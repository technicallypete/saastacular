package access

import (
	"context"
	"log/slog"

	"core/internal/audit"
	"core/internal/observability"
	"core/internal/tenancy"
)

// eventDecision is the one catalogued log event of this package (events.json): one record per call, including
// BeginRequest and Resolve, with a bounded operation, capability, outcome and reason plus request/trace correlation
// when present. Never IDs, names, slugs, roles, profile or policy values, tokens, claims or driver text. Correlation is
// diagnostic only: this file is the only one that reads it, and no decision depends on it.
const eventDecision = "access.decision"

// The logged operations: the two context operations and the eight guarded operations.
const (
	opBeginRequest          = "access.begin_request"
	opResolve               = "access.resolve"
	opGetOrganization       = "organization.get"
	opGetOrganizationPolicy = "organization.get_policy"
	opSetOrganizationPolicy = "organization.set_policy"
	opGetGroup              = "group.get"
	opRenameGroup           = "group.rename"
	opChangeGroupSlug       = "group.change_slug"
	opGetBusinessProfile    = "business_profile.get"
	opUpdateBusinessProfile = "business_profile.update"
	opAuditRead             = "audit.read" // G011: the guarded audit read
	// G018: the seven guarded onboarding operations (each also an access.denied operation, 00014).
	opSuggestGroupSlug           = "group.suggest_slug"
	opCreateGroup                = "group.create"
	opListGroups                 = "group.list"
	opGetGroupOnboarding         = "group.onboarding.get"
	opCompleteOnboardingNew      = "group.onboarding.complete_new"
	opCompleteOnboardingExisting = "group.onboarding.complete_existing"
	opListBusinessProfiles       = "business_profile.list"
	// G019: the seventeen membership, ownership, invitation and retirement operations (each also an access.denied
	// operation, 00015; r2 addendum §7 adds membership.self).
	opMembershipList        = "membership.list"
	opMembershipSelf        = "membership.self"
	opMembershipChangeRole  = "membership.change_role"
	opMembershipRemove      = "membership.remove"
	opMembershipLeave       = "membership.leave"
	opOwnershipPromote      = "ownership.promote"
	opOwnershipDemote       = "ownership.demote"
	opOwnershipTransfer     = "ownership.transfer"
	opInvitationList        = "invitation.list"
	opInvitationMine        = "invitation.mine"
	opInvitationCandidates  = "invitation.candidates"
	opInvitationIssue       = "invitation.issue"
	opInvitationRevoke      = "invitation.revoke"
	opInvitationGet         = "invitation.get"
	opInvitationAccept      = "invitation.accept"
	opInvitationConsume     = "invitation.consume"
	opRetirementEligibility = "retirement.eligibility"
	// G058: the enrollment admission (an Organization invitation claim plus lift, or open admission), also an
	// access.denied operation (00016).
	opEnrollmentAdmit = "enrollment.admit"
	// G058 slice 1b: the guarded enrollment policy read (organization.read) and write (organization.manage_policy).
	opEnrollmentPolicyGet    = "enrollment.policy.get"
	opEnrollmentPolicyUpdate = "enrollment.policy.update"
	// G058 request slice: the applicant's request, status read and abandonment (own credential, no capability; the
	// request admission is enrollment.admit) and the three manager operations (organization.manage_members). Each is an
	// access.denied operation of 00016; only enrollment.request (with enrollment.admit) records its known denials.
	opEnrollmentRequest        = "enrollment.request"
	opEnrollmentStatus         = "enrollment.status"
	opEnrollmentAbandon        = "enrollment.abandon"
	opEnrollmentRequestsList   = "enrollment.requests.list"
	opEnrollmentRequestsDecide = "enrollment.requests.decide"
	opEnrollmentRequestsRevoke = "enrollment.requests.revoke"
	// G060: the owner-only Group-management override write (organization.manage_policy; also an access.denied operation,
	// 00017). Its read is organization.get_policy (organization.read).
	opSetGroupOverride = "organization.set_group_override"
	// G020 (Plan r2 §4.1): the seven guarded label operations (each also an access.denied operation, 00020). No record
	// carries a label name, target id or payload, and no label operation appends access.override_used.
	opLabelDefinitionList   = "label.definition.list"
	opLabelDefinitionCreate = "label.definition.create"
	opLabelDefinitionUpdate = "label.definition.update"
	opLabelDefinitionDelete = "label.definition.delete"
	opLabelAssignmentList   = "label.assignment.list"
	opLabelAssign           = "label.assign"
	opLabelUnassign         = "label.unassign"
)

// The outcomes and their levels: established (info) for a context operation's success, allow (info) for a guarded
// operation's, deny (warn) and failed (error: unavailable or internal) for both.
const (
	outcomeEstablished = "established"
	outcomeAllow       = "allow"
	outcomeDeny        = "deny"
	outcomeFailed      = "failed"
)

// logContext records a context operation: it carries no capability, because establishing authentication and context
// is not a policy decision.
func (s *Service) logContext(ctx context.Context, op string, f *failure) {
	s.logDecision(ctx, op, "", outcomeEstablished, reasonContextEstablished, f)
}

// logGuarded records a guarded operation with its single capability.
func (s *Service) logGuarded(ctx context.Context, op string, c Capability, f *failure) {
	s.logDecision(ctx, op, c, outcomeAllow, reasonAllowed, f)
}

// logGuardedAllow is logGuarded with the decision's allow reason (G060: allowed_override for the Organization override
// path; "" or any other value records allowed).
func (s *Service) logGuardedAllow(ctx context.Context, op string, c Capability, allow Reason, f *failure) {
	if allow != reasonAllowedOverride {
		allow = reasonAllowed
	}
	s.logDecision(ctx, op, c, outcomeAllow, allow, f)
}

func (s *Service) logDecision(ctx context.Context, op string, c Capability, success string, successReason Reason, f *failure) {
	level, outcome, reason := slog.LevelInfo, success, successReason
	if f != nil {
		level, outcome, reason = slog.LevelWarn, outcomeDeny, f.reason
		if f.code == CodeUnavailable || f.code == CodeInternal {
			level, outcome = slog.LevelError, outcomeFailed
		}
	}
	attrs := []slog.Attr{slog.String("operation", op)}
	if c != "" {
		attrs = append(attrs, slog.String("capability", string(c)))
	}
	attrs = append(attrs, slog.String("outcome", outcome), slog.String("reason", string(reason)))
	if rid := observability.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid), slog.String("trace_id", observability.TraceID(ctx)))
	}
	s.logger.LogAttrs(context.Background(), level, eventDecision, attrs...)
}

// eventStorage is the one catalogued record of every guarded storage call (G009 Plan r1 §4.8; events.json): the
// operation, outcome and closed reason, the Put/Delete stage, UUID attribution (R9: actor, Organization, optional
// Group, object once known), the issued TTL of a signed URL, the physical outcome of a delete, and request/trace
// correlation when present. Never a URL, signature, credential, key, payload, content type, size, filename or provider
// or driver text. Operational evidence, not the G011 audit; storage calls do not also write access.decision.
const eventStorage = "storage.operation"

// The storage operations.
const (
	opObjectPut      = "object.put"
	opObjectGet      = "object.get"
	opObjectMetadata = "object.metadata"
	opObjectSign     = "object.sign"
	opObjectDelete   = "object.delete"
)

// The Put and Delete stages.
const (
	stageT1         = "t1"
	stageProvider   = "provider"
	stageVerify     = "verify"
	stageT2         = "t2"
	stageReconcile  = "reconcile"
	stageCompensate = "compensate"
	stageAudit      = "audit" // G011: the signed-URL issuance record failed, so no URL was returned
)

// The physical outcomes of a delete.
const (
	physicalConfirmed = "confirmed"
	physicalPending   = "pending"
)

// storageRecord collects one storage call's record fields as the call proceeds.
type storageRecord struct {
	op       string
	bound    bool // the binding check passed: the attribution below is trusted
	actor    tenancy.UserID
	org      tenancy.OrganizationID
	group    tenancy.GroupID
	object   tenancy.ObjectID
	stage    string
	ttl      int64  // seconds; sign only
	physical string // physicalConfirmed or physicalPending; delete only
	reason   Reason // the allow reason when not "allowed" (reconciled_ready, or a pending delete's cause)
}

// logStorage writes the storage.operation record of one call: allow (info) on success; deny (warn) for a refusal;
// failed (error) for unavailable or internal, and for a ready object whose provider bytes are missing.
func (s *Service) logStorage(ctx context.Context, r *storageRecord, f *failure) {
	level, outcome, reason := slog.LevelInfo, outcomeAllow, reasonAllowed
	if r.reason != "" {
		reason = r.reason
	}
	if f != nil {
		level, outcome, reason = slog.LevelWarn, outcomeDeny, f.reason
		if f.code == CodeUnavailable || f.code == CodeInternal || f.reason == reasonObjectBytesMissing {
			level, outcome = slog.LevelError, outcomeFailed
		}
	}
	attrs := []slog.Attr{slog.String("operation", r.op), slog.String("outcome", outcome), slog.String("reason", string(reason))}
	if r.stage != "" {
		attrs = append(attrs, slog.String("stage", r.stage))
	}
	if r.bound {
		attrs = append(attrs, slog.String("actor_id", r.actor.String()), slog.String("organization_id", r.org.String()))
		if !zero(r.group) {
			attrs = append(attrs, slog.String("group_id", r.group.String()))
		}
		if !zero(r.object) {
			attrs = append(attrs, slog.String("object_id", r.object.String()))
		}
	}
	if r.ttl > 0 && f == nil {
		attrs = append(attrs, slog.Int64("ttl_seconds", r.ttl))
	}
	if r.physical != "" && f == nil {
		attrs = append(attrs, slog.String("physical", r.physical))
	}
	if rid := observability.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid), slog.String("trace_id", observability.TraceID(ctx)))
	}
	s.logger.LogAttrs(context.Background(), level, eventStorage, attrs...)
}

// eventAuditRecordFailed is the G011 diagnostic of an independent audit record that could not be written (events.json):
// the catalog action only, plus correlation when present. The caller's outcome is unchanged and no durable row is
// claimed; if the database is unavailable this record is the only trace.
const eventAuditRecordFailed = "audit.record_failed"

func (s *Service) logAuditRecordFailed(ctx context.Context, action audit.Action) {
	attrs := []slog.Attr{slog.String("action", string(action))}
	if rid := observability.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid), slog.String("trace_id", observability.TraceID(ctx)))
	}
	s.logger.LogAttrs(context.Background(), slog.LevelError, eventAuditRecordFailed, attrs...)
}
