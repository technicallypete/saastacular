package config

import (
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// productionShaped is today's production-shaped configuration: https edge, origins, issuer and object store, the
// plaintext internal web -> core URL, intelligence and gateway disabled, no evidence references.
func productionShaped() map[string]string {
	return map[string]string{
		"APP_ENV":                 "production",
		"IDENTITY_COOKIE_PROFILE": "https",
		"IDENTITY_PUBLIC_ORIGINS": "https://app.sample.example.test",
		"IDENTITY_WORKOS_ISSUER":  "https://auth.sample.example.test",
		"CORE_INTERNAL_URL":       "http://core:8180",
		"OBJECT_STORE_ENDPOINT":   "https://objects.sample.example.test",
	}
}

func allEvidence() map[string]string {
	m := map[string]string{}
	for _, item := range evidenceItem {
		m["READINESS_EVIDENCE_"+item] = "DEPLOY-REVIEW-2026-" + strings.ReplaceAll(strings.ToLower(item), "_", "-")
	}
	return m
}

func statuses(r Readiness) map[string]string {
	m := map[string]string{}
	for _, p := range r.Paths {
		m[p.Path] = p.Status
	}
	return m
}

func assess(t *testing.T, in map[string]string, caps Capabilities) Readiness {
	t.Helper()
	r, err := AssessReadiness(in, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Label != ReadinessLabel {
		t.Fatalf("label missing: %q", r.Label)
	}
	b, _ := json.Marshal(r)
	if strings.Contains(strings.ToLower(string(b)), "encrypted") {
		t.Fatal("a readiness report must never claim anything is encrypted")
	}
	return r
}

// U-RDY-1: the current production-shaped configuration is blocked with the exact prerequisite set.
func TestReadinessCurrentProductionShapeIsBlocked(t *testing.T) {
	r := assess(t, productionShaped(), CurrentCapabilities())
	want := map[string]string{
		PathBrowserWeb: StatusBlockedUnknown, PathWebCore: StatusBlockedPlaintext, PathCorePostgres: StatusBlockedUnsupport,
		PathMigratePostgres: StatusBlockedUnsupport, PathCoreRedis: StatusBlockedUnsupport,
		PathCoreObjectStore: StatusConfigVerified, PathSignedDownload: StatusConfigVerified, PathCoreWorkOS: StatusConfigVerified,
		PathCoreIntelligence: StatusDisabled, PathGateway: StatusDisabled, PathPostgresAtRest: StatusBlockedUnknown,
		PathObjectAtRest: StatusBlockedUnknown, PathBackupAtRest: StatusBlockedUnknown, PathRedisAtRest: StatusNotApplicable,
		PathAppEnv: StatusConfigVerified,
	}
	if got := statuses(r); !reflect.DeepEqual(got, want) {
		t.Fatalf("statuses\n got %v\nwant %v", got, want)
	}
	if r.Result != ReadinessBlocked {
		t.Fatalf("result %q", r.Result)
	}
	var blocked []string
	for _, p := range []string{PathBrowserWeb, PathWebCore, PathCorePostgres, PathMigratePostgres, PathCoreRedis,
		PathPostgresAtRest, PathObjectAtRest, PathBackupAtRest} {
		blocked = append(blocked, p+": "+prerequisite[p])
	}
	if !reflect.DeepEqual(r.Prerequisites, blocked) {
		t.Fatalf("prerequisites\n got %q\nwant %q", r.Prerequisites, blocked)
	}
	if len(r.LocalExceptions) != 0 {
		t.Fatalf("production has no local exceptions, got %v", r.LocalExceptions)
	}
}

// Evidence can never lift a path the current code cannot secure.
func TestReadinessEvidenceCannotLiftUnsupported(t *testing.T) {
	in := productionShaped()
	in["CORE_INTERNAL_URL"] = "https://core.internal.sample.example.test"
	maps.Copy(in, allEvidence())
	r := assess(t, in, CurrentCapabilities())
	s := statuses(r)
	for _, p := range []string{PathCorePostgres, PathMigratePostgres, PathCoreRedis} {
		if s[p] != StatusBlockedUnsupport {
			t.Errorf("%s: %s", p, s[p])
		}
	}
	if r.Result != ReadinessBlocked {
		t.Fatal("unsupported paths must keep the result blocked")
	}
	for _, p := range []string{PathBrowserWeb, PathWebCore, PathPostgresAtRest, PathObjectAtRest, PathBackupAtRest} {
		if s[p] != StatusExternalDeclared {
			t.Errorf("%s: %s, want external-declared", p, s[p])
		}
	}
}

// U-RDY-2: the development fixture is blocked and lists its local exceptions.
func TestReadinessDevelopmentFixture(t *testing.T) {
	in := map[string]string{
		"APP_ENV": "development", "IDENTITY_COOKIE_PROFILE": "loopback-http",
		"IDENTITY_PUBLIC_ORIGINS": "http://localhost:3100", "IDENTITY_WORKOS_ISSUER": "https://auth.sample.example.test",
		"CORE_INTERNAL_URL": "http://core:8180", "OBJECT_STORE_ENDPOINT": "http://minio:9000",
		"INTELLIGENCE_URL": "http://intelligence:8000",
	}
	r := assess(t, in, CurrentCapabilities())
	if r.Result != ReadinessBlocked {
		t.Fatal("development must be blocked")
	}
	want := []string{PathBrowserWeb, PathWebCore, PathCoreObjectStore, PathSignedDownload, PathCoreIntelligence, PathAppEnv}
	if !reflect.DeepEqual(r.LocalExceptions, want) {
		t.Fatalf("local exceptions %v, want %v", r.LocalExceptions, want)
	}
	if s := statuses(r); s[PathAppEnv] != StatusBlockedLocalExcep || s[PathBrowserWeb] != StatusBlockedLocalExcep {
		t.Fatalf("statuses %v", s)
	}
}

// U-RDY-3: negatives, each blocking the named path in production.
func TestReadinessNegatives(t *testing.T) {
	cases := []struct {
		name string
		set  map[string]string
		del  []string
		path string
		want string
	}{
		{"http internal web->core", map[string]string{"CORE_INTERNAL_URL": "http://core.internal:8180"}, nil, PathWebCore, StatusBlockedPlaintext},
		{"missing web->core", nil, []string{"CORE_INTERNAL_URL"}, PathWebCore, StatusBlockedUnknown},
		{"http intelligence", map[string]string{"INTELLIGENCE_URL": "http://intelligence:8000"}, nil, PathCoreIntelligence, StatusBlockedPlaintext},
		{"http object endpoint", map[string]string{"OBJECT_STORE_ENDPOINT": "http://minio:9000"}, nil, PathCoreObjectStore, StatusBlockedPlaintext},
		{"http signing endpoint", map[string]string{"OBJECT_STORE_SIGNING_ENDPOINT": "http://minio:9000"}, nil, PathSignedDownload, StatusBlockedPlaintext},
		{"object endpoint with a path", map[string]string{"OBJECT_STORE_ENDPOINT": "https://objects.sample.example.test/bucket"}, nil, PathCoreObjectStore, StatusBlockedUnknown},
		{"PGSSLMODE disable", map[string]string{"PGSSLMODE": "disable"}, nil, PathCorePostgres, StatusBlockedPlaintext},
		{"PGSSLMODE prefer", map[string]string{"PGSSLMODE": "prefer"}, nil, PathCorePostgres, StatusBlockedPlaintext},
		{"PGSSLMODE allow", map[string]string{"PGSSLMODE": "allow"}, nil, PathCorePostgres, StatusBlockedPlaintext},
		{"PGSSLMODE require", map[string]string{"PGSSLMODE": "require"}, nil, PathCorePostgres, StatusBlockedBypass},
		{"PGSSLMODE verify-full", map[string]string{"PGSSLMODE": "verify-full"}, nil, PathCorePostgres, StatusBlockedUnsupport},
		{"ambient PGHOST", map[string]string{"PGHOST": "db.example.test"}, nil, PathCorePostgres, StatusBlockedUnsupport},
		{"loopback-http cookies", map[string]string{"IDENTITY_COOKIE_PROFILE": "loopback-http"}, nil, PathBrowserWeb, StatusBlockedLocalExcep},
		{"http origin", map[string]string{"IDENTITY_PUBLIC_ORIGINS": "https://a.example.test,http://b.example.test"}, nil, PathBrowserWeb, StatusBlockedPlaintext},
		{"no origin listed", map[string]string{"IDENTITY_PUBLIC_ORIGINS": " , "}, nil, PathBrowserWeb, StatusBlockedUnknown},
		{"test API base override", map[string]string{"IDENTITY_WORKOS_API_BASE_URL": "http://identity-emulator:8080"}, nil, PathCoreWorkOS, StatusBlockedLocalExcep},
		{"http issuer", map[string]string{"IDENTITY_WORKOS_ISSUER": "http://auth.example.test"}, nil, PathCoreWorkOS, StatusBlockedPlaintext},
		{"boolean edge evidence", map[string]string{"CORE_INTERNAL_URL": "https://core.example.test", "READINESS_EVIDENCE_EDGE_TLS": "true"}, nil, PathBrowserWeb, StatusBlockedUnknown},
		{"boolean at-rest evidence", map[string]string{"READINESS_EVIDENCE_POSTGRES_AT_REST": "YES"}, nil, PathPostgresAtRest, StatusBlockedUnknown},
		{"missing backup evidence", map[string]string{"READINESS_EVIDENCE_POSTGRES_AT_REST": "DR-1234"}, nil, PathBackupAtRest, StatusBlockedUnknown},
		{"non-production APP_ENV", map[string]string{"APP_ENV": "test"}, nil, PathAppEnv, StatusBlockedLocalExcep},
		{"missing APP_ENV", nil, []string{"APP_ENV"}, PathAppEnv, StatusBlockedUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := productionShaped()
			maps.Copy(in, tc.set)
			for _, k := range tc.del {
				delete(in, k)
			}
			r := assess(t, in, CurrentCapabilities())
			if got := statuses(r)[tc.path]; got != tc.want {
				t.Fatalf("%s: got %s, want %s", tc.path, got, tc.want)
			}
			if r.Result != ReadinessBlocked {
				t.Fatal("result must be blocked")
			}
		})
	}
}

// U-RDY-4: secret-named, unknown, duplicate and malformed input is refused before evaluation, naming keys only.
func TestReadinessRefusesInvalidInput(t *testing.T) {
	for _, k := range []string{"DB_PASSWORD", "OBJECT_STORE_SECRET_KEY", "INTELLIGENCE_SERVICE_TOKEN", "OPENROUTER_API_KEY",
		"IDENTITY_WORKOS_API_KEY_FILE", "MY_PASSWORD_HINT", "SOME_SECRET", "VAULT_TOKEN", "X_API_KEY"} {
		in := productionShaped()
		in[k] = "Cx0123456789abcdef"
		_, err := AssessReadiness(in, CurrentCapabilities())
		if !errors.Is(err, ErrReadinessInput) || !strings.Contains(err.Error(), k) || strings.Contains(err.Error(), "Cx0123456789abcdef") {
			t.Errorf("%s: want a refusal naming the key only, got %v", k, err)
		}
	}
	secretNamed := "DB_PASSWORD"
	secretLine := secretNamed + "=" + "Cx0123456789abcdef\n" // assembled so the source holds no scanner-shaped assignment
	for _, in := range []string{"APP_ENV=production\nAPP_ENV=production\n", "APP_ENV production\n", "lower=1\n",
		"UNKNOWN_SETTING=1\n", "READINESS_EVIDENCE_NOT_AN_ITEM=REF-1\n", "READINESS_GATEWAY_ENABLED=maybe\n",
		secretLine} {
		_, err := ParseReadinessInput(strings.NewReader(in))
		if !errors.Is(err, ErrReadinessInput) || strings.Contains(err.Error(), "Cx0123456789abcdef") {
			t.Errorf("input %q: want ErrReadinessInput without values, got %v", in, err)
		}
	}
	// Oversize input is refused, never silently truncated.
	big := "APP_ENV=production\n" + strings.Repeat("# padding\n", readinessInputMax/10+1)
	if _, err := ParseReadinessInput(strings.NewReader(big)); !errors.Is(err, ErrReadinessInput) || !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("oversize input: %v", err)
	}
	got, err := ParseReadinessInput(strings.NewReader("# comment\n\nAPP_ENV=production\nCORE_INTERNAL_URL=https://core.example.test\n"))
	if err != nil || got["APP_ENV"] != "production" || got["CORE_INTERNAL_URL"] != "https://core.example.test" || len(got) != 2 {
		t.Fatalf("valid input: %v %v", got, err)
	}
}

// U-RDY-5: a synthetic qualifying configuration with injected capabilities and every reference is configuration-ready,
// carrying the label.
func TestReadinessSyntheticPositive(t *testing.T) {
	caps := CurrentCapabilities()
	caps[PathCorePostgres], caps[PathMigratePostgres], caps[PathCoreRedis] = ModeExternal, ModeExternal, ModeExternal
	in := productionShaped()
	in["CORE_INTERNAL_URL"] = "https://core.internal.sample.example.test"
	in["INTELLIGENCE_URL"] = "https://intelligence.internal.sample.example.test"
	in["OBJECT_STORE_SIGNING_ENDPOINT"] = "https://downloads.sample.example.test"
	in["READINESS_GATEWAY_ENABLED"] = "true"
	maps.Copy(in, allEvidence())
	r := assess(t, in, caps)
	if r.Result != ReadinessReady {
		t.Fatalf("result %q, statuses %v, prerequisites %v", r.Result, statuses(r), r.Prerequisites)
	}
	for _, p := range r.Paths {
		if !slices.Contains([]string{StatusConfigVerified, StatusExternalDeclared, StatusNotApplicable}, p.Status) {
			t.Errorf("%s: %s", p.Path, p.Status)
		}
	}
	if len(r.Prerequisites) != 0 || len(r.LocalExceptions) != 0 {
		t.Fatal("a ready report lists no prerequisites or local exceptions")
	}
	// One missing at-rest reference blocks it again.
	delete(in, "READINESS_EVIDENCE_BACKUP_AT_REST")
	if r := assess(t, in, caps); r.Result != ReadinessBlocked {
		t.Fatal("missing at-rest evidence must block")
	}
}

// U-RDY-6: CurrentCapabilities is pinned; changing a value needs a reviewed code change.
func TestCurrentCapabilitiesPinned(t *testing.T) {
	want := Capabilities{
		"browser-web": "external", "web-core": "external", "core-postgres": "unsupported", "migrate-postgres": "unsupported",
		"core-redis": "unsupported", "core-object-store": "in-platform", "signed-download": "in-platform",
		"core-workos": "in-platform", "core-intelligence": "external", "intelligence-gateway": "in-platform",
		"postgres-at-rest": "external", "object-storage-at-rest": "external", "backups-at-rest": "external",
		"redis-at-rest": "in-platform",
	}
	if got := CurrentCapabilities(); !reflect.DeepEqual(got, want) {
		t.Fatalf("CurrentCapabilities changed:\n got %v\nwant %v", got, want)
	}
}
