package tenancy

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// G058 Organization enrollment admission (Plan r3 §1.4, §4.3 enrollment_admit; readiness correction-01 §1 row 6). The
// one admission write is the claim-writer body app.enrollment_admit, called inside the caller's IDENTITY-shape
// transaction (the access facade owns the transaction, its context statement and the audit of refusals).

// AdmitKind is the admission kind of one enrollment_admit call: lift (after an Organization invitation claim in the same
// transaction; AdmitEnrollmentTx), open (AdmitEnrollmentTx) and request (the applicant's approved request;
// AdmitRequestTx).
type AdmitKind string

// The admission kinds.
const (
	AdmitKindLift    AdmitKind = "lift"
	AdmitKindOpen    AdmitKind = "open"
	AdmitKindRequest AdmitKind = "request"
)

// AdmitOutcome is one outcome row of app.enrollment_admit: the closed set of Plan r3 §1.4 that the body returns as a row.
// The lift kind's newer-restriction refusal is not a row: the body raises it (CodeEnrollmentRestricted) so that the
// claim executed earlier in the same transaction rolls back.
type AdmitOutcome string

// The closed outcome set of enrollment_admit rows.
const (
	AdmitAdmitted             AdmitOutcome = "admitted"
	AdmitAlreadyMember        AdmitOutcome = "already_member"
	AdmitRecipientIneligible  AdmitOutcome = "recipient_ineligible"
	AdmitEnrollmentExpired    AdmitOutcome = "enrollment_expired"
	AdmitEnrollmentClosed     AdmitOutcome = "enrollment_closed"
	AdmitEnrollmentRestricted AdmitOutcome = "enrollment_restricted"
	AdmitNotFound             AdmitOutcome = "not_found"
	// The request kind's own refusals: the request is still pending (including an expired pending request, which is
	// never consumable), it was denied, or the approval is no longer current authority (revoked, expired, its decider no
	// longer manages the Organization, or consumed without a current membership).
	AdmitPending         AdmitOutcome = "pending"
	AdmitDenied          AdmitOutcome = "denied"
	AdmitApprovalRevoked AdmitOutcome = "approval_revoked"
)

// admitOutcomes is the closed outcome set; any other value is internal (fail closed).
var admitOutcomes = map[AdmitOutcome]bool{
	AdmitAdmitted: true, AdmitAlreadyMember: true, AdmitRecipientIneligible: true, AdmitEnrollmentExpired: true,
	AdmitEnrollmentClosed: true, AdmitEnrollmentRestricted: true, AdmitNotFound: true,
	AdmitPending: true, AdmitDenied: true, AdmitApprovalRevoked: true,
}

// Admitted reports whether the outcome commits: admitted or already_member. Every other outcome is a refusal that the
// caller rolls back (the body wrote nothing on those paths).
func (o AdmitOutcome) Admitted() bool { return o == AdmitAdmitted || o == AdmitAlreadyMember }

// AdmitResult is one admission outcome. Binding is the admitted Organization membership binding: set for admitted
// (required), optionally for already_member (the reconciled binding), and zero for every refusal.
type AdmitResult struct {
	Outcome AdmitOutcome
	Binding BindingID
}

// AdmitEnrollmentTx runs app.enrollment_admit inside the caller's transaction, as the context user (IDENTITY shape:
// no Organization or Group context). digest is the enrollment credential digest; ref is the invitation the caller just
// claimed in the same transaction (lift, required) or nil (open). A zero digest or Organization, any other kind (request
// has its own primitive, AdmitRequestTx) or a reference that does not match the kind is invalid_id before any statement. The lift kind's newer-restriction RAISE
// (P0001 enrollment_restricted) is CodeEnrollmentRestricted and has aborted the transaction; every returned row maps to
// its closed outcome, and an unknown outcome or an admitted row without a binding is internal.
func AdmitEnrollmentTx(ctx context.Context, tx pgx.Tx, digest [32]byte, org OrganizationID, kind AdmitKind, ref *InvitationID) (AdmitResult, error) {
	if digest == ([32]byte{}) || zero(org) {
		return AdmitResult{}, newError(CodeInvalidID)
	}
	request := pgtype.UUID{}
	switch kind {
	case AdmitKindLift:
		if ref == nil || zero(*ref) {
			return AdmitResult{}, newError(CodeInvalidID)
		}
		request = uuidArg(*ref)
	case AdmitKindOpen:
		if ref != nil {
			return AdmitResult{}, newError(CodeInvalidID)
		}
	default:
		return AdmitResult{}, newError(CodeInvalidID)
	}
	return admitTx(ctx, tx, digest, org, kind, request)
}

// AdmitRequestTx runs app.enrollment_admit(digest, org, 'request', id) inside the caller's IDENTITY-shape transaction:
// the subject's own approved, unconsumed, unexpired request id of org whose decider still manages the Organization
// becomes the membership binding, its consumption and the lift of the restrictions it covers, in one body. A zero
// digest, Organization or request is invalid_id before any statement. Every returned row maps to its closed outcome
// (pending, denied and approval_revoked are this kind's own refusals, with nothing written).
func AdmitRequestTx(ctx context.Context, tx pgx.Tx, digest [32]byte, org OrganizationID, id EnrollmentRequestID) (AdmitResult, error) {
	if digest == ([32]byte{}) || zero(org) || zero(id) {
		return AdmitResult{}, newError(CodeInvalidID)
	}
	return admitTx(ctx, tx, digest, org, AdmitKindRequest, uuidArg(id))
}

func admitTx(ctx context.Context, tx pgx.Tx, digest [32]byte, org OrganizationID, kind AdmitKind, request pgtype.UUID) (AdmitResult, error) {
	out, err := scanAdmit(tx.QueryRow(ctx, sqlEnrollmentAdmit, digest[:], uuidArg(org), string(kind), request))
	if err != nil {
		return AdmitResult{}, classify(one(err, CodeInternal))
	}
	return out, nil
}

// G058 request slice: access requests (Plan r3 §1.4/§1.5; 00016 enrollment_request, enrollment_requests,
// enrollment_requests_immutable; Codex round-8 dispositions and request-c1 rulings).

// EnrollmentRequestID is the canonical UUID of an access request (generated inside enrollment_request).
type EnrollmentRequestID [16]byte

func (id EnrollmentRequestID) String() string { return formatUUID(id) }

// ParseEnrollmentRequestID parses an access request ID; see ParseOrganizationID.
func ParseEnrollmentRequestID(s string) (EnrollmentRequestID, error) {
	return parseUUID[EnrollmentRequestID](s)
}

// The stored request states (enrollment_requests_status_valid). An expired pending request stays pending: expiry is a
// derived fact (ExpiresAt), it keeps occupying the one pending slot and is never consumable or renewed (Codex round-8
// disposition 3).
const (
	RequestPending  = "pending"
	RequestApproved = "approved"
	RequestDenied   = "denied"
	RequestRevoked  = "revoked"
)

// RequestStatusValid reports whether s is one of the four stored request states.
func RequestStatusValid(s string) bool {
	return s == RequestPending || s == RequestApproved || s == RequestDenied || s == RequestRevoked
}

// EnrollmentRequest is one stored access request as the runtime reads it (all thirteen columns). AuthoritySeq is the
// restriction watermark the approval guard derived (0 unless approved or revoked).
type EnrollmentRequest struct {
	ID                EnrollmentRequestID
	OrganizationID    OrganizationID
	UserID            UserID
	Status            string
	CreatedAt         time.Time
	ExpiresAt         time.Time
	DecidedAt         *time.Time
	DecidedByUserID   *UserID
	RevokedAt         *time.Time
	RevokedByUserID   *UserID
	ConsumedAt        *time.Time
	ConsumedBindingID *BindingID
	AuthoritySeq      int64
}

// Expired reports whether the request's lifetime has passed at now (derived; the stored status is unchanged).
func (r EnrollmentRequest) Expired(now time.Time) bool { return !r.ExpiresAt.After(now) }

func scanEnrollmentRequest(row pgx.Row) (EnrollmentRequest, error) {
	var r EnrollmentRequest
	var id, org, user, decider, revoker, binding pgtype.UUID
	if err := row.Scan(&id, &org, &user, &r.Status, &r.CreatedAt, &r.ExpiresAt, &r.DecidedAt, &decider, &r.RevokedAt, &revoker,
		&r.ConsumedAt, &binding, &r.AuthoritySeq); err != nil {
		return EnrollmentRequest{}, err
	}
	r.ID, r.OrganizationID, r.UserID = EnrollmentRequestID(id.Bytes), OrganizationID(org.Bytes), UserID(user.Bytes)
	if decider.Valid {
		u := UserID(decider.Bytes)
		r.DecidedByUserID = &u
	}
	if revoker.Valid {
		u := UserID(revoker.Bytes)
		r.RevokedByUserID = &u
	}
	if binding.Valid {
		b := BindingID(binding.Bytes)
		r.ConsumedBindingID = &b
	}
	return r, nil
}

// RequestOutcome is one outcome row of app.enrollment_request: pending (with the one pending request: new, or the
// existing one returned idempotently, expired or not) or a refusal without a request (recipient_ineligible,
// enrollment_expired, not_found, enrollment_closed, already_member).
type RequestOutcome string

// The closed outcome set of enrollment_request rows.
const (
	RequestOutcomePending             RequestOutcome = "pending"
	RequestOutcomeRecipientIneligible RequestOutcome = "recipient_ineligible"
	RequestOutcomeEnrollmentExpired   RequestOutcome = "enrollment_expired"
	RequestOutcomeNotFound            RequestOutcome = "not_found"
	RequestOutcomeEnrollmentClosed    RequestOutcome = "enrollment_closed"
	RequestOutcomeAlreadyMember       RequestOutcome = "already_member"
)

var requestOutcomes = map[RequestOutcome]bool{
	RequestOutcomePending: true, RequestOutcomeRecipientIneligible: true, RequestOutcomeEnrollmentExpired: true,
	RequestOutcomeNotFound: true, RequestOutcomeEnrollmentClosed: true, RequestOutcomeAlreadyMember: true,
}

// RequestResult is one request outcome; Request is set exactly for pending.
type RequestResult struct {
	Outcome RequestOutcome
	Request EnrollmentRequestID
}

// scanRequestOutcome maps one enrollment_request row onto its closed outcome: an unknown value, pending without a
// request, or a refusal carrying one is internal (fail closed).
func scanRequestOutcome(row pgx.Row) (RequestResult, error) {
	var outcome string
	var id pgtype.UUID
	if err := row.Scan(&outcome, &id); err != nil {
		return RequestResult{}, err
	}
	o := RequestOutcome(outcome)
	if !requestOutcomes[o] || (o == RequestOutcomePending) != id.Valid {
		return RequestResult{}, newError(CodeInternal)
	}
	return RequestResult{Outcome: o, Request: EnrollmentRequestID(id.Bytes)}, nil
}

// RequestEnrollmentTx runs app.enrollment_request(digest, org) inside the caller's IDENTITY-shape transaction (no
// Organization or Group context): the body rechecks the subject, its live credential, the request_access mode and
// current membership, and returns the one pending request, inserting it (with its enrollment.requested capture) only
// when none is pending. A zero digest or Organization is invalid_id before any statement. A concurrent first request of
// the same subject resolves inside the body to the committed pending row (ON CONFLICT on one_pending); the body's
// fallback when that row cannot be read is 40001 enrollment_request_conflict, which is unavailable here (class 40, the
// transaction aborted, nothing committed): the caller reports it as aborted and never retries it internally.
func RequestEnrollmentTx(ctx context.Context, tx pgx.Tx, digest [32]byte, org OrganizationID) (RequestResult, error) {
	if digest == ([32]byte{}) || zero(org) {
		return RequestResult{}, newError(CodeInvalidID)
	}
	out, err := scanRequestOutcome(tx.QueryRow(ctx, sqlEnrollmentRequest, digest[:], uuidArg(org)))
	if err != nil {
		return RequestResult{}, classify(one(err, CodeInternal))
	}
	return out, nil
}

// EnrollmentRequestTx reads request id of org inside the caller's transaction, as row-level security admits it (the
// applicant's own row, or any row of the context Organization for its managers). A row the context cannot see is
// enrollment_request_not_found.
func EnrollmentRequestTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id EnrollmentRequestID) (EnrollmentRequest, error) {
	if zero(org) || zero(id) {
		return EnrollmentRequest{}, newError(CodeInvalidID)
	}
	r, err := scanEnrollmentRequest(tx.QueryRow(ctx, sqlSelectEnrollmentRequest, uuidArg(org), uuidArg(id)))
	if e := classify(one(err, CodeEnrollmentRequestNotFound)); e != nil {
		return EnrollmentRequest{}, e
	}
	return r, nil
}

// DecideEnrollmentRequestTx decides the pending request id of org as actor (approve: approved, else denied) inside the
// caller's Organization-context transaction and returns the decided row. The guard enrollment_requests_immutable
// derives authority_seq on approval under the Organization lock and refuses a decider other than ctx_user(); the
// enrollment.decided capture and, for a denial, the denied restriction row are written by the same statement (a refused
// append or ledger insert aborts it). Zero rows are disambiguated by a re-read: invisible → enrollment_request_not_found,
// any other state → not_pending. A zero Organization, request or actor is invalid_id before any statement.
func DecideEnrollmentRequestTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id EnrollmentRequestID, actor UserID, approve bool) (EnrollmentRequest, error) {
	if zero(org) || zero(id) || zero(actor) {
		return EnrollmentRequest{}, newError(CodeInvalidID)
	}
	status := RequestDenied
	if approve {
		status = RequestApproved
	}
	r, err := scanEnrollmentRequest(tx.QueryRow(ctx, sqlDecideEnrollmentRequest, uuidArg(org), uuidArg(id), status, uuidArg(actor)))
	if err == nil {
		return r, nil
	}
	if !isNoRows(err) {
		return EnrollmentRequest{}, classify(err)
	}
	if _, err := EnrollmentRequestTx(ctx, tx, org, id); err != nil {
		return EnrollmentRequest{}, err
	}
	return EnrollmentRequest{}, newError(CodeEnrollmentNotPending)
}

// RevokeEnrollmentApprovalTx revokes the unconsumed approval id of org as actor inside the caller's Organization-context
// transaction (Plan r3 §1.4 "Approval revocation"). revoked reports whether this call changed the row: false with the
// stored row for a request already revoked (idempotent: no second event or restriction). Zero rows otherwise re-read:
// invisible → enrollment_request_not_found; a consumed approval → already_consumed (the membership stands); pending or
// denied → not_approved. The enrollment.revoked capture and the revoked restriction row are written by the same
// statement. A zero Organization, request or actor is invalid_id before any statement.
func RevokeEnrollmentApprovalTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id EnrollmentRequestID, actor UserID) (r EnrollmentRequest, revoked bool, err error) {
	if zero(org) || zero(id) || zero(actor) {
		return EnrollmentRequest{}, false, newError(CodeInvalidID)
	}
	r, err = scanEnrollmentRequest(tx.QueryRow(ctx, sqlRevokeEnrollmentApproval, uuidArg(org), uuidArg(id), uuidArg(actor)))
	if err == nil {
		return r, true, nil
	}
	if !isNoRows(err) {
		return EnrollmentRequest{}, false, classify(err)
	}
	cur, err := EnrollmentRequestTx(ctx, tx, org, id)
	switch {
	case err != nil:
		return EnrollmentRequest{}, false, err
	case cur.Status == RequestRevoked:
		return cur, false, nil
	case cur.Status == RequestApproved && cur.ConsumedAt != nil:
		return EnrollmentRequest{}, false, newError(CodeEnrollmentAlreadyConsumed)
	}
	return EnrollmentRequest{}, false, newError(CodeEnrollmentNotApproved)
}

// EnrollmentRequestsPage reads one keyset page of org's requests after the cursor (zero: from the start), at most
// limit+1 rows so the caller can tell whether another page follows; status "" lists every state, otherwise one stored
// state. A zero Organization, a limit outside 1..MaxMembersPage or an unknown status is invalid_id before any statement.
func EnrollmentRequestsPage(ctx context.Context, tx pgx.Tx, org OrganizationID, status string, after EnrollmentRequestID, limit int) ([]EnrollmentRequest, error) {
	if zero(org) || limit < 1 || limit > MaxMembersPage || (status != "" && !RequestStatusValid(status)) {
		return nil, newError(CodeInvalidID)
	}
	filter := pgtype.Text{String: status, Valid: status != ""}
	rows, err := tx.Query(ctx, sqlEnrollmentRequestsPage, uuidArg(org), filter, uuidArg(after), limit+1)
	out, err := collect(rows, err, scanEnrollmentRequest)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// EnrollmentState is the applicant's own reconciliation read (Plan r3 §1.3 step 6): AdmittedBinding is the binding the
// credential's T2 proof names (nil until an admission committed); Request is the latest request for the Organization
// (nil when none exists).
type EnrollmentState struct {
	AdmittedBinding *BindingID
	Request         *EnrollmentRequest
}

// EnrollmentStateTx reads user's enrollment state for org inside the caller's IDENTITY-shape transaction, by read only:
// the live credential digest's admitted proof, then the latest request. A credential that is missing, revoked, expired
// or not the caller's is enrollment_expired; a zero digest, Organization or user is invalid_id before any statement.
func EnrollmentStateTx(ctx context.Context, tx pgx.Tx, digest [32]byte, org OrganizationID, user UserID) (EnrollmentState, error) {
	if digest == ([32]byte{}) || zero(org) || zero(user) {
		return EnrollmentState{}, newError(CodeInvalidID)
	}
	var out EnrollmentState
	var at pgtype.Timestamptz
	var binding pgtype.UUID
	err := tx.QueryRow(ctx, sqlEnrollmentCredentialState, uuidArg(org), uuidArg(user), digest[:]).Scan(&at, &binding)
	if e := classify(one(err, CodeEnrollmentExpired)); e != nil {
		return EnrollmentState{}, e
	}
	if at.Valid && binding.Valid {
		b := BindingID(binding.Bytes)
		out.AdmittedBinding = &b
	}
	r, err := scanEnrollmentRequest(tx.QueryRow(ctx, sqlOwnEnrollmentRequest, uuidArg(org), uuidArg(user)))
	switch {
	case err == nil:
		out.Request = &r
	case !isNoRows(err):
		return EnrollmentState{}, classify(err)
	}
	return out, nil
}

// AbandonEnrollmentTx revokes user's own live enrollment credential digest for org inside the caller's IDENTITY-shape
// transaction (revoke_reason abandoned). An expired but unrevoked own row is revoked too (safe: revocation never widens
// authority). Idempotent: a credential already revoked or unknown changes nothing and is not an error. A zero digest,
// Organization or user is invalid_id before any statement.
func AbandonEnrollmentTx(ctx context.Context, tx pgx.Tx, digest [32]byte, org OrganizationID, user UserID) error {
	if digest == ([32]byte{}) || zero(org) || zero(user) {
		return newError(CodeInvalidID)
	}
	if _, err := tx.Exec(ctx, sqlAbandonEnrollment, uuidArg(org), uuidArg(user), digest[:]); err != nil {
		return classify(err)
	}
	return nil
}

// scanAdmit maps one admission row onto its closed outcome: an unknown value, or admitted without a binding, is
// internal; a refusal never carries a binding.
func scanAdmit(row pgx.Row) (AdmitResult, error) {
	var outcome string
	var binding pgtype.UUID
	if err := row.Scan(&outcome, &binding); err != nil {
		return AdmitResult{}, err
	}
	o := AdmitOutcome(outcome)
	if !admitOutcomes[o] {
		return AdmitResult{}, newError(CodeInternal)
	}
	out := AdmitResult{Outcome: o}
	switch {
	case o == AdmitAdmitted && !binding.Valid:
		return AdmitResult{}, newError(CodeInternal)
	case o.Admitted() && binding.Valid:
		out.Binding = BindingID(binding.Bytes)
	}
	return out, nil
}

// G058 slice 1b: the Organization enrollment policy (Plan r3 §1.5; selected-design addendum §A.1/§A.5). The mode is
// closed at rest (NOT NULL + organizations_enrollment_mode_valid); the readiness fact is the operator-set pair
// enrollment_history_resolved_at/_basis (NULL/NULL = unresolved legacy history), written only through the superuser
// channel. organizations_open_requires_resolved_history is the authoritative open guard.

// EnrollmentMode is one Organization's enrollment mode.
type EnrollmentMode string

// The closed mode set (00016 organizations_enrollment_mode_valid).
const (
	EnrollmentInviteOnly    EnrollmentMode = "invite_only"
	EnrollmentRequestAccess EnrollmentMode = "request_access"
	EnrollmentOpen          EnrollmentMode = "open"
)

// Valid reports whether m is one of the three modes.
func (m EnrollmentMode) Valid() bool {
	return m == EnrollmentInviteOnly || m == EnrollmentRequestAccess || m == EnrollmentOpen
}

// ParseEnrollmentMode maps a stored mode onto the closed set: any unknown value reads as invite_only (closed).
func ParseEnrollmentMode(s string) EnrollmentMode {
	if m := EnrollmentMode(s); m.Valid() {
		return m
	}
	return EnrollmentInviteOnly
}

// EnrollmentPolicy is the policy read of one Organization (addendum §A.5 {mode, open_ready, history_basis}): OpenReady
// reports resolved membership history (open may be enabled); HistoryBasis is the recorded basis, "" while unresolved.
type EnrollmentPolicy struct {
	Mode         EnrollmentMode
	OpenReady    bool
	HistoryBasis string
}

// ParseEnrollmentPolicy maps the stored mode and readiness pair onto an EnrollmentPolicy, fail closed: an unknown mode
// is invite_only, and readiness holds only when both readiness columns are set (the shape CHECK makes a half-set pair
// impossible at rest; it reads as unresolved).
func ParseEnrollmentPolicy(mode string, resolvedAt pgtype.Timestamptz, basis pgtype.Text) EnrollmentPolicy {
	out := EnrollmentPolicy{Mode: ParseEnrollmentMode(mode)}
	if resolvedAt.Valid && basis.Valid && basis.String != "" {
		out.OpenReady, out.HistoryBasis = true, basis.String
	}
	return out
}

func scanEnrollmentPolicy(row pgx.Row) (EnrollmentPolicy, error) {
	var mode string
	var at pgtype.Timestamptz
	var basis pgtype.Text
	if err := row.Scan(&mode, &at, &basis); err != nil {
		return EnrollmentPolicy{}, err
	}
	return ParseEnrollmentPolicy(mode, at, basis), nil
}

// ReadEnrollmentPolicyTx reads org's enrollment policy inside the caller's transaction (the guarded facade has already
// read and, for a write, locked the Organization row). A zero Organization is invalid_id; a missing row is
// organization_not_found.
func ReadEnrollmentPolicyTx(ctx context.Context, tx pgx.Tx, org OrganizationID) (EnrollmentPolicy, error) {
	if zero(org) {
		return EnrollmentPolicy{}, newError(CodeInvalidID)
	}
	out, err := scanEnrollmentPolicy(tx.QueryRow(ctx, sqlReadEnrollmentPolicy, uuidArg(org)))
	if e := classify(one(err, CodeOrganizationNotFound)); e != nil {
		return EnrollmentPolicy{}, e
	}
	return out, nil
}

// ReadPublicEnrollmentMode reads the public enrollment mode of org through the policy reader's definer function, as one
// statement on q (no context; an absent Organization reads invite_only). A zero Organization is invalid_id before any
// statement; an unknown stored value parses closed (invite_only).
func ReadPublicEnrollmentMode(ctx context.Context, q Querier, org OrganizationID) (EnrollmentMode, error) {
	if zero(org) {
		return "", newError(CodeInvalidID)
	}
	var mode string
	if err := q.QueryRow(ctx, sqlPublicEnrollmentMode, uuidArg(org)).Scan(&mode); err != nil {
		return "", err
	}
	return ParseEnrollmentMode(mode), nil
}

// SetEnrollmentModeTx assigns org's enrollment mode inside the caller's transaction (an absolute set: the current value
// is a successful no-op without an audit row) and returns the stored policy. A zero Organization is invalid_id and a mode
// outside the closed set invalid_enrollment_mode, both before any statement; open on an unresolved Organization is the
// CHECK's enrollment_not_ready (23514); the owner guard's refusal and any other refusal are internal.
func SetEnrollmentModeTx(ctx context.Context, tx pgx.Tx, org OrganizationID, mode EnrollmentMode) (EnrollmentPolicy, error) {
	if zero(org) {
		return EnrollmentPolicy{}, newError(CodeInvalidID)
	}
	if !mode.Valid() {
		return EnrollmentPolicy{}, newError(CodeInvalidEnrollmentMode)
	}
	out, err := scanEnrollmentPolicy(tx.QueryRow(ctx, sqlSetEnrollmentMode, uuidArg(org), string(mode)))
	if e := classify(one(err, CodeOrganizationNotFound)); e != nil {
		return EnrollmentPolicy{}, e
	}
	return out, nil
}
