package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/tenancy"
	"core/internal/tenancy/access"
)

// G018 (Plan r5 §5.3, §6.2 ONB-H): the onboarding adapter's closed surface without a database — the eight pairs and
// their method/path closure, the pre-session refusals (ambiguous parameters, rejected origins, no session), strict
// bodies and the whole-payload budgets, the refusal mapping, and the D-5 bounded copy of branding/api. The
// authenticated routes run against real PostgreSQL through the facade (TestOnboarding) and the browser suites.

const (
	orgA  = "0f18a000-0000-4000-8000-0000000000a1"
	hostA = "https://org-a.localhost:543"
	gid   = "0f18a000-0000-4000-8000-00000000b001"
)

// fakeDB answers only the resolver's hostname_state statement; anything else is a failure of the test. ctxs records the
// context of each statement and its deadline as seen when the statement ran (the completion deadline tests).
type fakeDB struct {
	stmts []string
	ctxs  []stmtCtx
}

type stmtCtx struct {
	ctx      context.Context
	left     time.Duration
	deadline bool
}

type row struct{ scan func(...any) error }

func (r row) Scan(dest ...any) error { return r.scan(dest...) }

func (f *fakeDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("fakeDB: no transaction expected")
}

func (f *fakeDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	dl, ok := ctx.Deadline()
	f.ctxs = append(f.ctxs, stmtCtx{ctx: ctx, left: time.Until(dl), deadline: ok})
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
	h, db := fixtureHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	return mux, db
}

func fixtureHandler(t *testing.T) (*Handler, *fakeDB) {
	t.Helper()
	db := &fakeDB{}
	d := config.Domains{BaseDomain: "localhost", CNAMETarget: "edge.localhost", HostPort: "543", AuthHost: "auth.localhost"}
	return &Handler{cfg: Config{Resolver: domains.NewResolver(db, d.Hostnames())}, acc: access.NewService(db, nil, nil),
		log: slog.New(slog.DiscardHandler)}, db
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

// The eight pairs (the six paths), each with a body that passes the decoder for its mutation.
var pairs = []struct{ method, path string }{
	{"GET", "/internal/v1/groups"}, {"POST", "/internal/v1/groups/slug-suggestion"}, {"POST", "/internal/v1/groups"},
	{"GET", "/internal/v1/groups/" + gid}, {"POST", "/internal/v1/groups/" + gid + "/business-profile"},
	{"GET", "/internal/v1/business-profiles"}, {"GET", "/internal/v1/business-profiles/" + gid},
	{"PUT", "/internal/v1/business-profiles/" + gid},
}

func mutationBody(t *testing.T, path, origin string) map[string]any {
	t.Helper()
	b := map[string]any{"origin": origin, "csrf_token": "x"}
	switch {
	case strings.HasSuffix(path, "/business-profile"):
		b["mode"], b["profile"] = "new", map[string]any{"display_name": "Acme"}
	case strings.HasPrefix(path, "/internal/v1/business-profiles/"):
		b["profile"] = map[string]any{"display_name": "Acme"}
	default:
		b["name"] = "Research"
	}
	return b
}

func TestOnboardingHTTP(t *testing.T) {
	mux, db := fixture(t)
	if len(Paths) != 6 {
		t.Fatalf("%d paths, want 6", len(Paths))
	}
	for _, p := range pairs {
		name := p.method + " " + p.path
		if p.method == "GET" {
			// No session: unauthenticated (401, closed body); a foreign or malformed origin: origin_rejected; a missing or
			// repeated origin or an unknown parameter: ambiguous_parameter.
			closedAnswer(t, name+" no session", serve(mux, "GET", p.path+q(hostA), nil, ""), 401, "unauthenticated")
			closedAnswer(t, name+" foreign origin", serve(mux, "GET", p.path+q("https://elsewhere.test:543"), nil, ""), 403, "origin_rejected")
			closedAnswer(t, name+" http origin", serve(mux, "GET", p.path+q("http://org-a.localhost:543"), nil, ""), 403, "origin_rejected")
			closedAnswer(t, name+" no origin", serve(mux, "GET", p.path, nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" two origins", serve(mux, "GET", p.path+q(hostA)+"&origin=x", nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" unknown parameter", serve(mux, "GET", p.path+q(hostA)+"&x=1", nil, ""), 400, "ambiguous_parameter")
			continue
		}
		ok := jsonBody(t, mutationBody(t, p.path, hostA))
		closedAnswer(t, name+" no session", serve(mux, p.method, p.path, ok, "application/json"), 401, "unauthenticated")
		closedAnswer(t, name+" foreign origin", serve(mux, p.method, p.path, jsonBody(t, mutationBody(t, p.path, "https://elsewhere.test:543")), "application/json"), 403, "origin_rejected")
		closedAnswer(t, name+" not JSON", serve(mux, p.method, p.path, ok, "text/plain"), 400, "invalid_request")
		closedAnswer(t, name+" query", serve(mux, p.method, p.path+q(hostA), ok, "application/json"), 400, "ambiguous_parameter")
		unknown := mutationBody(t, p.path, hostA)
		unknown["owner_user_id"] = gid // a forged creator is an unknown field
		closedAnswer(t, name+" unknown field", serve(mux, p.method, p.path, jsonBody(t, unknown), "application/json"), 400, "invalid_request")
		closedAnswer(t, name+" trailing data", serve(mux, p.method, p.path, append(ok, []byte(" {}")...), "application/json"), 400, "invalid_request")
		// A trailing closing delimiter is trailing data too (dec.More() alone would miss it); trailing whitespace is not,
		// so that body reaches the session check.
		for _, tail := range []string{"]", "}", " ]", "\n}"} {
			closedAnswer(t, name+fmt.Sprintf(" trailing %q", tail), serve(mux, p.method, p.path, append(append([]byte{}, ok...), tail...), "application/json"), 400, "invalid_request")
		}
		closedAnswer(t, name+" trailing whitespace", serve(mux, p.method, p.path, append(append([]byte{}, ok...), " \r\n\t"...), "application/json"), 401, "unauthenticated")
		closedAnswer(t, name+" malformed", serve(mux, p.method, p.path, []byte("{"), "application/json"), 400, "invalid_request")
	}

	// Every method outside each path's closed set is method_not_allowed (closed JSON, not the mux's text answer).
	served := map[string]bool{}
	for _, p := range pairs {
		served[p.method+" "+p.path] = true
	}
	for _, p := range pairs {
		for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
			if served[m+" "+p.path] {
				continue
			}
			if p.path == "/internal/v1/groups/slug-suggestion" && m == "GET" {
				continue // a {id} value of GET /groups/{id}: not_found after the session (below)
			}
			closedAnswer(t, m+" "+p.path, serve(mux, m, p.path+q(hostA), nil, ""), 405, "method_not_allowed")
		}
	}
	closedAnswer(t, "GET slug-suggestion", serve(mux, "GET", "/internal/v1/groups/slug-suggestion"+q(hostA), nil, ""), 401, "unauthenticated")

	// Sub-path probing below both roots: not_found (closed JSON).
	for _, p := range []string{"/internal/v1/groups/" + gid + "/x", "/internal/v1/business-profiles/x/y", "/internal/v1/groups/x/y/z",
		"/internal/v1/groups/", "/internal/v1/business-profiles/" + gid + "/business-profile"} {
		closedAnswer(t, "sub-path "+p, serve(mux, "GET", p+q(hostA), nil, ""), 404, "not_found")
	}
	for _, s := range db.stmts {
		if s != "hostname_state" {
			t.Fatalf("a statement other than the resolver's ran: %v", db.stmts)
		}
	}

	// G060 (Plan r0 §4.2, r1 C7): a Group row is exactly id, name, slug, business_profile_id, complete, self_role,
	// can_manage and management_source; the four nonmember/member by override cases render truthfully, and an override
	// is never a self membership.
	gID, _ := tenancy.ParseGroupID(gid)
	for _, c := range []struct {
		name   string
		in     access.GroupOnboarding
		self   any
		manage bool
		source any
	}{
		{"nonmember, override off", access.GroupOnboarding{Group: tenancy.Group{ID: gID, Name: "G", Slug: "g"}}, nil, false, nil},
		{"nonmember, override on", access.GroupOnboarding{Group: tenancy.Group{ID: gID, Name: "G", Slug: "g"}, CanManage: true,
			ManagementSource: "organization_override"}, nil, true, "organization_override"},
		{"member, override off", access.GroupOnboarding{Group: tenancy.Group{ID: gID, Name: "G", Slug: "g"},
			SelfRole: tenancy.GroupMember}, "group_member", false, nil},
		{"member, override on", access.GroupOnboarding{Group: tenancy.Group{ID: gID, Name: "G", Slug: "g"}, SelfRole: tenancy.GroupMember,
			CanManage: true, ManagementSource: "organization_override"}, "group_member", true, "organization_override"},
		{"group admin", access.GroupOnboarding{Group: tenancy.Group{ID: gID, Name: "G", Slug: "g"}, SelfRole: tenancy.GroupAdmin,
			CanManage: true, ManagementSource: "group_role"}, "group_admin", true, "group_role"},
	} {
		out := groupJSON(c.in)
		keys := make([]string, 0, len(out))
		for k := range out {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		if strings.Join(keys, ",") != "business_profile_id,can_manage,complete,id,management_source,name,self_role,slug" {
			t.Errorf("%s: group keys %v", c.name, keys)
		}
		if out["self_role"] != c.self || out["can_manage"] != c.manage || out["management_source"] != c.source {
			t.Errorf("%s: hints self_role=%v can_manage=%v management_source=%v", c.name, out["self_role"], out["can_manage"],
				out["management_source"])
		}
	}
}

func TestOnboardingRefusalMapping(t *testing.T) {
	for code, want := range map[access.Code]string{
		access.CodeUnauthenticated: "unauthenticated", access.CodeContextDenied: "not_available", access.CodeForbidden: "forbidden",
		access.CodeNotFound: "not_found", access.CodeConflict: "conflict", access.CodeInvalidInput: "invalid_request",
		access.CodeUnavailable: "onboarding_unavailable", access.CodeInternal: "onboarding_unavailable",
	} {
		if got := failure(&access.Error{Code: code}); got != want {
			t.Errorf("%s -> %s, want %s", code, got, want)
		}
	}
	if got := failure(errors.New("other")); got != "onboarding_unavailable" {
		t.Errorf("a non-access error -> %s", got)
	}
	if len(status) != 11 {
		t.Errorf("%d refusal codes, want 11", len(status))
	}
	// The field code accompanies invalid_request and conflict only; nothing else carries detail.
	for _, c := range []struct {
		err  *access.Error
		keys string
	}{
		{&access.Error{Code: access.CodeConflict, Field: "slug"}, "error,field"},
		{&access.Error{Code: access.CodeInvalidInput, Field: "invalid_cursor"}, "error,field"},
		{&access.Error{Code: access.CodeForbidden, Field: "x"}, "error"},
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
		if strings.Join(sortStrings(keys), ",") != c.keys {
			t.Errorf("%v: %v", c.err.Code, out)
		}
	}
}

func sortStrings(s []string) []string {
	for i := range s {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
	return s
}

// TestOnboardingBodyBudgets (ONB-H9): the budgets are recomputed from the tenancy maxima; the max-field all-astral
// payload is admitted raw and fully escaped (decoded intact; it reaches the session check), the validators still refuse
// a 4,001-code-point description, and one byte over the budget is invalid_request before any decode.
func TestOnboardingBodyBudgets(t *testing.T) {
	const perCodePoint, envelope = 12, 1024
	profile := perCodePoint*(tenancy.MaxDisplayNameLength+tenancy.MaxLegalNameLength+tenancy.MaxDescriptionLength+
		tenancy.MaxIndustryLength+tenancy.MaxWebsiteLength) + envelope
	small := perCodePoint*tenancy.MaxNameLength + tenancy.MaxSlugLength + envelope
	if profile != 82780 || profile > maxProfileBody || maxProfileBody != 96<<10 {
		t.Errorf("profile budget %d (computed %d)", maxProfileBody, profile)
	}
	if small != 3487 || small > maxBody || maxBody != 8<<10 {
		t.Errorf("small budget %d (computed %d)", maxBody, small)
	}

	mux, _ := fixture(t)
	astral := func(n int) string { return strings.Repeat("\U0001F600", n) }
	site := "https://example.test/" + strings.Repeat("a", tenancy.MaxWebsiteLength-len("https://example.test/"))
	fields := map[string]any{"display_name": astral(tenancy.MaxDisplayNameLength), "legal_name": astral(tenancy.MaxLegalNameLength),
		"description": astral(tenancy.MaxDescriptionLength), "industry": astral(tenancy.MaxIndustryLength), "website": site}
	raw := jsonBody(t, map[string]any{"origin": hostA, "csrf_token": "x", "profile": fields})
	escaped := escapeAll(raw)
	if len(escaped) > maxProfileBody || len(escaped) <= len(raw) {
		t.Fatalf("escaped payload %d bytes (raw %d)", len(escaped), len(raw))
	}
	for name, b := range map[string][]byte{"raw": raw, "escaped": escaped} {
		var in body
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil || in.Profile == nil || in.Profile.Description == nil || *in.Profile.Description != fields["description"] {
			t.Fatalf("%s: not decoded intact: %v", name, err)
		}
		if _, err := tenancy.ValidateBusinessContext(in.Profile.context()); err != nil {
			t.Errorf("%s: the max-field context is refused: %v", name, err)
		}
		// Admitted by the budget and the decoder: the request reaches the session check.
		closedAnswer(t, name+" max-field PUT", serve(mux, "PUT", "/internal/v1/business-profiles/"+gid, b, "application/json"), 401, "unauthenticated")
	}

	over := astral(tenancy.MaxDescriptionLength + 1)
	b := jsonBody(t, map[string]any{"origin": hostA, "csrf_token": "x", "profile": map[string]any{"display_name": "A", "description": over}})
	var in body
	if err := json.Unmarshal(b, &in); err != nil || len(b) > maxProfileBody {
		t.Fatalf("4,001-code-point body: %v (%d bytes)", err, len(b))
	}
	if _, err := tenancy.ValidateBusinessContext(in.Profile.context()); tenancy.CodeOf(err) != tenancy.CodeInvalidDescription {
		t.Errorf("4,001-code-point description: %v", err)
	}

	pad := func(budget int) []byte {
		b := []byte(`{"origin":"` + hostA + `","csrf_token":"x","name":"A"}`)
		return append(b, bytes.Repeat([]byte(" "), budget+1-len(b))...)
	}
	closedAnswer(t, "maxProfileBody+1", serve(mux, "PUT", "/internal/v1/business-profiles/"+gid, pad(maxProfileBody), "application/json"), 400, "invalid_request")
	closedAnswer(t, "maxProfileBody+1 completion", serve(mux, "POST", "/internal/v1/groups/"+gid+"/business-profile", pad(maxProfileBody), "application/json"), 400, "invalid_request")
	for _, p := range []string{"/internal/v1/groups", "/internal/v1/groups/slug-suggestion"} {
		closedAnswer(t, "maxBody+1 "+p, serve(mux, "POST", p, pad(maxBody), "application/json"), 400, "invalid_request")
		if w := serve(mux, "POST", p, pad(maxBody-1), "application/json"); w.Code != 401 {
			t.Errorf("a body at the budget on %s: %d", p, w.Code)
		}
	}
}

// escapeAll re-encodes every string character of a JSON document as \uXXXX (astral code points as surrogate pairs).
func escapeAll(doc []byte) []byte {
	var out bytes.Buffer
	in := false
	s := string(doc)
	for i := 0; i < len(s); {
		r, size := rune(s[i]), 1
		if r >= 0x80 {
			r, size = utf8.DecodeRuneInString(s[i:])
		}
		switch {
		case r == '"' && (i == 0 || s[i-1] != '\\'):
			in = !in
			out.WriteByte('"')
		case in && r == '\\':
			out.WriteString(s[i : i+2])
			size = 2
		case in && r > 0xFFFF:
			r -= 0x10000
			out.WriteString(jsonEscape(0xD800 + (r >> 10)))
			out.WriteString(jsonEscape(0xDC00 + (r & 0x3FF)))
		case in:
			out.WriteString(jsonEscape(r))
		default:
			out.WriteRune(r)
		}
		i += size
	}
	return out.Bytes()
}

func jsonEscape(r rune) string {
	const hex = "0123456789abcdef"
	return `\u` + string([]byte{hex[r>>12&0xF], hex[r>>8&0xF], hex[r>>4&0xF], hex[r&0xF]})
}

// parsed is one source file of the D-5 copy check.
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

// TestOnboardingCopyIsExact (D-5): the session-scope behaviour is a bounded copy of branding/api (itself pinned to
// domains/api), not a shared helper; this pins begin, scope and originHost byte for byte and the eight shared refusal
// codes with their wire values and statuses.
func TestOnboardingCopyIsExact(t *testing.T) {
	brd, onb := parseSource(t, "../../branding/api/http.go"), parseSource(t, "http.go")
	for _, name := range []string{"begin", "scope", "originHost"} {
		if a, b := brd.decl(t, name), onb.decl(t, name); a != b {
			t.Errorf("%s differs from branding/api:\n--- branding\n%s\n--- onboarding\n%s", name, a, b)
		}
	}
	shared := []string{"errInvalidRequest", "errAmbiguous", "errOriginRejected", "errUnauthenticated", "errCSRF", "errNotAvailable",
		"errForbidden", "errNotFound"}
	bc, oc := brd.consts(), onb.consts()
	bs, ost := brd.statuses(t), onb.statuses(t)
	for _, n := range shared {
		if bc[n] == "" || bc[n] != oc[n] || bs[n] == "" || bs[n] != ost[n] {
			t.Errorf("%s: branding %s/%s, onboarding %s/%s", n, bc[n], bs[n], oc[n], ost[n])
		}
	}
	if bs["errUnavailable"] != "503" || ost["errUnavailable"] != "503" || oc["errUnavailable"] != `"onboarding_unavailable"` {
		t.Errorf("errUnavailable: branding %s, onboarding %s %s", bs["errUnavailable"], oc["errUnavailable"], ost["errUnavailable"])
	}
	for n, v := range oc {
		if strings.HasPrefix(n, "err") && ost[n] == "" {
			t.Errorf("%s (%s) has no status", n, v)
		}
	}
	if len(ost) != len(status) {
		t.Errorf("parsed %d statuses, the table has %d", len(ost), len(status))
	}
}

func TestOnboardingNewRefusesIncompleteConfiguration(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("an incomplete configuration was accepted")
	}
	if hostTag("portal.acme.test") == hostTag("portal.acme.tes") || len(hostTag("x")) != 12 {
		t.Error("host tag")
	}
}

// TestOnboardingCompletionAdmission (availability addendum): the completion admission set never waits, admits one
// completion per subject key (two sessions of one user share the key) and at most maxCompletionsInFlight keys, and
// releases on every outcome; a stale release never frees another holder's entry. Through the mux, a refused or
// exhausted admission never precedes the authenticated begin: with every entry held an unauthenticated completion is
// still unauthenticated, not onboarding_unavailable, and takes no entry.
func TestOnboardingCompletionAdmission(t *testing.T) {
	var a admission
	alice, bob := [16]byte{0: 1}, [16]byte{0: 2}
	refusedAtOnce := func(what string, key [16]byte) {
		t.Helper()
		got := make(chan bool, 1)
		go func() { _, ok := a.try(key); got <- ok }()
		select {
		case ok := <-got:
			if ok {
				t.Errorf("%s: admitted", what)
			}
		case <-time.After(time.Second):
			t.Fatalf("%s: try waited", what)
		}
	}
	admitted := func(what string, key [16]byte) func() {
		t.Helper()
		rel, ok := a.try(key)
		if !ok {
			t.Fatalf("%s: refused", what)
		}
		return rel
	}

	// Session 1 of alice holds; session 2 of the same subject is refused without waiting; bob is admitted.
	rel := admitted("alice session 1", alice)
	refusedAtOnce("alice session 2 while session 1 holds", alice)
	relB := admitted("bob while alice holds", bob)
	rel()
	rel2 := admitted("alice after release", alice)
	rel() // stale and repeated: must not free rel2's entry
	refusedAtOnce("alice after a stale release", alice)
	rel2()
	relB()

	// Release on panic (the deferred release of a completion that panics) and on cancellation.
	func() {
		defer func() { _ = recover() }()
		r := admitted("alice before a panic", alice)
		defer r()
		panic("completion panic")
	}()
	admitted("alice after the panic", alice)()
	ctx, cancel := context.WithCancel(context.Background())
	held, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		r, ok := a.try(alice)
		if !ok {
			close(held)
			return
		}
		defer r()
		close(held)
		<-ctx.Done() // a completion that ends on its request's cancellation
	}()
	<-held
	refusedAtOnce("alice while a cancellable completion holds", alice)
	cancel()
	<-done
	admitted("alice after the cancellation", alice)()

	// Cap: maxCompletionsInFlight distinct subjects fill the set; another subject is refused at once; one release
	// admits a new subject.
	rels := make([]func(), 0, maxCompletionsInFlight)
	for i := range maxCompletionsInFlight {
		rels = append(rels, admitted(fmt.Sprintf("subject %d", i), [16]byte{0: 0xa0, 1: byte(i >> 8), 2: byte(i)}))
	}
	refusedAtOnce("a new subject at the cap", bob)

	// Placement: the full set does not turn an unauthenticated completion into onboarding_unavailable.
	h, _ := fixtureHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	for i := range maxCompletionsInFlight {
		if _, ok := h.admit.try([16]byte{0: 0xb0, 1: byte(i >> 8), 2: byte(i)}); !ok {
			t.Fatalf("handler subject %d refused", i)
		}
	}
	path := "/internal/v1/groups/" + gid + "/business-profile"
	closedAnswer(t, "completion at the cap without a session", serve(mux, "POST", path, jsonBody(t, mutationBody(t, path, hostA)), "application/json"), 401, "unauthenticated")
	if n := len(h.admit.active); n != maxCompletionsInFlight {
		t.Errorf("handler admission holds %d entries, want %d", n, maxCompletionsInFlight)
	}

	rels[0]()
	admitted("bob after one release at the cap", bob)()
	for _, r := range rels {
		r()
	}
	if n := len(a.active); n != 0 {
		t.Errorf("%d entries left after every release", n)
	}
}

// TestOnboardingCompletionDeadline (availability addendum): only the completion route runs under the request
// deadline; the resolver statement of its begin already sees it (at most completionRequestDeadline, an earlier caller
// deadline kept), the bounded context ends with the request, and a stalled body ends at the read deadline as
// invalid_request without a decision.
func TestOnboardingCompletionDeadline(t *testing.T) {
	mux, db := fixture(t)
	for _, p := range pairs {
		db.ctxs = nil
		if p.method == "GET" {
			serve(mux, "GET", p.path+q(hostA), nil, "")
		} else {
			serve(mux, p.method, p.path, jsonBody(t, mutationBody(t, p.path, hostA)), "application/json")
		}
		if len(db.ctxs) == 0 {
			t.Fatalf("%s %s: no resolver statement", p.method, p.path)
		}
		s := db.ctxs[len(db.ctxs)-1]
		if strings.HasSuffix(p.path, "/business-profile") {
			if !s.deadline || s.left > completionRequestDeadline || s.left < completionRequestDeadline-time.Second {
				t.Errorf("completion: deadline %v, %v left; want at most %v", s.deadline, s.left, completionRequestDeadline)
			}
			if s.ctx.Err() == nil {
				t.Error("completion: the bounded context outlived the request")
			}
		} else if s.deadline {
			t.Errorf("%s %s: unexpected deadline (%v left)", p.method, p.path, s.left)
		}
	}

	// An earlier caller deadline is kept.
	path := "/internal/v1/groups/" + gid + "/business-profile"
	caller, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	r := httptest.NewRequest("POST", path, bytes.NewReader(jsonBody(t, mutationBody(t, path, hostA)))).WithContext(caller)
	r.Header.Set("Content-Type", "application/json")
	db.ctxs = nil
	mux.ServeHTTP(httptest.NewRecorder(), r)
	if len(db.ctxs) == 0 || !db.ctxs[0].deadline || db.ctxs[0].left > 2*time.Second {
		t.Errorf("earlier caller deadline not kept: %+v", db.ctxs)
	}

	// The read deadline: the completion mutation behind bounded (a short bound here) refuses a stalled body.
	h, db2 := fixtureHandler(t)
	const bound = 300 * time.Millisecond
	srv := httptest.NewServer(bounded(bound, h.mutation("complete_onboarding", maxProfileBody, h.completeGroup)))
	defer srv.Close()
	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	start := time.Now()
	if _, err := io.WriteString(conn, "POST "+path+" HTTP/1.1\r\nHost: core\r\nContent-Type: application/json\r\nContent-Length: 200\r\n\r\n{\"origin\":"); err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("stalled body: no answer: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if el := time.Since(start); resp.StatusCode != 400 || out["error"] != "invalid_request" || el < bound || el > 5*time.Second {
		t.Errorf("stalled body: %d %v after %v, want 400 invalid_request after the %v read deadline", resp.StatusCode, out, el, bound)
	}
	if len(db2.stmts) != 0 {
		t.Errorf("stalled body reached the resolver: %v", db2.stmts)
	}
}
