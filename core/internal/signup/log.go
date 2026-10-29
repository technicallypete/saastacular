package signup

import (
	"context"
	"log/slog"
	"slices"

	"core/internal/observability"
)

// EventRequest is the one catalogued log event of root signup (events.json signup.request): one record per signup
// request with the operation, the closed outcome and, for a refusal or failure, its closed reason, plus request/trace
// correlation when present. Never a credential, a continuation secret, a name, a slug, an origin, an ID or a payload.
const EventRequest = "signup.request"

// The logged operations (the signup/api routes, plus the closed method/path answers).
const (
	OpStart     = "start"
	OpStatus    = "status"
	OpSlug      = "slug"
	OpCreate    = "create"
	OpChallenge = "challenge"
	OpProve     = "prove"
	OpComplete  = "complete"
	OpAbandon   = "abandon"
	OpMethod    = "method"
	OpPath      = "path"
)

var logOperations = []string{OpStart, OpStatus, OpSlug, OpCreate, OpChallenge, OpProve, OpComplete, OpAbandon, OpMethod, OpPath}

// logReasons are the closed refusal/failure reasons of signup.request (the signup codes plus the HTTP adapter's).
var logReasons = []string{string(CodeInvalidRequest), string(CodeUnauthenticated), string(CodeCSRFFailed), string(CodeSlugTaken),
	string(CodeIntentUsed), string(CodeUserDisabled), string(CodeContinuationInvalid), string(CodeContextMismatch),
	string(CodeUnconfigured), string(CodeUnavailable), "origin_rejected", "rate_limited", "not_found", "method_not_allowed"}

// LogRequest writes one signup.request record: reason "" is ok (info); signup_unavailable is failed (error); any other
// reason is refused (warn). An operation or reason outside the closed sets is not logged.
func LogRequest(ctx context.Context, logger *slog.Logger, op, reason string) {
	if logger == nil || !slices.Contains(logOperations, op) || (reason != "" && !slices.Contains(logReasons, reason)) {
		return
	}
	attrs := []slog.Attr{slog.String("operation", op)}
	level := slog.LevelInfo
	switch reason {
	case "":
		attrs = append(attrs, slog.String("outcome", "ok"))
	case string(CodeUnavailable):
		level = slog.LevelError
		attrs = append(attrs, slog.String("outcome", "failed"), slog.String("reason", reason))
	default:
		level = slog.LevelWarn
		attrs = append(attrs, slog.String("outcome", "refused"), slog.String("reason", reason))
	}
	if rid := observability.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid), slog.String("trace_id", observability.TraceID(ctx)))
	}
	logger.LogAttrs(context.Background(), level, EventRequest, attrs...)
}
