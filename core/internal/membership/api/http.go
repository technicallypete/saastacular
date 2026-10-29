// Package api is the internal HTTP surface of G019 scoped invitations, membership and ownership administration (Plan
// r1 §5.3 as amended by the r2 addendum §7; internal web-BFF routes only, not a public API; no MCP change). It composes
// the access facade (every route calls one guarded operation; the raw tenancy service is unreachable and tenancy is not
// imported), identity (the browser session and its CSRF token) and the G016 resolver (the trusted host -> Organization
// decision of every tenant route). The session-scope behaviour (begin, scope, originHost) is a bounded copy of
// core/internal/onboarding/api (no shared helper, no change there); http_test pins the copy byte for byte. Every answer
// is no-store JSON without CORS headers; errors carry closed codes, never names, addresses, tokens or identifiers of
// foreign objects; one membership.request event per request (operation, outcome, closed reason, host tag), never a
// token, an address or a payload. There is no Organization-invitation consumption route (G058) and no retire route.
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
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/domains"
	"core/internal/identity"
	"core/internal/tenancy/access"
)

// The eighteen method/path pairs on sixteen paths (r2 addendum §7; G060 adds the Group-authority read and owner-only
// write on /internal/v1/organization/group-authority).
const (
	RouteListMembers           = "GET /internal/v1/members"
	RouteSelfMember            = "GET /internal/v1/members/self"
	RouteChangeRole            = "POST /internal/v1/members/{user_id}/role"
	RouteRemoveMember          = "POST /internal/v1/members/{user_id}/remove"
	RoutePromoteOwner          = "POST /internal/v1/members/{user_id}/promote"
	RouteDemoteOwner           = "POST /internal/v1/members/{user_id}/demote"
	RouteLeaveScope            = "POST /internal/v1/members/leave"
	RouteTransferOwnership     = "POST /internal/v1/ownership/transfer"
	RouteListInvitations       = "GET /internal/v1/invitations"
	RouteIssueInvitation       = "POST /internal/v1/invitations"
	RouteMyInvitations         = "GET /internal/v1/invitations/mine"
	RouteInvitationCandidates  = "GET /internal/v1/invitations/candidates"
	RouteGetInvitation         = "GET /internal/v1/invitations/{id}"
	RouteRevokeInvitation      = "POST /internal/v1/invitations/{id}/revoke"
	RouteAcceptInvitation      = "POST /internal/v1/invitations/{id}/accept"
	RouteRetirementEligibility = "GET /internal/v1/account/retirement-eligibility"
	RouteGetGroupAuthority     = "GET /internal/v1/organization/group-authority"
	RouteSetGroupAuthority     = "POST /internal/v1/organization/group-authority"
)

// Paths are the sixteen served paths: every other method on them is method_not_allowed.
var Paths = []string{
	"/internal/v1/members", "/internal/v1/members/self", "/internal/v1/members/{user_id}/role",
	"/internal/v1/members/{user_id}/remove", "/internal/v1/members/{user_id}/promote",
	"/internal/v1/members/{user_id}/demote", "/internal/v1/members/leave", "/internal/v1/ownership/transfer",
	"/internal/v1/invitations", "/internal/v1/invitations/mine", "/internal/v1/invitations/candidates",
	"/internal/v1/invitations/{id}", "/internal/v1/invitations/{id}/revoke", "/internal/v1/invitations/{id}/accept",
	"/internal/v1/account/retirement-eligibility", "/internal/v1/organization/group-authority",
}

// literalOfID are the literal segments matched before the /invitations/{id} template: their other methods reach that
// template's method closure (a separate method-less pattern would conflict with GET /invitations/{id}), and a request
// for them never parses the literal as an id.
var literalOfID = map[string]bool{"/internal/v1/invitations/mine": true, "/internal/v1/invitations/candidates": true}

// maxBody is the whole-payload request budget of every mutation (Plan r1 §5.3).
const maxBody = 8 << 10

// maxLifetimeHours bounds expires_in_hours before the duration is formed (no overflow); the facade decides the range.
const maxLifetimeHours = 1 << 20

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

// New returns a handler with its own access facade (like the domains, branding and onboarding handlers').
func New(cfg Config) (*Handler, error) {
	if cfg.Pool == nil || cfg.Identity == nil || cfg.Resolver == nil {
		return nil, errors.New("membership api: incomplete configuration")
	}
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Handler{cfg: cfg, acc: access.NewService(cfg.Pool, cfg.Identity, log), log: log}, nil
}

// Register mounts every route on mux, a closed method_not_allowed answer for the other methods of each served path and
// a closed not_found answer below the four path roots.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(RouteListMembers, h.listMembers)
	mux.HandleFunc(RouteSelfMember, h.selfMember)
	mux.HandleFunc(RouteChangeRole, h.mutation("change_role", h.changeRole))
	mux.HandleFunc(RouteRemoveMember, h.mutation("remove_member", h.removeMember))
	mux.HandleFunc(RoutePromoteOwner, h.mutation("promote_owner", h.promoteOwner))
	mux.HandleFunc(RouteDemoteOwner, h.mutation("demote_owner", h.demoteOwner))
	mux.HandleFunc(RouteLeaveScope, h.mutation("leave_scope", h.leaveScope))
	mux.HandleFunc(RouteTransferOwnership, h.mutation("transfer_ownership", h.transferOwnership))
	mux.HandleFunc(RouteListInvitations, h.listInvitations)
	mux.HandleFunc(RouteIssueInvitation, h.mutation("issue_invitation", h.issueInvitation))
	mux.HandleFunc(RouteMyInvitations, h.myInvitations)
	mux.HandleFunc(RouteInvitationCandidates, h.invitationCandidates)
	mux.HandleFunc(RouteGetInvitation, h.getInvitation)
	mux.HandleFunc(RouteRevokeInvitation, h.mutation("revoke_invitation", h.revokeInvitation))
	mux.HandleFunc(RouteAcceptInvitation, h.mutation("accept_invitation", h.acceptInvitation))
	mux.HandleFunc(RouteRetirementEligibility, h.retirementEligibility)
	mux.HandleFunc(RouteGetGroupAuthority, h.groupAuthority)
	mux.HandleFunc(RouteSetGroupAuthority, h.mutation("set_group_authority", h.setGroupAuthority))
	for _, p := range Paths {
		if !literalOfID[p] {
			mux.HandleFunc(p, h.closed("method", errMethod))
		}
	}
	for _, root := range []string{"/internal/v1/members/", "/internal/v1/invitations/", "/internal/v1/ownership/",
		"/internal/v1/account/"} {
		mux.HandleFunc(root, h.closed("path", errNotFound))
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// The closed refusal codes: onboarding/api's with their values and statuses (http_test pins the equivalence), with
// errUnavailable under membership's own value.
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
	errUnavailable     = "membership_unavailable"
)

var status = map[string]int{
	errInvalidRequest: 400, errAmbiguous: 400, errOriginRejected: 403, errUnauthenticated: 401, errCSRF: 403,
	errNotAvailable: 403, errForbidden: 403, errNotFound: 404, errMethod: 405, errConflict: 409, errUnavailable: 503,
}

// recorder observes one answer for the membership.request event.
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

// event writes one membership.request record: operation, closed outcome and reason, and the origin's host tag.
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
	h.log.LogAttrs(ctx, level, "membership.request", attrs...)
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

// failure maps a facade error onto a closed refusal; any unavailable (possibly durable, including a class-40 abort the
// facade reports as unavailable) is membership_unavailable.
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
// only: owner_subject, owner_authority, last_owner, dependent_bindings, recipient_ineligible, invitation_limit, …).
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

// setKind assigns the requested scope kind; the facade validates it (invalid_request field scope).
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

// nullable renders an optional string as JSON null when absent.
func nullable(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func memberJSON(m access.Member) map[string]any {
	return map[string]any{"user_id": formatID(m.UserID), "display_name": nullable(m.DisplayName), "role": m.Role,
		"self": m.Self, "disabled": m.Disabled, "can_change_role": m.CanChangeRole, "can_remove": m.CanRemove,
		"can_promote": m.CanPromote, "can_demote": m.CanDemote, "can_transfer": m.CanTransfer}
}

// invitationJSON renders an invitation; intended_address appears only when the facade supplied it (the issuing scope's
// administrators), never in the recipient view.
func invitationJSON(v access.Invitation) map[string]any {
	out := map[string]any{"id": formatID(v.ID), "scope": string(v.Scope.Kind), "group_id": nil, "role": v.Role,
		"issuer_display_name": nullable(v.IssuerDisplayName), "intended_user_id": nil,
		"expires_at": v.ExpiresAt.UTC().Format(time.RFC3339), "status": v.Status, "consumed_by": nil}
	if v.Scope.Group != nil {
		out["group_id"] = formatID(*v.Scope.Group)
	}
	if v.IntendedUserID != nil {
		out["intended_user_id"] = formatID(*v.IntendedUserID)
	}
	if v.IntendedAddress != nil {
		out["intended_address"] = *v.IntendedAddress
	}
	if v.ConsumedBy != nil {
		out["consumed_by"] = formatID(*v.ConsumedBy)
	}
	return out
}

func invitationPageJSON(p access.InvitationPage) map[string]any {
	rows := make([]map[string]any, 0, len(p.Invitations))
	for _, v := range p.Invitations {
		rows = append(rows, invitationJSON(v))
	}
	out := map[string]any{"invitations": rows, "next": nil}
	if p.Next != nil {
		out["next"] = formatID(*p.Next)
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

func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "list_members", rec, &origin)
	s, q, ok := h.read(rec, r, &origin, "scope", "group_id", "after", "limit", "role")
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
	p, err := h.acc.ListMembers(s.ctx, s.tc, sc, q["after"], limit, q["role"])
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	rows := make([]map[string]any, 0, len(p.Members))
	for _, m := range p.Members {
		rows = append(rows, memberJSON(m))
	}
	out := map[string]any{"members": rows, "next": nil}
	if p.Next != nil {
		out["next"] = formatID(*p.Next)
	}
	writeJSON(rec, http.StatusOK, out)
}

func (h *Handler) selfMember(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "self_member", rec, &origin)
	s, q, ok := h.read(rec, r, &origin, "scope", "group_id")
	if !ok {
		return
	}
	defer s.end()
	sc, ok := queryScope(q)
	if !ok {
		h.refuse(rec, errInvalidRequest, "invalid_id")
		return
	}
	m, err := h.acc.SelfMembership(s.ctx, s.tc, sc)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, map[string]any{"member": map[string]any{"user_id": formatID(m.UserID),
		"display_name": nullable(m.DisplayName), "role": m.Role, "self": true, "can_leave": m.CanLeave,
		"last_owner": m.LastOwner, "can_manage": m.CanManage}})
}

func (h *Handler) listInvitations(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "list_invitations", rec, &origin)
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
	p, err := h.acc.ListInvitations(s.ctx, s.tc, sc, q["after"], limit)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, invitationPageJSON(p))
}

func (h *Handler) myInvitations(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "my_invitations", rec, &origin)
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
	p, err := h.acc.ListMyInvitations(s.ctx, s.tc, q["after"], limit)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, invitationPageJSON(p))
}

func (h *Handler) invitationCandidates(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "invitation_candidates", rec, &origin)
	s, q, ok := h.read(rec, r, &origin, "group_id", "after", "limit")
	if !ok {
		return
	}
	defer s.end()
	limit, ok := page(q)
	if !ok {
		h.refuse(rec, errInvalidRequest, "invalid_page_size")
		return
	}
	var g access.GroupID
	if !parseID(&g, q["group_id"]) {
		h.refuse(rec, errInvalidRequest, "invalid_id")
		return
	}
	p, err := h.acc.ListInvitationCandidates(s.ctx, s.tc, g, q["after"], limit)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	rows := make([]map[string]any, 0, len(p.Candidates))
	for _, c := range p.Candidates {
		rows = append(rows, map[string]any{"user_id": formatID(c.UserID), "display_name": nullable(c.DisplayName)})
	}
	out := map[string]any{"candidates": rows, "next": nil}
	if p.Next != nil {
		out["next"] = formatID(*p.Next)
	}
	writeJSON(rec, http.StatusOK, out)
}

func (h *Handler) getInvitation(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "get_invitation", rec, &origin)
	s, _, ok := h.read(rec, r, &origin)
	if !ok {
		return
	}
	defer s.end()
	var id access.Invitation
	if !parseID(&id.ID, r.PathValue("id")) {
		h.fail(rec, errNotFound)
		return
	}
	v, err := h.acc.GetInvitation(s.ctx, s.tc, id.ID)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, map[string]any{"invitation": invitationJSON(v)})
}

// retirementEligibility answers the session subject's own sole-owned scope ids (IDENTITY shape: no tenant context, no
// resolver decision; the credential is resolved at its bound origin by the facade, the /api/account trust). There is no
// retire action.
func (h *Handler) retirementEligibility(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "retirement_eligibility", rec, &origin)
	q, ok := query(r)
	if !ok {
		h.fail(rec, errAmbiguous)
		return
	}
	origin = q["origin"]
	v := r.Header.Values("Authorization")
	if len(v) > 1 {
		h.fail(rec, errAmbiguous)
		return
	}
	cred := ""
	if len(v) == 1 {
		if c, ok := strings.CutPrefix(v[0], "Session "); ok && !strings.ContainsAny(c, " \t") {
			cred = c
		}
	}
	if _, ok := originHost(origin); !ok {
		h.fail(rec, errOriginRejected)
		return
	}
	if cred == "" {
		h.fail(rec, errUnauthenticated)
		return
	}
	b, err := h.acc.RetirementEligibility(r.Context(), cred, origin)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	scopes := make([]map[string]any, 0, len(b.Scopes))
	for _, x := range b.Scopes {
		scopes = append(scopes, map[string]any{"organization_id": formatID(x.OrganizationID),
			"scope_type": string(x.ScopeType), "scope_id": formatID(x.ScopeID)})
	}
	writeJSON(rec, http.StatusOK, map[string]any{"count": b.Count, "scopes": scopes, "truncated": b.Truncated})
}

// body is every JSON mutation's body; each route reads its own fields and refuses the others.
type body struct {
	Origin          string  `json:"origin"`
	CSRFToken       string  `json:"csrf_token"`
	Scope           *string `json:"scope"`
	GroupID         *string `json:"group_id"`
	Role            *string `json:"role"`
	RecipientUserID *string `json:"recipient_user_id"`
	FormerRole      *string `json:"former_role"`
	IntendedUserID  *string `json:"intended_user_id"`
	IntendedAddress *string `json:"intended_address"`
	ExpiresInHours  *int    `json:"expires_in_hours"`
	Token           *string `json:"token"`
	Enabled         *bool   `json:"enabled"` // G060: the Group-management override write
}

// fields names the optional members present in the body.
func (in body) fields() map[string]bool {
	out := map[string]bool{}
	for name, set := range map[string]bool{"scope": in.Scope != nil, "group_id": in.GroupID != nil, "role": in.Role != nil,
		"recipient_user_id": in.RecipientUserID != nil, "former_role": in.FormerRole != nil,
		"intended_user_id": in.IntendedUserID != nil, "intended_address": in.IntendedAddress != nil,
		"expires_in_hours": in.ExpiresInHours != nil, "token": in.Token != nil, "enabled": in.Enabled != nil} {
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

// scope is the body's administered scope (scope required, group_id optional by shape).
func (in body) scope() (access.Scope, bool) { return scopeOf(*in.Scope, in.GroupID) }

// mutation reads at most maxBody bytes (more is invalid_request before any decode, guard or write), decodes the body
// strictly, authenticates with CSRF, resolves the context and runs act; it writes one membership.request.
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

// scoped decodes the subject path value and the body's scope of a member route: an unparsable subject is not_found; a
// body of the wrong shape or a malformed Group id is invalid_request.
func (h *Handler) scoped(w *recorder, r *http.Request, in body, required ...string) (access.Scope, access.Member, bool) {
	var m access.Member
	if r.PathValue("user_id") != "" && !parseID(&m.UserID, r.PathValue("user_id")) {
		h.fail(w, errNotFound)
		return access.Scope{}, m, false
	}
	if !in.shape(append([]string{"scope"}, required...), []string{"group_id"}) {
		h.fail(w, errInvalidRequest)
		return access.Scope{}, m, false
	}
	sc, ok := in.scope()
	if !ok {
		h.refuse(w, errInvalidRequest, "invalid_id")
		return access.Scope{}, m, false
	}
	return sc, m, true
}

// groupAuthorityJSON renders the active Organization's Group-management override state (G060).
func groupAuthorityJSON(v access.GroupAuthorityView) map[string]any {
	return map[string]any{"override_enabled": v.OverrideEnabled, "can_change": v.CanChange}
}

// groupAuthority reads the override (every current Organization member; can_change is the owner hint).
func (h *Handler) groupAuthority(w http.ResponseWriter, r *http.Request) {
	rec, origin := &recorder{ResponseWriter: w}, ""
	defer h.event(r.Context(), "group_authority", rec, &origin)
	s, _, ok := h.read(rec, r, &origin)
	if !ok {
		return
	}
	defer s.end()
	v, err := h.acc.GetGroupAuthority(s.ctx, s.tc)
	if err != nil {
		h.refuseErr(rec, err)
		return
	}
	writeJSON(rec, http.StatusOK, groupAuthorityJSON(v))
}

// setGroupAuthority assigns the override (the current Organization owner only; anyone else is forbidden). The body is
// exactly origin, csrf_token and enabled.
func (h *Handler) setGroupAuthority(w *recorder, _ *http.Request, s scope, in body) {
	if !in.shape([]string{"enabled"}, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.SetGroupManagementOverride(s.ctx, s.tc, *in.Enabled)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, groupAuthorityJSON(v))
}

func (h *Handler) changeRole(w *recorder, r *http.Request, s scope, in body) {
	sc, m, ok := h.scoped(w, r, in, "role")
	if !ok {
		return
	}
	v, err := h.acc.ChangeMemberRole(s.ctx, s.tc, sc, m.UserID, *in.Role)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"member": memberJSON(v)})
}

func (h *Handler) removeMember(w *recorder, r *http.Request, s scope, in body) {
	sc, m, ok := h.scoped(w, r, in)
	if !ok {
		return
	}
	if err := h.acc.RemoveMember(s.ctx, s.tc, sc, m.UserID); err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *Handler) promoteOwner(w *recorder, r *http.Request, s scope, in body) {
	sc, m, ok := h.scoped(w, r, in)
	if !ok {
		return
	}
	v, err := h.acc.PromoteToOwner(s.ctx, s.tc, sc, m.UserID)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"member": memberJSON(v)})
}

func (h *Handler) demoteOwner(w *recorder, r *http.Request, s scope, in body) {
	sc, m, ok := h.scoped(w, r, in, "role")
	if !ok {
		return
	}
	v, err := h.acc.DemoteOwner(s.ctx, s.tc, sc, m.UserID, *in.Role)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"member": memberJSON(v)})
}

func (h *Handler) leaveScope(w *recorder, r *http.Request, s scope, in body) {
	sc, _, ok := h.scoped(w, r, in)
	if !ok {
		return
	}
	if err := h.acc.LeaveScope(s.ctx, s.tc, sc); err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *Handler) transferOwnership(w *recorder, r *http.Request, s scope, in body) {
	sc, _, ok := h.scoped(w, r, in, "recipient_user_id", "former_role")
	if !ok {
		return
	}
	var m access.Member
	if !parseID(&m.UserID, *in.RecipientUserID) {
		h.refuse(w, errInvalidRequest, "invalid_id")
		return
	}
	if err := h.acc.TransferOwnership(s.ctx, s.tc, sc, m.UserID, *in.FormerRole); err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *Handler) issueInvitation(w *recorder, _ *http.Request, s scope, in body) {
	if !in.shape([]string{"scope", "role"}, []string{"group_id", "intended_user_id", "intended_address", "expires_in_hours"}) ||
		(in.IntendedUserID == nil) == (in.IntendedAddress == nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	sc, ok := in.scope()
	if !ok {
		h.refuse(w, errInvalidRequest, "invalid_id")
		return
	}
	var rcp access.Recipient
	if in.IntendedUserID != nil {
		var m access.Member
		if !parseID(&m.UserID, *in.IntendedUserID) {
			h.refuse(w, errInvalidRequest, "invalid_id")
			return
		}
		rcp.UserID = &m.UserID
	} else {
		rcp.Address = in.IntendedAddress
	}
	// Absent is the facade's default lifetime; a present value outside 1..maxLifetimeHours hours is passed as an
	// out-of-range duration so that the facade refuses it with its own field code.
	var lifetime time.Duration
	if in.ExpiresInHours != nil {
		lifetime = -1
		if n := *in.ExpiresInHours; n >= 1 && n <= maxLifetimeHours {
			lifetime = time.Duration(n) * time.Hour
		}
	}
	v, err := h.acc.IssueInvitation(s.ctx, s.tc, sc, *in.Role, rcp, lifetime)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"invitation": invitationJSON(v.Invitation), "token": v.Token})
}

func (h *Handler) revokeInvitation(w *recorder, r *http.Request, s scope, in body) {
	var id access.Invitation
	if !parseID(&id.ID, r.PathValue("id")) {
		h.fail(w, errNotFound)
		return
	}
	if !in.shape(nil, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	if err := h.acc.RevokeInvitation(s.ctx, s.tc, id.ID); err != nil {
		h.refuseErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *Handler) acceptInvitation(w *recorder, r *http.Request, s scope, in body) {
	var id access.Invitation
	if !parseID(&id.ID, r.PathValue("id")) {
		h.fail(w, errNotFound)
		return
	}
	if !in.shape([]string{"token"}, nil) {
		h.fail(w, errInvalidRequest)
		return
	}
	v, err := h.acc.AcceptGroupInvitation(s.ctx, s.tc, id.ID, *in.Token)
	if err != nil {
		h.refuseErr(w, err)
		return
	}
	out := map[string]any{"outcome": v.Outcome, "group_id": nil}
	if v.Binding != nil && v.Binding.GroupID != nil {
		out["group_id"] = formatID(*v.Binding.GroupID)
	}
	writeJSON(w, http.StatusOK, out)
}
