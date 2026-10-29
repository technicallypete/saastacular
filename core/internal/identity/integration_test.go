//go:build integration

// Integration tests (B-IDENTITY-IT) against a disposable PostgreSQL initialized by the rendered role script and
// migrated here to the reviewed set, with the fake WorkOS of package identitytest and the real WorkOS adapter. They
// use only the exported identity API and fail, never skip, when the database settings are absent. Fixture success
// proves application enforcement only, never provider behaviour.
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, APP_ENV, and from
// version 6 (G008) DB_SUPERUSER and DB_SUPERUSER_PASSWORD: row-level security binds the owner too (FORCE), so fixture
// counts and other setup-side reads use a separately attested superuser connection, never an assertion path.
package identity_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/identity/workos"
	"core/internal/migrate"
	"core/migrations"
)

const (
	itClient  = "client_01INTEGRATION000000000000"
	originA   = "https://org-a.localhost:543"
	originB   = "https://org-b.localhost:543"
	idleTTL   = 30 * time.Minute
	absoluteT = 8 * time.Hour
	// G016 (Plan r4 §1.7), G057: tenant hosts resolve through Organization platform labels (org-a, org-b under the test
	// platform domain localhost, port 543) and verified domains; the shared auth host is the only IDENTITY_PUBLIC_ORIGINS
	// entry.
	orgAID     = "0f16a000-0000-4000-8000-0000000000a1"
	orgBID     = "0f16b000-0000-4000-8000-0000000000b1"
	authOrigin = "https://auth.localhost:543"
)

// itHosts is the integration HostResolver: the overlay's classification (<label>.localhost:543 platform hosts, any other
// name on port 543 a custom host, auth.localhost non-tenant) over the real app.hostname_state reader, one statement on
// the querier identity passes (the login or consume transaction, or the issuance transaction for the final guard).
type itHosts struct{}

func (itHosts) ResolveOrigin(ctx context.Context, q identity.Querier, origin string) (identity.HostResolution, error) {
	host, ok := strings.CutPrefix(origin, "https://")
	name, port, found := strings.Cut(host, ":")
	if !ok || !found || port != "543" || name == "" || strings.ContainsAny(name, "/?#@:") {
		return identity.HostResolution{}, nil
	}
	kind, key := identity.HostCustom, name
	if label, ok := strings.CutSuffix(name, ".localhost"); ok {
		if label == "" || strings.Contains(label, ".") || label == "auth" {
			return identity.HostResolution{}, nil
		}
		kind, key = identity.HostPlatform, label
	}
	var org *string
	var routing bool
	err := q.QueryRow(ctx, `SELECT organization_id::text, routing FROM app.hostname_state($1, $2)`, kind, key).Scan(&org, &routing)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return identity.HostResolution{}, nil
	case err != nil:
		return identity.HostResolution{}, err
	case org == nil || !routing:
		return identity.HostResolution{}, nil
	}
	return identity.HostResolution{Kind: kind, OrganizationID: *org, Routing: true}, nil
}

// provisionMu serializes fixture provisioning (concurrent sign-in tests provision one principal from many goroutines).
var provisionMu sync.Mutex

// seedTenants creates the two fixture Organizations (idempotent; superuser setup connection). G057 (D-2): org-a and
// org-b are their fixed synthetic platform labels (the hosts); the friendly slugs are metadata and never route.
func seedTenants(t *testing.T) {
	t.Helper()
	if _, err := setupConn(t).Exec(context.Background(), `INSERT INTO app.organizations (id, name, slug, platform_label) VALUES
  ($1::text::uuid, 'Org A', 'alpha-agency', 'org-a'), ($2::text::uuid, 'Org B', 'bravo-agency', 'org-b') ON CONFLICT (id) DO NOTHING`,
		orgAID, orgBID); err != nil {
		t.Fatal(err)
	}
}

// superDSN is the disposable instance's superuser connection string (setup only, as setupConn).
func superDSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", it.cfg.Host, it.cfg.Port, it.cfg.Database,
		os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD"))
}

// provisionE gives a provider subject of this harness's issuer a user, a principal and Organization memberships (org-a
// and org-b unless orgs names others; idempotent). Under D-2(a) (Plan r4 §1.7.1) a tenant-host sign-in completes only
// for a current member, so an unmapped principal never completes a first sign-in; the fixture provisions the account
// first, as an operator does. It never calls t.Fatal (goroutine-safe).
func (h *harness) provisionE(subject string, orgs ...string) (string, error) {
	if len(orgs) == 0 {
		orgs = []string{orgAID, orgBID}
	}
	provisionMu.Lock()
	defer provisionMu.Unlock()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, superDSN())
	if err != nil {
		return "", err
	}
	defer conn.Close(ctx)
	var uid string
	err = conn.QueryRow(ctx, `SELECT user_id::text FROM app.identity_principals WHERE issuer = $1 AND subject = $2`, h.fake.Issuer, subject).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		err = conn.QueryRow(ctx, `WITH u AS (INSERT INTO app.users (display_name) VALUES ('Synthetic User') RETURNING id),
  p AS (INSERT INTO app.identity_principals (issuer, subject, user_id) SELECT $1, $2, id FROM u)
SELECT id::text FROM u`, h.fake.Issuer, subject).Scan(&uid)
	}
	if err != nil {
		return "", err
	}
	for _, org := range orgs {
		if _, err := conn.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
SELECT $1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership'
WHERE NOT EXISTS (SELECT 1 FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid
                    AND scope_type = 'organization' AND source = 'membership')`, org, uid); err != nil {
			return "", err
		}
	}
	return uid, nil
}

func (h *harness) provision(subject string, orgs ...string) string {
	h.t.Helper()
	uid, err := h.provisionE(subject, orgs...)
	if err != nil {
		h.t.Fatalf("provision %s: %v", subject, err)
	}
	return uid
}

// code provisions the grant's principal and issues the fake provider's authorization code for challenge.
func (h *harness) code(challenge string, g identitytest.Grant) string {
	h.t.Helper()
	h.provision(g.UserID)
	return h.fake.IssueCode(challenge, g)
}

var it struct {
	cfg       migrate.Config
	runtimePW string
	pool      *pgxpool.Pool
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil || os.Getenv("DB_RUNTIME_PASSWORD") == "" {
		fmt.Fprintln(os.Stderr, "integration settings: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD and APP_ENV are required")
		os.Exit(1)
	}
	it.cfg, it.runtimePW = cfg, os.Getenv("DB_RUNTIME_PASSWORD")
	r, err := migrate.Open(ctx, cfg, migrations.FS)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open migrator:", err)
		os.Exit(1)
	}
	if err := r.Up(ctx, io.Discard); err != nil {
		fmt.Fprintln(os.Stderr, "migrate up:", err)
		os.Exit(1)
	}
	_ = r.Close()
	it.pool = runtimePool(nil)
	code := m.Run()
	it.pool.Close()
	if err := releaseFixtureSlugs(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "release fixture slugs:", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

// releaseFixtureSlugs clears the slugs and platform labels of the two seedTenants Organizations after the package's
// tests (G016, G057). The identity overlay runs these tests on its own database and then applies
// infra/identity/mcp-seed.sql, which gives the labels org-a, org-b and org-c (and their slugs) to the Organizations of
// infra/identity/mcp-hosts.tsv; a fixture Organization still holding a slug or label would make that seed fail on the
// unique index. The Organizations themselves stay (as before G016).
func releaseFixtureSlugs(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, superDSN())
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, `UPDATE app.organizations SET slug = NULL, platform_label = NULL WHERE id IN ($1::text::uuid, $2::text::uuid)
  AND (slug IS NOT NULL OR platform_label IS NOT NULL)`, orgAID, orgBID)
	return err
}

func runtimePool(t *testing.T) *pgxpool.Pool {
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		panic(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = it.cfg.Host, uint16(it.cfg.Port), it.cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = it.cfg.RuntimeRole(), it.runtimePW
	pc.ConnConfig.ConnectTimeout = 3 * time.Second
	pc.MaxConns = 40
	p, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		panic(err)
	}
	if t != nil {
		t.Cleanup(p.Close)
	}
	return p
}

// ownerConn connects as the migrator and assumes the owner role (catalog inspection, cleanup, fault injection).
func ownerConn(t *testing.T) *pgx.Conn {
	t.Helper()
	cc, err := pgx.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	cc.Host, cc.Port, cc.Database, cc.User, cc.Password = it.cfg.Host, uint16(it.cfg.Port), it.cfg.Database, it.cfg.MigratorRole(), it.cfg.Password
	cc.TLSConfig, cc.Fallbacks = nil, nil
	conn, err := pgx.ConnectConfig(context.Background(), cc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	if _, err := conn.Exec(context.Background(), "SET ROLE "+pgx.Identifier{it.cfg.OwnerRole()}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	return conn
}

// setupConn connects as the disposable instance's superuser (G008, decided Q4): setup-side counts, fixtures and
// constraint controls only, never an isolation assertion. It attests rolsuper before use.
func setupConn(t *testing.T) *pgx.Conn {
	t.Helper()
	user, pw := os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD")
	if user == "" || pw == "" {
		t.Fatal("integration settings: DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required from version 6")
	}
	cc, err := pgx.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	cc.Host, cc.Port, cc.Database, cc.User, cc.Password = it.cfg.Host, uint16(it.cfg.Port), it.cfg.Database, user, pw
	cc.TLSConfig, cc.Fallbacks = nil, nil
	conn, err := pgx.ConnectConfig(context.Background(), cc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	var super bool
	if err := conn.QueryRow(context.Background(), `SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&super); err != nil || !super {
		t.Fatalf("setup connection is not a superuser (%v)", err)
	}
	return conn
}

// runtimeConn connects as the runtime role (grant and denial checks).
func runtimeConn(t *testing.T) *pgx.Conn {
	t.Helper()
	cc, err := pgx.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	cc.Host, cc.Port, cc.Database, cc.User, cc.Password = it.cfg.Host, uint16(it.cfg.Port), it.cfg.Database, it.cfg.RuntimeRole(), it.runtimePW
	cc.TLSConfig, cc.Fallbacks = nil, nil
	conn, err := pgx.ConnectConfig(context.Background(), cc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

// resetIdentity empties every identity table (owner TRUNCATE; runtime has no DELETE on most of them). CASCADE (G005):
// app.role_bindings references app.users, so its membership bindings are emptied with the users; Organizations and
// Groups remain. The statement is byte-identical to the workos_staging fixture's.
func resetIdentity(t *testing.T) {
	t.Helper()
	if _, err := ownerConn(t).Exec(context.Background(),
		`TRUNCATE app.login_transactions, app.sessions, app.identity_principals, app.users, app.identity_provider_binding CASCADE`); err != nil {
		t.Fatal(err)
	}
}

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock() *clock { return &clock{t: time.Now().UTC().Truncate(time.Second)} }

func (c *clock) Now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) Advance(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

type logBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *logBuf) Write(p []byte) (int, error) { l.mu.Lock(); defer l.mu.Unlock(); return l.b.Write(p) }
func (l *logBuf) String() string              { l.mu.Lock(); defer l.mu.Unlock(); return l.b.String() }

type harness struct {
	t    *testing.T
	fake *identitytest.Fake
	clk  *clock
	cfg  identity.Config
	svc  *identity.Service
	logs *logBuf
}

type option func(*identity.Config)

func withMaxAuthAge(d time.Duration) option { return func(c *identity.Config) { c.MaxAuthAge = d } }
func withClient(id string) option           { return func(c *identity.Config) { c.ClientID = id } }
func withIssuer(iss string) option          { return func(c *identity.Config) { c.Issuer = iss } }
func unconfigured() option                  { return func(c *identity.Config) { c.Provider = nil } }

// newHarness builds a service over the shared runtime pool and a fresh fake WorkOS. Each harness has its own fake
// (own issuer URL), so tests that need one namespace reset the tables first.
func newHarness(t *testing.T, opts ...option) *harness {
	t.Helper()
	f := identitytest.New(t, itClient)
	clk := newClock()
	f.SetNow(clk.Now)
	a, err := workos.New(workos.Config{APIKey: f.APIKey, ClientID: itClient, Issuer: f.Issuer, BaseURL: f.URL(), Algorithms: []string{"RS256"}, Now: clk.Now})
	if err != nil {
		t.Fatal(err)
	}
	logs := &logBuf{}
	seedTenants(t)
	cfg := identity.Config{Provider: a, Issuer: f.Issuer, ClientID: itClient, Origins: identity.NewOrigins([]string{authOrigin}),
		Hosts: itHosts{}, AuthOrigin: authOrigin,
		Idle: idleTTL, Max: absoluteT, Now: clk.Now, Logger: slog.New(slog.NewJSONHandler(logs, nil))}
	for _, o := range opts {
		o(&cfg)
	}
	return &harness{t: t, fake: f, clk: clk, cfg: cfg, svc: identity.NewService(identity.NewStore(it.pool), cfg), logs: logs}
}

// sibling builds a second service (another replica or a restarted instance) on its own pool with the same
// configuration.
func (h *harness) sibling() *identity.Service {
	h.t.Helper()
	return identity.NewService(identity.NewStore(runtimePool(h.t)), h.cfg)
}

// startE begins a login and returns the transaction's state and challenge as the provider would see them. It never
// calls t.Fatal, so it is safe in goroutines.
func (h *harness) startE(origin, binding string) (identity.LoginOutput, string, string, error) {
	out, err := h.svc.Login(context.Background(), identity.LoginInput{Origin: origin, Binding: binding, ReturnTo: "/account"})
	if err != nil {
		return identity.LoginOutput{}, "", "", err
	}
	u, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		return identity.LoginOutput{}, "", "", err
	}
	return out, u.Query().Get("state"), u.Query().Get("code_challenge"), nil
}

func (h *harness) start(origin, binding string) (identity.LoginOutput, string, string) {
	h.t.Helper()
	out, state, challenge, err := h.startE(origin, binding)
	if err != nil {
		h.t.Fatalf("login start: %v", err)
	}
	return out, state, challenge
}

func grant(user, sid string) identitytest.Grant {
	return identitytest.Grant{UserID: user, SID: sid, FirstName: "Synthetic", LastName: "User"}
}

func query(code, state string) string {
	return "code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state)
}

// signIn runs a complete login for a principal on origin, optionally with a live credential (goroutine-safe).
func (h *harness) signIn(origin, cred string, g identitytest.Grant) (identity.CallbackOutput, error) {
	out, state, challenge, err := h.startE(origin, "")
	if err != nil {
		return identity.CallbackOutput{}, err
	}
	if _, err := h.provisionE(g.UserID); err != nil {
		return identity.CallbackOutput{}, err
	}
	code := h.fake.IssueCode(challenge, g)
	return h.svc.Callback(context.Background(), identity.CallbackInput{Origin: origin, RawQuery: query(code, state),
		Binding: out.Binding, Credential: cred})
}

func (h *harness) mustSignIn(origin string, g identitytest.Grant) identity.CallbackOutput {
	h.t.Helper()
	out, err := h.signIn(origin, "", g)
	if err != nil || out.Credential == "" {
		h.t.Fatalf("sign-in: %+v %v", out, err)
	}
	return out
}

func (h *harness) session(cred string) identity.SessionInfo {
	h.t.Helper()
	return h.sessionAt(cred, originA)
}

// sessionAt resolves cred at origin (G016 security amendment: a session resolves only at its issuing origin).
func (h *harness) sessionAt(cred, origin string) identity.SessionInfo {
	h.t.Helper()
	info, err := h.svc.Session(context.Background(), cred, origin)
	if err != nil {
		h.t.Fatalf("session: %v", err)
	}
	return info
}

func csrfOf(info identity.SessionInfo) string {
	return base64.RawURLEncoding.EncodeToString(info.CSRFToken)
}

func count(t *testing.T, q string, args ...any) int {
	t.Helper()
	var n int
	if err := setupConn(t).QueryRow(context.Background(), q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// I-MAP-CONC: 32 concurrent first sign-ins for one principal create exactly one user and one principal.
func TestConcurrentFirstSignIns(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	var wg sync.WaitGroup
	var ok atomic.Int32
	users := sync.Map{}
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := h.signIn(originA, "", grant("user_01CONC", fmt.Sprintf("session_%02d", i)))
			if err == nil {
				ok.Add(1)
				if info, err := h.svc.Session(context.Background(), out.Credential, originA); err == nil {
					users.Store(info.UserID, true)
				}
			}
		}()
	}
	wg.Wait()
	n := 0
	users.Range(func(any, any) bool { n++; return true })
	if ok.Load() != 32 || n != 1 {
		t.Fatalf("successes %d, distinct users %d", ok.Load(), n)
	}
	if u, p := count(t, `SELECT count(*) FROM app.users`), count(t, `SELECT count(*) FROM app.identity_principals`); u != 1 || p != 1 {
		t.Fatalf("users %d principals %d", u, p)
	}
}

// Eight parallel identical callbacks consume the transaction exactly once and issue at most one session.
func TestParallelIdenticalCallbacksConsumeOnce(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	out, state, challenge := h.start(originA, "")
	code := h.code(challenge, grant("user_01PAR", "session_par"))
	var wg sync.WaitGroup
	var signedIn, invalid atomic.Int32
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.svc.Callback(context.Background(), identity.CallbackInput{Origin: originA, RawQuery: query(code, state), Binding: out.Binding})
			switch err {
			case nil:
				signedIn.Add(1)
			case identity.ErrTransactionInvalid:
				invalid.Add(1)
			}
		}()
	}
	wg.Wait()
	if signedIn.Load() != 1 || invalid.Load() != 7 || count(t, `SELECT count(*) FROM app.sessions`) != 1 {
		t.Fatalf("signed in %d, invalid %d", signedIn.Load(), invalid.Load())
	}
}

// I-ATOMIC: a failure between mapping and session commits nothing usable (no orphan user, principal or session).
func TestSignInIsAtomic(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	own := ownerConn(t)
	ctx := context.Background()
	for _, stmt := range []string{
		`CREATE FUNCTION app.g004_it_fail() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected'; END $$`,
		`CREATE TRIGGER g004_it_fail BEFORE INSERT ON app.sessions FOR EACH ROW EXECUTE FUNCTION app.g004_it_fail()`,
	} {
		if _, err := own.Exec(ctx, stmt); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		c := ownerConn(t)
		_, _ = c.Exec(ctx, `DROP TRIGGER IF EXISTS g004_it_fail ON app.sessions`)
		_, _ = c.Exec(ctx, `DROP FUNCTION IF EXISTS app.g004_it_fail()`)
	})
	if _, err := h.signIn(originA, "", grant("user_01ATOM", "session_atom")); err != identity.ErrUnavailable {
		t.Fatalf("injected failure: %v", err)
	}
	// G016 D-2(a): the principal is provisioned (one user, one principal) before its sign-in; the failure adds nothing.
	if n := count(t, `SELECT ((SELECT count(*) FROM app.users) + (SELECT count(*) FROM app.identity_principals) + (SELECT count(*) FROM app.sessions))::int`); n != 2 {
		t.Fatalf("%d rows (want the 2 provisioned) after the failure", n)
	}
}

// I-LIFE: idle and absolute expiry with a controlled clock; renewal never extends the absolute expiry and writes at
// most once per 60 s.
func TestSessionLifetimes(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	cred := h.mustSignIn(originA, grant("user_01LIFE", "session_life")).Credential

	h.clk.Advance(idleTTL - time.Minute)
	if _, err := h.svc.Session(ctx, cred, originA); err != nil {
		t.Fatalf("within idle: %v", err)
	}
	h.clk.Advance(idleTTL - time.Minute) // renewed above, so still live
	if _, err := h.svc.Session(ctx, cred, originA); err != nil {
		t.Fatalf("after renewal: %v", err)
	}
	h.clk.Advance(idleTTL + time.Second)
	if _, err := h.svc.Session(ctx, cred, originA); err != identity.ErrUnauthenticated {
		t.Fatalf("idle expiry: %v", err)
	}

	// Absolute expiry wins over continuous activity.
	cred = h.mustSignIn(originA, grant("user_01LIFE", "session_life2")).Credential
	total := time.Duration(0)
	for total+20*time.Minute < absoluteT {
		h.clk.Advance(20 * time.Minute)
		total += 20 * time.Minute
		info, err := h.svc.Session(ctx, cred, originA)
		if err != nil {
			t.Fatalf("active session at %v: %v", total, err)
		}
		if !info.ExpiresAt.After(h.clk.Now()) {
			t.Fatal("expiry not in the future")
		}
	}
	h.clk.Advance(absoluteT - total + time.Second)
	if _, err := h.svc.Session(ctx, cred, originA); err != identity.ErrUnauthenticated {
		t.Fatalf("absolute expiry: %v", err)
	}
	if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE idle_expires_at > absolute_expires_at`); n != 0 {
		t.Fatalf("%d sessions with idle beyond absolute", n)
	}

	// Renewal writes at most once per 60 s.
	cred = h.mustSignIn(originA, grant("user_01LIFE", "session_life3")).Credential
	before := count(t, `SELECT extract(epoch FROM renewed_at)::int FROM app.sessions WHERE provider_session_id = 'session_life3'`)
	h.clk.Advance(30 * time.Second)
	_, _ = h.svc.Session(ctx, cred, originA)
	if after := count(t, `SELECT extract(epoch FROM renewed_at)::int FROM app.sessions WHERE provider_session_id = 'session_life3'`); after != before {
		t.Fatal("renewed within 60 s")
	}
}

// Revoke versus renew: 1,000 iterations of concurrent renewal and revocation never resurrect a revoked session.
func TestRevokeVersusRenewNeverResurrects(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		cred := h.mustSignIn(originA, grant("user_01RACE", fmt.Sprintf("session_race_%d", i))).Credential
		info := h.session(cred)
		h.clk.Advance(61 * time.Second) // make the next resolution renew
		var wg sync.WaitGroup
		var logoutErr error
		wg.Add(2)
		go func() { defer wg.Done(); _, _ = h.svc.Session(ctx, cred, originA) }()
		go func() {
			defer wg.Done()
			_, logoutErr = h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: csrfOf(info), Origin: originA, Credential: cred})
		}()
		wg.Wait()
		if logoutErr != nil {
			t.Fatalf("iteration %d: logout %v", i, logoutErr)
		}
		if _, err := h.svc.Session(ctx, cred, originA); err != identity.ErrUnauthenticated {
			t.Fatalf("iteration %d: revoked session resolved (%v)", i, err)
		}
		if i%100 == 99 {
			resetIdentity(t) // keep the table small; the binding is re-created by the next sign-in
		}
	}
}

// I-REPLICA and restart: revocation on one instance is seen by another; a new instance keeps valid sessions and the
// revoked denial (PostgreSQL is the only state).
func TestReplicasAndRestart(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	a1 := h.mustSignIn(originA, grant("user_01REP", "session_rep1")).Credential
	a2 := h.mustSignIn(originA, grant("user_01REP", "session_rep2")).Credential
	replica := h.sibling()
	info, err := replica.Session(ctx, a1, originA)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: csrfOf(info), Origin: originA, Credential: a1}); err != nil {
		t.Fatal(err)
	}
	if _, err := replica.Session(ctx, a1, originA); err != identity.ErrUnauthenticated {
		t.Fatalf("replica resolved a revoked session: %v", err)
	}
	restarted := h.sibling()
	if _, err := restarted.Session(ctx, a2, originA); err != nil {
		t.Fatalf("restart lost a valid session: %v", err)
	}
	if _, err := restarted.Session(ctx, a1, originA); err != identity.ErrUnauthenticated {
		t.Fatalf("restart resurrected a revoked session: %v", err)
	}
}

// I-OUTAGE: an unreachable store is a bounded identity_unavailable, never a cached or authenticated success.
func TestStoreOutageIsUnavailable(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	cred := h.mustSignIn(originA, grant("user_01OUT", "session_out")).Credential
	pc, _ := pgxpool.ParseConfig("")
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = it.cfg.Host, 1, it.cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = it.cfg.RuntimeRole(), it.runtimePW
	pc.ConnConfig.ConnectTimeout = time.Second
	dead, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	defer dead.Close()
	down := identity.NewService(identity.NewStore(dead), h.cfg)
	start := time.Now()
	for name, err := range map[string]error{
		"session": func() error { _, err := down.Session(context.Background(), cred, originA); return err }(),
		"logout": func() error {
			_, err := down.Logout(context.Background(), identity.LogoutInput{Scope: identity.ScopeCurrent, Origin: originA, Credential: cred})
			return err
		}(),
		"login": func() error {
			_, err := down.Login(context.Background(), identity.LoginInput{Origin: originA})
			return err
		}(),
	} {
		if err != identity.ErrUnavailable {
			t.Fatalf("outage %s: %v", name, err)
		}
	}
	if time.Since(start) > 10*time.Second {
		t.Fatal("outage handling not bounded")
	}
	if _, err := h.svc.Session(context.Background(), cred, originA); err != nil {
		t.Fatalf("recovery: %v", err)
	}
}

// Disable versus sign-in: whatever the interleaving, a disabled user never ends with a live session.
func TestDisableVersusSignInRace(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	info := h.session(h.mustSignIn(originA, grant("user_01DIS", "session_dis0")).Credential)
	for i := 0; i < 50; i++ {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = h.signIn(originA, "", grant("user_01DIS", fmt.Sprintf("session_dis_%d", i)))
		}()
		go func() { defer wg.Done(); _ = h.svc.DisableUser(ctx, info.UserID) }()
		wg.Wait()
		if n := count(t, `SELECT count(*)::int FROM app.sessions s JOIN app.users u ON u.id = s.user_id WHERE u.status = 'disabled' AND s.revoked_at IS NULL`); n != 0 {
			t.Fatalf("iteration %d: %d unrevoked sessions for a disabled user", i, n)
		}
		if err := h.svc.EnableUser(ctx, info.UserID); err != nil {
			t.Fatal(err)
		}
	}
}

// Logout-all versus a concurrent new sign-in: sessions existing at the revocation are revoked; a session created
// afterwards survives; nothing is half-applied.
func TestLogoutAllVersusConcurrentSignIn(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	c1 := h.mustSignIn(originA, grant("user_01ALLR", "session_allr1")).Credential
	info := h.session(c1)
	var later string
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: csrfOf(info), Origin: originA, Credential: c1})
	}()
	go func() {
		defer wg.Done()
		if out, err := h.signIn(originA, "", grant("user_01ALLR", "session_allr2")); err == nil {
			later = out.Credential
		}
	}()
	wg.Wait()
	if _, err := h.svc.Session(ctx, c1, originA); err != identity.ErrUnauthenticated {
		t.Fatalf("logout-all left the original session: %v", err)
	}
	if later == "" {
		t.Fatal("concurrent sign-in failed")
	}
	var revoked bool
	if err := setupConn(t).QueryRow(ctx, `SELECT revoked_at IS NOT NULL FROM app.sessions WHERE provider_session_id = 'session_allr2'`).Scan(&revoked); err != nil {
		t.Fatal(err)
	}
	if _, err := h.svc.Session(ctx, later, originA); revoked != (err == identity.ErrUnauthenticated) {
		t.Fatalf("revoked=%v but resolution err=%v", revoked, err)
	}
	after := h.mustSignIn(originA, grant("user_01ALLR", "session_allr3")).Credential
	if _, err := h.svc.Session(ctx, after, originA); err != nil {
		t.Fatalf("new sign-in after logout-all: %v", err)
	}
}

// U-LOG (fixture canaries) and R9: identity records carry only closed outcome/scope values and correlation.
func TestLogsCarryNoValues(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	g := grant("user_01LOGCANARY", "session_logcanary")
	g.FirstName, g.LastName, g.Email = "CanaryFirst", "<script>CanaryLast</script>", "canary-mail@example.test"
	out, state, challenge := h.start(originA, "")
	code := h.code(challenge, g)
	// Unknown parameters and an error_description without an error are ignored (never echoed or logged).
	cb, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA,
		RawQuery: query(code, state) + "&error_description=canary-desc&return_to=canary-return", Binding: out.Binding})
	if err != nil {
		t.Fatal(err)
	}
	info := h.session(cb.Credential)
	_, _ = h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfOf(info), Origin: originA, Credential: cb.Credential})
	logs := h.logs.String()
	for _, canary := range []string{code, state, out.Binding, cb.Credential, csrfOf(info), "user_01LOGCANARY", "session_logcanary",
		"CanaryFirst", "CanaryLast", "canary-mail", "canary-desc", "canary-return", info.UserID, h.fake.APIKey, "eyJ", "code="} {
		if strings.Contains(logs, canary) {
			t.Errorf("log contains %q", canary)
		}
	}
	for _, ev := range []string{"auth.login.started", "auth.callback.completed", "auth.session.revoked", "auth.provider.revocation"} {
		if !strings.Contains(logs, `"msg":"`+ev+`"`) {
			t.Errorf("event %s missing", ev)
		}
	}
}
