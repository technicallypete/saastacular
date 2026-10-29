---
name: sdlc-management
description: Establish or adapt a repository's agent-driven software delivery management structure (owner, supervisor, coordinator and worker roles; Goal, Spec, Plan and Result artifacts; budgets; validation, integration and closeout policy) from parameterized templates. Use when adopting this workflow in a repository, instantiating management files, or applying its lifecycle rules; not for session messaging (use claude-session-handoff) or resuming supervision (use repository-supervisor).
license: MIT
metadata:
  version: "0.1.0"
---

# SDLC management

This skill packages a management workflow for repositories where a human owner directs, a supervisor agent owns
Goals, Specs, reviews and progress, and an implementer agent owns Plans, code, validation and Results. It provides
policy references, parameterized templates and a stdlib checker. It contains no transport code, assigns no role and
grants no authority by being loaded.

## Boundaries

| Skill | Responsibility |
|---|---|
| `sdlc-management` (this) | Management structure, policy templates and lifecycle rules. |
| [repository-supervisor](../repository-supervisor/SKILL.md) | Resuming supervision from records; reviews, dependencies and progress. |
| [claude-session-handoff](../claude-session-handoff/SKILL.md) | Verified session discovery and routing, receipts, coordinator pin and closeout archiving. |

Load this skill alone for adoption or lifecycle questions. Load a companion only when its job is needed (messaging or
supervision); routine status checks never require reloading all three.

## When to use

- Adopting the workflow in a new or existing repository.
- Instantiating a Goal, Spec, Plan, Result, review or progress archive from the templates.
- Applying budget, validation, integration, coordination or closeout rules during a Goal.

Not for: product or architecture decisions, permission or provider changes, creating sessions, or CI redesign.

## Native worker naming

Prefer task/role prefixes when supported. Treat unsupported Claude naming as a
harness bug/limitation and continue with its native isolated `agent-<id>` path,
recording task/role/agent/path/branch/HEAD. Naming alone never requires a stop,
exception or operator prompt; ownership/isolation and native permissions remain.
See [coordination](references/coordination.md#naming-dispatch-and-relocation-gate).

## Inputs

A parameter file copied from [parameters.example.json](templates/parameters.example.json): project name, release,
actor names, canonical sources, service roots, branches, CI provider and commands, validation options, budget and
effort defaults, and an optional delivery barrier. Every value is explicit; there are no hidden project defaults.

## Authority precedence

Harness and platform instructions, then explicit owner direction, then the adopting repository's canonical
governance and decisions, then this skill and its templates. A policy change records its true scope and authority.

## Adoption procedure (summary)

1. Read existing governance; record parameters and adoption authority in a Plan.
2. Install the bundle directory `restruct-sdlc/` whole into `.agents/skills/` **without overwriting**: an absent
   bundle is copied, an identical one left alone, a differing one kept for a recorded owner decision, symlinks refused.
3. Render each template to its destination; **create only missing files**. An existing file gets
   `<path>.sdlc-proposed` and a recorded owner merge decision.
4. Append the ignore lines; create the release directories.
5. Run the checker in adoption mode and fix every finding.

Full procedure, destination rule, risk tiers, installation limits and foundation-skill composition:
[adoption](references/adoption.md).

## Lifecycle summary

Goal-to-closeout flow, role boundaries, automatic planning/execution transitions,
validation, evidence retention and cleanup defaults are summarized in the
[workflow overview](references/workflow-overview.md). The linked policy references
below define the enforceable procedures.

Record Goal/Spec priority from owner direction or supervisor judgment, normally high.
Plans target 3–5 implementation workers for high or 1–3 for low when feasible;
kickoff follows the accepted allocation. The planner may remain one agent. See [parallel scheduling](references/coordination.md#default-to-safe-parallel-work-lanes).

## References

| Reference | Covers |
|---|---|
| [roles-lifecycle.md](references/roles-lifecycle.md) | Ownership matrix, eleven Goal fields, Spec/Plan/Result contents, IDs and history, four states, precedence, delegated review and relay. |
| [budgets.md](references/budgets.md) | B3/B5/B7/B15, checkpoints, Goal-level attempts, shared allowance, extensions, ledger, effort. |
| [adoption.md](references/adoption.md) | Parameters, risk tiers, procedure, installation and helper path limits, composition with foundation skills. |
| [coordination.md](references/coordination.md) | One coordinator, one owner per task, recovery, compact repair handoffs, receipts versus acknowledgment, the naming, dispatch and relocation gate. |
| [validation-integration.md](references/validation-integration.md) | Pre-batch reconciliation, finish-fix-rerun-advance, manual feedback order, merge requests, options, GitLab/glab adapter. |
| [closeout-risks.md](references/closeout-risks.md) | Verified cleanup, archive-before-prune, PROGRESS compaction, risk and performance reporting. |
| [workflow-overview.md](references/workflow-overview.md) | Goal lifecycle, routine transitions, validation and closeout defaults. |

## Templates

| Template | Destination |
|---|---|
| [root/AGENTS.md.tmpl](templates/root/AGENTS.md.tmpl) | `AGENTS.md` |
| [root/CLAUDE.md.tmpl](templates/root/CLAUDE.md.tmpl) | `CLAUDE.md` |
| [root/gitignore.lines](templates/root/gitignore.lines) | appended to `.gitignore` |
| [management/BUDGET_POLICY.md.tmpl](templates/management/BUDGET_POLICY.md.tmpl) | `management/BUDGET_POLICY.md` |
| [management/SUPERVISOR-HANDOFF.md.tmpl](templates/management/SUPERVISOR-HANDOFF.md.tmpl) | `management/SUPERVISOR-HANDOFF.md` |
| [release/ROADMAP.md.tmpl](templates/management/release/ROADMAP.md.tmpl) | `management/<release>/ROADMAP.md` |
| [release/PROGRESS.md.tmpl](templates/management/release/PROGRESS.md.tmpl) | `management/<release>/PROGRESS.md` |
| [release/BACKLOG.md.tmpl](templates/management/release/BACKLOG.md.tmpl) | `management/<release>/BACKLOG.md` |
| [goals/GOAL.md.tmpl](templates/management/release/goals/GOAL.md.tmpl) | `goals/<ID>-<stem>.md` |
| [specs/SPEC.md.tmpl](templates/management/release/specs/SPEC.md.tmpl) | `specs/<ID>-<stem>-spec.md` |
| [plans/PLAN.md.tmpl](templates/management/release/plans/PLAN.md.tmpl) | `plans/<ID>-<stem>-plan.md` |
| [results/RESULT.md.tmpl](templates/management/release/results/RESULT.md.tmpl) | `results/<ID>-<stem>-result.md` |
| [results/REVIEW.md.tmpl](templates/management/release/results/REVIEW.md.tmpl) | `results/<ID>-implementation-codex-review.md` and other reviews |
| [archive/PROGRESS-ARCHIVE.md.tmpl](templates/management/release/archive/PROGRESS-ARCHIVE.md.tmpl) | `archive/PROGRESS-YYYY-MM-DD.md` |
| [docs/RISKS.md.tmpl](templates/docs/RISKS.md.tmpl) | `docs/RISKS.md` |
| [docs/agents/BRANCH-WORKFLOW.md.tmpl](templates/docs/agents/BRANCH-WORKFLOW.md.tmpl) | `docs/agents/BRANCH-WORKFLOW.md` |
| [docs/agents/CLAUDE-SESSION-HANDOFF.md.tmpl](templates/docs/agents/CLAUDE-SESSION-HANDOFF.md.tmpl) | `docs/agents/CLAUDE-SESSION-HANDOFF.md` |

Templates use the `.tmpl` suffix so agents never read them as directory-scoped instructions in skill storage.

## Supported limits

Tested on Linux/WSL with Python 3 (standard library) and Git. Messaging and closeout depend on the companion
helper's verified Claude Code versions and its committed path under `.agents/skills/restruct-sdlc/`. The Claude
parent alias is supported as documented. Not supported: other install paths, separately installed children, discovery
after `cd`, personal installation, macOS or native Windows, and unverified CLI versions. See [adoption](references/adoption.md#installation-storage-versus-discovery).

## Package verification

[scripts/check_sdlc.py](scripts/check_sdlc.py) prints a JSON summary; exit 0 passes, 1 reports findings, 2 is a
usage error. Run it with `python3 -B` so no bytecode lands in the skill directory.

```sh
python3 -B scripts/check_sdlc.py package PATH [--skills-root DIR] [--literals FILE]
python3 -B scripts/check_sdlc.py adoption REPO --release NAME [--literals FILE]
python3 -B -m unittest discover -s scripts -p 'test_check_sdlc.py'
```

For a Git-backed bundle, verify a copy without its VCS metadata rather than weakening the check, for example
`rsync -a --exclude=/.git BUNDLE/ /tmp/bundle-check/ && python3 -B scripts/check_sdlc.py package /tmp/bundle-check`.
`package` checks a skill directory or a bundle root: frontmatter, the 200-line limit, intra-package and sibling-skill
links, template placeholders and adoption links, reachability, file types, forbidden content and the bundle
manifest. `adoption` checks an instantiated repository: required paths, rendered links, leftover placeholders, Goal
fields, ignore lines and forbidden content. Tests: [scripts/test_check_sdlc.py](scripts/test_check_sdlc.py).

## Model-aware effort override — 2026-10-15

Default to Fable 5.1 high for planning and test corrections, including failure
diagnosis and corrective test/fix work; use xhigh selectively for especially
hard work. Default to Opus 5.5 high for implementation and routine test execution.
Default the Claude coordinator to Sonnet 5.5 high; the owner may choose Opus 5.5
high for coordinator work. If the owner selects Opus 5.5 for planning or test
corrections/diagnosis, use xhigh; routine implementation, test execution and
coordination stay high. This overrides generic role-effort defaults for those
models/phases. Honor owner choices, verify actual settings or report unknown,
and apply only at the next phase/safe boundary without restarting completed work,
interrupting a batch or resetting budgets. Prose does not configure the runtime.

## Fable weekly-limit fallback

On confirmed Fable weekly exhaustion, use Opus xhigh for planning/test corrections
and high for routine implementation/tests. Coordinator default is Sonnet 5.5
high, with Opus 5.5 high available by owner choice. Follow the
[full fallback policy](references/budgets.md#fable-weekly-limit-fallback--owner-direction-2026-10-15)
for safe transitions, verified reset, preserved state/budgets and native permissions.


Operational defaults: use [proportionate validation and bounded flake handling](references/validation-integration.md#validation-economy-and-recurring-failures),
[automatic routine transitions](references/roles-lifecycle.md#continue-routine-transitions-without-owner-nudges),
[compact status/handoffs](references/coordination.md#concise-operational-messages) and the
[closeout receipt](references/closeout-risks.md#routine-closeout-receipt). Reuse existing tools and records.

Default to [targeted reruns and attributable coverage](references/validation-integration.md#targeted-reruns-and-attributable-coverage-default):
reuse unchanged passes and overlap frozen-source review with independent validation;
retain real failure history and required final/hosted gates.


## Mandatory task-acceptance cleanup — owner direction, 2026-10-17

At every final task ACCEPT, immediately remove all finished task-owned worktrees
and temporary resources under the [mandatory cleanup policy](references/closeout-risks.md#mandatory-cleanup-at-every-task-acceptance).
Unlock finished agent-tool worktrees; preserve unique commits and dirty/untracked
work before force removal when needed. Clean task-owned containers, unused images
(including attributable untagged IDs), volumes and disposable evidence without
another routine approval. Keep active coordinator/successor/preview resources and
specific recorded retention/native-permission holds. Verify a compact receipt;
no silent lock/dirty-work deferral, blanket host prune or extra test campaign.


## Private files for new tasks — 2026-10-17

Use the main integration checkout's ignored `.local/{evidence,secrets,tmp}/` for
new task-private files, shared across worker worktrees. Follow the [local storage
policy](references/adoption.md#repository-local-private-files): task-scoped paths, private permissions, Git/build/upload exclusions
and separate evidence/secret cleanup. Existing tasks keep current paths through
acceptance cleanup; do not migrate active bindings or native session state.

Cleanup is continuously accounted for at allocation, handback/batch end and
interruption/restart, under [mandatory resource accountability](references/closeout-risks.md#continuous-resource-accountability--mandatory).
Verify every category; abandoned resources block new resource-consuming work.
Reclaim finished worker trees at handback, not only at final acceptance.
