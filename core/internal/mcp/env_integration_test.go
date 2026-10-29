//go:build integration

// G014 B-MCP-IT (Plan r8 §10.3): the shared environment of the MCP integration tests. Real HTTP (httptest over the
// production constructor mcp.New and Handler.Wrap), the real access facade over a pool logged in as the actual runtime
// role, the real G004 identity service (fake WorkOS and the real adapter) on a controllable clock, the real Connect
// validator with the identitytest Connect issuer as the test signer (the C10 static trust anchor), and the official SDK
// client. Fixtures use the attested superuser of the disposable database; assertions never do.
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, DB_SUPERUSER,
// DB_SUPERUSER_PASSWORD and APP_ENV; MCP_LOG_FILE (the core records are appended) and, for TestMCPRecovery,
// MCP_INVALIDATE_SQL_B64 (infra/recovery/sql/invalidate.sql as the harness read it at the bound base).
package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/identity/workos"
	"core/internal/migrate"
	"core/internal/observability"
	"core/internal/tenancy"
	"core/migrations"
)

const (
	itClientG004 = "client_01MCPINTEGRATION0000000000"
	itConnect    = "client_01MCPCONNECTCLIENT0000000"
	itConnect2   = "client_01MCPOTHERCLIENT000000000"
	itHostA      = "org-a.localhost:543"
	itHostB      = "org-b.localhost:543"
)

type itClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *itClock) Now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *itClock) Advance(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.b.Write(p) }
func (s *syncBuf) String() string              { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

type itUser struct {
	name, sub, id string
	uid           tenancy.UserID
	at            map[string]string // G016 A2: the browser credential prepared for each host, issued at "https://"+host
}

type mcpEnv struct {
	cfg     migrate.Config
	pw      string
	setup   *pgx.Conn
	setupDB *pgxpool.Pool
	pool    *pgxpool.Pool
	fake    *identitytest.Fake
	clk     *itClock
	id      *identity.Service
	h       *Handler
	srv     *httptest.Server
	model   *tenancy.Service
	logs    *syncBuf
	orgA    tenancy.OrganizationID
	orgB    tenancy.OrganizationID
	a1, a2  tenancy.Group
	b1      tenancy.Group
	users   map[string]*itUser
	secrets []string // observed token values (privacy scan)
	// G016 A2: hosts maps each fixture tenant host to the Organization its browser origin is issued for (see route);
	// origins maps each issued browser credential to the origin it is bound to (see csrf).
	mu      sync.Mutex
	hosts   map[string]tenancy.OrganizationID
	origins map[string]string
}

// loginHost is the browser-identity resolver of the MCP environment (G016 A2): the origin "https://"+host of each
// registered fixture host routes to the Organization registered for it, which is the Organization the serving
// handler's HostLookup maps that host to; any other origin routes nowhere. It reads nothing from the database.
type loginHost struct{ e *mcpEnv }

func (h loginHost) ResolveOrigin(_ context.Context, _ identity.Querier, origin string) (identity.HostResolution, error) {
	host, ok := strings.CutPrefix(origin, "https://")
	if !ok {
		return identity.HostResolution{}, nil
	}
	h.e.mu.Lock()
	org, ok := h.e.hosts[host]
	h.e.mu.Unlock()
	if !ok {
		return identity.HostResolution{}, nil
	}
	return identity.HostResolution{Kind: identity.HostPlatform, OrganizationID: org.String(), Routing: true}, nil
}

// route registers host's browser origin for org in the identity resolver until the registering test ends (G016 A2).
// Every registration mirrors the Organization the handler serving that host maps it to; a host already routed to
// another Organization is fatal (a fixture never reassigns a browser origin).
func (e *mcpEnv) route(t *testing.T, host string, org tenancy.OrganizationID) {
	t.Helper()
	e.mu.Lock()
	defer e.mu.Unlock()
	if prev, ok := e.hosts[host]; ok {
		if prev != org {
			t.Fatalf("route %s: already routed to another Organization", host)
		}
		return
	}
	e.hosts[host] = org
	t.Cleanup(func() {
		e.mu.Lock()
		delete(e.hosts, host)
		e.mu.Unlock()
	})
}

// hostOrg returns the Organization registered for host; an unregistered host is fatal.
func (e *mcpEnv) hostOrg(t *testing.T, host string) tenancy.OrganizationID {
	t.Helper()
	e.mu.Lock()
	org, ok := e.hosts[host]
	e.mu.Unlock()
	if !ok {
		t.Fatalf("host %s is not routed", host)
	}
	return org
}

// admit prepares one fixture sign-in at host under G016 D-2(a), where a tenant-host session is issued only to a current
// member of the host's Organization: the subject's user and principal exist and it holds a membership of that
// Organization. A membership is added only when the subject holds no Organization-scope binding of it; the returned
// function removes exactly the row it added (by id) right after the sign-in, so an existing owner/admin/member binding
// is never touched and a membership-negative subject is membership-negative again before any assertion (G016 A2).
func (e *mcpEnv) admit(t *testing.T, sub, host string) func() {
	t.Helper()
	ctx := context.Background()
	org := e.hostOrg(t, host)
	var uid string
	err := e.setupDB.QueryRow(ctx, `SELECT user_id::text FROM app.identity_principals WHERE issuer = $1 AND subject = $2`, e.fake.Issuer, sub).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		err = e.setupDB.QueryRow(ctx, `WITH u AS (INSERT INTO app.users (display_name) VALUES ('Fixture User') RETURNING id),
  p AS (INSERT INTO app.identity_principals (issuer, subject, user_id) SELECT $1, $2, id FROM u)
SELECT id::text FROM u`, e.fake.Issuer, sub).Scan(&uid)
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

var itSeq atomic.Int64

func itConn(t *testing.T, cfg migrate.Config, user, pw string) *pgx.Conn {
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

func itPool(t *testing.T, cfg migrate.Config, user, pw string) *pgxpool.Pool {
	t.Helper()
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
	return p
}

// setupMCP migrates to the reviewed set and builds the environment with the standard fixture: Organizations A (Groups A1,
// A2) and B (Group B1); agency staff ownerA (org_owner A) and adminB (org_admin B); client users memberA1 (org_member A,
// group_member A1) and cross (A1 and B1 member); one non-member; and registry hosts for A and B. G016 A2: every user
// holds a browser credential issued at A's origin; cross also a distinct one issued at B's origin.
func setupMCP(t *testing.T) *mcpEnv {
	t.Helper()
	ctx := context.Background()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil {
		t.Fatalf("integration settings: %v", err)
	}
	e := &mcpEnv{cfg: cfg, pw: os.Getenv("DB_RUNTIME_PASSWORD"), users: map[string]*itUser{}, logs: &syncBuf{},
		hosts: map[string]tenancy.OrganizationID{}, origins: map[string]string{}}
	super, superPW := os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD")
	if e.pw == "" || super == "" || superPW == "" {
		t.Fatal("integration settings: DB_RUNTIME_PASSWORD, DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required")
	}
	r, err := migrate.Open(ctx, cfg, migrations.FS)
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	var out bytes.Buffer
	if err := r.Up(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t22\n") {
		t.Fatalf("migrate up: %v %q", err, out.String())
	}
	_ = r.Close()
	e.setup = itConn(t, cfg, super, superPW)
	e.setupDB = itPool(t, cfg, super, superPW)
	var isSuper bool
	if err := e.setup.QueryRow(ctx, `SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&isSuper); err != nil || !isSuper {
		t.Fatalf("setup connection is not a superuser (%v)", err)
	}
	// Earlier tests of the suite bind another provider namespace; this environment's fake WorkOS is a new one.
	if _, err := e.setup.Exec(ctx, `DELETE FROM app.identity_provider_binding`); err != nil {
		t.Fatal(err)
	}
	e.pool = itPool(t, cfg, cfg.RuntimeRole(), e.pw)
	e.fake = identitytest.New(t, itClientG004)
	e.clk = &itClock{t: time.Now().UTC().Truncate(time.Second)}
	e.fake.SetNow(e.clk.Now)
	a, err := workos.New(workos.Config{APIKey: e.fake.APIKey, ClientID: itClientG004, Issuer: e.fake.Issuer, BaseURL: e.fake.URL(),
		Algorithms: []string{"RS256"}, Now: e.clk.Now})
	if err != nil {
		t.Fatal(err)
	}
	logger := observability.NewLogger(e.logs, slog.LevelInfo, nil)
	e.model = tenancy.NewService(e.setupDB, nil)
	e.id = identity.NewService(identity.NewStore(e.pool), identity.Config{Provider: a, Issuer: e.fake.Issuer, ClientID: itClientG004,
		Hosts: loginHost{e: e}, Idle: 30 * time.Minute, Max: 8 * time.Hour,
		Now: e.clk.Now, Logger: logger})

	e.orgA, e.orgB = e.org(t, "Organization A"), e.org(t, "Organization B")
	// G016 A2: A's and B's browser origins are issued for the Organizations the handler maps their hosts to.
	e.route(t, itHostA, e.orgA)
	e.route(t, itHostB, e.orgB)
	for _, n := range []string{"ownerA", "adminB", "memberA1", "cross", "nonmember", "dave"} {
		e.newUser(t, n)
	}
	e.a1 = e.group(t, e.orgA, "ownerA", "Group A1", fmt.Sprintf("a1-%d", itSeq.Add(1)))
	e.a2 = e.group(t, e.orgA, "ownerA", "Group A2", fmt.Sprintf("a2-%d", itSeq.Add(1)))
	e.b1 = e.group(t, e.orgB, "adminB", "Group B1", fmt.Sprintf("b1-%d", itSeq.Add(1)))
	e.grantOrg(t, e.orgA, "ownerA", tenancy.OrgOwner)
	e.grantOrg(t, e.orgB, "adminB", tenancy.OrgAdmin)
	e.grantOrg(t, e.orgA, "memberA1", tenancy.OrgMember)
	e.grantGroup(t, e.orgA, e.a1.ID, "memberA1", tenancy.GroupMember)
	e.grantOrg(t, e.orgA, "cross", tenancy.OrgMember)
	e.grantGroup(t, e.orgA, e.a1.ID, "cross", tenancy.GroupMember)
	e.grantOrg(t, e.orgB, "cross", tenancy.OrgMember)
	e.grantGroup(t, e.orgB, e.b1.ID, "cross", tenancy.GroupMember)
	e.grantOrg(t, e.orgA, "dave", tenancy.OrgMember)
	// G016 A2: the multi-tenant client also holds a distinct credential issued at B's origin (IT-MCP-4/6/13, listing).
	e.bind(t, "cross", itHostB)

	e.h = e.handler(t, map[string]tenancy.OrganizationID{itHostA: e.orgA, itHostB: e.orgB}, nil)
	mux := http.NewServeMux()
	mux.HandleFunc(RouteConnectionsList, e.h.List)
	mux.HandleFunc(RouteConnectionsCreate, e.h.Create)
	mux.HandleFunc(RouteConnectionsRevoke, e.h.Revoke)
	e.srv = httptest.NewServer(observability.Middleware(logger, e.h.Wrap(mux)))
	t.Cleanup(e.srv.Close)
	if f := os.Getenv("MCP_LOG_FILE"); f != "" {
		t.Cleanup(func() {
			fh, err := os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err == nil {
				_, err = fh.WriteString(e.logs.String())
				_ = fh.Close()
			}
			if err != nil {
				t.Errorf("write the core log: %v", err)
			}
		})
	}
	return e
}

// handler builds a handler over the environment's pool and identity with the given registry; v nil uses the static
// test trust anchor of the fake Connect issuer.
func (e *mcpEnv) handler(t *testing.T, hosts map[string]tenancy.OrganizationID, v Verifier) *Handler {
	t.Helper()
	if v == nil {
		cv, err := workos.NewConnectValidator(workos.ConnectConfig{Issuer: e.fake.ConnectIssuer(), Algorithms: []string{"RS256"},
			JWKS: e.fake.ConnectJWKS(), Now: e.clk.Now})
		if err != nil {
			t.Fatal(err)
		}
		v = cv
	}
	reg := staticLookup{hosts: map[string]string{}}
	for h, o := range hosts {
		reg.hosts[h] = o.String()
	}
	return e.lookupHandler(t, reg, v)
}

// lookupHandler builds a handler like handler over an explicit HostLookup (G016: the real durable resolver over the
// environment's pool, or a static alias map).
func (e *mcpEnv) lookupHandler(t *testing.T, lookup HostLookup, v Verifier) *Handler {
	t.Helper()
	if v == nil {
		cv, err := workos.NewConnectValidator(workos.ConnectConfig{Issuer: e.fake.ConnectIssuer(), Algorithms: []string{"RS256"},
			JWKS: e.fake.ConnectJWKS(), Now: e.clk.Now})
		if err != nil {
			t.Fatal(err)
		}
		v = cv
	}
	h, err := New(Config{Issuer: e.fake.ConnectIssuer(), Hosts: lookup, Verifier: v, Identity: e.id, Pool: e.pool,
		Lifetimes: identity.MCPLifetimes{Pending: 30 * time.Minute, Idle: 30 * 24 * time.Hour, Absolute: 180 * 24 * time.Hour},
		Title:     "Integration Product", Version: "0.0.0-it", Logger: observability.NewLogger(e.logs, slog.LevelInfo, nil)})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// serveWith routes the environment's HTTP server through h (setup routes and the data path) until the calling test
// ends, then restores the environment's own handler and server.
func (e *mcpEnv) serveWith(t *testing.T, h *Handler) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc(RouteConnectionsList, h.List)
	mux.HandleFunc(RouteConnectionsCreate, h.Create)
	mux.HandleFunc(RouteConnectionsRevoke, h.Revoke)
	srv := httptest.NewServer(observability.Middleware(observability.NewLogger(e.logs, slog.LevelInfo, nil), h.Wrap(mux)))
	prevSrv, prevH := e.srv, e.h
	e.srv, e.h = srv, h
	t.Cleanup(func() {
		srv.Close()
		e.srv, e.h = prevSrv, prevH
	})
}

func (e *mcpEnv) org(t *testing.T, name string) tenancy.OrganizationID {
	t.Helper()
	o, err := e.model.CreateOrganization(context.Background(), name)
	if err != nil {
		t.Fatal(err)
	}
	return o.ID
}

func (e *mcpEnv) group(t *testing.T, org tenancy.OrganizationID, creator, name, slug string) tenancy.Group {
	t.Helper()
	g, _, err := e.model.CreateGroup(context.Background(), org, e.users[creator].uid, name, slug)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func (e *mcpEnv) grantOrg(t *testing.T, org tenancy.OrganizationID, who string, r tenancy.OrganizationRole) {
	t.Helper()
	if _, err := e.model.GrantOrganizationRole(context.Background(), org, e.users[who].uid, r); err != nil {
		t.Fatal(err)
	}
}

func (e *mcpEnv) grantGroup(t *testing.T, org tenancy.OrganizationID, g tenancy.GroupID, who string, r tenancy.GroupRole) {
	t.Helper()
	if _, err := e.model.GrantGroupRole(context.Background(), org, g, e.users[who].uid, r); err != nil {
		t.Fatal(err)
	}
}

// newUser completes one real G004 sign-in of a new subject at A's origin and records that browser credential (G016
// A2: bound to A's origin and Organization; credentials for other hosts come from bind).
func (e *mcpEnv) newUser(t *testing.T, name string) *itUser {
	t.Helper()
	ctx := context.Background()
	n := itSeq.Add(1)
	u := &itUser{name: name, sub: fmt.Sprintf("user_01MCPIT%010d", n), at: map[string]string{}}
	cred := e.signIn(t, u.sub, itHostA)
	info, err := e.id.Session(ctx, cred, "https://"+itHostA)
	if err != nil {
		t.Fatal(err)
	}
	u.at[itHostA] = cred
	u.id = info.UserID
	if u.uid, err = tenancy.ParseUserID(info.UserID); err != nil {
		t.Fatal(err)
	}
	e.users[name] = u
	return u
}

// signIn completes one real G004 sign-in of sub at "https://"+host (G016 A2: the session is bound to that origin and to
// the Organization registered for host) and returns the credential; the admission is released right after.
func (e *mcpEnv) signIn(t *testing.T, sub, host string) string {
	t.Helper()
	ctx := context.Background()
	origin := "https://" + host
	defer e.admit(t, sub, host)()
	out, err := e.id.Login(ctx, identity.LoginInput{Origin: origin, ReturnTo: "/account"})
	if err != nil {
		t.Fatalf("login start: %v", err)
	}
	u, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	code := e.fake.IssueCode(u.Query().Get("code_challenge"), identitytest.Grant{UserID: sub, FirstName: "Fixture", LastName: "User",
		SID: fmt.Sprintf("session_01MCPIT%08d", itSeq.Add(1))})
	cb, err := e.id.Callback(ctx, identity.CallbackInput{Origin: origin, RawQuery: "code=" + url.QueryEscape(code) + "&state=" +
		url.QueryEscape(u.Query().Get("state")), Binding: out.Binding})
	if err != nil || cb.Credential == "" {
		t.Fatalf("callback: %v", err)
	}
	e.mu.Lock()
	e.origins[cb.Credential] = origin
	e.mu.Unlock()
	return cb.Credential
}

// bind prepares who's browser credential for each host through the real sign-in flow at "https://"+host, replacing any
// earlier credential for that host (G016 A2). Tests call it explicitly: at setup after the grants, after a logout,
// disable or invalidation the test itself performs, and as the explicit renewal right after the test clock moved past
// the browser session lifetimes (30m idle, 8h maximum). It is never chosen after a refusal.
func (e *mcpEnv) bind(t *testing.T, who string, hosts ...string) {
	t.Helper()
	u := e.users[who]
	for _, h := range hosts {
		u.at[h] = e.signIn(t, u.sub, h)
	}
}

// cred returns who's browser credential prepared for host (G016 A2); none prepared is fatal. It never renews.
func (e *mcpEnv) cred(t *testing.T, who, host string) string {
	t.Helper()
	c, ok := e.users[who].at[host]
	if !ok {
		t.Fatalf("no browser credential of %s prepared for %s", who, host)
	}
	return c
}

// csrf returns the CSRF token of cred's session, read at the credential's own origin (G016 A2).
func (e *mcpEnv) csrf(t *testing.T, cred string) string {
	t.Helper()
	e.mu.Lock()
	origin, ok := e.origins[cred]
	e.mu.Unlock()
	if !ok {
		t.Fatal("csrf: not a fixture-issued credential")
	}
	info, err := e.id.Session(context.Background(), cred, origin)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(info.CSRFToken)
}

// internal calls the internal setup API with the browser credential.
func (e *mcpEnv) internal(t *testing.T, method, path, cred string, body any) (int, map[string]any) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.srv.URL+path, rd)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cred != "" {
		req.Header.Set("Authorization", "Session "+cred)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

type conn struct {
	id, host, url string
	owner         *itUser
}

// create creates a pending connection of who on host through the internal API with who's credential prepared for host
// (G016 A2: it never signs in again; renewal is the test's explicit bind).
func (e *mcpEnv) create(t *testing.T, who, host string) conn {
	t.Helper()
	u := e.users[who]
	cred := e.cred(t, who, host)
	code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections", cred,
		map[string]string{"origin": "https://" + host, "csrf_token": e.csrf(t, cred)})
	if code != http.StatusCreated {
		t.Fatalf("create for %s on %s: %d %v", who, host, code, out)
	}
	c := out["connection"].(map[string]any)
	return conn{id: c["connection_id"].(string), host: host, url: c["url"].(string), owner: u}
}

// token mints a Connect access token of who for audience aud (default client).
func (e *mcpEnv) token(who, aud string, over map[string]any) string {
	client := itConnect
	if c, ok := over["client_id"].(string); ok {
		client = c
	}
	tok := e.fake.SignConnect(e.users[who].sub, client, aud, over)
	e.secrets = append(e.secrets, tok)
	return tok
}

func (e *mcpEnv) tok(c conn) string { return e.token(c.owner.name, c.url, nil) }

// rawPost sends one MCP POST for connection path id on host with token ("" for none) and returns the status, headers and
// decoded JSON body.
func (e *mcpEnv) rawPost(t *testing.T, host, id, token, body string) (int, http.Header, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, e.srv.URL+ConnectionPathPrefix+id, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Host = host
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2025-11-25")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, resp.Header, out
}

var rpcSeq atomic.Int64

// rpc calls one JSON-RPC method and returns the HTTP status and the decoded response.
func (e *mcpEnv) rpc(t *testing.T, c conn, token, method string, params any) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": rpcSeq.Add(1), "method": method, "params": params})
	code, _, out := e.rawPost(t, c.host, c.id, token, string(b))
	return code, out
}

// call calls a tool and returns the HTTP status, the structured content and whether it is a tool error.
func (e *mcpEnv) call(t *testing.T, c conn, token, tool string, args any) (int, map[string]any, bool) {
	t.Helper()
	if args == nil {
		args = map[string]any{}
	}
	code, out := e.rpc(t, c, token, "tools/call", map[string]any{"name": tool, "arguments": args})
	if code != http.StatusOK {
		return code, nil, false
	}
	res, _ := out["result"].(map[string]any)
	if res == nil {
		return code, map[string]any{"rpc_error": out["error"]}, true
	}
	sc, _ := res["structuredContent"].(map[string]any)
	isErr, _ := res["isError"].(bool)
	return code, sc, isErr
}

// mustCall calls a tool and fails unless it succeeds.
func (e *mcpEnv) mustCall(t *testing.T, c conn, token, tool string, args any) map[string]any {
	t.Helper()
	code, sc, isErr := e.call(t, c, token, tool, args)
	if code != http.StatusOK || isErr {
		t.Fatalf("%s on %s: %d %v", tool, c.id, code, sc)
	}
	return sc
}

// wantToolError calls a tool and fails unless it is the given closed tool error.
func (e *mcpEnv) wantToolError(t *testing.T, c conn, token, tool string, args any, code string) {
	t.Helper()
	status, sc, isErr := e.call(t, c, token, tool, args)
	if status != http.StatusOK || !isErr || sc["error"] != code {
		t.Fatalf("%s on %s: %d %v, want tool error %s", tool, c.id, status, sc, code)
	}
}

// want401 fails unless a protected call with token is refused with 401 and runs nothing.
func (e *mcpEnv) want401(t *testing.T, c conn, token, what string) {
	t.Helper()
	if code, out := e.rpc(t, c, token, "tools/call", map[string]any{"name": ToolGetCurrentContext, "arguments": map[string]any{}}); code != http.StatusUnauthorized {
		t.Fatalf("%s: %d %v, want 401", what, code, out)
	}
}

// hostRT injects the tenant Host and the bearer into the SDK client's requests.
type hostRT struct {
	host, token string
}

func (h hostRT) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Host = h.host
	r.Header.Set("Authorization", "Bearer "+h.token)
	return http.DefaultTransport.RoundTrip(r)
}

// dial connects the official SDK client to connection c with token.
func (e *mcpEnv) dial(t *testing.T, c conn, token string) (*sdk.ClientSession, error) {
	t.Helper()
	return e.dialWith(t, c, token, nil)
}

// dialWith is dial with client options (G015 CAT-5: the list-changed handler).
func (e *mcpEnv) dialWith(t *testing.T, c conn, token string, opts *sdk.ClientOptions) (*sdk.ClientSession, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tr := &sdk.StreamableClientTransport{Endpoint: e.srv.URL + ConnectionPathPrefix + c.id,
		HTTPClient: &http.Client{Transport: hostRT{host: c.host, token: token}}, MaxRetries: -1, DisableStandaloneSSE: true}
	return sdk.NewClient(&sdk.Implementation{Name: "g014-it", Version: "0.0.0"}, opts).Connect(ctx, tr, nil)
}

// G015 fixture helpers (Plan r4 §6.1, R-g: existing model operations and owner SQL only).

// profile creates a BusinessProfile of org.
func (e *mcpEnv) profile(t *testing.T, org tenancy.OrganizationID, name string) tenancy.BusinessProfile {
	t.Helper()
	p, err := e.model.CreateBusinessProfile(context.Background(), org, tenancy.BusinessContext{DisplayName: name})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// link associates profile p with Group g; unlink removes that association.
func (e *mcpEnv) link(t *testing.T, org tenancy.OrganizationID, g tenancy.GroupID, p tenancy.BusinessProfileID) {
	t.Helper()
	if _, err := e.model.LinkGroupBusinessProfile(context.Background(), org, g, p); err != nil {
		t.Fatal(err)
	}
}

func (e *mcpEnv) unlink(t *testing.T, org tenancy.OrganizationID, g tenancy.GroupID, p tenancy.BusinessProfileID) {
	t.Helper()
	if err := e.model.UnlinkGroupBusinessProfile(context.Background(), org, g, p); err != nil {
		t.Fatal(err)
	}
}

// setFlag assigns an Organization's group_admins_edit_business_profiles flag directly (owner SQL), for fixtures only.
func (e *mcpEnv) setFlag(t *testing.T, org tenancy.OrganizationID, on bool) {
	t.Helper()
	if _, err := e.setup.Exec(context.Background(), `UPDATE app.organizations SET group_admins_edit_business_profiles = $2 WHERE id = $1::text::uuid`,
		org.String(), on); err != nil {
		t.Fatal(err)
	}
}

// list returns the tool names of one raw tools/list call, in the order served, failing on a refusal or a cursor.
func (e *mcpEnv) list(t *testing.T, c conn, token string) []string {
	t.Helper()
	code, out := e.rpc(t, c, token, "tools/list", map[string]any{})
	res, ok := out["result"].(map[string]any)
	if code != http.StatusOK || !ok {
		t.Fatalf("tools/list on %s: %d %v", c.id, code, out)
	}
	if _, cursor := res["nextCursor"]; cursor {
		t.Fatalf("tools/list returned a cursor: %v", res)
	}
	var names []string
	for _, x := range res["tools"].([]any) {
		names = append(names, x.(map[string]any)["name"].(string))
	}
	return names
}

// listError returns the JSON-RPC error message of a refused tools/list ("" when it succeeded).
func (e *mcpEnv) listError(t *testing.T, c conn, token string) (int, string) {
	t.Helper()
	code, out := e.rpc(t, c, token, "tools/list", map[string]any{})
	if er, ok := out["error"].(map[string]any); ok {
		return code, fmt.Sprintf("%v %v", er["code"], er["message"])
	}
	return code, ""
}

// rawCall sends one tools/call for name with a fixed JSON-RPC id and returns the status and the exact response body.
func (e *mcpEnv) rawCall(t *testing.T, c conn, token, name string) (int, string) {
	t.Helper()
	b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 4242, "method": "tools/call", "params": map[string]any{"name": name, "arguments": map[string]any{}}})
	req, err := http.NewRequest(http.MethodPost, e.srv.URL+ConnectionPathPrefix+c.id, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Host = c.host
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2025-11-25")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(body)
}

// mcpDenials counts the access.denied mcp.request audit rows of user u (superuser read).
func (e *mcpEnv) mcpDenials(t *testing.T, u *itUser) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) FROM audit.events
  WHERE action = 'access.denied' AND operation = 'mcp.request' AND actor_user_id = $1::text::uuid`, u.id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// authorizeRecords returns the access.decision records of operation mcp.authorize written so far, each reduced to its
// level, outcome and reason (the correlation and time vary per request).
func (e *mcpEnv) authorizeRecords(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, line := range strings.Split(e.logs.String(), "\n") {
		var r map[string]any
		if json.Unmarshal([]byte(line), &r) != nil || r["event"] != "access.decision" || r["operation"] != "mcp.authorize" {
			continue
		}
		out = append(out, fmt.Sprintf("%v|%v|%v", r["level"], r["outcome"], r["reason"]))
	}
	return out
}

// state returns the stored lifecycle of connection id (superuser read).
func (e *mcpEnv) state(t *testing.T, id string) (state, client, group string, version int64) {
	t.Helper()
	var c, g *string
	if err := e.setup.QueryRow(context.Background(), `SELECT state, client_id, selected_group_id::text, context_version
FROM app.mcp_connections WHERE id = $1::text::uuid`, id).Scan(&state, &c, &g, &version); err != nil {
		t.Fatalf("state of %s: %v", id, err)
	}
	if c != nil {
		client = *c
	}
	if g != nil {
		group = *g
	}
	return state, client, group, version
}

// events lists the audit actions recorded for resource id (superuser read), oldest first.
func (e *mcpEnv) events(t *testing.T, resource string) []string {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT action || ':' || coalesce(group_id::text, '-') || ':' || actor_kind
FROM audit.events WHERE resource_id = $1::text::uuid ORDER BY occurred_at, id`, resource)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

// activate activates c with the default client through one lifecycle call (ping), and checks it.
func (e *mcpEnv) activate(t *testing.T, c conn) string {
	t.Helper()
	tok := e.tok(c)
	if code, out := e.rpc(t, c, tok, "ping", map[string]any{}); code != http.StatusOK {
		t.Fatalf("activate %s: %d %v", c.id, code, out)
	}
	if st, client, _, _ := e.state(t, c.id); st != identity.MCPStateActive || client != itConnect {
		t.Fatalf("after activation: %s %s", st, client)
	}
	return tok
}
