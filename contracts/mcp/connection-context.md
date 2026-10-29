# MCP connection context — behavioral contract

Status: owner-approved design contract, 2026-10-11 (architecture decision ADR-005; its record is not part of this distribution, and this contract states the requirements in full). Owning implementation Goals: G014–G016. This defines required behavior, not an implemented endpoint, MCP wire schema or credential format. Core owns policy; web, clients and intelligence cannot supply authority.

## Organization binding

1. A tenant-specific connection URL resolves through server-registered or verified routing to exactly one canonical Organization UUID. Unknown/unverified routes fail without selecting a tenant. Production URL form and normalization/proxy trust are pinned during planning; `https://acme.platform.example/mcp` is illustrative, not a deployed contract. A registered Organization route on a configured shared origin supports G014 before G016's hostname lifecycle.
2. Establish connection authorization only after authenticating the global user and validating current membership of the resolved Organization. A URL, slug, Organization ID, session ID, provider claim or model argument alone grants nothing. No default-tenant fallback.
3. Bind the connection's authorization and MCP session to that UUID for their lifetime. Endpoint, tenant-specific transport credential and session must agree. Reject A's session/credential presented at B, including when the user belongs to both. Authenticate globally only to establish a separate authorized connection; never retarget an existing one.
4. A route/slug/domain rename cannot silently reassign an existing session to another UUID. Requests through an alias still require agreement with the established binding. The G014/G016 Plans pin route reassignment and session invalidation behavior without permitting cross-tenant reuse.
5. Do not expose Organization/agency-switching tools. Ordinary tool parameters, headers, cookies or resource IDs cannot alter the bound Organization. Cross-tenant targets are rejected through the accepted non-disclosing error contract.

## Optional Group context

6. Organization-only context is valid. `context.list_groups` shows only authorized Groups of the bound Organization. `context.switch_group` selects an authorized Group or explicitly clears selection to Organization-only context; both retain the Organization binding. Failed selection leaves the previous context unchanged.
7. Group switching is foundation control-plane behavior. Its application UI/tool use may be domain-specific, but Group authority is evaluated by core's shared policy. A selected Group restricts Group-derived access according to that policy; clearing selection does not synthesize broader grants.
8. Group selection and catalog versions are connection-local. Two connections for one user, whether to different Organizations or the same Organization, cannot modify one another's context through Group selection. Browser tenant/Group navigation likewise cannot silently retarget a live MCP connection.
9. The authenticated tenant-bound control surface is current user, current context, authorized Group listing/selection and logout. Global Organization discovery for configuring another connection is a separately authorized application concern, outside this connection's catalog. The G014 Plan pins exact method/payload/error and logout scope; no new global logout side effects are implied here.

## Discovery, lifecycle and execution

10. Resolve current capabilities for this connection; refresh/invalidate discovery on Group selection and relevant membership/role/link/policy changes. Hidden direct calls fail like unknown tools. Previously visible tools do not retain authority after revocation or catalog staleness.
11. Every execution independently validates live authenticated user/session, bound Organization, current Organization membership, optional Group, exact resource ownership, capability and applicable approval policy. Transport reconnect/resumption/replay must preserve the immutable binding and validate current authority. Concurrent context changes and tool execution need a reviewed linearization/version contract; a stale catalog or client context assertion is insufficient.
12. Preserve the accepted expiry/revocation/disable and privacy/audit contracts. Record connection establishment, authorized Group changes and denied cross-tenant attempts through the reviewed bounded attribution rules; no raw credentials or sensitive membership/profile payloads. No inferred MFA assurance or new mandatory enrollment.

## Required verification in the owning Goals

| Scenario | Required outcome |
|---|---|
| User belongs to A and B, separate connections with different roles | Each catalog and execution uses only its own Organization and current role facts |
| A URL without authentication or without A membership | Connection authorization denied; no tenant data/tools disclosed |
| Unknown host/route or untrusted forwarded host | No Organization resolution/authorization fallback |
| A credential/session replayed at B, including same authenticated user | Rejected; neither connection's binding changes |
| Organization-switch tool/direct domain argument attempts | No switching surface; cross-tenant call rejected |
| Authorized Group select/clear within A | Context changes only within A; exact catalog/permissions refreshed |
| Foreign/unauthorized/unknown Group selector | Rejected without changing prior context or disclosing foreign data |
| Two simultaneous connections and browser navigation | Group/context changes stay isolated to the targeted connection |
| Membership/role/link/policy revocation, expiry or disable after discovery | Subsequent invocation denied even through a stale catalog |
| Reconnect/resumption and concurrent switch/execution | Immutable Organization and reviewed ordering/version guarantees hold |
| Slug/domain alias change or route reassignment | UUID identity preserved; mismatched old session cannot reach a new tenant |

Use real accepted policy/isolation boundaries where required. G014 owns connection and Group-control evidence; G015 owns discovery/stale/concurrency evidence; G016 owns host/domain-resolution cases. Full hostname lifecycle is not a prerequisite invented for G014. Acceptance requires actual implementation and attributable evidence under each Goal, not this document's presence. No schema, protocol library, product workflow or implementation is authorized here.

## Owner amendment — per-connection URLs and lifecycle, 2026-10-15

This supplements the numbered contract and supersedes the earlier illustrative
shared-origin URL as the intended user-facing layout. Use
`https://{organization_slug}.${PLATFORM_BASE_DOMAIN}/mcp/c/{connection_uuid}`
Host registration and the connection record must agree
on the canonical Organization UUID; unknown/mismatched hosts fail closed. G014
reviews configured-host feasibility; G016 still owns provisioning/DNS/TLS/custom
domains. No arbitrary host-derived issuer/audience or early G016 implementation.

The server records the connection UUID, canonical user/Organization, bound client,
optional Group and lifecycle state. One logical connection is normally enough
for one user/Organization/AI client; multiple connections remain supported. Agency
staff and client users use the same connection mechanism with their current
Organization/Group permissions; neither the URL nor MCP enrollment grants roles.
Group switching stays inside that connection. A second URL creates independent
selection/logout state; copying the same URL reuses one logical connection.

Routine token expiry/refresh or user reauthentication for a still-valid connection
retains its URL and user/client/Organization binding. This does not permit
reauthentication to revive a revoked, invalidated, missing or expired-setup record.
Explicit connection logout/revocation is terminal; replacement has a fresh UUID
and URL. Expired unactivated setup also requires a fresh connection. Recovery
invalidation covers pending and active connection state.

WorkOS per-connection issuance is the selected proposal for Plan revision,
conditional on qualification and acceptance. Provider refresh may continue after
local revocation, but every protected MCP call checks live local authority and
refuses access. Unknown records are never created from tokens. Qualify exact
concrete tenant-hosted audiences, subject mapping and an old-family refresh
attempt targeting a new resource; an old family must not acquire that audience.
The URL is not a bearer credential. No new provider probe/configuration or
implementation is authorized by this documentation amendment.

Verification extends existing G014 ACs, not a new suite: staff/client permission
fixtures, same-URL shared state versus distinct-URL isolation, allowed within-Org
Group switching, stable URL through ordinary renewal/reauthentication, terminal
revocation/expired setup/recovery refusal, and wrong-host/audience rejection.

## Live access changes — clarification, 2026-10-15

An active connection retains identity and tenant binding, not a lifetime grant.
Subsequent protected calls use current authority without requiring reconnection
or token refresh. A downgrade removes affected capabilities; lost selected-Group
access refuses protected operations without automatic selection/clear. Explicit
select/clear and exact-binding termination use their reviewed current-authority
rules. Organization membership loss, disablement and revocation deny protected
access despite still-valid provider credentials or a cached tools/list.

Use accepted core transaction semantics: guarded reads may finish on their
already-authorized snapshot; guarded mutations serialize relied-on authority
with revocation. Do not promise instantaneous cancellation or rollback of committed
work, and do not authorize a write solely from an earlier adapter check. G014
pins this boundary for its control-state mutations; G015 handles broader dynamic
catalog invalidation/notifications and concurrency. Client catalog refresh timing
does not weaken server-side execution authorization. Reuse existing controls
and cover actual gaps, rather than creating another general concurrency suite.

## Owner amendment — managed OAuth grants and local connection authority, 2026-10-15

The owner selected option B: WorkOS remains the browser identity and MCP token
issuer. This amendment supersedes the earlier demand that an old refresh family
must never acquire another connection audience. QP-3 remains a recorded failed
experiment, not a passing test. No independently resource-bound refresh family
or unique provider authorization instance is promised for each connection.

The accepted exposure is that one user's client refresh credential may obtain
tokens for that user's other already-authorized resources. Broader reach to
unconsented resources, other hosts or future replacement connections is untested
and must be described as potential exposure. A URL is still a selector, not a
credential; knowledge of it never authorizes a request.

Numbered rules 3/8 and their verification distinguish local connections from
provider grants: direct reuse of A's access token/session at B is refused. A
separately issued token with exact B audience is evaluated against B's own stored
user/client/Organization, liveness and current permissions. A call targeting A
cannot modify B's Group/lifecycle state; a legitimately authenticated call
targeting B may do so under B's authority even if the provider refresh credential
also serves another connection. Do not claim credential-level independence.

Connection logout revokes that local binding only. Unknown/terminal connections
remain refused; tokens cannot resurrect them. A replacement requires new
authenticated browser setup and a new UUID, without a guarantee of distinct
WorkOS consent or fresh password entry. Compromised credentials require a separate
explicitly authorized user/client-wide containment procedure, including supported
provider grant/family revocation and local invalidation. Qualify the provider's
effect on all old families after fresh provider authorization before G014
implementation, without app bootstrap. Prove local invalidation/replacement holds
and old-access refusal during approved implementation before acceptance;
revoking local rows alone is insufficient. No new global-logout tool is implied.

Retain exact issuer/audience/subject/client checks and all live tenant/Group,
recovery and resource-ownership guarantees. This is a deliberate change to grant
compartmentalization, not a waiver of tenant authorization. WorkOS revocation
qualification, subject mapping and revised Spec/ADR/Plan/input gates remain;
no new provider experiment, settings change or application implementation is
authorized by recording this decision.


## Owner-approved G016 alias lifecycle target — 2026-10-16

Architecture decision ADR-007 (its record is not part of this distribution)
approves retaining historical MCP records without automatic revocation on alias
removal/slug change. An exact stored host must currently route to the bound
Organization; removed/foreign aliases refuse, never retarget resources or Group
state. Historical rows remain listable/revocable through a current authorized host;
new aliases require separate authorized connections. Terminal/expiry rules remain.
This selects G016's implementation contract; it does not claim rollout or authorize
live domain/provider changes.

## G016 implemented alias behaviour — 2026-10-16 (branch, pending acceptance)

- Host lookup is the durable resolver (Organization slugs and verified, routing-ready custom domains); the static host
  registry (`MCP_TENANT_HOSTS_FILE`) is retired and refused. Unknown or non-routing hosts are 404 before any store
  access; a store failure is 503, never a clean unknown host.
- A request is admitted only when its host resolves to the bound Organization and equals the stored `host[:port]`:
  a removed or renamed alias is 404, a host reassigned to another Organization 401, another current alias of the same
  Organization `resource_mismatch`. No resource, audience or Group state is rewritten; nothing is revoked
  automatically.
- The setup listing returns every connection of the user in the current host's Organization, whatever host it was
  created on, with that host shown; revoke is by connection ID through any current host of the Organization; create
  keeps the exact-origin check ([MCP.md](../../docs/foundation/MCP.md) "Hosts and aliases").
