# Claude project permission settings

Owner direction,2026-10-18. `.claude/settings.json` is the canonical shared rule
set for this repository (SaaStacular ships none; create it under your own owner
authority when you adopt the workflow). The owner explicitly retained `Bash(glab mr:*)` and
`Bash(glab api:*)`; do not silently replace them with a narrower proposal.

| Location | Purpose | Git |
|---|---|---|
| `.claude/settings.json` | Reviewed shared project rules | Tracked |
| `.claude/settings.local.json` | This operator's machine-specific/repository exceptions | Ignored |
| `~/.claude/settings.json` | Cross-project preferences; classifier settings when native scope requires it | Personal; outside repository |

Repository-specific home paths, harness-lock paths and personal evidence/deletion
exceptions belong in the local file rather than becoming defaults for every
project. Existing personal rules are not automatically migrated or deleted.
Ignore local settings and Claude worktrees, retaining shared settings. Keep
credentials out of every committed configuration and documentation file.

Portable adoption distinguishes [tool permissions and classifier rules](../../.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#tool-permissions-and-classifier-rules).
This repository's exact grants and classifier choices below are local history, not
settings to copy into another project.

## Authorization and verification

Shared permission changes need actual owner authorization; the standing direction
below supplies it for narrowly scoped rules needed by already-approved work. Publish under the
owner's specified branch/MR policy; the explicit2026-10-18 instruction allowed
this installation directly on staging. That exception does not grant future
settings self-modification, automatic native prompt answers or mode changes.
Existing Goal scopes, budgets, GitLab review/merge gates and owner stops remain;
a tool allow rule is not task authority. Preserve actual native rejections and
operator-required controls.

Check `.claude/settings.json` JSON and exclusions after changes. `/core/**` is
settings-source/primary-working-directory relative in project rules; `//` denotes
filesystem-absolute paths. Verify the actual loaded sources/rules through Claude's
supported `/permissions` interface, especially after a merge or for a session
started at `core/`, `web/` or `intelligence/`. Presence on staging is not proof that
an older session/worker loaded the file. Do not assume parent settings inheritance
or make divergent service copies.
No broad application tests, worker restart or repeated full-repository reading is
required merely for a config/documentation update.

Primary reference: [Claude permission documentation](https://code.claude.com/docs/en/permissions).
Portable policy: [adoption](../../.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#repository-scoped-claude-permissions).
Installation/history: record it in the project's own management records.


## Test unblock — owner direction,2026-10-18

The owner explicitly requested removal of the guard blocking approved tests.
Isolated unit/static checks follow AGENTS.md's lock exception: no shared harness
lock for read-only source/cache and disposable private state; shared mutable
cache preparation and integration resources still serialize. Use native plain
commands, retaining worktree isolation.

Codex installed ignored local Edit rules for the reviewed matrix_test.go correction:
project-relative path plus this machine's generated-worker path pattern. Existing
settings and both shared glab rules are preserved. These authorize the reviewed
file edit through Claude's supported permission mechanism; Plan/scope and true
security-test negatives remain. No permission mode, global settings, native CLI
binary or hook changed. Verify next actual edit/test calls before claiming the
block is resolved; file existence alone is not proof of loaded permission.


## Owner removes CI Bypass classifier rule — 2026-10-18

Explicit owner instruction: “Remove the ci bylass security guard.” Codex removed
only the CI Bypass soft-deny rule through supported autoMode configuration in
~/.claude/settings.json. Claude ignores autoMode in repository settings, so this
is an explicit exception to preferring repo-local permission configuration.
Shared/project permissions and both glab rules remain unchanged.

Effective CLI config verified: 72→71 soft-deny rules; all other classifier sections
and prior home settings preserved. Auto-Mode Bypass and Security Test Removal
remain distinct rules. This config change does not authorize skipping a Goal's
required CI or merging without its existing review/authority.

The supported removal replaces the soft_deny list with the current effective
list minus that rule (no $defaults marker). Future built-in soft-deny updates
therefore require reconciliation after a Claude upgrade. No CLI binary, hook,
permission mode, live session or application/CI implementation changed.
Reference: https://code.claude.com/docs/en/auto-mode-config


## Standing permission-request workflow — 2026-10-18

1. Claude sends Codex the exact blocked operation/refusal and literal permission
   rules or patch needed for the approved task, with scope justification.
2. Codex reviews and installs those bounded rules without another routine owner
   prompt: shared project settings for reusable project rules, ignored local
   settings for task/machine exceptions; native-required home scope separately
   identified. Preserve all existing settings and validate JSON/exclusions.
3. Codex notifies Claude of the exact settings file and installed delta.
4. Claude verifies effective loading and actual native operation, then continues
   the same task/budget. Report a remaining native refusal precisely; installation
   alone is not an unblock. No alternate executor or automatic mode change.

This is explicit standing owner authorization for routine approved-scope allow
rules. Unrequested broader access, global classifier-rule removal and task-scope
expansion retain their own authorization. Portable adoption names its available
editor; it does not bundle our paths or require Codex in Claude-only deployments.
Canonical reusable procedure: [permission change handoff](../../.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#permission-change-handoff).


## Reviewed-merge classifier exception — 2026-10-19

Owner authorized a narrow Merge Without Review exception for this repository
and native worktrees. Codex changed only that existing home `autoMode.soft_deny`
entry, retaining the original rule outside scope, all 71 entries and unrelated
settings. Supported native effective configuration confirms the change; live
coordinator loading/actual operation require verification. This is local history,
not a portable settings payload.

Independently recorded exact-SHA supervisor approval under delegated authority
is recognized with current required CI and applicable Duo/findings disposition.
RC promotion retains separate owner authority and truthful named-waiver records.
Read-only listings/authorized closeout are not merge operations. Unreviewed merges,
self-approval and unapproved protection bypasses remain subject to their guards.
See the [portable procedure](../../.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#reviewed-merges-and-classifier-false-positives)
and the project's installation record.


## Private database scheduling and native worktree guard — 2026-10-19

The worktree command-shape check has no disable setting; no Bash allow rule is
installed to override it. See the official [worktree isolation documentation](https://code.claude.com/docs/en/worktrees#how-claude-code-enforces-isolation).
The owner instead approved private per-run database scheduling for all tasks:
AGENTS.md's private database section permits two concurrent runs, a third only
with explicit owner approval. Qualifying runs use supported plain Docker commands
and private resources, not hidden lock wrappers or another executor. Shared
builds/cache preparation/Compose/browser locks and native refusals remain.


## Browser pool and sequential private builds — owner amendment, 2026-10-19

The owner extended resource scheduling: private DB/browser runs share cap2, third
by explicit owner approval; private builds use one exclusive sequential build slot.
See AGENTS.md private browser section. Qualifying run-owned browser/Compose stacks
and reserved private builds use native plain commands without the shared harness
lock. Shared stacks/cache preparation/ports/data retain it. No permission settings,
classifier or native worktree check is removed; no alternate executor/wrapper.
This resolves the resource-rule question, not any later concrete native refusal.


## Reusable project test rules — owner direction, 2026-10-19

Four permanent shared rules now allow build/run/Compose/tagged-image cleanup using
`platform.test_run` and `platform-test-*` (see tracked settings for exact argument
order). Future task/run identity goes in the suffix/value. Two generic evidence
Write/Edit rules live only in ignored local settings for the existing machine's
private evidence root; no secrets-root/native state grant added. Both glab
rules and all other settings preserved. Remove task-specific temporary entries
at their task cleanup, never the permanent rules.
Native loading/actual operation must be verified, and ownership/slots/complete
command review remain mandatory; patterns do not enforce a resource sandbox.
Portable source: [adoption guidance](../../.agents/skills/restruct-sdlc/sdlc-management/references/adoption.md#reusable-private-test-permission-rules).
