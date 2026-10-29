> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G014 — Tenant-bound MCP connections work specification

Goal: [G014](../goals/G014-tenant-bound-mcp-connections.md)

Epic: EPIC-005 — MCP control plane

State: Owner-approved definition and per-connection UX amendment; Plan revision released, no execution release

## Rationale and inspected baseline

ADR-005 requires a separately authorized external connection per Organization.
G014 makes that boundary real before G015 exposes dynamic domain tools. A
registered URL chooses a candidate tenant; authentication and core policy decide
access. Global identity never implies global authority.

Inspection: staging `df3d1b0`, 2026-10-15. G012 is finally ACCEPTED at integrated
`5e1e987`; G013 is planning, not accepted or implemented. Migrations currently end
at version 8. Do not reserve the next migration number or assume G013's proposed
schema/tooling has shipped. Reconcile its actual final delivery before planning
and again before execution.

`core/internal/mcp/server.go` constructs a dormant empty official Go SDK server;
`core/cmd/server/main.go` mounts identity/health only. There is no accepted external
MCP transport, token issuer, route registry, Group-list facade or MCP lifecycle
audit catalog. G004's global browser session is not an already tenant-bound MCP
credential. G007 TenantContext is request-scoped and invalid after its scope ends.

Canonical: Architecture §§3/4/21/37/52/53, ADR-003/004/005/006,
[connection-context contract](../../../../contracts/mcp/connection-context.md),
IDENTITY/AUTHORIZATION/DATABASE-ISOLATION/AUDIT/SECURITY and final G013 recovery.
Conflicts escalate; historical generator switching examples do not override ADR-005.

## Boundaries

In scope: one registered shared-origin external HTTP MCP interface and the five
control tools below, their authentication/context/lifecycle guarantees, minimal
necessary storage/audit/schema/configuration and current-guidance reconciliation.
Core retains identity, authorization and audit ownership; web is only an adapter
if the approved authorization flow needs it. No intelligence execution is added.

G015 owns permission-filtered domain catalogs, change notifications, stale-catalog
versions and the broader concurrent selection/execution matrix. G014 must still
secure every tool it actually exposes and isolate live connections. G016 owns
verified hostname aliases, custom domains, DNS/CNAME and TLS lifecycle. Use an
explicit registered route on a configured origin now; do not depend on those
later features. G021's product-feature barrier remains closed.

No product schema/tools, global tenant directory, Organization switch, outgoing
MCP, agent/service accounts, billing/branding/onboarding UI, provider migration,
new broker/authorization-server platform, scheduler/cache framework, CI redesign
or production deployment is implicitly in scope.

## Requirements

### R1 — Pin the external wire and authentication boundary

Use the existing official Go MCP SDK for real HTTP transport; in-memory tests
alone do not qualify. The Plan pins supported protocol revision(s), initial
client(s), methods/metadata, version negotiation, resource URI, credential
issuance/renewal/expiry and authentication-error/discovery behavior. Distinguish
logical application authorization/connection state from SDK objects or optional
protocol session identifiers. A client-supplied session identifier is no proof.

Authenticate every protected request before private tool metadata/data is served.
Permit only explicitly documented public protocol/authorization metadata.
Missing, malformed, expired, revoked, wrong-issuer/purpose/resource credentials
and unknown authority fail closed. Apply Origin/proxy-trust/request-size/time
bounds for the chosen transport; no wildcard browser credential access. An
unavailable verifier/store cannot become anonymous or cached authorization.

Human identity remains WorkOS behind core's accepted provider boundary. Provider
access/ID tokens, global browser cookies and internal service credentials are
not forwarded unchanged as tenant-specific MCP authorization. Bootstrap must
bind the authenticated subject, intended registered Organization, resource and
client/transaction; a callback/exchange selector mismatch cannot mint a differently
bound credential. Preserve optional MFA and unknown-assurance semantics.

The owner's client/authentication choice is open below. If HTTP OAuth is chosen,
pin the applicable MCP authorization/RFC requirements, metadata/discovery and
resource/audience validation; no token passthrough or token in URL/query. Do not
claim generic-client OAuth compatibility from a private/manual bearer fixture.
A new issuer platform/vendor or material identity contract change escalates
before execution rather than being an incidental prerequisite.

### R2 — Registered route and immutable Organization binding

Each configured tenant route resolves exactly one canonical Organization UUID.
Registration is trusted configuration or reviewed existing-store registration,
not public tenant CRUD. Unknown, ambiguous, malformed or unregistered routing
fails without default-tenant fallback. Host/forwarded headers, caller/model
Organization fields, cookies and MCP metadata cannot override the registration.
Pin normalization and the configured public origin; no request-controlled issuer,
resource URL or arbitrary outbound discovery destination.

Establishment authenticates the global subject and checks current Organization
membership. Bind the resulting authorization and logical connection to that UUID.
Every request requires route, resource/purpose credential, connection/session and
subject agreement. A's credential or session at B is rejected even when that same
user belongs to both. Combining one connection's credential with another's state
cannot substitute authorization. Another Organization needs separate authorization.

Names/slugs are routing aliases only. Registry rename/reassignment cannot retarget
an existing connection: preserve its UUID agreement or invalidate/re-authorize it.
Durable canonical domain records use UUIDs; opaque transport IDs/credential hashes
are not alternate Organization identities. Framework table strategies need explicit
justification; no BusinessProfile tenant layer is introduced.

### R3 — Fresh core authority and the finite control surface

Expose only these control-plane tools (besides required protocol methods):

| Tool | Required semantics |
|---|---|
| account.get_current_user | Authenticated subject's explicitly permitted public identity fields; no credential/provider token/session secret or global membership directory |
| account.get_current_context | Current bound Organization UUID and optional selected Group, plus only documented non-authoritative display/version metadata |
| context.list_groups | Only currently selectable Groups of this Organization under accepted core policy, with bounded/paginated output |
| context.switch_group | Select an authorized Group of this Organization, or explicitly clear to Organization-only context |
| logout | End this connection's authorization under the reviewed lifecycle/scope contract |

No Organization/agency switch or global Organization-list tool, generic SQL/CRUD,
product tool, or unguarded hidden direct-call surface. Unknown/retired switching
names are unavailable. Tool schemas/responses/errors and public optional fields
are pinned in the Plan; arbitrary extra fields never expand authority.

Store only connection binding/selection and required lifecycle state, not a live
TenantContext, transaction, role set or authorization result. On each invocation
create a fresh trusted request scope and resolve/guard through core's accepted
identity/policy boundary; end it after the request. Do not copy a prior request's
opaque envelope, use browser selectors as authority or cache grants in Redis.

A bounded policy-backed Group-list primitive may be added. Return only necessary
identity/routing/display fields, filter before disclosure and bound work/results.
Reuse the existing owner/admin/exact-Group selectability policy; no independent
role model, per-Group profile leakage or unbounded per-Group query fan-out.
Clearing Group selection grants nothing beyond current Organization authority.

### R4 — Selection and live-connection isolation

Start a newly authorized connection with explicit Organization-only context
unless the reviewed bootstrap separately validates a requested initial Group.
Selection/clear validates current authority and atomically publishes either the
whole new connection-local selection or no change. Foreign, unknown, malformed
and unselectable Group errors cannot replace/clear the previous selection or
reveal foreign Group details.

Two connections for the same user, to A/B or both to A, keep independent selection
and lifecycle state. Browser navigation cannot change either. Reconnect/resume,
retries and transport loss never select a default Organization, resurrect a
logged-out binding or copy another connection's context. State loss/restart must
invalidate or recover only an exactly verified binding, never infer one.

Pin the minimum selection/request ordering needed to make these actual controls
atomic and isolated, including one bounded in-flight selection case. G015's
broader concurrency/catalog work is deferred; it cannot defer this boundary.
No replicated session/cache infrastructure is required merely for scalability.

### R5 — Current authority, termination and recovery

All protected data, listing, selection and refresh invocations recheck live
user/session and Organization membership. Expiry, revocation, disablement, lost Organization membership or
invalid selected Group fail closed before protected data/effects. Do not silently
clear an invalid Group to widen authority; explicit clear may succeed only after
current Organization authorization. Re-grants do not revive expired/revoked
credentials without the reviewed reauthorization flow.

A narrow self-termination path may prove ownership of the exact existing
route/tenant/connection authorization and end only that binding without requiring
current Organization/selected-Group authority. It returns a generic idempotent
terminal response, discloses no metadata, issues no replacement authorization and
cannot target another connection. A session identifier alone is insufficient;
already expired/revoked state never reactivates. The Plan pins this proof/error
contract. Normal data/context/refresh paths retain all live-authority checks.

Pin connection lifetime, refresh/rotation, transport disconnect versus logout,
restart/resume, idempotency and unknown-commit behavior. **Proposed owner-review
default: logout ends only this connection and its tenant transport authorization**;
no implicit browser, other-connection, all-user or WorkOS sign-out. Explicit global
user/session revocation remains effective wherever the accepted identity contract
requires it. The Plan must resolve any linkage to an existing application session
before claiming this isolation. Revoke-first audit-failure behavior follows the
existing owner policy; do not undo revocation to obtain an audit record.

Any new credential/session/auth-transaction state joins final G013's restored-auth
invalidation, private-backup inventory and retention/replay handoffs. Prove an old
MCP credential cannot resume after recovery and a fresh authorization can work.
Ephemeral state may instead fail closed on restart/restore when explicitly
reviewed; no durable session table is mandated by this Spec.

### R6 — Minimal audit, privacy and repository transition

Define the minimum additive durable audit mapping for successful connection
establishment, authorized Group select/clear and termination, plus established-
identity cross-tenant refusals where safe attribution is possible. The current
29-action catalog contains no MCP connection/context actions; the Plan must
inventory exact proposed actions/reasons, producers, attribution, atomicity and
failure behavior and justify any additive migration. Do not label context
changes as identity sign-in/session-issued events without the actual identity
change. No arbitrary event stream or generalized audit redesign.

Accepted core denial events may be reused only with their actual supported shape;
otherwise review the bounded delta. Pre-authentication/unknown-route failures
remain sanitized operational evidence without invented subject/tenant attribution.
Strict authority-increasing changes need the reviewed audit guarantee; termination
keeps revocation priority and truthful gaps. Retain audit append-only/RLS/retention
protections and bounded independent-failure recording; no network I/O in audit-
holding transactions. Operational logs are not a substitute for required durable
lifecycle evidence.

No credential, refresh value, raw provider token or sensitive membership/profile
payload in logs, URLs, Result or artifacts; authorization codes/state are also
excluded from logs/artifacts. A one-use short-lived code/state may travel only in
the specifically reviewed authorization-protocol redirect, with the selected
PKCE/transaction binding, exact registered redirect validation, no-store and
referrer protections. This is not permission
to put bearer/session credentials in query strings. Protect stored credential
representations and private test inputs through existing procedures.

Replace B-MCP-SOURCE's now-obsolete prohibition on authenticated SDK/transport/
tenancy linkage and its old main.go/import/route pins with precise reviewed
positive/negative exposure checks. Preserve exclusions for migration, fixture/
identity-test, isolated-DDL and standalone validation-tool/schema-conformance code
in runtime binaries; never disable the whole boundary check. The official SDK
itself links its schema package: permit only the exact reviewed SDK-required
schema dependencies, not a blanket ban or an open-ended validator allowance.
The Plan pins that narrow successor runtime-import boundary. Preserve existing identity/health behavior.

Reconcile current reusable MCP guidance/fixtures and accepted customization
provenance that still imply Organization switching. Preserve historical G000
Goal/Spec/Plan/Result and closed evidence; document the successor mapping. No
public export/release or general SaaS-foundation rewrite is authorized here.

### R7 — Bounded validation and attribution

Use one synthetic two-Organization fixture, overlapping identities/different
roles and two Groups in A plus one in B. Include two simultaneous A connections
and a B connection. Exercise actual HTTP requests through the SDK transport and
real accepted identity/policy/RLS paths, not only mocked adapter returns.
Reuse existing test issuer/provider setup; no new live-provider qualification
campaign or production access. If a chosen client flow needs operator testing,
name it in the Plan and finish it before final CI under existing policy.

Use focused parser/transport/unit cases and finite integration negatives first.
Add only checks for the new adapter and required security/recovery boundaries;
reuse inherited proofs with unchanged input provenance. Do not copy all G010 or
G013 cases, create a separate oracle/resource registry or run full stacks for
individual malformed inputs. Existing path selection governs final local/MR/
staging coverage: migrations and shared guards select full under the current
matrix. Pin actual changed paths/selection; no automatic CI redesign or repeated
unchanged controls. Failure/interruption history and output scanning stay intact.

Record exact source/target, approved inputs, SDK/protocol/client versions,
commands, hashes and retained/reused evidence in the Result. Existing Git-first,
batch/reconciliation, manual-feedback ordering, MR/Duo disposition and staging
status rules apply. Owned resources and raw local evidence are temporary; close
with scoped cleanup/compact receipt, preserving unknown/shared resources.

## Acceptance criteria and verification

| ID | Required outcome | Objective verification |
|---|---|---|
| AC1 | External MCP transport and the five-tool surface work under the reviewed protocol/authentication contract. | Real HTTP SDK client negotiation/discovery and authenticated calls; exact allowed tool set, unknown/retired switching names refused. Public metadata versus private tools pinned; in-memory-only evidence insufficient. |
| AC2 | Registered routing selects one canonical tenant without granting access. | Known A/B route→UUID mapping, unknown/malformed/ambiguous routes and forged Host/forwarded/selectors fail without fallback; route rename/reassignment cannot retarget existing authorization. |
| AC3 | Establishment authenticates and authorizes the intended subject/Organization/resource. | Valid member succeeds; absent/malformed/expired/wrong-purpose credentials and authenticated nonmember fail. Bootstrap subject/resource/Organization mismatch issues no usable authorization. |
| AC4 | Endpoint, tenant credential and connection binding must agree. | Direct A-access-token/session replay at B, mixed connection credential/state and same-global-user mismatches fail; neither connection changes. Under the dated managed-grant amendment, a separately issued valid B-audience token is evaluated under B's exact binding/live authority. Old/revoked/state-lost records cannot be inferred from caller selectors. |
| AC5 | Current identity/context output is bounded and non-authoritative. | Assert documented response fields, correct subject/immutable Organization/optional Group and no raw secrets/provider tokens/global membership enumeration. Use fresh real request scopes; ended/prior-request envelope reuse refused. |
| AC6 | Group listing exposes only currently selectable Groups. | Owner/admin and exact-Group-member fixtures return expected A-only sets; B/unknown/unselectable Groups absent. Bound/pagination controls and no per-Group authorization query fan-out documented/tested on the finite fixture. |
| AC7 | Group select/clear is atomic and cannot broaden authority. | Authorized select and explicit clear succeed; unknown/foreign/unselectable/malformed selection preserves prior state and nondisclosing errors; cleared context gains no new grants. One bounded selection/request interleaving follows the pinned ordering. |
| AC8 | Stored selection and lifecycle are connection-local. | Two A connections plus B for one user: selection, clearing and browser navigation affect only the targeted connection; reconnect/restart either verifies exact binding or rejects, never copies/falls back. Distinct provider refresh-family authority is not promised under the dated managed-grant amendment. |
| AC9 | Live revocation, local termination and compromised-client containment are enforced. | Expiry, disablement, session/Organization revocation and selected-Group loss deny subsequent protected control calls. Reviewed connection-local logout/idempotency, same-binding termination after membership loss and audit-fault path proven; unrelated state preserved for normal logout. Separately authorized user/client compromise containment satisfies RV-1 in the dated amendment, including all old refresh families and fresh replacements. |
| AC10 | Audit/recovery/privacy contracts cover the new state. | Exact additive event/grant/migration inventory and real-runtime refusal evidence; required lifecycle successes/refusals attributable, audit faults truthful without false authority success. Old MCP credential rejected after accepted recovery invalidation/state loss, fresh authorization succeeds; no raw sensitive output. |
| AC11 | Current repository/reusable guidance reflects authenticated tenant-bound MCP. | B-MCP-SOURCE successor checks allow only approved authenticated wiring while forbidden runtime imports/bypasses still fail; identity/health regressions pass. Current Organization-switch guidance/fixtures reconciled with provenance, historical G000 artifacts unchanged. |
| AC12 | Delivery meets bounded evidence, review and closeout gates. | Source-bound AC1–AC11 map, focused and actual path-selected CI, read-only security review, B7/checkpoint4 ledger, failed/reused evidence, scoped cleanup receipt, Codex actual-diff/final ACCEPT and G015 handoff. G021 remains closed. |

## Expected artifacts and completion gates

Goal/Spec and later Claude Plan/Result; authenticated MCP adapter and finite
control tools; justified core policy/list/storage/audit changes; additive migration
only if the approved design requires it; minimum contract/operating/security/
recovery/current-guidance updates; focused tests and existing-harness wiring.
Claude selects file layout and exact methods/schema after the owner hold releases.
No detailed Plan, new snapshot system or implementation is part of this draft.

Completion requires all twelve ACs, accepted source/security deviations, required
staging disposition, scoped cleanup and final Codex ACCEPT. Integrating files or
passing a protocol handshake alone is not completion. G015 receives immutable
connection/session identity, context transition/lifecycle, revocation/audit and
protocol contracts; it adds dynamic discovery without replacing those controls.

## Human input and Plan decisions before execution

**Owner review before Claude planning remains an explicit gate.** Resolve or
explicitly delegate the following choices at that review; this draft selects no
new vendor or authorization-server deployment:

1. Initial real client compatibility target and authorization experience: standard
   remote OAuth-capable clients versus deliberately limited private/manual
   credentials. A narrow mode must be labelled and cannot claim general support.
2. Authorization issuer/bootstrap responsibility, resource/audience, client
   registration/consent and identity-session linkage. WorkOS AuthKit browser login
   alone does not settle these. A new issuer platform/security architecture needs
   owner/ADR review or a separately bounded goal, not silent expansion.
3. Confirm the proposed connection-local logout default. Pin lifetime/renewal and
   restart/resume semantics in the Plan; no new Redis/durable store by assumption.

During planning resolve exact registered shared-origin route/normalization,
protocol versions, payload/errors, minimal audit/migration delta, final G013
invalidation handoff, actual CI selection and any operator-test prerequisites.
Flag concrete database/query/index/fan-out/locking costs in PG-006 before approval;
propose bounded representative observations where necessary, not a benchmark
campaign or speculative performance SLO. Native/provider prerequisite uncertainty
is not permission to implement a weaker authentication mechanism.

No preapproved deviation weakens tenant binding, current authority, privacy,
audit/recovery truth or required evidence. Routine file/fixture details stay in
the Plan. Owner approval of the definition alone does not accept implementation.

## Primary protocol references (planning inputs, not a compatibility claim)

- [Pinned Go SDK v1.8.0 protocol documentation](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/docs/protocol.md):
  supports different protocol lifecycle models. Logical application binding must
  not depend on assuming every negotiated version supplies a transport session ID.
- [MCP authorization revision 2026-07-28](https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization):
  pin the requirements of the chosen supported revision during planning.
- [MCP security guidance](https://modelcontextprotocol.io/docs/2026-07-28/tutorials/security/security_best_practices):
  review resource audience, token passthrough and session-hijack boundaries.

References were inspected read-only while drafting; no SDK install, probe,
protocol-version change or credential/provider action occurred.

## Owner approval / planning-phase amendment — 2026-10-15

Owner approved this definition after the Group-clear/Organization-permissions
clarification. The initial human-review-before-planning gate is satisfied.
The five tools, twelve ACs, tenancy/permission boundaries and connection-local
logout default are unchanged. No Organization-level authority is created by
clearing selection; existing Organization role permissions remain in force.

Claude may now prepare a provisional Plan proposing the initial client/auth
experience and issuer/bootstrap/wire contract. This supersedes the requirement
to settle those choices before drafting a Plan; it does not silently settle them.
New platforms/vendors or material unresolved authentication decisions still need
owner/ADR disposition before Plan approval/execution. Reconcile final G013's
actual recovery interfaces; final predecessor ACCEPT, reviewed Plan and current
input/budget/resource release still precede implementation.

No G014 implementation, credential provisioning, provider changes, tool probes
or new test/validation campaign is authorized by definition approval. Planning
only may proceed alongside G013 without using its shared validation environment.

## Owner amendment — tenant-hosted per-connection UX, 2026-10-15

Authority: after confirming the URL should be acme.app.foundation.test/mcp/c/{id},
per-user connection tracking, within-connection Group switching and stable URL
through ordinary reauthentication, the owner directed **“Ok update with those
details.”** This approves the UX/contract direction for the WorkOS per-connection
proposal and authorizes the existing planner's bounded Plan revision. It is not
Plan ACCEPT, provider-qualification approval, implementation release or G013 F4
risk acceptance. Previous Plan r1/review/probe history remains unchanged.

The canonical [connection-contract amendment](../../../../contracts/mcp/connection-context.md#owner-amendment--per-connection-urls-and-lifecycle-2026-10-15)
and Architecture §3 define the amended requirements. This dated amendment takes
precedence over the earlier shared-origin-only boundary and undecided-UX text:

- Target URL: https://{organization_slug}.${PLATFORM_BASE_DOMAIN}/mcp/c/{UUID}.
  Registered host, resource, canonical user/client and immutable Organization
  binding must agree. URLs remain selectors, never authorization credentials.
  A different Host cannot reuse an otherwise valid connection credential.
- Store connection binding, optional Group and lifecycle state; do not persist
  trusted request contexts or cached roles. One per user/Org/AI client is a UX
  default, not an enforced uniqueness limit. Both agency staff and client users
  are eligible according to existing membership/roles, with no additional grants.
- One URL means one logical connection. It may switch/clear authorized Groups
  within its Org; it is not per Group. URL reuse shares selection and logout;
  independently configured clients/devices should create separate URLs.
- Normal credential refresh/reauthentication for a still-valid connection keeps
  the URL and exact user/client/Org binding. Credential expiry alone is not
  permanent connection revocation. Logout/revocation, recovery invalidation,
  missing state or expired pending setup cannot be reversed by presenting tokens;
  replacement setup uses a new UUID/URL. Plan pins connection versus credential
  lifetimes and atomic activation, including first-token client binding.
- R1 clarification: dedicated WorkOS Connect credentials may authenticate the
  subject only after issuer/namespace and exact resource qualification. Ordinary
  login/ID tokens, browser cookies and internal credentials remain unacceptable
  as MCP transport authorization. No email-based identity linking.
- R5 clarification: provider refresh issuance is distinct from Beginify
  authorization. WorkOS may refresh after local revocation; Beginify must deny
  protected access using live local state. Local context/authority changes retain
  all current checks; do not claim local checks execute at WorkOS's endpoint.

Before accepting the revised Plan, qualify two concrete tenant-hosted connection
audiences, stability on refresh, negative old-family refresh retargeting to a new
resource, and exact G004 subject mapping. Path-wildcard support does not establish
wildcard-host support. Identify reviewed provider registration limits/configuration
and explicitly configured host readiness. No new probe/provider mutation is
released here. G016's provisioning/DNS/TLS/custom-domain work stays deferred;
escalate an actual prerequisite rather than reverting to a shared host or pulling
that work forward silently. Minimal authenticated connection setup is in scope;
general onboarding, tenant CRUD and product screens remain out of scope.

Map the added evidence into AC2–AC4/AC6–AC10 and AC12: host/Org/audience mismatch;
agency-staff and exact client-Group roles; same-URL shared state and distinct-URL
isolation; ordinary renewal/reauthentication with unchanged URL; revoked,
expired-setup and restored-state refusal. Existing twelve ACs/B7/checkpoint4,
security review, G013 final ACCEPT and final input/resource/implementation gates
remain. No new generic validation framework or duplicated predecessor suite.

## Clarification — access changes during a live connection, 2026-10-15

This makes R3–R5 and AC6–AC10 explicit, without changing the permission model
or adding a suite. Authorization is not snapshotted for the connection lifetime.
Role, membership, user/session status and selected-Group authority are resolved
through the existing live core boundary on protected invocations. No reconnect
or token refresh is required for a committed access change to take effect.

- A role downgrade removes the affected capability, not unrelated permissions.
  Lost selected-Group access refuses Group-scoped/protected requests; do not
  silently select another Group or widen to Organization-only context. Explicit
  authorized select/clear and same-binding self-termination retain R5's exceptions.
- Lost Organization membership, user disablement or connection/session revocation
  denies protected access regardless of a valid provider token or cached catalog.
- Reuse accepted AUTHORIZATION guarded-read/write concurrency semantics. A guarded
  read uses its authorized snapshot; do not promise cancellation of an already
  authorized read or reversal of committed effects. Mutations serialize relied-on
  authority with revocation at the existing transaction guard/linearization
  boundary, not merely a middleware check before an unguarded write. The Plan
  identifies the boundary for each of G014's state-changing control operations.

G014 demonstrates these cases on its five control tools using existing fixtures:
change authority after connection establishment, then call again without a new
connection/token; refuse stale selected-Group/cross-tenant access; preserve the
explicit clear/logout exceptions and unrelated connection state. Reuse the
existing bounded selection/request interleaving and transaction evidence; add
only tests for actual uncovered behavior, with AC9 live-revocation and AC10 audit
coverage. No copied predecessor concurrency matrix or new invalidation service.

G015 owns the broader dynamic domain catalog: recompute/invalidate tools/list
on relevant Group/role/membership/link/policy changes, with the supported protocol's
notification/re-list behavior pinned in its future Plan. Cached client UI cannot
confer authority; hidden/direct calls remain independently checked. No claim of
instantaneous client repaint or broader G015 implementation is made by G014.
Existing Plan revision should map this clarification; no separate replan or new
Goal required. Existing review/input/dependency/budget/resource gates remain.

## Owner lifetime and qualification disposition — 2026-10-15

Owner accepted the recommendations with **30 minutes for setup**. Pending
connection setup therefore defaults to 30 minutes; active connections expire
at 30 days idle or 180 days absolute. Terminal expiry requires a new connection
URL; ordinary valid-connection renewal keeps it. These defaults supersede the
Plan r3's proposed 15-minute pending window. Configurable ranges and enforcement
remain Plan-reviewed; no authority/lifecycle bypass is permitted.

The dated owner disposition (not included)
authorizes exactly the disposable WorkOS qualification and bounded PG-006
observations described there. Use a 30-minute idle probe setup cap and at most
30 minutes after explicit operator Start. No existing provider settings/records
are changed. Actual qualification, Plan/ADR/input/resource acceptance and final
G013 ACCEPT remain implementation gates; this is not G014 implementation release.

## Owner amendment — managed WorkOS grants and compromise containment, 2026-10-15

Authority: owner answered **"Ok"** to the recommendation to retain WorkOS login
and MCP issuance, accepting broader same-user/client credential reach and requiring
a supported compromised-client revocation procedure before implementation. This
selects assessment option B, not the proposed dedicated issuer. It supersedes
the earlier per-connection independent OAuth authorization/refresh-family and
negative-retarget acceptance requirements. QP-3 remains historically failed;
never relabel the observed retarget as a passing old test. All unchanged
tenant/Group, liveness, audit/recovery and scope guarantees remain.

### Required managed-grant behavior (R1–R5 / AC3–AC4 / AC8–AC10)

- WorkOS authenticates the global identity and issues dedicated Connect tokens.
  Existing WorkOS SSO may skip password/MFA or another consent screen. A new local
  connection still requires authenticated browser bootstrap and current tenant
  membership; a token cannot create a connection or synthesize a tenant grant.
- Each protected call requires exact issuer/resource audience and current bound
  user/client, registered host, immutable Organization, live connection/session,
  membership and selected-Group authority. Keep namespace mapping without email
  linking, wrong-purpose token refusal and accepted request transaction guards.
- Direct replay of A's access token/session at B fails even for one global user.
  A separately issued token with exact B audience may authenticate to B when all
  B bindings and current authority agree, even if the provider refreshed it using
  a credential that also serves another connection. Refresh families are not
  treated as independent per-connection security compartments.
- Stored Group/lifecycle state remains separate: a call targeting A cannot
  mutate B's selection or terminate B. A correctly authenticated, separately
  targeted B call may act under B's permissions. Do not claim the presenting
  provider family was uniquely or freshly authorized for B.
- Normal connection logout ends only that local connection. It does not revoke
  the provider grant or claim to contain a leaked refresh value. Unknown,
  revoked, expired-setup, missing or recovery-invalidated records remain terminal.
  Replacement requires authenticated new browser setup and a new UUID; it need
  not involve fresh password entry or a distinct provider grant.
- The owner accepts broader credential exposure: one refresh value may obtain
  tokens for other resources already authorized by that user/client. Actual QP
  observed one same-host case. Cross-host, unconsented and future replacement
  reach remain unqualified risks, not established provider guarantees.

### RV-1 — compromised-client containment release gate (AC9 / AC10 / AC12)

**Provider gate before implementation release:** identify and qualify a supported
operation that revokes all pre-existing refresh families for the affected user
and OAuth client. Use provider-only authorization flows and audience-only test
resources; no core database, future connection UI or implementation is needed.
Pin the real authorized operator/management operation, scope, attribution,
failure/uncertainty semantics, propagation bound and issued-token limits in the
Plan. An undocumented consent-delete effect is not proof of containment.

Minimum bounded provider qualification: establish two families for one user/client;
run supported grant revocation; show their pre-operation refresh values cannot
regain usable tokens by ordinary refresh or resource retargeting. Perform a fresh
provider authorization for a new test resource and show it works while those old
values remain unusable. This is provider authorization, not future application
replacement setup. Do not claim instantaneous invalidation beyond the evidence.

**Local implementation acceptance (AC9 / AC10), after execution release:** prove
that the separately authorized compromise procedure invalidates all affected
local pending/active connections; pre-issued access tokens fail against those
records; fresh authenticated replacement setup works only after containment is
established. Provider failure or uncertain completion leaves containment unresolved
and replacement activation held under the reviewed procedure. Bind these local
checks to the accepted provider facts; no application bootstrap or implementation
is demanded before Plan/execution approval. Normal connection logout stays local;
a supported broader emergency operation needs its actual side effects disclosed.

The Plan may reuse existing broader local revocation/disable mechanisms and a
bounded operator runbook; do not invent a generalized revocation registry,
scheduler, extra MCP tool or identity platform for this gate. Escalate an actual
missing supported operation instead of weakening the requirement. If it cannot
be established, return to the owner before selecting another issuer/model.

This requirement does **not** authorize a new provider probe or mutation. First
resolve documented support using existing sources/records. Any necessary live
qualification needs a separately scoped owner authorization, exact owned records,
operator readiness, deadlines and cleanup. Combine related subject/refresh and
revocation observations in one bounded run where possible; no repeated campaign.

### Reconciliation and completion

Retain twelve ACs; AC4/AC8 now distinguish local state from credential-family
independence, and AC9 includes RV-1. Preserve the current Plan r4 and prior failed
evidence. The existing planner prepares only the bounded revision required by
this decision and identifies exact remaining provider/owner gates. Qualify exact
issuer/audience, subject mapping and ordinary refresh under the selected model;
no silent acceptance of generic environment-client or wildcard audiences.

OD-1 stays 30-minute pending setup, 30-day idle and 180-day absolute connection
life. Those limits govern local connection state, not every resource a provider
refresh grant might serve. B7/checkpoint4, security review, final G013 ACCEPT,
Plan/input/budget/resource gates and the G021 product barrier remain. No dedicated
issuer, new dependency, application code or implementation is released here.
