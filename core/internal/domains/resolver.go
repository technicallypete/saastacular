package domains

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"core/internal/hostname"
)

// Querier is satisfied by *pgxpool.Pool and pgx.Tx: ResolveWith runs its one statement on whatever the caller passes
// (Plan r4 §5.5), so it owns no transaction.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Resolution is the routing decision of one Host value. OrganizationID is set only when Routing or Issuance holds for
// a custom host, or for a platform slug that exists, or for the exact platform root aliased to the default tenant.
type Resolution struct {
	Host           hostname.Host
	Kind           hostname.Kind
	OrganizationID string
	Routing        bool // the host currently routes to OrganizationID
	Issuance       bool // a certificate may be issued for the host (custom hosts only; the E1 tls-ask answer)
	RootAlias      bool // G071: the exact platform root, resolved as a platform host of the pinned default tenant
}

// hostnameStateSQL is the resolver's single statement: one unique-index probe in the auth-reader function.
const hostnameStateSQL = `SELECT organization_id::text, routing, issuance FROM app.hostname_state($1, $2)`

// ResolveWith normalizes and classifies a raw Host value under cfg and, for platform and custom names only, reads the
// routing and issuance predicates with one statement on q. A malformed host is returned as KindNone without error; a
// store failure is an error (503), never a clean unknown host. No cache: invalidation is the commit.
func ResolveWith(ctx context.Context, q Querier, cfg hostname.Config, raw string) (Resolution, error) {
	h, err := hostname.Normalize(raw, cfg.Port)
	if err != nil {
		return Resolution{Kind: hostname.KindNone}, nil
	}
	c := hostname.Classify(h, cfg)
	res := Resolution{Host: h, Kind: c.Kind}
	key := h.Name
	switch c.Kind {
	case hostname.KindPlatform:
		key = c.Slug
	case hostname.KindCustom:
	default:
		return res, nil
	}
	var org *string
	var routing, issuance bool
	err = q.QueryRow(ctx, hostnameStateSQL, string(c.Kind), key).Scan(&org, &routing, &issuance)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return res, nil
	case err != nil:
		return Resolution{}, err
	}
	if org != nil {
		res.OrganizationID, res.Routing, res.Issuance = *org, routing, issuance
	}
	return res, nil
}

// Resolver is the durable resolver on the pool: one connection acquisition and one autocommit statement per call.
type Resolver struct {
	q   Querier
	cfg hostname.Config
	// defaultTenant is the Organization id the exact platform root aliases (G071), pinned once at startup; "" = none.
	defaultTenant string
}

// NewResolver returns a resolver on q (the pool) for cfg.
func NewResolver(q Querier, cfg hostname.Config) *Resolver { return &Resolver{q: q, cfg: cfg} }

// WithDefaultTenant pins the Organization the exact platform root aliases (G071 Plan r4 S4) and returns r. The server
// calls it once, before serving, with the id SelectDefaultTenant selected ("" pins nothing); it is never changed while
// the process runs.
func (r *Resolver) WithDefaultTenant(organizationID string) *Resolver {
	r.defaultTenant = organizationID
	return r
}

// Config returns the classification configuration.
func (r *Resolver) Config() hostname.Config { return r.cfg }

// Resolve resolves one raw Host value on the pool.
func (r *Resolver) Resolve(ctx context.Context, raw string) (Resolution, error) {
	return r.resolve(ctx, r.q, raw)
}

// ResolveWith resolves on the caller's querier (a transaction) with the resolver's configuration.
func (r *Resolver) ResolveWith(ctx context.Context, q Querier, raw string) (Resolution, error) {
	return r.resolve(ctx, q, raw)
}

// resolve is ResolveWith plus the pinned root alias (G071 S4): with a default tenant pinned, the exact platform root
// (a public entry name that is the platform domain itself) is a routing platform host of that Organization, answered
// without a statement. www.<domain>, the auth host, the CNAME target, reserved and tenant labels, custom hosts and
// malformed hosts resolve exactly as without the alias.
func (r *Resolver) resolve(ctx context.Context, q Querier, raw string) (Resolution, error) {
	if r.defaultTenant != "" {
		if h, err := hostname.Normalize(raw, r.cfg.Port); err == nil &&
			hostname.Classify(h, r.cfg).Kind == hostname.KindPublicEntry && hostname.IsPlatformRoot(h, r.cfg) {
			return Resolution{Host: h, Kind: hostname.KindPlatform, OrganizationID: r.defaultTenant, Routing: true,
				RootAlias: true}, nil
		}
	}
	return ResolveWith(ctx, q, r.cfg, raw)
}

// Lookup returns the Organization an exact Host value routes to and its normalized host[:port]. The value must already
// be in normalized form (no case folding or trailing-dot stripping is admitted here: the MCP exactness rule). Anything
// that does not route is ErrUnknownHost; a store failure is returned as is.
func (r *Resolver) Lookup(ctx context.Context, raw string) (organizationID, hostPort string, err error) {
	res, err := r.Resolve(ctx, raw)
	if err != nil {
		return "", "", err
	}
	if !res.Routing || res.Host.HostPort() != raw {
		return "", "", ErrUnknownHost
	}
	return res.OrganizationID, raw, nil
}

// IssuanceAllowed is the E1 tls-ask answer for a bare hostname (no port): true only for a custom hostname whose live row
// has a fresh ownership proof (verified, active or TLS-suspended). Platform hosts use the shared wildcard certificate.
func (r *Resolver) IssuanceAllowed(ctx context.Context, name string) (bool, error) {
	if name == "" || strings.Contains(name, ":") {
		return false, nil
	}
	res, err := ResolveWith(ctx, r.q, hostname.Config{BaseDomain: r.cfg.BaseDomain, CNAMETarget: r.cfg.CNAMETarget,
		AuthHost: r.cfg.AuthHost, ReservedLabels: r.cfg.ReservedLabels}, name)
	if err != nil {
		return false, err
	}
	return res.Kind == hostname.KindCustom && res.Issuance && res.Host.Name == name, nil
}
