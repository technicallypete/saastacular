# Closeout and risk reporting

Resource tracking and cleanup run throughout execution. Final ACCEPT is an additional cleanup checkpoint,
not the first one. Closeout removes temporary resources without losing unique work, live ownership or history.

## Direct supervisor publication of acceptance and closeout

The default route for management-only acceptance, post-merge evidence and cleanup
closeout records is supervisor review followed by direct publication to the
configured integration branch under its publication authority. Do not open a new
acceptance, Result-only or cleanup-receipt MR as another routine delivery gate.

The coordinator prepares an append-only Result and cleanup receipt on its existing
task branch/checkpoint or through the existing handback mechanism. Provide exact
commit/path or content binding, attributable evidence, and actual removed/held/
unresolved resource outcomes. The supervisor verifies the submitted bytes and
evidence, requests any bounded corrections through the same handback, and imports
the reviewed Result with history preserved. It records the final disposition only
when required gates are satisfied, updates PROGRESS and related closeout records,
and commits/pushes the completed owned paths from the integration checkout.

This route adds no Result-only MR, automated-review request, CI or local validation
run solely for recording acceptance/cleanup. Documentation checks and review still
apply; the coordinator cannot self-accept or push directly to the integration
branch. Continue authorized cleanup while reporting its true state; an incomplete
receipt is not a completed cleanup or an acceptance waiver. Keep private logs,
credentials, raw secret-bearing evidence and unrelated changes out of publication.

An explicit owner instruction or actual protected-branch requirement may require
an MR for these records; state that exception instead of assuming it. Report native
refusals without bypassing them. Actual source/executable configuration repairs
found during closeout follow the ordinary source-MR review and validation route;
management publication never grants direct application integration authority.
Existing completed closeout MRs and their history remain intact.

## Continuous resource accountability — mandatory

Owner direction, 2026-10-22: cleanup is non-negotiable and tracked throughout
execution, including failed, canceled and interrupted work. A task still being
in flight is not permission to retain abandoned resources. This strengthens the
older acceptance-only trigger for every adopting repository.

- Record each resource in the existing task/run inventory when it is allocated,
  before handing off or using it: task/run and responsible owner, exact container
  and image IDs/tags, named and anonymous volumes, networks, worktree path/branch/
  HEAD, temporary paths and retention reason. Capture mount/volume ownership before
  removing its container. Include nested runs, one-off/probe containers, build
  exports and partially created resources. No untracked scratch creation.
- Keep the inventory current as resources are created, transferred and removed.
  Reconcile actual resources before dispatch/startup/resume, at every worker
  handback and test-batch end, on success/failure/cancellation, and at acceptance.
  Event-driven accounting and bounded checkpoint audits are continuous tracking;
  this does not authorize coordinator polling or another watcher/framework.
- The creator owns cleanup; the coordinator reconciles the task inventory and
  the supervisor verifies its receipt. Stop and reap resource-creating children
  before removal, verify absence afterwards and check for late recreation.
  Cleanup errors are terminal failures, never swallowed by a passing test status.
- Reclaim finished worker worktrees immediately after their handback is preserved;
  do not accumulate them until task acceptance. Preserve unique commits with refs
  or a compact bundle, and dirty/untracked source with a patch/selected files.
  Remove disposable builds, copied dependencies and scratch; a worker's obsolete
  Git lock is not a retention reason. Keep only current coordinator, worker or
  frozen-validation checkouts with recorded live ownership.
- A batch/handback cannot be declared clean, and new resource-consuming work
  cannot start, while its abandoned resources remain unresolved. Existing safe
  source-only work may continue. Track every unresolved item, responsible owner,
  reason and concrete cleanup action in the same Result/receipt; silence or
  unknown ownership never counts as complete coverage. Explicit live/required
  retention is named and time/phase bounded, not an indefinite cleanup exception.
- On crash, reboot or forced termination, reconcile the previous inventory with
  real project resources before resuming resource-consuming work. Traps alone do
  not cover power loss or SIGKILL. Clean abandoned resources first; preserve the
  interrupted status and minimal failure/recovery evidence. Do not present an
  unverified guarantee as successful cleanup.
- Cover containers, unused task images (including untagged IDs), all attached
  named/anonymous volumes, networks, worktrees and disposable temporary files in
  the receipt: allocated, removed, legitimately active/retained and unresolved
  counts, with no unaccounted remainder. Root and worker scratch, external run
  directories and failure paths are included. Keep credentials, native session
  state, committed history and unrelated/shared host resources protected.

Use existing inventories, cleanup/finalizers and compact receipts. Repair leaks
and prove the relevant success/failure/interruption/recovery paths with focused
checks before reopening the affected workload. A documentation update is not
proof that a defective executable cleanup path has been fixed. No host-wide
prune, redundant full suite, new resource database or permission bypass.

## Project-scoped Docker labels and cleanup discovery

Require explicit project attribution for resources created by Compose, direct
Docker commands, nested tests and helper/build containers. Every newly created,
exclusively project-owned object that supports labels carries a stable repository
identity in an adopter-selected, nonreserved namespace. Disposable allocations
also carry run identity and their responsible owner/lifecycle. Keep persistent
development data, active previews and shared resources distinguishable from
throwaway tests; a project match alone never authorizes removing all its objects.

- Set `COMPOSE_PROJECT_NAME` explicitly in the task/command environment. Disposable
  test/preview stacks use a collision-resistant name scoped to project and run;
  a persistent development workspace keeps its documented stable stack name.
  Account for higher-priority `docker compose -p` and reject conflicting identity
  rather than silently selecting a different stack. Keep the same identity for
  creation, discovery, recovery and teardown; avoid a global shell setting that
  redirects concurrent work.
- Compose's automatic project labels group its managed objects. Put common custom
  labels on each applicable service, network and named-volume declaration and on
  project-built images where supported; container labels do not propagate to
  attached volumes/networks. Do not forge the reserved `com.docker.compose.*`
  namespace. Direct `docker run/create`, network/volume creation and image builds
  pass equivalent labels explicitly; `COMPOSE_PROJECT_NAME` does not label them.
- Label genuinely project-built images with stable project identity; add run
  identity only when their lifetime is run-specific. Do not rebuild, copy or
  relabel shared/pulled images merely for bookkeeping, nor label shared images or
  external volumes as exclusively owned. Record their actual IDs and shared-use
  disposition. Anonymous volumes and other objects not labelable at creation
  require positive exact-ID/parent attribution captured before the parent goes.
- Use label queries for ordinary discovery and reconcile them with the existing
  run inventory/registry and creation receipts. Keep only the existing compact
  exact-ID, ownership, retention and recovery records needed to verify cleanup;
  do not add a second manifest/database/framework or hash the whole host. Labels
  simplify finding resources but do not replace crash/reboot recovery state,
  cleanup outcomes, anonymous-volume bindings or worktree/tmp accounting.
- Never infer removal authority from a label/name alone. Confirm responsible owner,
  lifecycle, exact identity and current use before removal; stop creators first
  and verify absence. Missing/conflicting ownership or a failed query is a hold.
  Preserve legacy readers until their reviewed retirement condition is satisfied.

Apply the rule to new/adapted creator paths. Existing bound tasks retain their
reviewed label/proof contracts until a safe, explicitly reviewed adoption; do not
recreate active resources or restart tests merely to relabel them. Label changes
need focused discovery, foreign/live/persistent protection, failure/interruption
and cleanup proofs through the existing workflow. Documentation alone is not an
implementation or cleanup guarantee.

## Verify before removing anything

Before removing a worktree, branch or in-flight marker, or closing a disposable merge request, verify:
- **Integrated content:** the integration branch contains the accepted changes (ancestry, or content and
  provenance reconciliation for rebased or squashed work).
- **Unique work:** no unmerged commits, dirty files or unpreserved evidence. Dirty work is reconciled or preserved
  outside the checkout first; force removal is allowed only under the mandatory cleanup policy below.
- **Live ownership:** no live session, process, validation run or required resource uses it. A live coordinator's
  directory is **never** deleted.

Then use `git worktree remove` for verified obsolete checkouts and `git worktree prune` for stale registrations.
Local branch deletion is separate; keep unique commits unless reconciled or archived. Close each disposable draft or
qualification merge request once nothing depends on it, keeping its discussion and pipeline links; a delivery
request that still carries unintegrated work stays open. Record removed paths, closed requests and retained
exceptions with reasons in the closeout record.

Clear an in-flight marker or lock only after verifying its owner ended; never clear another live run's marker.
Append a superseding status; never rewrite historical facts or turn withheld or failed evidence into a pass.

## Mandatory cleanup at every task acceptance

Owner direction, 2026-10-17: every final task ACCEPT triggers immediate cleanup
of all finished task-owned workers/worktrees, temporary branches, containers,
images, volumes and disposable local evidence. Do not wait for an epic closeout
or another owner prompt. Plan approval or conditional integration is not final
ACCEPT. Acceptance history remains; closeout is incomplete until cleanup has a
verified receipt or an explicit live-use, retention or native-permission hold.

- Verify the task's accepted integration and that each worker/process has ended.
  An agent-tool Git worktree lock on a finished worker is not a retention reason:
  use `git worktree unlock`, then remove the checkout. This standing authority
  includes `git worktree remove --force` after preservation below; do not request
  another approval merely because the finished checkout is locked or dirty.
- Before removing a dirty checkout, save a binary-capable `git diff --binary HEAD`
  patch outside it, plus any unique untracked files (a Git diff does not include
  them). Preserve necessary ignored files separately; do not copy credentials or
  reproducible build bulk into a patch archive. Record source HEAD, paths and the
  verified preservation location in the existing receipt. Preserve local-only
  commits with an exact retained branch/ref or durable Git bundle before removing
  their checkout; never delete the sole reference to unique work. Once reconciled
  or integrated, remove the temporary preservation copy/ref too.
- Remove all verified task-owned containers and volumes and unused task-built
  images, including attributable untagged images by exact ID. Remove temporary
  source branches after their MR merges or closes without merge, once unique
  work is integrated or durably preserved and no open MR/live task uses the ref.
  A closed unmerged MR is not acceptance and is not permission to discard unique
  work. Preserve immutable tags and failure evidence. Protect configured
  long-lived branches. Do not use host-wide prune, tag/age alone as ownership,
  or remove another task's/shared base images, volumes or builder cache.
- Audit remote branches explicitly at acceptance/closeout. GitLab's source-branch
  deletion setting is only a convenience and does not prove all task refs were
  removed. Compare `git ls-remote --heads origin` with the task/MR inventory;
  confirm whether the MR merged or closed unmerged, ownership, live use and
  unique-content integration or durable preservation before deleting each
  disposable task branch. Preserve unresolved
  work in an existing Result/retained ref with a reason and release condition.
  Audit local refs separately (`git branch -vv`, `git worktree list`): remote
  deletion does not remove local branches. Refresh stale local `main` only when
  it has no unique commits. Record removed/preserved refs in the same receipt.
- Purge task-local logs, screenshots, traces, downloads and build scratch after
  retaining the compact Result/review and required hosted artifact links. Almost
  nothing should remain locally by default. A preservation patch or selected
  proof needed for unresolved unique work has a named owner and release condition;
  it is not an indefinite raw-evidence archive.
- Keep the live coordinator, active successor/preview checkouts and resources,
  required private session/observer state and committed management history.
  Transfer any real dependent use explicitly; an obsolete worker lock alone is
  never a live-use hold. Native permission refusals remain real holds, reported
  once for supported owner handling; unlocking a stale Git worktree lock is not
  permission to bypass a native access denial or a live harness/process lock.

Use the existing inventory and one bounded cleanup pass. Record removed counts,
measured reclaimed bytes (or unknown), preserved unique work and exact holds with
owner/release condition. The supervisor verifies the receipt and updates progress;
do not silently skip worktrees, images or evidence. No new framework, test run,
full replan or budget reset is required. This supersedes older blanket prohibitions
on force-removing preserved dirty checkouts or unlocking finished worker worktrees.

## Session records: archive before prune

Delivery records are private, machine-local and gitignored. After the supervisor or owner verifies a task's standing
terminal disposition from committed records, the companion helper's `compact` archives that task's whole delivery
groups into the private archive with a manifest, verifies it and only then prunes them from the live log (dry run
first; apply bound to the reviewed selection, from the main checkout, in a quiet window). Active, ambiguous, failed,
unconfirmed and incomplete records keep their bytes. Before removing a linked worktree, copy any worktree-local
delivery log byte for byte into the same archive and record both hashes. See
[claude-session-handoff](../../claude-session-handoff/SKILL.md) for exact commands and refusals.

## Bounded routine cleanup

- Record task/job ownership when creating temporary paths, containers, images and
  volumes, using the existing run inventory. Capture exact IDs and mounted volume
  names while the creating container still exists. Use the existing interruption
  cleanup; names, age and current non-use alone do not establish ownership.
- The coordinator performs routine cleanup; the supervisor checks its compact
  receipt. Reuse the existing cleanup mechanism, retained decisions and inventories.
  Use one scoped preflight and one removal pass, with immediate checks for mutable
  targets. Do not create another Goal, Plan, registry or verification framework, or
  require new per-file hashes for disposable bulk output. Existing approved hash,
  retention and recovery requirements remain binding.
- Check task-relevant processes and mounts. Unrelated host services or container
  churn are not blockers by themselves. Hold an uncertain or live item, explain
  why, and clean the independently verified remainder. Required active-run resource
  cleanup gates still apply; this does not permit starting another batch with
  unresolved required resources.
- After interruption, preserve the partial receipt and continue unfinished items;
  verify completed items rather than redo them. Do not restart unrelated tests or
  repeatedly rebuild inventories because one item is held.
- On a native permission refusal, report the exact action and reason once and
  request supported owner handling. No repeated approval relays, rephrased retries,
  settings changes or executor substitution to bypass a denial. Include genuine
  human mid-turn decisions when checking authorization; never invent one.

Closeout records need the task, ownership source, removed/retained counts and bytes,
retained proof locations, and held items with reasons. Partial cleanup is reported
as partial. Unverified volumes remain held; a targeted exception does not stall
all verified cleanup. This is a bounded operational step, not another test campaign.

## Local validation evidence: retain a compact record, purge bulk output

Default closeout policy: after final ACCEPT, integration and required qualification,
remove bulky task-owned local validation artifacts; do not retain every harness run
indefinitely. Apply this to already-closed tasks at the next cleanup as well. A
running, blocked, review-pending or conditionally integrated task is not eligible.
A canceled/superseded task needs a verified terminal disposition and confirmation
that no investigation, appeal, dependent task or explicit retention hold needs its
artifacts. Do not infer eligibility from an idle process or a deleted branch.

Before deletion, the coordinator inventories owned evidence and the supervisor
verifies a compact retained record linked from Result/closeout. Keep acceptance
and review history, exact source/input/command/tool/configuration hashes, per-check
outcomes, failed/skipped/interrupted history and fixes, required manual feedback,
CI job/pipeline references and necessary selected proof artifacts. Retain input
snapshots/oracles or other bytes still required to verify a decision or resume a
dependent task. A hash identifies deleted content; it does not keep that content
available. Record explicit retention holds, their owner, reason and release gate.

Then purge redundant raw logs, traces/videos, report copies, extracted/build scratch,
temporary fixtures and bulky per-run outputs covered by that record. Compress
selected retained proof where useful; compressing all runs is not a substitute
for purging redundant output. Recheck ownership, resolved containment, symlinks
and live users before removal; use exact inventoried paths, never broad glob/home/
state-directory deletion. Shared or unverified content stays put with a reason.
Do not purge credentials, native session transcripts, handoff logs/archives, active
observer state or committed Goal/Spec/Plan/Result/review/PROGRESS history under this
rule. Delivery-record archive-before-prune remains a separate unchanged procedure.

Record paths/categories, retained locations and hashes, deletion time, file count
and bytes before/after/reclaimed, plus holds. Annotate old local links as purged
with the retained record; never rewrite a failed run as passing or claim deleted
artifacts can be downloaded/reused. Local cleanup does not change hosted artifact
expiry, CI gates, Docker ownership or shared builder-cache pruning authority.
If retention, safety or summary completeness cannot be verified, report the hold
instead of deleting. No further owner prompt is needed for routine verified
closeout under this default; an explicit owner retention instruction takes priority.


### Local evidence is temporary — owner direction, 2026-10-14

Default durable evidence locations are committed Result/review records and the
existing CI pipeline artifacts, not a permanent local evidence archive. Hosted
runs already publish artifacts; do not keep duplicate downloads after review.
During active work keep only what is needed to diagnose, review or resume it.
At verified closeout purge raw logs/reports, traces/screenshots, copied tools/SDKs,
fixtures and caches; retain compact commands, outcomes (including failures),
source references and pipeline/job links in Git. Local-only/manual evidence may
be published as sanitized artifacts through an existing authorized publication
path when suitable; do not create another pipeline or upload secrets just to
archive it. No new artifact service or CI redesign is required.

Keep local proof only for a specific outstanding verification/dependency or
required manual record unavailable elsewhere. Record why, owner and release
condition; closure releases routine scratch holds, not an explicit unresolved
investigation. Hosted artifacts may expire: preserve a necessary selected proof
before expiry or acknowledge that the raw bytes are unavailable. A hash or CI
link does not guarantee indefinite retention. Expiry alone is not a reason to
mirror every artifact locally. Purging raw files preserves their documented
failures and acceptance history. Private live bindings, code approvals, session
transcripts/handoff archives and observer state are separate operational state,
not disposable validation evidence. Apply bounded ownership/live-use checks;
no per-file hashing campaign or new cleanup system is required.

## Progress compaction

After final ACCEPT and integration closeout, the supervisor assesses PROGRESS.
Every epic closeout is a mandatory periodic compaction checkpoint: compact
accumulated superseded entries, or record that none need moving. If active-task
bindings prevent a safe change, record the specific hold and next safe checkpoint.
Record this disposition before marking epic closeout complete; no new validation
campaign or tracking framework is needed.

During compaction:
- keep one current entry per Goal with its disposition, acceptance and review links, dependencies, blockers and
  next gates;
- move superseded updates into a committed dated archive, `management/<release>/archive/PROGRESS-YYYY-MM-DD.md`
  (header template `templates/management/release/archive/PROGRESS-ARCHIVE.md.tmpl`), preserving original text,
  dates, decisions, failures and evidence references, and link it from PROGRESS;
- check moved links and anchors, adjusting relative references so they still resolve;
- if PROGRESS is a bound input of an active task, defer or notify that task's coordinator to reconcile before its
  next batch. Compaction alone never justifies another full test run.

## Risks and performance

Record risks in `docs/RISKS.md` (template `templates/docs/RISKS.md.tmpl`). Surface query, index, locking,
contention and other performance concerns there with the observed evidence, affected component and proposed
follow-up. Recording a concern does not impose a vendor-specific architecture and never authorizes silent tuning of
application code; changes go through a Goal.


## Routine closeout receipt

At the existing verified closeout, perform the authorized ownership-scoped cleanup
without another routine proceed prompt. Retain unique work and specific live or
uncertain holds; do not turn one hold into a stop for all verified cleanup. No host-
wide volume/image prune or shared builder-cache deletion is implied.

Use the existing Result/closeout receipt, not a new resource database. Summarize:
`removed: worktrees/branches/containers/images/volumes/evidence; reclaimed: <measured
bytes or unknown>; holds: <item, reason, release condition>; builder cache: <observed
size/reclaimable or unavailable>`. A read-only builder-cache observation is enough;
no cache benchmark or additional qualification run. Distinguish logical deleted
bytes from measured filesystem space and shared cache from task-owned resources.
Resume interrupted cleanup from its receipt. The supervisor verifies it, updates
progress and assesses required compaction as part of the same closeout transition.


## Container artifact ownership

When containers produce host-retained reports, arrange operator-removable output
at creation. Preserve the browser/tool's reviewed non-root user; do not assume
its UID matches the host operator. Reuse existing stdout/archive transfer or
container copy-out with host ownership instead of report bind mounts that create
foreign-owned trees. Capture owned container IDs/mounts before removal; transfer
and cleanup failures remain failures, including when the test itself passes.

A retained one-off must have positive run ownership before copying or removal;
use an immutable ID, protect start-inventory/foreign resources and cover aborts
with existing project cleanup. Do not add a new resource framework, switch the
whole browser to root, weaken assertions or rely on recurring sudo/chown/broad
permission grants to make routine closeout work. Existing explicit owner/native
handling still governs historical remnants; no privilege workaround is implied.


## Private files for new tasks — 2026-10-17

Use the main integration checkout's ignored `.local/{evidence,secrets,tmp}/` for
new task-private files, shared across worker worktrees. Follow the [local storage
policy](adoption.md#repository-local-private-files): task-scoped paths, private permissions, Git/build/upload exclusions
and separate evidence/secret cleanup. Existing tasks keep current paths through
acceptance cleanup; do not migrate active bindings or native session state.
