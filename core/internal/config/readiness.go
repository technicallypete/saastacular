package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

// Encryption and transport readiness (G012 Plan r1 §6). AssessReadiness is a pure evaluator over non-secret
// configuration plus declared evidence references. It reports at most "configuration-ready" and never says that
// anything is encrypted: external items are declared, not verified, and the result is not a deployed-platform
// attestation. Paths the code cannot secure today (CurrentCapabilities) stay blocked whatever evidence is declared.

// ReadinessLabel is carried by every report.
const ReadinessLabel = "configuration-only; external items declared, not verified; not a deployed-platform attestation"

// Overall results.
const (
	ReadinessReady   = "configuration-ready"
	ReadinessBlocked = "blocked"
)

// Per-path statuses.
const (
	StatusConfigVerified    = "config-verified"
	StatusExternalDeclared  = "external-declared"
	StatusDisabled          = "disabled"
	StatusNotApplicable     = "not-applicable"
	StatusBlockedPlaintext  = "blocked-plaintext"
	StatusBlockedBypass     = "blocked-verification-bypass"
	StatusBlockedUnsupport  = "blocked-unsupported"
	StatusBlockedUnknown    = "blocked-unknown"
	StatusBlockedLocalExcep = "blocked-local-exception"
)

// Mode is how a path can be secured with the current code.
type Mode string

// Capability modes.
const (
	ModeInPlatform  Mode = "in-platform" // the platform's own configuration secures it; checked from configuration
	ModeExternal    Mode = "external"    // secured outside the platform; needs a declared evidence reference
	ModeUnsupported Mode = "unsupported" // no reviewed way to secure it exists in the code
)

// Readiness paths, in report order (Plan §6.1).
const (
	PathBrowserWeb       = "browser-web"
	PathWebCore          = "web-core"
	PathCorePostgres     = "core-postgres"
	PathMigratePostgres  = "migrate-postgres"
	PathCoreRedis        = "core-redis"
	PathCoreObjectStore  = "core-object-store"
	PathSignedDownload   = "signed-download"
	PathCoreWorkOS       = "core-workos"
	PathCoreIntelligence = "core-intelligence"
	PathGateway          = "intelligence-gateway"
	PathPostgresAtRest   = "postgres-at-rest"
	PathObjectAtRest     = "object-storage-at-rest"
	PathBackupAtRest     = "backups-at-rest"
	PathRedisAtRest      = "redis-at-rest"
	PathAppEnv           = "app-env"
)

var readinessPaths = []string{PathBrowserWeb, PathWebCore, PathCorePostgres, PathMigratePostgres, PathCoreRedis,
	PathCoreObjectStore, PathSignedDownload, PathCoreWorkOS, PathCoreIntelligence, PathGateway, PathPostgresAtRest,
	PathObjectAtRest, PathBackupAtRest, PathRedisAtRest}

// Capabilities gives each path a Mode.
type Capabilities map[string]Mode

// CurrentCapabilities pins today's code facts: core and migrate reach PostgreSQL and core reaches Redis without a
// reviewed TLS configuration (findings F-1; Plan §6.1), so those paths are unsupported. Changing a value needs a
// reviewed code change; readiness_test.go pins this table.
func CurrentCapabilities() Capabilities {
	return Capabilities{
		PathBrowserWeb: ModeExternal, PathWebCore: ModeExternal, PathCorePostgres: ModeUnsupported,
		PathMigratePostgres: ModeUnsupported, PathCoreRedis: ModeUnsupported, PathCoreObjectStore: ModeInPlatform,
		PathSignedDownload: ModeInPlatform, PathCoreWorkOS: ModeInPlatform, PathCoreIntelligence: ModeExternal,
		PathGateway: ModeInPlatform, PathPostgresAtRest: ModeExternal, PathObjectAtRest: ModeExternal,
		PathBackupAtRest: ModeExternal, PathRedisAtRest: ModeInPlatform,
	}
}

// Evidence items: READINESS_EVIDENCE_<ITEM> holds a reference (a document, ticket or record identifier) for an
// externally secured path.
var evidenceItem = map[string]string{
	PathBrowserWeb: "EDGE_TLS", PathWebCore: "CORE_TLS", PathCoreIntelligence: "INTELLIGENCE_TLS",
	PathCorePostgres: "POSTGRES_TLS", PathMigratePostgres: "MIGRATE_TLS", PathCoreRedis: "REDIS_TLS",
	PathCoreObjectStore: "OBJECT_STORE_TLS", PathSignedDownload: "SIGNING_TLS", PathCoreWorkOS: "WORKOS_TLS",
	PathGateway: "GATEWAY_TLS", PathPostgresAtRest: "POSTGRES_AT_REST", PathObjectAtRest: "OBJECT_AT_REST",
	PathBackupAtRest: "BACKUP_AT_REST", PathRedisAtRest: "REDIS_AT_REST",
}

// prerequisite names the deployment work that lifts a blocked path.
var prerequisite = map[string]string{
	PathBrowserWeb:       "external TLS termination for the browser edge with the https cookie profile and https origins",
	PathWebCore:          "TLS for web to core (an https CORE_INTERNAL_URL to an external terminator)",
	PathCorePostgres:     "verified TLS from core to PostgreSQL (no reviewed configuration exists; F-1)",
	PathMigratePostgres:  "verified TLS from migrate to PostgreSQL (no reviewed configuration exists)",
	PathCoreRedis:        "verified TLS from core to Redis, or removing Redis (no reviewed configuration exists)",
	PathCoreObjectStore:  "an https object-store endpoint",
	PathSignedDownload:   "an https signed-download endpoint",
	PathCoreWorkOS:       "the default https WorkOS API base and an https issuer",
	PathCoreIntelligence: "TLS for core to intelligence (an https INTELLIGENCE_URL to an external terminator)",
	PathGateway:          "the fixed https gateway base",
	PathPostgresAtRest:   "evidence of PostgreSQL at-rest encryption by the provider or disk",
	PathObjectAtRest:     "evidence of object-storage at-rest encryption by the provider",
	PathBackupAtRest:     "evidence of backup at-rest encryption (backups are G013 scope)",
	PathRedisAtRest:      "keeping Redis ephemeral",
	PathAppEnv:           "APP_ENV=production",
}

// readinessKeys is the closed set of configuration names the evaluator reads (besides PG* and evidence names).
var readinessKeys = []string{"APP_ENV", "IDENTITY_COOKIE_PROFILE", "IDENTITY_PUBLIC_ORIGINS",
	"IDENTITY_WORKOS_API_BASE_URL", "IDENTITY_WORKOS_ISSUER", "CORE_INTERNAL_URL", "OBJECT_STORE_ENDPOINT",
	"OBJECT_STORE_SIGNING_ENDPOINT", "INTELLIGENCE_URL", "READINESS_GATEWAY_ENABLED"}

// readinessSecretNames are the §2.1 secret names; any of them, or a name containing a secret-like word, is refused.
var readinessSecretNames = []string{"DB_PASSWORD", "DB_MIGRATOR_PASSWORD", "DB_RUNTIME_PASSWORD",
	"DB_RETENTION_PASSWORD", "POSTGRES_SUPERUSER_PASSWORD", "REDIS_PASSWORD", "OBJECT_STORE_SECRET_KEY",
	"MINIO_ROOT_PASSWORD", "INTELLIGENCE_SERVICE_TOKEN", "OPENROUTER_API_KEY", "IDENTITY_WORKOS_API_KEY",
	"IDENTITY_WORKOS_API_KEY_FILE", "IDENTITY_TEST_API_KEY", "IDENTITY_STAGING_WORKOS_API_KEY"}

var (
	readinessKeyRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,127}$`)
	evidenceRefRe  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/#@+-]{2,255}$`)
	booleanLike    = []string{"true", "false", "yes", "no", "y", "n", "1", "0", "on", "off", "ok", "done", "enabled", "disabled", "none", "null", "n/a", "na", "tbd", "todo"}
)

// readinessInputMax bounds the evaluator input; larger input is refused, never truncated.
const readinessInputMax = 64 << 10

// ErrReadinessInput reports invalid evaluator input; its message names keys only, never values.
var ErrReadinessInput = errors.New("invalid readiness input")

// PathStatus is one path's result.
type PathStatus struct {
	Path   string `json:"path"`
	Mode   Mode   `json:"mode"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// Readiness is the evaluator report.
type Readiness struct {
	Result          string       `json:"result"`
	Label           string       `json:"label"`
	Paths           []PathStatus `json:"paths"`
	Prerequisites   []string     `json:"prerequisites"`
	LocalExceptions []string     `json:"local_exceptions"`
}

// ParseReadinessInput reads KEY=VALUE lines (blank lines and # comments ignored). A malformed line, a duplicate key,
// a secret-named key or a key outside the closed input set is refused; errors name keys or line numbers only.
func ParseReadinessInput(r io.Reader) (map[string]string, error) {
	in := map[string]string{}
	b, err := io.ReadAll(io.LimitReader(r, readinessInputMax+1))
	if err != nil {
		return nil, fmt.Errorf("%w: the input cannot be read", ErrReadinessInput)
	}
	if len(b) > readinessInputMax {
		return nil, fmt.Errorf("%w: the input exceeds %d bytes", ErrReadinessInput, readinessInputMax)
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || !readinessKeyRe.MatchString(k) {
			return nil, fmt.Errorf("%w: line %d is not KEY=VALUE", ErrReadinessInput, n)
		}
		if _, dup := in[k]; dup {
			return nil, fmt.Errorf("%w: %s is given more than once (ambiguous)", ErrReadinessInput, k)
		}
		in[k] = v
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%w: line %d cannot be read", ErrReadinessInput, n+1)
	}
	if err := checkReadinessKeys(in); err != nil {
		return nil, err
	}
	return in, nil
}

func secretLikeKey(k string) bool {
	if slices.Contains(readinessSecretNames, k) {
		return true
	}
	for _, w := range []string{"PASSWORD", "SECRET", "TOKEN", "API_KEY"} {
		if strings.Contains(k, w) {
			return true
		}
	}
	return false
}

func checkReadinessKeys(in map[string]string) error {
	var secret, unknown []string
	for k := range in {
		switch {
		case secretLikeKey(k):
			secret = append(secret, k)
		case slices.Contains(readinessKeys, k), strings.HasPrefix(k, "PG"):
		case strings.HasPrefix(k, "READINESS_EVIDENCE_") && knownEvidence(strings.TrimPrefix(k, "READINESS_EVIDENCE_")):
		default:
			unknown = append(unknown, k)
		}
	}
	slices.Sort(secret)
	slices.Sort(unknown)
	if len(secret) > 0 {
		return fmt.Errorf("%w: secret-named input is refused: %s", ErrReadinessInput, strings.Join(secret, ", "))
	}
	if len(unknown) > 0 {
		return fmt.Errorf("%w: unknown keys: %s", ErrReadinessInput, strings.Join(unknown, ", "))
	}
	if g := in["READINESS_GATEWAY_ENABLED"]; g != "" && g != "true" && g != "false" {
		return fmt.Errorf("%w: READINESS_GATEWAY_ENABLED must be true or false", ErrReadinessInput)
	}
	return nil
}

func knownEvidence(item string) bool {
	for _, v := range evidenceItem {
		if v == item {
			return true
		}
	}
	return false
}

// evidenceRef reports whether the path's evidence variable holds a reference; a boolean-like value is not one.
func evidenceRef(in map[string]string, path string) bool {
	v := strings.TrimSpace(in["READINESS_EVIDENCE_"+evidenceItem[path]])
	return evidenceRefRe.MatchString(v) && !slices.Contains(booleanLike, strings.ToLower(v))
}

// AssessReadiness evaluates input against caps. It refuses secret-named or unknown input before evaluation.
func AssessReadiness(input map[string]string, caps Capabilities) (Readiness, error) {
	if err := checkReadinessKeys(input); err != nil {
		return Readiness{}, err
	}
	get := func(k string) string { return strings.TrimSpace(input[k]) }
	appEnv := get("APP_ENV")
	rep := Readiness{Label: ReadinessLabel, Prerequisites: []string{}, LocalExceptions: []string{}}

	for _, p := range readinessPaths {
		mode := caps[p]
		if mode == "" {
			mode = ModeUnsupported
		}
		st, detail := assessPath(p, mode, input, get)
		rep.Paths = append(rep.Paths, PathStatus{Path: p, Mode: mode, Status: st, Detail: detail})
	}
	envStatus, envDetail := StatusConfigVerified, ""
	if appEnv != "production" {
		envStatus = StatusBlockedLocalExcep
		envDetail = "APP_ENV is not production; development allowances (loopback-http cookies, http endpoints, test API overrides) are local exceptions"
		if appEnv != "development" && appEnv != "test" {
			envStatus, envDetail = StatusBlockedUnknown, "APP_ENV is missing or not development, test or production"
		}
	}
	rep.Paths = append(rep.Paths, PathStatus{Path: PathAppEnv, Mode: ModeInPlatform, Status: envStatus, Detail: envDetail})

	rep.Result = ReadinessReady
	for _, ps := range rep.Paths {
		if strings.HasPrefix(ps.Status, "blocked-") {
			rep.Result = ReadinessBlocked
			rep.Prerequisites = append(rep.Prerequisites, ps.Path+": "+prerequisite[ps.Path])
			if ps.Status == StatusBlockedLocalExcep || (appEnv == "development" || appEnv == "test") && localAllowance(ps) {
				rep.LocalExceptions = append(rep.LocalExceptions, ps.Path)
			}
		}
	}
	return rep, nil
}

// localAllowance reports whether a blocked status is one the development and test environments allow.
func localAllowance(ps PathStatus) bool {
	switch ps.Path {
	case PathBrowserWeb, PathWebCore, PathCoreObjectStore, PathSignedDownload, PathCoreWorkOS, PathCoreIntelligence:
		return ps.Status == StatusBlockedPlaintext
	}
	return false
}

func assessPath(p string, mode Mode, in map[string]string, get func(string) string) (string, string) {
	external := func() (string, string) {
		if evidenceRef(in, p) {
			return StatusExternalDeclared, "declared by READINESS_EVIDENCE_" + evidenceItem[p] + "; not verified"
		}
		return StatusBlockedUnknown, "no READINESS_EVIDENCE_" + evidenceItem[p] + " reference"
	}
	// Disabled and not-applicable paths first: they need no capability.
	switch p {
	case PathCoreIntelligence:
		if get("INTELLIGENCE_URL") == "" {
			return StatusDisabled, "INTELLIGENCE_URL is unset"
		}
	case PathGateway:
		if get("READINESS_GATEWAY_ENABLED") != "true" {
			return StatusDisabled, "the AI gateway is not enabled"
		}
	case PathRedisAtRest:
		return StatusNotApplicable, "Redis keeps no persistent data (ephemeral)"
	}
	if p == PathCorePostgres {
		if st, d := pgEnvironment(in); st != "" {
			return st, d
		}
	}
	if mode == ModeUnsupported {
		return StatusBlockedUnsupport, "no reviewed configuration secures this path in the current code"
	}

	switch p {
	case PathBrowserWeb:
		cp := get("IDENTITY_COOKIE_PROFILE")
		if cp == "loopback-http" {
			return StatusBlockedLocalExcep, "IDENTITY_COOKIE_PROFILE loopback-http"
		}
		if cp != "" && cp != "https" {
			return StatusBlockedUnknown, "IDENTITY_COOKIE_PROFILE is not https"
		}
		origins := get("IDENTITY_PUBLIC_ORIGINS")
		if origins == "" {
			return StatusBlockedUnknown, "IDENTITY_PUBLIC_ORIGINS is missing"
		}
		listed := 0
		for _, o := range strings.Split(origins, ",") {
			o = strings.TrimSpace(o)
			if o == "" {
				continue
			}
			listed++
			if strings.HasPrefix(o, "http://") {
				return StatusBlockedPlaintext, "IDENTITY_PUBLIC_ORIGINS lists an http origin"
			}
			if !validOrigin(o, false) {
				return StatusBlockedUnknown, "IDENTITY_PUBLIC_ORIGINS lists an invalid origin"
			}
		}
		if listed == 0 {
			return StatusBlockedUnknown, "IDENTITY_PUBLIC_ORIGINS lists no origin"
		}
	case PathWebCore, PathCoreIntelligence:
		name := map[string]string{PathWebCore: "CORE_INTERNAL_URL", PathCoreIntelligence: "INTELLIGENCE_URL"}[p]
		switch scheme := urlScheme(get(name)); scheme {
		case "https":
		case "http":
			return StatusBlockedPlaintext, name + " is http"
		default:
			return StatusBlockedUnknown, name + " is missing or not an http(s) URL"
		}
	case PathCoreObjectStore, PathSignedDownload:
		data, signing := get("OBJECT_STORE_ENDPOINT"), get("OBJECT_STORE_SIGNING_ENDPOINT")
		if data == "" {
			return StatusBlockedUnknown, "OBJECT_STORE_ENDPOINT is missing"
		}
		if signing == "" {
			signing = data
		}
		name, ep := "OBJECT_STORE_ENDPOINT", data
		if p == PathSignedDownload {
			name, ep = "OBJECT_STORE_SIGNING_ENDPOINT", signing
		}
		// The accepted G009 endpoint rules stay the single source: any production problem blocks both paths.
		if probs := objectEndpointProblems(data, get("OBJECT_STORE_SIGNING_ENDPOINT"), "production"); len(probs) > 0 {
			if s, ok := endpointScheme(ep); ok && s == "http" {
				return StatusBlockedPlaintext, name + " is http"
			}
			return StatusBlockedUnknown, "the object-store endpoints fail the production endpoint rules"
		}
	case PathCoreWorkOS:
		if b := get("IDENTITY_WORKOS_API_BASE_URL"); b != "" && strings.TrimSuffix(b, "/") != IdentityDefaultAPIBase {
			return StatusBlockedLocalExcep, "IDENTITY_WORKOS_API_BASE_URL overrides the default (test only)"
		}
		switch urlScheme(get("IDENTITY_WORKOS_ISSUER")) {
		case "https":
		case "http":
			return StatusBlockedPlaintext, "IDENTITY_WORKOS_ISSUER is http"
		default:
			return StatusBlockedUnknown, "IDENTITY_WORKOS_ISSUER is missing or not an https URL"
		}
	}

	if mode == ModeExternal {
		return external()
	}
	return StatusConfigVerified, ""
}

// pgEnvironment refuses any PG* setting: core's pool reads ambient PG* input ahead of its host override, which is not
// a supported transport configuration (F-1).
func pgEnvironment(in map[string]string) (string, string) {
	var keys []string
	for k := range in {
		if strings.HasPrefix(k, "PG") {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return "", ""
	}
	slices.Sort(keys)
	switch strings.ToLower(strings.TrimSpace(in["PGSSLMODE"])) {
	case "disable", "allow", "prefer":
		return StatusBlockedPlaintext, "PGSSLMODE permits plaintext"
	case "require":
		return StatusBlockedBypass, "PGSSLMODE require does not verify the server certificate"
	}
	return StatusBlockedUnsupport, "ambient PG* settings are not a supported configuration: " + strings.Join(keys, ", ")
}

func urlScheme(s string) string {
	u, err := url.Parse(s)
	if err != nil || u.Host == "" || u.User != nil {
		return ""
	}
	if u.Scheme == "https" || u.Scheme == "http" {
		return u.Scheme
	}
	return ""
}
