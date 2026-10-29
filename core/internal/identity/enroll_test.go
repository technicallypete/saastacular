package identity

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/audit"
)

// mintingProvider is a Provider whose Exchange returns vi and calls the input's minter mints times with the given
// projection (a foreign identity when foreign is set). Other methods are unused.
type mintingProvider struct {
	Provider
	vi       VerifiedIdentity
	foreign  *VerifiedIdentity
	mints    int
	address  string
	verified bool
	at       time.Time
}

func (p mintingProvider) Exchange(_ context.Context, in Exchange) (VerifiedIdentity, error) {
	bind := p.vi
	if p.foreign != nil {
		bind = *p.foreign
	}
	for i := 0; i < p.mints; i++ {
		in.Minter.Mint(bind, p.address, p.verified, p.at)
	}
	return p.vi, nil
}

// TestAddressEvidenceUnforgeable (G058 Plan r3 §5.1; readiness correction 01 §1 row 1): a zero minter mints a zero
// value; a struct literal is zero; a fresh minter binds the exchanged issuer/subject; a second mint poisons the minter;
// identity reads back exactly the one minted value per exchange; and in production code only the provider adapter calls
// Mint and only package identity constructs a minter (text pin over core/internal, the TestAccessBoundary pattern).
func TestAddressEvidenceUnforgeable(t *testing.T) {
	at := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	vi := VerifiedIdentity{Issuer: "https://issuer.test", Subject: "user_01UNF"}
	if ev := (EvidenceMinter{}).Mint(vi, "hana@identity.test", true, at); !ev.IsZero() {
		t.Fatalf("zero minter minted %+v", ev)
	}
	if !(AddressEvidence{}).IsZero() {
		t.Fatal("struct literal is not zero")
	}
	m := newEvidenceMinter()
	if ev := m.Mint(VerifiedIdentity{Subject: "user_01UNF"}, "hana@identity.test", true, at); !ev.IsZero() {
		t.Fatal("minted evidence for an identity without an issuer")
	}
	m = newEvidenceMinter()
	ev := m.Mint(vi, "hana@identity.test", true, at)
	if ev.IsZero() || ev.issuer != vi.Issuer || ev.subject != vi.Subject || !ev.verified || !ev.observedAt.Equal(at) || m.minted() != ev {
		t.Fatalf("minted %+v, read back %+v", ev, m.minted())
	}
	if again := m.Mint(vi, "other@identity.test", true, at); !again.IsZero() || !m.minted().IsZero() {
		t.Fatal("a second mint did not poison the minter")
	}

	// F3 (slice-1a round-1 review): a copied minter shares the one sink — a mint through the copy after a mint through
	// the original poisons both, so a copy cannot carry a second value past identity's read-back.
	orig := newEvidenceMinter()
	cp := orig
	if ev := orig.Mint(vi, "hana@identity.test", true, at); ev.IsZero() {
		t.Fatal("first mint through the original returned zero")
	}
	if ev := cp.Mint(vi, "copy@identity.test", true, at); !ev.IsZero() || !orig.minted().IsZero() || !cp.minted().IsZero() {
		t.Fatal("a mint through a copied minter did not poison the shared sink")
	}
	// F3: two concurrent completed mints (joined before the read-back) leave zero evidence, whatever their order.
	conc := newEvidenceMinter()
	done := make(chan struct{}, 2)
	for _, addr := range []string{"a@identity.test", "b@identity.test"} {
		go func(addr string, m EvidenceMinter) {
			m.Mint(vi, addr, true, at)
			done <- struct{}{}
		}(addr, conc)
	}
	<-done
	<-done
	if !conc.minted().IsZero() {
		t.Fatal("two concurrent mints left non-zero evidence")
	}

	// exchangeWithEvidence: one mint is returned; none or two yield zero evidence.
	for _, c := range []struct {
		name  string
		mints int
		zero  bool
	}{{"one", 1, false}, {"none", 0, true}, {"two", 2, true}} {
		s := &Service{cfg: Config{Provider: mintingProvider{vi: vi, mints: c.mints, address: "hana@identity.test", verified: true, at: at}}}
		got, ev, err := s.exchangeWithEvidence(context.Background(), "code", "verifier")
		if err != nil || got != vi || ev.IsZero() != c.zero {
			t.Errorf("%s mints: %+v %+v %v", c.name, got, ev, err)
		}
	}

	// Production call-site pins (non-test files under core/internal).
	root := filepath.Join("..")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src, rel := string(b), filepath.ToSlash(path)
		inIdentity := filepath.Dir(rel) == "../identity"
		if strings.Contains(src, ".Mint(") && rel != "../identity/workos/adapter.go" {
			t.Errorf("%s calls Mint (only the provider adapter may)", rel)
		}
		if strings.Contains(src, "newEvidenceMinter(") && !inIdentity {
			t.Errorf("%s constructs an evidence minter", rel)
		}
		if strings.Contains(src, "AddressEvidence{") && !inIdentity {
			t.Errorf("%s builds AddressEvidence", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join("workos", "adapter.go")); err != nil || strings.Count(string(b), ".Mint(") != 1 {
		t.Errorf("the adapter must call Mint exactly once (%v)", err)
	}
}

// TestAttestFromEvidenceRefusesBeforeAttestation (readiness correction 01 §1 row 2): zero, unverified, foreign-bound
// (issuer or subject) and zero-time evidence are recipient_proof_unavailable before AttestVerifiedAddress runs — the
// Service here has no store, so reaching the attestation would answer identity_unavailable instead. An address that
// does not normalize is refused by AttestVerifiedAddress's own validation (recipient_proof_unavailable); valid evidence
// reaches the store (identity_unavailable without one). The canonical-user comparison is covered against PostgreSQL.
func TestAttestFromEvidenceRefusesBeforeAttestation(t *testing.T) {
	at := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	vi := VerifiedIdentity{Issuer: "https://issuer.test", Subject: "user_01ATT"}
	h := EnrollmentHandle{UserID: "8c6f0a3e-1d2b-4c5a-9e7f-0a1b2c3d4e5f"}
	mint := func(bind VerifiedIdentity, address string, verified bool, at time.Time) AddressEvidence {
		return newEvidenceMinter().Mint(bind, address, verified, at)
	}
	s := &Service{}
	for _, c := range []struct {
		name string
		ev   AddressEvidence
		want error
	}{
		{"zero", AddressEvidence{}, ErrRecipientProofUnavailable},
		{"unverified", mint(vi, "hana@identity.test", false, at), ErrRecipientProofUnavailable},
		{"foreign issuer", mint(VerifiedIdentity{Issuer: "https://other.test", Subject: vi.Subject}, "hana@identity.test", true, at), ErrRecipientProofUnavailable},
		{"foreign subject", mint(VerifiedIdentity{Issuer: vi.Issuer, Subject: "user_01OTHER"}, "hana@identity.test", true, at), ErrRecipientProofUnavailable},
		{"zero time", mint(vi, "hana@identity.test", true, time.Time{}), ErrRecipientProofUnavailable},
		{"unnormalizable address", mint(vi, "not-an-address", true, at), ErrRecipientProofUnavailable},
		{"valid evidence reaches the store", mint(vi, "hana@identity.test", true, at), ErrUnavailable},
	} {
		va, err := s.attestFromEvidence(context.Background(), c.ev, vi, h)
		if !errors.Is(err, c.want) || !va.IsZero() {
			t.Errorf("%s: %v %+v, want %v", c.name, err, va, c.want)
		}
	}
}

// hostSpy is a HostResolver that records the Querier and origin it was asked about and answers res/err.
type hostSpy struct {
	res    HostResolution
	err    error
	q      Querier
	origin string
	calls  int
}

func (h *hostSpy) ResolveOrigin(_ context.Context, q Querier, origin string) (HostResolution, error) {
	h.q, h.origin = q, origin
	h.calls++
	return h.res, h.err
}

// spyQuerier is a Querier identity never reads through in these unit cases (the spy resolver ignores it).
type spyQuerier struct{ Querier }

// TestCheckEnrollmentHost (Codex round-2 host decision, identity half): the configured resolver is called exactly once
// on the caller's Querier with the credential's exact origin; a routing platform or custom host of the credential's
// Organization passes; another Organization, a non-routing or unknown-kind answer, a malformed handle and an
// unconfigured resolver are host_changed (the last two without a resolver call); a resolver failure is
// identity_unavailable; a nil service or Querier fails closed.
func TestCheckEnrollmentHost(t *testing.T) {
	ctx := context.Background()
	org := "8c6f0a3e-1d2b-4c5a-9e7f-0a1b2c3d4e5f"
	h := EnrollmentHandle{UserID: "u", OrganizationID: org, Origin: "https://acme.example.test", Digest: [32]byte{1}}
	q := spyQuerier{}
	for _, c := range []struct {
		name  string
		res   HostResolution
		err   error
		h     EnrollmentHandle
		want  error
		calls int
	}{
		{"platform host", HostResolution{Kind: HostPlatform, OrganizationID: org, Routing: true}, nil, h, nil, 1},
		{"custom host", HostResolution{Kind: HostCustom, OrganizationID: org, Routing: true}, nil, h, nil, 1},
		{"reassigned", HostResolution{Kind: HostPlatform, OrganizationID: "other", Routing: true}, nil, h, ErrHostChanged, 1},
		{"not routing", HostResolution{Kind: HostPlatform, OrganizationID: org}, nil, h, ErrHostChanged, 1},
		{"routes nowhere", HostResolution{}, nil, h, ErrHostChanged, 1},
		{"unknown kind", HostResolution{Kind: "relay", OrganizationID: org, Routing: true}, nil, h, ErrHostChanged, 1},
		{"resolver failure", HostResolution{}, errors.New("store"), h, ErrUnavailable, 1},
		{"no origin", HostResolution{}, nil, EnrollmentHandle{OrganizationID: org}, ErrHostChanged, 0},
		{"overlong origin", HostResolution{}, nil, EnrollmentHandle{OrganizationID: org, Origin: "https://" + strings.Repeat("a", 300)}, ErrHostChanged, 0},
		{"no organization", HostResolution{}, nil, EnrollmentHandle{Origin: h.Origin}, ErrHostChanged, 0},
	} {
		spy := &hostSpy{res: c.res, err: c.err}
		s := &Service{cfg: Config{Hosts: spy}}
		err := s.CheckEnrollmentHost(ctx, q, c.h)
		if !errors.Is(err, c.want) || (c.want == nil && err != nil) || spy.calls != c.calls {
			t.Errorf("%s: %v (calls %d), want %v (calls %d)", c.name, err, spy.calls, c.want, c.calls)
		}
		if c.calls == 1 && (spy.q != Querier(q) || spy.origin != c.h.Origin) {
			t.Errorf("%s: resolver asked on %v about %q", c.name, spy.q, spy.origin)
		}
	}
	if err := (&Service{}).CheckEnrollmentHost(ctx, q, h); !errors.Is(err, ErrHostChanged) {
		t.Errorf("unconfigured resolver: %v", err)
	}
	var nilSvc *Service
	if err := nilSvc.CheckEnrollmentHost(ctx, q, h); !errors.Is(err, ErrUnavailable) {
		t.Errorf("nil service: %v", err)
	}
	if err := (&Service{cfg: Config{Hosts: &hostSpy{}}}).CheckEnrollmentHost(ctx, nil, h); !errors.Is(err, ErrUnavailable) {
		t.Errorf("nil querier: %v", err)
	}
}

// fakeAdmitter is an Admitter that records its calls and answers out/err.
type fakeAdmitter struct {
	out   AdmitOutcome
	err   error
	calls int
	sel   AdmitSelector
}

func (f *fakeAdmitter) AdmitAtCallback(_ context.Context, _ EnrollmentHandle, sel AdmitSelector) (AdmitOutcome, error) {
	f.calls++
	f.sel = sel
	return f.out, f.err
}

// TestAdmitAtCallbackOrchestration (identity's T2 half): without a composed Admitter every admission is
// identity_unavailable; a malformed selector is invalid_request; a recipient-proof refusal stops before the Admitter
// (recipient_proof_unavailable, the consumed callback's auth.sign_in denied reason); the open kind never forwards an
// invitation or address; a returned status passes through unchanged (the Admitter recorded it) and an Admitter error is
// identity_unavailable. A client-supplied Address is always discarded.
func TestAdmitAtCallbackOrchestration(t *testing.T) {
	ctx := context.Background()
	vi := VerifiedIdentity{Issuer: "https://issuer.test", Subject: "user_01ADM"}
	h := EnrollmentHandle{UserID: "8c6f0a3e-1d2b-4c5a-9e7f-0a1b2c3d4e5f"}
	inv := InvitationRef{9}
	forged := &VerifiedAddress{}
	if _, err := (&Service{}).admitAtCallback(ctx, h, vi, AddressEvidence{}, AdmitSelector{Kind: AdmitKindOpen}); !errors.Is(err, ErrUnavailable) {
		t.Errorf("no admitter: %v", err)
	}
	for name, sel := range map[string]AdmitSelector{
		"unknown kind":          {Kind: "lift"},
		"invitation without id": {Kind: AdmitKindInvitation, InvitationToken: "t"},
		"invitation no token":   {Kind: AdmitKindInvitation, Invitation: &inv},
	} {
		a := &fakeAdmitter{}
		if _, err := (&Service{admitter: a}).admitAtCallback(ctx, h, vi, AddressEvidence{}, sel); !errors.Is(err, ErrInvalidRequest) || a.calls != 0 {
			t.Errorf("%s: %v (admitter calls %d)", name, err, a.calls)
		}
	}
	a := &fakeAdmitter{out: AdmitOutcome{Status: AdmitAdmitted}}
	_, err := (&Service{admitter: a}).admitAtCallback(ctx, h, vi, AddressEvidence{},
		AdmitSelector{Kind: AdmitKindInvitation, Invitation: &inv, InvitationToken: "t", Address: forged})
	if !errors.Is(err, ErrRecipientProofUnavailable) || a.calls != 0 {
		t.Errorf("zero evidence: %v (admitter calls %d)", err, a.calls)
	}
	want, _ := audit.ParseUUID(h.UserID)
	if e := signInFailure(err, h.UserID); e.Outcome != audit.Denied || e.Reason != audit.ReasonRecipientProofUnavailable || e.Subject != want {
		t.Errorf("proof failure event: %+v", e)
	}
	if e := signInFailure(err, ""); e.Outcome != audit.Denied || e.Reason != audit.ReasonRecipientProofUnavailable || e.Subject != (audit.UUID{}) {
		t.Errorf("unattributed proof failure event: %+v", e)
	}
	for _, c := range []struct {
		name   string
		a      *fakeAdmitter
		status string
		err    error
	}{
		{"admitted", &fakeAdmitter{out: AdmitOutcome{Status: AdmitAdmitted, BindingID: "b"}}, AdmitAdmitted, nil},
		{"recorded refusal", &fakeAdmitter{out: AdmitOutcome{Status: AdmitEnrollmentClosed}}, AdmitEnrollmentClosed, nil},
		{"admitter error", &fakeAdmitter{err: errors.New("access: unavailable")}, "", ErrUnavailable},
	} {
		out, err := (&Service{admitter: c.a}).admitAtCallback(ctx, h, vi, AddressEvidence{},
			AdmitSelector{Kind: AdmitKindOpen, Invitation: &inv, InvitationToken: "t", Address: forged})
		if out.Status != c.status || !errors.Is(err, c.err) || (c.err == nil && err != nil) || c.a.calls != 1 {
			t.Errorf("%s: %+v %v (calls %d)", c.name, out, err, c.a.calls)
		}
		if c.a.sel.Address != nil || c.a.sel.Invitation != nil || c.a.sel.InvitationToken != "" || c.a.sel.Kind != AdmitKindOpen {
			t.Errorf("%s: open admission forwarded %+v", c.name, c.a.sel)
		}
	}
	// G058 HTTP slice (request_on_arrival): the request kind forwards no invitation or address either; its pending status
	// (or a recorded refusal) passes through, an Admitter error is identity_unavailable.
	for _, c := range []struct {
		name   string
		a      *fakeAdmitter
		status string
		err    error
	}{
		{"request submitted", &fakeAdmitter{out: AdmitOutcome{Status: EnrollmentPending}}, EnrollmentPending, nil},
		{"request refused", &fakeAdmitter{out: AdmitOutcome{Status: AdmitEnrollmentClosed}}, AdmitEnrollmentClosed, nil},
		{"request unknown outcome", &fakeAdmitter{err: errors.New("access: unavailable")}, "", ErrUnavailable},
	} {
		out, err := (&Service{admitter: c.a}).admitAtCallback(ctx, h, vi, AddressEvidence{},
			AdmitSelector{Kind: AdmitKindRequest, Invitation: &inv, InvitationToken: "t", Address: forged})
		if out.Status != c.status || !errors.Is(err, c.err) || (c.err == nil && err != nil) || c.a.calls != 1 {
			t.Errorf("%s: %+v %v (calls %d)", c.name, out, err, c.a.calls)
		}
		if c.a.sel.Address != nil || c.a.sel.Invitation != nil || c.a.sel.InvitationToken != "" || c.a.sel.Kind != AdmitKindRequest {
			t.Errorf("%s: request forwarded %+v", c.name, c.a.sel)
		}
	}

	// T2 kind selection (Codex round-4 correction): an explicit stored invitation intent selects the invitation kind
	// whether or not the invitation is still pending, so a stale (expired, revoked or consumed) invitation on an open
	// Organization is adjudicated by the claim and never becomes an open admission; only a login without an invitation
	// is admitted through open enrollment. enrollResult.Invitation is T1's StoredInvitation != "".
	tok := strings.Repeat("A", 43)
	withInv := CallbackInput{Invitation: &inv, InvitationToken: tok}
	for _, c := range []struct {
		name string
		r    enrollResult
		in   CallbackInput
		kind AdmitKind
	}{
		{"stale expired invitation, open Organization", enrollResult{Mode: "open", Invitation: true}, withInv, AdmitKindInvitation},
		{"stale revoked invitation, open Organization", enrollResult{Mode: "open", Invitation: true}, withInv, AdmitKindInvitation},
		{"stale consumed invitation, open Organization", enrollResult{Mode: "open", Invitation: true}, withInv, AdmitKindInvitation},
		{"stale invitation, request_access Organization", enrollResult{Mode: "request_access", Invitation: true}, withInv, AdmitKindInvitation},
		{"pending invitation, invite_only Organization", enrollResult{Mode: "invite_only", Invitation: true}, withInv, AdmitKindInvitation},
		{"no invitation, open Organization (positive)", enrollResult{Mode: "open"}, CallbackInput{}, AdmitKindOpen},
		{"no invitation, request_access Organization", enrollResult{Mode: "request_access"}, CallbackInput{}, ""},
		// request_on_arrival selects the request kind on request_access only; an invitation intent or open mode wins, and
		// invite_only (a pending invitation admitted T1) never submits a request.
		{"request on arrival, request_access Organization", enrollResult{Mode: "request_access", Request: true}, CallbackInput{}, AdmitKindRequest},
		{"request on arrival, open Organization", enrollResult{Mode: "open", Request: true}, CallbackInput{}, AdmitKindOpen},
		{"request on arrival with an invitation", enrollResult{Mode: "request_access", Invitation: true, Request: true}, withInv, AdmitKindInvitation},
		{"request on arrival, invite_only Organization", enrollResult{Mode: "invite_only", Request: true}, CallbackInput{}, ""},
	} {
		sel := callbackSelector(c.r, c.in)
		if sel.Kind != c.kind || sel.Address != nil {
			t.Errorf("%s: selector %+v, want kind %q", c.name, sel, c.kind)
		}
		switch c.kind {
		case AdmitKindInvitation:
			if sel.Invitation == nil || *sel.Invitation != inv || sel.InvitationToken != tok {
				t.Errorf("%s: invitation intent not forwarded: %+v", c.name, sel)
			}
		default:
			if sel.Invitation != nil || sel.InvitationToken != "" {
				t.Errorf("%s: invitation forwarded without intent: %+v", c.name, sel)
			}
		}
	}
}

// TestEnrollmentCallbackInputs (Codex round-3 decisions 1-2, unit half): the Admitter binds once and never as nil (a
// refused binding changes nothing); the invitation id renders canonically and a forwarded token has G019's form; a
// malformed forwarded intent and a login carrying enrollment facts without the intent are refused before any database
// work; a T1 policy/intent denial is one access.denied enrollment.start event without an Organization, Group or resource.
func TestEnrollmentCallbackInputs(t *testing.T) {
	ctx := context.Background()
	var nilSvc *Service
	a := &fakeAdmitter{}
	if err := nilSvc.BindAdmitter(a); err == nil {
		t.Error("nil service bound an admitter")
	}
	s := &Service{}
	if err := s.BindAdmitter(nil); err == nil || s.admitter != nil {
		t.Errorf("nil admitter: %v (%v)", err, s.admitter)
	}
	if err := s.BindAdmitter(a); err != nil || s.admitter != Admitter(a) {
		t.Fatalf("first binding: %v", err)
	}
	if err := s.BindAdmitter(&fakeAdmitter{}); err == nil || s.admitter != Admitter(a) {
		t.Errorf("second binding: %v", err)
	}

	ref := InvitationRef{0x0f, 0x58, 0xe0, 0, 0, 0, 0x40, 0, 0x80, 0, 0, 0, 0, 0, 0, 0x2a}
	if got := invitationString(ref); got != "0f58e000-0000-4000-8000-00000000002a" {
		t.Errorf("invitation id rendered %q", got)
	}
	good := strings.Repeat("A", 43)
	for tok, want := range map[string]bool{"": true, good: true, strings.Repeat("A", 42): false, strings.Repeat("A", 44): false,
		strings.Repeat("A", 42) + "=": false, strings.Repeat("A", 42) + "+": false} {
		if invitationTokenForm(tok) != want {
			t.Errorf("token form %q: want %v", tok, want)
		}
	}

	db := &countingDB{}
	svc := NewService(newStore(db), Config{})
	for name, in := range map[string]CallbackInput{
		"token without invitation": {RawQuery: "code=c&state=s", InvitationToken: good},
		"invitation without token": {RawQuery: "code=c&state=s", Invitation: &ref},
		"malformed token":          {RawQuery: "code=c&state=s", Invitation: &ref, InvitationToken: "short"},
	} {
		if _, err := svc.Callback(ctx, in); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("%s: %v", name, err)
		}
	}
	hosts := &hostSpy{}
	login := NewService(newStore(db), Config{Provider: mintingProvider{}, Hosts: hosts})
	for name, in := range map[string]LoginInput{
		"invitation without intent": {Origin: "https://acme.example.test", Invitation: &ref},
		"request without intent":    {Origin: "https://acme.example.test", Request: true},
	} {
		if _, err := login.Login(ctx, in); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if db.begins != 0 || db.queries != 0 || hosts.calls != 0 {
		t.Errorf("database or resolver reached before the refusals: %+v, resolver %d", db, hosts.calls)
	}

	for _, err := range []error{ErrEnrollmentClosed, ErrInvitationMismatch} {
		e := startDenial(err)
		if e.Action != audit.AccessDenied || e.Outcome != audit.Denied || e.Operation != audit.OpEnrollmentStart ||
			string(e.Reason) != AsError(err).Code || e.Organization != (audit.UUID{}) || e.Group != (audit.UUID{}) ||
			e.ResourceType != "" || e.Subject != (audit.UUID{}) {
			t.Errorf("start denial %v: %+v", err, e)
		}
		if !slicesContains(audit.Reasons, e.Reason) {
			t.Errorf("start denial reason %s outside the catalog", e.Reason)
		}
	}
}

// credentialRowDB answers app.auth_resolve_enrollment with one scripted row (or err) and records its arguments.
type credentialRowDB struct {
	countingDB
	origin string
	err    error
	args   []any
}

type scriptedRow struct{ scan func(...any) error }

func (r scriptedRow) Scan(dest ...any) error { return r.scan(dest...) }

func (d *credentialRowDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	d.queries++
	d.args = args
	return scriptedRow{func(dest ...any) error {
		if !strings.Contains(sql, "app.auth_resolve_enrollment($1, $2)") {
			return errors.New("credentialRowDB: unexpected statement")
		}
		if d.err != nil {
			return d.err
		}
		*dest[0].(*string), *dest[1].(*string), *dest[2].(*string) = "s", "8c6f0a3e-1d2b-4c5a-9e7f-0a1b2c3d4e5f", "0f58a000-0000-4000-8000-0000000000a1"
		*dest[3].(*string), *dest[4].(*[]byte), *dest[5].(*time.Time) = d.origin, make([]byte, 32), time.Unix(1800, 0)
		return nil
	}}
}

// TestEnrollmentCredential (G058 HTTP slice; Plan r3 §1.1 enrollment credential, ENR-4/ENR-21 unit half): a malformed
// credential or an absent/overlong origin is unauthenticated before any statement; one auth reader statement resolves a
// live row by the credential's digest; no row, or a row minted for another origin, is unauthenticated (a credential is
// bound to its exact origin, never usable on another host); a store failure is identity_unavailable; the handle carries
// the digest, never the raw credential. The CSRF matcher is constant-time over the decoded token and refuses anything
// but the exact 32 bytes.
func TestEnrollmentCredential(t *testing.T) {
	ctx := context.Background()
	good := encodeSecret(make([]byte, 32))
	db := &countingDB{}
	svc := NewService(newStore(db), Config{})
	for name, c := range map[string][2]string{
		"empty credential":     {"", "https://acme.example.test"},
		"malformed credential": {"not a secret", "https://acme.example.test"},
		"short credential":     {strings.Repeat("A", 42), "https://acme.example.test"},
		"no origin":            {good, ""},
		"overlong origin":      {good, "https://" + strings.Repeat("a", 300)},
	} {
		if _, err := svc.EnrollmentCredential(ctx, c[0], c[1]); !errors.Is(err, ErrUnauthenticated) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if db.queries != 0 || db.begins != 0 {
		t.Fatalf("database reached before the refusals: %+v", db)
	}
	const origin = "https://acme.example.test"
	rdb := &credentialRowDB{origin: origin}
	info, err := NewService(newStore(rdb), Config{}).EnrollmentCredential(ctx, good, origin)
	raw, _ := decodeSecret(good)
	var want [32]byte
	copy(want[:], hashSecret(raw))
	if err != nil || info.Handle.Origin != origin || info.Handle.Digest != want || info.Handle.UserID == "" ||
		info.Handle.OrganizationID == "" || len(info.CSRFToken) != 32 || rdb.queries != 1 || rdb.begins != 0 {
		t.Fatalf("live credential: %+v %v (queries %d)", info, err, rdb.queries)
	}
	if b, ok := rdb.args[0].([]byte); !ok || string(b) == good || len(b) != 32 {
		t.Errorf("the statement received %T, not the 32-byte digest", rdb.args[0])
	}
	for name, d := range map[string]*credentialRowDB{
		"another origin": {origin: "https://other.example.test"},
		"no row":         {err: pgx.ErrNoRows},
	} {
		if _, err := NewService(newStore(d), Config{}).EnrollmentCredential(ctx, good, origin); !errors.Is(err, ErrUnauthenticated) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := NewService(newStore(&credentialRowDB{err: errors.New("down")}), Config{}).EnrollmentCredential(ctx, good, origin); !errors.Is(err, ErrUnavailable) {
		t.Errorf("store failure: %v", err)
	}
	tok := make([]byte, 32)
	tok[0] = 7
	ok := EnrollmentCredentialInfo{CSRFToken: tok}
	for token, want := range map[string]bool{encodeSecret(tok): true, encodeSecret(make([]byte, 32)): false, "": false,
		encodeSecret(tok) + "=": false, encodeSecret(tok)[:42]: false} {
		if EnrollmentCSRFMatches(ok, token) != want {
			t.Errorf("csrf %q: want %v", token, want)
		}
	}
	if EnrollmentCSRFMatches(EnrollmentCredentialInfo{}, encodeSecret(tok)) {
		t.Error("a credential without a CSRF token matched")
	}
}

func slicesContains[T comparable](s []T, v T) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
