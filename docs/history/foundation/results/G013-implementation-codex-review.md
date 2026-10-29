> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G013 implementation review — 2026-10-15 UTC

**CORRECTION LOOP.** Reviewed source `8607534e25ac07e97524713d864805911fba60ac`
on staging `aae1101`, Result SHA-256
`65914e55593eec198570d834cce31386dc3786f3b36a45d444c825e975785459`.
Two bounded read-only reviews examined recovery/retention source and AC evidence.
All 47 changed paths match the approved seventeen-pattern scope. Existing
migration/grant model and accepted F4 residual are not reopened.

## Required corrections

### C1 — Replay must cover the current ledger (AC4)

recovery.sh writes a restore-bound replay receipt without the replayed ledger
head/sequence (around 808ff). Reopen verifies the current attestation, then accepts
any same-restore complete/gap replay (around 890–900). A new restriction appended
and attested after replay can therefore pass reopen without being applied.

Bind a successful replay receipt to the exact validated ledger identity, head
and through-sequence. Refuse ledger drift during replay; reopen must require
agreement among the current validated/attested head and that replay receipt,
in addition to the existing backup/restore/target binding. A different head
requires replay, not a new attestation alone. Keep existing explicit gap/risk
semantics and truthful failure states; no new ledger framework.

Minimum regression: replay through N; append/attest a valid N+1 restriction;
reopen refuses and target stays closed. Replay the new head; verify restriction
applied and normal gate succeeds. Cover a ledger change during replay with the
existing fixture/seam. Reuse the current recovery drill.

### C2 — Failed receipt writes cannot bypass fence cleanup (AC1/AC7)

The EXIT guard inherits errexit. Its unguarded INCOMPLETE write at recovery.sh:55
can fail before fence_release at 57, despite the guard promising failed-record-
write cleanup. Make cleanup steps independent of receipt/log write success,
preserve a nonzero outcome and attempt release of exactly the fence established
by that command. Report release failure truthfully; do not claim cleanup succeeded
without verification. No guarantee against SIGKILL/host failure is introduced.

Minimum regression: force marker write failure in the disposable existing fence
fixture; show release is still attempted, actual fence state is checked and the
command remains nonzero. Ensure test cleanup handles its own failure safely.

### C3 — Correct residual-risk disposition wording (AC8)

SECURITY.md:166–167 and RECOVERY.md:200 still call F4 pending. Reference the owner
acceptance recorded at 31444dd, retaining the risk, scope and operational safeguards.
No new owner decision, grant change or test campaign for this wording correction.
The immutable restore-identity-line digest deviation is acceptable: fields appended
by staged object completion need not invalidate the stable restore binding.

## Failed-run and cleanup dispositions

- Verified b6-02 reports 104 PASS/3 FAIL/7 INFO; all four G013 IDs pass. The run
  remains failed, never relabelled green.
- Codex's own contract-literal defect is corrected on staging `dcb68a4`;
  neutral wording correction (not included). Reconcile
  it and collect the covering B-LITERALS pass; no scanner exception.
- Actual b7-02 output proves VS2/VS3 PASS, exit 0, on `537d124`/staging `aae1101`.
  Only the Result differs between that candidate and this reviewed head. The
  other 439 controls passed in b6-02. Explicitly bind/reuse that unchanged covering
  evidence in the Result; it does not establish the load-sensitive control defect
  is fixed and does not turn the old aggregate/tally into a successful run.
- Preserve the unproven beginify-core-check image as an explicit ownership hold.
  Record exact before/after IDs and cleanup exit 18 honestly. No deletion from its
  name/label, invented capture row, cleanup-tool weakening or global prune. The
  hold is not an assertion of clean complete removal; owned resources and final
  closeout still require their scoped checks.

## Authorized correction and validation scope

Continue as **attempt 2 of B7** (one completed; checkpoint after 4), without budget
reset. Fix only C1–C3 in the existing allowed files. Fetch/rebase/reconcile before
the next batch, pin changed protected documentation bytes correctly, and preserve
failed/provisional history. Reuse unchanged passing coverage with explicit source
and validation-input provenance.

Rerun existing B-RECOVERY-DRILL and the covering maintenance drill (shared cleanup
code), B-SHELL, B-LITERALS and applicable scope/protected/privacy checks. Add only
the focused cases above. No unchanged full local suite or all 441 CI controls
merely for these fixes; actual changed inputs/failures govern any broader need.
Update Result last, push for safekeeping and return exact head/hash for re-review.
No MR/integration acceptance yet. Required MR/Duo/staging disposition and final
Codex ACCEPT remain; G014 execution is not released. Notify the existing
coordinator automatically; no new implementer or duplicate review campaign.


## Attempt 2 correction re-review — 2026-10-15 UTC

**ACCEPT** — source corrections and attributable covering evidence accepted at
`63d50722166dcbbb0493dbb686f91b03e4a56f71`, Result SHA-256
`7a1e0d4f98de794f102801b1b8d6a56332260d4e9045d6b433e7fc043b9ad070`.
A bounded independent read-only reviewer found no substantive remaining C1–C3
blocker. Codex inspected the actual correction diff and covering evidence:

- C1: replay captures the validated ledger UUID/head/through-sequence, applies
  only that range and fails incomplete on head drift. Reopen requires the current
  attested head to match the replay receipt. RD-O proves post-replay append and
  attestation refuse reopen, mid-replay drift refuses, and a fresh replay removes
  the membership before the normal gate succeeds.
- C2: the EXIT guard disables errexit for independent cleanup, attempts release
  despite the failed marker, reports the recorded verified release state and
  preserves nonzero exit. The explicit capture-failure path releases first.
  RD-B-NEG verifies marker-write failure, no COMPLETE, released fences and restored
  ACLs; no SIGKILL/host-failure guarantee is added.
- C3: both runbooks now cite the owner's accepted F4 disposition without changing
  grants or removing its safeguards.

Verified b8-01: candidate `5d05d3f928017d44526a01092c0776be8358e5b8`, target
`4f195c58c9e2aa1b8b7a9fa71d8c8f7d04426cd1`, 53 PASS/0 FAIL, run and finalize
exit 0. Recovery/maintenance drills, shell/literal/scope/protected/privacy checks
pass. results.tsv SHA-256:
`e641dbe43bad05e779774382e171f837c077fda45ef73a1fac1a4ddca71b177b`.
The four changed recovery/harness/security/runbook files are byte-identical to
this tested candidate. Verified b8-03: candidate
`5a6f995d0cb046602f5c9c3fc7784b6ac420b127`, target `9800db4`, 22 PASS/0 FAIL,
run and finalize exit 0; results.tsv SHA-256:
`324a3f29cf22a25d152b4bc4bdd2f0a338af68366b7347a9a2ea53c3ff752640`.
Only the Result differs between b8-03's candidate and the submitted head.
The failed b8-02 prerequisite-only run remains recorded. Unchanged earlier
coverage and focused VS2/VS3 evidence are reused under the prior disposition;
none of the historic failed aggregate runs is relabelled passing.

The exact submitted Result is imported into staging. Attempt usage remains 2/B7,
checkpoint after 4. No new broad local run is requested merely for this review.
Rebase/reconcile onto current remote staging before opening the feature MR;
management-only review/probe records do not themselves require repeating unchanged
application tests. Verify exact content/input provenance and rerun only relevant
changes. The accepted source may proceed through MR/fast/Duo review and authorized
integration under the existing workflow; remind the owner of any Duo findings
before merge. No new architecture or scope authority is granted.

This is implementation-review acceptance for integration, **not final Goal
ACCEPT**. Applicable post-merge staging disposition, final Result/cleanup and
closeout remain required. Preserve the unproven image ownership hold, do not
prune unrelated resources, and retain the historical failures. G014 implementation
remains held on final G013 acceptance and its own provider/Plan/ADR/input gates.


## MR !29 Duo disposition — 2026-10-15 UTC

**CORRECTION LOOP.** Reviewed MR head
`4ea4cfb7a37fa5279abb7353ca911e695872ba4c`; MR fast pipeline `2952382395`
passed. Codex and a bounded independent read-only reviewer verified Duo's two
findings. The source matches the accepted attempt-2 source; this reopens only
these findings, not settled architecture or the F4 residual.

### D1 — preserve an externally owned fence on capture interruption

Blocking AC1/AC7 and Plan §3.1/3.2. `cmd_backup` installs `fence_abort` for
TERM/INT/HUP even under `--use-fence`, while `fence_abort` releases FENCE_FILE
unconditionally. This can reopen an operator-established fence that the backup
promises to leave held. Preserve the ownership/guard mode before clearing it;
signal cleanup may release only the fence established by that command. Keep
nonzero outcome and truthful incomplete-marker handling. Do not change explicit
`unfence` authority or the normal integrated-capture cleanup.

Minimum validation: establish an external fence, interrupt its `--use-fence`
capture, prove nonzero/no COMPLETE and that the actual fence, database ACL and
provider restrictions remain held. Then explicit `unfence` must restore the prior
state. Keep the existing integrated-capture interruption/release positive.
Use the existing focused recovery drill, plus B-SHELL and relevant
scope/protected/privacy checks; no unchanged full local suite or all CI controls
solely for this correction.

### D2 — document bounded purge head-of-line blocking

The oldest candidate batch can be fully skipped (`not_preserved`, `bytes_present`
or `provider_unavailable`), yielding zero deletions and ending the invocation
before younger eligible tombstones. The current implementation correctly reports
remaining work as partial/nonzero and preserves unsafe-to-purge rows. Spec R5/AC6
and Plan §6.5 do not require starvation-free traversal; no keyset paging redesign
is requested or authorized here.

Document this limit explicitly in RECOVERY.md: unchanged retries can encounter
the same blocked batch. Resolve the actual blockers through the existing trusted
ledger preservation/synchronization, authorized deletion or provider recovery,
then retry. Do not delete blockers to force progress, erase replay evidence, or
claim all eligible rows were visited. Existing skipped/partial maintenance
coverage is reusable; this disclosure needs no additional broad test campaign.

### Release and history

Authorize one bounded correction round as **attempt 3 of B7**, checkpoint after 4;
no counter reset. Fetch/rebase/reconcile before its focused batch. Update Result
last with failed/passing evidence and exact provenance, push for safekeeping,
reply to Duo and return for Codex re-review before resolving the findings/merging.
The earlier source approval is superseded only for these two outstanding items.
Notify the owner before merge of these Duo findings and their final disposition.
MR/staging/final closeout gates remain; fast passing alone does not release merge
or G014 implementation. The live G014 owner probe continues independently.


## Attempt 3 / MR !29 re-review — 2026-10-15 UTC

**ACCEPT — corrections for integration.** Reviewed source
`7bd8ced703da9590d6b94e854f12feee770dccab`, Result SHA-256
`e85a378bfcbf918546d5706c4921a8360feaaffd5fcf1726ce1fe0096c4a8aea`,
imported byte-identically. Codex and the same bounded read-only reviewer found
no remaining D1/D2 blocker. Earlier C1–C3/F4 dispositions remain settled.

- D1: signal cleanup saves ownership before clearing the guard, excludes external
  fences, preserves nonzero status and independent cleanup. RD-B-INT proves
  actual ACL/runtime-login/provider restrictions remain held after interrupted
  external capture, then explicit unfence restores prior state. Existing integrated
  interruption and failed-marker cases still pass. Signal handling ends after
  capture completion in both modes.
- D2: the runbook explicitly describes oldest-batch blocking, truthful partial/
  nonzero output, lack of progress on unchanged retries and authorized repair
  before retry. The documented limitation is accepted; no paging redesign.

Verified a3-02 results.tsv SHA-256
`a8a28528310841bc2ed7f8ff0869881b6bcb3ef558b661f960d95cad5d4d0d7b`:
53 PASS, zero FAIL/UNMET, five INFO; run/finalize exit 0. Verified a3-03 SHA-256
`630a597ed5b80c33a65459afc91c38da7be731315348c52e4a185d683f56d9c3`:
22 PASS, zero FAIL/UNMET, five INFO; run/finalize exit 0. The Result's table says
zero INFO for these rows; that bookkeeping typo does not alter the verified
required checks. Correct it in the normal closeout update; no separate test or
correction round is needed. Failed prebatch a3-01 remains recorded as blocked
before tests, not a passing batch.

Recovery/harness/runbook bytes match tested `399d9d4`; their identity survives
rebase onto `fd21bdc`. Only the Result differs from reconciled candidate `7415180`.
Merged SECURITY.md SHA-256 is
`66e80909c686de57700c754443bbf12e287fb6850548e18130e9d9d7f4b04229`,
preserving G013's accepted F4 wording and staging's G014 planned-boundary note.
Accepted unchanged-evidence reuse and the unproven image ownership hold remain;
no historical failed broad run is relabelled passing.

Usage remains 3/B7, checkpoint after 4. Resolve the two reviewed Duo discussions
and proceed through the existing conditional integration workflow: reconcile the
latest target, retain exact content/input provenance, pass MR fast checks and
review/dispose of any new Duo findings before merge. Under standing owner
integration authority, enable auto-merge/merge once those gates close, with source
branch deletion enabled. No repeated unchanged broad local run is requested for
management-only deltas. A relevant content change returns for affected review.

This is not final Goal ACCEPT. Required staging disposition, updated Result and
scoped cleanup/closeout remain. G014 may continue planning but not implementation
until final G013 acceptance and its own Plan/provider/input/resource gates close.

## Failed staging cleanup and attempt-4 release — 2026-10-15 UTC

**CORRECTION LOOP — bounded resource lifecycle fix; attempt 4 of B7 authorized.**

Verified pipeline **2952592867**, gate job **17214158089**, exact merged source
`8b0823a17a675dcc369d722c95c5539855f68e81`. Hosted results contain 77 required
PASS rows, zero FAIL/UNMET, and 441/441 CI controls PASS. Finalization nevertheless
exited **14**: `the harness left 3 resources behind`. The staging gate failed;
G013 is not finally ACCEPTED and G014 implementation remains held.

The artifact's `inventory-leftover.txt` contains exactly three anonymous volume
IDs plus the separately allowed `project-archetype-render-cache`. The latter is
not a fourth new leak. `cleanup-inventory.tsv` lists four removed image rows;
`owned_remaining=0` means no registered resources remained, not that no unregistered
volume leaked. Do not weaken either inventory or acceptance checking.

### Finding and minimum correction

Codex and an independent read-only reviewer identified a high-confidence source
match: G013's detached PostgreSQL-image psql clients `adminw`, `rtw` and `lock`
are removed without their inherited anonymous data volumes at run.sh lines
3199, 3237 and 3526; `rd_purge` repeats volume-less removal. `docker_run_owned`
records their container IDs, but does not retain their mounted volume IDs before
these containers disappear. Exact mapping of each hosted anonymous ID to a removed
container was not retained, so this attribution remains strongly inferred until
the focused reproduction, not a claim of hosted mount-level proof.

Claude may correct the G013 drill's exact owned client lifecycles and normal/abort
cleanup in `infra/validation/run.sh`, within the existing scope. Capture actual
owned mount IDs before container removal using existing ownership/start-inventory
rules, and remove the anonymous volumes with their owned containers. Reuse current
registry/cleanup primitives; no generic registry redesign, global volume sweep,
allowlisted anonymous leak or change to operator fences/data/recovery semantics.
Keep unrelated, shared and pre-existing resources protected. No reversion of the
working recovery feature is requested.

### Validation, budget and integration

Reproduce the three detached client lifecycles in the existing fixture, including
normal and abort cleanup; record actual mounts and verify owned anonymous volumes
disappear while a pre-existing control volume remains. Use relevant shell/static
checks and affected recovery/maintenance evidence with explicit unchanged-evidence
provenance. No new validation framework or unchanged full local campaign merely
for a management/SHA change; applicable changed-behavior and hosted path-selected
gates still apply. Do not claim the original failed pipeline passed.

Use the existing verified G013 task owner/worktree and Git-first reconciliation;
Fable high is the current test-correction default where available. No permission
mode change or duplicate implementation worker. This is one shared Goal-level
attempt 4, with inner fix/retest batches inside it; no budget reset or extension.
Return source diff, focused proof and appended Result for review and the B7
checkpoint after attempt 4, before a further meaningful attempt. Only then follow
the existing authorized MR/fast/Duo/integration route and collect the required
staging disposition. Do not retry the unchanged failed gate as the remedy.

G014's owner-authorized provider-only QP-B is independent and runs no Docker/core
database; preserve that live probe. Final acceptance/cleanup and G021 remain gates.


## Attempt-4 cleanup review and B7 checkpoint — 2026-10-15 UTC

**ACCEPT — bounded correction for integration; B7 checkpoint satisfied. Not final Goal ACCEPT.**

Reviewed exact remote submission `1cacf95ce2828b0d776488a6718dcb9fc3287b18`
on `g013-drill-volume-cleanup` (base `515f866`), the actual two-file diff and
independent read-only review. Imported Result SHA-256
`6c6c69254b9a29db2fc78f51a58832e6e540582fba35a08c686cb10a1d28a11b`.
Submitted `infra/validation/run.sh` SHA-256
`ad26ead9a9ff1f01b13c3a11b1fa6da5c4b99d436640fc7f7821b81663d17fb0`
is identical to tested `8668e5b`.

The closed drill callers create fresh PostgreSQL-image anonymous mounts and
supply no arbitrary volume arguments. The helper captures mounted IDs before
removal, records new anonymous mounts through existing inventory controls, and
removes/verifies them on normal and abort paths. Named controls are preserved.
This accepts those specific callers, not a general-purpose arbitrary-container
cleanup guarantee. No application/recovery/fence semantics or gate were weakened.

Independently inspected private a4-00 reproduction: former leak reproduced,
10/10 normal/abort/idempotence/control checks PASS. Inspected a4-01 report:
53 PASS, 5 INFO, zero FAIL/UNMET; 29 recovery steps PASS; leftover inventory
is empty (zero bytes); cleanup reports removed=0/refused=0/unproven=0/
remaining=0/failed=0. Independent reviewer confirmed the unchanged 71-to-71
anonymous-volume set and surviving pre-existing named control. Existing
functional/control passes retain their explicit unchanged-content provenance;
original failed staging finalization remains failed. No redundant local suite
is requested for management-only reconciliation.

Four of seven Goal-level attempts used; checkpoint review is satisfied and three
remain within existing scope and escalation rules, with no budget extension.
Fetch/rebase onto current origin/staging, reconcile Result/governance bindings,
verify tested executable bytes unchanged, then open the correction MR to staging.
Standing feature-integration authority applies after required fast/Duo review;
notify Codex of findings before merge. Retain source deletion and auto-merge intent.
The existing G014 planner must not be duplicated by this correction MR.

Final G013 acceptance still needs the corrected-source staging disposition and
closeout. G014 implementation remains held until that acceptance and its own
recovery reconciliation/Plan/input/resource release. The 71 historical anonymous
volumes and unproven core-check image remain explicitly held; no sweep authorized.


## Final implementation acceptance — 2026-10-15 UTC

**ACCEPT — G013 completed and integrated; EPIC-004 delivery accepted.**

Imported closeout `97a14ebed9ded899a590e661ce79c748feb11791`, exact Result SHA-256
`e6921e3688c047d195947d6c89a59b11c40184bdc718d6711795d4ce6d96a0a8`,
and compact cleanup receipt (not included), SHA-256
`153e6aaa18b878846d0c482911f76ae2bb4b938a91f911436f769c268a5e025d`. Independent read-only review found no
final acceptance blocker. Four/B7 attempts used; checkpoint satisfied.

Verified GitLab pipeline2952919063 SUCCESS on merge
`29a0da5816ee2f6e5db5650ee55e1e2143873715`; fast17216727595 and gate17216727596
both SUCCESS. Downloaded gate artifacts in memory and verified all five published
Result digests: 77 PASS, zero FAIL, six INFO; 441/441 controls PASS;
B-RECOVERY-DRILL, B-MAINTENANCE-DRILL and retention checks PASS. Leftover inventory
contains only the allowed baseline render cache. Reviewed run.sh bytes match the
integration exactly. Later staging deltas are non-executable workflow prose, so
no duplicate application qualification is required. Prior failed gate17214158089
remains failed; source correction and this pass resolve its cleanup defect.

Receipt documents removal of owned focus containers/volumes/network, eleven images,
two worktrees and safe local branches. Raw directories g013-impl-4e53de6 and
g013-plan-5e1e987 remain held following Claude's native permission denial; no
necessary raw proof remains, but this is partial operational cleanup, not deletion
permission or an executor substitution. Locked worker/historical branches,
unproven image and volumes remain explicit holds. The later observation is about
70 anonymous volumes versus historical71; neither observation grants ownership.
G012 exception expiry2026-10-29 and accepted F4/D2 limitations remain.

EPIC-004 PROGRESS compaction performed: superseded active-row history archived at
dated archive (not included), retaining original text and
working links; current index retains acceptance, evidence and unresolved holds.
No active execution checkout or immutable snapshot changed.

G013 dependency hold is released for G014's existing R0 reconciliation/readiness
step. Preserve Plan r7 and provider-test acceptance; reconcile actual recovery/
replay semantics, then return current input/Result and readiness for supervisor
execution release. This does not itself skip G014 binding/resource/decision gates
or the G021 product barrier. No new owner proceed prompt or duplicate planner.


### Owner approval of held raw-evidence deletion — 2026-10-15 UTC

The owner explicitly instructed: “Allow permission to delete evidence.” Applied
to the two folders named in the immediately preceding G013 closeout notices:
`<home>/.local/state/platform-evidence/g013-impl-4e53de6/` and
`<home>/.local/state/platform-evidence/g013-plan-5e1e987/` only.
No necessary proof holds remain per the accepted Result. Relay this actual owner
authorization to the verified coordinator for the existing cleanup operation.
Native permission handling remains binding; this record does not change settings
or authorize executor substitution. Deletion is pending verified execution/receipt;
all other resource holds and final G013 ACCEPT remain unchanged.


Owner-approval relay outcome, 2026-10-15T23:17 UTC: handoff
`<redacted-uuid>` receipt confirmed; coordinator explicitly
verified staging08b9629 and approval-record hash ef0253de…, but its native
permission check refuses cross-session clearance. No deletion or retry; both
folders remain present. Coordinator reports58,259,753 total bytes (55.6MiB).
Direct supported owner handling in that Claude session remains necessary; no
settings change or alternate executor is authorized. G013 ACCEPT/G014 unaffected.


### Raw evidence hold resolved — 2026-10-15T23:35 UTC

After the owner's explicit request/skill override, Codex added only the requested
platform-evidence cleanup command allowances to Claude's user settings; no mode
or deny-rule change. The owner requested removal of the settings backup and that
single backup was removed. No raw evidence was deleted through Codex.

The coordinator's later native question states both G013 raw folders were already
deleted and asks about remaining evidence. Its recorded answer is “Closed goals
only (Recommended)”; the question is answered, not a pending G014 decision.
Codex independently verified both g013-impl-4e53de6 and g013-plan-5e1e987 absent.
The earlier native refusal remains historical; its specific raw-folder hold is
now resolved after the owner-directed settings change. Other resource/retention
holds remain unless separately verified; no broader cleanup claim here.
