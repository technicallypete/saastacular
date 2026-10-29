package access

import (
	"bytes"
	"context"
	"crypto/rand"
	"time"

	"core/internal/audit"
	"core/internal/tenancy"
)

// G018 guarded onboarding (Plan r5 §1.3, §1.4, §5.1): Group creation under group.create and onboarding completion under
// group.complete_onboarding (with business_profile.select_for_onboarding for an existing profile). Each operation is one
// guarded transaction on its own facts; the database policies of 00014 admit exactly the statements the tenancy
// primitives execute, so a facade bypass meets the same refusal.

// GroupOnboarding is a Group with its derived onboarding state: complete exactly when the Group has its single
// Business Profile association. There is no status column and no client marker.
type GroupOnboarding struct {
	Group             tenancy.Group
	BusinessProfileID *tenancy.BusinessProfileID // nil while onboarding is incomplete
	Complete          bool                       // BusinessProfileID != nil
	// G060 display hints from the same facts (Plan r0 §4.1, D-9): the subject's own Group role (zero when it holds none),
	// whether it may administer the Group's ordinary metadata, members and invitations, and through which path. They
	// authorize nothing: every action re-decides on its own locked facts. An override is never a self membership.
	SelfRole         tenancy.GroupRole
	CanManage        bool
	ManagementSource string // "group_role", "organization_override" or "" (none)
}

// The G060 management sources of GroupOnboarding.ManagementSource.
const (
	managementGroupRole            = "group_role"
	managementOrganizationOverride = "organization_override"
)

func onboardingOf(f Facts, g tenancy.GroupID) GroupOnboarding {
	out := GroupOnboarding{Group: findGroup(f, g)}
	if p, ok := association(f, g); ok {
		out.BusinessProfileID, out.Complete = &p, true
	}
	return out
}

// withAuthority sets the G060 hints of a Group on the facts of its decision: the own role from the subject's Group
// bindings and group.manage_members decided on the same facts (the explicit Group path first, the override second).
func withAuthority(out GroupOnboarding, sel contextSelection, f Facts) GroupOnboarding {
	g := out.Group.ID
	if roles, ok := groupRoles(sel, f); ok {
		out.SelfRole = roles[g]
	}
	ok, r := decide(capGroupManageMembers, sel, f, Target{Kind: kindGroup, Organization: sel.Organization, Group: g})
	out.CanManage, out.ManagementSource = ok, ""
	switch {
	case ok && r == reasonAllowedOverride:
		out.ManagementSource = managementOrganizationOverride
	case ok:
		out.ManagementSource = managementGroupRole
	}
	return out
}

// newKey returns a random UUID v4 for a keyed insert (the Group and profile ids are generated here, as object ids are).
func newKey[T ~[16]byte]() (T, bool) {
	var id T
	if _, err := rand.Read(id[:]); err != nil {
		return T{}, false
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return id, !zero(id)
}

// onboardingQuery is the fact query of a completion (Plan r5 §1.4): the target Group locked FOR NO KEY UPDATE in the
// ascending Group-id order and its association read after the lock. For an existing profile p it also reads every Group
// of the subject (the candidate paths, locked FOR SHARE around the target) with their associations and the profile. In
// selected-Group context the candidates are the selected Group; a target outside it is read without an exclusive lock
// (the context gate denies it).
func onboardingQuery(g tenancy.GroupID, p tenancy.BusinessProfileID) func(*TenantContext) tenancy.FactQuery {
	return func(tc *TenantContext) tenancy.FactQuery {
		q := tenancy.FactQuery{Profile: p, Associations: true}
		switch {
		case !zero(tc.group) && tc.group != g:
			q.Groups = []tenancy.GroupID{g, tc.group}
			return q
		case !zero(tc.group) || zero(p):
			q.Groups = []tenancy.GroupID{g}
		}
		q.TargetGroup = g
		return q
	}
}

// onboardingReadQuery reads the target Group (and the selected Group) with their associations, without locks.
func onboardingReadQuery(g tenancy.GroupID) func(*TenantContext) tenancy.FactQuery {
	return func(tc *TenantContext) tenancy.FactQuery {
		q := groupQuery(g, false)(tc)
		q.Associations = true
		return q
	}
}

// createFailure maps a creation write failure: a taken slug is conflict on the slug field.
func createFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	if tenancy.CodeOf(err) == tenancy.CodeSlugConflict {
		return &failure{code: CodeConflict, field: "slug", reason: reasonWriteConflict}
	}
	return writeFailure(ctx, tx, err)
}

// CreateGroup creates a Group in the active Organization (group.create: every current Organization membership role,
// Organization-only context). The verified context user becomes its initial group_owner in the same statement (the
// 00014 bootstrap trigger); no Organization binding is written and there is no creator parameter. An empty slug takes
// the deterministic suggestion base of the name. A taken slug is conflict (field slug); unavailable may be durable.
func (s *Service) CreateGroup(ctx context.Context, tc *TenantContext, name, slug string) (GroupOnboarding, error) {
	var out GroupOnboarding
	pre := func() *failure {
		if _, err := tenancy.ValidateName(name); err != nil {
			return inputFailure(err, reasonInputInvalid)
		}
		if slug != "" {
			if err := tenancy.ValidateSlug(slug); err != nil {
				return inputFailure(err, reasonInputInvalid)
			}
		}
		return nil
	}
	err := s.guard(ctx, tc, opCreateGroup, capGroupCreate, true, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts) *failure {
			id, ok := newKey[tenancy.GroupID]()
			if !ok {
				return fail(CodeInternal, reasonStoreUnavailable)
			}
			g, err := tenancy.CreateGroupTx(ctx, tx, f.Organization.ID, id, name, slug)
			if err != nil {
				return createFailure(ctx, tx, err)
			}
			// The 00014 bootstrap trigger made the verified context user the Group's owner in the same statement.
			out = GroupOnboarding{Group: g, SelfRole: tenancy.GroupOwner, CanManage: true, ManagementSource: managementGroupRole}
			return nil
		})
	if err != nil {
		return GroupOnboarding{}, err
	}
	return out, nil
}

// GetGroupOnboarding reads an exact Group with its onboarding state (group.read).
func (s *Service) GetGroupOnboarding(ctx context.Context, tc *TenantContext, g tenancy.GroupID) (GroupOnboarding, error) {
	var out GroupOnboarding
	err := s.guard(ctx, tc, opGetGroupOnboarding, capGroupRead, false, Target{Kind: kindGroup, Group: g}, onboardingReadQuery(g), targetID(g),
		func(_ *trackedTx, f Facts) *failure {
			out = withAuthority(onboardingOf(f, g), tc.selection(), f)
			return nil
		})
	if err != nil {
		return GroupOnboarding{}, err
	}
	return out, nil
}

// completed is the state check on the locked facts: a Group that already has its profile is conflict
// (association_exists) with zero writes; the association unique key is the backstop.
func completed(f Facts, g tenancy.GroupID) *failure {
	if _, ok := association(f, g); ok {
		return &failure{code: CodeConflict, field: string(tenancy.CodeAssociationExists), reason: reasonWriteConflict}
	}
	return nil
}

// The completion budgets (availability addendum to Plan r5; AUTHORIZATION.md "Availability: denied exclusive-write
// requests"): a completion locks its target Group FOR NO KEY UPDATE before the decision, so its transaction, BEGIN
// through COMMIT, runs under a fixed decision budget and its deferred rollback under a fixed cleanup budget (the
// storageTx pattern). The denial audit keeps its own detached auditRecordBudget and auditCleanupBudget.
const (
	completionDecisionBudget = 5 * time.Second
	completionCleanupBudget  = 5 * time.Second
)

// guardBounded is guard for the two onboarding completions: guarded runs under ctx bounded by completionDecisionBudget
// (an earlier caller deadline is kept) with completionCleanupBudget for its rollback, and act receives that bounded
// context, so every statement of the completion's writes and their failure classification run under the decision
// budget. The access.decision record, the denial audit and the operational record keep the caller's ctx, as in guard.
// A deadline that interrupts COMMIT keeps inTx's commit classification (possibly durable: unavailable).
func (s *Service) guardBounded(ctx context.Context, tc *TenantContext, op string, c Capability, t Target,
	query func(*TenantContext) tenancy.FactQuery, pre func() *failure, act func(ctx context.Context, tx *trackedTx, f Facts) *failure) error {
	dctx, cancel := context.WithTimeout(ctx, completionDecisionBudget)
	var allow Reason
	f := s.guardedTx(dctx, tc, c, true, t, query, pre, completionCleanupBudget, &allow,
		func(tx *trackedTx, f Facts) *failure { return act(dctx, tx, f) })
	cancel()
	s.recordGuarded(ctx, tc, op, f)
	s.logGuardedAllow(ctx, op, c, allow, f)
	if f != nil {
		return f.err()
	}
	return nil
}

// CompleteGroupOnboardingWithNewProfile completes an incomplete Group with a new Business Profile created from c
// (group.complete_onboarding). The profile and its association commit together or not at all; the database admits the
// association because the profile was inserted by this transaction (00014 app.rls_profile_eligible).
func (s *Service) CompleteGroupOnboardingWithNewProfile(ctx context.Context, tc *TenantContext, g tenancy.GroupID, c tenancy.BusinessContext) (GroupOnboarding, error) {
	var out GroupOnboarding
	pre := func() *failure {
		if f := targetID(g)(); f != nil {
			return f
		}
		if _, err := tenancy.ValidateBusinessContext(c); err != nil {
			return inputFailure(err, reasonInputInvalid)
		}
		return nil
	}
	// act's ctx is the decision-bounded context (guardBounded): it shadows the caller's for every write below.
	err := s.guardBounded(ctx, tc, opCompleteOnboardingNew, capGroupCompleteOnboarding, Target{Kind: kindGroup, Group: g},
		onboardingQuery(g, tenancy.BusinessProfileID{}), pre,
		func(ctx context.Context, tx *trackedTx, f Facts) *failure {
			if fl := completed(f, g); fl != nil {
				return fl
			}
			p, ok := newKey[tenancy.BusinessProfileID]()
			if !ok {
				return fail(CodeInternal, reasonStoreUnavailable)
			}
			if _, err := tenancy.CreateBusinessProfileTx(ctx, tx, f.Organization.ID, p, c); err != nil {
				return writeFailure(ctx, tx, err)
			}
			if _, err := tenancy.LinkGroupBusinessProfileTx(ctx, tx, f.Organization.ID, g, p); err != nil {
				return writeFailure(ctx, tx, err)
			}
			out = withAuthority(GroupOnboarding{Group: findGroup(f, g), BusinessProfileID: &p, Complete: true}, tc.selection(), f)
			return s.overrideUsed(ctx, tx, tc, f, capGroupCompleteOnboarding, Target{Kind: kindGroup, Group: g}, opCompleteOnboardingNew,
				tenancy.GroupID{})
		})
	if err != nil {
		return GroupOnboarding{}, err
	}
	return out, nil
}

// CompleteGroupOnboardingWithProfile completes an incomplete Group with an existing Business Profile p
// (group.complete_onboarding on g, then business_profile.select_for_onboarding on p, on the same locked facts). An
// unreadable profile is not_found; a readable one without a qualifying path is forbidden. The new Group itself never
// qualifies a profile.
func (s *Service) CompleteGroupOnboardingWithProfile(ctx context.Context, tc *TenantContext, g tenancy.GroupID, p tenancy.BusinessProfileID) (GroupOnboarding, error) {
	var out GroupOnboarding
	pre := func() *failure {
		if f := targetID(g)(); f != nil {
			return f
		}
		return targetID(p)()
	}
	// act's ctx is the decision-bounded context (guardBounded): it shadows the caller's for the write below.
	err := s.guardBounded(ctx, tc, opCompleteOnboardingExisting, capGroupCompleteOnboarding, Target{Kind: kindGroup, Group: g},
		onboardingQuery(g, p), pre,
		func(ctx context.Context, tx *trackedTx, f Facts) *failure {
			sel := tc.selection()
			t := Target{Kind: kindProfile, Organization: tc.organization, Profile: p}
			if ok, r := decide(capProfileSelect, sel, f, t); !ok {
				res, id, grp := targetScope(t)
				return markDenied(denial(capProfileSelect, sel, f, t, r), sel, f, res, id, grp)
			}
			if fl := completed(f, g); fl != nil {
				return fl
			}
			if _, err := tenancy.LinkGroupBusinessProfileTx(ctx, tx, f.Organization.ID, g, p); err != nil {
				return writeFailure(ctx, tx, err)
			}
			out = withAuthority(GroupOnboarding{Group: findGroup(f, g), BusinessProfileID: &p, Complete: true}, tc.selection(), f)
			// G060 (r1 C5): one record per override-allowed decision, each naming its actual target: the Group, and the
			// selected profile with that Group.
			if fl := s.overrideUsed(ctx, tx, tc, f, capGroupCompleteOnboarding, Target{Kind: kindGroup, Group: g},
				opCompleteOnboardingExisting, tenancy.GroupID{}); fl != nil {
				return fl
			}
			return s.overrideUsed(ctx, tx, tc, f, capProfileSelect, t, opCompleteOnboardingExisting, g)
		})
	if err != nil {
		return GroupOnboarding{}, err
	}
	return out, nil
}

// G018 slice 2 (Plan r5 §1.5, §5.1): the identifier and context types of the onboarding surface, named here so that
// its HTTP adapter (core/internal/onboarding/api) never imports the tenancy package (the importer boundary).
type (
	GroupID           = tenancy.GroupID
	BusinessProfileID = tenancy.BusinessProfileID
	BusinessContext   = tenancy.BusinessContext
)

// ParseGroupID parses a canonical Group UUID text; a malformed identifier is an error.
func ParseGroupID(s string) (GroupID, error) { return tenancy.ParseGroupID(s) }

// ParseBusinessProfileID parses a canonical Business Profile UUID text; a malformed identifier is an error.
func ParseBusinessProfileID(s string) (BusinessProfileID, error) {
	return tenancy.ParseBusinessProfileID(s)
}

// OnboardingGroupPage is one ListGroups page (ascending Group id); Next is the cursor of the following page (the last
// scanned id), nil on the final page. (access.GroupPage is the G014 MCP selectable-Group page.)
type OnboardingGroupPage struct {
	Groups []GroupOnboarding
	Next   *tenancy.GroupID
}

// BusinessProfileSummary is one row of the Business Profiles list: never the full record, its associations or its Groups.
// G060 (Plan r1 C2): CanSelect (business_profile.select_for_onboarding) and CanUpdate (business_profile.profile.update)
// are display hints decided on the same facts; they authorize nothing.
type BusinessProfileSummary struct {
	ID          tenancy.BusinessProfileID
	DisplayName string
	Industry    *string
	CanSelect   bool
	CanUpdate   bool
}

// ProfilePage is one ListBusinessProfiles page; Next is the last scanned id (a page may hold fewer admitted rows than
// its limit while Next is present), nil on the final page.
type ProfilePage struct {
	Profiles []BusinessProfileSummary
	Next     *tenancy.BusinessProfileID
}

// pageInput validates a list page request: limit 1..100 (invalid_page_size) and an empty or canonical UUID cursor
// (invalid_cursor), as the G014 keyset pattern.
func pageInput[T ~[16]byte](after string, limit int, parse func(string) (T, error)) (T, *failure) {
	var cur T
	if limit < 1 || limit > tenancy.MaxOnboardingPage {
		return cur, &failure{code: CodeInvalidInput, field: "invalid_page_size", reason: reasonInputInvalid}
	}
	if after != "" {
		c, err := parse(after)
		if err != nil {
			return cur, &failure{code: CodeInvalidInput, field: "invalid_cursor", reason: reasonInputInvalid}
		}
		cur = c
	}
	return cur, nil
}

// listed runs one list operation: the binding check, then one read-only transaction with the session liveness, the
// facts, the context revalidation and act, which decides each row with capability c on those facts (there is no single
// target). It writes the access.decision record of op (capability c) and, for a recordable denial, the access.denied
// event, as guard does.
func (s *Service) listed(ctx context.Context, tc *TenantContext, op string, c Capability, query func(*TenantContext) tenancy.FactQuery,
	pre func() *failure, act func(tx *trackedTx, f Facts, sel contextSelection) *failure) error {
	f := func() *failure {
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
		q.Organization, q.User = tc.organization, tc.subject
		return s.inContextTx(ctx, false, dbContext{user: tc.subject, organization: tc.organization, group: tc.group}, cleanupUnbounded, func(tx *trackedTx) *failure {
			if err := s.id.CheckLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
				return sessionFailure(ctx, tx, err)
			}
			facts, err := tenancy.ReadFacts(ctx, tx, q)
			if err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
			if r := contextCheck(sel, facts); r != "" {
				return markDenied(fail(CodeContextDenied, r), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
			}
			return act(tx, facts, sel)
		})
	}()
	s.recordGuarded(ctx, tc, op, f)
	s.logGuarded(ctx, op, c, f)
	if f != nil {
		return f.err()
	}
	return nil
}

// SuggestGroupSlug returns an advisory slug for a Group name in the active Organization (group.create, Organization-only
// context): the name's base or the first free base-2 ... base-99 among the slugs the context can read. It reserves
// nothing; CreateGroup's unique key decides (conflict, field slug).
func (s *Service) SuggestGroupSlug(ctx context.Context, tc *TenantContext, name string) (string, error) {
	var out string
	pre := func() *failure {
		if _, err := tenancy.ValidateName(name); err != nil {
			return inputFailure(err, reasonInputInvalid)
		}
		return nil
	}
	err := s.guard(ctx, tc, opSuggestGroupSlug, capGroupCreate, false, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts) *failure {
			v, err := tenancy.SuggestGroupSlugTx(ctx, tx, f.Organization.ID, name)
			if err != nil {
				if c := tenancy.CodeOf(err); c == tenancy.CodeSlugRequired || c == tenancy.CodeSlugUnavailable {
					return inputFailure(err, reasonInputInvalid)
				}
				return writeFailure(ctx, tx, err)
			}
			out = v
			return nil
		})
	if err != nil {
		return "", err
	}
	return out, nil
}

// ListGroups returns one page of the active Organization's Groups the context may read (group.read per row), ordered by
// id after the cursor, each with its derived onboarding state: facts, one page statement and one associations
// statement. A selected-Group context lists only that Group (from its facts). A non-empty slug is the exact filter of
// conflict recovery.
func (s *Service) ListGroups(ctx context.Context, tc *TenantContext, after string, limit int, slug string) (OnboardingGroupPage, error) {
	var out OnboardingGroupPage
	var cur tenancy.GroupID
	pre := func() *failure {
		var f *failure
		if cur, f = pageInput(after, limit, tenancy.ParseGroupID); f != nil {
			return f
		}
		if slug != "" {
			if err := tenancy.ValidateSlug(slug); err != nil {
				return inputFailure(err, reasonInputInvalid)
			}
		}
		return nil
	}
	query := func(tc *TenantContext) tenancy.FactQuery {
		if !zero(tc.group) {
			return tenancy.FactQuery{Groups: []tenancy.GroupID{tc.group}, Associations: true}
		}
		return tenancy.FactQuery{} // every Group binding of the subject, and its Groups
	}
	err := s.listed(ctx, tc, opListGroups, capGroupRead, query, pre, func(tx *trackedTx, f Facts, sel contextSelection) *failure {
		out.Groups = []GroupOnboarding{}
		if !zero(sel.Group) {
			g := findGroup(f, sel.Group)
			if bytes.Compare(g.ID[:], cur[:]) > 0 && (slug == "" || g.Slug == slug) {
				if ok, _ := decide(capGroupRead, sel, f, Target{Kind: kindGroup, Organization: tc.organization, Group: g.ID}); ok {
					out.Groups = append(out.Groups, withAuthority(onboardingOf(f, g.ID), sel, f))
				}
			}
			return nil
		}
		groups, assoc, err := tenancy.GroupsPage(ctx, tx, tc.organization, cur, limit, slug)
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if len(groups) > limit {
			groups = groups[:limit]
			next := groups[limit-1].ID
			out.Next = &next
		}
		pf := f
		pf.Groups, pf.Associations = groups, assoc
		for _, g := range groups {
			if ok, _ := decide(capGroupRead, sel, pf, Target{Kind: kindGroup, Organization: tc.organization, Group: g.ID}); ok {
				out.Groups = append(out.Groups, withAuthority(onboardingOf(pf, g.ID), sel, pf))
			}
		}
		return nil
	})
	if err != nil {
		return OnboardingGroupPage{}, err
	}
	return out, nil
}

// ListBusinessProfiles is the Organization's Business Profiles page: the profiles scanned ascending from the cursor (one
// page statement after the facts) for which business_profile.profile.read allows on the facts (G060 Plan r1 C2: an
// Organization owner or admin every ordinary profile of the Organization, metadata only; a Group role the profiles linked
// to its Groups, restricted to the selected Group in selected-Group context; foreign tenants never). Each row carries the
// can_select and can_update hints decided on the same facts; listing never authorizes attaching or editing, which
// re-decide on their own locked facts. Rows are summaries; Next is the last scanned id.
func (s *Service) ListBusinessProfiles(ctx context.Context, tc *TenantContext, after string, limit int) (ProfilePage, error) {
	var out ProfilePage
	var cur tenancy.BusinessProfileID
	pre := func() (f *failure) {
		cur, f = pageInput(after, limit, tenancy.ParseBusinessProfileID)
		return f
	}
	query := func(tc *TenantContext) tenancy.FactQuery {
		q := profileQuery(tenancy.BusinessProfileID{})(tc)
		q.Associations = true
		return q
	}
	err := s.listed(ctx, tc, opListBusinessProfiles, capProfileRead, query, pre, func(tx *trackedTx, f Facts, sel contextSelection) *failure {
		rows, err := tenancy.ProfilesPage(ctx, tx, tc.organization, cur, limit)
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if len(rows) > limit {
			rows = rows[:limit]
			next := rows[limit-1].ID
			out.Next = &next
		}
		pf := f
		pf.Profiles = rows
		out.Profiles = []BusinessProfileSummary{}
		for _, p := range rows {
			t := Target{Kind: kindProfile, Organization: tc.organization, Profile: p.ID}
			if ok, _ := decide(capProfileRead, sel, pf, t); ok {
				canSelect, _ := decide(capProfileSelect, sel, pf, t)
				canUpdate, _ := decide(capProfileUpdate, sel, pf, t)
				out.Profiles = append(out.Profiles, BusinessProfileSummary{ID: p.ID, DisplayName: p.Context.DisplayName, Industry: p.Context.Industry,
					CanSelect: canSelect, CanUpdate: canUpdate})
			}
		}
		return nil
	})
	if err != nil {
		return ProfilePage{}, err
	}
	return out, nil
}

// GetBusinessProfileView is GetBusinessProfile (business_profile.profile.read, the same operation and record) with the
// can_update display hint (Plan r5 §5.3, Q3): a second decision of business_profile.profile.update on the same
// snapshot. The hint authorizes nothing: UpdateBusinessProfile always re-decides on fresh locked facts.
func (s *Service) GetBusinessProfileView(ctx context.Context, tc *TenantContext, p tenancy.BusinessProfileID) (tenancy.BusinessProfile, bool, error) {
	var out tenancy.BusinessProfile
	var canUpdate bool
	err := s.guard(ctx, tc, opGetBusinessProfile, capProfileRead, false, Target{Kind: kindProfile, Profile: p}, profileQuery(p), targetID(p),
		func(_ *trackedTx, f Facts) *failure {
			out = findProfile(f, p)
			canUpdate, _ = decide(capProfileUpdate, tc.selection(), f, Target{Kind: kindProfile, Organization: tc.organization, Profile: p})
			return nil
		})
	if err != nil {
		return tenancy.BusinessProfile{}, false, err
	}
	return out, canUpdate, nil
}
