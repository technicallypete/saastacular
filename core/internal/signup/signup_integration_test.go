//go:build integration

// Integration tests of root signup (G062 Plan r5 §5.2 SUP-5..SUP-8; B-SIGNUP-IT) against a disposable PostgreSQL
// initialized by the rendered role script and migrated here to the reviewed set, with the fake WorkOS of package
// identitytest behind the real WorkOS adapter. They fail, never skip, when the database settings are absent. Fixture
// success proves application enforcement only, never provider behaviour. Run with -p 1 (the fault triggers are
// installed and dropped on the shared catalog).
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, APP_ENV,
// DB_SUPERUSER and DB_SUPERUSER_PASSWORD (setup-side counts and fault triggers only, never an assertion path).
package signup

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/hostname"
	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/identity/workos"
	"core/internal/migrate"
	"core/migrations"
)

const (
	itClient   = "client_01SIGNUPINTEGRATION0000000"
	rootOrigin = "https://localhost:543"
	slugPrefix = "sut-" // every slug these tests create; released after the package's tests
)

var it struct {
	cfg       migrate.Config
	runtimePW string
	pool      *pgxpool.Pool
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil || os.Getenv("DB_RUNTIME_PASSWORD") == "" || os.Getenv("DB_SUPERUSER") == "" {
		fmt.Fprintln(os.Stderr, "integration settings: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, APP_ENV, DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required")
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
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		panic(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = cfg.Host, uint16(cfg.Port), cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = cfg.RuntimeRole(), it.runtimePW
	pc.ConnConfig.ConnectTimeout = 3 * time.Second
	pc.MaxConns = 20
	if it.pool, err = pgxpool.NewWithConfig(ctx, pc); err != nil {
		panic(err)
	}
	code := m.Run()
	it.pool.Close()
	if err := releaseSlugs(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "release slugs:", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

func superDSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", it.cfg.Host, it.cfg.Port, it.cfg.Database,
		os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD"))
}

// releaseSlugs clears the platform slugs these tests created, so later fixtures on the same database never collide.
func releaseSlugs(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, superDSN())
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, `UPDATE app.organizations SET slug = NULL WHERE slug LIKE 'sut-%'`)
	return err
}

// setupConn is the attested superuser connection (setup-side counts and fault triggers only).
func setupConn(t *testing.T) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(context.Background(), superDSN())
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

// runtimeConn is a plain runtime connection (privilege denials).
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

// itHosts is the overlay classification (<label>.localhost:543 platform hosts over the real app.hostname_state reader,
// keyed by the platform label since 00021; the root and auth hosts are not tenant hosts; other names on :543 are custom).
type itHosts struct{}

func (itHosts) ResolveOrigin(ctx context.Context, q identity.Querier, origin string) (identity.HostResolution, error) {
	host, ok := strings.CutPrefix(origin, "https://")
	name, port, found := strings.Cut(host, ":")
	if !ok || !found || port != "543" || name == "" || strings.ContainsAny(name, "/?#@:") {
		return identity.HostResolution{}, nil
	}
	kind, key := identity.HostCustom, name
	if label, ok := strings.CutSuffix(name, ".localhost"); ok {
		if label == "" || strings.Contains(label, ".") || label == "auth" || label == "www" {
			return identity.HostResolution{}, nil
		}
		kind, key = identity.HostPlatform, label
	} else if name == "localhost" {
		return identity.HostResolution{}, nil
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

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) Now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) Advance(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

type harness struct {
	t    *testing.T
	fake *identitytest.Fake
	clk  *clock
	prov identity.Provider
	id   *identity.Service
	svc  *Service
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	f := identitytest.New(t, itClient)
	clk := &clock{t: time.Now().UTC().Truncate(time.Second)}
	f.SetNow(clk.Now)
	a, err := workos.New(workos.Config{APIKey: f.APIKey, ClientID: itClient, Issuer: f.Issuer, BaseURL: f.URL(), Algorithms: []string{"RS256"}, Now: clk.Now})
	if err != nil {
		t.Fatal(err)
	}
	// Each harness has its own fake issuer; the provider namespace is bound once per database, so reset it first.
	if _, err := setupConn(t).Exec(context.Background(), `DELETE FROM app.identity_provider_binding`); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, fake: f, clk: clk, prov: a}
	h.wire(it.pool)
	return h
}

// on returns a harness sharing h's fake provider, issuer binding and clock whose identity store uses pool.
func (h *harness) on(pool *pgxpool.Pool) *harness {
	n := &harness{t: h.t, fake: h.fake, clk: h.clk, prov: h.prov}
	n.wire(pool)
	return n
}

func (h *harness) wire(pool *pgxpool.Pool) {
	h.id = identity.NewService(identity.NewStore(pool), identity.Config{Provider: h.prov, Issuer: h.fake.Issuer, ClientID: itClient,
		Origins: identity.NewOrigins([]string{"https://auth.localhost:543"}), Hosts: itHosts{}, RootOrigin: rootOrigin,
		Idle: 30 * time.Minute, Max: 8 * time.Hour, Now: h.clk.Now, Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))})
	h.svc = NewService(Config{Identity: h.id, RootOrigin: rootOrigin, Now: h.clk.Now,
		Hostname: hostname.Config{BaseDomain: "localhost", Port: "543", AuthHost: "auth.localhost"}})
}

// genericPlanPool is a one-connection runtime pool whose backend starts with plan_cache_mode = force_generic_plan (a
// startup parameter, so it is also the RESET value): every statement of a flow run on it, the provisioning transaction
// included, uses that single backend with generic plans.
func genericPlanPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = it.cfg.Host, uint16(it.cfg.Port), it.cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = it.cfg.RuntimeRole(), it.runtimePW
	pc.ConnConfig.ConnectTimeout = 3 * time.Second
	pc.ConnConfig.RuntimeParams["plan_cache_mode"] = "force_generic_plan"
	pc.MinConns, pc.MaxConns = 0, 1
	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func grant(subject, sid string) identitytest.Grant {
	return identitytest.Grant{UserID: subject, SID: sid, FirstName: "Synthetic", LastName: "User"}
}

func callbackQuery(code, state string) string {
	return "code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state)
}

// login runs one login start and returns the output with the provider-visible state and challenge.
func (h *harness) startAt(t *testing.T, in identity.LoginInput) (identity.LoginOutput, string, string) {
	t.Helper()
	out, err := h.id.Login(context.Background(), in)
	if err != nil {
		t.Fatalf("login start %s: %v", in.Origin, err)
	}
	u, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	return out, u.Query().Get("state"), u.Query().Get("code_challenge")
}

// credential runs the root signup-intent login and T1 for subject and returns the signup credential and its resolution.
func (h *harness) credential(t *testing.T, subject, sid string) (string, identity.SignupCredentialInfo) {
	t.Helper()
	out, state, challenge := h.startAt(t, identity.LoginInput{Origin: rootOrigin, SignupIntent: true})
	code := h.fake.IssueCode(challenge, grant(subject, sid))
	cb, err := h.id.Callback(context.Background(), identity.CallbackInput{Origin: rootOrigin, RawQuery: callbackQuery(code, state), Binding: out.Binding})
	if err != nil || cb.Signup == nil || cb.Outcome != identity.OutcomeSignup || cb.ReturnTo != "/signup/details" || cb.Credential != "" {
		t.Fatalf("signup callback: %+v %v", cb, err)
	}
	info, err := h.svc.Resolve(context.Background(), cb.Signup.Credential, rootOrigin)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	return cb.Signup.Credential, info
}

// resolve re-reads a credential (the browser presents it again).
func (h *harness) resolve(t *testing.T, cred string) identity.SignupCredentialInfo {
	t.Helper()
	info, err := h.svc.Resolve(context.Background(), cred, rootOrigin)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	return info
}

func uniqueSlug(t *testing.T) string {
	t.Helper()
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return slugPrefix + hex.EncodeToString(b)
}

func uniqueSubject(t *testing.T) string {
	t.Helper()
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return "user_01SUP" + hex.EncodeToString(b)
}

// counts returns, for slug, the number of Organizations, org_owner bindings and organization.created/membership.granted
// audit rows (superuser setup reads).
func counts(t *testing.T, slug string) (orgs, owners, created, granted int) {
	t.Helper()
	err := setupConn(t).QueryRow(context.Background(), `
SELECT (SELECT count(*) FROM app.organizations WHERE slug = $1),
       (SELECT count(*) FROM app.role_bindings b JOIN app.organizations o ON o.id = b.organization_id
          WHERE o.slug = $1 AND b.role = 'org_owner'),
       (SELECT count(*) FROM audit.events e JOIN app.organizations o ON o.id = e.organization_id
          WHERE o.slug = $1 AND e.action = 'organization.created'),
       (SELECT count(*) FROM audit.events e JOIN app.organizations o ON o.id = e.organization_id
          WHERE o.slug = $1 AND e.action = 'membership.granted')`, slug).Scan(&orgs, &owners, &created, &granted)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func assertOnePair(t *testing.T, slug string) {
	t.Helper()
	if o, w, c, g := counts(t, slug); o != 1 || w != 1 || c != 1 || g != 1 {
		t.Errorf("%s: orgs %d, owners %d, organization.created %d, membership.granted %d; want exactly one each", slug, o, w, c, g)
	}
}

// row reads the continuation columns of the credential's signup row (superuser).
type sessionRow struct {
	Generation                                  int
	Provisioned, RefConsumed, NonceConsumed     bool
	ProofSet, ProofConsumed, Completed, Revoked bool
	ChallengeSet, NonceSet                      bool
}

func readRow(t *testing.T, info identity.SignupCredentialInfo) sessionRow {
	t.Helper()
	var r sessionRow
	err := setupConn(t).QueryRow(context.Background(), `
SELECT continuation_generation, provisioned_at IS NOT NULL, ref_consumed_at IS NOT NULL, nonce_consumed_at IS NOT NULL,
       proof_hash IS NOT NULL, proof_consumed_at IS NOT NULL, completed_at IS NOT NULL, revoked_at IS NOT NULL,
       challenge_hash IS NOT NULL, nonce_hash IS NOT NULL
FROM app.signup_sessions WHERE credential_hash = $1`, info.Handle.Digest()).Scan(&r.Generation, &r.Provisioned, &r.RefConsumed,
		&r.NonceConsumed, &r.ProofSet, &r.ProofConsumed, &r.Completed, &r.Revoked, &r.ChallengeSet, &r.NonceSet)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func query(raw, key string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}

func originOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func newBinding(t *testing.T) string {
	t.Helper()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return encode(b)
}

func wantCode(t *testing.T, what string, err error, code Code) {
	t.Helper()
	if CodeOf(err) != code {
		t.Errorf("%s: got %v, want %s", what, err, code)
	}
}

// withLabels is a service sharing h's identity and clock whose platform labels come from r, with reserved as
// PLATFORM_RESERVED_LABELS (G057 T4: deterministic label draws through the real writer body).
func (h *harness) withLabels(r io.Reader, reserved ...string) *Service {
	return NewService(Config{Identity: h.id, RootOrigin: rootOrigin, Now: h.clk.Now, LabelRand: r,
		Hostname: hostname.Config{BaseDomain: "localhost", Port: "543", AuthHost: "auth.localhost", ReservedLabels: reserved}})
}

// randomBlock is one label block from crypto/rand, so deterministic label sequences never collide across runs.
func randomBlock(t *testing.T) []byte {
	t.Helper()
	b := make([]byte, LabelBytes)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return b
}

// storedAddress returns the platform label of the Organization holding slug and the target origin stored on the
// credential's signup row (superuser setup reads).
func storedAddress(t *testing.T, slug string, info identity.SignupCredentialInfo) (label, origin string) {
	t.Helper()
	var l, o *string
	err := setupConn(t).QueryRow(context.Background(), `SELECT (SELECT platform_label FROM app.organizations WHERE slug = $1),
       (SELECT target_origin FROM app.signup_sessions WHERE credential_hash = $2)`, slug, info.Handle.Digest()).Scan(&l, &o)
	if err != nil {
		t.Fatal(err)
	}
	if l != nil {
		label = *l
	}
	if o != nil {
		origin = *o
	}
	return label, origin
}

// labelOrigin is the platform label host origin of the overlay.
func labelOrigin(label string) string { return "https://" + label + ".localhost:543" }

// routes counts the app.hostname_state platform rows for key (superuser setup read of the routing reader).
func routes(t *testing.T, key string) int {
	t.Helper()
	var n int
	if err := setupConn(t).QueryRow(context.Background(), `SELECT count(*) FROM app.hostname_state('platform', $1)`, key).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// assertGenerated checks a stored platform label: <adjective>-<noun>-<four digits> (the generator's shape).
func assertGenerated(t *testing.T, label string) {
	t.Helper()
	if !generatedShape.MatchString(label) {
		t.Errorf("platform label %q is not a generated adjective-noun-four-digit label", label)
	}
}

// TestSignupProvisioning (SUP-5, SUP-7): atomic creation with the two captures attributed to the creator; injected
// failures after the Organization INSERT and after the binding retain nothing; lost response / acknowledged-create retry /
// credential gone; changed payload; concurrent identical retries; disabled creator and the disable race; six provisions on
// one backend under plan_cache_mode = force_generic_plan.
func TestSignupProvisioning(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)

	t.Run("created", func(t *testing.T) {
		cred, info := h.credential(t, uniqueSubject(t), "sid-1")
		if StatusOf(info) != StatusPending {
			t.Fatalf("status %s", StatusOf(info))
		}
		slug := uniqueSlug(t)
		p, err := h.svc.Provision(ctx, info, "  Acme Holdings  ", slug)
		if err != nil || p.Outcome != OutcomeCreated || p.Generation != 1 || p.Slug != slug || p.OrganizationID == "" {
			t.Fatalf("provision: %+v %v", p, err)
		}
		// G057 T4: the Organization carries a generated platform label, the continuation and the stored target origin are
		// that label's host, the slug is metadata only (its host routes nowhere) and the label host routes to the UUID.
		label, origin := storedAddress(t, slug, info)
		assertGenerated(t, label)
		if originOf(p.Continue) != labelOrigin(label) || origin != labelOrigin(label) || query(p.Continue, "ref") == "" ||
			strings.Contains(p.Continue, slug) {
			t.Errorf("continue %q, stored origin %q; want the label origin %s", p.Continue, origin, labelOrigin(label))
		}
		if n := routes(t, slug); n != 0 {
			t.Errorf("the slug host routes (%d rows)", n)
		}
		var routed string
		if err := setupConn(t).QueryRow(ctx, `SELECT organization_id::text FROM app.hostname_state('platform', $1)`, label).Scan(&routed); err != nil ||
			routed != p.OrganizationID {
			t.Errorf("label host routes to %q (%v), want %s", routed, err, p.OrganizationID)
		}
		assertOnePair(t, slug)
		var name, actor string
		if err := setupConn(t).QueryRow(ctx, `SELECT o.name, (SELECT e.actor_user_id::text FROM audit.events e WHERE e.organization_id = o.id
  AND e.action = 'organization.created') FROM app.organizations o WHERE o.slug = $1`, slug).Scan(&name, &actor); err != nil {
			t.Fatal(err)
		}
		if name != "Acme Holdings" || actor != info.Handle.UserID() {
			t.Errorf("stored name %q actor %q (want trimmed name, creator %s)", name, actor, info.Handle.UserID())
		}
		if st := StatusOf(h.resolve(t, cred)); st != StatusCreated {
			t.Errorf("status after create %s", st)
		}
	})

	// G057 T4 (Plan r5 §3.2): a generated value reserved by configuration is drawn first and never issued; it reaches
	// no statement, and the row carries the second draw.
	t.Run("a configured reserved generated value is never issued", func(t *testing.T) {
		rb, wb := randomBlock(t), randomBlock(t)
		reserved, want := labelOf(t, rb), labelOf(t, wb)
		r := &labelSeq{blocks: [][]byte{rb, wb}}
		svc := h.withLabels(r, reserved)
		_, info := h.credential(t, uniqueSubject(t), "sid-rl")
		slug := uniqueSlug(t)
		p, err := svc.Provision(ctx, info, "Reserved Label", slug)
		if err != nil || p.Outcome != OutcomeCreated || r.draws != 2 {
			t.Fatalf("provision: %+v %v after %d draws", p, err, r.draws)
		}
		if label, origin := storedAddress(t, slug, info); label != want || origin != labelOrigin(want) || originOf(p.Continue) != labelOrigin(want) {
			t.Errorf("issued %q (%q, continue %q); want the second draw %q", label, origin, p.Continue, want)
		}
		if n := routes(t, reserved); n != 0 {
			t.Errorf("the reserved value routes (%d rows)", n)
		}
		assertOnePair(t, slug)
	})

	t.Run("injected failures retain nothing", func(t *testing.T) {
		sc := setupConn(t)
		if _, err := sc.Exec(ctx, `
CREATE FUNCTION public.g062_fault() RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog AS $$
BEGIN
  IF EXISTS (SELECT 1 FROM app.organizations o WHERE o.id = NEW.organization_id AND o.name = TG_ARGV[0]) THEN
    RAISE EXCEPTION 'g062 injected fault';
  END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER g062_fault_binding BEFORE INSERT ON app.role_bindings FOR EACH ROW EXECUTE FUNCTION public.g062_fault('Fault Binding');
CREATE TRIGGER g062_fault_audit BEFORE INSERT ON audit.events FOR EACH ROW EXECUTE FUNCTION public.g062_fault('Fault Audit');`); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_, _ = setupConn(t).Exec(context.Background(), `DROP TRIGGER IF EXISTS g062_fault_binding ON app.role_bindings;
DROP TRIGGER IF EXISTS g062_fault_audit ON audit.events; DROP FUNCTION IF EXISTS public.g062_fault()`)
		})
		for _, name := range []string{"Fault Binding", "Fault Audit"} {
			_, info := h.credential(t, uniqueSubject(t), "sid-f")
			slug := uniqueSlug(t)
			_, err := h.svc.Provision(ctx, info, name, slug)
			wantCode(t, name, err, CodeUnavailable)
			if o, w, c, g := counts(t, slug); o+w+c+g != 0 {
				t.Errorf("%s retained writes: %d %d %d %d", name, o, w, c, g)
			}
			if r := readRow(t, info); r.Provisioned || r.Generation != 0 {
				t.Errorf("%s: signup row changed %+v", name, r)
			}
			// the same credential still provisions another intent afterwards
			if p, err := h.svc.Provision(ctx, info, "Recovered", slug); err != nil || p.Outcome != OutcomeCreated {
				t.Errorf("%s: recovery provision %+v %v", name, p, err)
			}
		}
	})

	t.Run("lost response, acknowledged retry, re-mint resets the proof", func(t *testing.T) {
		cred, info := h.credential(t, uniqueSubject(t), "sid-2")
		slug := uniqueSlug(t)
		h.svc.afterProvision = func() error { return errors.New("connection reset after COMMIT") }
		_, err := h.svc.Provision(ctx, info, "Lost Response", slug)
		h.svc.afterProvision = nil
		wantCode(t, "lost response", err, CodeUnavailable)
		// the retry draws a different label: already_created answers the stored origin, never the retry's label
		label, origin := storedAddress(t, slug, info)
		assertGenerated(t, label)
		retryBlock := randomBlock(t)
		retryLabel := labelOf(t, retryBlock)
		h.svc.cfg.LabelRand = bytes.NewReader(retryBlock)
		p2, err := h.svc.Provision(ctx, info, "Lost Response", slug)
		h.svc.cfg.LabelRand = nil
		if err != nil || p2.Outcome != OutcomeAlreadyCreated || p2.Generation != 2 || p2.Slug != slug {
			t.Fatalf("(a) retry: %+v %v", p2, err)
		}
		if retryLabel == label || originOf(p2.Continue) != origin || origin != labelOrigin(label) || strings.Contains(p2.Continue, retryLabel) {
			t.Errorf("(a) retry continue %q (retry label %q); want the stored origin %s", p2.Continue, retryLabel, origin)
		}
		if l, o := storedAddress(t, slug, info); l != label || o != origin || routes(t, retryLabel) != 0 {
			t.Errorf("(a) retry changed the address: %q %q (retry label routes: %d)", l, o, routes(t, retryLabel))
		}
		assertOnePair(t, slug)
		// (b) acknowledged create, continuation through step 3, then a failed hop: an identical re-POST re-mints g+1 and
		// resets nonce_consumed_at and proof_hash atomically under the provision writer.
		target := originOf(p2.Continue)
		b := newBinding(t)
		ch, err := h.svc.Challenge(ctx, target, b, query(p2.Continue, "ref"))
		if err != nil {
			t.Fatalf("challenge: %v", err)
		}
		complete, err := h.svc.Prove(ctx, h.resolve(t, cred), query(ch.Prove, "n"))
		if err != nil {
			t.Fatalf("prove: %v", err)
		}
		if r := readRow(t, info); !r.NonceConsumed || !r.ProofSet {
			t.Fatalf("after prove: %+v", r)
		}
		p3, err := h.svc.Provision(ctx, h.resolve(t, cred), "Lost Response", slug)
		if err != nil || p3.Outcome != OutcomeAlreadyCreated || p3.Generation != 3 {
			t.Fatalf("(b) retry: %+v %v", p3, err)
		}
		if r := readRow(t, info); r.Generation != 3 || r.RefConsumed || r.ChallengeSet || r.NonceSet || r.NonceConsumed || r.ProofSet ||
			r.ProofConsumed || r.Completed {
			t.Errorf("after re-mint: %+v (want generation 3 with every continuation column reset)", r)
		}
		if _, err := h.svc.Complete(ctx, target, b, ch.Value, query(complete, "p")); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("generation-2 secrets after re-mint: %v", err)
		}
		if _, err := h.svc.Challenge(ctx, target, b, query(p2.Continue, "ref")); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("generation-2 reference after re-mint: %v", err)
		}
		assertOnePair(t, slug)
		// the same re-mint as the runtime: no privilege on the continuation columns
		var pg *pgconn.PgError
		_, err = runtimeConn(t).Exec(ctx, `UPDATE app.signup_sessions SET nonce_consumed_at = NULL, proof_hash = NULL, ref_hash = NULL WHERE false`)
		if !errors.As(err, &pg) || pg.Code != "42501" {
			t.Errorf("runtime re-mint: %v, want 42501", err)
		}
		// changed payload
		if _, err := h.svc.Provision(ctx, h.resolve(t, cred), "Other Name", slug); CodeOf(err) != CodeIntentUsed {
			t.Errorf("changed name: %v", err)
		}
		if _, err := h.svc.Provision(ctx, h.resolve(t, cred), "Lost Response", uniqueSlug(t)); CodeOf(err) != CodeIntentUsed {
			t.Errorf("changed slug: %v", err)
		}
		// (c) credential gone: abandoned ⇒ unauthenticated; no re-mint
		info = h.resolve(t, cred)
		if err := h.svc.Abandon(ctx, info); err != nil {
			t.Fatal(err)
		}
		if _, err := h.svc.Resolve(ctx, cred, rootOrigin); CodeOf(err) != CodeUnauthenticated {
			t.Errorf("resolve after abandon: %v", err)
		}
		if _, err := h.svc.Provision(ctx, info, "Lost Response", slug); CodeOf(err) != CodeUnauthenticated {
			t.Errorf("provision with a revoked credential: %v", err)
		}
		if r := readRow(t, info); r.Generation != 3 || !r.Revoked {
			t.Errorf("after abandon: %+v", r)
		}
		assertOnePair(t, slug)
	})

	t.Run("expired credential", func(t *testing.T) {
		cred, info := h.credential(t, uniqueSubject(t), "sid-3")
		h.clk.Advance(identity.SignupLifetime + time.Second)
		if _, err := h.svc.Resolve(ctx, cred, rootOrigin); CodeOf(err) != CodeUnauthenticated {
			t.Errorf("resolve expired: %v", err)
		}
		if _, err := h.svc.Provision(ctx, info, "Too Late", uniqueSlug(t)); CodeOf(err) != CodeUnauthenticated {
			t.Errorf("provision expired: %v", err)
		}
	})

	t.Run("concurrent identical retries", func(t *testing.T) {
		cred, info := h.credential(t, uniqueSubject(t), "sid-4")
		slug := uniqueSlug(t)
		if _, err := h.svc.Provision(ctx, info, "Concurrent", slug); err != nil {
			t.Fatal(err)
		}
		info = h.resolve(t, cred)
		var wg sync.WaitGroup
		res := make([]Provisioned, 4)
		errs := make([]error, 4)
		for i := range res {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				res[i], errs[i] = h.svc.Provision(ctx, info, "Concurrent", slug)
			}(i)
		}
		wg.Wait()
		gens := map[int]bool{}
		for i := range res {
			if errs[i] != nil || res[i].Outcome != OutcomeAlreadyCreated {
				t.Fatalf("retry %d: %+v %v", i, res[i], errs[i])
			}
			gens[res[i].Generation] = true
		}
		if len(gens) != 4 || !gens[2] || !gens[5] {
			t.Errorf("generations %v, want 2..5 each once (serialized on the row)", gens)
		}
		if r := readRow(t, info); r.Generation != 5 {
			t.Errorf("live generation %d, want 5", r.Generation)
		}
		live := 0
		for i := range res {
			if _, err := h.svc.Challenge(ctx, originOf(res[i].Continue), newBinding(t), query(res[i].Continue, "ref")); err == nil {
				live++
			}
		}
		if live != 1 {
			t.Errorf("%d live references, want exactly one", live)
		}
		assertOnePair(t, slug)
	})

	t.Run("disabled creator and disable race", func(t *testing.T) {
		_, info := h.credential(t, uniqueSubject(t), "sid-5")
		if err := h.id.DisableUser(ctx, info.Handle.UserID()); err != nil {
			t.Fatal(err)
		}
		slug := uniqueSlug(t)
		if _, err := h.svc.Provision(ctx, info, "Disabled", slug); CodeOf(err) != CodeUserDisabled && CodeOf(err) != CodeUnauthenticated {
			t.Errorf("disabled creator: %v", err)
		}
		if o, _, _, _ := counts(t, slug); o != 0 {
			t.Error("a disabled creator created an Organization")
		}
		for i := 0; i < 5; i++ {
			_, info := h.credential(t, uniqueSubject(t), "sid-r")
			slug := uniqueSlug(t)
			var wg sync.WaitGroup
			var perr, derr error
			wg.Add(2)
			go func() { defer wg.Done(); _, perr = h.svc.Provision(ctx, info, "Race", slug) }()
			go func() { defer wg.Done(); derr = h.id.DisableUser(ctx, info.Handle.UserID()) }()
			wg.Wait()
			if derr != nil {
				t.Fatal(derr)
			}
			o, w, c, g := counts(t, slug)
			switch CodeOf(perr) {
			case "":
				if o != 1 || w != 1 || c != 1 || g != 1 {
					t.Errorf("race %d created: %v %d %d %d %d", i, perr, o, w, c, g)
				}
			case CodeUserDisabled, CodeUnauthenticated:
				if o+w+c+g != 0 {
					t.Errorf("race %d refused but retained %d %d %d %d", i, o, w, c, g)
				}
			default:
				t.Errorf("race %d: %v", i, perr)
			}
		}
	})

	// Generic plan (G062 round-5 correction, review note 14): the whole flow, the provisioning transaction included, runs
	// on one backend with plan_cache_mode = force_generic_plan, six times. The audit attribution body evaluated for the
	// provision writer's captures must not need a privilege that writer lacks on a branch it does not take: each provision
	// creates the Organization, its org_owner binding and the organization.created / membership.granted captures.
	t.Run("generic plan on the provisioning backend", func(t *testing.T) {
		pool := genericPlanPool(t)
		g := h.on(pool)
		backend := func() (int32, string) {
			t.Helper()
			var pid int32
			var mode string
			if err := pool.QueryRow(ctx, `SELECT pg_backend_pid(), current_setting('plan_cache_mode')`).Scan(&pid, &mode); err != nil {
				t.Fatal(err)
			}
			return pid, mode
		}
		pid, mode := backend()
		if mode != "force_generic_plan" {
			t.Fatalf("plan_cache_mode %q on the provisioning backend", mode)
		}
		for i := 0; i < 6; i++ {
			_, info := g.credential(t, uniqueSubject(t), fmt.Sprintf("sid-g%d", i))
			slug := uniqueSlug(t)
			p, err := g.svc.Provision(ctx, info, fmt.Sprintf("Generic Plan %d", i), slug)
			if err != nil || p.Outcome != OutcomeCreated || p.Slug != slug || p.OrganizationID == "" {
				t.Fatalf("provision %d under a generic plan: %+v %v", i, p, err)
			}
			assertOnePair(t, slug)
			if p2, m2 := backend(); p2 != pid || m2 != mode {
				t.Fatalf("provision %d ran on backend %d (%s), want %d (%s)", i, p2, m2, pid, mode)
			}
		}
	})
}

// TestSignupSlugRace (SUP-6): two users race distinct intents on one slug: exactly one created, the loser slug_taken with
// zero retained writes and a still-pending row; the winner's platform label resolves through app.hostname_state to the
// winner only and the slug routes nowhere (G057). A forced label collision is retried within the label budget.
func TestSignupSlugRace(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)
	for i := 0; i < 5; i++ {
		_, a := h.credential(t, uniqueSubject(t), "sid-a")
		_, b := h.credential(t, uniqueSubject(t), "sid-b")
		slug := uniqueSlug(t)
		var wg sync.WaitGroup
		var pa, pb Provisioned
		var ea, eb error
		wg.Add(2)
		go func() { defer wg.Done(); pa, ea = h.svc.Provision(ctx, a, "Alpha", slug) }()
		go func() { defer wg.Done(); pb, eb = h.svc.Provision(ctx, b, "Beta", slug) }()
		wg.Wait()
		winner, loser, wp, le := a, b, pa, eb
		if ea != nil {
			winner, loser, wp, le = b, a, pb, ea
		}
		if wp.Outcome != OutcomeCreated || CodeOf(le) != CodeSlugTaken {
			t.Fatalf("race %d: %+v %v / %+v %v", i, pa, ea, pb, eb)
		}
		assertOnePair(t, slug)
		if r := readRow(t, loser); r.Provisioned {
			t.Errorf("race %d: loser row provisioned", i)
		}
		// G057: the winner's platform label routes to the winner; the slug itself routes nowhere
		label, _ := storedAddress(t, slug, winner)
		var org string
		if err := setupConn(t).QueryRow(ctx, `SELECT organization_id::text FROM app.hostname_state('platform', $1)`, label).Scan(&org); err != nil || org != wp.OrganizationID {
			t.Errorf("race %d: hostname_state %q %v, want %s", i, org, err, wp.OrganizationID)
		}
		if n := routes(t, slug); n != 0 {
			t.Errorf("race %d: the slug host routes (%d rows)", i, n)
		}
		// the loser's suggestion skips the taken slug
		sg, err := h.svc.SuggestSlug(ctx, loser, slug)
		if err != nil || !sg.Available || sg.Suggested == slug || !strings.HasPrefix(sg.Suggested, slug+"-") {
			t.Errorf("race %d: suggestion %+v %v", i, sg, err)
		}
	}
	// reserved labels never reach SQL; a configured one is refused like the floor
	_, c := h.credential(t, uniqueSubject(t), "sid-c")
	for _, s := range []string{"www", "auth", "api", "example.com", "a.b", "acme.localhost"} {
		if _, err := h.svc.Provision(ctx, c, "Reserved", s); CodeOf(err) != CodeInvalidRequest {
			t.Errorf("reserved %q: %v", s, err)
		}
	}

	// G057 T4 (Plan r5 §3.5 concurrency): two users with distinct slugs draw the same first label (deterministic reader
	// doubles). The first commit wins the label index; the other's body answers label_taken with nothing retained, and
	// its bounded retry creates the Organization with its second draw. Both are created, the labels differ, each label
	// routes to its own Organization.
	t.Run("forced label collision", func(t *testing.T) {
		shared, secondA, secondB := randomBlock(t), randomBlock(t), randomBlock(t)
		ra := &labelSeq{blocks: [][]byte{shared, secondA}}
		rb := &labelSeq{blocks: [][]byte{shared, secondB}}
		sa, sb := h.withLabels(ra), h.withLabels(rb)
		_, a := h.credential(t, uniqueSubject(t), "sid-la")
		_, b := h.credential(t, uniqueSubject(t), "sid-lb")
		slugA, slugB := uniqueSlug(t), uniqueSlug(t)
		var wg sync.WaitGroup
		var pa, pb Provisioned
		var ea, eb error
		wg.Add(2)
		go func() { defer wg.Done(); pa, ea = sa.Provision(ctx, a, "Label A", slugA) }()
		go func() { defer wg.Done(); pb, eb = sb.Provision(ctx, b, "Label B", slugB) }()
		wg.Wait()
		if ea != nil || eb != nil || pa.Outcome != OutcomeCreated || pb.Outcome != OutcomeCreated {
			t.Fatalf("collision: %+v %v / %+v %v", pa, ea, pb, eb)
		}
		la, oa := storedAddress(t, slugA, a)
		lb, ob := storedAddress(t, slugB, b)
		first := labelOf(t, shared)
		aWon := la == first && ra.draws == 1 && lb == labelOf(t, secondB) && rb.draws == 2
		bWon := lb == first && rb.draws == 1 && la == labelOf(t, secondA) && ra.draws == 2
		if aWon == bWon {
			t.Errorf("labels %q / %q after %d / %d draws; want the shared first draw issued once and the other retried with its second draw",
				la, lb, ra.draws, rb.draws)
		}
		if la == lb || oa != labelOrigin(la) || ob != labelOrigin(lb) || originOf(pa.Continue) != oa || originOf(pb.Continue) != ob {
			t.Errorf("addresses %q %q / %q %q; continues %q / %q", la, oa, lb, ob, pa.Continue, pb.Continue)
		}
		for _, c := range []struct{ label, org string }{{la, pa.OrganizationID}, {lb, pb.OrganizationID}} {
			var org string
			if err := setupConn(t).QueryRow(ctx, `SELECT organization_id::text FROM app.hostname_state('platform', $1)`, c.label).Scan(&org); err != nil || org != c.org {
				t.Errorf("label %q routes to %q (%v), want %s", c.label, org, err, c.org)
			}
		}
		assertOnePair(t, slugA)
		assertOnePair(t, slugB)
	})
}

// continuation runs one provisioning and returns its pieces.
type flow struct {
	cred   string
	info   identity.SignupCredentialInfo
	slug   string
	target string
	prov   Provisioned
}

func (h *harness) provisioned(t *testing.T, subject string) flow {
	t.Helper()
	cred, info := h.credential(t, subject, "sid-"+subject[len(subject)-4:])
	slug := uniqueSlug(t)
	p, err := h.svc.Provision(context.Background(), info, "Continuation "+slug, slug)
	if err != nil {
		t.Fatalf("provision: %v", err)
	}
	return flow{cred: cred, info: h.resolve(t, cred), slug: slug, target: originOf(p.Continue), prov: p}
}

// pinnedCallback finishes the pinned login with subject at the provider.
func (h *harness) pinnedCallback(target string, out identity.LoginOutput, subject, sid string) (identity.CallbackOutput, error) {
	u, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		return identity.CallbackOutput{}, err
	}
	code := h.fake.IssueCode(u.Query().Get("code_challenge"), grant(subject, sid))
	return h.id.Callback(context.Background(), identity.CallbackInput{Origin: target, RawQuery: callbackQuery(code, u.Query().Get("state")),
		Binding: out.Binding})
}

// TestSignupContinuation (SUP-8): the four hops with two simulated browsers and two root credentials; every refusal is
// "no matching row, nothing consumed" unless a consumption is named.
func TestSignupContinuation(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)

	t.Run("happy path to tenant_created", func(t *testing.T) {
		subject := uniqueSubject(t)
		f := h.provisioned(t, subject)
		b := newBinding(t)
		ch, err := h.svc.Challenge(ctx, f.target, b, query(f.prov.Continue, "ref"))
		if err != nil || ch.MaxAge != 120 || originOf(ch.Prove) != rootOrigin || ch.Value == "" {
			t.Fatalf("challenge: %+v %v", ch, err)
		}
		complete, err := h.svc.Prove(ctx, h.resolve(t, f.cred), query(ch.Prove, "n"))
		if err != nil || originOf(complete) != f.target {
			t.Fatalf("prove: %q %v", complete, err)
		}
		out, err := h.svc.Complete(ctx, f.target, b, ch.Value, query(complete, "p"))
		if err != nil || out.AuthorizationURL == "" || out.Binding != b {
			t.Fatalf("complete: %+v %v", out, err)
		}
		if r := readRow(t, f.info); !r.Completed {
			t.Errorf("row not completed: %+v", r)
		}
		if st := StatusOf(h.resolve(t, f.cred)); st != StatusCompleted {
			t.Errorf("status %s", st)
		}
		// same user, new provider session ⇒ accepted (no SID comparison)
		cb, err := h.pinnedCallback(f.target, out, subject, "sid-new-provider-session")
		if err != nil || cb.Outcome != identity.OutcomeTenantCreated || cb.ReturnTo != "/" || cb.Credential == "" {
			t.Fatalf("pinned callback: %+v %v", cb, err)
		}
		sess, err := h.id.Session(ctx, cb.Credential, f.target)
		if err != nil || sess.OrganizationID != f.prov.OrganizationID || sess.UserID != f.info.Handle.UserID() {
			t.Errorf("session: %+v %v", sess, err)
		}
		// each secret replayed after its legitimate use matches nothing
		if _, err := h.svc.Challenge(ctx, f.target, b, query(f.prov.Continue, "ref")); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("R replay: %v", err)
		}
		if _, err := h.svc.Prove(ctx, h.resolve(t, f.cred), query(ch.Prove, "n")); CodeOf(err) != CodeContextMismatch {
			t.Errorf("N replay: %v", err)
		}
		if _, err := h.svc.Complete(ctx, f.target, b, ch.Value, query(complete, "p")); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("C+P replay: %v", err)
		}
		assertOnePair(t, f.slug)
	})

	t.Run("captured R, captured N, captured P", func(t *testing.T) {
		f := h.provisioned(t, uniqueSubject(t))
		other := h.provisioned(t, uniqueSubject(t)).info // a provisioned row of another user (the SQL path)
		ref := query(f.prov.Continue, "ref")
		b1, b2 := newBinding(t), newBinding(t)
		// browser 2 spends R (the only consumption) and gets C2/N2
		ch2, err := h.svc.Challenge(ctx, f.target, b2, ref)
		if err != nil {
			t.Fatalf("browser 2 challenge: %v", err)
		}
		// browser 2 has no __Host-signup: with another user's credential the proof matches no row
		if _, err := h.svc.Prove(ctx, other, query(ch2.Prove, "n")); CodeOf(err) != CodeContextMismatch {
			t.Errorf("foreign credential prove: %v", err)
		}
		if r := readRow(t, f.info); r.NonceConsumed || r.ProofSet {
			t.Errorf("foreign prove changed the row: %+v", r)
		}
		// browser 1's continue finds R spent
		if _, err := h.svc.Challenge(ctx, f.target, b1, ref); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("browser 1 challenge after capture: %v", err)
		}
		// captured N: browser 1 (the original credential) is tricked into step 3 ⇒ P issued into browser 1, but
		// browser 1 lacks C2 ⇒ completion matches nothing and P stays unconsumed
		complete, err := h.svc.Prove(ctx, h.resolve(t, f.cred), query(ch2.Prove, "n"))
		if err != nil {
			t.Fatalf("captured-N prove: %v", err)
		}
		wrongC := encode(make([]byte, 32))
		if _, err := h.svc.Complete(ctx, f.target, b1, wrongC, query(complete, "p")); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("completion without C2: %v", err)
		}
		if r := readRow(t, f.info); r.ProofConsumed || r.Completed {
			t.Errorf("wrong-C completion changed the row: %+v", r)
		}
		// captured P with another generation's C after the original browser's "Try again"
		name := "Continuation " + f.slug
		p2, err := h.svc.Provision(ctx, h.resolve(t, f.cred), name, f.slug)
		if err != nil || p2.Outcome != OutcomeAlreadyCreated {
			t.Fatalf("try again: %+v %v", p2, err)
		}
		if _, err := h.svc.Complete(ctx, f.target, b2, ch2.Value, query(complete, "p")); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("generation-1 C+P after re-mint: %v", err)
		}
		if _, err := h.svc.Prove(ctx, h.resolve(t, f.cred), query(ch2.Prove, "n")); CodeOf(err) != CodeContextMismatch {
			t.Errorf("generation-1 N after re-mint: %v", err)
		}
		ch, err := h.svc.Challenge(ctx, f.target, b1, query(p2.Continue, "ref"))
		if err != nil {
			t.Fatalf("generation-2 challenge: %v", err)
		}
		c2, err := h.svc.Prove(ctx, h.resolve(t, f.cred), query(ch.Prove, "n"))
		if err != nil {
			t.Fatalf("generation-2 prove: %v", err)
		}
		if _, err := h.svc.Complete(ctx, f.target, b2, ch2.Value, query(c2, "p")); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("captured P with browser 2's stale C: %v", err)
		}
		if _, err := h.svc.Complete(ctx, f.target, b1, ch.Value, query(c2, "p")); err != nil {
			t.Errorf("browser 1 completion: %v", err)
		}
		assertOnePair(t, f.slug)
	})

	t.Run("wrong root context", func(t *testing.T) {
		subject := uniqueSubject(t)
		f := h.provisioned(t, subject)
		// another signup row of the same user (a second credential)
		_, same := h.credential(t, subject, "sid-second")
		ch, err := h.svc.Challenge(ctx, f.target, newBinding(t), query(f.prov.Continue, "ref"))
		if err != nil {
			t.Fatal(err)
		}
		n := query(ch.Prove, "n")
		if _, err := h.svc.Prove(ctx, same, n); CodeOf(err) != CodeContextMismatch {
			t.Errorf("same user, other row: %v", err)
		}
		// expired generation (clock advanced past 120 s at the proof)
		h.clk.Advance(ContinuationTTL + time.Second)
		if _, err := h.svc.Prove(ctx, h.resolve(t, f.cred), n); CodeOf(err) != CodeContextMismatch {
			t.Errorf("expired generation: %v", err)
		}
		if r := readRow(t, f.info); r.NonceConsumed {
			t.Errorf("refused proofs consumed the nonce: %+v", r)
		}
		// revoked credential (resolved before the abandon) matches nothing
		info := h.resolve(t, f.cred)
		if err := h.svc.Abandon(ctx, info); err != nil {
			t.Fatal(err)
		}
		if _, err := h.svc.Prove(ctx, info, n); CodeOf(err) != CodeContextMismatch {
			t.Errorf("revoked credential: %v", err)
		}
	})

	t.Run("wrong tenant host and expiry at the challenge", func(t *testing.T) {
		f := h.provisioned(t, uniqueSubject(t))
		g := h.provisioned(t, uniqueSubject(t))
		ref := query(f.prov.Continue, "ref")
		// G057: the hops run on the label host; the slug host is a wrong tenant host (it routes nowhere)
		labelName := strings.TrimSuffix(strings.TrimPrefix(f.target, "https://"), ":543")
		for _, origin := range []string{g.target, "https://" + labelName + ":544", "http://" + labelName + ":543",
			"https://localhost:543", "https://custom.example:543", "https://nowhere-" + f.slug + ".localhost:543",
			"https://" + f.slug + ".localhost:543"} {
			if _, err := h.svc.Challenge(ctx, origin, newBinding(t), ref); CodeOf(err) != CodeContinuationInvalid {
				t.Errorf("challenge at %s: %v", origin, err)
			}
		}
		if r := readRow(t, f.info); r.RefConsumed {
			t.Errorf("a mismatched origin consumed R: %+v", r)
		}
		h.clk.Advance(ContinuationTTL + time.Second)
		if _, err := h.svc.Challenge(ctx, f.target, newBinding(t), ref); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("expired R: %v", err)
		}
	})

	t.Run("completion at another host, account substitution, removed owner, host moved", func(t *testing.T) {
		run := func(t *testing.T) (flow, string, string, string) {
			subject := uniqueSubject(t)
			f := h.provisioned(t, subject)
			b := newBinding(t)
			ch, err := h.svc.Challenge(ctx, f.target, b, query(f.prov.Continue, "ref"))
			if err != nil {
				t.Fatal(err)
			}
			complete, err := h.svc.Prove(ctx, h.resolve(t, f.cred), query(ch.Prove, "n"))
			if err != nil {
				t.Fatal(err)
			}
			return f, b, ch.Value, query(complete, "p")
		}
		f, b, c, p := run(t)
		g := h.provisioned(t, uniqueSubject(t))
		if _, err := h.svc.Complete(ctx, g.target, b, c, p); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("completion at another tenant: %v", err)
		}
		out, err := h.svc.Complete(ctx, f.target, b, c, p)
		if err != nil {
			t.Fatal(err)
		}
		// account substitution at the provider ⇒ signup_user_mismatch, consumed, no session
		intruder := uniqueSubject(t)
		cb, err := h.pinnedCallback(f.target, out, intruder, "sid-x")
		var ie *identity.Error
		if !errors.As(err, &ie) || ie.Code != identity.ErrSignupUserMismatch.Code || cb.Credential != "" {
			t.Errorf("substitution: %+v %v", cb, err)
		}
		var mapped int
		if err := setupConn(t).QueryRow(ctx, `SELECT count(*) FROM app.identity_principals WHERE subject = $1`, intruder).Scan(&mapped); err != nil || mapped != 0 {
			t.Errorf("substitution created a mapping (%d, %v)", mapped, err)
		}
		// removed owner binding before the completion ⇒ refused (no pinned row)
		f2, b2, c2, p2 := run(t)
		superNoTriggers(t, "app.role_bindings", `DELETE FROM app.role_bindings WHERE organization_id = $1::uuid`, f2.prov.OrganizationID)
		if _, err := h.svc.Complete(ctx, f2.target, b2, c2, p2); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("removed owner: %v", err)
		}
		if r := readRow(t, f2.info); !r.Completed {
			t.Errorf("removed owner: the generation should be completed with no login row (recovery (b)): %+v", r)
		}
		// host moved between completion and callback ⇒ host_changed
		f3, b3, c3, p3 := run(t)
		out3, err := h.svc.Complete(ctx, f3.target, b3, c3, p3)
		if err != nil {
			t.Fatal(err)
		}
		// G057: a slug change is metadata and moves nothing; only the platform label is the host (moved here as setup)
		label3 := strings.TrimSuffix(strings.TrimPrefix(f3.target, "https://"), ".localhost:543")
		superNoTriggers(t, "app.organizations", `UPDATE app.organizations SET slug = $2 WHERE id = $1::uuid`, f3.prov.OrganizationID, uniqueSlug(t))
		if n := routes(t, label3); n != 1 {
			t.Errorf("a slug change moved the label host (%d rows)", n)
		}
		superNoTriggers(t, "app.organizations", `UPDATE app.organizations SET platform_label = $2 WHERE id = $1::uuid`, f3.prov.OrganizationID,
			labelOf(t, randomBlock(t)))
		_, err = h.pinnedCallback(f3.target, out3, subjectOf(t, f3.info), "sid-y")
		if !errors.As(err, &ie) || ie.Code != identity.ErrHostChanged.Code {
			t.Errorf("host moved: %v", err)
		}
	})
}

// subjectOf returns the provider subject mapped to a signup credential's user (superuser read).
func subjectOf(t *testing.T, info identity.SignupCredentialInfo) string {
	t.Helper()
	var s string
	if err := setupConn(t).QueryRow(context.Background(), `SELECT subject FROM app.identity_principals WHERE user_id = $1::uuid`,
		info.Handle.UserID()).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// superNoTriggers runs one fixture statement as the superuser with table's user triggers disabled, in one transaction
// (setup only: removes an owner binding or moves a slug without the guards and captures).
func superNoTriggers(t *testing.T, table, stmt string, args ...any) {
	t.Helper()
	ctx := context.Background()
	tx, err := setupConn(t).Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "ALTER TABLE "+table+" DISABLE TRIGGER USER"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, stmt, args...); err != nil {
		t.Fatalf("%s: %v", stmt, err)
	}
	if _, err := tx.Exec(ctx, "ALTER TABLE "+table+" ENABLE TRIGGER USER"); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
