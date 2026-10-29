package access

import (
	"context"

	"core/internal/audit"
	"core/internal/tenancy"
)

// G060 explicit Group authority (ADR-008; Plan r0 §1.3-§1.4, r1 C5): the owner-only Group-management override of the
// active Organization and the atomic audit of administration allowed through it. The override is one persisted boolean
// (00017 app.organizations.group_management_override, default false). Only a current Organization owner changes it
// (organization.manage_policy, through its own operation organization.set_group_override; the generic
// SetOrganizationPolicy never sets it), and the database guard app.organizations_group_override_guard refuses any other
// session role. While it is enabled, current Organization owners and admins administer ordinary Group metadata,
// onboarding, members and invitations without a Group binding; it never confers Group content, Group ownership, a
// synthetic membership or a self-granted role.

// GroupAuthorityView is the override's state for the active Organization: whether it is enabled, and whether the caller
// may change it (a hint decided on the same facts; it authorizes nothing).
type GroupAuthorityView struct {
	OverrideEnabled bool
	CanChange       bool
}

// GetGroupAuthority reads the active Organization's override (organization.read: every current Organization member; the
// operation and record of organization.get_policy). The value is read with the Organization row; a read failure is
// unavailable, never an allow.
func (s *Service) GetGroupAuthority(ctx context.Context, tc *TenantContext) (GroupAuthorityView, error) {
	var out GroupAuthorityView
	err := s.guard(ctx, tc, opGetOrganizationPolicy, capOrganizationRead, false, Target{Kind: kindOrganization}, organizationQuery, nil,
		func(_ *trackedTx, f Facts) *failure {
			can, _ := decide(capOrganizationManagePolicy, tc.selection(), f, Target{Kind: kindOrganization, Organization: tc.organization})
			out = GroupAuthorityView{OverrideEnabled: f.Policy.GroupManagementOverride, CanChange: can}
			return nil
		})
	if err != nil {
		return GroupAuthorityView{}, err
	}
	return out, nil
}

// SetGroupManagementOverride assigns the active Organization's override (organization.manage_policy: the current
// org_owner only, whether or not it holds any Group role; an Organization admin is forbidden even when it owns a Group).
// It is an absolute set: setting the current value is a successful no-op. The Organization row is locked FOR NO KEY
// UPDATE up front (policyWriteQuery), so every guarded Group write, which holds that row FOR SHARE while it relies on the
// override, and every Group invitation consumption, which locks it FOR SHARE before its issuer check, either commits
// before this change or observes it. The capture trigger records organization.policy_changed (reason
// group_management_override_enabled or _disabled) in the same transaction.
func (s *Service) SetGroupManagementOverride(ctx context.Context, tc *TenantContext, enabled bool) (GroupAuthorityView, error) {
	var out GroupAuthorityView
	err := s.guard(ctx, tc, opSetGroupOverride, capOrganizationManagePolicy, true, Target{Kind: kindOrganization}, policyWriteQuery, nil,
		func(tx *trackedTx, f Facts) *failure {
			v, err := tenancy.SetGroupManagementOverrideTx(ctx, tx, f.Organization.ID, enabled)
			if err != nil {
				if tenancy.CodeOf(err) == tenancy.CodeOwnerAuthority {
					// The database guard: only reachable if the facts and the statement disagree on the owner.
					return &failure{code: CodeForbidden, field: string(tenancy.CodeOwnerAuthority), reason: reasonRoleNotPermitted}
				}
				return writeFailure(ctx, tx, err)
			}
			out = GroupAuthorityView{OverrideEnabled: v, CanChange: true}
			return nil
		})
	if err != nil {
		return GroupAuthorityView{}, err
	}
	return out, nil
}

// overrideUsed appends access.override_used on the guarded write's own transaction when decision c on t was allowed only
// through the Organization override path (Plan r1 C5): the explicit Group path is decided first, so a caller who also
// qualifies through its own Group role records nothing. The event names the actual mutated resource: the Group for a
// Group target; the Business Profile for a profile target, with group set only when the write ran against that Group
// (an existing-profile onboarding completion). An append refusal or failure refuses the write (unavailable): the
// mutation and its record commit together or not at all.
func (s *Service) overrideUsed(ctx context.Context, tx *trackedTx, tc *TenantContext, f Facts, c Capability, t Target, op string,
	group tenancy.GroupID) *failure {
	t.Organization = tc.organization
	if ok, r := decide(c, tc.selection(), f, t); !ok || r != reasonAllowedOverride {
		return nil
	}
	e := audit.Event{Action: audit.AccessOverrideUsed, Outcome: audit.Success, Operation: audit.Operation(op),
		Organization: audit.UUID(tc.organization)}
	switch t.Kind {
	case kindGroup:
		e.ResourceType, e.ResourceID, e.Group = audit.ResourceGroup, audit.UUID(t.Group), audit.UUID(t.Group)
	case kindProfile:
		e.ResourceType, e.ResourceID, e.Group = audit.ResourceBusinessProfile, audit.UUID(t.Profile), audit.UUID(group)
	default:
		return fail(CodeInternal, reasonFactInconsistent)
	}
	if err := audit.Append(ctx, tx, e); err != nil {
		return fail(CodeUnavailable, failureReason(ctx, tx.cause))
	}
	return nil
}
