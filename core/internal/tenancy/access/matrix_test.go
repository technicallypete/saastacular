// U-POL-MATRIX (G007 Plan r1 §6.1 and §10; Spec AC3): the independent expected-result oracle for the pure evaluator
// decide. The oracle delegate wrote it from Spec R1 to R3, Plan r1 §3.3 with its interpretation pins Q8(a)-(d), the
// pinned decide signature, the role constants of core/internal/tenancy and the coordinator's interface-only type
// shapes, without reading policy.go or any other implementation file of this package. Every expectation is a
// hand-written literal; none is computed from decide, from another evaluator or from the R3 symbols. Only the allow
// result is compared; the internal Reason is not.
//
// A disagreement between this file and decide is resolved by the implementer against the Spec, or escalated; neither
// side is edited to match the other (Plan r1 §10).
//
// Fixture (Plan r1 §6.3 naming): the active Organization A has Groups g1, g2 and g3 and the new unattached Group g4;
// profile P1 is shared by g1 and g2, P2 is linked to g3 and P3 is unlinked. Organization B owns gB and PB, which are
// never facts of A except in the inconsistent-fact cases. Per pin Q8(a), every Group-role subject also holds an
// org_member binding in A, so the Organization gate is held constant and a Group-role column shows the combination of
// its Group path with the org_member paths.
//
// Each primary, scope and manageable case runs twice: with the complete facts of A (every binding, Group, association
// and profile, whatever the selected context), which proves that decide applies the selected-Group restriction itself
// (Plan r1 §3.3, candidate Groups), and with the subset that the Plan r1 §3.7 fact query loads for that decision.
package access

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"core/internal/tenancy"
)

// The pinned evaluator entry point (Plan r1 §3.3) by its stated names. A compile error here is a signature gap for the
// implementer to resolve against the Plan, not a reason to edit this oracle.
var _ func(Capability, contextSelection, Facts, Target) (bool, Reason) = decide

// Fixed canonical UUIDs of the fixture, distinct across every ID type.
var (
	oracleOrgA = oracleMustOrg("a0000000-0000-4000-8000-00000000000a")
	oracleOrgB = oracleMustOrg("b0000000-0000-4000-8000-00000000000b")
	oracleGrp1 = oracleMustGroup("a1000000-0000-4000-8000-000000000001")
	oracleGrp2 = oracleMustGroup("a2000000-0000-4000-8000-000000000002")
	oracleGrp3 = oracleMustGroup("a3000000-0000-4000-8000-000000000003")
	oracleGrp4 = oracleMustGroup("a4000000-0000-4000-8000-000000000004")
	oracleGrpB = oracleMustGroup("b1000000-0000-4000-8000-00000000000b")
	oracleGrpX = oracleMustGroup("c1000000-0000-4000-8000-00000000000c")
	oraclePrf1 = oracleMustProfile("a1000000-0000-4000-9000-000000000001")
	oraclePrf2 = oracleMustProfile("a2000000-0000-4000-9000-000000000002")
	oraclePrf3 = oracleMustProfile("a3000000-0000-4000-9000-000000000003")
	oraclePrfB = oracleMustProfile("b1000000-0000-4000-9000-00000000000b")
	oraclePrfX = oracleMustProfile("c1000000-0000-4000-9000-00000000000c")
)

// oracleOrgOnly is the zero Group selector: Organization-only context.
var oracleOrgOnly tenancy.GroupID

// oracleAllow and oracleDeny are the literal expectations; oracleFlagOn and oracleFlagOff are the two values of the
// R4 group_admins_edit_business_profiles flag.
const oracleAllow = true
const oracleDeny = false
const oracleFlagOn = true
const oracleFlagOff = false

// oracleCapKind pairs a Spec R3 capability with its Spec R2 target kind.
type oracleCapKind struct {
	name string
	kind TargetKind
}

// oracleCapabilities are the 11 capabilities of Spec R3 in Spec order, then G016's twelfth (organization.manage_domains,
// Plan r4 §5.0.1: Organization owner or admin, the manage_members rule) and G017's thirteenth
// (organization.manage_branding, owner decision D-1: Organization owner or admin, the same rule), then G018's
// group.create (every Organization membership role, superseding D-1; Organization-only context, so every selected-Group
// cell denies) and group.complete_onboarding (all for Organization owner/admin, own for group_owner/group_admin; the
// group.update rule). Spec R2: the Organization capabilities and
// audit.read target the active Organization, the three Group capabilities an exact Group and the four BusinessProfile
// capabilities an exact BusinessProfile.
var oracleCapabilities = []oracleCapKind{
	{"organization.read", "organization"},
	{"organization.manage_members", "organization"},
	{"organization.manage_policy", "organization"},
	{"group.read", "group"},
	{"group.update", "group"},
	{"group.manage_members", "group"},
	{"business_profile.profile.read", "business_profile"},
	{"business_profile.profile.update", "business_profile"},
	{"business_profile.select_for_onboarding", "business_profile"},
	{"business_profile.privileged_settings.manage", "business_profile"},
	{"audit.read", "organization"},
	{"organization.manage_domains", "organization"},
	{"organization.manage_branding", "organization"},
	{"group.create", "organization"},
	{"group.complete_onboarding", "group"},
	// G019 (Plan r1 §1.3): organization.manage_owners (org_owner), group.manage_owners (the exact Group's group_owner
	// only; no Organization-owner override), organization.leave (every Organization role) and group.leave (own Group
	// role only; an Organization role alone has nothing to leave).
	{"organization.manage_owners", "organization"},
	{"group.manage_owners", "group"},
	{"organization.leave", "organization"},
	{"group.leave", "group"},
	// G020 (Plan r2 §1): label.organization.manage (Organization owner/admin; the Organization-owned labels are private
	// to them), label.group.manage (the exact Group's owner/admin) and label.group.read (any role of the exact Group).
	// Group-owned labels are Group content: an Organization role alone, with or without the G060 override, confers
	// nothing on either Group label capability.
	{"label.organization.manage", "organization"},
	{"label.group.manage", "group"},
	{"label.group.read", "group"},
}

// oracleColumns are the six roles of Spec R3 in column order, as the core/internal/tenancy role constants.
var oracleColumns = []string{
	string(tenancy.OrgOwner), string(tenancy.OrgAdmin), string(tenancy.OrgMember),
	string(tenancy.GroupOwner), string(tenancy.GroupAdmin), string(tenancy.GroupMember),
}

// oracleRow is one literal row of a primary table: six cells in oracleColumns order, then the capability. A cell is A
// for allow, D for deny or X for not asserted. X marks the org_member column in selected-Group context: no selected
// Group context can be established for that subject, because Resolve refuses it (Plan r1 §3.2 step 3), and
// IT-MATRIX-1 asserts that refusal at Resolve.
type oracleRow struct {
	cells      string
	capability string
}

// oracleOrgOnlyFlagOn is the primary scenario in Organization-only context with the flag true. In every primary cell
// the subject holds exactly the column role (a Group role is held in g1, together with org_member per pin Q8(a)) and
// the target is Organization A, Group g1 or profile P1 according to the capability's kind.
var oracleOrgOnlyFlagOn = []oracleRow{
	// org_owner, org_admin, org_member, group_owner, group_admin, group_member
	{"A A A A A A", "organization.read"},
	{"A A D D D D", "organization.manage_members"},
	{"A D D D D D", "organization.manage_policy"},
	{"A A D A A A", "group.read"},
	{"D D D A A D", "group.update"},
	{"D D D A A D", "group.manage_members"},
	{"A A D A A A", "business_profile.profile.read"},
	{"D D D A A D", "business_profile.profile.update"},
	{"D D D A A D", "business_profile.select_for_onboarding"},
	{"A D D D D D", "business_profile.privileged_settings.manage"},
	{"A D D D D D", "audit.read"},
	{"A A D D D D", "organization.manage_domains"},
	{"A A D D D D", "organization.manage_branding"},
	{"A A A A A A", "group.create"},
	{"D D D A A D", "group.complete_onboarding"},
	{"A D D D D D", "organization.manage_owners"},
	{"D D D A D D", "group.manage_owners"},
	{"A A A A A A", "organization.leave"},
	{"D D D A A A", "group.leave"},
	{"A A D D D D", "label.organization.manage"},
	{"D D D A A D", "label.group.manage"},
	{"D D D A A A", "label.group.read"},
}

// oracleOrgOnlyFlagOff is the primary scenario in Organization-only context with the flag false.
var oracleOrgOnlyFlagOff = []oracleRow{
	// org_owner, org_admin, org_member, group_owner, group_admin, group_member
	{"A A A A A A", "organization.read"},
	{"A A D D D D", "organization.manage_members"},
	{"A D D D D D", "organization.manage_policy"},
	{"A A D A A A", "group.read"},
	{"D D D A A D", "group.update"},
	{"D D D A A D", "group.manage_members"},
	{"A A D A A A", "business_profile.profile.read"},
	{"D D D D D D", "business_profile.profile.update"},
	{"D D D A A D", "business_profile.select_for_onboarding"},
	{"A D D D D D", "business_profile.privileged_settings.manage"},
	{"A D D D D D", "audit.read"},
	{"A A D D D D", "organization.manage_domains"},
	{"A A D D D D", "organization.manage_branding"},
	{"A A A A A A", "group.create"},
	{"D D D A A D", "group.complete_onboarding"},
	{"A D D D D D", "organization.manage_owners"},
	{"D D D A D D", "group.manage_owners"},
	{"A A A A A A", "organization.leave"},
	{"D D D A A A", "group.leave"},
	{"A A D D D D", "label.organization.manage"},
	{"D D D A A D", "label.group.manage"},
	{"D D D A A A", "label.group.read"},
}

// oracleSelectedFlagOn is the primary scenario in the selected-Group context g1 with the flag true.
var oracleSelectedFlagOn = []oracleRow{
	// org_owner, org_admin, org_member, group_owner, group_admin, group_member
	{"A A X A A A", "organization.read"},
	{"A A X D D D", "organization.manage_members"},
	{"A D X D D D", "organization.manage_policy"},
	{"A A X A A A", "group.read"},
	{"D D X A A D", "group.update"},
	{"D D X A A D", "group.manage_members"},
	{"A A X A A A", "business_profile.profile.read"},
	{"D D X A A D", "business_profile.profile.update"},
	{"D D X A A D", "business_profile.select_for_onboarding"},
	{"A D X D D D", "business_profile.privileged_settings.manage"},
	{"A D X D D D", "audit.read"},
	{"A A X D D D", "organization.manage_domains"},
	{"A A X D D D", "organization.manage_branding"},
	{"D D X D D D", "group.create"},
	{"D D X A A D", "group.complete_onboarding"},
	{"A D X D D D", "organization.manage_owners"},
	{"D D X A D D", "group.manage_owners"},
	{"A A X A A A", "organization.leave"},
	{"D D X A A A", "group.leave"},
	{"A A X D D D", "label.organization.manage"},
	{"D D X A A D", "label.group.manage"},
	{"D D X A A A", "label.group.read"},
}

// oracleSelectedFlagOff is the primary scenario in the selected-Group context g1 with the flag false.
var oracleSelectedFlagOff = []oracleRow{
	// org_owner, org_admin, org_member, group_owner, group_admin, group_member
	{"A A X A A A", "organization.read"},
	{"A A X D D D", "organization.manage_members"},
	{"A D X D D D", "organization.manage_policy"},
	{"A A X A A A", "group.read"},
	{"D D X A A D", "group.update"},
	{"D D X A A D", "group.manage_members"},
	{"A A X A A A", "business_profile.profile.read"},
	{"D D X D D D", "business_profile.profile.update"},
	{"D D X A A D", "business_profile.select_for_onboarding"},
	{"A D X D D D", "business_profile.privileged_settings.manage"},
	{"A D X D D D", "audit.read"},
	{"A A X D D D", "organization.manage_domains"},
	{"A A X D D D", "organization.manage_branding"},
	{"D D X D D D", "group.create"},
	{"D D X A A D", "group.complete_onboarding"},
	{"A D X D D D", "organization.manage_owners"},
	{"D D X A D D", "group.manage_owners"},
	{"A A X A A A", "organization.leave"},
	{"D D X A A A", "group.leave"},
	{"A A X D D D", "label.organization.manage"},
	{"D D X A A D", "label.group.manage"},
	{"D D X A A A", "label.group.read"},
}

// G060 oracleOrgOnlyFlagOnOverride is the Organization-only, flag-true scenario with the Organization's Group-management
// override enabled: Organization owners and admins administer every Group of A (all) without a Group binding; nothing
// else changes (content, ownership and leave never follow from it).
var oracleOrgOnlyFlagOnOverride = []oracleRow{
	// org_owner, org_admin, org_member, group_owner, group_admin, group_member
	{"A A A A A A", "organization.read"},
	{"A A D D D D", "organization.manage_members"},
	{"A D D D D D", "organization.manage_policy"},
	{"A A D A A A", "group.read"},
	{"A A D A A D", "group.update"},
	{"A A D A A D", "group.manage_members"},
	{"A A D A A A", "business_profile.profile.read"},
	{"A A D A A D", "business_profile.profile.update"},
	{"A A D A A D", "business_profile.select_for_onboarding"},
	{"A D D D D D", "business_profile.privileged_settings.manage"},
	{"A D D D D D", "audit.read"},
	{"A A D D D D", "organization.manage_domains"},
	{"A A D D D D", "organization.manage_branding"},
	{"A A A A A A", "group.create"},
	{"A A D A A D", "group.complete_onboarding"},
	{"A D D D D D", "organization.manage_owners"},
	{"D D D A D D", "group.manage_owners"},
	{"A A A A A A", "organization.leave"},
	{"D D D A A A", "group.leave"},
	{"A A D D D D", "label.organization.manage"},
	{"D D D A A D", "label.group.manage"},
	{"D D D A A A", "label.group.read"},
}

// G060 oracleSelectedFlagOffOverride is the selected-g1, flag-false scenario with the override enabled.
var oracleSelectedFlagOffOverride = []oracleRow{
	// org_owner, org_admin, org_member, group_owner, group_admin, group_member
	{"A A X A A A", "organization.read"},
	{"A A X D D D", "organization.manage_members"},
	{"A D X D D D", "organization.manage_policy"},
	{"A A X A A A", "group.read"},
	{"A A X A A D", "group.update"},
	{"A A X A A D", "group.manage_members"},
	{"A A X A A A", "business_profile.profile.read"},
	{"A A X D D D", "business_profile.profile.update"},
	{"A A X A A D", "business_profile.select_for_onboarding"},
	{"A D X D D D", "business_profile.privileged_settings.manage"},
	{"A D X D D D", "audit.read"},
	{"A A X D D D", "organization.manage_domains"},
	{"A A X D D D", "organization.manage_branding"},
	{"D D X D D D", "group.create"},
	{"A A X A A D", "group.complete_onboarding"},
	{"A D X D D D", "organization.manage_owners"},
	{"D D X A D D", "group.manage_owners"},
	{"A A X A A A", "organization.leave"},
	{"D D X A A A", "group.leave"},
	{"A A X D D D", "label.organization.manage"},
	{"D D X A A D", "label.group.manage"},
	{"D D X A A A", "label.group.read"},
}

// oracleTable is one primary table with its context mode (selected means the selected Group g1), flag value and (G060)
// Group-management override value.
type oracleTable struct {
	name     string
	selected bool
	flagOn   bool
	rows     []oracleRow
	override bool
}

// oraclePrimaryTables returns the primary tables in a fixed order: the four override-off (default) tables, then the
// two G060 override-on tables.
func oraclePrimaryTables() []oracleTable {
	return []oracleTable{
		{"org-only/flag-on", false, oracleFlagOn, oracleOrgOnlyFlagOn, false},
		{"org-only/flag-off", false, oracleFlagOff, oracleOrgOnlyFlagOff, false},
		{"selected-g1/flag-on", true, oracleFlagOn, oracleSelectedFlagOn, false},
		{"selected-g1/flag-off", true, oracleFlagOff, oracleSelectedFlagOff, false},
		{"org-only/flag-on/override-on", false, oracleFlagOn, oracleOrgOnlyFlagOnOverride, true},
		{"selected-g1/flag-off/override-on", true, oracleFlagOff, oracleSelectedFlagOffOverride, true},
	}
}

// oracleScopeCase is one literal own/all/linked/flag/selected-Group variant (Spec R2 and R3; Plan r1 §3.3 and pins
// Q8(b) and Q8(d)). sel is the selected Group, oracleOrgOnly for Organization-only context.
type oracleScopeCase struct {
	id         string
	subject    string
	sel        tenancy.GroupID
	target     Target
	flagOn     bool
	capability string
	want       bool
}

// oracleScopeCases are the scope variants. Each ID names one subject, context, target and flag value.
var oracleScopeCases = []oracleScopeCase{
	// own versus all in Organization-only context (Spec R3 own/all). Spec R2: Group-derived authority never authorizes
	// an unrelated Group record, even g2, which shares P1 with g1. all covers every Group of A, including unattached g4.
	{"S01", "group_owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.read", oracleDeny},
	{"S01", "group_owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleDeny},
	{"S01", "group_owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S02", "group_admin", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.read", oracleDeny},
	{"S02", "group_admin", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleDeny},
	{"S02", "group_admin", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S03", "group_member", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.read", oracleDeny},
	{"S04", "org_owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.read", oracleAllow},
	{"S04", "org_owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleDeny},
	{"S04", "org_owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S05", "org_owner", oracleOrgOnly, oracleG(oracleGrp4), oracleFlagOn, "group.read", oracleAllow},
	{"S05", "org_owner", oracleOrgOnly, oracleG(oracleGrp4), oracleFlagOn, "group.update", oracleDeny},
	{"S06", "org_admin", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.read", oracleAllow},
	{"S06", "org_admin", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.update", oracleDeny},
	{"S06", "org_admin", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S07", "org_member", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.read", oracleDeny},
	{"S07", "org_member", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleDeny},
	{"S07", "org_member", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S08", "g1owner+g3admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "group.read", oracleAllow},
	{"S08", "g1owner+g3admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "group.update", oracleAllow},
	{"S08", "g1owner+g3admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "group.manage_members", oracleAllow},
	{"S09", "g1owner+g3admin", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.read", oracleAllow},
	{"S09", "g1owner+g3admin", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.update", oracleAllow},
	{"S09", "g1owner+g3admin", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleAllow},
	{"S10", "g1owner+g2owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleAllow},
	{"S11", "g1owner+g2owner", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.read", oracleDeny},
	{"S12", "g4owner", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "group.read", oracleDeny},

	// Selected-Group context gate for group.* (Plan r1 §3.3 gate 3; pin Q8(b): an org owner/admin's all path is
	// restricted to the selected Group; Spec R2: a Group-target operation in selected context requires that exact Group).
	{"S13", "org_owner", oracleGrp1, oracleG(oracleGrp2), oracleFlagOn, "group.read", oracleDeny},
	{"S13", "org_owner", oracleGrp1, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleDeny},
	{"S13", "org_owner", oracleGrp1, oracleG(oracleGrp2), oracleFlagOn, "group.manage_members", oracleDeny},
	// G018 group.complete_onboarding follows the group.update rule: own for group_owner/group_admin, all (unattached g4
	// included) for Organization owner/admin, restricted to the selected Group.
	{"S01", "group_owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.complete_onboarding", oracleDeny},
	{"S02", "group_admin", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.complete_onboarding", oracleDeny},
	{"S04", "org_owner", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.complete_onboarding", oracleDeny},
	{"S05", "org_owner", oracleOrgOnly, oracleG(oracleGrp4), oracleFlagOn, "group.complete_onboarding", oracleDeny},
	{"S07", "org_member", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.complete_onboarding", oracleDeny},
	{"S13", "org_owner", oracleGrp1, oracleG(oracleGrp2), oracleFlagOn, "group.complete_onboarding", oracleDeny},
	{"S14", "org_admin", oracleGrp1, oracleG(oracleGrp3), oracleFlagOn, "group.read", oracleDeny},
	{"S14", "org_admin", oracleGrp1, oracleG(oracleGrp3), oracleFlagOn, "group.update", oracleDeny},
	{"S14", "org_admin", oracleGrp1, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S15", "org_admin", oracleGrp2, oracleG(oracleGrp2), oracleFlagOn, "group.read", oracleAllow},
	{"S15", "org_admin", oracleGrp2, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleDeny},
	{"S15", "org_admin", oracleGrp2, oracleG(oracleGrp2), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S16", "group_owner", oracleGrp1, oracleG(oracleGrp2), oracleFlagOn, "group.read", oracleDeny},
	{"S16", "group_owner", oracleGrp1, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleDeny},
	{"S17", "g1owner+g3admin", oracleGrp3, oracleG(oracleGrp1), oracleFlagOn, "group.read", oracleDeny},
	{"S17", "g1owner+g3admin", oracleGrp3, oracleG(oracleGrp1), oracleFlagOn, "group.update", oracleDeny},
	{"S17", "g1owner+g3admin", oracleGrp3, oracleG(oracleGrp1), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S18", "g1owner+g3admin", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.read", oracleAllow},
	{"S18", "g1owner+g3admin", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.update", oracleAllow},
	{"S18", "g1owner+g3admin", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleAllow},
	{"S19", "org_admin+g3member", oracleGrp3, oracleG(oracleGrp1), oracleFlagOn, "group.read", oracleDeny},
	{"S19", "org_admin+g3member", oracleGrp3, oracleG(oracleGrp1), oracleFlagOn, "group.update", oracleDeny},
	{"S20", "org_admin+g3member", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.read", oracleAllow},
	{"S20", "org_admin+g3member", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.update", oracleDeny},
	{"S20", "org_admin+g3member", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S21", "g4owner", oracleGrp4, oracleG(oracleGrp4), oracleFlagOn, "group.read", oracleAllow},
	{"S21", "g4owner", oracleGrp4, oracleG(oracleGrp4), oracleFlagOn, "group.update", oracleAllow},
	{"S21", "g4owner", oracleGrp4, oracleG(oracleGrp4), oracleFlagOn, "group.manage_members", oracleAllow},
	{"S22", "g4owner", oracleGrp4, oracleG(oracleGrp1), oracleFlagOn, "group.read", oracleDeny},

	// linked in Organization-only context: candidate Groups are every Group of the subject in A (Plan r1 §3.3); allow
	// paths combine per capability, not per whole role (Spec R3); Organization owner/admin paths include unlinked
	// profiles (Spec R2); the flag affects only the inherited update (Spec R3, R4); member-only reads only (Spec R3).
	{"S23", "group_owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S23", "group_owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S23", "group_owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S23", "group_owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.privileged_settings.manage", oracleDeny},
	{"S24", "org_owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S24", "org_owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S24", "org_owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S24", "org_owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.privileged_settings.manage", oracleAllow},
	{"S25", "org_admin", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S25", "org_admin", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S25", "org_admin", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S25", "org_admin", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.privileged_settings.manage", oracleDeny},
	{"S26", "group_owner", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S26", "group_owner", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S26", "group_owner", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S27", "g3admin", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S27", "g3admin", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S27", "g3admin", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S27", "g3admin", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.privileged_settings.manage", oracleDeny},
	{"S28", "g3admin", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOff, "business_profile.profile.read", oracleAllow},
	{"S28", "g3admin", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOff, "business_profile.profile.update", oracleDeny},
	{"S28", "g3admin", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOff, "business_profile.select_for_onboarding", oracleAllow},
	{"S29", "g3admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S29", "g3admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S29", "g3admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S30", "g1member+g2admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S30", "g1member+g2admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S30", "g1member+g2admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S31", "g1member+g2admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOff, "business_profile.profile.read", oracleAllow},
	{"S31", "g1member+g2admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOff, "business_profile.profile.update", oracleDeny},
	{"S31", "g1member+g2admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOff, "business_profile.select_for_onboarding", oracleAllow},
	{"S32", "group_member", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S32", "group_member", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S32", "group_member", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S32", "group_member", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.privileged_settings.manage", oracleDeny},
	{"S33", "org_member", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S34", "org_member", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S35", "g1owner+g2owner", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S35", "g1owner+g2owner", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S36", "g4owner", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S36", "g4owner", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S36", "g4owner", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S37", "g4owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S37", "g4owner", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},

	// linked in selected-Group context: only the associated selected Group counts (Spec R2; Plan r1 §3.3 candidate
	// Groups = {selected Group}); Organization owner/admin profile paths are independent of the selected Group and
	// include unlinked profiles (Plan r1 §3.3; Spec R2); a new unattached Group cannot authorize P1 (Spec AC4).
	{"S38", "g1owner+g3admin", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S38", "g1owner+g3admin", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S38", "g1owner+g3admin", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S39", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S39", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S39", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S40", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOff, "business_profile.profile.read", oracleAllow},
	{"S40", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOff, "business_profile.profile.update", oracleDeny},
	{"S40", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOff, "business_profile.select_for_onboarding", oracleAllow},
	{"S41", "g1owner+g3admin", oracleGrp3, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S41", "g1owner+g3admin", oracleGrp3, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S41", "g1owner+g3admin", oracleGrp3, oracleP(oraclePrf2), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S42", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S42", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S42", "g1owner+g3admin", oracleGrp1, oracleP(oraclePrf2), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S43", "g1member+g2admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S43", "g1member+g2admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S43", "g1member+g2admin", oracleGrp1, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S44", "g1member+g2admin", oracleGrp2, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S44", "g1member+g2admin", oracleGrp2, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S44", "g1member+g2admin", oracleGrp2, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S45", "org_admin+g3member", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S45", "org_admin+g3member", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S45", "org_admin+g3member", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S45", "org_admin+g3member", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.privileged_settings.manage", oracleDeny},
	{"S46", "org_admin+g3member", oracleGrp3, oracleP(oraclePrf1), oracleFlagOff, "business_profile.profile.update", oracleDeny},
	{"S47", "org_owner", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S47", "org_owner", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S47", "org_owner", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},
	{"S47", "org_owner", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.privileged_settings.manage", oracleAllow},
	{"S48", "org_owner", oracleGrp1, oracleP(oraclePrf3), oracleFlagOff, "business_profile.profile.read", oracleAllow},
	{"S48", "org_owner", oracleGrp1, oracleP(oraclePrf3), oracleFlagOff, "business_profile.profile.update", oracleDeny},
	{"S48", "org_owner", oracleGrp1, oracleP(oraclePrf3), oracleFlagOff, "business_profile.select_for_onboarding", oracleDeny},
	{"S48", "org_owner", oracleGrp1, oracleP(oraclePrf3), oracleFlagOff, "business_profile.privileged_settings.manage", oracleAllow},
	{"S49", "group_owner", oracleGrp1, oracleP(oraclePrf1), oracleFlagOff, "business_profile.profile.read", oracleAllow},
	{"S49", "group_owner", oracleGrp1, oracleP(oraclePrf1), oracleFlagOff, "business_profile.profile.update", oracleDeny},
	{"S49", "group_owner", oracleGrp1, oracleP(oraclePrf1), oracleFlagOff, "business_profile.select_for_onboarding", oracleAllow},
	{"S49", "group_owner", oracleGrp1, oracleP(oraclePrf1), oracleFlagOff, "business_profile.privileged_settings.manage", oracleDeny},
	{"S50", "g4owner", oracleGrp4, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.read", oracleDeny},
	{"S50", "g4owner", oracleGrp4, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleDeny},
	{"S50", "g4owner", oracleGrp4, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleDeny},

	// No Group path confers privileged settings, Organization policy, member management or audit (Spec R3); a
	// selected Group does not restrict the Organization capabilities, which target the active Organization (Spec R2;
	// Plan r1 §3.3 gate 3 applies to group.* only).
	{"S51", "g1owner+g3admin", oracleOrgOnly, oracleP(oraclePrf1), oracleFlagOn, "business_profile.privileged_settings.manage", oracleDeny},
	{"S52", "g1owner+g3admin", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "organization.read", oracleAllow},
	{"S52", "g1owner+g3admin", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "organization.manage_members", oracleDeny},
	{"S52", "g1owner+g3admin", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "organization.manage_policy", oracleDeny},
	{"S52", "g1owner+g3admin", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "audit.read", oracleDeny},
	{"S53", "org_admin+g3member", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.profile.read", oracleAllow},
	{"S53", "org_admin+g3member", oracleOrgOnly, oracleP(oraclePrf2), oracleFlagOn, "business_profile.privileged_settings.manage", oracleDeny},
	{"S54", "org_admin+g3member", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "organization.read", oracleAllow},
	{"S54", "org_admin+g3member", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "organization.manage_members", oracleAllow},
	{"S54", "org_admin+g3member", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "organization.manage_policy", oracleDeny},
	{"S54", "org_admin+g3member", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "audit.read", oracleDeny},
	{"S55", "org_owner", oracleGrp2, oracleO(oracleOrgA), oracleFlagOn, "organization.read", oracleAllow},
	{"S55", "org_owner", oracleGrp2, oracleO(oracleOrgA), oracleFlagOn, "organization.manage_policy", oracleAllow},
	{"S55", "org_owner", oracleGrp2, oracleO(oracleOrgA), oracleFlagOn, "audit.read", oracleAllow},
	{"S56", "group_admin", oracleGrp1, oracleO(oracleOrgA), oracleFlagOn, "organization.read", oracleAllow},
	{"S56", "group_admin", oracleGrp1, oracleO(oracleOrgA), oracleFlagOn, "organization.manage_members", oracleDeny},

	// G060 (ADR-008; Spec R2): the same Organization owner/admin subjects with the Organization's Group-management
	// override enabled ("+ovr") administer every Group of A and every profile (unlinked included) without a Group binding,
	// restricted to the selected Group; the default (above) denies them every ordinary administration capability while
	// keeping the metadata reads. A selected other Group, privileged settings, Group ownership and leave stay denied.
	{"S04o", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleAllow},
	{"S04o", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.manage_members", oracleAllow},
	{"S05o", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp4), oracleFlagOn, "group.update", oracleAllow},
	{"S06o", "org_admin+ovr", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.update", oracleAllow},
	{"S06o", "org_admin+ovr", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleAllow},
	{"S04o", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.complete_onboarding", oracleAllow},
	{"S05o", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp4), oracleFlagOn, "group.complete_onboarding", oracleAllow},
	{"S15o", "org_admin+ovr", oracleGrp2, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleAllow},
	{"S15o", "org_admin+ovr", oracleGrp2, oracleG(oracleGrp2), oracleFlagOn, "group.manage_members", oracleAllow},
	{"S20o", "org_admin+g3member+ovr", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.update", oracleAllow},
	{"S20o", "org_admin+g3member+ovr", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleAllow},
	{"S24o", "org_owner+ovr", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S24o", "org_owner+ovr", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S25o", "org_admin+ovr", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S25o", "org_admin+ovr", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S45o", "org_admin+g3member+ovr", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S45o", "org_admin+g3member+ovr", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S46o", "org_admin+g3member+ovr", oracleGrp3, oracleP(oraclePrf1), oracleFlagOff, "business_profile.profile.update", oracleAllow},
	{"S47o", "org_owner+ovr", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.profile.update", oracleAllow},
	{"S47o", "org_owner+ovr", oracleGrp3, oracleP(oraclePrf1), oracleFlagOn, "business_profile.select_for_onboarding", oracleAllow},
	{"S48o", "org_owner+ovr", oracleGrp1, oracleP(oraclePrf3), oracleFlagOff, "business_profile.profile.update", oracleAllow},
	{"S48o", "org_owner+ovr", oracleGrp1, oracleP(oraclePrf3), oracleFlagOff, "business_profile.select_for_onboarding", oracleAllow},
	{"S13o", "org_owner+ovr", oracleGrp1, oracleG(oracleGrp2), oracleFlagOn, "group.update", oracleDeny},
	{"S14o", "org_admin+ovr", oracleGrp1, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleDeny},
	{"S25o", "org_admin+ovr", oracleOrgOnly, oracleP(oraclePrf3), oracleFlagOn, "business_profile.privileged_settings.manage", oracleDeny},
	{"S04r", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.manage_owners", oracleDeny},
	{"S04r", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "group.leave", oracleDeny},

	// G020 (Plan r2 §1, §5.2 T1): the label capabilities. Organization-owned labels belong to Organization owners/admins
	// in either context; Group-owned labels need an own role in the exact Group: an Organization admin (frank) without
	// one is denied with the override off and on, in Organization-only and selected context; a dual-role subject (Grace)
	// holds both paths independently and gains nothing in a Group where she has no role; a Group member (dave) reads only
	// his own Group's labels and never manages or reads the Organization's.
	{"L01", "org_admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L01", "org_admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.read", oracleDeny},
	{"L01", "org_admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "group.read", oracleAllow},
	{"L02", "org_admin+ovr", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L02", "org_admin+ovr", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.read", oracleDeny},
	{"L02", "org_admin+ovr", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "group.update", oracleAllow},
	{"L03", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp4), oracleFlagOff, "label.group.manage", oracleDeny},
	{"L03", "org_owner+ovr", oracleOrgOnly, oracleG(oracleGrp4), oracleFlagOff, "label.group.read", oracleDeny},
	{"L04", "org_admin+ovr", oracleGrp1, oracleG(oracleGrp1), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L04", "org_admin+ovr", oracleGrp1, oracleG(oracleGrp1), oracleFlagOn, "label.group.read", oracleDeny},
	{"L05", "org_admin+ovr", oracleGrp1, oracleO(oracleOrgA), oracleFlagOn, "label.organization.manage", oracleAllow},
	{"L06", "org_admin+g1admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.manage", oracleAllow},
	{"L06", "org_admin+g1admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.read", oracleAllow},
	{"L07", "org_admin+g1admin", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "label.organization.manage", oracleAllow},
	{"L08", "org_admin+g1admin", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L08", "org_admin+g1admin", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "label.group.read", oracleDeny},
	{"L09", "org_admin+g1admin+ovr", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L09", "org_admin+g1admin+ovr", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "label.group.read", oracleDeny},
	{"L10", "org_admin+g1admin", oracleGrp1, oracleG(oracleGrp1), oracleFlagOn, "label.group.manage", oracleAllow},
	{"L11", "org_admin+g1admin", oracleGrp2, oracleG(oracleGrp1), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L12", "group_member", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.read", oracleAllow},
	{"L12", "group_member", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L13", "group_member", oracleOrgOnly, oracleG(oracleGrp2), oracleFlagOn, "label.group.read", oracleDeny},
	{"L14", "group_member", oracleGrp1, oracleO(oracleOrgA), oracleFlagOn, "label.organization.manage", oracleDeny},
	{"L15", "group_owner", oracleGrp1, oracleG(oracleGrp1), oracleFlagOff, "label.group.manage", oracleAllow},
	{"L16", "group_owner", oracleGrp1, oracleG(oracleGrp2), oracleFlagOff, "label.group.read", oracleDeny},
	{"L17", "org_admin+g3member", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "label.group.read", oracleAllow},
	{"L17", "org_admin+g3member", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L18", "org_admin+g3member+ovr", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L18", "org_admin+g3member+ovr", oracleGrp3, oracleG(oracleGrp3), oracleFlagOn, "group.manage_members", oracleAllow},
	{"L19", "org_owner", oracleGrp2, oracleO(oracleOrgA), oracleFlagOff, "label.organization.manage", oracleAllow},
	{"L20", "org_member", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "label.organization.manage", oracleDeny},
	{"L21", "g1owner+g3admin", oracleOrgOnly, oracleO(oracleOrgA), oracleFlagOn, "label.organization.manage", oracleDeny},
	{"L22", "g1owner+g3admin", oracleOrgOnly, oracleG(oracleGrp3), oracleFlagOn, "label.group.manage", oracleAllow},
	{"L23", "g1member+g2admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.manage", oracleDeny},
	{"L23", "g1member+g2admin", oracleOrgOnly, oracleG(oracleGrp1), oracleFlagOn, "label.group.read", oracleAllow},
}

// oracleManageableCase is one literal manageable-profile set (pin Q8(d): the profiles P of the active Organization for
// which decide(business_profile.select_for_onboarding, P) allows in the current context). want lists the allowed
// candidates in oracleCandidates order, separated by single spaces.
type oracleManageableCase struct {
	id      string
	subject string
	sel     tenancy.GroupID
	flagOn  bool
	want    string
}

// oracleManageableCases are the manageable-set expectations (Spec R3 paragraph on the management list; Spec AC4 and
// AC7): distinct profiles, member-only reading does not make a profile manageable, selected context limits only the
// Group-derived paths, the flag does not affect selection eligibility and B's profile is never listed.
var oracleManageableCases = []oracleManageableCase{
	{"M01", "org_owner", oracleOrgOnly, oracleFlagOn, ""},
	{"M02", "org_admin", oracleOrgOnly, oracleFlagOn, ""},
	{"M03", "org_member", oracleOrgOnly, oracleFlagOn, ""},
	{"M04", "group_owner", oracleOrgOnly, oracleFlagOn, "P1"},
	{"M05", "group_admin", oracleOrgOnly, oracleFlagOff, "P1"},
	{"M06", "group_member", oracleOrgOnly, oracleFlagOn, ""},
	{"M07", "g3admin", oracleOrgOnly, oracleFlagOn, "P2"},
	{"M08", "g1owner+g3admin", oracleOrgOnly, oracleFlagOn, "P1 P2"},
	{"M09", "g1owner+g3admin", oracleGrp1, oracleFlagOn, "P1"},
	{"M10", "g1owner+g3admin", oracleGrp3, oracleFlagOn, "P2"},
	{"M11", "g1member+g2admin", oracleOrgOnly, oracleFlagOn, "P1"},
	{"M12", "g1member+g2admin", oracleGrp1, oracleFlagOn, ""},
	{"M13", "g1member+g2admin", oracleGrp2, oracleFlagOn, "P1"},
	{"M14", "g1owner+g2owner", oracleOrgOnly, oracleFlagOn, "P1"},
	{"M15", "g4owner", oracleOrgOnly, oracleFlagOn, ""},
	{"M16", "g4owner", oracleGrp4, oracleFlagOn, ""},
	{"M17", "org_admin+g3member", oracleGrp3, oracleFlagOn, ""},
	{"M18", "org_owner", oracleGrp2, oracleFlagOff, ""},
	{"M19", "g1owner-no-org", oracleOrgOnly, oracleFlagOn, ""},
	{"M20", "group_member", oracleGrp1, oracleFlagOn, ""},
	// G060: the selectable set of Organization owners/admins is every profile only while the override is enabled.
	{"M21", "org_owner+ovr", oracleOrgOnly, oracleFlagOn, "P1 P2 P3"},
	{"M22", "org_admin+ovr", oracleOrgOnly, oracleFlagOff, "P1 P2 P3"},
	{"M23", "org_admin+g3member+ovr", oracleGrp3, oracleFlagOn, "P1 P2 P3"},
	{"M24", "org_owner+ovr", oracleGrp2, oracleFlagOff, "P1 P2 P3"},
}

// oracleNamedProfile labels a candidate profile of the manageable-set cases.
type oracleNamedProfile struct {
	label string
	id    tenancy.BusinessProfileID
}

// oracleCandidates are the profiles probed for the manageable set: A's three profiles and B's PB.
var oracleCandidates = []oracleNamedProfile{{"P1", oraclePrf1}, {"P2", oraclePrf2}, {"P3", oraclePrf3}, {"PB", oraclePrfB}}

// Unknown inputs. None may be mapped to a known capability, role, source or kind: no prefix, substring or case
// matching, no ranking and no removed capability (Spec R2 and R3).
var oracleUnknownCapabilities = []string{
	"", "*", "organization", "organization.", "organization.*", "organization.delete", "ORGANIZATION.READ",
	"organization.read ", " organization.read", "organization.read.all", "organization.manage_mfa_policy", "group",
	"group.*", "group.transfer_ownership", "business_profile.read", "business_profile.profile",
	"business_profile.profile.*", "business_profile.transfer_ownership", "business_profile.privileged_settings",
	"audit", "audit.write", "billing.manage",
}

var oracleUnknownOrgRoles = []tenancy.OrganizationRole{
	"", "owner", "admin", "member", "org_superuser", "ORG_OWNER", "org_owner ", "org_", "group_owner",
	"org_owner,org_admin",
}

var oracleUnknownGroupRoles = []tenancy.GroupRole{
	"", "owner", "admin", "group_superuser", "GROUP_OWNER", "group_owner ", "group_", "org_owner",
	"group_member,group_owner",
}

var oracleUnknownSources = []string{"", "invitation", "system", "Membership", "membership "}

var oracleUnknownKinds = []TargetKind{"", "unknown", "Group", "profile", "business_profile.profile", "organization "}

// oracleMustOrg parses a fixture Organization ID.
func oracleMustOrg(s string) tenancy.OrganizationID {
	id, err := tenancy.ParseOrganizationID(s)
	if err != nil {
		panic("oracle: invalid fixture ID " + s)
	}
	return id
}

// oracleMustGroup parses a fixture Group ID.
func oracleMustGroup(s string) tenancy.GroupID {
	id, err := tenancy.ParseGroupID(s)
	if err != nil {
		panic("oracle: invalid fixture ID " + s)
	}
	return id
}

// oracleMustProfile parses a fixture BusinessProfile ID.
func oracleMustProfile(s string) tenancy.BusinessProfileID {
	id, err := tenancy.ParseBusinessProfileID(s)
	if err != nil {
		panic("oracle: invalid fixture ID " + s)
	}
	return id
}

// oracleGroupRow is a Group row of the facts.
func oracleGroupRow(org tenancy.OrganizationID, id tenancy.GroupID, name, slug string) tenancy.Group {
	return tenancy.Group{ID: id, OrganizationID: org, Name: name, Slug: slug}
}

// oracleProfileRow is a BusinessProfile row of the facts.
func oracleProfileRow(org tenancy.OrganizationID, id tenancy.BusinessProfileID, name string) tenancy.BusinessProfile {
	return tenancy.BusinessProfile{ID: id, OrganizationID: org, Context: tenancy.BusinessContext{DisplayName: name}}
}

// oracleAssociation is a current GroupBusinessProfile association of the facts.
func oracleAssociation(g tenancy.GroupID, p tenancy.BusinessProfileID) tenancy.AssociationFact {
	return tenancy.AssociationFact{GroupID: g, BusinessProfileID: p}
}

// oracleOrgBinding is the subject's Organization-scope membership binding in A with the given role.
func oracleOrgBinding(role tenancy.OrganizationRole) tenancy.BindingFact {
	return tenancy.BindingFact{OrganizationID: oracleOrgA, Role: string(role), Source: tenancy.SourceMembership}
}

// oracleGroupBinding is the subject's Group-scope membership binding in a Group of A with the given role.
func oracleGroupBinding(g tenancy.GroupID, role tenancy.GroupRole) tenancy.BindingFact {
	return tenancy.BindingFact{OrganizationID: oracleOrgA, GroupID: g, Role: string(role), Source: tenancy.SourceMembership}
}

// oracleBaseFacts returns the facts of Organization A without any subject binding.
func oracleBaseFacts(flagOn bool) Facts {
	var f Facts
	f.Organization = tenancy.Organization{ID: oracleOrgA, Name: "Organization A"}
	f.Policy = tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: flagOn}
	f.Groups = []tenancy.Group{
		oracleGroupRow(oracleOrgA, oracleGrp1, "Group 1", "group-1"),
		oracleGroupRow(oracleOrgA, oracleGrp2, "Group 2", "group-2"),
		oracleGroupRow(oracleOrgA, oracleGrp3, "Group 3", "group-3"),
		oracleGroupRow(oracleOrgA, oracleGrp4, "Group 4", "group-4"),
	}
	f.Associations = []tenancy.AssociationFact{
		oracleAssociation(oracleGrp1, oraclePrf1),
		oracleAssociation(oracleGrp2, oraclePrf1),
		oracleAssociation(oracleGrp3, oraclePrf2),
	}
	f.Profiles = []tenancy.BusinessProfile{
		oracleProfileRow(oracleOrgA, oraclePrf1, "Profile 1"),
		oracleProfileRow(oracleOrgA, oraclePrf2, "Profile 2"),
		oracleProfileRow(oracleOrgA, oraclePrf3, "Profile 3"),
	}
	return f
}

// oracleSubjectFacts returns the complete facts of A for a named subject. The six column subjects follow pin Q8(a):
// an Organization role alone, or org_member plus the Group role in g1. The others are org_member plus the named Group
// roles, org_admin plus group_member in g3, a Group binding in g1 without any Organization binding, and no binding.
func oracleSubjectFacts(t testing.TB, subject string, flagOn bool) Facts {
	t.Helper()
	// G060: a "+ovr" suffix names the same subject with the Organization's Group-management override enabled.
	override := strings.HasSuffix(subject, "+ovr")
	subject = strings.TrimSuffix(subject, "+ovr")
	f := oracleBaseFacts(flagOn)
	f.Policy.GroupManagementOverride = override
	member := []tenancy.BindingFact{oracleOrgBinding(tenancy.OrgMember)}
	switch subject {
	case string(tenancy.OrgOwner):
		f.OrganizationBindings = []tenancy.BindingFact{oracleOrgBinding(tenancy.OrgOwner)}
	case string(tenancy.OrgAdmin):
		f.OrganizationBindings = []tenancy.BindingFact{oracleOrgBinding(tenancy.OrgAdmin)}
	case string(tenancy.OrgMember):
		f.OrganizationBindings = member
	case string(tenancy.GroupOwner):
		f.OrganizationBindings = member
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, tenancy.GroupOwner)}
	case string(tenancy.GroupAdmin):
		f.OrganizationBindings = member
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, tenancy.GroupAdmin)}
	case string(tenancy.GroupMember):
		f.OrganizationBindings = member
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, tenancy.GroupMember)}
	case "g1owner+g2owner":
		f.OrganizationBindings = member
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, tenancy.GroupOwner), oracleGroupBinding(oracleGrp2, tenancy.GroupOwner)}
	case "g1owner+g3admin":
		f.OrganizationBindings = member
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, tenancy.GroupOwner), oracleGroupBinding(oracleGrp3, tenancy.GroupAdmin)}
	case "g1member+g2admin":
		f.OrganizationBindings = member
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, tenancy.GroupMember), oracleGroupBinding(oracleGrp2, tenancy.GroupAdmin)}
	case "g3admin":
		f.OrganizationBindings = member
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp3, tenancy.GroupAdmin)}
	case "g4owner":
		f.OrganizationBindings = member
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp4, tenancy.GroupOwner)}
	case "org_admin+g3member":
		f.OrganizationBindings = []tenancy.BindingFact{oracleOrgBinding(tenancy.OrgAdmin)}
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp3, tenancy.GroupMember)}
	case "org_admin+g1admin":
		// G020: the dual-role subject (fixture Grace: A admin and research admin), whose Organization and Group paths are
		// independent and never combine.
		f.OrganizationBindings = []tenancy.BindingFact{oracleOrgBinding(tenancy.OrgAdmin)}
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, tenancy.GroupAdmin)}
	case "g1owner-no-org":
		f.GroupBindings = []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, tenancy.GroupOwner)}
	case "no-bindings":
	default:
		t.Fatalf("oracle: unknown subject %q", subject)
	}
	return f
}

// oracleSelOf is the context selection of Organization A with the given selected Group (oracleOrgOnly for none).
func oracleSelOf(g tenancy.GroupID) contextSelection {
	return contextSelection{Organization: oracleOrgA, Group: g}
}

// oracleO is an Organization target.
func oracleO(o tenancy.OrganizationID) Target {
	return Target{Kind: "organization", Organization: o}
}

// oracleG is a Group target.
func oracleG(g tenancy.GroupID) Target {
	return Target{Kind: "group", Group: g}
}

// oracleP is a BusinessProfile target.
func oracleP(p tenancy.BusinessProfileID) Target {
	return Target{Kind: "business_profile", Profile: p}
}

// oracleNaturalTarget is the primary-scenario target of a capability's Spec R2 kind: Organization A, g1 or P1.
func oracleNaturalTarget(t testing.TB, capability string) Target {
	t.Helper()
	for _, c := range oracleCapabilities {
		if c.name != capability {
			continue
		}
		switch c.kind {
		case "organization":
			return oracleO(oracleOrgA)
		case "group":
			return oracleG(oracleGrp1)
		case "business_profile":
			return oracleP(oraclePrf1)
		}
	}
	t.Fatalf("oracle: %q is not one of the 22 capabilities", capability)
	return Target{}
}

// oracleLoaded returns the subset of the complete facts f that the Plan r1 §3.7 fact query loads for one decision.
// The candidate Groups are the selected Group in selected context, the target Group for a Group target in
// Organization-only context, and otherwise every Group of the subject. It keeps the Organization row, the policy and
// the Organization bindings, the subject's Group bindings for the candidates, the candidate and target Group rows,
// their associations, and only the target profile.
func oracleLoaded(f Facts, sel contextSelection, tg Target) Facts {
	candidates := map[tenancy.GroupID]bool{}
	switch {
	case sel.Group != oracleOrgOnly:
		candidates[sel.Group] = true
	case tg.Kind == "group":
		candidates[tg.Group] = true
	default:
		for _, b := range f.GroupBindings {
			candidates[b.GroupID] = true
		}
	}
	rows := map[tenancy.GroupID]bool{}
	for g := range candidates {
		rows[g] = true
	}
	if tg.Kind == "group" {
		rows[tg.Group] = true
	}
	var out Facts
	out.Organization = f.Organization
	out.Policy = f.Policy
	out.OrganizationBindings = f.OrganizationBindings
	for _, b := range f.GroupBindings {
		if candidates[b.GroupID] {
			out.GroupBindings = append(out.GroupBindings, b)
		}
	}
	for _, g := range f.Groups {
		if rows[g.ID] {
			out.Groups = append(out.Groups, g)
		}
	}
	for _, a := range f.Associations {
		if rows[a.GroupID] {
			out.Associations = append(out.Associations, a)
		}
	}
	for _, p := range f.Profiles {
		if tg.Kind == "business_profile" && p.ID == tg.Profile {
			out.Profiles = append(out.Profiles, p)
		}
	}
	return out
}

// oracleParseCells splits one literal row strictly: exactly six space-separated cells, each A, D or X.
func oracleParseCells(t testing.TB, row oracleRow) []string {
	t.Helper()
	cells := strings.Split(row.cells, " ")
	if len(cells) != len(oracleColumns) {
		t.Fatalf("oracle table: %s has %d cells, want %d", row.capability, len(cells), len(oracleColumns))
	}
	for _, c := range cells {
		if c != "A" && c != "D" && c != "X" {
			t.Fatalf("oracle table: %s has cell %q", row.capability, c)
		}
	}
	return cells
}

// oracleExpectPrimary returns the literal primary-scenario expectation of one cell and whether it is asserted. It is
// also the accessor for IT-MATRIX-1 (Plan r1 §6.3), which applies the same table to DB-loaded facts; selected means the
// selected Group g1.
func oracleExpectPrimary(t testing.TB, selected, flagOn bool, capability, column string) (bool, bool) {
	t.Helper()
	return oracleExpectPrimaryOverride(t, selected, flagOn, false, capability, column)
}

// oracleExpectPrimaryOverride is oracleExpectPrimary for a given value of the G060 Group-management override (the
// override-on tables exist for Organization-only/flag-on and selected-g1/flag-off).
func oracleExpectPrimaryOverride(t testing.TB, selected, flagOn, override bool, capability, column string) (bool, bool) {
	t.Helper()
	col := slices.Index(oracleColumns, column)
	if col < 0 {
		t.Fatalf("oracle: unknown column %q", column)
	}
	for _, tbl := range oraclePrimaryTables() {
		if tbl.selected != selected || tbl.flagOn != flagOn || tbl.override != override {
			continue
		}
		for _, row := range tbl.rows {
			if row.capability == capability {
				cell := oracleParseCells(t, row)[col]
				return cell == "A", cell != "X"
			}
		}
	}
	t.Fatalf("oracle: no primary cell for %q", capability)
	return false, false
}

// oracleExpectShape asserts the allow result of decide, called by its pinned signature, for one fact shape.
func oracleExpectShape(t *testing.T, shape string, f Facts, sel contextSelection, tg Target, capability string, want bool) {
	t.Helper()
	if got, _ := decide(Capability(capability), sel, f, tg); got != want {
		t.Errorf("decide(%q) with the %s facts: allow = %t, want %t", capability, shape, got, want)
	}
}

// oracleExpectBoth asserts the same expectation for the complete facts and for their loaded subset.
func oracleExpectBoth(t *testing.T, f Facts, sel contextSelection, tg Target, capability string, want bool) {
	t.Helper()
	oracleExpectShape(t, "complete", f, sel, tg, capability, want)
	oracleExpectShape(t, "loaded", oracleLoaded(f, sel, tg), sel, tg, capability, want)
}

// TestOracleTables checks the shape of the literal tables and their only intended differences, so that a transcription
// slip in this oracle fails here rather than as a policy disagreement. It never calls decide.
func TestOracleTables(t *testing.T) {
	wantColumns := []string{"org_owner", "org_admin", "org_member", "group_owner", "group_admin", "group_member"}
	if !slices.Equal(oracleColumns, wantColumns) {
		t.Fatalf("columns %q, want %q", oracleColumns, wantColumns)
	}
	if len(oracleCapabilities) != 22 {
		t.Fatalf("%d capabilities, want 22 (Spec R3's 11 plus G016 organization.manage_domains, G017 organization.manage_branding, G018 group.create and group.complete_onboarding, G019 organization.manage_owners, group.manage_owners, organization.leave and group.leave and G020 label.organization.manage, label.group.manage and label.group.read)", len(oracleCapabilities))
	}
	tables := oraclePrimaryTables()
	for _, tbl := range tables {
		if len(tbl.rows) != len(oracleCapabilities) {
			t.Fatalf("%s: %d rows, want %d", tbl.name, len(tbl.rows), len(oracleCapabilities))
		}
		for i, row := range tbl.rows {
			if row.capability != oracleCapabilities[i].name {
				t.Errorf("%s row %d: %q, want %q (Spec R3 order)", tbl.name, i, row.capability, oracleCapabilities[i].name)
			}
			oracleParseCells(t, row)
		}
	}
	diff := func(a, b oracleTable) []string {
		var out []string
		for i := range a.rows {
			ca, cb := oracleParseCells(t, a.rows[i]), oracleParseCells(t, b.rows[i])
			for j := range ca {
				if ca[j] != cb[j] {
					out = append(out, a.rows[i].capability+"/"+oracleColumns[j])
				}
			}
		}
		return out
	}
	// The flag values differ only in the two linked+flag cells (Spec R3, R4). Organization-only and selected-g1
	// context differ only in the org_member column, for which no selected Group context exists, and in every column of
	// G018's group.create, which only Organization-only context admits.
	flagDiff := []string{"business_profile.profile.update/group_owner", "business_profile.profile.update/group_admin"}
	var contextDiff []string
	for _, c := range oracleCapabilities {
		if c.name == "group.create" {
			for _, column := range oracleColumns {
				contextDiff = append(contextDiff, c.name+"/"+column)
			}
			continue
		}
		contextDiff = append(contextDiff, c.name+"/org_member")
	}
	if got := diff(tables[0], tables[1]); !slices.Equal(got, flagDiff) {
		t.Errorf("org-only flag values differ at %q, want %q", got, flagDiff)
	}
	if got := diff(tables[2], tables[3]); !slices.Equal(got, flagDiff) {
		t.Errorf("selected-g1 flag values differ at %q, want %q", got, flagDiff)
	}
	if got := diff(tables[0], tables[2]); !slices.Equal(got, contextDiff) {
		t.Errorf("flag-on context modes differ at %q, want %q", got, contextDiff)
	}
	if got := diff(tables[1], tables[3]); !slices.Equal(got, contextDiff) {
		t.Errorf("flag-off context modes differ at %q, want %q", got, contextDiff)
	}
	// G060: the override differs from the default only in the Organization owner/admin cells of the five ordinary Group
	// administration capabilities (group.update, group.manage_members, business_profile.profile.update,
	// business_profile.select_for_onboarding and group.complete_onboarding), in either context and flag value.
	var overrideDiff []string
	for _, c := range []string{"group.update", "group.manage_members", "business_profile.profile.update",
		"business_profile.select_for_onboarding", "group.complete_onboarding"} {
		overrideDiff = append(overrideDiff, c+"/org_owner", c+"/org_admin")
	}
	slices.SortFunc(overrideDiff, func(a, b string) int {
		ia, ib := slices.IndexFunc(oracleCapabilities, func(c oracleCapKind) bool { return strings.HasPrefix(a, c.name+"/") }),
			slices.IndexFunc(oracleCapabilities, func(c oracleCapKind) bool { return strings.HasPrefix(b, c.name+"/") })
		if ia != ib {
			return ia - ib
		}
		return strings.Compare(b, a) // org_owner before org_admin
	})
	if got := diff(tables[4], tables[0]); !slices.Equal(got, overrideDiff) {
		t.Errorf("org-only flag-on override values differ at %q, want %q", got, overrideDiff)
	}
	if got := diff(tables[5], tables[3]); !slices.Equal(got, overrideDiff) {
		t.Errorf("selected-g1 flag-off override values differ at %q, want %q", got, overrideDiff)
	}
	names := map[string]bool{}
	scenarios := map[string]string{}
	for _, c := range oracleScopeCases {
		oracleNaturalTarget(t, c.capability)
		oracleSubjectFacts(t, c.subject, c.flagOn)
		name := c.id + "/" + c.subject + "/" + c.capability
		if names[name] {
			t.Errorf("duplicate scope case %s", name)
		}
		names[name] = true
		scenario := fmt.Sprintf("%s|%v|%v|%t", c.subject, c.sel, c.target, c.flagOn)
		if prev, ok := scenarios[c.id]; ok && prev != scenario {
			t.Errorf("scope case %s names two scenarios", c.id)
		}
		scenarios[c.id] = scenario
	}
	ids := map[string]bool{}
	for _, c := range oracleManageableCases {
		oracleSubjectFacts(t, c.subject, c.flagOn)
		if ids[c.id] {
			t.Errorf("duplicate manageable case %s", c.id)
		}
		ids[c.id] = true
	}
}

// TestOracleMatrix (U-POL-MATRIX): every one of the 66 Spec R3 cells (11 capabilities by 6 roles), plus the six cells of
// G016's organization.manage_domains, the six of G017's organization.manage_branding and the twelve of G018's
// group.create and group.complete_onboarding, in its primary
// scenario, in Organization-only and selected-g1 context, with both values of the R4 flag.
func TestOracleMatrix(t *testing.T) {
	// Spec R3 as written. Y: a role at the target scope; all: any Group in the active Organization; own: the exact
	// Group binding; linked: a current authorized Group plus its current BusinessProfile association; linked+flag: also
	// the R4 flag; -: deny. The Organization-membership and context gates apply to every cell.
	//
	// capability                                   org_owner  org_admin  org_member  group_owner  group_admin  group_member
	// organization.read                            Y          Y          Y           -            -            -
	// organization.manage_members                  Y          Y          -           -            -            -
	// organization.manage_policy                   Y          -          -           -            -            -
	// group.read                                   all        all        -           own          own          own
	// group.update                                 all        all        -           own          own          -
	// group.manage_members                         all        all        -           own          own          -
	// business_profile.profile.read                Y          Y          -           linked       linked       linked
	// business_profile.profile.update              Y          Y          -           linked+flag  linked+flag  -
	// business_profile.select_for_onboarding       Y          Y          -           linked       linked       -
	// business_profile.privileged_settings.manage  Y          -          -           -            -            -
	// audit.read                                   Y          -          -           -            -            -
	// organization.manage_domains (G016)           Y          Y          -           -            -            -
	// organization.manage_branding (G017)          Y          Y          -           -            -            -
	// group.create (G018; Organization-only)      Y          Y          Y           -            -            -
	// group.complete_onboarding (G018)             all        all        -           own          own          -
	// organization.manage_owners (G019)            Y          -          -           -            -            -
	// group.manage_owners (G019)                   -          -          -           own          -            -
	// organization.leave (G019)                    Y          Y          Y           -            -            -
	// group.leave (G019)                           -          -          -           own          own          own
	// label.organization.manage (G020)             Y          Y          -           -            -            -
	// label.group.manage (G020; no override)       -          -          -           own          own          -
	// label.group.read (G020; no override)         -          -          -           own          own          own
	//
	// The literal tables above restate these cells for concrete subjects: by pin Q8(a) a Group-role column also holds
	// org_member, so its organization.read cell is A through org_member.
	asserted := 0
	for _, tbl := range oraclePrimaryTables() {
		sel := oracleSelOf(oracleOrgOnly)
		if tbl.selected {
			sel = oracleSelOf(oracleGrp1)
		}
		for _, c := range oracleCapabilities {
			for _, column := range oracleColumns {
				want, ok := oracleExpectPrimaryOverride(t, tbl.selected, tbl.flagOn, tbl.override, c.name, column)
				if !ok {
					continue
				}
				asserted++
				name := tbl.name + "/" + c.name + "/" + column
				t.Run(name, func(t *testing.T) {
					f := oracleSubjectFacts(t, column, tbl.flagOn)
					f.Policy.GroupManagementOverride = tbl.override
					oracleExpectBoth(t, f, sel, oracleNaturalTarget(t, c.name), c.name, want)
				})
			}
		}
	}
	// 114 cells in each Organization-only table and 95 in each selected table, where org_member is not asserted (G018: the
	// two capabilities add 6+6+5+5 cells each, 286 -> 330; G019: the four capabilities add 6+6+5+5 each, 330 -> 418;
	// G060: the two override-on tables add 114 + 95, 418 -> 627; G020: the three label capabilities add 6+6+5+5+6+5 each,
	// with explicit deny cells for every Organization-role cell of label.group.* in the override-on tables, 627 -> 726).
	if asserted != 726 {
		t.Fatalf("asserted %d primary cells, want 726", asserted)
	}
}

// TestOracleScope (U-POL-MATRIX scope variants): own, all, linked and linked+flag paths with the selected-Group
// restriction and the independent Organization paths.
func TestOracleScope(t *testing.T) {
	for _, c := range oracleScopeCases {
		name := c.id + "/" + c.subject + "/" + c.capability
		t.Run(name, func(t *testing.T) {
			f := oracleSubjectFacts(t, c.subject, c.flagOn)
			oracleExpectBoth(t, f, oracleSelOf(c.sel), c.target, c.capability, c.want)
		})
	}
}

// TestOracleManageable (pin Q8(d)): the manageable-profile set as the profiles that decide allows for
// business_profile.select_for_onboarding, compared with literal sets.
func TestOracleManageable(t *testing.T) {
	for _, c := range oracleManageableCases {
		name := c.id + "/" + c.subject
		t.Run(name, func(t *testing.T) {
			f := oracleSubjectFacts(t, c.subject, c.flagOn)
			sel := oracleSelOf(c.sel)
			var complete, loaded []string
			for _, p := range oracleCandidates {
				tg := oracleP(p.id)
				if got, _ := decide(Capability("business_profile.select_for_onboarding"), sel, f, tg); got {
					complete = append(complete, p.label)
				}
				if got, _ := decide(Capability("business_profile.select_for_onboarding"), sel, oracleLoaded(f, sel, tg), tg); got {
					loaded = append(loaded, p.label)
				}
			}
			if got := strings.Join(complete, " "); got != c.want {
				t.Errorf("manageable set with the complete facts %q, want %q", got, c.want)
			}
			if got := strings.Join(loaded, " "); got != c.want {
				t.Errorf("manageable set with the loaded facts %q, want %q", got, c.want)
			}
		})
	}
}

// TestOracleGates: the Organization gate held constant, the target gate and inconsistent context deny every capability
// before the matrix, whatever role the subject holds (Spec R1 and R2; Plan r1 §3.3 gates 1 and 2).
func TestOracleGates(t *testing.T) {
	// A Group binding without an Organization binding, or no binding at all, establishes no tenant access (Spec R1;
	// Spec AC2 and AC3), in both context modes and with both flag values.
	for _, subject := range []string{"g1owner-no-org", "no-bindings"} {
		for _, flagOn := range []bool{oracleFlagOn, oracleFlagOff} {
			for _, g := range []tenancy.GroupID{oracleOrgOnly, oracleGrp1} {
				for _, c := range oracleCapabilities {
					name := fmt.Sprintf("org-gate/%s/flag-%t/selected-%t/%s", subject, flagOn, g != oracleOrgOnly, c.name)
					t.Run(name, func(t *testing.T) {
						f := oracleSubjectFacts(t, subject, flagOn)
						oracleExpectShape(t, "complete", f, oracleSelOf(g), oracleNaturalTarget(t, c.name), c.name, oracleDeny)
					})
				}
			}
		}
	}
	// A mismatched target kind denies even when the caller holds the named role (Spec R2).
	for _, subject := range []string{"org_owner", "group_owner"} {
		for _, c := range oracleCapabilities {
			for _, tg := range []Target{oracleO(oracleOrgA), oracleG(oracleGrp1), oracleP(oraclePrf1)} {
				if tg.Kind == c.kind {
					continue
				}
				name := fmt.Sprintf("target-kind/%s/%s/%s", subject, c.name, tg.Kind)
				t.Run(name, func(t *testing.T) {
					f := oracleSubjectFacts(t, subject, oracleFlagOn)
					oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), tg, c.name, oracleDeny)
				})
			}
		}
	}
	// A foreign, unknown or all-zero target of the right kind denies (Spec R1 and R2: exact target ownership in the
	// active Organization; role grants never cross Organizations).
	for _, subject := range []string{"org_owner", "group_owner"} {
		for _, c := range oracleCapabilities {
			var targets []Target
			switch c.kind {
			case "organization":
				targets = []Target{oracleO(oracleOrgB), oracleO(tenancy.OrganizationID{})}
			case "group":
				targets = []Target{oracleG(oracleGrpB), oracleG(oracleGrpX), oracleG(tenancy.GroupID{})}
			case "business_profile":
				targets = []Target{oracleP(oraclePrfB), oracleP(oraclePrfX), oracleP(tenancy.BusinessProfileID{})}
			}
			for i, tg := range targets {
				name := fmt.Sprintf("foreign-target/%s/%s/%d", subject, c.name, i)
				t.Run(name, func(t *testing.T) {
					f := oracleSubjectFacts(t, subject, oracleFlagOn)
					oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), tg, c.name, oracleDeny)
				})
			}
		}
	}
	// Inconsistent context fails closed (Spec R1: all-zero IDs and inconsistent context; Plan r1 §3.3 gate 1: the facts
	// carry the active Organization), even for org_owner.
	inconsistent := func(label string, f Facts, sel contextSelection) {
		for _, c := range oracleCapabilities {
			name := "inconsistent-context/" + label + "/" + c.name
			t.Run(name, func(t *testing.T) {
				oracleExpectShape(t, "complete", f, sel, oracleNaturalTarget(t, c.name), c.name, oracleDeny)
			})
		}
	}
	absent := oracleSubjectFacts(t, "org_owner", oracleFlagOn)
	absent.Organization = tenancy.Organization{}
	inconsistent("facts-organization-absent", absent, oracleSelOf(oracleOrgOnly))
	foreign := oracleSubjectFacts(t, "org_owner", oracleFlagOn)
	foreign.Organization = tenancy.Organization{ID: oracleOrgB, Name: "Organization B"}
	inconsistent("facts-organization-foreign", foreign, oracleSelOf(oracleOrgOnly))
	inconsistent("selection-organization-zero", oracleSubjectFacts(t, "org_owner", oracleFlagOn), contextSelection{})
	inconsistent("selection-organization-foreign", oracleSubjectFacts(t, "org_owner", oracleFlagOn), contextSelection{Organization: oracleOrgB})
}

// TestOracleUnknownInputs: unknown capabilities, roles, sources and target kinds and inconsistent bindings deny (Spec
// R2; Plan r1 §3.3: a closed switch, no ranking, prefix matching or global authority).
func TestOracleUnknownInputs(t *testing.T) {
	for _, capability := range oracleUnknownCapabilities {
		for _, subject := range []string{"org_owner", "group_owner"} {
			for _, tg := range []Target{oracleO(oracleOrgA), oracleG(oracleGrp1), oracleP(oraclePrf1)} {
				name := fmt.Sprintf("capability/%q/%s/%s", capability, subject, tg.Kind)
				t.Run(name, func(t *testing.T) {
					f := oracleSubjectFacts(t, subject, oracleFlagOn)
					oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), tg, capability, oracleDeny)
				})
			}
		}
	}
	for _, role := range oracleUnknownOrgRoles {
		for _, c := range oracleCapabilities {
			name := fmt.Sprintf("org-role/%q/%s", role, c.name)
			t.Run(name, func(t *testing.T) {
				f := oracleBaseFacts(oracleFlagOn)
				f.OrganizationBindings = []tenancy.BindingFact{oracleOrgBinding(role)}
				oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), oracleNaturalTarget(t, c.name), c.name, oracleDeny)
			})
		}
	}
	for _, source := range oracleUnknownSources {
		for _, c := range oracleCapabilities {
			name := fmt.Sprintf("org-source/%q/%s", source, c.name)
			t.Run(name, func(t *testing.T) {
				f := oracleBaseFacts(oracleFlagOn)
				binding := oracleOrgBinding(tenancy.OrgOwner)
				binding.Source = source
				f.OrganizationBindings = []tenancy.BindingFact{binding}
				oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), oracleNaturalTarget(t, c.name), c.name, oracleDeny)
			})
		}
	}
	// Exactly one current Organization binding (Plan r1 §3.3 gate 1); an Organization binding naming a Group is
	// inconsistent.
	scoped := oracleOrgBinding(tenancy.OrgOwner)
	scoped.GroupID = oracleGrp1
	bindingSets := [][]tenancy.BindingFact{
		{oracleOrgBinding(tenancy.OrgOwner), oracleOrgBinding(tenancy.OrgMember)},
		{oracleOrgBinding(tenancy.OrgOwner), oracleOrgBinding(tenancy.OrgOwner)},
		{scoped},
	}
	for i, bindings := range bindingSets {
		for _, c := range oracleCapabilities {
			name := fmt.Sprintf("org-bindings/%d/%s", i, c.name)
			t.Run(name, func(t *testing.T) {
				f := oracleBaseFacts(oracleFlagOn)
				f.OrganizationBindings = bindings
				oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), oracleNaturalTarget(t, c.name), c.name, oracleDeny)
			})
		}
	}
	// An unknown or inconsistent Group-scope fact next to a valid org_member binding grants nothing. org_member's own
	// organization.read path is not asserted here: the Plan does not pin whether such a fact denies the whole decision
	// or only contributes no path (NOTES.md, ambiguity A2).
	var groupFacts [][]tenancy.BindingFact
	for _, role := range oracleUnknownGroupRoles {
		groupFacts = append(groupFacts, []tenancy.BindingFact{oracleGroupBinding(oracleGrp1, role)})
	}
	for _, source := range oracleUnknownSources {
		binding := oracleGroupBinding(oracleGrp1, tenancy.GroupOwner)
		binding.Source = source
		groupFacts = append(groupFacts, []tenancy.BindingFact{binding})
	}
	unscoped := oracleGroupBinding(oracleGrp1, tenancy.GroupOwner)
	unscoped.GroupID = oracleOrgOnly
	groupFacts = append(groupFacts, []tenancy.BindingFact{unscoped})
	for i, bindings := range groupFacts {
		for _, c := range oracleCapabilities {
			if c.name == "organization.read" || c.name == "group.create" { // G018: org_member's own creation path
				continue
			}
			name := fmt.Sprintf("group-binding/%d/%q/%q/%s", i, bindings[0].Role, bindings[0].Source, c.name)
			// G019 (Plan r1 §1.3): the fixture keeps exactly one valid org_member binding in A, whose own
			// organization.leave is allowed; the unknown or inconsistent Group fact neither denies nor widens it. Every
			// other capability denies.
			want := c.name == "organization.leave"
			t.Run(name, func(t *testing.T) {
				f := oracleSubjectFacts(t, string(tenancy.OrgMember), oracleFlagOn)
				f.GroupBindings = bindings
				oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), oracleNaturalTarget(t, c.name), c.name, want)
			})
		}
	}
	for _, kind := range oracleUnknownKinds {
		for _, c := range oracleCapabilities {
			name := fmt.Sprintf("target-kind/%q/%s", kind, c.name)
			t.Run(name, func(t *testing.T) {
				f := oracleSubjectFacts(t, "org_owner", oracleFlagOn)
				tg := Target{Kind: kind, Organization: oracleOrgA, Group: oracleGrp1, Profile: oraclePrf1}
				oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), tg, c.name, oracleDeny)
			})
		}
	}
}

// TestOracleCrossOrganization: role grants never cross Organizations, and rows of another Organization never authorize
// in the active one (Spec R1 to R3).
func TestOracleCrossOrganization(t *testing.T) {
	run := func(label string, f Facts, tg Target, capability string) {
		name := label + "/" + capability
		t.Run(name, func(t *testing.T) {
			oracleExpectShape(t, "complete", f, oracleSelOf(oracleOrgOnly), tg, capability, oracleDeny)
		})
	}
	ownerInB := oracleBaseFacts(oracleFlagOn)
	ownerInB.OrganizationBindings = []tenancy.BindingFact{{OrganizationID: oracleOrgB, Role: string(tenancy.OrgOwner), Source: tenancy.SourceMembership}}
	memberAndOwnerInB := oracleSubjectFacts(t, string(tenancy.OrgMember), oracleFlagOn)
	memberAndOwnerInB.OrganizationBindings = append(memberAndOwnerInB.OrganizationBindings, ownerInB.OrganizationBindings[0])
	groupOwnerInB := oracleSubjectFacts(t, string(tenancy.OrgMember), oracleFlagOn)
	groupOwnerInB.GroupBindings = []tenancy.BindingFact{{OrganizationID: oracleOrgB, GroupID: oracleGrp1, Role: string(tenancy.GroupOwner), Source: tenancy.SourceMembership}}
	for _, c := range oracleCapabilities {
		target := oracleNaturalTarget(t, c.name)
		run("org-binding-only-in-B", ownerInB, target, c.name)
		// org_member's own organization.read path in A is not asserted next to a foreign binding (NOTES.md, A2), nor its
		// G018 group.create path.
		if c.name != "organization.read" && c.name != "group.create" {
			run("org-member-in-A-and-owner-in-B", memberAndOwnerInB, target, c.name)
			if c.name != "organization.leave" {
				run("group-binding-claiming-B", groupOwnerInB, target, c.name)
				continue
			}
			// G019 (Plan r1 §1.3): this fixture keeps exactly one valid org_member binding in A, so organization.leave is
			// that member's own departure from A and allowed; the claimed B Group binding grants nothing. The
			// org-binding-only-in-B and org-member-in-A-and-owner-in-B fixtures above stay DENY for leave.
			t.Run("group-binding-claiming-B/"+c.name, func(t *testing.T) {
				oracleExpectShape(t, "complete", groupOwnerInB, oracleSelOf(oracleOrgOnly), target, c.name, oracleAllow)
			})
		}
	}
	groupRowInB := oracleSubjectFacts(t, "org_owner", oracleFlagOn)
	groupRowInB.Groups[0] = oracleGroupRow(oracleOrgB, oracleGrp1, "Group 1", "group-1")
	profileRowInB := oracleSubjectFacts(t, "org_owner", oracleFlagOn)
	profileRowInB.Profiles[0] = oracleProfileRow(oracleOrgB, oraclePrf1, "Profile 1")
	foreignRows := oracleSubjectFacts(t, "org_owner", oracleFlagOn)
	foreignRows.Groups = append(foreignRows.Groups, oracleGroupRow(oracleOrgB, oracleGrpB, "Group B", "group-b"))
	foreignRows.Profiles = append(foreignRows.Profiles, oracleProfileRow(oracleOrgB, oraclePrfB, "Profile B"))
	foreignRows.Associations = append(foreignRows.Associations, oracleAssociation(oracleGrpB, oraclePrfB))
	foreignLink := oracleSubjectFacts(t, string(tenancy.GroupOwner), oracleFlagOn)
	foreignLink.Profiles = append(foreignLink.Profiles, oracleProfileRow(oracleOrgB, oraclePrfB, "Profile B"))
	foreignLink.Associations = append(foreignLink.Associations, oracleAssociation(oracleGrp1, oraclePrfB))
	for _, c := range oracleCapabilities {
		switch c.kind {
		case "group":
			run("target-group-row-in-B", groupRowInB, oracleG(oracleGrp1), c.name)
			run("foreign-group-row", foreignRows, oracleG(oracleGrpB), c.name)
		case "business_profile":
			run("target-profile-row-in-B", profileRowInB, oracleP(oraclePrf1), c.name)
			run("foreign-profile-row", foreignRows, oracleP(oraclePrfB), c.name)
			run("group-link-to-profile-in-B", foreignLink, oracleP(oraclePrfB), c.name)
		}
	}
}

// U-TOOL-ELIGIBILITY (G015 Plan r4 §6.2, AC2): the discovery rule eligible and its fact options factQueryFor. For
// every canonical capability, subject, context (Organization-only, selected g1, selected g3), flag value and link state
// (P1 linked to g1, or that link removed), eligible on the facts factQueryFor would load must equal "some target of the
// oracle world is allowed by decide on the complete facts" (decide itself is held to the literal oracle above).
// Literal cases pin the Organization owner/admin paths without Group bindings, unlinked profiles, the linked+flag path
// and link removal; nothing else evaluates visibility.
func TestToolEligibility(t *testing.T) {
	subjects := append(slices.Clone(oracleColumns), "g1owner+g2owner", "g1owner+g3admin", "g1member+g2admin", "g3admin",
		"g4owner", "org_admin+g3member", "g1owner-no-org", "no-bindings",
		// G060: the Organization owner/admin subjects with the Group-management override enabled.
		"org_owner+ovr", "org_admin+ovr", "org_admin+g3member+ovr",
		// G020: the dual-role subject (Grace), with the override off and on.
		"org_admin+g1admin", "org_admin+g1admin+ovr")
	targets := map[TargetKind][]Target{
		kindOrganization: {oracleO(oracleOrgA)},
		kindGroup:        {oracleG(oracleGrp1), oracleG(oracleGrp2), oracleG(oracleGrp3), oracleG(oracleGrp4)},
		kindProfile:      {oracleP(oraclePrf1), oracleP(oraclePrf2), oracleP(oraclePrf3)},
	}
	facts := func(subject string, flagOn, unlinked bool) Facts {
		f := oracleSubjectFacts(t, subject, flagOn)
		if unlinked {
			f.Associations = slices.DeleteFunc(slices.Clone(f.Associations), func(a tenancy.AssociationFact) bool {
				return a.GroupID == oracleGrp1
			})
		}
		return f
	}
	n := 0
	for _, c := range oracleCapabilities {
		req, err := RequireCapability(c.name)
		if err != nil {
			t.Fatalf("RequireCapability(%q): %v", c.name, err)
		}
		reqs := map[string]ToolRequirement{"test.tool": req}
		for _, subject := range subjects {
			for _, g := range []tenancy.GroupID{oracleOrgOnly, oracleGrp1, oracleGrp3} {
				for _, flagOn := range []bool{oracleFlagOn, oracleFlagOff} {
					for _, unlinked := range []bool{false, true} {
						n++
						name := fmt.Sprintf("%s/%s/sel=%v/flag=%t/unlinked=%t", c.name, subject, g, flagOn, unlinked)
						complete := facts(subject, flagOn, unlinked)
						sel := oracleSelOf(g)
						loaded := eligibilityLoaded(complete, factQueryFor(reqs, sel))
						want := false
						if contextCheck(sel, complete) == "" {
							for _, tg := range targets[c.kind] {
								if ok, _ := decide(Capability(c.name), sel, complete, tg); ok {
									want = true
								}
							}
						}
						if got := eligible(req, sel, loaded); got != want {
							t.Errorf("%s: eligible on the loaded facts = %t, want %t", name, got, want)
						}
						if got := eligible(req, sel, complete); got != want {
							t.Errorf("%s: eligible on the complete facts = %t, want %t", name, got, want)
						}
					}
				}
			}
		}
	}
	if n != len(oracleCapabilities)*len(subjects)*3*2*2 { // G020: twenty-two capabilities
		t.Fatalf("%d eligibility cases", n)
	}

	// Literal pins (r3 item 1; G060): an Organization admin with no Group binding reads any Group and any profile, but
	// administers them (any Group and any profile, including the unlinked P3, in Organization-only context and,
	// restricted to the selected Group, in selected context) only while the Organization's override is enabled; a Group
	// admin of g1 holds no Organization path, so its profile capability needs g1's link and (for update) the flag.
	for _, p := range []struct {
		capability, subject string
		sel                 tenancy.GroupID
		flagOn, unlinked    bool
		want                bool
	}{
		{"group.update", "org_admin", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"group.update", "org_admin", oracleGrp1, oracleFlagOn, false, oracleDeny},
		{"business_profile.profile.update", "org_admin", oracleOrgOnly, oracleFlagOff, true, oracleDeny},
		{"business_profile.profile.update", "org_admin", oracleGrp1, oracleFlagOff, true, oracleDeny},
		{"group.read", "org_admin", oracleOrgOnly, oracleFlagOn, false, oracleAllow},
		{"business_profile.profile.read", "org_admin", oracleGrp1, oracleFlagOff, true, oracleAllow},
		{"group.update", "org_admin+ovr", oracleOrgOnly, oracleFlagOn, false, oracleAllow},
		{"group.update", "org_admin+ovr", oracleGrp1, oracleFlagOn, false, oracleAllow},
		{"group.manage_members", "org_owner+ovr", oracleGrp1, oracleFlagOn, false, oracleAllow},
		{"business_profile.profile.update", "org_admin+ovr", oracleOrgOnly, oracleFlagOff, true, oracleAllow},
		{"business_profile.profile.update", "org_admin+ovr", oracleGrp1, oracleFlagOff, true, oracleAllow},
		{"business_profile.privileged_settings.manage", "org_admin+ovr", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"group.manage_owners", "org_owner+ovr", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"business_profile.privileged_settings.manage", "org_admin", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"business_profile.privileged_settings.manage", "org_owner", oracleGrp1, oracleFlagOn, false, oracleAllow},
		{"organization.manage_policy", "org_admin", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"organization.manage_policy", "org_owner", oracleGrp1, oracleFlagOn, false, oracleAllow},
		{"audit.read", "org_owner", oracleOrgOnly, oracleFlagOn, false, oracleAllow},
		{"audit.read", "group_owner", oracleGrp1, oracleFlagOn, false, oracleDeny},
		{"organization.manage_branding", "org_admin", oracleOrgOnly, oracleFlagOn, false, oracleAllow},
		{"organization.manage_branding", "group_owner", oracleGrp1, oracleFlagOn, false, oracleDeny},
		{"group.update", "group_admin", oracleOrgOnly, oracleFlagOn, false, oracleAllow},
		{"group.update", "group_member", oracleGrp1, oracleFlagOn, false, oracleDeny},
		{"business_profile.profile.update", "group_admin", oracleGrp1, oracleFlagOn, false, oracleAllow},
		{"business_profile.profile.update", "group_admin", oracleGrp1, oracleFlagOff, false, oracleDeny},
		{"business_profile.profile.update", "group_admin", oracleGrp1, oracleFlagOn, true, oracleDeny},
		{"business_profile.profile.update", "group_admin", oracleOrgOnly, oracleFlagOn, true, oracleDeny},
		{"business_profile.profile.read", "group_member", oracleOrgOnly, oracleFlagOff, false, oracleAllow},
		{"business_profile.select_for_onboarding", "g1owner+g3admin", oracleGrp3, oracleFlagOn, true, oracleAllow},
		{"business_profile.select_for_onboarding", "g4owner", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		// G020: label.organization.manage is the Organization owner/admin path in either context; the Group label
		// capabilities follow decide's explicit Group path only (the override is irrelevant).
		{"label.organization.manage", "org_admin", oracleOrgOnly, oracleFlagOn, false, oracleAllow},
		{"label.organization.manage", "org_admin+ovr", oracleGrp1, oracleFlagOn, false, oracleAllow},
		{"label.organization.manage", "group_owner", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"label.organization.manage", "org_member", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"label.group.manage", "org_admin", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"label.group.manage", "org_admin+ovr", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
		{"label.group.manage", "org_owner+ovr", oracleGrp1, oracleFlagOn, false, oracleDeny},
		{"label.group.read", "org_owner+ovr", oracleOrgOnly, oracleFlagOff, false, oracleDeny},
		{"label.group.manage", "group_admin", oracleOrgOnly, oracleFlagOn, false, oracleAllow},
		{"label.group.manage", "group_member", oracleGrp1, oracleFlagOn, false, oracleDeny},
		{"label.group.read", "group_member", oracleGrp1, oracleFlagOn, false, oracleAllow},
		{"label.group.read", "org_admin+g3member", oracleGrp3, oracleFlagOn, false, oracleAllow},
		{"label.group.read", "org_admin+g3member", oracleGrp1, oracleFlagOn, false, oracleDeny},
		{"label.group.manage", "org_admin+g1admin+ovr", oracleGrp1, oracleFlagOn, false, oracleAllow},
		{"label.group.manage", "org_admin+g1admin+ovr", oracleGrp3, oracleFlagOn, false, oracleDeny},
		{"label.group.read", "g1owner-no-org", oracleOrgOnly, oracleFlagOn, false, oracleDeny},
	} {
		req, _ := RequireCapability(p.capability)
		complete := facts(p.subject, p.flagOn, p.unlinked)
		sel := oracleSelOf(p.sel)
		if got := eligible(req, sel, eligibilityLoaded(complete, factQueryFor(map[string]ToolRequirement{"t": req}, sel))); got != p.want {
			t.Errorf("literal %+v: eligible = %t", p, got)
		}
	}

	// Literal empty-Organization pins (Duo D1; Spec supervisor clarification): with no Group, association or profile
	// row, the Organization owner/admin path stays a role-level hint that asserts no target exists; privileged settings
	// stay owner-only; a non-admin has no Group-derived path; a failed context check (no membership, or a selected Group
	// that does not exist) lists nothing.
	empty := func(subject string) Facts {
		f := oracleSubjectFacts(t, subject, oracleFlagOn)
		f.GroupBindings, f.Groups, f.Associations, f.Profiles = nil, nil, nil, nil
		return f
	}
	for _, p := range []struct {
		capability, subject string
		sel                 tenancy.GroupID
		want                bool
	}{
		{"group.update", "org_owner", oracleOrgOnly, oracleDeny},
		{"business_profile.profile.update", "org_owner", oracleOrgOnly, oracleDeny},
		{"group.update", "org_owner+ovr", oracleOrgOnly, oracleAllow},
		{"business_profile.profile.update", "org_owner+ovr", oracleOrgOnly, oracleAllow},
		{"group.read", "org_owner", oracleOrgOnly, oracleAllow},
		{"business_profile.privileged_settings.manage", "org_owner", oracleOrgOnly, oracleAllow},
		{"group.update", "org_admin", oracleOrgOnly, oracleDeny},
		{"group.update", "org_admin+ovr", oracleOrgOnly, oracleAllow},
		{"business_profile.profile.read", "org_admin", oracleOrgOnly, oracleAllow},
		{"business_profile.privileged_settings.manage", "org_admin", oracleOrgOnly, oracleDeny},
		{"group.read", "org_member", oracleOrgOnly, oracleDeny},
		{"business_profile.profile.read", "org_member", oracleOrgOnly, oracleDeny},
		{"business_profile.select_for_onboarding", "org_member", oracleOrgOnly, oracleDeny},
		{"group.update", "org_owner", oracleGrp1, oracleDeny},
		{"business_profile.profile.update", "org_admin", oracleGrp1, oracleDeny},
		{"group.update", "no-bindings", oracleOrgOnly, oracleDeny},
		{"business_profile.profile.read", "no-bindings", oracleOrgOnly, oracleDeny},
		// G020: the Organization label path is role-level; no Group label path exists without a Group role.
		{"label.organization.manage", "org_owner", oracleOrgOnly, oracleAllow},
		{"label.organization.manage", "org_admin+ovr", oracleOrgOnly, oracleAllow},
		{"label.group.manage", "org_owner+ovr", oracleOrgOnly, oracleDeny},
		{"label.group.read", "org_admin+ovr", oracleOrgOnly, oracleDeny},
	} {
		req, _ := RequireCapability(p.capability)
		complete, sel := empty(p.subject), oracleSelOf(p.sel)
		if got := eligible(req, sel, eligibilityLoaded(complete, factQueryFor(map[string]ToolRequirement{"t": req}, sel))); got != p.want {
			t.Errorf("empty Organization %+v: eligible = %t", p, got)
		}
	}

	// The control requirements hold exactly when the context holds; the zero requirement never does.
	member, lost := oracleSubjectFacts(t, "group_member", oracleFlagOn), oracleSubjectFacts(t, "g1owner-no-org", oracleFlagOn)
	for _, r := range []ToolRequirement{RequireBinding(), RequireMembership(), RequireContext()} {
		if !eligible(r, oracleSelOf(oracleGrp1), member) || eligible(r, oracleSelOf(oracleGrp3), member) ||
			eligible(r, oracleSelOf(oracleOrgOnly), lost) {
			t.Errorf("control requirement %+v: wrong context dependence", r)
		}
	}
	if eligible(ToolRequirement{}, oracleSelOf(oracleOrgOnly), oracleSubjectFacts(t, "org_owner", oracleFlagOn)) {
		t.Error("the zero requirement is eligible")
	}

	// The capability names are closed: the twenty-two canonical values only (G019 made group.manage_owners canonical; G020
	// adds the three label capabilities).
	for _, u := range oracleUnknownCapabilities {
		if r, err := RequireCapability(u); err == nil || CodeOf(err) != CodeInvalidInput || r != (ToolRequirement{}) {
			t.Errorf("RequireCapability(%q) = %+v, %v; want invalid_input", u, r, err)
		}
	}

	// The fact options depend on the catalog's shape only (Plan r4 §1, §5.5).
	org, _ := RequireCapability("audit.read")
	grp, _ := RequireCapability("group.update")
	prf, _ := RequireCapability("business_profile.profile.update")
	for _, q := range []struct {
		name        string
		reqs        map[string]ToolRequirement
		sel         tenancy.GroupID
		groups      []tenancy.GroupID
		allProfiles bool
	}{
		{"none, Organization-only", nil, oracleOrgOnly, []tenancy.GroupID{}, false},
		{"none, selected", nil, oracleGrp1, []tenancy.GroupID{oracleGrp1}, false},
		{"controls only", map[string]ToolRequirement{"a": RequireBinding(), "b": RequireContext()}, oracleOrgOnly, []tenancy.GroupID{}, false},
		{"Organization kind", map[string]ToolRequirement{"a": org}, oracleGrp1, []tenancy.GroupID{oracleGrp1}, false},
		{"Group kind, Organization-only", map[string]ToolRequirement{"a": org, "b": grp}, oracleOrgOnly, nil, true},
		{"profile kind, selected", map[string]ToolRequirement{"a": prf}, oracleGrp3, []tenancy.GroupID{oracleGrp3}, true},
	} {
		got := factQueryFor(q.reqs, oracleSelOf(q.sel))
		if got.Organization != oracleOrgA || (got.Groups == nil) != (q.groups == nil) || !slices.Equal(got.Groups, q.groups) ||
			got.AllProfiles != q.allProfiles || got.Lock || !zero(got.TargetGroup) || !zero(got.Profile) || !zero(got.User) {
			t.Errorf("factQueryFor %s = %+v", q.name, got)
		}
	}
}

// eligibilityLoaded returns the subset of the complete facts f that tenancy.ReadFacts reads for q (facts.go): the
// Organization row, policy and bindings; the subject's Group bindings of the candidate Groups (every bound Group when
// Groups is nil); those Group rows; their associations and every profile when AllProfiles is set.
func eligibilityLoaded(f Facts, q tenancy.FactQuery) Facts {
	candidates := map[tenancy.GroupID]bool{}
	if q.Groups == nil {
		for _, b := range f.GroupBindings {
			candidates[b.GroupID] = true
		}
	}
	for _, g := range q.Groups {
		candidates[g] = true
	}
	out := Facts{Organization: f.Organization, Policy: f.Policy, OrganizationBindings: f.OrganizationBindings}
	for _, b := range f.GroupBindings {
		if candidates[b.GroupID] {
			out.GroupBindings = append(out.GroupBindings, b)
		}
	}
	for _, g := range f.Groups {
		if candidates[g.ID] {
			out.Groups = append(out.Groups, g)
		}
	}
	if q.AllProfiles {
		if len(out.Groups) > 0 {
			for _, a := range f.Associations {
				if candidates[a.GroupID] {
					out.Associations = append(out.Associations, a)
				}
			}
		}
		out.Profiles = f.Profiles
	}
	return out
}
