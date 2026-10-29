---
name: repository-supervisor
description: Resume repository supervision from management records, review Claude Plans or Results, and track Goal dependencies and acceptance. Use for supervisor handoffs or continued planning/review; leaves implementation with Claude Code.
---

# Repository supervisor

Recover current supervisor state from the repository rather than an opening prompt
or a replay of earlier conversation. Follow the supervisor role in AGENTS.md when
the user has assigned it to this agent, including an explicitly delegated GLM or
DeepSeek session. Loading the skill does not assign authority, approve work or
authorize merges. Claude Code remains the primary implementer; major product and
architecture decisions remain with the owner.

When the owner selects Claude supervision, ChatGPT/Codex is limited to Work Spec writing and architecture reviews. Do not continue routine supervision, monitoring, acceptance or dispatch under this skill in that mode. Follow the recorded role assignment; merely documenting this option does not activate it.

## Recover state

1. Resolve the current checkout's Git root. Read root AGENTS.md and CLAUDE.md,
   canonical architecture/security/decisions and applicable ADRs, budget policy,
   and the active release's ROADMAP.md and PROGRESS.md as required by governance.
2. Read `management/SUPERVISOR-HANDOFF.md` if present as a dated state reference.
   Prefer current committed records and verified state over its older observations.
   Use another release's existing management structure when governance selects it.
3. Identify active Goals and the specific user-authorized next action. Read their
   Goal, Spec, exact approved Plan/inventory, latest review/release/relocation
   confirmation and Result when present. Distinguish design approval, execution
   release, conditional integration and final implementation ACCEPT.
4. Verify checkout/branch/HEAD, outstanding edits, relevant predecessor status,
   immutable input bindings and current cycle/checkpoint usage. Read-only discovery
   first; do not checkout/rebase someone else's active worktree or refresh bound
   inputs to match the integration branch.

For session checks and authorized messages, load the sibling
[Claude handoff skill](../claude-session-handoff/SKILL.md). Use its compact Python
helper and current registration; do not embed another socket implementation here.
Route to the pinned preferred coordinator by omitting `--session` only while the
helper's `status` is `pinned-live-verified`; otherwise target a verified session
explicitly with `--session`. Verify actual receipt and dispatch. Report model prose as reported status until
validation artifacts independently support it. Continuous watching is unnecessary.

## Apply workflow updates to live repository guidance

When the supervisor updates Restruct SDLC workflow rules, it also updates the
corresponding live governance files in the adopting repository in the same change
batch, then commits and pushes to that repository's configured integration branch
under its publication policy. Updating reusable references or adoption templates
alone is insufficient. Preserve repository-specific settings and existing task
bindings; record any conflict or explicit owner deferral rather than silently
leaving the live guidance stale. This applies to every adopting repository.


## Checkout and delegated review defaults

Treat unsupported Claude task/role worktree naming as a harness bug/limitation;
continue authorized work with native generated isolated names. Verify/record
actual task, role, owner, path, branch and HEAD; do not demand another naming
exception, delay dispatch or rename a live session. Isolation/permission failures
remain blockers; a cosmetic name alone never is.

Use the main integration checkout for supervision; specifications and read-only
reviews normally need no separate worktree. Prefer bounded subagents for Spec
drafting/research and formal reviews, keeping one writer per artifact. Reviewers
are read-only by default. If a Codex worker needs isolation, use
`.codex/worktrees/<task-id>-agent-<id>-<role>`, recording checkout/branch/owner;
Claude workers use `.claude/worktrees/`. Preserve active checkouts.

Record Goal/Spec priority: owner choice first; otherwise determine or ask when needed,
normally defaulting to high. Plans target 3–5 implementation workers for high or
1–3 for low when feasible; record constraints and exclude coordinator/reviewers.
The planner may remain one agent; kickoff checks readiness and follows the Plan.
Keep one task owner/writer per artifact, source/resource/stop and acceptance gates;
see [parallel scheduling](../sdlc-management/references/coordination.md#default-to-safe-parallel-work-lanes).

When a new watcher notice or owner-requested status check reveals a review-ready Claude Plan/Result,
start a bounded subagent review without another owner prompt. Verify the exact
submission and avoid duplicate review of unchanged accepted content. Integrate
findings and issue the supervisor's disposition; a child cannot approve for you.
If subagents are unavailable, report it and review directly. Automatically relay
the decision and remaining gates under the procedure below. Owner interruption
or redirection remains available, and mandatory stops/approvals remain binding;
this adds no background watcher or blanket implementation authority.

The Claude session-handoff watcher already polls the coordinator and delivers
milestone notices. The supervisor does not poll coordinator sessions, ping the
watcher or repeat unchanged status reads unless the owner explicitly asks for a
bounded status check. Verify linked repository records on a new notice. Send needed coordinator decisions or requests immediately using nonblocking
delivery. Continue supervisor work without waiting for acknowledgment or polling
the coordinator transcript; the watcher delivers responses asynchronously. Receipt
is not acknowledgment or execution.

For a bounded delegated review, keep the supervisor turn open using the harness's
asynchronous agent wait/notification interface; the owner may message or redirect
meanwhile. Collect the findings, then finalize, record and relay the review before
the final reply. Use bounded waits without repeated inference-based polling.
Waiting uses no model tokens; reviewer work and result processing do. If an ended
turn cannot be resumed by a notification, say so rather than promising automatic
wake-up; no scheduler, extra monitoring agent or session creation is implied.
An interruption of your turn alone is not stop-all authority: on resuming, check
the children's actual status, keep completed findings and authorized running work,
and do not cancel or re-dispatch automatically. An explicit stop-all or a
conflicting scope/security/owner direction governs; client cancellation may not
reach child agents, so verify rather than assume.

When used with `saas-foundation`, `sdlc-management` takes precedence for
overlapping lifecycle/management process. Preserve the foundation's renderer
safety and technical checks and the repository's canonical invariants/decisions.
Neither skill grants a supervisor role merely by being loaded.

Apply the current local-after-merge policy: supervisor/Duo and MR fast review, then merge, then the existing implementer coordinator automatically owns affected local qualification of the exact merged SHA. Full local runs on the integration tip are owner-requested or supervisor-recommended; record the reason and freeze the observed tip. Explicit technically justified full-run requirements remain. A ready authorized successor may implement in isolation during that run after merge; predecessor final ACCEPT still gates successor integration. No automatic post-merge integration/release-branch or tag-push pipeline, including lightweight checks; preserve feature MR fast, owner-selected full RC-promotion MR qualification, explicit diagnostic/provider triggers and owner holds.

For ordinary feature registration/schema/count-pin edits, exclude full CI controls;
review only affected product checks and control cases. A shared harness pathname
or aggregate PASS row is not a broad-coverage reason. Apply
[feature control selection](../sdlc-management/references/validation-integration.md#ordinary-feature-registration-does-not-require-full-ci-controls).

Publish reviewed management-only acceptance/cleanup handbacks directly under the
integration-branch publication authority; do not require another Result-only MR,
automated review or CI run. Follow [closeout publication](../sdlc-management/references/closeout-risks.md#direct-supervisor-publication-of-acceptance-and-closeout).

## Supervise the next authorized action

Before any automatic transition, check recorded owner specification/work/phase/stop bounds. Missing Specs stop
continuation unless explicitly authorized within a recorded endpoint/range or pinned epic. Spec-only mandates
stop at documentation; full-lifecycle mandates retain every gate. Stops override existing Specs/accepted Plans,
including lookahead planning during CI; clarify ambiguous endpoints before crossing and do not skip to later tasks.
Record owner instruction, membership, phase/endpoint and resume authority in existing PROGRESS/handoff records.
Policy examples grant no actual epic/range. See [mandates and stops](../sdlc-management/references/roles-lifecycle.md#bounded-specification-authority-and-owner-stop-boundaries).

When a predecessor implementation MR first opens against the integration branch (drafts included), verify
the MR/source/target/candidate and automatically dispatch provisional planning for the earliest eligible task with an
existing ready supervisor-owned Goal/Spec; no separate owner prompt. Reuse its current planner if assigned. Label
pending predecessor acceptance provisional, record MR/candidate baseline and holds, and preserve explicit owner
deferrals. Do not wait for CI/Duo, ready status or merge; ignore planning-only/placeholder/probe MRs. Planning
does not release implementation, direct-tier work, test resources or dependency gates. Follow the adopting
repository's rule or [portable policy](../sdlc-management/references/roles-lifecycle.md#automatic-planning-after-integration).

After reviewed predecessor source is merged, reassess the provisional Plan against that exact integration.
Record/relay release for isolated implementation once Plan/input/scope, owner-decision, budget and resource gates
close; it may overlap predecessor post-merge affected local qualification/closeout without another routine owner prompt. Final predecessor
ACCEPT still gates successor integration/final acceptance. Material changes need bounded updates/review; pause
affected work for unresolved security/architecture/contracts or failed qualification, preserving unaffected work.
The predecessor implementation owner keeps its staging repairs/ledger in parallel with successor work; do not
assign them to successor workers by default. Preserve one writer, explicit holds and delivery barriers.
A merge/Plan alone is not release; apply current model-aware effort. Follow
[release policy](../sdlc-management/references/roles-lifecycle.md#automatic-reassessment-and-execution-release).

- **Goal/Spec:** reconcile the actual accepted predecessor and roadmap; define a
  bounded measurable outcome, non-goals, dependencies, AC evidence, budget and
  stops using repository conventions. A recorded owner epic mandate may cover
  sequential specification/delivery; keep separate bounded Goals and reviews,
  and never infer epic authority or expand its pinned membership.
- **Plan review:** compare the exact Plan and changed inputs with the Spec, all
  ACs, security/architecture, budget and required evidence. Reuse accepted controls.
  Check cited reuse/extend/new test coverage against actual gaps; request bounded
  corrections for unjustified duplication. Distinct policy/RLS/integration boundaries
  can need similar scenarios; require their purpose, not a separate test inventory.
  Resolve a mechanical binding change without reopening settled design unless a
  substantive defect or changed assumption requires it. Import accepted Plan,
  inventory/history and review bytes into the local integration branch as required.
- **Implementation review:** inspect actual diffs and attributable AC evidence,
  including real runtime-role/isolation negatives where required. Passing tests
  alone are insufficient. Give bounded corrections; do not rewrite Claude's code.
- **Progress/integration:** update existing management records. Follow current
  rebase, local/manual/hosted validation, MR, owner-merge and branch-cleanup policies.
  Never infer merge/action authority from a review or session delivery.

Use exactly the governance outcomes for formal review: **ACCEPT**,
**CORRECTION LOOP**, or **ARCHITECTURE DECISION REQUIRED**. State what ACCEPT
accepts and the remaining gates. Corrections identify the finding, guarantee,
minimum relevant retests and authorized budget. Use the [compact repair handoff](../sdlc-management/references/coordination.md#compact-repair-handoff)
for failures: relevant revision/diff, focused evidence, tried fixes, hypothesis and next test; reference existing records. Stop/escalate scope, architecture,
security, public-contract, irreversible-model and resource conflicts as required;
do not extend/reset budgets or weaken evidence to finish.

Preserve UUID canonical identities, current tenancy/authority semantics and the
G021 product-feature barrier. Read service-root instructions/skills for service
work. The skill contains no substitute permission matrix, CI policy or fixed
execution baseline; those belong to current approved repository records.

For [mechanical corrections](../sdlc-management/references/roles-lifecycle.md#autonomous-mechanical-corrections),
let the released implementer fix approved-scope selectors/fixtures/pins/catalog parity
and harness bookkeeping directly. Review the eventual diff/evidence; do not turn
routine progress into a new approval/replan or duplicate review. Material scope,
security/contracts, unapproved schema/constraints, budgets and explicit holds remain stops.

## Leave durable state

Record decisions in goal-specific reviews/Results and index them in PROGRESS.
For cross-session continuity, update the existing supervisor handoff with dated
active SHAs/checkouts, links to contracts, current phase/cycle usage, open findings
and next gates. Distinguish direct evidence from reports and stale observations.
Keep approval history and bindings intact; an updated state note is not a release.

At every epic closeout, perform the repository's PROGRESS compaction checkpoint before marking closeout complete.
Archive superseded updates with working links and preserve the concise current index, acceptance and failure history.
Record a no-change outcome when nothing needs moving, or a specific active-binding deferral and next safe checkpoint.
This is supervisor housekeeping, not another validation campaign or permission to rebind active work.

Send concise references to those records using the handoff helper. Its root
management delivery log tracks attempts/receipt IDs, not acceptance or execution.
After every completed review, notify the responsible verified Claude session
without another owner prompt, including outcome, reviewed inputs, next action and
remaining gates. Verify coordinator relay where applicable, receipt and explicit
acknowledgment; report pending/unavailable delivery without automatic resends or
session creation. Notification conveys the recorded review authority only.
Do not store full conversations or secrets. Commit/push, hosted writes, merges and
new implementation require existing owner authorization. On subsequent turns,
read changed sections and unresolved findings instead of replaying all sources.


## Fable weekly-limit fallback — owner direction, 2026-10-15

When Fable's weekly allowance is confirmed exhausted, automatically use Opus 5.5
**xhigh for planning and test-failure diagnosis/corrections** until Fable resets.
Opus implementation, ordinary test execution/reporting and coordination remain
**high**. This standing fallback needs no repeat owner approval; honor an explicit
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


Use the adopting repository’s low-overhead delivery policy; when the companion is installed, see
[validation/failure handling](../sdlc-management/references/validation-integration.md#validation-economy-and-recurring-failures),
[routine transitions](../sdlc-management/references/roles-lifecycle.md#continue-routine-transitions-without-owner-nudges),
[concise messages](../sdlc-management/references/coordination.md#concise-operational-messages) and
[closeout receipts](../sdlc-management/references/closeout-risks.md#routine-closeout-receipt).
Preserve standalone operation and every existing authority/native-permission boundary.

Use [targeted reruns and attributable coverage](../sdlc-management/references/validation-integration.md#targeted-reruns-and-attributable-coverage-default)
for local qualification. Review frozen source while independent tests run; preserve
preliminary/final distinctions, input provenance and later-delta review. A partial
rerun never becomes a full passing report or waives required hosted checks.

Before approving local validation or accepting its selection, check the actual
executable scope from the first run. Reject unexplained broad umbrella suites;
require a concrete technical guarantee/dependency reason before approving an
exception. Reuse an unchanged approved exception without another permission loop.
The companion's focused-selection policy applies when installed; required hosted
jobs, security evidence and native gates remain unchanged.

For planning and review, apply [early risk review and short correction loops](../sdlc-management/references/validation-integration.md#early-risk-review-and-short-correction-loops):
changed trust boundaries first, approved risky slice early, bounded caller/pin
inventory and consolidated corrections. Use existing phases/gates and compact
handoffs; do not restart active work or create another process framework.

Apply [supervisor session continuity](../sdlc-management/references/coordination.md#supervisor-session-continuity)
when the companion is installed: native context limits, checkpoint/compaction
first, durable role/delegate handoff and explicit observer thread rebind. Honor
recorded owner transition exceptions; a long transcript alone does not require
retirement or authorize a replacement session.


## Mandatory task-acceptance cleanup — owner direction, 2026-10-17

At every final task ACCEPT, immediately remove all finished task-owned worktrees
and temporary resources under the [mandatory cleanup policy](../sdlc-management/references/closeout-risks.md#mandatory-cleanup-at-every-task-acceptance).
Unlock finished agent-tool worktrees; preserve unique commits and dirty/untracked
work before force removal when needed. Clean task-owned containers, unused images
(including attributable untagged IDs), volumes and disposable evidence without
another routine approval. Keep active coordinator/successor/preview resources and
specific recorded retention/native-permission holds. Verify a compact receipt;
no silent lock/dirty-work deferral, blanket host prune or extra test campaign.


## Private files for new tasks — 2026-10-17

Use the main integration checkout's ignored `.local/{evidence,secrets,tmp}/` for
new task-private files, shared across worker worktrees. Follow the [local storage
policy](../sdlc-management/references/adoption.md#repository-local-private-files): task-scoped paths, private permissions, Git/build/upload exclusions
and separate evidence/secret cleanup. Existing tasks keep current paths through
acceptance cleanup; do not migrate active bindings or native session state.

## Repository-scoped permission configuration

Keep shared project rules tracked and repository-specific personal exceptions
ignored; home settings are for intentional cross-project preferences. Verify the
actual loaded rules for the session/worker launch directory. Owner-authorized
configuration is distinct from task authority; native rejection gates remain.
Follow the adopting repository's permission placement/adoption procedure.

## Mandatory short owner reports

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

For integration repairs, independently reconcile the Result's complete failed-run
inventory with actual suite logs before pre-integration ACCEPT; see
[repair completeness](../sdlc-management/references/validation-integration.md#complete-repair-inventory-and-supervisor-verification).
Do not accept a later passing subset as proof that earlier failures cleared.


## Direct owner replies — 2026-10-19

For ordinary replies to direct owner prompts, omit the `COORDINATOR_STATUS` footer unless the owner explicitly asks for supervisor relay. Exception: an owner decision that changes active work must immediately emit a `decision` footer, including when the same message answers the owner. Do not wait for another turn, worker handback or review-ready commit. Autonomous handoffs and required reviews, failure records and blockers still notify; existing legacy-prose/native-question detection is unchanged.


## Coordination economy — owner direction, 2026-10-20

Use150k active-context checkpoint/200k safe handoff assessment, respecting smaller native limits. Claude and Codex may initiate supported native self-compaction at safe points under standing owner authority. Apply at safe boundaries without interrupting validation, resetting budgets or copying a full oversized transcript. Handle routine supervisor-authored prose and mechanical changes directly with focused checks; preserve substantive independent architecture/security and required submitted Plan/Result/source reviews. Batch independent reads and related findings, give delegates only necessary bounded context, and do not repeat unchanged polling, receipt acknowledgments or reviews. Preserve prompt failure intervention and truthful evidence; usage estimates are not actual bills.


## Codex routine work and Sol delegates — owner direction, 2026-10-20

Routine supervisor status, management publication, handoffs, prose/configuration and mechanical changes run directly in the main session; no subagent by default. Delegated Goal/Spec drafting and substantive independent reviews default to `gpt-6.1-sol` through repository `.codex/config.toml` (`[agents].default_subagent_model`), even when the owner selects a cheaper main model. Explicit approved child overrides win; verify actual loading/dispatch rather than assume a running session refreshed. Use bounded fresh inputs where model inheritance prevents an override; no extra persistent session, test spawn, full-history fork or separate quota is implied. Preserve formal review/authority gates and Claude's separately selected implementation model.

Cleanup is continuously accounted for at allocation, handback/batch end and
interruption/restart, under [mandatory resource accountability](../sdlc-management/references/closeout-risks.md#continuous-resource-accountability--mandatory).
Verify every category; abandoned resources block new resource-consuming work.
Reclaim finished worker trees at handback, not only at final acceptance.
