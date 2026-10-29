package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ctxKey int

const traceContextKey ctxKey = iota

const (
	sourceInbound   = "inbound"
	sourceGenerated = "generated"

	// maxTraceparentLen bounds an inbound traceparent before it is parsed.
	maxTraceparentLen = 512

	completionEvent = "http.request.completed"
	routeUnmatched  = "unmatched"
)

var (
	requestIDRe   = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)
	traceparentRe = regexp.MustCompile(`^00-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$`)
	// routePatternRe accepts only a method-qualified ServeMux pattern without
	// a host. The mux can store a raw request path in Request.Pattern for some
	// redirects; such a value never has a method prefix, so it is reported as
	// unmatched instead of being logged.
	routePatternRe = regexp.MustCompile(`^[A-Z]+ (/[A-Za-z0-9._~/{}$-]{0,63})$`)
)

// TraceContext is the per-request diagnostic context: the effective request
// ID, the trace, this hop's span and, when the inbound traceparent was valid,
// the caller's parent span. Flags holds the outbound trace flags ("00" or
// "01"); only the sampled bit of an inbound value is kept.
type TraceContext struct {
	RequestID       string
	RequestIDSource string
	TraceID         string
	TraceSource     string
	SpanID          string
	ParentSpanID    string
	Flags           string
}

// parseRequestID accepts exactly one well-formed X-Request-Id field value
// that contains no configured secret.
func parseRequestID(values []string, isSecret func(string) bool) (string, bool) {
	if len(values) != 1 || !requestIDRe.MatchString(values[0]) {
		return "", false
	}
	if isSecret != nil && isSecret(values[0]) {
		return "", false
	}
	return values[0], true
}

// parseTraceparent applies the version-00 profile: exactly one field value of
// at most 512 bytes that fully matches 00-<32 hex>-<16 hex>-<2 hex> in
// lowercase, with nonzero trace and parent IDs. Every other version is
// unsupported. The returned flags keep only the sampled bit.
func parseTraceparent(values []string) (traceID, parentID, flags string, ok bool) {
	if len(values) != 1 || len(values[0]) > maxTraceparentLen {
		return "", "", "", false
	}
	m := traceparentRe.FindStringSubmatch(values[0])
	if m == nil || allZero(m[1]) || allZero(m[2]) {
		return "", "", "", false
	}
	raw, err := hex.DecodeString(m[3])
	if err != nil {
		return "", "", "", false
	}
	flags = "00"
	if raw[0]&0x01 == 0x01 {
		flags = "01"
	}
	return m[1], m[2], flags, true
}

// ParseTraceparent extracts the trace ID from a single traceparent value
// under the version-00 profile. It rejects malformed values, every version
// other than 00 and all-zero identifiers.
func ParseTraceparent(h string) (traceID string, ok bool) {
	traceID, _, _, ok = parseTraceparent([]string{h})
	return traceID, ok
}

func allZero(s string) bool { return strings.Trim(s, "0") == "" }

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// nonZeroHex returns n random bytes as lowercase hex, never all zeros.
func nonZeroHex(n int) string {
	for {
		if s := randomHex(n); !allZero(s) {
			return s
		}
	}
}

// newTraceContext derives the request's diagnostic context from its headers.
// Rejected inbound values are replaced and never kept.
func newTraceContext(h http.Header, isSecret func(string) bool) TraceContext {
	tc := TraceContext{Flags: "00"}
	if id, ok := parseRequestID(h.Values("X-Request-Id"), isSecret); ok {
		tc.RequestID, tc.RequestIDSource = id, sourceInbound
	} else {
		tc.RequestID, tc.RequestIDSource = randomHex(16), sourceGenerated
	}
	if traceID, parentID, flags, ok := parseTraceparent(h.Values("traceparent")); ok {
		tc.TraceID, tc.TraceSource, tc.ParentSpanID, tc.Flags = traceID, sourceInbound, parentID, flags
	} else {
		tc.TraceID, tc.TraceSource = nonZeroHex(16), sourceGenerated
	}
	tc.SpanID = nonZeroHex(8)
	return tc
}

// FromContext returns the diagnostic context stored by Middleware.
func FromContext(ctx context.Context) (TraceContext, bool) {
	tc, ok := ctx.Value(traceContextKey).(TraceContext)
	return tc, ok
}

// RequestID returns the request ID stored in ctx.
func RequestID(ctx context.Context) string { tc, _ := FromContext(ctx); return tc.RequestID }

// TraceID returns the trace ID stored in ctx.
func TraceID(ctx context.Context) string { tc, _ := FromContext(ctx); return tc.TraceID }

// SpanID returns this hop's span ID stored in ctx.
func SpanID(ctx context.Context) string { tc, _ := FromContext(ctx); return tc.SpanID }

// InjectOutbound sets exactly X-Request-Id and traceparent on an outbound
// request from the context of the request being served. The outbound parent
// is this hop's span. Nothing is set when ctx carries no context.
func InjectOutbound(ctx context.Context, h http.Header) {
	tc, ok := FromContext(ctx)
	if !ok {
		return
	}
	h.Set("X-Request-Id", tc.RequestID)
	h.Set("traceparent", "00-"+tc.TraceID+"-"+tc.SpanID+"-"+tc.Flags)
}

// statusRecorder forwards every call unchanged and records the first final
// status: the first WriteHeader, or an implicit 200 on the first Write.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 && ((code >= 200 && code <= 999) || code == http.StatusSwitchingProtocols) {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(p []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(p)
}

// Unwrap exposes the underlying writer to http.ResponseController.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// Middleware assigns a request ID (accepting a well-formed inbound one that
// carries no configured secret), continues a valid inbound traceparent or
// starts a new trace, creates this hop's span, and logs exactly one
// http.request.completed record per request. The record never carries
// headers, bodies, query strings or the raw path. A handler panic is logged as
// a failure and the connection is then aborted with http.ErrAbortHandler.
func Middleware(logger *slog.Logger, next http.Handler) http.Handler {
	isSecret := secretMatcher(logger)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		tc := newTraceContext(r.Header, isSecret)
		r2 := r.WithContext(context.WithValue(r.Context(), traceContextKey, tc))
		w.Header().Set("X-Request-Id", tc.RequestID)
		rec := &statusRecorder{ResponseWriter: w}
		logged := false
		defer func() {
			p := recover()
			if !logged {
				logged = true
				logCompletion(logger, r2, rec.status, tc, time.Since(start), p != nil)
			}
			if p != nil {
				panic(http.ErrAbortHandler)
			}
		}()
		next.ServeHTTP(rec, r2)
	})
}

func logCompletion(logger *slog.Logger, r *http.Request, status int, tc TraceContext, elapsed time.Duration, panicked bool) {
	outcome, errorType := "completed", ""
	switch {
	case panicked:
		outcome, errorType = "failed", "panic"
	case r.Context().Err() != nil:
		outcome = "cancelled"
	}
	if status == 0 && outcome == "completed" {
		status = http.StatusOK // net/http sends an implicit 200 after the handler returns
	}
	level := slog.LevelInfo
	switch {
	case outcome == "failed":
		level = slog.LevelError
	case status >= 500:
		level = slog.LevelWarn
	}
	attrs := make([]slog.Attr, 0, 14)
	attrs = append(attrs,
		slog.String("method", normalizeMethod(r.Method)),
		slog.String("route", routeOf(r.Pattern)),
		slog.Int("status", status),
		slog.Int64("duration_ms", elapsed.Milliseconds()),
		slog.String("request_id", tc.RequestID),
		slog.String("request_id_source", tc.RequestIDSource),
		slog.String("trace_id", tc.TraceID),
		slog.String("trace_source", tc.TraceSource),
		slog.String("span_id", tc.SpanID),
	)
	if tc.ParentSpanID != "" {
		attrs = append(attrs, slog.String("parent_span_id", tc.ParentSpanID))
	}
	attrs = append(attrs, slog.String("trace_flags", tc.Flags), slog.String("outcome", outcome))
	if errorType != "" {
		attrs = append(attrs, slog.String("error_type", errorType))
	}
	logger.LogAttrs(context.Background(), level, completionEvent, attrs...)
}

func normalizeMethod(m string) string {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return m
	}
	return "OTHER"
}

// routeOf returns the path of the matched, method-qualified ServeMux pattern
// (for example "GET /healthz" gives "/healthz"), else "unmatched".
func routeOf(pattern string) string {
	if m := routePatternRe.FindStringSubmatch(pattern); m != nil {
		return m[1]
	}
	return routeUnmatched
}
