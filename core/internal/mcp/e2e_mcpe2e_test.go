//go:build mcpe2e

// G014 B-MCP-E2E (Plan r8 §10.5): the MCP end-to-end driver. It runs in the core check image in the identity
// project's web network namespace, after the identity browser run and its probes, against the real core, replica,
// web and TLS test proxy of the disposable identity environment (infra/compose/identity.yml). Every MCP request goes
// through the TLS edge (https://org-a.localhost:543 and https://org-b.localhost:543, the run CA as the only root);
// setup actions go through web's form routes on the edge with the browser sessions the mcp-* browser projects handed
// over (M-HANDOFF); listings read core's internal API directly, as web does. Connect access tokens are minted with the
// per-run Connect test signer of identitysetup, whose public key set is core's test trust anchor (C10, APP_ENV=test
// only). Every minted token is printed on its own "MCP-OBSERVED <token>" stdout line for the harness's secret registry,
// which removes those lines from the saved log (E2E-MCP-7).
//
// Required environment: MCP_E2E_HANDOFF, MCP_E2E_CA, MCP_E2E_SIGNER, MCP_E2E_ISSUER, MCP_E2E_SUBJECT (dave's WorkOS
// user ID), MCP_E2E_USER (dave's account ID), MCP_E2E_ORG_A, MCP_E2E_ORG_B, MCP_E2E_EDGE (host:port of the proxy)
// and MCP_E2E_CORE (core's internal base URL).
package mcp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"core/internal/identity/identitytest"
)

const (
	e2eClient      = "client_01MCPE2ECLIENT00000000000"
	e2eOtherClient = "client_01MCPE2EOTHERCLIENT000000"
	e2eKeyID       = "mcp-connect-test" // identitysetup's ConnectKeyID
	e2eResearch    = "0f14a000-0000-4000-8000-0000000000a1"
	e2eDesign      = "0f14a000-0000-4000-8000-0000000000a2"
	e2eOperations  = "0f14b000-0000-4000-8000-0000000000b1"
)

type e2eHandoff struct {
	OriginA    string `json:"origin_a"`
	OriginB    string `json:"origin_b"`
	LiveA      string `json:"live_a"`
	CSRFA      string `json:"csrf_a"`
	LiveB      string `json:"live_b"`
	CSRFB      string `json:"csrf_b"`
	SpareA     string `json:"spare_a"`
	SpareCSRFA string `json:"spare_csrf_a"`
	A1         string `json:"a1"`
	A2         string `json:"a2"`
	B1         string `json:"b1"`
}

type e2eEnv struct {
	h                 e2eHandoff
	edge, core        *http.Client
	coreURL           string
	key               *rsa.PrivateKey
	issuer, sub, user string
	orgA, orgB        string
	minted            atomic.Int64
}

func e2eGetenv(t *testing.T, name string) string {
	t.Helper()
	v := os.Getenv(name)
	if v == "" {
		t.Fatalf("%s is required", name)
	}
	return v
}

func setupE2E(t *testing.T) *e2eEnv {
	t.Helper()
	e := &e2eEnv{issuer: e2eGetenv(t, "MCP_E2E_ISSUER"), sub: e2eGetenv(t, "MCP_E2E_SUBJECT"), user: e2eGetenv(t, "MCP_E2E_USER"),
		orgA: e2eGetenv(t, "MCP_E2E_ORG_A"), orgB: e2eGetenv(t, "MCP_E2E_ORG_B"), coreURL: e2eGetenv(t, "MCP_E2E_CORE")}
	b, err := os.ReadFile(e2eGetenv(t, "MCP_E2E_HANDOFF"))
	if err != nil || json.Unmarshal(b, &e.h) != nil || e.h.A1 == "" || e.h.A2 == "" || e.h.B1 == "" || e.h.LiveA == "" || e.h.SpareA == "" {
		t.Fatalf("handoff unreadable or incomplete")
	}
	caPEM, err := os.ReadFile(e2eGetenv(t, "MCP_E2E_CA"))
	if err != nil {
		t.Fatal("run CA unreadable")
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		t.Fatal("run CA invalid")
	}
	keyPEM, err := os.ReadFile(e2eGetenv(t, "MCP_E2E_SIGNER"))
	if err != nil {
		t.Fatal("Connect test signer unreadable")
	}
	blk, _ := pem.Decode(keyPEM)
	if blk == nil {
		t.Fatal("Connect test signer invalid")
	}
	k, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		t.Fatal("Connect test signer invalid")
	}
	e.key = k.(*rsa.PrivateKey)
	edge := e2eGetenv(t, "MCP_E2E_EDGE")
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	e.edge = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			// Every edge name resolves to the proxy, as a browser on the test machine would reach it.
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, edge)
			},
			TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12},
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	e.core = &http.Client{Timeout: 30 * time.Second}
	return e
}

// token mints a Connect access token of the subject for audience aud and client; over replaces claims (nil deletes).
func (e *e2eEnv) token(t *testing.T, aud, client string, over map[string]any) string {
	t.Helper()
	sid := make([]byte, 8)
	_, _ = rand.Read(sid)
	now := time.Now()
	c := map[string]any{"iss": e.issuer, "sub": e.sub, "aud": aud, "client_id": client, "iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(), "sid": "session_" + hex.EncodeToString(sid), "scope": "openid offline_access"}
	for k, v := range over {
		if v == nil {
			delete(c, k)
		} else {
			c[k] = v
		}
	}
	tok := identitytest.SignWith(e.key, "RS256", e2eKeyID, c)
	fmt.Println("MCP-OBSERVED " + tok)
	e.minted.Add(1)
	return tok
}

// post sends one MCP JSON-RPC POST through the edge and returns the status, headers and decoded body.
func (e *e2eEnv) post(t *testing.T, target, tok, body string) (int, http.Header, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2025-11-25")
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := e.edge.Do(req)
	if err != nil {
		t.Fatalf("edge request: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out)
	return resp.StatusCode, resp.Header, out
}

var e2eSeq atomic.Int64

func (e *e2eEnv) rpc(t *testing.T, target, tok, method string, params any) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": e2eSeq.Add(1), "method": method, "params": params})
	code, _, out := e.post(t, target, tok, string(b))
	return code, out
}

// call calls a tool and returns the HTTP status, the structured content and whether it is a tool error.
func (e *e2eEnv) call(t *testing.T, target, tok, tool string, args any) (int, map[string]any, bool) {
	t.Helper()
	if args == nil {
		args = map[string]any{}
	}
	code, out := e.rpc(t, target, tok, "tools/call", map[string]any{"name": tool, "arguments": args})
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

func (e *e2eEnv) mustCall(t *testing.T, target, tok, tool string, args any) map[string]any {
	t.Helper()
	code, sc, isErr := e.call(t, target, tok, tool, args)
	if code != http.StatusOK || isErr {
		t.Fatalf("%s: %d %v", tool, code, sc)
	}
	return sc
}

func (e *e2eEnv) wantToolError(t *testing.T, target, tok, tool string, args any, want string) {
	t.Helper()
	code, sc, isErr := e.call(t, target, tok, tool, args)
	if code != http.StatusOK || !isErr || sc["error"] != want {
		t.Fatalf("%s: %d %v, want tool error %s", tool, code, sc, want)
	}
}

func (e *e2eEnv) want401(t *testing.T, target, tok, what string) {
	t.Helper()
	if code, out := e.rpc(t, target, tok, "tools/call", map[string]any{"name": ToolGetCurrentContext, "arguments": map[string]any{}}); code != http.StatusUnauthorized {
		t.Fatalf("%s: %d %v, want 401", what, code, out)
	}
}

// group returns the selected group ID of a context result ("" when none).
func e2eGroup(sc map[string]any) string {
	g, _ := sc["group"].(map[string]any)
	id, _ := g["group_id"].(string)
	return id
}

// form posts a web form route on the edge with a browser session and returns the status and Location.
func (e *e2eEnv) form(t *testing.T, origin, path, live string, fields url.Values) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, origin+path, strings.NewReader(fields.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", origin)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "__Host-sid="+live)
	resp, err := e.edge.Do(req)
	if err != nil {
		t.Fatalf("web form request: %v", err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode, resp.Header.Get("Location")
}

// list reads the session's live connections on origin from core's internal API (as web does).
func (e *e2eEnv) list(t *testing.T, origin, live string) []string {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, e.coreURL+"/internal/v1/mcp/connections?origin="+url.QueryEscape(origin), nil)
	req.Header.Set("Authorization", "Session "+live)
	resp, err := e.core.Do(req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	defer resp.Body.Close()
	var out struct {
		Connections []struct {
			URL string `json:"url"`
		} `json:"connections"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&out) != nil {
		t.Fatalf("list on %s: %d", origin, resp.StatusCode)
	}
	var urls []string
	for _, c := range out.Connections {
		urls = append(urls, c.URL)
	}
	return urls
}

// create creates a connection through web's form route on the edge and returns its URL.
func (e *e2eEnv) create(t *testing.T, origin, live, csrf string) string {
	t.Helper()
	before := e.list(t, origin, live)
	code, loc := e.form(t, origin, "/connections/mcp/create", live, url.Values{"csrf_token": {csrf}})
	if code != http.StatusSeeOther || loc != "/connections/mcp?status=created" {
		t.Fatalf("web create on %s: %d %q", origin, code, loc)
	}
	var added []string
	for _, u := range e.list(t, origin, live) {
		if !slices.Contains(before, u) {
			added = append(added, u)
		}
	}
	if len(added) != 1 || !strings.HasPrefix(added[0], origin+ConnectionPathPrefix) {
		t.Fatalf("web create on %s added %v", origin, added)
	}
	return added[0]
}

// e2eRT adds the bearer to the SDK client's requests through the edge.
type e2eRT struct {
	base http.RoundTripper
	tok  string
}

func (r e2eRT) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+r.tok)
	return r.base.RoundTrip(req)
}

func TestMCPE2E(t *testing.T) {
	e := setupE2E(t)
	ctx := context.Background()
	h := e.h
	var a1Tok, a2Tok, b1Tok, a3URL string

	t.Run("E2E-MCP-1 activation and all five tools through the TLS edge", func(t *testing.T) {
		// Discovery: no token is 401 with the per-connection metadata URL; the metadata names the resource and issuer.
		code, hdr, _ := e.post(t, h.A1, "", `{"jsonrpc":"2.0","id":1,"method":"ping"}`)
		id := strings.TrimPrefix(h.A1, h.OriginA+ConnectionPathPrefix)
		meta := h.OriginA + MetadataPathPrefix + id
		if code != http.StatusUnauthorized || hdr.Get("WWW-Authenticate") != `Bearer resource_metadata="`+meta+`"` {
			t.Fatalf("no token: %d %q", code, hdr.Get("WWW-Authenticate"))
		}
		resp, err := e.edge.Get(meta)
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&doc)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK || doc["resource"] != h.A1 || !slices.Equal(e2eStrings(doc["authorization_servers"]), []string{e.issuer}) {
			t.Fatalf("metadata: %d %v", resp.StatusCode, doc)
		}
		// The official SDK client activates the pending connection and sees exactly the five tools.
		a1Tok = e.token(t, h.A1, e2eClient, nil)
		cctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		tr := &sdk.StreamableClientTransport{Endpoint: h.A1, HTTPClient: &http.Client{Transport: e2eRT{base: e.edge.Transport, tok: a1Tok}},
			MaxRetries: -1, DisableStandaloneSSE: true}
		cs, err := sdk.NewClient(&sdk.Implementation{Name: "g014-e2e", Version: "0.0.0"}, nil).Connect(cctx, tr, nil)
		if err != nil {
			t.Fatalf("SDK connect through the edge: %v", err)
		}
		defer cs.Close()
		if init := cs.InitializeResult(); init == nil || !slices.Contains(ProtocolVersions, init.ProtocolVersion) {
			t.Fatalf("initialize: %+v", init)
		}
		tl, err := cs.ListTools(cctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, tool := range tl.Tools {
			names = append(names, tool.Name)
		}
		slices.Sort(names)
		want := slices.Clone(ToolNames)
		slices.Sort(want)
		if !slices.Equal(names, want) {
			t.Fatalf("catalog = %v", names)
		}
		if u := e.mustCall(t, h.A1, a1Tok, ToolGetCurrentUser, nil); u["user_id"] != e.user {
			t.Fatalf("get_current_user: %v", u)
		}
		c := e.mustCall(t, h.A1, a1Tok, ToolGetCurrentContext, nil)
		if c["organization_id"] != e.orgA || c["organization_name"] != "Org A" || e2eGroup(c) != "" {
			t.Fatalf("get_current_context: %v", c)
		}
		gl := e.mustCall(t, h.A1, a1Tok, ToolListGroups, nil)
		groups, _ := gl["groups"].([]any)
		if len(groups) != 1 || groups[0].(map[string]any)["group_id"] != e2eResearch || gl["next_cursor"] != nil {
			t.Fatalf("list_groups: %v", gl)
		}
		if s := e.mustCall(t, h.A1, a1Tok, ToolSwitchGroup, map[string]any{"group_id": e2eResearch}); e2eGroup(s) != e2eResearch {
			t.Fatalf("switch_group: %v", s)
		}
		e.wantToolError(t, h.A1, a1Tok, ToolSwitchGroup, map[string]any{"group_id": e2eDesign}, "group_not_selectable")
		// logout, on a further connection created through web's setup route on the edge: terminal.
		a3URL = e.create(t, h.OriginA, h.LiveA, h.CSRFA)
		a3Tok := e.token(t, a3URL, e2eClient, nil)
		if out := e.mustCall(t, a3URL, a3Tok, ToolLogout, nil); out["outcome"] != "terminated" {
			t.Fatalf("logout: %v", out)
		}
		e.want401(t, a3URL, a3Tok, "after the logout tool")
		e.want401(t, a3URL, e.token(t, a3URL, e2eClient, nil), "a fresh token after the logout tool")
		if slices.Contains(e.list(t, h.OriginA, h.LiveA), a3URL) {
			t.Fatal("a logged-out connection is still listed")
		}
	})

	t.Run("E2E-MCP-2 separate connections keep isolated stored state; one URL shares it", func(t *testing.T) {
		// A second token family on the same URL (same client, new provider session) shares the stored selection.
		if c := e.mustCall(t, h.A1, e.token(t, h.A1, e2eClient, nil), ToolGetCurrentContext, nil); e2eGroup(c) != e2eResearch {
			t.Fatalf("second family on the same URL: %v", c)
		}
		a2Tok = e.token(t, h.A2, e2eClient, nil)
		if c := e.mustCall(t, h.A2, a2Tok, ToolGetCurrentContext, nil); c["organization_id"] != e.orgA || e2eGroup(c) != "" {
			t.Fatalf("second connection on A: %v", c)
		}
		b1Tok = e.token(t, h.B1, e2eClient, nil)
		c := e.mustCall(t, h.B1, b1Tok, ToolGetCurrentContext, nil)
		if c["organization_id"] != e.orgB || c["organization_name"] != "Org B" || e2eGroup(c) != "" {
			t.Fatalf("connection on B: %v", c)
		}
		gl := e.mustCall(t, h.B1, b1Tok, ToolListGroups, nil)
		if groups, _ := gl["groups"].([]any); len(groups) != 1 || groups[0].(map[string]any)["group_id"] != e2eOperations {
			t.Fatalf("list_groups on B: %v", gl)
		}
		e.wantToolError(t, h.B1, b1Tok, ToolSwitchGroup, map[string]any{"group_id": e2eResearch}, "group_not_selectable")
		if s := e.mustCall(t, h.B1, b1Tok, ToolSwitchGroup, map[string]any{"group_id": e2eOperations}); e2eGroup(s) != e2eOperations {
			t.Fatalf("switch_group on B: %v", s)
		}
		// Neither selection leaked: A1 keeps research, A2 none.
		if c := e.mustCall(t, h.A1, a1Tok, ToolGetCurrentContext, nil); e2eGroup(c) != e2eResearch {
			t.Fatalf("A1 after B's switch: %v", c)
		}
		if c := e.mustCall(t, h.A2, a2Tok, ToolGetCurrentContext, nil); e2eGroup(c) != "" {
			t.Fatalf("A2 after the other switches: %v", c)
		}
	})

	t.Run("E2E-MCP-3 cross-host and cross-connection replay is refused", func(t *testing.T) {
		idA1 := strings.TrimPrefix(h.A1, h.OriginA+ConnectionPathPrefix)
		onB := h.OriginB + ConnectionPathPrefix + idA1
		e.want401(t, onB, a1Tok, "A1's token on B's host")
		e.want401(t, onB, e.token(t, onB, e2eClient, nil), "A1's UUID on B's host with an exact B audience")
		e.want401(t, h.A2, a1Tok, "A1's token on A2")
		e.want401(t, h.A1, b1Tok, "B1's token on A1")
		e.want401(t, h.A1, e.token(t, h.A1, e2eOtherClient, nil), "another client on the activated A1")
		e.want401(t, h.A1, e.token(t, h.A1, e2eClient, map[string]any{"iss": e.issuer + "/other"}), "another issuer")
		e.want401(t, h.A1, e.token(t, h.A1, e2eClient, map[string]any{"exp": time.Now().Add(-2 * time.Minute).Unix()}), "an expired token")
		e.want401(t, h.A1, identitytest.Tamper(a1Tok), "a tampered token")
		// The legitimate token still works afterwards.
		e.mustCall(t, h.A1, a1Tok, ToolGetCurrentContext, nil)
	})

	t.Run("E2E-MCP-4 the edge exposes only the MCP paths on tenant hosts", func(t *testing.T) {
		for _, p := range []string{"/internal/v1/mcp/connections", "/internal/v1/auth/session", "/healthz", "/mcp/x", "/.well-known/oauth-protected-resource"} {
			resp, err := e.edge.Get(h.OriginA + p)
			if err != nil {
				t.Fatal(err)
			}
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("%s at the edge: %d, want 404 (web)", p, resp.StatusCode)
			}
		}
		req, _ := http.NewRequest(http.MethodPost, h.OriginA+"/internal/v1/mcp/connections", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := e.edge.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			t.Fatalf("internal create reachable at the edge: %d", resp.StatusCode)
		}
		// The attack origin never serves MCP.
		resp, err = e.edge.Get("https://attacker.localhost:543" + ConnectionPathPrefix + strings.TrimPrefix(h.A1, h.OriginA+ConnectionPathPrefix))
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("attacker origin MCP path: %d", resp.StatusCode)
		}
	})

	t.Run("E2E-MCP-5 web revoke and sign-out everywhere are terminal; signing out this browser leaves connections", func(t *testing.T) {
		id := strings.TrimPrefix(h.A2, h.OriginA+ConnectionPathPrefix)
		if code, loc := e.form(t, h.OriginA, "/connections/mcp/revoke", h.LiveA, url.Values{"csrf_token": {h.CSRFA}, "connection_id": {id}}); code != http.StatusSeeOther || loc != "/connections/mcp?status=revoked" {
			t.Fatalf("web revoke: %d %q", code, loc)
		}
		e.want401(t, h.A2, a2Tok, "after the web revoke")
		e.want401(t, h.A2, e.token(t, h.A2, e2eClient, nil), "a fresh token after the web revoke")
		// Signing out the spare browser session (current session only) leaves the connections working.
		if code, _ := e.form(t, h.OriginA, "/auth/logout", h.SpareA, url.Values{"csrf_token": {h.SpareCSRFA}}); code != http.StatusSeeOther {
			t.Fatalf("browser sign-out: %d", code)
		}
		e.mustCall(t, h.A1, a1Tok, ToolGetCurrentContext, nil)
		e.mustCall(t, h.B1, b1Tok, ToolGetCurrentContext, nil)
		// Sign out everywhere ends every connection of the user, on both hosts.
		if code, _ := e.form(t, h.OriginA, "/auth/logout-all", h.LiveA, url.Values{"csrf_token": {h.CSRFA}}); code != http.StatusSeeOther {
			t.Fatalf("sign out everywhere: %d", code)
		}
		e.want401(t, h.A1, a1Tok, "A1 after sign out everywhere")
		e.want401(t, h.B1, b1Tok, "B1 after sign out everywhere")
		e.want401(t, h.A1, e.token(t, h.A1, e2eClient, nil), "a fresh A1 token after sign out everywhere")
	})

	t.Run("E2E-MCP-7 every minted token is registered for the scans", func(t *testing.T) {
		if n := e.minted.Load(); n < 12 {
			t.Fatalf("only %d tokens minted", n)
		}
	})
}

func e2eStrings(v any) []string {
	a, _ := v.([]any)
	var out []string
	for _, x := range a {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}
