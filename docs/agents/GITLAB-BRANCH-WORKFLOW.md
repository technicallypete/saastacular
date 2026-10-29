# GitLab branch integration and cleanup

## Docker resource attribution

Use explicit project/run ownership labels for every newly created exclusively
owned Docker object that supports them, through Compose and direct commands.
Set `COMPOSE_PROJECT_NAME` per stack/run; direct creators need equivalent explicit
labels. Discover by labels, then verify exact identity/lifecycle using the existing
compact run registry. Preserve anonymous/shared exceptions and reboot recovery;
no extra cleanup manifest, label-only deletion or relabeling of active bound work.

Follow [label discovery and existing cleanup records](../../distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#project-scoped-docker-labels-and-cleanup-discovery).

## Acceptance and cleanup closeout publication

Routine acceptance and cleanup closeout records use the supervisor's direct publication route to the configured integration branch. The coordinator hands back the Result, attributable qualification evidence and cleanup receipt with exact source references; the supervisor reviews them, records the acceptance disposition and imports/publishes the completed Result, review, PROGRESS and related closeout records in scoped commits. Do not create a separate acceptance, Result-only or cleanup-receipt MR, request another automated review or launch CI solely to publish those records. The implementer receives no direct integration-branch push or self-acceptance authority. An explicit owner instruction or an actual branch-protection requirement can select an MR instead; report that reason. Actual code or executable configuration repairs retain the normal source-MR workflow. Preserve review, evidence, cleanup and native permission gates; publication alone is not ACCEPT.

## Required MR pipeline monitoring owner

Every required GitLab MR pipeline has one explicitly named monitoring owner,
normally the coordinator. Record the owner with the MR/pipeline ID and immutable
source SHA when the MR is opened or the pipeline is created. The owner verifies
the expected pipeline actually started, tracks it to a terminal outcome, reports
the result and routes the first failure for diagnosis. The Claude session-handoff
watcher observes Claude sessions; it does not inspect GitLab MRs, pipelines or
jobs. Auto-merge, a review request or a watcher notice does not prove that anyone
is monitoring the pipeline. If ownership changes, record a direct handoff and
acknowledgment; never leave a required pipeline implicitly owned. Keep updates
event-driven and concise, with no repetitive unchanged status reports. This
monitoring rule does not add or waive review, CI, merge or release authority.

## Post-merge staging qualification runs locally

Ordinary application feature work does **not** require the full `ci-controls` suite merely because it updates product test registration or expectations inside harness files. This includes routes/import allowlists, migration versions, schema/catalog counts, browser matrices, test invocations, required-check IDs and their resulting count pins. Run affected product checks and only the directly affected control cases using existing focused selectors. A changed harness pathname, copied fixture, mechanical count update or desire for an aggregate PASS row is not a reason to select all controls. Full CI controls require a recorded concrete reason involving changed shared runner behavior or an explicitly selected full staging/RC qualification. Required-check aggregates apply to selected coverage; do not add `ci-controls` merely to make `B-CI-CONTROLS` required. Preserve actual outcomes and report focused control proof without claiming an unexecuted full suite passed.

Owner direction, 2026-10-22 UTC:

Default sequence for implementation MRs: focused pre-merge checks and current hosted fast/supervisor/Duo review → authorized merge → affected local batches/suites on the exact merged integration SHA. The existing implementer coordinator automatically launches, owns and tracks those selected checks; no extra owner command or full pre-merge gate. Select from the actual change and integration delta, including affected dependencies, shared fixtures, migrations, disabled variants and security boundaries. Record executable selectors and coverage rationale in the existing Plan/Result; preserve immutable inputs, actual outcomes and cleanup. Final Goal ACCEPT requires that attributable affected coverage and the other Goal gates, rather than a blanket full run per feature.

Run full local qualification on the integration branch tip when the owner requests it **or the supervisor recommends it**, recording the reason and freezing the exact observed tip for the run. Full staging runs and full hosted RC-promotion runs provide broader opportunities to find gaps in feature selection; unexecuted checks are never reported as passing. Ordinary feature MRs retain hosted fast checks; RC promotion MRs retain required full hosted qualification. Automatic integration/release-branch post-merge and tag-push pipelines remain disabled, including lightweight checks; separately authorized diagnostic/provider runs remain distinct.

After a failure, rerun the failed checks and checks affected by the fix, reusing attributable unchanged passes. A broad local selection needs a concrete coverage reason; explicit technically justified full-run requirements remain binding. Reconcile active work at a safe boundary without restarting running batches or rewriting approved snapshots. Ready authorized successor implementation may overlap post-merge checks in isolation, respecting resource reservations; predecessor final ACCEPT still gates successor integration/final acceptance. Preserve owner holds, manual/security gates, documentation-only exemptions and failure history. This supersedes the blanket full-local-run-after-every-feature default.

[Current policy](../../distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#post-merge-qualification-runs-locally).

Ready, authorized successor implementation may overlap that local run **after predecessor merge**, in isolation with shared-resource locks respected. Predecessor final ACCEPT still gates successor merge/final acceptance, and its own owner retains repairs. Explicit owner/phase holds and documentation-only exemptions remain. This supersedes the earlier pre-merge/unmerged-successor variant. See [current policy](../../distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#local-qualification-after-merge).

## Default parallel scheduling

The Goal/Spec records priority, normally high if unspecified. The Plan targets
3–5 parallel implementation workers for high or 1–3 for low whenever feasible,
recording constraints; kickoff follows that allocation after readiness checks.
Independent, dependency-ready work and validation batches run in parallel by
default when ownership, writable paths, inputs and resource reservations permit.
Each dependency chain retains finish/fix/rerun gates; a failure holds its batch and
dependents, while unrelated safe work may continue. Preserve task-specific
sequencing, frozen validation sources, the shared test pool, sequential builds and
all merge/qualification gates. See [parallel scheduling](../../distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#default-to-safe-parallel-work-lanes).

## Duo feedback before merge — 2026-10-11

Before merging an MR, Codex inspects its current Duo review comments and
unresolved discussion threads using glab. Remind the owner of outstanding Duo
findings before carrying out the merge, identifying the affected behavior,
Codex's assessment and proposed disposition (correction, documented risk or
reasoned rejection). Passing CI does not dispose of review findings. Record the
disposition in the Goal review/Result and reply on the MR when authorized;
resolving a thread must not imply an unimplemented correction is complete.
Check the latest reviewed source after a refresh; a truncated Duo review is not
comprehensive coverage. This reminder does not add a separate approval request
when existing owner authority and the recorded disposition already cover the
merge. An unreviewed blocking security/correctness finding stops integration.

## Staging MR auto-merge default — 2026-10-20

Owner directs all non-promotion MRs targeting `staging`—including feature/Goal,
maintenance and repair MRs—to use auto-merge after every required review and
pre-merge gate clears. This is standing staging-integration authority; no second
owner merge prompt is needed. For an eligible MR, enable it with the reviewed
source SHA:
`glab mr merge MR_IID --auto-merge --sha REVIEWED_SHA --remove-source-branch --yes`.
Verify GitLab accepted the request and record whether it is scheduled or already
merged; a command exit alone is not proof. The owner may opt out for an MR.

Before arming, verify Codex pre-integration ACCEPT, applicable Duo review and
finding disposition, target/source reconciliation, required pre-merge evidence,
and every manual or owner hold. Never arm while a required review or
manual/optional qualification is pending: GitLab may merge immediately when its
native checks are green and does not enforce every repository gate. Early/draft
MRs record the auto-merge intent and enable it only after all holds clear. An
explicit owner hold/opt-out on a specific MR takes precedence. Disarm and
reassess on a relevant new commit or blocking finding. Preserve failed/manual
gate status honestly.

The bundled Restruct SDLC copy under `distributions/skills/restruct-sdlc/` is a
pinned copy; its upstream project remains the source of truth, and updating the
bundled copy is an explicit, reviewed version change. For the corresponding live
guidance in this adopting repository, publish directly to `staging` under
supervisor publication authority; do not open a separate MR solely for those
documentation changes. This exception does not alter the normal MR routes for
executable changes or RC promotion.

Post-merge staging disposition and final Goal ACCEPT remain distinct. Retain
source deletion for temporary branches. This default does not apply to
`staging` → `main`/RC promotion MRs, or authorize protected-branch deletion,
release promotion or final release tags.

## Feature and Goal branches

Owner clarification, 2026-10-10: final Codex pre-integration code-review acceptance
supplies standing authority to push the reviewed temporary branch and open its
MR to staging. Do not ask the owner again for these two actions or require that
future review before starting approved implementation. Merge authorization and
final Goal ACCEPT remain distinct; preserve validation and exact-content binding.

1. Implement in the Goal's approved isolated checkout and branch. Preserve its reviewed source/evidence bindings and obtain Codex pre-integration review before integration. Final ACCEPT awaits required post-merge affected local qualification and other Goal gates. Draft MRs may be opened earlier for review or CI; they do not authorize implementation or merging.
2. Before every local test/validation batch on the feature branch, including focused development, preliminary and background runs, fetch the remote target and rebase the temporary source branch in its worktree onto that fetched target tip if needed (`origin/staging` for ordinary Goal integration). Read-only safety/input checks needed to prepare the rebase may precede it. Finish required manual feedback and self-review, then repeat target readiness before pushing/opening the MR. Resolve conflicts while preserving both accepted changes; verify the fetched target SHA is an ancestor of the final source SHA. Record both SHAs, reconcile changed approved inputs/validation bindings before further execution, and update the Result/evidence mapping. Run required focused pre-MR checks against that rebased candidate; affected staging qualification runs locally after merge on the frozen exact merged SHA. Check the remote target again immediately before opening; if it advanced, refresh/rebase again, reconcile and rerun only affected checks, retaining valid unchanged evidence with explicit content/input provenance. Then push the reconciled source and use `glab mr create` with an explicit source and target `staging`. Keep the project's enabled "Delete source branch after merge" default and confirm the MR retains it; `--remove-source-branch` is optional when needed to set it explicitly. Supply a concrete title/description and validation evidence. Check for an existing MR for that source/target before creating another.
3. If staging advances after the MR opens, refresh/rebase the source again before integration and reconcile any conflicts under the existing scope and validation rules. Reconcile exact accepted content and applicable CI evidence; an MR is not permission to overwrite accepted work or bypass required checks. Do not recreate a broad validation campaign when unchanged evidence remains applicable.
4. Merge after Codex pre-integration review, applicable Duo review/findings disposition, required focused pre-merge evidence and passing MR fast checks. The standing staging auto-merge rule supplies integration authority once all gates clear; there is no routine owner merge prompt. Under the current `local-after-merge` policy, the existing Claude coordinator automatically owns affected local qualification of the exact merged SHA; automatic hosted staging full qualification is disabled. Final ACCEPT/completion and dependent integration wait for that qualification and other Goal gates. Isolated dependent implementation may overlap the post-merge local run. Use glab to arm auto-merge for the reviewed SHA. Do not integrate feature branches by direct push, local merge pushed to staging or cherry-pick instead of the MR route unless the owner explicitly approves an exception.
5. Resolve the source branch after either merge or MR closure. If merged, confirm the target contains the accepted changes and verify the exact remote source tip. If closed without merge, confirm it is unmerged, no other open MR or live task uses the branch, and unique commits/evidence are integrated or durably preserved in existing records, a tag, retained ref or bundle. Then confirm provider deletion or delete only that exact task-owned disposable source. Closure is not acceptance; do not discard unique unpreserved work or delete long-lived/protected branches or immutable tags. Record the MR outcome, source SHA and resulting staging SHA when merged.

   At Goal acceptance/closeout, separately audit all task-owned remote refs; do
   not infer cleanup from GitLab's “Delete source branch” setting or the MR
   source alone. Compare `git ls-remote --heads origin` with the task/MR
   inventory, confirm no open MR or live owner, and verify unique content is
   integrated or preserved in the existing Result/retained ref before deletion.
   Delete only accepted, reconciled task branches; retain unresolved or unique
   work with a reason and release condition. Reconcile local branches separately
   (`git branch -vv`, `git worktree list`): remote deletion leaves local refs
   behind. Refresh stale local `main` only when it has no unique work. Include
   exact project-owned Docker image IDs in the same receipt; retain shared/reused
   images and never use host-wide prune.

Owner direction, 2026-10-09: the pre-MR rebase rule applies to temporary feature/Goal worktree branches, including correction passes. Use the remote target, not a potentially stale local staging branch. Do not rewrite long-lived `main`, `staging` or `identity-qualification` for this rule; staging-to-main promotions keep their existing route. When an owned, already-published source needs a rebase, preserve its prior delivery SHA and push with an exact expected-remote-tip `--force-with-lease`, never an unconditional force push. Stop for unexpected third-party source changes. Rebased commits have new identities: old passing pipeline evidence is historical until reconciled to the new exact candidate. Under the current `local-after-merge` placement, focused checks are required before MR opening and affected local qualification runs after merge on the exact merged SHA; the merged-SHA checkout stays frozen. This supersedes the older pre-MR full-coverage/post-merge hosted-gate placement. The rebase rule does not grant implementation ACCEPT or merge authority.

Example (replace source/title/description for the actual reviewed change):

```sh
glab mr create --source-branch G004-identity --target-branch staging \
  --title "Add global identity and sessions" \
  --description "Core-owned identity and sessions; validation evidence is linked in the G004 Result." \
  --yes
```

Verify the project setting `remove_source_branch_after_merge=true` via glab for your GitLab project. Use this default for temporary staging MRs and verify cleanup after merging; the setting applies to merges and does not clean up an unmerged MR. After MR closure, follow step 5's ownership, live-use and preservation checks before deleting its source branch. Local branches, worktrees, transcripts and review evidence are separate resources. The owner's 2026-10-10 closeout direction below authorizes verified obsolete worktree cleanup; remote deletion alone is not proof that a local checkout is safe to remove.

## Owner work and stop boundaries

Before next-task planning or release, check
[recorded specification authority and owner stops](../../AGENTS.md#bounded-specification-authority-and-owner-stop-boundaries).
An existing Spec/Plan never overrides a stop. Missing Specs require an explicit
bounded owner mandate (inclusive endpoint, task set/range or pinned epic), not
an assumption from the previous merge. Observe phase limits and prevent advance
planning past the stop while current CI/closeout continues. Resolve ambiguous
before/after wording before crossing the boundary; preserve work at safe pauses.

## Planning overlap after integration

After verified integration into staging, Codex applies
[automatic planning](../../AGENTS.md#automatic-planning-after-integration): start
the earliest eligible roadmap task with an existing ready Goal/Spec, without
another owner prompt. Pending staging CI/final ACCEPT/cleanup permits provisional
planning, not implementation or predecessor acceptance. Record baseline, open
gates and one planner's ownership; reconcile integrated inputs under the release policy.
Specific owner holds remain. No Spec means report the gap unless an explicit
bounded supervisor-specification mandate covers it; no invented implementer scope.

<a id="execution-release-after-predecessor-acceptance"></a>
## Execution release after reviewed predecessor integration

After required pre-integration review and predecessor merge to staging, Codex
reassesses and may release the next approved task's isolated implementation while
post-merge affected local qualification/closeout runs under [the release policy](../../AGENTS.md#automatic-reassessment-and-execution-release).
Plan/input/scope, other dependencies, decisions, budget/resources and recorded
release remain mandatory. Final predecessor ACCEPT still gates successor merge
and final acceptance. Diagnose failures; pause only affected work and assign scoped
repairs to the predecessor implementation owner, using its own ledger/branch
in parallel with successor implementation. Unresolved security,
architecture/contracts, explicit task holds, native gates and G021 remain.
No extra routine owner prompt, unchanged full rerun or acceptance inferred from merge.

## Mechanical corrections during released implementation

Use [standing correction authority](../../AGENTS.md#autonomous-mechanical-corrections)
for in-scope test/fixture/pin/catalog parity and harness/import bookkeeping.
Record focused evidence and continue without a separate approval round; preserve
assertions, exact scope and escalation for security/contracts/unapproved schema.
This changes no required CI, retry permission, final review or merge authority.

## Shared skill source and public distribution sync

For a publishable change to a maintained shared skill, sync its mapped portable
files and manifest to the owner-selected distribution checkout, then commit/push
both repositories through their authorized routes. Verify both remote commits;
matching local files alone is not completion. Preserve explicit adaptations,
README/release metadata and held unaccepted work. Record partial syncs and held
divergence; no silent public sweep or automatic tag/release/live cutover. See
[portable guidance](../../distributions/skills/restruct-sdlc/sdlc-management/references/adoption.md#maintaining-a-shared-package-in-two-repositories).

## Long-lived branches

At each staging integration/observed external advance, Codex assesses all known
active planning, implementation and validation tasks under
[AGENTS.md](../../AGENTS.md#codex-assessment-of-staging-advances). Record whether
each task continues, reconciles at a safe boundary or pauses immediately, with
the staging SHAs and reason. Notify verified task owners and verify acknowledgment.
If tests must stop, preserve interrupted evidence and verify owned-process/resource
cleanup before rebase; restart affected coverage after binding reconciliation.
Do not automatically restart every run or cancel hosted pipelines for a new SHA.

Owner direction, 2026-10-18: newer commits on the same MR automatically supersede
its running `validate:checks` job. GitLab project auto-cancellation is enabled;
the MR workflow selects `on_new_commit: interruptible`, and only that job's MR
rule sets `interruptible: true`. This avoids parallel checks for successive
pushes/rebases. It does not suppress required checks on the latest candidate or
turn a canceled run into a pass. Integration-branch runs, manually started MR
integration jobs and credential-bearing identity qualification retain their
non-interruptible behavior; unrelated MRs do not supersede one another. No push
starts a pipeline (feature branches, `staging`, `main` and every tag; owner,
2026-10-19): staging qualification is the coordinator's affected local suites on the
exact merged SHA and no cloud pipeline runs after a staging merge. The MR
documentation-only suppression remains in place.
Existing pipelines keep their creation-time configuration; active feature
branches must reconcile the CI update before their later pipelines use it.

- `main`: release/promotion target. Under the current RC flow, promote only a qualified RC through its temporary `release/<RC>` MR (source deletion on; see the RC promotion runbook below). A direct `staging` → `main` MR, if ever used, explicitly disables source-branch deletion, retains staging, and follows the applicable promotion/CI rules. The project deletion default also applies to this target, so do not assume changing the target disables it.
- `staging`: integration branch; retain it.
- `identity-qualification`: long-lived protected test branch while the WorkOS qualification workflow is active. It selects the exact implementation commit for credential-bearing identity checks; it is not a feature or integration branch. Update it only for authorized qualification, using the reviewed protected workflow and secret boundary. It has no role as a merge source to staging and is not deleted after each qualification. Retire it only when that workflow is explicitly retired or replaced. It is not required permanently: review a bounded qualification-simplification follow-up in the project backlog. Keep it and the current gates until that replacement/retirement is approved.

Keep branch protections, exact `identity-staging` variable scope and named external-action requirements. This policy does not launch pipelines, release G004/G005 execution, change budgets or waive the G021 foundation barrier. It does not retroactively rewrite accepted direct integrations or require new validation of their unchanged content.

## Temporary solo-operator gate placement — 2026-10-09 (superseded by local post-merge qualification above)

Required affected/full local coverage precedes the MR. Its fast job runs
automatically, and its gate is optional/manual (`allow_failure: true` only in
that MR rule). A green MR therefore does not claim full hosted coverage. Staging
pushes retain automatic path-selected integration/full validation, with unchanged
documentation-only bypass. Codex can approve conditional integration based on the
actual diff and local/MR evidence; final Goal ACCEPT, dependent integration and
promotion to main wait for a qualifying staging disposition. Dependent isolated
implementation may overlap staging CI under the execution-release policy above. On staging failure, correct
or revert before promotion. This owner policy supersedes earlier MR-hosted-final
sequencing, including active G005, without altering frozen reviewed inputs.

## Earlier rebase-before-final rule — superseded by the clarification below

Owner makes this order explicit for implementation and correction passes:
implementation/focused checks/manual feedback → fetch/rebase remote target and
resolve/reconcile → required final local qualification → push/MR. This refines
the earlier pre-MR rule; it avoids validating a candidate immediately before a
planned rebase. Preliminary checks are allowed. Later target advances require
reconciliation and affected retests, not an automatic whole-suite rerun merely
because rebasing changed commit IDs. Relevant manual evidence must also be
repeated when its tested behavior changes. Preserve historical Plan/input bytes
and record active-plan sequencing amendments in the Result.

## Rebase before every local validation batch — 2026-10-09 clarification

The earlier preliminary-check exception is superseded. Before each local test,
build, lint/type-check, controls or validation batch on a temporary feature/Goal
branch, fetch the remote target and verify its SHA is an ancestor of the candidate
HEAD. Rebase/reconcile if needed before launch, including during development;
there is no preliminary/cycle-1/background exemption. Record target/candidate
SHAs and approved-input reconciliation. Multiple commands may form one batch;
refresh before a later batch rather than rebasing between its commands. Do not
repeat a rebase when the target is unchanged.

Read-only Git/input-integrity/resource-safety preparation is allowed first. Unsafe
rebase, dirty-work preservation or immutable-binding conflicts must be resolved
under approved procedures or escalated before tests. Do not silently refresh
frozen inputs. This applies to future active G006/G051 batches without changing
historical approved Plan/snapshot bytes or automatically canceling running work.
Retain failed/interrupted evidence and reuse unchanged passes only with explicit
content/input provenance. Hosted CI tests its submitted SHA without rebasing;
long-lived integration/protected refs are not rewritten by this rule.

## Staging pipeline acceptance status — 2026-10-09

Current clarification: under `local-after-merge`, there is no automatic hosted
full staging pipeline. A canceled or green hosted fast pipeline never replaces
actual affected local qualification on the exact merged SHA. The historical
`success`/`canceled` status rule below applies only when an adopting repository
explicitly configures a hosted post-merge gate; it does not define this
repository's current default. Record canceled evidence as canceled, never as
passing tests.

## Temporary draft and qualification MR closeout — 2026-10-11

Task closeout includes every temporary draft MR opened for implementation,
validation or qualification. Once its purpose is complete and no required run or
review still needs it, close each unmerged disposable MR using glab and verify
that it no longer appears among open MRs. Preserve its discussion, pipeline links
and evidence; closing the MR removes it from active work without deleting its
historical record. A delivery MR that still carries unintegrated work must remain
open until integrated or explicitly canceled/superseded.

Identify MRs by the task's recorded IID, source/target branches and ownership,
not merely their draft status or title. Before deleting an owned disposable probe
branch, verify its exact remote tip, retain required evidence, confirm that no
live run or other MR depends on it and that it contains no unpreserved delivery
work. This is a bounded exception for disposable qualification branches, not
permission to delete arbitrary unmerged branches. For a closed unmerged MR,
also confirm that unique commits and failure evidence are preserved and that no
other open MR or live task uses the branch. Keep long-lived branches and other
tasks' MRs intact.

Record closed MR URLs, removed probe tips and retained exceptions in the Result
or closeout review alongside branch/worktree cleanup. No pending
qualification is waived by cleanup.

## Local worktree and in-flight marker closeout — 2026-10-10

After a temporary Goal/feature MR is merged or closed without merge and its
remote source is deleted, close out its local Claude worktrees too. For a merge,
confirm integrated content; for closure, preserve unique commits and evidence
before deleting the source. In both cases confirm there are no outstanding local
commits/edits or live session, process or required resource still using the
checkout. Remote deletion without verified integration or durable preservation
is not permission to discard unmerged work. Retain long-lived checkouts and
active planning/execution worktrees.

Use `git worktree remove` for verified obsolete checkouts, then `git worktree
prune` for stale registrations. Do not force-remove dirty work: reconcile it or
preserve a recoverable archive outside the checkout first. Rebased/squashed work
needs content/provenance reconciliation when ancestry does not establish merge.
Record removed paths, retained exceptions and their reasons. Local branch
deletion is separate; preserve unique commits unless reconciled or archived.

Close stale in-flight/status markers belonging to finished worktree sessions in
the same closeout. Verify ownership and that the session/run ended before
clearing an ephemeral marker or lock; never clear another live run's marker.
Update mutable current-status records with the terminal disposition and retained
evidence references. Preserve historical Result/review records, delivery logs,
failed/interrupted evidence, transcript receipt markers and immutable bindings;
append a superseding status rather than rewriting historical facts. Do not turn
withheld/failed validation artifacts into passing evidence by removing markers.

This is integration closeout, not an automatic validation-job deletion hook or
a new cleanup framework. G051 records its closeout using this procedure without
expanding Docker ownership or removing an executing checkout.

Session-handoff closeout also covers stale pending/in-flight task indicators.
Reconcile each indicator against the actual session/task outcome (completed,
canceled, superseded or retained with a blocker); do not infer completion from
socket delivery, receipt, remote branch deletion or an idle process. Close or
supersede the current indicator with its task/session reference and evidence.
`management/SESSION-HANDOFFS.jsonl` is machine-local, gitignored/untracked
append-only delivery history, not a current-task queue: retain prepared/sent/receipt records and unique handoff IDs.
Record terminal task disposition in the existing Result/PROGRESS record (and
append a closure record if a task tracker needs it); never erase or rewrite
historical delivery attempts just to make the task appear finished.

## Delivery-log archive before prune — 2026-10-11 (G052)

The append-only rule above still governs the live delivery log: nothing is erased
to make a task look finished, and an unverified or active task's records stay live.
Its one exception is the handoff skill's `compact`. After Codex or the owner
verifies a standing terminal disposition (completed, canceled or superseded) from
committed integration records, `compact` copies that task's whole delivery groups,
byte for byte, into the private machine-local archive with a manifest, verifies the
archive, and only then removes those records from the live log. That is archiving,
not erasure: every byte stays in the live log or in a verified archive, failed and
unconfirmed deliveries keep their status, and archives, transcripts and approvals
are never deleted. Run it only from the main checkout under the
[skill's closeout procedure](../../.agents/skills/restruct-sdlc/claude-session-handoff/SKILL.md);
the first live apply is owner- or Codex-operated. Record the archive basename and
manifest SHA-256 with the closeout.

**Worktree-local delivery logs.** Before removing a linked worktree, check it for
`management/SESSION-HANDOFFS.jsonl`. If one exists, copy it byte for byte into the
same private archive and record both SHA-256 values in the closeout record. Do not
merge it into the main log, and do not remove the worktree if the copy fails:

```sh
(umask 077
 common=$(realpath "$(git -C MAIN_CHECKOUT rev-parse --path-format=absolute --git-common-dir)")
 archive="$HOME/.local/state/claude-session-handoff/$(printf %s "$common" | sha256sum | cut -d' ' -f1)/archive"
 mkdir -p "$archive"
 dest="$archive/$(date -u +%Y%m%dT%H%M%SZ)-worktree-$(basename WORKTREE).jsonl"
 cp --no-clobber WORKTREE/management/SESSION-HANDOFFS.jsonl "$dest" &&
   sha256sum WORKTREE/management/SESSION-HANDOFFS.jsonl "$dest")
```

The two hashes must match. This adds no migration of historical logs and no
automatic relocation; delivery logging stays on an explicit main-checkout `--repo`.


## Supervisor progress-index closeout — 2026-10-11

After final Goal ACCEPT, Codex also assesses PROGRESS.md compaction under
[Supervisor progress-index closeout](../../AGENTS.md#supervisor-progress-index-closeout).
Keep the current status, acceptance/review links and unresolved cross-task gates
in the index; archive superseded updates as dated committed Markdown within the
release management directory, preserving historical decisions and evidence.
This is separate from machine-local session-log pruning and safe branch/worktree
cleanup. Do not compact active task records or silently refresh bound inputs.
Every epic closeout is a mandatory compaction checkpoint: archive accumulated
superseded updates, or record that none exist or why active bindings require
deferral and the next safe checkpoint. Codex records this disposition before
closing out the epic; compaction alone requires no test rerun.


## Safekeeping checkpoints before review — 2026-10-11

[AGENTS.md](../../AGENTS.md#implementation-round-safekeeping-pushes) grants Claude
standing authority to commit/push a scoped temporary-branch checkpoint after
each implementation or correction round. This early backup is distinct from the
reviewed push/MR integration step above; pending validation is recorded honestly.
Preserve secrets/local-state exclusions and active immutable inputs. Final
rebase/reconciliation, review, validation and separate merge authority still apply.

## Local validation evidence closeout — owner direction, 2026-10-13

After final ACCEPT, integration and required qualification, apply the SDLC skill's
[compact-record and bulk-purge policy](../../distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#local-validation-evidence-retain-a-compact-record-purge-bulk-output).
This replaces indefinite retention of redundant task-owned local validation logs,
traces/videos, report copies and scratch outputs, including eligible older closed
Goals at their next cleanup. Verify retained acceptance/provenance/failure records
and necessary proof, exact owned paths, containment and no active/held/dependent
use before deletion; record retained hashes/locations, purged paths/count/bytes and
reclaimed space. Raw artifacts deliberately purged are unavailable, not passing
or reusable evidence. Preserve committed management history and the separate
unchanged session-transcript/delivery-log/archive protections above. Active tasks
and explicit retention holds stay protected; no blanket state-directory cleanup.
This is a policy amendment, not authorization to purge evidence needed by G008 or
G054, alter hosted expiry/CI, or prune shared Docker caches.

## Prospective exact-delta exception — 2026-10-14

For G009 onward, the later AGENTS.md future-goal rule supersedes mandatory rebase
for a fetched delta verified irrelevant to both validation and execution authority.
Record the target range, candidate and reason in the existing Result. No blanket
folder exemption; unknown/relevant changes reconcile. Resource safeguards still
run; final qualification and MR opening still require reconciliation. Active
G008 bindings and historical qualification are unchanged.

## Temporary local evidence — 2026-10-14

Keep durable review/outcome records in Git and use existing hosted CI artifacts.
Local raw output is temporary; purge duplicate downloads and eligible closed-task
bulk at closeout. Explicit necessary-proof holds state reason and release gate;
expiry does not justify mirroring every hosted run. Follow the SDLC evidence policy.
Session/binding/observer state, active work and unresolved investigations remain protected.


## Low-overhead delivery refinement — 2026-10-15

Follow [validation economy and recurring failures](../../distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#validation-economy-and-recurring-failures)
for focused evidence reuse, one evidenced job retry, recurring-failure repair,
superseded-run cancellation and the narrow verified workflow-prose CI-skip option.
No path-selector change or blanket directory exemption; required application
qualification and failed-check gates remain. Continue already-authorized MR/Duo/
merge transitions without another routine owner prompt. Use the existing
[closeout receipt](../../distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#routine-closeout-receipt).


## Mandatory task-acceptance cleanup — owner direction, 2026-10-17

At every final task ACCEPT, immediately remove all finished task-owned worktrees
and temporary resources under the [mandatory cleanup policy](../../distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#mandatory-cleanup-at-every-task-acceptance).
Unlock finished agent-tool worktrees; preserve unique commits and dirty/untracked
work before force removal when needed. Clean task-owned containers, unused images
(including attributable untagged IDs), volumes and disposable evidence without
another routine approval. Keep active coordinator/successor/preview resources and
specific recorded retention/native-permission holds. Verify a compact receipt;
no silent lock/dirty-work deferral, blanket host prune or extra test campaign.


## First failure requires intervention — 2026-10-18

Apply the [first-failure intervention rule](../../distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#first-failure-requires-intervention--2026-10-18): inspect every task pipeline outcome, diagnose the first failure, and test a correction or evidenced retry hypothesis before another diagnostic run. Native local-test blocks require a supported configuration remedy, not repeated failing hosted runs. Preserve checkpoint authority and failed evidence.


## Staging repair gate — owner direction, 2026-10-18

Owner direction after G019's repeated post-merge staging failures. One repair
MR missed a failure that was already in the previous run. Each miss cost about
an hour of staging time plus another MR, merge-approval and integration cycle.
A staging repair must fix every known failure before it merges.

- **Inventory every failure first.** From each failed run's `ci-evidence/results.tsv`
  (plus junit and the failing suite logs), list every non-PASS required check.
  Include checks from earlier failed runs of the same task that later runs did not
  re-execute. Record the list in the existing Result with the pipeline, job and SHA.
  Map each entry to a correction or to an evidenced disposition (for example,
  derived aggregate or timing hypothesis). Open no repair MR while an entry is unmapped.
- **Prove every failed check before merge.** On the exact repair candidate, rerun
  each failed check with its focused local selector and record it passing.
  Derive expected values in pins from the hosted output byte for byte. When a
  check cannot be reproduced locally, start the repair MR's manual
  `validate:integration` job and require it to pass before merging.
  Do not use the post-merge staging run to find out whether a repair works.
- **Handoff.** The repair handoff carries the inventory table: check, failing
  run, correction or disposition, and pre-merge proof. Codex review and the
  owner's merge approval read that table, not only the latest failure.
- Timing or flake hypotheses keep the existing
  [one evidenced job retry](../../distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#validation-economy-and-recurring-failures)
  rule; a recurrence needs focused diagnosis. The post-merge staging gate, the
  acceptance status rule and all native merge gates are unchanged.


### Local failure evidence retention — 2026-10-21

Local qualification and repair runs use `--evidence-dir` in task-private storage.
The wrapper retains its scanned report in `<evidence-dir>/candidate/`, including
`results.tsv`, suite logs, diagnostics and `junit.xml` when available. Verify these
files after finalization before a retry, repair handoff or acceptance review.
Keep original failure and later-run records separately. Withheld detail is an
explicit evidence gap; do not bypass scans or count JUnit phase/aggregate entries
as distinct underlying test failures. Follow the [portable retention rule](../../distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#preserve-local-validation-evidence-before-cleanup).

### Supervisor verification of repair completeness — 2026-10-18

Codex owns independent reconciliation of the repair inventory against the actual
failed runs and their suite logs, including earlier unresolved failures. A Result
summary, latest-run green subset or successful underlying test is insufficient
when its wrapper/pin still fails. Review all affected pin consumers, not only the
first mismatched count. Each row must say fixed with attributable passing proof,
unresolved, not rerun (with valid unchanged evidence referenced), derived failure,
or an explicitly authorized evidenced retry/disposition. Missing/skipped/canceled
or unavailable evidence is never a pass; a timing hypothesis is not a fixed defect.

Before pre-integration ACCEPT, Codex records completeness, candidate/input
attribution and remaining holds in the existing review. Reject incomplete repairs
before another merge/run cycle. Reuse unchanged valid evidence after relevant-delta
reconciliation; do not rerun unrelated suites merely because the SHA changed.
Use the existing Result table and review, not a separate tracker or new blanket
full-suite gate. One job retry remains bounded by the existing retry rule; preserve
both outcomes and stop for focused repair on recurrence. No acceptance is inferred
from a merge or a passing retry.

## RC-qualified main promotion — current owner direction, 2026-10-19 UTC

Owner requests staging RC/version → coordinator creates the annotated RC and internally opens a tag-pinned temporary-branch MR to main → the MR automatically runs the complete hosted suite → **optional advisory Duo review; changes become planned follow-ups** → owner-authorized manual promotion after success, or explicitly selected RC auto-merge → verify main content and delete the temporary source branch; if the MR closes without merge, preserve the RC tag and all qualification/failure evidence before deleting its disposable source branch → **wait for the owner's separate decision to tag main**.

The owner never manages the temporary branch. GitLab MRs need a source branch, so the coordinator pins release/<canonical-tag> to the exact RC commit and removes only that branch after verified merge or after the promotion MR closes unmerged and the tag/evidence are preserved. Closure never qualifies or promotes the RC. Keep staging/main/immutable RC tags and failure evidence. A release/ prefix alone cannot qualify the MR: verify RC object/commit, MR source, current complete evidence and prospective/actual main content. No automatic main merge follows green CI.

RC/final tag pushes and main pushes start no pipeline; the promotion MR is the sole full hosted release qualification. Its full job is required/automatic, including documentation-only RCs. Ordinary feature-MR Duo and other Goal/security/manual gates remain. Normal staging qualification stays coordinator-owned affected local suites after merge; full local tip runs are owner-requested or supervisor-recommended. The CI rules, the `mr-rc-promotion` attribution and `infra/ci/release-verify.sh` implement this flow once their MR is integrated; until then, and until a first real RC promotion runs, only fixture/CI-rule proof exists. Do not create a concrete RC expecting this flow before integration.

### RC promotion runbook (coordinator, under the owner's named instruction)

Every step runs from a fetched checkout (`git fetch origin --prune --tags`), saves its inputs as files and records the
helper's single result line in the release receipt (`release-verify.sh` writes the receipt; keep it with the RC's
Result). Inputs are data: listings from `git ls-remote origin` (complete, unfiltered) and JSON from authenticated
`glab api` calls; nothing in them is executed. A FAIL line stops the flow; never substitute the moving staging tip.

1. **Cut** (owner names `X.Y.Z-rc.N` and the reviewed staging commit): `git ls-remote origin >ls-cut.txt`, then
   `bash infra/ci/release-verify.sh name X.Y.Z-rc.N ls-cut.txt receipt.env` (canonical name; tag and
   `release/X.Y.Z-rc.N` absent remotely and locally; staging at cut recorded). Only then
   `git tag -a X.Y.Z-rc.N -m "..." <staging SHA>` and `git push origin refs/tags/X.Y.Z-rc.N` (no pipeline starts).
   Then `git ls-remote origin >ls-tag.txt` and `bash infra/ci/release-verify.sh candidate X.Y.Z-rc.N ls-tag.txt receipt.env`
   (one remote tag object and peel line equal to the local annotated tag; on first binding the peel must equal the
   staging commit recorded at cut; later runs compare the recorded identity and staging ancestry).
2. **Promotion MR** (internal plumbing): create `release/X.Y.Z-rc.N` at the peeled commit
   (`glab api -X POST projects/:id/repository/branches -f branch=release/X.Y.Z-rc.N -f ref=<peeled SHA>`) and open the
   MR to `main` with squash off and source deletion on. Duo review is optional and advisory for this MR; record actionable findings as scheduled/planned follow-ups
   under the policy below. No Duo approval gate or project-wide review-setting change is introduced. The MR's detached pipeline runs
   `validate:checks` and the required `validate:integration` automatically (full, `rc-promotion`).
3. **Qualify** after the pipeline ends: save `glab api projects/:id/merge_requests/<iid>` (mr.json),
   `glab api projects/:id/pipelines/<pipeline id>` (pipeline.json) and `glab api projects/:id/pipelines/<id>/jobs`
   (jobs.json, without retried jobs), download both jobs' `ci-evidence` artifacts, fetch staging, then
   `bash infra/ci/release-verify.sh qualified X.Y.Z-rc.N receipt.env <project id> mr.json pipeline.json jobs.json FAST_DIR GATE_DIR`.
   A PASS makes the RC **eligible only**; report it and wait for the owner's explicit merge instruction naming it.
4. **Promote** (owner instruction): fetch `main`; `bash infra/ci/release-verify.sh promotable X.Y.Z-rc.N receipt.env <main SHA>`;
   immediately `git ls-remote origin >ls-recheck.txt`, a fresh mr.json and jobs.json and
   `bash infra/ci/release-verify.sh recheck X.Y.Z-rc.N receipt.env ls-recheck.txt mr.json jobs.json` (the latest pipeline must be
   freshly `success` with the qualified job IDs; a retried job needs `qualified` re-run with its artifacts first);
   then merge at once through the API with
   `sha=<peeled SHA>` and squash off (manual mode: no queued auto-merge; no direct main push). GitLab binds only the source SHA; the
   target is not compared atomically. If `recheck` reports that main moved, re-run `promotable` on the new main and
   `recheck` again; a conflict or a different tree needs a new RC from staging.
5. **Verify**: fetch `main`; `git ls-remote origin >ls-promoted.txt`;
   `bash infra/ci/release-verify.sh promoted X.Y.Z-rc.N receipt.env <main merge SHA> ls-promoted.txt` (parents exactly expected
   main and RC; tree = RC tree; the merge is on the freshly listed remote main, which may have advanced since; `final`
   repeats that membership check pre and post). On a mismatch stop and escalate: no release claim, no final tag, no
   automatic repair. Then confirm `release/X.Y.Z-rc.N` is deleted after merge or MR closure (delete only that exact temporary branch after verifying its tip, no other open MR/live use, and retained RC tag plus evidence; closure is not promotion).
   Main pushes start no pipeline; no further local or hosted full run follows.
   Only when the promotion MR has already merged (an owner-selected auto-merge), steps 3–5 are replaced by one evidence-only
   `bash infra/ci/release-verify.sh merged X.Y.Z-rc.N receipt.env <project id> mr.json pipeline.json jobs.json FAST_DIR GATE_DIR <main merge SHA> ls-merged.txt`
   on the candidate receipt (it records `VERIFICATION_MODE=merged-evidence` and infers no merge mode, actor or authority); the manual steps 3–5 stay the default.
6. **Final tag** (a separate, later owner decision): `git ls-remote origin >ls-final-pre.txt`;
   `bash infra/ci/release-verify.sh final pre X.Y.Z-rc.N X.Y.Z receipt.env ls-final-pre.txt`; create the annotated `X.Y.Z` on the
   verified main merge and push it (no pipeline); `git ls-remote origin >ls-final.txt`;
   `bash infra/ci/release-verify.sh final post X.Y.Z-rc.N X.Y.Z receipt.env ls-final.txt`.

### Superseded RC flow history

The following earlier tag-pipeline policy is retained as history only; the current owner sequence above supersedes it.

## RC-qualified main promotion — owner request, 2026-10-19 UTC

Target workflow: canonical SemVer RC such as `0.1.0-rc.1` pins a reviewed staging commit and automatically receives complete hosted qualification. Promote only that qualified snapshot through a tag-pinned temporary branch/MR to main, with prospective and actual main tree identical to the tested tree. A changed target/content requires reassessment/new RC; never promote the moving staging tip under older proof. Main push and ordinary release tags do not start duplicate pipelines. Normal staging merges retain coordinator-owned affected local qualification and acceptance holds; full local tip runs are owner-requested or supervisor-recommended.

Capability Spec and implementation state at that time: **CI implementation pending; current entrypoint rejects all tags.** Do not cut a tag expecting this new pipeline until it is integrated. This request creates no tag/release or main-promotion action now.

### Owner-controlled RC release sequence — 2026-10-19 UTC

Once the pending capability is integrated: owner requests an RC of integrated staging work → annotated MAJOR.MINOR.PATCH-rc.N at the reviewed staging commit → mandatory complete hosted CI for that tag → **wait for explicit owner promotion instruction after success** → reviewed tag-pinned temporary-branch MR to main → verify actual main parents/tree against qualified RC → final MAJOR.MINOR.PATCH tag on verified main under the release instruction. No automatic promotion from green CI and no moving staging tip substituted for the tested tag. Main pushes and final tags have no automatic pipeline; no duplicate full run follows promotion. This request establishes workflow only, with no concrete version/tag/promotion action now.

The temporary promotion branch is internal coordinator plumbing: create/pin it and open the MR as part of the owner's authorized promotion, then verify its removal after merge or MR closure once unique work/evidence is retained and no other open MR/live use depends on it. Do not ask the owner to create/manage a branch or add a branch-approval step. Present only RC request, CI outcome, owner promotion decision and final release outcome.

### Standing authority for an owner-authorized RC series — 2026-10-21

When the owner authorizes a named RC series (or explicitly directs the next
candidate), that authorization covers the coordinator's bounded RC work without
per-step or per-candidate reconfirmation: diagnose and repair qualification
failures within the already approved scope, integrate the repair through the
normal staging review and merge gates, cut the next sequential immutable RC tag
from the latest staging commit observed at cut time, open its tag-pinned
promotion MR, and run the required full qualification. If qualification fails,
preserve the failure evidence and continue with the next candidate after an
in-scope repair; do not waive, hide, or relabel a failed check. An instruction
given to the supervisor in this session remains the owner's authority when
faithfully relayed to the coordinator; do not ask the owner to repeat it in the
coordinator session. The coordinator still reports material results and stops
for scope expansion, security/architecture decisions, exhausted budget, or
other existing escalation conditions. This standing authority does not permit
merging a promotion MR to `main`, enabling auto-merge absent its separate owner
instruction, creating a stable release tag, or deploying/publishing a release;
those remain separate owner decisions.


## Owner-selected RC auto-merge — 2026-10-19

The owner may explicitly enable auto-merge for a named RC promotion MR or a clearly recorded scope of RC MRs. Manual promotion remains the default when no such instruction exists. This owner instruction authorizes queuing promotion before CI finishes; once every required full-qualification job succeeds, the provider may merge without another owner prompt. Record the owner instruction, MR, immutable source SHA and selected mode. Before arming, verify the tag/source identity, required complete pipeline, non-squash merge, and conflict-free prospective target tree equal to the RC tree; source-pin the request. Preserve target-movement monitoring and post-merge parent/tree verification without claiming an atomic target lock. Retain and independently verify full project/pipeline/job/artifact proof before claiming qualified release. No failed, skipped, optional or partial coverage counts as qualification. Do not fabricate manual-mode proof or change API snapshots to satisfy a verifier. A manual-only helper's refusal of queued/merged MRs is a mode limitation, not a passing auto-mode proof; use truthful independent verification or a separately reviewed mode-aware adapter. Final release tagging remains a separate owner decision. Feature-MR auto-merge defaults alone never authorize an RC-to-release merge.

A named-MR instruction and any waiver apply to that MR only; they do not select another RC's merge mode or satisfy its qualification. Keep the manual runbook commands above for manual mode; do not claim `release-verify.sh qualified/recheck` supports auto-mode API states before an actual reviewed adapter change.


## Optional RC Duo review and planned follow-ups — 2026-10-19

RC promotion MRs may use Duo as optional advisory review; Duo is not a required promotion approval gate. Triage every actionable finding and record its disposition in the existing release Result/review. Changes arising from that review must be scheduled and planned as bounded follow-up work: identify the responsible Goal/task or backlog item, intended milestone, scope, acceptance criteria and required focused validation under normal authority. Do not leave actionable recommendations only in an MR comment or silently claim them fixed. Do not change the code or move/recreate an existing RC tag to satisfy a review. If a follow-up changes the release candidate or must ship before promotion, integrate it through the normal staging workflow, then cut and fully qualify a new RC tag under the applicable release authorization. The owner may instead retain a documented prerelease follow-up while the current RC proceeds, as explicitly selected for RC1's limiter finding. Preserve ordinary feature-MR Duo requirements, required full RC CI, security/owner dispositions, selected auto/manual merge mode and the separate final-tag decision. This policy creates no automatic RC2, unbounded repair mandate or new tracker.


## Owner-waived canceled RC qualification — 2026-10-19

The owner may explicitly cancel a named RC promotion MR's pipeline and direct that cancellation to satisfy the current RC promotion gate, treating it as success for the workflow decision only. Record an owner-authorized qualification waiver with the RC/tag/source SHA, MR/pipeline IDs, actual canceled/partial/failed job states, any known findings and the owner's instruction. Never report the canceled pipeline or unexecuted tests as PASS, fabricate green evidence, erase prior failures or claim complete qualification. This exception is specific to the named candidate/pipeline; it is not the default cancellation policy for feature MRs or other RCs. Schedule a follow-up immutable RC tag to run the complete required suite, carrying forward unresolved findings and the qualification debt in the existing release Result/backlog/Plan. The follow-up RC's full suite is required unless the owner separately changes that instruction. The owner may choose this waiver alongside the recorded auto/manual promotion mode; provider auto-merge still depends on the provider's actual cancellation behavior. If existing provider rules prevent merging a canceled pipeline, report that concrete limitation rather than changing project-wide protections, manufacturing a success job or treating a queue as a completed merge. No pipeline is canceled merely because this option is documented, and final release tagging remains separately owner-controlled.


## RC source selection at cut time

For an owner-authorized RC cut, fetch the configured integration branch and tag
its latest remote commit observed at cut time (`origin/staging` here). Do not
require an older acceptance/review SHA or exclude later integrated documentation
or implementation merely because it arrived after the cut was requested. No
additional owner confirmation or documentation-only ancestry check is needed
solely because staging advanced. An explicitly owner-pinned source or release
content exclusion still takes precedence.

Record the selected full commit SHA/tree and create the immutable annotated RC
tag there. A concurrent later staging push belongs to a later candidate; never
move the tag to chase it. The promotion branch/MR and required full qualification
use that exact tagged source. This source-selection rule does not accept pending
Goals, waive findings/CI, authorize an unrequested RC/main merge/final tag or
change the selected promotion mode. Existing tags and pipelines stay unchanged.

## Descriptive tag and release notes — owner direction, 2026-10-20

Keep tag names canonical SemVer for automation, but make every new annotated tag
message and GitLab Release entry identify the substantive Goals/tasks included.
Summarize each included task by ID and short title, call out material additions
since the prior candidate/release, and record the exact source commit. The GitLab
Release notes are the user-facing summary; do not use a generic version-only
description. For a final release, summarize the complete delivered scope since
the prior stable release and state qualification accurately. Never imply that
unaccepted or merely planned work shipped. Published tag objects are immutable:
do not move or recreate a tag to improve its annotation. If an existing release
entry is sparse, improve its notes while preserving the tag and its source.

## Continuous cleanup checkpoints — 2026-10-22

Track owned resources at allocation and removal in the existing run/task inventory;
reconcile before startup/resume/dispatch, at every handback and batch end, and after
failure/cancellation. Reclaim finished worker trees and scratch immediately after
preserving their handback. Any unresolved abandoned resources block new resource-
consuming work until repaired; verify actual absence, including nested resources
and anonymous volumes. Follow [continuous accountability](../../distributions/skills/restruct-sdlc/sdlc-management/references/closeout-risks.md#continuous-resource-accountability--mandatory).
This adds no coordinator polling or duplicate inventory framework.
