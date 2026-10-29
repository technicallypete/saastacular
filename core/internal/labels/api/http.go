// Package api is the internal HTTP surface of G020 generic labels (Plan r2 §4.2, D-9; internal web-BFF routes only, not
// a public API; no MCP change). It composes the access label facade (every route calls one guarded operation; the raw
// tenancy service is unreachable and tenancy is not imported), identity (the browser session and its CSRF token) and the
// G016 resolver (the trusted host -> Organization decision of every route). The session-scope behaviour and the shared
// request helpers are a bounded copy of core/internal/membership/api (no shared helper, no change there); http_test pins
// the copy byte for byte. Every answer is no-store JSON without CORS headers; errors carry closed codes, never names or
// identifiers of foreign objects; one labels.request event per request (operation, outcome, closed reason, host tag),
// never a label name, a label or target id, or a payload.
//
// Assignment routes address only the production group target kind (/internal/v1/groups/{id}/labels...): an
// Organization-owned definition assigned to a Group. No route exposes a Group-owned assignment: the production target
// registry has no Group-owned kind (Plan r2 §1.4; the registry is closed to group), so a Group-owned definition is
// listed, created, updated and deleted here but never applied.
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

	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/domains"
	"core/internal/identity"
	"core/internal/tenancy/access"
)

// The seven method/path pairs on six paths (Plan r2 §4.2), C-sorted.
const (
	RouteListGroupLabels    = "GET /internal/v1/groups/{id}/labels"
	RouteListLabels         = "GET /internal/v1/labels"
	RouteAssignGroupLabel   = "POST /internal/v1/groups/{id}/labels/assign"
	RouteUnassignGroupLabel = "POST /internal/v1/groups/{id}/labels/unassign"
	RouteCreateLabel        = "POST /internal/v1/labels"
	RouteDeleteLabel        = "POST /internal/v1/labels/{id}/delete"
	RouteUpdateLabel        = "POST /internal/v1/labels/{id}/update"
)

// Paths are the six served paths (C-sorted): every other method on them is method_not_allowed.
var Paths = []string{
	"/internal/v1/groups/{id}/labels", "/internal/v1/groups/{id}/labels/assign", "/internal/v1/groups/{id}/labels/unassign",
	"/internal/v1/labels", "/internal/v1/labels/{id}/delete", "/internal/v1/labels/{id}/update",
}

// maxBody is the whole-payload request budget of every mutation (as membership/api).
const maxBody = 8 << 10

// Config is the handler configuration.
type Config struct {
	Pool     *pgxpool.Pool
	Identity *identity.Service
	Resolver *domains.Resolver
	Logger   *slog.Logger
}

// Handler serves the routes.
type Handler struct {
	cfg Config
	acc *access.Service
	log *slog.Logger
}

// New returns a handler with its own access facade (like the membership and onboarding handlers').
func New(cfg Config) (*Handler, error) {
	if cfg.Pool == nil || cfg.Identity == nil || cfg.Resolver == nil {
		return nil, errors.New("labels api: incomplete configuration")
	}
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Handler{cfg: cfg, acc: access.NewService(cfg.Pool, cfg.Identity, log), log: log}, nil
}

// Register mounts every route on mux, a closed method_not_allowed answer for the other methods of each served path and
// a closed not_found answer below /internal/v1/labels/ (below /internal/v1/groups/ onboarding/api's root answers).
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(RouteListGroupLabels, h.listGroupLabels)
	mux.HandleFunc(RouteListLabels, h.listLabels)
	mux.HandleFunc(RouteAssignGroupLabel, h.mutation("assign_label", h.assignLabel))
	mux.HandleFunc(RouteUnassignGroupLabel, h.mutation("unassign_label", h.unassignLabel))
	mux.HandleFunc(RouteCreateLabel, h.mutation("create_label", h.createLabel))
	mux.HandleFunc(RouteDeleteLabel, h.mutation("delete_label", h.deleteLabel))
	mux.HandleFunc(RouteUpdateLabel, h.mutation("update_label", h.updateLabel))
	for _, p := range Paths {
		mux.HandleFunc(p, h.closed("method", errMethod))
	}
	mux.HandleFunc("/internal/v1/labels/", h.closed("path", errNotFound))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// The closed refusal codes: membership/api's with their values and statuses, with errUnavailable under labels' own
// value.
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
	errUnavailable     = "labels_unavailable"
)

var status = map[string]int{
	errInvalidRequest: 400, errAmbiguous: 400, errOriginRejected: 403, errUnauthenticated: 401, errCSRF: 403,
	errNotAvailable: 403, errForbidden: 403, errNotFound: 404, errMethod: 405, errConflict: 409, errUnavailable: 503,
}

// recorder observes one answer for the labels.request event.
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

// event writes one labels.request record: operation, closed outcome and reason, and the origin's host tag (never a
// name, a label or target id, or a payload).
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
	h.log.LogAttrs(ctx, level, "labels.request", attrs...)
}

// closed answers a method outside a served path's closed set, or a path below the root that is not served.
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

// failure maps a facade error onto a closed refusal; any unavailable (possibly durable, including a class-40 abort the
// facade reports as unavailable) is labels_unavailable.
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
// only: invalid_name, invalid_visibility, visibility_not_allowed, organization_context_required, name_conflict,
// scope_full, already_assigned, not_assigned, …).
func (h *Handler) refuseErr(w http.ResponseWriter, err error) {
	code := failure(err)
	var ae *access.Error
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

// page parses the optional limit (absent: 0, the facade's default page); a malformed value is invalid_request field
// invalid_page_size, and the facade decides the range.
func page(q map[string]string) (int, bool) {
	v, ok := q["limit"]
	if !ok {
		return 0, true
	}
	n, err := strconv.Atoi(v)
	return n, err == nil && strconv.Itoa(n) == v
}

// parseID parses a canonical identifier into dst with the facade's parser (ParseGroupID shares tenancy's single
// canonical UUID parser for every id type); dst is any of the facade's [16]byte id types, so the adapter never imports
// tenancy and the facade's exported surface is unchanged.
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

// setKind assigns the requested scope kind; the facade validates it (invalid_request field invalid_scope).
func setKind[T ~string](dst *T, v string) { *dst = T(v) }

// scopeOf builds the administered scope from the scope kind and the optional Group id: a malformed Group id is
// invalid_request field invalid_id; a missing or extra Group, or an unknown kind, is the facade's scope refusal.
func scopeOf(kind string, group *string) (access.Scope, bool) {
	var sc access.Scope
	setKind(&sc.Kind, kind)
	if group != nil {
		var g access.GroupID
		if !parseID(&g, *group) {
			return access.Scope{}, false
		}
		sc.Group = &g
	}
	return sc, true
}

// queryScope is scopeOf over the scope and group_id query parameters.
func queryScope(q map[string]string) (access.Scope, bool) {
	var group *string
	if g, ok := q["group_id"]; ok {
		group = &g
	}
	return scopeOf(q["scope"], group)
}

// labelJSON renders one definition: {id, scope, group_id, name, visibility} (group_id null for an Organization-owned
// definition); never the name key or created_at.
func labelJSON(v access.LabelDefinitionView) map[string]any {
	out := map[string]any{"id": formatID(v.ID), "scope": string(v.Scope), "group_id": nil, "name": v.Name,
		"visibility": string(v.Visibility)}
	if v.Group != nil {
		out["group_id"] = formatID(*v.Group)
	}
	return out
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

// listLabels answers one keyset page of the definitions of the requested scope (scope=organization, or scope=group with
// group_id); the facade decides the capability and the target_members subset.
func (h *Handler) listLabels(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "list_labels", rec, &origin)
	s, q, ok := h.read(rec, r, &origin, "scope", "group_id", "after", "limit")
	if !ok {
		return
	}
	defer s.end()
	limit, ok := page(q)
	if !ok {
		h.refuse(rec, errInvalidRequest, "invalid_page_size")
		return
	}
	sc, ok := queryScope(q)
	if !ok {
		h.refuse(rec, errInvalidRequest, "invalid_id")
		return
	}
	p, err := h.acc.ListLabelDefinitions(s.ctx, s.tc, sc, q["after"], limit)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	rows := make([]map[string]any, 0, len(p.Labels))
	for _, v := range p.Labels {
		rows = append(rows, labelJSON(v))
	}
	out := map[string]any{"labels": rows, "next": nil}
	if p.Next != nil {
		out["next"] = formatID(*p.Next)
	}
	writeJSON(rec, http.StatusOK, out)
}

// groupTarget is the production group target of a /groups/{id}/labels route: an unparsable Group id is not_found (the
// same answer as an unknown or foreign Group). The selector carries the canonical form.
func groupTarget(r *http.Request) (access.LabelTargetSelector, bool) {
	var g access.GroupID
	if !parseID(&g, r.PathValue("id")) {
		return access.LabelTargetSelector{}, false
	}
	return access.LabelTargetSelector{Group: formatID(g)}, true
}

// pathLabel is the canonical definition id of a /labels/{id}/... route: an unparsable id is not_found (the same answer
// as an unknown, foreign or unreadable definition).
func pathLabel(r *http.Request) (string, bool) {
	var v access.LabelDefinitionView
	if !parseID(&v.ID, r.PathValue("id")) {
		return "", false
	}
	return formatID(v.ID), true
}

// listGroupLabels answers the labels assigned to one Group (Organization owners/admins only; the facade decides).
func (h *Handler) listGroupLabels(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "list_group_labels", rec, &origin)
	s, _, ok := h.read(rec, r, &origin)
	if !ok {
		return
	}
	defer s.end()
	t, ok := groupTarget(r)
	if !ok {
		h.fail(rec, errNotFound)
		return
	}
	as, err := h.acc.ListLabelAssignments(s.ctx, s.tc, t)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	rows := make([]map[string]any, 0, len(as))
	for _, a := range as {
		rows = append(rows, map[string]any{"label_id": formatID(a.LabelID), "name": a.Name})
	}
	writeJSON(rec, http.StatusOK, map[string]any{"labels": rows})
}

// body is every JSON mutation's body; each route reads its own fields and refuses the others.
type body struct {
	Origin     string  `json:"origin"`
	CSRFToken  string  `json:"csrf_token"`
	Scope      *string `json:"scope"`
	GroupID    *string `json:"group_id"`
	Name       *string `json:"name"`
	Visibility *string `json:"visibility"`
	LabelID    *string `json:"label_id"`
}

// fields names the optional members present in the body.
func (in body) fields() map[string]bool {
	out := map[string]bool{}
	for name, set := range map[string]bool{"scope": in.Scope != nil, "group_id": in.GroupID != nil, "name": in.Name != nil,
		"visibility": in.Visibility != nil, "label_id": in.LabelID != nil} {
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

// mutation reads at most maxBody bytes (more is invalid_request before any decode, guard or write), decodes the body
// strictly, authenticates with CSRF, resolves the context and runs act; it writes one labels.request.
func (h *Handler) mutation(op string, act func(w *recorder, r *http.Request, s scope, in body)) http.HandlerFunc {
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
		raw, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
		if err != nil || len(raw) > maxBody {
			h.fail(rec, errInvalidRequest)
			return
		}
		var in body
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		// Exactly one JSON value: the second Decode must report io.EOF (trailing whitespace is fine).
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

// createLabel: exactly {origin, csrf_token, scope, group_id?, name, visibility}; the facade refuses target_members in
// the Organization scope (visibility_not_allowed) and decides the capability. 200 the row.
func (h *Handler) createLabel(w *recorder, _ *http.Request, s scope, in body) {
	if !in.shape([]string{"scope", "name", "visibility"}, []string{"group_id"}) {
		h.fail(w, errInvalidRequest)
		return
	}
	sc, ok := scopeOf(*in.Scope, in.GroupID)
	if !ok {
		h.refuse(w, errInvalidRequest, "invalid_id")
		return
	}
	v, err := h.acc.CreateLabelDefinition(s.ctx, s.tc, sc, *in.Name, *in.Visibility)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, labelJSON(v))
}

// updateLabel: exactly {origin, csrf_token, name, visibility} (an absolute set). 200 the row.
func (h *Handler) updateLabel(w *recorder, r *http.Request, s scope, in body) {
	id, ok := pathLabel(r)
	if !ok {
		h.fail(w, errNotFound)
		return
	}
	if !in.shape([]string{"name", "visibility"}, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.UpdateLabelDefinition(s.ctx, s.tc, id, *in.Name, *in.Visibility)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, labelJSON(v))
}

// deleteLabel: exactly {origin, csrf_token}. 200 {"deleted": true, "assignments_removed": n}.
func (h *Handler) deleteLabel(w *recorder, r *http.Request, s scope, in body) {
	id, ok := pathLabel(r)
	if !ok {
		h.fail(w, errNotFound)
		return
	}
	if !in.shape(nil, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	res, err := h.acc.DeleteLabelDefinition(s.ctx, s.tc, id)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "assignments_removed": res.AssignmentsRemoved})
}

// assignment decodes the Group path id (not_found when unparsable) and exactly {origin, csrf_token, label_id} (a
// malformed label_id is invalid_request field invalid_id) of an assign or unassign.
func (h *Handler) assignment(w *recorder, r *http.Request, in body) (string, access.LabelTargetSelector, bool) {
	t, ok := groupTarget(r)
	if !ok {
		h.fail(w, errNotFound)
		return "", access.LabelTargetSelector{}, false
	}
	if !in.shape([]string{"label_id"}, nil) {
		h.fail(w, errInvalidRequest)
		return "", access.LabelTargetSelector{}, false
	}
	var v access.LabelDefinitionView
	if !parseID(&v.ID, *in.LabelID) {
		h.refuse(w, errInvalidRequest, "invalid_id")
		return "", access.LabelTargetSelector{}, false
	}
	return formatID(v.ID), t, true
}

// assignLabel applies an Organization-owned definition to the Group. 200 {"assigned": true}.
func (h *Handler) assignLabel(w *recorder, r *http.Request, s scope, in body) {
	id, t, ok := h.assignment(w, r, in)
	if !ok {
		return
	}
	if err := h.acc.AssignLabel(s.ctx, s.tc, id, t); err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assigned": true})
}

// unassignLabel removes the assignment. 200 {"assigned": false}.
func (h *Handler) unassignLabel(w *recorder, r *http.Request, s scope, in body) {
	id, t, ok := h.assignment(w, r, in)
	if !ok {
		return
	}
	if err := h.acc.UnassignLabel(s.ctx, s.tc, id, t); err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assigned": false})
}
