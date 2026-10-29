package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

const canary = "Cx7d1e2f3a4b5c6d70"

type seen struct {
	mu       sync.Mutex
	requests []*http.Request
	bodies   [][]byte
}

func (s *seen) add(r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, r)
	s.bodies = append(s.bodies, b)
}

func server(t *testing.T, s *seen, h func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s != nil {
			s.add(r)
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func runCLI(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func one(t *testing.T, out string) Result {
	t.Helper()
	var r Result
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("output is not one JSON object: %q", out)
	}
	return r
}

func write(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestHeadersBodyAndOutput(t *testing.T) {
	s := &seen{}
	srv := server(t, s, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "req-response-01")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, `{"status":"unready"}`)
	})
	hf := write(t, "auth", "Bearer "+canary+"\n")
	bf := write(t, "body", "payload-"+canary)
	code, out, errOut := runCLI(t, "--url", srv.URL+"/api/ready", "--method", "POST",
		"--header", "X-Request-Id: first-id-0001", "--header", "X-Request-Id:second-id-0002",
		"--header", "traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"--header-file", "Authorization="+hf, "--body-file", bf)
	if code != exitPass {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if strings.Contains(out+errOut, canary) {
		t.Fatalf("a value read from a file was printed")
	}
	r := one(t, out)
	want := sha256.Sum256([]byte(`{"status":"unready"}`))
	if r.Status != 503 || r.XRequestID == nil || *r.XRequestID != "req-response-01" || r.XRequestIDCount != 1 ||
		r.CacheControl == nil || *r.CacheControl != "no-store" || r.Body == nil || *r.Body != `{"status":"unready"}` ||
		r.BodySHA256 != hex.EncodeToString(want[:]) || r.BodyBytes != 20 || r.Error != "" || r.Sent != nil || r.DurationMS < 0 {
		t.Errorf("result: %s", out)
	}
	got := s.requests[0]
	if ids := got.Header.Values("X-Request-Id"); len(ids) != 2 || ids[0] != "first-id-0001" || ids[1] != "second-id-0002" {
		t.Errorf("duplicate headers must arrive as separate field lines: %q", ids)
	}
	if got.Header.Get("Authorization") != "Bearer "+canary || string(s.bodies[0]) != "payload-"+canary || got.Method != "POST" {
		t.Errorf("header file or body not sent")
	}
	if got.Header.Get("Traceparent") != "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" {
		t.Errorf("traceparent not sent")
	}
}

func TestURLFileIsNeverPrinted(t *testing.T) {
	s := &seen{}
	srv := server(t, s, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok") })
	uf := write(t, "url", srv.URL+"/api/health?token="+canary+"&q="+canary+"\n")
	out := filepath.Join(t.TempDir(), "out.json")
	code, stdout, errOut := runCLI(t, "--url-file", uf, "--out", out)
	b, _ := os.ReadFile(out)
	if code != exitPass || stdout != "" || strings.Contains(string(b)+errOut, canary) || strings.Contains(string(b), "/api/health") {
		t.Errorf("url-file output: %d %q %q", code, b, errOut)
	}
	if s.requests[0].URL.Query().Get("token") != canary {
		t.Errorf("the URL from the file was not used")
	}
	// An invalid URL is refused without echoing it.
	bad := write(t, "bad", "ftp://x/"+canary)
	if code, o, e := runCLI(t, "--url-file", bad); code != exitUsage || strings.Contains(o+e, canary) {
		t.Errorf("invalid url: %d", code)
	}
}

func TestParallelGeneratesDistinctContexts(t *testing.T) {
	s := &seen{}
	srv := server(t, s, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", r.Header.Get("X-Request-Id"))
	})
	code, out, _ := runCLI(t, "--url", srv.URL, "--parallel", "8")
	if code != exitPass {
		t.Fatalf("exit %d", code)
	}
	var results []Result
	if err := json.Unmarshal([]byte(out), &results); err != nil || len(results) != 8 {
		t.Fatalf("parallel output must be an array of 8: %q", out)
	}
	tpRe := regexp.MustCompile(`^00-[0-9a-f]{32}-[0-9a-f]{16}-01$`)
	ids, tps := map[string]bool{}, map[string]bool{}
	for _, r := range results {
		if r.Sent == nil || !requestIDRe.MatchString(r.Sent.XRequestID) || len(r.Sent.XRequestID) != 32 || !tpRe.MatchString(r.Sent.Traceparent) {
			t.Fatalf("generated context: %+v", r.Sent)
		}
		if r.XRequestID == nil || *r.XRequestID != r.Sent.XRequestID {
			t.Errorf("response ID does not match the generated one")
		}
		ids[r.Sent.XRequestID], tps[r.Sent.Traceparent] = true, true
	}
	if len(ids) != 8 || len(tps) != 8 || len(s.requests) != 8 {
		t.Errorf("contexts are not distinct: %d ids, %d traceparents, %d requests", len(ids), len(tps), len(s.requests))
	}
	// Supplied headers are not replaced or echoed.
	code, out, _ = runCLI(t, "--url", srv.URL, "--parallel", "2", "--header", "X-Request-Id: supplied-0001")
	var two []Result
	if err := json.Unmarshal([]byte(out), &two); err != nil || code != exitPass || len(two) != 2 {
		t.Fatalf("parallel 2: %q", out)
	}
	for _, r := range two {
		if r.Sent == nil || r.Sent.XRequestID != "" || r.Sent.Traceparent == "" {
			t.Errorf("only the missing header is generated: %+v", r.Sent)
		}
	}
	// Without --parallel nothing is generated.
	_, out, _ = runCLI(t, "--url", srv.URL)
	last := s.requests[len(s.requests)-1]
	if one(t, out).Sent != nil || last.Header.Get("X-Request-Id") != "" || last.Header.Get("Traceparent") != "" {
		t.Errorf("a single request must carry no generated context")
	}
}

func TestAbortTimeoutAndErrors(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	slow := server(t, nil, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		case <-time.After(5 * time.Second):
		}
	})
	code, out, _ := runCLI(t, "--url", slow.URL, "--abort-after", "100ms")
	if r := one(t, out); code != exitPass || r.Error != "aborted" || r.Status != 0 || r.DurationMS < 50 || r.DurationMS > 3000 {
		t.Errorf("abort: %s", out)
	}
	code, out, _ = runCLI(t, "--url", slow.URL, "--timeout", "150ms")
	if r := one(t, out); code != exitPass || r.Error != "timeout" || r.Status != 0 {
		t.Errorf("timeout: %s", out)
	}
	// Abort while the body is streaming still reports aborted with status 0.
	streaming := server(t, nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		io.WriteString(w, "partial")
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	})
	if _, out, _ := runCLI(t, "--url", streaming.URL, "--abort-after", "100ms"); one(t, out).Error != "aborted" || one(t, out).Status != 0 {
		t.Errorf("abort during body: %s", out)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	if _, out, _ := runCLI(t, "--url", "http://"+addr+"/"); one(t, out).Error != "refused" {
		t.Errorf("refused: %s", out)
	}
	ok := server(t, nil, func(w http.ResponseWriter, r *http.Request) {})
	hf := write(t, "bad", "a\x00b")
	if _, out, _ := runCLI(t, "--url", ok.URL, "--header-file", "X-Bad="+hf); one(t, out).Error != "invalid_request" {
		t.Errorf("control character value: %s", out)
	}
}

func TestBodyRendering(t *testing.T) {
	for name, tc := range map[string]struct {
		body  string
		shown bool
	}{
		"small utf-8":  {"héllo", true},
		"exactly 4096": {strings.Repeat("a", 4096), true},
		"too large":    {strings.Repeat("a", 4097), false},
		"not utf-8":    {"\xff\xfe", false},
		"empty":        {"", true},
	} {
		srv := server(t, nil, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Request-Id", "bad id with spaces")
			w.Header().Add("Cache-Control", strings.Repeat("a", 257))
			io.WriteString(w, tc.body)
		})
		_, out, _ := runCLI(t, "--url", srv.URL)
		r := one(t, out)
		sum := sha256.Sum256([]byte(tc.body))
		if (r.Body != nil) != tc.shown || r.BodyBytes != len(tc.body) || r.BodySHA256 != hex.EncodeToString(sum[:]) {
			t.Errorf("%s: %s", name, out)
		}
		if r.XRequestID != nil || r.XRequestIDCount != 1 || r.CacheControl != nil {
			t.Errorf("%s: invalid response header values must not be printed: %s", name, out)
		}
	}
}

func TestUsage(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"--url", "http://127.0.0.1:1", "--url-file", "/x"},
		{"--url", "http://127.0.0.1:1", "--parallel", "0"},
		{"--url", "http://127.0.0.1:1", "--timeout", "0s"},
		{"--url", "http://127.0.0.1:1", "--header", "no-colon"},
		{"--url", "http://127.0.0.1:1", "--header", "Bad Name: x"},
		{"--url", "http://127.0.0.1:1", "--header-file", "X-A"},
		{"--url", "http://127.0.0.1:1", "--header-file", "X-A=/does/not/exist"},
		{"--url", "http://127.0.0.1:1", "--body-file", "/does/not/exist"},
		{"--url", "http://127.0.0.1:1", "--method", "GE T"},
		{"--url", "not a url"},
		{"--url", "http://127.0.0.1:1", "extra"},
	} {
		if code, _, _ := runCLI(t, args...); code != exitUsage {
			t.Errorf("%v: exit %d, want %d", args, code, exitUsage)
		}
	}
}
