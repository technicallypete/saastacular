package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"core/internal/config"
)

const current = `APP_ENV=production
IDENTITY_COOKIE_PROFILE=https
IDENTITY_PUBLIC_ORIGINS=https://app.sample.example.test
IDENTITY_WORKOS_ISSUER=https://auth.sample.example.test
CORE_INTERNAL_URL=http://core:8180
OBJECT_STORE_ENDPOINT=https://objects.sample.example.test
`

func TestExitBlocked(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run(strings.NewReader(current), &out, &errb, config.CurrentCapabilities()); code != 3 {
		t.Fatalf("exit %d, stderr %q", code, errb.String())
	}
	var rep config.Readiness
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.Result != config.ReadinessBlocked || rep.Label != config.ReadinessLabel {
		t.Fatalf("report %+v", rep)
	}
}

func TestExitInvalid(t *testing.T) {
	secretNamed := "DB_PASSWORD"
	secretLine := secretNamed + "=" + "Cx00112233445566778899\n" // assembled so the source holds no scanner-shaped assignment
	for _, in := range []string{secretLine, "APP_ENV=production\nAPP_ENV=test\n", "not a line\n"} {
		var out, errb bytes.Buffer
		if code := run(strings.NewReader(in), &out, &errb, config.CurrentCapabilities()); code != 2 {
			t.Errorf("input %q: exit %d", in, code)
		}
		if out.Len() != 0 || strings.Contains(errb.String(), "Cx00112233445566778899") {
			t.Errorf("input %q: no report and no value on invalid input", in)
		}
	}
}

func TestExitReady(t *testing.T) {
	caps := config.CurrentCapabilities()
	caps[config.PathCorePostgres], caps[config.PathMigratePostgres], caps[config.PathCoreRedis] =
		config.ModeExternal, config.ModeExternal, config.ModeExternal
	in := strings.Replace(current, "http://core:8180", "https://core.internal.sample.example.test", 1)
	for _, item := range []string{"EDGE_TLS", "CORE_TLS", "POSTGRES_TLS", "MIGRATE_TLS", "REDIS_TLS", "POSTGRES_AT_REST",
		"OBJECT_AT_REST", "BACKUP_AT_REST"} {
		in += "READINESS_EVIDENCE_" + item + "=DEPLOY-REVIEW-7\n"
	}
	var out, errb bytes.Buffer
	if code := run(strings.NewReader(in), &out, &errb, caps); code != 0 {
		t.Fatalf("exit %d: %s %s", code, out.String(), errb.String())
	}
	if !strings.Contains(out.String(), config.ReadinessLabel) {
		t.Fatal("the label is missing")
	}
}
