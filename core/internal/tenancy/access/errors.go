package access

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/identity"
	"core/internal/tenancy"
)

// Code is the closed external error category of the facade (docs/foundation/AUTHORIZATION.md). An inaccessible target
// is indistinguishable from an unknown one: both are not_found.
type Code string

// The external codes.
const (
	CodeUnauthenticated Code = "unauthenticated" // absent, invalid, expired or revoked session; disabled user
	CodeContextDenied   Code = "context_denied"  // request scope or context binding, or the context no longer holds
	CodeNotFound        Code = "not_found"       // unknown, foreign, kind-mismatched, outside-context or unreadable target
	CodeForbidden       Code = "forbidden"       // readable target, capability denied
	CodeInvalidInput    Code = "invalid_input"   // malformed selector or payload; Field carries the tenancy code
	CodeConflict        Code = "conflict"        // slug_conflict; G018: Field slug or association_exists
	// G016 domain lifecycle codes.
	CodeHostTaken            Code = "host_taken"             // a live claim of the hostname exists (any Organization)
	CodeDomainStateInvalid   Code = "domain_state_invalid"   // the operation does not apply to the row's state, or a raced change
	CodeDomainRetryExhausted Code = "domain_retry_exhausted" // the explicit withdrawal retry bound is reached
	CodeUnavailable          Code = "unavailable"            // identity or database unavailable, abort, cancellation, lost commit reply
	CodeInternal             Code = "internal"               // unexpected tenancy internal error

	// G017: a branding write whose effective palette fails a pinned contrast pair; Field is "<pair id>:<appearance>".
	CodeContrastFailed Code = "contrast_failed"
)

// Error is every error the facade returns. Error() is "access: <code>"; it never carries IDs, names, roles, policy
// values or driver text. Field is set only for invalid_input and (G017) contrast_failed, and for SetBrandingLogo's closed
// logo outcomes after its upload (publication_outcome_unknown, compensated, compensation_pending, compensation_failed),
// which alone also set Object (the uploaded object's ID, in the caller's own active Organization) and ObjectState (its
// actual live state as read or as G009 returned it: ready, deleting, deleted, absent or unread).
type Error struct {
	Code        Code
	Field       string
	Object      string
	ObjectState string
}

func (e *Error) Error() string { return "access: " + string(e.Code) }

// CodeOf returns the code of an access error, or "" for any other error.
func CodeOf(err error) Code {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// failure is one denied or failed call: its external code and field, and the internal reason that only the
// access.decision record carries. denied is set only for a recordable denial (G011, markDenied): the scope its
// transaction established, carried to the independent access.denied record and nowhere else.
type failure struct {
	code   Code
	field  string
	reason Reason
	denied *denialScope
	// G017: SetBrandingLogo's uploaded object and its live state (Error.Object, Error.ObjectState).
	object, objectState string
}

func (f *failure) err() error {
	return &Error{Code: f.code, Field: f.field, Object: f.object, ObjectState: f.objectState}
}

func fail(code Code, reason Reason) *failure { return &failure{code: code, reason: reason} }

// failureReason derives the reason of a store failure from the context and the first driver error the transaction saw:
// canceled for a cancellation or deadline, aborted for a definitive transaction abort (class 40: serialization failure,
// deadlock), store_unavailable otherwise.
func failureReason(ctx context.Context, cause error) Reason {
	var pg *pgconn.PgError
	switch {
	case ctx.Err() != nil || errors.Is(cause, context.Canceled) || errors.Is(cause, context.DeadlineExceeded):
		return reasonCanceled
	case errors.As(cause, &pg) && len(pg.Code) == 5 && pg.Code[:2] == "40":
		return reasonAborted
	}
	return reasonStoreUnavailable
}

// commitFailure classifies a failed COMMIT (G007 Plan r1 §4.5; Codex review 1bed0ddb §C2). A definitive abort is
// reported only when it is established: the COMMIT was never sent (the driver reports the error safe to retry; the
// reason is failureReason's), the server answered the COMMIT with ROLLBACK (pgx.ErrTxCommitRollback; aborted), or an
// ordinary command-severity ERROR rejected the COMMIT, so the server rolled the transaction back (aborted).
//
// Everything else may follow a COMMIT that was sent and executed, so the outcome is commit_outcome_unknown and rollback
// is never claimed: a session-ending FATAL (for example 57P01, administrator shutdown) or a PANIC (for example 57P02,
// crash shutdown), a server error whose severity is missing or not recognized, and a lost reply (EOF, connection
// reset, cancellation or deadline during COMMIT). Every case is the public code unavailable, with no retry.
func commitFailure(ctx context.Context, err error) *failure {
	var pg *pgconn.PgError
	switch {
	case pgconn.SafeToRetry(err):
		return fail(CodeUnavailable, failureReason(ctx, err))
	case errors.Is(err, pgx.ErrTxCommitRollback):
		return fail(CodeUnavailable, reasonAborted)
	case errors.As(err, &pg) && commandError(pg):
		return fail(CodeUnavailable, reasonAborted)
	}
	return fail(CodeUnavailable, reasonCommitOutcomeUnknown)
}

// commandError reports whether a server error ended only the command (severity ERROR), not the session (FATAL) or the
// server (PANIC). It reads the nonlocalized severity (sent by PostgreSQL 9.6 and later) and falls back to the localized
// one only when no nonlocalized severity was sent; only the exact value ERROR counts, so a missing, translated or
// unrecognized severity is never taken as proof of rollback.
func commandError(pg *pgconn.PgError) bool {
	severity := pg.SeverityUnlocalized
	if severity == "" {
		severity = pg.Severity
	}
	return severity == "ERROR"
}

// sessionFailure maps a session-liveness error: unauthenticated, or unavailable with the transaction's reason.
func sessionFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	if errors.Is(err, identity.ErrUnauthenticated) {
		return fail(CodeUnauthenticated, reasonSessionInvalid)
	}
	return fail(CodeUnavailable, failureReason(ctx, tx.cause))
}

// writeFailure maps a tenancy error of a guarded write. Payloads are validated before authorization, so an input code
// here comes from a database rule; a missing target was row-locked by the facts and cannot occur.
func writeFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	switch c := tenancy.CodeOf(err); c {
	case tenancy.CodeSlugConflict:
		return fail(CodeConflict, reasonWriteConflict)
	case tenancy.CodeAssociationExists: // G018: the association unique key behind a completion's locked state check
		return &failure{code: CodeConflict, field: string(c), reason: reasonWriteConflict}
	case tenancy.CodeGroupNotFound, tenancy.CodeBusinessProfileNotFound, tenancy.CodeOrganizationNotFound:
		return fail(CodeNotFound, reasonTargetNotFound)
	case tenancy.CodeUnavailable:
		return fail(CodeUnavailable, failureReason(ctx, tx.cause))
	case tenancy.CodeInternal, "":
		return fail(CodeInternal, reasonStoreUnavailable)
	default:
		return &failure{code: CodeInvalidInput, field: string(c), reason: reasonInputInvalid}
	}
}

// membershipFailure maps a tenancy error of a G019 membership, ownership or invitation write (Plan r1 §1.4 "Failure
// semantics"): the authority guard's and claim bodies' refusals onto conflict, forbidden or invalid_input with their
// field; a binding or invitation that is gone or invisible onto not_found; everything else as writeFailure (a class-40
// abort stays unavailable with reason aborted, r3 addendum).
func membershipFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	switch c := tenancy.CodeOf(err); c {
	case tenancy.CodeLastOwner, tenancy.CodeDependentBindings, tenancy.CodeRecipientIneligible, tenancy.CodeIssuerAuthorityLost,
		tenancy.CodeInvitationLimit, tenancy.CodeInvitationNotPending, tenancy.CodeMembershipExists:
		return &failure{code: CodeConflict, field: string(c), reason: reasonWriteConflict}
	case tenancy.CodeOwnerAuthority, tenancy.CodeSelfEscalation, tenancy.CodeOwnerSubject:
		// G060: self_escalation is the guard's refusal of an own Group role change (00017 step (3b)); owner_subject its
		// refusal of an administrative Organization departure of a Group-owner subject (step (3), amendment01).
		return &failure{code: CodeForbidden, field: string(c), reason: reasonRoleNotPermitted}
	case tenancy.CodeSelfTarget:
		return &failure{code: CodeInvalidInput, field: fieldSelfTarget, reason: reasonInputInvalid}
	case tenancy.CodeContextRequired:
		return &failure{code: CodeInvalidInput, field: fieldContext, reason: reasonInputInvalid}
	case tenancy.CodeMembershipNotFound, tenancy.CodeInvitationNotFound, tenancy.CodeUserNotFound:
		return fail(CodeNotFound, reasonTargetNotFound)
	}
	return writeFailure(ctx, tx, err)
}

// The G019 failure fields that are not tenancy codes (Plan r1 §1.4).
const (
	fieldContext      = "context"       // an Organization departure or removal needs an Organization-only context
	fieldSelfTarget   = "self_target"   // an administrative change of the actor's own binding
	fieldOwnerSubject = "owner_subject" // an ordinary change or removal of an owner (Spec R1)
	fieldNotOwner     = "not_owner"     // a demotion of a subject who is not an owner
	fieldLimit        = "limit"         // a page limit outside 1..MaxMembersPage
	fieldCursor       = "after"         // a malformed page cursor
	fieldFilter       = "filter"        // an unknown member filter
	fieldScope        = "scope"         // a malformed scope
	fieldRecipient    = "recipient"     // a recipient of the wrong form for the scope
)

// fieldRetirementBlocked refuses a trusted retirement while the subject solely owns a scope (Result addendum A1-01).
const fieldRetirementBlocked = "retirement_blocked"

// The G058 enrollment admission fields (Plan r3 §1.4 closed outcome set): the refused outcome of a conflict or
// unauthenticated admission, and the invalid_input field of a malformed enrollment handle.
const (
	fieldEnrollment           = "enrollment"            // a zero or non-canonical handle user, Organization or digest
	fieldEnrollmentClosed     = "enrollment_closed"     // open admission while the Organization is not open
	fieldEnrollmentRestricted = "enrollment_restricted" // an unlifted restriction (open) or one not older than the invitation (lift)
	fieldEnrollmentExpired    = "enrollment_expired"    // the enrollment credential is expired, revoked or not the caller's
	fieldHostChanged          = "host_changed"          // the credential's origin no longer routes to its Organization
)

// The G058 semantic reasons of an admission refusal (Codex round-2 audit/log decision): the same reason names the
// access.decision record and the independent access.denied record (audit catalog, 00016; logrecord events.json);
// session_invalid (the enrollment_expired aggregate) and target_not_found are the existing reasons.
const (
	reasonEnrollmentClosed     Reason = "enrollment_closed"
	reasonEnrollmentRestricted Reason = "enrollment_restricted"
	reasonRecipientIneligible  Reason = "recipient_ineligible"
	reasonHostChanged          Reason = "host_changed"
)

// G058 slice 1b (selected-design addendum §A.5): open refused while the Organization's membership history is unresolved
// — conflict with this field (HTTP 409 enrollment_not_ready) and the same access.decision reason; no UPDATE is issued
// and no audit row is written (not an access.denied reason). The invalid_input field of an unknown mode is the tenancy
// code invalid_enrollment_mode.
const (
	fieldEnrollmentNotReady         = "enrollment_not_ready"
	reasonEnrollmentNotReady Reason = "enrollment_not_ready"
)

// G058 request slice (Plan r3 §1.4/§1.5/§5.2): the request kind's admission refusals (conflict with the outcome as
// field: pending and denied are status answers, never recorded denials, so their decision reason is write_conflict;
// approval_revoked is a recorded denial with its semantic reason), the applicant's request already answered by a
// current membership (already_member, not recorded), the manager operations' refusals (not_pending, not_approved,
// already_consumed and enrollment_request_transition, each the tenancy code as field) and their invalid_input fields
// (an unknown decision, an unknown status filter).
const (
	fieldPending         = "pending"
	fieldDenied          = "denied"
	fieldApprovalRevoked = "approval_revoked"
	fieldAlreadyMember   = "already_member"
	fieldDecision        = "decision"
	fieldStatus          = "status"
)

// reasonApprovalRevoked is the access.decision and access.denied reason of a request admission whose approval is no
// longer current authority (00016 catalog reason approval_revoked; enrollment.admit only).
const reasonApprovalRevoked Reason = "approval_revoked"

// ErrEnrollmentNotReady is UpdateEnrollmentPolicy's refusal of open on an Organization with unresolved membership
// history (409 enrollment_not_ready); errors.Is matches it by code and field.
var ErrEnrollmentNotReady error = &Error{Code: CodeConflict, Field: fieldEnrollmentNotReady}

// Is reports whether target is an *Error with the same code, field and object outcome (ErrEnrollmentNotReady).
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t != nil && *t == *e
}

// objectStoreFailure maps a tenancy error of a stored-object primitive (G009): a missing, invisible or no longer
// matching row is not_found; a store failure is unavailable with the transaction's reason; anything else (a refused
// insert or update: a row-level security, trigger or constraint refusal) is internal, never a policy outcome, since the
// decision was already made on the facts.
func objectStoreFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	switch tenancy.CodeOf(err) {
	case tenancy.CodeObjectNotFound:
		return fail(CodeNotFound, reasonTargetNotFound)
	case tenancy.CodeUnavailable:
		return fail(CodeUnavailable, failureReason(ctx, tx.cause))
	}
	return fail(CodeInternal, reasonStoreUnavailable)
}

// The G020 label failure fields (Plan r2 §1.3, §1.4, §4.1). The conflict fields and target_unsupported and
// visibility_not_allowed equal the tenancy codes of the same refusal (00020's named constraints, mapped by
// tenancy.classify); target_ambiguous and organization_context_required are refused in this package before any SQL.
const (
	fieldNameConflict                = string(tenancy.CodeNameConflict)         // conflict: the scoped name key is taken
	fieldScopeFull                   = string(tenancy.CodeScopeFull)            // conflict: the owner scope holds 200 definitions
	fieldAlreadyAssigned             = string(tenancy.CodeAlreadyAssigned)      // conflict: a duplicate assignment
	fieldNotAssigned                 = string(tenancy.CodeNotAssigned)          // conflict: unassign of an absent assignment
	fieldTargetUnsupported           = string(tenancy.CodeTargetUnsupported)    // invalid_input: unknown kind or inadmissible owner scope
	fieldVisibilityNotAllowed        = string(tenancy.CodeVisibilityNotAllowed) // invalid_input: target_members on an Organization label
	fieldTargetAmbiguous             = "target_ambiguous"                       // invalid_input: more than one target named
	fieldOrganizationContextRequired = "organization_context_required"          // invalid_input: Organization label delete in a selected Group
)

// labelFailure maps a tenancy error of a guarded label primitive: the conflict refusals onto conflict with their field;
// a definition the context cannot see (or a total-key refusal naming an absent definition) onto not_found; the
// input-shaped refusals of the label tables onto invalid_input with their field; everything else as writeFailure (a
// class-40 abort stays unavailable with reason aborted).
func labelFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	switch c := tenancy.CodeOf(err); c {
	case tenancy.CodeNameConflict, tenancy.CodeScopeFull, tenancy.CodeAlreadyAssigned, tenancy.CodeNotAssigned:
		return &failure{code: CodeConflict, field: string(c), reason: reasonWriteConflict}
	case tenancy.CodeLabelNotFound:
		return fail(CodeNotFound, reasonTargetNotFound)
	case tenancy.CodeTargetUnsupported, tenancy.CodeVisibilityNotAllowed, tenancy.CodeInvalidVisibility, tenancy.CodeInvalidScope,
		tenancy.CodeInvalidName:
		return &failure{code: CodeInvalidInput, field: string(c), reason: reasonInputInvalid}
	}
	return writeFailure(ctx, tx, err)
}

// labelInputFailure is a G020 refusal decided in this package before any SQL (invalid_input with field).
func labelInputFailure(field string) *failure {
	return &failure{code: CodeInvalidInput, field: field, reason: reasonInputInvalid}
}

// inputFailure maps a pure validation error (a tenancy parser or validator) onto invalid_input with its code.
func inputFailure(err error, reason Reason) *failure {
	return &failure{code: CodeInvalidInput, field: string(tenancy.CodeOf(err)), reason: reason}
}
