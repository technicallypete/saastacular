package tenancy

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// SQL of the model. Every statement is parameterized and names the owning Organization explicitly. Writes target
// organizations, groups, role_bindings, business_profiles and group_business_profiles only; role values are written to
// role_bindings only, and membership and access-path reads go through the read-only projections only (U-BOUNDARY
// parses these constants).
const (
	sqlOrganizationExists = `SELECT EXISTS (SELECT 1 FROM app.organizations WHERE id = $1)`
	sqlGroupExists        = `SELECT EXISTS (SELECT 1 FROM app.groups WHERE organization_id = $1 AND id = $2)`
	sqlUserExists         = `SELECT EXISTS (SELECT 1 FROM app.users WHERE id = $1)`
	// G019 (F-01 refinement a): CreateOrganizationWithOwner's active-status owner lock before the bootstrap writes; FOR
	// SHARE conflicts with identity's FOR UPDATE (DisableUser, FinalizeRetirement), so the status read after a wait is the
	// committed one. It names the user because the Organization does not exist yet.
	sqlLockOwnerStatus = `SELECT status FROM app.users WHERE id = $1 FOR SHARE`

	sqlInsertOrganization = `INSERT INTO app.organizations (name) VALUES ($1) RETURNING id, name, created_at`
	sqlSelectOrganization = `SELECT id, name, created_at FROM app.organizations WHERE id = $1`
	sqlRenameOrganization = `UPDATE app.organizations SET name = $2 WHERE id = $1 RETURNING id, name, created_at`

	groupColumns     = `id, organization_id, name, slug, created_at`
	sqlInsertGroup   = `INSERT INTO app.groups (organization_id, name, slug) VALUES ($1, $2, $3) RETURNING ` + groupColumns
	sqlSelectGroup   = `SELECT ` + groupColumns + ` FROM app.groups WHERE organization_id = $1 AND id = $2`
	sqlSelectBySlug  = `SELECT ` + groupColumns + ` FROM app.groups WHERE organization_id = $1 AND slug = $2`
	sqlListGroups    = `SELECT ` + groupColumns + ` FROM app.groups WHERE organization_id = $1 ORDER BY slug`
	sqlRenameGroup   = `UPDATE app.groups SET name = $3 WHERE organization_id = $1 AND id = $2 RETURNING ` + groupColumns
	sqlChangeSlug    = `UPDATE app.groups SET slug = $3 WHERE organization_id = $1 AND id = $2 RETURNING ` + groupColumns
	sqlSlugsInUse    = `SELECT slug FROM app.groups WHERE organization_id = $1 AND slug = ANY ($2)`
	bindingColumns   = `id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source, created_at`
	sqlInsertBinding = `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) ` +
		`VALUES ($1, $2, $3, $4, $5, $6, 'membership') RETURNING ` + bindingColumns
	// $4 is the scope ID: the Organization ID for Organization scope, the Group ID for Group scope.
	sqlChangeRole = `UPDATE app.role_bindings SET role = $5 ` +
		`WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = $3 AND scope_id = $4 AND source = 'membership' ` +
		`RETURNING ` + bindingColumns
	sqlRemoveBinding = `DELETE FROM app.role_bindings ` +
		`WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = $3 AND scope_id = $4 AND source = 'membership'`

	sqlGetOrgMembership   = `SELECT user_id, organization_id, role FROM app.organization_memberships WHERE organization_id = $1 AND user_id = $2`
	sqlListOrgMemberships = `SELECT user_id, organization_id, role FROM app.organization_memberships WHERE organization_id = $1 ORDER BY user_id`
	sqlGetGroupMembership = `SELECT user_id, organization_id, group_id, role FROM app.group_memberships ` +
		`WHERE organization_id = $1 AND group_id = $2 AND user_id = $3`
	sqlListGroupMemberships = `SELECT user_id, organization_id, group_id, role FROM app.group_memberships ` +
		`WHERE organization_id = $1 AND group_id = $2 ORDER BY user_id`

	// G006 (migration 00004_business_profiles.sql): BusinessProfiles, their Group associations and the read-only access
	// paths. Writes target business_profiles and group_business_profiles only; access facts are read through the
	// projection only.
	sqlBusinessProfileExists = `SELECT EXISTS (SELECT 1 FROM app.business_profiles WHERE organization_id = $1 AND id = $2)`
	sqlGroupHasAssociation   = `SELECT EXISTS (SELECT 1 FROM app.group_business_profiles WHERE organization_id = $1 AND group_id = $2)`
	businessProfileColumns   = `id, organization_id, display_name, legal_name, description, industry, website, created_at`
	sqlInsertBusinessProfile = `INSERT INTO app.business_profiles (organization_id, display_name, legal_name, description, industry, website) ` +
		`VALUES ($1, $2, $3, $4, $5, $6) RETURNING ` + businessProfileColumns
	sqlSelectBusinessProfile = `SELECT ` + businessProfileColumns + ` FROM app.business_profiles WHERE organization_id = $1 AND id = $2`
	sqlUpdateBusinessProfile = `UPDATE app.business_profiles SET display_name = $3, legal_name = $4, description = $5, industry = $6, website = $7 ` +
		`WHERE organization_id = $1 AND id = $2 RETURNING ` + businessProfileColumns
	associationColumns   = `id, organization_id, group_id, business_profile_id, created_at`
	sqlInsertAssociation = `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1, $2, $3) ` +
		`RETURNING ` + associationColumns
	sqlSelectGroupAssociation = `SELECT a.id, a.organization_id, a.group_id, a.business_profile_id, a.created_at, p.id, p.organization_id, ` +
		`p.display_name, p.legal_name, p.description, p.industry, p.website, p.created_at FROM app.group_business_profiles a ` +
		`JOIN app.business_profiles p ON p.organization_id = a.organization_id AND p.id = a.business_profile_id ` +
		`WHERE a.organization_id = $1 AND a.group_id = $2`
	sqlListProfileAssociations = `SELECT ` + associationColumns + ` FROM app.group_business_profiles ` +
		`WHERE organization_id = $1 AND business_profile_id = $2 ORDER BY group_id`
	sqlDeleteAssociation      = `DELETE FROM app.group_business_profiles WHERE organization_id = $1 AND group_id = $2 AND business_profile_id = $3`
	accessPathColumns         = `user_id, organization_id, business_profile_id, group_id, group_role`
	sqlListProfileAccessPaths = `SELECT ` + accessPathColumns + ` FROM app.business_profile_access_paths ` +
		`WHERE organization_id = $1 AND business_profile_id = $2 ORDER BY user_id, group_id`
	sqlListUserAccessPaths = `SELECT ` + accessPathColumns + ` FROM app.business_profile_access_paths ` +
		`WHERE organization_id = $1 AND user_id = $2 ORDER BY business_profile_id, group_id`

	// G007 (migration 00005_organization_policy.sql and docs/foundation/AUTHORIZATION.md): the fact reads of ReadFacts in
	// the order of the authorization contract, each in a snapshot form and a locking form, the guarded writes and the
	// Group row lock of UnlinkGroupBusinessProfile. A set is read by one statement (= ANY), never one per Group.
	organizationFactColumns      = `id, name, created_at, group_admins_edit_business_profiles, group_management_override`
	sqlFactOrganization          = `SELECT ` + organizationFactColumns + ` FROM app.organizations WHERE id = $1`
	sqlFactOrganizationShare     = sqlFactOrganization + ` FOR SHARE`
	sqlFactOrganizationExclusive = sqlFactOrganization + ` FOR NO KEY UPDATE`
	// The subject's Organization binding and its Group bindings (candidate Groups, or all of its Groups in the
	// Organization). Snapshot reads go through the projections. A row lock needs UPDATE on a column of the relation it
	// names, and the projections are SELECT-only, so the three locking forms name role_bindings itself with exactly the
	// projections' filters (scope_type and the membership source): the same columns, rows, lock mode and order, through
	// the existing UPDATE(role) grant (Plan r1 §11 fallback; Codex review 432d5f17 §1). They are the only direct reads of
	// role_bindings; U-BOUNDARY pins them by name and exact text.
	sqlFactOrganizationBinding     = `SELECT organization_id, role FROM app.organization_memberships WHERE organization_id = $1 AND user_id = $2`
	sqlFactOrganizationBindingLock = `SELECT organization_id, role FROM app.role_bindings ` +
		`WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = 'organization' AND source = 'membership' FOR SHARE`
	sqlFactGroupBindings = `SELECT organization_id, group_id, role FROM app.group_memberships ` +
		`WHERE organization_id = $1 AND user_id = $2 AND group_id = ANY ($3) ORDER BY group_id`
	sqlFactGroupBindingsLock = `SELECT organization_id, group_id, role FROM app.role_bindings ` +
		`WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = 'group' AND source = 'membership' AND group_id = ANY ($3) ` +
		`ORDER BY group_id FOR SHARE`
	sqlFactAllGroupBindings = `SELECT organization_id, group_id, role FROM app.group_memberships ` +
		`WHERE organization_id = $1 AND user_id = $2 ORDER BY group_id`
	sqlFactAllGroupBindingsLock = `SELECT organization_id, group_id, role FROM app.role_bindings ` +
		`WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = 'group' AND source = 'membership' ORDER BY group_id FOR SHARE`
	// Group rows: the update target alone, exclusively and first; then the other candidates, shared, in id order.
	sqlFactTargetGroupLock = sqlSelectGroup + ` FOR NO KEY UPDATE`
	sqlFactGroups          = `SELECT ` + groupColumns + ` FROM app.groups WHERE organization_id = $1 AND id = ANY ($2) ORDER BY id`
	sqlFactGroupsLock      = sqlFactGroups + ` FOR SHARE`
	// Associations of the read Groups: always a plain read in a new statement (runtime cannot row-lock associations;
	// the Group row is their lock surrogate).
	sqlFactAssociations = `SELECT group_id, business_profile_id FROM app.group_business_profiles ` +
		`WHERE organization_id = $1 AND group_id = ANY ($2) ORDER BY group_id`
	sqlFactAllProfiles = `SELECT ` + businessProfileColumns + ` FROM app.business_profiles WHERE organization_id = $1 ORDER BY id`
	// The one policy write: an absolute assignment of the flag.
	sqlSetOrganizationPolicy = `UPDATE app.organizations SET group_admins_edit_business_profiles = $2 WHERE id = $1 ` +
		`RETURNING group_admins_edit_business_profiles`
	// G060: the owner-only override write (00017; the guard and the capture trigger run on it), an absolute assignment.
	sqlSetGroupManagementOverride = `UPDATE app.organizations SET group_management_override = $2 WHERE id = $1 ` +
		`RETURNING group_management_override`
	// UnlinkGroupBusinessProfile's Group precheck: it locks the Group row before the DELETE, so an unlink conflicts with
	// every guarded write that relies on that Group's association.
	sqlLockGroupForUnlink = `SELECT id FROM app.groups WHERE organization_id = $1 AND id = $2 FOR NO KEY UPDATE`

	// G014 (docs/foundation/MCP.md): one page of the Groups the subject may select in Organization-only context, ordered
	// by Group UUID after a cursor. An Organization owner or admin may select every Group of the Organization (one range
	// read of groups by its (organization_id, id) key); any other member exactly the Groups it holds a binding in (one
	// read of the group_memberships projection joined to its Groups). One statement per page, never one per Group.
	sqlSelectableGroupsManager = `SELECT ` + groupColumns + ` FROM app.groups WHERE organization_id = $1 AND id > $2 ORDER BY id LIMIT $3`
	sqlSelectableGroupsMember  = `SELECT g.id, g.organization_id, g.name, g.slug, g.created_at FROM app.group_memberships m ` +
		`JOIN app.groups g ON g.organization_id = m.organization_id AND g.id = m.group_id ` +
		`WHERE m.organization_id = $1 AND m.user_id = $2 AND g.id > $3 ORDER BY g.id LIMIT $4`

	// G008 (00006_row_level_security.sql, docs/foundation/DATABASE-ISOLATION.md): the one context statement. It sets
	// all four transaction-local fields ($4 is always '' here: no login binding in tenant shape) and returns one row only
	// when none of them was already set on the connection; no row means prior authority, and the caller discards the
	// connection. core/internal/identity holds a byte-identical copy (U-RLS-BOUNDARY).
	//
	// G011 (00008_audit_events.sql, docs/foundation/AUDIT.md): the same statement sets the two transaction-local audit
	// correlation settings ($5 request ID, $6 trace ID; diagnostic only, never compared or used in a decision) and resets
	// the revocation gap marker to the constant '', so a stale session-level value is never read as a gap. They confer no
	// authority and are outside the prior-context guard.
	sqlInitContext = `SELECT pg_catalog.set_config('app.user_id', $1, true), pg_catalog.set_config('app.organization_id', $2, true),
       pg_catalog.set_config('app.group_id', $3, true), pg_catalog.set_config('app.login_binding', $4, true),
       pg_catalog.set_config('audit.request_id', $5, true), pg_catalog.set_config('audit.trace_id', $6, true),
       pg_catalog.set_config('audit.revocation_gap', '', true)
WHERE coalesce(pg_catalog.current_setting('app.user_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.organization_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.group_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.login_binding', true), '') = ''`
)

// G018 (Plan r5 §4.7): the keyed onboarding inserts, both without RETURNING. S-G1: the creator holds no binding on the new
// Group until the 00014 bootstrap trigger has run, so the row is read afterwards with sqlSelectGroup. S-P1: an unlinked
// profile is invisible to a Group-derived creator.
const (
	sqlInsertGroupKeyed           = `INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1, $2, $3, $4)`
	sqlInsertBusinessProfileKeyed = `INSERT INTO app.business_profiles (id, organization_id, display_name, legal_name, description, industry, website) ` +
		`VALUES ($1, $2, $3, $4, $5, $6, $7)`
	// G018 (Plan r5 §1.5): the onboarding list pages, one keyset statement each (ascending id after a cursor, limit+1
	// rows); row-level security restricts them to the rows the context may read. The slug filter of ListGroups reuses
	// sqlSelectBySlug.
	sqlFactGroupsPage   = `SELECT ` + groupColumns + ` FROM app.groups WHERE organization_id = $1 AND id > $2 ORDER BY id LIMIT $3`
	sqlFactProfilesPage = `SELECT ` + businessProfileColumns + ` FROM app.business_profiles WHERE organization_id = $1 AND id > $2 ORDER BY id LIMIT $3`
)

// G019 (00015_membership_invitations.sql; Plan r1 §1.4, §4.3, §4.4): membership and ownership administration. The
// runtime role reads exactly the eight factBindingColumns of role_bindings (00015 adds id; created_at stays ungranted),
// so every runtime statement over role_bindings names those columns only.
const (
	factBindingColumns = `id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source`
	// The Subject fact (ReadFacts): the subject's and, for a transfer, the actor's binding in the one fact scope, in one
	// statement ordered by id, FOR NO KEY UPDATE (admitted by role_bindings_update's USING; an invisible row is an absent
	// fact). $2 is the subject set, $3/$4 the scope.
	sqlSubjectBindingsLock = `SELECT ` + factBindingColumns + ` FROM app.role_bindings ` +
		`WHERE organization_id = $1 AND subject_user_id = ANY ($2) AND scope_type = $3 AND scope_id = $4 AND source = 'membership' ` +
		`ORDER BY id FOR NO KEY UPDATE`
	// The guarded writes, keyed by the binding id the Subject fact locked; app.role_bindings_authority_guard re-enforces
	// owner authority, promotion eligibility, the dependent-grant rule and last-owner cardinality on each statement.
	sqlChangeRoleByID = `UPDATE app.role_bindings SET role = $3 WHERE organization_id = $1 AND id = $2 AND source = 'membership' ` +
		`RETURNING ` + factBindingColumns
	sqlRemoveBindingByID = `DELETE FROM app.role_bindings WHERE organization_id = $1 AND id = $2 AND source = 'membership'`
	// The definer count reader (answers 0 for a scope the context cannot see).
	sqlOwnerCount = `SELECT app.rls_owner_count($1, $2)`
	// The context user's sole-owned scopes (definer reader; IDENTITY-shape context; identifiers only).
	sqlSoleOwnerScopes = `SELECT organization_id, scope_type, scope_id FROM app.rls_sole_owner_scopes()`
	// The context user's own Organizations for trusted retirement (definer reader; IDENTITY-shape context; identifiers
	// only; ascending keyset page after $1, NULL: from the start, at most $2 <= 100 rows; Result addendum A1-01).
	sqlOwnOrganizationsPage = `SELECT organization_id FROM app.rls_own_organizations($1, $2)`
	// Member pages: one keyset statement each over the projection joined to the co-member users rows (users_select);
	// $3 (or $4) '' lists every member, a role lists only the enabled holders of that role (owner_candidates).
	sqlOrganizationMembersPage = `SELECT m.user_id, u.display_name, u.status, m.role FROM app.organization_memberships m ` +
		`JOIN app.users u ON u.id = m.user_id WHERE m.organization_id = $1 AND m.user_id > $2 ` +
		`AND ($3 = '' OR (m.role = $3 AND u.status = 'active')) ORDER BY m.user_id LIMIT $4`
	sqlGroupMembersPage = `SELECT m.user_id, u.display_name, u.status, m.role FROM app.group_memberships m ` +
		`JOIN app.users u ON u.id = m.user_id WHERE m.organization_id = $1 AND m.group_id = $2 AND m.user_id > $3 ` +
		`AND ($4 = '' OR (m.role = $4 AND u.status = 'active')) ORDER BY m.user_id LIMIT $5`
	// The least-disclosing Group invitation picker (D-13): the definer reader returns enabled Organization members that
	// hold no binding in the Group, after the cursor (NULL: from the start), at most least($3, 101) rows.
	sqlInvitationCandidatesPage = `SELECT user_id, display_name FROM app.rls_group_invitation_candidates($1, $2, $3)`
)

// G019 invitations (00015; Plan r1 §1.4 "Invitations", r2 addendum D-14a). The runtime role reads every column except
// token_digest; the recipient forms select NULL in place of intended_address (the address reaches only the issuing
// scope's administrators). $1 is always the Organization.
const (
	invitationColumns = `id, organization_id, scope_type, scope_id, group_id, role, issuer_user_id, intended_user_id, ` +
		`intended_address, created_at, expires_at, revoked_at, consumed_at, consumed_by_user_id, consumed_binding_id`
	invitationRecipientColumns = `id, organization_id, scope_type, scope_id, group_id, role, issuer_user_id, intended_user_id, ` +
		`NULL::text, created_at, expires_at, revoked_at, consumed_at, consumed_by_user_id, consumed_binding_id`
	// One keyed INSERT without RETURNING (the id is core-generated); expires_at is computed from the database clock so
	// the invitations_lifetime CHECK (against created_at DEFAULT now()) never sees application clock skew. The BEFORE
	// INSERT trigger app.invitations_bound repeats the scope lock and the pending bound.
	sqlInsertInvitation = `INSERT INTO app.invitations (id, organization_id, scope_type, scope_id, group_id, role, issuer_user_id, ` +
		`intended_user_id, intended_address, token_digest, expires_at) ` +
		`VALUES ($2, $1, $3, $4, $5, $6, $7, $8, $9, $10, now() + make_interval(secs => $11))`
	sqlSelectInvitation = `SELECT ` + invitationColumns + ` FROM app.invitations WHERE organization_id = $1 AND id = $2`
	// Revocation: only a pending row (neither revoked nor consumed); zero rows are disambiguated by sqlInvitationExists.
	sqlRevokeInvitation = `UPDATE app.invitations SET revoked_at = now() WHERE organization_id = $1 AND id = $2 ` +
		`AND revoked_at IS NULL AND consumed_at IS NULL RETURNING ` + invitationColumns
	sqlInvitationExists = `SELECT EXISTS (SELECT 1 FROM app.invitations WHERE organization_id = $1 AND id = $2)`
	// Keyset pages (cursor = invitation id, limit+1 rows): a scope's invitations (invitations_scope_idx) and a recipient's
	// open invitations (invitations_recipient_idx: neither consumed nor revoked; expiry is derived status).
	sqlInvitationsPage = `SELECT ` + invitationColumns + ` FROM app.invitations ` +
		`WHERE organization_id = $1 AND scope_type = $2 AND scope_id = $3 AND id > $4 ORDER BY id LIMIT $5`
	sqlMyInvitationsPage = `SELECT ` + invitationRecipientColumns + ` FROM app.invitations ` +
		`WHERE organization_id = $1 AND intended_user_id = $2 AND consumed_at IS NULL AND revoked_at IS NULL AND id > $3 ` +
		`ORDER BY id LIMIT $4`
	// The pending unexpired count of a scope (invitations_pending_idx); expired history never counts against the bound.
	sqlInvitationsPendingCount = `SELECT count(*) FROM app.invitations WHERE organization_id = $1 AND scope_type = $2 AND scope_id = $3 ` +
		`AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at > now()`
	// The two consumption transitions (claim-writer bodies). invitation_claim_organization is EXECUTE-granted to the
	// owner role only in G019 (D-15 pending, G058): a runtime call is refused with 42501.
	sqlClaimGroupInvitation = `SELECT outcome, organization_id, scope_type, scope_id, group_id, role, binding_id ` +
		`FROM app.invitation_claim_group($1, $2)`
	sqlClaimOrganizationInvitation = `SELECT outcome, organization_id, scope_type, scope_id, group_id, role, binding_id ` +
		`FROM app.invitation_claim_organization($1, $2, $3, $4, $5)`
	// G058 enrollment admission (enrollment.go AdmitEnrollmentTx): the admission body (00016; cw-owned SECURITY DEFINER,
	// EXECUTE runtime and owner) with the credential digest, the Organization, the kind and the request reference (the
	// claimed invitation for lift, NULL for open).
	sqlEnrollmentAdmit = `SELECT outcome, binding_id FROM app.enrollment_admit($1, $2, $3, $4)`
	// G058 slice 1b enrollment policy (enrollment.go): the runtime reads the mode and the two readiness columns under
	// organizations_select (table-level SELECT, 00003) and updates the mode only (column grant
	// organizations_enrollment_update, 00016; the owner guard and the open-requires-ready CHECK apply).
	sqlReadEnrollmentPolicy = `SELECT enrollment_mode, enrollment_history_resolved_at, enrollment_history_basis ` +
		`FROM app.organizations WHERE id = $1`
	sqlSetEnrollmentMode = `UPDATE app.organizations SET enrollment_mode = $2 WHERE id = $1 ` +
		`RETURNING enrollment_mode, enrollment_history_resolved_at, enrollment_history_basis`
	// G058 HTTP slice: the public mode of an Organization the caller resolved from a trusted host (no subject, no
	// context): the policy reader's definer function (an absent row is invite_only; it discloses the mode only).
	sqlPublicEnrollmentMode = `SELECT app.rls_enrollment_mode($1)`

	// G058 request slice (enrollment.go; 00016 enrollment_requests, Plan r3 §1.4/§1.5). The applicant's request is the
	// claim-writer body (IDENTITY shape; the INSERT and its enrollment.requested capture happen inside it). The runtime
	// reads requests under enrollment_requests_select (the applicant's own rows, or the context Organization's for its
	// managers) and decides or revokes under enrollment_requests_decide with the five decision columns only: the decider
	// and revoker are the acting user ($4/$3, re-checked = ctx_user() by enrollment_requests_immutable, which also derives
	// authority_seq on approval); the statement predicates name the only legal source state, so zero rows are
	// disambiguated by a re-read. The applicant's credential state and its abandonment touch the caller's own
	// enrollment_sessions row (enrollment_sessions_select/_update: user_id = ctx_user()).
	enrollmentRequestColumns = `id, organization_id, user_id, status, created_at, expires_at, decided_at, decided_by_user_id, ` +
		`revoked_at, revoked_by_user_id, consumed_at, consumed_binding_id, authority_seq`
	sqlEnrollmentRequest       = `SELECT outcome, request_id FROM app.enrollment_request($1, $2)`
	sqlSelectEnrollmentRequest = `SELECT ` + enrollmentRequestColumns + ` FROM app.enrollment_requests WHERE organization_id = $1 AND id = $2`
	sqlDecideEnrollmentRequest = `UPDATE app.enrollment_requests SET status = $3, decided_at = now(), decided_by_user_id = $4 ` +
		`WHERE organization_id = $1 AND id = $2 AND status = 'pending' RETURNING ` + enrollmentRequestColumns
	sqlRevokeEnrollmentApproval = `UPDATE app.enrollment_requests SET status = 'revoked', revoked_at = now(), revoked_by_user_id = $3 ` +
		`WHERE organization_id = $1 AND id = $2 AND status = 'approved' AND consumed_at IS NULL RETURNING ` + enrollmentRequestColumns
	// Keyset page (cursor = request id, limit+1 rows; enrollment_requests_scope_idx), optionally one status.
	sqlEnrollmentRequestsPage = `SELECT ` + enrollmentRequestColumns + ` FROM app.enrollment_requests ` +
		`WHERE organization_id = $1 AND ($2::text IS NULL OR status = $2) AND id > $3 ORDER BY id LIMIT $4`
	// The applicant's latest request for the Organization (enrollment_requests_user_idx).
	sqlOwnEnrollmentRequest = `SELECT ` + enrollmentRequestColumns + ` FROM app.enrollment_requests ` +
		`WHERE organization_id = $1 AND user_id = $2 ORDER BY created_at DESC, id DESC LIMIT 1`
	sqlEnrollmentCredentialState = `SELECT admitted_at, admitted_binding_id FROM app.enrollment_sessions ` +
		`WHERE organization_id = $1 AND user_id = $2 AND credential_hash = $3 AND revoked_at IS NULL AND expires_at > now()`
	sqlAbandonEnrollment = `UPDATE app.enrollment_sessions SET revoked_at = now(), revoke_reason = 'abandoned' ` +
		`WHERE organization_id = $1 AND user_id = $2 AND credential_hash = $3 AND revoked_at IS NULL`
)

func uuidArg[T ~[16]byte](id T) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

func exists(ctx context.Context, q pgx.Tx, notFound Code, sql string, args ...any) error {
	var ok bool
	if err := q.QueryRow(ctx, sql, args...).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return newError(notFound)
	}
	return nil
}

// one maps a missing row onto the given not-found code.
func one(err error, notFound Code) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return newError(notFound)
	}
	return err
}

func scanOrganization(row pgx.Row) (Organization, error) {
	var o Organization
	var id pgtype.UUID
	var at time.Time
	if err := row.Scan(&id, &o.Name, &at); err != nil {
		return Organization{}, err
	}
	o.ID, o.CreatedAt = OrganizationID(id.Bytes), at
	return o, nil
}

func scanGroup(row pgx.Row) (Group, error) {
	var g Group
	var id, org pgtype.UUID
	if err := row.Scan(&id, &org, &g.Name, &g.Slug, &g.CreatedAt); err != nil {
		return Group{}, err
	}
	g.ID, g.OrganizationID = GroupID(id.Bytes), OrganizationID(org.Bytes)
	return g, nil
}

func scanBinding(row pgx.Row) (Binding, error) {
	var b Binding
	var id, org, subject, scope, group pgtype.UUID
	var scopeType string
	if err := row.Scan(&id, &org, &subject, &scopeType, &scope, &group, &b.Role, &b.Source, &b.CreatedAt); err != nil {
		return Binding{}, err
	}
	b.ID, b.OrganizationID, b.SubjectUserID = BindingID(id.Bytes), OrganizationID(org.Bytes), UserID(subject.Bytes)
	b.ScopeType, b.ScopeID = ScopeType(scopeType), scope.Bytes
	if group.Valid {
		g := GroupID(group.Bytes)
		b.GroupID = &g
	}
	return b, nil
}

func scanOrgMembership(row pgx.Row) (OrganizationMembership, error) {
	var m OrganizationMembership
	var user, org pgtype.UUID
	var role string
	if err := row.Scan(&user, &org, &role); err != nil {
		return OrganizationMembership{}, err
	}
	m.UserID, m.OrganizationID, m.Role = UserID(user.Bytes), OrganizationID(org.Bytes), OrganizationRole(role)
	return m, nil
}

func scanGroupMembership(row pgx.Row) (GroupMembership, error) {
	var m GroupMembership
	var user, org, group pgtype.UUID
	var role string
	if err := row.Scan(&user, &org, &group, &role); err != nil {
		return GroupMembership{}, err
	}
	m.UserID, m.OrganizationID, m.GroupID, m.Role = UserID(user.Bytes), OrganizationID(org.Bytes), GroupID(group.Bytes), GroupRole(role)
	return m, nil
}

// collect scans every row with scan and closes the rows.
func collect[T any](rows pgx.Rows, err error, scan func(pgx.Row) (T, error)) ([]T, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func scanBusinessProfile(row pgx.Row) (BusinessProfile, error) {
	var p BusinessProfile
	var id, org pgtype.UUID
	c := &p.Context
	if err := row.Scan(&id, &org, &c.DisplayName, &c.LegalName, &c.Description, &c.Industry, &c.Website, &p.CreatedAt); err != nil {
		return BusinessProfile{}, err
	}
	p.ID, p.OrganizationID = BusinessProfileID(id.Bytes), OrganizationID(org.Bytes)
	return p, nil
}

func scanAssociation(row pgx.Row) (GroupBusinessProfile, error) {
	var a GroupBusinessProfile
	var id, org, group, profile pgtype.UUID
	if err := row.Scan(&id, &org, &group, &profile, &a.CreatedAt); err != nil {
		return GroupBusinessProfile{}, err
	}
	a.ID, a.OrganizationID, a.GroupID, a.BusinessProfileID = AssociationID(id.Bytes), OrganizationID(org.Bytes), GroupID(group.Bytes),
		BusinessProfileID(profile.Bytes)
	return a, nil
}

func scanAssociationProfile(row pgx.Row) (GroupBusinessProfile, BusinessProfile, error) {
	var a GroupBusinessProfile
	var p BusinessProfile
	var aid, aorg, group, profile, pid, porg pgtype.UUID
	c := &p.Context
	if err := row.Scan(&aid, &aorg, &group, &profile, &a.CreatedAt, &pid, &porg, &c.DisplayName, &c.LegalName, &c.Description,
		&c.Industry, &c.Website, &p.CreatedAt); err != nil {
		return GroupBusinessProfile{}, BusinessProfile{}, err
	}
	a.ID, a.OrganizationID, a.GroupID, a.BusinessProfileID = AssociationID(aid.Bytes), OrganizationID(aorg.Bytes), GroupID(group.Bytes),
		BusinessProfileID(profile.Bytes)
	p.ID, p.OrganizationID = BusinessProfileID(pid.Bytes), OrganizationID(porg.Bytes)
	return a, p, nil
}

func scanAccessPath(row pgx.Row) (BusinessProfileAccessPath, error) {
	var ap BusinessProfileAccessPath
	var user, org, profile, group pgtype.UUID
	var role string
	if err := row.Scan(&user, &org, &profile, &group, &role); err != nil {
		return BusinessProfileAccessPath{}, err
	}
	ap.UserID, ap.OrganizationID, ap.BusinessProfileID, ap.GroupID, ap.GroupRole = UserID(user.Bytes), OrganizationID(org.Bytes),
		BusinessProfileID(profile.Bytes), GroupID(group.Bytes), GroupRole(role)
	return ap, nil
}
