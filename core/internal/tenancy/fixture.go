package tenancy

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Fixture-only primitives (G019 Plan r1 §1.4 "Trusted bootstrap and the legacy entry points", C2/Q4). These Service
// methods are NOT application operations: CreateOrganization creates an ownerless tenant, and the Grant*/Change*/
// Remove* primitives write role_bindings without the owner-cardinality, owner-authority or dependent-grant rules of the
// guarded facade (core/internal/tenancy/access). They keep their names and signatures (moved from service.go, no
// test-rename campaign) and exist for test fixtures on a privileged pool only. The supported creation shape is
// CreateOrganizationWithOwner (membership.go). TestTenancyFixturePrimitives pins that no production file outside this
// one calls them; adding a production caller fails the package's tests.

// CreateOrganization creates an Organization with a database-generated UUID. It grants no membership: the result is
// ownerless, which is why it is fixture-only (use CreateOrganizationWithOwner).
func (s *Service) CreateOrganization(ctx context.Context, name string) (Organization, error) {
	var out Organization
	err := s.mutate(ctx, opOrganizationCreate, "", func() error {
		n, err := ValidateName(name)
		if err != nil {
			return err
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			out, err = scanOrganization(tx.QueryRow(ctx, sqlInsertOrganization, n))
			return err
		})
	})
	if err != nil {
		return Organization{}, err
	}
	return out, nil
}

// --- Memberships: writes go to role_bindings only; reads go through the projections only --------------------------

// grant inserts one membership binding after the scope and subject prechecks.
func (s *Service) grant(ctx context.Context, org OrganizationID, scope ScopeType, scopeID [16]byte, group pgtype.UUID,
	u UserID, role string, precheck func(pgx.Tx) error) (Binding, error) {
	var out Binding
	err := s.mutate(ctx, opMembershipGrant, scope, func() error {
		if zero(org) || zero(u) || scopeID == [16]byte{} {
			return newError(CodeInvalidID)
		}
		if role == "" {
			return newError(CodeInvalidRole)
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			if err := precheck(tx); err != nil {
				return err
			}
			if err := exists(ctx, tx, CodeUserNotFound, sqlUserExists, uuidArg(u)); err != nil {
				return err
			}
			var err error
			out, err = scanBinding(tx.QueryRow(ctx, sqlInsertBinding, uuidArg(org), uuidArg(u), string(scope),
				pgtype.UUID{Bytes: scopeID, Valid: true}, group, role))
			return err
		})
	})
	if err != nil {
		return Binding{}, err
	}
	return out, nil
}

// change updates the role of the one membership binding in place; its UUID is unchanged.
func (s *Service) change(ctx context.Context, org OrganizationID, scope ScopeType, scopeID [16]byte, u UserID, role string,
	precheck func(pgx.Tx) error) (Binding, error) {
	var out Binding
	err := s.mutate(ctx, opMembershipChange, scope, func() error {
		if zero(org) || zero(u) || scopeID == [16]byte{} {
			return newError(CodeInvalidID)
		}
		if role == "" {
			return newError(CodeInvalidRole)
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			if err := precheck(tx); err != nil {
				return err
			}
			var err error
			out, err = scanBinding(tx.QueryRow(ctx, sqlChangeRole, uuidArg(org), uuidArg(u), string(scope),
				pgtype.UUID{Bytes: scopeID, Valid: true}, role))
			return one(err, CodeMembershipNotFound)
		})
	})
	if err != nil {
		return Binding{}, err
	}
	return out, nil
}

// remove deletes the one membership binding; a later grant creates a new binding with a new UUID.
func (s *Service) remove(ctx context.Context, org OrganizationID, scope ScopeType, scopeID [16]byte, u UserID,
	precheck func(pgx.Tx) error) error {
	return s.mutate(ctx, opMembershipRemove, scope, func() error {
		if zero(org) || zero(u) || scopeID == [16]byte{} {
			return newError(CodeInvalidID)
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			if err := precheck(tx); err != nil {
				return err
			}
			tag, err := tx.Exec(ctx, sqlRemoveBinding, uuidArg(org), uuidArg(u), string(scope), pgtype.UUID{Bytes: scopeID, Valid: true})
			if err != nil {
				return err
			}
			if tag.RowsAffected() != 1 {
				return newError(CodeMembershipNotFound)
			}
			return nil
		})
	})
}

func (s *Service) orgPrecheck(ctx context.Context, org OrganizationID) func(pgx.Tx) error {
	return func(tx pgx.Tx) error {
		return exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org))
	}
}

func (s *Service) groupPrecheck(ctx context.Context, org OrganizationID, g GroupID) func(pgx.Tx) error {
	return func(tx pgx.Tx) error {
		return exists(ctx, tx, CodeGroupNotFound, sqlGroupExists, uuidArg(org), uuidArg(g))
	}
}

func orgRole(r OrganizationRole) string {
	if !r.Valid() {
		return ""
	}
	return string(r)
}

func groupRole(r GroupRole) string {
	if !r.Valid() {
		return ""
	}
	return string(r)
}

// GrantOrganizationRole creates the subject's one Organization membership binding; a second grant is
// membership_exists. It synthesizes no Group membership. Fixture-only: no owner-authority or cardinality rule.
func (s *Service) GrantOrganizationRole(ctx context.Context, org OrganizationID, u UserID, r OrganizationRole) (Binding, error) {
	return s.grant(ctx, org, ScopeOrganization, org, pgtype.UUID{}, u, orgRole(r), s.orgPrecheck(ctx, org))
}

// ChangeOrganizationRole changes the role of the subject's Organization membership in place. Fixture-only.
func (s *Service) ChangeOrganizationRole(ctx context.Context, org OrganizationID, u UserID, r OrganizationRole) (Binding, error) {
	return s.change(ctx, org, ScopeOrganization, org, u, orgRole(r), s.orgPrecheck(ctx, org))
}

// RemoveOrganizationMembership deletes the subject's Organization membership binding. Group bindings, the global
// user and its sessions are unchanged. Fixture-only: the guarded departure removes dependent Group bindings first.
func (s *Service) RemoveOrganizationMembership(ctx context.Context, org OrganizationID, u UserID) error {
	return s.remove(ctx, org, ScopeOrganization, org, u, s.orgPrecheck(ctx, org))
}

// GrantGroupRole creates the subject's one membership binding in a Group of the Organization. It synthesizes no
// Organization membership. group_owner is an ordinary role here: ownership transfer and last-owner policy are not
// part of these fixture-only primitives.
func (s *Service) GrantGroupRole(ctx context.Context, org OrganizationID, g GroupID, u UserID, r GroupRole) (Binding, error) {
	return s.grant(ctx, org, ScopeGroup, g, uuidArg(g), u, groupRole(r), s.groupPrecheck(ctx, org, g))
}

// ChangeGroupRole changes the role of the subject's Group membership in place. Fixture-only.
func (s *Service) ChangeGroupRole(ctx context.Context, org OrganizationID, g GroupID, u UserID, r GroupRole) (Binding, error) {
	return s.change(ctx, org, ScopeGroup, g, u, groupRole(r), s.groupPrecheck(ctx, org, g))
}

// RemoveGroupMembership deletes the subject's Group membership binding. Fixture-only.
func (s *Service) RemoveGroupMembership(ctx context.Context, org OrganizationID, g GroupID, u UserID) error {
	return s.remove(ctx, org, ScopeGroup, g, u, s.groupPrecheck(ctx, org, g))
}
