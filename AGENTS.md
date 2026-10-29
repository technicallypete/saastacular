# Repository Management Governance

SaaStacular ships the **Restruct SDLC** workflow as its default bundled development
dependency. The complete package lives at `distributions/skills/restruct-sdlc/`
(under its own MIT licenses; the repository is Apache-2.0), with two relative
discovery aliases, `.agents/skills/restruct-sdlc` (Codex and other agents that scan
`.agents/skills`) and `.claude/skills/restruct-sdlc` (Claude Code), both targeting
`../../distributions/skills/restruct-sdlc`. The rules in this file and in
[CLAUDE.md](CLAUDE.md) are that workflow's live governance for this repository;
the runbooks under [docs/agents/](docs/agents/) and the bundled references hold
the detailed procedures. The `management/` records these rules refer to are not
shipped: bootstrap them with the bundled `sdlc-management` skill
([adoption guide](distributions/skills/restruct-sdlc/sdlc-management/references/adoption.md)).
The application's configure, build, test, migrate and run commands do not need
any agent tooling.

**Using another workflow.** Remove both aliases (`.agents/skills/restruct-sdlc`
and `.claude/skills/restruct-sdlc`) and replace or update AGENTS.md, CLAUDE.md
and `docs/agents/` with your own rules. Removing the aliases alone disables skill
discovery but not the workflow instructions in these files. See
[docs/DEVELOPMENT-WORKFLOW.md](docs/DEVELOPMENT-WORKFLOW.md).

Default role assignment (unless the owner explicitly selects the Claude-supervisor option below):

Codex supervises planning, architecture/security conformance, plan review, implementation acceptance, and progress. Claude Code is the primary implementer. Codex must not become the primary implementation agent unless the user explicitly requests it, and must not rewrite Claude's implementation during correction review.

## Repository orientation

SaaStacular is a repository template for single-tenant and multi-tenant applications:
a Go `core`, a Next.js `web` surface, an optional internal Python `intelligence`
service, PostgreSQL, Redis and S3-compatible object storage (MinIO locally), with the
migrations, Docker configuration, CI definition and tests that exercise them. See
[README.md](README.md) and [docs/SECURITY.md](docs/SECURITY.md).

Everything runs in Docker from the repository root; no host Go, Python, Node,
PostgreSQL, Redis or MinIO is needed. The README's quick start holds the exact,
harness-executed blocks:

- configure: `infra/local/configure-env.sh` (creates `.env` from `.env.example`; never overwrites it);
- start: `docker compose --env-file .env -f infra/compose/local.yml up -d --build --wait`;
- migrate, check, browser tests and contract conformance: the `platform:*` blocks in [README.md](README.md#quick-start);
- validation profiles: `bash infra/ci/validate.sh --profile fast|integration|full`, or `--profile affected --base <commit>`;
- stop: `docker compose --env-file .env -f infra/compose/local.yml down`.

The secret scan, the validation harness and the affected profile need a Git
repository (see the README section "Developer checks need a Git repository").
Suites, profiles and check IDs are indexed in [docs/foundation/TEST-INDEX.md](docs/foundation/TEST-INDEX.md).

Architecture constraints:

- `core` owns domain authority: identity and sessions, tenancy, authorization, audit, storage, domains, MCP and migrations.
- `web` is a browser surface and thin server-side bridge to `core`: no authority, policy or domain rules, and no `NEXT_PUBLIC_*` backend URLs.
- `intelligence` is optional internal AI execution behind a model-gateway boundary, with no data-store access and no browser reachability.
- PostgreSQL holds durable state, Redis is ephemeral and object storage is S3-compatible. External providers stay behind owned adapters.
- Identity is global; the Organization is the tenant boundary, resolved on the server from the authenticated session and the request host, never from client-supplied or model-supplied identifiers. Row-level security is defense in depth, not the only control.
- Product name, domains and database name are configuration: they live only in `.env` and `.env.example`.
- The UI kit source in `web/src` stays unaltered; change themes through the tokens described in [docs/foundation/WEB-SHELL.md](docs/foundation/WEB-SHELL.md).
- Never commit `.env` or other secrets; the secret scan refuses local-only files.

The mechanisms behind these rules are documented in [docs/foundation/](docs/foundation/)
(start with TENANCY, AUTHORIZATION, DATABASE-ISOLATION, IDENTITY and MCP).

## Post-merge staging qualification runs locally

Owner direction, 2026-10-22 UTC, superseding the blanket full-local-run default of 2026-10-19:

Default sequence for implementation MRs: focused pre-merge checks and current hosted fast/supervisor/Duo review → authorized merge → affected local batches/suites on the exact merged integration SHA. The existing implementer coordinator automatically launches, owns and tracks those selected checks; no extra owner command or full pre-merge gate. Select from the actual change and integration delta, including affected dependencies, shared fixtures, migrations, disabled variants and security boundaries. Record executable selectors and coverage rationale in the existing Plan/Result; preserve immutable inputs, actual outcomes and cleanup. Final Goal ACCEPT requires that attributable affected coverage and the other Goal gates, rather than a blanket full run per feature.

Run full local qualification on the integration branch tip when the owner requests it **or the supervisor recommends it**, recording the reason and freezing the exact observed tip for the run. Full staging runs and full hosted RC-promotion runs provide broader opportunities to find gaps in feature selection; unexecuted checks are never reported as passing. Ordinary feature MRs retain hosted fast checks; RC promotion MRs retain required full hosted qualification. Automatic integration/release-branch post-merge and tag-push pipelines remain disabled, including lightweight checks; separately authorized diagnostic/provider runs remain distinct.

After a failure, rerun the failed checks and checks affected by the fix, reusing attributable unchanged passes. A broad local selection needs a concrete coverage reason; explicit technically justified full-run requirements remain binding. Reconcile active work at a safe boundary without restarting running batches or rewriting approved snapshots. Ready authorized successor implementation may overlap post-merge checks in isolation, respecting resource reservations; predecessor final ACCEPT still gates successor integration/final acceptance. Preserve owner holds, manual/security gates, documentation-only exemptions and failure history. This supersedes the blanket full-local-run-after-every-feature default.

Ordinary application feature work does **not** require the full `ci-controls` suite merely because it updates product test registration or expectations inside harness files. This includes routes/import allowlists, migration versions, schema/catalog counts, browser matrices, test invocations, required-check IDs and their resulting count pins. Run affected product checks and only the directly affected control cases using existing focused selectors. A changed harness pathname, copied fixture, mechanical count update or desire for an aggregate PASS row is not a reason to select all controls. Full CI controls require a recorded concrete reason involving changed shared runner behavior or an explicitly selected full staging/RC qualification. Required-check aggregates apply to selected coverage; do not add `ci-controls` merely to make `B-CI-CONTROLS` required. Preserve actual outcomes and report focused control proof without claiming an unexecuted full suite passed.

See [the current policy](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#post-merge-qualification-runs-locally).

Ready, authorized successor implementation may overlap that local run **after predecessor merge**, in isolation with shared-resource locks respected. Predecessor final ACCEPT still gates successor merge/final acceptance, and its own owner retains repairs. Explicit owner/phase holds and documentation-only exemptions remain. This supersedes the earlier pre-merge/unmerged-successor variant. See [current policy](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#post-merge-qualification-runs-locally).


The merged-SHA local qualification checkout remains frozen: development-branch fetch/rebase preparation does not rebase that tested merge. Inspect later target changes separately and reconcile their actual effects; do not silently switch the qualification source or repeat unchanged coverage for management-only commits.

## Owner-selected Claude supervision

Owner direction, 2026-10-19 UTC: the owner may choose Claude Code as supervisor. In that mode, Claude owns orchestration, Plan review, implementation acceptance, progress and closeout, using the existing implementation/review workflow. ChatGPT/Codex participates **only in Work Spec writing and architecture reviews**; it is not a second supervisor or a routine approval, monitoring or dispatch gate. Major product/architecture decisions remain with the owner. Record the selected roles in the existing governance/handoff when the owner selects this mode; this option alone does not change the current assignment, restart sessions or alter active task bindings.


## Management ownership matrix

| Artifact or responsibility | Owner | Purpose |
|---|---|---|
| Architecture / product direction | Product owner + ChatGPT | Canonical product and architecture decisions. |
| Roadmap | Codex | Dependency-ordered epics and stable goals. |
| Goal | Codex | Bounded outcome, scope, execution policy, autonomy, budgets, stop conditions, and suggested delegation. |
| Work Spec | Codex | Functional/technical requirements, architectural invariants, acceptance criteria, security requirements, and required validation. |
| Implementation Plan | Claude Code | Concrete approach, files/modules, sequence, migrations, tests, subagents, and acceptance-criteria mapping. |
| Implementation | Claude Code | Execute the approved plan within scope. |
| Tests / validation | Claude Code | Create and run required checks, diagnose failures, and report evidence. |
| Result report | Claude Code | Record changes, validation, acceptance checklist, deviations, and unresolved issues. |
| Plan review | Codex | Check scope, feasibility, architecture, security, and validation coverage before execution. |
| Implementation acceptance | Codex | Review diffs and evidence against the Goal, Spec, architecture, and security requirements. |
| Progress tracking | Codex | Maintain active, completed, blocked, and upcoming goals. |
| Major architecture/product decisions | Product owner | Decide escalated changes; neither agent may silently decide. |

## Canonical sources

Before planning, implementing, or reviewing, read:

- AGENTS.md and CLAUDE.md
- docs/ARCHITECTURE.md (not shipped; until the project adds one, the README's "Architecture at a glance" and docs/foundation/ serve as the architecture reference)
- docs/SECURITY.md
- docs/DECISIONS.md and relevant accepted ADRs
- management/v0/ROADMAP.md
- management/v0/PROGRESS.md

For a specific goal, also read its Goal, Work Spec, Plan, and Result when present. `docs/ARCHITECTURE.md` is authoritative over historical discussion. Conflicts with architecture or accepted ADRs require escalation, not silent resolution by either agent.

Preserve the roadmap's foundation barrier: no product-specific feature implementation until the reusable archetype and core foundation are accepted through G021. Approval of an individual goal does not waive that barrier.

## Service working directories and scoped skills

For service-specific planning, implementation and review, start the agent session
or delegated task in that service's root (`web/`, `core/` or `intelligence/`)
within the approved checkout. Read repository-wide governance and the active
Goal/Spec/Plan first, and load applicable service instructions and skills.
For UI work, explicitly read the repository-root-relative
`docs/foundation/WEB-SHELL.md` (resolve the Git root first when the
session starts in a service directory)
for application UI work: the UI kit source in `web/src` stays unaltered and themes
change through its tokens. For repository skills (such as the bundled
`restruct-sdlc`), use the repository copy rather than
assuming a same-named personal skill is equivalent.

For work spanning services, use the repository root and explicitly load each
relevant service's instructions and skills. A shell `cd` in an existing session
does not establish that the agent refreshed skill discovery; explicitly read the
applicable skill when continuing such a session. Run repository-level Compose,
CI and validation commands from the repository root as documented. Directory
selection does not change the approved checkout, input bindings, scope or gates.

## Versioned management structure

Use the release-specific management directory; v0 uses:

```text
management/v0/
  ROADMAP.md
  PROGRESS.md
  BACKLOG.md
  goals/
  specs/
  plans/
  results/
```

Artifacts are goal-specific, for example:

```text
goals/G000-project-archetype.md
specs/G000-project-archetype-spec.md
plans/G000-project-archetype-plan.md
results/G000-project-archetype-result.md
```

Goal identifiers are stable; do not renumber or reuse them. Do not overwrite prior goal artifacts or replace one goal's files with another's. Revisions within the same goal must preserve review history: append dated revisions/correction records or create explicitly linked revision files; do not erase prior approvals, failures, or evidence. Closed goal artifacts remain historical records.

PROGRESS.md is the mutable index of active, completed, blocked, and upcoming goals, with links to artifacts and review outcomes. Codex updates it at handoffs, blockers, and acceptance; only an ACCEPT marks a goal completed. Claude reports status/blockers in the Result and returns control to Codex. BACKLOG.md holds proposed future work, not execution authorization.

### Automatic publication of supervisor progress updates

Owner clarification, 2026-10-24: Routine acceptance and cleanup closeout records use the supervisor's direct publication route to the configured integration branch. The coordinator hands back the Result, attributable qualification evidence and cleanup receipt with exact source references; the supervisor reviews them, records the acceptance disposition and imports/publishes the completed Result, review, PROGRESS and related closeout records in scoped commits. Do not create a separate acceptance, Result-only or cleanup-receipt MR, request another automated review or launch CI solely to publish those records. The implementer receives no direct integration-branch push or self-acceptance authority. An explicit owner instruction or an actual branch-protection requirement can select an MR instead; report that reason. Actual code or executable configuration repairs retain the normal source-MR workflow. Preserve review, evidence, cleanup and native permission gates; publication alone is not ACCEPT.

Owner direction, 2026-10-12: enabled by default, whenever Codex completes a
PROGRESS.md update, Codex commits and pushes that update and its associated
management artifacts to `staging` from the main integration checkout, without
another owner confirmation. This includes received Result/status records that
Codex imports and references, and related reviews or decisions. Codex verifies
the imported bytes and records whether evidence is reported or independently
reviewed; publication alone is not ACCEPT.

This is an agent workflow, not a file watcher or per-save Git hook. Publish one
completed update batch, stage exact owned paths, exclude secrets/machine-local
handoff logs and unrelated edits, and preserve unique work and review history.
Fetch and reconcile the integration branch safely before pushing; never force
push staging. If a conflict, protected-branch rule or native permission gate
blocks publication, preserve the update and report the blocker.

Owner direction, 2026-10-20: Codex commits and pushes each completed requested
repository update batch to the configured integration branch (here:
`staging`) after focused verification, including documentation, Goal/Spec and
configuration updates; do not leave completed requested changes only in the
local checkout. Batch related files, stage only owned paths, fetch/reconcile
before push, and never force-push the integration branch. This applies without
another routine owner prompt. An explicit local-only instruction overrides it;
real conflicts, active binding hazards or native/protected-branch refusals must
be preserved and reported rather than bypassed. In-progress drafts are not
published until the requested update batch is complete and verified.

Claude continues to implement and publish checkpoints on its task branch; this
rule grants no direct Claude staging push or application integration authority.
Assess whether the staging update affects active task bindings and notify the
coordinator at the appropriate safe point under the existing reconciliation rule.
The owner may disable automatic publication for an explicit task/session; record
that override. Reusable adoption exposes this documented option as enabled by
default, with its integration branch explicitly configured (`staging` here).

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
[continuous resource accountability](distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#continuous-resource-accountability--mandatory).

## Supervisor progress-index closeout

### Bounded operational cleanup

Owner direction, 2026-10-14: record task-created resource ownership in the existing
run inventory at creation, including exact container/image IDs and mounted volume
names before container removal. Reuse existing cleanup with one scoped preflight,
one removal pass and a compact receipt. Hold uncertain/live items and finish the
verified remainder; unrelated host services and container churn alone are not
blockers. Preserve interrupted receipts and resume unfinished items. Do not add
another cleanup framework, approval loop or test campaign. Report native refusals
once for supported owner handling; never bypass them. Required active-run cleanup
gates and existing retention/input bindings remain unchanged. Full procedure:
[bounded routine cleanup](distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#bounded-routine-cleanup).

Owner direction, 2026-10-11: Codex owns PROGRESS.md housekeeping after each
Goal's final ACCEPT and integration closeout. Assess whether repeated or
superseded updates should be compacted; conditional integration, submission or
Plan approval alone does not close a Goal. Keep one current status entry per
Goal, with its acceptance/review links, dependencies, active blockers and next
gates. Keep unresolved decisions and cross-task effects visible.

Move superseded historical updates into a committed, dated Markdown archive
within the same release's management directory (for v0,
management/v0/archive/PROGRESS-YYYY-MM-DD.md). Preserve their original text,
dates, decisions and evidence references; link the archive from PROGRESS.md.
Retain the acceptance disposition and artifact links in the current index.
Do not delete or rewrite Goal/Spec/Plan/Result/review history, infer acceptance
from a merge, or turn an archived failure/cancellation into a pass. Check moved
links/anchors and adjust relative references so they still resolve.

Use a documentation-only closeout change, under existing commit/push authority.
Do not compact entries needed by active tasks or edit their immutable snapshots
or execution checkouts. If the index is a bound input of an active task, either
defer compaction to a safe closeout window or notify its coordinator to reconcile
the integration-branch change under the accepted procedure before a later batch.
Preserve safe running batches and reuse unchanged validation evidence; compaction
alone does not justify another full test run. This policy requires assessment
at closeout, not a rewrite of the entire index after every status update.

Owner direction, 2026-10-14: every epic closeout is a required periodic
PROGRESS compaction checkpoint owned by Codex. Compact accumulated superseded
updates into the existing dated archive structure and retain a concise current
index with links. If nothing needs moving, record that outcome; if active-task
bindings prevent safe compaction, record the specific deferral and next safe
checkpoint. Epic closeout is not complete without that recorded disposition.
This adds no test run, new tracking framework or permission to alter history.

### Apply reusable workflow changes to live guidance

Owner direction, 2026-10-20: when the supervisor updates Restruct SDLC workflow rules, it also updates the
corresponding live governance files in the adopting repository in the same change
batch, then commits and pushes to that repository's configured integration branch
under its publication policy. Updating reusable references or adoption templates
alone is insufficient. Preserve repository-specific settings and existing task
bindings; record any conflict or explicit owner deferral rather than silently
leaving the live guidance stale. This applies to every adopting repository.

The bundled copy under `distributions/skills/restruct-sdlc/` is a pinned copy of
the upstream Restruct SDLC project, which remains its source of truth. Updating
the bundled copy is an explicit, reviewed version change. In this adopting
repository, publish the corresponding live governance updates directly to
`staging` under the supervisor publication authority; do not open a separate MR
solely for those Restruct SDLC guidance changes. This exception does not change
the normal MR workflow for executable code, feature work, repairs or RC
promotion.


## Supervisor → implementer workflow

### Default parallel scheduling

Owner direction, 2026-10-26, strengthening the 2026-10-22 rule: when authoring a
Goal/Spec, record **high** or **low** priority and its source/rationale. The owner
should identify low-priority work explicitly. If unspecified, the supervisor may
determine priority or ask the owner when the distinction matters; normally use
**high** without blocking specification on a routine priority question. An explicit
owner choice takes precedence; attribute supervisor choices as supervisor decisions.

During planning, target **3–5 parallel implementation workers for high priority**
and **1–3 for low priority**, whenever useful independent work and native/resource
capacity permit. These are implementation worker targets, excluding the planner,
coordinator and reviewers. Preserve dependency, exclusive write ownership and
approved budget limits; do not manufacture work to satisfy a worker count.
Record the worker count by phase, roles, nonconflicting writable paths, bound
inputs, dependencies, integration owner, evidence and resource reservations in
the existing Plan, carrying forward the Goal/Spec priority. Record concrete
dependency, ownership or capacity reasons for fewer workers than the target.
The planner may remain a single agent.

At implementation kickoff, verify readiness and dispatch the dependency-ready
workers specified in the accepted Plan; follow its allocation and sequence.
Record any actual allocation change and its reason in the existing Result;
material scope/design/gate changes retain normal review. As dependencies clear,
start the next planned workers without another routine owner prompt. Low-priority
work may use one implementation worker within its 1–3 target and
yields capacity to the current high-priority implementation. Concentrate capacity
within that task; concurrent Goals retain their existing authority/dependency
gates, and routine supervisor staging updates may proceed.

Keep one accountable task owner and one writer per artifact. Reclaim finished
workers promptly; do not create duplicate work merely to occupy agents. A failure
holds its batch and dependent work; unrelated safe batches may proceed. Preserve
task-specific sequencing, source freezes, owner stops and all review/acceptance
gates. Private DB/browser tests share the existing two-run pool (third only with
owner approval); image builds remain sequential. Apply this to active work at a
safe handoff, preserving accepted Plan history and frozen runs. Use existing
coordinator facilities and records. See [parallel scheduling](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#default-to-safe-parallel-work-lanes).

### Supervisor checkout, delegation and automatic reviews

Owner direction, 2026-10-12: the supervisor normally works in the main integration
checkout and does not need its own worktree for specifications or reviews.
Prefer bounded subagents for Spec drafting/research and formal Plan/Result
reviews; reviewers are read-only by default. Share the main checkout for
read-only work. If a Codex worker genuinely needs checkout isolation, place its
worktree under `.codex/worktrees/`, separate from Claude's `.claude/worktrees/`,
and record task, role, branch and ownership. Protect active work and private
worktree state; do not create a worktree merely to perform a review.

When a new Claude watcher notice or an owner-requested status check surfaces a
review-ready Plan/Result, the supervisor initiates a bounded subagent review without
waiting for another owner request. Verify the exact artifact and phase first;
do not review unfinished work as a final submission or repeat an unchanged
completed review. If delegation is unavailable, report that limitation and
perform the review directly rather than leaving authorized work idle. The
supervisor integrates findings, owns the final disposition and records it.

Owner direction, 2026-10-12: for bounded delegated reviews, keep the supervisor
turn open and use the harness's asynchronous agent wait/notification interface.
The owner may message or redirect while the supervisor waits. Collect findings,
finalize the review and relay it before ending the turn; do not require another
owner status prompt merely because reviewers finished after an early final reply.
Use bounded waits and communicate meaningful progress during long work, without
repeated inference-based polling. The Claude session-handoff watcher already polls
the coordinator and delivers notices; the supervisor does not poll coordinator
sessions or ping the watcher unless the owner specifically asks. Waiting itself
uses no model tokens; reviewer work and the supervisor's processing of results
still do. If the harness cannot resume an ended turn from a notification, state that limit rather than promising
autonomous wake-up. No scheduler, extra monitoring agent or session creation is
authorized by this rule; the Claude transcript observer has a separate scope.

Owner clarification, 2026-10-12: interruption of the supervisor's turn alone
does not direct it to stop all delegated agents. On resumption, inspect their
actual status, preserve completed findings and reuse continuing authorized work;
do not automatically cancel or redispatch everyone. Honor an explicit stop-all
instruction and stop affected agents when a scope/security change or conflicting
owner direction makes their work unsafe or obsolete. Client cancellation behavior
may vary; do not assume children survived or stopped without checking.

Immediately send needed decisions, corrections, requests and remaining gates to
the verified Claude coordinator. Use nonblocking delivery and continue supervisor
work without waiting for acknowledgment or polling the coordinator transcript;
the watcher delivers responses asynchronously. A direct owner request permits one bounded
status check. Receipt is not acknowledgment or execution. The owner can interrupt
or redirect; routine review and relay do not require repeated approval.
Existing architecture/security/scope stops and owner merge authority remain binding. This does not add background
polling, grant implementation authority before Plan approval or create sessions.

### Combining management and foundation skills

When both skills are used, `sdlc-management` supersedes overlapping
foundation/renderer-skill guidance about management artifacts, responsibilities,
authorization, budgets, progress, reviews, validation sequencing, Git integration
and closeout. Foundation stages become inputs to Goals and acceptance criteria.
The renderer's dry-run plan is not a Claude Implementation Plan. Preserve resolved
input review, informed write authorization, conflict handling, renderer safety
and required technical checks; existing authority satisfies repeated confirmation
only within its explicit scope.

Canonical repository governance/architecture and accepted decisions remain
authoritative. Management precedence never waives technical/security invariants,
dependencies or delivery barriers. Escalate unresolved technical conflicts.
Do not modify the foundation skill to apply this composition rule.

### Task-prefixed Claude worktrees

Owner direction, 2026-10-14: prefer `<task-id>-agent-<agent-id>-<role>` for
Claude workers when the native tool supports it safely (`-plan`, `-check`,
`-impl`, etc.). **Naming alone must never block planning, implementation or
review.** Treat unsupported task prefixes/role suffixes as a Claude harness
naming bug/limitation pending resolution. Use Claude's native isolated-worker
creation and generated `agent-<id>` path immediately; no task-specific exception,
owner question, extra approval or naming workaround is required.

Record the task ID, role, actual agent/session ID, Git worktree path, branch and
HEAD in the existing dispatch/Result record, verifying isolation and intended
ownership before task writes. Do not infer identity or authority from a name.
Wrong repository/branch/owner or unsafe isolation remains a blocker; a generated
basename alone is not. No new manifest or duplicate coordinator is needed.

This supersedes the 2026-10-12 mandatory-prefix/task-specific-exception rule and
its role-suffix gate. Historical exceptions remain valid records and need no
rewrite. Already-active bound paths stay unchanged. Do not rename a live worker,
pre-create an inaccessible path, restart sessions, install hooks or bypass a
native permission refusal just to enforce naming. No evidence rerun or replan
is required solely to use a supported native name.

Keep the single coordinator's preferred `0-agent-<id>` name when safely supported;
that prefix remains reserved for it. Verify one active coordinator through the
owner-selected pin/discovery, not directory names. Existing coordinator
relocation remains optional and subject to supported safe-point checks; defer
an unsafe rename while continuing authorized work in the existing checkout.
Claude worker state stays under `.claude/worktrees/`; Codex workers that need
isolation stay under `.codex/worktrees/`. Preserve actual harness locks and
permission/checkout/input boundaries; this changes a naming convention only.

For feature/Goal integration, follow [GitLab branch integration and cleanup](docs/agents/GITLAB-BRANCH-WORKFLOW.md): fetch and rebase the feature/Goal worktree branch onto the remote target before opening the MR, reconcile conflicts and exact input/evidence bindings, use glab to open an MR targeting `staging`, retain the enabled source-branch deletion default, and verify cleanup after merge or MR closure under the runbook's preservation checks. Retain long-lived `main`, `staging` and `identity-qualification`. This route does not replace implementation ACCEPT, applicable CI, owner integration authority or exact input/evidence bindings.

At task acceptance, separately audit all task-owned remote refs; automatic MR
source-branch deletion is not a complete cleanup check. Reconcile accepted local
refs and exact task-owned Docker image IDs under the same ownership/unique-content
rules. Keep shared images, active work and unresolved unique branches; never use
host-wide prune. See the branch workflow and bounded cleanup procedure.

Owner direction, 2026-10-20: all non-promotion MRs targeting `staging`, including
feature/Goal, maintenance and repair MRs, default to auto-merge after required
Codex review, applicable Duo review/findings disposition, rebasing/input
reconciliation, required pre-merge evidence and other recorded holds are clear.
This is standing integration authority; do not ask for another routine merge
prompt. Early/draft MRs retain the intent but cannot arm auto-merge until every
hold clears. An explicit owner hold or opt-out for a specific MR takes
precedence. This does not apply to staging-to-main/RC promotions, final release
tags, or final Goal ACCEPT/post-integration qualification.

For local session discovery, progress checks and authenticated Codex → Claude Code messaging, use [Claude session handoffs](docs/agents/CLAUDE-SESSION-HANDOFF.md). Socket delivery is not acknowledgment or execution approval; verify the intended session and preserve its bound checkout/inputs. The runbook records the version-specific local interface and does not expand task authority.

Owner clarification, 2026-10-10: Codex's final pre-integration code-review
acceptance authorizes the implementer to push the reviewed temporary feature
branch and open its MR to staging under the existing branch workflow, without
another owner permission question. This conditional standing authority does not
block approved implementation while code review is pending. The 2026-10-20
staging auto-merge rule above now supplies merge authority after all its stated
gates clear; changes after review still require reconciliation. Final Goal ACCEPT
still follows the applicable staging disposition.

1. Codex selects a dependency-ready goal from the roadmap and writes its Goal and Work Spec. Major product/architecture questions go to the product owner.
2. Claude inspects the repository and writes the goal-specific Plan, unless the Goal explicitly authorizes the direct-execution tier below. Otherwise Claude does not implement until Codex explicitly approves the Plan.
3. Codex reviews the plan against scope, dependencies, architecture, security, budgets, and every acceptance criterion. Review uses exactly one outcome below; plan ACCEPT authorizes execution, not implementation completion.
4. Claude implements the approved plan, validates, self-reviews, and records the Result within the execution policy and budgets. Escalation conditions take precedence over autonomy.
5. Codex reviews the Result, actual diff, tests, tenancy/isolation evidence, and architecture conformance. Passing tests alone is insufficient.
6. Codex accepts, requests bounded corrections, or escalates. After acceptance, Codex updates progress and identifies the next dependency-ready goal.

### Bounded specification authority and owner stop boundaries

Owner direction, 2026-10-14: automatic continuation does **not** authorize Codex
to create the next Goal/Work Spec. Without explicit recorded specification
authority, stop and report a missing/draft Spec; do not expand the roadmap loop.

The owner may authorize automatic specification through a named inclusive
Goal/Spec endpoint, an explicit task set/range, or an entire named epic. Record
the actual owner instruction/reference, allowed phase(s), starting point, exact
included task IDs and inclusive/exclusive endpoint in PROGRESS and the existing
supervisor handoff when applicable. Pin the roadmap commit/membership: dependency
order governs, not numeric IDs, and later inserted Goals do not silently enlarge
the mandate. A task-range instruction authorizes only its stated phases;
specification-only wording is a phase stop, not permission to start planning or
implementation merely because a new Spec now exists. Separately authorized
in-flight tasks are unaffected. A full-epic
instruction may explicitly include missing Goals/Specs and the normal delivery
loop, with separate bounded Goals, reviews, budgets and acceptance throughout.
No new authorization registry or unbounded "finish the roadmap" inference.

An **independent owner stop boundary** overrides automatic specification,
planning and execution, even when a ready Spec or approved Plan already exists.
Check it before each next-task dispatch, including advance planning while the
current task's CI runs. Examples (not active authorizations):

- "Auto-spec through G013 inclusive" permits the stated specification range;
  stop authoring at that endpoint. Other phases need their own applicable authority.
- "Finish EPIC-004, including its missing Goals/Specs, then stop" permits the
  recorded epic membership under normal gates; do not start the following epic.
- "Stop before G013" excludes G013's specification, planning and execution.
- "Finish G013, then stop" includes its acceptance/closeout and excludes all
  later task activity, even speculative planning with an existing Spec.
- A phase-specific instruction such as "plan G013, then stop" stops at that
  phase, not after implementation.

Resolve "stop at"/"the next task" against current active work and record the
interpretation. If before/after or task identity is ambiguous, hold before
crossing the possible boundary and clarify; do not assume permission to finish
the boundary task. A new stop cancels queued continuation beyond it; pause
already-running out-of-bound work at a safe point, preserving artifacts and
owned-resource cleanup. "Stop now" retains immediate-stop semantics.

The earlier/more restrictive authorization limit or stop wins until the owner
explicitly changes it; record revocation/resumption without erasing history.
Missing prerequisites outside the mandate, owner decisions, budget exhaustion,
native permissions and G021 still stop affected work. This policy defines how
to record future mandates; it authorizes no new task range or epic by itself.
Do not bypass a stop by selecting a later independent task or extending the
recorded endpoint; reopening the boundary needs explicit owner authority.

<a id="automatic-planning-after-integration"></a>
### Automatic planning at implementation MR creation

Owner clarification, 2026-10-15: when a predecessor Goal/task's implementation MR
is first opened against `staging`, Codex verifies the MR and automatically starts
planning the earliest eligible next roadmap task with a ready supervisor-owned
Goal/Spec. A draft implementation MR qualifies; a planning-only, placeholder or
probe-only MR does not. Do not wait for MR fast/Duo, ready status, merge, staging
CI, final acceptance or cleanup merely to start this provisional planning.
Verify the actual MR, source/target and candidate SHA, existing planner ownership,
and recorded owner work/stop bounds. Resume the assigned planner; do not duplicate
or restart it for later pushes. Missing Specs require an existing explicit bounded
specification mandate, not scope invented by an implementer.

Record the MR/candidate baseline and pending dependencies/decisions in PROGRESS
and the planning handoff. Planning is provisional: reconcile relevant MR revisions
and the integrated predecessor before execution under the release policy below. A closed/reworked predecessor
MR requires readiness reassessment, not automatic execution. Specific owner stops
and unresolved planning prerequisites remain binding. This earlier trigger
supersedes the prior integration-only trigger; integration is a fallback readiness
checkpoint if initial MR discovery was missed, not another dispatch.

Use the verified coordinator, one task owner, current model-aware planning effort
and existing worktree/naming rules. Planning permits inspection and Plan/history
artifacts only, not application/test implementation, installs, builds, shared-test
resources, new infrastructure or unauthorized probes. Implementation still requires
the release policy below, Plan ACCEPT/current reassessment, bindings, budgets,
owner decisions, resource readiness and recorded supervisor release. No new watcher,
manifest or permission authority is introduced.

### Automatic reassessment and execution release

Owner direction, 2026-10-17, superseding the final-ACCEPT-before-start default
of 2026-10-14. Default: once a predecessor's source has passed required pre-integration review
and is merged into the integration branch, reassess the next authorized task's
approved provisional Plan against that exact integration and current Goal/Spec.
Implementation may overlap the predecessor's running staging/integration checks
and closeout; do not wait for final predecessor ACCEPT solely to start it.
Record the merged revision, pending qualification and affected dependency in the
existing binding/Result and PROGRESS. A merge or Plan existing alone is not release.

Before release, verify Plan ACCEPT (or explicit direct-tier authority), current
input/scope binding, owner decisions, budget, isolated task branch/checkout and
resource readiness. Other required dependencies must be satisfied. For material
changed assumptions, request bounded implementer-owned Plan updates and review;
otherwise record reconciliation without rewriting immutable approved Plan bytes
or repeating unchanged validation. Record and relay supervisor execution release;
no additional routine owner proceed prompt is needed. Reuse the task owner and
current model-aware effort, with no budget reset.

Predecessor final ACCEPT and its required staging disposition still gate the next
task's integration and final ACCEPT. Pending or failed checks remain pending or
failed. Unresolved security, architecture or contract changes that affect the next
task block its affected implementation. Explicit task/owner holds, manual gates,
native permissions and delivery barriers override this default; do not silently
rewrite an active Goal/Spec's stricter binding or expand a phase mandate.

If predecessor qualification fails, diagnose attributable evidence and pause only
affected or unsafe dependent work; preserve unaffected work. The predecessor's
assigned implementation owner retains responsibility for diagnosing/fixing its
staging failures, on its repair branch and existing attempt ledger. Run those
repairs in parallel with successor implementation; do not transfer them to the
successor's workers by default. Any worker reassignment must be explicit and keep
predecessor ownership, exact repair scope and focused validation. Keep repair and
successor changes attributable, prevent concurrent writers/shared-resource clashes,
and reconcile changed interfaces before dependent work resumes. No silent scope
expansion, automatic retry campaign or failed-to-pass conversion.

Planning still starts at the first eligible implementation MR; implementation
starts only after reviewed source integration and recorded release. No new watcher,
scheduler, artifact framework or blanket specification authority is introduced.

Accepted Plan/history import and the G021 foundation barrier remain mandatory.

### Accepted Plans belong in the local integration branch

Owner direction, 2026-10-09 UTC: when accepting a Plan, including a provisional
design acceptance, Codex imports the exact reviewed Plan into the main local
integration checkout (`staging` by default). Import its required planning
inventories, referenced revision history and Codex review record together, and
update PROGRESS.md. Verify imported Plan/inventory bytes against the reviewed
SHA-256 values; archive any superseded local revision before replacing it.
Do not leave the accepted Plan only in an isolated planning worktree.

Import management artifacts only; preserve the source/execution checkout and
its bound inputs. Record outstanding execution gates explicitly: import does
not grant implementation acceptance, release provisional dependencies or
rebind execution to the integration branch. Commit/push follows the owner's
authorization; feature implementation integration retains the MR workflow.
If an import conflicts with local edits, preserve both versions and resolve
the conflict before reporting the acceptance handoff complete.

Do not silently expand a goal. Changes to Goal/Spec remain Codex-owned and require explicit review; major product/architecture decisions remain product-owner-owned. No major infrastructure dependency or vendor may be introduced without an accepted ADR and the necessary reviewed scope/plan update.

## Implementation-round safekeeping pushes

Owner direction, 2026-10-11: after every completed implementation round, including
correction rounds, Claude commits the scoped work and pushes a checkpoint to its
temporary task branch on origin. This is standing authority for safekeeping
before final code review or validation completion. Record the commit, round and
actual validation status in the Result; pending/failed checks remain explicit.
Do not wait for Codex acceptance solely to back up completed work.

Inspect staged paths/content and credential exclusions before publication; do
not include secrets, machine-local session records, raw evidence or other tasks'
edits. During active immutable validation, push the existing committed checkpoint
without changing bound files. Use normal pushes; an approved rebase may use only
the existing exact-tip lease procedure, never unconditional force or rewrites of
long-lived/protected branches. Stop on unexpected remote changes.

A checkpoint push does not open an MR, authorize integration, mark the Goal
accepted or waive fetch/rebase/binding/manual/test gates. Final reviewed MR
readiness retains the existing workflow. Temporary branch pushes without an MR
remain covered by the existing duplicate-pipeline suppression; an existing MR
may run its automatic checks when its source advances. Do not launch extra
pipelines or full validation merely to publish a backup.

## Goal format

Every future goal file must contain:

1. Goal ID and title
2. Objective
3. In Scope
4. Out of Scope
5. Dependencies
6. Completion Definition
7. Execution Policy
8. Iteration Budget
9. Context / Scope Budget
10. Suggested Subagents / Delegation
11. Stop / Escalation Conditions

Codex states the goal's allowed autonomy, any explicit budget overrides, and stop conditions. Existing goals are not silently rewritten by governance changes; their revision remains a separate authorized planning action. Repository execution defaults apply unless an explicitly reviewed goal overrides them within the canonical constraints.

## Execution policy

After plan approval, Claude may autonomously inspect the repository, implement the approved plan, create/modify tests, run builds/lint/type checks/tests, diagnose failures, iterate, and refactor within approved scope. Inspection and plan preparation are allowed before approval; implementation is not.

Claude must stop affected work and escalate to Codex for:

- Architecture contradictions or conflicts with docs/ARCHITECTURE.md or an accepted ADR.
- Scope expansion.
- Security-model changes.
- Destructive or risky migrations.
- Public API/MCP contract changes.
- New major infrastructure or vendors.
- Irreversible data-model decisions.

Identify such changes during planning and obtain the required explicit review before execution. A plan approval is not blanket authorization for unreviewed material changes discovered afterward. Codex must also stop and escalate major architecture/product decisions to the product owner. Neither agent may silently make them.

### Autonomous mechanical corrections

Owner direction, 2026-10-17: after task execution release, Claude directly fixes
mechanical inconsistencies within approved scope, exact file allow-list and budget:
stale test selectors/fixtures, approved-value catalog parity/count/version pins,
missing imports and harness file lists. No extra approval or full Plan revision
for an unambiguous implementation consistency fix. Diagnose the cause, preserve
assertions/negative controls, run focused affected checks, record evidence and any
Plan wording reconciliation in the Result, and continue the current worker.
Immutable Plan/history and frozen running inputs stay intact. Codex reviews the
actual completed diff/evidence before integration/final acceptance as required.

Scope expansion, ambiguity, security/authority/tenancy or public contracts,
unapproved schema/constraint/grant/data or risky migration changes, major vendors,
budgets, explicit holds and native refusals still escalate. Never weaken a control
or alter expected values to conceal a regression. This does not grant a full-gate
retry, new test campaign, merge permission or new specification authority.
Full portable rule: [mechanical corrections](distributions/skills/restruct-sdlc/sdlc-management/references/roles-lifecycle.md#autonomous-mechanical-corrections).

### Risk-based execution for future Goals

Owner direction, 2026-10-06: as work moves from foundation engineering to
surface features, reduce process overhead in proportion to risk. Codex states
the execution tier in each future Goal's Execution Policy; agents may not
self-classify a task to bypass its approved gates. Keep required Goal/Spec
fields concise, with measurable outcomes and scope boundaries. Do not generate
lengthy artifacts solely to reproduce unchanged architecture or controls.

- **Direct execution:** bounded copy, styling and presentation changes within
  existing components and approved behavior may execute from an explicitly
  authorized Goal/Spec without a separate Plan file or another Plan-approval
  handoff. Record source/baseline, changed files, focused relevant checks and
  brief Result evidence; Codex reviews the actual diff before acceptance.
- **Ordinary feature execution:** use a concise Spec and short Claude Plan
  covering concrete changes, criterion mapping, validation and actual risks.
  One approval handoff authorizes autonomous implementation within that scope;
  routine choices do not require repeated permission or replanning.
- **Architecture/security/data execution:** preserve explicit Spec/Plan review,
  escalation and stronger required evidence for identity/authentication,
  authorization/tenant isolation, security-model changes, migrations or durable
  identity decisions, public API/MCP contracts, major dependencies/vendors and
  infrastructure. A UI for these behaviors is not automatically low risk.

Direct execution does not cover new schema/data mutations, authority/session
behavior, contract changes or unreviewed dependencies. If a presentation task
crosses those boundaries, stop the affected work for scope/tier review. For all
tiers, retain accessibility/security requirements, exact baseline/input binding
through accepted procedures, budget/stop rules, meaningful validation and the
accepted CI trigger matrix. Test scope follows actual changed behavior; do not
create duplicative control suites or remove required full coverage.

This policy applies through explicitly authorized future Goals. Existing
Goal/Plan bindings and dependencies are unchanged unless separately revised;
G003/G048 execution gates and the G021 product-feature barrier remain intact.

## Claude effort selection and operator reminder

**Current model-aware override (2026-10-15):** use the table in the model-selection
section below; it supersedes older blanket planning-xhigh wording for Fable.

Owner direction, revised 2026-10-11: Claude planning defaults to **xhigh**;
implementation defaults to **high**. The owner/operator may explicitly choose
**xhigh** for an implementation phase that needs deeper reasoning. This supersedes
the earlier implementation-xhigh default, including subsequent work on active
G007/G051/G052; preserve historical Plan and evidence bytes. Do not interrupt a
running validation batch or restart completed work solely to change effort.

Owner clarification, 2026-10-12: the persistent Claude coordinator defaults to
**high** for orchestration, status and handoffs. Planning workers default to
**xhigh**; implementation workers default to **high**. Set worker effort
explicitly through supported dispatch/session controls where available, rather
than assuming the coordinator's setting supplies the planning default. For
complex debugging or security decisions, the owner/operator may selectively
raise the affected work to xhigh; this changes reasoning effort, not authority.
If the coordinator itself performs planning, apply the planning-phase default.
Do not change a running session or interrupt a batch merely to enforce these
documented defaults; verify and report actual configuration as below.

Before starting a planning or implementation phase, state the intended effort
and remind the operator to select it in the active Claude session, or ask their
preference when unresolved. If the owner already specified the effort for that
phase, acknowledge it without requesting the same decision again. A reminder
is not an additional execution approval or a reason to stop already authorized
work; use the phase default unless the owner chooses otherwise. Existing Plan,
input-binding, budget and escalation gates still apply.

Verify actual runtime effort where the CLI/session metadata permits it,
including inherited subagent effort; do not claim that a handoff's wording
sets or enforces the runtime setting. Record unverifiable effort as unknown.
If a mismatch is found, notify the operator and correct subsequent work under
the supported session controls, preserving completed work and review history.
Do not restart a completed Plan or invalidate passing evidence solely because
its historical effort was lower. No model switch, permission change or budget
extension is implied by effort selection.

### Model selection — owner direction, 2026-10-15

Default to **Fable 5.1 high for planning and test corrections**, including
failure diagnosis and corrective test/fix work. Default to **Opus 5.5 high for
implementation and routine test execution/reporting**. Default the Claude
coordinator to **Sonnet 5.5 high**; the owner may choose **Opus 5.5 high** for
coordinator work. The owner may explicitly choose Opus for planning or test corrections; use xhigh
for those phases. Fable xhigh is selective for especially difficult work. These
model-selection defaults supersede earlier wording that limited Fable preference
to complex cases. They are workload recommendations, not a measured claim of
superiority or faster Docker/build/test runtime.

Owner refinement, 2026-10-15: use model-aware effort defaults:

| Selected model / phase | Effort |
|---|---|
| Fable 5.1 planning, review, diagnosis or corrections | high |
| Fable especially difficult architecture/security/concurrency work | xhigh selectively |
| Opus 5.5 planning | xhigh |
| Opus 5.5 test-failure diagnosis and corrective test/fix work | xhigh |
| Opus routine implementation, ordinary test execution/reporting | high |
| Sonnet 5.5 coordinator (default) | high |
| Opus 5.5 coordinator (owner-selected option) | high |

This supersedes the earlier blanket planning-xhigh rule for Fable and the generic
implementation-high rule when Opus is doing test corrections. Before the next
phase, state the proposed model/effort and honor owner choices; configure through
supported controls and verify actual values when available, otherwise report unknown.
A policy note or handoff does not change an existing worker. Do not interrupt a
batch, restart completed work, reset budgets or weaken review/tests to switch
models. Preserve sessions, checkouts, bindings and native permissions.


Fable test-correction workers may escalate from high to xhigh when diagnosis is
uncertain, a fix fails again, or the correction becomes substantially more
complex. This owner-authorized effort increase needs no repeat approval. Apply
through supported controls at the next safe boundary and record the reason and
actual setting (or unknown) in the Result. Do not interrupt a running batch.
Existing repeated-issue escalation and budget stops still apply; higher effort
does not reset attempts or expand scope.

Current Claude examples follow [Anthropic model guidance](https://platform.claude.com/docs/en/models/overview);
reassess names/availability before future adoption. Other providers/harnesses may
use equivalent owner-selected tiers; no new provider dependency is introduced.

## Execution Budgets

All Goals and implementation agents must follow
`management/BUDGET_POLICY.md` (bootstrapped at adoption; see
[budgets](distributions/skills/restruct-sdlc/sdlc-management/references/budgets.md)).

Codex assigns the Goal's budget class when authoring the Goal.
Claude may recommend changes during planning but may not change an approved
budget without Codex authorization.

## Iteration budget

The default budget is **5 goal-level execution attempts** under management/BUDGET_POLICY.md. Within one attempt Claude may autonomously implement, loop through scoped fixes and test batches, and delegate approved work until the Goal is review-ready or the attempt fails. Routine failures, commands, batches and subagent loops do not each consume an increment. A retry after an unsuccessful attempt or a supervisor-authorized correction round consumes another attempt; record boundaries and outcomes. Preserve historical counts and explicitly amend active Goal budgets before applying this definition.

If materially failing after that budget, Claude stops, documents the blocker and attempted approaches, updates the goal-specific Result, and returns control to Codex. Do not reduce tests or validation to fit the budget. Codex reviews whether to correct, split, or revise the goal/budget; Claude does not reset or extend it unilaterally.

## Context / scope budget

Prefer goals that can be understood, implemented, tested, and reviewed in one bounded implementation cycle. If Claude concludes during planning that the goal is too large for reliable completion within available context, do not implement. Explain the proposed split in the Plan and wait for Codex review. Codex owns the resulting Goal/Spec/roadmap changes; a split cannot silently enlarge scope.

## Suggested subagents / delegation

Codex records Goal/Spec priority and may suggest useful subagent roles. Claude's
Plan targets 3–5 independent parallel implementation workers for high priority or
1–3 for low priority when feasible, with exclusive file ownership and integration
responsibility. Record constraints that reduce the allocation. Kickoff follows
the accepted Plan; planning itself may use one agent.

Suitable uses include repository exploration, test analysis, security review, dependency/library research, isolated implementation in non-overlapping files, and post-implementation review. Do not delegate product decisions, architecture authority, scope changes, overlapping edits, or final integration responsibility. Claude remains responsible for reviewing, validating, and integrating all delegated work. Delegated opinions do not replace Codex acceptance or product-owner decisions.

## Work Spec format

Every Work Spec contains requirements, architecture constraints, security invariants, expected artifacts, acceptance criteria, required tests, required validation, and explicit non-goals where ambiguity is likely. Specs define **WHAT must be true**, not low-level **HOW to code it**. Requirements must be assessable and preserve canonical invariants; implementation choices belong in Claude's Plan.

## Implementation Plan format

Claude's Plan contains:

- Implementation approach.
- Files/modules to create or modify.
- Database/schema changes, including migrations and risk handling (or explicitly none).
- Contracts/interfaces, including public API/MCP implications.
- Test strategy and required validation.
- Goal/Spec priority and parallel implementation allocation by phase: target 3–5
  workers for high or 1–3 for low when feasible, with ownership, dependencies,
  resource limits and constraints reducing the allocation. The planner may remain one agent.
- Implementation sequence.
- Mapping to each acceptance criterion.
- Risks/questions requiring escalation, including scope/context concerns.

## Proportional planning and reuse of accepted controls

Scale Plan detail and review evidence to the Goal's scope, risk and unresolved
decisions. A narrow change needs its concrete changes, acceptance-criteria
mapping, validation commands and genuine prerequisites. Unchanged architecture
and accepted execution safeguards may be referenced by path/version; do not
rebuild or duplicate them for each Goal.

Reuse the repository's accepted harness, approved-input wrapper, isolation and
cleanup procedures. For a demonstrated gap, propose the smallest correction
that preserves the required guarantee. A new execution framework, resource
registry or generalized orchestration library is additional engineering scope,
even when placed outside the repository; it needs an explicit scope decision
rather than being an incidental prerequisite for a UI or other narrow Goal.

Owner direction, 2026-10-15: planning must inspect relevant existing tests and
cite reused/extended/new coverage in the existing Plan/AC mapping. New tests need
an actual gap or distinct changed boundary; do not duplicate suites, fixtures or
accepted invariants simply for a new Goal. Codex reviews this justification.
See [coverage reuse during planning](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#reuse-existing-coverage-during-planning);
required regression, security and selected local/hosted coverage remain binding.

Before approval, specify test scenarios and the evidence they will produce.
Implement those tests and collect required application evidence during approved
execution. A bounded disposable planning probe is justified when it resolves a
concrete uncertainty affecting feasibility or approval; it must not become a
parallel implementation of the future validation system. Do not demand new
pre-approval demonstrations of unchanged, accepted controls without an
identified defect or changed assumption.

Codex correction requests must name the finding, required guarantee and minimum
validation. Use the [compact repair handoff](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#compact-repair-handoff)
for failures: reference exact revisions/diff, focused evidence, attempted fixes,
current hypothesis and next smallest test in the existing Result/review context. If process work starts dominating the requested outcome, Codex must
reassess and narrow the procedural request, or propose separately scoped work.
Preserve review history and record which directions are superseded.

### Context and token discipline

#### Temporary local evidence

Owner direction, 2026-10-14: use committed Results/reviews and existing CI artifacts
for durable evidence. Local raw output is temporary during active diagnosis/review;
purge closed-task bulk and duplicate downloads, preserving compact outcomes and
explicit necessary-proof exceptions with release conditions. Artifact expiry is
recorded honestly, not a reason to mirror every run. Session/binding/observer state
is separate. Follow the [evidence cleanup policy](distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#local-evidence-is-temporary--owner-direction-2026-10-14).

#### Git-first provenance for new work

Owner direction, 2026-10-14: Git commit and path are the default binding for
committed implementation and reviewed management inputs. Keep a compact Result
with tested candidate/target commits, commands, outcomes, runtime/tool/image
references and CI links. Identify relevant dirty/untracked changes; prefer scoped
committed checkpoints before final qualification. Extra snapshots/hashes cover
only inputs Git does not capture. Do not duplicate committed source with per-file
inventories or retained whole-checkout copies by default. Preserve necessary proof,
secret-safe handling and failed history; purge disposable output at closeout.
Existing approved snapshots/wrappers and exact bindings, including active G008,
remain required until an explicit reviewed amendment. No replan, relaxed acceptance
or security/CI/resource gate follows. See
[Git-first provenance](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#git-first-provenance).

Owner direction, 2026-10-09: minimize repeated context loading and procedural
handoffs. After required initial reading, review changed sections and unresolved
findings; re-read unchanged sources only for a relevant change, contradiction or
lost context. Reference accepted controls by committed path/version rather than
reproducing their text or reopening settled review findings. Mechanical execution
binding uses the accepted procedure, not another design-review cycle unless a
substantive defect or changed assumption is found.

Keep status checks to bounded relevant transcript/log excerpts and handoffs to
the outcome, exact input references, next action and remaining gates. Store
durable decisions and readiness state in existing review/Result/PROGRESS records
so a fresh session can resume without replaying conversation history. Avoid
duplicate summaries and raw terminal redraw output. This changes communication
and reading overhead only; required approvals, input bindings, validation,
budgets, escalation and acceptance evidence remain unchanged. Do not refresh an
active execution checkout's bound governance inputs solely for this note.

### Keep CI changes separate and bounded

After the foundational CI profile is accepted, future Goals default to reusing
it. Add or adjust tests for the Goal's changed behavior and, where necessary,
make the smallest change to the existing job/path mapping. Do not bundle runner
migrations, harness redesign, new execution controls, generalized orchestration
or broad test-framework work into an ordinary feature/UI Goal. A substantial CI
change requires its own bounded Goal/Spec and explicit approval; the presence of
a possible optimization is not a prerequisite to unrelated implementation.

CI gates applicable merges/releases. Independent approved implementation may
proceed in isolated checkouts when its declared dependencies permit it; avoid
inventing a blanket dependency on all CI improvements. Explain any proposed CI
blocker by the specific required guarantee or unavailable dependency. Existing
Goal dependencies remain binding until explicitly revised; this policy does
not release G003/G048 or the G021 foundation barrier.

Use the accepted fast/affected/full trigger matrix. Run full when required,
reuse passing evidence only for its exact tested content/validation inputs,
and repeat qualification or full validation only for a relevant change,
failure or unresolved guarantee. Do not start an additional benchmark or
unchanged-control proof campaign during a feature Goal. Report timing/cost
observations as follow-up proposals rather than silently expanding the work.

Proportional planning alone does not waive applicable Plan approval; a future
Goal's explicit direct-execution tier may omit the separate Plan as above.
Neither policy waives exact baseline/input binding,
fail-closed validation, resource protection, acceptance evidence, security
invariants, budgets, escalation or the G021 foundation barrier. Simplifying a
procedure must retain those guarantees and every applicable acceptance criterion.

### Validation batches: finish, fix, rerun, then advance

Owner direction, 2026-10-10: organize tests/validation into bounded, named batches
using the existing suites and controls. Plans identify batch order, selection and
evidence; do not build a new runner or split an indivisible required suite solely
to satisfy this rule.

Within a batch, ordinary test failures may continue to collect other valid failures
and results. Do not start a dependent batch while its prerequisite failures remain
unresolved; independent safe batches may run in parallel under the scheduling rule
above. At the batch boundary, preserve its complete outcomes and verify owned
process/resource/lock cleanup. Gather all outstanding failures in that dependency
chain, diagnose and fix them, then rerun failed cases/controls and checks
affected by those fixes. Advance only after those failures are resolved and the
required reruns pass. Repeat this finish → diagnose/fix → rerun → advance sequence
at every subsequent batch. Do not restart unrelated, already-passing batches;
reuse them only with exact content/input provenance. Keep an outstanding-findings
checklist in the existing Result so an earlier failure cannot be forgotten.

Continuing after an ordinary assertion failure does not mean continuing unsafe or
invalid work. Resource/security failures, invalid prerequisites, corrupted evidence,
required fail-closed gates or an existing stop/escalation condition stop the affected
work immediately. Dependent checks that cannot run meaningfully are recorded as
blocked/skipped, never passing. Independent batches may continue only while their
inputs, ownership and resources remain valid; do not bypass a failed prerequisite.

Before each rerun or later batch, follow fetch/rebase/input reconciliation and the
Docker Desktop safeguard. Record failed/interrupted and rerun evidence honestly.
Fix/test inner loops normally share the current Goal execution attempt; batching does not
reset the allowance or independently consume another attempt. Scope or budget blockers return to Codex.
Mandatory final regression/qualification and hosted gates remain required after fixes
are stable; this sequencing rule does not reduce coverage or change CI triggers.
It applies to active and future local execution without rewriting historical Plan
or snapshot bytes; reconcile its sequencing amendment through current bindings.

### Proportional UI validation and focused debugging

For future UI Goals, choose validation from the changed behavior and shared
boundaries. Ordinary page/copy/layout changes default to focused component or
browser checks, relevant accessibility/responsive cases and the accepted
fast/affected CI profile. Shared components, global tokens/themes/fonts,
portals, first paint, authentication or other security boundaries warrant wider
regressions. Do not copy G048's full validation matrix into every later screen
change merely because it is UI work.

Current CI limitation, recorded 2026-10-08: `infra/ci/profile-paths.tsv`
classifies all changes under `web/tests/browser/` and `web/tests/fixtures/` as
foundation/full for applicable gate pipelines, even when the change tests one
ordinary feature. Focused local debugging does not override that mapping. Each
future Plan must state both its focused local checks and the actual CI profile
selected by its changed paths; do not promise affected-only hosted validation
when the map requires full. Separating feature-specific browser coverage from
shared foundation/security coverage is a bounded follow-up for the project
backlog, not an implicit CI change or a prerequisite to independent feature work.

When a browser case fails, inspect its trace/logs and reproduce that scenario
with targeted tests first. Retest the fix and relevant adjacent cases before
another broad run. Shared harness edits still require their covering checks.
Once fixes are stable, run the Goal's required final regression and hosted
checks against the final tree. Repeat a full run only for changed relevant
inputs, a failure or an unresolved guarantee; a focused pass does not replace
mandatory acceptance evidence. Record failed and interrupted runs honestly.

This is guidance for future scoping and the debugging inner loop. It does not
change existing approved Goal/Spec/Plan matrices, bindings or CI triggers, and
does not waive accessibility, security, required final validation or G021.

### Future-goal overhead reduction — 2026-10-14

Owner authorizes direct adoption from G009 planning onward. Before each batch,
fetch and assess the exact delta for effects on validation (including indirect
check/build/runtime/fixture/evidence dependencies) and execution authority,
dependencies/governance/approved bindings. If demonstrably irrelevant to both,
record the compared SHAs and reason in the existing Result and defer rebase/input
reconstruction. Relevant or uncertain changes reconcile normally. Always reconcile
before final qualification/MR; rerun affected checks only and retain resource gates.
This narrowly supersedes the unconditional pre-batch rebase rule below for new
work; G008 and existing approved bindings keep their current requirements.
Committed inputs use Git references; unconsumed progress documents need no extra
snapshots. CI meta-tests test the harness and are selected for CI/harness/control
dependency changes or unknowns, rather than every application/migration change.
Required application/security coverage and actual hosted path selection remain.
No new spec/Plan or tests are required for this owner-directed one-off; its code
is pushed untested at the owner's instruction, not represented as verified.

### Temporary solo-operator validation and integration policy

Owner direction, 2026-10-09: select required coverage from the accepted
path/behavior matrix; ordinary changes retain focused coverage. Complete
required manual feedback and self-review before final qualification. Preserve
exact tested content, validation inputs, statuses, resource cleanup and
failed/interrupted evidence. Do not repeat unchanged local coverage merely to
collect a second identical pass. The full-suite placement is governed by the
current local-after-merge clarification below; this supersedes the original
pre-MR full-coverage placement.

Owner clarification, 2026-10-09: **fetch and rebase before every local test or
validation batch on a temporary feature/Goal branch**, not only before final
qualification or MR opening. This includes focused development tests, builds,
lint/type checks, integration/browser tests, controls and broad/full runs; calling
a run preliminary, cycle-1 or background does not exempt it. Before launch, fetch
the remote target (normally origin/staging), verify that its fetched SHA is an
ancestor of the candidate HEAD, and rebase if it is not. Resolve conflicts and
reconcile approved-input bindings before tests; record target/candidate SHAs and
binding provenance with the batch evidence. If the target has not advanced, do
not perform a redundant rebase. A single batch may contain multiple commands;
check again before launching a later batch, not between commands in one batch.

Read-only Git, status, input-integrity and resource-safety inspections needed to
prepare a safe rebase are prerequisites, not application test batches. If dirty
work or immutable bindings prevent a safe rebase, preserve the work and reconcile
under the approved procedure or escalate; do not start tests against the old base.
Do not refresh bound inputs silently or rewrite long-lived main/staging/identity-
qualification branches. Hosted CI validates its exact submitted commit and must
not rebase it during a job.

The default order is implementation → fetch/rebase/reconcile → focused local
checks and required manual feedback/self-review → fetch/rebase/reconcile again
if needed → push/open MR → hosted fast checks and required supervisor/Duo review
→ authorized merge → affected local qualification on the exact merged integration
SHA. The full suite is not a pre-MR gate under the default `local-after-merge`
placement. An explicitly configured alternative placement governs only when the
adopting owner selected it. Already-running batches retain honest historical
evidence; this clarification does not cancel them or make their results final.
Reuse unchanged passing evidence only with explicit content/input provenance; a
changed SHA alone does not require a full rerun.

Check the remote target again before MR opening/integration. If it advanced,
refresh/rebase and reconcile again, then rerun only checks affected by the new
content or validation inputs. A changed commit SHA alone does not require
repeating the entire suite: reuse unchanged evidence only with explicit content
and input provenance. Relevant changes also require affected manual retests. MR
fast checks remain automatic. No automatic hosted full staging pipeline runs
after merge under the current placement; required RC promotion MRs retain their
separate full hosted qualification.

Codex reviews the actual diff and passing local/MR evidence before integration.
Merge to staging is conditional integration, not final Goal ACCEPT. After merge,
the existing Claude coordinator automatically owns affected local qualification of
the exact merged SHA; record its actual result and reconcile it to that content.
Only after required qualification and the other Goal gates clear may Codex mark
the Goal ACCEPT/completed and integrate dependent Goals. Isolated dependent
implementation may overlap the local run under the automatic execution-release
policy. Any failure blocks dependent integration; pause affected work until
resolved, preserving unaffected work. Main promotion requires the applicable
staging disposition and retains its separate checks.

The 2026-10-09 post-merge hosted-pipeline status rule is historical and does not
replace the current exact-merged-SHA local qualification. A canceled or green
fast pipeline is not evidence that affected local qualification ran. Record every
pipeline truthfully; canceled/unrun tests are never reported as passing.

The 2026-10-22 affected-post-merge policy supersedes the blanket full-local-run
default of 2026-10-19 and the 2026-10-09 temporary hosted-pipeline placement. Preserve scopes, budgets, distinct
local/identity/operator evidence, security guarantees, owner merge authority and
G021. Protected identity qualification, explicit web/API runs and schedules
remain separately governed.

Owner direction, 2026-10-10: Docker Desktop on WSL never releases a bind-mounted
host path until WSL shuts down. Leaked paths fill its VM mount table, and then
every container start fails with a misleading "no space left on device". See
[OPERATIONS.md, Docker Desktop bind-mount leak](docs/foundation/OPERATIONS.md).
On Docker Desktop/WSL, run `bash infra/ci/docker-desktop-mounts.sh` with the
fetch/rebase step before each local validation batch; `validate-controls.sh`
runs it itself. If it fails, do not start the batch. Copy any evidence that is
still needed out of `/tmp`, then recover with `wsl --shutdown` as documented.
Harness and test code must not bind-mount a new host directory per run or per
case. Stream inputs and outputs through stdin/stdout or use a fixed path. This
check is an operator-environment safeguard, not an acceptance criterion.

### Codex assessment of staging advances

Owner direction, 2026-10-10: whenever Codex integrates changes into staging, or
discovers an externally advanced staging tip, assess the delta against every
known active Claude planning, implementation and validation task. Include direct
staging commits as well as merges. Codex owns this integration assessment; do not
leave collision detection solely to the implementer's next fetch.

Compare old/new staging SHAs and changed behavior, shared files, dependencies,
contracts, validation inputs and approved bindings with each task's scope and
current phase. Record the disposition briefly in the existing handoff/Result or
PROGRESS: unaffected work continues; relevant nonurgent changes require a handoff
and reconciliation at a safe boundary; conflicts, invalidated assumptions or
unsafe/obsolete running validation require an immediate pause instruction.
Uncertain relevant impact requires inspection or a safe pause, not an assumed
exemption. A commit SHA change alone is not a reason to restart validation.

For interrupted tests, Claude stops and reaps owned children, preserves logs and
failed/interrupted status, verifies ownership-based cleanup and releases its run
lock before rebasing. Never edit/rebase a checkout underneath running tests.
Preserve dirty work, resolve conflicts and reconcile exact approved inputs before
resuming. Restart only checks affected by changed content/inputs or an unresolved
guarantee; reuse unchanged evidence only with attributable provenance. Planning
also reassesses changed assumptions before handback; original reviewed snapshots
and history are not rewritten.

Send instructions to the verified task owner (through the preferred coordinator
when established) and verify acknowledgment. Delivery is not proof of a halt,
cleanup or rebase. Existing pre-batch fetch/rebase rules remain mandatory even
when Codex assesses a delta as nonurgent. This is event-driven supervision, not a
new continuous watcher, permission to kill unrelated processes or automatic
hosted-pipeline cancellation. Architecture/scope conflicts still escalate.

### Manual testing and operator feedback before final CI

Owner direction, 2026-10-08: every Plan requiring manual testing or operator
feedback must explicitly sequence that work before final CI qualification.
First complete the focused local checks needed to make the candidate and review
environment ready; then conduct the required manual tests and obtain the required
operator feedback; resolve findings and repeat affected checks/reviews; finally
collect required local automated evidence before the MR and applicable hosted
evidence under the temporary solo-operator policy above.
This applies to all Goals and Plan types, including browser testing, usability
review, operational exercises and other required operator feedback, not only UI
Plans or G004. Do not launch final qualification while required feedback is pending.

The Plan identifies manual/feedback stages, their readiness conditions, expected
evidence and how findings are resolved before final qualification. Bind test and
feedback evidence to the exact application, configuration and observer/test
inputs. A later relevant change requires repeating the affected tests/reviews;
unchanged evidence may be reused only under the accepted provenance procedure.
Do not start operator timers until the operator explicitly confirms readiness.
Record the sequence and any failures in the Result.

If a manual test or feedback stage requires a CI-produced artifact or hosted environment, identify
that dependency in the Plan and run only the prerequisite CI work first; explain
the exception rather than silently postponing required manual testing or feedback. Preserve
mandatory final CI, accepted triggers and existing input bindings. This rule does
not authorize skipping or cancelling automatic pipelines, changing permissions
or rewriting historical approvals. Reconcile an existing bound sequence through
an explicit supervisor amendment; G004 C7's owner-authorized sequencing amendment
is recorded separately.

## Result format

Claude's Result contains:

- Status, distinguishing submission/readiness from Codex acceptance.
- Implementation summary and files changed.
- Migrations/schema changes.
- Contracts/API/MCP changes.
- Tests added.
- Exact commands/validation performed and outcomes; skipped/failed checks are explicit.
- Acceptance-criteria checklist with supporting evidence.
- Architecture deviations, including explicitly none when applicable.
- Unresolved issues, blockers, attempted approaches, and iteration/correction-cycle history.
- Suggested follow-up work, which does not authorize new scope.

## Codex review and correction loop

Owner direction, 2026-10-11: after completing any Plan, implementation, correction
or execution-binding review, Codex must notify the responsible Claude session
without waiting for a separate owner request. Record the outcome first, then send
a concise handoff referencing that record, the reviewed inputs, permitted next
action and remaining gates. Use the verified coordinator when applicable and
verify relay to the task worker. Check receipt and explicit acknowledgment;
report pending acknowledgment or an unavailable session honestly, without
automatically resending or creating another session. A review notification
conveys only the authority in that review; it does not authorize an unapproved
merge, scope change or permission change. Later commit/push updates should refer
to the same decision rather than repeat the review.

For each formal review, return exactly one outcome:

- **ACCEPT**
- **CORRECTION LOOP**
- **ARCHITECTURE DECISION REQUIRED**

For CORRECTION LOOP, provide a bounded list of exact corrections tied to requirements/findings and specify validation to rerun. Codex does not rewrite Claude's implementation. Claude fixes only the requested issues, reruns required validation and relevant regression checks, and appends correction evidence to the Result before returning control. A correction pass consumes a goal-level execution attempt within the remaining Goal allowance unless Codex explicitly assigns a separate allowance; it does not automatically reset the budget. Track correction passes and their attempt boundaries separately from inner debugging loops.

If the same material issue survives two correction cycles, Codex must assess whether the Goal or Spec is defective and whether splitting or revising it is required. Use ARCHITECTURE DECISION REQUIRED when appropriate and escalate major decisions to the product owner. Do not continue an endless correction loop or restart counters to avoid escalation.


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


## Low-overhead delivery defaults — owner direction, 2026-10-15

Apply the existing SDLC references for [proportionate checks, recurring failures and
policy-only publication](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#validation-economy-and-recurring-failures),
[automatic routine transitions](distributions/skills/restruct-sdlc/sdlc-management/references/roles-lifecycle.md#continue-routine-transitions-without-owner-nudges),
[concise messages](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#concise-operational-messages) and
[routine closeout receipts](distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#routine-closeout-receipt).
Continue already-authorized actions without owner nudges; preserve all actual
acceptance/security/scope/budget/native-permission gates. This amendment adds a
narrow verified workflow-prose CI-skip option, not a directory-wide CI exemption.
Active application qualification and historical failures remain required/recorded.

Container-produced retained artifacts must be operator-removable at creation.
Follow [artifact ownership](distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#container-artifact-ownership):
retain reviewed non-root execution, use existing transfer/copy-out paths and
verify ownership, failure propagation and normal cleanup without recurring sudo.

Owner note, 2026-10-16: assess coordinator history at task handoffs/closeout and
when old messages dominate context; use [safe context compaction](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#coordinator-context-compaction-at-safe-boundaries)
with current task/decision/evidence references preserved. Keep the same session,
watcher binding and safe running batches; no model switch, transcript purge or new
monitor is implied. Context size alone does not attribute allowance consumption.

Owner default, 2026-10-16: [targeted reruns and attributable coverage](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#targeted-reruns-and-attributable-coverage-default)
applies to active and future tasks. Reuse unchanged passing cases instead of
repeating whole batches/control suites; retain honest failed/canceled history and
actual required coverage. Review frozen source in parallel with independent checks.
Required hosted/atomic checks, cleanup and final acceptance gates remain; no new
aggregation framework or automatic passing status.

Owner clarification, 2026-10-16: apply [bounded worker context handoffs](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#worker-context-and-phase-handoffs)
when oversized obsolete history warrants them; preserve one write owner, exact
state/evidence and phase-specific effort. No mandatory fresh agent per phase,
full retest, budget reset or new tracking system. Report owner-only native
compaction/effort controls once; policy text does not perform those actions.

Owner default, 2026-10-16: worker checkpoint target is **150,000 active context
tokens**, with an upper handoff threshold of **200,000**, under
[the existing context policy](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#worker-context-and-phase-handoffs).
Prepare concise transfer state around 150k at a safe boundary; checkpointing alone
does not force immediate agent rotation or another review.
At a safe step/batch boundary, checkpoint and transfer concise state to a fresh
worker; no full-transcript fork, duplicate writer, retest or budget change solely
for rotation. Record a brief completion-imminent exception or unavailable context
measurement; coordinator identity/compaction remains separate.

Owner direction, 2026-10-16: [focused executable checks from the first run](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#focused-selection-before-the-first-run)
are the default for all local implementation/correction/final validation. Verify
what selectors actually run before dispatch; broad umbrella suites require a
recorded technical coverage reason and supervisor disposition before launch.
Existing reviewed justification suffices; no repeated owner permission loop.
Convenience or a missing narrow harness selector is no automatic broad fallback.
Reuse attributable evidence and preserve required hosted/atomic/security gates.

Owner direction, 2026-10-16: apply [early risk review and short correction loops](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#early-risk-review-and-short-correction-loops)
inside existing planning/review: changed trust boundaries first, risky functional
slice early, bounded caller/pin inventory and batched actionable findings. Preserve
focused validation, context handoffs and all gates; no new process or G016 restart.

Owner direction, 2026-10-23: strengthen the existing Plan with settled decisions,
unproven assumptions and the smallest early proof of the highest-risk assumption
inside approved implementation. Estimate coding effort separately from build/test
elapsed time and review/queue waits; record executable selectors, dependencies and
resource reservations using known timings or honest estimates/unknowns. Material
scope additions require a bounded impact delta and applicable approval. A failed
proof or material overrun needs a compact Result checkpoint and bounded correction,
not another reporting framework. Apply only to remaining work at a safe boundary;
do not restart active planning or validation. See the linked early-risk policy.

Owner direction, 2026-10-16: [coordinator checkpoints and retirement](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#coordinator-checkpoints-and-retirement)
uses a 150k active-context checkpoint and 200k compaction-or-handoff threshold,
not a hard kill. Prefer compaction; retire at a safe boundary for persistent
context/decision/session problems. Replacement needs recorded creation authority,
verified native cutover, one coordinator and observer rebinding. No live retirement
or automatic session-launch implementation is authorized by this documentation.

Owner clarification, 2026-10-16: the session lifecycle policy is effective now,
not deferred to G055. Apply [supervisor continuity](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#supervisor-session-continuity)
to Codex as well, with actual native/model limits and compaction first.
Native context limits, security stops and acceptance gates remain. This records
policy, not an automatic launcher or authority to bypass existing session-creation checks.


## Mandatory task-acceptance cleanup — owner direction, 2026-10-17

At every final task ACCEPT, immediately remove all finished task-owned worktrees
and temporary resources under the [mandatory cleanup policy](distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#mandatory-cleanup-at-every-task-acceptance).
Unlock finished agent-tool worktrees; preserve unique commits and dirty/untracked
work before force removal when needed. Clean task-owned containers, unused images
(including attributable untagged IDs), volumes and disposable evidence without
another routine approval. Keep active coordinator/successor/preview resources and
specific recorded retention/native-permission holds. Verify a compact receipt;
no silent lock/dirty-work deferral, blanket host prune or extra test campaign.


## Private files for new tasks — 2026-10-17

Use the main integration checkout's ignored `.local/{evidence,secrets,tmp}/` for
new task-private files, shared across worker worktrees. Follow the [local storage
policy](distributions/skills/restruct-sdlc/sdlc-management/references/adoption.md#repository-local-private-files): task-scoped paths, private permissions, Git/build/upload exclusions
and separate evidence/secret cleanup. Existing tasks keep current paths through
acceptance cleanup; do not migrate active bindings or native session state.

## Repository-scoped Claude permission rules — 2026-10-18

Follow [permission placement and verification](docs/agents/CLAUDE-PERMISSIONS.md):
tracked shared `.claude/settings.json`, ignored repository-specific personal
`.claude/settings.local.json`, home settings only for intentional cross-project
preferences. Verify actual native loading in worker/service launch directories;
permission rules never replace task/review/merge authority or native gates.
Distinguish those tool rules from [classifier configuration](distributions/skills/restruct-sdlc/sdlc-management/references/adoption.md#tool-permissions-and-classifier-rules):
use its native required scope, preserve owner-selected defaults, and verify actual
approved operations. Machine/task exceptions and disabled classifier rules are
not portable adoption defaults.

## Status writing — owner direction, 2026-10-18

Preserve useful detail; abbreviate prose with short labeled bullets and fragments.
Trim filler/repetition, not decisions, evidence, uncertainty or gates. Follow
[operational messages](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#concise-operational-messages).
This does not change status-marker syntax or require a dashboard-only report.


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
success are distinct. Follow the [permission handoff](distributions/skills/restruct-sdlc/sdlc-management/references/adoption.md#permission-change-handoff).
This standing authority covers bounded allow rules, not unrequested global guard
removal, permission-mode changes, broader task authority or bypassing a remaining
native refusal. Keep the current task, budget and single write owner.


## Usage correction — owner direction, 2026-10-18

Enforce [existing context and coordination limits](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#enforce-existing-usage-limits)
at safe boundaries: one implementation worker by default, justified bounded nested
delegation, delta-only continuation reads, no unchanged-state acknowledgment or
repeated closed-worker inventories. Keep current indexes concise and preserve
history in archives. Required security/CI proof and native gates are unchanged.

Local validation also follows [change-impact selection before merge](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#change-impact-selection-before-merge): map applicable shared fixtures, migration paths, pins, disabled variants and concurrency effects to focused checks in the existing Plan. Reconcile material deltas; no new broad-run default or duplicate planning artifact.

Owner direction,2026-10-18: Codex supervision with a Claude implementation coordinator remains the supported delivery baseline. Live workflow changes retain their existing release gates.


## Direct owner replies and supervisor status — owner direction, 2026-10-19

For ordinary replies to direct owner prompts, omit the `COORDINATOR_STATUS` footer unless the owner explicitly asks for supervisor relay. Exception: an owner decision that changes active work must immediately emit a `decision` footer, including when the same message answers the owner. Do not wait for another turn, worker handback or review-ready commit. Autonomous handoffs and required reviews, failure records and blockers still notify; existing legacy-prose/native-question detection is unchanged.


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

Use descriptive annotated-tag messages and GitLab Release notes for every new
candidate and release: list included Goal/task IDs and short titles, material
changes since the previous candidate/release, and the exact source commit.
Keep canonical SemVer tag names. Final release notes summarize all delivered
scope since the previous stable release and state qualification truthfully; omit
planned or unaccepted work. Never rewrite a published tag to improve its message;
update the associated Release notes instead. See the
[tag/release description policy](docs/agents/GITLAB-BRANCH-WORKFLOW.md#descriptive-tag-and-release-notes--owner-direction-2026-10-20).


## Optional RC Duo review and planned follow-ups — 2026-10-19

RC promotion MRs may use Duo as optional advisory review; Duo is not a required promotion approval gate. Triage every actionable finding and record its disposition in the existing release Result/review. Changes arising from that review must be scheduled and planned as bounded follow-up work: identify the responsible Goal/task or backlog item, intended milestone, scope, acceptance criteria and required focused validation under normal authority. Do not leave actionable recommendations only in an MR comment or silently claim them fixed. Do not change the code or move/recreate an existing RC tag to satisfy a review. If a follow-up changes the release candidate or must ship before promotion, integrate it through the normal staging workflow, then cut and fully qualify a new RC tag under the applicable release authorization. The owner may instead retain a documented prerelease follow-up while the current RC proceeds, as explicitly selected for RC1's limiter finding. Preserve ordinary feature-MR Duo requirements, required full RC CI, security/owner dispositions, selected auto/manual merge mode and the separate final-tag decision. This policy creates no automatic RC2, unbounded repair mandate or new tracker.


## Owner-waived canceled RC qualification — 2026-10-19

The owner may explicitly cancel a named RC promotion MR's pipeline and direct that cancellation to satisfy the current RC promotion gate, treating it as success for the workflow decision only. Record an owner-authorized qualification waiver with the RC/tag/source SHA, MR/pipeline IDs, actual canceled/partial/failed job states, any known findings and the owner's instruction. Never report the canceled pipeline or unexecuted tests as PASS, fabricate green evidence, erase prior failures or claim complete qualification. This exception is specific to the named candidate/pipeline; it is not the default cancellation policy for feature MRs or other RCs. Schedule a follow-up immutable RC tag to run the complete required suite, carrying forward unresolved findings and the qualification debt in the existing release Result/backlog/Plan. The follow-up RC's full suite is required unless the owner separately changes that instruction. The owner may choose this waiver alongside the recorded auto/manual promotion mode; provider auto-merge still depends on the provider's actual cancellation behavior. If existing provider rules prevent merging a canceled pipeline, report that concrete limitation rather than changing project-wide protections, manufacturing a success job or treating a queue as a completed merge. No pipeline is canceled merely because this option is documented, and final release tagging remains separately owner-controlled.


## Reviewed-merge classifier reconciliation — 2026-10-19

Owner requested narrowing the Merge Without Review false positive. Follow
[reviewed merges and classifier false positives](distributions/skills/restruct-sdlc/sdlc-management/references/adoption.md#reviewed-merges-and-classifier-false-positives):
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
[reusable test permissions](distributions/skills/restruct-sdlc/sdlc-management/references/adoption.md#reusable-private-test-permission-rules).
Preserve existing provider/GitLab rules. Keep active task-specific bridge rules
through safe transition/cleanup; permanent rules remain. Verify loaded sources and
actual operation. Prefix/allow matching does not prove task ownership or waive
native checks, test concurrency/build sequencing, scope or validation.


## Context and coordination economy — owner direction, 2026-10-20

Current checkpoint/handoff defaults are150k/200k active context, superseding300k/350k at the next safe boundary for workers and coordinator/supervisor assessment. Claude and Codex may initiate supported native self-compaction at safe points under standing owner authority. Preserve running batches, exact bindings, children, budgets and smaller native limits. Follow [coordination economy](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#lower-context-thresholds-and-coordination-economy--owner-direction-2026-10-20): directly review routine prose/mechanical changes, reserve delegation for meaningful independent work and required substantive reviews, batch searches/edits/handoffs, read deltas and suppress unchanged polls/acknowledgments. No full-history fork, duplicate worker, test rerun or automatic replacement merely for a checkpoint; no weakening of security/source/validation gates.


Owner amendment,2026-10-20: [safe self-compaction](distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#authorized-self-compaction-at-safe-points--owner-amendment-2026-10-20) is authorized for Claude/Codex without a repeat prompt where native tools support it; supersedes the prior owner-only policy. Save current state first, preserve active work, and verify native completion. An operator-only interface remains a capability limitation, not a reason to fabricate input or replace sessions.


## Codex routine work and Sol delegates — owner direction, 2026-10-20

Routine supervisor status, management publication, handoffs, prose/configuration and mechanical changes run directly in the main session; no subagent by default. The checked-in repository default for the main Codex session is `gpt-6.1-sol` at `low` reasoning effort. Delegated Goal/Spec drafting and substantive independent reviews default to `gpt-6.1-sol` at high reasoning effort through repository `.codex/config.toml`. Both are repository defaults and are user-overridable; explicit session settings and approved per-spawn overrides take precedence. Verify actual loading/dispatch rather than assume a running session refreshed. Use bounded fresh inputs where model inheritance prevents an override; no extra persistent session, test spawn, full-history fork or separate quota is implied. Preserve formal review/authority gates and Claude's separately selected implementation model.


## Restruct SDLC skill-suite packaging

The bundled Restruct SDLC package is a thin parent
([entrypoint](distributions/skills/restruct-sdlc/SKILL.md)) with the three preserved
child skills `sdlc-management`, `repository-supervisor` and `claude-session-handoff`.
It is SaaStacular's default bundled development dependency and lives at
`distributions/skills/restruct-sdlc/` with the aliases `.agents/skills/restruct-sdlc`
and `.claude/skills/restruct-sdlc` (relative symlinks to
`../../distributions/skills/restruct-sdlc`). Commands and links in these guides use
the bundle through those paths. Its MIT notices
([LICENSE](distributions/skills/restruct-sdlc/LICENSE) and the `sdlc-management`
LICENSE) stay with it. To use another workflow, remove both aliases and replace
or update AGENTS.md, CLAUDE.md and `docs/agents/` with your own rules; see
[docs/DEVELOPMENT-WORKFLOW.md](docs/DEVELOPMENT-WORKFLOW.md).

## Retain local validation diagnostics — owner direction, 2026-10-21

Before retrying or handing back a failed local run, verify its scanned check ledger,
underlying failure logs and supplemental JUnit survive temporary cleanup in the
explicit private evidence directory. Record any withheld/missing detail and keep
failed and later outcomes distinct; wrapper/phase failures are not underlying test
IDs. Follow [local evidence retention](distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#preserve-local-validation-evidence-before-cleanup).
