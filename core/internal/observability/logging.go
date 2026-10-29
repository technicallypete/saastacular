// Package observability provides structured, redacted logging and request
// and trace correlation. Logs never carry secret values or credential fields.
package observability

import (
	"context"
	"io"
	"log/slog"
	"regexp"
	"strings"
)

const redacted = "[REDACTED]"

// serviceName is the value of the envelope's service key on every record.
const serviceName = "core"

var sensitiveKey = regexp.MustCompile(`(?i)(password|passwd|secret|token|api_?key|authorization|cookie|dsn|credential)`)

// NewLogger returns a JSON logger that redacts sensitive attribute keys and
// any occurrence of the given secret values. Every record carries the
// envelope: time (RFC 3339 UTC, millisecond precision, Z), a lowercase level,
// service and event (the record message).
func NewLogger(w io.Writer, level slog.Level, secrets []string) *slog.Logger {
	base := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level, ReplaceAttr: envelopeAttr}).
		WithAttrs([]slog.Attr{slog.String("service", serviceName)})
	return slog.New(&redactingHandler{next: base, secrets: nonEmpty(secrets)})
}

// envelopeAttr rewrites slog's built-in keys into the log envelope. Only
// top-level built-in attributes are changed.
func envelopeAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return a
	}
	switch a.Key {
	case slog.TimeKey:
		if a.Value.Kind() == slog.KindTime {
			return slog.String(slog.TimeKey, a.Value.Time().UTC().Format("2006-01-02T15:04:05.000Z"))
		}
	case slog.LevelKey:
		if l, ok := a.Value.Any().(slog.Level); ok {
			return slog.String(slog.LevelKey, levelName(l))
		}
	case slog.MessageKey:
		return slog.Attr{Key: "event", Value: a.Value}
	}
	return a
}

// levelName maps a slog level onto the envelope's four level names.
func levelName(l slog.Level) string {
	switch {
	case l < slog.LevelInfo:
		return "debug"
	case l < slog.LevelWarn:
		return "info"
	case l < slog.LevelError:
		return "warn"
	default:
		return "error"
	}
}

// secretMatcher reports whether a string contains a secret value configured
// on logger's redacting handler. It never matches for other loggers.
func secretMatcher(logger *slog.Logger) func(string) bool {
	if logger != nil {
		if h, ok := logger.Handler().(*redactingHandler); ok {
			return h.containsSecret
		}
	}
	return func(string) bool { return false }
}

func nonEmpty(in []string) []string {
	var out []string
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

type redactingHandler struct {
	next    slog.Handler
	secrets []string
}

func (h *redactingHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.next.Enabled(ctx, l)
}

func (h *redactingHandler) Handle(ctx context.Context, r slog.Record) error {
	out := slog.NewRecord(r.Time, r.Level, h.scrub(r.Message), r.PC)
	r.Attrs(func(a slog.Attr) bool {
		out.AddAttrs(h.redact(a))
		return true
	})
	return h.next.Handle(ctx, out)
}

func (h *redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clean := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		clean[i] = h.redact(a)
	}
	return &redactingHandler{next: h.next.WithAttrs(clean), secrets: h.secrets}
}

func (h *redactingHandler) WithGroup(name string) slog.Handler {
	return &redactingHandler{next: h.next.WithGroup(name), secrets: h.secrets}
}

func (h *redactingHandler) redact(a slog.Attr) slog.Attr {
	if sensitiveKey.MatchString(a.Key) {
		return slog.String(a.Key, redacted)
	}
	v := a.Value.Resolve()
	switch v.Kind() {
	case slog.KindGroup:
		group := v.Group()
		clean := make([]any, 0, len(group))
		for _, g := range group {
			clean = append(clean, h.redact(g))
		}
		return slog.Group(a.Key, clean...)
	case slog.KindString:
		return slog.String(a.Key, h.scrub(v.String()))
	default:
		if s := v.String(); h.containsSecret(s) {
			return slog.String(a.Key, h.scrub(s))
		}
		return slog.Attr{Key: a.Key, Value: v}
	}
}

func (h *redactingHandler) containsSecret(s string) bool {
	for _, sec := range h.secrets {
		if strings.Contains(s, sec) {
			return true
		}
	}
	return false
}

func (h *redactingHandler) scrub(s string) string {
	for _, sec := range h.secrets {
		s = strings.ReplaceAll(s, sec, redacted)
	}
	return s
}
