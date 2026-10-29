package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/time/rate"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/identity"
	"core/internal/tenancy/access"
)

// G058 ENR-H (Plan r3 §5.2, §6.2 ENR-H1..H9; r0 §6.2 group 5; HTTP execution release 2026-10-18): the enrollment
// adapter's closed surface without a database — the method/path pairs and their closure, the root and sub-paths, the
// public projection end to end over a scripted resolver and the two definer reads, the applicant credential's
// pre-facade refusals (scheme, repeated header, origin, dead, foreign-origin and reassigned credential, CSRF), strict
// bodies and the budget, the limiter, the refusal mapping with its field codes, the minimal-disclosure answer shapes, the
// closed log event, and the bounded copy of onboarding/api. The facades run against real PostgreSQL in TestEnrollment,
// TestEnrollmentAdmission and TestEnrollmentRequests (access) and the identity callback in ENR-INTENT.

const (
	orgA  = "0f58a000-0000-4000-8000-0000000000a1"
	orgB  = "0f58a000-0000-4000-8000-0000000000b1"
	hostA = "https://org-a.localhost:543"
	rid   = "0f58a000-0000-4000-8000-00000000d001"
	uid   = "0f58a000-0000-4000-8000-00000000c001"
)

// fakeDB answers the resolver's hostname_state statement (org-a routes to orgA), the policy reader's public mode
// (open) and the auth reader's public branding (no row); anything else fails, and so does any transaction unless tx is
// set (TestEnrollmentAnswerShapes' scripted status transaction).
type fakeDB struct {
	stmts []string
	tx    pgx.Tx
}

type row struct{ scan func(...any) error }

func (r row) Scan(dest ...any) error { return r.scan(dest...) }

func (f *fakeDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	f.stmts = append(f.stmts, "begin")
	if f.tx != nil {
		return f.tx, nil
	}
	return nil, errors.New("fakeDB: no transaction")
}

// statusTx is the read-only status transaction of a live credential with no request and no admission: the context
// statement (seven columns) and the credential state (two NULL columns) answer, the own-request read finds no row.
type statusTx struct{ pgx.Tx }

func (statusTx) QueryRow(context.Context, string, ...any) pgx.Row {
	return row{func(dest ...any) error {
		if len(dest) == 7 || len(dest) == 2 {
			return nil
		}
		return pgx.ErrNoRows
	}}
}

func (statusTx) Commit(context.Context) error   { return nil }
func (statusTx) Rollback(context.Context) error { return nil }

func (f *fakeDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	switch {
	case strings.Contains(sql, "hostname_state"):
		f.stmts = append(f.stmts, "hostname_state")
		if args[0].(string)+"|"+args[1].(string) != "platform|org-a" {
			return row{func(...any) error { return pgx.ErrNoRows }}
		}
		return row{func(dest ...any) error {
			org := orgA
			*dest[0].(**string), *dest[1].(*bool), *dest[2].(*bool) = &org, true, false
			return nil
		}}
	case strings.Contains(sql, "app.rls_enrollment_mode($1)"):
		f.stmts = append(f.stmts, "rls_enrollment_mode")
		return row{func(dest ...any) error { *dest[0].(*string) = "open"; return nil }}
	case strings.Contains(sql, "app.public_branding($1)"):
		f.stmts = append(f.stmts, "public_branding")
		return row{func(...any) error { return pgx.ErrNoRows }}
	}
	f.stmts = append(f.stmts, "unexpected")
	return row{func(...any) error { return errors.New("fakeDB: unexpected statement") }}
}

// fakeCred is the applicant-credential boundary: a scripted resolution and membership answer.
type fakeCred struct {
	info               identity.EnrollmentCredentialInfo
	err                error
	member             bool
	calls              int
	lastCred, lastOrig string
}

func (f *fakeCred) EnrollmentCredential(_ context.Context, cred, origin string) (identity.EnrollmentCredentialInfo, error) {
	f.calls++
	f.lastCred, f.lastOrig = cred, origin
	return f.info, f.err
}

func (f *fakeCred) Membership(context.Context, string, string) (bool, error) { return f.member, nil }

// liveCred is a resolved credential of orgA with a known CSRF token.
func liveCred(org string) identity.EnrollmentCredentialInfo {
	csrf := make([]byte, 32)
	csrf[0] = 9
	return identity.EnrollmentCredentialInfo{Handle: identity.EnrollmentHandle{UserID: uid, OrganizationID: org, Origin: hostA,
		Digest: [32]byte{1}}, CSRFToken: csrf, ExpiresAt: time.Unix(1800, 0)}
}

// goodCSRF is the base64url form of liveCred's token.
var goodCSRF = base64.RawURLEncoding.EncodeToString(liveCred(orgA).CSRFToken)

func fixture(t *testing.T, cred credentials, log *slog.Logger) (*http.ServeMux, *fakeDB) {
	t.Helper()
	return fixtureWith(t, cred, log, newLimiter(credentialRate, credentialBurst), newLimiter(publicRate, publicBurst))
}

// fixtureWith is fixture with the given limiters.
func fixtureWith(t *testing.T, cred credentials, log *slog.Logger, own, public *limiter) (*http.ServeMux, *fakeDB) {
	t.Helper()
	db := &fakeDB{}
	d := config.Domains{BaseDomain: "localhost", CNAMETarget: "edge.localhost", HostPort: "543", AuthHost: "auth.localhost"}
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	h := &Handler{cfg: Config{Resolver: domains.NewResolver(db, d.Hostnames())}, acc: access.NewService(db, nil, nil), cred: cred,
		log: log, own: own, public: public}
	mux := http.NewServeMux()
	h.Register(mux)
	return mux, db
}

func serve(mux *http.ServeMux, method, target string, body []byte, ct string, auth ...string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, bytes.NewReader(body))
	if ct != "" {
		r.Header.Set("Content-Type", ct)
	}
	for _, a := range auth {
		r.Header.Add("Authorization", a)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}

// closedAnswer asserts a closed JSON refusal with the no-store headers and no CORS header, and returns its body.
func closedAnswer(t *testing.T, name string, w *httptest.ResponseRecorder, status int, code string) map[string]any {
	t.Helper()
	hd := w.Header()
	if hd.Get("Cache-Control") != "no-store" || hd.Get("X-Content-Type-Options") != "nosniff" || hd.Get("Content-Type") != "application/json" {
		t.Errorf("%s: headers %v", name, hd)
	}
	for k := range hd {
		if strings.HasPrefix(strings.ToLower(k), "access-control-") {
			t.Errorf("%s: CORS header %s", name, k)
		}
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("%s: body %q: %v", name, w.Body.String(), err)
	}
	if w.Code != status || out["error"] != code {
		t.Errorf("%s: %d %v, want %d %s", name, w.Code, out, status, code)
	}
	return out
}

func q(origin string) string { return "?" + url.Values{"origin": {origin}}.Encode() }

func jsonBody(t *testing.T, v map[string]any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

const enr = "Enrollment " + "EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE"

// The pairs and their trust shape: "public" (none), "own" (the applicant's enrollment credential), "session".
var pairs = []struct{ method, path, shape string }{
	{"GET", "/internal/v1/enrollment/public", "public"}, {"GET", "/internal/v1/enrollment/status", "own"},
	{"POST", "/internal/v1/enrollment/request", "own"}, {"POST", "/internal/v1/enrollment/admit", "own"},
	{"POST", "/internal/v1/enrollment/abandon", "own"}, {"GET", "/internal/v1/enrollment/policy", "session"},
	{"POST", "/internal/v1/enrollment/policy", "session"}, {"GET", "/internal/v1/enrollment/requests", "session"},
	{"POST", "/internal/v1/enrollment/requests/" + rid + "/decide", "session"},
	{"POST", "/internal/v1/enrollment/requests/" + rid + "/revoke", "session"},
}

func mutationBody(path, origin string) map[string]any {
	b := map[string]any{"origin": origin, "csrf_token": goodCSRF}
	switch {
	case strings.HasSuffix(path, "/admit"):
		b["kind"] = "open"
	case strings.HasSuffix(path, "/policy"):
		b["mode"] = "request_access"
	case strings.HasSuffix(path, "/decide"):
		b["decision"] = "approve"
	}
	return b
}

// TestEnrollmentHTTP (ENR-H1..H9 closed surface; ENR-4 d pending-credential binding): every pair's pre-facade refusals,
// the method/path closure, the root and sub-paths; no statement other than the resolver's and the public reads runs.
func TestEnrollmentHTTP(t *testing.T) {
	cred := &fakeCred{err: identity.ErrUnauthenticated}
	mux, db := fixture(t, cred, nil)
	if len(Paths) != 9 || len(pairs) != 10 {
		t.Fatalf("%d paths and %d pairs, want 9 and 10", len(Paths), len(pairs))
	}
	for _, p := range pairs {
		name := p.method + " " + p.path
		if p.method == "GET" {
			closedAnswer(t, name+" no origin", serve(mux, "GET", p.path, nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" two origins", serve(mux, "GET", p.path+q(hostA)+"&origin=x", nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" unknown parameter", serve(mux, "GET", p.path+q(hostA)+"&x=1", nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" http origin", serve(mux, "GET", p.path+q("http://org-a.localhost:543"), nil, ""), 403, "origin_rejected")
			if p.shape == "public" {
				closedAnswer(t, name+" unknown host", serve(mux, "GET", p.path+q("https://elsewhere.localhost:543"), nil, ""), 404, "unknown_host")
				continue
			}
			closedAnswer(t, name+" foreign origin", serve(mux, "GET", p.path+q("https://elsewhere.test:543"), nil, ""), 403, "origin_rejected")
			closedAnswer(t, name+" no credential", serve(mux, "GET", p.path+q(hostA), nil, ""), 401, "unauthenticated")
			closedAnswer(t, name+" two credentials", serve(mux, "GET", p.path+q(hostA), nil, "", "Session a", "Enrollment b"), 400, "ambiguous_parameter")
			continue
		}
		ok := jsonBody(t, mutationBody(p.path, hostA))
		closedAnswer(t, name+" no credential", serve(mux, p.method, p.path, ok, "application/json"), 401, "unauthenticated")
		closedAnswer(t, name+" foreign origin", serve(mux, p.method, p.path, jsonBody(t, mutationBody(p.path, "https://elsewhere.test:543")), "application/json"), 403, "origin_rejected")
		closedAnswer(t, name+" not JSON", serve(mux, p.method, p.path, ok, "text/plain"), 400, "invalid_request")
		closedAnswer(t, name+" query", serve(mux, p.method, p.path+q(hostA), ok, "application/json"), 400, "ambiguous_parameter")
		unknown := mutationBody(p.path, hostA)
		unknown["organization_id"] = orgA // a forged tenant is an unknown field
		closedAnswer(t, name+" unknown field", serve(mux, p.method, p.path, jsonBody(t, unknown), "application/json"), 400, "invalid_request")
		for _, forged := range []string{"user_id", "address", "email_verified", "role", "session"} {
			b := mutationBody(p.path, hostA)
			b[forged] = "x"
			closedAnswer(t, name+" forged "+forged, serve(mux, p.method, p.path, jsonBody(t, b), "application/json"), 400, "invalid_request")
		}
		for _, tail := range []string{" {}", "]", "}"} {
			closedAnswer(t, name+fmt.Sprintf(" trailing %q", tail), serve(mux, p.method, p.path, append(append([]byte{}, ok...), tail...), "application/json"), 400, "invalid_request")
		}
		closedAnswer(t, name+" malformed", serve(mux, p.method, p.path, []byte("{"), "application/json"), 400, "invalid_request")
		big := mutationBody(p.path, hostA)
		big["csrf_token"] = strings.Repeat("x", maxBody)
		closedAnswer(t, name+" over budget", serve(mux, p.method, p.path, jsonBody(t, big), "application/json"), 400, "invalid_request")
	}

	// Every method outside each path's closed set is method_not_allowed (closed JSON, not the mux's text answer).
	served := map[string]bool{}
	for _, p := range pairs {
		served[p.method+" "+p.path] = true
	}
	for _, p := range pairs {
		for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
			if !served[m+" "+p.path] {
				closedAnswer(t, m+" "+p.path, serve(mux, m, p.path+q(hostA), nil, ""), 405, "method_not_allowed")
			}
		}
	}
	// The root and sub-path probing: not_found (closed JSON) without a credential.
	for _, p := range []string{"/internal/v1/enrollment", "/internal/v1/enrollment/", "/internal/v1/enrollment/x",
		"/internal/v1/enrollment/requests/" + rid, "/internal/v1/enrollment/requests/" + rid + "/x",
		"/internal/v1/enrollment/requests/" + rid + "/approve", "/internal/v1/enrollment/status/x", "/internal/v1/enrollment/session",
		"/internal/v1/enrollment/issue"} {
		closedAnswer(t, "sub-path "+p, serve(mux, "GET", p+q(hostA), nil, ""), 404, "not_found")
	}
	if cred.calls != 0 {
		t.Fatalf("the credential was resolved %d times before an Enrollment header", cred.calls)
	}
	for _, s := range db.stmts {
		if s != "hostname_state" {
			t.Fatalf("a statement other than the resolver's ran: %v", db.stmts)
		}
	}
}

// TestEnrollmentPublic (ENR-H public projection; ENR-1 public half): the resolved Organization's mode and label through
// the two definer reads, no credential read; the answer carries exactly mode and label (no readiness, no member data).
func TestEnrollmentPublic(t *testing.T) {
	cred := &fakeCred{}
	mux, db := fixture(t, cred, nil)
	w := serve(mux, "GET", "/internal/v1/enrollment/public"+q(hostA), nil, "", enr)
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || w.Code != 200 || len(out) != 2 || out["mode"] != "open" ||
		out["label"] != nil {
		t.Fatalf("public: %d %v", w.Code, out)
	}
	if strings.Join(db.stmts, ",") != "hostname_state,rls_enrollment_mode,public_branding" || cred.calls != 0 {
		t.Errorf("public statements %v, credential reads %d", db.stmts, cred.calls)
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("public cache-control %q", w.Header().Get("Cache-Control"))
	}
}

// TestEnrollmentCredentialBinding (ENR-4 d, ENR-H applicant routes): only "Authorization: Enrollment <credential>" is
// read; a session credential or any other scheme is no credential (401); the credential is resolved at the request's
// exact origin; a dead credential (expired, revoked, disabled user) is 401; a credential minted for another
// Organization than the origin routes to now (a reassigned host) is 401; a store failure is 503; a mutation with a
// wrong CSRF token is 403 csrf_failed before any facade; a live credential reaches the facade (here unavailable: no
// transaction can open, so 503 enrollment_unavailable, never a guessed success).
func TestEnrollmentCredentialBinding(t *testing.T) {
	status := "/internal/v1/enrollment/status" + q(hostA)
	for _, auth := range []string{"Session " + strings.Repeat("E", 43), "Bearer x", "enrollment x", "Enrollment", "Enrollment a b"} {
		cred := &fakeCred{info: liveCred(orgA)}
		mux, _ := fixture(t, cred, nil)
		closedAnswer(t, "scheme "+auth, serve(mux, "GET", status, nil, "", auth), 401, "unauthenticated")
		if cred.calls != 0 {
			t.Errorf("scheme %q resolved a credential", auth)
		}
	}
	cred := &fakeCred{err: identity.ErrUnauthenticated}
	mux, _ := fixture(t, cred, nil)
	closedAnswer(t, "dead credential", serve(mux, "GET", status, nil, "", enr), 401, "unauthenticated")
	if cred.lastCred != strings.TrimPrefix(enr, "Enrollment ") || cred.lastOrig != hostA {
		t.Errorf("resolved %q at %q", cred.lastCred, cred.lastOrig)
	}
	mux, _ = fixture(t, &fakeCred{info: liveCred(orgB)}, nil)
	closedAnswer(t, "reassigned host", serve(mux, "GET", status, nil, "", enr), 401, "unauthenticated")
	mux, _ = fixture(t, &fakeCred{err: identity.ErrUnavailable}, nil)
	closedAnswer(t, "store failure", serve(mux, "GET", status, nil, "", enr), 503, "enrollment_unavailable")

	for _, path := range []string{"/internal/v1/enrollment/request", "/internal/v1/enrollment/admit", "/internal/v1/enrollment/abandon"} {
		b := mutationBody(path, hostA)
		b["csrf_token"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
		mux, db := fixture(t, &fakeCred{info: liveCred(orgA)}, nil)
		closedAnswer(t, path+" wrong csrf", serve(mux, "POST", path, jsonBody(t, b), "application/json", enr), 403, "csrf_failed")
		delete(b, "csrf_token")
		closedAnswer(t, path+" no csrf", serve(mux, "POST", path, jsonBody(t, b), "application/json", enr), 403, "csrf_failed")
		closedAnswer(t, path+" live", serve(mux, "POST", path, jsonBody(t, mutationBody(path, hostA)), "application/json", enr), 503, "enrollment_unavailable")
		if n := strings.Count(strings.Join(db.stmts, ","), "begin"); n != 1 {
			t.Errorf("%s: %d transactions, want the facade's one: %v", path, n, db.stmts)
		}
	}
	mux, db := fixture(t, &fakeCred{info: liveCred(orgA)}, nil)
	closedAnswer(t, "live status", serve(mux, "GET", status, nil, "", enr), 503, "enrollment_unavailable")
	if strings.Count(strings.Join(db.stmts, ","), "begin") != 1 {
		t.Errorf("status: %v", db.stmts)
	}
	// The admission body: kind open, or request with a canonical request id; anything else is invalid_request field
	// kind (or request_id) before any facade.
	for name, c := range map[string]struct {
		b     map[string]any
		field string
	}{
		"no kind":            {map[string]any{}, "kind"},
		"unknown kind":       {map[string]any{"kind": "invitation"}, "kind"},
		"open with id":       {map[string]any{"kind": "open", "request_id": rid}, "kind"},
		"request without id": {map[string]any{"kind": "request"}, "kind"},
		"malformed id":       {map[string]any{"kind": "request", "request_id": "x"}, "request_id"},
		"zero id":            {map[string]any{"kind": "request", "request_id": "00000000-0000-0000-0000-000000000000"}, "request_id"},
		"mode on admit":      {map[string]any{"kind": "open", "mode": "open"}, "kind"},
	} {
		mux, db := fixture(t, &fakeCred{info: liveCred(orgA)}, nil)
		b := map[string]any{"origin": hostA, "csrf_token": goodCSRF}
		for k, v := range c.b {
			b[k] = v
		}
		out := closedAnswer(t, "admit "+name, serve(mux, "POST", "/internal/v1/enrollment/admit", jsonBody(t, b), "application/json", enr), 400, "invalid_request")
		if out["field"] != c.field || strings.Contains(strings.Join(db.stmts, ","), "begin") {
			t.Errorf("admit %s: field %v, statements %v", name, out["field"], db.stmts)
		}
	}
	mux, db = fixture(t, &fakeCred{info: liveCred(orgA)}, nil)
	b := map[string]any{"origin": hostA, "csrf_token": goodCSRF, "kind": "request", "request_id": rid}
	closedAnswer(t, "admit request kind", serve(mux, "POST", "/internal/v1/enrollment/admit", jsonBody(t, b), "application/json", enr), 503, "enrollment_unavailable")
	for _, path := range []string{"/internal/v1/enrollment/request", "/internal/v1/enrollment/abandon"} {
		b := mutationBody(path, hostA)
		b["kind"] = "open"
		closedAnswer(t, path+" extra member", serve(mux, "POST", path, jsonBody(t, b), "application/json", enr), 400, "invalid_request")
	}
	_ = db
}

// TestEnrollmentRateLimited (Plan r0 §1.7): the configured rates are 5/s burst 10 per credential and 20/s burst 40 per
// resolved Organization (pinned); with refill frozen (rate 0, the configured bursts) the applicant routes allow the
// burst per credential before resolving it, then 429 rate_limited without a credential read; another credential has its
// own bucket; the public route allows its burst per Organization; a limiter never holds more than maxLimiters keys.
func TestEnrollmentRateLimited(t *testing.T) {
	if credentialRate != 5 || credentialBurst != 10 || publicRate != 20 || publicBurst != 40 || maxLimiters != 10000 {
		t.Fatal("limiter constants differ from Plan r0 §1.7")
	}
	cred := &fakeCred{err: identity.ErrUnauthenticated}
	mux, _ := fixtureWith(t, cred, nil, newLimiter(0, credentialBurst), newLimiter(0, publicBurst))
	status := "/internal/v1/enrollment/status" + q(hostA)
	for i := 0; i < credentialBurst; i++ {
		closedAnswer(t, fmt.Sprintf("burst %d", i), serve(mux, "GET", status, nil, "", enr), 401, "unauthenticated")
	}
	closedAnswer(t, "over burst", serve(mux, "GET", status, nil, "", enr), 429, "rate_limited")
	if cred.calls != credentialBurst {
		t.Errorf("credential reads %d, want %d (the limited request resolves nothing)", cred.calls, credentialBurst)
	}
	closedAnswer(t, "another credential", serve(mux, "GET", status, nil, "", "Enrollment "+strings.Repeat("F", 43)), 401, "unauthenticated")
	pub := "/internal/v1/enrollment/public" + q(hostA)
	for i := 0; i < publicBurst; i++ {
		if w := serve(mux, "GET", pub, nil, ""); w.Code != 200 {
			t.Fatalf("public burst %d: %d", i, w.Code)
		}
	}
	closedAnswer(t, "public over burst", serve(mux, "GET", pub, nil, ""), 429, "rate_limited")
	l := newLimiter(1, 1)
	for i := 0; i < maxLimiters+5; i++ {
		l.allow(fmt.Sprintf("k%d", i))
	}
	if len(l.m) > maxLimiters {
		t.Errorf("limiter holds %d keys", len(l.m))
	}
	if keyOf("abc") == "abc" || len(keyOf("abc")) != 64 {
		t.Error("the limiter key is not a digest")
	}
}

// TestEnrollmentLimiterChurn (Duo 4042726305; RC2 limiter follow-up): on a fake clock with the configured credential
// rate (5/s burst 10, full refill 2s) a depleted bucket survives churn to capacity; the map never exceeds maxLimiters; a
// new key at unsafe capacity is refused without replacing or adding an entry; after the clock advances only fully
// refilled buckets are reclaimed (depleted and partial ones are kept); and every scan, even a successful one, defers the
// next by burst/rate, so a new key inside that window is refused while an eligible bucket stays present (no scan ran).
func TestEnrollmentLimiterChurn(t *testing.T) {
	t0 := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	window := time.Duration(credentialBurst) * time.Second / credentialRate
	// setup returns a limiter on clock *at holding depleted key "A" and one-token-used churn keys up to capacity minus
	// spare, all at t0.
	setup := func(t *testing.T, at *time.Time, spare int) *limiter {
		t.Helper()
		*at = t0
		l := newLimiter(credentialRate, credentialBurst)
		l.now = func() time.Time { return *at }
		deplete(t, l, "A")
		for i := 0; len(l.m) < maxLimiters-spare; i++ {
			if !l.allow(fmt.Sprintf("c%d", i)) {
				t.Fatalf("churn key c%d refused below capacity", i)
			}
		}
		return l
	}
	t.Run("exhausted_key_preserved", func(t *testing.T) {
		var at time.Time
		l := setup(t, &at, 0)
		for i := 0; i < 50; i++ {
			l.allow(fmt.Sprintf("x%d", i))
		}
		if l.allow("A") {
			t.Fatal("the depleted key regained a token after churn to capacity")
		}
		if _, ok := l.m["A"]; !ok {
			t.Fatal("the depleted key's bucket was dropped")
		}
	})
	t.Run("bounded", func(t *testing.T) {
		var at time.Time
		l := setup(t, &at, 0)
		for i := 0; i < 2*maxLimiters; i++ {
			l.allow(fmt.Sprintf("x%d", i))
			if i%997 == 0 {
				at = at.Add(window / 3)
			}
			if len(l.m) > maxLimiters {
				t.Fatalf("limiter holds %d keys", len(l.m))
			}
		}
	})
	t.Run("unsafe_capacity_refused", func(t *testing.T) {
		var at time.Time
		l := setup(t, &at, 0)
		before := make(map[string]*rate.Limiter, len(l.m))
		for k, b := range l.m {
			before[k] = b
		}
		for _, k := range []string{"new1", "new2"} {
			if l.allow(k) {
				t.Errorf("new key %s admitted at capacity without a fully refilled bucket", k)
			}
		}
		if len(l.m) != len(before) {
			t.Fatalf("size %d, want %d", len(l.m), len(before))
		}
		for k, b := range before {
			if l.m[k] != b {
				t.Fatalf("entry %s replaced or removed", k)
			}
		}
	})
	t.Run("safe_reclaim_then_deferral", func(t *testing.T) {
		var at time.Time
		l := setup(t, &at, 1)
		at = t0.Add(window / 2)
		deplete(t, l, "D") // half refilled at t0+window; fully refilled at t0+3/2·window
		at = t0.Add(window)
		bA, bD := l.m["A"], l.m["D"]
		deplete(t, l, "A") // A refilled by t0+window; re-depleted now
		if !l.allow("c0") {
			t.Fatal("existing churn key c0 refused")
		}
		if !l.allow("n1") { // the first scan: frees the fully refilled churn keys
			t.Fatal("a new key was refused although fully refilled buckets were reclaimable")
		}
		want := map[string]bool{"A": true, "D": true, "c0": true, "n1": true}
		if len(l.m) != len(want) {
			t.Fatalf("after reclaim the limiter holds %d keys, want %d (A, D, c0, n1)", len(l.m), len(want))
		}
		for k := range want {
			if _, ok := l.m[k]; !ok {
				t.Fatalf("key %s missing after reclaim", k)
			}
		}
		if l.m["A"] != bA || l.m["D"] != bD {
			t.Fatal("a depleted or partial bucket was recreated")
		}
		if l.allow("A") {
			t.Fatal("the re-depleted key was reset by reclaim")
		}
		// Renewed pressure right after the successful scan: refill capacity at the same instant.
		for i := 0; len(l.m) < maxLimiters; i++ {
			if !l.allow(fmt.Sprintf("r%d", i)) {
				t.Fatalf("refill key r%d refused below capacity", i)
			}
		}
		// Inside the window D (and every r key) is fully refilled: a scan would delete it.
		at = t0.Add(window + window*3/4)
		if got := l.m["D"].TokensAt(at); got < credentialBurst {
			t.Fatalf("D holds %v tokens, want a fully refilled bucket", got)
		}
		if l.allow("n2") {
			t.Fatal("a new key was admitted inside the scan window")
		}
		if _, ok := l.m["D"]; !ok || len(l.m) != maxLimiters {
			t.Fatalf("a scan ran inside the window (D present %v, size %d)", ok, len(l.m))
		}
		if !l.allow("r0") {
			t.Fatal("an existing key with tokens was refused inside the window")
		}
		// At the window's end a scan runs again: every bucket is fully refilled except r1, re-depleted at this instant.
		at = t0.Add(2 * window)
		deplete(t, l, "r1")
		if !l.allow("n2") {
			t.Fatal("a new key was refused after the window although buckets were reclaimable")
		}
		if _, ok := l.m["D"]; ok {
			t.Fatal("the fully refilled bucket D was not reclaimed after the window")
		}
		if _, ok := l.m["r1"]; !ok || len(l.m) != 2 {
			t.Fatalf("after the second scan r1 present %v, size %d, want r1 and n2 only", ok, len(l.m))
		}
		if l.allow("r1") {
			t.Fatal("the re-depleted key was reset by the second scan")
		}
	})
}

// deplete takes key's whole burst on l at l's current instant and checks the next request is refused.
func deplete(t *testing.T, l *limiter, key string) {
	t.Helper()
	for i := 0; i < l.burst; i++ {
		if !l.allow(key) {
			t.Fatalf("%s refused at token %d of its burst", key, i)
		}
	}
	if l.allow(key) {
		t.Fatalf("%s admitted over its burst", key)
	}
}

// TestEnrollmentRefusalMapping (ENR-H refusal mapping): the facade codes map onto the closed set; the applicant's
// enrollment_closed is its own 403 without a field; the field code accompanies invalid_request, forbidden and conflict
// only (enrollment_restricted, approval_revoked, host_changed, not_pending, already_consumed, not_approved,
// enrollment_not_ready); an unauthenticated (expired or revoked credential), not_found or unavailable answer carries no
// detail.
func TestEnrollmentRefusalMapping(t *testing.T) {
	for code, want := range map[access.Code]string{
		access.CodeUnauthenticated: "unauthenticated", access.CodeContextDenied: "not_available", access.CodeForbidden: "forbidden",
		access.CodeNotFound: "not_found", access.CodeConflict: "conflict", access.CodeInvalidInput: "invalid_request",
		access.CodeUnavailable: "enrollment_unavailable", access.CodeInternal: "enrollment_unavailable",
	} {
		if got := failure(&access.Error{Code: code}); got != want {
			t.Errorf("%s -> %s, want %s", code, got, want)
		}
	}
	if got := failure(errors.New("other")); got != "enrollment_unavailable" {
		t.Errorf("a non-access error -> %s", got)
	}
	if len(status) != 14 {
		t.Errorf("%d refusal codes, want 14", len(status))
	}
	for _, c := range []struct {
		err   *access.Error
		code  int
		keys  string
		error string
		field string
	}{
		{&access.Error{Code: access.CodeConflict, Field: "enrollment_closed"}, 403, "error", "enrollment_closed", ""},
		{&access.Error{Code: access.CodeConflict, Field: "enrollment_restricted"}, 409, "error,field", "conflict", "enrollment_restricted"},
		{&access.Error{Code: access.CodeConflict, Field: "approval_revoked"}, 409, "error,field", "conflict", "approval_revoked"},
		{&access.Error{Code: access.CodeConflict, Field: "host_changed"}, 409, "error,field", "conflict", "host_changed"},
		{&access.Error{Code: access.CodeConflict, Field: "not_pending"}, 409, "error,field", "conflict", "not_pending"},
		{&access.Error{Code: access.CodeConflict, Field: "already_consumed"}, 409, "error,field", "conflict", "already_consumed"},
		{&access.Error{Code: access.CodeConflict, Field: "enrollment_not_ready"}, 409, "error,field", "conflict", "enrollment_not_ready"},
		{&access.Error{Code: access.CodeForbidden, Field: "x"}, 403, "error,field", "forbidden", "x"},
		{&access.Error{Code: access.CodeInvalidInput, Field: "status"}, 400, "error,field", "invalid_request", "status"},
		{&access.Error{Code: access.CodeUnauthenticated, Field: "enrollment_expired"}, 401, "error", "unauthenticated", ""},
		{&access.Error{Code: access.CodeNotFound, Field: "x"}, 404, "error", "not_found", ""},
		{&access.Error{Code: access.CodeUnavailable, Field: "commit_outcome_unknown"}, 503, "error", "enrollment_unavailable", ""},
	} {
		w := httptest.NewRecorder()
		(&Handler{}).refuseErr(w, c.err)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		keys := []string{}
		for k := range out {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if w.Code != c.code || strings.Join(keys, ",") != c.keys || out["error"] != c.error || (c.field != "" && out["field"] != c.field) {
			t.Errorf("%v/%s: %d %v", c.err.Code, c.err.Field, w.Code, out)
		}
	}
}

// TestEnrollmentAnswerShapes (ENR-17 minimal disclosure, ENR-H answers): the applicant's own request view has no subject,
// decider, revoker, binding or authority column; the manager's view adds only the applicant id and a null display name;
// the policy answer is exactly mode, open_ready and history_basis (null while unresolved).
func TestEnrollmentAnswerShapes(t *testing.T) {
	var v access.EnrollmentRequestView
	if !parseID(&v.ID, rid) || !parseID(&v.UserID, uid) {
		t.Fatal("ids")
	}
	v.Status, v.CreatedAt, v.ExpiresAt = "approved", time.Unix(0, 0), time.Unix(86400, 0)
	decided := time.Unix(60, 0)
	v.DecidedAt = &decided
	keys := func(m map[string]any) string {
		var out []string
		for k := range m {
			out = append(out, k)
		}
		sort.Strings(out)
		return strings.Join(out, ",")
	}
	own := ownRequestJSON(v)
	if keys(own) != "consumed,created_at,decided_at,expired,expires_at,id,revoked_at,status" || own["id"] != rid ||
		own["decided_at"] != "1970-01-01T00:01:00Z" || own["revoked_at"] != nil {
		t.Errorf("applicant view %v", own)
	}
	b, _ := json.Marshal(own)
	if strings.Contains(string(b), uid) || len(b) > 600 {
		t.Errorf("applicant view discloses the subject or exceeds its budget: %s", b)
	}
	mgr := requestJSON(v)
	if keys(mgr) != "consumed,created_at,decided_at,display_name,expired,expires_at,id,revoked_at,status,user_id" ||
		mgr["user_id"] != uid || mgr["display_name"] != nil {
		t.Errorf("manager view %v", mgr)
	}
	if p := policyJSON("invite_only", false, ""); keys(p) != "history_basis,mode,open_ready" || p["history_basis"] != nil || p["open_ready"] != false {
		t.Errorf("policy %v", p)
	}
	if p := policyJSON("open", true, "created_after_ledger"); p["history_basis"] != "created_after_ledger" {
		t.Errorf("policy basis %v", p)
	}
	// ENR-17d: the applicant's status envelope is exactly status, request, member and csrf_token; it names neither the
	// subject nor the Organization (no user_id, organization_id, origin or address).
	mux, db := fixture(t, &fakeCred{info: liveCred(orgA), member: false}, nil)
	db.tx = statusTx{}
	w := serve(mux, "GET", "/internal/v1/enrollment/status"+q(hostA), nil, "", enr)
	var st map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &st); err != nil || w.Code != 200 {
		t.Fatalf("status: %d %q %v", w.Code, w.Body.String(), err)
	}
	if keys(st) != "csrf_token,member,request,status" || st["status"] != "none" || st["request"] != nil || st["member"] != false ||
		st["csrf_token"] != goodCSRF {
		t.Errorf("status envelope %v", st)
	}
	for _, leak := range []string{uid, orgA, hostA, "org-a", "user_id", "organization_id", "origin", "address"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Errorf("status envelope discloses %q: %s", leak, w.Body.String())
		}
	}
}

// TestEnrollmentEventIsClosed: one enrollment.request per request with the operation and closed reason only; never the
// credential, the origin, a token, a user or a request id.
func TestEnrollmentEventIsClosed(t *testing.T) {
	var buf bytes.Buffer
	mux, _ := fixture(t, &fakeCred{err: identity.ErrUnauthenticated}, slog.New(slog.NewJSONHandler(&buf, nil)))
	for path, op := range map[string]string{"/internal/v1/enrollment/status": "status", "/internal/v1/enrollment/public": "public",
		"/internal/v1/enrollment/requests": "list_requests", "/internal/v1/enrollment/policy": "get_policy"} {
		buf.Reset()
		serve(mux, "GET", path+q(hostA), nil, "", enr)
		var rec map[string]any
		if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
			t.Fatalf("%s: event %q: %v", path, buf.String(), err)
		}
		if rec["msg"] != "enrollment.request" || rec["operation"] != op || rec["host_tag"] == nil {
			t.Errorf("%s: event %v", path, rec)
		}
		for _, s := range []string{strings.TrimPrefix(enr, "Enrollment "), hostA, "org-a.localhost", orgA, uid} {
			if strings.Contains(buf.String(), s) {
				t.Errorf("%s: event carries %q", path, s)
			}
		}
	}
}

// TestEnrollmentRoutesPinned: the Route constants are exactly the reviewed pairs (the CORE_ENROLLMENT_ROUTES mechanism
// reads the same lines), and every pair's path is a served path.
func TestEnrollmentRoutesPinned(t *testing.T) {
	src, err := os.ReadFile("http.go")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, m := range regexp.MustCompile(`(?m)^\s+Route[A-Za-z]+\s+= "([^"]+)"$`).FindAllSubmatch(src, -1) {
		got = append(got, string(m[1]))
	}
	sort.Strings(got)
	want := "GET /internal/v1/enrollment/policy|GET /internal/v1/enrollment/public|GET /internal/v1/enrollment/requests|" +
		"GET /internal/v1/enrollment/status|POST /internal/v1/enrollment/abandon|POST /internal/v1/enrollment/admit|" +
		"POST /internal/v1/enrollment/policy|POST /internal/v1/enrollment/request|POST /internal/v1/enrollment/requests/{id}/decide|" +
		"POST /internal/v1/enrollment/requests/{id}/revoke"
	if strings.Join(got, "|") != want {
		t.Errorf("routes:\n got %s\nwant %s", strings.Join(got, "|"), want)
	}
	paths := map[string]bool{}
	for _, p := range Paths {
		paths[p] = true
	}
	for _, r := range got {
		if _, p, _ := strings.Cut(r, " "); !paths[p] {
			t.Errorf("%s: path not served", r)
		}
	}
}

// decl returns the exact source bytes of the named function or single type declaration (without its doc comment).
func decl(t *testing.T, path, name string) string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		var n ast.Node
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Name.Name == name {
				n = d
			}
		case *ast.GenDecl:
			if d.Tok == token.TYPE && len(d.Specs) == 1 && d.Specs[0].(*ast.TypeSpec).Name.Name == name {
				n = d
			}
		}
		if n != nil {
			return string(src[fset.Position(n.Pos()).Offset:fset.Position(n.End()).Offset])
		}
	}
	t.Fatalf("%s: %s not declared", path, name)
	return ""
}

// TestEnrollmentCopyIsExact: the session-scope behaviour is a bounded copy of onboarding/api, not a shared helper; this
// pins begin, scope, originHost, query, hostTag and writeJSON byte for byte.
func TestEnrollmentCopyIsExact(t *testing.T) {
	for _, name := range []string{"begin", "scope", "originHost", "query", "hostTag", "writeJSON"} {
		if a, b := decl(t, "../../onboarding/api/http.go", name), decl(t, "http.go", name); a != b {
			t.Errorf("%s differs from onboarding/api:\n--- onboarding\n%s\n--- enrollment\n%s", name, a, b)
		}
	}
}

func TestEnrollmentNewRefusesIncompleteConfiguration(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("an incomplete configuration was accepted")
	}
	var nilH *Handler
	if err := nilH.Bind(); err == nil {
		t.Fatal("a nil handler bound an admitter")
	}
	if err := (&Handler{}).Bind(); err == nil {
		t.Fatal("an unconfigured handler bound an admitter")
	}
}
