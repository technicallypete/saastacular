package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// U-MCP-CONFIG (G014 Plan r8 §3.6, §10.1): the documented disabled state, the defaults and bounds, and every startup
// refusal, including the test trust anchor (C10) outside APP_ENV=test. Messages name variables only.

const (
	mcpOrgA     = "4a000000-0000-4000-8000-00000000000a"
	mcpIssuer   = "https://example-env.authkit.app"
	mcpRegistry = "# retired trusted tenant hosts\norg-a.localhost:543\t" + mcpOrgA + "\n"
)

func mcpWriteFile(t *testing.T, name, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func mcpEnv(t *testing.T, over map[string]string) Getenv {
	t.Helper()
	base := map[string]string{"MCP_CONNECT_ISSUER": mcpIssuer}
	for k, v := range over {
		if v == "<unset>" {
			delete(base, k)
		} else {
			base[k] = v
		}
	}
	return func(k string) string { return base[k] }
}

func mustIdentity(t *testing.T) Identity {
	t.Helper()
	id, err := LoadIdentity(identityEnv(nil), "test")
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestLoadMCPDisabled(t *testing.T) {
	m, err := LoadMCP(func(string) string { return "" }, "production", Identity{})
	if err != nil || m.Configured() {
		t.Fatalf("all unset must be the disabled state: %+v %v", m, err)
	}
}

func TestLoadMCPDefaults(t *testing.T) {
	m, err := LoadMCP(mcpEnv(t, nil), "development", mustIdentity(t))
	if err != nil {
		t.Fatal(err)
	}
	if !m.Configured() || m.Issuer != mcpIssuer || m.PendingTTL != 30*time.Minute || m.Idle != 30*24*time.Hour ||
		m.Absolute != 180*24*time.Hour || len(m.AllowedOrigins) != 0 || m.TestJWKS != nil || len(m.Algorithms) != 1 {
		t.Fatalf("defaults: %+v", m)
	}
	m, err = LoadMCP(mcpEnv(t, map[string]string{"MCP_PENDING_TTL": "2m", "MCP_CONNECTION_IDLE": "24h", "MCP_CONNECTION_ABSOLUTE": "17520h",
		"MCP_ALLOWED_ORIGINS": "https://client.example, https://client.example"}), "production", mustIdentity(t))
	if err != nil || m.PendingTTL != 2*time.Minute || m.Idle != 24*time.Hour || m.Absolute != 730*24*time.Hour || len(m.AllowedOrigins) != 1 {
		t.Fatalf("bounds: %+v %v", m, err)
	}
}

func TestLoadMCPRefusals(t *testing.T) {
	jwks := mcpWriteFile(t, "jwks.json", `{"keys":[{"kty":"RSA"}]}`)
	for name, c := range map[string]struct {
		env    map[string]string
		appEnv string
		id     func(*testing.T) Identity
		want   string
	}{
		"partial: no issuer":         {map[string]string{"MCP_CONNECT_ISSUER": "<unset>", "MCP_PENDING_TTL": "30m"}, "test", mustIdentity, "MCP_CONNECT_ISSUER is required"},
		"partial: only lifetimes":    {map[string]string{"MCP_CONNECT_ISSUER": "<unset>", "MCP_PENDING_TTL": "5m"}, "test", mustIdentity, "MCP_CONNECT_ISSUER is required"},
		"MCP without identity":       {nil, "test", func(*testing.T) Identity { return Identity{} }, "MCP requires identity"},
		"http issuer in production":  {map[string]string{"MCP_CONNECT_ISSUER": "http://example-env.authkit.app"}, "production", mustIdentity, "MCP_CONNECT_ISSUER must be"},
		"issuer with query":          {map[string]string{"MCP_CONNECT_ISSUER": mcpIssuer + "?x=1"}, "test", mustIdentity, "MCP_CONNECT_ISSUER must be"},
		"issuer with trailing slash": {map[string]string{"MCP_CONNECT_ISSUER": mcpIssuer + "/"}, "test", mustIdentity, "MCP_CONNECT_ISSUER must be"},
		"retired registry (test)":    {map[string]string{"MCP_TENANT_HOSTS_FILE": mcpWriteFile(t, "h.tsv", mcpRegistry)}, "test", mustIdentity, "MCP_TENANT_HOSTS_FILE is retired"},
		"retired registry (prod)":    {map[string]string{"MCP_TENANT_HOSTS_FILE": "/nonexistent/hosts.tsv"}, "production", mustIdentity, "MCP_TENANT_HOSTS_FILE is retired"},
		"retired registry alone":     {map[string]string{"MCP_CONNECT_ISSUER": "<unset>", "MCP_TENANT_HOSTS_FILE": "/x"}, "development", mustIdentity, "MCP_TENANT_HOSTS_FILE is retired"},
		"pending too short":          {map[string]string{"MCP_PENDING_TTL": "1m"}, "test", mustIdentity, "MCP_PENDING_TTL is outside"},
		"pending too long":           {map[string]string{"MCP_PENDING_TTL": "61m"}, "test", mustIdentity, "MCP_PENDING_TTL is outside"},
		"idle too long":              {map[string]string{"MCP_CONNECTION_IDLE": "2161h"}, "test", mustIdentity, "MCP_CONNECTION_IDLE is outside"},
		"absolute malformed":         {map[string]string{"MCP_CONNECTION_ABSOLUTE": "180d"}, "test", mustIdentity, "MCP_CONNECTION_ABSOLUTE is outside"},
		"idle above absolute":        {map[string]string{"MCP_CONNECTION_IDLE": "720h", "MCP_CONNECTION_ABSOLUTE": "48h"}, "test", mustIdentity, "must not exceed"},
		"allowed origin http":        {map[string]string{"MCP_ALLOWED_ORIGINS": "http://client.example"}, "test", mustIdentity, "MCP_ALLOWED_ORIGINS must list"},
		"test anchor in production":  {map[string]string{"MCP_CONNECT_JWKS_FILE": jwks}, "production", mustIdentity, "MCP_CONNECT_JWKS_FILE is accepted only with APP_ENV=test"},
		"test anchor in development": {map[string]string{"MCP_CONNECT_JWKS_FILE": jwks}, "development", mustIdentity, "MCP_CONNECT_JWKS_FILE is accepted only with APP_ENV=test"},
		"unreadable test anchor":     {map[string]string{"MCP_CONNECT_JWKS_FILE": "/nonexistent/jwks.json"}, "test", mustIdentity, "MCP_CONNECT_JWKS_FILE cannot be read"},
	} {
		_, err := LoadMCP(mcpEnv(t, c.env), c.appEnv, c.id(t))
		if err == nil || !strings.HasPrefix(err.Error(), "invalid configuration: ") || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want a refusal naming %q", name, err, c.want)
			continue
		}
		for _, v := range []string{mcpIssuer + "?", mcpOrgA, "org-c.localhost", "/nonexistent", "client.example"} {
			if strings.Contains(err.Error(), v) {
				t.Errorf("%s: the message carries a value (%s)", name, v)
			}
		}
	}
	// http issuer is accepted with APP_ENV=test (the fixture issuer), and the test anchor is read there.
	m, err := LoadMCP(mcpEnv(t, map[string]string{"MCP_CONNECT_ISSUER": "http://fake-workos:8080", "MCP_CONNECT_JWKS_FILE": jwks}), "test", mustIdentity(t))
	if err != nil || len(m.TestJWKS) == 0 {
		t.Fatalf("test configuration: %+v %v", m, err)
	}
}
