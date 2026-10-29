package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// U-CFG (G004 Plan r5 §8.2): missing, partial and malformed identity configuration; production refusals.

const (
	testClient = "client_01TESTCLIENT0000000000000"
	testIssuer = "https://api.workos.com/user_management/client_01TESTCLIENT0000000000000"
	testKey    = "sk_test_" + "identityconfigcanary" // assembled, as the fake does, so no key-shaped literal is committed
)

func identityEnv(over map[string]string) Getenv {
	base := map[string]string{
		"IDENTITY_PROVIDER":         "workos",
		"IDENTITY_WORKOS_CLIENT_ID": testClient,
		"IDENTITY_WORKOS_ISSUER":    testIssuer,
		"IDENTITY_WORKOS_API_KEY":   testKey,
		"IDENTITY_PUBLIC_ORIGINS":   "https://org-a.localhost:543,https://org-b.localhost:543",
	}
	for k, v := range over {
		if v == "<unset>" {
			delete(base, k)
		} else {
			base[k] = v
		}
	}
	return func(k string) string { return base[k] }
}

func TestLoadIdentityDefaults(t *testing.T) {
	id, err := LoadIdentity(identityEnv(nil), "development")
	if err != nil {
		t.Fatal(err)
	}
	if !id.Configured() || id.APIBaseURL != "https://api.workos.com" || id.CookieProfile != "https" ||
		id.SessionIdle != 30*time.Minute || id.SessionMax != 8*time.Hour || id.MaxAuthAge != 0 ||
		!slices.Equal(id.Algorithms, []string{"RS256"}) || len(id.PublicOrigins) != 2 {
		t.Fatalf("defaults: %+v", id)
	}
	if got := id.SecretValues(); !slices.Equal(got, []string{testKey}) {
		t.Fatalf("secret values = %v", got)
	}
}

func TestLoadIdentityUnconfigured(t *testing.T) {
	none := func(string) string { return "" }
	id, err := LoadIdentity(none, "development")
	if err != nil || id.Configured() || id.SecretValues() != nil {
		t.Fatalf("unconfigured development: %+v %v", id, err)
	}
	// The rendered placeholder value is refused: the historical OIDC placeholder does not define the AuthKit API.
	if _, err := LoadIdentity(func(k string) string {
		if k == "IDENTITY_PROVIDER" {
			return "oidc"
		}
		return ""
	}, "development"); err == nil || !strings.Contains(err.Error(), "IDENTITY_PROVIDER") {
		t.Fatalf("oidc placeholder accepted: %v", err)
	}
	if _, err := LoadIdentity(none, "production"); err == nil || !strings.Contains(err.Error(), "IDENTITY_WORKOS_CLIENT_ID") {
		t.Fatalf("production accepted unconfigured identity: %v", err)
	}
}

func TestLoadIdentityRefusals(t *testing.T) {
	cases := []struct {
		name   string
		env    map[string]string
		appEnv string
		want   []string
	}{
		{"partial: no issuer", map[string]string{"IDENTITY_WORKOS_ISSUER": "<unset>"}, "development", []string{"IDENTITY_WORKOS_ISSUER is required"}},
		{"partial: no key", map[string]string{"IDENTITY_WORKOS_API_KEY": "<unset>"}, "test", []string{"IDENTITY_WORKOS_API_KEY is required"}},
		{"partial: no client", map[string]string{"IDENTITY_WORKOS_CLIENT_ID": "<unset>"}, "development", []string{"IDENTITY_WORKOS_CLIENT_ID is required"}},
		{"settings without a client", map[string]string{"IDENTITY_WORKOS_CLIENT_ID": "<unset>", "IDENTITY_WORKOS_ISSUER": "<unset>", "IDENTITY_WORKOS_API_KEY": "<unset>"}, "development", []string{"IDENTITY_PUBLIC_ORIGINS is set but"}},
		{"both key sources", map[string]string{"IDENTITY_WORKOS_API_KEY_FILE": "/nonexistent"}, "development", []string{"mutually exclusive"}},
		{"unreadable key file", map[string]string{"IDENTITY_WORKOS_API_KEY": "<unset>", "IDENTITY_WORKOS_API_KEY_FILE": "/nonexistent/key"}, "development", []string{"IDENTITY_WORKOS_API_KEY_FILE cannot be read"}},
		{"malformed client", map[string]string{"IDENTITY_WORKOS_CLIENT_ID": "abc"}, "development", []string{"not a WorkOS client ID"}},
		{"http issuer", map[string]string{"IDENTITY_WORKOS_ISSUER": "http://api.workos.com"}, "production", []string{"exact https issuer"}},
		{"issuer with query", map[string]string{"IDENTITY_WORKOS_ISSUER": "https://api.workos.com/?x=1"}, "development", []string{"exact https issuer"}},
		{"emulator base outside test", map[string]string{"IDENTITY_WORKOS_API_BASE_URL": "https://emulator.localhost"}, "development", []string{"only with APP_ENV=test"}},
		{"emulator base in production", map[string]string{"IDENTITY_WORKOS_API_BASE_URL": "https://emulator.localhost"}, "production", []string{"only with APP_ENV=test"}},
		{"http default-host base outside test", map[string]string{"IDENTITY_WORKOS_API_BASE_URL": "http://api.workos.com"}, "production", []string{"only with APP_ENV=test"}},
		{"base with path", map[string]string{"IDENTITY_WORKOS_API_BASE_URL": "https://api.workos.com/v1"}, "test", []string{"without path"}},
		{"loopback-http in production", map[string]string{"IDENTITY_COOKIE_PROFILE": "loopback-http", "IDENTITY_PUBLIC_ORIGINS": "http://localhost:3100"}, "production", []string{"loopback-http is refused in production"}},
		{"unknown cookie profile", map[string]string{"IDENTITY_COOKIE_PROFILE": "lax"}, "development", []string{"IDENTITY_COOKIE_PROFILE must be"}},
		{"missing origins", map[string]string{"IDENTITY_PUBLIC_ORIGINS": "<unset>"}, "development", []string{"IDENTITY_PUBLIC_ORIGINS is required"}},
		{"http origin under https", map[string]string{"IDENTITY_PUBLIC_ORIGINS": "http://org-a.localhost"}, "development", []string{"exact https origins"}},
		{"origin with path", map[string]string{"IDENTITY_PUBLIC_ORIGINS": "https://org-a.localhost/x"}, "development", []string{"exact https origins"}},
		{"origin with userinfo", map[string]string{"IDENTITY_PUBLIC_ORIGINS": "https://u@org-a.localhost"}, "development", []string{"exact https origins"}},
		{"wildcard origin", map[string]string{"IDENTITY_PUBLIC_ORIGINS": "https://*.localhost"}, "development", []string{"exact https origins"}},
		{"non-loopback http under loopback-http", map[string]string{"IDENTITY_COOKIE_PROFILE": "loopback-http", "IDENTITY_PUBLIC_ORIGINS": "http://org-a.localhost"}, "development", []string{"exact https origins"}},
		{"short lifetime outside test", map[string]string{"IDENTITY_SESSION_IDLE": "30s"}, "production", []string{"between 5m and 24h"}},
		{"long lifetime", map[string]string{"IDENTITY_SESSION_ABSOLUTE": "48h"}, "development", []string{"between 5m and 24h"}},
		{"idle beyond absolute", map[string]string{"IDENTITY_SESSION_IDLE": "9h"}, "development", []string{"must not exceed"}},
		{"malformed duration", map[string]string{"IDENTITY_MAX_AUTH_AGE": "soon"}, "development", []string{"IDENTITY_MAX_AUTH_AGE must be a positive duration"}},
		{"symmetric algorithm", map[string]string{"IDENTITY_JWT_ALGORITHMS": "HS256"}, "development", []string{"only asymmetric"}},
		{"none algorithm", map[string]string{"IDENTITY_JWT_ALGORITHMS": "RS256,none"}, "development", []string{"only asymmetric"}},
		{"provider other than workos", map[string]string{"IDENTITY_PROVIDER": "keycloak"}, "development", []string{"IDENTITY_PROVIDER must be workos"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			id, err := LoadIdentity(identityEnv(c.env), c.appEnv)
			if err == nil {
				t.Fatalf("accepted: %+v", id)
			}
			msg := err.Error()
			if !strings.HasPrefix(msg, "invalid configuration: ") {
				t.Fatalf("message lost its prefix: %q", msg)
			}
			for _, w := range c.want {
				if !strings.Contains(msg, w) {
					t.Errorf("message %q lacks %q", msg, w)
				}
			}
			// Names only: no configured value appears in the error.
			for _, v := range []string{testKey, testClient, testIssuer, "emulator.localhost", "org-a.localhost", "keycloak"} {
				if strings.Contains(msg, v) {
					t.Errorf("message %q contains a configured value", msg)
				}
			}
			if id.Configured() || id.APIKey != "" {
				t.Errorf("a refused configuration returned values: %+v", id)
			}
		})
	}
}

func TestLoadIdentityTestEnvironmentAllowances(t *testing.T) {
	id, err := LoadIdentity(identityEnv(map[string]string{
		"IDENTITY_WORKOS_API_BASE_URL": "https://workos-emulator.localhost/",
		"IDENTITY_WORKOS_ISSUER":       "http://workos-emulator.localhost:4000",
		"IDENTITY_SESSION_IDLE":        "20s",
		"IDENTITY_SESSION_ABSOLUTE":    "90s",
		"IDENTITY_MAX_AUTH_AGE":        "2m",
		"IDENTITY_JWT_ALGORITHMS":      "RS256, ES256",
	}), "test")
	if err != nil {
		t.Fatal(err)
	}
	if id.APIBaseURL != "https://workos-emulator.localhost" || id.SessionIdle != 20*time.Second ||
		id.MaxAuthAge != 2*time.Minute || !slices.Equal(id.Algorithms, []string{"RS256", "ES256"}) {
		t.Fatalf("test allowances: %+v", id)
	}
	// An http issuer is a test-environment allowance only.
	if _, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_WORKOS_ISSUER": "http://workos-emulator.localhost:4000"}), "development"); err == nil {
		t.Fatal("http issuer accepted outside APP_ENV=test")
	}
}

// TestLoadIdentityCookieProfileGate is core's half of the double gate on http origins (checkpoint C4 item 7): web's
// non-production loopback default cannot enable an http provider flow, because configured core defaults to https and
// accepts http origins only under an explicit non-production loopback-http profile, for loopback hosts only.
func TestLoadIdentityCookieProfileGate(t *testing.T) {
	loopback := []string{"http://localhost:3100", "http://127.0.0.1:3100"}
	for _, appEnv := range []string{"development", "test"} {
		for _, o := range loopback {
			for _, profile := range []string{"<unset>", "https"} {
				if _, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_COOKIE_PROFILE": profile, "IDENTITY_PUBLIC_ORIGINS": o}), appEnv); err == nil ||
					!strings.Contains(err.Error(), "IDENTITY_PUBLIC_ORIGINS must list exact https origins") {
					t.Errorf("%s, profile %s: http origin %s not refused: %v", appEnv, profile, o, err)
				}
			}
			id, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_COOKIE_PROFILE": "loopback-http", "IDENTITY_PUBLIC_ORIGINS": o + ",https://org-a.localhost:543"}), appEnv)
			if err != nil || id.CookieProfile != "loopback-http" || !slices.Equal(id.PublicOrigins, []string{o, "https://org-a.localhost:543"}) {
				t.Errorf("%s: explicit loopback-http with %s: %+v %v", appEnv, o, id, err)
			}
		}
		for _, o := range []string{"http://org-a.localhost", "http://localhost.example:3100", "http://127.0.0.2:3100", "http://[::1]:3100", "http://LOCALHOST:3100"} {
			if _, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_COOKIE_PROFILE": "loopback-http", "IDENTITY_PUBLIC_ORIGINS": o}), appEnv); err == nil {
				t.Errorf("%s: non-loopback http origin %s accepted under loopback-http", appEnv, o)
			}
		}
	}
	// Production refuses the relaxation whatever the origins.
	for _, o := range append(loopback, "https://org-a.localhost:543") {
		if _, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_COOKIE_PROFILE": "loopback-http", "IDENTITY_PUBLIC_ORIGINS": o}), "production"); err == nil ||
			!strings.Contains(err.Error(), "IDENTITY_COOKIE_PROFILE loopback-http is refused in production") {
			t.Errorf("production loopback-http with %s: %v", o, err)
		}
	}
}

func TestLoadIdentityKeyFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "key")
	if err := os.WriteFile(p, []byte(testKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	id, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_WORKOS_API_KEY": "<unset>", "IDENTITY_WORKOS_API_KEY_FILE": p}), "production")
	if err != nil || id.APIKey != testKey {
		t.Fatalf("key file: %v", err)
	}
	empty := filepath.Join(dir, "empty")
	if err := os.WriteFile(empty, []byte("\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_WORKOS_API_KEY": "<unset>", "IDENTITY_WORKOS_API_KEY_FILE": empty}), "development"); err == nil {
		t.Fatal("empty key file accepted")
	}
}

// An unrelated MFA-labelled variable has no effect: there is no MFA configuration (M2); unknown names are ignored.
func TestLoadIdentityIgnoresMFAConfiguration(t *testing.T) {
	id, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_MFA_POLICY": "provider-required", "IDENTITY_MFA_METHODS": "totp"}), "production")
	if err != nil {
		t.Fatal(err)
	}
	if !id.Configured() {
		t.Fatal("configuration lost")
	}
}
