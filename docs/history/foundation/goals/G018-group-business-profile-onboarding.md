> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G018 — Group creation and create-or-select Business Profile onboarding

Epic: EPIC-007 — Group and Business Profile onboarding

Status: Defined; authorized delivery under the dated mandate extension below,
with provisional planning at G017's implementation MR and normal execution gates

Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation

Date: 2026-10-16 UTC; owner direction 2026-10-15 local

## Objective

Let an authorized existing Organization member create an owned Group, then create
and attach or select an eligible Business Profile. Present onboarding as complete
only with exactly one same-Organization profile, preserving derived permissions.

Superseding owner D-1, 2026-10-17: generic self-service creation permits current
Organization owners, admins and ordinary members in the active Organization to
create their own new Group. The verified creator becomes its initial owner,
without Organization elevation or other-Group access. The earlier owner/admin-only
choice is retained as superseded history in the Spec/DECISIONS. This decision
does not approve the pending Plan or release implementation; public signup and
membership enrollment remain separate from this authenticated foundation flow.
G058 owns generic enrollment after G019 and before G021, then reuses this flow;
G042/G043 add later product/commercial composition. This reference does not
expand G018 implementation scope or require restarting its current Plan.

## In Scope

- Authenticated Group creation, atomic initial creator ownership (one or more
  owners permitted thereafter, never an exactly-one-owner constraint) and suggested/editable slug,
  using canonical UUIDs and existing server validation/uniqueness constraints.
- Resumable incomplete Group onboarding; atomic new-profile creation/attachment
  or authorized existing-profile selection, completion and bounded retry handling.
- Guarded Group/profile lists, profile reads and ordinary profile edits under
  current Organization/Group policy, including the existing edit-policy flag.
- Sibling Groups/Business Profiles management areas and one shared profile page;
  Group shortcut opens that canonical page without copying state or authority.
- The smallest reviewed core capability/grant/RLS and lifecycle additions needed
  to expose these previously trusted-only primitives; thin UI/BFF, audit and tests.

## Out of Scope

Public signup, Organization creation/membership enrollment, invitations/member
administration/ownership transfer (G019), completed-profile reassignment/removal,
profile deletion/merge, billing/legal/privileged settings, advisors/consultants,
payments/workshops, product work objects or AI intake; domain/auth/branding redesign;
third tenancy tier, independent profile roles or duplicated owner rows; generalized
onboarding engine, retry framework, runner/harness redesign or vendor adoption.

## Dependencies

- Final accepted/integrated G017 and its G016/G014/G015 routing, identity and
  connection contracts; actual accepted implementation must replace draft assumptions.
- Accepted G005/G006 creator ownership/cardinality, G007/G008 policy/RLS, storage,
  audit, security/recovery and G048 UI foundation. ADR-004 governs profile authority.
- [Work Spec](../specs/G018-group-business-profile-onboarding-spec.md), Architecture
  §53.4/§53.5, AUTHORIZATION, DATABASE-ISOLATION, SECURITY and current governance.

## Completion Definition

Spec AC1–AC10 have attributable reviewed source and required validation, including
completion/revocation/concurrency negatives and manual UX feedback before final
qualification. Integration disposition and scoped cleanup are complete; Codex
records final ACCEPT. No product-feature release before G021 is implied.

## Execution Policy

**Architecture/security/data tier.** The owner's “Spec” authorizes this definition
only. The earlier EPIC-005/006 autopilot still ends at G017. **Do not dispatch a
G018 Plan or implementation**, including at a G017 implementation MR, without new
explicit owner phase authority. No G019 definition or other work is authorized.

After a later planning release, Claude reconciles actual accepted predecessors
and proposes the finite authenticated operation/capability and runtime grant/
RLS contract. Do not expose trusted primitives or broaden runtime grants by
assumption. Plan ACCEPT, dependency ACCEPTs, decisions, Git-first binding,
budget/resource readiness and supervisor execution release precede implementation.
Routine implementation choices may remain autonomous within that approved scope.

## Iteration Budget

**Provisional B7: seven shared Goal-level execution attempts; checkpoint after
attempt 4**, per BUDGET_POLICY (not included). G008 deliberately removed
runtime creation/link grants; safely adding authenticated creation, creator-role
bootstrap and race-proof completion requires security-sensitive multi-row/RLS
changes, beyond presentation-only B5 work. Reassess toward B5 if the later accepted
repository already supplies these guarantees. Definition/planning consumes no
implementation attempt; scoped inner loops/delegates share the parent's budget.

Use current model-aware phase defaults when eventually dispatched; record actual
settings or unknown. A rebase, interruption or session change never resets counts.

## Context / Scope Budget

One existing Organization→Group model, one shared Business Profile association,
current finite Group/BusinessContext fields and five bounded scenario groups.
Reuse two-Organization policy/RLS fixtures, one profile shared by two Groups and
one inaccessible profile. No new commercial onboarding state machine, exhaustive
role permutation campaign or full stack/suite per form/tenant/profile.

## Suggested Subagents / Delegation

One implementer owns overlapping core/web work. Up to two bounded delegates may
review creation/completion authorization and UX/accessibility; reviewers default
read-only. Claude integrates work and one B7 ledger through the verified existing
coordinator/native isolated-worker mapping. Naming alone is not a blocker.

## Stop / Escalation Conditions

Any owner phase hold; unaccepted/changed G017; missing authenticated creation
authority; broad runtime insert/link/role privileges; new Group authorizing an
unrelated existing profile; completed Group with zero/multiple profiles; stale
authority or retry creating partial/duplicate committed objects; selected-Group
restriction bypass; independent profile roles/third tenancy tier; new public
membership/transfer/reassignment or product scope; architecture/ADR conflict;
unsafe migration, unexpected query fan-out/locks, major dependency/vendor, native
permission refusal, context overrun, repeated material failure or budget exhaustion.

## Owner mandate extension — 2026-10-16

Owner **“I think we can authorize autopilot up to 19 spec”** supersedes this
artifact’s earlier definition-only/G018 planning and execution hold. Under the
recorded exact mandate (not included),
G018 may plan provisionally at G017’s first actual implementation MR, then proceed
through normal reviewed delivery and closeout after dependency ACCEPTs, Plan
ACCEPT, exact inputs, decisions/budget/resources and supervisor execution release.
No additional routine owner proceed prompt is required. Existing functional
requirements, provisional B7/checkpoint4 and escalation/validation rules remain.

Codex may define G019 after G018’s Plan settles the relevant interfaces, but must
**stop before G019 Claude planning/implementation**. G020+ is not authorized.
This is a bounded phase mandate, not live provider/deployment authority or a
waiver of the G021 barrier. Earlier authority text remains historical.


## Dependency clarification — G016/G057 qualification split (2026-10-16)

Under the owner-approved [split](../results/G016-G057-qualification-split.md),
G016 final ACCEPT covers the reviewed implementation/local security boundary;
P-1/P-2 actual-provider/deployed HTTPS proof moves to G057 before G021. This Goal
consumes accepted predecessor routing/session contracts and does not wait for
G057. Its own manual visual/UX feedback, security, Plan, dependency, input,
budget and supervisor-release gates remain. No live custom-host readiness claim,
new infrastructure authority or waiver of implementation defects follows.


## Owner amendment — one or more owners (2026-10-16)

Owner approved **multiple owners, with at least one required**, then directed
“write this in for 18 and 19 goals and specs.” This supersedes the preceding
exactly-one-owner proposal; exactly one Business Profile per completed Group is
unchanged. G018 creates the Group and the verified creator's initial owner binding
atomically, including unfinished Groups. Do not add a unique-owner constraint,
parallel owner column or extra admin binding; owner already includes admin powers.

[G019](G019-scoped-invitations-membership-ownership.md) owns additional-owner
promotion, atomic transfer and universal last-owner protection for Organizations
and Groups, including account retirement, database/runtime guards and legacy
bootstrap reconciliation. G018 must preserve these semantics in its creation
contract and focused AC1 evidence, without implementing G019 membership/transfer
UI early or gaining a forward dependency on G019. G019 definitions are now
explicitly authorized as provisional drafts; its planning/execution hold remains.
Existing budgets, G017 dependency, bindings/review/release and G021 stay unchanged.

## Owner amendment — overlap G018 implementation with G017 staging CI (2026-10-17)

Owner instruction: “We should start 018 impl before the hour long staging tests.
Anything broken, 018 workers can fix it.” This supersedes this Goal/Spec's earlier
requirement for final G017 ACCEPT **before G018 implementation** and the preparation-only
hold recorded in the G018 r5 review. G017 source is reviewed and integrated at
`3cc1822826754ff70cf6346809f62b417080832c`; staging pipeline `2958742766` is
still running. It is not recorded as passed or finally accepted.

G018 may execute its accepted r5 design on an isolated task branch once the
current-target R0 reconciliation, exact scope/input binding, initial Result,
resource readiness and Codex release are confirmed. Do not wait for G017 staging
CI/final ACCEPT solely to start that implementation. Other accepted dependencies,
security/architecture requirements, B7/checkpoint4, focused validation and the G021
barrier remain. G018 integration and final ACCEPT still require final G017 ACCEPT.

If G017's gate exposes defects, the same available workers may perform bounded
repairs after Codex records the failing revision/evidence, exact repair scope,
branch/writer and validation selection. Keep predecessor repair and G018 changes
attributable and their existing attempt ledgers separate; no budget reset, broad
rewrite or silent expansion of the G018 allow-list. Pause only work affected by
a changed contract or unsafe shared resource; preserve unaffected completed work.
No need to restart G018 or repeat passing checks for unrelated management changes.
G019/G058 planning/execution holds are unchanged.

## Owner clarification — predecessor retains staging repairs (2026-10-17)

The owner clarifies that staging pipeline failures remain the predecessor task's
implementation responsibility, as before. Its assigned owner handles scoped
repairs using the predecessor branch and attempt ledger, in parallel with the
successor's isolated implementation. Do not assign those failures to successor
workers by default or wait for predecessor qualification solely to start successor
work. Pause only affected/unsafe work; predecessor final ACCEPT still gates
successor integration/final acceptance. This supersedes the earlier wording about
the same available/successor workers fixing predecessor failures. Any worker
reassignment must be explicit and retain predecessor attribution, scope and budget.
No implementation restart, new test run, Plan rewrite or budget reset is required.

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

## Owner M1 UX amendment — 2026-10-18 UTC

Owner M1 feedback requests direct entry sign-in, inline Group-address conflict
feedback and dynamic blur-time slug hints. Approved bounded implementation uses
the existing guarded `/auth/login` POST for home and onboarding sign-in actions;
keeps `/sign-in` and all existing error/logout-return flows; moves definitive
create conflicts into the form with fresh authorized holder lookup; retains the
separate unknown-COMMIT recovery page without auto-retry or false success.

Add one authenticated same-origin GET BFF route `/groups/address-check?slug=…`,
reusing the current guarded exact-slug list. Closed advisory states are invalid,
in_use (a readable holder), or not_seen (filtered absence, never availability).
No holder names/IDs in JSON, no new core endpoint/capability, no schema/vendor or
global uniqueness oracle. Backend create remains authoritative under races.
Clear hints on edit; generation/value checks, finite request lifetime and closed
bounded responses prevent stale or malformed hints. Preserve accessible field
association, native submit and script-free behavior.

See the exact Claude proposal (not included) and
binding review/release (not included).
This dated amendment supersedes only the prior separate-conflict-page presentation
and adds the reviewed BFF surface. Original approved Plan/history remains intact;
no new standalone Goal/Spec or general authentication redesign. Remaining owner
M1 questions and mandatory MR/Duo/staging acceptance gates stay in force.
