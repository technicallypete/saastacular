> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G010 implementation review

## Staging gate failure — G7d disposition, 2026-10-14 UTC

Pipeline `2947802084` on exact integrated `c961e56` failed gate job
`17179052627`; fast job `17179052626` passed. Hosted artifacts verify 440/441
CI controls passed, with G7d's sole reported failure "the guard was not identified
exactly once". Its other assertions (exit 143, guard-death diagnostic and owned
context stop) did not fail. This is not proof of the exact cause: the published
log does not retain the killed-PID count. Do not call the check passed or waive
it because G010 did not change the control. Final G010/EPIC-003 ACCEPT and G011
execution remain held. Prior source/local acceptance remains scoped as recorded.

Codex authorizes the coordinator to obtain **three focused G7d repetitions**
using the existing control runner and normal target/input/environment/resource
gates. Record the actual killed-record count and safe PID-selection evidence
if available, retaining no environment/credentials. No full local suite,
benchmark, broad CI redesign or test weakening. Diagnosis/retries alone do not
spend a new implementation attempt; B5 still has one submitted attempt.

If all three repetitions pass unchanged and no new relevant fault is found,
one retry of the failed hosted gate on the same integrated source is authorized;
verify the retry job/source and record both the failed run and retry separately.
If any repetition fails, stop the focused batch, report the count/selection and
smallest proposed correction to Codex. If the hosted retry fails, return the
new evidence before another retry. No speculative source patch or direct push
is authorized by this diagnostic disposition. Use the existing task owner, not
a second coordinator or duplicate validation campaign. Cleanup owned diagnostic
resources under the existing procedure. This management-only record does not
invalidate unchanged G010 source or require a new Plan.

## G7d focused evidence and authorized retry verified — 2026-10-14 UTC

Codex inspected the coordinator's `g010-impl-325c136/g7d-diagnosis.md` and
all three run status/summary/control logs. Existing command
`bash infra/ci/validate-controls.sh --only G7d --keep-scratch` ran unchanged
integrated `c961e56` at 19:08:01, 19:08:17 and 19:08:40 UTC. Each exited 0:
one selected control passed, exactly one killed-record line (fixture PID 58),
six shim Docker calls and one owned stop. Environment/input/lock gates are
recorded; each scratch directory was removed. No source correction or full local
run. The hosted failure's exact cause remains unproven; local passes are not
a root-cause finding or a waiver of the failed hosted run.

The authorized retry condition is satisfied. GitLab verifies retry gate job
`17180444863` is running in the same staging pipeline `2947802084`, on exact
`c961e562f1041b51d5f45d0e9bbf172813c030d1`; fast remains passed. Preserve failed
job `17179052627` as history. No additional retry is authorized on another
failure. Final G010/EPIC-003 acceptance and G011 release still require the
qualifying retry disposition, final reconciliation and bounded closeout.
No new execution attempt consumed.

## Attempt 1 pre-integration review — 2026-10-14 UTC

**ACCEPT — scoped implementation and local evidence for integration.**
This is not final Goal acceptance or EPIC-003 closure.

Reviewed candidate `71c8c47b2ee3591d5bd370720717e5823e1a807f` on branch
`g010-foundation-isolation-attack-suite`, based on staging
`4a21892`. Imported Result byte-exact, SHA-256
`2a732fab43d9531b2fdce1a14aee96180f0ec9abf5389b209d305742a13dda8d`.
Approved Plan r1 remains `ca6f2176cc7019dff8fc8912b90e53951ed5210efcb82ac3566957a551b76a78`.
Two bounded independent read-only reviewers covered actual source and
Result/provenance; Codex reconciled their findings and the actual delta. No
Codex application tests ran and no Claude implementation was rewritten.

### Source and requirements

- All seven changed paths are within the approved scope: two new tests, minimum
  obj_it wiring, three foundation documents and Result. No production code,
  schema, dependency, public contract, capability, CI framework or new check ID.
- Oracle r1 is the original author's byte-exact artifact, SHA-256
  `e50f30cdbe50eed7bae6cf9574aabcd3c9ccedce58b99ab0a00afa2be25c7582`.
  Its 82 literal expectations have measured outcomes; missing/duplicate/unused
  expectations fail. The two prior contract/citation rulings are applied with
  r0 retained. No observation-generated policy expectations.
- Ten composition steps cover tenant/context substitution, profile non-transitivity,
  uploader authority, live-fact changes/pool reuse and staged storage residuals.
  Denials require exact errors, earlier working positives, zero provider effects
  and unchanged owned state. SQL uses attested runtime credentials/exact row sets;
  setup privilege remains separate from assertion authority.
- D2 verifies returned IDs, backend identity and empty context after ordinary
  rollback. E1 verifies compensation and non-publication. E3 verifies the live
  object, committed revocation, no later provider calls, deleting/absent state
  and authorized repair. The acknowledged bearer/runtime-credential trust limits
  remain explicit, without new revocation/compliance claims.
- Existing B-OBJ-IT runs the additional named test, requires 17 passing subtests
  and exactly 82/0, retains predecessor checks and privacy scans. Security-review
  fixes are present; no substantive new source correction is required.

### Validation and provenance

The six test/harness/document content blobs and application/infrastructure/
contract trees match across b12 `3f1f3f2`, b16 `36e1322` and final submission.
Later rebases contain only management/governance changes, explicitly reconciled.
The Result is report-only and was appended after validation under the accepted
protected-content procedure.

- AC8 b10/b11: disposable `f8f412a` reverses exactly C1-04 forbidden→ok. The
  existing gate fails B-OBJ-IT plus its dependent aggregate, and the focused
  output names only that disagreement. Restoration is verified by the clean b12
  candidate/correct oracle; no current remote reference contains the disposable
  commit. This is not a historical GitLab server-audit proof of never pushed.
- b12 `all unit`: 96 PASS, B-CI-CONTROLS and its dependent aggregate FAIL;
  TestIsolationAttacks is 17/17 and 82/0. VS2 is the failed control.
- b14 retains its separate VC7 failure while VS2 passes. Focused diagnostics
  passed. b16 is the successful covering failed-suite rerun: exit 0, 441/441
  CI controls, all 20 required IDs PASS, 21 passing records, no required skips.
- Reuse b12's unaffected passing coverage plus b16's replacement for the failed
  suite. Both runs' leftover inventories are empty. No duplicate unchanged full
  run is required solely to obtain a single all-green local report.
- VS2/VC7 causes remain unproven; their earlier failures are not rewritten as
  passes. A future failed required hosted check still blocks final acceptance.

AC1–AC8 and local AC9 evidence support integration. AC9 hosted disposition and
AC10 final cleanup/epic closeout remain. Attempt 1 of B5 is submitted; no correction
attempt or budget extension is needed. Checkpoint after attempt 3 remains.

### Authorized next action and remaining gates

Under the owner's existing G010 unattended-loop/MR authority, Claude may fetch,
reconcile/rebase onto current origin/staging, push and open a non-draft MR to
staging with source deletion enabled. This management-only review/import does
not require another local qualification run. Bind evidence to unchanged content
and report the current target/candidate. MR fast applies; an optional manual MR
full gate is not requested. Inspect and dispose Duo findings before eligible
SHA-bound auto-merge; remind the owner of unresolved findings, if any.

Required staging path-selected coverage remains applicable. Only its qualifying
disposition plus source/Result reconciliation, bounded owned-resource/evidence
cleanup and final Codex ACCEPT closes G010. Codex then records EPIC-003's mandatory
PROGRESS compaction disposition, including the prior G008 deferral. Preserve
failed/interrupted history; purge temporary raw evidence at verified closeout,
keeping Git records/CI artifacts and explicit necessary-proof holds.

G011 provisional planning may continue at xhigh, reconciling final G010 state;
G011 implementation is not released. G012+ dependencies and the G021 product
barrier remain binding. No architecture decision, new framework or extra local
control campaign is needed for this acceptance.

## MR !26 / Duo disposition — 2026-10-14 UTC

**ACCEPT — both Duo notes are nonblocking robustness follow-ups.**
MR source `7a749e938a8a79fc4b7099e5adcc78421d499d11` has the same six G010
content blobs as reviewed `71c8c47`. Pipeline `2947758142` succeeded, fast job
`17178748258` passed. Duo's completed review reported no actionable issue and
no resolvable findings, then identified two minor robustness observations.
Codex and the existing bounded source reviewer independently inspected both:

- E1 resets its recording-only spy hook on the happy path. A fatal exit could
  leave it installed for later subtests; it only records operation names and
  returns nil, and E1's fatal exit already fails the suite. Deferred cleanup is
  useful future failure isolation, not a correctness/merge blocker.
- Oracle check reports a missing agreed positive using t.Errorf, then can mark
  the local observation agreed. Transitive bookkeeping could be clearer, but
  the existing error permanently fails the subtest/Go invocation and the pinned
  harness cannot turn it into a pass. No false-positive acceptance path identified.

No correction, extra test campaign or implementation attempt is requested.
Consider both changes when next editing this test; do not claim they were fixed.
The owner was notified of the observations and this proposed disposition before
merge. Existing standing feature-integration authority applies.

Target advances since the MR baseline are G011 audit specification/review and
management publication records only; no tested G010 source/build/test/harness
input changed. Claude may complete the current-target reconciliation required by
the branch procedure and merge/enable SHA-bound auto-merge with source deletion.
Any rebased candidate must retain those reviewed content blobs; reuse unchanged
local coverage rather than another broad run. New actionable findings or actual
content changes still return for review. Qualifying staging disposition, cleanup
and final EPIC-003/Goal acceptance remain pending.

## Conditional integration verified — 2026-10-14 UTC

MR !26 merged at `c961e562f1041b51d5f45d0e9bbf172813c030d1`, source
`7a749e938a8a79fc4b7099e5adcc78421d499d11`, after the recorded pre-integration
ACCEPT, passing fast and nonblocking Duo disposition. Source deletion is enabled.
The main local staging checkout fast-forwarded to the integrated commit. Staging
pipeline `2947802084` is running; final G010 ACCEPT/EPIC-003 closure remains held.
No repeat local tests were required by the intervening documentation-only target
changes. G011 planning continues, reconciling these actual integrated tests/harness/
foundation documents at its next safe boundary; no G011 implementation release.


## Final implementation acceptance — 2026-10-14 UTC

**ACCEPT — G010 and EPIC-003 closed.** All ten acceptance criteria are satisfied.
Source/local acceptance above remains valid; no source correction was required.
One submitted attempt of B5 was used. No checkpoint-3 requirement arose.

Codex and a bounded read-only hosted-evidence reviewer verified pipeline
2947802084 (not included)
and its authorized retry gate
17180444863 (not included)
SUCCESS on exact `c961e562f1041b51d5f45d0e9bbf172813c030d1`.
Fast `17179052626` also succeeded. The retry took 3126.8 seconds:

- 69 PASS, zero FAIL/UNMET/SKIP: 68 underlying required IDs plus their completeness
  aggregate, each once; all 441 distinct controls PASS, including G7d.
- B-OBJ-IT: G010 17/17 subtests, 82 expectations, zero disagreements.
- Final trace: overall status 0; cleanup, publication and scrub all status 0;
  owned registry removed. Summary publication fields are pre-publication snapshots.
- Artifact ZIP independently inspected in memory only, SHA-256
  `53660e9c72f2cc92ac031c90587edd82c0313223404e36f06964f4122e2c7afa`.

Original gate `17179052627` remains FAILED (440/441 controls, G7d identification
count; 67 other checks PASS). Three focused unchanged runs and this retry do
not establish its original cause. VS2/VC7 historical failures likewise remain
unproven and failed, rather than being rewritten as passes.

All six delivered content blobs and the core/infra/web/intelligence/contracts/
foundation-document trees match reviewed `71c8c47`, merge `c961e56` and inspected
integration HEAD `d4c07bd`. Later differences are management/governance/future
architecture only. No duplicate application validation was necessary.

Imported the report-only [final Result](G010-foundation-isolation-attack-suite-result.md)
from `784f1441c3f8745bac94e36c727b1eee9e680179`, byte-exact SHA-256
`7b6e64cb0e49d69817a8cf99722a7243ec9e1d3219c6871a59835add8bb03dc6`.
Its compact replacement retains the material commands, outcomes, oracle rulings,
AC mapping and budget; the preceding full Result remains in Git history.
The phrase "never-pushed" in its AC8 row is the implementer's reported history;
Codex verified no current remote reference contains the disposable commit, not a
historical GitLab server-audit proof. No necessary local raw-proof hold remains.

Imported the cleanup receipt (not included), SHA-256
`8739925b1f761ac40f0a267e63564e29cacb554f4ee2be90b8e20175b2d7a5c1`.
Independently checked implementation worktree absence, remote refs and absence of
G010 focus images/volume/containers. Its three temporary evidence folders were
purged after compact records were made. The planner worktree remains explicitly
held by the native tool lock; remote `g010-plan` and safekept `g010-closeout`
remain until that supported closeout completes. They do not imply live G010
implementation or invalidate acceptance. No force unlock/removal is authorized.
Coordinator, G011/G054 workers, unproven old volumes and unrelated services remain.

Codex completed the mandatory PROGRESS compaction at EPIC-003 closeout, fulfilling
G008's recorded deferral. The full previous index is archived with source commit,
original hash and relocated-body hash; no Goal/Spec/Plan history or failed evidence
was erased. No test campaign is required for index/report maintenance.

G011's predecessor is satisfied: reconcile its already accepted Plan and prepared
inputs against current staging and record execution release. G011 receives audit
mapping/read isolation; G013 receives tombstones/pending rows; G014/G015 receive
scope-trust boundaries; G018/G019 receive membership-administration observations.
These are successor handoffs, not newly implemented guarantees. G021 remains closed;
no public adapters/product features, deployment, main promotion or release authorized.
