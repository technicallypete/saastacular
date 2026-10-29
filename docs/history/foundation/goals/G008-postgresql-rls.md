> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G008 — PostgreSQL RLS and least-privilege roles

Epic: EPIC-003 — Authorization and isolation
Status: r3 — Accepted-predecessor reconciliation complete; specification only
Owner: Codex (Goal/Spec/review); Claude Code (Plan/implementation/validation)

## Objective

Enforce the accepted tenant/scope policy in PostgreSQL as defense in depth, protect global identity visibility, and integrate transaction-local trusted context into the existing core database boundary. Prove isolation and context cleanup through the actual runtime role, including reuse of the same physical connection.

## In Scope

- Exact classification of accepted application relations and privileges; ENABLE/FORCE RLS and command-specific policies on every existing tenant-owned table and the global identity table.
- Organization, Group, BusinessProfile, association, RoleBinding and Organization-policy isolation consistent with accepted G007 semantics; read-only invoker membership projections.
- Core-owned transaction/context integration, fresh authorization and safe pooled-connection lifecycle; bounded authentication and authority-reader paths needed to avoid circular policy resolution.
- Additive migrations, least-privilege role/helper controls, database integrity and real PostgreSQL isolation/concurrency tests.
- Focused documentation, a portable role-provisioning procedure for environments without Docker init scripts, minimum existing harness/CI changes and final evidence.

## Out of Scope

- New tenant/domain entities, permission semantics, authority stores, product features or foundation-barrier release.
- Public HTTP/BFF/MCP contracts, tenant switching/discovery, domains, onboarding, invitations, member-administration UI, ownership transfer, billing or MFA policy changes.
- Object storage (G009), complete integrated attack-suite acceptance (G010), durable audit/retention (G011/G013), production deployment/cloud provisioning or provider changes.
- New pooling service, generic policy engine, orchestration/CI redesign, benchmark campaign, unrelated refactoring or changes to another active task checkout/input binding.

## Dependencies

- **G007 implementation ACCEPT and committed integration**, transitively G004–G006 and accepted bootstrap/observability/CI/UI foundations. Provisional designs do not satisfy this gate.
- Canonical architecture, accepted ADRs, AGENTS/CLAUDE, BUDGET_POLICY, ROADMAP/PROGRESS, the [Work Spec](../specs/G008-postgresql-rls-spec.md) and actual predecessor Results/reviews/deviations.
- Reconcile the actual schema, queries, session validation, guarded operations, capability matrix and authorization ordering before execution. Bind a fresh committed baseline, approved inputs and exact scope using accepted procedures.
- Claude's reviewed Plan must settle relation/command permissions, policy-reader and pre-tenant authentication boundaries, migration/rollback safety and connection cleanup. Named hosted actions require their applicable authorization.

## Completion Definition

1. Work Spec AC1–AC14 pass on the final submitted tree with attributable runtime-role, migration, local and hosted evidence; no required evidence is skipped.
2. Result records exact policies/grants/helpers, protected/exempt relations, context lifecycle, trust limitations, G009/G010 handoff, failures/deviations and shared B7 accounting.
3. Codex reviews actual diffs/evidence and returns ACCEPT; Progress is updated. G009 then becomes dependency-ready, without declaring G010 or G021 complete.

## Execution Policy

**Architecture/security/data tier:** separate Claude Plan and explicit Codex approval are required. This reconciliation authorizes specification work only, not Claude dispatch, implementation, provider/CI actions or changes to active task inputs. G004–G007 predecessor gates are satisfied; a separate Claude planning handoff and Plan ACCEPT with a fresh execution binding are still required. G054 manual testing is independent.

After approval, Claude may implement, test and diagnose within that binding. Reuse accepted wrappers, isolated fixtures, cleanup and CI. Diagnose failed SQL/pool cases with focused reproduction before required final regression; do not rebuild unchanged execution controls or repeat unrelated provider qualification. No developer/production destructive SQL is authorized.

## Iteration Budget

**B7: maximum seven shared Goal-level execution attempts**, checkpoint after attempt 4, under BUDGET_POLICY. This r3 explicitly reconciles the unstarted Goal to the current attempt-based policy; inner debug loops and subagents share the allowance. RLS, authority lookup recursion, authentication integration and pool lifecycle are coupled security-sensitive work; B5 would understate the required evidence. Claude forecasts attempts and proposes a split before execution if B7 is insufficient. No reset, autonomous extension or weaker tests. Drafting consumes zero attempts.

## Context / Scope Budget

One bounded database-enforcement increment over the accepted foundation. Include only existing relations and the minimum database helpers/context integration needed for their protection. A broader identity redesign, authorization engine, provisioning system or new deployment platform is additional scope and returns to Codex. The Plan pins exact files and schema/privilege changes rather than guessing future migration numbers or tables.

## Suggested Subagents / Delegation

Claude may use one bounded SQL/policy implementation delegate and one read-only security/test reviewer, with non-overlapping file ownership and one integration owner. All corrective work shares B7. No delegated architecture/product decisions or competing stack runs. Planning defaults xhigh; implementation defaults high, with owner-selected xhigh for complex security/ordering work. Verify actual effort or report unknown; delegation does not multiply B7.

## Stop / Escalation Conditions

- Conflict with architecture, accepted ADRs, G007 capabilities/ordering, global UUID/session authority, projection rules or the G021 barrier.
- A required operation needs superuser/BYPASSRLS, runtime ownership/privileged membership, a general definer bypass or client-controlled privilege/context flags.
- Pre-tenant authentication or authority resolution cannot be protected without broad predecessor changes; policy recursion or revocation races cannot satisfy the Spec.
- Destructive/risky migration, weakened identity/privacy guarantees, schema/public-contract expansion, new major vendor/pooler or deployment work.
- An incomplete relation inventory, unattested runtime connection, failing negative control, unavailable required evidence or unsafe cleanup cannot be resolved within scope.
- B7 exhaustion, repeated non-convergence or context/scope overflow. Preserve state and report the exact remaining gap; no green skip or deferral of G008 criteria to G010.

## Owner revision — 2026-10-08

Follow ADR-004 and G007 r2 for Group owner and derived profile access without an independent BE authority store. Secure the current role/attachment joins or read-only projection; removal of the last qualifying path cannot retain profile-management authority. r1 Goal (not included) is preserved. Fourteen criteria, B7 and existing dependency/implementation/G021 gates remain.

## Owner naming clarification — 2026-10-08

Use canonical `BusinessProfile` (UI: **Business Profile**) instead of `BusinessEntity`; the shared association becomes `GroupBusinessProfile`, future identifier spelling `business_profile_id`, and capability prefix `business_profile.*`. UI placement is **Group settings → Business Profile**; `GroupProfile` is not the reusable record name. Earlier names in archived/detailed planning records are historical. Stable Goal IDs, management filenames/links, UUID identity and all cardinality/derived-role/active-tenant guarantees remain. Claude must reconcile affected provisional Plan/schema/interface/fixture pins before execution; this does not authorize edits to closed artifacts, source Skills or active G004 inputs. Criteria, budgets, checkpoints and G021 are unchanged.

## Accepted-state revision — 2026-10-13

r3 follows accepted G007 and the current version-5 schema/guarded facade.
Exact r2 archive (not included) preserves earlier assumptions.
Spec AC1–AC14, B7 and the security/data tier remain. Preserve G007 request/session
binding, snapshot/row-lock order, conservative commit outcome and PG-001; assess
new RLS lookup/lock costs under PG-002 before selecting a mitigation. No new
capabilities, domain schema, public adapters, infrastructure or G021 waiver.
See the reconciliation record (not included).

## Execution amendment 01 — 2026-10-13

Bounded continuation (not included) adds the exact
migration-inventory unit test, CI count controls and operations tables to scope;
88 immutable inputs and 15 scope expressions replace 90/12 under recorded
reconciliation. Independent evidence-only oracle delegation is retained, and the
internal transaction-helper split/migration role switching are accepted subject
to their existing guard and database evidence. Plan r3 bytes, fourteen ACs,
B7/checkpoint 4 and G021 remain unchanged; attempt 1 continues. No implementation
ACCEPT or merge is granted by this amendment.
