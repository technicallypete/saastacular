package observability

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	hex32Re     = regexp.MustCompile(`^[0-9a-f]{32}$`)
	hex16Re     = regexp.MustCompile(`^[0-9a-f]{16}$`)
	envTimeRe   = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$`)
	intRe       = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)
	requestIDOK = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)
)

// syncBuffer is a goroutine-safe log sink.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// newCanary returns a distinctive random marker that must never be logged.
func newCanary() string { return "cnry" + randomHex(8) }

// decodeRecords parses every output line as exactly one JSON object.
func decodeRecords(t *testing.T, out string) []map[string]any {
	t.Helper()
	if out != "" && !strings.HasSuffix(out, "\n") {
		t.Fatalf("output is not newline-terminated: %q", out)
	}
	var recs []map[string]any
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		dec := json.NewDecoder(bytes.NewReader(sc.Bytes()))
		dec.UseNumber()
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("non-JSON line %q: %v", sc.Text(), err)
		}
		if dec.More() {
			t.Fatalf("more than one JSON value on a line: %q", sc.Text())
		}
		recs = append(recs, m)
	}
	return recs
}

func completionRecords(t *testing.T, out string) []map[string]any {
	t.Helper()
	var recs []map[string]any
	for _, r := range decodeRecords(t, out) {
		if r["event"] == completionEvent {
			assertCompletionShape(t, r)
			recs = append(recs, r)
		}
	}
	return recs
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func isString(v any) bool { _, ok := v.(string); return ok }

func isInt(v any) bool {
	n, ok := v.(json.Number)
	return ok && intRe.MatchString(n.String())
}

func assertEnvelope(t *testing.T, r map[string]any) {
	t.Helper()
	for _, k := range []string{"time", "level", "service", "event"} {
		if !isString(r[k]) {
			t.Fatalf("envelope key %q missing or not a string: %v", k, r)
		}
	}
	if !envTimeRe.MatchString(r["time"].(string)) {
		t.Fatalf("time is not RFC 3339 UTC with milliseconds: %v", r["time"])
	}
	if !slices.Contains([]string{"debug", "info", "warn", "error"}, r["level"].(string)) {
		t.Fatalf("level not lowercase envelope level: %v", r["level"])
	}
	if r["service"] != "core" {
		t.Fatalf("service = %v", r["service"])
	}
	if _, ok := r["msg"]; ok {
		t.Fatalf("msg key must be renamed to event: %v", r)
	}
}

// assertCompletionShape checks the exact key set, types and value domains of a
// completion record.
func assertCompletionShape(t *testing.T, r map[string]any) {
	t.Helper()
	assertEnvelope(t, r)
	want := []string{"duration_ms", "event", "level", "method", "outcome", "request_id", "request_id_source",
		"route", "service", "span_id", "status", "time", "trace_flags", "trace_id", "trace_source"}
	if _, ok := r["parent_span_id"]; ok {
		want = append(want, "parent_span_id")
	}
	if r["outcome"] == "failed" {
		want = append(want, "error_type")
	}
	sort.Strings(want)
	if got := sortedKeys(r); !slices.Equal(got, want) {
		t.Fatalf("completion keys = %v, want %v", got, want)
	}
	if !isInt(r["status"]) || !isInt(r["duration_ms"]) {
		t.Fatalf("status/duration_ms must be non-negative integers: %v", r)
	}
	for _, k := range []string{"method", "route", "request_id", "request_id_source", "trace_id", "trace_source", "span_id", "trace_flags", "outcome"} {
		if !isString(r[k]) {
			t.Fatalf("%s is not a string: %v", k, r)
		}
	}
	if !slices.Contains([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "OTHER"}, r["method"].(string)) {
		t.Fatalf("method not normalized: %v", r["method"])
	}
	if !regexp.MustCompile(`^(/[A-Za-z0-9._~/{}$-]*|unmatched)$`).MatchString(r["route"].(string)) {
		t.Fatalf("route not normalized: %v", r["route"])
	}
	if !requestIDOK.MatchString(r["request_id"].(string)) {
		t.Fatalf("request_id malformed: %v", r["request_id"])
	}
	switch r["request_id_source"] {
	case "inbound":
	case "generated":
		if !hex32Re.MatchString(r["request_id"].(string)) {
			t.Fatalf("generated request_id is not 32 hex: %v", r["request_id"])
		}
	default:
		t.Fatalf("request_id_source = %v", r["request_id_source"])
	}
	if !hex32Re.MatchString(r["trace_id"].(string)) || allZero(r["trace_id"].(string)) {
		t.Fatalf("trace_id malformed: %v", r["trace_id"])
	}
	if !hex16Re.MatchString(r["span_id"].(string)) || allZero(r["span_id"].(string)) {
		t.Fatalf("span_id malformed: %v", r["span_id"])
	}
	parent, hasParent := r["parent_span_id"]
	switch r["trace_source"] {
	case "inbound":
		if !hasParent || !isString(parent) || !hex16Re.MatchString(parent.(string)) || allZero(parent.(string)) {
			t.Fatalf("inbound trace without a valid parent_span_id: %v", r)
		}
	case "generated":
		if hasParent {
			t.Fatalf("generated trace with parent_span_id: %v", r)
		}
		if r["trace_flags"] != "00" {
			t.Fatalf("generated trace flags = %v", r["trace_flags"])
		}
	default:
		t.Fatalf("trace_source = %v", r["trace_source"])
	}
	if r["trace_flags"] != "00" && r["trace_flags"] != "01" {
		t.Fatalf("trace_flags = %v", r["trace_flags"])
	}
	status, _ := r["status"].(json.Number).Int64()
	wantLevel := "info"
	switch r["outcome"] {
	case "failed":
		if r["error_type"] != "panic" {
			t.Fatalf("failed record error_type = %v", r["error_type"])
		}
		wantLevel = "error"
	case "completed", "cancelled":
		if status >= 500 {
			wantLevel = "warn"
		}
	default:
		t.Fatalf("outcome = %v", r["outcome"])
	}
	if r["level"] != wantLevel {
		t.Fatalf("level = %v, want %s: %v", r["level"], wantLevel, r)
	}
}

func healthMux(h http.HandlerFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h)
	mux.HandleFunc("GET /readyz", h)
	return mux
}

// startServer serves handler with the server's error log captured.
func startServer(t *testing.T, handler http.Handler, errLog io.Writer) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(handler)
	srv.Config.ErrorLog = log.New(errLog, "", 0)
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

func TestParseRequestIDTable(t *testing.T) {
	secret := "sec" + newCanary()
	isSecret := func(s string) bool { return strings.Contains(s, secret) }
	for _, tc := range []struct {
		name   string
		values []string
		ok     bool
	}{
		{"valid", []string{"req-12345678"}, true},
		{"valid minimum length", []string{"abcdefgh"}, true},
		{"valid maximum length", []string{strings.Repeat("a", 64)}, true},
		{"valid punctuation", []string{"A.b_c-D.9"}, true},
		{"absent", nil, false},
		{"empty", []string{""}, false},
		{"too short", []string{"abcdefg"}, false},
		{"oversized 65", []string{strings.Repeat("a", 65)}, false},
		{"oversized 513", []string{strings.Repeat("a", 513)}, false},
		{"oversized 8 KiB", []string{strings.Repeat("a", 8192)}, false},
		{"space", []string{"bad id 12345"}, false},
		{"trailing newline", []string{"req-12345678\n"}, false},
		{"slash", []string{"req/12345678"}, false},
		{"non-ASCII", []string{"req-1234567é"}, false},
		{"duplicated field", []string{"req-12345678", "req-87654321"}, false},
		{"duplicated identical field", []string{"req-12345678", "req-12345678"}, false},
		{"comma-joined", []string{"req-12345678, req-87654321"}, false},
		{"comma-joined without space", []string{"req-12345678,req-87654321"}, false},
		{"secret-valued", []string{"req-" + secret}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id, ok := parseRequestID(tc.values, isSecret)
			if ok != tc.ok {
				t.Fatalf("parseRequestID(%q) ok = %v, want %v", tc.values, ok, tc.ok)
			}
			if ok && id != tc.values[0] {
				t.Fatalf("id = %q", id)
			}
			if !ok && id != "" {
				t.Fatalf("rejected value returned %q", id)
			}
		})
	}
}

func TestParseTraceparentProfileTable(t *testing.T) {
	tr, p := nonZeroHex(16), nonZeroHex(8)
	valid := "00-" + tr + "-" + p + "-01"
	for _, tc := range []struct {
		name   string
		values []string
		ok     bool
		flags  string
	}{
		{"valid sampled", []string{valid}, true, "01"},
		{"valid unsampled", []string{"00-" + tr + "-" + p + "-00"}, true, "00"},
		{"unknown flag bits with sampled", []string{"00-" + tr + "-" + p + "-03"}, true, "01"},
		{"unknown flag bits without sampled", []string{"00-" + tr + "-" + p + "-fe"}, true, "00"},
		{"all flag bits", []string{"00-" + tr + "-" + p + "-ff"}, true, "01"},
		{"absent", nil, false, ""},
		{"empty", []string{""}, false, ""},
		{"garbage", []string{"garbage"}, false, ""},
		{"all-zero trace", []string{"00-" + strings.Repeat("0", 32) + "-" + p + "-01"}, false, ""},
		{"all-zero parent", []string{"00-" + tr + "-" + strings.Repeat("0", 16) + "-01"}, false, ""},
		{"uppercase trace", []string{"00-" + strings.ToUpper(tr) + "-" + p + "-01"}, false, ""},
		{"uppercase flags", []string{"00-" + tr + "-" + p + "-0A"}, false, ""},
		{"version 01", []string{"01-" + tr + "-" + p + "-01"}, false, ""},
		{"version ff", []string{"ff-" + tr + "-" + p + "-01"}, false, ""},
		{"version cc", []string{"cc-" + tr + "-" + p + "-01"}, false, ""},
		{"version 01 with extra field", []string{"01-" + tr + "-" + p + "-01-abcd"}, false, ""},
		{"extra field on 00", []string{valid + "-abcd"}, false, ""},
		{"short trace", []string{"00-" + tr[:31] + "-" + p + "-01"}, false, ""},
		{"long parent", []string{"00-" + tr + "-" + p + "0-01"}, false, ""},
		{"non-hex", []string{"00-" + tr[:31] + "g-" + p + "-01"}, false, ""},
		{"internal whitespace", []string{"00-" + tr + " -" + p + "-01"}, false, ""},
		{"leading whitespace", []string{" " + valid}, false, ""},
		{"trailing whitespace", []string{valid + "\t"}, false, ""},
		{"trailing newline", []string{valid + "\n"}, false, ""},
		{"duplicated field", []string{valid, valid}, false, ""},
		{"comma-joined", []string{valid + "," + valid}, false, ""},
		{"comma-joined with space", []string{valid + ", " + valid}, false, ""},
		{"oversized 513", []string{valid + strings.Repeat("0", 513-len(valid))}, false, ""},
		{"oversized 8 KiB", []string{valid + strings.Repeat("0", 8192-len(valid))}, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotTrace, gotParent, flags, ok := parseTraceparent(tc.values)
			if ok != tc.ok {
				t.Fatalf("parseTraceparent(%q) ok = %v, want %v", tc.values, ok, tc.ok)
			}
			if ok && (gotTrace != tr || gotParent != p || flags != tc.flags) {
				t.Fatalf("got %q %q %q, want %q %q %q", gotTrace, gotParent, flags, tr, p, tc.flags)
			}
			if !ok && (gotTrace != "" || gotParent != "" || flags != "") {
				t.Fatalf("rejected value returned data: %q %q %q", gotTrace, gotParent, flags)
			}
			if len(tc.values) == 1 {
				if id, ok := ParseTraceparent(tc.values[0]); ok != tc.ok || (ok && id != tr) {
					t.Fatalf("ParseTraceparent(%q) = %q, %v", tc.values[0], id, ok)
				}
			}
		})
	}
}

func TestGeneratedContextIsWellFormedAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for range 500 {
		tc := newTraceContext(http.Header{}, nil)
		if !hex32Re.MatchString(tc.RequestID) || tc.RequestIDSource != "generated" {
			t.Fatalf("request ID = %q (%s)", tc.RequestID, tc.RequestIDSource)
		}
		if !hex32Re.MatchString(tc.TraceID) || allZero(tc.TraceID) || tc.TraceSource != "generated" {
			t.Fatalf("trace ID = %q (%s)", tc.TraceID, tc.TraceSource)
		}
		if !hex16Re.MatchString(tc.SpanID) || allZero(tc.SpanID) {
			t.Fatalf("span ID = %q", tc.SpanID)
		}
		if tc.ParentSpanID != "" || tc.Flags != "00" {
			t.Fatalf("generated context has parent %q flags %q", tc.ParentSpanID, tc.Flags)
		}
		for _, v := range []string{tc.RequestID, tc.TraceID, tc.SpanID} {
			if seen[v] {
				t.Fatalf("identifier repeated: %s", v)
			}
			seen[v] = true
		}
	}
}

// TestMiddlewareContextTable drives every accepted and rejected inbound form
// through the middleware and checks the context, the response header and the
// completion record. A rejected raw value never appears in any output.
func TestMiddlewareContextTable(t *testing.T) {
	secret := "sec" + newCanary()
	type row struct {
		name        string
		reqIDs      []string
		traceparent []string
		wantReqID   string // "" means generated
		wantTrace   string // "" means generated
		wantParent  string
		wantFlags   string
		absent      []string // raw markers that must never appear
	}
	var rows []row
	add := func(r row) { rows = append(rows, r) }

	tr, p := nonZeroHex(16), nonZeroHex(8)
	add(row{name: "valid inbound context", reqIDs: []string{"req-valid-0001"}, traceparent: []string{"00-" + tr + "-" + p + "-01"},
		wantReqID: "req-valid-0001", wantTrace: tr, wantParent: p, wantFlags: "01"})
	tr, p = nonZeroHex(16), nonZeroHex(8)
	add(row{name: "unknown flag bits are not kept", traceparent: []string{"00-" + tr + "-" + p + "-fe"},
		wantTrace: tr, wantParent: p, wantFlags: "00"})
	add(row{name: "absent"})

	for _, c := range []struct{ name, value string }{
		{"malformed request ID", "bad id " + newCanary()},
		{"request ID with newline", "req-" + newCanary() + "\n"},
		{"oversized request ID 65", padTo(newCanary(), 65)},
		{"oversized request ID 513", padTo(newCanary(), 513)},
		{"oversized request ID 8 KiB", padTo(newCanary(), 8192)},
		{"comma-joined request IDs", "req-" + newCanary() + ", req-" + newCanary()},
		{"secret-valued request ID", "req-" + secret},
	} {
		marker := c.value
		if i := strings.Index(c.value, "cnry"); i >= 0 {
			marker = c.value[i : i+20]
		}
		add(row{name: c.name, reqIDs: []string{c.value}, absent: []string{marker}})
	}
	dupA, dupB := "req-"+newCanary(), "req-"+newCanary()
	add(row{name: "duplicated request ID field", reqIDs: []string{dupA, dupB}, absent: []string{dupA, dupB}})

	for _, c := range []struct {
		name string
		mk   func(tr, p string) []string
	}{
		{"all-zero trace", func(_, p string) []string { return []string{"00-" + strings.Repeat("0", 32) + "-" + p + "-01"} }},
		{"all-zero parent", func(tr, _ string) []string { return []string{"00-" + tr + "-" + strings.Repeat("0", 16) + "-01"} }},
		{"uppercase", func(tr, p string) []string { return []string{"00-" + strings.ToUpper(tr) + "-" + p + "-01"} }},
		{"version 01", func(tr, p string) []string { return []string{"01-" + tr + "-" + p + "-01"} }},
		{"version ff", func(tr, p string) []string { return []string{"ff-" + tr + "-" + p + "-01"} }},
		{"extra fields", func(tr, p string) []string { return []string{"00-" + tr + "-" + p + "-01-" + newCanary()} }},
		{"internal whitespace", func(tr, p string) []string { return []string{"00-" + tr + "- " + p + "-01"} }},
		{"duplicated field", func(tr, p string) []string {
			return []string{"00-" + tr + "-" + p + "-01", "00-" + tr + "-" + p + "-01"}
		}},
		{"comma-joined", func(tr, p string) []string { return []string{"00-" + tr + "-" + p + "-01, 00-" + tr + "-" + p + "-01"} }},
		{"oversized 513", func(tr, p string) []string { return []string{padTo("00-"+tr+"-"+p+"-01", 513)} }},
		{"oversized 8 KiB", func(tr, p string) []string { return []string{padTo("00-"+tr+"-"+p+"-01", 8192)} }},
	} {
		tr, p := nonZeroHex(16), nonZeroHex(8)
		add(row{name: "traceparent " + c.name, traceparent: c.mk(tr, p),
			absent: []string{tr, strings.ToUpper(tr), p}})
	}

	for _, rw := range rows {
		t.Run(rw.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := NewLogger(&logs, slog.LevelInfo, []string{secret})
			var got TraceContext
			var gotOK bool
			h := Middleware(logger, healthMux(func(w http.ResponseWriter, r *http.Request) {
				got, gotOK = FromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/healthz?token=query-value", nil)
			for _, v := range rw.reqIDs {
				req.Header.Add("X-Request-Id", v)
			}
			for _, v := range rw.traceparent {
				req.Header.Add("traceparent", v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if !gotOK {
				t.Fatal("no context stored")
			}
			if rec.Header().Get("X-Request-Id") != got.RequestID || len(rec.Header().Values("X-Request-Id")) != 1 {
				t.Fatalf("response X-Request-Id = %q, context %q", rec.Header().Values("X-Request-Id"), got.RequestID)
			}
			if rw.wantReqID != "" {
				if got.RequestID != rw.wantReqID || got.RequestIDSource != "inbound" {
					t.Fatalf("request ID = %q (%s), want %q", got.RequestID, got.RequestIDSource, rw.wantReqID)
				}
			} else if !hex32Re.MatchString(got.RequestID) || got.RequestIDSource != "generated" {
				t.Fatalf("request ID not generated: %q (%s)", got.RequestID, got.RequestIDSource)
			}
			if rw.wantTrace != "" {
				if got.TraceID != rw.wantTrace || got.TraceSource != "inbound" || got.ParentSpanID != rw.wantParent || got.Flags != rw.wantFlags {
					t.Fatalf("trace context = %+v", got)
				}
				if got.SpanID == rw.wantParent {
					t.Fatal("span ID copied from the inbound parent")
				}
			} else if got.TraceSource != "generated" || got.ParentSpanID != "" || got.Flags != "00" || !hex32Re.MatchString(got.TraceID) {
				t.Fatalf("trace context not generated: %+v", got)
			}
			if !hex16Re.MatchString(got.SpanID) || allZero(got.SpanID) {
				t.Fatalf("span ID = %q", got.SpanID)
			}

			recs := completionRecords(t, logs.String())
			if len(recs) != 1 {
				t.Fatalf("want exactly one completion record, got %d: %s", len(recs), logs.String())
			}
			r := recs[0]
			if r["request_id"] != got.RequestID || r["request_id_source"] != got.RequestIDSource ||
				r["trace_id"] != got.TraceID || r["trace_source"] != got.TraceSource ||
				r["span_id"] != got.SpanID || r["trace_flags"] != got.Flags {
				t.Fatalf("record does not match context %+v: %v", got, r)
			}
			if got.ParentSpanID != "" && r["parent_span_id"] != got.ParentSpanID {
				t.Fatalf("parent_span_id = %v", r["parent_span_id"])
			}
			if r["route"] != "/healthz" || r["method"] != "GET" || r["outcome"] != "completed" {
				t.Fatalf("unexpected record: %v", r)
			}
			out := logs.String() + fmt.Sprint(rec.Header())
			for _, m := range append(rw.absent, "query-value", secret) {
				if strings.Contains(out, m) {
					t.Fatalf("raw value %q reflected or logged: %s", m, out)
				}
			}
		})
	}
}

func padTo(s string, n int) string { return s + strings.Repeat("a", n-len(s)) }

// TestMiddlewareConcurrentRequestsDoNotShareContext sends 64 concurrent
// requests, half with inbound context and half without, and checks that each
// handler, response and record sees only its own context.
func TestMiddlewareConcurrentRequestsDoNotShareContext(t *testing.T) {
	logs := &syncBuffer{}
	errLog := &syncBuffer{}
	logger := NewLogger(logs, slog.LevelInfo, nil)
	srv := startServer(t, Middleware(logger, healthMux(func(w http.ResponseWriter, r *http.Request) {
		before, _ := FromContext(r.Context())
		time.Sleep(time.Duration(rand.IntN(15)) * time.Millisecond)
		after, _ := FromContext(r.Context())
		if before != after {
			http.Error(w, "context changed", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s %s %s %s", after.RequestID, after.TraceID, after.ParentSpanID, after.SpanID)
	})), errLog)

	const n = 64
	type sent struct{ reqID, trace, parent, respID, body string }
	results := make([]sent, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/readyz", nil)
			if err != nil {
				t.Error(err)
				return
			}
			var s sent
			if i%2 == 0 {
				s.reqID, s.trace, s.parent = fmt.Sprintf("req-concurrent-%03d", i), nonZeroHex(16), nonZeroHex(8)
				req.Header.Set("X-Request-Id", s.reqID)
				req.Header.Set("traceparent", "00-"+s.trace+"-"+s.parent+"-01")
			}
			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("status %d: %s", resp.StatusCode, body)
			}
			s.respID, s.body = resp.Header.Get("X-Request-Id"), string(body)
			results[i] = s
		}(i)
	}
	close(start)
	wg.Wait()
	srv.Close()
	if t.Failed() {
		t.FailNow()
	}

	byID := map[string]map[string]any{}
	for _, r := range completionRecords(t, logs.String()) {
		id := r["request_id"].(string)
		if byID[id] != nil {
			t.Fatalf("request ID %s logged twice", id)
		}
		byID[id] = r
	}
	if len(byID) != n {
		t.Fatalf("want %d records, got %d", n, len(byID))
	}
	spans := map[string]bool{}
	for i, s := range results {
		f := strings.Split(s.body, " ")
		if len(f) != 4 {
			t.Fatalf("request %d body %q", i, s.body)
		}
		if f[0] != s.respID {
			t.Fatalf("request %d: handler saw %s, response header %s", i, f[0], s.respID)
		}
		if s.reqID != "" && (f[0] != s.reqID || f[1] != s.trace || f[2] != s.parent) {
			t.Fatalf("request %d saw another request's context: %q", i, s.body)
		}
		if s.reqID == "" && (!hex32Re.MatchString(f[0]) || f[2] != "") {
			t.Fatalf("request %d: generated context expected: %q", i, s.body)
		}
		r := byID[s.respID]
		if r == nil || r["trace_id"] != f[1] || r["span_id"] != f[3] {
			t.Fatalf("request %d: record mismatch %v", i, r)
		}
		if spans[f[3]] {
			t.Fatalf("span %s reused", f[3])
		}
		spans[f[3]] = true
	}
	if errLog.String() != "" {
		t.Fatalf("server error log not empty: %s", errLog.String())
	}
}

func TestMiddlewareCancelledRequestLogsCancelledAndKeepsNoState(t *testing.T) {
	logs := &syncBuffer{}
	errLog := &syncBuffer{}
	logger := NewLogger(logs, slog.LevelInfo, nil)
	entered := make(chan struct{})
	srv := startServer(t, Middleware(logger, healthMux(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" {
			close(entered)
			<-r.Context().Done()
			return
		}
		tc, _ := FromContext(r.Context())
		fmt.Fprintf(w, "%s %s %s", tc.RequestID, tc.RequestIDSource, tc.ParentSpanID)
	})), errLog)

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/readyz", nil)
	req.Header.Set("X-Request-Id", "req-cancelled-0001")
	req.Header.Set("traceparent", "00-"+nonZeroHex(16)+"-"+nonZeroHex(8)+"-01")
	done := make(chan error, 1)
	go func() {
		resp, err := srv.Client().Do(req)
		if resp != nil {
			_ = resp.Body.Close()
		}
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("handler not reached")
	}
	cancel()
	if err := <-done; err == nil {
		t.Fatal("cancelled request succeeded")
	}
	deadline := time.Now().Add(5 * time.Second)
	for len(completionRecords(t, logs.String())) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	// A following request without context gets a fresh, generated context.
	resp, err := srv.Client().Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	f := strings.Split(string(body), " ")
	if len(f) != 3 || f[0] == "req-cancelled-0001" || f[1] != "generated" || f[2] != "" {
		t.Fatalf("state leaked into the next request: %q", body)
	}
	srv.Close()

	recs := completionRecords(t, logs.String())
	if len(recs) != 2 {
		t.Fatalf("want 2 records, got %d: %s", len(recs), logs.String())
	}
	var cancelled map[string]any
	for _, r := range recs {
		if r["request_id"] == "req-cancelled-0001" {
			cancelled = r
		}
	}
	if cancelled == nil || cancelled["outcome"] != "cancelled" || cancelled["route"] != "/readyz" ||
		cancelled["status"].(json.Number).String() != "0" || cancelled["level"] != "info" {
		t.Fatalf("cancelled record = %v", cancelled)
	}
	if errLog.String() != "" {
		t.Fatalf("server error log not empty: %s", errLog.String())
	}
}

func TestMiddlewarePanicLogsOnceAndAbortsSilently(t *testing.T) {
	for _, tc := range []struct {
		name       string
		h          http.HandlerFunc
		wantStatus string
	}{
		{"before headers", func(http.ResponseWriter, *http.Request) { panic("boom") }, "0"},
		{"after 503", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			panic("boom")
		}, "503"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cn := newCanary()
			logs := &syncBuffer{}
			errLog := &syncBuffer{}
			logger := NewLogger(logs, slog.LevelInfo, nil)
			srv := startServer(t, Middleware(logger, healthMux(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if p := recover(); p != nil {
						panic(fmt.Sprintf("%v %s", p, cn))
					}
				}()
				tc.h(w, r)
			})), errLog)
			resp, err := srv.Client().Get(srv.URL + "/healthz")
			if err == nil {
				_ = resp.Body.Close()
				if tc.wantStatus == "0" {
					t.Fatalf("panicking handler produced a response: %d", resp.StatusCode)
				}
			}
			srv.Close()

			recs := completionRecords(t, logs.String())
			if len(recs) != 1 {
				t.Fatalf("want exactly one record, got %d: %s", len(recs), logs.String())
			}
			r := recs[0]
			if r["outcome"] != "failed" || r["error_type"] != "panic" || r["level"] != "error" ||
				r["status"].(json.Number).String() != tc.wantStatus || r["route"] != "/healthz" {
				t.Fatalf("panic record = %v", r)
			}
			if errLog.String() != "" {
				t.Fatalf("server error log not empty (ErrAbortHandler must be silent): %q", errLog.String())
			}
			if strings.Contains(logs.String(), cn) {
				t.Fatalf("panic value logged: %s", logs.String())
			}
		})
	}

	// Without a server, the middleware re-panics with http.ErrAbortHandler.
	var logs bytes.Buffer
	h := Middleware(NewLogger(&logs, slog.LevelInfo, nil), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(errors.New("original"))
	}))
	func() {
		defer func() {
			if p := recover(); p != http.ErrAbortHandler {
				t.Fatalf("re-panic value = %v", p)
			}
		}()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
	}()
	if recs := completionRecords(t, logs.String()); len(recs) != 1 || recs[0]["route"] != "unmatched" {
		t.Fatalf("records = %v", recs)
	}
}

// TestMiddlewareStatusAndResponseBehaviour checks the recorded status against
// the status a client observes with and without the middleware.
func TestMiddlewareStatusAndResponseBehaviour(t *testing.T) {
	for _, tc := range []struct {
		name       string
		h          http.HandlerFunc
		wantStatus int
	}{
		{"implicit 200 on first write", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }, 200},
		{"implicit 200 without a write", func(http.ResponseWriter, *http.Request) {}, 200},
		{"explicit 503", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("unready"))
		}, 503},
		{"duplicate WriteHeader keeps the first", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusInternalServerError)
		}, 201},
		{"WriteHeader after a write", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
			w.WriteHeader(http.StatusInternalServerError)
		}, 200},
		{"informational before final", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusEarlyHints)
			w.WriteHeader(http.StatusAccepted)
		}, 202},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs := &syncBuffer{}
			logger := NewLogger(logs, slog.LevelInfo, nil)
			with := startServer(t, Middleware(logger, healthMux(tc.h)), io.Discard)
			without := startServer(t, healthMux(tc.h), io.Discard)
			observed := func(url string) (int, string) {
				resp, err := http.Get(url + "/healthz")
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				b, _ := io.ReadAll(resp.Body)
				return resp.StatusCode, string(b)
			}
			sWith, bWith := observed(with.URL)
			sWithout, bWithout := observed(without.URL)
			if sWith != sWithout || bWith != bWithout || sWith != tc.wantStatus {
				t.Fatalf("with middleware %d %q, without %d %q, want status %d", sWith, bWith, sWithout, bWithout, tc.wantStatus)
			}
			with.Close()
			recs := completionRecords(t, logs.String())
			if len(recs) != 1 || recs[0]["status"].(json.Number).String() != fmt.Sprint(tc.wantStatus) {
				t.Fatalf("records = %v", recs)
			}
		})
	}
}

func TestRouteAndMethodNormalization(t *testing.T) {
	cn := newCanary()
	var logs bytes.Buffer
	h := Middleware(NewLogger(&logs, slog.LevelInfo, nil), healthMux(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for _, tc := range []struct {
		method, target    string
		wantCode          int
		wantRoute, wantMt string
	}{
		{http.MethodGet, "/healthz", 200, "/healthz", "GET"},
		{http.MethodGet, "/readyz?token=" + cn, 200, "/readyz", "GET"},
		{http.MethodHead, "/healthz", 200, "/healthz", "HEAD"},
		{http.MethodGet, "/missing-" + cn, 404, "unmatched", "GET"},
		{http.MethodPost, "/healthz", 405, "unmatched", "POST"},
		{http.MethodOptions, "/readyz", 405, "unmatched", "OPTIONS"},
		{"PROPFIND", "/healthz", 405, "unmatched", "OTHER"},
		{"TRACE", "/" + cn, 404, "unmatched", "OTHER"},
	} {
		logs.Reset()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.target, nil))
		recs := completionRecords(t, logs.String())
		if rec.Code != tc.wantCode || len(recs) != 1 || recs[0]["route"] != tc.wantRoute || recs[0]["method"] != tc.wantMt {
			t.Fatalf("%s %s: code %d, records %v", tc.method, tc.target, rec.Code, recs)
		}
		if strings.Contains(logs.String(), cn) || strings.Contains(logs.String(), "missing") {
			t.Fatalf("raw path or query logged: %s", logs.String())
		}
	}
	for pattern, want := range map[string]string{
		"GET /healthz":           "/healthz",
		"GET /readyz":            "/readyz",
		"GET /items/{id}":        "/items/{id}",
		"":                       "unmatched",
		"/healthz":               "unmatched", // a raw path stored by a mux redirect
		"/tree/":                 "unmatched",
		"GET example.test/x":     "unmatched",
		"GET /a b":               "unmatched",
		"GET /" + padTo("x", 70): "unmatched",
	} {
		if got := routeOf(pattern); got != want {
			t.Errorf("routeOf(%q) = %q, want %q", pattern, got, want)
		}
	}
}

func TestCompletionRecordKeysAndTypes(t *testing.T) {
	var logs bytes.Buffer
	h := Middleware(NewLogger(&logs, slog.LevelInfo, nil), healthMux(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{}"))
	}))
	// With an inbound parent: parent_span_id is present.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("traceparent", "00-"+nonZeroHex(16)+"-"+nonZeroHex(8)+"-01")
	h.ServeHTTP(httptest.NewRecorder(), req)
	// Without: parent_span_id is absent.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/readyz", nil))

	recs := decodeRecords(t, logs.String())
	if len(recs) != 2 {
		t.Fatalf("want 2 records: %s", logs.String())
	}
	for i, r := range recs {
		if r["event"] != "http.request.completed" {
			t.Fatalf("event = %v", r["event"])
		}
		assertCompletionShape(t, r)
		_, hasParent := r["parent_span_id"]
		if hasParent != (i == 0) || r["level"] != "info" || r["status"].(json.Number).String() != "200" {
			t.Fatalf("record %d = %v", i, r)
		}
	}
	if recs[0]["trace_flags"] != "01" || recs[1]["trace_flags"] != "00" {
		t.Fatalf("trace_flags = %v, %v", recs[0]["trace_flags"], recs[1]["trace_flags"])
	}
}

func TestEnvelopeOnOtherRecords(t *testing.T) {
	var logs bytes.Buffer
	logger := NewLogger(&logs, slog.LevelDebug, nil)
	logger.Debug("d")
	logger.Info("core listening", slog.Int("port", 8180))
	logger.Warn("w")
	logger.Error("e")
	recs := decodeRecords(t, logs.String())
	want := []string{"debug", "info", "warn", "error"}
	if len(recs) != len(want) {
		t.Fatalf("records: %s", logs.String())
	}
	for i, r := range recs {
		assertEnvelope(t, r)
		if r["level"] != want[i] {
			t.Fatalf("level = %v, want %s", r["level"], want[i])
		}
	}
	if recs[1]["event"] != "core listening" || recs[1]["port"].(json.Number).String() != "8180" {
		t.Fatalf("record = %v", recs[1])
	}
}

// TestHandlerOutputByteIdentical compares the raw response bytes of a handler
// with and without the middleware. Apart from X-Request-Id (and Date, which
// varies per response), they are identical.
func TestHandlerOutputByteIdentical(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"json 200": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"alive"}` + "\n"))
		},
		"json 503": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unready"}` + "\n"))
		},
		"implicit 200": func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("plain")) },
		"empty":        func(http.ResponseWriter, *http.Request) {},
	}
	for name, hf := range handlers {
		t.Run(name, func(t *testing.T) {
			with := startServer(t, Middleware(NewLogger(io.Discard, slog.LevelInfo, nil), healthMux(hf)), io.Discard)
			without := startServer(t, healthMux(hf), io.Discard)
			a, idLines := rawResponse(t, with.URL)
			b, noIDLines := rawResponse(t, without.URL)
			if idLines != 1 || noIDLines != 0 {
				t.Fatalf("X-Request-Id lines: with %d, without %d", idLines, noIDLines)
			}
			if !bytes.Equal(a, b) {
				t.Fatalf("response bytes differ:\n%q\n%q", a, b)
			}
		})
	}
}

// rawResponse returns the raw HTTP/1.1 response with the Date and
// X-Request-Id lines removed, and the number of X-Request-Id lines.
func rawResponse(t *testing.T, url string) ([]byte, int) {
	t.Helper()
	conn, err := net.Dial("tcp", strings.TrimPrefix(url, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := io.WriteString(conn, "GET /healthz HTTP/1.1\r\nHost: test\r\nConnection: close\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}
	var out []byte
	ids := 0
	for _, line := range bytes.SplitAfter(raw, []byte("\r\n")) {
		switch {
		case bytes.HasPrefix(line, []byte("Date: ")):
		case bytes.HasPrefix(line, []byte("X-Request-Id: ")):
			ids++
		default:
			out = append(out, line...)
		}
	}
	return out, ids
}

func TestInjectOutboundSetsExactlyTwoHeaders(t *testing.T) {
	h := http.Header{}
	InjectOutbound(context.Background(), h)
	if len(h) != 0 {
		t.Fatalf("headers set without a context: %v", h)
	}
	tc := TraceContext{RequestID: "req-12345678", RequestIDSource: "inbound", TraceID: nonZeroHex(16), TraceSource: "inbound",
		SpanID: nonZeroHex(8), ParentSpanID: nonZeroHex(8), Flags: "01"}
	ctx := context.WithValue(context.Background(), traceContextKey, tc)
	InjectOutbound(ctx, h)
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if !slices.Equal(keys, []string{"Traceparent", "X-Request-Id"}) {
		t.Fatalf("outbound header keys = %v", keys)
	}
	if h.Get("X-Request-Id") != tc.RequestID || h.Get("traceparent") != "00-"+tc.TraceID+"-"+tc.SpanID+"-01" {
		t.Fatalf("outbound headers = %v", h)
	}
	if RequestID(ctx) != tc.RequestID || TraceID(ctx) != tc.TraceID || SpanID(ctx) != tc.SpanID {
		t.Fatal("context accessors disagree")
	}
}
