package api

import (
	"bytes"
	"context"
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

	"github.com/jackc/pgx/v5"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/tenancy/access"
)

// G019 (Plan r1 §5.3, r2 addendum §7; §6.2 MEM-H): the membership adapter's closed surface without a database — the
// sixteen pairs on fifteen paths and their method/path closure, the literal segments matched before {id}, the
// pre-session refusals (ambiguous parameters, rejected origins, no session), strict bodies and the budget, the refusal
// mapping with its field codes, the id helpers, and the bounded copy of onboarding/api. The authenticated routes run
// against real PostgreSQL through the facade (TestMembership) and the browser suites.

const (
	orgA  = "0f19a000-0000-4000-8000-0000000000a1"
	hostA = "https://org-a.localhost:543"
	uid   = "0f19a000-0000-4000-8000-00000000c001"
	gid   = "0f19a000-0000-4000-8000-00000000b001"
	iid   = "0f19a000-0000-4000-8000-00000000d001"
)

// fakeDB answers only the resolver's hostname_state statement; anything else is a failure of the test.
type fakeDB struct{ stmts []string }

type row struct{ scan func(...any) error }

func (r row) Scan(dest ...any) error { return r.scan(dest...) }

func (f *fakeDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("fakeDB: no transaction expected")
}

func (f *fakeDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	if !strings.Contains(sql, "hostname_state") {
		f.stmts = append(f.stmts, "unexpected")
		return row{func(...any) error { return errors.New("fakeDB: unexpected statement") }}
	}
	f.stmts = append(f.stmts, "hostname_state")
	if args[0].(string)+"|"+args[1].(string) != "platform|org-a" {
		return row{func(...any) error { return pgx.ErrNoRows }}
	}
	return row{func(dest ...any) error {
		org := orgA
		*dest[0].(**string), *dest[1].(*bool), *dest[2].(*bool) = &org, true, false
		return nil
	}}
}

func fixture(t *testing.T) (*http.ServeMux, *fakeDB) {
	t.Helper()
	db := &fakeDB{}
	d := config.Domains{BaseDomain: "localhost", CNAMETarget: "edge.localhost", HostPort: "543", AuthHost: "auth.localhost"}
	h := &Handler{cfg: Config{Resolver: domains.NewResolver(db, d.Hostnames())}, acc: access.NewService(db, nil, nil),
		log: slog.New(slog.DiscardHandler)}
	mux := http.NewServeMux()
	h.Register(mux)
	return mux, db
}

func serve(mux *http.ServeMux, method, target string, body []byte, ct string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, bytes.NewReader(body))
	if ct != "" {
		r.Header.Set("Content-Type", ct)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}

// closedAnswer asserts a closed JSON refusal with the no-store headers and no CORS header.
func closedAnswer(t *testing.T, name string, w *httptest.ResponseRecorder, status int, code string) {
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

const eligibility = "/internal/v1/account/retirement-eligibility"

// The sixteen pairs (the fifteen paths), each mutation with a body that passes the decoder.
var pairs = []struct{ method, path string }{
	{"GET", "/internal/v1/members"}, {"GET", "/internal/v1/members/self"},
	{"POST", "/internal/v1/members/" + uid + "/role"}, {"POST", "/internal/v1/members/" + uid + "/remove"},
	{"POST", "/internal/v1/members/" + uid + "/promote"}, {"POST", "/internal/v1/members/" + uid + "/demote"},
	{"POST", "/internal/v1/members/leave"}, {"POST", "/internal/v1/ownership/transfer"},
	{"GET", "/internal/v1/invitations"}, {"POST", "/internal/v1/invitations"},
	{"GET", "/internal/v1/invitations/mine"}, {"GET", "/internal/v1/invitations/candidates"},
	{"GET", "/internal/v1/invitations/" + iid}, {"POST", "/internal/v1/invitations/" + iid + "/revoke"},
	{"POST", "/internal/v1/invitations/" + iid + "/accept"}, {"GET", eligibility},
	// G060: the Group-management override read and owner-only write.
	{"GET", groupAuthority}, {"POST", groupAuthority},
}

const groupAuthority = "/internal/v1/organization/group-authority"

func mutationBody(path, origin string) map[string]any {
	b := map[string]any{"origin": origin, "csrf_token": "x"}
	switch {
	case strings.HasSuffix(path, "/role"), strings.HasSuffix(path, "/demote"):
		b["scope"], b["role"] = "organization", "org_member"
	case strings.HasSuffix(path, "/transfer"):
		b["scope"], b["recipient_user_id"], b["former_role"] = "organization", uid, "org_admin"
	case path == "/internal/v1/invitations":
		b["scope"], b["group_id"], b["role"], b["intended_user_id"] = "group", gid, "group_member", uid
	case strings.HasSuffix(path, "/accept"):
		b["token"] = "t"
	case strings.HasSuffix(path, "/revoke"):
	case path == groupAuthority:
		b["enabled"] = true
	default:
		b["scope"] = "organization"
	}
	return b
}

func TestMembershipHTTP(t *testing.T) {
	mux, db := fixture(t)
	if len(Paths) != 16 || len(pairs) != 18 {
		t.Fatalf("%d paths and %d pairs, want 16 and 18", len(Paths), len(pairs))
	}
	for _, p := range pairs {
		name := p.method + " " + p.path
		if p.method == "GET" {
			closedAnswer(t, name+" no session", serve(mux, "GET", p.path+q(hostA), nil, ""), 401, "unauthenticated")
			if p.path != eligibility {
				// The eligibility route has no resolver decision (IDENTITY shape): a well-formed foreign origin is refused
				// by the session binding, after the credential.
				closedAnswer(t, name+" foreign origin", serve(mux, "GET", p.path+q("https://elsewhere.test:543"), nil, ""), 403, "origin_rejected")
			}
			closedAnswer(t, name+" http origin", serve(mux, "GET", p.path+q("http://org-a.localhost:543"), nil, ""), 403, "origin_rejected")
			closedAnswer(t, name+" no origin", serve(mux, "GET", p.path, nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" two origins", serve(mux, "GET", p.path+q(hostA)+"&origin=x", nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" unknown parameter", serve(mux, "GET", p.path+q(hostA)+"&x=1", nil, ""), 400, "ambiguous_parameter")
			r := httptest.NewRequest("GET", p.path+q(hostA), nil)
			r.Header.Add("Authorization", "Session a")
			r.Header.Add("Authorization", "Session b")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			closedAnswer(t, name+" two credentials", w, 400, "ambiguous_parameter")
			continue
		}
		ok := jsonBody(t, mutationBody(p.path, hostA))
		closedAnswer(t, name+" no session", serve(mux, p.method, p.path, ok, "application/json"), 401, "unauthenticated")
		closedAnswer(t, name+" foreign origin", serve(mux, p.method, p.path, jsonBody(t, mutationBody(p.path, "https://elsewhere.test:543")), "application/json"), 403, "origin_rejected")
		closedAnswer(t, name+" not JSON", serve(mux, p.method, p.path, ok, "text/plain"), 400, "invalid_request")
		closedAnswer(t, name+" query", serve(mux, p.method, p.path+q(hostA), ok, "application/json"), 400, "ambiguous_parameter")
		unknown := mutationBody(p.path, hostA)
		unknown["organization_id"] = orgA // a forged tenant is an unknown field
		closedAnswer(t, name+" unknown field", serve(mux, p.method, p.path, jsonBody(t, unknown), "application/json"), 400, "invalid_request")
		for _, tail := range []string{" {}", "]", "}", "\n}"} {
			closedAnswer(t, name+fmt.Sprintf(" trailing %q", tail), serve(mux, p.method, p.path, append(append([]byte{}, ok...), tail...), "application/json"), 400, "invalid_request")
		}
		closedAnswer(t, name+" trailing whitespace", serve(mux, p.method, p.path, append(append([]byte{}, ok...), " \r\n\t"...), "application/json"), 401, "unauthenticated")
		closedAnswer(t, name+" malformed", serve(mux, p.method, p.path, []byte("{"), "application/json"), 400, "invalid_request")
		big := mutationBody(p.path, hostA)
		big["csrf_token"] = strings.Repeat("x", maxBody)
		closedAnswer(t, name+" over budget", serve(mux, p.method, p.path, jsonBody(t, big), "application/json"), 400, "invalid_request")
	}

	// Every method outside each path's closed set is method_not_allowed (closed JSON, not the mux's text answer); the
	// literal segments' other methods reach the /invitations/{id} closure.
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

	// Sub-path probing below the four roots: not_found (closed JSON); an unparsable id is not_found after the session.
	for _, p := range []string{"/internal/v1/members/" + uid, "/internal/v1/members/" + uid + "/x", "/internal/v1/members/self/x",
		"/internal/v1/invitations/" + iid + "/x", "/internal/v1/invitations/mine/x", "/internal/v1/ownership/",
		"/internal/v1/ownership/x", "/internal/v1/account/", "/internal/v1/account/retire", "/internal/v1/members/",
		"/internal/v1/invitations/" + iid + "/consume"} {
		closedAnswer(t, "sub-path "+p, serve(mux, "GET", p+q(hostA), nil, ""), 404, "not_found")
	}
	for _, s := range db.stmts {
		if s != "hostname_state" {
			t.Fatalf("a statement other than the resolver's ran: %v", db.stmts)
		}
	}
}

// TestMembershipLiteralBeforeID: mine and candidates reach their own handlers, never GET /invitations/{id} with the
// literal as an id (the event names the operation).
func TestMembershipLiteralBeforeID(t *testing.T) {
	var buf bytes.Buffer
	db := &fakeDB{}
	d := config.Domains{BaseDomain: "localhost", CNAMETarget: "edge.localhost", HostPort: "543", AuthHost: "auth.localhost"}
	h := &Handler{cfg: Config{Resolver: domains.NewResolver(db, d.Hostnames())}, acc: access.NewService(db, nil, nil),
		log: slog.New(slog.NewJSONHandler(&buf, nil))}
	mux2 := http.NewServeMux()
	h.Register(mux2)
	for path, op := range map[string]string{"/internal/v1/invitations/mine": "my_invitations",
		"/internal/v1/invitations/candidates": "invitation_candidates", "/internal/v1/invitations/" + iid: "get_invitation",
		"/internal/v1/members/self": "self_member"} {
		buf.Reset()
		serve(mux2, "GET", path+q(hostA), nil, "")
		var rec map[string]any
		if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
			t.Fatalf("%s: event %q: %v", path, buf.String(), err)
		}
		if rec["msg"] != "membership.request" || rec["operation"] != op || rec["reason"] != "unauthenticated" {
			t.Errorf("%s: event %v, want operation %s", path, rec, op)
		}
		for _, k := range []string{"origin", "token", "intended_address", "user_id"} {
			if _, ok := rec[k]; ok {
				t.Errorf("%s: event carries %s", path, k)
			}
		}
	}
}

func TestMembershipRefusalMapping(t *testing.T) {
	for code, want := range map[access.Code]string{
		access.CodeUnauthenticated: "unauthenticated", access.CodeContextDenied: "not_available", access.CodeForbidden: "forbidden",
		access.CodeNotFound: "not_found", access.CodeConflict: "conflict", access.CodeInvalidInput: "invalid_request",
		access.CodeUnavailable: "membership_unavailable", access.CodeInternal: "membership_unavailable",
	} {
		if got := failure(&access.Error{Code: code}); got != want {
			t.Errorf("%s -> %s, want %s", code, got, want)
		}
	}
	if got := failure(errors.New("other")); got != "membership_unavailable" {
		t.Errorf("a non-access error -> %s", got)
	}
	if len(status) != 11 {
		t.Errorf("%d refusal codes, want 11", len(status))
	}
	// The field code accompanies invalid_request, forbidden and conflict only; nothing else carries detail.
	for _, c := range []struct {
		err  *access.Error
		keys string
	}{
		{&access.Error{Code: access.CodeConflict, Field: "last_owner"}, "error,field"},
		{&access.Error{Code: access.CodeForbidden, Field: "owner_subject"}, "error,field"},
		{&access.Error{Code: access.CodeInvalidInput, Field: "after"}, "error,field"},
		{&access.Error{Code: access.CodeNotFound, Field: "x"}, "error"},
		{&access.Error{Code: access.CodeUnavailable, Field: "store_unavailable"}, "error"},
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
		if strings.Join(keys, ",") != c.keys {
			t.Errorf("%v: %v", c.err.Code, out)
		}
	}
}

// TestMembershipIDHelpers: the canonical parser refuses the literal segments, the zero id and non-canonical forms, and
// formatID is the facade's lowercase form.
func TestMembershipIDHelpers(t *testing.T) {
	var m access.Member
	for _, bad := range []string{"mine", "candidates", "self", "leave", "", "00000000-0000-0000-0000-000000000000",
		"{" + uid + "}", strings.ReplaceAll(uid, "-", "")} {
		if parseID(&m.UserID, bad) {
			t.Errorf("%q parsed", bad)
		}
	}
	if !parseID(&m.UserID, strings.ToUpper(uid)) || formatID(m.UserID) != uid {
		t.Errorf("round trip: %s", formatID(m.UserID))
	}
	if _, ok := scopeOf("group", ptr("x")); ok {
		t.Error("a malformed Group id built a scope")
	}
	if sc, ok := scopeOf("group", ptr(gid)); !ok || string(sc.Kind) != "group" || sc.Group == nil || formatID(*sc.Group) != gid {
		t.Errorf("group scope: %v %v", sc, ok)
	}
}

func ptr(s string) *string { return &s }

// TestMembershipBodyShapes: each route's member set is closed (required present, nothing else).
func TestMembershipBodyShapes(t *testing.T) {
	role, scope, tok := ptr("org_member"), ptr("organization"), ptr("t")
	cases := []struct {
		in       body
		req, opt []string
		want     bool
	}{
		{body{Scope: scope, Role: role}, []string{"scope", "role"}, []string{"group_id"}, true},
		{body{Scope: scope}, []string{"scope", "role"}, []string{"group_id"}, false},
		{body{Scope: scope, Role: role, Token: tok}, []string{"scope", "role"}, []string{"group_id"}, false},
		{body{Token: tok}, []string{"token"}, nil, true},
		{body{}, nil, nil, true},
		{body{Scope: scope}, nil, nil, false},
		// G060: the override write is exactly enabled; a scope, role or extra member is refused.
		{body{Enabled: new(bool)}, []string{"enabled"}, nil, true},
		{body{}, []string{"enabled"}, nil, false},
		{body{Enabled: new(bool), Scope: scope}, []string{"enabled"}, nil, false},
	}
	for i, c := range cases {
		if got := c.in.shape(c.req, c.opt); got != c.want {
			t.Errorf("case %d: %v, want %v", i, got, c.want)
		}
	}
}

// TestMembershipRoutesPinned: the Route constants are exactly the eighteen reviewed pairs (G060 adds the Group-authority
// read and write; the CORE_MEMBERSHIP_ROUTES mechanism reads the same lines), and every pair's path is a served path.
func TestMembershipRoutesPinned(t *testing.T) {
	src, err := os.ReadFile("http.go")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, m := range regexp.MustCompile(`(?m)^\s+Route[A-Za-z]+\s+= "([^"]+)"$`).FindAllSubmatch(src, -1) {
		got = append(got, string(m[1]))
	}
	sort.Strings(got)
	want := "GET /internal/v1/account/retirement-eligibility|GET /internal/v1/invitations|GET /internal/v1/invitations/candidates|" +
		"GET /internal/v1/invitations/mine|GET /internal/v1/invitations/{id}|GET /internal/v1/members|GET /internal/v1/members/self|" +
		"GET /internal/v1/organization/group-authority|" +
		"POST /internal/v1/invitations|POST /internal/v1/invitations/{id}/accept|POST /internal/v1/invitations/{id}/revoke|" +
		"POST /internal/v1/members/leave|POST /internal/v1/members/{user_id}/demote|POST /internal/v1/members/{user_id}/promote|" +
		"POST /internal/v1/members/{user_id}/remove|POST /internal/v1/members/{user_id}/role|POST /internal/v1/organization/group-authority|" +
		"POST /internal/v1/ownership/transfer"
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

// TestMembershipCopyIsExact: the session-scope behaviour is a bounded copy of onboarding/api, not a shared helper; this
// pins begin, scope, originHost, query, hostTag and writeJSON byte for byte.
func TestMembershipCopyIsExact(t *testing.T) {
	for _, name := range []string{"begin", "scope", "originHost", "query", "hostTag", "writeJSON"} {
		if a, b := decl(t, "../../onboarding/api/http.go", name), decl(t, "http.go", name); a != b {
			t.Errorf("%s differs from onboarding/api:\n--- onboarding\n%s\n--- membership\n%s", name, a, b)
		}
	}
}

func TestMembershipNewRefusesIncompleteConfiguration(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("an incomplete configuration was accepted")
	}
}
