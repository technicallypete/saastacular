package identity

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/url"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/audit"
	"core/internal/observability"
)

// Bounds on server-side auth network operations (Spec R3; Plan r5 §3.6).
const (
	transactionTTL   = 10 * time.Minute
	bindingMaxAge    = 600 // seconds; the binding cookie's Max-Age, refreshed on each start
	callbackDeadline = 10 * time.Second
	exchangeDeadline = 8 * time.Second
	revokeDeadline   = 5 * time.Second
	clockLeeway      = 30 * time.Second
	logoutRefTTL     = 10 * time.Minute
)

// Config wires the service. Provider is nil when identity is unconfigured: provider operations then answer
// identity_unconfigured while session resolution and local logout keep working (Plan r5 §7.3).
type Config struct {
	Provider   Provider
	Issuer     string
	ClientID   string
	Origins    Origins      // the exact non-tenant origins (IDENTITY_PUBLIC_ORIGINS: the shared auth host)
	Hosts      HostResolver // G016: tenant-host origins resolve through Organization slugs and verified domains
	AuthOrigin string       // G016 B2: the shared callback relay origin (https://<PLATFORM_AUTH_HOST>[:port]); "" disables it
	// G062 (Plan r5 §1.2, §4.1): the exact platform root origin (https://<PLATFORM_BASE_DOMAIN>[:PLATFORM_HOST_PORT]), the
	// only origin a signup-intent login starts on and whose own /auth/callback completes it; "" disables root signup.
	RootOrigin string
	Idle       time.Duration
	Max        time.Duration
	MaxAuthAge time.Duration // zero: freshness is not required
	Now        func() time.Time
	Logger     *slog.Logger
}

// Service implements the core identity operations over the PostgreSQL store.
type Service struct {
	store *Store
	cfg   Config
	// G058: the callback-time admission service (Plan r3 §5.1); nil until composed, and then every admission is
	// identity_unavailable (fail closed).
	admitter Admitter
}

// NewService returns a service. A nil Now uses the wall clock; a nil Logger discards records.
func NewService(store *Store, cfg Config) *Service {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	if store != nil {
		store.logger = cfg.Logger // G011 audit diagnostics (audit.record_failed, audit.gap)
	}
	return &Service{store: store, cfg: cfg}
}

// Configured reports whether a provider is configured.
func (s *Service) Configured() bool { return s.cfg.Provider != nil }

func (s *Service) now() time.Time { return s.cfg.Now().UTC().Truncate(time.Microsecond) }

// event writes one closed-vocabulary identity record: only outcome/scope plus, inside a request, its request and
// trace correlation (events.json).
func (s *Service) event(ctx context.Context, level slog.Level, name string, attrs ...slog.Attr) {
	if rid := observability.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid), slog.String("trace_id", observability.TraceID(ctx)))
	}
	s.cfg.Logger.LogAttrs(context.Background(), level, name, attrs...)
}

func outcomeOf(err error) string { return AsError(err).Code }

// CheckNamespace compares the configured issuer and client with the bound namespace row. It is used at startup
// (logged once) and before every provider operation.
func (s *Service) CheckNamespace(ctx context.Context) error {
	issuer, client, found, err := s.store.binding(ctx)
	if err != nil {
		return err
	}
	if found && (issuer != s.cfg.Issuer || client != s.cfg.ClientID) {
		return ErrNamespaceMismatch
	}
	return nil
}

// LogNamespaceAtStartup records one identity.namespace_mismatch event (no values) when the configuration no longer
// matches the bound namespace.
func (s *Service) LogNamespaceAtStartup(ctx context.Context) {
	if !s.Configured() {
		return
	}
	if err := s.CheckNamespace(ctx); err == ErrNamespaceMismatch {
		s.cfg.Logger.LogAttrs(context.Background(), slog.LevelWarn, "identity.namespace_mismatch")
	}
}

// LoginInput is the core login-start request.
type LoginInput struct {
	Origin     string
	ReturnTo   string
	Credential string // optional forwarded session credential
	Binding    string // optional existing login binding
	// G058 (Plan r3 §1.3 step 1): an enrollment-intent login, optionally for one Organization invitation and with a
	// request on arrival. Invitation and Request require EnrollmentIntent; the login's return path is /join/status.
	EnrollmentIntent bool
	Invitation       *InvitationRef
	Request          bool
	// G062 (Plan r5 §1.2): a platform signup-intent login, accepted only on the exact root origin; it pins no
	// Organization, its return path is /signup/details and its callback is the root's own /auth/callback (no relay).
	SignupIntent bool
}

// LoginOutput carries the hosted authorization URL and the (new or reused) login binding.
type LoginOutput struct {
	AuthorizationURL string
	Binding          string
	BindingMaxAge    int
}

// Login starts one login transaction bound to the browser's login binding (Plan r5 §4.2).
func (s *Service) Login(ctx context.Context, in LoginInput) (out LoginOutput, err error) {
	defer func() {
		outcome := "started"
		if err != nil {
			outcome = outcomeOf(err)
		}
		s.event(ctx, slog.LevelInfo, "auth.login.started", slog.String("outcome", outcome))
	}()
	// An unconfigured deployment has no origin allow-list to judge against: every start is identity_unconfigured.
	if !s.Configured() {
		return LoginOutput{}, ErrUnconfigured
	}
	// G016: a login starts only on a tenant host that currently routes; the resolution runs inside the transaction that
	// stores the login (below). Structurally impossible origins are refused here, before anything else.
	if s.cfg.Hosts == nil || in.Origin == "" || len(in.Origin) > 300 {
		return LoginOutput{}, ErrOriginRejected
	}
	// G058: the enrollment facts exist only on an enrollment-intent login (refused before anything else is read).
	if !in.EnrollmentIntent && (in.Invitation != nil || in.Request) {
		return LoginOutput{}, ErrInvalidRequest
	}
	// G062: the signup intent excludes the enrollment facts and exists only on the exact root origin (never www, a tenant,
	// the auth host or a custom host); the root has no session, so a forwarded credential is not consulted.
	if in.SignupIntent {
		if in.EnrollmentIntent {
			return LoginOutput{}, ErrInvalidRequest
		}
		if s.cfg.RootOrigin == "" || in.Origin != s.cfg.RootOrigin {
			return LoginOutput{}, ErrOriginRejected
		}
		return s.signupLogin(ctx, in)
	}
	if in.Credential != "" {
		// G016 security amendment (Plan amendment A2): already_authenticated only for a session bound to this origin and
		// to the Organization the origin resolves to now; on a reassigned host the old Organization's cookie (same
		// origin) falls through to a new sign-in. The resolution is a pool statement before the transaction, exactly as
		// SessionHost; the in-transaction resolution below stays the pinned snapshot for the INSERT.
		tn, err := s.tenantOrigin(ctx, s.store.db, in.Origin)
		if err != nil {
			return LoginOutput{}, err
		}
		live, err := s.Session(ctx, in.Credential, in.Origin)
		switch {
		case err == nil && live.OrganizationID == tn.OrganizationID:
			return LoginOutput{}, ErrAlreadyAuthenticated
		case err != nil && err != ErrUnauthenticated:
			return LoginOutput{}, err
		}
	}
	if err := s.CheckNamespace(ctx); err != nil {
		return LoginOutput{}, err
	}
	binding, ok := decodeSecret(in.Binding)
	if !ok {
		if binding, err = newSecret(); err != nil {
			return LoginOutput{}, ErrUnavailable
		}
	}
	stateRaw, err := newSecret()
	if err != nil {
		return LoginOutput{}, ErrUnavailable
	}
	verifierRaw, err := newSecret()
	if err != nil {
		return LoginOutput{}, ErrUnavailable
	}
	state, verifier := encodeSecret(stateRaw), encodeSecret(verifierRaw)
	redirectOrigin := in.Origin
	now := s.now()
	maxAge := 0
	if s.cfg.MaxAuthAge > 0 {
		maxAge = int(s.cfg.MaxAuthAge / time.Second)
	}
	// G016 (Plan r4 §1.7 leg 1, §5.5): the tenant-host resolution is one statement inside the login transaction, before
	// the INSERT: the Organization is pinned on the row, and a custom host's state carries its origin tag with the
	// shared relay origin as redirect_uri. A host that routes nowhere is origin_rejected with nothing inserted.
	if err := s.store.insertLogin(ctx, func(tx pgx.Tx) (loginTxn, error) {
		tn, err := s.tenantOrigin(ctx, tx, in.Origin)
		if err != nil {
			return loginTxn{}, err
		}
		callbackOrigin := ""
		if tn.Kind == HostCustom {
			if s.cfg.AuthOrigin == "" {
				return loginTxn{}, ErrOriginRejected // B2 unconfigured: a custom host has no registered callback
			}
			callbackOrigin = s.cfg.AuthOrigin
			state, redirectOrigin = taggedState(state, in.Origin), callbackOrigin
		}
		t := loginTxn{
			StateHash: hashEncoded(state), BindingHash: hashSecret(binding), Verifier: verifier, ClientID: s.cfg.ClientID,
			Origin: in.Origin, ReturnPath: ReturnPath(in.ReturnTo), MaxAge: maxAge, CreatedAt: now, ExpiresAt: now.Add(transactionTTL),
			OrganizationID: tn.OrganizationID, CallbackOrigin: callbackOrigin,
		}
		if in.EnrollmentIntent {
			t.ReturnPath, t.EnrollmentIntent, t.RequestOnArrival = enrollmentReturnPath, true, in.Request
			if in.Invitation != nil {
				t.InvitationID = invitationString(*in.Invitation)
			}
		}
		return t, nil
	}); err != nil {
		return LoginOutput{}, err
	}
	u, err := s.cfg.Provider.AuthorizationURL(ctx, AuthRequest{
		State: state, CodeChallenge: s256Challenge(verifier), RedirectURI: CallbackURL(redirectOrigin), MaxAge: maxAge,
	})
	if err != nil {
		return LoginOutput{}, AsError(err)
	}
	return LoginOutput{AuthorizationURL: u, Binding: encodeSecret(binding), BindingMaxAge: bindingMaxAge}, nil
}

// CallbackInput is the core callback request: the raw query of the browser's callback navigation, the login binding
// and any forwarded session credential.
type CallbackInput struct {
	Origin     string
	RawQuery   string
	Binding    string
	Credential string
	// G058 (Plan r3 §1.3 step 2 (b), §1.6): the BFF-forwarded invitation intent of an enrollment-intent login (from the
	// one-use join cookie). Both or neither; a malformed pair is refused before the transaction is consumed.
	Invitation      *InvitationRef
	InvitationToken string
}

// CallbackOutput is a successful callback. Credential is set for signed_in and session_rotated. G058: Enrollment is set
// (with Outcome enrollment_status) for an enrollment-intent callback answered by the enrollment branch; it never carries a
// session.
type CallbackOutput struct {
	Outcome    string
	Credential string
	MaxAge     int
	ReturnTo   string
	Enrollment *EnrollmentOutcome
	// G062: Signup is set (with Outcome signup and ReturnTo /signup/details) for a root signup-intent callback (T1); it
	// carries the signup credential only, never a session.
	Signup *SignupOutcome
}

// Callback validates and completes one login (Plan r5 §4.2, §3.3). Duplicate protocol parameters, an unconfigured
// provider, a wrong origin, a namespace mismatch and a wrong or missing binding consume nothing; every later failure
// has consumed the one-use transaction.
//
// G011: a failure after the one-use transaction was consumed records one auth.sign_in denied or failed event through
// the bounded identity recorder, before the operational record: user_disabled attributed to the verified mapped
// subject, identity_conflict to the live session's subject, every other reason unattributed. Pre-consume refusals stay
// operational only (anonymous, no state change). A success is recorded atomically by signIn.
func (s *Service) Callback(ctx context.Context, in CallbackInput) (out CallbackOutput, err error) {
	ctx, cancel := context.WithTimeout(ctx, callbackDeadline)
	defer cancel()
	defer func() {
		outcome := out.Outcome
		if err != nil {
			outcome = outcomeOf(err)
		}
		s.event(ctx, slog.LevelInfo, "auth.callback.completed", slog.String("outcome", outcome))
	}()
	// G058: recorded is set when the enrollment branch already recorded the failure's one durable event (access.denied
	// enrollment.start), so the auth.sign_in failure record is not added (one truthful phase-owned event).
	consumed, subject, recorded := false, "", false
	defer func() {
		if consumed && err != nil && !recorded {
			s.store.recordIndependent(ctx, subject, signInFailure(err, subject))
		}
	}()

	params, perr := url.ParseQuery(in.RawQuery)
	if perr != nil {
		return CallbackOutput{}, ErrAmbiguousParameter
	}
	for _, k := range []string{"code", "state", "error", "error_description"} {
		if len(params[k]) > 1 {
			return CallbackOutput{}, ErrAmbiguousParameter
		}
	}
	// G058: the forwarded invitation intent is both-or-neither with a well-formed token, refused before consumption.
	if (in.Invitation == nil) != (in.InvitationToken == "") || !invitationTokenForm(in.InvitationToken) {
		return CallbackOutput{}, ErrInvalidRequest
	}
	if !s.Configured() {
		return CallbackOutput{}, ErrUnconfigured
	}
	if s.cfg.Hosts == nil || in.Origin == "" {
		return CallbackOutput{}, ErrOriginRejected
	}
	if err := s.CheckNamespace(ctx); err != nil {
		return CallbackOutput{}, err
	}
	state := params.Get("state")
	binding, ok := decodeSecret(in.Binding)
	if state == "" || len(state) > 512 || !ok {
		return CallbackOutput{}, ErrTransactionInvalid
	}
	now := s.now()
	// G016: the pre-exchange host check is one statement inside the consume transaction (Plan r4 §1.7 leg 3, §5.5).
	txn, res, err := s.store.consumeLogin(ctx, hashEncoded(state), hashSecret(binding), s.cfg.ClientID, in.Origin, now,
		func(tx pgx.Tx, t loginTxn) (bool, error) {
			if t.SignupIntent { // G062: a signup-intent login completes only on the configured root origin
				return s.cfg.RootOrigin != "" && t.Origin == s.cfg.RootOrigin, nil
			}
			tn, err := s.tenantOrigin(ctx, tx, in.Origin)
			if err == ErrUnavailable {
				return false, err
			}
			return err == nil && t.OrganizationID != "" && tn.OrganizationID == t.OrganizationID, nil
		})
	if err != nil {
		return CallbackOutput{}, err
	}
	switch res {
	case consumeNone:
		return CallbackOutput{}, ErrTransactionInvalid
	case consumeExpired:
		consumed = true
		return CallbackOutput{}, ErrTransactionExpired
	}
	consumed = true

	// G016 pre-exchange checks (Plan r4 §1.7 leg 3), after the one-use consumption and before any exchange: a relayed
	// login must name the configured relay origin, and the host must still route to the login's Organization (a
	// removed or reassigned domain is host_changed: consumed, no exchange, no session).
	if txn.CallbackOrigin != "" && txn.CallbackOrigin != s.cfg.AuthOrigin {
		return CallbackOutput{}, ErrHostChanged
	}
	if txn.hostChanged {
		if txn.SignupIntent {
			return CallbackOutput{}, ErrSignupOriginRejected
		}
		return CallbackOutput{}, ErrHostChanged
	}

	if _, present := params["error"]; present {
		if params.Get("error") == "access_denied" {
			return CallbackOutput{}, ErrLoginCanceled
		}
		return CallbackOutput{}, ErrProviderError
	}
	code := params.Get("code")
	if code == "" || len(code) > 1024 {
		return CallbackOutput{}, ErrCodeRejected
	}
	exCtx, exCancel := context.WithTimeout(ctx, exchangeDeadline)
	var vi VerifiedIdentity
	var ev AddressEvidence
	if txn.EnrollmentIntent {
		// G058: the enrollment branch's exchange carries a fresh sealed minter (the invitation kind's recipient proof).
		vi, ev, err = s.exchangeWithEvidence(exCtx, code, txn.Verifier)
	} else {
		vi, err = s.cfg.Provider.Exchange(exCtx, Exchange{Code: code, CodeVerifier: txn.Verifier})
	}
	exCancel()
	if err != nil {
		return CallbackOutput{}, AsError(err)
	}
	if vi.Issuer != s.cfg.Issuer || vi.ClientID != s.cfg.ClientID {
		return CallbackOutput{}, ErrProviderError // the adapter already enforces both; never accept a drift
	}
	if txn.MaxAge > 0 && now.Sub(vi.AuthTime) > time.Duration(txn.MaxAge)*time.Second+clockLeeway {
		return CallbackOutput{}, ErrStaleAuthentication
	}
	// G062 T1: the root signup branch (no session, no tenant context; a forwarded credential is not consulted).
	if txn.SignupIntent {
		return s.signupCallback(ctx, vi, in.Origin, now, &subject)
	}

	var live *SessionInfo
	if in.Credential != "" {
		// G016 security amendment: continuity only for a session bound to this origin and to the Organization pinned on
		// the login; any other session is ignored (a new session is issued and the stale row stays untouched, refused
		// everywhere else by its own binding).
		info, err := s.Session(ctx, in.Credential, in.Origin)
		switch {
		case err == nil && info.OrganizationID == txn.OrganizationID:
			live = &info
		case err == nil:
			// bound to another Organization (a reassigned host): not this sign-in's session
		case err != ErrUnauthenticated:
			return CallbackOutput{}, err
		}
	}
	if txn.EnrollmentIntent {
		eo, signIn, eerr := s.enrollCallback(ctx, in, txn, vi, ev, live, now, &subject, &recorded)
		if eerr != nil || !signIn {
			return eo, eerr
		}
		// a subject already signed in on this host, or already a member: the unchanged ordinary sign-in below
	}
	p := signInParams{Identity: vi, Live: live, Now: now, Idle: s.cfg.Idle, Max: s.cfg.Max,
		Origin: in.Origin, Organization: txn.OrganizationID, Hosts: s.cfg.Hosts}
	if txn.SignupHandoff {
		// G062: the pinned tenant handoff (Plan r5 §1.4 step 4): the expected user is fixed on the row; the shape CHECK
		// guarantees it is set, and an empty value is refused rather than unpinned.
		if txn.ExpectedUserID == "" {
			return CallbackOutput{}, ErrSignupUserMismatch
		}
		p.ExpectedUser = txn.ExpectedUserID
	}
	result, err := s.store.signIn(ctx, p)
	if err != nil {
		if err == ErrUserDisabled {
			subject = result.Subject
		}
		return CallbackOutput{}, err
	}
	if result.Conflict {
		subject = live.UserID
		// A live session for another user: it stays untouched and no mapping exists for this principal. The extra
		// provider session is revoked best effort; only the outcome is recorded.
		s.revokeExtraProviderSession(ctx, vi.ProviderSessionID)
		return CallbackOutput{}, ErrIdentityConflict
	}
	out = CallbackOutput{Outcome: result.Outcome, ReturnTo: txn.ReturnPath}
	if result.Credential != nil {
		out.Credential = encodeSecret(result.Credential)
		out.MaxAge = int(result.ExpiresAt.Sub(now) / time.Second)
	}
	return out, nil
}

// signInFailure is the auth.sign_in event of a post-consume callback failure: the closed denied reasons, the closed
// failed reasons, and identity_unavailable (failed) for any other code. subject is set only for user_disabled,
// identity_conflict and (G058) recipient_proof_unavailable, the consumed enrollment callback's recipient-proof failure,
// attributed only to a subject the callback actually established.
func signInFailure(err error, subject string) audit.Event {
	e := audit.Event{Action: audit.AuthSignIn, Outcome: audit.Failed, Reason: audit.ReasonIdentityUnavailable}
	switch r := audit.Reason(AsError(err).Code); r {
	case audit.ReasonTransactionExpired, audit.ReasonLoginCanceled, audit.ReasonCodeRejected, audit.ReasonImpersonationRefused,
		audit.ReasonMFAUnverified, audit.ReasonStaleAuthentication, audit.ReasonIdentityConflict, audit.ReasonUserDisabled,
		audit.ReasonHostChanged, audit.ReasonOrganizationNotMember, audit.ReasonRecipientProofUnavailable,
		audit.ReasonSignupOriginRejected, audit.ReasonSignupUserMismatch:
		e.Outcome, e.Reason = audit.Denied, r
	case audit.ReasonProviderError, audit.ReasonProviderTimeout, audit.ReasonIdentityNamespaceMismatch:
		e.Reason = r
	}
	if subject != "" && (e.Reason == audit.ReasonUserDisabled || e.Reason == audit.ReasonIdentityConflict ||
		e.Reason == audit.ReasonRecipientProofUnavailable) {
		e.Subject, _ = audit.ParseUUID(subject)
	}
	return e
}

func (s *Service) revokeExtraProviderSession(ctx context.Context, sid string) {
	outcome := OutcomeProviderFailed
	if sid != "" {
		rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), revokeDeadline)
		if err := s.cfg.Provider.RevokeProviderSession(rctx, sid); err == nil {
			outcome = OutcomeProviderRevoked
		}
		cancel()
	}
	s.event(ctx, slog.LevelInfo, "auth.provider.revocation", slog.String("outcome", outcome), slog.String("scope", "identity_conflict"))
}

// Session resolves a forwarded credential to its live session at the request's exact origin. It needs only PostgreSQL,
// so it behaves identically whether the provider is configured or not. G016 security amendment: a session resolves
// only at the origin it was issued for (an absent or overlong origin is unauthenticated before any statement); the
// result carries the session's Organization binding for the caller's comparison with the origin's current Organization.
func (s *Service) Session(ctx context.Context, credential, origin string) (SessionInfo, error) {
	raw, ok := decodeSecret(credential)
	if !ok || origin == "" || len(origin) > 300 {
		return SessionInfo{}, ErrUnauthenticated
	}
	return s.store.resolve(ctx, hashSecret(raw), origin, s.now(), s.cfg.Idle)
}

// Membership answers whether userID currently holds the Organization membership binding of orgID (G017 D-6): one
// statement through the auth-reader function app.is_organization_member, with no lock, context or renewal. The session
// route asks it only after the session resolved at, and is bound to, the origin's current Organization, so it can never
// make a credential usable on another host; it is an indicator fact and grants nothing.
func (s *Service) Membership(ctx context.Context, userID, orgID string) (bool, error) {
	var member bool
	if err := s.store.db.QueryRow(ctx, sqlIsOrganizationMember, userID, orgID).Scan(&member); err != nil {
		return false, unavailable(err)
	}
	return member, nil
}

// LockLiveSession checks, inside the caller's READ COMMITTED read-write transaction, that the session is live and its
// user active, and holds FOR SHARE on the user row and then the session row until that transaction ends (G007 guarded
// writes). A known session and user ID come from an earlier Session resolution, never from a client. It returns nil,
// ErrUnauthenticated or ErrUnavailable; it renews nothing and logs nothing.
func (s *Service) LockLiveSession(ctx context.Context, tx pgx.Tx, sessionID, userID string) error {
	return s.store.liveSession(ctx, tx, sessionID, userID, s.now(), true)
}

// CheckLiveSession is LockLiveSession without row locks, for a REPEATABLE READ READ ONLY snapshot (G007 guarded reads
// and context resolution).
func (s *Service) CheckLiveSession(ctx context.Context, tx pgx.Tx, sessionID, userID string) error {
	return s.store.liveSession(ctx, tx, sessionID, userID, s.now(), false)
}

// DisableUser disables a user and revokes its sessions (trusted service method; no route).
func (s *Service) DisableUser(ctx context.Context, userID string) error {
	err := s.store.DisableUser(ctx, userID, s.now())
	if err == nil {
		s.event(ctx, slog.LevelInfo, "auth.session.revoked", slog.String("outcome", OutcomeLocalRevoked), slog.String("scope", "disabled"))
	}
	return err
}

// EnableUser re-enables a user (trusted service method; no route). Old sessions stay revoked.
func (s *Service) EnableUser(ctx context.Context, userID string) error {
	return s.store.EnableUser(ctx, userID, s.now())
}

// G062 platform root signup (Plan r5 §1.2-§1.4, §3.4, §4.2-§4.3). Identity owns the database context of every signup
// statement (the pinned sqlInitContext lives only here and in tenancy): the T1 signup row, the credential resolution, the
// IDENTITY shape of a resolved signup credential, the tenant continuation's LOGIN shape with its runtime host resolution,
// and the pinned handoff login. core/internal/signup holds the provisioning, challenge and prove statements.

// signupLogin stores one signup-intent login transaction on the root origin (Plan r5 §1.2): no Organization, return
// path /signup/details, the root's own /auth/callback as redirect URI (no relay). Login already checked the origin.
func (s *Service) signupLogin(ctx context.Context, in LoginInput) (LoginOutput, error) {
	if err := s.CheckNamespace(ctx); err != nil {
		return LoginOutput{}, err
	}
	binding, ok := decodeSecret(in.Binding)
	if !ok {
		var err error
		if binding, err = newSecret(); err != nil {
			return LoginOutput{}, ErrUnavailable
		}
	}
	stateRaw, err := newSecret()
	if err != nil {
		return LoginOutput{}, ErrUnavailable
	}
	verifierRaw, err := newSecret()
	if err != nil {
		return LoginOutput{}, ErrUnavailable
	}
	state, verifier := encodeSecret(stateRaw), encodeSecret(verifierRaw)
	now := s.now()
	maxAge := 0
	if s.cfg.MaxAuthAge > 0 {
		maxAge = int(s.cfg.MaxAuthAge / time.Second)
	}
	if err := s.store.insertTransaction(ctx, loginTxn{
		StateHash: hashEncoded(state), BindingHash: hashSecret(binding), Verifier: verifier, ClientID: s.cfg.ClientID,
		Origin: in.Origin, ReturnPath: signupReturnPath, MaxAge: maxAge, CreatedAt: now, ExpiresAt: now.Add(transactionTTL),
		SignupIntent: true,
	}); err != nil {
		return LoginOutput{}, err
	}
	u, err := s.cfg.Provider.AuthorizationURL(ctx, AuthRequest{
		State: state, CodeChallenge: s256Challenge(verifier), RedirectURI: CallbackURL(in.Origin), MaxAge: maxAge,
	})
	if err != nil {
		return LoginOutput{}, AsError(err)
	}
	return LoginOutput{AuthorizationURL: u, Binding: encodeSecret(binding), BindingMaxAge: bindingMaxAge}, nil
}

// signupCallback is the T1 branch of a consumed signup-intent callback (Plan r5 §1.2): the exact (issuer, subject)
// mapping is found or created (ADR-007 G062 amendment; never merged by e-mail), the user must be active, and one signup
// row is inserted in the IDENTITY shape of that user with auth.sign_in success signup_authenticated, atomically. No
// session, no tenant context, no Organization.
func (s *Service) signupCallback(ctx context.Context, vi VerifiedIdentity, origin string, now time.Time, subject *string) (CallbackOutput, error) {
	r, err := s.store.signupT1(ctx, vi, origin, now)
	if err != nil {
		if err == ErrUserDisabled {
			*subject = r.Subject
		}
		return CallbackOutput{}, err
	}
	return CallbackOutput{Outcome: OutcomeSignup, ReturnTo: signupReturnPath, MaxAge: int(r.ExpiresAt.Sub(now) / time.Second),
		Signup: &SignupOutcome{Credential: encodeSecret(r.Credential)}}, nil
}

// sqlAuthResolveSignup is the auth reader's bootstrap resolver (00018): a live, unrevoked signup credential of an active
// user, before any context exists. No digest is returned.
const sqlAuthResolveSignup = `SELECT session_id::text, user_id::text, origin, csrf_token, expires_at, coalesce(organization_id::text, ''), ` +
	`coalesce(slug, ''), coalesce(target_origin, ''), provisioned_at IS NOT NULL, continuation_generation, completed_at IS NOT NULL ` +
	`FROM app.auth_resolve_signup($1, $2)`

// SignupCredential resolves a forwarded root signup credential ("Authorization: Signup <credential>") at the request's
// exact origin, which must be the configured root origin (Plan r5 §4.2). A malformed credential, another or absent
// origin (both before any statement), no live row or a row minted for another origin is unauthenticated; a store failure
// is identity_unavailable. The result is never a session and constructs no tenant context.
func (s *Service) SignupCredential(ctx context.Context, credential, origin string) (SignupCredentialInfo, error) {
	raw, ok := decodeSecret(credential)
	if !ok || origin == "" || len(origin) > 300 || s.cfg.RootOrigin == "" || origin != s.cfg.RootOrigin {
		return SignupCredentialInfo{}, ErrUnauthenticated
	}
	digest := hashSecret(raw)
	var info SignupCredentialInfo
	var stored string
	err := s.store.db.QueryRow(ctx, sqlAuthResolveSignup, digest, s.now()).Scan(&info.Handle.sessionID, &info.Handle.userID,
		&stored, &info.CSRFToken, &info.ExpiresAt, &info.OrganizationID, &info.Slug, &info.TargetOrigin, &info.Provisioned,
		&info.Generation, &info.Completed)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return SignupCredentialInfo{}, ErrUnauthenticated
	case err != nil:
		return SignupCredentialInfo{}, unavailable(err)
	case stored != origin:
		return SignupCredentialInfo{}, ErrUnauthenticated
	}
	info.Handle.origin = stored
	copy(info.Handle.digest[:], digest)
	return info, nil
}

// SignupCSRFMatches compares a submitted CSRF token with a resolved signup credential's (base64url, constant time).
func SignupCSRFMatches(info SignupCredentialInfo, token string) bool {
	got, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(info.CSRFToken) != secretLen {
		return false
	}
	return len(got) == len(info.CSRFToken) && subtle.ConstantTimeCompare(got, info.CSRFToken) == 1
}

// SignupIdentityTx runs fn in one READ COMMITTED transaction in the IDENTITY shape of the resolved signup credential's
// user (Plan r5 §3.4: provisioning and the step-3 prove). The user comes only from SignupCredential (the handle's fields
// are unexported); a zero handle is unauthenticated before any statement.
func (s *Service) SignupIdentityTx(ctx context.Context, h SignupHandle, fn func(tx pgx.Tx) error) error {
	if h.userID == "" || h.sessionID == "" {
		return ErrUnauthenticated
	}
	return s.store.signupIdentityTx(ctx, h.userID, fn)
}

// SignupContinuationTx runs fn in one READ COMMITTED transaction in the LOGIN shape of the browser's login binding (a
// fresh random binding when none is presented: the shape only, never authority) after resolving the exact tenant origin
// on that transaction (Plan r5 §1.4 step 2): fn receives the runtime-resolved request Organization. A host that routes
// nowhere is host_changed; a custom host (never a signup target) is continuation_invalid.
func (s *Service) SignupContinuationTx(ctx context.Context, origin, binding string, fn func(tx pgx.Tx, organizationID string) error) error {
	raw, ok := decodeSecret(binding)
	if !ok {
		var err error
		if raw, err = newSecret(); err != nil {
			return ErrUnavailable
		}
	}
	return s.store.loginTx(ctx, hashSecret(raw), func(tx pgx.Tx) error {
		org, err := s.continuationOrganization(ctx, tx, origin)
		if err != nil {
			return err
		}
		return fn(tx, org)
	})
}

// continuationOrganization resolves a continuation hop's exact origin on q: a routing platform host's Organization;
// a host that routes nowhere is host_changed, a custom host continuation_invalid, a store failure unavailable.
func (s *Service) continuationOrganization(ctx context.Context, q Querier, origin string) (string, error) {
	tn, err := s.tenantOrigin(ctx, q, origin)
	switch {
	case err == ErrOriginRejected:
		return "", ErrHostChanged
	case err != nil:
		return "", err
	case tn.Kind != HostPlatform:
		return "", ErrContinuationInvalid
	}
	return tn.OrganizationID, nil
}

// PinnedLoginInput is one tenant completion hop (Plan r5 §1.4 step 4): the exact tenant origin, the browser's login
// binding (or none), and the SHA-256 digests of the tenant challenge C (from the host-only __Host-sct cookie) and the
// proof P (from the query). Raw secrets never reach identity.
type PinnedLoginInput struct {
	Origin    string
	Binding   string
	Challenge []byte
	Proof     []byte
}

// sqlSignupChallengeComplete is the provision writer's completion body (00018): matches only when C and P belong to the
// same row and generation, for this origin and the runtime-resolved Organization; zero rows is continuation_invalid.
const sqlSignupChallengeComplete = `SELECT user_id::text, organization_id::text FROM app.signup_challenge_complete($1, $2, $3, $4::uuid, $5)`

// LoginPinned is the automatic continuation's completion (Plan r5 §1.4 step 4, §3.4): one short LOGIN transaction
// resolves the tenant host and calls app.signup_challenge_complete (a wrong C or P changes nothing); a second, the
// ordinary login-start transaction, re-resolves the host, requires current membership (runtime
// app.is_organization_member) and inserts the pinned login row (signup_handoff, expected_user_id, return_path "/") under
// the unchanged login_transactions_insert policy. The pin restricts which user may complete; it grants nothing. A failure
// after the completion leaves the generation completed and no login row (recovery (b)).
func (s *Service) LoginPinned(ctx context.Context, in PinnedLoginInput) (out LoginOutput, err error) {
	defer func() {
		outcome := "started"
		if err != nil {
			outcome = outcomeOf(err)
		}
		s.event(ctx, slog.LevelInfo, "auth.login.started", slog.String("outcome", outcome))
	}()
	if !s.Configured() {
		return LoginOutput{}, ErrUnconfigured
	}
	if s.cfg.Hosts == nil || in.Origin == "" || len(in.Origin) > 300 || len(in.Challenge) != secretLen || len(in.Proof) != secretLen {
		return LoginOutput{}, ErrContinuationInvalid
	}
	if err := s.CheckNamespace(ctx); err != nil {
		return LoginOutput{}, err
	}
	binding, ok := decodeSecret(in.Binding)
	if !ok {
		if binding, err = newSecret(); err != nil {
			return LoginOutput{}, ErrUnavailable
		}
	}
	now := s.now()
	var userID, orgID string
	if err := s.store.loginTx(ctx, hashSecret(binding), func(tx pgx.Tx) error {
		org, err := s.continuationOrganization(ctx, tx, in.Origin)
		if err != nil {
			return err
		}
		err = tx.QueryRow(ctx, sqlSignupChallengeComplete, in.Challenge, in.Proof, in.Origin, org, now).Scan(&userID, &orgID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrContinuationInvalid
		}
		return unavailable(err)
	}); err != nil {
		return LoginOutput{}, err
	}
	stateRaw, err := newSecret()
	if err != nil {
		return LoginOutput{}, ErrUnavailable
	}
	verifierRaw, err := newSecret()
	if err != nil {
		return LoginOutput{}, ErrUnavailable
	}
	state, verifier := encodeSecret(stateRaw), encodeSecret(verifierRaw)
	maxAge := 0
	if s.cfg.MaxAuthAge > 0 {
		maxAge = int(s.cfg.MaxAuthAge / time.Second)
	}
	if err := s.store.insertLogin(ctx, func(tx pgx.Tx) (loginTxn, error) {
		org, err := s.continuationOrganization(ctx, tx, in.Origin)
		if err != nil {
			return loginTxn{}, err
		}
		if org != orgID {
			return loginTxn{}, ErrHostChanged
		}
		var member bool
		if err := tx.QueryRow(ctx, sqlIsOrganizationMember, userID, orgID).Scan(&member); err != nil {
			return loginTxn{}, unavailable(err)
		}
		if !member {
			return loginTxn{}, ErrNotMember
		}
		return loginTxn{
			StateHash: hashEncoded(state), BindingHash: hashSecret(binding), Verifier: verifier, ClientID: s.cfg.ClientID,
			Origin: in.Origin, ReturnPath: "/", MaxAge: maxAge, CreatedAt: now, ExpiresAt: now.Add(transactionTTL),
			OrganizationID: orgID, SignupHandoff: true, ExpectedUserID: userID,
		}, nil
	}); err != nil {
		return LoginOutput{}, err
	}
	u, err := s.cfg.Provider.AuthorizationURL(ctx, AuthRequest{
		State: state, CodeChallenge: s256Challenge(verifier), RedirectURI: CallbackURL(in.Origin), MaxAge: maxAge,
	})
	if err != nil {
		return LoginOutput{}, AsError(err)
	}
	return LoginOutput{AuthorizationURL: u, Binding: encodeSecret(binding), BindingMaxAge: bindingMaxAge}, nil
}

// signupDenialReasons are the access.denied reasons a signup operation may carry (00018 events_attribute).
var signupDenialReasons = []audit.Reason{audit.ReasonSignupInvalid, audit.ReasonIntentUsed, audit.ReasonSlugTaken,
	audit.ReasonContinuationInvalid, audit.ReasonSignupContextMismatch, audit.ReasonHostChanged}

// signupOperations are the closed signup operations (00018).
var signupOperations = []audit.Operation{audit.OpSignupStart, audit.OpSignupStatus, audit.OpSignupSlug, audit.OpSignupCreate,
	audit.OpSignupChallenge, audit.OpSignupProve, audit.OpSignupComplete, audit.OpSignupAbandon}

// RecordSignupDenial records one access.denied event of a signup operation through the bounded independent recorder:
// no Organization, Group or resource; attributed to the resolved signup credential's user, or unattributed for a zero
// handle (a hop without a credential). Any other operation or reason is refused without a statement (the closed 00018
// shape). It accepts no user identifier: attribution comes only from SignupCredential.
func (s *Service) RecordSignupDenial(ctx context.Context, h SignupHandle, op audit.Operation, reason audit.Reason) {
	if !slices.Contains(signupOperations, op) || !slices.Contains(signupDenialReasons, reason) {
		return
	}
	s.store.recordIndependent(ctx, h.userID, audit.Event{Action: audit.AccessDenied, Outcome: audit.Denied, Reason: reason, Operation: op})
}
