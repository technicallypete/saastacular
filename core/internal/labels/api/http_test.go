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

// G020 (Plan r2 §4.2, T9): the labels adapter's closed surface without a database — the seven pairs on six paths and
// their method/path closure, the pre-session refusals (ambiguous parameters, rejected origins, no session), strict
// bodies and the budget, the not_found parity of unparsable path ids, the refusal mapping with its field codes, the
// labels.request event (no name, id or payload) and the bounded copy of membership/api. The authenticated routes run
// against real PostgreSQL through the facade (TestLabels) and the browser suite (labels.spec.ts).

const (
	orgA  = "0f20a000-0000-4000-8000-0000000000a1"
	hostA = "https://org-a.localhost:543"
	gid   = "0f20a000-0000-4000-8000-00000000b001"
	lid   = "0f20a000-0000-4000-8000-00000000e001"
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

func handler(db *fakeDB, log *slog.Logger) *Handler {
	d := config.Domains{BaseDomain: "localhost", CNAMETarget: "edge.localhost", HostPort: "543", AuthHost: "auth.localhost"}
	return &Handler{cfg: Config{Resolver: domains.NewResolver(db, d.Hostnames())}, acc: access.NewService(db, nil, nil), log: log}
}

func fixture(t *testing.T) (*http.ServeMux, *fakeDB) {
	t.Helper()
	db := &fakeDB{}
	mux := http.NewServeMux()
	handler(db, slog.New(slog.DiscardHandler)).Register(mux)
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

// The seven pairs (the six paths), each mutation with a body that passes the decoder.
var pairs = []struct{ method, path string }{
	{"GET", "/internal/v1/groups/" + gid + "/labels"}, {"GET", "/internal/v1/labels"},
	{"POST", "/internal/v1/groups/" + gid + "/labels/assign"}, {"POST", "/internal/v1/groups/" + gid + "/labels/unassign"},
	{"POST", "/internal/v1/labels"}, {"POST", "/internal/v1/labels/" + lid + "/delete"},
	{"POST", "/internal/v1/labels/" + lid + "/update"},
}

func mutationBody(path, origin string) map[string]any {
	b := map[string]any{"origin": origin, "csrf_token": "x"}
	switch {
	case strings.HasSuffix(path, "/assign"), strings.HasSuffix(path, "/unassign"):
		b["label_id"] = lid
	case path == "/internal/v1/labels":
		b["scope"], b["group_id"], b["name"], b["visibility"] = "group", gid, "Priority", "target_members"
	case strings.HasSuffix(path, "/update"):
		b["name"], b["visibility"] = "Priority", "owner_scope"
	}
	return b
}

func TestLabelsHTTP(t *testing.T) {
	mux, db := fixture(t)
	if len(Paths) != 6 || len(pairs) != 7 {
		t.Fatalf("%d paths and %d pairs, want 6 and 7", len(Paths), len(pairs))
	}
	for _, p := range pairs {
		name := p.method + " " + p.path
		if p.method == "GET" {
			closedAnswer(t, name+" no session", serve(mux, "GET", p.path+q(hostA), nil, ""), 401, "unauthenticated")
			closedAnswer(t, name+" foreign origin", serve(mux, "GET", p.path+q("https://elsewhere.test:543"), nil, ""), 403, "origin_rejected")
			closedAnswer(t, name+" http origin", serve(mux, "GET", p.path+q("http://org-a.localhost:543"), nil, ""), 403, "origin_rejected")
			closedAnswer(t, name+" origin with path", serve(mux, "GET", p.path+q(hostA+"/x"), nil, ""), 403, "origin_rejected")
			closedAnswer(t, name+" no origin", serve(mux, "GET", p.path, nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" two origins", serve(mux, "GET", p.path+q(hostA)+"&origin=x", nil, ""), 400, "ambiguous_parameter")
			closedAnswer(t, name+" unknown parameter", serve(mux, "GET", p.path+q(hostA)+"&organization_id="+orgA, nil, ""), 400, "ambiguous_parameter")
			if p.path == "/internal/v1/labels" {
				closedAnswer(t, name+" two scopes", serve(mux, "GET", p.path+q(hostA)+"&scope=group&scope=organization", nil, ""), 400, "ambiguous_parameter")
				closedAnswer(t, name+" all parameters", serve(mux, "GET", p.path+q(hostA)+"&scope=group&group_id="+gid+"&after="+lid+"&limit=5", nil, ""), 401, "unauthenticated")
			} else {
				// The Group labels read takes no parameter beyond origin.
				closedAnswer(t, name+" scope parameter", serve(mux, "GET", p.path+q(hostA)+"&scope=organization", nil, ""), 400, "ambiguous_parameter")
			}
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
		closedAnswer(t, name+" json charset", serve(mux, p.method, p.path, ok, "application/json; charset=utf-8"), 401, "unauthenticated")
		closedAnswer(t, name+" foreign origin", serve(mux, p.method, p.path, jsonBody(t, mutationBody(p.path, "https://elsewhere.test:543")), "application/json"), 403, "origin_rejected")
		closedAnswer(t, name+" http origin", serve(mux, p.method, p.path, jsonBody(t, mutationBody(p.path, "http://org-a.localhost:543")), "application/json"), 403, "origin_rejected")
		noOrigin := mutationBody(p.path, hostA)
		delete(noOrigin, "origin")
		closedAnswer(t, name+" no origin", serve(mux, p.method, p.path, jsonBody(t, noOrigin), "application/json"), 403, "origin_rejected")
		closedAnswer(t, name+" not JSON", serve(mux, p.method, p.path, ok, "text/plain"), 400, "invalid_request")
		closedAnswer(t, name+" form", serve(mux, p.method, p.path, ok, "application/x-www-form-urlencoded"), 400, "invalid_request")
		closedAnswer(t, name+" query", serve(mux, p.method, p.path+q(hostA), ok, "application/json"), 400, "ambiguous_parameter")
		for _, forged := range []string{"organization_id", "selected_group_id", "user_id", "kind", "id", "target_id", "assignments"} {
			unknown := mutationBody(p.path, hostA)
			unknown[forged] = orgA // a forged tenant, subject or target is an unknown field
			closedAnswer(t, name+" unknown field "+forged, serve(mux, p.method, p.path, jsonBody(t, unknown), "application/json"), 400, "invalid_request")
		}
		for _, tail := range []string{" {}", "]", "}", "\n}"} {
			closedAnswer(t, name+fmt.Sprintf(" trailing %q", tail), serve(mux, p.method, p.path, append(append([]byte{}, ok...), tail...), "application/json"), 400, "invalid_request")
		}
		closedAnswer(t, name+" trailing whitespace", serve(mux, p.method, p.path, append(append([]byte{}, ok...), " \r\n\t"...), "application/json"), 401, "unauthenticated")
		closedAnswer(t, name+" malformed", serve(mux, p.method, p.path, []byte("{"), "application/json"), 400, "invalid_request")
		closedAnswer(t, name+" array", serve(mux, p.method, p.path, []byte("[]"), "application/json"), 400, "invalid_request")
		wrongType := mutationBody(p.path, hostA)
		wrongType["csrf_token"] = 1
		closedAnswer(t, name+" wrong type", serve(mux, p.method, p.path, jsonBody(t, wrongType), "application/json"), 400, "invalid_request")
		big := mutationBody(p.path, hostA)
		big["csrf_token"] = strings.Repeat("x", maxBody)
		closedAnswer(t, name+" over budget", serve(mux, p.method, p.path, jsonBody(t, big), "application/json"), 400, "invalid_request")
		// Exactly maxBody bytes is within the budget (then refused only for the missing session).
		edge := mutationBody(p.path, hostA)
		edge["csrf_token"] = ""
		pad := maxBody - len(jsonBody(t, edge))
		edge["csrf_token"] = strings.Repeat("x", pad)
		if b := jsonBody(t, edge); len(b) != maxBody {
			t.Fatalf("%s: edge body %d bytes", name, len(b))
		}
		closedAnswer(t, name+" at budget", serve(mux, p.method, p.path, jsonBody(t, edge), "application/json"), 401, "unauthenticated")
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

	// Sub-path probing below /internal/v1/labels/: not_found (closed JSON).
	for _, p := range []string{"/internal/v1/labels/", "/internal/v1/labels/" + lid, "/internal/v1/labels/" + lid + "/x",
		"/internal/v1/labels/" + lid + "/assign", "/internal/v1/labels/" + lid + "/update/x", "/internal/v1/labels/x/y/z"} {
		closedAnswer(t, "sub-path "+p, serve(mux, "GET", p+q(hostA), nil, ""), 404, "not_found")
	}
	for _, s := range db.stmts {
		if s != "hostname_state" {
			t.Fatalf("a statement other than the resolver's ran: %v", db.stmts)
		}
	}

	// One labels.request per request, named by operation; never a name, a label or Group id, a token or a payload.
	var buf bytes.Buffer
	mux2 := http.NewServeMux()
	handler(&fakeDB{}, slog.New(slog.NewJSONHandler(&buf, nil))).Register(mux2)
	ops := map[string]string{"GET /internal/v1/labels": "list_labels", "GET /internal/v1/groups/" + gid + "/labels": "list_group_labels",
		"POST /internal/v1/labels": "create_label", "POST /internal/v1/labels/" + lid + "/update": "update_label",
		"POST /internal/v1/labels/" + lid + "/delete": "delete_label", "POST /internal/v1/groups/" + gid + "/labels/assign": "assign_label",
		"POST /internal/v1/groups/" + gid + "/labels/unassign": "unassign_label"}
	for _, p := range pairs {
		buf.Reset()
		if p.method == "GET" {
			serve(mux2, "GET", p.path+q(hostA), nil, "")
		} else {
			serve(mux2, "POST", p.path, jsonBody(t, mutationBody(p.path, hostA)), "application/json")
		}
		var rec map[string]any
		if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
			t.Fatalf("%s: event %q: %v", p.path, buf.String(), err)
		}
		want := ops[p.method+" "+p.path]
		if rec["msg"] != "labels.request" || rec["operation"] != want || rec["outcome"] != "refused" ||
			rec["reason"] != "unauthenticated" || rec["level"] != "WARN" {
			t.Errorf("%s %s: event %v, want operation %s", p.method, p.path, rec, want)
		}
		for k := range rec {
			if !map[string]bool{"time": true, "level": true, "msg": true, "operation": true, "outcome": true, "reason": true, "host_tag": true}[k] {
				t.Errorf("%s: event carries %s", p.path, k)
			}
		}
		for _, secret := range []string{"Priority", gid, lid, orgA, "org-a.localhost", "csrf"} {
			if strings.Contains(buf.String(), secret) {
				t.Errorf("%s: event contains %q: %s", p.path, secret, buf.String())
			}
		}
	}
	buf.Reset()
	serve(mux2, "DELETE", "/internal/v1/labels", nil, "")
	if !strings.Contains(buf.String(), `"operation":"method"`) || !strings.Contains(buf.String(), `"reason":"method_not_allowed"`) {
		t.Errorf("method event: %s", buf.String())
	}
}

// TestLabelsRefusalMapping: every access.Code maps onto its closed refusal and status; the field code accompanies
// invalid_request, forbidden and conflict only; unparsable path ids are not_found (the unknown/foreign answer) and a
// malformed body id or Group id is invalid_request invalid_id, all before any facade call.
func TestLabelsRefusalMapping(t *testing.T) {
	for code, want := range map[access.Code]string{
		access.CodeUnauthenticated: "unauthenticated", access.CodeContextDenied: "not_available", access.CodeForbidden: "forbidden",
		access.CodeNotFound: "not_found", access.CodeConflict: "conflict", access.CodeInvalidInput: "invalid_request",
		access.CodeUnavailable: "labels_unavailable", access.CodeInternal: "labels_unavailable",
		access.CodeHostTaken: "labels_unavailable", access.CodeDomainStateInvalid: "labels_unavailable",
	} {
		if got := failure(&access.Error{Code: code}); got != want {
			t.Errorf("%s -> %s, want %s", code, got, want)
		}
	}
	if got := failure(errors.New("other")); got != "labels_unavailable" {
		t.Errorf("a non-access error -> %s", got)
	}
	if got := failure(fmt.Errorf("wrapped: %w", &access.Error{Code: access.CodeNotFound})); got != "not_found" {
		t.Errorf("a wrapped not_found -> %s", got)
	}
	want := map[string]int{"invalid_request": 400, "ambiguous_parameter": 400, "origin_rejected": 403, "unauthenticated": 401,
		"csrf_failed": 403, "not_available": 403, "forbidden": 403, "not_found": 404, "method_not_allowed": 405, "conflict": 409,
		"labels_unavailable": 503}
	if len(status) != len(want) {
		t.Errorf("%d refusal codes, want %d", len(status), len(want))
	}
	for c, s := range want {
		if status[c] != s {
			t.Errorf("%s: status %d, want %d", c, status[c], s)
		}
	}
	for _, c := range []struct {
		err    *access.Error
		status int
		keys   string
	}{
		{&access.Error{Code: access.CodeConflict, Field: "already_assigned"}, 409, "error,field"},
		{&access.Error{Code: access.CodeConflict, Field: "name_conflict"}, 409, "error,field"},
		{&access.Error{Code: access.CodeInvalidInput, Field: "visibility_not_allowed"}, 400, "error,field"},
		{&access.Error{Code: access.CodeInvalidInput, Field: "organization_context_required"}, 400, "error,field"},
		{&access.Error{Code: access.CodeForbidden, Field: "x"}, 403, "error,field"},
		{&access.Error{Code: access.CodeNotFound, Field: "x"}, 404, "error"},
		{&access.Error{Code: access.CodeContextDenied, Field: "x"}, 403, "error"},
		{&access.Error{Code: access.CodeUnavailable, Field: "store_unavailable"}, 503, "error"},
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
		if strings.Join(keys, ",") != c.keys || w.Code != c.status {
			t.Errorf("%v: %d %v", c.err.Code, w.Code, out)
		}
	}

	// The id helpers: the canonical parser refuses non-canonical forms and the zero id; formatID is lowercase.
	for _, bad := range []string{"", "x", "labels", "00000000-0000-0000-0000-000000000000", "{" + lid + "}", strings.ReplaceAll(lid, "-", ""), lid + " "} {
		r := httptest.NewRequest("POST", "/", nil)
		r.SetPathValue("id", bad)
		if _, ok := pathLabel(r); ok {
			t.Errorf("label id %q parsed", bad)
		}
		if _, ok := groupTarget(r); ok {
			t.Errorf("Group id %q parsed", bad)
		}
	}
	r := httptest.NewRequest("POST", "/", nil)
	r.SetPathValue("id", strings.ToUpper(lid))
	if id, ok := pathLabel(r); !ok || id != lid {
		t.Errorf("label id round trip: %q %v", id, ok)
	}
	r.SetPathValue("id", strings.ToUpper(gid))
	if tg, ok := groupTarget(r); !ok || tg != (access.LabelTargetSelector{Group: gid}) {
		t.Errorf("Group target: %+v %v", tg, ok)
	}

	// 404 parity and body refusals of each mutation, decided before any facade call (scope is empty: reaching the
	// facade would be a test failure through a nil context).
	h := handler(&fakeDB{}, slog.New(slog.DiscardHandler))
	name, vis, label := ptr("Priority"), ptr("owner_scope"), ptr(lid)
	for _, c := range []struct {
		name   string
		act    func(*recorder, *http.Request, scope, body)
		id     string
		in     body
		status int
		code   string
		field  string
	}{
		{"update malformed id", h.updateLabel, "x", body{Name: name, Visibility: vis}, 404, "not_found", ""},
		{"update zero id", h.updateLabel, "00000000-0000-0000-0000-000000000000", body{Name: name, Visibility: vis}, 404, "not_found", ""},
		{"update without visibility", h.updateLabel, lid, body{Name: name}, 400, "invalid_request", ""},
		{"update with scope", h.updateLabel, lid, body{Name: name, Visibility: vis, Scope: ptr("group")}, 400, "invalid_request", ""},
		{"delete malformed id", h.deleteLabel, "labels", body{}, 404, "not_found", ""},
		{"delete with name", h.deleteLabel, lid, body{Name: name}, 400, "invalid_request", ""},
		{"assign malformed Group", h.assignLabel, "x", body{LabelID: label}, 404, "not_found", ""},
		{"assign without label_id", h.assignLabel, gid, body{}, 400, "invalid_request", ""},
		{"assign with group_id", h.assignLabel, gid, body{LabelID: label, GroupID: ptr(gid)}, 400, "invalid_request", ""},
		{"assign malformed label_id", h.assignLabel, gid, body{LabelID: ptr("x")}, 400, "invalid_request", "invalid_id"},
		{"unassign malformed Group", h.unassignLabel, "00000000-0000-0000-0000-000000000000", body{LabelID: label}, 404, "not_found", ""},
		{"unassign malformed label_id", h.unassignLabel, gid, body{LabelID: ptr("")}, 400, "invalid_request", "invalid_id"},
		{"create without visibility", h.createLabel, "", body{Scope: ptr("organization"), Name: name}, 400, "invalid_request", ""},
		{"create with label_id", h.createLabel, "", body{Scope: ptr("organization"), Name: name, Visibility: vis, LabelID: label}, 400, "invalid_request", ""},
		{"create malformed group_id", h.createLabel, "", body{Scope: ptr("group"), GroupID: ptr("x"), Name: name, Visibility: vis}, 400, "invalid_request", "invalid_id"},
	} {
		w := httptest.NewRecorder()
		rec := &recorder{ResponseWriter: w}
		r := httptest.NewRequest("POST", "/", nil)
		r.SetPathValue("id", c.id)
		c.act(rec, r, scope{}, c.in)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		if w.Code != c.status || out["error"] != c.code || rec.code != c.code {
			t.Errorf("%s: %d %v (recorded %q), want %d %s", c.name, w.Code, out, rec.code, c.status, c.code)
		}
		if f, _ := out["field"].(string); f != c.field {
			t.Errorf("%s: field %q, want %q", c.name, f, c.field)
		}
	}
}

func ptr(s string) *string { return &s }

// TestLabelsBodyShapes: each route's member set is closed (required present, nothing else), and the answer objects
// carry exactly the §4.2 keys.
func TestLabelsBodyShapes(t *testing.T) {
	sc, g, name, vis, label := ptr("group"), ptr(gid), ptr("n"), ptr("owner_scope"), ptr(lid)
	create := [2][]string{{"scope", "name", "visibility"}, {"group_id"}}
	update := [2][]string{{"name", "visibility"}, nil}
	assign := [2][]string{{"label_id"}, nil}
	none := [2][]string{nil, nil}
	cases := []struct {
		in    body
		shape [2][]string
		want  bool
	}{
		{body{Scope: sc, GroupID: g, Name: name, Visibility: vis}, create, true},
		{body{Scope: ptr("organization"), Name: name, Visibility: vis}, create, true},
		{body{Scope: sc, Name: name}, create, false},
		{body{Name: name, Visibility: vis}, create, false},
		{body{Scope: sc, GroupID: g, Name: name, Visibility: vis, LabelID: label}, create, false},
		{body{Name: name, Visibility: vis}, update, true},
		{body{Name: name}, update, false},
		{body{Name: name, Visibility: vis, GroupID: g}, update, false},
		{body{}, none, true},
		{body{Name: name}, none, false},
		{body{LabelID: label}, assign, true},
		{body{}, assign, false},
		{body{LabelID: label, Scope: sc}, assign, false},
	}
	for i, c := range cases {
		if got := c.in.shape(c.shape[0], c.shape[1]); got != c.want {
			t.Errorf("case %d: %v, want %v", i, got, c.want)
		}
	}
	// The decoder knows exactly the seven members of the union body.
	var tags []string
	for _, f := range regexp.MustCompile("`json:\"([a-z_]+)\"`").FindAllStringSubmatch(decl(t, "http.go", "body"), -1) {
		tags = append(tags, f[1])
	}
	if got := strings.Join(tags, ","); got != "origin,csrf_token,scope,group_id,name,visibility,label_id" {
		t.Errorf("body members %s", got)
	}

	// The definition row: exactly id, scope, group_id, name, visibility (group_id null for an Organization-owned one).
	var v access.LabelDefinitionView
	if !parseID(&v.ID, lid) {
		t.Fatal("label id")
	}
	v.Name, v.Scope, v.Visibility = "Priority", "organization", "owner_scope"
	keys := func(m map[string]any) string {
		ks := []string{}
		for k := range m {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		return strings.Join(ks, ",")
	}
	org := labelJSON(v)
	if keys(org) != "group_id,id,name,scope,visibility" || org["group_id"] != nil || org["id"] != lid || org["scope"] != "organization" ||
		org["name"] != "Priority" || org["visibility"] != "owner_scope" {
		t.Errorf("organization row %v", org)
	}
	var grp access.GroupID
	if !parseID(&grp, gid) {
		t.Fatal("group id")
	}
	v.Scope, v.Group, v.Visibility = "group", &grp, "target_members"
	if row := labelJSON(v); keys(row) != "group_id,id,name,scope,visibility" || row["group_id"] != gid || row["scope"] != "group" ||
		row["visibility"] != "target_members" {
		t.Errorf("group row %v", row)
	}
	// The answers of the other routes (source pins: the literal keys of each writeJSON).
	src, err := os.ReadFile("http.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`map[string]any{"labels": rows, "next": nil}`, `map[string]any{"labels": rows}`,
		`map[string]any{"label_id": formatID(a.LabelID), "name": a.Name}`,
		`map[string]any{"deleted": true, "assignments_removed": res.AssignmentsRemoved}`,
		`map[string]any{"assigned": true}`, `map[string]any{"assigned": false}`,
	} {
		if !bytes.Contains(src, []byte(want)) {
			t.Errorf("answer %s not found", want)
		}
	}
}

// TestLabelsRoutesPinned: the Route constants are exactly the seven reviewed pairs (the CORE_LABELS_ROUTES mechanism
// reads the same lines), declared C-sorted, every pair's path is a served path, the paths are C-sorted, and the
// patterns coexist on one mux with onboarding/api's /internal/v1/groups patterns and membership/api's (main.go).
func TestLabelsRoutesPinned(t *testing.T) {
	src, err := os.ReadFile("http.go")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, m := range regexp.MustCompile(`(?m)^\s+Route[A-Za-z]+\s+= "([^"]+)"$`).FindAllSubmatch(src, -1) {
		got = append(got, string(m[1]))
	}
	if !sort.StringsAreSorted(got) || !sort.StringsAreSorted(Paths) {
		t.Errorf("not C-sorted: %v %v", got, Paths)
	}
	want := "GET /internal/v1/groups/{id}/labels|GET /internal/v1/labels|POST /internal/v1/groups/{id}/labels/assign|" +
		"POST /internal/v1/groups/{id}/labels/unassign|POST /internal/v1/labels|POST /internal/v1/labels/{id}/delete|" +
		"POST /internal/v1/labels/{id}/update"
	if strings.Join(got, "|") != want {
		t.Errorf("routes:\n got %s\nwant %s", strings.Join(got, "|"), want)
	}
	paths := map[string]bool{}
	for _, p := range Paths {
		paths[p] = true
	}
	used := map[string]bool{}
	for _, r := range got {
		_, p, _ := strings.Cut(r, " ")
		if !paths[p] {
			t.Errorf("%s: path not served", r)
		}
		used[p] = true
	}
	if len(used) != len(Paths) {
		t.Errorf("a served path has no route: %v", Paths)
	}

	// Coexistence: register the other adapters' patterns as their Register does (each Route, the method-less closure of
	// each served path except the literals matched before an {id} template, and the closed roots), then this one (a
	// conflict panics).
	mux := http.NewServeMux()
	nop := func(http.ResponseWriter, *http.Request) {}
	seen := map[string]bool{}
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			mux.HandleFunc(p, nop)
		}
	}
	literal := map[string]bool{"/internal/v1/groups/slug-suggestion": true, "/internal/v1/invitations/mine": true,
		"/internal/v1/invitations/candidates": true}
	for _, f := range []string{"../../onboarding/api/http.go", "../../membership/api/http.go"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range regexp.MustCompile(`(?m)^\s+Route[A-Za-z]+\s+= "([^"]+)"$`).FindAllSubmatch(b, -1) {
			add(string(m[1]))
			if _, p, _ := strings.Cut(string(m[1]), " "); !literal[p] {
				add(p)
			}
		}
		for _, m := range regexp.MustCompile(`"(/internal/v1/[a-z-]+/)"`).FindAllSubmatch(b, -1) {
			add(string(m[1]))
		}
	}
	if !seen["/internal/v1/groups/"] || !seen["GET /internal/v1/groups/{id}"] || !seen["GET /internal/v1/members"] {
		t.Fatalf("other adapters' patterns not found: %d", len(seen))
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("labels patterns conflict: %v", r)
			}
		}()
		handler(&fakeDB{}, slog.New(slog.DiscardHandler)).Register(mux)
	}()
	// A Group sub-path that is not a labels route still reaches onboarding's root, not a labels handler.
	if _, p := mux.Handler(httptest.NewRequest("GET", "/internal/v1/groups/"+gid+"/labels/x", nil)); p != "/internal/v1/groups/" {
		t.Errorf("Group labels sub-path matched %q", p)
	}
	if _, p := mux.Handler(httptest.NewRequest("GET", "/internal/v1/groups/"+gid+"/labels", nil)); p != RouteListGroupLabels {
		t.Errorf("Group labels read matched %q", p)
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

// TestLabelsCopyIsExact: the session-scope behaviour and the shared request helpers are a bounded copy of
// membership/api (itself pinned to onboarding/api), not a shared helper; this pins them byte for byte.
func TestLabelsCopyIsExact(t *testing.T) {
	for _, name := range []string{"begin", "scope", "originHost", "query", "hostTag", "writeJSON", "recorder", "fail", "refuse",
		"closed", "failure", "refuseErr", "page", "parseID", "formatID", "setKind", "scopeOf", "queryScope", "read", "shape",
		"mutation"} {
		if a, b := decl(t, "../../membership/api/http.go", name), decl(t, "http.go", name); a != b {
			t.Errorf("%s differs from membership/api:\n--- membership\n%s\n--- labels\n%s", name, a, b)
		}
	}
	if _, err := New(Config{}); err == nil {
		t.Error("an incomplete configuration was accepted")
	}
}
