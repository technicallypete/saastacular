package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/objectstore"
	"core/internal/tenancy/access"
)

// G017 (Plan r2 §5.3, §5.5): the branding API's anonymous routes (public projection, logo bytes) over a fake querier
// (the G016 hostname_state resolver statement and the app.public_branding projection) and a fake object store; the
// logo route's pre-read bounds; the closed refusal mapping; and the D-5 bounded copy of domains/api (the shared refusal
// codes and statuses, and begin/scope/originHost byte for byte). The authenticated routes run against real PostgreSQL
// in TestBranding (the facade) and the browser suites.

const (
	orgA  = "0f17a000-0000-4000-8000-0000000000a1"
	orgC  = "0f17c000-0000-4000-8000-0000000000c1"
	objA  = "0f17a000-0000-4000-8000-00000000f001"
	hostA = "https://org-a.localhost:543"
	hostC = "https://portal.acme.test:543"
)

// fakeDB is the Service's DB: no transaction (the anonymous routes never open one) and single statements routed by
// their text to the resolver's hostname_state rows (kind|key) or the public_branding rows (Organization id).
type fakeDB struct {
	hosts    map[string]*string
	branding map[string]brandingRow
	err      error
	stmts    []string
}

type brandingRow struct {
	revision    int64
	label       *string
	light, dark map[string]string
	logo        []byte
	contentType string
	key         string
}

type row struct {
	scan func(dest ...any) error
}

func (r row) Scan(dest ...any) error { return r.scan(dest...) }

func noRow(...any) error { return pgx.ErrNoRows }

func (f *fakeDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("fakeDB: no transaction expected")
}

func (f *fakeDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	switch {
	case f.err != nil:
		return row{func(...any) error { return f.err }}
	case strings.Contains(sql, "hostname_state"):
		f.stmts = append(f.stmts, "hostname_state")
		org, ok := f.hosts[args[0].(string)+"|"+args[1].(string)]
		if !ok {
			return row{noRow}
		}
		return row{func(dest ...any) error {
			*dest[0].(**string), *dest[1].(*bool), *dest[2].(*bool) = org, true, false
			return nil
		}}
	case strings.Contains(sql, "app.public_branding($1)"):
		f.stmts = append(f.stmts, "public_branding")
		id, err := uuidString(args[0])
		if err != nil {
			return row{func(...any) error { return err }}
		}
		b, ok := f.branding[id]
		if !ok {
			return row{noRow}
		}
		return row{func(dest ...any) error {
			*dest[0].(*int64), *dest[1].(**string) = b.revision, b.label
			*dest[2].(*map[string]string), *dest[3].(*map[string]string) = b.light, b.dark
			if b.logo != nil {
				sum := sha256.Sum256(b.logo)
				size := int64(len(b.logo))
				*dest[4].(*[]byte), *dest[5].(**string), *dest[6].(**int64), *dest[7].(**string) = sum[:], &b.contentType, &size, &b.key
			}
			return nil
		}}
	}
	f.stmts = append(f.stmts, "unexpected")
	return row{func(...any) error { return errors.New("fakeDB: unexpected statement") }}
}

// uuidString renders the projection's argument (a pgtype.UUID-like value) as its canonical text.
func uuidString(v any) (string, error) {
	vv, ok := v.(driver.Valuer)
	if !ok {
		return "", errors.New("fakeDB: not a uuid argument")
	}
	s, err := vv.Value()
	if err != nil {
		return "", err
	}
	str, ok := s.(string)
	if !ok {
		return "", errors.New("fakeDB: not a uuid value")
	}
	return str, nil
}

// fakeObjects is the owned store: Get returns the stored bytes and binding; everything else is unexpected.
type fakeObjects struct {
	data    map[string][]byte
	binding map[string]objectstore.Binding
	gets    int
}

func (o *fakeObjects) Get(_ context.Context, key objectstore.Key) (objectstore.Info, io.ReadCloser, error) {
	o.gets++
	d, ok := o.data[key.String()]
	if !ok {
		return objectstore.Info{}, nil, objectstore.ErrNotFound
	}
	return objectstore.Info{Size: int64(len(d)), Binding: o.binding[key.String()]}, io.NopCloser(bytes.NewReader(d)), nil
}

func (o *fakeObjects) Put(context.Context, objectstore.Key, io.Reader, int64, string, objectstore.Binding) error {
	return errors.New("fakeObjects: put")
}
func (o *fakeObjects) Delete(context.Context, objectstore.Key) error {
	return errors.New("fakeObjects: delete")
}
func (o *fakeObjects) Metadata(context.Context, objectstore.Key) (objectstore.Info, error) {
	return objectstore.Info{}, errors.New("fakeObjects: metadata")
}
func (o *fakeObjects) SignedGetURL(context.Context, objectstore.Key, time.Duration, string) (string, error) {
	return "", errors.New("fakeObjects: sign")
}

func strp(s string) *string { return &s }

// fixture: org-a.localhost routes to A (label, overrides, a PNG logo); portal.acme.test routes to C (no branding row).
func fixture(t *testing.T) (*Handler, *fakeDB, *fakeObjects, []byte) {
	t.Helper()
	logo := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{7}, 64)...)
	key, err := objectstore.NewKey(orgA, objA)
	if err != nil {
		t.Fatal(err)
	}
	db := &fakeDB{
		hosts: map[string]*string{"platform|org-a": strp(orgA), "custom|portal.acme.test": strp(orgC)},
		branding: map[string]brandingRow{orgA: {revision: 4, label: strp("Acme"), light: map[string]string{"primary": "#123456"},
			dark: map[string]string{}, logo: logo, contentType: "image/png", key: key.String()}},
	}
	objs := &fakeObjects{data: map[string][]byte{key.String(): logo},
		binding: map[string]objectstore.Binding{key.String(): {Object: objA, Organization: orgA}}}
	d := config.Domains{BaseDomain: "localhost", CNAMETarget: "edge.localhost", HostPort: "543", AuthHost: "auth.localhost"}
	acc := access.NewServiceWithObjects(db, nil, nil, objs, access.ObjectConfig{MaxObjectBytes: 1 << 20,
		DefaultTTL: time.Minute, MaxTTL: 5 * time.Minute, TransferTimeout: 5 * time.Second})
	h := &Handler{cfg: Config{Resolver: domains.NewResolver(db, d.Hostnames())}, acc: acc, log: slog.New(slog.DiscardHandler)}
	return h, db, objs, logo
}

func get(t *testing.T, h http.HandlerFunc, target string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
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

// noStoreJSON asserts the headers of every answer except a current logo, and that no CORS header is ever set.
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

func publicURL(origin string) string {
	return "/internal/v1/branding/public?" + url.Values{"origin": {origin}}.Encode()
}

func logoURL(origin, v string) string {
	return "/internal/v1/branding/logo?" + url.Values{"origin": {origin}, "v": {v}}.Encode()
}

func TestPublicRoute(t *testing.T) {
	h, db, objs, logo := fixture(t)
	sum := sha256.Sum256(logo)
	v := hex.EncodeToString(sum[:6])

	w := get(t, h.Public, publicURL(hostA))
	noStoreJSON(t, "A", w)
	out := decode(t, w)
	if w.Code != 200 || keys(out) != "dark,display_label,light,logo,revision" {
		t.Fatalf("A: %d %v", w.Code, out)
	}
	logoOut, _ := out["logo"].(map[string]any)
	if out["revision"] != float64(4) || out["display_label"] != "Acme" || keys(logoOut) != "content_type,v" ||
		logoOut["v"] != v || logoOut["content_type"] != "image/png" {
		t.Errorf("A: projection %v", out)
	}
	if light, _ := out["light"].(map[string]any); len(light) != 1 || light["primary"] != "#123456" {
		t.Errorf("A: light %v", out["light"])
	}
	if objs.gets != 0 {
		t.Errorf("the projection read the object store (%d gets)", objs.gets)
	}

	// No branding row: the platform defaults at revision 0, without a label or a logo.
	w = get(t, h.Public, publicURL(hostC))
	noStoreJSON(t, "C", w)
	if out := decode(t, w); w.Code != 200 || keys(out) != "dark,light,revision" || out["revision"] != float64(0) {
		t.Errorf("C: %d %v", w.Code, out)
	}

	for _, c := range []struct {
		name, target string
		code         int
		want         string
	}{
		{"unknown host", publicURL("https://nobody.localhost:543"), 404, "unknown_host"},
		{"not an exact origin", publicURL("https://org-a.localhost:543/"), 404, "unknown_host"},
		{"http origin", publicURL("http://org-a.localhost:543"), 404, "unknown_host"},
		{"missing origin", "/internal/v1/branding/public", 400, "ambiguous_parameter"},
		{"repeated origin", "/internal/v1/branding/public?origin=" + url.QueryEscape(hostA) + "&origin=" + url.QueryEscape(hostA), 400, "ambiguous_parameter"},
	} {
		w := get(t, h.Public, c.target)
		noStoreJSON(t, c.name, w)
		if out := decode(t, w); w.Code != c.code || keys(out) != "error" || out["error"] != c.want {
			t.Errorf("%s: %d %v, want %d %s", c.name, w.Code, out, c.code, c.want)
		}
	}

	// Only the two reviewed statements ran; a store failure is unavailable, never a clean unknown host.
	for _, s := range db.stmts {
		if s != "hostname_state" && s != "public_branding" {
			t.Errorf("statement %s", s)
		}
	}
	db.err = errors.New("store down")
	w = get(t, h.Public, publicURL(hostA))
	noStoreJSON(t, "store down", w)
	if out := decode(t, w); w.Code != 503 || out["error"] != "branding_unavailable" {
		t.Errorf("store failure: %d %v", w.Code, out)
	}
}

func TestLogoRoute(t *testing.T) {
	h, db, objs, logo := fixture(t)
	sum := sha256.Sum256(logo)
	v := hex.EncodeToString(sum[:6])

	w := get(t, h.Logo, logoURL(hostA, v))
	hd := w.Header()
	if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), logo) {
		t.Fatalf("current logo: %d (%d bytes)", w.Code, w.Body.Len())
	}
	for k, want := range map[string]string{"Content-Type": "image/png", "Content-Length": strconv.Itoa(len(logo)),
		"Cache-Control": "public, max-age=86400, immutable", "ETag": `"` + v + `"`, "X-Content-Type-Options": "nosniff",
		"Content-Disposition": "inline"} {
		if hd.Get(k) != want {
			t.Errorf("logo header %s = %q, want %q", k, hd.Get(k), want)
		}
	}
	if len(hd) != 6 {
		t.Errorf("logo headers %v", hd)
	}

	// A stale or foreign v, a host routing to an Organization without that logo, or a host routing nowhere: a no-store
	// 404 with no bytes read.
	gets := objs.gets
	for _, c := range []struct {
		name, target, want string
	}{
		{"stale v", logoURL(hostA, "000000000000"), "not_found"},
		{"short v", logoURL(hostA, v[:11]), "not_found"},
		{"uppercase v", logoURL(hostA, strings.ToUpper(v)), "not_found"},
		{"foreign host", logoURL(hostC, v), "not_found"},
		{"unknown host", logoURL("https://nobody.localhost:543", v), "unknown_host"},
	} {
		w := get(t, h.Logo, c.target)
		noStoreJSON(t, c.name, w)
		if out := decode(t, w); w.Code != 404 || out["error"] != c.want {
			t.Errorf("%s: %d %v, want 404 %s", c.name, w.Code, out, c.want)
		}
	}
	if objs.gets != gets {
		t.Errorf("a refused logo request read the object store")
	}
	for _, target := range []string{"/internal/v1/branding/logo?origin=" + url.QueryEscape(hostA),
		logoURL(hostA, v) + "&v=" + v} {
		w := get(t, h.Logo, target)
		noStoreJSON(t, target, w)
		if out := decode(t, w); w.Code != 400 || out["error"] != "ambiguous_parameter" {
			t.Errorf("%s: %d %v", target, w.Code, out)
		}
	}

	// Bytes that no longer match the projection (length, digest or binding) are never served.
	key, _ := objectstore.NewKey(orgA, objA)
	objs.data[key.String()] = append(bytes.Clone(logo[:len(logo)-1]), 8)
	w = get(t, h.Logo, logoURL(hostA, v))
	noStoreJSON(t, "tampered", w)
	if out := decode(t, w); w.Code != 503 || out["error"] != "branding_unavailable" {
		t.Errorf("tampered bytes: %d %v", w.Code, out)
	}
	objs.data[key.String()] = logo
	objs.binding[key.String()] = objectstore.Binding{Object: objA, Organization: orgC}
	if w := get(t, h.Logo, logoURL(hostA, v)); w.Code != 503 {
		t.Errorf("foreign binding: %d", w.Code)
	}
	delete(objs.data, key.String())
	if w := get(t, h.Logo, logoURL(hostA, v)); w.Code != 404 || w.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("missing bytes: %d", w.Code)
	}
	db.err = errors.New("store down")
	if w := get(t, h.Logo, logoURL(hostA, v)); w.Code != 503 || w.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("store failure: %d", w.Code)
	}
}

// The logo upload's bounds hold before any body read, session or store access.
func TestSetLogoPreReadBounds(t *testing.T) {
	h, db, _, _ := fixture(t)
	post := func(origin, csrf, version string, length int64, body io.Reader) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/internal/v1/branding/logo?"+url.Values{"origin": {origin}}.Encode(), body)
		if csrf != "" {
			r.Header.Set("X-CSRF-Token", csrf)
		}
		if version != "" {
			r.Header.Set("X-Expected-Version", version)
		}
		r.Header.Set("Authorization", "Session x")
		r.ContentLength = length
		w := httptest.NewRecorder()
		h.SetLogo(w, r)
		return w
	}
	read := func(n int64) *countingReader { return &countingReader{n: n} }
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
		{"missing csrf", hostA, "", "0", 10, 400, "ambiguous_parameter"},
		{"missing version", hostA, "c", "", 10, 400, "ambiguous_parameter"},
		// At the bound the request proceeds to the session scope: an origin that routes nowhere is rejected there.
		{"at the bound", "https://nobody.localhost:543", "c", "0", 524288, 403, "origin_rejected"},
	} {
		body := read(c.length)
		w := post(c.origin, c.csrf, c.version, c.length, body)
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

type countingReader struct{ n, read int64 }

func (c *countingReader) Read(p []byte) (int, error) {
	c.read += int64(len(p))
	return 0, io.EOF
}

func TestRefusalMapping(t *testing.T) {
	for code, want := range map[access.Code]string{
		access.CodeUnauthenticated: "unauthenticated", access.CodeContextDenied: "not_available", access.CodeForbidden: "forbidden",
		access.CodeNotFound: "not_found", access.CodeConflict: "version_conflict", access.CodeContrastFailed: "contrast_failed",
		access.CodeUnavailable: "branding_unavailable", access.CodeInternal: "branding_unavailable",
	} {
		if got := failure(&access.Error{Code: code}); got != want {
			t.Errorf("%s -> %s, want %s", code, got, want)
		}
	}
	for field, want := range map[string]string{"logo_too_large": "payload_too_large", "logo_type": "logo_invalid",
		"logo_dimensions": "logo_invalid", "logo_malformed": "logo_invalid", "normalized_too_large": "logo_invalid",
		"display_label": "invalid_request"} {
		if got := failure(&access.Error{Code: access.CodeInvalidInput, Field: field}); got != want {
			t.Errorf("invalid_input/%s -> %s, want %s", field, got, want)
		}
	}
	if got := failure(&access.Error{Code: access.CodeUnavailable, Field: "publication_outcome_unknown"}); got != "publication_outcome_unknown" {
		t.Errorf("unknown publication outcome -> %s", got)
	}
	if got := failure(errors.New("other")); got != "branding_unavailable" {
		t.Errorf("a non-access error -> %s", got)
	}
	if len(status) != 16 {
		t.Errorf("%d refusal codes, want 16", len(status))
	}

	// The upload outcome detail carries the object and its live state, and only for those four fields.
	w := httptest.NewRecorder()
	(&Handler{}).refuseErr(w, &access.Error{Code: access.CodeUnavailable, Field: "publication_outcome_unknown", Object: objA, ObjectState: "ready"})
	if out := decode(t, w); w.Code != 503 || keys(out) != "error,upload" || keys(out["upload"].(map[string]any)) != "object_id,object_state,outcome" {
		t.Errorf("upload outcome detail: %d %v", w.Code, out)
	}
	w = httptest.NewRecorder()
	(&Handler{}).refuseErr(w, &access.Error{Code: access.CodeUnavailable, Field: "store_unavailable"})
	if out := decode(t, w); w.Code != 503 || keys(out) != "error" {
		t.Errorf("unavailable detail: %d %v", w.Code, out)
	}
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

// D-5: the session-scope behaviour is a bounded copy of domains/api, not a shared helper; this pins the copy.
func TestDomainsCopyIsExact(t *testing.T) {
	dom, brd := parseSource(t, "../../domains/api/http.go"), parseSource(t, "http.go")
	for _, name := range []string{"begin", "scope", "originHost"} {
		if a, b := dom.decl(t, name), brd.decl(t, name); a != b {
			t.Errorf("%s differs from domains/api:\n--- domains\n%s\n--- branding\n%s", name, a, b)
		}
	}
	shared := []string{"errInvalidRequest", "errAmbiguous", "errUnsupported", "errOriginRejected", "errUnauthenticated", "errCSRF",
		"errNotAvailable", "errForbidden", "errNotFound", "errUnknownHost"}
	dc, bc := dom.consts(), brd.consts()
	ds, bs := dom.statuses(t), brd.statuses(t)
	for _, n := range shared {
		if dc[n] == "" || dc[n] != bc[n] || ds[n] == "" || ds[n] != bs[n] {
			t.Errorf("%s: domains %s/%s, branding %s/%s", n, dc[n], ds[n], bc[n], bs[n])
		}
	}
	if ds["errUnavailable"] != "503" || bs["errUnavailable"] != "503" || bc["errUnavailable"] != `"branding_unavailable"` {
		t.Errorf("errUnavailable: domains %s, branding %s %s", ds["errUnavailable"], bc["errUnavailable"], bs["errUnavailable"])
	}
	// Every branding code is in its status table, and the copied codes keep their wire values in this package.
	for n, v := range bc {
		if strings.HasPrefix(n, "err") && bs[n] == "" {
			t.Errorf("%s (%s) has no status", n, v)
		}
	}
	if len(bs) != len(status) {
		t.Errorf("parsed %d statuses, the table has %d", len(bs), len(status))
	}
}

func TestNewRefusesIncompleteConfiguration(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("an incomplete configuration was accepted")
	}
	if hostTag("portal.acme.test") == hostTag("portal.acme.tes") || len(hostTag("x")) != 12 {
		t.Error("host tag")
	}
}
