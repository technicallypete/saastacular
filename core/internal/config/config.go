// Package config loads and validates core configuration from the
// environment. Product, domain and database identity are always read at
// runtime; nothing here embeds a product literal.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Getenv reads one environment variable.
type Getenv func(string) string

// SecretProvider is the boundary for secret material. The local
// implementation reads the environment; production deployments substitute a
// secrets-manager-backed implementation without touching callers.
type SecretProvider interface {
	Secret(name string) (string, error)
}

// EnvSecrets reads secrets from the environment.
type EnvSecrets struct{ Getenv Getenv }

// Secret returns the named secret or an error wrapping ErrSecretNotSet when it is unset (G012).
func (e EnvSecrets) Secret(name string) (string, error) {
	v := e.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("secret %s: %w", name, ErrSecretNotSet)
	}
	return v, nil
}

// DB holds PostgreSQL connection settings as separate components, so
// passwords never need URL escaping.
type DB struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// Redis holds Redis connection settings.
type Redis struct {
	Host     string
	Port     int
	Password string
}

// ObjectStore holds S3-compatible object store settings. The G009 storage settings (SigningEndpoint and the bounds)
// are optional and defaulted; nothing uses them until a reviewed adapter wires the guarded storage operations
// (docs/foundation/OBJECT-STORAGE.md).
type ObjectStore struct {
	Endpoint       string
	Region         string
	Bucket         string
	ForcePathStyle bool
	AccessKey      string
	SecretKey      string

	SigningEndpoint     string        // the trusted endpoint signed URLs name; defaults to Endpoint
	MaxObjectBytes      int64         // 1..1073741824, default 26214400 (25 MiB)
	SignedURLDefaultTTL time.Duration // 1..300 s and at most SignedURLMaxTTL, default 60 s
	SignedURLMaxTTL     time.Duration // 1..300 s, default 300 s
	TransferTimeout     time.Duration // 1..900 s, default 120 s
}

// The G009 object-storage defaults and bounds (Plan r1 §4.3, §4.7).
const (
	ObjectDefaultMaxBytes        int64 = 26214400
	ObjectHardMaxBytes           int64 = 1073741824
	ObjectDefaultSignedURLTTL          = 60 * time.Second
	ObjectHardMaxSignedURLTTL          = 300 * time.Second
	ObjectDefaultTransferTimeout       = 120 * time.Second
	ObjectHardMaxTransferTimeout       = 900 * time.Second
)

// Intelligence holds the optional internal intelligence service settings.
// It is zero when intelligence is disabled.
type Intelligence struct {
	URL   string
	Token string
}

// Enabled reports whether the intelligence service is configured.
func (i Intelligence) Enabled() bool { return i.URL != "" }

// Config is the validated core configuration.
type Config struct {
	ProductName             string
	PlatformBaseDomain      string
	CustomDomainCNAMETarget string
	DatabaseName            string
	AppEnv                  string
	ListenAddr              string
	DB                      DB
	Redis                   Redis
	ObjectStore             ObjectStore
	Intelligence            Intelligence
}

// SecretValues lists loaded secret values so log output can redact them.
func (c Config) SecretValues() []string {
	var out []string
	for _, v := range []string{c.DB.Password, c.Redis.Password, c.ObjectStore.SecretKey, c.Intelligence.Token} {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

var (
	hostLabelRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	dbNameRe    = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)
)

var appEnvs = map[string]bool{"development": true, "test": true, "production": true}

// Load reads and validates configuration. Every problem is reported at once
// and startup fails closed.
func Load(getenv Getenv, secrets SecretProvider) (Config, error) {
	var problems []string
	req := func(name string) string {
		v := strings.TrimSpace(getenv(name))
		if v == "" {
			problems = append(problems, name+" is required")
		}
		return v
	}
	secret := func(name string) string { // G012: one provider call, closed problem text, provider text discarded
		v, problem := resolveSecret(secrets, name)
		if problem != "" {
			problems = append(problems, problem)
		}
		return v
	}
	port := func(name string) int {
		raw := req(name)
		if raw == "" {
			return 0
		}
		p, err := strconv.Atoi(raw)
		if err != nil || p < 1 || p > 65535 {
			problems = append(problems, name+" must be a port number")
		}
		return p
	}

	c := Config{
		ProductName:             req("PRODUCT_NAME"),
		PlatformBaseDomain:      req("PLATFORM_BASE_DOMAIN"),
		CustomDomainCNAMETarget: req("CUSTOM_DOMAIN_CNAME_TARGET"),
		DatabaseName:            req("DATABASE_NAME"),
		AppEnv:                  req("APP_ENV"),
		ListenAddr:              getenv("CORE_LISTEN_ADDR"),
	}
	if c.ListenAddr == "" {
		c.ListenAddr = ":8180"
	}
	for _, h := range []struct{ name, value string }{
		{"PLATFORM_BASE_DOMAIN", c.PlatformBaseDomain},
		{"CUSTOM_DOMAIN_CNAME_TARGET", c.CustomDomainCNAMETarget},
	} {
		// G016: the identity overlay serves <slug>.localhost, so the single-label platform domain localhost is accepted
		// with APP_ENV=test only.
		if h.value != "" && !validHostname(h.value) && !(c.AppEnv == "test" && h.name == "PLATFORM_BASE_DOMAIN" && h.value == "localhost") {
			problems = append(problems, h.name+" must be a lowercase hostname")
		}
	}
	if c.DatabaseName != "" && !dbNameRe.MatchString(c.DatabaseName) {
		problems = append(problems, "DATABASE_NAME is not a valid database name")
	}
	if c.AppEnv != "" && !appEnvs[c.AppEnv] {
		problems = append(problems, "APP_ENV must be development, test or production")
	}

	c.DB = DB{Host: req("DB_HOST"), Port: port("DB_PORT"), Name: c.DatabaseName, User: req("DB_USER"), Password: secret("DB_PASSWORD")}
	c.Redis = Redis{Host: req("REDIS_HOST"), Port: port("REDIS_PORT"), Password: secret("REDIS_PASSWORD")}

	pathStyle := strings.ToLower(req("OBJECT_STORE_FORCE_PATH_STYLE"))
	if pathStyle != "" && pathStyle != "true" && pathStyle != "false" {
		problems = append(problems, "OBJECT_STORE_FORCE_PATH_STYLE must be true or false")
	}
	c.ObjectStore = ObjectStore{
		Endpoint:       req("OBJECT_STORE_ENDPOINT"),
		Region:         req("OBJECT_STORE_REGION"),
		Bucket:         req("OBJECT_STORE_BUCKET"),
		ForcePathStyle: pathStyle == "true",
		AccessKey:      req("OBJECT_STORE_ACCESS_KEY"),
		SecretKey:      secret("OBJECT_STORE_SECRET_KEY"),
	}
	problems = append(problems, loadObjectSettings(getenv, &c.ObjectStore, c.AppEnv)...)

	if url := strings.TrimSpace(getenv("INTELLIGENCE_URL")); url != "" {
		c.Intelligence = Intelligence{URL: url, Token: secret("INTELLIGENCE_SERVICE_TOKEN")}
	} else if getenv("INTELLIGENCE_SERVICE_TOKEN") != "" {
		problems = append(problems, "INTELLIGENCE_SERVICE_TOKEN is set but INTELLIGENCE_URL is not")
	}

	if len(problems) > 0 {
		return Config{}, errors.New("invalid configuration: " + strings.Join(problems, "; "))
	}
	return c, nil
}

// loadObjectSettings reads the optional G009 storage settings into o and checks the endpoint transport rules for appEnv:
// each endpoint is an absolute http(s) URL with a host and an optional port only; with APP_ENV=production both the
// data and the signing endpoint are https; an http data endpoint with an https signing endpoint is refused everywhere.
// core/internal/objectstore re-enforces the same rules at construction. Problems name only variables.
func loadObjectSettings(getenv Getenv, o *ObjectStore, appEnv string) []string {
	var problems []string
	o.SigningEndpoint = strings.TrimSpace(getenv("OBJECT_STORE_SIGNING_ENDPOINT"))
	bounded := func(name string, def, max int64) int64 {
		raw := strings.TrimSpace(getenv(name))
		if raw == "" {
			return def
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 1 || n > max || strconv.FormatInt(n, 10) != raw {
			problems = append(problems, fmt.Sprintf("%s must be an integer from 1 to %d", name, max))
			return def
		}
		return n
	}
	o.MaxObjectBytes = bounded("OBJECT_STORE_MAX_OBJECT_BYTES", ObjectDefaultMaxBytes, ObjectHardMaxBytes)
	maxTTL := bounded("OBJECT_STORE_SIGNED_URL_MAX_TTL", int64(ObjectHardMaxSignedURLTTL/time.Second), int64(ObjectHardMaxSignedURLTTL/time.Second))
	defTTL := bounded("OBJECT_STORE_SIGNED_URL_DEFAULT_TTL", int64(ObjectDefaultSignedURLTTL/time.Second), int64(ObjectHardMaxSignedURLTTL/time.Second))
	if defTTL > maxTTL {
		problems = append(problems, "OBJECT_STORE_SIGNED_URL_DEFAULT_TTL must not exceed OBJECT_STORE_SIGNED_URL_MAX_TTL")
	}
	o.SignedURLMaxTTL, o.SignedURLDefaultTTL = time.Duration(maxTTL)*time.Second, time.Duration(defTTL)*time.Second
	o.TransferTimeout = time.Duration(bounded("OBJECT_STORE_TRANSFER_TIMEOUT", int64(ObjectDefaultTransferTimeout/time.Second),
		int64(ObjectHardMaxTransferTimeout/time.Second))) * time.Second
	if o.Endpoint != "" {
		problems = append(problems, objectEndpointProblems(o.Endpoint, o.SigningEndpoint, appEnv)...)
	} else if o.SigningEndpoint != "" {
		if _, ok := endpointScheme(o.SigningEndpoint); !ok {
			problems = append(problems, "OBJECT_STORE_SIGNING_ENDPOINT must be an absolute http or https URL with a host, an optional port and nothing else")
		}
	}
	return problems
}

// objectEndpointProblems is the endpoint transport rule set (G009 Plan r1 §4.7); an empty signing endpoint is the
// data endpoint.
func objectEndpointProblems(data, signing, appEnv string) []string {
	var problems []string
	if signing == "" {
		signing = data
	}
	ds, dok := endpointScheme(data)
	ss, sok := endpointScheme(signing)
	if !dok {
		problems = append(problems, "OBJECT_STORE_ENDPOINT must be an absolute http or https URL with a host, an optional port and nothing else")
	}
	if !sok {
		problems = append(problems, "OBJECT_STORE_SIGNING_ENDPOINT must be an absolute http or https URL with a host, an optional port and nothing else")
	}
	if dok && sok {
		if appEnv == "production" {
			if ds != "https" {
				problems = append(problems, "OBJECT_STORE_ENDPOINT must be https with APP_ENV=production")
			}
			if ss != "https" {
				problems = append(problems, "OBJECT_STORE_SIGNING_ENDPOINT must be https with APP_ENV=production")
			}
		}
		if ds == "http" && ss == "https" {
			problems = append(problems, "OBJECT_STORE_SIGNING_ENDPOINT must not be https while OBJECT_STORE_ENDPOINT is http")
		}
	}
	return problems
}

// endpointScheme returns the scheme of a well-formed object-store endpoint: an absolute http(s) URL with a host and an
// optional numeric port, without userinfo, path, query or fragment.
func endpointScheme(s string) (string, bool) {
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Hostname() == "" || u.User != nil ||
		u.Path != "" || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" ||
		strings.ContainsAny(s, " \t\r\n") {
		return "", false
	}
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != p {
			return "", false
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return "", false
	}
	return u.Scheme, true
}

// Identity holds the WorkOS AuthKit identity configuration (G004). It is zero when identity is unconfigured: then
// provider operations answer identity_unconfigured, session resolution still works, and production refuses to
// start. Web and intelligence never receive these values.
type Identity struct {
	ClientID      string
	Issuer        string
	APIKey        string // secret: one of IDENTITY_WORKOS_API_KEY or IDENTITY_WORKOS_API_KEY_FILE
	APIBaseURL    string
	PublicOrigins []string
	SessionIdle   time.Duration
	SessionMax    time.Duration
	Algorithms    []string
	MaxAuthAge    time.Duration // zero: freshness is not required
	CookieProfile string        // "https" or "loopback-http"
}

// Configured reports whether a WorkOS client is configured.
func (i Identity) Configured() bool { return i.ClientID != "" }

// SecretValues lists the identity secrets so log output can redact them.
func (i Identity) SecretValues() []string {
	if i.APIKey == "" {
		return nil
	}
	return []string{i.APIKey}
}

// Identity defaults and bounds (Plan r5 §4.3/§4.6; the default API base is WorkOS's production API).
const (
	IdentityDefaultAPIBase = "https://api.workos.com"
	identityDefaultIdle    = 30 * time.Minute
	identityDefaultMax     = 8 * time.Hour
	identityMinLifetime    = 5 * time.Minute
	identityMaxLifetime    = 24 * time.Hour
)

var (
	identityAlgorithms = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "ES256", "ES384"}
	workosClientIDRe   = regexp.MustCompile(`^client_[A-Za-z0-9]{1,248}$`)
)

// LoadIdentity reads and validates the identity configuration for appEnv. Every problem is reported at once and
// names only variables, never values. Partial configuration is invalid in every environment; production refuses
// unconfigured identity, the loopback-http cookie profile, a non-default or non-https API base and lifetimes
// outside [5 min, 24 h].
func LoadIdentity(getenv Getenv, appEnv string) (Identity, error) {
	var problems []string
	add := func(p string) { problems = append(problems, p) }
	get := func(name string) string { return strings.TrimSpace(getenv(name)) }
	production := appEnv == "production"

	if p := get("IDENTITY_PROVIDER"); p != "" && p != "workos" {
		add("IDENTITY_PROVIDER must be workos")
	}
	id := Identity{ClientID: get("IDENTITY_WORKOS_CLIENT_ID"), Issuer: get("IDENTITY_WORKOS_ISSUER")}
	keyValue, keyFile := getenv("IDENTITY_WORKOS_API_KEY"), get("IDENTITY_WORKOS_API_KEY_FILE")
	switch {
	case keyValue != "" && keyFile != "":
		add("IDENTITY_WORKOS_API_KEY and IDENTITY_WORKOS_API_KEY_FILE are mutually exclusive")
	case keyFile != "":
		b, err := os.ReadFile(keyFile)
		if err != nil {
			add("IDENTITY_WORKOS_API_KEY_FILE cannot be read")
		} else if id.APIKey = strings.TrimSpace(string(b)); id.APIKey == "" {
			add("IDENTITY_WORKOS_API_KEY_FILE is empty")
		}
	default:
		id.APIKey = strings.TrimSpace(keyValue)
	}
	if strings.ContainsAny(id.APIKey, " \t\r\n") {
		add("IDENTITY_WORKOS_API_KEY must be a single token")
	}

	set := 0
	for _, v := range []string{id.ClientID, id.Issuer, id.APIKey} {
		if v != "" {
			set++
		}
	}
	others := []string{"IDENTITY_WORKOS_API_BASE_URL", "IDENTITY_PUBLIC_ORIGINS", "IDENTITY_SESSION_IDLE",
		"IDENTITY_SESSION_ABSOLUTE", "IDENTITY_JWT_ALGORITHMS", "IDENTITY_MAX_AUTH_AGE"}
	switch {
	case set == 0:
		for _, n := range others {
			if get(n) != "" {
				add(n + " is set but IDENTITY_WORKOS_CLIENT_ID is not")
			}
		}
		if production {
			add("IDENTITY_WORKOS_CLIENT_ID is required in production")
		}
		if cp := get("IDENTITY_COOKIE_PROFILE"); cp != "" && cp != "https" && cp != "loopback-http" {
			add("IDENTITY_COOKIE_PROFILE must be https or loopback-http")
		} else if cp == "loopback-http" && production {
			add("IDENTITY_COOKIE_PROFILE loopback-http is refused in production")
		}
		if len(problems) > 0 {
			return Identity{}, errors.New("invalid configuration: " + strings.Join(problems, "; "))
		}
		return Identity{}, nil
	case set < 3:
		for _, n := range []struct{ name, value string }{
			{"IDENTITY_WORKOS_CLIENT_ID", id.ClientID}, {"IDENTITY_WORKOS_ISSUER", id.Issuer},
			{"IDENTITY_WORKOS_API_KEY", id.APIKey},
		} {
			if n.value == "" {
				add(n.name + " is required when identity is configured")
			}
		}
	}
	if id.ClientID != "" && !workosClientIDRe.MatchString(id.ClientID) {
		add("IDENTITY_WORKOS_CLIENT_ID is not a WorkOS client ID")
	}
	if id.Issuer != "" {
		u, err := url.Parse(id.Issuer)
		if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || len(id.Issuer) > 2048 ||
			(u.Scheme != "https" && !(u.Scheme == "http" && appEnv == "test")) {
			add("IDENTITY_WORKOS_ISSUER must be an exact https issuer")
		}
	}

	id.APIBaseURL = get("IDENTITY_WORKOS_API_BASE_URL")
	if id.APIBaseURL == "" {
		id.APIBaseURL = IdentityDefaultAPIBase
	}
	if u, err := url.Parse(id.APIBaseURL); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		add("IDENTITY_WORKOS_API_BASE_URL must be a base URL without path, query or credentials")
	} else if id.APIBaseURL != IdentityDefaultAPIBase && appEnv != "test" {
		add("IDENTITY_WORKOS_API_BASE_URL may differ from the default only with APP_ENV=test")
	}
	id.APIBaseURL = strings.TrimSuffix(id.APIBaseURL, "/")

	id.CookieProfile = get("IDENTITY_COOKIE_PROFILE")
	if id.CookieProfile == "" {
		id.CookieProfile = "https"
	}
	if id.CookieProfile != "https" && id.CookieProfile != "loopback-http" {
		add("IDENTITY_COOKIE_PROFILE must be https or loopback-http")
	} else if id.CookieProfile == "loopback-http" && production {
		add("IDENTITY_COOKIE_PROFILE loopback-http is refused in production")
	}

	raw := get("IDENTITY_PUBLIC_ORIGINS")
	if raw == "" {
		add("IDENTITY_PUBLIC_ORIGINS is required when identity is configured")
	}
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if !validOrigin(o, id.CookieProfile == "loopback-http") {
			add("IDENTITY_PUBLIC_ORIGINS must list exact https origins")
			break
		}
		if !slices.Contains(id.PublicOrigins, o) {
			id.PublicOrigins = append(id.PublicOrigins, o)
		}
	}

	dur := func(name string, def time.Duration) time.Duration {
		v := get(name)
		if v == "" {
			return def
		}
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			add(name + " must be a positive duration")
			return def
		}
		return d
	}
	id.SessionIdle = dur("IDENTITY_SESSION_IDLE", identityDefaultIdle)
	id.SessionMax = dur("IDENTITY_SESSION_ABSOLUTE", identityDefaultMax)
	id.MaxAuthAge = dur("IDENTITY_MAX_AUTH_AGE", 0)
	if id.SessionIdle > id.SessionMax {
		add("IDENTITY_SESSION_IDLE must not exceed IDENTITY_SESSION_ABSOLUTE")
	}
	if appEnv != "test" {
		for _, d := range []struct {
			name string
			v    time.Duration
		}{{"IDENTITY_SESSION_IDLE", id.SessionIdle}, {"IDENTITY_SESSION_ABSOLUTE", id.SessionMax}} {
			if d.v < identityMinLifetime || d.v > identityMaxLifetime {
				add(d.name + " must be between 5m and 24h outside APP_ENV=test")
			}
		}
	}

	algs := get("IDENTITY_JWT_ALGORITHMS")
	if algs == "" {
		algs = "RS256"
	}
	for _, a := range strings.Split(algs, ",") {
		a = strings.TrimSpace(a)
		if !slices.Contains(identityAlgorithms, a) {
			add("IDENTITY_JWT_ALGORITHMS may list only asymmetric JWS algorithms")
			break
		}
		if !slices.Contains(id.Algorithms, a) {
			id.Algorithms = append(id.Algorithms, a)
		}
	}

	if len(problems) > 0 {
		return Identity{}, errors.New("invalid configuration: " + strings.Join(problems, "; "))
	}
	return id, nil
}

// validOrigin accepts an exact origin: https scheme (http only for loopback hosts under the loopback-http profile),
// a host, an optional port, and nothing else.
func validOrigin(o string, loopbackHTTP bool) bool {
	u, err := url.Parse(o)
	if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" ||
		u.Opaque != "" || u.String() != o || strings.ToLower(o) != o {
		return false
	}
	host := u.Hostname()
	for _, l := range strings.Split(host, ".") {
		if !hostLabelRe.MatchString(l) {
			return false
		}
	}
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != p {
			return false
		}
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		h := u.Hostname()
		return loopbackHTTP && (h == "localhost" || h == "127.0.0.1")
	}
	return false
}

func validHostname(h string) bool {
	if len(h) > 253 {
		return false
	}
	labels := strings.Split(h, ".")
	if len(labels) < 2 {
		return false
	}
	for _, l := range labels {
		if !hostLabelRe.MatchString(l) {
			return false
		}
	}
	return true
}
