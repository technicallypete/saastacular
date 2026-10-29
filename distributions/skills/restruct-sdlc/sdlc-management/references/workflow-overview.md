# Workflow overview

## Goal lifecycle

Goal and Spec (supervisor) → Plan (implementer; omitted only when the Goal explicitly authorizes the direct tier) → Plan review (ACCEPT, CORRECTION LOOP or ARCHITECTURE DECISION REQUIRED) → implementation and validation → Result → pre-integration review → conditional integration (owner merge) → final ACCEPT → closeout. Only final ACCEPT completes a Goal.

Review-ready submissions trigger delegated review and automatic decision relay without another owner prompt. The supervisor keeps the review turn open with the harness's asynchronous wait/notification interface and relays before its final reply; do not repeatedly poll. An ended turn may not wake automatically. An interruption alone is not stop-all: inspect child status and preserve completed or authorized running work unless an explicit stop-all or conflicting direction applies. Owner interruption and mandatory gates remain.

By default (`SUPERVISOR_AUTO_PUBLISH=true`, owner may disable), the supervisor pushes each completed PROGRESS update and its management records to the explicit integration branch; the implementer keeps its task branch and merge requests. Attribute relayed owner authorizations truthfully; they never override native permission gates.

By default (`OBSERVER_AUTO_ACTIVATE=true`, a separate option), the companion milestone observer starts only after coordinator and supervisor bindings are explicitly verified; otherwise it stays inactive and reported. Effort defaults by role: coordinator high (xhigh when planning), planning workers xhigh, implementation workers high.

## Routine transitions

The supervisor records Goal/Spec priority from explicit owner direction or its own
judgment, asking when needed and normally defaulting unspecified work to high.
Plans target 3–5 parallel implementation workers for high or 1–3 for low whenever
useful independent work, ownership, native/resource capacity and budget permit;
record constraints reducing the count and exclude planner/coordinator/reviewers. The planner
may remain one agent. Kickoff verifies readiness and follows the accepted worker
allocation/sequence. Low-priority work may use one worker within its 1–3 target
and yields capacity to the main task. Keep one task owner and one writer per artifact;
each validation lane retains its own finish/fix/rerun gate, and a failure holds only
that lane and its dependents. See [parallel work lanes](coordination.md#default-to-safe-parallel-work-lanes).

Released implementers fix [mechanical inconsistencies](roles-lifecycle.md#autonomous-mechanical-corrections) within scope and budget; preserve assertions, evidence, substantive escalation and final review.

Automatic continuation stops at missing Specs unless the owner records bounded specification authority (named endpoint/range or pinned epic). Independent owner stop/phase limits override continuation even with an existing Spec or accepted Plan; do not look ahead beyond the stop. See [mandates and stops](roles-lifecycle.md#bounded-specification-authority-and-owner-stop-boundaries).

At the first eligible implementation MR (drafts included), verify it and begin provisional next-task planning with an existing Goal/Spec; owner stops remain. Planning alone grants no implementation authority. See [planning](roles-lifecycle.md#automatic-planning-after-integration).

After reviewed predecessor integration, release isolated implementation once start gates close, overlapping CI. Final predecessor ACCEPT gates successor integration/final acceptance; security/contracts and explicit holds remain. See [execution release](roles-lifecycle.md#automatic-reassessment-and-execution-release).

## Validation and closeout defaults

Feature MRs retain hosted fast checks and review; after merge the coordinator runs affected local batches/suites on the exact merged SHA. Full local integration-tip runs are owner-requested or supervisor-recommended, with their source frozen and reason recorded. Full hosted RC promotion runs remain required. These two broader checkpoints can reveal gaps in feature selection; they do not turn unexecuted checks into passes. Preserve explicit technically justified full-run requirements and all acceptance/cleanup gates.

Ordinary feature test registration, schema expectations and resulting count-pin
updates do not require full CI controls. Use affected product checks and focused
control cases under [feature control selection](validation-integration.md#ordinary-feature-registration-does-not-require-full-ci-controls);
a shared harness pathname or aggregate PASS row does not justify every control.

Validation follows [Git-first provenance and proportionate coverage](validation-integration.md): assess target deltas before batches, reconcile relevant/unknown changes and always reconcile before final/MR. Use [focused executable selections](validation-integration.md#focused-selection-before-the-first-run); broad local exceptions need a technical reason before launch. Retain required application/security/hosted checks. Existing approved bindings remain in force until explicitly amended.

For management-only acceptance/cleanup records, use [direct supervisor publication](closeout-risks.md#direct-supervisor-publication-of-acceptance-and-closeout):
coordinator handback → supervisor review/import → scoped integration-branch commit.
Do not add a Result-only MR, automated review or test run solely for closeout.
Actual source repairs and explicit owner/protected-branch exceptions retain their gates.

After verified terminal closeout, retain compact attributable evidence and purge redundant task-owned local validation output by default; protect active inputs and session archives. Default durable proof to Git records and CI artifacts; local raw evidence is temporary. Purge duplicate downloads and closed-task bulk; explicit necessary-proof holds name a release condition. Record resource ownership at creation and use [bounded routine cleanup](closeout-risks.md#bounded-routine-cleanup): one scoped pass, a compact receipt and individual holds instead of another verification campaign.

Every epic closeout requires a supervisor-owned [PROGRESS compaction checkpoint](closeout-risks.md#progress-compaction): archive superseded updates, record none needed, or document an active-binding hold and next safe checkpoint. Preserve current status and linked history; this housekeeping alone requires no test rerun.

## Continuous cleanup checkpoints — 2026-10-22

Require project/run ownership labels on exclusively owned label-capable Docker resources, whether created through Compose or direct commands. Use `COMPOSE_PROJECT_NAME` per stack and the existing compact registry for exact IDs, anonymous/shared exceptions and recovery; labels improve discovery without adding another cleanup manifest. See [Docker attribution](closeout-risks.md#project-scoped-docker-labels-and-cleanup-discovery).

Track owned resources at allocation and removal in the existing run/task inventory;
reconcile before startup/resume/dispatch, at every handback and batch end, and after
failure/cancellation. Reclaim finished worker trees and scratch immediately after
preserving their handback. Any unresolved abandoned resources block new resource-
consuming work until repaired; verify actual absence, including nested resources
and anonymous volumes. Follow [continuous accountability](closeout-risks.md#continuous-resource-accountability--mandatory).
This adds no coordinator polling or duplicate inventory framework.
