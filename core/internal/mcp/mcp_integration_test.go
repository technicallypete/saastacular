//go:build integration

package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"core/internal/domains"
	"core/internal/hostname"
	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/identity/workos"
	"core/internal/observability"
	"core/internal/tenancy"
)

// G014 B-MCP-IT (Plan r8 §10.3): IT-MCP-1..14 (IT-MCP-7's ordering cases O-MCP-1..4, IT-MCP-12(d) and E-MCP-1..3 run
// in core/internal/tenancy/access, which owns the inert hooks) and TestMCPRecovery. One environment per top-level test.

func TestMCPTransport(t *testing.T) {
	e := setupMCP(t)
	ctx := context.Background()

	// IT-MCP-1: discovery and transport through the actual HTTP path, with the official SDK client.
	t.Run("IT-MCP-1 discovery, stateless metadata and transport", func(t *testing.T) {
		pending := e.create(t, "memberA1", itHostA)
		active := e.create(t, "memberA1", itHostA)
		e.activate(t, active)
		revoked := e.create(t, "memberA1", itHostA)
		e.activate(t, revoked)
		credA := e.cred(t, "memberA1", itHostA)
		if code, _ := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/revoke", credA,
			map[string]string{"origin": "https://" + itHostA, "connection_id": revoked.id, "csrf_token": e.csrf(t, credA)}); code != http.StatusOK {
			t.Fatalf("revoke: %d", code)
		}
		unknown := conn{id: "7f000000-0000-4000-8000-0000000000aa", host: itHostA, url: resourceURL(itHostA, "7f000000-0000-4000-8000-0000000000aa"), owner: e.users["memberA1"]}
		meta := func(h *Handler, id string) (int, string) {
			req, _ := http.NewRequest(http.MethodGet, e.srv.URL+MetadataPathPrefix+id, nil)
			req.Host = itHostA
			w := &recorder{h: http.Header{}}
			h.Wrap(http.NotFoundHandler()).ServeHTTP(w, req)
			return w.code, w.body.String()
		}
		var docs []string
		for _, c := range []conn{pending, active, unknown, revoked} {
			code, doc := meta(e.h, c.id)
			if code != http.StatusOK {
				t.Fatalf("metadata of %s: %d", c.id, code)
			}
			docs = append(docs, strings.Replace(doc, c.id, "ID", 1))
		}
		if docs[0] != docs[1] || docs[1] != docs[2] || docs[2] != docs[3] {
			t.Fatalf("metadata differs by connection state: %v", docs)
		}
		// The same metadata while the store is unavailable (no lookup); protected calls then answer 503.
		down := e.handler(t, map[string]tenancy.OrganizationID{itHostA: e.orgA}, nil)
		closedPool := itPool(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		closedPool.Close()
		downID := identity.NewService(identity.NewStore(closedPool), identity.Config{Issuer: e.fake.Issuer, ClientID: itClientG004, Now: e.clk.Now})
		down.id = downID
		if code, doc := meta(down, pending.id); code != http.StatusOK || strings.Replace(doc, pending.id, "ID", 1) != docs[0] {
			t.Fatalf("metadata with the store down: %d %s", code, doc)
		}
		req, _ := http.NewRequest(http.MethodPost, ConnectionPathPrefix+active.id, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		req.Host = itHostA
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Authorization", "Bearer "+e.tok(active))
		w := &recorder{h: http.Header{}}
		down.Wrap(http.NotFoundHandler()).ServeHTTP(w, req)
		if w.code != http.StatusServiceUnavailable {
			t.Fatalf("protected call with the store down: %d", w.code)
		}
		// No token: 401 with the per-connection discovery header; unknown and revoked connections are 401 and run nothing.
		code, hdr, _ := e.rawPost(t, itHostA, pending.id, "", `{"jsonrpc":"2.0","id":1,"method":"ping"}`)
		if code != http.StatusUnauthorized || hdr.Get("WWW-Authenticate") != `Bearer resource_metadata="`+metadataURL(itHostA, pending.id)+`"` {
			t.Fatalf("no token: %d %q", code, hdr.Get("WWW-Authenticate"))
		}
		e.want401(t, unknown, e.token("memberA1", unknown.url, nil), "unknown connection")
		e.want401(t, revoked, e.tok(revoked), "revoked connection")
		// 404 for an unregistered host or a malformed path.
		if code, _, _ := e.rawPost(t, "org-c.localhost:543", active.id, e.tok(active), "{}"); code != http.StatusNotFound {
			t.Fatalf("unregistered host: %d", code)
		}
		if code, _, _ := e.rawPost(t, itHostA, active.id+"/x", e.tok(active), "{}"); code != http.StatusNotFound {
			t.Fatalf("malformed path: %d", code)
		}
		// The official SDK client: negotiation, the exact catalog, a tool call, and an unknown or retired tool refused.
		cs, err := e.dial(t, active, e.tok(active))
		if err != nil {
			t.Fatalf("SDK connect: %v", err)
		}
		defer cs.Close()
		init := cs.InitializeResult()
		if init == nil || init.ServerInfo.Name != ServerName || init.ServerInfo.Title != "Integration Product" ||
			!slices.Contains(ProtocolVersions, init.ProtocolVersion) || init.Capabilities.Tools == nil || init.Capabilities.Tools.ListChanged {
			t.Fatalf("initialize: %+v", init)
		}
		tl, err := cs.ListTools(ctx, nil)
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
		res, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: ToolGetCurrentContext})
		if err != nil || res.IsError {
			t.Fatalf("get_current_context: %v %+v", err, res)
		}
		for _, n := range []string{"account.switch_organization", "account.list_organizations", "nope"} {
			if _, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: n}); err == nil {
				t.Errorf("%s accepted", n)
			}
		}
		if err := cs.Ping(ctx, nil); err != nil {
			t.Fatalf("ping: %v", err)
		}
		// An unclaimed protocol version header is refused.
		req2, _ := http.NewRequest(http.MethodPost, e.srv.URL+ConnectionPathPrefix+active.id, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		req2.Host = itHostA
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Accept", "application/json, text/event-stream")
		req2.Header.Set("MCP-Protocol-Version", "2024-11-05")
		req2.Header.Set("Authorization", "Bearer "+e.tok(active))
		resp, err := http.DefaultClient.Do(req2)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("unclaimed protocol version: %d", resp.StatusCode)
		}
	})

	// IT-MCP-2: host and route.
	t.Run("IT-MCP-2 host, route and resolution reassignment", func(t *testing.T) {
		c := e.create(t, "cross", itHostA)
		tok := e.activate(t, c)
		// The same UUID on B's host, with a token whose audience is that B URL: the registry maps B to Organization B,
		// where the subject has no such connection.
		onB := conn{id: c.id, host: itHostB, url: resourceURL(itHostB, c.id), owner: c.owner}
		e.want401(t, onB, e.token("cross", onB.url, nil), "connection reached on another host")
		// A's token presented on B's host fails the exact audience.
		e.want401(t, onB, tok, "A token on B host")
		// Forwarded headers are ignored: they never select a host.
		req, _ := http.NewRequest(http.MethodPost, e.srv.URL+ConnectionPathPrefix+c.id, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		req.Host = "org-c.localhost:543"
		req.Header.Set("X-Forwarded-Host", itHostA)
		req.Header.Set("Forwarded", "host="+itHostA)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("forwarded host: %d", resp.StatusCode)
		}
		// Registry reassignment (A's host now bound to B): the existing connection is never retargeted.
		moved := e.handler(t, map[string]tenancy.OrganizationID{itHostA: e.orgB}, nil)
		req, _ = http.NewRequest(http.MethodPost, ConnectionPathPrefix+c.id, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		req.Host = itHostA
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Authorization", "Bearer "+e.tok(c))
		w := &recorder{h: http.Header{}}
		moved.Wrap(http.NotFoundHandler()).ServeHTTP(w, req)
		if w.code != http.StatusUnauthorized {
			t.Fatalf("reassigned registry: %d", w.code)
		}

		// G016 (Plan r4 §1.6): the same rules through the real durable resolver: Organization platform labels (G057: the
		// label, not the slug, is the platform host; fixed synthetic labels here) and a verified, TLS-ready custom alias in
		// the database, resolved per request (no registry, no cache).
		n := itSeq.Add(1)
		labelA, labelB := fmt.Sprintf("mcp-a-%d", n), fmt.Sprintf("mcp-b-%d", n)
		labelHost, otherHost, alias := labelA+".localhost:543", labelB+".localhost:543", fmt.Sprintf("mcp-%d.alias.example:543", n)
		exec := func(sql string, args ...any) {
			t.Helper()
			if _, err := e.setup.Exec(ctx, sql, args...); err != nil {
				t.Fatal(err)
			}
		}
		exec(`UPDATE app.organizations SET platform_label = $2 WHERE id = $1::text::uuid`, e.orgA.String(), labelA)
		exec(`UPDATE app.organizations SET platform_label = $2 WHERE id = $1::text::uuid`, e.orgB.String(), labelB)
		var domainID string
		if err := e.setup.QueryRow(ctx, `INSERT INTO app.organization_domains (id, organization_id, hostname, generation, state, challenge_value,
  challenge_expires_at, verified_at, verified_until, tls_state, tls_checked_at, tls_not_after, created_by)
VALUES (gen_random_uuid(), $1::text::uuid, $2, 1, 'active', repeat('a', 52), now() + interval '1 hour', now(), now() + interval '1 day',
        'ready', now(), now() + interval '30 days', $3::text::uuid) RETURNING id::text`,
			e.orgA.String(), strings.TrimSuffix(alias, ":543"), e.users["ownerA"].uid.String()).Scan(&domainID); err != nil {
			t.Fatal(err)
		}
		e.serveWith(t, e.lookupHandler(t, domains.NewResolver(e.pool, hostname.Config{BaseDomain: "localhost", CNAMETarget: "edge.localhost",
			Port: "543"}), nil))
		// G016 A2: the identity resolver routes the label and alias origins to Organization A, as the durable resolver does
		// from these rows; cross holds a credential issued at each.
		e.route(t, labelHost, e.orgA)
		e.route(t, alias, e.orgA)
		e.bind(t, "cross", labelHost, alias)
		rc := e.create(t, "cross", labelHost)
		rtok := e.activate(t, rc)
		e.mustCall(t, rc, rtok, ToolGetCurrentContext, nil)
		// A at B unchanged: the same UUID on B's label host (B's audience) reaches no connection.
		onOther := conn{id: rc.id, host: otherHost, url: resourceURL(otherHost, rc.id), owner: rc.owner}
		e.want401(t, onOther, e.token("cross", onOther.url, nil), "resolver: connection reached on another Organization's host")
		// A verified custom alias of Organization A routes like its label host; the create URL keeps the exact alias origin.
		ac := e.create(t, "cross", alias)
		if !strings.HasPrefix(ac.url, "https://"+alias+"/") {
			t.Fatalf("alias create URL %s", ac.url)
		}
		atok := e.activate(t, ac)
		e.mustCall(t, ac, atok, ToolGetCurrentContext, nil)
		// A removed alias stops routing at that commit: 404 with no store access (the connection row is untouched).
		st0, _, _, v0 := e.state(t, ac.id)
		exec(`UPDATE app.organization_domains SET state = 'removed', removed_at = now(), tls_state = 'withdrawing', version = version + 1
WHERE id = $1::text::uuid`, domainID)
		if code, _, _ := e.rawPost(t, alias, ac.id, atok, `{"jsonrpc":"2.0","id":1,"method":"ping"}`); code != http.StatusNotFound {
			t.Fatalf("removed alias: %d, want 404", code)
		}
		if st1, _, _, v1 := e.state(t, ac.id); st1 != st0 || v1 != v0 {
			t.Fatalf("the removed-host request changed the connection: %s/%d -> %s/%d", st0, v0, st1, v1)
		}
		// A reassigned label (A's label host now resolving to B; setup only, no runtime role writes labels) never retargets
		// the existing connection: 401.
		exec(`UPDATE app.organizations SET platform_label = NULL WHERE id = $1::text::uuid`, e.orgA.String())
		exec(`UPDATE app.organizations SET platform_label = $2 WHERE id = $1::text::uuid`, e.orgB.String(), labelA)
		e.want401(t, rc, rtok, "resolver: reassigned label host")
		exec(`UPDATE app.organizations SET platform_label = NULL WHERE id = ANY($1::text[]::uuid[])`, []string{e.orgA.String(), e.orgB.String()})
	})

	// IT-MCP-3: setup and activation.
	t.Run("IT-MCP-3 setup and activation", func(t *testing.T) {
		// A non-member is refused at setup with the generic answer.
		nonCred := e.cred(t, "nonmember", itHostA)
		if code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections", nonCred,
			map[string]string{"origin": "https://" + itHostA, "csrf_token": e.csrf(t, nonCred)}); code != http.StatusForbidden || out["error"] != "not_available" {
			t.Fatalf("non-member setup: %d %v", code, out)
		}
		// Setup without CSRF, from an unregistered origin, or without a session.
		u := e.users["memberA1"]
		uCred := e.cred(t, "memberA1", itHostA)
		for name, c := range map[string]struct {
			cred string
			body map[string]string
			code int
			err  string
		}{
			"no csrf":    {uCred, map[string]string{"origin": "https://" + itHostA, "csrf_token": ""}, http.StatusForbidden, "csrf_failed"},
			"bad origin": {uCred, map[string]string{"origin": "https://org-c.localhost:543", "csrf_token": e.csrf(t, uCred)}, http.StatusForbidden, "origin_rejected"},
			"no session": {"", map[string]string{"origin": "https://" + itHostA, "csrf_token": "x"}, http.StatusUnauthorized, "unauthenticated"},
		} {
			if code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections", c.cred, c.body); code != c.code || out["error"] != c.err {
				t.Errorf("%s: %d %v", name, code, out)
			}
		}
		// G016 A2: a credential bound to A's origin and Organization is unauthenticated at B's origin, for create and for
		// the listing, even for a member of both Organizations (its own CSRF token); nothing is created.
		crossA := e.cred(t, "cross", itHostA)
		connections := func() int {
			t.Helper()
			var n int
			if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.mcp_connections`).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		before := connections()
		if code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections", crossA,
			map[string]string{"origin": "https://" + itHostB, "csrf_token": e.csrf(t, crossA)}); code != http.StatusUnauthorized || out["error"] != "unauthenticated" {
			t.Errorf("A-bound credential creating at B's origin: %d %v", code, out)
		}
		if code, out := e.internal(t, http.MethodGet, "/internal/v1/mcp/connections?"+url.Values{"origin": {"https://" + itHostB}}.Encode(), crossA, nil); code != http.StatusUnauthorized || out["error"] != "unauthenticated" {
			t.Errorf("A-bound credential listing at B's origin: %d %v", code, out)
		}
		if after := connections(); after != before {
			t.Errorf("the refused A-bound requests at B's origin changed the connections: %d -> %d", before, after)
		}
		// A member's pending connection; a wrong subject never activates it.
		c := e.create(t, "memberA1", itHostA)
		e.want401(t, c, e.token("cross", c.url, nil), "wrong subject")
		if st, _, _, _ := e.state(t, c.id); st != identity.MCPStatePending {
			t.Fatalf("wrong subject changed the state: %s", st)
		}
		// Token-type confusion: a G004 access token and an ID-token-shaped token are refused.
		g4 := e.fake.Sign(e.fake.ClaimsFor(identitytest.Grant{UserID: u.sub, SID: "session_01CONFUSION"}))
		e.want401(t, c, g4, "G004 access token")
		e.want401(t, c, e.token("memberA1", itConnect, nil), "ID-token shape (client audience)")
		// Membership lost before activation: refused, still pending; restored: the same connection activates.
		if err := e.model.RemoveGroupMembership(ctx, e.orgA, e.a1.ID, u.uid); err != nil {
			t.Fatal(err)
		}
		if err := e.model.RemoveOrganizationMembership(ctx, e.orgA, u.uid); err != nil {
			t.Fatal(err)
		}
		e.want401(t, c, e.tok(c), "non-member activation")
		if st, _, _, _ := e.state(t, c.id); st != identity.MCPStatePending {
			t.Fatalf("refused activation changed the state: %s", st)
		}
		e.grantOrg(t, e.orgA, "memberA1", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "memberA1", tenancy.GroupMember)
		tok := e.activate(t, c)
		// A second client after activation is refused; the bound client stays.
		e.want401(t, c, e.token("memberA1", c.url, map[string]any{"client_id": itConnect2}), "second client")
		if _, client, _, _ := e.state(t, c.id); client != itConnect {
			t.Fatalf("client rebound: %s", client)
		}
		e.mustCall(t, c, tok, ToolGetCurrentUser, nil)
		// Expired pending setup is terminal.
		late := e.create(t, "memberA1", itHostA)
		e.clk.Advance(31 * time.Minute)
		e.want401(t, late, e.tok(late), "expired pending setup")
		e.clk.Advance(-31 * time.Minute)
		if st, _, _, _ := e.state(t, late.id); st != identity.MCPStatePending {
			t.Fatalf("expired setup activated: %s", st)
		}
		// The activation audit event: one mcp.connection_established, attributed to the owner.
		if got := e.events(t, c.id); !slices.Equal(got, []string{"mcp.connection_established:-:user"}) {
			t.Fatalf("activation events: %v", got)
		}
		// G063: create accepts an optional friendly name (normalized; the legacy request shape stays valid), and every listed
		// row carries display_name: the name, or null for an unnamed connection. The name is never in the create answer.
		nameCred := e.cred(t, "memberA1", itHostA)
		code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections", nameCred, map[string]any{"origin": "https://" + itHostA,
			"csrf_token": e.csrf(t, nameCred), "display_name": "\u00A0 Claude Desktop \u3000"})
		if code != http.StatusCreated {
			t.Fatalf("named create: %d %v", code, out)
		}
		namedConn, _ := out["connection"].(map[string]any)
		if _, leaked := namedConn["display_name"]; leaked || len(namedConn) != 3 {
			t.Errorf("the create answer changed shape: %v", namedConn)
		}
		namedID, _ := namedConn["connection_id"].(string)
		code, out = e.internal(t, http.MethodGet, "/internal/v1/mcp/connections?"+url.Values{"origin": {"https://" + itHostA}}.Encode(), nameCred, nil)
		if code != http.StatusOK {
			t.Fatalf("listing: %d %v", code, out)
		}
		rows, _ := out["connections"].([]any)
		seen := 0
		for _, x := range rows {
			m := x.(map[string]any)
			v, has := m["display_name"]
			if !has {
				t.Errorf("listed row %v has no display_name key", m["connection_id"])
			}
			switch m["connection_id"] {
			case namedID:
				seen++
				if v != "Claude Desktop" {
					t.Errorf("named row: display_name %v, want the trimmed name", v)
				}
			case c.id, late.id:
				seen++
				if v != nil {
					t.Errorf("legacy unnamed row %v: display_name %v, want null", m["connection_id"], v)
				}
			}
		}
		if seen != 3 {
			t.Errorf("listing saw %d of the named and two legacy rows", seen)
		}
		setupListingPages(t, e)
	})

	// IT-MCP-4: binding agreement.
	t.Run("IT-MCP-4 endpoint, credential and connection binding", func(t *testing.T) {
		c1 := e.create(t, "cross", itHostA)
		t1 := e.activate(t, c1)
		c2 := e.create(t, "cross", itHostA)
		cb := e.create(t, "cross", itHostB)
		tb := e.activate(t, cb)
		// C1's audience on C2's path, and B's token at A, are refused; neither connection changes.
		e.want401(t, c2, t1, "C1 audience on C2")
		e.want401(t, c1, tb, "B token at A")
		if st, _, _, _ := e.state(t, c2.id); st != identity.MCPStatePending {
			t.Fatalf("C2 changed: %s", st)
		}
		// A separately issued exact-audience token for C2 (same user and client) is evaluated only under C2.
		t2 := e.tok(c2)
		e.mustCall(t, c2, t2, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		if _, _, g, _ := e.state(t, c1.id); g != "" {
			t.Fatalf("C1 selection changed by a C2 call: %s", g)
		}
		// A token whose audience is a revoked connection's resource never reaches another connection.
		e.mustCall(t, c1, t1, ToolLogout, nil)
		e.want401(t, c1, t1, "revoked connection")
		e.want401(t, c2, t1, "revoked connection's token on C2")
		e.mustCall(t, c2, t2, ToolGetCurrentContext, nil)
		// A mismatch on the subject's own connection (here a second client on the bound connection) is a recorded
		// resource_mismatch denial; an audience mismatch is refused during verification, before any mapping.
		e.want401(t, c2, e.token("cross", c2.url, map[string]any{"client_id": itConnect2}), "second client on C2")
		var n int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND reason = 'resource_mismatch'
  AND operation = 'mcp.request' AND organization_id IS NULL AND actor_user_id = $1::text::uuid`, e.users["cross"].id).Scan(&n); err != nil || n == 0 {
			t.Fatalf("resource_mismatch events: %d %v", n, err)
		}
	})

	// IT-MCP-5: identity and context output, fresh scope per call.
	t.Run("IT-MCP-5 identity and context output", func(t *testing.T) {
		c := e.create(t, "ownerA", itHostA)
		tok := e.activate(t, c)
		u := e.mustCall(t, c, tok, ToolGetCurrentUser, nil)
		if u["user_id"] != e.users["ownerA"].id || u["display_name"] != "Fixture User" || len(u) != 2 {
			t.Fatalf("get_current_user: %v", u)
		}
		ctxOut := e.mustCall(t, c, tok, ToolGetCurrentContext, nil)
		if ctxOut["organization_id"] != e.orgA.String() || ctxOut["organization_name"] != "Organization A" || ctxOut["group"] != nil ||
			ctxOut["context_version"] != float64(0) || len(ctxOut) != 4 {
			t.Fatalf("get_current_context: %v", ctxOut)
		}
		for _, v := range []string{e.users["ownerA"].sub, itConnect, tok} {
			for _, out := range []map[string]any{u, ctxOut} {
				for _, x := range out {
					if s, ok := x.(string); ok && s == v {
						t.Fatalf("output carries a provider value")
					}
				}
			}
		}
	})

	// IT-MCP-6: listing.
	t.Run("IT-MCP-6 selectable Groups", func(t *testing.T) {
		owner := e.create(t, "ownerA", itHostA)
		to := e.activate(t, owner)
		member := e.create(t, "memberA1", itHostA)
		tm := e.activate(t, member)
		ids := func(out map[string]any) []string {
			var s []string
			for _, g := range out["groups"].([]any) {
				s = append(s, g.(map[string]any)["group_id"].(string))
			}
			return s
		}
		wantOwner := []string{e.a1.ID.String(), e.a2.ID.String()}
		slices.Sort(wantOwner)
		if got := ids(e.mustCall(t, owner, to, ToolListGroups, nil)); !slices.Equal(got, wantOwner) {
			t.Fatalf("staff listing: %v", got)
		}
		if got := ids(e.mustCall(t, member, tm, ToolListGroups, nil)); !slices.Equal(got, []string{e.a1.ID.String()}) {
			t.Fatalf("client listing: %v", got)
		}
		// Pagination with an opaque cursor; a cursor of another Organization is invalid.
		p1 := e.mustCall(t, owner, to, ToolListGroups, map[string]any{"limit": 1})
		next, _ := p1["next_cursor"].(string)
		if len(ids(p1)) != 1 || next == "" {
			t.Fatalf("page 1: %v", p1)
		}
		p2 := e.mustCall(t, owner, to, ToolListGroups, map[string]any{"limit": 1, "cursor": next})
		if got := append(ids(p1), ids(p2)...); !slices.Equal(got, wantOwner) || p2["next_cursor"] != nil {
			t.Fatalf("pages: %v %v", got, p2)
		}
		cb := e.create(t, "cross", itHostB)
		tb := e.activate(t, cb)
		e.wantToolError(t, cb, tb, ToolListGroups, map[string]any{"cursor": next}, CodeInvalidInput)
		e.wantToolError(t, owner, to, ToolListGroups, map[string]any{"limit": 0}, CodeInvalidInput)
		e.wantToolError(t, owner, to, ToolListGroups, map[string]any{"organization_id": e.orgB.String()}, CodeInvalidInput)
	})

	// IT-MCP-7: selection refusals and no-ops (the ordering cases run in the access package).
	t.Run("IT-MCP-7 selection", func(t *testing.T) {
		c := e.create(t, "memberA1", itHostA)
		tok := e.activate(t, c)
		out := e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		if g := out["group"].(map[string]any); g["group_id"] != e.a1.ID.String() || out["context_version"] != float64(1) {
			t.Fatalf("select: %v", out)
		}
		// Selecting the current Group is a no-op (no version change, no event).
		if out := e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()}); out["context_version"] != float64(1) {
			t.Fatalf("no-op select: %v", out)
		}
		// Unknown, foreign and unselectable targets are the same refusal; the selection is unchanged.
		for _, g := range []string{e.a2.ID.String(), e.b1.ID.String(), "7f000000-0000-4000-8000-0000000000bb"} {
			e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": g}, CodeGroupNotSelectable)
		}
		e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": "NOT-A-UUID"}, CodeInvalidInput)
		e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{}, CodeInvalidInput)
		e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String(), "organization_id": e.orgB.String()}, CodeInvalidInput)
		if _, _, g, v := e.state(t, c.id); g != e.a1.ID.String() || v != 1 {
			t.Fatalf("refusals changed the selection: %s %d", g, v)
		}
		// Explicit clear; clearing again is a no-op.
		if out := e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": nil}); out["group"] != nil || out["context_version"] != float64(2) {
			t.Fatalf("clear: %v", out)
		}
		if out := e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": nil}); out["context_version"] != float64(2) {
			t.Fatalf("no-op clear: %v", out)
		}
		if got := e.events(t, c.id); !slices.Equal(got, []string{"mcp.connection_established:-:user",
			"mcp.context_changed:" + e.a1.ID.String() + ":user", "mcp.context_changed:-:user"}) {
			t.Fatalf("selection events: %v", got)
		}
	})
}

func TestMCPLifecycle(t *testing.T) {
	e := setupMCP(t)
	ctx := context.Background()

	// IT-MCP-8: same URL versus separate URLs.
	t.Run("IT-MCP-8 same URL shares, separate URLs isolate stored state", func(t *testing.T) {
		c1 := e.create(t, "cross", itHostA)
		f1 := e.activate(t, c1)
		f2 := e.tok(c1) // a second token family of the same user and client for the same URL
		c2 := e.create(t, "cross", itHostA)
		g2 := e.activate(t, c2)
		e.mustCall(t, c1, f1, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		if out := e.mustCall(t, c1, f2, ToolGetCurrentContext, nil); out["group"] == nil {
			t.Fatal("the same URL does not share the selection")
		}
		if out := e.mustCall(t, c2, g2, ToolGetCurrentContext, nil); out["group"] != nil {
			t.Fatal("a separate URL shares the selection")
		}
		// Browser logout (current) leaves both connections; logging out C1 ends it for both families, not C2.
		crossA := e.cred(t, "cross", itHostA)
		if _, err := e.id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: e.csrf(t, crossA), Origin: "https://" + itHostA, Credential: crossA}); err != nil {
			t.Fatal(err)
		}
		e.bind(t, "cross", itHostA) // G016 A2: the explicit re-sign-in after the logout, at A's origin
		e.mustCall(t, c1, f1, ToolGetCurrentContext, nil)
		e.mustCall(t, c1, f2, ToolLogout, nil)
		e.want401(t, c1, f1, "first family after logout")
		e.want401(t, c1, f2, "second family after logout")
		e.mustCall(t, c2, g2, ToolGetCurrentContext, nil)

		// G016 (Plan r4 §1.6, IT-MCP-3/8 listing): through a current alias host of Organization A, the setup listing shows
		// the user's rows of every host of the Organization with their own host; a row created on the old host is revoked
		// through the alias; create keeps the exact alias origin.
		aliasHost := fmt.Sprintf("org-a-current-%d.localhost:543", itSeq.Add(1))
		e.serveWith(t, e.handler(t, map[string]tenancy.OrganizationID{itHostA: e.orgA, itHostB: e.orgB, aliasHost: e.orgA}, nil))
		// G016 A2: the alias origin is issued for Organization A, as the serving handler maps it; cross holds a credential
		// issued at it for the alias listing, revoke and create.
		e.route(t, aliasHost, e.orgA)
		e.bind(t, "cross", aliasHost)
		aliasCred := e.cred(t, "cross", aliasHost)
		old := e.create(t, "cross", itHostA)
		e.activate(t, old)
		code, out := e.internal(t, http.MethodGet, "/internal/v1/mcp/connections?"+url.Values{"origin": {"https://" + aliasHost}}.Encode(), aliasCred, nil)
		if code != http.StatusOK {
			t.Fatalf("alias listing: %d %v", code, out)
		}
		found := false
		cs, _ := out["connections"].([]any)
		for _, x := range cs {
			if m := x.(map[string]any); m["connection_id"] == old.id {
				found = true
				if ru, _ := m["url"].(string); !strings.HasPrefix(ru, "https://"+itHostA+"/") {
					t.Errorf("the old-host row is listed with URL %v, want its own host", m["url"])
				}
			}
		}
		if !found {
			t.Fatalf("the alias listing omits the old-host row: %v", out)
		}
		if code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/revoke", aliasCred, map[string]string{"origin": "https://" + aliasHost,
			"connection_id": old.id, "csrf_token": e.csrf(t, aliasCred)}); code != http.StatusOK {
			t.Fatalf("revoke of the old-host row through the alias: %d %v", code, out)
		}
		// A connection terminates through its session (mcp_connections.state is only pending|active).
		var revoked bool
		if err := e.setup.QueryRow(context.Background(), `SELECT s.revoked_at IS NOT NULL FROM app.mcp_connections c
JOIN app.sessions s ON s.id = c.session_id WHERE c.id = $1::text::uuid`, old.id).Scan(&revoked); err != nil || !revoked {
			t.Fatalf("the old-host row's session is not revoked (%v)", err)
		}
		if nc := e.create(t, "cross", aliasHost); !strings.HasPrefix(nc.url, "https://"+aliasHost+"/") {
			t.Fatalf("alias create URL %s, want the exact alias origin", nc.url)
		}
	})

	// IT-MCP-9: lifecycle.
	t.Run("IT-MCP-9 renewal keeps the URL; terminal states refuse", func(t *testing.T) {
		c := e.create(t, "memberA1", itHostA)
		tok := e.activate(t, c)
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		// Refresh (a new token, same audience, user and client) and reauthentication keep the binding and the selection.
		e.clk.Advance(2 * time.Hour)
		refreshed := e.tok(c)
		if out := e.mustCall(t, c, refreshed, ToolGetCurrentContext, nil); out["group"] == nil {
			t.Fatal("refresh lost the selection")
		}
		// Web revoke is terminal; a token the provider still treats as valid is refused locally (R5).
		// G016 A2: the 2h advance above ended memberA1's browser session (30m idle): explicit renewal at A's origin.
		e.bind(t, "memberA1", itHostA)
		uCred := e.cred(t, "memberA1", itHostA)
		if code, _ := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/revoke", uCred,
			map[string]string{"origin": "https://" + itHostA, "connection_id": c.id, "csrf_token": e.csrf(t, uCred)}); code != http.StatusOK {
			t.Fatalf("revoke: %d", code)
		}
		e.want401(t, c, e.tok(c), "after web revoke")
		// Idle and absolute expiry are terminal.
		idle := e.create(t, "memberA1", itHostA)
		ti := e.activate(t, idle)
		e.clk.Advance(31 * 24 * time.Hour)
		e.want401(t, idle, e.tok(idle), "idle expiry")
		_ = ti
		e.bind(t, "memberA1", itHostA) // G016 A2: explicit renewal after the 31-day advance
		abs := e.create(t, "memberA1", itHostA)
		e.activate(t, abs)
		for i := 0; i < 7; i++ { // renewed every 29 days until the 180-day absolute expiry
			e.clk.Advance(29 * 24 * time.Hour)
			if i < 6 {
				e.mustCall(t, abs, e.tok(abs), ToolGetCurrentContext, nil)
			}
		}
		e.want401(t, abs, e.tok(abs), "absolute expiry")
		// A disabled owner's connections end; re-enable revives none.
		e.bind(t, "dave", itHostA) // G016 A2: explicit renewal; the advances above ended dave's browser session
		d := e.create(t, "dave", itHostA)
		e.activate(t, d)
		if err := e.id.DisableUser(ctx, e.users["dave"].id); err != nil {
			t.Fatal(err)
		}
		if err := e.id.EnableUser(ctx, e.users["dave"].id); err != nil {
			t.Fatal(err)
		}
		e.want401(t, d, e.tok(d), "after disable and enable")
		e.bind(t, "dave", itHostA) // the explicit re-sign-in after the disable
		// G016 A2: this subtest moved the clock past every browser session's lifetime; explicit renewal of the
		// credentials the later subtests present (ownerA at A for IT-MCP-11, cross at A and B for IT-MCP-13).
		e.bind(t, "ownerA", itHostA)
		e.bind(t, "cross", itHostA, itHostB)
	})

	// IT-MCP-11: the authority matrix.
	t.Run("IT-MCP-11 membership loss, selected-Group loss and key-set outage", func(t *testing.T) {
		u := e.newUser(t, "matrix")
		e.grantOrg(t, e.orgA, "matrix", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a2.ID, "matrix", tenancy.GroupMember)
		c := e.create(t, "matrix", itHostA)
		tok := e.activate(t, c)
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a2.ID.String()})
		// Selected-Group access lost: normal calls fail closed; explicit clear works; never an auto-clear.
		if err := e.model.RemoveGroupMembership(ctx, e.orgA, e.a2.ID, u.uid); err != nil {
			t.Fatal(err)
		}
		for _, tool := range []string{ToolGetCurrentUser, ToolGetCurrentContext, ToolListGroups} {
			e.wantToolError(t, c, tok, tool, nil, CodeContextDenied)
		}
		if code, out := e.rpc(t, c, tok, "tools/list", map[string]any{}); code != http.StatusOK || out["error"].(map[string]any)["message"] != CodeContextDenied {
			t.Fatalf("tools/list after selected-Group loss: %d %v", code, out)
		}
		e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()}, CodeContextDenied)
		if _, _, g, _ := e.state(t, c.id); g != e.a2.ID.String() {
			t.Fatalf("auto-cleared: %s", g)
		}
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": nil})
		// Organization membership lost: every protected call and the clear are denied; logout still works.
		if err := e.model.RemoveOrganizationMembership(ctx, e.orgA, u.uid); err != nil {
			t.Fatal(err)
		}
		for _, tool := range []string{ToolGetCurrentUser, ToolGetCurrentContext, ToolListGroups} {
			e.wantToolError(t, c, tok, tool, nil, CodeContextDenied)
		}
		e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": nil}, CodeContextDenied)
		// Restored membership makes the same connection usable again (it was not revoked).
		e.grantOrg(t, e.orgA, "matrix", tenancy.OrgMember)
		e.mustCall(t, c, tok, ToolGetCurrentContext, nil)
		if err := e.model.RemoveOrganizationMembership(ctx, e.orgA, u.uid); err != nil {
			t.Fatal(err)
		}
		e.mustCall(t, c, tok, ToolLogout, nil)
		e.want401(t, c, tok, "after logout without membership")
		// Key set unavailable: 503 on protected calls, never anonymous.
		network, err := workos.NewConnectValidator(workos.ConnectConfig{Issuer: e.fake.ConnectIssuer(), Algorithms: []string{"RS256"}, Now: e.clk.Now})
		if err != nil {
			t.Fatal(err)
		}
		e.fake.FailJWKS(http.StatusServiceUnavailable)
		defer e.fake.FailJWKS(0)
		h := e.handler(t, map[string]tenancy.OrganizationID{itHostA: e.orgA}, network)
		c2 := e.create(t, "ownerA", itHostA)
		req, _ := http.NewRequest(http.MethodPost, ConnectionPathPrefix+c2.id, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		req.Host = itHostA
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Authorization", "Bearer "+e.tok(c2))
		w := &recorder{h: http.Header{}}
		h.Wrap(http.NotFoundHandler()).ServeHTTP(w, req)
		if w.code != http.StatusServiceUnavailable {
			t.Fatalf("key set down: %d", w.code)
		}
	})

	// IT-MCP-12 (a)-(c): live access changes without a new token.
	t.Run("IT-MCP-12 live access changes", func(t *testing.T) {
		staff := e.newUser(t, "staff")
		e.grantOrg(t, e.orgA, "staff", tenancy.OrgAdmin)
		c := e.create(t, "staff", itHostA)
		tok := e.activate(t, c)
		count := func() int { return len(e.mustCall(t, c, tok, ToolListGroups, nil)["groups"].([]any)) }
		if n := count(); n < 2 {
			t.Fatalf("staff sees %d Groups", n)
		}
		// (a, c) A downgrade applies to the next call: only own Groups (none).
		if _, err := e.model.ChangeOrganizationRole(ctx, e.orgA, staff.uid, tenancy.OrgMember); err != nil {
			t.Fatal(err)
		}
		if n := count(); n != 0 {
			t.Fatalf("after the downgrade: %d Groups", n)
		}
		// (b) The selected Group's binding is removed while selected: calls fail closed, clear succeeds, no auto-clear.
		e.grantGroup(t, e.orgA, e.a1.ID, "staff", tenancy.GroupMember)
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		if err := e.model.RemoveGroupMembership(ctx, e.orgA, e.a1.ID, staff.uid); err != nil {
			t.Fatal(err)
		}
		e.wantToolError(t, c, tok, ToolGetCurrentContext, nil, CodeContextDenied)
		if _, _, g, _ := e.state(t, c.id); g != e.a1.ID.String() {
			t.Fatal("auto-cleared")
		}
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": nil})
	})

	// IT-MCP-13: Organization-bound reads through runtime SQL.
	t.Run("IT-MCP-13 Organization-bound reads", func(t *testing.T) {
		ca := e.create(t, "cross", itHostA)
		ta := e.activate(t, ca)
		cb := e.create(t, "cross", itHostB)
		e.activate(t, cb)
		rt := itConn(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		count := func(org string) int {
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SELECT pg_catalog.set_config('app.user_id', $1, true), pg_catalog.set_config('app.organization_id', $2, true)`,
				e.users["cross"].id, org); err != nil {
				t.Fatal(err)
			}
			var n int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.mcp_connections WHERE id = ANY ($1::uuid[])`, []string{ca.id, cb.id}).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		if count(e.orgA.String()) != 1 || count(e.orgB.String()) != 1 || count("") != 0 {
			t.Fatalf("Organization-bound reads: A %d, B %d, IDENTITY %d", count(e.orgA.String()), count(e.orgB.String()), count(""))
		}
		// After losing A membership the same-binding logout still revokes A's connection; B is unaffected.
		if err := e.model.RemoveGroupMembership(ctx, e.orgA, e.a1.ID, e.users["cross"].uid); err != nil {
			t.Fatal(err)
		}
		if err := e.model.RemoveOrganizationMembership(ctx, e.orgA, e.users["cross"].uid); err != nil {
			t.Fatal(err)
		}
		e.wantToolError(t, ca, ta, ToolGetCurrentContext, nil, CodeContextDenied)
		e.mustCall(t, ca, ta, ToolLogout, nil)
		e.want401(t, ca, ta, "A after logout")
		e.mustCall(t, cb, e.tok(cb), ToolGetCurrentContext, nil)
		e.grantOrg(t, e.orgA, "cross", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "cross", tenancy.GroupMember)
	})

	// IT-MCP-14: RV-1 local containment.
	t.Run("IT-MCP-14 RV-1 local containment", func(t *testing.T) {
		victim := e.newUser(t, "victim")
		other := e.newUser(t, "bystander")
		for _, w := range []string{"victim", "bystander"} {
			e.grantOrg(t, e.orgA, w, tenancy.OrgMember)
		}
		pend := e.create(t, "victim", itHostA)
		a1 := e.create(t, "victim", itHostA)
		t1 := e.activate(t, a1)
		a2 := e.create(t, "victim", itHostA)
		t2 := e.activate(t, a2)
		ob := e.create(t, "bystander", itHostA)
		tb := e.activate(t, ob)
		// (a) Trusted DisableUser ends every pending and active connection of the user; pre-issued tokens get 401.
		if err := e.id.DisableUser(ctx, victim.id); err != nil {
			t.Fatal(err)
		}
		e.want401(t, a1, t1, "active 1 after disable")
		e.want401(t, a2, t2, "active 2 after disable")
		e.want401(t, pend, e.tok(pend), "pending after disable")
		e.mustCall(t, ob, tb, ToolGetCurrentContext, nil)
		// (b) The hold: setup is refused while disabled and nothing re-enables automatically.
		if code, _ := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections", e.cred(t, "victim", itHostA),
			map[string]string{"origin": "https://" + itHostA, "csrf_token": "x"}); code != http.StatusUnauthorized {
			t.Fatalf("setup while held: %d", code)
		}
		var status string
		if err := e.setup.QueryRow(ctx, `SELECT status FROM app.users WHERE id = $1::text::uuid`, victim.id).Scan(&status); err != nil || status != "disabled" {
			t.Fatalf("hold: %s %v", status, err)
		}
		// (c) Release: the old connections stay terminal; a fresh setup with a new UUID activates; old tokens still 401.
		if err := e.id.EnableUser(ctx, victim.id); err != nil {
			t.Fatal(err)
		}
		e.bind(t, "victim", itHostA) // the explicit re-sign-in after the release (G016 A2: at A's origin)
		e.want401(t, a1, t1, "old connection after release")
		fresh := e.create(t, "victim", itHostA)
		tf := e.activate(t, fresh)
		e.mustCall(t, fresh, tf, ToolGetCurrentContext, nil)
		e.want401(t, fresh, t1, "old token on the replacement")
		// (d) A user already disabled before the procedure stays disabled with disabled_at unchanged; no user.enabled.
		pre := e.newUser(t, "predisabled")
		e.grantOrg(t, e.orgA, "predisabled", tenancy.OrgMember)
		pc := e.create(t, "predisabled", itHostA)
		ptok := e.activate(t, pc)
		if err := e.id.DisableUser(ctx, pre.id); err != nil {
			t.Fatal(err)
		}
		var before time.Time
		if err := e.setup.QueryRow(ctx, `SELECT disabled_at FROM app.users WHERE id = $1::text::uuid`, pre.id).Scan(&before); err != nil {
			t.Fatal(err)
		}
		// The procedure: prior state recorded as disabled, so no own hold and no release (nothing is called).
		var after time.Time
		var enabled int
		if err := e.setup.QueryRow(ctx, `SELECT disabled_at, (SELECT count(*) FROM audit.events WHERE action = 'user.enabled'
  AND subject_user_id = $1::text::uuid) FROM app.users WHERE id = $1::text::uuid`, pre.id).Scan(&after, &enabled); err != nil {
			t.Fatal(err)
		}
		if !after.Equal(before) || enabled != 0 {
			t.Fatalf("pre-disabled user changed: %v %v %d", before, after, enabled)
		}
		e.want401(t, pc, ptok, "pre-disabled user's connection")
		_ = other
	})

	// IT-MCP-15 (G063 Plan r1 §6): the setup rename over the real HTTP path and stack.
	t.Run("IT-MCP-15 rename", func(t *testing.T) {
		serveRename(t, e, e.h)
		e.bind(t, "cross", itHostA, itHostB) // G016 A2: explicit renewal; earlier subtests moved the clock
		e.newUser(t, "alice15")
		e.newUser(t, "bob15")
		e.grantOrg(t, e.orgA, "alice15", tenancy.OrgMember)
		e.grantOrg(t, e.orgA, "bob15", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "alice15", tenancy.GroupMember)
		emoji80 := strings.Repeat("\U0001F600", 80)
		aCred := func() string { return e.cred(t, "alice15", itHostA) }
		rename := func(cred, host, id string, name any) (int, map[string]any) {
			t.Helper()
			return e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/rename", cred, map[string]any{"origin": "https://" + host,
				"connection_id": id, "csrf_token": e.csrf(t, cred), "display_name": name})
		}
		listed := func(cred, host string) map[string]map[string]any {
			t.Helper()
			code, out := e.internal(t, http.MethodGet, "/internal/v1/mcp/connections?"+url.Values{"origin": {"https://" + host}}.Encode(), cred, nil)
			if code != http.StatusOK {
				t.Fatalf("listing: %d %v", code, out)
			}
			m := map[string]map[string]any{}
			for _, x := range out["connections"].([]any) {
				r := x.(map[string]any)
				m[r["connection_id"].(string)] = r
			}
			return m
		}
		nameOf := func(id string) *string {
			t.Helper()
			var n *string
			if err := e.setup.QueryRow(ctx, `SELECT display_name FROM app.mcp_connections WHERE id = $1::text::uuid`, id).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		// facts is every connection and session column except the name (id, resource, host, user, Organization, client,
		// selected Group, context version, activation and the session's expiries, renewal and revocation).
		facts := func(id string) string {
			t.Helper()
			var s string
			if err := e.setup.QueryRow(ctx, `SELECT ((to_jsonb(c) - 'display_name') || jsonb_build_object('session', to_jsonb(s)))::text
  FROM app.mcp_connections c JOIN app.sessions s ON s.id = c.session_id WHERE c.id = $1::text::uuid`, id).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
		renamedEvents := func(id string) int {
			t.Helper()
			n := 0
			for _, ev := range e.events(t, id) {
				if strings.HasPrefix(ev, "mcp.connection_renamed:") {
					n++
				}
			}
			return n
		}
		wantRenamed := func(what string, code int, out map[string]any) {
			t.Helper()
			if code != http.StatusOK || len(out) != 1 || out["outcome"] != "renamed" {
				t.Errorf("%s: %d %v, want 200 {outcome: renamed}", what, code, out)
			}
		}
		notFound := map[string]any{"error": "not_found"}

		// Rename and clear a pending and an active connection; nothing but the name changes.
		pending := e.create(t, "alice15", itHostA)
		active := e.create(t, "alice15", itHostA)
		e.activate(t, active)
		e.mustCall(t, active, e.tok(active), ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		for label, c := range map[string]conn{"pending": pending, "active": active} {
			before := facts(c.id)
			code, out := rename(aCred(), itHostA, c.id, "Work laptop")
			wantRenamed(label+" rename", code, out)
			if v := listed(aCred(), itHostA)[c.id]["display_name"]; v != "Work laptop" {
				t.Errorf("%s: listed name %v", label, v)
			}
			code, out = rename(aCred(), itHostA, c.id, "   ")
			wantRenamed(label+" clear", code, out)
			if v, has := listed(aCred(), itHostA)[c.id]["display_name"]; !has || v != nil {
				t.Errorf("%s: listed name after clear %v (key %v)", label, v, has)
			}
			if after := facts(c.id); after != before {
				t.Errorf("%s: the rename changed a fact:\n before %s\n after  %s", label, before, after)
			}
		}
		// The padded 80-non-BMP vector is stored as the 80 code points and listed back byte-identically; duplicates are
		// allowed; markup is kept verbatim (the UI escapes it).
		code, out := rename(aCred(), itHostA, pending.id, "\u3000\u3000"+emoji80+"\u00A0 ")
		wantRenamed("padded non-BMP", code, out)
		if n := nameOf(pending.id); n == nil || *n != emoji80 {
			t.Errorf("padded non-BMP stored as %v", n)
		}
		if v := listed(aCred(), itHostA)[pending.id]["display_name"]; v != emoji80 {
			t.Errorf("padded non-BMP listed as %v", v)
		}
		for _, id := range []string{pending.id, active.id} {
			code, out := rename(aCred(), itHostA, id, "<b>Same</b> & co")
			wantRenamed("duplicate name", code, out)
		}
		// A same-value rename succeeds and records no second event; the event names only IDs and closed metadata.
		before := renamedEvents(active.id)
		code, out = rename(aCred(), itHostA, active.id, "<b>Same</b> & co")
		wantRenamed("same value", code, out)
		if n := renamedEvents(active.id); n != before {
			t.Errorf("a same-value rename recorded %d more events", n-before)
		}
		var shape string
		if err := e.setup.QueryRow(ctx, `SELECT concat_ws('|', outcome, coalesce(reason, '-'), coalesce(operation, '-'), coalesce(group_id::text, '-'),
  resource_type, (organization_id = $2::text::uuid)::text, (subject_user_id = $3::text::uuid)::text, actor_kind,
  (actor_user_id = $3::text::uuid)::text, (position('Same' in e::text) > 0 OR position('laptop' in e::text) > 0)::text)
  FROM audit.events e WHERE action = 'mcp.connection_renamed' AND resource_id = $1::text::uuid ORDER BY occurred_at DESC LIMIT 1`,
			active.id, e.orgA.String(), e.users["alice15"].id).Scan(&shape); err != nil {
			t.Fatal(err)
		}
		if shape != "success|-|-|-|mcp_connection|true|true|user|true|false" {
			t.Errorf("mcp.connection_renamed row shape: %s", shape)
		}

		// Non-disclosing negatives: another user's row, an unknown UUID and a member of both Organizations renaming through
		// the other Organization's host all answer the same not_found, and change nothing.
		e.bind(t, "cross", itHostA, itHostB)
		crossRow := e.create(t, "cross", itHostA)
		const unknownID = "7f000000-0000-4000-8000-00000000f015"
		if code, out := rename(aCred(), itHostA, unknownID, "Intruder"); code != http.StatusNotFound || !maps.Equal(out, notFound) {
			t.Errorf("unknown connection: %d %v, want 404 not_found", code, out)
		}
		for name, c := range map[string]struct{ cred, host, id string }{
			"another user's row":        {e.cred(t, "bob15", itHostA), itHostA, active.id},
			"through the other tenant":  {e.cred(t, "cross", itHostB), itHostB, crossRow.id},
			"another tenant's user row": {e.cred(t, "cross", itHostB), itHostB, active.id},
		} {
			beforeFacts, beforeName := facts(c.id), nameOf(c.id)
			code, out := rename(c.cred, c.host, c.id, "Intruder")
			if code != http.StatusNotFound || !maps.Equal(out, notFound) {
				t.Errorf("%s: %d %v, want 404 not_found (byte-identical to unknown)", name, code, out)
			}
			if facts(c.id) != beforeFacts || !slices.Equal(strs(nameOf(c.id)), strs(beforeName)) {
				t.Errorf("%s: the refused rename changed the row", name)
			}
		}
		// D-4: a caller who lost the Organization membership is not_available for the rename (its revoke stays available).
		if err := e.model.RemoveGroupMembership(ctx, e.orgA, e.a1.ID, e.users["alice15"].uid); err != nil {
			t.Fatal(err)
		}
		if err := e.model.RemoveOrganizationMembership(ctx, e.orgA, e.users["alice15"].uid); err != nil {
			t.Fatal(err)
		}
		if code, out := rename(aCred(), itHostA, active.id, "Lost"); code != http.StatusForbidden || out["error"] != "not_available" {
			t.Errorf("non-member rename: %d %v, want 403 not_available", code, out)
		}
		e.grantOrg(t, e.orgA, "alice15", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "alice15", tenancy.GroupMember)
		if n := nameOf(active.id); n == nil || *n != "<b>Same</b> & co" {
			t.Errorf("the non-member rename changed the name: %v", n)
		}
		// Session, CSRF, origin, method and header refusals.
		for name, c := range map[string]struct {
			cred string
			body map[string]any
			code int
			err  string
		}{
			"no csrf": {aCred(), map[string]any{"origin": "https://" + itHostA, "connection_id": active.id, "csrf_token": "", "display_name": "x"},
				http.StatusForbidden, "csrf_failed"},
			"bad origin": {aCred(), map[string]any{"origin": "https://org-c.localhost:543", "connection_id": active.id, "csrf_token": e.csrf(t, aCred()),
				"display_name": "x"}, http.StatusForbidden, "origin_rejected"},
			"no session": {"", map[string]any{"origin": "https://" + itHostA, "connection_id": active.id, "csrf_token": "x", "display_name": "x"},
				http.StatusUnauthorized, "unauthenticated"},
			"A-bound credential at B": {e.cred(t, "cross", itHostA), map[string]any{"origin": "https://" + itHostB, "connection_id": crossRow.id,
				"csrf_token": e.csrf(t, e.cred(t, "cross", itHostA)), "display_name": "x"}, http.StatusUnauthorized, "unauthenticated"},
			"no name":        {aCred(), map[string]any{"origin": "https://" + itHostA, "connection_id": active.id, "csrf_token": e.csrf(t, aCred())}, http.StatusBadRequest, "invalid_request"},
			"null name":      {aCred(), map[string]any{"origin": "https://" + itHostA, "connection_id": active.id, "csrf_token": e.csrf(t, aCred()), "display_name": nil}, http.StatusBadRequest, "invalid_request"},
			"control":        {aCred(), map[string]any{"origin": "https://" + itHostA, "connection_id": active.id, "csrf_token": e.csrf(t, aCred()), "display_name": "a\tb"}, http.StatusBadRequest, "invalid_name"},
			"81 code points": {aCred(), map[string]any{"origin": "https://" + itHostA, "connection_id": active.id, "csrf_token": e.csrf(t, aCred()), "display_name": strings.Repeat("a", 81)}, http.StatusBadRequest, "invalid_name"},
			"malformed uuid": {aCred(), map[string]any{"origin": "https://" + itHostA, "connection_id": "not-a-uuid", "csrf_token": e.csrf(t, aCred()), "display_name": "x"}, http.StatusBadRequest, "invalid_request"},
			"unknown field":  {aCred(), map[string]any{"origin": "https://" + itHostA, "connection_id": active.id, "csrf_token": e.csrf(t, aCred()), "display_name": "x", "client_id": "c"}, http.StatusBadRequest, "invalid_request"},
		} {
			code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/rename", c.cred, c.body)
			if code != c.code || out["error"] != c.err {
				t.Errorf("%s: %d %v, want %d %s", name, code, out, c.code, c.err)
			}
		}
		if n := nameOf(active.id); n == nil || *n != "<b>Same</b> & co" {
			t.Errorf("a refused request changed the name: %v", n)
		}
		if code, _ := e.internal(t, http.MethodGet, "/internal/v1/mcp/connections/rename", aCred(), nil); code != http.StatusMethodNotAllowed {
			t.Errorf("GET rename: %d, want 405", code)
		}
		// The wire negatives of TestSetupNameDecoding over the real server, with raw bytes: refused, nothing changed.
		raw := func(nameField string) string {
			return `{"origin":"https://` + itHostA + `","connection_id":"` + active.id + `","csrf_token":"` + e.csrf(t, aCred()) + `"` + nameField + `}`
		}
		for name, c := range map[string]struct {
			body string
			code int
			want string
		}{
			"invalid UTF-8":       {raw(",\"display_name\":\"\xff\""), http.StatusBadRequest, `{"error":"invalid_name"}`},
			"lone surrogate":      {raw(`,"display_name":"\ud800"`), http.StatusBadRequest, `{"error":"invalid_name"}`},
			"reversed surrogates": {raw(`,"display_name":"\udc00\ud800"`), http.StatusBadRequest, `{"error":"invalid_name"}`},
			"duplicate key":       {raw(`,"display_name":"a","display_name":"b"`), http.StatusBadRequest, `{"error":"invalid_request"}`},
			"non-string":          {raw(`,"display_name":["a"]`), http.StatusBadRequest, `{"error":"invalid_request"}`},
		} {
			code, body := rawSetup(t, e, "/internal/v1/mcp/connections/rename", c.body, aCred(), false)
			if code != c.code || body != c.want {
				t.Errorf("raw %s: %d %s, want %d %s", name, code, body, c.code, c.want)
			}
		}
		if code, body := rawSetup(t, e, "/internal/v1/mcp/connections/rename", raw(`,"display_name":"x"`), aCred(), true); code != http.StatusBadRequest ||
			body != `{"error":"ambiguous_parameter"}` {
			t.Errorf("duplicate Authorization: %d %s", code, body)
		}
		if code, body := rawSetup(t, e, "/internal/v1/mcp/connections/rename", raw(`,"display_name":"\ud83d\ude00"`), aCred(), false); code != http.StatusOK ||
			body != `{"outcome":"renamed"}` {
			t.Errorf("escaped surrogate pair: %d %s", code, body)
		}
		if n := nameOf(active.id); n == nil || *n != "\U0001F600" {
			t.Errorf("the escaped pair stored %v", n)
		}

		// Target-layer terminal classes with a live caller session: not_found and no change.
		terminal := func(label string, c conn, prep func()) {
			t.Helper()
			prep()
			beforeName := nameOf(c.id)
			code, out := rename(aCred(), itHostA, c.id, "After")
			if code != http.StatusNotFound || !maps.Equal(out, notFound) {
				t.Errorf("%s: %d %v, want 404 not_found", label, code, out)
			}
			if !slices.Equal(strs(nameOf(c.id)), strs(beforeName)) {
				t.Errorf("%s: the name changed", label)
			}
		}
		revoked := e.create(t, "alice15", itHostA)
		terminal("revoked through setup", revoked, func() {
			if code, _ := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/revoke", aCred(),
				map[string]string{"origin": "https://" + itHostA, "connection_id": revoked.id, "csrf_token": e.csrf(t, aCred())}); code != http.StatusOK {
				t.Fatalf("revoke: %d", code)
			}
		})
		loggedOut := e.create(t, "alice15", itHostA)
		lt := e.activate(t, loggedOut)
		terminal("logout tool", loggedOut, func() { e.mustCall(t, loggedOut, lt, ToolLogout, nil) })
		expiredSetup := e.create(t, "alice15", itHostA)
		terminal("expired setup", expiredSetup, func() {
			e.clk.Advance(31 * time.Minute)
			e.bind(t, "alice15", itHostA) // G016 A2: explicit renewal after the advance (the caller's session is live)
		})
		idle := e.create(t, "alice15", itHostA)
		e.activate(t, idle)
		terminal("idle-expired session", idle, func() {
			e.clk.Advance(31 * 24 * time.Hour)
			e.bind(t, "alice15", itHostA)
		})
		// A disabled owner is the browser-auth layer: the caller's session ends with the owner (unauthenticated).
		dCred := e.cred(t, "alice15", itHostA)
		dCSRF := e.csrf(t, dCred)
		dc := e.create(t, "alice15", itHostA)
		if err := e.id.DisableUser(ctx, e.users["alice15"].id); err != nil {
			t.Fatal(err)
		}
		if code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/rename", dCred, map[string]any{"origin": "https://" + itHostA,
			"connection_id": dc.id, "csrf_token": dCSRF, "display_name": "Disabled"}); code != http.StatusUnauthorized || out["error"] != "unauthenticated" {
			t.Errorf("disabled owner: %d %v, want 401 unauthenticated", code, out)
		}
		if n := nameOf(dc.id); n != nil {
			t.Errorf("disabled owner: the name changed to %q", *n)
		}
		if err := e.id.EnableUser(ctx, e.users["alice15"].id); err != nil {
			t.Fatal(err)
		}
		e.bind(t, "alice15", itHostA)

		// Alias regression: a row created on the old host is renamed through a current alias host of the same
		// Organization; its stored host and resource stay.
		aliasHost := fmt.Sprintf("org-a-rename-%d.localhost:543", itSeq.Add(1))
		serveRename(t, e, e.handler(t, map[string]tenancy.OrganizationID{itHostA: e.orgA, itHostB: e.orgB, aliasHost: e.orgA}, nil))
		e.route(t, aliasHost, e.orgA)
		e.bind(t, "alice15", aliasHost)
		old := e.create(t, "alice15", itHostA)
		oldFacts := facts(old.id)
		code, out = rename(e.cred(t, "alice15", aliasHost), aliasHost, old.id, "Through the alias")
		wantRenamed("alias rename", code, out)
		if v := listed(e.cred(t, "alice15", aliasHost), aliasHost)[old.id]; v == nil || v["display_name"] != "Through the alias" ||
			!strings.HasPrefix(v["url"].(string), "https://"+itHostA+"/") {
			t.Errorf("alias listing of the renamed old-host row: %v", v)
		}
		if facts(old.id) != oldFacts {
			t.Error("the alias rename changed the old-host row's facts")
		}

		// The diagnostic records: mcp.setup rename events with the closed outcome/reason; no name in any record.
		logs := e.logs.String()
		if !strings.Contains(logs, `"operation":"rename"`) {
			t.Error("no mcp.setup rename record")
		}
		for _, n := range []string{"Work laptop", "Same</b>", "Through the alias", "Intruder", emoji80[:8]} {
			if strings.Contains(logs, n) {
				t.Errorf("a diagnostic record carries the name %q", n)
			}
		}
	})
}

// TestMCPAudit is IT-MCP-10's audit and privacy cases: the exact events, a strict audit fault on activation and on
// selection, and the privacy scan of the core records.
func TestMCPAudit(t *testing.T) {
	e := setupMCP(t)
	ctx := context.Background()
	c := e.create(t, "ownerA", itHostA)
	// An audit fault refuses activation (strict): the connection stays pending.
	fault := func(on bool) {
		stmt := `ALTER TABLE audit.events DROP CONSTRAINT IF EXISTS it_mcp_fault`
		if on {
			stmt = `ALTER TABLE audit.events ADD CONSTRAINT it_mcp_fault CHECK (action NOT LIKE 'mcp.%') NOT VALID`
		}
		if _, err := e.setup.Exec(ctx, stmt); err != nil {
			t.Fatal(err)
		}
	}
	fault(true)
	if code, _ := e.rpc(t, c, e.tok(c), "ping", map[string]any{}); code != http.StatusServiceUnavailable {
		t.Fatalf("activation with an audit fault: %d", code)
	}
	if st, _, _, _ := e.state(t, c.id); st != identity.MCPStatePending {
		t.Fatalf("activation committed without its event: %s", st)
	}
	fault(false)
	tok := e.activate(t, c)
	fault(true)
	e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()}, CodeUnavailable)
	if _, _, g, _ := e.state(t, c.id); g != "" {
		t.Fatalf("selection committed without its event: %s", g)
	}
	fault(false)
	e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
	e.mustCall(t, c, tok, ToolLogout, nil)
	got := e.events(t, c.id)
	if !slices.Equal(got, []string{"mcp.connection_established:-:user", "mcp.context_changed:" + e.a1.ID.String() + ":user"}) {
		t.Fatalf("events: %v", got)
	}
	// The logout is the existing session.revoked of the connection's session (reason logout).
	var n int
	if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM audit.events ev JOIN app.mcp_connections c ON c.session_id = ev.resource_id
  WHERE c.id = $1::text::uuid AND ev.action = 'session.revoked' AND ev.reason = 'logout'`, c.id).Scan(&n); err != nil || n != 1 {
		t.Fatalf("session.revoked: %d %v", n, err)
	}
	// Privacy: no token value, subject or client identifier reaches the core records.
	logs := e.logs.String()
	for _, s := range append(slices.Clone(e.secrets), e.users["ownerA"].sub) {
		if strings.Contains(logs, s) {
			t.Fatal("a token or provider value reached the core records")
		}
	}
}

// TestMCPDiscovery is G015 CAT-1..5 (Plan r4 §6.2; Q8: the five discovery scenarios in package mcp, distinct from the
// migrate package's TestMCPCatalog): requirement-aware discovery through the real HTTP path with the test-only
// capability-mapped tools of catalog_fixture_test.go (no effect, no check; invocation counters). It runs before
// TestMCPRecovery, which must stay last in its database.
func TestMCPDiscovery(t *testing.T) {
	e := setupMCP(t)
	ctx := context.Background()
	top := t
	counters := useFixtures(t, e.h)
	fixturesOn := func() { counters = useFixtures(top, e.h) }
	sorted := func(extra ...string) []string {
		all := append(slices.Clone(ToolNames), extra...)
		slices.Sort(all)
		return all
	}
	same := func(t *testing.T, what string, got, want []string) {
		t.Helper()
		if !slices.Equal(got, want) {
			t.Fatalf("%s: catalog %v, want %v", what, got, want)
		}
	}
	version := func(t *testing.T, c conn) (string, int64) {
		t.Helper()
		_, _, g, v := e.state(t, c.id)
		return g, v
	}
	// The fixture world (Plan r4 §6.1): profile P linked to A1 and shared with A2, an unlinked profile of A, an
	// Organization admin of A without any Group binding and a Group admin of A1.
	p := e.profile(t, e.orgA, "Shared Profile")
	e.profile(t, e.orgA, "Unlinked Profile")
	e.link(t, e.orgA, e.a1.ID, p.ID)
	e.link(t, e.orgA, e.a2.ID, p.ID)
	e.setFlag(t, e.orgA, true)
	e.newUser(t, "adminA")
	e.grantOrg(t, e.orgA, "adminA", tenancy.OrgAdmin)
	gadmin := e.newUser(t, "gadminA1")
	e.grantOrg(t, e.orgA, "gadminA1", tenancy.OrgMember)
	e.grantGroup(t, e.orgA, e.a1.ID, "gadminA1", tenancy.GroupAdmin)

	t.Run("CAT-1 the catalog follows the matrix per role, context, link and policy", func(t *testing.T) {
		owner := e.create(t, "ownerA", itHostA)
		otok := e.activate(t, owner)
		// The production catalog is unaffected by the fixtures: without the seam, exactly the five tools.
		clearFixtures(e.h)
		same(t, "production", e.list(t, owner, otok), sorted())
		fixturesOn()
		same(t, "owner", e.list(t, owner, otok), sorted(fixtureNames...))
		// The official SDK client sees the same sorted list without a cursor.
		cs, err := e.dial(t, owner, otok)
		if err != nil {
			t.Fatal(err)
		}
		res, err := cs.ListTools(ctx, nil)
		_ = cs.Close()
		if err != nil {
			t.Fatal(err)
		}
		if res.NextCursor != "" {
			t.Fatalf("SDK list cursor %q", res.NextCursor)
		}
		var names []string
		for _, tool := range res.Tools {
			names = append(names, tool.Name)
		}
		same(t, "owner via the SDK", names, sorted(fixtureNames...))

		member := e.create(t, "memberA1", itHostA)
		same(t, "group member", e.list(t, member, e.activate(t, member)), sorted())

		// The Organization admin without Group bindings. G060 (r0 A7): with the Group-management override off its
		// Organization path confers no Group or profile management, so neither entry is listed (Organization-only or A1
		// selected). With the override on (owner-only setting; explicit fixture setup through the setup connection, the
		// setFlag pattern, reset when CAT-1 ends): the all and Y paths in Organization-only context and, with A1
		// selected, the Group entry (target restricted to A1) and the profile entry (unlinked profiles stay reachable).
		setOverride := func(on bool) {
			t.Helper()
			if _, err := e.setup.Exec(ctx, `UPDATE app.organizations SET group_management_override = $2 WHERE id = $1::text::uuid`,
				e.orgA.String(), on); err != nil {
				t.Fatal(err)
			}
		}
		admin := e.create(t, "adminA", itHostA)
		atok := e.activate(t, admin)
		adminWant := sorted(fixtureGroupUpdate, fixtureProfileUpdate)
		same(t, "org admin, Organization-only, override off", e.list(t, admin, atok), sorted())
		e.mustCall(t, admin, atok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		same(t, "org admin, A1 selected, override off", e.list(t, admin, atok), sorted())
		setOverride(true)
		t.Cleanup(func() { setOverride(false) })
		same(t, "org admin, A1 selected", e.list(t, admin, atok), adminWant)
		admin2 := e.create(t, "adminA", itHostA)
		same(t, "org admin, Organization-only", e.list(t, admin2, e.activate(t, admin2)), adminWant)

		// The Group admin of A1: own Group and linked+flag profile paths only.
		ga := e.create(t, "gadminA1", itHostA)
		gtok := e.activate(t, ga)
		same(t, "group admin", e.list(t, ga, gtok), adminWant)
		gsel := e.create(t, "gadminA1", itHostA)
		gstok := e.activate(t, gsel)
		e.mustCall(t, gsel, gstok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		same(t, "group admin, A1 selected", e.list(t, gsel, gstok), adminWant)
		// Link removal: the Group admin loses the profile entry; the Organization admin (override on) keeps it, also with A1
		// selected.
		e.unlink(t, e.orgA, e.a1.ID, p.ID)
		same(t, "group admin after unlink", e.list(t, ga, gtok), sorted(fixtureGroupUpdate))
		same(t, "group admin after unlink, A1 selected", e.list(t, gsel, gstok), sorted(fixtureGroupUpdate))
		same(t, "org admin after unlink, A1 selected", e.list(t, admin, atok), adminWant)
		e.link(t, e.orgA, e.a1.ID, p.ID)
		same(t, "group admin after relink", e.list(t, ga, gtok), adminWant)
		// Policy off: the linked+flag path closes; the Organization path (override on) does not depend on the flag.
		e.setFlag(t, e.orgA, false)
		same(t, "group admin, policy off", e.list(t, ga, gtok), sorted(fixtureGroupUpdate))
		same(t, "org admin, policy off", e.list(t, admin, atok), adminWant)
		e.setFlag(t, e.orgA, true)
		same(t, "group admin, policy on", e.list(t, ga, gtok), adminWant)
		// A downgrade to group member removes both entries on the next list.
		if _, err := e.model.ChangeGroupRole(ctx, e.orgA, e.a1.ID, gadmin.uid, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		same(t, "group admin downgraded", e.list(t, ga, gtok), sorted())
		if _, err := e.model.ChangeGroupRole(ctx, e.orgA, e.a1.ID, gadmin.uid, tenancy.GroupAdmin); err != nil {
			t.Fatal(err)
		}
		for _, n := range fixtureNames {
			if counters.get(n) != 0 {
				t.Fatalf("listing invoked %s", n)
			}
		}
	})

	t.Run("CAT-2 selection and external changes govern the next list without a version change", func(t *testing.T) {
		u := e.newUser(t, "switcher")
		e.grantOrg(t, e.orgA, "switcher", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "switcher", tenancy.GroupAdmin)
		e.grantGroup(t, e.orgA, e.a2.ID, "switcher", tenancy.GroupMember)
		c := e.create(t, "switcher", itHostA)
		tok := e.activate(t, c)
		full := sorted(fixtureGroupUpdate, fixtureProfileUpdate)
		same(t, "Organization-only", e.list(t, c, tok), full)
		// Local changes: a select and a clear commit before their responses; the next list is the new context's.
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a2.ID.String()})
		same(t, "A2 selected", e.list(t, c, tok), sorted())
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": nil})
		same(t, "cleared", e.list(t, c, tok), full)
		if g, v := version(t, c); g != "" || v != 2 {
			t.Fatalf("after select and clear: %q %d", g, v)
		}
		// Definitive refusals change nothing: a foreign target, invalid input.
		e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.b1.ID.String()}, CodeGroupNotSelectable)
		e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": "not-a-uuid"}, CodeInvalidInput)
		if g, v := version(t, c); g != "" || v != 2 {
			t.Fatalf("after refused selects: %q %d", g, v)
		}
		same(t, "after refused selects", e.list(t, c, tok), full)
		// context_denied: after the selected Group is lost, a select elsewhere is refused and changes nothing; the list
		// stays refused as a whole; the clear restores it.
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a2.ID.String()})
		if err := e.model.RemoveGroupMembership(ctx, e.orgA, e.a2.ID, u.uid); err != nil {
			t.Fatal(err)
		}
		if code, msg := e.listError(t, c, tok); code != http.StatusOK || msg != "-32040 context_denied" {
			t.Fatalf("list after the selected Group was lost: %d %q", code, msg)
		}
		e.wantToolError(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()}, CodeContextDenied)
		if g, v := version(t, c); g != e.a2.ID.String() || v != 3 {
			t.Fatalf("after the refused select: %q %d", g, v)
		}
		if code, msg := e.listError(t, c, tok); msg != "-32040 context_denied" {
			t.Fatalf("list after the refused select: %d %q", code, msg)
		}
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": nil})
		same(t, "cleared after the loss", e.list(t, c, tok), full)
		// External changes: each governs the next list with the same credential and no version change.
		external := []struct {
			name   string
			change func()
			want   []string
		}{
			{"downgrade", func() { _, _ = e.model.ChangeGroupRole(ctx, e.orgA, e.a1.ID, u.uid, tenancy.GroupMember) }, sorted()},
			{"grant", func() { _, _ = e.model.ChangeGroupRole(ctx, e.orgA, e.a1.ID, u.uid, tenancy.GroupAdmin) }, full},
			{"unlink", func() { e.unlink(t, e.orgA, e.a1.ID, p.ID) }, sorted(fixtureGroupUpdate)},
			{"relink", func() { e.link(t, e.orgA, e.a1.ID, p.ID) }, full},
			{"policy off", func() { e.setFlag(t, e.orgA, false) }, sorted(fixtureGroupUpdate)},
			{"policy on", func() { e.setFlag(t, e.orgA, true) }, full},
		}
		for _, x := range external {
			x.change()
			same(t, x.name, e.list(t, c, tok), x.want)
			if g, v := version(t, c); g != "" || v != 4 {
				t.Fatalf("%s changed the selection revision: %q %d", x.name, g, v)
			}
		}
		// A store outage is 503 with no catalog (never a partial or cached one).
		down := e.handler(t, map[string]tenancy.OrganizationID{itHostA: e.orgA}, nil)
		closedPool := itPool(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		closedPool.Close()
		down.id = identity.NewService(identity.NewStore(closedPool), identity.Config{Issuer: e.fake.Issuer, ClientID: itClientG004, Now: e.clk.Now})
		req, _ := http.NewRequest(http.MethodPost, ConnectionPathPrefix+c.id, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
		req.Host = itHostA
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Authorization", "Bearer "+e.tok(c))
		w := &recorder{h: http.Header{}}
		down.Wrap(http.NotFoundHandler()).ServeHTTP(w, req)
		if w.code != http.StatusServiceUnavailable || strings.Contains(w.body.String(), "tools") {
			t.Fatalf("list with the store down: %d %s", w.code, w.body.String())
		}
		// Terminal states: revoke, disable and idle or absolute expiry refuse the next list; a later re-grant revives
		// nothing.
		swCred := e.cred(t, "switcher", itHostA)
		if code, _ := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/revoke", swCred,
			map[string]string{"origin": "https://" + itHostA, "connection_id": c.id, "csrf_token": e.csrf(t, swCred)}); code != http.StatusOK {
			t.Fatalf("revoke: %d", code)
		}
		if code, _ := e.rpc(t, c, e.tok(c), "tools/list", map[string]any{}); code != http.StatusUnauthorized {
			t.Fatalf("list after revoke: %d", code)
		}
		e.grantGroup(t, e.orgA, e.a2.ID, "switcher", tenancy.GroupAdmin)
		if code, _ := e.rpc(t, c, e.tok(c), "tools/list", map[string]any{}); code != http.StatusUnauthorized {
			t.Fatalf("list after revoke and re-grant: %d", code)
		}
		d := e.create(t, "switcher", itHostA)
		e.activate(t, d)
		if err := e.id.DisableUser(ctx, u.id); err != nil {
			t.Fatal(err)
		}
		if err := e.id.EnableUser(ctx, u.id); err != nil {
			t.Fatal(err)
		}
		if code, _ := e.rpc(t, d, e.tok(d), "tools/list", map[string]any{}); code != http.StatusUnauthorized {
			t.Fatalf("list after disable and enable: %d", code)
		}
		e.bind(t, "switcher", itHostA) // the explicit re-sign-in after the disable (G016 A2: at A's origin)
		idle := e.create(t, "switcher", itHostA)
		e.activate(t, idle)
		e.clk.Advance(31 * 24 * time.Hour)
		if code, _ := e.rpc(t, idle, e.tok(idle), "tools/list", map[string]any{}); code != http.StatusUnauthorized {
			t.Fatalf("list after idle expiry: %d", code)
		}
		// A connection renewed on every list still ends at the absolute limit.
		e.bind(t, "switcher", itHostA) // the explicit renewal after the 31-day advance (G016 A2: at A's origin)
		abs := e.create(t, "switcher", itHostA)
		e.activate(t, abs)
		for i := 0; i < 7; i++ {
			e.clk.Advance(29 * 24 * time.Hour)
			if i < 6 {
				e.list(t, abs, e.tok(abs))
			}
		}
		if code, _ := e.rpc(t, abs, e.tok(abs), "tools/list", map[string]any{}); code != http.StatusUnauthorized {
			t.Fatalf("list after absolute expiry: %d", code)
		}
		// G016 A2: this subtest moved the clock past every browser session's lifetime; explicit renewal of the credential
		// the later subtests present (memberA1 at A for CAT-3).
		e.bind(t, "memberA1", itHostA)
	})

	t.Run("CAT-3 hidden equals unknown; stale calls fail closed", func(t *testing.T) {
		const hidden = fixtureGroupUpdate
		m := e.create(t, "memberA1", itHostA)
		mtok := e.activate(t, m)
		member := e.users["memberA1"]
		// The same requested name with the seam empty (absent) and set but ineligible (hidden): byte-identical answers,
		// no handler run, no audit row and one allow decision record each.
		before, recs := e.mcpDenials(t, member), len(e.authorizeRecords(t))
		clearFixtures(e.h)
		code1, unknown := e.rawCall(t, m, mtok, hidden)
		fixturesOn()
		code2, hid := e.rawCall(t, m, mtok, hidden)
		if code1 != http.StatusOK || code2 != http.StatusOK || unknown != hid ||
			!strings.Contains(unknown, `"code":-32602`) || !strings.Contains(unknown, `unknown tool \"`+hidden+`\"`) {
			t.Fatalf("hidden versus unknown:\n%d %s\n%d %s", code1, unknown, code2, hid)
		}
		if got := e.authorizeRecords(t)[recs:]; !slices.Equal(got, []string{"info|allow|allowed", "info|allow|allowed"}) {
			t.Fatalf("decision records: %v", got)
		}
		if e.mcpDenials(t, member) != before || counters.get(hidden) != 0 {
			t.Fatal("a hidden call recorded a denial or ran the tool")
		}
		// Another never-existing name shares the classification and template, differing only in the echoed name.
		if _, other := e.rawCall(t, m, mtok, "test.never_existed"); other != strings.Replace(unknown, hidden, "test.never_existed", 1) {
			t.Fatalf("never-existing name: %s", other)
		}
		// A tool listed earlier and hidden after a downgrade: the paired authorized call runs once, then unknown.
		h := e.newUser(t, "hider")
		e.grantOrg(t, e.orgA, "hider", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "hider", tenancy.GroupAdmin)
		hc := e.create(t, "hider", itHostA)
		htok := e.activate(t, hc)
		if !slices.Contains(e.list(t, hc, htok), hidden) {
			t.Fatal("not listed before the downgrade")
		}
		if out := e.mustCall(t, hc, htok, hidden, nil); out["invoked"] != true || counters.get(hidden) != 1 {
			t.Fatalf("authorized fixture call: %v %d", out, counters.get(hidden))
		}
		if _, err := e.model.ChangeGroupRole(ctx, e.orgA, e.a1.ID, h.uid, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		if _, stale := e.rawCall(t, hc, htok, hidden); stale != unknown || counters.get(hidden) != 1 {
			t.Fatalf("stale call after the downgrade: %s", stale)
		}
		// C-6: after Organization membership loss and, separately, after selected-Group loss, the hidden and the unknown
		// name answer byte-identically, each with exactly one access.denied mcp.request row and one deny record with
		// identical fields; the list is refused as a whole; logout (and the clear) still succeed.
		refused := func(t *testing.T, c conn, tok string, u *itUser) {
			t.Helper()
			var bodies, records []string
			for _, seam := range []bool{false, true} {
				if seam {
					fixturesOn()
				} else {
					clearFixtures(e.h)
				}
				d0, r0 := e.mcpDenials(t, u), len(e.authorizeRecords(t))
				code, body := e.rawCall(t, c, tok, hidden)
				if code != http.StatusOK || e.mcpDenials(t, u) != d0+1 {
					t.Fatalf("seam %t: %d, denials %d -> %d", seam, code, d0, e.mcpDenials(t, u))
				}
				got := e.authorizeRecords(t)[r0:]
				if len(got) != 1 || !strings.HasPrefix(got[0], "warn|deny|") {
					t.Fatalf("seam %t: decision records %v", seam, got)
				}
				bodies, records = append(bodies, body), append(records, got[0])
			}
			if bodies[0] != bodies[1] || bodies[0] != unknown || records[0] != records[1] {
				t.Fatalf("refusal differs: %v %v", bodies, records)
			}
			if _, msg := e.listError(t, c, tok); msg != "-32040 context_denied" {
				t.Fatalf("list: %q", msg)
			}
		}
		lost := e.newUser(t, "lostmember")
		e.grantOrg(t, e.orgA, "lostmember", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "lostmember", tenancy.GroupAdmin)
		lc := e.create(t, "lostmember", itHostA)
		ltok := e.activate(t, lc)
		if err := e.model.RemoveOrganizationMembership(ctx, e.orgA, lost.uid); err != nil {
			t.Fatal(err)
		}
		refused(t, lc, ltok, lost)
		// The five tools skip the C-6 check: the clear and logout add no denial row and no decision record.
		quiet := func(t *testing.T, u *itUser, call func()) {
			t.Helper()
			d0, r0 := e.mcpDenials(t, u), len(e.authorizeRecords(t))
			call()
			if d, r := e.mcpDenials(t, u), len(e.authorizeRecords(t)); d != d0 || r != r0 {
				t.Fatalf("control call ran the C-6 check: denials %d -> %d, records %d -> %d", d0, d, r0, r)
			}
		}
		quiet(t, lost, func() { e.mustCall(t, lc, ltok, ToolLogout, nil) })
		sel := e.newUser(t, "lostgroup")
		e.grantOrg(t, e.orgA, "lostgroup", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a2.ID, "lostgroup", tenancy.GroupAdmin)
		sc := e.create(t, "lostgroup", itHostA)
		stok := e.activate(t, sc)
		e.mustCall(t, sc, stok, ToolSwitchGroup, map[string]any{"group_id": e.a2.ID.String()})
		if err := e.model.RemoveGroupMembership(ctx, e.orgA, e.a2.ID, sel.uid); err != nil {
			t.Fatal(err)
		}
		refused(t, sc, stok, sel)
		quiet(t, sel, func() { e.mustCall(t, sc, stok, ToolSwitchGroup, map[string]any{"group_id": nil}) })
		quiet(t, sel, func() { e.mustCall(t, sc, stok, ToolLogout, nil) })
		// A store failure is unavailable for both names and records no denial.
		closedPool := itPool(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		closedPool.Close()
		cv, err := workos.NewConnectValidator(workos.ConnectConfig{Issuer: e.fake.ConnectIssuer(), Algorithms: []string{"RS256"},
			JWKS: e.fake.ConnectJWKS(), Now: e.clk.Now})
		if err != nil {
			t.Fatal(err)
		}
		down, err := New(Config{Issuer: e.fake.ConnectIssuer(), Hosts: staticLookup{hosts: map[string]string{itHostA: e.orgA.String()}},
			Verifier: cv, Identity: e.id, Pool: closedPool, Lifetimes: identity.MCPLifetimes{Pending: 30 * time.Minute, Idle: 30 * 24 * time.Hour, Absolute: 180 * 24 * time.Hour},
			Title: "Integration Product", Version: "0.0.0-it"})
		if err != nil {
			t.Fatal(err)
		}
		var outage []string
		for _, seam := range []bool{false, true} {
			if seam {
				useFixtures(top, down)
			} else {
				clearFixtures(down)
			}
			body := fmt.Sprintf(`{"jsonrpc":"2.0","id":4242,"method":"tools/call","params":{"name":%q,"arguments":{}}}`, hidden)
			req, _ := http.NewRequest(http.MethodPost, ConnectionPathPrefix+m.id, strings.NewReader(body))
			req.Host = itHostA
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")
			req.Header.Set("MCP-Protocol-Version", "2025-11-25")
			req.Header.Set("Authorization", "Bearer "+mtok)
			w := &recorder{h: http.Header{}}
			d0 := e.mcpDenials(t, member)
			down.Wrap(http.NotFoundHandler()).ServeHTTP(w, req)
			if w.code != http.StatusOK || !strings.Contains(w.body.String(), `"code":-32041`) || e.mcpDenials(t, member) != d0 {
				t.Fatalf("store down, seam %t: %d %s", seam, w.code, w.body.String())
			}
			outage = append(outage, w.body.String())
		}
		fixturesOn()
		if outage[0] != outage[1] {
			t.Fatalf("store down differs: %v", outage)
		}
		// Exact-target refusal despite an earlier listing: the guarded switch rechecks the target.
		x := e.newUser(t, "lister")
		e.grantOrg(t, e.orgA, "lister", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "lister", tenancy.GroupMember)
		e.grantGroup(t, e.orgA, e.a2.ID, "lister", tenancy.GroupMember)
		xc := e.create(t, "lister", itHostA)
		xtok := e.activate(t, xc)
		if n := len(e.mustCall(t, xc, xtok, ToolListGroups, nil)["groups"].([]any)); n != 2 {
			t.Fatalf("listed %d Groups", n)
		}
		if err := e.model.RemoveGroupMembership(ctx, e.orgA, e.a2.ID, x.uid); err != nil {
			t.Fatal(err)
		}
		e.wantToolError(t, xc, xtok, ToolSwitchGroup, map[string]any{"group_id": e.a2.ID.String()}, CodeGroupNotSelectable)
		if g, v := version(t, xc); g != "" || v != 0 {
			t.Fatalf("refused switch changed the selection: %q %d", g, v)
		}
		e.mustCall(t, xc, xtok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		if g, v := version(t, xc); g != e.a1.ID.String() || v != 1 {
			t.Fatalf("authorized switch: %q %d", g, v)
		}
	})

	t.Run("CAT-4 catalogs are connection-local", func(t *testing.T) {
		e.newUser(t, "isolated")
		e.grantOrg(t, e.orgA, "isolated", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "isolated", tenancy.GroupAdmin)
		e.grantGroup(t, e.orgA, e.a2.ID, "isolated", tenancy.GroupMember)
		e.grantOrg(t, e.orgB, "isolated", tenancy.OrgMember)
		e.grantGroup(t, e.orgB, e.b1.ID, "isolated", tenancy.GroupMember)
		e.bind(t, "isolated", itHostB) // G016 A2: a distinct credential issued at B's origin
		c1 := e.create(t, "isolated", itHostA)
		t1 := e.activate(t, c1)
		c2 := e.create(t, "isolated", itHostA)
		t2 := e.activate(t, c2)
		cb := e.create(t, "isolated", itHostB)
		tb := e.activate(t, cb)
		e.mustCall(t, c2, t2, ToolSwitchGroup, map[string]any{"group_id": e.a2.ID.String()})
		same(t, "C2 before", e.list(t, c2, t2), sorted())
		same(t, "B before", e.list(t, cb, tb), sorted())
		e.mustCall(t, c1, t1, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		inA1 := sorted(fixtureGroupUpdate, fixtureProfileUpdate)
		same(t, "C1 with A1", e.list(t, c1, t1), inA1)
		same(t, "C2 after C1's select", e.list(t, c2, t2), sorted())
		same(t, "B after C1's select", e.list(t, cb, tb), sorted())
		// A fresh SDK session (reconnect) on C1 lists C1's stored context, nothing of C2's.
		cs, err := e.dial(t, c1, e.tok(c1))
		if err != nil {
			t.Fatal(err)
		}
		res, err := cs.ListTools(ctx, nil)
		_ = cs.Close()
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, tool := range res.Tools {
			names = append(names, tool.Name)
		}
		same(t, "C1 after reconnect", names, inA1)
		// Renewal-due calls list fresh and keep each connection's own context.
		e.clk.Advance(29 * 24 * time.Hour)
		same(t, "C2 renewed", e.list(t, c2, e.tok(c2)), sorted())
		same(t, "C1 renewed", e.list(t, c1, e.tok(c1)), inA1)
		if g, v := version(t, c1); g != e.a1.ID.String() || v != 1 {
			t.Fatalf("C1: %q %d", g, v)
		}
		if g, v := version(t, c2); g != e.a2.ID.String() || v != 1 {
			t.Fatalf("C2: %q %d", g, v)
		}
		if g, v := version(t, cb); g != "" || v != 0 {
			t.Fatalf("B: %q %d", g, v)
		}
	})

	t.Run("CAT-5 no notification; the selection revision is not a catalog signal", func(t *testing.T) {
		a3 := e.group(t, e.orgA, "ownerA", "Group A3", fmt.Sprintf("a3-%d", itSeq.Add(1)))
		u := e.newUser(t, "notified")
		e.grantOrg(t, e.orgA, "notified", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "notified", tenancy.GroupAdmin)
		e.grantGroup(t, e.orgA, e.a2.ID, "notified", tenancy.GroupAdmin)
		e.grantGroup(t, e.orgA, a3.ID, "notified", tenancy.GroupMember)
		c := e.create(t, "notified", itHostA)
		tok := e.activate(t, c)
		var notified atomic.Int64
		opts := &sdk.ClientOptions{ToolListChangedHandler: func(context.Context, *sdk.ToolListChangedRequest) { notified.Add(1) }}
		cs, err := e.dialWith(t, c, tok, opts)
		if err != nil {
			t.Fatal(err)
		}
		defer cs.Close()
		listed := func(t *testing.T) []string {
			t.Helper()
			res, err := cs.ListTools(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, tool := range res.Tools {
				names = append(names, tool.Name)
			}
			return names
		}
		revision := func(t *testing.T) any {
			t.Helper()
			res, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: ToolGetCurrentContext})
			if err != nil || res.IsError {
				t.Fatalf("context: %v %+v", err, res)
			}
			return res.StructuredContent.(map[string]any)["context_version"]
		}
		full := sorted(fixtureGroupUpdate, fixtureProfileUpdate)
		same(t, "initial", listed(t), full)
		// Local delta: the select's response carries the new revision and the re-list differs.
		r0 := revision(t)
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": a3.ID.String()})
		r1 := revision(t)
		same(t, "A3 selected", listed(t), sorted())
		// External delta: observed on the next list with the revision unchanged.
		if _, err := e.model.ChangeGroupRole(ctx, e.orgA, a3.ID, u.uid, tenancy.GroupAdmin); err != nil {
			t.Fatal(err)
		}
		same(t, "after the external grant", listed(t), sorted(fixtureGroupUpdate))
		if r := revision(t); r != r1 {
			t.Fatalf("external change moved the revision: %v -> %v", r1, r)
		}
		// A no-op select leaves the revision and the list identical.
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": a3.ID.String()})
		if r := revision(t); r != r1 {
			t.Fatalf("no-op select moved the revision: %v -> %v", r1, r)
		}
		same(t, "after the no-op select", listed(t), sorted(fixtureGroupUpdate))
		// A select between Groups with equal catalogs bumps the revision with an identical list.
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		ra := revision(t)
		inA1 := listed(t)
		e.mustCall(t, c, tok, ToolSwitchGroup, map[string]any{"group_id": e.a2.ID.String()})
		rb := revision(t)
		same(t, "A2 versus A1", listed(t), inA1)
		same(t, "A1", inA1, full)
		if r0 == r1 || r1 == ra || ra == rb {
			t.Fatalf("revisions %v %v %v %v", r0, r1, ra, rb)
		}
		// Reconnect: a new SDK session re-lists the same stored context.
		cs2, err := e.dialWith(t, c, e.tok(c), opts)
		if err != nil {
			t.Fatal(err)
		}
		res, err := cs2.ListTools(ctx, nil)
		_ = cs2.Close()
		if err != nil || len(res.Tools) != len(full) {
			t.Fatalf("reconnect list: %v %v", err, res)
		}
		if n := notified.Load(); n != 0 {
			t.Fatalf("%d list-changed notifications received", n)
		}
		if cs.InitializeResult().Capabilities.Tools.ListChanged {
			t.Fatal("listChanged advertised")
		}
	})

	// G020 (Plan r2 §4.3, T10; D-1 decided): label-capability parity through the three test-only label fixture tools on
	// the same seam. This is evidence of finite label-capability eligibility, current-context refusal and hidden
	// direct-call behaviour only, never of a production label operation (the production catalog stays the five control
	// tools: TestToolCatalog and CAT-1's production row). Subjects (the access fixture mapping of the G020 Result): frank
	// is an Organization admin of A without any Group binding; bob administers A1 and is a plain member of A2; dave is a
	// member of A1; dana holds only the Organization membership.
	t.Run("CAT-6 label capabilities: Organization staff never reach the Group label tools; hidden equals unknown", func(t *testing.T) {
		labels := useLabelFixtures(top, e.h)
		defer fixturesOn() // the G015 four-tool seam for anything after this subtest
		labelsOn := func() { labels = useLabelFixtures(top, e.h) }
		setOverride := func(on bool) {
			t.Helper()
			if _, err := e.setup.Exec(ctx, `UPDATE app.organizations SET group_management_override = $2 WHERE id = $1::text::uuid`,
				e.orgA.String(), on); err != nil {
				t.Fatal(err)
			}
		}
		e.newUser(t, "labelfrank")
		e.grantOrg(t, e.orgA, "labelfrank", tenancy.OrgAdmin)
		e.newUser(t, "labelbob")
		e.grantOrg(t, e.orgA, "labelbob", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "labelbob", tenancy.GroupAdmin)
		e.grantGroup(t, e.orgA, e.a2.ID, "labelbob", tenancy.GroupMember)
		e.newUser(t, "labeldave")
		e.grantOrg(t, e.orgA, "labeldave", tenancy.OrgMember)
		e.grantGroup(t, e.orgA, e.a1.ID, "labeldave", tenancy.GroupMember)
		e.newUser(t, "labeldana")
		e.grantOrg(t, e.orgA, "labeldana", tenancy.OrgMember)
		orgTool := sorted(fixtureLabelsListOrganization)
		groupTools := sorted(fixtureLabelsListGroup, fixtureLabelsReadGroup)

		// frank: the Organization tool only, with the override off or on, Organization-only or with A1 selected.
		fc := e.create(t, "labelfrank", itHostA)
		ftok := e.activate(t, fc)
		same(t, "frank, override off", e.list(t, fc, ftok), orgTool)
		setOverride(true)
		t.Cleanup(func() { setOverride(false) })
		same(t, "frank, override on", e.list(t, fc, ftok), orgTool)
		fsel := e.create(t, "labelfrank", itHostA)
		fstok := e.activate(t, fsel)
		e.mustCall(t, fsel, fstok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		same(t, "frank, override on, A1 selected", e.list(t, fsel, fstok), orgTool)
		setOverride(false)
		same(t, "frank, override off, A1 selected", e.list(t, fsel, fstok), orgTool)

		// bob: both Group tools through A1 and never the Organization tool; with A2 selected only the read tool.
		bc := e.create(t, "labelbob", itHostA)
		btok := e.activate(t, bc)
		same(t, "bob, Organization-only", e.list(t, bc, btok), groupTools)
		bsel := e.create(t, "labelbob", itHostA)
		bstok := e.activate(t, bsel)
		e.mustCall(t, bsel, bstok, ToolSwitchGroup, map[string]any{"group_id": e.a1.ID.String()})
		same(t, "bob, A1 selected", e.list(t, bsel, bstok), groupTools)
		e.mustCall(t, bsel, bstok, ToolSwitchGroup, map[string]any{"group_id": e.a2.ID.String()})
		same(t, "bob, A2 selected", e.list(t, bsel, bstok), sorted(fixtureLabelsReadGroup))

		// dave: the read tool through A1; dana: none.
		dc := e.create(t, "labeldave", itHostA)
		same(t, "dave", e.list(t, dc, e.activate(t, dc)), sorted(fixtureLabelsReadGroup))
		nc := e.create(t, "labeldana", itHostA)
		same(t, "dana", e.list(t, nc, e.activate(t, nc)), sorted())

		// Hidden equals unknown (C-1): the same requested name with the seam empty and set but ineligible answers
		// byte-identically and runs no handler.
		hiddenIsUnknown := func(t *testing.T, what string, c conn, tok, name string) {
			t.Helper()
			clearFixtures(e.h)
			code1, unknown := e.rawCall(t, c, tok, name)
			labelsOn()
			code2, hid := e.rawCall(t, c, tok, name)
			if code1 != http.StatusOK || code2 != http.StatusOK || unknown != hid ||
				!strings.Contains(unknown, `"code":-32602`) || !strings.Contains(unknown, `unknown tool \"`+name+`\"`) {
				t.Fatalf("%s: hidden versus unknown:\n%d %s\n%d %s", what, code1, unknown, code2, hid)
			}
			if labels.get(name) != 0 {
				t.Fatalf("%s: the hidden call ran %s", what, name)
			}
		}
		hiddenIsUnknown(t, "frank calls the Group manage tool", fc, ftok, fixtureLabelsListGroup)
		hiddenIsUnknown(t, "frank calls the Group read tool", fc, ftok, fixtureLabelsReadGroup)
		hiddenIsUnknown(t, "bob calls the Organization tool", bc, btok, fixtureLabelsListOrganization)
		hiddenIsUnknown(t, "bob with A2 selected calls the Group manage tool", bsel, bstok, fixtureLabelsListGroup)
		hiddenIsUnknown(t, "dana calls the Group read tool", nc, e.tok(nc), fixtureLabelsReadGroup)
		// The paired eligible calls run their handler once each.
		if out := e.mustCall(t, fc, ftok, fixtureLabelsListOrganization, nil); out["invoked"] != true || labels.get(fixtureLabelsListOrganization) != 1 {
			t.Fatalf("frank's Organization tool call: %v %d", out, labels.get(fixtureLabelsListOrganization))
		}
		if out := e.mustCall(t, bc, btok, fixtureLabelsListGroup, nil); out["invoked"] != true || labels.get(fixtureLabelsListGroup) != 1 {
			t.Fatalf("bob's Group tool call: %v %d", out, labels.get(fixtureLabelsListGroup))
		}
		// An external downgrade governs the next call: bob demoted to member of A1 loses the manage tool (hidden, unknown).
		if _, err := e.model.ChangeGroupRole(ctx, e.orgA, e.a1.ID, e.users["labelbob"].uid, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		same(t, "bob demoted", e.list(t, bc, btok), sorted(fixtureLabelsReadGroup))
		hiddenIsUnknown(t, "bob demoted calls the Group manage tool", bc, btok, fixtureLabelsListGroup)
	})
}

// TestMCPRecovery is IT-MCP-10's recovery case with the actual G013 restore invalidation (infra/recovery/sql/
// invalidate.sql at the bound base, passed by the harness): pending and active connections refuse afterwards, a deleted
// (post-checkpoint) connection refuses, and a fresh setup works. It must run last in its database: the invalidation
// expires every session.
func TestMCPRecovery(t *testing.T) {
	e := setupMCP(t)
	ctx := context.Background()
	raw, err := base64.StdEncoding.DecodeString(os.Getenv("MCP_INVALIDATE_SQL_B64"))
	if err != nil || len(raw) == 0 {
		t.Fatal("MCP_INVALIDATE_SQL_B64 is required")
	}
	if sum := sha256.Sum256(raw); hex.EncodeToString(sum[:]) != "931b4e066807fc2dfcb96d0949380c90b3a780965325c605c5d9fa504014e312" {
		t.Fatal("the invalidation SQL is not the bound G013 file")
	}
	pending := e.create(t, "memberA1", itHostA)
	active := e.create(t, "memberA1", itHostA)
	tok := e.activate(t, active)
	gone := e.create(t, "memberA1", itHostA)
	gtok := e.activate(t, gone)
	// The post-checkpoint connection is absent after a restore: remove its rows (superuser), as a restore would.
	if _, err := e.setup.Exec(ctx, `DELETE FROM app.mcp_connections WHERE id = $1::text::uuid`, gone.id); err != nil {
		t.Fatal(err)
	}
	// G063 (Plan r1 §6, r1 C5 browser-auth layer): the rename route is served; the pre-invalidation browser credential and
	// its CSRF token are captured before the invalidation.
	serveRename(t, e, e.h)
	preCred := e.cred(t, "memberA1", itHostA)
	preCSRF := e.csrf(t, preCred)
	if _, err := e.setup.Exec(ctx, string(raw)); err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	// The invalidation expires sessions at the database clock; the service clock of this environment is fixed at setup,
	// so it moves past the invalidation time, as real time does after a restore.
	e.clk.Advance(time.Minute)
	e.want401(t, active, tok, "active after invalidation")
	e.want401(t, pending, e.tok(pending), "pending after invalidation")
	e.want401(t, gone, gtok, "absent after restore")
	// G063: a rename of the pending and the active connection with the pre-invalidation credential is unauthenticated
	// (the caller's session is invalidated before any target lookup); nothing changes.
	for label, c := range map[string]conn{"pending": pending, "active": active} {
		code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/rename", preCred, map[string]any{"origin": "https://" + itHostA,
			"connection_id": c.id, "csrf_token": preCSRF, "display_name": "After restore"})
		if code != http.StatusUnauthorized || out["error"] != "unauthenticated" {
			t.Errorf("%s rename with the invalidated credential: %d %v, want 401 unauthenticated", label, code, out)
		}
	}
	// A fresh browser sign-in and setup work.
	e.bind(t, "memberA1", itHostA) // the explicit re-sign-in after the invalidation (G016 A2: at A's origin)
	// G063: after the re-sign-in the invalidated connections are terminal targets (not_found) and the absent one unknown.
	for label, c := range map[string]conn{"pending": pending, "active": active, "absent": gone} {
		cred := e.cred(t, "memberA1", itHostA)
		code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/rename", cred, map[string]any{"origin": "https://" + itHostA,
			"connection_id": c.id, "csrf_token": e.csrf(t, cred), "display_name": "After restore"})
		if code != http.StatusNotFound || out["error"] != "not_found" {
			t.Errorf("%s rename after the re-sign-in: %d %v, want 404 not_found", label, code, out)
		}
	}
	var renamed int
	if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.mcp_connections WHERE id = ANY($1::text[]::uuid[]) AND display_name IS NOT NULL`,
		[]string{pending.id, active.id}).Scan(&renamed); err != nil || renamed != 0 {
		t.Errorf("an invalidated connection was renamed: %d %v", renamed, err)
	}
	fresh := e.create(t, "memberA1", itHostA)
	ft := e.activate(t, fresh)
	e.mustCall(t, fresh, ft, ToolGetCurrentContext, nil)
}

// recorder is a minimal http.ResponseWriter for direct handler calls.
type recorder struct {
	h    http.Header
	code int
	body strings.Builder
}

func (r *recorder) Header() http.Header { return r.h }
func (r *recorder) Write(b []byte) (int, error) {
	if r.code == 0 {
		r.code = http.StatusOK
	}
	return r.body.Write(b)
}
func (r *recorder) WriteHeader(code int) {
	if r.code == 0 {
		r.code = code
	}
}

// setupListingPages is IT-MCP-3's setup listing pagination (Duo follow-up correction; proposal §4.1 and addendum 01):
// 101 live connections with a timestamp tie across the page boundary, newest-first traversal without omissions or
// duplicates, anchor deletion, an older row revoked from page 2, the 100-row boundary and new-URL visibility, malformed
// and foreign-context cursors (including an alias origin of the same Organization), edited positions that still list
// only the caller's own live rows, the time domain, and per-page live session and membership authority.
func setupListingPages(t *testing.T, e *mcpEnv) {
	t.Helper()
	ctx := context.Background()
	pager := e.newUser(t, "pager")
	e.grantOrg(t, e.orgA, "pager", tenancy.OrgMember)
	e.grantOrg(t, e.orgB, "pager", tenancy.OrgMember)
	originA, originB, aliasHost := "https://"+itHostA, "https://"+itHostB, "org-a-alias.localhost:543"
	// G016 A2: the alias origin is issued for Organization A (as the alias registry below maps it); pager holds a
	// distinct credential issued at A's, B's and the alias origin.
	e.route(t, aliasHost, e.orgA)
	e.bind(t, "pager", itHostB, aliasHost)
	pagerA, pagerB := e.cred(t, "pager", itHostA), e.cred(t, "pager", itHostB)
	list := func(cred, origin string, extra ...string) (int, map[string]any) {
		t.Helper()
		q := url.Values{"origin": {origin}}
		for i := 0; i+1 < len(extra); i += 2 {
			q.Add(extra[i], extra[i+1])
		}
		return e.internal(t, http.MethodGet, "/internal/v1/mcp/connections?"+q.Encode(), cred, nil)
	}
	type row struct {
		id string
		at time.Time
	}
	ids := func(out map[string]any) []string {
		got := []string{}
		cs, _ := out["connections"].([]any)
		for _, c := range cs {
			got = append(got, c.(map[string]any)["connection_id"].(string))
		}
		return got
	}
	page := func(cursor string) ([]string, any) {
		t.Helper()
		var code int
		var out map[string]any
		if cursor == "" {
			code, out = list(pagerA, originA)
		} else {
			code, out = list(pagerA, originA, "cursor", cursor)
		}
		if code != http.StatusOK || len(out) != 3 {
			t.Fatalf("listing page: %d %v", code, out)
		}
		return ids(out), out["next_cursor"]
	}
	refused := func(name string, code int, out map[string]any, wantCode int, want string) {
		t.Helper()
		if code != wantCode || out["error"] != want || len(out) != 1 {
			t.Errorf("%s: %d %v (want %d %s)", name, code, out, wantCode, want)
		}
	}
	revoke := func(id string) {
		t.Helper()
		code, out := e.internal(t, http.MethodPost, "/internal/v1/mcp/connections/revoke", pagerA,
			map[string]string{"origin": originA, "connection_id": id, "csrf_token": e.csrf(t, pagerA)})
		if code != http.StatusOK || out["outcome"] != "revoked" {
			t.Fatalf("revoke: %d %v", code, out)
		}
	}
	// 3 tied connections at one instant, then 98 at distinct later instants: newest first, the tie takes positions
	// 99-101, so the page boundary falls inside it.
	var made []row
	var moved time.Duration
	tick := func() {
		e.clk.Advance(time.Millisecond)
		moved += time.Millisecond
	}
	defer func() { e.clk.Advance(-moved) }()
	for i := 0; i < 101; i++ {
		if i >= 3 {
			tick()
		}
		made = append(made, row{e.create(t, "pager", itHostA).id, e.clk.Now()})
	}
	want := slices.Clone(made)
	slices.SortFunc(want, func(a, b row) int {
		if c := b.at.Compare(a.at); c != 0 {
			return c
		}
		return strings.Compare(b.id, a.id)
	})
	order := func(rs []row) []string {
		out := []string{}
		for _, r := range rs {
			out = append(out, r.id)
		}
		return out
	}
	p1, next := page("")
	cursor, _ := next.(string)
	if !slices.Equal(p1, order(want[:100])) || cursor == "" || p1[0] != made[100].id {
		t.Fatalf("page 1: %d rows, next %v, newest first %v", len(p1), next, slices.Equal(p1, order(want[:100])))
	}
	p2, next2 := page(cursor)
	if !slices.Equal(p2, order(want[100:])) || next2 != nil {
		t.Fatalf("page 2: %v next %v (want %v)", p2, next2, order(want[100:]))
	}
	if again, n := page(""); !slices.Equal(again, p1) || n != next {
		t.Fatal("a second traversal differs")
	}
	// The anchor (page 1's last row) revoked: the old cursor still resumes after its position.
	revoke(p1[99])
	if p2b, _ := page(cursor); !slices.Equal(p2b, p2) {
		t.Fatalf("page 2 after the anchor was revoked: %v", p2b)
	}
	// An older, page-2 row revoked: it is gone from every page; 99 rows remain on one page.
	revoke(p2[0])
	if p, n := page(""); len(p) != 99 || n != nil || slices.Contains(p, p2[0]) || slices.Contains(p, p1[99]) {
		t.Fatalf("after the older revoke: %d rows, next %v", len(p), n)
	}
	if p, n := page(cursor); len(p) != 0 || n != nil {
		t.Fatalf("page 2 after the older revoke: %v %v", p, n)
	}
	// Exactly 100 rows: no next page, and the new connection is the first row. One more: a next page again.
	tick()
	fresh := e.create(t, "pager", itHostA)
	if p, n := page(""); len(p) != 100 || n != nil || p[0] != fresh.id {
		t.Fatalf("100 rows: %d, next %v", len(p), n)
	}
	tick()
	newest := e.create(t, "pager", itHostA)
	p1, next = page("")
	if cursor, _ = next.(string); len(p1) != 100 || cursor == "" || p1[0] != newest.id {
		t.Fatalf("101 rows again: %d, next %v", len(p1), next)
	}
	// Malformed cursors and query shapes: 400 invalid_request.
	tag := setupCursorTag(pager.id, e.orgA.String(), originA)
	raw := func(micro, id string) string {
		return base64.RawURLEncoding.EncodeToString([]byte("s1|" + micro + "|" + id + "|" + tag))
	}
	for name, extra := range map[string][]string{
		"empty": {"cursor", ""}, "garbage": {"cursor", "!!"}, "padded": {"cursor", cursor + "="},
		"overlong":     {"cursor", strings.Repeat("A", 129)},
		"tool cursor":  {"cursor", encodeCursor(e.orgA.String(), p1[0])},
		"altered tag":  {"cursor", base64.RawURLEncoding.EncodeToString([]byte("s1|1|" + p1[0] + "|g" + tag[1:]))},
		"leading zero": {"cursor", raw("01", p1[0])}, "zero UUID": {"cursor", raw("1", "00000000-0000-0000-0000-000000000000")},
		"time zero": {"cursor", raw("0", p1[0])}, "above the domain": {"cursor", raw("253402300800000000", p1[0])},
		"int64 overflow": {"cursor", raw("9223372036854775808", p1[0])},
		"repeated key":   {"cursor", cursor, "cursor", cursor}, "unknown key": {"cursor", cursor, "limit", "5"},
	} {
		code, out := list(pagerA, originA, extra...)
		refused(name, code, out, http.StatusBadRequest, "invalid_request")
	}
	// Foreign contexts: another user, another Organization's origin, and an alias origin of the same Organization.
	code, out := list(e.cred(t, "memberA1", itHostA), originA, "cursor", cursor)
	refused("another user's cursor", code, out, http.StatusBadRequest, "invalid_request")
	code, out = list(pagerB, originB, "cursor", cursor)
	refused("an org-A cursor on org B", code, out, http.StatusBadRequest, "invalid_request")
	alias := e.handler(t, map[string]tenancy.OrganizationID{itHostA: e.orgA, aliasHost: e.orgA}, nil)
	viaAlias := func(host string) (int, map[string]any) { // G016 A2: pager's credential issued at that host's origin
		req := httptest.NewRequest(http.MethodGet, "/internal/v1/mcp/connections?"+url.Values{"origin": {"https://" + host}, "cursor": {cursor}}.Encode(), nil)
		req.Header.Set("Authorization", "Session "+e.cred(t, "pager", host))
		w := httptest.NewRecorder()
		alias.List(w, req)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	code, out = viaAlias(aliasHost)
	refused("an alias origin of the same Organization", code, out, http.StatusBadRequest, "invalid_request")
	if code, out = viaAlias(itHostA); code != http.StatusOK || len(ids(out)) != 1 {
		t.Errorf("the same cursor on its own origin through the alias registry: %d %v", code, out)
	}
	// Authority comes before the cursor: a non-member and an anonymous caller get the usual answers.
	code, out = list(e.cred(t, "nonmember", itHostA), originA, "cursor", cursor)
	refused("a non-member with a cursor", code, out, http.StatusForbidden, "not_available")
	code, out = list("", originA, "cursor", cursor)
	refused("no session with a cursor", code, out, http.StatusUnauthorized, "unauthenticated")
	// Edited positions with a matching context tag stay inside the caller's own live rows.
	own := map[string]bool{}
	for _, cur := range []string{"", cursor} {
		p, _ := page(cur)
		for _, id := range p {
			own[id] = true
		}
	}
	foreignA := e.create(t, "memberA1", itHostA)
	foreignB := e.create(t, "cross", itHostB)
	at := func(id string) time.Time {
		var ts time.Time
		if err := e.setup.QueryRow(ctx, `SELECT created_at FROM app.mcp_connections WHERE id = $1`, id).Scan(&ts); err != nil {
			t.Fatal(err)
		}
		return ts
	}
	for name, p := range map[string]identity.MCPListPosition{
		"another user's anchor":     {CreatedAt: at(foreignA.id), ID: foreignA.id},
		"an org-B anchor":           {CreatedAt: at(foreignB.id), ID: foreignB.id},
		"a random anchor":           {CreatedAt: e.clk.Now().Add(-time.Hour), ID: "7e000000-0000-4000-8000-0000000000e7"},
		"the domain maximum":        {CreatedAt: time.UnixMicro(setupCursorMaxMicro), ID: "7e000000-0000-4000-8000-0000000000e8"},
		"the domain minimum":        {CreatedAt: time.UnixMicro(setupCursorMinMicro), ID: "7e000000-0000-4000-8000-0000000000e9"},
		"a position inside the tie": {CreatedAt: made[0].at, ID: "ffffffff-ffff-4fff-bfff-ffffffffffff"},
	} {
		c, ok := encodeSetupCursor(pager.id, e.orgA.String(), originA, p)
		if !ok {
			t.Fatalf("%s: unencodable", name)
		}
		got, _ := page(c)
		for _, id := range got {
			if !own[id] {
				t.Errorf("%s listed a row that is not the caller's own live connection", name)
			}
		}
		switch name {
		case "the domain maximum":
			if !slices.Equal(got, p1) {
				t.Errorf("%s: not the first page", name)
			}
		case "the domain minimum":
			if len(got) != 0 {
				t.Errorf("%s: %d rows", name, len(got))
			}
		}
	}
	// Per-page live authority: membership lost after page 1 refuses page 2; restored, it lists again; an ended
	// browser session is unauthenticated.
	if err := e.model.RemoveOrganizationMembership(ctx, e.orgA, pager.uid); err != nil {
		t.Fatal(err)
	}
	code, out = list(pagerA, originA, "cursor", cursor)
	refused("page 2 after membership loss", code, out, http.StatusForbidden, "not_available")
	e.grantOrg(t, e.orgA, "pager", tenancy.OrgMember)
	if p, _ := page(cursor); len(p) != len(own)-100 {
		t.Errorf("page 2 after the membership was restored: %d rows", len(p))
	}
	if _, err := e.id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: e.csrf(t, pagerA), Origin: originA, Credential: pagerA}); err != nil {
		t.Fatal(err)
	}
	code, out = list(pagerA, originA, "cursor", cursor)
	refused("page 2 after the browser session ended", code, out, http.StatusUnauthorized, "unauthenticated")
	e.bind(t, "pager", itHostA) // the explicit re-sign-in after the logout (G016 A2: at A's origin)
}

// serveRename routes the environment's HTTP server through h with the four setup routes, the G063 rename included
// (env_integration_test.go's server and serveWith mount the three pre-G063 routes), until the calling test ends; it then
// restores the previous handler and server.
func serveRename(t *testing.T, e *mcpEnv, h *Handler) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc(RouteConnectionsList, h.List)
	mux.HandleFunc(RouteConnectionsCreate, h.Create)
	mux.HandleFunc(RouteConnectionsRevoke, h.Revoke)
	mux.HandleFunc(RouteConnectionsRename, h.Rename)
	srv := httptest.NewServer(observability.Middleware(observability.NewLogger(e.logs, slog.LevelInfo, nil), h.Wrap(mux)))
	prevSrv, prevH := e.srv, e.h
	e.srv, e.h = srv, h
	t.Cleanup(func() {
		srv.Close()
		e.srv, e.h = prevSrv, prevH
	})
}

// rawSetup posts raw body bytes to an internal setup route with the browser credential (a second Authorization header
// when twice is set) and returns the status and the trimmed body.
func rawSetup(t *testing.T, e *mcpEnv, path, body, cred string, twice bool) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, e.srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cred != "" {
		req.Header.Add("Authorization", "Session "+cred)
	}
	if twice {
		req.Header.Add("Authorization", "Session "+cred)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, strings.TrimSpace(string(b))
}

// strs is a comparable form of an optional name.
func strs(s *string) []string {
	if s == nil {
		return nil
	}
	return []string{*s}
}
