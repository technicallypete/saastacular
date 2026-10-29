package observability

import (
	"bytes"
	"context"
	"log"
	"log/slog"
)

const serverErrorEvent = "http.server.error"

// serverErrorPrefixes maps the prefixes of net/http's Server.ErrorLog lines
// (Go 1.27.1 call sites) to fixed categories. Anything else is "other".
var serverErrorPrefixes = []struct {
	prefix, category string
}{
	{"http: superfluous response.WriteHeader call", "superfluous_write_header"},
	{"http: response.WriteHeader on hijacked connection", "hijacked_write_header"},
	{"http: response.Write on hijacked connection", "hijacked_write"},
	{"http: invalid Content-Length", "invalid_content_length"},
	{"http: WriteHeader called with both Transfer-Encoding", "conflicting_length_headers"},
	{"http: panic serving", "handler_panic"},
	{"http: TLS handshake error", "tls_handshake_error"},
	{"http: Accept error", "accept_error"},
}

// ServerErrorLog returns a logger for http.Server.ErrorLog. Each line
// net/http writes becomes one http.server.error record (level warn) holding
// only a fixed error_type category; the line's text, which can carry caller
// locations, addresses and panic values, is discarded.
func ServerErrorLog(logger *slog.Logger) *log.Logger {
	return log.New(classifyingWriter{logger: logger}, "", 0)
}

type classifyingWriter struct{ logger *slog.Logger }

// Write emits one record per call and never fails, so net/http is never
// blocked by logging.
func (c classifyingWriter) Write(p []byte) (int, error) {
	if c.logger != nil {
		c.logger.LogAttrs(context.Background(), slog.LevelWarn, serverErrorEvent,
			slog.String("error_type", classifyServerError(p)))
	}
	return len(p), nil
}

func classifyServerError(line []byte) string {
	for _, p := range serverErrorPrefixes {
		if bytes.HasPrefix(line, []byte(p.prefix)) {
			return p.category
		}
	}
	return "other"
}
