> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G000 — Reusable multi-tenant SaaS project archetype Skill

Epic: EPIC-000 — Reusable multi-tenant SaaS archetype  
Status: Active — specification prepared; implementation plan pending review  
Implementation owner: Claude Code  
Architecture, security, and acceptance reviewer: Codex

## Objective

Create and validate a repository-owned, reusable Skill that guides creation of a generic multi-tenant SaaS project. The Skill must preserve the canonical architecture while accepting project identity, domain, database, provider, and optional intelligence configuration. It must be usable by another project without product-specific assumptions or source-code search/replace.

This goal delivers the Skill and its validation evidence. Instantiating the actual platform begins in G001; completing the core foundation remains the responsibility of G001–G021.

## In scope

- A self-contained `skills/project-archetype/SKILL.md` entry point with supporting references, configuration/templates, checklists, and bounded validation fixtures as specified in the [work specification](../specs/G000-project-archetype-spec.md).
- Explicit inputs, defaults, validation, invocation workflow, output contract, and safe handling of existing destination files.
- Docker-first project setup with container-provided runtimes, configured host ports, service health/readiness, and reproducible validation commands.
- Guidance for `web`, `core`, optional `intelligence`, `contracts`, and `infra` service boundaries, global identity, Organization/Group/BusinessEntity tenancy, shared RoleBinding policy, PostgreSQL RLS, storage isolation, audit, and MCP context/discovery.
- Generic AI configuration and ModelGateway boundaries when intelligence is enabled, including an OpenRouter-first adapter strategy.
- Security/observability/CI and SOC 2 readiness evidence hooks; tenant-isolation validation requirements and minimal disposable examples.
- Evaluation and composition of available commodity skills, with documented dependencies and usable fallback guidance where optional skills are unavailable.
- Validation in disposable generic destinations, including renamed/re-domained configuration and intelligence disabled, with an acceptance checklist and result evidence.

## Out of scope

- Application implementation or bootstrap in this platform's `web`, `core`, `intelligence`, `contracts`, or `infra` directories; that starts with G001.
- Completing the production identity, tenancy, authorization, audit, MCP, domain, or onboarding systems assigned to later foundation goals. Fixtures demonstrate the Skill's contracts without implementing the full foundation.
- Product-specific feature domains, workflows, schemas, tools, prompts, seed data, branding, and integrations.
- Billing/payment functionality, commercial catalogs, marketplace or matching engines, and product delivery objects.
- Production deployment, provider account provisioning, paid services, live model traffic, or platform-wide installation of the Skill.
- Major infrastructure additions without an ADR, speculative scaling, or claims of SOC 2 compliance.

## Dependencies

No prior implementation goal is required. Read AGENTS.md (not included), ARCHITECTURE.md (not included), [SECURITY.md](../../../SECURITY.md), [DECISIONS.md](../../../DECISIONS.md), ROADMAP.md (not included), and PROGRESS.md (not included).

The architecture is authoritative, especially its canonical tenancy override, reusable archetype, revised epic sequence, configuration/gateway conventions, folder naming, and Docker-first policies. SECURITY.md and DECISIONS.md are currently empty; do not infer additional approved decisions from them.

Claude must prepare `management/v0/plans/G000-project-archetype-plan.md` for Codex review before execution. Resolve architectural conflicts through the repository's review/ADR process. Missing optional commodity skills must not silently weaken the archetype's invariants.

## Completion definition

G000 is complete only when:

1. The repository-owned Skill and all required support artifacts exist, are internally linked, and meet the work specification.
2. Disposable validation demonstrates generic configuration portability, Docker-only project runtimes, optional intelligence behavior, correct boundaries, and the specified security/isolation examples. Results distinguish executed checks from guidance reserved for later implementation goals.
3. No platform application implementation or product-specific domain has been introduced by this goal.
4. Claude records artifacts, commands, outcomes, limitations, and acceptance evidence in `management/v0/results/G000-project-archetype-result.md`.
5. Codex reviews the Skill, diff, architectural/security requirements, and evidence and returns **ACCEPT**. Otherwise the result is **CORRECTION LOOP** or **ARCHITECTURE DECISION REQUIRED**.
6. PROGRESS.md records acceptance and G001 becomes eligible. G000 acceptance alone does not open the product-feature barrier; G021 must accept the entire foundation.

## Amendment record

### A1 — Execution budget amendment (2026-10-04)

Authorized by the product owner at the start of implementation, after Codex approved plan r4.

- The iteration budget is changed from the default 5 to a **hard ceiling of 15 meaningful implement → test → diagnose → fix → retest cycles** for the entire goal. The ceiling is shared by the main agent and every subagent, and the counting rules in plan §6.1 apply.
- **Mandatory checkpoints at cycles 8 and 12.** At each one, the execution log in the Result records:
  - cycles consumed per agent;
  - suite status;
  - remaining work;
  - a forecast against the ceiling.
- This amendment changes the execution budget only. It does not change the objective, scope, requirements, acceptance criteria or architecture. The original goal text above is unchanged.
