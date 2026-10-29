> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G048 — Adopt Vestive source components, bundled fonts and blueprint imagery

Epic: EPIC-001 — Repository and local platform
Status: Revised draft r2; Radix architecture decision owner-approved; Claude Plan and implementation not authorized
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget: B5 — maximum five shared meaningful cycles; checkpoint after cycle 3

## Objective

Use the supplied Vestive component source in the generic shell, bundle its named fonts locally, and add static blueprint SVG backgrounds conveying the beginning of an architectural design. Deliver actual source-derived components and rendered typography, rather than another values-only approximation.

This is a separate owner-requested foundation increment. G002/G047 acceptance, historical inputs and consumed budgets remain unchanged. The owner explicitly chooses Radix instead of Base UI on 2026-10-06. Follow [ADR-001](../../../adr/ADR-001-radix-vestive-components.md): adopt the selected kit components with their Radix primitives, migrate all existing Base UI consumers and remove the unused Base UI dependency. Literal adoption does not authorize inaccessible source behavior or changing the appearance/token architecture.

## In Scope

- Adopt the six selected source modules: typography, layout, bullet-list, site-header, button and dialog, including their nonconflicting visual styles and named type roles.
- Migrate the existing button/dialog consumers, test fixtures and scoped portal adapter from Base UI to Radix; use the source Slot/asChild composition and reconcile exact dependency/guard policies.
- Wire the existing shell to those components; retain configured identity, neutral content, navigation, appearance control and help flow.
- Bundle Oswald, Barlow Condensed and Roboto from reviewed, pinned upstream artifacts, including license/copyright notices; serve them from the application origin.
- Add original static SVG architectural drafting imagery to the overview background, with Light/Dark treatment, semantic styling, safe markup and no interaction.
- Focused source/fidelity, font, SVG, accessibility and inherited regression checks, exact provenance, documentation and a separate Result.
- Minimum reviewed updates to existing validation guards that currently forbid copied source/fonts/assets or require the old component geometry.

## Out of Scope

- Installing the full 46-module kit, a production component gallery, additional user flows or unconsumed widget families.
- Additional primitive suites beyond the required Radix Dialog/Slot, next-themes, replacement of React/Next.js or the appearance engine, or a generic design-system publishing platform.
- Marketing copy, logos/social assets, pricing, ribbon/node motion, external font/asset requests or a live architectural diagram editor.
- Organization branding storage/schema, authentication, domain/product models, public API/MCP changes, tenant selection or G003 operational refactors.
- Editing preserved skill sources, historical accepted artifacts, archetype templates/render manifest or unrelated cleanup.

## Dependencies

- G001, G002 and G047 implementation ACCEPT; G047 checkpoint `cbffd6da5d9936c90fdfd9e51d84d36c0f17ad81` is accepted but main integration is pending at drafting.
- **G003 implementation ACCEPT and integration before G048 implementation.** Its current execution checkout, snapshots and batch authorization must remain untouched. Read-only source research can occur now; a Claude Plan handoff follows reassessment of the accepted post-G003 tree.
- Canonical governance/architecture/security/decisions, roadmap/progress, BUDGET_POLICY (not included), G001/G002/G047 Results/reviews and the [G048 Work Spec](../specs/G048-vestive-components-fonts-blueprints-spec.md).
- The accepted [Radix decision ADR-001](../../../adr/ADR-001-radix-vestive-components.md); pin the exact Radix dependency delta and all affected consumer/portal/test contracts in the Plan.
- Confirm the selected kit code's origin, ownership/permission and required upstream notices before copying. Pin component and font inputs, font coverage/weights, source-to-consumer changes and visual measurement expectations before Plan approval.
- Separate Codex Plan ACCEPT and fresh committed baseline/input binding. Neither this draft nor an earlier snapshot authorizes execution.

## Completion Definition

1. Work Spec AC1–AC10 pass, including source lineage, actual custom-font rendering, visible blueprint backgrounds, both appearances and full inherited validation.
2. Result records exact source/assets/licenses, deviations, screenshots, commands/results, cleanup, failures and shared B5 accounting.
3. Codex reviews the actual diff and evidence, returns ACCEPT and updates progress. The G021 integrated foundation review includes this increment; no product-feature barrier is released.

## Execution Policy

Drafting authorizes management definition and read-only inspection only. Claude owns the later Plan, implementation, tests and Result. No application/assets/dependencies or active execution inputs are changed by this handoff; no build, commit, push or hosted action is authorized.

After separate approval, implement only reviewed paths and the scoped Radix migration. Reuse the accepted wrapper, harness, Docker/browser/CI validation, isolation and cleanup. Destructive tests stay on owned disposable copies. Inherit G003's actual accepted hosted-evidence requirements and obtain any needed external-action authorization separately. Do not develop a new execution framework.

## Iteration Budget

**B5: maximum five shared meaningful implementation/debug cycles; checkpoint after cycle 3.** Forecast three to four, conditional on source rights and a passing accepted baseline. Six consumed source modules, three font families and one decorative background are bounded; the button/dialog primitive migration is included, but no whole-library rollout or wider framework replacement. All required evidence stays inside G048.

Planning must recommend a split or reviewed budget revision if Radix migration, source compatibility or font/layout changes make B5 unreliable. No unilateral extension, budget reset or weakened test. Use the owner's xhigh preference for planning; routine inspection/drafting consumes no cycle.

## Context / Scope Budget

One generic-shell component adoption with fonts, a bounded button/dialog Radix migration and a static decorative extension. Selected source modules may expose their supplied variants in dedicated test fixtures, but unconsumed modules remain references. Do not turn this into a platform-wide component migration or a new rendering/validation engine.

## Suggested Subagents / Delegation

Claude may use one bounded source/font-provenance researcher or one read-only visual/accessibility reviewer. The Plan specifies ownership, integration and shared cycle accounting. Architecture, scope decisions and final integration remain with their owners. This guidance does not request Codex implementation delegation during drafting.

## Stop / Escalation Conditions

- G003 remains unaccepted, accepted baseline is unavailable, reviewed inputs drift or resources conflict with active/developer stacks.
- Kit provenance or font/asset redistribution terms remain unresolved; copying cannot silently fall back to values-only implementation.
- Need for primitive/framework changes beyond approved Radix Dialog/Slot, a runtime package outside the reviewed dependency delta, a token API break, appearance-engine rewrite, scope-breaking portal behavior or public contract/schema change.
- Required source visual behavior conflicts with accessibility or scope composition and cannot be resolved by the Spec's limited adaptations.
- Whole-library/product expansion, unsafe/dynamic SVG, external runtime assets, weakening inherited checks, nonconvergence or B5 exhaustion.

## Revision Record

- 2026-10-06 — Owner requests actual Vestive components/fonts and blueprint SVG backgrounds after identifying G047's approved font fallback and component-geometry limitations. Separate G048 drafted; historical acceptance and active G003 work are unchanged.

- 2026-10-06 — r2: owner approves Radix adoption. Supersedes r1's Base UI retention/porting and no-primitive-migration restrictions only. Original r1 (not included) is preserved byte-exactly. B5, ten criteria, fonts/blueprints, G003 prerequisite and foundation barrier remain. No implementation approval.
