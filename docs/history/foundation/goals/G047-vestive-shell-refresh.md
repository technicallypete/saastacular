> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G047 — Adopt the successor Vestive visuals in the existing shell

Epic: EPIC-001 — Repository and local platform
Status: Prepared for Claude planning; implementation not authorized
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget: B3 — maximum three shared meaningful cycles; checkpoint after cycle 2

## Owner-authorized combined handoff — 2026-10-05

The owner requests G047 followed by G003 before a combined Codex implementation review. Follow the supervisor execution authorization (not included): complete every G047 criterion, final validation and Result before the phase transition. Do not mark G047 accepted. Its B3 and scope remain independent of G003 B7. This changes the intermediate handoff timing only.

## Objective

Adapt the consolidated Vestive kit's Light/Dark palettes and visual component states to the accepted generic web shell, existing button variants, appearance control and help dialog. Preserve the accepted behavioral and semantic-token contracts. Complete this bounded follow-up before G003 implementation.

G047 is the next unused roadmap identifier, inserted by dependency order before G003. It does not renumber existing goals or reopen G002 acceptance. G002's seven consumed cycles and historical source hashes remain unchanged.

## In Scope

- Source-to-consumer mapping of successor palette roles, typography, spacing, ruled structure, flat geometry and relevant button/dialog states.
- Repository-owned styling of existing shell regions and controls only, in both appearances, using the current semantic token names and Base UI primitives.
- Focused browser assertions, before/after screenshots, provenance documentation, and exact-path reconciliation of consumer customization and validation guards where required by this change.
- Full accepted-stack regression validation and a separate G047 Result.

## Out of Scope

- Importing the whole kit, installing components not used by the existing shell, building a component gallery or a comprehensive component library.
- Radix or framework replacement, dependency/version changes, new fonts/assets or external asset requests, ribbon/node-field effects, marketing copy or logos.
- Appearance-engine rewrites, new token APIs, Organization persistence, authentication, domain schema/features, new API/MCP routes, G003 operational/CI work or unrelated cleanup.
- Editing accepted G002 artifacts, source kit files, project-archetype Skill, generator templates or original render manifest.

## Dependencies

- G001 and G002 accepted; G002 acceptance commit `7193ed84237be9362b7b596bdd135dd592447b07` is historical evidence, not a current execution binding.
- Read canonical governance, architecture/security/decisions, roadmap/progress, BUDGET_POLICY, G002 artifacts and acceptance/deviation notes, and the [G047 Work Spec](../specs/G047-vestive-shell-refresh-spec.md).
- Use the consolidated `skills/vestive-design-skill`, especially its integration guide, kit palette and matching button/dialog/typography/layout sources. Pin the complete reviewed input inventory in the Plan.
- Resolve source-use provenance for any copied code before Plan approval. A visual adaptation does not authorize copying unlicensed components/assets.
- Codex must approve the Plan and fresh execution baseline/snapshot. Uncommitted skill consolidation and management/Plan work must be explicitly committed or exactly bound; no broad dirty-tree exception.

## Completion Definition

1. Work Spec AC1–AC7 pass with attributable evidence and no required scenario skipped or deferred.
2. Codex reviews the actual diff, both-appearance screenshots, validation and provenance, returns ACCEPT, and updates progress.
3. G003's Plan is reconciled to the accepted refresh and submitted for fresh scope/baseline review before G003 execution. G047 acceptance does not approve G003 or waive its runner/hosted-action prerequisites.

## Execution Policy

This handoff authorizes inspection and Claude Plan preparation only, at `management/v0/plans/G047-vestive-shell-refresh-plan.md`. No application changes begin before explicit Codex Plan ACCEPT and execution binding. Claude owns implementation and Result reporting; Codex does not perform the visual implementation.

After approval, implement and validate only exact reviewed paths. Reuse the approved-input validation wrapper with a fresh G047 snapshot/digest; historical G002 and G003 continuity snapshots are not authorization. Preserve existing environment files and Docker data/resources. Destructive/outage tests remain on owned disposable copies. No commit, push, pipeline trigger or external configuration action is authorized here.

## Iteration Budget

**B3: at most three meaningful implementation/debug cycles shared across agents; checkpoint after cycle 2 before cycle 3.** Separate accounting from G002's exhausted B7 and G003's unused B7. Expected work is a style adaptation and regression update, not a new theme system. If planning demonstrates that the bounded work cannot fit B3, recommend a split or budget revision and stop; do not extend it or weaken checks.

## Context / Scope Budget

One coherent existing-shell refresh. All additional kit primitives remain reference material. An implementation-ready Plan must enumerate the precise visual inventory and independent expected values, adaptation decisions, changed paths and guard updates. No future component rollout or product design program is authorized.

## Suggested Subagents / Delegation

Prefer one implementer; at most one read-only visual/accessibility reviewer is suggested. Claude decides in the Plan and retains integration responsibility. Any corrective implementation counts toward the same B3. Use the owner's requested xhigh effort for planning.

## Stop / Escalation Conditions

- Architecture/security conflict, primitive/framework switch, dependency change, source-use rights uncertainty, new assets/font loading, public contract or schema change.
- Need to weaken accepted checks, change semantic token meanings/APIs or appearance/portal mechanisms, alter protected source/template files, or expand beyond existing shell controls.
- Missing approved baseline, drifting source inputs, developer-resource conflicts, nonconvergent validation or B3 exhaustion.
- Any proposed product-specific work: the foundation barrier remains closed through G021, including this added foundation goal.

## Revision Record

- 2026-10-05 — Product owner requests adoption of the updated Vestive skill before G003 execution. Codex creates a separate bounded follow-up and execution dependency; no application implementation or Plan approval.
