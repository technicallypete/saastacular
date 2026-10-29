package access

import (
	"context"

	"core/internal/tenancy"
)

// The G020 guarded label operations (Plan r2 §1.3, §4.1). Every operation is one guard or guardEvaluated call: the
// binding check and every input refusal before any transaction, then, in one transaction, the session, the ordered facts
// (users, session, Organization, own bindings, Group rows), the decision(s), and only then the label rows (a definition
// FOR NO KEY UPDATE for update and delete, FOR SHARE for assign and unassign), so a label row is never locked before the
// Group rows. Audit rows are written by the capture triggers of 00020, never here; no override path exists, so
// access.override_used is never appended. No name, target id or payload reaches a record.

// LabelTargetSelector is the caller's untrusted target selector (wire text): Group (a Group id) for the production
// group kind, or Kind and ID for another registered kind (tests only); never both (target_ambiguous).
type LabelTargetSelector struct {
	Group string
	Kind  string
	ID    string
}

// LabelDefinitionView is one label definition as the facade answers it. Group is nil for an Organization-owned
// definition and the owning Group for a Group-owned one. Never carries the name key or created_at.
type LabelDefinitionView struct {
	ID         tenancy.LabelDefinitionID
	Scope      tenancy.ScopeType // tenancy.ScopeOrganization | tenancy.ScopeGroup
	Group      *tenancy.GroupID
	Name       string
	Visibility tenancy.LabelVisibility // "owner_scope" | "target_members"
}

// LabelDefinitionPage is one keyset page (by definition id); Next is the id cursor of the following page, nil on the last.
type LabelDefinitionPage struct {
	Labels []LabelDefinitionView
	Next   *tenancy.LabelDefinitionID
}

// LabelAssignmentView is one label assigned to a target: the definition id and its name.
type LabelAssignmentView struct {
	LabelID tenancy.LabelDefinitionID
	Name    string
}

// LabelDeleteResult is a completed definition deletion: the number of assignments removed with it.
type LabelDeleteResult struct {
	AssignmentsRemoved int64
}

// ListLabelDefinitions: label.definition.list. sc.Kind organization -> label.organization.manage; group (sc.Group set)
// -> label.group.manage, falling back to label.group.read for the target_members subset. after "" = first page, else a
// canonical definition id (invalid_cursor); limit 0 = tenancy.DefaultLabelPage, else 1..tenancy.MaxLabelPage
// (invalid_page_size). Read-only REPEATABLE READ.
func (s *Service) ListLabelDefinitions(ctx context.Context, tc *TenantContext, sc Scope, after string, limit int) (LabelDefinitionPage, error) {
	var out LabelDefinitionPage
	var from tenancy.LabelDefinitionID
	pre := labelScopePre(sc, func() *failure {
		var f *failure
		from, limit, f = labelPageInput(after, limit)
		return f
	})
	page := func(tx *trackedTx, f Facts, visibility tenancy.LabelVisibility) *failure {
		rows, err := tenancy.LabelDefinitionsPage(ctx, tx, f.Organization.ID, sc.group(), visibility, from, limit)
		if err != nil {
			return labelFailure(ctx, tx, err)
		}
		out = labelDefinitionPage(rows, limit)
		return nil
	}
	if sc.Kind != tenancy.ScopeGroup {
		err := s.guard(ctx, tc, opLabelDefinitionList, capLabelOrganizationManage, false, Target{Kind: kindOrganization}, organizationQuery, pre,
			func(tx *trackedTx, f Facts) *failure { return page(tx, f, "") })
		if err != nil {
			return LabelDefinitionPage{}, err
		}
		return out, nil
	}
	g := sc.group()
	err := s.guardEvaluated(ctx, tc, opLabelDefinitionList, false, Target{Kind: kindGroup, Group: g}, groupQuery(g, false), pre,
		func(tx *trackedTx, f Facts, evaluated *evaluation) *failure {
			sel := tc.selection()
			t := Target{Kind: kindGroup, Organization: f.Organization.ID, Group: g}
			// The managers' decision first; a denial falls back to the readers' decision on the same facts, which admits
			// only the target_members subset. The record carries the capability that allowed (or the last that denied).
			if ok, r := decide(capLabelGroupManage, sel, f, t); ok {
				evaluated.capability, evaluated.reason, evaluated.facts = capLabelGroupManage, r, f
				return page(tx, f, "")
			}
			if fl := labelDecide(sel, f, capLabelGroupRead, t, evaluated); fl != nil {
				return fl
			}
			return page(tx, f, tenancy.LabelVisibilityTargetMembers)
		})
	if err != nil {
		return LabelDefinitionPage{}, err
	}
	return out, nil
}

// CreateLabelDefinition: label.definition.create. visibility is "owner_scope" or "target_members" (invalid_visibility);
// target_members in the Organization scope is invalid_input visibility_not_allowed before SQL; name through
// tenancy.ValidateLabelName (invalid_name). Organization scope: policyWriteQuery; Group scope: groupQuery(g, true).
func (s *Service) CreateLabelDefinition(ctx context.Context, tc *TenantContext, sc Scope, name, visibility string) (LabelDefinitionView, error) {
	var out LabelDefinitionView
	var stored string
	var vis tenancy.LabelVisibility
	pre := labelScopePre(sc, func() *failure {
		var f *failure
		stored, vis, f = labelDefinitionInput(sc.Kind, name, visibility)
		return f
	})
	c, t, query := capLabelOrganizationManage, Target{Kind: kindOrganization}, policyWriteQuery
	if sc.Kind == tenancy.ScopeGroup {
		c, t, query = capLabelGroupManage, Target{Kind: kindGroup, Group: sc.group()}, groupQuery(sc.group(), true)
	}
	err := s.guard(ctx, tc, opLabelDefinitionCreate, c, true, t, query, pre,
		func(tx *trackedTx, f Facts) *failure {
			id, ok := newKey[tenancy.LabelDefinitionID]()
			if !ok {
				return fail(CodeInternal, reasonStoreUnavailable)
			}
			d, err := tenancy.InsertLabelDefinitionTx(ctx, tx, tenancy.LabelDefinitionSpec{ID: id, OrganizationID: f.Organization.ID,
				Group: sc.group(), Name: stored, Visibility: vis})
			if err != nil {
				return labelFailure(ctx, tx, err)
			}
			out = labelDefinitionView(d)
			return nil
		})
	if err != nil {
		return LabelDefinitionView{}, err
	}
	return out, nil
}

// UpdateLabelDefinition: label.definition.update (guardEvaluated: row read first, then manage of its owner scope); an
// absolute set of name and visibility; id is a canonical definition id (invalid_id).
func (s *Service) UpdateLabelDefinition(ctx context.Context, tc *TenantContext, id, name, visibility string) (LabelDefinitionView, error) {
	var out LabelDefinitionView
	var label tenancy.LabelDefinitionID
	var stored string
	var vis tenancy.LabelVisibility
	pre := func() *failure {
		var f *failure
		if label, f = labelID(id); f != nil {
			return f
		}
		// The owner scope is unknown before the row is read: the Organization refusal of target_members follows the read.
		stored, vis, f = labelDefinitionInput(tenancy.ScopeGroup, name, visibility)
		return f
	}
	err := s.guardEvaluated(ctx, tc, opLabelDefinitionUpdate, true, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts, evaluated *evaluation) *failure {
			d, fl := s.labelOwnerManage(ctx, tx, tc, f, label, evaluated)
			if fl != nil {
				return fl
			}
			if d.OwnerScopeType == tenancy.ScopeOrganization && vis != tenancy.LabelVisibilityOwnerScope {
				return labelInputFailure(fieldVisibilityNotAllowed)
			}
			if _, err := tenancy.LockLabelDefinitionTx(ctx, tx, f.Organization.ID, label, true); err != nil {
				return labelFailure(ctx, tx, err)
			}
			// Setting the current values succeeds without effect; the capture trigger's WHEN clauses record nothing.
			u, err := tenancy.UpdateLabelDefinitionTx(ctx, tx, f.Organization.ID, label, stored, vis)
			if err != nil {
				return labelFailure(ctx, tx, err)
			}
			out = labelDefinitionView(u)
			return nil
		})
	if err != nil {
		return LabelDefinitionView{}, err
	}
	return out, nil
}

// DeleteLabelDefinition: label.definition.delete; an Organization-owned definition with a selected Group in the trusted
// context is invalid_input organization_context_required (before any write).
func (s *Service) DeleteLabelDefinition(ctx context.Context, tc *TenantContext, id string) (LabelDeleteResult, error) {
	var out LabelDeleteResult
	var label tenancy.LabelDefinitionID
	pre := func() *failure {
		var f *failure
		label, f = labelID(id)
		return f
	}
	err := s.guardEvaluated(ctx, tc, opLabelDefinitionDelete, true, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts, evaluated *evaluation) *failure {
			d, fl := s.labelOwnerManage(ctx, tx, tc, f, label, evaluated)
			if fl != nil {
				return fl
			}
			// r1 finding 3: the selected-Group filter of the assignments DELETE policy would leave other Groups'
			// assignments behind, so an Organization-owned definition is deleted only in the Organization-wide context.
			// The refusal is uniform (independent of any assignment) and precedes every lock and write.
			if d.OwnerScopeType == tenancy.ScopeOrganization && !zero(tc.group) {
				return labelInputFailure(fieldOrganizationContextRequired)
			}
			if _, err := tenancy.LockLabelDefinitionTx(ctx, tx, f.Organization.ID, label, true); err != nil {
				return labelFailure(ctx, tx, err)
			}
			n, err := tenancy.DeleteLabelDefinitionAssignmentsTx(ctx, tx, f.Organization.ID, label)
			if err != nil {
				return labelFailure(ctx, tx, err)
			}
			if err := tenancy.DeleteLabelDefinitionTx(ctx, tx, f.Organization.ID, label); err != nil {
				return labelFailure(ctx, tx, err)
			}
			out.AssignmentsRemoved = n
			return nil
		})
	if err != nil {
		return LabelDeleteResult{}, err
	}
	return out, nil
}

// ListLabelAssignments: label.assignment.list for one target. group target: label.organization.manage and group.read on
// the exact Group; a registered Group-owned kind: label.group.read/manage on the owning Group plus the adapter's
// readable. Ordered by definition id.
func (s *Service) ListLabelAssignments(ctx context.Context, tc *TenantContext, t LabelTargetSelector) ([]LabelAssignmentView, error) {
	out := []LabelAssignmentView{}
	var target labelTarget
	var adapter labelTargetAdapter
	pre := func() *failure {
		var f *failure
		target, adapter, f = labelTargetInput(t)
		return f
	}
	err := s.guardEvaluated(ctx, tc, opLabelAssignmentList, false, Target{Kind: kindOrganization}, labelTargetQuery(t, false), pre,
		func(tx *trackedTx, f Facts, evaluated *evaluation) *failure {
			sel := tc.selection()
			owner, owning := tenancy.ScopeOrganization, tenancy.GroupID{}
			if target.Kind == labelTargetGroup {
				g, fl := s.labelGroupTarget(ctx, tx, sel, f, target, adapter, tenancy.LabelDefinitionID{}, evaluated)
				if fl != nil {
					return fl
				}
				owning = g
			} else {
				// A registered Group-owned kind: resolve the owning Group in this transaction, read its facts, then the
				// managers' or readers' decision on it and the adapter's current read authority on the target.
				var fl *failure
				if owning, fl = labelResolve(ctx, tx, f, target, adapter); fl != nil {
					return fl
				}
				if f, fl = labelGroupFacts(ctx, tx, tc, f, owning, false); fl != nil {
					return fl
				}
				gt := Target{Kind: kindGroup, Organization: f.Organization.ID, Group: owning}
				if ok, r := decide(capLabelGroupManage, sel, f, gt); ok {
					evaluated.capability, evaluated.reason, evaluated.facts = capLabelGroupManage, r, f
				} else if fl := labelDecide(sel, f, capLabelGroupRead, gt, evaluated); fl != nil {
					return fl
				}
				if !adapter.readable(sel, f, owning) {
					return markDenied(notFound(), sel, f, "", [16]byte{}, tenancy.GroupID{})
				}
				owner = tenancy.ScopeGroup
			}
			rows, err := tenancy.LabelAssignmentsByTarget(ctx, tx, f.Organization.ID, string(target.Kind), target.ID)
			if err != nil {
				return labelFailure(ctx, tx, err)
			}
			// Row-level security admits only readers of each definition (a Group member: target_members only); the owner
			// facts are re-checked here so a row of another owner scope never answers for this target.
			for _, a := range rows {
				if labelAssignmentOwned(a, owner, owning) {
					out = append(out, LabelAssignmentView{LabelID: a.LabelID, Name: a.LabelName})
				}
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AssignLabel: label.assign of definition id to the target; duplicate is conflict already_assigned.
func (s *Service) AssignLabel(ctx context.Context, tc *TenantContext, id string, t LabelTargetSelector) error {
	return s.labelAssignment(ctx, tc, opLabelAssign, id, t, true)
}

// UnassignLabel: label.unassign; an absent assignment is conflict not_assigned.
func (s *Service) UnassignLabel(ctx context.Context, tc *TenantContext, id string, t LabelTargetSelector) error {
	return s.labelAssignment(ctx, tc, opLabelUnassign, id, t, false)
}

// labelAssignment is assign (insert) or unassign: the definition is read first (invisible: not_found) and admitted for
// the target kind (target_unsupported); a group target needs label.organization.manage and group.read on the exact Group,
// whose row the facts locked FOR NO KEY UPDATE; a registered Group-owned kind resolves its owning Group in this
// transaction, which must be the definition's (else not_found), then label.group.manage on it. The definition is locked
// FOR SHARE after the Group rows, then the one INSERT or DELETE runs.
func (s *Service) labelAssignment(ctx context.Context, tc *TenantContext, op, id string, t LabelTargetSelector, insert bool) error {
	var label tenancy.LabelDefinitionID
	var target labelTarget
	pre := func() *failure {
		var f *failure
		if label, f = labelID(id); f != nil {
			return f
		}
		target, _, f = labelTargetInput(t)
		return f
	}
	return s.guardEvaluated(ctx, tc, op, true, Target{Kind: kindOrganization}, labelTargetQuery(t, true), pre,
		func(tx *trackedTx, f Facts, evaluated *evaluation) *failure {
			sel := tc.selection()
			d, err := tenancy.LabelDefinitionTx(ctx, tx, f.Organization.ID, label)
			if err != nil {
				return labelFailure(ctx, tx, err)
			}
			adapter, fl := admitLabelTarget(d.OwnerScopeType, target)
			if fl != nil {
				return fl
			}
			var owning tenancy.GroupID
			if d.OwnerScopeType == tenancy.ScopeOrganization {
				if owning, fl = s.labelGroupTarget(ctx, tx, sel, f, target, adapter, d.ID, evaluated); fl != nil {
					return fl
				}
			} else {
				if f, fl = labelGroupFacts(ctx, tx, tc, f, d.GroupID, true); fl != nil {
					return fl
				}
				if owning, fl = labelResolve(ctx, tx, f, target, adapter); fl != nil {
					return fl
				}
				if owning != d.GroupID {
					return notFound()
				}
				gt := Target{Kind: kindGroup, Organization: f.Organization.ID, Group: owning, Label: d.ID}
				if fl := labelDecide(sel, f, capLabelGroupManage, gt, evaluated); fl != nil {
					return fl
				}
			}
			if d, err = tenancy.LockLabelDefinitionTx(ctx, tx, f.Organization.ID, label, false); err != nil {
				return labelFailure(ctx, tx, err)
			}
			if insert {
				aid, ok := newKey[tenancy.LabelAssignmentID]()
				if !ok {
					return fail(CodeInternal, reasonStoreUnavailable)
				}
				err = tenancy.InsertLabelAssignmentTx(ctx, tx, tenancy.LabelAssignmentSpec{ID: aid, Definition: d,
					TargetType: string(target.Kind), TargetID: target.ID, TargetGroup: owning})
			} else {
				err = tenancy.DeleteLabelAssignmentTx(ctx, tx, f.Organization.ID, label, string(target.Kind), target.ID)
			}
			if err != nil {
				return labelFailure(ctx, tx, err)
			}
			return nil
		})
}

// labelOwnerManage reads definition label (invisible: not_found) and decides the manage capability of its owner scope:
// label.organization.manage for an Organization-owned definition on the operation's facts; label.group.manage for a
// Group-owned one on that Group's facts, re-read in the same transaction (FOR SHARE on a write, after the Organization and
// own-binding locks the facts already hold; never FOR NO KEY UPDATE, so a selected owning Group's shared lock is not
// upgraded), as invitationAdmin does. The definition row itself is not locked here.
func (s *Service) labelOwnerManage(ctx context.Context, tx *trackedTx, tc *TenantContext, f Facts, label tenancy.LabelDefinitionID,
	evaluated *evaluation) (tenancy.LabelDefinition, *failure) {
	d, err := tenancy.LabelDefinitionTx(ctx, tx, f.Organization.ID, label)
	if err != nil {
		return tenancy.LabelDefinition{}, labelFailure(ctx, tx, err)
	}
	c, t := capLabelOrganizationManage, Target{Kind: kindOrganization, Organization: f.Organization.ID, Label: d.ID}
	if d.OwnerScopeType == tenancy.ScopeGroup {
		var fl *failure
		if f, fl = labelGroupFacts(ctx, tx, tc, f, d.GroupID, true); fl != nil {
			return tenancy.LabelDefinition{}, fl
		}
		c, t = capLabelGroupManage, Target{Kind: kindGroup, Organization: f.Organization.ID, Group: d.GroupID, Label: d.ID}
	}
	if fl := labelDecide(tc.selection(), f, c, t, evaluated); fl != nil {
		return tenancy.LabelDefinition{}, fl
	}
	return d, nil
}

// labelGroupTarget decides an Organization-owned label operation on a group target: label.organization.manage on the
// Organization (the definition label, when known, names the audit resource), then group.read on the exact target Group
// resolved from the facts (absent: not_found). It returns the target Group; the record reports
// label.organization.manage, or group.read when that second decision denies.
func (s *Service) labelGroupTarget(ctx context.Context, tx *trackedTx, sel contextSelection, f Facts, target labelTarget,
	adapter labelTargetAdapter, label tenancy.LabelDefinitionID, evaluated *evaluation) (tenancy.GroupID, *failure) {
	ot := Target{Kind: kindOrganization, Organization: f.Organization.ID, Label: label}
	if fl := labelDecide(sel, f, capLabelOrganizationManage, ot, evaluated); fl != nil {
		return tenancy.GroupID{}, fl
	}
	allowed := *evaluated
	g, fl := labelResolve(ctx, tx, f, target, adapter)
	if fl != nil {
		return tenancy.GroupID{}, fl
	}
	if fl := labelDecide(sel, f, capGroupRead, Target{Kind: kindGroup, Organization: f.Organization.ID, Group: g}, evaluated); fl != nil {
		return tenancy.GroupID{}, fl
	}
	*evaluated = allowed
	return g, nil
}

// labelResolve resolves the target through its adapter inside the operation's transaction: its owning Group, or
// not_found when it does not exist in the active Organization.
func labelResolve(ctx context.Context, tx *trackedTx, f Facts, target labelTarget, adapter labelTargetAdapter) (tenancy.GroupID, *failure) {
	g, ok, err := adapter.resolve(ctx, tx, f, f.Organization.ID, target.ID)
	if err != nil {
		return tenancy.GroupID{}, labelFailure(ctx, tx, err)
	}
	if !ok || zero(g) {
		return tenancy.GroupID{}, notFound()
	}
	return g, nil
}

// labelDecide decides c on t with facts f, recording c as the evaluated capability; a denial is the ordinary
// not_found/forbidden mapping, marked for the access.denied record.
func labelDecide(sel contextSelection, f Facts, c Capability, t Target, evaluated *evaluation) *failure {
	evaluated.capability = c
	ok, r := decide(c, sel, f, t)
	if !ok {
		res, id, g := targetScope(t)
		return markDenied(denial(c, sel, f, t, r), sel, f, res, id, g)
	}
	evaluated.reason, evaluated.facts = r, f
	return nil
}

// labelGroupFacts re-reads the facts of the Organization, the subject's bindings and Group g (plus the selected Group)
// in the operation's transaction, locked for a write (the invitationAdmin pattern), so a Group-owned definition's owning
// Group is decided on current facts before any label row is locked.
func labelGroupFacts(ctx context.Context, tx *trackedTx, tc *TenantContext, f Facts, g tenancy.GroupID, write bool) (Facts, *failure) {
	groups := []tenancy.GroupID{g}
	if !zero(tc.group) && tc.group != g {
		groups = append(groups, tc.group)
	}
	nf, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: f.Organization.ID, User: tc.subject, Groups: groups, Lock: write})
	if err != nil {
		return Facts{}, fail(CodeUnavailable, failureReason(ctx, tx.cause))
	}
	return nf, nil
}

// labelTargetQuery is the fact query of an assignment operation. A group target (either selector form) reads the
// target Group through groupQuery (locked FOR NO KEY UPDATE for a write); every other kind reads the Organization facts
// (its owning Group is resolved and read in the transaction). An ambiguous or malformed selector reads the Organization
// facts only; the input check refuses it before any transaction.
func labelTargetQuery(t LabelTargetSelector, write bool) func(*TenantContext) tenancy.FactQuery {
	group := t.Group
	if group == "" && t.Kind == string(labelTargetGroup) {
		group = t.ID
	} else if t.Kind != "" || t.ID != "" {
		return organizationQuery
	}
	g, err := tenancy.ParseGroupID(group)
	if err != nil || zero(g) {
		return organizationQuery
	}
	return groupQuery(g, write)
}

// labelAssignmentOwned reports whether an assignment row belongs to the owner scope the operation decided: an
// Organization-owned definition for a group target; the exact owning Group's definition for a Group-owned kind.
func labelAssignmentOwned(a tenancy.LabelAssignment, owner tenancy.ScopeType, owning tenancy.GroupID) bool {
	switch owner {
	case tenancy.ScopeOrganization:
		return a.OwnerScopeType == tenancy.ScopeOrganization && a.OwnerScopeID == [16]byte(a.OrganizationID)
	case tenancy.ScopeGroup:
		return a.OwnerScopeType == tenancy.ScopeGroup && a.OwnerScopeID == [16]byte(owning) && a.TargetGroupID == owning
	}
	return false
}

// labelScopePre validates the scope (invalid_scope), then runs more.
func labelScopePre(sc Scope, more func() *failure) func() *failure {
	return func() *failure {
		if !sc.valid() {
			return labelInputFailure(string(tenancy.CodeInvalidScope))
		}
		return more()
	}
}

// labelID parses a definition id: a canonical nonzero UUID, otherwise invalid_id.
func labelID(s string) (tenancy.LabelDefinitionID, *failure) {
	id, err := tenancy.ParseLabelDefinitionID(s)
	if err != nil || zero(id) {
		return tenancy.LabelDefinitionID{}, labelInputFailure(string(tenancy.CodeInvalidID))
	}
	return id, nil
}

// labelPageInput validates a definition page request: limit 0 is DefaultLabelPage, otherwise 1..MaxLabelPage
// (invalid_page_size); after is "" or a canonical nonzero definition id (invalid_cursor).
func labelPageInput(after string, limit int) (tenancy.LabelDefinitionID, int, *failure) {
	switch {
	case limit == 0:
		limit = tenancy.DefaultLabelPage
	case limit < 1 || limit > tenancy.MaxLabelPage:
		return tenancy.LabelDefinitionID{}, 0, labelInputFailure(fieldInvalidPageSize)
	}
	if after == "" {
		return tenancy.LabelDefinitionID{}, limit, nil
	}
	id, err := tenancy.ParseLabelDefinitionID(after)
	if err != nil || zero(id) {
		return tenancy.LabelDefinitionID{}, 0, labelInputFailure(fieldInvalidCursor)
	}
	return id, limit, nil
}

// The page refusals of the label list (the G014 keyset vocabulary of pageInput).
const (
	fieldInvalidPageSize = "invalid_page_size"
	fieldInvalidCursor   = "invalid_cursor"
)

// labelDefinitionInput validates a definition's visibility (invalid_visibility; target_members in the Organization scope
// is visibility_not_allowed) and name (tenancy.ValidateLabelName: invalid_name), returning the stored name.
func labelDefinitionInput(scope tenancy.ScopeType, name, visibility string) (string, tenancy.LabelVisibility, *failure) {
	vis, err := tenancy.ParseLabelVisibility(visibility)
	if err != nil {
		return "", "", inputFailure(err, reasonInputInvalid)
	}
	if scope == tenancy.ScopeOrganization && vis != tenancy.LabelVisibilityOwnerScope {
		return "", "", labelInputFailure(fieldVisibilityNotAllowed)
	}
	stored, err := tenancy.ValidateLabelName(name)
	if err != nil {
		return "", "", inputFailure(err, reasonInputInvalid)
	}
	return stored, vis, nil
}

// labelTargetInput parses the caller's selector (target_ambiguous, target_unsupported, invalid_id) and, before any SQL,
// checks that some owner scope may target its kind: the group kind admits Organization-owned definitions, every other
// registered kind Group-owned ones. It returns the parsed target and its adapter.
func labelTargetInput(s LabelTargetSelector) (labelTarget, labelTargetAdapter, *failure) {
	t, f := parseLabelTarget(labelSelector{group: s.Group, kind: s.Kind, id: s.ID})
	if f != nil {
		return labelTarget{}, nil, f
	}
	owner := tenancy.ScopeGroup
	if t.Kind == labelTargetGroup {
		owner = tenancy.ScopeOrganization
	}
	a, f := admitLabelTarget(owner, t)
	if f != nil {
		return labelTarget{}, nil, f
	}
	return t, a, nil
}

// labelDefinitionView is the facade's answer for a definition row.
func labelDefinitionView(d tenancy.LabelDefinition) LabelDefinitionView {
	v := LabelDefinitionView{ID: d.ID, Scope: d.OwnerScopeType, Name: d.Name, Visibility: d.Visibility}
	if d.OwnerScopeType == tenancy.ScopeGroup {
		g := d.GroupID
		v.Group = &g
	}
	return v
}

// labelDefinitionPage turns a limit+1 read into a page: a row beyond limit means a next page, whose cursor is the last
// returned id.
func labelDefinitionPage(rows []tenancy.LabelDefinition, limit int) LabelDefinitionPage {
	out := LabelDefinitionPage{Labels: []LabelDefinitionView{}}
	if len(rows) > limit {
		rows = rows[:limit]
		next := rows[limit-1].ID
		out.Next = &next
	}
	for _, d := range rows {
		out.Labels = append(out.Labels, labelDefinitionView(d))
	}
	return out
}
