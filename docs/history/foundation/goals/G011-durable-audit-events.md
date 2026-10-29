> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G011 — Durable audit and event conventions

Epic: EPIC-004 — Audit and compliance foundation  
Status: r0 + owner amendments — provisional planning released; implementation gated  
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation  
Date: 2026-10-14 UTC

## Objective

Make existing foundation mutations and security decisions durably attributable
through controlled audit records in the existing PostgreSQL database. Enforce
append-only runtime access and current-authority tenant reads without changing
the accepted identity, Organization/Group, policy or storage security contracts.

## In Scope

- ADR-006's dedicated `audit` schema, additive migration, ownership/grants/RLS,
  controlled append and bounded internal read interfaces owned by Go core.
- Hybrid recording preference: database triggers for committed row changes and
  Go-owned recording for denials and events outside row mutations. The Plan may
  justify a combined mutation/audit SQL function where it better preserves the
  same guarantees with no extra client round trip; no speed claim is assumed.
- A small core-internal audit-aware helper, reusing existing transaction/cleanup
  code for application-recorded atomic mutations; no external utility dependency
  or generalized transaction framework. Trigger-owned events are appended once.
- A finite event catalog and wiring for existing identity/session lifecycle,
  tenant/profile/membership/policy mutations, security denials/context changes
  and guarded storage issuance/mutations, with explicit attribution and outcomes.
- Successful, denied, failed, rolled-back and uncertain-outcome semantics;
  recording failures must not grant access or falsely report committed success.
- Real runtime-role, isolation, privacy, migration and focused failure evidence;
  minimum existing harness wiring and foundation operational documentation.
- Restricted maintenance boundary and G012/G013 handoff, not retention execution.

## Out of Scope

Product features, audit UI/export, public HTTP/BFF/MCP contracts, new tenant or
profile roles, administration/onboarding workflows, live WorkOS or R2 exercises,
queues/outbox/event bus, event sourcing, log ingestion, SIEM, cryptographic ledger,
backup/restore, retention jobs, legal hold, production deployment, new vendors,
generalized audit/test frameworks or a compliance certification claim.
No new policy of auditing every ordinary read, heartbeat or stream chunk.

## Dependencies

- **Final G010 implementation ACCEPT and committed integration before execution**,
  including its EPIC-003 disposition and any conformance corrections. G010 Plan
  r1/binding is accepted and implementation authorized, not finally accepted.
- Accepted G004–G009 interfaces, Results, deviations and residuals; ADR-003/004/006;
  canonical architecture, current governance and accepted validation/cleanup rules.
- [Work Spec](../specs/G011-durable-audit-events-spec.md). Reconcile the actual
  G010 source and current migration/catalog state at planning/execution binding;
  do not reserve a migration number or assume unchanged G010 source in advance.

Specification work is authorized. Claude Plan dispatch needs a separate release;
a provisional Plan, if later authorized, must label G010 assumptions. No G011
implementation is authorized by this Goal's creation.

## Completion Definition

Work Spec AC1–AC12 have attributable evidence. Codex accepts the actual diff,
event coverage and failure semantics; applicable integration/staging disposition,
resource cleanup and Result are complete. G012 becomes dependency-ready; G013
receives the maintenance/retention contract. G021 remains closed.

## Execution Policy

**Architecture/security/data tier.** Explicit Claude Plan and Codex Plan ACCEPT,
final accepted G010 reconciliation and Git-first execution binding precede work.
After release, implement and debug autonomously within reviewed paths and event
inventory. Preserve current authority/locking, optional MFA and UUID identity.
No audit outage may become an authorization bypass or reverse a completed
security revocation. Owner's 2026-10-14 ruling additionally requires completing
logout/disablement when only audit recording fails, with bounded failure recording
and an explicit gap diagnostic (Spec R4/AC5); ordinary mutations remain atomic.
Pin per-operation transaction/error semantics in the Plan.

Reuse existing fixtures, real PostgreSQL/runtime logins, fake identity provider,
private MinIO and accepted test entry points. Use focused failure batches before
required final selected coverage. Do not duplicate predecessor matrices, test
the unchanged harness as a separate project or run full once per event category.
Follow current MR/Duo/auto-merge, staged acceptance and bounded cleanup procedures.

## Iteration Budget

**B7: maximum seven shared Goal-level execution attempts; checkpoint after
attempt 4**, under BUDGET_POLICY (not included). New security-sensitive
schema/grants and transaction/failure wiring across existing foundation modules
justify B7. Planning spends none; delegates and inner fix/retest batches share
the allowance. Claude must forecast against the finite catalog and propose a
split before execution if it cannot complete within this boundary.

## Context / Scope Budget

One durable audit boundary over existing foundation operations. Enumerate current
producers and distinguish future hooks from implemented event coverage. Do not
turn operational logs into an ingestion pipeline or add every possible business
event. Reuse accepted catalog/pool/privilege controls; add only the audit delta.
Performance assessment is bounded to the new append/read/locking paths, linked
to PG-004; no throughput campaign, partitioning platform or speculative indexes.

## Suggested Subagents / Delegation

One implementation/integration owner; optionally one non-overlapping database/
security contributor and one independent read-only reviewer covering attribution,
transaction failure and runtime-role tests. One shared-stack owner. Planning
xhigh, implementation high; verify actual effort or report unknown. Follow task-
prefixed role paths and native-tool exception policy; no inherited G010 exception.

## Stop / Escalation Conditions

- Dependency or Plan/binding unavailable; architecture/ADR or expected-event
  semantics conflict; event coverage silently excludes an existing required flow.
- New authority/security model, product workflow, public API/MCP, infrastructure,
  vendor, queue, generalized framework or destructive data-model change is needed.
- Audit design requires runtime ownership/BYPASSRLS, arbitrary cross-tenant reads,
  fabricated actor/scope, secret persistence or unbounded metadata/queries.
- Failure policy could undo revocation, report rollback as success, repeat provider
  effects on ambiguous commit or promise atomicity with an external provider.
- PG-004 reveals materially new scans/fan-out, contention or feasibility risks:
  record and flag to the owner before proceeding with a materially changed design.
- Native permission refusal, uncertain destructive cleanup, repeated material
  nonconvergence or B7 exhaustion. Preserve evidence and return to Codex.

## Owner planning release — 2026-10-14 UTC

The owner explicitly authorizes: "if spec is finshed for 011 we can start 011
planning". This releases Claude Plan dispatch and supersedes the earlier planning
hold; implementation remains gated. The verified coordinator dispatches one
planner at xhigh in parallel with G010, using current committed staging and
recording its exact SHA, twelve-AC Spec and both audit mechanism/helper amendments.
Create `management/v0/plans/G011-durable-audit-events-plan.md`, with only necessary
linked planning history. No application/test writing, dependency install, build,
Docker run or validation campaign. Do not occupy G010's shared test environment.

The Plan is provisional on G010 final acceptance/diff and EPIC-003 disposition.
Reconcile actual final contracts/catalog/corrections before execution, alongside
explicit Codex Plan ACCEPT and Git-first binding. Keep the finite event inventory,
transaction/failure and trusted-attribution choices, trigger/function/Go mapping,
existing-helper reuse, query/lock assessment, AC mapping and expected validation
profile/runtime concrete. Reference accepted controls; do not rebuild them.
Surface real decisions and PG-004 costs without inventing numeric targets or
starting a benchmark. B7/checkpoint4 applies only to later approved execution.

Prefer `g011-agent-<id>-plan`. If the native isolated worker cannot select or
safely use that path, this task-specific record permits one G011 planner's native
`agent-<id>` path. Before writes record/verify task, agent, role, repository,
checkout, branch, actual effort or unknown, and management-only ownership. No
live rename, duplicate planner, permission bypass or future-role exception.
Safekeep the Plan branch and return exact commit/path/hash for Codex review.
G010 continues uninterrupted; no G011 implementation or G021 release follows.
