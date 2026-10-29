> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G009 — Tenant-aware object storage boundary

Epic: EPIC-003 — Authorization and isolation
Status: r1 — Plan r1 ACCEPT; execution binding confirmed, implementation released after preflight
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation
Date: 2026-10-13 UTC

## Objective

Provide core-owned, tenant-authorized binary storage with UUID metadata in
PostgreSQL, scoped local MinIO access and bounded signed downloads. An object
key or shared BusinessProfile must never substitute for exact-object authority.

## In Scope

- Owned ObjectStore/provider adapter and guarded Put/Get/Delete/SignedGetURL/Metadata operations.
- Minimal Organization/optional Group metadata, immutable UUID/key/scope, RLS and additive migration.
- Separate read/upload/delete action mapping, current Group-member uploads and
  own-upload deletion, fresh identity/context checks and safe operation ordering.
- Bounded payloads, exact signed GETs, failure/cancellation/ambiguous-outcome recovery.
- Real runtime-role/MinIO isolation evidence, scoped credentials, operational issuance logs and docs.
- Minimum covering harness/CI changes, preservation of root/off/rename behavior and cleanup.

## Out of Scope

Product attachments/resumes/conversations/exports or UI; public HTTP/BFF/MCP
routes; signed uploads, multipart/resumable transfer, object replacement/versioning,
CDN/public buckets, antivirus/DLP/workflow processing; BusinessProfile authority
store; identity redesign; live R2/cloud provisioning; durable AuditEvent/retention
(G011/G013); generalized repair daemon/queue; CI redesign or unrelated refactoring.
No G010 completion claim or G021 product-barrier release.

## Dependencies

- **G008 final implementation ACCEPT and committed integration**, transitively G004–G007.
  G008 is implemented and merged at `81b186c`; correction source review is ACCEPT.
  Final G008 ACCEPT is recorded on passing staging pipeline `2945531965`;
  Plan correction/approval and fresh execution binding remain.
- Actual accepted G008 schema/grants/helpers/context/COMMIT/pool boundary must be
  reconciled before G009 Plan approval and execution binding. Do not guess migration
  numbers or treat its planned version 6 as accepted catalog evidence.
- Existing private MinIO bucket/scoped core credential, pinned SDK, current
  governance/ADRs/roadmap and [Work Spec](../specs/G009-tenant-object-storage-spec.md).
- Owner-approved Spec R3: members upload within authorized Groups and delete
  their own uploads while that access remains; others' uploads and Organization
  mutations retain owner/admin rules. Disposition (not included).
  G054 testing is independent, not a dependency.

## Completion Definition

AC1–AC14 have attributable evidence; required local/hosted disposition and actual
source review satisfy current governance; Codex marks final ACCEPT. Deliver
explicit G010 isolation and G011 audit handoffs. No storage guarantee is deferred
to G010 and no product workflow or public release is authorized.

## Execution Policy

**Architecture/security/data tier.** This draft authorizes specification work
only. Separate planning authorization, a Claude Plan, explicit Codex Plan ACCEPT,
dependency reconciliation and fresh committed execution binding precede implementation.
Reuse accepted controls, rebase/batched testing, manual-before-final ordering,
scoped cleanup and MR-to-staging workflow. No new execution framework or live
cloud account is necessary. Preserve concurrent G008/G054 checkout/input bindings.

## Iteration Budget

**B5: maximum five shared Goal-level execution attempts, checkpoint after
attempt 3**, under BUDGET_POLICY. Forecast 2–4; planning spends none. All inner
loops/delegates share the allowance. Claude must propose a bounded split if the
actual reconciliation makes B5 infeasible; no autonomous extension or weaker tests.

## Context / Scope Budget

One reusable storage increment: metadata, adapter, guarded operations and their
fault/isolation evidence. No product object relationships, general storage engine,
background orchestration or new authorization framework. Internal lifecycle fields
are permitted only to implement this Goal's durable failure-state guarantees.

## Suggested Subagents / Delegation

One implementation/integration owner; optional non-overlapping adapter/metadata
worker and read-only security reviewer. One shared stack owner, no competing
writers. Planning xhigh, implementation high unless owner chooses otherwise;
verify actual effort or report unknown. Reviewers cannot grant acceptance.

## Stop / Escalation Conditions

- G008 incomplete or bound source drift unreconciled.
- Own-upload deletion attempts to bypass current Group access, scope/context,
  immutable verified creator attribution or the separate action matrix.
- Changed capability semantics, public contracts, architecture/security model or scope.
- Need runtime owner/BYPASSRLS, broad definer/admin flag or object-store root credentials.
- Unbounded stale authority, incomplete metadata made accessible, or destructive
  compensation after an uncertain commit cannot be prevented within scope.
- Risky migration/data deletion, unsupported provider semantics, credential leakage,
  unbounded stream/network/lock lifetime or unavailable required evidence.
- Unexplained query/index/lock behavior under the performance-register rule.
- Budget exhaustion, repeated nonconvergence or a broader cloud/repair framework.

## Revision history

- r0, 2026-10-13: initial draft; exact archive (not included).
- r1, 2026-10-13: owner resolves Group-member upload and scoped own-upload deletion.
  B5/checkpoint 3, fourteen criteria, G008 acceptance and G021 remain unchanged.
  No planning dispatch or implementation is authorized by this policy revision.

## Conditional planning authorization — 2026-10-14 UTC

Owner authorizes G009 planning once G008 MR !23's applicable checks pass
and Duo findings are resolved or explicitly dispositioned. This permits a
provisional Claude Plan using the exact reviewed G008 candidate before final
G008 acceptance; it supersedes specification-only authorization for planning
only when those conditions are verified by Codex. Record the provisional
source commit and reconcile any subsequent G008 changes. G008 final ACCEPT
and committed integration, actual-state reconciliation and explicit G009
Plan ACCEPT/execution binding remain prerequisites to implementation.
B5/checkpoint 3, fourteen ACs, scope and G021 remain unchanged.

## Owner-authorized G009 delivery loop — 2026-10-14 UTC

Owner authorizes the supervisor/coordinator to proceed autonomously through
G009 planning, Codex Plan review, Claude implementation, validation and bounded
correction/review loops once G008 is merged and its existing final acceptance/
dependency gates are satisfied. No additional owner start prompt is required.
The earlier conditional provisional-planning permission remains; it does not
release implementation before accepted, committed G008 integration.

Codex still reviews and explicitly accepts the Plan before Claude implements.
Use the actual accepted G008 state and current Spec, B5 shared attempts/checkpoint
3, fourteen ACs and risk/escalation rules. Scope, security-model, architecture or
budget changes still return for the required decision. Opening the reviewed MR
follows existing authorization; this instruction does not grant a G009 merge,
release, cloud provisioning, weaker evidence or G021 product-feature exception.

## Provisional planning released — 2026-10-14 UTC

Codex verified passing MR fast checks and disposed Duo findings for G008
`d0e8d40`, with correction/source review ACCEPT. MR !23 merged as `81b186c`.
The earlier owner-authorized provisional-planning conditions are satisfied.
Claude may draft G009's Plan at xhigh using this integrated source and current
Spec, explicitly recording final G008 acceptance as an implementation gate.
Reconcile any later accepted source changes before Plan approval/execution.
No G009 implementation starts from this planning release alone.
