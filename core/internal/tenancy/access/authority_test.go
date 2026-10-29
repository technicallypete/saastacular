package access

import (
	"testing"

	"core/internal/tenancy"
)

// G060 unit evidence (Plan r0 §5.2 T1/T4/T9, r1 C5/C7): the owner-only override write capability, the Group display
// hints and the override allow reason, decided on synthetic facts. The database guard, the audited write and the live
// revocation run in authority_integration_test.go / membership_integration_test.go (b2).

// TestAuthorityOwnerOnlySetting: organization.manage_policy (the override write's capability) allows exactly the
// current Organization owner, whether or not it holds a Group role, in either context; an Organization admin (also as
// a Group owner), a Group owner or admin and an ordinary member are refused, and the refusal is forbidden (the
// Organization stays readable) rather than not_found.
func TestAuthorityOwnerOnlySetting(t *testing.T) {
	for _, c := range []struct {
		name   string
		org    tenancy.OrganizationRole
		groups []tenancy.BindingFact
		want   bool
	}{
		{"org_owner without Group role", tenancy.OrgOwner, nil, true},
		{"org_owner who is a Group member", tenancy.OrgOwner, []tenancy.BindingFact{groupBinding(g1, tenancy.GroupMember)}, true},
		{"org_admin", tenancy.OrgAdmin, nil, false},
		{"org_admin who owns a Group", tenancy.OrgAdmin, []tenancy.BindingFact{groupBinding(g1, tenancy.GroupOwner)}, false},
		{"Group owner", tenancy.OrgMember, []tenancy.BindingFact{groupBinding(g1, tenancy.GroupOwner)}, false},
		{"Group admin", tenancy.OrgMember, []tenancy.BindingFact{groupBinding(g3, tenancy.GroupAdmin)}, false},
		{"org_member", tenancy.OrgMember, nil, false},
	} {
		for _, override := range []bool{false, true} {
			for _, sel := range []contextSelection{orgOnly, selG1} {
				f := factsA(c.org, c.groups...)
				f.Policy.GroupManagementOverride = override
				if contextCheck(sel, f) != "" {
					continue // a selected Group the subject cannot select never reaches the decision
				}
				ok, r := decide(capOrganizationManagePolicy, sel, f, targetOrgA)
				if ok != c.want {
					t.Errorf("%s (override %t, sel %v): manage_policy %v %s, want %v", c.name, override, sel.Group, ok, r, c.want)
				}
				if !ok {
					if fl := denial(capOrganizationManagePolicy, sel, f, Target{Kind: kindOrganization, Organization: orgA}, r); fl.code != CodeForbidden {
						t.Errorf("%s: refusal %s, want forbidden", c.name, fl.code)
					}
				}
			}
		}
	}
	// The generic policy capability is the same owner-only one; the override itself is never an input to it.
	f := factsA(tenancy.OrgAdmin)
	f.Policy.GroupManagementOverride = true
	if ok, _ := decide(capOrganizationManagePolicy, orgOnly, f, targetOrgA); ok {
		t.Error("an enabled override let an Organization admin change policy")
	}
}

// TestAuthorityGroupHints (r1 C7): the four nonmember/member by override cases of the Group hints, plus the explicit
// Group roles and the ordinary member; an override is never a self membership and never a self role.
func TestAuthorityGroupHints(t *testing.T) {
	g := GroupOnboarding{Group: tenancy.Group{ID: g3, OrganizationID: orgA}}
	for _, c := range []struct {
		name     string
		org      tenancy.OrganizationRole
		groups   []tenancy.BindingFact
		override bool
		self     tenancy.GroupRole
		manage   bool
		source   string
	}{
		{"org_admin nonmember, override off", tenancy.OrgAdmin, nil, false, "", false, ""},
		{"org_admin nonmember, override on", tenancy.OrgAdmin, nil, true, "", true, managementOrganizationOverride},
		{"org_admin member, override off", tenancy.OrgAdmin, []tenancy.BindingFact{groupBinding(g3, tenancy.GroupMember)}, false,
			tenancy.GroupMember, false, ""},
		{"org_admin member, override on", tenancy.OrgAdmin, []tenancy.BindingFact{groupBinding(g3, tenancy.GroupMember)}, true,
			tenancy.GroupMember, true, managementOrganizationOverride},
		{"org_owner nonmember, override on", tenancy.OrgOwner, nil, true, "", true, managementOrganizationOverride},
		{"Group admin, override off", tenancy.OrgMember, []tenancy.BindingFact{groupBinding(g3, tenancy.GroupAdmin)}, false,
			tenancy.GroupAdmin, true, managementGroupRole},
		{"org_admin Group owner, override on", tenancy.OrgAdmin, []tenancy.BindingFact{groupBinding(g3, tenancy.GroupOwner)}, true,
			tenancy.GroupOwner, true, managementGroupRole},
		{"org_member, override on", tenancy.OrgMember, nil, true, "", false, ""},
	} {
		f := factsA(c.org, c.groups...)
		f.Policy.GroupManagementOverride = c.override
		got := withAuthority(g, orgOnly, f)
		if got.SelfRole != c.self || got.CanManage != c.manage || got.ManagementSource != c.source {
			t.Errorf("%s: self %q manage %t source %q, want %q %t %q", c.name, got.SelfRole, got.CanManage, got.ManagementSource,
				c.self, c.manage, c.source)
		}
	}
}

// TestAuthorityOverrideReason (r1 C5): the override's allow reason marks exactly the writes that would be refused on the
// explicit Group path, for every ordinary administration capability; a caller who also qualifies through its own Group
// role is allowed with the ordinary reason (so it records no access.override_used).
func TestAuthorityOverrideReason(t *testing.T) {
	groupCaps := []Capability{capGroupUpdate, capGroupManageMembers, capGroupCompleteOnboarding}
	staff := factsA(tenancy.OrgAdmin, groupBinding(g1, tenancy.GroupAdmin))
	staff.Policy.GroupManagementOverride = true
	for _, c := range groupCaps {
		if ok, r := decide(c, orgOnly, staff, targetG(g1)); !ok || r != reasonAllowed {
			t.Errorf("%s on own admin Group: %v %s, want allowed", c, ok, r)
		}
		if ok, r := decide(c, orgOnly, staff, targetG(g2)); !ok || r != reasonAllowedOverride {
			t.Errorf("%s on an unbound Group: %v %s, want allowed_override", c, ok, r)
		}
	}
	// g1 and g2 are linked to p1: the linked Group path (flag on) wins; p3 is unlinked: only the override reaches it.
	if ok, r := decide(capProfileUpdate, orgOnly, staff, targetP(p1)); !ok || r != reasonAllowed {
		t.Errorf("profile update via linked admin Group: %v %s", ok, r)
	}
	for _, c := range []Capability{capProfileUpdate, capProfileSelect} {
		if ok, r := decide(c, orgOnly, staff, targetP(p3)); !ok || r != reasonAllowedOverride {
			t.Errorf("%s of an unlinked profile: %v %s, want allowed_override", c, ok, r)
		}
	}
	// The flag gates only the linked path; the override path ignores it (as the Organization path always did).
	staff.Policy.GroupAdminsEditBusinessProfiles = false
	if ok, r := decide(capProfileUpdate, orgOnly, staff, targetP(p1)); !ok || r != reasonAllowedOverride {
		t.Errorf("profile update with the flag off: %v %s, want allowed_override", ok, r)
	}
	// Reads never carry the override reason: metadata is the ordinary Organization path.
	for _, c := range []Capability{capGroupRead, capProfileRead} {
		tg := targetG(g2)
		if c == capProfileRead {
			tg = targetP(p3)
		}
		if ok, r := decide(c, orgOnly, staff, tg); !ok || r != reasonAllowed {
			t.Errorf("%s: %v %s, want allowed", c, ok, r)
		}
	}
	if !allowed(reasonAllowed) || !allowed(reasonAllowedOverride) || allowed(reasonRoleNotPermitted) {
		t.Error("allowed() misclassifies a reason")
	}
}
