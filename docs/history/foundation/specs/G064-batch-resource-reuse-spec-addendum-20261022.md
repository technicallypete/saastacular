> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G064 addendum — batch-owned Docker infrastructure and individual test evidence

Date: 2026-10-22. Owner-authorized extension of the existing merged, still-open
G064 remediation. Read with the [original Spec](G064-integration-test-isolation-spec.md)
and continuous cleanup mandate (not included).
This is part of G064, not a new Goal, a G061 parallel orchestrator or a restart
of the accepted isolation/reporting architecture. Preserve original artifacts,
review history, failed runs and the current Goal-level attempt ledger.

## Immediate sequencing and authority

The owner directs this work to begin now, before completing the current repair
cycle with another full suite. Claude immediately prepares an append-only Plan
addendum and reconciles the preserved correction-11 source with current staging.
Codex reviews that bounded delta and releases its implementation. Original Plan
bytes and accepted history stay unchanged; this direction is not Plan ACCEPT.
Source-only planning/development may overlap safe focused work on the existing
cleanup repairs. The repairs remain part of the combined G064 delivery.

Do not run a full profile to finish correction 11 and then another full profile
for batching. Defer pending full runs; safely stop any currently running G064
full run under the owner's new direction, preserve attributable partial/failure
records, and stop/reap/clean its resources. Reuse unchanged source and evidence;
throwing away useful work or resetting the budget is unnecessary.

After the combined changes, use focused affected-case/batch proofs, source review,
required fast/Duo gates and integration before one consolidated final qualification
run if required. Existing G064 full runs remain FAILED; this is not a waiver or
a claim of a passing full run. Do not launch a second broad/full attempt after a
failure automatically: retain evidence, diagnose, and apply the standing focused
rerun/coverage disposition. Required hosted RC qualification is unchanged.

## BR1 — Reuse compatible infrastructure per batch

Use a batch-owned PostgreSQL/MinIO/Redis/application stack for compatible test
invocations rather than repeatedly creating equivalent stacks per test. Reuse a
package/batch runner container where the image, bound source, execution identity,
environment, networking and writable fixture boundaries are demonstrably compatible.
Images are built once per bound source/configuration as needed, with serialized
builds; no per-test rebuild or new scheduler/pooling platform.

Claude's Plan maps existing concrete suites/invocations to batches, exact reused
resources and teardown points. Start only the services that batch actually needs.
No cross-task or cross-source reuse, shared mutable template, weakened credential
isolation, broad new privileges or production schema/role changes. Keep the
existing two-run private DB/browser capacity and shared-resource constraints.

## BR2 — Preserve independent state and infrastructure scenarios

Infrastructure reuse must preserve G064's fresh database lifetime for each
independent invocation, verified closed templates, complete cluster-global
role/settings residue checks, pool/session closure and direct clean-start selectors.
Keep object-store fixtures, writable paths, browser contexts/cookies and identities
isolated/reset at their reviewed boundaries. If state is unsafe after a failure,
hold/reset or tear down the batch; never pass contaminated state to another test.
Do not replace independent tests with an order-dependent aggregate batch.

Retain disposable scenario resources where container lifecycle, cancellation,
startup configuration, provisioning, restore/recovery or another infrastructure
mutation is the assertion. CI controls are examples; a suite being large is not
itself an exception. List the concrete exception and technical reason in the Plan.
Deliberate seed/transition/verify and transaction/ordering chains remain intact.

## BR3 — Individual failure reporting and reruns remain mandatory

Every selected test/subtest keeps its own outcome, timing, package/name, source,
selector, attributable sanitized log and existing supplemental JUnit/TSV evidence.
Expected-test completeness and child exit-status propagation remain fail-closed:
one failed case, process/setup error, timeout, missing outcome or truncated stream
cannot become a passing batch. A persistent runner must not hide a failed child.

Keep supported individual/direct selectors runnable alone from clean prerequisites
and fresh state, without executing the whole batch. A focused rerun retains the
original failure and reports its own exact source/command/outcome. Preserve bounded
per-test diagnostics even when the runner/stack is shared; reuse infrastructure,
not a batch-level summary in place of individual failure details. This supplements,
and does not replace, the authoritative required-check/completeness ledger.

## BR4 — Resource budgets and continuous cleanup

Record batch/scenario ownership and exact IDs, including anonymous volumes and
partially created/nested resources, in the existing inventory at allocation.
Remove scenario resources at case termination; remove reused stack/runner resources
at batch termination. Cover success, test/setup failure, timeout, cancellation and
reboot recovery. Stop/reap creators before sweeping, verify absence afterwards,
and fail/hold dispatch on any abandoned/unaccounted resources. Finished worker
worktrees/scratch are reclaimed at handback after compact source preservation.

Plan explicit maximum simultaneous project-owned containers, networks and volumes
for ordinary batches and every justified scenario exception. About 10–20 live
containers is an initial ordinary-batch planning estimate, not a verified count
or an allowance to exceed the reviewed limits. Derive network/volume counts from
the actual Compose definitions and approved overlapping batches. Separately report
cumulative creations, peak live resources and terminal leftovers: hundreds of
short-lived starts are different from hundreds left alive. Image/cache counts
are separate from running memory use. Other projects' resources stay protected.

Use counts/timings from the necessary focused proofs and any already retained
comparable evidence; do not run a separate old/new full-suite benchmark. Report
unknowns honestly, and verify zero unaccounted disposable project resources at
terminal cleanup. Do not promise that a trap executes during a power loss.

## Additional acceptance evidence

| Criterion | Required proof |
|---|---|
| BR-AC1 | Reviewed batch/resource/exception matrix, exact touched paths, source bindings, maximum live resource limits and retained per-case teardown exceptions. Show actual reuse, not just renamed tests. |
| BR-AC2 | At least two representative independent invocations reuse the same stack/runner where approved while receiving fresh state; a marker/fault cannot leak across invocations. Reversed-order and direct single-selector proofs cover each changed isolation mechanism, reusing unchanged prior evidence where valid. |
| BR-AC3 | A deliberately failing test/subtest in a reused runner yields its individual failure/log and nonzero batch outcome. The isolated selector can be rerun without the whole batch; original failure and focused rerun remain separately attributable. Setup failure, timeout/interruption, missing/truncated output and secret-safe diagnostics retain their guarantees. |
| BR-AC4 | Actual resource receipts cover ordinary reuse and nested/scenario cancellation, late recreation and restart recovery. All resource categories are accounted for, disposable leftovers are zero, and cleanup failures prevent further affected dispatch. |
| BR-AC5 | Result reports observed peak-live/cumulative/final counts and required focused validations. No full run occurs between the old repairs and this addition; final qualification is consolidated under the sequencing above. |

Existing AC1–AC7 and security/architecture invariants remain. Add no new required
check IDs unless the Plan demonstrates a gap and obtains the usual parity review.
Use existing harness/cleanup/reporting mechanisms. Claude owns the concrete Plan
and any allow-list amendment; propose precise additional patterns before touching
out-of-list files. Runtime repair, resource reuse and reporter conformance are
reviewed as attributable deltas, without reopening unrelated application work.
