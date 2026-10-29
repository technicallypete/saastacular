package identity

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

// maxBody bounds every core auth request body (Plan r5 §7.2).
const maxBody = 16 << 10

// authService is the operation set the HTTP layer needs; *Service implements it.
type authService interface {
	Login(context.Context, LoginInput) (LoginOutput, error)
	Callback(context.Context, CallbackInput) (CallbackOutput, error)
	Session(context.Context, string, string) (SessionInfo, error) // G016 security amendment: credential, origin
	Logout(context.Context, LogoutInput) (LogoutOutput, error)
	ProviderLogoutReturn(context.Context, string) (string, error)
	SessionHost(context.Context, string) (HostResolution, error) // G016: the session route's ?origin=
	Relay(context.Context, string) (string, error)               // G016 B2: the shared callback relay decision
	Membership(context.Context, string, string) (bool, error)    // G017 D-6: the session route's membership indicator
}

// Handler serves the core internal auth contract under /internal/v1/auth/ (Plan r5 §7.2). Identity comes only from
// the "Authorization: Session <credential>" header resolved against PostgreSQL; user, role, MFA and forwarding
// headers from callers are never read. Responses are JSON, no-store, and carry no CORS headers.
//
// G062 (Plan r5 §4.5 (i)-(ii)): the root signup login start shares one global bucket, and a callback on the exact root
// origin takes a token from its login binding's bucket and from one global root-callback bucket (a churned binding
// header cannot exceed the global bucket; a pinned binding cannot exceed its own). Tenant callbacks are not limited
// here. A nil limiter admits everything (a Handler built without NewHandler, in unit tests).
type Handler struct {
	svc       authService
	root      string   // the exact platform root origin ("" disables the root limiters)
	rootStart *limiter // signup/start: 2/s burst 20, one key
	cbBinding *limiter // signup/callback/<sha256(binding)>: 1/s burst 5
	cbGlobal  *limiter // signup/callback: 5/s burst 50, one key
}

// The root limiter rates (Plan r5 §4.5).
const (
	rootStartRate, rootStartBurst          = 2, 20
	rootCallbackRate, rootCallbackBurst    = 1, 5
	rootCallbackGlobal, rootCallbackGBurst = 5, 50
)

// NewHandler returns the HTTP handler set for svc.
func NewHandler(svc *Service) Handler {
	return Handler{svc: svc, root: svc.cfg.RootOrigin, rootStart: newLimiter(rootStartRate, rootStartBurst),
		cbBinding: newLimiter(rootCallbackRate, rootCallbackBurst), cbGlobal: newLimiter(rootCallbackGlobal, rootCallbackGBurst)}
}

// rootCallbackKey is the per-binding limiter key of a root callback (a digest; the binding itself is never kept).
func rootCallbackKey(binding string) string {
	sum := sha256.Sum256([]byte("signup/callback/" + binding))
	return hex.EncodeToString(sum[:])
}

// The exact route patterns, mounted by cmd/server.
const (
	RouteLogin          = "POST /internal/v1/auth/login"
	RouteCallback       = "POST /internal/v1/auth/callback"
	RouteSession        = "GET /internal/v1/auth/session"
	RouteLogout         = "POST /internal/v1/auth/logout"
	RouteProviderReturn = "POST /internal/v1/auth/provider-logout/return"
	// G016 B2: the shared callback relay decision.
	RouteRelay = "GET /internal/v1/auth/relay"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, e *Error) {
	writeJSON(w, e.Status, map[string]string{"error": e.Code})
}

// decodeJSON enforces the mutation input rules in order: exactly one application/json Content-Type (415), then a
// body of at most 16 KiB holding exactly one JSON object with only known fields (400).
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) *Error {
	cts := r.Header.Values("Content-Type")
	if len(cts) != 1 {
		return ErrUnsupportedMediaType
	}
	mt, params, err := mime.ParseMediaType(cts[0])
	if err != nil || mt != "application/json" {
		return ErrUnsupportedMediaType
	}
	for k, v := range params {
		if k != "charset" || !strings.EqualFold(v, "utf-8") {
			return ErrUnsupportedMediaType
		}
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		return ErrInvalidRequest
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return ErrInvalidRequest
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return ErrInvalidRequest
	}
	return nil
}

// single returns the one value of a header; ok is false when it is repeated. Absent yields "", true.
func single(r *http.Request, name string) (string, bool) {
	v := r.Header.Values(name)
	switch len(v) {
	case 0:
		return "", true
	case 1:
		return v[0], true
	}
	return "", false
}

// credentialOf extracts "Authorization: Session <credential>". A repeated header is ambiguous; any other scheme or
// shape yields no credential.
func credentialOf(r *http.Request) (string, *Error) {
	v, ok := single(r, "Authorization")
	if !ok {
		return "", ErrAmbiguousParameter
	}
	cred, found := strings.CutPrefix(v, "Session ")
	if !found || strings.ContainsAny(cred, " \t") {
		return "", nil
	}
	return cred, nil
}

func bindingOf(r *http.Request) (string, *Error) {
	v, ok := single(r, "X-Login-Binding")
	if !ok {
		return "", ErrAmbiguousParameter
	}
	return v, nil
}

type loginBody struct {
	Origin   string `json:"origin"`
	ReturnTo string `json:"return_to,omitempty"`
	// G058 (Plan r3 §1.3 step 1, §5.2): an enrollment-intent login ("enroll"), optionally for one Organization invitation
	// (canonical id) and with a request on arrival.
	Intent       string `json:"intent,omitempty"`
	InvitationID string `json:"invitation_id,omitempty"`
	Request      bool   `json:"request,omitempty"`
}

// The login intents: G058 enrollment and G062 platform root signup.
const (
	intentEnroll = "enroll"
	intentSignup = "signup"
)

// parseInvitationRef parses a canonical lowercase hyphenated, non-zero invitation id; anything else is refused.
func parseInvitationRef(s string) (InvitationRef, bool) {
	var r InvitationRef
	if len(s) != 36 {
		return r, false
	}
	compact := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	b, err := hex.DecodeString(compact)
	if err != nil || len(b) != 16 {
		return r, false
	}
	copy(r[:], b)
	if r == (InvitationRef{}) || invitationString(r) != s {
		return InvitationRef{}, false
	}
	return r, true
}

// enrollmentLogin maps the G058 login fields: no intent and no enrollment facts is the ordinary login; intent "enroll"
// with an optional canonical invitation id and request flag is the enrollment login; any other intent or a malformed id
// is invalid_request (facts without the intent are refused by Login itself).
func enrollmentLogin(in loginBody) (bool, *InvitationRef, *Error) {
	if in.Intent != "" && in.Intent != intentEnroll && in.Intent != intentSignup {
		return false, nil, ErrInvalidRequest
	}
	if in.Intent == intentSignup && (in.InvitationID != "" || in.Request) {
		return false, nil, ErrInvalidRequest // G062: the signup intent carries no enrollment fact
	}
	var inv *InvitationRef
	if in.InvitationID != "" {
		r, ok := parseInvitationRef(in.InvitationID)
		if !ok {
			return false, nil, ErrInvalidRequest
		}
		inv = &r
	}
	return in.Intent == intentEnroll, inv, nil
}

type secretBody struct {
	Value  string `json:"value"`
	MaxAge int    `json:"max_age"`
}

// Login is POST /internal/v1/auth/login.
func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in loginBody
	if e := decodeJSON(w, r, &in); e != nil {
		writeError(w, e)
		return
	}
	intent, inv, e := enrollmentLogin(in)
	if e != nil {
		writeError(w, e)
		return
	}
	signup := in.Intent == intentSignup
	if signup && !h.rootStart.allow("signup/start") {
		writeError(w, ErrRateLimited)
		return
	}
	cred, e := credentialOf(r)
	if e == nil {
		var b string
		if b, e = bindingOf(r); e == nil {
			out, err := h.svc.Login(r.Context(), LoginInput{Origin: in.Origin, ReturnTo: in.ReturnTo, Credential: cred, Binding: b,
				EnrollmentIntent: intent, Invitation: inv, Request: in.Request, SignupIntent: signup})
			if err != nil {
				writeError(w, AsError(err))
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"authorization_url": out.AuthorizationURL,
				"binding":           secretBody{Value: out.Binding, MaxAge: out.BindingMaxAge},
			})
			return
		}
	}
	writeError(w, e)
}

type callbackBody struct {
	Origin   string `json:"origin"`
	RawQuery string `json:"raw_query"`
	// G058 (Plan r3 §1.6, §5.2): the BFF-forwarded invitation intent from the one-use join cookie; both or neither.
	InvitationID    string `json:"invitation_id,omitempty"`
	InvitationToken string `json:"invitation_token,omitempty"`
}

// Callback is POST /internal/v1/auth/callback. G058 (Plan r3 §1.3 step 3, §5.2): an enrollment-intent callback answered by
// the enrollment branch never carries a session: a confirmed admission is 200 {"outcome": "admitted"|"already_member",
// "next": "sign_in"} (the browser then starts the ordinary sign-in; the status is historical, never membership or
// authority); a pending or refused admission is 202 {"enrollment": {"status", "credential": {value, max_age}}} with the
// enrollment credential only in its own field. Refusals keep the closed error vocabulary (403 enrollment_closed, 409
// invitation_mismatch, 409 identity_conflict, 503 identity_unavailable on an unknown outcome).
func (h Handler) Callback(w http.ResponseWriter, r *http.Request) {
	var in callbackBody
	if e := decodeJSON(w, r, &in); e != nil {
		writeError(w, e)
		return
	}
	var inv *InvitationRef
	if in.InvitationID != "" {
		ref, ok := parseInvitationRef(in.InvitationID)
		if !ok {
			writeError(w, ErrInvalidRequest)
			return
		}
		inv = &ref
	}
	cred, e := credentialOf(r)
	if e != nil {
		writeError(w, e)
		return
	}
	b, e := bindingOf(r)
	if e != nil {
		writeError(w, e)
		return
	}
	// G062: a root-origin callback takes a token from its binding's bucket and from the global root bucket, before any
	// statement (both are always charged, so a churned binding still spends the global bucket).
	if h.root != "" && in.Origin == h.root {
		okBinding, okGlobal := h.cbBinding.allow(rootCallbackKey(b)), h.cbGlobal.allow("signup/callback")
		if !okBinding || !okGlobal {
			writeError(w, ErrRateLimited)
			return
		}
	}
	out, err := h.svc.Callback(r.Context(), CallbackInput{Origin: in.Origin, RawQuery: in.RawQuery, Binding: b, Credential: cred,
		Invitation: inv, InvitationToken: in.InvitationToken})
	if err != nil {
		writeError(w, AsError(err))
		return
	}
	if en := out.Enrollment; en != nil {
		if en.Credential == "" {
			writeJSON(w, http.StatusOK, map[string]any{"outcome": en.Status, "next": en.Next})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"enrollment": map[string]any{"status": en.Status,
			"credential": secretBody{Value: en.Credential, MaxAge: out.MaxAge}}})
		return
	}
	// G062 (Plan r5 §4.2): a root signup-intent callback answers 202 with the signup credential only, never a session.
	if su := out.Signup; su != nil {
		writeJSON(w, http.StatusAccepted, map[string]any{"signup": map[string]any{
			"credential": secretBody{Value: su.Credential, MaxAge: out.MaxAge}, "next": out.ReturnTo}})
		return
	}
	resp := map[string]any{"outcome": out.Outcome, "return_to": out.ReturnTo}
	if out.Credential != "" {
		resp["session"] = secretBody{Value: out.Credential, MaxAge: out.MaxAge}
	}
	writeJSON(w, http.StatusOK, resp)
}

// Session is GET /internal/v1/auth/session. It returns the caller's own session, including the CSRF token for the
// server-rendered account forms; it never answers for anyone else. G017 D-6: the answer carries membership ("member" |
// "none"), whether the user currently belongs to the origin's Organization.
func (h Handler) Session(w http.ResponseWriter, r *http.Request) {
	cred, e := credentialOf(r)
	if e != nil {
		writeError(w, e)
		return
	}
	if cred == "" {
		writeError(w, ErrUnauthenticated)
		return
	}
	// G016: the session is answered only for a tenant host that currently routes, together with its resolution, so an
	// authenticated page needs one core round trip; a removed or reassigned host is unknown_host. Security amendment:
	// ?origin= is required, the session must be bound to exactly that origin and to the Organization it resolves to
	// now; any other session is unauthenticated, the same answer as a dead credential.
	q, present := r.URL.Query()["origin"]
	switch {
	case !present:
		writeError(w, ErrInvalidRequest)
		return
	case len(q) != 1:
		writeError(w, ErrAmbiguousParameter)
		return
	}
	tn, err := h.svc.SessionHost(r.Context(), q[0])
	if err != nil {
		writeError(w, AsError(err))
		return
	}
	info, err := h.svc.Session(r.Context(), cred, q[0])
	if err != nil {
		writeError(w, AsError(err))
		return
	}
	if info.OrganizationID != tn.OrganizationID {
		writeError(w, ErrUnauthenticated)
		return
	}
	// G017 D-6: whether the user is a current member of the origin's Organization, re-read on every call (a revoked
	// member is "none" on the next request); asked only after the binding check above, so it never widens what the
	// credential reaches.
	member, err := h.svc.Membership(r.Context(), info.UserID, tn.OrganizationID)
	if err != nil {
		writeError(w, AsError(err))
		return
	}
	membership := "none"
	if member {
		membership = "member"
	}
	var display any
	if info.DisplayName != "" {
		display = info.DisplayName
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":       info.UserID,
		"display_name":  display,
		"mfa_assurance": info.MFAAssurance,
		"auth_time":     info.AuthTime.UTC().Format(time.RFC3339),
		"expires_at":    info.ExpiresAt.UTC().Format(time.RFC3339),
		"csrf_token":    base64.RawURLEncoding.EncodeToString(info.CSRFToken),
		"host":          hostJSON(tn),
		"membership":    membership,
	})
}

func hostJSON(t HostResolution) any {
	if !t.Routing {
		return nil
	}
	return map[string]any{"kind": t.Kind, "organization_id": t.OrganizationID}
}

// Relay is GET /internal/v1/auth/relay?state=<state> (no credential): the shared callback relay decision on the auth host
// (Plan r4 §1.7 leg 2). 200 {callback} names the initiating custom host's callback; anything else is relay_refused (the
// fixed 404 page in web). Exactly one state parameter; nothing is consumed or stored.
func (h Handler) Relay(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()["state"]
	if len(q) != 1 {
		writeError(w, ErrRelayRefused)
		return
	}
	cb, err := h.svc.Relay(r.Context(), q[0])
	if err != nil {
		writeError(w, AsError(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"callback": cb})
}

type logoutBody struct {
	Scope     string `json:"scope"`
	CSRFToken string `json:"csrf_token"`
	Origin    string `json:"origin"`
}

// Logout is POST /internal/v1/auth/logout.
func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var in logoutBody
	if e := decodeJSON(w, r, &in); e != nil {
		writeError(w, e)
		return
	}
	cred, e := credentialOf(r)
	if e != nil {
		writeError(w, e)
		return
	}
	out, err := h.svc.Logout(r.Context(), LogoutInput{Scope: in.Scope, CSRF: in.CSRFToken, Origin: in.Origin, Credential: cred})
	if err != nil {
		writeError(w, AsError(err))
		return
	}
	resp := map[string]any{"outcome": out.Outcome}
	if out.ProviderLogoutURL != "" {
		resp["provider_logout_url"] = out.ProviderLogoutURL
	}
	if out.ReturnTo != "" {
		resp["return_to"] = out.ReturnTo
	}
	if out.LogoutRef != "" {
		resp["logout_ref"] = out.LogoutRef
	}
	writeJSON(w, http.StatusOK, resp)
}

type returnBody struct {
	LogoutRef string `json:"logout_ref"`
}

// ProviderReturn is POST /internal/v1/auth/provider-logout/return.
func (h Handler) ProviderReturn(w http.ResponseWriter, r *http.Request) {
	var in returnBody
	if e := decodeJSON(w, r, &in); e != nil {
		writeError(w, e)
		return
	}
	outcome, err := h.svc.ProviderLogoutReturn(r.Context(), in.LogoutRef)
	if err != nil {
		writeError(w, AsError(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"outcome": outcome})
}
