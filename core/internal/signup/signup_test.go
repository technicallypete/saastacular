package signup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"core/internal/hostname"
	"core/internal/identity"
)

func unitService() *Service {
	id := identity.NewService(identity.NewStore(nil), identity.Config{RootOrigin: "https://localhost:543"})
	return NewService(Config{Identity: id, Hostname: hostname.Config{BaseDomain: "localhost", Port: "543"},
		RootOrigin: "https://localhost:543"})
}

// TestSignupInputsRefusedBeforeSQL: every malformed input is refused with its closed code before any statement (the
// unit service has no database; reaching SQL would panic on the nil pool).
func TestSignupInputsRefusedBeforeSQL(t *testing.T) {
	ctx := context.Background()
	s := unitService()
	var zero identity.SignupCredentialInfo
	for _, name := range []string{"", "   ", "a\x00b", "a\u0085b", strings.Repeat("x", 201), "\xff"} {
		if _, err := s.Provision(ctx, zero, name, "acme"); CodeOf(err) != CodeInvalidRequest || AsError(err).Field != FieldName {
			t.Errorf("Provision(name %q) = %v", name, err)
		}
		if _, err := s.SuggestSlug(ctx, zero, name); CodeOf(err) != CodeInvalidRequest || AsError(err).Field != FieldName {
			t.Errorf("SuggestSlug(name %q) = %v", name, err)
		}
	}
	for _, slug := range []string{"", "www", "auth", "Acme", "a.b", "acme-"} {
		if _, err := s.Provision(ctx, zero, "Acme", slug); CodeOf(err) != CodeInvalidRequest || AsError(err).Field != FieldSlug {
			t.Errorf("Provision(slug %q) = %v", slug, err)
		}
	}
	if sg, err := s.SuggestSlug(ctx, zero, "日本"); err != nil || sg != (Suggestion{}) {
		t.Errorf("SuggestSlug(no base) = %+v %v", sg, err)
	}
	for _, ref := range []string{"", "x", strings.Repeat("A", 43) + "=", strings.Repeat("A", 44)} {
		if _, err := s.Challenge(ctx, "https://acme.localhost:543", "", ref); CodeOf(err) != CodeContinuationInvalid {
			t.Errorf("Challenge(ref %q) = %v", ref, err)
		}
	}
	if _, err := s.Challenge(ctx, "", "", encode(make([]byte, 32))); CodeOf(err) != CodeContinuationInvalid {
		t.Errorf("Challenge(no origin) = %v", err)
	}
	if _, err := s.Prove(ctx, zero, "bad"); CodeOf(err) != CodeContextMismatch {
		t.Errorf("Prove(bad nonce) = %v", err)
	}
	if _, err := s.Prove(ctx, zero, encode(make([]byte, 32))); CodeOf(err) != CodeContextMismatch {
		t.Errorf("Prove(unprovisioned) = %v", err)
	}
	if _, err := s.Complete(ctx, "https://acme.localhost:543", "", "bad", encode(make([]byte, 32))); CodeOf(err) != CodeContinuationInvalid {
		t.Errorf("Complete(bad challenge) = %v", err)
	}
	if _, err := s.Resolve(ctx, "not-a-credential", "https://localhost:543"); CodeOf(err) != CodeUnauthenticated {
		t.Errorf("Resolve(malformed) = %v", err)
	}
	if _, err := s.Resolve(ctx, encode(make([]byte, 32)), "https://www.localhost:543"); CodeOf(err) != CodeUnauthenticated {
		t.Errorf("Resolve(www origin) = %v", err)
	}
}

// TestSignupUnconfigured: without identity, a root origin or a platform domain every operation is signup_unconfigured.
func TestSignupUnconfigured(t *testing.T) {
	ctx := context.Background()
	for _, s := range []*Service{NewService(Config{}), NewService(Config{RootOrigin: "https://localhost:543"}), nil} {
		if s.Configured() {
			t.Fatal("Configured() = true")
		}
	}
	if got := (*Service)(nil).RootOrigin(); got != "" {
		t.Errorf("nil RootOrigin() = %q", got)
	}
	if got := NewService(Config{RootOrigin: "https://localhost:543"}).RootOrigin(); got != "https://localhost:543" {
		t.Errorf("RootOrigin() = %q", got)
	}
	s := NewService(Config{Hostname: hostname.Config{BaseDomain: "localhost"}})
	var zero identity.SignupCredentialInfo
	if _, err := s.Resolve(ctx, "x", "y"); CodeOf(err) != CodeUnconfigured {
		t.Errorf("Resolve = %v", err)
	}
	if _, err := s.Provision(ctx, zero, "Acme", "acme"); CodeOf(err) != CodeUnconfigured {
		t.Errorf("Provision = %v", err)
	}
	if _, err := s.Challenge(ctx, "o", "", "r"); CodeOf(err) != CodeUnconfigured {
		t.Errorf("Challenge = %v", err)
	}
	if _, err := s.Prove(ctx, zero, "n"); CodeOf(err) != CodeUnconfigured {
		t.Errorf("Prove = %v", err)
	}
	if _, err := s.Complete(ctx, "o", "", "c", "p"); CodeOf(err) != CodeUnconfigured {
		t.Errorf("Complete = %v", err)
	}
	if err := s.Abandon(ctx, zero); CodeOf(err) != CodeUnconfigured {
		t.Errorf("Abandon = %v", err)
	}
}

// TestSignupIntentAndSecrets: the intent digest is SHA-256(name || 0x00 || slug) (so "ab"+"c" and "a"+"bc" differ), and
// the secret codec accepts exactly one canonical 32-byte base64url encoding.
func TestSignupIntentAndSecrets(t *testing.T) {
	want := sha256.Sum256([]byte("Acme\x00acme"))
	if !bytes.Equal(Intent("Acme", "acme"), want[:]) {
		t.Error("Intent differs from SHA-256(name || 0x00 || slug)")
	}
	if bytes.Equal(Intent("ab", "c"), Intent("a", "bc")) {
		t.Error("Intent is ambiguous across the separator")
	}
	b, err := newSecret()
	if err != nil || len(b) != 32 {
		t.Fatal(err)
	}
	if got, ok := decode(encode(b)); !ok || !bytes.Equal(got, b) {
		t.Error("decode(encode(b)) failed")
	}
	enc := encode(b)
	for _, bad := range []string{enc + "A", enc[:42], enc + "=", "+" + enc[1:], "/" + enc[1:], enc[:42] + "B", ""} {
		if _, ok := decode(bad); ok {
			t.Errorf("decode(%q) accepted", bad)
		}
	}
}

// TestStatusOf: pending until provisioned, created while a continuation is outstanding, completed afterwards.
func TestStatusOf(t *testing.T) {
	for _, c := range []struct {
		info identity.SignupCredentialInfo
		want string
	}{
		{identity.SignupCredentialInfo{}, StatusPending},
		{identity.SignupCredentialInfo{Provisioned: true, Generation: 1}, StatusCreated},
		{identity.SignupCredentialInfo{Provisioned: true, Generation: 2, Completed: true}, StatusCompleted},
	} {
		if got := StatusOf(c.info); got != c.want {
			t.Errorf("StatusOf(%+v) = %q, want %q", c.info, got, c.want)
		}
	}
}

// TestLogRequestIsClosed (SUP-13): one signup.request record with the operation, the closed outcome and reason and
// nothing else; operations and reasons outside the closed sets are not logged.
func TestLogRequestIsClosed(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx := context.Background()
	LogRequest(ctx, l, OpCreate, "")
	LogRequest(ctx, l, OpCreate, string(CodeSlugTaken))
	LogRequest(ctx, l, OpProve, string(CodeUnavailable))
	LogRequest(ctx, l, "drop_tables", "")
	LogRequest(ctx, l, OpCreate, "acme")
	LogRequest(ctx, nil, OpCreate, "")
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("records = %d:\n%s", len(lines), buf.String())
	}
	want := []map[string]string{
		{"level": "INFO", "msg": EventRequest, "operation": "create", "outcome": "ok"},
		{"level": "WARN", "msg": EventRequest, "operation": "create", "outcome": "refused", "reason": "slug_taken"},
		{"level": "ERROR", "msg": EventRequest, "operation": "prove", "outcome": "failed", "reason": "signup_unavailable"},
	}
	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatal(err)
		}
		delete(m, "time")
		if len(m) != len(want[i]) {
			t.Errorf("record %d has keys %v", i, m)
		}
		for k, v := range want[i] {
			if m[k] != v {
				t.Errorf("record %d %s = %v, want %s", i, k, m[k], v)
			}
		}
	}
}
