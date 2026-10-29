package access

import (
	"context"
	"errors"
	"time"

	"core/internal/identity"
	"core/internal/tenancy"
)

// G058 Organization enrollment admission (Plan r3 §1.3 step 3, §1.4, §5.1; readiness correction-01 §1 rows 3 and 6;
// Codex round-2 host and audit/log decisions). Admission runs without a tenant context or session: the subject is the
// canonical user of a resolved enrollment credential (identity.EnrollmentHandle), and the Organization is the one that
// credential was minted for. Each admission is ONE IDENTITY-shape transaction of that user: first identity's host guard
// (CheckEnrollmentHost, on this transaction, before any claim or admission write), then app.enrollment_admit once
// (after G019's Organization invitation claim for the invitation kind); only admitted and already_member commit, every
// refusal rolls the whole transaction back (the claim included). Known denials (target_not_found, enrollment_closed,
// enrollment_restricted, recipient_ineligible, session_invalid, host_changed) are recorded as one access.denied
// enrollment.admit event without an Organization, attributed to the established user, in the independent audit
// transaction after the rollback (recordDenial); the access.decision record names the same semantic reason.

// EnrollmentAdmission is a committed admission: Outcome admitted (BindingID is the new or reinstated Organization
// membership binding) or already_member (no write; BindingID is the reconciled binding when the body returns it).
type EnrollmentAdmission struct {
	Outcome   string
	BindingID tenancy.BindingID
}

// enrollmentIdentity is the identity boundary of the admission host guard; *identity.Service satisfies it (the existing
// trusted-service composition pattern, as mcpIdentity and retirementFinalizer). A Service whose identity boundary does
// not provide it refuses every admission as unavailable (fail closed).
type enrollmentIdentity interface {
	CheckEnrollmentHost(ctx context.Context, q identity.Querier, h identity.EnrollmentHandle) error
}

// checkEnrollmentHost runs identity's host guard on the admission transaction tx: host_changed (conflict) when the
// credential's origin no longer routes to its Organization; unavailable on a resolver failure or when the identity
// boundary does not provide the guard (a nil identity service fails inside the guard itself).
func (s *Service) checkEnrollmentHost(ctx context.Context, tx *trackedTx, h identity.EnrollmentHandle) *failure {
	ei, ok := s.id.(enrollmentIdentity)
	if !ok {
		return fail(CodeUnavailable, reasonStoreUnavailable)
	}
	switch err := ei.CheckEnrollmentHost(ctx, tx, h); {
	case err == nil:
		return nil
	case errors.Is(err, identity.ErrHostChanged):
		return &failure{code: CodeConflict, field: fieldHostChanged, reason: reasonHostChanged}
	case tx.cause != nil:
		return fail(CodeUnavailable, failureReason(ctx, tx.cause))
	}
	return fail(CodeUnavailable, reasonStoreUnavailable)
}

// enrollmentHandle validates a resolved enrollment handle before any statement: a canonical non-zero user and
// Organization and a non-zero credential digest. Anything else is invalid_input (enrollment).
func enrollmentHandle(h identity.EnrollmentHandle) (tenancy.UserID, tenancy.OrganizationID, *failure) {
	user, err := tenancy.ParseUserID(h.UserID)
	if err != nil || zero(user) || user.String() != h.UserID {
		return tenancy.UserID{}, tenancy.OrganizationID{}, invalid(fieldEnrollment)
	}
	org, err := tenancy.ParseOrganizationID(h.OrganizationID)
	if err != nil || zero(org) || org.String() != h.OrganizationID || h.Digest == ([32]byte{}) {
		return tenancy.UserID{}, tenancy.OrganizationID{}, invalid(fieldEnrollment)
	}
	return user, org, nil
}

// admitRefusal maps a refusal row of app.enrollment_admit onto its failure (nil for admitted and already_member), each
// with its semantic reason: enrollment_closed and enrollment_restricted are conflict with the outcome as field;
// recipient_ineligible is G019's conflict; enrollment_expired is unauthenticated with session_invalid (the missing,
// revoked or expired credential aggregate; no disabled user or expired login is claimed); not_found is the closed
// not_found. The request kind: approval_revoked is the recorded conflict with its semantic reason; pending and denied
// are status conflicts (write_conflict; never recorded). Any other value is internal.
func admitRefusal(o tenancy.AdmitOutcome) *failure {
	switch o {
	case tenancy.AdmitAdmitted, tenancy.AdmitAlreadyMember:
		return nil
	case tenancy.AdmitApprovalRevoked:
		return &failure{code: CodeConflict, field: fieldApprovalRevoked, reason: reasonApprovalRevoked}
	case tenancy.AdmitPending:
		return conflict(fieldPending)
	case tenancy.AdmitDenied:
		return conflict(fieldDenied)
	case tenancy.AdmitEnrollmentClosed:
		return &failure{code: CodeConflict, field: fieldEnrollmentClosed, reason: reasonEnrollmentClosed}
	case tenancy.AdmitEnrollmentRestricted:
		return &failure{code: CodeConflict, field: fieldEnrollmentRestricted, reason: reasonEnrollmentRestricted}
	case tenancy.AdmitRecipientIneligible:
		return &failure{code: CodeConflict, field: string(tenancy.CodeRecipientIneligible), reason: reasonRecipientIneligible}
	case tenancy.AdmitEnrollmentExpired:
		return &failure{code: CodeUnauthenticated, field: fieldEnrollmentExpired, reason: reasonSessionInvalid}
	case tenancy.AdmitNotFound:
		return fail(CodeNotFound, reasonTargetNotFound)
	}
	return fail(CodeInternal, reasonStoreUnavailable)
}

// admitFailure maps an error of tenancy.AdmitEnrollmentTx: the lift kind's newer-restriction RAISE is the
// enrollment_restricted conflict (the transaction is aborted and rolls back with the claim); a store failure is
// unavailable with the transaction's reason; anything else is internal.
func admitFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	switch tenancy.CodeOf(err) {
	case tenancy.CodeEnrollmentRestricted:
		return &failure{code: CodeConflict, field: fieldEnrollmentRestricted, reason: reasonEnrollmentRestricted}
	case tenancy.CodeUnavailable:
		return fail(CodeUnavailable, failureReason(ctx, tx.cause))
	}
	return fail(CodeInternal, reasonStoreUnavailable)
}

// enrollmentClaimFailure is G019's claim failure (claimFailure) mapped onto the enrollment closed set (Codex round-3
// decision 4): a disabled recipient (recipient_ineligible) keeps its conflict with the semantic reason; an invitation
// whose issuer no longer holds authority (issuer_authority_lost) is the closed not_found (target_not_found), exactly as
// a missing, foreign, revoked or expired invitation. Every other claim failure (invalid input, an actual class-40 or
// store failure) is G019's unchanged; G019's own callers keep their mapping (claimFailure is not changed).
func enrollmentClaimFailure(f *failure) *failure {
	switch {
	case f == nil || f.code != CodeConflict:
		return f
	case f.field == string(tenancy.CodeRecipientIneligible):
		return &failure{code: CodeConflict, field: f.field, reason: reasonRecipientIneligible}
	case f.field == string(tenancy.CodeIssuerAuthorityLost):
		return fail(CodeNotFound, reasonTargetNotFound)
	}
	return f
}

// admissionDenialReason is the access.denied reason of a known admission denial, or "" when the failure is not
// recorded (invalid input, a consumed replay, another G019 claim conflict, unavailable, internal).
func admissionDenialReason(f *failure) Reason {
	switch {
	case f == nil:
		return ""
	case f.code == CodeNotFound:
		return reasonTargetNotFound
	case f.code == CodeUnauthenticated && f.field == fieldEnrollmentExpired:
		return reasonSessionInvalid
	case f.code == CodeConflict && (f.reason == reasonEnrollmentClosed || f.reason == reasonEnrollmentRestricted ||
		f.reason == reasonRecipientIneligible || f.reason == reasonHostChanged || f.reason == reasonApprovalRevoked):
		return f.reason
	}
	return ""
}

// finishAdmission records a known denial of user's admission as one access.denied event without an Organization (G019's
// NULL-context shape; the independent audit transaction runs after the admission transaction rolled back and released
// its connection), then the access.decision record with the same semantic reason. A failure before the user was
// established is never recorded.
func (s *Service) finishAdmission(ctx context.Context, user tenancy.UserID, f *failure) {
	s.finishEnrollment(ctx, opEnrollmentAdmit, user, f)
}

// finishEnrollment is finishAdmission for one own-credential enrollment operation op: enrollment.admit and
// enrollment.request record their known denials (the request's set is the admission's minus the restriction and
// approval reasons, which its body never produces); enrollment.status and enrollment.abandon record none (a read and an
// idempotent self-revocation), only the access.decision record.
func (s *Service) finishEnrollment(ctx context.Context, op string, user tenancy.UserID, f *failure) {
	if r := admissionDenialReason(f); r != "" && !zero(user) && (op == opEnrollmentAdmit || op == opEnrollmentRequest) {
		s.recordDenial(ctx, user, tenancy.OrganizationID{}, op, &failure{code: f.code, field: f.field, reason: r, denied: &denialScope{}})
	}
	s.logDecision(ctx, op, noCapability, outcomeAllow, reasonAllowed, f)
}

// enrollmentTx runs one own-credential enrollment operation op of h's user: the handle is validated before any
// statement, then fn runs in one IDENTITY-shape transaction of that user (no tenant context, no session; write or
// read-only), and the call ends with finishEnrollment. Every failure of fn rolls the whole transaction back.
func (s *Service) enrollmentTx(ctx context.Context, op string, write bool, h identity.EnrollmentHandle,
	fn func(tx *trackedTx, user tenancy.UserID, org tenancy.OrganizationID) *failure) *failure {
	user, org, f := enrollmentHandle(h)
	if f == nil {
		f = s.inContextTx(ctx, write, dbContext{user: user}, cleanupUnbounded, func(tx *trackedTx) *failure {
			return fn(tx, user, org)
		})
	}
	s.finishEnrollment(ctx, op, user, f)
	return f
}

// admitInvitation admits h's user to h's Organization through an Organization invitation (Plan r3 §5.1 with the
// readiness DTO reconciliation): ev must be the identity-attested evidence of the same canonical user. A zero ev, a
// malformed handle, a user mismatch or a malformed invitation id or token is refused before any statement. Then one
// IDENTITY-shape transaction runs identity's host guard, G019's claim (app.invitation_claim_organization, refusals
// exactly as ConsumeOrganizationInvitation) and, for an invitation of h's Organization, app.enrollment_admit(digest,
// org, 'lift', invitation); a host change, any refusal, a claim of another Organization (not_found) or the lift RAISE
// (enrollment_restricted) rolls back the claim with everything else (the invitation stays pending).
func (s *Service) admitInvitation(ctx context.Context, ev identity.VerifiedAddress, h identity.EnrollmentHandle, id tenancy.InvitationID,
	token string) (tenancy.AdmitResult, error) {
	var out tenancy.AdmitResult
	var user tenancy.UserID
	f := func() *failure {
		if ev.IsZero() {
			return invalid(fieldEvidence)
		}
		u, org, fl := enrollmentHandle(h)
		if fl != nil {
			return fl
		}
		if ev.UserID() != h.UserID {
			return invalid(fieldEvidence)
		}
		if fl := tokenPre(id, token)(); fl != nil {
			return fl
		}
		user = u
		claim := &tenancy.OrganizationClaim{Issuer: ev.Issuer(), Subject: ev.Subject(), Address: ev.Address()}
		return s.inContextTx(ctx, true, dbContext{user: user}, cleanupUnbounded, func(tx *trackedTx) *failure {
			if fl := s.checkEnrollmentHost(ctx, tx, h); fl != nil {
				return fl
			}
			c, err := tenancy.ClaimInvitationTx(ctx, tx, id, token, claim)
			if err != nil {
				if consumed, fl := claimFailure(ctx, tx, err); !consumed {
					return enrollmentClaimFailure(fl)
				}
				// A consumed replay without a qualifying binding admits nothing (nothing is restored).
				return conflict(string(tenancy.CodeInvitationNotPending))
			}
			if c.OrganizationID != org || c.ScopeType != tenancy.ScopeOrganization {
				return fail(CodeNotFound, reasonTargetNotFound)
			}
			r, err := tenancy.AdmitEnrollmentTx(ctx, tx, h.Digest, org, tenancy.AdmitKindLift, &id)
			if err != nil {
				return admitFailure(ctx, tx, err)
			}
			if fl := admitRefusal(r.Outcome); fl != nil {
				return fl
			}
			if r.Outcome == tenancy.AdmitAdmitted && r.Binding != c.BindingID {
				return fail(CodeInternal, reasonStoreUnavailable) // the admitted binding is the claimed one
			}
			out = r
			return nil
		})
	}()
	s.finishAdmission(ctx, user, f)
	if f != nil {
		return tenancy.AdmitResult{}, f.err()
	}
	return out, nil
}

// AdmitInvitationAtCallback is the callback-time invitation admission (admitInvitation) with its access-owned result.
func (s *Service) AdmitInvitationAtCallback(ctx context.Context, ev identity.VerifiedAddress, h identity.EnrollmentHandle,
	id tenancy.InvitationID, token string) (EnrollmentAdmission, error) {
	r, err := s.admitInvitation(ctx, ev, h, id, token)
	if err != nil {
		return EnrollmentAdmission{}, err
	}
	return EnrollmentAdmission{Outcome: string(r.Outcome), BindingID: r.Binding}, nil
}

// AdmitOpen admits h's user to h's Organization through open enrollment: one IDENTITY-shape transaction running
// identity's host guard, then app.enrollment_admit(digest, org, 'open', NULL), which rechecks the mode
// (enrollment_closed unless open), the user, the credential and the restriction ledger (enrollment_restricted while any
// restriction is unlifted). A malformed handle is refused before any statement; every refusal rolls back.
func (s *Service) AdmitOpen(ctx context.Context, h identity.EnrollmentHandle) (EnrollmentAdmission, error) {
	return s.admitOwn(ctx, h, func(tx *trackedTx, org tenancy.OrganizationID) (tenancy.AdmitResult, error) {
		return tenancy.AdmitEnrollmentTx(ctx, tx, h.Digest, org, tenancy.AdmitKindOpen, nil)
	})
}

// AdmitRequest admits h's user to h's Organization through the user's own approved request id (Plan r3 §1.4 "approved
// request"): one IDENTITY-shape transaction running identity's host guard, then app.enrollment_admit(digest, org,
// 'request', id), which locks the request row (serializing with a revocation and a second consumer), rechecks the mode
// (request_access or open), the approval (approved, unconsumed, unexpired, its decider still org_owner or org_admin) and
// the restriction ledger (an unlifted restriction newer than the approval's authority_seq refuses with zero writes),
// then writes the binding, the consumption and the lift of the covered restrictions. A zero request id or a malformed
// handle is refused before any statement; every refusal rolls back. Refusals: pending and denied (status conflicts),
// approval_revoked (recorded), enrollment_restricted, enrollment_closed, not_found (also a request of another subject or
// Organization), recipient_ineligible, enrollment_expired, host_changed.
func (s *Service) AdmitRequest(ctx context.Context, h identity.EnrollmentHandle, id tenancy.EnrollmentRequestID) (EnrollmentAdmission, error) {
	if zero(id) {
		f := invalid(fieldEnrollment)
		s.finishAdmission(ctx, tenancy.UserID{}, f)
		return EnrollmentAdmission{}, f.err()
	}
	return s.admitOwn(ctx, h, func(tx *trackedTx, org tenancy.OrganizationID) (tenancy.AdmitResult, error) {
		return tenancy.AdmitRequestTx(ctx, tx, h.Digest, org, id)
	})
}

// admitOwn is the shared body of the open and request admissions: enrollmentTx (enrollment.admit, write) running the
// host guard, then the admission call, mapped through admitFailure and admitRefusal.
func (s *Service) admitOwn(ctx context.Context, h identity.EnrollmentHandle,
	call func(tx *trackedTx, org tenancy.OrganizationID) (tenancy.AdmitResult, error)) (EnrollmentAdmission, error) {
	var out tenancy.AdmitResult
	f := s.enrollmentTx(ctx, opEnrollmentAdmit, true, h, func(tx *trackedTx, _ tenancy.UserID, org tenancy.OrganizationID) *failure {
		if fl := s.checkEnrollmentHost(ctx, tx, h); fl != nil {
			return fl
		}
		r, err := call(tx, org)
		if err != nil {
			return admitFailure(ctx, tx, err)
		}
		if fl := admitRefusal(r.Outcome); fl != nil {
			return fl
		}
		out = r
		return nil
	})
	if f != nil {
		return EnrollmentAdmission{}, f.err()
	}
	return EnrollmentAdmission{Outcome: string(out.Outcome), BindingID: out.Binding}, nil
}

// G058 request slice: access requests (Plan r3 §1.4/§1.5; Codex round-8 dispositions 1-4 and the request-c1 rulings).
// The applicant operations (RequestEnrollment, EnrollmentStatus, AbandonEnrollment; AdmitRequest above) run on h's own
// credential in IDENTITY shape; the manager operations (ListEnrollmentRequests, DecideEnrollmentRequest,
// RevokeEnrollmentApproval) run through the guarded facade with organization.manage_members in the active
// Organization, the two writes with the Organization row locked FOR NO KEY UPDATE before the request row (Plan r3 §1.4
// decision/revocation order (3') -> (4)).

// EnrollmentRequestView is one access request as a reader sees it. Status is the stored state (pending, approved,
// denied, revoked); Expired is derived from ExpiresAt (an expired pending request keeps its slot, is never consumable
// and is resolved by an administrator; Codex round-8 disposition 3); Consumed reports an approval spent by its
// admission. DisplayName is always nil: the existing users row-level security does not expose a non-member applicant's
// profile, and no lookup or widened read is added (disposition 4); consumers show the identifiers. No decider,
// revoker, binding or authority value is exposed.
type EnrollmentRequestView struct {
	ID          tenancy.EnrollmentRequestID
	UserID      tenancy.UserID
	DisplayName *string
	Status      string
	Expired     bool
	Consumed    bool
	CreatedAt   time.Time
	ExpiresAt   time.Time
	DecidedAt   *time.Time
	RevokedAt   *time.Time
}

func requestView(r tenancy.EnrollmentRequest, now time.Time) EnrollmentRequestView {
	return EnrollmentRequestView{ID: r.ID, UserID: r.UserID, Status: r.Status, Expired: r.Expired(now), Consumed: r.ConsumedAt != nil,
		CreatedAt: r.CreatedAt, ExpiresAt: r.ExpiresAt, DecidedAt: r.DecidedAt, RevokedAt: r.RevokedAt}
}

// EnrollmentRequestPage is one keyset page of requests; Next is the cursor of the following page (nil on the last page).
type EnrollmentRequestPage struct {
	Requests []EnrollmentRequestView
	Next     *tenancy.EnrollmentRequestID
}

// EnrollmentStatusAdmitted is EnrollmentStatus's status when the credential's T2 proof shows a committed admission;
// EnrollmentStatusNone when there is neither a proof nor a request.
const (
	EnrollmentStatusAdmitted = "admitted"
	EnrollmentStatusNone     = "none"
)

// EnrollmentStatus is the applicant's reconciliation read (Plan r3 §1.3 step 6): Status is admitted when the
// credential's own T2 proof shows a committed admission, otherwise the latest request's stored state, or none; Request
// is that latest request (nil when none exists). It claims nothing the rows do not prove (current membership is not
// readable without a tenant context; a second admission call answers already_member by the body).
type EnrollmentStatus struct {
	Status  string
	Request *EnrollmentRequestView
}

// requestOutcomeFailure maps a refusal row of app.enrollment_request onto its failure: recipient_ineligible,
// enrollment_closed and not_found as the admission's (recorded), enrollment_expired as unauthenticated session_invalid
// (recorded), already_member as a status conflict (not recorded). pending is not a failure; any other value is internal.
func requestOutcomeFailure(o tenancy.RequestOutcome) *failure {
	switch o {
	case tenancy.RequestOutcomeAlreadyMember:
		return conflict(fieldAlreadyMember)
	case tenancy.RequestOutcomeRecipientIneligible:
		return admitRefusal(tenancy.AdmitRecipientIneligible)
	case tenancy.RequestOutcomeEnrollmentExpired:
		return admitRefusal(tenancy.AdmitEnrollmentExpired)
	case tenancy.RequestOutcomeNotFound:
		return admitRefusal(tenancy.AdmitNotFound)
	case tenancy.RequestOutcomeEnrollmentClosed:
		return admitRefusal(tenancy.AdmitEnrollmentClosed)
	}
	return fail(CodeInternal, reasonStoreUnavailable)
}

// ownStoreFailure maps a tenancy error of an applicant operation: a store failure is unavailable with the transaction's
// reason (a class-40 abort, including enrollment_request's one_pending fallback 40001, is aborted: definitely rolled
// back, never success; an unknown COMMIT outcome is decided by inContextTx as commit_outcome_unknown, never aborted);
// an invisible request after its own insert is internal; anything else is internal.
func ownStoreFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	if tenancy.CodeOf(err) == tenancy.CodeUnavailable {
		return fail(CodeUnavailable, failureReason(ctx, tx.cause))
	}
	return fail(CodeInternal, reasonStoreUnavailable)
}

// RequestEnrollment submits h's user's access request to h's Organization (own credential; Plan r3 §1.5): one
// IDENTITY-shape transaction running identity's host guard, then app.enrollment_request once (no internal retry). The
// answer is the one pending request: new (inserted with its enrollment.requested capture by the applicant) or the
// existing one, returned idempotently while pending, expired or not (an expired pending request is never renewed; an
// administrator resolves it). A concurrent first request of the same subject from another credential resolves to the
// same row inside the body. Refusals: enrollment_closed (mode not request_access), recipient_ineligible,
// enrollment_expired, not_found, host_changed (each recorded), already_member (a current member; not recorded) and
// unavailable (aborted for a class-40 abort, commit_outcome_unknown for an unknown COMMIT; a repeat returns the same
// pending row, so nothing is duplicated).
func (s *Service) RequestEnrollment(ctx context.Context, h identity.EnrollmentHandle) (EnrollmentRequestView, error) {
	var out EnrollmentRequestView
	f := s.enrollmentTx(ctx, opEnrollmentRequest, true, h, func(tx *trackedTx, _ tenancy.UserID, org tenancy.OrganizationID) *failure {
		if fl := s.checkEnrollmentHost(ctx, tx, h); fl != nil {
			return fl
		}
		res, err := tenancy.RequestEnrollmentTx(ctx, tx, h.Digest, org)
		if err != nil {
			return ownStoreFailure(ctx, tx, err)
		}
		if res.Outcome != tenancy.RequestOutcomePending {
			return requestOutcomeFailure(res.Outcome)
		}
		r, err := tenancy.EnrollmentRequestTx(ctx, tx, org, res.Request)
		if err != nil {
			return ownStoreFailure(ctx, tx, err)
		}
		out = requestView(r, time.Now())
		return nil
	})
	if f != nil {
		return EnrollmentRequestView{}, f.err()
	}
	return out, nil
}

// EnrollmentStatus reads h's user's enrollment state for h's Organization by read only (own credential; read-only
// IDENTITY-shape transaction): admitted when the credential's T2 proof is set, otherwise the latest request's state, or
// none. A missing, revoked or expired credential is unauthenticated (enrollment_expired, session_invalid; not recorded).
func (s *Service) EnrollmentStatus(ctx context.Context, h identity.EnrollmentHandle) (EnrollmentStatus, error) {
	var out EnrollmentStatus
	f := s.enrollmentTx(ctx, opEnrollmentStatus, false, h, func(tx *trackedTx, user tenancy.UserID, org tenancy.OrganizationID) *failure {
		st, err := tenancy.EnrollmentStateTx(ctx, tx, h.Digest, org, user)
		switch {
		case tenancy.CodeOf(err) == tenancy.CodeEnrollmentExpired:
			return admitRefusal(tenancy.AdmitEnrollmentExpired)
		case err != nil:
			return ownStoreFailure(ctx, tx, err)
		}
		out.Status = EnrollmentStatusNone
		if st.Request != nil {
			v := requestView(*st.Request, time.Now())
			out.Request, out.Status = &v, v.Status
		}
		if st.AdmittedBinding != nil {
			out.Status = EnrollmentStatusAdmitted
		}
		return nil
	})
	if f != nil {
		return EnrollmentStatus{}, f.err()
	}
	return out, nil
}

// AbandonEnrollment revokes h's user's own enrollment credential (own credential; IDENTITY-shape write; revoke_reason
// abandoned). An expired but unrevoked own row is revoked too (safe: revocation never widens authority). Idempotent: an
// already revoked credential changes nothing and succeeds. A pending request is not withdrawn (it stays for an
// administrator's decision).
func (s *Service) AbandonEnrollment(ctx context.Context, h identity.EnrollmentHandle) error {
	f := s.enrollmentTx(ctx, opEnrollmentAbandon, true, h, func(tx *trackedTx, user tenancy.UserID, org tenancy.OrganizationID) *failure {
		if err := tenancy.AbandonEnrollmentTx(ctx, tx, h.Digest, org, user); err != nil {
			return ownStoreFailure(ctx, tx, err)
		}
		return nil
	})
	if f != nil {
		return f.err()
	}
	return nil
}

// requestFailure maps a tenancy error of a manager request operation: an invisible request is not_found; a request no
// longer pending, not approved, already consumed, or a refused transition is conflict with the tenancy code as field;
// everything else as writeFailure.
func requestFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	switch c := tenancy.CodeOf(err); c {
	case tenancy.CodeEnrollmentRequestNotFound:
		return fail(CodeNotFound, reasonTargetNotFound)
	case tenancy.CodeEnrollmentNotPending, tenancy.CodeEnrollmentNotApproved, tenancy.CodeEnrollmentAlreadyConsumed,
		tenancy.CodeEnrollmentRequestTransition:
		return conflict(string(c))
	}
	return writeFailure(ctx, tx, err)
}

// The two decisions of DecideEnrollmentRequest.
const (
	DecisionApprove = "approve"
	DecisionDeny    = "deny"
)

// ListEnrollmentRequests reads one keyset page of the active Organization's requests (organization.manage_members),
// all states or one stored state ("" or pending, approved, denied, revoked). limit 0 is MaxMembersPage; a limit outside
// 1..MaxMembersPage, a malformed cursor or an unknown status is invalid_input before any statement.
func (s *Service) ListEnrollmentRequests(ctx context.Context, tc *TenantContext, after string, limit int, status string) (EnrollmentRequestPage, error) {
	var out EnrollmentRequestPage
	var from tenancy.EnrollmentRequestID
	pre := func() *failure {
		var f *failure
		if limit, f = pageLimit(limit); f != nil {
			return f
		}
		if status != "" && !tenancy.RequestStatusValid(status) {
			return invalid(fieldStatus)
		}
		from, f = cursor(after, tenancy.ParseEnrollmentRequestID)
		return f
	}
	err := s.guard(ctx, tc, opEnrollmentRequestsList, capOrganizationManageMembers, false, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts) *failure {
			rows, err := tenancy.EnrollmentRequestsPage(ctx, tx, f.Organization.ID, status, from, limit)
			if err != nil {
				return requestFailure(ctx, tx, err)
			}
			if len(rows) > limit {
				next := rows[limit-1].ID
				out.Next, rows = &next, rows[:limit]
			}
			now := time.Now()
			out.Requests = make([]EnrollmentRequestView, 0, len(rows))
			for _, r := range rows {
				out.Requests = append(out.Requests, requestView(r, now))
			}
			return nil
		})
	if err != nil {
		return EnrollmentRequestPage{}, err
	}
	return out, nil
}

// DecideEnrollmentRequest approves or denies (DecisionApprove, DecisionDeny) the pending request id of the active
// Organization (organization.manage_members). The Organization row is locked FOR NO KEY UPDATE with the facts, then the
// one UPDATE records the acting manager as the decider (enrollment_requests_immutable re-checks decider = ctx_user()
// and, on approval, derives authority_seq itself; the facade computes nothing). The same statement appends
// enrollment.decided (reason NULL) and, for a denial, the denied restriction row; a refused append or ledger insert
// rolls the decision back. An approval writes no binding: the applicant's own admission does. A request that is
// unknown, foreign or invisible is not_found; one no longer pending conflict (not_pending). Deciding an expired pending
// request is allowed (an administrator resolves it) but never refreshes its lifetime: an expired approval is refused at
// admission (approval_revoked).
func (s *Service) DecideEnrollmentRequest(ctx context.Context, tc *TenantContext, id tenancy.EnrollmentRequestID, decision string) (EnrollmentRequestView, error) {
	var out EnrollmentRequestView
	pre := func() *failure {
		if f := targetID(id)(); f != nil {
			return f
		}
		if decision != DecisionApprove && decision != DecisionDeny {
			return invalid(fieldDecision)
		}
		return nil
	}
	err := s.guard(ctx, tc, opEnrollmentRequestsDecide, capOrganizationManageMembers, true, Target{Kind: kindOrganization}, policyWriteQuery, pre,
		func(tx *trackedTx, f Facts) *failure {
			r, err := tenancy.DecideEnrollmentRequestTx(ctx, tx, f.Organization.ID, id, tc.subject, decision == DecisionApprove)
			if err != nil {
				return requestFailure(ctx, tx, err)
			}
			out = requestView(r, time.Now())
			return nil
		})
	if err != nil {
		return EnrollmentRequestView{}, err
	}
	return out, nil
}

// RevokeEnrollmentApproval revokes the unconsumed approval id of the active Organization (organization.manage_members;
// Plan r3 §1.4 "Approval revocation"), with the Organization row locked FOR NO KEY UPDATE first. The UPDATE takes the
// request row lock that the admission also takes, so a revocation and a consumption serialize. The acting manager is
// the recorded revoker; the same statement appends enrollment.revoked and the revoked restriction row. Idempotent: an
// already revoked request answers its stored row with no second event. A consumed approval is conflict
// (already_consumed: the membership stands; removal is RemoveMember); a pending or denied request conflict
// (not_approved); an unknown, foreign or invisible one not_found.
func (s *Service) RevokeEnrollmentApproval(ctx context.Context, tc *TenantContext, id tenancy.EnrollmentRequestID) (EnrollmentRequestView, error) {
	var out EnrollmentRequestView
	err := s.guard(ctx, tc, opEnrollmentRequestsRevoke, capOrganizationManageMembers, true, Target{Kind: kindOrganization}, policyWriteQuery, targetID(id),
		func(tx *trackedTx, f Facts) *failure {
			r, _, err := tenancy.RevokeEnrollmentApprovalTx(ctx, tx, f.Organization.ID, id, tc.subject)
			if err != nil {
				return requestFailure(ctx, tx, err)
			}
			out = requestView(r, time.Now())
			return nil
		})
	if err != nil {
		return EnrollmentRequestView{}, err
	}
	return out, nil
}

// AdmitAtCallback implements identity.Admitter (Plan r3 §5.1): it converts identity's selector, runs the invitation
// admission (with identity's attested address), the open admission or, for the request kind (request_on_arrival on a
// request_access Organization, Plan r3 §1.3 step 3), the user's own access request (RequestEnrollment: status pending,
// never a membership), and converts the result to identity's closed AdmitOutcome. A known refusal, already recorded by
// this Service, is a status with a nil error, so identity records nothing further; a request by a current member is the
// unrecorded status already_member (the ordinary sign-in then decides); any other failure (invalid input, another G019
// claim conflict, unavailable including an unknown COMMIT, internal) is an error.
func (s *Service) AdmitAtCallback(ctx context.Context, h identity.EnrollmentHandle, sel identity.AdmitSelector) (identity.AdmitOutcome, error) {
	var a EnrollmentAdmission
	var err error
	switch sel.Kind {
	case identity.AdmitKindInvitation:
		if sel.Invitation == nil || sel.Address == nil {
			return identity.AdmitOutcome{}, invalid(fieldEnrollment).err()
		}
		a, err = s.AdmitInvitationAtCallback(ctx, *sel.Address, h, tenancy.InvitationID(*sel.Invitation), sel.InvitationToken)
	case identity.AdmitKindOpen:
		a, err = s.AdmitOpen(ctx, h)
	case identity.AdmitKindRequest:
		if _, err := s.RequestEnrollment(ctx, h); err != nil {
			var e *Error
			if errors.As(err, &e) && e.Code == CodeConflict && e.Field == fieldAlreadyMember {
				return identity.AdmitOutcome{Status: identity.AdmitAlreadyMember}, nil
			}
			if status := admissionStatus(err); status != "" {
				return identity.AdmitOutcome{Status: status}, nil
			}
			return identity.AdmitOutcome{}, err
		}
		return identity.AdmitOutcome{Status: identity.EnrollmentPending}, nil
	default:
		return identity.AdmitOutcome{}, invalid(fieldEnrollment).err()
	}
	if err != nil {
		if status := admissionStatus(err); status != "" {
			return identity.AdmitOutcome{Status: status}, nil
		}
		return identity.AdmitOutcome{}, err
	}
	out := identity.AdmitOutcome{Status: a.Outcome}
	if !zero(a.BindingID) {
		out.BindingID = a.BindingID.String()
	}
	return out, nil
}

// EnrollmentPublic is the public enrollment projection of an Organization resolved from a trusted host (Plan r3 §1.5
// PublicEnrollment): its mode and its public display label (the G017 public branding label; nil when none). It
// discloses no readiness, request, member or user data.
type EnrollmentPublic struct {
	Mode  string
	Label *string
}

// PublicEnrollment returns the public enrollment projection of organization, which the caller resolved from the
// request's host through the trusted G016 resolver. Like PublicBranding it is never authorized by a credential and opens
// no context: two pool statements through the policy reader's and the auth reader's definer functions. A malformed
// Organization is not_found; a store failure unavailable.
func (s *Service) PublicEnrollment(ctx context.Context, organization string) (EnrollmentPublic, error) {
	org, err := tenancy.ParseOrganizationID(organization)
	if err != nil || zero(org) {
		return EnrollmentPublic{}, fail(CodeNotFound, reasonTargetNotFound).err()
	}
	q, ok := s.db.(tenancy.Querier)
	if !ok {
		return EnrollmentPublic{}, fail(CodeUnavailable, reasonStoreUnavailable).err()
	}
	mode, err := tenancy.ReadPublicEnrollmentMode(ctx, q, org)
	if err != nil {
		return EnrollmentPublic{}, fail(CodeUnavailable, failureReason(ctx, err)).err()
	}
	p, f := s.publicRead(ctx, organization)
	if f != nil {
		return EnrollmentPublic{}, f.err()
	}
	return EnrollmentPublic{Mode: string(mode), Label: p.Label}, nil
}

// G058 slice 1b: the guarded enrollment policy (Plan r3 §1.5; selected-design addendum §A.5). Both operations run in
// the active Organization's context through the guarded facade (binding, session, facts, context revalidation and the
// single capability decision); a denial is recorded exactly as every other guarded operation's.

// GetEnrollmentPolicy reads the active Organization's enrollment policy (organization.read): the mode, whether open may
// be enabled (resolved membership history) and the recorded basis ("" while unresolved).
func (s *Service) GetEnrollmentPolicy(ctx context.Context, tc *TenantContext) (tenancy.EnrollmentPolicy, error) {
	var out tenancy.EnrollmentPolicy
	err := s.guard(ctx, tc, opEnrollmentPolicyGet, capOrganizationRead, false, Target{Kind: kindOrganization}, organizationQuery, nil,
		func(tx *trackedTx, f Facts) *failure {
			p, err := tenancy.ReadEnrollmentPolicyTx(ctx, tx, f.Organization.ID)
			if err != nil {
				return writeFailure(ctx, tx, err)
			}
			out = p
			return nil
		})
	if err != nil {
		return tenancy.EnrollmentPolicy{}, err
	}
	return out, nil
}

// notReady is the open refusal on an Organization whose membership history is unresolved (ErrEnrollmentNotReady).
func notReady() *failure {
	return &failure{code: CodeConflict, field: fieldEnrollmentNotReady, reason: reasonEnrollmentNotReady}
}

// UpdateEnrollmentPolicy assigns the active Organization's enrollment mode (organization.manage_policy: a current
// org_owner only); an absolute set, so the current value is a successful no-op without an audit row. A mode outside the
// closed set is invalid_input (invalid_enrollment_mode) before any statement. The Organization row is locked FOR NO KEY
// UPDATE with the facts; for open the readiness pair is then read in the same transaction and an unresolved history is
// ErrEnrollmentNotReady with no UPDATE issued and nothing recorded. Otherwise the one UPDATE runs (00016's owner guard,
// the open-requires-ready CHECK, whose 23514 maps to the same refusal, and the enrollment.policy_changed capture in this
// transaction). request_access and invite_only are never gated by readiness.
func (s *Service) UpdateEnrollmentPolicy(ctx context.Context, tc *TenantContext, mode tenancy.EnrollmentMode) (tenancy.EnrollmentPolicy, error) {
	var out tenancy.EnrollmentPolicy
	pre := func() *failure {
		if !mode.Valid() {
			return &failure{code: CodeInvalidInput, field: string(tenancy.CodeInvalidEnrollmentMode), reason: reasonInputInvalid}
		}
		return nil
	}
	err := s.guard(ctx, tc, opEnrollmentPolicyUpdate, capOrganizationManagePolicy, true, Target{Kind: kindOrganization}, policyWriteQuery, pre,
		func(tx *trackedTx, f Facts) *failure {
			if mode == tenancy.EnrollmentOpen {
				cur, err := tenancy.ReadEnrollmentPolicyTx(ctx, tx, f.Organization.ID)
				if err != nil {
					return writeFailure(ctx, tx, err)
				}
				if !cur.OpenReady {
					return notReady()
				}
			}
			p, err := tenancy.SetEnrollmentModeTx(ctx, tx, f.Organization.ID, mode)
			switch {
			case tenancy.CodeOf(err) == tenancy.CodeEnrollmentNotReady:
				return notReady()
			case err != nil:
				return writeFailure(ctx, tx, err)
			}
			out = p
			return nil
		})
	if err != nil {
		return tenancy.EnrollmentPolicy{}, err
	}
	return out, nil
}

// admissionStatus is identity's closed status of a known admission refusal error, or "" for any other error.
func admissionStatus(err error) string {
	var e *Error
	if !errors.As(err, &e) {
		return ""
	}
	switch {
	case e.Code == CodeNotFound:
		return identity.AdmitNotFound
	case e.Code == CodeUnauthenticated && e.Field == fieldEnrollmentExpired:
		return identity.AdmitEnrollmentExpired
	case e.Code != CodeConflict:
		return ""
	}
	switch e.Field {
	case fieldEnrollmentClosed:
		return identity.AdmitEnrollmentClosed
	case fieldEnrollmentRestricted:
		return identity.AdmitEnrollmentRestricted
	case string(tenancy.CodeRecipientIneligible):
		return identity.AdmitRecipientIneligible
	case fieldHostChanged:
		return identity.AdmitHostChanged
	case string(tenancy.CodeInvitationNotPending):
		return identity.AdmitAlreadyConsumed
	}
	return ""
}
