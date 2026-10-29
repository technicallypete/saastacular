//go:build integration

// B-POLICY-IT (G007 Plan r1 §6.3): TestAccessPolicy (and TestAccessFootprint, footprint_integration_test.go) against a
// disposable PostgreSQL initialized by the rendered role script and migrated here to the reviewed set
// (infra/validation/run.sh, suite "migrations"). Sessions are issued by the real G004 Login/Callback against
// identitytest.Fake and the WorkOS adapter. Every access operation runs through a pool logged in as the actual runtime
// role, wrapped by the counting hookDB; fixtures and conflicting writers use the accepted tenancy and identity
// services. Assertions compare scoped row-set hashes and print hashes only. Fails, never skips, when settings are
// absent.
//
// From version 6 (G008 RC-8) row-level security binds the runtime role and, through FORCE, the owner: tenancy fixtures,
// conflicting tenancy writers and every setup-side read (row hashes, flag fixture, footprint seed) use the separately
// attested superuser of the disposable instance (setupConn/setupPool, decided Q4), never an assertion path. Access
// operations still run only through the runtime pool, and identity flows through the real G004 services.
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, DB_SUPERUSER,
// DB_SUPERUSER_PASSWORD, APP_ENV, TENANCY_CANARY_FILE (per-run canaries, one per line) and TENANCY_LOG_FILE (the
// access.decision records are appended).
package access

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/identity/workos"
	"core/internal/migrate"
	"core/internal/observability"
	"core/internal/tenancy"
	"core/migrations"
)

const (
	itClient = "client_01ACCESSPOLICY000000000000"
	itOrigin = "https://org-a.localhost:543"
	itIdle   = 30 * time.Minute
	itMax    = 8 * time.Hour
)

// --- environment -----------------------------------------------------------------------------------------------------

type itSyncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *itSyncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}
func (s *itSyncBuf) String() string { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

type itClock struct {
	mu sync.Mutex
	t  time.Time
}

func newITClock() *itClock { return &itClock{t: time.Now().UTC().Truncate(time.Second)} }

func (c *itClock) Now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *itClock) Advance(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

type itEnv struct {
	cfg      migrate.Config
	pw       string
	pool     *pgxpool.Pool
	setup    *pgx.Conn     // attested superuser: fixtures and setup-side reads only (G008 RC-8)
	setupDB  *pgxpool.Pool // attested superuser pool for the tenancy model fixtures and conflicting writers
	logs     *itSyncBuf
	canaries []string
	fake     *identitytest.Fake
	clk      *itClock
	id       *identity.Service
	hook     *hookDB
	svc      *Service
	model    *tenancy.Service
	// G016: the fixture sign-in Organization behind itOrigin (see itHost and signInWith).
	loginOrg tenancy.OrganizationID
	// G016 A2: the origin every fixture credential was issued at (one per credential; set by signInAt only).
	origins sync.Map
}

// itHost is the fixture tenant-host resolver of the access environment (G016): itOrigin routes to the environment's
// sign-in Organization and itOriginOf(org) to the test-world Organization org (G016 A2: a browser session is bound to
// the Organization of the origin it was issued at, so each world Organization has its own tenant origin); anything else
// routes nowhere. It reads nothing (the routing rules are the domains package's own tests).
type itHost struct{ org string }

const itOrgHostSuffix = ".org.localhost:543"

// itOriginOf is the fixture tenant origin of test-world Organization org.
func itOriginOf(org tenancy.OrganizationID) string {
	return "https://" + org.String() + itOrgHostSuffix
}

func (h itHost) ResolveOrigin(_ context.Context, _ identity.Querier, origin string) (identity.HostResolution, error) {
	if origin == itOrigin {
		return identity.HostResolution{Kind: identity.HostPlatform, OrganizationID: h.org, Routing: true}, nil
	}
	if rest, ok := strings.CutPrefix(origin, "https://"); ok {
		if id, ok := strings.CutSuffix(rest, itOrgHostSuffix); ok {
			if org, err := tenancy.ParseOrganizationID(id); err == nil && org.String() == id {
				return identity.HostResolution{Kind: identity.HostPlatform, OrganizationID: id, Routing: true}, nil
			}
		}
	}
	return identity.HostResolution{}, nil
}

func itConnect(t *testing.T, cfg migrate.Config, user, pw string) *pgx.Conn {
	t.Helper()
	cc, err := pgx.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	cc.Host, cc.Port, cc.Database, cc.User, cc.Password = cfg.Host, uint16(cfg.Port), cfg.Database, user, pw
	cc.ConnectTimeout = 5 * time.Second
	cc.TLSConfig, cc.Fallbacks = nil, nil
	c, err := pgx.ConnectConfig(context.Background(), cc)
	if err != nil {
		t.Fatalf("connect as %s: %v", user, err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })
	return c
}

// itSetup opens the separately attested superuser connection and pool of the disposable instance (G008 RC-8, decided
// Q4). Both attest rolsuper before use; an isolation assertion must never use them (TestRLSIsolation's
// privileged-assertion negative).
func itSetup(t *testing.T, cfg migrate.Config) (*pgx.Conn, *pgxpool.Pool) {
	t.Helper()
	user, pw := os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD")
	if user == "" || pw == "" {
		t.Fatal("integration settings: DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required from version 6")
	}
	c := itConnect(t, cfg, user, pw)
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = cfg.Host, uint16(cfg.Port), cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = user, pw
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	pc.MaxConns = 16
	p, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	for _, q := range []interface {
		QueryRow(context.Context, string, ...any) pgx.Row
	}{c, p} {
		var super bool
		if err := q.QueryRow(context.Background(), `SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&super); err != nil || !super {
			t.Fatalf("setup connection is not a superuser (%v)", err)
		}
	}
	return c, p
}

// runtimePool returns a pool logged in as the runtime role; tracer, when set, observes its statements.
func (e *itEnv) runtimePool(t *testing.T, tracer pgx.QueryTracer) *pgxpool.Pool {
	t.Helper()
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = e.cfg.Host, uint16(e.cfg.Port), e.cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = e.cfg.RuntimeRole(), e.pw
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	pc.ConnConfig.Tracer = tracer
	pc.MaxConns = 16
	p, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

// setupIT migrates to the reviewed set, opens the runtime pool and the owner connection, and builds the identity
// service (the fake WorkOS and the real adapter, on a controllable clock) and the access service over the hookDB. The
// access.decision records are appended to TENANCY_LOG_FILE at cleanup.
func setupIT(t *testing.T) *itEnv {
	t.Helper()
	ctx := context.Background()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil {
		t.Fatalf("integration settings: %v", err)
	}
	e := &itEnv{cfg: cfg, pw: os.Getenv("DB_RUNTIME_PASSWORD")}
	canaryFile, logFile := os.Getenv("TENANCY_CANARY_FILE"), os.Getenv("TENANCY_LOG_FILE")
	if e.pw == "" || canaryFile == "" || logFile == "" {
		t.Fatal("integration settings: DB_RUNTIME_PASSWORD, TENANCY_CANARY_FILE and TENANCY_LOG_FILE are required")
	}
	b, err := os.ReadFile(canaryFile)
	if err != nil {
		t.Fatalf("canary file: %v", err)
	}
	if e.canaries = strings.Fields(string(b)); len(e.canaries) < 3 {
		t.Fatal("canary file: at least three canaries are required")
	}
	r, err := migrate.Open(ctx, cfg, migrations.FS)
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	var out bytes.Buffer
	// G011: the audit triggers are active for B-POLICY-IT, B-RLS-IT and B-OBJ-IT. G013/G014/G016/G017/G018/G019/G058/G060/G062/G063/G020/G057/G071:
	// the latest version is 22, so they (and G010's attack suite) also run with the retention policies and grants, the MCP connection objects,
	// the membership guard, policies and invitation table, the organization enrollment tables and functions, and the platform labels present.
	if err := r.Up(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t22\n") {
		t.Fatalf("migrate up: %v %q", err, out.String())
	}
	_ = r.Close()
	e.setup, e.setupDB = itSetup(t, cfg)
	// Earlier tests of the suite bind another provider namespace; this test's fake WorkOS is a new one. Only the
	// namespace row is removed (owner, disposable database); every other row stays.
	if _, err := e.setup.Exec(ctx, `DELETE FROM app.identity_provider_binding`); err != nil {
		t.Fatal(err)
	}
	e.pool = e.runtimePool(t, nil)
	e.fake = identitytest.New(t, itClient)
	e.clk = newITClock()
	e.fake.SetNow(e.clk.Now)
	e.model = tenancy.NewService(e.setupDB, nil)
	e.loginOrg = e.org(t, "Sign-in fixture")
	e.id = e.identity(t, e.clk, e.pool)
	e.logs = &itSyncBuf{}
	e.hook = &hookDB{pool: e.pool}
	e.svc = NewService(e.hook, e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
	t.Cleanup(func() {
		// World-readable: bounded codes only, read by the nonroot log checker. Appended: TestAccessPolicy and
		// TestAccessFootprint run in one process and share the file.
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			_, err = f.WriteString(e.logs.String())
			_ = f.Close()
		}
		if err != nil {
			t.Errorf("write the access log: %v", err)
		}
	})
	return e
}

// identity returns an identity service over pool on clock clk, sharing this environment's fake WorkOS.
func (e *itEnv) identity(t *testing.T, clk *itClock, pool *pgxpool.Pool) *identity.Service {
	t.Helper()
	a, err := workos.New(workos.Config{APIKey: e.fake.APIKey, ClientID: itClient, Issuer: e.fake.Issuer, BaseURL: e.fake.URL(),
		Algorithms: []string{"RS256"}, Now: clk.Now})
	if err != nil {
		t.Fatal(err)
	}
	return identity.NewService(identity.NewStore(pool), identity.Config{Provider: a, Issuer: e.fake.Issuer, ClientID: itClient,
		Hosts: itHost{org: e.loginOrg.String()}, Idle: itIdle, Max: itMax, Now: clk.Now})
}

// admit prepares one fixture sign-in under G016 D-2(a), where a tenant-host session is issued only to a current member
// of the host's Organization org: the subject's user and principal exist (created as an operator would, with the
// display name the sign-in would set) and it holds a membership of org. A membership is added only when the subject
// holds no Organization-scope binding of org; the returned function removes exactly the row it added (by id) after the
// sign-in, so an existing owner/admin/member binding is never touched and every later assertion sees exactly the
// memberships the test grants (G016 A2 fixture correction).
func (e *itEnv) admit(t *testing.T, org tenancy.OrganizationID, sub, display string) func() {
	t.Helper()
	ctx := context.Background()
	var uid string
	err := e.setupDB.QueryRow(ctx, `SELECT user_id::text FROM app.identity_principals WHERE issuer = $1 AND subject = $2`, e.fake.Issuer, sub).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		err = e.setupDB.QueryRow(ctx, `WITH u AS (INSERT INTO app.users (display_name) VALUES ($3) RETURNING id),
  p AS (INSERT INTO app.identity_principals (issuer, subject, user_id) SELECT $1, $2, id FROM u)
SELECT id::text FROM u`, e.fake.Issuer, sub, display).Scan(&uid)
	}
	if err != nil {
		t.Fatalf("admit %s: %v", sub, err)
	}
	var added string
	err = e.setupDB.QueryRow(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership')
ON CONFLICT (subject_user_id, scope_type, scope_id) DO NOTHING RETURNING id::text`, org.String(), uid).Scan(&added)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil // an existing Organization-scope binding: nothing added, nothing to remove
	}
	if err != nil {
		t.Fatalf("admit %s: %v", sub, err)
	}
	return func() {
		if added == "" {
			return
		}
		if _, err := e.setupDB.Exec(ctx, `DELETE FROM app.role_bindings WHERE id = $1::text::uuid`, added); err != nil {
			t.Errorf("release %s: %v", sub, err)
		}
	}
}

// service returns a separate access service (its own hookDB and log) over identity service id.
func (e *itEnv) service(id *identity.Service) (*Service, *hookDB) {
	h := &hookDB{pool: e.pool}
	return NewService(h, id, observability.NewLogger(e.logs, slog.LevelInfo, nil)), h
}

// --- identity fixtures -------------------------------------------------------------------------------------------------

type itUser struct {
	sub     string
	first   string
	cred    string // bound to the sign-in fixture Organization at itOrigin
	id      tenancy.UserID
	session string
	// at holds the explicitly prepared credential per test-world Organization (G016 A2; see bind and in).
	at map[tenancy.OrganizationID]string
}

var itSeq atomic.Int64

// signInWith completes one real G004 login of subject sub through id at the sign-in fixture Organization's origin
// itOrigin and returns the new credential. The fake's default access-token claims carry a vendor organization, role
// admin and permissions [all]; they grant nothing.
func (e *itEnv) signInWith(t *testing.T, id *identity.Service, sub, first string, claims map[string]any, method, live string) string {
	t.Helper()
	return e.signInAt(t, id, e.loginOrg, sub, first, claims, method, live)
}

// itOriginFor is the fixture origin of org: itOrigin for the sign-in fixture Organization, itOriginOf otherwise.
func (e *itEnv) itOriginFor(org tenancy.OrganizationID) string {
	if org == e.loginOrg {
		return itOrigin
	}
	return itOriginOf(org)
}

// originOf is the origin credential cred was issued at (itOrigin for anything signInAt did not issue, e.g. an invalid
// credential of a negative case). It never depends on an outcome: each credential has exactly one origin.
func (e *itEnv) originOf(cred string) string {
	if o, ok := e.origins.Load(cred); ok {
		return o.(string)
	}
	return itOrigin
}

// signInAt completes one real G004 login of subject sub through id at org's own origin, so the session is bound to
// org (G016 A2), and returns the new credential.
func (e *itEnv) signInAt(t *testing.T, id *identity.Service, org tenancy.OrganizationID, sub, first string, claims map[string]any, method, live string) string {
	t.Helper()
	ctx := context.Background()
	origin := e.itOriginFor(org)
	defer e.admit(t, org, sub, first+" Fixture")()
	out, err := id.Login(ctx, identity.LoginInput{Origin: origin, ReturnTo: "/account"})
	if err != nil {
		t.Fatalf("login start: %v", err)
	}
	u, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	sid := fmt.Sprintf("session_01IT%08d", itSeq.Add(1))
	code := e.fake.IssueCode(u.Query().Get("code_challenge"), identitytest.Grant{UserID: sub, FirstName: first, LastName: "Fixture",
		SID: sid, Claims: claims, Method: method})
	cb, err := id.Callback(ctx, identity.CallbackInput{Origin: origin, RawQuery: "code=" + url.QueryEscape(code) + "&state=" +
		url.QueryEscape(u.Query().Get("state")), Binding: out.Binding, Credential: live})
	if err != nil || cb.Credential == "" {
		t.Fatalf("callback: %+v %v", cb.Outcome, err)
	}
	e.origins.Store(cb.Credential, origin)
	return cb.Credential
}

// bind prepares u's credential for each test-world Organization in orgs through the real sign-in flow at that
// Organization's own origin (G016 A2). Setup only: never inside a measured, audited or locked operation, and never as a
// reaction to a refusal. A subject without a membership of org is admitted only for the issuance (see admit); the
// admission is removed before bind returns, so a membership-negative subject stays membership-negative.
func (e *itEnv) bind(t *testing.T, u *itUser, orgs ...tenancy.OrganizationID) {
	t.Helper()
	for _, org := range orgs {
		if u.at == nil {
			u.at = map[tenancy.OrganizationID]string{}
		}
		u.at[org] = e.signInAt(t, e.id, org, u.sub, u.first, nil, "", "")
	}
}

// in is u's prepared credential for org; a missing one fails the test (no credential is issued or substituted here).
func (u *itUser) in(t *testing.T, org tenancy.OrganizationID) string {
	t.Helper()
	c := u.at[org]
	if c == "" {
		t.Fatalf("fixture: no credential prepared for %s at Organization %s", u.sub, itH(org))
	}
	return c
}

// newUser signs in a new global user (a new WorkOS subject) whose display name carries a canary.
func (e *itEnv) newUser(t *testing.T) *itUser {
	t.Helper()
	n := itSeq.Add(1)
	u := &itUser{sub: fmt.Sprintf("user_01ACCESS%08d", n), first: e.canaries[int(n)%len(e.canaries)] + " org_owner"}
	u.cred = e.signInWith(t, e.id, u.sub, u.first, nil, "", "")
	info, err := e.id.Session(context.Background(), u.cred, itOrigin)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	if u.id, err = tenancy.ParseUserID(info.UserID); err != nil {
		t.Fatal(err)
	}
	u.session = info.SessionID
	return u
}

// relogin issues another session of the same user and returns its credential.
func (e *itEnv) relogin(t *testing.T, u *itUser) string {
	t.Helper()
	return e.signInWith(t, e.id, u.sub, "Again", nil, "", "")
}

func (e *itEnv) csrf(t *testing.T, cred string) string {
	t.Helper()
	info, err := e.id.Session(context.Background(), cred, e.originOf(cred))
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(info.CSRFToken)
}

func (e *itEnv) logout(t *testing.T, cred, scope string) {
	t.Helper()
	if _, err := e.id.Logout(context.Background(), identity.LogoutInput{Scope: scope, CSRF: e.csrf(t, cred), Origin: e.originOf(cred), Credential: cred}); err != nil {
		t.Fatalf("logout %s: %v", scope, err)
	}
}

// --- tenancy fixtures --------------------------------------------------------------------------------------------------

func (e *itEnv) org(t *testing.T, name string) tenancy.OrganizationID {
	t.Helper()
	o, err := e.model.CreateOrganization(context.Background(), name)
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	return o.ID
}

func (e *itEnv) group(t *testing.T, org tenancy.OrganizationID, creator *itUser, name, slug string) tenancy.Group {
	t.Helper()
	g, _, err := e.model.CreateGroup(context.Background(), org, creator.id, name, slug)
	if err != nil {
		t.Fatalf("create group %s: %v", slug, err)
	}
	return g
}

func (e *itEnv) profile(t *testing.T, org tenancy.OrganizationID, name string) tenancy.BusinessProfile {
	t.Helper()
	legal, desc := name+" Legal", "Context "+name
	p, err := e.model.CreateBusinessProfile(context.Background(), org, tenancy.BusinessContext{DisplayName: name, LegalName: &legal, Description: &desc})
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	return p
}

func (e *itEnv) link(t *testing.T, org tenancy.OrganizationID, g tenancy.GroupID, p tenancy.BusinessProfileID) {
	t.Helper()
	if _, err := e.model.LinkGroupBusinessProfile(context.Background(), org, g, p); err != nil {
		t.Fatalf("link: %v", err)
	}
}

func (e *itEnv) grantOrg(t *testing.T, org tenancy.OrganizationID, u *itUser, r tenancy.OrganizationRole) {
	t.Helper()
	if _, err := e.model.GrantOrganizationRole(context.Background(), org, u.id, r); err != nil {
		t.Fatalf("grant %s: %v", r, err)
	}
}

func (e *itEnv) grantGroup(t *testing.T, org tenancy.OrganizationID, g tenancy.GroupID, u *itUser, r tenancy.GroupRole) {
	t.Helper()
	if _, err := e.model.GrantGroupRole(context.Background(), org, g, u.id, r); err != nil {
		t.Fatalf("grant %s: %v", r, err)
	}
}

// setFlag assigns an Organization's policy flag directly (owner SQL), for fixtures only.
func (e *itEnv) setFlag(t *testing.T, org tenancy.OrganizationID, on bool) {
	t.Helper()
	if _, err := e.setup.Exec(context.Background(), `UPDATE app.organizations SET group_admins_edit_business_profiles = $2 WHERE id = $1::text::uuid`,
		org.String(), on); err != nil {
		t.Fatal(err)
	}
}

// withOverride has owner turn org's Group-management override on through the facade (G060: the Organization path of
// Group and Business Profile administration requires it) and returns the matching turn-off; the override is also
// turned off when the calling (sub)test ends (an absolute set: a repeated off records nothing). Call it while the
// hook is unarmed.
func (e *itEnv) withOverride(t *testing.T, owner *itUser, org tenancy.OrganizationID) (off func()) {
	t.Helper()
	set := func(on bool) {
		t.Helper()
		c, tc := e.as(t, owner, org)
		if v, err := e.svc.SetGroupManagementOverride(c, tc, on); err != nil || v.OverrideEnabled != on {
			t.Fatalf("owner sets the Group-management override to %t: %+v %v", on, v, err)
		}
	}
	set(true)
	t.Cleanup(func() { set(false) })
	return func() { set(false) }
}

// --- row-set hashes ----------------------------------------------------------------------------------------------------

// state hashes every fixture row of the given Organizations: the Organization rows (with the flag), Groups, bindings,
// profiles and associations.
func (e *itEnv) state(t *testing.T, orgs ...tenancy.OrganizationID) string {
	t.Helper()
	ids := make([]string, len(orgs))
	for i, o := range orgs {
		ids[i] = o.String()
	}
	var s string
	err := e.setup.QueryRow(context.Background(), `
SELECT coalesce((SELECT md5(string_agg(o::text, E'\n' ORDER BY o.id)) FROM app.organizations o WHERE o.id::text = ANY ($1::text[])), '-') || '|' ||
       coalesce((SELECT md5(string_agg(g::text, E'\n' ORDER BY g.id)) FROM app.groups g WHERE g.organization_id::text = ANY ($1::text[])), '-') || '|' ||
       coalesce((SELECT md5(string_agg(b::text, E'\n' ORDER BY b.id)) FROM app.role_bindings b WHERE b.organization_id::text = ANY ($1::text[])), '-') || '|' ||
       coalesce((SELECT md5(string_agg(p::text, E'\n' ORDER BY p.id)) FROM app.business_profiles p WHERE p.organization_id::text = ANY ($1::text[])), '-') || '|' ||
       coalesce((SELECT md5(string_agg(a::text, E'\n' ORDER BY a.id)) FROM app.group_business_profiles a WHERE a.organization_id::text = ANY ($1::text[])), '-')`,
		ids).Scan(&s)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// part hashes one relation of the given Organizations (owner connection).
func (e *itEnv) part(t *testing.T, rel string, org tenancy.OrganizationID) string {
	t.Helper()
	var s string
	col := "organization_id"
	if rel == "organizations" {
		col = "id"
	}
	if err := e.setup.QueryRow(context.Background(), `SELECT coalesce(md5(string_agg(r::text, E'\n' ORDER BY r::text)), '-') FROM app.`+rel+
		` r WHERE r.`+col+` = $1::text::uuid`, org.String()).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func (e *itEnv) renewedAt(t *testing.T, session string) time.Time {
	t.Helper()
	var at time.Time
	if err := e.setup.QueryRow(context.Background(), `SELECT renewed_at FROM app.sessions WHERE id = $1::text::uuid`, session).Scan(&at); err != nil {
		t.Fatal(err)
	}
	return at
}

// itH is what a failure message prints for a value that may carry a name: a short hash prefix, never the value.
func itH(v any) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%+v", v))))[:12] }

// lastRecord returns the last access.decision record.
func (e *itEnv) lastRecord(t *testing.T) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(e.logs.String()), "\n")
	var rec map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &rec); err != nil {
		t.Fatalf("decision record: %v", err)
	}
	return rec
}

func (e *itEnv) lastReason(t *testing.T) string {
	t.Helper()
	r, _ := e.lastRecord(t)["reason"].(string)
	return r
}

// --- request helpers ---------------------------------------------------------------------------------------------------

// begin starts one request scope of svc for cred under base; the scope is ended at cleanup.
func begin(t *testing.T, svc *Service, base context.Context, cred string) (context.Context, func()) {
	t.Helper()
	rctx, end, err := svc.BeginRequest(base, cred, itOrigin)
	if err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	t.Cleanup(end)
	return rctx, end
}

// beginAs starts one request scope of svc for u's credential prepared for org, at org's own origin (G016 A2: the scope
// is confined to org); the scope is ended at cleanup.
func beginAs(t *testing.T, svc *Service, base context.Context, u *itUser, org tenancy.OrganizationID) (context.Context, func()) {
	t.Helper()
	rctx, end, err := svc.BeginRequest(base, u.in(t, org), itOriginOf(org))
	if err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	t.Cleanup(end)
	return rctx, end
}

func resolve(t *testing.T, svc *Service, rctx context.Context, org tenancy.OrganizationID, g tenancy.GroupID) *TenantContext {
	t.Helper()
	req := ResolveRequest{OrganizationID: org.String()}
	if !zero(g) {
		req.GroupID = g.String()
	}
	tc, err := svc.Resolve(rctx, req)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	return tc
}

func wantCode(t *testing.T, what string, err error, code Code) {
	t.Helper()
	if CodeOf(err) != code || (code == "" && err != nil) {
		t.Errorf("%s: %v, want %q", what, err, code)
	}
}

// itOps are the eight guarded operations against Group g and profile p.
func itOps(s *Service, g tenancy.GroupID, p tenancy.BusinessProfileID) []facadeOp {
	return []facadeOp{
		{"GetOrganization", func(ctx context.Context, tc *TenantContext) error { _, err := s.GetOrganization(ctx, tc); return err }},
		{"GetOrganizationPolicy", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.GetOrganizationPolicy(ctx, tc)
			return err
		}},
		{"SetOrganizationPolicy", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.SetOrganizationPolicy(ctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: true})
			return err
		}},
		{"GetGroup", func(ctx context.Context, tc *TenantContext) error { _, err := s.GetGroup(ctx, tc, g); return err }},
		{"RenameGroup", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.RenameGroup(ctx, tc, g, "Renamed By Test")
			return err
		}},
		{"ChangeGroupSlug", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.ChangeGroupSlug(ctx, tc, g, "renamed-by-test")
			return err
		}},
		{"GetBusinessProfile", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.GetBusinessProfile(ctx, tc, p)
			return err
		}},
		{"UpdateBusinessProfile", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.UpdateBusinessProfile(ctx, tc, p, tenancy.BusinessContext{DisplayName: "Updated By Test"})
			return err
		}},
	}
}

// loadDecision evaluates capability c for tc on facts loaded from the database in one REPEATABLE READ READ ONLY
// snapshot with the facade's own fact query for the target kind (IT-MATRIX-1, IT-MFA-1).
func (e *itEnv) loadDecision(t *testing.T, tc *TenantContext, c Capability, tg Target) bool {
	t.Helper()
	ctx := context.Background()
	var q tenancy.FactQuery
	switch tg.Kind {
	case kindGroup:
		q = groupQuery(tg.Group, false)(tc)
	case kindProfile:
		q = profileQuery(tg.Profile)(tc)
	default:
		q = organizationQuery(tc)
	}
	q.Organization, q.User = tc.organization, tc.subject
	tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Version 6: the facts are read in the trusted context of tc, exactly as the facade reads them.
	if err := tenancy.InitContext(ctx, tx, tc.subject, tc.organization, tc.group); err != nil {
		t.Fatalf("context: %v", err)
	}
	f, err := tenancy.ReadFacts(ctx, tx, q)
	if err != nil {
		t.Fatalf("read facts: %v", err)
	}
	tg.Organization = tc.organization
	ok, _ := decide(c, tc.selection(), f, tg)
	return ok
}

// --- hookDB: the runtime pool, counting transactions, statements and locked rows, with step hooks and commit modes ----

type commitMode int

const (
	commitNormal       commitMode = iota
	commitDefinitive              // the transaction is rolled back and COMMIT reports a serialization failure (40001)
	commitLostReply               // COMMIT really commits, then the reply is lost (io.ErrUnexpectedEOF)
	commitLostNoCommit            // G009: the transaction is rolled back, then the COMMIT reply is lost (io.ErrUnexpectedEOF)
)

// The statement steps of the authorization contract (Plan r1 §4.1), recognized by their SQL.
var stepPatterns = []struct {
	step string
	re   *regexp.Regexp
}{
	{"C", regexp.MustCompile(`^SELECT pg_catalog\.set_config\('app\.user_id'`)}, // G008: the context statement, first
	{"U", regexp.MustCompile(`FROM app\.users WHERE id = \$1 AND status = 'active'`)},
	{"S", regexp.MustCompile(`FROM app\.sessions WHERE id = \$1 AND user_id = \$2`)},
	{"O", regexp.MustCompile(`^SELECT id, name, created_at, group_admins_edit_business_profiles, group_management_override FROM app\.organizations`)}, // G060: + the override column
	{"OB", regexp.MustCompile(`FROM app\.organization_memberships|FROM app\.role_bindings .*scope_type = 'organization'`)},
	{"GB", regexp.MustCompile(`FROM app\.group_memberships|FROM app\.role_bindings .*scope_type = 'group'`)},
	{"Gt", regexp.MustCompile(`FROM app\.groups WHERE organization_id = \$1 AND id = \$2`)},
	{"G", regexp.MustCompile(`FROM app\.groups WHERE organization_id = \$1 AND id = ANY`)},
	{"A", regexp.MustCompile(`FROM app\.group_business_profiles WHERE organization_id = \$1 AND group_id = ANY`)},
	{"P", regexp.MustCompile(`FROM app\.business_profiles WHERE organization_id = \$1 AND id = \$2`)},
	{"PALL", regexp.MustCompile(`FROM app\.business_profiles WHERE organization_id = \$1 ORDER BY id`)},
	{"W", regexp.MustCompile(`^UPDATE app\.`)},
	// G018: the keyed onboarding inserts (S-G1, S-P1, S-P2), so a fault or pause can name each statement.
	{"IG", regexp.MustCompile(`^INSERT INTO app\.groups \(id,`)},
	{"IP", regexp.MustCompile(`^INSERT INTO app\.business_profiles \(id,`)},
	{"IA", regexp.MustCompile(`^INSERT INTO app\.group_business_profiles `)},
	// G019: a binding removal (leave, removal, transfer to none, departure), so an ordering can pause before it (O-MEM-1).
	{"D", regexp.MustCompile(`^DELETE FROM app\.role_bindings `)},
}

func stepOf(sql string) string {
	for _, p := range stepPatterns {
		if p.re.MatchString(sql) {
			return p.step
		}
	}
	return "?"
}

var lockingClause = regexp.MustCompile(`\bFOR (SHARE|NO KEY UPDATE|UPDATE|KEY SHARE)\b`)

// lockedStmt is one locking statement and the number of rows it returned (and so locked).
type lockedStmt struct {
	step string
	mode string
	rows int
}

type hookDB struct {
	pool *pgxpool.Pool

	mu      sync.Mutex
	begins  int
	stmts   []string // steps, in order
	locked  []lockedStmt
	before  map[string]func()
	after   map[string]func()
	failOn  string
	commit  commitMode
	commits int
	// capture records every statement's text and arguments (E-PG-1 replays them under EXPLAIN).
	capture  bool
	captured []capturedStmt
}

type capturedStmt struct {
	step string
	sql  string
	args []any
}

func (h *hookDB) capturedStmts() []capturedStmt {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.captured)
}

// reset clears the counters, hooks and modes.
func (h *hookDB) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.begins, h.stmts, h.locked, h.before, h.after, h.failOn, h.commit, h.commits = 0, nil, nil, nil, nil, "", commitNormal, 0
	h.capture, h.captured = false, nil
}

func (h *hookDB) counts() (begins int, stmts []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.begins, slices.Clone(h.stmts)
}

func (h *hookDB) lockedRows() []lockedStmt {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.locked)
}

// on sets a hook to run before or after the statement of step.
func (h *hookDB) on(when, step string, fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	m := &h.before
	if when == "after" {
		m = &h.after
	}
	if *m == nil {
		*m = map[string]func(){}
	}
	(*m)[step] = fn
}

// BeginTx opens the transaction and then runs the "BEGIN" before-hook: a pause there is after BEGIN and before the
// operation's first statement (and before identity samples its clock for the session predicate).
func (h *hookDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	h.mu.Lock()
	h.begins++
	begin := h.before["BEGIN"]
	h.mu.Unlock()
	tx, err := h.pool.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}
	if begin != nil {
		begin()
	}
	return &hookTx{Tx: tx, h: h}, nil
}

type hookTx struct {
	pgx.Tx
	h *hookDB
}

var errHookInjected = &pgconn.PgError{Code: "08006", Message: "injected connection failure"}

// enter records a statement, runs its before hook and reports whether it is replaced by an injected failure.
func (t *hookTx) enter(sql string, args []any) (step string, injected bool) {
	step = stepOf(sql)
	t.h.mu.Lock()
	t.h.stmts = append(t.h.stmts, step)
	if t.h.capture {
		t.h.captured = append(t.h.captured, capturedStmt{step: step, sql: sql, args: slices.Clone(args)})
	}
	before, fail := t.h.before[step], t.h.failOn == step
	t.h.mu.Unlock()
	if before != nil {
		before()
	}
	return step, fail
}

func (t *hookTx) leave(step, sql string, rows int) {
	t.h.mu.Lock()
	if m := lockingClause.FindStringSubmatch(sql); m != nil {
		t.h.locked = append(t.h.locked, lockedStmt{step: step, mode: m[1], rows: rows})
	}
	after := t.h.after[step]
	t.h.mu.Unlock()
	if after != nil {
		after()
	}
}

type hookRow struct {
	row  pgx.Row
	done func(rows int)
}

func (r hookRow) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	n := 0
	if err == nil {
		n = 1
	}
	r.done(n)
	return err
}

type errRow struct{ err error }

func (r errRow) Scan(...any) error { return r.err }

func (t *hookTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	step, fail := t.enter(sql, args)
	if fail {
		return errRow{errHookInjected}
	}
	return hookRow{row: t.Tx.QueryRow(ctx, sql, args...), done: func(n int) { t.leave(step, sql, n) }}
}

type hookRows struct {
	pgx.Rows
	n    int
	done func(rows int)
	once sync.Once
}

func (r *hookRows) Next() bool {
	if r.Rows.Next() {
		r.n++
		return true
	}
	r.once.Do(func() { r.done(r.n) })
	return false
}

func (t *hookTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	step, fail := t.enter(sql, args)
	if fail {
		return nil, errHookInjected
	}
	rows, err := t.Tx.Query(ctx, sql, args...)
	if err != nil {
		return rows, err
	}
	return &hookRows{Rows: rows, done: func(n int) { t.leave(step, sql, n) }}, nil
}

func (t *hookTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	step, fail := t.enter(sql, args)
	if fail {
		return pgconn.CommandTag{}, errHookInjected
	}
	tag, err := t.Tx.Exec(ctx, sql, args...)
	t.leave(step, sql, int(tag.RowsAffected()))
	return tag, err
}

func (t *hookTx) Commit(ctx context.Context) error {
	t.h.mu.Lock()
	mode := t.h.commit
	t.h.commits++
	before := t.h.before["COMMIT"]
	t.h.mu.Unlock()
	if before != nil {
		before()
	}
	switch mode {
	case commitDefinitive:
		// A server rejecting COMMIT with an ordinary command error sends severity ERROR (nonlocalized and localized).
		_ = t.Tx.Rollback(ctx)
		return &pgconn.PgError{Severity: "ERROR", SeverityUnlocalized: "ERROR", Code: "40001", Message: "injected serialization failure at commit"}
	case commitLostReply:
		if err := t.Tx.Commit(ctx); err != nil {
			return err
		}
		return io.ErrUnexpectedEOF
	case commitLostNoCommit:
		_ = t.Tx.Rollback(ctx)
		return io.ErrUnexpectedEOF
	}
	return t.Tx.Commit(ctx)
}

// waitWaiters waits until at least n backends of this database wait for a lock (pg_locks joined to pg_stat_activity,
// both readable by the runtime role), the accepted waitBlocked synchronization.
func (e *itEnv) waitWaiters(t *testing.T, n int) {
	t.Helper()
	obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if e.waiters(t, obs) >= n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("fewer than %d lock waits observed", n)
}

func (e *itEnv) waiters(t *testing.T, obs *pgx.Conn) int {
	t.Helper()
	var c int
	if err := obs.QueryRow(context.Background(), `SELECT count(DISTINCT l.pid) FROM pg_locks l JOIN pg_stat_activity a ON a.pid = l.pid
WHERE NOT l.granted AND a.datname = current_database() AND a.pid <> pg_backend_pid()`).Scan(&c); err != nil {
		t.Fatal(err)
	}
	return c
}

// --- the fixture world (Plan r1 §6.3) ----------------------------------------------------------------------------------

// itWorld: Organizations A and B; Groups g1, g2, g3 and the new unattached gNew in A, gB in B; profiles p1 (shared by
// g1 and g2), p2 (on g3), p3 (unlinked) in A and pB (on gB) in B. Users: owner, admin and member of A; gOwner, gAdmin
// and gMember (org_member of A plus that role in g1); multi (org_member of A, group_admin of g3, org_owner of B);
// gOnly (group_admin of g1 without an Organization binding); unattached (org_member, creator and owner of gNew);
// foreign (org_owner of B); creator (creates g1-g3, holds no Organization binding).
type itWorld struct {
	A, B                                          tenancy.OrganizationID
	g1, g2, g3, gNew, gB                          tenancy.Group
	p1, p2, p3, pB                                tenancy.BusinessProfile
	owner, admin, member, gOwner, gAdmin, gMember *itUser
	multi, gOnly, unattached, foreign, creator    *itUser
}

func (e *itEnv) world(t *testing.T) *itWorld {
	t.Helper()
	c := e.canaries
	w := &itWorld{}
	w.A, w.B = e.org(t, "Access A "+c[0]), e.org(t, "Access B "+c[1])
	for _, u := range []**itUser{&w.owner, &w.admin, &w.member, &w.gOwner, &w.gAdmin, &w.gMember, &w.multi, &w.gOnly, &w.unattached,
		&w.foreign, &w.creator} {
		*u = e.newUser(t)
	}
	w.g1 = e.group(t, w.A, w.creator, "Group One "+c[2], "group-one")
	w.g2 = e.group(t, w.A, w.creator, "Group Two "+c[0], "group-two")
	w.g3 = e.group(t, w.A, w.creator, "Group Three "+c[1], "group-three")
	w.gNew = e.group(t, w.A, w.unattached, "Group New "+c[2], "group-new")
	w.gB = e.group(t, w.B, w.foreign, "Group B "+c[0], "group-one")
	e.grantOrg(t, w.A, w.owner, tenancy.OrgOwner)
	e.grantOrg(t, w.A, w.admin, tenancy.OrgAdmin)
	for _, u := range []*itUser{w.member, w.gOwner, w.gAdmin, w.gMember, w.multi, w.unattached} {
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
	}
	e.grantGroup(t, w.A, w.g1.ID, w.gOwner, tenancy.GroupOwner)
	e.grantGroup(t, w.A, w.g1.ID, w.gAdmin, tenancy.GroupAdmin)
	e.grantGroup(t, w.A, w.g1.ID, w.gMember, tenancy.GroupMember)
	e.grantGroup(t, w.A, w.g3.ID, w.multi, tenancy.GroupAdmin)
	e.grantOrg(t, w.B, w.multi, tenancy.OrgOwner)
	e.grantGroup(t, w.A, w.g1.ID, w.gOnly, tenancy.GroupAdmin)
	e.grantOrg(t, w.B, w.foreign, tenancy.OrgOwner)
	w.p1, w.p2, w.p3 = e.profile(t, w.A, "Profile One "+c[1]), e.profile(t, w.A, "Profile Two "+c[2]), e.profile(t, w.A, "Profile Three "+c[0])
	w.pB = e.profile(t, w.B, "Profile B "+c[1])
	e.link(t, w.A, w.g1.ID, w.p1.ID)
	e.link(t, w.A, w.g2.ID, w.p1.ID)
	e.link(t, w.A, w.g3.ID, w.p2.ID)
	e.link(t, w.B, w.gB.ID, w.pB.ID)
	// G016 A2: one prepared credential per (user, Organization) at that Organization's origin. gOnly (no Organization
	// binding), foreign at A and owner at B are membership-negative: admitted only for the issuance (see bind).
	for _, u := range []*itUser{w.owner, w.admin, w.member, w.gOwner, w.gAdmin, w.gMember, w.gOnly, w.unattached, w.foreign} {
		e.bind(t, u, w.A)
	}
	e.bind(t, w.multi, w.A, w.B)
	e.bind(t, w.foreign, w.B)
	e.bind(t, w.owner, w.B)
	return w
}

// --- the policy test -----------------------------------------------------------------------------------------------

func TestAccessPolicy(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	svc := e.svc

	t.Run("runtime identity", func(t *testing.T) {
		var cur, sess string
		var super, bypass bool
		var memberships int
		err := e.pool.QueryRow(ctx, `SELECT current_user::text, session_user::text, r.rolsuper, r.rolbypassrls,
  (SELECT count(*) FROM pg_auth_members m WHERE m.member = r.oid)::int FROM pg_roles r WHERE r.rolname = current_user`).Scan(&cur, &sess, &super, &bypass, &memberships)
		if err != nil {
			t.Fatal(err)
		}
		if cur != e.cfg.RuntimeRole() || sess != cur || super || bypass || memberships != 0 {
			t.Fatalf("access pool: current=%s session=%s super=%v bypassrls=%v memberships=%d", cur, sess, super, bypass, memberships)
		}
	})

	w := e.world(t)
	var zeroG tenancy.GroupID

	t.Run("IT-CTX-1 positives for every role", func(t *testing.T) {
		before := e.state(t, w.A, w.B)
		for _, u := range []*itUser{w.owner, w.admin, w.member, w.gOwner, w.gAdmin, w.gMember} {
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			tc := resolve(t, svc, rctx, w.A, zeroG)
			if tc.Subject() != u.id || tc.Organization() != w.A {
				t.Fatal("resolved subject or Organization differs")
			}
			if _, ok := tc.Group(); ok {
				t.Fatal("Organization-only context reports a Group")
			}
			if o, err := svc.GetOrganization(rctx, tc); err != nil || o.ID != w.A {
				t.Errorf("GetOrganization: %v", err)
			}
		}
		// Selected-Group context: an Organization owner or admin may select any Group of A; a Group role only its Group.
		for _, u := range []*itUser{w.owner, w.admin} {
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			for _, g := range []tenancy.Group{w.g1, w.g2, w.g3, w.gNew} {
				tc := resolve(t, svc, rctx, w.A, g.ID)
				if sel, ok := tc.Group(); !ok || sel != g.ID {
					t.Errorf("selected %s", itH(g.ID))
				}
			}
		}
		for _, u := range []*itUser{w.gOwner, w.gAdmin, w.gMember} {
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			resolve(t, svc, rctx, w.A, w.g1.ID)
			for _, g := range []tenancy.GroupID{w.g2.ID, w.g3.ID, w.gNew.ID} {
				_, err := svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String(), GroupID: g.String()})
				wantCode(t, "Group role selects another Group", err, CodeContextDenied)
			}
		}
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-CTX-2 unauthenticated sessions at BeginRequest", func(t *testing.T) {
		before := e.state(t, w.A, w.B)
		wantUnauth := func(what string, s *Service, cred string) {
			t.Helper()
			_, end, err := s.BeginRequest(ctx, cred, itOrigin)
			end()
			wantCode(t, what, err, CodeUnauthenticated)
		}
		wantUnauth("absent", svc, "")
		wantUnauth("garbage", svc, "not-a-credential")
		wantUnauth("well-formed unknown", svc, base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
		// Expired: the same sessions seen by an identity service whose clock is past the idle expiry.
		late := newITClock()
		late.t = e.clk.Now().Add(itIdle + time.Minute)
		lateSvc, _ := e.service(e.identity(t, late, e.pool))
		u := e.newUser(t)
		wantUnauth("expired", lateSvc, u.cred)
		// Revoked (logout), logout-all (both sessions) and disabled.
		r := e.newUser(t)
		e.logout(t, r.cred, identity.ScopeCurrent)
		wantUnauth("revoked", svc, r.cred)
		a := e.newUser(t)
		second := e.relogin(t, a)
		e.logout(t, a.cred, identity.ScopeAll)
		wantUnauth("logout-all first", svc, a.cred)
		wantUnauth("logout-all second", svc, second)
		d := e.newUser(t)
		if err := e.id.DisableUser(ctx, d.id.String()); err != nil {
			t.Fatal(err)
		}
		wantUnauth("disabled", svc, d.cred)
		if rec := e.lastRecord(t); rec["operation"] != opBeginRequest || rec["outcome"] != outcomeDeny || rec["reason"] != string(reasonSessionInvalid) {
			t.Errorf("BeginRequest denial record: %v", rec)
		}
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-CTX-3 revocation after BeginRequest at Resolve", func(t *testing.T) {
		for name, revoke := range map[string]func(u *itUser){
			"logout":  func(u *itUser) { e.logout(t, u.in(t, w.A), identity.ScopeCurrent) },
			"disable": func(u *itUser) { _ = e.id.DisableUser(ctx, u.id.String()) },
		} {
			u := e.newUser(t)
			e.grantOrg(t, w.A, u, tenancy.OrgAdmin)
			e.bind(t, u, w.A)
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			resolve(t, svc, rctx, w.A, zeroG)
			revoke(u)
			_, err := svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String()})
			wantCode(t, name+": Resolve in the same scope", err, CodeUnauthenticated)
		}
	})

	t.Run("IT-CTX-4 provider claims and role-like names grant nothing", func(t *testing.T) {
		// The fake's default claims (org_id, role admin, roles [admin], permissions [all]) plus explicit ones, and a
		// display name made of role words: none of them is an input.
		n := itSeq.Add(1)
		u := &itUser{sub: fmt.Sprintf("user_01CLAIMS%08d", n), at: map[tenancy.OrganizationID]string{}}
		// G016 A2: one credential per Organization, each issued at that Organization's origin with the subject admitted
		// only for the issuance (membership-negative again before any assertion).
		for _, org := range []tenancy.OrganizationID{w.A, w.B} {
			u.at[org] = e.signInAt(t, e.id, org, u.sub, "org_owner admin", map[string]any{"org_id": w.A.String(), "role": "org_owner",
				"permissions": []string{"organization.manage_policy", "all"}, "groups": []string{w.g1.ID.String()}}, "", "")
		}
		info, err := e.id.Session(ctx, u.in(t, w.A), itOriginOf(w.A))
		if err != nil {
			t.Fatal(err)
		}
		u.id, _ = tenancy.ParseUserID(info.UserID)
		before := e.state(t, w.A, w.B)
		var rctx context.Context
		for _, org := range []tenancy.OrganizationID{w.A, w.B} {
			octx, _ := beginAs(t, svc, ctx, u, org)
			if org == w.A {
				rctx = octx
			}
			_, err := svc.Resolve(octx, ResolveRequest{OrganizationID: org.String()})
			wantCode(t, "claims-only subject", err, CodeContextDenied)
		}
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		_, err = svc.SetOrganizationPolicy(rctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
		wantCode(t, "role-claim member sets the policy", err, CodeForbidden)
		_, err = svc.GetGroup(rctx, tc, w.g1.ID)
		wantCode(t, "Group-claim member reads g1", err, CodeNotFound)
		if got, _ := svc.GetOrganizationPolicy(rctx, tc); !got.GroupAdminsEditBusinessProfiles {
			t.Fatal("the flag changed")
		}
	})

	t.Run("IT-CTX-5 selectors", func(t *testing.T) {
		before := e.state(t, w.A, w.B)
		rctx, _ := beginAs(t, svc, ctx, w.gMember, w.A)
		for name, req := range map[string]ResolveRequest{
			"malformed": {OrganizationID: "acme"}, "zero": {OrganizationID: "00000000-0000-0000-0000-000000000000"},
			"braced": {OrganizationID: "{" + w.A.String() + "}"}, "unhyphenated": {OrganizationID: strings.ReplaceAll(w.A.String(), "-", "")},
			"malformed Group": {OrganizationID: w.A.String(), GroupID: "g1"}, "zero Group": {OrganizationID: w.A.String(), GroupID: "00000000-0000-0000-0000-000000000000"},
		} {
			_, err := svc.Resolve(rctx, req)
			wantCode(t, name, err, CodeInvalidInput)
		}
		unknownOrg, unknownGroup := "0f8fad5b-d9cb-469f-a165-70867728950e", "7c9e6679-7425-40de-944b-e07fc1f90ae7"
		var msgs []string
		for name, req := range map[string]ResolveRequest{
			"unknown Organization": {OrganizationID: unknownOrg}, "foreign Organization": {OrganizationID: w.B.String()},
			"unknown Group": {OrganizationID: w.A.String(), GroupID: unknownGroup}, "foreign Group": {OrganizationID: w.A.String(), GroupID: w.gB.ID.String()},
			"non-member Group": {OrganizationID: w.A.String(), GroupID: w.g2.ID.String()},
		} {
			_, err := svc.Resolve(rctx, req)
			wantCode(t, name, err, CodeContextDenied)
			if err != nil {
				msgs = append(msgs, err.Error())
			}
		}
		slices.Sort(msgs)
		if len(slices.Compact(msgs)) != 1 {
			t.Errorf("unknown, foreign and non-member selectors are distinguishable: %v", msgs)
		}
		// An Organization admin may select any Group of A, never a foreign one.
		actx, _ := beginAs(t, svc, ctx, w.admin, w.A)
		resolve(t, svc, actx, w.A, w.gNew.ID)
		_, err := svc.Resolve(actx, ResolveRequest{OrganizationID: w.A.String(), GroupID: w.gB.ID.String()})
		wantCode(t, "admin selects a foreign Group", err, CodeContextDenied)
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-CTX-6 Group binding without Organization binding", func(t *testing.T) {
		before := e.state(t, w.A, w.B)
		rctx, _ := beginAs(t, svc, ctx, w.gOnly, w.A)
		for _, g := range []tenancy.GroupID{zeroG, w.g1.ID} {
			req := ResolveRequest{OrganizationID: w.A.String()}
			if !zero(g) {
				req.GroupID = g.String()
			}
			_, err := svc.Resolve(rctx, req)
			wantCode(t, "Group-binding-only subject", err, CodeContextDenied)
		}
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed: a membership was minted")
		}
	})

	t.Run("IT-CTX-7 cross-request, same subject", func(t *testing.T) {
		for name, bases := range map[string][2]context.Context{
			"equal inbound X-Request-Id": {requestCtx(t, "it-ctx7-shared-id"), requestCtx(t, "it-ctx7-shared-id")},
			"absent request IDs":         {ctx, ctx},
		} {
			before := e.state(t, w.A, w.B)
			r1, _ := beginAs(t, svc, bases[0], w.owner, w.A)
			r2, _ := beginAs(t, svc, bases[1], w.owner, w.A)
			tc1, tc2 := resolve(t, svc, r1, w.A, zeroG), resolve(t, svc, r2, w.A, zeroG)
			for _, op := range itOps(svc, w.g1.ID, w.p3.ID) {
				for dir, call := range map[string]func() error{
					"R1 under R2": func() error { return op.call(r2, tc1) }, "R2 under R1": func() error { return op.call(r1, tc2) },
				} {
					e.hook.reset()
					err := call()
					begins, _ := e.hook.counts()
					if CodeOf(err) != CodeContextDenied || e.lastReason(t) != string(reasonScopeMismatch) || begins != 0 {
						t.Errorf("%s %s (%s): %v, reason %s, %d BeginTx", op.name, dir, name, err, e.lastReason(t), begins)
					}
				}
			}
			if e.state(t, w.A, w.B) != before {
				t.Fatalf("%s: rows changed", name)
			}
		}
	})

	t.Run("IT-CTX-8 cross-subject", func(t *testing.T) {
		for name, bases := range map[string][2]context.Context{
			"equal request IDs":  {requestCtx(t, "it-ctx8-equal-id"), requestCtx(t, "it-ctx8-equal-id")},
			"absent request IDs": {ctx, ctx},
		} {
			before := e.state(t, w.A, w.B)
			ra, _ := beginAs(t, svc, bases[0], w.owner, w.A)
			rb, _ := beginAs(t, svc, bases[1], w.gMember, w.A)
			tca, tcb := resolve(t, svc, ra, w.A, zeroG), resolve(t, svc, rb, w.A, zeroG)
			cross := map[string]func() error{
				"A read under B":  func() error { _, err := svc.GetOrganization(rb, tca); return err },
				"A group under B": func() error { _, err := svc.GetGroup(rb, tca, w.g2.ID); return err },
				"A write under B": func() error {
					_, err := svc.SetOrganizationPolicy(rb, tca, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
					return err
				},
				"B read under A": func() error { _, err := svc.GetBusinessProfile(ra, tcb, w.p1.ID); return err },
			}
			for what, call := range cross {
				e.hook.reset()
				err := call()
				begins, _ := e.hook.counts()
				if CodeOf(err) != CodeContextDenied || begins != 0 {
					t.Errorf("%s (%s): %v with %d BeginTx", what, name, err, begins)
				}
			}
			if _, err := svc.GetOrganization(ra, tca); err != nil {
				t.Errorf("A under its own scope: %v", err)
			}
			if _, err := svc.GetBusinessProfile(rb, tcb, w.p1.ID); err != nil {
				t.Errorf("B under its own scope: %v", err)
			}
			if e.state(t, w.A, w.B) != before {
				t.Fatalf("%s: rows changed", name)
			}
		}
	})

	t.Run("IT-CTX-9 legitimate same request", func(t *testing.T) {
		// G060: the owner's update of the unlinked p3 and rename of gNew (no Group role) are Organization-path writes,
		// admitted only with the Group-management override (explicit setup; the OFF denials are IT-FLAG-4/IT-REVOKE-8 and
		// the authority suites).
		e.withOverride(t, w.owner, w.A)
		rctx, _ := beginAs(t, svc, requestCtx(t, "it-ctx9-request"), w.owner, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		if o, err := svc.GetOrganization(rctx, tc); err != nil || o.ID != w.A {
			t.Fatalf("GetOrganization: %v", err)
		}
		desc := "Same request " + e.canaries[2]
		if p, err := svc.UpdateBusinessProfile(rctx, tc, w.p3.ID, tenancy.BusinessContext{DisplayName: w.p3.Context.DisplayName, Description: &desc}); err != nil ||
			p.Context.Description == nil || *p.Context.Description != desc {
			t.Fatalf("UpdateBusinessProfile: %v", err)
		}
		deadline, cancel := context.WithTimeout(rctx, time.Minute)
		defer cancel()
		if p, err := svc.GetBusinessProfile(deadline, tc, w.p3.ID); err != nil || p.Context.Description == nil || *p.Context.Description != desc {
			t.Fatalf("GetBusinessProfile under a deadline child: %v", err)
		}
		tcg := resolve(t, svc, rctx, w.A, w.gNew.ID)
		if g, err := svc.RenameGroup(rctx, tcg, w.gNew.ID, "Group New Renamed"); err != nil || g.Name != "Group New Renamed" {
			t.Fatalf("RenameGroup: %v", err)
		}
		if g, err := svc.GetGroup(rctx, tcg, w.gNew.ID); err != nil || g.Name != "Group New Renamed" {
			t.Fatalf("GetGroup: %v", err)
		}
	})

	t.Run("IT-CTX-10 outside or after the request", func(t *testing.T) {
		before := e.state(t, w.A, w.B)
		rctx, end := beginAs(t, svc, ctx, w.admin, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		check := func(what string, reason Reason, call func() error) {
			t.Helper()
			e.hook.reset()
			err := call()
			begins, _ := e.hook.counts()
			if CodeOf(err) != CodeContextDenied || e.lastReason(t) != string(reason) || begins != 0 {
				t.Errorf("%s: %v, reason %s, %d BeginTx", what, err, e.lastReason(t), begins)
			}
		}
		for _, op := range itOps(svc, w.g1.ID, w.p3.ID) {
			check(op.name+" under Background", reasonScopeMissing, func() error { return op.call(ctx, tc) })
		}
		adminA := w.admin.in(t, w.A)
		renewed := e.renewedAt(t, func() string { info, _ := e.id.Session(ctx, adminA, itOriginOf(w.A)); return info.SessionID }())
		check("nested BeginRequest", reasonScopeNested, func() error {
			_, nend, err := svc.BeginRequest(rctx, adminA, itOriginOf(w.A))
			nend()
			return err
		})
		info, _ := e.id.Session(ctx, adminA, itOriginOf(w.A))
		if e.renewedAt(t, info.SessionID) != renewed {
			t.Error("nested BeginRequest touched the session row")
		}
		end()
		check("Resolve after end", reasonScopeEnded, func() error {
			_, err := svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String()})
			return err
		})
		for _, op := range itOps(svc, w.g1.ID, w.p3.ID) {
			check(op.name+" after end", reasonScopeEnded, func() error { return op.call(rctx, tc) })
			check(op.name+" after end (WithoutCancel)", reasonScopeEnded, func() error { return op.call(context.WithoutCancel(rctx), tc) })
		}
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-CTX-11 a browser scope resolves only its session's Organization (G016 A2)", func(t *testing.T) {
		// multi holds bindings of both A and B: each Organization is a separate request with its own credential; the
		// A scope never resolves B (organization_not_permitted before any snapshot) and vice versa.
		before := e.state(t, w.A, w.B)
		actx, _ := beginAs(t, svc, ctx, w.multi, w.A)
		if tc := resolve(t, svc, actx, w.A, zeroG); tc.Organization() != w.A {
			t.Fatal("A scope resolved another Organization")
		}
		_, err := svc.Resolve(actx, ResolveRequest{OrganizationID: w.B.String()})
		wantCode(t, "A scope resolves B", err, CodeContextDenied)
		if e.lastReason(t) != string(reasonOrganizationNotPermitted) {
			t.Errorf("A scope resolves B: reason %s", e.lastReason(t))
		}
		bctx, _ := beginAs(t, svc, ctx, w.multi, w.B)
		if tc := resolve(t, svc, bctx, w.B, zeroG); tc.Organization() != w.B {
			t.Fatal("B scope resolved another Organization")
		}
		_, err = svc.Resolve(bctx, ResolveRequest{OrganizationID: w.A.String()})
		wantCode(t, "B scope resolves A", err, CodeContextDenied)
		// The A credential presented at B's origin is unauthenticated (no other credential is tried).
		_, end, err := svc.BeginRequest(ctx, w.multi.in(t, w.A), itOriginOf(w.B))
		end()
		wantCode(t, "A credential at B's origin", err, CodeUnauthenticated)
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-MATRIX-1 the independent table on DB-loaded facts", func(t *testing.T) {
		// The oracle delegate's literal table (matrix_test.go, written from Spec R3 only), applied to facts loaded from
		// PostgreSQL: one subject per column (an Organization role alone, or org_member plus the Group role in g1), the
		// targets A, g1 and p1, both context modes and both flag values.
		subjects := map[string]*itUser{}
		for _, col := range oracleColumns {
			u := e.newUser(t)
			switch tenancy.OrganizationRole(col) {
			case tenancy.OrgOwner, tenancy.OrgAdmin, tenancy.OrgMember:
				e.grantOrg(t, w.A, u, tenancy.OrganizationRole(col))
			default:
				e.grantOrg(t, w.A, u, tenancy.OrgMember)
				e.grantGroup(t, w.A, w.g1.ID, u, tenancy.GroupRole(col))
			}
			e.bind(t, u, w.A)
			subjects[col] = u
		}
		targets := map[TargetKind]Target{kindOrganization: {Kind: kindOrganization}, kindGroup: {Kind: kindGroup, Group: w.g1.ID},
			kindProfile: {Kind: kindProfile, Profile: w.p1.ID}}
		cells := 0
		for _, flag := range []bool{true, false} {
			e.setFlag(t, w.A, flag)
			for _, selected := range []bool{false, true} {
				for _, col := range oracleColumns {
					u := subjects[col]
					rctx, _ := beginAs(t, svc, ctx, u, w.A)
					sel := zeroG
					if selected {
						sel = w.g1.ID
					}
					req := ResolveRequest{OrganizationID: w.A.String()}
					if selected {
						req.GroupID = sel.String()
					}
					tc, err := svc.Resolve(rctx, req)
					if selected && col == string(tenancy.OrgMember) {
						wantCode(t, "org_member selects g1", err, CodeContextDenied)
						continue
					}
					if err != nil {
						t.Fatalf("%s selected=%v: %v", col, selected, err)
					}
					for _, c := range oracleCapabilities {
						want, asserted := oracleExpectPrimary(t, selected, flag, c.name, col)
						if !asserted {
							continue
						}
						cells++
						if got := e.loadDecision(t, tc, Capability(c.name), targets[c.kind]); got != want {
							t.Errorf("%s / %s (selected=%v, flag=%v): allow=%v, want %v", c.name, col, selected, flag, got, want)
						}
					}
					// The operation-backed reads agree with the table.
					reads := map[string]func() error{
						"organization.read":             func() error { _, err := svc.GetOrganization(rctx, tc); return err },
						"group.read":                    func() error { _, err := svc.GetGroup(rctx, tc, w.g1.ID); return err },
						"business_profile.profile.read": func() error { _, err := svc.GetBusinessProfile(rctx, tc, w.p1.ID); return err },
					}
					for c, call := range reads {
						want, _ := oracleExpectPrimary(t, selected, flag, c, col)
						if err := call(); (err == nil) != want || (err != nil && CodeOf(err) != CodeNotFound) {
							t.Errorf("facade %s / %s (selected=%v, flag=%v): %v, want allow=%v", c, col, selected, flag, err, want)
						}
					}
				}
			}
		}
		e.setFlag(t, w.A, true)
		// G018 adds 6+6+5+5 cells per capability (286 -> 330); G019's four capabilities likewise (330 -> 418); G020's three
		// label capabilities likewise (418 -> 484; this loop has no override-on tables, so not the unit matrix's 6+6+5+5+6+5).
		if cells != 484 {
			t.Errorf("asserted %d cells, want 484", cells)
		}
		// B is foreign: the A-owner's facts never reach B's Group or profile.
		rctx, _ := beginAs(t, svc, ctx, subjects[string(tenancy.OrgOwner)], w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		if e.loadDecision(t, tc, capGroupRead, Target{Kind: kindGroup, Group: w.gB.ID}) ||
			e.loadDecision(t, tc, capProfileRead, Target{Kind: kindProfile, Profile: w.pB.ID}) {
			t.Error("a foreign target was allowed")
		}
		// An unavailable store and a cancellation deny (never allow) and change nothing.
		before := e.state(t, w.A)
		e.hook.reset()
		e.hook.failOn = "OB"
		_, err := svc.GetOrganization(rctx, tc)
		wantCode(t, "store failure", err, CodeUnavailable)
		e.hook.reset()
		cctx, cancel := context.WithCancel(rctx)
		e.hook.on("after", "O", cancel)
		_, err = svc.UpdateBusinessProfile(cctx, tc, w.p3.ID, tenancy.BusinessContext{DisplayName: "Canceled"})
		wantCode(t, "cancellation", err, CodeUnavailable)
		if e.lastReason(t) != string(reasonCanceled) {
			t.Errorf("cancellation reason %s", e.lastReason(t))
		}
		cancel()
		e.hook.reset()
		if e.state(t, w.A) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-PROFILE-1 Group-derived management through either attached Group", func(t *testing.T) {
		g2admin := e.newUser(t)
		e.grantOrg(t, w.A, g2admin, tenancy.OrgMember)
		e.grantGroup(t, w.A, w.g2.ID, g2admin, tenancy.GroupAdmin)
		e.bind(t, g2admin, w.A)
		for _, u := range []*itUser{w.gOwner, w.gAdmin, g2admin} {
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			tc := resolve(t, svc, rctx, w.A, zeroG)
			if _, err := svc.UpdateBusinessProfile(rctx, tc, w.p1.ID, w.p1.Context); err != nil {
				t.Errorf("update p1: %v", err)
			}
			if !e.loadDecision(t, tc, capProfileSelect, Target{Kind: kindProfile, Profile: w.p1.ID}) {
				t.Error("select p1 denied")
			}
			for _, c := range []Capability{capProfilePrivileged} {
				if e.loadDecision(t, tc, c, Target{Kind: kindProfile, Profile: w.p1.ID}) {
					t.Errorf("%s allowed through a Group role", c)
				}
			}
			for _, c := range []Capability{capOrganizationManagePolicy, capAuditRead, capOrganizationManageMembers} {
				if e.loadDecision(t, tc, c, Target{Kind: kindOrganization}) {
					t.Errorf("%s allowed through a Group role", c)
				}
			}
			_, err := svc.SetOrganizationPolicy(rctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
			wantCode(t, "Group role sets the policy", err, CodeForbidden)
		}
		rctx, _ := beginAs(t, svc, ctx, w.gMember, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		before := e.state(t, w.A)
		if _, err := svc.GetBusinessProfile(rctx, tc, w.p1.ID); err != nil {
			t.Errorf("member read: %v", err)
		}
		_, err := svc.UpdateBusinessProfile(rctx, tc, w.p1.ID, tenancy.BusinessContext{DisplayName: "Member Edit"})
		wantCode(t, "member update", err, CodeForbidden)
		if e.loadDecision(t, tc, capProfileSelect, Target{Kind: kindProfile, Profile: w.p1.ID}) {
			t.Error("member may select p1")
		}
		if e.state(t, w.A) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-PROFILE-2 a selected Group restricts inheritance", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
		e.grantGroup(t, w.A, w.g1.ID, u, tenancy.GroupOwner)
		e.grantGroup(t, w.A, w.g3.ID, u, tenancy.GroupAdmin)
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, svc, ctx, u, w.A)
		sel := resolve(t, svc, rctx, w.A, w.g3.ID)
		before := e.state(t, w.A)
		_, err := svc.UpdateBusinessProfile(rctx, sel, w.p1.ID, tenancy.BusinessContext{DisplayName: "Through g1"})
		wantCode(t, "selected g3 updates p1 through g1", err, CodeNotFound)
		_, err = svc.GetBusinessProfile(rctx, sel, w.p1.ID)
		wantCode(t, "selected g3 reads p1", err, CodeNotFound)
		if e.state(t, w.A) != before {
			t.Fatal("rows changed")
		}
		if _, err := svc.GetBusinessProfile(rctx, sel, w.p2.ID); err != nil {
			t.Errorf("selected g3 reads p2: %v", err)
		}
		org := resolve(t, svc, rctx, w.A, zeroG)
		if _, err := svc.UpdateBusinessProfile(rctx, org, w.p1.ID, w.p1.Context); err != nil {
			t.Errorf("Organization-only update of p1 through g1: %v", err)
		}
	})

	t.Run("IT-PROFILE-3 the manageable set", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
		e.grantGroup(t, w.A, w.g1.ID, u, tenancy.GroupOwner)
		e.grantGroup(t, w.A, w.g2.ID, u, tenancy.GroupAdmin)
		e.grantGroup(t, w.A, w.g3.ID, u, tenancy.GroupMember)
		ids := func(ps []tenancy.BusinessProfile) []tenancy.BusinessProfileID {
			out := []tenancy.BusinessProfileID{}
			for _, p := range ps {
				out = append(out, p.ID)
			}
			slices.SortFunc(out, func(a, b tenancy.BusinessProfileID) int { return bytes.Compare(a[:], b[:]) })
			return out
		}
		sorted := func(ps ...tenancy.BusinessProfileID) []tenancy.BusinessProfileID {
			slices.SortFunc(ps, func(a, b tenancy.BusinessProfileID) int { return bytes.Compare(a[:], b[:]) })
			return ps
		}
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, svc, ctx, u, w.A)
		got, err := svc.manageableProfiles(rctx, resolve(t, svc, rctx, w.A, zeroG))
		if err != nil || !slices.Equal(ids(got), []tenancy.BusinessProfileID{w.p1.ID}) {
			t.Errorf("Organization-only manageable set: %v %v", itH(ids(got)), err)
		}
		got, err = svc.manageableProfiles(rctx, resolve(t, svc, rctx, w.A, w.g3.ID))
		if err != nil || len(got) != 0 {
			t.Errorf("selected g3 manageable set: %v %v", itH(ids(got)), err)
		}
		// G060 (r0 A4/D-5): the owner holds no Group role, so without the override its selectable set is empty; with the
		// override the Organization path selects every profile of A, as before G060.
		octx, _ := beginAs(t, svc, ctx, w.owner, w.A)
		got, err = svc.manageableProfiles(octx, resolve(t, svc, octx, w.A, w.g3.ID))
		if err != nil || len(got) != 0 {
			t.Errorf("owner manageable set, override off: %v %v", itH(ids(got)), err)
		}
		off := e.withOverride(t, w.owner, w.A)
		octx, _ = beginAs(t, svc, ctx, w.owner, w.A)
		got, err = svc.manageableProfiles(octx, resolve(t, svc, octx, w.A, w.g3.ID))
		if err != nil || !slices.Equal(ids(got), sorted(w.p1.ID, w.p2.ID, w.p3.ID)) {
			t.Errorf("owner manageable set, override on: %v %v", itH(ids(got)), err)
		}
		off()
		mctx, _ := beginAs(t, svc, ctx, w.gMember, w.A)
		if got, err := svc.manageableProfiles(mctx, resolve(t, svc, mctx, w.A, zeroG)); err != nil || len(got) != 0 {
			t.Errorf("member-only manageable set: %v %v", len(got), err)
		}
	})

	t.Run("IT-PROFILE-4 a new unattached Group authorizes no existing profile", func(t *testing.T) {
		rctx, _ := beginAs(t, svc, ctx, w.unattached, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		before := e.state(t, w.A)
		for _, p := range []tenancy.BusinessProfileID{w.p1.ID, w.p2.ID, w.p3.ID} {
			_, err := svc.UpdateBusinessProfile(rctx, tc, p, tenancy.BusinessContext{DisplayName: "Unattached"})
			wantCode(t, "unattached owner updates an unrelated profile", err, CodeNotFound)
		}
		if got, err := svc.manageableProfiles(rctx, tc); err != nil || len(got) != 0 {
			t.Errorf("manageable set: %d %v", len(got), err)
		}
		if e.state(t, w.A) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-PROFILE-5 a shared profile read reveals only the profile", func(t *testing.T) {
		rctx, _ := beginAs(t, svc, ctx, w.gMember, w.A)
		p, err := svc.GetBusinessProfile(rctx, resolve(t, svc, rctx, w.A, zeroG), w.p1.ID)
		if err != nil || p.ID != w.p1.ID || p.OrganizationID != w.A {
			t.Fatalf("read p1: %v", err)
		}
		rt := reflect.TypeOf(p)
		var fields []string
		for i := range rt.NumField() {
			fields = append(fields, rt.Field(i).Name)
		}
		if strings.Join(fields, ",") != "ID,OrganizationID,Context,CreatedAt" {
			t.Fatalf("profile result fields: %v", fields)
		}
	})

	t.Run("IT-FLAG-1 default true on pre-existing and new Organizations", func(t *testing.T) {
		rctx, _ := beginAs(t, svc, ctx, w.member, w.A)
		if pol, err := svc.GetOrganizationPolicy(rctx, resolve(t, svc, rctx, w.A, zeroG)); err != nil || !pol.GroupAdminsEditBusinessProfiles {
			t.Fatalf("A: %+v %v", pol, err)
		}
		n := e.org(t, "Fresh Organization "+e.canaries[1])
		u := e.newUser(t)
		e.grantOrg(t, n, u, tenancy.OrgMember)
		e.bind(t, u, n)
		nctx, _ := beginAs(t, svc, ctx, u, n)
		if pol, err := svc.GetOrganizationPolicy(nctx, resolve(t, svc, nctx, n, zeroG)); err != nil || !pol.GroupAdminsEditBusinessProfiles {
			t.Fatalf("new Organization: %+v %v", pol, err)
		}
	})

	t.Run("IT-FLAG-2 the owner sets false then true", func(t *testing.T) {
		rctx, _ := beginAs(t, svc, ctx, w.owner, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		for _, v := range []bool{false, false, true, true} {
			pol, err := svc.SetOrganizationPolicy(rctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: v})
			if err != nil || pol.GroupAdminsEditBusinessProfiles != v {
				t.Fatalf("set %v: %+v %v", v, pol, err)
			}
			if pol, err := svc.GetOrganizationPolicy(rctx, tc); err != nil || pol.GroupAdminsEditBusinessProfiles != v {
				t.Fatalf("get after set %v: %+v %v", v, pol, err)
			}
		}
	})

	t.Run("IT-FLAG-3 only the owner may set it; other tenants never", func(t *testing.T) {
		before := e.state(t, w.A, w.B)
		for _, u := range []*itUser{w.admin, w.member, w.gOwner, w.gAdmin, w.gMember} {
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			_, err := svc.SetOrganizationPolicy(rctx, resolve(t, svc, rctx, w.A, zeroG), tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
			wantCode(t, "non-owner sets the flag", err, CodeForbidden)
		}
		// B's owner cannot resolve A; A's owner cannot resolve B.
		for _, c := range []struct {
			u   *itUser
			org tenancy.OrganizationID
		}{{w.foreign, w.A}, {w.owner, w.B}} {
			rctx, _ := beginAs(t, svc, ctx, c.u, c.org)
			_, err := svc.Resolve(rctx, ResolveRequest{OrganizationID: c.org.String()})
			wantCode(t, "cross-tenant context", err, CodeContextDenied)
		}
		// multi owns B and is a member of A: its B context sets only B's flag; its A context cannot.
		rctx, _ := beginAs(t, svc, ctx, w.multi, w.A)
		_, err := svc.SetOrganizationPolicy(rctx, resolve(t, svc, rctx, w.A, zeroG), tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
		wantCode(t, "member of A sets A's flag", err, CodeForbidden)
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
		// G016 A2: the B context is a separate request with multi's B credential (the A scope is confined to A).
		brctx, _ := beginAs(t, svc, ctx, w.multi, w.B)
		bctx := resolve(t, svc, brctx, w.B, zeroG)
		if _, err := svc.SetOrganizationPolicy(brctx, bctx, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: true}); err != nil {
			t.Errorf("owner of B sets B: %v", err)
		}
	})

	t.Run("IT-FLAG-4 false denies only Group-inherited updates", func(t *testing.T) {
		e.setFlag(t, w.A, false)
		defer e.setFlag(t, w.A, true)
		for _, u := range []*itUser{w.gOwner, w.gAdmin} {
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			tc := resolve(t, svc, rctx, w.A, zeroG)
			before := e.state(t, w.A)
			_, err := svc.UpdateBusinessProfile(rctx, tc, w.p1.ID, tenancy.BusinessContext{DisplayName: "Flag Off"})
			wantCode(t, "Group-inherited update with the flag off", err, CodeForbidden)
			if e.lastReason(t) != string(reasonEditPolicyDisabled) {
				t.Errorf("reason %s", e.lastReason(t))
			}
			if e.state(t, w.A) != before {
				t.Fatal("rows changed")
			}
			if _, err := svc.GetBusinessProfile(rctx, tc, w.p1.ID); err != nil {
				t.Errorf("read with the flag off: %v", err)
			}
			if got, err := svc.manageableProfiles(rctx, tc); err != nil || len(got) != 1 || got[0].ID != w.p1.ID {
				t.Errorf("selection with the flag off: %d %v", len(got), err)
			}
		}
		// G060 (r0 A4/D-5): the Organization path needs the Group-management override, not the flag: off, the owner and
		// the admin (no Group role) are refused with nothing changed; on, both update p1 with the flag still off.
		for _, u := range []*itUser{w.owner, w.admin} {
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			before := e.state(t, w.A)
			_, err := svc.UpdateBusinessProfile(rctx, resolve(t, svc, rctx, w.A, zeroG), w.p1.ID, w.p1.Context)
			wantCode(t, "Organization path with the flag and the override off", err, CodeForbidden)
			if e.state(t, w.A) != before {
				t.Fatal("rows changed")
			}
		}
		e.withOverride(t, w.owner, w.A)
		for _, u := range []*itUser{w.owner, w.admin} {
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			if _, err := svc.UpdateBusinessProfile(rctx, resolve(t, svc, rctx, w.A, zeroG), w.p1.ID, w.p1.Context); err != nil {
				t.Errorf("Organization path with the flag off and the override on: %v", err)
			}
		}
	})

	t.Run("IT-FLAG-5 toggling changes no binding or link", func(t *testing.T) {
		bindings, links := e.part(t, "role_bindings", w.A), e.part(t, "group_business_profiles", w.A)
		rctx, _ := beginAs(t, svc, ctx, w.owner, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		for _, v := range []bool{false, true} {
			if _, err := svc.SetOrganizationPolicy(rctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: v}); err != nil {
				t.Fatal(err)
			}
			if e.part(t, "role_bindings", w.A) != bindings || e.part(t, "group_business_profiles", w.A) != links {
				t.Fatal("a binding or link changed")
			}
		}
	})

	t.Run("IT-FACADE-1 every method's allow and deny path", func(t *testing.T) {
		removed := e.newUser(t)
		e.grantOrg(t, w.A, removed, tenancy.OrgAdmin)
		e.bind(t, removed, w.A)
		rctxR, _ := beginAs(t, svc, ctx, removed, w.A)
		tcR := resolve(t, svc, rctxR, w.A, zeroG)
		if err := e.model.RemoveOrganizationMembership(ctx, w.A, removed.id); err != nil {
			t.Fatal(err)
		}
		type side struct {
			u    *itUser
			want Code
		}
		cases := []struct {
			name        string
			allow, deny side
			call        func(context.Context, *TenantContext) error
		}{
			{"GetOrganization", side{w.member, ""}, side{removed, CodeContextDenied}, func(c context.Context, tc *TenantContext) error { _, err := svc.GetOrganization(c, tc); return err }},
			{"GetOrganizationPolicy", side{w.member, ""}, side{removed, CodeContextDenied}, func(c context.Context, tc *TenantContext) error {
				_, err := svc.GetOrganizationPolicy(c, tc)
				return err
			}},
			{"SetOrganizationPolicy", side{w.owner, ""}, side{w.admin, CodeForbidden}, func(c context.Context, tc *TenantContext) error {
				_, err := svc.SetOrganizationPolicy(c, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: true})
				return err
			}},
			{"GetGroup", side{w.gMember, ""}, side{w.member, CodeNotFound}, func(c context.Context, tc *TenantContext) error { _, err := svc.GetGroup(c, tc, w.g1.ID); return err }},
			{"RenameGroup", side{w.gAdmin, ""}, side{w.gMember, CodeForbidden}, func(c context.Context, tc *TenantContext) error {
				_, err := svc.RenameGroup(c, tc, w.g1.ID, w.g1.Name)
				return err
			}},
			// G060: the allow sides of the Group and profile writes are real Group roles (g1's owner; g1's admin through the
			// p1 link with the flag on); without the override an Organization owner/admin is refused (IT-FLAG-4).
			{"ChangeGroupSlug", side{w.gOwner, ""}, side{w.gMember, CodeForbidden}, func(c context.Context, tc *TenantContext) error {
				_, err := svc.ChangeGroupSlug(c, tc, w.g1.ID, w.g1.Slug)
				return err
			}},
			{"GetBusinessProfile", side{w.gMember, ""}, side{w.member, CodeNotFound}, func(c context.Context, tc *TenantContext) error {
				_, err := svc.GetBusinessProfile(c, tc, w.p1.ID)
				return err
			}},
			{"UpdateBusinessProfile", side{w.gAdmin, ""}, side{w.gMember, CodeForbidden}, func(c context.Context, tc *TenantContext) error {
				_, err := svc.UpdateBusinessProfile(c, tc, w.p1.ID, w.p1.Context)
				return err
			}},
		}
		for _, c := range cases {
			rctx, _ := beginAs(t, svc, ctx, c.allow.u, w.A)
			wantCode(t, c.name+" allow", c.call(rctx, resolve(t, svc, rctx, w.A, zeroG)), "")
			before := e.state(t, w.A, w.B)
			if c.deny.u == removed {
				wantCode(t, c.name+" deny", c.call(rctxR, tcR), c.deny.want)
			} else {
				dctx, _ := beginAs(t, svc, ctx, c.deny.u, w.A)
				wantCode(t, c.name+" deny", c.call(dctx, resolve(t, svc, dctx, w.A, zeroG)), c.deny.want)
			}
			if e.state(t, w.A, w.B) != before {
				t.Fatalf("%s: a denial changed rows", c.name)
			}
		}
	})

	t.Run("IT-FACADE-2 definitive aborts change nothing", func(t *testing.T) {
		// G060: the owner (no g2 role) reaches the write and its faults only through the Group-management override
		// (explicit setup before any fault is armed); without it the call is a denial whose access.denied record runs
		// into the armed COMMIT fault instead (the attempt-2 IT-LOG-1 audit.record_failed records).
		e.withOverride(t, w.owner, w.A)
		rctx, _ := beginAs(t, svc, ctx, w.owner, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		defer e.hook.reset()
		run := func(what string, setup func(cancel func()), reason Reason) {
			t.Helper()
			e.hook.reset()
			cctx, cancel := context.WithCancel(rctx)
			defer cancel()
			setup(cancel)
			before := e.state(t, w.A)
			_, err := svc.RenameGroup(cctx, tc, w.g2.ID, "Aborted "+what)
			wantCode(t, what, err, CodeUnavailable)
			if e.lastReason(t) != string(reason) {
				t.Errorf("%s: reason %s, want %s", what, e.lastReason(t), reason)
			}
			if e.state(t, w.A) != before {
				t.Fatalf("%s: rows changed", what)
			}
		}
		run("write-statement failure", func(func()) { e.hook.failOn = "W" }, reasonStoreUnavailable)
		run("server-rejected COMMIT", func(func()) { e.hook.commit = commitDefinitive }, reasonAborted)
		run("cancellation before COMMIT", func(cancel func()) { e.hook.on("after", "W", cancel) }, reasonCanceled)
		run("cancellation in the facts", func(cancel func()) { e.hook.on("after", "GB", cancel) }, reasonCanceled)
	})

	t.Run("IT-FACADE-3 a lost COMMIT reply", func(t *testing.T) {
		// G060: the unlinked p3 is updatable by the owner only through the Group-management override (set before the
		// fault is armed; see IT-FACADE-2).
		e.withOverride(t, w.owner, w.A)
		rctx, _ := beginAs(t, svc, ctx, w.owner, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		defer e.hook.reset()
		e.hook.reset()
		e.hook.commit = commitLostReply
		desc := "Lost reply " + e.canaries[0]
		_, err := svc.UpdateBusinessProfile(rctx, tc, w.p3.ID, tenancy.BusinessContext{DisplayName: w.p3.Context.DisplayName, Description: &desc})
		wantCode(t, "lost reply", err, CodeUnavailable)
		if e.lastReason(t) != string(reasonCommitOutcomeUnknown) {
			t.Errorf("reason %s", e.lastReason(t))
		}
		begins, stmts := e.hook.counts()
		writes := 0
		for _, s := range stmts {
			if s == "W" {
				writes++
			}
		}
		if begins != 1 || writes != 1 {
			t.Errorf("attempts: %d BeginTx, statements %v (exactly one attempt, no retry)", begins, stmts)
		}
		// The write is durable; a guarded read, subject to current authority, reconciles it.
		e.hook.reset()
		p, err := svc.GetBusinessProfile(rctx, tc, w.p3.ID)
		if err != nil || p.Context.Description == nil || *p.Context.Description != desc {
			t.Fatalf("reconcile: %v", err)
		}
	})

	t.Run("IT-FACADE-4 the operation to capability mapping", func(t *testing.T) {
		// G060: the owner's gNew and p3 writes are Organization-path writes (Group-management override, explicit setup).
		e.withOverride(t, w.owner, w.A)
		rctx, _ := beginAs(t, svc, ctx, w.owner, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		want := map[string][2]string{
			"GetOrganization": {opGetOrganization, "organization.read"}, "GetOrganizationPolicy": {opGetOrganizationPolicy, "organization.read"},
			"SetOrganizationPolicy": {opSetOrganizationPolicy, "organization.manage_policy"}, "GetGroup": {opGetGroup, "group.read"},
			"RenameGroup": {opRenameGroup, "group.update"}, "ChangeGroupSlug": {opChangeGroupSlug, "group.update"},
			"GetBusinessProfile":    {opGetBusinessProfile, "business_profile.profile.read"},
			"UpdateBusinessProfile": {opUpdateBusinessProfile, "business_profile.profile.update"},
		}
		for _, op := range itOps(svc, w.gNew.ID, w.p3.ID) {
			if op.name == "ChangeGroupSlug" {
				continue // its slug is covered below without renaming gNew's slug
			}
			if err := op.call(rctx, tc); err != nil {
				t.Fatalf("%s: %v", op.name, err)
			}
			rec := e.lastRecord(t)
			if rec["operation"] != want[op.name][0] || rec["capability"] != want[op.name][1] || rec["outcome"] != outcomeAllow {
				t.Errorf("%s record: %v", op.name, rec)
			}
		}
		if _, err := svc.ChangeGroupSlug(rctx, tc, w.gNew.ID, w.gNew.Slug); err != nil {
			t.Fatal(err)
		}
		if rec := e.lastRecord(t); rec["operation"] != opChangeGroupSlug || rec["capability"] != "group.update" {
			t.Errorf("ChangeGroupSlug record: %v", rec)
		}
	})

	t.Run("IT-FACADE-5 inaccessible equals unknown", func(t *testing.T) {
		unknownG, _ := tenancy.ParseGroupID("7c9e6679-7425-40de-944b-e07fc1f90ae7")
		unknownP, _ := tenancy.ParseBusinessProfileID("7c9e6679-7425-40de-944b-e07fc1f90ae7")
		rctx, _ := beginAs(t, svc, ctx, w.gAdmin, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		before := e.state(t, w.A, w.B)
		defer e.hook.reset()
		for what, calls := range map[string][]func() error{
			"GetGroup": {
				func() error { _, err := svc.GetGroup(rctx, tc, w.gB.ID); return err },
				func() error { _, err := svc.GetGroup(rctx, tc, unknownG); return err },
				func() error { _, err := svc.GetGroup(rctx, tc, w.g2.ID); return err },
			},
			"RenameGroup": {
				func() error { _, err := svc.RenameGroup(rctx, tc, w.gB.ID, "X"); return err },
				func() error { _, err := svc.RenameGroup(rctx, tc, unknownG, "X"); return err },
				func() error { _, err := svc.RenameGroup(rctx, tc, w.g2.ID, "X"); return err },
			},
			"UpdateBusinessProfile": {
				func() error { _, err := svc.UpdateBusinessProfile(rctx, tc, w.pB.ID, w.p1.Context); return err },
				func() error { _, err := svc.UpdateBusinessProfile(rctx, tc, unknownP, w.p1.Context); return err },
				func() error { _, err := svc.UpdateBusinessProfile(rctx, tc, w.p3.ID, w.p1.Context); return err },
			},
		} {
			var msgs []string
			for _, call := range calls {
				e.hook.reset()
				err := call()
				wantCode(t, what, err, CodeNotFound)
				if _, stmts := e.hook.counts(); slices.Contains(stmts, "W") {
					t.Errorf("%s: a write reached the database", what)
				}
				if err != nil {
					msgs = append(msgs, err.Error())
				}
			}
			if len(slices.Compact(msgs)) != 1 {
				t.Errorf("%s: foreign, unknown and inaccessible differ: %v", what, msgs)
			}
		}
		if e.state(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-FACADE-6 payload rules before authorization, conflicts and no-ops", func(t *testing.T) {
		rctx, _ := beginAs(t, svc, ctx, w.gMember, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		defer e.hook.reset()
		for what, c := range map[string]struct {
			call  func() error
			field tenancy.Code
		}{
			"invalid name":  {func() error { _, err := svc.RenameGroup(rctx, tc, w.g1.ID, ""); return err }, tenancy.CodeInvalidName},
			"invalid slug":  {func() error { _, err := svc.ChangeGroupSlug(rctx, tc, w.g1.ID, "Bad Slug"); return err }, tenancy.CodeInvalidSlug},
			"reserved slug": {func() error { _, err := svc.ChangeGroupSlug(rctx, tc, w.g1.ID, "settings"); return err }, tenancy.CodeSlugReserved},
			"invalid website": {func() error {
				site := "ftp://example.com/"
				_, err := svc.UpdateBusinessProfile(rctx, tc, w.p1.ID, tenancy.BusinessContext{DisplayName: "X", Website: &site})
				return err
			}, tenancy.CodeInvalidWebsite},
		} {
			e.hook.reset()
			err := c.call()
			var ae *Error
			begins, _ := e.hook.counts()
			if !errors.As(err, &ae) || ae.Code != CodeInvalidInput || ae.Field != string(c.field) || begins != 0 {
				t.Errorf("%s: %v (field %v) with %d BeginTx", what, err, ae, begins)
			}
		}
		// G060: the slug conflict and no-op are reached by g1's owner (a real Group role; the Organization owner without
		// the override is refused before the write).
		octx, _ := beginAs(t, svc, ctx, w.gOwner, w.A)
		otc := resolve(t, svc, octx, w.A, zeroG)
		before := e.state(t, w.A)
		_, err := svc.ChangeGroupSlug(octx, otc, w.g1.ID, w.g2.Slug)
		wantCode(t, "slug collision", err, CodeConflict)
		if e.state(t, w.A) != before {
			t.Fatal("a conflict changed rows")
		}
		if g, err := svc.ChangeGroupSlug(octx, otc, w.g1.ID, w.g1.Slug); err != nil || g.Slug != w.g1.Slug {
			t.Errorf("setting the current slug: %v", err)
		}
	})

	t.Run("IT-REVOKE-1 Organization removal and downgrade", func(t *testing.T) {
		// G060: the admin's update of the unlinked p3 is an Organization-path write (Group-management override); the
		// downgrade then removes the Organization path itself.
		e.withOverride(t, w.owner, w.A)
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgAdmin)
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, svc, ctx, u, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		if _, err := svc.UpdateBusinessProfile(rctx, tc, w.p3.ID, w.p3.Context); err != nil {
			t.Fatal(err)
		}
		profiles := e.part(t, "business_profiles", w.A)
		if _, err := e.model.ChangeOrganizationRole(ctx, w.A, u.id, tenancy.OrgMember); err != nil {
			t.Fatal(err)
		}
		_, err := svc.UpdateBusinessProfile(rctx, tc, w.p3.ID, tenancy.BusinessContext{DisplayName: "Downgraded"})
		wantCode(t, "update after downgrade", err, CodeNotFound)
		resolve(t, svc, rctx, w.A, zeroG) // still a member
		if err := e.model.RemoveOrganizationMembership(ctx, w.A, u.id); err != nil {
			t.Fatal(err)
		}
		_, err = svc.GetOrganization(rctx, tc)
		wantCode(t, "read after removal", err, CodeContextDenied)
		_, err = svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String()})
		wantCode(t, "fresh Resolve after removal", err, CodeContextDenied)
		o := e.newUser(t)
		e.grantOrg(t, w.A, o, tenancy.OrgOwner)
		e.bind(t, o, w.A)
		octx, _ := beginAs(t, svc, ctx, o, w.A)
		otc := resolve(t, svc, octx, w.A, zeroG)
		if _, err := e.model.ChangeOrganizationRole(ctx, w.A, o.id, tenancy.OrgAdmin); err != nil {
			t.Fatal(err)
		}
		_, err = svc.SetOrganizationPolicy(octx, otc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
		wantCode(t, "policy after owner downgrade", err, CodeForbidden)
		if e.part(t, "business_profiles", w.A) != profiles {
			t.Fatal("a profile changed")
		}
		// Selected-Group membership loss.
		s := e.newUser(t)
		e.grantOrg(t, w.A, s, tenancy.OrgMember)
		e.grantGroup(t, w.A, w.g3.ID, s, tenancy.GroupAdmin)
		e.bind(t, s, w.A)
		sctx, _ := beginAs(t, svc, ctx, s, w.A)
		stc := resolve(t, svc, sctx, w.A, w.g3.ID)
		if err := e.model.RemoveGroupMembership(ctx, w.A, w.g3.ID, s.id); err != nil {
			t.Fatal(err)
		}
		_, err = svc.GetOrganization(sctx, stc)
		wantCode(t, "operation after selected-Group loss", err, CodeContextDenied)
		_, err = svc.Resolve(sctx, ResolveRequest{OrganizationID: w.A.String(), GroupID: w.g3.ID.String()})
		wantCode(t, "fresh selected Resolve after loss", err, CodeContextDenied)
		resolve(t, svc, sctx, w.A, zeroG)
	})

	t.Run("IT-REVOKE-2 unlink", func(t *testing.T) {
		g := e.group(t, w.A, w.creator, "Unlink Group", "unlink-group")
		p := e.profile(t, w.A, "Unlink Profile "+e.canaries[2])
		e.link(t, w.A, g.ID, p.ID)
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
		e.grantGroup(t, w.A, g.ID, u, tenancy.GroupAdmin)
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, svc, ctx, u, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		if _, err := svc.UpdateBusinessProfile(rctx, tc, p.ID, p.Context); err != nil {
			t.Fatal(err)
		}
		if err := e.model.UnlinkGroupBusinessProfile(ctx, w.A, g.ID, p.ID); err != nil {
			t.Fatal(err)
		}
		_, err := svc.UpdateBusinessProfile(rctx, tc, p.ID, tenancy.BusinessContext{DisplayName: "After Unlink"})
		wantCode(t, "update after unlink", err, CodeNotFound)
		resolve(t, svc, rctx, w.A, g.ID) // a link does not affect context resolution
	})

	t.Run("IT-REVOKE-3 flag false", func(t *testing.T) {
		rctx, _ := beginAs(t, svc, ctx, w.gAdmin, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		if _, err := svc.UpdateBusinessProfile(rctx, tc, w.p1.ID, w.p1.Context); err != nil {
			t.Fatal(err)
		}
		e.setFlag(t, w.A, false)
		defer e.setFlag(t, w.A, true)
		_, err := svc.UpdateBusinessProfile(rctx, tc, w.p1.ID, tenancy.BusinessContext{DisplayName: "Flag Off"})
		wantCode(t, "update after the flag turned off", err, CodeForbidden)
	})

	t.Run("IT-REVOKE-4 disable, and enabling revives nothing", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgOwner)
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, svc, ctx, u, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		if err := e.id.DisableUser(ctx, u.id.String()); err != nil {
			t.Fatal(err)
		}
		_, err := svc.GetOrganization(rctx, tc)
		wantCode(t, "read after disable", err, CodeUnauthenticated)
		_, err = svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String()})
		wantCode(t, "Resolve after disable", err, CodeUnauthenticated)
		if err := e.id.EnableUser(ctx, u.id.String()); err != nil {
			t.Fatal(err)
		}
		_, err = svc.SetOrganizationPolicy(rctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
		wantCode(t, "write after enable (revoked session)", err, CodeUnauthenticated)
		_, end, err := svc.BeginRequest(ctx, u.in(t, w.A), itOriginOf(w.A))
		end()
		wantCode(t, "old credential after enable", err, CodeUnauthenticated)
	})

	t.Run("IT-REVOKE-5 logout", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgAdmin)
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, svc, ctx, u, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		e.logout(t, u.in(t, w.A), identity.ScopeCurrent)
		_, err := svc.UpdateBusinessProfile(rctx, tc, w.p3.ID, w.p3.Context)
		wantCode(t, "write after logout", err, CodeUnauthenticated)
		_, err = svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String()})
		wantCode(t, "Resolve after logout", err, CodeUnauthenticated)
	})

	t.Run("IT-REVOKE-6 logout-all", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgAdmin)
		other := e.relogin(t, u)
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, svc, ctx, u, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		e.logout(t, other, identity.ScopeAll)
		_, err := svc.GetGroup(rctx, tc, w.g1.ID)
		wantCode(t, "read after logout-all", err, CodeUnauthenticated)
		_, err = svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String()})
		wantCode(t, "Resolve after logout-all", err, CodeUnauthenticated)
	})

	t.Run("IT-REVOKE-7 expiry", func(t *testing.T) {
		clk := newITClock()
		clk.t = e.clk.Now()
		xsvc, _ := e.service(e.identity(t, clk, e.pool))
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgAdmin)
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, xsvc, ctx, u, w.A)
		tc := resolve(t, xsvc, rctx, w.A, zeroG)
		if _, err := xsvc.GetOrganization(rctx, tc); err != nil {
			t.Fatal(err)
		}
		clk.Advance(itIdle + time.Second)
		_, err := xsvc.UpdateBusinessProfile(rctx, tc, w.p3.ID, w.p3.Context)
		wantCode(t, "write after expiry", err, CodeUnauthenticated)
		_, err = xsvc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String()})
		wantCode(t, "Resolve after expiry", err, CodeUnauthenticated)
	})

	t.Run("IT-REVOKE-8 the last management path, member read and regrants", func(t *testing.T) {
		gx := e.group(t, w.A, w.creator, "Path X", "path-x")
		gy := e.group(t, w.A, w.creator, "Path Y", "path-y")
		p := e.profile(t, w.A, "Path Profile "+e.canaries[1])
		e.link(t, w.A, gx.ID, p.ID)
		e.link(t, w.A, gy.ID, p.ID)
		u := e.newUser(t)
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
		e.grantGroup(t, w.A, gx.ID, u, tenancy.GroupOwner)
		e.grantGroup(t, w.A, gy.ID, u, tenancy.GroupAdmin)
		e.bind(t, u, w.A)
		rctx, _ := beginAs(t, svc, ctx, u, w.A)
		tc := resolve(t, svc, rctx, w.A, zeroG)
		manageable := func() bool {
			ps, err := svc.manageableProfiles(rctx, tc)
			if err != nil {
				t.Fatal(err)
			}
			return slices.ContainsFunc(ps, func(x tenancy.BusinessProfile) bool { return x.ID == p.ID })
		}
		bindingCount := func() string { return e.part(t, "role_bindings", w.A) }
		if err := e.model.RemoveGroupMembership(ctx, w.A, gx.ID, u.id); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.UpdateBusinessProfile(rctx, tc, p.ID, p.Context); err != nil || !manageable() {
			t.Fatalf("the gy path still allows: %v", err)
		}
		profileRows := e.part(t, "business_profiles", w.A)
		bindings := bindingCount()
		if _, err := e.model.ChangeGroupRole(ctx, w.A, gy.ID, u.id, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		bindings2 := bindingCount()
		_, err := svc.UpdateBusinessProfile(rctx, tc, p.ID, tenancy.BusinessContext{DisplayName: "No Path"})
		wantCode(t, "update without a management path", err, CodeForbidden)
		if manageable() {
			t.Error("the profile stayed manageable")
		}
		if _, err := svc.GetBusinessProfile(rctx, tc, p.ID); err != nil {
			t.Errorf("member read retained: %v", err)
		}
		if e.part(t, "business_profiles", w.A) != profileRows || bindingCount() != bindings2 || bindings == bindings2 {
			t.Error("a profile row changed, or a binding was created or synchronized")
		}
		// G060: an Organization path allows the same profile only with the Group-management override (off: refused,
		// nothing changed).
		actx, _ := beginAs(t, svc, ctx, w.admin, w.A)
		_, err = svc.UpdateBusinessProfile(actx, resolve(t, svc, actx, w.A, zeroG), p.ID, p.Context)
		wantCode(t, "Organization path, override off", err, CodeForbidden)
		if e.part(t, "business_profiles", w.A) != profileRows {
			t.Error("a refused Organization-path update changed a profile row")
		}
		off := e.withOverride(t, w.owner, w.A)
		actx, _ = beginAs(t, svc, ctx, w.admin, w.A)
		if _, err := svc.UpdateBusinessProfile(actx, resolve(t, svc, actx, w.A, zeroG), p.ID, p.Context); err != nil {
			t.Errorf("Organization path, override on: %v", err)
		}
		off()
		// A Group regrant does not create Organization membership.
		if err := e.model.RemoveOrganizationMembership(ctx, w.A, u.id); err != nil {
			t.Fatal(err)
		}
		e.grantGroup(t, w.A, gx.ID, u, tenancy.GroupOwner)
		_, err = svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String()})
		wantCode(t, "Resolve after a Group regrant without membership", err, CodeContextDenied)
	})

	t.Run("IT-MFA-1 provider methods and MFA-like claims change nothing", func(t *testing.T) {
		var vectors []string
		for i, m := range []struct {
			method string
			claims map[string]any
		}{
			{"Password", nil}, {"GoogleOAuth", map[string]any{"amr": []string{"pwd", "mfa"}}},
			{"SSO", map[string]any{"acr": "aal2", "mfa": true}}, {"MagicAuth", map[string]any{"amr": []string{"otp"}, "assurance": "high"}},
		} {
			sub := fmt.Sprintf("user_01MFA%02d%06d", i, itSeq.Add(1))
			cred := e.signInAt(t, e.id, w.A, sub, "Mfa Fixture", m.claims, m.method, "")
			info, err := e.id.Session(ctx, cred, itOriginOf(w.A))
			if err != nil || info.MFAAssurance != identity.MFAAssuranceUnknown {
				t.Fatalf("%s assurance %q (%v)", m.method, info.MFAAssurance, err)
			}
			if err := identity.RequireVerifiedMFA(info); !errors.Is(err, identity.ErrMFAUnverified) {
				t.Errorf("%s: RequireVerifiedMFA %v", m.method, err)
			}
			u := &itUser{sub: sub, at: map[tenancy.OrganizationID]string{w.A: cred}}
			u.id, _ = tenancy.ParseUserID(info.UserID)
			e.grantOrg(t, w.A, u, tenancy.OrgMember)
			e.grantGroup(t, w.A, w.g1.ID, u, tenancy.GroupAdmin)
			rctx, _ := beginAs(t, svc, ctx, u, w.A)
			var v strings.Builder
			for _, g := range []tenancy.GroupID{zeroG, w.g1.ID} {
				tc := resolve(t, svc, rctx, w.A, g)
				for _, c := range oracleCapabilities {
					tg := map[TargetKind]Target{kindOrganization: {Kind: kindOrganization}, kindGroup: {Kind: kindGroup, Group: w.g1.ID},
						kindProfile: {Kind: kindProfile, Profile: w.p1.ID}}[c.kind]
					fmt.Fprintf(&v, "%t,", e.loadDecision(t, tc, Capability(c.name), tg))
				}
			}
			vectors = append(vectors, v.String())
		}
		if len(slices.Compact(vectors)) != 1 {
			t.Errorf("decisions differ across sign-in methods: %v", vectors)
		}
	})

	t.Run("IT-LOG-1 access.decision records only, without values", func(t *testing.T) {
		logs := e.logs.String()
		uuid := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
		var allow, deny, failed, ctxOps, sharedID int
		for i, line := range strings.Split(strings.TrimSpace(logs), "\n") {
			var rec map[string]any
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatalf("record %d is not JSON", i)
			}
			if rec["event"] != eventDecision {
				t.Errorf("record %d: event %v", i, rec["event"])
			}
			for _, c := range e.canaries {
				if strings.Contains(line, c) {
					t.Errorf("record %d carries a canary", i)
				}
			}
			if uuid.MatchString(line) {
				t.Errorf("record %d carries a UUID", i)
			}
			op, _ := rec["operation"].(string)
			_, hasCap := rec["capability"]
			switch op {
			case opBeginRequest, opResolve:
				ctxOps++
				if hasCap || rec["outcome"] == outcomeAllow || rec["reason"] == string(reasonAllowed) {
					t.Errorf("context record %d: %v", i, itH(rec))
				}
			default:
				if !hasCap || rec["outcome"] == outcomeEstablished {
					t.Errorf("guarded record %d: %v", i, itH(rec))
				}
			}
			switch rec["outcome"] {
			case outcomeAllow:
				allow++
			case outcomeDeny:
				deny++
			case outcomeFailed:
				failed++
			}
			if rec["request_id"] == "it-ctx7-shared-id" {
				sharedID++
			}
		}
		if allow == 0 || deny == 0 || failed == 0 || ctxOps == 0 || sharedID < 2 {
			t.Errorf("records: allow %d deny %d failed %d context %d shared-ID %d", allow, deny, failed, ctxOps, sharedID)
		}
	})
}
