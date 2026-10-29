package health_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"core/internal/health"
	"core/internal/observability"
)

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func randHex(t *testing.T, n int) string {
	t.Helper()
	for {
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			t.Fatal(err)
		}
		if s := hex.EncodeToString(b); strings.Trim(s, "0") != "" {
			return s
		}
	}
}

// headerRecorder is an intelligence double that records the headers of every
// probe it receives.
type headerRecorder struct {
	mu   sync.Mutex
	seen []http.Header
}

func (h *headerRecorder) handler(delay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		h.seen = append(h.seen, r.Header.Clone())
		h.mu.Unlock()
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				return
			}
		}
		if r.URL.Path != "/healthz" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *headerRecorder) headers() []http.Header {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.seen)
}

// coreFront serves core's readiness route through the middleware, exactly as
// cmd/server wires it, with only the optional intelligence check configured.
func coreFront(t *testing.T, logs io.Writer, intelligenceURL string) *httptest.Server {
	t.Helper()
	h := health.Handler{ProductName: "Sample", Checks: []health.Check{health.Intelligence(intelligenceURL)}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.Liveness)
	mux.HandleFunc("GET /readyz", h.Readiness)
	srv := httptest.NewServer(observability.Middleware(observability.NewLogger(logs, slog.LevelInfo, nil), mux))
	t.Cleanup(srv.Close)
	return srv
}

func completionRecord(t *testing.T, out string) map[string]any {
	t.Helper()
	var recs []map[string]any
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("non-JSON log line %q", sc.Text())
		}
		if m["event"] == "http.request.completed" {
			recs = append(recs, m)
		}
	}
	if len(recs) != 1 {
		t.Fatalf("want exactly one completion record, got %d: %s", len(recs), out)
	}
	return recs[0]
}

func headerKeys(h http.Header) []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// The probe carries exactly the two diagnostic headers plus Go's transport
// defaults (User-Agent, Accept-Encoding).
var wantProbeHeaders = []string{"Accept-Encoding", "Traceparent", "User-Agent", "X-Request-Id"}

func TestReadinessProbePropagatesOnlyDiagnosticContext(t *testing.T) {
	trace, parent := randHex(t, 16), randHex(t, 8)
	cookie, auth, baggage, state, arbitrary := randHex(t, 8), randHex(t, 8), randHex(t, 8), randHex(t, 8), randHex(t, 8)
	for _, tc := range []struct {
		name        string
		reqID       string
		traceparent string
		wantFlags   string
	}{
		{"valid sampled context", "req-propagate-0001", "00-" + trace + "-" + parent + "-01", "01"},
		{"unknown flag bits dropped", "req-propagate-0002", "00-" + trace + "-" + parent + "-02", "00"},
		{"no inbound context", "", "", "00"},
		{"invalid inbound context replaced", "bad id " + arbitrary, "01-" + trace + "-" + parent + "-01", "00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := &headerRecorder{}
			intel := httptest.NewServer(rec.handler(0))
			defer intel.Close()
			logs := &lockedBuffer{}
			front := coreFront(t, logs, intel.URL)

			req, _ := http.NewRequest(http.MethodGet, front.URL+"/readyz?token="+arbitrary, nil)
			req.Header.Set("Cookie", "session="+cookie)
			req.Header.Set("Authorization", "Bearer "+auth)
			req.Header.Set("baggage", "userId="+baggage)
			req.Header.Set("tracestate", "vendor="+state)
			req.Header.Set("X-Arbitrary-Header", arbitrary)
			req.Header.Set("X-Api-Key", auth)
			if tc.reqID != "" {
				req.Header.Set("X-Request-Id", tc.reqID)
			}
			if tc.traceparent != "" {
				req.Header.Set("traceparent", tc.traceparent)
			}
			resp, err := front.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"optional":{"intelligence":"ok"}`) {
				t.Fatalf("readiness = %d %s", resp.StatusCode, body)
			}
			front.Close()
			r := completionRecord(t, logs.String())

			seen := rec.headers()
			if len(seen) != 1 {
				t.Fatalf("want one probe, got %d", len(seen))
			}
			got := seen[0]
			if keys := headerKeys(got); !slices.Equal(keys, wantProbeHeaders) {
				t.Fatalf("probe header set = %v, want %v", keys, wantProbeHeaders)
			}
			for _, h := range []string{"X-Request-Id", "Traceparent"} {
				if len(got.Values(h)) != 1 {
					t.Fatalf("%s sent %d times", h, len(got.Values(h)))
				}
			}
			for _, v := range []string{cookie, auth, baggage, state, arbitrary} {
				if strings.Contains(strings.Join(got.Values("X-Request-Id"), "")+got.Get("Traceparent"), v) {
					t.Fatalf("inbound value %q forwarded", v)
				}
			}

			respID := resp.Header.Get("X-Request-Id")
			if got.Get("X-Request-Id") != respID || r["request_id"] != respID {
				t.Fatalf("request ID: probe %q, response %q, record %v", got.Get("X-Request-Id"), respID, r["request_id"])
			}
			coreSpan, _ := r["span_id"].(string)
			wantTP := "00-" + r["trace_id"].(string) + "-" + coreSpan + "-" + tc.wantFlags
			if got.Get("Traceparent") != wantTP || r["trace_flags"] != tc.wantFlags {
				t.Fatalf("probe traceparent = %q, want %q (record %v)", got.Get("Traceparent"), wantTP, r)
			}
			valid := strings.HasPrefix(tc.traceparent, "00-")
			if valid {
				if r["trace_id"] != trace || r["parent_span_id"] != parent || coreSpan == parent {
					t.Fatalf("trace not continued with core's span as parent: %v", r)
				}
			} else if r["trace_id"] == trace || r["trace_source"] != "generated" {
				t.Fatalf("invalid traceparent adopted: %v", r)
			}
			if tc.reqID == "req-propagate-0001" || tc.reqID == "req-propagate-0002" {
				if respID != tc.reqID {
					t.Fatalf("valid request ID not kept: %q", respID)
				}
			} else if len(respID) != 32 || r["request_id_source"] != "generated" {
				t.Fatalf("request ID not generated: %q", respID)
			}
			if strings.Contains(logs.String(), arbitrary) {
				t.Fatalf("inbound value logged: %s", logs.String())
			}
		})
	}
}

func TestReadinessBoundedBySlowIntelligence(t *testing.T) {
	rec := &headerRecorder{}
	intel := httptest.NewServer(rec.handler(5 * time.Second))
	defer intel.Close()
	logs := &lockedBuffer{}
	front := coreFront(t, logs, intel.URL)

	req, _ := http.NewRequest(http.MethodGet, front.URL+"/readyz", nil)
	req.Header.Set("X-Request-Id", "req-slow-intelligence")
	start := time.Now()
	resp, err := front.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	elapsed := time.Since(start)
	if elapsed < 1500*time.Millisecond || elapsed > 3500*time.Millisecond {
		t.Fatalf("readiness took %v, want about 2s", elapsed)
	}
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"optional":{"intelligence":"unavailable"}`) {
		t.Fatalf("readiness = %d %s", resp.StatusCode, body)
	}
	front.Close()
	r := completionRecord(t, logs.String())
	if r["request_id"] != "req-slow-intelligence" || r["status"] != float64(200) || r["outcome"] != "completed" || r["route"] != "/readyz" {
		t.Fatalf("record = %v", r)
	}
	if seen := rec.headers(); len(seen) != 1 || seen[0].Get("X-Request-Id") != "req-slow-intelligence" {
		t.Fatalf("probe headers = %v", seen)
	}
}

// TestHealthHandlersUnchangedByMiddleware compares status, headers (apart from
// X-Request-Id) and body bytes of the real health handlers with and without
// the middleware.
func TestHealthHandlersUnchangedByMiddleware(t *testing.T) {
	okProbe := func(context.Context) error { return nil }
	for _, checks := range [][]health.Check{
		{{Name: "postgres", Required: true, Probe: okProbe}},
		{{Name: "postgres", Required: true, Probe: func(context.Context) error { return context.DeadlineExceeded }}},
	} {
		h := health.Handler{ProductName: "Sample", Checks: checks}
		mux := http.NewServeMux()
		mux.HandleFunc("GET /healthz", h.Liveness)
		mux.HandleFunc("GET /readyz", h.Readiness)
		wrapped := observability.Middleware(observability.NewLogger(io.Discard, slog.LevelInfo, nil), mux)
		for _, path := range []string{"/healthz", "/readyz", "/missing"} {
			a, b := httptest.NewRecorder(), httptest.NewRecorder()
			wrapped.ServeHTTP(a, httptest.NewRequest(http.MethodGet, path, nil))
			mux.ServeHTTP(b, httptest.NewRequest(http.MethodGet, path, nil))
			if len(a.Header().Values("X-Request-Id")) != 1 {
				t.Fatalf("%s: X-Request-Id missing", path)
			}
			a.Header().Del("X-Request-Id")
			if a.Code != b.Code || !bytes.Equal(a.Body.Bytes(), b.Body.Bytes()) || !headersEqual(a.Header(), b.Header()) {
				t.Fatalf("%s differs: %d %v %q vs %d %v %q", path, a.Code, a.Header(), a.Body, b.Code, b.Header(), b.Body)
			}
		}
	}
}

func headersEqual(a, b http.Header) bool {
	if !slices.Equal(headerKeys(a), headerKeys(b)) {
		return false
	}
	for k := range a {
		if !slices.Equal(a[k], b[k]) {
			return false
		}
	}
	return true
}

// Ensure the exported context accessors used by other packages keep working
// inside a probe.
func TestProbeContextAccessors(t *testing.T) {
	var got string
	h := observability.Middleware(observability.NewLogger(io.Discard, slog.LevelInfo, nil), http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		hdr := http.Header{}
		observability.InjectOutbound(r.Context(), hdr)
		got = hdr.Get("traceparent")
		if !strings.HasSuffix(got, "-"+observability.SpanID(r.Context())+"-00") {
			t.Errorf("traceparent %q does not carry this hop's span", got)
		}
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if got == "" {
		t.Fatal("no traceparent injected")
	}
}
