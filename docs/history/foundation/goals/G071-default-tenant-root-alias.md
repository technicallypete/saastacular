> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G071 — Optional default-tenant root alias

Date: 2026-10-26 UTC
Current status (Codex final ACCEPT, 2026-10-27): G071 merged as `37bde74a`;
hosted fast/full qualification, Duo review and cleanup are complete. D-10's
local invocation is **UNEXECUTED** and accepted by equivalent hosted coverage,
not recorded as a local pass. RC18 is tag-only and unqualified. See the
[final Result](../results/G071-default-tenant-root-alias-result.md).

Prior planning status (superseded 2026-10-27): Spec r3 and Plan r2 were accepted on the prior binding. The 2026-10-27
owner documentation amendment and current execution allocation require bounded
Plan reconciliation. G069 merged as `09440305`; G071 implementation still requires
the corrected Plan ACCEPT, recorded release, and safe source/resource handoff.
G069 final ACCEPT continues to gate G071 integration and final acceptance.
Priority: **High** — supervisor determination, 2026-10-27, from the owner's
explicit reorder of G071 immediately after G069 and before G070 and the
cross-service identity, tenancy, and migration scope; the owner gave no
low-priority direction.
Owner: Codex Goal/Spec/architecture/review; Claude Code Plan/implementation/validation/Result.

## Objective and authority

Let a deployment operator set `DEFAULT_TENANT_SLUG` to an existing Organization's
slug and, after restart, serve that tenant's normal experience at the configured
platform root hostname. Unset or unknown slug retains normal multi-tenant behavior;
unknown logs a warning. An active valid alias makes root `/signup` and its web
subpaths 404; `/join` stays policy-controlled. Keep Organization provisioning
APIs and the underlying multi-tenant data/authority model intact.

For multi-tenant operation, `PLATFORM_ROOT_PAGE=signup_redirect|landing` selects
root presentation, defaulting to a temporary `/signup` redirect. The existing
simple landing remains selectable. A valid default-tenant alias always takes
precedence; this setting never redirects its tenant entry to a signup 404.

Owner authorizes a **new** Goal/Spec with “Okay, spec that out right now” and
“make this something to do right after O-7-O.” This is G071, distinct from the
larger deferred G059. No creation, seeding, special single-tenant database mode or
global Organization-provisioning disable switch is included. The later owner
route refinement limits only the active root’s signup web routes.

## In Scope

[Work Spec](../specs/G071-default-tenant-root-alias-spec.md): optional server-side
configuration, read-only startup slug-to-UUID selection, exact-root routing,
existing auth/session/gate compatibility, warning/fallback, deployment examples
and focused proof, plus the multi-tenant root-page selector. For an active tenant
alias the root URL stays in the browser; no redirect to a generated
tenant subdomain is required. Existing other tenant/custom hosts continue working.
The owner-directed 2026-10-27 G057 staging-observation path-normalization
correction is also included as a bounded adjacent hardening item; it changes no
runtime routing behavior and is defined in the Work Spec addendum.

## Out of Scope

Tenant creation/seeding, owner bootstrap, data migration into one tenant, deleting
other tenants, enrollment changes, Group auto-creation, direct public configuration,
hot reload, arbitrary host fallback, new DNS/TLS provisioning, a new routing
framework, UI redesign or changes to the unaltered brand kit. Preserve existing
Organization provisioning APIs and normal root signup whenever the alias is inactive.

## Dependencies and scheduling

Initial definition baseline: staging `443d2c95`; later ordering supersedes the
original after-G070 schedule. Consume the actual accepted application/G069 source
at the normal handoff; G071 does not depend on G070 hosting or final acceptance.
Never rebase a running predecessor proof checkout.

Owner correction: “We should queue this for before 070 btw.” Current order is
**G069 → G071 → G070 → G021**, with G057's existing qualification lane preserved.
The existing coordinator may prepare the compact G071 Plan as its next planning
priority; no duplicate planner. Code waits for Plan ACCEPT, recorded release and
safe source/resource ownership. G069 final ACCEPT gates G071 integration; normal
isolated successor overlap after predecessor merge remains allowed. G071 final
ACCEPT precedes G070 integration/next delivery phase. Preserve already-dispatched
G070 work through safe checkpoints; no discarding unique source or interrupting
safe frozen runs. Do not dispatch a new G070 phase or integrate it ahead of G071.
G070 later consumes the optional setting in its existing deployment config path,
reconciling actual effects without restarting its approved Plan or budget. G057
Part B still depends on ready G070, not on a separate G071 live rollout. Existing
G057/M1 holds remain; G021 requires all final ACCEPTs, including G071. G059 stays
deferred and RC authority/numbering is unchanged.

## Completion Definition

All Spec ACs have attributable source and focused evidence, current hosted fast/
Duo disposition, authorized staging merge, selected exact-merged-SHA checks and
all-category cleanup. Codex reviews the actual diff/Result and gives final ACCEPT.
Publication of this definition is not delivery or an execution release.

## Execution Policy

Bounded routing/identity configuration change: a compact Claude Plan and Codex
Plan ACCEPT/release precede code. Prefer existing resolver, signup-intent and
cookie/identity mechanisms; do not turn this into a tenancy-model rewrite. Explain
the minimum protected slug lookup and any necessary narrow SQL reader/migration
before approval; broad runtime grants or owner credentials are unacceptable.
Keep configuration authoritative in core and avoid independent web selections.

## Iteration Budget

**B3: maximum three shared Goal execution attempts; checkpoint after attempt 2**,
under BUDGET_POLICY (not included). Planning spends zero attempts;
record attempt 1 at released implementation. Use current model/effort policy.

## Context / Scope Budget

Two small env settings: root-host alias and multi-tenant root presentation, with
essential auth/signup compatibility.
No target line-count promise: state the actual bounded source/validation work in
the Plan. No new framework, full local suite or full ci-controls by default.

## Suggested Subagents / Delegation

At the owner's direction, begin implementation immediately after Codex accepts
and releases the reconciled Plan, without another owner start prompt. Use the
maximum safe number of dependency-ready, nonconflicting implementation workers
allowed by that Plan and current resource limits, targeting 3–5 for this
high-priority Goal where feasible. Keep one task owner/integrator and one writer
per artifact; the Plan must explain any smaller allocation or lanes that cannot
run concurrently. Independent read-only auth/SQL review may overlap where useful.
Respect G069/G070 ownership, the two-private-run pool, sequential builds and
immediate resource accounting/reclamation. Plan acceptance does not waive its
scope, attempt budget, G069 final-ACCEPT integration gate or other recorded holds.

## Stop / Escalation Conditions

Escalate signup/provisioning suppression beyond the approved active-root web paths, implicit membership/ownership, unsafe Host/
Origin rewriting, broadened RLS/runtime authority, stored-data conversion, public
tenant lookup, a new provider/domain rollout, unsafe resource overlap or exhausted
budget. Preserve failures and unaffected work; do not fall back on DB outages as
though the configured slug were proven absent.

## History

- 2026-10-26 — Owner selects existing-tenant root alias with graceful unknown-slug
  fallback and untouched Organization creation; requests a fresh Spec, immediately
  after G070. G059 is not reopened. Specification/scheduling only at publication.
- 2026-10-26 — Later owner correction queues G071 **before G070**. This supersedes
  the prior ordering only, preserving existing checkpoints, scope and review gates.
- 2026-10-26 — Later route decision: with a known active default tenant, root
  `/signup` and its subpaths are 404; `/join` remains normal enrollment.
  No backend provisioning-disable change; unset/unknown restores normal signup.
- 2026-10-26 — Owner adds optional multi-tenant root presentation: default
  `/signup` redirect or existing landing, ignored when a valid tenant alias is active.
- 2026-10-27 — Owner directs the unresolved RC17 Duo finding in G057's staging
  observation helper to be rolled into G071, rather than a standalone maintenance
  MR. The bounded test-only correction and its focused offline proof are in the
  dated Work Spec addendum; RC17 remains frozen and unchanged.
- 2026-10-27 — Owner directs implementation to start immediately after the
  reconciled Plan receives Codex ACCEPT/release, using the maximum safe number
  of independent implementation workers. No extra owner start prompt is needed;
  Plan, resource, budget, and G069 integration/acceptance gates remain binding.
- 2026-10-27 — Codex final ACCEPT after MR !112 merged at `37bde74a`, hosted
  fast/full qualification and Duo approval. D-10's local invocation remains
  UNEXECUTED and is accepted by equivalent hosted coverage. Cleanup and ref
  disposition are recorded in the final Result; RC18 remains tag-only and
  unqualified. G070's predecessor hold is cleared.
