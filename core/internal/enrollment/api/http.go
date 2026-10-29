// Package api is the G058 Organization enrollment composition (Plan r3 §5.1/§5.2; Codex round-3 decision 2; HTTP
// execution release of 2026-10-18). It composes the callback-time admission (one facade delegates identity's Admitter
// to this package's own access facade; Bind hands it to identity once, at startup, before the server serves) and the
// internal enrollment HTTP surface under /internal/v1/enrollment (internal web-BFF routes only, not a public API; no MCP
// change). Like the domains, branding, onboarding and membership handlers it builds its own access facade: the raw
// tenancy service is unreachable and tenancy is not imported.
//
// Three trust shapes: the public projection (no credential; the Organization is the trusted G016 resolver's decision
// for the request origin); the applicant's own enrollment credential ("Authorization: Enrollment <credential>",
// resolved at its exact origin by identity and bound to the Organization the origin routes to now; never a session, no
// tenant context); and the browser session of an Organization manager (the session-scope behaviour is a bounded copy of
// core/internal/onboarding/api, pinned byte for byte by http_test). Every answer is no-store JSON without CORS headers;
// errors carry closed codes, never names, addresses, credentials or identifiers of foreign objects; one
// enrollment.request event per request (operation, outcome, closed reason, host tag), never a credential, a token or a
// payload. Status admitted is the historical T2 commitment, never a membership or session grant: a confirmed admission
// answers next sign_in, and the ordinary fresh sign-in with its current membership guards decides.
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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"

	"core/internal/domains"
	"core/internal/identity"
	"core/internal/tenancy/access"
)

// The enrollment method/path pairs (Plan r3 §5.2 = r0 §5.2's table plus the revoke route): the anonymous public
// projection, the applicant's four own-credential routes, and the manager's policy, list, decision and revocation.
const (
	RoutePublic        = "GET /internal/v1/enrollment/public"
	RouteStatus        = "GET /internal/v1/enrollment/status"
	RouteRequest       = "POST /internal/v1/enrollment/request"
	RouteAdmit         = "POST /internal/v1/enrollment/admit"
	RouteAbandon       = "POST /internal/v1/enrollment/abandon"
	RouteGetPolicy     = "GET /internal/v1/enrollment/policy"
	RouteUpdatePolicy  = "POST /internal/v1/enrollment/policy"
	RouteListRequests  = "GET /internal/v1/enrollment/requests"
	RouteDecideRequest = "POST /internal/v1/enrollment/requests/{id}/decide"
	RouteRevokeRequest = "POST /internal/v1/enrollment/requests/{id}/revoke"
)

// Paths are the served paths: every other method on them is method_not_allowed.
var Paths = []string{
	"/internal/v1/enrollment/public", "/internal/v1/enrollment/status", "/internal/v1/enrollment/request",
	"/internal/v1/enrollment/admit", "/internal/v1/enrollment/abandon", "/internal/v1/enrollment/policy",
	"/internal/v1/enrollment/requests", "/internal/v1/enrollment/requests/{id}/decide",
	"/internal/v1/enrollment/requests/{id}/revoke",
}

// maxBody is the whole-payload request budget of every mutation (Plan r0 §5.2).
const maxBody = 8 << 10

// The limiter (Plan r0 §1.7, r3 §1.7 "limiter as r0"; the MCP x/time/rate pattern): per enrollment credential 5 req/s
// burst 10 on the applicant routes, per resolved Organization 20 req/s burst 40 on the public route; each map holds at
// most maxLimiters keys (when full, only fully refilled buckets are reclaimed, else a new key is refused; see limiter).
// A key is a digest, never a credential.
const (
	credentialRate  = 5
	credentialBurst = 10
	publicRate      = 20
	publicBurst     = 40
	maxLimiters     = 10000
)

// Config is the composition configuration.
type Config struct {
	Pool     *pgxpool.Pool
	Identity *identity.Service
	Resolver *domains.Resolver
	Logger   *slog.Logger
}

// credentials is identity's applicant-credential boundary of the enrollment routes; *identity.Service satisfies it.
type credentials interface {
	EnrollmentCredential(ctx context.Context, credential, origin string) (identity.EnrollmentCredentialInfo, error)
	Membership(ctx context.Context, userID, orgID string) (bool, error)
}

// Handler is the enrollment composition: its own access facade, the identity service it binds and the routes.
type Handler struct {
	cfg    Config
	id     *identity.Service
	acc    *access.Service
	cred   credentials
	log    *slog.Logger
	own    *limiter
	public *limiter
}

// New returns the composition with its own access facade. An incomplete configuration is refused (nothing is bound).
func New(cfg Config) (*Handler, error) {
	if cfg.Pool == nil || cfg.Identity == nil || cfg.Resolver == nil {
		return nil, errors.New("enrollment api: incomplete configuration")
	}
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Handler{cfg: cfg, id: cfg.Identity, acc: access.NewService(cfg.Pool, cfg.Identity, log), cred: cfg.Identity, log: log,
		own: newLimiter(credentialRate, credentialBurst), public: newLimiter(publicRate, publicBurst)}, nil
}

// admitter is the one facade identity holds: it delegates each callback-time admission to the access facade's
// AdmitAtCallback (identity-owned DTOs only; access converts and records its own known denials).
type admitter struct{ acc *access.Service }

func (a admitter) AdmitAtCallback(ctx context.Context, h identity.EnrollmentHandle, sel identity.AdmitSelector) (identity.AdmitOutcome, error) {
	return a.acc.AdmitAtCallback(ctx, h, sel)
}

// Bind binds the admission facade into identity once, before serving. A nil or unconfigured handler, or a second
// binding, is refused; identity then keeps refusing every admission as identity_unavailable (fail closed).
func (h *Handler) Bind() error {
	if h == nil || h.id == nil || h.acc == nil {
		return errors.New("enrollment api: not configured")
	}
	return h.id.BindAdmitter(admitter{acc: h.acc})
}

// Register mounts every route on mux, a closed method_not_allowed answer for the other methods of each served path and
// a closed not_found answer at and below the /internal/v1/enrollment root.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(RoutePublic, h.publicProjection)
	mux.HandleFunc(RouteStatus, h.status)
	mux.HandleFunc(RouteRequest, h.applicant("request", h.request))
	mux.HandleFunc(RouteAdmit, h.applicant("admit", h.admit))
	mux.HandleFunc(RouteAbandon, h.applicant("abandon", h.abandon))
	mux.HandleFunc(RouteGetPolicy, h.getPolicy)
	mux.HandleFunc(RouteUpdatePolicy, h.mutation("update_policy", h.updatePolicy))
	mux.HandleFunc(RouteListRequests, h.listRequests)
	mux.HandleFunc(RouteDecideRequest, h.mutation("decide_request", h.decideRequest))
	mux.HandleFunc(RouteRevokeRequest, h.mutation("revoke_request", h.revokeRequest))
	for _, p := range Paths {
		mux.HandleFunc(p, h.closed("method", errMethod))
	}
	mux.HandleFunc("/internal/v1/enrollment", h.closed("path", errNotFound))
	mux.HandleFunc("/internal/v1/enrollment/", h.closed("path", errNotFound))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// The closed refusal codes: onboarding/api's with their values and statuses, errUnavailable under enrollment's own
// value, plus rate_limited (429), the public projection's unknown_host (404) and the applicant's enrollment_closed (403).
const (
	errInvalidRequest   = "invalid_request"
	errAmbiguous        = "ambiguous_parameter"
	errOriginRejected   = "origin_rejected"
	errUnauthenticated  = "unauthenticated"
	errCSRF             = "csrf_failed"
	errNotAvailable     = "not_available"
	errForbidden        = "forbidden"
	errNotFound         = "not_found"
	errMethod           = "method_not_allowed"
	errConflict         = "conflict"
	errUnavailable      = "enrollment_unavailable"
	errRateLimited      = "rate_limited"
	errUnknownHost      = "unknown_host"
	errEnrollmentClosed = "enrollment_closed"
)

var status = map[string]int{
	errInvalidRequest: 400, errAmbiguous: 400, errOriginRejected: 403, errUnauthenticated: 401, errCSRF: 403,
	errNotAvailable: 403, errForbidden: 403, errNotFound: 404, errMethod: 405, errConflict: 409, errUnavailable: 503,
	errRateLimited: 429, errUnknownHost: 404, errEnrollmentClosed: 403,
}

// recorder observes one answer for the enrollment.request event.
type recorder struct {
	http.ResponseWriter
	code string
}

func (h *Handler) fail(w http.ResponseWriter, code string) {
	h.refuse(w, code, "")
}

// refuse writes a closed refusal with an optional closed field code.
func (h *Handler) refuse(w http.ResponseWriter, code, field string) {
	if r, ok := w.(*recorder); ok {
		r.code = code
	}
	out := map[string]any{"error": code}
	if field != "" {
		out["field"] = field
	}
	writeJSON(w, status[code], out)
}

// hostTag is the privacy-safe log form of a hostname: the first 12 hex characters of its SHA-256 (as branding/api).
func hostTag(name string) string {
	sum := sha256.Sum256([]byte("domain/1\x00" + name))
	return hex.EncodeToString(sum[:6])
}

// event writes one enrollment.request record: operation, closed outcome and reason, and the origin's host tag.
func (h *Handler) event(ctx context.Context, op string, rec *recorder, origin *string) {
	attrs := []slog.Attr{slog.String("operation", op)}
	level := slog.LevelInfo
	switch {
	case rec.code == "":
		attrs = append(attrs, slog.String("outcome", "ok"))
	case rec.code == errUnavailable:
		level = slog.LevelError
		attrs = append(attrs, slog.String("outcome", "failed"), slog.String("reason", rec.code))
	default:
		level = slog.LevelWarn
		attrs = append(attrs, slog.String("outcome", "refused"), slog.String("reason", rec.code))
	}
	if host, ok := originHost(*origin); ok {
		attrs = append(attrs, slog.String("host_tag", hostTag(host)))
	}
	h.log.LogAttrs(ctx, level, "enrollment.request", attrs...)
}

// closed answers a method outside a served path's closed set, or a path at or below the root that is not served.
func (h *Handler) closed(op, code string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w}
		origin := ""
		defer h.event(r.Context(), op, rec, &origin)
		h.fail(rec, code)
	}
}

// originHost returns host[:port] of an exact https origin.
func originHost(origin string) (string, bool) {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" ||
		u.Fragment != "" || u.Opaque != "" || len(origin) > 300 {
		return "", false
	}
	return u.Host, true
}

// scope is one authenticated, resolved request.
type scope struct {
	ctx context.Context
	end func()
	tc  *access.TenantContext
}

// begin authenticates the browser session (and, for a mutation, its CSRF token), resolves the setup origin to the
// Organization it currently routes to, opens the access scope and resolves that Organization's context.
func (h *Handler) begin(w http.ResponseWriter, r *http.Request, origin, csrf string, mutation bool) (scope, bool) {
	v := r.Header.Values("Authorization")
	if len(v) > 1 {
		h.fail(w, errAmbiguous)
		return scope{}, false
	}
	cred := ""
	if len(v) == 1 {
		if c, ok := strings.CutPrefix(v[0], "Session "); ok && !strings.ContainsAny(c, " \t") {
			cred = c
		}
	}
	host, ok := originHost(origin)
	if !ok {
		h.fail(w, errOriginRejected)
		return scope{}, false
	}
	org, _, err := h.cfg.Resolver.Lookup(r.Context(), host)
	switch {
	case errors.Is(err, domains.ErrUnknownHost):
		h.fail(w, errOriginRejected)
		return scope{}, false
	case err != nil:
		h.fail(w, errUnavailable)
		return scope{}, false
	}
	if cred == "" {
		h.fail(w, errUnauthenticated)
		return scope{}, false
	}
	// G016 security amendment: the session must be bound to exactly this origin and to the Organization the origin
	// resolves to now (a reassigned host keeps its origin); anything else is unauthenticated, as a dead credential. The
	// CSRF comparison then inherits the binding through info.
	info, err := h.cfg.Identity.Session(r.Context(), cred, origin)
	switch {
	case errors.Is(err, identity.ErrUnauthenticated):
		h.fail(w, errUnauthenticated)
		return scope{}, false
	case err != nil:
		h.fail(w, errUnavailable)
		return scope{}, false
	case info.OrganizationID != org:
		h.fail(w, errUnauthenticated)
		return scope{}, false
	}
	if mutation && !identity.CSRFMatches(info, csrf) {
		h.fail(w, errCSRF)
		return scope{}, false
	}
	actx, end, err := h.acc.BeginRequest(r.Context(), cred, origin)
	if err != nil {
		h.fail(w, failure(err))
		return scope{}, false
	}
	tc, err := h.acc.Resolve(actx, access.ResolveRequest{OrganizationID: org})
	if err != nil {
		end()
		h.fail(w, failure(err))
		return scope{}, false
	}
	return scope{ctx: actx, end: end, tc: tc}, true
}

// failure maps a facade error onto a closed refusal; any unavailable (possibly durable: a class-40 abort or an unknown
// COMMIT outcome the facade reports as unavailable) is enrollment_unavailable, never a guessed success.
func failure(err error) string {
	var ae *access.Error
	if !errors.As(err, &ae) {
		return errUnavailable
	}
	switch ae.Code {
	case access.CodeUnauthenticated:
		return errUnauthenticated
	case access.CodeContextDenied:
		return errNotAvailable
	case access.CodeForbidden:
		return errForbidden
	case access.CodeNotFound:
		return errNotFound
	case access.CodeConflict:
		return errConflict
	case access.CodeInvalidInput:
		return errInvalidRequest
	}
	return errUnavailable
}

// refuseErr writes the refusal of a facade error with its closed field code (invalid_request, forbidden and conflict
// only). The applicant's enrollment_closed conflict is its own 403 code without a field (Plan r0 §5.2).
func (h *Handler) refuseErr(w http.ResponseWriter, err error) {
	code := failure(err)
	var ae *access.Error
	if errors.As(err, &ae) && code == errConflict && ae.Field == errEnrollmentClosed {
		h.fail(w, errEnrollmentClosed)
		return
	}
	if errors.As(err, &ae) && (code == errInvalidRequest || code == errForbidden || code == errConflict) {
		h.refuse(w, code, ae.Field)
		return
	}
	h.fail(w, code)
}

// query returns the single-valued query parameters of a GET: origin is required; the optional ones appear at most once;
// any other parameter is ambiguous.
func query(r *http.Request, optional ...string) (map[string]string, bool) {
	q := r.URL.Query()
	out := map[string]string{}
	for k, v := range q {
		if len(v) != 1 || (k != "origin" && !strings.Contains(" "+strings.Join(optional, " ")+" ", " "+k+" ")) {
			return nil, false
		}
		out[k] = v[0]
	}
	_, ok := out["origin"]
	return out, ok
}

// page parses the optional limit (absent: 0, the facade's default page of 100); a malformed value is invalid_request
// field invalid_page_size, and the facade decides the range.
func page(q map[string]string) (int, bool) {
	v, ok := q["limit"]
	if !ok {
		return 0, true
	}
	n, err := strconv.Atoi(v)
	return n, err == nil && strconv.Itoa(n) == v
}

// parseID parses a canonical identifier into dst with the facade's parser (tenancy's single canonical UUID parser);
// dst is any of the facade's [16]byte id types, so the adapter never imports tenancy.
func parseID[T ~[16]byte](dst *T, s string) bool {
	g, err := access.ParseGroupID(s)
	if err != nil {
		return false
	}
	*dst = T(g)
	return true
}

// formatID is the canonical lowercase hyphenated form of any [16]byte id (the facade's String of the same bytes).
func formatID[T ~[16]byte](id T) string { return access.GroupID(id).String() }

// withMode calls a facade operation that takes the facade's enrollment-mode type with the body's mode (the facade
// validates it: invalid_request field invalid_enrollment_mode), so the adapter never imports tenancy.
func withMode[M ~string, R any](f func(context.Context, *access.TenantContext, M) (R, error), ctx context.Context,
	tc *access.TenantContext, mode string) (R, error) {
	return f(ctx, tc, M(mode))
}

// timestamp renders a time as RFC 3339 UTC; nullableTime renders an optional time as JSON null when absent.
func timestamp(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return timestamp(*t)
}

// ownRequestJSON is the applicant's view of their own request: no subject, decider, revoker, binding, authority or
// other tenant column (minimal disclosure, ENR-17).
func ownRequestJSON(v access.EnrollmentRequestView) map[string]any {
	return map[string]any{"id": formatID(v.ID), "status": v.Status, "expired": v.Expired, "consumed": v.Consumed,
		"created_at": timestamp(v.CreatedAt), "expires_at": timestamp(v.ExpiresAt), "decided_at": nullableTime(v.DecidedAt),
		"revoked_at": nullableTime(v.RevokedAt)}
}

// requestJSON is a manager's view: the applicant's own view plus the applicant's user id and display name (always null:
// no profile read is widened; Codex round-8 disposition 4).
func requestJSON(v access.EnrollmentRequestView) map[string]any {
	out := ownRequestJSON(v)
	out["user_id"] = formatID(v.UserID)
	out["display_name"] = nil
	if v.DisplayName != nil {
		out["display_name"] = *v.DisplayName
	}
	return out
}

// limiter is a bounded set of token buckets keyed by a digest. The map is never cleared and a bucket that is not fully
// refilled is never dropped, so filling or churning the map with unrelated keys cannot restore a depleted bucket. When
// the map holds maxLimiters keys a new key triggers a reclaim scan that deletes only fully refilled buckets (dropping
// one is indistinguishable from recreating it); if the scan frees nothing the new key is refused (rate_limited) and no
// existing bucket is touched. Every scan, successful or not, defers the next one by the full-refill time burst/rate, so
// new keys at capacity inside that window are refused without scanning (at most one O(maxLimiters) scan per window).
// With no refill (rate <= 0) a used bucket can never become fully refilled: there is no scan and no window, and a new
// key at capacity is refused.
type limiter struct {
	mu    sync.Mutex
	m     map[string]*rate.Limiter
	r     rate.Limit
	burst int
	now   func() time.Time // the clock (time.Now; a test injects a fake), read once per allow under mu
	full  time.Duration    // burst/rate, the full-refill time and scan window; 0 when rate <= 0 (no reclaim)
	next  time.Time        // the earliest instant the next reclaim scan may run
}

func newLimiter(perSecond, burst int) *limiter {
	l := &limiter{m: map[string]*rate.Limiter{}, r: rate.Limit(perSecond), burst: burst, now: time.Now}
	if perSecond > 0 {
		l.full = time.Duration(float64(burst) / float64(perSecond) * float64(time.Second))
	}
	return l
}

// allow takes one token from key's bucket. A new key at capacity is refused unless reclaim frees room.
func (l *limiter) allow(key string) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.m[key]
	if !ok {
		if len(l.m) >= maxLimiters && !l.reclaim(now) {
			return false
		}
		b = rate.NewLimiter(l.r, l.burst)
		l.m[key] = b
	}
	return b.AllowN(now, 1)
}

// reclaim (l.mu held) deletes the fully refilled buckets at now and reports whether any was freed. It scans only at or
// after l.next and then moves l.next to now+full whatever the outcome; with no refill it never scans.
func (l *limiter) reclaim(now time.Time) bool {
	if l.full <= 0 || now.Before(l.next) {
		return false
	}
	l.next = now.Add(l.full)
	burst, freed := float64(l.burst), false
	for k, b := range l.m {
		if b.TokensAt(now) >= burst {
			delete(l.m, k)
			freed = true
		}
	}
	return freed
}

// keyOf is the limiter key of a forwarded credential (its SHA-256; the credential itself is never kept).
func keyOf(credential string) string {
	sum := sha256.Sum256([]byte("enrollment/1\x00" + credential))
	return hex.EncodeToString(sum[:])
}

// publicProjection is GET /internal/v1/enrollment/public?origin= (no credential): the Organization the origin's host
// routes to, its mode and public label. 200 {"mode", "label"}; a host that routes nowhere is 404 unknown_host.
func (h *Handler) publicProjection(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "public", rec, &origin)
	q, ok := query(r)
	if !ok {
		h.fail(rec, errAmbiguous)
		return
	}
	origin = q["origin"]
	host, ok := originHost(origin)
	if !ok {
		h.fail(rec, errOriginRejected)
		return
	}
	org, _, err := h.cfg.Resolver.Lookup(r.Context(), host)
	switch {
	case errors.Is(err, domains.ErrUnknownHost):
		h.fail(rec, errUnknownHost)
		return
	case err != nil:
		h.fail(rec, errUnavailable)
		return
	}
	if !h.public.allow(org) {
		h.fail(rec, errRateLimited)
		return
	}
	p, err := h.acc.PublicEnrollment(r.Context(), org)
	if err != nil {
		h.fail(rec, failure(err))
		return
	}
	var label any
	if p.Label != nil {
		label = *p.Label
	}
	writeJSON(rec, http.StatusOK, map[string]any{"mode": p.Mode, "label": label})
}

// beginOwn authenticates the applicant's own enrollment credential ("Authorization: Enrollment <credential>"; and, for a
// mutation, its CSRF token): a repeated header is ambiguous; the origin must be an exact https origin that routes now
// (else origin_rejected); a missing credential is unauthenticated; the per-credential limiter answers rate_limited; the
// credential must resolve at exactly this origin and be bound to the Organization the origin routes to now (a dead,
// foreign-origin or reassigned credential is unauthenticated, as a dead one). It never constructs a tenant context.
func (h *Handler) beginOwn(w http.ResponseWriter, r *http.Request, origin, csrf string, mutation bool) (identity.EnrollmentCredentialInfo, bool) {
	v := r.Header.Values("Authorization")
	if len(v) > 1 {
		h.fail(w, errAmbiguous)
		return identity.EnrollmentCredentialInfo{}, false
	}
	cred := ""
	if len(v) == 1 {
		if c, ok := strings.CutPrefix(v[0], "Enrollment "); ok && !strings.ContainsAny(c, " \t") {
			cred = c
		}
	}
	host, ok := originHost(origin)
	if !ok {
		h.fail(w, errOriginRejected)
		return identity.EnrollmentCredentialInfo{}, false
	}
	org, _, err := h.cfg.Resolver.Lookup(r.Context(), host)
	switch {
	case errors.Is(err, domains.ErrUnknownHost):
		h.fail(w, errOriginRejected)
		return identity.EnrollmentCredentialInfo{}, false
	case err != nil:
		h.fail(w, errUnavailable)
		return identity.EnrollmentCredentialInfo{}, false
	}
	if cred == "" {
		h.fail(w, errUnauthenticated)
		return identity.EnrollmentCredentialInfo{}, false
	}
	if !h.own.allow(keyOf(cred)) {
		h.fail(w, errRateLimited)
		return identity.EnrollmentCredentialInfo{}, false
	}
	info, err := h.cred.EnrollmentCredential(r.Context(), cred, origin)
	switch {
	case errors.Is(err, identity.ErrUnauthenticated):
		h.fail(w, errUnauthenticated)
		return identity.EnrollmentCredentialInfo{}, false
	case err != nil:
		h.fail(w, errUnavailable)
		return identity.EnrollmentCredentialInfo{}, false
	case info.Handle.OrganizationID != org:
		h.fail(w, errUnauthenticated)
		return identity.EnrollmentCredentialInfo{}, false
	}
	if mutation && !identity.EnrollmentCSRFMatches(info, csrf) {
		h.fail(w, errCSRF)
		return identity.EnrollmentCredentialInfo{}, false
	}
	return info, true
}

// status is GET /internal/v1/enrollment/status?origin= (own enrollment credential): the reconciliation read. 200
// {"status", "request": {…}|null, "member": bool, "csrf_token"}: status is admitted (the credential's historical T2
// proof), the latest request's state, or none; member is the current membership re-read on every call (an admitted but
// removed applicant is member false; nothing here grants or restores access); csrf_token is the credential's own token
// for the BFF's forms. A dead credential is 401 unauthenticated (the BFF clears the enrollment cookie).
func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "status", rec, &origin)
	q, ok := query(r)
	if !ok {
		h.fail(rec, errAmbiguous)
		return
	}
	origin = q["origin"]
	info, ok := h.beginOwn(rec, r, origin, "", false)
	if !ok {
		return
	}
	st, err := h.acc.EnrollmentStatus(r.Context(), info.Handle)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	member, err := h.cred.Membership(r.Context(), info.Handle.UserID, info.Handle.OrganizationID)
	if err != nil {
		h.fail(rec, errUnavailable)
		return
	}
	var req any
	if st.Request != nil {
		req = ownRequestJSON(*st.Request)
	}
	writeJSON(rec, http.StatusOK, map[string]any{"status": st.Status, "request": req, "member": member,
		"csrf_token": base64.RawURLEncoding.EncodeToString(info.CSRFToken)})
}

// body is every JSON mutation's body; each route reads its own fields and refuses the others.
type body struct {
	Origin    string  `json:"origin"`
	CSRFToken string  `json:"csrf_token"`
	Kind      *string `json:"kind"`
	RequestID *string `json:"request_id"`
	Mode      *string `json:"mode"`
	Decision  *string `json:"decision"`
}

// fields names the optional members present in the body.
func (in body) fields() map[string]bool {
	out := map[string]bool{}
	for name, set := range map[string]bool{"kind": in.Kind != nil, "request_id": in.RequestID != nil, "mode": in.Mode != nil,
		"decision": in.Decision != nil} {
		if set {
			out[name] = true
		}
	}
	return out
}

// shape reports whether every required member is present and nothing outside required and optional is.
func (in body) shape(required, optional []string) bool {
	got := in.fields()
	for _, n := range required {
		if !got[n] {
			return false
		}
		delete(got, n)
	}
	for _, n := range optional {
		delete(got, n)
	}
	return len(got) == 0
}

// decode reads at most maxBody bytes (more is invalid_request before any decode, credential or write) of a query-free
// application/json mutation and decodes exactly one JSON object with known members only.
func (h *Handler) decode(w *recorder, r *http.Request) (body, bool) {
	if len(r.URL.Query()) != 0 {
		h.fail(w, errAmbiguous)
		return body{}, false
	}
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "application/json" {
		h.fail(w, errInvalidRequest)
		return body{}, false
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil || len(raw) > maxBody {
		h.fail(w, errInvalidRequest)
		return body{}, false
	}
	var in body
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	// Exactly one JSON value: the second Decode must report io.EOF (trailing whitespace is fine).
	if err := dec.Decode(&in); err != nil || !errors.Is(dec.Decode(new(json.RawMessage)), io.EOF) {
		h.fail(w, errInvalidRequest)
		return body{}, false
	}
	return in, true
}

// applicant is an applicant mutation: decode, then the own credential with its CSRF token, then act; one
// enrollment.request.
func (h *Handler) applicant(op string, act func(w *recorder, r *http.Request, info identity.EnrollmentCredentialInfo, in body)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec, origin := &recorder{ResponseWriter: w}, ""
		defer h.event(r.Context(), op, rec, &origin)
		in, ok := h.decode(rec, r)
		if !ok {
			return
		}
		origin = in.Origin
		info, ok := h.beginOwn(rec, r, in.Origin, in.CSRFToken, true)
		if !ok {
			return
		}
		act(rec, r, info, in)
	}
}

// request is POST /internal/v1/enrollment/request {origin, csrf_token}: the applicant's own access request. 200
// {"request"} (new or the existing pending one; idempotent); 403 enrollment_closed; 409 conflict (already_member,
// host_changed, recipient_ineligible); 503 enrollment_unavailable (aborted or an unknown COMMIT: a repeat returns the
// same pending request).
func (h *Handler) request(w *recorder, r *http.Request, info identity.EnrollmentCredentialInfo, in body) {
	if !in.shape(nil, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.RequestEnrollment(r.Context(), info.Handle)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"request": ownRequestJSON(v)})
}

// admit is POST /internal/v1/enrollment/admit {origin, csrf_token, kind: "open"|"request", request_id?}: the applicant's
// own admission. 200 {"outcome": "admitted"|"already_member", "next": "sign_in"} on a confirmed admission (never a
// session: the BFF clears the enrollment cookie and starts the ordinary sign-in); 403 enrollment_closed; 409 conflict
// with the closed outcome (enrollment_restricted, approval_revoked, pending, denied, recipient_ineligible, host_changed);
// 404 not_found; 401 unauthenticated (expired, revoked); 503 enrollment_unavailable (unknown outcome: the next status
// read reconciles and a repeat answers already_member without a second grant).
func (h *Handler) admit(w *recorder, r *http.Request, info identity.EnrollmentCredentialInfo, in body) {
	var (
		a   access.EnrollmentAdmission
		err error
	)
	switch {
	case in.shape([]string{"kind"}, nil) && *in.Kind == "open":
		a, err = h.acc.AdmitOpen(r.Context(), info.Handle)
	case in.shape([]string{"kind", "request_id"}, nil) && *in.Kind == "request":
		var id access.EnrollmentRequestView
		if !parseID(&id.ID, *in.RequestID) {
			h.refuse(w, errInvalidRequest, "request_id")
			return
		}
		a, err = h.acc.AdmitRequest(r.Context(), info.Handle, id.ID)
	default:
		h.refuse(w, errInvalidRequest, "kind")
		return
	}
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"outcome": a.Outcome, "next": identity.NextSignIn})
}

// abandon is POST /internal/v1/enrollment/abandon {origin, csrf_token}: revokes the applicant's own credential
// (idempotent; a pending request stays for an administrator). 200 {}.
func (h *Handler) abandon(w *recorder, r *http.Request, info identity.EnrollmentCredentialInfo, in body) {
	if !in.shape(nil, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	if err := h.acc.AbandonEnrollment(r.Context(), info.Handle); err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

// read begins a session GET of op: the query, then the session scope.
func (h *Handler) read(w *recorder, r *http.Request, origin *string, optional ...string) (scope, map[string]string, bool) {
	q, ok := query(r, optional...)
	if !ok {
		h.fail(w, errAmbiguous)
		return scope{}, nil, false
	}
	*origin = q["origin"]
	s, ok := h.begin(w, r, q["origin"], "", false)
	return s, q, ok
}

// mutation is a manager mutation: decode, then the session with its CSRF token and the context, then act.
func (h *Handler) mutation(op string, act func(w *recorder, r *http.Request, s scope, in body)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec, origin := &recorder{ResponseWriter: w}, ""
		defer h.event(r.Context(), op, rec, &origin)
		in, ok := h.decode(rec, r)
		if !ok {
			return
		}
		origin = in.Origin
		s, ok := h.begin(rec, r, in.Origin, in.CSRFToken, true)
		if !ok {
			return
		}
		defer s.end()
		act(rec, r, s, in)
	}
}

// policyJSON renders the enrollment policy (selected-design addendum §A.5): the mode, whether open may be enabled and
// the recorded basis (null while unresolved). No legacy details or user data.
func policyJSON(mode string, openReady bool, basis string) map[string]any {
	var b any
	if basis != "" {
		b = basis
	}
	return map[string]any{"mode": mode, "open_ready": openReady, "history_basis": b}
}

// getPolicy is GET /internal/v1/enrollment/policy?origin= (session; organization.read).
func (h *Handler) getPolicy(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "get_policy", rec, &origin)
	s, _, ok := h.read(rec, r, &origin)
	if !ok {
		return
	}
	defer s.end()
	p, err := h.acc.GetEnrollmentPolicy(s.ctx, s.tc)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, policyJSON(string(p.Mode), p.OpenReady, p.HistoryBasis))
}

// updatePolicy is POST /internal/v1/enrollment/policy {origin, csrf_token, mode} (session; organization.manage_policy).
// 200 {"mode", "open_ready", "history_basis"}; 403 forbidden; 409 conflict field enrollment_not_ready (open on an
// unresolved history; nothing written).
func (h *Handler) updatePolicy(w *recorder, r *http.Request, s scope, in body) {
	if !in.shape([]string{"mode"}, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	p, err := withMode(h.acc.UpdateEnrollmentPolicy, s.ctx, s.tc, *in.Mode)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, policyJSON(string(p.Mode), p.OpenReady, p.HistoryBasis))
}

// listRequests is GET /internal/v1/enrollment/requests?origin=&after=&limit=&status= (session;
// organization.manage_members): {"requests": […], "next"} (≤ 100; keyset by id).
func (h *Handler) listRequests(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "list_requests", rec, &origin)
	s, q, ok := h.read(rec, r, &origin, "after", "limit", "status")
	if !ok {
		return
	}
	defer s.end()
	limit, ok := page(q)
	if !ok {
		h.refuse(rec, errInvalidRequest, "invalid_page_size")
		return
	}
	p, err := h.acc.ListEnrollmentRequests(s.ctx, s.tc, q["after"], limit, q["status"])
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	rows := make([]map[string]any, 0, len(p.Requests))
	for _, v := range p.Requests {
		rows = append(rows, requestJSON(v))
	}
	out := map[string]any{"requests": rows, "next": nil}
	if p.Next != nil {
		out["next"] = formatID(*p.Next)
	}
	writeJSON(rec, http.StatusOK, out)
}

// decideRequest is POST /internal/v1/enrollment/requests/{id}/decide {origin, csrf_token, decision: "approve"|"deny"}
// (session; organization.manage_members). 200 {"request"}; 409 conflict field not_pending; 404 not_found.
func (h *Handler) decideRequest(w *recorder, r *http.Request, s scope, in body) {
	var id access.EnrollmentRequestView
	if !parseID(&id.ID, r.PathValue("id")) {
		h.fail(w, errNotFound)
		return
	}
	if !in.shape([]string{"decision"}, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.DecideEnrollmentRequest(s.ctx, s.tc, id.ID, *in.Decision)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"request": requestJSON(v)})
}

// revokeRequest is POST /internal/v1/enrollment/requests/{id}/revoke {origin, csrf_token} (session;
// organization.manage_members). 200 {"request"} (idempotent on an already revoked approval); 409 conflict field
// already_consumed or not_approved; 404 not_found.
func (h *Handler) revokeRequest(w *recorder, r *http.Request, s scope, in body) {
	var id access.EnrollmentRequestView
	if !parseID(&id.ID, r.PathValue("id")) {
		h.fail(w, errNotFound)
		return
	}
	if !in.shape(nil, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.RevokeEnrollmentApproval(s.ctx, s.tc, id.ID)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"request": requestJSON(v)})
}
