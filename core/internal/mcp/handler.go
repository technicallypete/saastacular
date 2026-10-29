package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/time/rate"

	"core/internal/identity"
	"core/internal/identity/workos"
	"core/internal/observability"
	"core/internal/tenancy/access"
)

// Transport bounds (Plan r8 §3.6, r1 §3.6; C5).
const (
	maxBody         = 64 << 10 // MCP request body
	maxInternalBody = 16 << 10 // internal setup request body
	requestDeadline = 10 * time.Second
	perConnRate     = 20 // tokens per second per connection
	perConnBurst    = 40
	maxInFlight     = 64
	maxLimiters     = 10000
)

// Verifier validates a Connect access token for one exact resource; *workos.ConnectValidator satisfies it.
type Verifier interface {
	Validate(ctx context.Context, raw, resource string) (workos.ConnectClaims, error)
}

// Identity is the identity boundary of the transport; *identity.Service satisfies it.
type Identity interface {
	MapConnectSubject(ctx context.Context, subject string) (string, error)
	ReadMCPConnection(ctx context.Context, userID, organizationID, connectionID string) (identity.MCPConnection, error)
	RenewMCPConnection(ctx context.Context, c identity.MCPConnection, idle time.Duration) error
	RevokeMCPConnection(ctx context.Context, c identity.MCPConnection) error
	RevokeOwnMCPConnection(ctx context.Context, userID, organizationID, connectionID string) error
	Session(ctx context.Context, credential, origin string) (identity.SessionInfo, error)
	Now() time.Time
}

// Config wires the handler. Every field is required except AllowedOrigins and Logger.
type Config struct {
	Issuer         string     // MCP_CONNECT_ISSUER, named in the protected-resource metadata
	Hosts          HostLookup // G016: the durable resolver (Organization slugs and verified domains)
	Verifier       Verifier
	Identity       *identity.Service
	Pool           *pgxpool.Pool // the runtime-role pool of the access facade
	Lifetimes      identity.MCPLifetimes
	AllowedOrigins []string // browser origins allowed to call the MCP endpoint (none by default)
	Title          string   // the configured product name (server info title)
	Version        string
	Logger         *slog.Logger
}

// Handler serves the per-connection MCP endpoint, its protected-resource metadata and the internal setup routes.
type Handler struct {
	cfg      Config
	id       Identity
	acc      *access.Service
	stream   *sdk.StreamableHTTPHandler
	inFlight chan struct{}

	mu       sync.Mutex
	limiters map[string]*rate.Limiter

	// fixture is this handler's private test-only catalog override (catalog_fixture_test.go): nil in production, where
	// the handler serves productionCatalog. A published catalog is immutable; tests replace or clear the pointer only.
	fixture atomic.Pointer[catalog]
}

// New returns a handler over the access facade it builds on cfg.Pool and cfg.Identity (the only importer of the facade
// is this package), or an error for an incomplete configuration.
func New(cfg Config) (*Handler, error) {
	if cfg.Issuer == "" || cfg.Verifier == nil || cfg.Identity == nil || cfg.Pool == nil || cfg.Hosts == nil ||
		cfg.Lifetimes.Pending <= 0 || cfg.Lifetimes.Idle <= 0 || cfg.Lifetimes.Absolute < cfg.Lifetimes.Idle {
		return nil, errors.New("mcp: incomplete configuration")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	h := &Handler{cfg: cfg, id: cfg.Identity, acc: access.NewService(cfg.Pool, cfg.Identity, cfg.Logger),
		inFlight: make(chan struct{}, maxInFlight), limiters: map[string]*rate.Limiter{}}
	// The one streamable HTTP handler call site (B-MCP-SOURCE): stateless JSON responses; no Mcp-Session-Id is issued
	// or read (S0); GET and DELETE answer 405. The server of each request is built per call from its verified binding
	// (callState); the SDK's loopback rebinding protection is replaced by the exact registry Host check that runs first
	// (S0 reviewed setting), and the cross-origin check by the Origin allow-list below.
	h.stream = sdk.NewStreamableHTTPHandler(func(r *http.Request) *sdk.Server {
		if st := stateOf(r.Context()); st != nil {
			return st.server
		}
		return nil
	}, &sdk.StreamableHTTPOptions{
		Stateless:                  true,
		JSONResponse:               true,
		DisableLocalhostProtection: true,
		MaxRequestBodyBytes:        maxBody,
	})
	return h, nil
}

// Wrap returns next with the two public MCP path families intercepted on their raw path (before the mux could clean or
// redirect it): every request under them is answered here, and nothing else changes.
func (h *Handler) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.EscapedPath()
		switch {
		case strings.HasPrefix(raw, ConnectionPathPrefix):
			h.serveConnection(w, r, raw)
		case strings.HasPrefix(raw, MetadataPathPrefix):
			h.serveMetadata(w, r, raw)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	hd := w.Header()
	hd.Set("Content-Type", "application/json")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// serveMetadata answers the stateless protected-resource metadata (Plan r8 §3.2, r3 item 1): built only from a
// registered host and a syntactically valid canonical UUID path, with no store lookup, so it carries no existence or
// tenant signal. scopes_supported is omitted (C7 alternative: no scope set was qualified); scopes never confer authority.
func (h *Handler) serveMetadata(w http.ResponseWriter, r *http.Request, raw string) {
	host, err := lookupHost(r.Context(), h.cfg.Hosts, r.Host)
	if err != nil && !unknownHost(err) {
		plainError(w, http.StatusServiceUnavailable, "unavailable")
		return
	}
	id, okPath := connectionID(raw, MetadataPathPrefix)
	if err != nil || !okPath || r.URL.RawQuery != "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"resource":                 resourceURL(host.Host, id),
		"authorization_servers":    []string{h.cfg.Issuer},
		"bearer_methods_supported": []string{"header"},
	})
}

// unauthorized answers 401 with the per-connection resource metadata in WWW-Authenticate; invalid adds
// error="invalid_token". It names nothing about the connection's existence or state.
func unauthorized(w http.ResponseWriter, host, id string, invalid bool) {
	v := `Bearer resource_metadata="` + metadataURL(host, id) + `"`
	body := map[string]string{"error": "unauthenticated"}
	if invalid {
		v = `Bearer error="invalid_token", resource_metadata="` + metadataURL(host, id) + `"`
		body["error"] = "invalid_token"
	}
	w.Header().Set("WWW-Authenticate", v)
	writeJSON(w, http.StatusUnauthorized, body)
}

func plainError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

// bearerOf extracts exactly one "Authorization: Bearer <token>" value. absent is true without any Authorization header;
// ok is false for a repeated header, another scheme or a malformed token.
func bearerOf(r *http.Request) (token string, absent, ok bool) {
	v := r.Header.Values("Authorization")
	switch len(v) {
	case 0:
		return "", true, false
	case 1:
	default:
		return "", false, false
	}
	scheme, tok, found := strings.Cut(v[0], " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || tok == "" || len(tok) > maxInternalBody {
		return "", false, false
	}
	for i := 0; i < len(tok); i++ {
		c := tok[i]
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
			return "", false, false
		}
	}
	return tok, false, true
}

// limiter returns the per-connection token bucket (C5).
func (h *Handler) limiter(id string) *rate.Limiter {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, ok := h.limiters[id]
	if !ok {
		if len(h.limiters) >= maxLimiters {
			h.limiters = map[string]*rate.Limiter{}
		}
		l = rate.NewLimiter(rate.Limit(perConnRate), perConnBurst)
		h.limiters[id] = l
	}
	return l
}

// serveConnection is one protected MCP call (Plan r8 §4.3): exact path and registry host (else 404), method, query,
// Origin and admission; the body bound and the batch refusal; the bearer (absent 401 with discovery, malformed 400);
// token verification for the exact resource (no database); the subject mapping (E2); the connection read and binding
// agreement (host, audience, client, live session and owner); a fresh access scope; activation of a pending connection
// (M2) or due renewal (M5) before any tool runs; then the per-request server.
func (h *Handler) serveConnection(w http.ResponseWriter, r *http.Request, raw string) {
	rctx := r.Context()
	host, err := lookupHost(rctx, h.cfg.Hosts, r.Host)
	if err != nil && !unknownHost(err) {
		h.requestRefused(rctx, w, http.StatusServiceUnavailable, "unavailable", reasonUnavailable)
		return
	}
	id, okPath := connectionID(raw, ConnectionPathPrefix)
	if err != nil || !okPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		h.requestRefused(rctx, w, http.StatusMethodNotAllowed, "method_not_allowed", reasonMethodNotAllowed)
		return
	}
	if r.URL.RawQuery != "" || r.URL.ForceQuery {
		h.requestRefused(rctx, w, http.StatusBadRequest, "invalid_request", reasonInvalidRequest)
		return
	}
	if o := r.Header.Values("Origin"); len(o) > 1 || (len(o) == 1 && !slices.Contains(h.cfg.AllowedOrigins, o[0])) {
		h.requestRefused(rctx, w, http.StatusForbidden, "origin_rejected", reasonOriginRejected)
		return
	}
	// The body is read under its own deadline before an in-flight slot is claimed, so a slow or stalled body never
	// holds a slot (the server sets only a header timeout). A writer without deadline support keeps the size bound.
	_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(requestDeadline))
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			h.requestRefused(rctx, w, http.StatusRequestEntityTooLarge, "request_too_large", reasonRequestTooLarge)
			return
		}
		h.requestRefused(rctx, w, http.StatusBadRequest, "invalid_request", reasonInvalidRequest)
		return
	}
	select {
	case h.inFlight <- struct{}{}:
		defer func() { <-h.inFlight }()
	default:
		h.requestRefused(rctx, w, http.StatusTooManyRequests, "rate_limited", reasonRateLimited)
		return
	}
	ctx, cancel := context.WithTimeout(rctx, requestDeadline)
	defer cancel()
	if t := bytes.TrimLeft(body, " \t\r\n"); len(t) > 0 && t[0] == '[' {
		h.requestRefused(ctx, w, http.StatusBadRequest, "batch_unsupported", reasonBatchUnsupported)
		return
	}

	token, absent, okBearer := bearerOf(r)
	switch {
	case absent:
		h.requestEvent(ctx, reasonTokenMissing)
		unauthorized(w, host.Host, id, false)
		return
	case !okBearer:
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_request", resource_metadata="`+metadataURL(host.Host, id)+`"`)
		h.requestRefused(ctx, w, http.StatusBadRequest, "invalid_request", reasonInvalidRequest)
		return
	}
	resource := resourceURL(host.Host, id)
	claims, err := h.cfg.Verifier.Validate(ctx, token, resource)
	switch {
	case errors.Is(err, identity.ErrUnavailable):
		h.requestRefused(ctx, w, http.StatusServiceUnavailable, "unavailable", reasonUnavailable)
		return
	case err != nil:
		h.requestEvent(ctx, reasonTokenInvalid)
		unauthorized(w, host.Host, id, true)
		return
	}
	if !h.limiter(id).Allow() {
		h.requestRefused(ctx, w, http.StatusTooManyRequests, "rate_limited", reasonRateLimited)
		return
	}
	user, err := h.id.MapConnectSubject(ctx, claims.Subject)
	if err != nil {
		h.refuse(ctx, w, host.Host, id, err, reasonSubjectUnmapped)
		return
	}
	conn, err := h.id.ReadMCPConnection(ctx, user, host.OrganizationID, id)
	if err != nil {
		h.refuse(ctx, w, host.Host, id, err, reasonConnectionUnknown)
		return
	}
	// Binding agreement: the stored host, resource and (once active) client must equal the request's; the registry still
	// maps the host to the stored Organization (the read is scoped to it). A mismatch on the subject's own connection is
	// a recorded resource_mismatch denial.
	if conn.Host() != host.Host || conn.Resource() != resource || conn.OrganizationID() != host.OrganizationID ||
		(conn.Active() && conn.ClientID() != claims.ClientID) {
		h.acc.RecordResourceMismatch(ctx, conn)
		h.requestEvent(ctx, reasonBindingMismatch)
		unauthorized(w, host.Host, id, true)
		return
	}
	actx, end, err := h.acc.BeginVerifiedRequest(ctx, conn)
	if err != nil {
		h.refuse(ctx, w, host.Host, id, err, reasonContextDenied)
		return
	}
	defer end()
	if !conn.Active() {
		if err := h.acc.ActivateConnection(actx, claims.ClientID); err != nil {
			h.connectionRefused(ctx, w, host.Host, id, err)
			return
		}
		h.connectionEvent(ctx, "activated", "")
	} else if conn.RenewalDue(h.id.Now()) {
		if err := h.id.RenewMCPConnection(actx, conn, h.cfg.Lifetimes.Idle); err != nil {
			h.connectionRefused(ctx, w, host.Host, id, err)
			return
		}
		h.connectionEvent(ctx, "renewed", "")
	}
	st := &callState{}
	st.server = h.newServer(actx, conn)
	r2 := r.WithContext(context.WithValue(actx, stateKey{}, st))
	r2.Body = io.NopCloser(bytes.NewReader(body))
	r2.ContentLength = int64(len(body))
	h.stream.ServeHTTP(w, r2)
}

// outage reports whether err is an outage of the store, the key set or the facade (503, never anonymous).
func outage(err error) bool {
	return errors.Is(err, identity.ErrUnavailable) || access.CodeOf(err) == access.CodeUnavailable || access.CodeOf(err) == access.CodeInternal
}

// refuse maps a failure before the tools run: an outage is 503 (never anonymous), everything else the same 401. The
// mcp.request event names reason (or unavailable for an outage).
func (h *Handler) refuse(ctx context.Context, w http.ResponseWriter, host, id string, err error, reason string) {
	if outage(err) {
		h.requestRefused(ctx, w, http.StatusServiceUnavailable, "unavailable", reasonUnavailable)
		return
	}
	h.requestEvent(ctx, reason)
	unauthorized(w, host, id, true)
}

// connectionRefused maps a failed activation or renewal like refuse, recording mcp.connection instead.
func (h *Handler) connectionRefused(ctx context.Context, w http.ResponseWriter, host, id string, err error) {
	switch {
	case outage(err):
		h.connectionEvent(ctx, "failed", reasonUnavailable)
		plainError(w, http.StatusServiceUnavailable, "unavailable")
	case access.CodeOf(err) == access.CodeContextDenied:
		h.connectionEvent(ctx, "refused", reasonContextDenied)
		unauthorized(w, host, id, true)
	default:
		h.connectionEvent(ctx, "refused", reasonConnectionUnknown)
		unauthorized(w, host, id, true)
	}
}

// The closed reasons of the MCP operational events (Plan r8 §6.2; events.json mcp.request and mcp.connection). The
// events carry no connection UUID, token, subject, client ID, host or path: only the outcome, the reason and the
// request and trace correlation.
const (
	reasonMethodNotAllowed  = "method_not_allowed"
	reasonInvalidRequest    = "invalid_request"
	reasonOriginRejected    = "origin_rejected"
	reasonRateLimited       = "rate_limited"
	reasonRequestTooLarge   = "request_too_large"
	reasonBatchUnsupported  = "batch_unsupported"
	reasonTokenMissing      = "token_missing"
	reasonTokenInvalid      = "token_invalid"
	reasonSubjectUnmapped   = "subject_unmapped"
	reasonConnectionUnknown = "connection_unknown" // unknown, terminal (revoked or expired) or an expired pending setup
	reasonBindingMismatch   = "binding_mismatch"
	reasonContextDenied     = "context_denied"
	reasonUnavailable       = "unavailable"
)

// event writes one operational record with the request and trace correlation of ctx.
func (h *Handler) event(ctx context.Context, level slog.Level, name string, attrs ...slog.Attr) {
	if h.cfg.Logger == nil {
		return
	}
	if rid := observability.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid), slog.String("trace_id", observability.TraceID(ctx)))
	}
	h.cfg.Logger.LogAttrs(context.Background(), level, name, attrs...)
}

// requestEvent records one refused (or, with reasonUnavailable, failed) protected MCP request.
func (h *Handler) requestEvent(ctx context.Context, reason string) {
	level, outcome := slog.LevelWarn, "refused"
	if reason == reasonUnavailable {
		level, outcome = slog.LevelError, "failed"
	}
	h.event(ctx, level, "mcp.request", slog.String("outcome", outcome), slog.String("reason", reason))
}

// requestRefused records the request event and answers status with the closed code.
func (h *Handler) requestRefused(ctx context.Context, w http.ResponseWriter, status int, code, reason string) {
	h.requestEvent(ctx, reason)
	plainError(w, status, code)
}

// connectionEvent records one lifecycle step of a connection: activated or renewed (no reason), refused or failed.
func (h *Handler) connectionEvent(ctx context.Context, outcome, reason string) {
	level := slog.LevelInfo
	attrs := []slog.Attr{slog.String("outcome", outcome)}
	switch outcome {
	case "refused":
		level = slog.LevelWarn
	case "failed":
		level = slog.LevelError
	}
	if reason != "" {
		attrs = append(attrs, slog.String("reason", reason))
	}
	h.event(ctx, level, "mcp.connection", attrs...)
}

// callState carries the per-request server to the streamable handler's server lookup.
type callState struct{ server *sdk.Server }

type stateKey struct{}

func stateOf(ctx context.Context) *callState {
	st, _ := ctx.Value(stateKey{}).(*callState)
	return st
}
