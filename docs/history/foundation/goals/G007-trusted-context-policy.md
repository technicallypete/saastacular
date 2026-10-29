> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G007 — Trusted context and shared policy evaluation

Epic: EPIC-003 — Authorization and isolation
Status: r3 — Reassessed against accepted G006; ready for Claude planning, implementation gated
Owner: Codex (Goal/Spec/review); Claude Code (Plan/implementation/validation)

## Objective

Resolve trusted Organization and optional Group context from an authenticated global user, evaluate one deny-by-default capability policy over the accepted G005/G006 facts, and enforce it on a bounded internal core operation surface. Prove capability-specific BusinessProfile inheritance and immediate effects of committed authority changes without copied grants or premature public tenant access.

## In Scope

- Core-owned, request-scoped TenantContext resolution using accepted G004 sessions and current Organization/Group facts.
- The closed capability matrix in the Work Spec, derived from the accepted archetype, with Organization and Group-derived BusinessProfile paths kept distinct.
- One Organization policy flag controlling Group-owner/admin ordinary Business Profile edits; minimum persistence/grants/migration if needed, with no configurable policy language.
- Guarded internal Organization/Group/BusinessProfile reads, ordinary Group/profile updates and owner-only policy updates; evaluate future-sensitive capabilities without implementing their workflows.
- Real runtime-role database, policy, forgery, revocation, failure/concurrency and boundary tests; focused documentation and existing CI integration.

## Out of Scope

- PostgreSQL RLS/pool-context setup (G008), object storage (G009), complete attack-suite acceptance (G010), durable AuditEvent (G011).
- New public HTTP/BFF/MCP endpoints, tenant switching/session persistence or discovery (G014/G015), hostname/custom-domain resolution (G016).
- Membership administration, invitations, Group creation/onboarding completion, ownership transfer or billing/payment/tax/legal workflows. Their capabilities may be evaluated; their side effects are not authorized.
- Product roles/features, UI or design changes, Organization-enforced MFA, new passkey/factor integrations, provider/session protocol changes, deployment, vendors or generalized execution/CI frameworks.
- Rewriting accepted migrations, duplicate authority stores or active G004 checkout/input changes.

## Dependencies

- **G006 implementation ACCEPT and committed integration**, satisfied by the [final G006 review](../results/G006-implementation-codex-review.md#final-goal-acceptance--2026-10-10), merge `4486b41`, transitively accepted G004/G005 and the bootstrap/CI/visual foundation. Planning observation is staging `6e38fdadd9fa1b2fe551d77314d0f1f5dd658c55`, not an execution binding.
- Canonical ARCHITECTURE, DECISIONS/ADRs, AGENTS/CLAUDE, BUDGET_POLICY, roadmap/progress, the [Work Spec](../specs/G007-trusted-context-policy-spec.md), and accepted predecessor Results/reviews/deviations when available.
- Reconcile actual identity/session interfaces, tenancy model/grants/migrations and validation/CI before execution; bind fresh committed inputs and an exact scope using the accepted procedures.
- Explicit review of the capability/context contract and any policy migration in Claude's Plan. Hosted evidence actions require their named authorization; no new live-provider setup is required for policy fixtures.
- G051 is independent and is not a planning/implementation prerequisite. Reconcile whichever CI implementation is actually integrated before execution and later validation; do not assume its proposed selector, reporting or timing behavior is accepted.

## Completion Definition

1. Work Spec AC1–AC14 pass on the submitted implementation with attributable local and hosted evidence and no required skips.
2. Result records the exact policy/internal-operation/schema inventory, G008 handoff, failures/deviations, provenance, cleanup and shared B7 accounting.
3. Codex reviews the actual diff and evidence, returns ACCEPT and updates Progress. G008 then becomes dependency-ready; G007 does not claim RLS or public tenant isolation completion.
4. G021 remains closed; no product-specific feature or later workflow is accepted by this Goal.

## Execution Policy

**Architecture/security/data tier:** separate Claude Plan and explicit Codex approval are required. G007 is ready for read-only inspection and Plan preparation against accepted implementations; it is no longer provisional predecessor planning. This reassessment does not dispatch Claude or authorize implementation. Capability/context decisions, exact internal transaction contracts, scope/input binding and named hosted authority retain their approval gates.

After approval, Claude may autonomously implement, test and diagnose within the bound scope. Reuse the accepted harness/wrapper, isolation/cleanup and actual CI trigger matrix. Fetch/rebase/reconcile before every local validation batch, including focused checks; use the Docker Desktop mount safeguard where applicable. Finish any required operator feedback before final qualification. Collect required local full evidence before the MR, automatic MR fast and the applicable post-integration staging disposition under the current solo-operator policy; its status rule does not waive known failing checks or local evidence. Do not rebuild unchanged execution safeguards or create new proof campaigns. Other checkouts, bound inputs and resources remain protected.

## Iteration Budget

**B7: maximum seven shared meaningful implementation/debug cycles**, checkpoint after cycle 4, under BUDGET_POLICY. This is first authorization integration across sessions, tenant facts and capability inheritance; B3/B5 would understate its concurrency/security scope. Claude forecasts cycles during planning and recommends a split before implementation if B7 is not credible. No reset or unilateral extension; zero cycles consumed by specification drafting.

## Context / Scope Budget

One context resolver, one closed evaluator, one policy flag and one bounded guarded internal facade. No rule editor, arbitrary policy DSL, distributed invalidation/epoch service, new authority database or transport work. Reuse predecessor services rather than refactoring their entire model. The Plan identifies the exact method surface and minimum shared-file changes.

## Suggested Subagents / Delegation

During approved Claude work, one isolated policy/test delegate and one read-only security reviewer may help. Claude decides in its Plan, owns integration and shares B7 across delegates. Non-overlapping file ownership is required. Architecture/product choices, scope and acceptance cannot be delegated. No Codex subagents are dispatched by this Goal.

## Stop / Escalation Conditions

- Missing predecessor acceptance, incompatible accepted session/model contract, invalid binding or unavailable required evidence.
- Need to infer authority from provider roles, email, hostname, display name, user/model claims or an unvalidated context; cache stale positive authorization or copy inherited grants.
- Matrix/context semantics contradict canonical architecture; need for tenant/MFA policy beyond the single reviewed flag, global admin or unreviewed sensitive-operation authority.
- New public contracts, RLS, storage, switching/onboarding/billing/product work, major dependencies/vendors, security-model changes or destructive existing-data migrations.
- Required guarantees need broad predecessor redesign or a new generalized execution/invalidation framework; return the smallest concrete gap for scope review.
- Repeated nonconvergence, cycle-4 checkpoint failure, B7 exhaustion or unowned resources. No test weakening or deferred G007 evidence.

## Revision Record

- 2026-10-07 — r1 drafted at owner request, observing staging `29312802c6b33c10df1254df8e8bb663cc6cba29`. G004 executes separately; G005/G006 Plans are provisional design approvals. No Plan/Result, implementation, provider/CI action, commit or push is created by drafting.

## Owner revision — 2026-10-08

ADR-004 removes independent BusinessEntity roles and adds Group owner. Updated Spec r2 defines eleven capabilities over six roles; fourteen criteria and B7 remain. Group-owner/admin profile management derives from current attached-Group facts, member grants read only, Organization authority remains independent, and losing the last qualifying path immediately removes management without synchronization. r1 Goal (not included) is preserved. Plan approval/predecessor gates/G021 remain; no implementation.

## Owner naming clarification — 2026-10-08

Use canonical `BusinessProfile` (UI: **Business Profile**) instead of `BusinessEntity`; the shared association becomes `GroupBusinessProfile`, future identifier spelling `business_profile_id`, and capability prefix `business_profile.*`. UI placement is **Group settings → Business Profile**; `GroupProfile` is not the reusable record name. Earlier names in archived/detailed planning records are historical. Stable Goal IDs, management filenames/links, UUID identity and all cardinality/derived-role/active-tenant guarantees remain. Claude must reconcile affected provisional Plan/schema/interface/fixture pins before execution; this does not authorize edits to closed artifacts, source Skills or active G004 inputs. Criteria, budgets, checkpoints and G021 are unchanged.

## Accepted-baseline reassessment — 2026-10-10

r3 preserves the exact r2 Goal (not included) and
r2 Spec (not included). See the
reassessment (not included). Actual identity,
RoleBinding and BusinessProfile implementations are accepted and integrated;
no G007 Plan exists yet. Scope remains one resolver/evaluator, one policy flag
and the bounded internal facade, AC1–AC14 and B7/checkpoint four. Plan preparation
must address the existing independent transactions rather than infer an atomic
authorization/write boundary. No implementation, delegation, hosted action,
budget consumption or G021 release occurs here.

ADR-004's later UX clarification supersedes the exclusive Group-settings
placement above: Groups and Business Profiles are sibling management areas,
with Group settings linking to the same shared page. G018 owns that UI;
G007 introduces neither page nor route.

## Execution amendment — 2026-10-10

Codex blocker disposition (not included) approves the named RoleBinding lock fallback, records the existing small-fixture sequential-scan observation and permits the exact reviewed shared scanner correction at one added path. Original Plan/snapshot history and all ACs remain. Explicit attempt accounting: zero historical consumed counts retained, active initial attempt 1, B7 total shared goal-level attempts with checkpoint after attempt 4. Inner debugging loops/subagents do not independently consume increments. Reconcile amended inputs/scope before tests; no G051 completion dependency is introduced.
