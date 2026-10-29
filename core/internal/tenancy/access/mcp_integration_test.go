//go:build integration

// G014 B-MCP-IT, facade part (Plan r8 §4.4, §10.3, §11.6): the selection ordering cases O-MCP-1..4 with the inert hook
// (IT-MCP-7; O-MCP-3 is IT-MCP-12(d)), the activation race (one winner), and the bounded PG-006 observations E-MCP-1..3
// with constant statement counts in the Group count. G015 adds O-MCP-7/8 (the eligibility snapshot under a held switch
// and under a change committed after the catalog is computed) and E-MCP-4 (eligibility statements per catalog shape).
// Same environment as B-POLICY-IT (setupIT): the real runtime-role pool through the counting hookDB and the real G004
// identity service.
package access

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
	"core/internal/observability"
	"core/internal/tenancy"
)

var mcpTestLifetimes = identity.MCPLifetimes{Pending: 30 * time.Minute, Idle: 24 * time.Hour, Absolute: 48 * time.Hour}

// mcpConn creates a pending connection of u in org (setup through the facade, in a request scope of u's credential for
// org: G016 A2) and returns its identity binding.
func (e *itEnv) mcpConn(t *testing.T, u *itUser, org tenancy.OrganizationID) identity.MCPConnection {
	t.Helper()
	ctx := context.Background()
	rctx, end := beginAs(t, e.svc, ctx, u, org)
	defer end()
	tc := resolve(t, e.svc, rctx, org, tenancy.GroupID{})
	id, err := identity.NewMCPConnectionID()
	if err != nil {
		t.Fatal(err)
	}
	host := "org-a.localhost:543"
	if err := e.svc.CreateMCPConnection(rctx, tc, id, host, "https://"+host+"/mcp/c/"+id, mcpTestLifetimes, nil); err != nil {
		t.Fatalf("create: %v", err)
	}
	c, err := e.id.ReadMCPConnection(ctx, u.id.String(), org.String(), id)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return c
}

// verified begins a verified scope for c (re-read, so the binding is current) and returns it with its end function.
func (e *itEnv) verified(t *testing.T, u *itUser, org tenancy.OrganizationID, c identity.MCPConnection) (context.Context, func()) {
	t.Helper()
	cur, err := e.id.ReadMCPConnection(context.Background(), u.id.String(), org.String(), c.ID())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	vctx, end, err := e.svc.BeginVerifiedRequest(context.Background(), cur)
	if err != nil {
		t.Fatalf("begin verified: %v", err)
	}
	return vctx, end
}

func (e *itEnv) activeConn(t *testing.T, u *itUser, org tenancy.OrganizationID) identity.MCPConnection {
	t.Helper()
	c := e.mcpConn(t, u, org)
	vctx, end := e.verified(t, u, org, c)
	defer end()
	if err := e.svc.ActivateConnection(vctx, "client_01ORDERING"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	return c
}

func (e *itEnv) selection(t *testing.T, u *itUser, org tenancy.OrganizationID, c identity.MCPConnection) (string, int64) {
	t.Helper()
	vctx, end := e.verified(t, u, org, c)
	defer end()
	cc, err := e.svc.GetConnectionContext(vctx)
	if err != nil {
		t.Fatalf("context: %v", err)
	}
	g := ""
	if cc.Group != nil {
		g = cc.Group.ID.String()
	}
	return g, cc.ContextVersion
}

func TestMCPOrdering(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	owner, member := e.newUser(t), e.newUser(t)
	org := e.org(t, "MCP Ordering "+e.canaries[0])
	e.grantOrg(t, org, owner, tenancy.OrgOwner)
	e.grantOrg(t, org, member, tenancy.OrgMember)
	g1 := e.group(t, org, owner, "Ordering One", fmt.Sprintf("ord-one-%d", itSeq.Add(1)))
	g2 := e.group(t, org, owner, "Ordering Two", fmt.Sprintf("ord-two-%d", itSeq.Add(1)))
	e.grantGroup(t, org, g1.ID, member, tenancy.GroupMember)
	e.grantGroup(t, org, g2.ID, member, tenancy.GroupMember)
	e.bind(t, owner, org) // G016 A2: each user's credential for org (setup only)
	e.bind(t, member, org)

	// hold installs a selectLocked hook that signals entry and waits for release.
	hold := func() (entered, release chan struct{}) {
		entered, release = make(chan struct{}), make(chan struct{})
		var once sync.Once
		e.svc.mcpHooks = &mcpHooks{selectLocked: func() {
			once.Do(func() { close(entered) })
			<-release
		}}
		return entered, release
	}
	defer func() { e.svc.mcpHooks = nil }()

	t.Run("activation race: one winner, same client only", func(t *testing.T) {
		c := e.mcpConn(t, member, org)
		var wg sync.WaitGroup
		errs := make([]error, 4)
		for i := range errs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				vctx, end, err := e.svc.BeginVerifiedRequest(ctx, c)
				if err != nil {
					errs[i] = err
					return
				}
				defer end()
				client := "client_01RACEA"
				if i%2 == 1 {
					client = "client_01RACEB"
				}
				errs[i] = e.svc.ActivateConnection(vctx, client)
			}(i)
		}
		wg.Wait()
		ok := 0
		for _, err := range errs {
			if err == nil {
				ok++
			} else if CodeOf(err) != CodeUnauthenticated {
				t.Errorf("activation error: %v", err)
			}
		}
		if ok != 2 { // the winner and the other caller with the same client
			t.Fatalf("%d activations succeeded, want 2 (one client)", ok)
		}
		var n int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'mcp.connection_established' AND resource_id = $1::text::uuid`,
			c.ID()).Scan(&n); err != nil || n != 1 {
			t.Fatalf("established events: %d %v", n, err)
		}
	})

	t.Run("O-MCP-1 a held switch does not block a read, which sees the old pair, then the new", func(t *testing.T) {
		c := e.activeConn(t, member, org)
		entered, release := hold()
		done := make(chan error, 1)
		go func() {
			vctx, end := e.verified(t, member, org, c)
			defer end()
			_, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String())
			done <- err
		}()
		<-entered
		if g, v := e.selection(t, member, org, c); g != "" || v != 0 {
			t.Fatalf("read during the held switch: %q %d", g, v)
		}
		close(release)
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		if g, v := e.selection(t, member, org, c); g != g1.ID.String() || v != 1 {
			t.Fatalf("after the switch: %q %d", g, v)
		}
	})

	t.Run("O-MCP-2 concurrent switches serialize; each validates current facts", func(t *testing.T) {
		c := e.activeConn(t, member, org)
		entered, release := hold()
		done := make(chan error, 1)
		go func() {
			vctx, end := e.verified(t, member, org, c)
			defer end()
			_, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String())
			done <- err
		}()
		<-entered
		e.svc.mcpHooks = nil // the second switch is not held
		vctx, end := e.verified(t, member, org, c)
		if _, err := e.svc.SelectConnectionGroup(vctx, g2.ID.String()); err != nil {
			t.Fatalf("second switch: %v", err)
		}
		end()
		close(release)
		// The held switch read the old selection before the second committed: it writes nothing (a conflict, which the
		// tool reports as unavailable) and the client reconciles.
		if err := <-done; CodeOf(err) != CodeConflict {
			t.Fatalf("held switch after a concurrent commit: %v", err)
		}
		if g, v := e.selection(t, member, org, c); g != g2.ID.String() || v != 1 {
			t.Fatalf("after both: %q %d", g, v)
		}
		// A foreign target changes nothing.
		foreign := e.org(t, "MCP Foreign "+e.canaries[1])
		fg := e.group(t, foreign, owner, "Foreign", fmt.Sprintf("foreign-%d", itSeq.Add(1)))
		vctx, end = e.verified(t, member, org, c)
		defer end()
		if _, err := e.svc.SelectConnectionGroup(vctx, fg.ID.String()); CodeOf(err) != CodeGroupNotSelectable {
			t.Fatalf("foreign target: %v", err)
		}
		if g, v := e.selection(t, member, org, c); g != g2.ID.String() || v != 1 {
			t.Fatalf("foreign target changed the selection: %q %d", g, v)
		}
	})

	t.Run("O-MCP-2b a selection committed between the decision snapshot and the write makes the switch write nothing", func(t *testing.T) {
		c := e.activeConn(t, member, org)
		entered, release := make(chan struct{}), make(chan struct{})
		var once sync.Once
		e.svc.mcpHooks = &mcpHooks{selectDecided: func() {
			once.Do(func() { close(entered) })
			<-release
		}}
		done := make(chan error, 1)
		go func() {
			vctx, end := e.verified(t, member, org, c)
			defer end()
			_, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String())
			done <- err
		}()
		<-entered
		e.svc.mcpHooks = nil
		vctx, end := e.verified(t, member, org, c)
		if _, err := e.svc.SelectConnectionGroup(vctx, g2.ID.String()); err != nil {
			t.Fatalf("intervening switch: %v", err)
		}
		end()
		close(release)
		// The decided switch validated an empty selection; g2 is selected now, so it writes nothing.
		if err := <-done; CodeOf(err) != CodeConflict {
			t.Fatalf("switch over an unvalidated selection: %v", err)
		}
		if g, v := e.selection(t, member, org, c); g != g2.ID.String() || v != 1 {
			t.Fatalf("after both: %q %d", g, v)
		}
	})

	t.Run("O-MCP-3 (IT-MCP-12 d) a membership removal waits for the held switch, then the next call is denied", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, org, u, tenancy.OrgMember)
		e.grantGroup(t, org, g1.ID, u, tenancy.GroupMember)
		e.bind(t, u, org) // G016 A2: before the connection and any held switch
		c := e.activeConn(t, u, org)
		entered, release := hold()
		done := make(chan error, 1)
		go func() {
			vctx, end := e.verified(t, u, org, c)
			defer end()
			_, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String())
			done <- err
		}()
		<-entered
		removed := make(chan error, 1)
		go func() { removed <- e.model.RemoveOrganizationMembership(ctx, org, u.id) }()
		select {
		case err := <-removed:
			t.Fatalf("the removal did not wait for the held switch: %v", err)
		case <-time.After(300 * time.Millisecond):
		}
		// E-MCP-3: the locks the held switch holds (pg_locks of its backend, by relation).
		mcpLocks(t, e)
		close(release)
		if err := <-done; err != nil {
			t.Fatalf("held switch: %v", err)
		}
		if err := <-removed; err != nil {
			t.Fatal(err)
		}
		e.svc.mcpHooks = nil
		vctx, end := e.verified(t, u, org, c)
		defer end()
		if _, err := e.svc.GetConnectionContext(vctx); CodeOf(err) != CodeContextDenied {
			t.Fatalf("call after the removal: %v", err)
		}
	})

	t.Run("O-MCP-5 a removal from the selected Group committed after the decision refuses a switch away; a clear still succeeds", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, org, u, tenancy.OrgMember)
		e.grantGroup(t, org, g1.ID, u, tenancy.GroupMember)
		e.grantGroup(t, org, g2.ID, u, tenancy.GroupMember)
		e.bind(t, u, org) // G016 A2: before the connection and any held switch
		c := e.activeConn(t, u, org)
		vctx, end := e.verified(t, u, org, c)
		if _, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String()); err != nil {
			t.Fatalf("select g1: %v", err)
		}
		end()
		entered, release := make(chan struct{}), make(chan struct{})
		var once sync.Once
		e.svc.mcpHooks = &mcpHooks{selectDecided: func() {
			once.Do(func() { close(entered) })
			<-release
		}}
		done := make(chan error, 1)
		go func() {
			vctx, end := e.verified(t, u, org, c)
			defer end()
			_, err := e.svc.SelectConnectionGroup(vctx, g2.ID.String())
			done <- err
		}()
		<-entered
		e.svc.mcpHooks = nil
		// The decision validated g1; the removal from g1 commits before the guarded write takes its locks.
		if err := e.model.RemoveGroupMembership(ctx, org, g1.ID, u.id); err != nil {
			t.Fatal(err)
		}
		denials := func() (int, string) {
			var n int
			var reasons string
			if err := e.setup.QueryRow(ctx, `SELECT count(*), coalesce(string_agg(reason, ','), '') FROM audit.events
  WHERE action = 'access.denied' AND operation = 'mcp.request' AND actor_user_id = $1::text::uuid`, u.id.String()).Scan(&n, &reasons); err != nil {
				t.Fatal(err)
			}
			return n, reasons
		}
		before, _ := denials()
		close(release)
		if err := <-done; CodeOf(err) != CodeContextDenied {
			t.Fatalf("switch away from a lost Group: %v", err)
		}
		// S1: the refusal decided on the locked facts is a durable access.denied record (the Group no longer selectable).
		if after, reasons := denials(); after != before+1 || reasons != "group_not_selectable" {
			t.Fatalf("access.denied records for the refused switch: %d -> %d (%s)", before, after, reasons)
		}
		// The context is now invalid for reads (the selected Group is lost), so the stored row is read directly.
		var g string
		var v int64
		if err := e.setup.QueryRow(ctx, `SELECT coalesce(selected_group_id::text, ''), context_version FROM app.mcp_connections WHERE id = $1::text::uuid`,
			c.ID()).Scan(&g, &v); err != nil || g != g1.ID.String() || v != 1 {
			t.Fatalf("the refused switch changed the selection: %q %d %v", g, v, err)
		}
		// Spec R5: an explicit clear still succeeds after the loss.
		vctx, end = e.verified(t, u, org, c)
		defer end()
		if _, err := e.svc.SelectConnectionGroup(vctx, ""); err != nil {
			t.Fatalf("clear after the loss: %v", err)
		}
		if g, v := e.selection(t, u, org, c); g != "" || v != 2 {
			t.Fatalf("after the clear: %q %d", g, v)
		}
	})

	t.Run("O-MCP-6 a removal from the selected Group waits for a held switch away from it", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, org, u, tenancy.OrgMember)
		e.grantGroup(t, org, g1.ID, u, tenancy.GroupMember)
		e.grantGroup(t, org, g2.ID, u, tenancy.GroupMember)
		e.bind(t, u, org) // G016 A2: before the connection and any held switch
		c := e.activeConn(t, u, org)
		vctx, end := e.verified(t, u, org, c)
		if _, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String()); err != nil {
			t.Fatalf("select g1: %v", err)
		}
		end()
		entered, release := hold()
		done := make(chan error, 1)
		go func() {
			vctx, end := e.verified(t, u, org, c)
			defer end()
			_, err := e.svc.SelectConnectionGroup(vctx, g2.ID.String())
			done <- err
		}()
		<-entered
		removed := make(chan error, 1)
		go func() { removed <- e.model.RemoveGroupMembership(ctx, org, g1.ID, u.id) }()
		select {
		case err := <-removed:
			t.Fatalf("the removal of the previous Group did not wait for the held switch: %v", err)
		case <-time.After(300 * time.Millisecond):
		}
		close(release)
		if err := <-done; err != nil {
			t.Fatalf("held switch: %v", err)
		}
		if err := <-removed; err != nil {
			t.Fatal(err)
		}
		e.svc.mcpHooks = nil
		if g, v := e.selection(t, u, org, c); g != g2.ID.String() || v != 2 {
			t.Fatalf("after the switch and the removal: %q %d", g, v)
		}
	})

	// G015 (Plan r4 §5.3 C-4/C-5): the eligibility snapshot is one (selection, facts) pair and authorizes nothing.
	updateReq, err := RequireCapability("group.update")
	if err != nil {
		t.Fatal(err)
	}
	catalogReqs := map[string]ToolRequirement{"test.group_update": updateReq}
	catalog := func(t *testing.T, u *itUser, c identity.MCPConnection) ([]string, error) {
		t.Helper()
		vctx, end := e.verified(t, u, org, c)
		defer end()
		return e.svc.AuthorizeCatalog(vctx, catalogReqs)
	}

	t.Run("O-MCP-7 a catalog computed while a switch is held is the old pair; one spanning its commit stays old", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, org, u, tenancy.OrgMember)
		e.grantGroup(t, org, g1.ID, u, tenancy.GroupAdmin)
		e.grantGroup(t, org, g2.ID, u, tenancy.GroupMember)
		e.bind(t, u, org) // G016 A2: before the connection and any held switch
		c := e.activeConn(t, u, org)
		vctx, end := e.verified(t, u, org, c)
		if _, err := e.svc.SelectConnectionGroup(vctx, g2.ID.String()); err != nil {
			t.Fatalf("select g2: %v", err)
		}
		end()
		if names, err := catalog(t, u, c); err != nil || len(names) != 0 {
			t.Fatalf("catalog in g2: %v %v", names, err)
		}
		// A switch to g1 is held after its locks; a catalog does not wait for it and sees the old selection (g2).
		switchEntered, switchRelease := make(chan struct{}), make(chan struct{})
		catalogEntered, catalogRelease := make(chan struct{}), make(chan struct{})
		var switchOnce, catalogOnce sync.Once
		holdCatalog := false
		var mu sync.Mutex
		e.svc.mcpHooks = &mcpHooks{
			selectLocked: func() {
				switchOnce.Do(func() { close(switchEntered) })
				<-switchRelease
			},
			catalogComputed: func() {
				mu.Lock()
				hold := holdCatalog
				mu.Unlock()
				if hold {
					catalogOnce.Do(func() { close(catalogEntered) })
					<-catalogRelease
				}
			},
		}
		done := make(chan error, 1)
		go func() {
			vctx, end := e.verified(t, u, org, c)
			defer end()
			_, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String())
			done <- err
		}()
		<-switchEntered
		if names, err := catalog(t, u, c); err != nil || len(names) != 0 {
			t.Fatalf("catalog during the held switch: %v %v", names, err)
		}
		if g, v := e.selection(t, u, org, c); g != g2.ID.String() || v != 1 {
			t.Fatalf("selection during the held switch: %q %d", g, v)
		}
		// A catalog whose snapshot opened before the commit returns the old pair even after the switch commits.
		mu.Lock()
		holdCatalog = true
		mu.Unlock()
		type out struct {
			names []string
			err   error
		}
		spanning := make(chan out, 1)
		go func() {
			vctx, end := e.verified(t, u, org, c)
			defer end()
			names, err := e.svc.AuthorizeCatalog(vctx, catalogReqs)
			spanning <- out{names, err}
		}()
		<-catalogEntered
		close(switchRelease)
		if err := <-done; err != nil {
			t.Fatalf("held switch: %v", err)
		}
		close(catalogRelease)
		if r := <-spanning; r.err != nil || len(r.names) != 0 {
			t.Fatalf("spanning catalog: %v %v", r.names, r.err)
		}
		e.svc.mcpHooks = nil
		// The next snapshot is the new pair.
		if names, err := catalog(t, u, c); err != nil || !slices.Equal(names, []string{"test.group_update"}) {
			t.Fatalf("catalog after the switch: %v %v", names, err)
		}
		if g, v := e.selection(t, u, org, c); g != g1.ID.String() || v != 2 {
			t.Fatalf("selection after the switch: %q %d", g, v)
		}
	})

	t.Run("O-MCP-8 a change committed after the catalog is computed governs the next request and the tools' own checks", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, org, u, tenancy.OrgMember)
		e.grantGroup(t, org, g1.ID, u, tenancy.GroupAdmin)
		e.grantGroup(t, org, g2.ID, u, tenancy.GroupAdmin)
		e.bind(t, u, org) // G016 A2: before the connection and any held switch
		c := e.activeConn(t, u, org)
		vctx, end := e.verified(t, u, org, c)
		if _, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String()); err != nil {
			t.Fatalf("select g1: %v", err)
		}
		end()
		// (a) A downgrade committed inside the snapshot window: this request's names are its own snapshot's; the next
		// request no longer finds the capability eligible.
		var once sync.Once
		e.svc.mcpHooks = &mcpHooks{catalogComputed: func() {
			once.Do(func() {
				if _, err := e.model.ChangeGroupRole(ctx, org, g1.ID, u.id, tenancy.GroupMember); err != nil {
					t.Errorf("downgrade: %v", err)
				}
			})
		}}
		if names, err := catalog(t, u, c); err != nil || !slices.Equal(names, []string{"test.group_update"}) {
			t.Fatalf("catalog computed before the downgrade: %v %v", names, err)
		}
		e.svc.mcpHooks = nil
		if names, err := catalog(t, u, c); err != nil || len(names) != 0 {
			t.Fatalf("next catalog after the downgrade: %v %v", names, err)
		}
		// (b) The selected Group lost inside the snapshot window: the next catalog is refused as a whole, the
		// production tools refuse on their own checks, never on the earlier result; the clear still succeeds.
		var once2 sync.Once
		e.svc.mcpHooks = &mcpHooks{catalogComputed: func() {
			once2.Do(func() {
				if err := e.model.RemoveGroupMembership(ctx, org, g1.ID, u.id); err != nil {
					t.Errorf("removal: %v", err)
				}
			})
		}}
		if _, err := catalog(t, u, c); err != nil {
			t.Fatalf("catalog computed before the removal: %v", err)
		}
		e.svc.mcpHooks = nil
		if _, err := catalog(t, u, c); CodeOf(err) != CodeContextDenied {
			t.Fatalf("next catalog after the removal: %v", err)
		}
		vctx, end = e.verified(t, u, org, c)
		defer end()
		if _, err := e.svc.GetConnectionContext(vctx); CodeOf(err) != CodeContextDenied {
			t.Fatalf("read tool after the removal: %v", err)
		}
		if _, err := e.svc.SelectConnectionGroup(vctx, g2.ID.String()); CodeOf(err) != CodeContextDenied {
			t.Fatalf("switch away after the removal: %v", err)
		}
		if _, err := e.svc.SelectConnectionGroup(vctx, ""); err != nil {
			t.Fatalf("clear after the removal: %v", err)
		}
	})

	t.Run("O-MCP-4 a logout committed before the switch makes it unauthenticated", func(t *testing.T) {
		c := e.activeConn(t, member, org)
		vctx, end := e.verified(t, member, org, c)
		defer end()
		if err := e.id.RevokeMCPConnection(ctx, c); err != nil {
			t.Fatal(err)
		}
		if _, err := e.svc.SelectConnectionGroup(vctx, g1.ID.String()); CodeOf(err) != CodeUnauthenticated {
			t.Fatalf("switch after logout: %v", err)
		}
		var g *string
		if err := e.setup.QueryRow(ctx, `SELECT selected_group_id::text FROM app.mcp_connections WHERE id = $1::text::uuid`, c.ID()).Scan(&g); err != nil || g != nil {
			t.Fatalf("selection after a refused switch: %v %v", g, err)
		}
	})

	t.Run("O-MCP-9 a rename and a revocation serialize on the target session in both orders, without a deadlock", func(t *testing.T) {
		rec := &recordTracer{pauseTracer: &pauseTracer{}}
		tsvc := NewService(e.runtimePool(t, rec), e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
		name := "Ordering laptop"
		state := func(c identity.MCPConnection) (named *string, revoked bool, liveness string) {
			t.Helper()
			if err := e.setup.QueryRow(ctx, `SELECT c.display_name, s.revoked_at IS NOT NULL,
  s.idle_expires_at::text || '|' || s.absolute_expires_at::text || '|' || s.renewed_at::text
  FROM app.mcp_connections c JOIN app.sessions s ON s.id = c.session_id WHERE c.id = $1::text::uuid`, c.ID()).Scan(&named, &revoked, &liveness); err != nil {
				t.Fatal(err)
			}
			return named, revoked, liveness
		}
		renameCall := func(svc *Service, c identity.MCPConnection) func() error {
			rctx, _ := beginAs(t, svc, ctx, member, org)
			tc := resolve(t, svc, rctx, org, tenancy.GroupID{})
			return func() error { return svc.RenameMCPConnection(rctx, tc, c.ID(), &name) }
		}

		// (a) rename first: paused after its connection-row lock (the target session FOR SHARE is held), a revocation's
		// session UPDATE waits; released, the rename commits, then the revocation. The statement order is pinned.
		c := e.activeConn(t, member, org)
		_, _, live0 := state(c)
		call := renameCall(tsvc, c)
		rec.take()
		p := rec.arm("SELECT state, pending_expires_at FROM app.mcp_connections")
		ren := start(call)
		p.wait(t, "rename row lock")
		rev := start(func() error { return e.id.RevokeOwnMCPConnection(ctx, member.id.String(), org.String(), c.ID()) })
		e.waitWaiters(t, 1)
		if rev.finished() {
			t.Fatal("(a) the revocation did not wait for the rename's target-session lock")
		}
		close(p.release)
		rerr, verr := ren.wait(t, "rename"), rev.wait(t, "revocation")
		if rerr != nil || verr != nil || isDeadlock(rerr) || isDeadlock(verr) {
			t.Fatalf("(a) rename %v, revocation %v", rerr, verr)
		}
		if ren.at > rev.at {
			t.Errorf("(a) the revocation finished before the rename it waited for")
		}
		named, revoked, live := state(c)
		if named == nil || *named != name || !revoked || live != live0 {
			t.Errorf("(a) final state: name %v, revoked %v, liveness %s (before %s)", named, revoked, live, live0)
		}
		var order []string
		for _, sql := range rec.take() {
			k := renameStatementKind(sql)
			if k != "" && (len(order) == 0 || order[len(order)-1] != k || k != "facts") {
				order = append(order, k)
			}
		}
		ordering(t, map[string]any{"case": "O-MCP-9", "order": strings.Join(order, ",")})
		if got := strings.Join(order, ","); got != "user,session,target,user,session,facts,row,write" {
			t.Errorf("(a) rename statement order %s, want user,session,target,user,session,facts,row,write", got)
		}

		// (b) revocation first: committed, then the rename is not_found and writes nothing.
		c = e.activeConn(t, member, org)
		call = renameCall(e.svc, c)
		if err := e.id.RevokeOwnMCPConnection(ctx, member.id.String(), org.String(), c.ID()); err != nil {
			t.Fatal(err)
		}
		_, _, live0 = state(c)
		if err := call(); CodeOf(err) != CodeNotFound {
			t.Errorf("(b) rename after the revocation: %v, want not_found", err)
		}
		if named, _, live := state(c); named != nil || live != live0 {
			t.Errorf("(b) the refused rename changed the row: name %v, liveness %s (before %s)", named, live, live0)
		}

		// (c) revocation held open: the rename waits at the target session FOR SHARE, not at the row, with no deadlock;
		// after the revocation commits it observes it and is not_found.
		c = e.activeConn(t, member, org)
		call = renameCall(e.svc, c)
		y, err := e.setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = y.Rollback(ctx) }()
		if _, err := y.Exec(ctx, `UPDATE app.sessions SET revoked_at = now(), revoke_reason = 'logout' WHERE id = $1::text::uuid`, c.SessionID()); err != nil {
			t.Fatal(err)
		}
		ren = start(call)
		e.waitWaiters(t, 1)
		if ren.finished() {
			t.Fatal("(c) the rename did not wait for the held revocation")
		}
		var rowLocked bool
		if err := e.setup.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_locks l JOIN pg_stat_activity a ON a.pid = l.pid
  WHERE NOT l.granted AND a.datname = current_database() AND a.query LIKE 'SELECT state, pending_expires_at FROM app.mcp_connections%')`).Scan(&rowLocked); err != nil {
			t.Fatal(err)
		}
		if rowLocked {
			t.Error("(c) the rename waits at the connection row, not at the target session")
		}
		if err := y.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if err := ren.wait(t, "rename"); CodeOf(err) != CodeNotFound || isDeadlock(err) {
			t.Errorf("(c) rename after the held revocation committed: %v, want not_found", err)
		}
		if named, revoked, _ := state(c); named != nil || !revoked {
			t.Errorf("(c) final state: name %v, revoked %v", named, revoked)
		}
	})
}

// mcpLocks prints the relation-level locks held by sessions other than this test's (descriptive E-MCP-3) and fails on a
// lock of a relation the switch must not lock.
func mcpLocks(t *testing.T, e *itEnv) {
	t.Helper()
	var s string
	if err := e.setup.QueryRow(context.Background(), `SELECT coalesce(string_agg(DISTINCT c.relname || ':' || l.mode, ',' ORDER BY c.relname || ':' || l.mode), '')
FROM pg_locks l JOIN pg_class c ON c.oid = l.relation
WHERE l.pid <> pg_backend_pid() AND c.relnamespace = 'app'::regnamespace AND c.relkind = 'r' AND l.granted`).Scan(&s); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(map[string]any{"evidence": "E-MCP-3", "locks": s})
	fmt.Println("E-MCP-3 " + string(b))
	for _, l := range strings.Split(s, ",") {
		if strings.HasPrefix(l, "stored_objects:") || strings.HasPrefix(l, "business_profiles:") {
			fmt.Println("E-MCP-TRIGGER " + l)
			t.Errorf("unexpected lock %s", l)
		}
	}
}

// TestMCPFootprint is E-MCP-1/2 (EXPLAIN of both listing forms and of the connection read) and the constant statement
// count in the Group count (Plan r8 §11.6).
func TestMCPFootprint(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	owner, member := e.newUser(t), e.newUser(t)
	org := e.org(t, "MCP Footprint "+e.canaries[0])
	e.grantOrg(t, org, owner, tenancy.OrgOwner)
	e.grantOrg(t, org, member, tenancy.OrgMember)
	g := e.group(t, org, owner, "Footprint", fmt.Sprintf("fp-%d", itSeq.Add(1)))
	e.grantGroup(t, org, g.ID, member, tenancy.GroupMember)
	e.bind(t, owner, org) // G016 A2: each user's credential for org, before any measured window
	e.bind(t, member, org)
	counts := func(u *itUser, c identity.MCPConnection) int {
		vctx, end := e.verified(t, u, org, c)
		defer end()
		e.hook.reset()
		if _, err := e.svc.ListSelectableGroups(vctx, "", 100); err != nil {
			t.Fatal(err)
		}
		_, stmts := e.hook.counts()
		return len(stmts)
	}
	oc, mc := e.activeConn(t, owner, org), e.activeConn(t, member, org)
	// G015 E-MCP-4 (Plan r4 §5.5): the eligibility snapshot's statements per catalog shape. The production shape (no
	// capability-mapped tool: tools/list, and the C-6 error path of a tools/call naming anything else) must equal the
	// G014 context check; the capability shape (Group, profile and Organization kinds) adds at most four statements;
	// all stay constant in the Group and profile count.
	reqs := map[string]ToolRequirement{}
	for name, c := range map[string]string{"test.group_update": "group.update", "test.profile_update": "business_profile.profile.update",
		"test.audit_read": "audit.read"} {
		r, err := RequireCapability(c)
		if err != nil {
			t.Fatal(err)
		}
		reqs[name] = r
	}
	stmts := func(u *itUser, c identity.MCPConnection, f func(context.Context) error) int {
		vctx, end := e.verified(t, u, org, c)
		defer end()
		e.hook.reset()
		if err := f(vctx); err != nil {
			t.Fatal(err)
		}
		_, s := e.hook.counts()
		return len(s)
	}
	connection := func(ctx context.Context) error { return e.svc.AuthorizeConnection(ctx) }
	production := func(ctx context.Context) error {
		_, err := e.svc.AuthorizeCatalog(ctx, map[string]ToolRequirement{})
		return err
	}
	capability := func(ctx context.Context) error {
		_, err := e.svc.AuthorizeCatalog(ctx, reqs)
		return err
	}
	msel := e.activeConn(t, member, org)
	vctx, end := e.verified(t, member, org, msel)
	if _, err := e.svc.SelectConnectionGroup(vctx, g.ID.String()); err != nil {
		t.Fatal(err)
	}
	end()
	emcp4 := func() map[string]int {
		return map[string]int{
			"owner_context": stmts(owner, oc, connection), "owner_production": stmts(owner, oc, production),
			"owner_capability": stmts(owner, oc, capability), "member_context": stmts(member, mc, connection),
			"member_production": stmts(member, mc, production), "member_capability": stmts(member, mc, capability),
			"member_selected_context": stmts(member, msel, connection), "member_selected_production": stmts(member, msel, production),
			"member_selected_capability": stmts(member, msel, capability),
		}
	}
	before := []int{counts(owner, oc), counts(member, mc)}
	e4before := emcp4()
	for i := 0; i < 30; i++ {
		gx := e.group(t, org, owner, fmt.Sprintf("Extra %d", i), fmt.Sprintf("fp-extra-%d-%d", i, itSeq.Add(1)))
		e.grantGroup(t, org, gx.ID, member, tenancy.GroupMember)
		e.link(t, org, gx.ID, e.profile(t, org, fmt.Sprintf("Extra Profile %d", i)).ID)
	}
	if _, err := e.setup.Exec(ctx, `ANALYZE app.groups; ANALYZE app.role_bindings; ANALYZE app.mcp_connections`); err != nil {
		t.Fatal(err)
	}
	after := []int{counts(owner, oc), counts(member, mc)}
	if !slices.Equal(before, after) {
		fmt.Println("E-MCP-TRIGGER statement fan-out")
		t.Fatalf("statement counts grew with the Group count: %v -> %v", before, after)
	}
	e4after := emcp4()
	if !maps.Equal(e4before, e4after) {
		fmt.Println("E-MCP-TRIGGER eligibility statement fan-out")
		t.Fatalf("eligibility statement counts grew with the Group and profile count: %v -> %v", e4before, e4after)
	}
	for _, who := range []string{"owner", "member", "member_selected"} {
		ctxN, prod, capN := e4after[who+"_context"], e4after[who+"_production"], e4after[who+"_capability"]
		if prod != ctxN || capN < ctxN || capN > ctxN+4 {
			fmt.Println("E-MCP-TRIGGER eligibility statement bound")
			t.Errorf("%s: context %d, production catalog %d, capability catalog %d (want production = context, capability <= context+4)",
				who, ctxN, prod, capN)
		}
	}
	b4, _ := json.Marshal(map[string]any{"evidence": "E-MCP-4", "statements": e4after, "extra_groups": 30, "extra_profiles": 30})
	fmt.Println("E-MCP-4 " + string(b4))
	// E-MCP-1 and E-MCP-2: the plans, as the owner over the same statements, printed for review (descriptive).
	for name, q := range map[string]struct {
		sql  string
		args []any
	}{
		"E-MCP-1 manager form": {`EXPLAIN (COSTS OFF) SELECT id, organization_id, name, slug, created_at FROM app.groups WHERE organization_id = $1::text::uuid AND id > '00000000-0000-0000-0000-000000000000' ORDER BY id LIMIT 101`, []any{org.String()}},
		"E-MCP-1 member form":  {`EXPLAIN (COSTS OFF) SELECT g.id FROM app.group_memberships m JOIN app.groups g ON g.organization_id = m.organization_id AND g.id = m.group_id WHERE m.organization_id = $1::text::uuid AND m.user_id = $2::text::uuid ORDER BY g.id LIMIT 101`, []any{org.String(), member.id.String()}},
		"E-MCP-2 connection":   {`EXPLAIN (COSTS OFF) SELECT c.id FROM app.mcp_connections c JOIN app.sessions s ON s.id = c.session_id JOIN app.users u ON u.id = c.user_id WHERE c.id = $1::text::uuid`, []any{mc.ID()}},
	} {
		rows, err := e.setup.Query(ctx, q.sql, q.args...)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var plan []string
		for rows.Next() {
			var l string
			if err := rows.Scan(&l); err != nil {
				t.Fatal(err)
			}
			plan = append(plan, strings.TrimSpace(l))
		}
		rows.Close()
		b, _ := json.Marshal(map[string]any{"evidence": name, "plan": plan, "statements": after})
		fmt.Println("E-MCP " + string(b))
	}
}

// recordTracer is a pauseTracer that also records every statement's SQL (G063 O-MCP-9: the rename's statement order).
type recordTracer struct {
	*pauseTracer
	mu   sync.Mutex
	sqls []string
}

func (r *recordTracer) TraceQueryStart(ctx context.Context, c *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	r.mu.Lock()
	r.sqls = append(r.sqls, d.SQL)
	r.mu.Unlock()
	return r.pauseTracer.TraceQueryStart(ctx, c, d)
}

// take returns and clears the recorded statements.
func (r *recordTracer) take() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.sqls
	r.sqls = nil
	return out
}

// renameStatementKind classifies one statement of a rename transaction: the live user and session locks (the browser
// session and then the target 'mcp' session take the same two statements), the target session read, the locked facts,
// the connection row lock and the name write ("" for anything else, such as the context statement).
func renameStatementKind(sql string) string {
	switch {
	case strings.Contains(sql, "FROM app.users WHERE id = $1 AND status = 'active' FOR SHARE"):
		return "user"
	case strings.Contains(sql, "FROM app.sessions WHERE id = $1 AND user_id = $2") && strings.HasSuffix(strings.TrimSpace(sql), "FOR SHARE"):
		return "session"
	case strings.Contains(sql, "SELECT session_id::text FROM app.mcp_connections"):
		return "target"
	case strings.Contains(sql, "SELECT state, pending_expires_at FROM app.mcp_connections"):
		return "row"
	case strings.Contains(sql, "UPDATE app.mcp_connections SET display_name"):
		return "write"
	case strings.Contains(sql, "app.organizations") || strings.Contains(sql, "app.role_bindings"):
		return "facts"
	}
	return ""
}
