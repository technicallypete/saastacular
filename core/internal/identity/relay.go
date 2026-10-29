package identity

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
)

// G016 (Plan r4 §1.7; ADR-007 D-2/D-3): tenant-host origins are validated by resolving their host, not by a static
// list; IDENTITY_PUBLIC_ORIGINS stays the exact allow-list of non-tenant origins (the shared auth host). Identity
// receives the resolver through the interface below, which it declares itself: it never imports the domains package.

// Querier is one statement's database boundary: the pool (one autocommit statement) or the caller's transaction.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Host kinds of a HostResolution.
const (
	HostPlatform = "platform"
	HostCustom   = "custom"
)

// HostResolution is the routing answer for one exact origin: Kind is HostPlatform or HostCustom for a tenant host that
// currently routes to OrganizationID (Routing true); anything else routes nowhere.
type HostResolution struct {
	Kind           string
	OrganizationID string
	Routing        bool
}

// HostResolver resolves an exact origin (https://host[:port]) with one statement on q. A host that routes nowhere is a
// zero HostResolution and no error; an error is a store failure (never a clean unknown host).
type HostResolver interface {
	ResolveOrigin(ctx context.Context, q Querier, origin string) (HostResolution, error)
}

// tenant is a resolved tenant-host origin.
type tenant struct {
	Origin         string
	Kind           string
	OrganizationID string
}

// tenantOrigin resolves origin on q: a routing platform or custom host, or ErrOriginRejected (an unconfigured resolver,
// a non-tenant origin such as the shared auth host, or a host that routes nowhere). A store failure is unavailable.
func (s *Service) tenantOrigin(ctx context.Context, q Querier, origin string) (tenant, error) {
	if s.cfg.Hosts == nil || origin == "" || len(origin) > 300 {
		return tenant{}, ErrOriginRejected
	}
	res, err := s.cfg.Hosts.ResolveOrigin(ctx, q, origin)
	if err != nil {
		return tenant{}, ErrUnavailable
	}
	if !res.Routing || res.OrganizationID == "" || (res.Kind != HostPlatform && res.Kind != HostCustom) {
		return tenant{}, ErrOriginRejected
	}
	return tenant{Origin: origin, Kind: res.Kind, OrganizationID: res.OrganizationID}, nil
}

// SessionHost resolves the session route's ?origin= with one pool statement: the routing tenant host's resolution, or
// ErrUnknownHost for an origin that routes nowhere (a removed or reassigned host; Plan r4 §1.7.1 case C).
func (s *Service) SessionHost(ctx context.Context, origin string) (HostResolution, error) {
	t, err := s.tenantOrigin(ctx, s.store.db, origin)
	switch {
	case err == ErrOriginRejected:
		return HostResolution{}, ErrUnknownHost
	case err != nil:
		return HostResolution{}, err
	}
	return HostResolution{Kind: t.Kind, OrganizationID: t.OrganizationID, Routing: true}, nil
}

// stateTagSep separates a login state's random part from its origin tag.
const stateTagSep = "."

// taggedState appends the initiating origin to a custom-host login state: <random>.<base64url(origin)>. The tag is
// opaque to the provider and carries no secret; the stored state hash covers the whole value.
func taggedState(random, origin string) string {
	return random + stateTagSep + base64.RawURLEncoding.EncodeToString([]byte(origin))
}

// stateOrigin decodes the origin tag of a relayed state ("" when absent or undecodable).
func stateOrigin(state string) string {
	i := strings.LastIndex(state, stateTagSep)
	if i < 1 || i == len(state)-1 {
		return ""
	}
	b, err := base64.RawURLEncoding.DecodeString(state[i+1:])
	if err != nil || len(b) == 0 || len(b) > 300 {
		return ""
	}
	o := string(b)
	u, err := url.Parse(o)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" ||
		u.Fragment != "" || u.Opaque != "" || u.String() != o {
		return ""
	}
	return o
}

// The closed relay refusal reasons (auth.relay).
const (
	relayUnconfigured = "relay_unconfigured"
	relayStateInvalid = "state_invalid"
	relayNotCustom    = "not_custom_host"
	relayUnavailable  = "unavailable"
)

// Relay is the shared callback relay decision on the auth host (Plan r4 §1.7 leg 2; B2): it reads only the state
// value's origin tag and answers the initiating custom host's callback URL when, and only when, that origin is a
// routing-ready custom host and the relay is configured. The destination is derived from resolution, never from a
// free-form parameter. Nothing is stored, consumed or set; the record carries the outcome and a closed reason only.
// The caller (web on the auth host) appends the provider's raw query byte-identical.
func (s *Service) Relay(ctx context.Context, state string) (callback string, err error) {
	reason := ""
	defer func() {
		attrs := []slog.Attr{slog.String("outcome", "relayed")}
		if reason != "" {
			attrs = []slog.Attr{slog.String("outcome", "refused"), slog.String("reason", reason)}
		}
		s.event(ctx, slog.LevelInfo, "auth.relay", attrs...)
	}()
	if s.cfg.AuthOrigin == "" || s.cfg.Hosts == nil {
		reason = relayUnconfigured
		return "", ErrRelayRefused
	}
	origin := ""
	if state != "" && len(state) <= 512 {
		origin = stateOrigin(state)
	}
	if origin == "" {
		reason = relayStateInvalid
		return "", ErrRelayRefused
	}
	t, err := s.tenantOrigin(ctx, s.store.db, origin)
	switch {
	case err == ErrUnavailable:
		reason = relayUnavailable
		return "", ErrUnavailable
	case err != nil || t.Kind != HostCustom:
		reason = relayNotCustom
		return "", ErrRelayRefused
	}
	return CallbackURL(t.Origin), nil
}
