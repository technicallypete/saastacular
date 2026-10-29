package access

import (
	"reflect"
	"regexp"
	"slices"
	"testing"

	"core/internal/tenancy"
)

// U-POL-GATE, U-POL-SCOPE and U-POL-MFA (G007 Plan r1 §6.1): the gates, scope rules and inputs of the evaluator. The
// independent literal matrix is matrix_test.go (written from the Spec by the oracle delegate).

func uid[T ~[16]byte](b byte) T {
	var x T
	x[0], x[6], x[8], x[15] = b, 0x40, 0x80, 1
	return x
}

// The fixture of Plan r1 §6.3 at unit level: Organizations A and B; Groups g1, g2, g3 and the new unattached gNew in A,
// gB in B; profiles p1 (shared by g1 and g2), p2 (on g3), p3 (unlinked) in A and pB (on gB) in B.
var (
	orgA, orgB            = uid[tenancy.OrganizationID](0xa1), uid[tenancy.OrganizationID](0xb1)
	g1, g2, g3, gNew, gB  = uid[tenancy.GroupID](0x01), uid[tenancy.GroupID](0x02), uid[tenancy.GroupID](0x03), uid[tenancy.GroupID](0x04), uid[tenancy.GroupID](0x0b)
	p1, p2, p3, pB        = uid[tenancy.BusinessProfileID](0x11), uid[tenancy.BusinessProfileID](0x12), uid[tenancy.BusinessProfileID](0x13), uid[tenancy.BusinessProfileID](0x1b)
	allCapabilities       = []Capability{capOrganizationRead, capOrganizationManageMembers, capOrganizationManagePolicy, capGroupRead, capGroupUpdate, capGroupManageMembers, capProfileRead, capProfileUpdate, capProfileSelect, capProfilePrivileged, capAuditRead, capOrganizationManageDomains, capOrganizationManageBranding, capGroupCreate, capGroupCompleteOnboarding, capOrganizationManageOwners, capGroupManageOwners, capOrganizationLeave, capGroupLeave, capLabelOrganizationManage, capLabelGroupManage, capLabelGroupRead}
	orgOnly, selG1, selG3 = contextSelection{Organization: orgA}, contextSelection{Organization: orgA, Group: g1}, contextSelection{Organization: orgA, Group: g3}
	targetOrgA            = Target{Kind: kindOrganization, Organization: orgA}
	targetG               = func(g tenancy.GroupID) Target { return Target{Kind: kindGroup, Group: g} }
	targetP               = func(p tenancy.BusinessProfileID) Target { return Target{Kind: kindProfile, Profile: p} }
	groupOf               = map[tenancy.GroupID]tenancy.OrganizationID{g1: orgA, g2: orgA, g3: orgA, gNew: orgA, gB: orgB}
	profileOf             = map[tenancy.BusinessProfileID]tenancy.OrganizationID{p1: orgA, p2: orgA, p3: orgA, pB: orgB}
	fixtureAssociations   = []tenancy.AssociationFact{{GroupID: g1, BusinessProfileID: p1}, {GroupID: g2, BusinessProfileID: p1}, {GroupID: g3, BusinessProfileID: p2}}
	fixtureGroupsA        = []tenancy.Group{{ID: g1, OrganizationID: orgA}, {ID: g2, OrganizationID: orgA}, {ID: g3, OrganizationID: orgA}, {ID: gNew, OrganizationID: orgA}}
	fixtureProfilesA      = []tenancy.BusinessProfile{{ID: p1, OrganizationID: orgA}, {ID: p2, OrganizationID: orgA}, {ID: p3, OrganizationID: orgA}}
	membershipSource      = tenancy.SourceMembership
	groupBinding          = func(g tenancy.GroupID, role tenancy.GroupRole) tenancy.BindingFact {
		return tenancy.BindingFact{OrganizationID: groupOf[g], GroupID: g, Role: string(role), Source: membershipSource}
	}
)

// factsA returns Organization A's facts for a subject with the given Organization role ("" for none) and Group roles.
func factsA(orgRole tenancy.OrganizationRole, groups ...tenancy.BindingFact) Facts {
	f := Facts{
		Organization: tenancy.Organization{ID: orgA}, Policy: tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: true},
		GroupBindings: groups, Groups: fixtureGroupsA, Associations: fixtureAssociations, Profiles: fixtureProfilesA,
	}
	if orgRole != "" {
		f.OrganizationBindings = []tenancy.BindingFact{{OrganizationID: orgA, Role: string(orgRole), Source: membershipSource}}
	}
	return f
}

func wantDecision(t *testing.T, what string, c Capability, ctx contextSelection, f Facts, tg Target, allow bool, reason Reason) {
	t.Helper()
	tg.Organization = orgA
	if tg.Kind == "" {
		tg.Organization = tenancy.OrganizationID{}
	}
	ok, r := decide(c, ctx, f, tg)
	if ok != allow || r != reason {
		t.Errorf("%s: %s -> (%v, %s), want (%v, %s)", what, c, ok, r, allow, reason)
	}
}

func targetFor(c Capability, g tenancy.GroupID, p tenancy.BusinessProfileID) Target {
	switch k, _ := capabilityKind(c); k {
	case kindGroup:
		return targetG(g)
	case kindProfile:
		return targetP(p)
	}
	return targetOrgA
}

// U-POL-GATE: the Organization gate, unknown capabilities, roles, sources and kinds, empty and inconsistent facts, and
// target ownership and kind.
func TestPolicyGates(t *testing.T) {
	t.Run("no Organization binding denies all 11", func(t *testing.T) {
		// Even with every Group role and link a Group-role subject could hold: a Group binding never establishes tenant
		// access.
		f := factsA("", groupBinding(g1, tenancy.GroupOwner), groupBinding(g3, tenancy.GroupAdmin))
		for _, c := range allCapabilities {
			for _, ctx := range []contextSelection{orgOnly, selG1} {
				wantDecision(t, "group bindings only", c, ctx, f, targetFor(c, g1, p1), false, reasonOrganizationNotPermitted)
			}
		}
	})
	t.Run("empty, nil and foreign facts", func(t *testing.T) {
		for name, f := range map[string]Facts{
			"zero facts":          {},
			"Organization absent": {OrganizationBindings: factsA(tenancy.OrgOwner).OrganizationBindings},
			"other Organization":  func() Facts { f := factsA(tenancy.OrgOwner); f.Organization.ID = orgB; return f }(),
			"nil binding slice":   func() Facts { f := factsA(tenancy.OrgOwner); f.OrganizationBindings = nil; return f }(),
			"empty binding slice": func() Facts {
				f := factsA(tenancy.OrgOwner)
				f.OrganizationBindings = []tenancy.BindingFact{}
				return f
			}(),
		} {
			for _, c := range allCapabilities {
				wantDecision(t, name, c, orgOnly, f, targetFor(c, g1, p1), false, reasonOrganizationNotPermitted)
			}
		}
		for _, c := range allCapabilities {
			wantDecision(t, "zero context", c, contextSelection{}, factsA(tenancy.OrgOwner), targetFor(c, g1, p1), false, reasonOrganizationNotPermitted)
		}
	})
	t.Run("unknown capabilities, no prefix matching or ranking", func(t *testing.T) {
		for _, c := range []Capability{"", "organization", "organization.*", "organization.read.all", "ORGANIZATION.READ", " organization.read",
			"group", "business_profile", "business_profile.ownership.transfer", "billing.manage", "audit.write", "*", "admin"} {
			wantDecision(t, "unknown "+string(c), c, orgOnly, factsA(tenancy.OrgOwner), targetOrgA, false, reasonCapabilityUnknown)
		}
	})
	t.Run("unknown or inconsistent Organization bindings deny everything", func(t *testing.T) {
		mutate := map[string]func(*tenancy.BindingFact){
			"unknown role":       func(b *tenancy.BindingFact) { b.Role = "org_superadmin" },
			"role with space":    func(b *tenancy.BindingFact) { b.Role = "org_owner " },
			"uppercase role":     func(b *tenancy.BindingFact) { b.Role = "ORG_OWNER" },
			"Group role as org":  func(b *tenancy.BindingFact) { b.Role = "group_owner" },
			"empty role":         func(b *tenancy.BindingFact) { b.Role = "" },
			"invitation source":  func(b *tenancy.BindingFact) { b.Source = "invitation" },
			"system source":      func(b *tenancy.BindingFact) { b.Source = "system" },
			"empty source":       func(b *tenancy.BindingFact) { b.Source = "" },
			"foreign binding":    func(b *tenancy.BindingFact) { b.OrganizationID = orgB },
			"binding with Group": func(b *tenancy.BindingFact) { b.GroupID = g1 },
		}
		for name, m := range mutate {
			f := factsA(tenancy.OrgOwner)
			b := f.OrganizationBindings[0]
			m(&b)
			f.OrganizationBindings = []tenancy.BindingFact{b}
			for _, c := range allCapabilities {
				wantDecision(t, name, c, orgOnly, f, targetFor(c, g1, p1), false, reasonFactInconsistent)
			}
		}
		f := factsA(tenancy.OrgOwner)
		f.OrganizationBindings = append(f.OrganizationBindings, tenancy.BindingFact{OrganizationID: orgA, Role: "org_member", Source: membershipSource})
		for _, c := range allCapabilities {
			wantDecision(t, "two Organization bindings", c, orgOnly, f, targetFor(c, g1, p1), false, reasonFactInconsistent)
		}
	})
	t.Run("unknown or foreign Group bindings confer nothing", func(t *testing.T) {
		for name, b := range map[string]tenancy.BindingFact{
			"unknown Group role":  {OrganizationID: orgA, GroupID: g1, Role: "group_superowner", Source: membershipSource},
			"Org role on a Group": {OrganizationID: orgA, GroupID: g1, Role: "org_owner", Source: membershipSource},
			"system source":       {OrganizationID: orgA, GroupID: g1, Role: "group_owner", Source: "system"},
			"foreign binding":     {OrganizationID: orgB, GroupID: g1, Role: "group_owner", Source: membershipSource},
			"no Group":            {OrganizationID: orgA, Role: "group_owner", Source: membershipSource},
		} {
			f := factsA(tenancy.OrgMember, b)
			for _, c := range []Capability{capGroupRead, capGroupUpdate, capGroupManageMembers} {
				wantDecision(t, name, c, orgOnly, f, targetG(g1), false, reasonRoleNotPermitted)
			}
			for _, c := range []Capability{capProfileRead, capProfileUpdate, capProfileSelect} {
				wantDecision(t, name, c, orgOnly, f, targetP(p1), false, reasonRoleNotPermitted)
			}
		}
		f := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner), groupBinding(g1, tenancy.GroupMember))
		wantDecision(t, "two bindings of one Group", capGroupRead, orgOnly, f, targetG(g1), false, reasonFactInconsistent)
	})
	t.Run("an inconsistent association confers nothing", func(t *testing.T) {
		f := factsA(tenancy.OrgMember, groupBinding(g3, tenancy.GroupOwner))
		f.Associations = append(slices.Clone(f.Associations), tenancy.AssociationFact{GroupID: g3, BusinessProfileID: p3})
		for _, p := range []tenancy.BusinessProfileID{p2, p3} {
			wantDecision(t, "Group with two profiles", capProfileRead, orgOnly, f, targetP(p), false, reasonNoLinkedPath)
		}
	})
	t.Run("target kind mismatch and unknown kind", func(t *testing.T) {
		f := factsA(tenancy.OrgOwner)
		for _, c := range allCapabilities {
			k, _ := capabilityKind(c)
			for _, tg := range []Target{targetOrgA, targetG(g1), targetP(p1)} {
				if tg.Kind == k {
					continue
				}
				wantDecision(t, "mismatch "+string(tg.Kind), c, orgOnly, f, tg, false, reasonTargetKindMismatch)
			}
			wantDecision(t, "unknown kind", c, orgOnly, f, Target{Kind: "tenant", Group: g1, Profile: p1}, false, reasonFactInconsistent)
			wantDecision(t, "empty kind", c, orgOnly, f, Target{}, false, reasonFactInconsistent)
		}
	})
	t.Run("foreign, unknown and zero targets", func(t *testing.T) {
		f := factsA(tenancy.OrgOwner)
		f.Groups = append(slices.Clone(f.Groups), tenancy.Group{ID: gB, OrganizationID: orgB})
		f.Profiles = append(slices.Clone(f.Profiles), tenancy.BusinessProfile{ID: pB, OrganizationID: orgB})
		unknownG, unknownP := uid[tenancy.GroupID](0x7e), uid[tenancy.BusinessProfileID](0x7e)
		for _, c := range []Capability{capGroupRead, capGroupUpdate, capGroupManageMembers, capGroupCompleteOnboarding, capLabelGroupManage, capLabelGroupRead} {
			for _, g := range []tenancy.GroupID{gB, unknownG, {}} {
				wantDecision(t, "Group target", c, orgOnly, f, targetG(g), false, reasonTargetNotFound)
			}
		}
		for _, c := range []Capability{capProfileRead, capProfileUpdate, capProfileSelect, capProfilePrivileged} {
			for _, p := range []tenancy.BusinessProfileID{pB, unknownP, {}} {
				wantDecision(t, "profile target", c, orgOnly, f, targetP(p), false, reasonTargetNotFound)
			}
		}
		for _, c := range []Capability{capOrganizationRead, capOrganizationManagePolicy, capAuditRead, capOrganizationManageDomains, capOrganizationManageBranding, capGroupCreate, capLabelOrganizationManage} {
			for _, o := range []tenancy.OrganizationID{orgB, {}} {
				if ok, r := decide(c, orgOnly, f, Target{Kind: kindOrganization, Organization: o}); ok || r != reasonTargetNotFound {
					t.Errorf("Organization target: %s -> (%v, %s)", c, ok, r)
				}
			}
		}
	})
	t.Run("store errors and cancellation never allow", func(t *testing.T) {
		// A failed fact read yields no facts at all (the facade never evaluates partial facts): zero facts deny.
		for _, c := range allCapabilities {
			wantDecision(t, "zero facts", c, orgOnly, Facts{}, targetFor(c, g1, p1), false, reasonOrganizationNotPermitted)
		}
	})
}

// U-POL-SCOPE: selected-Group restriction, independent Organization paths, unlinked profiles, the new unattached Group,
// the edit flag and the manageable-set rules.
func TestPolicyScope(t *testing.T) {
	t.Run("selected Group restricts Group-derived profile paths", func(t *testing.T) {
		f := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupAdmin), groupBinding(g3, tenancy.GroupAdmin))
		wantDecision(t, "org-only via g1", capProfileUpdate, orgOnly, f, targetP(p1), true, reasonAllowed)
		wantDecision(t, "selected g3 cannot reach p1 through g1", capProfileUpdate, selG3, f, targetP(p1), false, reasonNoLinkedPath)
		wantDecision(t, "selected g3 cannot read p1 through g1", capProfileRead, selG3, f, targetP(p1), false, reasonNoLinkedPath)
		wantDecision(t, "selected g3 reaches p2", capProfileUpdate, selG3, f, targetP(p2), true, reasonAllowed)
		wantDecision(t, "selected g1 reaches p1", capProfileSelect, selG1, f, targetP(p1), true, reasonAllowed)
	})
	t.Run("Organization owner and admin paths are independent of the selected Group", func(t *testing.T) {
		// G060 (ADR-008): the metadata read is the Organization path regardless of the override; ordinary administration
		// is the Organization path only while the override is enabled (allowed_override), otherwise denied.
		for _, r := range []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin} {
			f := factsA(r)
			on := factsA(r)
			on.Policy.GroupManagementOverride = true
			for _, p := range []tenancy.BusinessProfileID{p1, p2, p3} {
				wantDecision(t, string(r)+" selected g3", capProfileRead, selG3, f, targetP(p), true, reasonAllowed)
				wantDecision(t, string(r)+" org-only", capProfileRead, orgOnly, f, targetP(p), true, reasonAllowed)
				for _, c := range []Capability{capProfileUpdate, capProfileSelect} {
					wantDecision(t, string(r)+" selected g3", c, selG3, f, targetP(p), false, reasonRoleNotPermitted)
					wantDecision(t, string(r)+" org-only", c, orgOnly, f, targetP(p), false, reasonRoleNotPermitted)
					wantDecision(t, string(r)+" selected g3, override", c, selG3, on, targetP(p), true, reasonAllowedOverride)
					wantDecision(t, string(r)+" org-only, override", c, orgOnly, on, targetP(p), true, reasonAllowedOverride)
				}
			}
			// Pin (b): the Group `all` path is restricted by a selected Group to that Group.
			wantDecision(t, string(r)+" selected g3 renames g1", capGroupUpdate, selG3, f, targetG(g1), false, reasonTargetOutsideContext)
			wantDecision(t, string(r)+" selected g3 renames g3", capGroupUpdate, selG3, f, targetG(g3), false, reasonRoleNotPermitted)
			wantDecision(t, string(r)+" org-only renames g1", capGroupUpdate, orgOnly, f, targetG(g1), false, reasonRoleNotPermitted)
			wantDecision(t, string(r)+" org-only reads g1", capGroupRead, orgOnly, f, targetG(g1), true, reasonAllowed)
			wantDecision(t, string(r)+" selected g3 renames g1, override", capGroupUpdate, selG3, on, targetG(g1), false, reasonTargetOutsideContext)
			wantDecision(t, string(r)+" selected g3 renames g3, override", capGroupUpdate, selG3, on, targetG(g3), true, reasonAllowedOverride)
			wantDecision(t, string(r)+" org-only renames g1, override", capGroupUpdate, orgOnly, on, targetG(g1), true, reasonAllowedOverride)
			wantDecision(t, string(r)+" org-only manages g1, override", capGroupManageMembers, orgOnly, on, targetG(g1), true, reasonAllowedOverride)
			wantDecision(t, string(r)+" no Group ownership, override", capGroupManageOwners, orgOnly, on, targetG(g1), false, reasonRoleNotPermitted)
			wantDecision(t, string(r)+" nothing to leave, override", capGroupLeave, orgOnly, on, targetG(g1), false, reasonRoleNotPermitted)
		}
	})
	t.Run("G060 the explicit Group path is decided before the override", func(t *testing.T) {
		f := factsA(tenancy.OrgAdmin, groupBinding(g1, tenancy.GroupAdmin), groupBinding(g3, tenancy.GroupMember))
		f.Policy.GroupManagementOverride = true
		wantDecision(t, "own admin Group", capGroupManageMembers, orgOnly, f, targetG(g1), true, reasonAllowed)
		wantDecision(t, "member Group", capGroupManageMembers, orgOnly, f, targetG(g3), true, reasonAllowedOverride)
		wantDecision(t, "unbound Group", capGroupUpdate, orgOnly, f, targetG(g2), true, reasonAllowedOverride)
		f.Policy.GroupManagementOverride = false
		wantDecision(t, "own admin Group, override off", capGroupManageMembers, orgOnly, f, targetG(g1), true, reasonAllowed)
		wantDecision(t, "member Group, override off", capGroupManageMembers, orgOnly, f, targetG(g3), false, reasonRoleNotPermitted)
		wantDecision(t, "member Group read, override off", capGroupRead, orgOnly, f, targetG(g3), true, reasonAllowed)
		member := factsA(tenancy.OrgMember)
		member.Policy.GroupManagementOverride = true
		wantDecision(t, "org_member, override", capGroupUpdate, orgOnly, member, targetG(g1), false, reasonRoleNotPermitted)
		wantDecision(t, "org_member read, override", capGroupRead, orgOnly, member, targetG(g1), false, reasonRoleNotPermitted)
	})
	t.Run("unlinked profile: no Group role reaches it", func(t *testing.T) {
		f := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner), groupBinding(g2, tenancy.GroupOwner), groupBinding(g3, tenancy.GroupOwner))
		for _, c := range []Capability{capProfileRead, capProfileUpdate, capProfileSelect} {
			wantDecision(t, "p3", c, orgOnly, f, targetP(p3), false, reasonNoLinkedPath)
		}
		wantDecision(t, "privileged settings", capProfilePrivileged, orgOnly, f, targetP(p1), false, reasonRoleNotPermitted)
	})
	t.Run("a new unattached Group authorizes no existing profile", func(t *testing.T) {
		f := factsA(tenancy.OrgMember, groupBinding(gNew, tenancy.GroupOwner))
		for _, p := range []tenancy.BusinessProfileID{p1, p2, p3} {
			wantDecision(t, "select", capProfileSelect, orgOnly, f, targetP(p), false, reasonNoLinkedPath)
			wantDecision(t, "update", capProfileUpdate, orgOnly, f, targetP(p), false, reasonNoLinkedPath)
		}
		wantDecision(t, "own Group", capGroupUpdate, orgOnly, f, targetG(gNew), true, reasonAllowed)
		wantDecision(t, "another Group", capGroupRead, orgOnly, f, targetG(g1), false, reasonRoleNotPermitted)
	})
	t.Run("the edit flag gates only Group-inherited updates", func(t *testing.T) {
		off := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner), groupBinding(g2, tenancy.GroupMember))
		off.Policy.GroupAdminsEditBusinessProfiles = false
		wantDecision(t, "owner update, flag off", capProfileUpdate, orgOnly, off, targetP(p1), false, reasonEditPolicyDisabled)
		wantDecision(t, "owner read, flag off", capProfileRead, orgOnly, off, targetP(p1), true, reasonAllowed)
		wantDecision(t, "owner select, flag off", capProfileSelect, orgOnly, off, targetP(p1), true, reasonAllowed)
		admin := factsA(tenancy.OrgMember, groupBinding(g3, tenancy.GroupAdmin))
		admin.Policy.GroupAdminsEditBusinessProfiles = false
		wantDecision(t, "admin update, flag off", capProfileUpdate, orgOnly, admin, targetP(p2), false, reasonEditPolicyDisabled)
		for _, r := range []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin} {
			f := factsA(r)
			f.Policy.GroupAdminsEditBusinessProfiles = false
			// G060: without the override the Organization path administers nothing; with it the flag does not apply.
			wantDecision(t, string(r)+" update, flag off", capProfileUpdate, orgOnly, f, targetP(p1), false, reasonRoleNotPermitted)
			f.Policy.GroupManagementOverride = true
			wantDecision(t, string(r)+" update, flag off, override", capProfileUpdate, orgOnly, f, targetP(p1), true, reasonAllowedOverride)
		}
	})
	t.Run("member-only paths allow read only", func(t *testing.T) {
		f := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupMember))
		wantDecision(t, "read", capProfileRead, orgOnly, f, targetP(p1), true, reasonAllowed)
		wantDecision(t, "update", capProfileUpdate, orgOnly, f, targetP(p1), false, reasonRoleNotPermitted)
		wantDecision(t, "select", capProfileSelect, orgOnly, f, targetP(p1), false, reasonRoleNotPermitted)
	})
	t.Run("manageable-set rules", func(t *testing.T) {
		set := func(ctx contextSelection, f Facts) []tenancy.BusinessProfileID {
			var out []tenancy.BusinessProfileID
			for _, p := range f.Profiles {
				if ok, _ := decide(capProfileSelect, ctx, f, Target{Kind: kindProfile, Organization: orgA, Profile: p.ID}); ok {
					out = append(out, p.ID)
				}
			}
			return out
		}
		// Owner of g1 and g2 (both linked to p1), member of g3 (p2): p1 once; member-only p2 and unlinked p3 excluded.
		f := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner), groupBinding(g2, tenancy.GroupAdmin), groupBinding(g3, tenancy.GroupMember))
		f.Profiles = append(slices.Clone(f.Profiles), tenancy.BusinessProfile{ID: pB, OrganizationID: orgB})
		if got := set(orgOnly, f); !slices.Equal(got, []tenancy.BusinessProfileID{p1}) {
			t.Errorf("org-only manageable set: %v", got)
		}
		if got := set(selG3, f); len(got) != 0 {
			t.Errorf("selected g3 manageable set: %v", got)
		}
		if got := set(selG1, f); !slices.Equal(got, []tenancy.BusinessProfileID{p1}) {
			t.Errorf("selected g1 manageable set: %v", got)
		}
		admin := factsA(tenancy.OrgAdmin)
		if got := set(selG3, admin); len(got) != 0 {
			t.Errorf("org admin manageable set without the override: %v", got)
		}
		admin.Policy.GroupManagementOverride = true
		if got := set(selG3, admin); !slices.Equal(got, []tenancy.BusinessProfileID{p1, p2, p3}) {
			t.Errorf("org admin manageable set with the override: %v", got)
		}
	})
	t.Run("Group-role and Organization paths combine per capability (pin a)", func(t *testing.T) {
		f := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner))
		wantDecision(t, "organization.read via org_member", capOrganizationRead, orgOnly, f, targetOrgA, true, reasonAllowed)
		wantDecision(t, "no manage_members via group_owner", capOrganizationManageMembers, orgOnly, f, targetOrgA, false, reasonRoleNotPermitted)
		wantDecision(t, "no policy via group_owner", capOrganizationManagePolicy, orgOnly, f, targetOrgA, false, reasonRoleNotPermitted)
		wantDecision(t, "no audit via group_owner", capAuditRead, orgOnly, f, targetOrgA, false, reasonRoleNotPermitted)
		wantDecision(t, "no domains via group_owner", capOrganizationManageDomains, orgOnly, f, targetOrgA, false, reasonRoleNotPermitted)
		wantDecision(t, "no branding via group_owner", capOrganizationManageBranding, orgOnly, f, targetOrgA, false, reasonRoleNotPermitted)
	})
	t.Run("G018 group.create: every Organization membership role, Organization-only context", func(t *testing.T) {
		for _, r := range []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin, tenancy.OrgMember} {
			f := factsA(r)
			wantDecision(t, string(r)+" org-only", capGroupCreate, orgOnly, f, targetOrgA, true, reasonAllowed)
			wantDecision(t, string(r)+" selected g1", capGroupCreate, selG1, f, targetOrgA, false, reasonRoleNotPermitted)
		}
		owner := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner))
		wantDecision(t, "group_owner selected own g1", capGroupCreate, selG1, owner, targetOrgA, false, reasonRoleNotPermitted)
		// A Group role without a current Organization binding never creates (superseding D-1).
		noOrg := factsA("", groupBinding(g1, tenancy.GroupOwner))
		wantDecision(t, "group_owner without Organization binding", capGroupCreate, orgOnly, noOrg, targetOrgA, false, reasonOrganizationNotPermitted)
	})
	t.Run("G018 group.complete_onboarding: all for Organization owner/admin, own for group owner/admin", func(t *testing.T) {
		for _, r := range []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin} {
			f := factsA(r)
			// G060: the Organization path completes onboarding only while the override is enabled.
			wantDecision(t, string(r)+" gNew", capGroupCompleteOnboarding, orgOnly, f, targetG(gNew), false, reasonRoleNotPermitted)
			wantDecision(t, string(r)+" selected g3 on g1", capGroupCompleteOnboarding, selG3, f, targetG(g1), false, reasonTargetOutsideContext)
			f.Policy.GroupManagementOverride = true
			wantDecision(t, string(r)+" gNew, override", capGroupCompleteOnboarding, orgOnly, f, targetG(gNew), true, reasonAllowedOverride)
			wantDecision(t, string(r)+" selected g3 on g1, override", capGroupCompleteOnboarding, selG3, f, targetG(g1), false, reasonTargetOutsideContext)
		}
		f := factsA(tenancy.OrgMember, groupBinding(gNew, tenancy.GroupOwner), groupBinding(g1, tenancy.GroupAdmin), groupBinding(g3, tenancy.GroupMember))
		wantDecision(t, "group_owner own gNew", capGroupCompleteOnboarding, orgOnly, f, targetG(gNew), true, reasonAllowed)
		wantDecision(t, "group_admin own g1", capGroupCompleteOnboarding, orgOnly, f, targetG(g1), true, reasonAllowed)
		wantDecision(t, "group_member g3", capGroupCompleteOnboarding, orgOnly, f, targetG(g3), false, reasonRoleNotPermitted)
		wantDecision(t, "unrelated g2", capGroupCompleteOnboarding, orgOnly, f, targetG(g2), false, reasonRoleNotPermitted)
		wantDecision(t, "org_member alone", capGroupCompleteOnboarding, orgOnly, factsA(tenancy.OrgMember), targetG(gNew), false, reasonRoleNotPermitted)
	})
	t.Run("G020 label capabilities: Organization labels for owners/admins, Group labels only through an own Group role", func(t *testing.T) {
		for _, r := range []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin} {
			for _, override := range []bool{false, true} {
				f := factsA(r)
				f.Policy.GroupManagementOverride = override
				what := string(r)
				if override {
					what += ", override"
				}
				wantDecision(t, what+" Organization labels", capLabelOrganizationManage, orgOnly, f, targetOrgA, true, reasonAllowed)
				wantDecision(t, what+" Organization labels, selected g3", capLabelOrganizationManage, selG3, f, targetOrgA, true, reasonAllowed)
				// The override is never consulted: no allowed_override, a plain role denial.
				wantDecision(t, what+" manages g1 labels", capLabelGroupManage, orgOnly, f, targetG(g1), false, reasonRoleNotPermitted)
				wantDecision(t, what+" reads g1 labels", capLabelGroupRead, orgOnly, f, targetG(g1), false, reasonRoleNotPermitted)
				wantDecision(t, what+" reads g1 labels, selected g1", capLabelGroupRead, selG1, f, targetG(g1), false, reasonRoleNotPermitted)
				wantDecision(t, what+" selected g3 reads g1 labels", capLabelGroupRead, selG3, f, targetG(g1), false, reasonTargetOutsideContext)
			}
		}
		// Dual role (Grace): both paths, independently; nothing in a Group without a role.
		grace := factsA(tenancy.OrgAdmin, groupBinding(g1, tenancy.GroupAdmin))
		grace.Policy.GroupManagementOverride = true
		wantDecision(t, "dual role Organization labels", capLabelOrganizationManage, orgOnly, grace, targetOrgA, true, reasonAllowed)
		wantDecision(t, "dual role own Group", capLabelGroupManage, orgOnly, grace, targetG(g1), true, reasonAllowed)
		wantDecision(t, "dual role other Group", capLabelGroupManage, orgOnly, grace, targetG(g2), false, reasonRoleNotPermitted)
		wantDecision(t, "dual role other Group read", capLabelGroupRead, orgOnly, grace, targetG(g2), false, reasonRoleNotPermitted)
		// Group roles: owner/admin manage and read, member reads; none reaches the Organization labels.
		f := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner), groupBinding(g2, tenancy.GroupAdmin), groupBinding(g3, tenancy.GroupMember))
		wantDecision(t, "group_owner manages", capLabelGroupManage, orgOnly, f, targetG(g1), true, reasonAllowed)
		wantDecision(t, "group_admin manages", capLabelGroupManage, orgOnly, f, targetG(g2), true, reasonAllowed)
		wantDecision(t, "group_member cannot manage", capLabelGroupManage, orgOnly, f, targetG(g3), false, reasonRoleNotPermitted)
		wantDecision(t, "group_member reads", capLabelGroupRead, orgOnly, f, targetG(g3), true, reasonAllowed)
		wantDecision(t, "no role in gNew", capLabelGroupRead, orgOnly, f, targetG(gNew), false, reasonRoleNotPermitted)
		wantDecision(t, "Group roles never manage Organization labels", capLabelOrganizationManage, orgOnly, f, targetOrgA, false, reasonRoleNotPermitted)
		noOrg := factsA("", groupBinding(g1, tenancy.GroupOwner))
		wantDecision(t, "group_owner without Organization binding", capLabelGroupRead, orgOnly, noOrg, targetG(g1), false, reasonOrganizationNotPermitted)
	})
}

// U-POL-MFA: the evaluator has no assurance, provider, claim, credential or host input anywhere in its argument types,
// and its decision is a function of the facts alone.
func TestPolicyHasNoAssuranceInput(t *testing.T) {
	forbidden := regexp.MustCompile(`(?i)mfa|assurance|factor|provider|claim|credential|token|session|email|host|header|cookie`)
	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type, string)
	walk = func(rt reflect.Type, path string) {
		for rt.Kind() == reflect.Pointer || rt.Kind() == reflect.Slice || rt.Kind() == reflect.Map {
			if rt.Kind() == reflect.Map {
				walk(rt.Key(), path+"[key]")
			}
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct || seen[rt] {
			return
		}
		seen[rt] = true
		for i := range rt.NumField() {
			fd := rt.Field(i)
			if forbidden.MatchString(fd.Name) {
				t.Errorf("evaluator input %s.%s", path, fd.Name)
			}
			walk(fd.Type, path+"."+fd.Name)
		}
	}
	dt := reflect.TypeOf(decide)
	for i := range dt.NumIn() {
		walk(dt.In(i), dt.In(i).String())
	}
	if dt.NumIn() != 4 || dt.In(0) != reflect.TypeOf(Capability("")) || dt.In(1) != reflect.TypeOf(contextSelection{}) ||
		dt.In(2) != reflect.TypeOf(Facts{}) || dt.In(3) != reflect.TypeOf(Target{}) {
		t.Fatalf("decide signature: %v", dt)
	}
	// Identical facts give identical decisions, whatever produced them.
	for _, c := range allCapabilities {
		f := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupAdmin))
		a1, r1 := decide(c, orgOnly, f, targetFor(c, g1, p1))
		a2, r2 := decide(c, orgOnly, factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupAdmin)), targetFor(c, g1, p1))
		if a1 != a2 || r1 != r2 {
			t.Errorf("%s: not a function of the facts", c)
		}
	}
}
