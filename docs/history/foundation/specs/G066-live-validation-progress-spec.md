> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G066 Work Spec — Simple live test progress

Revision: r2, 2026-10-23 — owner reduction; r1 (not included) preserved.
Goal: [G066](../goals/G066-live-validation-progress.md).

## Required outcome

After each existing recorded test/check finishes, show one line in the live job
console, for example `27/67 PASS` or `28/67 FAIL`. The user can see how far the
run has progressed and notice a failure before the final job result.

## Requirements

1. Use the existing shared test/check completion logging path used by hosted fast
   and integration jobs. Show completed count, total selected for that job/run,
   and that result's PASS/FAIL status. Derive the total from existing selection;
   67 is an example, not a constant. Informational/start/working messages do not
   increment completed tests. Reuse the runner's existing result granularity;
   no new instrumentation of every service framework's nested subtest.
2. Emit during execution, not only in the end-of-run report. Use actual recorded
   outcomes; do not turn missing/unexecuted tests into passes or hide a failure
   behind a later passing line. The existing final result/exit status remains
   authoritative. Raw diagnostic messages, credentials and command payloads stay
   out of this numeric/status progress line.
3. Keep this a minimal addition to existing logging. Preserve existing wait,
   deadline, signal, cleanup, scanning and result behavior. No new progress
   framework, scheduler, reporting system or resource-policy work. The previously
   planned reader caps/identity checks/failure marker and VG1–VG3 control expansion
   are superseded as G066 deliverables; existing safety controls are retained.

## Acceptance and qualification

| ID | Outcome | Verification |
|---|---|---|
| AC1 | Live completed/selected-total and PASS/FAIL after recorded test/check completion. | Current hosted fast job console demonstrates the shared path while execution continues; inspect count/total/status against its existing results. |
| AC2 | The small diff preserves existing results, failure/exit and confidentiality behavior. | Source review and existing hosted fast checks; use a small assertion in existing fast coverage only if needed to verify formatting/FAIL handling. |
| AC3 | Change is integrated and attributable, with ordinary cleanup and a brief Result. | Reviewed source, successful current hosted fast, normal staging merge/tree verification and cleanup receipt, then Codex final ACCEPT. |

Hosted **fast checks only** qualify this shared logging change. No local/full
suite, integration-job run, additional CI-control matrix, dedicated proof campaign,
or separate affected post-merge run is required for G066. This is the owner's
specific exception for this small output change; later RC full hosted qualification
remains required. Preserve native/source-review gates and actual failures.

## Execution and scope

Implementation target **10–30 minutes**, excluding hosted execution/queue time.
Reuse the current writer and release; record the small method/scope change in the
existing Plan/Result without another planning/approval cycle. If it cannot stay
small, report the concrete obstacle promptly instead of expanding the task.
No budget reset: current B3/attempt ledger remains. Existing source review and
staging merge authority apply. RC9 follows final G066 ACCEPT.

## Authority and supersession

Owner, 2026-10-23: “after individual tests ... 27/67 status pass|fail ... that's all
I want”, followed by “Impl should only need 10-30 min” and the shared-path/hosted-
fast-only qualification direction. This supersedes r1's additional guarantees,
AC matrix, the expanded Plan and its correction/release detail wherever they
require more than this minimal behavior. Original artifacts remain historical.
G065 is accepted; no application/authentication/tenancy, dependencies, CI triggers
or broader harness behavior changes are authorized.
