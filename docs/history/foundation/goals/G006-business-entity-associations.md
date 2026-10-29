> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G006 — BusinessProfile and Group associations

Epic: EPIC-002 — Identity and tenant foundation
Status: r3 — owner-approved single-profile cardinality and derived access; provisional Plan reconciliation required; implementation not authorized
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget class: B5 — maximum five shared meaningful cycles; checkpoint after cycle 3

## Objective

Add Organization-owned, UUID-identified BusinessProfiles containing reusable business context, single-profile-per-Group associations reusable across multiple Groups and read-only access facts derived from Organization/Group RoleBindings. Prove shared context reuse inside one Organization and rejection of cross-Organization links by service validation and PostgreSQL constraints.

BusinessProfile is the canonical domain object; Business Profile is its generic UI/documentation label. No product UI or commerce implementation is implied. Preserve a future BusinessProfile-to-many-BillingProfiles relationship without creating billing schema. The product-feature barrier remains closed through G021.

## Context and Rationale

Drafting observes main HEAD `19f945af7b7bfd1b27cb395fd56b0c2cec3b163b`; G001/G002 are accepted, G047/G003 execute separately, and G004/G005 definitions are drafts. Their proposed models are dependencies to reassess after acceptance, not existing implementation to rely on now.

The canonical architecture and accepted archetype require BusinessProfile reuse across Groups in one Organization. A Group-owned one-to-one profile or a copied profile/grant per Group would violate that model. G006 extends the actual accepted G005 tenant/binding integrity mechanism; G007 later evaluates Organization and capability-specific Group authority.

## In Scope

- Minimal Organization-scoped BusinessProfile persistence with immutable UUID, bounded display/business-context fields and no Group parent.
- UUID-identified GroupBusinessProfile links, same-Organization composite foreign keys, service prevalidation, at most one profile association per Group and shared-record reads/updates.
- Read-only profile-access facts from current Group roles and attachments; no independent BusinessProfile owner/admin grants.
- Internal core profile/link/access-fact operations needed for model tests, transaction failure/concurrency controls and deletion/reference protection.
- Focused additive migrations, actual runtime-role SQL tests, container/CI integration, model documentation, exact provenance and a G006 Result.

## Out of Scope

- New policy engine, trusted request context, inherited capabilities, RLS or full attack-suite acceptance; G007–G010 own these.
- Profile/onboarding/editor/member-management UI, public HTTP/BFF/MCP contracts, invitations, automatic Group ownership/grants or completed-onboarding state.
- BillingProfile tables/foreign keys, tax/payment configuration, pricing, invoicing, commercial/provider integrations and reporting/AI workflows.
- Cross-Organization BusinessProfile sharing, merge/deduplication by business name, tenant transfer, ownership-transfer workflow or new lifecycle policy.
- Branding/appearance/hostname changes, files/storage, durable audit, provider/session changes, production deployment and product-specific schema/features.
- Accepted Skill/source-kit edits, template-manifest rewrites, broad refactors or unrelated cleanup.

## Dependencies

- **G005 implementation ACCEPT**, including its actual Organization/Group UUIDs, RoleBinding authority/projections, scope integrity, migrations/runtime grants and Result. G004/G003/G047 acceptance is inherited transitively.
- Canonical governance/architecture/security/decisions, roadmap/progress, BUDGET_POLICY (not included), accepted archetype tenancy contract, predecessor Specs/Results/reviews and the [G006 Work Spec](../specs/G006-business-entity-associations-spec.md).
- Reassess the actual accepted post-G005 schema and binding representation before Claude planning. Obtain a fresh supervisor baseline/input binding at Plan approval; no speculative schema/table names or historical batch snapshot are authoritative.
- Before approval, pin profile-field/error/transaction contracts, derived-access contracts and migration rollback effects. Resolve any required model/architecture decision with Codex/owner.
- Existing approved Docker/PostgreSQL/GitLab validation facilities and separately authorized hosted evidence actions. No extra vendor or live provider setup is required for synthetic model fixtures.

## Completion Definition

1. Work Spec AC1–AC12 pass, including actual runtime SQL, cross-Organization link negatives, shared-context reuse, derived-access integrity, migration and required CI evidence.
2. Result identifies exact schema/internal contracts, diff, failures/deviations, full inherited regression outcomes, shared B5 accounting and cycle-3 checkpoint.
3. Codex reviews actual implementation and attributable evidence, returns ACCEPT and updates PROGRESS. Passing tests alone is insufficient.
4. G007 becomes dependency-ready for a separate Goal/Spec/Plan handoff. No policy/RLS, onboarding, billing or product feature is accepted by G006; the G021 barrier remains closed.

## Execution Policy

This draft authorizes management specification work and read-only prerequisite research only. Claude Plan preparation waits for G005 acceptance and a Codex handoff. No implementation, dependency installation, migration, build, hosted/provider action or commit/push is authorized by drafting.

After explicit G006 Plan ACCEPT, Claude implements/tests only the reviewed core model, additive migrations, internal contracts and exact docs/harness/CI paths. No public profile/grant/provisioning route or client-controlled bootstrap is allowed. Any unreviewed contract, data-model or security expansion returns to Codex before execution.

Reuse accepted wrappers/harness/isolation/cleanup and core-only datastore credentials. Destructive/race/outage tests use positively owned disposable copies/databases; protect developer state and unrelated resources. Shared-harness edits trigger affected static/control checks within the same cycle and final full regressions. Do not build another execution framework or broaden exclusions.

## Iteration Budget

**B5: maximum five shared meaningful implementation/debug cycles**, checkpoint after cycle 3, following BUDGET_POLICY. This is a bounded extension of accepted tenant/binding/migration controls: one context entity, one association and one read-only derived-access representation. Policy/RLS, public UI and commerce remain separate, so B7 is not assigned speculatively.

Claude must recommend a reviewed revision/split before implementation if the accepted G005 representation makes this bounded extension infeasible within B5. Record aggregate delegate cycles and convergence; stop on exhaustion or thrashing. Drafting/routine commands consume no cycle and do not authorize extra process engineering. Use the owner's xhigh preference for planning.

## Context / Scope Budget

One reusable BusinessProfile model, one-profile-per-Group links inside one Organization, and derived reads of the existing Organization/Group RoleBinding mechanism. No new authority store, generic metadata platform, onboarding state machine, policy evaluator or external integration. Keep minimum identity/Group integration within one bounded Goal.

## Suggested Subagents / Delegation

Claude may use one non-overlapping database/test delegate or a read-only integrity/security reviewer. The Plan names file ownership/integration and shared-cycle accounting. Do not delegate architecture, inheritance semantics, scope decisions or final acceptance. This is optional Claude execution guidance, not Codex delegation authorization during drafting.

## Stop / Escalation Conditions

- G005 not accepted, predecessor schema/authority assumptions invalid, missing reviewed inputs or required validation access.
- Need for a one-Group BusinessProfile parent, cross-Organization sharing, writable membership mirror, copied inherited grants or independent BusinessProfile authority.
- Premature policy/RLS/public profile/onboarding/billing/product implementation or client/provider/model-supplied authority.
- Irreversible model change, destructive existing-data migration, transfer/cascade/merge semantics, new vendor/infrastructure or inherited check weakening.
- Unowned resources, scope/context expansion, repeated nonconvergence or B5 exhaustion. Return the blocker; do not defer required evidence to G007.

## Revision Record

- 2026-10-05 — Initial G006 Goal/Spec drafted with G005 at the owner's request. Dependencies remain unaccepted; no Plan, implementation, commit or push is performed by drafting.


## Parallel provisional planning — 2026-10-07

Codex authorizes drafting a provisional G006 Plan in an isolated clone while G004/G005 remain unimplemented/unaccepted. This supersedes only the requirement to wait for G005 ACCEPT before planning. Use G004 Goal/Spec r3 and design-accepted Plan r4, and G005 Goal/Spec and provisional design-accepted Plan r2 with its Codex review. Label exact schema/module/migration/grant/harness assumptions provisional. Reconcile the actual accepted G005 Result and predecessor state before final approval and execution binding.

Planning only at xhigh, optionally one isolated Claude planning subagent. No implementation, installations, builds, Docker/provider/CI actions, active G048 checkout/input/resource edits, commit or push. G005 ACCEPT/integration remains an implementation prerequisite; B5/checkpoint three, all ACs, UUID/shared RoleBinding/same-Organization invariants and G021 remain. Return substantive ambiguities to Codex rather than invent accepted facts or expand scope.

## Owner-approved cardinality clarification — 2026-10-08

The owner requires exactly one Business Profile per completed Group, reusable by multiple Groups within its Organization. Global users may manage multiple profiles across their authorized Organizations, with UI visibility restricted to the active Organization and its policy. BusinessEntity is reusable profile/context, not a third tenancy tier; Group-owned resources retain Organization/Group scope. G006 enforces at most one link per Group; unfinished foundation Groups may have no link until G018 enforces onboarding completion. Client-facing creation/selection remains G018, not this internal model Goal. See [the decision](../../../adr/ADR-004-group-business-profile-cardinality.md) and [scope addendum](../specs/G006-business-profile-cardinality-addendum.md).

The preserved r1 Goal (not included), provisional Plan r1 and its approval are historical. Their permission for multiple profiles on one Group is superseded; Claude must reconcile its Plan before execution. B5/checkpoint three, AC1–AC12, predecessor gates, direct-grant policy and G021 remain unchanged. No implementation or active G004 input change is authorized.

## Owner authority revision — 2026-10-08

r3 removes independent BusinessEntity owner/admin grants. Current Group-owner/admin bindings plus attachments supply management facts; removal/downgrade/unlink needs no second role update. Organization authority remains independent. The r2 Goal (not included) preserves the preceding direction. Plan r1 requires a bounded amendment for the simpler schema and derived-fact fixtures; no budget extension or implementation release.

## Owner naming clarification — 2026-10-08

Use canonical `BusinessProfile` (UI: **Business Profile**) instead of `BusinessEntity`; the shared association becomes `GroupBusinessProfile`, future identifier spelling `business_profile_id`, and capability prefix `business_profile.*`. UI placement is **Group settings → Business Profile**; `GroupProfile` is not the reusable record name. Earlier names in archived/detailed planning records are historical. Stable Goal IDs, management filenames/links, UUID identity and all cardinality/derived-role/active-tenant guarantees remain. Claude must reconcile affected provisional Plan/schema/interface/fixture pins before execution; this does not authorize edits to closed artifacts, source Skills or active G004 inputs. Criteria, budgets, checkpoints and G021 are unchanged.
