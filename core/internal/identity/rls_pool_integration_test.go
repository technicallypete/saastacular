//go:build integration

// B-RLS-IT, identity part (G008 Plan r3 §8.3): RLS-POOL I1-I5, RLS-AUTH E5 (f) and the I-MAP-CONC extension, through the
// identity store's own database boundary (storeDB) on a runtime pool constrained to one observed backend. Fault
// injection wraps the real pool; no statement is simulated except the injected failure itself. Setup-side counts and
// row hashes use the separately attested superuser (DB_SUPERUSER, DB_SUPERUSER_PASSWORD); assertions run as the runtime
// role. The database is migrated by TestMain (integration_test.go).
//
// G016 (amendment 02): store-level sign-ins pass the production final issuance guard. The fixture gives them a fixed
// test HostResolver over the real app.hostname_state reader (the Organization with platform label org-a and its
// origin) and provisions each subject's user, principal and Organization membership through the attested superuser
// before it signs in, as an operator does under D-2(a). TestRLSFirstSignInRace is the first-sign-in negative, and
// TestIssuanceGuardFailsClosed shows that the guard itself refuses a missing resolver or Organization with no change.
package identity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/migrate"
)

// --- environment -------------------------------------------------------------------------------------------------------

type rlsPoolEnv struct {
	cfg   migrate.Config
	pool  *pgxpool.Pool // runtime, MaxConns 1
	setup *pgx.Conn     // attested superuser
	db    *faultDB
	store *Store
	org   string // the fixture Organization (platform label org-a) that rlsOrigin routes to
}

// The fixture: rlsLabel is the fixture Organization's fixed synthetic platform label, the host of rlsOrigin (G057 D-2);
// rlsSlug is its friendly slug, metadata that never routes.
const (
	rlsIssuer = "https://issuer.rls.invalid"
	rlsOrigin = "https://org-a.localhost:543"
	rlsLabel  = "org-a"
	rlsSlug   = "alpha-agency"
	rlsOrgID  = "0f16a000-0000-4000-8000-0000000000a1" // the identity harness's org-a (seedTenants)
)

// rlsHosts is the fixed test HostResolver of these store-level cases: rlsOrigin is the platform host of label org-a,
// resolved by one statement on the querier the store passes (the issuance transaction) through the real
// app.hostname_state reader; every other origin routes nowhere.
type rlsHosts struct{}

func (rlsHosts) ResolveOrigin(ctx context.Context, q Querier, origin string) (HostResolution, error) {
	if origin != rlsOrigin {
		return HostResolution{}, nil
	}
	var org *string
	var routing bool
	err := q.QueryRow(ctx, `SELECT organization_id::text, routing FROM app.hostname_state($1, $2)`, HostPlatform, rlsLabel).Scan(&org, &routing)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return HostResolution{}, nil
	case err != nil:
		return HostResolution{}, err
	case org == nil || !routing:
		return HostResolution{}, nil
	}
	return HostResolution{Kind: HostPlatform, OrganizationID: *org, Routing: true}, nil
}

func newRLSPoolEnv(t *testing.T) *rlsPoolEnv {
	t.Helper()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil {
		t.Fatalf("integration settings: %v", err)
	}
	user, pw := os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD")
	if user == "" || pw == "" || os.Getenv("DB_RUNTIME_PASSWORD") == "" {
		t.Fatal("integration settings: DB_RUNTIME_PASSWORD, DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required")
	}
	e := &rlsPoolEnv{cfg: cfg}
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = cfg.Host, uint16(cfg.Port), cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = cfg.RuntimeRole(), os.Getenv("DB_RUNTIME_PASSWORD")
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	pc.MaxConns, pc.MinConns = 1, 0
	if e.pool, err = pgxpool.NewWithConfig(context.Background(), pc); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.pool.Close)
	cc, err := pgx.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	cc.Host, cc.Port, cc.Database, cc.User, cc.Password = cfg.Host, uint16(cfg.Port), cfg.Database, user, pw
	cc.TLSConfig, cc.Fallbacks = nil, nil
	if e.setup, err = pgx.ConnectConfig(context.Background(), cc); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.setup.Close(context.Background()) })
	var attest string
	if err := e.setup.QueryRow(context.Background(), `SELECT rolsuper::text FROM pg_roles WHERE rolname = current_user`).Scan(&attest); err != nil || attest != "true" {
		t.Fatalf("setup connection is not a superuser (%v)", err)
	}
	if err := e.pool.QueryRow(context.Background(), `SELECT format('%s|%s|%s|%s', session_user, current_user, r.rolsuper, r.rolbypassrls)
FROM pg_roles r WHERE r.rolname = current_user`).Scan(&attest); err != nil || attest != cfg.RuntimeRole()+"|"+cfg.RuntimeRole()+"|f|f" {
		t.Fatalf("runtime attestation: %q (%v)", attest, err)
	}
	e.db = &faultDB{pool: e.pool}
	e.store = newStore(e.db)
	if _, err := e.setup.Exec(context.Background(),
		`TRUNCATE app.login_transactions, app.sessions, app.identity_principals, app.users, app.identity_provider_binding CASCADE`); err != nil {
		t.Fatal(err)
	}
	// G016, G057: the labelled fixture Organization (idempotent; the identity harness seeds the same row).
	if _, err := e.setup.Exec(context.Background(), `INSERT INTO app.organizations (id, name, slug, platform_label) VALUES ($1::text::uuid,
  'Org A', $2, $3) ON CONFLICT DO NOTHING`, rlsOrgID, rlsSlug, rlsLabel); err != nil {
		t.Fatal(err)
	}
	if err := e.setup.QueryRow(context.Background(), `SELECT id::text FROM app.organizations WHERE platform_label = $1`, rlsLabel).Scan(&e.org); err != nil {
		t.Fatalf("fixture Organization: %v", err)
	}
	return e
}

// provision gives subject sub of rlsIssuer a user and a principal and, when member, a current membership of the fixture
// Organization, through the attested superuser (idempotent). It returns the user.
func (e *rlsPoolEnv) provision(t *testing.T, sub string, member bool) string {
	t.Helper()
	ctx := context.Background()
	var uid string
	err := e.setup.QueryRow(ctx, `SELECT user_id::text FROM app.identity_principals WHERE issuer = $1 AND subject = $2`, rlsIssuer, sub).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		err = e.setup.QueryRow(ctx, `WITH u AS (INSERT INTO app.users (display_name) VALUES ('Pool Fixture') RETURNING id),
  p AS (INSERT INTO app.identity_principals (issuer, subject, user_id) SELECT $1, $2, id FROM u)
SELECT id::text FROM u`, rlsIssuer, sub).Scan(&uid)
	}
	if err != nil {
		t.Fatalf("provision %s: %v", sub, err)
	}
	if member {
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
SELECT $1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership'
WHERE NOT EXISTS (SELECT 1 FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid
                    AND scope_type = 'organization' AND source = 'membership')`, e.org, uid); err != nil {
			t.Fatalf("provision %s membership: %v", sub, err)
		}
	}
	return uid
}

// backend returns the backend PID of the single pooled connection (a new one after a discard).
func (e *rlsPoolEnv) backend(t *testing.T) int {
	t.Helper()
	var pid int
	if err := e.pool.QueryRow(context.Background(), `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	return pid
}

// settings returns the four context settings as the pooled connection sees them outside a transaction.
func (e *rlsPoolEnv) settings(t *testing.T) string {
	t.Helper()
	var s string
	if err := e.pool.QueryRow(context.Background(), `SELECT concat_ws('|', coalesce(current_setting('app.user_id', true), ''),
  coalesce(current_setting('app.organization_id', true), ''), coalesce(current_setting('app.group_id', true), ''),
  coalesce(current_setting('app.login_binding', true), ''))`).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// poison sets a session-level context value on the single pooled connection (the deliberately unsafe control).
func (e *rlsPoolEnv) poison(t *testing.T, name, value string) {
	t.Helper()
	c, err := e.pool.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Release()
	if _, err := c.Exec(context.Background(), `SELECT set_config($1, $2, false)`, name, value); err != nil {
		t.Fatal(err)
	}
}

func (e *rlsPoolEnv) count(t *testing.T, q string, args ...any) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (e *rlsPoolEnv) hash(t *testing.T, table string) string {
	t.Helper()
	var h string
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) || ':' || coalesce(md5(string_agg(r::text, E'\n' ORDER BY r::text)), '-') FROM app.`+table+` r`).Scan(&h); err != nil {
		t.Fatal(err)
	}
	return h
}

func randomHex(t *testing.T, n int) string {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

var rlsSeq int

// signInParams are the parameters of a tenant-host sign-in on rlsOrigin (pinned to the fixture Organization, with the
// fixed test resolver), as the callback passes them to the store.
func (e *rlsPoolEnv) signInParams(sub string, now time.Time) signInParams {
	rlsSeq++
	return signInParams{Identity: VerifiedIdentity{Issuer: rlsIssuer, ClientID: "client_01RLSPOOL", Subject: sub,
		ProviderSessionID: fmt.Sprintf("session_rls%04d", rlsSeq), DisplayName: "Pool Fixture", AuthTime: now},
		Now: now, Idle: 30 * time.Minute, Max: 8 * time.Hour, Origin: rlsOrigin, Organization: e.org, Hosts: rlsHosts{}}
}

// sessionsOf counts the sessions of subject sub's user (setup side).
func (e *rlsPoolEnv) sessionsOf(t *testing.T, sub string) int {
	t.Helper()
	return e.count(t, `SELECT count(*)::int FROM app.sessions s JOIN app.identity_principals p ON p.user_id = s.user_id
WHERE p.issuer = $1 AND p.subject = $2`, rlsIssuer, sub)
}

// signIn issues a session for subject sub through the store (clean, no fault) and returns its credential hash and user.
func (e *rlsPoolEnv) signIn(t *testing.T, sub string, now time.Time) ([]byte, string) {
	t.Helper()
	res, err := e.store.signIn(context.Background(), e.signInParams(sub, now))
	if err != nil || res.Credential == nil {
		t.Fatalf("sign-in: %v", err)
	}
	info, err := e.store.resolve(context.Background(), hashSecret(res.Credential), rlsOrigin, now, 30*time.Minute)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	return hashSecret(res.Credential), info.UserID
}

// --- fault injection ---------------------------------------------------------------------------------------------------

var errInjected = &pgconn.PgError{Code: "XX000", Severity: "ERROR", SeverityUnlocalized: "ERROR", Message: "injected statement failure"}

type faultDB struct {
	pool *pgxpool.Pool

	mu           sync.Mutex
	failSQL      string             // a statement containing this text fails before it is sent
	cancelSQL    string             // the request context is canceled just before a statement containing this text
	cancel       context.CancelFunc //
	lostCommit   bool               // COMMIT really commits, then its reply is "lost"
	failRollback bool               // ROLLBACK is not sent and reports a failure
}

func (d *faultDB) reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.failSQL, d.cancelSQL, d.cancel, d.lostCommit, d.failRollback = "", "", nil, false, false
}

func (d *faultDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, err := d.pool.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}
	return &faultTx{Tx: tx, d: d}, nil
}

func (d *faultDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return d.pool.QueryRow(ctx, sql, args...)
}

type faultTx struct {
	pgx.Tx
	d *faultDB
}

type errRow struct{ err error }

func (r errRow) Scan(...any) error { return r.err }

func (t *faultTx) enter(sql string) error {
	t.d.mu.Lock()
	defer t.d.mu.Unlock()
	if t.d.cancelSQL != "" && strings.Contains(sql, t.d.cancelSQL) && t.d.cancel != nil {
		t.d.cancel()
	}
	if t.d.failSQL != "" && strings.Contains(sql, t.d.failSQL) {
		return errInjected
	}
	return nil
}

func (t *faultTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if err := t.enter(sql); err != nil {
		return pgconn.CommandTag{}, err
	}
	return t.Tx.Exec(ctx, sql, args...)
}

func (t *faultTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if err := t.enter(sql); err != nil {
		return errRow{err}
	}
	return t.Tx.QueryRow(ctx, sql, args...)
}

// Commit with lostCommit sends COMMIT, which the server executes, then loses the connection before the reply is
// read: the connection is closed while still held, the transaction releases it, and the caller sees a lost reply.
func (t *faultTx) Commit(ctx context.Context) error {
	t.d.mu.Lock()
	lost := t.d.lostCommit
	t.d.mu.Unlock()
	if !lost {
		return t.Tx.Commit(ctx)
	}
	c := t.Tx.Conn()
	if _, err := c.Exec(ctx, "COMMIT"); err != nil {
		return err
	}
	_ = c.Close(context.Background())
	_ = t.Tx.Rollback(ctx) // releases the closed connection; the pool destroys it
	return io.ErrUnexpectedEOF
}

// Rollback with failRollback loses the connection while the rollback is attempted: the rollback fails and the pool
// destroys the connection on release.
func (t *faultTx) Rollback(ctx context.Context) error {
	t.d.mu.Lock()
	fail := t.d.failRollback
	t.d.mu.Unlock()
	if !fail {
		return t.Tx.Rollback(ctx)
	}
	_ = t.Tx.Conn().Close(context.Background())
	if err := t.Tx.Rollback(ctx); err != nil {
		return err
	}
	return errors.New("injected rollback failure")
}

// --- tests -------------------------------------------------------------------------------------------------------------

func TestRLSIdentityPool(t *testing.T) {
	e := newRLSPoolEnv(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	empty := "|||"
	// G016 D-2(a): every subject signing in below is a provisioned member of the fixture Organization.
	for _, sub := range []string{"user_01RLSI1", "user_01RLSI2", "user_01RLSI2D", "user_01RLSI3", "user_01RLSI3D", "user_01RLSI4",
		"user_01RLSI4X", "user_01RLSI5"} {
		e.provision(t, sub, true)
	}

	t.Run("I1 statement error in sign-in and renewal: nothing issued, clean reuse", func(t *testing.T) {
		pid := e.backend(t)
		users, sessions := e.count(t, `SELECT count(*)::int FROM app.users`), e.count(t, `SELECT count(*)::int FROM app.sessions`)
		e.db.failSQL = "INSERT INTO app.sessions"
		_, err := e.store.signIn(ctx, e.signInParams("user_01RLSI1", now))
		e.db.reset()
		if err != ErrUnavailable {
			t.Fatalf("sign-in with a failing statement: %v", err)
		}
		if u, s := e.count(t, `SELECT count(*)::int FROM app.users`), e.count(t, `SELECT count(*)::int FROM app.sessions`); u != users || s != sessions {
			t.Fatalf("rows changed: users %d->%d sessions %d->%d", users, u, sessions, s)
		}
		if got := e.backend(t); got != pid || e.settings(t) != empty {
			t.Fatalf("not a clean reuse: backend %d->%d settings %q", pid, got, e.settings(t))
		}
		cred, _ := e.signIn(t, "user_01RLSI1", now)
		later := now.Add(2 * time.Minute)
		var renewed time.Time
		_ = e.setup.QueryRow(ctx, `SELECT renewed_at FROM app.sessions WHERE credential_hash = $1`, cred).Scan(&renewed)
		e.db.failSQL = "SET idle_expires_at"
		_, err = e.store.resolve(ctx, cred, rlsOrigin, later, 30*time.Minute)
		e.db.reset()
		if err != ErrUnavailable {
			t.Fatalf("renewal with a failing statement: %v", err)
		}
		var after time.Time
		_ = e.setup.QueryRow(ctx, `SELECT renewed_at FROM app.sessions WHERE credential_hash = $1`, cred).Scan(&after)
		if !after.Equal(renewed) {
			t.Fatal("renewal persisted despite the failure")
		}
		if got := e.backend(t); got != pid || e.settings(t) != empty {
			t.Fatalf("not a clean reuse after renewal failure: backend %d->%d", pid, got)
		}
	})

	t.Run("I2 cancellation in sign-in, consume and disable: no success; discard or clean reuse", func(t *testing.T) {
		for _, c := range []struct {
			name, at string
			call     func(context.Context) error
		}{
			{"sign-in", "INSERT INTO app.sessions", func(ctx context.Context) error {
				_, err := e.store.signIn(ctx, e.signInParams("user_01RLSI2", now))
				return err
			}},
			{"consume", "DELETE FROM app.login_transactions\nWHERE", func(ctx context.Context) error {
				b := []byte(randomHex(t, 16))
				bh := hashSecret(b)
				if err := e.store.insertTransaction(context.Background(), loginTxn{StateHash: hashSecret([]byte("s" + string(b))), BindingHash: bh,
					Verifier: strings.Repeat("v", 43), ClientID: "client_01RLSPOOL", Origin: "https://o", ReturnPath: "/", CreatedAt: now,
					ExpiresAt: now.Add(10 * time.Minute)}); err != nil {
					t.Fatalf("start: %v", err)
				}
				_, res, err := e.store.consumeTransaction(ctx, hashSecret([]byte("s"+string(b))), bh, "client_01RLSPOOL", "https://o", now)
				if err == nil && res == consumeOK {
					return nil
				}
				if err == nil {
					return errors.New("not consumed")
				}
				return err
			}},
			{"disable", "UPDATE app.users SET status", func(ctx context.Context) error {
				_, uid := e.signIn(t, "user_01RLSI2D", now)
				return e.store.DisableUser(ctx, uid, now)
			}},
		} {
			pid := e.backend(t)
			cctx, cancel := context.WithCancel(ctx)
			e.db.cancelSQL, e.db.cancel = c.at, cancel
			err := c.call(cctx)
			e.db.reset()
			cancel()
			if err == nil {
				t.Errorf("%s: canceled call succeeded", c.name)
			}
			got := e.backend(t)
			outcome := "clean reuse"
			if got != pid {
				outcome = "discarded"
			} else if e.settings(t) != empty {
				t.Errorf("%s: reused connection carries settings", c.name)
			}
			fmt.Printf("RLS-POOL {\"case\":\"I2\",\"call\":%q,\"outcome\":%q}\n", c.name, outcome)
		}
	})

	t.Run("I3 lost COMMIT reply in sign-in and disable: identity_unavailable, possibly committed, next backend new", func(t *testing.T) {
		pid := e.backend(t)
		e.db.lostCommit = true
		_, err := e.store.signIn(ctx, e.signInParams("user_01RLSI3", now))
		e.db.reset()
		if err != ErrUnavailable {
			t.Fatalf("sign-in with a lost COMMIT reply: %v", err)
		}
		committed := e.sessionsOf(t, "user_01RLSI3") // the principal is provisioned: the session is the committed effect
		if got := e.backend(t); got == pid {
			t.Fatal("the connection of an unknown COMMIT outcome was reused")
		}
		_, uid := e.signIn(t, "user_01RLSI3D", now)
		pid = e.backend(t)
		e.db.lostCommit = true
		err = e.store.DisableUser(ctx, uid, now)
		e.db.reset()
		if err != ErrUnavailable {
			t.Fatalf("disable with a lost COMMIT reply: %v", err)
		}
		disabled := e.count(t, `SELECT count(*)::int FROM app.users WHERE id = $1 AND status = 'disabled'`, uid)
		if got := e.backend(t); got == pid {
			t.Fatal("the connection of an unknown COMMIT outcome was reused")
		}
		fmt.Printf("RLS-POOL {\"case\":\"I3\",\"sign_in_committed\":%d,\"disable_committed\":%d}\n", committed, disabled)
	})

	t.Run("I4 poisoned connection: E1 unaffected and not discarded; the next initialized transaction refuses and discards", func(t *testing.T) {
		cred, uid := e.signIn(t, "user_01RLSI4", now)
		_, other := e.signIn(t, "user_01RLSI4X", now)
		e.poison(t, "app.user_id", other)
		pid := e.backend(t)
		info, err := e.store.resolve(ctx, cred, rlsOrigin, now.Add(10*time.Second), 30*time.Minute)
		if err != nil || info.UserID != uid {
			t.Fatalf("non-renewing E1 on a poisoned connection: %v (own row: %v)", err, info.UserID == uid)
		}
		if got := e.backend(t); got != pid {
			t.Fatal("the E1 autocommit path discarded the connection (it runs no guard)")
		}
		_, err = e.store.resolve(ctx, cred, rlsOrigin, now.Add(5*time.Minute), 30*time.Minute)
		if err != ErrUnavailable {
			t.Fatalf("renewal on a poisoned connection: %v, want identity_unavailable", err)
		}
		if got := e.backend(t); got == pid {
			t.Fatal("the poisoned connection was not discarded")
		}
		if e.settings(t) != empty {
			t.Fatal("the new connection carries settings")
		}
		for _, f := range []func() error{
			func() error { return e.store.revokeSession(ctx, uid, info.SessionID, revokeLogout, now) },
			func() error { _, err := e.store.signIn(ctx, e.signInParams("user_01RLSI4", now)); return err },
		} {
			e.poison(t, "app.user_id", other)
			pid := e.backend(t)
			if err := f(); err != ErrUnavailable {
				t.Errorf("initialized transaction on a poisoned connection: %v", err)
			}
			if e.backend(t) == pid {
				t.Error("not discarded")
			}
		}
	})

	t.Run("E5 (f) production login start on a poisoned connection deletes nothing; recovery is clean", func(t *testing.T) {
		svc := NewService(e.store, Config{Provider: rlsStubProvider{}, Issuer: rlsIssuer, ClientID: "client_01RLSPOOL",
			Origins: NewOrigins([]string{"https://org-a.localhost:543"}), Idle: 30 * time.Minute, Max: 8 * time.Hour,
			Now: func() time.Time { return now }, Hosts: rlsHosts{}})
		if _, err := e.setup.Exec(ctx, `TRUNCATE app.login_transactions`); err != nil {
			t.Fatal(err)
		}
		seed := func(binding []byte, expired bool) {
			t.Helper()
			created := "now()"
			if expired {
				created = "now() - interval '20 minutes'"
			}
			if _, err := e.setup.Exec(ctx, `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin,
  return_path, created_at, expires_at) VALUES (sha256(convert_to($1, 'UTF8')), $2, repeat('v', 43), 'client_01RLSPOOL', 'https://o', '/', `+
				created+`, `+created+` + interval '10 minutes')`, randomHex(t, 16), binding); err != nil {
				t.Fatal(err)
			}
		}
		b := hashSecret([]byte(randomHex(t, 16)))
		seed(b, false) // the previous browser's pending, unexpired login of binding b
		seed(hashSecret([]byte(randomHex(t, 16))), true)
		seed(hashSecret([]byte(randomHex(t, 16))), true)
		seed(hashSecret([]byte(randomHex(t, 16))), false)
		before := e.hash(t, "login_transactions")
		for _, poison := range []struct{ name, value string }{
			{"app.login_binding", hex.EncodeToString(b)},
			{"app.user_id", "6f6e6c79-7465-4173-a974-6973706f6973"},
			{"app.login_binding", "not-hex"},
		} {
			e.poison(t, poison.name, poison.value)
			pid := e.backend(t)
			_, err := svc.Login(ctx, LoginInput{Origin: "https://org-a.localhost:543"})
			if err != ErrUnavailable {
				t.Errorf("login start on a connection poisoned with %s: %v, want identity_unavailable", poison.name, err)
			}
			if got := e.hash(t, "login_transactions"); got != before {
				t.Errorf("login start on a connection poisoned with %s changed login transactions", poison.name)
			}
			if e.backend(t) == pid {
				t.Errorf("the connection poisoned with %s was not discarded", poison.name)
			}
		}
		out, err := svc.Login(ctx, LoginInput{Origin: "https://org-a.localhost:543"})
		if err != nil {
			t.Fatalf("clean login start: %v", err)
		}
		if n := e.count(t, `SELECT count(*)::int FROM app.login_transactions WHERE expires_at <= now()`); n != 0 {
			t.Errorf("%d expired transactions survived the clean cleanup", n)
		}
		if n := e.count(t, `SELECT count(*)::int FROM app.login_transactions WHERE binding_hash = $1`, b); n != 1 {
			t.Errorf("binding b's unexpired transaction: %d rows, want 1", n)
		}
		if n := e.count(t, `SELECT count(*)::int FROM app.login_transactions`); n != 3 {
			t.Errorf("%d transactions after the clean start, want 3 (two unexpired fixtures and the new one)", n)
		}
		raw, ok := decodeSecret(out.Binding)
		if !ok {
			t.Fatal("binding")
		}
		state := strings.TrimPrefix(out.AuthorizationURL, "https://auth.rls.invalid/?state=")
		_, res, err := e.store.consumeTransaction(ctx, hashEncoded(state), hashSecret(raw), "client_01RLSPOOL", "https://org-a.localhost:543", now)
		if err != nil || res != consumeOK {
			t.Fatalf("consume of the new transaction: %v %v", res, err)
		}
		if _, res, _ := e.store.consumeTransaction(ctx, hashEncoded(state), hashSecret(raw), "client_01RLSPOOL", "https://org-a.localhost:543", now); res != consumeNone {
			t.Fatal("second consume succeeded")
		}
	})

	t.Run("I5 a rollback failure closes the connection", func(t *testing.T) {
		pid := e.backend(t)
		e.db.failSQL, e.db.failRollback = "INSERT INTO app.sessions", true
		_, err := e.store.signIn(ctx, e.signInParams("user_01RLSI5", now))
		e.db.reset()
		if err != ErrUnavailable {
			t.Fatalf("sign-in: %v", err)
		}
		if e.backend(t) == pid {
			t.Fatal("the connection of a failed rollback was reused")
		}
		if e.sessionsOf(t, "user_01RLSI5") != 0 {
			t.Fatal("rows committed")
		}
	})
}

// TestRLSFirstSignInRace is the I-MAP-CONC extension (decided Q3) revised by G016 D-2(a) (Plan r4 §1.7.1, amendment 02)
// into the first-sign-in negative: 32 concurrent first sign-ins of one unmapped principal on a routing tenant host are
// each refused organization_not_member by the final issuance guard (an unmapped principal is never a member); nothing
// is created (no user, principal, session or audit event), and no sign-in reaches the mapping-race path (the
// savepoint hook never runs).
func TestRLSFirstSignInRace(t *testing.T) {
	e := newRLSPoolEnv(t)
	ctx := context.Background()
	pc := e.pool.Config()
	pc.MaxConns = 40
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	store := NewStore(pool)
	var mu sync.Mutex
	hooked := 0
	beforeWinnerInit = func(context.Context, pgx.Tx) {
		mu.Lock()
		hooked++
		mu.Unlock()
	}
	t.Cleanup(func() { beforeWinnerInit = nil })
	now := time.Now().UTC().Truncate(time.Microsecond)
	// The namespace binding exists first, so no sign-in waits on another's uncommitted binding row.
	if _, err := e.setup.Exec(ctx, `INSERT INTO app.identity_provider_binding (provider, issuer, client_id) VALUES ('workos', $1, 'client_01RLSPOOL') ON CONFLICT DO NOTHING`, rlsIssuer); err != nil {
		t.Fatal(err)
	}
	before := e.hash(t, "users") + e.hash(t, "identity_principals") + e.hash(t, "sessions")
	events := e.count(t, `SELECT count(*)::int FROM audit.events`)
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make([]error, 32)
	for i := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := e.signInParams("user_01RLSRACE", now)
			p.Identity.ProviderSessionID = fmt.Sprintf("session_race%02d", i)
			<-start
			_, errs[i] = store.signIn(ctx, p)
		}()
	}
	close(start)
	wg.Wait()
	refused := 0
	for i, err := range errs {
		if err == ErrNotMember {
			refused++
		} else {
			t.Errorf("sign-in %d: %v, want organization_not_member", i, err)
		}
	}
	if got := e.hash(t, "users") + e.hash(t, "identity_principals") + e.hash(t, "sessions"); got != before {
		t.Fatal("a refused first sign-in created a user, principal or session")
	}
	if n := e.count(t, `SELECT count(*)::int FROM app.identity_principals WHERE subject = 'user_01RLSRACE'`); n != 0 {
		t.Fatalf("%d principals for the unmapped subject, want 0", n)
	}
	if n := e.count(t, `SELECT count(*)::int FROM audit.events`); n != events {
		t.Fatalf("audit events %d -> %d: a refused sign-in committed history", events, n)
	}
	if hooked != 0 {
		t.Fatalf("%d sign-ins reached the mapping-race path, want 0", hooked)
	}
	fmt.Printf("RLS-POOL {\"case\":\"I-MAP-CONC\",\"sign_ins\":32,\"refused_not_member\":%d}\n", refused)
}

// TestIssuanceGuardFailsClosed (G016 amendment 02): the final issuance guard inside the production signIn refuses a
// sign-in without a resolver or without a pinned Organization (host_changed: no store-level bypass), a host that routes
// nowhere or to another Organization (host_changed), a mapped non-member and an unmapped principal
// (organization_not_member), each with no user, principal, session, namespace-binding or audit change; the valid
// fixture then passes the same guard and issues exactly one session.
func TestIssuanceGuardFailsClosed(t *testing.T) {
	e := newRLSPoolEnv(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	e.provision(t, "user_01GUARDM", true)
	e.provision(t, "user_01GUARDN", false)
	snapshot := func() string {
		return e.hash(t, "users") + e.hash(t, "identity_principals") + e.hash(t, "sessions") + e.hash(t, "identity_provider_binding") +
			fmt.Sprint(e.count(t, `SELECT count(*)::int FROM audit.events`))
	}
	before := snapshot()
	for _, c := range []struct {
		name string
		edit func(*signInParams)
		want error
	}{
		{"missing resolver", func(p *signInParams) { p.Hosts = nil }, ErrHostChanged},
		{"missing Organization", func(p *signInParams) { p.Organization = "" }, ErrHostChanged},
		{"host routes nowhere", func(p *signInParams) { p.Origin = "https://nowhere.localhost:543" }, ErrHostChanged},
		{"host routes to another Organization", func(p *signInParams) { p.Organization = "0f16b000-0000-4000-8000-0000000000b1" }, ErrHostChanged},
		{"mapped non-member", func(p *signInParams) { p.Identity.Subject = "user_01GUARDN" }, ErrNotMember},
		{"unmapped principal", func(p *signInParams) { p.Identity.Subject = "user_01GUARDU" }, ErrNotMember},
	} {
		p := e.signInParams("user_01GUARDM", now)
		c.edit(&p)
		res, err := e.store.signIn(ctx, p)
		if err != c.want || res.Credential != nil {
			t.Errorf("%s: %v (credential issued: %v), want %v", c.name, err, res.Credential != nil, c.want)
		}
		if got := snapshot(); got != before {
			t.Errorf("%s: the refused sign-in changed users, principals, sessions, the namespace binding or audit history", c.name)
		}
		fmt.Printf("RLS-POOL {\"case\":\"GUARD\",\"refusal\":%q,\"outcome\":%q}\n", c.name, outcomeOf(err))
	}
	res, err := e.store.signIn(ctx, e.signInParams("user_01GUARDM", now))
	if err != nil || res.Credential == nil || res.Outcome != "signed_in" {
		t.Fatalf("valid fixture through the same guard: %v %q", err, res.Outcome)
	}
	if n := e.sessionsOf(t, "user_01GUARDM"); n != 1 {
		t.Fatalf("valid fixture sessions = %d, want 1", n)
	}
}

// rlsStubProvider is the minimal provider for login start (it is never asked to exchange or revoke).
type rlsStubProvider struct{}

func (rlsStubProvider) AuthorizationURL(_ context.Context, r AuthRequest) (string, error) {
	return "https://auth.rls.invalid/?state=" + r.State, nil
}
func (rlsStubProvider) Exchange(context.Context, Exchange) (VerifiedIdentity, error) {
	return VerifiedIdentity{}, errors.New("unused")
}
func (rlsStubProvider) RevokeProviderSession(context.Context, string) error {
	return errors.New("unused")
}
func (rlsStubProvider) ProviderLogoutURL(string, string) (string, error) {
	return "", errors.New("unused")
}
