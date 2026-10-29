package access

// G009 independent ORACLE: expected outcomes for tenant object storage.
//
// Role: task G009, role oracle. Written only from the input files below,
// without reading any repository, migration 00007, objects*.go, Plan §3.2/§4.2
// or policy/SQL/decision code. Pure data plus tiny helpers; stdlib only and no
// imports. Every package-level identifier carries the "oe" prefix, except the
// two required entry points objectsExpectedFixture() and objectsExpectedCases().
//
// Inputs (SHA-256 copied from inputs/SHA256SUMS and verified with sha256sum -c):
//
//	51a487844d74e913d195c963095c2a381deeb293d1829085f2f026f060955a68  ADR-004-group-business-profile-cardinality.md
//	292d9c6861a91dc450d4a8d5ebfa6c02c93f9e4ddc8ffdaaf3bb43ab53b41527  ARCHITECTURE-s40-object-storage.md
//	4709983daddadf640a5f58b236e946c9b0db32ec60a862dbc097e5bd20159247  ARCHITECTURE-s53.7-authorization-inheritance.md
//	b5aaff584493908b44b1ed41bd3fe73477430e7b8c74753485c5e2ab0877aa72  AUTHORIZATION.md
//	73072a1086e74ed7e579efbea9dae12aefc7617ee1f639efda7a982e26880760  DATABASE-ISOLATION-context-validity.md
//	6ebc905ecd205f684c12c5b157b2b1a59edf928482e9e854712c1acbafd48c82  G009-storage-write-policy-r1.md
//	5e9f4699eb670bca33497072f0219badc69bf1d524d1be2d13c4daa7be49a348  G009-tenant-object-storage-spec.md
//	72ef7876488d845f111e97264c18ebe9b0301887f5a5619ec9bec68015d8eb6f  PLAN-oracle-case-schema-excerpt.md
//
// Ruling input (outside inputs/, verified separately with sha256sum):
//
//	a466be31437b7210bde512371560a318137e48239fd6085e943e8d90826e7502  inputs-ruling-RU05.md
//
// Revisions:
//   - r0: original oracle.
//   - r1 = gofmt of r0 (whitespace alignment only).
//   - r2: RU05 per Codex ruling a466be31 (inputs-ruling-RU05.md). New outcome
//     oeRLSImmutableGenerated, used by RU05 only. No other case changed.
//
// Citation abbreviations used in Why: "Spec Rn" = G009 Spec requirement;
// "matrix" = the owner-approved R3 table; "WP r1 bN" = write-policy r1 bullet N
// (1 upload, 2 own delete, 3 owner/admin delete and Organization mutation,
// 4 removal/downgrade); "AUTH" = AUTHORIZATION.md (gates 1-3, pins (a)-(d),
// B1-B6 binding checks, Errors table); "DBISO" = DATABASE-ISOLATION context
// validity; "ARCH40"/"ARCH53.7" = Architecture sections; "PLAN" = Plan excerpt.
//
// Reading rules (apply to every case):
//   - Cases are independent. Each one is evaluated against the initial fixture
//     state below, never against the effect of an earlier case.
//   - Every service denial (any outcome other than oeAllow) means: zero provider
//     calls, no metadata change and no disclosure of the object (PLAN OBJ-READ /
//     OBJ-WRITE; Spec AC3/AC4 "no denied provider calls", "unchanged DB/provider state").
//   - A read row is expanded into three cases (metadata, get, sign) by
//     oeRead; the Spec gives all three the same read decision (Spec R3, AC3).
//     A sign case without an explicit TTL expects the 60 s default (Spec R5).
//   - Fixture "Bindings" are the current facts at decision time. "Former"
//     bindings are history only and confer nothing (Spec R3, WP r1 b4).
//   - Contexts and rows name fixture keys; the glue maps keys to UUIDs.
//
// AMBIGUITY notes (the chosen reading is what the cases encode):
//
// AMBIGUITY A1 (no selection): "no selection" is read as no TenantContext at
// all (nil/caller-built, AUTH B1), external context_denied. An
// Organization-only TenantContext is the "selected Organization" form. For
// RLS, "no selection" is the EMPTY shape (all settings empty): zero rows, no
// error, and INSERT fails WITH CHECK (DBISO; PLAN OBJ-RLS "With no context it
// sees nothing").
//
// AMBIGUITY A2 (Organization objects under a selected Group): R3 says
// "Selected Group restricts Group targets; Organization authority remains
// independent". The matrix Put cell says "within selected scope". The oracle
// reads "within selected scope" as constraining Group targets only. An
// Organization-scoped object is therefore readable by any current Organization
// member, and writable/deletable by an Organization owner/admin, in a
// selected-Group context too (AUTH gate 3 applies only to group.* targets).
//
// AMBIGUITY A3 (non-ready objects): the Spec requires that incomplete or
// deleted objects are never returned or signed, but names no external code.
// The oracle expects not_found (the same category as a missing ID, Spec R3) for
// metadata/get/sign of pending or deleting objects, even for an authorized
// reader. Delete of a non-ready object and repeated delete are Plan-pinned
// (Spec R7 "Repeated delete/retry behavior is pinned"), so no case is given.
//
// AMBIGUITY A4 (RLS visibility of non-ready rows): the service may need to see
// in-scope pending/deleting rows to finalize or repair. RLS SELECT cases
// therefore constrain only (i) in-scope READY rows: MustSee, and (ii) every
// out-of-scope row in any state: MustNotSee. In-scope non-ready rows are
// unconstrained.
//
// AMBIGUITY A5 (state names): the lifecycle states are the logical names
// pending (prepared, not yet durable/ready), ready and deleting (logical
// deletion committed). The Plan pins the real names; the glue maps them.
//
// AMBIGUITY A6 ("the same global UUID across Organizations"): the object UUID
// is a canonical unique identifier (Spec R2), so two objects sharing one UUID
// are not fixtured. The oracle reads the phrase as (i) one global user UUID
// ("cross") holding bindings and uploads in both Organizations, and (ii) an
// existing object UUID presented under the other Organization's context. If
// the Plan instead permits per-Organization object-UUID collisions, more
// cases are needed (return to Codex).
//
// AMBIGUITY A7 (invalid-input code): the oracle uses invalid_input from AUTH's
// closed error list for malformed selectors and TTL values. The inputs do not
// define invalid_request.
//
// AMBIGUITY A8 (forbidden versus not_found for Put): AUTH says the choice
// "uses the read capability of the same kind on the same facts". For Put, the
// read capability is the one for the target scope. An Organization-scoped Put
// by a member who can read the Organization is forbidden. A Group Put by a
// subject without group.read on that Group (or outside the selected Group) is
// not_found.
//
// AMBIGUITY A9 (TTL encoding): the TTL is whole seconds. "Unset" is distinct
// from zero: unset gives the 60 s default, and an explicit 0 is rejected
// (Spec R5 "Reject invalid, zero/negative/overflow and above-cap values"). The
// overflow case uses a seconds value whose nanosecond duration overflows
// int64. Config-level TTL bounds are Plan-pinned and not covered.
//
// AMBIGUITY A10 (RLS denial form): for an UPDATE of a row that is visible but
// not authorized for that transition, or an illegal state transition,
// PostgreSQL may filter it (0 rows), raise 42501 (WITH CHECK) or a trigger may
// raise. The oracle accepts any of these (oeRLSDeniedNoChange), provided the
// row is unchanged. For immutable columns the Plan picks column
// privilege (42501) or a trigger exception per column. The oracle accepts
// either (oeRLSImmutable), provided the row is unchanged. r2: per Codex
// ruling a466be31, RU05 alone (a non-DEFAULT assignment to the
// database-generated storage key, logical column object_key) also accepts
// SQLSTATE 428C9 (oeRLSImmutableGenerated), with the row unchanged. No other
// immutable case accepts 428C9.
//
// AMBIGUITY A11 (RLS insert state): only the initial pending state may be
// inserted. A row inserted already ready or deleting must fail (Spec R7 "Do
// not make metadata READY ... before the declared durable upload outcome").
// The mechanism may be WITH CHECK (42501), a CHECK constraint (23514) or a
// trigger; oeRLSStateRejected accepts any statement error with no row
// inserted.
//
// AMBIGUITY A12 (finalize by a non-creator): whether another authorized
// member or admin may move someone else's pending row to ready is Plan-pinned.
// Only the creator's finalize (allow) and an out-of-scope finalize (0 rows)
// are given.
//
// AMBIGUITY A13 (RLS and session liveness): the context-validity section does
// not tie RLS to session/user liveness, so RLS cases do not vary session or
// user state. Liveness is covered at the service level.
//
// AMBIGUITY A14 (precedence): the binding check B1-B6 runs "before any
// transaction" and before session liveness (AUTH). A forged context with a
// revoked session is therefore context_denied. Session liveness precedes the
// membership/selectability recheck (AUTH Resolve step 3 order), so a revoked
// session with a stale context is unauthenticated. The relative order of
// payload validation and the binding check is unspecified, so no case
// combines them. Invalid-input cases always use a valid context and an
// authorized actor.
//
// AMBIGUITY A15 ("revoked or former admin"): formerOrgAdmin was downgraded from
// org_admin to org_member and stays in the Organization; its Organization
// mutations are forbidden. revokedOrgAdmin was removed from the Organization
// entirely, so any context in that Organization is context_denied (AUTH: no
// Organization membership), not not_found.
//
// AMBIGUITY A16 (FORCE owner subject): the Plan does not name the app.user_id
// of the owner session. The oracle uses aOwner (org_owner of orgA). Under
// FORCE the owner is subject to the same policies, so an in-context orgA
// INSERT by it is expected to succeed (positive pair for the foreign-INSERT
// denial).
//
// AMBIGUITY A17 (selector casing): AUTH lists malformed, braced, unhyphenated,
// empty and all-zero selectors as invalid_input. It does not mention uppercase
// hex, so no uppercase case is given.
//
// AMBIGUITY A18 (Org-only context and Group targets): in an Organization-only
// context there is no context gate, so a Group member reaches its own Group's
// objects. Read, own-delete and Put are allowed exactly as in that
// Group's selected context (AUTH gate 3 applies only in a selected-Group
// context).
//
// AMBIGUITY A19 (Group roles and Organization objects): group_owner/group_admin
// give no Organization-object mutation. Such a subject reads Organization
// objects only through its org_member binding (AUTH pin (a)), so its
// Organization Put/Delete is forbidden, not not_found.

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

type oeRole string

const (
	oeRoleOrgOwner    oeRole = "org_owner"
	oeRoleOrgAdmin    oeRole = "org_admin"
	oeRoleOrgMember   oeRole = "org_member"
	oeRoleGroupOwner  oeRole = "group_owner"
	oeRoleGroupAdmin  oeRole = "group_admin"
	oeRoleGroupMember oeRole = "group_member"
)

type oeScopeType string

const (
	oeScopeOrganization oeScopeType = "organization"
	oeScopeGroup        oeScopeType = "group"
)

// oeBinding is one RoleBinding (source membership). ScopeKey names an
// Organization key for oeScopeOrganization, a Group key for oeScopeGroup.
type oeBinding struct {
	Scope    oeScopeType
	ScopeKey string
	Role     oeRole
}

type oeUserState string

const (
	oeUserActive   oeUserState = "active"
	oeUserDisabled oeUserState = "disabled"
)

type oeSessionState string

const (
	oeSessionLive    oeSessionState = "live"
	oeSessionRevoked oeSessionState = "revoked"
	oeSessionExpired oeSessionState = "expired"
)

type oeUser struct {
	Key            string
	ID             string
	State          oeUserState
	Session        oeSessionState
	Bindings       []oeBinding // current facts at decision time
	FormerBindings []oeBinding // history only; confers nothing
	Note           string
}

type oeOrganization struct {
	Key string
	ID  string
}

type oeGroup struct {
	Key             string
	ID              string
	OrganizationKey string
}

// oeProfile is a BusinessProfile and the Groups attached to it (ADR-004:
// at most one profile per Group, one profile reusable by several Groups).
type oeProfile struct {
	Key             string
	ID              string
	OrganizationKey string
	GroupKeys       []string
}

type oeObjectState string

const (
	oeStatePending  oeObjectState = "pending"
	oeStateReady    oeObjectState = "ready"
	oeStateDeleting oeObjectState = "deleting"
)

// oeObject is one stored-object metadata row. GroupKey "" means
// Organization-scoped. Stored=false marks an ID that exists nowhere (the glue
// must not insert it).
type oeObject struct {
	Key             string
	ID              string
	OrganizationKey string
	GroupKey        string
	CreatorKey      string
	State           oeObjectState
	Stored          bool
	Note            string
}

type oeFixture struct {
	Organizations []oeOrganization
	Groups        []oeGroup
	Profiles      []oeProfile
	Users         []oeUser
	Objects       []oeObject
}

type oeContextKind string

const (
	oeCtxNone         oeContextKind = "none"         // no TenantContext; RLS EMPTY shape
	oeCtxOrganization oeContextKind = "organization" // TENANT-ORG
	oeCtxGroup        oeContextKind = "group"        // TENANT-GROUP
	oeCtxRaw          oeContextKind = "raw"          // RLS only: literal settings in Raw
)

// oeBindingFault selects an AUTH binding-check failure (B1-B6) applied to an
// otherwise valid context. All give external context_denied.
type oeBindingFault string

const (
	oeFaultNone            oeBindingFault = ""
	oeFaultForged          oeBindingFault = "forged_B1"           // nil, zero-valued or caller-built
	oeFaultForeignService  oeBindingFault = "foreign_service_B2"  // issued by another Service
	oeFaultScopeMissing    oeBindingFault = "scope_missing_B3"    // ctx carries no scope
	oeFaultScopeEnded      oeBindingFault = "scope_ended_B4"      // scope ended
	oeFaultScopeMismatch   oeBindingFault = "scope_mismatch_B5"   // context from another request
	oeFaultSubjectMismatch oeBindingFault = "subject_mismatch_B6" // another user's context under this scope (B5 and B6)
)

// oeRawSettings are literal transaction-local settings for RLS-only cases.
type oeRawSettings struct {
	UserID         string
	OrganizationID string
	GroupID        string
	LoginBinding   string
}

// oeContext is the request context of a case. Stale means the context was
// resolved while valid and its basis (membership or Group selectability) was
// revoked before the operation; RLS cases evaluate current facts only.
type oeContext struct {
	Kind            oeContextKind
	OrganizationKey string
	GroupKey        string
	Stale           bool
	Fault           oeBindingFault
	Raw             oeRawSettings
}

type oeAction string

const (
	oeActMetadata      oeAction = "metadata"
	oeActGet           oeAction = "get"
	oeActSign          oeAction = "sign"
	oeActPut           oeAction = "put"
	oeActDelete        oeAction = "delete"
	oeActRLSSelect     oeAction = "rls_select"
	oeActRLSInsert     oeAction = "rls_insert"
	oeActRLSTransition oeAction = "rls_transition"
	oeActRLSImmutable  oeAction = "rls_update_immutable"
	oeActRLSDelete     oeAction = "rls_delete"
)

type oeDBRole string

const (
	oeDBNone    oeDBRole = ""        // service-level case
	oeDBRuntime oeDBRole = "runtime" // runtime login, direct SQL without application predicates
	oeDBOwner   oeDBRole = "owner"   // table owner under FORCE ROW LEVEL SECURITY
)

type oeOutcome string

const (
	// Service-level outcomes (AUTH Errors table).
	oeAllow           oeOutcome = "allow"
	oeUnauthenticated oeOutcome = "unauthenticated"
	oeContextDenied   oeOutcome = "context_denied"
	oeNotFound        oeOutcome = "not_found"
	oeForbidden       oeOutcome = "forbidden"
	oeInvalidInput    oeOutcome = "invalid_input"

	// RLS outcomes. oeAllow for RLS means the statement succeeds; INSERT and
	// UPDATE affect exactly one row; SELECT satisfies MustSee/MustNotSee.
	oeRLSNoRows             oeOutcome = "rls_no_rows"             // SELECT succeeds and returns none of the fixture rows
	oeRLSZeroRows           oeOutcome = "rls_zero_rows"           // UPDATE succeeds and affects 0 rows
	oeRLSWithCheck          oeOutcome = "rls_with_check_42501"    // new row violates row-level security policy
	oeRLSStateRejected      oeOutcome = "rls_state_rejected"      // statement error (42501, 23514 or trigger); no row inserted
	oeRLSDeniedNoChange     oeOutcome = "rls_denied_no_change"    // 0 rows, 42501 or trigger exception; the row is unchanged
	oeRLSImmutable          oeOutcome = "rls_immutable"           // 42501 privilege or trigger exception; the row is unchanged
	oeRLSImmutableGenerated oeOutcome = "rls_immutable_generated" // 428C9, 42501 privilege or trigger exception; the row is unchanged (RU05 only)
	oeSQLPrivilege          oeOutcome = "sqlstate_42501_privilege"
	oeSQL22P02              oeOutcome = "sqlstate_22P02"
	oeSQL22023              oeOutcome = "sqlstate_22023"
)

// oeTTL is a sign request TTL in whole seconds; Set=false means not supplied.
type oeTTL struct {
	Set     bool
	Seconds int64
}

// oeRow is an RLS INSERT row. GroupKey "" means Organization-scoped.
type oeRow struct {
	OrganizationKey string
	GroupKey        string
	CreatorKey      string
	State           oeObjectState
}

// oeCase is one actor x context x object x action expectation.
//
//   - Object: fixture object key for read/delete/RLS update/delete cases.
//   - RawObjectID / RawGroupID: literal selector strings (invalid-input cases).
//   - PutGroupKey: Put target; "" = Organization-scoped object in the context's
//     active Organization.
//   - TTL / ExpectTTLSeconds: sign only. ExpectTTLSeconds is the issued TTL on allow.
//   - Row: RLS insert. ToState: RLS transition. Column/NewValue: RLS immutable
//     update (logical column names; NewValue names a fixture key or "<fresh>").
//   - MustSee / MustNotSee: RLS select object keys.
//   - Pair: ID of the nearest allowed variant (set on every denial).
type oeCase struct {
	ID               string
	Action           oeAction
	DBRole           oeDBRole
	Actor            string
	Ctx              oeContext
	Object           string
	RawObjectID      string
	PutGroupKey      string
	RawGroupID       string
	TTL              oeTTL
	Row              oeRow
	ToState          oeObjectState
	Column           string
	NewValue         string
	Expect           oeOutcome
	ExpectTTLSeconds int64
	MustSee          []string
	MustNotSee       []string
	Why              string
	Pair             string
}

// ---------------------------------------------------------------------------
// Fixture
// ---------------------------------------------------------------------------

func oeOrgB(org string, role oeRole) oeBinding {
	return oeBinding{Scope: oeScopeOrganization, ScopeKey: org, Role: role}
}

func oeGrpB(group string, role oeRole) oeBinding {
	return oeBinding{Scope: oeScopeGroup, ScopeKey: group, Role: role}
}

func objectsExpectedFixture() oeFixture {
	live, active := oeSessionLive, oeUserActive
	return oeFixture{
		Organizations: []oeOrganization{
			{Key: "orgA", ID: "10000000-0000-4000-8000-0000000000a1"},
			{Key: "orgB", ID: "10000000-0000-4000-8000-0000000000b1"},
		},
		Groups: []oeGroup{
			{Key: "g1", ID: "20000000-0000-4000-8000-0000000000a1", OrganizationKey: "orgA"},
			{Key: "g2", ID: "20000000-0000-4000-8000-0000000000a2", OrganizationKey: "orgA"}, // sibling of g1
			{Key: "gb", ID: "20000000-0000-4000-8000-0000000000b1", OrganizationKey: "orgB"}, // foreign Organization
		},
		Profiles: []oeProfile{
			// One BusinessProfile shared by sibling Groups g1 and g2 (ADR-004).
			{Key: "p1", ID: "30000000-0000-4000-8000-0000000000a1", OrganizationKey: "orgA", GroupKeys: []string{"g1", "g2"}},
			{Key: "pb", ID: "30000000-0000-4000-8000-0000000000b1", OrganizationKey: "orgB", GroupKeys: []string{"gb"}},
		},
		Users: []oeUser{
			{Key: "aOwner", ID: "40000000-0000-4000-8000-000000000001", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgOwner)}},
			{Key: "aAdmin", ID: "40000000-0000-4000-8000-000000000002", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgAdmin)}},
			{Key: "aMember", ID: "40000000-0000-4000-8000-000000000003", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember)}, Note: "Organization member without any Group binding"},
			{Key: "g1Owner", ID: "40000000-0000-4000-8000-000000000004", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g1", oeRoleGroupOwner)}},
			{Key: "g1Admin", ID: "40000000-0000-4000-8000-000000000005", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g1", oeRoleGroupAdmin)}},
			{Key: "g1MemberX", ID: "40000000-0000-4000-8000-000000000006", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g1", oeRoleGroupMember)}},
			{Key: "g1MemberY", ID: "40000000-0000-4000-8000-000000000007", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g1", oeRoleGroupMember)}},
			{Key: "g2Member", ID: "40000000-0000-4000-8000-000000000008", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g2", oeRoleGroupMember)},
				Note:     "reads shared profile p1 through g2; must never reach g1 objects"},
			{Key: "multi", ID: "40000000-0000-4000-8000-000000000009", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g1", oeRoleGroupMember), oeGrpB("g2", oeRoleGroupAdmin)},
				Note:     "member of g1 and admin of g2 (scope discovery with more than one Group binding)"},
			{Key: "downgraded", ID: "40000000-0000-4000-8000-000000000010", State: active, Session: live,
				Bindings:       []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g1", oeRoleGroupMember)},
				FormerBindings: []oeBinding{oeGrpB("g1", oeRoleGroupAdmin)},
				Note:           "g1 admin downgraded to g1 member; uploaded oG1D while admin"},
			{Key: "formerOrgAdmin", ID: "40000000-0000-4000-8000-000000000011", State: active, Session: live,
				Bindings:       []oeBinding{oeOrgB("orgA", oeRoleOrgMember)},
				FormerBindings: []oeBinding{oeOrgB("orgA", oeRoleOrgAdmin)},
				Note:           "org_admin downgraded to org_member; uploaded oAOrgFormer while admin"},
			{Key: "removedG1", ID: "40000000-0000-4000-8000-000000000012", State: active, Session: live,
				Bindings:       []oeBinding{oeOrgB("orgA", oeRoleOrgMember)},
				FormerBindings: []oeBinding{oeGrpB("g1", oeRoleGroupMember)},
				Note:           "removed from g1, still in orgA; uploaded oG1R and prepared oG1RPending while a member"},
			{Key: "revokedOrgAdmin", ID: "40000000-0000-4000-8000-000000000013", State: active, Session: live,
				Bindings:       nil,
				FormerBindings: []oeBinding{oeOrgB("orgA", oeRoleOrgAdmin)},
				Note:           "former org_admin removed from orgA entirely; uploaded oAOrgRevoked while admin"},
			{Key: "cross", ID: "40000000-0000-4000-8000-000000000014", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g1", oeRoleGroupMember),
					oeOrgB("orgB", oeRoleOrgMember), oeGrpB("gb", oeRoleGroupMember)},
				Note: "one global user UUID with memberships and uploads in both Organizations"},
			{Key: "bOwner", ID: "40000000-0000-4000-8000-000000000015", State: active, Session: live,
				Bindings: []oeBinding{oeOrgB("orgB", oeRoleOrgOwner)}, Note: "foreign Organization owner"},
			{Key: "sessionRevoked", ID: "40000000-0000-4000-8000-000000000016", State: active, Session: oeSessionRevoked,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgMember), oeGrpB("g1", oeRoleGroupMember)},
				Note:     "bindings intact, session revoked before the operation; uploaded oG1S"},
			{Key: "sessionExpired", ID: "40000000-0000-4000-8000-000000000017", State: active, Session: oeSessionExpired,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgOwner)}},
			{Key: "disabled", ID: "40000000-0000-4000-8000-000000000018", State: oeUserDisabled, Session: live,
				Bindings: []oeBinding{oeOrgB("orgA", oeRoleOrgAdmin)}},
		},
		Objects: []oeObject{
			{Key: "oAOrg1", ID: "50000000-0000-4000-8000-000000000001", OrganizationKey: "orgA", CreatorKey: "aOwner", State: oeStateReady, Stored: true},
			{Key: "oAOrgFormer", ID: "50000000-0000-4000-8000-000000000002", OrganizationKey: "orgA", CreatorKey: "formerOrgAdmin", State: oeStateReady, Stored: true},
			{Key: "oAOrgRevoked", ID: "50000000-0000-4000-8000-000000000003", OrganizationKey: "orgA", CreatorKey: "revokedOrgAdmin", State: oeStateReady, Stored: true},
			{Key: "oG1X", ID: "50000000-0000-4000-8000-000000000004", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "g1MemberX", State: oeStateReady, Stored: true},
			{Key: "oG1Y", ID: "50000000-0000-4000-8000-000000000005", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "g1MemberY", State: oeStateReady, Stored: true},
			{Key: "oG1D", ID: "50000000-0000-4000-8000-000000000006", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "downgraded", State: oeStateReady, Stored: true},
			{Key: "oG1R", ID: "50000000-0000-4000-8000-000000000007", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "removedG1", State: oeStateReady, Stored: true},
			{Key: "oG1C", ID: "50000000-0000-4000-8000-000000000008", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "cross", State: oeStateReady, Stored: true},
			{Key: "oG1S", ID: "50000000-0000-4000-8000-000000000009", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "sessionRevoked", State: oeStateReady, Stored: true},
			{Key: "oG1Pending", ID: "50000000-0000-4000-8000-000000000010", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "g1MemberX", State: oeStatePending, Stored: true},
			{Key: "oG1Deleting", ID: "50000000-0000-4000-8000-000000000011", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "g1MemberX", State: oeStateDeleting, Stored: true},
			{Key: "oG1RPending", ID: "50000000-0000-4000-8000-000000000012", OrganizationKey: "orgA", GroupKey: "g1", CreatorKey: "removedG1", State: oeStatePending, Stored: true},
			{Key: "oG2M", ID: "50000000-0000-4000-8000-000000000013", OrganizationKey: "orgA", GroupKey: "g2", CreatorKey: "g2Member", State: oeStateReady, Stored: true},
			{Key: "oBOrg", ID: "50000000-0000-4000-8000-000000000014", OrganizationKey: "orgB", CreatorKey: "bOwner", State: oeStateReady, Stored: true},
			{Key: "oGBC", ID: "50000000-0000-4000-8000-000000000015", OrganizationKey: "orgB", GroupKey: "gb", CreatorKey: "cross", State: oeStateReady, Stored: true},
			{Key: "oMissing", ID: "50000000-0000-4000-8000-0000000000ff", Stored: false, Note: "canonical UUID that exists nowhere"},
		},
	}
}

// Object-key sets for RLS SELECT expectations.
var (
	oeSetAOrg        = []string{"oAOrg1", "oAOrgFormer", "oAOrgRevoked"}
	oeSetG1Ready     = []string{"oG1X", "oG1Y", "oG1D", "oG1R", "oG1C", "oG1S"}
	oeSetG1NonReady  = []string{"oG1Pending", "oG1Deleting", "oG1RPending"}
	oeSetG2          = []string{"oG2M"}
	oeSetB           = []string{"oBOrg", "oGBC"}
	oeSetAllA        = oeJoin(oeSetAOrg, oeSetG1Ready, oeSetG1NonReady, oeSetG2)
	oeSetAllStored   = oeJoin(oeSetAllA, oeSetB)
	oeSetG1All       = oeJoin(oeSetG1Ready, oeSetG1NonReady)
	oeSetGroupsAllA  = oeJoin(oeSetG1All, oeSetG2)
	oeSetNothingSeen = []string{}
)

// ---------------------------------------------------------------------------
// Tiny helpers
// ---------------------------------------------------------------------------

func oeJoin(sets ...[]string) []string {
	var out []string
	for _, s := range sets {
		out = append(out, s...)
	}
	return out
}

func oeNoCtx() oeContext { return oeContext{Kind: oeCtxNone} }

func oeOrgCtx(org string) oeContext {
	return oeContext{Kind: oeCtxOrganization, OrganizationKey: org}
}

func oeGroupCtx(org, group string) oeContext {
	return oeContext{Kind: oeCtxGroup, OrganizationKey: org, GroupKey: group}
}

func oeStale(c oeContext) oeContext { c.Stale = true; return c }

func oeFault(c oeContext, f oeBindingFault) oeContext { c.Fault = f; return c }

func oeRawCtx(r oeRawSettings) oeContext { return oeContext{Kind: oeCtxRaw, Raw: r} }

// oeIsDenial reports whether an outcome is a denial (zero provider calls and
// unchanged state are then required for service cases).
func oeIsDenial(o oeOutcome) bool { return o != oeAllow }

// oeRead expands one read decision into metadata, get and sign cases
// (sign with the TTL unset, so the 60 s default on allow).
func oeRead(id, actor string, ctx oeContext, object string, expect oeOutcome, why, pair string) []oeCase {
	return oeReadAny(id, actor, ctx, object, "", expect, why, pair)
}

func oeReadRaw(id, actor string, ctx oeContext, raw string, expect oeOutcome, why, pair string) []oeCase {
	return oeReadAny(id, actor, ctx, "", raw, expect, why, pair)
}

func oeReadAny(id, actor string, ctx oeContext, object, raw string, expect oeOutcome, why, pair string) []oeCase {
	var out []oeCase
	for _, a := range []oeAction{oeActMetadata, oeActGet, oeActSign} {
		c := oeCase{ID: id + "/" + string(a), Action: a, Actor: actor, Ctx: ctx, Object: object, RawObjectID: raw, Expect: expect, Why: why}
		if pair != "" {
			c.Pair = pair + "/" + string(a)
		}
		if a == oeActSign && expect == oeAllow {
			c.ExpectTTLSeconds = 60
		}
		out = append(out, c)
	}
	return out
}

func oeSign(id, actor string, ctx oeContext, object string, ttl oeTTL, expect oeOutcome, ttlSeconds int64, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActSign, Actor: actor, Ctx: ctx, Object: object, TTL: ttl, Expect: expect, ExpectTTLSeconds: ttlSeconds, Why: why, Pair: pair}
}

func oePut(id, actor string, ctx oeContext, group string, expect oeOutcome, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActPut, Actor: actor, Ctx: ctx, PutGroupKey: group, Expect: expect, Why: why, Pair: pair}
}

func oePutRaw(id, actor string, ctx oeContext, rawGroup string, expect oeOutcome, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActPut, Actor: actor, Ctx: ctx, RawGroupID: rawGroup, Expect: expect, Why: why, Pair: pair}
}

func oeDel(id, actor string, ctx oeContext, object string, expect oeOutcome, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActDelete, Actor: actor, Ctx: ctx, Object: object, Expect: expect, Why: why, Pair: pair}
}

func oeDelRaw(id, actor string, ctx oeContext, raw string, expect oeOutcome, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActDelete, Actor: actor, Ctx: ctx, RawObjectID: raw, Expect: expect, Why: why, Pair: pair}
}

func oeSel(id string, role oeDBRole, actor string, ctx oeContext, expect oeOutcome, see, notSee []string, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActRLSSelect, DBRole: role, Actor: actor, Ctx: ctx, Expect: expect, MustSee: see, MustNotSee: notSee, Why: why, Pair: pair}
}

func oeIns(id string, role oeDBRole, actor string, ctx oeContext, row oeRow, expect oeOutcome, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActRLSInsert, DBRole: role, Actor: actor, Ctx: ctx, Row: row, Expect: expect, Why: why, Pair: pair}
}

func oeTrans(id string, role oeDBRole, actor string, ctx oeContext, object string, to oeObjectState, expect oeOutcome, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActRLSTransition, DBRole: role, Actor: actor, Ctx: ctx, Object: object, ToState: to, Expect: expect, Why: why, Pair: pair}
}

func oeImm(id, actor string, ctx oeContext, object, column, newValue string, expect oeOutcome, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActRLSImmutable, DBRole: oeDBRuntime, Actor: actor, Ctx: ctx, Object: object, Column: column, NewValue: newValue, Expect: expect, Why: why, Pair: pair}
}

func oeRLSDel(id, actor string, ctx oeContext, object string, expect oeOutcome, why, pair string) oeCase {
	return oeCase{ID: id, Action: oeActRLSDelete, DBRole: oeDBRuntime, Actor: actor, Ctx: ctx, Object: object, Expect: expect, Why: why, Pair: pair}
}

func oeR(org, group, creator string, state oeObjectState) oeRow {
	return oeRow{OrganizationKey: org, GroupKey: group, CreatorKey: creator, State: state}
}

// ---------------------------------------------------------------------------
// Cases
// ---------------------------------------------------------------------------

func objectsExpectedCases() []oeCase {
	var cs []oeCase
	add := func(c ...oeCase) { cs = append(cs, c...) }

	A, B := oeOrgCtx("orgA"), oeOrgCtx("orgB")
	AG1, AG2 := oeGroupCtx("orgA", "g1"), oeGroupCtx("orgA", "g2")
	BGB := oeGroupCtx("orgB", "gb")
	unset := oeTTL{}
	ttl := func(s int64) oeTTL { return oeTTL{Set: true, Seconds: s} }
	rt, ow := oeDBRuntime, oeDBOwner
	P, R, D := oeStatePending, oeStateReady, oeStateDeleting

	// ---- READ (metadata/get/sign): Organization-scoped objects ----
	add(oeRead("R01", "aOwner", A, "oAOrg1", oeAllow, "Spec R3 matrix: Organization object read via organization.read (org_owner)", "")...)
	add(oeRead("R02", "aMember", A, "oAOrg1", oeAllow, "Spec R3 matrix: any current Organization member (org_member organization.read)", "")...)
	add(oeRead("R03", "g1MemberX", AG1, "oAOrg1", oeAllow, "Spec R3: Organization authority independent of selected Group; AUTH pin (a) org_member path; AMBIGUITY A2", "")...)
	add(oeRead("R04", "g2Member", A, "oAOrgFormer", oeAllow, "Spec R3 matrix: Organization object readable by current members regardless of uploader", "")...)
	add(oeRead("R05", "formerOrgAdmin", A, "oAOrgFormer", oeAllow, "Spec R3: downgraded admin is still an org_member and reads", "")...)
	add(oeRead("R06", "revokedOrgAdmin", oeStale(A), "oAOrgRevoked", oeContextDenied, "AUTH Resolve/Errors: no current Organization membership -> context_denied even for own upload; AMBIGUITY A15", "R05")...)
	add(oeRead("R07", "bOwner", B, "oAOrg1", oeNotFound, "Spec R3: foreign object ID same external category; AUTH gate 2", "R08")...)
	add(oeRead("R08", "bOwner", B, "oBOrg", oeAllow, "Spec R3 matrix: org_owner reads own Organization object", "")...)
	add(oeRead("R09", "aOwner", A, "oBOrg", oeNotFound, "Spec R3: object of another Organization is not_found; owner role in orgA gives nothing in orgB", "R01")...)
	add(oeRead("R10", "aOwner", B, "oBOrg", oeContextDenied, "AUTH Resolve: foreign Organization selection without membership -> context_denied", "R08")...)
	add(oeRead("R11", "aMember", oeNoCtx(), "oAOrg1", oeContextDenied, "AUTH B1: no TenantContext; AMBIGUITY A1", "R02")...)
	add(oeRead("R12", "aAdmin", AG1, "oAOrgRevoked", oeAllow, "Spec R3: Organization object of a removed uploader stays readable by current members; AMBIGUITY A2", "")...)

	// ---- READ: Group-scoped objects ----
	add(oeRead("R20", "aOwner", A, "oG1X", oeNotFound, "G060 Spec R3 / ADR-008: a Group object is content; Organization owner/admin without a role in its Group -> not_found under either override value (was org_owner 'all')", "R26")...)
	add(oeRead("R21", "aAdmin", AG1, "oG1X", oeNotFound, "G060 Spec R3 / ADR-008: a Group object is content; Organization owner/admin without a role in its Group -> not_found under either override value; selecting the Group confers nothing", "R25")...)
	add(oeRead("R22", "aAdmin", AG2, "oG1X", oeNotFound, "Spec R3 selected Group restricts Group targets; AUTH pin (b), gate 3 -> not_found", "R25")...)
	add(oeRead("R23", "g1Owner", AG1, "oG1Y", oeAllow, "AUTH matrix group_owner 'own' group.read", "")...)
	add(oeRead("R24", "g1Admin", A, "oG1Y", oeAllow, "AUTH matrix group_admin 'own'; Org-only context has no context gate; AMBIGUITY A18", "")...)
	add(oeRead("R25", "g1MemberX", AG1, "oG1Y", oeAllow, "Spec R3 matrix Group read: group_member reads any object of its Group, not only own", "")...)
	add(oeRead("R26", "g1MemberX", A, "oG1X", oeAllow, "Spec R3 matrix; AMBIGUITY A18", "")...)
	add(oeRead("R27", "aMember", A, "oG1X", oeNotFound, "AUTH matrix: org_member has no group.read; inaccessible -> not_found (Spec R3)", "R26")...)
	add(oeRead("R28", "g2Member", A, "oG1X", oeNotFound, "Spec R3: profile p1 shared with visible g2 never reveals g1 objects; ARCH53.7 no sibling access", "R26")...)
	add(oeRead("R29", "g2Member", AG2, "oG1X", oeNotFound, "Spec R3 shared-profile sibling; AUTH gate 3", "R30")...)
	add(oeRead("R30", "g2Member", AG2, "oG2M", oeAllow, "Spec R3 matrix group_member own Group", "")...)
	add(oeRead("R31", "g1MemberX", AG2, "oG2M", oeContextDenied, "AUTH Resolve: g2 not selectable without org_owner/org_admin or a g2 binding", "R30")...)
	add(oeRead("R32", "g1MemberX", AG1, "oG2M", oeNotFound, "Spec R3 sibling Group object; AUTH gate 3", "R30")...)
	add(oeRead("R33", "multi", A, "oG2M", oeAllow, "AUTH matrix group_admin own g2; PLAN r1 scope discovery with several bindings", "")...)
	add(oeRead("R34", "multi", A, "oG1X", oeAllow, "AUTH matrix group_member own g1; PLAN r1 scope discovery", "")...)
	add(oeRead("R35", "multi", AG1, "oG2M", oeNotFound, "Spec R3 selected Group restricts Group targets even with a g2 binding", "R33")...)
	add(oeRead("R36", "removedG1", A, "oG1R", oeNotFound, "Spec R3: being the uploader grants nothing after removal; creator equality never bypasses (R2)", "R37")...)
	add(oeRead("R37", "g1Admin", AG1, "oG1R", oeAllow, "Spec R3 matrix: object of a removed uploader stays readable by current g1 callers", "")...)
	add(oeRead("R38", "removedG1", oeStale(AG1), "oG1R", oeContextDenied, "AUTH Errors: Group selectability no longer holds -> context_denied", "R39")...)
	add(oeRead("R39", "removedG1", A, "oAOrg1", oeAllow, "Spec R3 matrix: still an org_member, reads Organization objects", "")...)
	add(oeRead("R40", "cross", AG1, "oG1C", oeAllow, "Spec R3 matrix group_member in orgA/g1", "")...)
	add(oeRead("R41", "cross", B, "oGBC", oeAllow, "Spec R3 matrix group_member in orgB/gb (Org-only context)", "")...)
	add(oeRead("R42", "cross", A, "oGBC", oeNotFound, "Spec R3: same global UUID across tenants does not waive the boundary; AMBIGUITY A6", "R41")...)
	add(oeRead("R43", "cross", B, "oG1C", oeNotFound, "Spec R3: own upload in orgA invisible from orgB context; AMBIGUITY A6", "R40")...)
	add(oeRead("R44", "cross", oeGroupCtx("orgA", "gb"), "oGBC", oeContextDenied, "AUTH Resolve: selected Group must belong to the Organization", "R41")...)
	add(oeRead("R45", "aOwner", A, "oMissing", oeNotFound, "Spec R3: missing ID same external category as foreign/inaccessible", "R01")...)
	add(oeRead("R46", "aOwner", A, "oG1Pending", oeNotFound, "Spec R2/R7: incomplete object never returned or signed; AMBIGUITY A3", "R26")...)
	add(oeRead("R47", "g1MemberX", AG1, "oG1Pending", oeNotFound, "Spec R2/R7: own pending upload not readable before READY; AMBIGUITY A3", "R25")...)
	add(oeRead("R48", "aOwner", A, "oG1Deleting", oeNotFound, "Spec R7: no new guarded read/signing after logical deletion commits; AMBIGUITY A3", "R26")...)
	add(oeRead("R49", "aMember", B, "oMissing", oeContextDenied, "PLAN OBJ-READ: invalid context with an unknown object is context_denied", "R02")...)
	add(oeRead("R50", "sessionRevoked", AG1, "oG1S", oeUnauthenticated, "AUTH Errors: revoked session -> unauthenticated; Spec R8; even for own upload", "R51")...)
	add(oeRead("R51", "g1MemberX", AG1, "oG1S", oeAllow, "Spec R3 matrix: live g1 member reads the object", "")...)
	add(oeRead("R52", "sessionExpired", A, "oAOrg1", oeUnauthenticated, "AUTH Errors: expired session -> unauthenticated", "R01")...)
	add(oeRead("R53", "disabled", A, "oAOrg1", oeUnauthenticated, "AUTH Errors: disabled user -> unauthenticated; Spec R8", "R01")...)
	add(oeRead("R54", "aOwner", oeFault(A, oeFaultForged), "oAOrg1", oeContextDenied, "AUTH B1 context_forged", "R01")...)
	add(oeRead("R55", "aOwner", oeFault(A, oeFaultForeignService), "oAOrg1", oeContextDenied, "AUTH B2 issued by another Service", "R01")...)
	add(oeRead("R56", "aOwner", oeFault(A, oeFaultScopeMissing), "oAOrg1", oeContextDenied, "AUTH B3 scope missing", "R01")...)
	add(oeRead("R57", "aOwner", oeFault(A, oeFaultScopeEnded), "oAOrg1", oeContextDenied, "AUTH B4 scope ended (stale request)", "R01")...)
	add(oeRead("R58", "aOwner", oeFault(A, oeFaultScopeMismatch), "oAOrg1", oeContextDenied, "AUTH B5 cross-request substitution", "R01")...)
	add(oeRead("R59", "aOwner", oeFault(A, oeFaultSubjectMismatch), "oAOrg1", oeContextDenied, "AUTH B5/B6 another user's context", "R01")...)
	add(oeRead("R60", "sessionRevoked", oeFault(AG1, oeFaultForged), "oG1S", oeContextDenied, "AUTH binding check precedes session liveness; AMBIGUITY A14", "R51")...)
	add(oeRead("R61", "downgraded", AG1, "oG1Y", oeAllow, "Spec R3: downgraded admin keeps g1 member read", "")...)
	add(oeRead("R62", "g1Owner", A, "oG2M", oeNotFound, "AUTH matrix 'own': g1 owner has no path to sibling g2", "R30")...)
	add(oeRead("R63", "aAdmin", A, "oGBC", oeNotFound, "Spec R3 foreign Organization object", "R41")...)
	add(oeReadRaw("R70", "g1MemberX", A, "not-a-uuid", oeInvalidInput, "Spec R4 selector validation; AUTH invalid_input; AMBIGUITY A7", "R26")...)
	add(oeReadRaw("R71", "g1MemberX", A, "00000000-0000-0000-0000-000000000000", oeInvalidInput, "Spec R2 nonzero UUID; AUTH all-zero selector invalid_input", "R26")...)
	add(oeReadRaw("R72", "g1MemberX", A, "{50000000-0000-4000-8000-000000000004}", oeInvalidInput, "AUTH braced selector invalid_input", "R26")...)
	add(oeReadRaw("R73", "g1MemberX", A, "50000000000040008000000000000004", oeInvalidInput, "AUTH unhyphenated selector invalid_input", "R26")...)
	add(oeReadRaw("R74", "g1MemberX", A, "", oeInvalidInput, "AUTH empty selector invalid_input", "R26")...)
	add(oeReadRaw("R75", "g1MemberX", A, "../20000000-0000-4000-8000-0000000000a1/50000000-0000-4000-8000-000000000004", oeInvalidInput, "Spec R4: key/path/traversal never reaches the adapter", "R26")...)
	add(oeReadRaw("R76", "g1MemberX", A, "groups%2F20000000-0000-4000-8000-0000000000a1", oeInvalidInput, "Spec R4: encoded separator / prefix is not a selector; ARCH40 prefixes confer no authority", "R26")...)

	// ---- SIGN: TTL bounds (Spec R5), authorized actor and object ----
	add(oeSign("S01", "g1MemberX", AG1, "oG1X", unset, oeAllow, 60, "Spec R5 default TTL 60 s", ""))
	add(oeSign("S02", "g1MemberX", AG1, "oG1X", ttl(1), oeAllow, 1, "Spec R5 lower bound 1 s", ""))
	add(oeSign("S03", "g1MemberX", AG1, "oG1X", ttl(300), oeAllow, 300, "Spec R5 hard maximum 300 s", ""))
	add(oeSign("S04", "g1MemberX", AG1, "oG1X", ttl(60), oeAllow, 60, "Spec R5 explicit in-range TTL", ""))
	add(oeSign("S05", "g1MemberX", AG1, "oG1X", ttl(0), oeInvalidInput, 0, "Spec R5 reject zero; AMBIGUITY A9", "S02"))
	add(oeSign("S06", "g1MemberX", AG1, "oG1X", ttl(-1), oeInvalidInput, 0, "Spec R5 reject negative", "S02"))
	add(oeSign("S07", "g1MemberX", AG1, "oG1X", ttl(301), oeInvalidInput, 0, "Spec R5 reject above-cap", "S03"))
	add(oeSign("S08", "g1MemberX", AG1, "oG1X", ttl(9223372037), oeInvalidInput, 0, "Spec R5 reject overflow; AMBIGUITY A9", "S03"))
	add(oeSign("S09", "aMember", A, "oAOrg1", ttl(300), oeAllow, 300, "Spec R3/R5: Organization object sign by org_member at the cap", ""))
	add(oeSign("S10", "g2Member", A, "oG1X", ttl(30), oeNotFound, 0, "Spec R3/R5: valid TTL does not make an inaccessible object signable", "S11"))
	add(oeSign("S11", "g1MemberX", A, "oG1X", ttl(30), oeAllow, 30, "Spec R3/R5", ""))

	// ---- PUT: Organization-scoped target (PutGroupKey "") ----
	add(oePut("P01", "aOwner", A, "", oeAllow, "Spec R3 matrix Put Organization: owner/admin only", ""))
	add(oePut("P02", "aAdmin", AG1, "", oeAllow, "Spec R3: Organization authority independent of selected Group; AMBIGUITY A2", ""))
	add(oePut("P03", "aMember", A, "", oeForbidden, "Spec R3 matrix Organization Put owner/admin only; readable scope -> forbidden (AMBIGUITY A8)", "P04"))
	add(oePut("P04", "aAdmin", A, "", oeAllow, "Spec R3 matrix Organization Put org_admin", ""))
	add(oePut("P05", "g1Owner", A, "", oeForbidden, "WP r1 b3: Organization mutations stay Organization owner/admin-only; AMBIGUITY A19", "P04"))
	add(oePut("P06", "formerOrgAdmin", A, "", oeForbidden, "Spec R3: former admin is an ordinary member now", "P04"))
	add(oePut("P07", "g1MemberX", AG1, "", oeForbidden, "Spec R3: a Group member may not upload Organization objects", "P02"))
	add(oePut("P08", "revokedOrgAdmin", oeStale(A), "", oeContextDenied, "AUTH: Organization membership gone -> context_denied; AMBIGUITY A15", "P04"))
	add(oePut("P09", "aOwner", oeNoCtx(), "", oeContextDenied, "AUTH B1; AMBIGUITY A1", "P01"))
	add(oePut("P10", "aOwner", B, "", oeContextDenied, "AUTH Resolve: foreign Organization selection", "P01"))
	add(oePut("P11", "aOwner", oeFault(A, oeFaultForged), "", oeContextDenied, "AUTH B1 forged context", "P01"))
	add(oePut("P12", "disabled", A, "", oeUnauthenticated, "AUTH Errors disabled user; Spec R8", "P04"))
	add(oePut("P13", "bOwner", B, "", oeAllow, "Spec R3 matrix org_owner of orgB", ""))
	add(oePut("P14", "sessionExpired", A, "", oeUnauthenticated, "AUTH Errors expired session", "P01"))

	// ---- PUT: Group-scoped target ----
	add(oePut("P20", "g1MemberX", AG1, "g1", oeAllow, "WP r1 b1 / Spec R3 matrix: Group member uploads to its authorized Group", ""))
	add(oePut("P21", "g1MemberX", A, "g1", oeAllow, "Spec R3 matrix within scope; Org-only context; AMBIGUITY A18", ""))
	add(oePut("P22", "g1Admin", AG1, "g1", oeAllow, "Spec R3 matrix group_admin", ""))
	add(oePut("P23", "g1Owner", A, "g1", oeAllow, "Spec R3 matrix group_owner", ""))
	add(oePut("P24", "aAdmin", AG1, "g1", oeNotFound, "G060 Spec R3 / ADR-008: a Group object is content; Organization owner/admin without a role in its Group -> not_found under either override value (was org_admin within selected Group)", "P22"))
	add(oePut("P25", "aOwner", A, "g2", oeNotFound, "G060 Spec R3 / ADR-008: a Group object is content; Organization owner/admin without a role in its Group -> not_found under either override value (was org_owner any Group)", "P23"))
	add(oePut("P26", "aAdmin", AG2, "g1", oeNotFound, "Spec R3 'within selected scope'; AUTH gate 3 target outside selected Group -> not_found", "P22"))
	add(oePut("P27", "g1MemberX", AG1, "g2", oeNotFound, "Spec R3 sibling Group; AMBIGUITY A8", "P28"))
	add(oePut("P28", "g2Member", AG2, "g2", oeAllow, "WP r1 b1", ""))
	add(oePut("P29", "g1MemberX", A, "g2", oeNotFound, "Spec R3 sibling Group in Org-only context", "P28"))
	add(oePut("P30", "aMember", A, "g1", oeNotFound, "Spec R3 matrix: org_member has no Group path; no group.read -> not_found (AMBIGUITY A8)", "P21"))
	add(oePut("P31", "g2Member", A, "g1", oeNotFound, "Spec R3 shared profile p1 confers no Group write", "P21"))
	add(oePut("P32", "removedG1", A, "g1", oeNotFound, "WP r1 b4: removal ends Group upload", "P21"))
	add(oePut("P33", "removedG1", oeStale(AG1), "g1", oeContextDenied, "AUTH Errors: Group selectability no longer holds", "P20"))
	add(oePut("P34", "bOwner", B, "gb", oeNotFound, "G060 Spec R3 / ADR-008: a Group object is content; Organization owner/admin without a role in its Group -> not_found under either override value (was org_owner of orgB)", "P20"))
	add(oePut("P35", "bOwner", B, "g1", oeNotFound, "AUTH gate 2 foreign Group target; Spec R3", "P20"))
	add(oePut("P36", "aOwner", A, "gb", oeNotFound, "AUTH gate 2 Group of another Organization", "P23"))
	add(oePut("P37", "cross", A, "gb", oeNotFound, "Spec R3 same global UUID across tenants; AMBIGUITY A6", "P38"))
	add(oePut("P38", "cross", BGB, "gb", oeAllow, "WP r1 b1 in orgB", ""))
	add(oePut("P39", "sessionRevoked", AG1, "g1", oeUnauthenticated, "AUTH Errors revoked session; Spec R3/R8", "P20"))
	add(oePut("P40", "downgraded", AG1, "g1", oeAllow, "Spec R3: downgraded admin keeps member upload", ""))
	add(oePut("P41", "multi", AG1, "g2", oeNotFound, "Spec R3 selected Group restricts Group targets", "P42"))
	add(oePut("P42", "multi", A, "g2", oeAllow, "Spec R3 matrix group_admin of g2", ""))
	add(oePutRaw("P43", "g1MemberX", AG1, "not-a-uuid", oeInvalidInput, "Spec R4 selector validation; AMBIGUITY A7", "P20"))
	add(oePutRaw("P44", "g1MemberX", AG1, "00000000-0000-0000-0000-000000000000", oeInvalidInput, "AUTH all-zero selector", "P20"))
	add(oePutRaw("P45", "g1MemberX", AG1, "g1/../g2", oeInvalidInput, "Spec R4 path/traversal never reaches the adapter", "P20"))

	// ---- DELETE: Organization-scoped objects ----
	add(oeDel("D01", "aOwner", A, "oAOrg1", oeAllow, "Spec R3 matrix Organization Delete owner/admin", ""))
	add(oeDel("D02", "aAdmin", A, "oAOrgFormer", oeAllow, "Spec R3 matrix org_admin deletes any Organization object", ""))
	add(oeDel("D03", "aMember", A, "oAOrg1", oeForbidden, "Spec R3 matrix Organization Delete owner/admin only; readable -> forbidden", "D04"))
	add(oeDel("D04", "aAdmin", A, "oAOrg1", oeAllow, "Spec R3 matrix", ""))
	add(oeDel("D05", "formerOrgAdmin", A, "oAOrgFormer", oeForbidden, "Spec R3: Organization objects keep admin-only mutation, including a former admin's uploads", "D02"))
	add(oeDel("D06", "g1Owner", A, "oAOrg1", oeForbidden, "WP r1 b3; AMBIGUITY A19", "D04"))
	add(oeDel("D07", "aAdmin", AG1, "oAOrg1", oeAllow, "Spec R3 Organization authority independent of selected Group; AMBIGUITY A2", ""))
	add(oeDel("D08", "revokedOrgAdmin", oeStale(A), "oAOrgRevoked", oeContextDenied, "AUTH: membership gone -> context_denied; own upload irrelevant; AMBIGUITY A15", "D09"))
	add(oeDel("D09", "aAdmin", A, "oAOrgRevoked", oeAllow, "Spec R3 matrix", ""))
	add(oeDel("D10", "bOwner", B, "oAOrg1", oeNotFound, "Spec R3 foreign object", "D04"))
	add(oeDel("D11", "aOwner", A, "oBOrg", oeNotFound, "Spec R3 foreign object", "D12"))
	add(oeDel("D12", "bOwner", B, "oBOrg", oeAllow, "Spec R3 matrix", ""))
	add(oeDel("D13", "g1MemberX", AG1, "oAOrg1", oeForbidden, "Spec R3 Organization mutation owner/admin only", "D07"))
	add(oeDel("D14", "g1Admin", AG1, "oAOrg1", oeForbidden, "WP r1 b3; AMBIGUITY A19", "D07"))
	add(oeDel("D15", "sessionExpired", A, "oAOrg1", oeUnauthenticated, "AUTH Errors expired session", "D01"))

	// ---- DELETE: Group-scoped objects ----
	add(oeDel("D20", "g1MemberX", AG1, "oG1X", oeAllow, "WP r1 b2: member deletes own upload with current Group access", ""))
	add(oeDel("D21", "g1MemberX", A, "oG1X", oeAllow, "WP r1 b2; Org-only context keeps current g1 access; AMBIGUITY A18", ""))
	add(oeDel("D22", "g1MemberX", AG1, "oG1Y", oeForbidden, "Spec R3: another user's upload denied to an ordinary member; readable -> forbidden", "D23"))
	add(oeDel("D23", "g1Admin", AG1, "oG1Y", oeAllow, "Spec R3 matrix group_admin deletes any object in scope", ""))
	add(oeDel("D24", "g1Owner", A, "oG1Y", oeAllow, "Spec R3 matrix group_owner", ""))
	add(oeDel("D25", "aAdmin", AG1, "oG1Y", oeNotFound, "G060 Spec R3 / ADR-008: a Group object is content; Organization owner/admin without a role in its Group -> not_found under either override value (was org_admin)", "D23"))
	add(oeDel("D26", "aOwner", A, "oG1X", oeNotFound, "G060 Spec R3 / ADR-008: a Group object is content; Organization owner/admin without a role in its Group -> not_found under either override value (was org_owner)", "D24"))
	add(oeDel("D27", "aAdmin", AG2, "oG1Y", oeNotFound, "Spec R3 selected Group restricts; AUTH pin (b) gate 3", "D23"))
	add(oeDel("D28", "g1MemberY", AG1, "oG1Y", oeAllow, "WP r1 b2 own upload", ""))
	add(oeDel("D29", "g1MemberY", AG1, "oG1X", oeForbidden, "Spec R3 another uploader's same-Group object", "D28"))
	add(oeDel("D30", "downgraded", AG1, "oG1D", oeAllow, "Spec R3/WP r1 b4: downgraded member keeps own-upload delete while Group access remains", ""))
	add(oeDel("D31", "downgraded", AG1, "oG1Y", oeForbidden, "Spec R3/WP r1 b4: downgrade removes deleting others' uploads", "D30"))
	add(oeDel("D32", "downgraded", A, "oG1X", oeForbidden, "Spec R3/WP r1 b4 (Org-only context)", "D30"))
	add(oeDel("D33", "removedG1", A, "oG1R", oeNotFound, "Spec R3/WP r1 b4: Group removal removes own-delete; no group.read -> not_found", "D34"))
	add(oeDel("D34", "g1Admin", AG1, "oG1R", oeAllow, "Spec R3 matrix group_admin", ""))
	add(oeDel("D35", "removedG1", oeStale(AG1), "oG1R", oeContextDenied, "AUTH Errors Group selectability gone", "D34"))
	add(oeDel("D36", "sessionRevoked", AG1, "oG1S", oeUnauthenticated, "Spec R3: revoking the session removes own-delete; AUTH Errors", "D37"))
	add(oeDel("D37", "g1Admin", AG1, "oG1S", oeAllow, "Spec R3 matrix group_admin", ""))
	add(oeDel("D38", "g2Member", A, "oG1X", oeNotFound, "Spec R3 shared profile / sibling Group", "D21"))
	add(oeDel("D39", "g2Member", AG2, "oG2M", oeAllow, "WP r1 b2 own upload", ""))
	add(oeDel("D40", "g1MemberX", AG1, "oG2M", oeNotFound, "Spec R3 sibling Group", "D39"))
	add(oeDel("D41", "g1MemberX", AG2, "oG2M", oeContextDenied, "AUTH Resolve g2 not selectable", "D39"))
	add(oeDel("D42", "aMember", A, "oG1X", oeNotFound, "AUTH matrix org_member no group.read", "D21"))
	add(oeDel("D43", "multi", A, "oG2M", oeAllow, "Spec R3 matrix group_admin of g2 deletes another's upload", ""))
	add(oeDel("D44", "multi", A, "oG1X", oeForbidden, "Spec R3 ordinary member of g1, not own upload; g2 admin role confers nothing on g1", "D43"))
	add(oeDel("D45", "multi", AG1, "oG2M", oeNotFound, "Spec R3 selected Group restricts", "D43"))
	add(oeDel("D46", "cross", AG1, "oG1C", oeAllow, "WP r1 b2 own upload in orgA/g1", ""))
	add(oeDel("D47", "cross", BGB, "oGBC", oeAllow, "WP r1 b2 own upload in orgB/gb", ""))
	add(oeDel("D48", "cross", A, "oGBC", oeNotFound, "Spec R3 same global UUID / own upload in another Organization; AMBIGUITY A6", "D47"))
	add(oeDel("D49", "cross", B, "oG1C", oeNotFound, "Spec R3; AMBIGUITY A6", "D46"))
	add(oeDel("D50", "cross", BGB, "oG1C", oeNotFound, "Spec R3; AMBIGUITY A6", "D46"))
	add(oeDel("D51", "bOwner", B, "oG1X", oeNotFound, "Spec R3 foreign tenant", "D26"))
	add(oeDel("D52", "aOwner", A, "oMissing", oeNotFound, "Spec R3 missing ID", "D26"))
	add(oeDel("D53", "aOwner", oeNoCtx(), "oG1X", oeContextDenied, "AUTH B1; AMBIGUITY A1", "D26"))
	add(oeDel("D54", "aOwner", oeFault(A, oeFaultScopeEnded), "oG1X", oeContextDenied, "AUTH B4", "D26"))
	add(oeDel("D55", "g1MemberX", oeFault(AG1, oeFaultSubjectMismatch), "oG1X", oeContextDenied, "AUTH B5/B6: own-upload delete via another user's context denied", "D20"))
	add(oeDel("D56", "disabled", A, "oG1X", oeUnauthenticated, "AUTH Errors disabled user", "D26"))
	add(oeDel("D57", "g1Owner", AG1, "oG1X", oeAllow, "Spec R3 matrix group_owner", ""))
	add(oeDel("D58", "aMember", B, "oMissing", oeContextDenied, "PLAN: invalid context with unknown object -> context_denied", "D04"))
	add(oeDel("D59", "g1MemberX", oeGroupCtx("orgA", "gb"), "oGBC", oeContextDenied, "AUTH Resolve: Group of another Organization", "D47"))
	add(oeDelRaw("D60", "g1MemberX", AG1, "not-a-uuid", oeInvalidInput, "Spec R4; AMBIGUITY A7", "D20"))
	add(oeDelRaw("D61", "g1MemberX", AG1, "orgs/10000000-0000-4000-8000-0000000000a1/50000000-0000-4000-8000-000000000005", oeInvalidInput, "Spec R3/R4: key/prefix never grants mutation and never reaches the adapter", "D20"))

	// ---- RLS SELECT (runtime login, no application predicates) ----
	add(oeSel("RS01", rt, "aOwner", A, oeAllow, oeSetAOrg, oeJoin(oeSetG1Ready, oeSetG2, oeSetB), "Spec R2 RLS mirrors read; Organization objects for org_owner; AMBIGUITY A4; G060 (Plan r0 A10): a Group object needs an explicit role in that exact Group under either Group-management setting", ""))
	add(oeSel("RS02", rt, "aOwner", AG1, oeAllow, oeSetAOrg, oeJoin(oeSetG1Ready, oeSetG2, oeSetB), "Spec R3 selected Group restricts Group rows; Org rows independent; AMBIGUITY A2/A4; G060 (Plan r0 A10): a Group object needs an explicit role in that exact Group under either Group-management setting", ""))
	add(oeSel("RS03", rt, "aAdmin", AG2, oeAllow, oeSetAOrg, oeJoin(oeSetG1All, oeSetG2, oeSetB), "Spec R3; AUTH pin (b); G060 (Plan r0 A10): a Group object needs an explicit role in that exact Group under either Group-management setting", ""))
	add(oeSel("RS04", rt, "aMember", A, oeAllow, oeSetAOrg, oeJoin(oeSetGroupsAllA, oeSetB), "Spec R3 org_member reads Organization objects only", ""))
	add(oeSel("RS05", rt, "g1MemberX", A, oeAllow, oeJoin(oeSetAOrg, oeSetG1Ready), oeJoin(oeSetG2, oeSetB), "Spec R3; AMBIGUITY A18", ""))
	add(oeSel("RS06", rt, "g1MemberX", AG1, oeAllow, oeJoin(oeSetAOrg, oeSetG1Ready), oeJoin(oeSetG2, oeSetB), "Spec R3", ""))
	add(oeSel("RS07", rt, "g1MemberX", AG2, oeRLSNoRows, oeSetNothingSeen, oeSetAllStored, "DBISO: unselectable Group -> rls_org_role NULL -> every tenant row denied", "RS06"))
	add(oeSel("RS08", rt, "g2Member", A, oeAllow, oeJoin(oeSetAOrg, oeSetG2), oeJoin(oeSetG1All, oeSetB), "Spec R3 shared profile never reveals sibling objects", ""))
	add(oeSel("RS09", rt, "multi", A, oeAllow, oeJoin(oeSetAOrg, oeSetG1Ready, oeSetG2), oeSetB, "Spec R3 union of own Group paths in Org-only context", ""))
	add(oeSel("RS10", rt, "multi", AG1, oeAllow, oeJoin(oeSetAOrg, oeSetG1Ready), oeJoin(oeSetG2, oeSetB), "Spec R3 selected Group restricts", ""))
	add(oeSel("RS11", rt, "removedG1", A, oeAllow, oeSetAOrg, oeJoin(oeSetGroupsAllA, oeSetB), "Spec R3: removed uploader sees neither own oG1R nor own pending oG1RPending", ""))
	add(oeSel("RS12", rt, "removedG1", AG1, oeRLSNoRows, oeSetNothingSeen, oeSetAllStored, "DBISO: Group no longer selectable", "RS11"))
	add(oeSel("RS13", rt, "revokedOrgAdmin", A, oeRLSNoRows, oeSetNothingSeen, oeSetAllStored, "DBISO: no Organization binding -> no tenant rows, including own upload", "RS04"))
	add(oeSel("RS14", rt, "cross", A, oeAllow, oeJoin(oeSetAOrg, oeSetG1Ready), oeJoin(oeSetG2, oeSetB), "Spec R3 same global UUID: own orgB upload oGBC invisible in orgA context", ""))
	add(oeSel("RS15", rt, "cross", BGB, oeAllow, oeSetB, oeSetAllA, "Spec R3: own orgA upload oG1C invisible in orgB context", ""))
	add(oeSel("RS16", rt, "bOwner", B, oeAllow, []string{"oBOrg"}, oeJoin([]string{"oGBC"}, oeSetAllA), "Spec R3 foreign tenant isolation; G060 (Plan r0 A10): a Group object needs an explicit role in that exact Group under either Group-management setting (bOwner holds no gB role)", ""))
	add(oeSel("RS17", rt, "aOwner", B, oeRLSNoRows, oeSetNothingSeen, oeSetAllStored, "DBISO: foreign Organization selection denies every tenant row", "RS01"))
	add(oeSel("RS18", rt, "aOwner", oeNoCtx(), oeRLSNoRows, oeSetNothingSeen, oeSetAllStored, "DBISO EMPTY shape / PLAN 'with no context it sees nothing'; AMBIGUITY A1", "RS01"))
	add(oeSel("RS19", rt, "formerOrgAdmin", A, oeAllow, oeSetAOrg, oeJoin(oeSetGroupsAllA, oeSetB), "Spec R3 former admin is an ordinary member", ""))
	add(oeSel("RS20", rt, "downgraded", AG1, oeAllow, oeJoin(oeSetAOrg, oeSetG1Ready), oeJoin(oeSetG2, oeSetB), "Spec R3 downgraded keeps member read", ""))
	add(oeSel("RS21", rt, "aOwner", oeGroupCtx("orgA", "gb"), oeRLSNoRows, oeSetNothingSeen, oeSetAllStored, "DBISO: selected Group outside the Organization", "RS01"))
	add(oeSel("RS22", rt, "aOwner", oeRawCtx(oeRawSettings{UserID: "40000000-0000-4000-8000-000000000001", OrganizationID: "not-a-uuid"}), oeSQL22P02, nil, nil, "DBISO parsers: malformed UUID raises 22P02 when a candidate row is evaluated", "RS01"))
	add(oeSel("RS23", rt, "aOwner", oeRawCtx(oeRawSettings{UserID: "40000000-0000-4000-8000-000000000001", OrganizationID: "00000000-0000-0000-0000-000000000000"}), oeSQL22023, nil, nil, "DBISO parsers: zero UUID raises 22023", "RS01"))
	add(oeSel("RS24", rt, "aOwner", oeRawCtx(oeRawSettings{UserID: "40000000-0000-4000-8000-000000000001", GroupID: "20000000-0000-4000-8000-0000000000a1"}), oeSQL22023, nil, nil, "DBISO parsers: Group without Organization is not a valid shape -> 22023", "RS02"))

	// ---- RLS INSERT (runtime) ----
	add(oeIns("RI01", rt, "g1MemberX", AG1, oeR("orgA", "g1", "g1MemberX", P), oeAllow, "WP r1 b1 / Spec R2: creator = context user, initial pending state", ""))
	add(oeIns("RI02", rt, "g1MemberX", A, oeR("orgA", "g1", "g1MemberX", P), oeAllow, "WP r1 b1; AMBIGUITY A18", ""))
	add(oeIns("RI03", rt, "g1MemberX", AG1, oeR("orgA", "g1", "g1MemberY", P), oeRLSWithCheck, "Spec R2: caller-supplied/forged creator rejected; AC4 forged creator", "RI01"))
	add(oeIns("RI04", rt, "g1MemberX", AG1, oeR("orgA", "g1", "g1MemberX", R), oeRLSStateRejected, "Spec R7 no READY before durable upload; PLAN wrong state; AMBIGUITY A11", "RI01"))
	add(oeIns("RI05", rt, "g1MemberX", A, oeR("orgA", "g2", "g1MemberX", P), oeRLSWithCheck, "PLAN wrong Group; Spec R3 sibling Group", "RI02"))
	add(oeIns("RI06", rt, "g1MemberX", AG1, oeR("orgB", "gb", "g1MemberX", P), oeRLSWithCheck, "PLAN wrong Organization", "RI01"))
	add(oeIns("RI07", rt, "g1MemberX", A, oeR("orgA", "", "g1MemberX", P), oeRLSWithCheck, "Spec R3 Organization Put owner/admin only", "RI08"))
	add(oeIns("RI08", rt, "aAdmin", A, oeR("orgA", "", "aAdmin", P), oeAllow, "Spec R3 matrix Organization Put org_admin", ""))
	add(oeIns("RI09", rt, "aAdmin", AG1, oeR("orgA", "g1", "aAdmin", P), oeRLSWithCheck, "Spec R3 matrix: formerly org_admin Group Put within selected Group; G060 (Plan r0 A10): a Group object needs an explicit role in that exact Group under either Group-management setting", ""))
	add(oeIns("RI10", rt, "aAdmin", AG2, oeR("orgA", "g1", "aAdmin", P), oeRLSWithCheck, "Spec R3 within selected scope", "RI09"))
	add(oeIns("RI11", rt, "aMember", A, oeR("orgA", "", "aMember", P), oeRLSWithCheck, "Spec R3 Organization Put owner/admin only", "RI08"))
	add(oeIns("RI12", rt, "aMember", A, oeR("orgA", "g1", "aMember", P), oeRLSWithCheck, "Spec R3 no Group path", "RI02"))
	add(oeIns("RI13", rt, "removedG1", A, oeR("orgA", "g1", "removedG1", P), oeRLSWithCheck, "WP r1 b4 removed member", "RI02"))
	add(oeIns("RI14", rt, "formerOrgAdmin", A, oeR("orgA", "", "formerOrgAdmin", P), oeRLSWithCheck, "Spec R3 former admin", "RI08"))
	add(oeIns("RI15", rt, "g1Owner", A, oeR("orgA", "", "g1Owner", P), oeRLSWithCheck, "WP r1 b3; AMBIGUITY A19", "RI08"))
	add(oeIns("RI16", rt, "g1Owner", AG1, oeR("orgA", "g1", "g1Owner", P), oeAllow, "Spec R3 matrix group_owner", ""))
	add(oeIns("RI17", rt, "cross", A, oeR("orgB", "gb", "cross", P), oeRLSWithCheck, "Spec R3 same global UUID; row outside active Organization", "RI18"))
	add(oeIns("RI18", rt, "cross", BGB, oeR("orgB", "gb", "cross", P), oeAllow, "WP r1 b1 in orgB", ""))
	add(oeIns("RI19", rt, "aOwner", oeNoCtx(), oeR("orgA", "", "aOwner", P), oeRLSWithCheck, "DBISO EMPTY shape: no tenant authority; AMBIGUITY A1", "RI20"))
	add(oeIns("RI20", rt, "aOwner", A, oeR("orgA", "", "aOwner", P), oeAllow, "Spec R3 matrix org_owner", ""))
	add(oeIns("RI21", rt, "aOwner", B, oeR("orgB", "", "aOwner", P), oeRLSWithCheck, "DBISO foreign selection", "RI20"))
	add(oeIns("RI22", rt, "aAdmin", A, oeR("orgA", "", "aOwner", P), oeRLSWithCheck, "Spec R2: admins cannot forge creator either", "RI08"))
	add(oeIns("RI23", rt, "g1MemberX", AG1, oeR("orgA", "g1", "g1MemberX", D), oeRLSStateRejected, "PLAN wrong state; AMBIGUITY A11", "RI01"))

	// ---- RLS state transitions (runtime) ----
	add(oeTrans("RT01", rt, "g1MemberX", AG1, "oG1Pending", R, oeAllow, "Spec R7 finalize by creator with current Group Put authority; AMBIGUITY A12", ""))
	add(oeTrans("RT02", rt, "removedG1", A, "oG1RPending", R, oeRLSZeroRows, "Spec R7: revocation after preparation cannot publish; row out of scope", "RT01"))
	add(oeTrans("RT03", rt, "g1MemberX", AG1, "oG1X", D, oeAllow, "WP r1 b2 own-upload logical delete", ""))
	add(oeTrans("RT04", rt, "g1MemberX", AG1, "oG1Y", D, oeRLSDeniedNoChange, "Spec R3 another uploader; WP r1 runtime RLS distinction; AMBIGUITY A10", "RT03"))
	add(oeTrans("RT05", rt, "g1Admin", AG1, "oG1Y", D, oeAllow, "Spec R3 matrix group_admin", ""))
	add(oeTrans("RT06", rt, "aAdmin", A, "oAOrg1", D, oeAllow, "Spec R3 matrix Organization Delete org_admin", ""))
	add(oeTrans("RT07", rt, "aMember", A, "oAOrg1", D, oeRLSDeniedNoChange, "Spec R3 Organization Delete owner/admin only; AMBIGUITY A10", "RT06"))
	add(oeTrans("RT08", rt, "formerOrgAdmin", A, "oAOrgFormer", D, oeRLSDeniedNoChange, "Spec R3 former-admin uploader denied; AMBIGUITY A10", "RT09"))
	add(oeTrans("RT09", rt, "aOwner", A, "oAOrgFormer", D, oeAllow, "Spec R3 matrix org_owner", ""))
	add(oeTrans("RT10", rt, "downgraded", AG1, "oG1D", D, oeAllow, "WP r1 b4 downgraded keeps own delete", ""))
	add(oeTrans("RT11", rt, "downgraded", AG1, "oG1Y", D, oeRLSDeniedNoChange, "WP r1 b4 downgraded loses others' delete; AMBIGUITY A10", "RT10"))
	add(oeTrans("RT12", rt, "aOwner", A, "oBOrg", D, oeRLSZeroRows, "Spec R3 foreign row invisible", "RT06"))
	add(oeTrans("RT13", rt, "cross", A, "oGBC", D, oeRLSZeroRows, "Spec R3 same global UUID; own orgB row invisible from orgA", "RT14"))
	add(oeTrans("RT14", rt, "cross", BGB, "oGBC", D, oeAllow, "WP r1 b2 own upload in orgB", ""))
	add(oeTrans("RT15", rt, "g2Member", A, "oG1X", D, oeRLSZeroRows, "Spec R3 sibling row invisible", "RT03"))
	add(oeTrans("RT16", rt, "g1MemberX", AG1, "oG1X", P, oeRLSDeniedNoChange, "Spec R2/R7 illegal ready->pending; AMBIGUITY A10", "RT03"))
	add(oeTrans("RT17", rt, "aAdmin", A, "oG1Deleting", R, oeRLSDeniedNoChange, "Spec R7 deleted object never returns to READY; AMBIGUITY A10", "RT06"))
	add(oeTrans("RT18", rt, "aAdmin", AG2, "oG1Y", D, oeRLSZeroRows, "Spec R3 outside selected Group: row invisible", "RT05"))
	add(oeTrans("RT19", rt, "g1MemberX", AG2, "oG1X", D, oeRLSZeroRows, "DBISO unselectable Group context denies every tenant row", "RT03"))
	add(oeTrans("RT20", rt, "aAdmin", AG1, "oG1X", D, oeRLSZeroRows, "Spec R3 matrix: formerly org_admin in selected Group; G060 (Plan r0 A10): a Group object needs an explicit role in that exact Group under either Group-management setting (the row is outside USING)", ""))
	add(oeTrans("RT21", rt, "removedG1", A, "oG1R", D, oeRLSZeroRows, "WP r1 b4 removed uploader loses own delete", "RT03"))

	// ---- RLS immutable columns (runtime, authorized updater) ----
	add(oeImm("RU01", "aAdmin", AG1, "oG1X", "id", "<fresh>", oeRLSImmutable, "Spec R2 UUID immutable; AMBIGUITY A10", "RT20"))
	add(oeImm("RU02", "aAdmin", AG1, "oG1X", "organization_id", "orgB", oeRLSImmutable, "Spec R2 owning Organization immutable, no reparenting", "RT20"))
	add(oeImm("RU03", "aAdmin", AG1, "oG1X", "group_id", "g2", oeRLSImmutable, "Spec R2 owning Group immutable, no reparenting", "RT20"))
	add(oeImm("RU04", "aAdmin", AG1, "oG1X", "creator_id", "aAdmin", oeRLSImmutable, "Spec R2 creator identity immutable", "RT20"))
	add(oeImm("RU05", "aAdmin", AG1, "oG1X", "object_key", "<fresh>", oeRLSImmutableGenerated, "Spec R2 key immutable; forged key cannot redirect; generated storage key 428C9 per Codex ruling a466be31 (r2)", "RT20"))
	add(oeImm("RU06", "g1MemberX", AG1, "oG1X", "creator_id", "g1MemberY", oeRLSImmutable, "Spec R2 creator immutable even for the creator", "RT03"))
	add(oeImm("RU07", "g1MemberY", AG1, "oG1X", "creator_id", "g1MemberY", oeRLSDeniedNoChange, "Spec R2/R3: re-attributing creator cannot create an own-delete path; AMBIGUITY A10", "RT03"))

	// ---- RLS DELETE (runtime): always a privilege error ----
	add(oeRLSDel("RD01", "g1MemberX", AG1, "oG1X", oeSQLPrivilege, "PLAN OBJ-RLS: DELETE always 42501; logical deletion is the transition", "RT03"))
	add(oeRLSDel("RD02", "aOwner", A, "oAOrg1", oeSQLPrivilege, "PLAN OBJ-RLS: DELETE always 42501", "RT06"))
	add(oeRLSDel("RD03", "aAdmin", oeNoCtx(), "oG1X", oeSQLPrivilege, "PLAN OBJ-RLS: privilege check precedes RLS", "RT20"))

	// ---- FORCE owner (PLAN OBJ-RLS correction 6) ----
	add(oeSel("RF01", ow, "aOwner", A, oeAllow, oeSetAOrg, oeJoin(oeSetG1Ready, oeSetG2, oeSetB), "PLAN FORCE owner: sees exactly orgA's in-scope objects; AMBIGUITY A16; G060 (Plan r0 A10): a Group object needs an explicit role in that exact Group under either Group-management setting", ""))
	add(oeSel("RF02", ow, "aOwner", AG1, oeAllow, oeSetAOrg, oeJoin(oeSetG1Ready, oeSetG2, oeSetB), "PLAN FORCE owner: selected g1 does not see g2; G060 (Plan r0 A10): a Group object needs an explicit role in that exact Group under either Group-management setting", ""))
	add(oeSel("RF03", ow, "aOwner", oeNoCtx(), oeRLSNoRows, oeSetNothingSeen, oeSetAllStored, "PLAN FORCE owner: no context sees nothing", "RF01"))
	add(oeIns("RF04", ow, "aOwner", A, oeR("orgB", "", "aOwner", P), oeRLSWithCheck, "PLAN FORCE owner: INSERT of an orgB row fails WITH CHECK", "RF05"))
	add(oeIns("RF05", ow, "aOwner", A, oeR("orgA", "", "aOwner", P), oeAllow, "FORCE: same policies as runtime; AMBIGUITY A16", ""))
	add(oeTrans("RF06", ow, "aOwner", A, "oBOrg", D, oeRLSZeroRows, "PLAN FORCE owner: UPDATE of orgB rows affects 0 rows", "RF07"))
	add(oeTrans("RF07", ow, "aOwner", A, "oAOrg1", D, oeAllow, "FORCE: in-scope owner transition; AMBIGUITY A16", ""))

	return cs
}
