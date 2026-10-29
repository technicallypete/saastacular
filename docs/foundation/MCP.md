# Tenant-bound MCP connections

Core serves MCP to external AI clients as tenant-bound connections: each connection has its own URL and is bound to one
Organization, one user and one client.
The wire contract is [contracts/mcp/control-plane.md](../../contracts/mcp/control-plane.md); the behavioral contract is
[contracts/mcp/connection-context.md](../../contracts/mcp/connection-context.md). This guide covers configuration,
the setup flow, the moving parts, the evidence and the limits.

## How it works

1. A signed-in member opens My Account (`https://{tenant-host}/account`) or the compatibility page
   `https://{tenant-host}/connections/mcp`, which render the same connection manager, and creates a connection,
   optionally with a friendly name. Core resolves the
   Organization from the request host ([DOMAINS.md](DOMAINS.md): an Organization slug under the platform domain or a
   routing-ready verified custom domain), checks the browser session, the CSRF token and the membership, and
   creates a pending connection with its own URL `https://{tenant-host}/mcp/c/{uuid}` and its own `mcp` session row.
2. The member pastes the URL into an AI client. The client discovers the authorization server from the per-connection
   protected-resource metadata, signs in through WorkOS AuthKit Connect and receives an access token whose only audience
   is that URL.
3. Every call is verified in core (issuer, audience, lifetime, client, subject mapping through the identity principal),
   matched against the stored connection (host, user, client, live session), then reauthorized against the user's
   current membership and selected Group by the access facade ([AUTHORIZATION.md](AUTHORIZATION.md)). The first call
   activates the connection and binds the client.
4. The connection ends with the `logout` tool, a revoke in the connection manager, sign-out everywhere, `DisableUser`, idle or
   absolute expiry, or a restore ([RECOVERY.md](RECOVERY.md)). Signing out one browser session leaves connections working.

**Friendly names.** A connection may carry a name chosen by its owner (at creation, or later through rename;
an empty name clears it, and "Unnamed connection" is shown for an unnamed or legacy row). The name is presentation
metadata only: it is never identity, client binding, tenant selection, authority or a claim about the client (the
manager shows the actually bound client separately), and it never enters the MCP endpoint, tools, discovery or the
transport binding reads. Core's Unicode rule is authoritative ([the contract](../../contracts/mcp/control-plane.md#internal-setup-contract-core-internal-network-only));
names are rendered as escaped text. A rename needs the caller's live browser session, the CSRF token, current
membership in the host's Organization and a target that is one of the caller's own live pending or active
connections; unknown, foreign and terminal targets answer the same `not_found`. It locks, in order, the live user, the
browser session, the target connection's `mcp` session (FOR SHARE: the locking read sees a committed revocation and
blocks a concurrent one until the rename commits), the Organization facts and finally the connection row, the same
user → session → row order as activation, selection and `revokeSession`, so it cannot deadlock with them. It updates
`display_name` only (audited as `mcp.connection_renamed` with no name) and never revives, renews or extends anything.

A user with access to several Organizations creates one connection per Organization (on that Organization's host).
Two connections to the same Organization keep separate stored Group selections. One URL used by two token families of
the same client shares that connection's state. There is no claim of credential isolation between connections of one
user and client: the provider's refresh grant is per user and client.

## Configuration (core)

All MCP variables unset is the disabled state: no MCP route, metadata or setup route exists and the setup page reports
that connections are unavailable. A partial configuration fails startup with one `core.exit` `config_invalid` record
naming variables only.

| Variable | Meaning |
|---|---|
| `MCP_CONNECT_ISSUER` | the exact AuthKit Connect issuer (https; http only with `APP_ENV=test`); its key set is `<issuer>/oauth2/jwks` |
| `MCP_TENANT_HOSTS_FILE` | **retired**: the former static host registry, replaced by durable host resolution ([DOMAINS.md](DOMAINS.md)); set in any environment it is `config_invalid` (B-MCP-E2E E2E-MCP-6) |
| `MCP_ALLOWED_ORIGINS` | browser origins allowed to call the endpoint (none by default; native clients send no `Origin`) |
| `MCP_PENDING_TTL` | unused-connection lifetime, 2–60 min (default 30 min) |
| `MCP_CONNECTION_IDLE` | idle lifetime, 24 h–90 days (default 30 days); renewal at most once an hour, capped at the absolute lifetime |
| `MCP_CONNECTION_ABSOLUTE` | absolute lifetime, 24 h–730 days (default 180 days), not below the idle lifetime |
| `MCP_CONNECT_JWKS_FILE` | **test only**: a fixed key set (C10), accepted only with `APP_ENV=test`; refused in production (B-IDENTITY-GUARDS N11, B-MCP-E2E) |

MCP requires the WorkOS identity configuration ([IDENTITY.md](IDENTITY.md)) and migration version 10 ([MIGRATIONS.md](MIGRATIONS.md)); the
host-resolution code needs version 11 applied first (migrate, then deploy). Tenant hosts are data, not configuration: the durable resolver behind
`mcp.HostLookup` answers from Organization slugs and verified domains on every request (one autocommit statement before
admission, no cache).

## Hosts and aliases

A connection stores the exact `host[:port]` and resource it was created on. A request is admitted only when its host
resolves to the bound Organization **and** equals the stored host ([connection-context.md](../../contracts/mcp/connection-context.md)):

- an alias that no longer routes (a removed custom domain, a changed slug) is 404 before any store access; a host now
  routing to another Organization is 401; a different current alias of the same Organization is `resource_mismatch`;
- nothing is retargeted and nothing is revoked automatically; a new alias needs a new connection created on it;
- **setup listing**: the setup page lists every connection of the user in the Organization the current host routes to,
  whatever host each was created on, and shows each connection's host next to its URL; revoke is by connection ID
  through any current host of that Organization. Create keeps the exact-origin check (the new URL is on the current
  host). `control-plane.md` is unchanged.

## Components

| Part | Where |
|---|---|
| Schema | `00010_mcp_connections.sql`: `sessions.kind`, `app.mcp_connections`, three Organization-bound policies, the guard trigger, two audit actions; `00019_mcp_connection_names.sql`: the nullable `display_name`, its CHECK, two runtime column grants and the `mcp.connection_renamed` capture ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md), [AUDIT.md](AUDIT.md)) |
| Identity | `core/internal/identity/mcp*.go`: connection records, activation, renewal, revocation, subject mapping |
| Token validation | `core/internal/identity/workos/connect.go`: the Connect validator (bounded key-set fetch, or the test anchor) |
| Facade | `core/internal/tenancy/access/mcp.go`: the verified request scope and the guarded MCP operations |
| Transport | `core/internal/mcp`: host lookup (`registry.go`, the resolver adapter), metadata, admission, the stateless streamable handler, the five tools, the internal setup routes |
| Catalog | `core/internal/mcp/catalog.go`: the closed catalog spec, one requirement per tool, registration that depends on the requirement |
| Web | the shared server-rendered manager `web/components/account/connection-manager.tsx`, embedded in `/account` through `web/components/account/account-sections.tsx` and in `web/app/connections/mcp/` (the compatibility page; the create, revoke and rename form routes), and the server-only `web/lib/mcp-client.ts` |
| Test overlay | `infra/compose/identity.yml`, `infra/identity/{mcp-seed.sql,proxy.conf}` (slugs, domains and the historical old-host connection), `infra/identity/mcp-hosts.tsv` (the harness's host-to-UUID map only; core reads no registry), identitysetup's Connect test signer ([IDENTITY.md](IDENTITY.md)) |

Logs: `mcp.request` for every protected request refused before the tools run, `mcp.connection` for activation, renewal
and their refusals, and `mcp.setup` for every setup request (closed outcome and reason fields only;
[the contract](../../contracts/mcp/control-plane.md#operational-events)); each guarded MCP operation writes one
`access.decision` (operations `access.begin_verified_request` and `mcp.*`, no capability); a connection termination
writes `auth.session.revoked` with scope `mcp`. No token, subject, client ID, host or connection URL is logged.

## Evidence

| Check | What it proves |
|---|---|
| B-MCP-CATALOG | the exact version-10 catalog delta, ten rolled-back missing-protection detectors, the 00010 up/down/reapply cases; `TestMCPNamesCatalog`/`TestMCPNamesMigration`: the exact version-19 delta, the name-only runtime DML, detectors and the 00019 cases |
| B-MCP-IT | transport, binding, lifecycle, authority, audit, RV-1 local containment and restore invalidation against real PostgreSQL and the official SDK client; discovery (`TestMCPDiscovery` CAT-1..5); facade ordering (O-MCP-7/8; O-MCP-9 rename against revocation in both orders) and footprint (E-MCP-4); IT-MCP-15 (rename) and the rename recovery cases |
| B-UNIT-CORE (`TestToolEligibility`, `TestToolCatalog`, `TestErrorMapping`, `TestAccessBoundary`) | eligibility against `decide` over the oracle world; the production catalog is exactly the five tools; the unknown-tool envelope; the exported discovery surface |
| B-MCP-E2E | the same contract through the TLS edge of the disposable identity environment with the setup page's sessions; configuration refusals of the real core image |
| B-IDENTITY-E2E (mcp-light, mcp-dark) | the setup page: sign-in-first, the not-member refusal, create, list including the historical old-host row (M-LIST), revoke, accessibility of the not-member state; the not-available branch for a session whose membership ends later is unchanged and not exercised by the browser since durable host resolution (IT-MCP-11 proves the transport/tool denial only); named and unnamed creates (M-SETUP), rename, clear, duplicates, the padded non-BMP name, markup as text and refused names (M-RENAME), the manager on My Account with its pages, notices and invalid-page isolation (M-ACCOUNT) and its accessibility (M-A11Y) |
| B-UNIT-WEB, B-BROWSER-ROOT | the web name rule on the shared vector table, strict raw-byte name reading and the form routes (`mcp-connections`, `mcp-forms`); the rendered My Account composition with MCP unavailable on the fixture build (S-ACCOUNT-MCP-UNAVAILABLE) |
| B-MCP-SOURCE, B-MCP-HTTP, B-WEB-BOUNDARY, B-IDENTITY-GUARDS | the import and route boundary, nothing reachable on the unconfigured stack, the web boundary, the test-anchor guard |

**Successor mapping.** The archetype requirements MCP-01..05 described session-wide Organization switching with
per-session catalogs (the generator Skill's isolation fixture, retired with the generator and not part of this distribution). Under the tenant-bound connection model:
MCP-01 (control-plane tools, scope) → IT-MCP-3/5/6/7 and E2E-MCP-1/2; MCP-02 (catalog, arguments cannot change scope)
→ IT-MCP-1/5 and U-MCP-TOOLS; MCP-03 (every execution reauthorizes) → IT-MCP-9/11/12 and E2E-MCP-5; MCP-04 (per-request
server, unauthenticated refused) → IT-MCP-1/2/4 and E2E-MCP-1/3; MCP-05 (production mount with identity) → B-MCP-SOURCE,
B-MCP-E2E and the identity overlay. The historical fixture is not evidence for the tenant-bound contract.

## Discovery and stale context

Discovery depends on each tool's requirement and is computed for every request; nothing is cached
([control-plane.md](../../contracts/mcp/control-plane.md), section "Discovery, stale context and notifications";
[AUTHORIZATION.md](AUTHORIZATION.md#mcp-discovery)).

- Each tool declares one requirement: the exact binding (`logout`), Organization membership
  (`context.switch_group`), membership plus a valid selection (the read tools), or one canonical capability. The five
  production tools are registered on every request's server, and nothing runs in front of their own checks. That keeps
  the explicit clear after a lost Group and the `logout` after a lost membership reachable. A capability-mapped tool
  (none in production; test-only fixtures exercise the path) is registered only when that request's snapshot finds it
  eligible.
- `tools/list` stays refused as a whole when the context no longer holds. A call to a tool that is not eligible gets
  the same unknown-tool answer as a name that never existed, in a valid context and in a refused one. The refusal is
  recorded truthfully: one decision record, plus `access.denied` when the context failed.
- A listed tool still checks its own execution. An earlier list or `context_version` authorizes nothing.
- `tools.listChanged` is `false`, and no list-change notification is delivered: the stateless JSON transport has no
  stream. `context_version` is a selection revision, not a catalog revision. Clients list again after a switch, an
  error or a reconnect.
- **A catalog a client caches is a hint.** It can be stale the moment it is returned. Only the call's own
  authorization counts.
- Cost: production adds nothing to `tools/list`, the five tools, `initialize` or `ping`. A call naming a tool that does
  not exist runs one context snapshot (an error path, rate-limited). A capability-mapped catalog adds at most four
  statements per snapshot, constant in the Group and profile count (E-MCP-4). A future production capability tool
  states its own cost.
- **Explicit Group authority.** `context.list_groups` and `context.switch_group` still list and select every Group
  of the Organization for its owners and admins (selection is metadata), but a selection confers no
  Group management or content: capability-mapped Group/profile management tools are eligible for Organization staff
  without a Group role only while the Organization's Group-management override is on, and the override is read on every
  request (B-MCP-IT `TestMCPDiscovery` CAT-1 lists none with it off and the earlier set with it on). Nothing about the
  setting is cached in the connection or the catalog.
- **Labels.** No production label tool exists: the five production tools, `tools.listChanged`
  and [control-plane.md](../../contracts/mcp/control-plane.md) are unchanged, and labels are reached only through the
  web BFF's internal routes ([WEB-SHELL.md](WEB-SHELL.md#labels-pages)). Eligibility of the three label
  capabilities follows `decide`: `label.organization.manage` for Organization owners/admins (the override is
  irrelevant), `label.group.manage` and `label.group.read` only with an explicit role in the exact Group (never through
  Organization authority or the override; [AUTHORIZATION.md](AUTHORIZATION.md#guarded-label-operations)). Three
  test-only capability-mapped fixture tools (`labels.list_organization`, `labels.list_group`, `labels.read_group`, in
  `catalog_fixture_test.go`) prove finite eligibility, current-context refusal and hidden-equals-unknown direct calls
  (`TestToolEligibility`, `TestMCPDiscovery`) only; they are not a production label operation and are never reported
  as one. A production tenant-data tool remains a separately reviewed public-contract change.

## Limits and claims

- Evidence covers the tested client, user and configuration only. There is no claim about arbitrary MCP clients, no
  revocation SLA and no immediate JWT invalidation: an issued access token stays cryptographically valid until it
  expires, and is refused locally once its connection is terminal.
- Logout and revoke end the local connection, not the provider grant. Containing a compromised client is the operator
  runbook in [OPERATIONS.md](OPERATIONS.md) ("Compromised MCP client containment"); re-enabling a user, including after
  a restore, follows its prior-state, explicit-clearance and other-hold checks ([RECOVERY.md](RECOVERY.md)).
- No Organization-switching tool, no tenant data tools, no consent page in this application, no MCP surface for
  intelligence. Domain tools belong to later stages.
