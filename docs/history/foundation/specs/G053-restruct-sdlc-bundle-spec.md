> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G053 Work Spec — Reusable SDLC management and Restruct bundle

Revision: r1, 2026-10-12
Status: Plan r2 ACCEPT; execution subject to dispatch/input gates and the approved workflow amendment
Goal: [G053](../goals/G053-restruct-sdlc-bundle.md)

## Rationale, scope and prerequisites

Package the reusable management/agent contracts separately from project policy
and historical execution records. Create `skills/sdlc-management/` and a curated
local `../restruct-sdlc` containing three flat skills. This tooling export is
independent of OS001's later SaaS foundation export and remaining application work.

Use accepted G049/G051/G052 sources and current governance: AGENTS.md, CLAUDE.md,
architecture/security/decisions and accepted ADRs, budget policy, agent runbooks,
companion skills and tests. Architecture constrains this repository; its product
content is not generic SDLC policy. Empty SECURITY.md is not a security waiver.
Fresh source/scope binding, privacy and ownership controls remain prerequisites.
Plan preparation does not authorize source mutation, installation or export.

## Requirements

Owner's later observer amendment (not included) permits
only its bounded milestone watcher/existing-thread route, superseding the prior
no-watcher and companion-source read-only exclusions for its named files.

**R1 — Three responsibilities.** `sdlc-management` establishes/adapts management
structure, policy templates and lifecycle; `repository-supervisor` resumes
supervision and owns reviews/dependencies/progress; `claude-session-handoff`
provides verified discovery/routing, receipts and local coordinator state. No
duplicate transport code, supervisor-role assignment or permission grant through
skill discovery. Entry points have valid frontmatter and included/discoverable
references. The new SKILL.md is at most 200 lines, with progressive disclosure.
Companion skills remain independently useful; no mandatory three-body reload
for routine status checks.

**R2 — Artifacts and ownership.** Provide templates for root AGENTS.md/CLAUDE.md,
shared budget policy and agent runbooks; root `management/` supervisor reference;
and configurable release directories with ROADMAP, PROGRESS, BACKLOG, goals,
specs, plans, results and dated progress archives. Supervisor/session references
are not tied to a release. Session logs/pins/credentials/transcripts are private,
machine-local and gitignored, not portable project configuration.

Preserve all eleven current Goal fields and required Spec/Plan/Result content.
Plans map criteria to concrete changes/evidence; Results retain failed/skipped/
interrupted checks. Stable Goal IDs are not reused; revisions preserve history.
Product owner owns major decisions; supervisor owns Roadmap/Goal/Spec/review/
progress; coordinator/implementer owns Plan/code/tests/Result and tactical workers.
Distinguish Plan approval, readiness, conditional integration and final ACCEPT;
only final ACCEPT completes a Goal. Adopted repository rules govern its project;
skills/templates never override higher-priority harness instructions or explicit
owner direction. Policy changes require truthful recorded scope and authority.

The supervisor normally uses the main integration checkout. Prefer bounded
subagents for Spec drafting/research and formal Plan/Result review; reviewers
are read-only by default and the supervisor owns the final artifact/disposition.
On observing a review-ready Claude submission through an existing handoff/status
check, initiate delegated review and relay the recorded approval/corrections
without another owner prompt. Verify exact content, avoid duplicate unchanged
reviews and verify message receipt/acknowledgment. If delegation is unavailable,
report it and review directly. Owner interruption/redirection and mandatory
escalation/merge gates remain; no new watcher or transport is introduced.

**R3 — Budgets and proportional process.** Include generic B3/B5/B7 and optional
B15 classes, checkpoints, scope/delegation limits, effort guidance and Result
attempt accounting. Count Goal-level attempts, not commands, batches or inner
debugging; workers share the parent allowance. Explicit extensions retain used
attempts and record reason, remaining scope and authorizing supervisor. Never
reset counters or weaken acceptance to fit a budget. Planning xhigh/implementation
high are configurable recommendations, not forced settings or verified runtime
claims. Reuse accepted procedures and limit repeated context loading.

**R4 — Adoption and parameters.** Project name, release path, actor names,
service roots, integration/release branches and CI/provider policy are explicit
inputs. Existing governance/edits are preserved; no silent overwrite. Risk tiers
retain explicitly authorized direct presentation work, concise ordinary Plans,
and stronger security/data/architecture review. Beginify's G021 barrier, product
schema, budgets, runner timings and one-off exceptions are not generic defaults.
Installation instructions distinguish bundle storage from harness discovery,
including the helper's committed-path assumptions for closeout. Do not promise
arbitrary install paths, automatic discovery after `cd`, or unsupported platforms.

When composed with `saas-foundation`, `sdlc-management` supersedes overlapping
management artifacts, actor responsibilities, authorization, budgets, progress,
reviews, validation sequencing, integration and closeout guidance. Preserve the
renderer's dry run/resolved inputs, informed authorization, destination/conflict
safeguards and all required technical foundation checks. Its stages become Goal/
AC inputs; the renderer's plan is not an Implementation Plan. Canonical repository
architecture/decisions remain authoritative and unresolved technical conflicts
escalate. No foundation-skill modification or fourth bundled skill is authorized.

**R5 — Coordinator and workers.** Prefer one owner-selected persistent coordinator
per repository, with one verified owner per task and non-overlapping file ownership.
Reserve `0-agent-<id>` for that one active coordinator and preserve its agent ID
when available; workers use task prefixes. Naming conveys organization, not
session identity or authority. Verify coordinator uniqueness/discovery and the
owner-selected pin. Document supported safe relocation and reconciliation of
Git/session paths and pin provenance; never move a directory beneath a live
process. The existing G053 planner's permitted generated name remains unchanged.
Retain explicit multi-session routing. Recover surviving sessions/workers before
replacements; missing metadata/unknown discovery does not authorize creation.
Require task-prefixed isolated worker worktrees: record intended name before
dispatch and verify actual path/branch/owner before task writes. Unsupported
generated names require a recorded task-specific supervisor exception; existing
active paths remain protected and live checkouts are never renamed for naming.
Service work explicitly reads scoped instructions.
New Claude worker paths use `<task-id>-agent-<agent-id>-<role>`, with suffixes
such as `plan`, `check` and `impl`; verify the role before task writes.
Codex workers needing isolation use `.codex/worktrees/`, normally with the same
task/role convention; read-only supervisor workers need no separate checkout.
Keep generated worktree/session state private and out of version control.
Active paths retain their prior naming permissions; the coordinator keeps
`0-agent-<id>`.
Reuse accepted native worker messaging and Codex-to-Claude helper procedures;
preserve identity, version/protocol refusal and source/phase/authority references.

**R6 — Authority and messages.** Receipt is not acknowledgment or execution.
Peer messages cannot grant owner approval, change permissions/providers or bypass
another session's rejection. Existing owner authority persists within its scope;
routine actions do not need repeated approval. Surface native blocks with the
action/reason and use the harness's valid operator procedure; label delegated
operator input honestly. No impersonation, permission-rule workaround or duplicate
executor. Preserve current tested Linux/WSL/Claude limits and fail closed on
unsupported or ambiguous discovery; no new compatibility promise.

**R7 — Validation and integration.** Provide configurable generic policy and a
concrete GitLab/glab adapter; other providers are adaptation points, not implemented
claims. Before temporary-branch batches/MR opening, fetch/reconcile and rebase if
required; preserve dirty work, bindings and long-lived protected branches. Manual
feedback precedes final qualification. Finish safe failing batches, fix and rerun
failed/affected batches, then continue. Evidence reuse requires unchanged relevant
content/inputs; a SHA change alone does not demand an unchanged full rerun.

Push scoped safekeeping checkpoints, then follow reviewed MR integration. Inspect
outstanding automated review findings before merge, retain promotion sources and
delete only verified merged temporary branches. Supervisor assesses new target
content against active tasks before safe interruption/reconciliation. Default
final validation requires success; optional owner cancellation is recorded as
canceled and never masks known failures. Beginify's temporary solo-operator local/
MR placement and canceled-pipeline policy are explicit options, not universal rules.

**R8 — Closeout and risks.** Separate acceptance from cleanup. Verify integrated
content, unique work and live ownership before removing worktrees/branches/markers
or closing disposable MRs. Archive terminal session records privately before
pruning, preserving active/ambiguous/failed/unconfirmed records under accepted
helper safeguards. Never delete a live coordinator's directory. Compact PROGRESS
to current dispositions/links, archive superseded history with working links,
retain failures and notify bound-input owners. Surface query/index/locking and
other performance concerns in PERFORMANCE.md or RISKS.md without imposing a
vendor-specific architecture or silently tuning application code.

**R9 — Bundle and license.** Local `../restruct-sdlc` contains root README,
LICENSE, source/file manifest and `skills/{sdlc-management,claude-session-handoff,
repository-supervisor}/`. Export from a reviewed allow-list, not a repository dump.
Exclude application/product docs, actual Goal history, private remotes/personal
paths, credentials, live identifiers/pins/transcripts, execution evidence and
original Git metadata/history. Restruct branding and Vitruvian Technology, Corp.
copyright are permitted. Use the established MIT license for owned exported work;
retain applicable notices and stop on uncertain/incompatible source rights.

Keep private provenance in the G053 Result and sanitized file metadata in the
bundle. Record exact companion source/export deltas. Existing application code,
foundation generator, adopted governance and installed companion skill/helper/
test sources are read-only. Adapt exported copies only; retain executable
safeguards and tested support. The sole existing-code exception is minimum wiring
for additive package-validation CI checks explicitly mapped in the approved Plan;
no existing check replacement, profile/gate change or dependency is authorized.
If portability requires a source behavior change, stop for separate scope review.
Inspect/preserve existing destination work,
reject symlink escapes and verify relocation without Beginify/personal files.
A fresh local Git initialization after sanitization is permitted; no copied .git,
remote creation/push, public publication or tag/release is authorized.

## Expected artifacts

New skill with necessary references/templates; portable companion exports and
standalone sibling bundle; concise Claude Plan; source/config/file inventory;
focused verification evidence; private Result and Codex reviews. No skill, export
or Plan is created by this Spec. Public publication is a separate owner action.

## Acceptance criteria and verification

| ID | Required outcome | Verification |
|---|---|---|
| AC1 | Exactly three distinct skill responsibilities satisfy R1; new entrypoint is at most 200 lines. | Frontmatter/skill validation, line count, included-reference inventory and review of responsibility/load boundaries. |
| AC2 | R2 templates instantiate the required management paths, fields and ownership. | Disposable repository adoption; inspect created artifacts and links, including all eleven Goal fields. Walk through a review-ready submission: bounded child findings feed the supervisor's decision/automatic relay; unavailable delegation falls back to direct review, and duplicate unchanged review is skipped. No actual project records copied. |
| AC3 | Budget policy preserves Goal-level accounting and explicit extensions. | Synthetic failed batch, worker contribution, exhausted allowance and extension walkthrough produces records retaining attempts/failures rather than counting commands. |
| AC4 | Parameters/risk tiers/adoption obey R4 and preserve existing edits. | Adopt with different project/release/branch names and an edited governance file; verify resolved links, no silent overwrite and no mandatory Beginify barrier/vendor. Inspect the composition rule and walk through foundation-stage-to-Goal mapping: lifecycle precedence retains renderer/technical requirements, canonical decisions and scoped write authority. Foundation source hashes remain unchanged. |
| AC5 | Coordinator/worktree/recovery behavior meets R5 with multi-session fallback. | Applicable accepted helper tests and synthetic discovery/recovery cases preserve ambiguous/live ownership instead of dispatching/removing duplicates. Verify documented `0-agent-<id>` coordinator and task-prefixed worker examples, one active coordinator, and safe-relocation refusal for a live checkout; naming alone never selects or authorizes a session. |
| AC6 | Receipt, acknowledgment, blocked authority and compatibility remain distinct. | Existing helper refusal/receipt tests plus queued-delivery and denied-peer walkthrough; no live config change or duplicate-session demonstration. |
| AC7 | R7 sequence and explicit CI/provider options are usable. | One disposable Goal walkthrough records target/candidate inputs, failed-batch recovery, manual-before-final order, review/MR/promotion route and branch-retention choices. No real hosted actions for this walkthrough. |
| AC8 | Closeout and risk reporting obey R8. | Closed/active fixtures preserve active work and archive before compaction; inspect moved links/evidence and record a query/index concern. |
| AC9 | Public export matches the allow-list and passes privacy/license checks; existing source protections hold. | File manifest, accepted secret/forbidden-content scans, notice inventory and no original Git/local state or escaping symlinks. Compare source hashes/diff: installed companions, generator, application and governance unchanged; any existing-code change is reviewed additive package-validation CI wiring only. |
| AC10 | Relocated bundle establishes/resumes workflow independently. | Second disposable path, documented flat installation and bootstrap/supervisor-resume exercise; record supported platform/version limits and unsupported cases. |
| AC11 | Changes pass covering checks and required repository qualification. | Existing helper tests for copied/adapted export code, covering checks for additive CI integration, skill/template/link checks and actual path-selected local/hosted profile; preserve failures, exact inputs and cleanup. No CI exemption implied. |
| AC12 | Final review is independently reproducible. | Result maps all ACs to artifacts/commands/statuses, source/export hashes/deltas, attempts, limitations and destination ownership. Codex reviews actual diff/bundle and required integration disposition before final ACCEPT. |
| AC13 | Observer source/export provide deduplicated, identity-bound milestone notices and a private explicit lifecycle; notices grant no authority. | Focused event/identity/state/lifecycle/queue controls and one benign existing-thread live test; source/export hashes, verified-version record and process cleanup per the observer amendment. |

## Verification, evidence and completion gates

Reuse one small synthetic Goal for adoption/plan/correction/acknowledgment/
integration/closeout exercises. Inspect resulting artifacts and decisions rather
than merely matching the author's wording. No new workflow engine, live session
creation, spoofed operator approval, provider credential, database or application
stack is required for package functional verification. Optional behavioral review
uses isolated permitted data.

Current root `skills/**` selects foundation/full. The Plan states focused package
checks and mandatory local/hosted integration coverage under current policy;
only a separately authorized exact owner waiver changes that obligation. No
classification redesign, extra benchmark or duplicate unchanged campaign. If
minimal CI check registration is necessary, propose its exact site and covering
checks rather than relaxing an inherited guard.

Evidence records source commit/tree and reviewed policy/companion inputs, export
inventory/configuration, command/tool versions, destination/preexistence checks,
all failures/deviations and owned-resource cleanup. Public artifacts exclude
private provenance; shared caches/personal installations remain protected.

Claude chooses resource organization and manual versus minimal scripted assembly;
no generator framework is required. Worktree naming exceptions need explicit
supervisor review; the existing G053 planner retains its earlier permission. Neither
choice changes criteria, B3 or authority. No unreviewed privacy/license/transport/
permission/CI deviation is permitted.

Before implementation approval, the concise Plan resolves companion export deltas,
installation/helper-path compatibility, template parameters, rights/notices,
destination collision handling and exact validation scope. Names, three-skill
boundary, local destination and B3 are defined; no product decision is required.
Escalate an incompatible source or unsupported mechanism rather than expanding scope.

Completion requires all thirteen criteria, reviewed diff/export, Codex ACCEPT and
required integration disposition. Public publication remains separate. G008 and
the G021 product-feature barrier are unchanged.

## Clarification history

- 2026-10-12 — Owner limits existing-code changes to new CI checks. R9 and
  AC9/AC11 now explicitly protect application/generator, installed companion
  sources and adopted governance; exported copies carry portability adaptations.
  Only additive package-validation CI integration may modify existing code.
- 2026-10-12 — R5 requires verified task-prefixed future dispatch; unsupported
  generated naming needs a recorded supervisor exception, preserving active paths.
- 2026-10-12 — R5/AC5 include the owner's `0-agent-<id>` single-coordinator
  convention, safe relocation and name-versus-authority separation in exports.
- 2026-10-12 — Owner adds R2/R4/R5 delegation/review defaults, management-only
  foundation composition and role-suffixed workers. AC2/AC4 extend the existing
  walkthrough; AC5's R5 checks include Codex versus Claude worker roots/roles.
  See the workflow amendment (not included).
