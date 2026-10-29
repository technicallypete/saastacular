# Adoption, parameters and composition

Adoption instantiates the management structure in a target repository. It is an agent-performed procedure checked
by `scripts/check_sdlc.py`; there is no generator framework.

## Repository-local private files

Default for newly started tasks: use `.local/` in the **main integration checkout**,
shared by all of its linked worker worktrees, with `evidence/<task-or-run>/`,
`secrets/<task-or-run>/` and `tmp/<task-or-run>/`. Record that absolute root in the
existing dispatch/Result; do not resolve it against a worker's own checkout or
create a different `.local/` per worker. Use existing supported path options;
this policy does not rewrite tools, immutable bindings or fixed native state.
Supported isolated-worker fallback: a worker that cannot reach the main checkout's `.local/` uses the documented
home-state path recorded in its dispatch (the legacy `~/.local` evidence/secret root) and never creates a
per-worker `.local/`.

Ignore `/.local/` in Git. Exclude `.local/` from every applicable Docker build
context (including Dockerfile-specific ignore files, which override a root
`.dockerignore`) and source/export/upload inputs. Publish only explicitly selected,
sanitized evidence; never upload the whole root or its secrets. Keep private
folders mode0700 and secret files mode0600; do not force-add private files.
This is operator/task-local storage, not an application production secret store.

Evidence and scratch follow mandatory acceptance cleanup. Secrets remain separate:
never recursively delete the shared root as evidence cleanup. Remove a finished
task's disposable credentials only under their recorded ownership/lifecycle;
retain shared active credentials without copying them into evidence or patches.
Durable evidence still belongs in compact committed records and authorized hosted
artifacts; the local folder is temporary, not a permanent archive.

Existing/in-flight tasks may retain their current home-directory evidence/secret
paths through final acceptance and cleanup. Do not move active files, change their
bound hashes, or rerun validation for this default. Existing shared credentials
remain at their current recorded paths until deliberately migrated; new tasks put
newly created private files under the new root, without duplicating shared secrets.
Native session transcripts, observer/transport state and operating-system temporary
files keep their supported locations. No blanket home-directory migration,
settings change or new artifact/secret-management framework is introduced.

## Repository-scoped Claude permissions

Keep reviewed, shared project rules in tracked `.claude/settings.json`. Keep
personal or machine-specific exceptions for that repository in ignored
`.claude/settings.local.json`; reserve `~/.claude/settings.json` for preferences
intentionally shared across projects and native classifier settings that require
home scope (see below). Do not publish home paths, credentials or
operator-specific evidence/lock/deletion rules as portable project defaults.
Ignore `/.claude/settings.local.json` and `/.claude/worktrees/`, while leaving
shared settings tracked. Never overwrite existing rules during adoption.

Changes to shared allow rules require actual owner authorization and repository
review/publication under that authority. Preserve the owner's chosen rule breadth;
do not silently narrow or broaden it. Tool permission does not replace Goal/Plan,
scope, budget, merge, production or native approval authority. This policy grants
no automatic permission-mode switch, settings self-modification or ability to
route around a native rejection.

Verify current tool documentation before choosing path syntax. In current Claude
project settings `/path` anchors at the session primary working directory/settings
source; `//path` is filesystem-absolute. A machine-specific absolute rule belongs
in personal local settings, not the shared file. Project trust, managed deny/ask
rules and hooks still apply. Verify loaded settings through the supported native
permission interface, especially for service-root launches and existing workers;
committed bytes, a merge and a successful unrelated command are not proof of
loading. Do not create/restart workers or rerun application tests solely to prove
configuration discovery. Record the actual limitation and use a safe boundary.

Before adopting, check all documented launch directories and actual worker loading;
reference one canonical rule set, avoiding divergent service/worktree copies.
Validate JSON and Git exclusions for config-only changes. Carry unresolved native
discovery/access issues into the existing consolidation/adoption Plan rather than
adding a permission framework or treating the task allow-list as a sandbox.

### Tool permissions and classifier rules

Treat tool allow/ask/deny rules and auto-mode classifier rules as separate layers.
Check the installed version's supported configuration scopes. Current Claude reads
`autoMode` from personal home settings, managed settings or supported invocation
configuration, not repository `.claude/settings*.json`. Repository-local tool
permissions remain the default; when native configuration requires home scope,
record that explicit exception and its cross-project effect. Do not create
ineffective project copies or change permission mode merely to finish adoption.

Classifier customization is an owner choice, not a distribution default. Preserve
existing settings and rule breadth; never ship personal paths, exact task-file
exceptions, blanket shell grants or disabled security rules in reusable templates.
For an authorized rule change, inspect effective rules, make the smallest supported
change, and compare the effective result. Replacing a complete default rule list
pins its remaining rules; reconcile upstream defaults after tool upgrades.

Verify loading with the supported permission/config interface and the next actual
approved edit/test/commit operation. Tool rules match command forms: a `git commit`
prefix need not cover `git -c ... commit`. A successful edit does not establish
that tests or commits are permitted. Do not add probe campaigns or restart workers
just to verify loading. On first refusal, capture the actual operation and native
reason, identify the layer, and use existing owner authority for a supported fix.
A remaining native gate stays explicit; do not switch executor or impersonate an
owner response. Preserve safe independent work and all evidence.

References: [tool permission configuration](https://code.claude.com/docs/en/permissions)
and [classifier scope and defaults](https://code.claude.com/docs/en/auto-mode-config).

### Permission change handoff

At adoption, name the authorized settings editor and record whether the owner
has granted standing authority for narrowly scoped permissions needed by approved
work. In a Claude-to-Codex deployment, Claude proposes the rules and Codex edits
the settings; another supported deployment names its available editor. This does
not require a permanent Codex supervisor in a Claude-only workflow.

On a permission block, the coordinator sends that editor the exact tool/command
or file, native refusal, task/scope justification and proposed literal rules or
settings patch. The editor checks scope and applies the smallest supported change
under the recorded authority, preserving existing settings. Use tracked project
settings for reusable shared rules, ignored project-local settings for task/machine
exceptions, and home scope only when the native configuration requires it. Validate
JSON/exclusions and report the file and exact installed delta back to the coordinator.

With standing authority, do not bounce routine approved-scope rules back to the
owner or leave them as instructions for the owner to paste. Broader authority,
security-rule removal or global policy changes require their own authorization.
A peer's request alone is not an owner grant.

The coordinator then verifies effective loading and the next actual approved
operation through the native tool. Report installed, loaded and operation-succeeded
states separately. On success, continue the same task/budget with one owner; on a
remaining native refusal, report its exact operation/reason and supported owner
handling once. No alternate executor, fabricated owner response, permission-mode
switch or broad test rerun is implied. Use existing handoff/Result records, not a
new permission registry or approval framework.


### Reviewed merges and classifier false positives

When the owner delegates merge authority to an independent supervisor, an
auto-mode rule requiring a human review must not silently replace that workflow.
An owner-authorized settings editor may narrow a false-positive classifier rule
through supported native configuration, limited to the adopting repository and
its worktrees. Recognize verified independent supervisor approval of the exact
source SHA and the workflow's current required CI/review disposition. Preserve
separate owner authority for release promotion and apply a named waiver only to
its explicitly waived gate; canceled or unperformed tests never become PASS.

Read-only status/ref/worktree listings and authorized post-merge closeout are not
merge operations. Missing review or authority, self-approval, unreviewed source,
unrelated targets and unapproved check/protection bypasses remain blocked. This
is an opt-in owner configuration, not a distributed disabled-guard default or
a blanket merge grant. Preserve the original rule outside the selected scope,
other classifier categories and all unrelated settings. Never bundle personal
paths or a full pinned native rule list into the reusable package.

Use the existing permission-change handoff: exact refusal and proposed patch,
authorized editor installation, effective-config comparison, then coordinator
verification of the next legitimate operation. Distinguish installed, loaded
and operation-succeeded; no synthetic merge, extra CI run, alternate executor or
mode switch merely to prove the exception. Reconcile pinned defaults on upgrades.

## Parameters

Every parameter is explicit. [`templates/parameters.example.json`](../templates/parameters.example.json) lists
each one with a description and a generic default. Copy it, edit the values for the target project and keep the
copy outside the shipped skill (for example in the adoption Plan).

- Identity and actors: `PROJECT_NAME`, `RELEASE` (default `v0`), `OWNER_ROLE`, `ARCHITECTURE_AUTHORITY`,
  `SUPERVISOR_AGENT`, `IMPLEMENTER_AGENT`.
- Sources and layout: `CANONICAL_SOURCES` (list), `SERVICE_ROOTS` (list, default empty).
- Branches and CI: `INTEGRATION_BRANCH`, `RELEASE_BRANCH`, `PROTECTED_BRANCHES` (list), `CI_PROVIDER`
  (`gitlab` or `other`), `MR_TOOL`, `FAST_CHECK_COMMAND`, `FULL_CHECK_COMMAND`, `PRE_BATCH_ENV_CHECK` (optional).
- Supervisor publication: `SUPERVISOR_AUTO_PUBLISH` (default `true`; see
  [roles](roles-lifecycle.md#supervisor-progress-publication-option-enabled-by-default)), publishing to the explicit
  `INTEGRATION_BRANCH`.
- Validation options: `FINAL_VALIDATION_PLACEMENT` (`local-after-merge` default: coordinator-owned affected local qualification of the merged SHA; full local tip runs on owner request or supervisor recommendation, no automatic post-merge cloud pipeline; other placements require explicit adoption),
  `CANCELED_PIPELINE_SATISFIES` (default `false`).
  An owner may also select [RC promotion qualification](validation-integration.md#owner-controlled-rc-promotion-option): full hosted proof belongs to a tag-pinned promotion MR, main/release and tag pushes have no duplicate automatic pipeline, promotion and final tagging remain distinct owner decisions. Configure/verify the adopting provider adapter explicitly; no portable script or project-wide permission change is implied.
- Budget and effort: `DEFAULT_BUDGET` (`B5`), `COORDINATOR_EFFORT` (`high`), `PLANNING_EFFORT` (`xhigh`),
  `IMPLEMENTATION_EFFORT` (`high`).
- Observer activation: `OBSERVER_AUTO_ACTIVATE` (default `true`, independent of supervisor publication).
- `DELIVERY_BARRIER` (optional, default empty): a roadmap barrier the project declares, if any.

**Rendering.** Replace each `{{NAME}}` with its value. A list renders as its items joined by `, `; an empty value
renders as `(none)`; booleans render as `true`/`false`. No other syntax exists: no conditionals or loops.

**Not generic defaults.** A particular project's delivery barrier, product/tenancy/tool-contract rules, identity
provider, UI kit, port conventions, branch and actor names, dated owner directions, one-off exceptions, runner
timings and environment workarounds are never defaults. An environment safeguard becomes `PRE_BATCH_ENV_CHECK`.

## Risk tiers

The supervisor states the tier in each Goal's Execution Policy; agents never self-classify to skip a gate.

- **Direct execution:** explicitly authorized, bounded copy/styling/presentation work inside existing behavior.
  No separate Plan; the Result records baseline, changed files and focused checks; the diff is still reviewed.
- **Ordinary:** a concise Spec and a short Plan (changes, criterion mapping, validation, real risks). One Plan
  approval authorizes autonomous implementation within scope.
- **Security, data or architecture:** explicit Spec/Plan review and stronger evidence for identity,
  authorization or isolation, security-model changes, migrations, durable identity decisions, public contracts,
  major dependencies or infrastructure.

## Procedure

1. Read the target's existing governance. Record the parameter values and the adoption authority in a Plan.
2. Install the `restruct-sdlc/` bundle into `.agents/skills/` with the collision-safe procedure in
   [Installing the skills](#installing-the-skills), record any collision decision, and commit.
3. For each template, compute its destination: strip `templates/`; `root/` maps to the repository root;
   `management/release/` maps to `management/<RELEASE>/`; drop `.tmpl`. Instance templates (GOAL, SPEC, PLAN,
   RESULT, REVIEW, PROGRESS-ARCHIVE) are copied per Goal or per archive with an ID-based filename.
4. **Create only missing files.** For an existing file, write the rendering beside it as `<path>.sdlc-proposed`
   and record the owner's merge decision. Existing governance is never silently overwritten or merged.
   4a. Render [`templates/root/.codex/config.toml.tmpl`](../templates/root/.codex/config.toml.tmpl) to
   `.codex/config.toml` only when `CODEX_CONFIG_TEMPLATE` is `true`; an existing file gets
   `.codex/config.toml.sdlc-proposed` instead. Never edit account-wide Codex settings; verify actual dispatch
   separately (see [Codex subagent model defaults](#repository-scoped-codex-subagent-model-defaults)).
5. Append [`templates/root/gitignore.lines`](../templates/root/gitignore.lines) to `.gitignore` (session log and
   private worktree roots and `/.local/`), skipping lines already present. Add `.local`
   to the root and each applicable Dockerfile-specific ignore file without replacing
   existing exclusions; exclude it from artifact/export selection too.
6. Create the release directories `goals/`, `specs/`, `plans/`, `results/` and `archive/`.
7. **Observer activation** (when `OBSERVER_AUTO_ACTIVATE` is `true`). Once the owner-selected coordinator is pinned
   and `pinned-live-verified`, and the existing supervisor UUID has been supplied explicitly (never inferred from a
   name or chosen among several sessions), run the companion's `watch.py bind` and then `watch.py start`, using the
   30-second default, with no further enablement question. If any binding is missing, ambiguous, unverified or
   retired, leave the observer inactive and record the unmet prerequisite. Never create sessions or change
   permissions to satisfy it. `start` is idempotent for a running observer. `false` opts out.
8. Run `python3 -B .agents/skills/restruct-sdlc/sdlc-management/scripts/check_sdlc.py adoption . --release <RELEASE>` and fix
   every finding. Optionally pass `--literals FILE` with project-private strings that must not appear. Private
   worker checkouts under `.claude/worktrees/` and `.codex/worktrees/` are not adopted content and are skipped.

## Installing the skills

Install the `restruct-sdlc/` bundle directory **whole**; its three children are never installed separately. Never
copy over an existing installation. From the distribution root (the directory that contains `restruct-sdlc/`):

```sh
TARGET=/path/to/target-repo
dest="$TARGET/.agents/skills/restruct-sdlc"
if [ -L "$TARGET/.agents" ] || [ -L "$TARGET/.agents/skills" ] || [ -L "$dest" ]; then
  echo "STOP: symlinked destination; nothing copied"
elif [ ! -e "$dest" ]; then
  mkdir -p "$TARGET/.agents/skills" && cp -R restruct-sdlc "$dest" && echo "INSTALLED restruct-sdlc"
elif [ -z "$(find "$dest" -type l)" ] && diff -r -q restruct-sdlc "$dest" > /dev/null; then
  echo "IDENTICAL restruct-sdlc: left unchanged"
else
  echo "COLLISION restruct-sdlc: existing content kept; record an owner merge or replacement decision"
fi
```

- **Absent:** copied. **Identical** (no symlinks, no differences): left unchanged.
- **Different** (edited, older or newer): the existing directory stays byte-identical. Record the collision, the
  `diff -r` summary and the owner's merge or replacement decision in the adoption Plan or Result; the bundle copy
  remains the proposal. A replacement happens only after that recorded decision.
- **Symlinked** `.agents`, `.agents/skills` or bundle destination: stop; nothing is written through a symlink.
- **Optional Claude alias:** the relative symlink `.claude/skills/restruct-sdlc -> ../../.agents/skills/restruct-sdlc`
  (create it only where no `.claude/skills/restruct-sdlc` exists; never replace another entry).
- **Earlier flat installations** (`.agents/skills/<child>/` from a previous package) are not a maintained layout:
  install the bundle, retarget local references, then retire the flat copies under a recorded owner decision.

## Installation: storage versus discovery

The supported layout is the single bundle directory `.agents/skills/restruct-sdlc/` (real files), committed in the
target repository, with its children at `.agents/skills/restruct-sdlc/<child>/`. Codex discovers the parent
`restruct-sdlc` skill there; the parent routes to the one child a task needs through a contained relative link, and
Claude sessions read the needed `SKILL.md` explicitly, as the adopted AGENTS.md requires. A distribution root
(`README.md`, `LICENSE`, `MANIFEST.tsv`, `restruct-sdlc/`) is **storage**, not a discovery location. A
source-maintainer repository may keep the canonical tree elsewhere (for example `distributions/skills/restruct-sdlc/`) and expose
it through relative aliases `.agents/skills/restruct-sdlc` and `.claude/skills/restruct-sdlc`.

The handoff helper's committed-path assumptions constrain adopters:
- `compact` verifies the helper committed at `.agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py`
  (or, in a source-maintainer repository, at `distributions/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py`);
- task labels match `G` plus three digits, and document references must lie under `management/`;
- closure records match `management/<release>/results/<ID>-implementation-codex-review*.md`, with
  `--verified-by codex|owner`, and the sender label is `codex-supervisor`;
- messaging is verified on Claude Code 2.1.290, 2.1.291 and 2.1.296, the creation gate only on 2.1.296;
- the helper needs Linux `/proc`.

The templates therefore default to `G###` IDs, root `management/`, `management/<release>/` and that review record
name. Changing them would change helper behavior, which needs separate scope.

**Unsupported:** other install paths (`compact` refuses them), separately installed or discovered children (the
Claude parent alias is supported as documented; children are not separately installed), discovery after `cd`,
personal installation, macOS or native Windows, other CLIs or versions (the helper fails closed), and any
promise of automatic discovery.

## Maintaining a shared package in two repositories

When the owner maintains a canonical project package and a separate distribution
repository, a publishable shared-package change includes syncing the mapped files,
updating provenance/manifest hashes and committing/pushing both repositories
through their authorized routes. Verify both remote commits before reporting
the sync complete. File equality alone is not publication; two independent Git
histories require two pushes. Do not push unrelated project/private files.

Mapped portable files should match the reviewed source; README, license/release
metadata and explicitly documented adaptations need not match the project tree.
Preserve pending unaccepted work and record any held divergence with its reason
and release condition instead of sweeping it into a public commit. If one push
fails, report the partial state and complete the outstanding sync when authorized;
never claim both are current. Tags/releases, breaking migrations, live cutover,
specific publication holds and native permissions retain their separate gates.
No nested Git repository, new submodule or synchronization daemon is implied.

At task acceptance/closeout, explicitly audit task-owned remote refs with
`git ls-remote --heads`; provider source-branch deletion is not proof that all
task branches were removed. Verify whether each MR merged or closed without
merge, confirm no other open MR or live task uses the ref, and ensure unique
content/evidence is integrated or durably preserved before deleting a
disposable branch. Closure is not acceptance. Audit local branches/worktrees
separately because remote deletion leaves local refs behind. Keep unclear or
unique work with a documented hold. Preserve immutable tags and failure
evidence. Include only exact task-owned Docker image IDs in cleanup;
shared/reused images stay, and host-wide prune is not part of adoption.

## Composing with a foundation or renderer skill

When this skill is used together with `saas-foundation` (or a similar renderer skill), **`sdlc-management`
supersedes only the overlapping management process**: management artifacts, actor responsibilities,
authorization flow, budgets, progress, reviews, validation sequencing, Git integration and closeout.

It does **not** supersede, and must preserve:
- the renderer's dry run and resolved-input review, informed write authorization, destination and conflict
  safeguards, and renderer safety rules;
- every required technical foundation check and technical verification;
- the repository's canonical governance, architecture and accepted decisions, including technical/security
  invariants, dependencies and delivery barriers.

Mapping: each foundation stage becomes an input to a Goal and its acceptance criteria. The renderer's dry-run plan
is **not** an Implementation Plan; the Goal still needs its own Plan and approvals. Existing authority satisfies a
repeated confirmation only within its explicit scope. Where the renderer's historical technical guidance differs
from accepted repository decisions, management precedence resolves nothing: canonical decisions prevail and an
unresolved technical conflict escalates. The foundation skill is neither edited nor bundled to apply this rule,
and neither skill assigns a supervisor role by being loaded.


## Owner-selected RC auto-merge — 2026-10-19

The owner may explicitly enable auto-merge for a named RC promotion MR or a clearly recorded scope of RC MRs. Manual promotion remains the default when no such instruction exists. This owner instruction authorizes queuing promotion before CI finishes; once every required full-qualification job succeeds, the provider may merge without another owner prompt. Record the owner instruction, MR, immutable source SHA and selected mode. Before arming, verify the tag/source identity, required complete pipeline, non-squash merge, and conflict-free prospective target tree equal to the RC tree; source-pin the request. Preserve target-movement monitoring and post-merge parent/tree verification without claiming an atomic target lock. Retain and independently verify full project/pipeline/job/artifact proof before claiming qualified release. No failed, skipped, optional or partial coverage counts as qualification. Do not fabricate manual-mode proof or change API snapshots to satisfy a verifier. A manual-only helper's refusal of queued/merged MRs is a mode limitation, not a passing auto-mode proof; use truthful independent verification or a separately reviewed mode-aware adapter. Final release tagging remains a separate owner decision. Feature-MR auto-merge defaults alone never authorize an RC-to-release merge.


## Optional RC Duo review and planned follow-ups — 2026-10-19

RC promotion MRs may use Duo as optional advisory review; Duo is not a required promotion approval gate. Triage every actionable finding and record its disposition in the existing release Result/review. Changes arising from that review must be scheduled and planned as bounded follow-up work: identify the responsible Goal/task or backlog item, intended milestone, scope, acceptance criteria and required focused validation under normal authority. Do not leave actionable recommendations only in an MR comment or silently claim them fixed. Do not change the code or move/recreate an existing RC tag to satisfy a review. If a follow-up changes the release candidate or must ship before promotion, integrate it through the normal staging workflow, then cut and fully qualify a new RC tag under the applicable release authorization. The owner may instead retain a documented prerelease follow-up while the current RC proceeds. Preserve ordinary feature-MR Duo requirements, required full RC CI, security/owner dispositions, selected auto/manual merge mode and the separate final-tag decision. This policy creates no automatic RC2, unbounded repair mandate or new tracker.


## Owner-waived canceled RC qualification — 2026-10-19

The owner may explicitly cancel a named RC promotion MR's pipeline and direct that cancellation to satisfy the current RC promotion gate, treating it as success for the workflow decision only. Record an owner-authorized qualification waiver with the RC/tag/source SHA, MR/pipeline IDs, actual canceled/partial/failed job states, any known findings and the owner's instruction. Never report the canceled pipeline or unexecuted tests as PASS, fabricate green evidence, erase prior failures or claim complete qualification. This exception is specific to the named candidate/pipeline; it is not the default cancellation policy for feature MRs or other RCs. Schedule a follow-up immutable RC tag to run the complete required suite, carrying forward unresolved findings and the qualification debt in the existing release Result/backlog/Plan. The follow-up RC's full suite is required unless the owner separately changes that instruction. The owner may choose this waiver alongside the recorded auto/manual promotion mode; provider auto-merge still depends on the provider's actual cancellation behavior. If existing provider rules prevent merging a canceled pipeline, report that concrete limitation rather than changing project-wide protections, manufacturing a success job or treating a queue as a completed merge. No pipeline is canceled merely because this option is documented, and final release tagging remains separately owner-controlled.


### Reusable private-test permission rules

With explicit owner authorization, use stable project-owned test resource names
rather than a new permission grant per Goal. Merge these example narrow Bash
allows into tracked `.claude/settings.json`, preserving every existing rule:

```json
[
  "Bash(docker build --label platform.test_run=* -t platform-test-*)",
  "Bash(docker run --rm --network none --label platform.test_run=*)",
  "Bash(docker compose -p platform-test-*)",
  "Bash(docker image rm platform-test-*)"
]
```

Choose/record the adopting project's reserved prefix/label (example `platform-test-`
and `platform.test_run`) and include task/run identity in names and label values.
Commands must use the reviewed argument order to match. This covers private image
builds, network-disabled disposable checks, project-scoped Compose and tagged-image
cleanup; unmatched commands/untagged-ID cleanup retain normal native handling.
Do not replace these with blanket Docker, arbitrary-shell or host-prune grants.
Allow matching is a workflow convenience, not a resource-ownership sandbox:
review the complete actual command, contexts/mounts/options/targets and task
ownership; never append unrelated cleanup targets or use a prefix as authority.

Evidence Write/Edit rules use the owner's repository-specific evidence root in
ignored `.claude/settings.local.json` when machine-absolute, never published home
paths. Secrets/settings/native session directories are not generic evidence roots.
Shared provider/GitLab rules and other owner preferences remain unchanged.
Permanent reusable rules survive task cleanup; remove only task-specific bridge
rules listed in that task's installation receipt. Keep an active task's existing
names/allowances until its safe transition or acceptance cleanup; do not restart,
rename resources or repeat tests merely to adopt the prefix.

Check JSON, Git exclusions, actual loaded sources and actual tool success separately.
Supported allow rules normally resolve before the auto classifier, with documented
exceptions; they do not disable native worktree checks, protected-path/managed
constraints or session permission controls. Installed is not loaded or successful.
Keep the shared private DB/browser cap, sequential build slot, ownership/cleanup,
source binding, Goal scope and review/validation gates. This is a portable adoption
pattern, not automatic settings self-modification or a blanket bypass.


### Repository-scoped Codex subagent model defaults

Where the owner selects mixed-model supervision, put shared project Codex settings in `.codex/config.toml`: `[agents]` with `default_subagent_model = "gpt-6.1-sol"` is the current owner's example. Leave the main-session model/effort and unrelated personal configuration unchanged. Explicit per-agent overrides take precedence; verify actual configuration loading/trust and dispatch separately, do not launch a test agent merely to prove the setting. No extra persistent session is needed. See [model/delegation defaults](coordination.md#codex-supervisor-model-and-delegation-defaults--owner-direction-2026-10-20). This optional adoption pattern does not install account-wide settings or add a required package file. Routine supervision stays in the main session; Goal/Spec drafting and substantive independent reviews use bounded delegates.
