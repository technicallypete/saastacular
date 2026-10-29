// Package hostname is the pure hostname normalization and classification contract of G016 (Plan r4 §1.1): it turns
// the single Host header value the trusted edge delivered into a portless lowercase key and classifies it against the
// deployment configuration. It performs no I/O and holds no tenant fact; resolution of a classified name to an
// Organization is the domains package's job. Names select; they never authorize.
package hostname

import (
	"errors"
	"slices"
	"strings"
)

// MaxHeaderBytes bounds the raw Host value (253-byte name, ':' and a five-digit port, plus a trailing dot).
const MaxHeaderBytes = 261

// Errors returned by Normalize. Each is a malformed host for every caller (no tenant is selected).
var (
	ErrMalformed = errors.New("hostname: malformed host")
	ErrNonASCII  = errors.New("hostname: non-ASCII or IDNA host")
	ErrPort      = errors.New("hostname: port not accepted")
)

// Host is a normalized host: Name is the portless lowercase key; Port is the accepted port ("" when none was sent).
type Host struct {
	Name string
	Port string
}

// HostPort returns name[:port] exactly as validated (the form stored by MCP connections).
func (h Host) HostPort() string {
	if h.Port == "" {
		return h.Name
	}
	return h.Name + ":" + h.Port
}

// Origin returns https://name[:port].
func (h Host) Origin() string { return "https://" + h.HostPort() }

// Kind is the closed classification of a normalized host.
type Kind string

// Classification kinds. KindNone is a host under the platform domain that can never route (a nested name or a label
// outside the slug grammar); it is answered as an unknown host without a lookup.
const (
	KindPublicEntry Kind = "public_entry"
	KindReserved    Kind = "reserved"
	KindPlatform    Kind = "platform"
	KindCustom      Kind = "custom"
	KindNone        Kind = "none"
)

// Config is the deployment configuration classification depends on. Every name is a normalized portless hostname.
type Config struct {
	BaseDomain     string   // PLATFORM_BASE_DOMAIN (B)
	CNAMETarget    string   // CUSTOM_DOMAIN_CNAME_TARGET (T)
	AuthHost       string   // PLATFORM_AUTH_HOST: the shared auth host of the B2 relay ("" when unset)
	Port           string   // PLATFORM_HOST_PORT ("" in production: no port accepted)
	ReservedLabels []string // PLATFORM_RESERVED_LABELS: additional reserved first labels
}

// Classification is the result of Classify. Slug is set only for KindPlatform and holds the platform label (G057:
// the generated opaque address label of an Organization, resolved through its platform_label); the friendly
// Organization slug is metadata and never routes. The field keeps its historical name.
type Classification struct {
	Kind Kind
	Slug string
}

// groupReserved mirrors the Group reserved slug set of 00003_tenancy.sql (tenancy.ReservedSlugs; parity-tested).
var groupReserved = []string{
	"account", "admin", "api", "auth", "billing", "blueprint", "fonts", "health", "healthz", "internal", "login",
	"logout", "mcp", "members", "message", "messages", "new", "onboarding", "readyz", "rpc", "settings", "sign-in",
	"sse", "system",
}

// hostReserved is the hostname-specific addition to the floor (Plan §1.1).
var hostReserved = []string{"www", "auth", "custom", "app", "mail", "smtp", "ns", "ns1", "ns2", "cdn", "static", "status"}

// ReservedLabels returns the reserved first-label floor, sorted and de-duplicated: the Group reserved set plus the
// hostname additions. It is pinned identically in the organizations_slug_reserved CHECK of 00011 (parity-tested).
func ReservedLabels() []string {
	out := slices.Concat(groupReserved, hostReserved)
	slices.Sort(out)
	return slices.Compact(out)
}

// Normalize validates and normalizes one Host header value. port is the configured PLATFORM_HOST_PORT: a port in the
// value is accepted only when it equals port exactly (so none is accepted when port is empty). The value is lowercased
// (ASCII only), one trailing dot is stripped, and the name must be a dotted sequence of 1-63 byte labels of
// [a-z0-9-] without leading or trailing hyphens, at most 253 bytes. IP literals, userinfo, paths, schemes, whitespace,
// non-ASCII bytes and xn-- (IDNA) labels are refused: no IDNA mapping is attempted in v0.
func Normalize(raw, port string) (Host, error) {
	if raw == "" || len(raw) > MaxHeaderBytes {
		return Host{}, ErrMalformed
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] >= 0x80 {
			return Host{}, ErrNonASCII
		}
	}
	s := strings.ToLower(raw)
	name, p := s, ""
	if i := strings.LastIndexByte(s, ':'); i >= 0 {
		name, p = s[:i], s[i+1:]
		if p == "" || len(p) > 5 || strings.Trim(p, "0123456789") != "" {
			return Host{}, ErrMalformed
		}
		if p != port {
			return Host{}, ErrPort
		}
	}
	name = strings.TrimSuffix(name, ".")
	if err := validName(name); err != nil {
		return Host{}, err
	}
	return Host{Name: name, Port: p}, nil
}

// ValidName reports whether name is already a normalized portless hostname (the form stored and configured).
func ValidName(name string) bool { return validName(name) == nil }

func validName(name string) error {
	if name == "" || len(name) > 253 {
		return ErrMalformed
	}
	labels := strings.Split(name, ".")
	for _, l := range labels {
		if len(l) < 1 || len(l) > 63 || l[0] == '-' || l[len(l)-1] == '-' {
			return ErrMalformed
		}
		for i := 0; i < len(l); i++ {
			c := l[i]
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return ErrMalformed
			}
		}
		if strings.HasPrefix(l, "xn--") {
			return ErrNonASCII
		}
	}
	// An all-numeric last label is an IPv4 literal (or a lookalike of one); no top-level domain is numeric.
	if strings.Trim(labels[len(labels)-1], "0123456789") == "" {
		return ErrMalformed
	}
	return nil
}

// Classify assigns a normalized host its kind under cfg: the platform domain and www.<domain> are the public entry;
// the CNAME target, the shared auth host and <reserved>.<domain> are reserved; <label>.<domain> with one label in the
// slug grammar is a platform label (KindPlatform; a generated 26-character base32 label fits that grammar); any other
// name under the domain is KindNone; everything else is a custom name. No default tenant is classified here; the
// resolver may alias the exact root (G071).
func Classify(h Host, cfg Config) Classification {
	n := h.Name
	switch {
	case n == cfg.BaseDomain || n == "www."+cfg.BaseDomain:
		return Classification{Kind: KindPublicEntry}
	case n == cfg.CNAMETarget || (cfg.AuthHost != "" && n == cfg.AuthHost):
		return Classification{Kind: KindReserved}
	}
	label, under := strings.CutSuffix(n, "."+cfg.BaseDomain)
	if !under {
		return Classification{Kind: KindCustom}
	}
	if strings.Contains(label, ".") || !slugGrammar(label) {
		return Classification{Kind: KindNone}
	}
	if IsReservedLabel(label, cfg) {
		return Classification{Kind: KindReserved}
	}
	return Classification{Kind: KindPlatform, Slug: label}
}

// PlatformRootOrigin returns the exact platform root origin under cfg, https://<PLATFORM_BASE_DOMAIN>[:PLATFORM_HOST_PORT]
// (G062 Plan r5 §4.1): the only origin root signup starts and completes on, and the canonical signup origin both public
// entry names (the platform domain and www.<domain>) link to. It is "" when no platform domain is configured.
func PlatformRootOrigin(cfg Config) string {
	if cfg.BaseDomain == "" {
		return ""
	}
	return Host{Name: cfg.BaseDomain, Port: cfg.Port}.Origin()
}

// IsPlatformRoot reports whether a normalized host is the platform domain itself (G062): www.<domain> is a public entry
// but never the root signup host.
func IsPlatformRoot(h Host, cfg Config) bool { return cfg.BaseDomain != "" && h.Name == cfg.BaseDomain }

// IsReservedLabel reports whether label is in the floor or reserved by configuration: the first label of the CNAME
// target and of the auth host when they sit directly under the platform domain, and PLATFORM_RESERVED_LABELS.
func IsReservedLabel(label string, cfg Config) bool {
	if slices.Contains(groupReserved, label) || slices.Contains(hostReserved, label) ||
		slices.Contains(cfg.ReservedLabels, label) {
		return true
	}
	for _, n := range []string{cfg.CNAMETarget, cfg.AuthHost} {
		if l, ok := strings.CutSuffix(n, "."+cfg.BaseDomain); ok && l == label {
			return true
		}
	}
	return false
}

// CustomClaimAllowed reports whether a normalized name may be claimed as a custom domain: at least two labels, and
// neither the platform domain, the CNAME target, the auth host nor any name under the platform domain.
func CustomClaimAllowed(name string, cfg Config) bool {
	if !strings.Contains(name, ".") {
		return false
	}
	return name != cfg.BaseDomain && name != cfg.CNAMETarget && (cfg.AuthHost == "" || name != cfg.AuthHost) &&
		!strings.HasSuffix(name, "."+cfg.BaseDomain)
}

// slugGrammar is the Organization/Group slug grammar: 1-63 bytes of [a-z0-9]+(-[a-z0-9]+)*.
func slugGrammar(s string) bool {
	if len(s) < 1 || len(s) > 63 || s[0] == '-' || s[len(s)-1] == '-' || strings.Contains(s, "--") {
		return false
	}
	return strings.Trim(s, "abcdefghijklmnopqrstuvwxyz0123456789-") == ""
}

// SlugGrammar is the exported form of the slug grammar check.
func SlugGrammar(s string) bool { return slugGrammar(s) }
