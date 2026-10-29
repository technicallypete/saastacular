//go:build integration

// G018 B-ONBOARDING-IT (Plan r5 §6.2): TestOnboarding against the access integration environment (real PostgreSQL at
// version 14, the attested runtime role, the real guarded facade). Slice 1 cases: ONB-1 permitted creators (org owner,
// admin and member) with the trigger-written owner binding and T4/T7 attribution; ONB-2 creation denials (selected
// Group context, Organization membership removed after Resolve) with zero rows and an access.denied row persisted
// through the normal runtime path, and a member's new Group granting nothing elsewhere; ONB-4 the same slug twice
// (conflict, field slug) and independently in another Organization; ONB-5 new-profile completion (Group owner flag off
// and on, Organization admin) with T10/T12; ONB-8 a second completion of a complete Group (conflict, no second profile);
// ONB-9 existing-profile selection (attached-Group owner, Organization admin incl. an unlinked profile; unrelated,
// member-only, unlinked and foreign profiles refused, the new Group alone never qualifying); ONB-13 member self-service.
package access

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

type onbWorld struct {
	org, other                                     tenancy.OrganizationID
	owner, admin, member, gowner, gmember, foreign *itUser
	gA, gB                                         tenancy.Group // gA (gowner owns, gmember member) attached to pA; gB owned by gowner, attached to pB
	pA, pB, pU, pF                                 tenancy.BusinessProfile
}

func (e *itEnv) onbWorld(t *testing.T) onbWorld {
	t.Helper()
	w := onbWorld{owner: e.newUser(t), admin: e.newUser(t), member: e.newUser(t), gowner: e.newUser(t), gmember: e.newUser(t), foreign: e.newUser(t)}
	w.org = e.org(t, "Onboarding "+e.canaries[0])
	w.other = e.org(t, "Onboarding Other")
	e.grantOrg(t, w.org, w.owner, tenancy.OrgOwner)
	e.grantOrg(t, w.org, w.admin, tenancy.OrgAdmin)
	for _, u := range []*itUser{w.member, w.gowner, w.gmember} {
		e.grantOrg(t, w.org, u, tenancy.OrgMember)
	}
	n := itSeq.Add(1)
	w.gA = e.group(t, w.org, w.owner, "Onboarding A", fmt.Sprintf("onb-a-%d", n))
	w.gB = e.group(t, w.org, w.owner, "Onboarding B", fmt.Sprintf("onb-b-%d", n))
	e.grantGroup(t, w.org, w.gA.ID, w.gowner, tenancy.GroupOwner)
	e.grantGroup(t, w.org, w.gB.ID, w.gowner, tenancy.GroupOwner)
	e.grantGroup(t, w.org, w.gA.ID, w.gmember, tenancy.GroupMember)
	w.pA, w.pB, w.pU = e.profile(t, w.org, "Profile A"), e.profile(t, w.org, "Profile B"), e.profile(t, w.org, "Profile Unlinked")
	e.link(t, w.org, w.gA.ID, w.pA.ID)
	e.link(t, w.org, w.gB.ID, w.pB.ID)
	e.grantOrg(t, w.other, w.foreign, tenancy.OrgOwner)
	w.pF = e.profile(t, w.other, "Profile Foreign")
	for _, u := range []*itUser{w.owner, w.admin, w.member, w.gowner, w.gmember, w.foreign} {
		e.bind(t, u, w.org)
	}
	e.bind(t, w.foreign, w.other)
	return w
}

// onbOwnerBindings lists the Group-scope bindings of g as subject/role pairs (owner connection).
func (e *itEnv) onbOwnerBindings(t *testing.T, g tenancy.GroupID) []string {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT subject_user_id::text || '/' || role || '/' || source FROM app.role_bindings
WHERE scope_type = 'group' AND group_id = $1::text::uuid ORDER BY subject_user_id`, g.String())
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

// onbEvents lists the audit actions whose resource is id, oldest first, with their actor.
func (e *itEnv) onbEvents(t *testing.T, org tenancy.OrganizationID, id string) []string {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT action || '/' || coalesce(actor_user_id::text, '-') FROM audit.events
WHERE organization_id = $1::text::uuid AND (resource_id::text = $2 OR group_id::text = $2) AND action <> 'access.denied'
ORDER BY occurred_at, id`, org.String(), id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

func (e *itEnv) onbDenied(t *testing.T, org tenancy.OrganizationID, op string) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND operation = $2
AND organization_id IS NOT DISTINCT FROM $1::text::uuid`, org.String(), op).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (e *itEnv) onbCount(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestOnboarding(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	w := e.onbWorld(t)
	newCtx := func(name string) tenancy.BusinessContext { return tenancy.BusinessContext{DisplayName: name} }

	t.Run("ONB-1 permitted creators", func(t *testing.T) {
		for i, u := range []*itUser{w.owner, w.admin, w.member} {
			rctx, tc := e.as(t, u, w.org)
			orgBindings := e.part(t, "organization_memberships", w.org)
			name := fmt.Sprintf("Created Group %d", i)
			got, err := e.svc.CreateGroup(rctx, tc, name, "")
			if err != nil {
				t.Fatalf("creator %d: %v", i, err)
			}
			if got.Complete || got.BusinessProfileID != nil || got.Group.Name != name || got.Group.Slug != fmt.Sprintf("created-group-%d", i) ||
				got.Group.OrganizationID != w.org || zero(got.Group.ID) {
				t.Errorf("creator %d: %+v", i, got)
			}
			want := []string{u.id.String() + "/group_owner/membership"}
			if b := e.onbOwnerBindings(t, got.Group.ID); fmt.Sprint(b) != fmt.Sprint(want) {
				t.Errorf("creator %d bindings %v, want %v", i, b, want)
			}
			if ev := e.onbEvents(t, w.org, got.Group.ID.String()); len(ev) != 2 || ev[0] != "group.created/"+u.id.String() ||
				ev[1] != "membership.granted/"+u.id.String() {
				t.Errorf("creator %d audit %v, want group.created then membership.granted by the creator", i, ev)
			}
			if e.part(t, "organization_memberships", w.org) != orgBindings {
				t.Errorf("creator %d: an Organization binding changed", i)
			}
			if g, err := e.svc.GetGroupOnboarding(rctx, tc, got.Group.ID); err != nil || g.Complete || g.Group.ID != got.Group.ID {
				t.Errorf("creator %d read-back: %+v %v", i, g, err)
			}
		}
	})

	t.Run("ONB-2 creation denials persist access.denied and change nothing", func(t *testing.T) {
		n0 := e.onbDenied(t, w.org, opCreateGroup)
		// Selected-Group context: the Group owner of gA in gA never creates (forbidden, role_not_permitted).
		rctx, _ := beginAs(t, e.svc, ctx, w.gowner, w.org)
		stc := resolve(t, e.svc, rctx, w.org, w.gA.ID)
		_, err := e.svc.CreateGroup(rctx, stc, "Selected Context", "selected-context")
		wantCode(t, "selected-Group context", err, CodeForbidden)
		if r := e.lastReason(t); r != string(reasonRoleNotPermitted) {
			t.Errorf("selected-Group reason %q", r)
		}
		// The Organization binding removed after Resolve (a Group role alone remains): context_denied, recorded.
		u := e.newUser(t)
		e.grantOrg(t, w.org, u, tenancy.OrgMember)
		e.grantGroup(t, w.org, w.gB.ID, u, tenancy.GroupOwner)
		e.bind(t, u, w.org)
		uctx, utc := e.as(t, u, w.org)
		if _, err := e.setup.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND scope_type = 'organization'`, u.id.String()); err != nil {
			t.Fatal(err)
		}
		before2 := e.state(t, w.org, w.other)
		_, err = e.svc.CreateGroup(uctx, utc, "Group Role Only", "group-role-only")
		wantCode(t, "Group role without Organization binding", err, CodeContextDenied)
		if e.state(t, w.org, w.other) != before2 {
			t.Error("a denied creation changed state")
		}
		if n := e.onbDenied(t, w.org, opCreateGroup); n != n0+1 {
			t.Errorf("access.denied group.create rows in the Organization %d, want %d (the selected-Group refusal)", n, n0+1)
		}
		// The context refusal is recorded without an established Organization (IDENTITY shape), attributed to u.
		if n := e.onbCount(t, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND operation = 'group.create'
AND reason = 'organization_not_permitted' AND organization_id IS NULL AND actor_user_id = $1::text::uuid`, u.id.String()); n != 1 {
			t.Errorf("access.denied group.create rows of the Group-role-only subject %d, want 1", n)
		}
		// Input refusals precede the transaction: no row, no denial.
		mctx, mtc := e.as(t, w.member, w.org)
		_, err = e.svc.CreateGroup(mctx, mtc, "", "")
		wantCode(t, "empty name", err, CodeInvalidInput)
		_, err = e.svc.CreateGroup(mctx, mtc, "Reserved", "new")
		wantCode(t, "reserved slug", err, CodeInvalidInput)
	})

	t.Run("ONB-4 slug conflict and cross-Organization independence", func(t *testing.T) {
		rctx, tc := e.as(t, w.admin, w.org)
		slug := fmt.Sprintf("dup-%d", itSeq.Add(1))
		if _, err := e.svc.CreateGroup(rctx, tc, "Dup One", slug); err != nil {
			t.Fatal(err)
		}
		before := e.state(t, w.org)
		_, err := e.svc.CreateGroup(rctx, tc, "Dup Two", slug)
		var ae *Error
		if CodeOf(err) != CodeConflict || !asError(err, &ae) || ae.Field != "slug" {
			t.Errorf("duplicate slug: %v", err)
		}
		if e.state(t, w.org) != before {
			t.Error("a conflicting creation changed state")
		}
		fctx, ftc := e.as(t, w.foreign, w.other)
		if _, err := e.svc.CreateGroup(fctx, ftc, "Dup Elsewhere", slug); err != nil {
			t.Errorf("same slug in another Organization: %v", err)
		}
	})

	t.Run("ONB-5 new-profile completion", func(t *testing.T) {
		for _, c := range []struct {
			name string
			u    *itUser
			flag bool
		}{{"Group owner flag off", w.member, false}, {"Group owner flag on", w.member, true}, {"Organization admin", w.admin, true}} {
			e.setFlag(t, w.org, c.flag)
			rctx, tc := e.as(t, c.u, w.org)
			g, err := e.svc.CreateGroup(rctx, tc, "Complete "+c.name, "")
			if err != nil {
				t.Fatalf("%s create: %v", c.name, err)
			}
			got, err := e.svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g.Group.ID, newCtx("New "+c.name))
			if err != nil || !got.Complete || got.BusinessProfileID == nil {
				t.Fatalf("%s complete: %+v %v", c.name, got, err)
			}
			p := *got.BusinessProfileID
			if n := e.onbCount(t, `SELECT count(*) FROM app.group_business_profiles WHERE group_id = $1::text::uuid AND business_profile_id = $2::text::uuid`,
				g.Group.ID.String(), p.String()); n != 1 {
				t.Errorf("%s association rows %d", c.name, n)
			}
			if ev := e.onbEvents(t, w.org, p.String()); len(ev) != 2 || ev[0] != "business_profile.created/"+c.u.id.String() ||
				ev[1] != "business_profile.linked/"+c.u.id.String() {
				t.Errorf("%s audit %v, want business_profile.created then business_profile.linked", c.name, ev)
			}
			if r, err := e.svc.GetGroupOnboarding(rctx, tc, g.Group.ID); err != nil || !r.Complete || *r.BusinessProfileID != p {
				t.Errorf("%s read-back %+v %v", c.name, r, err)
			}
			if bp, err := e.svc.GetBusinessProfile(rctx, tc, p); err != nil || bp.Context.DisplayName != "New "+c.name {
				t.Errorf("%s profile read %+v %v", c.name, bp, err)
			}
		}
		e.setFlag(t, w.org, true)
	})

	t.Run("ONB-8 a complete Group answers conflict and gains nothing", func(t *testing.T) {
		rctx, tc := e.as(t, w.gowner, w.org)
		g, err := e.svc.CreateGroup(rctx, tc, "Once Only", "")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g.Group.ID, newCtx("First")); err != nil {
			t.Fatal(err)
		}
		before := e.state(t, w.org)
		_, err = e.svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g.Group.ID, newCtx("Second"))
		var ae *Error
		if CodeOf(err) != CodeConflict || !asError(err, &ae) || ae.Field != string(tenancy.CodeAssociationExists) {
			t.Errorf("second new completion: %v", err)
		}
		_, err = e.svc.CompleteGroupOnboardingWithProfile(rctx, tc, g.Group.ID, w.pA.ID)
		wantCode(t, "existing after complete", err, CodeConflict)
		if e.state(t, w.org) != before {
			t.Error("a conflicting completion changed state")
		}
	})

	t.Run("ONB-9 existing-profile selection", func(t *testing.T) {
		create := func(u *itUser) (context.Context, *TenantContext, tenancy.GroupID) {
			rctx, tc := e.as(t, u, w.org)
			g, err := e.svc.CreateGroup(rctx, tc, fmt.Sprintf("Select %d", itSeq.Add(1)), "")
			if err != nil {
				t.Fatal(err)
			}
			return rctx, tc, g.Group.ID
		}
		// Refused with zero effect: unlinked (not readable: not_found), foreign (not_found), member-only pA for gmember
		// who owns a new Group (readable: forbidden), and the new Group alone qualifying nothing (pB for a fresh member).
		n0 := e.onbDenied(t, w.org, opCompleteOnboardingExisting)
		gctx, gtc, gg := create(w.gowner)
		wantCode(t, "unlinked pU for a Group owner", func() error { _, err := e.svc.CompleteGroupOnboardingWithProfile(gctx, gtc, gg, w.pU.ID); return err }(), CodeNotFound)
		wantCode(t, "foreign pF", func() error { _, err := e.svc.CompleteGroupOnboardingWithProfile(gctx, gtc, gg, w.pF.ID); return err }(), CodeNotFound)
		mctx, mtc, mg := create(w.gmember)
		wantCode(t, "member-only pA", func() error { _, err := e.svc.CompleteGroupOnboardingWithProfile(mctx, mtc, mg, w.pA.ID); return err }(), CodeForbidden)
		if r := e.lastReason(t); r != string(reasonNoLinkedPath) {
			t.Errorf("member-only reason %q, want no_linked_path", r)
		}
		nctx, ntc, ng := create(w.member)
		wantCode(t, "new Group alone, pB", func() error { _, err := e.svc.CompleteGroupOnboardingWithProfile(nctx, ntc, ng, w.pB.ID); return err }(), CodeNotFound)
		for _, x := range []tenancy.GroupID{gg, mg, ng} {
			if n := e.onbCount(t, `SELECT count(*) FROM app.group_business_profiles WHERE group_id = $1::text::uuid`, x.String()); n != 0 {
				t.Errorf("a refused selection attached a profile (%d rows)", n)
			}
		}
		if n := e.onbDenied(t, w.org, opCompleteOnboardingExisting); n != n0+4 {
			t.Errorf("access.denied complete_existing rows %d, want %d", n, n0+4)
		}
		// Allowed: the owner of attached gA selects pA; the Organization admin selects the unlinked pU only through the
		// Group-management override (G060 r0 A4/D-5: off, the readable pU is refused with nothing attached).
		rctx, tc, g := create(w.gowner)
		if got, err := e.svc.CompleteGroupOnboardingWithProfile(rctx, tc, g, w.pA.ID); err != nil || !got.Complete || *got.BusinessProfileID != w.pA.ID {
			t.Errorf("Group owner selects pA: %+v %v", got, err)
		}
		actx, atc, ag := create(w.admin)
		wantCode(t, "Organization admin selects unlinked pU, override off", func() error {
			_, err := e.svc.CompleteGroupOnboardingWithProfile(actx, atc, ag, w.pU.ID)
			return err
		}(), CodeForbidden)
		if n := e.onbCount(t, `SELECT count(*) FROM app.group_business_profiles WHERE group_id = $1::text::uuid`, ag.String()); n != 0 {
			t.Errorf("a refused selection attached a profile (%d rows)", n)
		}
		e.withOverride(t, w.owner, w.org)
		actx, atc = e.as(t, w.admin, w.org)
		if _, err := e.svc.CompleteGroupOnboardingWithProfile(actx, atc, ag, w.pU.ID); err != nil {
			t.Errorf("Organization admin selects unlinked pU, override on: %v", err)
		}
		// Another member's Group: the member cannot read or complete it.
		wantCode(t, "another member's Group read", func() error { _, err := e.svc.GetGroupOnboarding(nctx, ntc, mg); return err }(), CodeNotFound)
		wantCode(t, "another member's Group complete", func() error {
			_, err := e.svc.CompleteGroupOnboardingWithNewProfile(nctx, ntc, mg, newCtx("Hijack"))
			return err
		}(), CodeNotFound)
	})

	t.Run("ONB-13 member self-service", func(t *testing.T) {
		u := e.newUser(t)
		e.grantOrg(t, w.org, u, tenancy.OrgMember)
		e.bind(t, u, w.org)
		rctx, tc := e.as(t, u, w.org)
		g1, err := e.svc.CreateGroup(rctx, tc, "Member One", "")
		if err != nil {
			t.Fatal(err)
		}
		done, err := e.svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g1.Group.ID, newCtx("Member Profile"))
		if err != nil {
			t.Fatal(err)
		}
		// The member now owns g1 attached to its own new profile: that profile qualifies for its next Group.
		g2, err := e.svc.CreateGroup(rctx, tc, "Member Two", "")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.svc.CompleteGroupOnboardingWithProfile(rctx, tc, g2.Group.ID, *done.BusinessProfileID); err != nil {
			t.Errorf("member selects its own attached profile: %v", err)
		}
		g3, err := e.svc.CreateGroup(rctx, tc, "Member Three", "")
		if err != nil {
			t.Fatal(err)
		}
		wantCode(t, "member selects pA without a path", func() error {
			_, err := e.svc.CompleteGroupOnboardingWithProfile(rctx, tc, g3.Group.ID, w.pA.ID)
			return err
		}(), CodeNotFound)
	})

	t.Run("ONB-10 paged lists traverse completely and decide every row", func(t *testing.T) {
		org := e.org(t, "Onboarding Lists "+e.canaries[0])
		owner, gown, mem := e.newUser(t), e.newUser(t), e.newUser(t)
		e.grantOrg(t, org, owner, tenancy.OrgOwner)
		e.grantOrg(t, org, gown, tenancy.OrgMember)
		e.grantOrg(t, org, mem, tenancy.OrgMember)
		for _, u := range []*itUser{owner, gown, mem} {
			e.bind(t, u, org)
		}
		n := itSeq.Add(1)
		var gs []tenancy.Group
		for i := range 7 {
			gs = append(gs, e.group(t, org, owner, fmt.Sprintf("List %d", i), fmt.Sprintf("onb-list-%d-%d", n, i)))
		}
		e.grantGroup(t, org, gs[0].ID, mem, tenancy.GroupMember)
		e.grantGroup(t, org, gs[1].ID, gown, tenancy.GroupOwner)
		e.grantGroup(t, org, gs[0].ID, gown, tenancy.GroupMember) // gown can read ps[0] (member-only) but not select it
		var ps []tenancy.BusinessProfile
		for i := range 7 {
			ps = append(ps, e.profile(t, org, fmt.Sprintf("List Profile %d", i)))
		}
		e.link(t, org, gs[0].ID, ps[0].ID) // member-only for mem
		e.link(t, org, gs[1].ID, ps[1].ID) // gown's attached profile

		groups := func(u *itUser, limit int) (ids []string, complete map[string]bool) {
			t.Helper()
			rctx, tc := e.as(t, u, org)
			complete = map[string]bool{}
			after := ""
			for pages := 0; ; pages++ {
				p, err := e.svc.ListGroups(rctx, tc, after, limit, "")
				if err != nil || len(p.Groups) > limit || pages > 10 {
					t.Fatalf("ListGroups page %d: %v (%d rows)", pages, err, len(p.Groups))
				}
				for _, g := range p.Groups {
					if len(ids) > 0 && g.Group.ID.String() <= ids[len(ids)-1] {
						t.Fatalf("not strictly ascending: %s after %s", g.Group.ID, ids[len(ids)-1])
					}
					ids = append(ids, g.Group.ID.String())
					complete[g.Group.ID.String()] = g.Complete && g.BusinessProfileID != nil
				}
				if p.Next == nil {
					return ids, complete
				}
				after = p.Next.String()
			}
		}
		// hints records each listed row's can_select/can_update hints of the last profiles call (G060 r1 C2).
		hints := map[string][2]bool{}
		profiles := func(u *itUser, limit int) (ids []string, pages int) {
			t.Helper()
			rctx, tc := e.as(t, u, org)
			after := ""
			clear(hints)
			for ; ; pages++ {
				p, err := e.svc.ListBusinessProfiles(rctx, tc, after, limit)
				if err != nil || len(p.Profiles) > limit || pages > 10 {
					t.Fatalf("ListBusinessProfiles page %d: %v", pages, err)
				}
				for _, x := range p.Profiles {
					ids = append(ids, x.ID.String())
					hints[x.ID.String()] = [2]bool{x.CanSelect, x.CanUpdate}
				}
				if p.Next == nil {
					return ids, pages + 1
				}
				after = p.Next.String()
			}
		}
		var allG, allP []string
		for _, g := range gs {
			allG = append(allG, g.ID.String())
		}
		for _, p := range ps {
			allP = append(allP, p.ID.String())
		}
		slices.Sort(allG)
		slices.Sort(allP)

		ids, complete := groups(owner, 3)
		if !slices.Equal(ids, allG) {
			t.Errorf("owner Groups %v, want every Group of the Organization %v", ids, allG)
		}
		for i, g := range gs {
			if complete[g.ID.String()] != (i < 2) {
				t.Errorf("Group %d complete=%v", i, complete[g.ID.String()])
			}
		}
		if ids, _ := groups(mem, 3); !slices.Equal(ids, []string{gs[0].ID.String()}) {
			t.Errorf("member Groups %v, want only its own", ids)
		}
		// G060 r1 C2: the list is the readable metadata (profile.read) with independent can_select/can_update hints. The
		// owner (no override) reads every profile; it created every Group, so its Group-owner path selects and updates the
		// two linked profiles (flag on) and nothing unlinked; with the override every row is selectable and updatable.
		psID := func(i int) string { return ps[i].ID.String() }
		checkHints := func(who string, want map[string][2]bool) {
			t.Helper()
			for id, h := range want {
				if got, ok := hints[id]; !ok || got != h {
					t.Errorf("%s hints of %s: %v (listed %t), want can_select/can_update %v", who, id, got, ok, h)
				}
			}
		}
		if ids, pages := profiles(owner, 3); !slices.Equal(ids, allP) || pages != 3 {
			t.Errorf("owner profiles %v in %d pages, want every profile in 3", ids, pages)
		}
		ownerOff := map[string][2]bool{psID(0): {true, true}, psID(1): {true, true}}
		for i := 2; i < len(ps); i++ {
			ownerOff[psID(i)] = [2]bool{false, false}
		}
		checkHints("owner, override off", ownerOff)
		off := e.withOverride(t, owner, org)
		if ids, pages := profiles(owner, 3); !slices.Equal(ids, allP) || pages != 3 {
			t.Errorf("owner profiles with the override %v in %d pages, want every profile in 3", ids, pages)
		}
		ownerOn := map[string][2]bool{}
		for i := range ps {
			ownerOn[psID(i)] = [2]bool{true, true}
		}
		checkHints("owner, override on", ownerOn)
		off()
		// gown: its two readable profiles (ps[0] through its gs[0] membership, ps[1] through its owned gs[1]), one per
		// page at limit 1 (row-level security admits exactly those two rows: two pages); only ps[1] is selectable and
		// updatable. mem: the member-only ps[0] is listed read-only (neither selectable nor updatable).
		readable := []string{psID(0), psID(1)}
		slices.Sort(readable)
		if ids, pages := profiles(gown, 1); !slices.Equal(ids, readable) || pages != 2 {
			t.Errorf("Group owner profiles %v in %d pages, want %v in 2", ids, pages, readable)
		}
		checkHints("Group owner", map[string][2]bool{psID(0): {false, false}, psID(1): {true, true}})
		if ids, _ := profiles(mem, 3); !slices.Equal(ids, []string{psID(0)}) {
			t.Errorf("member profiles %v, want only the member-only %s", ids, psID(0))
		}
		checkHints("member", map[string][2]bool{psID(0): {false, false}})

		rctx, tc := e.as(t, owner, org)
		for _, c := range []struct {
			after string
			limit int
			field string
		}{{"", 0, "invalid_page_size"}, {"", 101, "invalid_page_size"}, {"x", 10, "invalid_cursor"}} {
			_, err := e.svc.ListGroups(rctx, tc, c.after, c.limit, "")
			var ae *Error
			if !asError(err, &ae) || ae.Code != CodeInvalidInput || ae.Field != c.field {
				t.Errorf("ListGroups %q/%d: %v", c.after, c.limit, err)
			}
			_, err = e.svc.ListBusinessProfiles(rctx, tc, c.after, c.limit)
			if !asError(err, &ae) || ae.Code != CodeInvalidInput || ae.Field != c.field {
				t.Errorf("ListBusinessProfiles %q/%d: %v", c.after, c.limit, err)
			}
		}
		// The exact slug filter: the owner finds the Group, a member without access finds nothing (no existence signal).
		if p, err := e.svc.ListGroups(rctx, tc, "", 10, gs[3].Slug); err != nil || len(p.Groups) != 1 || p.Groups[0].Group.ID != gs[3].ID || p.Next != nil {
			t.Errorf("owner slug filter: %+v %v", p, err)
		}
		mctx, mtc := e.as(t, mem, org)
		if p, err := e.svc.ListGroups(mctx, mtc, "", 10, gs[3].Slug); err != nil || len(p.Groups) != 0 {
			t.Errorf("member slug filter on another Group: %+v %v", p, err)
		}
		// Selected-Group context restricts both lists to the selected Group (Plan r5 §1.5): the owner selected in gs[0]
		// lists only gs[0] (complete), nothing after it, and the slug filter of another Group finds nothing; gown selected
		// in its owned gs[1] lists gs[1] and only ps[1] (its other Organization-only candidate gs[0] is out of scope);
		// gown selected in gs[0] (member there) lists gs[0] and no selectable profile.
		selected := func(u *itUser, g tenancy.GroupID) (context.Context, *TenantContext) {
			t.Helper()
			sctx, _ := beginAs(t, e.svc, ctx, u, org) // beginAs registers its own cleanup
			return sctx, resolve(t, e.svc, sctx, org, g)
		}
		gid := func(p OnboardingGroupPage) []string {
			out := []string{}
			for _, g := range p.Groups {
				out = append(out, g.Group.ID.String())
			}
			return out
		}
		pid := func(p ProfilePage) []string {
			out := []string{}
			for _, x := range p.Profiles {
				out = append(out, x.ID.String())
			}
			return out
		}
		sctx, stc := selected(owner, gs[0].ID)
		if p, err := e.svc.ListGroups(sctx, stc, "", 10, ""); err != nil || !slices.Equal(gid(p), []string{gs[0].ID.String()}) || !p.Groups[0].Complete || p.Next != nil {
			t.Errorf("owner selected gs[0] Groups %v next %v: %v", gid(p), p.Next, err)
		}
		if p, err := e.svc.ListGroups(sctx, stc, gs[0].ID.String(), 10, ""); err != nil || len(p.Groups) != 0 || p.Next != nil {
			t.Errorf("owner selected gs[0] Groups after gs[0] %v: %v", gid(p), err)
		}
		if p, err := e.svc.ListGroups(sctx, stc, "", 10, gs[3].Slug); err != nil || len(p.Groups) != 0 {
			t.Errorf("owner selected gs[0] slug filter on gs[3] %v: %v", gid(p), err)
		}
		sctx, stc = selected(gown, gs[1].ID)
		if p, err := e.svc.ListGroups(sctx, stc, "", 10, ""); err != nil || !slices.Equal(gid(p), []string{gs[1].ID.String()}) {
			t.Errorf("Group owner selected gs[1] Groups %v: %v", gid(p), err)
		}
		if p, err := e.svc.ListBusinessProfiles(sctx, stc, "", 10); err != nil || !slices.Equal(pid(p), []string{ps[1].ID.String()}) {
			t.Errorf("Group owner selected gs[1] profiles %v: %v", pid(p), err)
		}
		sctx, stc = selected(gown, gs[0].ID)
		if p, err := e.svc.ListGroups(sctx, stc, "", 10, ""); err != nil || !slices.Equal(gid(p), []string{gs[0].ID.String()}) {
			t.Errorf("Group owner selected gs[0] Groups %v: %v", gid(p), err)
		}
		// G060 r1 C2: selected gs[0], where gown is a member, lists the linked ps[0] read-only (the selected context
		// limits the readable path to gs[0]; neither hint is set).
		if p, err := e.svc.ListBusinessProfiles(sctx, stc, "", 10); err != nil || !slices.Equal(pid(p), []string{ps[0].ID.String()}) ||
			p.Profiles[0].CanSelect || p.Profiles[0].CanUpdate {
			t.Errorf("Group owner selected gs[0] (member there) profiles %v %+v: %v", pid(p), p.Profiles, err)
		}
		// Another Organization's rows never appear (w.org's Groups and profiles are absent from this traversal).
		if slices.Contains(ids, w.gA.ID.String()) {
			t.Error("a foreign Group was listed")
		}
	})

	t.Run("ONB-11 retry after an unknown outcome is an authorized conflict", func(t *testing.T) {
		rctx, tc := e.as(t, w.member, w.org)
		n := itSeq.Add(1)
		slug := fmt.Sprintf("onb-retry-%d", n)
		g, err := e.svc.CreateGroup(rctx, tc, "Retry", slug)
		if err != nil {
			t.Fatal(err)
		}
		// The retried create (same payload, or a changed name) is conflict on the slug, never a second Group.
		for _, name := range []string{"Retry", "Retry Changed"} {
			_, err := e.svc.CreateGroup(rctx, tc, name, slug)
			var ae *Error
			if !asError(err, &ae) || ae.Code != CodeConflict || ae.Field != "slug" {
				t.Errorf("retried create %q: %v", name, err)
			}
		}
		p, err := e.svc.ListGroups(rctx, tc, "", 10, slug)
		if err != nil || len(p.Groups) != 1 || p.Groups[0].Group.ID != g.Group.ID || p.Groups[0].Group.Name != "Retry" || p.Groups[0].Complete {
			t.Errorf("slug recovery: %+v %v", p, err)
		}
		done, err := e.svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g.Group.ID, newCtx("Retry Profile"))
		if err != nil {
			t.Fatal(err)
		}
		before := e.state(t, w.org)
		_, err = e.svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g.Group.ID, newCtx("Retry Profile Changed"))
		wantCode(t, "retried completion", err, CodeConflict)
		got, err := e.svc.GetGroupOnboarding(rctx, tc, g.Group.ID)
		if err != nil || !got.Complete || *got.BusinessProfileID != *done.BusinessProfileID {
			t.Errorf("committed profile after the retry: %+v %v", got, err)
		}
		if e.state(t, w.org) != before {
			t.Error("a retried completion changed state")
		}
	})

	t.Run("ONB-12 the can_update hint is display only", func(t *testing.T) {
		// G060 (r0 A4/D-5): the Organization path of profile.update needs the Group-management override, independent of
		// the flag (the owner's own gA ownership is a Group path, so the flag decides it while the override is off).
		var off func()
		for _, c := range []struct {
			name string
			u    *itUser
			flag bool
			ovr  bool
			want bool
		}{
			{"Organization owner, override off", w.owner, false, false, false}, {"Organization admin, override off", w.admin, false, false, false},
			{"Group owner flag off", w.gowner, false, false, false}, {"Group owner flag on", w.gowner, true, false, true},
			{"Group member flag on", w.gmember, true, false, false},
			{"Organization owner, override on", w.owner, false, true, true}, {"Organization admin, override on", w.admin, false, true, true},
			{"Group owner flag off, override on", w.gowner, false, true, false},
		} {
			if c.ovr && off == nil {
				off = e.withOverride(t, w.owner, w.org)
			}
			e.setFlag(t, w.org, c.flag)
			rctx, tc := e.as(t, c.u, w.org)
			p, can, err := e.svc.GetBusinessProfileView(rctx, tc, w.pA.ID)
			if err != nil || p.ID != w.pA.ID || can != c.want {
				t.Errorf("%s: can_update %v (%v), want %v", c.name, can, err, c.want)
			}
		}
		if off != nil {
			off()
		}
		// Hint true, then the Group binding is revoked before the write: the update re-decides and is refused.
		e.setFlag(t, w.org, true)
		u := e.newUser(t)
		e.grantOrg(t, w.org, u, tenancy.OrgMember)
		e.grantGroup(t, w.org, w.gA.ID, u, tenancy.GroupOwner)
		e.bind(t, u, w.org)
		rctx, tc := e.as(t, u, w.org)
		if _, can, err := e.svc.GetBusinessProfileView(rctx, tc, w.pA.ID); err != nil || !can {
			t.Fatalf("hint before the revocation: %v %v", can, err)
		}
		if _, err := e.setup.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND scope_type = 'group'`, u.id.String()); err != nil {
			t.Fatal(err)
		}
		before := e.state(t, w.org)
		_, err := e.svc.UpdateBusinessProfile(rctx, tc, w.pA.ID, newCtx("Hijacked"))
		if c := CodeOf(err); c != CodeForbidden && c != CodeNotFound {
			t.Errorf("update after the revocation: %v", err)
		}
		if e.state(t, w.org) != before {
			t.Error("a refused update changed state")
		}
		e.setFlag(t, w.org, false)
	})

	t.Run("ONB-3 creation atomicity and the deterministic bootstrap race", func(t *testing.T) {
		// (a) A failure injected after the Group statement (S-G1) and before COMMIT: neither the Group nor its owner binding.
		s, h := e.service(e.id)
		rctx, _ := beginAs(t, s, ctx, w.member, w.org)
		tc := resolve(t, s, rctx, w.org, tenancy.GroupID{})
		h.on("after", "IG", func() { h.setFail("Gt") })
		before := e.state(t, w.org)
		_, err := s.CreateGroup(rctx, tc, "Faulted Create", fmt.Sprintf("faulted-%d", itSeq.Add(1)))
		wantCode(t, "fault after S-G1", err, CodeUnavailable)
		if e.state(t, w.org) != before {
			t.Error("a faulted creation left a row")
		}
		if _, stmts := h.counts(); !slices.Contains(stmts, "IG") {
			t.Errorf("the fault did not follow S-G1: %v", stmts)
		}
		// (b) Two creators interleaved at the statement seam: the first pauses after its keyed insert (uncommitted Group
		// and trigger-written binding), the second runs. Different slugs: both commit. Same slug: the second waits on the
		// unique key, then answers conflict. At every observation point each committed Group has exactly one owner binding.
		ownerless := func() int {
			return e.onbCount(t, `SELECT count(*) FROM app.groups g WHERE g.organization_id = $1::text::uuid AND NOT EXISTS (
SELECT 1 FROM app.role_bindings b WHERE b.scope_type = 'group' AND b.group_id = g.id AND b.role = 'group_owner')`, w.org.String())
		}
		for _, same := range []bool{false, true} {
			s1, h1 := e.service(e.id)
			s2, _ := e.service(e.id)
			c1, _ := beginAs(t, s1, ctx, w.member, w.org)
			t1 := resolve(t, s1, c1, w.org, tenancy.GroupID{})
			c2, _ := beginAs(t, s2, ctx, w.admin, w.org)
			t2 := resolve(t, s2, c2, w.org, tenancy.GroupID{})
			slug1 := fmt.Sprintf("race-%d", itSeq.Add(1))
			slug2 := slug1
			if !same {
				slug2 = fmt.Sprintf("race-%d", itSeq.Add(1))
			}
			p := newPause()
			h1.on("after", "IG", p.hook)
			var g1, g2 GroupOnboarding
			a := start(func() error { var err error; g1, err = s1.CreateGroup(c1, t1, "Race One", slug1); return err })
			p.wait(t, "first creator")
			if n := e.onbCount(t, `SELECT count(*) FROM app.groups WHERE organization_id = $1::text::uuid AND slug = $2`, w.org.String(), slug1); n != 0 || ownerless() != 0 {
				t.Errorf("same=%v: an uncommitted Group is observable (%d) or a Group lacks its owner", same, n)
			}
			b := start(func() error { var err error; g2, err = s2.CreateGroup(c2, t2, "Race Two", slug2); return err })
			if same {
				e.waitWaiters(t, 1)
			} else if err := b.wait(t, "second creator"); err != nil {
				t.Errorf("different slugs: second creator %v", err)
			}
			if ownerless() != 0 {
				t.Errorf("same=%v: a committed Group without its owner binding", same)
			}
			close(p.release)
			if err := a.wait(t, "first creator"); err != nil {
				t.Errorf("same=%v: first creator %v", same, err)
			}
			err2 := b.wait(t, "second creator")
			if same {
				var ae *Error
				if CodeOf(err2) != CodeConflict || !asError(err2, &ae) || ae.Field != "slug" {
					t.Errorf("same slug: second creator %v, want conflict on slug", err2)
				}
			} else if want := []string{w.admin.id.String() + "/group_owner/membership"}; fmt.Sprint(e.onbOwnerBindings(t, g2.Group.ID)) != fmt.Sprint(want) {
				t.Errorf("second Group bindings %v", e.onbOwnerBindings(t, g2.Group.ID))
			}
			if want := []string{w.member.id.String() + "/group_owner/membership"}; fmt.Sprint(e.onbOwnerBindings(t, g1.Group.ID)) != fmt.Sprint(want) {
				t.Errorf("same=%v: first Group bindings %v", same, e.onbOwnerBindings(t, g1.Group.ID))
			}
			if ownerless() != 0 {
				t.Errorf("same=%v: a Group without its owner binding after both commits", same)
			}
		}
	})

	t.Run("ONB-6 completion faults leave no orphan profile", func(t *testing.T) {
		rctx, tc := e.as(t, w.member, w.org)
		g, err := e.svc.CreateGroup(rctx, tc, "Fault Target", "")
		if err != nil {
			t.Fatal(err)
		}
		profiles := func() int {
			return e.onbCount(t, `SELECT count(*) FROM app.business_profiles WHERE organization_id = $1::text::uuid`, w.org.String())
		}
		incomplete := func(what string) {
			if n := e.onbCount(t, `SELECT count(*) FROM app.group_business_profiles WHERE group_id = $1::text::uuid`, g.Group.ID.String()); n != 0 {
				t.Errorf("%s: the Group gained an association", what)
			}
		}
		// Validation: zero writes, no transaction.
		s, h := e.service(e.id)
		sctx, _ := beginAs(t, s, ctx, w.member, w.org)
		stc := resolve(t, s, sctx, w.org, tenancy.GroupID{})
		h.reset()
		_, err = s.CompleteGroupOnboardingWithNewProfile(sctx, stc, g.Group.ID, newCtx(""))
		wantCode(t, "invalid context", err, CodeInvalidInput)
		if n, _ := h.counts(); n != 0 {
			t.Errorf("validation began %d transactions", n)
		}
		// A failure at S-P2 (the association statement) after S-P1 succeeded: the profile does not survive.
		n0 := profiles()
		h.setFail("IA")
		_, err = s.CompleteGroupOnboardingWithNewProfile(sctx, stc, g.Group.ID, newCtx("Fault At Link"))
		wantCode(t, "fault at S-P2", err, CodeUnavailable)
		h.setFail("")
		if _, stmts := h.counts(); !slices.Contains(stmts, "IP") {
			t.Errorf("S-P1 did not run before the fault: %v", stmts)
		}
		if profiles() != n0 {
			t.Error("an orphan profile survived the S-P2 fault")
		}
		incomplete("S-P2 fault")
		// The strict audit path: a planted fault refusing only the business_profile.linked row rolls everything back.
		name := fmt.Sprintf("onb_audit_fault_%d", itSeq.Add(1))
		if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events ADD CONSTRAINT `+name+` CHECK (action <> 'business_profile.linked') NOT VALID`); err != nil {
			t.Fatal(err)
		}
		_, err = e.svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g.Group.ID, newCtx("Fault At Audit"))
		if _, derr := e.setup.Exec(ctx, `ALTER TABLE audit.events DROP CONSTRAINT `+name); derr != nil {
			t.Fatal(derr)
		}
		if err == nil {
			t.Error("completion under an audit fault succeeded")
		}
		if profiles() != n0 {
			t.Error("an orphan profile survived the audit fault")
		}
		incomplete("audit fault")
		// Afterwards the same completion succeeds once.
		if got, err := e.svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g.Group.ID, newCtx("After Faults")); err != nil || !got.Complete {
			t.Errorf("completion after the faults: %+v %v", got, err)
		}
	})

	t.Run("ONB-2d fixed-actor runtime SQL cannot bootstrap or misplace a Group", func(t *testing.T) {
		rt := e.rlsRuntime(t)
		rctx, tc := e.as(t, w.member, w.org)
		own, err := e.svc.CreateGroup(rctx, tc, "Member Existing", "")
		if err != nil {
			t.Fatal(err)
		}
		foreignG := e.group(t, w.other, w.foreign, "Foreign Group", fmt.Sprintf("foreign-%d", itSeq.Add(1)))
		before := e.state(t, w.org, w.other)
		for _, c := range []struct {
			who string
			u   *itUser
		}{{"owner", w.owner}, {"member", w.member}} {
			uid, oid := c.u.id.String(), w.org.String()
			rlsTx(t, rt, ptr(uid), ptr(oid), ptr(""), nil, func(tx pgx.Tx) {
				for _, b := range []struct {
					what string
					org  string
					g    tenancy.GroupID
				}{{"another owner's Group", oid, w.gA.ID}, {"its own existing Group", oid, own.Group.ID}, {"a Group of another Organization", w.other.String(), foreignG.ID}} {
					st, n := rlsStmt(t, tx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'group', $3::text::uuid, $3::text::uuid, 'group_owner', 'membership')`, b.org, uid, b.g.String())
					if st != "42501" || n != 0 {
						t.Errorf("%s: owner binding on %s: SQLSTATE %q rows %d, want 42501", c.who, b.what, st, n)
					}
				}
				st, _ := rlsStmt(t, tx, `INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1::text::uuid, $2::text::uuid, 'Foreign Placement', $3)`,
					rlsUUID(t), w.other.String(), fmt.Sprintf("misplaced-%d", itSeq.Add(1)))
				if st != "42501" {
					t.Errorf("%s: Group insert into another Organization: SQLSTATE %q, want 42501", c.who, st)
				}
			})
		}
		// Selected-Group context (the Group owner of gA in gA) and a Group-role-only subject (Organization binding gone).
		ins := func(tx pgx.Tx) string {
			st, _ := rlsStmt(t, tx, `INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1::text::uuid, $2::text::uuid, 'Refused', $3)`,
				rlsUUID(t), w.org.String(), fmt.Sprintf("refused-%d", itSeq.Add(1)))
			return st
		}
		rlsTx(t, rt, ptr(w.gowner.id.String()), ptr(w.org.String()), ptr(w.gA.ID.String()), nil, func(tx pgx.Tx) {
			if st := ins(tx); st != "42501" {
				t.Errorf("selected-Group context Group insert: SQLSTATE %q, want 42501", st)
			}
		})
		if e.state(t, w.org, w.other) != before {
			t.Error("a refused runtime statement changed state")
		}
		u := e.newUser(t)
		e.grantOrg(t, w.org, u, tenancy.OrgMember)
		e.grantGroup(t, w.org, w.gB.ID, u, tenancy.GroupOwner)
		if _, err := e.setup.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND scope_type = 'organization'`, u.id.String()); err != nil {
			t.Fatal(err)
		}
		before = e.state(t, w.org, w.other)
		rlsTx(t, rt, ptr(u.id.String()), ptr(w.org.String()), ptr(""), nil, func(tx pgx.Tx) {
			if st := ins(tx); st != "42501" {
				t.Errorf("Group-role-only Group insert: SQLSTATE %q, want 42501", st)
			}
		})
		if e.state(t, w.org, w.other) != before {
			t.Error("a refused runtime statement changed state")
		}
	})

	t.Run("ONB-9e direct eligibility reader calls", func(t *testing.T) {
		rt := e.rlsRuntime(t)
		eligible := func(user, org, group *string, p tenancy.BusinessProfileID) (bool, string) {
			var got bool
			var st string
			rlsTx(t, rt, user, org, group, nil, func(tx pgx.Tx) {
				st = stateOrUnknown(tx.QueryRow(ctx, `SELECT app.rls_profile_eligible($1::text::uuid)`, p.String()).Scan(&got))
			})
			return got, st
		}
		// The legitimate context (the owner of attached gA, Organization-only): true, the only positive.
		v := e.newUser(t)
		e.grantOrg(t, w.org, v, tenancy.OrgMember)
		e.grantGroup(t, w.org, w.gA.ID, v, tenancy.GroupOwner)
		if ok, st := eligible(ptr(v.id.String()), ptr(w.org.String()), ptr(""), w.pA.ID); !ok || st != "" {
			t.Errorf("legitimate context: %v %q, want true", ok, st)
		}
		// No context at all: false.
		if ok, st := eligible(nil, nil, nil, w.pA.ID); ok || st != "" {
			t.Errorf("no context: %v %q, want false", ok, st)
		}
		// A selected Group of another Organization, and a same-Organization Group v is not bound to: false.
		foreignG := e.group(t, w.other, w.foreign, "Reader Foreign", fmt.Sprintf("reader-foreign-%d", itSeq.Add(1)))
		if ok, _ := eligible(ptr(v.id.String()), ptr(w.org.String()), ptr(foreignG.ID.String()), w.pA.ID); ok {
			t.Error("a selected Group of another Organization made pA eligible")
		}
		if ok, _ := eligible(ptr(v.id.String()), ptr(w.org.String()), ptr(w.gB.ID.String()), w.pA.ID); ok {
			t.Error("an unbound selected Group made pA eligible")
		}
		// The attack C2 shape: v's Organization binding removed by the trusted path, its gA binding and a selected-gA
		// context retained: false, and the association INSERT is refused.
		gN := e.group(t, w.org, w.owner, "Reader Target", fmt.Sprintf("reader-target-%d", itSeq.Add(1)))
		e.grantGroup(t, w.org, gN.ID, v, tenancy.GroupOwner)
		if _, err := e.setup.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND scope_type = 'organization'`, v.id.String()); err != nil {
			t.Fatal(err)
		}
		if ok, _ := eligible(ptr(v.id.String()), ptr(w.org.String()), ptr(w.gA.ID.String()), w.pA.ID); ok {
			t.Error("a retained selected-Group context without Organization membership made pA eligible")
		}
		for _, sel := range []string{w.gA.ID.String(), gN.ID.String(), ""} {
			rlsTx(t, rt, ptr(v.id.String()), ptr(w.org.String()), ptr(sel), nil, func(tx pgx.Tx) {
				st, n := rlsStmt(t, tx, `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid)`,
					w.org.String(), gN.ID.String(), w.pA.ID.String())
				if st != "42501" || n != 0 {
					t.Errorf("association insert after the Organization binding removal (selected %q): SQLSTATE %q rows %d, want 42501", sel, st, n)
				}
			})
		}
	})
}
