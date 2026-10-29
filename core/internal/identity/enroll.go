package identity

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/audit"
)

// G058 enrollment (Plan r3 §1.3, §5.1; readiness correction 01 §1; Codex round-2/round-3 callback decisions). This file
// holds identity's side of the enrollment-intent callback: the sealed per-exchange AddressEvidence is verified against
// the exchanged identity before the merged G019 AttestVerifiedAddress builds the VerifiedAddress that access consumes;
// the admission host guard (CheckEnrollmentHost); the callback's first transaction (store.enroll, T1: continuity,
// invitation intent, host, policy, mapping, the enrollment credential) and its admission hand-off (admitAtCallback, T2)
// through the Admitter bound at startup (BindAdmitter).

// evidenceSourceWorkOS is the evidence kind AttestVerifiedAddress records for adapter-minted evidence.
const evidenceSourceWorkOS = "workos_authenticate"

// EnrollmentSessionLifetime is the absolute lifetime of an enrollment credential (Plan r3 §1.7; the table's
// enrollment_sessions_lifetime CHECK).
const EnrollmentSessionLifetime = 30 * time.Minute

// enrollmentReturnPath is the one return path of an enrollment-intent login (00016 login_transactions_return_path_check).
const enrollmentReturnPath = "/join/status"

// The T1 statements (Plan r3 §1.3 step 2): the policy reader's public mode and the claim writer's pending-invitation
// predicate (both SECURITY DEFINER, readable before any context), and the issuing INSERT of the credential row in the
// IDENTITY shape of the mapped subject (enrollment_sessions_insert; the issuing columns only).
const (
	sqlEnrollmentPolicy = `SELECT app.rls_enrollment_mode($1::uuid), app.rls_invitation_pending(NULLIF($2, '')::uuid, $1::uuid)`
	sqlInsertEnrollment = `INSERT INTO app.enrollment_sessions (id, user_id, organization_id, origin, credential_hash, csrf_token, ` +
		`provider_session_id, created_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
)

// errAdmitterBinding refuses a nil Admitter or a second binding.
var errAdmitterBinding = errors.New("identity: admitter binding refused")

// BindAdmitter binds the callback-time admission service once, at startup and before serving (G058 Codex round-3
// decision 2: composed by core/internal/enrollment/api). A nil Admitter or a second binding is refused and changes
// nothing; until a binding succeeds every admission is identity_unavailable (fail closed).
func (s *Service) BindAdmitter(a Admitter) error {
	if s == nil || a == nil || s.admitter != nil {
		return errAdmitterBinding
	}
	s.admitter = a
	return nil
}

// invitationString is the canonical lowercase hyphenated form of an invitation id.
func invitationString(r InvitationRef) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", r[0:4], r[4:6], r[6:8], r[8:10], r[10:16])
}

// invitationTokenForm reports whether a forwarded invitation token is absent or has G019's form (43 base64url
// characters of 32 bytes, tenancy.ValidInvitationToken's rule; identity does not import tenancy).
func invitationTokenForm(token string) bool {
	if token == "" {
		return true
	}
	if len(token) != 43 {
		return false
	}
	b, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(b) == 32
}

// enrollParams are the inputs of the callback's first enrollment transaction.
type enrollParams struct {
	Identity     VerifiedIdentity
	Live         *SessionInfo // the forwarded live session bound to this origin and Organization, if any
	Now          time.Time
	Origin       string
	Organization string // pinned on the login transaction
	Hosts        HostResolver
	// The stored login's invitation (txn.invitation_id) and the BFF-forwarded one.
	StoredInvitation    string
	ForwardedInvitation *InvitationRef
	// RequestOnArrival is the stored login's request_on_arrival (Plan r3 §1.3 steps 1 and 3).
	RequestOnArrival bool
}

// The next step after T1.
const (
	enrollNext     = iota // an enrollment credential committed: continue to T2
	enrollSignIn          // the subject is already signed in on this host or already a member: the ordinary sign-in
	enrollConflict        // a live session for another user: nothing committed (identity_conflict)
)

// enrollResult is T1's committed effect.
type enrollResult struct {
	Next       int
	Handle     EnrollmentHandle
	Credential []byte // raw enrollment credential (enrollNext only)
	ExpiresAt  time.Time
	Mode       string // the enrollment mode T1 read (the T2 kind selector; the admission body rechecks it)
	// Invitation is the explicit stored invitation intent of the login (txn.invitation_id set), independent of whether
	// the invitation is still pending: an expired, revoked or consumed invitation is adjudicated by T2's claim and never
	// becomes an open admission (Codex round-4 correction).
	Invitation bool
	// Request is the stored login's request_on_arrival: on a request_access Organization T2 submits the user's own access
	// request (the request kind); it never admits.
	Request bool
	Subject string // the mapped subject when established (attribution of a T1 denial)
}

// enroll is the callback's first enrollment transaction (T1, Plan r3 §1.3 step 2), in this order: the namespace binding
// (as signIn), the mapped subject (E2); (a) continuity: a live session for another principal is identity_conflict with
// nothing created, the same subject continues with the ordinary sign-in; (b) the forwarded invitation must equal the
// stored one (invitation_mismatch); (c) the host must still route to the login's Organization on this transaction
// (host_changed); (d) the Organization admits the login only when its mode is request_access or open or the stored
// invitation is pending (enrollment_closed; nothing created); then a current member continues with the ordinary
// sign-in, (e) an unknown principal is mapped (createMapping), and (g) the enrollment credential row is inserted in the
// IDENTITY shape of the mapped subject and committed. No provider fact but the provider session id (the logout-scope
// key) is stored. Every refusal rolls back; an unknown COMMIT is identity_unavailable.
func (s *Store) enroll(ctx context.Context, p enrollParams) (enrollResult, error) {
	var res enrollResult
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initEmpty(ctx, tx); err != nil {
			return err
		}
		vi := p.Identity
		if _, err := tx.Exec(ctx, `
INSERT INTO app.identity_provider_binding (provider, issuer, client_id, bound_at) VALUES ('workos', $1, $2, $3)
ON CONFLICT (provider) DO NOTHING`, vi.Issuer, vi.ClientID, p.Now); err != nil {
			return unavailable(err)
		}
		var boundIssuer, boundClient string
		if err := tx.QueryRow(ctx, `SELECT issuer, client_id FROM app.identity_provider_binding WHERE provider = 'workos'`).
			Scan(&boundIssuer, &boundClient); err != nil {
			return unavailable(err)
		}
		if boundIssuer != vi.Issuer || boundClient != vi.ClientID {
			return ErrNamespaceMismatch
		}
		mapped, err := principalSubject(ctx, tx, vi.Issuer, vi.Subject)
		if err != nil {
			return err
		}
		res.Subject = mapped
		// (a) continuity first.
		if p.Live != nil {
			if mapped == "" || mapped != p.Live.UserID {
				res.Next = enrollConflict
				return errNoCommit
			}
			res.Next = enrollSignIn
			return errNoCommit
		}
		// (b) the invitation intent is bound to the stored login transaction.
		fwd := ""
		if p.ForwardedInvitation != nil {
			fwd = invitationString(*p.ForwardedInvitation)
		}
		if fwd != p.StoredInvitation {
			return ErrInvitationMismatch
		}
		// (c) the host, on this transaction.
		if p.Hosts == nil || p.Organization == "" {
			return ErrHostChanged
		}
		hr, err := p.Hosts.ResolveOrigin(ctx, tx, p.Origin)
		if err != nil {
			return unavailable(err)
		}
		if !hr.Routing || hr.OrganizationID != p.Organization {
			return ErrHostChanged
		}
		// (d) the Organization's policy (or a pending invitation of it) must admit this login.
		var mode string
		var pending bool
		if err := tx.QueryRow(ctx, sqlEnrollmentPolicy, p.Organization, p.StoredInvitation).Scan(&mode, &pending); err != nil {
			return unavailable(err)
		}
		if mode != "request_access" && mode != "open" && !pending {
			return ErrEnrollmentClosed
		}
		// The gate above admits on mode or a pending invitation; the T2 selector keeps the stored intent itself.
		res.Mode, res.Invitation, res.Request = mode, p.StoredInvitation != "", p.RequestOnArrival
		// (f) a current member signs in the ordinary way (nothing is created here).
		if mapped != "" {
			var member bool
			if err := tx.QueryRow(ctx, sqlIsOrganizationMember, mapped, p.Organization).Scan(&member); err != nil {
				return unavailable(err)
			}
			if member {
				res.Next = enrollSignIn
				return errNoCommit
			}
			if err := initIdentity(ctx, tx, mapped); err != nil {
				return err
			}
		} else if mapped, err = s.createMapping(ctx, tx, vi, p.Now); err != nil {
			// (e) the first local identity of an admitted principal.
			return err
		}
		res.Subject = mapped
		// (g) the enrollment credential.
		if vi.ProviderSessionID == "" || len(vi.ProviderSessionID) > 255 {
			return ErrUnavailable
		}
		id, err := newUserID()
		if err != nil {
			return ErrUnavailable
		}
		cred, err := newSecret()
		if err != nil {
			return ErrUnavailable
		}
		csrf, err := newSecret()
		if err != nil {
			return ErrUnavailable
		}
		expires := p.Now.Add(EnrollmentSessionLifetime)
		digest := hashSecret(cred)
		if _, err := tx.Exec(ctx, sqlInsertEnrollment, id, mapped, p.Organization, p.Origin, digest, csrf, vi.ProviderSessionID,
			p.Now, expires); err != nil {
			return unavailable(err)
		}
		res.Next, res.Credential, res.ExpiresAt = enrollNext, cred, expires
		res.Handle = EnrollmentHandle{UserID: mapped, OrganizationID: p.Organization, Origin: p.Origin}
		copy(res.Handle.Digest[:], digest)
		return nil
	})
	switch {
	case errors.Is(err, errNoCommit):
		return res, nil
	case err != nil:
		return enrollResult{Subject: res.Subject}, err
	}
	return res, nil
}

// startDenial is the access.denied enrollment.start event of a T1 policy/intent denial (Codex round-2/3 decisions): no
// Organization, Group or resource; attributed by the recorder's IDENTITY shape only when the callback established the
// mapped subject.
func startDenial(err error) audit.Event {
	return audit.Event{Action: audit.AccessDenied, Outcome: audit.Denied, Reason: audit.Reason(AsError(err).Code),
		Operation: audit.OpEnrollmentStart}
}

// enrollCallback is the callback's enrollment branch after the exchange (Plan r3 §1.3 steps 2-3). T1 (store.enroll);
// a T1 policy/intent denial (enrollment_closed, invitation_mismatch) is recorded once as access.denied enrollment.start
// and *recorded is set, so no auth.sign_in failure is added; other T1 failures keep their auth.sign_in ownership. signIn
// is true when the subject continues with the unchanged ordinary sign-in. After a committed T1, T2 runs through the bound
// Admitter (callbackSelector): the invitation kind whenever the login carries a stored invitation intent, open when the
// mode was open, the request kind on request_access when the login carried request_on_arrival (the user's own access
// request; success is status pending), else no admission and status pending. A confirmed admission answers its status
// with next sign_in and no credential (status admitted is the historical T2 commitment, never a session: the browser
// then starts the ordinary sign-in, whose guards decide); pending or a known refusal (recorded by the admitting
// service) answers its status with the enrollment credential; a recipient-proof failure or identity conflict is the
// callback's auth.sign_in failure attributed to the established subject; any other failure is identity_unavailable.
func (s *Service) enrollCallback(ctx context.Context, in CallbackInput, txn loginTxn, vi VerifiedIdentity, ev AddressEvidence,
	live *SessionInfo, now time.Time, subject *string, recorded *bool) (CallbackOutput, bool, error) {
	r, err := s.store.enroll(ctx, enrollParams{Identity: vi, Live: live, Now: now, Origin: in.Origin,
		Organization: txn.OrganizationID, Hosts: s.cfg.Hosts, StoredInvitation: txn.InvitationID,
		ForwardedInvitation: in.Invitation, RequestOnArrival: txn.RequestOnArrival})
	switch {
	case errors.Is(err, ErrEnrollmentClosed), errors.Is(err, ErrInvitationMismatch):
		*recorded = true
		s.store.recordIndependent(ctx, r.Subject, startDenial(err))
		return CallbackOutput{}, false, err
	case err != nil:
		return CallbackOutput{}, false, err
	}
	switch r.Next {
	case enrollSignIn:
		return CallbackOutput{}, true, nil
	case enrollConflict:
		*subject = live.UserID
		s.revokeExtraProviderSession(ctx, vi.ProviderSessionID)
		return CallbackOutput{}, false, ErrIdentityConflict
	}
	sel := callbackSelector(r, in)
	pending := CallbackOutput{Outcome: OutcomeEnrollmentStatus, ReturnTo: txn.ReturnPath, MaxAge: int(r.ExpiresAt.Sub(now) / time.Second),
		Enrollment: &EnrollmentOutcome{Status: EnrollmentPending, Credential: encodeSecret(r.Credential)}}
	if sel.Kind == "" {
		return pending, false, nil
	}
	out, err := s.admitAtCallback(ctx, r.Handle, vi, ev, sel)
	if err != nil {
		*subject = r.Handle.UserID
		return CallbackOutput{}, false, err
	}
	switch out.Status {
	case AdmitAdmitted, AdmitAlreadyMember:
		return CallbackOutput{Outcome: OutcomeEnrollmentStatus, ReturnTo: txn.ReturnPath,
			Enrollment: &EnrollmentOutcome{Status: out.Status, Next: NextSignIn}}, false, nil
	}
	pending.Enrollment.Status = out.Status
	return pending, false, nil
}

// callbackSelector is T2's kind for a committed T1: the invitation kind whenever the login carries an explicit stored
// invitation intent (pending or not: the claim adjudicates an expired, revoked or consumed invitation and returns its
// closed refusal, even on an open Organization), else open when T1 read mode open, else the request kind when T1 read
// request_access and the login carried request_on_arrival, else none (no write; status pending).
func callbackSelector(r enrollResult, in CallbackInput) AdmitSelector {
	switch {
	case r.Invitation:
		return AdmitSelector{Kind: AdmitKindInvitation, Invitation: in.Invitation, InvitationToken: in.InvitationToken}
	case r.Mode == "open":
		return AdmitSelector{Kind: AdmitKindOpen}
	case r.Mode == "request_access" && r.Request:
		return AdmitSelector{Kind: AdmitKindRequest}
	}
	return AdmitSelector{}
}

// exchangeWithEvidence redeems the code with a fresh sealed minter and returns the validated identity together with the
// evidence the provider adapter minted for this exchange (the zero value when it minted none or more than one).
func (s *Service) exchangeWithEvidence(ctx context.Context, code, verifier string) (VerifiedIdentity, AddressEvidence, error) {
	m := newEvidenceMinter()
	vi, err := s.cfg.Provider.Exchange(ctx, Exchange{Code: code, CodeVerifier: verifier, Minter: m})
	if err != nil {
		return VerifiedIdentity{}, AddressEvidence{}, err
	}
	return vi, m.minted(), nil
}

// attestFromEvidence verifies ev for the exchanged identity vi and the enrollment handle h, in this order and before any
// attestation: (a) zero evidence, (b) an address the provider did not report verified, (c) evidence bound to another
// issuer or subject than the exchanged identity (a copied or foreign value), (d) a zero observation time — each is
// recipient_proof_unavailable with nothing attested or written. Only then (e) the merged AttestVerifiedAddress
// re-validates the inputs, normalizes the address and resolves the canonical user itself; an address that does not
// normalize is recipient_proof_unavailable, an unmapped principal identity_conflict, a store failure
// identity_unavailable. Finally (f) the resolved canonical user must be the enrollment handle's user, else
// identity_conflict. The returned VerifiedAddress is the only evidence access receives (no client proof exists).
func (s *Service) attestFromEvidence(ctx context.Context, ev AddressEvidence, vi VerifiedIdentity, h EnrollmentHandle) (VerifiedAddress, error) {
	switch {
	case ev.IsZero(), !ev.verified, ev.issuer != vi.Issuer || ev.subject != vi.Subject, ev.observedAt.IsZero():
		return VerifiedAddress{}, ErrRecipientProofUnavailable
	}
	va, err := s.AttestVerifiedAddress(ctx, ev.issuer, ev.subject, ev.address, ev.observedAt, evidenceSourceWorkOS)
	switch {
	case errors.Is(err, ErrInvalidRequest):
		return VerifiedAddress{}, ErrRecipientProofUnavailable
	case errors.Is(err, ErrUnauthenticated):
		return VerifiedAddress{}, ErrIdentityConflict
	case err != nil:
		return VerifiedAddress{}, ErrUnavailable
	}
	if h.UserID == "" || va.UserID() != h.UserID {
		return VerifiedAddress{}, ErrIdentityConflict
	}
	return va, nil
}

// CheckEnrollmentHost is the final host guard of an enrollment admission (Plan r3 §1.3 step 3, §1.4; Codex round-2 host
// decision). The admitting service calls it inside its own admission transaction q, before any invitation claim or
// admission write: the enrollment credential's exact stored origin h.Origin (equal to the request origin by the
// credential's resolution) must still route, through the configured resolver and on q itself, to the credential's
// Organization. A malformed handle, an unconfigured resolver, or an origin that routes nowhere or to another
// Organization is host_changed; a resolver (store) failure is identity_unavailable. Like issuanceGuard it is a plain READ
// COMMITTED snapshot read without a lock: a host change committing after the read is not covered at commit time.
func (s *Service) CheckEnrollmentHost(ctx context.Context, q Querier, h EnrollmentHandle) error {
	if s == nil || q == nil {
		return ErrUnavailable
	}
	if s.cfg.Hosts == nil || h.Origin == "" || len(h.Origin) > 300 || h.OrganizationID == "" {
		return ErrHostChanged
	}
	res, err := s.cfg.Hosts.ResolveOrigin(ctx, q, h.Origin)
	if err != nil {
		return ErrUnavailable
	}
	if !res.Routing || res.OrganizationID != h.OrganizationID || (res.Kind != HostPlatform && res.Kind != HostCustom) {
		return ErrHostChanged
	}
	return nil
}

// admitAtCallback is identity's side of the callback-time admission (T2, Plan r3 §1.3 step 3) for an enrollment handle
// T1 established: for the invitation kind the sealed evidence of this exchange is verified and attested first
// (attestFromEvidence; a refusal is the consumed callback's own auth.sign_in failure and nothing is admitted), then the
// composed Admitter runs the one admission transaction with the attested VerifiedAddress. A returned status (admitted,
// already_member or a known refusal the Admitter recorded itself) is passed through unchanged; only an Admitter error,
// or no composed Admitter, is identity_unavailable. Client input never supplies the address.
func (s *Service) admitAtCallback(ctx context.Context, h EnrollmentHandle, vi VerifiedIdentity, ev AddressEvidence,
	sel AdmitSelector) (AdmitOutcome, error) {
	if s.admitter == nil {
		return AdmitOutcome{}, ErrUnavailable
	}
	sel.Address = nil
	switch sel.Kind {
	case AdmitKindInvitation:
		if sel.Invitation == nil || sel.InvitationToken == "" {
			return AdmitOutcome{}, ErrInvalidRequest
		}
		va, err := s.attestFromEvidence(ctx, ev, vi, h)
		if err != nil {
			return AdmitOutcome{}, err
		}
		sel.Address = &va
	case AdmitKindOpen, AdmitKindRequest:
		sel.Invitation, sel.InvitationToken = nil, ""
	default:
		return AdmitOutcome{}, ErrInvalidRequest
	}
	out, err := s.admitter.AdmitAtCallback(ctx, h, sel)
	if err != nil {
		return AdmitOutcome{}, ErrUnavailable
	}
	return out, nil
}

// sqlAuthResolveEnrollment is the auth reader's bootstrap resolver (00016): a live, unrevoked credential of an active
// user, before any context exists.
const sqlAuthResolveEnrollment = `SELECT session_id::text, user_id::text, organization_id::text, origin, csrf_token, expires_at ` +
	`FROM app.auth_resolve_enrollment($1, $2)`

// EnrollmentCredential resolves a forwarded enrollment credential ("Authorization: Enrollment <credential>") at the
// request's exact origin (Plan r3 §1.1, §5.1): one pool statement through app.auth_resolve_enrollment (unrevoked,
// unexpired, active user). A malformed credential, an absent or overlong origin (both before any statement), no live row
// or a row minted for another origin is unauthenticated; a store failure is identity_unavailable. The result is never a
// session, constructs no tenant context and renews nothing; the own-credential operations take its Handle, and the
// caller compares the Organization the origin routes to now with Handle.OrganizationID.
func (s *Service) EnrollmentCredential(ctx context.Context, credential, origin string) (EnrollmentCredentialInfo, error) {
	raw, ok := decodeSecret(credential)
	if !ok || origin == "" || len(origin) > 300 {
		return EnrollmentCredentialInfo{}, ErrUnauthenticated
	}
	digest := hashSecret(raw)
	var info EnrollmentCredentialInfo
	var id, stored string
	err := s.store.db.QueryRow(ctx, sqlAuthResolveEnrollment, digest, s.now()).
		Scan(&id, &info.Handle.UserID, &info.Handle.OrganizationID, &stored, &info.CSRFToken, &info.ExpiresAt)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return EnrollmentCredentialInfo{}, ErrUnauthenticated
	case err != nil:
		return EnrollmentCredentialInfo{}, unavailable(err)
	case stored != origin:
		return EnrollmentCredentialInfo{}, ErrUnauthenticated
	}
	info.Handle.Origin = stored
	copy(info.Handle.Digest[:], digest)
	return info, nil
}

// EnrollmentCSRFMatches compares a submitted CSRF token with a resolved enrollment credential's (base64url, constant
// time; the CSRFMatches rule of the browser session).
func EnrollmentCSRFMatches(info EnrollmentCredentialInfo, token string) bool {
	got, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(info.CSRFToken) != secretLen {
		return false
	}
	return len(got) == len(info.CSRFToken) && subtle.ConstantTimeCompare(got, info.CSRFToken) == 1
}
