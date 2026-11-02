package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/tenancy/access"
)

// G074 (Plan r3.1 §3.7; B-UNIT-CORE, TestProfileAPI*): the profile API's routes, its refusals before the session scope
// (no body byte read and no statement but the resolver's), the closed refusal mapping including the avatar HOLD (503 with
// its reason and Retry-After), the avatar header set, the JSON shapes, the strict parsers and the bounded copy of the
// branding/api session scope. The authenticated routes run against real PostgreSQL and MinIO in TestProfile and
// TestProfileAvatar (the facade) and the browser suites.

const (
	orgA  = "0f74a000-0000-4000-8000-0000000000a1"
	userA = "0f74a000-0000-4000-8000-00000000c0de"
	hostA = "https://org-a.localhost:543"
)

// G075: two Groups of A (any canonical ids: no case here reaches the facade's transactions).
const (
	groupR = "0f75a000-0000-4000-8000-0000000000b1"
	groupZ = "0f75a000-0000-4000-8000-0000000000b2"
)

// fakeDB is the Service's DB: no transaction (no route here reaches the facade's transactions) and the resolver's
// hostname_state statement only.
type fakeDB struct {
	hosts map[string]*string
	stmts []string
}

type row struct {
	scan func(dest ...any) error
}

func (r row) Scan(dest ...any) error { return r.scan(dest...) }

func (f *fakeDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	f.stmts = append(f.stmts, "begin")
	return nil, errors.New("fakeDB: no transaction expected")
}

func (f *fakeDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	if !strings.Contains(sql, "hostname_state") {
		f.stmts = append(f.stmts, "unexpected")
		return row{func(...any) error { return errors.New("fakeDB: unexpected statement") }}
	}
	f.stmts = append(f.stmts, "hostname_state")
	org, ok := f.hosts[args[0].(string)+"|"+args[1].(string)]
	if !ok {
		return row{func(...any) error { return pgx.ErrNoRows }}
	}
	return row{func(dest ...any) error {
		*dest[0].(**string), *dest[1].(*bool), *dest[2].(*bool) = org, true, false
		return nil
	}}
}

func strp(s string) *string { return &s }

// fixture: org-a.localhost routes to A; nothing else routes.
func fixture(t *testing.T) (*Handler, *fakeDB) {
	t.Helper()
	db := &fakeDB{hosts: map[string]*string{"platform|org-a": strp(orgA)}}
	d := config.Domains{BaseDomain: "localhost", CNAMETarget: "edge.localhost", HostPort: "543", AuthHost: "auth.localhost"}
	h := &Handler{cfg: Config{Resolver: domains.NewResolver(db, d.Hostnames())}, acc: access.NewService(db, nil, nil),
		log: slog.New(slog.DiscardHandler)}
	return h, db
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("body %q: %v", w.Body.String(), err)
	}
	return out
}

func keys(m map[string]any) string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return strings.Join(ks, ",")
}

// noStoreJSON asserts the headers of every answer except avatar bytes, and that no CORS header is ever set.
func noStoreJSON(t *testing.T, name string, w *httptest.ResponseRecorder) {
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
}

// countingReader records how many body bytes a handler asked for.
type countingReader struct{ read int64 }

func (c *countingReader) Read(p []byte) (int, error) {
	c.read += int64(len(p))
	return 0, io.EOF
}

func TestProfileAPIRoutes(t *testing.T) {
	h, _ := fixture(t)
	mux := http.NewServeMux()
	h.Register(mux)
	for _, c := range []struct{ method, path, want string }{
		{http.MethodGet, "/internal/v1/profiles/me", RouteOwn},
		{http.MethodPost, "/internal/v1/profiles/me", RouteUpdateName},
		{http.MethodPost, "/internal/v1/profiles/me/avatar", RouteSetAvatar},
		{http.MethodPost, "/internal/v1/profiles/me/avatar/remove", RouteRemoveAvatar},
		{http.MethodGet, "/internal/v1/profiles/" + userA, RouteGet},
		{http.MethodGet, "/internal/v1/profiles/" + userA + "/avatar", RouteAvatar},
		{http.MethodGet, "/internal/v1/profiles/" + userA + "/groups", RouteGroups},
		{http.MethodPost, "/internal/v1/profiles/" + userA + "/groups/" + groupR + "/label", RouteGroupLabel},
	} {
		if _, pattern := mux.Handler(httptest.NewRequest(c.method, c.path, nil)); pattern != c.want {
			t.Errorf("%s %s routes to %q, want %q", c.method, c.path, pattern, c.want)
		}
	}
	// No route issues a signed URL or serves a POST on another user's profile.
	for _, c := range []struct{ method, path string }{
		{http.MethodPost, "/internal/v1/profiles/" + userA},
		{http.MethodPost, "/internal/v1/profiles/" + userA + "/avatar"},
		{http.MethodGet, "/internal/v1/profiles/" + userA + "/avatar/url"},
	} {
		if _, pattern := mux.Handler(httptest.NewRequest(c.method, c.path, nil)); pattern == RouteSetAvatar || pattern == RouteUpdateName ||
			strings.HasSuffix(pattern, "/url") {
			t.Errorf("%s %s routes to %q", c.method, c.path, pattern)
		}
	}
	// G075: the cards read is GET only and the label update POST only, each on exactly its own path.
	for _, c := range []struct{ method, path string }{
		{http.MethodPost, "/internal/v1/profiles/" + userA + "/groups"},
		{http.MethodGet, "/internal/v1/profiles/" + userA + "/groups/" + groupR + "/label"},
		{http.MethodGet, "/internal/v1/profiles/" + userA + "/groups/" + groupR},
		{http.MethodPost, "/internal/v1/profiles/" + userA + "/groups/" + groupR},
		{http.MethodPost, "/internal/v1/profiles/" + userA + "/groups/" + groupR + "/label/x"},
	} {
		if _, pattern := mux.Handler(httptest.NewRequest(c.method, c.path, nil)); pattern == RouteGroups || pattern == RouteGroupLabel {
			t.Errorf("%s %s routes to %q", c.method, c.path, pattern)
		}
	}
	// The route set is exactly the eight reviewed pairs (G074 six, G075 two); infra/validation/run.sh
	// CORE_PROFILE_ROUTES pins the same constants, sorted and joined the same way.
	var routes []string
	for n, v := range parseSource(t, "http.go").consts() {
		if strings.HasPrefix(n, "Route") {
			routes = append(routes, strings.Trim(v, `"`))
		}
	}
	sort.Strings(routes)
	if got := strings.Join(routes, "|"); got != "GET /internal/v1/profiles/me|GET /internal/v1/profiles/{user_id}|"+
		"GET /internal/v1/profiles/{user_id}/avatar|GET /internal/v1/profiles/{user_id}/groups|POST /internal/v1/profiles/me|"+
		"POST /internal/v1/profiles/me/avatar|POST /internal/v1/profiles/me/avatar/remove|"+
		"POST /internal/v1/profiles/{user_id}/groups/{group_id}/label" {
		t.Errorf("route set %s", got)
	}
}

func TestProfileAPIPreSessionRefusals(t *testing.T) {
	h, db := fixture(t)
	get := func(hf http.HandlerFunc, target, path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, target, nil)
		if path != "" {
			r.SetPathValue("user_id", path)
		}
		w := httptest.NewRecorder()
		hf(w, r)
		return w
	}
	q := func(p string, kv ...string) string {
		v := url.Values{}
		for i := 0; i+1 < len(kv); i += 2 {
			v.Add(kv[i], kv[i+1])
		}
		return p + "?" + v.Encode()
	}
	for _, c := range []struct {
		name string
		w    *httptest.ResponseRecorder
		code int
		want string
	}{
		{"own without origin", get(h.Own, "/internal/v1/profiles/me", ""), 400, "ambiguous_parameter"},
		{"own with two origins", get(h.Own, q("/internal/v1/profiles/me", "origin", hostA, "origin", hostA), ""), 400, "ambiguous_parameter"},
		{"own over http", get(h.Own, q("/internal/v1/profiles/me", "origin", "http://org-a.localhost:543"), ""), 403, "origin_rejected"},
		{"own at a host that routes nowhere", get(h.Own, q("/internal/v1/profiles/me", "origin", "https://nobody.localhost:543"), ""), 403, "origin_rejected"},
		{"own without a credential", get(h.Own, q("/internal/v1/profiles/me", "origin", hostA), ""), 401, "unauthenticated"},
		{"profile without a credential", get(h.Get, q("/internal/v1/profiles/"+userA, "origin", hostA), userA), 401, "unauthenticated"},
		{"avatar without v", get(h.Avatar, q("/internal/v1/profiles/"+userA+"/avatar", "origin", hostA), userA), 400, "ambiguous_parameter"},
		{"avatar with two v", get(h.Avatar, q("/internal/v1/profiles/"+userA+"/avatar", "origin", hostA, "v", "1", "v", "2"), userA), 400, "ambiguous_parameter"},
		{"avatar without a credential", get(h.Avatar, q("/internal/v1/profiles/"+userA+"/avatar", "origin", hostA, "v", "1"), userA), 401, "unauthenticated"},
	} {
		noStoreJSON(t, c.name, c.w)
		if out := decode(t, c.w); c.w.Code != c.code || out["error"] != c.want || keys(out) != "error" {
			t.Errorf("%s: %d %v, want %d %s", c.name, c.w.Code, out, c.code, c.want)
		}
	}

	post := func(hf http.HandlerFunc, ct, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/internal/v1/profiles/me", strings.NewReader(body))
		if ct != "" {
			r.Header.Set("Content-Type", ct)
		}
		w := httptest.NewRecorder()
		hf(w, r)
		return w
	}
	update, remove := h.mutation("update_name", h.updateName), h.mutation("remove_avatar", h.removeAvatar)
	for _, c := range []struct {
		name string
		w    *httptest.ResponseRecorder
		code int
		want string
	}{
		{"form body", post(update, "application/x-www-form-urlencoded", "name=x"), 415, "unsupported_media_type"},
		{"no content type", post(update, "", "{}"), 415, "unsupported_media_type"},
		{"malformed JSON", post(update, "application/json", "{"), 400, "invalid_request"},
		{"no expected version", post(update, "application/json", `{"origin":"`+hostA+`","name":"x"}`), 400, "invalid_request"},
		{"negative expected version", post(update, "application/json", `{"origin":"`+hostA+`","expected_version":-1,"name":"x"}`), 400, "invalid_request"},
		{"expected version over int32", post(update, "application/json", `{"origin":"`+hostA+`","expected_version":2147483648,"name":"x"}`), 400, "invalid_request"},
		{"unknown field", post(update, "application/json", `{"origin":"`+hostA+`","expected_version":0,"email":"x"}`), 400, "invalid_request"},
		{"trailing data", post(remove, "application/json", `{"origin":"`+hostA+`","expected_version":0} {}`), 400, "invalid_request"},
		{"no credential", post(update, "application/json", `{"origin":"`+hostA+`","expected_version":0,"name":"x"}`), 401, "unauthenticated"},
	} {
		noStoreJSON(t, c.name, c.w)
		if out := decode(t, c.w); c.w.Code != c.code || out["error"] != c.want {
			t.Errorf("%s: %d %v, want %d %s", c.name, c.w.Code, out, c.code, c.want)
		}
	}

	avatar := func(origin, csrf, version string, length int64, body io.Reader) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/internal/v1/profiles/me/avatar", body)
		for k, v := range map[string]string{"X-Origin": origin, "X-CSRF-Token": csrf, "X-Expected-Version": version} {
			if v != "" {
				r.Header.Set(k, v)
			}
		}
		r.Header.Set("Authorization", "Session x")
		r.ContentLength = length
		w := httptest.NewRecorder()
		h.SetAvatar(w, r)
		return w
	}
	for _, c := range []struct {
		name                  string
		origin, csrf, version string
		length                int64
		code                  int
		want                  string
	}{
		{"413 over the bound", hostA, "c", "0", 524289, 413, "payload_too_large"},
		{"413 far over the bound", hostA, "c", "0", 1 << 30, 413, "payload_too_large"},
		{"unknown length", hostA, "c", "0", -1, 400, "invalid_request"},
		{"bad version", hostA, "c", "x", 10, 400, "invalid_request"},
		{"negative version", hostA, "c", "-1", 10, 400, "invalid_request"},
		{"version over int32", hostA, "c", "2147483648", 10, 400, "invalid_request"},
		{"missing origin", "", "c", "0", 10, 400, "ambiguous_parameter"},
		{"missing csrf", hostA, "", "0", 10, 400, "ambiguous_parameter"},
		{"missing version", hostA, "c", "", 10, 400, "ambiguous_parameter"},
		// At the bound the request proceeds to the session scope: an origin that routes nowhere is rejected there.
		{"at the bound", "https://nobody.localhost:543", "c", "0", 524288, 403, "origin_rejected"},
	} {
		body := &countingReader{}
		w := avatar(c.origin, c.csrf, c.version, c.length, body)
		noStoreJSON(t, c.name, w)
		if out := decode(t, w); w.Code != c.code || out["error"] != c.want {
			t.Errorf("%s: %d %v, want %d %s", c.name, w.Code, out, c.code, c.want)
		}
		if body.read != 0 {
			t.Errorf("%s: %d body bytes read", c.name, body.read)
		}
	}
	for _, s := range db.stmts {
		if s != "hostname_state" {
			t.Errorf("statement %s before the session scope", s)
		}
	}
}

func TestProfileAPIRefusalMapping(t *testing.T) {
	for code, want := range map[access.Code]string{
		access.CodeUnauthenticated: "unauthenticated", access.CodeContextDenied: "not_available", access.CodeForbidden: "forbidden",
		access.CodeNotFound: "not_found", access.CodeConflict: "version_conflict", access.CodeUnavailable: "unavailable",
		access.CodeInternal: "unavailable",
	} {
		if got := failure(&access.Error{Code: code}); got != want {
			t.Errorf("%s -> %s, want %s", code, got, want)
		}
	}
	for field, want := range map[string]string{"avatar_too_large": "payload_too_large", "avatar_type": "invalid_input",
		"avatar_dimensions": "invalid_input", "avatar_malformed": "invalid_input", "normalized_too_large": "invalid_input",
		"avatar_size": "invalid_request", "name": "invalid_request", "expected_version": "invalid_request"} {
		if got := failure(&access.Error{Code: access.CodeInvalidInput, Field: field}); got != want {
			t.Errorf("invalid_input/%s -> %s, want %s", field, got, want)
		}
	}
	if got := failure(errors.New("other")); got != "unavailable" {
		t.Errorf("a non-access error -> %s", got)
	}
	if len(status) != 14 {
		t.Errorf("%d refusal codes, want 14", len(status))
	}

	refuse := func(err error) (*httptest.ResponseRecorder, map[string]any) {
		w := httptest.NewRecorder()
		(&Handler{}).refuseErr(w, err)
		return w, decode(t, w)
	}
	// The three HOLD reasons: 503 unavailable with the reason and Retry-After (30 s without a facade delay).
	for _, reason := range []string{"avatar_cleanup_pending", "avatar_upload_in_progress", "commit_outcome_unknown"} {
		w, out := refuse(&access.Error{Code: access.CodeUnavailable, Field: reason})
		noStoreJSON(t, reason, w)
		if w.Code != 503 || keys(out) != "error,reason" || out["error"] != "unavailable" || out["reason"] != reason ||
			w.Header().Get("Retry-After") != "30" {
			t.Errorf("%s: %d %v Retry-After %q", reason, w.Code, out, w.Header().Get("Retry-After"))
		}
	}
	// Any other unavailable outcome carries no reason and no Retry-After.
	for _, err := range []error{&access.Error{Code: access.CodeUnavailable}, &access.Error{Code: access.CodeUnavailable, Field: "compensated"},
		&access.Error{Code: access.CodeInternal, Field: "avatar_cleanup_pending"}} {
		w, out := refuse(err)
		if w.Code != 503 || keys(out) != "error" || w.Header().Get("Retry-After") != "" {
			t.Errorf("%v: %d %v Retry-After %q", err, w.Code, out, w.Header().Get("Retry-After"))
		}
	}
	// invalid_input and invalid_request carry the field; 413 too; not_found is exactly {"error":"not_found"}.
	for _, c := range []struct {
		err  error
		code int
		keys string
	}{
		{&access.Error{Code: access.CodeInvalidInput, Field: "avatar_type"}, 400, "error,field"},
		{&access.Error{Code: access.CodeInvalidInput, Field: "name"}, 400, "error,field"},
		{&access.Error{Code: access.CodeInvalidInput, Field: "avatar_too_large"}, 413, "error,field"},
		{&access.Error{Code: access.CodeConflict, Field: "version_conflict"}, 409, "error"},
		{&access.Error{Code: access.CodeNotFound}, 404, "error"},
	} {
		w, out := refuse(c.err)
		noStoreJSON(t, "refusal", w)
		if w.Code != c.code || keys(out) != c.keys {
			t.Errorf("%v: %d %v, want %d %s", c.err, w.Code, out, c.code, c.keys)
		}
	}
}

func TestProfileAPIAvatarHeaders(t *testing.T) {
	data := []byte("\x89PNG\r\n\x1a\nnot really")
	w := httptest.NewRecorder()
	writeAvatar(w, access.AccountAvatar{Version: 3, ContentType: "image/png"}, data)
	want := map[string]string{"Content-Type": "image/png", "Content-Length": "18", "Cache-Control": "private, max-age=300",
		"Etag": `"v3"`, "X-Content-Type-Options": "nosniff", "Content-Disposition": "inline"}
	if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), data) || len(w.Header()) != len(want) {
		t.Fatalf("avatar answer: %d %v", w.Code, w.Header())
	}
	for k, v := range want {
		if got := w.Header().Get(k); got != v {
			t.Errorf("%s: %q, want %q", k, got, v)
		}
	}
}

func TestProfileAPIJSONShapes(t *testing.T) {
	var p access.AccountProfile
	if !parseID(&p.UserID, userA) {
		t.Fatal("parse")
	}
	own := ownJSON(access.AccountProfileView{UserID: p.UserID, EffectiveName: "Provider Name", Version: 0})
	if keys(own) != "avatar,effective_name,name,user_id,version" || own["name"] != nil || own["avatar"] != nil || own["user_id"] != userA {
		t.Errorf("own view without a row: %v", own)
	}
	own = ownJSON(access.AccountProfileView{UserID: p.UserID, Name: strp("Explicit"), EffectiveName: "Explicit",
		Avatar: &access.AccountAvatar{Version: 2, ContentType: "image/jpeg"}, Version: 4})
	if a, ok := own["avatar"].(map[string]any); !ok || keys(a) != "content_type,version" || own["name"] != "Explicit" {
		t.Errorf("own view: %v", own)
	}
	other := otherJSON(access.AccountProfile{UserID: p.UserID, Name: "Effective", Avatar: &access.AccountAvatar{Version: 1, ContentType: "image/png"}})
	if keys(other) != "avatar,name,user_id" || other["name"] != "Effective" {
		t.Errorf("other view: %v", other)
	}
}

func TestProfileAPIParsers(t *testing.T) {
	var p access.AccountProfile
	for _, s := range []string{"", "x", "00000000-0000-0000-0000-000000000000", strings.ToUpper(userA), userA + " ", "{" + userA + "}"} {
		if parseID(&p.UserID, s) {
			t.Errorf("parseID accepted %q", s)
		}
	}
	if !parseID(&p.UserID, userA) || formatID(p.UserID) != userA {
		t.Error("parseID/formatID round trip")
	}
	for s, want := range map[string]int32{"1": 1, "42": 42, "2147483647": 2147483647} {
		if v, ok := parseVersion(s); !ok || v != want {
			t.Errorf("parseVersion(%q) = %d %t", s, v, ok)
		}
	}
	for _, s := range []string{"", "0", "-1", "+1", "01", "1.0", "2147483648", " 1"} {
		if _, ok := parseVersion(s); ok {
			t.Errorf("parseVersion accepted %q", s)
		}
	}
}

// parsed is one source file of the copy check.
type parsed struct {
	fset *token.FileSet
	file *ast.File
	src  []byte
}

func parseSource(t *testing.T, path string) parsed {
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
	return parsed{fset, f, src}
}

func (p parsed) consts() map[string]string {
	out := map[string]string{}
	for _, d := range p.file.Decls {
		if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.CONST {
			for _, s := range g.Specs {
				vs := s.(*ast.ValueSpec)
				for i, n := range vs.Names {
					if i < len(vs.Values) {
						if lit, ok := vs.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							out[n.Name] = lit.Value
						}
					}
				}
			}
		}
	}
	return out
}

func (p parsed) statuses(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, d := range p.file.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.VAR {
			continue
		}
		for _, s := range g.Specs {
			vs := s.(*ast.ValueSpec)
			if len(vs.Names) == 1 && vs.Names[0].Name == "status" {
				for _, e := range vs.Values[0].(*ast.CompositeLit).Elts {
					kv := e.(*ast.KeyValueExpr)
					out[kv.Key.(*ast.Ident).Name] = kv.Value.(*ast.BasicLit).Value
				}
			}
		}
	}
	if len(out) == 0 {
		t.Fatal("no status table")
	}
	return out
}

// decl returns the exact source bytes of the named function or type declaration (without its doc comment).
func (p parsed) decl(t *testing.T, name string) string {
	t.Helper()
	for _, d := range p.file.Decls {
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
			return string(p.src[p.fset.Position(n.Pos()).Offset:p.fset.Position(n.End()).Offset])
		}
	}
	t.Fatalf("%s not declared", name)
	return ""
}

// The session-scope behaviour is a bounded copy of branding/api (itself the domains/api copy), not a shared helper.
func TestProfileAPISessionCopyIsExact(t *testing.T) {
	brd, prf := parseSource(t, "../../branding/api/http.go"), parseSource(t, "http.go")
	for _, name := range []string{"begin", "scope", "originHost", "writeJSON", "fail", "refuse"} {
		if a, b := brd.decl(t, name), prf.decl(t, name); a != b {
			t.Errorf("%s differs from branding/api:\n--- branding\n%s\n--- profile\n%s", name, a, b)
		}
	}
	shared := []string{"errInvalidRequest", "errAmbiguous", "errUnsupported", "errOriginRejected", "errUnauthenticated", "errCSRF",
		"errNotAvailable", "errForbidden", "errNotFound", "errUnknownHost", "errConflict", "errTooLarge"}
	bc, pc := brd.consts(), prf.consts()
	bs, ps := brd.statuses(t), prf.statuses(t)
	for _, n := range shared {
		if bc[n] == "" || bc[n] != pc[n] || bs[n] == "" || bs[n] != ps[n] {
			t.Errorf("%s: branding %s/%s, profile %s/%s", n, bc[n], bs[n], pc[n], ps[n])
		}
	}
	if ps["errUnavailable"] != "503" || pc["errUnavailable"] != `"unavailable"` {
		t.Errorf("errUnavailable: %s %s", pc["errUnavailable"], ps["errUnavailable"])
	}
	for n, v := range pc {
		if strings.HasPrefix(n, "err") && ps[n] == "" {
			t.Errorf("%s (%s) has no status", n, v)
		}
	}
	if len(ps) != len(status) {
		t.Errorf("parsed %d statuses, the table has %d", len(ps), len(status))
	}
}

func TestProfileAPINewRefusesIncompleteConfiguration(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("an incomplete configuration was accepted")
	}
}

// G075 (Plan r2 §3.4, §3.5, §3.9; B-UNIT-CORE, TestProfileAPI*): the two Group member card routes and the selected-Group
// context selector (C10). Every case here stops before the facade's transactions; the authenticated card and label
// outcomes run against real PostgreSQL in TestGroupMemberProfile and RLS-GMP, and through the browser in E-GMP-*.

// groupsGet calls the cards read with the raw query (set verbatim, so that an unparsable query reaches the handler).
func groupsGet(h *Handler, rawQuery string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/internal/v1/profiles/"+userA+"/groups", nil)
	r.URL.RawQuery = rawQuery
	r.SetPathValue("user_id", userA)
	w := httptest.NewRecorder()
	h.Groups(w, r)
	return w
}

// labelPost calls the label update of userA's label in groupR.
func labelPost(h *Handler, ct string, body io.Reader) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/internal/v1/profiles/"+userA+"/groups/"+groupR+"/label", body)
	if ct != "" {
		r.Header.Set("Content-Type", ct)
	}
	r.SetPathValue("user_id", userA)
	r.SetPathValue("group_id", groupR)
	w := httptest.NewRecorder()
	h.SetGroupLabel(w, r)
	return w
}

// labelJSON is a well-formed label update body with extra members appended.
func labelJSON(extra string) string {
	return `{"origin":"` + hostA + `","csrf_token":"c","expected_version":0,"role_label":"Lead"` + extra + `}`
}

// The selector grammar: absent is the Organization-only request; a present value is exactly one canonical lowercase
// Group UUID, returned byte-exact; an empty, malformed or repeated value is refused, never read as absent.
func TestProfileAPIContextSelectorGrammar(t *testing.T) {
	if g, ok := contextSelector(nil, false); !ok || g != "" {
		t.Errorf("absent: %q %t", g, ok)
	}
	if g, ok := contextSelector([]string{groupR}, true); !ok || g != groupR {
		t.Errorf("valid: %q %t", g, ok)
	}
	for _, c := range []struct {
		name string
		vals []string
	}{
		{"empty", []string{""}},
		{"no value", []string{}},
		{"repeated", []string{groupR, groupR}},
		{"two Groups", []string{groupR, groupZ}},
		{"repeated with an empty value", []string{groupR, ""}},
		{"not a UUID", []string{"not-a-uuid"}},
		{"upper case", []string{strings.ToUpper(groupR)}},
		{"braced", []string{"{" + groupR + "}"}},
		{"leading space", []string{" " + groupR}},
		{"trailing space", []string{groupR + " "}},
		{"no hyphens", []string{strings.ReplaceAll(groupR, "-", "")}},
		{"urn", []string{"urn:uuid:" + groupR}},
		{"zero UUID", []string{"00000000-0000-0000-0000-000000000000"}},
	} {
		if g, ok := contextSelector(c.vals, true); ok || g != "" {
			t.Errorf("%s: accepted as %q", c.name, g)
		}
	}
	if g, ok := bodySelector(nil); !ok || g != "" {
		t.Errorf("absent member: %q %t", g, ok)
	}
	if g, ok := bodySelector(json.RawMessage(`"` + groupR + `"`)); !ok || g != groupR {
		t.Errorf("valid member: %q %t", g, ok)
	}
	for _, raw := range []string{`null`, `""`, `1`, `true`, `{}`, `["` + groupR + `"]`, `"not-a-uuid"`} {
		if g, ok := bodySelector(json.RawMessage(raw)); ok || g != "" {
			t.Errorf("member %s: accepted as %q", raw, g)
		}
	}
}

// A present selector that is empty, malformed or repeated is invalid_request context_group on both routes, before the
// session scope and before any statement (not even the resolver's), and its value is never echoed; a valid selector and
// an absent one proceed to the session scope.
func TestProfileAPIContextSelectorRefusals(t *testing.T) {
	h, db := fixture(t)
	q := func(kv ...string) string {
		v := url.Values{}
		for i := 0; i+1 < len(kv); i += 2 {
			v.Add(kv[i], kv[i+1])
		}
		return v.Encode()
	}
	refused := func(name string, w *httptest.ResponseRecorder, value string) {
		t.Helper()
		noStoreJSON(t, name, w)
		out := decode(t, w)
		if w.Code != 400 || keys(out) != "error,field" || out["error"] != "invalid_request" || out["field"] != "context_group" {
			t.Errorf("%s: %d %v, want 400 invalid_request context_group", name, w.Code, out)
		}
		if value != "" && strings.Contains(w.Body.String(), value) {
			t.Errorf("%s: the selector is echoed", name)
		}
	}
	for _, c := range []struct {
		name  string
		query string
		value string
	}{
		{"empty", q("origin", hostA, "context_group", ""), ""},
		{"present without a value", "origin=" + url.QueryEscape(hostA) + "&context_group", ""},
		{"malformed", q("origin", hostA, "context_group", "not-a-uuid"), "not-a-uuid"},
		{"upper case", q("origin", hostA, "context_group", strings.ToUpper(groupR)), strings.ToUpper(groupR)},
		{"zero UUID", q("origin", hostA, "context_group", "00000000-0000-0000-0000-000000000000"), ""},
		{"repeated", q("origin", hostA, "context_group", groupR, "context_group", groupR), ""},
		{"two Groups", q("origin", hostA, "context_group", groupR, "context_group", groupZ), ""},
		{"repeated with an empty value", q("origin", hostA, "context_group", groupR, "context_group", ""), ""},
	} {
		refused("GET "+c.name, groupsGet(h, c.query), c.value)
	}
	for _, c := range []struct {
		name  string
		extra string
		value string
	}{
		{"empty", `,"context_group":""`, ""},
		{"null", `,"context_group":null`, ""},
		{"number", `,"context_group":1`, ""},
		{"object", `,"context_group":{}`, ""},
		{"array", `,"context_group":["` + groupR + `"]`, ""},
		{"malformed", `,"context_group":"not-a-uuid"`, "not-a-uuid"},
		{"upper case", `,"context_group":"` + strings.ToUpper(groupR) + `"`, strings.ToUpper(groupR)},
		{"leading space", `,"context_group":" ` + groupR + `"`, ""},
		{"zero UUID", `,"context_group":"00000000-0000-0000-0000-000000000000"`, ""},
		{"repeated", `,"context_group":"` + groupR + `","context_group":"` + groupR + `"`, ""},
		{"a Group then null", `,"context_group":"` + groupR + `","context_group":null`, ""},
		{"null then a Group", `,"context_group":null,"context_group":"` + groupR + `"`, ""},
		{"repeated in another case", `,"context_group":"` + groupR + `","Context_Group":"` + groupZ + `"`, ""},
	} {
		refused("POST "+c.name, labelPost(h, "application/json", strings.NewReader(labelJSON(c.extra))), c.value)
	}
	if len(db.stmts) != 0 {
		t.Errorf("statements before a selector refusal: %v", db.stmts)
	}
	// A valid selector and an absent one pass the grammar and stop at the session scope (no credential here).
	for _, c := range []struct {
		name string
		w    *httptest.ResponseRecorder
	}{
		{"GET valid", groupsGet(h, q("origin", hostA, "context_group", groupR))},
		{"GET absent", groupsGet(h, q("origin", hostA))},
		{"POST valid", labelPost(h, "application/json", strings.NewReader(labelJSON(`,"context_group":"`+groupR+`"`)))},
		{"POST absent", labelPost(h, "application/json", strings.NewReader(labelJSON("")))},
	} {
		noStoreJSON(t, c.name, c.w)
		if out := decode(t, c.w); c.w.Code != 401 || keys(out) != "error" || out["error"] != "unauthenticated" {
			t.Errorf("%s: %d %v, want 401 unauthenticated", c.name, c.w.Code, out)
		}
	}
	if len(db.stmts) != 4 {
		t.Errorf("%d statements, want the resolver's 4", len(db.stmts))
	}
	for _, s := range db.stmts {
		if s != "hostname_state" {
			t.Errorf("statement %s before the session scope", s)
		}
	}
}

// fakeResolver records the context requests narrow makes and answers with tc or err.
type fakeResolver struct {
	reqs []access.ResolveRequest
	tc   *access.TenantContext
	err  error
}

func (f *fakeResolver) Resolve(_ context.Context, req access.ResolveRequest) (*access.TenantContext, error) {
	f.reqs = append(f.reqs, req)
	if f.err != nil {
		return nil, f.err
	}
	return f.tc, nil
}

// The selector transport (C10): absent keeps begin's Organization-only context and resolves nothing; present, Resolve
// receives exactly the scope's Organization and the selector byte-exact and decides: a selectable Group's context is
// used; an unknown, foreign or unselectable one (context_denied) is not_found with exactly the body of an invisible
// subject, and of the facade's selected-other denial (a target Group other than the selected one), never the selector.
func TestProfileAPIContextSelectorTransport(t *testing.T) {
	ctx := context.Background()
	base, selected := &access.TenantContext{}, &access.TenantContext{}
	f := &fakeResolver{tc: selected}
	if tc, code := narrow(ctx, f, base, orgA, ""); tc != base || code != "" || len(f.reqs) != 0 {
		t.Errorf("absent: %p %q %v", tc, code, f.reqs)
	}
	if tc, code := narrow(ctx, f, base, orgA, groupR); tc != selected || code != "" || len(f.reqs) != 1 ||
		f.reqs[0] != (access.ResolveRequest{OrganizationID: orgA, GroupID: groupR}) {
		t.Errorf("selectable: %p %q %v", tc, code, f.reqs)
	}
	for _, c := range []struct {
		err  error
		want string
	}{
		{&access.Error{Code: access.CodeContextDenied}, "not_found"},
		{&access.Error{Code: access.CodeInvalidInput, Field: "invalid_id"}, "invalid_request"},
		{&access.Error{Code: access.CodeUnauthenticated}, "unauthenticated"},
		{&access.Error{Code: access.CodeUnavailable}, "unavailable"},
		{errors.New("other"), "unavailable"},
	} {
		f := &fakeResolver{err: c.err}
		if tc, code := narrow(ctx, f, base, orgA, groupZ); tc != nil || code != c.want || len(f.reqs) != 1 || f.reqs[0].GroupID != groupZ {
			t.Errorf("%v: %p %q %v, want %s", c.err, tc, code, f.reqs, c.want)
		}
	}

	// One answer for an unselectable selector and for the facade's not_found (an invisible subject; a selected-other
	// target Group, Plan r2 §3.4): status, headers and body byte for byte.
	unselectable, invisible := httptest.NewRecorder(), httptest.NewRecorder()
	(&Handler{}).failSelector(unselectable, errNotFound)
	(&Handler{}).refuseErr(invisible, &access.Error{Code: access.CodeNotFound})
	noStoreJSON(t, "unselectable selector", unselectable)
	if invisible.Code != unselectable.Code || invisible.Body.String() != unselectable.Body.String() ||
		len(invisible.Header()) != len(unselectable.Header()) {
		t.Errorf("not_found: %d %q, unselectable selector: %d %q", invisible.Code, invisible.Body.String(), unselectable.Code,
			unselectable.Body.String())
	}
	if unselectable.Code != 404 || unselectable.Body.String() != "{\"error\":\"not_found\"}\n" {
		t.Errorf("unselectable selector: %d %q", unselectable.Code, unselectable.Body.String())
	}
	// A selector refusal names the field, never the value.
	invalid := httptest.NewRecorder()
	(&Handler{}).failSelector(invalid, errInvalidRequest)
	if out := decode(t, invalid); invalid.Code != 400 || keys(out) != "error,field" || out["field"] != "context_group" {
		t.Errorf("invalid selector: %d %v", invalid.Code, out)
	}
}

// The selector is read only by the two Group member card routes: begin (the byte-exact branding/api copy) resolves the
// Organization only, the six G074 handlers never read the selector, the selector reaches only narrow, and the label
// update passes the path's target Group, never the selector, to the facade.
func TestProfileAPIContextSelectorScope(t *testing.T) {
	prf := parseSource(t, "http.go")
	if b := prf.decl(t, "begin"); !strings.Contains(b, "access.ResolveRequest{OrganizationID: org})") || strings.Contains(b, "GroupID") {
		t.Errorf("begin resolves more than the Organization:\n%s", b)
	}
	for _, name := range []string{"Own", "mutation", "updateName", "removeAvatar", "SetAvatar", "Get", "Avatar"} {
		d := prf.decl(t, name)
		for _, s := range []string{"narrow(", "Selector(", "fieldContextGroup", "context_group", "ResolveRequest", "ContextGroup"} {
			if strings.Contains(d, s) {
				t.Errorf("%s reads the selector (%s)", name, s)
			}
		}
	}
	if d := prf.decl(t, "narrow"); strings.Count(d, "access.ResolveRequest{OrganizationID: org, GroupID: group}") != 1 {
		t.Errorf("narrow does not pass the selector as the Group selector:\n%s", d)
	}
	g := prf.decl(t, "Groups")
	if strings.Count(g, "narrow(") != 1 || strings.Count(g, "contextSelector(vals, present)") != 1 {
		t.Errorf("Groups does not resolve its selector exactly once:\n%s", g)
	}
	l := prf.decl(t, "SetGroupLabel")
	if strings.Count(l, "narrow(") != 1 || strings.Count(l, "in.ContextGroup") != 1 || strings.Count(l, "bodySelector(in.ContextGroup)") != 1 ||
		strings.Count(l, `parseID(&g, r.PathValue("group_id"))`) != 1 || strings.Count(l, "h.acc.SetGroupMemberLabel(s.ctx, tc, g, p.UserID,") != 1 {
		t.Errorf("SetGroupLabel does not keep the target Group and the selector apart:\n%s", l)
	}
}

// The two routes' bounds before the session scope: the cards read's strict query (an unparsable query or an unknown
// parameter is refused, so that no selector is dropped silently) and origin multiplicity; the label update's media type,
// its 8 KiB body bound (no more than the bound plus one byte read), strict JSON (unknown, repeated and trailing members)
// and expected version range. The G074 mutation body accepts neither new member. Nothing but the resolver's statement
// runs.
func TestProfileAPIGroupRoutesPreSession(t *testing.T) {
	h, db := fixture(t)
	esc := url.QueryEscape
	for _, c := range []struct {
		name  string
		query string
		code  int
		want  string
	}{
		{"no origin", "", 400, "ambiguous_parameter"},
		{"two origins", "origin=" + esc(hostA) + "&origin=" + esc(hostA), 400, "ambiguous_parameter"},
		{"unknown parameter", "origin=" + esc(hostA) + "&v=1", 400, "invalid_request"},
		{"the selector in another case", "origin=" + esc(hostA) + "&Context_Group=" + groupR, 400, "invalid_request"},
		{"the page's parameter name", "origin=" + esc(hostA) + "&context=" + groupR, 400, "invalid_request"},
		{"unparsable escape", "origin=" + esc(hostA) + "&context_group=%zz", 400, "invalid_request"},
		{"semicolon separator", "origin=" + esc(hostA) + ";context_group=" + groupR, 400, "invalid_request"},
		{"over http", "origin=" + esc("http://org-a.localhost:543"), 403, "origin_rejected"},
		{"at a host that routes nowhere", "origin=" + esc("https://nobody.localhost:543"), 403, "origin_rejected"},
		{"no credential", "origin=" + esc(hostA), 401, "unauthenticated"},
	} {
		w := groupsGet(h, c.query)
		noStoreJSON(t, "GET "+c.name, w)
		if out := decode(t, w); w.Code != c.code || keys(out) != "error" || out["error"] != c.want {
			t.Errorf("GET %s: %d %v, want %d %s", c.name, w.Code, out, c.code, c.want)
		}
	}

	version := func(v string) string {
		return `{"origin":"` + hostA + `","csrf_token":"c","expected_version":` + v + `,"role_label":"x"}`
	}
	for _, c := range []struct {
		name string
		ct   string
		body string
		code int
		want string
	}{
		{"form body", "application/x-www-form-urlencoded", "role_label=x", 415, "unsupported_media_type"},
		{"no content type", "", labelJSON(""), 415, "unsupported_media_type"},
		{"malformed JSON", "application/json", "{", 400, "invalid_request"},
		{"not an object", "application/json", "[]", 400, "invalid_request"},
		{"no expected version", "application/json", `{"origin":"` + hostA + `","csrf_token":"c","role_label":"x"}`, 400, "invalid_request"},
		{"negative expected version", "application/json", version("-1"), 400, "invalid_request"},
		{"expected version over int32", "application/json", version("2147483648"), 400, "invalid_request"},
		{"fractional expected version", "application/json", version("1.5"), 400, "invalid_request"},
		{"unknown member name", "application/json", labelJSON(`,"name":"x"`), 400, "invalid_request"},
		{"unknown member user_id", "application/json", labelJSON(`,"user_id":"` + userA + `"`), 400, "invalid_request"},
		{"unknown member group_id", "application/json", labelJSON(`,"group_id":"` + groupZ + `"`), 400, "invalid_request"},
		{"repeated expected version", "application/json", labelJSON(`,"expected_version":1`), 400, "invalid_request"},
		{"repeated label", "application/json", labelJSON(`,"role_label":"Other"`), 400, "invalid_request"},
		{"repeated label in another case", "application/json", labelJSON(`,"Role_Label":"Other"`), 400, "invalid_request"},
		{"trailing data", "application/json", labelJSON("") + " {}", 400, "invalid_request"},
		{"no credential", "application/json", labelJSON(""), 401, "unauthenticated"},
	} {
		w := labelPost(h, c.ct, strings.NewReader(c.body))
		noStoreJSON(t, "POST "+c.name, w)
		if out := decode(t, w); w.Code != c.code || keys(out) != "error" || out["error"] != c.want {
			t.Errorf("POST %s: %d %v, want %d %s", c.name, w.Code, out, c.code, c.want)
		}
	}

	// The body bound: over 8 KiB is invalid_request after at most the bound plus one byte; at the bound the request
	// proceeds to the session scope (the label's own bound is the facade's).
	prefix, suffix := `{"origin":"`+hostA+`","csrf_token":"c","expected_version":0,"role_label":"`, `"}`
	for _, c := range []struct {
		name string
		size int
		code int
		want string
	}{
		{"at the bound", maxBody, 401, "unauthenticated"},
		{"one byte over the bound", maxBody + 1, 400, "invalid_request"},
		{"far over the bound", 1 << 20, 400, "invalid_request"},
	} {
		body := []byte(prefix + strings.Repeat("x", c.size-len(prefix)-len(suffix)) + suffix)
		rd := bytes.NewReader(body)
		w := labelPost(h, "application/json", rd)
		noStoreJSON(t, c.name, w)
		if out := decode(t, w); w.Code != c.code || out["error"] != c.want {
			t.Errorf("%s: %d %v, want %d %s", c.name, w.Code, out, c.code, c.want)
		}
		if read := int64(len(body)) - int64(rd.Len()); read > maxBody+1 {
			t.Errorf("%s: %d body bytes read", c.name, read)
		}
	}

	// The G074 name update accepts neither the selector nor a role label.
	update := h.mutation("update_name", h.updateName)
	for _, extra := range []string{`,"context_group":"` + groupR + `"`, `,"role_label":"x"`} {
		r := httptest.NewRequest(http.MethodPost, "/internal/v1/profiles/me",
			strings.NewReader(`{"origin":"`+hostA+`","expected_version":0,"name":"x"`+extra+`}`))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		update(w, r)
		if out := decode(t, w); w.Code != 400 || out["error"] != "invalid_request" {
			t.Errorf("update_name with %s: %d %v", extra, w.Code, out)
		}
	}
	for _, s := range db.stmts {
		if s != "hostname_state" {
			t.Errorf("statement %s before the session scope", s)
		}
	}
}

// The cards answer (C8): exactly user_id, cards and narrow_required; each card exactly its five fields, with a null label
// when there is none; an empty list is [] (never null); an authorized overflow is no card and narrow_required, never the
// overflowing card, a partial list, a total or a cursor. The label update answers exactly the card.
func TestProfileAPIGroupCardShapes(t *testing.T) {
	var c access.GroupMemberCards
	var g access.GroupID
	if !parseID(&c.Subject, userA) || !parseID(&g, groupR) {
		t.Fatal("parse")
	}
	answer := func(v map[string]any) (*httptest.ResponseRecorder, map[string]any) {
		w := httptest.NewRecorder()
		writeJSON(w, http.StatusOK, v)
		return w, decode(t, w)
	}
	w, out := answer(cardsJSON(c))
	noStoreJSON(t, "no cards", w)
	if w.Body.String() != `{"cards":[],"narrow_required":false,"user_id":"`+userA+`"}`+"\n" {
		t.Errorf("no cards: %s", w.Body.String())
	}

	c.Cards = []access.GroupMemberCard{
		{GroupID: g, GroupName: "Research", RoleLabel: strp("Lead"), CanEdit: true, Version: 3},
		{GroupID: g, GroupName: "Design"},
	}
	w, out = answer(cardsJSON(c))
	cards, _ := out["cards"].([]any)
	if keys(out) != "cards,narrow_required,user_id" || len(cards) != 2 || out["narrow_required"] != false || out["user_id"] != userA {
		t.Fatalf("two cards: %s", w.Body.String())
	}
	first, _ := cards[0].(map[string]any)
	second, _ := cards[1].(map[string]any)
	if keys(first) != "can_edit,group_id,group_name,role_label,version" || first["group_id"] != groupR || first["group_name"] != "Research" ||
		first["role_label"] != "Lead" || first["can_edit"] != true || first["version"] != float64(3) {
		t.Errorf("first card: %v", first)
	}
	if keys(second) != keys(first) || second["role_label"] != nil || second["can_edit"] != false || second["version"] != float64(0) {
		t.Errorf("second card: %v", second)
	}

	over := access.GroupMemberCards{Subject: c.Subject, NarrowRequired: true}
	for i := 0; i < 201; i++ {
		id := g
		id[15] = byte(i)
		over.Cards = append(over.Cards, access.GroupMemberCard{GroupID: id, GroupName: "Overflow", CanEdit: true, Version: 1})
	}
	w, _ = answer(cardsJSON(over))
	if w.Code != 200 || w.Body.String() != `{"cards":[],"narrow_required":true,"user_id":"`+userA+`"}`+"\n" {
		t.Errorf("overflow: %s", w.Body.String())
	}

	w, out = answer(cardJSON(access.GroupMemberCard{GroupID: g, GroupName: "Research", CanEdit: true, Version: 1}))
	if keys(out) != "can_edit,group_id,group_name,role_label,version" || out["role_label"] != nil || out["version"] != float64(1) {
		t.Errorf("label answer: %s", w.Body.String())
	}
}

// The two routes' profile.request records carry the operation, the closed outcome and the reason only: never a label, a
// selector value, a Group or a subject.
func TestProfileAPIGroupRecords(t *testing.T) {
	h, _ := fixture(t)
	var buf bytes.Buffer
	h.log = slog.New(slog.NewJSONHandler(&buf, nil))
	secret := "Secret-Role-Label"
	labelPost(h, "application/json", strings.NewReader(`{"origin":"`+hostA+`","csrf_token":"c","expected_version":0,"role_label":"`+secret+
		`","context_group":"`+strings.ToUpper(groupZ)+`"}`))
	groupsGet(h, "origin="+url.QueryEscape(hostA)+"&context_group="+strings.ToUpper(groupZ))
	labelPost(h, "application/json", strings.NewReader(labelJSON(`,"context_group":"`+groupZ+`"`)))
	got := buf.String()
	for _, want := range []string{`"operation":"update_group_label"`, `"operation":"get_groups"`, `"outcome":"refused"`,
		`"reason":"invalid_request"`, `"reason":"unauthenticated"`} {
		if !strings.Contains(got, want) {
			t.Errorf("no record with %s", want)
		}
	}
	for _, leak := range []string{secret, "Lead", strings.ToUpper(groupZ), groupZ, groupR, userA, "context_group", "role_label"} {
		if strings.Contains(got, leak) {
			t.Errorf("a record carries %q", leak)
		}
	}
	if n := strings.Count(got, "\n"); n != 3 {
		t.Errorf("%d records, want 3", n)
	}
}
