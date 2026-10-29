// Package api is the internal HTTP surface of G062 platform root signup under /internal/v1/signup (Plan r5 §4.2, §4.5;
// internal web-BFF routes only, not a public API; no MCP change). It adapts core/internal/signup and never builds a
// tenant context, a session or an access facade.
//
// Two trust shapes: the root routes (status, slug, organizations, prove, abandon) take the caller's own root signup
// credential ("Authorization: Signup <credential>"), resolved by identity at the exact platform root origin (every other
// scheme, a Session credential included, is no credential); the tenant continuation hops (challenge, complete) take no
// credential at all (an Authorization header is ignored) and only the browser's login binding ("X-Login-Binding", as
// /internal/v1/auth/login), with the one-use continuation secrets carrying the binding. Every answer is no-store JSON
// without CORS headers; errors carry closed codes, never names, slugs, origins, credentials or continuation secrets; one
// signup.request event per request (operation, closed outcome and reason, request/trace correlation). The root routes
// that change state carry the credential's own csrf_token in the body; prove is the reviewed exception (D-15): its only
// effect is spending the presenting credential's own current nonce.
package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"core/internal/identity"
	"core/internal/signup"
)

// The signup method/path pairs (Plan r5 §4.2).
const (
	RouteStatus        = "GET /internal/v1/signup/status"
	RouteSlug          = "POST /internal/v1/signup/slug"
	RouteOrganizations = "POST /internal/v1/signup/organizations"
	RouteChallenge     = "POST /internal/v1/signup/challenge"
	RouteProve         = "POST /internal/v1/signup/prove"
	RouteComplete      = "POST /internal/v1/signup/complete"
	RouteAbandon       = "POST /internal/v1/signup/abandon"
)

// Paths are the served paths: every other method on them is method_not_allowed.
var Paths = []string{
	"/internal/v1/signup/status", "/internal/v1/signup/slug", "/internal/v1/signup/organizations",
	"/internal/v1/signup/challenge", "/internal/v1/signup/prove", "/internal/v1/signup/complete",
	"/internal/v1/signup/abandon",
}

// maxBody is the whole-payload request budget of every mutation (Plan r5 §4.2: JSON ≤ 8 KiB).
const maxBody = 8 << 10

// The buckets (Plan r5 §4.5 (iii)-(iv)): per signup credential 5 req/s burst 10 on every root route (status, slug,
// organizations, prove, abandon); per resolved user one create per 10 s, burst 5, so a fresh credential cannot replenish
// it; per tenant origin 5 req/s burst 20 on each continuation hop. Keys are digests, never a credential, user id or
// origin. The root login start and the root callback buckets live in the identity HTTP adapter.
const (
	credentialRate  = 5
	credentialBurst = 10
	userEvery       = 10 * time.Second
	userBurst       = 5
	hopRate         = 5
	hopBurst        = 20
)

// service is the signup operation set the adapter needs; *signup.Service implements it.
type service interface {
	Configured() bool
	RootOrigin() string
	Resolve(ctx context.Context, credential, origin string) (identity.SignupCredentialInfo, error)
	SuggestSlug(ctx context.Context, info identity.SignupCredentialInfo, name string) (signup.Suggestion, error)
	Provision(ctx context.Context, info identity.SignupCredentialInfo, name, slug string) (signup.Provisioned, error)
	Challenge(ctx context.Context, origin, binding, ref string) (signup.Challenge, error)
	Prove(ctx context.Context, info identity.SignupCredentialInfo, nonce string) (string, error)
	Complete(ctx context.Context, origin, binding, challenge, proof string) (identity.LoginOutput, error)
	Abandon(ctx context.Context, info identity.SignupCredentialInfo) error
}

// Config is the adapter configuration.
type Config struct {
	Service *signup.Service
	Logger  *slog.Logger
}

// Handler serves the signup routes.
type Handler struct {
	svc       service
	log       *slog.Logger
	own       *limiter // per signup credential, every root route
	user      *limiter // per resolved user, create only
	challenge *limiter // per tenant origin, the challenge hop
	complete  *limiter // per tenant origin, the completion hop
	// userOf is the per-user bucket's subject of a resolved credential (the handle's user; a package test substitutes it,
	// because only identity mints a handle).
	userOf func(identity.SignupCredentialInfo) string
}

// New returns the adapter. A missing service is refused; an unconfigured service is mounted and answers
// signup_unconfigured on every route.
func New(cfg Config) (*Handler, error) {
	if cfg.Service == nil {
		return nil, errors.New("signup api: incomplete configuration")
	}
	return newHandler(cfg.Service, cfg.Logger), nil
}

func newHandler(svc service, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Handler{svc: svc, log: log, own: newLimiter(credentialRate, credentialBurst), user: newSlowLimiter(userEvery, userBurst),
		challenge: newLimiter(hopRate, hopBurst), complete: newLimiter(hopRate, hopBurst),
		userOf: func(info identity.SignupCredentialInfo) string { return info.Handle.UserID() }}
}

// newSlowLimiter is the pinned limiter with a refill slower than one token per second (one token per every; the copy's
// constructor takes whole tokens per second). Only the construction differs; allow and reclaim are the pinned ones.
func newSlowLimiter(every time.Duration, burst int) *limiter {
	return &limiter{m: map[string]*rate.Limiter{}, r: rate.Every(every), burst: burst, now: time.Now, full: every * time.Duration(burst)}
}

// Register mounts every route on mux, a closed method_not_allowed answer for the other methods of each served path and
// a closed not_found answer at and below the /internal/v1/signup root.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(RouteStatus, h.status)
	mux.HandleFunc(RouteSlug, h.root(signup.OpSlug, true, h.slug))
	mux.HandleFunc(RouteOrganizations, h.root(signup.OpCreate, true, h.create))
	mux.HandleFunc(RouteProve, h.root(signup.OpProve, false, h.prove))
	mux.HandleFunc(RouteAbandon, h.root(signup.OpAbandon, true, h.abandon))
	mux.HandleFunc(RouteChallenge, h.hop(signup.OpChallenge, h.challengeHop))
	mux.HandleFunc(RouteComplete, h.hop(signup.OpComplete, h.completeHop))
	for _, p := range Paths {
		mux.HandleFunc(p, h.closed(signup.OpMethod, errMethod))
	}
	mux.HandleFunc("/internal/v1/signup", h.closed(signup.OpPath, errNotFound))
	mux.HandleFunc("/internal/v1/signup/", h.closed(signup.OpPath, errNotFound))
}

// The closed refusal codes and their statuses. conflict carries the closed field slug_taken or intent_used (logged as
// that field); invalid_request may carry the field name or slug.
const (
	errInvalidRequest  = "invalid_request"
	errOriginRejected  = "origin_rejected"
	errUnauthenticated = "unauthenticated"
	errCSRF            = "csrf_failed"
	errUserDisabled    = "user_disabled"
	errConflict        = "conflict"
	errContinuation    = "continuation_invalid"
	errContextMismatch = "signup_context_mismatch"
	errRateLimited     = "rate_limited"
	errNotFound        = "not_found"
	errMethod          = "method_not_allowed"
	errUnconfigured    = "signup_unconfigured"
	errUnavailable     = "signup_unavailable"
)

var status = map[string]int{
	errInvalidRequest: 400, errOriginRejected: 403, errUnauthenticated: 401, errCSRF: 403, errUserDisabled: 403,
	errConflict: 409, errContinuation: 404, errContextMismatch: 403, errRateLimited: 429, errNotFound: 404, errMethod: 405,
	errUnconfigured: 503, errUnavailable: 503,
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// recorder observes one answer for the signup.request event: reason is the closed log reason ("" for success).
type recorder struct {
	http.ResponseWriter
	reason string
}

// refuse writes a closed refusal with an optional closed field. A conflict is logged as its field.
func (h *Handler) refuse(w *recorder, code, field string) {
	w.reason = code
	if code == errConflict {
		w.reason = field
	}
	out := map[string]any{"error": code}
	if field != "" {
		out["field"] = field
	}
	writeJSON(w, status[code], out)
}

func (h *Handler) fail(w *recorder, code string) { h.refuse(w, code, "") }

// refuseErr maps a signup error onto its closed refusal.
func (h *Handler) refuseErr(w *recorder, err error) {
	e := signup.AsError(err)
	switch e.Code {
	case signup.CodeInvalidRequest:
		h.refuse(w, errInvalidRequest, e.Field)
	case signup.CodeSlugTaken, signup.CodeIntentUsed:
		h.refuse(w, errConflict, string(e.Code))
	case signup.CodeUnauthenticated:
		h.fail(w, errUnauthenticated)
	case signup.CodeCSRFFailed:
		h.fail(w, errCSRF)
	case signup.CodeUserDisabled:
		h.fail(w, errUserDisabled)
	case signup.CodeContinuationInvalid:
		h.fail(w, errContinuation)
	case signup.CodeContextMismatch:
		h.fail(w, errContextMismatch)
	case signup.CodeUnconfigured:
		h.fail(w, errUnconfigured)
	default:
		h.fail(w, errUnavailable)
	}
}

// event writes the request's one signup.request record.
func (h *Handler) event(ctx context.Context, op string, rec *recorder) {
	signup.LogRequest(ctx, h.log, op, rec.reason)
}

// closed answers a method outside a served path's closed set, or a path at or below the root that is not served.
func (h *Handler) closed(op, code string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w}
		defer h.event(r.Context(), op, rec)
		h.fail(rec, code)
	}
}

// body is every JSON mutation's body; each route requires its own members and refuses the others.
type body struct {
	Origin    *string `json:"origin"`
	CSRFToken *string `json:"csrf_token"`
	Name      *string `json:"name"`
	Slug      *string `json:"slug"`
	Ref       *string `json:"ref"`
	Nonce     *string `json:"nonce"`
	Challenge *string `json:"challenge"`
	Proof     *string `json:"proof"`
}

// shape reports whether exactly the named members are present.
func (in body) shape(names ...string) bool {
	present := map[string]bool{"origin": in.Origin != nil, "csrf_token": in.CSRFToken != nil, "name": in.Name != nil,
		"slug": in.Slug != nil, "ref": in.Ref != nil, "nonce": in.Nonce != nil, "challenge": in.Challenge != nil,
		"proof": in.Proof != nil}
	n := 0
	for _, name := range names {
		if !present[name] {
			return false
		}
		n++
	}
	for _, set := range present {
		if set {
			n--
		}
	}
	return n == 0
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// decode reads at most maxBody bytes (more is invalid_request before any decode, credential or statement) of a
// query-free application/json mutation and decodes exactly one JSON object with known members only.
func (h *Handler) decode(w *recorder, r *http.Request) (body, bool) {
	if len(r.URL.Query()) != 0 {
		h.fail(w, errInvalidRequest)
		return body{}, false
	}
	cts := r.Header.Values("Content-Type")
	if len(cts) != 1 {
		h.fail(w, errInvalidRequest)
		return body{}, false
	}
	mt, params, err := mime.ParseMediaType(cts[0])
	if err != nil || mt != "application/json" {
		h.fail(w, errInvalidRequest)
		return body{}, false
	}
	for k, v := range params {
		if k != "charset" || !strings.EqualFold(v, "utf-8") {
			h.fail(w, errInvalidRequest)
			return body{}, false
		}
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil || len(raw) > maxBody {
		h.fail(w, errInvalidRequest)
		return body{}, false
	}
	var in body
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	// Exactly one JSON object: the second Decode must report io.EOF (trailing whitespace is fine).
	if err := dec.Decode(&in); err != nil || !errors.Is(dec.Decode(new(json.RawMessage)), io.EOF) {
		h.fail(w, errInvalidRequest)
		return body{}, false
	}
	return in, true
}

// keyOf is a limiter key: the SHA-256 of a scoped value (the value itself is never kept).
func keyOf(scope, v string) string {
	sum := sha256.Sum256([]byte("signup/1\x00" + scope + "\x00" + v))
	return hex.EncodeToString(sum[:])
}

// credentialOf extracts "Authorization: Signup <credential>": a repeated header is invalid_request; any other scheme or
// shape (a Session or Enrollment credential included) is no credential.
func credentialOf(r *http.Request) (string, bool) {
	v := r.Header.Values("Authorization")
	switch len(v) {
	case 0:
		return "", true
	case 1:
	default:
		return "", false
	}
	c, ok := strings.CutPrefix(v[0], "Signup ")
	if !ok || c == "" || strings.ContainsAny(c, " \t") {
		return "", true
	}
	return c, true
}

// bindingOf returns the browser's login binding header ("" when absent); a repeated header is refused.
func bindingOf(r *http.Request) (string, bool) {
	v := r.Header.Values("X-Login-Binding")
	switch len(v) {
	case 0:
		return "", true
	case 1:
		return v[0], true
	}
	return "", false
}

// authenticate authenticates a root request: the exact root origin (else origin_rejected), the Signup credential (a
// missing one is unauthenticated), the per-credential bucket, identity's resolution at that origin and, for a form
// mutation, the credential's own CSRF token.
func (h *Handler) authenticate(w *recorder, r *http.Request, origin, csrf string, mutation bool) (identity.SignupCredentialInfo, bool) {
	if !h.svc.Configured() {
		h.fail(w, errUnconfigured)
		return identity.SignupCredentialInfo{}, false
	}
	cred, ok := credentialOf(r)
	if !ok {
		h.fail(w, errInvalidRequest)
		return identity.SignupCredentialInfo{}, false
	}
	if root := h.svc.RootOrigin(); root == "" || origin != root {
		h.fail(w, errOriginRejected)
		return identity.SignupCredentialInfo{}, false
	}
	if cred == "" {
		h.fail(w, errUnauthenticated)
		return identity.SignupCredentialInfo{}, false
	}
	if !h.own.allow(keyOf("credential", cred)) {
		h.fail(w, errRateLimited)
		return identity.SignupCredentialInfo{}, false
	}
	info, err := h.svc.Resolve(r.Context(), cred, origin)
	if err != nil {
		h.refuseErr(w, err)
		return identity.SignupCredentialInfo{}, false
	}
	if mutation && !identity.SignupCSRFMatches(info, csrf) {
		h.fail(w, errCSRF)
		return identity.SignupCredentialInfo{}, false
	}
	return info, true
}

// status is GET /internal/v1/signup/status?origin= (own signup credential): the reconciliation read of the root forms.
// 200 {"status": pending|created|completed, "csrf_token", "organization"?: {"slug", "target_origin"}} (organization once
// provisioned); a dead credential is 401 unauthenticated (the BFF clears the signup cookie).
func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), signup.OpStatus, rec)
	q := r.URL.Query()
	origin, ok := q["origin"]
	if !ok || len(origin) != 1 || len(q) != 1 {
		h.fail(rec, errInvalidRequest)
		return
	}
	info, ok := h.authenticate(rec, r, origin[0], "", false)
	if !ok {
		return
	}
	out := map[string]any{"status": signup.StatusOf(info), "csrf_token": base64.RawURLEncoding.EncodeToString(info.CSRFToken)}
	if info.Provisioned {
		out["organization"] = map[string]any{"slug": info.Slug, "target_origin": info.TargetOrigin}
	}
	writeJSON(rec, http.StatusOK, out)
}

// root is a root mutation: decode, the route's exact body members, the own credential (with its CSRF token when csrf),
// then act; one signup.request.
func (h *Handler) root(op string, csrf bool, act func(w *recorder, r *http.Request, info identity.SignupCredentialInfo, in body)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w}
		defer h.event(r.Context(), op, rec)
		in, ok := h.decode(rec, r)
		if !ok {
			return
		}
		var names []string
		switch op {
		case signup.OpSlug:
			names = []string{"origin", "csrf_token", "name"}
		case signup.OpCreate:
			names = []string{"origin", "csrf_token", "name", "slug"}
		case signup.OpProve:
			names = []string{"origin", "nonce"}
		default:
			names = []string{"origin", "csrf_token"}
		}
		if !in.shape(names...) {
			h.fail(rec, errInvalidRequest)
			return
		}
		info, ok := h.authenticate(rec, r, str(in.Origin), str(in.CSRFToken), csrf)
		if !ok {
			return
		}
		act(rec, r, info, in)
	}
}

// slug is POST /internal/v1/signup/slug {origin, csrf_token, name}: the advisory suggestion. 200 {"suggested",
// "available"} (suggested null when the name has no ASCII letter or digit); nothing is reserved.
func (h *Handler) slug(w *recorder, r *http.Request, info identity.SignupCredentialInfo, in body) {
	s, err := h.svc.SuggestSlug(r.Context(), info, *in.Name)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	var suggested any
	if s.Suggested != "" {
		suggested = s.Suggested
	}
	writeJSON(w, http.StatusOK, map[string]any{"suggested": suggested, "available": s.Available})
}

// create is POST /internal/v1/signup/organizations {origin, csrf_token, name, slug}: constrained provisioning, after the
// per-user bucket. 201 {"outcome": created|already_created, "organization": {"id", "slug"}, "continue"} (an identical
// intent re-mints generation g+1); 409 conflict field slug_taken|intent_used; 400 invalid_request field name|slug; 403
// user_disabled; 401 unauthenticated; 503 signup_unavailable (unknown commit: an identical retry answers
// already_created).
func (h *Handler) create(w *recorder, r *http.Request, info identity.SignupCredentialInfo, in body) {
	if !h.user.allow(keyOf("user", h.userOf(info))) {
		h.fail(w, errRateLimited)
		return
	}
	p, err := h.svc.Provision(r.Context(), info, *in.Name, *in.Slug)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"outcome": p.Outcome,
		"organization": map[string]any{"id": p.OrganizationID, "slug": p.Slug}, "continue": p.Continue})
}

// prove is POST /internal/v1/signup/prove {origin, nonce} (own signup credential; no CSRF token, D-15): continuation
// step 3. 200 {"complete"}; 403 signup_context_mismatch (nothing consumed); 401 unauthenticated.
func (h *Handler) prove(w *recorder, r *http.Request, info identity.SignupCredentialInfo, in body) {
	u, err := h.svc.Prove(r.Context(), info, *in.Nonce)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"complete": u})
}

// abandon is POST /internal/v1/signup/abandon {origin, csrf_token}: revokes the caller's own credential and the current
// generation. 200 {}.
func (h *Handler) abandon(w *recorder, r *http.Request, info identity.SignupCredentialInfo, _ body) {
	if err := h.svc.Abandon(r.Context(), info); err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

// exactOrigin reports whether origin is an exact https origin (scheme and host[:port] only).
func exactOrigin(origin string) bool {
	u, err := url.Parse(origin)
	return err == nil && len(origin) <= 300 && u.Scheme == "https" && u.Host != "" && u.User == nil && u.Path == "" &&
		u.RawPath == "" && u.RawQuery == "" && !u.ForceQuery && u.Fragment == "" && u.Opaque == "" && u.String() == origin
}

// hop is a tenant continuation hop: decode, the route's exact body members, an exact https origin (anything else is
// continuation_invalid), the browser's login binding, the per-tenant-origin bucket, then act. Any Authorization header
// is ignored: a hop has no credential and no authority beyond its one-use secrets.
func (h *Handler) hop(op string, act func(w *recorder, r *http.Request, origin, binding string, in body)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names, bucket := []string{"origin", "ref"}, h.challenge
		if op == signup.OpComplete {
			names, bucket = []string{"origin", "challenge", "proof"}, h.complete
		}
		rec := &recorder{ResponseWriter: w}
		defer h.event(r.Context(), op, rec)
		in, ok := h.decode(rec, r)
		if !ok {
			return
		}
		if !in.shape(names...) {
			h.fail(rec, errInvalidRequest)
			return
		}
		if !h.svc.Configured() {
			h.fail(rec, errUnconfigured)
			return
		}
		binding, ok := bindingOf(r)
		if !ok {
			h.fail(rec, errInvalidRequest)
			return
		}
		origin := *in.Origin
		if !exactOrigin(origin) {
			h.fail(rec, errContinuation)
			return
		}
		if !bucket.allow(keyOf(op, origin)) {
			h.fail(rec, errRateLimited)
			return
		}
		act(rec, r, origin, binding, in)
	}
}

// challengeHop is POST /internal/v1/signup/challenge {origin, ref} (tenant origin; X-Login-Binding): continuation step 2.
// 200 {"challenge": {"value", "max_age": 120}, "prove"}; 404 continuation_invalid.
func (h *Handler) challengeHop(w *recorder, r *http.Request, origin, binding string, in body) {
	c, err := h.svc.Challenge(r.Context(), origin, binding, *in.Ref)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"challenge": map[string]any{"value": c.Value, "max_age": c.MaxAge}, "prove": c.Prove})
}

// completeHop is POST /internal/v1/signup/complete {origin, challenge, proof} (tenant origin; X-Login-Binding):
// continuation step 4. 200 the ordinary login answer {"authorization_url", "binding": {"value", "max_age"}} (the pinned
// login; return_to "/"); 404 continuation_invalid.
func (h *Handler) completeHop(w *recorder, r *http.Request, origin, binding string, in body) {
	out, err := h.svc.Complete(r.Context(), origin, binding, *in.Challenge, *in.Proof)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"authorization_url": out.AuthorizationURL,
		"binding": map[string]any{"value": out.Binding, "max_age": out.BindingMaxAge}})
}
