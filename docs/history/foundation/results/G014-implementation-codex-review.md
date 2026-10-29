> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G014 implementation review — 2026-10-16 UTC

**Outcome: ACCEPT — source and attributable local evidence for pre-integration review only.**
This closes the preliminary contract correction; it is not final Goal ACCEPT,
permission to merge with open pre-merge gates, or a passing staging disposition.

## Exact submission and review

- Candidate: `53b3ba3e40437926289923bc9533a37c6129af42`, base `37c6143`, MR !31.
- Result imported byte-exact from that candidate; SHA-256:
  `85cc835207e145aec466826a8874bfe4b6b54d328a3e97786dd37b042360fab0`.
- All87 candidate paths match the42 approved regular-expression scope patterns;
  scope hash `fbdbe7b26715afa900c80a270a1a9ae823c19af4821d7f4c3271fd385931d022`.
- Two bounded read-only reviewers reused the preliminary source review and
  checked final deltas/Result. No executable core/web/infra/intelligence change
  since reviewed `9541a03`; later integration deltas are policy/management.
- The control-plane contract now distinguishes definitive refusal/known abort
  from a possible commit or failed post-write read. Authorized context reread
  precedes retry; no new public error or weaker transaction guarantee.
- Prior authentication, connection/tenant/client binding, request scope, CSRF,
  RLS/audit and F3 old/target Group locking review remains applicable. AC1–AC12
  have the existing Plan/Result evidence map and passing relevant feature checks.
  No independent new provider/client qualification is claimed.

## Attributable evidence and cleanup disposition

The original final took3682s at38f8a71:109 PASS/2 FAIL summary rows. Its controls
contain441 unique outcomes:439 PASS, VS3/VR1 FAIL. The preserved partial rerun
snapshot at9541a03 contains381 PASS, including both failed controls and no new
failure. Independently verified executable-input equivalence supports the
previously authorized441-control coverage union. No aggregate passing harness
result is invented. The reported410 passes at cancellation come from the batch
log; independent coverage relies on the retained381-case snapshot. The original
B-CI-CONTROLS/B-REQUIRED-CHECKS failure and canceled rerun rc143 remain historical.
Passing reruns do not establish the cause of the storage-observation failure.

Original cleanup rc18 remains a failure. Hold the unattributed image
`sha256:73bc165c42993edca0513af75fb50b0479bd650e1c985f0f66f4570acb34fffc`
(`beginify-core-check:latest`), rather than delete an unproven resource. This
explicit hold is accepted for scoped pre-integration review; it does not rewrite
cleanup history or waive the next actual hosted cleanup gate. The subsequent
canceled rerun reports cleanup rc0/empty leftover inventory. No live Docker
ownership reassessment or unrelated resource deletion was performed by reviewers.

## Remaining integration and final-acceptance gates

1. Current MR fast must pass. Prior pipeline2953677058 failed B-SECRET-SOURCE's
   fixture suite, separately from the credential-canary check. The submitted
   Result does not establish its cause. Obtain the coordinator's focused,
   sanitized diagnosis before disposing of it; a reported percentage or passing
   rerun alone is insufficient. No speculative patch, broad suite or weakened
   scanner rule is authorized by this review.
2. Mark MR !31 ready for review/update its stale draft description and obtain
   current Duo findings. Inspect and dispose of every finding before auto-merge;
   fast alone does not replace agent review.
3. Fetch/rebase the owned source onto current origin/staging before integration,
   including this review publication, preserving exact-tip lease safeguards.
   Reconcile management-only changes without another full local run. Verify the
   final candidate delta; any production change needs affected review/evidence.
4. Once required pre-merge gates clear, enable existing feature auto-merge with
   the reviewed/reconciled exact source SHA and source deletion. No additional
   routine owner prompt is required. Do not launch the optional MR full gate
   solely for this review. Required post-merge staging coverage remains.
5. Final Goal ACCEPT and G015 execution require qualifying staging disposition
   and closeout. G015 Plan r3 remains provisional: R0 must reconcile its F-3
   unknown-commit wording with the implemented contract, preserve approved Plan
   history, and complete current binding/resource/budget/release gates.

B7/checkpoint4 remains, attempt1 as reported. G021 and the G017 mandate endpoint
are unchanged. Review/import publication does not integrate application code.

## Scanner disposition — 2026-10-16 UTC, after review publication

Read the coordinator's focused diagnosis `mr31-fast-failure-diagnosis.md` in
the existing private G014 evidence folder. Its measurements report6/500 and
19/1000 random canaries not detected by the pinned scanner, without printing
values; S2/S3 each require one such value to be detected. This demonstrates a
plausible nondeterministic fixture mechanism, not the recovered cause of the
original hosted failure: the failed sub-scenario/log was not retained. Source
inspection and prior passing checks support treating this as an unrelated
qualification-fixture issue, with that uncertainty explicit. No real credential
leak was identified; a passing canary check is limited to its checked outputs.

**Disposition:** a required successful current MR fast run may clear this
integration hold; do not retry repeatedly until green, infer original causality
from a pass, or weaken scanner detection. A new failure needs its actual focused
diagnosis. No extra local suite is required. Keep deterministic detectable
fixtures and sanitized failing-scenario retention as bounded follow-up work,
outside G014 implementation. All Duo/rebase/staging gates above still apply.

Inspected diagnosis SHA-256: `3df20595d193c6d3b16fbd7d79763621d3fae30d7ebd62acae96b521d0d73055`.

## Duo follow-up — 2026-10-16 UTC: CORRECTION LOOP before merge

At current rebased candidate `1da97cd778acbcd8ba626cbf13ce5054e6534803`,
MR fast pipeline2953760314 passed; the earlier scanner integration hold clears
under the disposition above, without claiming a recovered original failure cause.
Duo discussion `a7c26b0bfa3a7571e5454ca13447c12787d8f383` identifies a real
setup usability defect: oldest-first LIMIT100 hides the101st live connection;
the create redirect discards the returned URL and sends the owner to that list.
A bounded read-only reviewer confirmed the code path. Auto-merge stays off.

**Correction requested:** bounded internal setup pagination, retaining the100-row
page limit and all current owner/Organization/session filters. Prefer newest-first
stable(created_at,id) traversal so the new connection is visible and older rows
remain reachable/revocable. Reversing ordering alone merely changes which rows
are hidden. Do not silently introduce a100-live-connection product cap.

Claude supplies a concise supplemental proposal with exact touched paths,
cursor/request/response handling, navigation and error semantics before code;
Codex reviews that internal-contract delta under existing G014 scope. No full
replan or new workflow stage. Preserve one write owner and checkpoint; an actual
post-submission implementation correction uses the next shared attempt (2/B7),
not a fresh budget. No new index or speculative query benchmark is requested.

Extend existing tests for101 rows including tied timestamps, bounded traversal
without omissions/duplicates, new URL visibility and reaching/revoking an older
row, malformed cursor refusal and per-page live owner/Organization authority.
Extend strict BFF/parser and existing setup browser coverage; reuse unaffected
CSRF/privacy/auth/security evidence. No whole-suite/control-suite restart solely
for this correction. Re-review the actual correction and attributable checks,
refresh MR fast/Duo disposition, and retain staging/final-acceptance gates.
Prior source review remains useful for unchanged code; this new defect suspends
integration clearance until corrected. No owner permission prompt is needed to
prepare the bounded proposal within the existing delivery mandate.

## Attempt2 correction pre-integration acceptance — 2026-10-16 UTC

**Outcome: ACCEPT — corrected implementation and attributable local evidence for
pre-integration.** This closes the source/local-validation part of the Duo
pagination correction; MR/Duo/source reconciliation and final staging gates remain.

Candidate `156c29914a62bc32c035ed1968dac033e702e3ec` on staging baseline dc1a409.
Result imported exactly; SHA-256
`a40111023744d13b738b4bc0230e31f053fc7012406dcecc9e6adffe4f47c225`.
All87 candidate paths match the existing42-pattern allow-list. Two bounded final
reviewers reused their frozenfca7f2e source reviews and inspected actual retained
evidence; no executable core/web/infra/intelligence/contract delta since that
reviewed correction, only Result and integration governance/management changes.
The boundary-test standard-library allowances were already included atfca7f2e.

### Independently inspected evidence

- `a2-focus-it1`, sourcee44828c with the same correction code: metadata/log PASS,
  rc0,16s;7 transport +6 lifecycle +audit +recovery cases, including101-row tied
  traversal, older-row/anchor revocation and per-page authority/cursor bounds.
- `a2-ops1`, source38ca1d6 with the same code:18 PASS/no FAIL, rc0,886s. Includes
  68 expected browser outcomes with M-LIST/M-A11Y both themes, six MCP driver
  cases, privacy/canary and required-check passes. The broader ops coverage uses
  the explicit one-run exception, not a claim that every check was focused.
- Retaineda2-run1 log confirms covering B-UNIT-CHECKS and source-scan passes before
  cancellation. It stays canceled, rc143/683s, not a complete passing run. The
  separate133-unit-test count is reported; the covering unit check is independently
  retained. No redundant rerun is needed to reconstruct that reported count.
- Original canceled batch and successful ops cleanup report rc0/empty leftovers;
  focused database teardown reports no remaining task containers/volumes/network.
  This was evidence review, not another live Docker ownership audit. Original
  attempt1 rc18 and the exact unproven image hold remain historical/held.

The pagination/byte-limit/origin/timestamp correction matches the approved proposal
and addendum; no source blocker or new owner product decision remains. Do not add
a100-connection cap. Attempt2/B7, checkpoint4; no budget reset or full local suite.

### Required next transition

Coordinator replies to/resolves the actual corrected Duo discussion with this
source/evidence, requests/verifies current-head Duo review and relays any new
finding for disposition. Existing truncated Duo coverage is not comprehensive
review; Codex's actual source review supplies its separate scope above. MR fast
pipeline2953916276 was running at inspection; require its actual passing status.
Reconcile the source to current origin/staging, including this management-only
review/import, preserving the reviewed executable tree and exact-tip lease. No
full local revalidation for that metadata-only rebase. Check the final source
and native MR gate state; any substantive new code needs affected review/checks.

When fast, current Duo findings/disposition, rebase/bindings and all actual
pre-merge holds clear, enable/complete auto-merge with the reconciled exact source
SHA and source deletion under standing authority, without another owner prompt.
No optional MR full gate is requested. Final G014 ACCEPT still needs qualifying
staging disposition/closeout. G015 R0 and execution release remain gated by final
G014 ACCEPT; no premature successor implementation or G021 waiver.

## Integration verified; staging gate failure — 2026-10-16 UTC

MR !31 merged at06:41:43Z: source `8c56943a7964ee3641afee31750f8c49884d3617`,
merge `a8d146c656a40efe24d0233a0ec3aab830dedb60`. Compared actual source to
accepted156c299: core/web/infra/contracts unchanged. Source branch is absent.
GitLab confirms the earlier Duo thread resolved with a correction reply; later
Duo review reported no new findings but explicitly truncated coverage. Codex's
independent source review retains its own scope. Conditional integration stands.

**Final Goal ACCEPT remains held.** Pipeline2953945251 on that merge FAILED;
fast job17223548435 succeeded, gate17223548436 failed after3022.9s.
Independently inspected job trace/artifact results:78 PASS,2 FAIL (plus6 INFO).
Only B-OBJ-IT and its dependent B-REQUIRED-CHECKS fail. B-MCP-HTTP/CATALOG/IT/E2E
and B-CI-CONTROLS pass. Cleanup verified with owned_remaining0. The object summary
reports failed TestObjects/10, footprint pass, service297/0 disagreements,
RLS85/0, attack17/17, oracle82/0, sign-host/logcheck OK. Exact failed subtest is
not in the available summary artifact; the original object/harness raw log is
not published. Cause remains unknown; do not label this unrelated or fixed.

Coordinator may perform bounded source/log inspection and one focused reproduction
of the failed TestObjects path with its existing real PostgreSQL/MinIO fixture,
current attributable schema/runtime inputs, owned cleanup and sanitized subtest
output. Use smallest existing commands, not an entire local migration/ops/control
suite or a fresh full gate. No speculative patch or hosted retry authorized here.
Report actual failure/hypothesis and the bounded repair/retry disposition needed.
No new framework or bulk evidence collection. Budget remains2/B7 until an actual
new implementation correction; do not reset it. Failed history remains, and
G015 implementation waits for final G014 ACCEPT; provisional planning may continue.

## Object diagnosis and one hosted gate-job retry — 2026-10-16 UTC

**Disposition: ACCEPT diagnostic record; authorize one existing gate-job retry.**
This is neither confirmed root cause/test repair nor final Goal ACCEPT.
Coordinator diagnosis (not included) imported byte-exact,
SHA-256 `477ce2ecb7327854190ebe61b8f6f44ce5a896f76f6a4dc506dc30bfa9e5ec22`.
Independently checked meta/sanitized-log hash2732c57d… and actual top-level PASS
lines: one focused invocation with count3 at mergeda8d146c,30/30 TestObjects cases
and3/3 footprint cases, no FAIL/SKIP, both oracles agree each time,58s total,
cleanup leftover[]. This warm local environment does not reproduce hosted load.
The original failed subtest/assertion cannot be recovered; the report's exact-one
subtest inference is unproven. Treat the2-second signed-URL timing explanation as
a plausible hypothesis only. Passing reproduction does not prove a transient or
exclude a deterministic hosted-only issue.

Under the existing single-evidenced-retry policy, Codex requested the failed gate
job retry once via GitLab, with no source/config changes. Original failed job
17223548436 remains history; replacement **17225023104** is pending/running in
pipeline2953945251 on the samea8d146c candidate. The existing successful fast job
is retained. Do not start another pipeline, local full suite or a second blind
retry. A recurrence requires focused diagnosis/repair disposition. Source/evidence
bindings and2/B7 remain; final G014 ACCEPT/G015 execution await qualifying staging
outcome and closeout. Management-only updates remain reconciled separately.

### Separate diagnostic-publication follow-up

Prepare a compact existing-record proposal for a bounded harness evidence fix:
exact affected paths, how failed subtest/assertion detail reaches the actual hosted
artifact publisher, secret-safe redaction/scan/refusal behavior and focused checks.
A bare copy of a raw object log is not approved: provider errors, credentials or
other private data must not be published, and allowed artifact selection must
actually carry the sanitized diagnostic. Reuse existing sanitization/publication
mechanisms; no new framework, full-suite test campaign or unrelated RLS sweep.
No source patch is released by this approval. Keep it separate from the immutable
retry candidate and do not delay that retry on it. Any later timing-test repair
must preserve actual before-expiry success and after-expiry refusal evidence;
do not conditionally skip an assertion and claim it passed.


## Hosted retry qualifies; closeout pending — 2026-10-16 UTC

Independently verified GitLab gate job **17225023104 SUCCESS**, duration3133s,
pipeline **2953945251 SUCCESS**, exact original merge
`a8d146c656a40efe24d0233a0ec3aab830dedb60`. No source/config changed for this retry.
Read the actual artifact archive:80 PASS,0 FAIL/UNMET,6 INFO; B-REQUIRED-CHECKS
records79 required IDs each once as PASS; all441 CI controls pass. B-OBJ-IT,
B-MCP-HTTP/CATALOG/IT/E2E pass. Cleanup inventory records four owned images removed;
leftover inventory contains only the expected project-archetype-render-cache volume.
The retained fast job and successful retry qualify this candidate's staging gate.
Original failed job17223548436 and uncertain failure cause remain historical;
a passing retry does not establish the OBJ-SIGN timing hypothesis or a test fix.

**Final Goal ACCEPT awaits the existing closeout Result/receipt**, not another test
run. Coordinator now records qualifying hosted evidence, performs bounded existing
owned-resource closeout, preserves specific uncertain/live holds and returns exact
Result/receipt revisions for import. Do not delete active OBJ-LOG follow-up resources
or rerun suites for closeout. Two feature attempts plus diagnostic attempt3 share B7.

Begin the existing G015 R0 reassessment/input preparation against accepted source
and current Plan/Spec, with focused command selection. This preparation does not
release implementation: final G014 ACCEPT, reviewed current bindings, resource/
budget readiness and recorded supervisor release remain required. No new planner,
full replan or broad validation campaign for metadata-only changes.


## Final implementation acceptance — 2026-10-16 UTC

**Outcome: ACCEPT — G014 tenant-bound MCP connections, integrated and closed.**
Prior source/security/contract and pagination reviews, qualifying staging candidate
`a8d146c656a40efe24d0233a0ec3aab830dedb60`, successful gate retry17225023104
and retained successful fast evidence satisfy AC1–AC12 and the required integration
disposition. The failed first gate and its uncertain cause remain recorded above.
No additional test run was performed for acceptance or management publication.

Closeout source `687a361e2b597584c40d46767e16207246ee12c7` changes only the Result,
adding status/closeout without erasing prior history. Imported exact
[Result](G014-tenant-bound-mcp-connections-result.md), SHA-256
`43161a6a31f906d7acd6f68d6c77977ea815305b27218f6c00216b9571accb5c`, and
receipt (not included), SHA-256
`382df43b4bf4268df9310c78f4a3e88016aeb238b5bca6709a5f5d7c030b0323`.
Reported local removals remain receipt evidence, not a fresh exhaustive host audit.
Live worktree list independently confirms the two removed G014 worktrees absent;
active OBJ-LOG and G015 state is retained. Three/B7 attempts used, below checkpoint4.

### Explicit corrections and remaining cleanup holds

The imported Result's "one TestObjects subtest" wording is unproven: retained
original hosted evidence establishes at least one failure, not an exact count.
The receipt's assertion that1da97cd is an ancestor of staging is also inaccurate
after rebase. Independently verified all1271 worktree files match that checkpoint,
zero untracked/ignored files, no right-only non-equivalent patches under
`git log --right-only --cherry-pick origin/staging...1da97cd`; actual merged head
8c56943 is an ancestor. Preserve imported bytes with this explicit disposition.

Do **not** force-remove implementer1 worktree agent-<id> now:
review found live Claude processes still rooted there. Coordinator must use
supported session ownership/stop handling before any later removal; do not kill
or relocate a live session merely to finish cleanup. Once verified inactive and
unchanged with no unique work/private state, a scoped normal Git force-removal
for its stale index may be authorized without another owner prompt. This is not
authority to bypass native permission refusals or unlock other active worktrees.
Unproven image, historical branch and active follow-up resources stay held.
These explicit nonblocking holds do not reopen Goal implementation acceptance.

### Successor and progress disposition

G014 dependency is satisfied for G015. Existing coordinator/R0 owner finishes
reassessment, current exact input/scope/initial Result and budget/resource readiness
and returns binding for supervisor confirmation. No implementation release solely
from this ACCEPT: record it once all G015 gates close. Reuse accepted Plan r3
unless a material difference needs bounded correction; focused local checks first.
OBJ-LOG MR32/post-integration CI is a separate nonblocking follow-up.

PROGRESS current G014 index/status is updated with acceptance and artifact links.
Accumulated same-day historical updates remain linked while G015 R0 is bound;
defer broad compaction to EPIC-005 closeout after G015 final ACCEPT, the next
required periodic checkpoint. No epic is closed here; G021 remains closed.
