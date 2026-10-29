> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G071 Work Spec — Optional default-tenant root alias

Revision r3 — 2026-10-26 UTC. Priority: **High** — supervisor determination
from the owner's explicit reorder of G071 immediately after G069 and before G070;
the owner gave no low-priority direction. The Plan should target 3–5 parallel
implementation workers where independent paths and resource limits permit, and
record concrete reasons for any lower allocation. Plan r2 was accepted against
the pre-amendment Spec. The owner documentation amendment dated 2026-10-27 below
requires bounded Plan reconciliation. G069 merged as `09440305`; implementation
requires the corrected Plan ACCEPT, recorded release, and safe source/resource
handoff. G069 final ACCEPT gates G071 integration and final acceptance.
[Goal](../goals/G071-default-tenant-root-alias.md) /
[ADR-010](../../../adr/ADR-010-default-tenant-root-alias.md).

## 1. Required behavior

The operator creates an Organization through the existing normal flow, learns its
slug, then sets the optional **server-side** `DEFAULT_TENANT_SLUG` in deployment
configuration and restarts/redeploys. No seed/create/adoption operation is added.
The deployment operator may select any existing Organization in that deployment;
Organization ownership is not a prerequisite for this trusted configuration.
Users retain every existing authentication/membership/authority requirement.

| Startup configuration | Effective behavior |
|---|---|
| Unset or empty | Multi-tenant operation; root presentation follows PLATFORM_ROOT_PAGE (default signup redirect). |
| Existing Organization slug | Resolve its canonical UUID and activate the exact-root alias for this running configuration. |
| Unknown/non-resolving slug | Emit a clear structured warning, leave the override inactive and continue normal multi-tenant operation. Create nothing. |

Use the existing Organization slug syntax/normalization; malformed values receive
the same warning/inactive behavior rather than selecting a partial or guessed
match. Organization display names and generated platform labels are not alternate
selection keys. A database/permission/transport failure is **not** proof of absence:
use existing startup/readiness failure handling and report the failure honestly.

Selection takes effect on startup, not env-file edits alone. Pin the selected
UUID for that configuration; no watcher, per-request slug search or automatic
rebinding to another Organization within the process. Restarting with a different
valid slug selects its UUID; removing the flag or restarting with an unknown slug
restores normal behavior. Do not rename, mutate or delete either Organization.
Preserve existing liveness/authorization checks if the selected entity disappears.

### Multi-tenant root presentation

Add server-side `PLATFORM_ROOT_PAGE=signup_redirect|landing`. **Default:
`signup_redirect`**, including unset/empty. An invalid enum value warns and uses
that default; it cannot become a redirect URL. This selects presentation only
when no valid default-tenant alias is active, including unknown-slug fallback.
Do not infer the mode from the number of Organizations in the database.

| Effective mode | Request to the exact platform root `/` |
|---|---|
| Valid active DEFAULT_TENANT_SLUG | Tenant landing; PLATFORM_ROOT_PAGE has no effect. |
| Multi-tenant + signup_redirect (default) | Temporary same-origin redirect to `/signup`. |
| Multi-tenant + landing | Existing generic foundation landing with Sign Up action. |

Redirect only GET/HEAD of exact `/` on the configured public root origin, using
307 and a relative `Location: /signup`, without permanent caching (`no-store`).
Do not redirect `/signup` itself, other paths, POST actions or other hosts. Do not
forward arbitrary query parameters to signup, begin provider authentication or
create an Organization merely by following this redirect. Existing signup still
requires its normal user action and guards. Avoid a loop by resolving the valid
tenant alias first; its signup 404 policy takes precedence over root presentation.

The existing simple landing remains available through `landing`; removing the
variable selects the new redirect default. This dated default change supersedes
the former unconditional platform-root landing, not the loopback-only generic
development page. A separate marketing hostname/site is independent of this app;
no marketing site, new domain or hosting deployment is included.

## 2. Root routing and unchanged flows

Apply the alias only to the **exact configured platform root hostname/origin**,
under existing scheme/port normalization. Serve the tenant landing, branding,
sign-in, account and existing tenant routes on that origin; keep its address in
the browser. Do not spoof Host/Origin or internally pretend the browser requested
the opaque-label hostname. Core owns the trusted mapping; web uses that decision.
Existing API/MCP consumers of host resolution must remain consistent with it.

The root's usual Organization-creation CTA is replaced by the tenant landing.
**While a valid default tenant is active, the root's `/signup` and `/signup/*`
web routes return the existing 404 response**, including direct GET and form POST
paths. Use the existing host/path gate; hiding a link alone is insufficient. Do
not redirect these paths to `/join` or Group creation. This is the owner's later
simplification of the earlier direct-signup-compatibility requirement.

Keep existing Organization provisioning APIs, permissions and implementation
unchanged: no global creation-disable policy or new backend switch. With the
alias inactive (unset/empty/unknown slug), the existing root signup pages,
SignupIntent, forms/proofs and continuation flow return unchanged. Do not confuse
old tenant/session state with platform signup credentials during that transition.

New-Organization continuation still targets that newly created tenant's generated
platform hostname. Root aliasing must not newly admit `/signup/continue` or
`/signup/complete` on the root or change the existing continuation destination rules.

Route inventory (verified against current application source):

| Route | Meaning retained with the alias |
|---|---|
| `/` | Tenant landing when alias active; otherwise PLATFORM_ROOT_PAGE chooses signup redirect or existing landing. |
| `/signup` and its root subpaths | 404 on an active valid alias; normal Organization signup when inactive. No `/organization/signup` route exists. |
| `/join` | Enrollment into the selected Organization under its existing open/request/invite policy. |
| `/groups/new` and existing onboarding paths | Authenticated, authorized Group/Business Profile creation after admission. |
| `/sign-in`, `/auth/*`, account and tenant API paths | Existing tenant flow using the root's actual origin and the trusted selected UUID. |

Do not automatically redirect `/signup` to `/join` or Group creation. The tenant
landing already exposes the appropriate Sign in/Join actions. Inventory the small
root-specific host/path gate and auth-origin branches in the Plan; preserve the
existing signup intent/API code for normal mode. No route namespace redesign or broad
authentication rewrite is required merely because the root gains an alias.

Existing public enrollment and Group/Business Profile onboarding work through
their current rules. Merely arriving at the root grants no membership, Group,
Organization role or ownership. No automatic Group signup mode is invented.

Other platform-label/custom hosts, `www`, shared auth/reserved hosts, probes and
unknown hosts retain existing behavior. Do not route arbitrary hosts to the
default Organization. Existing tenant subdomains continue working. No wildcard
DNS, certificate, new domain proof or real WorkOS mutation is required by this
feature; reuse the root's existing deployment configuration.

## 3. Security and implementation constraints

Preserve UUID identity, tenancy/RLS, membership/role/revocation checks, exact-origin
cookies, CSRF, one-use login/signup states, callback privacy and final issuance
guards. A root session or login transaction for Organization A must not grant
Organization B access after an A→B restart. Unsetting the alias must not reinterpret
a former tenant credential as a platform signup credential. Preserve exact MCP
host/audience/connection binding; never rewrite existing tokens or connections.

Use a bounded least-privilege slug-to-UUID lookup. Current `app.hostname_state`
intentionally resolves opaque platform labels, while `signup_slug_taken` only
reports availability. The Plan must identify a suitable private existing reader
or propose the smallest narrowly scoped reader/migration; do not repurpose all
platform-label routing as slug routing, grant broad Organization SELECT, use an
owner/superuser runtime connection or expose a public slug-to-tenant directory.
No new DML is required for selecting this existing Organization.

Keep one authoritative server configuration and normal deployment propagation;
no `NEXT_PUBLIC` selector, request-supplied tenant override or second web mapping.
Document warnings/effective state without logging secrets or dumping environments.
Shared staging may leave DEFAULT_TENANT_SLUG unset and explicitly select `landing`
if the existing landing is wanted; isolated automated configuration proofs
qualify the option. It is intended for production adoption, not gated to production
in a way that makes ordinary focused validation impossible.

## 4. Expected artifacts and bounded scope

Claude supplies a compact Plan with actual config/resolver/auth/gate paths, the
protected lookup, existing tests to extend, any exact migration/grant inventory
and targeted selectors. Use normal server deployment/env configuration and generic
examples; add the two optional settings to relevant env/docs/allowlists. G070 later
consumes them through its approved deployment transport, without becoming a G071
implementation or test prerequisite.
Preserve its unset behavior and local stores. No application kit/style changes,
bootstrap/seed feature, new process manager, schema tenancy-mode flag or provider
framework. If a small SQL reader is necessary, retain its focused catalog/ACL/
migration proof; do not convert that registration into a full ci-controls run.

## 5. Acceptance and focused validation

| AC | Required outcome and proof |
|---|---|
| AC1 — Configuration | Unset/empty, existing, unknown and invalid slug cases behave as specified; unknown warns and startup continues, while an unavailable lookup is not mislabeled unknown. Repeated startup does not create or modify Organizations. |
| AC2 — Usable root alias | Selected tenant landing/branding and authenticated tenant navigation work at the root's real origin, with existing membership/enrollment rules. Other tenant/custom/auth/www/unknown-host behavior remains attributable and unchanged. |
| AC3 — Route boundary | Valid active alias: root `/signup` and its subpaths return 404 for GET/POST, while `/join` and tenant sign-in retain existing policy gates. Unset/unknown alias restores normal root signup. Organization provisioning APIs/permissions and generated-host continuation destinations stay unchanged. |
| AC4 — Isolation and switching | Reuse focused resolver/identity checks to prove no implicit membership, no cross-tenant credential acceptance on A→B restart and no signup/session confusion on removal/unknown fallback; preserve relevant host/audience/CSRF negatives. |
| AC5 — Least privilege and operation | Source/SQL review and focused integration evidence for any changed reader/grants/migration; no broad runtime authority, data rewrite or new tenant. Generic deployment docs explain setting/removing the flag and restart semantics. |
| AC6 — Delivery and cleanup | Applicable hosted fast, source/Duo disposition, authorized staging merge, affected exact-merged-SHA checks, Result AC mapping/failure history and verified all-category cleanup; Codex final ACCEPT. |
| AC7 — Multi-tenant root presentation | Default/unset signup redirect and explicit landing work only at root GET/HEAD; redirect has fixed same-origin location/no-store and no loop/side effect/query forwarding. Valid alias wins for both enum values; unknown slug falls back to the selected multi-tenant presentation. Existing signup, Join, other hosts/paths and loopback behavior retain their documented handling. |

Prefer table-driven existing config/resolver/identity tests and a small affected
browser sample for root tenant sign-in, active-alias signup 404, preserved Join
policy, normal-mode signup restoration and the small root-presentation/precedence
matrix. Reuse
existing synthetic tenants/provider fixture machinery in private tests; no live
production mutation or new live provider qualification campaign. The Plan chooses
executable narrow selectors and explains any shared dependencies. Normal hosted
fast remains; no mandatory full local suite, full ci-controls, broad migrations/
browser umbrella or new harness framework. Ordinary registration/pin updates use
only affected controls. Report unexecuted checks as such, reuse attributable passes
and preserve failed runs; selected post-merge coverage is coordinator-owned.

## 6. Scheduling and history

The original instruction placed G071 after G070. The owner's later “We should
queue this for before 070 btw” supersedes that order: **G069 → G071 → G070**,
then G021 after all required acceptance/qualification. Prepare G071's compact Plan
as the next planning priority; code still requires Plan ACCEPT/release and source/
resource handoff. Preserve existing G070 worker checkpoints/frozen runs and hold
new G070 phases/integration behind G071; no Plan/budget reset. G071 does not need
G070's cloud environment for its focused qualification. G059 remains deferred.
Existing G057 qualification continues, with Part B consuming ready G070 later.
No full suite, new RC, cloud/provider action or code execution is authorized here.


Revision history: r0 kept direct root signup available with the alias; r1 clarified
the route inventory and moved G071 before G070. The owner's later “maybe just slash
sign up if the default tenant is set and is a known slug ... becomes a 404 page”
selects r2's active-alias web restriction. `/join` explicitly remains available;
backend Organization provisioning remains unchanged. The published r0 remains in Git; the intermediate routing/order decisions are
recorded here. No frozen Plan or implementation snapshot has been rewritten.

r3 adds the owner's optional multi-tenant root presentation setting, with signup
redirect as the selected default and `landing` retaining the existing simple page.
It is ignored for a valid active default tenant. No marketing-host deployment,
Organization creation policy or new authentication flow is included.

## Owner documentation carry-through — 2026-10-27

During G071 implementation, update the platform `README.md` configuration and
operations guidance to describe both supported settings, `DEFAULT_TENANT_SLUG`
and `PLATFORM_ROOT_PAGE`: how to set/remove them, restart semantics, unknown-slug
warning/fallback, active-root `/signup` behavior, and the fact that this aliases
an already-existing Organization without changing the multi-tenant data model or
provisioning APIs. Keep examples generic and consistent with `.env.example` and
`docs/foundation/DOMAINS.md`. This makes the README an explicit G071 documentation
deliverable and narrows the existing generic-doc requirement in §4/AC5.

Carry the same accepted configuration into OS001's neutral exported README/setup
guidance when OS001 resumes and pins the accepted G021 source containing G071.
OS001's Goal and Spec are amended to require that carry-through; OS001 remains
PAUSED. This is the existing-tenant root-alias option, not a new G059-style
single-tenant database mode, seed operation, or provisioning-disable feature.
Because accepted Plan r2 explicitly excluded a README change, Claude must submit
a bounded Plan reconciliation for Codex review before performing the newly
specified documentation work. Preserve its other scope, source bindings,
budgets, and predecessor gates.

## Supervisor priority classification — 2026-10-27

Codex classifies G071 as **high priority** because the owner explicitly ordered it
immediately after G069 and before G070, and its scope crosses identity, tenancy,
and migration behavior. The owner gave no low-priority direction. This is
scheduling metadata only; it does not expand product behavior. The Implementation
Plan targets 3–5 implementation workers on independent paths where dependencies
and resource limits permit, recording concrete reasons for any lower allocation.

## Owner-directed G057 observation hardening — 2026-10-27

The owner directs the unresolved Duo finding on RC promotion MR !110 to be rolled
into G071, rather than opened as a separate maintenance task. This is one bounded
adjacent correction to G057's staging-domain observation helper; it does not alter
G071's tenant-routing behavior or any production request path.

**AC8 — Closed path allow-list rejects traversal.** In
`core/internal/identity/workos/staging_test.go`, `stgPathAllowed` must reject a
decoded path containing any `.` or `..` path segment before applying either the
exact-prefix or segment-boundary-prefix rule. The offline
`TestStagingDomainObservationFixtures` table must prove that literal
`/sign-in/../x` and percent-encoded `/sign-in/%2e%2e/x` redirects produce
`stgUnknown`, while `/sign-in` and `/sign-in/mfa` remain accepted under the
`/sign-in` allow-list and `/sign-inx` remains unknown. The test must exercise URL
parsing so encoded dot segments are checked in the decoded `URL.Path` consumed
by the helper. No provider credentials, live WorkOS call, product-runtime change,
or full `ci-controls` run is required for this item.

Add the exact offline selector and its build-tag invocation to the reconciled
Plan and include the helper path in its change-scope inventory. Reuse an existing
Core/test lane; this item does not justify another worker, attempt, standalone
MR, or broad qualification. Preserve the RC17 source/tag and its qualification
evidence as frozen; this addendum records the finding as deferred to G071, not
as already fixed. No RC merge or promotion authority is granted.

This addendum supplements §§4–5 with AC8. It does not rewrite the reviewed r3
text or settle any of the existing Plan-r3 corrections.

## AC7 implementation correction — absolute same-origin redirect Location — 2026-10-27

The private B1 run exposed an adapter contract issue: Next 16.4.1 returned HTTP
500 for the root redirect in E-DT-5/E-DT-6. The server reported
`TypeError: Invalid URL` with input `/signup`. The failed run is preserved in
the G071 evidence; the Result records its exact source and artifacts.

This addendum supersedes only the earlier **relative** `Location: /signup`
detail in §1 and AC7. The exact configured root must still return 307 on GET or
HEAD of `/` only when `signup_redirect` applies, but `Location` must be the
absolute URL `new URL("/signup", signupOrigin).toString()`, using the trusted
`signupOrigin` returned by the core host resolver. Do not construct it from the
request URL or an untrusted Host/Origin value. Keep the fixed `/signup` path,
same origin, `Cache-Control: no-store`, no body or cookie, and do not forward the
incoming query. The active-tenant precedence and every other host/path/method
remain unchanged. Keep the existing proxy gate; a route-handler rewrite is not
part of this correction.

Update the reconciled Plan's S17, U5 and E-DT-5/E-DT-6 assertions to require the
absolute same-origin Location and to verify the fixed path and empty query. Keep
the failed run and its evidence immutable; rerun the failed and affected root
checks after the fix, reusing attributable unaffected passes. No full suite or
`ci-controls` run is added by this correction.
