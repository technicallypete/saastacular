package identity

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/audit"
	"core/internal/observability"
)

// Store is the PostgreSQL-authoritative identity state (schema app, 00002_identity.sql), protected by row-level
// security from version 6 (00006_row_level_security.sql; docs/foundation/DATABASE-ISOLATION.md). It connects as the
// runtime role. Credential resolution (E1), provider-principal lookup (E2) and logout-reference lookup (E3) call the
// three bounded SELECT-only reader functions, which ignore ambient settings. Every other read or write of a protected
// table runs in a transaction whose first statement initializes the transaction-local context: IDENTITY (a subject
// E1/E2/E3 established, or the trusted lifecycle caller's target), LOGIN (the hash of the browser's login binding) or
// EMPTY (only before the expired-login cleanup). The provider binding is public namespace configuration without
// row-level security. There is no cache: every resolution reads PostgreSQL. All times come from the service clock,
// except the expired-login cleanup, which uses the database clock.
type Store struct {
	db storeDB
	// logger receives the G011 audit diagnostics (audit.record_failed, audit.gap); NewService sets it from its Config.
	logger *slog.Logger
}

// storeDB is the store's database boundary; *pgxpool.Pool satisfies it. Tests substitute a hooking implementation to
// inject faults into identity transactions.
type storeDB interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// NewStore returns a store over pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{db: pool} }

// newStore returns a store over any storeDB (fault-injection tests).
func newStore(db storeDB) *Store { return &Store{db: db} }

// unavailable wraps any database failure as identity_unavailable, never as an authentication outcome.
func unavailable(err error) error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return ErrUnavailable
}

// renewEvery bounds session renewal writes (Plan r5 §4.3).
const renewEvery = 60 * time.Second

// sqlInitContext is the one context statement, byte-identical to core/internal/tenancy's (U-RLS-BOUNDARY): it sets all
// four transaction-local fields and returns one row only when none of them was already set on the connection. G011: it
// also sets the two unguarded audit correlation settings and resets the revocation gap marker (docs/foundation/AUDIT.md).
const sqlInitContext = `SELECT pg_catalog.set_config('app.user_id', $1, true), pg_catalog.set_config('app.organization_id', $2, true),
       pg_catalog.set_config('app.group_id', $3, true), pg_catalog.set_config('app.login_binding', $4, true),
       pg_catalog.set_config('audit.request_id', $5, true), pg_catalog.set_config('audit.trace_id', $6, true),
       pg_catalog.set_config('audit.revocation_gap', '', true)
WHERE coalesce(pg_catalog.current_setting('app.user_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.organization_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.group_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.login_binding', true), '') = ''`

// The three bounded reader calls (E1-E3) and the expired-login cleanup (E5).
const (
	sqlAuthResolveSession = `SELECT session_id::text, user_id::text, display_name, mfa_assurance, auth_time, idle_expires_at, ` +
		`absolute_expires_at, renewed_at, csrf_token, provider_session_id, origin, organization_id::text ` +
		`FROM app.auth_resolve_session($1, $2, $3)`
	sqlAuthPrincipalSubject = `SELECT app.auth_principal_subject($1, $2)::text`
	sqlAuthLogoutRef        = `SELECT session_id::text, user_id::text FROM app.auth_logout_ref($1, $2)`
	// E5: exactly this statement (no WHERE, no RETURNING), only immediately after initEmpty in the login-start
	// transaction. In the EMPTY shape row-level security admits exactly the rows expired by the database clock.
	sqlExpireLoginTransactions = `DELETE FROM app.login_transactions`
	// G058 (Plan r3 §1.3 step 1, §4.4): the expired enrollment credentials, removed by the same login-start
	// transaction right after the login cleanup, in the same EMPTY shape, where enrollment_sessions_expire admits
	// exactly the rows expired by the database clock (no WHERE, no RETURNING).
	sqlExpireEnrollmentSessions = `DELETE FROM app.enrollment_sessions`
	// G062 (Plan r5 §3.5): the expired root signup credentials, removed by the same login-start transaction right after
	// the enrollment cleanup, in the same EMPTY shape, where signup_sessions_expire admits exactly the rows expired by
	// the database clock (no WHERE, no RETURNING).
	sqlExpireSignupSessions = `DELETE FROM app.signup_sessions`
)

// G062 (Plan r5 §4.3; SUP-10): the pending root signup credentials are revoked beside the enrollment credentials, in the
// same transaction and scope: the bound paths (current-session logout, provider logout, supersession) revoke only the
// rows minted from the provider session of the local session being revoked; logout-all and DisableUser revoke every
// unrevoked row of the user. Revocation is terminal.
const (
	sqlRevokeSignupAll   = `UPDATE app.signup_sessions SET revoked_at = $2, revoke_reason = $3 WHERE user_id = $1 AND revoked_at IS NULL`
	sqlRevokeSignupBound = `UPDATE app.signup_sessions SET revoked_at = $3, revoke_reason = $4 WHERE user_id = $1 AND revoked_at IS NULL ` +
		`AND provider_session_id = (SELECT s.provider_session_id FROM app.sessions s WHERE s.id = $2 AND s.user_id = $1)`
)

// revokeSignupSessionsAll revokes every unrevoked root signup credential of userID with reason, inside tx (the caller's
// IDENTITY transaction for userID; signup_sessions_revoke admits only that user's rows).
func revokeSignupSessionsAll(ctx context.Context, tx pgx.Tx, userID, reason string, now time.Time) error {
	if _, err := tx.Exec(ctx, sqlRevokeSignupAll, userID, now, reason); err != nil {
		return unavailable(err)
	}
	return nil
}

// G058 pending enrollment credentials are revoked at the revoking path's actual scope (Plan r3 §1.3 step 5, D-18;
// readiness correction 01 §3, D-11), in that path's transaction, after its sessions UPDATE and before its revocation-gap
// read. The broader paths (logout-all, DisableUser, FinalizeRetirement) revoke every unrevoked row of the user; the bound
// paths (current-session logout, provider logout, supersession) revoke only the rows minted from the provider session of
// the local session being revoked. Revocation is terminal: re-enabling a user revives no enrollment row.
const (
	sqlRevokeEnrollmentAll   = `UPDATE app.enrollment_sessions SET revoked_at = $2, revoke_reason = $3 WHERE user_id = $1 AND revoked_at IS NULL`
	sqlRevokeEnrollmentBound = `UPDATE app.enrollment_sessions SET revoked_at = $3, revoke_reason = $4 WHERE user_id = $1 AND revoked_at IS NULL ` +
		`AND provider_session_id = (SELECT s.provider_session_id FROM app.sessions s WHERE s.id = $2 AND s.user_id = $1)`
)

// revokeEnrollmentSessionsAll revokes every unrevoked enrollment credential of userID with reason, inside tx (the
// caller's IDENTITY transaction for userID; the own-row policy admits only that user's rows). It is the one shared
// statement of revokeAll, DisableUser and finalizeRetirement.
func revokeEnrollmentSessionsAll(ctx context.Context, tx pgx.Tx, userID, reason string, now time.Time) error {
	if _, err := tx.Exec(ctx, sqlRevokeEnrollmentAll, userID, now, reason); err != nil {
		return unavailable(err)
	}
	return nil
}

// initContext runs the context statement as the first statement of tx with the given user and login binding (both
// empty for EMPTY), and the request and trace correlation of ctx as the diagnostic audit correlation (G011). No row
// means the connection carried prior context: it is closed (the pool destroys it on release) and the call fails closed
// before anything protected runs.
func initContext(ctx context.Context, tx pgx.Tx, user, binding string) error {
	return initContextShape(ctx, tx, user, "", binding)
}

// initContextShape is initContext with an Organization: "" for EMPTY, LOGIN and IDENTITY; an Organization UUID (with a
// user and no binding) for the TENANT-ORG shape of an MCP connection read (G014). It never sets a Group.
func initContextShape(ctx context.Context, tx pgx.Tx, user, org, binding string) error {
	var a, b, c, d, e, f, g string
	err := tx.QueryRow(ctx, sqlInitContext, user, org, "", binding, observability.RequestID(ctx), observability.TraceID(ctx)).
		Scan(&a, &b, &c, &d, &e, &f, &g)
	if errors.Is(err, pgx.ErrNoRows) {
		discard(tx)
		return ErrUnavailable
	}
	if err != nil {
		return ErrUnavailable
	}
	return nil
}

// initEmpty verifies that the connection carries no context and leaves all four fields empty (the EMPTY shape).
func initEmpty(ctx context.Context, tx pgx.Tx) error { return initContext(ctx, tx, "", "") }

// initIdentity initializes the IDENTITY shape for userID: a subject core established (E1-E3, sign-in) or the trusted
// lifecycle caller's target, never request input.
func initIdentity(ctx context.Context, tx pgx.Tx, userID string) error {
	return initContext(ctx, tx, userID, "")
}

// initTenantOrg initializes the TENANT-ORG shape for userID in org (G014): the mapped subject of a verified MCP token
// and the Organization the trusted host registry names, never request input. Row-level security then admits only the
// subject's own MCP connections of that Organization.
func initTenantOrg(ctx context.Context, tx pgx.Tx, userID, org string) error {
	return initContextShape(ctx, tx, userID, org, "")
}

// initLogin initializes the LOGIN shape for the SHA-256 of the browser's login binding.
func initLogin(ctx context.Context, tx pgx.Tx, bindingHash []byte) error {
	return initContext(ctx, tx, "", hex.EncodeToString(bindingHash))
}

// discard closes the transaction's connection while the transaction still holds it; the deferred rollback then
// releases it and pgxpool destroys a closed connection on release, so it is never handed out again. It is called only
// while the connection is acquired (prior context found by an initializer).
func discard(tx pgx.Tx) {
	if c := tx.Conn(); c != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = c.Close(ctx)
		cancel()
	}
}

// begin starts one identity transaction (READ COMMITTED READ WRITE).
func (s *Store) begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite})
	if err != nil {
		return nil, unavailable(err)
	}
	return tx, nil
}

// inTx runs fn in one transaction on one connection; fn's first statement is one of the initializers. Success only
// after Commit returns nil; any other outcome is identity_unavailable. A connection whose COMMIT or ROLLBACK outcome
// is not established is never reused: pgx closes a connection on a connection-level failure (a lost reply), and
// pgxpool destroys every connection that is closed or not idle when the transaction releases it. Closing it here,
// after the release, could close a connection the pool had already handed to another caller.
func (s *Store) inTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := s.begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)
	if err := fn(tx); err != nil {
		return err
	}
	return commit(ctx, tx)
}

// rollback is the deferred rollback of every identity transaction: a no-op after a commit. A failed rollback leaves
// a closed or non-idle connection, which pgxpool destroys on release.
func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(context.WithoutCancel(ctx))
}

// commit commits tx. On failure the outcome is not established: identity_unavailable, never success, and no retry.
func commit(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		return ErrUnavailable
	}
	return nil
}

// errNoCommit ends an inTxAudited transaction without committing: the helper rolls back and returns it (signIn's
// existing rollback-only outcomes, already_signed_in and identity_conflict).
var errNoCommit = errors.New("identity: no commit")

// inTxAudited is inTx for a mutation with its required audit append (G011, R4; docs/foundation/AUDIT.md): the same
// begin, deferred rollback and commit, so the identity semantics and the unknown-COMMIT truth are unchanged. fn performs
// the mutation and returns the core-owned events it requires; they are appended through audit.Append on the same
// transaction after fn succeeds and before COMMIT. An append failure is a failed transaction: nothing commits,
// identity_unavailable.
func (s *Store) inTxAudited(ctx context.Context, fn func(tx pgx.Tx) ([]audit.Event, error)) error {
	tx, err := s.begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)
	events, err := fn(tx)
	if err != nil {
		return err
	}
	for _, e := range events {
		if err := audit.Append(ctx, tx, e); err != nil {
			return ErrUnavailable
		}
	}
	return commit(ctx, tx)
}

// The bounds of an independent audit record (G011 Plan §4.3): BEGIN through COMMIT, and the deferred rollback.
const (
	auditRecordBudget  = 2 * time.Second
	auditCleanupBudget = 2 * time.Second
)

// recordAuditEvents is the bounded identity recorder of independent core-owned events (G1 denied/failed after the
// one-use transaction was consumed, and G6 after a committed revocation with an audit gap). It does not use inTx: it
// opens its own transaction on a context detached from the caller's cancellation and bounded by auditRecordBudget, and
// its deferred rollback runs on a fresh context bounded by auditCleanupBudget (an expired cleanup budget makes pgx
// close the held connection, so pgxpool destroys it on release). It initializes IDENTITY for subject, or EMPTY for an
// unattributed event; a prior context discards the connection and fails before anything is appended. No retry.
func (s *Store) recordAuditEvents(ctx context.Context, subject string, events ...audit.Event) error {
	base := context.WithoutCancel(ctx)
	bctx, cancel := context.WithTimeout(base, auditRecordBudget)
	defer cancel()
	tx, err := s.begin(bctx)
	if err != nil {
		return err
	}
	defer func() {
		rctx, rcancel := context.WithTimeout(base, auditCleanupBudget)
		_ = tx.Rollback(rctx)
		rcancel()
	}()
	if subject != "" {
		err = initIdentity(bctx, tx, subject)
	} else {
		err = initEmpty(bctx, tx)
	}
	if err != nil {
		return err
	}
	for _, e := range events {
		if err := audit.Append(bctx, tx, e); err != nil {
			return ErrUnavailable
		}
	}
	return commit(bctx, tx)
}

// The G011 audit diagnostics (events.json): audit.record_failed (an independent record could not be written; the
// caller's outcome is unchanged) and audit.gap (a committed revocation whose row-change event could not be written).
const (
	eventAuditRecordFailed = "audit.record_failed"
	eventAuditGap          = "audit.gap"
)

// The audit.gap independent-record outcomes.
const (
	gapRecorded     = "recorded"
	gapFailed       = "failed"
	gapNotAttempted = "not_attempted"
)

// diagnostic writes one G011 audit diagnostic (error level) with request/trace correlation when present.
func (s *Store) diagnostic(ctx context.Context, name string, attrs ...slog.Attr) {
	if s.logger == nil {
		return
	}
	if rid := observability.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid), slog.String("trace_id", observability.TraceID(ctx)))
	}
	s.logger.LogAttrs(context.Background(), slog.LevelError, name, attrs...)
}

// recordIndependent records events with the bounded recorder; a failure writes one audit.record_failed diagnostic per
// event and never changes the caller's outcome.
func (s *Store) recordIndependent(ctx context.Context, subject string, events ...audit.Event) {
	if err := s.recordAuditEvents(ctx, subject, events...); err != nil {
		for _, e := range events {
			s.diagnostic(ctx, eventAuditRecordFailed, slog.String("action", string(e.Action)))
		}
	}
}

// afterRevocation completes a revocation-priority transaction (M5, owner decision; Spec R4/AC5). gap is the marker read
// before COMMIT and err the transaction's outcome. Without a gap it returns err unchanged. With a gap and a committed
// revocation it attempts one independent revocation.unaudited record (resource: the session for logout and
// provider_logout, the user for logout_all and disabled), reports audit.gap and returns success: an audit-only fault
// never prevents or undoes a local revocation. With a gap and a failed or unknown COMMIT it reports audit.gap
// (not_attempted) and returns the failure (identity_unavailable), never success.
func (s *Store) afterRevocation(ctx context.Context, userID, reason string, resource audit.ResourceType, resourceID string, gap []string, err error) error {
	if len(gap) == 0 {
		return err
	}
	if err != nil {
		s.diagnostic(ctx, eventAuditGap, slog.String("operation", reason), slog.String("independent", gapNotAttempted))
		return err
	}
	subject, _ := audit.ParseUUID(userID)
	rid, _ := audit.ParseUUID(resourceID)
	e := audit.Event{Action: audit.RevocationUnaudited, Outcome: audit.Success, Reason: audit.Reason(reason),
		ResourceType: resource, ResourceID: rid, Subject: subject}
	independent := gapFailed
	if s.recordAuditEvents(ctx, userID, e) == nil {
		independent = gapRecorded
	}
	s.diagnostic(ctx, eventAuditGap, slog.String("operation", reason), slog.String("independent", independent))
	return nil
}

// revocationGap reads the gap marker as the last statement of a revocation transaction, before COMMIT.
func revocationGap(ctx context.Context, tx pgx.Tx) ([]string, error) {
	gap, err := audit.RevocationGap(ctx, tx)
	if err != nil {
		return nil, ErrUnavailable
	}
	return gap, nil
}

// newUserID returns a random (version 4) canonical UUID for a first sign-in (decided Q3): core generates it because the
// IDENTITY context must name the subject before its row exists. It is never caller-supplied.
func newUserID() (string, error) {
	b, err := newSecret()
	if err != nil {
		return "", err
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// loginTxn is one pending login transaction.
type loginTxn struct {
	StateHash   []byte
	BindingHash []byte
	Verifier    string
	ClientID    string
	Origin      string
	ReturnPath  string
	MaxAge      int
	CreatedAt   time.Time
	ExpiresAt   time.Time
	// G016: the Organization pinned at a tenant-host login and, for a custom host, the shared relay origin.
	OrganizationID string
	CallbackOrigin string
	// G058 (Plan r3 §1.3 step 1, §4.1): the enrollment-intent login's branch facts (the callback's only branch selector).
	EnrollmentIntent bool
	InvitationID     string
	RequestOnArrival bool
	// G062 (Plan r5 §3.1): a root signup-intent login (no Organization, return path /signup/details) and a pinned tenant
	// handoff login (the Organization and the expected user fixed by the continuation's completion).
	SignupIntent   bool
	SignupHandoff  bool
	ExpectedUserID string
	// hostChanged is the consume transaction's pre-exchange host check: the origin no longer routes to OrganizationID.
	hostChanged bool
}

// binding returns the bound provider namespace, if any (E6: public namespace configuration, no row-level security).
func (s *Store) binding(ctx context.Context) (issuer, clientID string, found bool, err error) {
	err = s.db.QueryRow(ctx, `SELECT issuer, client_id FROM app.identity_provider_binding WHERE provider = 'workos'`).
		Scan(&issuer, &clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, unavailable(err)
	}
	return issuer, clientID, true, nil
}

// insertTransaction is login start (F3): in one transaction it verifies an empty context (EMPTY), removes the
// transactions expired by the database clock (E5) and, in the same EMPTY shape, the expired enrollment credentials
// (G058), initializes LOGIN for the browser's binding and records the new transaction. Prior context on the connection fails closed before anything is deleted; the cleanup and the insert
// commit together.
// G016 (Plan r4 §5.5): admit runs first inside the transaction (the tenant-host resolution, one statement on tx) and
// completes the row; its refusal rolls back with nothing inserted.
func (s *Store) insertLogin(ctx context.Context, admit func(tx pgx.Tx) (loginTxn, error)) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initEmpty(ctx, tx); err != nil {
			return err
		}
		t, err := admit(tx)
		if err != nil {
			return err
		}
		var maxAge *int
		if t.MaxAge > 0 {
			maxAge = &t.MaxAge
		}
		if _, err := tx.Exec(ctx, sqlExpireLoginTransactions); err != nil {
			return unavailable(err)
		}
		if _, err := tx.Exec(ctx, sqlExpireEnrollmentSessions); err != nil {
			return unavailable(err)
		}
		if _, err := tx.Exec(ctx, sqlExpireSignupSessions); err != nil {
			return unavailable(err)
		}
		if err := initLogin(ctx, tx, t.BindingHash); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO app.login_transactions
  (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, max_age, created_at, expires_at, organization_id,
   callback_origin, enrollment_intent, invitation_id, request_on_arrival, signup_intent, signup_handoff, expected_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, '')::uuid, NULLIF($11, ''), $12, NULLIF($13, '')::uuid, $14, $15, $16,
        NULLIF($17, '')::uuid)`,
			t.StateHash, t.BindingHash, t.Verifier, t.ClientID, t.Origin, t.ReturnPath, maxAge, t.CreatedAt, t.ExpiresAt,
			t.OrganizationID, t.CallbackOrigin, t.EnrollmentIntent, t.InvitationID, t.RequestOnArrival, t.SignupIntent,
			t.SignupHandoff, t.ExpectedUserID)
		return unavailable(err)
	})
}

// insertTransaction stores a complete row with no admission step (store-level callers and tests).
func (s *Store) insertTransaction(ctx context.Context, t loginTxn) error {
	return s.insertLogin(ctx, func(pgx.Tx) (loginTxn, error) { return t, nil })
}

// consumeTransaction consumes with no host check (store-level callers and tests).
func (s *Store) consumeTransaction(ctx context.Context, stateHash, bindingHash []byte, clientID, origin string, now time.Time) (loginTxn, consumeResult, error) {
	return s.consumeLogin(ctx, stateHash, bindingHash, clientID, origin, now, func(pgx.Tx, loginTxn) (bool, error) { return true, nil })
}

type consumeResult int

const (
	consumeNone    consumeResult = iota // unknown, replayed, wrong binding/client/origin: nothing consumed
	consumeExpired                      // matched but expired: deleted, unusable
	consumeOK
)

// consumeTransaction (F4) atomically consumes the transaction only when state, binding, client and origin all match,
// in the LOGIN shape of the presented binding. A wrong or missing binding never consumes the victim's transaction;
// parallel identical callbacks consume once.
//
// G016 (Plan r4 §1.7 leg 3, §5.5): hostOK runs inside the same transaction after a successful consumption (one
// statement on tx); a false answer still commits the consumption and is reported as hostChanged, a store failure
// rolls everything back (nothing consumed).
func (s *Store) consumeLogin(ctx context.Context, stateHash, bindingHash []byte, clientID, origin string, now time.Time,
	hostOK func(tx pgx.Tx, t loginTxn) (bool, error)) (loginTxn, consumeResult, error) {
	t := loginTxn{StateHash: stateHash, BindingHash: bindingHash, ClientID: clientID, Origin: origin}
	var maxAge *int
	found := false
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initLogin(ctx, tx, bindingHash); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `
DELETE FROM app.login_transactions
WHERE state_hash = $1 AND binding_hash = $2 AND client_id = $3 AND origin = $4
RETURNING pkce_verifier, return_path, max_age, created_at, expires_at, coalesce(organization_id::text, ''),
          coalesce(callback_origin, ''), enrollment_intent, coalesce(invitation_id::text, ''), request_on_arrival,
          signup_intent, signup_handoff, coalesce(expected_user_id::text, '')`,
			stateHash, bindingHash, clientID, origin).Scan(&t.Verifier, &t.ReturnPath, &maxAge, &t.CreatedAt, &t.ExpiresAt,
			&t.OrganizationID, &t.CallbackOrigin, &t.EnrollmentIntent, &t.InvitationID, &t.RequestOnArrival,
			&t.SignupIntent, &t.SignupHandoff, &t.ExpectedUserID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return unavailable(err)
		}
		found = true
		ok, err := hostOK(tx, t)
		if err != nil {
			return err
		}
		t.hostChanged = !ok
		return nil
	})
	if err != nil {
		return loginTxn{}, consumeNone, err
	}
	if !found {
		return loginTxn{}, consumeNone, nil
	}
	if maxAge != nil {
		t.MaxAge = *maxAge
	}
	if !t.ExpiresAt.After(now) {
		return loginTxn{}, consumeExpired, nil
	}
	return t, consumeOK, nil
}

// resolve (F1) returns the live session for a credential hash through the E1 reader, renewing its idle expiry at most
// once per renewEvery in an IDENTITY transaction of the resolved subject. Revoked, expired, unknown and disabled-user
// sessions are unauthenticated. Renewal never touches revoked_at or the absolute expiry; if a revocation commits
// between the read and the renewal, the request is refused rather than served from the stale read. G016 security
// amendment: the reader returns only a browser session bound to exactly origin (with its Organization); an unbound,
// mcp or other-origin row is unauthenticated.
func (s *Store) resolve(ctx context.Context, credHash []byte, origin string, now time.Time, idle time.Duration) (SessionInfo, error) {
	var (
		info              SessionInfo
		display, sid      *string
		idleExp, absolute time.Time
		renewed           time.Time
	)
	err := s.db.QueryRow(ctx, sqlAuthResolveSession, credHash, now, origin).
		Scan(&info.SessionID, &info.UserID, &display, &info.MFAAssurance, &info.AuthTime, &idleExp, &absolute, &renewed,
			&info.CSRFToken, &sid, &info.Origin, &info.OrganizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return SessionInfo{}, ErrUnauthenticated
	}
	if err != nil {
		return SessionInfo{}, unavailable(err)
	}
	if now.Sub(renewed) >= renewEvery {
		err := s.inTx(ctx, func(tx pgx.Tx) error {
			if err := initIdentity(ctx, tx, info.UserID); err != nil {
				return err
			}
			err := tx.QueryRow(ctx, `
UPDATE app.sessions
SET idle_expires_at = LEAST($3::timestamptz, absolute_expires_at), renewed_at = $2
WHERE id = $1 AND revoked_at IS NULL AND idle_expires_at > $2
RETURNING idle_expires_at`, info.SessionID, now, now.Add(idle)).Scan(&idleExp)
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrUnauthenticated
			}
			return unavailable(err)
		})
		if err != nil {
			return SessionInfo{}, err
		}
	}
	if display != nil {
		info.DisplayName = *display
	}
	if sid != nil {
		info.ProviderSessionID = *sid
	}
	info.ExpiresAt = idleExp
	if absolute.Before(idleExp) {
		info.ExpiresAt = absolute
	}
	return info, nil
}

// Transaction-scoped liveness of one known session (G007, docs/foundation/AUTHORIZATION.md): the user is active and the
// session belongs to it, is unrevoked and has not reached either expiry at the service clock's now. The locking forms
// take FOR SHARE on the user row first and then on the session row, the order of DisableUser (users FOR UPDATE, then
// the sessions UPDATE), so a guarded write holding them is ordered before any later disable, logout, logout-all,
// provider logout or supersession of that session, or it observes it. They never renew a session. They run in the
// caller's tenant-shaped transaction, whose context admits exactly the subject's own rows.
const (
	sqlLiveUser        = `SELECT id FROM app.users WHERE id = $1 AND status = 'active'`
	sqlLiveSession     = `SELECT id FROM app.sessions WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL AND idle_expires_at > $3 AND absolute_expires_at > $3`
	sqlLockLiveUser    = sqlLiveUser + ` FOR SHARE`
	sqlLockLiveSession = sqlLiveSession + ` FOR SHARE`
)

// liveSession runs the two liveness statements in tx (the caller's transaction): unauthenticated when either row is
// absent, identity_unavailable on any other failure (the transaction is then unusable; the caller rolls it back).
func (s *Store) liveSession(ctx context.Context, tx pgx.Tx, sessionID, userID string, now time.Time, lock bool) error {
	userSQL, sessionSQL := sqlLiveUser, sqlLiveSession
	if lock {
		userSQL, sessionSQL = sqlLockLiveUser, sqlLockLiveSession
	}
	var id string
	if err := tx.QueryRow(ctx, userSQL, userID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUnauthenticated
		}
		return ErrUnavailable
	}
	if err := tx.QueryRow(ctx, sessionSQL, sessionID, userID, now).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUnauthenticated
		}
		return ErrUnavailable
	}
	return nil
}

// signInParams is one validated callback ready to commit.
type signInParams struct {
	Identity VerifiedIdentity
	Live     *SessionInfo // the forwarded live session, if any
	Now      time.Time
	Idle     time.Duration
	Max      time.Duration
	// G016: the tenant-host origin, the Organization pinned at login and the resolver of the final issuance guard.
	Origin       string
	Organization string
	Hosts        HostResolver
	// G062 (Plan r5 §1.4 user pin): the expected user of a pinned tenant handoff login ("" otherwise). The verified
	// principal must map to it before any write (signup_user_mismatch; no mapping is created), and a success is the
	// tenant_created outcome.
	ExpectedUser string
}

// sqlIsOrganizationMember is the D-2(a) membership read through the auth-reader function (no lock).
const sqlIsOrganizationMember = `SELECT app.is_organization_member($1::uuid, $2::uuid)`

// issuanceGuard is the final tenant-host session-issuance guard (Plan r4 §1.7.1), run inside signIn after the
// principal mapping decision and before the first write of either branch: the host must still route to the login's
// Organization (host_changed) and, under D-2(a), the mapped user must currently be a member of it
// (organization_not_member; an unmapped principal is never a member, so nothing is created). Both are plain snapshot
// reads of SECURITY DEFINER readers at READ COMMITTED: no lock and no network inside the transaction. A change that
// commits after these reads is refused at the session's next request by per-request resolution and authorization.
// Fail closed: a sign-in without a resolver or without a pinned Organization is host_changed (no store-level bypass).
func issuanceGuard(ctx context.Context, tx pgx.Tx, p signInParams, mapped string) error {
	if p.Hosts == nil || p.Organization == "" {
		return ErrHostChanged
	}
	res, err := p.Hosts.ResolveOrigin(ctx, tx, p.Origin)
	if err != nil {
		return unavailable(err)
	}
	if !res.Routing || res.OrganizationID != p.Organization {
		return ErrHostChanged
	}
	if mapped == "" {
		return ErrNotMember
	}
	var member bool
	if err := tx.QueryRow(ctx, sqlIsOrganizationMember, mapped, p.Organization).Scan(&member); err != nil {
		return unavailable(err)
	}
	if !member {
		return ErrNotMember
	}
	return nil
}

// signInResult reports the committed effect.
type signInResult struct {
	Outcome    string // signed_in, session_rotated, already_signed_in
	Credential []byte // raw credential of a new session (signed_in, session_rotated)
	ExpiresAt  time.Time
	Conflict   bool   // a live session for another user: nothing was committed
	Subject    string // with ErrUserDisabled: the verified mapped subject (G011 attribution of auth.sign_in user_disabled)
}

// signIn (F2) commits a validated callback in one transaction: the namespace binding (first mapping inserts it), the
// verified principal's subject through the E2 reader, the continuity effect (Plan r5 §4.2), find-or-create of the exact
// (issuer, subject) mapping, the active-status check under FOR SHARE (serialized with DisableUser's FOR UPDATE), and the
// new session, every write in the IDENTITY shape of the mapped subject. Nothing usable is published on any failure.
//
// G011: the transaction runs through inTxAudited, so both success branches (signed_in, session_rotated) append their
// auth.sign_in success event in the same transaction, and the row changes (user, supersession, session) are recorded by
// the capture triggers. already_signed_in and identity_conflict commit nothing (errNoCommit). ErrUserDisabled returns
// the mapped subject in Subject.
func (s *Store) signIn(ctx context.Context, p signInParams) (signInResult, error) {
	var res signInResult
	err := s.inTxAudited(ctx, func(tx pgx.Tx) ([]audit.Event, error) {
		vi := p.Identity
		if _, err := tx.Exec(ctx, `
INSERT INTO app.identity_provider_binding (provider, issuer, client_id, bound_at) VALUES ('workos', $1, $2, $3)
ON CONFLICT (provider) DO NOTHING`, vi.Issuer, vi.ClientID, p.Now); err != nil {
			return nil, unavailable(err)
		}
		var boundIssuer, boundClient string
		if err := tx.QueryRow(ctx, `SELECT issuer, client_id FROM app.identity_provider_binding WHERE provider = 'workos'`).
			Scan(&boundIssuer, &boundClient); err != nil {
			return nil, unavailable(err)
		}
		if boundIssuer != vi.Issuer || boundClient != vi.ClientID {
			return nil, ErrNamespaceMismatch
		}

		mapped, err := principalSubject(ctx, tx, vi.Issuer, vi.Subject)
		if err != nil {
			return nil, err
		}
		// G062: a pinned handoff completes only for its expected user, before any write or continuity effect.
		if p.ExpectedUser != "" && mapped != p.ExpectedUser {
			return nil, ErrSignupUserMismatch
		}
		rotated, signedIn := OutcomeSessionRotated, OutcomeSignedIn
		rotatedReason, signedInReason := audit.ReasonSessionRotated, audit.ReasonSignedIn
		if p.ExpectedUser != "" {
			rotated, signedIn = OutcomeTenantCreated, OutcomeTenantCreated
			rotatedReason, signedInReason = audit.ReasonTenantCreated, audit.ReasonTenantCreated
		}

		if p.Live != nil {
			// Continuity: the browser already has a live session for user A.
			if mapped == "" || mapped != p.Live.UserID {
				res = signInResult{Conflict: true} // no mapping is created for the other principal
				return nil, errNoCommit
			}
			if p.Live.ProviderSessionID == vi.ProviderSessionID {
				res = signInResult{Outcome: OutcomeAlreadySignedIn} // SSO returned the same provider session
				return nil, errNoCommit
			}
			if err := issuanceGuard(ctx, tx, p, mapped); err != nil {
				return nil, err // before the supersession: the live session stays untouched
			}
			if err := initIdentity(ctx, tx, mapped); err != nil {
				return nil, err
			}
			// An additional provider session for A: supersede the old local session, bind a new one to the new sid.
			tag, err := tx.Exec(ctx, `
UPDATE app.sessions SET revoked_at = $2, revoke_reason = 'superseded' WHERE id = $1 AND revoked_at IS NULL`,
				p.Live.SessionID, p.Now)
			if err != nil {
				return nil, unavailable(err)
			}
			if tag.RowsAffected() != 1 {
				res = signInResult{Conflict: true} // revoked concurrently: treat as not continuous
				return nil, errNoCommit
			}
			// G058 D-18 bound path: the pending enrollment credentials minted from the superseded provider session.
			if _, err := tx.Exec(ctx, sqlRevokeEnrollmentBound, mapped, p.Live.SessionID, p.Now, revokeSuperseded); err != nil {
				return nil, unavailable(err)
			}
			// G062: and the pending root signup credentials minted from it.
			if _, err := tx.Exec(ctx, sqlRevokeSignupBound, mapped, p.Live.SessionID, p.Now, revokeSuperseded); err != nil {
				return nil, unavailable(err)
			}
			r, err := s.issue(ctx, tx, mapped, p)
			if err != nil {
				return nil, disabledSubject(&res, mapped, err)
			}
			r.Outcome = rotated
			res = r
			return []audit.Event{signInEvent(mapped, rotatedReason)}, nil
		}

		if err := issuanceGuard(ctx, tx, p, mapped); err != nil {
			return nil, err // before createMapping and issue: no user, mapping or session is created
		}
		if mapped == "" {
			if mapped, err = s.createMapping(ctx, tx, vi, p.Now); err != nil {
				return nil, err
			}
		} else if err := initIdentity(ctx, tx, mapped); err != nil {
			return nil, err
		}
		r, err := s.issue(ctx, tx, mapped, p)
		if err != nil {
			return nil, disabledSubject(&res, mapped, err)
		}
		r.Outcome = signedIn
		res = r
		return []audit.Event{signInEvent(mapped, signedInReason)}, nil
	})
	switch {
	case errors.Is(err, errNoCommit):
		return res, nil
	case errors.Is(err, ErrUserDisabled):
		return signInResult{Subject: res.Subject}, err
	case err != nil:
		return signInResult{}, err
	}
	return res, nil
}

// sqlInsertSignup is T1's issuing INSERT of the root signup row in the IDENTITY shape of the mapped subject
// (signup_sessions_insert; the issuing columns only).
const sqlInsertSignup = `INSERT INTO app.signup_sessions (id, user_id, origin, credential_hash, csrf_token, provider_session_id, ` +
	`created_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

// signupT1Result is T1's committed effect.
type signupT1Result struct {
	Credential []byte
	ExpiresAt  time.Time
	Subject    string // the mapped subject (attribution of user_disabled)
}

// signupT1 is the root signup callback's transaction (G062 Plan r5 §1.2): the namespace binding (as signIn), the exact
// (issuer, subject) mapping found through E2 or created (createMapping; never linked by e-mail), the active-status check
// under FOR SHARE (serialized with DisableUser), and one signup row with a fresh credential and CSRF token, all in the
// IDENTITY shape of the mapped subject, with auth.sign_in success signup_authenticated in the same transaction. No
// session, Organization or tenant context is created.
func (s *Store) signupT1(ctx context.Context, vi VerifiedIdentity, origin string, now time.Time) (signupT1Result, error) {
	var res signupT1Result
	err := s.inTxAudited(ctx, func(tx pgx.Tx) ([]audit.Event, error) {
		if _, err := tx.Exec(ctx, `
INSERT INTO app.identity_provider_binding (provider, issuer, client_id, bound_at) VALUES ('workos', $1, $2, $3)
ON CONFLICT (provider) DO NOTHING`, vi.Issuer, vi.ClientID, now); err != nil {
			return nil, unavailable(err)
		}
		var boundIssuer, boundClient string
		if err := tx.QueryRow(ctx, `SELECT issuer, client_id FROM app.identity_provider_binding WHERE provider = 'workos'`).
			Scan(&boundIssuer, &boundClient); err != nil {
			return nil, unavailable(err)
		}
		if boundIssuer != vi.Issuer || boundClient != vi.ClientID {
			return nil, ErrNamespaceMismatch
		}
		mapped, err := principalSubject(ctx, tx, vi.Issuer, vi.Subject)
		if err != nil {
			return nil, err
		}
		if mapped != "" {
			if err := initIdentity(ctx, tx, mapped); err != nil {
				return nil, err
			}
		} else if mapped, err = s.createMapping(ctx, tx, vi, now); err != nil {
			return nil, err
		}
		res.Subject = mapped
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM app.users WHERE id = $1 FOR SHARE`, mapped).Scan(&status); err != nil {
			return nil, unavailable(err)
		}
		if status != "active" {
			return nil, ErrUserDisabled
		}
		if vi.ProviderSessionID == "" || len(vi.ProviderSessionID) > 255 {
			return nil, ErrUnavailable
		}
		id, err := newUserID()
		if err != nil {
			return nil, ErrUnavailable
		}
		cred, err := newSecret()
		if err != nil {
			return nil, ErrUnavailable
		}
		csrf, err := newSecret()
		if err != nil {
			return nil, ErrUnavailable
		}
		expires := now.Add(SignupLifetime)
		if _, err := tx.Exec(ctx, sqlInsertSignup, id, mapped, origin, hashSecret(cred), csrf, vi.ProviderSessionID, now, expires); err != nil {
			return nil, unavailable(err)
		}
		res.Credential, res.ExpiresAt = cred, expires
		return []audit.Event{signInEvent(mapped, audit.ReasonSignupAuthenticated)}, nil
	})
	switch {
	case errors.Is(err, ErrUserDisabled):
		return signupT1Result{Subject: res.Subject}, err
	case err != nil:
		return signupT1Result{}, err
	}
	return res, nil
}

// signupIdentityTx runs fn in one transaction in the IDENTITY shape of userID (a resolved signup credential's user,
// G062 Plan r5 §3.4); fn issues only signup statements (provisioning body, own-row prove).
func (s *Store) signupIdentityTx(ctx context.Context, userID string, fn func(tx pgx.Tx) error) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initIdentity(ctx, tx, userID); err != nil {
			return err
		}
		return fn(tx)
	})
}

// loginTx runs fn in one transaction in the LOGIN shape of bindingHash (G062 Plan r5 §3.4: the continuation challenge
// and completion bodies refuse any other shape).
func (s *Store) loginTx(ctx context.Context, bindingHash []byte, fn func(tx pgx.Tx) error) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initLogin(ctx, tx, bindingHash); err != nil {
			return err
		}
		return fn(tx)
	})
}

// disabledSubject keeps the mapped subject of an ErrUserDisabled refusal for its audit attribution.
func disabledSubject(res *signInResult, mapped string, err error) error {
	if errors.Is(err, ErrUserDisabled) {
		res.Subject = mapped
	}
	return err
}

// signInEvent is the auth.sign_in success event of the mapped subject (the IDENTITY context of the transaction).
func signInEvent(mapped string, reason audit.Reason) audit.Event {
	subject, _ := audit.ParseUUID(mapped)
	return audit.Event{Action: audit.AuthSignIn, Outcome: audit.Success, Reason: reason, Subject: subject}
}

// principalSubject is the E2 reader call in q: the user mapped to the exact (issuer, subject), or "" when unmapped.
func principalSubject(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, issuer, subject string) (string, error) {
	var mapped *string
	if err := q.QueryRow(ctx, sqlAuthPrincipalSubject, issuer, subject).Scan(&mapped); err != nil {
		return "", unavailable(err)
	}
	if mapped == nil {
		return "", nil
	}
	return *mapped, nil
}

// createMapping creates the user (with a core-generated UUID, decided Q3) and its exact principal mapping in the
// IDENTITY shape of that new subject, inside a savepoint. Concurrent first sign-ins for one principal race on the
// principal primary key: the loser rolls back its savepoint (removing its user row and reverting its context to
// empty), reads the winner's mapping through E2 and initializes IDENTITY for the winner, so exactly one user and one
// principal exist and every session belongs to the winner.
func (s *Store) createMapping(ctx context.Context, tx pgx.Tx, vi VerifiedIdentity, now time.Time) (string, error) {
	id, err := newUserID()
	if err != nil {
		return "", ErrUnavailable
	}
	sp, err := tx.Begin(ctx)
	if err != nil {
		return "", unavailable(err)
	}
	if err := initIdentity(ctx, sp, id); err != nil {
		_ = sp.Rollback(ctx)
		return "", err
	}
	var userID, mapped string
	err = sp.QueryRow(ctx, `INSERT INTO app.users (id, display_name, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id::text`,
		id, nullable(vi.DisplayName), now).Scan(&userID)
	if err != nil {
		_ = sp.Rollback(ctx)
		return "", unavailable(err)
	}
	err = sp.QueryRow(ctx, `
INSERT INTO app.identity_principals (issuer, subject, user_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $4)
ON CONFLICT (issuer, subject) DO NOTHING RETURNING user_id::text`, vi.Issuer, vi.Subject, userID, now).Scan(&mapped)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		if err := sp.Rollback(ctx); err != nil {
			return "", unavailable(err)
		}
		if beforeWinnerInit != nil {
			beforeWinnerInit(ctx, tx)
		}
		if mapped, err = principalSubject(ctx, tx, vi.Issuer, vi.Subject); err != nil {
			return "", err
		}
		if mapped == "" {
			return "", ErrUnavailable
		}
		if err := initIdentity(ctx, tx, mapped); err != nil {
			return "", err
		}
		return mapped, nil
	case err != nil:
		_ = sp.Rollback(ctx)
		return "", unavailable(err)
	}
	if err := sp.Commit(ctx); err != nil {
		return "", unavailable(err)
	}
	return mapped, nil
}

// beforeWinnerInit is a test hook (nil in production): it observes a losing first sign-in's transaction after its
// savepoint rollback and before the winner's context is initialized (I-MAP-CONC).
var beforeWinnerInit func(ctx context.Context, tx pgx.Tx)

// issue checks the user is active (FOR SHARE, serialized with DisableUser), refreshes the untrusted display name and
// inserts a new session with a fresh credential and CSRF token. It runs in the IDENTITY shape of userID.
func (s *Store) issue(ctx context.Context, tx pgx.Tx, userID string, p signInParams) (signInResult, error) {
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM app.users WHERE id = $1 FOR SHARE`, userID).Scan(&status); err != nil {
		return signInResult{}, unavailable(err)
	}
	if status != "active" {
		return signInResult{}, ErrUserDisabled
	}
	if _, err := tx.Exec(ctx, `
UPDATE app.users SET display_name = $2, updated_at = $3 WHERE id = $1 AND display_name IS DISTINCT FROM $2`,
		userID, nullable(p.Identity.DisplayName), p.Now); err != nil {
		return signInResult{}, unavailable(err)
	}
	cred, err := newSecret()
	if err != nil {
		return signInResult{}, ErrUnavailable
	}
	csrf, err := newSecret()
	if err != nil {
		return signInResult{}, ErrUnavailable
	}
	absolute := p.Now.Add(p.Max)
	idle := p.Now.Add(p.Idle)
	if idle.After(absolute) {
		idle = absolute
	}
	// G016 security amendment: the session is bound to its issuing exact origin and the Organization pinned on the login
	// (the sessions_insert policy refuses a browser row without both; issuanceGuard already refused an empty one).
	if _, err := tx.Exec(ctx, `
INSERT INTO app.sessions (user_id, credential_hash, csrf_token, provider_session_id, auth_time, created_at, renewed_at,
                          idle_expires_at, absolute_expires_at, origin, organization_id)
VALUES ($1, $2, $3, $4, $5, $6, $6, $7, $8, $9, $10)`,
		userID, hashSecret(cred), csrf, nullable(p.Identity.ProviderSessionID), p.Identity.AuthTime, p.Now, idle, absolute,
		p.Origin, p.Organization); err != nil {
		return signInResult{}, unavailable(err)
	}
	return signInResult{Credential: cred, ExpiresAt: absolute}, nil
}

// lookupUser reads the user mapped to (issuer, subject) through the E2 reader, read-only.
func (s *Store) lookupUser(ctx context.Context, issuer, subject string) (string, bool, error) {
	id, err := principalSubject(ctx, s.db, issuer, subject)
	if err != nil {
		return "", false, err
	}
	return id, id != "", nil
}

// revokeSession revokes one session of userID (the subject F1 resolved); revoking an already revoked session changes
// nothing. G011 revocation priority (M5): the capture trigger records session.revoked in the same transaction, or, on
// an audit-only fault, sets the gap marker read here before COMMIT (afterRevocation). G058 (D-18 bound path): the same
// transaction revokes only the pending enrollment credentials minted from this session's provider session.
func (s *Store) revokeSession(ctx context.Context, userID, sessionID, reason string, now time.Time) error {
	var gap []string
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initIdentity(ctx, tx, userID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE app.sessions SET revoked_at = $2, revoke_reason = $3 WHERE id = $1 AND revoked_at IS NULL`,
			sessionID, now, reason); err != nil {
			return unavailable(err)
		}
		// G058 D-18 bound path: only the pending enrollment credentials minted from this session's provider session.
		if _, err := tx.Exec(ctx, sqlRevokeEnrollmentBound, userID, sessionID, now, reason); err != nil {
			return unavailable(err)
		}
		// G062: and the pending root signup credentials minted from it.
		if _, err := tx.Exec(ctx, sqlRevokeSignupBound, userID, sessionID, now, reason); err != nil {
			return unavailable(err)
		}
		var err error
		gap, err = revocationGap(ctx, tx)
		return err
	})
	return s.afterRevocation(ctx, userID, reason, audit.ResourceSession, sessionID, gap, err)
}

// revokeAll revokes every unrevoked session of one user in one statement (every revoked_at IS NULL row, expired ones
// included, with no count cap). A session committed after this statement's snapshot is not affected. G011: one
// session.revoked event per revoked session, or the gap path of afterRevocation. G058 (D-18 broader path): every pending
// enrollment credential of the user is revoked in the same transaction.
func (s *Store) revokeAll(ctx context.Context, userID, reason string, now time.Time) (int64, error) {
	var n int64
	var gap []string
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initIdentity(ctx, tx, userID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `UPDATE app.sessions SET revoked_at = $2, revoke_reason = $3 WHERE user_id = $1 AND revoked_at IS NULL`,
			userID, now, reason)
		if err != nil {
			return unavailable(err)
		}
		n = tag.RowsAffected()
		if err := revokeEnrollmentSessionsAll(ctx, tx, userID, reason, now); err != nil {
			return err
		}
		if err := revokeSignupSessionsAll(ctx, tx, userID, reason, now); err != nil {
			return err
		}
		gap, err = revocationGap(ctx, tx)
		return err
	})
	if err := s.afterRevocation(ctx, userID, reason, audit.ResourceUser, userID, gap, err); err != nil {
		return 0, err
	}
	return n, nil
}

// recordProviderLogout stores the provider revocation outcome and the hashed one-use logout reference on a session of
// userID already revoked with reason provider_logout.
func (s *Store) recordProviderLogout(ctx context.Context, userID, sessionID, outcome string, refHash []byte, refExpires time.Time) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initIdentity(ctx, tx, userID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
UPDATE app.sessions SET provider_revocation = $2, provider_logout_ref_hash = $3, provider_logout_ref_expires_at = $4
WHERE id = $1 AND revoke_reason = 'provider_logout' AND provider_logout_ref_hash IS NULL`, sessionID, outcome, refHash, refExpires)
		if err != nil {
			return unavailable(err)
		}
		if tag.RowsAffected() != 1 {
			return ErrUnavailable
		}
		return nil
	})
}

// consumeLogoutRef (F6) returns the recorded provider outcome for a logout reference exactly once: the E3 reader finds
// the session and its subject, then an IDENTITY transaction marks the reference returned.
func (s *Store) consumeLogoutRef(ctx context.Context, refHash []byte, now time.Time) (string, bool, error) {
	var sessionID, userID string
	err := s.db.QueryRow(ctx, sqlAuthLogoutRef, refHash, now).Scan(&sessionID, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, unavailable(err)
	}
	var outcome *string
	found := false
	err = s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initIdentity(ctx, tx, userID); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `
UPDATE app.sessions SET provider_logout_returned_at = $2
WHERE id = $3 AND provider_logout_ref_hash = $1 AND provider_logout_returned_at IS NULL AND provider_logout_ref_expires_at > $2
RETURNING provider_revocation`, refHash, now, sessionID).Scan(&outcome)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return unavailable(err)
		}
		found = true
		return nil
	})
	if err != nil {
		return "", false, err
	}
	if !found || outcome == nil {
		return "", false, nil
	}
	return *outcome, true, nil
}

// DisableUser disables a user and revokes all of its sessions in one transaction (trusted; no route), in the IDENTITY
// shape of the target the trusted in-process caller names (E7). FOR UPDATE serializes with sign-in's FOR SHARE, so no
// active session can exist for a disabled user. G011 revocation priority (M5): user.disabled (system) and one
// session.revoked (system) per revoked session are recorded in the same transaction, or the gap path of afterRevocation
// runs after COMMIT. G058 (D-11): the same transaction revokes every pending enrollment credential of the user, after the
// sessions UPDATE and before the gap read, exactly as FinalizeRetirement does (revokeEnrollmentSessionsAll, 'disabled').
func (s *Store) DisableUser(ctx context.Context, userID string, now time.Time) error {
	if !canonicalUserID(userID) {
		return ErrInvalidRequest // rejected before any database work; the context parsers would refuse it as malformed
	}
	var gap []string
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
		if _, err := tx.Exec(ctx, `
UPDATE app.users SET status = 'disabled', disabled_at = COALESCE(disabled_at, $2), updated_at = $2 WHERE id = $1`, userID, now); err != nil {
			return unavailable(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE app.sessions SET revoked_at = $2, revoke_reason = 'disabled' WHERE user_id = $1 AND revoked_at IS NULL`,
			userID, now); err != nil {
			return unavailable(err)
		}
		if err := revokeEnrollmentSessionsAll(ctx, tx, userID, revokeDisabled, now); err != nil {
			return err
		}
		if err := revokeSignupSessionsAll(ctx, tx, userID, revokeDisabled, now); err != nil {
			return err
		}
		var err error
		gap, err = revocationGap(ctx, tx)
		return err
	})
	return s.afterRevocation(ctx, userID, revokeDisabled, audit.ResourceUser, userID, gap, err)
}

// EnableUser re-enables a user (trusted; no route; E7). Old sessions and enrollment credentials stay revoked; only new
// sign-ins create sessions.
func (s *Store) EnableUser(ctx context.Context, userID string, now time.Time) error {
	if !canonicalUserID(userID) {
		return ErrInvalidRequest
	}
	return s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initIdentity(ctx, tx, userID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `UPDATE app.users SET status = 'active', disabled_at = NULL, updated_at = $2 WHERE id = $1`, userID, now)
		if err != nil {
			return unavailable(err)
		}
		if tag.RowsAffected() != 1 {
			return ErrInvalidRequest
		}
		return nil
	})
}

// zeroUUID names no user: core generates version 4 UUIDs, and the context parsers refuse the zero UUID.
const zeroUUID = "00000000-0000-0000-0000-000000000000"

// canonicalUserID reports whether s has the only form app.ctx_user() accepts: a nonzero, lowercase, hyphenated UUID
// (8-4-4-4-12 hexadecimal digits). Anything else would fail the context parser with 22P02 or 22023, which the store
// would report as identity_unavailable rather than invalid_request.
func canonicalUserID(s string) bool {
	if len(s) != 36 || s == zeroUUID {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return false
			}
		}
	}
	return true
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
