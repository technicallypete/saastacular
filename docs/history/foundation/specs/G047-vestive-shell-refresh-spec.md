> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G047 — Successor Vestive shell visuals work specification

Goal: [G047-vestive-shell-refresh.md](../goals/G047-vestive-shell-refresh.md)
Epic: EPIC-001 — Repository and local platform
State: Prepared for Claude planning; implementation not authorized

## Owner-authorized combined handoff — 2026-10-05

The product owner requests G047 → G003 followed by a combined Codex review. The supervisor execution authorization (not included) permits the phase transition only after all G047 criteria, final validation and Result are complete. No intermediate acceptance is claimed, and all seven criteria, evidence obligations, B3 and scope remain unchanged.

## Rationale and Authority

The accepted G002 shell implements the earlier Vestive reference through its own styling. The successor kit adds a supplied Light palette and detailed component states. Updating a skill does not change application CSS. This follow-up adopts the successor visuals in the existing foundation surface before G003, under ARCHITECTURE.md's “Evolution of the Vestive Design Input and Component Library” rules.

G002 remains accepted against its historical inputs. Do not revise its Plan/Result, acceptance hashes or exhausted budget. G047 has independent artifacts, evidence and B3 accounting. Canonical architecture/governance and accessible behavior take precedence over reference code; marketing-copy and dark-only instructions remain excluded.

## Scope and Prerequisites

Refresh existing shell header/navigation, appearance control, three generic content regions and footer; existing button variants and the existing help dialog, including close/overlay treatments. No new user flow, component catalogue, business content or routing behavior is required. Test-only existing fixtures may exercise existing variants/states that are not displayed in the shell.

G001/G002 acceptance and the complete consolidated skill are prerequisites. Before execution, Codex approves a Claude Plan, exact starting state and input snapshot/digest. Source input manifest `skills/vestive-design-skill/references/shadcn-kit.sha256` currently hashes to `fcda5a273daee8117598ab32a244a36d898112c27bd94b3908fbe7a362c79163`; the 63 preserved kit files must match it. The Plan also pins entry-point/integration-guide bytes and every other selected source. This intake fingerprint is not an execution approval.

The kit's code/assets lack a complete licensing/provenance record. The Plan must distinguish independently implemented visual adaptations from copied material. Establish origin and usage permission/notices for any copied code before approval; no new asset/font or whole-component copying is implicitly authorized. If that prerequisite cannot be satisfied, escalate that part rather than silently ship it.

## Requirements

### R1 — Observable successor adoption

Provide a mapping for each existing semantic color token in both appearances: current value/role, relevant successor source value/role, proposed value, consuming surfaces/states and any adaptation. Account for typography, spacing, geometry and structure used by the scoped shell/button/dialog. Implement the reviewed mapping, rather than merely documenting that the kit was read.

Use the successor warm-paper Light ground (`vx-bg` #f4f2ee), alternate Light surface (#ebe8e3), Light raised surface (#ffffff), Dark grounds (#0d0d0d/#101010/#131313), and relevant text/accent/danger/control roles on their mapped consumers. Exceptions require a documented measurable contrast/usability reason approved in the Plan. Not every source role belongs in the existing token set: record unused roles, do not introduce charts, ratings or marketing primitives to consume them.

Map by role, not identical slot names. In particular, accepted `accent` means emphasis ink, not the kit's rule-colored hover surface. Preserve all 16 existing semantic color names, override input shape and mode-aware resolution. Keep palette literals centralized; component source consumes semantic styling. Retain zero radius, ruled structure and local/system typography roles. The kit's percent corner cuts may be adapted to avoid clipped labels/focus, as already established in G002.

Kit palette/state values are visual inputs, not permission to reduce contrast. Its #0075ff with white normal-size text fails 4.5:1; retain or derive a passing primary treatment and explicitly map that exception. Make the same measurement-based decision for secondary, destructive, hover and focus states. Do not copy the kit's color assertions without measuring the rendered application.

### R2 — Behavioral and architecture invariants

Preserve Base UI button/dialog APIs, server/client boundaries, configured identity, generic content/navigation, skip link, landmarks, keyboard interactions, modal focus/inertness, close/return-focus and scoped portal placement. No new routes or contract changes. No Radix, next-themes, new packages/version changes, font files or external requests. No wholesale kit CSS/component import or motion/assets.

Light/Dark/System remain available, default System. Keep browser-local persistence, invalid/unavailable-storage fallback, OS tracking only in System, parser-blocking head initialization, earliest-painted palette evidence and hydration behavior. Do not rewrite the appearance engine to adopt reference selectors. Keep user preference distinct from source palettes and scope-local Organization-style overrides; no Organization data/schema/persistence.

### R3 — Accessibility and scope composition

Measure actual painted text/surface pairs at rest and relevant hover/focus/selected/disabled states in both appearances, including translucent dialog/backdrop combinations. Require >=4.5:1 normal text, >=3:1 large text under the existing definition, and >=3:1 for required UI/focus indicators. Retain existing per-control/geometry assertions; report disabled-state exemptions explicitly rather than treating them as failures or inventing exemptions for enabled states.

At 320 CSS-pixel width and desktop, preserve no horizontal page overflow, readable/nonclipped labels and focus, 200% zoom usability and reduced-motion behavior. For touched shell/control text, add a spacing trial with line height 1.5, paragraph spacing 2em, letter spacing .12em and word spacing .16em; no hidden/overlapped text or lost controls. This bounded trial does not claim full WCAG certification or resolve every historical accessibility follow-up.

Verify default, sibling and nested scopes and dialogs with overrides in each resolved appearance. Light-only overrides retain inherited Dark values and vice versa; switching user mode re-resolves scopes without leaking overrides to neighbors or moving portals outside their intended region.

### R4 — Validation and repository integrity

Keep accepted G001/G002 assertions and dependency/wrapper controls. Update independent palette/state expectations only for reviewed visual changes, with source mappings explaining each update; no self-derived expected values that make tests tautological. Existing suite/test counts are historical baselines, not permission to skip coverage or limits on added assertions.

Exact-path changes may include web styling/components, affected browser expectations/tests, WEB-SHELL documentation, consumer customization register and project harness guards/disabled-variant reconciliation when necessary. The Plan must enumerate each such change and retain protected-file, dependency and source checks. No broad allow-list/scan exclusions. Shared harness changes require same-cycle static/control checks plus affected browser regressions. Preserve original renderer manifest and project-archetype source; project-owned customizations remain attributable.

Run the complete current project harness through the fresh approved-input wrapper against the delivered tree, including real-root and intelligence-disabled browser matrices, identity rename, builds, checks and inherited datastore/migration/isolation/outage/persistence scenarios. Destructive scenarios remain disposable. Failed checks and post-run drift fail acceptance; preserve actual harness status. No hosted CI evidence is required for G047 and no G003 CI work is authorized.

## Security and Architectural Constraints

No production/application changes outside the reviewed visual/test/provenance surface; core remains authoritative, web remains UI/thin BFF and intelligence internal. No service authority, credentials or tenancy behavior changes. Never introduce product schema before its owning Goal. Future durable product/domain identities remain UUIDs; slugs/names are alternate/routing identifiers. Infrastructure/framework key strategies are unaffected. No future product feature is authorized before G021 barrier acceptance.

## Expected Artifacts and Evidence

- Claude-owned G047 Plan and Result, plus narrowly affected WEB-SHELL/provenance docs.
- Exact approved commit/start inventory, input paths/hashes and snapshot digest; pre/post verification of immutable inputs and evidence of no unapproved drift.
- Source-to-consumer mapping and copy/adaptation/licensing provenance; prior G002 acceptance remains linked as historical evidence.
- Comparable actual-shell and open-help-dialog screenshots before/after in both Light and Dark, at desktop and narrow width. Identify viewport, browser/container, fonts and revision/tree. Capture open overlays after transitions settle; screenshots supplement measured assertions, not first-paint timing evidence.
- Exact validation commands, suite/control results, failures/retests, change inventory, cycle counts/checkpoint and owned-resource cleanup record. Evidence must not contain credentials or secret-bearing environment values.

## Acceptance Criteria and Verification Strategy

| ID | Required outcome | Required verification/evidence |
|---|---|---|
| AC1 | Scoped consumers implement the reviewed successor visual mapping. | Full per-token/per-appearance and scoped structure/state mapping; independent expected-value/style assertions for mapped grounds, typography/structure and button/dialog states; before/after actual shell and overlay screenshots show those consumers. Each retained old value is justified, not silently omitted. |
| AC2 | Existing interaction/configuration and dependency contracts are preserved. | Keyboard/landmark/navigation/skip/dialog/focus regression suite; configuration-only identity rename; dependency controls and package/module diff show no additions or version/category changes; no new routes, marketing literals, assets/fonts or network requests. |
| AC3 | Appearance and first-paint guarantees remain effective. | All accepted Light/Dark/System, persistence, storage failure, live OS-change, explicit-mode isolation, delayed-framework-JS/paint-correlated earliest-frame and production/dev hydration scenarios pass with updated independent palette expectations. Missing paint evidence fails, and later screenshots cannot substitute. |
| AC4 | Scope overrides remain separate from preference and isolated. | Existing fixture scenarios in both appearances plus partial Light-only/Dark-only overrides, siblings/nesting and scoped dialogs; independent assertions on scope/overlay styles before/after mode changes and absence of neighbor leakage. Fixture routes remain absent from production. |
| AC5 | Updated visuals meet the specified rendered accessibility checks. | Rendered contrast matrix for actual mapped text/UI/focus surfaces and states; keyboard/reduced-motion, 320px reflow, desktop/200% zoom, angle-label/focus and touched-text spacing trials pass. Record exact measurements, exemptions and evidence limits. |
| AC6 | Full stack and validation/provenance regressions pass against the submitted tree. | Approved-wrapper full harness exit 0 with no required skipped/unmet checks; enabled/disabled browser coverage and inherited scenarios pass; all dependency/wrapper/static controls retained, protected inputs unchanged, NUL-safe full change inventory equals exact reviewed scope and postcheck shows no drift. |
| AC7 | Handoff is attributable, bounded and safe for G003 reconciliation. | Result maps AC1–AC7 to artifacts/commands, source manifest and approved input digest, copy/adaptation provenance, actual cycle/checkpoint and cleanup record; no schema/product/CI implementation or closed-record rewrites. Codex ACCEPT recorded separately; G003 must subsequently bind the accepted refreshed baseline. |

## Permitted Deviations and Gate Conditions

Plan-reviewed contrast adjustments, role translations, local/system font fallback and nonclipping angle adaptations are permitted, with explicit values/consumers and passing independent evidence. Source tokens not used by this bounded surface may remain unused. No behavioral/token-contract/primitive substitution or acceptance-evidence deferral is permitted. New deviations return to Codex before acceptance; architecture changes go to the product owner.

Before implementation: reviewed source/provenance, exact visual mapping and paths, B3 feasibility, Plan ACCEPT and fresh baseline binding. Before completion: all criteria and full regressions, truthful Result and Codex ACCEPT. Before G003 execution: accepted G047 baseline, G003 Plan reconciliation and fresh approval/binding, plus G003's existing runner and hosted-authorization prerequisites.

## Risks / Questions Before Implementation

- Resolve copy-versus-independent-adaptation provenance; missing kit licences block unreviewed source copying, not preservation as a reference.
- Validate which kit hues need accessible alternatives, especially primary/secondary fills, subtle borders, focus and destructive states. Preserve role semantics despite source slot differences.
- The exact starting HEAD must reflect the accepted inputs and unrelated pending planning/skill work. No approved execution snapshot exists yet; do not promote continuity evidence to approval.
- If the owner intends all kit components or motion/brand assets rather than this existing-shell refresh, that is a separate scope decision; this Spec does not authorize it.

## Revision Record

- 2026-10-05 — Initial definition of the owner-requested successor visual adoption before G003. Seven objectively verifiable criteria; no implementation performed or authorized.
