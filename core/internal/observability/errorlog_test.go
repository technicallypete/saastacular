package observability

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"sort"
	"strings"
	"testing"
)

func serverErrorRecords(t *testing.T, out string) []string {
	t.Helper()
	var types []string
	for _, r := range decodeRecords(t, out) {
		assertEnvelope(t, r)
		if r["event"] != "http.server.error" || r["level"] != "warn" {
			t.Fatalf("unexpected record: %v", r)
		}
		if got := sortedKeys(r); !slices.Equal(got, []string{"error_type", "event", "level", "service", "time"}) {
			t.Fatalf("keys = %v", got)
		}
		types = append(types, r["error_type"].(string))
	}
	return types
}

// TestServerErrorLogWithRealServer installs ServerErrorLog on a real server
// whose handlers write a superfluous header, panic with a canary value and
// panic with http.ErrAbortHandler.
func TestServerErrorLogWithRealServer(t *testing.T) {
	cn := newCanary()
	logs := &syncBuffer{}
	logger := NewLogger(logs, slog.LevelInfo, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /double", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("GET /panic", func(http.ResponseWriter, *http.Request) { panic("value " + cn) })
	mux.HandleFunc("GET /abort", func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) })
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.ErrorLog = ServerErrorLog(logger)
	srv.Start()
	defer srv.Close()

	// A fresh connection per request: the transport never retries on a reused one.
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	for _, path := range []string{"/double", "/panic", "/abort"} {
		resp, err := client.Get(srv.URL + path)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
	}
	srv.Close()

	out := logs.String()
	types := serverErrorRecords(t, out)
	sort.Strings(types)
	if !slices.Equal(types, []string{"handler_panic", "superfluous_write_header"}) {
		t.Fatalf("categories = %v (ErrAbortHandler must log nothing): %s", types, out)
	}
	for _, leak := range []string{cn, "goroutine", "errorlog_test.go", "127.0.0.1", "superfluous response"} {
		if strings.Contains(out, leak) {
			t.Fatalf("server error text %q leaked: %s", leak, out)
		}
	}
}

func TestClassifyingWriterTable(t *testing.T) {
	cn := newCanary()
	for _, tc := range []struct {
		line, want string
	}{
		{"http: superfluous response.WriteHeader call from main.handler." + cn + " (main.go:12)\n", "superfluous_write_header"},
		{"http: response.WriteHeader on hijacked connection from main." + cn + " (main.go:3)\n", "hijacked_write_header"},
		{"http: response.Write on hijacked connection from main." + cn + " (main.go:4)\n", "hijacked_write"},
		{"http: invalid Content-Length of \"" + cn + "\"\n", "invalid_content_length"},
		{"http: WriteHeader called with both Transfer-Encoding of \"" + cn + "\" and a Content-Length of 5\n", "conflicting_length_headers"},
		{"http: panic serving 10.0.0.1:5555: " + cn + "\ngoroutine 1 [running]:\n", "handler_panic"},
		{"http: TLS handshake error from 10.0.0.1:5555: " + cn + "\n", "tls_handshake_error"},
		{"http: Accept error: " + cn + "; retrying in 5ms\n", "accept_error"},
		{"http2: server: error reading preface from client " + cn + "\n", "other"},
		{cn + "\n", "other"},
		{"", "other"},
	} {
		var buf bytes.Buffer
		w := classifyingWriter{logger: NewLogger(&buf, slog.LevelInfo, nil)}
		n, err := w.Write([]byte(tc.line))
		if n != len(tc.line) || err != nil {
			t.Fatalf("Write(%q) = %d, %v", tc.line, n, err)
		}
		types := serverErrorRecords(t, buf.String())
		if !slices.Equal(types, []string{tc.want}) {
			t.Fatalf("Write(%q) categories = %v, want [%s]", tc.line, types, tc.want)
		}
		if strings.Contains(buf.String(), cn) {
			t.Fatalf("line text leaked: %s", buf.String())
		}
	}
}

func TestServerErrorLogOneRecordPerLine(t *testing.T) {
	cn := newCanary()
	var buf bytes.Buffer
	l := ServerErrorLog(NewLogger(&buf, slog.LevelInfo, nil))
	l.Printf("http: panic serving %s: %v\n%s", "10.0.0.1:1", cn, "goroutine 7 [running]:\nmain.go:1")
	l.Printf("http: Accept error: %v; retrying in %v", cn, "5ms")
	types := serverErrorRecords(t, buf.String())
	if !slices.Equal(types, []string{"handler_panic", "accept_error"}) {
		t.Fatalf("categories = %v", types)
	}
	if strings.Contains(buf.String(), cn) {
		t.Fatalf("text leaked: %s", buf.String())
	}
	// A nil logger discards instead of failing.
	ServerErrorLog(nil).Printf("http: Accept error: %v", cn)
}
