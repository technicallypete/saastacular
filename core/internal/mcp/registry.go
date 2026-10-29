package mcp

import (
	"context"
	"errors"
	"net/url"

	"core/internal/domains"
)

// Host is the routing entry of one request host: the exact lowercase host[:port], the canonical Organization UUID it
// currently routes to and its setup origin (https://host[:port]).
type Host struct {
	Host           string
	OrganizationID string
	Origin         string
}

// HostLookup resolves an exact request Host value to the Organization it currently routes to (G016 Plan r4 §1.6). The
// production implementation is the durable resolver (Organization slugs and verified, routing-ready custom domains);
// the static MCP_TENANT_HOSTS_FILE registry is retired. Lookup returns domains.ErrUnknownHost for a host that routes
// nowhere (the request is 404 before any store access) and any other error for a store failure (503, never a clean
// unknown host). The value is matched exactly: no case folding or trimming. Organization identity still comes only from
// this answer and the connection record; Host or forwarded headers can only cause refusal.
type HostLookup interface {
	Lookup(ctx context.Context, host string) (organizationID, hostPort string, err error)
}

// ErrUnknownHost is the lookup's unknown-host answer.
var ErrUnknownHost = domains.ErrUnknownHost

// lookupHost returns the routing entry of an exact request Host value.
func lookupHost(ctx context.Context, l HostLookup, host string) (Host, error) {
	org, hp, err := l.Lookup(ctx, host)
	if err != nil {
		return Host{}, err
	}
	if hp != host {
		return Host{}, ErrUnknownHost
	}
	return Host{Host: hp, OrganizationID: org, Origin: "https://" + hp}, nil
}

// lookupOrigin returns the routing entry of an exact setup origin (https://host[:port] and nothing else).
func lookupOrigin(ctx context.Context, l HostLookup, origin string) (Host, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" ||
		u.Fragment != "" || u.Opaque != "" || "https://"+u.Host != origin {
		return Host{}, ErrUnknownHost
	}
	return lookupHost(ctx, l, u.Host)
}

// unknownHost reports whether a lookup error is the unknown-host answer (otherwise it is a store failure).
func unknownHost(err error) bool { return errors.Is(err, ErrUnknownHost) }
