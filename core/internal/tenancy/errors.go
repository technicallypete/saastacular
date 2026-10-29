package tenancy

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/jackc/pgx/v5/pgconn"
)

// Code is a bounded error category. Codes never carry IDs, names, slugs or driver text.
type Code string

// Error codes (docs/foundation/TENANCY.md). Unknown and foreign selectors produce the identical not-found code.
const (
	CodeInvalidID            Code = "invalid_id"
	CodeInvalidName          Code = "invalid_name"
	CodeInvalidSlug          Code = "invalid_slug"
	CodeSlugReserved         Code = "slug_reserved"
	CodeSlugRequired         Code = "slug_required"
	CodeInvalidRole          Code = "invalid_role"
	CodeOrganizationNotFound Code = "organization_not_found"
	CodeGroupNotFound        Code = "group_not_found"
	CodeUserNotFound         Code = "user_not_found"
	CodeMembershipNotFound   Code = "membership_not_found"
	CodeSlugConflict         Code = "slug_conflict"
	CodeMembershipExists     Code = "membership_exists"
	CodeSlugUnavailable      Code = "slug_unavailable"
	CodeUnavailable          Code = "unavailable"
	CodeInternal             Code = "internal"

	// G006: BusinessProfile context fields, profiles and Group associations.
	CodeInvalidDisplayName      Code = "invalid_display_name"
	CodeInvalidLegalName        Code = "invalid_legal_name"
	CodeInvalidDescription      Code = "invalid_description"
	CodeInvalidIndustry         Code = "invalid_industry"
	CodeInvalidWebsite          Code = "invalid_website"
	CodeBusinessProfileNotFound Code = "business_profile_not_found"
	CodeAssociationNotFound     Code = "association_not_found"
	CodeAssociationExists       Code = "association_exists"

	// G008: InitContext found prior request context on the connection (docs/foundation/DATABASE-ISOLATION.md). The
	// caller discards the connection and fails closed; nothing protected has run.
	CodeContextPresent Code = "context_present"

	// G009: stored-object selectors and inputs (docs/foundation/OBJECT-STORAGE.md). The storage operations report the
	// input codes as the Field of access invalid_input.
	CodeObjectNotFound     Code = "object_not_found"
	CodeInvalidContentType Code = "invalid_content_type"
	CodeObjectSizeRequired Code = "object_size_required"
	CodeObjectTooLarge     Code = "object_too_large"
	CodeObjectSizeMismatch Code = "object_size_mismatch"
	CodeMissingBody        Code = "missing_body"
	CodeInvalidTTL         Code = "invalid_ttl"

	// G019: invitation inputs (Plan r1 §1.4). The guarded operations report them as the Field of access invalid_input.
	CodeInvalidAddress  Code = "invalid_address"
	CodeInvalidToken    Code = "invalid_token"
	CodeInvalidLifetime Code = "invalid_lifetime"

	// G019: the refusals of 00015's guards and claim bodies, named by their RAISE ... CONSTRAINT (Plan r1 §1.4 failure
	// semantics; r2 §4: a class-40 abort stays unavailable). The access layer maps them onto its public codes and fields.
	CodeLastOwner            Code = "last_owner"
	CodeOwnerAuthority       Code = "owner_authority"
	CodeRecipientIneligible  Code = "recipient_ineligible"
	CodeDependentBindings    Code = "dependent_bindings"
	CodeContextRequired      Code = "context_required"
	CodeIssuerAuthorityLost  Code = "issuer_authority_lost"
	CodeInvitationLimit      Code = "invitation_limit"
	CodeInvitationNotPending Code = "invitation_not_pending"
	CodeInvitationNotFound   Code = "invitation_not_found"

	// G060 (00017_group_authority.sql): the authority guard's own-Group-role refusal (42501 CONSTRAINT self_escalation:
	// a runtime or owner session changing the role of its own Group binding without that Group's owner role) and a Group
	// invitation whose recipient is its issuer (refused before SQL; the claim body also never consumes one).
	CodeSelfEscalation Code = "self_escalation"
	CodeSelfTarget     Code = "self_target"
	// G060 amendment01: the authority guard's Organization-departure cascade refuses an administrative removal of a
	// subject who owns a Group of the Organization (42501 CONSTRAINT owner_subject; that Group's succession comes first).
	CodeOwnerSubject Code = "owner_subject"

	// G058: the lift kind of app.enrollment_admit raises P0001 with exactly this message when an unlifted restriction is
	// not older than the claimed invitation (Plan r3 §1.4), so the claim in the same transaction rolls back.
	CodeEnrollmentRestricted Code = "enrollment_restricted"
	// G058 slice 1b: an enrollment mode outside the closed set (organizations_enrollment_mode_valid; the access policy
	// write validates it before any statement), and open on an Organization whose membership history is unresolved
	// (organizations_open_requires_resolved_history, the authoritative CHECK; addendum §A.5 defense in depth).
	CodeInvalidEnrollmentMode Code = "invalid_enrollment_mode"
	CodeEnrollmentNotReady    Code = "enrollment_not_ready"
	// G058 request slice (Plan r3 §1.4 "Approval revocation", §1.5; 00016 enrollment_requests): a request the context
	// cannot see (unknown, foreign, or not the caller's); a decision of a request that is no longer pending; a revocation
	// of a request that is pending or denied; a revocation of an approval already consumed by its admission; a refused
	// state transition (enrollment_requests_immutable, 42501 CONSTRAINT enrollment_request_transition: an illegal
	// transition, a changed identity or authority, or a decision/revocation attributed to anyone but the acting user);
	// and an applicant credential that is missing, revoked, expired or not the caller's.
	CodeEnrollmentRequestNotFound   Code = "enrollment_request_not_found"
	CodeEnrollmentNotPending        Code = "not_pending"
	CodeEnrollmentNotApproved       Code = "not_approved"
	CodeEnrollmentAlreadyConsumed   Code = "already_consumed"
	CodeEnrollmentRequestTransition Code = "enrollment_request_transition"
	CodeEnrollmentExpired           Code = "enrollment_expired"

	// G020 (00020_labels.sql; Plan r2 §1.2, §4.1): label inputs and the named refusals of the two label tables. The
	// guarded label operations report them as the Field of access invalid_input or conflict (the conflict codes are the
	// public field names), and label_not_found as not_found.
	CodeInvalidVisibility    Code = "invalid_visibility"     // a visibility outside owner_scope/target_members
	CodeInvalidScope         Code = "invalid_scope"          // an owner scope that is not consistent (label_definitions_scope_consistent)
	CodeVisibilityNotAllowed Code = "visibility_not_allowed" // target_members on an Organization-owned definition
	CodeTargetUnsupported    Code = "target_unsupported"     // an assignment shape production storage does not admit
	CodeLabelNotFound        Code = "label_not_found"        // a definition the context cannot see (unknown, foreign, unreadable)
	CodeNameConflict         Code = "name_conflict"          // the scoped ASCII-case-insensitive name key is taken
	CodeScopeFull            Code = "scope_full"             // the owner scope already holds MaxLabelDefinitionsPerScope definitions
	CodeAlreadyAssigned      Code = "already_assigned"       // the definition is already assigned to the target
	CodeNotAssigned          Code = "not_assigned"           // the definition is not assigned to the target
)

// Error is every error the Service returns. Error() is "tenancy: <code>". Unwrap exposes only a context cancellation
// or deadline error; PostgreSQL messages and details, which can contain row values, never leave the package.
type Error struct {
	Code  Code
	cause error
}

func (e *Error) Error() string { return "tenancy: " + string(e.Code) }

// Unwrap returns context.Canceled or context.DeadlineExceeded when that caused the error, otherwise nil.
func (e *Error) Unwrap() error { return e.cause }

func newError(c Code) *Error { return &Error{Code: c} }

// CodeOf returns the code of a tenancy error, or "" for any other error.
func CodeOf(err error) Code {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// rejected reports whether a code is a rejection (validation, not found or conflict) rather than a failure.
func rejected(c Code) bool { return c != CodeUnavailable && c != CodeInternal }

// uniqueCodes, foreignKeyCodes and checkCodes map the named constraints of 00003_tenancy.sql and
// 00004_business_profiles.sql onto codes. Any other constraint is unexpected and maps to internal.
var (
	uniqueCodes = map[string]Code{
		"groups_organization_slug_key":      CodeSlugConflict,
		"role_bindings_subject_scope_key":   CodeMembershipExists,
		"group_business_profiles_group_key": CodeAssociationExists,
		// G020 (00020_labels.sql): the scoped name key and the one-assignment-per-target key. The primary keys and the
		// composite owner key are keyed-insert collisions (labelKeyCollisions: unavailable), never conflicts.
		"label_definitions_scope_name_key": CodeNameConflict,
		"label_assignments_unique":         CodeAlreadyAssigned,
	}
	foreignKeyCodes = map[string]Code{
		"groups_organization_fkey":                      CodeOrganizationNotFound,
		"role_bindings_organization_fkey":               CodeOrganizationNotFound,
		"role_bindings_group_fkey":                      CodeGroupNotFound,
		"role_bindings_subject_fkey":                    CodeUserNotFound,
		"business_profiles_organization_fkey":           CodeOrganizationNotFound,
		"group_business_profiles_organization_fkey":     CodeOrganizationNotFound,
		"group_business_profiles_group_fkey":            CodeGroupNotFound,
		"group_business_profiles_business_profile_fkey": CodeBusinessProfileNotFound,
		"stored_objects_organization_fkey":              CodeOrganizationNotFound,
		"stored_objects_group_fkey":                     CodeGroupNotFound,
		"stored_objects_creator_fkey":                   CodeUserNotFound,
		// G019 (00015_membership_invitations.sql).
		"invitations_organization_fkey":  CodeOrganizationNotFound,
		"invitations_group_fkey":         CodeGroupNotFound,
		"invitations_issuer_fkey":        CodeUserNotFound,
		"invitations_intended_user_fkey": CodeUserNotFound,
		"invitations_consumed_by_fkey":   CodeUserNotFound,
		// G020 (00020_labels.sql): the total composite assignment -> definition key (absent, foreign or owner-mismatched
		// definition) and the owner and target Group keys.
		"label_definitions_organization_fkey": CodeOrganizationNotFound,
		"label_definitions_group_fkey":        CodeGroupNotFound,
		"label_assignments_label_fkey":        CodeLabelNotFound,
		"label_assignments_target_group_fkey": CodeGroupNotFound,
	}
	checkCodes = map[string]Code{
		"organizations_name_valid":             CodeInvalidName,
		"groups_name_valid":                    CodeInvalidName,
		"groups_slug_format":                   CodeInvalidSlug,
		"groups_slug_reserved":                 CodeSlugReserved,
		"organizations_id_nonzero":             CodeInvalidID,
		"groups_id_nonzero":                    CodeInvalidID,
		"role_bindings_id_nonzero":             CodeInvalidID,
		"role_bindings_subject_nonzero":        CodeInvalidID,
		"role_bindings_scope_consistent":       CodeInvalidRole,
		"business_profiles_display_name_valid": CodeInvalidDisplayName,
		"business_profiles_legal_name_valid":   CodeInvalidLegalName,
		"business_profiles_description_valid":  CodeInvalidDescription,
		"business_profiles_industry_valid":     CodeInvalidIndustry,
		"business_profiles_website_valid":      CodeInvalidWebsite,
		"business_profiles_id_nonzero":         CodeInvalidID,
		"group_business_profiles_id_nonzero":   CodeInvalidID,
		"stored_objects_id_nonzero":            CodeInvalidID,
		"stored_objects_size_valid":            CodeObjectTooLarge,
		"stored_objects_content_type_valid":    CodeInvalidContentType,
		// G019 (00015_membership_invitations.sql).
		"invitations_id_nonzero":         CodeInvalidID,
		"invitations_scope_consistent":   CodeInvalidRole,
		"invitations_address_normalized": CodeInvalidAddress,
		"invitations_lifetime":           CodeInvalidLifetime,
		// G058 (00016_organization_enrollment.sql).
		"organizations_enrollment_mode_valid":          CodeInvalidEnrollmentMode,
		"organizations_open_requires_resolved_history": CodeEnrollmentNotReady,
		// G020 (00020_labels.sql): the name rule (ValidateLabelName is never weaker), the owner-scope shape, the
		// Organization-scope visibility rule, the closed production assignment shape and the nonzero core-generated ids.
		"label_definitions_name_valid":       CodeInvalidName,
		"label_definitions_scope_consistent": CodeInvalidScope,
		"label_definitions_visibility_valid": CodeVisibilityNotAllowed,
		"label_definitions_id_nonzero":       CodeInvalidID,
		"label_assignments_target_shape":     CodeTargetUnsupported,
		"label_assignments_id_nonzero":       CodeInvalidID,
	}
	// G019: 00015's guard and claim refusals raised as integrity_constraint_violation (23000) and as
	// insufficient_privilege (42501), keyed by the CONSTRAINT name of the RAISE. Any other 23000 or 42501 (an
	// immutable-key trigger of an earlier migration, an audit-event refusal, a missing grant) stays internal.
	integrityCodes = map[string]Code{
		"last_owner":            CodeLastOwner,
		"recipient_ineligible":  CodeRecipientIneligible,
		"dependent_bindings":    CodeDependentBindings,
		"issuer_authority_lost": CodeIssuerAuthorityLost,
		"invitation_limit":      CodeInvitationLimit,
		// The only runtime-reachable change it refuses is a state transition of a row that is no longer pending.
		"invitations_keys_immutable": CodeInvitationNotPending,
		// G020: the per-scope definition bound (app.label_definitions_bound: scope-row lock, then count; 200).
		"label_definitions_bound": CodeScopeFull,
	}
	privilegeCodes = map[string]Code{
		"owner_authority":  CodeOwnerAuthority,
		"context_required": CodeContextRequired,
		// The claim bodies' context-shape refusal: the same caller error as context_required.
		"invitation_context": CodeContextRequired,
		// G058: the enrollment bodies' IDENTITY-shape refusal (enrollment_request, enrollment_admit) and the request state
		// machine's refusal (enrollment_requests_immutable, which also enforces decider/revoker = ctx_user()).
		"enrollment_context":            CodeContextRequired,
		"enrollment_request_transition": CodeEnrollmentRequestTransition,
		// G060: the override guard (a change by anyone but the current Organization owner) and step (3b) of the
		// authority guard (an own Group role change).
		"group_override_owner": CodeOwnerAuthority,
		"self_escalation":      CodeSelfEscalation,
		// G060 amendment01: step (3) of the authority guard (an administrative departure of a Group-owner subject).
		"owner_subject": CodeOwnerSubject,
	}
	// G058: the raise_exception (P0001) refusals of 00016's bodies, keyed by the exact message (the bodies set no
	// constraint name). Any other P0001 stays internal. enrollment_request's one_pending fallback (40001
	// enrollment_request_conflict, raised when the conflicting pending row cannot be read) is not mapped here: it is
	// class 40 and stays unavailable like every serialization failure, so the access facade reports it as aborted (a
	// definite rollback, never success and never an unknown COMMIT outcome); no layer retries it internally.
	raisedCodes = map[string]Code{
		"enrollment_restricted": CodeEnrollmentRestricted,
	}
)

// classify maps a store error onto an Error. A tenancy Error passes through unchanged.
func classify(err error) *Error {
	if err == nil {
		return nil
	}
	var te *Error
	if errors.As(err, &te) {
		return te
	}
	if ce := contextCause(err); ce != nil {
		return &Error{Code: CodeUnavailable, cause: ce}
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return newError(pgCode(pg))
	}
	if transient(err) {
		return newError(CodeUnavailable)
	}
	return newError(CodeInternal)
}

// unavailable maps a failure to begin or commit a transaction: always unavailable, keeping a context cause.
func unavailable(err error) *Error {
	return &Error{Code: CodeUnavailable, cause: contextCause(err)}
}

func pgCode(pg *pgconn.PgError) Code {
	var m map[string]Code
	switch pg.Code {
	case "23505":
		m = uniqueCodes
	case "23503":
		m = foreignKeyCodes
	case "23514":
		m = checkCodes
	case "23000":
		if c, ok := integrityCodes[pg.ConstraintName]; ok {
			return c
		}
		return CodeInternal
	case "42501":
		if c, ok := privilegeCodes[pg.ConstraintName]; ok {
			return c
		}
		return CodeInternal
	case "P0001":
		if c, ok := raisedCodes[pg.Message]; ok {
			return c
		}
		return CodeInternal
	}
	if m != nil {
		if c, ok := m[pg.ConstraintName]; ok {
			return c
		}
		return CodeInternal
	}
	if len(pg.Code) == 5 {
		switch pg.Code[:2] {
		case "08", "40", "53", "57":
			return CodeUnavailable
		}
	}
	return CodeInternal
}

func contextCause(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	}
	return nil
}

// transient reports a connection-level failure without a server error: a refused or broken connection, a timeout
// or an error the driver marks as safe to retry because nothing was sent.
func transient(err error) bool {
	var ce *pgconn.ConnectError
	var ne net.Error
	return errors.As(err, &ce) || errors.As(err, &ne) || pgconn.Timeout(err) || pgconn.SafeToRetry(err) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed)
}
