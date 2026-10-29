package config

import (
	"errors"
	"net"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"core/internal/hostname"
)

// Domains is the G016 hostname-routing and custom-domain lifecycle configuration (docs/foundation/DOMAINS.md). Every
// value is public deployment configuration; none is a secret. The platform domain and CNAME target come from Config.
type Domains struct {
	BaseDomain      string   // PLATFORM_BASE_DOMAIN (from Config)
	CNAMETarget     string   // CUSTOM_DOMAIN_CNAME_TARGET (from Config)
	HostPort        string   // PLATFORM_HOST_PORT: the only port accepted in a Host value ("" in production)
	ReservedLabels  []string // PLATFORM_RESERVED_LABELS
	AuthHost        string   // PLATFORM_AUTH_HOST: the shared auth (callback relay) host; "" disables the relay
	ChallengeLabel  string   // CUSTOM_DOMAIN_CHALLENGE_LABEL
	ChallengeTTL    time.Duration
	ProofFreshness  time.Duration
	DNSTimeout      time.Duration
	DNSFixture      []byte // CUSTOM_DOMAIN_DNS_FIXTURE_FILE contents (APP_ENV=test only)
	TLSAdapter      string // DOMAIN_TLS_ADAPTER: edge-probe (default) or fixture (APP_ENV=test only)
	TLSEdgeAddr     string // DOMAIN_TLS_EDGE_ADDR: host:port the edge-probe dials (default <CNAME target>:443)
	TLSTimeout      time.Duration
	TLSFixture      []byte // DOMAIN_TLS_FIXTURE_FILE contents (APP_ENV=test only)
	WithdrawalRetry int    // DOMAIN_WITHDRAWAL_RETRY_MAX
	// G071 (Plan r4 S1, S15): the optional default-tenant root alias and the multi-tenant root presentation. Both are
	// read trimmed and are never a configuration error: an empty slug is unset and any other value is judged once at
	// startup (domains.SelectDefaultTenant); RootPage holds the raw value until the server stores the effective mode
	// (domains.RootPresentation) before the handlers are built.
	DefaultTenantSlug string // DEFAULT_TENANT_SLUG
	RootPage          string // PLATFORM_ROOT_PAGE
}

// AuthOrigin returns the shared auth host's origin (https://host[:port]), or "" when the relay is not configured.
func (d Domains) AuthOrigin() string {
	if d.AuthHost == "" {
		return ""
	}
	if d.HostPort != "" {
		return "https://" + d.AuthHost + ":" + d.HostPort
	}
	return "https://" + d.AuthHost
}

// RootOrigin returns the exact platform root origin (https://<PLATFORM_BASE_DOMAIN>[:PLATFORM_HOST_PORT]; G062 Plan r5
// §4.1): the root signup origin and the signup_origin of both public entry names.
func (d Domains) RootOrigin() string { return hostname.PlatformRootOrigin(d.Hostnames()) }

// Hostnames returns the classification configuration of the hostname package.
func (d Domains) Hostnames() hostname.Config {
	return hostname.Config{BaseDomain: d.BaseDomain, CNAMETarget: d.CNAMETarget, AuthHost: d.AuthHost, Port: d.HostPort,
		ReservedLabels: d.ReservedLabels}
}

// Domain defaults and bounds (Plan r4 §5.4).
const (
	DomainDefaultChallengeLabel = "_platform-verify"
	domainDefaultChallengeTTL   = 168 * time.Hour
	domainDefaultFreshness      = 720 * time.Hour
	domainDefaultDNSTimeout     = 5 * time.Second
	domainDefaultTLSTimeout     = 5 * time.Second
	domainDefaultRetry          = 5
)

var domainVariables = []string{"PLATFORM_HOST_PORT", "PLATFORM_RESERVED_LABELS", "PLATFORM_AUTH_HOST",
	"CUSTOM_DOMAIN_CHALLENGE_LABEL", "CUSTOM_DOMAIN_CHALLENGE_TTL", "CUSTOM_DOMAIN_PROOF_FRESHNESS",
	"CUSTOM_DOMAIN_DNS_TIMEOUT", "CUSTOM_DOMAIN_DNS_FIXTURE_FILE", "DOMAIN_TLS_ADAPTER", "DOMAIN_TLS_EDGE_ADDR",
	"DOMAIN_TLS_TIMEOUT", "DOMAIN_TLS_FIXTURE_FILE", "DOMAIN_WITHDRAWAL_RETRY_MAX", "DEFAULT_TENANT_SLUG",
	"PLATFORM_ROOT_PAGE"}

// DomainVariables returns the names LoadDomains reads, sorted.
func DomainVariables() []string {
	out := slices.Clone(domainVariables)
	slices.Sort(out)
	return out
}

// LoadDomains reads and validates the domain configuration for c.AppEnv. Every problem is reported at once and names
// only variables. Production refuses a host port and both test fixtures; the shared
// auth host's origin must be one of IDENTITY_PUBLIC_ORIGINS when identity is configured, and it must be the only
// entry there (that list is the exact allow-list of non-tenant origins, Plan r4 §1.7).
func LoadDomains(getenv Getenv, c Config, id Identity) (Domains, error) {
	var problems []string
	add := func(p string) { problems = append(problems, p) }
	get := func(name string) string { return strings.TrimSpace(getenv(name)) }
	production, test := c.AppEnv == "production", c.AppEnv == "test"

	d := Domains{
		BaseDomain: c.PlatformBaseDomain, CNAMETarget: c.CustomDomainCNAMETarget,
		HostPort: get("PLATFORM_HOST_PORT"), AuthHost: get("PLATFORM_AUTH_HOST"),
		ChallengeLabel: get("CUSTOM_DOMAIN_CHALLENGE_LABEL"), TLSAdapter: get("DOMAIN_TLS_ADAPTER"),
		TLSEdgeAddr: get("DOMAIN_TLS_EDGE_ADDR"), DefaultTenantSlug: get("DEFAULT_TENANT_SLUG"),
		RootPage: get("PLATFORM_ROOT_PAGE"),
	}
	if d.HostPort != "" {
		if n, err := strconv.Atoi(d.HostPort); err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != d.HostPort {
			add("PLATFORM_HOST_PORT must be a port number")
		} else if production {
			add("PLATFORM_HOST_PORT is refused in production")
		}
	}
	for _, l := range strings.Split(get("PLATFORM_RESERVED_LABELS"), ",") {
		if l = strings.TrimSpace(l); l == "" {
			continue
		}
		if !hostLabelRe.MatchString(l) || strings.ToLower(l) != l {
			add("PLATFORM_RESERVED_LABELS must list lowercase hostname labels")
			break
		}
		if !slices.Contains(d.ReservedLabels, l) {
			d.ReservedLabels = append(d.ReservedLabels, l)
		}
	}
	if d.AuthHost != "" {
		if !validHostname(d.AuthHost) && !(test && d.AuthHost == "localhost") {
			add("PLATFORM_AUTH_HOST must be a lowercase hostname")
		} else if d.AuthHost == d.BaseDomain || d.AuthHost == d.CNAMETarget {
			add("PLATFORM_AUTH_HOST must differ from PLATFORM_BASE_DOMAIN and CUSTOM_DOMAIN_CNAME_TARGET")
		} else if id.Configured() && !slices.Contains(id.PublicOrigins, d.AuthOrigin()) {
			add("PLATFORM_AUTH_HOST must be listed in IDENTITY_PUBLIC_ORIGINS")
		}
	}
	if id.Configured() {
		hc := d.Hostnames()
		for _, o := range id.PublicOrigins {
			u, err := url.Parse(o)
			if err != nil {
				continue // LoadIdentity already refused it
			}
			h, err := hostname.Normalize(u.Host, d.HostPort)
			if err != nil {
				add("IDENTITY_PUBLIC_ORIGINS must use PLATFORM_HOST_PORT")
				break
			}
			if k := hostname.Classify(h, hc).Kind; k != hostname.KindReserved && k != hostname.KindPublicEntry {
				add("IDENTITY_PUBLIC_ORIGINS may list only non-tenant origins")
				break
			}
		}
	}
	if d.ChallengeLabel == "" {
		d.ChallengeLabel = DomainDefaultChallengeLabel
	} else if !strings.HasPrefix(d.ChallengeLabel, "_") || len(d.ChallengeLabel) < 2 || len(d.ChallengeLabel) > 63 ||
		strings.Trim(d.ChallengeLabel[1:], "abcdefghijklmnopqrstuvwxyz0123456789-") != "" {
		add("CUSTOM_DOMAIN_CHALLENGE_LABEL must be one underscore-prefixed lowercase label")
	}

	dur := func(name string, def, lo, hi time.Duration) time.Duration {
		v := get(name)
		if v == "" {
			return def
		}
		dv, err := time.ParseDuration(v)
		if err != nil || dv < lo || dv > hi {
			add(name + " must be a duration between " + lo.String() + " and " + hi.String())
			return def
		}
		return dv
	}
	d.ChallengeTTL = dur("CUSTOM_DOMAIN_CHALLENGE_TTL", domainDefaultChallengeTTL, time.Hour, 720*time.Hour)
	d.ProofFreshness = dur("CUSTOM_DOMAIN_PROOF_FRESHNESS", domainDefaultFreshness, 24*time.Hour, 2160*time.Hour)
	d.DNSTimeout = dur("CUSTOM_DOMAIN_DNS_TIMEOUT", domainDefaultDNSTimeout, time.Second, 15*time.Second)
	d.TLSTimeout = dur("DOMAIN_TLS_TIMEOUT", domainDefaultTLSTimeout, time.Second, 15*time.Second)

	d.WithdrawalRetry = domainDefaultRetry
	if v := get("DOMAIN_WITHDRAWAL_RETRY_MAX"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 20 {
			add("DOMAIN_WITHDRAWAL_RETRY_MAX must be between 1 and 20")
		} else {
			d.WithdrawalRetry = n
		}
	}

	switch d.TLSAdapter {
	case "":
		d.TLSAdapter = "edge-probe"
	case "edge-probe":
	case "fixture":
		if !test {
			add("DOMAIN_TLS_ADAPTER fixture is accepted only with APP_ENV=test")
		}
	default:
		add("DOMAIN_TLS_ADAPTER must be edge-probe or fixture")
	}
	if d.TLSEdgeAddr == "" {
		if d.CNAMETarget != "" {
			d.TLSEdgeAddr = net.JoinHostPort(d.CNAMETarget, "443")
		}
	} else if h, p, err := net.SplitHostPort(d.TLSEdgeAddr); err != nil || h == "" || p == "" {
		add("DOMAIN_TLS_EDGE_ADDR must be host:port")
	}

	fixture := func(name string) []byte {
		path := get(name)
		if path == "" {
			return nil
		}
		if !test {
			add(name + " is accepted only with APP_ENV=test")
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil || len(b) == 0 || len(b) > 1<<20 {
			add(name + " cannot be read")
			return nil
		}
		return b
	}
	d.DNSFixture = fixture("CUSTOM_DOMAIN_DNS_FIXTURE_FILE")
	d.TLSFixture = fixture("DOMAIN_TLS_FIXTURE_FILE")
	if d.TLSAdapter == "fixture" && d.TLSFixture == nil && test {
		add("DOMAIN_TLS_FIXTURE_FILE is required with DOMAIN_TLS_ADAPTER fixture")
	}

	if len(problems) > 0 {
		return Domains{}, errors.New("invalid configuration: " + strings.Join(problems, "; "))
	}
	return d, nil
}
