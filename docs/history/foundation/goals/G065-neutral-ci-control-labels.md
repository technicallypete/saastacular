> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G065 — Goal-neutral Docker labels for CI controls

Status: FINAL ACCEPT — Codex, 2026-10-23. Implementation MR !87 merged as `4afca837`; closing Result MR !88 merged as `efe6a9e1`. Hosted qualification, matching executable trees and scoped cleanup are verified. [Final acceptance](../results/G065-final-acceptance-codex-review-20261023.md). Earlier dated scope/release records below remain historical.
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation.
Authority: verified direct owner request in coordinator session `<redacted-uuid>`, 2026-10-22T19:45:57.194Z, message `<redacted-uuid>`: "Tell codex to make that goal". The preceding exchange proposes a separate label-prefix task after G064 final ACCEPT, preserving G064's existing names and evidence until then.

Current scope is governed by the [R1 + R8 amendment](#owner-scope-reduction--r1--r8-only-2026-10-23).
The earlier all-resource metadata rollout (R2–R7) is dropped and not scheduled;
its dated amendment below is historical. No common metadata rollout, new Compose
project-name policy or repository-wide resource-label implementation belongs to G065.

The [RC8 harness-repair amendment](#owner-rc8-harness-repair-amendment--2026-10-23)
adds only the three bounded repair items and focused proofs specified there.

The [2026-10-23 clean-cutover amendment](#owner-clean-cutover-amendment--2026-10-23)
supersedes the earlier dual-generation compatibility and reader-retirement requirements.

Current sequencing is governed by the [2026-10-23 clarification](#owner-start-clarification--rc8-independent-2026-10-23): G065 follows accepted G064 and can proceed while RC8 qualifies independently.

Spec: [G065 Work Spec](../specs/G065-neutral-ci-control-labels-spec.md).

## Objective

Replace the legacy Goal-number-based Docker label namespace used by the CI
controls harness with a stable Goal-neutral namespace, preserving resource
attribution, startup holds and cleanup through a verified cleanup and coordinated
cutover of affected writers/readers.

## In Scope

- Inventory the producers and consumers of `g003.controls.run`, `g003.control`,
  `g003.outer` and `g003.controls.owner`; define an explicit equivalent mapping
  to the neutral namespace in the Work Spec/Plan before implementation.
- Update applicable creators, owner checks, cleanup/recovery readers, counts,
  focused fixtures and current runbook text together. Preserve owner-key values,
  run/control/nesting identity and exact resource-ID records.
- Stop affected creators, clean up their exact owned old-label resources and
  verify absence before changing all inventoried writers/readers together.
  After cutover use the new keys only; no dual-generation compatibility or
  separate legacy-reader retirement phase. Uncertain/live required retention
  holds cutover rather than authorizing unsafe deletion.
- Focused cleanup-boundary, new-label, protected-resource and affected
  cleanup/recovery checks using existing selectors/fixtures and attributable
  unchanged G064 evidence; record actual source and resource outcomes.

## Out of Scope

Any change to current G064 implementation, labels or proof bindings before its
final acceptance; renaming every historical G003 helper, storage path, environment
variable or Goal artifact; unrelated Docker namespaces such as `g001.*`;
production tenant/authentication behavior; a new resource registry, daemon,
scheduler, reporting framework or broad performance campaign; rewriting old
evidence or authorizing cleanup from a label alone.

## Dependencies

G064 final ACCEPT and current-input/resource reconciliation are required before
this task's planning/execution starts. Reuse G064's accepted accounting, owner and
recovery mechanisms. Preserve all existing architecture/security, cleanup,
resource reservations and G021 delivery-barrier rules.

This is a separate queued maintenance task after G064, not an addition to G064's
remediation or a new prerequisite for RC8, G020, the staging preview or G021.
Its exact execution slot remains to be assigned within the existing owner work
bounds; creating this Goal does not dispatch a planner or implementation worker.

## Completion Definition

The reviewed Spec's mapping and clean-cutover requirements have attributable
evidence against the actual diff. Affected old creators and their resources are
verified gone at the cutover boundary; active writers/readers then use the new
keys together without disturbing live/foreign runs. Source review, required hosted fast
checks, authorized staging integration, affected post-merge local checks and
verified task cleanup are complete. Codex records final ACCEPT. Preserve earlier
failures and historical labels in old evidence; no blanket local full run is
required solely for naming, unless concrete shared impact warrants one.

## Execution Policy

Validation/cleanup infrastructure change: a concise explicit Claude Plan and
Codex Plan ACCEPT/current-input execution release are required. Write/review the
Work Spec before planning. Goal creation alone releases no implementation,
builds, Docker work, new permissions or hosted actions. Keep one task owner and
one writer per artifact; use the current model/effort and MR policies.

## Iteration Budget

**B3: three shared Goal-level execution attempts, checkpoint after attempt 2**,
under BUDGET_POLICY (not included). Attempt 1 is spent on the original
rename candidate; the owner-directed RC8 repair does not reset the counter.
Claude's amendment Plan must confirm the remaining budget contains the bounded
repair work; expand neither scope nor budget silently.

## Context / Scope Budget

The four control-label keys and their actual producers/readers only. Reuse the
accepted G064 design and focused proof infrastructure. Respect the current
150k checkpoint / 200k safe-handoff thresholds and smaller native limits.

## Suggested Subagents / Delegation

One Claude implementation owner; optional bounded read-only review of the
exact-key producer/reader and cleanup/cutover mapping. Workers share B3 and existing
resource limits. No extra coordinator or mandatory agent split.

## Stop / Escalation Conditions

Stop affected work for changed ownership/removal authority, resource misattribution,
an incomplete cleanup boundary, production changes, unavailable
proof, expanded file scope, native refusal or exhausted budget. Preserve exact
IDs and non-green records; do not bypass cleanup holds or rewrite G064 outcomes.

## History

- 2026-10-22 — Created from the verified owner request, queued after G064 final
  ACCEPT. No Spec/Plan or execution release issued.


## Owner scope amendment — project/run labels and simpler cleanup, 2026-10-22

Historical scope: the R2–R7 rollout described in this section is superseded and
dropped by the 2026-10-23 R1 + R8 reduction below. It is not an active G065 deliverable.

The owner requests adding consistent Docker project attribution to this queued
Goal and to reusable/live workflow guidance: label resources created through both
Compose and direct Docker commands, use the Compose project-name environment
variable appropriately, and simplify cleanup around that attribution.

This extends the original four-key-only scope above to project-controlled Docker
creator/reader paths needed for a common ownership-label contract. It does not
require renaming unrelated legacy keys such as `g001.*`; those may coexist with
new common project metadata. Preserve the original four-key neutral-prefix
transition and all earlier authority/evidence history.

- Define the exact common label schema in the Spec/Plan: stable repository ID,
  disposable run ID and responsible ownership/lifecycle. Use an appropriate
  nonreserved namespace. Keep persistent development data, previews and shared
  objects distinguishable from disposable test allocations.
- Set and propagate `COMPOSE_PROJECT_NAME` per stack/run, checking any overriding
  `-p` value. Disposable stacks get unique names; persistent development stacks
  keep documented stable names. Apply custom labels at the relevant service,
  network, named-volume and project-build levels. Direct Docker creators pass
  equivalent labels explicitly; they do not inherit Compose naming.
- Cover nested/helper/one-off resources and partial creations. For objects that
  cannot receive these labels at creation, including anonymous volumes and
  shared downloaded images, retain exact-ID/parent or shared-use records. Do not
  rebuild shared images or recreate active resources solely for labels.
- Make label queries the normal discovery route and reconcile them with existing
  compact ownership/cleanup/recovery records. Reuse the accepted G064 registry;
  remove redundant bookkeeping only after its recovery/evidence consumers are
  accounted for. No separate manifest, database, watcher or cleanup framework.
  Labels cannot replace persistent recovery state, verified removal receipts,
  anonymous-volume binding, or worktree/tmp tracking.
- Prove Compose and direct allocation attribution, concurrent-run separation,
  legacy/mixed ownership handling, protected live/foreign/persistent/shared
  resources, interruption recovery and cleanup through existing focused checks.
  Unknown/conflicting ownership remains a hold; project-label matches alone do
  not authorize deletion. Preserve the normal affected post-merge policy and
  broader full-run recommendation only for demonstrated impact.

The task remains queued after G064 final ACCEPT. This amendment does not change
active G064 labels, implementation, Plan, budgets or qualification. Spec/Plan,
exact path inventory and execution release remain pending. Proposed B3 remains
unspent and must be reassessed at Plan review if the actual inventory exceeds its
bounded scope; no budget extension is implied. This adds no RC8/G020/G021 gate.

[Reusable rule](../../../../skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#project-scoped-docker-labels-and-cleanup-discovery).
[Docker project-name precedence](https://docs.docker.com/compose/how-tos/project-name/),
[network labels](https://docs.docker.com/reference/compose-file/networks/#labels),
[volume labels](https://docs.docker.com/reference/compose-file/volumes/#labels).

Publication verification: source bundle, management child and public distribution package checks passed; whitespace checks passed and the disposable package copy was removed. Public Restruct `78f0fe6acfe8e9704d26bff2692c2c635752822d` is pushed to GitHub and GitLab. Documentation/Goal changes only; no implementation, Docker operation or test campaign was launched.


## Owner sequencing amendment — 2026-10-22

The owner directs the Docker-labeling work (G065/O6-5) to follow G064/O6-4
immediately, then resume the existing roadmap order. G064 final ACCEPT is the
gate before G065 planning begins. G065 needs a Codex-authored Work Spec and a
Claude-authored Implementation Plan, followed by normal Plan review, execution,
validation and final acceptance. Neither the Work Spec nor Plan has been started.
G065 final ACCEPT is now required before RC8; after RC8, the existing sequence
continues with G020, the combined staging preview, optional RC9, G057, G021 and
RC10. Any existing authorized G020 isolated implementation may continue under
its current input reconciliation and execution release, but its integration
and final acceptance remain behind the recorded G064/RC8 gates.

This supersedes the earlier Goal text saying G065 would not be a prerequisite for
RC8 or the later chain. It does not change G064's active scope, evidence, labels,
Plan, budget or qualification. No G065 planner or implementation worker is
released before G064 final ACCEPT.

## Owner early Spec authorization — 2026-10-22

The owner directed Codex to write the G065 Work Spec while G064 is still active.
This is a narrow definition-only exception to the prior timing statement: it
permits the Codex-owned Spec to be completed before G064 final ACCEPT. It does
not release Claude's Plan, implementation, Docker/build/test work or any other
execution. G064 final ACCEPT and current-input/resource reconciliation remain
the gates before G065 planning and implementation. The delivery order remains
G064 → G065 → RC8, and G065 final ACCEPT remains required before RC8.

## Owner clean-cutover amendment — 2026-10-23

The owner asked whether affected runs could simply be cleaned up before changing
labels and cutting over, then instructed Codex: "Update the spec for that".
[Work Spec r2](../specs/G065-neutral-ci-control-labels-spec.md) selects that
approach: stop affected old creators, verify exact owned resources are cleaned
up, update producers/readers together, run focused checks and resume on the new
contract. It supersedes earlier requirements for legacy-only/mixed resources,
dual-label compatibility and a later reader-retirement phase. The original Spec
is preserved as r1 (not included).

Common project/run labels, explicit Compose naming, exact-ID ownership/recovery
and protected/shared-resource handling remain in scope. Use the accepted G064
mechanisms and unchanged evidence; this does not add another cleanup framework
or compatibility test campaign. A retained or uncertain old resource holds the
cutover instead of permitting deletion or silently widening the design.

This is the selected reduced cutover design, not acceptance, execution release,
budget expansion or a roadmap reorder. G064 final ACCEPT, a reviewed G065 Plan,
the existing B3 budget and G065-before-RC8 sequence remain binding.

## Owner scope reduction — R1 + R8 only, 2026-10-23

The owner selected the coordinator's proposal to keep only R1 (the exact four-key
CI-control label rename) and R8 (verified cleanup and coordinated cutover).
Direct message `<redacted-uuid>`, 08:01:02.900Z, in the
verified coordinator session supplies the decision; the 08:01:23 handoff states
that R2–R7 are dropped and not scheduled. [Spec r3](../specs/G065-neutral-ci-control-labels-spec.md)
is authoritative. r2 (not included) preserves
the previous definition.

This supersedes the earlier wording that common metadata remains in G065.
Remove its all-resource labeling, Compose project-name rollout, compatibility
and additional resource-type validation deliverables. Existing safety/cleanup
governance remains binding without creating new G065 implementation. AC1, AC2
and AC8 cover the bounded rename/cutover; AC3–AC7 are retired. Reuse current
mechanisms, existing fixtures and attributable unchanged evidence.

The concise Plan/review, G064 final-ACCEPT gate, proposed B3 budget and roadmap
order remain. No planning, implementation or resource allocation is released
by this scope amendment. The coordinator's 2–4 hour estimate is reported only,
not a measured completion guarantee or a new budget.


## Owner start clarification — RC8 independent, 2026-10-23

The owner explicitly states: “Rc8 qualification is not a blocker.” This removes
the RC8 qualification wait introduced in the prior roadmap update. G064 FINAL
ACCEPT is complete; Claude may start the concise R1 + R8 Plan now. Following
Plan ACCEPT and current-input/resource reconciliation, implementation, integration
and final acceptance proceed through their own existing gates without waiting
for RC8 qualification. The tag and its frozen source remain independent.

This supersedes earlier sequencing statements requiring G065 before RC8 or RC8
qualification before G065. Scope, the proposed B3 budget and validation obligations
are unchanged. The Plan should give a concrete coding/check/runtime estimate
from its actual file inventory and existing selectors.


## Owner validation-timing exception — local pre-merge qualification (superseded by r6)

Historical decision: the owner stated “This can also use pre merge instead of post merge.” For G065,
run the selected affected local checks on the reconciled candidate before merge;
that evidence may replace the post-merge local run. Record the tested SHA/tree,
commands, outcomes and cleanup, then verify the actual merged SHA/tree against
them. Matching trees require no duplicate post-merge execution. If the candidate
or integration changes after testing, inspect the actual delta and rerun only
affected checks; document any reuse of unchanged evidence, including management-
only differences. Do not silently transfer results to changed executable inputs.

This supersedes G065's default post-merge placement and the review's request for
a separate merged-SHA selection. It changes timing only: keep the focused
coverage, source/hosted review, cleanup and final acceptance gates. No local full
or 528-control run is added, and RC8 continues independently.


## Namespace correction — 2026-10-23

Use Work Spec r4's product-neutral `sdlc.ci.run`, `sdlc.ci.control`,
`sdlc.ci.outer` and `sdlc.ci.owner` targets. The original approved target prefix
caused the existing `B-LITERALS` guard to fail; preserve that failure and the
r3 Spec snapshot. The guard remains unchanged. This mapping-only correction
is within the existing implementation release/attempt ledger; Claude updates
its existing Plan/Result and continues the approved focused pre-merge checks
without a new Plan round, broader rollout or full run.


## Owner RC8 harness-repair amendment — 2026-10-23

The owner directly instructed: “Fix rc8 issue with 065,” then clarified:
“Add it to G065.” This adds a bounded repair of the RC8 hosted harness defects
to G065; it does not retroactively qualify, move or replace the immutable
`0.2.0-rc.8` tag. Its hosted failure record remains FAILED.

The Work Spec r5 amendment covers three repairs from the retained RC8 diagnosis:
(1) make CI-control image listing and removal fail closed and consistent with
owned cleanup; (2) record Compose-built VC8/VC9 images by exact ID in the control
inventory; and (3) make a failed persistence start preserve per-check results,
join background checks and stop/reap the harness witness on abort. Focused
selectors prove these paths and cleanup. The registry `DeadlineExceeded` is a
separate, unproven infrastructure symptom; do not patch it with retries, frontend
pinning or timeout changes.

Claude must append a bounded Plan amendment, bind it to the retained diagnosis,
and receive Codex Plan review before writing the added repair code. Existing B3
budget and attempt count are unchanged (attempt 1 remains spent); the Plan must
confirm the remaining budget contains the amendment or stop for owner decision.
Keep MR !87 draft until the amendment, repair, focused pre-merge checks, hosted
fast pipeline, reviews and reconciliation are complete. No full local suite or
528-control campaign is added. After integration, any later RC candidate still
requires its own full hosted qualification; no release/tag action is authorized
by this amendment.


## Owner hosted full pre-merge qualification — 2026-10-23

The owner directed: “waive the local pre merge run, and just trigger
`validate:integration` on CI.” This replaces G065's local pre-merge qualification
exception. Require `validate:checks` and manually start `validate:integration`
for the same final MR SHA; together they are the selected full hosted profile.
Do not trigger the manual job on the current incomplete !87 head. After merge,
verify the merged executable trees match the qualified SHA and skip a duplicate
local post-merge run when they match. A changed executable input must be
requalified before merge; docs/management-only deltas are reconciled without a
test run.

The owner also directed closing RC8 promotion MR !86. Its immutable tag and
failed qualification remain unchanged; G065 includes the repair, and this
qualification choice authorizes no RC tag or main promotion.
