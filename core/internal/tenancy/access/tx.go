package access

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/tenancy"
)

// inTx is the single transaction-owning call site of the facade: every guarded operation and Resolve authorizes and
// acts in exactly one transaction on one connection. Reads use REPEATABLE READ READ ONLY (one snapshot, no row locks);
// writes use READ COMMITTED READ WRITE with the ordered row locks of the authorization contract. fn's failure and any
// statement failure roll back (the deferred rollback is a no-op after a successful commit and also covers a panic). A
// failed COMMIT is classified by commitFailure, which claims a rollback only when one is established; success is
// returned only after Commit returns nil. There is no retry. A connection whose COMMIT or ROLLBACK outcome is not
// established is never reused (G008): pgx closes a connection on a connection-level failure, and pgxpool destroys any
// closed or non-idle connection when the transaction releases it. Closing it after that release could close a
// connection the pool had already handed to another caller, so the only explicit close is discard, while acquired.
//
// inTx is the transaction lifecycle only; every production call goes through inContextTx, which makes the context
// statement the first statement of the transaction (U-RLS-BOUNDARY).
//
// cleanup bounds the deferred rollback (G009 Plan r1 §4.3a, E11). cleanupUnbounded (0), used by every G007/G008
// operation, keeps the rollback on context.WithoutCancel(ctx) without a deadline, exactly as before. A positive value
// (the storage operations' cleanup budget) bounds it with context.WithTimeout(context.WithoutCancel(ctx), cleanup): an
// expired budget interrupts the rollback, pgx treats the failed rollback as a real rollback error and closes the
// connection while it is still held, and pgxpool destroys the closed connection on release, so no post-release close
// is added. After a successful COMMIT the deferred rollback is a no-op without I/O.
func (s *Service) inTx(ctx context.Context, write bool, cleanup time.Duration, fn func(tx *trackedTx) *failure) *failure {
	opts := pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
	if write {
		opts = pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite}
	}
	raw, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return fail(CodeUnavailable, failureReason(ctx, err))
	}
	defer func() {
		if cleanup <= 0 {
			_ = raw.Rollback(context.WithoutCancel(ctx))
			return
		}
		rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanup)
		_ = raw.Rollback(rctx)
		cancel()
	}()
	tx := &trackedTx{Tx: raw}
	if f := fn(tx); f != nil {
		return f
	}
	if err := raw.Commit(ctx); err != nil {
		return commitFailure(ctx, err)
	}
	return nil
}

// inContextTx runs fn in inTx after initializing the transaction-local request context (G008;
// docs/foundation/DATABASE-ISOLATION.md): tenancy.InitContext with dbc, the trusted subject, active Organization and
// selected Group, is the first statement, so row-level security admits only that context's rows. Prior context on the
// connection (a session-level value) is refused before anything protected runs: the connection is closed, so the pool
// destroys it on release, and the call fails closed as unavailable (store_unavailable).
//
// cleanup is passed to inTx: cleanupUnbounded for the G007/G008 operations, the storage cleanup budget for storageTx.
func (s *Service) inContextTx(ctx context.Context, write bool, dbc dbContext, cleanup time.Duration, fn func(tx *trackedTx) *failure) *failure {
	return s.inTx(ctx, write, cleanup, func(tx *trackedTx) *failure {
		if err := tenancy.InitContext(ctx, tx, dbc.user, dbc.organization, dbc.group); err != nil {
			if tenancy.CodeOf(err) == tenancy.CodeContextPresent {
				discard(tx.Tx)
				return fail(CodeUnavailable, reasonStoreUnavailable)
			}
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		return fn(tx)
	})
}

// cleanupUnbounded is the cleanup argument of every G007/G008 transaction: the deferred rollback runs on
// context.WithoutCancel(ctx) without a deadline, unchanged.
const cleanupUnbounded time.Duration = 0

// dbContext is the trusted database context of one transaction (tenancy.InitContext): the authenticated subject, the
// active Organization and the selected Group (zero: Organization-only).
type dbContext struct {
	user         tenancy.UserID
	organization tenancy.OrganizationID
	group        tenancy.GroupID
}

// discard closes the transaction's connection while the transaction still holds it (prior context found by the
// context statement). The deferred rollback then releases it, and pgxpool destroys a closed connection on release, so
// it is never handed out again.
func discard(tx pgx.Tx) {
	if c := tx.Conn(); c != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = c.Close(ctx)
		cancel()
	}
}

// trackedTx records the first driver error of its statements (a missing row is not one), so that the reason of a
// failure (aborted, canceled, store_unavailable) is derived without passing driver text anywhere. The identity and
// tenancy helpers map every failure onto their own closed codes.
type trackedTx struct {
	pgx.Tx
	cause error
}

func (t *trackedTx) note(err error) {
	if err != nil && t.cause == nil && !errors.Is(err, pgx.ErrNoRows) {
		t.cause = err
	}
}

func (t *trackedTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tag, err := t.Tx.Exec(ctx, sql, args...)
	t.note(err)
	return tag, err
}

func (t *trackedTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return trackedRow{row: t.Tx.QueryRow(ctx, sql, args...), t: t}
}

func (t *trackedTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := t.Tx.Query(ctx, sql, args...)
	t.note(err)
	if err != nil {
		return rows, err
	}
	return &trackedRows{Rows: rows, t: t}, nil
}

type trackedRow struct {
	row pgx.Row
	t   *trackedTx
}

func (r trackedRow) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	r.t.note(err)
	return err
}

type trackedRows struct {
	pgx.Rows
	t *trackedTx
}

func (r *trackedRows) Scan(dest ...any) error {
	err := r.Rows.Scan(dest...)
	r.t.note(err)
	return err
}

func (r *trackedRows) Err() error {
	err := r.Rows.Err()
	r.t.note(err)
	return err
}
