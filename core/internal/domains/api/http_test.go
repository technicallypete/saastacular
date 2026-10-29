package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/tenancy/access"
)

// G016 (Plan r4 §5.3): the domains API's unauthenticated routes (hosts/resolve, tls-ask) over a fake resolver store,
// the exact-origin rule, and the closed refusal mapping. The authenticated routes run against real PostgreSQL in
// TestDomains (the facade) and the identity/browser suites.

// fakeStore answers app.hostname_state by (kind, key); a nil entry is "no row"; err fails every statement.
type fakeStore struct {
	rows  map[string]fakeRow
	err   error
	calls int
}

type fakeRow struct {
	org               *string
	routing, issuance bool
	none              bool
	err               error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if r.none {
		return pgx.ErrNoRows
	}
	*dest[0].(**string) = r.org
	*dest[1].(*bool) = r.routing
	*dest[2].(*bool) = r.issuance
	return nil
}

func (f *fakeStore) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	f.calls++
	if f.err != nil {
		return fakeRow{err: f.err}
	}
	if r, ok := f.rows[args[0].(string)+"|"+args[1].(string)]; ok {
		return r
	}
	return fakeRow{none: true}
}

func strp(s string) *string { return &s }

const (
	orgA  = "0f16a000-0000-4000-8000-0000000000a1"
	orgC  = "0f16c000-0000-4000-8000-0000000000c1"
	orgDT = "0f071000-0000-4000-8000-0000000000d1" // G071: the pinned default tenant
)

func testHandler(store *fakeStore) *Handler {
	d := config.Domains{BaseDomain: "localhost", CNAMETarget: "edge.localhost", HostPort: "543", AuthHost: "auth.localhost"}
	return &Handler{cfg: Config{Resolver: domains.NewResolver(store, d.Hostnames()), Domains: d}}
}

func get(t *testing.T, h http.HandlerFunc, target string) (int, http.Header, map[string]any) {
	t.Helper()
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest(http.MethodGet, target, nil))
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, w.Header(), out
}

func TestResolveRoute(t *testing.T) {
	store := &fakeStore{rows: map[string]fakeRow{
		"platform|org-a":            {org: strp(orgA), routing: true},
		"custom|portal.acme.test":   {org: strp(orgC), routing: true, issuance: true},
		"custom|pending.acme.test":  {org: strp(orgC), routing: false, issuance: false},
		"custom|verified.acme.test": {org: strp(orgC), routing: false, issuance: true},
	}}
	h := testHandler(store)
	for _, c := range []struct {
		origin string
		code   int
		want   map[string]any
	}{
		{"https://org-a.localhost:543", 200, map[string]any{"kind": "platform", "organization_id": orgA}},
		{"https://portal.acme.test:543", 200, map[string]any{"kind": "custom", "organization_id": orgC,
			"callback_origin": "https://auth.localhost:543"}},
		{"https://pending.acme.test:543", 404, map[string]any{"error": "unknown_host"}},
		{"https://verified.acme.test:543", 404, map[string]any{"error": "unknown_host"}},
		{"https://nobody.localhost:543", 404, map[string]any{"error": "unknown_host"}},
		{"https://unknown.example:543", 404, map[string]any{"error": "unknown_host"}},
		// G062: both public entry names answer the exact root origin as signup_origin and never a location. G071 S16:
		// both carry root_page (this handler's mode is unset, which is signup_redirect).
		{"https://localhost:543", 200, map[string]any{"kind": "public_entry", "signup_origin": "https://localhost:543",
			"root_page": "signup_redirect"}},
		{"https://www.localhost:543", 200, map[string]any{"kind": "public_entry", "signup_origin": "https://localhost:543",
			"root_page": "signup_redirect"}},
		{"http://localhost:543", 404, map[string]any{"error": "unknown_host"}},
		{"https://WWW.localhost:543", 404, map[string]any{"error": "unknown_host"}},
		{"https://auth.localhost:543", 200, map[string]any{"kind": "auth"}},
		// Only an exact (already normalized) origin is answered.
		{"https://ORG-A.localhost:543", 404, map[string]any{"error": "unknown_host"}},
		{"https://org-a.localhost.:543", 404, map[string]any{"error": "unknown_host"}},
		{"http://org-a.localhost:543", 404, map[string]any{"error": "unknown_host"}},
		{"https://user@org-a.localhost:543", 404, map[string]any{"error": "unknown_host"}},
		{"https://org-a.localhost:543/path", 404, map[string]any{"error": "unknown_host"}},
	} {
		code, hd, out := get(t, h.Resolve, "/internal/v1/hosts/resolve?"+url.Values{"origin": {c.origin}}.Encode())
		if code != c.code || len(out) != len(c.want) {
			t.Errorf("%s: %d %v, want %d %v", c.origin, code, out, c.code, c.want)
			continue
		}
		for k, v := range c.want {
			if out[k] != v {
				t.Errorf("%s: %s = %v, want %v", c.origin, k, out[k], v)
			}
		}
		if hd.Get("Cache-Control") != "no-store" || hd.Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("%s: headers %v", c.origin, hd)
		}
	}
	if code, _, out := get(t, h.Resolve, "/internal/v1/hosts/resolve?origin=https://org-a.localhost:543&origin=https://org-a.localhost:543"); code != 400 || out["error"] != "ambiguous_parameter" {
		t.Errorf("repeated origin: %d %v", code, out)
	}
	store.err = errors.New("store down")
	if code, _, out := get(t, h.Resolve, "/internal/v1/hosts/resolve?origin=https://org-a.localhost:543"); code != 503 || out["error"] != "domains_unavailable" {
		t.Errorf("store failure: %d %v, want 503 domains_unavailable (never a clean unknown host)", code, out)
	}

	// G071 (Plan r4 S4, S5; §5.1 U3 alias cells): with a default tenant pinned, the exact root answers the platform kind
	// with root_alias and without a statement; every other host answers exactly as unpinned and never carries
	// root_alias; only an exact origin is answered.
	pstore := &fakeStore{rows: map[string]fakeRow{"platform|org-a": {org: strp(orgA), routing: true}}}
	ph := testHandler(pstore)
	ph.cfg.Resolver.WithDefaultTenant(orgDT)
	for _, c := range []struct {
		origin string
		code   int
		want   map[string]any
		calls  int
	}{
		{"https://localhost:543", 200, map[string]any{"kind": "platform", "organization_id": orgDT, "root_alias": true}, 0},
		{"https://www.localhost:543", 200, map[string]any{"kind": "public_entry", "signup_origin": "https://localhost:543",
			"root_page": "signup_redirect"}, 0},
		{"https://org-a.localhost:543", 200, map[string]any{"kind": "platform", "organization_id": orgA}, 1},
		{"https://auth.localhost:543", 200, map[string]any{"kind": "auth"}, 0},
		{"https://edge.localhost:543", 200, map[string]any{"kind": "reserved"}, 0},
		{"https://nobody.localhost:543", 404, map[string]any{"error": "unknown_host"}, 1},
		{"https://unknown.example:543", 404, map[string]any{"error": "unknown_host"}, 1},
		{"https://a.b.localhost:543", 404, map[string]any{"error": "unknown_host"}, 0},
		{"https://LOCALHOST:543", 404, map[string]any{"error": "unknown_host"}, 0},
		{"https://localhost.:543", 404, map[string]any{"error": "unknown_host"}, 0},
		{"http://localhost:543", 404, map[string]any{"error": "unknown_host"}, 0},
		{"https://localhost:543/signup", 404, map[string]any{"error": "unknown_host"}, 0},
	} {
		pstore.calls = 0
		code, hd, out := get(t, ph.Resolve, "/internal/v1/hosts/resolve?"+url.Values{"origin": {c.origin}}.Encode())
		if code != c.code || len(out) != len(c.want) || pstore.calls != c.calls {
			t.Errorf("pinned %s: %d %v (%d statements), want %d %v (%d)", c.origin, code, out, pstore.calls, c.code, c.want,
				c.calls)
			continue
		}
		for k, v := range c.want {
			if out[k] != v {
				t.Errorf("pinned %s: %s = %v, want %v", c.origin, k, out[k], v)
			}
		}
		if hd.Get("Cache-Control") != "no-store" || hd.Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("pinned %s: headers %v", c.origin, hd)
		}
	}
	// The pinned root needs no statement, so a store outage still answers it; a tenant host stays 503.
	pstore.err, pstore.calls = errors.New("store down"), 0
	if code, _, out := get(t, ph.Resolve, "/internal/v1/hosts/resolve?origin=https://localhost:543"); code != 200 ||
		out["kind"] != "platform" || out["organization_id"] != orgDT || out["root_alias"] != true || pstore.calls != 0 {
		t.Errorf("pinned root with the store down: %d %v (%d statements)", code, out, pstore.calls)
	}
	if code, _, out := get(t, ph.Resolve, "/internal/v1/hosts/resolve?origin=https://org-a.localhost:543"); code != 503 || out["error"] != "domains_unavailable" {
		t.Errorf("pinned resolver, tenant host with the store down: %d %v", code, out)
	}
	// Unpinned (the handler above), the root never carries root_alias.
	store.err = nil
	if _, _, out := get(t, h.Resolve, "/internal/v1/hosts/resolve?origin=https://localhost:543"); out["kind"] != "public_entry" || out["root_alias"] != nil {
		t.Errorf("unpinned root: %v", out)
	}

	// G071 (Plan r4 S16; §5.1 U3 precedence cells of the §5.3 matrix): alias state × PLATFORM_ROOT_PAGE. Every
	// public_entry answer carries root_page with the effective mode and nothing else (keys === 3); a pinned root wins
	// for every mode (platform + root_alias, no root_page, no statement); an inactive selection (unknown or invalid slug:
	// SelectDefaultTenant pins "", as unset does) leaves the root a public_entry answering the selected mode; www carries
	// the same member in every cell; every other answer is unchanged and never carries it. The server hands the handler
	// the effective mode (an invalid value is already signup_redirect, U2); unset is signup_redirect.
	for _, m := range []struct{ page, want string }{
		{domains.RootPageSignupRedirect, "signup_redirect"},
		{domains.RootPageLanding, "landing"},
		{"", "signup_redirect"},
	} {
		for _, pin := range []string{orgDT, ""} {
			mstore := &fakeStore{rows: map[string]fakeRow{
				"platform|org-a":          {org: strp(orgA), routing: true},
				"custom|portal.acme.test": {org: strp(orgC), routing: true, issuance: true},
			}}
			mh := testHandler(mstore)
			mh.cfg.Domains.RootPage = m.page
			mh.cfg.Resolver.WithDefaultTenant(pin)
			entry := map[string]any{"kind": "public_entry", "signup_origin": "https://localhost:543", "root_page": m.want}
			root := entry
			if pin != "" {
				root = map[string]any{"kind": "platform", "organization_id": orgDT, "root_alias": true}
			}
			check := func(origin string, code int, want map[string]any, calls int) {
				t.Helper()
				mstore.calls = 0
				gotCode, hd, out := get(t, mh.Resolve, "/internal/v1/hosts/resolve?"+url.Values{"origin": {origin}}.Encode())
				if gotCode != code || len(out) != len(want) || mstore.calls != calls {
					t.Errorf("mode %q pin %q %s: %d %v (%d statements), want %d %v (%d)", m.page, pin, origin, gotCode, out,
						mstore.calls, code, want, calls)
					return
				}
				for k, v := range want {
					if out[k] != v {
						t.Errorf("mode %q pin %q %s: %s = %v, want %v", m.page, pin, origin, k, out[k], v)
					}
				}
				if hd.Get("Cache-Control") != "no-store" || hd.Get("X-Content-Type-Options") != "nosniff" {
					t.Errorf("mode %q pin %q %s: headers %v", m.page, pin, origin, hd)
				}
			}
			check("https://localhost:543", 200, root, 0)
			check("https://www.localhost:543", 200, entry, 0)
			check("https://org-a.localhost:543", 200, map[string]any{"kind": "platform", "organization_id": orgA}, 1)
			check("https://portal.acme.test:543", 200, map[string]any{"kind": "custom", "organization_id": orgC,
				"callback_origin": "https://auth.localhost:543"}, 1)
			check("https://auth.localhost:543", 200, map[string]any{"kind": "auth"}, 0)
			check("https://nobody.localhost:543", 404, map[string]any{"error": "unknown_host"}, 1)
			check("https://localhost:543/signup", 404, map[string]any{"error": "unknown_host"}, 0)
			// A store outage never touches the root or www (no statement either way); a tenant host stays 503.
			mstore.err = errors.New("store down")
			check("https://localhost:543", 200, root, 0)
			check("https://www.localhost:543", 200, entry, 0)
			check("https://org-a.localhost:543", 503, map[string]any{"error": "domains_unavailable"}, 1)
		}
	}
}

func TestTLSAskRoute(t *testing.T) {
	store := &fakeStore{rows: map[string]fakeRow{
		"custom|portal.acme.test":    {org: strp(orgC), routing: true, issuance: true},
		"custom|pending.acme.test":   {org: strp(orgC)},
		"custom|suspended.acme.test": {org: strp(orgC), issuance: true},
		"platform|org-a":             {org: strp(orgA), routing: true},
	}}
	h := testHandler(store)
	for _, c := range []struct {
		q    string
		code int
	}{
		{"domain=portal.acme.test", 200},
		{"domain=suspended.acme.test", 200},
		{"domain=pending.acme.test", 403},
		{"domain=unknown.acme.test", 403},
		{"domain=org-a.localhost", 403},
		{"domain=portal.acme.test:543", 403},
		{"domain=PORTAL.acme.test", 403},
		{"", 403},
		{"domain=portal.acme.test&domain=portal.acme.test", 403},
	} {
		w := httptest.NewRecorder()
		h.TLSAsk(w, httptest.NewRequest(http.MethodGet, "/internal/v1/domains/tls-ask?"+c.q, nil))
		if w.Code != c.code || w.Body.Len() != 0 || w.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("%q: %d (body %d bytes)", c.q, w.Code, w.Body.Len())
		}
	}
	store.err = errors.New("store down")
	w := httptest.NewRecorder()
	h.TLSAsk(w, httptest.NewRequest(http.MethodGet, "/internal/v1/domains/tls-ask?domain=portal.acme.test", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("store failure: %d, want 503", w.Code)
	}

	// G071 (U3): the aliased root is never issuable (the platform wildcard covers it; issuance stays custom-only), and
	// neither alias state nor the root presentation mode (S16, §5.3) changes that for the root or www.
	store.err = nil
	for _, page := range []string{domains.RootPageSignupRedirect, domains.RootPageLanding, ""} {
		for _, pin := range []string{orgDT, ""} {
			ph := testHandler(store)
			ph.cfg.Domains.RootPage = page
			ph.cfg.Resolver.WithDefaultTenant(pin)
			for _, q := range []string{"domain=localhost", "domain=localhost:543", "domain=www.localhost"} {
				store.calls = 0
				w := httptest.NewRecorder()
				ph.TLSAsk(w, httptest.NewRequest(http.MethodGet, "/internal/v1/domains/tls-ask?"+q, nil))
				if w.Code != http.StatusForbidden || w.Body.Len() != 0 || store.calls != 0 {
					t.Errorf("mode %q pin %q %q: %d (%d statements), want 403 without a statement", page, pin, q, w.Code,
						store.calls)
				}
			}
			// A custom host's issuance is unchanged by either setting.
			w := httptest.NewRecorder()
			ph.TLSAsk(w, httptest.NewRequest(http.MethodGet, "/internal/v1/domains/tls-ask?domain=portal.acme.test", nil))
			if w.Code != http.StatusOK {
				t.Errorf("mode %q pin %q portal.acme.test: %d, want 200", page, pin, w.Code)
			}
		}
	}
}

func TestRefusalMapping(t *testing.T) {
	for code, want := range map[access.Code]string{
		access.CodeUnauthenticated: "unauthenticated", access.CodeContextDenied: "not_available", access.CodeForbidden: "forbidden",
		access.CodeNotFound: "not_found", access.CodeHostTaken: "host_taken", access.CodeConflict: "slug_taken",
		access.CodeDomainStateInvalid: "domain_state_invalid", access.CodeDomainRetryExhausted: "domain_retry_exhausted",
		access.CodeUnavailable: "domains_unavailable", access.CodeInternal: "domains_unavailable",
	} {
		if got := failure(&access.Error{Code: code}); got != want {
			t.Errorf("%s -> %s, want %s", code, got, want)
		}
	}
	for field, want := range map[string]string{"host_invalid": "host_invalid", "slug_reserved": "slug_reserved", "invalid_slug": "invalid_request"} {
		if got := failure(&access.Error{Code: access.CodeInvalidInput, Field: field}); got != want {
			t.Errorf("invalid_input/%s -> %s, want %s", field, got, want)
		}
	}
	if got := failure(errors.New("other")); got != "domains_unavailable" {
		t.Errorf("a non-access error -> %s", got)
	}
	for code, st := range status {
		if st < 400 || st > 503 {
			t.Errorf("%s: status %d", code, st)
		}
	}
	if len(status) != 18 {
		t.Errorf("%d refusal codes, want 18", len(status))
	}
}

func TestOriginHostIsExact(t *testing.T) {
	for origin, want := range map[string]string{"https://org-a.localhost:543": "org-a.localhost:543", "https://portal.acme.test": "portal.acme.test"} {
		if got, ok := originHost(origin); !ok || got != want {
			t.Errorf("%s: %q %v", origin, got, ok)
		}
	}
	for _, bad := range []string{"", "http://a.test", "https://", "https://u@a.test", "https://a.test/", "https://a.test?x", "https://a.test#f",
		"a.test", "https://a.test:543/mcp"} {
		if _, ok := originHost(bad); ok {
			t.Errorf("%q accepted", bad)
		}
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
