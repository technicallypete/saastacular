package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/time/rate"

	"core/internal/identity"
	"core/internal/identity/workos"
	"core/internal/tenancy/access"
)

// U-MCP-HOST, U-MCP-URL, U-MCP-META, U-MCP-TOOLS, U-MCP-ERR and U-MCP-BOUNDARY (G014 Plan r8 §10.1): everything the
// transport decides before any store, and the package boundary.

const (
	testHostA = "org-a.localhost:543"
	testHostB = "org-b.localhost:543"
	testOrgA  = "4a000000-0000-4000-8000-00000000000a"
	testOrgB  = "4b000000-0000-4000-8000-00000000000b"
	testID    = "7f000000-0000-4000-8000-000000000001"
	issuer    = "https://example-env.authkit.app"
)

// staticLookup is the test double of the durable resolver (G016): an exact host[:port] -> Organization map; fail
// simulates a store failure.
type staticLookup struct {
	hosts map[string]string
	fail  error
}

func (s staticLookup) Lookup(_ context.Context, host string) (string, string, error) {
	if s.fail != nil {
		return "", "", s.fail
	}
	if o, ok := s.hosts[host]; ok {
		return o, host, nil
	}
	return "", "", ErrUnknownHost
}

func testRegistry() staticLookup {
	return staticLookup{hosts: map[string]string{testHostA: testOrgA, testHostB: testOrgB}}
}

// tripIdentity fails the test on any call: the pre-store paths must never reach identity.
type tripIdentity struct{ t *testing.T }

func (x tripIdentity) MapConnectSubject(context.Context, string) (string, error) {
	x.t.Error("identity called")
	return "", identity.ErrUnavailable
}
func (x tripIdentity) ReadMCPConnection(context.Context, string, string, string) (identity.MCPConnection, error) {
	x.t.Error("identity called")
	return identity.MCPConnection{}, identity.ErrUnavailable
}
func (x tripIdentity) RenewMCPConnection(context.Context, identity.MCPConnection, time.Duration) error {
	x.t.Error("identity called")
	return identity.ErrUnavailable
}
func (x tripIdentity) RevokeMCPConnection(context.Context, identity.MCPConnection) error {
	x.t.Error("identity called")
	return identity.ErrUnavailable
}
func (x tripIdentity) RevokeOwnMCPConnection(context.Context, string, string, string) error {
	x.t.Error("identity called")
	return identity.ErrUnavailable
}
func (x tripIdentity) Session(context.Context, string, string) (identity.SessionInfo, error) {
	x.t.Error("identity called")
	return identity.SessionInfo{}, identity.ErrUnavailable
}
func (x tripIdentity) Now() time.Time { return time.Now() }

// stubVerifier answers every token with err (nil: an unexpected success fails the test through tripIdentity).
type stubVerifier struct {
	err   error
	calls int
}

func (v *stubVerifier) Validate(context.Context, string, string) (workos.ConnectClaims, error) {
	v.calls++
	return workos.ConnectClaims{}, v.err
}

func testHandler(t *testing.T, v Verifier, origins ...string) *Handler {
	t.Helper()
	return &Handler{cfg: Config{Issuer: issuer, Hosts: testRegistry(), Verifier: v, AllowedOrigins: origins},
		id: tripIdentity{t}, inFlight: make(chan struct{}, maxInFlight), limiters: map[string]*rate.Limiter{}}
}

func serve(h *Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(299) })).ServeHTTP(w, r)
	return w
}

func post(host, target, body string, headers map[string]string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	r.Host = host
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

const initBody = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}`

func TestConnectionIDPathMatrix(t *testing.T) {
	for _, p := range []string{"/mcp/c/" + testID} {
		if id, ok := connectionID(p, ConnectionPathPrefix); !ok || id != testID {
			t.Errorf("%s refused", p)
		}
	}
	for _, p := range []string{
		"/mcp/c/", "/mcp/c/" + testID + "/", "/mcp/c/" + strings.ToUpper(testID), "/mcp/c/" + testID + "/x",
		"/mcp/c/%37f000000-0000-4000-8000-000000000001", "/mcp/c/7f000000000040008000000000000001", "/mcp/c//" + testID,
		"/mcp/c/00000000-0000-0000-0000-000000000000", "/mcp/c/{" + testID + "}", "/mcp/C/" + testID, "/MCP/c/" + testID,
	} {
		if _, ok := connectionID(p, ConnectionPathPrefix); ok {
			t.Errorf("%s accepted", p)
		}
	}
	if resourceURL(testHostA, testID) != "https://org-a.localhost:543/mcp/c/"+testID ||
		metadataURL(testHostA, testID) != "https://org-a.localhost:543/.well-known/oauth-protected-resource/mcp/c/"+testID {
		t.Fatal("URL construction")
	}
}

// foldingLookup answers a case-folded host (a resolver that normalizes).
type foldingLookup struct{}

func (foldingLookup) Lookup(_ context.Context, host string) (string, string, error) {
	return testOrgA, strings.ToLower(host), nil
}

// TestLookupIsExact (G016): the HostLookup boundary is exact and unfolded; an origin must be exactly https://host[:port];
// a store failure is never an unknown host.
func TestLookupIsExact(t *testing.T) {
	ctx := context.Background()
	r := testRegistry()
	if h, err := lookupHost(ctx, r, testHostA); err != nil || h.OrganizationID != testOrgA || h.Origin != "https://"+testHostA {
		t.Fatal("exact host")
	}
	for _, h := range []string{"ORG-A.localhost:543", "org-a.localhost", "org-a.localhost:543.", " org-a.localhost:543", "org-c.localhost:543", ""} {
		if _, err := lookupHost(ctx, r, h); !unknownHost(err) {
			t.Errorf("%q matched", h)
		}
	}
	// A lookup answering a different host[:port] than asked (a folding resolver) is refused.
	if _, err := lookupHost(ctx, foldingLookup{}, "ORG-A.localhost:543"); !unknownHost(err) {
		t.Error("a folded answer must not match")
	}
	if h, err := lookupOrigin(ctx, r, "https://"+testHostB); err != nil || h.Host != testHostB {
		t.Fatal("origin")
	}
	for _, o := range []string{"http://" + testHostB, "https://" + testHostB + "/", "https://" + testHostB + "?x", "https://u@" + testHostB, testHostB} {
		if _, err := lookupOrigin(ctx, r, o); !unknownHost(err) {
			t.Errorf("origin %q matched", o)
		}
	}
	store := errors.New("store down")
	if _, err := lookupHost(ctx, staticLookup{fail: store}, testHostA); err == nil || unknownHost(err) {
		t.Fatal("a store failure must not be an unknown host")
	}
}

func TestMetadataIsStateless(t *testing.T) {
	h := testHandler(t, &stubVerifier{})
	for _, id := range []string{testID, "7f000000-0000-4000-8000-0000000000ff"} {
		r := httptest.NewRequest(http.MethodGet, MetadataPathPrefix+id, nil)
		r.Host = testHostA
		r.Header.Set("X-Forwarded-Host", testHostB)
		w := serve(h, r)
		var doc map[string]any
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &doc) != nil {
			t.Fatalf("metadata: %d %s", w.Code, w.Body)
		}
		want := map[string]any{"resource": "https://" + testHostA + "/mcp/c/" + id, "authorization_servers": []any{issuer},
			"bearer_methods_supported": []any{"header"}}
		if b1, b2 := mustJSON(doc), mustJSON(want); b1 != b2 {
			t.Fatalf("metadata = %s, want %s", b1, b2)
		}
		if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("headers %v", w.Header())
		}
	}
	for name, c := range map[string]struct{ host, path string }{
		"unregistered host": {"org-c.localhost:543", MetadataPathPrefix + testID},
		"upper-case host":   {"ORG-A.localhost:543", MetadataPathPrefix + testID},
		"malformed path":    {testHostA, MetadataPathPrefix + testID + "/"},
		"query":             {testHostA, MetadataPathPrefix + testID + "?x=1"},
	} {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		r.Host = c.host
		if w := serve(h, r); w.Code != http.StatusNotFound {
			t.Errorf("%s: %d", name, w.Code)
		}
	}
	r := httptest.NewRequest(http.MethodPost, MetadataPathPrefix+testID, nil)
	r.Host = testHostA
	if w := serve(h, r); w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST metadata: %d", w.Code)
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestTransportRefusalsBeforeTheStore(t *testing.T) {
	v := &stubVerifier{err: identity.ErrUnauthenticated}
	h := testHandler(t, v, "https://allowed.example")
	meta := `resource_metadata="https://org-a.localhost:543/.well-known/oauth-protected-resource/mcp/c/` + testID + `"`
	path := ConnectionPathPrefix + testID

	// Unknown host or malformed path: 404 with no fallback; everything outside the two families reaches next.
	for name, r := range map[string]*http.Request{
		"unregistered host": post("org-c.localhost:543", path, initBody, nil),
		"forwarded host":    post("org-c.localhost:543", path, initBody, map[string]string{"X-Forwarded-Host": testHostA, "Forwarded": "host=" + testHostA}),
		"trailing slash":    post(testHostA, path+"/", initBody, nil),
		"extra segment":     post(testHostA, path+"/x", initBody, nil),
		"upper-case uuid":   post(testHostA, ConnectionPathPrefix+strings.ToUpper(testID), initBody, nil),
	} {
		if w := serve(h, r); w.Code != http.StatusNotFound {
			t.Errorf("%s: %d", name, w.Code)
		}
	}
	if w := serve(h, post(testHostA, "/mcp", initBody, nil)); w.Code != 299 {
		t.Errorf("/mcp must reach the mux: %d", w.Code)
	}

	get := httptest.NewRequest(http.MethodGet, path, nil)
	get.Host = testHostA
	del := httptest.NewRequest(http.MethodDelete, path, nil)
	del.Host = testHostA
	for name, c := range map[string]struct {
		r    *http.Request
		code int
	}{
		"GET":              {get, http.StatusMethodNotAllowed},
		"DELETE":           {del, http.StatusMethodNotAllowed},
		"query":            {post(testHostA, path+"?x=1", initBody, nil), http.StatusBadRequest},
		"empty query":      {post(testHostA, path+"?", initBody, nil), http.StatusBadRequest},
		"origin":           {post(testHostA, path, initBody, map[string]string{"Origin": "https://evil.example"}), http.StatusForbidden},
		"batch":            {post(testHostA, path, "  ["+initBody+"]", nil), http.StatusBadRequest},
		"too large":        {post(testHostA, path, strings.Repeat(" ", maxBody+1), nil), http.StatusRequestEntityTooLarge},
		"no token":         {post(testHostA, path, initBody, nil), http.StatusUnauthorized},
		"basic scheme":     {post(testHostA, path, initBody, map[string]string{"Authorization": "Basic YTpi"}), http.StatusBadRequest},
		"bearer, no token": {post(testHostA, path, initBody, map[string]string{"Authorization": "Bearer "}), http.StatusBadRequest},
		"bearer, bad char": {post(testHostA, path, initBody, map[string]string{"Authorization": "Bearer a b"}), http.StatusBadRequest},
		"invalid token":    {post(testHostA, path, initBody, map[string]string{"Authorization": "Bearer eyJ.eyJ.sig"}), http.StatusUnauthorized},
	} {
		w := serve(h, c.r)
		if w.Code != c.code {
			t.Errorf("%s: %d, want %d", name, w.Code, c.code)
		}
		if name == "GET" && w.Header().Get("Allow") != "POST" {
			t.Errorf("405 without Allow: POST")
		}
		if name == "no token" && w.Header().Get("WWW-Authenticate") != "Bearer "+meta {
			t.Errorf("no token: WWW-Authenticate = %q", w.Header().Get("WWW-Authenticate"))
		}
		if name == "invalid token" && w.Header().Get("WWW-Authenticate") != `Bearer error="invalid_token", `+meta {
			t.Errorf("invalid token: WWW-Authenticate = %q", w.Header().Get("WWW-Authenticate"))
		}
	}
	// An allowed Origin passes to the bearer check.
	if w := serve(h, post(testHostA, path, initBody, map[string]string{"Origin": "https://allowed.example"})); w.Code != http.StatusUnauthorized {
		t.Errorf("allowed origin: %d", w.Code)
	}
	// Two Authorization headers are malformed.
	r := post(testHostA, path, initBody, nil)
	r.Header.Add("Authorization", "Bearer a")
	r.Header.Add("Authorization", "Bearer b")
	if w := serve(h, r); w.Code != http.StatusBadRequest {
		t.Errorf("repeated Authorization: %d", w.Code)
	}
	// A key-set or verifier outage is 503, never anonymous.
	v.err = identity.ErrUnavailable
	if w := serve(h, post(testHostA, path, initBody, map[string]string{"Authorization": "Bearer eyJ.eyJ.sig"})); w.Code != http.StatusServiceUnavailable {
		t.Errorf("verifier outage: %d", w.Code)
	}
	// Admission: a full in-flight bound answers 429 before any token work.
	calls := v.calls
	for i := 0; i < maxInFlight; i++ {
		h.inFlight <- struct{}{}
	}
	if w := serve(h, post(testHostA, path, initBody, map[string]string{"Authorization": "Bearer eyJ.eyJ.sig"})); w.Code != http.StatusTooManyRequests {
		t.Errorf("in-flight bound: %d", w.Code)
	}
	if v.calls != calls {
		t.Error("the verifier ran past the admission bound")
	}
}

// A slow or stalled body never holds an in-flight slot (S-review F1): with every slot taken, a request whose body is
// still arriving gets no answer until the body ends; only then is a slot claimed (here: none free, 429).
func TestStalledBodyHoldsNoSlot(t *testing.T) {
	v := &stubVerifier{err: errors.New("x")}
	h := testHandler(t, v)
	srv := httptest.NewServer(h.Wrap(http.NotFoundHandler()))
	defer srv.Close()
	for i := 0; i < maxInFlight; i++ {
		h.inFlight <- struct{}{}
	}
	pr, pw := io.Pipe()
	req, err := http.NewRequest(http.MethodPost, srv.URL+ConnectionPathPrefix+testID, pr)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = testHostA
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer eyJ.eyJ.sig")
	done := make(chan int, 1)
	go func() {
		resp, err := srv.Client().Do(req)
		if err != nil {
			done <- 0
			return
		}
		_ = resp.Body.Close()
		done <- resp.StatusCode
	}()
	if _, err := pw.Write([]byte(`{"jsonrpc":"2.0",`)); err != nil {
		t.Fatal(err)
	}
	select {
	case code := <-done:
		t.Fatalf("answered %d while the body was still arriving (a slot was claimed first)", code)
	case <-time.After(300 * time.Millisecond):
	}
	_, _ = pw.Write([]byte(`"id":1,"method":"ping"}`))
	_ = pw.Close()
	select {
	case code := <-done:
		if code != http.StatusTooManyRequests {
			t.Fatalf("after the body: %d, want 429", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no answer after the body ended")
	}
	if v.calls != 0 {
		t.Error("the verifier ran past the admission bound")
	}
}

func TestToolCatalog(t *testing.T) {
	want := []string{"account.get_current_user", "account.get_current_context", "context.list_groups", "context.switch_group", "logout"}
	if !slices.Equal(ToolNames, want) {
		t.Fatalf("tools = %v", ToolNames)
	}
	grammar := regexp.MustCompile(`^[a-z][a-z_]*(\.[a-z][a-z_]*)?$`)
	for _, n := range ToolNames {
		in, out, ok := ToolSchemas(n)
		if !ok || in["type"] != "object" || in["additionalProperties"] != false || out["type"] != "object" || out["additionalProperties"] != false {
			t.Errorf("%s schemas: %v %v", n, in, out)
		}
		if !grammar.MatchString(n) {
			t.Errorf("%s breaks the tool-name grammar", n)
		}
	}
	for _, retired := range []string{"account.switch_organization", "account.switch_agency", "account.list_organizations"} {
		if _, _, ok := ToolSchemas(retired); ok || slices.Contains(ToolNames, retired) {
			t.Errorf("%s present", retired)
		}
	}
	in, _, _ := ToolSchemas(ToolSwitchGroup)
	if r, _ := in["required"].([]string); !slices.Equal(r, []string{"group_id"}) {
		t.Errorf("switch_group requires %v", in["required"])
	}
	in, _, _ = ToolSchemas(ToolListGroups)
	if _, req := in["required"]; req {
		t.Error("list_groups has required arguments")
	}
	// No argument of any tool names an Organization.
	for _, n := range ToolNames {
		in, _, _ := ToolSchemas(n)
		for k := range in["properties"].(map[string]any) {
			if strings.Contains(k, "organization") || strings.Contains(k, "tenant") || strings.Contains(k, "agency") {
				t.Errorf("%s takes %s", n, k)
			}
		}
	}

	// G015 (Plan r4 §1, §6.2): the production spec is exactly the five control tools with their requirements and no
	// capability-mapped tool; a new handler serves it (the fixture seam is unset outside tests).
	wantReq := map[string]access.ToolRequirement{
		ToolGetCurrentUser: access.RequireContext(), ToolGetCurrentContext: access.RequireContext(),
		ToolListGroups: access.RequireContext(), ToolSwitchGroup: access.RequireMembership(), ToolLogout: access.RequireBinding(),
	}
	if len(productionCatalog.capability) != 0 || len(productionCatalog.reqs) != 0 || len(productionCatalog.control) != len(ToolNames) {
		t.Fatalf("production catalog: %d control, %d capability", len(productionCatalog.control), len(productionCatalog.capability))
	}
	for i, e := range productionCatalog.control {
		if e.name != ToolNames[i] || e.req != wantReq[e.name] || e.capabilityMapped() {
			t.Errorf("production entry %d: %s %+v", i, e.name, e.req)
		}
	}
	if c := testHandler(t, &stubVerifier{}).catalog(); len(c.capability) != 0 || len(c.control) != len(ToolNames) {
		t.Error("a new handler serves a catalog with extra entries")
	}
	// The fixture spec adds four capability-mapped entries and keeps the controls.
	extra, _ := fixtureEntries(t)
	c, err := catalogSpec(append(productionEntries(), extra...))
	if err != nil || len(c.control) != len(ToolNames) || len(c.capability) != 4 || len(c.reqs) != 4 {
		t.Fatalf("fixture catalog: %v %d %d", err, len(c.control), len(c.capability))
	}
	if c.isControl(fixtureGroupUpdate) || !c.isControl(ToolLogout) || c.isControl("") {
		t.Error("isControl")
	}
	// G020 (Plan r2 §4.3, T10; D-1): the test-only label fixture spec adds exactly three capability-mapped entries on the
	// three canonical label capabilities, follows the tool-name grammar and keeps the controls; the production spec above
	// is unchanged (five control tools, no capability-mapped tool) and never names a label tool.
	lextra, _ := labelFixtureEntries(t)
	lc, err := catalogSpec(append(productionEntries(), lextra...))
	if err != nil || len(lc.control) != len(ToolNames) || len(lc.capability) != 3 || len(lc.reqs) != 3 {
		t.Fatalf("label fixture catalog: %v %d %d", err, len(lc.control), len(lc.capability))
	}
	for _, n := range labelFixtureNames {
		req, rerr := access.RequireCapability(labelFixtureCapabilities[n])
		if rerr != nil || lc.reqs[n] != req || lc.isControl(n) || !grammar.MatchString(n) {
			t.Errorf("label fixture %s: %+v %v", n, lc.reqs[n], rerr)
		}
		if _, _, ok := ToolSchemas(n); ok || slices.Contains(ToolNames, n) {
			t.Errorf("label fixture %s is a production tool", n)
		}
		if _, ok := productionCatalog.reqs[n]; ok {
			t.Errorf("the production catalog maps %s", n)
		}
	}
	// Unknown capabilities, zero (unknown) requirements, repeated or empty names and incomplete entries are refused.
	if _, err := access.RequireCapability("group.delete"); access.CodeOf(err) != access.CodeInvalidInput {
		t.Errorf("unknown capability: %v", err)
	}
	for name, mutate := range map[string]func(e *catalogEntry){
		"zero requirement": func(e *catalogEntry) { e.req = access.ToolRequirement{} },
		"repeated name":    func(e *catalogEntry) { e.name = ToolLogout },
		"empty name":       func(e *catalogEntry) { e.name = "" },
		"no handler":       func(e *catalogEntry) { e.handler = nil },
		"no input schema":  func(e *catalogEntry) { e.input = nil },
		"non-object input": func(e *catalogEntry) { e.input = map[string]any{"type": "string"} },
	} {
		bad := slices.Clone(extra)
		mutate(&bad[0])
		if _, err := catalogSpec(append(productionEntries(), bad...)); !errors.Is(err, errCatalog) {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestArgumentDecoding(t *testing.T) {
	var none noArgs
	for _, raw := range []string{"", "null", "{}", " { } "} {
		if !decodeArgs(json.RawMessage(raw), &none) {
			t.Errorf("%q refused", raw)
		}
	}
	for _, raw := range []string{"[]", `{"x":1}`, `{} {}`, `"s"`, `{`, "1"} {
		if decodeArgs(json.RawMessage(raw), &none) {
			t.Errorf("%q accepted", raw)
		}
	}
}

func TestCursor(t *testing.T) {
	c := encodeCursor(testOrgA, testID)
	if g, ok := decodeCursor(c, testOrgA); !ok || g != testID {
		t.Fatal("round trip")
	}
	for name, s := range map[string]string{
		"other Organization": c, "garbage": "!!", "padded": c + "=", "wrong version": encodeCursor(testOrgA, testID)[1:],
		"zero group": encodeCursor(testOrgA, "00000000-0000-0000-0000-000000000000"),
	} {
		org := testOrgA
		if name == "other Organization" {
			org = testOrgB
		}
		if _, ok := decodeCursor(s, org); ok {
			t.Errorf("%s accepted", name)
		}
	}
	testSetupCursor(t)
	testSetupListWireBound(t)
}

// testSetupCursor is the setup listing cursor (Duo follow-up correction; addendum 01 A2-A4): a strict round trip in
// one (user, Organization, trusted origin) context, refusal of every other context and shape, and the time domain.
func testSetupCursor(t *testing.T) {
	t.Helper()
	const user, other = "5a000000-0000-4000-8000-0000000000aa", "5b000000-0000-4000-8000-0000000000bb"
	originA, alias := "https://"+testHostA, "https://org-a-alias.localhost:543"
	at := time.Date(2026, 10, 16, 12, 0, 0, 123456000, time.UTC)
	raw := func(micro, id, tag string) string {
		return base64.RawURLEncoding.EncodeToString([]byte("s1|" + micro + "|" + id + "|" + tag))
	}
	tag := setupCursorTag(user, testOrgA, originA)
	c, ok := encodeSetupCursor(user, testOrgA, originA, identity.MCPListPosition{CreatedAt: at, ID: testID})
	if !ok || len(c) > setupCursorMaxLen {
		t.Fatalf("encode: %q %v", c, ok)
	}
	if p, ok := decodeSetupCursor(c, user, testOrgA, originA); !ok || !p.CreatedAt.Equal(at) || p.ID != testID {
		t.Fatalf("round trip: %+v %v", p, ok)
	}
	// The worst-case cursor stays within the documented 128-character bound (it is at most 122).
	if w, ok := encodeSetupCursor(user, testOrgA, originA, identity.MCPListPosition{CreatedAt: time.UnixMicro(setupCursorMaxMicro), ID: testID}); !ok || len(w) > 122 {
		t.Fatalf("maximum cursor: %d %v", len(w), ok)
	}
	// Another context: user, Organization, trusted origin (including an alias of the same Organization), or swapped.
	for name, ctx := range map[string][3]string{
		"other user": {other, testOrgA, originA}, "other Organization": {user, testOrgB, originA},
		"other origin, same Organization": {user, testOrgA, alias}, "swapped user and Organization": {testOrgA, user, originA},
	} {
		if _, ok := decodeSetupCursor(c, ctx[0], ctx[1], ctx[2]); ok {
			t.Errorf("%s accepted", name)
		}
	}
	flipped := []byte(tag)
	flipped[0] ^= 1
	for name, s := range map[string]string{
		"empty": "", "garbage": "!!", "padded": c + "=", "overlong": strings.Repeat("A", setupCursorMaxLen+1),
		"tool cursor":       encodeCursor(testOrgA, testID),
		"wrong version":     base64.RawURLEncoding.EncodeToString([]byte("s2|1|" + testID + "|" + tag)),
		"three fields":      base64.RawURLEncoding.EncodeToString([]byte("s1|1|" + testID)),
		"five fields":       raw("1", testID, tag+"|x"),
		"flipped tag":       raw("1", testID, string(flipped)),
		"uppercase tag":     raw("1", testID, strings.ToUpper(tag)),
		"zero UUID":         raw("1", "00000000-0000-0000-0000-000000000000", tag),
		"uppercase UUID":    raw("1", strings.ToUpper(testID), tag),
		"time zero":         raw("0", testID, tag),
		"leading zero":      raw("01", testID, tag),
		"negative":          raw("-1", testID, tag),
		"plus sign":         raw("+1", testID, tag),
		"above the domain":  raw("253402300800000000", testID, tag),
		"int64 maximum":     raw("9223372036854775807", testID, tag),
		"int64 overflow":    raw("9223372036854775808", testID, tag),
		"not a number":      raw("1e6", testID, tag),
		"standard alphabet": strings.NewReplacer("-", "+", "_", "/").Replace(raw("1", testID, tag)) + "+/",
	} {
		if _, ok := decodeSetupCursor(s, user, testOrgA, originA); ok {
			t.Errorf("%s accepted", name)
		}
	}
	if _, ok := decodeCursor(c, testOrgA); ok {
		t.Error("a setup cursor accepted as a tool cursor")
	}
	// The inclusive time domain: both edges accepted, and the encoder refuses outside it (the listing fails closed).
	for _, micro := range []string{"1", "253402300799999999"} {
		if _, ok := decodeSetupCursor(raw(micro, testID, tag), user, testOrgA, originA); !ok {
			t.Errorf("domain edge %s refused", micro)
		}
	}
	for name, p := range map[string]identity.MCPListPosition{
		"epoch":            {CreatedAt: time.UnixMicro(0), ID: testID},
		"above the domain": {CreatedAt: time.UnixMicro(setupCursorMaxMicro + 1), ID: testID},
		"zero UUID":        {CreatedAt: at, ID: "00000000-0000-0000-0000-000000000000"},
	} {
		if _, ok := encodeSetupCursor(user, testOrgA, originA, p); ok {
			t.Errorf("encoder accepted %s", name)
		}
	}
}

// testSetupListWireBound encodes a worst-case listing through the real writer (addendum 01 A1): 100 active rows with a
// 2048-byte client ID of characters the encoder escapes to six bytes, a 310-byte resource URL on a 253-character host
// with port 65535, timestamps at the domain maximum, a 43-character CSRF token and a 128-character cursor. Its size is
// the BFF's list-success bound (web/lib/mcp-client.ts MCP_LIST_MAX_RESPONSE_BYTES), and no other client ID exceeds it.
// G063 (Plan r1 §5.3): every row also carries display_name; the worst case is MCPNameMax code points the encoder escapes
// to six bytes each (80 "<" -> 482 bytes with the quotes, 498 per row with the key and separator, +49 800 in all). No
// other accepted name (non-BMP, ampersands, quotes, backslashes) exceeds it, and an unnamed row costs less.
func testSetupListWireBound(t *testing.T) {
	t.Helper()
	const bound = 1_327_022
	host := strings.Repeat(strings.Repeat("a", 63)+".", 3) + strings.Repeat("a", 61) + ":65535"
	resource := resourceURL(host, testID)
	if len(resource) != 310 {
		t.Fatalf("resource length %d", len(resource))
	}
	top := time.UnixMicro(setupCursorMaxMicro).UTC()
	worstName := strings.Repeat("<", identity.MCPNameMax)
	sizeNamed := func(client string, name *string) int {
		rows := make([]map[string]any, 0, identity.MCPListPage)
		for range identity.MCPListPage {
			rows = append(rows, connectionView(identity.MCPConnectionInfo{ID: testID, Resource: resource, State: identity.MCPStateActive,
				ClientID: client, CreatedAt: top, ActivatedAt: top, DisplayName: name}))
		}
		w := httptest.NewRecorder()
		writeJSON(w, http.StatusOK, map[string]any{"connections": rows, "csrf_token": strings.Repeat("c", 43), "next_cursor": strings.Repeat("n", 128)})
		return w.Body.Len()
	}
	size := func(client string) int { return sizeNamed(client, &worstName) }
	if got := size(strings.Repeat("<", 2048)); got != bound {
		t.Fatalf("worst-case listing: %d bytes, want %d", got, bound)
	}
	if got := sizeNamed(strings.Repeat("<", 2048), nil); got != bound-100*(482-4) {
		t.Errorf("worst-case unnamed listing: %d bytes, want %d", got, bound-100*(482-4))
	}
	for name, n := range map[string]string{
		"non-BMP": strings.Repeat("\U0001F600", identity.MCPNameMax), "ampersand": strings.Repeat("&", identity.MCPNameMax),
		"quotes and backslashes": strings.Repeat("\"\\", identity.MCPNameMax/2), "markup": strings.Repeat("<a>&", identity.MCPNameMax/4),
	} {
		if v, err := identity.NormalizeMCPDisplayName(n); err != nil || v == nil || *v != n {
			t.Fatalf("%s: not an accepted name (%v)", name, err)
		}
		if got := sizeNamed(strings.Repeat("<", 2048), &n); got > bound {
			t.Errorf("name %s: %d bytes > %d", name, got, bound)
		}
	}
	for name, client := range map[string]string{
		"ampersand": strings.Repeat("&", 2048), "control": strings.Repeat("\x01", 2048),
		"line separator": strings.Repeat(" ", 682) + "aa", "mixed": strings.Repeat("<\x01 &\"\\", 256),
	} {
		if len(client) > 2048 {
			t.Fatalf("%s: %d bytes", name, len(client))
		}
		if got := size(client); got > bound {
			t.Errorf("%s: %d bytes > %d", name, got, bound)
		}
	}
}

func TestErrorMapping(t *testing.T) {
	for in, want := range map[access.Code]string{
		access.CodeContextDenied: CodeContextDenied, access.CodeUnauthenticated: CodeContextDenied,
		access.CodeGroupNotSelectable: CodeGroupNotSelectable, access.CodeInvalidInput: CodeInvalidInput,
		access.CodeUnavailable: CodeUnavailable, access.CodeInternal: CodeUnavailable, access.CodeConflict: CodeUnavailable,
	} {
		if got := toolCode(&access.Error{Code: in}); got != want {
			t.Errorf("%s -> %s, want %s", in, got, want)
		}
	}
	if toolCode(errors.New("x")) != CodeUnavailable {
		t.Error("unknown error")
	}
	if e := rpcError(&access.Error{Code: access.CodeContextDenied}); e.Code != RPCCodeContextDenied || e.Message != CodeContextDenied {
		t.Errorf("rpc context_denied: %+v", e)
	}
	if e := rpcError(&access.Error{Code: access.CodeUnavailable}); e.Code != RPCCodeUnavailable || e.Message != CodeUnavailable {
		t.Errorf("rpc unavailable: %+v", e)
	}
	res := toolError(CodeInvalidInput, FieldInvalidCursor)
	if !res.IsError || mustJSON(res.StructuredContent) != `{"error":"invalid_input","field":"invalid_cursor"}` {
		t.Errorf("tool error: %+v", res)
	}
	// G015 C-1/C-6: a hidden tool is answered by the SDK's own unknown-tool path: invalid params, the requested name
	// echoed in a fixed template. Nothing in this package composes that envelope.
	cs, _ := connect(t, NewServer(Options{}))
	defer cs.Close()
	for _, name := range []string{fixtureGroupUpdate, "test.never_existed"} {
		_, err := cs.CallTool(context.Background(), &sdk.CallToolParams{Name: name})
		var we *jsonrpc.Error
		if !errors.As(err, &we) || we.Code != jsonrpc.CodeInvalidParams || we.Message != "unknown tool "+strconv.Quote(name) {
			t.Errorf("unknown tool %s: %v", name, err)
		}
	}
}

// U-MCP-BOUNDARY: the package's production imports, no SQL, exactly one streamable handler call site and no other
// transport.
func TestPackageBoundary(t *testing.T) {
	allowed := map[string]bool{
		"bytes": true, "context": true, "encoding/base64": true, "encoding/json": true, "errors": true, "io": true,
		"log/slog": true, "mime": true, "net/http": true, "slices": true, "strings": true, "sync": true,
		"sync/atomic": true, "time": true, // sync/atomic: Handler.fixture, the G015 D2 catalog override
		"crypto/sha256": true, "crypto/subtle": true, "encoding/hex": true, "strconv": true, // the setup listing cursor
		"github.com/jackc/pgx/v5/pgxpool": true, "github.com/modelcontextprotocol/go-sdk/mcp": true,
		"github.com/modelcontextprotocol/go-sdk/jsonrpc": true, "golang.org/x/time/rate": true,
		"core/internal/identity": true, "core/internal/identity/workos": true, "core/internal/observability": true,
		"core/internal/tenancy/access": true, "unicode/utf8": true, // G063 unicode/utf8: the strict friendly-name token
		// G016: the HostLookup consumer names the resolver's unknown-host sentinel and parses exact origins.
		"core/internal/domains": true, "net/url": true,
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	handlers := 0
	for _, e := range entries {
		n := e.Name()
		if !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, n, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if !allowed[p] {
				t.Errorf("%s imports %s", n, p)
			}
		}
		ast.Inspect(f, func(x ast.Node) bool {
			switch v := x.(type) {
			case *ast.BasicLit:
				if v.Kind == token.STRING && regexp.MustCompile(`\bapp\.|\b(SELECT|INSERT INTO|DELETE FROM|FOR SHARE)\b`).MatchString(v.Value) {
					t.Errorf("%s holds an SQL literal: %s", n, v.Value)
				}
			case *ast.SelectorExpr:
				switch v.Sel.Name {
				case "NewStreamableHTTPHandler":
					handlers++
				case "NewSSEHandler", "SSEServerTransport", "StdioTransport", "CommandTransport", "NewInMemoryTransports":
					t.Errorf("%s uses %s", n, v.Sel.Name)
				}
			}
			return true
		})
	}
	if handlers != 1 {
		t.Errorf("%d NewStreamableHTTPHandler call sites, want 1", handlers)
	}
}

// sessionCountingIdentity is tripIdentity whose Session answers unauthenticated and counts the calls (G063
// TestSetupNameDecoding): a refusal before begin leaves the count at zero; a request that passes the name settles
// reaches the session read.
type sessionCountingIdentity struct {
	tripIdentity
	sessions *int
}

func (x sessionCountingIdentity) Session(context.Context, string, string) (identity.SessionInfo, error) {
	*x.sessions++
	return identity.SessionInfo{}, identity.ErrUnauthenticated
}

// TestSetupNameDecoding (G063 Plan r1 §4.3, §6; r1 C4): the endpoint-level wire negatives of the friendly name on both
// setup mutations, with raw body bytes. Malformed UTF-8 and lone, reversed or unpaired surrogate escapes are refused as
// invalid_name, never replaced; null, a duplicate (including a case-folded alias the struct decoder would accept) and a
// non-string value are invalid_request; the rename requires the name, the create does not; every name refusal is
// answered before any session or CSRF check. A valid escaped pair or raw non-BMP text reaches the session read. The
// legacy fields keep decodeSetup's last-wins behavior.
func TestSetupNameDecoding(t *testing.T) {
	sessions := 0
	h := testHandler(t, &stubVerifier{})
	h.id = sessionCountingIdentity{tripIdentity{t}, &sessions}
	origin := `"origin":"https://` + testHostA + `"`
	bodies := map[string]func(name string) string{
		"create": func(name string) string { return `{` + origin + `,"csrf_token":"c"` + name + `}` },
		"rename": func(name string) string {
			return `{` + origin + `,"connection_id":"` + testID + `","csrf_token":"c"` + name + `}`
		},
	}
	handlers := map[string]func(http.ResponseWriter, *http.Request){"create": h.Create, "rename": h.Rename}
	call := func(op, body string) (int, string) {
		r := httptest.NewRequest(http.MethodPost, "/internal/v1/mcp/connections", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Session credential")
		w := httptest.NewRecorder()
		handlers[op](w, r)
		return w.Code, strings.TrimSpace(w.Body.String())
	}
	refusals := map[string]struct {
		field string
		code  int
		want  string
	}{
		"invalid UTF-8 byte":         {",\"display_name\":\"\xff\"", http.StatusBadRequest, "invalid_name"},
		"overlong UTF-8":             {",\"display_name\":\"a\xc0\xafb\"", http.StatusBadRequest, "invalid_name"},
		"encoded surrogate bytes":    {",\"display_name\":\"\xed\xa0\x80\"", http.StatusBadRequest, "invalid_name"},
		"lone high surrogate escape": {`,"display_name":"\ud800"`, http.StatusBadRequest, "invalid_name"},
		"lone low surrogate escape":  {`,"display_name":"\udc00"`, http.StatusBadRequest, "invalid_name"},
		"reversed surrogate pair":    {`,"display_name":"\udc00\ud800"`, http.StatusBadRequest, "invalid_name"},
		"high surrogate then text":   {`,"display_name":"\uD800x"`, http.StatusBadRequest, "invalid_name"},
		"high surrogate then BMP":    {`,"display_name":"\ud800A"`, http.StatusBadRequest, "invalid_name"},
		"escaped newline":            {`,"display_name":"a\nb"`, http.StatusBadRequest, "invalid_name"},
		"escaped NEL":                {`,"display_name":"\u0085a"`, http.StatusBadRequest, "invalid_name"},
		"81 code points":             {`,"display_name":"` + strings.Repeat("a", 81) + `"`, http.StatusBadRequest, "invalid_name"},
		"null":                       {`,"display_name":null`, http.StatusBadRequest, "invalid_request"},
		"number":                     {`,"display_name":5`, http.StatusBadRequest, "invalid_request"},
		"object":                     {`,"display_name":{}`, http.StatusBadRequest, "invalid_request"},
		"duplicate key":              {`,"display_name":"a","display_name":"b"`, http.StatusBadRequest, "invalid_request"},
		"case-folded duplicate":      {`,"Display_Name":"a","display_name":"b"`, http.StatusBadRequest, "invalid_request"},
	}
	for op, body := range bodies {
		for name, c := range refusals {
			sessions = 0
			code, got := call(op, body(c.field))
			if code != c.code || got != `{"error":"`+c.want+`"}` {
				t.Errorf("%s, %s: %d %s, want %d %s", op, name, code, got, c.code, c.want)
			}
			if sessions != 0 {
				t.Errorf("%s, %s: the session was read before the name was settled", op, name)
			}
		}
		for name, field := range map[string]string{
			"escaped surrogate pair": `,"display_name":"\ud83d\ude00"`,
			"upper-case pair":        `,"display_name":"\uD83D\uDE00"`,
			"raw non-BMP":            `,"display_name":"` + strings.Repeat("\U0001F600", 80) + `"`,
			"padded clear":           `,"display_name":"   "`,
			"markup":                 `,"display_name":"<script>alert(1)</script>"`,
		} {
			sessions = 0
			if code, got := call(op, body(field)); code != http.StatusUnauthorized || got != `{"error":"unauthenticated"}` || sessions != 1 {
				t.Errorf("%s, %s: %d %s (%d session reads), want the session read and 401", op, name, code, got, sessions)
			}
		}
	}
	// The rename requires the name; the create does not.
	sessions = 0
	if code, got := call("rename", bodies["rename"]("")); code != http.StatusBadRequest || got != `{"error":"invalid_request"}` || sessions != 0 {
		t.Errorf("rename without display_name: %d %s (%d)", code, got, sessions)
	}
	if code, _ := call("create", bodies["create"]("")); code != http.StatusUnauthorized || sessions != 1 {
		t.Errorf("create without display_name: %d (%d session reads), want the legacy shape to reach the session", code, sessions)
	}
	// A malformed connection UUID is invalid_request after the origin and before the session.
	sessions = 0
	bad := `{` + origin + `,"connection_id":"` + strings.ToUpper(testID) + `","csrf_token":"c","display_name":"x"}`
	if code, got := call("rename", bad); code != http.StatusBadRequest || got != `{"error":"invalid_request"}` || sessions != 0 {
		t.Errorf("rename with a malformed UUID: %d %s (%d)", code, got, sessions)
	}
	// Legacy fields keep last-wins duplicates: the second origin (the registered one) decides.
	sessions = 0
	dup := `{"origin":"https://evil.example","origin":"https://` + testHostA + `","csrf_token":"c","display_name":"x"}`
	if code, _ := call("create", dup); code != http.StatusUnauthorized || sessions != 1 {
		t.Errorf("duplicate origin: %d (%d session reads), want last-wins", code, sessions)
	}
	// The pure helpers.
	for raw, ok := range map[string]bool{`"a"`: true, `"\ud83d\ude00"`: true, `"\uD83D\uDE00"`: true, `"\\ud800"`: true, `"\ud800"`: false, `"\udfff"`: false,
		`"\ud800\ud800"`: false, `"\ud800\"`: false} {
		if surrogatesPaired([]byte(raw)) != ok {
			t.Errorf("surrogatesPaired(%s) != %v", raw, ok)
		}
	}
}
