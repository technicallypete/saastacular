# Claude Code Repository Role

SaaStacular uses **Restruct SDLC** as its default development dependency, pinned
as an unaltered GitHub submodule at `.agents/skills/restruct-sdlc/` (MIT; the
repository is Apache-2.0). Claude discovers it through the relative symlink
`.claude/skills/restruct-sdlc` → `../../.agents/skills/restruct-sdlc`.
Initialize it with `git submodule update --init -- .agents/skills/restruct-sdlc`
before agent-driven work. Use the repository's pinned version instead of an
unrelated personal copy. Linux/WSL support symlink discovery; native Windows
requires symlink support.

The full workflow rules below remain in force. Application configure, build,
test, migrate and runtime commands do not require submodule initialization.
To use another workflow, remove the submodule and Claude alias and replace or
update AGENTS.md, CLAUDE.md and `docs/agents/` with your own rules; see
[Development workflow](docs/DEVELOPMENT-WORKFLOW.md).

Default role assignment (unless the owner explicitly selects the Claude-supervisor option below):

Claude Code owns implementation planning, implementation, tests/validation, and Result reporting. Codex owns the roadmap, Goal, Work Spec, plan review, acceptance, and progress tracking. Product owner + ChatGPT own canonical architecture/product direction; the product owner decides major changes.

## Post-merge staging qualification runs locally

Owner direction, 2026-10-22 UTC, superseding the blanket full-local-run default of 2026-10-19:

Default sequence for implementation MRs: focused pre-merge checks and current hosted fast/supervisor/Duo review → authorized merge → affected local batches/suites on the exact merged integration SHA. The existing implementer coordinator automatically launches, owns and tracks those selected checks; no extra owner command or full pre-merge gate. Select from the actual change and integration delta, including affected dependencies, shared fixtures, migrations, disabled variants and security boundaries. Record executable selectors and coverage rationale in the existing Plan/Result; preserve immutable inputs, actual outcomes and cleanup. Final Goal ACCEPT requires that attributable affected coverage and the other Goal gates, rather than a blanket full run per feature.

Run full local qualification on the integration branch tip when the owner requests it **or the supervisor recommends it**, recording the reason and freezing the exact observed tip for the run. Full staging runs and full hosted RC-promotion runs provide broader opportunities to find gaps in feature selection; unexecuted checks are never reported as passing. Ordinary feature MRs retain hosted fast checks; RC promotion MRs retain required full hosted qualification. Automatic integration/release-branch post-merge and tag-push pipelines remain disabled, including lightweight checks; separately authorized diagnostic/provider runs remain distinct.

After a failure, rerun the failed checks and checks affected by the fix, reusing attributable unchanged passes. A broad local selection needs a concrete coverage reason; explicit technically justified full-run requirements remain binding. Reconcile active work at a safe boundary without restarting running batches or rewriting approved snapshots. Ready authorized successor implementation may overlap post-merge checks in isolation, respecting resource reservations; predecessor final ACCEPT still gates successor integration/final acceptance. Preserve owner holds, manual/security gates, documentation-only exemptions and failure history. This supersedes the blanket full-local-run-after-every-feature default.

Ordinary application feature work does **not** require the full `ci-controls` suite merely because it updates product test registration or expectations inside harness files. This includes routes/import allowlists, migration versions, schema/catalog counts, browser matrices, test invocations, required-check IDs and their resulting count pins. Run affected product checks and only the directly affected control cases using existing focused selectors. A changed harness pathname, copied fixture, mechanical count update or desire for an aggregate PASS row is not a reason to select all controls. Full CI controls require a recorded concrete reason involving changed shared runner behavior or an explicitly selected full staging/RC qualification. Required-check aggregates apply to selected coverage; do not add `ci-controls` merely to make `B-CI-CONTROLS` required. Preserve actual outcomes and report focused control proof without claiming an unexecuted full suite passed.

See [the current policy](.agents/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#post-merge-qualification-runs-locally).

Ready, authorized successor implementation may overlap that local run **after predecessor merge**, in isolation with shared-resource locks respected. Predecessor final ACCEPT still gates successor merge/final acceptance, and its own owner retains repairs. Explicit owner/phase holds and documentation-only exemptions remain. This supersedes the earlier pre-merge/unmerged-successor variant. See [current policy](.agents/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#post-merge-qualification-runs-locally).


The merged-SHA local qualification checkout remains frozen: development-branch fetch/rebase preparation does not rebase that tested merge. Inspect later target changes separately and reconcile their actual effects; do not silently switch the qualification source or repeat unchanged coverage for management-only commits.

## Owner-selected Claude supervision

Owner direction, 2026-10-19 UTC: the owner may choose Claude Code as supervisor. In that mode, Claude owns orchestration, Plan review, implementation acceptance, progress and closeout, using the existing implementation/review workflow. ChatGPT/Codex participates **only in Work Spec writing and architecture reviews**; it is not a second supervisor or a routine approval, monitoring or dispatch gate. Major product/architecture decisions remain with the owner. Record the selected roles in the existing governance/handoff when the owner selects this mode; this option alone does not change the current assignment, restart sessions or alter active task bindings.

Owner direction, 2026-10-20: the supervisor does not poll coordinator sessions or ping the session-handoff watcher unless the owner specifically asks. The watcher already polls and delivers notices. Rely on its notices, verify linked repository records, send needed coordinator actions immediately using nonblocking delivery, and continue work without waiting for a response; the watcher delivers acknowledgment asynchronously. Receipt is not acknowledgment or execution.


[AGENTS.md](AGENTS.md) defines the shared ownership matrix, artifact formats, execution policy, budgets, delegation rules, and review workflow. Follow it in full; this file describes Claude's implementation handoff and does not create a separate governance model.

## Acceptance and cleanup closeout publication

Routine acceptance and cleanup closeout records use the supervisor's direct publication route to the configured integration branch. The coordinator hands back the Result, attributable qualification evidence and cleanup receipt with exact source references; the supervisor reviews them, records the acceptance disposition and imports/publishes the completed Result, review, PROGRESS and related closeout records in scoped commits. Do not create a separate acceptance, Result-only or cleanup-receipt MR, request another automated review or launch CI solely to publish those records. The implementer receives no direct integration-branch push or self-acceptance authority. An explicit owner instruction or an actual branch-protection requirement can select an MR instead; report that reason. Actual code or executable configuration repairs retain the normal source-MR workflow. Preserve review, evidence, cleanup and native permission gates; publication alone is not ACCEPT.

## Required reading and artifacts

Before planning or implementation, read AGENTS.md, CLAUDE.md, docs/ARCHITECTURE.md (not shipped; until the project adds one, use the README's architecture section and docs/foundation/), docs/SECURITY.md, docs/DECISIONS.md, relevant accepted ADRs, management/v0/ROADMAP.md, and management/v0/PROGRESS.md. Then read the active goal's Goal, Spec, Plan, and Result when present.

Use `management/v0/{goals,specs,plans,results}/` with the stable goal ID and descriptive stem, for example `plans/G000-project-archetype-plan.md` and `results/G000-project-archetype-result.md`. Preserve prior artifacts and review history. Do not mark yourself accepted or modify Codex-owned Goal/Spec/roadmap/progress to authorize work.

Follow AGENTS.md's **Service working directories and scoped skills**: start
service-focused sessions/tasks in the service root after reading repository
governance and approved goal inputs. Explicitly load the relevant service skill
when working from the repository root or continuing an existing session; for UI
work, follow docs/foundation/WEB-SHELL.md (UI kit source unaltered; themes through tokens). Run
root-level Compose, CI and validation commands from the repository root.

## Autonomous mechanical corrections

After execution release, fix in-scope selector/fixture/pin/catalog parity and
harness/import bookkeeping directly under [standing authority](AGENTS.md#autonomous-mechanical-corrections).
Diagnose, preserve assertions, record focused evidence in the Result and continue
without another approval/replan handoff. Escalate actual scope/security/contracts,
unapproved schema/constraints, ambiguity and budget/native stops. Final review and
required gates remain; do not restart the worker or repeat broad checks for wording.

## Workflow and autonomy

The supervisor records Goal/Spec priority: explicit owner choice first, otherwise
supervisor judgment or a focused owner question, normally defaulting to high.
Carry that priority into the Plan and target 3–5 parallel implementation workers
for high or 1–3 for low whenever independent work, ownership, native/resource
capacity and budget permit; record constraints reducing the count. These counts
exclude the planner/coordinator/reviewers. The planner may remain one agent.
At implementation kickoff, verify
readiness and dispatch the accepted Plan's ready workers, following its allocation
and sequence. Record actual deviations and constraints in the Result; retain
normal review for material changes. Low-priority tasks may use one worker within
their 1–3 target and yield capacity to the main task.
Follow [parallel scheduling](AGENTS.md#default-parallel-scheduling), with one
writer per artifact, attributable inputs/evidence, prompt reclamation, the shared
test pool, sequential builds and frozen sources. A failure holds its batch and
dependents; unrelated safe work may continue under existing gates.

Honor [recorded specification/work limits and owner stops](AGENTS.md#bounded-specification-authority-and-owner-stop-boundaries)
before any continuation, even with an existing Spec/approved Plan. Missing Specs
belong to the supervisor only under explicit bounded owner authority. Do not
plan or execute beyond a stop; clarify an ambiguous boundary before crossing it.

When a predecessor implementation MR first opens (drafts included), the supervisor
automatically starts provisional planning for the next eligible
existing Goal/Spec's planning under [the planning rule](AGENTS.md#automatic-planning-after-integration),
without another owner prompt. Acknowledge the bounded dispatch, check for an
existing planner and label pending predecessor acceptance provisional. This is
planning only; implementation and all dependency/approval/input gates remain.

Once the supervisor reassesses the provisional Plan and records execution
release under [the release rule](AGENTS.md#automatic-reassessment-and-execution-release),
begin/resume the approved task without another routine owner prompt. Isolated
implementation may overlap staging CI after reviewed predecessor integration and
closed start gates; final predecessor ACCEPT still gates successor integration/
final acceptance. Plan existence, merge or CI alone is not release. Material Plan
corrections stay Claude-owned and require review; specific holds remain.

For ordinary replies to direct owner prompts, omit the `COORDINATOR_STATUS` footer unless the owner explicitly asks for supervisor relay. Exception: an owner decision that changes active work must immediately emit a `decision` footer, including when the same message answers the owner. Do not wait for another turn, worker handback or review-ready commit. Autonomous handoffs and required reviews, failure records and blockers still notify; existing legacy-prose/native-question detection is unchanged.

For autonomous supervisor handoffs, the coordinator finishes with one plain final line:
`COORDINATOR_STATUS: working | TASK-ID | Brief current status.`
Owner-decision updates (`decision`), pending questions (`question`) and review handoffs (`review-ready`) take priority over
another task's routine work; name the actionable task. Use `working` only when no handoff is pending.
Follow the classes/limits in
[Coordinator final status line](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#coordinator-final-status-line).
This status is a report, not approval; Plan/Result submissions use `review-ready`, not self-acceptance.

For failed-batch or correction handoffs, use the
[compact repair checklist](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#compact-repair-handoff).
Reference existing Result/artifacts; send focused evidence rather than full logs.

After every completed implementation/correction round, follow
[Implementation-round safekeeping pushes](AGENTS.md#implementation-round-safekeeping-pushes):
commit/push the scoped checkpoint to the temporary task branch, even when final
review or validation is pending. Report actual status; this does not authorize
MR opening or integration. Preserve active bound files and publication exclusions.

Codex's final pre-integration code-review acceptance authorizes pushing the
reviewed feature branch and opening its MR to staging without another owner
permission request. Pending review does not block approved implementation;
merging still requires its separate owner authorization. Follow AGENTS.md's
standing authority and the existing validation/reconciliation requirements.

Prefer task/role-prefixed worker names when safely supported, following
[AGENTS.md](AGENTS.md#task-prefixed-claude-worktrees). Owner direction, 2026-10-14:
unsupported naming is a Claude harness bug/limitation, never a dispatch blocker.
Use the native isolated worker and its generated `agent-<id>` directory without
another naming exception or permission question. Record task/role/agent/path/
branch/HEAD and verify actual ownership/isolation. This supersedes the earlier
mandatory-prefix and task-specific exception requirements. Do not rename live
workers, create duplicate sessions or bypass native permissions for naming.

The single coordinator prefers `0-agent-<id>` when supported; safe relocation
checks remain. Codex normally supervises from the integration checkout and puts
isolated workers under `.codex/worktrees/`. Submit artifacts through the existing
coordinator; Codex reviews/relays decisions with all other gates intact.
Local session handoffs follow [the runbook](docs/agents/CLAUDE-SESSION-HANDOFF.md).
Acknowledge actual receipt and report worker/effort/checkouts; queued messages
still do not waive Plan approval, budgets or input bindings.

1. Read the Goal and Spec and inspect the repository. Unless the Goal explicitly authorizes direct execution under AGENTS.md's risk-based policy, write the goal-specific Plan using the AGENTS.md format. Include files/modules, schemas/migrations, interfaces/contracts, tests, sequence, delegation, every acceptance criterion, and escalation risks; scale detail to the actual change.
2. If the goal exceeds reliable available context, recommend a bounded split in the Plan and stop before implementing. Wait for Codex review.
3. Stop for explicit Codex Plan ACCEPT before implementation unless the Goal already authorizes the direct-execution tier. That tier needs no separate Plan or additional Plan-approval round; it still requires bounded scope, validation, Result and Codex implementation acceptance. Do not select a lighter tier yourself.
4. Implement the approved plan. Autonomously inspect, create/modify tests, run builds/lint/type checks/tests, diagnose, iterate, and refactor within approved scope.
5. Run all required validation and self-review architecture/security compliance. Record the Result using the AGENTS.md format, including exact command outcomes, acceptance evidence, deviations, blockers, attempted approaches, and cycle counts.
6. Return control to Codex for exactly one review outcome: ACCEPT, CORRECTION LOOP, or ARCHITECTURE DECISION REQUIRED. Only Codex may accept completion and update progress.

Before every local test/validation batch on the temporary feature branch, follow [GitLab branch integration and cleanup](docs/agents/GITLAB-BRANCH-WORKFLOW.md): fetch the remote target, rebase the temporary worktree source if needed, resolve conflicts and reconcile approved inputs before launching tests. Verify the fetched target SHA is an ancestor of the candidate HEAD. After required manual feedback and self-review, repeat that readiness check before pushing and creating the MR. The default affected local qualification runs after merge on its frozen merged-SHA checkout; do not rebase that checkout under development-branch rules. Record target/source SHAs; do not substitute a stale local target or carry old pipeline success to a new commit without reconciliation. Keep long-lived branches intact and follow the temporary solo-operator validation policy below.

Codex assesses staging advances against all in-flight tasks, including planning.
Acknowledge targeted pause/reconciliation instructions. Before rebasing an
interrupted validation checkout, stop/reap owned children, preserve interrupted
evidence, verify owned cleanup and release the run lock. Reconcile bindings and
rerun affected checks; do not restart unchanged coverage solely for a new SHA.
Follow [the integration assessment rule](AGENTS.md#codex-assessment-of-staging-advances).

At Plan handback, provide the exact saved Plan, required inventories and
revision-history paths with their SHA-256 values so Codex can import accepted
planning artifacts into the local integration branch as required by AGENTS.md.
Keep the planning/execution checkout's bound inputs unchanged during that import.

## Effort reminder

Follow [model-aware effort selection](AGENTS.md#model-selection--owner-direction-2026-10-15):
Default to Fable 5.1 high for planning and test corrections; reserve its xhigh
for especially difficult work. Default to Opus 5.5 high for implementation and routine test execution/reporting.
Default Claude coordinator model is Sonnet 5.5 high; owner may choose Opus 5.5 high.
If the owner chooses Opus for planning or test corrections, use xhigh. Routine
implementation and test execution/reporting stay high; coordinator choices stay high. Configure through supported controls and verify actual
or inherited values, otherwise report unknown. Honor existing owner choices;
prose does not set runtime configuration. Apply at the next phase/safe boundary;
never interrupt a batch, redo completed work or reset budgets solely for effort.

Fable may raise problematic test corrections from high to xhigh at a safe boundary
without repeat owner approval; record the reason/actual setting and preserve budgets.

## Budgets and escalation

Follow AGENTS.md's **Proportional planning and reuse of accepted controls**.
Inspect relevant existing tests before planning new ones. In the existing Plan's
validation/AC mapping cite test paths/check IDs as reuse, extend or new, and state
the uncovered behavior or distinct boundary for each new test. Prefer extending
existing fixtures; no duplicate suite or fresh check solely for this Goal. Retain
required regressions/CI and current evidence reconciliation.
Follow its **Risk-based execution for future Goals**: direct execution for
explicitly authorized bounded presentation changes, concise Spec/Plan for
ordinary features, and explicit review/stronger evidence for security,
architecture and data changes. Existing Goal bindings/gates are not waived.
Its **Keep CI changes separate and bounded** subsection is the default for
future Goals: reuse accepted CI, add focused tests/minimum mapping changes,
and scope substantial runner/harness work separately. Preserve declared
dependencies and required merge/release gates.
Follow AGENTS.md's **Validation batches: finish, fix, rerun, then advance**:
let a safe batch finish collecting failures, preserve results and verify cleanup,
fix outstanding failures in that dependency chain, rerun failed cases/controls and affected checks, then advance.
Do not start a dependent batch with unresolved prerequisite failures or repeat unrelated passing
batches without changed inputs. Unsafe/invalid work stops immediately. Preserve
pre-batch reconciliation, budgets and required final coverage; track findings in
the existing Result.

Follow AGENTS.md's **Proportional UI validation and focused debugging**: use
focused reproduction/retests while diagnosing UI failures, then collect the
required final evidence. Scope future UI validation to affected behavior; do
not inherit G048's broad matrix automatically. Existing approved matrices and
CI triggers remain binding.

Follow AGENTS.md's **Manual testing and operator feedback before final CI** in
every Plan that needs either: state readiness/evidence requirements, complete
focused local checks, obtain required manual results and operator feedback,
resolve findings, then collect required local coverage before the MR and hosted
staging qualification after conditional integration. Bind evidence to
tested/reviewed inputs and repeat affected tests/reviews after relevant changes.
Identify any CI artifact/environment prerequisite explicitly; retain mandatory
final checks and existing bindings. Required feedback pending means the candidate
is not ready for final qualification.

Follow AGENTS.md's **Temporary solo-operator validation and integration policy**:
fetch/rebase/reconcile before every local test/validation batch, including focused
and preliminary runs. On Docker Desktop/WSL, also run
`bash infra/ci/docker-desktop-mounts.sh` first and recover before the batch if it
fails; do not add per-run bind mounts to harness code. Finish implementation, required manual feedback and self-
review; refresh/reconcile again before required final affected/full local coverage
and pushing/opening the MR. Read-only safety/binding inspections may precede the
rebase; preserve bound inputs and stop if safe reconciliation is unavailable. If the target advances again,
reconcile and rerun only affected checks, retaining unchanged evidence with exact
content/input provenance. Do not repeat unchanged local checks. MR fast runs automatically; its integration/full gate is
optional/manual. The existing path-based staging gate supplies required hosted
evidence after an explicitly authorized conditional merge. Codex reviews before
integration and uses the owner staging status rule: success/canceled satisfies
the CI acceptance gate; any failure fails, and pending/running does not qualify.
Record cancellations honestly; local evidence and other Goal criteria remain.
Preserve input bindings, distinct evidence, budgets and G021; record this owner
amendment instead of altering approved historical Plan/snapshot bytes.

Reference and reuse accepted safeguards; keep narrow Plans focused on their
actual changes and tests. Describe implementation-time validation before
approval, then build/run it during authorized execution. If satisfying a
correction appears to require a new execution framework or substantial process
engineering, report the smallest alternative to Codex before expanding it.
External scratch code does not make that expansion free or implicitly authorized.

Default budget: **5 goal-level execution attempts**, unless the reviewed Goal explicitly sets another budget. Scoped implement/test/fix loops and approved subagents share one attempt; routine failures do not each spend an increment. Apply management/BUDGET_POLICY.md attempt boundaries and preserve historical counts. Active Goals need explicit supervisor reconciliation before changing their accounting. If materially failing afterward, stop, update the Result with blockers and attempted approaches, and return control to Codex. Do not weaken validation, silently extend the budget, or hide skipped/failed checks.

Stop affected work and escalate architecture contradictions, scope expansion, security-model changes, destructive/risky migrations, public API/MCP contract changes, new major infrastructure/vendors, irreversible data-model decisions, and conflicts with architecture or accepted ADRs. Identify planned changes in the Plan for explicit review; newly discovered changes require a further handoff. Major architecture/product decisions go through Codex to the product owner. An ADR proposal is not approval to implement it.

For CORRECTION LOOP, fix only Codex's bounded requested issues and rerun required/relevant validation. Preserve previous Result evidence and append the correction record. Do not autonomously add follow-up features. Track correction passes; if the same material issue remains after two cycles, return it to Codex for Goal/Spec reassessment or architecture escalation rather than continuing indefinitely.

## Delegation

For feature/Goal integration, follow [GitLab branch integration and cleanup](docs/agents/GITLAB-BRANCH-WORKFLOW.md): use glab MRs to `staging` with source-branch deletion, verify cleanup after merge or MR closure under the runbook's preservation checks, and retain the long-lived qualification/integration/release branches. An MR does not waive Codex acceptance or authorize a merge on its own.

Consider Codex's suggested roles, then decide tactically in the Plan whether subagents help. Use them for exploration, test analysis, security review, library research, isolated non-overlapping implementation, or post-implementation review. State file ownership and integration strategy. Do not delegate product/architecture authority, scope changes, overlapping edits, or final integration responsibility. Validate and integrate all delegated work yourself.

## Architecture constraints

- `core` owns facts, authoritative domain logic, authorization, MCP, and public streaming; `intelligence` owns internal inference; `web` remains a thin UI/BFF boundary.
- Identity is global; Organization is the tenant boundary, Group the collaboration scope, and BusinessEntity reusable business context. Use shared RoleBinding policy and server-resolved context with PostgreSQL RLS defense in depth.
- Dynamic MCP discovery follows active Organization/Group permissions; each execution reauthorizes independently. AI never determines tenant scope or authority.
- PostgreSQL is durable state; Redis is ephemeral; MinIO is local S3-compatible storage. External providers remain behind owned adapters.
- Project runtimes and validation are Docker-first with no host Go/Python/Node required by default; preserve configured standard+100 host port conventions.
- Product/domain names and client-facing role labels resolve through configuration; preserve semantic branding tokens and non-redundant service folder names.
- AI uses `AI_MODEL_<CAPABILITY>=<provider>:<model>` and an OpenRouter-first ModelGateway boundary, without making the gateway a domain concept.
- Preserve the G000–G021 foundation barrier before product-specific work. Follow the architecture for full requirements; this list does not replace it.
- `web` has no authority, policy or domain rules and no `NEXT_PUBLIC_*` backend URLs; `intelligence` stays internal with no data-store access.
- Resolve the tenant (the Organization) on the server from the authenticated session and the request host, never from client-supplied or model-supplied identifiers.
- Keep product name, domains and database name in `.env` and `.env.example` only. Configure with `infra/local/configure-env.sh`; it never overwrites `.env`. Never commit `.env` or other secrets.
- Leave the UI kit source in `web/src` unaltered; change themes through the tokens in [docs/foundation/WEB-SHELL.md](docs/foundation/WEB-SHELL.md).
- Run everything in Docker from the repository root with the README's `platform:*` blocks and `bash infra/ci/validate.sh --profile fast|integration|full` (or `affected --base <commit>`); developer checks need a Git repository. Check IDs and suites are indexed in [docs/foundation/TEST-INDEX.md](docs/foundation/TEST-INDEX.md).

## Future-goal validation exception — 2026-10-14

For newly planned work from G009, apply AGENTS.md's exact-delta exception before
rebasing for irrelevant staging changes; final/MR reconciliation remains mandatory.
Use Git-bound inputs and proportionate coverage. Existing G008 approvals remain.


## Fable weekly-limit fallback — owner direction, 2026-10-15

When Fable's weekly allowance is confirmed exhausted, automatically use Opus 5.5
**xhigh for planning and test-failure diagnosis/corrections** until Fable resets.
Opus implementation and ordinary test execution/reporting remain **high**.
Coordinator default remains Sonnet 5.5 high, with Opus 5.5 high available by
owner choice. This standing fallback needs no repeat owner approval; honor an explicit
owner override. Record the limit, fallback model/effort and actual setting (or
unknown) in the existing Result/dispatch record. A generic tool error is not
proof of a model limit.

Use supported controls at the next safe boundary, preserving the existing task,
artifacts, checkout, bindings, reviews and attempt count. Do not interrupt a
running validation batch, redo completed work, reset budgets or weaken gates.
If native switching is unavailable or refused, report the blocker once; do not
bypass permissions. Once Fable's allowance reset is verified, return to its normal
model/effort defaults at the next safe phase boundary, without restarting work or
repeatedly probing the allowance. This is model selection only, not scope,
execution-release or budget-extension authority.


Use [low-overhead delivery defaults](AGENTS.md#low-overhead-delivery-defaults--owner-direction-2026-10-15):
focused checks and bounded failure handling, automatic authorized transitions,
concise handoffs and the existing cleanup receipt. Preserve active batch/bindings
and required staging qualification; do not request another routine proceed prompt.

Owner note, 2026-10-16: assess coordinator history at task handoffs/closeout and
when old messages dominate context; use [safe context compaction](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#coordinator-context-compaction-at-safe-boundaries)
with current task/decision/evidence references preserved. Keep the same session,
watcher binding and safe running batches; no model switch, transcript purge or new
monitor is implied. Context size alone does not attribute allowance consumption.

Use [targeted reruns and attributable coverage](.agents/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#targeted-reruns-and-attributable-coverage-default)
as the standing default: preserve attributable passes, rerun failures/dependencies,
and prepare frozen source for concurrent review. Do not repeat whole control suites
solely for a fresh wrapper PASS; final evidence, cleanup and hosted gates remain.

Owner clarification, 2026-10-16: apply [bounded worker context handoffs](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#worker-context-and-phase-handoffs)
when oversized obsolete history warrants them; preserve one write owner, exact
state/evidence and phase-specific effort. No mandatory fresh agent per phase,
full retest, budget reset or new tracking system. Report owner-only native
compaction/effort controls once; policy text does not perform those actions.

Owner default, 2026-10-16: worker checkpoint target is **150,000 active context
tokens**, with an upper handoff threshold of **200,000**, under
[the existing context policy](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#worker-context-and-phase-handoffs).
Prepare concise transfer state around 150k at a safe boundary; checkpointing alone
does not force immediate agent rotation or another review.
At a safe step/batch boundary, checkpoint and transfer concise state to a fresh
worker; no full-transcript fork, duplicate writer, retest or budget change solely
for rotation. Record a brief completion-imminent exception or unavailable context
measurement; coordinator identity/compaction remains separate.

Owner direction, 2026-10-16: [focused executable checks from the first run](.agents/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#focused-selection-before-the-first-run)
are the default for all local implementation/correction/final validation. Verify
what selectors actually run before dispatch; broad umbrella suites require a
recorded technical coverage reason and supervisor disposition before launch.
Existing reviewed justification suffices; no repeated owner permission loop.
Convenience or a missing narrow harness selector is no automatic broad fallback.
Reuse attributable evidence and preserve required hosted/atomic/security gates.

Use [early risk review and short correction loops](.agents/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#early-risk-review-and-short-correction-loops)
within the existing Plan: inventory dependent callers/pins, prove the highest-risk
approved slice early, and return consolidated findings with minimum affected checks.
No extra phase agents, broad reruns or restart of active work is implied.

Owner direction, 2026-10-23: use the existing Plan to separate settled decisions
from unproven assumptions and name a small early proof of the riskiest assumption
in approved implementation. Give executable selectors, dependency/resource order
and estimates separating coding, build/test elapsed time and review/queue waits;
mark unknowns honestly. Material scope additions need an approved bounded impact
delta. A failed proof or material overrun needs a compact Result checkpoint and
correction. Apply to remaining work at a safe boundary, without restarting work
or creating a new timing/reporting process. Follow the linked early-risk policy.

Owner direction, 2026-10-16: [coordinator checkpoints and retirement](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#coordinator-checkpoints-and-retirement)
uses a 150k active-context checkpoint and 200k compaction-or-handoff threshold,
not a hard kill. Prefer compaction; retire at a safe boundary for persistent
context/decision/session problems. Replacement needs recorded creation authority,
verified native cutover, one coordinator and observer rebinding. No live retirement
or automatic session-launch implementation is authorized by this documentation.

Owner clarification, 2026-10-16: the session lifecycle policy is effective now,
not deferred to G055. Apply [supervisor continuity](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#supervisor-session-continuity)
to Codex as well, with actual native/model limits and compaction first.
Native context limits, security stops and acceptance gates remain. This records
policy, not an automatic launcher or authority to bypass existing session-creation checks.


## Continuous resource accountability — owner direction, 2026-10-22

Use explicit project/run ownership labels for every newly created exclusively
owned Docker object that supports them, through Compose and direct commands.
Set `COMPOSE_PROJECT_NAME` per stack/run; direct creators need equivalent explicit
labels. Discover by labels, then verify exact identity/lifecycle using the existing
compact run registry. Preserve anonymous/shared exceptions and reboot recovery;
no extra cleanup manifest, label-only deletion or relabeling of active bound work.

Cleanup is non-negotiable throughout execution, not only at final acceptance.
Record every container/image/volume/network, worktree and temporary path in the
existing task/run inventory at allocation, including nested/partial creations;
update ownership and removal as they happen. Reconcile actual resources before
startup/resume/dispatch, at each worker handback and batch end, and on failure,
cancellation and acceptance. Reclaim finished worker trees and disposable scratch
as soon as the handback and unique source are preserved. Stop/reap creators before
cleanup and verify absence afterwards. Failed cleanup blocks new resource-consuming
work until reconciled; legitimate live/required retention must have a named owner,
reason and release condition. A crash/reboot requires recovery reconciliation;
traps alone cannot ensure cleanup. Report all resource categories and unresolved
counts in the existing receipt. Preserve credentials, native session state,
required minimal failure proof and unrelated/shared resources. This is event-driven
accounting, not coordinator polling. Repair and focus-test executable leaks before
resuming the affected workload; prose alone is insufficient. See
[continuous resource accountability](.agents/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#continuous-resource-accountability--mandatory).

## Mandatory task-acceptance cleanup — owner direction, 2026-10-17

At every final task ACCEPT, immediately remove all finished task-owned worktrees
and temporary resources under the [mandatory cleanup policy](.agents/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#mandatory-cleanup-at-every-task-acceptance).
Unlock finished agent-tool worktrees; preserve unique commits and dirty/untracked
work before force removal when needed. Clean task-owned containers, unused images
(including attributable untagged IDs), volumes and disposable evidence without
another routine approval. Keep active coordinator/successor/preview resources and
specific recorded retention/native-permission holds. Verify a compact receipt;
no silent lock/dirty-work deferral, blanket host prune or extra test campaign.


## Private files for new tasks — 2026-10-17

Use the main integration checkout's ignored `.local/{evidence,secrets,tmp}/` for
new task-private files, shared across worker worktrees. Follow the [local storage
policy](.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#repository-local-private-files): task-scoped paths, private permissions, Git/build/upload exclusions
and separate evidence/secret cleanup. Existing tasks keep current paths through
acceptance cleanup; do not migrate active bindings or native session state.

## Repository-scoped Claude permission rules — 2026-10-18

Follow [permission placement and verification](docs/agents/CLAUDE-PERMISSIONS.md):
tracked shared `.claude/settings.json`, ignored repository-specific personal
`.claude/settings.local.json`, home settings only for intentional cross-project
preferences. Verify actual native loading in worker/service launch directories;
permission rules never replace task/review/merge authority or native gates.
Distinguish those tool rules from [classifier configuration](.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#tool-permissions-and-classifier-rules):
use its native required scope, preserve owner-selected defaults, and verify actual
approved operations. Machine/task exceptions and disabled classifier rules are
not portable adoption defaults.

## Status writing — mandatory short default, 2026-10-18

Routine owner-facing updates MUST be at most 80 words and three short bullets
(or one short paragraph), plus a status marker only when required for a supervisor handoff. Report only the
new outcome, next action and any decision/blocker requiring attention. No verbose
reports, task rosters, repeated history, hashes, policy recaps or unchanged gates.
Routine acknowledgments, dispatch receipts and unchanged waiting states get no
separate report. Put evidence, inventories and full handbacks in the existing
Result/review and link it; do not paste them into chat. Detailed chat is allowed
only when the owner explicitly asks for detail. A failure/decision must remain
clear within the limit; do not hide it in a link. Preserve truthful uncertainty,
source/final-ACCEPT distinctions and the exact status-marker syntax. This hard
default supersedes earlier guidance to preserve detailed status prose or avoid
fixed-length reports; it changes no technical, native-permission or review gate.

See [operational messages](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#concise-operational-messages).


### Isolated unit checks do not take the shared harness lock

Owner direction,2026-10-18: remove blanket serialization of all Docker commands.
Focused unit/static checks run without local-harness.lock when they use read-only
source, disposable per-run writable temporary/build state, no database/Compose
stack, no published ports and no shared mutable fixtures. Existing dependency
caches may be mounted read-only after preparation; Go uses -mod=readonly,
GOPROXY=off and a disposable GOCACHE, with test containers network-disabled.
Do not write shared caches concurrently: dependency downloads/cache preparation,
Compose/integration/browser stacks and shared ports/data/fixtures retain their
existing resource lock. Keep bounded resource ownership, cleanup and input binding.

Use the native plain Docker command in the verified task checkout. This owner
change supersedes the earlier G019 blanket-lock instruction for qualifying checks;
there is no requirement for the owner to execute each unit-test command. Native
worktree isolation remains; no hidden wrapper or different executor is needed.
Run the approved focused selectors, record actual outcomes, and correct failures
within scope. A new native refusal is reported precisely rather than prompting
another all-tests ban. This changes resource scheduling, not test expectations,
security/tenancy invariants, task scope, budgets or acceptance requirements.


## First failure requires intervention — 2026-10-18

Each required GitLab MR pipeline has one explicitly named monitoring owner,
normally the coordinator, recorded with the MR/pipeline ID and immutable source
SHA when the MR is opened or the pipeline is created. The owner verifies the
expected pipeline started, tracks it to a terminal outcome, reports the outcome,
and routes the first failure for diagnosis. The Claude session watcher observes
coordinator-session notices; it does not monitor GitLab MRs, pipelines or jobs.
A general task assignment, watcher notice, review request or auto-merge setting
is not evidence that a pipeline has an active monitor. If ownership changes,
record a direct handoff and acknowledgment so monitoring has no gap. Keep updates
event-driven and concise; do not send repeated unchanged status polls. Auto-merge
does not replace monitoring or alter review/merge authority. Before another
CI-triggering push, the owner checks the previous outcome and records the
pipeline, SHA, failed job/test and correction owner in the existing Result.
The supervisor acts on a reported blocker/failure promptly instead of waiting for
a final handback. A failed CI run is executed-and-failed evidence even when local
execution was blocked; never summarize that state as tests not executed.

On the first failure, inspect its available log and run the smallest reproducer.
Do not repeatedly push or retry unchanged known-failing source to rediscover the
same error. The next validation run must test a specific correction or an evidenced
retry hypothesis within existing authority. Keep independent work and safekeeping
authority; explain an unavoidable checkpoint-triggered run rather than treating
it as another diagnostic attempt. Do not suppress required CI for application
changes or erase failed evidence. Existing MR supersession handles redundant runs.

On the first native permission refusal, separate configuration/resource scheduling
from product security: report the exact refused operation and propose the narrow
supported remedy immediately. Codex uses existing owner authority to correct
workflow/configuration where authorized; ask only for genuinely missing authority.
Verify an actual successful tool call. Do not repeatedly dispatch blocked workers,
require owner-run commands by default, or compensate with repeated CI runs.
Keep isolation and real security/permission boundaries; this is prompt diagnosis
and supported configuration, not permission to evade a native refusal.


## Claude requests permissions; Codex installs; Claude verifies — 2026-10-18

Owner standing direction: Claude reports exact missing permission rules and the
blocked operation/task justification to Codex. Codex must review and install the
narrow rules needed for already-approved work, choose the appropriate supported
settings file, preserve existing settings and validate the delta, then notify
Claude of the exact file/rules installed. Do not require another routine owner
approval or leave approved-scope rules for the owner to paste.

Claude must verify native loading and the actual approved edit/test/commit before
reporting the block cleared and continuing. Installation, loading and operation
success are distinct. Follow the [permission handoff](.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#permission-change-handoff).
This standing authority covers bounded allow rules, not unrequested global guard
removal, permission-mode changes, broader task authority or bypassing a remaining
native refusal. Keep the current task, budget and single write owner.


## Usage correction — owner direction, 2026-10-18

Enforce [existing context and coordination limits](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#enforce-existing-usage-limits)
at safe boundaries: one implementation worker by default, justified bounded nested
delegation, delta-only continuation reads, no unchanged-state acknowledgment or
repeated closed-worker inventories. Keep current indexes concise and preserve
history in archives. Required security/CI proof and native gates are unchanged.

Local validation also follows [change-impact selection before merge](.agents/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#change-impact-selection-before-merge): map applicable shared fixtures, migration paths, pins, disabled variants and concurrency effects to focused checks in the existing Plan. Reconcile material deltas; no new broad-run default or duplicate planning artifact.

Owner direction,2026-10-18: Codex supervision with a Claude implementation coordinator remains the supported delivery baseline. Live workflow changes retain their existing release gates.

UI work reads `docs/foundation/WEB-SHELL.md` relative to the Git repository root (use that absolute root when starting in a service directory). No agent skill is needed to work on the UI; the UI kit source in `web/src` stays unaltered.


## Owner-selected RC promotion auto-merge — 2026-10-19

The owner may explicitly enable auto-merge for a named RC promotion MR or a clearly recorded scope of RC MRs. Manual promotion remains the default when no such instruction exists. This owner instruction authorizes queuing promotion before CI finishes; once every required full-qualification job succeeds, the provider may merge without another owner prompt. Record the owner instruction, MR, immutable source SHA and selected mode. Before arming, verify the tag/source identity, required complete pipeline, non-squash merge, and conflict-free prospective target tree equal to the RC tree; source-pin the request. Preserve target-movement monitoring and post-merge parent/tree verification without claiming an atomic target lock. Retain and independently verify full project/pipeline/job/artifact proof before claiming qualified release. No failed, skipped, optional or partial coverage counts as qualification. Do not fabricate manual-mode proof or change API snapshots to satisfy a verifier. A manual-only helper's refusal of queued/merged MRs is a mode limitation, not a passing auto-mode proof; use truthful independent verification or a separately reviewed mode-aware adapter. Final release tagging remains a separate owner decision. Feature-MR auto-merge defaults alone never authorize an RC-to-release merge.

When the owner authorizes a named RC series or directly requests its next
candidate, that authority covers the coordinator's bounded series work without
per-step or per-candidate reconfirmation: in-scope qualification repairs,
normal staging integration, the next sequential immutable RC tag from the
latest staging commit observed at cut time, its tag-pinned promotion MR, and
required full qualification. Preserve every failed run; never waive or report
it as passing. An owner instruction given to the supervisor remains valid when
faithfully relayed to the coordinator; do not ask the owner to repeat it in
another session. Scope, budget and escalation stops remain binding. This does
not authorize promotion-MR merge, auto-merge without its separate owner
instruction, a stable release tag, or deployment/publication; those remain
separate owner decisions. See the [RC series authority rule](docs/agents/GITLAB-BRANCH-WORKFLOW.md#standing-authority-for-an-owner-authorized-rc-series--2026-10-21).

Make every new annotated tag message and GitLab Release entry descriptive of its
contents: list included Goal/task IDs and short titles, material changes since
the previous candidate/release, and the exact source commit. Keep canonical
SemVer tag names. Final release notes summarize delivered scope since the prior
stable release and state qualification truthfully; do not list planned or
unaccepted work. Published tag objects are immutable; improve sparse GitLab
Release notes without moving or recreating the tag. See the
[tag/release description policy](docs/agents/GITLAB-BRANCH-WORKFLOW.md#descriptive-tag-and-release-notes--owner-direction-2026-10-20).


## Optional RC Duo review and planned follow-ups — 2026-10-19

RC promotion MRs may use Duo as optional advisory review; Duo is not a required promotion approval gate. Triage every actionable finding and record its disposition in the existing release Result/review. Changes arising from that review must be scheduled and planned as bounded follow-up work: identify the responsible Goal/task or backlog item, intended milestone, scope, acceptance criteria and required focused validation under normal authority. Do not leave actionable recommendations only in an MR comment or silently claim them fixed. Do not change the code or move/recreate an existing RC tag to satisfy a review. If a follow-up changes the release candidate or must ship before promotion, integrate it through the normal staging workflow, then cut and fully qualify a new RC tag under the applicable release authorization. The owner may instead retain a documented prerelease follow-up while the current RC proceeds, as explicitly selected for RC1's limiter finding. Preserve ordinary feature-MR Duo requirements, required full RC CI, security/owner dispositions, selected auto/manual merge mode and the separate final-tag decision. This policy creates no automatic RC2, unbounded repair mandate or new tracker.


## Owner-waived canceled RC qualification — 2026-10-19

The owner may explicitly cancel a named RC promotion MR's pipeline and direct that cancellation to satisfy the current RC promotion gate, treating it as success for the workflow decision only. Record an owner-authorized qualification waiver with the RC/tag/source SHA, MR/pipeline IDs, actual canceled/partial/failed job states, any known findings and the owner's instruction. Never report the canceled pipeline or unexecuted tests as PASS, fabricate green evidence, erase prior failures or claim complete qualification. This exception is specific to the named candidate/pipeline; it is not the default cancellation policy for feature MRs or other RCs. Schedule a follow-up immutable RC tag to run the complete required suite, carrying forward unresolved findings and the qualification debt in the existing release Result/backlog/Plan. The follow-up RC's full suite is required unless the owner separately changes that instruction. The owner may choose this waiver alongside the recorded auto/manual promotion mode; provider auto-merge still depends on the provider's actual cancellation behavior. If existing provider rules prevent merging a canceled pipeline, report that concrete limitation rather than changing project-wide protections, manufacturing a success job or treating a queue as a completed merge. No pipeline is canceled merely because this option is documented, and final release tagging remains separately owner-controlled.


## Reviewed-merge classifier reconciliation — 2026-10-19

Owner requested narrowing the Merge Without Review false positive. Follow
[reviewed merges and classifier false positives](.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#reviewed-merges-and-classifier-false-positives):
verified independent supervisor review under delegated merge authority satisfies
the review requirement without another routine human approval. Required CI,
source-SHA binding, applicable findings disposition and separate RC owner
authority remain. Read-only follow-up/authorized closeout is not a merge. Native
exceptions require owner authorization and supported configuration; installed,
loaded and operation-succeeded remain distinct. No blanket disabled-guard default.


### Private per-run database tests — owner direction, 2026-10-19

Private database test runs do not take local-harness.lock for any task when each
run owns its database, network, credentials and disposable storage; publishes no
host ports; uses read-only source and prepared read-only dependency caches; and
shares no writable fixtures/cache or existing stack. Use unique run labels/names,
record exact resource ownership before removal, and verify bounded cleanup.
Use plain supported native Docker commands in the bound checkout. Network access
is limited to the private test services; no unrelated host/project services.

Allow at most **two concurrent private database runs across this repository's
coordinator and workers**. The owner may explicitly approve a third run when
asked; no automatic third slot or wider concurrency. Use the existing coordinator
and run inventory for slot ownership; verify live runs before dispatch and release
slots after cleanup. If ownership/count cannot be established, queue the run.
No new scheduler, lock service or monitoring framework is required.

This supersedes the database-stack lock requirement only for qualifying private
runs. Shared image builds, dependency/cache preparation, Compose/browser stacks,
shared ports/data/fixtures and other shared mutations retain their locks. A test
needing unavailable pinned images waits for their approved preparation; do not
substitute an unpinned image or claim unexecuted coverage. Native guards, required
validation, exact-source evidence, task scope, budgets and release gates remain.


### Private browser runs and sequential builds — owner direction, 2026-10-19

Extend private-run scheduling to browser tests, including their fully private
Compose stacks. Database-only runs and browser stacks share **one repository-wide
pool of two concurrent test runs** across coordinator/workers; a third requires
explicit owner approval. A browser stack's internal database belongs to that
single run, not another slot. Existing no-third-without-approval rule remains.

Each run owns its project, containers, network, database, credentials and writable
storage/fixtures; mounts source and prepared caches read-only; and reuses no live
stack. Browser ports, if required, are individually reserved nonconflicting
loopback ports, never shared fixed ports. Use unique names/labels, recorded slot
and resource ownership, exact tested-source attribution and verified cleanup.
Qualifying private browser/Compose runs do not need local-harness.lock. Shared
stacks, shared/fixed ports and shared mutable fixtures retain their existing lock.

**Builds are sequential: one image build at a time across this repository's
coordinator and workers.** Reserve one build slot in the existing coordinator/run
inventory, verify no owned build is active, and queue additional builds. Individual
images of a multi-service build must not be built concurrently. A private build
may use a plain native Docker command without the harness lock while holding this
exclusive scheduling slot: use task-owned tags, a bound build context and disposable
outputs; do not overwrite active image tags or prepare a shared writable cache.
Shared cache preparation and other shared mutations still use their existing lock.
A third test-run approval does not increase build concurrency. Release reservations
on completion; queue on uncertain ownership. No new scheduler or executor.

This supersedes the earlier blanket browser/Compose/image-build lock requirement
only for these private, reserved operations. Existing safe runs keep ownership;
reconcile at their next safe boundary without restart or budget reset. Native
worktree checks remain enabled: no wrapper, mode change or alternate executor to
evade a refusal. Test selection, required coverage and review/acceptance gates
remain unchanged. Missing support or unsafe resource collisions remain blockers.


## Reusable project test permissions

Owner-authorized private test commands use a stable project test prefix/label,
shared build/run/Compose/tagged-image cleanup allows in tracked project settings,
and machine-specific evidence Write/Edit paths in ignored local settings. Follow
[reusable test permissions](.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#reusable-private-test-permission-rules).
Preserve existing provider/GitLab rules. Keep active task-specific bridge rules
through safe transition/cleanup; permanent rules remain. Verify loaded sources and
actual operation. Prefix/allow matching does not prove task ownership or waive
native checks, test concurrency/build sequencing, scope or validation.


## Context and coordination economy — owner direction, 2026-10-20

Current checkpoint/handoff defaults are150k/200k active context, superseding300k/350k at the next safe boundary for workers and coordinator/supervisor assessment. Claude and Codex may initiate supported native self-compaction at safe points under standing owner authority. Preserve running batches, exact bindings, children, budgets and smaller native limits. Follow [coordination economy](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#lower-context-thresholds-and-coordination-economy--owner-direction-2026-10-20): directly review routine prose/mechanical changes, reserve delegation for meaningful independent work and required substantive reviews, batch searches/edits/handoffs, read deltas and suppress unchanged polls/acknowledgments. No full-history fork, duplicate worker, test rerun or automatic replacement merely for a checkpoint; no weakening of security/source/validation gates.


Owner amendment,2026-10-20: [safe self-compaction](.agents/skills/restruct-sdlc/sdlc-management/references/coordination.md#authorized-self-compaction-at-safe-points--owner-amendment-2026-10-20) is authorized for Claude/Codex without a repeat prompt where native tools support it; supersedes the prior owner-only policy. Save current state first, preserve active work, and verify native completion. An operator-only interface remains a capability limitation, not a reason to fabricate input or replace sessions.


## Codex routine work and Sol delegates — owner direction, 2026-10-20

Routine supervisor status, management publication, handoffs, prose/configuration and mechanical changes run directly in the main session; no subagent by default. The checked-in repository default for the main Codex session is `gpt-6.1-sol` at `low` reasoning effort. Delegated Goal/Spec drafting and substantive independent reviews default to `gpt-6.1-sol` at high reasoning effort through repository `.codex/config.toml`. Both are repository defaults and are user-overridable; explicit session settings and approved per-spawn overrides take precedence. Verify actual loading/dispatch rather than assume a running session refreshed. Use bounded fresh inputs where model inheritance prevents an override; no extra persistent session, test spawn, full-history fork or separate quota is implied. Preserve formal review/authority gates and Claude's separately selected implementation model.


## Restruct SDLC skill-suite packaging

The GitHub Restruct submodule at `.agents/skills/restruct-sdlc/` is the default
development dependency. Its parent entrypoint routes to the three complete child
skills `sdlc-management`, `repository-supervisor` and `claude-session-handoff`.
The `.claude/skills/restruct-sdlc` alias points to that location. Updates are
explicit reviewed Git submodule pin changes; its MIT notices stay with it.
To use another workflow, delete the submodule and alias and replace or update
AGENTS.md, this file and `docs/agents/` with your own rules; see
[Development workflow](docs/DEVELOPMENT-WORKFLOW.md).

## Retain local validation diagnostics — owner direction, 2026-10-21

Before retrying or handing back a failed local run, verify its scanned check ledger,
underlying failure logs and supplemental JUnit survive temporary cleanup in the
explicit private evidence directory. Record any withheld/missing detail and keep
failed and later outcomes distinct; wrapper/phase failures are not underlying test
IDs. Follow [local evidence retention](.agents/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#preserve-local-validation-evidence-before-cleanup).
