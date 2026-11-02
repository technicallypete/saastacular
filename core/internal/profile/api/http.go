// Package api is the internal HTTP surface of G074 global account profiles (Plan r3.1 §3.7; internal routes only: web
// is the only caller, not a public API; no MCP change). It composes the access facade (every route calls one guarded
// profile operation; tenancy is not imported), identity (the browser session and its CSRF token) and the G016 resolver
// (the trusted host -> Organization decision of every route). The session-scope behaviour (begin, scope, originHost) is a
// bounded copy of core/internal/branding/api (itself the domains/api copy; no shared helper, no change there); http_test
// pins the copy byte for byte. Every answer except avatar bytes is no-store JSON without CORS headers; avatar bytes are
// private (max-age 300 s), versioned by the URL and the ETag, and decided now on every request (no conditional request,
// no signed URL). Errors carry closed codes, never a name, an address, a key, a size or a foreign identifier; one
// profile.request event per request (operation, closed outcome and reason), never a name, a token or a payload. G075
// (Plan r2 §3.5) adds the two Group member card routes and their optional selected-Group context selector
// (context_group), parsed here with the facade's canonical parser and resolved by the facade's unchanged Resolve; begin
// itself stays the byte-exact copy and the six G074 routes never read the selector.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/identity"
	"core/internal/objectstore"
	"core/internal/tenancy/access"
)

// The eight method/path pairs on seven paths: the six G074 pairs (Plan r3.1 §3.7) and the two G075 Group member card
// pairs (Plan r2 §3.5).
const (
	RouteOwn          = "GET /internal/v1/profiles/me"
	RouteUpdateName   = "POST /internal/v1/profiles/me"
	RouteSetAvatar    = "POST /internal/v1/profiles/me/avatar"
	RouteRemoveAvatar = "POST /internal/v1/profiles/me/avatar/remove"
	RouteGet          = "GET /internal/v1/profiles/{user_id}"
	RouteAvatar       = "GET /internal/v1/profiles/{user_id}/avatar"
	RouteGroups       = "GET /internal/v1/profiles/{user_id}/groups"
	RouteGroupLabel   = "POST /internal/v1/profiles/{user_id}/groups/{group_id}/label"
)

// fieldContextGroup is the selected-Group context selector (G075 Plan r2 C10): an optional query parameter of the cards
// read, an optional body member of the label update, and the refused field of an empty, malformed or repeated value.
const fieldContextGroup = "context_group"

const (
	maxBody = 8 << 10
	// maxAvatarBody is the avatar route's Content-Length bound (larger is 413 before any read).
	maxAvatarBody = 524288
	// avatarCache is the disclosed bounded private cache window of avatar bytes (Plan r3.1 §3.4).
	avatarCache = "private, max-age=300"
	// holdRetry is the Retry-After of a HOLD whose error carries no delay.
	holdRetry = 30 * time.Second
)

// Config is the handler configuration.
type Config struct {
	Pool        *pgxpool.Pool
	Identity    *identity.Service
	Resolver    *domains.Resolver
	ObjectStore config.ObjectStore
	AppEnv      string
	Logger      *slog.Logger
}

// Handler serves the routes.
type Handler struct {
	cfg Config
	acc *access.Service
	log *slog.Logger
}

// New returns a handler. Its access facade is its own (like the branding handler's), built with the owned object store
// from the loaded settings exactly as core/internal/branding/api builds it; a store the settings cannot build is a
// configuration error.
func New(cfg Config) (*Handler, error) {
	if cfg.Pool == nil || cfg.Identity == nil || cfg.Resolver == nil {
		return nil, errors.New("profile api: incomplete configuration")
	}
	o := cfg.ObjectStore
	store, err := objectstore.New(objectstore.Config{Endpoint: o.Endpoint, SigningEndpoint: o.SigningEndpoint, Region: o.Region,
		Bucket: o.Bucket, ForcePathStyle: o.ForcePathStyle, AccessKey: o.AccessKey, SecretKey: o.SecretKey}, cfg.AppEnv)
	if err != nil {
		return nil, errors.New("profile api: object store configuration")
	}
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	acc := access.NewServiceWithObjects(cfg.Pool, cfg.Identity, log, store, access.ObjectConfig{MaxObjectBytes: o.MaxObjectBytes,
		DefaultTTL: o.SignedURLDefaultTTL, MaxTTL: o.SignedURLMaxTTL, TransferTimeout: o.TransferTimeout})
	return &Handler{cfg: cfg, acc: acc, log: log}, nil
}

// Register mounts every route on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(RouteOwn, h.Own)
	mux.HandleFunc(RouteUpdateName, h.mutation("update_name", h.updateName))
	mux.HandleFunc(RouteSetAvatar, h.SetAvatar)
	mux.HandleFunc(RouteRemoveAvatar, h.mutation("remove_avatar", h.removeAvatar))
	mux.HandleFunc(RouteGet, h.Get)
	mux.HandleFunc(RouteAvatar, h.Avatar)
	mux.HandleFunc(RouteGroups, h.Groups)
	mux.HandleFunc(RouteGroupLabel, h.SetGroupLabel)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// The closed refusal codes. The first ten are the session-scope and host codes copied from branding/api (and so
// domains/api) with their values and statuses, and errUnavailable keeps their 503 under this package's own value; the
// rest are the profile routes' own.
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
	errConflict        = "version_conflict"
	errInvalidInput    = "invalid_input"
	errTooLarge        = "payload_too_large"
	errUnavailable     = "unavailable"
)

var status = map[string]int{
	errInvalidRequest: 400, errAmbiguous: 400, errUnsupported: 415, errOriginRejected: 403, errUnauthenticated: 401,
	errCSRF: 403, errNotAvailable: 403, errForbidden: 403, errNotFound: 404, errUnknownHost: 404, errConflict: 409,
	errInvalidInput: 400, errTooLarge: 413, errUnavailable: 503,
}

// holdReasons are the closed repairable HOLD reasons of the avatar upload (503 with Retry-After).
var holdReasons = map[string]bool{"avatar_cleanup_pending": true, "avatar_upload_in_progress": true, "commit_outcome_unknown": true}

// recorder observes one answer for the profile.request event.
type recorder struct {
	http.ResponseWriter
	code string
}

func (h *Handler) fail(w http.ResponseWriter, code string) {
	h.refuse(w, code, nil)
}

// refuse writes a closed refusal with optional closed detail fields (field, reason).
func (h *Handler) refuse(w http.ResponseWriter, code string, extra map[string]any) {
	if r, ok := w.(*recorder); ok {
		r.code = code
	}
	out := map[string]any{"error": code}
	for k, v := range extra {
		out[k] = v
	}
	writeJSON(w, status[code], out)
}

// event writes one profile.request record: operation and closed outcome and reason.
func (h *Handler) event(ctx context.Context, op string, rec *recorder) {
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
	h.log.LogAttrs(ctx, level, "profile.request", attrs...)
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
	case access.CodeConflict:
		return errConflict
	case access.CodeInvalidInput:
		switch ae.Field {
		case "avatar_too_large":
			return errTooLarge
		case "avatar_type", "avatar_dimensions", "avatar_malformed", "normalized_too_large":
			return errInvalidInput
		}
		return errInvalidRequest
	}
	return errUnavailable
}

// retryAfter is the Retry-After value of a HOLD: the facade's delay in whole seconds, rounded up (holdRetry when the error
// carries none).
func retryAfter(err error) string {
	d, ok := access.RetryAfterOf(err)
	if !ok {
		d = holdRetry
	}
	s := int(math.Ceil(d.Seconds()))
	if s < 1 {
		s = 1
	}
	return strconv.Itoa(s)
}

// refuseErr writes the refusal of a facade error with its closed detail: the invalid field, or a HOLD's reason with its
// Retry-After.
func (h *Handler) refuseErr(w http.ResponseWriter, err error) {
	code := failure(err)
	var ae *access.Error
	if !errors.As(err, &ae) || ae.Field == "" {
		h.fail(w, code)
		return
	}
	switch {
	case ae.Code == access.CodeUnavailable && holdReasons[ae.Field]:
		w.Header().Set("Retry-After", retryAfter(err))
		h.refuse(w, code, map[string]any{"reason": ae.Field})
	case code == errInvalidRequest || code == errInvalidInput || code == errTooLarge:
		h.refuse(w, code, map[string]any{"field": ae.Field})
	default:
		h.fail(w, code)
	}
}

// avatarJSON is the public avatar reference (version and content type) or null.
func avatarJSON(a *access.AccountAvatar) any {
	if a == nil {
		return nil
	}
	return map[string]any{"version": a.Version, "content_type": a.ContentType}
}

// ownJSON is the caller's own view: the explicit name (null: none), the effective name, the avatar and the optimistic
// version.
func ownJSON(v access.AccountProfileView) map[string]any {
	var name any
	if v.Name != nil {
		name = *v.Name
	}
	return map[string]any{"user_id": formatID(v.UserID), "name": name, "effective_name": v.EffectiveName, "avatar": avatarJSON(v.Avatar),
		"version": v.Version}
}

// otherJSON is another user's profile as the caller may see it now: exactly the user, the effective name and the avatar.
func otherJSON(p access.AccountProfile) map[string]any {
	return map[string]any{"user_id": formatID(p.UserID), "name": p.Name, "avatar": avatarJSON(p.Avatar)}
}

func (h *Handler) answerOwn(w *recorder, v access.AccountProfileView, err error) {
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ownJSON(v))
}

// parseID parses a canonical identifier into dst with the facade's parser (ParseGroupID shares tenancy's single canonical
// UUID parser for every id type), so the adapter never imports tenancy; only the canonical lowercase form is accepted (one
// URL per avatar version) and the zero UUID is refused.
func parseID[T ~[16]byte](dst *T, s string) bool {
	g, err := access.ParseGroupID(s)
	if err != nil || g == (access.GroupID{}) || g.String() != s {
		return false
	}
	*dst = T(g)
	return true
}

// formatID is the canonical lowercase hyphenated form of any [16]byte id (the facade's String of the same bytes).
func formatID[T ~[16]byte](id T) string { return access.GroupID(id).String() }

// parseVersion parses a canonical positive avatar version (no sign, no leading zero, at most int32).
func parseVersion(s string) (int32, bool) {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil || n < 1 || strconv.FormatInt(n, 10) != s {
		return 0, false
	}
	return int32(n), true
}

// Own is GET /internal/v1/profiles/me?origin= (session): the caller's own profile view.
func (h *Handler) Own(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "get_own", rec)
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
	v, err := h.acc.GetOwnAccountProfile(s.ctx, s.tc)
	h.answerOwn(rec, v, err)
}

// body is every JSON mutation's body; each route reads its own fields.
type body struct {
	Origin          string  `json:"origin"`
	CSRFToken       string  `json:"csrf_token"`
	ExpectedVersion *int64  `json:"expected_version"`
	Name            *string `json:"name"`
}

// mutation decodes the body, authenticates with CSRF, resolves the context and runs act; it writes one profile.request.
func (h *Handler) mutation(op string, act func(w *recorder, s scope, in body, version int32)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w}
		defer h.event(r.Context(), op, rec)
		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mt != "application/json" {
			h.fail(rec, errUnsupported)
			return
		}
		var in body
		dec := json.NewDecoder(io.LimitReader(r.Body, maxBody+1))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil || dec.More() || in.ExpectedVersion == nil || *in.ExpectedVersion < 0 ||
			*in.ExpectedVersion > math.MaxInt32 {
			h.fail(rec, errInvalidRequest)
			return
		}
		s, ok := h.begin(rec, r, in.Origin, in.CSRFToken, true)
		if !ok {
			return
		}
		defer s.end()
		act(rec, s, in, int32(*in.ExpectedVersion))
	}
}

// updateName is POST /internal/v1/profiles/me: sets the explicit name, or clears it with "" (the current provider name
// applies again). The name is required.
func (h *Handler) updateName(w *recorder, s scope, in body, version int32) {
	if in.Name == nil {
		h.refuse(w, errInvalidRequest, map[string]any{"field": "name"})
		return
	}
	v, err := h.acc.UpdateAccountProfileName(s.ctx, s.tc, version, in.Name)
	h.answerOwn(w, v, err)
}

// removeAvatar is POST /internal/v1/profiles/me/avatar/remove: withdraws the current avatar (no avatar: no-op).
func (h *Handler) removeAvatar(w *recorder, s scope, in body, version int32) {
	if in.Name != nil {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.RemoveAccountAvatar(s.ctx, s.tc, version)
	h.answerOwn(w, v, err)
}

// SetAvatar is POST /internal/v1/profiles/me/avatar (session, raw body): headers X-Origin, X-CSRF-Token and
// X-Expected-Version, Content-Length at most 512 KiB (larger is 413 before any read; an unknown length is
// invalid_request). The claimed content type is ignored; the facade detects, validates and normalizes the image. A HOLD
// is 503 unavailable with its reason and Retry-After.
func (h *Handler) SetAvatar(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "set_avatar", rec)
	origin, csrf, ver := r.Header.Values("X-Origin"), r.Header.Values("X-CSRF-Token"), r.Header.Values("X-Expected-Version")
	if len(origin) != 1 || len(csrf) != 1 || len(ver) != 1 {
		h.fail(rec, errAmbiguous)
		return
	}
	version, err := strconv.ParseInt(ver[0], 10, 32)
	switch {
	case err != nil || version < 0:
		h.fail(rec, errInvalidRequest)
		return
	case r.ContentLength > maxAvatarBody:
		h.fail(rec, errTooLarge)
		return
	case r.ContentLength < 0:
		h.fail(rec, errInvalidRequest)
		return
	}
	s, ok := h.begin(rec, r, origin[0], csrf[0], true)
	if !ok {
		return
	}
	defer s.end()
	v, err := h.acc.SetAccountAvatar(s.ctx, s.tc, int32(version), r.Body, r.ContentLength)
	h.answerOwn(rec, v, err)
}

// Get is GET /internal/v1/profiles/{user_id}?origin= (session): another user's (or the caller's own) profile as the
// caller may see it now. An unknown, foreign, unrelated or no longer related subject, and a malformed id, are the same
// 404 not_found.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "get", rec)
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
	var p access.AccountProfile
	if !parseID(&p.UserID, r.PathValue("user_id")) {
		h.fail(rec, errNotFound)
		return
	}
	p, err := h.acc.GetAccountProfile(s.ctx, s.tc, p.UserID)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, otherJSON(p))
}

// Avatar is GET /internal/v1/profiles/{user_id}/avatar?origin=&v= (session): the subject's current avatar bytes when v
// is its current avatar version, decided now on every request. Anything else (an unrelated subject, a stale or
// malformed v, no avatar) is a no-store 404; a store failure a no-store 503.
func (h *Handler) Avatar(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "get_avatar", rec)
	q, vs := r.URL.Query()["origin"], r.URL.Query()["v"]
	if len(q) != 1 || len(vs) != 1 {
		h.fail(rec, errAmbiguous)
		return
	}
	s, ok := h.begin(rec, r, q[0], "", false)
	if !ok {
		return
	}
	defer s.end()
	var p access.AccountProfile
	version, vok := parseVersion(vs[0])
	if !parseID(&p.UserID, r.PathValue("user_id")) || !vok {
		h.fail(rec, errNotFound)
		return
	}
	a, data, err := h.acc.GetAccountAvatar(s.ctx, s.tc, p.UserID, version)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeAvatar(rec, a, data)
}

// writeAvatar writes avatar bytes with the reviewed header set only.
func writeAvatar(w http.ResponseWriter, a access.AccountAvatar, data []byte) {
	hd := w.Header()
	hd.Set("Content-Type", a.ContentType)
	hd.Set("Content-Length", strconv.Itoa(len(data)))
	hd.Set("Cache-Control", avatarCache)
	hd.Set("ETag", `"v`+strconv.FormatInt(int64(a.Version), 10)+`"`)
	hd.Set("X-Content-Type-Options", "nosniff")
	hd.Set("Content-Disposition", "inline")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// G075 Group member cards (Plan r2 §3.4, §3.5, §3.9; C8, C10). The two routes read the subject's cards and update one
// role label through the guarded facade operations ListGroupMemberCards and SetGroupMemberLabel. Each accepts one
// optional selected-Group context selector, context_group, which is checked with the facade's canonical parser
// (parseID; this package never imports tenancy) before the session scope and resolved by the facade's unchanged Resolve
// in the same request scope (narrow): Resolve decides whether the Group is selectable, and the selector only narrows.
// An absent selector is the explicit Organization-only request; a present one that is empty, malformed or repeated is
// invalid_request context_group, never read as absent (which would broaden the request to Organization-only). An
// unselectable selector is the same 404 not_found as an invisible subject, and the selector is never echoed or recorded.
// begin stays the byte-exact branding/api copy (Organization-only); the six G074 routes never read the selector.

// contextSelector reads the optional selector from its values: absent is "" (Organization-only); present, it must be
// exactly one canonical lowercase Group UUID (the zero UUID refused), returned byte-exact.
func contextSelector(vals []string, present bool) (string, bool) {
	if !present {
		return "", true
	}
	var g access.GroupID
	if len(vals) != 1 || !parseID(&g, vals[0]) {
		return "", false
	}
	return vals[0], true
}

// groupsQuery parses the cards read's query strictly: a query that does not parse, or that names a parameter other than
// origin and context_group, is refused (url.Values silently drops a malformed pair, which would turn a present selector
// into an absent one).
func groupsQuery(raw string) (url.Values, bool) {
	q, err := url.ParseQuery(raw)
	if err != nil {
		return nil, false
	}
	for k := range q {
		if k != "origin" && k != fieldContextGroup {
			return nil, false
		}
	}
	return q, true
}

// bodySelector reads the optional selector of a JSON body: an absent member is "" (Organization-only); a present member
// must be one string contextSelector accepts (null, a number or an empty string is refused, never read as absent).
func bodySelector(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", true
	}
	var s *string
	if err := json.Unmarshal(raw, &s); err != nil || s == nil {
		return "", false
	}
	return contextSelector([]string{*s}, true)
}

// repeatedMember returns the folded name of the first repeated top-level member of one JSON object ("" when every member
// is distinct). encoding/json matches member names without regard to case and silently keeps the last of a repeated
// member, so a repeated member (the selector above all) is refused here instead. ok is false for anything but one
// well-formed object.
func repeatedMember(data []byte) (string, bool) {
	dec := json.NewDecoder(bytes.NewReader(data))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return "", false
	}
	seen := map[string]bool{}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return "", false
		}
		k, ok := t.(string)
		if !ok {
			return "", false
		}
		k = strings.ToLower(strings.ToUpper(k))
		if seen[k] {
			return k, true
		}
		seen[k] = true
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return "", false
		}
	}
	return "", true
}

// contextResolver is the facade's Resolve (an interface only so that the selector transport is testable without a
// database).
type contextResolver interface {
	Resolve(ctx context.Context, req access.ResolveRequest) (*access.TenantContext, error)
}

// narrow resolves the request's context with the validated selector in the same request scope: absent ("") keeps the
// Organization-only context begin resolved; present, the selector is passed byte-exact as ResolveRequest.GroupID with
// the scope's Organization, and Resolve decides whether that Group is selectable (Resolve records its policy refusals as
// it always does). Its context_denied (an unknown, foreign or unselectable Group) is not_found, the identity of an
// invisible subject; an invalid selector is invalid_request; anything else keeps its usual refusal.
func narrow(ctx context.Context, res contextResolver, tc *access.TenantContext, org, group string) (*access.TenantContext, string) {
	if group == "" {
		return tc, ""
	}
	sel, err := res.Resolve(ctx, access.ResolveRequest{OrganizationID: org, GroupID: group})
	switch {
	case err == nil:
		return sel, ""
	case access.CodeOf(err) == access.CodeContextDenied:
		return nil, errNotFound
	}
	return nil, failure(err)
}

// failSelector writes a selector refusal: invalid_request names the selector field (never its value); not_found is
// exactly the body of an invisible subject.
func (h *Handler) failSelector(w http.ResponseWriter, code string) {
	if code == errInvalidRequest {
		h.refuse(w, code, map[string]any{"field": fieldContextGroup})
		return
	}
	h.fail(w, code)
}

// cardJSON is one Group member card: exactly the Group, its name, the role label (null: none), the edit hint and the
// label's optimistic version (0: no label yet); never a permission role, another member, a count or a Business Profile.
func cardJSON(c access.GroupMemberCard) map[string]any {
	var label any
	if c.RoleLabel != nil {
		label = *c.RoleLabel
	}
	return map[string]any{"group_id": formatID(c.GroupID), "group_name": c.GroupName, "role_label": label, "can_edit": c.CanEdit,
		"version": c.Version}
}

// cardsJSON is the subject's cards as the caller may see them now: the complete authorized list or, on an authorized
// overflow (C8), no card and narrow_required; never a partial list, the overflowing card, a total or a cursor.
func cardsJSON(c access.GroupMemberCards) map[string]any {
	cards := []map[string]any{}
	if !c.NarrowRequired {
		for _, card := range c.Cards {
			cards = append(cards, cardJSON(card))
		}
	}
	return map[string]any{"user_id": formatID(c.Subject), "cards": cards, "narrow_required": c.NarrowRequired}
}

// Groups is GET /internal/v1/profiles/{user_id}/groups?origin=&context_group= (session): the subject's Group member
// cards as the caller may see them now, restricted to the selected Group when context_group is present. A subject the
// relationship predicate refuses, a malformed id and an unselectable selector are the same 404 not_found; a visible
// subject without a card is an empty list.
func (h *Handler) Groups(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "get_groups", rec)
	q, ok := groupsQuery(r.URL.RawQuery)
	if !ok {
		h.fail(rec, errInvalidRequest)
		return
	}
	origin := q["origin"]
	if len(origin) != 1 {
		h.fail(rec, errAmbiguous)
		return
	}
	vals, present := q[fieldContextGroup]
	group, ok := contextSelector(vals, present)
	if !ok {
		h.failSelector(rec, errInvalidRequest)
		return
	}
	s, ok := h.begin(rec, r, origin[0], "", false)
	if !ok {
		return
	}
	defer s.end()
	var c access.GroupMemberCards
	if !parseID(&c.Subject, r.PathValue("user_id")) {
		h.fail(rec, errNotFound)
		return
	}
	tc, code := narrow(s.ctx, h.acc, s.tc, formatID(s.tc.Organization()), group)
	if code != "" {
		h.failSelector(rec, code)
		return
	}
	c, err := h.acc.ListGroupMemberCards(s.ctx, tc, c.Subject)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, cardsJSON(c))
}

// labelBody is the label update's body: its own type, so the G074 mutation body neither reads nor accepts its members.
type labelBody struct {
	Origin          string          `json:"origin"`
	CSRFToken       string          `json:"csrf_token"`
	ExpectedVersion *int64          `json:"expected_version"`
	RoleLabel       *string         `json:"role_label"`
	ContextGroup    json.RawMessage `json:"context_group"`
}

// SetGroupLabel is POST /internal/v1/profiles/{user_id}/groups/{group_id}/label (session, CSRF; JSON body origin,
// csrf_token, expected_version, role_label and the optional context_group): sets the subject's role label in the target
// Group {group_id} ("" clears it) when the label is still at expected_version (0: no label yet) and answers the
// subject's card for that Group. The target Group and the selector are parsed separately and never substituted for one
// another; a target other than the selected Group is the facade's not_found. A malformed id, an unknown or foreign
// target, a subject without a current membership in it and an unselectable selector are the same 404 not_found; a
// readable Group with a denied capability is 403 forbidden; a stale version 409 version_conflict.
func (h *Handler) SetGroupLabel(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "update_group_label", rec)
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "application/json" {
		h.fail(rec, errUnsupported)
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil || len(data) > maxBody {
		h.fail(rec, errInvalidRequest)
		return
	}
	var in labelBody
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil || dec.More() || in.ExpectedVersion == nil || *in.ExpectedVersion < 0 ||
		*in.ExpectedVersion > math.MaxInt32 {
		h.fail(rec, errInvalidRequest)
		return
	}
	repeated, ok := repeatedMember(data)
	switch {
	case repeated == fieldContextGroup:
		h.failSelector(rec, errInvalidRequest)
		return
	case !ok || repeated != "":
		h.fail(rec, errInvalidRequest)
		return
	}
	group, ok := bodySelector(in.ContextGroup)
	if !ok {
		h.failSelector(rec, errInvalidRequest)
		return
	}
	s, ok := h.begin(rec, r, in.Origin, in.CSRFToken, true)
	if !ok {
		return
	}
	defer s.end()
	var p access.AccountProfile // the subject's id type, without naming tenancy
	var g access.GroupID
	if !parseID(&p.UserID, r.PathValue("user_id")) || !parseID(&g, r.PathValue("group_id")) {
		h.fail(rec, errNotFound)
		return
	}
	if in.RoleLabel == nil {
		h.refuse(rec, errInvalidRequest, map[string]any{"field": "role_label"})
		return
	}
	tc, code := narrow(s.ctx, h.acc, s.tc, formatID(s.tc.Organization()), group)
	if code != "" {
		h.failSelector(rec, code)
		return
	}
	card, err := h.acc.SetGroupMemberLabel(s.ctx, tc, g, p.UserID, int32(*in.ExpectedVersion), in.RoleLabel)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, cardJSON(card))
}
