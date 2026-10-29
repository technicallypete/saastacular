# Validation and integration

## Post-merge qualification runs locally

Default sequence for implementation MRs: focused pre-merge checks and current hosted fast/supervisor/Duo review → authorized merge → affected local batches/suites on the exact merged integration SHA. The existing implementer coordinator automatically launches, owns and tracks those selected checks; no extra owner command or full pre-merge gate. Select from the actual change and integration delta, including affected dependencies, shared fixtures, migrations, disabled variants and security boundaries. Record executable selectors and coverage rationale in the existing Plan/Result; preserve immutable inputs, actual outcomes and cleanup. Final Goal ACCEPT requires that attributable affected coverage and the other Goal gates, rather than a blanket full run per feature.

Run full local qualification on the integration branch tip when the owner requests it **or the supervisor recommends it**, recording the reason and freezing the exact observed tip for the run. Full staging runs and full hosted RC-promotion runs provide broader opportunities to find gaps in feature selection; unexecuted checks are never reported as passing. Ordinary feature MRs retain hosted fast checks; RC promotion MRs retain required full hosted qualification. Automatic integration/release-branch post-merge and tag-push pipelines remain disabled, including lightweight checks; separately authorized diagnostic/provider runs remain distinct.

After a failure, rerun the failed checks and checks affected by the fix, reusing attributable unchanged passes. A broad local selection needs a concrete coverage reason; explicit technically justified full-run requirements remain binding. Reconcile active work at a safe boundary without restarting running batches or rewriting approved snapshots. Ready authorized successor implementation may overlap post-merge checks in isolation, respecting resource reservations; predecessor final ACCEPT still gates successor integration/final acceptance. Preserve owner holds, manual/security gates, documentation-only exemptions and failure history. This supersedes the blanket full-local-run-after-every-feature default.

## Ordinary feature registration does not require full CI controls

Routine product work may update product-specific assertions and test registration
inside shared validation scripts: permitted routes/imports, migration versions,
schema/catalog inventories, browser scenario matrices, test invocations, required
check IDs and expected totals in control fixtures. These edits do not by themselves
change the generic runner's execution behavior or invalidate every control.

For such changes, exclude the full `ci-controls` suite from ordinary feature
qualification. Run the affected application/security/migration checks and the
specific control cases whose inputs, assertions or expected totals changed, using
the existing focused selector. Record the changed guarantee and exact selector in
the existing Plan/Result. Reuse attributable unaffected passes. A harness pathname,
a copied fixture, mechanical registration/count-pin changes, or a missing full-suite
PASS row is not justification for selecting every control.

Changes to shared execution semantics (dispatch/profile selection, gate enforcement,
process supervision, isolation, resource accounting/cleanup or evidence handling)
require their affected control cases. The supervisor may select the whole control
suite only with a concrete recorded dependency/coverage reason explaining why the
focused set is insufficient. Do not replace that analysis with "harness changed".
An explicitly requested/recommended full staging run and required full hosted RC
qualification retain their complete coverage.

Required-check accounting does not create a new coverage requirement merely to
obtain an aggregate PASS: select checks from actual impact first. In adapters where
`B-CI-CONTROLS` represents the entire suite, do not select that umbrella suite just
to satisfy its own row; preserve focused control evidence in the existing Result.
Report an unexecuted suite as unexecuted, never as PASS. A genuine adapter limit
needs a bounded stated remedy/disposition; it must not silently force every feature
through the entire control suite. Apply this rule before the first qualification
run and when reviewing remaining selections at a safe boundary.

## Owner mandate and stop bounds

Check [recorded mandates/stops](roles-lifecycle.md#bounded-specification-authority-and-owner-stop-boundaries)
before every continuation. A ready Spec/accepted Plan or green pipeline does not
override an owner stop. Do not start successor lookahead while finishing the
endpoint task's CI/closeout. Auto-specification needs explicit bounded owner
authority; an epic mandate neither expands membership nor resets Goal budgets.

## Planning overlap

At planning and dispatch boundaries, default to [safe parallel batches](coordination.md#default-to-safe-parallel-work-lanes)
when their dependencies, ownership, bound inputs and resource reservations permit.
Finish → diagnose/fix → rerun → advance applies within each dependency chain;
unrelated safe batches may proceed while a failed batch and its dependents are held.
Do not parallelize an intentionally ordered proof or conflicting shared resources,
mutate a frozen validation source, or bypass task-specific sequencing/required gates.

At the first predecessor implementation MR (drafts included), apply
[automatic planning](roles-lifecycle.md#automatic-planning-after-integration)
for the next eligible existing Goal/Spec; integration is fallback discovery only. Provisional planning may overlap CI,
acceptance and cleanup. Implementation may overlap predecessor staging CI after
reviewed source integration and supervisor release under the policy below. Do not reserve the shared test
stack or repeat predecessor validation merely to write the next Plan.

## Reassessment and execution release

After reviewed predecessor source is merged, automatically reassess the
provisional Plan; permit isolated implementation during its affected local qualification using [the release policy](roles-lifecycle.md#automatic-reassessment-and-execution-release).
Material Plan changes require bounded implementer updates and review; unchanged
design requires reconciliation only. Verify every execution gate and record/
relay the supervisor release before work, without another routine owner prompt.
Predecessor final ACCEPT still gates successor integration/final acceptance. A failed
predecessor gate pauses affected work only. The predecessor implementation owner
keeps its staging repairs and attempt ledger, working in parallel with the successor;
do not shift repair responsibility to successor workers. Preserve separate resources. Use existing binding and focused validation; no unchanged full rerun for a SHA.

## Git-first provenance

For new work, Git is the source of truth for committed code and management
artifacts. Bind reviewed inputs by repository, commit and path; record the tested
candidate and fetched target commit in the Result. A commit identifies bytes, not
approval or passing tests. Check for relevant tracked/untracked changes before
claiming a run tested that commit; commit scoped implementation checkpoints before
final qualification rather than building another source-tracking framework.

Record actual commands, outcomes (including failures/interruptions), relevant tool
versions, image IDs/digests and non-secret runtime configuration references in a
compact Result. Link required CI jobs and selected proof artifacts. Git does not
capture external runtime inputs or establish what executed.

Snapshot/hash only relevant inputs not identified by Git: reviewed uncommitted
changes, external configuration/fixtures or mutable artifacts. Preserve necessary
bytes when verification or a dependent task needs them; a hash alone is not
recoverability. Use existing secret-safe private handling for credentials; never
commit secrets. Do not inventory/hash every committed file again, archive whole
checkout copies or create a parallel evidence registry by default.

Retain compact attributable outcomes and necessary proof, then use the existing
closeout policy to purge disposable local bulk output. Do not erase failed-run
history or treat a changed commit SHA alone as reason to repeat unchanged tests.

This default applies to new Goals/Plans. Existing approved snapshots, wrappers,
retention rules and exact input bindings remain required unless explicitly amended
and reviewed; active work is not silently rebound or replanned. Security/authority,
acceptance, rebase, CI and resource-safety gates are unchanged.

## Before every batch

A batch is a named group of local checks (focused tests, builds, lint, controls, full runs). Before launching each
batch on a temporary Goal branch:

1. Fetch the remote target (`origin/<INTEGRATION_BRANCH>`).
2. Assess the exact fetched target delta against tested behavior, check dependencies,
   fixtures/contracts, build/runtime configuration, external inputs and required evidence,
   plus the task's authority, dependencies, approved bindings and governance. If all are
   demonstrably unaffected, record target range/candidate/reason in the existing Result
   and defer rebase/input reconstruction. Otherwise rebase/reconcile before testing.
   Unknown impact follows reconciliation. Preserve dirty work and immutable bindings;
   never rewrite a protected branch or change a running test's checkout.
3. Run `PRE_BATCH_ENV_CHECK` when configured; if it fails, do not launch.
4. Record target and candidate SHAs and the binding provenance with the batch evidence.

Read-only Git, input-integrity and resource-safety inspections may precede the rebase. No redundant rebase when the
target has not moved. Always reconcile before final qualification and again before
MR opening, rerunning affected checks only. No directory-wide management/skills exemption.
Existing approved active contracts retain their required sequence until explicitly amended.

## Reuse existing coverage during planning

### Early risk review and short correction loops

Use the existing Plan and review, rather than another checklist artifact:

- Review changed trust boundaries first: what can a malicious tenant, domain/DNS
  owner or credential holder now control, capture or replay? Check global identity
  separately from credential authority. Include affected migration/legacy-row,
  revocation and downgrade behavior when durable security state changes.
- Order approved implementation to prove the riskiest end-to-end path and its
  principal denial early, before surrounding features. Use existing fixtures and
  a thin functional slice; no planning-time build/probe or extra test campaign.
- For nontrivial changes, distinguish settled design decisions from unproven
  assumptions in the existing Plan. Name the highest-risk assumption, the smallest
  executable proof and its success/failure criteria; reuse existing evidence when
  sufficient. Run that proof early in approved implementation before building the
  surrounding solution. Planning remains inspection-only unless execution was
  separately authorized; routine changes need no artificial spike or extra gate.
- Give executable validation selectors, dependency order and independent parallel
  batches. Estimate implementation effort separately from elapsed build/test time
  and review/queue waits. Cite available timings or mark estimates/unknowns honestly;
  identify required resource slots, serial builds and safe overlap. These are compact
  planning estimates, not guarantees or a new measurement/benchmark campaign.
- Keep execution inside the approved scope. A material addition needs a bounded
  delta stating its design, timing, resource, validation and budget effects, with
  the applicable approval before affected implementation. Defer unrelated extras;
  do not absorb them silently into an active task or reset its attempt ledger.
- If an early proof fails or execution materially exceeds the estimate, checkpoint
  the cause and remaining work in the existing Result. Distinguish coding/rework,
  validation elapsed time and coordination/waiting using available evidence; mark
  unavailable timings unknown. Review a bounded correction or justified split
  rather than repeating the whole Plan or adding more process. Implementation
  need not take less time than planning; repeated surprises require reassessment.
- Inventory affected callers, schema/version/count pins and actual executable
  selectors during planning using bounded searches. Include necessary paths in
  the approved touch-list; discovered design/scope gaps still stop affected work
  for review, rather than granting unrestricted file access.
- Keep a worker on a bounded coherent phase, with the existing context checkpoint
  and compact handoff at safe boundaries. Do not rotate useful workers merely
  per phase, forward entire transcripts or restart attributable tests/budgets.
- Diagnose and rerun the failed selector first. Reuse unchanged passing evidence;
  broaden only for a concrete changed dependency/guarantee under the existing
  supervisor exception rule. No broad run just for a fresh wrapper PASS.
- Batch related findings from the current bounded review into one actionable
  correction handoff with the guarantee, relevant files and minimum evidence.
  Consolidate available independent findings; do not serialize avoidable nudges
  or delay an urgent security hold waiting for unrelated review completion.

Apply these as improvements to existing execution, not a new approval stage,
mandatory phase-agent scheme, monitoring system, separate timing report or
framework. For active work, assess only the remaining delta at a safe boundary;
do not restart completed planning or running validation to satisfy this guidance.

Before proposing tests, inspect the relevant existing suites and fixtures. In the
Plan's existing validation/AC mapping, cite test paths or check IDs and distinguish
**reuse**, **extend**, and **new** coverage. State the actual uncovered behavior or
regression that each new case proves; prefer extending the existing fixture/test
at that boundary. One existing test may cover several acceptance criteria.

Do not clone a suite/fixture, rebuild an accepted invariant or add Goal-named
checks merely to give the new task its own evidence. Similar scenarios at distinct
boundaries (for example policy unit, runtime RLS and browser integration) may be
necessary: explain the distinct failure each catches. The supervisor requests a
bounded correction when proposed duplication has no such justification.

Record this in the normal Plan, not a new inventory/framework or planning-time
test campaign. Relevant regressions and required local/hosted/security coverage
remain mandatory; reuse of earlier passing results still requires reconciliation
of actual changed code, dependencies and runtime inputs. Test reuse is not a waiver
of current validation, manual feedback or evidence requirements.

## Proportionate coverage

Committed test/approval inputs use Git commit/path. Unconsumed status documents
need no duplicate test snapshot; approval/governance changes still require review.
Size local coverage to changed behavior, required security criteria and the actual
path/profile selection. Hosted staging is path-selected, docs-only pipelines may
be omitted, and cancellation is not passing-test evidence. Keep required local
coverage and honest failed/interrupted records; reuse unchanged attributable evidence.
The expensive CI meta-tests are distinct from application tests and may be selected
only for affected CI/harness/control dependencies; dependency and wrapper controls
remain separate obligations. Unknown dependencies select CI meta-tests. Existing
approved snapshots and Goal validation requirements are not silently waived.

## Finish, fix, rerun, advance

Within a batch, let ordinary failures finish collecting results. At the boundary, preserve every outcome and verify
owned cleanup. Gather all outstanding failures, fix them, rerun failed cases/controls and checks affected by the fixes,
then advance. Keep an outstanding-findings checklist in the Result. Unsafe or invalid work (resource or security
failures, corrupted evidence, fail-closed gates) stops immediately; dependent checks that cannot run are recorded as
blocked or skipped, never passing.

## Order of work

implementation → reconcile → focused checks → **required manual tests and operator feedback** → resolve findings
and repeat affected checks → reconcile again → push and open the merge request → hosted fast checks and required
supervisor/Duo review → authorized merge → final qualification at the configured placement. Under the default
`local-after-merge` placement, the coordinator runs affected local batches/suites on the exact merged integration SHA;
it is not a full pre-MR gate. An adopting owner may explicitly select another placement in the repository's
parameters. Final qualification never starts while required feedback is pending. Evidence is reused only for
unchanged relevant content and inputs, with recorded provenance; a changed commit SHA alone does not demand an
unchanged full rerun.

## Integration

- **Safekeeping pushes.** After each implementation or correction round, commit scoped work and push a checkpoint
  to the temporary branch. A checkpoint opens no merge request and authorizes nothing.
- **Reviewed integration.** After the supervisor's pre-integration review, check the target again; if it advanced,
  rebase and rerun only affected checks. Push and open the merge request to `INTEGRATION_BRANCH` with source-branch
  deletion enabled. Before merging, inspect outstanding automated review findings and unresolved threads; passing
  CI does not dispose of them. Follow the adopting owner's recorded integration mode; see
  [staging auto-merge](#owner-selected-staging-auto-merge).
- **Promotion** from `INTEGRATION_BRANCH` to `RELEASE_BRANCH` uses a separate merge request that **retains** the
  source branch. Long-lived branches (`PROTECTED_BRANCHES`) are never deleted or rewritten.
- **Cleanup after merge or MR closure.** After a merge, verify the accepted target content and exact remote source tip; after an MR closes without merge, verify that it is unmerged, no other open MR or live task uses the branch, and unique commits/evidence are integrated or durably preserved in existing records, a tag, retained ref or bundle. Then delete only the exact task-owned disposable source branch. MR closure is not acceptance and does not justify deleting unique unpreserved work. Preserve long-lived/protected branches, immutable tags and failure evidence.
- **Target advances.** When the integration branch advances, the supervisor assesses the delta against every active
  task: unaffected work continues; relevant changes reconcile at a safe boundary; conflicts pause immediately.
  Interrupted tests stop, reap owned children, preserve evidence and release locks before any rebase.

## Options

| Option | Values | Meaning |
|---|---|---|
| `FINAL_VALIDATION_PLACEMENT` | `local-after-merge` (default) | Fast/supervisor/Duo review, then merge; existing coordinator automatically owns affected local qualification of the exact integrated SHA. Full local runs on the integration tip are owner-requested or supervisor-recommended. No automatic post-merge cloud pipeline; staging MR checks and explicitly configured release/diagnostic qualification remain. |
| | `mr-hosted` | Explicit adopting-owner alternative: full coverage hosted before merge. |
| | `local-before-mr` | Legacy explicitly selected alternative; not the current post-merge default. |
| `CANCELED_PIPELINE_SATISFIES` | `false` (default) | Final validation needs `success`. |
| | `true` | An owner-canceled post-merge pipeline also satisfies the gate. It is recorded as canceled, never as passing, and any failed job or known failing check still fails. Pending or running never qualifies. |

## GitLab/glab adapter (`CI_PROVIDER=gitlab`, `MR_TOOL=glab`)

```sh
git fetch origin || exit 1                                       # stop: never rebase against stale refs
git merge-base --is-ancestor origin/INTEGRATION_BRANCH HEAD || git rebase origin/INTEGRATION_BRANCH
git push origin HEAD:refs/heads/TEMP_BRANCH                      # safekeeping checkpoint
glab mr list --source-branch TEMP_BRANCH --target-branch INTEGRATION_BRANCH   # avoid a duplicate MR
glab mr create --source-branch TEMP_BRANCH --target-branch INTEGRATION_BRANCH \
  --title "TITLE" --description "Evidence: see the Goal Result." --remove-source-branch --yes
glab mr view MR_IID --comments                                    # automated review findings before merge
glab mr merge MR_IID --auto-merge --sha REVIEWED_SHA --remove-source-branch --yes # when owner policy authorizes
git fetch origin INTEGRATION_BRANCH                              # freeze the merged SHA for coordinator-owned affected local qualification
glab mr create --source-branch INTEGRATION_BRANCH --target-branch RELEASE_BRANCH --title "Promote"
```

For the promotion request, confirm that source-branch deletion is **off**; a project-wide deletion default also
applies to that target, so check the setting rather than assuming the target changes it.

A published, owned temporary branch that must be rebased is pushed with an exact expected-tip
`--force-with-lease=TEMP_BRANCH:<old-sha>`, never an unconditional force. Hosted quota exhaustion is a resource
stop, reported as such.

## Owner-selected staging auto-merge

An adopting owner may set all non-promotion MRs targeting the integration branch
to auto-merge once required supervisor review, applicable Duo review/findings
disposition, target/input reconciliation, required pre-merge evidence and
manual/owner holds are clear. Record the policy in the repository's live
governance and branch runbook. When enabled, this is standing integration
authority; do not request another routine merge prompt. Early/draft MRs may
record the intent but cannot arm it until every gate clears. A specific owner
hold or opt-out takes precedence. Reassess on any new source commit or blocking
finding. This policy does not apply to integration-to-release/RC promotion MRs,
release tags, final Goal ACCEPT or post-merge qualification. GitLab may merge
immediately when native checks are already green, so verify every repository
review/evidence gate before enabling auto-merge.


## Owner-controlled RC promotion option

When an adopting owner selects release-candidate qualification, use this release flow: owner requests an annotated SemVer RC at the reviewed integration snapshot → coordinator internally opens a temporary tag-pinned source branch/MR to the release branch → that MR runs mandatory complete hosted qualification → owner-authorized promotion (manual after success by default, or explicit owner-selected auto-merge) → verify integrated content and remove only the temporary source branch, or remove it after the MR is closed without merge once the RC tag and all pass/failure evidence are preserved → owner separately decides when to create the final release tag. The temporary branch is provider plumbing, not another owner-managed step; retain long-lived branches, immutable RC tags and qualification evidence. Closing an unqualified/failed promotion MR never promotes or qualifies the RC.

For this option, RC/final tag pushes and release-branch pushes start no automatic pipeline. Admit the promotion MR even for documentation-only candidates and make its full coverage required rather than optional/manual. Verify actual tag object/peeled commit, MR source, complete project/pipeline/job/artifact coverage and prospective/actual integrated content. A branch-name prefix or green partial report is insufficient. Present target-movement checks honestly; do not claim an atomic target guarantee the provider does not support. Never create the final tag automatically from merge success.

The adopting owner may explicitly waive Duo/automated review for this promotion of already integrated content; record the confined waiver and preserve ordinary feature-review, security/manual acceptance and source-proof gates. It grants no general review bypass or project-wide permission change. Keep normal coordinator-owned local-after-merge integration qualification separate. The adopting CI/provider adapter must implement and verify this option before use; these instructions alone do not install tag/MR rules or authorize a concrete release.

## Autonomous mechanical repair

Apply [mechanical-correction authority](roles-lifecycle.md#autonomous-mechanical-corrections):
fix in-scope selectors, fixtures, approved-value catalog/count/version pins and
harness/import bookkeeping directly after diagnosis. Preserve assertions, negative
controls, scope and frozen inputs; record the focused failing/affected evidence in
the existing Result. No approval round, replan or new worker solely for such fixes.
Security/contracts, scope, unapproved production schema/constraints, budget and
native gates still escalate; full hosted retry authority is unchanged.

## Validation economy and recurring failures

Before a batch, choose the smallest checks covering the actual change and required
security/shared boundaries. During diagnosis run the failed scenario and adjacent
cases first; expand only for affected dependencies, unresolved failures or required
final coverage. Reuse valid unchanged evidence with source/input provenance. Do
not create another benchmark, harness proof campaign or duplicate full run merely
for a new SHA or policy edit. Existing path selection and Goal requirements govern.

A passing retry does not establish a root cause or prove a flake fixed. Inspect
failure output first. Under standing integration authority, the supervisor may
approve one retry of the failed job for an evidenced transient/unrelated failure,
within the existing Goal budget; retain both outcomes and record the hypothesis.
No blind repeated retries, whole-pipeline rerun when one job suffices, or treating
an unknown failure as unrelated. A recurrence requires focused diagnosis and a
bounded repair/disposition before another retry. Scope/security/budget and native
permission stops still apply; repeated-issue escalation remains in force.

On GitLab, for non-executable workflow-prose publication only, the supervisor may
use `git push -o ci.skip` after reviewing the entire staged delta and verifying no
application, tests/fixtures, executable scripts/templates, CI selectors, dependency
or runtime inputs changed. Markdown filename or management/skills location alone
is insufficient. Record the exact skipped publication and reason; skipped is not
passing and cannot satisfy or erase a required application staging gate. Unknown
impact follows normal CI. This exception does not change the path selector.

Cancel a superseded duplicate hosted run only after verifying the newer run covers
its exact changes and required suites, the older run has no failed required check,
and no independent qualification depends on it. Record both pipeline IDs and the
cancellation; preserve failed-run evidence and the still-required qualification.


## Preserve local validation evidence before cleanup

Use an explicit private evidence directory for every local qualification or repair run.
Before temporary cleanup, retain the scanned report: check ledger, detailed failing
suite/control logs, supplemental JUnit and relevant diagnostics, bound to the tested
commit and command. Verify those files remain readable after finalization before
retrying, handing back or closing the run. Preserve both failed and later outcomes;
raw logs and generated secrets still follow existing scans and secret cleanup.
If scanning withholds detail, retain only the safe marker/summary and record the
specific evidence gap; never copy unsafe raw files to defeat withholding. Distinguish
JUnit wrapper/aggregate/phase failures from underlying failing test IDs. If a wrapper
cannot retain diagnostics, report and repair that gap before another diagnostic run;
do not solve it with repeated full suites. Regress retention with a failed local run
and prove unsafe or unscanned evidence is still withheld. Purge retained reports only
under the existing terminal-task evidence cleanup policy.

## Complete repair inventory and supervisor verification

Before opening a staging/integration repair MR, reconcile every non-passing
required check across the task's failed runs, including earlier failures not
re-executed later. In the existing Result, record check/scenario, pipeline/job,
failing candidate, correction or evidenced disposition, and attributable passing
proof or remaining hold. Read suite logs as well as summaries; aggregate failures
must point to their underlying failures. Account for skipped/interrupted/canceled
checks without calling them product failures or passes. No entry may be silently
omitted or assumed cleared by a later subset.

Prove affected corrections with the smallest executable selector on the actual
candidate/inputs; a passing underlying test does not clear its failed wrapper or
pin. Derive pin expectations from actual catalog/output and inspect all affected
consumers. Reuse unchanged valid evidence after relevant-delta reconciliation.
If local proof is genuinely unavailable, use the existing supported pre-merge
integration entrypoint and require its success before merging; document why that
broader run is necessary. Do not default to a full suite or use post-merge staging
to discover whether an unproven repair works. Opening the necessary MR to access
its pre-merge job is distinct from merging it.

The supervisor independently compares that inventory with the failed artifacts,
verifies evidence attribution and records repair completeness before pre-integration
ACCEPT. Distinguish fixed/pass, unresolved, not rerun with valid evidence reuse,
derived aggregate and explicitly authorized retry/disposition. An unavailable
artifact, non-execution or timing hypothesis is not a pass. The existing one-job
retry rule still applies: retain both outcomes; recurrence needs focused diagnosis
and a bounded repair, not repeated unchanged runs. A passing retry does not prove
a root cause or repair. Final acceptance and post-integration gates remain.

Use existing Result/review records, one current failure inventory and existing
owners/budgets. No new tracking framework, duplicate full validation, retroactive
pass or permission bypass. Apply equally when the supervisor reviews its own
workflow/CI changes: inspect the actual delta and required self-checks.

## Focused selection before the first run

Focused checks are the default for **every local run**, including initial
implementation, corrections and final local qualification. Start with changed
behavior and required technical guarantees, then choose the smallest existing
package/test filters, service checks, browser projects and integration entrypoints
that cover them. Inspect what the command actually runs; a narrow task name or
"focused batch" label does not make an umbrella command narrow.

Before Plan approval and before launching a batch, the supervisor/coordinator
checks the proposed executable selectors against that scope. Reject unexplained
`all`, generic unit/migration/ops suites or full control suites when they include
unaffected work. Prefer supported filtered/direct existing test commands; record
coverage and required setup in the normal Plan/Result, without a new checklist,
manifest, test wrapper or reporting stage. Reuse valid attributable prior passes.

A broad local run is an exception requiring a concrete technical reason and
supervisor disposition **before launch**: for example a changed shared boundary,
a specified indivisible guarantee, or an identified unresolved dependency impact.
State what wider coverage is needed and why the narrower checks cannot establish
it. Existing reviewed Plan justification can supply that disposition; do not ask
again for an unchanged approved exception. Tool convenience, "to be safe", a fresh
wrapper PASS, or an unavailable narrow harness selector alone is not justification.
If selection is unavailable, report the exact limitation once and propose the
smallest safe existing command for supervisor disposition; do not automatically
fall back to the broad suite or modify the harness to add another framework.

If an overbroad run is already active, retain completed evidence and narrow its
remaining work at a supported safe boundary with owned cleanup. Required schema
setup, real affected regressions and atomic/global guarantees remain; never mark
unexecuted checks passed. Mandatory hosted MR/integration jobs retain their actual
configured scope. This is local selection policy, not permission to skip required
CI, weaken security evidence or change budgets. Apply to active tasks at the next
safe boundary; reassess inherited umbrella selections without a full replan.

### Change-impact selection before merge

Owner direction, 2026-10-18: choose local checks from the consequences of a
change, not only its feature/package directory. In the existing Plan validation
mapping, use bounded searches of the actual diff, callers and CI/harness selectors
to connect each affected guarantee to an existing executable check. Consider only
applicable dependencies: shared authentication/authorization fixtures; historical
migration paths and upgrade/down preconditions; schema/catalog/version/count pins;
startup imports and disabled-feature/copy variants; and concurrent or timing-sensitive
paths. Extend existing coverage only where it leaves a concrete gap.

The coordinator checks that mapping before the first affected batch and refreshes
it for material implementation or integration deltas. Codex reviews coverage of
those effects, not merely whether the selected commands look narrow. Reconcile
against the actual target integration before merge; run newly affected checks or
cite attributable unchanged evidence. A known omitted impact must be resolved
before merge, rather than delegated to staging discovery. Record a real harness
limitation and the smallest justified exception under the existing selection rule.

This is smarter selection within the current Plan/Result, not another inventory,
approval stage, mandatory matrix or broad local suite. Preserve running batches
and assess remaining active work at its next safe boundary; no full replan or
rerun of completed unaffected work. Report exactly what passed locally and what
any separately authorized provider qualification covers. The required combined staging qualification now runs locally on the exact merged revision; it is not inferred from hosted fast checks. Record its actual local environment and coverage.

## Targeted reruns and attributable coverage default

After failures, rerun the smallest failed case/control and checks affected by the
fix. Reuse unrelated passing results with unchanged relevant code, schema,
fixtures, dependencies, configuration and runtime inputs. A suite, batch or job
name is not itself a reason to repeat every case. CI/harness self-tests follow
the same dependency rule; ordinary feature work does not automatically invalidate
the entire control suite. Unknown impact or a real shared dependency requires
broader relevant coverage, with the reason recorded.

Complete local coverage may combine earlier attributable passes and later focused
reruns. Record the original/rerun source and input provenance and outstanding IDs
in the existing Result/review, not a new registry or aggregation framework. Never
rewrite failed/canceled runs as successful, synthesize a full-pass report or use
an older pass to conceal a new substantive failure. Preserve known root-cause
uncertainty; a passing retry does not prove a flake fixed. Required atomic/global
checks still need their actual evidence, and required hosted jobs remain unchanged.

If a rerun is unnecessarily broad, preserve its completed outcomes and use
supported cancellation at a safe case/control boundary with owned cleanup; do not
interrupt unsafe-to-stop work or declare unexecuted cases passed. Required coverage
and unresolved failures remain gates. No full restart solely for a new SHA,
management edit, wrapper tally or clean summary when attributable coverage suffices.

When implementation is frozen, perform bounded read-only source review in
parallel with independent validation. Label preliminary source findings honestly;
final acceptance still waits for the Result, reconciled code/inputs, required
checks and automated-review findings. Review later changes as deltas rather than
repeating the unchanged source review. Preserve the current task, owner and budget.

Apply this default to active tasks at the next safe boundary and future tasks;
it supersedes generic wording to rerun an entire failed batch. If a specific Goal
needs one indivisible run for a technical guarantee, preserve that guarantee and
record the bounded exception. This policy adds no script, test stage or monitor.


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
supported remedy immediately. The authorized operator uses existing owner authority to correct
workflow/configuration where authorized; ask only for genuinely missing authority.
Verify an actual successful tool call. Do not repeatedly dispatch blocked workers,
require owner-run commands by default, or compensate with repeated CI runs.
Keep isolation and real security/permission boundaries; this is prompt diagnosis
and supported configuration, not permission to evade a native refusal.


## Isolated checks and shared resource locks

Where the adopting repository uses a shared harness lock, apply it to shared
mutable resources, not automatically to every container command. Focused unit or
static checks can run without that lock with read-only source and prepared
read-only dependency caches, disposable per-run writable build/temp state, no
network, database/Compose stack, published ports or shared mutable fixtures.
Dependency downloads/shared cache preparation and integration/browser stacks
retain their existing resource locks. Use supported native commands and keep
worker isolation, bounded resource ownership, cleanup and source/input attribution.
This scheduling distinction neither changes test expectations nor waives required
hosted checks, security evidence, task scope or native permissions.


## Owner-selected RC auto-merge — 2026-10-19

The owner may explicitly enable auto-merge for a named RC promotion MR or a clearly recorded scope of RC MRs. Manual promotion remains the default when no such instruction exists. This owner instruction authorizes queuing promotion before CI finishes; once every required full-qualification job succeeds, the provider may merge without another owner prompt. Record the owner instruction, MR, immutable source SHA and selected mode. Before arming, verify the tag/source identity, required complete pipeline, non-squash merge, and conflict-free prospective target tree equal to the RC tree; source-pin the request. Preserve target-movement monitoring and post-merge parent/tree verification without claiming an atomic target lock. Retain and independently verify full project/pipeline/job/artifact proof before claiming qualified release. No failed, skipped, optional or partial coverage counts as qualification. Do not fabricate manual-mode proof or change API snapshots to satisfy a verifier. A manual-only helper's refusal of queued/merged MRs is a mode limitation, not a passing auto-mode proof; use truthful independent verification or a separately reviewed mode-aware adapter. Final release tagging remains a separate owner decision. Feature-MR auto-merge defaults alone never authorize an RC-to-release merge.


## Optional RC Duo review and planned follow-ups — 2026-10-19

RC promotion MRs may use Duo as optional advisory review; Duo is not a required promotion approval gate. Triage every actionable finding and record its disposition in the existing release Result/review. Changes arising from that review must be scheduled and planned as bounded follow-up work: identify the responsible Goal/task or backlog item, intended milestone, scope, acceptance criteria and required focused validation under normal authority. Do not leave actionable recommendations only in an MR comment or silently claim them fixed. Do not change the code or move/recreate an existing RC tag to satisfy a review. If a follow-up changes the release candidate or must ship before promotion, integrate it through the normal staging workflow, then cut and fully qualify a new RC tag under the applicable release authorization. The owner may instead retain a documented prerelease follow-up while the current RC proceeds. Preserve ordinary feature-MR Duo requirements, required full RC CI, security/owner dispositions, selected auto/manual merge mode and the separate final-tag decision. This policy creates no automatic RC2, unbounded repair mandate or new tracker.


## Owner-waived canceled RC qualification — 2026-10-19

The owner may explicitly cancel a named RC promotion MR's pipeline and direct that cancellation to satisfy the current RC promotion gate, treating it as success for the workflow decision only. Record an owner-authorized qualification waiver with the RC/tag/source SHA, MR/pipeline IDs, actual canceled/partial/failed job states, any known findings and the owner's instruction. Never report the canceled pipeline or unexecuted tests as PASS, fabricate green evidence, erase prior failures or claim complete qualification. This exception is specific to the named candidate/pipeline; it is not the default cancellation policy for feature MRs or other RCs. Schedule a follow-up immutable RC tag to run the complete required suite, carrying forward unresolved findings and the qualification debt in the existing release Result/backlog/Plan. The follow-up RC's full suite is required unless the owner separately changes that instruction. The owner may choose this waiver alongside the recorded auto/manual promotion mode; provider auto-merge still depends on the provider's actual cancellation behavior. If existing provider rules prevent merging a canceled pipeline, report that concrete limitation rather than changing project-wide protections, manufacturing a success job or treating a queue as a completed merge. No pipeline is canceled merely because this option is documented, and final release tagging remains separately owner-controlled.


## RC source selection at cut time

For an owner-authorized RC cut, fetch the configured integration branch and tag
its latest remote commit observed at cut time (using the adopting repository’s configured branch). Do not
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


### Private per-run database test scheduling

Focused database tests may omit the shared harness lock when every run owns its
database, private network, credentials and disposable storage; publishes no host
ports; mounts source and prepared dependency caches read-only; and shares no
writable fixtures/cache or existing stack. Use plain supported native commands,
unique resource names/labels, exact run ownership and verified bounded cleanup.
Network access is limited to that run's test services. No alternate executor or
hidden wrapper may evade a native refusal.

Default to at most **two concurrent private database runs per adopted repository**,
across its coordinator and workers. A third needs explicit owner approval when
asked; do not infer higher concurrency. Reuse the existing coordinator/run
inventory for slot ownership; verify live runs before dispatch, release slots
after cleanup and queue on uncertain ownership. No extra scheduler is required.
The adopting owner may explicitly choose a different recorded limit.

This exception applies across tasks and supersedes blanket database-stack locking
only for qualifying private runs. Shared image builds, mutable cache preparation,
Compose/browser stacks, shared ports/data/fixtures and other shared mutations
retain their locks. Unavailable pinned images remain a separate preparation
blocker; do not substitute unpinned images or claim unexecuted tests passed.
Task scope, validation coverage, source binding, budgets and native guards remain.


### Private browser runs and sequential image builds

Private browser runs (including fully private Compose stacks) share the adopted
repository's private database-test concurrency pool: default **two total test
runs**, a third only with explicit owner approval. Count a browser stack and its
internal database once. Each run owns project/names/labels, containers, network,
credentials and writable storage/fixtures; mounts source and prepared caches
read-only; and reuses no live stack. Required host ports must be uniquely reserved,
nonconflicting loopback ports, never shared fixed ports. Record slots/resources,
exact tested source and verified cleanup. These qualifying runs omit the shared
harness lock; shared stacks/ports/data/fixtures retain their existing lock.

Image builds are **sequential, one at a time across the repository's coordinator
and workers**, including individual images in a multi-service build. Reserve one
build slot through the existing coordinator/run inventory, verify current owned
builds, queue on uncertain ownership and release on completion. A private build
can use native plain Docker commands without the harness lock under this exclusive
reservation: task-owned tags, bound context, disposable outputs, no active-tag
replacement or shared writable-cache preparation. Shared cache preparation/other
shared mutations retain their lock. Third-test-run approval does not expand build
concurrency. No extra scheduler, executor, hidden wrapper or native-check bypass.
Reconcile live work at a safe boundary without restart, budget reset or coverage
waiver. This supersedes earlier blanket browser/Compose/build locking only for
qualifying private reserved operations. Adopting owners may explicitly override
limits; preserve the recorded choice and real native permission boundaries.

## Descriptive release artifacts

Keep tag names canonical SemVer. Every new annotated tag message and GitLab
Release entry lists included Goal/task IDs and short titles, material changes
since the preceding candidate/release, and the exact source commit. Final
release notes summarize delivered scope since the previous stable release and
state qualification truthfully; never imply planned or unaccepted work shipped.
Published tags are immutable: improve an existing GitLab Release entry rather
than moving or recreating a tag.
