> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G066 — Live validation check progress

Status: owner-reduced minimal implementation authorized, 2026-10-23. Spec r2 supersedes the expanded Plan; current writer/release and B3 ledger continue.
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation/Result.

## Objective

Show each completed validation check's ID and status in the local or hosted job
console during execution, preserving confidential diagnostics, authoritative
results, exit behavior and resource cleanup.

## In Scope

- Expose existing foreground/background harness check events through the wrapper
  for fast and broader profiles, without waiting for overall run completion.
- Emit validated check IDs/statuses and optional generated timestamps; detailed
  messages remain in the existing scanned artifacts.
- Reuse current shell helpers, evidence paths and process cleanup; provide focused
  proofs and a short operator note describing informational progress output.

## Out of Scope

Harness redesign, new scheduler/dashboard/service, test-selection or qualification
changes, dependency upgrades, application behavior, resource-policy changes,
streaming raw logs, new monitoring agents, or RC tag/promotion actions.

## Dependencies

[G065](G065-neutral-ci-control-labels.md) final ACCEPT and integrated MR !87.
Planning may proceed provisionally from current source without changing its frozen
qualification checkout. Reconcile integrated G065 before implementation release.
G066 runs immediately after G065, before the next RC tag and before G020
integration/final acceptance. Existing authorized isolated G020 implementation
may continue. The G021 foundation barrier remains binding.

## Completion Definition

The [Spec](../specs/G066-live-validation-progress-spec.md) AC1–AC6 have attributable
focused/fast evidence; current source/review/hosted-fast gates clear; affected
post-merge verification and cleanup complete; Codex accepts the diff and Result.
No manual owner walkthrough or full-suite qualification is required for this
bounded output change.

## Execution Policy

Ordinary execution with a concise Claude Plan and Codex Plan ACCEPT before code.
Owner direction: “make a spec for 066 and schedule that immediately after 065
ahead of another rc tag.” Definition and scheduling are authorized; normal
provisional planning and execution-release rules apply. Planning can start now;
implementation awaits G065 final ACCEPT, Plan ACCEPT and current-input/resource
reconciliation. Use current phase/model defaults; verify actual settings or
report unknown. Once released, execute and correct autonomously within scope.

## Iteration Budget

B3, checkpoint after attempt 2, zero attempts spent. One shared allowance for
implementation, focused validation and corrections under the
budget policy (not included). No reset through worker replacement.

## Context / Scope Budget

One wrapper/harness output path, its existing focused test locations and operator
documentation. No new framework, daemon, runtime or Docker stack for progress
transport. Plan names exact files/selectors, maps all ACs, and separates coding,
validation and waiting estimates. Escalate if this requires wider lifecycle redesign.

## Suggested Subagents / Delegation

One Claude implementation owner and the existing coordinator. Optional bounded
independent output-confidentiality/process review; no competing writers, additional
monitor or duplicate qualification run. Workers share the Goal's attempt allowance.

## Stop / Escalation Conditions

Raw diagnostics/secrets becoming console-visible; altered test/exit semantics;
unbounded buffering or process/resource leaks; needed scheduler, recovery or gate
changes; new dependencies; unclear ownership; predecessor hold, native refusal
or exhausted budget. Preserve actual failures and evidence.

## History

- 2026-10-23 — Owner authorized G066 definition and immediate placement after
  G065, before another RC. Supersedes its proposal-only backlog status.


## Owner RC milestone — 2026-10-23

Immediately after G066 final ACCEPT, cut `0.2.0-rc.9` from the latest observed
staging tip using the existing RC workflow and required full hosted qualification.
The owner also schedules `0.2.0-rc.10` immediately after G020 final ACCEPT; the
combined preview follows that cut, and the later post-G021 candidate becomes
RC11. This supersedes the prior unspecified next-RC placement, without changing
G066 implementation/validation scope or authorizing main promotion/final release.


## Corrected Plan acceptance and execution release — 2026-10-23

The corrected Plan acceptance (not included)
imports `cf34795b` and releases one Claude implementation owner after routine S0
reconciliation. G065 final ACCEPT has cleared the predecessor gate. Earlier
Plan-wait wording records the prior phase; scope/Spec and B3 are unchanged, with
zero implementation attempts at release. Follow the approved early/focused proofs,
current hosted fast and affected merged-SHA verification, then final ACCEPT/RC9.


## Owner minimal-scope override — 2026-10-23

Current [Spec r2](../specs/G066-live-validation-progress-spec.md) supersedes the
broader Objective/In Scope/Completion/Execution details and expanded Plan above:
one live `completed/total PASS|FAIL` line after each existing recorded test/check.
The owner targets 10–30 minutes of implementation and selects hosted fast checks
alone to qualify the shared fast/integration logging path. No new control matrix,
reader-cap/identity/failure-marker project, local/full/integration qualification
or separate post-merge run is required. Preserve existing safety/result behavior.

Continue the current implementation owner without a new planning cycle; reconcile
its small actual method in the Plan/Result. B3 and the current attempt are not
reset. Source review, successful current hosted fast, ordinary staging integration,
brief Result/cleanup and Codex final ACCEPT remain. RC9 follows immediately.
Earlier dated Goal/Plan/review detail remains history; it does not reimpose the
retired scope or validation gates.
