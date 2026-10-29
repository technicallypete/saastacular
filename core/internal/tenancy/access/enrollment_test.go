package access

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/identity"
	"core/internal/tenancy"
)

// TestAdmissionRefusesBeforeSQL (G058 ENR unit half): a zero VerifiedAddress, a zero or non-canonical handle user, a
// zero or non-canonical Organization, a zero digest and a malformed invitation id or token are refused as invalid_input
// before any statement (the Service has no pool here, so any statement would panic); no access.denied is recorded (it
// would need a pool) and the access.decision record is enrollment.admit deny input_invalid.
func TestAdmissionRefusesBeforeSQL(t *testing.T) {
	var logs bytes.Buffer
	s := newService(nil, nil, slog.New(slog.NewJSONHandler(&logs, nil)))
	ctx := context.Background()
	user, org := uid[tenancy.UserID](0xab), uid[tenancy.OrganizationID](0xcd) // hex letters, so ToUpper is non-canonical
	good := identity.EnrollmentHandle{UserID: user.String(), OrganizationID: org.String(), Origin: "https://a.example.test", Digest: [32]byte{7}}
	want := func(what string, err error, field string) {
		t.Helper()
		var e *Error
		if !errors.As(err, &e) || e.Code != CodeInvalidInput || e.Field != field {
			t.Errorf("%s: %v, want invalid_input %s", what, err, field)
		}
	}

	// The invitation kind: the evidence is checked first, so a zero ev is refused whatever the handle.
	_, err := s.admitInvitation(ctx, identity.VerifiedAddress{}, good, uid[tenancy.InvitationID](0x40), "token")
	want("zero evidence", err, fieldEvidence)
	_, err = s.AdmitInvitationAtCallback(ctx, identity.VerifiedAddress{}, identity.EnrollmentHandle{}, tenancy.InvitationID{}, "")
	want("zero evidence, zero handle", err, fieldEvidence)

	// The open kind: every malformed handle.
	upper := good
	upper.UserID = strings.ToUpper(good.UserID)
	upperOrg := good
	upperOrg.OrganizationID = strings.ToUpper(good.OrganizationID)
	noDigest := good
	noDigest.Digest = [32]byte{}
	noUser := good
	noUser.UserID = ""
	zeroOrg := good
	zeroOrg.OrganizationID = tenancy.OrganizationID{}.String()
	for name, h := range map[string]identity.EnrollmentHandle{
		"zero handle": {}, "non-canonical user": upper, "non-canonical organization": upperOrg, "zero digest": noDigest,
		"empty user": noUser, "zero organization": zeroOrg,
	} {
		_, err := s.AdmitOpen(ctx, h)
		want(name, err, fieldEnrollment)
	}
	if _, _, f := enrollmentHandle(good); f != nil {
		t.Errorf("canonical handle refused: %+v", f)
	}

	n := 0
	for _, line := range bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n")) {
		var r map[string]any
		if err := json.Unmarshal(line, &r); err != nil {
			t.Fatal(err)
		}
		n++
		if r["msg"] != eventDecision || r["operation"] != opEnrollmentAdmit || r["outcome"] != outcomeDeny || r["reason"] != string(reasonInputInvalid) {
			t.Errorf("decision record: %v", r)
		}
		if _, ok := r["capability"]; ok {
			t.Errorf("admission record carries a capability: %v", r)
		}
	}
	if n != 8 {
		t.Errorf("decision records: %d, want 8 (one per call)", n)
	}
}

// TestAdmissionOutcomeMapping (Codex round-2 audit/log decision): the closed admission outcomes map onto the facade's
// codes and fields with their semantic reasons (no write_conflict), the same reason the one access.denied record carries
// for every known denial (enrollment_closed, enrollment_restricted, recipient_ineligible, session_invalid for the
// enrollment_expired aggregate, target_not_found); the lift RAISE is the enrollment_restricted conflict.
func TestAdmissionOutcomeMapping(t *testing.T) {
	for _, c := range []struct {
		o      tenancy.AdmitOutcome
		code   Code
		field  string
		reason Reason
		denial Reason
	}{
		{tenancy.AdmitEnrollmentClosed, CodeConflict, fieldEnrollmentClosed, reasonEnrollmentClosed, reasonEnrollmentClosed},
		{tenancy.AdmitEnrollmentRestricted, CodeConflict, fieldEnrollmentRestricted, reasonEnrollmentRestricted, reasonEnrollmentRestricted},
		{tenancy.AdmitRecipientIneligible, CodeConflict, string(tenancy.CodeRecipientIneligible), reasonRecipientIneligible, reasonRecipientIneligible},
		{tenancy.AdmitEnrollmentExpired, CodeUnauthenticated, fieldEnrollmentExpired, reasonSessionInvalid, reasonSessionInvalid},
		{tenancy.AdmitNotFound, CodeNotFound, "", reasonTargetNotFound, reasonTargetNotFound},
		// G058 request slice: approval_revoked is a recorded denial with its semantic reason; pending and denied are
		// status conflicts (write_conflict), never recorded.
		{tenancy.AdmitApprovalRevoked, CodeConflict, fieldApprovalRevoked, reasonApprovalRevoked, reasonApprovalRevoked},
		{tenancy.AdmitPending, CodeConflict, fieldPending, reasonWriteConflict, ""},
		{tenancy.AdmitDenied, CodeConflict, fieldDenied, reasonWriteConflict, ""},
		{"consumed", CodeInternal, "", reasonStoreUnavailable, ""},
	} {
		f := admitRefusal(c.o)
		if f == nil || f.code != c.code || f.field != c.field || f.reason != c.reason || f.denied != nil {
			t.Errorf("%s: %+v", c.o, f)
			continue
		}
		if got := admissionDenialReason(f); got != c.denial {
			t.Errorf("%s: denial reason %q, want %q", c.o, got, c.denial)
		}
	}
	for _, o := range []tenancy.AdmitOutcome{tenancy.AdmitAdmitted, tenancy.AdmitAlreadyMember} {
		if f := admitRefusal(o); f != nil {
			t.Errorf("%s refused: %+v", o, f)
		}
	}
	ctx, tx := context.Background(), &trackedTx{}
	if f := admitFailure(ctx, tx, &tenancy.Error{Code: tenancy.CodeEnrollmentRestricted}); f == nil || f.code != CodeConflict ||
		f.field != fieldEnrollmentRestricted || f.reason != reasonEnrollmentRestricted || admissionDenialReason(f) != reasonEnrollmentRestricted {
		t.Errorf("lift raise: %+v", f)
	}
	// G019's claim refusal of a disabled recipient becomes the recorded recipient_ineligible; an invitation whose issuer
	// lost authority is the closed not_found / target_not_found (Codex round-3 decision 4); other claim failures are
	// unchanged (an invalid input stays invalid input, a store failure unavailable).
	if f := enrollmentClaimFailure(membershipFailure(ctx, tx, &tenancy.Error{Code: tenancy.CodeRecipientIneligible})); f == nil ||
		f.code != CodeConflict || f.reason != reasonRecipientIneligible || admissionDenialReason(f) != reasonRecipientIneligible {
		t.Errorf("claim recipient_ineligible: %+v", f)
	}
	if f := enrollmentClaimFailure(membershipFailure(ctx, tx, &tenancy.Error{Code: tenancy.CodeIssuerAuthorityLost})); f == nil ||
		f.code != CodeNotFound || f.reason != reasonTargetNotFound || admissionDenialReason(f) != reasonTargetNotFound ||
		admissionStatus(f.err()) != identity.AdmitNotFound {
		t.Errorf("claim issuer_authority_lost: %+v", f)
	}
	for _, in := range []*failure{nil, invalid(string(tenancy.CodeInvalidToken)), fail(CodeUnavailable, reasonAborted)} {
		if f := enrollmentClaimFailure(in); f != in {
			t.Errorf("claim failure %+v remapped to %+v", in, f)
		}
	}
	// An unauthenticated failure other than the enrollment credential's is not session_invalid's denial.
	if r := admissionDenialReason(fail(CodeUnauthenticated, reasonSessionInvalid)); r != "" {
		t.Errorf("foreign unauthenticated recorded as %s", r)
	}
	if f := admitFailure(ctx, tx, &tenancy.Error{Code: tenancy.CodeUnavailable}); f == nil || f.code != CodeUnavailable || admissionDenialReason(f) != "" {
		t.Errorf("unavailable: %+v", f)
	}
	for _, c := range []tenancy.Code{tenancy.CodeInternal, tenancy.CodeInvalidID, ""} {
		if f := admitFailure(ctx, tx, &tenancy.Error{Code: c}); f == nil || f.code != CodeInternal {
			t.Errorf("%q: %+v", c, f)
		}
	}
	// Invalid input, a consumed replay and G019's other claim conflicts are not recorded denials.
	for _, f := range []*failure{nil, invalid(fieldEvidence), conflict(string(tenancy.CodeInvitationNotPending)),
		conflict(string(tenancy.CodeIssuerAuthorityLost)), fail(CodeUnavailable, reasonAborted)} {
		if r := admissionDenialReason(f); r != "" {
			t.Errorf("%+v recorded as %s", f, r)
		}
	}
}

// hostIdentity is a spy identity boundary that also provides the admission host guard: it records the transaction it
// was handed and answers err.
type hostIdentity struct {
	spyIdentity
	err  error
	seen identity.Querier
	h    identity.EnrollmentHandle
}

func (h *hostIdentity) CheckEnrollmentHost(_ context.Context, q identity.Querier, eh identity.EnrollmentHandle) error {
	h.seen, h.h = q, eh
	return h.err
}

// TestEnrollmentHostGuard (Codex round-2 host decision, unit half): the guard is reached on the admission's own tracked
// transaction with the handle; host_changed is the recorded host_changed conflict, a resolver failure is unavailable
// (not a denial), and an identity boundary without the guard, or a nil identity service, fails closed as unavailable.
func TestEnrollmentHostGuard(t *testing.T) {
	ctx := context.Background()
	h := identity.EnrollmentHandle{UserID: "u", OrganizationID: "o", Origin: "https://a.example.test", Digest: [32]byte{1}}
	for _, c := range []struct {
		name   string
		err    error
		code   Code
		reason Reason
		denial Reason
	}{
		{"routes", nil, "", "", ""},
		{"host changed", identity.ErrHostChanged, CodeConflict, reasonHostChanged, reasonHostChanged},
		{"resolver failure", identity.ErrUnavailable, CodeUnavailable, reasonStoreUnavailable, ""},
	} {
		id := &hostIdentity{err: c.err}
		s := newService(nil, id, nil)
		tx := &trackedTx{}
		f := s.checkEnrollmentHost(ctx, tx, h)
		if id.seen != identity.Querier(tx) || id.h != h {
			t.Errorf("%s: guard not reached on the admission transaction (%v)", c.name, id.seen)
		}
		switch {
		case c.code == "" && f != nil:
			t.Errorf("%s: %+v", c.name, f)
		case c.code != "" && (f == nil || f.code != c.code || f.reason != c.reason || admissionDenialReason(f) != c.denial):
			t.Errorf("%s: %+v", c.name, f)
		}
	}
	if f := newService(nil, &spyIdentity{}, nil).checkEnrollmentHost(ctx, &trackedTx{}, h); f == nil || f.code != CodeUnavailable {
		t.Errorf("identity without the guard: %+v", f)
	}
	var nilID *identity.Service
	if f := newService(nil, nilID, nil).checkEnrollmentHost(ctx, &trackedTx{}, h); f == nil || f.code != CodeUnavailable {
		t.Errorf("nil identity service: %+v", f)
	}
	if f := newService(nil, nil, nil).checkEnrollmentHost(ctx, &trackedTx{}, h); f == nil || f.code != CodeUnavailable {
		t.Errorf("no identity boundary: %+v", f)
	}
}

// TestAdmitAtCallbackMapping (identity.Admitter, unit half): access satisfies identity's Admitter; a malformed selector
// is refused before any statement; every known refusal error converts to identity's closed status (recorded by access,
// so a nil error), and anything else stays an error.
func TestAdmitAtCallbackMapping(t *testing.T) {
	var _ identity.Admitter = (*Service)(nil)
	s := newService(nil, nil, nil)
	ctx := context.Background()
	inv := identity.InvitationRef{1}
	for name, sel := range map[string]identity.AdmitSelector{
		"unknown kind":              {Kind: "lift"},
		"request, malformed handle": {Kind: identity.AdmitKindRequest},
		"invitation without id":     {Kind: identity.AdmitKindInvitation, Address: &identity.VerifiedAddress{}},
		"invitation without proof":  {Kind: identity.AdmitKindInvitation, Invitation: &inv, InvitationToken: "t"},
	} {
		if out, err := s.AdmitAtCallback(ctx, identity.EnrollmentHandle{}, sel); err == nil || CodeOf(err) != CodeInvalidInput || out != (identity.AdmitOutcome{}) {
			t.Errorf("%s: %+v %v", name, out, err)
		}
	}
	for _, c := range []struct {
		f    *failure
		want string
	}{
		{&failure{code: CodeConflict, field: fieldEnrollmentClosed}, identity.AdmitEnrollmentClosed},
		{&failure{code: CodeConflict, field: fieldEnrollmentRestricted}, identity.AdmitEnrollmentRestricted},
		{&failure{code: CodeConflict, field: string(tenancy.CodeRecipientIneligible)}, identity.AdmitRecipientIneligible},
		{&failure{code: CodeConflict, field: fieldHostChanged}, identity.AdmitHostChanged},
		{&failure{code: CodeConflict, field: string(tenancy.CodeInvitationNotPending)}, identity.AdmitAlreadyConsumed},
		{&failure{code: CodeUnauthenticated, field: fieldEnrollmentExpired}, identity.AdmitEnrollmentExpired},
		{fail(CodeNotFound, reasonTargetNotFound), identity.AdmitNotFound},
		{&failure{code: CodeConflict, field: string(tenancy.CodeIssuerAuthorityLost)}, ""},
		{fail(CodeUnauthenticated, reasonSessionInvalid), ""},
		{invalid(fieldEvidence), ""},
		{fail(CodeUnavailable, reasonAborted), ""},
		{fail(CodeInternal, reasonStoreUnavailable), ""},
	} {
		if got := admissionStatus(c.f.err()); got != c.want {
			t.Errorf("%+v: status %q, want %q", c.f, got, c.want)
		}
	}
	if got := admissionStatus(errors.New("other")); got != "" {
		t.Errorf("foreign error: %q", got)
	}
	// G058 HTTP slice: the public projection refuses a malformed Organization (not_found) and a service without a
	// statement-capable pool (unavailable) before any statement.
	for org, want := range map[string]Code{"x": CodeNotFound, "00000000-0000-0000-0000-000000000000": CodeNotFound,
		"0f58a000-0000-4000-8000-0000000000a1": CodeUnavailable} {
		if p, err := s.PublicEnrollment(ctx, org); CodeOf(err) != want || p != (EnrollmentPublic{}) {
			t.Errorf("public %q: %+v %v, want %s", org, p, err, want)
		}
	}
}

// TestEnrollmentNotReadyError (G058 slice 1b, addendum §A.5): the open-readiness refusal is conflict with the field
// enrollment_not_ready (HTTP 409) and the same access.decision reason; errors.Is matches ErrEnrollmentNotReady by code
// and field only, so no other conflict, field or code (nor a nil or foreign target) matches it, and the error text stays
// the bounded "access: conflict". It is never a recordable denial (no access.denied row).
func TestEnrollmentNotReadyError(t *testing.T) {
	f := notReady()
	if f.code != CodeConflict || f.field != fieldEnrollmentNotReady || f.reason != reasonEnrollmentNotReady || f.denied != nil {
		t.Fatalf("not ready failure: %+v", f)
	}
	err := f.err()
	if !errors.Is(err, ErrEnrollmentNotReady) || err.Error() != "access: conflict" || CodeOf(err) != CodeConflict {
		t.Errorf("not ready error: %v", err)
	}
	if wrapped := errors.Join(errors.New("context"), err); !errors.Is(wrapped, ErrEnrollmentNotReady) {
		t.Error("wrapped not ready error does not match")
	}
	for name, other := range map[string]error{
		"write conflict":      fail(CodeConflict, reasonWriteConflict).err(),
		"enrollment closed":   (&failure{code: CodeConflict, field: fieldEnrollmentClosed, reason: reasonEnrollmentClosed}).err(),
		"invalid input field": (&failure{code: CodeInvalidInput, field: fieldEnrollmentNotReady}).err(),
		"invalid mode":        (&failure{code: CodeInvalidInput, field: string(tenancy.CodeInvalidEnrollmentMode)}).err(),
		"foreign error":       errors.New("access: conflict"),
	} {
		if errors.Is(other, ErrEnrollmentNotReady) {
			t.Errorf("%s matches ErrEnrollmentNotReady", name)
		}
	}
	var nilErr *Error
	if (&Error{Code: CodeConflict, Field: fieldEnrollmentNotReady}).Is(nilErr) {
		t.Error("a nil *Error target matches")
	}
	if admissionDenialReason(f) != "" {
		t.Error("enrollment_not_ready is an admission denial reason")
	}
}

// TestRequestOperationsRefuseBeforeSQL (G058 request slice, unit half): a zero request id and every malformed handle
// are invalid_input (enrollment) before any statement for the request admission and the applicant's request, status
// and abandonment (the Service has no pool, so a statement would panic); each call writes exactly one access.decision
// record with its own operation, deny input_invalid and no capability.
func TestRequestOperationsRefuseBeforeSQL(t *testing.T) {
	var logs bytes.Buffer
	s := newService(nil, nil, slog.New(slog.NewJSONHandler(&logs, nil)))
	ctx := context.Background()
	user, org := uid[tenancy.UserID](0x1b), uid[tenancy.OrganizationID](0x2c)
	good := identity.EnrollmentHandle{UserID: user.String(), OrganizationID: org.String(), Origin: "https://a.example.test", Digest: [32]byte{7}}
	bad := good
	bad.Digest = [32]byte{}
	want := func(what string, err error) {
		t.Helper()
		var e *Error
		if !errors.As(err, &e) || e.Code != CodeInvalidInput || e.Field != fieldEnrollment {
			t.Errorf("%s: %v, want invalid_input enrollment", what, err)
		}
	}
	_, err := s.AdmitRequest(ctx, good, tenancy.EnrollmentRequestID{})
	want("zero request id", err)
	_, err = s.AdmitRequest(ctx, bad, uid[tenancy.EnrollmentRequestID](0x3d))
	want("request admission, zero digest", err)
	_, err = s.RequestEnrollment(ctx, bad)
	want("request, zero digest", err)
	_, err = s.EnrollmentStatus(ctx, identity.EnrollmentHandle{})
	want("status, zero handle", err)
	want("abandon, zero digest", s.AbandonEnrollment(ctx, bad))
	var ops []string
	for _, line := range bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n")) {
		var r map[string]any
		if err := json.Unmarshal(line, &r); err != nil {
			t.Fatal(err)
		}
		if r["msg"] != eventDecision || r["outcome"] != outcomeDeny || r["reason"] != string(reasonInputInvalid) {
			t.Errorf("decision record: %v", r)
		}
		if _, ok := r["capability"]; ok {
			t.Errorf("own-credential record carries a capability: %v", r)
		}
		op, _ := r["operation"].(string)
		ops = append(ops, op)
	}
	if got, want := strings.Join(ops, ","), strings.Join([]string{opEnrollmentAdmit, opEnrollmentAdmit, opEnrollmentRequest,
		opEnrollmentStatus, opEnrollmentAbandon}, ","); got != want {
		t.Errorf("decision operations: %s, want %s", got, want)
	}
}

// TestRequestOutcomeMapping (G058 request slice): enrollment_request's refusals map as the admission's equivalents
// (recorded with the same semantic reasons on enrollment.request); already_member is an unrecorded status conflict; an
// unknown outcome is internal. The manager operations' refusals are not_found or conflict with the tenancy code as
// field (never recorded). An applicant store failure is unavailable with the transaction's reason: a class-40 abort,
// the one_pending fallback 40001 included, is aborted (a definite rollback), anything else store_unavailable; it is
// never mapped to success or to a refusal.
func TestRequestOutcomeMapping(t *testing.T) {
	for _, c := range []struct {
		o      tenancy.RequestOutcome
		code   Code
		field  string
		denial Reason
	}{
		{tenancy.RequestOutcomeRecipientIneligible, CodeConflict, string(tenancy.CodeRecipientIneligible), reasonRecipientIneligible},
		{tenancy.RequestOutcomeEnrollmentExpired, CodeUnauthenticated, fieldEnrollmentExpired, reasonSessionInvalid},
		{tenancy.RequestOutcomeNotFound, CodeNotFound, "", reasonTargetNotFound},
		{tenancy.RequestOutcomeEnrollmentClosed, CodeConflict, fieldEnrollmentClosed, reasonEnrollmentClosed},
		{tenancy.RequestOutcomeAlreadyMember, CodeConflict, fieldAlreadyMember, ""},
		{tenancy.RequestOutcomePending, CodeInternal, "", ""}, // pending is the success path, never a failure
		{"approved", CodeInternal, "", ""},
	} {
		f := requestOutcomeFailure(c.o)
		if f == nil || f.code != c.code || f.field != c.field || admissionDenialReason(f) != c.denial {
			t.Errorf("%s: %+v", c.o, f)
		}
	}
	ctx := context.Background()
	for _, c := range []struct {
		code  tenancy.Code
		want  Code
		field string
	}{
		{tenancy.CodeEnrollmentRequestNotFound, CodeNotFound, ""},
		{tenancy.CodeEnrollmentNotPending, CodeConflict, "not_pending"},
		{tenancy.CodeEnrollmentNotApproved, CodeConflict, "not_approved"},
		{tenancy.CodeEnrollmentAlreadyConsumed, CodeConflict, "already_consumed"},
		{tenancy.CodeEnrollmentRequestTransition, CodeConflict, "enrollment_request_transition"},
		{tenancy.CodeInternal, CodeInternal, ""},
	} {
		f := requestFailure(ctx, &trackedTx{}, &tenancy.Error{Code: c.code})
		if f == nil || f.code != c.want || f.field != c.field || admissionDenialReason(f) != "" && c.want != CodeNotFound {
			t.Errorf("%s: %+v", c.code, f)
		}
	}
	serialization := &pgconn.PgError{Code: "40001", Message: "enrollment_request_conflict", SeverityUnlocalized: "ERROR"}
	if f := ownStoreFailure(ctx, &trackedTx{cause: serialization}, &tenancy.Error{Code: tenancy.CodeUnavailable}); f == nil ||
		f.code != CodeUnavailable || f.reason != reasonAborted || admissionDenialReason(f) != "" {
		t.Errorf("one_pending fallback: %+v, want unavailable aborted", f)
	}
	if f := ownStoreFailure(ctx, &trackedTx{cause: errors.New("connection reset")}, &tenancy.Error{Code: tenancy.CodeUnavailable}); f == nil ||
		f.code != CodeUnavailable || f.reason != reasonStoreUnavailable {
		t.Errorf("transport failure: %+v, want unavailable store_unavailable", f)
	}
	// An unknown COMMIT outcome is never aborted (commitFailure, used by inContextTx for every transaction here).
	if f := commitFailure(ctx, errors.New("unexpected EOF")); f.reason != reasonCommitOutcomeUnknown {
		t.Errorf("lost COMMIT reply: %+v, want commit_outcome_unknown", f)
	}
	if f := ownStoreFailure(ctx, &trackedTx{}, &tenancy.Error{Code: tenancy.CodeEnrollmentRequestNotFound}); f == nil || f.code != CodeInternal {
		t.Errorf("own request invisible after its insert: %+v, want internal", f)
	}
}
