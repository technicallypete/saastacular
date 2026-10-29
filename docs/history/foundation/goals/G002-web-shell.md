> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G002 — Accessible generic web shell and design-token foundation

Epic: EPIC-001 — Repository and local platform
Status: Plan r6 approved; Claude implementation authorized against the approved snapshot, subject to S0a/S0b/S0c gates
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget class: B7 — Complex (authorized at r2 review)

## Objective

Extend the accepted web skeleton into a small, accessible, generic application shell using repository-owned shadcn/ui components backed by Base UI. Establish tested semantic-token overrides suitable for future Organization branding while preserving runtime identity, the thin BFF boundary, and Docker-only workflows.

Use the repository's Vestive design skill (not included) as the visual-language reference for the delivered shell, adapted through semantic tokens and the Work Spec's accessibility, appearance-mode, and generic-content constraints. Ignore its marketing copy, content guidance, and dark-only rule; shell copy remains generic, identity remains configured, and Light/Dark/System support with System default remains required.

This completes the remaining G002 foundation obligations. It introduces no Organization records, authentication, tenant selection, or product features. The product-feature barrier stays closed through G021.

## Context and reassessment

On 2026-10-05, Codex verified that commit `b40507794f3ebcc0dc84cc66fbab8102c89978fb` (`G001 acceptance`) contains all 92 accepted implementation/Result paths plus the PROGRESS.md acceptance update. Both normal and ignored-inclusive working-tree inventories were empty before this planning work.

G002 remains dependency-ready and the correct next bounded Goal. Actual state narrows the roadmap's original wording:

- Next.js App Router, React, TypeScript, Tailwind v4/PostCSS, and initial semantic CSS variables already exist. Preserve and extend them rather than scaffold again.
- Runtime-configured page identity and metadata, server-only core readiness calls, and web health/readiness endpoints already work. Preserve them; no new backend contract is needed.
- shadcn/ui/Base UI, reusable shell components, interactive keyboard behavior, token-override browser evidence, and a real browser test suite are absent.
- `web`'s current `check` runs lint and type checks only. G001's Result calls these web checks/tests, but no web behavioral test suite exists at this baseline.
- The project harness currently assumes web dependency files match the original manifest, and its disabled variant copies only the G001 additions. G002 must reconcile these consumer-specific assumptions with its reviewed web customizations without changing the accepted Skill or historical evidence.

G003 is also dependency-ready. Its existing health/logging scaffolding will require reassessment when that Goal is prepared; it is not a prerequisite for this shell and receives no detailed Spec here.

## In Scope

- Preserve the existing `web/app`, `web/lib`, Tailwind integration, configured identity, and core readiness bridge.
- Initialize shadcn/ui for the existing Next.js project using Base UI, with committed component source and only the primitives needed by this Goal.
- Build a responsive generic shell at the existing root page: configured identity, navigation to functioning generic destinations or page sections, main content, skip navigation, and a small working Base UI interaction.
- Extend the semantic CSS-token contract for the installed primitives and shell; demonstrate two independently scoped styling overrides without implementing Organization configuration storage or resolution.
- Support accessible Light, Dark, and System appearance choices, defaulting to System; persist the browser's choice and test system-preference changes, both resolved palettes, and scoped overlays.
- Apply the Vestive skill's visual structure, typography hierarchy, and color roles to the bounded shell; document source mappings and adaptations for both appearances, accessibility, and configured identity.
- Add Docker-run browser and accessibility validation, including keyboard/focus behavior, computed token styles, responsive behavior, runtime rename, and frontend boundary checks.
- Make bounded project check/harness/documentation changes needed to validate the evolved consumer, retain renderer provenance, and preserve the applicable intelligence-disabled web variant.

## Out of Scope

- Rebuilding Tailwind, the BFF, health/readiness, migrations, or the accepted archetype; changing the manifest's original intended hashes.
- G003 CI workflows, metrics, new trace propagation, cross-service contract evolution, or broader observability.
- Authentication/session integration (G004), Organization/Group/BusinessEntity schema (G005–G006), authorization/RLS (G007–G008), or new public API/MCP contracts.
- Hostname-based Organization resolution, branding persistence/editors, arbitrary theme uploads, or the logged-out tenant shell (G016–G017).
- Tenant switching, fake memberships, mock authorization in runtime, domain screens, product navigation, seed data, AI/chat UI, billing, acquisition, or onboarding.
- Inventing a new detailed product brand or expanding the supplied design language, marketing pages/campaigns, external fonts/assets, large component galleries, all-component installation, hot reload, deployment, or vendor replacement. Use of the supplied Vestive visual guidance within the shell is in scope; its marketing content and business workflows are not.

## Dependencies

- **G001 accepted and committed:** PROGRESS.md (not included), [Goal](G001-repository-bootstrap.md), [Spec](../specs/G001-repository-bootstrap-spec.md), [Plan](../plans/G001-repository-bootstrap-plan.md), and [Result](../results/G001-repository-bootstrap-result.md). The Result's submission status is historical; the committed progress acceptance is authoritative.
- Review AGENTS.md (not included), CLAUDE.md (not included), ARCHITECTURE.md (not included), [SECURITY.md](../../../SECURITY.md), [DECISIONS.md](../../../DECISIONS.md), accepted ADRs if added, ROADMAP.md (not included), BUDGET_POLICY.md (not included), and the [G002 Work Spec](../specs/G002-web-shell-spec.md).
- Read Vestive SKILL.md (not included), its README.md (not included), styles, motion source, and reference example before revising the visual approach. This repository-local design input is required; retain its source files unchanged during G002 implementation.
- No dependency on G003, live providers, tenant data, an external design service, or a selected identity provider.

## Completion Definition

1. The G002 Work Spec's AC1–AC10 pass with real, containerized browser evidence against the evolved implementation.
2. Generic shell/components and token overrides work accessibly; runtime identity and service boundaries remain intact.
3. Existing bootstrap checks remain meaningful after explicitly attributed consumer customizations; unrelated accepted implementation and original provenance are preserved.
4. Claude submits a goal-specific Result with exact commands, acceptance mapping, diff, dependency/provenance facts, limitations, cleanup, and shared budget accounting.
5. Codex returns **ACCEPT** and updates progress. Submission or Plan approval alone is not completion.

## Execution Policy

The product owner's submission of Plan r1 for Codex review released this Goal and Work Spec for planning. Codex's 2026-10-05 r6 review approves the Plan and authorizes Claude to begin G002 against `/tmp/g002-r6-approved-inputs.sha256`, with that snapshot's own SHA-256 pinned in the supervisor approval. The r6 Plan SHA-256 is `30501d828768efdbc307217f5400827d1e0e55758befcc42af19d9782f979f41`; its exact external S0a script SHA-256 is `853b20dfe26b974479d1a794cd4ccd024f620edcfe2571e9d8ad5b252916b428`. Both are part of the approved execution contract. S0a must pass before repository writes; S0b/S0c must pass before S1. No other planning inputs or regenerated snapshot are authorized. This is Plan approval, not implementation acceptance.

After Plan approval, Claude may modify the reviewed web shell, components, tokens, tests, and necessary project check/harness/docs within scope, run containerized validation, and correct failures. The Plan must enumerate files, installed/generated components and pinned dependencies, browser tooling, test fixtures, intended consumer customization conflicts, and acceptance mapping. Existing major frameworks and runtime defaults stay in place.

No database/schema change is authorized. Durable product/domain entities retain UUID canonical identity; names/slugs are alternate display/routing identifiers, never authority. Infrastructure/framework keys may differ when justified, including existing migration metadata; G002 does not migrate them or introduce early domain entities.

Preserve existing developer `.env` and Docker resources. Destructive or environment-override trials use uniquely named disposable projects/copies proven to contain the current implementation. Cleanup removes only execution-owned resources. Do not commit, publish, or deploy unless separately requested.

## Iteration Budget

**B7: maximum seven meaningful implementation/debug cycles shared across all implementation agents; mandatory checkpoint after cycle 4.** Follow BUDGET_POLICY's definitions and stop at exhaustion or thrashing. No G001 allowance carries forward and no further extension is implicit.

B7 is now the smallest appropriate class for the expanded coupled frontend foundation: mode-aware scoped token resolution, pre-paint appearance initialization and hydration validation, supplied design-language integration, rendered contrast/geometry checks across both appearances and three scopes, and full consumer-provenance regressions. Claude forecasts five to six meaningful cycles; retaining B5 would leave insufficient confidence after the owner's additions. Keep one bounded G002 instead of repainting/revalidating the same token machinery in a separate Goal. This budget adds no schema, infrastructure vendor, or product scope.

Record aggregate cycles, diagnosis/fix/retest evidence, and completed/remaining criteria at the checkpoint. Read-only review and successful routine checks do not independently consume a cycle. Claude may recommend a smaller scope, split, or budget revision during planning; only Codex can authorize changes.

## Context / Scope Budget

One bounded web foundation increment: a root shell, the minimum reusable primitives it consumes, a semantic-token contract, Light/Dark/System appearance, and necessary validation. Avoid building a general design system or a future tenant-branding subsystem. If dependable completion requires broad harness reconstruction or unrelated framework upgrades, stop during planning and propose a bounded split.

Reasoning recommendation: High; reserve XHigh for a concrete architecture/security ambiguity.

## Suggested Subagents / Delegation

Prefer no implementation delegation for the tightly coupled shell/token work. Claude may use two bounded read-only reviewers as proposed in r2: accessibility/boundary/provenance and design-language fidelity. They edit no files; corrective implementation remains Claude's responsibility and shares the B7 ceiling. Any implementation delegation requires a reviewed Plan revision with non-overlapping file ownership. Architecture, scope, and final acceptance remain undelegated.

## Stop / Escalation Conditions

Stop affected work for architecture contradictions, scope expansion, security-model changes, public API/MCP changes, new major infrastructure/vendors, destructive migrations, or irreversible data-model decisions. Product-owner decisions go through Codex.

Also stop if Base UI adoption requires replacing the accepted framework stack, if token work requires tenant schema/resolution or runtime authorization fixtures, if the accepted Skill needs edits, if validation cannot run without weakening checks, or if context/budget is insufficient. Reconcile expected project customizations explicitly; do not silence provenance or scope assertions broadly. No product-specific work is authorized before G021 acceptance.

## Revision record

- 2026-10-05 — Initial draft based on committed G001 and actual web implementation; submitted with the Work Spec for product-owner review. No G002 implementation Plan or code created.
- 2026-10-05 — Planning handoff reconciled at the product owner's request to review Plan r1. This records the current release for planning, without asserting an independently verified earlier approval. Scope, AC1–AC8, B5, and the G021 barrier remain unchanged; r1 is not authorized for implementation.
- 2026-10-05 — Product owner explicitly adds Light/Dark/System appearance to G002, with System the default. Work Spec R8 and AC9 define behavior and validation; AC1–AC8 remain required. B5 is retained pending Claude's revised feasibility assessment. Plan r1's no-dark-theme choice is superseded; implementation still awaits revised Plan approval.
- 2026-10-05 — Product owner requires use of the repository-local Vestive design skill. Work Spec R9/AC10 define visual integration, semantic mapping, and adaptations; earlier neutral-only/no-design-skill assumptions are superseded. All functional/security criteria and the foundation barrier remain in force. Claude must incorporate this input into the revised Plan; no implementation approval is granted.
- 2026-10-05 — Product owner explicitly excludes marketing copy and the skill's dark-only rule. Apply visual guidance in both appearances, retaining System as the default mode.
- 2026-10-05 — Codex r2 review authorizes B7, checkpoint after cycle 4, and two read-only reviewers for the expanded coupled foundation; the proposed split is not selected. Plan r2 still requires two bounded planning corrections before implementation approval. Prior B5 decisions remain historical; no implementation cycles have been consumed by this review.
- 2026-10-05 — Codex r3 review retains B7 and all scope/criteria. The revised first-paint mechanism is adequate as a plan, but overlapping dependency-control expectations and final-validation failure propagation require correction before approval. No application files changed and no implementation is authorized.
- 2026-10-05 — Codex r4 review finds the independent dependency-rule/control expectations coherent and retains the first-paint evidence requirements. The proposed wrapper cannot serve S0 when its creation is scheduled for S5; reconcile its fail-closed bootstrap before approval. B7, checkpoint 4, scope and acceptance criteria remain unchanged. No implementation is authorized.
- 2026-10-05 — Codex r5 review finds the three-stage bootstrap ordering coherent, but the exact S0a script creates its evidence directory before rejecting an in-repository destination and permits Git's optional index refresh. Correct these two read-only violations, refresh the gate text/hash and demonstrate no repository writes before approval. B7 and all acceptance criteria remain unchanged; no implementation is authorized.
- 2026-10-05 — Codex approves Plan r6 after focused review and independent scratch checks resolve both S0a read-only defects. Claude is authorized against the exact approved snapshot and pinned Plan/script hashes, subject to every start/validation/stop gate. Explicitly approve B7, maximum seven shared meaningful cycles and checkpoint after cycle 4; no acceptance evidence is deferred. G002 implementation remains unaccepted, with no schema or product-feature authorization.
