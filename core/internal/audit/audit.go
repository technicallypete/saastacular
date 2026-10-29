// Package audit is the durable audit boundary of core (G011; ADR-006; docs/foundation/AUDIT.md): the closed event
// catalog, the one append primitive for core-owned events, the tenant reader and the revocation-gap marker read.
//
// Migration 00008_audit_events.sql owns the relation audit.events. Row changes of the eight audited app tables are
// recorded by database capture triggers in the same transaction; core appends only the events that are not row
// changes (authentication decisions, security denials, signed-URL issuance, external storage outcomes and revocation
// gaps), always on a transaction its caller owns. The database sets every server-owned field (id, schema version, time,
// actor and correlation) from the transaction context, refuses a core insert of a trigger-owned action and validates
// each action's shape, so Event carries the descriptive fields only.
//
// The package imports pgx only; identity and the access facade use it. It holds no names, slugs, emails, provider
// subjects, keys, URLs, payloads or driver text.
package audit

import (
	"slices"
)

// Action is one catalog action (§2 of the G011 Plan).
type Action string

// The 35 actions (G014 adds the two trigger-owned MCP connection actions; G016 the four trigger-owned domain and slug
// actions). T-actions (trigger-owned) are recorded only by
// the capture triggers; G-actions (core-owned) only by Append.
const (
	OrganizationCreated       Action = "organization.created"
	OrganizationRenamed       Action = "organization.renamed"
	OrganizationPolicyChanged Action = "organization.policy_changed"
	GroupCreated              Action = "group.created"
	GroupRenamed              Action = "group.renamed"
	GroupSlugChanged          Action = "group.slug_changed"
	MembershipGranted         Action = "membership.granted"
	MembershipRoleChanged     Action = "membership.role_changed"
	MembershipRemoved         Action = "membership.removed"
	BusinessProfileCreated    Action = "business_profile.created"
	BusinessProfileUpdated    Action = "business_profile.updated"
	BusinessProfileLinked     Action = "business_profile.linked"
	BusinessProfileUnlinked   Action = "business_profile.unlinked"
	UserCreated               Action = "user.created"
	UserDisabled              Action = "user.disabled"
	UserEnabled               Action = "user.enabled"
	SessionIssued             Action = "session.issued"
	SessionRevoked            Action = "session.revoked"
	SessionProviderRevocation Action = "session.provider_revocation_recorded"
	ObjectUploadStarted       Action = "object.upload_started"
	ObjectReady               Action = "object.ready"
	ObjectDeleteStarted       Action = "object.delete_started"
	ObjectDeleted             Action = "object.deleted"
	MCPConnectionEstablished  Action = "mcp.connection_established"
	MCPContextChanged         Action = "mcp.context_changed"
	OrganizationSlugChanged   Action = "organization.slug_changed"     // G016
	DomainClaimed             Action = "domain.claimed"                // G016
	DomainStateChanged        Action = "domain.state_changed"          // G016: reason = the new state
	DomainChecked             Action = "domain.checked"                // G016: reason = the check outcome
	BrandingChanged           Action = "organization.branding_changed" // G017: reason = updated, reset, logo_set, logo_removed
	// G019: the invitation lifecycle captures (audit.capture_invitations, 00015); never the digest or the address.
	InvitationIssued   Action = "invitation.issued"
	InvitationRevoked  Action = "invitation.revoked"
	InvitationConsumed Action = "invitation.consumed"
	// G058 slice 1b: the enrollment policy and readiness captures (00016 audit_capture_enrollment and
	// audit_capture_backfill); the readiness and backfill rows come only from the superuser operations channel.
	EnrollmentPolicyChanged       Action = "enrollment.policy_changed"
	EnrollmentReadinessChanged    Action = "enrollment.readiness_changed"      // reason = the basis or history_unresolved
	EnrollmentRestrictionBackfill Action = "enrollment.restriction_backfilled" // reason = operator_backfilled
	// G058 request slice: the access-request captures (00016 audit_capture_request_insert/_status); the request is the
	// resource, the applicant the subject; no reason (enrollment.decided does not say approved or denied: the request
	// row does).
	EnrollmentRequested Action = "enrollment.requested"
	EnrollmentDecided   Action = "enrollment.decided"
	EnrollmentRevoked   Action = "enrollment.revoked"
	// G063 (00019; Plan r1 D-1): a connection's friendly name was set, changed or cleared (audit_capture_name); the
	// connection is the resource, its owner the subject; never the name, a prior or new value or a reason.
	MCPConnectionRenamed Action = "mcp.connection_renamed"
	// G020 (00020; Plan r2 D-4): the label captures (audit_capture_label_definitions, audit_capture_label_assignments).
	// The definition is the resource, the owning Group (Group-owned definitions) or the target Group (assignments) the
	// group_id; never a name, a target type or a non-Group target id. label.updated carries reason renamed or
	// visibility_changed (one event per changed field); a no-op update records nothing.
	LabelCreated    Action = "label.created"
	LabelUpdated    Action = "label.updated"
	LabelDeleted    Action = "label.deleted"
	LabelAssigned   Action = "label.assigned"
	LabelUnassigned Action = "label.unassigned"

	AuthSignIn            Action = "auth.sign_in"
	AccessDenied          Action = "access.denied"
	ObjectURLIssued       Action = "object.url_issued"
	ObjectUploadFailed    Action = "object.upload_failed"
	ObjectDeletionPending Action = "object.deletion_pending"
	RevocationUnaudited   Action = "revocation.unaudited"
	// G060 (00017; Plan r1 C5): an ordinary Group administrative write allowed only through the Organization's
	// Group-management override, appended atomically on that write's transaction; it names the actual mutated Group or
	// Business Profile and the operation, never a reason, role, subject or token.
	AccessOverrideUsed Action = "access.override_used"
)

// TriggerActions are the 45 actions only a capture trigger records (G063 adds mcp.connection_renamed; G020 the five
// label.* actions).
var TriggerActions = []Action{
	OrganizationCreated, OrganizationRenamed, OrganizationPolicyChanged, GroupCreated, GroupRenamed, GroupSlugChanged,
	MembershipGranted, MembershipRoleChanged, MembershipRemoved, BusinessProfileCreated, BusinessProfileUpdated,
	BusinessProfileLinked, BusinessProfileUnlinked, UserCreated, UserDisabled, UserEnabled, SessionIssued, SessionRevoked,
	SessionProviderRevocation, ObjectUploadStarted, ObjectReady, ObjectDeleteStarted, ObjectDeleted, MCPConnectionEstablished,
	MCPContextChanged, OrganizationSlugChanged, DomainClaimed, DomainStateChanged, DomainChecked, BrandingChanged,
	InvitationIssued, InvitationRevoked, InvitationConsumed,
	EnrollmentPolicyChanged, EnrollmentReadinessChanged, EnrollmentRestrictionBackfill,
	EnrollmentRequested, EnrollmentDecided, EnrollmentRevoked, MCPConnectionRenamed,
	LabelCreated, LabelUpdated, LabelDeleted, LabelAssigned, LabelUnassigned,
}

// CoreActions are the 7 actions only Append records (G060 adds access.override_used).
var CoreActions = []Action{AuthSignIn, AccessDenied, ObjectURLIssued, ObjectUploadFailed, ObjectDeletionPending, RevocationUnaudited,
	AccessOverrideUsed}

// TriggerOwned reports whether a is recorded only by a capture trigger.
func TriggerOwned(a Action) bool { return slices.Contains(TriggerActions, a) }

// Outcome is an event's outcome.
type Outcome string

// The outcomes.
const (
	Success Outcome = "success"
	Denied  Outcome = "denied"
	Failed  Outcome = "failed"
	Unknown Outcome = "unknown"
)

// Outcomes is the closed outcome set.
var Outcomes = []Outcome{Success, Denied, Failed, Unknown}

// Reason is an event's closed reason.
type Reason string

// The reasons, grouped by the actions that use them.
const (
	// organization.policy_changed
	ReasonProfileEditEnabled  Reason = "group_admin_profile_edit_enabled"
	ReasonProfileEditDisabled Reason = "group_admin_profile_edit_disabled"
	// session.revoked and revocation.unaudited
	ReasonLogout         Reason = "logout"
	ReasonLogoutAll      Reason = "logout_all"
	ReasonProviderLogout Reason = "provider_logout"
	ReasonDisabled       Reason = "disabled"
	ReasonSuperseded     Reason = "superseded"
	// auth.sign_in
	ReasonSignedIn                  Reason = "signed_in"
	ReasonSessionRotated            Reason = "session_rotated"
	ReasonTransactionExpired        Reason = "transaction_expired"
	ReasonLoginCanceled             Reason = "login_canceled"
	ReasonCodeRejected              Reason = "code_rejected"
	ReasonImpersonationRefused      Reason = "impersonation_refused"
	ReasonMFAUnverified             Reason = "mfa_unverified"
	ReasonStaleAuthentication       Reason = "stale_authentication"
	ReasonIdentityConflict          Reason = "identity_conflict"
	ReasonUserDisabled              Reason = "user_disabled"
	ReasonProviderError             Reason = "provider_error"
	ReasonProviderTimeout           Reason = "provider_timeout"
	ReasonIdentityNamespaceMismatch Reason = "identity_namespace_mismatch"
	ReasonIdentityUnavailable       Reason = "identity_unavailable"
	// access.denied
	ReasonOrganizationNotPermitted Reason = "organization_not_permitted"
	ReasonGroupNotSelectable       Reason = "group_not_selectable"
	ReasonTargetNotFound           Reason = "target_not_found"
	ReasonTargetKindMismatch       Reason = "target_kind_mismatch"
	ReasonTargetOutsideContext     Reason = "target_outside_context"
	ReasonRoleNotPermitted         Reason = "role_not_permitted"
	ReasonNoLinkedPath             Reason = "no_linked_path"
	ReasonEditPolicyDisabled       Reason = "edit_policy_disabled"
	ReasonNotCreator               Reason = "not_creator"
	ReasonObjectNotReady           Reason = "object_not_ready"
	ReasonCapabilityUnknown        Reason = "capability_unknown"
	ReasonFactInconsistent         Reason = "fact_inconsistent"
	ReasonResourceMismatch         Reason = "resource_mismatch" // G014: operation mcp.request only, never with an Organization
	// object.upload_failed
	ReasonCompensated        Reason = "compensated"
	ReasonCompensationFailed Reason = "compensation_failed"
	ReasonNotCompensated     Reason = "not_compensated"
	// object.deletion_pending
	ReasonProviderUnavailable Reason = "provider_unavailable"
	ReasonTombstoneFailed     Reason = "tombstone_failed"
	ReasonTombstoneUnknown    Reason = "tombstone_unknown"
	// G016 auth.sign_in denials (the tenant-host issuance guard).
	ReasonOrganizationNotMember Reason = "organization_not_member"
	ReasonHostChanged           Reason = "host_changed"
	// G016 domain.state_changed (the new state) and domain.checked (the outcome).
	ReasonDomainVerified       Reason = "verified"
	ReasonDomainActive         Reason = "active"
	ReasonDomainSuspended      Reason = "suspended"
	ReasonDomainRemoved        Reason = "removed"
	ReasonCheckOK              Reason = "ok"
	ReasonCheckMismatch        Reason = "mismatch"
	ReasonCheckNXDomain        Reason = "nxdomain"
	ReasonCheckTransient       Reason = "transient"
	ReasonCheckTLSError        Reason = "tls_error"
	ReasonCheckWithdrawn       Reason = "withdrawn"
	ReasonCheckWithdrawError   Reason = "withdraw_error"
	ReasonCheckWithdrawUnknown Reason = "withdraw_unknown"
	// G017 organization.branding_changed (derived by the capture trigger from the old and new rows).
	ReasonBrandingUpdated     Reason = "updated"
	ReasonBrandingReset       Reason = "reset"
	ReasonBrandingLogoSet     Reason = "logo_set"
	ReasonBrandingLogoRemoved Reason = "logo_removed"
	// G058 access.denied enrollment refusals (00016 widens events_reason_valid and admits them in events_attribute).
	ReasonEnrollmentClosed     Reason = "enrollment_closed"
	ReasonEnrollmentRestricted Reason = "enrollment_restricted"
	ReasonApprovalRevoked      Reason = "approval_revoked"
	ReasonInvitationMismatch   Reason = "invitation_mismatch"
	// G058 Codex round-2 audit decision: the consumed recipient-proof failure (auth.sign_in denied), the admission
	// recipient refusal and the missing/revoked/expired enrollment credential aggregate (access.denied enrollment.admit).
	ReasonRecipientProofUnavailable Reason = "recipient_proof_unavailable"
	ReasonRecipientIneligible       Reason = "recipient_ineligible"
	ReasonSessionInvalid            Reason = "session_invalid"
	// G058 slice 1b: the readiness reasons of enrollment.readiness_changed / enrollment.restriction_backfilled.
	ReasonOperatorVerifiedAbsent Reason = "operator_verified_absent"
	ReasonOperatorBackfilled     Reason = "operator_backfilled"
	ReasonHistoryUnresolved      Reason = "history_unresolved"
	// G060 organization.policy_changed: the Group-management override was enabled or disabled (00017).
	ReasonOverrideEnabled  Reason = "group_management_override_enabled"
	ReasonOverrideDisabled Reason = "group_management_override_disabled"
	// G062 (00018, Plan r5 §3.6): the auth.sign_in successes of the root signup callback (T1) and of the pinned tenant
	// callback, the two signup sign-in denials, and the access.denied signup refusals (host_changed is reused).
	ReasonSignupAuthenticated   Reason = "signup_authenticated"
	ReasonTenantCreated         Reason = "tenant_created"
	ReasonSignupOriginRejected  Reason = "signup_origin_rejected"
	ReasonSignupUserMismatch    Reason = "signup_user_mismatch"
	ReasonSignupInvalid         Reason = "signup_invalid"
	ReasonIntentUsed            Reason = "intent_used"
	ReasonSlugTaken             Reason = "slug_taken"
	ReasonContinuationInvalid   Reason = "continuation_invalid"
	ReasonSignupContextMismatch Reason = "signup_context_mismatch"
	// G020 (00020): the two label.updated reasons.
	ReasonLabelRenamed           Reason = "renamed"
	ReasonLabelVisibilityChanged Reason = "visibility_changed"
)

// Reasons is the closed reason set (81).
var Reasons = []Reason{
	ReasonProfileEditEnabled, ReasonProfileEditDisabled,
	ReasonLogout, ReasonLogoutAll, ReasonProviderLogout, ReasonDisabled, ReasonSuperseded,
	ReasonSignedIn, ReasonSessionRotated, ReasonTransactionExpired, ReasonLoginCanceled, ReasonCodeRejected,
	ReasonImpersonationRefused, ReasonMFAUnverified, ReasonStaleAuthentication, ReasonIdentityConflict, ReasonUserDisabled,
	ReasonProviderError, ReasonProviderTimeout, ReasonIdentityNamespaceMismatch, ReasonIdentityUnavailable,
	ReasonOrganizationNotPermitted, ReasonGroupNotSelectable, ReasonTargetNotFound, ReasonTargetKindMismatch,
	ReasonTargetOutsideContext, ReasonRoleNotPermitted, ReasonNoLinkedPath, ReasonEditPolicyDisabled, ReasonNotCreator,
	ReasonObjectNotReady, ReasonCapabilityUnknown, ReasonFactInconsistent, ReasonResourceMismatch,
	ReasonCompensated, ReasonCompensationFailed, ReasonNotCompensated,
	ReasonProviderUnavailable, ReasonTombstoneFailed, ReasonTombstoneUnknown,
	ReasonOrganizationNotMember, ReasonHostChanged,
	ReasonDomainVerified, ReasonDomainActive, ReasonDomainSuspended, ReasonDomainRemoved, ReasonCheckOK, ReasonCheckMismatch,
	ReasonCheckNXDomain, ReasonCheckTransient, ReasonCheckTLSError, ReasonCheckWithdrawn, ReasonCheckWithdrawError,
	ReasonCheckWithdrawUnknown,
	ReasonBrandingUpdated, ReasonBrandingReset, ReasonBrandingLogoSet, ReasonBrandingLogoRemoved,
	ReasonEnrollmentClosed, ReasonEnrollmentRestricted, ReasonApprovalRevoked, ReasonInvitationMismatch,
	ReasonRecipientProofUnavailable, ReasonRecipientIneligible, ReasonSessionInvalid,
	ReasonOperatorVerifiedAbsent, ReasonOperatorBackfilled, ReasonHistoryUnresolved,
	ReasonOverrideEnabled, ReasonOverrideDisabled,
	ReasonSignupAuthenticated, ReasonTenantCreated, ReasonSignupOriginRejected, ReasonSignupUserMismatch,
	ReasonSignupInvalid, ReasonIntentUsed, ReasonSlugTaken, ReasonContinuationInvalid, ReasonSignupContextMismatch,
	ReasonLabelRenamed, ReasonLabelVisibilityChanged,
}

// Operation is the guarded operation of an access.denied event.
type Operation string

// The 36 operations of access.denied.
const (
	OpAccessResolve         Operation = "access.resolve"
	OpOrganizationGet       Operation = "organization.get"
	OpOrganizationGetPolicy Operation = "organization.get_policy"
	OpOrganizationSetPolicy Operation = "organization.set_policy"
	OpGroupGet              Operation = "group.get"
	OpGroupRename           Operation = "group.rename"
	OpGroupChangeSlug       Operation = "group.change_slug"
	OpBusinessProfileGet    Operation = "business_profile.get"
	OpBusinessProfileUpdate Operation = "business_profile.update"
	OpObjectPut             Operation = "object.put"
	OpObjectGet             Operation = "object.get"
	OpObjectMetadata        Operation = "object.metadata"
	OpObjectSign            Operation = "object.sign"
	OpObjectDelete          Operation = "object.delete"
	OpAuditRead             Operation = "audit.read"
	OpMCPRequest            Operation = "mcp.request" // G014: a protected MCP call
	// G016: the guarded domain operations and the Organization slug.
	OpDomainList            Operation = "domain.list"
	OpDomainClaim           Operation = "domain.claim"
	OpDomainRotate          Operation = "domain.rotate"
	OpDomainVerify          Operation = "domain.verify"
	OpDomainRecheck         Operation = "domain.recheck"
	OpDomainRemove          Operation = "domain.remove"
	OpDomainRetryWithdrawal Operation = "domain.retry_withdrawal"
	OpOrganizationSetSlug   Operation = "organization.set_slug"
	// G017: the guarded branding operations.
	OpBrandingGet        Operation = "branding.get"
	OpBrandingUpdate     Operation = "branding.update"
	OpBrandingReset      Operation = "branding.reset"
	OpBrandingSetLogo    Operation = "branding.set_logo"
	OpBrandingRemoveLogo Operation = "branding.remove_logo"
	// G018: the guarded Group onboarding operations (00014 widens events_operation_valid).
	OpGroupSuggestSlug                Operation = "group.suggest_slug"
	OpGroupCreate                     Operation = "group.create"
	OpGroupList                       Operation = "group.list"
	OpGroupOnboardingGet              Operation = "group.onboarding.get"
	OpGroupOnboardingCompleteNew      Operation = "group.onboarding.complete_new"
	OpGroupOnboardingCompleteExisting Operation = "group.onboarding.complete_existing"
	OpBusinessProfileList             Operation = "business_profile.list"
	// G019: the membership, ownership, invitation and retirement operations (00015 widens events_operation_valid).
	OpMembershipList        Operation = "membership.list"
	OpMembershipSelf        Operation = "membership.self"
	OpMembershipChangeRole  Operation = "membership.change_role"
	OpMembershipRemove      Operation = "membership.remove"
	OpMembershipLeave       Operation = "membership.leave"
	OpOwnershipPromote      Operation = "ownership.promote"
	OpOwnershipDemote       Operation = "ownership.demote"
	OpOwnershipTransfer     Operation = "ownership.transfer"
	OpInvitationList        Operation = "invitation.list"
	OpInvitationMine        Operation = "invitation.mine"
	OpInvitationCandidates  Operation = "invitation.candidates"
	OpInvitationIssue       Operation = "invitation.issue"
	OpInvitationRevoke      Operation = "invitation.revoke"
	OpInvitationGet         Operation = "invitation.get"
	OpInvitationAccept      Operation = "invitation.accept"
	OpInvitationConsume     Operation = "invitation.consume"
	OpRetirementEligibility Operation = "retirement.eligibility"
	// G058: the enrollment operations (00016 widens events_operation_valid). enrollment.start is the callback T1
	// enrollment-policy/intent denial (Codex round-3 decision 1).
	OpEnrollmentStart          Operation = "enrollment.start"
	OpEnrollmentPublic         Operation = "enrollment.public"
	OpEnrollmentStatus         Operation = "enrollment.status"
	OpEnrollmentRequest        Operation = "enrollment.request"
	OpEnrollmentAdmit          Operation = "enrollment.admit"
	OpEnrollmentAbandon        Operation = "enrollment.abandon"
	OpEnrollmentPolicyGet      Operation = "enrollment.policy.get"
	OpEnrollmentPolicyUpdate   Operation = "enrollment.policy.update"
	OpEnrollmentRequestsList   Operation = "enrollment.requests.list"
	OpEnrollmentRequestsDecide Operation = "enrollment.requests.decide"
	OpEnrollmentRequestsRevoke Operation = "enrollment.requests.revoke"
	// G060: the owner-only Group-management override write (00017 widens events_operation_valid).
	OpOrganizationSetGroupOverride Operation = "organization.set_group_override"
	// G062: the root signup and tenant continuation operations (00018 widens events_operation_valid).
	OpSignupStart     Operation = "signup.start"
	OpSignupStatus    Operation = "signup.status"
	OpSignupSlug      Operation = "signup.slug"
	OpSignupCreate    Operation = "signup.create"
	OpSignupChallenge Operation = "signup.challenge"
	OpSignupProve     Operation = "signup.prove"
	OpSignupComplete  Operation = "signup.complete"
	OpSignupAbandon   Operation = "signup.abandon"
	// G020: the guarded label operations (00020 widens events_operation_valid).
	OpLabelDefinitionList   Operation = "label.definition.list"
	OpLabelDefinitionCreate Operation = "label.definition.create"
	OpLabelDefinitionUpdate Operation = "label.definition.update"
	OpLabelDefinitionDelete Operation = "label.definition.delete"
	OpLabelAssignmentList   Operation = "label.assignment.list"
	OpLabelAssign           Operation = "label.assign"
	OpLabelUnassign         Operation = "label.unassign"
)

// Operations is the closed operation set (80).
var Operations = []Operation{
	OpAccessResolve, OpOrganizationGet, OpOrganizationGetPolicy, OpOrganizationSetPolicy, OpGroupGet, OpGroupRename,
	OpGroupChangeSlug, OpBusinessProfileGet, OpBusinessProfileUpdate, OpObjectPut, OpObjectGet, OpObjectMetadata,
	OpObjectSign, OpObjectDelete, OpAuditRead, OpMCPRequest,
	OpDomainList, OpDomainClaim, OpDomainRotate, OpDomainVerify, OpDomainRecheck, OpDomainRemove, OpDomainRetryWithdrawal,
	OpOrganizationSetSlug,
	OpBrandingGet, OpBrandingUpdate, OpBrandingReset, OpBrandingSetLogo, OpBrandingRemoveLogo,
	OpGroupSuggestSlug, OpGroupCreate, OpGroupList, OpGroupOnboardingGet, OpGroupOnboardingCompleteNew,
	OpGroupOnboardingCompleteExisting, OpBusinessProfileList,
	OpMembershipList, OpMembershipSelf, OpMembershipChangeRole, OpMembershipRemove, OpMembershipLeave, OpOwnershipPromote,
	OpOwnershipDemote, OpOwnershipTransfer, OpInvitationList, OpInvitationMine, OpInvitationCandidates, OpInvitationIssue,
	OpInvitationRevoke, OpInvitationGet, OpInvitationAccept, OpInvitationConsume, OpRetirementEligibility,
	OpEnrollmentStart, OpEnrollmentPublic, OpEnrollmentStatus, OpEnrollmentRequest, OpEnrollmentAdmit, OpEnrollmentAbandon,
	OpEnrollmentPolicyGet, OpEnrollmentPolicyUpdate, OpEnrollmentRequestsList, OpEnrollmentRequestsDecide,
	OpEnrollmentRequestsRevoke,
	OpOrganizationSetGroupOverride,
	OpSignupStart, OpSignupStatus, OpSignupSlug, OpSignupCreate, OpSignupChallenge, OpSignupProve, OpSignupComplete,
	OpSignupAbandon,
	OpLabelDefinitionList, OpLabelDefinitionCreate, OpLabelDefinitionUpdate, OpLabelDefinitionDelete, OpLabelAssignmentList,
	OpLabelAssign, OpLabelUnassign,
}

// ResourceType is the kind of an event's resource.
type ResourceType string

// The resource types.
const (
	ResourceOrganization    ResourceType = "organization"
	ResourceGroup           ResourceType = "group"
	ResourceRoleBinding     ResourceType = "role_binding"
	ResourceBusinessProfile ResourceType = "business_profile"
	ResourceUser            ResourceType = "user"
	ResourceSession         ResourceType = "session"
	ResourceStoredObject    ResourceType = "stored_object"
	ResourceMCPConnection   ResourceType = "mcp_connection"
	ResourceDomain          ResourceType = "organization_domain" // G016
	ResourceInvitation      ResourceType = "invitation"          // G019
	ResourceEnrollmentReq   ResourceType = "enrollment_request"  // G058
	ResourceLabel           ResourceType = "label"               // G020: the label definition
)

// ResourceTypes is the closed resource-type set (12).
var ResourceTypes = []ResourceType{ResourceOrganization, ResourceGroup, ResourceRoleBinding, ResourceBusinessProfile,
	ResourceUser, ResourceSession, ResourceStoredObject, ResourceMCPConnection, ResourceDomain, ResourceInvitation,
	ResourceEnrollmentReq, ResourceLabel}

// The actor kinds the database derives (never supplied).
const (
	ActorUser    = "user"
	ActorSystem  = "system"
	ActorUnknown = "unknown"
)

// The input codes of the guarded audit read (the access facade's invalid_input Field).
const (
	CodeInvalidPageSize  = "invalid_page_size"
	CodeInvalidTimeRange = "invalid_time_range"
	CodeInvalidCursor    = "invalid_cursor"
)

// UUID is a 16-byte identifier; the zero value means absent.
type UUID [16]byte

// Event is one core-owned event: the descriptive fields only. Zero values mean absent (NULL). Organization and Group are
// accepted by the database only when the Organization is the established context of the recording transaction and the
// Group belongs to it.
type Event struct {
	Action       Action
	Outcome      Outcome
	Reason       Reason
	Operation    Operation
	ResourceType ResourceType
	ResourceID   UUID
	Subject      UUID
	Organization UUID
	Group        UUID
	TTLSeconds   int
}

// ParseUUID parses a canonical lowercase hyphenated UUID; false for anything else.
func ParseUUID(s string) (UUID, bool) {
	var u UUID
	if len(s) != 36 {
		return u, false
	}
	j := 0
	for i := 0; i < 36; {
		switch i {
		case 8, 13, 18, 23:
			if s[i] != '-' {
				return UUID{}, false
			}
			i++
			continue
		}
		hi, ok1 := hexVal(s[i])
		lo, ok2 := hexVal(s[i+1])
		if !ok1 || !ok2 {
			return UUID{}, false
		}
		u[j] = hi<<4 | lo
		j++
		i += 2
	}
	return u, true
}

func hexVal(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	}
	return 0, false
}
