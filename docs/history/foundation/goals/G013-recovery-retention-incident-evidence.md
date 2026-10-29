> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G013 — Recovery, retention and incident evidence

Epic: EPIC-004 — Audit and compliance foundation
Status: r0 — defined; G011 implementing, G012 specification available
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation
Date: 2026-10-14 UTC

## Objective

Demonstrate a recoverable, authorization-safe PostgreSQL/object-store checkpoint
and bounded maintenance/incident procedures for the existing generic foundation.
Prove the procedures on disposable synthetic environments and document actual
recovery evidence and remaining deployment responsibilities. A successful local
restore is not production backup qualification or a privacy-compliance claim.

## In Scope

- One operator-invoked paired database/object backup and isolated restore, with
  a documented consistency window, integrity checks and truthful partial failure.
- Verification of restored data, object bytes, roles/grants/RLS/audit protections,
  invalidation of restored authentication and replay of applicable post-checkpoint
  deletion/restriction instructions before ordinary access resumes.
- Bounded, separately credentialed retention for eligible audit records and
  confirmed-deleted object metadata; conservative repair of pending deletions.
- Backup expiry/hold procedures, one failure/alert/incident exercise, and a finite
  recovery/retention responsibility matrix using existing logs/health controls.
- Minimal reviewed maintenance migration/privileges and current harness wiring
  if required; compact security/operating documentation and successor handoffs.

## Out of Scope

Production provisioning/data access, live R2/WorkOS/cloud recovery, cloud/KMS
selection, high availability, replication/WAL/PITR, zero-downtime distributed
snapshots, generic backup/job/retention frameworks, automatic schedules/daemons,
queues, SIEM/pager services, new public HTTP/MCP/admin UI, tenant-wide account
erasure, arbitrary business-record deletion, subject-rights portal/legal policy
engine, universal retention periods or RPO/RTO guarantees, product features.

## Dependencies

- **Final G012 ACCEPT and committed integration before implementation**, including
  actual accepted G011 audit and G012 secret/vendor/encryption contracts. Current
  G011 Plan r2/deviation review and G012 Spec are provisional source assumptions
  for this definition, not accepted implementation evidence.
- Accepted G001–G010 store/configuration, identity, tenant policy/RLS and storage
  lifecycles; ADR-002/003/004/006, Architecture §21.1/§56.1 and current governance.
- [Work Spec](../specs/G013-recovery-retention-incident-evidence-spec.md).
  No production environment is required for the synthetic exercise. Actual
  production policies/providers/targets remain owner/deployment decisions.

## Completion Definition

Spec AC1–AC10 have source-bound evidence: paired backup/restore, real runtime-role
security checks, restricted maintenance/fault behavior, anti-resurrection replay,
failure/alert exercise and a named production-responsibility record. Codex accepts
the actual diff/Result and required staging disposition. Scoped resources/raw
backup material are cleaned; EPIC-004's PROGRESS compaction disposition is recorded.
G014 receives recovery/session/maintenance boundaries; G021 remains closed.

## Execution Policy

**Architecture/security/data tier.** A concise Claude Plan, explicit Codex Plan
ACCEPT and current Git-first input binding precede implementation. Standing
provisional planning becomes eligible after G012 integration; G012 final ACCEPT
is required for execution. This definition dispatches no planner or implementer.

After release, implement/debug autonomously within the approved finite scope.
Use existing container/runtime/security/cleanup tools and one owned synthetic
source/restore exercise. Destructive assertions operate only on proven-owned
fixtures; real source deployments, unrelated services and production data remain
untouched. Restricted maintenance is an explicit reviewed change, not permission
to disable RLS/audit protections or distribute privileged credentials to apps.

## Iteration Budget

**B7: maximum seven shared Goal-level execution attempts; checkpoint after
attempt 4**, per BUDGET_POLICY (not included). Focused failed-batch
fix/retest loops share the attempt. No budget reset across workers or interrupted
tests; no extra broad campaign just to obtain another clean report.

Planning defaults **xhigh**, implementation **high**. Remind the operator at
dispatch and record actual setting as verified or unknown. No attempt consumed
by this definition. Claude forecasts effort/validation in the Plan; narrow or
split scope before execution if a new platform becomes a prerequisite.

## Context / Scope Budget

Existing PostgreSQL and private object storage; one two-Organization synthetic
checkpoint; one restored target; finite audit/object maintenance cases and one
incident exercise. Redis restores empty. Reuse existing fixtures and permission
proofs; add checks only for recovery/maintenance guarantees changed here. No
separate orchestration, resource registry or full copied predecessor matrix.

## Suggested Subagents / Delegation

One implementation/integration owner; up to two bounded delegates for recovery
analysis and read-only security review. Do not split overlapping migration,
retention and storage ownership without a concrete integration boundary.
Workers share B7 and record task/role/actual native worktree mapping. Unsupported
Claude naming never blocks dispatch; permissions/isolation remain mandatory.

## Stop / Escalation Conditions

Unreconciled predecessor changes; inability to establish a consistent two-store
checkpoint; unsafe target/source alias or missing resource ownership; lost
replay/hold information; unexplained restore omissions; scope-changing migrations,
new roles/vendors/public contracts or a broader recovery platform; runtime audit/
RLS weakening; new query/index/fan-out/lock risks; native permission refusal;
budget exhaustion. Preserve the safe remainder and report the concrete blocker.

Any production-data operation or actual retention/notification policy requires
its separately recorded owner authority. Fixture policy choices are not legal
retention or launch approvals. G056 owns assembled-product rights readiness;
this Goal neither waits for it nor implements its product scope. No G014 Spec
or larger auto-spec mandate is authorized by defining G013.
