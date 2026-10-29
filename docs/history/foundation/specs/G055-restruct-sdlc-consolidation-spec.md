> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G055 Work Spec — Preserved skill-suite packaging

Revision: r3, 2026-10-20. [Goal](../goals/G055-restruct-sdlc-consolidation.md).
Status: owner-directed respec; revised Claude Plan required before implementation.
Exact previous Spec (not included)
preserves the superseded unmerged v2 requirements and all earlier amendments.

## Owner direction and outcome

Preserve the three proven skills and distribute them together as **Restruct SDLC**.
A thin parent `restruct-sdlc` skill routes to their existing child `SKILL.md` files;
it does not absorb or rewrite their responsibilities, instructions or helper engines.
A mandatory major 2.0.0 upgrade and multi-day soak are not requirements. Any later
version/release follows actual reviewed scope and separate release authority.

Owner moved the outside repo to `../../Restruct/sdlc-skill` and permits completely
changing it, including discarding pending changes if needed. That local pathname
is independent of the installed skill name and Git remote repository name.

## Required package and discovery

Beginify source:

```text
skills/restruct-sdlc/
  SKILL.md
  sdlc-management/           # its SKILL.md and complete existing contents
  repository-supervisor/    # its SKILL.md and complete existing contents
  claude-session-handoff/   # its SKILL.md and complete existing contents
.agents/skills/restruct-sdlc -> ../../skills/restruct-sdlc
.claude/skills/restruct-sdlc -> ../../skills/restruct-sdlc
```

Outside distribution root:

```text
sdlc-skill/
  README.md
  LICENSE
  MANIFEST.tsv              # reuse existing inventory format as appropriate
  restruct-sdlc/
    SKILL.md
    sdlc-management/
    repository-supervisor/
    claude-session-handoff/
```

Other already-needed notices/metadata are allowed. No new installer is implied.
The public bundle contains real files, not links back to private Beginify paths.
Parent frontmatter name is `restruct-sdlc`; child names/frontmatter stay intact.
Four `SKILL.md` files are intentional. Only the parent needs native discovery;
it explicitly instructs the agent to read the needed child through a contained
relative link. Never depend on nested automatic discovery, separately installing
children, or loading all three bodies for routine status/messaging.

The complete management bootstrap is part of the `sdlc-management` child and the
outside bundle. Retain its current adoption instructions, parameters, both root
`AGENTS.md` and `CLAUDE.md` governance templates, and all release-scoped templates.
During explicit adoption, render either root template when its destination is
missing. If a destination already exists, preserve it and emit the established
`.sdlc-proposed` file for owner review; never overwrite project-specific guidance.
A chosen adoption can set up the full `management/<release>/` tree:
`ROADMAP.md`, `PROGRESS.md`, `BACKLOG.md`,
`goals/GOAL.md`, `specs/SPEC.md`, `plans/PLAN.md`, `results/RESULT.md`,
`results/REVIEW.md` and `archive/PROGRESS-ARCHIVE.md`, plus root budget and
supervisor-handoff records. Preserve per-task templates and archive/Result/review
procedures, not merely the `sdlc-management/SKILL.md` entrypoint. This is available
to a new or existing adopter who explicitly chooses and runs adoption; installing
or discovering the bundle alone does not create a management tree, fill project
decisions, start agents or activate a watcher. Bootstrap remains parameterized,
conflict-safe and subject to the adopting repository's owner authorization.
Include an optional repository-scoped `.codex/config.toml` template for the
selected main/subagent model and effort defaults. Current owner defaults are
`gpt-6-luna`/medium for the main session and `gpt-6.1-sol`/high for subagents.
Keep these values reviewable/parameterized; adoption renders the file only when
the owner selects it. If the target exists, preserve it and emit a proposal for
owner review. Never edit account-wide settings or claim a configured model was
used without verifying actual session dispatch.
The optional template's exact default content is:

```toml
# Shared model defaults for this repository. Individual sessions may override them.
model = "gpt-6-luna"
model_reasoning_effort = "medium"

[agents]
default_subagent_model = "gpt-6.1-sol"
default_subagent_reasoning_effort = "high"
```


## Preservation and relocation

Pin current committed standalone staging skills as behavioral baseline in the
existing Plan inventory. Preserve complete contents, modes, responsibilities,
notices and accepted policies. Unchanged moved files remain byte-identical.
Necessary sibling/resource/governance/template/checker/manifest path adaptations
are allowed; classify a bounded source→destination map and every non-move diff.
Clarify and preserve the existing communication rules: supervisors rely on
the existing Claude session-handoff watcher and do not poll coordinator sessions
unless the owner specifically asks. The watcher already polls; no watcher code or
new monitoring mechanism changes.
Do not rewrite instruction bodies for packaging, introduce consolidated helper
logic, or ship unaccepted G054/v2 features from the previous candidate. Reuse
accepted policy changes without treating its entire implementation as approved.

Move `skills/sdlc-management/` inside the bundle and retire its old standalone
source. Likewise relocate repository-supervisor and claude-session-handoff from
old discovery locations, then retire duplicate root discovery entries. Retarget
active AGENTS/CLAUDE/runbook/template/checker/consumer links and calls. Keep one
maintained source per child and unrelated skills intact. Closed records keep
original path/evidence text; Git history and the move map explain retired paths.

Keep current accepted workflow in children/templates: roles and owner-selected
supervision, phase/dependency/budget gates, model/effort defaults, 150k/200k context
limits and safe compaction, concise event-driven reports, permission handoffs,
private DB/browser cap two (third by permission) and sequential builds, reviewed
feature integration then coordinator-owned local qualification, RC promotion/
waiver/auto-merge and latest-staging cuts, descriptive tags/releases, complete
repair inventories with independent review, mandatory task cleanup and explicit
remote-branch/image checks. Preserve the supervisor no-poll rule in the bundled
coordination/supervisor instructions and root adoption templates: rely on watcher
notices, verify linked repository records and do not poll the coordinator or ping
the watcher absent a specific owner request. Retain the private-storage policy: new work uses the
ignored `.local/{evidence,secrets,tmp}/` in the main integration checkout, shared
with linked worker worktrees; existing/in-flight tasks may keep recorded `~/.local`
paths through acceptance, and isolated workers use only the documented supported
fallback. No blanket migration or secrets copying. Exact semantics come from
current canonical governance, not superseded prose in the unmerged candidate.

## Runtime and private-state invariants

Retain supported commands, authority/code-integrity checks, target-repository
selection, session discovery/receipt, observer identity/deduplication and private
namespaces. Resolving a skill through an alias must not silently select the wrong
target Git repository. Use contained canonical paths and existing supported
approval/binding mechanisms; never weaken checks to pass a relocation test.

Inventory live helper/observer consumers before deleting old paths. If necessary,
use the existing supported safe-point stop/bind/start procedure while preserving
one live observer and all state/history. Temporary old-script forwarders only
when actually needed/reviewed using an existing mechanism; no duplicate engines,
new watchers, automatic session replacement or state reset. Do not move personal
installed copies or private state as incidental packaging.

Use task/run-scoped ignored `.local/tmp/` for accessible coordinator scratch
instead of undifferentiated `/tmp`. Preserve G055's bound home-local evidence/
secret roots and supported isolated-worker fallback; controlled disposable HOME
fixtures may retain their existing temporary root. Exclude `.local`, credentials,
evidence, pins/transcripts/socket state and `.git` from both package exports.

## Outside checkout and publication

Verified local root: `<home>/Code/VitruvianTech/Restruct/sdlc-skill`.
Current origin: `https://github.com/technicallypete/restruct-sdlc.git`.
Inventory branch/HEAD/origin and tracked/untracked pending package changes first.
Owner explicitly authorizes discarding those pending changes and completely
replacing package/README content for this bundle; prior dirty-overlay, specific
README and stage-only export holds no longer apply. Preserve `.git`, committed
history, remote configuration and unrelated/private state. No force push/history
reset, GitHub rename, public tag/release or discard in other checkouts.

After approved Plan/review gates, normal commit/push of the reviewed bundle is
authorized through the existing two-repository publication route. Verify file/mode
hashes, licensing, neutral installation docs and no private material; document any
necessary neutral distribution adaptation. Verify actual remote commits in both
repositories before claiming publication. No submodule, package-manager/registry
framework, auto-updater or cross-project code-binding engine is added.

Project installation is the team default: copy/link the whole bundle and provide
parent discovery aliases as supported. Optional personal installation may use the
same bundle and existing target-selection/approval rules; do not add the abandoned
v2 home-runtime machinery to make this optional route work. Document a genuine
limitation rather than silently broadening scope.

## Plan revision and validation

Claude writes a concise linked revision (r4 recommended), refreshed input inventory
and allow-list; preserve previous Plans, reviews, candidate commits and the existing
Result. Record retained/reused/withdrawn v2 work, criterion mapping, actual budget
usage and any live cutover. No implementation against the new design until Codex
accepts/releases the revised Plan/scope/bindings. Latest Result reports attempt 1
of B5/checkpoint 3; verify usage, do not reset. Replanning is not an execution attempt.

Test changed guarantees: parent routing/child reads, both discovery aliases and
containment, complete source preservation, copy/install portability, active links,
relocated commands and current checker/affected child tests. Reuse existing tests,
fixtures and attributable unchanged passes; only add checks for uncovered changes.
A fake listing/file read is not actual harness discovery. No child native discovery
requirement, broad new compatibility campaign or arbitrary observation period.

Old M1/F1/G8a/G8b labels stay historical. The revised Plan maps still-relevant
live/discovery/security guarantees to focused evidence and removes consolidation-
only/home-engine/README gates. Owner's existing manual-review hold stays in force
if genuine owner action is still needed; do not automatically ask for obsolete
consolidation M1. Real native discovery/safe cutover and current repository CI/
integration gates cannot be waived by a packaging label. Feature fast/Duo → merge
→ local full still governs executable integration; this respec is documentation
only. Management-only SHA changes do not justify repeating unchanged coverage.

## Acceptance criteria

| ID | Required outcome | Evidence |
|---|---|---|
| AC1 | Parent and all three complete, named preserved child skills; four intentional SKILL.md files. | Baseline move/diff/mode inventory; parent ≤200 physical lines. |
| AC2 | Both discovery aliases resolve to canonical bundle; parent explicitly selects/loads children. | Containment/link checks, genuine supported parent-discovery evidence and focused child-read/routing smoke. |
| AC3 | Explicit adoption can bootstrap the complete parameterized management/docs tree and both root `AGENTS.md`/`CLAUDE.md` governance files. Missing files are rendered; existing files are preserved with `.sdlc-proposed` drafts. Optional project-scoped `.codex/config.toml` model defaults use the same opt-in/conflict-safe behavior. | Inventory shipped templates and use existing adoption/checker tests in a disposable repository to prove missing-file creation, existing-file preservation/proposal behavior and selected Codex config setup; package discovery alone creates nothing. |
| AC4 | Child behavior, notices and current accepted policy preserved; justified relocation/packaging edits plus the clarification of existing supervisor communication behavior only. | Actual source diff/move map; evidence the watcher behavior is unchanged; affected existing tests and attributed reused evidence. |
| AC5 | Old standalone source/discovery retired safely, active references work, unrelated skills/history preserved. | Consumer/link inventory and necessary existing safe-cutover proof. |
| AC6 | Helper target/authority/integrity/session/state guarantees survive relocation. | Relevant existing positive/negative tests and actual live-binding reconciliation when affected. |
| AC7 | Moved outside repo has real, portable, licensed bundle with correct installation docs and no private files. | Reviewed replacement inventory, source/export hashes and existing public/secret checks. |
| AC8 | Revised Plan/inputs/scope accepted, both repository updates published and required qualification/Result truthful. | Reviews, actual remote refs, commands/statuses, failure history and evidence attribution. |
| AC9 | Task resources, stale refs/worktrees/images and scratch reconciled at final ACCEPT. | Existing scoped cleanup receipt, preserving active/retained resources and native holds. |
| AC10 | Adoption preserves shared `.local` and legacy `~/.local` transition/fallback behavior without migrating active task state. | Verify rendered governance and adoption instructions identify the new-task root, linked-worktree sharing, ignore/export exclusions, existing-task retention through acceptance and supported isolated-worker fallback. |

AC1–AC10 replace the previous consolidation-specific fourteen-criterion design;
old results remain historical. Current delivery order, G020 implementation hold,
OS001 pause and G021 product barrier are unchanged.

## Historical inbound anchors

<a id="validation-overhead-inheritance--2026-10-14-utc"></a>
Preserved validation amendment (not included).

<a id="distribution-assessment-before-saastacular--2026-10-16"></a>
Original assessment (not included).
Current design uses the ordinary bundled directory and existing installation options.

Owner clarification, 2026-10-20: G055 must also preserve the reusable live-guidance
update rule: whenever a supervisor updates Restruct SDLC workflow rules, update
the adopting repository's corresponding live governance files in the same batch
and commit/push to its configured integration branch under its publication policy.
This applies to all adopting repositories, preserving local settings and task
bindings. Include this alongside the no-poll rule in Plan reconciliation; no
watcher implementation change or execution release is implied.

Owner direction, 2026-10-20: apply future Restruct SDLC workflow changes to
the corresponding live guidance in every adopting repository in the same update
batch, publishing to its configured integration branch under local policy. G055
must package and preserve this rule. Reconcile it with the no-poll update in the
next Plan submission; implementation remains held pending Plan review/release.

### Owner clarification: immediate asynchronous coordinator handoffs — 2026-10-20

The supervisor sends needed coordinator decisions, corrections and requests
immediately using nonblocking delivery, then continues its work without waiting
for a reply. The existing watcher delivers acknowledgment and responses
asynchronously. No-polling restricts status checks, not outbound communication.
G055 must preserve this clarification in its bundled guidance and templates.
