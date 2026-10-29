// Package api is the internal HTTP surface of G018 Group creation and Business Profile onboarding (Plan r5 §5.3;
// internal routes only: web is the only caller). It composes the access facade (every route calls one guarded
// operation; the raw tenancy service is unreachable), identity (the browser session and its CSRF token) and the G016
// resolver (the trusted host -> Organization decision of every route). The session-scope behaviour (begin, scope,
// originHost) is a bounded copy of core/internal/branding/api (D-5: no shared helper, no change there); http_test pins
// the copy byte for byte. Every answer is no-store JSON without CORS headers; errors carry closed codes, never names or
// identifiers of foreign objects; one onboarding.request event per request (operation, outcome, closed reason, host
// tag), never identifiers, names or payloads.
package api

import (
	"bytes"
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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/domains"
	"core/internal/identity"
	"core/internal/tenancy/access"
)

// The eight method/path pairs on six paths.
const (
	RouteListGroups    = "GET /internal/v1/groups"
	RouteSuggestSlug   = "POST /internal/v1/groups/slug-suggestion"
	RouteCreateGroup   = "POST /internal/v1/groups"
	RouteGetGroup      = "GET /internal/v1/groups/{id}"
	RouteCompleteGroup = "POST /internal/v1/groups/{id}/business-profile"
	RouteListProfiles  = "GET /internal/v1/business-profiles"
	RouteGetProfile    = "GET /internal/v1/business-profiles/{id}"
	RouteUpdateProfile = "PUT /internal/v1/business-profiles/{id}"
)

// Paths are the six served paths: every other method on them is method_not_allowed.
var Paths = []string{
	"/internal/v1/groups", "/internal/v1/groups/slug-suggestion", "/internal/v1/groups/{id}",
	"/internal/v1/groups/{id}/business-profile", "/internal/v1/business-profiles", "/internal/v1/business-profiles/{id}",
}

// The whole-payload request budgets (Plan r5 §5.3): every value the field validators accept fits in any valid JSON
// encoding (at most 12 bytes per code point, plus 1 KiB of keys, structure, mode, origin and csrf_token).
const (
	maxBody        = 8 << 10
	maxProfileBody = 96 << 10
	// defaultPage is the page size when the limit parameter is absent.
	defaultPage = 50
)

// Completion availability (availability addendum to Plan r5; AUTHORIZATION.md "Availability: denied exclusive-write
// requests"): a completion locks its target Group before the decision, so the completion route reads its body and runs
// under completionRequestDeadline (an earlier caller deadline is kept), and at most one completion per verified subject,
// and maxCompletionsInFlight in all, is admitted at a time in this process. Admission never waits: a refused request is
// the existing onboarding_unavailable 503. The limits are per core process (no shared state): with N serving core
// processes one subject may have up to N completions in flight, each bounded by the access facade's decision budget.
const (
	completionRequestDeadline = 10 * time.Second
	maxCompletionsInFlight    = 256
)

// admission is the completion admission set, keyed on the verified canonical subject as an unnamed UUID array (two
// sessions of one user share one entry). The zero value is ready.
type admission struct {
	mu     sync.Mutex
	active map[[16]byte]struct{}
}

// try admits key without waiting: false when key already holds an entry or maxCompletionsInFlight entries are held.
// The returned release removes the entry; it is idempotent.
func (a *admission) try(key [16]byte) (func(), bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, held := a.active[key]; held || len(a.active) >= maxCompletionsInFlight {
		return nil, false
	}
	if a.active == nil {
		a.active = map[[16]byte]struct{}{}
	}
	a.active[key] = struct{}{}
	var once sync.Once
	return func() {
		once.Do(func() {
			a.mu.Lock()
			delete(a.active, key)
			a.mu.Unlock()
		})
	}, true
}

// bounded runs next with the request context bounded by d (context.WithTimeout keeps an earlier caller deadline) and
// the connection's read deadline set to the same instant before the body is read, as the MCP handler does; a writer
// without deadline support is ignored there too. The context is canceled when next returns.
func bounded(d time.Duration, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		if dl, ok := ctx.Deadline(); ok {
			_ = http.NewResponseController(w).SetReadDeadline(dl)
		}
		next(w, r.WithContext(ctx))
	}
}

// Config is the handler configuration.
type Config struct {
	Pool     *pgxpool.Pool
	Identity *identity.Service
	Resolver *domains.Resolver
	Logger   *slog.Logger
}

// Handler serves the routes.
type Handler struct {
	cfg   Config
	acc   *access.Service
	log   *slog.Logger
	admit admission // the completion admission set (process-local)
}

// New returns a handler with its own access facade (like the domains and branding handlers').
func New(cfg Config) (*Handler, error) {
	if cfg.Pool == nil || cfg.Identity == nil || cfg.Resolver == nil {
		return nil, errors.New("onboarding api: incomplete configuration")
	}
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Handler{cfg: cfg, acc: access.NewService(cfg.Pool, cfg.Identity, log), log: log}, nil
}

// Register mounts every route on mux, a closed method_not_allowed answer for the other methods of each served path and
// a closed not_found answer below the two path roots.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(RouteListGroups, h.listGroups)
	mux.HandleFunc(RouteSuggestSlug, h.mutation("suggest_slug", maxBody, h.suggestSlug))
	mux.HandleFunc(RouteCreateGroup, h.mutation("create_group", maxBody, h.createGroup))
	mux.HandleFunc(RouteGetGroup, h.getGroup)
	mux.HandleFunc(RouteCompleteGroup, bounded(completionRequestDeadline, h.mutation("complete_onboarding", maxProfileBody, h.completeGroup)))
	mux.HandleFunc(RouteListProfiles, h.listProfiles)
	mux.HandleFunc(RouteGetProfile, h.getProfile)
	mux.HandleFunc(RouteUpdateProfile, h.mutation("update_profile", maxProfileBody, h.updateProfile))
	for _, p := range Paths {
		// The slug-suggestion path is also a {id} value of GET /groups/{id} (a malformed id: not_found); its other
		// methods reach the /groups/{id} method closure.
		if p != "/internal/v1/groups/slug-suggestion" {
			mux.HandleFunc(p, h.closed("method", errMethod))
		}
	}
	mux.HandleFunc("/internal/v1/groups/", h.closed("path", errNotFound))
	mux.HandleFunc("/internal/v1/business-profiles/", h.closed("path", errNotFound))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// The closed refusal codes. The first eight are the session-scope codes copied from branding/api with their values and
// statuses (http_test pins the equivalence); errUnavailable keeps branding's 503 under onboarding's own value.
const (
	errInvalidRequest  = "invalid_request"
	errAmbiguous       = "ambiguous_parameter"
	errOriginRejected  = "origin_rejected"
	errUnauthenticated = "unauthenticated"
	errCSRF            = "csrf_failed"
	errNotAvailable    = "not_available"
	errForbidden       = "forbidden"
	errNotFound        = "not_found"
	errMethod          = "method_not_allowed"
	errConflict        = "conflict"
	errUnavailable     = "onboarding_unavailable"
)

var status = map[string]int{
	errInvalidRequest: 400, errAmbiguous: 400, errOriginRejected: 403, errUnauthenticated: 401, errCSRF: 403,
	errNotAvailable: 403, errForbidden: 403, errNotFound: 404, errMethod: 405, errConflict: 409, errUnavailable: 503,
}

// recorder observes one answer for the onboarding.request event.
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

// event writes one onboarding.request record: operation, closed outcome and reason, and the origin's host tag.
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
	h.log.LogAttrs(ctx, level, "onboarding.request", attrs...)
}

// closed answers a method outside a served path's closed set, or a path below the roots that is not served.
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

// failure maps a facade error onto a closed refusal; any unavailable (possibly durable) is onboarding_unavailable.
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

// refuseErr writes the refusal of a facade error with its closed field code (invalid_request and conflict only).
func (h *Handler) refuseErr(w http.ResponseWriter, err error) {
	code := failure(err)
	var ae *access.Error
	if errors.As(err, &ae) && (code == errInvalidRequest || code == errConflict) {
		h.refuse(w, code, ae.Field)
		return
	}
	h.fail(w, code)
}

// groupJSON renders a Group with its onboarding state and the G060 display hints (Plan r0 §4.2, D-9): self_role (the
// caller's own Group role or null), can_manage and management_source ("group_role", "organization_override" or null).
// The hints authorize nothing; an override is never a self membership.
func groupJSON(g access.GroupOnboarding) map[string]any {
	out := map[string]any{"id": g.Group.ID.String(), "name": g.Group.Name, "slug": g.Group.Slug, "business_profile_id": nil,
		"complete": g.Complete, "self_role": nil, "can_manage": g.CanManage, "management_source": nil}
	if g.BusinessProfileID != nil {
		out["business_profile_id"] = g.BusinessProfileID.String()
	}
	if g.SelfRole != "" {
		out["self_role"] = string(g.SelfRole)
	}
	if g.ManagementSource != "" {
		out["management_source"] = g.ManagementSource
	}
	return out
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

// page parses the optional limit (default 50); a malformed value is invalid_request field invalid_page_size, and the
// facade decides the range.
func page(q map[string]string) (int, bool) {
	v, ok := q["limit"]
	if !ok {
		return defaultPage, true
	}
	n, err := strconv.Atoi(v)
	return n, err == nil && strconv.Itoa(n) == v
}

// read begins a GET of op: the query, then the session scope.
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

func (h *Handler) listGroups(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "list_groups", rec, &origin)
	s, q, ok := h.read(rec, r, &origin, "after", "limit", "slug")
	if !ok {
		return
	}
	defer s.end()
	limit, ok := page(q)
	if !ok {
		h.refuse(rec, errInvalidRequest, "invalid_page_size")
		return
	}
	p, err := h.acc.ListGroups(s.ctx, s.tc, q["after"], limit, q["slug"])
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	groups := make([]map[string]any, 0, len(p.Groups))
	for _, g := range p.Groups {
		groups = append(groups, groupJSON(g))
	}
	out := map[string]any{"groups": groups, "next": nil}
	if p.Next != nil {
		out["next"] = p.Next.String()
	}
	writeJSON(rec, http.StatusOK, out)
}

func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "get_group", rec, &origin)
	s, _, ok := h.read(rec, r, &origin)
	if !ok {
		return
	}
	defer s.end()
	g, err := access.ParseGroupID(r.PathValue("id"))
	if err != nil {
		h.fail(rec, errNotFound)
		return
	}
	v, err := h.acc.GetGroupOnboarding(s.ctx, s.tc, g)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, map[string]any{"group": groupJSON(v)})
}

func (h *Handler) listProfiles(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "list_profiles", rec, &origin)
	s, q, ok := h.read(rec, r, &origin, "after", "limit")
	if !ok {
		return
	}
	defer s.end()
	limit, ok := page(q)
	if !ok {
		h.refuse(rec, errInvalidRequest, "invalid_page_size")
		return
	}
	p, err := h.acc.ListBusinessProfiles(s.ctx, s.tc, q["after"], limit)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	rows := make([]map[string]any, 0, len(p.Profiles))
	for _, x := range p.Profiles {
		rows = append(rows, map[string]any{"id": x.ID.String(), "display_name": x.DisplayName, "industry": x.Industry,
			"can_select": x.CanSelect, "can_update": x.CanUpdate})
	}
	out := map[string]any{"business_profiles": rows, "next": nil}
	if p.Next != nil {
		out["next"] = p.Next.String()
	}
	writeJSON(rec, http.StatusOK, out)
}

func profileJSON(id access.BusinessProfileID, c access.BusinessContext) map[string]any {
	return map[string]any{"id": id.String(), "display_name": c.DisplayName, "legal_name": c.LegalName,
		"description": c.Description, "industry": c.Industry, "website": c.Website}
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "get_profile", rec, &origin)
	s, _, ok := h.read(rec, r, &origin)
	if !ok {
		return
	}
	defer s.end()
	p, err := access.ParseBusinessProfileID(r.PathValue("id"))
	if err != nil {
		h.fail(rec, errNotFound)
		return
	}
	v, canUpdate, err := h.acc.GetBusinessProfileView(s.ctx, s.tc, p)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, map[string]any{"business_profile": profileJSON(v.ID, v.Context), "can_update": canUpdate})
}

// profileBody is the editable context of a request; nil optional fields are absent.
type profileBody struct {
	DisplayName string  `json:"display_name"`
	LegalName   *string `json:"legal_name"`
	Description *string `json:"description"`
	Industry    *string `json:"industry"`
	Website     *string `json:"website"`
}

func (p profileBody) context() access.BusinessContext {
	return access.BusinessContext{DisplayName: p.DisplayName, LegalName: p.LegalName, Description: p.Description,
		Industry: p.Industry, Website: p.Website}
}

// body is every JSON mutation's body; each route reads its own fields and refuses the others.
type body struct {
	Origin            string       `json:"origin"`
	CSRFToken         string       `json:"csrf_token"`
	Name              *string      `json:"name"`
	Slug              *string      `json:"slug"`
	Mode              *string      `json:"mode"`
	Profile           *profileBody `json:"profile"`
	BusinessProfileID *string      `json:"business_profile_id"`
}

// mutation reads at most budget bytes (more is invalid_request before any decode, guard or write), decodes the body
// strictly, authenticates with CSRF, resolves the context and runs act; it writes one onboarding.request.
func (h *Handler) mutation(op string, budget int64, act func(w *recorder, r *http.Request, s scope, in body)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec, origin := &recorder{ResponseWriter: w}, ""
		defer h.event(r.Context(), op, rec, &origin)
		if len(r.URL.Query()) != 0 {
			h.fail(rec, errAmbiguous)
			return
		}
		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mt != "application/json" {
			h.fail(rec, errInvalidRequest)
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, budget+1))
		if err != nil || int64(len(raw)) > budget {
			h.fail(rec, errInvalidRequest)
			return
		}
		var in body
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		// Exactly one JSON value: the second Decode must report io.EOF (trailing whitespace is fine). dec.More() would
		// accept a trailing ']' or '}', since it only asks about the current array/object.
		if err := dec.Decode(&in); err != nil || !errors.Is(dec.Decode(new(json.RawMessage)), io.EOF) {
			h.fail(rec, errInvalidRequest)
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

func (h *Handler) suggestSlug(w *recorder, _ *http.Request, s scope, in body) {
	if in.Name == nil || in.Slug != nil || in.Mode != nil || in.Profile != nil || in.BusinessProfileID != nil {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.SuggestGroupSlug(s.ctx, s.tc, *in.Name)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"slug": v})
}

func (h *Handler) createGroup(w *recorder, _ *http.Request, s scope, in body) {
	if in.Name == nil || in.Mode != nil || in.Profile != nil || in.BusinessProfileID != nil {
		h.fail(w, errInvalidRequest)
		return
	}
	slug := ""
	if in.Slug != nil {
		slug = *in.Slug
	}
	g, err := h.acc.CreateGroup(s.ctx, s.tc, *in.Name, slug)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"group": groupJSON(g)})
}

func (h *Handler) completeGroup(w *recorder, r *http.Request, s scope, in body) {
	g, err := access.ParseGroupID(r.PathValue("id"))
	if err != nil {
		h.fail(w, errNotFound)
		return
	}
	if in.Mode == nil || in.Name != nil || in.Slug != nil {
		h.fail(w, errInvalidRequest)
		return
	}
	// Admission: after the authenticated begin (which already did its own database work) and before either completion's
	// transaction and Group lock. It grants nothing; the facade decides on the locked facts.
	release, ok := h.admit.try([16]byte(s.tc.Subject()))
	if !ok {
		h.fail(w, errUnavailable)
		return
	}
	defer release()
	var v access.GroupOnboarding
	switch {
	case *in.Mode == "new" && in.Profile != nil && in.BusinessProfileID == nil:
		v, err = h.acc.CompleteGroupOnboardingWithNewProfile(s.ctx, s.tc, g, in.Profile.context())
	case *in.Mode == "existing" && in.Profile == nil && in.BusinessProfileID != nil:
		p, perr := access.ParseBusinessProfileID(*in.BusinessProfileID)
		if perr != nil {
			h.refuse(w, errInvalidRequest, "invalid_id")
			return
		}
		v, err = h.acc.CompleteGroupOnboardingWithProfile(s.ctx, s.tc, g, p)
	default:
		h.fail(w, errInvalidRequest)
		return
	}
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"group": groupJSON(v)})
}

func (h *Handler) updateProfile(w *recorder, r *http.Request, s scope, in body) {
	p, err := access.ParseBusinessProfileID(r.PathValue("id"))
	if err != nil {
		h.fail(w, errNotFound)
		return
	}
	if in.Profile == nil || in.Name != nil || in.Slug != nil || in.Mode != nil || in.BusinessProfileID != nil {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.UpdateBusinessProfile(s.ctx, s.tc, p, in.Profile.context())
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"business_profile": profileJSON(v.ID, v.Context)})
}
