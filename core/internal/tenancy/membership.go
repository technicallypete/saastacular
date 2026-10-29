package tenancy

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// G019 membership and ownership primitives (Plan r1 §1.4, §5.1; 00015_membership_invitations.sql). Like facts.go they
// run in a transaction the caller owns, carry no actor and enforce no capability: the guarded facade
// (core/internal/tenancy/access) decides, and 00015's row-level security and app.role_bindings_authority_guard trigger
// re-enforce owner authority, promotion eligibility, the dependent-grant rule and last-owner cardinality on every
// statement. CreateOrganizationWithOwner is the one trusted Service method (a privileged pool; no route).

// MaxMembersPage bounds one member, candidate or invitation page (Plan r1 §1.7: limit 1..100).
const MaxMembersPage = 100

// Member is one row of a member page: the subject, its display name (nil when unset), its role in the scope and
// whether its global user is disabled.
type Member struct {
	UserID      UserID
	DisplayName *string
	Role        string
	Disabled    bool
}

// Candidate is one Group invitation candidate: an enabled Organization member without a binding in the Group.
type Candidate struct {
	UserID      UserID
	DisplayName *string
}

// CreateOrganizationWithOwner creates an Organization and the owner's org_owner membership binding in one transaction
// (Plan r1 §1.4 C2, MEM-1): the only supported creation shape. Either both rows commit or neither does; an unknown owner
// is user_not_found. It runs on the Service's pool, which for this method is a privileged (non-runtime) pool: the
// runtime role has no INSERT on organizations and the authority guard does not run for other roles. No route or
// production caller exists in G019.
//
// Before either write it locks the owner's users row FOR SHARE and requires status active (G019 F-01 refinement a):
// the lock conflicts with identity's FOR UPDATE (emergency disable, retirement finalization), so a creation that waited
// for a disable reads the committed status and refuses as recipient_ineligible, writing nothing; an owner row the pool
// cannot see is user_not_found (fail closed). Pool assumption: the supported privileged pool (the trusted setup path)
// is not bound by app.users row-level security for this read; a FORCE-bound owner pool without the owner's context sees
// no row and refuses. Arbitrary superuser SQL stays outside this guarantee.
func (s *Service) CreateOrganizationWithOwner(ctx context.Context, name string, owner UserID) (Organization, Binding, error) {
	var org Organization
	var b Binding
	err := s.mutate(ctx, opOrganizationCreate, "", func() error {
		if zero(owner) {
			return newError(CodeInvalidID)
		}
		n, err := ValidateName(name)
		if err != nil {
			return err
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			var status string
			if err := tx.QueryRow(ctx, sqlLockOwnerStatus, uuidArg(owner)).Scan(&status); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return newError(CodeUserNotFound)
				}
				return err
			}
			if status != "active" {
				return newError(CodeRecipientIneligible)
			}
			var err error
			if org, err = scanOrganization(tx.QueryRow(ctx, sqlInsertOrganization, n)); err != nil {
				return err
			}
			b, err = scanBinding(tx.QueryRow(ctx, sqlInsertBinding, uuidArg(org.ID), uuidArg(owner), string(ScopeOrganization),
				uuidArg(org.ID), pgtype.UUID{}, string(OrgOwner)))
			return err
		})
	})
	if err != nil {
		return Organization{}, Binding{}, err
	}
	return org, b, nil
}

// roleValid reports whether role is one of the three roles of the scope.
func roleValid(scope ScopeType, role string) bool {
	switch scope {
	case ScopeOrganization:
		return OrganizationRole(role).Valid()
	case ScopeGroup:
		return GroupRole(role).Valid()
	}
	return false
}

// ChangeRoleTx sets the role of the locked membership binding b (keyed by b.OrganizationID and b.ID) inside the
// caller's transaction and returns the stored binding (CreatedAt is zero: the runtime role does not read it). A role
// outside b's scope is invalid_role; a binding that is gone or invisible is membership_not_found; the guard's refusals
// arrive as last_owner, owner_authority or recipient_ineligible.
func ChangeRoleTx(ctx context.Context, tx pgx.Tx, b Binding, role string) (Binding, error) {
	if zero(b.OrganizationID) || zero(b.ID) {
		return Binding{}, newError(CodeInvalidID)
	}
	if !roleValid(b.ScopeType, role) {
		return Binding{}, newError(CodeInvalidRole)
	}
	out, err := scanFactBinding(tx.QueryRow(ctx, sqlChangeRoleByID, uuidArg(b.OrganizationID), uuidArg(b.ID), role))
	if e := classify(one(err, CodeMembershipNotFound)); e != nil {
		return Binding{}, e
	}
	return out, nil
}

// RemoveBindingTx deletes the locked membership binding b inside the caller's transaction. A binding that is gone or
// invisible is membership_not_found; the guard's refusals arrive as last_owner, owner_authority, context_required or
// dependent_bindings.
func RemoveBindingTx(ctx context.Context, tx pgx.Tx, b Binding) error {
	if zero(b.OrganizationID) || zero(b.ID) {
		return newError(CodeInvalidID)
	}
	tag, err := tx.Exec(ctx, sqlRemoveBindingByID, uuidArg(b.OrganizationID), uuidArg(b.ID))
	if e := classify(err); e != nil {
		return e
	}
	if tag.RowsAffected() != 1 {
		return newError(CodeMembershipNotFound)
	}
	return nil
}

// OwnerCount returns the number of owner bindings of the scope (an Organization, or a Group of the active
// Organization) through the definer reader app.rls_owner_count: 0 when the context cannot see the scope. It is truthful
// feedback and the facade's precheck; the guard's recount at the statement is the guarantee.
func OwnerCount(ctx context.Context, tx pgx.Tx, scope ScopeType, scopeID [16]byte) (int, error) {
	n, err := ownerCount(ctx, tx, scope, scopeID)
	if e := classify(err); e != nil {
		return 0, e
	}
	return n, nil
}

func ownerCount(ctx context.Context, tx pgx.Tx, scope ScopeType, scopeID [16]byte) (int, error) {
	if scope != ScopeOrganization && scope != ScopeGroup {
		return 0, newError(CodeInvalidRole)
	}
	if scopeID == [16]byte{} {
		return 0, newError(CodeInvalidID)
	}
	var n int32
	if err := tx.QueryRow(ctx, sqlOwnerCount, string(scope), pgtype.UUID{Bytes: scopeID, Valid: true}).Scan(&n); err != nil {
		return 0, err
	}
	return int(n), nil
}

// MembersPage reads at most limit+1 members of a scope ordered by user id after after (zero: from the start), in one
// statement, so the caller can tell whether a next page exists. group zero selects the Organization scope. With
// ownerCandidates only the enabled admins of the scope are listed (the promotion/transfer picker, C5). Row-level
// security limits the rows to those the context may read.
func MembersPage(ctx context.Context, tx pgx.Tx, org OrganizationID, group GroupID, after UserID, limit int, ownerCandidates bool) ([]Member, error) {
	if zero(org) || limit < 1 || limit > MaxMembersPage {
		return nil, newError(CodeInvalidID)
	}
	var rows pgx.Rows
	var err error
	if zero(group) {
		filter := ""
		if ownerCandidates {
			filter = string(OrgAdmin)
		}
		rows, err = tx.Query(ctx, sqlOrganizationMembersPage, uuidArg(org), uuidArg(after), filter, limit+1)
	} else {
		filter := ""
		if ownerCandidates {
			filter = string(GroupAdmin)
		}
		rows, err = tx.Query(ctx, sqlGroupMembersPage, uuidArg(org), uuidArg(group), uuidArg(after), filter, limit+1)
	}
	out, err := collect(rows, err, scanMember)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// InvitationCandidatesPage reads at most limit+1 invitation candidates of a Group of the active Organization ordered
// by user id after after (zero: from the start) through the definer reader app.rls_group_invitation_candidates, which
// answers only Organization owners/admins and the Group's owners/admins of a matching context, and nothing otherwise.
func InvitationCandidatesPage(ctx context.Context, tx pgx.Tx, group GroupID, after UserID, limit int) ([]Candidate, error) {
	if zero(group) || limit < 1 || limit > MaxMembersPage {
		return nil, newError(CodeInvalidID)
	}
	cursor := pgtype.UUID{}
	if !zero(after) {
		cursor = uuidArg(after)
	}
	rows, err := tx.Query(ctx, sqlInvitationCandidatesPage, uuidArg(group), cursor, int32(limit+1))
	out, err := collect(rows, err, scanCandidate)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// ScopeRef identifies one scope: an Organization (GroupID nil, ScopeID the Organization) or a Group of it.
type ScopeRef struct {
	OrganizationID OrganizationID
	ScopeType      ScopeType
	ScopeID        [16]byte
	GroupID        *GroupID
}

// SoleOwnerScopes returns the context user's sole-owned scopes, identifiers only, through the definer reader
// app.rls_sole_owner_scopes (Plan r1 §1.4 "Retirement"; IDENTITY-shape context): the subject sees only its own
// blocking scopes, never another member. Nothing without a context user.
func SoleOwnerScopes(ctx context.Context, tx pgx.Tx) ([]ScopeRef, error) {
	rows, err := tx.Query(ctx, sqlSoleOwnerScopes)
	out, err := collect(rows, err, scanScopeRef)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// OwnOrganizationsPage reads at most limit (1..MaxMembersPage) of the context user's own current Organization
// memberships, identifiers only, in ascending id after after (zero: from the start), through the definer reader
// app.rls_own_organizations (Result addendum A1-01; IDENTITY-shape context only): nothing about any other subject,
// and nothing outside that shape. A page shorter than limit is the last one.
func OwnOrganizationsPage(ctx context.Context, tx pgx.Tx, after OrganizationID, limit int) ([]OrganizationID, error) {
	if limit < 1 || limit > MaxMembersPage {
		return nil, newError(CodeInvalidID)
	}
	cursor := pgtype.UUID{}
	if !zero(after) {
		cursor = uuidArg(after)
	}
	rows, err := tx.Query(ctx, sqlOwnOrganizationsPage, cursor, int32(limit))
	out, err := collect(rows, err, func(row pgx.Row) (OrganizationID, error) {
		var id pgtype.UUID
		if err := row.Scan(&id); err != nil {
			return OrganizationID{}, err
		}
		return OrganizationID(id.Bytes), nil
	})
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

func scanScopeRef(row pgx.Row) (ScopeRef, error) {
	var org, scope pgtype.UUID
	var scopeType string
	if err := row.Scan(&org, &scopeType, &scope); err != nil {
		return ScopeRef{}, err
	}
	r := ScopeRef{OrganizationID: OrganizationID(org.Bytes), ScopeType: ScopeType(scopeType), ScopeID: scope.Bytes}
	if r.ScopeType == ScopeGroup {
		g := GroupID(scope.Bytes)
		r.GroupID = &g
	}
	return r, nil
}

// scanFactBinding scans the eight factBindingColumns (no created_at).
func scanFactBinding(row pgx.Row) (Binding, error) {
	var b Binding
	var id, org, subject, scope, group pgtype.UUID
	var scopeType string
	if err := row.Scan(&id, &org, &subject, &scopeType, &scope, &group, &b.Role, &b.Source); err != nil {
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

func scanMember(row pgx.Row) (Member, error) {
	var m Member
	var user pgtype.UUID
	var status string
	if err := row.Scan(&user, &m.DisplayName, &status, &m.Role); err != nil {
		return Member{}, err
	}
	m.UserID, m.Disabled = UserID(user.Bytes), status != userActive
	return m, nil
}

func scanCandidate(row pgx.Row) (Candidate, error) {
	var c Candidate
	var user pgtype.UUID
	if err := row.Scan(&user, &c.DisplayName); err != nil {
		return Candidate{}, err
	}
	c.UserID = UserID(user.Bytes)
	return c, nil
}

// userActive is the enabled value of app.users.status (00002_identity.sql).
const userActive = "active"
