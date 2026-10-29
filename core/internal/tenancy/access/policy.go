package access

import "core/internal/tenancy"

// Capability is one of the twenty-two closed capabilities (the eleven of Spec R3, G016 organization.manage_domains, G017
// organization.manage_branding, G018 group.create and group.complete_onboarding, G019 organization.manage_owners,
// group.manage_owners, organization.leave and group.leave, and G020 label.organization.manage, label.group.manage and
// label.group.read). It is evaluated only inside this package: no allow result, capability check or evaluator is
// exported.
type Capability string

// The capabilities (Spec R3).
const (
	capOrganizationRead          Capability = "organization.read"
	capOrganizationManageMembers Capability = "organization.manage_members"
	capOrganizationManagePolicy  Capability = "organization.manage_policy"
	capGroupRead                 Capability = "group.read"
	capGroupUpdate               Capability = "group.update"
	capGroupManageMembers        Capability = "group.manage_members"
	capProfileRead               Capability = "business_profile.profile.read"
	capProfileUpdate             Capability = "business_profile.profile.update"
	capProfileSelect             Capability = "business_profile.select_for_onboarding"
	capProfilePrivileged         Capability = "business_profile.privileged_settings.manage"
	capAuditRead                 Capability = "audit.read"
	// G016: the custom-domain lifecycle of the active Organization (owner or admin; Organization scope, role-level).
	capOrganizationManageDomains Capability = "organization.manage_domains"
	// G017: the active Organization's branding (owner or admin, owner decision D-1; Organization scope, role-level).
	capOrganizationManageBranding Capability = "organization.manage_branding"
	// G018: creating a Group in the active Organization (every current Organization membership role, superseding D-1;
	// Organization-only context; Organization scope, role-level).
	capGroupCreate Capability = "group.create"
	// G018: completing a Group's onboarding with a Business Profile (all for Organization owner/admin, own for
	// group_owner/group_admin; Group scope).
	capGroupCompleteOnboarding Capability = "group.complete_onboarding"
	// G019 (Plan r1 §1.3): ownership administration and self-departure. organization.manage_owners is org_owner only;
	// group.manage_owners is the exact Group's group_owner only (no implicit Organization-owner override, Spec R1);
	// organization.leave is every Organization membership role; group.leave is every role of the exact Group.
	capOrganizationManageOwners Capability = "organization.manage_owners"
	capGroupManageOwners        Capability = "group.manage_owners"
	capOrganizationLeave        Capability = "organization.leave"
	capGroupLeave               Capability = "group.leave"
	// G020 (Plan r2 §1): labels. label.organization.manage is the active Organization's owners and admins (the
	// Organization-owned definitions and their Group assignments, private to them; Organization scope, role-level).
	// label.group.manage (owner/admin) and label.group.read (owner/admin/member) need the caller's own role in the exact
	// Group: Group-owned labels are Group content, so Organization owners/admins and the G060 override confer nothing.
	capLabelOrganizationManage Capability = "label.organization.manage"
	capLabelGroupManage        Capability = "label.group.manage"
	capLabelGroupRead          Capability = "label.group.read"
)

// Reason is a closed internal reason code. It appears only in access.decision records, never in an error.
type Reason string

// The reasons (G007 Plan r1 §3.6).
const (
	reasonAllowed Reason = "allowed"
	// G060 (Plan r1 C5): an allow through the Organization owner/admin path of the Group-management override, decided
	// only after the explicit Group path denied. A write allowed this way appends access.override_used atomically; a
	// read records it in its access.decision record only.
	reasonAllowedOverride    Reason = "allowed_override"
	reasonContextEstablished Reason = "context_established"
	reasonSessionInvalid     Reason = "session_invalid"
	reasonSelectorInvalid    Reason = "selector_invalid"

	reasonContextForged   Reason = "context_forged"
	reasonScopeMissing    Reason = "context_scope_missing"
	reasonScopeEnded      Reason = "context_scope_ended"
	reasonScopeMismatch   Reason = "context_scope_mismatch"
	reasonSubjectMismatch Reason = "context_subject_mismatch"
	reasonScopeNested     Reason = "context_scope_nested"

	reasonOrganizationNotPermitted Reason = "organization_not_permitted"
	reasonGroupNotSelectable       Reason = "group_not_selectable"
	reasonTargetNotFound           Reason = "target_not_found"
	reasonTargetKindMismatch       Reason = "target_kind_mismatch"
	reasonTargetOutsideContext     Reason = "target_outside_context"

	reasonCapabilityUnknown  Reason = "capability_unknown"
	reasonRoleNotPermitted   Reason = "role_not_permitted"
	reasonNoLinkedPath       Reason = "no_linked_path"
	reasonEditPolicyDisabled Reason = "edit_policy_disabled"
	reasonFactInconsistent   Reason = "fact_inconsistent"

	reasonInputInvalid         Reason = "input_invalid"
	reasonWriteConflict        Reason = "write_conflict"
	reasonStoreUnavailable     Reason = "store_unavailable"
	reasonAborted              Reason = "aborted"
	reasonCanceled             Reason = "canceled"
	reasonCommitOutcomeUnknown Reason = "commit_outcome_unknown"
)

// TargetKind is the kind of a capability's target.
type TargetKind string

// The target kinds.
const (
	kindOrganization TargetKind = "organization"
	kindGroup        TargetKind = "group"
	kindProfile      TargetKind = "business_profile"
)

// Target is the exact object a capability is evaluated against: the active Organization, a Group or a BusinessProfile.
type Target struct {
	Kind         TargetKind
	Organization tenancy.OrganizationID
	Group        tenancy.GroupID
	Profile      tenancy.BusinessProfileID
	// G020: the label definition a label operation acts on (zero otherwise). decide never reads it; it only names the
	// audit resource of a recorded denial (targetScope: resource label).
	Label tenancy.LabelDefinitionID
}

// contextSelection is the active context of a decision: the Organization and, in selected-Group context, the Group.
type contextSelection struct {
	Organization tenancy.OrganizationID
	Group        tenancy.GroupID
}

// Facts are the raw facts a decision is made on, loaded in one transaction by tenancy.ReadFacts.
type Facts = tenancy.Facts

// capabilityKind returns the target kind of a capability, false for an unknown capability.
func capabilityKind(c Capability) (TargetKind, bool) {
	switch c {
	case capOrganizationRead, capOrganizationManageMembers, capOrganizationManagePolicy, capAuditRead, capOrganizationManageDomains,
		capOrganizationManageBranding, capGroupCreate, capOrganizationManageOwners, capOrganizationLeave, capLabelOrganizationManage:
		return kindOrganization, true
	case capGroupRead, capGroupUpdate, capGroupManageMembers, capGroupCompleteOnboarding, capGroupManageOwners, capGroupLeave,
		capLabelGroupManage, capLabelGroupRead:
		return kindGroup, true
	case capProfileRead, capProfileUpdate, capProfileSelect, capProfilePrivileged:
		return kindProfile, true
	}
	return "", false
}

// readCapability is the read capability of a capability's kind: forbidden versus not_found is decided by it on the
// same facts. G020: the Group label capabilities read through label.group.read, not group.read, so a subject without a
// role in the Group (an Organization owner/admin reading Group metadata, with or without the override) gets not_found
// for that Group's labels; label.organization.manage reads through organization.read like every Organization
// capability (the Organization is readable, the denial is forbidden).
func readCapability(c Capability) Capability {
	switch c {
	case capLabelGroupManage, capLabelGroupRead:
		return capLabelGroupRead
	}
	switch k, _ := capabilityKind(c); k {
	case kindGroup:
		return capGroupRead
	case kindProfile:
		return capProfileRead
	}
	return capOrganizationRead
}

func zero[T ~[16]byte](id T) bool { return id == T{} }

// organizationRole is the Organization gate: the facts carry the active Organization and exactly one current
// Organization membership binding of the subject in it, with a known role and the membership source.
func organizationRole(ctx contextSelection, f Facts) (tenancy.OrganizationRole, Reason) {
	if zero(ctx.Organization) || f.Organization.ID != ctx.Organization {
		return "", reasonOrganizationNotPermitted
	}
	switch len(f.OrganizationBindings) {
	case 0:
		return "", reasonOrganizationNotPermitted
	case 1:
	default:
		return "", reasonFactInconsistent
	}
	b := f.OrganizationBindings[0]
	r := tenancy.OrganizationRole(b.Role)
	if b.OrganizationID != ctx.Organization || !zero(b.GroupID) || b.Source != tenancy.SourceMembership || !r.Valid() {
		return "", reasonFactInconsistent
	}
	return r, ""
}

// groupRoles returns the subject's Group roles in the active Organization. A binding of another Organization, without a
// Group, with an unknown role or with any source but membership confers nothing; two bindings of one Group are
// inconsistent.
func groupRoles(ctx contextSelection, f Facts) (map[tenancy.GroupID]tenancy.GroupRole, bool) {
	out := map[tenancy.GroupID]tenancy.GroupRole{}
	for _, b := range f.GroupBindings {
		r := tenancy.GroupRole(b.Role)
		if b.OrganizationID != ctx.Organization || zero(b.GroupID) || b.Source != tenancy.SourceMembership || !r.Valid() {
			continue
		}
		if _, dup := out[b.GroupID]; dup {
			return nil, false
		}
		out[b.GroupID] = r
	}
	return out, true
}

// association returns the one profile associated with a Group, false when it has none or (inconsistently) several.
func association(f Facts, g tenancy.GroupID) (tenancy.BusinessProfileID, bool) {
	var p tenancy.BusinessProfileID
	n := 0
	for _, a := range f.Associations {
		if a.GroupID == g && !zero(a.BusinessProfileID) {
			if n > 0 && a.BusinessProfileID != p {
				return tenancy.BusinessProfileID{}, false
			}
			p, n = a.BusinessProfileID, n+1
		}
	}
	return p, n > 0
}

func hasGroup(f Facts, org tenancy.OrganizationID, g tenancy.GroupID) bool {
	for _, x := range f.Groups {
		if x.ID == g && x.OrganizationID == org && !zero(g) {
			return true
		}
	}
	return false
}

func hasProfile(f Facts, org tenancy.OrganizationID, p tenancy.BusinessProfileID) bool {
	for _, x := range f.Profiles {
		if x.ID == p && x.OrganizationID == org && !zero(p) {
			return true
		}
	}
	return false
}

// decide is the closed, deny-by-default evaluator of Spec R3 (G007 Plan r1 §3.3). It is pure: no clock, provider,
// assurance, name, hostname or claim input. Gates apply in order: the capability is known; the Organization gate; the
// target gate (kind and exact ownership); the context gate (a Group target must be the selected Group); the matrix.
func decide(cap Capability, ctx contextSelection, f Facts, t Target) (bool, Reason) {
	kind, ok := capabilityKind(cap)
	if !ok {
		return false, reasonCapabilityUnknown
	}
	orgRole, r := organizationRole(ctx, f)
	if r != "" {
		return false, r
	}
	switch t.Kind {
	case kindOrganization, kindGroup, kindProfile:
	default:
		return false, reasonFactInconsistent
	}
	if t.Kind != kind {
		return false, reasonTargetKindMismatch
	}
	switch kind {
	case kindOrganization:
		if t.Organization != ctx.Organization {
			return false, reasonTargetNotFound
		}
	case kindGroup:
		if !hasGroup(f, ctx.Organization, t.Group) {
			return false, reasonTargetNotFound
		}
		if !zero(ctx.Group) && t.Group != ctx.Group {
			return false, reasonTargetOutsideContext
		}
	case kindProfile:
		if !hasProfile(f, ctx.Organization, t.Profile) {
			return false, reasonTargetNotFound
		}
	}
	roles, ok := groupRoles(ctx, f)
	if !ok {
		return false, reasonFactInconsistent
	}
	orgAdmin := orgRole == tenancy.OrgOwner || orgRole == tenancy.OrgAdmin
	// G060 (ADR-008; Spec R1/R2): Organization owners and admins read Group and Business Profile metadata (orgAdmin);
	// ordinary Group administration needs an explicit Group owner/admin binding, or the Organization's default-off
	// override (orgMgr). Neither confers Group content (decideObject), Group ownership or membership.
	orgMgr := orgAdmin && f.Policy.GroupManagementOverride

	switch cap {
	case capOrganizationRead:
		return allowIf(orgRole.Valid())
	case capGroupCreate:
		// Organization-only context: a selected-Group context never creates a Group (Plan r5 §1.3); the readable
		// Organization makes the refusal forbidden, not not_found.
		if !zero(ctx.Group) {
			return false, reasonRoleNotPermitted
		}
		return allowIf(orgRole.Valid())
	case capOrganizationManageMembers, capOrganizationManageDomains, capOrganizationManageBranding, capLabelOrganizationManage:
		return allowIf(orgAdmin)
	case capOrganizationManagePolicy, capAuditRead, capProfilePrivileged, capOrganizationManageOwners:
		return allowIf(orgRole == tenancy.OrgOwner)
	case capOrganizationLeave:
		return allowIf(orgRole.Valid())
	case capGroupManageOwners:
		// Own only: an Organization owner or admin without the Group's owner binding is denied (Spec R1).
		return allowIf(roles[t.Group] == tenancy.GroupOwner)
	case capGroupLeave:
		// Own only: an Organization-role subject without a binding of the Group has nothing to leave.
		return allowIf(roles[t.Group].Valid())
	case capGroupRead:
		// Metadata: the Organization path holds regardless of the override (Spec R1).
		return allowIf(orgAdmin || roles[t.Group].Valid())
	case capLabelGroupManage:
		// G020: Group content (the decideObject rule): own owner/admin only; orgAdmin and orgMgr are never consulted.
		gr := roles[t.Group]
		return allowIf(gr == tenancy.GroupOwner || gr == tenancy.GroupAdmin)
	case capLabelGroupRead:
		// G020: own role of any kind; target_members visibility is the operation's row rule, not a matrix input.
		return allowIf(roles[t.Group].Valid())
	case capGroupUpdate, capGroupManageMembers, capGroupCompleteOnboarding:
		// The explicit Group path first; the override only when it denies (Plan r1 C5).
		if gr := roles[t.Group]; gr == tenancy.GroupOwner || gr == tenancy.GroupAdmin {
			return true, reasonAllowed
		}
		if orgMgr {
			return true, reasonAllowedOverride
		}
		return false, reasonRoleNotPermitted
	case capProfileRead:
		// Ordinary profile fields are metadata (Spec R1): every profile of the Organization for its owners and admins.
		if orgAdmin {
			return true, reasonAllowed
		}
		return linkedPath(cap, ctx, f, roles, t.Profile)
	case capProfileUpdate, capProfileSelect:
		// The linked Group path first (with the group_admins_edit_business_profiles flag for update); the override only
		// when it denies (Plan r0 D-5, r1 C5).
		ok, r := linkedPath(cap, ctx, f, roles, t.Profile)
		if ok {
			return true, r
		}
		if orgMgr {
			return true, reasonAllowedOverride
		}
		return false, r
	}
	return false, reasonCapabilityUnknown
}

// allowed reports an allow reason of decide (reasonAllowed or, G060, reasonAllowedOverride).
func allowed(r Reason) bool { return r == reasonAllowed || r == reasonAllowedOverride }

func allowIf(ok bool) (bool, Reason) {
	if ok {
		return true, reasonAllowed
	}
	return false, reasonRoleNotPermitted
}

// linkedPath is the Group-derived BusinessProfile path: a candidate Group with a qualifying Group role whose current
// association is the profile. Candidate Groups are the selected Group in selected-Group context, otherwise every Group
// of the subject in the active Organization. Read admits any Group role; update and selection admit owner and admin;
// update also needs the Organization's group_admins_edit_business_profiles flag.
func linkedPath(cap Capability, ctx contextSelection, f Facts, roles map[tenancy.GroupID]tenancy.GroupRole, p tenancy.BusinessProfileID) (bool, Reason) {
	qualifies := func(r tenancy.GroupRole) bool {
		if cap == capProfileRead {
			return r.Valid()
		}
		return r == tenancy.GroupOwner || r == tenancy.GroupAdmin
	}
	var candidates []tenancy.GroupID
	if !zero(ctx.Group) {
		if _, ok := roles[ctx.Group]; ok {
			candidates = []tenancy.GroupID{ctx.Group}
		}
	} else {
		for g := range roles {
			candidates = append(candidates, g)
		}
	}
	qualified := false
	for _, g := range candidates {
		if !qualifies(roles[g]) {
			continue
		}
		qualified = true
		if linked, ok := association(f, g); ok && linked == p {
			if cap == capProfileUpdate && !f.Policy.GroupAdminsEditBusinessProfiles {
				return false, reasonEditPolicyDisabled
			}
			return true, reasonAllowed
		}
	}
	if qualified {
		return false, reasonNoLinkedPath
	}
	return false, reasonRoleNotPermitted
}

// eligible is the G015 discovery rule of one tool requirement in the current context, on facts read with
// factQueryFor (docs/foundation/AUTHORIZATION.md "Capabilities and the matrix"; Plan r4 §1). The context must hold
// (contextCheck). The binding, membership and context requirements are then satisfied. A capability is eligible when
// decide allows an Organization capability on the active Organization; or, for a Group or BusinessProfile capability,
// when the Organization owner/admin path holds; or when decide allows a Group-derived path over the candidate Groups (the
// selected Group in selected context, otherwise every Group of the subject) and, for profiles, their current
// associations. The Organization owner/admin path is role-level, from the Organization binding alone (the all and Y
// cells: any Group, restricted to the selected Group that contextCheck validated; any profile, unlinked ones included,
// independent of the selection). It asserts no existing Group or profile target, so it also holds in an otherwise empty
// Organization; privileged settings stay owner-only. Eligibility is a hint: it proves no target exists, decides no
// target and authorizes nothing, and an unknown or foreign target keeps its non-disclosing refusal at call time.
func eligible(req ToolRequirement, sel contextSelection, f Facts) bool {
	if contextCheck(sel, f) != "" {
		return false
	}
	switch req.kind {
	case reqBinding, reqMembership, reqContext:
		return true
	case reqCapability:
	default:
		return false
	}
	kind, ok := capabilityKind(req.cap)
	if !ok {
		return false
	}
	orgRole, _ := organizationRole(sel, f)
	orgAdmin := orgRole == tenancy.OrgOwner || orgRole == tenancy.OrgAdmin
	if kind == kindOrganization {
		ok, _ := decide(req.cap, sel, f, Target{Kind: kindOrganization, Organization: sel.Organization})
		return ok
	}
	// The Organization path (the all and Y cells of the matrix); privileged settings are owner-only. G060: the metadata
	// reads hold for every Organization owner/admin; ordinary Group administration only while the Organization's
	// override is enabled (decide's orgMgr).
	switch req.cap {
	case capProfilePrivileged:
		return orgRole == tenancy.OrgOwner
	case capGroupRead, capProfileRead:
		if orgAdmin {
			return true
		}
	case capGroupUpdate, capGroupManageMembers, capGroupCompleteOnboarding, capProfileUpdate, capProfileSelect:
		if orgAdmin && f.Policy.GroupManagementOverride {
			return true
		}
	case capLabelGroupManage, capLabelGroupRead:
		// G020: no Organization path (Group content; the override is irrelevant): only the Group-derived paths below.
	}
	// The Group-derived paths, decided by decide on each candidate Group or its linked profile.
	roles, ok := groupRoles(sel, f)
	if !ok {
		return false
	}
	for g := range roles {
		if !zero(sel.Group) && g != sel.Group {
			continue
		}
		var t Target
		if kind == kindGroup {
			t = Target{Kind: kindGroup, Group: g}
		} else {
			p, linked := association(f, g)
			if !linked {
				continue
			}
			t = Target{Kind: kindProfile, Profile: p}
		}
		if ok, _ := decide(req.cap, sel, f, t); ok {
			return true
		}
	}
	return false
}

// factQueryFor is the fact query of an eligibility snapshot (Plan r4 §1 "Facts per requirement"): the snapshot's own
// Organization-only (no Group) or selected-Group facts, plus, when any requirement is a Group or BusinessProfile
// capability, every Group of the subject in Organization-only context (Groups nil) and every profile of the Organization
// with the candidate Groups' associations. The options depend on the catalog's shape, never on the subject's role. The
// caller sets the user.
func factQueryFor(reqs map[string]ToolRequirement, sel contextSelection) tenancy.FactQuery {
	q := tenancy.FactQuery{Organization: sel.Organization, Groups: []tenancy.GroupID{}}
	if !zero(sel.Group) {
		q.Groups = []tenancy.GroupID{sel.Group}
	}
	for _, r := range reqs {
		if r.kind != reqCapability {
			continue
		}
		if k, _ := capabilityKind(r.cap); k == kindGroup || k == kindProfile {
			if zero(sel.Group) {
				q.Groups = nil
			}
			q.AllProfiles = true
			break
		}
	}
	return q
}

// contextCheck revalidates the trusted context on the operation's own facts: the Organization gate, and in
// selected-Group context that the Group still belongs to the Organization and is selectable by the subject (an
// Organization owner or admin, or a current binding in that exact Group). A Group binding alone never establishes
// tenant access.
func contextCheck(ctx contextSelection, f Facts) Reason {
	orgRole, r := organizationRole(ctx, f)
	if r != "" {
		return r
	}
	if zero(ctx.Group) {
		return ""
	}
	if !hasGroup(f, ctx.Organization, ctx.Group) {
		return reasonGroupNotSelectable
	}
	if orgRole == tenancy.OrgOwner || orgRole == tenancy.OrgAdmin {
		return ""
	}
	roles, ok := groupRoles(ctx, f)
	if !ok {
		return reasonFactInconsistent
	}
	if _, bound := roles[ctx.Group]; !bound {
		return reasonGroupNotSelectable
	}
	return ""
}
