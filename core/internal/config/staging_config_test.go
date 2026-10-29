package config

import (
	"slices"
	"strings"
	"testing"
)

// G070 (Plan §2.2, r1.1 D-C1): the synthetic staging profile through all four loaders, in core's startup order (Load,
// LoadIdentity, LoadMCP, LoadDomains). stagingCoreTemplate is the environment infra/compose/staging.yml hands core for
// the values of infra/staging/examples/app.env.example (kept in step by review; this test's build context holds core/
// only): APP_ENV=development with every production-equivalent setting explicit, dummy secrets. G-1 is executable: the
// same template under APP_ENV=production is refused for its private http object store.

const stagingCoreTemplate = `
PRODUCT_NAME=Example Product
PLATFORM_BASE_DOMAIN=staging.example.net
CUSTOM_DOMAIN_CNAME_TARGET=custom.staging.example.net
DATABASE_NAME=example_app
APP_ENV=development
CORE_LISTEN_ADDR=:8180
DB_HOST=postgres
DB_PORT=5432
DB_USER=example_app_runtime
DB_PASSWORD=db-secret-value
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=redis-secret-value
OBJECT_STORE_ENDPOINT=http://minio:9000
OBJECT_STORE_REGION=us-east-1
OBJECT_STORE_BUCKET=example-objects
OBJECT_STORE_FORCE_PATH_STYLE=true
OBJECT_STORE_ACCESS_KEY=core-service
OBJECT_STORE_SECRET_KEY=object-secret-value
INTELLIGENCE_URL=http://intelligence:8000
INTELLIGENCE_SERVICE_TOKEN=intelligence-secret-value
PLATFORM_AUTH_HOST=auth.staging.example.net
IDENTITY_PROVIDER=workos
IDENTITY_COOKIE_PROFILE=https
DOMAIN_TLS_EDGE_ADDR=edge:443
`

// stagingIdentityTemplate is the optional identity triple (dummy values) plus the origin the VM renders with it.
const stagingIdentityTemplate = `
IDENTITY_WORKOS_CLIENT_ID=client_01STAGINGEXAMPLE
IDENTITY_WORKOS_ISSUER=https://api.workos.com/user_management/client_01STAGINGEXAMPLE
IDENTITY_WORKOS_API_KEY=workos-secret-value
IDENTITY_PUBLIC_ORIGINS=https://auth.staging.example.net
`

func stagingEnv(t *testing.T, blocks ...string) map[string]string {
	t.Helper()
	env := map[string]string{}
	for _, b := range blocks {
		for _, line := range strings.Split(b, "\n") {
			if line = strings.TrimSpace(line); line == "" {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				t.Fatalf("template line without '=': %q", line)
			}
			env[k] = v
		}
	}
	return env
}

// loadStaging runs the four loaders in core's startup order and stops at the first refusal.
func loadStaging(env map[string]string) (Config, Identity, MCP, Domains, error) {
	get := func(k string) string { return env[k] }
	c, err := Load(get, EnvSecrets{Getenv: get})
	if err != nil {
		return Config{}, Identity{}, MCP{}, Domains{}, err
	}
	id, err := LoadIdentity(get, c.AppEnv)
	if err != nil {
		return Config{}, Identity{}, MCP{}, Domains{}, err
	}
	m, err := LoadMCP(get, c.AppEnv, id)
	if err != nil {
		return Config{}, Identity{}, MCP{}, Domains{}, err
	}
	d, err := LoadDomains(get, c, id)
	if err != nil {
		return Config{}, Identity{}, MCP{}, Domains{}, err
	}
	return c, id, m, d, nil
}

func TestStagingTemplateCarriesNoTestOnlySetting(t *testing.T) {
	env := stagingEnv(t, stagingCoreTemplate, stagingIdentityTemplate)
	if env["APP_ENV"] != "development" {
		t.Fatalf("staging APP_ENV = %q, want development", env["APP_ENV"])
	}
	for k := range env {
		if strings.HasPrefix(k, "IDENTITY_TEST_") || slices.Contains([]string{"PLATFORM_HOST_PORT", "IDENTITY_WORKOS_API_BASE_URL",
			"CUSTOM_DOMAIN_DNS_FIXTURE_FILE", "DOMAIN_TLS_ADAPTER", "DOMAIN_TLS_FIXTURE_FILE", "MCP_CONNECT_JWKS_FILE"}, k) {
			t.Errorf("the staging template sets the test-only or fixture variable %s", k)
		}
	}
}

func TestStagingTemplateIdentityUnconfigured(t *testing.T) {
	c, id, m, d, err := loadStaging(stagingEnv(t, stagingCoreTemplate))
	if err != nil {
		t.Fatal(err)
	}
	if c.AppEnv != "development" || c.ListenAddr != ":8180" || c.DB.User != "example_app_runtime" ||
		c.ObjectStore.Endpoint != "http://minio:9000" || c.ObjectStore.SigningEndpoint != "" || !c.Intelligence.Enabled() ||
		len(c.SecretValues()) != 4 {
		t.Fatalf("config: app env %q, listen %q, db user %q, object endpoint %q, intelligence %v, %d secrets", c.AppEnv,
			c.ListenAddr, c.DB.User, c.ObjectStore.Endpoint, c.Intelligence.Enabled(), len(c.SecretValues()))
	}
	// Identity unconfigured: sign-in is unavailable (never bypassed); development starts, production would refuse.
	if id.Configured() || id.SecretValues() != nil || m.Configured() {
		t.Fatalf("identity configured %v, MCP configured %v", id.Configured(), m.Configured())
	}
	if d.BaseDomain != "staging.example.net" || d.CNAMETarget != "custom.staging.example.net" ||
		d.AuthHost != "auth.staging.example.net" || d.AuthOrigin() != "https://auth.staging.example.net" || d.HostPort != "" ||
		d.RootOrigin() != "https://staging.example.net" || d.TLSAdapter != "edge-probe" || d.TLSEdgeAddr != "edge:443" ||
		d.DNSFixture != nil || d.TLSFixture != nil {
		t.Fatalf("domains: %+v", d)
	}
}

func TestStagingTemplateIdentityConfigured(t *testing.T) {
	_, id, m, d, err := loadStaging(stagingEnv(t, stagingCoreTemplate, stagingIdentityTemplate))
	if err != nil {
		t.Fatal(err)
	}
	if !id.Configured() || id.CookieProfile != "https" || id.APIBaseURL != IdentityDefaultAPIBase ||
		!slices.Equal(id.PublicOrigins, []string{"https://auth.staging.example.net"}) || len(id.SecretValues()) != 1 {
		t.Fatalf("identity: configured %v, cookie profile %q, API base %q, origins %v", id.Configured(), id.CookieProfile,
			id.APIBaseURL, id.PublicOrigins)
	}
	if m.Configured() {
		t.Fatal("MCP must stay disabled in the staging template")
	}
	if d.AuthOrigin() != "https://auth.staging.example.net" || d.TLSAdapter != "edge-probe" || d.TLSEdgeAddr != "edge:443" {
		t.Fatalf("domains: %+v", d)
	}
}

// TestStagingProductionRefusesHTTPObjectStore pins G-1: the private MinIO endpoint (http://minio:9000) is refused under
// APP_ENV=production, which is why staging runs the development loader (D-C1) and is not production readiness.
func TestStagingProductionRefusesHTTPObjectStore(t *testing.T) {
	env := stagingEnv(t, stagingCoreTemplate, stagingIdentityTemplate)
	env["APP_ENV"] = "production"
	_, _, _, _, err := loadStaging(env)
	if err == nil {
		t.Fatal("APP_ENV=production with the staging object store must be refused")
	}
	for _, want := range []string{"OBJECT_STORE_ENDPOINT must be https with APP_ENV=production",
		"OBJECT_STORE_SIGNING_ENDPOINT must be https with APP_ENV=production"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q lacks %q", err, want)
		}
	}
	for _, secret := range []string{"db-secret-value", "redis-secret-value", "object-secret-value", "intelligence-secret-value"} {
		if strings.Contains(err.Error(), secret) {
			t.Error("the refusal leaks a secret value")
		}
	}
}

func TestStagingProductionRefusesUnconfiguredIdentity(t *testing.T) {
	env := stagingEnv(t, stagingCoreTemplate)
	_, err := LoadIdentity(func(k string) string { return env[k] }, "production")
	if err == nil || !strings.Contains(err.Error(), "IDENTITY_WORKOS_CLIENT_ID is required in production") {
		t.Fatalf("production without identity: %v", err)
	}
}

// TestStagingTemplateRootSettings: the two optional G071 root host settings the VM renders from staging.conf (an empty
// value as KEY=, unset) reach LoadDomains unchanged through the four loaders. LoadDomains never refuses them: an
// invalid PLATFORM_ROOT_PAGE is kept raw for domains.RootPresentation, which warns and uses the default at startup.
func TestStagingTemplateRootSettings(t *testing.T) {
	for _, c := range []struct{ slug, page string }{
		{"", ""},
		{"", "landing"},
		{"alpha-agency", "signup_redirect"},
		{"", "redirect"},
	} {
		env := stagingEnv(t, stagingCoreTemplate, "DEFAULT_TENANT_SLUG="+c.slug+"\nPLATFORM_ROOT_PAGE="+c.page)
		_, _, _, d, err := loadStaging(env)
		if err != nil {
			t.Errorf("DEFAULT_TENANT_SLUG=%q PLATFORM_ROOT_PAGE=%q refused: %v", c.slug, c.page, err)
			continue
		}
		if d.DefaultTenantSlug != c.slug || d.RootPage != c.page || d.TLSAdapter != "edge-probe" {
			t.Errorf("%q/%q loaded as %q/%q (TLS adapter %q)", c.slug, c.page, d.DefaultTenantSlug, d.RootPage, d.TLSAdapter)
		}
	}
}

// Each test-only knob added to the development staging template is refused by the loaders (no test overlay can be
// slipped into the public environment through configuration).
func TestStagingTemplateRefusesTestOnlySettings(t *testing.T) {
	for _, c := range []struct{ name, value, want string }{
		{"DOMAIN_TLS_ADAPTER", "fixture", "DOMAIN_TLS_ADAPTER fixture is accepted only with APP_ENV=test"},
		{"CUSTOM_DOMAIN_DNS_FIXTURE_FILE", "/etc/domains/dns.tsv", "CUSTOM_DOMAIN_DNS_FIXTURE_FILE is accepted only with APP_ENV=test"},
		{"DOMAIN_TLS_FIXTURE_FILE", "/etc/domains/tls.tsv", "DOMAIN_TLS_FIXTURE_FILE is accepted only with APP_ENV=test"},
		{"IDENTITY_WORKOS_API_BASE_URL", "http://emulator:4100", "IDENTITY_WORKOS_API_BASE_URL may differ from the default only with APP_ENV=test"},
		{"MCP_CONNECT_JWKS_FILE", "/identity/mcp/connect-jwks.json", "MCP_CONNECT_JWKS_FILE is accepted only with APP_ENV=test"},
	} {
		env := stagingEnv(t, stagingCoreTemplate, stagingIdentityTemplate)
		env[c.name] = c.value
		if _, _, _, _, err := loadStaging(env); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s=%s: got %v, want a refusal containing %q", c.name, c.value, err, c.want)
		}
	}
}
