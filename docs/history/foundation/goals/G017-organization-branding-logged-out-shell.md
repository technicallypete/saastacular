> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G017 — Organization branding and logged-out shell

Epic: EPIC-006 — Organization domains and branding

Status: Definition reviewed and ready; provisional planning follows G016 implementation MR

Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation

Date: 2026-10-16 UTC; owner direction 2026-10-15 local

## Objective

Apply an Organization's authorized name, logo and bounded semantic color overrides
to its trusted-host generic entry shell and authenticated context indicators,
without exposing private tenant information or changing identity/permissions.

## In Scope

- Owner M1 feedback, 2026-10-17: replace the shared Light/Dark/System segmented
  control with a compact icon-trigger dropdown/select (sun/moon/monitor for the
  selected preference), preserving appearance/accessibility/pre-paint behavior.
  See the dated Spec amendment; other preview visuals received positive feedback.
- Core-owned, UUID-scoped branding configuration with a finite privileged update,
  reset and logo lifecycle; least-privilege/RLS and audit integration.
- A minimal generic logged-out tenant shell resolved through G016, using an
  explicitly public branding projection and the existing sign-in entry.
- Authenticated current-Organization indicators derived from current core context,
  with no tenant/Group switching or membership expansion from visual selection.
- Existing semantic tokens, TokenScope/portal isolation, bundled fonts, adapted
  Vestive/Radix components and Light/Dark/System appearance behavior.
- Safe logo handling through existing ObjectStore boundaries; tenant-safe asset/
  response caching, configured platform fallback and focused validation.

## Out of Scope

G016 domain/DNS/TLS or cross-origin authentication redesign; G018 Group/profile
onboarding; public signup/recruitment/acquisition funnels; service catalogs,
workshops, testimonials, page builders, navigation editors or product screens;
arbitrary CSS/JavaScript/HTML/themes, remote asset/font fetching, custom font
uploads, new identity/storage vendors, role taxonomy or a third tenant tier;
whole-kit adoption, unrelated refactors, broad CI/test-framework changes.

## Dependencies

- G016 finally ACCEPTED/integrated before implementation, including trusted host
  resolution, removal behavior, public protocol exceptions and browser/MCP binding.
- Accepted G002/G048/G049 UI, token/appearance/portal and scoped-skill contracts;
  existing identity, tenancy, policy/RLS, storage, audit and security foundation.
- [Work Spec](../specs/G017-organization-branding-logged-out-shell-spec.md),
  Architecture §§37A/50/51/61, ADR-001/003/004/005/006, WEB-SHELL, AUTHORIZATION,
  OBJECT-STORAGE, SECURITY and PERFORMANCE; reconcile actual accepted predecessors.

## Completion Definition

Spec AC1–AC10 have reviewed implementation and attributable focused evidence,
including required manual visual feedback before final qualification. The public
projection/asset surface, privileged mutation mapping and no-brand fallback are
documented. Required integration/staging disposition and scoped cleanup are
complete; Codex records final ACCEPT. G021's product-feature barrier remains closed.

## Execution Policy

**Architecture/security/data tier with a bounded ordinary UI delta.** The owner's
“Spec 17 and continue auto for epic 5 and 6” authorizes this definition and the
normal reviewed lifecycle within the recorded EPIC-005/006 membership. It does not
authorize G018 or other later epics. Use the current supervisor-recorded mandate.

Start provisional Claude planning at G016's first verified implementation MR,
including draft, without waiting for CI/Duo/merge. Inspect/reconcile actual
predecessor contracts and unresolved configuration/security decisions; do not
implement while G016 is pending. Plan ACCEPT, every dependency ACCEPT, reviewed
mutation/public-projection contracts, current Git-first binding, budget/resource
readiness and recorded execution release precede implementation. No new owner
proceed prompt is required once those gates close; owner-owned decisions still
escalate. Reuse existing controls, not another validation/orchestration framework.

## Iteration Budget

**B5: five shared Goal-level execution attempts; checkpoint after attempt 3**,
under BUDGET_POLICY (not included). Existing UI/storage/auth components
bound the work; if the Plan exposes a larger public-asset or policy change,
request an explicit scope/budget decision before execution. Planning spends no
implementation attempt; inner loops and delegates share the parent attempt.

Apply current model-aware phase defaults at dispatch, verifying actual effort or
reporting unknown. An effort change does not reset budget or bypass approval.

## Context / Scope Budget

One branding record per Organization, the existing closed color-token vocabulary,
one logo publication path, minimal management controls and one reusable shell.
Five scenario groups use two Organizations, one no-brand fallback and selected
existing portal/appearance regressions. No full suite per tenant/theme, exhaustive
theme generator, duplicate policy oracle or cosmetic rewrite of accepted pages.

## Suggested Subagents / Delegation

One implementation owner coordinates core/web edits. Up to two bounded delegates
may review public-data/asset security and accessibility/token isolation; reviewers
are read-only by default. Claude integrates work and one B5 ledger. Use existing
coordinator/native isolated-worker ownership mapping; naming alone is no blocker.

## Stop / Escalation Conditions

Unaccepted/changed G016 contract; unknown-host branding fallback; private data in
public projection/cache; foreign or unvalidated logo publication; unbounded/active
asset or remote fetch; arbitrary style injection; weakened contrast, first paint,
portal or user-appearance guarantees; new role/capability/public surface beyond
reviewed mapping; auth/MCP/domain redesign; major dependency/vendor or destructive
migration; unexplained query/asset fan-out or contention; native permission refusal;
owner stop, context overrun, repeated material failure or exhausted budget. Return
the smallest bounded decision/correction rather than expanding the Goal.

## Owner continuation endpoint update — 2026-10-16

Owner **“I think we can authorize autopilot up to 19 spec”** extends continuation
after G017 under the current exact mandate (not included):
G018 may plan at this task’s first implementation MR and follow normal reviewed
delivery; Codex may define G019 after G018 Plan interfaces settle. Stop before
G019 planning/implementation and all G020+ work. This supersedes older G017-only
continuation stops, not this task’s functional scope, budget, review/dependency/
release requirements, live authority or G021.


## Dependency clarification — G016/G057 qualification split (2026-10-16)

Under the owner-approved [split](../results/G016-G057-qualification-split.md),
G016 final ACCEPT covers the reviewed implementation/local security boundary;
P-1/P-2 actual-provider/deployed HTTPS proof moves to G057 before G021. This Goal
consumes accepted predecessor routing/session contracts and does not wait for
G057. Its own manual visual/UX feedback, security, Plan, dependency, input,
budget and supervisor-release gates remain. No live custom-host readiness claim,
new infrastructure authority or waiver of implementation defects follows.


## Owner decision D-1 — Organization branding roles (2026-10-16)

Owner: **“017 owners and admins.”** Resolve D-1 to
`organization.manage_branding` for current `org_owner` and `org_admin` in the
trusted active Organization. The mapping covers branding update/reset, logo
upload/replacement/removal and the scoped branding-upload cleanup route under
the existing guarded object-deletion policy. Group owner/admin/member roles or
Organization membership alone do not grant this capability; an actor holding
one of those roles may act only if it also has a current qualifying Organization
role. Preserve live membership, origin/Organization session binding, tenant/RLS,
CSRF, stale-edit and audit checks; branding never grants identity or membership.

Claude appends the finite role mapping and affected positive/negative assertions
to the accepted r2 Plan/history, preserving its exact imported snapshot. No full
replan is needed. This closes only D-1: G016 final ACCEPT/current-target
reconciliation, input/budget/resource readiness and recorded execution release
remain required. B5/checkpoint3 and the G057 qualification split are unchanged.

## Mechanical correction autonomy — owner direction, 2026-10-17

Owner “Make that change” authorizes the current mechanical-correction policy (not included):
a released implementer handles in-scope selectors/fixtures, approved-value catalog/
count/version pins and harness/import bookkeeping directly, preserving assertions
and recording focused evidence/Plan wording reconciliation in the Result. No
separate approval or full replan for each mechanical inconsistency. Exact scope,
security/contracts, unapproved schema/constraints, budget/native stops, required
validation and final Codex review remain. Apply at the current safe boundary; no
worker restart, new test campaign, frozen-input rewrite or budget reset. For held
Goals this changes guidance only, not execution/live/publication authority.

## Staging repair scope05 — 2026-10-17 UTC

Codex review (not included) permits the exact
boundary.spec.ts appearance-selector correction (scope49→50), existing harness
branding inventories/copy lists and actual320px layout diagnosis/fix. Attempt2/B5,
checkpoint3; no weakened assertions, full local campaign or change to final gates.
The separately identified DNS hop-bound issue is proposal-only until reviewed.

## Owner includes G016 CNAME repair alongside G017 fixes — 2026-10-17

Owner: “The 016 cname issue should be fixed alongside 17 fixes.” Include restoration
of the existing immediate-CNAME-edge/three-hop proof contract in the current G017
staging repair effort. This supersedes any deferral of that correction to a later
roadmap Goal. It does not authorize weakening the TXT challenge/hop bound or
replacing a real fix with serial tests/retries. G016's historical acceptance stays
recorded; the newly verified defect and its correction remain explicit.

Use the current G017 correction attempt2/B5/checkpoint3, with a bounded DNS worker
in parallel to the mechanical/layout repair worker and G018 implementation.
Keep exact non-overlapping file/branch ownership and attributable DNS/web commits;
no new Goal, fresh budget or reopening the old G016 implementation loop. Coordinate
review/integration and qualify the combined repaired source; avoid separate full
local gate campaigns for each fix. G017 final acceptance remains held until the
DNS correction and all required G017 qualification pass.

Expedite the already-running concise DNS proposal and supervisor review of exact
paths, dependency pin, system-resolver policy and deterministic failure cases.
Then extend the current G017 repair allow-list and release that bounded production
correction without another routine owner prompt. Until those details are reviewed,
the current50-pattern scope remains; do not silently edit production DNS/module
files. Existing security/contract/risky-migration/native gates remain, and material
resolver-policy changes escalate. G018 is not assigned predecessor repairs.
Autopilot continues through G019/G058 and stops before G020 planning.

## DNS companion scope06 execution release — 2026-10-17 UTC

Codex design/scope release (not included)
accepts the exact authored r0+r1 proposal and six new exact patterns (50→56) for
the DNS helper/tests and core module manifests. Restore the approved immediate
edge/three-hop guarantee with typed CNAME queries, bounded resolver behavior and
reviewed x/net pin; preserve TXT/authority/public contracts. Shared G017 attempt2,
parallel bounded writer, no new budget/task. Required qualification still gates
acceptance; G018 continues. Earlier proposal-only production hold is now released.
