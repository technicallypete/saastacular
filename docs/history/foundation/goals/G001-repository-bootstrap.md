> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G001 — Bootstrap repository and local services from the archetype

Epic: EPIC-001 — Repository and local platform
Status: Prepared for Claude implementation planning; execution awaits Codex Plan approval
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget class: B5 — Standard

## Objective

Consume the accepted G000 project archetype to establish this repository's runnable, containerized platform skeleton. Preserve its validated boundaries and configuration behavior, add the bounded bootstrap obligations missing from the templates, and demonstrate the actual repository works without host language runtimes.

This is platform bootstrap, not completion of the tenant/security foundation. Product-specific feature work remains blocked through G021.

## In Scope

- Apply the repository-owned `skills/project-archetype/` renderer to the actual repository, preserving existing governance, architecture, management, and Skill artifacts.
- Establish `web/`, `core/`, `intelligence/`, `contracts/`, `infra/`, generated foundation documentation, environment examples, ignore rules, and renderer provenance.
- Adopt the existing Go, Next.js, FastAPI, PostgreSQL, Redis, and source-built local MinIO skeletons. Preserve their readiness, secret scoping, network boundaries, and configuration-only identity.
- Add pinned, containerized database migration tooling and a minimal baseline, with separate owner/migrator/runtime authority and disposable migration lifecycle tests. No tenant or product schema is introduced.
- Add dormant official MCP Go SDK server initialization with an empty catalog and in-process tests. No network MCP transport or authentication substitute is exposed before G004.
- Provide containerized bootstrap/check commands and repository-specific validation for startup, configuration overrides, safe repeat rendering, migrations, service boundaries, persistence, and intelligence enabled/disabled.
- Document actual bootstrap additions, inherited scaffolding, and obligations remaining for later goals.

## Out of Scope

- Modifying the accepted archetype, its templates, or G000 artifacts; a discovered archetype defect returns to Codex for a separately bounded correction.
- G002's accessible web shell, shadcn/ui initialization, and Organization token overrides; retain the rendered minimal UI and semantic tokens.
- G003's baseline CI, expanded observability, metrics, and contract evolution; retain existing health/logging/check scaffolding.
- Production identity/session integration, tenant tables, RoleBinding implementation, application RLS policies, audit ledger, signed-object authorization, MCP switching or dynamic capabilities, verified domains, and onboarding. These remain later foundation work.
- Copying G000 isolation fixtures or development authenticators into platform runtime code.
- Product feature schemas, workflows, seed data, prompts, tools, commerce, or detailed product branding.
- Hot reload, cloud deployment, external account provisioning, live AI requests, vendor replacements, or global Skill installation.
- Editing AGENTS.md, CLAUDE.md, canonical architecture/security/decision documents, or existing goal artifacts to authorize implementation.

## Dependencies

- **G000 accepted:** PROGRESS.md (not included) records Codex ACCEPT after correction loop 1. The [G000 Result](../results/G000-project-archetype-result.md) and [approved Plan](../plans/G000-project-archetype-plan.md) supply inherited evidence and limitations.
- Read AGENTS.md (not included), CLAUDE.md (not included), ARCHITECTURE.md (not included), [SECURITY.md](../../../SECURITY.md), [DECISIONS.md](../../../DECISIONS.md), ROADMAP.md (not included), BUDGET_POLICY.md (not included), and the [G001 Work Spec](../specs/G001-repository-bootstrap-spec.md).
- Inspect the accepted Skill, particularly its configuration contract, migration/MCP follow-up obligations, and invocation safety. Security and decision documents are empty at the starting baseline; no additional ADR is inferred.
- Claude writes `management/v0/plans/G001-repository-bootstrap-plan.md` and waits for Codex approval before rendering or implementation.

## Completion Definition

1. All G001 Work Spec acceptance criteria have concrete passing evidence from the actual repository or clearly identified disposable variants of its implementation.
2. The accepted renderer has supplied the skeleton; bootstrap additions are attributable and bounded. Existing canonical/management/Skill content is preserved.
3. Container-only startup/checks, migration role/lifecycle checks, dormant MCP initialization, intelligence-toggle checks, and security boundary controls pass. No real provider credentials or live model calls are required.
4. Claude submits `management/v0/results/G001-repository-bootstrap-result.md` with commands, files, provenance, acceptance mapping, limitations, and shared cycle accounting.
5. Codex returns **ACCEPT** and updates progress. G002 and G003 then become dependency-ready; the product-feature barrier remains closed.

## Execution Policy

Before Plan approval, inspect and plan only. After approval, Claude may render, implement the approved bootstrap additions, modify platform tests and setup documentation, run containerized builds/checks/tests, diagnose failures, and refactor within scope. The Plan must enumerate application paths, dependencies, migration metadata, and the dormant MCP boundary before execution.

This Goal assigns the bootstrap identity values in the Spec; they do not require another identity-collection round. Preview the resolved configuration and renderer actions in the Plan. Codex approval authorizes those reviewed actions. Unanticipated collisions or changes remain reviewable exceptions; never silently overwrite existing content. Destructive tests operate only on uniquely named disposable projects and volumes, never existing developer data.

## Iteration Budget

Follow management/BUDGET_POLICY.md (not included). **B5: maximum five meaningful implementation/debug cycles shared by Claude and all implementation subagents.** G000's B15 amendment does not carry forward.

Record each agent's usage, failures, diagnosis, fixes, and retests in the Result; concurrent subagent cycles count separately. At shared cycle 3, record criteria completed/remaining, suite status, repeated failures, uncertainty, scope convergence, and the forecast against five cycles. Successful routine validation and read-only review do not independently consume cycles.

Claude may recommend a budget change in planning but cannot change an approved budget without Codex authorization. At exhaustion or material thrashing, stop, preserve full validation requirements, document attempted approaches and blockers in the Result, and return control to Codex. Track any authorized correction passes separately; do not reset the Goal counter or grant subagents independent budgets.

## Context / Scope Budget

One bounded implementation cycle: adopt the already tested skeleton, add migration plumbing and dormant MCP initialization, then validate the integrated repository. Reuse the accepted artifacts instead of rebuilding the archetype or the later foundation. If this cannot fit reliably in available context, recommend an exact split in the Plan and wait for Codex review before implementing.

Reasoning recommendation: High, with XHigh reserved for a concrete security/architecture ambiguity requiring analysis.

## Suggested Subagents / Delegation

Claude decides tactically in the Plan. Prefer zero to two bounded helpers: one for migration-role/lifecycle validation or one for independent bootstrap/security review. If delegating implementation, assign non-overlapping files and cycle allowances within the shared five-cycle ceiling. Require cycle/status/blocker handoffs. Claude integrates and validates all work; product direction, architecture authority, scope changes, and final integration responsibility remain undelegated.

## Stop / Escalation Conditions

Stop affected work and return to Codex for architecture contradictions, scope expansion, security-model changes, destructive/risky migrations, new public API/MCP contracts, major infrastructure/vendors, irreversible data-model decisions, or conflict with canonical architecture/accepted ADRs. MCP network exposure or a temporary identity bypass is outside this approved boundary.

Also stop for unexpected destination conflicts affecting protected content, required changes to the accepted Skill, inability to validate without weakening checks, budget exhaustion, or a goal too large for reliable context. Existing MinIO replacement and production provider decisions remain product-owner decisions; bootstrap cannot resolve them silently. Codex escalates major architecture/product choices to the product owner.
