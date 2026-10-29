> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G064 — Integration-test isolation and per-test evidence

Status: Owner-prioritized definition; Claude planning authorized; implementation awaits Plan ACCEPT and execution release.
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation.
Authority: the owner's latest direct instruction prioritizes this task now with a Spec, Claude Plan and implementation, pausing G020. This authorizes G064 through normal review, qualification and closeout gates; it grants no new later-task specification range.
Spec: [G064 Work Spec](../specs/G064-integration-test-isolation-spec.md).

## Objective

Give each Go PostgreSQL integration package a private database cloned from a correctly migrated template, make independently intended tests runnable alone, and retain actual per-test outcomes through the existing validation harness.

## In Scope

- Test-only template provisioning, package database lifecycles and bounded fixture isolation/reset, including explicit exceptions for migration and recovery proofs.
- Inventory and remove accidental package/test order dependence while preserving deliberate stateful scenarios and their assertions.
- Existing Go JSON and Playwright reporter integration, sanitized per-test evidence retention, required-check parity and concise runbook updates.
- Focused correctness/security evidence, normal MR review/integration and one normal post-merge full local qualification.

## Out of Scope

G061 full-suite parallel orchestration; new scheduler/framework, Dagger, Bazel, vendor or build/cache platform; production schema, role/grant/configuration semantics or application features; live provider/infrastructure actions; host sizing changes; benchmarking campaigns; altered qualification/RC/CI policy or weaker security tests.

## Dependencies

Plan against current `staging` (definition baseline `63d493f25000a66b5c94256d1ea5886f6af1a964`), preserving accepted architecture/security and current harness contracts. The retention repair at `36d1189b` retains scanned local evidence under explicit `--evidence-dir`; it supplies no per-test output adapter.

The owner's latest 2026-10-21 direction prioritizes G064 now, superseding its earlier after-G020 scheduling. G020 is directed to pause at S1, preserving `origin/g020-impl` checkpoint `df2c46f4`, unmerged migration `00020`, its approved Plan and B5 attempt 1; confirm the coordinator's actual safe pause rather than infer acknowledgment. G020 acceptance is not a G064 prerequisite. G063's ongoing frozen qualification continues under its existing owner: do not rebase, restart, replace its source or alter its disposition. G063 final ACCEPT is not a G064 technical or scheduling gate. Respect shared-resource readiness without displacing either task's resources. G020 resumption requires G064 acceptance and current-input/readiness reconciliation; other RC7/main, preview, RC8/G057/G021 ordering remains separately governed.

## Completion Definition

Spec AC1–AC7 are supported by the actual diff, complete check/test inventory and attributable Result evidence. Focused pre-merge checks, supervisor/source and applicable Duo review, required hosted fast checks, authorized staging integration, coordinator-owned full local qualification on the exact merged staging SHA and scoped cleanup are complete. One attributable full run may satisfy multiple pending Goals only when they share the same reviewed candidate and all required bindings/coverage; do not require duplicate unchanged qualification. Codex records final ACCEPT. No mandatory manual M1 is added; the G021 foundation/product-feature barrier remains.

## Execution Policy

**Architecture/security/data tier:** explicit Claude Implementation Plan, Codex Plan ACCEPT, exact input/scope binding and supervisor execution release are required because the harness handles privileged test credentials, database lifecycles and security evidence. This Goal authorizes inspection and Plan artifacts now, not implementation, installs, builds or runtime probes before release. Claude owns concrete files/helpers and exception design; reuse existing primitives and do not modify production role/schema behavior to simplify tests.

After release, Claude may implement, validate and correct within reviewed scope autonomously. Preserve native permissions, task ownership, existing two-run DB/browser capacity, serialized shared image preparation and applicable locks. Use current model/effort policy and record actual settings or unknown. The retention repair's no-full-run exception does not apply to G064.

## Iteration Budget

**B5: five shared Goal-level execution attempts; checkpoint after attempt 3**, under BUDGET_POLICY (not included). Zero implementation attempts spent. Inner diagnosis/fix/test loops share the current attempt; no budget reset or silent extension. This bounded cross-package test-infrastructure change warrants B5; no security-model redesign is authorized.

## Context / Scope Budget

Current PostgreSQL integration packages and their existing harness/reporting consumers only. Plan inventories concrete package/test groups, reuse/extend/new coverage, exceptions and exact paths before implementation. **150k active-context checkpoint / 200k compaction-or-safe-handoff threshold**, subject to smaller native limits. Split/escalate if general orchestration or production changes become necessary.

## Suggested Subagents / Delegation

One Claude implementation owner/writer; optional up to two bounded read-only test-lifecycle/security reviewers, sharing B5. Keep one writer per artifact and no duplicate coordinator. Delegation does not add runtime slots or authorize separate campaigns.

## Stop / Escalation Conditions

Stop affected work for production role/schema/contract changes; weakened RLS/audit/runtime assertions or intentional fault/ordering semantics; unsafe template source/role binding or cross-task cleanup; shared-resource collision; unbounded fixture rewriting; unavailable required evidence; architecture/ADR conflict; native refusal; nonconvergence or exhausted budget. Preserve failures, unique work and active qualification bindings; return a bounded decision request.


## Owner-confirmed delivery order — 2026-10-21

The latest owner order is **G063 → G064 → RC7 → G020 → combined staging preview → RC8 (if needed) → G057 → G021 → RC9**. This supersedes earlier sequencing only. G064 planning and approved isolated implementation may overlap G063's frozen qualification; G063 final ACCEPT precedes G064 staging integration/final acceptance. Both must be accepted before RC7 is cut and fully qualified. G020 remains held through G064 final ACCEPT and RC7 qualification, then resumes its preserved checkpoint/budget after supervisor input/readiness reconciliation. Preview follows G020 acceptance; RC8 is conditional, and every RC actually cut requires its normal complete qualification. G057 and G021 follow before RC9. Existing main-promotion decisions, technical scope and budgets remain unchanged.

## Immediate owner-authorized extension — 2026-10-22

Fold [batch-owned Docker reuse](../specs/G064-batch-resource-reuse-spec-addendum-20261022.md)
into the current merged G064 remediation now, rather than opening another Goal
or finishing the current cycle with a full run first. Compatible invocations
share stack/runner resources; fresh state, security isolation, individual results
and direct individual reruns remain required. Claude owns an append-only Plan
addendum and exact scope/resource limits; bounded supervisor review releases the
new slice. Preserve correction-11 source, existing valid evidence and the current
shared B5 ledger; no automatic restart, extra attempt or full-suite campaign.
Pending G064 full runs are held/stopped safely and recorded; qualification is
consolidated after the combined repairs/reuse changes. Existing failed runs do
not become passes, and required final/hosted gates remain.

Cleanup removed disposable worker checkouts, not G020 source: all 17 recorded
G020 checkpoint tips were independently verified unchanged against the cleanup
inventory, with zero dirty G020 worktrees before removal. Recreate only a needed
inventoried worker checkout from the preserved refs. G020's existing authorization
and dependency holds are unchanged by this G064 extension.

## Final acceptance — 2026-10-23 UTC

**FINAL ACCEPT** after implementation MR !82, maintenance !84 and closing Result !85 integration. The [final review](../results/G064-final-acceptance-codex-review-20261023.md) records the owner-authorized pre-merge/focused continuation disposition, complete consolidated coverage, reconciled final tree, preserved failures and explicit resource/ref retention. This supersedes the initial status and original blanket post-merge full-run requirement above for this Goal's closeout. No mandatory manual preview or duplicate local run remains.
