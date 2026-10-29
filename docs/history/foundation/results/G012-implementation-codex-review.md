> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G012 implementation — Codex review

## Attempt 1 source/local-evidence review — 2026-10-15 UTC

**CORRECTION LOOP — two bounded fail-closed defects.** No MR/integration or final
Goal ACCEPT yet. Authorize correction pass 1 as Goal-level attempt 2 of B5,
checkpoint after attempt 3. No reset, extension, scope expansion or Plan rewrite.

Exact submitted head `6654e15d389162b5e0d401558a2d63de4de6802e`, tested code
`c9f874e4cffb515f33da3215ea278f7df0c2bb78`, base
`34e036258b486b079867ad8e277393b25e376ea3`. Imported the submitted Claude
Result byte-exactly, SHA-256:
`a53e195aa5a66bcd5c417061f60734a8aa220255426b58c1d7be307a709ac3f6`.
Accepted Plan r1, D0/owner amendments and nineteen-pattern scope remain binding.

Codex inspected actual source/Result with independent bounded source-security
and scope/evidence reviews. One reviewer verified coverage/provenance with no
material evidence issue; security source review found the two defects below.
No tests or implementation edits by Codex.

## Required corrections — Spec R4 / AC6 and AC9

1. **Reject line-scanner truncation.** In
   `infra/validation/tools/cmd/vulnpolicy/exceptions.go:102–105` and
   `parse.go:107–111,269–273`, the three bufio.Scanner consumers return parsed
   prefixes without checking Scanner.Err. A valid exception row followed by an
   over-limit line can hide a later invalid/expired row. A module/uv input can
   likewise terminate without a parse failure. Return an input error on scanner
   failure before using any partially parsed policy/graph. Explicit bounds may
   remain; increasing the buffer alone does not fix error handling.
   Minimum regressions: valid prefix plus over-limit line in exceptions, go.mod
   and uv input; include a later invalid/expired exception. Required outcome is
   invalid/incomplete input (CLI exit 2), never clean or merely unused-row status.
2. **Require complete OSV advisory references and actual EOF.** In
   `policy.go:255–279`, validation rejects advisory records missing a group but
   accepts group IDs with no corresponding advisory record. Such a package can
   keep its identity and a Low group score and pass incomplete input. Require
   every group ID to resolve to a nonempty advisory ID/record in that package;
   preserve existing reverse validation. In `parse.go:384–396`, Decoder.More is
   not a top-level EOF check: trailing closing delimiters can escape rejection.
   After the one expected document require a second Decode to yield io.EOF.
   Minimum regressions: group ID missing its record, empty advisory ID, trailing
   closing brace and bracket (and valid trailing whitespace remains accepted).
   Incomplete/malformed inputs must produce exit 2; existing valid clean/finding/
   exception results retain their intended outcomes.

Use the existing parser/evaluator and table-driven tests. No new scanner,
framework, policy threshold, exception, dependency or production code is needed.
These findings are source-level reasoning; Claude supplies the regression proof.

## Evidence retained and deviation dispositions

Independent verification supports 104 unique required PASS IDs and CI 441/441;
manifest/expected suites exactly match. Submitted head adds only Result to the
tested code. All 55 submitted paths match approved scope; final batch's 54 omit
that later report-only update. SECURITY's expected hash is e6bb1a70… . Prior
failures and original review corrections remain honestly recorded.

Go 1.27.2 is used by the actual four Dockerfiles, both module directives and
scanner; all thirteen prior stdlib findings clear. Original generator/skill/
manifest remains unchanged, 28 customizations and disabled-variant inheritance
pass. Rotation evidence covers old open connections, fresh-login denial,
termination, restart recovery and withdrawal.

The exact braces exception conditions are supported by real production-image
inspection and lint-input/call-path evidence tied to unchanged web bytes. Retain
High/EXCEPTED with October 29 expiry and earlier review conditions, never fixed.
No broader development-only exemption or extra Go exception is granted.

Disclosed residuals accepted within existing boundaries: private 0700 temporary
submitted-source staging removed by trap, without bind mount/network; root Go
containers limited to owned volumes, caps dropped/no-new-privileges/no secrets
or Docker socket, scans network-free. These do not authorize privileged runtime
service use. Small adjacent config hunks, pins INFO row, synthetic Go advisory
fixtures and extra gitlink coverage do not expand behavior/scope. The earlier
one-off version-request deviation stays recorded, not retroactively compliant.

## Minimum rerun and handoff

Reconcile current staging before the correction batch, preserving input/source
bindings and the raw failure history. Run validation-tools check/unit tests with
the added parser regressions and CLI invalid-input outcomes, plus B-DEPSCAN and
its covering B-CONTRACT-TOOLS/security publication checks and relevant format/
static checks. Keep source/exception canary and final scope/protected/provenance
accounting valid for the corrected tree; append exact evidence/hash mapping.
Reuse passing runtime/browser/rotation/migration/off evidence where bytes and
inputs remain unchanged. No second whole local all-unit campaign solely to
repeat those unchanged cases; applicable MR/staging CI is not waived.

The current allow-list already covers these fixes; no new paths or owner choice
is required. Existing implementer owns correction attempt 2, then submits Result
for automatic Codex re-review before MR. Remaining MR fast/hosted timing/Duo,
staging outcome and final cleanup/Goal ACCEPT remain. G013 planning waits for
G012 integration. G014 Goal/Spec human-input stop and G021 barrier remain.

## Correction pass 1 / attempt 2 — source and local evidence ACCEPT — 2026-10-15 UTC

**ACCEPT for MR integration review.** Both requested parser guarantees are
satisfied; this is not final Goal acceptance. Exact submitted head
`6780bcd5107a06186a190a6702dada4699690f2f`, tested correction
`67ef90c136134e7cda02d98a54058ba7aaf36601`, reconciled target
`3a852ba5bf479eb5717325767e8704e7438da343`. Submitted head changes only the
Result after the tested correction. Imported Result bytes have SHA-256
`c8efae0a251c1a348fac8c3e9d83f573345ac3e7f86fd34ddea16b388230bdc1`.

Independent bounded source/security and evidence reviewers verified the exact
correction delta. All three line parsers check Scanner.Err and reject partial
input. OSV group/advisory references are checked both ways with nonempty IDs;
a second JSON decode must yield EOF. Regressions cover valid-prefix truncation,
missing/empty references, trailing delimiters and second documents, CLI exit 2,
and valid trailing whitespace. No material residual finding. Only the four
reviewed vulnpolicy source/test files changed outside management; no policy,
exception, dependency, scope or runtime expansion. Prior correction history stays.

Focused c2-focus evidence is 29 unique required PASS IDs, exit 0, empty leftover
inventory. Exact results SHA-256:
`67344c9b72b4b9bb17bcaf9b5f85d812ef04873564f3523feb741aaa88c7743e`;
run log SHA-256:
`4e59449a14e3cdd40bda64ac9ad2f79521a33c7a85dfcdd3c096bd2402ac4743`.
Validation-tools checks include fourteen policy test groups. The focused IDs are
an exact subset of the prior 104-ID passing batch; the other 75 retain unchanged
runtime/browser/migration/rotation/variant/control inputs. The Result maps their
reuse explicitly. No duplicate full local run is required. Codex ran no tests
and made no implementation edits during this review.

Budget remains two used of B5, checkpoint after attempt 3. Accepted Plan r1,
D0/owner amendments, nineteen-pattern scope and exact braces expiry/conditions
remain binding. No new owner choice is needed.

### Authorized next action and remaining gates

The verified coordinator may fetch/reconcile current staging, push the reviewed
source and open its MR targeting staging under standing owner authority. This
review's management-only import does not itself require another broad test run;
record exact content/input reconciliation. Check required MR fast evidence,
first hosted scanner timing and current Duo/discussion findings before enabling
auto-merge for the reviewed current head. Submit substantive new findings for
bounded disposition; passing CI alone does not resolve them.

Integration is conditional pending the applicable staging disposition, scoped
cleanup and final Codex Goal ACCEPT. Verify source-branch deletion. G013 planning
may begin provisionally only after verified G012 integration; G013 execution
still requires final G012 ACCEPT and its own Plan/input/budget/resource release.
The G014 specification-only human-input stop and G021 product barrier remain.

## Hosted MR qualification hold — 2026-10-15 UTC

**CORRECTION LOOP — hosted failure requires diagnosis.** MR !28 remains open,
unmerged, auto-merge disabled. Pipeline `2949326116`, fast job `17189970959`,
source `27582356418d688695dbad79ccefe77358e22f83` on target `a8eb787` failed.
GitLab artifacts independently confirm B-DEPSCAN FAIL: synthetic govulncheck and
OSV fixture steps fail; advisory fetch, module setup, tool build and real-tree
scan pass. B-REQUIRED-CHECKS consequently fails. Other required fast IDs pass;
477-second pipeline duration is within the twenty-minute fast cap. This does not
invalidate the earlier local evidence, nor qualify the hosted run as passing.

Published artifacts retain the step summary but omit detailed fixture outputs;
the exact cause is unproven. The existing worker may continue read-only diagnosis
under the approved task. Capture exact failing outputs/assertions and compare
hosted inputs/environment with the local passing case before proposing a fix.
Do not weaken fixture expectations, label this transient without evidence, or
retry the identical job speculatively. No new owner decision, broad local rerun,
budget extension or expanded CI redesign is authorized. Submit the bounded cause,
patch/validation proposal and attempt accounting for supervisor disposition.
Prior source acceptance remains for its exact reviewed bytes; changed code needs
bounded re-review. MR fast, Duo, staging and cleanup/final ACCEPT gates remain.

## Hosted cidfile diagnosis and correction authorization — 2026-10-15 UTC

**CORRECTION LOOP — authorize attempt 3 of B5, no Plan revision.** The cause is
confirmed by source inspection and a clean-source reproduction at `2758235`:
CI supplies `--owned-registry`, forwarded as dependency-scan `--cid-dir`.
Repeated evaluate calls use the same `g012-<name>-<tag>.cid`; Docker refuses the
second and third fixture evaluations with exit 125, "container ID file found".
Local passing batches had omitted that integration. Reproduction results SHA:
`067cb109c8681e0cb709bac1d8408f549236e1ef81042b1fb33d691956676508`;
run log SHA:
`fd2300d6d06c021851f4c9d1f9aa0e1cde5629663eb98a2627c5a68c196ef024`.
Hosted raw per-evaluation outputs were not published; do not claim they were.

Authorize both bounded changes, already within the nineteen-pattern allow-list:

1. `infra/security/dependency-scan.sh`: each actual Docker invocation must have
   a distinct, initially absent cidfile, including command substitutions and
   pipelines. Reuse the existing BASHPID-plus-random suffix pattern in
   `infra/validation/lib.sh` or an equivalent invocation-safe identity. A plain
   shell-global counter alone is insufficient: subshell increments do not persist.
   Keep the flat registry cid directory and earlier cidfiles for cleanup proof;
   do not delete/reuse them, disable registration, alter ownership rules or weaken
   any fixture expectation. Handle identity-allocation failure closed.
2. `infra/validation/run.sh`: include bounded nonsecret failing-step assertion
   text/exit details in the existing published failure summary. Keep nonzero
   statuses and existing output/canary scans. No raw environment/credential dump,
   new publication framework or broad artifact/CI change.

After normal target/input reconciliation, run the source secret check and changed
script shell checks; reproduce dependency-scan with the real cid directory.
Both fixture families must give none/valid/expired outcomes 1/0/1, unique retained
cidfiles and no Docker 125; real dependencies still satisfy the exact existing
braces exception. Run covering `lint lint-base` with `--owned-registry`, verify
required-check accounting and zero owned containers/volumes after cleanup.
Prove the failure-summary path retains useful bounded detail and fails closed.
Reuse unrelated passing coverage; no full local rerun or repeated CI-controls
campaign. Append exact evidence, failed-run history and accounting to the Result.

Submit for bounded source/evidence re-review before pushing the fix to trigger a
new MR fast qualification. No speculative identical-job retry or auto-merge.
Attempt 2 ended at its submitted Result; this correction is attempt 3, not a reset
or extension. At its checkpoint record progress, remaining risks and forecast;
escalate material failure under the existing budget. MR/Duo/staging/final cleanup
and Goal ACCEPT remain; G013 integration dependency and G014 phase stop remain.

## Attempt 3 correction / checkpoint ACCEPT for MR qualification — 2026-10-15 UTC

**ACCEPT for corrected MR qualification and conditional integration.** Reviewed
`0aac74eeb7ac0856f0580bf6bc448b8c76126d8d`, tested correction
`d7ab6075b9d366a8e2b75754af8d2ed0a3ec6519`, target
`964dd8144d274a68701eddaca2e17f6aea66b3aa`. Result-only submission delta;
imported exact Result SHA-256:
`686244324c549a29e054df5cef1f9cd0eba6b56b809e97cb91c721bc086f2b5f`.
This supersedes the hosted diagnosis hold for the reviewed correction; the failed
pipeline remains failed historical evidence, not a waived or passing run.

Independent bounded source/security and evidence reviews found no material
residual. Only the two authorized scripts changed outside management. drun uses
BASHPID plus validated randomness, refuses allocation failure, retains earlier
cidfiles and preserves Docker/timeout/cleanup semantics. The existing FAIL
summary includes bounded assertion details; failure status and scans remain.
Codex made no application edits and ran no tests.

Direct cid-dir reproduction now passes all six steps: both fixture families
none/valid/expired 1/0/1, real scan PASS with the unchanged braces disposition,
18 distinct nonempty cidfiles and no owned leftovers. The real registry-enabled
lint/lint-base batch passes all 17 required IDs; cleanup finalization exit 0,
empty cleanup/leftover inventories. Failure-detail demo retains the original
assertions and enforces its bound. Exact evidence SHA-256 values:

- cid-run log: `71ecd5448174f9ee36fc9a7b1d466dc4f5ade988945145ade17b6c870714f315`
- cid-run results: `0baf00cb025f69b0bfc404261d84808bc0dad702b28fa8da574624aa3fbbd553`
- detail demo: `9aaa834b361f6f27548f0b07fe7252202f224b1b76259657a19022821d9ac03b`
- registry batch log: `4d6b54dd2fb84a8cce866137db0426aedad7e5753ba6251f9fe79ec63b9ff175`
- registry results: `4de2a096058784521d1bacf4e89e08b066f6da94b27966f936692066c64de53b`

Verified coverage union: 17 refreshed + 12 retained c2 + 75 retained final-01 =
104 required IDs, without omissions/additions. Unchanged evidence reuse remains
justified; no extra broad local suite. At the mandatory checkpoint, three/B5
attempts submitted, requirements implemented and concrete hosted-path defect
reproduced/fixed. Continue; two attempts remain, no extension or reset.

Coordinator is authorized to reconcile current target, publish the accepted
correction to existing MR !28 using an exact-tip lease, and collect a new fast
qualification. Preserve prior source refs; close the temporary c2review branch
only after its exact content is secured. Current fast and Duo disposition must
clear before auto-merge. A new code finding requires bounded disposition; do not
retry speculative failures or weaken expectations. Applicable staging disposition
and final scoped cleanup precede final Goal ACCEPT. After verified integration,
G013 provisional planning may start; its execution still needs final G012 ACCEPT
and its own release. G014 specification-only stop and G021 barrier remain.

## MR !28 fast / unavailable Duo disposition — 2026-10-15 UTC

**ACCEPT for conditional integration; final Goal ACCEPT pending.** GitLab reports
MR head `4305d11480ec5e29aa3298f06d43672a8dca20c8`, target `4094d71`, no conflicts.
Its non-management bytes match reviewed `0aac74e` exactly; the rebase delta is
only the three supervisor management records. Pipeline `2949373312` and fast job
`17190291257` SUCCESS; downloaded results independently confirm 35 PASS, zero
FAIL. B-DEPSCAN passes all six stages including both fixture families. Hosted
lint-base/dependency-scan stage reported 104 seconds; pipeline duration 543s,
within the twenty-minute fast cap. This supersedes the failed MR qualification
as current evidence while preserving its failure history.

Codex inspected all current MR discussions via glab: Duo sessions `9507542`
(started 04:39 UTC, old head) and `9508253` (05:08 UTC, current head) have posted
start notices but no result/findings by this disposition. The coordinator's
reported forty-minute waits do not establish their internal execution state.
There are no posted unresolved findings to dispose of. **Treat Duo as unavailable
for this MR**, not as completed, clean or comprehensive. Do not start a third
identical review or block indefinitely on missing service output.

The required actual-diff/security review is already complete through Codex and
independent bounded source/security reviewers, with both correction passes
reviewed and local/hosted evidence verified. This is a goal-specific supervisor
Duo disposition under the existing MR policy, not a new general bypass or an
owner-created waiver. It does not waive source/security review, tests or staging.
Any later posted blocking finding must be assessed before integration if still
pending, otherwise as a bounded staging correction before final acceptance.

Existing coordinator may reconcile the management-only target advance, preserve
non-management bytes and evidence provenance, recheck current MR discussions
and fast status, then enable auto-merge with the exact reconciled reviewed source
SHA. No broad local repeat for this management-only record. Record the unavailable
Duo disposition on the MR; do not say Duo approved. Verify actual merge/source
branch deletion and applicable staging pipeline, then return its evidence and
cleanup receipt for final Codex ACCEPT. G013 provisional planning may start after
verified integration, with execution held until final G012 ACCEPT/own release.
Budget remains three/B5 attempts; G014 specification-only stop remains.

## Staging evidence verified; closeout pending — 2026-10-15 UTC

Pipeline 2949510391 SUCCESS on integrated 5e1e987; jobs17191203943 fast35PASS and
17191203944 gate73PASS union104 required IDs with four intended overlaps,
no omissions/duplicates. Controls441/441; source-scan/depscan/rotation and inherited
audit checks PASS. One exact approved dependency exception remains visible,
zero unexcepted blockers. Source4305d11 to finalMR9c279861 differs only three
management files; MR/merge share tree8cbb5d2e74c395fe6d2212f42f23429603606a45.
Gate38m19s; pipeline46m30s. Independent evidence review used in-memory artifacts.
Gate results4782248930a503dfd0a941bbb672663cc5b557e7fe3691df2f5f0191516e4240;
CI controls ea307fe0e896144e721b1874aea18e82447fa81a78fcf008c7f68e1214ee396e.
Closeout Result/receipt and final Codex ACCEPT remain; coordinator was notified.

## Final implementation acceptance — 2026-10-15 UTC

**ACCEPT — G012 completed and integrated.** All ten ACs satisfied by the reviewed
source, attributable local/correction and hosted evidence, owner exception
conditions and scoped cleanup. Integration MR !28 merge
`5e1e98735f6224888889d516ef2397a0ea8dc89a`, exact reviewed tree
`8cbb5d2e74c395fe6d2212f42f23429603606a45`. Pipeline2949510391 SUCCESS:
fast17191203943 35PASS, gate17191203944 73PASS, union104 required IDs,
441/441 controls. Source/secret/dependency/rotation and inherited security checks
pass. Gate38m19s; pipeline46m30s. Prior failed runs, corrections, D0 deviation
and unavailable-Duo disposition remain historical, not retrospectively passing.

Imported report-only commit70661fb exactly; Result SHA-256:
`e9224e6197daa4d23ac594caaef6ab029bb1422f38a5e2887738030334b6f161`.
Imported coordinator cleanup receipt exactly at G012-closeout-receipt.md, SHA-256:
`f9964e9b6a7fcabfb5b29a67e5be4a6e9c5a8013e8493aaeea92aaa02eb64342`.
The Result's "Closeout owed" list describes the pre-cleanup submission; this
receipt and decision supersede it without rewriting Claude's submitted bytes.

Independent bounded closeout review verified both G012 worktrees, local branches,
owned focus resources/raw implementation and planning evidence, and four named
binding temporary files absent. Shared pinned Go/OSV images, an unproven shared
core-check image and remote plan/history remain explicitly held; unrelated/G013/
G054/native-locked work and unproven old resources were preserved. Remote temporary
closeout branch may be removed after exact import verification; receipt records
compact proof hashes, no necessary raw proof holds remain.

Three of B5 attempts used; checkpoint completed, no extension. The exact braces
3.0.3 exception remains an owner-accepted open risk through 2026-10-29 inclusive,
with earlier review/revocation conditions; never a fixed finding or automatic
renewal. Configuration-only encryption readiness does not qualify deployment.

G012 predecessor gate for G013 is satisfied. Existing G013 planner continues the
bounded r0 corrections; G013 execution remains held for its corrected Plan ACCEPT,
PG-005 owner disposition and current input/resource release. Do not start it solely
because this predecessor passed. G013 closes EPIC-004; perform PROGRESS compaction
there, not another validation campaign now. G014 specification-only human-input
stop, G021 product barrier and independent G054/G055/OS001 holds remain.
