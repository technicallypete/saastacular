//go:build integration

package access

import (
	"context"
	"fmt"
	"testing"

	"core/internal/audit"
	"core/internal/tenancy"
)

// TestMembership is the G019 b2 membership integration entry point (Plan r1 batch b2: harness row B-MEMBERSHIP-IT runs
// ./internal/tenancy/access/ -run '^TestMembership$' with an it_count pin on these subtests). Each subtest builds its own
// environment. Cases are added here as they are written. Subtest names follow the Plan's row-ID convention ("<ROW-ID>
// <description>", as TestRLSIsolation's "RLS-SCOPE …" and TestOnboarding's "ONB-1 …"), so a row selector is
// '^TestMembership$/^MEM-14_' (go test turns the space into "_", which keeps MEM-1_ apart from MEM-10_/MEM-14_).
// Pending (NOT EXECUTED) while the native test hold applies; DB evidence after Q1.
func TestMembership(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(*testing.T)
	}{
		{"MEM-10 consume an Organization invitation", mem10ConsumeOrganizationInvitation},
		{"MEM-11 consume negatives", mem11ConsumeOrganizationNegatives},
		{"MEM-14 own-Organizations reader", mem14OwnOrganizationsReader},
		{"MEM-14 retire every page", mem14RetireUserPages},
		{"MEM-14 retire stops and resumes", mem14RetireUserStopsAndResumes},
		{"MEM-14 finalize refuses a new membership", mem14FinalizeNewMembership},
		{"MEM-14 finalize refuses a rejoin", mem14FinalizeRejoin},
		{"MEM-14 finalize refuses a late sole ownership", mem14FinalizeSoleOwner},
		{"MEM-14 finalize waits in both lock orders", mem14FinalizeWaits},
		// G019 writer25 (Codex writer23 review, "bounded membership gaps required before merge" items 1–3).
		{"MEM-8 pending cap race and paging", mem8CapRaceAndPaging},
		{"MEM-9 visibility, upgrade and consumed revoke", mem9UpgradeVisibilityConsumedRevoke},
		{"MEM-2 both owners administer with hints", mem2BothOwnersAdminister},
		{"MEM-6 membership denial attribution", mem6DenialAttribution},
		{"MEM-7 transfer audit fault rollback", mem7TransferAuditFault},
		{"MEM-13 dependent departure and readmission", mem13DependentDepartureReadmission},
		// G060 amendment01 A3: a failure after the guard's nested Group DELETEs rolls the whole departure back.
		{"MEM-13b departure audit fault rollback", mem13bDepartureAuditFault},
	} {
		t.Run(c.name, c.run)
	}
}

// memOrgScope is the Organization scope selector.
var memOrgScope = Scope{Kind: tenancy.ScopeOrganization}

// memMember finds user in one administrators' members page of sc (limit 0, no filter).
func (e *itEnv) memMember(t *testing.T, what string, c context.Context, tc *TenantContext, sc Scope, user tenancy.UserID) Member {
	t.Helper()
	pg, err := e.svc.ListMembers(c, tc, sc, "", 0, "")
	if err != nil {
		t.Fatalf("%s: members: %v", what, err)
	}
	for _, m := range pg.Members {
		if m.UserID == user {
			return m
		}
	}
	t.Fatalf("%s: %s is not on the page", what, user)
	return Member{}
}

// memSelf reads the subject's own membership of sc.
func (e *itEnv) memSelf(t *testing.T, what string, c context.Context, tc *TenantContext, sc Scope) SelfMember {
	t.Helper()
	s, err := e.svc.SelfMembership(c, tc, sc)
	if err != nil {
		t.Fatalf("%s: self: %v", what, err)
	}
	return s
}

// mem2BothOwnersAdminister (MEM-2/MEM-5; review item 3): the existing two-owner fixture through the facade. Before the
// promotion the sole owner is the scope's last owner (LastOwner, no leave) and sees the admin as a promotion and
// transfer candidate and nobody as demotable; the admin sees no ownership hints. After PromoteToOwner both owners are
// not last (OwnerCount 2: each may leave, each sees the other as demotable and no longer as a candidate), both
// administer ordinary members through the facade, and the second owner demotes the first, who becomes an admin while
// the second is now the last owner.
func mem2BothOwnersAdminister(t *testing.T) {
	e := setupIT(t)
	o, p, m1, m2 := e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t)
	a := e.org(t, "MEM2 A")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.grantOrg(t, a, p, tenancy.OrgAdmin)
	e.grantOrg(t, a, m1, tenancy.OrgMember)
	e.grantOrg(t, a, m2, tenancy.OrgMember)
	e.bind(t, o, a) // G016 A2: the actors' credentials for A, setup only
	e.bind(t, p, a)
	octx, otc := e.as(t, o, a)
	pctx, ptc := e.as(t, p, a)
	owners := func() int {
		return e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE scope_id = $1::text::uuid AND role = 'org_owner'`, a.String())
	}

	// Before.
	if s := e.memSelf(t, "o before", octx, otc, memOrgScope); s.Role != string(tenancy.OrgOwner) || !s.LastOwner || s.CanLeave || !s.CanManage {
		t.Errorf("o before: %+v, want the last owner who may manage but not leave", s)
	}
	if m := e.memMember(t, "o sees p before", octx, otc, memOrgScope, p.id); !m.CanPromote || !m.CanTransfer || m.CanDemote || !m.CanChangeRole || !m.CanRemove {
		t.Errorf("o sees p before: %+v, want promote/transfer/change/remove and no demote", m)
	}
	if m := e.memMember(t, "o sees o before", octx, otc, memOrgScope, o.id); !m.Self || m.CanDemote || m.CanPromote || m.CanChangeRole || m.CanRemove {
		t.Errorf("o sees o before: %+v, want self with no hints", m)
	}
	if m := e.memMember(t, "p sees o before", pctx, ptc, memOrgScope, o.id); m.CanDemote || m.CanPromote || m.CanTransfer || m.CanChangeRole || m.CanRemove {
		t.Errorf("p sees o before: %+v, want no hints on an owner for an admin", m)
	}
	if n := owners(); n != 1 {
		t.Fatalf("owners before: %d", n)
	}

	// Promotion to a second owner.
	if m, err := e.svc.PromoteToOwner(octx, otc, memOrgScope, p.id); err != nil || m.Role != string(tenancy.OrgOwner) || m.UserID != p.id {
		t.Fatalf("promote: %v, %+v", err, m)
	}
	if n := owners(); n != 2 {
		t.Fatalf("owners after the promotion: %d", n)
	}
	for _, x := range []struct {
		what string
		c    context.Context
		tc   *TenantContext
		them tenancy.UserID
	}{{"o", octx, otc, p.id}, {"p", pctx, ptc, o.id}} {
		if s := e.memSelf(t, x.what+" after", x.c, x.tc, memOrgScope); s.Role != string(tenancy.OrgOwner) || s.LastOwner || !s.CanLeave || !s.CanManage {
			t.Errorf("%s after: %+v, want an owner who is not last and may leave and manage", x.what, s)
		}
		if m := e.memMember(t, x.what+" sees the other after", x.c, x.tc, memOrgScope, x.them); !m.CanDemote || m.CanPromote || m.CanTransfer || m.CanChangeRole || m.CanRemove {
			t.Errorf("%s sees the other after: %+v, want demote only", x.what, m)
		}
	}

	// Both owners administer ordinary members through the facade.
	if m, err := e.svc.ChangeMemberRole(octx, otc, memOrgScope, m1.id, string(tenancy.OrgAdmin)); err != nil || m.Role != string(tenancy.OrgAdmin) {
		t.Errorf("o changes m1: %v, %+v", err, m)
	}
	if m, err := e.svc.ChangeMemberRole(pctx, ptc, memOrgScope, m2.id, string(tenancy.OrgAdmin)); err != nil || m.Role != string(tenancy.OrgAdmin) {
		t.Errorf("p changes m2: %v, %+v", err, m)
	}
	if r1, r2 := e.memRole(t, a.String(), m1.id), e.memRole(t, a.String(), m2.id); r1 != string(tenancy.OrgAdmin) || r2 != string(tenancy.OrgAdmin) {
		t.Errorf("roles after both changes: m1 %s, m2 %s", r1, r2)
	}

	// The second owner demotes the first and becomes the last owner.
	if m, err := e.svc.DemoteOwner(pctx, ptc, memOrgScope, o.id, string(tenancy.OrgAdmin)); err != nil || m.Role != string(tenancy.OrgAdmin) {
		t.Fatalf("p demotes o: %v, %+v", err, m)
	}
	if s := e.memSelf(t, "p last", pctx, ptc, memOrgScope); !s.LastOwner || s.CanLeave {
		t.Errorf("p after demoting o: %+v, want the last owner", s)
	}
	if s := e.memSelf(t, "o demoted", octx, otc, memOrgScope); s.Role != string(tenancy.OrgAdmin) || s.LastOwner || !s.CanLeave || !s.CanManage {
		t.Errorf("o after the demotion: %+v, want an admin who may leave and manage", s)
	}
	if n := owners(); n != 1 {
		t.Errorf("owners at the end: %d", n)
	}
}

// mem6DenialAttribution (MEM-6; review item 3): representative facade denials of ordinary management touching owners
// and of unauthorized ownership management, each with the pinned code/field, unchanged roles and bindings, no success
// capture (no membership.role_changed/removed row in the Organization), exactly one access.decision deny record of the
// operation with reason role_not_permitted, and exactly one truthful access.denied audit row: operation, actor,
// Organization, the administered scope as the target (resource and Group), no subject and the reason. This holds for a
// capability denial (manage_owners refused by policy) and for an owner_subject refusal decided on the locked facts
// after organization.manage_members was granted (a recordable authorization refusal, AUDIT.md). Context/foreign/
// stale-session denials stay with the existing context, attack and audit rows.
func mem6DenialAttribution(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, q, gw, ga := e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t)
	a := e.org(t, "MEM6 A")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.grantOrg(t, a, q, tenancy.OrgAdmin)
	e.grantOrg(t, a, gw, tenancy.OrgMember)
	e.grantOrg(t, a, ga, tenancy.OrgMember)
	g := e.group(t, a, gw, "MEM6 Group", fmt.Sprintf("mem6-%d", itSeq.Add(1))) // gw: the trigger-written Group owner
	e.grantGroup(t, a, g.ID, ga, tenancy.GroupAdmin)
	for _, u := range []*itUser{o, q, gw, ga} {
		e.bind(t, u, a) // G016 A2: credentials for A, setup only
	}
	qctx, qtc := e.as(t, q, a)
	orctx, _ := beginAs(t, e.svc, ctx, o, a)
	otcG := resolve(t, e.svc, orctx, a, g.ID)
	garctx, _ := beginAs(t, e.svc, ctx, ga, a)
	gatc := resolve(t, e.svc, garctx, a, g.ID)
	gid := g.ID
	gScope := Scope{Kind: tenancy.ScopeGroup, Group: &gid}

	roles := func() string {
		return fmt.Sprintf("o=%s q=%s gw=%s ga=%s bindings=%d", e.memRole(t, a.String(), o.id), e.memRole(t, a.String(), q.id),
			e.memRole(t, g.ID.String(), gw.id), e.memRole(t, g.ID.String(), ga.id),
			e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid`, a.String()))
	}
	successes := func() int {
		return e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid
  AND action IN ('membership.role_changed', 'membership.removed')`, a.String())
	}
	type deniedRow struct{ operation, actor, org, group, resource, resourceID, subject, reason string }
	// denied lists the four actors' access.denied rows in write order (ids are gen_random_uuid(), so occurred_at orders;
	// a row's callers slice off the rows that existed before the call).
	denied := func() []deniedRow {
		rows, err := e.setup.Query(ctx, `SELECT operation, coalesce(actor_user_id::text, ''), coalesce(organization_id::text, ''),
  coalesce(group_id::text, ''), coalesce(resource_type, ''), coalesce(resource_id::text, ''), coalesce(subject_user_id::text, ''),
  reason FROM audit.events WHERE action = 'access.denied' AND actor_user_id IN ($1, $2, $3, $4)
  ORDER BY occurred_at, id`, o.id.String(), q.id.String(), gw.id.String(), ga.id.String())
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []deniedRow
		for rows.Next() {
			var r deniedRow
			if err := rows.Scan(&r.operation, &r.actor, &r.org, &r.group, &r.resource, &r.resourceID, &r.subject, &r.reason); err != nil {
				t.Fatal(err)
			}
			out = append(out, r)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return out
	}
	start := roles()

	for _, c := range []struct {
		name, op string
		call     func() error
		field    string
		actor    *itUser
		group    string // the audited Group target ('' for an Organization target)
	}{
		{"admin promotes self", opOwnershipPromote, func() error {
			_, err := e.svc.PromoteToOwner(qctx, qtc, memOrgScope, q.id)
			return err
		}, "", q, ""},
		{"admin changes an owner's role", opMembershipChangeRole, func() error {
			_, err := e.svc.ChangeMemberRole(qctx, qtc, memOrgScope, o.id, string(tenancy.OrgMember))
			return err
		}, fieldOwnerSubject, q, ""},
		{"admin removes an owner", opMembershipRemove, func() error {
			return e.svc.RemoveMember(qctx, qtc, memOrgScope, o.id)
		}, fieldOwnerSubject, q, ""},
		{"Organization owner demotes a Group owner", opOwnershipDemote, func() error {
			_, err := e.svc.DemoteOwner(orctx, otcG, gScope, gw.id, string(tenancy.GroupMember))
			return err
		}, "", o, g.ID.String()},
		{"Group admin promotes self", opOwnershipPromote, func() error {
			_, err := e.svc.PromoteToOwner(garctx, gatc, gScope, ga.id)
			return err
		}, "", ga, g.ID.String()},
	} {
		t.Run(c.name, func(t *testing.T) {
			before, from := len(denied()), len(e.logs.String())
			err := c.call()
			memWantCode(t, c.name, err, CodeForbidden, c.field)
			if now := roles(); now != start {
				t.Errorf("%s changed state: %s -> %s", c.name, start, now)
			}
			if n := successes(); n != 0 {
				t.Errorf("%s: %d success captures", c.name, n)
			}
			recs := e.memOpRecords(t, from, c.op)
			if len(recs) != 1 || recs[0]["outcome"] != outcomeDeny || recs[0]["reason"] != string(reasonRoleNotPermitted) {
				t.Errorf("%s: %s access.decision records %v, want one deny with reason role_not_permitted", c.name, c.op, recs)
			}
			got := denied()[before:]
			want := deniedRow{operation: c.op, actor: c.actor.id.String(), org: a.String(), group: c.group,
				resource: string(audit.ResourceOrganization), resourceID: a.String(), reason: string(reasonRoleNotPermitted)}
			if c.group != "" {
				want.resource, want.resourceID = string(audit.ResourceGroup), c.group
			}
			if len(got) != 1 || got[0] != want {
				t.Errorf("%s: access.denied rows %v, want exactly %v", c.name, got, want)
			}
		})
	}
}

// mem7TransferAuditFault (MEM-7; review item 3): the AT-ATOM audit-fault mechanism (a superuser NOT VALID check on
// audit.events) planted so that the recipient's promotion event passes and the actor's own demotion event, the later
// write after the first role mutation, fails: the transfer is refused, both bindings keep their roles, the owner count
// stays 1, no membership.role_changed row exists and the operation's access.decision record is failed. With the fault
// removed the same transfer succeeds as the control: recipient owner, actor admin, one owner, the two role_changed rows
// in order.
func mem7TransferAuditFault(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, p := e.newUser(t), e.newUser(t)
	a := e.org(t, "MEM7 A")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.grantOrg(t, a, p, tenancy.OrgAdmin)
	e.bind(t, o, a) // G016 A2: the actor's credential for A, setup only
	octx, otc := e.as(t, o, a)
	type change struct{ subject, previous, role string }
	changes := func() []change {
		rows, err := e.setup.Query(ctx, `SELECT subject_user_id::text, coalesce(previous_role, ''), coalesce(role, '') FROM audit.events
  WHERE organization_id = $1::text::uuid AND action = 'membership.role_changed' ORDER BY occurred_at, id`, a.String())
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []change
		for rows.Next() {
			var c change
			if err := rows.Scan(&c.subject, &c.previous, &c.role); err != nil {
				t.Fatal(err)
			}
			out = append(out, c)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return out
	}
	state := func() string {
		return fmt.Sprintf("o=%s p=%s owners=%d", e.memRole(t, a.String(), o.id), e.memRole(t, a.String(), p.id),
			e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE scope_id = $1::text::uuid AND role = 'org_owner'`, a.String()))
	}
	before := state()

	name := fmt.Sprintf("audit_fault_%d", itSeq.Add(1))
	if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events ADD CONSTRAINT `+name+` CHECK (NOT (action = 'membership.role_changed'
  AND subject_user_id = '`+o.id.String()+`')) NOT VALID`); err != nil {
		t.Fatal(err)
	}
	from := len(e.logs.String())
	err := e.svc.TransferOwnership(octx, otc, memOrgScope, p.id, string(tenancy.OrgAdmin))
	if err == nil {
		t.Fatalf("transfer under the audit fault succeeded (state %s)", state())
	}
	t.Logf("transfer under the audit fault: %v", err)
	if now := state(); now != before {
		t.Errorf("transfer under the audit fault changed state: %s -> %s", before, now)
	}
	if ch := changes(); len(ch) != 0 {
		t.Errorf("membership.role_changed rows after the faulted transfer: %v", ch)
	}
	if recs := e.memOpRecords(t, from, opOwnershipTransfer); len(recs) != 1 || recs[0]["outcome"] != outcomeFailed {
		t.Errorf("faulted transfer access.decision records %v, want one failed record", recs)
	}
	if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events DROP CONSTRAINT `+name); err != nil {
		t.Fatal(err)
	}

	// Control: the same transfer succeeds once the fault is removed.
	from = len(e.logs.String())
	if err := e.svc.TransferOwnership(octx, otc, memOrgScope, p.id, string(tenancy.OrgAdmin)); err != nil {
		t.Fatalf("control transfer: %v", err)
	}
	if now, want := state(), fmt.Sprintf("o=%s p=%s owners=1", tenancy.OrgAdmin, tenancy.OrgOwner); now != want {
		t.Errorf("control transfer state %s, want %s", now, want)
	}
	want := []change{{p.id.String(), string(tenancy.OrgAdmin), string(tenancy.OrgOwner)}, {o.id.String(), string(tenancy.OrgOwner), string(tenancy.OrgAdmin)}}
	if ch := changes(); len(ch) != 2 || ch[0] != want[0] || ch[1] != want[1] {
		t.Errorf("control transfer membership.role_changed rows %v, want %v", ch, want)
	}
	if recs := e.memOpRecords(t, from, opOwnershipTransfer); len(recs) != 1 || recs[0]["outcome"] != outcomeAllow {
		t.Errorf("control transfer access.decision records %v, want one allow record", recs)
	}

	// G060 amendment01 A3/release item 5: TransferOwnership(..., formerRole "none") with dependent Group bindings, the
	// fourth departure caller. In Organization b, o2 owns b, solely owns Group gs and is a member of Group gm (owned by
	// the admin p2). (i) The departure reaches gs's sole-owner binding in the guard's cascade: last_owner, and the whole
	// transfer rolls back (p2 stays admin, o2 keeps every binding, one owner, no role_changed or removed row). (ii) After
	// o2 hands gs to p2 (keeping the member role), the transfer succeeds atomically: p2 owner, o2 holds nothing in b, one
	// owner, membership.removed for gs, gm and the Organization binding.
	o2, p2 := e.newUser(t), e.newUser(t)
	b := e.org(t, "MEM7 B")
	e.grantOrg(t, b, o2, tenancy.OrgOwner)
	e.grantOrg(t, b, p2, tenancy.OrgAdmin)
	gs := e.group(t, b, o2, "MEM7 B sole", fmt.Sprintf("mem7-%d", itSeq.Add(1)))
	gm := e.group(t, b, p2, "MEM7 B member", fmt.Sprintf("mem7-%d", itSeq.Add(1)))
	e.grantGroup(t, b, gm.ID, o2, tenancy.GroupMember)
	e.grantGroup(t, b, gs.ID, p2, tenancy.GroupAdmin) // gs's succession recipient (a same-scope admin); o2 stays its sole owner
	e.bind(t, o2, b)
	o2ctx, o2tc := e.as(t, o2, b)
	bstate := func() string {
		return fmt.Sprintf("o2=%s p2=%s owners=%d o2bindings=%d", e.memRole(t, b.String(), o2.id), e.memRole(t, b.String(), p2.id),
			e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE scope_id = $1::text::uuid AND role = 'org_owner'`, b.String()),
			e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid`,
				b.String(), o2.id.String()))
	}
	bEvents := func(action string) int {
		return e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = $2`, b.String(), action)
	}
	bBefore, rcBefore, rmBefore := bstate(), bEvents("membership.role_changed"), bEvents("membership.removed")
	from = len(e.logs.String())
	err = e.svc.TransferOwnership(o2ctx, o2tc, memOrgScope, p2.id, roleNone)
	memWantCode(t, "transfer to none while solely owning a Group", err, CodeConflict, string(tenancy.CodeLastOwner))
	if now := bstate(); now != bBefore {
		t.Errorf("refused transfer to none changed state: %s -> %s", bBefore, now)
	}
	if rc, rm := bEvents("membership.role_changed"), bEvents("membership.removed"); rc != rcBefore || rm != rmBefore {
		t.Errorf("refused transfer to none wrote role_changed %d->%d, removed %d->%d", rcBefore, rc, rmBefore, rm)
	}
	if recs := e.memOpRecords(t, from, opOwnershipTransfer); len(recs) != 1 || recs[0]["outcome"] == outcomeAllow {
		t.Errorf("refused transfer to none access.decision records %v, want one non-allow record", recs)
	}
	gsid := gs.ID
	gsctx, _ := beginAs(t, e.svc, ctx, o2, b)
	gstc := resolve(t, e.svc, gsctx, b, gs.ID)
	if err := e.svc.TransferOwnership(gsctx, gstc, Scope{Kind: tenancy.ScopeGroup, Group: &gsid}, p2.id, string(tenancy.GroupMember)); err != nil {
		t.Fatalf("Group succession before the departure: %v", err)
	}
	from = len(e.logs.String())
	if err := e.svc.TransferOwnership(o2ctx, o2tc, memOrgScope, p2.id, roleNone); err != nil {
		t.Fatalf("transfer to none after the succession: %v", err)
	}
	if now, want := bstate(), fmt.Sprintf("o2=none p2=%s owners=1 o2bindings=0", tenancy.OrgOwner); now != want {
		t.Errorf("transfer to none state %s, want %s", now, want)
	}
	if n := e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'membership.removed'
  AND subject_user_id = $2::text::uuid AND (group_id IS NULL OR group_id IN ($3::text::uuid, $4::text::uuid))`, b.String(), o2.id.String(),
		gs.ID.String(), gm.ID.String()); n != 3 {
		t.Errorf("membership.removed rows of the departure: %d, want 3 (gs, gm, Organization)", n)
	}
	if recs := e.memOpRecords(t, from, opOwnershipTransfer); len(recs) != 1 || recs[0]["outcome"] != outcomeAllow {
		t.Errorf("transfer to none access.decision records %v, want one allow record", recs)
	}
}

// mem13bDepartureAuditFault (MEM-13b, G060 amendment01 A3): the MEM-7 AT-ATOM fault (a superuser NOT VALID check on
// audit.events) planted on the Organization row's membership.removed capture of subject u, which the guard's cascade
// writes after the nested Group DELETEs of the same statement: RemoveMember fails, all three bindings remain, no
// membership.removed row exists and the operation's access.decision record is failed; with the fault removed the same
// removal succeeds as the control (three rows). The distinct boundary: a failure raised after the nested Group DELETEs
// inside the one Organization DELETE statement.
func mem13bDepartureAuditFault(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, u := e.newUser(t), e.newUser(t)
	a := e.org(t, "MEM13b A")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.grantOrg(t, a, u, tenancy.OrgMember)
	g1 := e.group(t, a, o, "MEM13b G1", fmt.Sprintf("mem13b-%d", itSeq.Add(1)))
	g2 := e.group(t, a, o, "MEM13b G2", fmt.Sprintf("mem13b-%d", itSeq.Add(1)))
	e.grantGroup(t, a, g1.ID, u, tenancy.GroupMember)
	e.grantGroup(t, a, g2.ID, u, tenancy.GroupAdmin)
	e.bind(t, o, a) // G016 A2: the actor's credential for A, setup only
	octx, otc := e.as(t, o, a)
	bindings := func() int {
		return e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid`,
			a.String(), u.id.String())
	}
	removed := func() int {
		return e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'membership.removed'
  AND subject_user_id = $2::text::uuid`, a.String(), u.id.String())
	}
	if n := bindings(); n != 3 {
		t.Fatalf("bindings at the start: %d", n)
	}
	name := fmt.Sprintf("audit_fault_%d", itSeq.Add(1))
	if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events ADD CONSTRAINT `+name+` CHECK (NOT (action = 'membership.removed'
  AND subject_user_id = '`+u.id.String()+`' AND group_id IS NULL)) NOT VALID`); err != nil {
		t.Fatal(err)
	}
	from := len(e.logs.String())
	err := e.svc.RemoveMember(octx, otc, memOrgScope, u.id)
	if err == nil {
		t.Fatalf("removal under the audit fault succeeded (%d bindings)", bindings())
	}
	t.Logf("removal under the audit fault: %v", err)
	if n := bindings(); n != 3 {
		t.Errorf("bindings after the faulted removal: %d, want 3 (the nested Group DELETEs rolled back)", n)
	}
	if r1, r2 := e.memRole(t, g1.ID.String(), u.id), e.memRole(t, g2.ID.String(), u.id); r1 != string(tenancy.GroupMember) || r2 != string(tenancy.GroupAdmin) {
		t.Errorf("Group roles after the faulted removal: %s / %s", r1, r2)
	}
	if n := removed(); n != 0 {
		t.Errorf("membership.removed rows after the faulted removal: %d", n)
	}
	if recs := e.memOpRecords(t, from, opMembershipRemove); len(recs) != 1 || recs[0]["outcome"] != outcomeFailed {
		t.Errorf("faulted removal access.decision records %v, want one failed record", recs)
	}
	if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events DROP CONSTRAINT `+name); err != nil {
		t.Fatal(err)
	}
	// Control: the same removal succeeds once the fault is removed.
	from = len(e.logs.String())
	if err := e.svc.RemoveMember(octx, otc, memOrgScope, u.id); err != nil {
		t.Fatalf("control removal: %v", err)
	}
	if n, r := bindings(), removed(); n != 0 || r != 3 {
		t.Errorf("control removal: %d bindings, %d membership.removed rows, want 0 and 3", n, r)
	}
	if recs := e.memOpRecords(t, from, opMembershipRemove); len(recs) != 1 || recs[0]["outcome"] != outcomeAllow {
		t.Errorf("control removal access.decision records %v, want one allow record", recs)
	}
}

// mem13DependentDepartureReadmission (MEM-13; review item 2): a member who owns Group g1 and belongs to Group g2. The
// administrative Organization removal is refused up front (owner_subject) with all three bindings intact, no
// membership.removed row and exactly one truthful access.denied (the Organization as the target); after the required succession (the member transfers g1 and stays a member) one facade call in
// one transaction removes g1, g2 and the Organization binding (three membership.removed rows); a fresh readmission
// through an Organization invitation restores the Organization membership only, no Group binding.
func mem13DependentDepartureReadmission(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	owner := e.memOwnerService(t)
	o, p, u := e.newUser(t), e.newUser(t), e.newUser(t)
	a := e.org(t, "MEM13 A")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.grantOrg(t, a, p, tenancy.OrgMember)
	e.grantOrg(t, a, u, tenancy.OrgMember)
	g1 := e.group(t, a, u, "MEM13 G1", fmt.Sprintf("mem13-%d", itSeq.Add(1))) // u: the trigger-written owner of g1
	g2 := e.group(t, a, o, "MEM13 G2", fmt.Sprintf("mem13-%d", itSeq.Add(1)))
	e.grantGroup(t, a, g2.ID, u, tenancy.GroupMember)
	e.grantGroup(t, a, g1.ID, p, tenancy.GroupAdmin)
	e.bind(t, o, a) // G016 A2: the actors' credentials for A, setup only
	e.bind(t, u, a)
	octx, otc := e.as(t, o, a)
	bindings := func(scope string) int {
		return e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid
  AND ($3 = '' OR scope_type = $3)`, a.String(), u.id.String(), scope)
	}
	removed := func() (total, groups int) {
		total = e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'membership.removed'
  AND subject_user_id = $2::text::uuid`, a.String(), u.id.String())
		groups = e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'membership.removed'
  AND subject_user_id = $2::text::uuid AND group_id IN ($3::text::uuid, $4::text::uuid)`, a.String(), u.id.String(), g1.ID.String(), g2.ID.String())
		return
	}
	if n := bindings(""); n != 3 {
		t.Fatalf("bindings at the start: %d", n)
	}

	// (1) Before the succession the removal is refused up front with nothing written.
	err := e.svc.RemoveMember(octx, otc, memOrgScope, u.id)
	memWantCode(t, "removal of a Group owner", err, CodeForbidden, fieldOwnerSubject)
	if n, r := bindings(""), e.memRole(t, g1.ID.String(), u.id); n != 3 || r != string(tenancy.GroupOwner) {
		t.Errorf("after the refused removal: %d bindings, g1 role %s", n, r)
	}
	if total, _ := removed(); total != 0 {
		t.Errorf("membership.removed rows after the refused removal: %d", total)
	}
	// The dependent Group-owner refusal is the Organization removal's one truthful access.denied: the actor, the
	// Organization as the administered target, no Group and no subject.
	var denials []string
	rows, err := e.setup.Query(ctx, `SELECT concat_ws('|', operation, coalesce(organization_id::text, ''), coalesce(group_id::text, ''),
  coalesce(resource_type, ''), coalesce(resource_id::text, ''), coalesce(subject_user_id::text, ''), reason)
  FROM audit.events WHERE action = 'access.denied' AND actor_user_id = $1::text::uuid ORDER BY occurred_at, id`, o.id.String())
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			t.Fatal(err)
		}
		denials = append(denials, d)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	wantDenial := opMembershipRemove + "|" + a.String() + "||" + string(audit.ResourceOrganization) + "|" + a.String() + "||" +
		string(reasonRoleNotPermitted)
	if len(denials) != 1 || denials[0] != wantDenial {
		t.Errorf("access.denied rows of the refused removal %v, want exactly [%s]", denials, wantDenial)
	}

	// (2) The required succession: u transfers g1 to p and stays a member of g1.
	urctx, _ := beginAs(t, e.svc, ctx, u, a)
	utc := resolve(t, e.svc, urctx, a, g1.ID)
	g1id := g1.ID
	if err := e.svc.TransferOwnership(urctx, utc, Scope{Kind: tenancy.ScopeGroup, Group: &g1id}, p.id, string(tenancy.GroupMember)); err != nil {
		t.Fatalf("succession: %v", err)
	}
	if ru, rp := e.memRole(t, g1.ID.String(), u.id), e.memRole(t, g1.ID.String(), p.id); ru != string(tenancy.GroupMember) || rp != string(tenancy.GroupOwner) {
		t.Fatalf("after the succession: u %s, p %s", ru, rp)
	}

	// (3) One transaction removes g1, g2 and the Organization binding.
	svc, h := e.service(e.id)
	rctx, _ := beginAs(t, svc, ctx, o, a)
	tc := resolve(t, svc, rctx, a, tenancy.GroupID{})
	h.reset()
	if err := svc.RemoveMember(rctx, tc, memOrgScope, u.id); err != nil {
		t.Fatalf("removal after the succession: %v", err)
	}
	if begins, _ := h.counts(); begins != 1 {
		t.Errorf("the removal began %d transactions, want 1", begins)
	}
	if n := bindings(""); n != 0 {
		t.Errorf("bindings after the removal: %d", n)
	}
	if total, groups := removed(); total != 3 || groups != 2 {
		t.Errorf("membership.removed rows after the removal: %d (%d Group rows), want 3 (2)", total, groups)
	}
	// G060 amendment01 A3: the guard's cascade writes the rows as the acting user, the Group rows in ascending group_id,
	// then the Organization row (the same producer, order and attribution as the former Go sequence).
	var seq, want []string
	rows, err = e.setup.Query(ctx, `SELECT concat_ws('|', coalesce(actor_user_id::text, ''), coalesce(resource_type, ''), coalesce(group_id::text, ''))
  FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'membership.removed' AND subject_user_id = $2::text::uuid
  ORDER BY occurred_at, id`, a.String(), u.id.String())
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		seq = append(seq, s)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	lo, hi := g1.ID.String(), g2.ID.String()
	if hi < lo {
		lo, hi = hi, lo
	}
	for _, g := range []string{lo, hi, ""} {
		want = append(want, o.id.String()+"|"+string(audit.ResourceRoleBinding)+"|"+g)
	}
	if fmt.Sprint(seq) != fmt.Sprint(want) {
		t.Errorf("membership.removed rows of the removal %v, want %v (actor o, role_binding, Groups ascending, then the Organization)", seq, want)
	}

	// (4) Readmission restores the Organization membership only.
	addr := memAddr("readmit")
	iss := e.memIssueOrg(t, o, a, addr)
	out, err := owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, u.sub, addr), iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != tenancy.ClaimJoined || out.Binding == nil || out.Binding.Role != string(tenancy.OrgMember) {
		t.Fatalf("readmission: %v, %+v", err, out)
	}
	if org, groups := bindings("organization"), bindings("group"); org != 1 || groups != 0 {
		t.Errorf("after the readmission: %d Organization and %d Group bindings, want 1 and 0", org, groups)
	}
}
