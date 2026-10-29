package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// G016 (Plan r4 §5.4): the domain variables, their defaults and bounds, the test-only fixtures and the non-tenant rule
// for IDENTITY_PUBLIC_ORIGINS. Messages name variables only.

func domainsBase(appEnv string) Config {
	return Config{PlatformBaseDomain: "example.net", CustomDomainCNAMETarget: "custom.example.net", AppEnv: appEnv}
}

func domainsEnv(over map[string]string) Getenv {
	base := map[string]string{}
	for k, v := range over {
		base[k] = v
	}
	return func(k string) string { return base[k] }
}

func authIdentity(t *testing.T, origins string) Identity {
	t.Helper()
	id, err := LoadIdentity(identityEnv(map[string]string{"IDENTITY_PUBLIC_ORIGINS": origins}), "test")
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestLoadDomainsConfig(t *testing.T) {
	d, err := LoadDomains(domainsEnv(nil), domainsBase("development"), Identity{})
	if err != nil {
		t.Fatal(err)
	}
	if d.ChallengeLabel != "_platform-verify" || d.ChallengeTTL != 168*time.Hour || d.ProofFreshness != 720*time.Hour ||
		d.DNSTimeout != 5*time.Second || d.TLSTimeout != 5*time.Second || d.TLSAdapter != "edge-probe" ||
		d.TLSEdgeAddr != "custom.example.net:443" || d.WithdrawalRetry != 5 || d.AuthHost != "" || d.AuthOrigin() != "" ||
		d.HostPort != "" || d.RootOrigin() != "https://example.net" {
		t.Fatalf("defaults: %+v", d)
	}
	if hc := d.Hostnames(); hc.BaseDomain != "example.net" || hc.CNAMETarget != "custom.example.net" {
		t.Fatalf("hostnames: %+v", hc)
	}

	dir := t.TempDir()
	fixture := filepath.Join(dir, "dns.tsv")
	if err := os.WriteFile(fixture, []byte("x\tTXT\ty\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	d, err = LoadDomains(domainsEnv(map[string]string{
		"PLATFORM_HOST_PORT": "543", "PLATFORM_AUTH_HOST": "auth.example.net",
		"PLATFORM_RESERVED_LABELS": "blog, docs,blog", "CUSTOM_DOMAIN_CHALLENGE_TTL": "1h", "CUSTOM_DOMAIN_PROOF_FRESHNESS": "2160h",
		"CUSTOM_DOMAIN_DNS_TIMEOUT": "15s", "DOMAIN_TLS_ADAPTER": "fixture", "DOMAIN_TLS_FIXTURE_FILE": fixture,
		"CUSTOM_DOMAIN_DNS_FIXTURE_FILE": fixture, "DOMAIN_TLS_EDGE_ADDR": "proxy:543", "DOMAIN_WITHDRAWAL_RETRY_MAX": "20",
	}), domainsBase("test"), authIdentity(t, "https://auth.example.net:543"))
	if err != nil {
		t.Fatal(err)
	}
	if d.AuthOrigin() != "https://auth.example.net:543" || len(d.ReservedLabels) != 2 || d.ChallengeTTL != time.Hour ||
		d.ProofFreshness != 2160*time.Hour || d.TLSAdapter != "fixture" || len(d.TLSFixture) == 0 || len(d.DNSFixture) == 0 ||
		d.TLSEdgeAddr != "proxy:543" || d.WithdrawalRetry != 20 || d.RootOrigin() != "https://example.net:543" {
		t.Fatalf("explicit: %+v", d)
	}

	for name, c := range map[string]struct {
		env    map[string]string
		appEnv string
		id     Identity
		want   string
	}{
		"host port in production":           {map[string]string{"PLATFORM_HOST_PORT": "543"}, "production", Identity{}, "PLATFORM_HOST_PORT is refused in production"},
		"host port malformed":               {map[string]string{"PLATFORM_HOST_PORT": "0543"}, "test", Identity{}, "PLATFORM_HOST_PORT must be"},
		"reserved label uppercase":          {map[string]string{"PLATFORM_RESERVED_LABELS": "Blog"}, "test", Identity{}, "PLATFORM_RESERVED_LABELS must"},
		"auth host malformed":               {map[string]string{"PLATFORM_AUTH_HOST": "https://auth.example.net"}, "test", Identity{}, "PLATFORM_AUTH_HOST must be"},
		"auth host equals target":           {map[string]string{"PLATFORM_AUTH_HOST": "custom.example.net"}, "test", Identity{}, "PLATFORM_AUTH_HOST must differ"},
		"auth host not public":              {map[string]string{"PLATFORM_AUTH_HOST": "login.example.net"}, "test", authIdentity(t, "https://auth.example.net"), "PLATFORM_AUTH_HOST must be listed"},
		"tenant origin public":              {map[string]string{"PLATFORM_AUTH_HOST": "auth.example.net"}, "test", authIdentity(t, "https://auth.example.net,https://org-a.example.net"), "IDENTITY_PUBLIC_ORIGINS may list only non-tenant origins"},
		"custom origin public":              {nil, "test", authIdentity(t, "https://portal.acme.test"), "IDENTITY_PUBLIC_ORIGINS may list only non-tenant origins"},
		"public origin foreign port":        {nil, "test", authIdentity(t, "https://auth.example.net:8443"), "IDENTITY_PUBLIC_ORIGINS must use PLATFORM_HOST_PORT"},
		"challenge label no underscore":     {map[string]string{"CUSTOM_DOMAIN_CHALLENGE_LABEL": "verify"}, "test", Identity{}, "CUSTOM_DOMAIN_CHALLENGE_LABEL must"},
		"challenge ttl too long":            {map[string]string{"CUSTOM_DOMAIN_CHALLENGE_TTL": "721h"}, "test", Identity{}, "CUSTOM_DOMAIN_CHALLENGE_TTL must be"},
		"freshness too short":               {map[string]string{"CUSTOM_DOMAIN_PROOF_FRESHNESS": "23h"}, "test", Identity{}, "CUSTOM_DOMAIN_PROOF_FRESHNESS must be"},
		"dns timeout malformed":             {map[string]string{"CUSTOM_DOMAIN_DNS_TIMEOUT": "5"}, "test", Identity{}, "CUSTOM_DOMAIN_DNS_TIMEOUT must be"},
		"tls timeout too long":              {map[string]string{"DOMAIN_TLS_TIMEOUT": "16s"}, "test", Identity{}, "DOMAIN_TLS_TIMEOUT must be"},
		"retry out of range":                {map[string]string{"DOMAIN_WITHDRAWAL_RETRY_MAX": "21"}, "test", Identity{}, "DOMAIN_WITHDRAWAL_RETRY_MAX must be"},
		"tls adapter unknown":               {map[string]string{"DOMAIN_TLS_ADAPTER": "cloudflare"}, "test", Identity{}, "DOMAIN_TLS_ADAPTER must be"},
		"tls fixture adapter in production": {map[string]string{"DOMAIN_TLS_ADAPTER": "fixture"}, "production", Identity{}, "DOMAIN_TLS_ADAPTER fixture is accepted only with APP_ENV=test"},
		"tls fixture file missing":          {map[string]string{"DOMAIN_TLS_ADAPTER": "fixture"}, "test", Identity{}, "DOMAIN_TLS_FIXTURE_FILE is required"},
		"dns fixture in production":         {map[string]string{"CUSTOM_DOMAIN_DNS_FIXTURE_FILE": fixture}, "production", Identity{}, "CUSTOM_DOMAIN_DNS_FIXTURE_FILE is accepted only with APP_ENV=test"},
		"dns fixture in development":        {map[string]string{"CUSTOM_DOMAIN_DNS_FIXTURE_FILE": fixture}, "development", Identity{}, "CUSTOM_DOMAIN_DNS_FIXTURE_FILE is accepted only with APP_ENV=test"},
		"dns fixture unreadable":            {map[string]string{"CUSTOM_DOMAIN_DNS_FIXTURE_FILE": "/nonexistent/dns.tsv"}, "test", Identity{}, "CUSTOM_DOMAIN_DNS_FIXTURE_FILE cannot be read"},
		"edge addr malformed":               {map[string]string{"DOMAIN_TLS_EDGE_ADDR": "proxy"}, "test", Identity{}, "DOMAIN_TLS_EDGE_ADDR must be"},
	} {
		_, err := LoadDomains(domainsEnv(c.env), domainsBase(c.appEnv), c.id)
		if err == nil || !strings.HasPrefix(err.Error(), "invalid configuration: ") || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want a refusal naming %q", name, err, c.want)
			continue
		}
		for _, v := range []string{"www.example.org", "portal.acme.test", "/nonexistent", "cloudflare", "Blog"} {
			if strings.Contains(err.Error(), v) {
				t.Errorf("%s: the message carries a value (%s)", name, v)
			}
		}
	}

	// G071 (Plan r4 S1, S15; U1): DEFAULT_TENANT_SLUG and PLATFORM_ROOT_PAGE are never a configuration error in any
	// environment. Each is read trimmed (empty = unset) and kept as given otherwise: the slug is judged at startup by
	// domains.SelectDefaultTenant and the presentation resolved by domains.RootPresentation, never here.
	for _, appEnv := range []string{"production", "development", "test"} {
		for _, c := range []struct {
			slug, page         string
			wantSlug, wantPage string
		}{
			{"", "", "", ""},
			{"   ", " \t ", "", ""},
			{"alpha-agency", "signup_redirect", "alpha-agency", "signup_redirect"},
			{"  alpha-agency  ", " landing ", "alpha-agency", "landing"},
			{"Alpha-Agency", "Landing", "Alpha-Agency", "Landing"},
			{"not_a_slug", "redirect", "not_a_slug", "redirect"},
			{"www", "https://x", "www", "https://x"},
			{"a--b", "bogus", "a--b", "bogus"},
		} {
			d, err := LoadDomains(domainsEnv(map[string]string{"DEFAULT_TENANT_SLUG": c.slug, "PLATFORM_ROOT_PAGE": c.page}),
				domainsBase(appEnv), Identity{})
			if err != nil {
				t.Errorf("%s: DEFAULT_TENANT_SLUG=%q PLATFORM_ROOT_PAGE=%q refused: %v", appEnv, c.slug, c.page, err)
				continue
			}
			if d.DefaultTenantSlug != c.wantSlug || d.RootPage != c.wantPage {
				t.Errorf("%s: %q/%q loaded as %q/%q, want %q/%q", appEnv, c.slug, c.page, d.DefaultTenantSlug, d.RootPage,
					c.wantSlug, c.wantPage)
			}
		}
	}
	// Neither setting turns another variable's refusal into an acceptance, nor appears in a refusal message.
	if _, err := LoadDomains(domainsEnv(map[string]string{"DEFAULT_TENANT_SLUG": "alpha-agency", "PLATFORM_ROOT_PAGE": "landing",
		"PLATFORM_HOST_PORT": "543"}), domainsBase("production"), Identity{}); err == nil ||
		!strings.Contains(err.Error(), "PLATFORM_HOST_PORT is refused in production") ||
		strings.Contains(err.Error(), "DEFAULT_TENANT_SLUG") || strings.Contains(err.Error(), "PLATFORM_ROOT_PAGE") ||
		strings.Contains(err.Error(), "alpha-agency") || strings.Contains(err.Error(), "landing") {
		t.Errorf("a refusal next to the G071 settings: %v", err)
	}
	vars := DomainVariables()
	for _, name := range []string{"DEFAULT_TENANT_SLUG", "PLATFORM_ROOT_PAGE"} {
		if !slices.Contains(vars, name) {
			t.Errorf("DomainVariables does not list %s: %v", name, vars)
		}
	}
	if !slices.IsSorted(vars) {
		t.Errorf("DomainVariables is not sorted: %v", vars)
	}

	// The overlay's single-label platform domain is accepted with APP_ENV=test only.
	if _, err := Load(func(k string) string { return overlayEnv(k, "test") }, EnvSecrets{Getenv: func(k string) string { return overlayEnv(k, "test") }}); err != nil {
		t.Fatalf("localhost platform domain under test: %v", err)
	}
	if _, err := Load(func(k string) string { return overlayEnv(k, "development") }, EnvSecrets{Getenv: func(k string) string { return overlayEnv(k, "development") }}); err == nil ||
		!strings.Contains(err.Error(), "PLATFORM_BASE_DOMAIN must be a lowercase hostname") {
		t.Fatalf("localhost platform domain outside test: %v", err)
	}
}

func overlayEnv(k, appEnv string) string {
	return map[string]string{
		"PRODUCT_NAME": "Probe", "PLATFORM_BASE_DOMAIN": "localhost", "CUSTOM_DOMAIN_CNAME_TARGET": "custom.localhost",
		"DATABASE_NAME": "probe", "APP_ENV": appEnv, "DB_HOST": "postgres", "DB_PORT": "5432", "DB_USER": "probe_runtime",
		"DB_PASSWORD": "x", "REDIS_HOST": "redis", "REDIS_PORT": "6379", "REDIS_PASSWORD": "x",
		"OBJECT_STORE_ENDPOINT": "http://minio:9000", "OBJECT_STORE_REGION": "us-east-1", "OBJECT_STORE_BUCKET": "b",
		"OBJECT_STORE_FORCE_PATH_STYLE": "true", "OBJECT_STORE_ACCESS_KEY": "a", "OBJECT_STORE_SECRET_KEY": "s",
	}[k]
}

// TestMarketingSiteURLRemoved (G062 Plan r5 §4.1, D-6): MARKETING_SITE_URL is no longer a domain variable; production
// loads without it, and an environment that still sets it (well-formed or not) is ignored. The root origin follows
// PLATFORM_BASE_DOMAIN only.
func TestMarketingSiteURLRemoved(t *testing.T) {
	for _, v := range DomainVariables() {
		if v == "MARKETING_SITE_URL" {
			t.Fatal("MARKETING_SITE_URL is still a domain variable")
		}
	}
	for _, env := range []map[string]string{nil, {"MARKETING_SITE_URL": "https://www.example.org/"},
		{"MARKETING_SITE_URL": "/relative"}, {"MARKETING_SITE_URL": "http://www.example.org"}} {
		d, err := LoadDomains(domainsEnv(env), domainsBase("production"), Identity{})
		if err != nil {
			t.Fatalf("%v: %v", env, err)
		}
		if d.RootOrigin() != "https://example.net" {
			t.Errorf("%v: root origin %q", env, d.RootOrigin())
		}
	}
}
