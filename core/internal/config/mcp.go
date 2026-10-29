package config

import (
	"errors"
	"io"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"
)

// MCP is the tenant-bound MCP connection configuration (G014; docs/foundation/MCP.md). It is zero (disabled: no MCP route
// is mounted) when every MCP_* variable is unset. It holds no secret: the issuer and the lifetimes are public values,
// and the test trust anchor is a public key set accepted only with APP_ENV=test (C10). Tenant hosts are not
// configuration: since G016 they resolve through Organization slugs and verified domains (the static
// MCP_TENANT_HOSTS_FILE registry is retired and refused).
type MCP struct {
	Issuer         string   // exact WorkOS AuthKit Connect issuer (MCP_CONNECT_ISSUER)
	AllowedOrigins []string // MCP_ALLOWED_ORIGINS: browser origins allowed to call the MCP endpoint (none by default)
	PendingTTL     time.Duration
	Idle           time.Duration
	Absolute       time.Duration
	Algorithms     []string // the identity JWS allow-list (IDENTITY_JWT_ALGORITHMS)
	TestJWKS       []byte   // MCP_CONNECT_JWKS_FILE contents (APP_ENV=test only)
}

// Configured reports whether MCP is configured.
func (m MCP) Configured() bool { return m.Issuer != "" }

// MCP defaults and bounds (Plan r8 §3.6; owner OD-1: pending setup 30 minutes, idle 30 days, absolute 180 days).
const (
	mcpDefaultPending = 30 * time.Minute
	mcpMinPending     = 2 * time.Minute
	mcpMaxPending     = 60 * time.Minute
	mcpDefaultIdle    = 30 * 24 * time.Hour
	mcpMinIdle        = 24 * time.Hour
	mcpMaxIdle        = 90 * 24 * time.Hour
	mcpDefaultMax     = 180 * 24 * time.Hour
	mcpMinMax         = 24 * time.Hour
	mcpMaxMax         = 730 * 24 * time.Hour
	mcpMaxJWKS        = 256 << 10
)

var mcpVariables = []string{"MCP_CONNECT_ISSUER", "MCP_TENANT_HOSTS_FILE", "MCP_ALLOWED_ORIGINS", "MCP_PENDING_TTL",
	"MCP_CONNECTION_IDLE", "MCP_CONNECTION_ABSOLUTE", "MCP_CONNECT_JWKS_FILE"}

// LoadMCP reads and validates the MCP configuration for appEnv on top of the validated identity configuration. Every
// problem is reported at once and names only variables, never values. All unset is the documented disabled state;
// partial configuration, MCP without identity, the retired host registry (G016) and the test trust anchor outside
// APP_ENV=test are refused.
func LoadMCP(getenv Getenv, appEnv string, id Identity) (MCP, error) {
	var problems []string
	add := func(p string) { problems = append(problems, p) }
	get := func(name string) string { return strings.TrimSpace(getenv(name)) }
	test := appEnv == "test"

	set := false
	for _, n := range mcpVariables {
		if get(n) != "" {
			set = true
		}
	}
	if !set {
		return MCP{}, nil
	}
	m := MCP{Issuer: get("MCP_CONNECT_ISSUER"), Algorithms: slices.Clone(id.Algorithms)}
	if m.Issuer == "" {
		add("MCP_CONNECT_ISSUER is required when MCP is configured")
	}
	if get("MCP_TENANT_HOSTS_FILE") != "" {
		add("MCP_TENANT_HOSTS_FILE is retired: tenant hosts resolve through Organization slugs and verified domains")
	}
	if !id.Configured() {
		add("MCP requires identity: IDENTITY_WORKOS_CLIENT_ID is not set")
	}
	if m.Issuer != "" {
		u, err := url.Parse(m.Issuer)
		if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" ||
			len(m.Issuer) > 2048 || strings.HasSuffix(m.Issuer, "/") || (u.Scheme != "https" && !(u.Scheme == "http" && test)) {
			add("MCP_CONNECT_ISSUER must be an exact https issuer")
		}
	}

	if raw := get("MCP_ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			o = strings.TrimSpace(o)
			if o == "" {
				continue
			}
			if !validOrigin(o, false) {
				add("MCP_ALLOWED_ORIGINS must list exact https origins")
				break
			}
			if !slices.Contains(m.AllowedOrigins, o) {
				m.AllowedOrigins = append(m.AllowedOrigins, o)
			}
		}
	}

	dur := func(name string, def, lo, hi time.Duration) time.Duration {
		v := get(name)
		if v == "" {
			return def
		}
		d, err := time.ParseDuration(v)
		if err != nil || d < lo || d > hi {
			add(name + " is outside its bounds")
			return def
		}
		return d
	}
	m.PendingTTL = dur("MCP_PENDING_TTL", mcpDefaultPending, mcpMinPending, mcpMaxPending)
	m.Idle = dur("MCP_CONNECTION_IDLE", mcpDefaultIdle, mcpMinIdle, mcpMaxIdle)
	m.Absolute = dur("MCP_CONNECTION_ABSOLUTE", mcpDefaultMax, mcpMinMax, mcpMaxMax)
	if m.Idle > m.Absolute {
		add("MCP_CONNECTION_IDLE must not exceed MCP_CONNECTION_ABSOLUTE")
	}

	if f := get("MCP_CONNECT_JWKS_FILE"); f != "" {
		if !test {
			add("MCP_CONNECT_JWKS_FILE is accepted only with APP_ENV=test")
		} else if b, err := readBounded(f, mcpMaxJWKS); err != nil || len(strings.TrimSpace(string(b))) == 0 {
			add("MCP_CONNECT_JWKS_FILE cannot be read")
		} else {
			m.TestJWKS = b
		}
	}

	if len(problems) > 0 {
		return MCP{}, errors.New("invalid configuration: " + strings.Join(problems, "; "))
	}
	return m, nil
}

// readBounded reads at most limit bytes of path; a larger file is an error.
func readBounded(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("too large")
	}
	return b, nil
}
