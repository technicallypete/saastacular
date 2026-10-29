//go:build integration

// G014 identity integration (B-IDENTITY-IT; Plan r8 §10.2): IT-MCP-ID-1..5 against the version-10 database of the
// identity project, through the exported identity API and the actual runtime role. The access facade's guarded
// orchestration (M1-M3) is B-MCP-IT's; here the identity statements, the row-level security proofs, the M5 renewal and
// the revocation paths are exercised directly. Fixtures (Organizations, membership bindings, fault constraints) use the
// attested superuser; assertions never do.
package identity_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
)

var mcpLife = identity.MCPLifetimes{Pending: 30 * time.Minute, Idle: 30 * 24 * time.Hour, Absolute: 180 * 24 * time.Hour}

// mcpFixture is one signed-in user with an Organization membership.
type mcpFixture struct {
	h     *harness
	sub   string
	cred  string
	user  string
	org   string
	other string // an Organization without a membership of user
}

func newMCPFixture(t *testing.T) *mcpFixture {
	t.Helper()
	resetIdentity(t)
	h := newHarness(t)
	f := &mcpFixture{h: h, sub: fmt.Sprintf("user_01MCPID%010d", time.Now().UnixNano()%1e10)}
	f.cred = h.mustSignIn(originA, grant(f.sub, "session_01MCPID")).Credential
	info, err := h.svc.Session(context.Background(), f.cred, originA)
	if err != nil {
		t.Fatal(err)
	}
	f.user = info.UserID
	setup := setupConn(t)
	ctx := context.Background()
	if err := setup.QueryRow(ctx, `INSERT INTO app.organizations (name) VALUES ('MCP Identity Org') RETURNING id::text`).Scan(&f.org); err != nil {
		t.Fatal(err)
	}
	if err := setup.QueryRow(ctx, `INSERT INTO app.organizations (name) VALUES ('MCP Other Org') RETURNING id::text`).Scan(&f.other); err != nil {
		t.Fatal(err)
	}
	if _, err := setup.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, 'org_member', 'membership')`, f.org, f.user); err != nil {
		t.Fatal(err)
	}
	return f
}

// tenantTx opens a runtime transaction in the TENANT-ORG shape of (user, org) with the literal context statement.
func tenantTx(t *testing.T, user, org string) pgx.Tx {
	t.Helper()
	tx, err := it.pool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	if _, err := tx.Exec(context.Background(), `SELECT pg_catalog.set_config('app.user_id', $1, true),
  pg_catalog.set_config('app.organization_id', $2, true)`, user, org); err != nil {
		t.Fatal(err)
	}
	return tx
}

// create inserts a pending connection for the fixture user in org and commits; it returns the connection ID.
func (f *mcpFixture) create(t *testing.T, org string) (string, error) {
	t.Helper()
	id, err := identity.NewMCPConnectionID()
	if err != nil {
		t.Fatal(err)
	}
	tx := tenantTx(t, f.user, org)
	if err := f.h.svc.InsertMCPConnectionTx(context.Background(), tx, f.user, org, id, "org-a.localhost:543",
		"https://org-a.localhost:543/mcp/c/"+id, mcpLife, nil); err != nil {
		return "", err
	}
	return id, tx.Commit(context.Background())
}

func (f *mcpFixture) mustCreate(t *testing.T) string {
	t.Helper()
	id, err := f.create(t, f.org)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return id
}

func (f *mcpFixture) activate(t *testing.T, id, client string) identity.MCPActivation {
	t.Helper()
	c, err := f.h.svc.ReadMCPConnection(context.Background(), f.user, f.org, id)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	tx := tenantTx(t, f.user, f.org)
	a, err := f.h.svc.ActivateMCPConnectionTx(context.Background(), tx, c, client)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestMCPIdentityCreate(t *testing.T) {
	f := newMCPFixture(t)
	// IT-MCP-ID-1: the insert policies require the TENANT-ORG shape with a current Organization role.
	if _, err := f.create(t, f.other); err == nil {
		t.Fatal("a connection was created in an Organization without membership")
	}
	if _, err := f.create(t, ""); err == nil {
		t.Fatal("a connection was created in the IDENTITY shape")
	}
	id := f.mustCreate(t)
	c, err := f.h.svc.ReadMCPConnection(context.Background(), f.user, f.org, id)
	if err != nil || c.Active() || c.Resource() != "https://org-a.localhost:543/mcp/c/"+id || c.OrganizationID() != f.org {
		t.Fatalf("pending connection: %+v %v", c, err)
	}
	// The connection is invisible from another Organization's shape and to another user.
	if _, err := f.h.svc.ReadMCPConnection(context.Background(), f.user, f.other, id); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("read from another Organization: %v", err)
	}
	g := f.h.mustSignIn(originA, grant(f.sub+"X", "session_01MCPIDX"))
	info, err := f.h.svc.Session(context.Background(), g.Credential, originA)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.h.svc.ReadMCPConnection(context.Background(), info.UserID, f.org, id); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("read by another user: %v", err)
	}
	// IT-MCP-ID-3: E1 never resolves an mcp session, even with a known credential hash planted by the superuser.
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	known := base64.RawURLEncoding.EncodeToString(raw)
	if _, err := setupConn(t).Exec(context.Background(), `UPDATE app.sessions SET credential_hash = sha256($1::bytea)
WHERE id = (SELECT session_id FROM app.mcp_connections WHERE id = $2::text::uuid)`, raw, id); err != nil {
		t.Fatal(err)
	}
	if _, err := f.h.svc.Session(context.Background(), known, originA); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("E1 resolved an mcp session: %v", err)
	}
	// The subject maps only through an existing principal: an unmapped subject creates nothing.
	var before, after int
	setup := setupConn(t)
	if err := setup.QueryRow(context.Background(), `SELECT count(*) FROM app.users`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err := f.h.svc.MapConnectSubject(context.Background(), "user_01NEVERSEEN"); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("unmapped subject: %v", err)
	}
	if err := setup.QueryRow(context.Background(), `SELECT count(*) FROM app.users`).Scan(&after); err != nil || after != before {
		t.Fatalf("users %d -> %d (%v)", before, after, err)
	}
	if u, err := f.h.svc.MapConnectSubject(context.Background(), f.sub); err != nil || u != f.user {
		t.Fatalf("mapped subject: %q %v", u, err)
	}
}

func TestMCPIdentityActivation(t *testing.T) {
	f := newMCPFixture(t)
	ctx := context.Background()
	// IT-MCP-ID-2: concurrent first tokens: exactly one activation; the client is bound once.
	id := f.mustCreate(t)
	c, err := f.h.svc.ReadMCPConnection(ctx, f.user, f.org, id)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make([]identity.MCPActivation, 6)
	errs := make([]error, 6)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tx, err := it.pool.Begin(ctx)
			if err != nil {
				errs[i] = err
				return
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SELECT pg_catalog.set_config('app.user_id', $1, true), pg_catalog.set_config('app.organization_id', $2, true)`,
				f.user, f.org); err != nil {
				errs[i] = err
				return
			}
			if results[i], errs[i] = f.h.svc.ActivateMCPConnectionTx(ctx, tx, c, fmt.Sprintf("client_01ACT%d", i)); errs[i] == nil {
				errs[i] = tx.Commit(ctx)
			}
		}(i)
	}
	wg.Wait()
	winners, bound := 0, ""
	for i, r := range results {
		if errs[i] != nil {
			t.Fatalf("activation %d: %v", i, errs[i])
		}
		if r.Activated {
			winners++
			bound = r.ClientID
		}
	}
	for _, r := range results {
		if !r.Activated && (r.State != identity.MCPStateActive || r.ClientID != bound) {
			t.Fatalf("a losing activation saw %+v, want the bound client %s", r, bound)
		}
	}
	if winners != 1 {
		t.Fatalf("%d activations won", winners)
	}
	// An expired pending connection never activates.
	late := f.mustCreate(t)
	f.h.clk.Advance(31 * time.Minute)
	if _, err := f.h.svc.ReadMCPConnection(ctx, f.user, f.org, late); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("expired pending read: %v", err)
	}
	f.h.clk.Advance(-31 * time.Minute)
	// The guard refuses a second activation and any binding change, even through the runtime role directly.
	tx := tenantTx(t, f.user, f.org)
	if _, err := tx.Exec(ctx, `UPDATE app.mcp_connections SET client_id = 'client_01OTHER' WHERE id = $1::text::uuid`, id); err == nil {
		t.Fatal("the bound client changed")
	}
}

func TestMCPIdentityRenewalAndRevocation(t *testing.T) {
	f := newMCPFixture(t)
	ctx := context.Background()
	read := func(id string) identity.MCPConnection {
		t.Helper()
		c, err := f.h.svc.ReadMCPConnection(ctx, f.user, f.org, id)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		return c
	}
	idleOf := func(id string) (idle, abs time.Time) {
		t.Helper()
		if err := setupConn(t).QueryRow(ctx, `SELECT s.idle_expires_at, s.absolute_expires_at FROM app.sessions s
  JOIN app.mcp_connections c ON c.session_id = s.id WHERE c.id = $1::text::uuid`, id).Scan(&idle, &abs); err != nil {
			t.Fatal(err)
		}
		return idle, abs
	}
	// IT-MCP-ID-4, M5: due and live renews (capped at the absolute expiry); inside the throttle is a live no-op.
	id := f.mustCreate(t)
	f.activate(t, id, "client_01RENEW")
	i0, _ := idleOf(id)
	f.h.clk.Advance(2 * time.Hour)
	c := read(id)
	if !c.RenewalDue(f.h.clk.Now()) {
		t.Fatal("renewal not due after two hours")
	}
	if err := f.h.svc.RenewMCPConnection(ctx, c, mcpLife.Idle); err != nil {
		t.Fatalf("due renewal: %v", err)
	}
	i1, abs := idleOf(id)
	if !i1.After(i0) || i1.After(abs) {
		t.Fatalf("renewal did not extend within the absolute expiry: %v -> %v (absolute %v)", i0, i1, abs)
	}
	if err := f.h.svc.RenewMCPConnection(ctx, read(id), mcpLife.Idle); err != nil {
		t.Fatalf("throttled renewal: %v", err)
	}
	if i2, _ := idleOf(id); !i2.Equal(i1) {
		t.Fatal("a throttled renewal changed the expiry")
	}
	// Renewed by use every 29 days, the idle expiry is capped at the absolute expiry.
	for i := 0; i < 6; i++ {
		f.h.clk.Advance(29 * 24 * time.Hour)
		if err := f.h.svc.RenewMCPConnection(ctx, read(id), mcpLife.Idle); err != nil {
			t.Fatalf("renewal %d: %v", i, err)
		}
	}
	if i3, abs3 := idleOf(id); !i3.Equal(abs3) {
		t.Fatalf("renewal past the absolute expiry: idle %v absolute %v", i3, abs3)
	}
	// Terminal: a stale read cannot extend idle expiry, absolute expiry, a revocation, a wrong kind or a disabled user.
	terminal := func(name string, prep func(c identity.MCPConnection)) {
		t.Helper()
		f.h.clk.Advance(2 * time.Hour)
		cid := f.mustCreate(t)
		f.activate(t, cid, "client_01TERM")
		f.h.clk.Advance(2 * time.Hour)
		c := read(cid)
		before, _ := idleOf(cid)
		prep(c)
		if err := f.h.svc.RenewMCPConnection(ctx, c, mcpLife.Idle); !errors.Is(err, identity.ErrUnauthenticated) {
			t.Errorf("%s: renewal %v, want unauthenticated", name, err)
		}
		if after, _ := idleOf(cid); !after.Equal(before) {
			t.Errorf("%s: the expiry changed", name)
		}
	}
	terminal("idle expired", func(identity.MCPConnection) { f.h.clk.Advance(31 * 24 * time.Hour) })
	terminal("revoked", func(c identity.MCPConnection) {
		if err := f.h.svc.RevokeMCPConnection(ctx, c); err != nil {
			t.Fatal(err)
		}
	})
	terminal("browser kind", func(c identity.MCPConnection) {
		if _, err := setupConn(t).Exec(ctx, `UPDATE app.sessions SET kind = 'browser' WHERE id = $1::text::uuid`, c.SessionID()); err != nil {
			t.Fatal(err)
		}
	})
	terminal("absolute expired", func(identity.MCPConnection) { f.h.clk.Advance(181 * 24 * time.Hour) })

	// IT-MCP-ID-4: logout (the tool and the setup revoke), logout-all and disable revoke mcp sessions; re-enable revives
	// none; browser logout (current) leaves them.
	f2 := newMCPFixture(t)
	a := f2.mustCreate(t)
	f2.activate(t, a, "client_01A")
	b := f2.mustCreate(t)
	f2.activate(t, b, "client_01B")
	if _, err := f2.h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: mcpCSRF(t, f2.h, f2.cred), Origin: originA, Credential: f2.cred}); err != nil {
		t.Fatal(err)
	}
	if _, err := f2.h.svc.ReadMCPConnection(ctx, f2.user, f2.org, a); err != nil {
		t.Fatalf("browser logout ended a connection: %v", err)
	}
	if err := f2.h.svc.RevokeOwnMCPConnection(ctx, f2.user, f2.org, a); err != nil {
		t.Fatal(err)
	}
	if _, err := f2.h.svc.ReadMCPConnection(ctx, f2.user, f2.org, a); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("revoked connection: %v", err)
	}
	f2.cred = f2.h.mustSignIn(originA, grant(f2.sub, "session_01MCPID2")).Credential
	if _, err := f2.h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: mcpCSRF(t, f2.h, f2.cred), Origin: originA, Credential: f2.cred}); err != nil {
		t.Fatal(err)
	}
	if _, err := f2.h.svc.ReadMCPConnection(ctx, f2.user, f2.org, b); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("logout-all left a connection: %v", err)
	}
	f2.cred = f2.h.mustSignIn(originA, grant(f2.sub, "session_01MCPID3")).Credential
	d := f2.mustCreate(t)
	if err := f2.h.svc.DisableUser(ctx, f2.user); err != nil {
		t.Fatal(err)
	}
	if err := f2.h.svc.EnableUser(ctx, f2.user); err != nil {
		t.Fatal(err)
	}
	if _, err := f2.h.svc.ReadMCPConnection(ctx, f2.user, f2.org, d); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("re-enable revived a connection: %v", err)
	}
	// The audit-fault revocation path: an audit-only fault never prevents the revocation; the gap is recorded.
	f2.cred = f2.h.mustSignIn(originA, grant(f2.sub, "session_01MCPID4")).Credential
	e := f2.mustCreate(t)
	setup := setupConn(t)
	if _, err := setup.Exec(ctx, `ALTER TABLE audit.events ADD CONSTRAINT it_mcp_revoke_fault CHECK (action <> 'session.revoked') NOT VALID`); err != nil {
		t.Fatal(err)
	}
	ce, err := f2.h.svc.ReadMCPConnection(ctx, f2.user, f2.org, e)
	if err != nil {
		t.Fatal(err)
	}
	rerr := f2.h.svc.RevokeMCPConnection(ctx, ce)
	if _, err := setup.Exec(ctx, `ALTER TABLE audit.events DROP CONSTRAINT it_mcp_revoke_fault`); err != nil {
		t.Fatal(err)
	}
	if rerr != nil {
		t.Fatalf("revocation with an audit fault: %v", rerr)
	}
	if _, err := f2.h.svc.ReadMCPConnection(ctx, f2.user, f2.org, e); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatal("the revocation did not stand")
	}
	var n int
	if err := setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'revocation.unaudited' AND reason = 'logout'
  AND resource_id = $1::text::uuid`, ce.SessionID()).Scan(&n); err != nil || n != 1 {
		t.Fatalf("revocation.unaudited: %d %v", n, err)
	}
}

// TestMCPIdentityLogoutAllInterleaving is IT-MCP-ID-5: an M1 transaction holding the browser session FOR SHARE while
// logout-all waits on that row; the connection M1 commits survives logout-all's statement snapshot (disclosed), while
// a disable serializes on the user row and revokes it.
func TestMCPIdentityLogoutAllInterleaving(t *testing.T) {
	for _, viaDisable := range []bool{false, true} {
		f := newMCPFixture(t)
		ctx := context.Background()
		info, err := f.h.svc.Session(ctx, f.cred, originA)
		if err != nil {
			t.Fatal(err)
		}
		old := f.mustCreate(t)
		id, err := identity.NewMCPConnectionID()
		if err != nil {
			t.Fatal(err)
		}
		// M1's lock prefix (users, then the browser session, then the Organization binding), then the insert; held.
		tx := tenantTx(t, f.user, f.org)
		for _, q := range []struct {
			sql  string
			args []any
		}{{`SELECT id FROM app.users WHERE id = $1::text::uuid FOR SHARE`, []any{f.user}},
			{`SELECT id FROM app.sessions WHERE id = $1::text::uuid FOR SHARE`, []any{info.SessionID}},
			{`SELECT organization_id FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid AND scope_type = 'organization' AND source = 'membership' FOR SHARE`, []any{f.org, f.user}}} {
			rows, err := tx.Query(ctx, q.sql, q.args...)
			if err != nil {
				t.Fatal(err)
			}
			rows.Close()
			if rows.Err() != nil {
				t.Fatal(rows.Err())
			}
		}
		if err := f.h.svc.InsertMCPConnectionTx(ctx, tx, f.user, f.org, id, "org-a.localhost:543", "https://org-a.localhost:543/mcp/c/"+id, mcpLife, nil); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() {
			if viaDisable {
				done <- f.h.svc.DisableUser(ctx, f.user)
				return
			}
			_, err := f.h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: mcpCSRF(t, f.h, f.cred), Origin: originA, Credential: f.cred})
			done <- err
		}()
		select {
		case err := <-done:
			t.Fatalf("the revocation did not wait for M1 (%v)", err)
		case <-time.After(300 * time.Millisecond):
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		// The connection created before is terminal in both cases.
		if _, err := f.h.svc.ReadMCPConnection(ctx, f.user, f.org, old); !errors.Is(err, identity.ErrUnauthenticated) {
			t.Fatalf("disable=%v: the earlier connection survived", viaDisable)
		}
		_, err = f.h.svc.ReadMCPConnection(ctx, f.user, f.org, id)
		if viaDisable {
			if !errors.Is(err, identity.ErrUnauthenticated) {
				t.Fatalf("disable left the concurrent connection: %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("logout-all: the connection committed after its snapshot did not survive as disclosed: %v", err)
		}
		// It stays revocable, and its revocation is terminal.
		if err := f.h.svc.RevokeOwnMCPConnection(ctx, f.user, f.org, id); err != nil {
			t.Fatal(err)
		}
		if _, err := f.h.svc.ReadMCPConnection(ctx, f.user, f.org, id); !errors.Is(err, identity.ErrUnauthenticated) {
			t.Fatal("revocation of the surviving connection is not terminal")
		}
	}
}

func mcpCSRF(t *testing.T, h *harness, cred string) string {
	t.Helper()
	info, err := h.svc.Session(context.Background(), cred, originA)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	return csrfOf(info)
}

// TestMCPIdentityRename (G063 Plan r1 §6; r1 C5): the store and liveness layer of the rename in TENANT-ORG test
// transactions, through the exported identity API and the runtime role: the target 'mcp' session is locked by the same
// liveness statements the facade uses (step 3), then RenameMCPConnectionTx locks the row and writes the name only. Live
// pending and active connections are renamed and cleared with every other connection and session fact unchanged; a
// disabled owner (the only layer that reaches this class: the browser session dies with the owner), a G013-shaped
// expiry, a revocation, an expired setup and an idle expiry are all unauthenticated with the name unchanged; another
// Organization's shape sees no row; an unnormalized name is refused by the helper; create and list carry the name.
func TestMCPIdentityRename(t *testing.T) {
	f := newMCPFixture(t)
	ctx := context.Background()
	setup := setupConn(t)
	str := func(s string) *string { return &s }
	nameOf := func(id string) *string {
		t.Helper()
		var n *string
		if err := setup.QueryRow(ctx, `SELECT display_name FROM app.mcp_connections WHERE id = $1::text::uuid`, id).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	// facts is every connection and session column except the name.
	facts := func(id string) string {
		t.Helper()
		var s string
		if err := setup.QueryRow(ctx, `SELECT ((to_jsonb(c) - 'display_name') || jsonb_build_object('session', to_jsonb(s)))::text
  FROM app.mcp_connections c JOIN app.sessions s ON s.id = c.session_id WHERE c.id = $1::text::uuid`, id).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	sessionOf := func(id string) string {
		t.Helper()
		var s string
		if err := setup.QueryRow(ctx, `SELECT session_id::text FROM app.mcp_connections WHERE id = $1::text::uuid`, id).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	rename := func(org, id string, name *string) error {
		t.Helper()
		tx := tenantTx(t, f.user, org)
		sid, err := f.h.svc.ReadMCPConnectionSessionTx(ctx, tx, f.user, org, id)
		if err != nil {
			return err
		}
		if err := f.h.svc.LockLiveSession(ctx, tx, sid, f.user); err != nil {
			return err
		}
		if err := f.h.svc.RenameMCPConnectionTx(ctx, tx, f.user, org, id, name); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	pending := f.mustCreate(t)
	active := f.mustCreate(t)
	f.activate(t, active, "client_01NAMES")
	for label, id := range map[string]string{"pending": pending, "active": active} {
		before := facts(id)
		if err := rename(f.org, id, str("Laptop")); err != nil {
			t.Fatalf("%s rename: %v", label, err)
		}
		if n := nameOf(id); n == nil || *n != "Laptop" {
			t.Fatalf("%s: name %v", label, n)
		}
		if err := rename(f.org, id, str("Laptop")); err != nil { // same value: idempotent
			t.Fatalf("%s same-value rename: %v", label, err)
		}
		if err := rename(f.org, id, nil); err != nil {
			t.Fatalf("%s clear: %v", label, err)
		}
		if n := nameOf(id); n != nil {
			t.Fatalf("%s: name after clear %q", label, *n)
		}
		if after := facts(id); after != before {
			t.Errorf("%s: a rename changed a connection or session fact:\n before %s\n after  %s", label, before, after)
		}
	}
	// The helper refuses a value that is not already normalized (the facade normalizes first); nothing is written.
	for _, bad := range []string{" padded", "a\nb", ""} {
		tx := tenantTx(t, f.user, f.org)
		if err := f.h.svc.RenameMCPConnectionTx(ctx, tx, f.user, f.org, active, str(bad)); !errors.Is(err, identity.ErrInvalidMCPName) {
			t.Errorf("unnormalized %q: %v, want ErrInvalidMCPName", bad, err)
		}
	}
	// Another Organization's shape sees no row (and no session to lock).
	tx := tenantTx(t, f.user, f.other)
	if _, err := f.h.svc.ReadMCPConnectionSessionTx(ctx, tx, f.user, f.other, active); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Errorf("session read from another Organization: %v", err)
	}
	if err := f.h.svc.RenameMCPConnectionTx(ctx, tx, f.user, f.other, active, str("x")); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Errorf("rename from another Organization: %v", err)
	}
	// Create with a name and list it back (a legacy unnamed row lists nil).
	named, err := identity.NewMCPConnectionID()
	if err != nil {
		t.Fatal(err)
	}
	tx = tenantTx(t, f.user, f.org)
	if err := f.h.svc.InsertMCPConnectionTx(ctx, tx, f.user, f.org, named, "org-a.localhost:543", "https://org-a.localhost:543/mcp/c/"+named,
		mcpLife, str("Desk")); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	tx = tenantTx(t, f.user, f.org)
	if err := f.h.svc.InsertMCPConnectionTx(ctx, tx, f.user, f.org, named, "h", "r", mcpLife, str(" Desk")); !errors.Is(err, identity.ErrInvalidMCPName) {
		t.Errorf("insert with an unnormalized name: %v", err)
	}
	tx = tenantTx(t, f.user, f.org)
	page, err := f.h.svc.ListMCPConnectionsTx(ctx, tx, f.user, f.org, nil)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]*string{}
	for _, c := range page.Connections {
		seen[c.ID] = c.DisplayName
	}
	if n, ok := seen[named]; !ok || n == nil || *n != "Desk" {
		t.Errorf("listed name of the named connection: %v", n)
	}
	if n, ok := seen[active]; !ok || n != nil {
		t.Errorf("listed name of the unnamed connection: %v %v", ok, n)
	}

	// Terminal classes: unauthenticated and the name unchanged.
	terminal := func(label string, pend bool, prep func(id string), undo func()) {
		t.Helper()
		id := f.mustCreate(t)
		if !pend {
			f.activate(t, id, "client_01TERM")
		}
		prep(id)
		if err := rename(f.org, id, str("After")); !errors.Is(err, identity.ErrUnauthenticated) {
			t.Errorf("%s: rename %v, want unauthenticated", label, err)
		}
		if n := nameOf(id); n != nil {
			t.Errorf("%s: the name changed to %q", label, *n)
		}
		if undo != nil {
			undo()
		}
	}
	terminal("disabled owner", false, func(string) {
		if _, err := setup.Exec(ctx, `UPDATE app.users SET status = 'disabled', disabled_at = now() WHERE id = $1::text::uuid`, f.user); err != nil {
			t.Fatal(err)
		}
	}, func() {
		if _, err := setup.Exec(ctx, `UPDATE app.users SET status = 'active', disabled_at = NULL WHERE id = $1::text::uuid`, f.user); err != nil {
			t.Fatal(err)
		}
	})
	terminal("G013-shaped expiry", false, func(id string) {
		f.h.clk.Advance(time.Minute)
		if _, err := setup.Exec(ctx, `UPDATE app.sessions SET idle_expires_at = least(idle_expires_at, $2), absolute_expires_at = least(absolute_expires_at, $2)
  WHERE id = $1::text::uuid AND revoked_at IS NULL`, sessionOf(id), f.h.clk.Now()); err != nil {
			t.Fatal(err)
		}
	}, nil)
	terminal("revoked", false, func(id string) {
		if err := f.h.svc.RevokeOwnMCPConnection(ctx, f.user, f.org, id); err != nil {
			t.Fatal(err)
		}
	}, nil)
	terminal("expired setup", true, func(string) { f.h.clk.Advance(31 * time.Minute) }, nil)
	terminal("idle expired", false, func(string) { f.h.clk.Advance(31 * 24 * time.Hour) }, nil)
}
