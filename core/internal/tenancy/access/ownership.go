package access

import (
	"context"

	"core/internal/tenancy"
)

// G019 ownership administration (Plan r1 §1.4 "Promotion / demotion / transfer"). The capability is
// organization.manage_owners (org_owner) or group.manage_owners (the exact Group's group_owner; no Organization-owner
// override, Spec R1). The checks below run on the locked Subject facts for truthful feedback; the authority guard
// re-enforces every eligibility condition at the statement (same-scope enabled admin, Organization membership for a
// Group recipient, the last-owner count), so a raced change is refused there.

// PromoteToOwner makes an enabled same-scope admin an owner. A subject who is not an admin of the scope is
// recipient_ineligible.
func (s *Service) PromoteToOwner(ctx context.Context, tc *TenantContext, sc Scope, subject tenancy.UserID) (Member, error) {
	var out Member
	pre := scopePre(sc, func() *failure { return targetID(subject)() })
	err := s.guard(ctx, tc, opOwnershipPromote, sc.manageOwners(), true, sc.target(), sc.query(true, &subject, false, false), pre,
		func(tx *trackedTx, f Facts) *failure {
			b, ok := subjectBinding(f, sc, subject)
			if !ok {
				return notFound()
			}
			if b.Role != sc.adminRole() {
				return conflict(string(tenancy.CodeRecipientIneligible))
			}
			nb, err := tenancy.ChangeRoleTx(ctx, tx, b, sc.ownerRole())
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			out = Member{UserID: subject, Role: nb.Role, Self: subject == tc.subject}
			return nil
		})
	if err != nil {
		return Member{}, err
	}
	return out, nil
}

// DemoteOwner sets an owner's role to an admin or member role of the scope. A subject who is not an owner is not_owner;
// the last owner is refused (last_owner). The actor may demote itself only while another owner remains.
func (s *Service) DemoteOwner(ctx context.Context, tc *TenantContext, sc Scope, subject tenancy.UserID, role string) (Member, error) {
	var out Member
	pre := scopePre(sc, func() *failure {
		if f := targetID(subject)(); f != nil {
			return f
		}
		if !ordinaryRole(sc, role) {
			return invalid(string(tenancy.CodeInvalidRole))
		}
		return nil
	})
	err := s.guard(ctx, tc, opOwnershipDemote, sc.manageOwners(), true, sc.target(), sc.query(true, &subject, false, true), pre,
		func(tx *trackedTx, f Facts) *failure {
			b, ok := subjectBinding(f, sc, subject)
			if !ok {
				return notFound()
			}
			if b.Role != sc.ownerRole() {
				return conflict(fieldNotOwner)
			}
			if f.OwnerCount <= 1 {
				return conflict(string(tenancy.CodeLastOwner))
			}
			nb, err := tenancy.ChangeRoleTx(ctx, tx, b, role)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			out = Member{UserID: subject, Role: nb.Role, Self: subject == tc.subject}
			return nil
		})
	if err != nil {
		return Member{}, err
	}
	return out, nil
}

// TransferOwnership makes an enabled same-scope admin an owner, then sets the actor's own role to formerRole (an admin
// or member role of the scope) or, with "none", removes the actor from the scope. The recipient is updated first, so
// the guard's last-owner count on the actor's statement already sees the new owner. An Organization transfer with
// "none" is an Organization departure: it needs an Organization-only context, and the authority guard removes the
// actor's Group bindings inside the Organization DELETE (00017 amendment01; it refuses a Group's last owner and the
// whole transfer rolls back).
func (s *Service) TransferOwnership(ctx context.Context, tc *TenantContext, sc Scope, recipient tenancy.UserID, formerRole string) error {
	pre := scopePre(sc, func() *failure {
		if f := targetID(recipient)(); f != nil {
			return f
		}
		if tc != nil && recipient == tc.subject {
			return invalid(fieldSelfTarget)
		}
		if formerRole == roleNone {
			return organizationOnly(tc, sc)
		}
		if !ordinaryRole(sc, formerRole) {
			return invalid(string(tenancy.CodeInvalidRole))
		}
		return nil
	})
	return s.guard(ctx, tc, opOwnershipTransfer, sc.manageOwners(), true, sc.target(), sc.query(true, &recipient, true, false), pre,
		func(tx *trackedTx, f Facts) *failure {
			ab, ok := subjectBinding(f, sc, tc.subject)
			if !ok || ab.Role != sc.ownerRole() {
				return &failure{code: CodeForbidden, field: string(tenancy.CodeOwnerAuthority), reason: reasonRoleNotPermitted}
			}
			rb, ok := subjectBinding(f, sc, recipient)
			if !ok {
				return notFound()
			}
			if rb.Role != sc.adminRole() {
				return conflict(string(tenancy.CodeRecipientIneligible))
			}
			if _, err := tenancy.ChangeRoleTx(ctx, tx, rb, sc.ownerRole()); err != nil {
				return membershipFailure(ctx, tx, err)
			}
			if formerRole != roleNone {
				if _, err := tenancy.ChangeRoleTx(ctx, tx, ab, formerRole); err != nil {
					return membershipFailure(ctx, tx, err)
				}
				return nil
			}
			// An Organization departure's Group bindings are removed by the guard inside this DELETE (00017 amendment01).
			if err := tenancy.RemoveBindingTx(ctx, tx, ab); err != nil {
				return membershipFailure(ctx, tx, err)
			}
			return nil
		})
}
