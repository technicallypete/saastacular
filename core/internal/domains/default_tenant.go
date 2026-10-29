package domains

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"core/internal/hostname"
)

// G071 (Plan r4 S3, S14, S15; ADR-010): the optional default-tenant root alias and the multi-tenant root presentation,
// both selected once at startup from deployment configuration. The alias is pinned in the resolver for the life of the
// process (WithDefaultTenant); a different slug needs a restart. Neither value is ever searched per request.

// ErrDefaultTenantUnavailable is the fixed error of a default-tenant lookup that could not complete (a store, transport
// or permission failure, never a clean unknown slug). The server returns it from run, and core.exit records it as the
// default_tenant_unavailable category; the underlying error text is never kept or logged.
var ErrDefaultTenantUnavailable = errors.New("default tenant lookup unavailable")

// The default_tenant outcomes and inactive reasons (the closed S14 schema).
const (
	DefaultTenantActive   = "active"
	DefaultTenantInactive = "inactive"

	DefaultTenantInvalidSlug = "invalid_slug"
	DefaultTenantUnknownSlug = "unknown_slug"
)

// The PLATFORM_ROOT_PAGE modes (S15): exact values, no case folding.
const (
	RootPageSignupRedirect = "signup_redirect"
	RootPageLanding        = "landing"
)

// The startup log events (S14).
const (
	eventDefaultTenant    = "default_tenant"
	eventPlatformRootPage = "platform_root_page"
)

// defaultTenantSQL is the single startup statement: the auth-reader body of migration 00022, executable by the runtime
// role on the resolver's own pool (Plan r4 §3). It returns NULL for a slug no Organization holds.
const defaultTenantSQL = `SELECT app.organization_id_by_slug($1)::text`

// defaultTenantTimeout bounds the one startup lookup.
const defaultTenantTimeout = 10 * time.Second

// DefaultTenant is the startup outcome of DEFAULT_TENANT_SLUG. Outcome is "" when the setting is unset. Slug holds the
// canonical slug only for an active alias or an unknown slug; an invalid value is never kept, so it cannot be logged.
// OrganizationID is set only when the alias is active.
type DefaultTenant struct {
	Outcome        string
	Reason         string
	Slug           string
	OrganizationID string
}

// Active reports whether the exact platform root aliases an Organization.
func (t DefaultTenant) Active() bool {
	return t.Outcome == DefaultTenantActive && t.OrganizationID != ""
}

// SelectDefaultTenant judges the trimmed DEFAULT_TENANT_SLUG value raw once at startup (S3): "" is unset (no statement);
// a value outside the Organization slug grammar or reserved as a first label under cfg, exactly as given (no case
// folding or other normalization), is inactive/invalid_slug without a statement; otherwise one statement on q (the
// resolver's pool) under a 10 s bound reads the Organization id: NULL is inactive/unknown_slug, an id is active. Any
// store, transport or permission failure is ErrDefaultTenantUnavailable: an outage is never treated as absence.
func SelectDefaultTenant(ctx context.Context, q Querier, cfg hostname.Config, raw string) (DefaultTenant, error) {
	raw = strings.TrimSpace(raw) // environment whitespace only (config.LoadDomains already trims); nothing else
	if raw == "" {
		return DefaultTenant{}, nil
	}
	if !hostname.SlugGrammar(raw) || hostname.IsReservedLabel(raw, cfg) {
		return DefaultTenant{Outcome: DefaultTenantInactive, Reason: DefaultTenantInvalidSlug}, nil
	}
	lctx, cancel := context.WithTimeout(ctx, defaultTenantTimeout)
	defer cancel()
	var org *string
	if err := q.QueryRow(lctx, defaultTenantSQL, raw).Scan(&org); err != nil {
		return DefaultTenant{}, ErrDefaultTenantUnavailable
	}
	if org == nil || *org == "" {
		return DefaultTenant{Outcome: DefaultTenantInactive, Reason: DefaultTenantUnknownSlug, Slug: raw}, nil
	}
	return DefaultTenant{Outcome: DefaultTenantActive, Slug: raw, OrganizationID: *org}, nil
}

// Log writes the S14 default_tenant record of the outcome: info outcome=active with slug and organization_id; warn
// outcome=inactive with reason (and the slug only for unknown_slug: an invalid value is never logged). An unset setting
// writes nothing.
func (t DefaultTenant) Log(ctx context.Context, log *slog.Logger) {
	switch {
	case t.Active():
		log.LogAttrs(ctx, slog.LevelInfo, eventDefaultTenant, slog.String("outcome", DefaultTenantActive),
			slog.String("slug", t.Slug), slog.String("organization_id", t.OrganizationID))
	case t.Outcome == DefaultTenantInactive && t.Reason == DefaultTenantUnknownSlug:
		log.LogAttrs(ctx, slog.LevelWarn, eventDefaultTenant, slog.String("outcome", DefaultTenantInactive),
			slog.String("reason", DefaultTenantUnknownSlug), slog.String("slug", t.Slug))
	case t.Outcome == DefaultTenantInactive:
		log.LogAttrs(ctx, slog.LevelWarn, eventDefaultTenant, slog.String("outcome", DefaultTenantInactive),
			slog.String("reason", DefaultTenantInvalidSlug))
	}
}

// RootPresentation resolves the trimmed PLATFORM_ROOT_PAGE value raw to the effective multi-tenant root presentation
// (S15): "" and signup_redirect are signup_redirect, landing is landing, and anything else falls back to
// signup_redirect with invalid set (the caller writes LogInvalidRootPage). Values are exact; no Organization count is
// inferred.
func RootPresentation(raw string) (mode string, invalid bool) {
	switch strings.TrimSpace(raw) {
	case "", RootPageSignupRedirect:
		return RootPageSignupRedirect, false
	case RootPageLanding:
		return RootPageLanding, false
	default:
		return RootPageSignupRedirect, true
	}
}

// LogInvalidRootPage writes the S14 platform_root_page record of an invalid PLATFORM_ROOT_PAGE value: warn
// outcome=invalid_value fallback=signup_redirect and nothing else (the raw value is never logged).
func LogInvalidRootPage(ctx context.Context, log *slog.Logger) {
	log.LogAttrs(ctx, slog.LevelWarn, eventPlatformRootPage, slog.String("outcome", "invalid_value"),
		slog.String("fallback", RootPageSignupRedirect))
}
