package access

import (
	"time"

	"core/internal/tenancy"
)

// The storage-action mapping (G009 Plan r1 §4.2, escalation E1). Reads reuse G007's organization.read and group.read
// unchanged through decide; writes use this separate closed mapping. The eleven G007 capabilities, decide,
// capabilityKind and the G007 matrix are unchanged, and organization.manage_* is not borrowed.
//
//	action   Organization object     Group object (owning Group g)
//	read     organization.read       group.read on g
//	put      org_owner, org_admin    org_owner, org_admin, or any current role in g
//	delete   org_owner, org_admin    org_owner, org_admin, group_owner/group_admin of g; group_member of g only as creator

// The storage reasons (G009 Plan r1 §4.8), beside the G007 reasons; like those they appear only in records.
const (
	reasonNotCreator             Reason = "not_creator"
	reasonObjectNotReady         Reason = "object_not_ready"
	reasonObjectsUnconfigured    Reason = "objects_unconfigured"
	reasonProviderUnavailable    Reason = "provider_unavailable"
	reasonProviderOutcomeUnknown Reason = "provider_outcome_unknown"
	reasonObjectBytesMissing     Reason = "object_bytes_missing"
	reasonBindingMismatch        Reason = "binding_mismatch"
	reasonIntegrityMismatch      Reason = "integrity_mismatch"
	reasonCompensated            Reason = "compensated"
	reasonCompensationFailed     Reason = "compensation_failed"
	reasonReconciledReady        Reason = "reconciled_ready"
	reasonSourceNotReleased      Reason = "source_not_released"
)

// objectAction is one storage action.
type objectAction string

const (
	actionRead   objectAction = "read"
	actionPut    objectAction = "put"
	actionDelete objectAction = "delete"
)

// objectTarget is the immutable scope of a stored object (or of a Put): its owning Group (zero for an Organization
// object) and, for an existing object, its creator.
type objectTarget struct {
	group   tenancy.GroupID
	creator tenancy.UserID
}

// decideObject is the closed, deny-by-default storage evaluator. It is pure. The gates run in G007 order: the current
// Organization binding; the owning Group belongs to the active Organization (a foreign or unknown Group is
// target_not_found); the selected Group restricts Group targets (target_outside_context); then the mapping. Creator
// equality is evaluated only after the role gate and never grants scope. G060 (ADR-008; Spec R3): a Group object is
// content, so every action on it needs an explicit role in that exact Group; an Organization owner or admin without one
// is denied under either setting of the Group-management override, and Group metadata visibility confers nothing here.
// Organization objects keep their Organization rules.
func decideObject(a objectAction, sel contextSelection, f Facts, t objectTarget, subject tenancy.UserID) (bool, Reason) {
	if a == actionRead && zero(t.group) {
		return decide(capOrganizationRead, sel, f, Target{Kind: kindOrganization, Organization: sel.Organization})
	}
	orgRole, r := organizationRole(sel, f)
	if r != "" {
		return false, r
	}
	if !zero(t.group) {
		if !hasGroup(f, sel.Organization, t.group) {
			return false, reasonTargetNotFound
		}
		if !zero(sel.Group) && t.group != sel.Group {
			return false, reasonTargetOutsideContext
		}
	}
	roles, ok := groupRoles(sel, f)
	if !ok {
		return false, reasonFactInconsistent
	}
	orgAdmin := orgRole == tenancy.OrgOwner || orgRole == tenancy.OrgAdmin
	switch a {
	case actionRead:
		return allowIf(roles[t.group].Valid())
	case actionPut:
		if zero(t.group) {
			return allowIf(orgAdmin)
		}
		return allowIf(roles[t.group].Valid())
	case actionDelete:
		if zero(t.group) {
			return allowIf(orgAdmin)
		}
		switch gr := roles[t.group]; {
		case gr == tenancy.GroupOwner || gr == tenancy.GroupAdmin:
			return true, reasonAllowed
		case gr == tenancy.GroupMember && !zero(t.creator) && t.creator == subject:
			return true, reasonAllowed
		case gr == tenancy.GroupMember:
			return false, reasonNotCreator
		}
		return false, reasonRoleNotPermitted
	}
	return false, reasonCapabilityUnknown
}

// objectDenial maps a denied storage decision: an unknown, foreign or outside-context target, or one the subject cannot
// read on the same facts (G007's same-facts read rule), is not_found; a readable target whose action is denied is
// forbidden.
func objectDenial(a objectAction, sel contextSelection, f Facts, t objectTarget, subject tenancy.UserID, r Reason) *failure {
	switch r {
	case reasonTargetNotFound, reasonTargetKindMismatch, reasonTargetOutsideContext:
		return fail(CodeNotFound, r)
	}
	if a != actionRead {
		if readable, _ := decideObject(actionRead, sel, f, t, subject); readable {
			return fail(CodeForbidden, r)
		}
	}
	return fail(CodeNotFound, r)
}

// objectFactQuery reads the facts of one storage decision: the owning Group (if any) and the selected Group (if any),
// never every binding of the subject (Groups is never nil), so fact work is constant in the subject's binding count.
func objectFactQuery(tc *TenantContext, owning tenancy.GroupID, lock bool) tenancy.FactQuery {
	groups := []tenancy.GroupID{}
	if !zero(owning) {
		groups = append(groups, owning)
	}
	if !zero(tc.group) && tc.group != owning {
		groups = append(groups, tc.group)
	}
	return tenancy.FactQuery{Organization: tc.organization, User: tc.subject, Groups: groups, Lock: lock}
}

// ObjectConfig holds the storage bounds of NewServiceWithObjects (G009 Plan r1 §4.3, §4.7), from config.Load's
// OBJECT_STORE_* settings. A value outside its bounds is refused.
type ObjectConfig struct {
	MaxObjectBytes  int64         // 1..1073741824
	DefaultTTL      time.Duration // whole seconds, 1 s..MaxTTL
	MaxTTL          time.Duration // whole seconds, 1..300 s
	TransferTimeout time.Duration // 1..900 s
}

// The hard caps and fixed budgets of the storage operations (G009 Plan r1 §4.3, §4.3a).
const (
	hardMaxTTL             = 300 * time.Second
	hardMaxTransfer        = 900 * time.Second
	storageDecisionBudget  = 5 * time.Second  // each storage transaction, BEGIN through COMMIT
	storageCleanupBudget   = 5 * time.Second  // each storage transaction's deferred rollback
	providerControlTimeout = 10 * time.Second // Head and Delete
	sourceReapGrace        = 5 * time.Second  // after Close, wait for an in-flight Read of the upload source
)

func (c ObjectConfig) valid() bool {
	whole := func(d time.Duration) bool { return d%time.Second == 0 }
	return c.MaxObjectBytes >= 1 && c.MaxObjectBytes <= tenancy.MaxObjectBytes &&
		c.MaxTTL >= time.Second && c.MaxTTL <= hardMaxTTL && whole(c.MaxTTL) &&
		c.DefaultTTL >= time.Second && c.DefaultTTL <= c.MaxTTL && whole(c.DefaultTTL) &&
		c.TransferTimeout >= time.Second && c.TransferTimeout <= hardMaxTransfer
}

// signedTTL validates a signed-URL TTL request: exactly one form, the default (DefaultTTL with TTLSeconds 0) or an
// explicit 1..MaxTTL seconds. It is checked before any int-to-duration conversion, so it cannot overflow.
func (c ObjectConfig) signedTTL(req SignedGetURLRequest) (time.Duration, bool) {
	switch {
	case req.DefaultTTL && req.TTLSeconds == 0:
		return c.DefaultTTL, true
	case !req.DefaultTTL && req.TTLSeconds >= 1 && req.TTLSeconds <= int64(c.MaxTTL/time.Second):
		return time.Duration(req.TTLSeconds) * time.Second, true
	}
	return 0, false
}

// inputFailureCode is invalid_input with a tenancy input code as its Field.
func inputFailureCode(c tenancy.Code) *failure {
	return &failure{code: CodeInvalidInput, field: string(c), reason: reasonInputInvalid}
}
