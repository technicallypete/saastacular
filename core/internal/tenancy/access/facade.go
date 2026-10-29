package access

import (
	"context"
	"time"

	"core/internal/audit"
	"core/internal/tenancy"
)

// The guarded facade (G007 Plan r1 §3.5): eight operations, each mapped to exactly one capability, each authorizing and
// acting in one transaction. The operation -> capability mapping is fixed here and is the only one.

// guard runs one guarded operation: the binding check before any transaction, then, in one transaction, the session
// liveness (locked for writes), the ordered facts, the context revalidation and the decision. act runs only on an
// allow, in the same transaction. A target the subject cannot read is not_found; a readable one whose capability is
// denied is forbidden. Exactly one access.decision record is written. G011: a recordable denial is recorded as one
// durable access.denied event after the denied transaction ended and before the operational record.
func (s *Service) guard(ctx context.Context, tc *TenantContext, op string, c Capability, write bool, t Target,
	query func(*TenantContext) tenancy.FactQuery, pre func() *failure, act func(tx *trackedTx, f Facts) *failure) error {
	var allow Reason
	f := s.guardedTx(ctx, tc, c, write, t, query, pre, cleanupUnbounded, &allow, act)
	s.recordGuarded(ctx, tc, op, f)
	s.logGuardedAllow(ctx, op, c, allow, f)
	if f != nil {
		return f.err()
	}
	return nil
}

func (s *Service) guarded(ctx context.Context, tc *TenantContext, c Capability, write bool, t Target,
	query func(*TenantContext) tenancy.FactQuery, pre func() *failure, act func(tx *trackedTx, f Facts) *failure) *failure {
	return s.guardedTx(ctx, tc, c, write, t, query, pre, cleanupUnbounded, nil, act)
}

// guardedTx is guarded with the cleanup bound of its deferred rollback: cleanupUnbounded for every operation through
// guarded (unchanged), completionCleanupBudget for the two onboarding completions through guardBounded. allow, when not
// nil, receives the allow reason of the decision (G060: reasonAllowedOverride for the Organization override path) for the
// access.decision record.
func (s *Service) guardedTx(ctx context.Context, tc *TenantContext, c Capability, write bool, t Target,
	query func(*TenantContext) tenancy.FactQuery, pre func() *failure, cleanup time.Duration, allow *Reason,
	act func(tx *trackedTx, f Facts) *failure) *failure {
	sc, f := s.bind(ctx, tc)
	if f != nil {
		return f
	}
	q := query(tc)
	if pre != nil {
		if f := pre(); f != nil {
			return f
		}
	}
	sel := tc.selection()
	t.Organization = tc.organization
	q.Organization, q.User, q.Lock = tc.organization, tc.subject, write
	return s.inContextTx(ctx, write, dbContext{user: tc.subject, organization: tc.organization, group: tc.group}, cleanup, func(tx *trackedTx) *failure {
		live := s.id.CheckLiveSession
		if write {
			live = s.id.LockLiveSession
		}
		if err := live(ctx, tx, sc.session, sc.subjectText); err != nil {
			return sessionFailure(ctx, tx, err)
		}
		facts, err := tenancy.ReadFacts(ctx, tx, q)
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if r := contextCheck(sel, facts); r != "" {
			return markDenied(fail(CodeContextDenied, r), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		// G019: an operation without a capability (noCapability: the subject's own membership, own invitations and
		// acceptance) runs on the revalidated context alone; act decides any per-row rule itself.
		if c != noCapability {
			ok, r := decide(c, sel, facts, t)
			if !ok {
				res, id, g := targetScope(t)
				return markDenied(denial(c, sel, facts, t, r), sel, facts, res, id, g)
			}
			if allow != nil {
				*allow = r
			}
		}
		return act(tx, facts)
	})
}

// noCapability marks a guarded operation that holds no capability (G019 Plan r1 §1.3 "none" rows): the binding,
// session and context checks still apply; decide is not called and the access.decision record carries no capability.
const noCapability Capability = ""

// denial maps a denied decision: an unknown, foreign, kind-mismatched or outside-context target, or one the subject
// cannot read with the read capability of the same kind on the same facts, is not_found; otherwise forbidden.
func denial(c Capability, sel contextSelection, f Facts, t Target, r Reason) *failure {
	switch r {
	case reasonTargetNotFound, reasonTargetKindMismatch, reasonTargetOutsideContext:
		return fail(CodeNotFound, r)
	}
	if readable, _ := decide(readCapability(c), sel, f, t); readable {
		return fail(CodeForbidden, r)
	}
	return fail(CodeNotFound, r)
}

// organizationQuery reads only the Organization facts, plus the selected Group's for the context revalidation.
func organizationQuery(tc *TenantContext) tenancy.FactQuery {
	groups := []tenancy.GroupID{}
	if !zero(tc.group) {
		groups = []tenancy.GroupID{tc.group}
	}
	return tenancy.FactQuery{Groups: groups}
}

// groupQuery reads the target Group's facts (and the selected Group's). A write locks the target exclusively first,
// unless the context gate will deny it (a target other than the selected Group), when nothing is locked exclusively.
func groupQuery(g tenancy.GroupID, write bool) func(*TenantContext) tenancy.FactQuery {
	return func(tc *TenantContext) tenancy.FactQuery {
		q := tenancy.FactQuery{Groups: []tenancy.GroupID{g}}
		if !zero(tc.group) && tc.group != g {
			q.Groups = append(q.Groups, tc.group)
		} else if write {
			q.TargetGroup = g
		}
		return q
	}
}

// profileQuery reads the profile and the Group-derived path facts: the selected Group in selected-Group context,
// otherwise every Group of the subject in the Organization.
func profileQuery(p tenancy.BusinessProfileID) func(*TenantContext) tenancy.FactQuery {
	return func(tc *TenantContext) tenancy.FactQuery {
		q := tenancy.FactQuery{Profile: p}
		if !zero(tc.group) {
			q.Groups = []tenancy.GroupID{tc.group}
		}
		return q
	}
}

// policyWriteQuery is organizationQuery with the Organization row locked FOR NO KEY UPDATE.
func policyWriteQuery(tc *TenantContext) tenancy.FactQuery {
	q := organizationQuery(tc)
	q.ExclusiveOrganization = true
	return q
}

func targetID[T ~[16]byte](id T) func() *failure {
	return func() *failure {
		if zero(id) {
			return &failure{code: CodeInvalidInput, field: string(tenancy.CodeInvalidID), reason: reasonInputInvalid}
		}
		return nil
	}
}

func findGroup(f Facts, g tenancy.GroupID) tenancy.Group {
	for _, x := range f.Groups {
		if x.ID == g {
			return x
		}
	}
	return tenancy.Group{}
}

func findProfile(f Facts, p tenancy.BusinessProfileID) tenancy.BusinessProfile {
	for _, x := range f.Profiles {
		if x.ID == p {
			return x
		}
	}
	return tenancy.BusinessProfile{}
}

// GetOrganization reads the active Organization (organization.read).
func (s *Service) GetOrganization(ctx context.Context, tc *TenantContext) (tenancy.Organization, error) {
	var out tenancy.Organization
	err := s.guard(ctx, tc, opGetOrganization, capOrganizationRead, false, Target{Kind: kindOrganization}, organizationQuery, nil,
		func(_ *trackedTx, f Facts) *failure {
			out = f.Organization
			return nil
		})
	if err != nil {
		return tenancy.Organization{}, err
	}
	return out, nil
}

// GetOrganizationPolicy reads the active Organization's policy (organization.read).
func (s *Service) GetOrganizationPolicy(ctx context.Context, tc *TenantContext) (tenancy.OrganizationPolicy, error) {
	var out tenancy.OrganizationPolicy
	err := s.guard(ctx, tc, opGetOrganizationPolicy, capOrganizationRead, false, Target{Kind: kindOrganization}, organizationQuery, nil,
		func(_ *trackedTx, f Facts) *failure {
			out = f.Policy
			return nil
		})
	if err != nil {
		return tenancy.OrganizationPolicy{}, err
	}
	return out, nil
}

// SetOrganizationPolicy assigns the active Organization's policy (organization.manage_policy: a current org_owner
// only). It is an absolute set: setting the current value is a successful no-op. The Organization row is locked
// FOR NO KEY UPDATE up front, so a shared lock is never upgraded.
func (s *Service) SetOrganizationPolicy(ctx context.Context, tc *TenantContext, p tenancy.OrganizationPolicy) (tenancy.OrganizationPolicy, error) {
	var out tenancy.OrganizationPolicy
	err := s.guard(ctx, tc, opSetOrganizationPolicy, capOrganizationManagePolicy, true, Target{Kind: kindOrganization}, policyWriteQuery, nil,
		func(tx *trackedTx, f Facts) *failure {
			var err error
			if out, err = tenancy.SetOrganizationPolicyTx(ctx, tx, f.Organization.ID, p); err != nil {
				return writeFailure(ctx, tx, err)
			}
			// G060: this generic policy write never sets the Group-management override (a distinct owner-only
			// operation, SetGroupManagementOverride); it reports the current value read under the exclusive lock.
			out.GroupManagementOverride = f.Policy.GroupManagementOverride
			return nil
		})
	if err != nil {
		return tenancy.OrganizationPolicy{}, err
	}
	return out, nil
}

// GetGroup reads an exact Group of the active Organization (group.read).
func (s *Service) GetGroup(ctx context.Context, tc *TenantContext, g tenancy.GroupID) (tenancy.Group, error) {
	var out tenancy.Group
	err := s.guard(ctx, tc, opGetGroup, capGroupRead, false, Target{Kind: kindGroup, Group: g}, groupQuery(g, false), targetID(g),
		func(_ *trackedTx, f Facts) *failure {
			out = findGroup(f, g)
			return nil
		})
	if err != nil {
		return tenancy.Group{}, err
	}
	return out, nil
}

// RenameGroup changes an exact Group's display name (group.update), with the accepted name contract.
func (s *Service) RenameGroup(ctx context.Context, tc *TenantContext, g tenancy.GroupID, name string) (tenancy.Group, error) {
	var out tenancy.Group
	pre := func() *failure {
		if f := targetID(g)(); f != nil {
			return f
		}
		if _, err := tenancy.ValidateName(name); err != nil {
			return inputFailure(err, reasonInputInvalid)
		}
		return nil
	}
	err := s.guard(ctx, tc, opRenameGroup, capGroupUpdate, true, Target{Kind: kindGroup, Group: g}, groupQuery(g, true), pre,
		func(tx *trackedTx, f Facts) *failure {
			var err error
			if out, err = tenancy.RenameGroupTx(ctx, tx, f.Organization.ID, g, name); err != nil {
				return writeFailure(ctx, tx, err)
			}
			return s.overrideUsed(ctx, tx, tc, f, capGroupUpdate, Target{Kind: kindGroup, Group: g}, opRenameGroup, tenancy.GroupID{})
		})
	if err != nil {
		return tenancy.Group{}, err
	}
	return out, nil
}

// ChangeGroupSlug changes an exact Group's slug (group.update), with the accepted slug contract: setting the current
// slug is a successful no-op, and a collision is conflict.
func (s *Service) ChangeGroupSlug(ctx context.Context, tc *TenantContext, g tenancy.GroupID, slug string) (tenancy.Group, error) {
	var out tenancy.Group
	pre := func() *failure {
		if f := targetID(g)(); f != nil {
			return f
		}
		if err := tenancy.ValidateSlug(slug); err != nil {
			return inputFailure(err, reasonInputInvalid)
		}
		return nil
	}
	err := s.guard(ctx, tc, opChangeGroupSlug, capGroupUpdate, true, Target{Kind: kindGroup, Group: g}, groupQuery(g, true), pre,
		func(tx *trackedTx, f Facts) *failure {
			var err error
			if out, err = tenancy.ChangeGroupSlugTx(ctx, tx, f.Organization.ID, g, slug); err != nil {
				return writeFailure(ctx, tx, err)
			}
			return s.overrideUsed(ctx, tx, tc, f, capGroupUpdate, Target{Kind: kindGroup, Group: g}, opChangeGroupSlug, tenancy.GroupID{})
		})
	if err != nil {
		return tenancy.Group{}, err
	}
	return out, nil
}

// GetBusinessProfile reads an exact BusinessProfile of the active Organization (business_profile.profile.read). It
// returns only the profile: never its associations, access paths or sibling memberships.
func (s *Service) GetBusinessProfile(ctx context.Context, tc *TenantContext, p tenancy.BusinessProfileID) (tenancy.BusinessProfile, error) {
	var out tenancy.BusinessProfile
	err := s.guard(ctx, tc, opGetBusinessProfile, capProfileRead, false, Target{Kind: kindProfile, Profile: p}, profileQuery(p), targetID(p),
		func(_ *trackedTx, f Facts) *failure {
			out = findProfile(f, p)
			return nil
		})
	if err != nil {
		return tenancy.BusinessProfile{}, err
	}
	return out, nil
}

// UpdateBusinessProfile replaces an exact BusinessProfile's context (business_profile.profile.update), with the
// accepted bounded context contract.
func (s *Service) UpdateBusinessProfile(ctx context.Context, tc *TenantContext, p tenancy.BusinessProfileID, c tenancy.BusinessContext) (tenancy.BusinessProfile, error) {
	var out tenancy.BusinessProfile
	pre := func() *failure {
		if f := targetID(p)(); f != nil {
			return f
		}
		if _, err := tenancy.ValidateBusinessContext(c); err != nil {
			return inputFailure(err, reasonInputInvalid)
		}
		return nil
	}
	err := s.guard(ctx, tc, opUpdateBusinessProfile, capProfileUpdate, true, Target{Kind: kindProfile, Profile: p}, profileQuery(p), pre,
		func(tx *trackedTx, f Facts) *failure {
			var err error
			if out, err = tenancy.UpdateBusinessProfileTx(ctx, tx, f.Organization.ID, p, c); err != nil {
				return writeFailure(ctx, tx, err)
			}
			return s.overrideUsed(ctx, tx, tc, f, capProfileUpdate, Target{Kind: kindProfile, Profile: p}, opUpdateBusinessProfile,
				tenancy.GroupID{})
		})
	if err != nil {
		return tenancy.BusinessProfile{}, err
	}
	return out, nil
}

// manageableProfiles is the manageable-profile set of the current context (G007 Plan r1 §3.3 pin d): every profile of
// the active Organization for which business_profile.select_for_onboarding allows. In Organization-only context it is
// deduplicated across the subject's attached Groups; selected-Group context limits only the Group-derived paths. It is
// unexported evidence for AC4/AC7 (no exported list before G018) and writes no record.
func (s *Service) manageableProfiles(ctx context.Context, tc *TenantContext) ([]tenancy.BusinessProfile, error) {
	sc, f := s.bind(ctx, tc)
	if f != nil {
		return nil, f.err()
	}
	sel := tc.selection()
	q := profileQuery(tenancy.BusinessProfileID{})(tc)
	q.Organization, q.User, q.AllProfiles = tc.organization, tc.subject, true
	var out []tenancy.BusinessProfile
	f = s.inContextTx(ctx, false, dbContext{user: tc.subject, organization: tc.organization, group: tc.group}, cleanupUnbounded, func(tx *trackedTx) *failure {
		if err := s.id.CheckLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
			return sessionFailure(ctx, tx, err)
		}
		facts, err := tenancy.ReadFacts(ctx, tx, q)
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if r := contextCheck(sel, facts); r != "" {
			return fail(CodeContextDenied, r)
		}
		for _, p := range facts.Profiles {
			if ok, _ := decide(capProfileSelect, sel, facts, Target{Kind: kindProfile, Organization: tc.organization, Profile: p.ID}); ok {
				out = append(out, p)
			}
		}
		return nil
	})
	if f != nil {
		return nil, f.err()
	}
	return out, nil
}
