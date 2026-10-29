package tenancy

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// admitRow is a synthetic enrollment_admit row (or scan error).
type admitRow struct {
	outcome string
	binding pgtype.UUID
	err     error
}

func (r admitRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*string) = r.outcome
	*dest[1].(*pgtype.UUID) = r.binding
	return nil
}

// TestScanAdmit (G058 ENR unit half): every row of the closed outcome set maps onto its outcome; admitted requires a
// binding; a refusal never carries one; an unknown or empty outcome is internal (fail closed); only admitted and
// already_member commit.
func TestScanAdmit(t *testing.T) {
	b, _ := ParseBindingID("3f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	for _, o := range []AdmitOutcome{AdmitAdmitted, AdmitAlreadyMember, AdmitRecipientIneligible, AdmitEnrollmentExpired,
		AdmitEnrollmentClosed, AdmitEnrollmentRestricted, AdmitNotFound, AdmitPending, AdmitDenied, AdmitApprovalRevoked} {
		out, err := scanAdmit(admitRow{outcome: string(o), binding: uuidArg(b)})
		if err != nil || out.Outcome != o {
			t.Errorf("%s: %+v %v", o, out, err)
			continue
		}
		if o.Admitted() != (o == AdmitAdmitted || o == AdmitAlreadyMember) {
			t.Errorf("%s: Admitted() = %t", o, o.Admitted())
		}
		if want := o.Admitted(); (out.Binding == b) != want {
			t.Errorf("%s: binding %v, want carried %t", o, out.Binding, want)
		}
	}
	if len(admitOutcomes) != 10 {
		t.Errorf("closed outcome set: %d, want 10 (seven of slice 1a plus the request kind's pending, denied, approval_revoked)", len(admitOutcomes))
	}
	if out, err := scanAdmit(admitRow{outcome: "already_member"}); err != nil || out.Outcome != AdmitAlreadyMember || !zero(out.Binding) {
		t.Errorf("already_member without binding: %+v %v", out, err)
	}
	for name, r := range map[string]admitRow{
		"admitted without binding": {outcome: "admitted"},
		"unknown outcome":          {outcome: "consumed"},
		"empty outcome":            {},
		"case variant":             {outcome: "Approval_revoked"},
		"removed outcome (D-17)":   {outcome: "admitted_sign_in_required"},
	} {
		if _, err := scanAdmit(r); CodeOf(err) != CodeInternal {
			t.Errorf("%s: %v, want internal", name, err)
		}
	}
	// A missing row (never produced by the set-returning body) is internal; a store error keeps its class.
	if CodeOf(classify(one(admitRow{err: pgx.ErrNoRows}.Scan(), CodeInternal))) != CodeInternal {
		t.Error("no rows is not internal")
	}
}

// TestAdmitEnrollmentTxRefusesBeforeSQL: a zero digest or Organization, an unknown kind, lift without (or with a zero)
// invitation and open with one are invalid_id before any statement (the transaction is nil, so a statement would panic).
func TestAdmitEnrollmentTxRefusesBeforeSQL(t *testing.T) {
	ctx := context.Background()
	d := [32]byte{1}
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	inv, _ := ParseInvitationID("1f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	zeroInv := InvitationID{}
	for name, c := range map[string]struct {
		digest [32]byte
		org    OrganizationID
		kind   AdmitKind
		ref    *InvitationID
	}{
		"zero digest":         {[32]byte{}, org, AdmitKindOpen, nil},
		"zero organization":   {d, OrganizationID{}, AdmitKindOpen, nil},
		"request kind":        {d, org, AdmitKindRequest, nil}, // AdmitRequestTx is the request kind's primitive
		"empty kind":          {d, org, "", nil},
		"lift without ref":    {d, org, AdmitKindLift, nil},
		"lift with zero ref":  {d, org, AdmitKindLift, &zeroInv},
		"open with reference": {d, org, AdmitKindOpen, &inv},
	} {
		if _, err := AdmitEnrollmentTx(ctx, nil, c.digest, c.org, c.kind, c.ref); CodeOf(err) != CodeInvalidID {
			t.Errorf("%s: %v, want invalid_id", name, err)
		}
	}
}

// TestClassifyEnrollmentRaise: the lift kind's RAISE (P0001, message exactly enrollment_restricted) is
// CodeEnrollmentRestricted; any other P0001 message is internal; the error text stays bounded.
func TestClassifyEnrollmentRaise(t *testing.T) {
	raise := func(msg string) error {
		return fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "P0001", Message: msg, Detail: "secret-canary-detail"})
	}
	if e := classify(raise("enrollment_restricted")); e == nil || e.Code != CodeEnrollmentRestricted || e.Error() != "tenancy: enrollment_restricted" {
		t.Errorf("enrollment_restricted raise: %v", e)
	}
	for _, msg := range []string{"enrollment_restricted ", "Enrollment_restricted", "enrollment_closed", ""} {
		if e := classify(raise(msg)); e == nil || e.Code != CodeInternal {
			t.Errorf("raise %q: %v, want internal", msg, e)
		}
	}
	// The bodies' argument refusals (22023) are internal: the primitives validate every argument before the statement.
	for _, msg := range []string{"enrollment_kind_invalid", "enrollment_request_invalid"} {
		if e := classify(&pgconn.PgError{Code: "22023", Message: msg}); e == nil || e.Code != CodeInternal {
			t.Errorf("%s: %v, want internal", msg, e)
		}
	}
}

// TestRequestPrimitivesRefuseBeforeSQL (G058 request slice): every request primitive refuses a zero digest,
// Organization, request, actor or user, a limit outside 1..MaxMembersPage and an unknown status filter with invalid_id
// before any statement (the transaction is nil, so a statement would panic).
func TestRequestPrimitivesRefuseBeforeSQL(t *testing.T) {
	ctx := context.Background()
	d := [32]byte{1}
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	id, _ := ParseEnrollmentRequestID("2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	user, _ := ParseUserID("3f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	var zd [32]byte
	var zo OrganizationID
	var zr EnrollmentRequestID
	var zu UserID
	calls := map[string]func() error{
		"admit request zero digest": func() error { _, err := AdmitRequestTx(ctx, nil, zd, org, id); return err },
		"admit request zero org":    func() error { _, err := AdmitRequestTx(ctx, nil, d, zo, id); return err },
		"admit request zero id":     func() error { _, err := AdmitRequestTx(ctx, nil, d, org, zr); return err },
		"request zero digest":       func() error { _, err := RequestEnrollmentTx(ctx, nil, zd, org); return err },
		"request zero org":          func() error { _, err := RequestEnrollmentTx(ctx, nil, d, zo); return err },
		"read zero org":             func() error { _, err := EnrollmentRequestTx(ctx, nil, zo, id); return err },
		"read zero id":              func() error { _, err := EnrollmentRequestTx(ctx, nil, org, zr); return err },
		"decide zero org":           func() error { _, err := DecideEnrollmentRequestTx(ctx, nil, zo, id, user, true); return err },
		"decide zero id":            func() error { _, err := DecideEnrollmentRequestTx(ctx, nil, org, zr, user, false); return err },
		"decide zero actor":         func() error { _, err := DecideEnrollmentRequestTx(ctx, nil, org, id, zu, true); return err },
		"revoke zero org":           func() error { _, _, err := RevokeEnrollmentApprovalTx(ctx, nil, zo, id, user); return err },
		"revoke zero id":            func() error { _, _, err := RevokeEnrollmentApprovalTx(ctx, nil, org, zr, user); return err },
		"revoke zero actor":         func() error { _, _, err := RevokeEnrollmentApprovalTx(ctx, nil, org, id, zu); return err },
		"page zero org":             func() error { _, err := EnrollmentRequestsPage(ctx, nil, zo, "", zr, 10); return err },
		"page limit 0":              func() error { _, err := EnrollmentRequestsPage(ctx, nil, org, "", zr, 0); return err },
		"page limit 101":            func() error { _, err := EnrollmentRequestsPage(ctx, nil, org, "", zr, MaxMembersPage+1); return err },
		"page unknown status":       func() error { _, err := EnrollmentRequestsPage(ctx, nil, org, "expired", zr, 10); return err },
		"page status case":          func() error { _, err := EnrollmentRequestsPage(ctx, nil, org, "Pending", zr, 10); return err },
		"state zero digest":         func() error { _, err := EnrollmentStateTx(ctx, nil, zd, org, user); return err },
		"state zero org":            func() error { _, err := EnrollmentStateTx(ctx, nil, d, zo, user); return err },
		"state zero user":           func() error { _, err := EnrollmentStateTx(ctx, nil, d, org, zu); return err },
		"abandon zero digest":       func() error { return AbandonEnrollmentTx(ctx, nil, zd, org, user) },
		"abandon zero org":          func() error { return AbandonEnrollmentTx(ctx, nil, d, zo, user) },
		"abandon zero user":         func() error { return AbandonEnrollmentTx(ctx, nil, d, org, zu) },
	}
	for name, call := range calls {
		if err := call(); CodeOf(err) != CodeInvalidID {
			t.Errorf("%s: %v, want invalid_id", name, err)
		}
	}
	for _, s := range []string{RequestPending, RequestApproved, RequestDenied, RequestRevoked} {
		if !RequestStatusValid(s) {
			t.Errorf("%s invalid", s)
		}
	}
	if _, err := ParseEnrollmentRequestID("00000000-0000-0000-0000-000000000000"); CodeOf(err) != CodeInvalidID {
		t.Errorf("zero request id parsed: %v", err)
	}
	if id.String() != "2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9" {
		t.Errorf("request id round trip: %s", id)
	}
}

// requestRow is a synthetic enrollment_request row.
type requestRow struct {
	outcome string
	id      pgtype.UUID
}

func (r requestRow) Scan(dest ...any) error {
	*dest[0].(*string) = r.outcome
	*dest[1].(*pgtype.UUID) = r.id
	return nil
}

// TestScanRequestOutcome: pending carries exactly its request; every refusal of the closed set carries none; an unknown
// outcome, pending without a request or a refusal with one is internal (fail closed).
func TestScanRequestOutcome(t *testing.T) {
	id, _ := ParseEnrollmentRequestID("2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	if out, err := scanRequestOutcome(requestRow{outcome: "pending", id: uuidArg(id)}); err != nil || out != (RequestResult{Outcome: RequestOutcomePending, Request: id}) {
		t.Errorf("pending: %+v %v", out, err)
	}
	for _, o := range []RequestOutcome{RequestOutcomeRecipientIneligible, RequestOutcomeEnrollmentExpired, RequestOutcomeNotFound,
		RequestOutcomeEnrollmentClosed, RequestOutcomeAlreadyMember} {
		if out, err := scanRequestOutcome(requestRow{outcome: string(o)}); err != nil || out.Outcome != o || !zero(out.Request) {
			t.Errorf("%s: %+v %v", o, out, err)
		}
		if _, err := scanRequestOutcome(requestRow{outcome: string(o), id: uuidArg(id)}); CodeOf(err) != CodeInternal {
			t.Errorf("%s with a request: %v, want internal", o, err)
		}
	}
	if len(requestOutcomes) != 6 {
		t.Errorf("closed request outcome set: %d, want 6", len(requestOutcomes))
	}
	for _, r := range []requestRow{{outcome: "pending"}, {outcome: "approved", id: uuidArg(id)}, {outcome: ""}, {outcome: "denied"}} {
		if _, err := scanRequestOutcome(r); CodeOf(err) != CodeInternal {
			t.Errorf("%+v: %v, want internal", r, err)
		}
	}
}

// TestClassifyRequestRefusals: the state machine's refusal (42501 CONSTRAINT enrollment_request_transition, also the
// decider/revoker-is-the-actor rule) is CodeEnrollmentRequestTransition; the bodies' context refusal is
// context_required; the one_pending fallback (40001 enrollment_request_conflict) is unavailable like every class-40
// abort (never success, never a mapped refusal); an unnamed 42501 (a missing column privilege, e.g. a supplied
// authority_seq) stays internal; no driver text escapes.
func TestClassifyRequestRefusals(t *testing.T) {
	pg := func(code, constraint, msg string) *Error {
		return classify(fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: code, ConstraintName: constraint, Message: msg, Detail: "secret-canary-detail"}))
	}
	if e := pg("42501", "enrollment_request_transition", "enrollment_request_transition"); e == nil || e.Code != CodeEnrollmentRequestTransition ||
		e.Error() != "tenancy: enrollment_request_transition" {
		t.Errorf("transition: %v", e)
	}
	if e := pg("42501", "enrollment_context", "enrollment request context refused"); e == nil || e.Code != CodeContextRequired {
		t.Errorf("context: %v", e)
	}
	if e := pg("40001", "", "enrollment_request_conflict"); e == nil || e.Code != CodeUnavailable || e.Error() != "tenancy: unavailable" {
		t.Errorf("one_pending fallback: %v, want unavailable", e)
	}
	if e := pg("42501", "", "permission denied for table enrollment_requests"); e == nil || e.Code != CodeInternal {
		t.Errorf("unnamed 42501: %v, want internal", e)
	}
	if e := pg("P0001", "", "enrollment_request_conflict"); e == nil || e.Code != CodeInternal {
		t.Errorf("conflict message under another SQLSTATE: %v, want internal", e)
	}
}

// TestEnrollmentPolicyParse (G058 ENR-1 unit half, addendum §A.6): the mode is closed (an unknown or empty stored value
// reads as invite_only); open_ready holds only with both readiness columns set and exposes the basis; an unresolved or
// half-set pair is not ready with no basis (fail closed).
func TestEnrollmentPolicyParse(t *testing.T) {
	for in, want := range map[string]EnrollmentMode{"invite_only": EnrollmentInviteOnly, "request_access": EnrollmentRequestAccess,
		"open": EnrollmentOpen, "": EnrollmentInviteOnly, "Open": EnrollmentInviteOnly, "closed": EnrollmentInviteOnly, " open": EnrollmentInviteOnly} {
		if got := ParseEnrollmentMode(in); got != want || !got.Valid() {
			t.Errorf("ParseEnrollmentMode(%q) = %q, want %q", in, got, want)
		}
	}
	for _, m := range []EnrollmentMode{"", "Open", "closed"} {
		if m.Valid() {
			t.Errorf("%q is valid", m)
		}
	}
	at := pgtype.Timestamptz{Time: time.Unix(1700000000, 0), Valid: true}
	basis := pgtype.Text{String: "created_after_ledger", Valid: true}
	if p := ParseEnrollmentPolicy("invite_only", at, basis); p != (EnrollmentPolicy{Mode: EnrollmentInviteOnly, OpenReady: true, HistoryBasis: "created_after_ledger"}) {
		t.Errorf("resolved: %+v", p)
	}
	for name, c := range map[string]struct {
		at    pgtype.Timestamptz
		basis pgtype.Text
	}{
		"unresolved":  {pgtype.Timestamptz{}, pgtype.Text{}},
		"time only":   {at, pgtype.Text{}},
		"basis only":  {pgtype.Timestamptz{}, basis},
		"empty basis": {at, pgtype.Text{Valid: true}},
	} {
		if p := ParseEnrollmentPolicy("request_access", c.at, c.basis); p != (EnrollmentPolicy{Mode: EnrollmentRequestAccess}) {
			t.Errorf("%s: %+v, want not ready", name, p)
		}
	}
	if p := ParseEnrollmentPolicy("bogus", at, basis); p.Mode != EnrollmentInviteOnly || !p.OpenReady {
		t.Errorf("unknown mode: %+v", p)
	}
}

// TestEnrollmentPolicyTxRefusesBeforeSQL: a zero Organization is invalid_id for the read and the write, and a mode outside
// the closed set is invalid_enrollment_mode, before any statement (the transaction is nil, so a statement would panic).
func TestEnrollmentPolicyTxRefusesBeforeSQL(t *testing.T) {
	ctx := context.Background()
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	if _, err := ReadEnrollmentPolicyTx(ctx, nil, OrganizationID{}); CodeOf(err) != CodeInvalidID {
		t.Errorf("read zero organization: %v", err)
	}
	if _, err := SetEnrollmentModeTx(ctx, nil, OrganizationID{}, EnrollmentOpen); CodeOf(err) != CodeInvalidID {
		t.Errorf("write zero organization: %v", err)
	}
	for _, m := range []EnrollmentMode{"", "closed", "OPEN"} {
		if _, err := SetEnrollmentModeTx(ctx, nil, org, m); CodeOf(err) != CodeInvalidEnrollmentMode {
			t.Errorf("mode %q: %v, want invalid_enrollment_mode", m, err)
		}
	}
}

// TestClassifyEnrollmentPolicyCheck: the open-requires-ready CHECK (23514) is enrollment_not_ready and the mode CHECK
// invalid_enrollment_mode; the guard refusals (42501 without a mapped constraint) and any other check stay internal.
func TestClassifyEnrollmentPolicyCheck(t *testing.T) {
	check := func(code, constraint string) *Error {
		return classify(fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: code, ConstraintName: constraint, Detail: "secret-canary-detail"}))
	}
	if e := check("23514", "organizations_open_requires_resolved_history"); e == nil || e.Code != CodeEnrollmentNotReady || e.Error() != "tenancy: enrollment_not_ready" {
		t.Errorf("open-requires-ready: %v", e)
	}
	if e := check("23514", "organizations_enrollment_mode_valid"); e == nil || e.Code != CodeInvalidEnrollmentMode {
		t.Errorf("mode check: %v", e)
	}
	for _, c := range [][2]string{{"23514", "organizations_enrollment_history_shape"}, {"42501", ""}, {"42501", "enrollment_policy_owner"},
		{"23000", "organizations_open_requires_resolved_history"}} {
		if e := check(c[0], c[1]); e == nil || e.Code != CodeInternal {
			t.Errorf("%v: %v, want internal", c, e)
		}
	}
}

// modeQuerier answers the public mode statement with mode (or err) and counts statements.
type modeQuerier struct {
	mode  string
	err   error
	calls int
	sql   string
}

type modeRow struct{ q *modeQuerier }

func (r modeRow) Scan(dest ...any) error {
	if r.q.err != nil {
		return r.q.err
	}
	*dest[0].(*string) = r.q.mode
	return nil
}

func (q *modeQuerier) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	q.calls++
	q.sql = sql
	return modeRow{q}
}

// TestReadPublicEnrollmentMode (G058 HTTP slice): a zero Organization is invalid_id before any statement; the one
// statement is the policy reader's definer call; a stored value parses closed (an unknown value is invite_only); a
// store error passes through.
func TestReadPublicEnrollmentMode(t *testing.T) {
	ctx := context.Background()
	q := &modeQuerier{mode: "open"}
	if _, err := ReadPublicEnrollmentMode(ctx, q, OrganizationID{}); CodeOf(err) != CodeInvalidID || q.calls != 0 {
		t.Fatalf("zero Organization: %v (calls %d)", err, q.calls)
	}
	org := OrganizationID{1}
	for stored, want := range map[string]EnrollmentMode{"open": EnrollmentOpen, "request_access": EnrollmentRequestAccess,
		"invite_only": EnrollmentInviteOnly, "bogus": EnrollmentInviteOnly} {
		q := &modeQuerier{mode: stored}
		if m, err := ReadPublicEnrollmentMode(ctx, q, org); err != nil || m != want || q.calls != 1 || q.sql != sqlPublicEnrollmentMode {
			t.Errorf("%q: %q %v (calls %d)", stored, m, err, q.calls)
		}
	}
	if _, err := ReadPublicEnrollmentMode(ctx, &modeQuerier{err: fmt.Errorf("down")}, org); err == nil {
		t.Error("a store error was swallowed")
	}
}
