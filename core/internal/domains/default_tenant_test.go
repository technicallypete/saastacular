package domains

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/hostname"
	"core/internal/observability"
)

// G071 (Plan r4 S3, S14, S15; §5.1 U2): the startup selection of the default-tenant root alias over a fake querier, the
// root presentation, and the exact closed record shapes on the real core logger.

// dtStore answers the one startup statement: org (nil = NULL) or err; it records every call.
type dtStore struct {
	org      *string
	err      error
	calls    int
	sql      string
	args     []any
	deadline time.Duration
}

type dtRow struct {
	org *string
	err error
}

func (r dtRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(**string) = r.org
	return nil
}

func (s *dtStore) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	s.calls++
	s.sql, s.args = sql, args
	if d, ok := ctx.Deadline(); ok {
		s.deadline = time.Until(d)
	}
	if s.err != nil {
		return dtRow{err: s.err}
	}
	return dtRow{org: s.org}
}

const dtOrg = "0f071000-0000-4000-8000-0000000000a1"

func dtStrp(s string) *string { return &s }

func dtConfig() hostname.Config {
	return hostname.Config{BaseDomain: "localhost", CNAMETarget: "edge.localhost", AuthHost: "auth.localhost", Port: "543",
		ReservedLabels: []string{"blog"}}
}

func dtCanary(t *testing.T) string {
	t.Helper()
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return "cnry" + hex.EncodeToString(b)
}

// dtRecords decodes every JSON line of buf and checks the core envelope of each.
func dtRecords(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("not a JSON record: %q", line)
		}
		if m["service"] != "core" || m["time"] == nil || m["level"] == nil || m["event"] == nil {
			t.Fatalf("envelope: %v", m)
		}
		out = append(out, m)
	}
	return out
}

func dtKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// dtOne asserts exactly one record with exactly the given level, event and members (besides the envelope).
func dtOne(t *testing.T, name string, buf *bytes.Buffer, level, event string, members map[string]string) {
	t.Helper()
	recs := dtRecords(t, buf)
	if len(recs) != 1 {
		t.Fatalf("%s: %d records, want 1: %s", name, len(recs), buf.String())
	}
	r := recs[0]
	want := []string{"event", "level", "service", "time"}
	for k := range members {
		want = append(want, k)
	}
	slices.Sort(want)
	if got := dtKeys(r); !slices.Equal(got, want) {
		t.Fatalf("%s: keys %v, want %v: %s", name, got, want, buf.String())
	}
	if r["level"] != level || r["event"] != event {
		t.Fatalf("%s: level/event %v/%v, want %s/%s", name, r["level"], r["event"], level, event)
	}
	for k, v := range members {
		if r[k] != v {
			t.Fatalf("%s: %s = %v, want %s", name, k, r[k], v)
		}
	}
}

func TestSelectDefaultTenant(t *testing.T) {
	ctx := context.Background()
	cfg := dtConfig()

	// Unset: no statement, no record.
	for _, raw := range []string{"", "   ", "\t\n"} {
		store := &dtStore{org: dtStrp(dtOrg)}
		dt, err := SelectDefaultTenant(ctx, store, cfg, raw)
		if err != nil || dt != (DefaultTenant{}) || dt.Active() || store.calls != 0 {
			t.Fatalf("unset %q: %+v %v (%d statements)", raw, dt, err, store.calls)
		}
		var buf bytes.Buffer
		dt.Log(ctx, observability.NewLogger(&buf, slog.LevelInfo, nil))
		if buf.Len() != 0 {
			t.Fatalf("unset %q logged: %s", raw, buf.String())
		}
	}

	// Invalid exactly as given (the TestSignupSlug contract: grammar plus the reserved floor, the CNAME-target and
	// auth-host labels and PLATFORM_RESERVED_LABELS; no case folding): inactive/invalid_slug without a statement, and the
	// record carries no slug, so the configured value never reaches the log.
	canary := dtCanary(t)
	for _, raw := range []string{"Alpha-Agency", "not_a_slug", "www", "auth", "a--b", "blog", "edge", "ácme", "-alpha",
		"alpha-", "alpha.agency", strings.Repeat("a", 64), "Alpha-Agency-" + canary, canary + "_x", canary + ".localhost"} {
		store := &dtStore{org: dtStrp(dtOrg)}
		dt, err := SelectDefaultTenant(ctx, store, cfg, raw)
		if err != nil || store.calls != 0 || dt.Active() || dt.Outcome != "inactive" || dt.Reason != "invalid_slug" ||
			dt.Slug != "" || dt.OrganizationID != "" {
			t.Fatalf("invalid %q: %+v %v (%d statements)", raw, dt, err, store.calls)
		}
		var buf bytes.Buffer
		dt.Log(ctx, observability.NewLogger(&buf, slog.LevelInfo, nil))
		dtOne(t, "invalid "+raw, &buf, "warn", "default_tenant", map[string]string{"outcome": "inactive", "reason": "invalid_slug"})
		if strings.Contains(buf.String(), raw) || strings.Contains(buf.String(), canary) {
			t.Fatalf("invalid %q: the configured value reached the log: %s", raw, buf.String())
		}
	}

	// Trimmed lookup: one statement, the exact auth-reader call with the trimmed canonical slug, under a 10 s bound.
	store := &dtStore{org: dtStrp(dtOrg)}
	dt, err := SelectDefaultTenant(ctx, store, cfg, "  alpha-agency  ")
	if err != nil || !dt.Active() || dt.Outcome != "active" || dt.Reason != "" || dt.Slug != "alpha-agency" ||
		dt.OrganizationID != dtOrg {
		t.Fatalf("active: %+v %v", dt, err)
	}
	if store.calls != 1 || store.sql != "SELECT app.organization_id_by_slug($1)::text" || len(store.args) != 1 ||
		store.args[0] != "alpha-agency" {
		t.Fatalf("active statement: %d %q %v", store.calls, store.sql, store.args)
	}
	if store.deadline <= 9*time.Second || store.deadline > 10*time.Second {
		t.Fatalf("active statement deadline %v, want the 10 s bound", store.deadline)
	}
	var buf bytes.Buffer
	dt.Log(ctx, observability.NewLogger(&buf, slog.LevelInfo, nil))
	dtOne(t, "active", &buf, "info", "default_tenant", map[string]string{"outcome": "active", "slug": "alpha-agency",
		"organization_id": dtOrg})

	// NULL: inactive/unknown_slug, one statement; the canonical slug is logged, no organization_id.
	store = &dtStore{}
	dt, err = SelectDefaultTenant(ctx, store, cfg, "nope-nobody")
	if err != nil || dt.Active() || dt.Outcome != "inactive" || dt.Reason != "unknown_slug" || dt.Slug != "nope-nobody" ||
		dt.OrganizationID != "" || store.calls != 1 || store.args[0] != "nope-nobody" {
		t.Fatalf("unknown: %+v %v (%d statements)", dt, err, store.calls)
	}
	buf.Reset()
	dt.Log(ctx, observability.NewLogger(&buf, slog.LevelInfo, nil))
	dtOne(t, "unknown", &buf, "warn", "default_tenant", map[string]string{"outcome": "inactive", "reason": "unknown_slug",
		"slug": "nope-nobody"})
	// An empty id is not an Organization either.
	store = &dtStore{org: dtStrp("")}
	if dt, err = SelectDefaultTenant(ctx, store, cfg, "nope-nobody"); err != nil || dt.Active() || dt.Reason != "unknown_slug" {
		t.Fatalf("empty id: %+v %v", dt, err)
	}

	// A store, transport or permission failure is the fixed error (never absence); its text is never carried.
	for _, cause := range []error{errors.New("dial tcp: " + canary), &pgconnLikeError{canary}, context.DeadlineExceeded} {
		store = &dtStore{err: cause}
		dt, err = SelectDefaultTenant(ctx, store, cfg, "alpha-agency")
		if !errors.Is(err, ErrDefaultTenantUnavailable) || err.Error() != "default tenant lookup unavailable" ||
			dt != (DefaultTenant{}) || dt.Active() || store.calls != 1 || strings.Contains(err.Error(), canary) {
			t.Fatalf("store failure %v: %+v %v", cause, dt, err)
		}
	}

	// Wiring as in cmd/server: the selected id pins the resolver (no statement for the root); an inactive outcome
	// carries no id and pins nothing.
	pinned := &dtStore{}
	r := NewResolver(pinned, cfg).WithDefaultTenant(DefaultTenant{Outcome: "active", Slug: "alpha-agency",
		OrganizationID: dtOrg}.OrganizationID)
	if res, err := r.Resolve(ctx, "localhost:543"); err != nil || !res.RootAlias || res.Kind != hostname.KindPlatform ||
		!res.Routing || res.OrganizationID != dtOrg || pinned.calls != 0 {
		t.Fatalf("pinned resolver: %+v %v (%d statements)", res, err, pinned.calls)
	}
	r = NewResolver(&dtStore{}, cfg).WithDefaultTenant(DefaultTenant{Outcome: "inactive", Reason: "unknown_slug",
		Slug: "nope-nobody"}.OrganizationID)
	if res, err := r.Resolve(ctx, "localhost:543"); err != nil || res.RootAlias || res.Kind != hostname.KindPublicEntry {
		t.Fatalf("unpinned resolver: %+v %v", res, err)
	}
}

// pgconnLikeError stands in for a permission refusal (SQLSTATE 42501) or a missing function (42883 against a version-21
// database): any error is unavailable.
type pgconnLikeError struct{ detail string }

func (e *pgconnLikeError) Error() string {
	return "ERROR: permission denied (SQLSTATE 42501) " + e.detail
}

func TestRootPresentation(t *testing.T) {
	ctx := context.Background()
	for _, c := range []struct {
		raw, mode string
		invalid   bool
	}{
		{"", "signup_redirect", false},
		{"signup_redirect", "signup_redirect", false},
		{"landing", "landing", false},
		{"Landing", "signup_redirect", true},
		{"SIGNUP_REDIRECT", "signup_redirect", true},
		{"redirect", "signup_redirect", true},
		{"https://x", "signup_redirect", true},
		{"signup", "signup_redirect", true},
	} {
		mode, invalid := RootPresentation(c.raw)
		if mode != c.mode || invalid != c.invalid {
			t.Errorf("%q: %q %v, want %q %v", c.raw, mode, invalid, c.mode, c.invalid)
		}
		// The server writes the warning exactly when invalid is set (cmd/server): valid modes emit nothing.
		var buf bytes.Buffer
		if invalid {
			LogInvalidRootPage(ctx, observability.NewLogger(&buf, slog.LevelInfo, nil))
			dtOne(t, "invalid "+c.raw, &buf, "warn", "platform_root_page",
				map[string]string{"outcome": "invalid_value", "fallback": "signup_redirect"})
		} else if buf.Len() != 0 {
			t.Errorf("%q logged: %s", c.raw, buf.String())
		}
	}

	// The raw value is never logged.
	canary := dtCanary(t)
	raw := "bogus-" + canary
	mode, invalid := RootPresentation(raw)
	if mode != "signup_redirect" || !invalid {
		t.Fatalf("%q: %q %v", raw, mode, invalid)
	}
	var buf bytes.Buffer
	LogInvalidRootPage(ctx, observability.NewLogger(&buf, slog.LevelInfo, nil))
	if strings.Contains(buf.String(), canary) || strings.Contains(buf.String(), "bogus") {
		t.Fatalf("the configured value reached the log: %s", buf.String())
	}
}
