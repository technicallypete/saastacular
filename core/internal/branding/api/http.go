// Package api is the internal HTTP surface of G017 Organization branding (Plan r2 §5.3, §5.5; internal routes only:
// web is the only caller). It composes the access facade (authorization, the guarded branding operations, the logo
// upload over the owned object store and the public projection), identity (the browser session and its CSRF token)
// and the G016 resolver (the trusted host -> Organization decision of every route). The session-scope behaviour
// (begin, scope, originHost) is a bounded copy of core/internal/domains/api (D-5: no shared helper, no change there);
// http_test pins the copy byte-for-byte. Every answer except a current logo is no-store JSON without CORS headers;
// hostnames appear in logs only as a 12-hex-character hash prefix.
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
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/config"
	"core/internal/domains"
	"core/internal/identity"
	"core/internal/objectstore"
	"core/internal/tenancy/access"
)

// The routes.
const (
	RoutePublic       = "GET /internal/v1/branding/public"
	RouteLogo         = "GET /internal/v1/branding/logo"
	RouteGet          = "GET /internal/v1/branding"
	RouteUpdate       = "POST /internal/v1/branding"
	RouteReset        = "POST /internal/v1/branding/reset"
	RouteSetLogo      = "POST /internal/v1/branding/logo"
	RouteRemoveLogo   = "POST /internal/v1/branding/logo/remove"
	RouteDeleteObject = "POST /internal/v1/branding/objects/delete"
)

const (
	maxBody = 8 << 10
	// maxLogoBody is the logo route's Content-Length bound (larger is 413 before any read).
	maxLogoBody = 524288
	// logoCache is the disclosed bounded cached-byte contract of a current logo (Plan r2 §5.5, C-4).
	logoCache = "public, max-age=86400, immutable"
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

// New returns a handler. Its access facade is its own (like the domains and MCP handlers'), built with the owned object
// store from the loaded settings; a store the settings cannot build is a configuration error.
func New(cfg Config) (*Handler, error) {
	if cfg.Pool == nil || cfg.Identity == nil || cfg.Resolver == nil {
		return nil, errors.New("branding api: incomplete configuration")
	}
	o := cfg.ObjectStore
	store, err := objectstore.New(objectstore.Config{Endpoint: o.Endpoint, SigningEndpoint: o.SigningEndpoint, Region: o.Region,
		Bucket: o.Bucket, ForcePathStyle: o.ForcePathStyle, AccessKey: o.AccessKey, SecretKey: o.SecretKey}, cfg.AppEnv)
	if err != nil {
		return nil, errors.New("branding api: object store configuration")
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
	mux.HandleFunc(RoutePublic, h.Public)
	mux.HandleFunc(RouteLogo, h.Logo)
	mux.HandleFunc(RouteGet, h.Get)
	mux.HandleFunc(RouteUpdate, h.mutation("update", h.update))
	mux.HandleFunc(RouteReset, h.mutation("reset", h.reset))
	mux.HandleFunc(RouteSetLogo, h.SetLogo)
	mux.HandleFunc(RouteRemoveLogo, h.mutation("remove_logo", h.removeLogo))
	mux.HandleFunc(RouteDeleteObject, h.mutation("delete_object", h.deleteObject))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// The closed refusal codes. The first ten are the session-scope and host codes copied from domains/api with their
// values and statuses, and errUnavailable keeps domains' 503 under branding's own value (eleven shared identifiers;
// http_test pins the equivalence); the rest are branding's own.
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
	errContrast        = "contrast_failed"
	errLogoInvalid     = "logo_invalid"
	errTooLarge        = "payload_too_large"
	errOutcomeUnknown  = "publication_outcome_unknown"
	errUnavailable     = "branding_unavailable"
)

var status = map[string]int{
	errInvalidRequest: 400, errAmbiguous: 400, errUnsupported: 415, errOriginRejected: 403, errUnauthenticated: 401,
	errCSRF: 403, errNotAvailable: 403, errForbidden: 403, errNotFound: 404, errUnknownHost: 404, errConflict: 409,
	errContrast: 400, errLogoInvalid: 400, errTooLarge: 413, errOutcomeUnknown: 503, errUnavailable: 503,
}

// recorder observes one answer for the branding.request / branding.public event.
type recorder struct {
	http.ResponseWriter
	code string
}

func (h *Handler) fail(w http.ResponseWriter, code string) {
	h.refuse(w, code, nil)
}

// refuse writes a closed refusal with optional closed detail fields (field, upload outcome).
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

// hostTag is the privacy-safe log form of a hostname: the first 12 hex characters of its SHA-256 (as domains/api).
func hostTag(name string) string {
	sum := sha256.Sum256([]byte("domain/1\x00" + name))
	return hex.EncodeToString(sum[:6])
}

// event writes one record: operation, closed outcome and reason, and the hostname tag when known.
func (h *Handler) event(ctx context.Context, name, op string, rec *recorder, host string) {
	attrs := []slog.Attr{slog.String("operation", op)}
	level := slog.LevelInfo
	switch {
	case rec.code == "":
		attrs = append(attrs, slog.String("outcome", "ok"))
	case rec.code == errUnavailable || rec.code == errOutcomeUnknown:
		level = slog.LevelError
		attrs = append(attrs, slog.String("outcome", "failed"), slog.String("reason", rec.code))
	default:
		level = slog.LevelWarn
		attrs = append(attrs, slog.String("outcome", "refused"), slog.String("reason", rec.code))
	}
	if host != "" {
		attrs = append(attrs, slog.String("host_tag", hostTag(host)))
	}
	h.log.LogAttrs(ctx, level, name, attrs...)
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

// publicOrg resolves an anonymous request's single origin parameter to the Organization its host routes to now.
func (h *Handler) publicOrg(w *recorder, r *http.Request) (org, host string, ok bool) {
	q := r.URL.Query()["origin"]
	if len(q) != 1 {
		h.fail(w, errAmbiguous)
		return "", "", false
	}
	host, ok = originHost(q[0])
	if !ok {
		h.fail(w, errUnknownHost)
		return "", "", false
	}
	org, _, err := h.cfg.Resolver.Lookup(r.Context(), host)
	switch {
	case errors.Is(err, domains.ErrUnknownHost):
		h.fail(w, errUnknownHost)
		return "", host, false
	case err != nil:
		h.fail(w, errUnavailable)
		return "", host, false
	}
	return org, host, true
}

// Public is GET /internal/v1/branding/public?origin= (no credential): the public projection of the Organization the
// origin's host routes to. 200 {revision, display_label?, light, dark, logo?: {v, content_type}}; 404 unknown_host.
func (h *Handler) Public(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	host := ""
	defer func() { h.event(r.Context(), "branding.public", "projection", rec, host) }()
	org, hst, ok := h.publicOrg(rec, r)
	host = hst
	if !ok {
		return
	}
	v, err := h.acc.PublicBranding(r.Context(), org)
	if err != nil {
		h.fail(rec, failure(err))
		return
	}
	out := map[string]any{"revision": v.Revision, "light": v.Light, "dark": v.Dark}
	if v.Label != nil {
		out["display_label"] = *v.Label
	}
	if v.Logo != nil {
		out["logo"] = map[string]any{"v": v.Logo.V, "content_type": v.Logo.ContentType}
	}
	writeJSON(rec, http.StatusOK, out)
}

// Logo is GET /internal/v1/branding/logo?origin=&v= (no credential): the current ready logo's bytes of the
// Organization the host routes to, only when v is its current revision; cacheable for 24 h (the disclosed bounded
// contract). A stale, foreign or missing v, or a host that routes nowhere, is a no-store 404.
func (h *Handler) Logo(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	host := ""
	defer func() { h.event(r.Context(), "branding.public", "logo", rec, host) }()
	vs := r.URL.Query()["v"]
	if len(vs) != 1 {
		h.fail(rec, errAmbiguous)
		return
	}
	org, hst, ok := h.publicOrg(rec, r)
	host = hst
	if !ok {
		return
	}
	ct, data, err := h.acc.PublicBrandingLogo(r.Context(), org, vs[0])
	if err != nil {
		h.fail(rec, failure(err))
		return
	}
	hd := rec.Header()
	hd.Set("Content-Type", ct)
	hd.Set("Content-Length", strconv.Itoa(len(data)))
	hd.Set("Cache-Control", logoCache)
	hd.Set("ETag", `"`+vs[0]+`"`)
	hd.Set("X-Content-Type-Options", "nosniff")
	hd.Set("Content-Disposition", "inline")
	rec.WriteHeader(http.StatusOK)
	_, _ = rec.Write(data)
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
	case access.CodeContrastFailed:
		return errContrast
	case access.CodeInvalidInput:
		switch ae.Field {
		case "logo_too_large":
			return errTooLarge
		case "logo_type", "logo_dimensions", "logo_malformed", "normalized_too_large":
			return errLogoInvalid
		}
		return errInvalidRequest
	case access.CodeUnavailable:
		if ae.Field == "publication_outcome_unknown" {
			return errOutcomeUnknown
		}
	}
	return errUnavailable
}

// refuseErr writes the refusal of a facade error with its closed detail: the invalid field, or the logo upload's
// outcome (the object and its actual live state) after the upload.
func (h *Handler) refuseErr(w http.ResponseWriter, err error) {
	code := failure(err)
	var ae *access.Error
	if !errors.As(err, &ae) || ae.Field == "" {
		h.fail(w, code)
		return
	}
	if ae.Object != "" {
		h.refuse(w, code, map[string]any{"upload": map[string]any{"outcome": ae.Field, "object_id": ae.Object, "object_state": ae.ObjectState}})
		return
	}
	switch code {
	case errInvalidRequest, errContrast, errLogoInvalid, errTooLarge:
		h.refuse(w, code, map[string]any{"field": ae.Field})
	default:
		h.fail(w, code)
	}
}

func logoJSON(l access.BrandingLogo) map[string]any {
	out := map[string]any{"object_id": l.ObjectID, "content_type": l.ContentType, "byte_size": l.ByteSize, "state": l.State}
	if l.SHA12 != "" {
		out["v"] = l.SHA12
	}
	return out
}

func viewJSON(v access.BrandingView) map[string]any {
	out := map[string]any{"version": v.Version, "display_label": v.Label, "light": v.Light, "dark": v.Dark, "logo": nil}
	if v.Logo != nil {
		out["logo"] = logoJSON(*v.Logo)
	}
	if v.Uploads != nil {
		ups := make([]map[string]any, 0, len(v.Uploads))
		for _, u := range v.Uploads {
			ups = append(ups, logoJSON(u))
		}
		out["uploads"], out["uploads_total"] = ups, v.UploadsTotal
	}
	return out
}

// Get is GET /internal/v1/branding?origin= (session): the active Organization's branding view (owner or admin).
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "branding.request", "get", rec, "")
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
	v, err := h.acc.GetBranding(s.ctx, s.tc)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, map[string]any{"branding": viewJSON(v)})
}

// body is every JSON mutation's body; each route reads its own fields.
type body struct {
	Origin          string            `json:"origin"`
	CSRFToken       string            `json:"csrf_token"`
	ExpectedVersion *int64            `json:"expected_version"`
	DisplayLabel    *string           `json:"display_label"`
	Light           map[string]string `json:"light"`
	Dark            map[string]string `json:"dark"`
	ObjectID        string            `json:"object_id"`
}

// mutation decodes the body, authenticates with CSRF, resolves the context and runs act; it writes one
// branding.request.
func (h *Handler) mutation(op string, act func(w *recorder, s scope, in body, version int64)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w}
		defer h.event(r.Context(), "branding.request", op, rec, "")
		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mt != "application/json" {
			h.fail(rec, errUnsupported)
			return
		}
		var in body
		dec := json.NewDecoder(io.LimitReader(r.Body, maxBody+1))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil || dec.More() || in.ExpectedVersion == nil || *in.ExpectedVersion < 0 {
			h.fail(rec, errInvalidRequest)
			return
		}
		s, ok := h.begin(rec, r, in.Origin, in.CSRFToken, true)
		if !ok {
			return
		}
		defer s.end()
		act(rec, s, in, *in.ExpectedVersion)
	}
}

func (h *Handler) answer(w *recorder, v access.BrandingView, err error) {
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"branding": viewJSON(v)})
}

func (h *Handler) update(w *recorder, s scope, in body, version int64) {
	v, err := h.acc.UpdateBranding(s.ctx, s.tc, access.BrandingUpdate{ExpectedVersion: version, Label: in.DisplayLabel,
		Light: in.Light, Dark: in.Dark})
	h.answer(w, v, err)
}

func (h *Handler) reset(w *recorder, s scope, _ body, version int64) {
	v, err := h.acc.ResetBranding(s.ctx, s.tc, version)
	h.answer(w, v, err)
}

func (h *Handler) removeLogo(w *recorder, s scope, _ body, version int64) {
	v, err := h.acc.RemoveBrandingLogo(s.ctx, s.tc, version)
	h.answer(w, v, err)
}

// deleteObject deletes one unreferenced logo upload through the facade's DeleteBrandingUpload (D-3): only an object
// the current view lists as an upload, and only while the view is still at the caller's version. The truthful outcome
// is deleted or deletion_pending.
func (h *Handler) deleteObject(w *recorder, s scope, in body, version int64) {
	res, err := h.acc.DeleteBrandingUpload(s.ctx, s.tc, version, in.ObjectID)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"outcome": string(res.Outcome)})
}

// SetLogo is POST /internal/v1/branding/logo?origin= (session, raw body): headers X-CSRF-Token and X-Expected-Version,
// Content-Length at most 512 KiB (larger is 413 before any read; an unknown length is invalid_request). The claimed
// content type is ignored; the server detects, validates and normalizes the image.
func (h *Handler) SetLogo(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer h.event(r.Context(), "branding.request", "set_logo", rec, "")
	q := r.URL.Query()["origin"]
	csrf, ver := r.Header.Values("X-CSRF-Token"), r.Header.Values("X-Expected-Version")
	if len(q) != 1 || len(csrf) != 1 || len(ver) != 1 {
		h.fail(rec, errAmbiguous)
		return
	}
	version, err := strconv.ParseInt(ver[0], 10, 64)
	switch {
	case err != nil || version < 0:
		h.fail(rec, errInvalidRequest)
		return
	case r.ContentLength > maxLogoBody:
		h.fail(rec, errTooLarge)
		return
	case r.ContentLength < 0:
		h.fail(rec, errInvalidRequest)
		return
	}
	s, ok := h.begin(rec, r, q[0], csrf[0], true)
	if !ok {
		return
	}
	defer s.end()
	v, err := h.acc.SetBrandingLogo(s.ctx, s.tc, version, r.Body, r.ContentLength)
	h.answer(rec, v, err)
}
