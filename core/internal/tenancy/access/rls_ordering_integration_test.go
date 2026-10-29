//go:build integration

// B-RLS-ORDERING (G008 Plan r3 §8.3): TestRLSOrdering against real PostgreSQL at version 6, with G007's deterministic
// barriers (hook pauses and lock-wait observation; no sleeps for ordering). Each case runs the guarded operation and one
// conflicting committed change in order A (guard first: the writer waits on the guard's row locks), B (writer first)
// and, where the writer takes a conflicting row lock, C (writer in flight: the guard waits). Either the guarded write is
// serialized before the change, or the changed authority is observed and the write is denied with the target unchanged.
//
//   - OR-1 profile write versus removal of the last qualifying Group path;
//   - OR-2 profile write versus unlink;
//   - OR-3 profile write versus the Organization flag turned off;
//   - OR-4 an old context, connection and prepared statement after each committed change (downgrade, removal, unlink,
//     flag, disable, session revoke, expiry), and manager demotion shrinking raw membership visibility immediately;
//   - OR-5 a real slug change upgrades the Group lock and orders a concurrent foreign-key insert behind it;
//   - OR-6 (G060 amendment01 A2) a raw owner UPDATE of the Group-management override queued behind an in-flight
//     DisableUser of that owner waits inside the guard (users FOR SHARE) and is refused after the wait.
//   - OR-7 (G060 attempt3, SR2-3) a raw Organization removal by an Organization admin who also owns the Group versus a
//     raw promotion of the removed subject from Group admin to Group co-owner: promotion committed first or in flight
//     (the cascade's locked read waits for it) → owner_subject with every binding and audit row of the subject kept;
//     removal in flight first → the promotion waits and then updates no row, nothing partial.
//   - OR-8 (G020 Plan r2 T6) a raw runtime delete of a label definition versus an in-flight facade assignment of it:
//     the delete waits and is refused 23503 by the total assignment → definition key, or (delete in flight first) the
//     assignment waits and observes not_found; never an orphan assignment.
//
// TestLabelsOrdering (G020 Plan r2 §1.3, T6 O-LAB-1..4) lives in this file rather than labels_integration_test.go
// (another writer's file in the same round; a recorded file-placement deviation) and reuses the same barriers.
//
// Guarded operations run through the runtime pool; conflicting tenancy writers are trusted model primitives on the
// attested setup pool (RC-8), identity writers are the real G004 services.
package access

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
	"core/internal/tenancy"
)

func TestRLSOrdering(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	org := e.org(t, "RLS Ordering "+e.canaries[0])
	creator := e.newUser(t)
	owner := e.newUser(t)
	e.grantOrg(t, org, owner, tenancy.OrgOwner)
	e.bind(t, owner, org) // G016 A2: owner's credential for org (setup only)
	zeroG := tenancy.GroupID{}
	seq := 0
	newGroup := func(t *testing.T) tenancy.Group {
		seq++
		return e.group(t, org, creator, "RLS Ordering Group", "rls-ordering-"+itH(seq)[:8])
	}
	newProfile := func(t *testing.T) tenancy.BusinessProfile { return e.profile(t, org, "RLS Ordering Profile") }
	groupAdmin := func(t *testing.T, g tenancy.GroupID) *itUser {
		u := e.newUser(t)
		e.grantOrg(t, org, u, tenancy.OrgMember)
		e.grantGroup(t, org, g, u, tenancy.GroupAdmin)
		e.bind(t, u, org) // G016 A2: u's credential for org, prepared before any ordering window
		return u
	}
	profileRow := func(p tenancy.BusinessProfileID) func() string {
		return func() string {
			var s string
			if err := e.setup.QueryRow(ctx, `SELECT md5(r::text) FROM app.business_profiles r WHERE r.id = $1::text::uuid`, p.String()).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
	}
	setupWriter := func() (*tenancy.Service, *hookDB) {
		h := &hookDB{pool: e.setupDB}
		return tenancy.NewService(h, nil), h
	}
	update := func(svc *Service, rctx context.Context, tc *TenantContext, p tenancy.BusinessProfile) func() error {
		return func() error {
			desc := "Ordered " + e.canaries[1]
			_, err := svc.UpdateBusinessProfile(rctx, tc, p.ID, tenancy.BusinessContext{DisplayName: p.Context.DisplayName, Description: &desc})
			return err
		}
	}

	type prepared struct {
		guard, writer func() error
		ghook, whook  *hookDB
		target        func() string
	}
	run := func(t *testing.T, id string, order byte, pr prepared, deny Code) {
		t.Helper()
		before := pr.target()
		switch order {
		case 'A':
			p := newPause()
			pr.ghook.on("before", "W", p.hook)
			g := start(pr.guard)
			p.wait(t, "guard")
			w := start(pr.writer)
			e.waitWaiters(t, 1)
			close(p.release)
			gerr, werr := g.wait(t, "guard"), w.wait(t, "writer")
			if gerr != nil || werr != nil || g.at > w.at {
				t.Errorf("%s A: guard %v (finished %d), writer %v (finished %d)", id, gerr, g.at, werr, w.at)
			}
			if pr.target() == before {
				t.Errorf("%s A: the guarded write did not persist", id)
			}
		case 'B':
			if err := pr.writer(); err != nil {
				t.Fatalf("%s B writer: %v", id, err)
			}
			if err := pr.guard(); CodeOf(err) != deny {
				t.Errorf("%s B: guard %v, want %s", id, err, deny)
			}
			if pr.target() != before {
				t.Errorf("%s B: the target changed", id)
			}
		case 'C':
			p := newPause()
			pr.whook.on("before", "COMMIT", p.hook)
			w := start(pr.writer)
			p.wait(t, "writer")
			g := start(pr.guard)
			e.waitWaiters(t, 1)
			close(p.release)
			werr, gerr := w.wait(t, "writer"), g.wait(t, "guard")
			if werr != nil || CodeOf(gerr) != deny {
				t.Errorf("%s C: writer %v, guard %v (want %s)", id, werr, gerr, deny)
			}
			if pr.target() != before {
				t.Errorf("%s C: the target changed", id)
			}
		}
		pr.ghook.reset()
		if pr.whook != nil {
			pr.whook.reset()
		}
		ordering(t, map[string]any{"case": id, "order": string(order)})
	}

	t.Run("OR-1 profile write versus removal of the last qualifying path", func(t *testing.T) {
		for _, order := range []byte("ABC") {
			g, p := newGroup(t), newProfile(t)
			e.link(t, org, g.ID, p.ID)
			u := groupAdmin(t, g.ID)
			gsvc, ghook := e.service(e.id)
			rctx, _ := beginAs(t, gsvc, ctx, u, org) // G016 A2
			tc := resolve(t, gsvc, rctx, org, zeroG)
			wm, wh := setupWriter()
			run(t, "OR-1", order, prepared{guard: update(gsvc, rctx, tc, p), ghook: ghook, whook: wh, target: profileRow(p.ID),
				writer: func() error { return wm.RemoveGroupMembership(ctx, org, g.ID, u.id) }}, CodeNotFound)
		}
	})
	t.Run("OR-2 profile write versus unlink", func(t *testing.T) {
		for _, order := range []byte("ABC") {
			g, p := newGroup(t), newProfile(t)
			e.link(t, org, g.ID, p.ID)
			u := groupAdmin(t, g.ID)
			gsvc, ghook := e.service(e.id)
			rctx, _ := beginAs(t, gsvc, ctx, u, org) // G016 A2
			tc := resolve(t, gsvc, rctx, org, zeroG)
			wm, wh := setupWriter()
			run(t, "OR-2", order, prepared{guard: update(gsvc, rctx, tc, p), ghook: ghook, whook: wh, target: profileRow(p.ID),
				writer: func() error { return wm.UnlinkGroupBusinessProfile(ctx, org, g.ID, p.ID) }}, CodeNotFound)
		}
	})
	t.Run("OR-3 profile write versus the flag turned off", func(t *testing.T) {
		for _, order := range []byte("ABC") {
			e.setFlag(t, org, true)
			g, p := newGroup(t), newProfile(t)
			e.link(t, org, g.ID, p.ID)
			u := groupAdmin(t, g.ID)
			gsvc, ghook := e.service(e.id)
			rctx, _ := beginAs(t, gsvc, ctx, u, org) // G016 A2
			tc := resolve(t, gsvc, rctx, org, zeroG)
			psvc, phook := e.service(e.id)
			prctx, _ := beginAs(t, psvc, ctx, owner, org) // G016 A2
			ptc := resolve(t, psvc, prctx, org, zeroG)
			run(t, "OR-3", order, prepared{guard: update(gsvc, rctx, tc, p), ghook: ghook, whook: phook, target: profileRow(p.ID),
				writer: func() error {
					_, err := psvc.SetOrganizationPolicy(prctx, ptc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
					return err
				}}, CodeForbidden)
		}
		e.setFlag(t, org, true)
	})

	t.Run("OR-6 a raw owner override UPDATE queued behind an in-flight DisableUser of that owner", func(t *testing.T) {
		o2 := e.newUser(t)
		e.grantOrg(t, org, o2, tenancy.OrgOwner) // a second owner; DisableUser keeps the binding
		tracer := &pauseTracer{}
		idTraced := e.identity(t, e.clk, e.runtimePool(t, tracer))
		dp := tracer.arm("revoke_reason = 'disabled'") // DisableUser holds o2's users row FOR UPDATE, before its COMMIT
		d := start(func() error { return idTraced.DisableUser(ctx, o2.id.String()) })
		dp.wait(t, "DisableUser in flight")
		rt := e.rlsRuntime(t)
		u, o := o2.id.String(), org.String()
		raw := start(func() error {
			tx, err := rt.Begin(ctx)
			if err != nil {
				return err
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if err := rlsSet(ctx, tx, &u, &o, ptr(""), nil); err != nil {
				return err
			}
			_, err = tx.Exec(ctx, `UPDATE app.organizations SET group_management_override = true WHERE id = $1::text::uuid`, o)
			if err != nil {
				return err
			}
			return tx.Commit(ctx)
		})
		obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		waited := false
		for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
			if raw.finished() {
				break
			}
			if e.waiters(t, obs) >= 1 {
				waited = true
				break
			}
		}
		if !waited {
			t.Fatal("the raw owner UPDATE neither waited for the in-flight DisableUser nor finished")
		}
		close(dp.release)
		if err := d.wait(t, "DisableUser"); err != nil {
			t.Fatalf("DisableUser: %v", err)
		}
		err := raw.wait(t, "raw owner UPDATE")
		if pgState(err) != "42501" || pgConstraint(err) != "group_override_owner" {
			t.Errorf("raw owner UPDATE after the wait: %v (SQLSTATE %q constraint %q), want 42501 group_override_owner", err, pgState(err), pgConstraint(err))
		}
		var v bool
		if err := e.setup.QueryRow(ctx, `SELECT group_management_override FROM app.organizations WHERE id = $1::text::uuid`, o).Scan(&v); err != nil || v {
			t.Errorf("the override after the refused UPDATE: %t (%v), want false", v, err)
		}
	})

	t.Run("OR-7 raw Organization removal versus a raw Group-admin to co-owner promotion of the same subject", func(t *testing.T) {
		// The remover is an Organization admin and the actual owner of g; the target is an Organization member, g's admin
		// and a member of g2 (owned by creator; the remover holds no g2 role and the override is off, so the cascade
		// reaches g2 only through the trigger-depth branch, including its row lock).
		obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		o := org.String()
		bindingID := func(t *testing.T, u *itUser, g *tenancy.GroupID) string {
			t.Helper()
			var id string
			q := `SELECT id::text FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid
  AND scope_type = 'organization'`
			args := []any{o, u.id.String()}
			if g != nil {
				q = `SELECT id::text FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid
  AND scope_type = 'group' AND group_id = $3::text::uuid`
				args = append(args, g.String())
			}
			if err := e.setup.QueryRow(ctx, q, args...).Scan(&id); err != nil {
				t.Fatal(err)
			}
			return id
		}
		// state is the subject's bindings (scope:role, ordered) and its audit rows per action.
		state := func(t *testing.T, u *itUser, g, g2 tenancy.GroupID) (bindings, audit string) {
			t.Helper()
			if err := e.setup.QueryRow(ctx, `SELECT coalesce(string_agg(x.v, ',' ORDER BY x.v COLLATE "C"), '') FROM (
  SELECT CASE WHEN rb.group_id = $3::text::uuid THEN 'g' WHEN rb.group_id = $4::text::uuid THEN 'g2' ELSE rb.scope_type END
    || ':' || rb.role AS v
  FROM app.role_bindings rb WHERE rb.organization_id = $1::text::uuid AND rb.subject_user_id = $2::text::uuid) x`,
				o, u.id.String(), g.String(), g2.String()).Scan(&bindings); err != nil {
				t.Fatal(err)
			}
			if err := e.setup.QueryRow(ctx, `SELECT coalesce(string_agg(x.v, ',' ORDER BY x.v COLLATE "C"), '') FROM (
  SELECT ev.action || '=' || count(*) AS v FROM audit.events ev WHERE ev.organization_id = $1::text::uuid AND ev.subject_user_id = $2::text::uuid
  GROUP BY ev.action) x`, o, u.id.String()).Scan(&audit); err != nil {
				t.Fatal(err)
			}
			return bindings, audit
		}
		// addAction returns the audit summary with one more row of action (same "action=n" form, byte order).
		addAction := func(audit, action string) string {
			counts := map[string]int{}
			for _, kv := range strings.Split(audit, ",") {
				if kv == "" {
					continue
				}
				k, v, _ := strings.Cut(kv, "=")
				n, _ := strconv.Atoi(v)
				counts[k] = n
			}
			counts[action]++
			keys := make([]string, 0, len(counts))
			for k, n := range counts {
				keys = append(keys, k+"="+strconv.Itoa(n))
			}
			sort.Strings(keys)
			return strings.Join(keys, ",")
		}
		// raw runs one statement as the remover in an Organization-only context on its own runtime connection; hold keeps
		// the transaction open until released (nil: commit at once).
		type rawTx struct {
			rows int64
			tx   pgx.Tx
		}
		begin := func(t *testing.T, conn *pgx.Conn, u *itUser) pgx.Tx {
			t.Helper()
			tx, err := conn.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			uid := u.id.String()
			if err := rlsSet(ctx, tx, &uid, &o, ptr(""), nil); err != nil {
				t.Fatal(err)
			}
			return tx
		}
		const promoteSQL = `UPDATE app.role_bindings SET role = 'group_owner' WHERE id = $1::text::uuid`
		const removeSQL = `DELETE FROM app.role_bindings WHERE id = $1::text::uuid`
		waitBlocked := func(t *testing.T, a *actor, what string) {
			t.Helper()
			for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
				if a.finished() {
					t.Fatalf("%s finished without waiting: %v", what, a.err)
				}
				if e.waiters(t, obs) >= 1 {
					return
				}
			}
			t.Fatalf("%s never waited", what)
		}
		fixture := func(t *testing.T) (*itUser, *itUser, tenancy.GroupID, tenancy.GroupID) {
			t.Helper()
			remover, target := e.newUser(t), e.newUser(t)
			e.grantOrg(t, org, remover, tenancy.OrgAdmin)
			seq++
			g := e.group(t, org, remover, "RLS Ordering Owned Group", "rls-ordering-owned-"+itH(seq)[:8]).ID
			g2 := newGroup(t).ID
			e.grantOrg(t, org, target, tenancy.OrgMember)
			e.grantGroup(t, org, g, target, tenancy.GroupAdmin)
			e.grantGroup(t, org, g2, target, tenancy.GroupMember)
			return remover, target, g, g2
		}
		const startBindings = "g2:group_member,g:group_admin,organization:org_member"
		for _, variant := range []string{"promotion committed first", "promotion in flight first", "removal in flight first"} {
			remover, target, g, g2 := fixture(t)
			if b, _ := state(t, target, g, g2); b != startBindings {
				t.Fatalf("%s: fixture bindings %q", variant, b)
			}
			_, audit0 := state(t, target, g, g2)
			orgB, gB := bindingID(t, target, nil), bindingID(t, target, &g)
			rtP, rtR := e.rlsRuntime(t), e.rlsRuntime(t)
			var p, r rawTx
			switch variant {
			case "promotion committed first", "promotion in flight first":
				p.tx = begin(t, rtP, remover)
				tag, err := p.tx.Exec(ctx, promoteSQL, gB)
				if err != nil || tag.RowsAffected() != 1 {
					t.Fatalf("%s: promotion %v (%d rows)", variant, err, tag.RowsAffected())
				}
				r.tx = begin(t, rtR, remover)
				var rerr error
				if variant == "promotion committed first" {
					if err := p.tx.Commit(ctx); err != nil {
						t.Fatalf("%s: promotion commit: %v", variant, err)
					}
					_, rerr = r.tx.Exec(ctx, removeSQL, orgB)
				} else {
					rem := start(func() error { _, err := r.tx.Exec(ctx, removeSQL, orgB); return err })
					waitBlocked(t, rem, "the removal behind the in-flight promotion")
					if err := p.tx.Commit(ctx); err != nil {
						t.Fatalf("%s: promotion commit: %v", variant, err)
					}
					rerr = rem.wait(t, "removal")
				}
				if pgState(rerr) != "42501" || pgConstraint(rerr) != "owner_subject" {
					t.Errorf("%s: removal %v (SQLSTATE %q constraint %q), want 42501 owner_subject", variant, rerr, pgState(rerr), pgConstraint(rerr))
				}
				_ = r.tx.Rollback(ctx)
				b, audit := state(t, target, g, g2)
				if b != "g2:group_member,g:group_owner,organization:org_member" {
					t.Errorf("%s: bindings %q, want all three kept with the promotion", variant, b)
				}
				if want := addAction(audit0, "membership.role_changed"); audit != want {
					t.Errorf("%s: audit %q, want %q (the promotion only; no removed row)", variant, audit, want)
				}
			case "removal in flight first":
				r.tx = begin(t, rtR, remover)
				tag, err := r.tx.Exec(ctx, removeSQL, orgB)
				if err != nil || tag.RowsAffected() != 1 {
					t.Fatalf("%s: removal %v (%d rows)", variant, err, tag.RowsAffected())
				}
				p.tx = begin(t, rtP, remover)
				pro := start(func() error {
					tag, err := p.tx.Exec(ctx, promoteSQL, gB)
					p.rows = tag.RowsAffected()
					return err
				})
				waitBlocked(t, pro, "the promotion behind the in-flight removal")
				if err := r.tx.Commit(ctx); err != nil {
					t.Fatalf("%s: removal commit: %v", variant, err)
				}
				if err := pro.wait(t, "promotion"); err != nil || p.rows != 0 {
					t.Errorf("%s: promotion %v (%d rows), want no row updated", variant, err, p.rows)
				}
				if err := p.tx.Commit(ctx); err != nil {
					t.Errorf("%s: empty promotion commit: %v", variant, err)
				}
				b, audit := state(t, target, g, g2)
				if b != "" {
					t.Errorf("%s: bindings %q remain, want none", variant, b)
				}
				want := addAction(addAction(addAction(audit0, "membership.removed"), "membership.removed"), "membership.removed")
				if audit != want {
					t.Errorf("%s: audit %q, want %q (three removed rows, no role change)", variant, audit, want)
				}
			}
			ordering(t, map[string]any{"case": "OR-7", "variant": variant})
		}
	})

	t.Run("OR-4 an old context, connection and prepared statement after each committed change", func(t *testing.T) {
		type change struct {
			name string
			do   func(u *itUser, g tenancy.Group, p tenancy.BusinessProfile, clk *itClock)
			deny Code
		}
		changes := []change{
			{"downgrade", func(u *itUser, g tenancy.Group, _ tenancy.BusinessProfile, _ *itClock) {
				if _, err := e.model.ChangeGroupRole(ctx, org, g.ID, u.id, tenancy.GroupMember); err != nil {
					t.Fatal(err)
				}
			}, CodeForbidden},
			{"removal", func(u *itUser, g tenancy.Group, _ tenancy.BusinessProfile, _ *itClock) {
				if err := e.model.RemoveGroupMembership(ctx, org, g.ID, u.id); err != nil {
					t.Fatal(err)
				}
			}, CodeNotFound},
			{"unlink", func(_ *itUser, g tenancy.Group, p tenancy.BusinessProfile, _ *itClock) {
				if err := e.model.UnlinkGroupBusinessProfile(ctx, org, g.ID, p.ID); err != nil {
					t.Fatal(err)
				}
			}, CodeNotFound},
			{"flag", func(*itUser, tenancy.Group, tenancy.BusinessProfile, *itClock) { e.setFlag(t, org, false) }, CodeForbidden},
			{"disable", func(u *itUser, _ tenancy.Group, _ tenancy.BusinessProfile, _ *itClock) {
				if err := e.id.DisableUser(ctx, u.id.String()); err != nil {
					t.Fatal(err)
				}
			}, CodeUnauthenticated},
			{"session revoke", func(u *itUser, _ tenancy.Group, _ tenancy.BusinessProfile, _ *itClock) {
				e.logout(t, u.in(t, org), identity.ScopeCurrent) // G016 A2: revokes the credential of the scope under test
			}, CodeUnauthenticated},
			{"expiry", func(_ *itUser, _ tenancy.Group, _ tenancy.BusinessProfile, clk *itClock) {
				clk.Advance(itIdle + time.Minute)
			}, CodeUnauthenticated},
		}
		for _, c := range changes {
			e.setFlag(t, org, true)
			g, p := newGroup(t), newProfile(t)
			e.link(t, org, g.ID, p.ID)
			u := groupAdmin(t, g.ID)
			clk := newITClock()
			id := e.identity(t, clk, e.pool)
			gsvc, ghook := e.service(id)
			rctx, _ := beginAs(t, gsvc, ctx, u, org) // G016 A2
			tc := resolve(t, gsvc, rctx, org, zeroG)
			if err := update(gsvc, rctx, tc, p)(); err != nil {
				t.Fatalf("%s: positive before the change: %v", c.name, err)
			}
			c.do(u, g, p, clk)
			before := profileRow(p.ID)()
			if err := update(gsvc, rctx, tc, p)(); CodeOf(err) != c.deny {
				t.Errorf("%s: the old context after the change: %v, want %s", c.name, err, c.deny)
			}
			if profileRow(p.ID)() != before {
				t.Errorf("%s: the target changed", c.name)
			}
			ghook.reset()
			ordering(t, map[string]any{"case": "OR-4", "change": c.name})
		}
		e.setFlag(t, org, true)
		// Manager demotion: on one runtime connection, the same prepared statement re-executed after the change sees only
		// the caller's own bindings.
		g := newGroup(t)
		mgr := groupAdmin(t, g.ID)
		member := e.newUser(t)
		e.grantOrg(t, org, member, tenancy.OrgMember)
		e.grantGroup(t, org, g.ID, member, tenancy.GroupMember)
		rt := e.rlsRuntime(t)
		const q = `SELECT count(*) FROM app.group_memberships WHERE group_id = $1`
		count := func() int {
			var n int
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, mgr.id.String(), org.String()); err != nil {
				t.Fatal(err)
			}
			if err := tx.QueryRow(ctx, q, g.ID.String()).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		if n := count(); n != 3 {
			t.Fatalf("manager sees %d Group bindings, want 3 (the creator's owner binding, its own and the member's)", n)
		}
		if _, err := e.model.ChangeGroupRole(ctx, org, g.ID, mgr.id, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		if n := count(); n != 1 {
			t.Errorf("after demotion the same prepared statement sees %d Group bindings, want only the caller's own", n)
		}
		ordering(t, map[string]any{"case": "OR-4", "change": "manager demotion"})
	})

	t.Run("OR-5 a real slug change upgrades the Group lock and orders a foreign-key insert behind it", func(t *testing.T) {
		g := newGroup(t)
		u := groupAdmin(t, g.ID)
		gsvc, ghook := e.service(e.id)
		rctx, _ := beginAs(t, gsvc, ctx, u, org) // G016 A2
		tc := resolve(t, gsvc, rctx, org, zeroG)
		other := e.newUser(t)
		p := newPause()
		ghook.on("before", "COMMIT", p.hook)
		guard := start(func() error { _, err := gsvc.ChangeGroupSlug(rctx, tc, g.ID, "rls-moved-"+itH(g.ID)[:8]); return err })
		p.wait(t, "guard")
		w := start(func() error {
			_, err := e.model.GrantGroupRole(ctx, org, g.ID, other.id, tenancy.GroupMember)
			return err
		})
		e.waitWaiters(t, 1)
		if w.finished() {
			t.Fatal("the foreign-key insert did not wait for the upgraded Group lock")
		}
		close(p.release)
		if err := guard.wait(t, "guard"); err != nil {
			t.Fatalf("slug change: %v", err)
		}
		if err := w.wait(t, "writer"); err != nil {
			t.Fatalf("grant after the slug change: %v", err)
		}
		if guard.at > w.at {
			t.Error("the insert finished before the slug change committed")
		}
		var locked string
		if err := e.setup.QueryRow(ctx, `SELECT slug FROM app.groups WHERE id = $1::text::uuid`, g.ID.String()).Scan(&locked); err != nil || locked != "rls-moved-"+itH(g.ID)[:8] {
			t.Errorf("slug %q (%v)", locked, err)
		}
		ghook.reset()
		ordering(t, map[string]any{"case": "OR-5"})
	})

	t.Run("OR-8 a raw definition delete versus an in-flight assignment of that definition", func(t *testing.T) {
		// G020 (Plan r2 §1.3, T6): the facade assignment holds the target Group FOR NO KEY UPDATE and then the definition
		// FOR SHARE until its COMMIT. A raw runtime DELETE of the definition by the Organization owner (Organization-wide
		// context; the label_definitions DELETE policy admits it) waits for that share lock and is then refused by the
		// total assignment → definition key (23503 label_assignments_label_fkey); with the delete in flight first, the
		// assignment waits on the deleted row and observes not_found. Either way no assignment without its definition.
		obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		o, u := org.String(), owner.id.String()
		waitBlocked := func(t *testing.T, a *actor, what string) {
			t.Helper()
			for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
				if a.finished() {
					t.Fatalf("%s finished without waiting: %v", what, a.err)
				}
				if e.waiters(t, obs) >= 1 {
					return
				}
			}
			t.Fatalf("%s never waited", what)
		}
		// definition inserts one Organization-owned definition of org as the attested superuser (fixture only).
		definition := func(t *testing.T) string {
			t.Helper()
			seq++
			id := rlsUUID(t)
			if _, err := e.setup.Exec(ctx, `INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility)
VALUES ($1::uuid, $2::text::uuid, 'organization', $2::text::uuid, NULL, $3, 'owner_scope')`, id, o, "RLS Ordering label "+itH(seq)[:8]); err != nil {
				t.Fatal(err)
			}
			return id
		}
		// state counts the definition, its assignment to g and every assignment of org without its definition.
		state := func(t *testing.T, def string, g tenancy.GroupID) string {
			t.Helper()
			var d, a, orphans int
			if err := e.setup.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.label_definitions WHERE id = $1::uuid)::int,
  (SELECT count(*) FROM app.label_assignments WHERE label_id = $1::uuid AND target_group_id = $2::text::uuid)::int,
  (SELECT count(*) FROM app.label_assignments x WHERE x.organization_id = $3::text::uuid
     AND NOT EXISTS (SELECT 1 FROM app.label_definitions y WHERE y.id = x.label_id))::int`, def, g.String(), o).Scan(&d, &a, &orphans); err != nil {
				t.Fatal(err)
			}
			return fmt.Sprintf("definitions=%d assignments=%d orphans=%d", d, a, orphans)
		}
		const rawDelete = `DELETE FROM app.label_definitions WHERE id = $1::uuid`
		for _, variant := range []string{"assign in flight first", "delete in flight first"} {
			g := newGroup(t)
			def := definition(t)
			gsvc, ghook := e.service(e.id)
			rctx, _ := beginAs(t, gsvc, ctx, owner, org)
			tc := resolve(t, gsvc, rctx, org, zeroG)
			assign := func() error { return gsvc.AssignLabel(rctx, tc, def, LabelTargetSelector{Group: g.ID.String()}) }
			rt := e.rlsRuntime(t)
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if err := rlsSet(ctx, tx, &u, &o, ptr(""), nil); err != nil {
				t.Fatal(err)
			}
			var want string
			switch variant {
			case "assign in flight first":
				p := newPause()
				ghook.on("before", "COMMIT", p.hook)
				a := start(assign)
				p.wait(t, "assign")
				del := start(func() error { _, err := tx.Exec(ctx, rawDelete, def); return err })
				waitBlocked(t, del, "the raw delete behind the in-flight assign")
				close(p.release)
				if err := a.wait(t, "assign"); err != nil {
					t.Errorf("%s: assign %v, want committed", variant, err)
				}
				derr := del.wait(t, "raw delete")
				if pgState(derr) != "23503" || pgConstraint(derr) != "label_assignments_label_fkey" {
					t.Errorf("%s: raw delete %v (SQLSTATE %q constraint %q), want 23503 label_assignments_label_fkey", variant, derr, pgState(derr), pgConstraint(derr))
				}
				_ = tx.Rollback(ctx)
				want = "definitions=1 assignments=1 orphans=0"
			case "delete in flight first":
				tag, err := tx.Exec(ctx, rawDelete, def)
				if err != nil || tag.RowsAffected() != 1 {
					t.Fatalf("%s: raw delete %v (%d rows)", variant, err, tag.RowsAffected())
				}
				a := start(assign)
				waitBlocked(t, a, "the assign behind the in-flight raw delete")
				if err := tx.Commit(ctx); err != nil {
					t.Fatalf("%s: raw delete commit: %v", variant, err)
				}
				if err := a.wait(t, "assign"); CodeOf(err) != CodeNotFound {
					t.Errorf("%s: assign after the committed delete %v, want not_found", variant, err)
				}
				want = "definitions=0 assignments=0 orphans=0"
			}
			if got := state(t, def, g.ID); got != want {
				t.Errorf("%s: %s, want %s", variant, got, want)
			}
			ghook.reset()
			ordering(t, map[string]any{"case": "OR-8", "variant": variant})
		}
	})
	_ = pgx.ErrNoRows
}

// TestLabelsOrdering (G020 Plan r2 §1.3, T6) runs the label operations of the facade against each other and against
// committed authority changes with the G007 barriers (a pause before COMMIT and an observed lock wait, no sleeps):
//
//   - O-LAB-1 assignment versus definition deletion, both orders: serialized on the definition row (FOR SHARE versus FOR
//     NO KEY UPDATE); assign first → the delete removes the new assignment too (READ COMMITTED statement snapshot);
//     delete first → the assignment observes not_found; never an orphan;
//   - O-LAB-2 visibility narrowed → the next read of an old context is empty; the Group role removed → the next read is
//     not_found; a selected-Group switch → another Group's labels are not_found from the new context;
//   - O-LAB-3 a duplicate assignment race: one already_assigned, one row, one audit row;
//   - O-LAB-4 creations racing the 200 bound: two Organization-scope creations serialize on the up-front Organization FOR
//     NO KEY UPDATE lock (one scope_full, 200 rows, no 40P01); two Group-scope creations in one Group likewise on the
//     Group row; a mixed Organization/Group pair commits both in either order without a deadlock.
func TestLabelsOrdering(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	zeroG := tenancy.GroupID{}
	seq := 0
	next := func() int { seq++; return seq }
	creator := e.newUser(t)
	// newOrg returns a new Organization with a bound owner.
	newOrg := func(t *testing.T) (tenancy.OrganizationID, *itUser) {
		t.Helper()
		o := e.org(t, fmt.Sprintf("Labels Ordering %d %s", next(), e.canaries[0]))
		u := e.newUser(t)
		e.grantOrg(t, o, u, tenancy.OrgOwner)
		e.bind(t, u, o) // G016 A2: the owner's credential for o, prepared before any ordering window
		return o, u
	}
	newGroup := func(t *testing.T, o tenancy.OrganizationID) tenancy.Group {
		t.Helper()
		return e.group(t, o, creator, "Labels Ordering Group", "labels-ordering-"+itH(next())[:8])
	}
	// member returns a new subject of o with the Organization role and the given Group roles, bound to o.
	member := func(t *testing.T, o tenancy.OrganizationID, r tenancy.OrganizationRole, groups map[tenancy.GroupID]tenancy.GroupRole) *itUser {
		t.Helper()
		u := e.newUser(t)
		e.grantOrg(t, o, u, r)
		for g, gr := range groups {
			e.grantGroup(t, o, g, u, gr)
		}
		e.bind(t, u, o)
		return u
	}
	// as resolves sel for u in o in a fresh request scope of a separate service (its own hookDB).
	type actorCtx struct {
		svc  *Service
		hook *hookDB
		rctx context.Context
		tc   *TenantContext
	}
	as := func(t *testing.T, u *itUser, o tenancy.OrganizationID, sel tenancy.GroupID) actorCtx {
		t.Helper()
		svc, h := e.service(e.id)
		rctx, _ := beginAs(t, svc, ctx, u, o)
		return actorCtx{svc: svc, hook: h, rctx: rctx, tc: resolve(t, svc, rctx, o, sel)}
	}
	orgScope := Scope{Kind: tenancy.ScopeOrganization}
	groupScope := func(g tenancy.GroupID) Scope { return Scope{Kind: tenancy.ScopeGroup, Group: &g} }
	target := func(g tenancy.GroupID) LabelTargetSelector { return LabelTargetSelector{Group: g.String()} }
	name := func() string { return fmt.Sprintf("Ordering label %d", next()) }
	field := func(err error) string {
		var ae *Error
		if errors.As(err, &ae) {
			return ae.Field
		}
		return ""
	}
	// serialize runs first paused before its COMMIT, starts second, waits until second waits for a lock, then releases
	// first; it returns both errors and whether first finished before second.
	serialize := func(t *testing.T, hFirst *hookDB, first, second func() error) (error, error, bool) {
		t.Helper()
		p := newPause()
		hFirst.on("before", "COMMIT", p.hook)
		a := start(first)
		p.wait(t, "first")
		b := start(second)
		e.waitWaiters(t, 1)
		close(p.release)
		ea, eb := a.wait(t, "first"), b.wait(t, "second")
		hFirst.reset()
		return ea, eb, a.at < b.at
	}
	// rows counts the definitions of an owner scope, the assignments of a definition and orphans of an Organization.
	scopeRows := func(t *testing.T, o tenancy.OrganizationID, scopeType string, scopeID [16]byte) int {
		t.Helper()
		var n int
		if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM app.label_definitions WHERE organization_id = $1::text::uuid
  AND owner_scope_type = $2 AND owner_scope_id = $3::uuid`, o.String(), scopeType, fmt.Sprintf("%x", scopeID[:])).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	labelState := func(t *testing.T, o tenancy.OrganizationID, def string) string {
		t.Helper()
		var d, a, orphans int
		if err := e.setup.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.label_definitions WHERE id = $1::text::uuid)::int,
  (SELECT count(*) FROM app.label_assignments WHERE label_id = $1::text::uuid)::int,
  (SELECT count(*) FROM app.label_assignments x WHERE x.organization_id = $2::text::uuid
     AND NOT EXISTS (SELECT 1 FROM app.label_definitions y WHERE y.id = x.label_id))::int`, def, o.String()).Scan(&d, &a, &orphans); err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("definitions=%d assignments=%d orphans=%d", d, a, orphans)
	}
	// auditOf summarizes the audit rows of one definition as "action=n" in byte order.
	auditOf := func(t *testing.T, def string) string {
		t.Helper()
		var s string
		if err := e.setup.QueryRow(ctx, `SELECT coalesce(string_agg(x.v, ',' ORDER BY x.v COLLATE "C"), '') FROM (
  SELECT ev.action || '=' || count(*) AS v FROM audit.events ev WHERE ev.resource_type = 'label' AND ev.resource_id = $1::text::uuid
  GROUP BY ev.action) x`, def).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	// seed inserts n definitions of one owner scope as the attested superuser (fixtures only; the bound trigger admits
	// only the runtime and owner roles, so the seed is not itself bounded, and the facade's next creation counts it).
	seed := func(t *testing.T, o tenancy.OrganizationID, g *tenancy.GroupID, n int) {
		t.Helper()
		typ, owner, group := "organization", o.String(), any(nil)
		if g != nil {
			typ, owner, group = "group", g.String(), g.String()
		}
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility)
SELECT gen_random_uuid(), $1::text::uuid, $2, $3::text::uuid, $4::text::uuid, 'Seed label ' || i, 'owner_scope' FROM generate_series(1, $5::int) i`,
			o.String(), typ, owner, group, n); err != nil {
			t.Fatal(err)
		}
	}

	org, owner := newOrg(t)
	admin := member(t, org, tenancy.OrgAdmin, nil)

	t.Run("O-LAB-1 assignment versus definition deletion in both orders, no orphan", func(t *testing.T) {
		for _, order := range []string{"assign first", "delete first"} {
			g := newGroup(t, org)
			oc := as(t, owner, org, zeroG)
			def, err := oc.svc.CreateLabelDefinition(oc.rctx, oc.tc, orgScope, name(), string(tenancy.LabelVisibilityOwnerScope))
			if err != nil {
				t.Fatalf("%s: create: %v", order, err)
			}
			id := def.ID.String()
			ac, dc := as(t, owner, org, zeroG), as(t, admin, org, zeroG)
			var removed LabelDeleteResult
			assign := func() error { return ac.svc.AssignLabel(ac.rctx, ac.tc, id, target(g.ID)) }
			del := func() error {
				var err error
				removed, err = dc.svc.DeleteLabelDefinition(dc.rctx, dc.tc, id)
				return err
			}
			var wantAudit string
			switch order {
			case "assign first":
				ea, ed, inOrder := serialize(t, ac.hook, assign, del)
				if ea != nil || ed != nil || !inOrder || removed.AssignmentsRemoved != 1 {
					t.Errorf("%s: assign %v, delete %v (%d removed, in order %t), want both, one removed", order, ea, ed, removed.AssignmentsRemoved, inOrder)
				}
				wantAudit = "label.assigned=1,label.created=1,label.deleted=1,label.unassigned=1"
			case "delete first":
				ed, ea, _ := serialize(t, dc.hook, del, assign)
				if ed != nil || removed.AssignmentsRemoved != 0 || CodeOf(ea) != CodeNotFound {
					t.Errorf("%s: delete %v (%d removed), assign %v, want the delete then not_found", order, ed, removed.AssignmentsRemoved, ea)
				}
				wantAudit = "label.created=1,label.deleted=1"
			}
			if got := labelState(t, org, id); got != "definitions=0 assignments=0 orphans=0" {
				t.Errorf("%s: %s, want nothing left and no orphan", order, got)
			}
			if got := auditOf(t, id); got != wantAudit {
				t.Errorf("%s: audit %q, want %q", order, got, wantAudit)
			}
			ordering(t, map[string]any{"case": "O-LAB-1", "order": order})
		}
	})

	t.Run("O-LAB-2 narrowed visibility, a removed Group role and a selected-Group switch govern the next read", func(t *testing.T) {
		g1, g2 := newGroup(t, org), newGroup(t, org)
		mgr := member(t, org, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g1.ID: tenancy.GroupAdmin})
		dave := member(t, org, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g1.ID: tenancy.GroupMember, g2.ID: tenancy.GroupMember})
		mc := as(t, mgr, org, zeroG)
		def, err := mc.svc.CreateLabelDefinition(mc.rctx, mc.tc, groupScope(g1.ID), name(), string(tenancy.LabelVisibilityTargetMembers))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		read := func(c actorCtx) (int, error) {
			p, err := c.svc.ListLabelDefinitions(c.rctx, c.tc, groupScope(g1.ID), "", 0)
			return len(p.Labels), err
		}
		old, sel1, sel2 := as(t, dave, org, zeroG), as(t, dave, org, g1.ID), as(t, dave, org, g2.ID)
		if n, err := read(old); err != nil || n != 1 {
			t.Fatalf("member reads the target_members label: %d %v, want 1", n, err)
		}
		// Selected-Group switch: the same subject with g1 selected reads it; with g2 selected g1 is outside the context.
		if n, err := read(sel1); err != nil || n != 1 {
			t.Errorf("g1 selected: %d %v, want 1", n, err)
		}
		if _, err := read(sel2); CodeOf(err) != CodeNotFound {
			t.Errorf("g2 selected reads g1's labels: %v, want not_found", err)
		}
		ordering(t, map[string]any{"case": "O-LAB-2", "change": "selected-Group switch"})
		// Narrowing commits; the old context's next read takes a new snapshot and sees none.
		if _, err := mc.svc.UpdateLabelDefinition(mc.rctx, mc.tc, def.ID.String(), def.Name, string(tenancy.LabelVisibilityOwnerScope)); err != nil {
			t.Fatalf("narrow: %v", err)
		}
		if n, err := read(old); err != nil || n != 0 {
			t.Errorf("after narrowing the old context reads %d %v, want 0 and no error", n, err)
		}
		ordering(t, map[string]any{"case": "O-LAB-2", "change": "visibility narrowed"})
		// The Group role removed: the old Organization-only context's next read is not_found (no label.group.read).
		if err := e.model.RemoveGroupMembership(ctx, org, g1.ID, dave.id); err != nil {
			t.Fatal(err)
		}
		if _, err := read(old); CodeOf(err) != CodeNotFound {
			t.Errorf("after the role removal the old context reads %v, want not_found", err)
		}
		ordering(t, map[string]any{"case": "O-LAB-2", "change": "Group role removed"})
	})

	t.Run("O-LAB-3 a duplicate assignment race: one already_assigned, one row, one audit row", func(t *testing.T) {
		g := newGroup(t, org)
		oc := as(t, owner, org, zeroG)
		def, err := oc.svc.CreateLabelDefinition(oc.rctx, oc.tc, orgScope, name(), string(tenancy.LabelVisibilityOwnerScope))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		id := def.ID.String()
		c1, c2 := as(t, owner, org, zeroG), as(t, admin, org, zeroG)
		// Both take the target Group FOR NO KEY UPDATE first: the duplicate queues behind the first and then meets the
		// committed row (23505 label_assignments_unique → conflict already_assigned).
		e1, e2, inOrder := serialize(t, c1.hook,
			func() error { return c1.svc.AssignLabel(c1.rctx, c1.tc, id, target(g.ID)) },
			func() error { return c2.svc.AssignLabel(c2.rctx, c2.tc, id, target(g.ID)) })
		if e1 != nil || CodeOf(e2) != CodeConflict || field(e2) != "already_assigned" || !inOrder {
			t.Errorf("assign %v, duplicate %v (field %q), in order %t; want the first, then conflict already_assigned", e1, e2, field(e2), inOrder)
		}
		if got := labelState(t, org, id); got != "definitions=1 assignments=1 orphans=0" {
			t.Errorf("%s, want one assignment", got)
		}
		if got := auditOf(t, id); got != "label.assigned=1,label.created=1" {
			t.Errorf("audit %q, want one label.assigned", got)
		}
		ordering(t, map[string]any{"case": "O-LAB-3"})
	})

	t.Run("O-LAB-4 creations racing the 200 bound serialize on the scope row; a mixed pair commits both", func(t *testing.T) {
		o4, owner4 := newOrg(t)
		g4 := newGroup(t, o4)
		ga, gb := member(t, o4, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g4.ID: tenancy.GroupAdmin}),
			member(t, o4, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g4.ID: tenancy.GroupAdmin})
		create := func(c actorCtx, sc Scope) func() error {
			return func() error {
				_, err := c.svc.CreateLabelDefinition(c.rctx, c.tc, sc, name(), string(tenancy.LabelVisibilityOwnerScope))
				return err
			}
		}
		bounded := func(t *testing.T, what string, first, second actorCtx, sc Scope, scopeType string, scopeID [16]byte) {
			t.Helper()
			ef, es, inOrder := serialize(t, first.hook, create(first, sc), create(second, sc))
			if ef != nil || CodeOf(es) != CodeConflict || field(es) != "scope_full" || !inOrder || isDeadlock(ef) || isDeadlock(es) {
				t.Errorf("%s: first %v, second %v (field %q), in order %t; want the 200th, then conflict scope_full, no deadlock",
					what, ef, es, field(es), inOrder)
			}
			if n := scopeRows(t, o4, scopeType, scopeID); n != tenancy.MaxLabelDefinitionsPerScope {
				t.Errorf("%s: %d definitions, want %d", what, n, tenancy.MaxLabelDefinitionsPerScope)
			}
			ordering(t, map[string]any{"case": "O-LAB-4", "pair": what})
		}
		// Organization pair: 199 seeded; both creations lock the Organization row FOR NO KEY UPDATE up front (step 3).
		seed(t, o4, nil, tenancy.MaxLabelDefinitionsPerScope-1)
		bounded(t, "organization", as(t, owner4, o4, zeroG), as(t, owner4, o4, zeroG), orgScope, "organization", o4)
		// Group pair on one Group: 199 seeded; both lock g4 FOR NO KEY UPDATE (step 6) after the Organization FOR SHARE.
		seed(t, o4, &g4.ID, tenancy.MaxLabelDefinitionsPerScope-1)
		bounded(t, "group", as(t, ga, o4, zeroG), as(t, gb, o4, zeroG), groupScope(g4.ID), "group", g4.ID)
		// Mixed pair in an Organization with room in both scopes: the Group creation's Organization FOR SHARE and the
		// Organization creation's FOR NO KEY UPDATE serialize in step order, without a cycle, in either order.
		o5, owner5 := newOrg(t)
		g5 := newGroup(t, o5)
		gc := member(t, o5, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g5.ID: tenancy.GroupAdmin})
		for _, first := range []string{"organization", "group"} {
			oc, gcc := as(t, owner5, o5, zeroG), as(t, gc, o5, zeroG)
			orgCreate, groupCreate := create(oc, orgScope), create(gcc, groupScope(g5.ID))
			var eo, eg error
			if first == "organization" {
				eo, eg, _ = serialize(t, oc.hook, orgCreate, groupCreate)
			} else {
				eg, eo, _ = serialize(t, gcc.hook, groupCreate, orgCreate)
			}
			if eo != nil || eg != nil || isDeadlock(eo) || isDeadlock(eg) {
				t.Errorf("mixed, %s first: Organization %v, Group %v; want both committed without a deadlock", first, eo, eg)
			}
			ordering(t, map[string]any{"case": "O-LAB-4", "pair": "mixed", "first": first})
		}
		if no, ng := scopeRows(t, o5, "organization", o5), scopeRows(t, o5, "group", g5.ID); no != 2 || ng != 2 {
			t.Errorf("mixed pairs left %d Organization and %d Group definitions, want 2 and 2", no, ng)
		}
	})
}
