// Package api is the internal HTTP surface of the G016 domain lifecycle (Plan r4 §5.3; D-5 M2: internal routes only,
// web and the trusted edge are the only callers). It composes the access facade (authorization and the guarded
// writes), identity (the browser session and its CSRF token) and the domains package (resolver, DNS checker, TLS
// adapter); every DNS/TLS step runs between two facade calls, outside every transaction. Every answer is no-store JSON
// without CORS headers. Hostnames appear in logs only as a 12-hex-character hash prefix; challenge values never.
package api

import (
	"context"
	"crypto/sha256"
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

	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/hostname"
	"core/internal/identity"
	"core/internal/tenancy/access"
)

// The routes.
const (
	RouteResolve         = "GET /internal/v1/hosts/resolve"
	RouteTLSAsk          = "GET /internal/v1/domains/tls-ask"
	RouteList            = "GET /internal/v1/domains"
	RouteClaim           = "POST /internal/v1/domains"
	RouteRotate          = "POST /internal/v1/domains/rotate"
	RouteVerify          = "POST /internal/v1/domains/verify"
	RouteRecheck         = "POST /internal/v1/domains/recheck"
	RouteRemove          = "POST /internal/v1/domains/remove"
	RouteRetryWithdrawal = "POST /internal/v1/domains/retry-withdrawal"
	RouteSetSlug         = "POST /internal/v1/organization/slug"
)

const maxBody = 4 << 10

// Config is the handler configuration.
type Config struct {
	Pool     *pgxpool.Pool
	Identity *identity.Service
	Resolver *domains.Resolver
	Checker  domains.Checker
	TLS      domains.TLSAdapter
	Domains  config.Domains
	Logger   *slog.Logger
}

// Handler serves the routes.
type Handler struct {
	cfg Config
	acc *access.Service
	log *slog.Logger
}

// New returns a handler. Its access facade is its own (like the MCP handler's).
func New(cfg Config) (*Handler, error) {
	if cfg.Pool == nil || cfg.Identity == nil || cfg.Resolver == nil || cfg.Checker == nil || cfg.TLS == nil {
		return nil, errors.New("domains api: incomplete configuration")
	}
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Handler{cfg: cfg, acc: access.NewService(cfg.Pool, cfg.Identity, log), log: log}, nil
}

// Register mounts every route on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(RouteResolve, h.Resolve)
	mux.HandleFunc(RouteTLSAsk, h.TLSAsk)
	mux.HandleFunc(RouteList, h.List)
	mux.HandleFunc(RouteClaim, h.mutation("claim", h.claim))
	mux.HandleFunc(RouteRotate, h.mutation("rotate", h.rotate))
	mux.HandleFunc(RouteVerify, h.mutation("verify", h.check(domains.CheckVerify)))
	mux.HandleFunc(RouteRecheck, h.mutation("recheck", h.check(domains.CheckRecheck)))
	mux.HandleFunc(RouteRemove, h.mutation("remove", h.remove))
	mux.HandleFunc(RouteRetryWithdrawal, h.mutation("retry_withdrawal", h.retry))
	mux.HandleFunc(RouteSetSlug, h.mutation("set_slug", h.setSlug))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// The closed refusal codes.
const (
	errInvalidRequest  = "invalid_request"
	errAmbiguous       = "ambiguous_parameter"
	errUnsupported     = "unsupported_media_type"
	errOriginRejected  = "origin_rejected"
	errUnauthenticated = "unauthenticated"
	errCSRF            = "csrf_failed"
	errNotAvailable    = "not_available"
	errForbidden       = "forbidden"
	errNotFound        = "not_found"
	errUnknownHost     = "unknown_host"
	errHostInvalid     = "host_invalid"
	errHostReserved    = "host_reserved"
	errHostTaken       = "host_taken"
	errSlugReserved    = "slug_reserved"
	errSlugTaken       = "slug_taken"
	errStateInvalid    = "domain_state_invalid"
	errRetryExhausted  = "domain_retry_exhausted"
	errUnavailable     = "domains_unavailable"
)

var status = map[string]int{
	errInvalidRequest: 400, errAmbiguous: 400, errUnsupported: 415, errOriginRejected: 403, errUnauthenticated: 401,
	errCSRF: 403, errNotAvailable: 403, errForbidden: 403, errNotFound: 404, errUnknownHost: 404, errHostInvalid: 400,
	errHostReserved: 400, errHostTaken: 409, errSlugReserved: 400, errSlugTaken: 409, errStateInvalid: 409,
	errRetryExhausted: 409, errUnavailable: 503,
}

// recorder observes one answer for the domain.request event.
type recorder struct {
	http.ResponseWriter
	code, reason string
}

func (h *Handler) fail(w http.ResponseWriter, code string) {
	if r, ok := w.(*recorder); ok {
		r.code = code
	}
	writeJSON(w, status[code], map[string]string{"error": code})
}

// hostTag is the privacy-safe log form of a hostname: the first 12 hex characters of its SHA-256.
func hostTag(name string) string {
	sum := sha256.Sum256([]byte("domain/1\x00" + name))
	return hex.EncodeToString(sum[:6])
}

// event writes one domain.request record: operation, closed outcome and reason, and the hostname tag when known.
func (h *Handler) event(ctx context.Context, op string, rec *recorder, host string) {
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
	if rec.reason != "" {
		attrs = append(attrs, slog.String("detail", rec.reason))
	}
	if host != "" {
		attrs = append(attrs, slog.String("host_tag", hostTag(host)))
	}
	h.log.LogAttrs(ctx, level, "domain.request", attrs...)
}

// Resolve is GET /internal/v1/hosts/resolve?origin=https://host[:port] (no credential): the routing decision of one
// origin for web's host gate and the callback relay. 200 {kind, organization_id?, callback_origin?, signup_origin?}; 404
// unknown_host for a host that routes nowhere; 503 on a store failure (never a clean unknown host). G062 (Plan r5 §1.1,
// §4.1): the platform domain and www.<domain> answer {kind: public_entry, signup_origin} with the exact configured
// platform root origin (www canonicalizes to it); there is no redirect location. G071 (Plan r4 S5): while a default
// tenant is pinned, the exact root answers {kind: platform, organization_id, root_alias: true} instead; root_alias is
// absent from every other answer. G071 (Plan r4 S16): every public_entry answer (the unaliased root and www) also
// carries root_page, the effective multi-tenant root presentation: exactly "signup_redirect" or "landing"; no other
// answer carries it.
func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()["origin"]
	if len(q) != 1 {
		h.fail(w, errAmbiguous)
		return
	}
	host, ok := originHost(q[0])
	if !ok {
		h.fail(w, errUnknownHost)
		return
	}
	res, err := h.cfg.Resolver.Resolve(r.Context(), host)
	if err != nil {
		h.fail(w, errUnavailable)
		return
	}
	if res.Host.Origin() != q[0] {
		h.fail(w, errUnknownHost) // the origin must already be exact (normalized) to be answered
		return
	}
	out := map[string]any{}
	switch res.Kind {
	case hostname.KindPublicEntry:
		// G071 S16: the mode is the effective one the server stored in Domains.RootPage before building the handlers
		// (domains.RootPresentation, idempotent on it); resolving it again here keeps the member closed (unset is
		// signup_redirect). The aliased root never reaches this case: the resolver answers it as platform first (S4).
		mode, _ := domains.RootPresentation(h.cfg.Domains.RootPage)
		out["kind"], out["signup_origin"], out["root_page"] = "public_entry", h.cfg.Domains.RootOrigin(), mode
	case hostname.KindReserved:
		out["kind"] = "reserved"
		if h.cfg.Domains.AuthHost != "" && res.Host.Name == h.cfg.Domains.AuthHost {
			out["kind"] = "auth"
		}
	case hostname.KindPlatform, hostname.KindCustom:
		if !res.Routing {
			h.fail(w, errUnknownHost)
			return
		}
		out["kind"], out["organization_id"] = string(res.Kind), res.OrganizationID
		if res.RootAlias {
			out["root_alias"] = true // G071 S5: present only on the active root alias's platform answer
		}
		if res.Kind == hostname.KindCustom && h.cfg.Domains.AuthOrigin() != "" {
			out["callback_origin"] = h.cfg.Domains.AuthOrigin()
		}
	default:
		h.fail(w, errUnknownHost)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// TLSAsk is the E1 on-demand TLS permission (Caddy's "ask"): GET /internal/v1/domains/tls-ask?domain=<hostname>. 200
// only for a custom hostname whose live row has a fresh ownership proof; 403 otherwise; 503 on a store failure. No
// body. A removed hostname is refused from the removal commit on.
func (h *Handler) TLSAsk(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()["domain"]
	w.Header().Set("Cache-Control", "no-store")
	if len(q) != 1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	ok, err := h.cfg.Resolver.IssuanceAllowed(r.Context(), q[0])
	switch {
	case err != nil:
		w.WriteHeader(http.StatusServiceUnavailable)
	case ok:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusForbidden)
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

// failure maps a facade error onto a closed refusal.
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
	case access.CodeHostTaken:
		return errHostTaken
	case access.CodeConflict:
		return errSlugTaken
	case access.CodeDomainStateInvalid:
		return errStateInvalid
	case access.CodeDomainRetryExhausted:
		return errRetryExhausted
	case access.CodeInvalidInput:
		switch ae.Field {
		case "host_invalid":
			return errHostInvalid
		case "slug_reserved":
			return errSlugReserved
		}
		return errInvalidRequest
	}
	return errUnavailable
}

func viewJSON(v access.DomainView) map[string]any {
	t := func(p *time.Time) any {
		if p == nil {
			return nil
		}
		return p.UTC().Format(time.RFC3339)
	}
	out := map[string]any{"id": v.ID, "hostname": v.Hostname, "generation": v.Generation, "state": string(v.State),
		"verified_until": t(v.VerifiedUntil), "tls_state": string(v.TLSState), "last_check_at": t(v.LastCheckAt),
		"last_check_outcome": nil, "check_failures": v.CheckFailures, "created_at": v.CreatedAt.UTC().Format(time.RFC3339),
		"removed_at": t(v.RemovedAt)}
	if v.LastCheckOutcome != "" {
		out["last_check_outcome"] = v.LastCheckOutcome
	}
	return out
}

func (h *Handler) claimJSON(c access.DomainClaim) map[string]any {
	out := viewJSON(c.DomainView)
	out["challenge"] = map[string]any{"type": "TXT", "name": c.ChallengeName, "value": c.ChallengeValue,
		"expires_at": c.ChallengeExpiresAt.UTC().Format(time.RFC3339)}
	out["cname_target"] = h.cfg.Domains.CNAMETarget
	return out
}

// List is GET /internal/v1/domains?origin= (session): the active Organization's domain rows (owner or admin).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "list", rec, "")
	q := r.URL.Query()["origin"]
	if len(q) != 1 {
		h.fail(rec, errAmbiguous)
		return
	}
	s, ok := h.begin(rec, r, q[0], "", false)
	if !ok {
		return
	}
	defer s.end()
	views, err := h.acc.ListDomains(s.ctx, s.tc)
	if err != nil {
		h.fail(rec, failure(err))
		return
	}
	out := make([]map[string]any, 0, len(views))
	for _, v := range views {
		out = append(out, viewJSON(v))
	}
	writeJSON(rec, http.StatusOK, map[string]any{"domains": out})
}

// body is every mutation's JSON body; each route reads its own fields.
type body struct {
	Origin    string `json:"origin"`
	CSRFToken string `json:"csrf_token"`
	ID        string `json:"id"`
	Hostname  string `json:"hostname"`
	Slug      string `json:"slug"`
}

// mutation decodes the body, authenticates with CSRF, resolves the context and runs act; it writes one domain.request.
func (h *Handler) mutation(op string, act func(w *recorder, s scope, in body) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w}
		host := ""
		defer func() { h.event(r.Context(), op, rec, host) }()
		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mt != "application/json" {
			h.fail(rec, errUnsupported)
			return
		}
		var in body
		dec := json.NewDecoder(io.LimitReader(r.Body, maxBody+1))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil || dec.More() {
			h.fail(rec, errInvalidRequest)
			return
		}
		s, ok := h.begin(rec, r, in.Origin, in.CSRFToken, true)
		if !ok {
			return
		}
		defer s.end()
		host = act(rec, s, in)
	}
}

func (h *Handler) claim(w *recorder, s scope, in body) string {
	n, err := hostname.Normalize(in.Hostname, "")
	if err != nil || !strings.Contains(n.Name, ".") {
		h.fail(w, errHostInvalid)
		return ""
	}
	hc := h.cfg.Domains.Hostnames()
	if !hostname.CustomClaimAllowed(n.Name, hc) || domains.Refused(n.Name, h.proofConfig()) {
		h.fail(w, errHostReserved)
		return n.Name
	}
	c, err := h.acc.ClaimDomain(s.ctx, s.tc, n.Name, h.cfg.Domains.ChallengeTTL, h.cfg.Domains.ChallengeLabel)
	if err != nil {
		h.fail(w, failure(err))
		return n.Name
	}
	writeJSON(w, http.StatusCreated, map[string]any{"domain": h.claimJSON(c)})
	return n.Name
}

func (h *Handler) rotate(w *recorder, s scope, in body) string {
	c, err := h.acc.RotateChallenge(s.ctx, s.tc, in.ID, h.cfg.Domains.ChallengeTTL, h.cfg.Domains.ChallengeLabel)
	if err != nil {
		h.fail(w, failure(err))
		return ""
	}
	writeJSON(w, http.StatusOK, map[string]any{"domain": h.claimJSON(c)})
	return c.Hostname
}

func (h *Handler) proofConfig() domains.ProofConfig {
	d := h.cfg.Domains
	return domains.ProofConfig{ChallengeLabel: d.ChallengeLabel, CNAMETarget: d.CNAMETarget, BaseDomain: d.BaseDomain, Timeout: d.DNSTimeout}
}

// check runs one verify or recheck: the facade's read step, the DNS proof (and, for a recheck with a valid proof, the TLS
// status of the row's own DomainRef) outside every transaction, then the facade's guarded write.
func (h *Handler) check(c domains.Check) func(w *recorder, s scope, in body) string {
	return func(w *recorder, s scope, in body) string {
		rec, err := h.acc.BeginDomainCheck(s.ctx, s.tc, in.ID, c)
		if err != nil {
			h.fail(w, failure(err))
			return ""
		}
		obs := domains.Observation{Proof: domains.Proof(s.ctx, h.cfg.Checker, h.proofConfig(), rec.Hostname, rec.ChallengeValue)}
		if c == domains.CheckRecheck && obs.Proof == domains.OutcomeOK {
			obs.TLS = h.cfg.TLS.Status(s.ctx, rec.Ref())
		}
		v, err := h.acc.RecordDomainCheck(s.ctx, s.tc, rec, c, obs, h.cfg.Domains.ProofFreshness)
		if err != nil {
			h.fail(w, failure(err))
			return rec.Hostname
		}
		writeJSON(w, http.StatusOK, map[string]any{"domain": viewJSON(v)})
		return rec.Hostname
	}
}

// withdraw calls the adapter's Remove on the removed row's own DomainRef after the removal (or reservation) committed and
// records the outcome on that row; a result for another row or generation is dropped (stale_result).
func (h *Handler) withdraw(w *recorder, s scope, removed domains.Record, reserved bool) access.DomainView {
	res := h.cfg.TLS.Remove(s.ctx, removed.Ref())
	v, err := h.acc.RecordWithdrawal(s.ctx, s.tc, removed, res, reserved)
	if err != nil {
		if access.CodeOf(err) == access.CodeDomainStateInvalid {
			w.reason = "stale_result"
		} else {
			w.reason = "withdrawal_unrecorded"
		}
		return access.DomainView{ID: removed.ID, Hostname: removed.Hostname, Generation: removed.Generation, State: removed.State,
			TLSState: removed.TLSState, CheckFailures: removed.CheckFailures, CreatedAt: removed.CreatedAt, RemovedAt: removed.RemovedAt}
	}
	return v
}

func (h *Handler) remove(w *recorder, s scope, in body) string {
	removed, err := h.acc.RemoveDomain(s.ctx, s.tc, in.ID)
	if err != nil {
		h.fail(w, failure(err))
		return ""
	}
	writeJSON(w, http.StatusOK, map[string]any{"domain": viewJSON(h.withdraw(w, s, removed, false))})
	return removed.Hostname
}

func (h *Handler) retry(w *recorder, s scope, in body) string {
	reserved, err := h.acc.RetryWithdrawal(s.ctx, s.tc, in.ID, h.cfg.Domains.WithdrawalRetry)
	if err != nil {
		h.fail(w, failure(err))
		return ""
	}
	writeJSON(w, http.StatusOK, map[string]any{"domain": viewJSON(h.withdraw(w, s, reserved, true))})
	return reserved.Hostname
}

func (h *Handler) setSlug(w *recorder, s scope, in body) string {
	if in.Slug != "" && hostname.IsReservedLabel(in.Slug, h.cfg.Domains.Hostnames()) {
		h.fail(w, errSlugReserved)
		return ""
	}
	slug, err := h.acc.SetOrganizationSlug(s.ctx, s.tc, in.Slug)
	if err != nil {
		h.fail(w, failure(err))
		return ""
	}
	var v any
	if slug != "" {
		v = slug
	}
	writeJSON(w, http.StatusOK, map[string]any{"slug": v})
	return ""
}
