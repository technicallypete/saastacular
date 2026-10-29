# ADR-005 — Tenant-bound external MCP connections

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Status: ACCEPTED by product owner, 2026-10-11
Authority: owner approved separate tenant-specific MCP connection URLs, immutable tenant binding per connection and within-tenant Group switching, then directed updates to contracts, architecture and related documentation.

## Decision

Each external MCP connection is authorized for exactly one canonical Organization UUID. Its tenant-specific URL is a registered/verified routing selector; authentication and current Organization membership remain mandatory. Another tenant requires another authorized connection. Remove in-connection Organization switching from the planned MCP surface.

Retain reusable optional Group selection within the bound Organization, including explicit clearing to Organization-only context. Group selection, MCP session authorization and capability catalogs are local to that connection. Identity remains global; authority does not. A connection/session/tenant-specific transport credential for A cannot be replayed as B, even if the same user belongs to both. A shared global identity login may support establishing each separately authorized connection, but is not a tenant grant.

Every execution checks current session/user state, Organization membership, selected Group and exact capability/resource ownership. Tool discovery is filtered and invalidated on relevant changes, never a substitute for execution checks. Unknown routes, mismatched endpoint/session/credential bindings, foreign selectors and revoked authority fail closed, without switching or fallback. Slugs/hostnames are alternate routing identifiers; existing bindings retain their canonical UUID.

## Scope and dependencies

G014 owns the reviewed endpoint/transport/authentication/session contract and control-plane Group tools. G015 owns dynamic catalogs, concurrent connection isolation and stale-context enforcement. G016 adds verified tenant hostname/custom-domain aliases. Preserve G014→G015→G016 sequencing: G014 may use a server-registered tenant-specific route on a configured shared origin before G016. No arbitrary Host/forwarded-header shortcut or early DNS/TLS lifecycle implementation.

The implementation-ready behavioral requirements are in [the connection contract](../../contracts/mcp/connection-context.md). Exact URL spelling, protocol authentication, audience/resource validation and connection/session lifecycle handling must be pinned in the G014 Plan; this record selects no vendor or wire schema.

## Supersession and transition

Supersedes Organization-switching portions of Architecture §§3/4/52/53.9, EPIC-005's former G014/G015 wording and historical G000/archetype MCP examples. Preserve accepted historical Plans/Specs/Results and existing archetype implementation/tests; they are not evidence of the new connection model. G014 must reconcile affected reusable MCP guidance/fixtures under reviewed scope before claiming the new contract is implemented; G021 verifies alignment before releasing product work or OS001 export.

Active G007/G051 execution contracts are unchanged: this is a future adapter decision, not a new test requirement for their internal policy or CI increments. No application code, schema, runtime skill/test implementation, authentication provider configuration or endpoint is changed here. Existing owner integration authority and the G021 foundation barrier remain.

## Owner UX addendum — 2026-10-15

Owner approved recording per-connection tenant-hosted URLs:
https://{organization_slug}.${PLATFORM_BASE_DOMAIN}/mcp/c/{connection_uuid}.
The intended UX has one connection per user/Organization/AI client, with optional
Group switching within it; multiple independent connections remain valid. Reusing
a URL reuses the logical connection and its selection/logout state. All users use existing local permissions [Editorial redaction:
source-product user types removed]. Routine credential renewal
or reauthentication retains a still-valid connection URL; terminal revocation,
invalidated/missing state and expired pending setup cannot be revived from tokens.

This supersedes URL-spelling deferral for the intended user-facing layout only.
G014 must qualify configured-host/provider audience feasibility without adding
G016 provisioning/DNS/TLS. WorkOS per-connection issuance is the proposal selected
for bounded Plan revision, still subject to issuer/subject/resource and refresh-
retarget qualification and formal Plan/security acceptance. Provider issuance can
continue after local revocation while Beginify denies protected access. The
Spec's dated amendment pins that distinction; no provider change or implementation
is authorized by this addendum. Shared-origin historical Plans remain preserved.

## Owner decision — WorkOS grants and local connection boundaries, 2026-10-15

The owner answered "Ok" to retaining WorkOS for login and MCP token issuance
(assessment option B), with a supported compromised-client revocation gate.
This supersedes the earlier independent refresh-family requirement and provider
negative-retarget acceptance condition; the failed QP-3 remains historical evidence.
No dedicated MCP issuer is selected. Assessment r1 is recorded in
G014's assessment review (not included).

Global identity and provider user/client grants may span already-authorized
resources. Local MCP connections remain bound to one canonical Organization,
with separate stored Group and lifecycle state. Raw access-token/session replay
from A to B remains refused; a valid separately issued B-audience token must meet
B's exact user/client binding, live membership and local authority. The owner
accepts broader credential reach; future replacement, cross-host and unconsented
reach were not proved by the single observed retarget.

Local connection logout does not terminate the provider grant. A compromised
client needs separately authorized provider grant/family revocation plus local
connection invalidation. G014 implementation is held until supported provider
revocation and old-family refusal after fresh provider authorization are qualified,
without future app bootstrap. Local invalidation/replacement holds and old-access
refusal are proved during approved implementation before acceptance. Unknown/terminal
records never revive from tokens; new authenticated browser bootstrap creates a
replacement, without requiring fresh password entry or a separate provider grant.

Architecture and the connection contract carry this interpretation; the dated
G014 Spec amendment supplies objective evidence gates. Reconcile the Plan before
acceptance. This accepts the model, not a provider revocation guarantee, new
experiment, dependency, settings change or implementation release. All final
G013, Plan/input/budget/resource and G021 gates remain.


## Owner-approved G016 domain/authentication refinement — 2026-10-16

[ADR-007](ADR-007-verified-domain-routing-and-authentication.md) records the owner's
explicit approval of tenant-host membership, shared WorkOS callback relay and
exact-host historical MCP retention, with unchanged canonical identity/authority.
It preserves this ADR's historical evidence. Design approval alone grants no live
provider/settings/deployment action or G016 execution release.
