//go:build integration

// G058 D-11 atomic rollback through the existing internal fault seam (readiness correction 01 §3; Codex disposition F2):
// the store under test is newStore over faultDB, which wraps the constrained runtime pool of rls_pool_integration_test.go
// and fails the added enrollment_sessions UPDATE before it is sent. Fixtures and every assertion read use the attested
// superuser of that environment; nothing in production is exported or changed for this test.
package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// seedEnrollmentRow inserts one pending (unrevoked, unexpired, unadmitted) enrollment credential of user, minted from
// provider session sid in the fixture Organization, through the attested superuser, and returns its credential hash.
func (e *rlsPoolEnv) seedEnrollmentRow(t *testing.T, user, sid string) []byte {
	t.Helper()
	secret, csrf := make([]byte, 32), make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(csrf); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(secret)
	if _, err := e.setup.Exec(context.Background(), `INSERT INTO app.enrollment_sessions
  (id, user_id, organization_id, origin, credential_hash, csrf_token, provider_session_id, expires_at)
VALUES (gen_random_uuid(), $1::text::uuid, $2::text::uuid, $3, $4, $5, $6, now() + interval '20 minutes')`,
		user, e.org, rlsOrigin, sum[:], csrf, sid); err != nil {
		t.Fatalf("seed enrollment row: %v", err)
	}
	return sum[:]
}

// TestEnrollmentRevocationFaultRollback (G058 D-11): with the enrollment_sessions UPDATE failing, FinalizeRetirement and
// DisableUser are identity_unavailable and write nothing: the users row stays active, the user's session and enrollment
// rows stay unrevoked and no audit event is appended (users, sessions, enrollment_sessions and audit.events are
// byte-identical). The same call without the fault then succeeds and revokes the enrollment row ('disabled'), so the
// injected statement is the only obstacle.
func TestEnrollmentRevocationFaultRollback(t *testing.T) {
	e := newRLSPoolEnv(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	pass := func(context.Context, pgx.Tx) error { return nil }
	for _, c := range []struct {
		name, sub string
		call      func(uid string) error
	}{
		{"retire", "user_01ENRFR", func(uid string) error { return e.store.FinalizeRetirement(ctx, uid, now, pass) }},
		{"disable", "user_01ENRFD", func(uid string) error { return e.store.DisableUser(ctx, uid, now) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			e.provision(t, c.sub, true)
			_, uid := e.signIn(t, c.sub, now)
			var sid string
			if err := e.setup.QueryRow(ctx, `SELECT provider_session_id FROM app.sessions WHERE user_id = $1::text::uuid`, uid).Scan(&sid); err != nil {
				t.Fatalf("fixture session: %v", err)
			}
			cred := e.seedEnrollmentRow(t, uid, sid)
			state := func() string {
				return e.hash(t, "users") + "|" + e.hash(t, "sessions") + "|" + e.hash(t, "enrollment_sessions") + "|" +
					e.hash(t, "role_bindings") + "|" + strconv.Itoa(e.count(t, `SELECT count(*)::int FROM audit.events`))
			}
			before := state()

			e.db.failSQL = "UPDATE app.enrollment_sessions"
			err := c.call(uid)
			e.db.reset()
			if err != ErrUnavailable {
				t.Fatalf("%s with a failing enrollment UPDATE: %v, want identity_unavailable", c.name, err)
			}
			if after := state(); after != before {
				t.Fatalf("partial writes after the injected failure: %s -> %s", before, after)
			}
			if n := e.count(t, `SELECT count(*)::int FROM app.users WHERE id = $1::text::uuid AND status = 'active'`, uid); n != 1 {
				t.Fatal("the user is not active after the rolled-back finalization")
			}
			if n := e.count(t, `SELECT count(*)::int FROM app.sessions WHERE user_id = $1::text::uuid AND revoked_at IS NULL`, uid); n != 1 {
				t.Fatalf("unrevoked sessions %d, want 1", n)
			}
			if n := e.count(t, `SELECT count(*)::int FROM app.enrollment_sessions WHERE credential_hash = $1 AND revoked_at IS NULL`, cred); n != 1 {
				t.Fatal("the enrollment row was revoked by the rolled-back transaction")
			}
			if n := e.count(t, `SELECT count(*)::int FROM audit.events WHERE action IN ('user.disabled', 'session.revoked', 'revocation.unaudited')
  AND subject_user_id = $1::text::uuid`, uid); n != 0 {
				t.Fatalf("%d user.disabled/session.revoked/revocation.unaudited events after the rollback", n)
			}

			// Control: without the fault the same call commits and revokes the enrollment row with reason disabled.
			if err := c.call(uid); err != nil {
				t.Fatalf("%s without the fault: %v", c.name, err)
			}
			if n := e.count(t, `SELECT count(*)::int FROM app.enrollment_sessions WHERE credential_hash = $1 AND revoke_reason = 'disabled'`, cred); n != 1 {
				t.Fatal("the clean call did not revoke the enrollment row")
			}
			if n := e.count(t, `SELECT count(*)::int FROM app.users WHERE id = $1::text::uuid AND status = 'disabled'`, uid); n != 1 {
				t.Fatal("the clean call did not disable the user")
			}
		})
	}
}
