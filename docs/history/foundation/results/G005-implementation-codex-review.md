> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G005 implementation — Codex review

Date: 2026-10-09 UTC
Outcome: **ACCEPT** — implemented and integrated.

Reviewed the actual integrated diff, Goal/Spec, approved Plan r4 and review, Result, local provenance/failure evidence and final hosted artifacts. Integrated commit `570b40caf47ff9257c5b5aa0588f60fecd58d0a6` has tree `a09895950c8077f7f50b116d05edeb76c81c657e`, identical to submitted source `f862f1843fb96636d8801e6f0ae751bfa3a9db5c`. Subsequent staging changes are management-only; accepted implementation content is unchanged.

The final pre-integration Result from local commit `0e32d938660903c55819d4da2cd66bf5a434e936` is imported byte-identically, SHA-256 `510314a4dbd929efdc98cfa0818df785951a6b2104dc132d297c144572d9008d`. Its earlier pending/open wording is historical and superseded by this completion record.

## Final hosted evidence independently verified

Staging pipeline 2931683534 (not included), commit/tree above:

- Fast job `17068999910`: success, 33/33 required checks, 368.022 seconds.
- Gate job `17068999911`: success, 81/81 required checks, 3902.628 seconds.
- GitLab test report: 114 successful outcomes, zero failed/error/skipped. Per-job repeated checks are not 114 distinct obligations.
- The gate's PASS set equals the full registry exactly, with no missing or extra ID. B-TENANCY-CATALOG reports 21/21 subtests; B-TENANCY-MIGRATION 6/6 plus persistence restart; B-TENANCY-IT 35/35 and log validation/planted-leak detection; B-TENANCY-HTTP 36 absence probes.
- Inherited root/off browser suites each report 141 tests, identity integration 57 and identity browser 56; 282 CI controls pass.
- Both final job traces record `final status 0`, all final phases zero, registry removed and no harness interruption. Artifact summaries are explicitly pre-publication; final job status/trace provides the later authoritative outcome.

Read-only evidence is retained outside the repository at `<home>/.local/state/platform-evidence/g005-codex-acceptance/`. Artifact archive SHA-256 values: fast `f256de628b2dda27ca2a7196923d57382bd941b6b8483c3bd09cd4d8d1c3a2a1`; gate `73a864cf1e37df06321c42c48c1fa6312f6a210944f0623f2d3040b96cbdf053`.

## Criterion disposition

| Criteria | Reviewed implementation and evidence |
|---|---|
| AC1–AC3 | Database-generated immutable UUIDs; Organization/Group composite integrity; closed role/source/scope CHECK; one membership per subject/scope. Scoped service SQL and runtime/catalog/concurrency tests preserve those constraints. |
| AC4–AC5 | RoleBinding is the sole writable authority; SELECT-only security-invoker membership projections. Role changes/removal and multi-scope identity tests preserve unrelated bindings and identity/session data. No implicit Organization membership. |
| AC6–AC7 | Organization-local slug uniqueness/reservations, immutable canonical IDs, scoped predicates and atomic Group-plus-creator-owner transaction. Exact conflict/fault/creator-deletion/concurrency assertions and rollback evidence are present. |
| AC8–AC9 | No binary links the tenancy primitives and candidate routes remain absent. Additive migration, exact runtime column grants, owner-only functions/metadata, restart/down/reapply and predecessor preservation are tested. G001/G004 migration bytes remain unchanged. |
| AC10–AC11 | Required inherited, privacy and variant coverage passes locally and hosted. Forced F-1 model failure produces B-TENANCY-IT/B-REQUIRED-CHECKS failure and exit 1. Hosted final content and all model/migration IDs are bound above. |
| AC12 | Approved-input reconciliation, BASE-aware local scope/protected checks, pre/post verification, owned cleanup, failed-run history and Result are attributable. B7 used one meaningful correction cycle; cycle-4 checkpoint was not reached. |

## Deviations and residuals

Accept the recorded checkout relocation, owner-directed validation/MR placement, approved ESC-1 option B, scoped identity-fixture warning and lock-wait mechanism. Step batching does not omit required final coverage. No architecture, provider/security-model, public contract or product-scope deviation is introduced.

Retain the documented residuals: no post-G005 live WorkOS qualification was required under option B; a future authorized qualification tests the updated fixture. PostgreSQL constraint DETAIL can contain UUIDs/slugs, while application operation/error logs remain bounded; the canary evidence is not a claim of redaction of every transformed display value. Lost commit replies may return unavailable after a durable commit. Those limits are documented, not hidden successes or new authorization.

G005 is accepted and integrated. G006's predecessor acceptance gate is satisfied. Its approved Plan r2/B5 still requires fresh baseline/input/scope binding, action authorization and S0 before execution; no application work is performed by this review. G007 policy, G008 RLS and G021 product-barrier acceptance remain separate.
