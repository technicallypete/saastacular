//go:build integration

// G008 independent RLS oracle: hand-written expected outcomes. Revision r2.
//
// SOURCES (the only files read; SHA-256 verified before writing):
//   G008-postgresql-rls-spec.md                  d3cb3c51f69bdf5475204aea6597a3de00e37b057df77b01241200ae8cedac6a
//   ADR-004-group-business-profile-cardinality.md 51a487844d74e913d195c963095c2a381deeb293d1829085f2f026f060955a68
//   AUTHORIZATION.md                             0363d9d177e06395e3f1db528081791cbeeb6a928d2b48c27eb5627bda884f22
//   schema-inventory.md                          be017c7d0affddf6a7bc5e123b06617ecae4ffda7e054663b99f97cfa14513c1
//   interface-r1.md                              b26e64c413eb3fca64f729fcfdc85334772f3021224ba8598d83a0e98ebcfa8c
//   G008-context-validity-clarification.md       4df154ebf6297926aba9a51d5a1f1f8a66067568b00b32677b2e244889898eee
//   (r0 used interface.md 3d11a60aaaada329d8b1dd9237bd2cb5a2fc8c44708d1178cb4a4f127cc70fe3, kept as history.)
// Authority: G008-execution-amendment-01.md (ccd0382440ae9d2d688b637f7cf6016b84122b336f69314093be806f8ddc66af), C2.
//
// REVISION r1 (2026-10-13). Reason: supervisor clarification relayed by the coordinator. The Spec requires an
// unauthorized UPDATE to cause no mutation and allows either filtering or boundary rejection; it does not pin that
// mechanism from the actor's need for row locks. r0's reading A1 chose "zero_rows" versus "denied" from that lock
// argument, which is Plan-owned mechanics, not independently authored authorization. Change: every r0 "zero_rows"
// and "denied" UPDATE expectation is now the semantic "no_update" of interface-r1 (no input explicitly pins an SQLSTATE
// or error-versus-zero form for any listed action), and the explanatory text (header A1, A6, conventions, Basis wording,
// one case name) no longer asserts a mechanism. Unchanged: fixture rows, contexts, exact visible sets, Reject classes,
// which actions are allowed ("ok") or forbidden, and paired positives. r0 is archived unchanged at
// archive/rls_expected_test.r0-8523cc47.go (8523cc47feca1a375264dd105a639220991d56378c60262b965f9513cecd0355).
// No outcome was supplied from runtime results.
//
// REVISION r2 (2026-10-13). Reason: the published G008 Spec clarification 'Context validity and global self identity'
// (G008-context-validity-clarification.md, 4df154ebf6297926aba9a51d5a1f1f8a66067568b00b32677b2e244889898eee). It
// pins that tenant-selection validity precedes capability checks: an unknown, foreign or unselectable selection is not
// a valid TenantContext and denies every tenant-owned row and projection, including Organization rows and the subject's
// own Organization membership/binding rows, while a structurally valid shape whose selectors fail data-dependent
// resolution still shows the established actor's global self row; structurally malformed, zero, missing, partial and
// illegal shapes keep their rejection. Changes: reading A5 and the SHAPES/derivation text are rewritten; case
// 'adminA2 with unselectable A1 selected' loses its Organization row and own Organization binding (self row only);
// case 'crossUser A with B1' moves from a rejected shape to a structurally valid shape with the self row only (Shape
// "valid", Reject ""); the Basis of the three foreign/non-member selection cases now cites the clarification (their
// visible sets were already self-only). Unchanged: every fixture row, every UPDATE expectation, and all other cases.
// Archives (unchanged): r1 at archive/rls_expected_test.r1-c968cfa5.go
// (c968cfa56a77df50db74b6d480751e5cb3ac2cdfdb04c30ca416c76d4b1260a7); r0 at archive/rls_expected_test.r0-8523cc47.go
// (8523cc47feca1a375264dd105a639220991d56378c60262b965f9513cecd0355). No outcome was supplied from runtime results.
//
// INDEPENDENCE. Every fixture row and every expected value below is a hand-written literal derived by reading the
// sources above. Nothing is computed from an evaluator, the schema, a migration, a helper or any implementation, none
// of which the author read. When the real database disagrees with an expectation, the disagreement is resolved against
// the Spec (and ADR-004/AUTHORIZATION.md) or escalated to Codex. It is never resolved by editing these expectations.
//
// CONVENTIONS.
//   Visible sets are order-insensitive exact sets. A nil slice inside a non-nil rlsVisible means the exact set is empty.
//   View entries: OrgMemberships/GroupMemberships name the role_bindings row ("<bindingLabel>");
//   ProfileAccessPaths name "<associationLabel>/<bindingLabel>" (schema-inventory view output columns).
//   Each rlsUpdate is an independent attempt from the unmodified fixture state (for example its own rolled-back
//   transaction). The implementer chooses the new value; it must differ from the current value and, for structural
//   columns, should otherwise satisfy constraints where possible (for example a same-Organization target) so that the
//   outcome is attributable to the isolation contract.
//   Expect values (interface-r1): "ok" (the authorized change is applied to that row) or "no_update" (the attempt must
//   not change that row, nor any other row). The no_update mechanism (filtering to zero rows or rejection at the
//   boundary) is Plan-owned and checked separately; no input explicitly pins it for any action listed here.
//   Reject "rejected-no-rows-no-change": every read of every covered relation (including the actor's own users row)
//   returns no row or fails, and no write changes any row. Error versus empty is not pinned (Spec R3: errors are not
//   allow fallbacks). Reject cases carry Visible == nil and no Updates.
//
// SHAPES. "valid": each non-empty label is set to that row's canonical UUID; an empty Org/Group is absent and the
// optional Group is explicitly cleared (Spec R3). Invalid shapes keep the other labelled fields of a paired valid case
// and replace exactly the named field (its label is then ""):
//   "no-context"         no context setting is initialized in the transaction at all.
//   "missing-actor"      the actor setting is never initialized; Org/Group from labels (partial initialization).
//   "empty-actor"        the actor setting is initialized to the empty string.
//   "zero-actor"         the actor setting is 00000000-0000-0000-0000-000000000000.
//   "malformed-actor"    the actor setting is non-UUID text (for example "not-a-uuid").
//   "zero-org"           the Organization setting is the all-zero UUID; Group cleared.
//   "malformed-org"      the Organization setting is non-UUID text; Group cleared.
//   "group-without-org"  the Group setting is the labelled Group's UUID; the Organization setting is absent or empty.
//   "zero-group"         the Group setting is the all-zero UUID; actor and Organization from labels.
//   "malformed-group"    the Group setting is non-UUID text; actor and Organization from labels.
// A structurally valid context whose selectors fail data-dependent resolution (a foreign or unknown Organization, a
// Group of another Organization, or an unselectable Group) uses Shape "valid" and is not a Reject case: it shows only
// the actor's global self row (r2, see A5). r1's "group-of-other-org" shape is retired for that reason.
//
// DERIVATION RULES (each case's Basis cites the specific sections):
//   Context validity first (r2 clarification): the selected Organization requires the actor's current Organization
//     membership, and a selected Group must belong to it and be selectable (org_owner/org_admin, or a binding in that
//     exact Group). Otherwise every tenant relation and projection below is empty and Users is the actor alone; the
//     rules that follow apply only to valid contexts.
//   Organizations: the active Organization only, when the actor holds an Organization binding there
//     (organization.read for org_owner/org_admin/org_member; not context-gated, gate 3 applies to group.* only).
//   Groups: org_owner/org_admin see every Group of the active Organization; Group roles see their own Groups; a selected
//     Group restricts both to that Group (matrix group.read, interpretation pin (b), gate 3, Spec R4).
//   Profiles: org_owner/org_admin see every profile of the active Organization including unlinked ones, independent of
//     the selected Group; Group owner/admin/member see the profile linked to a candidate Group (selected Group, else all
//     their Groups) (matrix business_profile.profile.read, 'linked', ADR-004).
//   Associations: visible only when both its Group and its profile are visible (Spec R4).
//   RoleBindings: the actor's own bindings within the context (see A2); plus, for org_owner/org_admin
//     (organization.manage_members), every Organization-scope binding, and (group.manage_members 'all') every
//     Group-scope binding of the in-context Groups; plus, for group_owner/group_admin (group.manage_members 'own'),
//     every binding of their in-context Groups (Spec R5).
//   Membership views and access paths: security_invoker projections, so exactly the rows of visible base rows (Spec R5,
//     AC7, schema-inventory); an access path requires both its association and its binding to be visible.
//   Users: the actor (self) plus the subjects of the RoleBindings visible in the same context (Spec R6, see A4).
//   Updates: "ok" when the G007 capability allows (group.update for name/slug; business_profile.profile.update for
//     profile context columns, linked+flag for Group roles; organization.manage_policy for the flag); otherwise
//     "no_update", as are all identity, parent, scope, role, source, subject, status and link-retarget changes (Spec R4,
//     AC6).
//
// AMBIGUITIES AND CHOSEN READINGS:
//   A1 (revised in r1) Mechanism of a refused UPDATE. The Spec leaves it to the Plan (Permitted Deviations: 'A denied
//      command may use explicit privilege denial or default-deny RLS as pinned'; R1; R10 'not every RLS denial is an SQL
//      error'). This oracle therefore asserts only the semantic outcome "no_update" for every forbidden attempt, whether
//      or not the actor can see the row. Authorization context kept as Basis, not as a mechanism claim: the actor's
//      users row, the active Organization row, the actor's own in-context bindings and candidate Groups are rows the
//      actor's own guarded writes lock (AUTHORIZATION Ordering steps 1, 3-6; G008 handoff 'Row locks under row-level
//      security'); Spec R8 and AC6 require that any privilege needed for those locks does not allow the forbidden
//      change. Whether such an attempt errors or affects zero rows is checked against the Plan, not here.
//   A2 Own bindings in selected-Group context: the actor's Organization binding(s) plus the binding(s) in the selected
//      Group; own bindings in other Groups are excluded (Spec R5 'own current bindings in the active Organization within
//      the selected scope'; the Organization binding is locked at Ordering step 4 in every write and holds gate 1).
//   A3 org_owner/org_admin in selected-Group context still see every Organization-scope binding
//      (organization.manage_members is an Organization capability; gate 3 restricts only group.* targets) but only the
//      selected Group's Group-scope bindings (pin (b)).
//   A4 'Identities sharing a scope the caller is permitted to inspect' (Spec R6) is read as the subjects of the bindings
//      the caller may inspect under Spec R5. Identity-only context (no active Organization) admits self only (no
//      directory from mere global login).
//   A5 (revised in r2; no longer an ambiguity, pinned by the context-validity clarification) A structurally valid
//      context whose tenant selectors fail data-dependent resolution (an Organization where the actor has no membership,
//      a same-Organization Group the actor cannot select, or a Group of another Organization) is not a valid
//      TenantContext: it denies every tenant-owned row and projection, including the Organization row and the actor's
//      own Organization binding, and no capability such as organization.read overrides it (clarification section
//      'Tenant-selection validity precedes capability checks'). It is not a rejected shape either: the established actor's
//      global self row stays visible, and no co-member identity is shown (clarification 'Structural rejection versus
//      global self liveness'; Spec R6 'admit self'). Structurally malformed, zero, missing, partial and illegal shapes
//      keep the Reject class. r1 read the unselectable-Group case as filtering, which kept the Organization row and own
//      Organization binding, and read the other-Organization Group as a rejected shape. The clarification supersedes
//      both readings.
//   A6 (revised in r1) UPDATE on group_business_profiles is not a permitted request command (no guarded operation
//      updates links; Spec R4 'retarget links'), so it is "no_update" for every actor whether or not the row is visible.
//      Spec R1 describes 'documented default-deny plus privilege denial where not permitted', but no input pins an exact
//      SQLSTATE, so no mechanism-specific form is used.
//   A7 rlsContext carries no session and all fixture users are "active": disabled-user, expired/revoked-session and
//      revocation-after-commit outcomes (AC3, AC8, AC10) are core/session-boundary and ordering tests outside this
//      static oracle. AC5's last-path removal is a state transition this interface cannot express; it is covered
//      statically by two Groups attached to one profile each yielding only its own path.
//   A8 Not pinned and therefore omitted: organizations.name (AUTHORIZATION mentions trusted Organization renames),
//      changes to another user's role or status, renames by pure group members, and structural updates of rows the
//      actor cannot see (except A6).
//   A9 The policy flag affects only inherited profile updates, never reads or selection (AUTHORIZATION Policy flag), so
//      Organization B (flag off) has the same read rules as Organization A (flag on).

package access

// --- fixture world (you choose its contents; the implementer creates exactly these rows) ---

type rlsOrg struct {
	Label                   string // e.g. "A", "B"
	GroupAdminsEditProfiles bool   // organizations.group_admins_edit_business_profiles
}

type rlsGroup struct {
	Label string // e.g. "A1"
	Org   string // rlsOrg.Label
}

type rlsUser struct {
	Label  string // e.g. "orgOwnerA"
	Status string // as the Spec allows (for example "active", "disabled")
}

type rlsBinding struct {
	Label     string // unique label for this role_bindings row
	Org       string // rlsOrg.Label
	User      string // rlsUser.Label
	ScopeType string // "organization" or "group"
	Group     string // rlsGroup.Label for group scope, "" for organization scope
	Role      string // a role named by AUTHORIZATION.md
	Source    string // "membership" unless the Spec gives another value you need
}

type rlsProfile struct {
	Label string // business profile label
	Org   string // rlsOrg.Label
}

type rlsAssociation struct {
	Label   string // group_business_profiles row label
	Org     string // rlsOrg.Label
	Group   string // rlsGroup.Label
	Profile string // rlsProfile.Label
}

// rlsLabelDef is one app.label_definitions row (G020): Organization-owned when Group is "", else owned by that Group.
type rlsLabelDef struct {
	Label      string // fixture label
	Org        string // rlsOrg.Label
	Group      string // "" = Organization-owned; otherwise the owning rlsGroup.Label
	Visibility string // "owner_scope" or "target_members" (Organization-owned rows are always "owner_scope")
}

// rlsLabelAssignment is one app.label_assignments row of the production shape (G020 Plan §1.2
// label_assignments_target_shape): an Organization-owned definition on a same-Organization Group (target_type 'group',
// target_id = target_group_id = Group).
type rlsLabelAssignment struct {
	Label string // fixture label
	Def   string // rlsLabelDef.Label (Organization-owned)
	Group string // target rlsGroup.Label
}

type rlsFixture struct {
	Orgs             []rlsOrg
	Groups           []rlsGroup
	Users            []rlsUser
	Bindings         []rlsBinding
	Profiles         []rlsProfile
	Associations     []rlsAssociation
	LabelDefinitions []rlsLabelDef        // G020
	LabelAssignments []rlsLabelAssignment // G020
}

// --- the request context a protected operation runs under ---

// rlsContext is the context a protected operation runs under. Shape is "valid" for a well-formed context, or one of
// the named invalid cases listed in the file header; the implementer maps each name to concrete setting values.
type rlsContext struct {
	Actor string // rlsUser.Label ("" means no authenticated user)
	Org   string // active Organization label ("" = none)
	Group string // selected Group label ("" = Organization-only, or none)
	Shape string // "valid", or a named invalid shape from the file header
	// G060: true = the Organization's group_management_override is set true by its owner inside the rolled-back
	// transaction before the case runs (the driver toggle); false = the default (off).
	Override bool
}

// --- expected outcomes ---

// rlsVisible holds visible row sets, as labels from the fixture. For views, an entry names the fixture row(s) it
// derives from, in the form "<bindingLabel>" (memberships) or "<associationLabel>/<bindingLabel>" (access paths).
// G020: LabelDefinitions and LabelAssignments name app.label_definitions / app.label_assignments fixture rows; as for
// every other field, nil means the exact set is empty (most cases see no label row and leave them unset).
type rlsVisible struct {
	RoleBindings       []string
	OrgMemberships     []string
	GroupMemberships   []string
	ProfileAccessPaths []string
	Users              []string
	Groups             []string
	Associations       []string
	Profiles           []string
	Organizations      []string
	LabelDefinitions   []string
	LabelAssignments   []string
}

// rlsUpdate is the outcome of a raw single-row UPDATE attempt against one fixture row (interface-r1 vocabulary): "ok"
// when the authorized change is applied, "no_update" when the attempt must not change the row by any mechanism.
type rlsUpdate struct {
	Table  string // "groups", "business_profiles", "organizations", "group_business_profiles", "role_bindings", "users"
	Row    string // fixture label
	Column string // the column the attempt changes
	Expect string // "ok" or "no_update"
	// Guard (r4, G019): for a "no_update" role change of role_bindings that version 15 admits through its UPDATE policy
	// but refuses in the owner guard, the named protection ("last_owner"): the driver must see exactly that named
	// constraint, never an unnamed error of the same class. "" keeps the r0-r3 policy mechanism (rlsMechanism).
	Guard string
}

type rlsCase struct {
	Name    string      // unique, descriptive
	Context rlsContext  // the request context
	Visible *rlsVisible // nil = this case asserts no visibility (invalid shapes); otherwise the exact sets
	Updates []rlsUpdate // raw single-row UPDATE attempts and their outcomes
	Reject  string      // for invalid shapes: "rejected-no-rows-no-change"; "" if valid
	Basis   string      // Spec/ADR/AUTHORIZATION section(s) that justify every expectation in this case
}

func rlsExpectedFixture() rlsFixture {
	return rlsFixture{
		Orgs: []rlsOrg{
			{Label: "A", GroupAdminsEditProfiles: true},
			{Label: "B", GroupAdminsEditProfiles: false},
		},
		Groups: []rlsGroup{
			{Label: "A1", Org: "A"},
			{Label: "A2", Org: "A"},
			{Label: "A3", Org: "A"},
			{Label: "A4", Org: "A"},
			{Label: "B1", Org: "B"},
			{Label: "B2", Org: "B"},
		},
		Users: []rlsUser{
			{Label: "ownerA", Status: "active"},
			{Label: "adminA", Status: "active"},
			{Label: "memberA", Status: "active"},
			{Label: "gOwnerA1", Status: "active"},
			{Label: "adminA2", Status: "active"},
			{Label: "memberA2", Status: "active"},
			{Label: "ownerA4", Status: "active"},
			{Label: "crossUser", Status: "active"},
			{Label: "ownerB", Status: "active"},
			{Label: "adminB", Status: "active"},
			{Label: "memberB1", Status: "active"},
			{Label: "outsider", Status: "active"},
		},
		Bindings: []rlsBinding{
			// Organization A.
			{Label: "bOwnerA", Org: "A", User: "ownerA", ScopeType: "organization", Group: "", Role: "org_owner", Source: "membership"},
			{Label: "bAdminA", Org: "A", User: "adminA", ScopeType: "organization", Group: "", Role: "org_admin", Source: "membership"},
			{Label: "bMemberA", Org: "A", User: "memberA", ScopeType: "organization", Group: "", Role: "org_member", Source: "membership"},
			{Label: "bGOwnerA1Org", Org: "A", User: "gOwnerA1", ScopeType: "organization", Group: "", Role: "org_member", Source: "membership"},
			{Label: "bGOwnerA1G1", Org: "A", User: "gOwnerA1", ScopeType: "group", Group: "A1", Role: "group_owner", Source: "membership"},
			{Label: "bGOwnerA1G3", Org: "A", User: "gOwnerA1", ScopeType: "group", Group: "A3", Role: "group_member", Source: "membership"},
			// G060 correction02 C02-2: the Organization admin adminA is an ordinary member of A1 (the raw own-Group-role
			// self-escalation rows with the override off and on).
			{Label: "bAdminAG1", Org: "A", User: "adminA", ScopeType: "group", Group: "A1", Role: "group_member", Source: "membership"},
			{Label: "bAdminA2Org", Org: "A", User: "adminA2", ScopeType: "organization", Group: "", Role: "org_member", Source: "membership"},
			{Label: "bAdminA2G2", Org: "A", User: "adminA2", ScopeType: "group", Group: "A2", Role: "group_admin", Source: "membership"},
			{Label: "bMemberA2Org", Org: "A", User: "memberA2", ScopeType: "organization", Group: "", Role: "org_member", Source: "membership"},
			{Label: "bMemberA2G2", Org: "A", User: "memberA2", ScopeType: "group", Group: "A2", Role: "group_member", Source: "membership"},
			{Label: "bOwnerA4Org", Org: "A", User: "ownerA4", ScopeType: "organization", Group: "", Role: "org_member", Source: "membership"},
			{Label: "bOwnerA4G4", Org: "A", User: "ownerA4", ScopeType: "group", Group: "A4", Role: "group_owner", Source: "membership"},
			{Label: "bCrossOrgA", Org: "A", User: "crossUser", ScopeType: "organization", Group: "", Role: "org_member", Source: "membership"},
			{Label: "bCrossA3", Org: "A", User: "crossUser", ScopeType: "group", Group: "A3", Role: "group_owner", Source: "membership"},
			// Organization B.
			{Label: "bOwnerB", Org: "B", User: "ownerB", ScopeType: "organization", Group: "", Role: "org_owner", Source: "membership"},
			{Label: "bAdminB", Org: "B", User: "adminB", ScopeType: "organization", Group: "", Role: "org_admin", Source: "membership"},
			{Label: "bCrossOrgB", Org: "B", User: "crossUser", ScopeType: "organization", Group: "", Role: "org_member", Source: "membership"},
			{Label: "bCrossB1", Org: "B", User: "crossUser", ScopeType: "group", Group: "B1", Role: "group_admin", Source: "membership"},
			{Label: "bMemberB1Org", Org: "B", User: "memberB1", ScopeType: "organization", Group: "", Role: "org_member", Source: "membership"},
			{Label: "bMemberB1G1", Org: "B", User: "memberB1", ScopeType: "group", Group: "B1", Role: "group_member", Source: "membership"},
			// outsider holds no binding anywhere.
		},
		Profiles: []rlsProfile{
			{Label: "pA12", Org: "A"},
			{Label: "pA3", Org: "A"},
			{Label: "pAUnlinked", Org: "A"},
			{Label: "pB1", Org: "B"},
			{Label: "pBUnlinked", Org: "B"},
		},
		Associations: []rlsAssociation{
			// A1 and A2 are sibling Groups sharing pA12; A3 has its own pA3; A4 and B2 have no profile (ADR-004).
			{Label: "gbpA1", Org: "A", Group: "A1", Profile: "pA12"},
			{Label: "gbpA2", Org: "A", Group: "A2", Profile: "pA12"},
			{Label: "gbpA3", Org: "A", Group: "A3", Profile: "pA3"},
			{Label: "gbpB1", Org: "B", Group: "B1", Profile: "pB1"},
		},
		// G020 (see "oracle G020" below): A3 plays the seed's research Group and A4 its design Group. lOrgA is assigned to
		// both; lOrgA2 is unassigned (the paired-positive and deletion targets). No Group-owned assignment exists: production
		// storage admits none (label_assignments_target_shape). Label rows are invisible to every pre-G020 relation.
		LabelDefinitions: []rlsLabelDef{
			{Label: "lOrgA", Org: "A", Group: "", Visibility: "owner_scope"},
			{Label: "lOrgA2", Org: "A", Group: "", Visibility: "owner_scope"},
			{Label: "lA3Own", Org: "A", Group: "A3", Visibility: "owner_scope"},
			{Label: "lA3Tm", Org: "A", Group: "A3", Visibility: "target_members"},
			{Label: "lA4Own", Org: "A", Group: "A4", Visibility: "owner_scope"},
			{Label: "lA4Tm", Org: "A", Group: "A4", Visibility: "target_members"},
			{Label: "lOrgB", Org: "B", Group: "", Visibility: "owner_scope"},
			{Label: "lB1Tm", Org: "B", Group: "B1", Visibility: "target_members"},
		},
		LabelAssignments: []rlsLabelAssignment{
			{Label: "asOrgAA3", Def: "lOrgA", Group: "A3"},
			{Label: "asOrgAA4", Def: "lOrgA", Group: "A4"},
			{Label: "asOrgBB1", Def: "lOrgB", Group: "B1"},
		},
	}
}

func rlsExpectedCases() []rlsCase {
	return []rlsCase{
		// ===== Organization A owner and admin =====
		{
			Name:    "ownerA org-only: every Organization A row; owner updates; foreign rows not updated; structural and self changes not applied",
			Context: rlsContext{Actor: "ownerA", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				// G060 (override off): Organization-scope bindings only; no Group roster row or access path.
				RoleBindings:       []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				OrgMemberships:     []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"ownerA", "adminA", "memberA", "gOwnerA1", "adminA2", "memberA2", "ownerA4", "crossUser"},
				Groups:             []string{"A1", "A2", "A3", "A4"},
				Associations:       []string{"gbpA1", "gbpA2", "gbpA3"},
				Profiles:           []string{"pA12", "pA3", "pAUnlinked"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lOrgA", "lOrgA2"},
				LabelAssignments:   []string{"asOrgAA3", "asOrgAA4"},
			},
			Updates: []rlsUpdate{
				// G060 (override off): no Group or Business Profile write through the Organization role (r0 A12/A13).
				{Table: "groups", Row: "A2", Column: "name", Expect: "no_update"},
				{Table: "groups", Row: "A4", Column: "slug", Expect: "no_update"},
				{Table: "business_profiles", Row: "pAUnlinked", Column: "display_name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "ok"},
				{Table: "groups", Row: "B1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pB1", Column: "display_name", Expect: "no_update"},
				{Table: "organizations", Row: "B", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "groups", Row: "A1", Column: "organization_id", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA3", Column: "organization_id", Expect: "no_update"},
				{Table: "group_business_profiles", Row: "gbpA1", Column: "group_id", Expect: "no_update"},
				{Table: "group_business_profiles", Row: "gbpA2", Column: "business_profile_id", Expect: "no_update"},
				{Table: "role_bindings", Row: "bOwnerA", Column: "role", Expect: "no_update", Guard: "last_owner"}, // r4: sole owner of A
				{Table: "role_bindings", Row: "bOwnerA", Column: "subject_user_id", Expect: "no_update"},
				{Table: "users", Row: "ownerA", Column: "status", Expect: "no_update"},
			},
			Basis: "G060 (ADR-008; Plan r0 A11-A13 as amended by S2 amendment01 A1.3/A1.5): without the Group-management override an Organization owner/admin without a Group role reads no Group roster (role_bindings Group rows, group_memberships, the joined access paths) and writes no Group or Business Profile row; Organization memberships, users, Group and profile metadata stay visible. AUTHORIZATION matrix org_owner: organization.read, organization.manage_members, organization.manage_policy, group.read/update 'all', business_profile.profile.read/update Y incl. unlinked; gates 1-2 (active Organization only). Spec R4 (only the active Organization's permitted rows; association needs Group and profile visibility; foreign UPDATE cannot mutate), R5 (membership visibility via organization.manage_members/group.manage_members; invoker projections), R6 (identities sharing an inspectable scope), R10 (not every RLS denial is an SQL error). Structural: AC6 and Spec R4 (reparent/retarget/subject/role changes must not apply), R1 (link UPDATE is not a permitted command, header A6). Self role/status: AUTHORIZATION Ordering steps 1 and 4 plus G008 handoff 'Row locks under row-level security' and Spec R8/AC6 (privileges needed for those locks must not become a mutation grant; header A1).",
		},
		{
			Name:    "ownerA selected A1: Group rows restricted to A1, profiles and Organization-scope bindings stay Organization-wide",
			Context: rlsContext{Actor: "ownerA", Org: "A", Group: "A1", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				OrgMemberships:     []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"ownerA", "adminA", "memberA", "gOwnerA1", "adminA2", "memberA2", "ownerA4", "crossUser"},
				Groups:             []string{"A1"},
				Associations:       []string{"gbpA1"},
				Profiles:           []string{"pA12", "pA3", "pAUnlinked"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lOrgA", "lOrgA2"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},                            // G060 (override off)
				{Table: "business_profiles", Row: "pA3", Column: "description", Expect: "no_update"},         // G060 (override off)
				{Table: "business_profiles", Row: "pAUnlinked", Column: "display_name", Expect: "no_update"}, // G060 (override off)
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "ok"},
				{Table: "groups", Row: "A2", Column: "name", Expect: "no_update"},
				{Table: "groups", Row: "A3", Column: "slug", Expect: "no_update"},
				{Table: "role_bindings", Row: "bOwnerA", Column: "role", Expect: "no_update", Guard: "last_owner"}, // r4: sole owner of A
			},
			Basis: "G060 (ADR-008; Plan r0 A11-A13 as amended by S2 amendment01 A1.3/A1.5): without the Group-management override an Organization owner/admin without a Group role reads no Group roster (role_bindings Group rows, group_memberships, the joined access paths) and writes no Group or Business Profile row; Organization memberships, users, Group and profile metadata stay visible. AUTHORIZATION interpretation pin (b) (org owner 'all' Group path restricted to the selected Group), gate 3 (context gate applies to group.* only, so organization.read/manage_members/manage_policy remain), matrix note 'Organization owner and admin profile paths are independent of the selected Group and include unlinked profiles'. Spec R4 (selected-Group restriction; association needs Group and profile visibility), R5 (header A3), R6 (header A4). Self role: Ordering step 4 + G008 handoff row locks + Spec R8/AC6 (header A1).",
		},
		{
			Name:    "ownerA identity-only: self only, no default Organization and no tenant row",
			Context: rlsContext{Actor: "ownerA", Org: "", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       nil,
				OrgMemberships:     nil,
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"ownerA"},
				Groups:             nil,
				Associations:       nil,
				Profiles:           nil,
				Organizations:      nil,
			},
			Updates: []rlsUpdate{
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
			},
			Basis: "Spec R3 (no default Organization, no tenant-zero/global sentinel), R6 (pre-tenant account self-read preserved; ordinary reads admit self; no directory access from mere global login), R10, AC8. AUTHORIZATION gate 1 (no active Organization binding, no tenant capability). Paired positives: 'ownerA org-only' updates the same rows.",
		},
		{
			Name:    "adminA org-only: every Organization A row; no policy update; admin updates profiles regardless of flag",
			Context: rlsContext{Actor: "adminA", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				// G060 (override off): Organization-scope bindings and the actor's own A1 member binding (own-row branch).
				RoleBindings:       []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA", "bAdminAG1"},
				OrgMemberships:     []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				GroupMemberships:   []string{"bAdminAG1"},
				ProfileAccessPaths: []string{"gbpA1/bAdminAG1"},
				Users:              []string{"ownerA", "adminA", "memberA", "gOwnerA1", "adminA2", "memberA2", "ownerA4", "crossUser"},
				Groups:             []string{"A1", "A2", "A3", "A4"},
				Associations:       []string{"gbpA1", "gbpA2", "gbpA3"},
				Profiles:           []string{"pA12", "pA3", "pAUnlinked"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lOrgA", "lOrgA2"},
				LabelAssignments:   []string{"asOrgAA3", "asOrgAA4"},
			},
			Updates: []rlsUpdate{
				// G060 (override off): no Group or profile write; A1 membership is member-level (no update, no flag path).
				{Table: "groups", Row: "A3", Column: "name", Expect: "no_update"},
				{Table: "groups", Row: "A1", Column: "slug", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pAUnlinked", Column: "description", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "groups", Row: "B2", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pBUnlinked", Column: "display_name", Expect: "no_update"},
				{Table: "groups", Row: "A2", Column: "organization_id", Expect: "no_update"},
				// r4 (G019): the raw org_admin -> org_member self-demotion of bAdminA is an admitted write at version 15; moved to
				// rlsExpectedMutations ("adminA org-only demotes itself to org_member"). Its owner promotion stays refused there.
				{Table: "role_bindings", Row: "bAdminA", Column: "scope_type", Expect: "no_update"},
				{Table: "users", Row: "adminA", Column: "status", Expect: "no_update"},
			},
			Basis: "G060 (ADR-008; Plan r0 A11-A13 as amended by S2 amendment01 A1.3/A1.5): without the Group-management override an Organization owner/admin without a Group role reads no Group roster (role_bindings Group rows, group_memberships, the joined access paths) and writes no Group or Business Profile row; Organization memberships, users, Group and profile metadata stay visible. correction02 C02-2 fixture: adminA is a member of A1 (own row and its pA12 path visible). AUTHORIZATION matrix org_admin: organization.read Y, organization.manage_members Y, organization.manage_policy '-', group.read/update 'all', business_profile.profile.read/update Y (Organization path, flag not required: 'linked+flag' applies to Group roles only); gates 1-2. Spec R4, R5, R6, R10. Policy flag not updatable: matrix manage_policy '-'; Ordering step 3 (this actor's guarded writes lock the Organization row) with G008 handoff row locks and Spec R8 (the lock privilege is not a mutation grant; header A1); paired positive 'ownerA org-only'. Structural/self: AC6, Spec R4, Ordering steps 1 and 4, Spec R8.",
		},
		{
			Name:    "adminA selected A4 (Group without profile): A4 only for Group rows, all profiles, no access path",
			Context: rlsContext{Actor: "adminA", Org: "A", Group: "A4", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				OrgMemberships:     []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"ownerA", "adminA", "memberA", "gOwnerA1", "adminA2", "memberA2", "ownerA4", "crossUser"},
				Groups:             []string{"A4"},
				Associations:       nil,
				Profiles:           []string{"pA12", "pA3", "pAUnlinked"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lOrgA", "lOrgA2"},
				LabelAssignments:   []string{"asOrgAA4"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A4", Column: "name", Expect: "no_update"},                      // G060 (override off)
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"}, // G060 (override off)
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				// r4 (G019): bAdminA role self-demotion moved to rlsExpectedMutations ("adminA selected A4 demotes itself").
			},
			Basis: "G060 (ADR-008; Plan r0 A11-A13 as amended by S2 amendment01 A1.3/A1.5): without the Group-management override an Organization owner/admin without a Group role reads no Group roster (role_bindings Group rows, group_memberships, the joined access paths) and writes no Group or Business Profile row; Organization memberships, users, Group and profile metadata stay visible. AUTHORIZATION Resolve step 3 (org_admin may select any Group of the Organization), pin (b), gate 3, matrix note (admin profile paths independent of the selected Group); ADR-004 (unfinished Groups may have no profile). Spec R4 (association needs the Group visible), R5 (header A3), R6. Flag/self not applied: matrix manage_policy '-', Ordering steps 3 and 4, G008 handoff row locks, Spec R8, AC6 (header A1); paired positive 'ownerA selected A1'.",
		},
		// ===== G060: Organization A staff with the Group-management override on (driver toggle; independent literals) =====
		{
			Name:    "ownerA org-only, override on: Group rosters and access paths visible; Group and profile writes admitted",
			Context: rlsContext{Actor: "ownerA", Org: "A", Group: "", Shape: "valid", Override: true},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA", "bGOwnerA1G1", "bGOwnerA1G3", "bAdminAG1", "bAdminA2G2", "bMemberA2G2", "bOwnerA4G4", "bCrossA3"},
				OrgMemberships:     []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				GroupMemberships:   []string{"bGOwnerA1G1", "bGOwnerA1G3", "bAdminAG1", "bAdminA2G2", "bMemberA2G2", "bOwnerA4G4", "bCrossA3"},
				ProfileAccessPaths: []string{"gbpA1/bGOwnerA1G1", "gbpA1/bAdminAG1", "gbpA2/bAdminA2G2", "gbpA2/bMemberA2G2", "gbpA3/bGOwnerA1G3", "gbpA3/bCrossA3"},
				Users:              []string{"ownerA", "adminA", "memberA", "gOwnerA1", "adminA2", "memberA2", "ownerA4", "crossUser"},
				Groups:             []string{"A1", "A2", "A3", "A4"},
				Associations:       []string{"gbpA1", "gbpA2", "gbpA3"},
				Profiles:           []string{"pA12", "pA3", "pAUnlinked"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lOrgA", "lOrgA2"},
				LabelAssignments:   []string{"asOrgAA3", "asOrgAA4"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A2", Column: "name", Expect: "ok"},
				{Table: "groups", Row: "A4", Column: "slug", Expect: "ok"},
				{Table: "business_profiles", Row: "pAUnlinked", Column: "display_name", Expect: "ok"},
				{Table: "groups", Row: "B1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pB1", Column: "display_name", Expect: "no_update"},
				{Table: "groups", Row: "A1", Column: "organization_id", Expect: "no_update"},
				{Table: "role_bindings", Row: "bOwnerA", Column: "role", Expect: "no_update", Guard: "last_owner"},
			},
			Basis: "G060 (ADR-008; Plan r0 §1.2 matrix column 'org_owner/admin, OVR on', A11-A13; amendment01 A1.5: the override-on re-run of the staff shapes): with the override on, Organization owners and admins read Group rosters and their access paths and write Group and Business Profile rows of the Organization (bounded administration); Group content, ownership and the selected-Group restriction are unchanged. Foreign rows and structural changes stay refused (Spec R4, AC6); the sole owner stays protected.",
		},
		{
			Name:    "ownerA selected A1, override on: the selected Group's roster only; sibling Groups stay outside the context",
			Context: rlsContext{Actor: "ownerA", Org: "A", Group: "A1", Shape: "valid", Override: true},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA", "bGOwnerA1G1", "bAdminAG1"},
				OrgMemberships:     []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				GroupMemberships:   []string{"bGOwnerA1G1", "bAdminAG1"},
				ProfileAccessPaths: []string{"gbpA1/bGOwnerA1G1", "gbpA1/bAdminAG1"},
				Users:              []string{"ownerA", "adminA", "memberA", "gOwnerA1", "adminA2", "memberA2", "ownerA4", "crossUser"},
				Groups:             []string{"A1"},
				Associations:       []string{"gbpA1"},
				Profiles:           []string{"pA12", "pA3", "pAUnlinked"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lOrgA", "lOrgA2"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A1", Column: "name", Expect: "ok"},
				{Table: "business_profiles", Row: "pA3", Column: "description", Expect: "ok"},
				{Table: "groups", Row: "A2", Column: "name", Expect: "no_update"},
				{Table: "groups", Row: "A3", Column: "slug", Expect: "no_update"},
			},
			Basis: "G060 (ADR-008; Plan r0 §1.2 matrix column 'org_owner/admin, OVR on', A11-A13; amendment01 A1.5: the override-on re-run of the staff shapes): with the override on, Organization owners and admins read Group rosters and their access paths and write Group and Business Profile rows of the Organization (bounded administration); Group content, ownership and the selected-Group restriction are unchanged. AUTHORIZATION pin (b) and gate 3: the selected Group restricts the Group path.",
		},
		{
			Name:    "adminA org-only, override on: Group rosters and access paths visible; Group and profile writes admitted; no policy update",
			Context: rlsContext{Actor: "adminA", Org: "A", Group: "", Shape: "valid", Override: true},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA", "bGOwnerA1G1", "bGOwnerA1G3", "bAdminAG1", "bAdminA2G2", "bMemberA2G2", "bOwnerA4G4", "bCrossA3"},
				OrgMemberships:     []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				GroupMemberships:   []string{"bGOwnerA1G1", "bGOwnerA1G3", "bAdminAG1", "bAdminA2G2", "bMemberA2G2", "bOwnerA4G4", "bCrossA3"},
				ProfileAccessPaths: []string{"gbpA1/bGOwnerA1G1", "gbpA1/bAdminAG1", "gbpA2/bAdminA2G2", "gbpA2/bMemberA2G2", "gbpA3/bGOwnerA1G3", "gbpA3/bCrossA3"},
				Users:              []string{"ownerA", "adminA", "memberA", "gOwnerA1", "adminA2", "memberA2", "ownerA4", "crossUser"},
				Groups:             []string{"A1", "A2", "A3", "A4"},
				Associations:       []string{"gbpA1", "gbpA2", "gbpA3"},
				Profiles:           []string{"pA12", "pA3", "pAUnlinked"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lOrgA", "lOrgA2"},
				LabelAssignments:   []string{"asOrgAA3", "asOrgAA4"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A3", Column: "name", Expect: "ok"},
				{Table: "groups", Row: "A1", Column: "slug", Expect: "ok"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "ok"},
				{Table: "business_profiles", Row: "pAUnlinked", Column: "description", Expect: "ok"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "groups", Row: "B2", Column: "name", Expect: "no_update"},
			},
			Basis: "G060 (ADR-008; Plan r0 §1.2 matrix column 'org_owner/admin, OVR on', A11-A13; amendment01 A1.5: the override-on re-run of the staff shapes): with the override on, Organization owners and admins read Group rosters and their access paths and write Group and Business Profile rows of the Organization (bounded administration); Group content, ownership and the selected-Group restriction are unchanged. manage_policy stays owner-only.",
		},
		{
			Name:    "adminA selected A4, override on: A4's roster only, no access path (A4 has no profile)",
			Context: rlsContext{Actor: "adminA", Org: "A", Group: "A4", Shape: "valid", Override: true},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA", "bOwnerA4G4"},
				OrgMemberships:     []string{"bOwnerA", "bAdminA", "bMemberA", "bGOwnerA1Org", "bAdminA2Org", "bMemberA2Org", "bOwnerA4Org", "bCrossOrgA"},
				GroupMemberships:   []string{"bOwnerA4G4"},
				ProfileAccessPaths: nil,
				Users:              []string{"ownerA", "adminA", "memberA", "gOwnerA1", "adminA2", "memberA2", "ownerA4", "crossUser"},
				Groups:             []string{"A4"},
				Associations:       nil,
				Profiles:           []string{"pA12", "pA3", "pAUnlinked"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lOrgA", "lOrgA2"},
				LabelAssignments:   []string{"asOrgAA4"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A4", Column: "name", Expect: "ok"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "ok"},
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
			},
			Basis: "G060 (ADR-008; Plan r0 §1.2 matrix column 'org_owner/admin, OVR on', A11-A13; amendment01 A1.5: the override-on re-run of the staff shapes): with the override on, Organization owners and admins read Group rosters and their access paths and write Group and Business Profile rows of the Organization (bounded administration); Group content, ownership and the selected-Group restriction are unchanged. ADR-004 (A4 has no profile).",
		},
		{
			Name:    "memberA org-only: Organization row, own binding and self only",
			Context: rlsContext{Actor: "memberA", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bMemberA"},
				OrgMemberships:     []string{"bMemberA"},
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"memberA"},
				Groups:             nil,
				Associations:       nil,
				Profiles:           nil,
				Organizations:      []string{"A"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pAUnlinked", Column: "display_name", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION matrix org_member: organization.read only (group.read '-', business_profile.profile.read '-', manage_members '-'). Spec R5 ('a caller can inspect its own current bindings'; broader visibility needs the scope-management capability), R6 (self only, no directory), R4/R10 (rows the actor cannot read are neither returned nor changed). Paired positives: 'gOwnerA1 org-only' renames A1 and updates pA12; 'ownerA org-only' updates pAUnlinked.",
		},
		// ===== Organization A Group roles; sibling Groups A1/A2 share pA12 =====
		{
			Name:    "gOwnerA1 org-only: owner of A1 and member of A3; sibling A2 and its link hidden although pA12 is shared",
			Context: rlsContext{Actor: "gOwnerA1", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				// G060 correction02 fixture: A1's owner sees adminA's A1 member binding, its access path and identity.
				RoleBindings:       []string{"bGOwnerA1Org", "bGOwnerA1G1", "bGOwnerA1G3", "bAdminAG1"},
				OrgMemberships:     []string{"bGOwnerA1Org"},
				GroupMemberships:   []string{"bGOwnerA1G1", "bGOwnerA1G3", "bAdminAG1"},
				ProfileAccessPaths: []string{"gbpA1/bGOwnerA1G1", "gbpA3/bGOwnerA1G3", "gbpA1/bAdminAG1"},
				Users:              []string{"gOwnerA1", "adminA"},
				Groups:             []string{"A1", "A3"},
				Associations:       []string{"gbpA1", "gbpA3"},
				Profiles:           []string{"pA12", "pA3"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lA3Tm"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A1", Column: "name", Expect: "ok"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "ok"},
				{Table: "groups", Row: "A3", Column: "name", Expect: "no_update"},
				{Table: "groups", Row: "A2", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA3", Column: "display_name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pAUnlinked", Column: "display_name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "role_bindings", Row: "bGOwnerA1Org", Column: "role", Expect: "no_update"},
				{Table: "role_bindings", Row: "bGOwnerA1G3", Column: "role", Expect: "no_update"},
				{Table: "role_bindings", Row: "bGOwnerA1G3", Column: "group_id", Expect: "no_update"},
				{Table: "users", Row: "gOwnerA1", Column: "status", Expect: "no_update"},
				{Table: "group_business_profiles", Row: "gbpA1", Column: "group_id", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION matrix group_owner/group_member: group.read 'own', group.update owner 'own' (member '-'), group.manage_members owner 'own' (A1 has no other binding), profile.read 'linked' (candidates = every subject Group), profile.update 'linked+flag' (Organization A flag true), member-only paths read only; pin (a). Spec R4 ('a shared profile exposes neither sibling Groups nor their links/memberships'; association needs both visibilities), R5, R6, AC5, AC7; ADR-004 (managing a profile does not grant access to all Groups using it). A3 rename, own bindings, users row and Organization flag not applied: group.update/manage_policy deny; an Organization-only pA12 update locks every candidate Group FOR SHARE (Ordering step 6), all own Group bindings (step 5), the Organization binding (step 4), the Organization row (step 3) and the users row (step 1), and the privilege those locks need must not become a mutation grant (G008 handoff row locks, Spec R8, AC6; header A1). pA3 member path: read only. Paired positives: 'crossUser A org-only' renames A3 and updates pA3; 'adminA2 org-only' renames A2; 'ownerA org-only' updates the flag and pAUnlinked. Link retarget: Spec R1/R4, AC6 (header A6).",
		},
		{
			Name:    "gOwnerA1 selected A1: only A1, its link and pA12; A3 path excluded by the selected Group",
			Context: rlsContext{Actor: "gOwnerA1", Org: "A", Group: "A1", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bGOwnerA1Org", "bGOwnerA1G1", "bAdminAG1"}, // G060 correction02 fixture
				OrgMemberships:     []string{"bGOwnerA1Org"},
				GroupMemberships:   []string{"bGOwnerA1G1", "bAdminAG1"},
				ProfileAccessPaths: []string{"gbpA1/bGOwnerA1G1", "gbpA1/bAdminAG1"},
				Users:              []string{"gOwnerA1", "adminA"},
				Groups:             []string{"A1"},
				Associations:       []string{"gbpA1"},
				Profiles:           []string{"pA12"},
				Organizations:      []string{"A"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A1", Column: "slug", Expect: "ok"},
				{Table: "business_profiles", Row: "pA12", Column: "description", Expect: "ok"},
				{Table: "groups", Row: "A3", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA3", Column: "display_name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "role_bindings", Row: "bGOwnerA1G1", Column: "role", Expect: "no_update", Guard: "last_owner"}, // r4: sole owner of A1
			},
			Basis: "AUTHORIZATION 'linked' (in selected-Group context the candidate is the selected Group only), gate 3, matrix group.read/update 'own'. Spec R4 (selected Group limits inherited paths), R5 (own bindings within the selected scope; header A2), R6. Flag/self not applied: matrix manage_policy '-', Ordering steps 3-5 for this actor's A1 writes, G008 handoff row locks, Spec R8, AC6 (header A1). Paired positives: 'gOwnerA1 org-only' sees A3/pA3; 'crossUser A org-only' updates pA3 and renames A3.",
		},
		{
			Name:    "gOwnerA1 selected A3: member path only; shared pA12 and A1 excluded",
			Context: rlsContext{Actor: "gOwnerA1", Org: "A", Group: "A3", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bGOwnerA1Org", "bGOwnerA1G3"},
				OrgMemberships:     []string{"bGOwnerA1Org"},
				GroupMemberships:   []string{"bGOwnerA1G3"},
				ProfileAccessPaths: []string{"gbpA3/bGOwnerA1G3"},
				Users:              []string{"gOwnerA1"},
				Groups:             []string{"A3"},
				Associations:       []string{"gbpA3"},
				Profiles:           []string{"pA3"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lA3Tm"},
			},
			Updates: []rlsUpdate{
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION 'linked' (candidate = selected Group A3 only), group.manage_members '-' for group_member, gate 3, pin (d) (selected-Group context restricts Group-derived paths). Spec R4 (selected Group limits inherited paths; shared profile exposes no sibling links/memberships, so crossUser's A3 owner binding stays hidden), R5 (header A2), R6. Paired positives: 'gOwnerA1 selected A1' updates pA12 and renames A1.",
		},
		{
			Name:    "adminA2 org-only: admin of sibling A2 sees A2 members and its own link only, edits shared pA12 with flag on",
			Context: rlsContext{Actor: "adminA2", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bAdminA2Org", "bAdminA2G2", "bMemberA2G2"},
				OrgMemberships:     []string{"bAdminA2Org"},
				GroupMemberships:   []string{"bAdminA2G2", "bMemberA2G2"},
				ProfileAccessPaths: []string{"gbpA2/bAdminA2G2", "gbpA2/bMemberA2G2"},
				Users:              []string{"adminA2", "memberA2"},
				Groups:             []string{"A2"},
				Associations:       []string{"gbpA2"},
				Profiles:           []string{"pA12"},
				Organizations:      []string{"A"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A2", Column: "name", Expect: "ok"},
				{Table: "groups", Row: "A2", Column: "slug", Expect: "ok"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "ok"},
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA3", Column: "display_name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				// r4 (G019): the raw group_admin -> group_member self-demotion of bAdminA2G2 is an admitted write at version 15;
				// moved to rlsExpectedMutations ("adminA2 org-only demotes itself to group_member").
				{Table: "users", Row: "adminA2", Column: "status", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION matrix group_admin: group.read/update/manage_members 'own', profile.read 'linked', profile.update 'linked+flag' (A flag true). Spec R4 (shared pA12 exposes neither sibling A1 nor gbpA1 nor A1 memberships), R5, R6, AC5 ('two linked Groups'; a second qualifying attached Group gives only its own path), AC7; ADR-004 additional owner decisions (admin cannot promote itself). Flag/self not applied: matrix manage_policy '-', Ordering steps 1, 3, 5, G008 handoff row locks, Spec R8, AC6 (header A1). Paired positives: 'gOwnerA1 org-only' renames A1; 'crossUser A org-only' updates pA3; 'ownerA org-only' updates the flag.",
		},
		{
			Name:    "adminA2 with unselectable A1 selected: invalid selection, no tenant or projection row, global self row only",
			Context: rlsContext{Actor: "adminA2", Org: "A", Group: "A1", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       nil,
				OrgMemberships:     nil,
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"adminA2"},
				Groups:             nil,
				Associations:       nil,
				Profiles:           nil,
				Organizations:      nil,
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "groups", Row: "A2", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
			},
			Basis: "Context-validity clarification, required case 1 (Group admin with Organization member status and no binding in the selected same-Organization Group: no tenant/projection rows; permitted global self row) and section 'Tenant-selection validity precedes capability checks' (an unselectable selection denies every tenant-owned row and projection, including Organization rows and the subject's own Organization binding; organization.read cannot override it). AUTHORIZATION Resolve step 3 (A1 not selectable: no org_owner/org_admin role and no A1 binding). Spec R6 (admit self), R3 trust limit (substituted scope ID with fixed actor). Reading per header A5. Paired positives: 'adminA2 org-only' sees the A Organization row and own bindings, renames A2 and updates pA12; 'gOwnerA1 selected A1' renames A1.",
		},
		{
			Name:    "memberA2 org-only: member reads shared pA12 through A2, sees no co-member and cannot edit",
			Context: rlsContext{Actor: "memberA2", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bMemberA2Org", "bMemberA2G2"},
				OrgMemberships:     []string{"bMemberA2Org"},
				GroupMemberships:   []string{"bMemberA2G2"},
				ProfileAccessPaths: []string{"gbpA2/bMemberA2G2"},
				Users:              []string{"memberA2"},
				Groups:             []string{"A2"},
				Associations:       []string{"gbpA2"},
				Profiles:           []string{"pA12"},
				Organizations:      []string{"A"},
			},
			Updates: []rlsUpdate{
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA3", Column: "display_name", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION matrix group_member: group.read 'own', profile.read 'linked', profile.update '-' ('Member-only paths allow reading only'), group.manage_members '-'. ADR-004 (members derive read only). Spec R4, R5 (own bindings only; no manage capability), R6 (self only), AC5. Paired positive: 'adminA2 org-only' updates pA12.",
		},
		{
			Name:    "ownerA4 org-only: owner of a Group without profile sees no profile and never reaches unlinked profiles",
			Context: rlsContext{Actor: "ownerA4", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bOwnerA4Org", "bOwnerA4G4"},
				OrgMemberships:     []string{"bOwnerA4Org"},
				GroupMemberships:   []string{"bOwnerA4G4"},
				ProfileAccessPaths: nil,
				Users:              []string{"ownerA4"},
				Groups:             []string{"A4"},
				Associations:       nil,
				Profiles:           nil,
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lA4Own", "lA4Tm"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A4", Column: "name", Expect: "ok"},
				{Table: "business_profiles", Row: "pAUnlinked", Column: "display_name", Expect: "no_update"},
				{Table: "groups", Row: "A3", Column: "name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "role_bindings", Row: "bOwnerA4G4", Column: "source", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION matrix group_owner (group.read/update 'own'; profile paths only through a current association; only Organization owner/admin paths include unlinked profiles). ADR-004 (unfinished Groups may have no profile; no independent BusinessProfile grants). Spec R4, R5, R6. Flag/source not applied: matrix manage_policy '-', Ordering steps 3 and 5 for this actor's A4 rename, G008 handoff row locks, Spec R4 (cannot change role/source), R8, AC6 (header A1). Paired positives: 'ownerA org-only' updates pAUnlinked and the flag; 'crossUser A org-only' renames A3.",
		},
		// ===== One user managing profiles through Groups in two Organizations =====
		{
			Name:    "crossUser in A org-only: manages pA3 through A3; Organization B rows and B management hidden",
			Context: rlsContext{Actor: "crossUser", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bCrossOrgA", "bCrossA3", "bGOwnerA1G3"},
				OrgMemberships:     []string{"bCrossOrgA"},
				GroupMemberships:   []string{"bCrossA3", "bGOwnerA1G3"},
				ProfileAccessPaths: []string{"gbpA3/bCrossA3", "gbpA3/bGOwnerA1G3"},
				Users:              []string{"crossUser", "gOwnerA1"},
				Groups:             []string{"A3"},
				Associations:       []string{"gbpA3"},
				Profiles:           []string{"pA3"},
				Organizations:      []string{"A"},
				LabelDefinitions:   []string{"lA3Own", "lA3Tm"},
			},
			Updates: []rlsUpdate{
				{Table: "business_profiles", Row: "pA3", Column: "display_name", Expect: "ok"},
				{Table: "groups", Row: "A3", Column: "name", Expect: "ok"},
				{Table: "business_profiles", Row: "pB1", Column: "display_name", Expect: "no_update"},
				{Table: "groups", Row: "B1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "organizations", Row: "B", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "role_bindings", Row: "bCrossA3", Column: "role", Expect: "no_update", Guard: "last_owner"}, // r4: sole owner of A3
			},
			Basis: "Spec inherited owner clarification and AC5 ('one user manages profiles through Groups in two Organizations; active-tenant queries exclude the other tenant'), R4 (only the active Organization's rows), R5 (group.manage_members 'own' exposes A3 bindings only), R6 (gOwnerA1 shares inspectable scope A3). AUTHORIZATION gate 1-2 (active Organization only), matrix group_owner, 'linked+flag' (A flag true). ADR-004 (a global user may manage profiles in multiple Organizations but sees only active-tenant ones). Flag/self not applied: matrix manage_policy '-', Ordering steps 3 and 5, G008 handoff row locks, Spec R8, AC6 (header A1). Paired positives: 'crossUser in B org-only' renames B1; 'ownerB org-only' updates pB1 and the B flag; 'adminA2 org-only' updates pA12.",
		},
		{
			Name:    "crossUser in B org-only (flag off): group_admin of B1 reads pB1 but cannot edit it; Organization A hidden",
			Context: rlsContext{Actor: "crossUser", Org: "B", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bCrossOrgB", "bCrossB1", "bMemberB1G1"},
				OrgMemberships:     []string{"bCrossOrgB"},
				GroupMemberships:   []string{"bCrossB1", "bMemberB1G1"},
				ProfileAccessPaths: []string{"gbpB1/bCrossB1", "gbpB1/bMemberB1G1"},
				Users:              []string{"crossUser", "memberB1"},
				Groups:             []string{"B1"},
				Associations:       []string{"gbpB1"},
				Profiles:           []string{"pB1"},
				Organizations:      []string{"B"},
				LabelDefinitions:   []string{"lB1Tm"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "B1", Column: "name", Expect: "ok"},
				{Table: "business_profiles", Row: "pB1", Column: "display_name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA3", Column: "display_name", Expect: "no_update"},
				{Table: "groups", Row: "A3", Column: "name", Expect: "no_update"},
				{Table: "organizations", Row: "B", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				// r4 (G019): the raw group_admin -> group_member self-demotion of bCrossB1 is an admitted write at version 15;
				// moved to rlsExpectedMutations ("crossUser in B org-only demotes itself to group_member").
				{Table: "users", Row: "crossUser", Column: "status", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION 'linked+flag' (Organization B group_admins_edit_business_profiles false: Group-role profile update denied), Policy flag ('affects inherited ordinary profile updates only, never reads or selection'), matrix group_admin group.update/manage_members 'own'. Spec AC5 (both policy values; active-tenant exclusion), R4, R5, R6. Flag/self not applied: matrix manage_policy '-', Ordering steps 1, 3, 5 for this actor's B1 rename, G008 handoff row locks, Spec R8, AC6 (header A1). Paired positives: 'ownerB org-only' and 'adminB org-only' update pB1; 'crossUser in A org-only' updates pA3 and renames A3.",
		},
		{
			Name:    "crossUser selected B1 (flag off): same B1 rows; rename allowed, profile edit still refused",
			Context: rlsContext{Actor: "crossUser", Org: "B", Group: "B1", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bCrossOrgB", "bCrossB1", "bMemberB1G1"},
				OrgMemberships:     []string{"bCrossOrgB"},
				GroupMemberships:   []string{"bCrossB1", "bMemberB1G1"},
				ProfileAccessPaths: []string{"gbpB1/bCrossB1", "gbpB1/bMemberB1G1"},
				Users:              []string{"crossUser", "memberB1"},
				Groups:             []string{"B1"},
				Associations:       []string{"gbpB1"},
				Profiles:           []string{"pB1"},
				Organizations:      []string{"B"},
				LabelDefinitions:   []string{"lB1Tm"},
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "B1", Column: "slug", Expect: "ok"},
				{Table: "business_profiles", Row: "pB1", Column: "description", Expect: "no_update"},
				{Table: "organizations", Row: "B", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION 'linked+flag' (candidate = selected B1; flag false), gate 3 (B1 is the selected Group), matrix group_admin 'own', manage_policy '-'. Spec R4, R5 (header A2), R6, R8 (header A1), AC5 (both policy values). Paired positives: 'ownerB org-only' updates pB1 and the B flag.",
		},
		// ===== Organization B owner, admin and member =====
		{
			Name:    "ownerB org-only (flag off): every Organization B row; owner path ignores the flag; Organization A filtered",
			Context: rlsContext{Actor: "ownerB", Org: "B", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				// G060 (override off; amendment01 A1.5 applied to the same staff shape in Organization B).
				RoleBindings:       []string{"bOwnerB", "bAdminB", "bCrossOrgB", "bMemberB1Org"},
				OrgMemberships:     []string{"bOwnerB", "bAdminB", "bCrossOrgB", "bMemberB1Org"},
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"ownerB", "adminB", "crossUser", "memberB1"},
				Groups:             []string{"B1", "B2"},
				Associations:       []string{"gbpB1"},
				Profiles:           []string{"pB1", "pBUnlinked"},
				Organizations:      []string{"B"},
				LabelDefinitions:   []string{"lOrgB"},
				LabelAssignments:   []string{"asOrgBB1"},
			},
			Updates: []rlsUpdate{
				{Table: "business_profiles", Row: "pB1", Column: "display_name", Expect: "no_update"}, // G060 (override off)
				{Table: "organizations", Row: "B", Column: "group_admins_edit_business_profiles", Expect: "ok"},
				{Table: "groups", Row: "B2", Column: "name", Expect: "no_update"}, // G060 (override off)
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
			},
			Basis: "G060 (ADR-008; Plan r0 A11-A13 as amended by S2 amendment01 A1.3/A1.5): without the Group-management override an Organization owner/admin without a Group role reads no Group roster (role_bindings Group rows, group_memberships, the joined access paths) and writes no Group or Business Profile row; Organization memberships, users, Group and profile metadata stay visible. AUTHORIZATION matrix org_owner (profile.update Y without the flag; manage_policy Y; group.read/update 'all'), Policy flag (affects inherited Group-role updates only), gates 1-2. Spec R4, R5, R6 (crossUser visible through its B bindings only), R10, AC4 (foreign rows not returned or changed), AC5 (both policy values). Paired positives: 'ownerA org-only' updates the A flag and A Group rows; 'adminA2 org-only' updates pA12.",
		},
		{
			Name:    "adminB org-only (flag off): every Organization B row; admin edits profiles; no policy update",
			Context: rlsContext{Actor: "adminB", Org: "B", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				// G060 (override off; amendment01 A1.5 applied to the same staff shape in Organization B).
				RoleBindings:       []string{"bOwnerB", "bAdminB", "bCrossOrgB", "bMemberB1Org"},
				OrgMemberships:     []string{"bOwnerB", "bAdminB", "bCrossOrgB", "bMemberB1Org"},
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"ownerB", "adminB", "crossUser", "memberB1"},
				Groups:             []string{"B1", "B2"},
				Associations:       []string{"gbpB1"},
				Profiles:           []string{"pB1", "pBUnlinked"},
				Organizations:      []string{"B"},
				LabelDefinitions:   []string{"lOrgB"},
				LabelAssignments:   []string{"asOrgBB1"},
			},
			Updates: []rlsUpdate{
				{Table: "business_profiles", Row: "pB1", Column: "display_name", Expect: "no_update"},       // G060 (override off)
				{Table: "business_profiles", Row: "pBUnlinked", Column: "description", Expect: "no_update"}, // G060 (override off)
				{Table: "groups", Row: "B1", Column: "slug", Expect: "no_update"},                           // G060 (override off)
				{Table: "organizations", Row: "B", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				// r4 (G019): the raw org_admin -> org_member self-demotion of bAdminB is an admitted write at version 15; moved to
				// rlsExpectedMutations ("adminB org-only demotes itself to org_member").
				{Table: "business_profiles", Row: "pA3", Column: "display_name", Expect: "no_update"},
			},
			Basis: "G060 (ADR-008; Plan r0 A11-A13 as amended by S2 amendment01 A1.3/A1.5): without the Group-management override an Organization owner/admin without a Group role reads no Group roster (role_bindings Group rows, group_memberships, the joined access paths) and writes no Group or Business Profile row; Organization memberships, users, Group and profile metadata stay visible. AUTHORIZATION matrix org_admin (profile.update Y independent of the flag; manage_policy '-'), Policy flag, gates 1-2. Spec R4, R5, R6, AC5. Flag/self not applied: Ordering steps 3 and 4, G008 handoff row locks, Spec R8, AC6 (header A1). Paired positives: 'ownerB org-only' updates the B flag; 'crossUser in A org-only' updates pA3.",
		},
		{
			Name:    "memberB1 org-only: member of B1 reads pB1; no co-member, no edit",
			Context: rlsContext{Actor: "memberB1", Org: "B", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       []string{"bMemberB1Org", "bMemberB1G1"},
				OrgMemberships:     []string{"bMemberB1Org"},
				GroupMemberships:   []string{"bMemberB1G1"},
				ProfileAccessPaths: []string{"gbpB1/bMemberB1G1"},
				Users:              []string{"memberB1"},
				Groups:             []string{"B1"},
				Associations:       []string{"gbpB1"},
				Profiles:           []string{"pB1"},
				Organizations:      []string{"B"},
				LabelDefinitions:   []string{"lB1Tm"},
			},
			Updates: []rlsUpdate{
				{Table: "business_profiles", Row: "pB1", Column: "display_name", Expect: "no_update"},
				{Table: "groups", Row: "B2", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pBUnlinked", Column: "display_name", Expect: "no_update"},
			},
			Basis: "AUTHORIZATION matrix group_member (group.read 'own', profile.read 'linked', no update, no manage_members). Spec R4, R5, R6. Paired positives: 'ownerB org-only' renames B2 and updates pB1; 'adminB org-only' updates pBUnlinked.",
		},
		// ===== Foreign tenant and global-login-only actors (well-formed context, header A5) =====
		{
			Name:    "memberB1 with Organization A active (foreign tenant): self only, every A row filtered",
			Context: rlsContext{Actor: "memberB1", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       nil,
				OrgMemberships:     nil,
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"memberB1"},
				Groups:             nil,
				Associations:       nil,
				Profiles:           nil,
				Organizations:      nil,
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "group_business_profiles", Row: "gbpA1", Column: "group_id", Expect: "no_update"},
			},
			Basis: "Context-validity clarification (a foreign selection is not a valid TenantContext and denies every tenant-owned row and projection; the established actor's global self row stays visible). AUTHORIZATION Resolve step 3 and gate 1 (no Organization A membership). Spec R3 trust limit (fixed actor, substituted tenant ID), R4 (foreign rows cannot be mutated), R6 (admit self; foreign identities undisclosed), AC4. Link UPDATE: not a permitted command (Spec R1/R4, header A6). Reading per header A5. Paired positives: 'ownerA org-only' updates the flag and A1; 'gOwnerA1 org-only' updates pA12.",
		},
		{
			Name:    "ownerB with Organization A active (foreign owner): an owner role in B confers nothing in A",
			Context: rlsContext{Actor: "ownerB", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       nil,
				OrgMemberships:     nil,
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"ownerB"},
				Groups:             nil,
				Associations:       nil,
				Profiles:           nil,
				Organizations:      nil,
			},
			Updates: []rlsUpdate{
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
				{Table: "groups", Row: "A2", Column: "name", Expect: "no_update"},
				{Table: "business_profiles", Row: "pAUnlinked", Column: "display_name", Expect: "no_update"},
			},
			Basis: "Context-validity clarification (a foreign selection is not a valid TenantContext: no tenant/projection row; global self row only). AUTHORIZATION Resolve step 3, gates 1-2 ('a foreign ... target denies, even for a subject holding the named role'; no global authority). Spec R3 trust limit, R4, R6, AC4. Reading per header A5. Paired positives: 'ownerA org-only' updates the flag, A2 and pAUnlinked.",
		},
		{
			Name:    "outsider identity-only: global login without any binding sees only itself",
			Context: rlsContext{Actor: "outsider", Org: "", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       nil,
				OrgMemberships:     nil,
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"outsider"},
				Groups:             nil,
				Associations:       nil,
				Profiles:           nil,
				Organizations:      nil,
			},
			Updates: []rlsUpdate{
				{Table: "groups", Row: "A1", Column: "name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
			},
			Basis: "Spec R6 (pre-tenant account self-read preserved before any Organization exists; no directory access from mere global login), R3 (no default Organization), AC8. Paired positives: 'ownerA org-only' updates A1 and the flag.",
		},
		{
			Name:    "outsider with Organization A active: non-member sees only itself",
			Context: rlsContext{Actor: "outsider", Org: "A", Group: "", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       nil,
				OrgMemberships:     nil,
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"outsider"},
				Groups:             nil,
				Associations:       nil,
				Profiles:           nil,
				Organizations:      nil,
			},
			Updates: []rlsUpdate{
				{Table: "business_profiles", Row: "pA12", Column: "display_name", Expect: "no_update"},
				{Table: "organizations", Row: "A", Column: "group_admins_edit_business_profiles", Expect: "no_update"},
			},
			Basis: "Context-validity clarification (selecting an Organization without current membership is not a valid TenantContext: no tenant/projection row; the live actor's global self row stays visible and the actor is not treated as unauthenticated). AUTHORIZATION Resolve step 3, gate 1. Spec R3 trust limit, R4, R6 (admit self only), AC3 (foreign scope denies), AC4. Reading per header A5. Paired positives: 'adminA2 org-only' updates pA12; 'ownerA org-only' updates the flag.",
		},
		// ===== Invalid context shapes: each varies one field of a paired valid case =====
		{
			Name:    "no context initialized: context-free access returns nothing",
			Context: rlsContext{Actor: "", Org: "", Group: "", Shape: "no-context"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (complete transaction-local context initialization before protected queries; missing fields fail closed with no data or mutation), R7/AC9 ('context-free access denies'), AC4 (missing context). Paired positive: 'ownerA org-only'.",
		},
		{
			Name:    "actor setting never initialized, Organization A set: partial initialization fails closed",
			Context: rlsContext{Actor: "", Org: "A", Group: "", Shape: "missing-actor"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (missing required fields and partial initialization fail closed; errors are not allow fallbacks), AC3, AC4. Paired positive: 'ownerA org-only' (same Organization, actor present).",
		},
		{
			Name:    "actor setting never initialized, Organization A and Group A1 set: partial initialization fails closed",
			Context: rlsContext{Actor: "", Org: "A", Group: "A1", Shape: "missing-actor"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (missing required field, partial initialization), AC3, AC4. Paired positive: 'gOwnerA1 selected A1'.",
		},
		{
			Name:    "actor setting empty string, Organization A set: fails closed",
			Context: rlsContext{Actor: "", Org: "A", Group: "", Shape: "empty-actor"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 ('Missing/empty required fields ... fail closed'), AC3, AC4. Paired positive: 'ownerA org-only'.",
		},
		{
			Name:    "actor setting all-zero UUID, Organization A set: fails closed",
			Context: rlsContext{Actor: "", Org: "A", Group: "", Shape: "zero-actor"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (zero UUIDs fail closed; no global sentinel), AC3. Paired positive: 'ownerA org-only'.",
		},
		{
			Name:    "actor setting all-zero UUID, identity-only: no self row is reachable through a sentinel",
			Context: rlsContext{Actor: "", Org: "", Group: "", Shape: "zero-actor"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (zero UUIDs fail closed), R6 (a supplied user UUID cannot select a bootstrap path; no remotely selectable synthetic identity), AC3, AC8. Paired positive: 'ownerA identity-only'.",
		},
		{
			Name:    "actor setting malformed, Organization A set: fails closed",
			Context: rlsContext{Actor: "", Org: "A", Group: "", Shape: "malformed-actor"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (malformed settings fail closed; errors are not allow fallbacks), AC3, AC4 (malformed context). Paired positive: 'ownerA org-only'.",
		},
		{
			Name:    "Organization setting all-zero UUID for ownerA: no tenant-zero sentinel",
			Context: rlsContext{Actor: "ownerA", Org: "", Group: "", Shape: "zero-org"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 ('zero UUIDs ... fail closed'; 'No default Organization, tenant-zero/global sentinel'), AC3; AUTHORIZATION Resolve step 2 (all-zero selectors invalid). Paired positive: 'ownerA org-only'.",
		},
		{
			Name:    "Organization setting malformed for ownerA: fails closed",
			Context: rlsContext{Actor: "ownerA", Org: "", Group: "", Shape: "malformed-org"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (malformed settings fail closed), AC3, AC4; AUTHORIZATION Resolve step 2 (malformed selectors invalid). Paired positive: 'ownerA org-only'.",
		},
		{
			Name:    "Group A1 set without an Organization for gOwnerA1: partial and inconsistent, fails closed",
			Context: rlsContext{Actor: "gOwnerA1", Org: "", Group: "A1", Shape: "group-without-org"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (partial initialization and inconsistent settings fail closed; Group is optional only within an active Organization), AC3; AUTHORIZATION Resolve (a Group selector is a Group of the selected Organization). Paired positive: 'gOwnerA1 selected A1'.",
		},
		{
			Name:    "Group setting all-zero UUID for gOwnerA1 in A: absent Group must be cleared, not zero",
			Context: rlsContext{Actor: "gOwnerA1", Org: "A", Group: "", Shape: "zero-group"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 ('Explicitly clear an absent optional Group'; zero UUIDs fail closed), AC3; AUTHORIZATION Resolve step 2. Paired positives: 'gOwnerA1 org-only' and 'gOwnerA1 selected A1'.",
		},
		{
			Name:    "Group setting malformed for gOwnerA1 in A: fails closed",
			Context: rlsContext{Actor: "gOwnerA1", Org: "A", Group: "", Shape: "malformed-group"},
			Visible: nil,
			Reject:  "rejected-no-rows-no-change",
			Basis:   "Spec R3 (malformed settings fail closed), AC3, AC4; AUTHORIZATION Resolve step 2. Paired positives: 'gOwnerA1 org-only' and 'gOwnerA1 selected A1'.",
		},
		// Structurally valid shape with a data-invalid selection: not a Reject case (r2, header A5).
		{
			Name:    "Organization A with Group B1 for crossUser (member of both): invalid selection, no tenant or shared-scope identity row, global self row only",
			Context: rlsContext{Actor: "crossUser", Org: "A", Group: "B1", Shape: "valid"},
			Visible: &rlsVisible{
				RoleBindings:       nil,
				OrgMemberships:     nil,
				GroupMemberships:   nil,
				ProfileAccessPaths: nil,
				Users:              []string{"crossUser"},
				Groups:             nil,
				Associations:       nil,
				Profiles:           nil,
				Organizations:      nil,
			},
			Basis: "Context-validity clarification, required case 2 (fixed authenticated actor selects Organization A with a Group belonging to B: no tenant/projection or shared-scope identity rows; permitted global self row) and section 'Structural rejection versus global self liveness' (a structurally valid shape whose selectors fail data-dependent resolution does not conceal the actor's global self row). AUTHORIZATION Resolve step 3 (the Group must belong to that Organization). Spec R3, R6 (admit self), AC3 (foreign scope). Reading per header A5 (r1 treated this as a rejected shape). Paired positives: 'crossUser in A org-only' and 'crossUser selected B1'.",
		},
	}
}

// --- oracle r3 (G018): the four INSERT policies and the association DELETE refusal --------------------------------------
//
// REVISION r3 (2026-10-17, G018 Plan r5 §6.2 group (5)). Reason: version 14 grants the runtime role keyed INSERT on
// app.groups and app.business_profiles, INSERT on app.group_business_profiles, and a trigger-only (depth > 0) owner
// binding INSERT on app.role_bindings. Additions only: the rlsInsert attempts below, each an independent raw single-row
// statement from the unmodified fixture state in its own rolled-back transaction. Unchanged: the fixture, every
// context, visible set, UPDATE expectation and Reject class above. Expect values: "ok" (exactly one row inserted) or
// "denied" (SQLSTATE 42501, no row).
// SOURCES: G018 Spec R1/R2/R3 and the Plan r5 §1.3 matrix, §1.4 creation/completion contracts and §5.6 settled
// interfaces (group.create for every current Organization membership role in Organization-only context, superseding
// D-1; the creator's owner binding only from the creation statement itself; a new profile by an Organization
// owner/admin or a Group owner/admin; an association by an Organization owner/admin, or by an owner/admin of the target
// Group for a profile that is either inserted by the same transaction or attached to a Group on which the subject
// holds group_owner/group_admin; no association DELETE). INDEPENDENCE NOTE (deviation, recorded in the Result): unlike
// r0–r2 these rows were authored by the G018 implementer, who had read migration 00014; each expectation is stated
// from the rules above and its Basis, not from a runtime result.
// LIMITATIONS (as A7): the fixture has no subject that owns an incomplete Group and also holds owner/admin on an
// attached Group, and no subject that can both update a committed profile and complete a Group; the eligibility
// positive through an attached Group and the UPDATEd-committed-profile refusal are therefore covered by ONB-9/ONB-9e,
// O-ONB-E and CT-ONB-1/R-ONB-1, not here. The fixture is not changed (it would change every visible set above).

type rlsInsert struct {
	Name    string
	Context rlsContext
	Op      string // "insert" or "delete"
	Table   string // "groups", "role_bindings", "business_profiles", "group_business_profiles"
	Org     string // organization_id label of the attempted row
	Group   string // role_bindings / group_business_profiles: the Group label
	Profile string // group_business_profiles: the profile label, or "new" (inserted first by the same transaction)
	Subject string // role_bindings: the subject label
	Row     string // delete: the association label
	Expect  string // "ok" or "denied"
	Basis   string
}

func rlsExpectedInserts() []rlsInsert {
	org := func(a, o string) rlsContext { return rlsContext{Actor: a, Org: o, Shape: "valid"} }
	ovr := func(a, o string) rlsContext { return rlsContext{Actor: a, Org: o, Shape: "valid", Override: true} } // G060
	sel := func(a, o, g string) rlsContext { return rlsContext{Actor: a, Org: o, Group: g, Shape: "valid"} }
	const create, prof, assoc, boot = "Plan r5 §1.3 group.create; Spec R1 (superseding D-1)", "Plan r5 §1.4 S-P1; matrix group.complete_onboarding", "Plan r5 §1.4 S-P2; ADR-004; Spec R2", "Plan r5 §1.4 S-G1 bootstrap; §5.6 item 2"
	return []rlsInsert{
		{Name: "r3 ownerA creates a Group in A", Context: org("ownerA", "A"), Op: "insert", Table: "groups", Org: "A", Expect: "ok", Basis: create},
		{Name: "r3 adminA creates a Group in A", Context: org("adminA", "A"), Op: "insert", Table: "groups", Org: "A", Expect: "ok", Basis: create},
		{Name: "r3 memberA creates a Group in A", Context: org("memberA", "A"), Op: "insert", Table: "groups", Org: "A", Expect: "ok", Basis: create},
		{Name: "r3 gOwnerA1 (org_member with Group roles) creates a Group in A", Context: org("gOwnerA1", "A"), Op: "insert", Table: "groups", Org: "A", Expect: "ok", Basis: create},
		{Name: "r3 ownerA in selected A1 cannot create", Context: sel("ownerA", "A", "A1"), Op: "insert", Table: "groups", Org: "A", Expect: "denied", Basis: create + "; Organization-only context"},
		{Name: "r3 adminA2 in selected A2 cannot create", Context: sel("adminA2", "A", "A2"), Op: "insert", Table: "groups", Org: "A", Expect: "denied", Basis: create + "; Organization-only context"},
		{Name: "r3 ownerB cannot create a Group in A from B", Context: org("ownerB", "B"), Op: "insert", Table: "groups", Org: "A", Expect: "denied", Basis: create + "; Organization boundary"},
		{Name: "r3 outsider cannot create a Group in A", Context: org("outsider", "A"), Op: "insert", Table: "groups", Org: "A", Expect: "denied", Basis: create + "; context validity (r2)"},
		{Name: "r3 ownerA cannot write its own owner binding on A4 directly", Context: org("ownerA", "A"), Op: "insert", Table: "role_bindings", Org: "A", Group: "A4", Subject: "ownerA", Expect: "denied", Basis: boot},
		{Name: "r3 memberA cannot write its own owner binding on A1 directly", Context: org("memberA", "A"), Op: "insert", Table: "role_bindings", Org: "A", Group: "A1", Subject: "memberA", Expect: "denied", Basis: boot},
		{Name: "r3 ownerA4 cannot write another subject's owner binding on A4", Context: org("ownerA4", "A"), Op: "insert", Table: "role_bindings", Org: "A", Group: "A4", Subject: "memberA", Expect: "denied", Basis: boot},
		{Name: "r3 ownerA creates a profile in A", Context: org("ownerA", "A"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "denied", Basis: prof + "; G060 r0 A13: without the override an Organization owner without a Group owner/admin role inserts no profile"},
		{Name: "r3 adminA creates a profile in A", Context: org("adminA", "A"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "denied", Basis: prof + "; G060 r0 A13: without the override (adminA is only a member of A1)"},
		{Name: "r3 ownerA creates a profile in A, override on", Context: ovr("ownerA", "A"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "ok", Basis: prof + "; G060 r0 A13: (MGR AND OVR)"},
		{Name: "r3 adminA creates a profile in A, override on", Context: ovr("adminA", "A"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "ok", Basis: prof + "; G060 r0 A13: (MGR AND OVR)"},
		{Name: "r3 ownerA4 (group_owner) creates a profile in A", Context: org("ownerA4", "A"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "ok", Basis: prof},
		{Name: "r3 adminA2 (group_admin) creates a profile in A", Context: org("adminA2", "A"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "ok", Basis: prof},
		{Name: "r3 ownerA4 in selected A4 creates a profile", Context: sel("ownerA4", "A", "A4"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "ok", Basis: prof},
		{Name: "r3 memberA (org_member only) cannot create a profile", Context: org("memberA", "A"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "denied", Basis: prof},
		{Name: "r3 memberA2 (group_member only) cannot create a profile", Context: org("memberA2", "A"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "denied", Basis: prof},
		{Name: "r3 gOwnerA1 in selected A3 (group_member there) cannot create a profile", Context: sel("gOwnerA1", "A", "A3"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "denied", Basis: prof + "; selected Group only"},
		{Name: "r3 ownerB cannot create a profile in A from B", Context: org("ownerB", "B"), Op: "insert", Table: "business_profiles", Org: "A", Expect: "denied", Basis: prof + "; Organization boundary"},
		{Name: "r3 ownerA attaches unlinked pAUnlinked to A4", Context: org("ownerA", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pAUnlinked", Expect: "denied", Basis: assoc + "; G060 r0 A13: Organization owner without the override"},
		{Name: "r3 adminA attaches pA3 to A4", Context: org("adminA", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pA3", Expect: "denied", Basis: assoc + "; G060 r0 A13: Organization admin without the override"},
		{Name: "r3 ownerA attaches unlinked pAUnlinked to A4, override on", Context: ovr("ownerA", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pAUnlinked", Expect: "ok", Basis: assoc + "; G060 r0 A13: Organization owner with the override"},
		{Name: "r3 adminA attaches pA3 to A4, override on", Context: ovr("adminA", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pA3", Expect: "ok", Basis: assoc + "; G060 r0 A13: Organization admin with the override"},
		{Name: "r3 ownerA4 attaches a profile it inserted in the same transaction", Context: org("ownerA4", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "new", Expect: "ok", Basis: assoc + "; new-profile branch"},
		{Name: "r3 ownerA4 in selected A4 attaches a profile it inserted in the same transaction", Context: sel("ownerA4", "A", "A4"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "new", Expect: "ok", Basis: assoc + "; new-profile branch"},
		{Name: "r3 ownerA4 cannot attach the unrelated shared pA12 (A4 alone qualifies nothing)", Context: org("ownerA4", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pA12", Expect: "denied", Basis: assoc + "; pre-existing eligibility"},
		{Name: "r3 ownerA4 cannot attach the unlinked pAUnlinked", Context: org("ownerA4", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pAUnlinked", Expect: "denied", Basis: assoc + "; unlinked profile"},
		{Name: "r3 ownerA4 cannot attach the foreign pB1", Context: org("ownerA4", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pB1", Expect: "denied", Basis: assoc + "; Organization boundary"},
		{Name: "r3 gOwnerA1 cannot attach its pA12 to A4 where it holds no role", Context: org("gOwnerA1", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pA12", Expect: "denied", Basis: assoc + "; target Group owner/admin"},
		{Name: "r3 memberA2 (group_member) cannot attach pA12 to A4", Context: org("memberA2", "A"), Op: "insert", Table: "group_business_profiles", Org: "A", Group: "A4", Profile: "pA12", Expect: "denied", Basis: assoc + "; member-only"},
		{Name: "r3 ownerA cannot delete the association gbpA1", Context: org("ownerA", "A"), Op: "delete", Table: "group_business_profiles", Row: "gbpA1", Expect: "denied", Basis: "Plan r5 §5.5: no association DELETE"},
		{Name: "r3 gOwnerA1 cannot delete the association gbpA1", Context: org("gOwnerA1", "A"), Op: "delete", Table: "group_business_profiles", Row: "gbpA1", Expect: "denied", Basis: "Plan r5 §5.5: no association DELETE"},
	}
}

// --- oracle r4 (G019): role_bindings UPDATE and DELETE under the version-15 management policy and owner guard ---------
//
// REVISION r4 (2026-10-18, G019 Plan §6.2 'TestRLSIsolation oracle r4', scope line 34; review disposition
// 'MEM-1 PASS; complete the planned independent RLS oracle r4'). Reason: version 15 replaces the lock-only
// role_bindings policy with real UPDATE/DELETE policies and an owner guard, so the runtime role may now change and
// remove membership bindings. SOURCE (the only semantic input; SHA-256 verified before writing):
//   G019-rls-oracle-r4-expectations.md  df91a0e87a2d9a94b470879176186988b8f95c26d994b422ded2e06d8ecd9c60
//   (a source-only matrix derived from the G019 Goal/Spec R1-R3, AUTHORIZATION and the accepted Plan §4.4-4.5/§6.2 by a
//   reviewer who read no implementation, test, Result or runtime output). Basis strings cite its sections.
// Every expectation below is a semantic literal from that matrix: the target role, whether the row changes, and the
// named protection that refuses it. Exact SQLSTATE and zero-rows-versus-WITH-CHECK mechanics stay in the driver.
// Changes to r0-r3 above (recorded inline, nothing deleted silently):
//   moved  the five raw self-demotions the matrix admits (adminA org-only / selected A4: org_admin -> org_member; adminB:
//          org_admin -> org_member; adminA2 and crossUser in B: group_admin -> group_member) leave the "no_update" lists
//          and become explicit positive readback cases here (their every other r0-r2 row, visible set and Basis is kept);
//   named  the four sole-owner self-demotions that stay refused (bOwnerA x2, bGOwnerA1G1, bCrossA3) carry Guard
//          "last_owner": the driver must see that named protection, never an unnamed error of the same class;
//   kept   every structural/immutable, foreign-tenant, selected-sibling and privilege row of r0-r3, the RLS-IMMUT lock
//          positives and its mutation negatives (its gOwnerA1 selected-A1 group_admin self-demotion is now the named
//          last_owner refusal), and the r3 trigger-only owner INSERT refusals.
// Expect values: "ok" (exactly one row changed; the readback shows the target role, or no row after DELETE) or
// "denied" (no row changes). Guard names the owner-guard protection of a denial; "" means the scope/management policy
// itself refuses (a hidden row or the managerial WITH CHECK), with no named constraint. Pre lists same-actor
// same-transaction admitted role changes that precede the attempt (the atomic owner transfer). Every case starts from
// the unmodified fixture in its own rolled-back transaction.
// MAPPED, NOT DUPLICATED (existing attributable coverage of the same boundary; the fixture is unchanged because any new
// row would change every visible set above): a disabled promotion recipient and the post-disable wait (tenancy
// TestTenancyModel 'MEM-1 inactive owner and the post-disable wait refuse'); a Group-ownership recipient without
// current Organization membership (no such subject exists in this fixture; facade coverage in the membership suites);
// 23000/42501 constraint-name to Code mapping (tenancy membership_test, access membership_test); competing owner
// changes keeping at least one owner, class-40 aborts and whole-transaction rollback (ordering O-MEM-1, O-MEM-3b,
// O-MEM-3d); immutable keys/source and the trigger-only owner INSERT (r0-r3 rows and RLS-IMMUT above); facade
// self_target refusals (membership suites; raw self rules differ by design, r4 §Self operations).

type rlsPreStep struct {
	Row  string // role_bindings label
	Role string // the admitted target role
}

type rlsMutation struct {
	Name    string
	Context rlsContext
	Pre     []rlsPreStep // admitted same-actor role changes before the attempt (atomic transfer), in order
	Op      string       // "update" or "delete"
	Row     string       // the role_bindings label the attempt targets
	Role    string       // update: the target role
	Expect  string       // "ok" or "denied"
	Guard   string       // denied: "owner_authority", "recipient_ineligible", "last_owner", "dependent_bindings", "context_required", "self_escalation" (G060), "owner_subject" (G060 amendment01), or "" (policy)
	Also    []string     // G060 amendment01: "ok" DELETE of an Organization binding: the subject's Group bindings the same statement removes
	Basis   string
}

func rlsExpectedMutations() []rlsMutation {
	org := func(a, o string) rlsContext { return rlsContext{Actor: a, Org: o, Shape: "valid"} }
	sel := func(a, o, g string) rlsContext { return rlsContext{Actor: a, Org: o, Group: g, Shape: "valid"} }
	// G060: the same contexts with the Organization's Group-management override set true by its owner (driver toggle).
	orgOn := func(a, o string) rlsContext { return rlsContext{Actor: a, Org: o, Shape: "valid", Override: true} }
	selOn := func(a, o, g string) rlsContext {
		return rlsContext{Actor: a, Org: o, Group: g, Shape: "valid", Override: true}
	}
	const a01 = "G060 S2 amendment01 A1.3/A1.5 (Codex ACCEPT): "

	const (
		auth   = "r4 §Authority: "
		self   = "r4 §Self operations and owner invariants: "
		ctxsel = "r4 §Selected-context interpretation, Spec R1: "
		proof  = "r4 §Required proof: "
	)
	return []rlsMutation{
		// ===== ordinary management (Manage = admin/member role changes and removal) =====
		{Name: "r4 ownerA org-only demotes adminA to org_member", Context: org("ownerA", "A"), Op: "update", Row: "bAdminA", Role: "org_member", Expect: "ok", Basis: auth + "Organization owner manages ordinary Organization bindings"},
		{Name: "r4 ownerA org-only promotes memberA to org_admin", Context: org("ownerA", "A"), Op: "update", Row: "bMemberA", Role: "org_admin", Expect: "ok", Basis: auth + "Organization owner manages ordinary Organization bindings"},
		{Name: "r4 adminA org-only promotes memberA to org_admin", Context: org("adminA", "A"), Op: "update", Row: "bMemberA", Role: "org_admin", Expect: "ok", Basis: auth + "Organization admin manages ordinary Organization bindings"},
		{Name: "r4 adminA selected A4 promotes memberA to org_admin (an Organization-role update is not a sibling-Group operation)", Context: sel("adminA", "A", "A4"), Op: "update", Row: "bMemberA", Role: "org_admin", Expect: "ok", Basis: ctxsel + "an ordinary Organization-role update is not a sibling-Group operation, unlike departure"},
		{Name: "r4 ownerA org-only promotes memberA2 to group_admin of A2", Context: org("ownerA", "A"), Op: "update", Row: "bMemberA2G2", Role: "group_admin", Expect: "denied", Guard: "", Basis: "G060 (ADR-008; Plan r1 C3, correction02 C02-2 and the Codex mechanical correction): without the Group-management override an Organization owner without a Group role changes no Group role (managerial WITH CHECK)"},
		{Name: "r4 adminA org-only demotes adminA2 to group_member of A2", Context: org("adminA", "A"), Op: "update", Row: "bAdminA2G2", Role: "group_member", Expect: "denied", Guard: "", Basis: "G060 (ADR-008; Plan r1 C3, correction02 C02-2 and the Codex mechanical correction): without the override an Organization admin without a Group role changes no Group role"},
		{Name: "r4 ownerA selected A2 promotes memberA2 to group_admin", Context: sel("ownerA", "A", "A2"), Op: "update", Row: "bMemberA2G2", Role: "group_admin", Expect: "denied", Guard: "", Basis: "G060 (ADR-008; Plan r1 C3, correction02 C02-2 and the Codex mechanical correction): selecting the Group confers no administration without the override"},
		{Name: "r4 adminA2 org-only (Group admin of A2) promotes memberA2 to group_admin", Context: org("adminA2", "A"), Op: "update", Row: "bMemberA2G2", Role: "group_admin", Expect: "ok", Basis: auth + "Target Group admin, Organization member: Manage target Group"},
		{Name: "r4 adminA2 selected A2 promotes memberA2 to group_admin", Context: sel("adminA2", "A", "A2"), Op: "update", Row: "bMemberA2G2", Role: "group_admin", Expect: "ok", Basis: auth + "Target Group admin manages the selected target Group"},
		// ===== admitted raw self-demotions (moved from the r0-r2 "no_update" lists) =====
		{Name: "r4 adminA org-only demotes itself to org_member", Context: org("adminA", "A"), Op: "update", Row: "bAdminA", Role: "org_member", Expect: "ok", Basis: self + "raw Organization admin to member self-demotion is permitted (facade self_target is separate); moved from 'adminA org-only'"},
		{Name: "r4 adminA selected A4 demotes itself to org_member", Context: sel("adminA", "A", "A4"), Op: "update", Row: "bAdminA", Role: "org_member", Expect: "ok", Basis: self + "raw Organization admin to member self-demotion; moved from 'adminA selected A4'"},
		{Name: "r4 adminB org-only demotes itself to org_member", Context: org("adminB", "B"), Op: "update", Row: "bAdminB", Role: "org_member", Expect: "ok", Basis: self + "raw Organization admin to member self-demotion; moved from 'adminB org-only'"},
		{Name: "r4 adminA2 org-only demotes itself to group_member of A2", Context: org("adminA2", "A"), Op: "update", Row: "bAdminA2G2", Role: "group_member", Expect: "denied", Guard: "self_escalation", Basis: "G060 (ADR-008; Plan r1 C3, correction02 C02-2 and the Codex mechanical correction): guard step (3b): a runtime session changes the role of its own Group binding only as that Group's owner (supersedes the r4 raw Group-admin self-demotion; Leave is unchanged)"},
		{Name: "r4 crossUser in B org-only demotes itself to group_member of B1", Context: org("crossUser", "B"), Op: "update", Row: "bCrossB1", Role: "group_member", Expect: "denied", Guard: "self_escalation", Basis: "G060 (ADR-008; Plan r1 C3, correction02 C02-2 and the Codex mechanical correction): guard step (3b) refuses an own Group role change without that Group's owner role"},
		{Name: "r4 adminA org-only cannot promote its own A1 member binding to group_admin (override off)", Context: org("adminA", "A"), Op: "update", Row: "bAdminAG1", Role: "group_admin", Expect: "denied", Guard: "self_escalation", Basis: "G060 correction02 C02-2 and the Codex mechanical correction: the own-row USING admits the row and guard step (3b) refuses before WITH CHECK, override off"},
		{Name: "r4 adminA org-only cannot promote its own A1 member binding to group_admin, override on", Context: orgOn("adminA", "A"), Op: "update", Row: "bAdminAG1", Role: "group_admin", Expect: "denied", Guard: "self_escalation", Basis: "G060 correction02 C02-2: the override path never raises the actor's own Group role (step (3b); WITH CHECK subject_user_id <> ctx_user)"},
		// ===== owner transitions by the exact scope's current owner =====
		{Name: "r4 ownerA org-only promotes adminA to a second org_owner", Context: org("ownerA", "A"), Op: "update", Row: "bAdminA", Role: "org_owner", Expect: "ok", Basis: self + "promotion requires an enabled existing same-scope admin; multiple owners are valid"},
		{Name: "r4 crossUser (owner of A3) promotes gOwnerA1 to a second group_owner of A3 after making it admin", Context: org("crossUser", "A"), Pre: []rlsPreStep{{"bGOwnerA1G3", "group_admin"}}, Op: "update", Row: "bGOwnerA1G3", Role: "group_owner", Expect: "ok", Basis: self + "Group-owner transition by the exact scope's current owner; recipient is a same-scope admin and a current Organization member; multiple owners are valid"},
		{Name: "r4 crossUser transfers A3: promotes gOwnerA1 first, then demotes itself to group_admin", Context: org("crossUser", "A"), Pre: []rlsPreStep{{"bGOwnerA1G3", "group_admin"}, {"bGOwnerA1G3", "group_owner"}}, Op: "update", Row: "bCrossA3", Role: "group_admin", Expect: "ok", Basis: self + "atomic transfer promotes the recipient before demoting the former owner"},
		{Name: "r4 ownerA transfers A: promotes adminA first, then demotes itself to org_admin", Context: org("ownerA", "A"), Pre: []rlsPreStep{{"bAdminA", "org_owner"}}, Op: "update", Row: "bOwnerA", Role: "org_admin", Expect: "ok", Basis: self + "atomic transfer promotes the recipient before demoting the former owner"},
		{Name: "r4 ownerA transfers A: promotes adminA first, then leaves A", Context: org("ownerA", "A"), Pre: []rlsPreStep{{"bAdminA", "org_owner"}}, Op: "delete", Row: "bOwnerA", Expect: "ok", Basis: self + "atomic transfer promotes the recipient before removing the former owner; departure in Organization-only context with no dependent Group binding"},
		// ===== removal and departure =====
		{Name: "r4 ownerA org-only removes memberA", Context: org("ownerA", "A"), Op: "delete", Row: "bMemberA", Expect: "ok", Basis: auth + "Manage permits removal of an ordinary Organization binding"},
		{Name: "r4 adminA org-only cannot remove memberA2 from A2 (override off)", Context: org("adminA", "A"), Op: "delete", Row: "bMemberA2G2", Expect: "denied", Guard: "", Basis: a01 + "the D-13 residual removed: without the override an Organization admin without a Group role neither sees nor deletes an ordinary Group binding at depth zero"},
		{Name: "r4 adminA org-only removes memberA2 from A2, override on", Context: orgOn("adminA", "A"), Op: "delete", Row: "bMemberA2G2", Expect: "ok", Basis: auth + "Organization admin removes an ordinary Group binding within a permitted Group; " + a01 + "admitted only with the override"},
		{Name: "r4 adminA2 org-only removes memberA2 from A2", Context: org("adminA2", "A"), Op: "delete", Row: "bMemberA2G2", Expect: "ok", Basis: auth + "Target Group admin manages (removes from) the target Group"},
		{Name: "r4 ownerA selected A2 cannot remove memberA2 from A2 (override off)", Context: sel("ownerA", "A", "A2"), Op: "delete", Row: "bMemberA2G2", Expect: "denied", Guard: "", Basis: a01 + "selecting the Group confers no roster or removal without the override"},
		{Name: "r4 ownerA selected A2 removes memberA2 from A2, override on", Context: selOn("ownerA", "A", "A2"), Op: "delete", Row: "bMemberA2G2", Expect: "ok", Basis: auth + "removal within the permitted (selected) Group; " + a01 + "admitted only with the override"},
		{Name: "r4 memberA org-only leaves A (no dependent Group binding)", Context: org("memberA", "A"), Op: "delete", Row: "bMemberA", Expect: "ok", Basis: "Spec R1 'Leave one's own membership' (current member acting on self); " + self + "departure requires Organization-only context and dependent Group cleanup, both satisfied"},
		{Name: "r4 gOwnerA1 org-only leaves A3 where it is a member", Context: org("gOwnerA1", "A"), Op: "delete", Row: "bGOwnerA1G3", Expect: "ok", Basis: "Spec R1 'Leave one's own membership' Group scope, subject to last-owner protection (not the owner of A3)"},
		// ===== self-escalation negatives =====
		{Name: "r4 memberA org-only cannot promote itself to org_admin", Context: org("memberA", "A"), Op: "update", Row: "bMemberA", Role: "org_admin", Expect: "denied", Guard: "", Basis: self + "member-to-admin self-promotion remains denied (managerial WITH CHECK controls UPDATE)"},
		{Name: "r4 memberA2 org-only cannot promote itself to group_admin of A2", Context: org("memberA2", "A"), Op: "update", Row: "bMemberA2G2", Role: "group_admin", Expect: "denied", Guard: "self_escalation", Basis: self + "member-to-admin self-promotion remains denied; " + "G060 (ADR-008; Plan r1 C3, correction02 C02-2 and the Codex mechanical correction): the own-row USING admits the row and the BEFORE UPDATE guard step (3b) refuses before WITH CHECK (override on or off)"},
		{Name: "r4 adminA org-only cannot promote itself to org_owner", Context: org("adminA", "A"), Op: "update", Row: "bAdminA", Role: "org_owner", Expect: "denied", Guard: "owner_authority", Basis: self + "admin-to-owner without current same-scope owner authority is denied; " + proof + "named owner_authority"},
		{Name: "r4 adminB org-only cannot promote itself to org_owner", Context: org("adminB", "B"), Op: "update", Row: "bAdminB", Role: "org_owner", Expect: "denied", Guard: "owner_authority", Basis: self + "admin-to-owner without owner authority"},
		{Name: "r4 adminA2 org-only cannot promote itself to group_owner of A2", Context: org("adminA2", "A"), Op: "update", Row: "bAdminA2G2", Role: "group_owner", Expect: "denied", Guard: "owner_authority", Basis: auth + "Target Group admin: Group owner transition Deny; " + self + "admin-to-owner without owner authority"},
		{Name: "r4 crossUser in B org-only cannot promote itself to group_owner of B1", Context: org("crossUser", "B"), Op: "update", Row: "bCrossB1", Role: "group_owner", Expect: "denied", Guard: "owner_authority", Basis: self + "admin-to-owner without owner authority"},
		// ===== Organization management never overrides Group-owner authority =====
		{Name: "r4 ownerA org-only cannot demote the owner of A1 (no A1 owner role)", Context: org("ownerA", "A"), Op: "update", Row: "bGOwnerA1G1", Role: "group_admin", Expect: "denied", Guard: "", Basis: auth + "Organization owner: Group owner transition requires the target Group owner role; " + self + "Organization-wide ordinary management never overrides Group-owner authority" + "; " + a01 + "without the override the Group row is outside the USING set (mechanism only)"},
		{Name: "r4 ownerA org-only cannot demote the owner of A1 (no A1 owner role), override on", Context: orgOn("ownerA", "A"), Op: "update", Row: "bGOwnerA1G1", Role: "group_admin", Expect: "denied", Guard: "owner_authority", Basis: auth + "Organization owner: Group owner transition requires the target Group owner role; " + self + "Organization-wide ordinary management never overrides Group-owner authority" + "; " + a01 + "with the override the row is visible and the guard's owner authority refuses"},
		{Name: "r4 ownerA org-only cannot promote adminA2 to group_owner of A2", Context: org("ownerA", "A"), Op: "update", Row: "bAdminA2G2", Role: "group_owner", Expect: "denied", Guard: "", Basis: auth + "Group owner transition requires the target Group owner role" + "; " + a01 + "without the override the Group row is outside the USING set (mechanism only)"},
		{Name: "r4 ownerA org-only cannot promote adminA2 to group_owner of A2, override on", Context: orgOn("ownerA", "A"), Op: "update", Row: "bAdminA2G2", Role: "group_owner", Expect: "denied", Guard: "owner_authority", Basis: auth + "Group owner transition requires the target Group owner role" + "; " + a01 + "with the override the row is visible and the guard's owner authority refuses"},
		{Name: "r4 ownerA selected A1 cannot demote the owner of A1", Context: sel("ownerA", "A", "A1"), Op: "update", Row: "bGOwnerA1G1", Role: "group_admin", Expect: "denied", Guard: "", Basis: auth + "selecting the Group confers no Group-owner authority" + "; " + a01 + "without the override the Group row is outside the USING set (mechanism only)"},
		{Name: "r4 ownerA selected A1 cannot demote the owner of A1, override on", Context: selOn("ownerA", "A", "A1"), Op: "update", Row: "bGOwnerA1G1", Role: "group_admin", Expect: "denied", Guard: "owner_authority", Basis: auth + "selecting the Group confers no Group-owner authority" + "; " + a01 + "with the override the row is visible and the guard's owner authority refuses"},
		{Name: "r4 adminA org-only cannot demote the owner of A4", Context: org("adminA", "A"), Op: "update", Row: "bOwnerA4G4", Role: "group_admin", Expect: "denied", Guard: "", Basis: auth + "Organization admin: Group owner transition requires the target Group owner role" + "; " + a01 + "without the override the Group row is outside the USING set (mechanism only)"},
		{Name: "r4 adminA org-only cannot demote the owner of A4, override on", Context: orgOn("adminA", "A"), Op: "update", Row: "bOwnerA4G4", Role: "group_admin", Expect: "denied", Guard: "owner_authority", Basis: auth + "Organization admin: Group owner transition requires the target Group owner role" + "; " + a01 + "with the override the row is visible and the guard's owner authority refuses"},
		{Name: "r4 adminA org-only cannot remove the owner of A4", Context: org("adminA", "A"), Op: "delete", Row: "bOwnerA4G4", Expect: "denied", Guard: "", Basis: auth + "Manage does not permit touching an owner; " + proof + "cover DELETE" + "; " + a01 + "without the override the Group row is outside the USING set (mechanism only)"},
		{Name: "r4 adminA org-only cannot remove the owner of A4, override on", Context: orgOn("adminA", "A"), Op: "delete", Row: "bOwnerA4G4", Expect: "denied", Guard: "owner_authority", Basis: auth + "Manage does not permit touching an owner; " + proof + "cover DELETE" + "; " + a01 + "with the override the row is visible and the guard's owner authority refuses"},
		// ===== Organization admin and the Organization owner =====
		{Name: "r4 adminA org-only cannot demote ownerA", Context: org("adminA", "A"), Op: "update", Row: "bOwnerA", Role: "org_admin", Expect: "denied", Guard: "owner_authority", Basis: auth + "Organization admin: Organization owner transition Deny"},
		{Name: "r4 adminA org-only cannot remove ownerA", Context: org("adminA", "A"), Op: "delete", Row: "bOwnerA", Expect: "denied", Guard: "owner_authority", Basis: auth + "Organization admin: Organization owner transition Deny (removal)"},
		{Name: "r4 adminA org-only cannot promote memberA to org_owner", Context: org("adminA", "A"), Op: "update", Row: "bMemberA", Role: "org_owner", Expect: "denied", Guard: "owner_authority", Basis: auth + "Organization admin: Organization owner transition Deny"},
		// ===== promotion recipients =====
		{Name: "r4 ownerA org-only cannot promote memberA (not an admin) straight to org_owner", Context: org("ownerA", "A"), Op: "update", Row: "bMemberA", Role: "org_owner", Expect: "denied", Guard: "recipient_ineligible", Basis: self + "promotion requires an enabled existing same-scope admin; " + proof + "non-admin recipient, named recipient_ineligible"},
		{Name: "r4 crossUser (owner of A3) cannot promote gOwnerA1's member binding straight to group_owner", Context: org("crossUser", "A"), Op: "update", Row: "bGOwnerA1G3", Role: "group_owner", Expect: "denied", Guard: "recipient_ineligible", Basis: self + "promotion requires an existing same-scope admin"},
		// ===== last owner =====
		{Name: "r4 ownerA org-only cannot demote itself as the sole owner of A", Context: org("ownerA", "A"), Op: "update", Row: "bOwnerA", Role: "org_admin", Expect: "denied", Guard: "last_owner", Basis: self + "demotion of the last assigned owner is refused; " + proof + "named last_owner"},
		{Name: "r4 ownerA org-only cannot leave A as its sole owner", Context: org("ownerA", "A"), Op: "delete", Row: "bOwnerA", Expect: "denied", Guard: "last_owner", Basis: self + "deletion of the last assigned owner is refused"},
		{Name: "r4 ownerB org-only cannot demote itself as the sole owner of B", Context: org("ownerB", "B"), Op: "update", Row: "bOwnerB", Role: "org_admin", Expect: "denied", Guard: "last_owner", Basis: self + "demotion of the last assigned owner is refused"},
		{Name: "r4 gOwnerA1 selected A1 cannot demote itself as the sole owner of A1", Context: sel("gOwnerA1", "A", "A1"), Op: "update", Row: "bGOwnerA1G1", Role: "group_admin", Expect: "denied", Guard: "last_owner", Basis: self + "demotion of the last assigned owner is refused (also the RLS-IMMUT group_admin negative)"},
		{Name: "r4 gOwnerA1 selected A1 cannot leave A1 as its sole owner", Context: sel("gOwnerA1", "A", "A1"), Op: "delete", Row: "bGOwnerA1G1", Expect: "denied", Guard: "last_owner", Basis: "Spec R1 Group leave subject to last-owner protection"},
		{Name: "r4 crossUser in A org-only cannot demote itself as the sole owner of A3", Context: org("crossUser", "A"), Op: "update", Row: "bCrossA3", Role: "group_admin", Expect: "denied", Guard: "last_owner", Basis: self + "demotion of the last assigned owner is refused"},
		// ===== Organization departure: Organization-only context and dependent Group cleanup =====
		{Name: "r4 gOwnerA1 org-only cannot leave A as the sole owner of A1", Context: org("gOwnerA1", "A"), Op: "delete", Row: "bGOwnerA1Org", Expect: "denied", Guard: "last_owner", Basis: self + "departure preserves every Group's owner; " + a01 + "the guard's cascade reaches A1's sole-owner binding and its nested step (4) refuses; the statement rolls back (override-independent)"},
		{Name: "r4 adminA2 org-only leaves A; its A2 binding departs in the same statement", Context: org("adminA2", "A"), Op: "delete", Row: "bAdminA2Org", Expect: "ok", Also: []string{"bAdminA2G2"}, Basis: self + "departure leaves no dormant Group grant; " + a01 + "raw self-departure is complete and atomic: the guard deletes the dependent Group binding first (override-independent)"},
		{Name: "r4 ownerA org-only cannot remove gOwnerA1 (owner of A1) from A", Context: org("ownerA", "A"), Op: "delete", Row: "bGOwnerA1Org", Expect: "denied", Guard: "owner_subject", Basis: self + "Group succession precedes Organization removal; " + a01 + "the guard's cascade refuses another subject's Group-owner binding (owner_subject)"},
		{Name: "r4 ownerA org-only cannot remove gOwnerA1 (owner of A1) from A, override on", Context: orgOn("ownerA", "A"), Op: "delete", Row: "bGOwnerA1Org", Expect: "denied", Guard: "owner_subject", Basis: self + "Group succession precedes Organization removal; " + a01 + "the cascade is override-independent: the override confers no Group-owner removal (owner_subject, not the superseded dependent_bindings)"},
		{Name: "r4 memberA2 selected A2 cannot leave A from a selected-Group context", Context: sel("memberA2", "A", "A2"), Op: "delete", Row: "bMemberA2Org", Expect: "denied", Guard: "context_required", Basis: self + "Organization departure requires Organization-only context; " + ctxsel + "departure expressly requires Organization-only context"},
		{Name: "r4 gOwnerA1 selected A1 cannot leave A from a selected-Group context", Context: sel("gOwnerA1", "A", "A1"), Op: "delete", Row: "bGOwnerA1Org", Expect: "denied", Guard: "context_required", Basis: self + "Organization departure requires Organization-only context"},
		// ===== foreign tenant, selected sibling and no administration of others =====
		{Name: "r4 ownerB with Organization A active cannot promote memberA", Context: org("ownerB", "A"), Op: "update", Row: "bMemberA", Role: "org_admin", Expect: "denied", Guard: "", Basis: auth + "Foreign/nonmember/invalid context: Deny"},
		{Name: "r4 ownerA org-only cannot demote adminB (Organization B row)", Context: org("ownerA", "A"), Op: "update", Row: "bAdminB", Role: "org_member", Expect: "denied", Guard: "", Basis: auth + "Manage does not permit foreign tenants"},
		{Name: "r4 ownerA org-only cannot remove memberB1 from B", Context: org("ownerA", "A"), Op: "delete", Row: "bMemberB1Org", Expect: "denied", Guard: "", Basis: auth + "Manage does not permit foreign tenants"},
		{Name: "r4 crossUser in A org-only cannot touch its own B1 admin binding from A", Context: org("crossUser", "A"), Op: "update", Row: "bCrossB1", Role: "group_member", Expect: "denied", Guard: "", Basis: self + "a cross-Organization member's active-context Group authority confers no mutation in their other Organization"},
		{Name: "r4 crossUser in B org-only cannot touch gOwnerA1's A3 binding from B", Context: org("crossUser", "B"), Op: "update", Row: "bGOwnerA1G3", Role: "group_admin", Expect: "denied", Guard: "", Basis: self + "active-context Group authority confers no mutation in the other Organization"},
		{Name: "r4 gOwnerA1 selected A1 cannot DELETE its own sibling A3 binding", Context: sel("gOwnerA1", "A", "A1"), Op: "delete", Row: "bGOwnerA1G3", Expect: "denied", Guard: "", Basis: ctxsel + "a selected Group cannot target a sibling Group, including raw self DELETE of that sibling binding"},
		{Name: "r4 gOwnerA1 selected A1 cannot update its own sibling A3 binding", Context: sel("gOwnerA1", "A", "A1"), Op: "update", Row: "bGOwnerA1G3", Role: "group_admin", Expect: "denied", Guard: "", Basis: ctxsel + "a selected Group cannot target a sibling Group"},
		{Name: "r4 ownerA selected A1 cannot promote memberA2 in sibling A2", Context: sel("ownerA", "A", "A1"), Op: "update", Row: "bMemberA2G2", Role: "group_admin", Expect: "denied", Guard: "", Basis: ctxsel + "selected-context restrictions remain for Organization owners/admins (Spec R1 table)"},
		{Name: "r4 ownerA selected A1 cannot remove memberA2 from sibling A2", Context: sel("ownerA", "A", "A1"), Op: "delete", Row: "bMemberA2G2", Expect: "denied", Guard: "", Basis: ctxsel + "selected-context restrictions remain"},
		{Name: "r4 adminA2 org-only cannot promote memberA2's Organization binding", Context: org("adminA2", "A"), Op: "update", Row: "bMemberA2Org", Role: "org_admin", Expect: "denied", Guard: "", Basis: auth + "Target Group admin, Organization member: no administration of others' Organization bindings"},
		{Name: "r4 adminA2 selected A2 cannot promote memberA's Organization binding", Context: sel("adminA2", "A", "A2"), Op: "update", Row: "bMemberA", Role: "org_admin", Expect: "denied", Guard: "", Basis: auth + "Target Group admin: no administration of others' Organization bindings"},
		{Name: "r4 memberA org-only cannot promote memberA2's Organization binding", Context: org("memberA", "A"), Op: "update", Row: "bMemberA2Org", Role: "org_admin", Expect: "denied", Guard: "", Basis: auth + "Organization member: no administration of others"},
		{Name: "r4 memberA2 org-only cannot remove adminA2 from A2", Context: org("memberA2", "A"), Op: "delete", Row: "bAdminA2G2", Expect: "denied", Guard: "", Basis: auth + "Target Group member: no administration of others"},
		{Name: "r4 gOwnerA1 org-only (member of A3) cannot demote the owner of A3", Context: org("gOwnerA1", "A"), Op: "update", Row: "bCrossA3", Role: "group_admin", Expect: "denied", Guard: "", Basis: auth + "Target Group member: no administration of others; Group ownership elsewhere confers nothing in A3"},
		{Name: "r4 outsider with Organization A active cannot promote memberA", Context: org("outsider", "A"), Op: "update", Row: "bMemberA", Role: "org_admin", Expect: "denied", Guard: "", Basis: auth + "Foreign/nonmember/invalid context: Deny"},
		{Name: "r4 memberB1 with Organization A active cannot remove memberA", Context: org("memberB1", "A"), Op: "delete", Row: "bMemberA", Expect: "denied", Guard: "", Basis: auth + "Foreign/nonmember/invalid context: Deny"},
	}
}

// --- oracle G020: generic labels (Plan r2 §1, §1.2, §3 step 4, §5.2 T3; §8 AC3/AC5) --------------------------------------
//
// REVISION G020 (2026-10-21, G020 Plan r2 §5.2 row T3). Reason: version 20 adds app.label_definitions and
// app.label_assignments (FORCE row-level security, command-specific policies to the runtime and owner roles). Additions
// only: the label rows of rlsExpectedFixture (created by oracleWorld like every other fixture row; invisible to every
// pre-G020 relation), the LabelDefinitions/LabelAssignments sets of the existing valid cases (every other field, every
// Updates row, every Reject class and every case is unchanged; the Reject cases now also cover the two label relations),
// and the label reads, raw writes and isolated foreign-key cases below. SOURCES: the G020 Plan r2 §1 authority table and
// bullets, §1.2 (exact columns, constraints and foreign keys), §3 step 4 (exact policy predicates), §5.2 T3 and §8
// AC3/AC5. INDEPENDENCE NOTE (deviation, as for r3/r4): written by the G020 implementation sub-writer from the Plan text,
// in parallel with and without reading migration 00020; no expectation comes from a runtime result. A disagreement is
// resolved against the Plan or escalated, never by editing an expectation to match. The first refusal of a multi-check
// case (the converse-direction pins 42501 for carol and 23514 for bob, case (iv) 23514) follows the Plan §5.2 order of a
// raw INSERT in PostgreSQL 17 (RLS WITH CHECK 42501 -> CHECK/NOT NULL 23514/23502 -> unique 23505 -> referential 23503);
// if the first b2 run observes a different first-refusal code, the Result records it and that pin is corrected under
// standing authority. The isolated 23503 cases and every zero-row/zero-audit assertion are not negotiable (Plan §5.2 T3).
//
// SUBJECTS. Plan §5.2 T3 names the seed identities of infra/identity/mcp-seed.sql; this oracle maps each to the fixture
// subject holding exactly the authority the Plan relies on, with A3 playing research and A4 playing design:
//   carol = ownerA    org_owner of A; no Group role.
//   frank = adminA    org_admin of A; its only Group role is member of A1, which owns no label.
//   bob   = crossUser group_owner of A3 (research), org_member of A; group_admin of B1 in Organization B.
//   dave  = gOwnerA1  group_member of A3 (research), org_member of A; its group_owner role is in A1 (no label).
//   Dana  = ownerA4   group_owner of A4 (design), org_member of A.
//   Grace = graceA    fixture-only (rlsExpectedLabelSubjects): org_admin of A and group_admin of A3. RLS-LABEL adds her
//                     two bindings after every other subtest and removes them before it ends (they would change every
//                     visible set above); the restored state is verified.
// DERIVATION (Plan §3 step 4). O is the valid context Organization, R the actor's Organization role there, G the selected
// Group, MGR = R is org_owner/org_admin, GMGR(g) = the actor's own role in g is group_owner/group_admin. An invalid
// context (header r2: an Organization without membership, an unselectable Group) admits no tenant row.
//   label_definitions SELECT: Organization-owned rows of O to MGR (the Organization-owned branch has no selected-Group
//     filter); Group-owned rows of O whose Group is G (any Group when none is selected) to GMGR(Group), or to any actual
//     role in that Group when visibility is target_members. An Organization role, the G060 override and a selected Group
//     confer nothing on Group-owned rows; a Group role confers nothing on Organization-owned rows.
//   label_definitions INSERT/UPDATE/DELETE: the manager predicate (MGR for Organization-owned rows; GMGR with the
//     selected-Group filter for Group-owned rows). Runtime UPDATE covers only name and visibility (any other column is a
//     privilege refusal, 42501); a row outside the UPDATE/DELETE USING set is not reached (no error, zero rows);
//     label_definitions_visibility_valid refuses target_members on an Organization-owned row (23514).
//   label_assignments SELECT: rows of O whose target Group is G (any when none is selected) and whose definition the
//     actor reads, target_type 'group'. INSERT/DELETE: Organization owner facts with MGR and target_type 'group', or
//     Group owner facts with GMGR(target Group), both under the selected-Group filter; no UPDATE grant (42501).
//     label_assignments_target_shape admits only Organization owner facts on the same Group (23514);
//     label_assignments_label_fkey is total over (organization_id, label_id, owner_scope_type, owner_scope_id) (23503)
//     and RESTRICTs deleting a definition that still has assignments (23503); label_assignments_target_group_fkey
//     (23503); label_assignments_unique (23505).

// rlsExpectedLabelSubjects returns the fixture-only subjects of RLS-LABEL (Grace) and their bindings.
func rlsExpectedLabelSubjects() ([]rlsUser, []rlsBinding) {
	return []rlsUser{{Label: "graceA", Status: "active"}},
		[]rlsBinding{
			{Label: "bGraceAOrg", Org: "A", User: "graceA", ScopeType: "organization", Group: "", Role: "org_admin", Source: "membership"},
			{Label: "bGraceAA3", Org: "A", User: "graceA", ScopeType: "group", Group: "A3", Role: "group_admin", Source: "membership"},
		}
}

// rlsLabelRead is a label-only visibility case: the exact label rows the context reads. The other relations are the
// existing cases' concern (Grace's bindings would change their sets).
type rlsLabelRead struct {
	Name        string
	Context     rlsContext
	Definitions []string // exact app.label_definitions set (nil = empty)
	Assignments []string // exact app.label_assignments set (nil = empty)
	Basis       string
}

// rlsLabelWrite is one raw runtime write on a label relation. Unless Commit is set it runs, like every rlsInsert, from
// the unmodified fixture in its own rolled-back transaction under a savepoint. Commit selects the isolated protocol of
// Plan §5.2 T3: its own committed runtime transaction with the attempt under its own savepoint, followed by superuser
// counts proving the attempted key has zero rows (exactly one for "ok") and zero audit.events rows (exactly one
// label.assigned for "ok"); an "ok" row is then removed by the same subject.
type rlsLabelWrite struct {
	Name       string
	Context    rlsContext
	Op         string // "insert", "update", "delete", or "delete_with_assignments" (the facade's two-statement delete)
	Table      string // "label_definitions" or "label_assignments"
	Org        string // insert: the organization_id label
	Owner      string // insert: "" = Organization owner facts (owner_scope_id = Org); otherwise the owning Group label
	Visibility string // label_definitions insert: "" = "owner_scope"
	Def        string // label_assignments insert: label_id = this definition label, or "unknown" (a fresh UUID)
	Target     string // label_assignments insert: the target Group label (target_id = target_group_id), or "unknown"
	Row        string // update/delete: the fixture row label
	Column     string // update: "name", "visibility" (toggled) or a key column
	Expect     string // "ok" (exactly one row), "zero" (no error, no row) or "denied" (State, no row)
	State      string // denied: the pinned (first-refusal) SQLSTATE
	Constraint string // denied: the constraint name the server must report ("" for a privilege/policy refusal)
	Commit     bool   // the isolated protocol (above)
	Basis      string
}

func rlsExpectedLabelReads() []rlsLabelRead {
	org := func(a, o string) rlsContext { return rlsContext{Actor: a, Org: o, Shape: "valid"} }
	ovr := func(a, o string) rlsContext { return rlsContext{Actor: a, Org: o, Shape: "valid", Override: true} }
	sel := func(a, o, g string) rlsContext { return rlsContext{Actor: a, Org: o, Group: g, Shape: "valid"} }
	selOvr := func(a, o, g string) rlsContext {
		return rlsContext{Actor: a, Org: o, Group: g, Shape: "valid", Override: true}
	}
	const p = "Plan r2 §1 authority table and bullets; §3 step 4 label_definitions_select and label_assignments_select; §5.2 T3"
	const inv = "header r2 context validity (an invalid context admits no tenant row); "
	return []rlsLabelRead{
		{Name: "label Grace org-only: Organization-owned rows through the Organization path, research's Group-owned rows through the Group path, design's not at all",
			Context: org("graceA", "A"), Definitions: []string{"lOrgA", "lOrgA2", "lA3Own", "lA3Tm"}, Assignments: []string{"asOrgAA3", "asOrgAA4"},
			Basis: p + "; §1 'a subject who independently holds both roles (Grace) ... the paths never combine'; Organization admin confers nothing on A4"},
		{Name: "label Grace org-only with the override on: unchanged",
			Context: ovr("graceA", "A"), Definitions: []string{"lOrgA", "lOrgA2", "lA3Own", "lA3Tm"}, Assignments: []string{"asOrgAA3", "asOrgAA4"},
			Basis: p + "; OVR effect none for label.group.*"},
		{Name: "label Grace selected A3: Organization-owned definitions, the A3 assignment and A3's Group-owned rows",
			Context: sel("graceA", "A", "A3"), Definitions: []string{"lOrgA", "lOrgA2", "lA3Own", "lA3Tm"}, Assignments: []string{"asOrgAA3"},
			Basis: p + "; assignments filtered to the selected Group"},
		{Name: "label Grace selected A4: Organization-owned definitions and the A4 assignment; neither A3's nor A4's Group-owned rows",
			Context: sel("graceA", "A", "A4"), Definitions: []string{"lOrgA", "lOrgA2"}, Assignments: []string{"asOrgAA4"},
			Basis: p + "; Group-owned rows only of the selected Group and only with an actual role there"},
		{Name: "label Grace with Organization B active: no label row",
			Context: org("graceA", "B"), Basis: inv + p},
		{Name: "label carol selected A3: Organization-owned definitions and only the A3 assignment; no Group-owned row",
			Context: sel("ownerA", "A", "A3"), Definitions: []string{"lOrgA", "lOrgA2"}, Assignments: []string{"asOrgAA3"},
			Basis: p + "; a Group-owned definition needs an explicit role in its Group"},
		{Name: "label carol selected A3 with the override on: still no Group-owned row",
			Context: selOvr("ownerA", "A", "A3"), Definitions: []string{"lOrgA", "lOrgA2"}, Assignments: []string{"asOrgAA3"},
			Basis: p + "; OVR effect none (Goal: the G060 override does not grant object-label access)"},
		{Name: "label frank selected A3 with the override off: Organization-owned rows only",
			Context: sel("adminA", "A", "A3"), Definitions: []string{"lOrgA", "lOrgA2"}, Assignments: []string{"asOrgAA3"},
			Basis: p},
		{Name: "label frank selected A3 with the override on: Organization-owned rows only",
			Context: selOvr("adminA", "A", "A3"), Definitions: []string{"lOrgA", "lOrgA2"}, Assignments: []string{"asOrgAA3"},
			Basis: p + "; OVR effect none"},
		{Name: "label bob selected A3: research's Group-owned rows; no Organization-owned row or assignment",
			Context: sel("crossUser", "A", "A3"), Definitions: []string{"lA3Own", "lA3Tm"},
			Basis: p + "; a Group role never reads Organization-owned labels (owner clarification 2026-10-18)"},
		{Name: "label bob selected A4 (unselectable): no label row",
			Context: sel("crossUser", "A", "A4"), Basis: inv + p},
		{Name: "label bob org-only with the override on: research's Group-owned rows, unchanged",
			Context: ovr("crossUser", "A"), Definitions: []string{"lA3Own", "lA3Tm"}, Basis: p + "; OVR effect none"},
		{Name: "label Dana selected A4: design's Group-owned rows only",
			Context: sel("ownerA4", "A", "A4"), Definitions: []string{"lA4Own", "lA4Tm"}, Basis: p},
		{Name: "label Dana selected A3 (unselectable): no label row",
			Context: sel("ownerA4", "A", "A3"), Basis: inv + p},
		{Name: "label dave org-only with the override on: only research's target_members definition",
			Context: ovr("gOwnerA1", "A"), Definitions: []string{"lA3Tm"}, Basis: p + "; members read only target_members definitions; OVR effect none"},
		{Name: "label dave selected A3 with the override on: only research's target_members definition",
			Context: selOvr("gOwnerA1", "A", "A3"), Definitions: []string{"lA3Tm"}, Basis: p + "; OVR effect none"},
	}
}

func rlsExpectedLabelWrites() []rlsLabelWrite {
	org := func(a, o string) rlsContext { return rlsContext{Actor: a, Org: o, Shape: "valid"} }
	ovr := func(a, o string) rlsContext { return rlsContext{Actor: a, Org: o, Shape: "valid", Override: true} }
	sel := func(a, o, g string) rlsContext { return rlsContext{Actor: a, Org: o, Group: g, Shape: "valid"} }
	const (
		dpol = "Plan r2 §3 step 4 label_definitions_insert/_update/_delete (the manager predicate); §1 authority table"
		apol = "Plan r2 §3 step 4 label_assignments_insert/_delete; §1 authority table"
		ddl  = "Plan r2 §1.2 and §3 steps 3-4 (grants, CHECKs and foreign keys)"
		t3   = "Plan r2 §5.2 T3 isolated cases"
	)
	return []rlsLabelWrite{
		// ===== label_definitions INSERT =====
		{Name: "label carol inserts an Organization-owned definition", Context: org("ownerA", "A"), Op: "insert", Table: "label_definitions", Org: "A", Expect: "ok", Basis: dpol},
		{Name: "label frank inserts an Organization-owned definition", Context: org("adminA", "A"), Op: "insert", Table: "label_definitions", Org: "A", Expect: "ok", Basis: dpol},
		{Name: "label Grace inserts an Organization-owned definition", Context: org("graceA", "A"), Op: "insert", Table: "label_definitions", Org: "A", Expect: "ok", Basis: dpol},
		{Name: "label carol selected A3 inserts an Organization-owned definition (no selected-Group filter on that branch)", Context: sel("ownerA", "A", "A3"), Op: "insert", Table: "label_definitions", Org: "A", Expect: "ok",
			Basis: dpol + "; the selected-context rule for Organization-owned operations is core's (§1.3), not the policy's"},
		{Name: "label bob cannot insert an Organization-owned definition", Context: org("crossUser", "A"), Op: "insert", Table: "label_definitions", Org: "A", Expect: "denied", State: "42501", Basis: dpol},
		{Name: "label dave cannot insert an Organization-owned definition", Context: org("gOwnerA1", "A"), Op: "insert", Table: "label_definitions", Org: "A", Expect: "denied", State: "42501", Basis: dpol},
		{Name: "label Dana cannot insert an Organization-owned definition", Context: org("ownerA4", "A"), Op: "insert", Table: "label_definitions", Org: "A", Expect: "denied", State: "42501", Basis: dpol},
		{Name: "label ownerB with B active cannot insert a definition of Organization A", Context: org("ownerB", "B"), Op: "insert", Table: "label_definitions", Org: "A", Expect: "denied", State: "42501", Basis: dpol + "; organization_id = the context Organization"},
		{Name: "label carol cannot insert an Organization-owned target_members definition", Context: org("ownerA", "A"), Op: "insert", Table: "label_definitions", Org: "A", Visibility: "target_members",
			Expect: "denied", State: "23514", Constraint: "label_definitions_visibility_valid", Basis: ddl + "; the policy admits carol, the CHECK refuses (Spec §2)"},
		{Name: "label bob inserts a research-owned definition", Context: org("crossUser", "A"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Expect: "ok", Basis: dpol},
		{Name: "label bob selected A3 inserts a research-owned definition", Context: sel("crossUser", "A", "A3"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Expect: "ok", Basis: dpol},
		{Name: "label Grace inserts a research-owned definition through her research admin role", Context: org("graceA", "A"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Expect: "ok", Basis: dpol},
		{Name: "label Grace selected A4 cannot insert a research-owned definition", Context: sel("graceA", "A", "A4"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Expect: "denied", State: "42501", Basis: dpol + "; selected-Group filter"},
		{Name: "label dave cannot insert a research-owned target_members definition", Context: org("gOwnerA1", "A"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Visibility: "target_members", Expect: "denied", State: "42501", Basis: dpol + "; members read only"},
		{Name: "label carol cannot insert a research-owned definition", Context: org("ownerA", "A"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Expect: "denied", State: "42501", Basis: dpol + "; orgAdmin confers nothing on Group content"},
		{Name: "label carol with the override on cannot insert a research-owned definition", Context: ovr("ownerA", "A"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Expect: "denied", State: "42501", Basis: dpol + "; OVR effect none"},
		{Name: "label frank with the override on cannot insert a research-owned definition", Context: ovr("adminA", "A"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Expect: "denied", State: "42501", Basis: dpol + "; OVR effect none"},
		{Name: "label Dana cannot insert a research-owned definition", Context: org("ownerA4", "A"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A3", Expect: "denied", State: "42501", Basis: dpol + "; a role in another Group confers nothing"},
		{Name: "label Dana inserts a design-owned definition", Context: org("ownerA4", "A"), Op: "insert", Table: "label_definitions", Org: "A", Owner: "A4", Expect: "ok", Basis: dpol},

		// ===== label_definitions UPDATE =====
		{Name: "label carol renames lOrgA", Context: org("ownerA", "A"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "name", Expect: "ok", Basis: dpol},
		{Name: "label frank renames lOrgA", Context: org("adminA", "A"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "name", Expect: "ok", Basis: dpol},
		{Name: "label Grace renames lOrgA", Context: org("graceA", "A"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "name", Expect: "ok", Basis: dpol},
		{Name: "label bob does not reach lOrgA", Context: org("crossUser", "A"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "name", Expect: "zero", Basis: dpol},
		{Name: "label dave does not reach lOrgA", Context: org("gOwnerA1", "A"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "name", Expect: "zero", Basis: dpol},
		{Name: "label Dana does not reach lOrgA", Context: org("ownerA4", "A"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "name", Expect: "zero", Basis: dpol},
		{Name: "label ownerB with B active does not reach lOrgA", Context: org("ownerB", "B"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "name", Expect: "zero", Basis: dpol + "; foreign Organization"},
		{Name: "label carol cannot make lOrgA target_members", Context: org("ownerA", "A"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "visibility",
			Expect: "denied", State: "23514", Constraint: "label_definitions_visibility_valid", Basis: ddl + "; the policy admits carol, the CHECK refuses"},
		{Name: "label carol cannot move lOrgA to Organization B", Context: org("ownerA", "A"), Op: "update", Table: "label_definitions", Row: "lOrgA", Column: "organization_id", Expect: "denied", State: "42501", Basis: ddl + "; UPDATE grant covers name and visibility only"},
		{Name: "label bob renames lA3Own", Context: org("crossUser", "A"), Op: "update", Table: "label_definitions", Row: "lA3Own", Column: "name", Expect: "ok", Basis: dpol},
		{Name: "label bob changes the visibility of lA3Own", Context: org("crossUser", "A"), Op: "update", Table: "label_definitions", Row: "lA3Own", Column: "visibility", Expect: "ok", Basis: dpol + "; Group-owned rows may use target_members"},
		{Name: "label Grace renames lA3Own", Context: org("graceA", "A"), Op: "update", Table: "label_definitions", Row: "lA3Own", Column: "name", Expect: "ok", Basis: dpol},
		{Name: "label dave reads but does not reach lA3Tm for UPDATE", Context: org("gOwnerA1", "A"), Op: "update", Table: "label_definitions", Row: "lA3Tm", Column: "name", Expect: "zero", Basis: dpol + "; members read only"},
		{Name: "label carol does not reach lA3Own", Context: org("ownerA", "A"), Op: "update", Table: "label_definitions", Row: "lA3Own", Column: "name", Expect: "zero", Basis: dpol},
		{Name: "label carol with the override on does not reach lA3Own", Context: ovr("ownerA", "A"), Op: "update", Table: "label_definitions", Row: "lA3Own", Column: "name", Expect: "zero", Basis: dpol + "; OVR effect none"},
		{Name: "label frank with the override on does not reach lA3Own", Context: ovr("adminA", "A"), Op: "update", Table: "label_definitions", Row: "lA3Own", Column: "name", Expect: "zero", Basis: dpol + "; OVR effect none"},
		{Name: "label Grace does not reach lA4Own", Context: org("graceA", "A"), Op: "update", Table: "label_definitions", Row: "lA4Own", Column: "name", Expect: "zero", Basis: dpol + "; Organization admin confers nothing on design"},
		{Name: "label Dana does not reach lA3Own", Context: org("ownerA4", "A"), Op: "update", Table: "label_definitions", Row: "lA3Own", Column: "name", Expect: "zero", Basis: dpol},
		{Name: "label bob cannot move lA3Own to A4", Context: org("crossUser", "A"), Op: "update", Table: "label_definitions", Row: "lA3Own", Column: "group_id", Expect: "denied", State: "42501", Basis: ddl + "; owner scope immutable; no UPDATE grant on group_id"},

		// ===== label_definitions DELETE =====
		{Name: "label carol deletes the unassigned lOrgA2", Context: org("ownerA", "A"), Op: "delete", Table: "label_definitions", Row: "lOrgA2", Expect: "ok", Basis: dpol},
		{Name: "label frank deletes the unassigned lOrgA2", Context: org("adminA", "A"), Op: "delete", Table: "label_definitions", Row: "lOrgA2", Expect: "ok", Basis: dpol},
		{Name: "label Grace deletes the unassigned lOrgA2", Context: org("graceA", "A"), Op: "delete", Table: "label_definitions", Row: "lOrgA2", Expect: "ok", Basis: dpol},
		{Name: "label bob does not reach lOrgA2 for DELETE", Context: org("crossUser", "A"), Op: "delete", Table: "label_definitions", Row: "lOrgA2", Expect: "zero", Basis: dpol},
		{Name: "label dave does not reach lOrgA2 for DELETE", Context: org("gOwnerA1", "A"), Op: "delete", Table: "label_definitions", Row: "lOrgA2", Expect: "zero", Basis: dpol},
		{Name: "label Dana does not reach lOrgA2 for DELETE", Context: org("ownerA4", "A"), Op: "delete", Table: "label_definitions", Row: "lOrgA2", Expect: "zero", Basis: dpol},
		{Name: "label ownerB with B active does not reach lOrgA2 for DELETE", Context: org("ownerB", "B"), Op: "delete", Table: "label_definitions", Row: "lOrgA2", Expect: "zero", Basis: dpol + "; foreign Organization"},
		{Name: "label bob deletes lA3Own", Context: org("crossUser", "A"), Op: "delete", Table: "label_definitions", Row: "lA3Own", Expect: "ok", Basis: dpol},
		{Name: "label Grace deletes lA3Own", Context: org("graceA", "A"), Op: "delete", Table: "label_definitions", Row: "lA3Own", Expect: "ok", Basis: dpol},
		{Name: "label dave does not reach lA3Tm for DELETE", Context: org("gOwnerA1", "A"), Op: "delete", Table: "label_definitions", Row: "lA3Tm", Expect: "zero", Basis: dpol + "; members read only"},
		{Name: "label carol does not reach lA3Own for DELETE", Context: org("ownerA", "A"), Op: "delete", Table: "label_definitions", Row: "lA3Own", Expect: "zero", Basis: dpol},
		{Name: "label carol with the override on does not reach lA3Own for DELETE", Context: ovr("ownerA", "A"), Op: "delete", Table: "label_definitions", Row: "lA3Own", Expect: "zero", Basis: dpol + "; OVR effect none"},
		{Name: "label Dana does not reach lA3Own for DELETE", Context: org("ownerA4", "A"), Op: "delete", Table: "label_definitions", Row: "lA3Own", Expect: "zero", Basis: dpol},
		{Name: "label Dana deletes lA4Tm", Context: org("ownerA4", "A"), Op: "delete", Table: "label_definitions", Row: "lA4Tm", Expect: "ok", Basis: dpol},

		// ===== label_assignments INSERT (production shape: Organization owner facts, target_type 'group') =====
		{Name: "label carol assigns lOrgA2 to A3", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3", Expect: "ok", Basis: apol},
		{Name: "label frank assigns lOrgA2 to A3", Context: org("adminA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3", Expect: "ok", Basis: apol},
		{Name: "label Grace assigns lOrgA2 to A3", Context: org("graceA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3", Expect: "ok", Basis: apol},
		{Name: "label carol selected A3 assigns lOrgA2 to A3", Context: sel("ownerA", "A", "A3"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3", Expect: "ok", Basis: apol},
		{Name: "label carol selected A4 cannot assign lOrgA2 to A3", Context: sel("ownerA", "A", "A4"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3", Expect: "denied", State: "42501", Basis: apol + "; selected-Group filter"},
		{Name: "label bob cannot assign the Organization-owned lOrgA2 to his A3", Context: org("crossUser", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3", Expect: "denied", State: "42501", Basis: apol + "; Organization owner facts need MGR"},
		{Name: "label dave cannot assign lOrgA2 to A3", Context: org("gOwnerA1", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3", Expect: "denied", State: "42501", Basis: apol},
		{Name: "label Dana cannot assign lOrgA2 to her A4", Context: org("ownerA4", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A4", Expect: "denied", State: "42501", Basis: apol + "; Organization owner facts need MGR"},
		{Name: "label ownerB with B active cannot assign lOrgA2 in Organization A", Context: org("ownerB", "B"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3", Expect: "denied", State: "42501", Basis: apol + "; organization_id = the context Organization"},
		{Name: "label carol cannot assign lOrgA to A3 twice", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA", Target: "A3",
			Expect: "denied", State: "23505", Constraint: "label_assignments_unique", Basis: ddl + "; label_assignments_unique (D-6 already_assigned at the facade)"},
		{Name: "label carol cannot assign lOrgA2 to an unknown Group", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "unknown",
			Expect: "denied", State: "23503", Constraint: "label_assignments_target_group_fkey", Basis: ddl + "; target existence"},
		{Name: "label carol cannot assign lOrgA2 to Organization B's Group B1", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "B1",
			Expect: "denied", State: "23503", Constraint: "label_assignments_target_group_fkey", Basis: ddl + "; the target Group must belong to the row's Organization"},

		// ===== label_assignments DELETE and UPDATE =====
		{Name: "label carol unassigns asOrgAA3", Context: org("ownerA", "A"), Op: "delete", Table: "label_assignments", Row: "asOrgAA3", Expect: "ok", Basis: apol},
		{Name: "label frank unassigns asOrgAA3", Context: org("adminA", "A"), Op: "delete", Table: "label_assignments", Row: "asOrgAA3", Expect: "ok", Basis: apol},
		{Name: "label Grace unassigns asOrgAA3", Context: org("graceA", "A"), Op: "delete", Table: "label_assignments", Row: "asOrgAA3", Expect: "ok", Basis: apol},
		{Name: "label carol selected A3 unassigns asOrgAA3", Context: sel("ownerA", "A", "A3"), Op: "delete", Table: "label_assignments", Row: "asOrgAA3", Expect: "ok", Basis: apol},
		{Name: "label carol selected A4 does not reach asOrgAA3", Context: sel("ownerA", "A", "A4"), Op: "delete", Table: "label_assignments", Row: "asOrgAA3", Expect: "zero", Basis: apol + "; selected-Group filter"},
		{Name: "label bob does not reach the Organization's label on his A3", Context: org("crossUser", "A"), Op: "delete", Table: "label_assignments", Row: "asOrgAA3", Expect: "zero", Basis: apol},
		{Name: "label dave does not reach asOrgAA3", Context: org("gOwnerA1", "A"), Op: "delete", Table: "label_assignments", Row: "asOrgAA3", Expect: "zero", Basis: apol},
		{Name: "label Dana does not reach the Organization's label on her A4", Context: org("ownerA4", "A"), Op: "delete", Table: "label_assignments", Row: "asOrgAA4", Expect: "zero", Basis: apol},
		{Name: "label ownerB with B active does not reach asOrgAA3", Context: org("ownerB", "B"), Op: "delete", Table: "label_assignments", Row: "asOrgAA3", Expect: "zero", Basis: apol + "; foreign Organization"},
		{Name: "label carol cannot retarget asOrgAA3", Context: org("ownerA", "A"), Op: "update", Table: "label_assignments", Row: "asOrgAA3", Column: "target_id", Expect: "denied", State: "42501", Basis: ddl + "; no runtime UPDATE grant on label_assignments"},

		// ===== isolated cases (Commit; Plan §5.2 T3). Valid production shape by carol Organization-wide unless stated. =====
		{Name: "label T3 (i) carol: unknown label_id -> 23503", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "unknown", Target: "A3",
			Expect: "denied", State: "23503", Constraint: "label_assignments_label_fkey", Commit: true, Basis: t3 + " (i); the policy and the CHECK admit, only the total FK refuses"},
		{Name: "label T3 (ii) carol: a foreign Organization's definition id with A's organization_id -> 23503", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgB", Target: "A3",
			Expect: "denied", State: "23503", Constraint: "label_assignments_label_fkey", Commit: true, Basis: t3 + " (ii)"},
		{Name: "label T3 (iii) carol: bob's research-owned definition id with Organization owner facts -> 23503", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lA3Own", Target: "A3",
			Expect: "denied", State: "23503", Constraint: "label_assignments_label_fkey", Commit: true, Basis: t3 + " (iii) owner-fact substitution: the definition exists with key ('group', research)"},
		{Name: "label T3 converse carol: an Organization-owned definition id with ('group', research) facts -> 42501 first", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Owner: "A3", Def: "lOrgA2", Target: "A3",
			Expect: "denied", State: "42501", Commit: true, Basis: t3 + " converse direction: the Group branch needs a research owner/admin role carol lacks; RLS precedes the CHECK (pinned first refusal)"},
		{Name: "label T3 converse bob: an Organization-owned definition id with ('group', research) facts -> 23514 first", Context: org("crossUser", "A"), Op: "insert", Table: "label_assignments", Org: "A", Owner: "A3", Def: "lOrgA2", Target: "A3",
			Expect: "denied", State: "23514", Constraint: "label_assignments_target_shape", Commit: true, Basis: t3 + " converse direction: the policy admits the research owner; the closed CHECK refuses (pinned first refusal)"},
		{Name: "label T3 (iv) bob: a research-owned definition with its exact copied facts on a group target -> 23514", Context: org("crossUser", "A"), Op: "insert", Table: "label_assignments", Org: "A", Owner: "A3", Def: "lA3Own", Target: "A3",
			Expect: "denied", State: "23514", Constraint: "label_assignments_target_shape", Commit: true, Basis: t3 + " (iv): the policy admits bob; production storage admits no Group-owned assignment"},
		{Name: "label T3 paired positive carol: exact copied owner facts insert on research (one row, one label.assigned)", Context: org("ownerA", "A"), Op: "insert", Table: "label_assignments", Org: "A", Def: "lOrgA2", Target: "A3",
			Expect: "ok", Commit: true, Basis: t3 + " paired positive"},
		{Name: "label T3 deletion cleanup carol: deleting lOrgA while assignments exist -> 23503 RESTRICT", Context: org("ownerA", "A"), Op: "delete", Table: "label_definitions", Row: "lOrgA",
			Expect: "denied", State: "23503", Constraint: "label_assignments_label_fkey", Commit: true, Basis: t3 + " deletion cleanup: no hidden orphan; the definition and both assignments remain, no audit row"},
		{Name: "label T3 deletion cleanup paired positive: carol deletes lOrgA's assignments, then lOrgA", Context: org("ownerA", "A"), Op: "delete_with_assignments", Table: "label_definitions", Row: "lOrgA",
			Expect: "ok", Basis: t3 + " deletion cleanup paired positive (§1.3 two-statement delete)"},
	}
}
