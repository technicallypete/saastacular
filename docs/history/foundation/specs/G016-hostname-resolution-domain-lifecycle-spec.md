> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G016 — Hostname resolution and domain lifecycle work specification

Goal: [G016](../goals/G016-hostname-resolution-domain-lifecycle.md)

Epic: EPIC-006 — Organization domains and branding

State: Definition ready under EPIC-006 mandate; provisional planning at G015 implementation MR; execution gated

## Purpose, rationale and authority

Make Organization-specific platform and custom hostnames usable as verified
routing aliases, while canonical UUID ownership and current authorization remain
independent of DNS, TLS, slugs and provider claims. G014 supplies configured-host
MCP connections; G015 supplies dynamic discovery. G016 owns their subsequent
hostname lifecycle and scalable browser-authentication composition.

The owner's 2026-10-15 instruction authorizes this draft only. It does not select
a production DNS/TLS vendor, approve a cross-origin session mechanism, release
planning/implementation or authorize the following Goal. G021 remains closed.

Definition baseline: integration checkout `staging`, inspected at
`75b546bb725c23c1397a413273fc588852e02a5b`. G014 Plan r8 is accepted and its
implementation is in progress; G015 is defined, not accepted implementation.
Reassess accepted integrated predecessors before G016 execution; do not reserve
a migration number or assume proposed predecessor interfaces are shipped.

Canonical inputs: AGENTS/CLAUDE, Architecture §§3/50/51/52/60/63, SECURITY,
DECISIONS, ADR-003/004/005/006, ROADMAP/PROGRESS, G014/G015 artifacts and eventual
Results, the [connection contract](../../../../contracts/mcp/connection-context.md),
[redirect-scaling addendum](G004-tenant-redirect-scaling-addendum.md), and foundation
IDENTITY/AUTHORIZATION/DATABASE-ISOLATION/AUDIT/RECOVERY plus PERFORMANCE.

## Scope, dependencies and non-goals

In scope: core-owned domain registration/lookup and lifecycle, bounded DNS
ownership/configuration checks, TLS adapter state, minimal authorized management
operations needed to exercise that lifecycle, scalable browser callback/sign-out
configuration, trusted web/core routing, MCP alias safety, configuration migration
and operating documentation. Durable domain records/migrations belong to G016;
their exact schema and next migration follow the actual accepted repository.

Execution depends on final G002/G015 ACCEPTs and their transitive identity,
tenancy, authorization, audit/security/recovery contracts. G014/G015 pending work
does not establish a passing G016 prerequisite. Owner decisions below and Plan,
binding, budget/resource gates must also close.

Exclude tenant branding/content (G017), public acquisition/marketing deployment,
onboarding, global tenant directory, product features, new external MCP management
tools, broad settings UI, enterprise identity administration, a new token issuer,
production cloud rollout, wildcard DNS hosting, a generalized job/certificate
controller, CI redesign and speculative indexes/benchmark campaigns. A live
provider/environment change requires explicit bounded authority, even if needed
for later qualification.

## Requirements and architectural/security invariants

### R1 — Canonical hostname identity and trusted request resolution

Domain records use canonical UUIDs and reference exactly one canonical Organization
UUID. Normalized hostnames are alternate identifiers with database-enforced global
uniqueness; duplicate/concurrent claims cannot select different Organizations.
Group and BusinessProfile are not additional domain tenant tiers. Names never
grant membership, and successful global authentication never grants tenant access.

Pin one hostname normalization contract in the Plan: case, trailing dot, IDNA or
explicit Unicode rejection, allowed port handling, label/suffix boundaries and
malformed/ambiguous authority rejection. Do not derive authority from URL userinfo,
path, arbitrary forwarding headers or a suffix substring. Match a registered
canonical hostname, not a guessed default tenant. Exact authentication origins
and MCP resource strings keep their stricter contracts; normalization must not
silently broaden an existing audience or browser transaction.

Only a configured trusted edge/proxy path may supply authoritative external scheme/
host metadata; direct client spoofing fails. The Plan pins that boundary and any
replacement of G004's ignored-forwarding-header behavior. Unknown, malformed,
unverified, suspended and removed hosts cannot select a tenant or receive its
private data. Edge TLS/HTTP routing and core binding must agree; alternate direct
core/web access cannot bypass these checks. Cache behavior, if any, has explicit
invalidation and cannot extend removed-domain authority.

### R2 — Platform suffix, reserved hosts and configuration transition

Platform tenant slugs resolve under the deployment-configured application suffix.
Reserve apex, `www`, the configured CNAME target, authentication/protocol and
infrastructure labels; exact reserved-set and slug-change behavior are pinned in
the Plan. Reject suffix confusion, sibling lookalikes and collisions with custom
claims. Editing a slug never changes an Organization UUID or rebinds sessions.

The intended Beginify deployment defaults are `PLATFORM_BASE_DOMAIN=app.foundation.test`,
`CUSTOM_DOMAIN_CNAME_TARGET=custom.app.foundation.test` and a configured marketing
destination `https://foundation.test`. Reconcile Architecture §51's historical `.ai`
examples and §50/generated-baseline defaults explicitly. These are configuration
values, not code constants. Change current templates/manifests/provenance together
under the accepted procedure; preserve historical snapshots and the environment
consistency check. Document rename/re-domain behavior and the ingress prerequisite
before telling an operator to point customer DNS at the CNAME target.

Apex/`www` public-entry routes redirect to configured marketing without selecting
an Organization or global workspace picker. Explicit callback, sign-out return,
MCP metadata/transport and other registered protocol routes retain their selected
reviewed routing; they must not become marketing redirects. G016 adds no marketing
site or agency-acquisition flow.

### R3 — Ownership, authorization and DNS lifecycle

Core owns domain-management authorization, transitions and audit attribution.
Pin a finite operation/role/capability matrix before Plan ACCEPT; unknown authority
denies, and Group/profile ownership or global identity alone cannot administer
Organization domains. Reuse scoped guards/RLS/transactions; no browser/provider
claim supplies Organization ownership. Lifecycle operations and any internal
administrative route are closed and non-public unless explicitly reviewed.

A custom hostname starts non-routing. Issue a fresh, Organization/domain-bound
ownership challenge and check the configured CNAME/ingress association. A shared
CNAME target alone is not sufficient to let another tenant claim a customer host.
Pin proof format/location, expiry, retry limits and verification freshness. Stale,
foreign, reused or wrong-host proof cannot activate a claim; concurrent claims and
verify/remove races cannot resurrect a removed record or alter its owner.

Provide bounded explicit verification/recheck/removal operations and failure/
suspension states. Distinguish confirmed loss/mismatch from transient resolver
failure, pin when a previous proof ceases to allow routing, and fail closed once
the selected freshness deadline expires. No unbounded polling/background service
is implied. Removal disables new tenant resolution and authentication immediately
under the declared transaction/cache ordering; reassignment needs fresh proof
and cannot authorize any previous tenant credential. Preserve canonical identity
and attributable audit/history where required rather than unsafe cascading deletes.

DNS/network adapters have cancellation/time bounds, finite records/redirect-chain
limits and no arbitrary URL fetch, supplied resolver, internal metadata endpoint
or unexpected certificate-management target. Do not weaken local/private-host
protection to make disposable fixtures work; fixture-only behavior must be isolated
from production.

### R4 — TLS readiness and adapter boundary

Routing readiness requires both current ownership/configuration proof and the
selected TLS readiness contract for that exact hostname. TLS availability alone
never proves tenant ownership. Model requested/pending/ready/error/expired or
equivalent explicit outcomes, with bounded issue/status/renew/remove adapter
operations and retry/idempotency semantics; no falsely successful activation on
provider timeout, unknown outcome or an invalid certificate.

Pin verification of hostname, trust/expiry and edge binding, renewal failure/
expiry behavior, withdrawal and late provider completion after removal. Keep
provider identifiers and credentials behind the adapter; never store certificate
private keys or secrets in management evidence or logs. Select the production
TLS/edge/DNS mechanism with owner/ADR approval as needed before Plan ACCEPT.
Disposable adapter/HTTPS fixtures prove application decisions, not a live
certificate provider or deployed ingress; required bounded live evidence and
operator prerequisites must be stated rather than silently deferred.

### Owner direction — infrastructure versus tenant API lifecycle (2026-10-15)

Use the Architecture §51 ownership split (not included):
Terraform is optional deployment tooling for shared platform DNS/wildcard,
ingress and TLS configuration. Tenant custom-hostname lifecycle uses one selected
provider API adapter; no Terraform plan/apply per tenant or generic provider
framework. Assign each concrete resource to exactly one reconciler. Customers
normally add their own DNS records, without giving Beginify their DNS credentials.

The Plan must state the resource boundary, applicable provider API operations,
idempotency/unknown-result handling and qualification prerequisites within the
existing R3/R4 and AC4–AC6 scenarios. This adds no Terraform installation,
production deployment, provider purchase, additional test campaign or early
planning/execution release. Provider choice and other draft decision gates remain.

### R5 — Scalable browser authentication and exact origin binding

Use one WorkOS web application/client per platform environment. Adding platform
Organizations must not require per-tenant applications/client IDs or individual
callback/sign-out entries. Select a provider-supported wildcard or shared-callback
contract with a concrete default and exact application-side host/origin validation.
A provider wildcard is not an application origin allowlist. Record remaining
custom-domain provider configuration requirements and any applicable service-tier
limits; do not claim unlimited or configuration-independent behavior untested.

Custom-domain authentication must preserve the same global canonical user UUID,
existing issuer/client namespace and current Organization membership checks.
Preserve Secure/HttpOnly host-only session cookies, browser-bound one-use state,
PKCE S256, exact initiating origin/destination and login-CSRF/replay protection.
No broad `Domain` cookie; no provider bearer, refresh token or reusable application
session credential in a URL. A shared callback/single-use cross-host handoff is a
candidate requiring owner/security approval, not selected by this Spec. The Plan
must show browser possession binding across origins, exact return validation,
expiry/one-use ordering and how changed/removed domain state blocks completion.

Sign-out preserves truthful local/provider outcomes, exact permitted return
destinations and existing global-versus-local scope. Authentication on A cannot
silently issue B's host cookie or grant B membership. No forced MFA or new MFA
assurance claim; preserve accepted optional enrollment and identity conflict rules.
Required real provider/manual checks occur before final qualification under the
existing sequencing rule; fixtures alone do not prove provider callback support.

### Owner-approved R3/R5 security amendment — 2026-10-16

The owner answered **“I approve”** to both recommendations in the
H-1/M-1 disposition (not included).
These requirements supersede conflicting global-browser-bearer and exclusive
unverified-claim assumptions; [ADR-007](../../../adr/ADR-007-verified-domain-routing-and-authentication.md#owner-approved-security-amendment--2026-10-16-utc)
records the architecture decision.

- Bind tenant-host browser application sessions to issuing exact origin and
  canonical Organization throughout core credential resolution and request/context
  consumers. Missing binding/origin, foreign Organization and reassigned host fail
  closed. Global identity/UUID and WorkOS remain; other tenant hosts/aliases use
  separate local sessions. Preserve membership checks, browser binding/PKCE,
  current/all/provider logout semantics and separate MCP session-kind/resource rules.
  The Plan must specify migration and legacy-session handling without an unbound
  fallback. Never infer authority solely from a caller-supplied Host.
- Permit concurrent non-routing unverified claims without a global hostname
  reservation. Acquire exclusive ownership atomically on successful valid DNS proof,
  protecting established verified/active ownership. Pin promotion conflicts,
  challenge ownership, hostname readers, generation/version fencing, retry and audit
  behavior; no partial activation, cross-tenant mutation or proof reuse.
- Extend existing AC3/AC5/AC8/AC11 evidence with focused cross-host/cross-Organization
  credential replay, missing binding/origin, reassignment, separate-host sign-in,
  scoped logout and concurrent pending/promotion cases. Reuse unaffected evidence;
  no new full ops/rename/control campaign follows from this amendment.

Claude supplies a bounded Plan/file-list amendment for supervisor review before
code changes. Existing B7/checkpoint4 history and budget accounting remain; no
reset or new Goal is introduced. Owner approval grants these design corrections,
not final implementation/MR acceptance or provider/deployment authority.

### R6 — MCP aliases preserve local connections and exact audiences

Inherit final G014/G015 and ADR-005. Each connection retains its stored canonical
Organization, user/client, exact resource URL/audience, host and lifecycle state;
Group selection remains connection-local. A new verified alias is only a route,
not permission to accept an old audience at a new host or rewrite a live connection.
Pin existing-URL retention, rejection/replacement or explicitly reviewed migration
behavior for domain removal/rename/reassignment before implementation. Unknown or
terminal records never revive from tokens; wrong-host/audience attempts change no
connection, selection or authority.

Keep protected-resource metadata and transport at their accepted registered paths
with no anonymous tenant-data leak or marketing redirect. Reconcile configured
G014 host/identity-origin registries with durable verified-domain resolution; do
not add a second competing source of truth. Record provider Resource Indicator/
audience registration requirements separately from browser callback scaling.
Preserve WorkOS's owner-accepted broader same-user/client refresh reach; do not
claim independent refresh families or credential containment that G014 did not
prove. A separately issued exact-audience credential still faces its target
connection's live local authority.

### R7 — Performance, evidence and bounded integration

Assess the actual per-request hostname lookup count, unique/index lookup path,
cache invalidation, DNS/TLS work placement, recheck set size and mutation contention
under PERFORMANCE.md. Do not perform DNS/certificate network verification on
every authenticated request or scan every domain to resolve one hostname. Record
bounded query/plan observations only for new lookup/uniqueness/ordering paths;
escalate unexplained scans/fan-out/locks without starting a performance campaign.

Keep credentials, challenge secrets, callback parameters and private tenant data
out of logs/evidence. Use existing audit/privacy/recovery boundaries and a finite
lifecycle event mapping; no new audit/retention framework. Document restore/
configuration reconciliation so stale external state cannot silently route a
restored record to a different tenant. Operator/provider unknowns stay explicit.

## Acceptance criteria and verification

| ID | Required outcome | Objective verification |
|---|---|---|
| AC1 | Normalization and trusted ingress resolve exactly one registered canonical Organization, with no tenant fallback. | Table-driven normalization and real HTTP edge/core probes cover valid canonical forms, malformed/ambiguous hosts, suffix lookalikes, disallowed ports and spoofed forwarding metadata. Unknown/unverified hosts reveal no tenant data or session; allowed fixtures prove the actual trusted path. |
| AC2 | Platform slugs and reserved/public hosts obey the configured routing boundary. | Two Organization subdomains resolve correctly; reserved apex/`www`/CNAME/infrastructure labels and collision attempts cannot be allocated. Public entry redirects only to configured marketing; selected protocol/callback paths preserve exact handling. |
| AC3 | Domain registration and mutation enforce ownership and concurrency constraints. | Real runtime-role policy/RLS checks show permitted management, denied ordinary/foreign actors and no cross-Organization disclosure. Concurrent duplicate claims/slug updates yield one owner; UUID identity stays unchanged and denied operations have zero effects. |
| AC4 | Only fresh matching DNS ownership and CNAME/configuration proof can reach verified readiness. | Bounded resolver fixtures cover correct proof, shared-target-only, wrong/foreign/stale/replayed proof, wrong target and resolver failure. Pending/failed cases never route; authorized retry follows the declared state transition. |
| AC5 | Recheck, removal and reassignment cannot retain or resurrect stale routing authority. | Exercise proof loss/freshness expiry, transient failure rule, remove-versus-verification race, late provider completion and fresh reassignment. Assert exact resulting state, no old-tenant resolution and no unauthorized old session/connection effect. |
| AC6 | TLS lifecycle controls routing readiness and records failure/unknown outcomes accurately. | Adapter contract tests plus actual disposable HTTPS checks cover pending, ready, wrong hostname/trust, expiry/renewal failure and removal. Verify bounded cancellation/idempotency and selected production adapter qualification evidence or explicit unresolved gate; a fake alone cannot close required provider evidence. |
| AC7 | Platform authentication scales without per-Organization WorkOS applications or individual callback/sign-out entries. | Authenticate through A and B, then add C and authenticate without those provider configuration additions. Record sanitized before/after configuration, actual supported pattern/default, canonical user mapping and membership denial; real provider/manual evidence covers what the emulator cannot. |
| AC8 | Verified custom-domain sign-in/sign-out preserve browser binding and existing security. | Browser navigates the actual HTTPS custom hostname under the selected reviewed mechanism; confirm host-only Secure/HttpOnly cookies and stable user UUID. Negatives cover altered return, foreign initiating browser, replay/expiry, removed domain and wrong membership, with no unauthorized session. Verify truthful local/provider sign-out and no reusable credentials in redirects. |
| AC9 | Domain aliases and lifecycle cannot retarget MCP credentials or connection context. | Reuse accepted SDK HTTP fixtures across platform/custom aliases and a removed/reassigned host. Assert exact stored audience/host agreement, A-at-B denial, no token-created/revived record and independent Group state; permitted separately authorized target connection succeeds. Verify accepted metadata path behavior and document provider audience-registration requirements. |
| AC10 | Configuration/provenance transition is coherent and remains brand/domain configurable. | Review and run existing relevant environment/template/rename checks against the revised defaults/manifests and one alternate suffix/CNAME/marketing configuration. Historical snapshots remain unchanged; no literal production-domain dependency or disabled consistency check. |
| AC11 | Domain state, privileged operations and lookup behavior retain audit/RLS/recovery/privacy boundaries with bounded cost. | Focused audit/denial/runtime-role and restore/configuration-mismatch cases plus source review. Record added query/index/lock observations and network deadlines; secrets/canaries do not appear in captured output. Review any performance/security escalation disposition. |
| AC12 | Delivery is attributable, qualified under actual CI mapping and cleaned up. | Result maps AC1–AC11 to reviewed/tested source and relevant input/content provenance, commands/statuses, manual/provider evidence and deviations. Record applicable MR/staging disposition and scoped cleanup receipt; absent/failed/held evidence is explicit. Codex final ACCEPT is recorded separately. |

## Verification strategy and provenance

Use six scenario groups: (1) canonical host/trusted edge/public reserved routing;
(2) management policy/uniqueness; (3) DNS/TLS verification and invalidation;
(4) browser scaling and custom-origin authentication/sign-out; (5) MCP host/audience
composition; (6) configuration/audit/recovery and delivery. Reuse finite existing
two-Organization identity/MCP fixtures, add the third platform growth case and one
custom domain, and use deterministic lifecycle barriers for actual races. Do not
multiply full stacks/suites across every hostname, role or state combination.

Pure normalization/state mapping use focused unit tests; domain ownership,
uniqueness and RLS require real PostgreSQL runtime-role evidence; composition
requires actual HTTPS/SDK/browser navigation at the tested origin. Use deterministic
DNS/TLS adapter fixtures for negative cases and bounded selected provider evidence
for real claims. No public DNS, provider settings or installed trust changes during
definition; later authorized qualification must protect unrelated environments.

Claude's Plan maps each AC to files/checks, specific manual/provider prerequisites,
failure/effect assertions and the actual accepted path-selected CI profile. Run
covering failed batches and affected regressions; retain unchanged evidence by
explicit content/input reconciliation. Shared foundation path changes may select
wider CI, but this Spec adds no independent runner/control campaign or full suite
per case. Required manual/operator evidence precedes final qualification, with an
explicit prerequisite-only exception if hosted artifacts are necessary.

Use existing Git-first bindings, exact source/target references, accepted validation
entry points, resource inventory and one scoped cleanup. Compact durable Results/
reviews, sanitized provider configuration facts and existing CI artifacts are the
evidence default; local raw evidence is temporary. Record failures, cancellations,
unknown outcomes and evidence expiry honestly. No new input manifest framework.

## Expected artifacts, permitted deviations and gates

Expected delivery: reviewed Claude Plan; bounded domain model/migration and
management operations; trusted lookup/routing and chosen DNS/TLS adapters;
reviewed browser/MCP integration; focused tests/minimal existing harness wiring;
configuration/provenance reconciliation; operator/domain-authentication runbook;
Result with AC evidence, query observations, attempts, deviations and cleanup.
Exact file list, migration number and mechanism belong to the later Plan.

No blanket deviations. Finite lifecycle names/internal layout and replaceable
adapter details may differ if observable AC outcomes remain identical and are
documented in the Plan. Changes to authority, browser/session transfer, audience
validation, DNS/TLS guarantees, provider/infra, schema risk or public contracts
require explicit review and owner/ADR decisions as applicable. Required evidence
cannot be deferred outside G016 to make an incomplete Goal pass.

Before Plan ACCEPT, resolve and record:

1. **DNS/TLS/edge mechanism:** adapter/provider, ownership challenge, verification
   freshness and renewal/removal model; required owner/ADR and scoped live test
   authority. Distinguish reusable implementation from production deployment.
2. **Browser/custom-domain authentication:** provider-supported scalable callback/
   sign-out configuration and exact cross-origin browser-binding contract; any
   shared callback/handoff changes G004 security behavior and needs explicit review.
3. **Domain administration:** finite owner/admin operation/capability matrix and
   minimal management surface; no implicit Group-owner or new public MCP authority.
4. **MCP alias transition:** old URL/audience retention, refusal/replacement or
   reviewed migration; resource registration and recovery effects under actual
   final G014/G015, without claiming stronger WorkOS credential isolation.
5. **Configuration rollout:** current generated-baseline to `.app` transition,
   reserved hosts/protocol exceptions, ingress prerequisite and required provider/
   operator evidence. Reconcile historical `.ai` examples explicitly.

These are draft decision gates, not permission for unbounded preapproval probes.
The supervisor decides whether an unresolved mechanism requires a bounded split;
neither agent silently implements a fallback. Final completion requires all ACs,
reviewed deviations/decisions, final dependency acceptance, required staging
disposition, cleanup and Codex final ACCEPT. This authorizes no G017 work or
product implementation before G021.

## Authority amendment — EPIC-006 continuation (2026-10-15)

The owner’s subsequent “Spec 17 and continue auto for epic 5 and 6” replaces this
Spec’s earlier draft-only/preplanning hold with the bounded normal delivery
authority recorded in PROGRESS (not included).
Provisional planning may compare the unresolved mechanisms at the initial G015
implementation MR; none of those alternatives is silently selected or qualified.
The listed owner/ADR choices, required evidence and Plan/dependency/binding/
budget/resource/supervisor-release gates remain before implementation. No new
probe or deployment authority is implied; AC1–AC12 and their validation scope are
unchanged. Stop at final G017/EPIC-006 closeout, before any G018 phase.

## Query tooling note — 2026-10-16

Follow Architecture §27.1 (not included):
Goose handles migrations; incremental sqlc adoption for ordinary queries is a
separate bounded follow-up, not a required deliverable or new acceptance criterion.
Hostname lookup, uniqueness and domain-state queries retain the existing pgx/
Goose approach unless a separately reviewed sqlc trial has already been accepted.
Do not couple DNS/TLS/auth qualification to query-generator adoption or refactor
inherited identity/RLS paths. Reconcile any accepted query-layer change in the
Plan while preserving trusted-host resolution, lock order and runtime boundaries.

## Owner amendment — introduce sqlc through G016 planning (2026-10-16)

The owner now asks to try sqlc in G016 planning. This supersedes the earlier query
tooling note only insofar as the bounded trial can be proposed inside G016 rather
than requiring a previously accepted separate follow-up. The Plan must record an
adopt/defer disposition for one suitable new ordinary domain data-access package.

If proposing adoption, pin generator/version, existing authoritative schema inputs
and actual Goose/procedural-SQL compatibility, UUID/null mappings, package/output
paths and reproducible generation. Keep generated methods on the caller-supplied
guarded pgx transaction/runtime role; authorization, trusted-host checks, lock
ordering and audit stay in existing core boundaries. No broad inherited query
conversion, second maintained schema, new runtime service or claimed query-speed
improvement. Identify exact file scope and any necessary existing dependency or
provenance-pin updates in the Plan for review.

Planning uses inspection/documentation only, with no tool installation, code
generation, build, provider probe or test campaign before execution release. If
reviewed adoption executes, verify generation/compilation and the already-required
relevant query/runtime-role cases; no new full suite, permanent CI stage or parallel
validation framework. Reuse existing evidence where applicable. If the trial would
obstruct domain/authentication delivery, defer with the reason in Plan/Result and
use handwritten pgx. AC1–AC12, B7 and all normal gates remain; planning does not
itself authorize installation or implementation, nor dispatch G016 early.


## Owner-approved design selection — 2026-10-16

The owner answered **“Approve those details”** to the recommended design package
following Plan r3 technical closure. [ADR-007](../../../adr/ADR-007-verified-domain-routing-and-authentication.md)
selects E1 Caddy target (production qualification remains explicit), membership
required at tenant-host sign-in, B2 shared WorkOS callback relay, retained exact-host
MCP records without automatic revocation, and DB-backed lookups with bounded
PG-007 observations. D5 internal-only surface and D6 reviewed provenance remain.
No live DNS, WorkOS settings, provisioning or deployment action is approved.
Plan/current binding/release and final evidence gates still apply; G021 stays closed.


## Owner amendment — live qualification split to G057 (2026-10-16)

The owner approved the [G016 → G057 scope/dependency amendment](../results/G016-G057-qualification-split.md).
It supersedes earlier mandatory-in-G016 real-provider/manual qualification and
no-deferral wording only for P-1/P-2: actual WorkOS callback/sign-out scaling and
live custom-host HTTPS/browser proof now belong to
G057 (not included), required before
G021. Mark these guarantees unqualified/transferred, never passed by fixtures.
G016 still requires reviewed implementation, the already-planned provider-test
extension, local security/HTTPS/emulator evidence, applicable MR/Duo and required
post-integration CI, Result and cleanup before final ACCEPT. Unresolved code or
security defects are not deferred by this amendment. P-3 remains optional; P-4
keeps its existing explicit production-adapter hold. No immutable Plan/history
is rewritten, no budget is reset and no broad rerun follows from management edits.
G017–G020 need accepted implementation contracts, not G057 live qualification;
the G021 foundation barrier additionally requires final G057 ACCEPT.
No owner DNS/VM/WorkOS/browser action or G057 dispatch is required/authorized now.

## Prospective root-entry supersession — G062, 2026-10-20

Owner replaces only R2/AC2's platform apex/public-entry marketing redirect with
[G062 root-host tenant signup](G062-root-host-tenant-signup-spec.md): generic
Sign Up-only landing and `/signup`, with tenant login at tenant hosts. G016 is
already accepted and is not deferred or struck. Preserve all other hostname,
reserved-label, verification, TLS, callback, session and MCP guarantees. Earlier
Plan/Result/acceptance evidence remains historical; new behavior is implemented
and qualified under G062 rather than retroactively attributed to G016.


## Prospective platform-address amendment — owner, 2026-10-23

G057's opaque-address amendment (not included)
prospectively replaces friendly-slug platform routing with a generated stable
opaque label directly mapped to Organization UUID. Friendly slugs remain metadata;
no routable slug alias/redirect or production compatibility period is required.
The change is implemented and qualified under G057's later reviewed Plan/release,
including affected signup/URL/seed consumers. This Goal's historical acceptance
and evidence remain unchanged; current code is not claimed to have changed.
