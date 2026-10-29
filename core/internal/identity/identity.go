// Package identity owns global users, their exact provider principals and PostgreSQL-authoritative opaque
// application sessions (G004, ADR-003). The WorkOS adapter (package identity/workos) implements Provider; nothing
// outside that package imports a provider SDK, JWT or JWKS library.
//
// Authority stays here: a session resolves to a canonical user UUID only through core's own PostgreSQL state.
// Session MFA assurance is always "unknown" (ADR-003 clarification, M2); no provider setting, method, claim,
// browser input or test result can change it, and RequireVerifiedMFA denies every session.
package identity

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

// Provider is the owned boundary to the identity provider. Implementations validate everything listed in Plan r5
// §3.3 before returning a VerifiedIdentity; callers never see provider tokens, claims or SDK types.
type Provider interface {
	// AuthorizationURL returns the hosted sign-in URL for one login transaction.
	AuthorizationURL(ctx context.Context, r AuthRequest) (string, error)
	// Exchange redeems an authorization code with the transaction's own PKCE verifier and returns the validated
	// identity.
	Exchange(ctx context.Context, in Exchange) (VerifiedIdentity, error)
	// RevokeProviderSession revokes exactly one provider session (sid). Success proves only that.
	RevokeProviderSession(ctx context.Context, sid string) error
	// ProviderLogoutURL returns the hosted logout URL that ends the browser's provider session for sid.
	ProviderLogoutURL(sid, returnTo string) (string, error)
}

// AuthRequest carries one transaction's authorization parameters.
type AuthRequest struct {
	State         string
	CodeChallenge string // BASE64URL(SHA-256(verifier)); method S256
	RedirectURI   string
	MaxAge        int // seconds; zero means the parameter is not sent
}

// Exchange carries the code from the callback and the verifier of the transaction that issued its challenge.
//
// G058 (Plan r3 §5.1; readiness correction 01 §1–§2): Minter is the per-call sealed evidence minter identity hands to
// the provider. A zero Minter (every caller outside identity's enrollment branch) mints nothing.
type Exchange struct {
	Code         string
	CodeVerifier string
	Minter       EvidenceMinter
}

// AddressEvidence is the provider adapter's observation, during one exchange, of the address of the exchanged principal
// and whether the provider reports it verified (G058 D-15(i)). Its fields are unexported and it has no JSON form: only
// EvidenceMinter.Mint produces a non-zero value, which binds the exchanged issuer and subject. Identity verifies it
// (attestFromEvidence) before the G019 AttestVerifiedAddress; it never crosses HTTP and is never stored.
type AddressEvidence struct {
	address    string
	verified   bool
	observedAt time.Time
	issuer     string
	subject    string
}

// IsZero reports whether e is the zero value (no evidence).
func (e AddressEvidence) IsZero() bool { return e == AddressEvidence{} }

// EvidenceMinter is the sealed, per-call minter of AddressEvidence. Only identity constructs a usable one
// (newEvidenceMinter) and passes it in one Exchange input; the provider adapter calls Mint at most once for that
// exchange and identity reads the minted value back after Exchange returns. A zero EvidenceMinter mints a zero value; a
// second Mint on one minter poisons it (zero evidence, fail closed).
type EvidenceMinter struct {
	sink *evidenceSink
}

type evidenceSink struct {
	mu    sync.Mutex
	ev    AddressEvidence
	mints int
}

// newEvidenceMinter returns a fresh minter for exactly one exchange.
func newEvidenceMinter() EvidenceMinter { return EvidenceMinter{sink: &evidenceSink{}} }

// Mint records the provider's projection of this exchange (address, the provider's verified flag and the adapter's
// observation time) bound to the exchanged identity's issuer and subject, and returns it. Pinned call sites: only the
// provider adapter (TestAddressEvidenceUnforgeable).
func (m EvidenceMinter) Mint(vi VerifiedIdentity, address string, verified bool, at time.Time) AddressEvidence {
	if m.sink == nil || vi.Issuer == "" || vi.Subject == "" {
		return AddressEvidence{}
	}
	m.sink.mu.Lock()
	defer m.sink.mu.Unlock()
	m.sink.mints++
	if m.sink.mints != 1 {
		m.sink.ev = AddressEvidence{}
		return AddressEvidence{}
	}
	m.sink.ev = AddressEvidence{address: address, verified: verified, observedAt: at.UTC(), issuer: vi.Issuer, subject: vi.Subject}
	return m.sink.ev
}

// minted returns the one value minted for this exchange, or the zero value (none, or more than one mint).
func (m EvidenceMinter) minted() AddressEvidence {
	if m.sink == nil {
		return AddressEvidence{}
	}
	m.sink.mu.Lock()
	defer m.sink.mu.Unlock()
	if m.sink.mints != 1 {
		return AddressEvidence{}
	}
	return m.sink.ev
}

// EnrollmentHandle is identity's view of one resolved enrollment credential (G058 Plan r3 §5.1): the canonical user,
// the Organization and exact origin it was minted for, and the credential digest. It carries no provider facts.
type EnrollmentHandle struct {
	UserID         string
	OrganizationID string
	Origin         string
	Digest         [32]byte
}

// EnrollmentCredentialInfo is one live enrollment credential resolved at its exact origin (Service.EnrollmentCredential):
// the handle the own-credential operations take, the credential's CSRF token (for the BFF's forms and the mutation
// check, EnrollmentCSRFMatches) and its absolute expiry. It is not a session and carries no provider fact.
type EnrollmentCredentialInfo struct {
	Handle    EnrollmentHandle
	CSRFToken []byte
	ExpiresAt time.Time
}

// InvitationRef is identity's own value of an Organization invitation id (G058 Plan r3 §5.1); access converts it to and
// from its tenancy id. Identity imports neither access nor tenancy.
type InvitationRef [16]byte

// AdmitKind is the admission kind of a callback-time admission (G058 Plan r3 §1.3 step 3).
type AdmitKind string

// The admission kinds. The request kind (an enrollment login with request_on_arrival on a request_access Organization)
// submits the user's own access request (app.enrollment_request) and never admits: its success is status pending.
const (
	AdmitKindInvitation AdmitKind = "invitation"
	AdmitKindOpen       AdmitKind = "open"
	AdmitKindRequest    AdmitKind = "request"
)

// AdmitSelector selects one admission of an enrollment handle's user (identity-owned DTO, readiness correction 01 D-3):
// for the invitation kind the invitation and its token, and Address, the VerifiedAddress identity attested from this
// exchange's sealed evidence (never client input).
type AdmitSelector struct {
	Kind            AdmitKind
	Invitation      *InvitationRef
	InvitationToken string
	Address         *VerifiedAddress
}

// AdmitOutcome is the closed result of one admission (Plan r3 §1.4): Status admitted or already_member (committed;
// BindingID names the Organization membership binding when known), pending (the request kind: the user's own access
// request committed or already pending; no membership), or a known refusal recorded by the admitting service
// (enrollment_closed, enrollment_restricted, recipient_ineligible, already_consumed, not_found, enrollment_expired,
// host_changed) with nothing written. An infrastructure failure is an error, never a status.
type AdmitOutcome struct {
	Status    string
	BindingID string
}

// The admission statuses (Plan r3 §1.4 closed outcome set, the subset this slice produces).
const (
	AdmitAdmitted             = "admitted"
	AdmitAlreadyMember        = "already_member"
	AdmitEnrollmentClosed     = "enrollment_closed"
	AdmitEnrollmentRestricted = "enrollment_restricted"
	AdmitRecipientIneligible  = "recipient_ineligible"
	AdmitAlreadyConsumed      = "already_consumed"
	AdmitNotFound             = "not_found"
	AdmitEnrollmentExpired    = "enrollment_expired"
	AdmitHostChanged          = "host_changed"
)

// Admitter performs the callback-time admission (T2) of an enrollment handle's user (G058 Plan r3 §5.1: cycle-free
// orchestration with identity-owned DTOs). The tenancy access service implements it; it records its own known denials
// (access.denied, enrollment.admit), so identity records nothing further for a returned status. The server composes it
// through core/internal/enrollment/api and binds it once before serving (Service.BindAdmitter).
type Admitter interface {
	AdmitAtCallback(ctx context.Context, h EnrollmentHandle, sel AdmitSelector) (AdmitOutcome, error)
}

// EnrollmentOutcome is the enrollment answer of an enrollment-intent callback (Plan r3 §1.3 step 3, §5.1): Status
// admitted or already_member with Next "sign_in" (a confirmed T2; no credential: the browser then starts the ordinary
// login), or the closed status of a pending or refused admission with the enrollment Credential (set once) and never a
// session.
type EnrollmentOutcome struct {
	Status     string
	Credential string
	Next       string
}

// The enrollment callback statuses beyond AdmitOutcome's: pending (an enrollment credential without an admission:
// request_access, with or without the request submitted on arrival) and the next step of a confirmed admission.
const (
	EnrollmentPending = "pending"
	NextSignIn        = "sign_in"
)

// VerifiedIdentity is a validated provider principal. It deliberately has no email, authentication method or
// assurance field.
type VerifiedIdentity struct {
	Issuer            string
	ClientID          string
	Subject           string
	ProviderSessionID string // the non-bearer provider session reference (sid)
	DisplayName       string // untrusted text; never a key
	AuthTime          time.Time
}

// MFAAssuranceUnknown is the only MFA assurance a G004 session can have.
const MFAAssuranceUnknown = "unknown"

// SessionInfo is the resolved state of one live application session.
type SessionInfo struct {
	SessionID         string
	UserID            string
	DisplayName       string
	MFAAssurance      string
	AuthTime          time.Time
	ExpiresAt         time.Time // the earlier of the idle and absolute expiry
	CSRFToken         []byte
	ProviderSessionID string
	// Origin and OrganizationID are the session's binding (G016 security amendment): the exact tenant-host origin that
	// issued it and that origin's canonical Organization at issuance. A consumer that resolves the request origin to an
	// Organization must also require OrganizationID to equal it (a reassigned host keeps its origin).
	Origin         string
	OrganizationID string
}

// Error is a closed, client-safe error: a fixed code and its HTTP status. It never carries provider bodies, claims
// or values.
type Error struct {
	Code   string
	Status int
}

func (e *Error) Error() string { return "identity: " + e.Code }

func newError(code string, status int) *Error { return &Error{Code: code, Status: status} }

// The closed error vocabulary of the core auth contract (Plan r5 §7.2; IDENTITY.md).
var (
	ErrInvalidRequest       = newError("invalid_request", http.StatusBadRequest)
	ErrAmbiguousParameter   = newError("ambiguous_parameter", http.StatusBadRequest)
	ErrUnsupportedMediaType = newError("unsupported_media_type", http.StatusUnsupportedMediaType)
	ErrOriginRejected       = newError("origin_rejected", http.StatusForbidden)
	ErrUnauthenticated      = newError("unauthenticated", http.StatusUnauthorized)
	ErrCSRFFailed           = newError("csrf_failed", http.StatusForbidden)
	ErrAlreadyAuthenticated = newError("already_authenticated", http.StatusConflict)
	ErrTransactionInvalid   = newError("transaction_invalid", http.StatusUnauthorized)
	ErrTransactionExpired   = newError("transaction_expired", http.StatusUnauthorized)
	ErrLoginCanceled        = newError("login_canceled", http.StatusUnauthorized)
	ErrCodeRejected         = newError("code_rejected", http.StatusUnauthorized)
	ErrIdentityConflict     = newError("identity_conflict", http.StatusConflict)
	ErrUserDisabled         = newError("user_disabled", http.StatusForbidden)
	ErrImpersonationRefused = newError("impersonation_refused", http.StatusForbidden)
	ErrStaleAuthentication  = newError("stale_authentication", http.StatusForbidden)
	ErrMFAUnverified        = newError("mfa_unverified", http.StatusForbidden)
	ErrProviderError        = newError("provider_error", http.StatusBadGateway)
	ErrProviderTimeout      = newError("provider_timeout", http.StatusGatewayTimeout)
	ErrNamespaceMismatch    = newError("identity_namespace_mismatch", http.StatusServiceUnavailable)
	ErrUnconfigured         = newError("identity_unconfigured", http.StatusServiceUnavailable)
	ErrUnavailable          = newError("identity_unavailable", http.StatusServiceUnavailable)
	// G016: the tenant-host issuance guard (Plan r4 §1.7.1) and the shared callback relay (§1.7).
	ErrHostChanged  = newError("host_changed", http.StatusForbidden)            // the host no longer routes to the login's Organization
	ErrNotMember    = newError("organization_not_member", http.StatusForbidden) // D-2(a): no current membership of that Organization
	ErrRelayRefused = newError("relay_refused", http.StatusNotFound)            // the relay answers its fixed 404
	ErrUnknownHost  = newError("unknown_host", http.StatusNotFound)             // the session route's origin routes nowhere
	// G058: the exchange produced no usable verified-address evidence for the invitation (Plan r3 §1.4 closed set).
	ErrRecipientProofUnavailable = newError("recipient_proof_unavailable", http.StatusConflict)
	// G058 callback T1 (Plan r3 §1.3 step 2 (b)/(d), §5.2): the Organization admits no enrollment for this login (403), and
	// the forwarded invitation does not match the one the login transaction stored (409). Both are consumed refusals with
	// nothing created, recorded as access.denied enrollment.start (never auth.sign_in).
	ErrEnrollmentClosed   = newError("enrollment_closed", http.StatusForbidden)
	ErrInvitationMismatch = newError("invitation_mismatch", http.StatusConflict)
	// G062 (Plan r5 §4.2, §4.5): a consumed signup-intent callback whose origin is no longer the platform root origin; a
	// pinned tenant callback whose verified principal is not the expected user (consumed; nothing issued); a tenant
	// continuation hop that matched no row (nothing consumed beyond a legitimate use); and the root limiters.
	ErrSignupOriginRejected = newError("signup_origin_rejected", http.StatusForbidden)
	ErrSignupUserMismatch   = newError("signup_user_mismatch", http.StatusForbidden)
	ErrContinuationInvalid  = newError("continuation_invalid", http.StatusNotFound)
	ErrRateLimited          = newError("rate_limited", http.StatusTooManyRequests)
)

// SignupHandle is identity's view of one resolved root signup credential (G062 Plan r5 §3.2, §4.3): the signup row, its
// user and exact root origin, and the credential digest. Its fields are unexported, so only SignupCredential mints one;
// SignupIdentityTx derives the IDENTITY shape from it and nothing else. It is never a session.
type SignupHandle struct {
	sessionID string
	userID    string
	origin    string
	digest    [32]byte
}

// UserID is the signup credential's canonical user.
func (h SignupHandle) UserID() string { return h.userID }

// Origin is the exact root origin the credential was minted for.
func (h SignupHandle) Origin() string { return h.origin }

// Digest is the SHA-256 of the credential (the provisioning body's own-row key); the credential itself is never kept.
func (h SignupHandle) Digest() []byte { return append([]byte(nil), h.digest[:]...) }

// SignupCredentialInfo is one live root signup credential resolved at its exact root origin (Service.SignupCredential,
// app.auth_resolve_signup): the handle, the CSRF token of the root forms, the absolute expiry and the status/recovery
// projection (no digest of any continuation secret is ever returned). OrganizationID, Slug and TargetOrigin are set once
// the row is provisioned.
type SignupCredentialInfo struct {
	Handle         SignupHandle
	CSRFToken      []byte
	ExpiresAt      time.Time
	OrganizationID string
	Slug           string
	TargetOrigin   string
	Provisioned    bool
	Generation     int
	Completed      bool
}

// SignupOutcome is the answer of a root signup-intent callback (T1): the signup credential, set once, and never a
// session.
type SignupOutcome struct {
	Credential string
}

// SignupLifetime is the absolute lifetime of a root signup credential (the signup_sessions_lifetime CHECK).
const SignupLifetime = 30 * time.Minute

// signupReturnPath is the one return path of a signup-intent login (00018 login_transactions_return_path_check).
const signupReturnPath = "/signup/details"

// AsError maps any error onto the closed vocabulary; unknown errors become identity_unavailable.
func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return ErrUnavailable
}

// Session kinds (G014; migration 00010): a browser session carries the G004 cookie credential; an MCP session is owned
// by exactly one MCP connection, has no issued bearer and is never resolved by E1.
const (
	SessionKindBrowser = "browser"
	SessionKindMCP     = "mcp"
)

// Logout scopes and outcomes (closed vocabularies; Plan r5 §4.4).
const (
	ScopeCurrent  = "current"
	ScopeAll      = "all"
	ScopeProvider = "provider"

	OutcomeSignedIn          = "signed_in"
	OutcomeSessionRotated    = "session_rotated"
	OutcomeAlreadySignedIn   = "already_signed_in"
	OutcomeEnrollmentStatus  = "enrollment_status" // G058: an enrollment-intent callback answered with EnrollmentOutcome
	OutcomeSignup            = "signup"            // G062: a root signup-intent callback answered with SignupOutcome (T1)
	OutcomeTenantCreated     = "tenant_created"    // G062: the pinned tenant callback's success (return_to stays "/")
	OutcomeLocalRevoked      = "local_revoked"
	OutcomeProviderRevoked   = "provider_session_revoked"
	OutcomeProviderFailed    = "provider_revoke_failed"
	OutcomeProviderRedirect  = "provider_logout_redirected"
	OutcomeProviderReturned  = "provider_logout_returned"
	revokeLogout             = "logout"
	revokeLogoutAll          = "logout_all"
	revokeProviderLogout     = "provider_logout"
	revokeDisabled           = "disabled"
	revokeSuperseded         = "superseded"
	providerRevocationOK     = "confirmed"
	providerRevocationFailed = "failed"
)
