> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G053 — Reusable SDLC management skill and Restruct bundle

Status: Plan r2 ACCEPT; execution subject to dispatch/input gates and the approved workflow amendment
Epic: EPIC-001 — Repository and agent-workflow follow-up
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation

## Objective

Extract the accepted management workflow into `skills/sdlc-management/` and
prepare a standalone local `../restruct-sdlc` bundle with that skill,
`claude-session-handoff` and `repository-supervisor`. Another repository must be
able to adopt it without Beginify's product, private history or machine state.

## In Scope

- Reusable management/governance templates, budget policy and agent procedures.
- Supervisor/coordinator responsibilities, worker worktrees and verified handoffs.
- Main-checkout supervision, bounded delegated Spec/review work, automatic
  review-ready submission handling/decision relay and isolated Codex workers
  under `.codex/worktrees/`.
- Management-process precedence when composed with `saas-foundation`, without
  changing that skill or waiving its technical obligations.
- Coordinator naming `0-agent-<id>` with one active coordinator, task-prefixed
  workers and safe relocation/provenance requirements; names do not grant authority.
- Portability review and packaging of the two accepted companion skills.
- Curated local sibling bundle, README, MIT notices and source/file manifest.
- Additive package-validation CI checks and minimum wiring when required by the
  approved Plan; existing checks and profile/gate behavior remain intact.
- Focused adoption/helper checks, applicable integration/CI evidence and Result.
- Bounded observer capability (not included):
  Claude implements/tests/exports; Codex reviews. This is the specific
  exception to companion-source read-only/no-new-system scope below.

## Out of Scope

Edits to existing application code, foundation generator, installed companion
skills/helpers/tests or adopted governance are out of scope. Adapt companion
copies in the exported bundle only. Application/schema changes, product features,
G008 implementation, G021 release,
OS001 foundation export, new orchestration/transport framework, permission/model/
provider changes, CI redesign, personal installations, cloud deployment, public
remote creation/push, and new tag/release publication.

## Dependencies

- G049, G051 and G052 ACCEPT/integration: skill layout, current CI policy and
  accepted coordinator/helper behavior. These prerequisites are satisfied.
- Current governance, architecture/security/decisions, budget policy, agent
  runbooks, companion sources/tests and the [Spec](../specs/G053-restruct-sdlc-bundle-spec.md).
- Drafting observes staging `84c322dc89f46a671cbf461338a1a08aae7b891e`;
  this is not an execution binding. Approval requires fresh source/scope inputs.

G053 is independent of G008 and does not require G021 to export SDLC tooling.
It does not alter their sequencing or release the product-feature barrier.

## Completion Definition

All thirteen Spec criteria have attributable evidence. Codex accepts actual diff,
local bundle and Result; source implementation integrates through the existing
MR-to-staging route with its required disposition. The sibling is a reviewable
publication candidate, not an authorized public publication.

## Execution Policy

Ordinary bounded tooling/documentation packaging: concise Claude Plan and Codex
approval before implementation. Reuse accepted controls. Existing source code is
read-only except for minimum additive package-validation CI integration approved
in the Plan; no replacement, weakening or refactoring of existing checks.
Exported executable safeguards and permission/session authority must be retained;
any necessary behavioral change escalates before coding. No new execution system
is authorized.

Planning defaults to xhigh, implementation to high; remind the operator and
report unverifiable effort as unknown. One existing coordinator owns dispatch/
integration. Future workers require verified task-prefixed worktrees; a generated
name requires a recorded task-specific Codex exception. The already-dispatched
planner retains its previously permitted path until safe handback. Do not migrate
live tasks or create another coordinator for tests.
Use `0-agent-<id>` for the one persistent coordinator; an existing checkout's
rename waits for its next supported safe point and preserves verified routing.

Package checks are focused. Current `skills/**` selects full coverage; the Plan
must account for required local/hosted qualification under current policy. No CI
exemption or full-suite waiver is granted. Reuse unchanged evidence with explicit
provenance rather than duplicating qualification campaigns.

## Iteration Budget

B3: at most three shared Goal-level attempts; checkpoint after attempt 2 before
using attempt 3. Expected one or two attempts. Commands, inner debugging loops
and subagent contributions do not independently consume attempts. Preserve all
failures; stop on exhaustion and obtain Codex authorization for an extension.

## Context / Scope Budget

Three skills, necessary templates/references and one local bundle. New SKILL.md
is at most 200 lines; substantial policy lives in linked resources. No historical
document dump, scheduler, policy engine or generator framework. Keep the Plan
short and resolve genuine portability questions without reopening accepted controls.

## Suggested Subagents / Delegation

One implementation owner; an optional read-only reviewer for portability,
authority boundaries and disposable adoption. No overlapping edits or delegation
of owner/supervisor decisions. Workers share the Goal's allowance.

## Stop / Escalation Conditions

- Uncertain/incompatible source rights or leaked private data.
- Unowned/nonempty sibling destination, unpreserved work or escaping symlinks.
- Need to weaken helper verification, change permissions/providers, expand CLI/
  platform support or launch duplicate live sessions.
- Architecture conflict, application work, substantial CI/harness changes,
  new infrastructure, scope/context expansion or nonconverging strategy.
- Missing required evidence, budget exhaustion or unapproved public publication.

## History

- 2026-10-12 — Owner authorizes Goal/Spec preparation, commit and push to staging.
  No Plan, skill implementation, sibling creation or execution approval issued.
- 2026-10-12 — Owner clarifies existing code is not to be changed apart from new
  CI checks. Source companions/governance remain read-only; portability edits
  belong to exported copies, with additive package-validation wiring only.
- 2026-10-12 — Future worker naming becomes a verified dispatch gate with explicit
  Codex exceptions. Preserve the existing planner's previously permitted path.
- 2026-10-12 — Owner reserves `0-agent-<id>` for the single persistent coordinator;
  include naming/safe relocation in the exported workflow, preserving G053's
  existing worker exception.
- 2026-10-12 — Owner adds role-suffixed worker names, Codex checkout/delegation
  defaults, automatic review/relay and management-only skill precedence; see
  the workflow amendment (not included). B3 and
  the twelve-criterion scope remain; existing accepted Plan bytes are preserved.

## Budget extension — 2026-10-12

Codex authorizes **B3 override: four total shared attempts** (previous maximum
three; three already used, historical counts preserved). Attempt 4 is limited to
AC13 pending-oversized-row same-inode truncation recovery and its focused test,
export/provenance and Result update. Convergence, scope and validation rationale
are recorded in attempt-3 review (not included).
No further feature scope, counter reset or automatic extension is authorized.

## MR correction budget revision — 2026-10-12

Owner requests fixes for the two confirmed Duo findings on MR !21. Codex revises
the budget to **B5, maximum five total attempts**; four used and preserved.
Attempt 5 is limited to JSONL LF framing and explicit-zero threshold rejection,
their focused tests/export/provenance and existing MR update. See
MR review (not included). No counter reset or automatic
extra attempt; no new feature or release authority.
