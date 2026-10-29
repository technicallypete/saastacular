package mcp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"core/internal/identity"
	"core/internal/tenancy/access"
)

// The internal setup API (Plan r8 §3.2, §3.4): the web setup page's thin adapter calls it with the browser session in
// "Authorization: Session <credential>". Identity comes only from that credential resolved against PostgreSQL; the
// Organization only from the trusted registry entry of the exact setup origin. Mutations need the session's CSRF token.
// Responses are JSON, no-store, with closed error codes and no CORS headers.

// The closed setup error codes.
const (
	setupInvalidRequest   = "invalid_request"
	setupAmbiguous        = "ambiguous_parameter"
	setupUnsupportedMedia = "unsupported_media_type"
	setupOriginRejected   = "origin_rejected"
	setupUnauthenticated  = "unauthenticated"
	setupCSRFFailed       = "csrf_failed"
	setupNotAvailable     = "not_available"
	setupNotFound         = "not_found"
	setupUnavailable      = "mcp_unavailable"
	// G063 (D-2): a friendly name that is malformed on the wire (invalid UTF-8, an unpaired surrogate escape) or fails
	// identity.NormalizeMCPDisplayName; bounded and non-echoing.
	setupInvalidName = "invalid_name"
)

var setupStatus = map[string]int{
	setupInvalidRequest: http.StatusBadRequest, setupAmbiguous: http.StatusBadRequest, setupInvalidName: http.StatusBadRequest,
	setupUnsupportedMedia: http.StatusUnsupportedMediaType, setupOriginRejected: http.StatusForbidden,
	setupUnauthenticated: http.StatusUnauthorized, setupCSRFFailed: http.StatusForbidden, setupNotAvailable: http.StatusForbidden,
	setupNotFound: http.StatusNotFound, setupUnavailable: http.StatusServiceUnavailable,
}

func setupError(w http.ResponseWriter, code string) {
	if sw, ok := w.(*setupWriter); ok {
		sw.code = code
	}
	writeJSON(w, setupStatus[code], map[string]string{"error": code})
}

// setupWriter observes one setup answer for its mcp.setup event: the status and, on a refusal, the closed code.
type setupWriter struct {
	http.ResponseWriter
	status int
	code   string
}

func (s *setupWriter) WriteHeader(status int) {
	if s.status == 0 {
		s.status = status
	}
	s.ResponseWriter.WriteHeader(status)
}

func (s *setupWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

func (s *setupWriter) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// setupEvent records one setup request (Plan r8 §6.2; events.json mcp.setup): the operation, the outcome and, on a
// refusal or failure, the closed code; never an origin, connection, session or CSRF value.
func (h *Handler) setupEvent(r *http.Request, operation string, sw *setupWriter) {
	attrs := []slog.Attr{slog.String("operation", operation)}
	level := slog.LevelInfo
	switch {
	case sw.status >= 200 && sw.status < 300:
		attrs = append(attrs, slog.String("outcome", "ok"))
	case sw.code == setupUnavailable:
		level = slog.LevelError
		attrs = append(attrs, slog.String("outcome", "failed"), slog.String("reason", sw.code))
	case sw.code != "":
		level = slog.LevelWarn
		attrs = append(attrs, slog.String("outcome", "refused"), slog.String("reason", sw.code))
	default:
		return
	}
	h.event(r.Context(), level, "mcp.setup", attrs...)
}

// sessionCredential extracts "Authorization: Session <credential>" ("" when absent or of another shape); a repeated
// header is ambiguous.
func sessionCredential(r *http.Request) (string, bool) {
	v := r.Header.Values("Authorization")
	switch len(v) {
	case 0:
		return "", true
	case 1:
	default:
		return "", false
	}
	cred, found := strings.CutPrefix(v[0], "Session ")
	if !found || strings.ContainsAny(cred, " \t") {
		return "", true
	}
	return cred, true
}

// decodeSetup enforces the mutation input rules in order: exactly one application/json Content-Type (415), then a body
// of at most 16 KiB holding exactly one JSON object with only known fields (400).
func decodeSetup(w http.ResponseWriter, r *http.Request, dst any) string {
	_, code := decodeSetupBody(w, r, dst)
	return code
}

// decodeSetupBody is decodeSetup that also returns the bounded raw body (for the strict name-field scan).
func decodeSetupBody(w http.ResponseWriter, r *http.Request, dst any) ([]byte, string) {
	cts := r.Header.Values("Content-Type")
	if len(cts) != 1 {
		return nil, setupUnsupportedMedia
	}
	mt, params, err := mime.ParseMediaType(cts[0])
	if err != nil || mt != "application/json" {
		return nil, setupUnsupportedMedia
	}
	for k, v := range params {
		if k != "charset" || !strings.EqualFold(v, "utf-8") {
			return nil, setupUnsupportedMedia
		}
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxInternalBody))
	if err != nil {
		return nil, setupInvalidRequest
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return nil, setupInvalidRequest
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, setupInvalidRequest
	}
	return body, ""
}

// The friendly-name field (G063 Plan r1 §4.3). encoding/json replaces invalid UTF-8 and invalid surrogate escapes with
// U+FFFD instead of failing, so the name is validated on its raw token before any Go string exists. Only this one key is
// read strictly; every other field keeps decodeSetup's behavior (last-wins duplicates, replacement decoding, then its
// own exact checks).
const setupNameKey = "display_name"

// nameField scans the top level of an already decoded, bounded setup body for the name key: the raw value of the last
// match and the number of matches. A key matches as encoding/json would match it to the struct field (exactly or
// case-insensitively), so a second spelling is a duplicate rather than a silently decoded alias.
func nameField(body []byte) (raw json.RawMessage, count int, ok bool) {
	dec := json.NewDecoder(bytes.NewReader(body))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, 0, false
	}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, 0, false
		}
		key, isKey := t.(string)
		if !isKey {
			return nil, 0, false
		}
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return nil, 0, false
		}
		if strings.EqualFold(key, setupNameKey) {
			raw, count = v, count+1
		}
	}
	return raw, count, true
}

// surrogatesPaired reports whether every \u escape of the raw JSON string token in the surrogate range D800-DFFF is a
// high surrogate immediately followed by a \u low surrogate (hexadecimal digits in either case).
func surrogatesPaired(raw []byte) bool {
	hex4 := func(at int) (rune, bool) {
		if at+4 > len(raw) {
			return 0, false
		}
		var r rune
		for _, c := range raw[at : at+4] {
			switch {
			case c >= '0' && c <= '9':
				r = r<<4 | rune(c-'0')
			case c >= 'a' && c <= 'f':
				r = r<<4 | rune(c-'a'+10)
			case c >= 'A' && c <= 'F':
				r = r<<4 | rune(c-'A'+10)
			default:
				return 0, false
			}
		}
		return r, true
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		i++ // the escaped character
		if i >= len(raw) {
			return false
		}
		if raw[i] != 'u' {
			continue
		}
		r, ok := hex4(i + 1)
		if !ok {
			return false
		}
		i += 4 // the last hexadecimal digit
		switch {
		case r >= 0xDC00 && r <= 0xDFFF:
			return false // a low surrogate without its high surrogate (lone or reversed)
		case r >= 0xD800 && r <= 0xDBFF:
			if i+2 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
				return false
			}
			lo, ok := hex4(i + 3)
			if !ok || lo < 0xDC00 || lo > 0xDFFF {
				return false
			}
			i += 6
		}
	}
	return true
}

// setupName reads the optional friendly name of a setup body: absent is (nil, false, ""); a duplicate, null or
// non-string value is invalid_request; malformed UTF-8 or an unpaired surrogate escape, or a value refused by
// identity.NormalizeMCPDisplayName, is invalid_name; a value that normalizes to empty is (nil, true, "") (clear).
func setupName(body []byte) (*string, bool, string) {
	raw, count, ok := nameField(body)
	raw = bytes.TrimSpace(raw) // JSON whitespace around the value token, if the decoder kept any
	switch {
	case !ok || count > 1:
		return nil, false, setupInvalidRequest
	case count == 0:
		return nil, false, ""
	case len(raw) == 0 || raw[0] != '"':
		return nil, true, setupInvalidRequest // null (a clear is ""), number, boolean, array or object
	case !utf8.Valid(raw) || !surrogatesPaired(raw):
		return nil, true, setupInvalidName
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, true, setupInvalidRequest
	}
	name, err := identity.NormalizeMCPDisplayName(s)
	if err != nil {
		return nil, true, setupInvalidName
	}
	return name, true, ""
}

// connectionView is one listed connection: the URL is a selector shown to its owner, never a credential. display_name is
// always present (null for an unnamed connection) and kept separate from the bound client.
func connectionView(i identity.MCPConnectionInfo) map[string]any {
	v := map[string]any{"connection_id": i.ID, "url": i.Resource, "state": i.State, "client_id": nil,
		"created_at": i.CreatedAt.UTC().Format(time.RFC3339), "activated_at": nil, "display_name": nil}
	if i.ClientID != "" {
		v["client_id"] = i.ClientID
	}
	if !i.ActivatedAt.IsZero() {
		v["activated_at"] = i.ActivatedAt.UTC().Format(time.RFC3339)
	}
	if i.DisplayName != nil {
		v["display_name"] = *i.DisplayName
	}
	return v
}

// setupFailure maps a facade error: a refused context (non-member) is the generic not_available; an ended session
// unauthenticated; a target that is not one of the caller's live connections (G063 rename) not_found; anything else
// mcp_unavailable.
func setupFailure(err error) string {
	switch access.CodeOf(err) {
	case access.CodeContextDenied:
		return setupNotAvailable
	case access.CodeUnauthenticated:
		return setupUnauthenticated
	case access.CodeNotFound:
		return setupNotFound
	}
	return setupUnavailable
}

// begin resolves the browser session and, for a mutation, its CSRF token; then opens the access scope and resolves the
// registry Organization (membership). It writes the response and returns ok false on any refusal.
func (h *Handler) begin(w http.ResponseWriter, r *http.Request, origin, csrf string, mutation bool) (ctx contextScope, ok bool) {
	return h.beginWith(w, r, origin, csrf, mutation, nil)
}

// beginWith is begin with an optional request check run after the origin is resolved and before the session is (G063:
// the rename's canonical connection UUID); a non-empty code refuses the request.
func (h *Handler) beginWith(w http.ResponseWriter, r *http.Request, origin, csrf string, mutation bool, afterOrigin func() string) (ctx contextScope, ok bool) {
	cred, single := sessionCredential(r)
	if !single {
		setupError(w, setupAmbiguous)
		return contextScope{}, false
	}
	host, err := lookupOrigin(r.Context(), h.cfg.Hosts, origin)
	switch {
	case err != nil && !unknownHost(err):
		setupError(w, setupUnavailable)
		return contextScope{}, false
	case err != nil:
		setupError(w, setupOriginRejected)
		return contextScope{}, false
	}
	if afterOrigin != nil {
		if code := afterOrigin(); code != "" {
			setupError(w, code)
			return contextScope{}, false
		}
	}
	if cred == "" {
		setupError(w, setupUnauthenticated)
		return contextScope{}, false
	}
	// G016 security amendment: the browser session must be bound to exactly this origin and to the Organization it
	// resolves to now; anything else is unauthenticated, as a dead credential.
	info, err := h.id.Session(r.Context(), cred, origin)
	switch {
	case errors.Is(err, identity.ErrUnauthenticated):
		setupError(w, setupUnauthenticated)
		return contextScope{}, false
	case err != nil:
		setupError(w, setupUnavailable)
		return contextScope{}, false
	case info.OrganizationID != host.OrganizationID:
		setupError(w, setupUnauthenticated)
		return contextScope{}, false
	}
	if mutation && !identity.CSRFMatches(info, csrf) {
		setupError(w, setupCSRFFailed)
		return contextScope{}, false
	}
	actx, end, err := h.acc.BeginRequest(r.Context(), cred, origin)
	if err != nil {
		setupError(w, setupFailure(err))
		return contextScope{}, false
	}
	tc, err := h.acc.Resolve(actx, access.ResolveRequest{OrganizationID: host.OrganizationID})
	if err != nil {
		end()
		setupError(w, setupFailure(err))
		return contextScope{}, false
	}
	return contextScope{ctx: actx, end: end, tc: tc, host: host, user: info.UserID,
		csrf: base64.RawURLEncoding.EncodeToString(info.CSRFToken)}, true
}

// contextScope is one resolved setup request: the access scope (end must be deferred), the resolved Organization-only
// context and the registry entry of the setup origin.
type contextScope struct {
	ctx  context.Context
	end  func()
	tc   *access.TenantContext
	host Host
	user string // the session's user: the cursor context together with the registry entry's Organization and origin
	csrf string // the session's CSRF token, returned by the listing so the setup page needs one core call
}

// The setup listing cursor (Duo follow-up correction): base64url (strict, no padding) of
// "s1|<created_at in Unix microseconds>|<anchor connection UUID>|<context tag>". The position is untrusted and
// editable: it only chooses where the newest-first traversal resumes, and every page re-runs the full authority and
// the owner, Organization and liveness filters, so any position lists only the caller's own live connections. The
// context tag is an unkeyed hash, not a MAC: it rejects a cursor presented in another user, Organization or setup
// origin context without putting those identifiers into a browser URL, and proves neither issuance nor integrity.
const (
	setupCursorVersion = "s1"
	setupCursorMaxLen  = 128
	// The accepted anchor time: 1970-01-01T00:00:00.000001Z through 9999-12-31T23:59:59.999999Z (the four-digit-year
	// RFC 3339 UTC timestamp contract, inside PostgreSQL timestamptz and int64).
	setupCursorMinMicro int64 = 1
	setupCursorMaxMicro int64 = 253402300799999999
)

// setupCursorTag is the context tag of user, Organization and the trusted registry setup origin.
func setupCursorTag(user, org, origin string) string {
	sum := sha256.Sum256([]byte("mcp-setup-cursor/1\x00" + user + "\x00" + org + "\x00" + origin))
	return hex.EncodeToString(sum[:16])
}

// encodeSetupCursor is the cursor after p in the context (user, org, origin); false when p is outside the accepted
// domain (the listing then fails closed).
func encodeSetupCursor(user, org, origin string, p identity.MCPListPosition) (string, bool) {
	us := p.CreatedAt.UnixMicro()
	if us < setupCursorMinMicro || us > setupCursorMaxMicro || !canonicalUUID(p.ID) {
		return "", false
	}
	raw := setupCursorVersion + "|" + strconv.FormatInt(us, 10) + "|" + p.ID + "|" + setupCursorTag(user, org, origin)
	return base64.RawURLEncoding.EncodeToString([]byte(raw)), true
}

// decodeSetupCursor strictly parses a cursor for the context (user, org, origin); any fault is false.
func decodeSetupCursor(s, user, org, origin string) (identity.MCPListPosition, bool) {
	if len(s) == 0 || len(s) > setupCursorMaxLen {
		return identity.MCPListPosition{}, false
	}
	b, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil {
		return identity.MCPListPosition{}, false
	}
	parts := strings.Split(string(b), "|")
	if len(parts) != 4 || parts[0] != setupCursorVersion || !canonicalMicro(parts[1]) || !canonicalUUID(parts[2]) ||
		subtle.ConstantTimeCompare([]byte(parts[3]), []byte(setupCursorTag(user, org, origin))) != 1 {
		return identity.MCPListPosition{}, false
	}
	us, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || us < setupCursorMinMicro || us > setupCursorMaxMicro {
		return identity.MCPListPosition{}, false
	}
	return identity.MCPListPosition{CreatedAt: time.UnixMicro(us).UTC(), ID: parts[2]}, true
}

// canonicalMicro accepts 1 to 18 decimal digits without a leading zero or sign.
func canonicalMicro(s string) bool {
	if len(s) == 0 || len(s) > 18 || s[0] == '0' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// List is GET /internal/v1/mcp/connections?origin=<setup origin>[&cursor=<next_cursor>]: one page (newest first, at
// most identity.MCPListPage) of the caller's own live connections in the Organization of that origin (current
// membership required), with the session's CSRF token for the page's forms and the next page's cursor (null on the
// last page). A cursor is checked only after the session, origin and membership; any cursor fault is invalid_request.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sw := &setupWriter{ResponseWriter: w}
	defer h.setupEvent(r, "list", sw)
	w = sw
	q := r.URL.Query()
	keys := 1
	cursor, hasCursor := q["cursor"]
	if hasCursor {
		keys = 2
	}
	if len(q["origin"]) != 1 || (hasCursor && len(cursor) != 1) || len(q) != keys {
		setupError(w, setupInvalidRequest)
		return
	}
	s, ok := h.begin(w, r, q.Get("origin"), "", false)
	if !ok {
		return
	}
	defer s.end()
	var after *identity.MCPListPosition
	if hasCursor {
		p, valid := decodeSetupCursor(cursor[0], s.user, s.host.OrganizationID, s.host.Origin)
		if !valid {
			setupError(w, setupInvalidRequest)
			return
		}
		after = &p
	}
	page, err := h.acc.ListMCPConnections(s.ctx, s.tc, after)
	if err != nil {
		setupError(w, setupFailure(err))
		return
	}
	out := make([]map[string]any, 0, len(page.Connections))
	for _, i := range page.Connections {
		out = append(out, connectionView(i))
	}
	var next any // JSON null on the last page
	if page.More && len(page.Connections) > 0 {
		last := page.Connections[len(page.Connections)-1]
		c, ok := encodeSetupCursor(s.user, s.host.OrganizationID, s.host.Origin, identity.MCPListPosition{CreatedAt: last.CreatedAt, ID: last.ID})
		if !ok {
			setupError(w, setupUnavailable) // an unencodable anchor fails closed
			return
		}
		next = c
	}
	writeJSON(w, http.StatusOK, map[string]any{"connections": out, "csrf_token": s.csrf, "next_cursor": next})
}

type createBody struct {
	Origin    string `json:"origin"`
	CSRFToken string `json:"csrf_token"`
	// DisplayName admits the optional key; its value is read strictly from the raw body by setupName (G063).
	DisplayName json.RawMessage `json:"display_name"`
}

// Create is POST /internal/v1/mcp/connections {origin, csrf_token[, display_name]}: a new pending connection of the
// caller in the Organization of that origin (M1), with the optional friendly name (G063; settled before the session is
// read, so a malformed name learns nothing about the session or CSRF). It returns the connection with its URL.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	sw := &setupWriter{ResponseWriter: w}
	defer h.setupEvent(r, "create", sw)
	w = sw
	var in createBody
	body, e := decodeSetupBody(w, r, &in)
	if e != "" {
		setupError(w, e)
		return
	}
	name, _, e := setupName(body)
	if e != "" {
		setupError(w, e)
		return
	}
	s, ok := h.begin(w, r, in.Origin, in.CSRFToken, true)
	if !ok {
		return
	}
	defer s.end()
	id, err := identity.NewMCPConnectionID()
	if err != nil {
		setupError(w, setupUnavailable)
		return
	}
	resource := resourceURL(s.host.Host, id)
	if err := h.acc.CreateMCPConnection(s.ctx, s.tc, id, s.host.Host, resource, h.cfg.Lifetimes, name); err != nil {
		setupError(w, setupFailure(err))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"connection": map[string]any{"connection_id": id, "url": resource,
		"state": identity.MCPStatePending}})
}

type renameBody struct {
	Origin       string          `json:"origin"`
	ConnectionID string          `json:"connection_id"`
	CSRFToken    string          `json:"csrf_token"`
	DisplayName  json.RawMessage `json:"display_name"` // required; read strictly by setupName
}

// Rename is POST /internal/v1/mcp/connections/rename {origin, connection_id, csrf_token, display_name} (G063 Plan r1
// §4.3): set, change or (with "" or whitespace only) clear the friendly name of one own live pending or active
// connection through a current host of its Organization (current membership required, D-4). Order: media type, body,
// the name field (before any session or CSRF check), the single Authorization header, the origin registry, the
// canonical connection UUID, the session bound to that origin and Organization, CSRF, then the facade. Unknown,
// foreign-user, foreign-tenant and terminal targets all answer the same not_found.
func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
	sw := &setupWriter{ResponseWriter: w}
	defer h.setupEvent(r, "rename", sw)
	w = sw
	var in renameBody
	body, e := decodeSetupBody(w, r, &in)
	if e != "" {
		setupError(w, e)
		return
	}
	name, present, e := setupName(body)
	if e == "" && !present {
		e = setupInvalidRequest
	}
	if e != "" {
		setupError(w, e)
		return
	}
	s, ok := h.beginWith(w, r, in.Origin, in.CSRFToken, true, func() string {
		if !canonicalUUID(in.ConnectionID) {
			return setupInvalidRequest
		}
		return ""
	})
	if !ok {
		return
	}
	defer s.end()
	if err := h.acc.RenameMCPConnection(s.ctx, s.tc, in.ConnectionID, name); err != nil {
		setupError(w, setupFailure(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"outcome": "renamed"})
}

type revokeBody struct {
	Origin       string `json:"origin"`
	ConnectionID string `json:"connection_id"`
	CSRFToken    string `json:"csrf_token"`
}

// Revoke is POST /internal/v1/mcp/connections/revoke {origin, connection_id, csrf_token}: terminate one own connection
// (M4; connection-local; exact ownership only, no membership needed). A second revoke of the same connection, like
// any connection that is not one of the caller's live connections, answers not_found.
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	sw := &setupWriter{ResponseWriter: w}
	defer h.setupEvent(r, "revoke", sw)
	w = sw
	var in revokeBody
	if e := decodeSetup(w, r, &in); e != "" {
		setupError(w, e)
		return
	}
	cred, single := sessionCredential(r)
	if !single {
		setupError(w, setupAmbiguous)
		return
	}
	host, err := lookupOrigin(r.Context(), h.cfg.Hosts, in.Origin)
	switch {
	case err != nil && !unknownHost(err):
		setupError(w, setupUnavailable)
		return
	case err != nil:
		setupError(w, setupOriginRejected)
		return
	}
	if !canonicalUUID(in.ConnectionID) {
		setupError(w, setupInvalidRequest)
		return
	}
	if cred == "" {
		setupError(w, setupUnauthenticated)
		return
	}
	info, err := h.id.Session(r.Context(), cred, in.Origin) // G016 security amendment: bound origin and Organization
	switch {
	case errors.Is(err, identity.ErrUnauthenticated):
		setupError(w, setupUnauthenticated)
		return
	case err != nil:
		setupError(w, setupUnavailable)
		return
	case info.OrganizationID != host.OrganizationID:
		setupError(w, setupUnauthenticated)
		return
	}
	if !identity.CSRFMatches(info, in.CSRFToken) {
		setupError(w, setupCSRFFailed)
		return
	}
	err = h.id.RevokeOwnMCPConnection(r.Context(), info.UserID, host.OrganizationID, in.ConnectionID)
	switch {
	case errors.Is(err, identity.ErrUnauthenticated):
		setupError(w, setupNotFound)
		return
	case err != nil:
		setupError(w, setupUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"outcome": "revoked"})
}
