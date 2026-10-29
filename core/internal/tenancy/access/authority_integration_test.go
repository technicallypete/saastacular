//go:build integration

package access

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

// TestAuthority is the G060 b2 entry point for the owner-only Group-management override and the bounded enabled mode
// (Plan r0 §5.2 T4/T5, r1 C3/C4/C5, correction02 C02-2): AUT-1..6 (owner-only write, direct runtime SQL, generic policy
// patch, default false, audit reasons) and MEM-OVR-1..8 (bounded administration, exclusions, self-grant refusals, the
// stale issuer and the independent-issuer upgrade). Subtest names follow the "<ROW-ID> <description>" convention, so a
// row selector is '^TestAuthority$/^MEM-OVR-8_'.
func TestAuthority(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(*testing.T)
	}{
		{"AUT-1 owner enables and disables with audit reasons", aut1OwnerEnablesDisables},
		{"AUT-2 admin, Group owner and member are refused", aut2NonOwnersRefused},
		{"AUT-3 direct runtime SQL by an admin meets the guard", aut3DirectSQLGuard},
		{"AUT-4 generic policy patch never sets the override", aut4GenericPolicyPatch},
		{"AUT-5 fresh Organizations default to false", aut5DefaultFalse},
		{"AUT-6 a disabled owner's raw runtime UPDATE meets the enabled-owner guard", aut6DisabledOwnerRawDenied},
		{"MEM-OVR-1 enabled staff administer ordinary members and invitations", memOvr1EnabledAdministers},
		{"MEM-OVR-2 exclusions under the override", memOvr2Exclusions},
		{"MEM-OVR-3 disabled override denies the next administration", memOvr3DisableDenies},
		{"MEM-OVR-5 self-targeted legacy intent is never consumed", memOvr5SelfIntent},
		{"MEM-OVR-6 override issuer, new member consumes", memOvr6NewMemberConsumes},
		{"MEM-OVR-7 owner self-demotion and member leave unchanged", memOvr7OwnerAndLeave},
		{"MEM-OVR-8 independent manager upgrades a member by invitation", memOvr8IndependentUpgrade},
		{"O-OVR-2 invitation issued under the override, override disabled", oOvr2StaleOverrideIssuer},
		{"O-OVR-4 issuer disabled before the claim", oOvr4IssuerDisabled},
	} {
		t.Run(c.name, c.run)
	}
}

// ovrWorld is one Organization with an owner o, an admin f (no Group binding), a Group g created by o and members m and u.
type ovrWorld struct {
	e          *itEnv
	a          tenancy.OrganizationID
	g          tenancy.Group
	o, f, m, u *itUser
}

func newOvrWorld(t *testing.T, label string) ovrWorld {
	t.Helper()
	e := setupIT(t)
	w := ovrWorld{e: e, o: e.newUser(t), f: e.newUser(t), m: e.newUser(t), u: e.newUser(t)}
	w.a = e.org(t, label)
	e.grantOrg(t, w.a, w.o, tenancy.OrgOwner)
	e.grantOrg(t, w.a, w.f, tenancy.OrgAdmin)
	e.grantOrg(t, w.a, w.m, tenancy.OrgMember)
	e.grantOrg(t, w.a, w.u, tenancy.OrgMember)
	w.g = e.group(t, w.a, w.o, label+" Group", fmt.Sprintf("ovr-%d", itSeq.Add(1)))
	e.grantGroup(t, w.a, w.g.ID, w.m, tenancy.GroupMember)
	for _, x := range []*itUser{w.o, w.f, w.m, w.u} {
		e.bind(t, x, w.a)
	}
	return w
}

func (w ovrWorld) scope() Scope {
	gid := w.g.ID
	return Scope{Kind: tenancy.ScopeGroup, Group: &gid}
}

func (w ovrWorld) setOverride(t *testing.T, enabled bool) {
	t.Helper()
	c, tc := w.e.as(t, w.o, w.a)
	v, err := w.e.svc.SetGroupManagementOverride(c, tc, enabled)
	if err != nil || v.OverrideEnabled != enabled {
		t.Fatalf("owner sets the override to %t: %+v %v", enabled, v, err)
	}
}

func (w ovrWorld) overrideValue(t *testing.T) bool {
	t.Helper()
	var v bool
	if err := w.e.setup.QueryRow(context.Background(), `SELECT group_management_override FROM app.organizations WHERE id = $1::text::uuid`,
		w.a.String()).Scan(&v); err != nil {
		t.Fatalf("override value: %v", err)
	}
	return v
}

func (w ovrWorld) events(t *testing.T, action, extra string, args ...any) int {
	t.Helper()
	return w.e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = $2`+extra,
		append([]any{w.a.String(), action}, args...)...)
}

func aut1OwnerEnablesDisables(t *testing.T) {
	w := newOvrWorld(t, "AUT1")
	c, tc := w.e.as(t, w.o, w.a)
	if v, err := w.e.svc.GetGroupAuthority(c, tc); err != nil || v.OverrideEnabled || !v.CanChange {
		t.Fatalf("owner reads the default: %+v %v", v, err)
	}
	w.setOverride(t, true)
	w.setOverride(t, true) // absolute set: a no-op records nothing
	w.setOverride(t, false)
	if w.overrideValue(t) {
		t.Error("the override stayed enabled")
	}
	on := w.events(t, "organization.policy_changed", " AND reason = 'group_management_override_enabled'")
	off := w.events(t, "organization.policy_changed", " AND reason = 'group_management_override_disabled'")
	if on != 1 || off != 1 {
		t.Errorf("policy_changed reasons: %d enabled, %d disabled, want 1 and 1", on, off)
	}
	if n := w.events(t, "organization.policy_changed", " AND actor_user_id = $3::text::uuid AND resource_id = organization_id", w.o.id.String()); n != 2 {
		t.Errorf("attributed policy_changed rows: %d, want 2", n)
	}
	fc, ftc := w.e.as(t, w.f, w.a)
	if v, err := w.e.svc.GetGroupAuthority(fc, ftc); err != nil || v.CanChange {
		t.Errorf("admin reads the setting: %+v %v (can_change must be false)", v, err)
	}
}

func aut2NonOwnersRefused(t *testing.T) {
	w := newOvrWorld(t, "AUT2")
	// f also owns its own Group: still no override authority.
	w.e.group(t, w.a, w.f, "AUT2 admin's Group", fmt.Sprintf("ovr-%d", itSeq.Add(1)))
	go1 := w.e.newUser(t)
	w.e.grantOrg(t, w.a, go1, tenancy.OrgMember)
	w.e.group(t, w.a, go1, "AUT2 member's Group", fmt.Sprintf("ovr-%d", itSeq.Add(1)))
	w.e.bind(t, go1, w.a)
	for _, x := range []*itUser{w.f, go1, w.m} {
		c, tc := w.e.as(t, x, w.a)
		_, err := w.e.svc.SetGroupManagementOverride(c, tc, true)
		memWantCode(t, "non-owner enables", err, CodeForbidden, "")
	}
	if w.overrideValue(t) {
		t.Error("a refused write changed the override")
	}
	if n := w.events(t, "access.denied", " AND operation = 'organization.set_group_override'"); n != 3 {
		t.Errorf("recorded override denials: %d, want 3", n)
	}
}

func aut3DirectSQLGuard(t *testing.T) {
	w := newOvrWorld(t, "AUT3")
	rt := w.e.rlsRuntime(t)
	defer rt.Close(context.Background())
	f, a := w.f.id.String(), w.a.String()
	rlsTx(t, rt, &f, &a, ptr(""), nil, func(tx pgx.Tx) {
		st, cn, _ := rlsStmtC(t, tx, `UPDATE app.organizations SET group_management_override = true WHERE id = $1::text::uuid`, a)
		if st != "42501" || cn != "group_override_owner" {
			t.Errorf("admin direct UPDATE: SQLSTATE %q constraint %q, want 42501 group_override_owner", st, cn)
		}
	})
	o := w.o.id.String()
	rlsTx(t, rt, &o, &a, ptr(""), nil, func(tx pgx.Tx) {
		st, _, n := rlsStmtC(t, tx, `UPDATE app.organizations SET group_management_override = true WHERE id = $1::text::uuid`, a)
		if st != "" || n != 1 {
			t.Errorf("owner direct UPDATE: SQLSTATE %q rows %d, want success", st, n)
		}
	})
	if w.overrideValue(t) {
		t.Error("the rolled-back runtime transactions changed the override")
	}
}

// aut6DisabledOwnerRawDenied (AUT-6, amendment01 A2): a second owner f (promoted by o through the facade) is disabled
// by the emergency DisableUser, which keeps its bindings; f's raw runtime UPDATE of the override is refused by the
// enabled-owner check of the guard (users FOR SHARE, status active) with 42501 group_override_owner, and the value is
// unchanged. The active owner's raw positive is AUT-3.
func aut6DisabledOwnerRawDenied(t *testing.T) {
	w := newOvrWorld(t, "AUT6")
	ctx := context.Background()
	c, tc := w.e.as(t, w.o, w.a)
	if _, err := w.e.svc.PromoteToOwner(c, tc, memOrgScope, w.f.id); err != nil {
		t.Fatalf("promote f to a second owner: %v", err)
	}
	if err := w.e.id.DisableUser(ctx, w.f.id.String()); err != nil {
		t.Fatalf("disable f: %v", err)
	}
	if r := w.e.memRole(t, w.a.String(), w.f.id); r != string(tenancy.OrgOwner) {
		t.Fatalf("the disabled owner's binding: %s, want it retained as %s", r, tenancy.OrgOwner)
	}
	rt := w.e.rlsRuntime(t)
	defer rt.Close(ctx)
	f, a := w.f.id.String(), w.a.String()
	rlsTx(t, rt, &f, &a, ptr(""), nil, func(tx pgx.Tx) {
		st, cn, n := rlsStmtC(t, tx, `UPDATE app.organizations SET group_management_override = true WHERE id = $1::text::uuid`, a)
		if st != "42501" || cn != "group_override_owner" || n != 0 {
			t.Errorf("disabled owner direct UPDATE: SQLSTATE %q constraint %q rows %d, want 42501 group_override_owner", st, cn, n)
		}
	})
	if w.overrideValue(t) {
		t.Error("the refused raw UPDATE changed the override")
	}
}

func aut4GenericPolicyPatch(t *testing.T) {
	w := newOvrWorld(t, "AUT4")
	c, tc := w.e.as(t, w.o, w.a)
	p, err := w.e.svc.SetOrganizationPolicy(c, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false, GroupManagementOverride: true})
	if err != nil || p.GroupManagementOverride || w.overrideValue(t) {
		t.Errorf("generic policy patch: %+v %v, override %t (must stay false)", p, err, w.overrideValue(t))
	}
}

func aut5DefaultFalse(t *testing.T) {
	w := newOvrWorld(t, "AUT5")
	if w.overrideValue(t) {
		t.Error("a fresh Organization has the override enabled")
	}
	c, tc := w.e.as(t, w.f, w.a)
	_, err := w.e.svc.ListMembers(c, tc, w.scope(), "", 0, "")
	memWantCode(t, "admin lists Group members by default", err, CodeForbidden, "")
	if _, err := w.e.svc.GetGroupOnboarding(c, tc, w.g.ID); err != nil {
		t.Errorf("admin reads Group metadata by default: %v", err)
	}
}

func memOvr1EnabledAdministers(t *testing.T) {
	w := newOvrWorld(t, "MEMOVR1")
	w.setOverride(t, true)
	c, tc := w.e.as(t, w.f, w.a)
	if _, err := w.e.svc.ListMembers(c, tc, w.scope(), "", 0, ""); err != nil {
		t.Fatalf("enabled admin lists members: %v", err)
	}
	if _, err := w.e.svc.RenameGroup(c, tc, w.g.ID, "MEMOVR1 renamed"); err != nil {
		t.Errorf("enabled admin renames: %v", err)
	}
	if _, err := w.e.svc.ChangeMemberRole(c, tc, w.scope(), w.m.id, string(tenancy.GroupAdmin)); err != nil {
		t.Errorf("enabled admin promotes a member to admin: %v", err)
	}
	if _, err := w.e.svc.IssueInvitation(c, tc, w.scope(), string(tenancy.GroupMember), Recipient{UserID: &w.u.id}, 0); err != nil {
		t.Errorf("enabled admin invites: %v", err)
	}
	if n := w.events(t, "access.override_used", " AND actor_user_id = $3::text::uuid AND resource_type = 'group' AND group_id = $4::text::uuid",
		w.f.id.String(), w.g.ID.String()); n != 3 {
		t.Errorf("access.override_used rows: %d, want 3 (rename, role change, invitation)", n)
	}
	g, err := w.e.svc.GetGroupOnboarding(c, tc, w.g.ID)
	if err != nil || g.SelfRole != "" || !g.CanManage || g.ManagementSource != managementOrganizationOverride {
		t.Errorf("hints under the override: %+v %v", g, err)
	}
	_, err = w.e.svc.SelfMembership(c, tc, w.scope())
	memWantCode(t, "override is not a self membership", err, CodeNotFound, "")
	// The Group owner's own administration records no override row.
	oc, otc := w.e.as(t, w.o, w.a)
	if _, err := w.e.svc.RenameGroup(oc, otc, w.g.ID, "MEMOVR1 by owner"); err != nil {
		t.Errorf("Group owner renames: %v", err)
	}
	if n := w.events(t, "access.override_used", " AND actor_user_id = $3::text::uuid", w.o.id.String()); n != 0 {
		t.Errorf("Group owner recorded %d override rows", n)
	}
}

func memOvr2Exclusions(t *testing.T) {
	w := newOvrWorld(t, "MEMOVR2")
	w.setOverride(t, true)
	c, tc := w.e.as(t, w.f, w.a)
	_, err := w.e.svc.IssueInvitation(c, tc, w.scope(), string(tenancy.GroupMember), Recipient{UserID: &w.f.id}, 0)
	memWantCode(t, "self-invitation", err, CodeInvalidInput, fieldSelfTarget)
	_, err = w.e.svc.IssueInvitation(c, tc, w.scope(), string(tenancy.GroupOwner), Recipient{UserID: &w.u.id}, 0)
	memWantCode(t, "owner-role invitation", err, CodeInvalidInput, string(tenancy.CodeInvalidRole))
	_, err = w.e.svc.PromoteToOwner(c, tc, w.scope(), w.m.id)
	if err == nil {
		t.Error("the override promoted to Group owner")
	}
	err = w.e.svc.LeaveScope(c, tc, w.scope())
	memWantCode(t, "leave without a binding", err, CodeForbidden, "")
	err = w.e.svc.RemoveMember(c, tc, w.scope(), w.o.id)
	memWantCode(t, "remove the Group owner", err, CodeForbidden, fieldOwnerSubject)
}

func memOvr3DisableDenies(t *testing.T) {
	w := newOvrWorld(t, "MEMOVR3")
	w.setOverride(t, true)
	c, tc := w.e.as(t, w.f, w.a)
	if _, err := w.e.svc.ListMembers(c, tc, w.scope(), "", 0, ""); err != nil {
		t.Fatalf("enabled: %v", err)
	}
	w.setOverride(t, false)
	_, err := w.e.svc.ListMembers(c, tc, w.scope(), "", 0, "")
	memWantCode(t, "after disable", err, CodeForbidden, "")
	// Genuine Group membership survives turning the override off.
	if r := w.e.memRole(t, w.g.ID.String(), w.m.id); r != string(tenancy.GroupMember) {
		t.Errorf("member role after disable: %s", r)
	}
}

func memOvr5SelfIntent(t *testing.T) {
	w := newOvrWorld(t, "MEMOVR5")
	// A legacy self-targeted intent (issuer = recipient = m) inserted by the superuser setup connection.
	token, digest, err := tenancy.NewInvitationToken()
	if err != nil {
		t.Fatal(err)
	}
	var id string
	if err := w.e.setup.QueryRow(context.Background(), `INSERT INTO app.invitations (id, organization_id, scope_type, scope_id, group_id, role,
  issuer_user_id, intended_user_id, token_digest, expires_at) VALUES (gen_random_uuid(), $1::text::uuid, 'group', $2::text::uuid,
  $2::text::uuid, 'group_admin', $3::text::uuid, $3::text::uuid, $4, now() + interval '1 day') RETURNING id::text`,
		w.a.String(), w.g.ID.String(), w.m.id.String(), digest).Scan(&id); err != nil {
		t.Fatalf("legacy intent: %v", err)
	}
	iid, _ := tenancy.ParseInvitationID(id)
	c, tc := w.e.as(t, w.m, w.a)
	_, err = w.e.svc.AcceptGroupInvitation(c, tc, iid, token)
	memWantCode(t, "self-issued claim", err, CodeNotFound, "")
	if n := w.e.memCount(t, `SELECT count(*) FROM app.invitations WHERE id = $1::text::uuid AND consumed_at IS NULL`, id); n != 1 {
		t.Error("the self-issued intent was consumed")
	}
	if r := w.e.memRole(t, w.g.ID.String(), w.m.id); r != string(tenancy.GroupMember) {
		t.Errorf("role after the refused claim: %s", r)
	}
}

func memOvr6NewMemberConsumes(t *testing.T) {
	w := newOvrWorld(t, "MEMOVR6")
	w.setOverride(t, true)
	c, tc := w.e.as(t, w.f, w.a)
	iss, err := w.e.svc.IssueInvitation(c, tc, w.scope(), string(tenancy.GroupMember), Recipient{UserID: &w.u.id}, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	uc, utc := w.e.as(t, w.u, w.a)
	out, err := w.e.svc.AcceptGroupInvitation(uc, utc, iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != tenancy.ClaimJoined {
		t.Fatalf("accept: %+v %v", out, err)
	}
	// The claim grants exactly one binding, in the Group (the fixture's Organization grant of u is not the claim's).
	if n := w.events(t, "membership.granted", " AND subject_user_id = $3::text::uuid AND group_id = $4::text::uuid", w.u.id.String(),
		w.g.ID.String()); n != 1 {
		t.Errorf("membership.granted in the Group: %d, want 1", n)
	}
}

func memOvr7OwnerAndLeave(t *testing.T) {
	w := newOvrWorld(t, "MEMOVR7")
	w.setOverride(t, true)
	mc, mtc := w.e.as(t, w.m, w.a)
	if err := w.e.svc.LeaveScope(mc, mtc, w.scope()); err != nil {
		t.Errorf("member leaves: %v", err)
	}
}

func memOvr8IndependentUpgrade(t *testing.T) {
	w := newOvrWorld(t, "MEMOVR8")
	w.setOverride(t, true)
	c, tc := w.e.as(t, w.f, w.a)
	iss, err := w.e.svc.IssueInvitation(c, tc, w.scope(), string(tenancy.GroupAdmin), Recipient{UserID: &w.m.id}, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	var before string
	if err := w.e.setup.QueryRow(context.Background(), `SELECT id::text FROM app.role_bindings WHERE scope_type = 'group'
  AND scope_id = $1::text::uuid AND subject_user_id = $2::text::uuid`, w.g.ID.String(), w.m.id.String()).Scan(&before); err != nil {
		t.Fatal(err)
	}
	counts := func() [3]int {
		return [3]int{w.events(t, "invitation.consumed", ""), w.events(t, "membership.role_changed", ""), w.events(t, "membership.granted", "")}
	}
	base := counts() // correction02 C02-2 deltas: consumed +1, role_changed +1, granted +0 (the fixture's grants precede)
	mc, mtc := w.e.as(t, w.m, w.a)
	out, err := w.e.svc.AcceptGroupInvitation(mc, mtc, iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != tenancy.ClaimUpgraded || out.Binding == nil || out.Binding.ID.String() != before ||
		out.Binding.Role != string(tenancy.GroupAdmin) {
		t.Fatalf("upgrade: %+v %v", out, err)
	}
	if got := counts(); got != [3]int{base[0] + 1, base[1] + 1, base[2]} {
		t.Errorf("consumed/role_changed/granted: %v from %v, want +1 +1 +0", got, base)
	}
	if n := w.events(t, "access.override_used", " AND operation = 'invitation.issue'"); n != 1 {
		t.Errorf("override_used for the issuance: %d", n)
	}
	out, err = w.e.svc.AcceptGroupInvitation(mc, mtc, iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != tenancy.ClaimAlreadyMember {
		t.Errorf("replay: %+v %v", out, err)
	}
}

func oOvr2StaleOverrideIssuer(t *testing.T) {
	w := newOvrWorld(t, "OOVR2")
	w.setOverride(t, true)
	c, tc := w.e.as(t, w.f, w.a)
	iss, err := w.e.svc.IssueInvitation(c, tc, w.scope(), string(tenancy.GroupMember), Recipient{UserID: &w.u.id}, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	w.setOverride(t, false)
	uc, utc := w.e.as(t, w.u, w.a)
	_, err = w.e.svc.AcceptGroupInvitation(uc, utc, iss.Invitation.ID, iss.Token)
	memWantCode(t, "claim after the override was disabled", err, CodeConflict, string(tenancy.CodeIssuerAuthorityLost))
	if n := w.e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE scope_type = 'group' AND scope_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid`, w.g.ID.String(), w.u.id.String()); n != 0 {
		t.Error("a stale override invitation granted a binding")
	}
}

func oOvr4IssuerDisabled(t *testing.T) {
	w := newOvrWorld(t, "OOVR4")
	w.setOverride(t, true)
	c, tc := w.e.as(t, w.f, w.a)
	iss, err := w.e.svc.IssueInvitation(c, tc, w.scope(), string(tenancy.GroupMember), Recipient{UserID: &w.u.id}, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := w.e.setup.Exec(context.Background(), `UPDATE app.users SET status = 'disabled', disabled_at = now() WHERE id = $1::text::uuid`,
		w.f.id.String()); err != nil {
		t.Fatalf("disable issuer: %v", err)
	}
	uc, utc := w.e.as(t, w.u, w.a)
	_, err = w.e.svc.AcceptGroupInvitation(uc, utc, iss.Invitation.ID, iss.Token)
	memWantCode(t, "claim after the issuer was disabled", err, CodeConflict, string(tenancy.CodeIssuerAuthorityLost))
}
