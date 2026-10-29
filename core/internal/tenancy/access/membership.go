package access

import (
	"context"

	"core/internal/tenancy"
)

// G019 membership administration (Plan r1 §1.3, §1.4, §5.1; r2 addendum §7). Every operation is one guarded call: the
// capability of the scope's kind is decided on the locked facts, and the write's eligibility is re-enforced at the
// statement by app.role_bindings_authority_guard (00015), whose refusals map through membershipFailure. The facade's own
// checks on the locked facts give truthful feedback; the guard is the guarantee.

// Scope is the administered scope: the active Organization (Group nil) or an exact Group of it.
type Scope struct {
	Kind  tenancy.ScopeType
	Group *tenancy.GroupID
}

// Member is one member of a scope with action hints read from the same snapshot. Hints confer nothing: every action
// re-decides on its own locked facts.
type Member struct {
	UserID      tenancy.UserID
	DisplayName *string
	Role        string
	Self        bool
	Disabled    bool

	CanChangeRole, CanRemove, CanPromote, CanDemote, CanTransfer bool
}

// MemberPage is one keyset page of members; Next is the cursor of the following page (nil on the last page).
type MemberPage struct {
	Members []Member
	Next    *tenancy.UserID
}

// SelfMember is the subject's own binding in a scope (r2 addendum §7, membership.self): the role, whether it may leave,
// whether it is the scope's last owner, and whether it may use the administrative lists. Its DisplayName stays nil (D-A3,
// Result addendum A1-01 §4: no tenancy reader supplies it); consumers show a truthful fallback.
type SelfMember struct {
	Member
	CanLeave  bool
	LastOwner bool
	CanManage bool
}

// MemberFilterOwnerCandidates lists only the enabled admins of the scope: the promotion and transfer picker (C5).
const MemberFilterOwnerCandidates = "owner_candidates"

// roleNone is TransferOwnership's formerRole for the actor's departure from the scope.
const roleNone = "none"

// valid reports a well-formed scope: an Organization scope names no Group, a Group scope a non-zero one.
func (sc Scope) valid() bool {
	switch sc.Kind {
	case tenancy.ScopeOrganization:
		return sc.Group == nil
	case tenancy.ScopeGroup:
		return sc.Group != nil && !zero(*sc.Group)
	}
	return false
}

func (sc Scope) group() tenancy.GroupID {
	if sc.Group == nil {
		return tenancy.GroupID{}
	}
	return *sc.Group
}

// target is the capability target of the scope.
func (sc Scope) target() Target {
	if sc.Kind == tenancy.ScopeGroup {
		return Target{Kind: kindGroup, Group: sc.group()}
	}
	return Target{Kind: kindOrganization}
}

// scopeID is the scope_id of the scope's bindings in the active Organization org.
func (sc Scope) scopeID(org tenancy.OrganizationID) [16]byte {
	if sc.Kind == tenancy.ScopeGroup {
		return sc.group()
	}
	return org
}

func (sc Scope) manageMembers() Capability {
	if sc.Kind == tenancy.ScopeGroup {
		return capGroupManageMembers
	}
	return capOrganizationManageMembers
}

func (sc Scope) manageOwners() Capability {
	if sc.Kind == tenancy.ScopeGroup {
		return capGroupManageOwners
	}
	return capOrganizationManageOwners
}

func (sc Scope) leave() Capability {
	if sc.Kind == tenancy.ScopeGroup {
		return capGroupLeave
	}
	return capOrganizationLeave
}

// The role names of a scope.
func (sc Scope) ownerRole() string {
	if sc.Kind == tenancy.ScopeGroup {
		return string(tenancy.GroupOwner)
	}
	return string(tenancy.OrgOwner)
}

func (sc Scope) adminRole() string {
	if sc.Kind == tenancy.ScopeGroup {
		return string(tenancy.GroupAdmin)
	}
	return string(tenancy.OrgAdmin)
}

// query is the fact query of a scope operation. An Organization scope reads the Organization facts, locking the
// Organization row FOR NO KEY UPDATE for a write (the scope serialization, §1.5); a Group scope reads the target Group
// as TargetGroup (FOR NO KEY UPDATE for a write) unless the context gate will deny it. A write adds the Subject fact
// (subject's, and with withActor the actor's, scope binding locked FOR NO KEY UPDATE); owners adds the owner count.
func (sc Scope) query(write bool, subject *tenancy.UserID, withActor, owners bool) func(*TenantContext) tenancy.FactQuery {
	return func(tc *TenantContext) tenancy.FactQuery {
		var q tenancy.FactQuery
		if sc.Kind == tenancy.ScopeGroup {
			g := sc.group()
			q = groupQuery(g, write)(tc)
			if !zero(tc.group) && tc.group != g {
				return q // the context gate denies; nothing is locked exclusively and no subject is read
			}
			q.TargetGroup = g
		} else {
			q = organizationQuery(tc)
			q.ExclusiveOrganization = write
		}
		if write && subject != nil {
			q.Subject, q.SubjectAndActor = subject, withActor
		}
		q.OwnerCount = owners
		return q
	}
}

// scopePre validates the scope, then runs more (nil: nothing else).
func scopePre(sc Scope, more func() *failure) func() *failure {
	return func() *failure {
		if !sc.valid() {
			return &failure{code: CodeInvalidInput, field: fieldScope, reason: reasonInputInvalid}
		}
		if more != nil {
			return more()
		}
		return nil
	}
}

// organizationOnly refuses an Organization departure or removal in a selected-Group context (C3): such a context hides
// sibling Group bindings, so the dependent removal could not be complete.
func organizationOnly(tc *TenantContext, sc Scope) *failure {
	if sc.Kind == tenancy.ScopeOrganization && tc != nil && !zero(tc.group) {
		return &failure{code: CodeInvalidInput, field: fieldContext, reason: reasonInputInvalid}
	}
	return nil
}

// subjectBinding returns user's binding of the scope among the locked Subject facts.
func subjectBinding(f Facts, sc Scope, user tenancy.UserID) (tenancy.Binding, bool) {
	id := sc.scopeID(f.Organization.ID)
	for _, b := range f.SubjectBindings {
		if b.SubjectUserID == user && b.ScopeType == sc.Kind && b.ScopeID == id {
			return b, true
		}
	}
	return tenancy.Binding{}, false
}

// ownRole returns the subject's own role in the scope from the ordinary facts ("" when it holds none).
func ownRole(f Facts, sc Scope, tc *TenantContext) string {
	if sc.Kind == tenancy.ScopeGroup {
		roles, _ := groupRoles(tc.selection(), f)
		return string(roles[sc.group()])
	}
	r, _ := organizationRole(tc.selection(), f)
	return string(r)
}

// ordinaryRole reports whether role is an admin or member role of the scope (never an owner role).
func ordinaryRole(sc Scope, role string) bool { return tenancy.InvitationRoleValid(sc.Kind, role) }

func notFound() *failure { return fail(CodeNotFound, reasonTargetNotFound) }

// ownerSubject is the owner_subject refusal of an administrative role change or removal (forbidden, role_not_permitted)
// after the scope's capability allowed it: a recordable denial (AUDIT.md), marked on the locked facts with the
// administered scope (the Organization or the exact Group) as its target, never the subject.
func ownerSubject(tc *TenantContext, sc Scope, f Facts) *failure {
	t := sc.target()
	t.Organization = f.Organization.ID
	res, id, g := targetScope(t)
	return markDenied(&failure{code: CodeForbidden, field: fieldOwnerSubject, reason: reasonRoleNotPermitted}, tc.selection(), f, res, id, g)
}

func conflict(field string) *failure {
	return &failure{code: CodeConflict, field: field, reason: reasonWriteConflict}
}

func invalid(field string) *failure {
	return &failure{code: CodeInvalidInput, field: field, reason: reasonInputInvalid}
}

// pageLimit validates a page limit: 0 is MaxMembersPage, otherwise 1..MaxMembersPage.
func pageLimit(limit int) (int, *failure) {
	switch {
	case limit == 0:
		return tenancy.MaxMembersPage, nil
	case limit < 1 || limit > tenancy.MaxMembersPage:
		return 0, invalid(fieldLimit)
	}
	return limit, nil
}

// cursor parses a keyset cursor: "" is the start of the list.
func cursor[T ~[16]byte](after string, parse func(string) (T, error)) (T, *failure) {
	if after == "" {
		return T{}, nil
	}
	id, err := parse(after)
	if err != nil || zero(id) {
		return T{}, invalid(fieldCursor)
	}
	return id, nil
}

// ListMembers reads one keyset page of the scope's members (organization.manage_members or group.manage_members), with
// hints from the same snapshot. filter "" lists everyone; MemberFilterOwnerCandidates only the enabled same-scope
// admins. limit 0 is MaxMembersPage.
func (s *Service) ListMembers(ctx context.Context, tc *TenantContext, sc Scope, after string, limit int, filter string) (MemberPage, error) {
	var out MemberPage
	var from tenancy.UserID
	pre := scopePre(sc, func() *failure {
		var f *failure
		if limit, f = pageLimit(limit); f != nil {
			return f
		}
		if from, f = cursor(after, tenancy.ParseUserID); f != nil {
			return f
		}
		if filter != "" && filter != MemberFilterOwnerCandidates {
			return invalid(fieldFilter)
		}
		return nil
	})
	err := s.guard(ctx, tc, opMembershipList, sc.manageMembers(), false, sc.target(), sc.query(false, nil, false, true), pre,
		func(tx *trackedTx, f Facts) *failure {
			rows, err := tenancy.MembersPage(ctx, tx, f.Organization.ID, sc.group(), from, limit, filter == MemberFilterOwnerCandidates)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			if len(rows) > limit {
				next := rows[limit-1].UserID
				out.Next, rows = &next, rows[:limit]
			}
			t := sc.target()
			t.Organization = f.Organization.ID
			owners, _ := decide(sc.manageOwners(), tc.selection(), f, t)
			actorOwner := ownRole(f, sc, tc) == sc.ownerRole()
			out.Members = make([]Member, 0, len(rows))
			for _, r := range rows {
				out.Members = append(out.Members, memberHints(sc, tc.subject, r, owners, actorOwner, f.OwnerCount))
			}
			return nil
		})
	if err != nil {
		return MemberPage{}, err
	}
	return out, nil
}

// memberHints derives one member's action hints: ordinary changes for non-owners other than the actor; promotion of an
// enabled admin and demotion of an owner while another owner remains, with the owner capability; transfer to an enabled
// admin by an owner.
func memberHints(sc Scope, actor tenancy.UserID, r tenancy.Member, owners, actorOwner bool, ownerCount int) Member {
	m := Member{UserID: r.UserID, DisplayName: r.DisplayName, Role: r.Role, Self: r.UserID == actor, Disabled: r.Disabled}
	owner, admin := r.Role == sc.ownerRole(), r.Role == sc.adminRole()
	m.CanChangeRole = !m.Self && !owner
	m.CanRemove = !m.Self && !owner
	m.CanPromote = owners && admin && !r.Disabled
	m.CanDemote = owners && owner && ownerCount > 1
	m.CanTransfer = owners && actorOwner && admin && !r.Disabled && !m.Self
	return m
}

// SelfMembership reads the subject's own binding in the scope (no capability; membership.self, r2 addendum §7): a
// non-member, an unknown or foreign Group or one outside the selected context is not_found.
func (s *Service) SelfMembership(ctx context.Context, tc *TenantContext, sc Scope) (SelfMember, error) {
	var out SelfMember
	err := s.guard(ctx, tc, opMembershipSelf, noCapability, false, sc.target(), sc.query(false, nil, false, true), scopePre(sc, nil),
		func(_ *trackedTx, f Facts) *failure {
			sel := tc.selection()
			if sc.Kind == tenancy.ScopeGroup && (!hasGroup(f, sel.Organization, sc.group()) || (!zero(sel.Group) && sel.Group != sc.group())) {
				return notFound()
			}
			role := ownRole(f, sc, tc)
			if role == "" {
				return notFound()
			}
			t := sc.target()
			t.Organization = f.Organization.ID
			leave, _ := decide(sc.leave(), sel, f, t)
			manage, _ := decide(sc.manageMembers(), sel, f, t)
			last := role == sc.ownerRole() && f.OwnerCount <= 1
			out = SelfMember{Member: Member{UserID: tc.subject, Role: role, Self: true}, CanLeave: leave && !last, LastOwner: last,
				CanManage: manage}
			return nil
		})
	if err != nil {
		return SelfMember{}, err
	}
	return out, nil
}

// ChangeMemberRole sets an admin or member subject's role to the other ordinary role (organization.manage_members or
// group.manage_members). The actor cannot target itself (self_target) and an owner subject is forbidden (owner_subject)
// whatever the requested role (Spec R1); ownership changes are PromoteToOwner, DemoteOwner and TransferOwnership.
func (s *Service) ChangeMemberRole(ctx context.Context, tc *TenantContext, sc Scope, subject tenancy.UserID, role string) (Member, error) {
	var out Member
	pre := scopePre(sc, func() *failure {
		if f := targetID(subject)(); f != nil {
			return f
		}
		if tc != nil && subject == tc.subject {
			return invalid(fieldSelfTarget)
		}
		if !ordinaryRole(sc, role) {
			return invalid(string(tenancy.CodeInvalidRole))
		}
		return nil
	})
	err := s.guard(ctx, tc, opMembershipChangeRole, sc.manageMembers(), true, sc.target(), sc.query(true, &subject, false, false), pre,
		func(tx *trackedTx, f Facts) *failure {
			b, ok := subjectBinding(f, sc, subject)
			if !ok {
				return notFound()
			}
			if b.Role == sc.ownerRole() {
				return ownerSubject(tc, sc, f)
			}
			nb, err := tenancy.ChangeRoleTx(ctx, tx, b, role)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			out = Member{UserID: subject, Role: nb.Role}
			return s.scopeOverrideUsed(ctx, tx, tc, f, sc, opMembershipChangeRole)
		})
	if err != nil {
		return Member{}, err
	}
	return out, nil
}

// RemoveMember removes a non-owner subject from the scope (organization.manage_members or group.manage_members). An
// Organization removal needs an Organization-only context (field context); deleting the Organization binding removes
// the subject's Group bindings of the Organization first, in ascending group_id, inside the same statement (the
// authority guard's step (3), 00017 amendment01); a subject who owns any Group of the Organization is forbidden
// (owner_subject, raised by that step): that Group's succession comes first (Spec R3).
func (s *Service) RemoveMember(ctx context.Context, tc *TenantContext, sc Scope, subject tenancy.UserID) error {
	pre := scopePre(sc, func() *failure {
		if f := targetID(subject)(); f != nil {
			return f
		}
		if tc != nil && subject == tc.subject {
			return invalid(fieldSelfTarget)
		}
		return organizationOnly(tc, sc)
	})
	return s.guard(ctx, tc, opMembershipRemove, sc.manageMembers(), true, sc.target(), sc.query(true, &subject, false, false), pre,
		func(tx *trackedTx, f Facts) *failure {
			b, ok := subjectBinding(f, sc, subject)
			if !ok {
				return notFound()
			}
			if b.Role == sc.ownerRole() {
				return ownerSubject(tc, sc, f)
			}
			if err := tenancy.RemoveBindingTx(ctx, tx, b); err != nil {
				if fl := membershipFailure(ctx, tx, err); fl == nil || fl.field != fieldOwnerSubject {
					return fl
				}
				// The guard's dependent Group-owner refusal is the Organization removal's recordable denial.
				return ownerSubject(tc, sc, f)
			}
			return s.scopeOverrideUsed(ctx, tx, tc, f, sc, opMembershipRemove)
		})
}

// scopeOverrideUsed is overrideUsed for a Group scope's group.manage_members write (G060); an Organization scope never
// uses the override.
func (s *Service) scopeOverrideUsed(ctx context.Context, tx *trackedTx, tc *TenantContext, f Facts, sc Scope, op string) *failure {
	if sc.Kind != tenancy.ScopeGroup {
		return nil
	}
	return s.overrideUsed(ctx, tx, tc, f, capGroupManageMembers, sc.target(), op, tenancy.GroupID{})
}

// LeaveScope removes the subject's own binding of the scope (organization.leave or group.leave). The last owner of the
// scope is refused (last_owner). An Organization departure needs an Organization-only context; the guard's step (3)
// removes the subject's own Group bindings of the Organization in ascending group_id inside the same DELETE and refuses
// the departure of a Group's last owner before the Organization row is deleted (last_owner; the statement rolls back).
func (s *Service) LeaveScope(ctx context.Context, tc *TenantContext, sc Scope) error {
	var self tenancy.UserID
	if tc != nil {
		self = tc.subject
	}
	pre := scopePre(sc, func() *failure { return organizationOnly(tc, sc) })
	return s.guard(ctx, tc, opMembershipLeave, sc.leave(), true, sc.target(), sc.query(true, &self, false, true), pre,
		func(tx *trackedTx, f Facts) *failure {
			b, ok := subjectBinding(f, sc, self)
			if !ok {
				return notFound()
			}
			if b.Role == sc.ownerRole() && f.OwnerCount <= 1 {
				return conflict(string(tenancy.CodeLastOwner))
			}
			if err := tenancy.RemoveBindingTx(ctx, tx, b); err != nil {
				return membershipFailure(ctx, tx, err)
			}
			return nil
		})
}
