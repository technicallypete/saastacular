package tenancy

import (
	"context"
	"log/slog"

	"core/internal/observability"
)

// eventOperation is the one catalogued log event of this package (events.json): one record per mutation, with a
// bounded operation, outcome and error code plus request/trace correlation when present. Never IDs, names, slugs,
// roles, driver errors or payloads; reads are not logged.
const eventOperation = "tenancy.operation"

// The logged mutation names.
const (
	opOrganizationCreate = "organization.create"
	opOrganizationRename = "organization.rename"
	opGroupCreate        = "group.create" // one record for the atomic Group plus creator-owner transaction
	opGroupRename        = "group.rename"
	opGroupChangeSlug    = "group.change_slug"
	opMembershipGrant    = "membership.grant"
	opMembershipChange   = "membership.change"
	opMembershipRemove   = "membership.remove"

	// G006: BusinessProfile and association mutations; they carry no scope_type.
	opBusinessProfileCreate = "business_profile.create"
	opBusinessProfileUpdate = "business_profile.update"
	opAssociationLink       = "association.link"
	opAssociationUnlink     = "association.unlink"

	// G019: the trusted ordinary retirement (access.RetireUser), a tenancy-level operation without an access.decision
	// (Plan r1 §5.2); it carries no scope_type.
	opRetirementRetire = "retirement.retire"
)

// LogRetirement writes access.RetireUser's one tenancy.operation record: operation retirement.retire, the outcome and,
// for a refusal or failure, its closed code (a tenancy code, the access refusal field retirement_blocked, or the access
// codes unavailable/internal/invalid_input). Never the subject, an Organization, a count or a cause.
func LogRetirement(ctx context.Context, logger *slog.Logger, code Code) {
	if logger == nil {
		return
	}
	var e *Error
	if code != "" {
		e = &Error{Code: code}
	}
	(&Service{logger: logger}).logOperation(ctx, opRetirementRetire, "", e)
}

// The logged outcomes and their levels: ok (info), rejected (warn: validation, not found, conflict) and failed (error:
// unavailable or internal).
const (
	outcomeOK       = "ok"
	outcomeRejected = "rejected"
	outcomeFailed   = "failed"
)

func (s *Service) logOperation(ctx context.Context, op string, scope ScopeType, e *Error) {
	level, outcome := slog.LevelInfo, outcomeOK
	if e != nil {
		level, outcome = slog.LevelError, outcomeFailed
		if rejected(e.Code) {
			level, outcome = slog.LevelWarn, outcomeRejected
		}
	}
	attrs := []slog.Attr{slog.String("operation", op)}
	if scope != "" {
		attrs = append(attrs, slog.String("scope_type", string(scope)))
	}
	attrs = append(attrs, slog.String("outcome", outcome))
	if e != nil {
		attrs = append(attrs, slog.String("error_code", string(e.Code)))
	}
	if rid := observability.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid), slog.String("trace_id", observability.TraceID(ctx)))
	}
	s.logger.LogAttrs(context.Background(), level, eventOperation, attrs...)
}
