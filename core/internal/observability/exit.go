package observability

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"regexp"
	"sort"
	"strings"
)

const exitEvent = "core.exit"

// Exit categories of a core.exit record.
const (
	exitConfigInvalid            = "config_invalid"
	exitDatabasePoolInvalid      = "database_pool_invalid"
	exitDefaultTenantUnavailable = "default_tenant_unavailable"
	exitListenFailed             = "listen_failed"
	exitServeFailed              = "serve_failed"
	exitShutdownTimeout          = "shutdown_timeout"
	exitUnknown                  = "unknown"
)

// configErrorPrefix, poolError and defaultTenantError are the fixed texts of
// the configuration, database pool and default-tenant lookup errors that run
// returns (config.Load, config.LoadIdentity, cmd/server and, G071,
// domains.ErrDefaultTenantUnavailable).
const (
	configErrorPrefix  = "invalid configuration: "
	poolError          = "cannot configure database pool"
	defaultTenantError = "default tenant lookup unavailable"
)

// configKeys is the allow-list of the environment variable names config.Load
// and config.LoadIdentity read. Only these exact names can appear in a
// config_invalid record; unit tests keep the list in sync with both loaders.
var configKeys = map[string]bool{
	"APP_ENV":                             true,
	"CORE_LISTEN_ADDR":                    true,
	"CUSTOM_DOMAIN_CNAME_TARGET":          true,
	"DATABASE_NAME":                       true,
	"DB_HOST":                             true,
	"DB_PASSWORD":                         true,
	"DB_PORT":                             true,
	"DB_USER":                             true,
	"IDENTITY_COOKIE_PROFILE":             true,
	"IDENTITY_JWT_ALGORITHMS":             true,
	"IDENTITY_MAX_AUTH_AGE":               true,
	"IDENTITY_PROVIDER":                   true,
	"IDENTITY_PUBLIC_ORIGINS":             true,
	"IDENTITY_SESSION_ABSOLUTE":           true,
	"IDENTITY_SESSION_IDLE":               true,
	"IDENTITY_WORKOS_API_BASE_URL":        true,
	"IDENTITY_WORKOS_API_KEY":             true,
	"IDENTITY_WORKOS_API_KEY_FILE":        true,
	"IDENTITY_WORKOS_CLIENT_ID":           true,
	"IDENTITY_WORKOS_ISSUER":              true,
	"INTELLIGENCE_SERVICE_TOKEN":          true,
	"INTELLIGENCE_URL":                    true,
	"OBJECT_STORE_ACCESS_KEY":             true,
	"OBJECT_STORE_BUCKET":                 true,
	"OBJECT_STORE_ENDPOINT":               true,
	"OBJECT_STORE_FORCE_PATH_STYLE":       true,
	"OBJECT_STORE_MAX_OBJECT_BYTES":       true,
	"OBJECT_STORE_REGION":                 true,
	"OBJECT_STORE_SECRET_KEY":             true,
	"OBJECT_STORE_SIGNED_URL_DEFAULT_TTL": true,
	"OBJECT_STORE_SIGNED_URL_MAX_TTL":     true,
	"OBJECT_STORE_SIGNING_ENDPOINT":       true,
	"OBJECT_STORE_TRANSFER_TIMEOUT":       true,
	"PLATFORM_BASE_DOMAIN":                true,
	"PRODUCT_NAME":                        true,
	"REDIS_HOST":                          true,
	"REDIS_PASSWORD":                      true,
	"REDIS_PORT":                          true,
}

var configKeyRe = regexp.MustCompile(`[A-Z][A-Z0-9_]*`)

// LogExit writes exactly one core.exit JSON record for an error returned by
// the server's run function. The record holds a fixed error_type category
// and, for invalid configuration, the allow-listed variable names involved.
// It never contains the error's text.
func LogExit(w io.Writer, err error) {
	errorType, keys := classifyExit(err)
	attrs := []slog.Attr{slog.String("error_type", errorType)}
	if errorType == exitConfigInvalid {
		attrs = append(attrs, slog.Any("config_keys", keys))
	}
	NewLogger(w, slog.LevelInfo, nil).LogAttrs(context.Background(), slog.LevelError, exitEvent, attrs...)
}

// classifyExit maps err onto a core.exit category. Error text is inspected
// only to recognize the fixed configuration and pool errors; it is never
// returned.
func classifyExit(err error) (string, []string) {
	for e := err; e != nil; e = errors.Unwrap(e) {
		msg := e.Error()
		if strings.HasPrefix(msg, configErrorPrefix) {
			return exitConfigInvalid, allowedConfigKeys(msg)
		}
		if msg == poolError {
			return exitDatabasePoolInvalid, nil
		}
		if msg == defaultTenantError {
			return exitDefaultTenantUnavailable, nil
		}
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Op == "listen" {
			return exitListenFailed, nil
		}
		return exitServeFailed, nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return exitShutdownTimeout, nil
	}
	return exitUnknown, nil
}

// allowedConfigKeys returns the sorted, de-duplicated allow-listed variable
// names that occur in a configuration error message. It is never nil.
func allowedConfigKeys(msg string) []string {
	seen := map[string]bool{}
	keys := []string{}
	for _, k := range configKeyRe.FindAllString(msg, -1) {
		if configKeys[k] && !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// ListenPort returns the port of a bound TCP listener address, or 0 for any
// other address.
func ListenPort(a net.Addr) int {
	if t, ok := a.(*net.TCPAddr); ok && t != nil {
		return t.Port
	}
	return 0
}
