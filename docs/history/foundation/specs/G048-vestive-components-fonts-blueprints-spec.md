> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G048 — Vestive component, font and blueprint adoption work specification

Goal: [G048-vestive-components-fonts-blueprints.md](../goals/G048-vestive-components-fonts-blueprints.md)
Epic: EPIC-001 — Repository and local platform
State: Revised draft r2; owner-approved Radix decision; separate Claude Plan approval required

## Goal, Rationale and Authority

Adopt the actual supplied Vestive source components used by the generic shell, load their intended fonts from bundled files, and establish decorative blueprint SVG imagery as a visual expression of beginning architecture.

G047 passed its approved values-only scope. It deliberately retained compact buttons, rectangular outline controls and local/system fallback fonts. Its browser evidence rendered Liberation Sans/Mono, not Oswald/Barlow Condensed/Roboto. Those limitations explain the owner's requested follow-up; they are not grounds for rewriting the prior acceptance record.

Follow AGENTS.md, CLAUDE.md, BUDGET_POLICY, accepted ADRs and ARCHITECTURE.md's **Frontend UI and Design-System Architecture** and **Evolution of the Vestive Design Input and Component Library**. Apply the owner-approved [Radix decision ADR-001](../../../adr/ADR-001-radix-vestive-components.md), preserving semantic scoped styling, thin BFF authority and the closed G021 product-feature barrier. Marketing-copy and dark-only guidance remain excluded.

For this Goal only, the owner's requests permit the scoped Base UI → Radix migration, selected component-source copying and bundled local fonts/new original SVGs once provenance is resolved. It supersedes G047's values-only/no-new-assets limits for the named inventory, not the architecture, accessibility or preserved-source rules. Do not import the whole skill, global resets, source theme engine or barrel exporting unused modules.

## Inspected State, Prerequisites and Execution Order

Drafting observes main HEAD `dbdc651e6f24aa0c20eeebfb25fc4df3d8e4a08b`; main PROGRESS.md contains the uncommitted G047 checkpoint acceptance. The accepted checkpoint is `cbffd6da5d9936c90fdfd9e51d84d36c0f17ad81`; it is not yet integrated into main. G003 is actively implemented elsewhere and is not accepted. This observed state is not an execution binding.

Implementation waits for accepted/integrated G003 and reassessment of its actual harness/CI. Do not add this work to the running G047 → G003 batch or mutate its approved inputs. Before approval require a Claude Plan, resolved selected-source rights/notices, exact font inputs, source/adaptation and Radix dependency/consumer migration inventories, measurable reference expectations and fresh supervisor baseline/snapshot binding.

Inherit G001's configured identity/ports, actual-root commands, datastore role separation, no application datastore access from web and disposable destructive tests; G002/G047's accessibility, appearance/first-paint, scoped portals and independent browser evidence; and G003's actual accepted contracts/logging/CI/provenance guarantees. Historical residuals are not blanket cleanup authorization. No product/domain schema is permitted; future canonical identities remain UUIDs and names/slugs remain alternates.

## Scope and Explicit Non-goals

Production scope is the existing header/nav, appearance control, overview, ruled informational grid, neutral About list, footer and help dialog. Existing content/IDs/navigation remain usable. Source header marketing defaults/brand images are replaced with configured application identity. Additional component states belong in existing dedicated test fixtures, never a production gallery.

Excluded: the rest of the kit; other primitive suites or React/Next.js framework replacement; remote fonts/assets; marketing/brand marks; ribbon/node effects; interactive blueprints; customer-uploaded SVG; domain/tenant storage, authentication, new BFF/API/MCP routes, G003 refactors, broad generator/skill changes and product features. Organization branding persistence stays with G016/G017.

## Requirements

### R1 — Actual Selected Component Adoption

Adopt these source modules from `skills/vestive-design-skill/assets/shadcn/components/ui/` into repository-owned consuming modules:

| Source module | Required production use |
|---|---|
| `typography.tsx` | Kicker, Heading and Text roles in overview/grid/dialog-compatible typography; TextLink for a suitable existing link |
| `layout.tsx` | Container, Section and RuledGrid around existing neutral regions |
| `bullet-list.tsx` | BulletList/BulletItem for the existing About list |
| `site-header.tsx` | Header/nav layout adapted to configured text identity and existing appearance/help controls |
| `button.tsx` | Source-derived variants and sizes for the existing help button and test fixture controls |
| `dialog.tsx` | Source-derived panel, title, description, close, overlay and structural pieces for the existing help dialog |

Actual adoption means traceable source-derived component implementations, imported and used by the built shell. Merely adding unused files, declaring source font names, reading the skill or matching a few palette values does not satisfy R1. Preserve reusable component composition and nonconflicting source visual declarations. The Plan lists each source hash, consumer path, exports/production consumers and every source-to-consumer alteration. Retain the source Radix Dialog and Slot/asChild composition; use existing `cn`/CVA facilities unless the Plan identifies and escalates a concrete compatibility gap. The owner authorizes adding only `@radix-ui/react-dialog` and `@radix-ui/react-slot`, with exact versions reviewed in the Plan, and removing `@base-ui/react` after all consumers are migrated. Record transitive additions/removals and licenses; do not add the whole Radix umbrella package, clsx/tailwind-merge or unrelated packages by following the kit install instructions.

Enumerate and migrate every existing Base UI button/dialog import, render-prop composition, ref/event contract and production/test call site to Radix/native button equivalents. Internal component APIs may change to the source APIs with explicit reviewed call-site migration; public HTTP/MCP and token override input contracts do not change. At completion, no application/test source imports Base UI and its packages are absent from the resolved lockfile dependency tree. Radix is the default for subsequent generic interactive primitives. No mixed or temporary Base UI dependency may be left as completed work. Copy only needed source modules, not `components/index.ts` or the kit's unconsumed assets/helpers.

### R2 — Source Visual Fidelity with Measured Exceptions

The Plan pins source-derived geometry/type/state expectations independently of production implementation. At 1280 × 800 and default text settings, implement the source's 1360px maximum container, `clamp(20px, 4vw, 40px)` gutter, 88px section spacing, 36px cell padding, 2px structural rules, zero radius/shadows and source heading/body/label metrics. Typography uses the supplied clamp expressions and font weights; retain distinct heading ink, body/lead/meta and gold emphasis roles rather than collapsing them for convenience.

Buttons use the source sizes (sm: 30px horizontal/14px vertical, default: 40px/18px, lg: 48px/22px, icon: 44px), typography and variant treatments. Filled and outline variants have the source angled silhouette; outline has a 2px angled frame. Ghost/link variants remain unclipped, with their source gold resting treatment and link border underline. The normal help trigger uses a source size, not the old fixed 32/36px compact override. Test all consumed variants/sizes; unconsumed legacy exports must remain compatible or have a reviewed call-site migration.

Dialog retains the source 560px maximum, viewport-minus-40px width, 36px horizontal/34px top/40px bottom padding, 20px gap, 2px frame, 32px close control at 20px inset, and Oswald title/Roboto description. Header keeps the source 76px minimum-height/ruled layout while supporting its additional existing appearance control. Neutral product identity replaces the supplied image/marketing defaults.

The complete allowed adaptation categories are: Radix/ref/attribute/portal/import and existing-call-site compatibility; configured neutral content/semantic HTML; semantic-token translation; measured contrast/focus correction; responsive/text-spacing/reflow correction; reduced-motion treatment; and local-font loading. Every exception needs a specific source declaration, replacement, reason, affected states and verification in the Plan. Historical G047 compact sizing or rectangular outline choices are not automatic exceptions. A different cut may be proposed only with measured source-label clipping evidence; retain an angled 2px outline frame. Move clipping to a visual layer where necessary so labels/focus remain intact.

At standard desktop measurements, unchanged numeric declarations must match within 1 CSS pixel (font weight/family, border widths, radius and variant selection must match exactly). Responsive adaptations must state breakpoints and expected values; source/reference differences are reviewed explicitly, not accepted solely because broad accessibility tests pass. Accessibility overrides literal reproduction where the supplied code fails, but silent aesthetic substitution is prohibited.

### R3 — Bundled Fonts That Actually Render

Bundle Oswald for display/headings/nav/buttons/dialog titles, Barlow Condensed for kickers/meta/bullet labels, and Roboto for body/lead/dialog descriptions. Include all weights/styles exercised by the selected components (at least Oswald 500/600/700, Barlow Condensed 400/600, Roboto 400/500/700), using reviewed variable or static artifacts. Mono indices keep the existing system mono role. CSS family strings or fonts installed only in the browser image are insufficient.

Font requests are same-origin; binaries and notices are in the production build/image. Builds must not fetch fonts from Google Fonts, a CDN or a package postinstall. Normal production/dev browsers and the Docker test browser use the same delivered assets. Pin upstream commit/release, artifact URL, original and delivered SHA-256, metadata/coverage and copyright/license text. Record any conversion/subsetting tools and resulting name/license obligations. Bundle the upstream notices; do not relabel third-party font copyright as Vitruvian Technology's.

Primary upstream references inspected during specification are the Google Fonts family directories and their OFL files: [Oswald](https://raw.githubusercontent.com/google/fonts/main/ofl/oswald/OFL.txt), [Barlow Condensed](https://raw.githubusercontent.com/google/fonts/main/ofl/barlowcondensed/OFL.txt), [Roboto](https://raw.githubusercontent.com/google/fonts/main/ofl/roboto/OFL.txt). These references are not pinned artifact approval; Claude must bind the exact binaries and corresponding notices in the Plan.

Acceptance must inspect actual rendered platform/custom fonts on representative production nodes at the required weights after loading. Distinguish internal font naming/variable axes from CSS aliases. Verify delivered bytes/provenance and absence of synthetic required weights. Missing/blocked fonts must leave readable fallback text and working controls, but fallback is never the successful normal-load font result. Preserve the first-paint palette guarantee; it does not require waiting for font loading before FCP or hiding text until every font arrives.

### R4 — Appearance and Scoped Visual Roles

Radix dialogs must use the existing scope-container resolver for both overlay and content. The source default body portal is not acceptable within a scoped region. The Plan pins initial scope-container readiness, SSR/hydration and mount/unmount behavior; an overlay must not transiently appear with root tokens before moving into scope. Preserve real modal semantics, focus trapping/restoration, background inertness and nested composition. [Radix Dialog documentation](https://www.radix-ui.com/primitives/docs/components/dialog) defines its portal container and modal APIs; exact package behavior must be tested, not inferred from Base UI evidence.

Preserve Light/Dark/System, default System, local preference persistence, invalid/unavailable-storage fallback, live OS changes only in System, parser-blocking head initialization, hydration and scoped overlay placement. Do not add source `.dark`/`.light` global selectors, next-themes or an alternative preference store.

Preserve the existing 16 semantic color names and override input contract. Source colors must be centralized in mode-aware tokens; source classes/literals must be translated rather than hardcoding brand values in components. Additional internal visual-role aliases are permitted to retain source ink/text/surface distinctions, provided the Plan defines their Light/Dark resolution and how relevant existing overrides propagate. They are not a new Organization branding API/schema. Root-only brand variables that bypass scope overrides are unacceptable. Any required override-contract break returns to Codex.

Keep user appearance, platform design roles and scope-local overrides technically distinct. Test partial mode overrides, nested/sibling isolation, runtime changes and scoped dialogs with the adopted components and blueprint treatment.

### R5 — Static Blueprint SVG Background Language

Add one original, reusable decorative blueprint composition to the overview background. Convey architectural drafting using at least three visible motif families: construction grid/guide lines, dimension lines with ticks, and outlined floor-plan or axonometric structural forms. These are abstract, unfinished architectural forms; no actual product/entity map, labels claiming features, logo, marketing slogan or customer data. Do not rely on a plain grid alone to satisfy the imagery requirement.

Use SVG/vector source, with an explicit viewBox and reproducible source file. Keep the shipped SVG payload at most 40 KiB uncompressed for the composition. No raster embed, external reference, executable script/event handlers, foreignObject, remote CSS/font, SMIL animation, live user input or expensive animated/filter effects. Internal references must resolve within the reviewed SVG. Use original project-authored geometry or separately reviewed licensed geometry; record provenance and hashes.

Linework has theme-aware semantic colors and a low-emphasis treatment in both Light and Dark. It must be visibly present in the final desktop and 320px overview screenshots, with masking/cropping permitted to protect text; not hidden behind an opaque section or enabled only in a test fixture. The Plan pins placement/viewBox/stroke/opacity/masking and theme-role mappings. Compare imagery-on versus imagery-off captures of the same region to prove painted strokes, and inspect the three motif families in source and rendered desktop evidence.

Decorative SVG is hidden from the accessibility tree, unfocusable, noninteractive and pointer-events none. It must not intercept controls, change document order/layout sizing or produce horizontal overflow. No animation is required; reduced-motion has no extra work to stop. Under forced-colors it may be hidden while all semantic content/focus remains usable. Missing imagery leaves the same content and behavior. Do not introduce a separate background preference or persistence.

### R6 — Accessibility on Actual Fonts and Imagery

Preserve keyboard/skip/nav/modal behavior, focus containment/return, inert background and accessible names. Verify uncut visible focus around angled controls, including the angled outline variant; require >=3:1 against adjacent painted surfaces. Normal text requires >=4.5:1 and qualifying large text >=3:1; required enabled control boundaries/icons require >=3:1. Disabled exemptions remain explicit. Source blue/green normal-size label pairs and faint text/control rules are not assumed to pass.

Measure text on the actual painted SVG/composited backgrounds and control rest/hover/pressed/focus/selected states in both appearances. Include source label sizes when deciding the large-text threshold. Averages over a background do not excuse a failing local region behind text. Use the accepted pixel/contrast tools with independent expectations and focused rendered samples.

At 320px and desktop, verify no page horizontal overflow, no clipped labels/focus or unreachable controls, 200% zoom/text usability and inherited reduced-motion behavior. Repeat the accepted line-height 1.5/paragraph 2em/letter .12em/word .16em spacing trials with actual bundled fonts. Reflow the source grid rather than retaining its 860px scrolling minimum. Also test a long configured product name and blocked-font fallback. Font completion must not cover/focus-shift controls or destabilize modal interaction.

### R7 — Narrow Guard Reconciliation and Regression Coverage

Existing controls forbid copied kit files/new assets, pin old component geometry and require Base UI dependencies/imports. This Goal authorizes the minimum reviewed change to those guards for the exact selected components, bundled font/notices, internal visual-role declarations and blueprint files and the exact approved Radix dependency/removal/consumer migration. Preserve protected source/template/manifest integrity and all unrelated checks. Do not add blanket source-copy, color-literal, dependency or SVG exclusions.

The Plan identifies old assertion → authorized replacement → independent test/negative control. Controls must reject an unexpected copied module/asset, an unapproved font or remote font URL, an unsafe SVG, and a normal-load fallback-font rendering falsely claimed as success. Dependency controls must accept exactly the reviewed Radix additions/Base UI removal, reject unapproved packages/version/category changes and package/lockfile mismatches independently, and report all applicable deterministic violation codes. Retain the established optional-dev justification rules and existing negative-control guarantees. Tests may replace Base UI-private selectors/focus-guard assumptions with Radix counterparts, but every behavior scenario and assertion guarantee must remain mapped and exercised. These are focused additions to the existing test surfaces, not a new validation system. Shared guard edits trigger relevant static/control checks in the same cycle.

Run the actual accepted full wrapper/harness against the reviewed tree, including root, intelligence-disabled and configured-rename browser coverage and applicable post-G003 CI checks. Preserve harness-status and post-run drift detection, cleanup and input binding. If final hosted evidence is inherited from accepted G003, obtain that evidence before completion; do not create another hosted workflow or weaken a gate because access is missing.

## Expected Artifacts and Evidence

- Source-derived consuming components, shell wiring, same-origin font assets/notices and original SVG/background treatment, only during approved execution.
- Reviewed `web/package.json`/lockfile delta, all migrated production/test consumers and scoped portal adapter, and exact dependency/import guard reconciliation for Radix Dialog/Slot and Base UI removal.
- Updated WEB-SHELL documentation with source lineage, exact-font loading, adaptation table, scoped-role/background behavior and known limits; narrow approved customization/guard updates.
- Existing browser/fixture test extensions and a goal-specific Result with AC1–AC10 checklist, failures/deviations, cycle/checkpoint accounting and cleanup.
- Exact baseline/input/source/asset hashes, full diff/path inventory and notices; protected kit manifest remains unchanged. Pin relevant selected source/guide bytes and reuse the existing full manifest rather than inventing another source registry.
- Before/after desktop/narrow Light/Dark shell and open-dialog screenshots; reference source comparison; font-request hashes/rendered-font reports; geometry/state, SVG, contrast/spacing, first-paint/hydration and scope evidence. Artifacts carry commit/tree/run identity and SHA-256 outside the repository under the accepted evidence convention. Screenshots wait for fonts/animations, while first-paint evidence uses the inherited separate earliest-paint mechanism.
- Full final wrapper/CI outcomes and resource start/end records. Logs/reports contain no secrets; destructive/negative mutations occur only in positively owned disposable fixtures/copies. No required failed/skipped scenario is represented as passed.

## Acceptance Criteria and Verification

| ID | Required outcome | Objective verification and evidence |
|---|---|---|
| AC1 | Six selected source modules are adopted and used with Radix, resolved rights and traceable migration. | Source/consumer hash and alteration table; ownership/upstream-notice record; diff review; production import/use inventory for R1; built page contains their rendered structures; exact approved Radix pins and all consumer migrations; no Base UI imports/dependency tree entries remain. No unused-file or values-only substitution. |
| AC2 | Source visual geometry/type/state treatments are delivered within reviewed exceptions. | Independent R2 reference table; desktop computed measurements within specified tolerance; screenshots/source comparison; all consumed button sizes/variants and dialog parts exercised. Angled 2px outline and source-sized production help trigger are explicitly checked. |
| AC3 | Bundled Oswald, Barlow Condensed and Roboto render in the real app. | Original/delivered hashes, metadata, axes/weight/coverage and notices; same-origin font response bytes match; clean Docker browser actual-font reports on heading/nav/button/dialog, kicker/bullet/meta and body/lead/description nodes. Blocked-font control is detected by the font oracle and independently proves readable/operable fallback. No host-only installation or family-string-only proof. |
| AC4 | Blueprint SVG imagery is present, safe and decorative in both appearances. | XML/source checks for R5 safety, payload and motif inventory; theme mappings; imagery-on/off painted-region evidence; desktop/320px Light/Dark screenshots; accessibility-tree/tab/pointer and missing-asset/forced-colors behavior. |
| AC5 | Light/Dark/System and first-paint/hydration behavior remain intact. | Inherited mode/storage/OS-change tests and delayed-framework-JS earliest-paint evidence on the changed production page; font/background loading does not substitute later screenshots for first-paint proof. Explicit modes ignore OS changes. |
| AC6 | Scoped overrides compose with source-derived components and imagery. | Independent expected-role mapping; partial Light/Dark, nested/sibling, runtime change and portal cases in both appearances; default scope source fidelity and overridden scope behavior verified. No Organization storage/API change. |
| AC7 | Accessibility passes with the delivered fonts/backgrounds and source geometry. | Rendered contrast at applicable states/surfaces, uncut focus, axe/keyboard/modal/inert/return-focus, 320px/desktop/200%/text-spacing, long-identity and reduced-motion tests. Explicit failure/no-skip for required measurements; negative occlusion/contrast controls remain meaningful. |
| AC8 | Copy/asset/geometry guards are reconciled without broad acceptance paths. | Reviewed old→new guard table and focused controls in R7; reject extra source/asset, unapproved/remote font, unsafe SVG and false fallback success for their intended reasons. Exact approved dependency delta and all inherited dependency-policy guarantees pass; controls reject unapproved additions, version/category changes, leftover Base UI and independent package/lock mismatches. No blanket package/import exception. |
| AC9 | Full inherited stack/browser/CI validation passes on the delivered tree. | Final wrapper pre/post hashes and preserved exit status; full actual-root and disabled stack checks, renamed identity browser checks and applicable accepted G003 CI evidence; no required FAIL/UNMET/skip or deferred hosted gate. |
| AC10 | Provenance, scope, cleanup and completion record are reviewable. | Exact reviewed path/input/source/asset inventory, protected skill manifest, source/font/license notices, screenshot/report checksums, Result AC mapping and shared B5/checkpoint; no schema/API/MCP/product changes or unowned resource removal; final resource inventory restored. Codex diff/evidence review and ACCEPT required. |

## Permitted Deviations and Completion Gate

Only R2's enumerated compatibility/accessibility adaptations are permitted, and their concrete substitutions require Plan review before execution. Do not reuse the prior font-fallback, compact-size or rectangular-outline choices as an automatic exception. Unexpected fidelity/rights/framework/token issues return to Codex; major architecture/product decisions return to the owner. No required acceptance evidence may be deferred to another Goal.

Completion requires AC1–AC10, the actual accepted validation gates and Codex ACCEPT. Passing screenshots/tests alone are insufficient. Reuse accepted controls proportionately; no new execution framework or extensive pre-approval demonstrations of unchanged controls. G003 continues independently; G048 is neither implemented nor accepted by this draft. G021 includes this foundation increment before releasing product features.

## Risks and Questions to Resolve Before Implementation

1. **Selected code provenance:** the supplied kit contains no complete license/origin record. The owner's instruction authorizes intended adoption but does not identify upstream obligations. Confirm original ownership and source lineage/notices before copying; do not invent a license or silently substitute an independent approximation.
2. **Exact fonts:** pin artifacts, notices, coverage, weights and any transformation in the Plan. Upstream family-directory references are not a binary hash; no remote runtime loading is allowed.
3. **Migration and fidelity:** the owner has decided Radix via ADR-001; no further decision is needed merely to use the two scoped packages. Pin versions/licenses, every consumer/API change, scope-container lifecycle and framework compatibility in the Plan. Preserve source appearance wherever no measured conflict exists. Additional primitive suites, package expansion or token-contract breaks require a separate review/decision.
4. **Background composition:** the motif/placement requirements are decided here; Claude proposes the exact static drawing and theme mappings in the Plan. Visual review checks it against those requirements, not an unspecified promise to resemble a blueprint.
5. **Accepted baseline/CI:** G003 acceptance/integration and its actual validation contract remain prerequisites. Reconcile exact paths/guards against that tree, without editing the currently running batch or treating pending evidence as accepted.

## Revision Record

- 2026-10-06 — Initial owner-requested G048 definition. Ten acceptance criteria, B5/checkpoint 3, actual selected source adoption and bundled fonts with a static blueprint-background extension. No implementation or Plan approval.

- 2026-10-06 — r2: apply owner-approved ADR-001. Supersedes r1's Base UI retention/no-Radix assumptions; scopes Radix Dialog/Slot dependencies, complete consumer migration/removal, portal lifecycle and unchanged behavior/control guarantees. Original r1 (not included) is preserved byte-exactly. AC1–AC10 and B5 remain; no implementation approval.
