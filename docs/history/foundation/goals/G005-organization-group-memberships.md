> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G005 — Organization and Group membership model

Epic: EPIC-002 — Identity and tenant foundation
Status: r2 — Group creator/owner correction requires provisional Plan amendment; implementation not authorized
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget class: B7 — maximum seven shared meaningful cycles; checkpoint after cycle 4

## Objective

Establish UUID-identified Organizations and Groups, editable Organization-local Group slugs, and scoped membership primitives backed by one mutable RoleBinding authority store. Prove that one accepted global identity can hold different roles in different Organizations and Groups without copied grants or a second identity.

This Goal establishes durable model integrity, not an externally usable tenant authorization system. Public tenancy operations remain closed until their policy, isolation and surface-owning Goals. The product-feature barrier remains closed through G021.

## Context and Rationale

Drafting observes main HEAD `19f945af7b7bfd1b27cb395fd56b0c2cec3b163b`. G001/G002 are accepted; G047/G003 execute separately and remain unaccepted. The G004 identity/session definition is a draft, not an implemented schema or an execution dependency already satisfied.

The roadmap still places G005 after G004. The canonical multi-tenant override in ARCHITECTURE and the accepted archetype require Organization as the tenant boundary, Group as collaboration scope, global user UUIDs, Organization-local Group slugs and shared RoleBinding. Memberships expose canonical read-only projections; they do not add another writable role store. Use the actual accepted G004 identity key and migration contracts when planning, rather than guessing their table names now.

## In Scope

- Minimal Organization and Group persistence, canonical UUIDs, owning `organization_id`, bounded display names and parent/reference integrity.
- Suggested/editable kebab-case Group slugs, Organization-local uniqueness, reserved-route validation and concurrent collision handling.
- Organization/Group RoleBinding scopes, reviewed role/source vocabulary, one membership role per subject/scope and read-only canonical membership projections.
- Internal core create/read/update and membership grant/change/remove primitives needed to exercise the model; no public provisioning or administration surface.
- Actual PostgreSQL constraint, concurrency, transaction, migration, runtime-grant and persistence tests, integrated into accepted container/CI checks.
- Focused model documentation, exact consumer provenance/customization reconciliation and a G005 Result.

## Out of Scope

- BusinessProfile, GroupBusinessProfile or BillingProfile schema; G006 owns profile/link facts. No independent BusinessProfile grants are planned.
- Trusted request TenantContext, capability evaluation, Group permission inheritance, RLS policies, storage authorization and full isolation acceptance; G007–G010 own these.
- Organization/Group HTTP/BFF/MCP routes, tenant switching, a user directory, membership administration UI, invitations and onboarding.
- Hostname resolution, Organization routing slugs/custom domains, persisted branding, appearance preferences or design-system changes.
- Ownership-transfer workflows, last-owner policy, cross-Organization moves, account deletion, durable audit ledger and production deployment.
- Product-specific entities, demo businesses, consultant/client roles or features, and changes to accepted Skills/template manifests.

## Dependencies

- **G004 implementation ACCEPT**, including its accepted global user/session/provider boundary, migrations, runtime grants, relevant CI and Result. G003/G047 acceptance is inherited through G004.
- Canonical governance, architecture/security/decisions, roadmap/progress, BUDGET_POLICY (not included), predecessor Specs/Results/review records, the accepted archetype tenancy contract and the [G005 Work Spec](../specs/G005-organization-group-memberships-spec.md).
- Before Claude planning, reassess the accepted post-G004 tree and its deviations. Codex must provide a fresh baseline/input binding at Plan approval; this draft's observed HEAD is not that binding.
- Resolve the Plan's concrete role/scope constraints, projection representation, slug contract and migration risks before approval. Any conflict with canonical authority semantics goes through Codex to the owner.
- Existing approved Docker/PostgreSQL/GitLab test facilities. No new identity provider, tenancy vendor or major infrastructure is required by this Goal. Separately authorize any necessary hosted evidence actions.

## Completion Definition

1. Work Spec AC1–AC12 have attributable passing evidence; model, direct-SQL integrity, concurrency, migration and required CI checks are neither skipped nor deferred.
2. Result identifies the exact diff, migration/schema/internal contracts, inherited regressions, failures/deviations, B7 accounting and cycle-4 checkpoint.
3. Codex reviews actual implementation and evidence, returns ACCEPT and updates PROGRESS. Test success alone is not acceptance.
4. No external tenancy surface, policy/RLS completion, BusinessProfile schema or product capability is claimed. G006 is still separately planned and the G021 barrier remains closed.

## Execution Policy

This draft authorizes management specification work and read-only prerequisite research only. It does not authorize a Claude Plan handoff ahead of G004 acceptance, implementation, dependency installation, migrations, builds, provider/hosted changes or commit/push.

After explicit G005 Plan ACCEPT, Claude may implement/test the enumerated core schema, internal contracts and exact harness/docs/CI paths. Model migrations and internal contracts are anticipated scope and must be reviewed before execution. Do not expose a model helper by adding an HTTP route, MCP tool, request-controlled bootstrap or production test authenticator.

Reuse accepted approved-input, harness, isolation and cleanup controls. Runtime datastore credentials remain core-only; destructive tests use owned disposable copies/databases. Preserve developer data and unrelated Docker/runner resources. Every shared-harness edit reruns its affected static/control checks within the same cycle and the full final regressions. No broad skip/exception or new execution framework.

## Iteration Budget

**B7: maximum seven shared meaningful implementation/debug cycles**, checkpoint after cycle 4, under BUDGET_POLICY. This is the first persisted tenant/authority model, with scope-safe references, read-only projections and concurrent migration/integrity evidence. That coupled security-sensitive foundation warrants B7 rather than a narrow CRUD allowance; it does not authorize capability or RLS implementation.

Claude records aggregate cycles across delegates, remaining criteria and convergence at the checkpoint. Routine commands and drafting consume no cycles. Stop on exhaustion or repeated nonconvergence; neither extra cycles nor weaker evidence are implied. Use the owner's xhigh preference for planning. Claude may recommend a smaller bounded split/budget before approval if the accepted identity schema changes feasibility.

## Context / Scope Budget

One Organization model, one Group model and one shared Organization/Group binding mechanism. Minimum identity integration; no new provider/session system, policy engine, public tenancy UI or generic orchestration library. Keep model invariants and evidence comprehensible in one Goal; scope growth returns to Codex.

## Suggested Subagents / Delegation

Claude may use one non-overlapping test/migration delegate and one read-only security reviewer. Name file ownership and integration responsibilities in the Plan; all corrective implementation counts against the same B7. Do not delegate role semantics, architecture decisions or final acceptance. This suggestion does not authorize Codex implementation/delegation during drafting.

## Stop / Escalation Conditions

- G004 not accepted, accepted identity/baseline mismatch, missing reviewed inputs or required test/hosted access.
- A second writable membership/role store, implicit Organization/Group/global authority, provider-claim grants or premature tenant context/policy/RLS/public surface.
- Need to change canonical hierarchy, transfer tenant ownership, introduce BusinessProfile/product schema, select a vendor or weaken database integrity.
- Irreversible/unreviewed model decisions, destructive existing-data migration, audit/security conflict or proposed cascade/last-owner lifecycle policy outside this Goal.
- Inherited checks cannot remain meaningful, resources are not positively owned, scope/context exceeds this Goal, thrashing or B7 exhaustion.

## Revision Record

- 2026-10-05 — Initial G005 Goal/Spec drafting authorized alongside G006. G004 and the active G047/G003 batch remain unaccepted prerequisites; no implementation, Plan, commit or push is performed by drafting.


## Owner-authorized parallel planning — 2026-10-07

The owner authorizes a provisional G005 Plan while G004 is not implemented/accepted. This supersedes only the earlier requirement to wait for G004 ACCEPT before planning. Read G004 Goal/Spec r3, accepted Plan r4 and its Codex review as provisional integration inputs; label proposed table/module/migration dependencies as assumptions, not accepted implementation facts. Reconcile against the actual accepted G004 Result, schema, grants and operational evidence before final Plan approval and execution binding.

Planning only, xhigh, in an isolated clone outside the active G048 checkout. No application/schema/test implementation, dependency installation, builds, Docker/provider/CI changes, commit or push. G004 ACCEPT/integration still gates G005 implementation. Scope, ACs, B7/checkpoint four, Codex Plan approval, exact execution binding and G021 remain unchanged. If provisional dependencies prevent a concrete decision, name it as an implementation prerequisite rather than expanding scope or inventing accepted facts.

## Owner revision — 2026-10-08

Under accepted ADR-004, Group roles are owner/admin/member and Group creation atomically assigns the explicit creator `group_owner` through the sole RoleBinding store. Public onboarding supplies the verified creator later in G018; no UI/transfer/last-owner workflow is introduced in G005. The r1 Goal (not included) and provisional Plan r2 approval are retained but require a bounded amendment for these changed requirements. Existing budget, predecessor gates and G021 remain.

## Owner naming clarification — 2026-10-08

Use canonical `BusinessProfile` (UI: **Business Profile**) instead of `BusinessEntity`; the shared association becomes `GroupBusinessProfile`, future identifier spelling `business_profile_id`, and capability prefix `business_profile.*`. UI placement is **Group settings → Business Profile**; `GroupProfile` is not the reusable record name. Earlier names in archived/detailed planning records are historical. Stable Goal IDs, management filenames/links, UUID identity and all cardinality/derived-role/active-tenant guarantees remain. Claude must reconcile affected provisional Plan/schema/interface/fixture pins before execution; this does not authorize edits to closed artifacts, source Skills or active G004 inputs. Criteria, budgets, checkpoints and G021 are unchanged.
