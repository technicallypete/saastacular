package access

import (
	"context"
	"time"

	"core/internal/audit"
	"core/internal/tenancy"
)

// The G011 durable audit of the facade (docs/foundation/AUDIT.md; G011 Plan §3, §4). Row changes of guarded writes and
// storage transactions are recorded by the database capture triggers inside their own transactions; this file records
// the events that are not row changes, each in a short independent audit transaction after the business transaction
// ended and released its connection: security denials (access.denied), signed-URL issuance (object.url_issued, before
// the URL is returned) and unconfirmed storage outcomes (object.upload_failed, object.deletion_pending). It also holds
// the one guarded audit read, ReadAuditEvents.

// The bounds of an independent audit record (G011 Plan §4.3): BEGIN through COMMIT, and the deferred rollback.
const (
	auditRecordBudget  = 2 * time.Second
	auditCleanupBudget = 2 * time.Second
)

// denialScope is what a recordable denial established on the denied transaction's facts: whether the Organization gate
// passed, and the target (only for forbidden: a target the subject can read on the same facts, so a rejected foreign
// selector never becomes a recorded UUID).
type denialScope struct {
	org      bool
	resource audit.ResourceType
	id       audit.UUID
	group    tenancy.GroupID
}

// recordableReasons are the closed access.denied reasons.
var recordableReasons = map[Reason]bool{
	reasonOrganizationNotPermitted: true, reasonGroupNotSelectable: true, reasonTargetNotFound: true,
	reasonTargetKindMismatch: true, reasonTargetOutsideContext: true, reasonRoleNotPermitted: true, reasonNoLinkedPath: true,
	reasonEditPolicyDisabled: true, reasonNotCreator: true, reasonObjectNotReady: true, reasonCapabilityUnknown: true,
	reasonFactInconsistent: true,
}

// markDenied marks f as a recordable denial when it is a policy refusal after a valid binding check: context_denied with
// organization_not_permitted or group_not_selectable, or not_found or forbidden with a closed denial reason. The
// Organization is established when the facts pass the Organization gate; the target (resource, id, Group) is kept only
// for forbidden.
func markDenied(f *failure, sel contextSelection, facts Facts, resource audit.ResourceType, id audit.UUID, group tenancy.GroupID) *failure {
	if f == nil || !recordableReasons[f.reason] {
		return f
	}
	switch f.code {
	case CodeContextDenied:
		if f.reason != reasonOrganizationNotPermitted && f.reason != reasonGroupNotSelectable {
			return f
		}
	case CodeNotFound, CodeForbidden:
	default:
		return f
	}
	_, r := organizationRole(sel, facts)
	d := &denialScope{org: r == ""}
	if d.org && f.code == CodeForbidden && resource != "" {
		d.resource, d.id, d.group = resource, id, group
	}
	f.denied = d
	return f
}

// targetScope is the audit resource of a guarded operation's target. G020: a label operation's target (label.* capability,
// Organization or Group kind, with the definition in Target.Label) is the resource label with the definition id and,
// for a Group-owned definition, its owning Group; an Organization-owned definition carries no Group.
func targetScope(t Target) (audit.ResourceType, audit.UUID, tenancy.GroupID) {
	if !zero(t.Label) {
		switch t.Kind {
		case kindOrganization:
			return audit.ResourceLabel, audit.UUID(t.Label), tenancy.GroupID{}
		case kindGroup:
			return audit.ResourceLabel, audit.UUID(t.Label), t.Group
		}
		return "", audit.UUID{}, tenancy.GroupID{}
	}
	switch t.Kind {
	case kindOrganization:
		return audit.ResourceOrganization, audit.UUID(t.Organization), tenancy.GroupID{}
	case kindGroup:
		return audit.ResourceGroup, audit.UUID(t.Group), t.Group
	case kindProfile:
		return audit.ResourceBusinessProfile, audit.UUID(t.Profile), tenancy.GroupID{}
	}
	return "", audit.UUID{}, tenancy.GroupID{}
}

// auditTx is the facade's independent recorder: a short write transaction through inContextTx (so the context
// statement, with the trusted subject and correlation, is first), on a context detached from the caller's cancellation
// and bounded by auditRecordBudget, with the deferred rollback bounded by auditCleanupBudget (the storage-budget
// mechanics of G009). A zero Organization selects the IDENTITY shape; an Organization record uses the TENANT-ORG shape
// without a selected Group. It inherits inTx's COMMIT classification, prior-context discard and bounded cleanup. No
// retry.
func (s *Service) auditTx(ctx context.Context, dbc dbContext, events ...audit.Event) *failure {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), auditRecordBudget)
	defer cancel()
	return s.inContextTx(ctx, true, dbc, auditCleanupBudget, func(tx *trackedTx) *failure {
		for _, e := range events {
			if err := audit.Append(ctx, tx, e); err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
		}
		return nil
	})
}

// record runs auditTx and writes one audit.record_failed diagnostic per event on failure; the caller's outcome is
// unchanged.
func (s *Service) record(ctx context.Context, dbc dbContext, events ...audit.Event) bool {
	if f := s.auditTx(ctx, dbc, events...); f != nil {
		for _, e := range events {
			s.logAuditRecordFailed(ctx, e.Action)
		}
		return false
	}
	return true
}

// recordDenial records one access.denied event for a marked denial of op by subject in org (the context's
// Organization), after the denied transaction released its connection and before the operational record. The
// Organization (and the target) are recorded only when established; otherwise the event is global (IDENTITY shape).
func (s *Service) recordDenial(ctx context.Context, subject tenancy.UserID, org tenancy.OrganizationID, op string, f *failure) {
	if f == nil || f.denied == nil {
		return
	}
	e := audit.Event{Action: audit.AccessDenied, Outcome: audit.Denied, Reason: audit.Reason(f.reason), Operation: audit.Operation(op)}
	dbc := dbContext{user: subject}
	if f.denied.org {
		e.Organization, dbc.organization = audit.UUID(org), org
		if f.denied.resource != "" {
			e.ResourceType, e.ResourceID, e.Group = f.denied.resource, f.denied.id, audit.UUID(f.denied.group)
		}
	}
	s.record(ctx, dbc, e)
}

// recordGuarded is recordDenial for a guarded or storage call of tc (nil-safe: an unbound call is never marked).
func (s *Service) recordGuarded(ctx context.Context, tc *TenantContext, op string, f *failure) {
	if f == nil || f.denied == nil || tc == nil {
		return
	}
	s.recordDenial(ctx, tc.subject, tc.organization, op, f)
}

// objectEvent is a storage outcome event of object id (Group g) in tc's Organization; withOrg false records it without
// the Organization (and Group) when the scope is no longer established (a later policy denial).
func objectEvent(tc *TenantContext, a audit.Action, o audit.Outcome, r audit.Reason, id tenancy.ObjectID, g tenancy.GroupID, withOrg bool) (dbContext, audit.Event) {
	e := audit.Event{Action: a, Outcome: o, Reason: r, ResourceType: audit.ResourceStoredObject, ResourceID: audit.UUID(id)}
	dbc := dbContext{user: tc.subject}
	if withOrg {
		e.Organization, e.Group, dbc.organization = audit.UUID(tc.organization), audit.UUID(g), tc.organization
	}
	return dbc, e
}

// recordObject records one storage outcome event (G4, G5) independently; a failure only writes the diagnostic.
func (s *Service) recordObject(ctx context.Context, tc *TenantContext, a audit.Action, o audit.Outcome, r audit.Reason, id tenancy.ObjectID,
	g tenancy.GroupID, withOrg bool) {
	dbc, e := objectEvent(tc, a, o, r, id, g, withOrg)
	s.record(ctx, dbc, e)
}

// compensation is the object.upload_failed reason of a compensation result.
func compensation(ok bool) audit.Reason {
	if ok {
		return audit.ReasonCompensated
	}
	return audit.ReasonCompensationFailed
}

// --- the guarded audit read --------------------------------------------------------------------------------------------

// The page bounds of ReadAuditEvents.
const (
	defaultAuditPage = 50
	maxAuditPage     = 100
)

// AuditReadRequest selects a page of the active Organization's audit events: the time window [Since, Until) (zero: no
// bound), the page size (0: 50; 1..100) and the continuation cursor (zero: the first page).
type AuditReadRequest struct {
	Since  time.Time
	Until  time.Time
	Limit  int
	Cursor AuditCursor
}

// AuditCursor is an opaque continuation cursor, bound to the Organization and the window it was issued for.
type AuditCursor struct {
	set   bool
	org   tenancy.OrganizationID
	since time.Time
	until time.Time
	at    time.Time
	id    audit.UUID
}

// AuditPage is one page of events, newest first. More reports whether Next continues it.
type AuditPage struct {
	Events []AuditEvent
	Next   AuditCursor
	More   bool
}

// AuditEvent is one stored audit event.
type AuditEvent = audit.Record

// ReadAuditEvents reads the active Organization's audit events (audit.read: the current org_owner only), newest first,
// through the guard in one REPEATABLE READ READ ONLY snapshot. Row-level security enforces the same rule. A selected
// Group grants nothing: an owner in selected-Group context reads the whole Organization; Group roles read nothing.
// Organization-less (global) events are never returned. A page is a snapshot: an event committed later with an
// earlier time may be absent from an ongoing traversal and appears in a fresh one. Successful reads are not audited.
func (s *Service) ReadAuditEvents(ctx context.Context, tc *TenantContext, req AuditReadRequest) (AuditPage, error) {
	var page AuditPage
	limit := req.Limit
	pre := func() *failure {
		if limit == 0 {
			limit = defaultAuditPage
		}
		switch {
		case limit < 1 || limit > maxAuditPage:
			return &failure{code: CodeInvalidInput, field: audit.CodeInvalidPageSize, reason: reasonInputInvalid}
		case !req.Since.IsZero() && !req.Until.IsZero() && !req.Until.After(req.Since):
			return &failure{code: CodeInvalidInput, field: audit.CodeInvalidTimeRange, reason: reasonInputInvalid}
		case req.Cursor != (AuditCursor{}) && (!req.Cursor.set || req.Cursor.org != tc.organization ||
			!req.Cursor.since.Equal(req.Since) || !req.Cursor.until.Equal(req.Until)):
			return &failure{code: CodeInvalidInput, field: audit.CodeInvalidCursor, reason: reasonInputInvalid}
		}
		return nil
	}
	err := s.guard(ctx, tc, opAuditRead, capAuditRead, false, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts) *failure {
			before, beforeID := req.Until, audit.UUID{}
			if req.Cursor.set {
				before, beforeID = req.Cursor.at, req.Cursor.id
			}
			recs, err := audit.ReadTenant(ctx, tx, audit.UUID(f.Organization.ID), req.Since, before, beforeID, limit+1)
			if err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
			if len(recs) > limit {
				recs = recs[:limit]
				last := recs[len(recs)-1]
				page.More = true
				page.Next = AuditCursor{set: true, org: tc.organization, since: req.Since, until: req.Until, at: last.OccurredAt, id: last.ID}
			}
			page.Events = recs
			return nil
		})
	if err != nil {
		return AuditPage{}, err
	}
	return page, nil
}
