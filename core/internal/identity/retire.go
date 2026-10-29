package identity

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/audit"
)

// G019 trusted retirement finalization (Result F-01, accepted design variant; Codex binding review "F-01 bounded design
// ACCEPT; scope amendment02"). Ordinary retirement (core/internal/tenancy/access RetireUser) departs every Organization
// first and then disables the user; this file closes the window between the last departure and the disable. It is not
// the emergency DisableUser (store.go), whose immediate semantics are unchanged. G058 (D-11): both paths revoke the
// user's pending enrollment credentials identically (revokeEnrollmentSessionsAll, reason 'disabled', after the sessions
// UPDATE and before the gap read), so re-enabling the user never revives an old enrollment credential. G062 (SUP-10): both
// paths likewise revoke every pending root signup credential (revokeSignupSessionsAll, reason 'disabled', right after the
// enrollment revocation and before the gap read).

// RetirementCheck is the trusted, read-only final check of ordinary retirement. It runs exactly once, synchronously,
// inside the finalization transaction, after the subject's users row is locked FOR UPDATE, so its statements read fresh
// READ COMMITTED snapshots taken after any lock wait. It may only read through tx (Query and QueryRow): it must not
// retain tx or anything read from it, commit, roll back, execute, prepare or open a nested transaction, and it must not
// change the database context. The guarded view of the transaction filters methods and post-return calls only: Query
// and QueryRow pass any SQL through and do not themselves enforce read-only statements or an unchanged context. That
// guarantee is the contract of the single trusted, pinned callback (access's final check), not a SQL sandbox. A non-nil
// result refuses the finalization: it is returned unchanged and nothing is written.
type RetirementCheck func(ctx context.Context, tx pgx.Tx) error

// errCheckMisuse reports a refused transaction method called by a RetirementCheck (or after it returned).
var errCheckMisuse = errors.New("identity: retirement check used a refused transaction method")

// FinalizeRetirement is the trusted, route-less finalization of ordinary retirement (no HTTP or MCP surface; its one
// caller is access.RetireUser): in one identity transaction the subject's users row is locked FOR UPDATE, check runs
// once, and only when it returns nil is the user disabled and every session, pending enrollment credential and pending
// root signup credential revoked, with DisableUser's revocation priority. A check refusal is returned unchanged; on success the auth.session.revoked (scope disabled) event is
// emitted as DisableUser emits it.
func (s *Service) FinalizeRetirement(ctx context.Context, userID string, check RetirementCheck) error {
	if !canonicalUserID(userID) || check == nil {
		return ErrInvalidRequest
	}
	if s == nil || s.store == nil {
		return ErrUnavailable
	}
	err := s.store.FinalizeRetirement(ctx, userID, s.now(), check)
	if err == nil {
		s.event(ctx, slog.LevelInfo, "auth.session.revoked", slog.String("outcome", OutcomeLocalRevoked), slog.String("scope", "disabled"))
	}
	return err
}

// FinalizeRetirement runs the finalization at now. A non-canonical user ID or a nil check is invalid_request before any
// database work.
func (s *Store) FinalizeRetirement(ctx context.Context, userID string, now time.Time, check RetirementCheck) error {
	if !canonicalUserID(userID) || check == nil {
		return ErrInvalidRequest
	}
	return s.finalizeRetirement(ctx, userID, now, check)
}

// finalizeRetirement is the one READ COMMITTED transaction, in the subject's IDENTITY shape (E7, as DisableUser):
// users(subject) FOR UPDATE → check (definer reads, no row locks) → users UPDATE (row already held) → sessions(subject)
// UPDATE → enrollment_sessions(subject) UPDATE (G058 D-11) → signup_sessions(subject) UPDATE (G062 SUP-10) → revocationGap last → COMMIT → afterRevocation. A refusal, a misuse of the guarded view or any failure leaves
// the deferred rollback to write nothing; afterRevocation and the success event never run for a check refusal. An
// unknown COMMIT is identity_unavailable (never a claimed disable).
func (s *Store) finalizeRetirement(ctx context.Context, userID string, now time.Time, check RetirementCheck) error {
	var gap []string
	var refused error
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initIdentity(ctx, tx, userID); err != nil {
			return err
		}
		var id string
		if err := tx.QueryRow(ctx, `SELECT id::text FROM app.users WHERE id = $1 FOR UPDATE`, userID).Scan(&id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidRequest
			}
			return unavailable(err)
		}
		view := &checkTx{tx: tx, open: true}
		cerr := check(ctx, view)
		if view.seal() {
			return ErrUnavailable
		}
		if cerr != nil {
			refused = cerr
			return cerr
		}
		if _, err := tx.Exec(ctx, `
UPDATE app.users SET status = 'disabled', disabled_at = COALESCE(disabled_at, $2), updated_at = $2 WHERE id = $1`, userID, now); err != nil {
			return unavailable(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE app.sessions SET revoked_at = $2, revoke_reason = 'disabled' WHERE user_id = $1 AND revoked_at IS NULL`,
			userID, now); err != nil {
			return unavailable(err)
		}
		// G058 D-11: every pending enrollment credential of the user, exactly as DisableUser revokes them.
		if err := revokeEnrollmentSessionsAll(ctx, tx, userID, revokeDisabled, now); err != nil {
			return err
		}
		// G062 SUP-10: and every pending root signup credential of the user, exactly as DisableUser revokes them.
		if err := revokeSignupSessionsAll(ctx, tx, userID, revokeDisabled, now); err != nil {
			return err
		}
		var err error
		gap, err = revocationGap(ctx, tx)
		return err
	})
	if refused != nil {
		return refused
	}
	return s.afterRevocation(ctx, userID, revokeDisabled, audit.ResourceUser, userID, gap, err)
}

// checkTx is the guarded view of the finalization transaction handed to a RetirementCheck: Query and QueryRow pass
// through while the check runs; every other method, and every method after the check returned, is refused and marks a
// misuse, which fails the finalization whatever the check returned. It filters methods, not statements: Query and
// QueryRow do not parse or restrict the SQL they forward, so read-only SQL and an unchanged context are guaranteed by
// the trusted pinned callback, not by this wrapper.
type checkTx struct {
	tx pgx.Tx

	mu     sync.Mutex
	open   bool
	misuse bool
}

// seal closes the view and reports whether any refused method was called.
func (c *checkTx) seal() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.open = false
	return c.misuse
}

// readable reports whether a read may pass through now; otherwise it records the misuse.
func (c *checkTx) readable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.open {
		c.misuse = true
	}
	return c.open
}

// refuse records a call of a method the check may never use.
func (c *checkTx) refuse() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.misuse = true
	return errCheckMisuse
}

func (c *checkTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if !c.readable() {
		return nil, errCheckMisuse
	}
	return c.tx.Query(ctx, sql, args...)
}

func (c *checkTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if !c.readable() {
		return refusedRow{}
	}
	return c.tx.QueryRow(ctx, sql, args...)
}

func (c *checkTx) Begin(context.Context) (pgx.Tx, error) { return nil, c.refuse() }
func (c *checkTx) Commit(context.Context) error          { return c.refuse() }
func (c *checkTx) Rollback(context.Context) error        { return c.refuse() }

func (c *checkTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, c.refuse()
}

func (c *checkTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	_ = c.refuse()
	return refusedBatch{}
}

// LargeObjects is refused: the returned zero value is unusable.
func (c *checkTx) LargeObjects() pgx.LargeObjects {
	_ = c.refuse()
	return pgx.LargeObjects{}
}

func (c *checkTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, c.refuse()
}

func (c *checkTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, c.refuse()
}

// Conn is refused: the check never reaches the underlying connection.
func (c *checkTx) Conn() *pgx.Conn {
	_ = c.refuse()
	return nil
}

// refusedRow is the row of a refused QueryRow.
type refusedRow struct{}

func (refusedRow) Scan(...any) error { return errCheckMisuse }

// refusedBatch is the result of a refused SendBatch.
type refusedBatch struct{}

func (refusedBatch) Exec() (pgconn.CommandTag, error) { return pgconn.CommandTag{}, errCheckMisuse }
func (refusedBatch) Query() (pgx.Rows, error)         { return nil, errCheckMisuse }
func (refusedBatch) QueryRow() pgx.Row                { return refusedRow{} }
func (refusedBatch) Close() error                     { return errCheckMisuse }
