> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G002 — Accessible web shell work specification

Goal: [G002-web-shell.md](../goals/G002-web-shell.md)
Epic: EPIC-001 — Repository and local platform
State: Plan r6 approved for implementation against the supervisor-pinned snapshot; implementation not yet accepted

## Authority and inspected baseline

Follow the Goal, repository governance, canonical architecture, budget policy, and roadmap. This Spec defines assessable outcomes; Claude's subsequent reviewed Plan owns component selection, files, dependency pins, layout details, and test tooling.

The clean inspected baseline is `b40507794f3ebcc0dc84cc66fbab8102c89978fb` (`G001 acceptance`). That commit records the accepted implementation and PROGRESS.md acceptance. G001's historical Result remains a pre-acceptance submission and must not be rewritten to describe this Goal.

Actual web state:

| Area | Already present | Remaining G002 obligation |
|---|---|---|
| Framework | Next.js 16.3.8, React/React DOM 19.3.0, strict TypeScript 5.9.3, App Router under `web/app`, npm lockfile | Retain the working framework/layout convention; add only needed UI/test dependencies |
| Styling | Tailwind 4.3.3, PostCSS integration, CSS variables mapped with `@theme inline` | Complete the tokens needed by installed components and test override behavior |
| Surface | Dynamic root page, configured title/product/domain, minimal main content | Accessible responsive generic shell and a working interaction |
| BFF | `server-only` config/core client; `/api/health`; `/api/ready` with bounded core readiness fetch | Preserve this boundary and behavior; add regression evidence, not a replacement API |
| Checks | Docker web check target runs ESLint and TypeScript, production standalone runtime | Add meaningful behavioral/browser validation in containers |
| Components | No `web/components`, shadcn config, or Base UI dependency | Repository-owned minimum shadcn/Base UI component foundation |
| Provenance | Original output `55bb52881a31a589`; three G001 registered customizations | Attribute new web customizations and keep consumer validation accurate |

G002 remains the next Goal because its UI obligations are unmet and need no additional backend authority. G003 stays independently eligible for later planning; no detailed G003 requirements are created here. SECURITY.md and DECISIONS.md are empty at this baseline and add no unstated ADR approval.

## Requirements

### R1 — Extend the accepted web foundation

Use the existing `web/app`, `web/lib`, npm lockfile, Next.js/React/TypeScript, Tailwind/PostCSS, Docker build/check/runtime, and request-time identity configuration. Do not re-scaffold the application, relocate it to `web/src`, or introduce a parallel UI framework/package manager. Keep generic layout and UI primitives separate from future domain components; create no empty domain modules.

Initialize shadcn/ui for this existing project with **Base UI primitives**. Commit the consumed component source and project configuration that makes future additions unambiguous. Use only components actually needed by this Goal; do not install an entire catalog. Generated components must not introduce product literals, raw brand colors, runtime policy, or datastore clients.

Record the CLI/registry/component provenance and exact dependency versions in the Plan and Result. Pin newly added direct runtime/development dependencies and container tools; preserve lockfile integrity. Installation may access registries during containerized setup/build; execution-time checks must use installed dependencies without downloads. No unrelated framework upgrade is authorized. The official [shadcn CLI](https://ui.shadcn.com/docs/cli) and [configuration documentation](https://ui.shadcn.com/docs/components-json) describe library selection and project aliases; Claude selects and reviews the exact compatible setup in its Plan.

### R2 — Small, functional, accessible shell

The existing `/` page must render a reusable generic application frame with configured product identity, meaningful navigation, a main landmark, and neutral foundation content. All navigation targets must exist; in-page destinations are sufficient. No placeholder product routes, account/Organization switchers, login affordances that imply implemented authentication, or domain dashboards are authorized.

Include a visible-on-focus skip link that moves focus to main content, clear heading hierarchy, semantic landmarks, and accessible names for controls. Interactive elements use proper links/buttons and keyboard behavior, with visible focus and no color-only meaning.

Consume at least one working shadcn/Base UI overlay interaction (for example generic shell help) to demonstrate the primitive boundary. It must open from the keyboard, have an accessible name, place focus appropriately, contain focus while modal, close on Escape and its close control, and return focus to the trigger. The Plan chooses the smallest suitable component set and useful generic content.

At 320 CSS-pixel width and a desktop viewport, content and controls remain usable without unintended horizontal page scrolling or obscured focus. Verify 200% text/zoom usability. Supported animations respect reduced-motion preferences. Avoid external fonts/assets and invented detailed product-brand decisions; use the supplied visual guidance within R9's constraints.

### R3 — Semantic tokens ready for later Organization styling

Retain the existing semantic names (`background`, `foreground`, `primary`, `primary-foreground`, `accent`, `muted-foreground`, `border`) and extend the contract only for installed component needs, including accessible focus styling. Map tokens consistently to Tailwind utilities; document names, purpose, default values, and override scope. Colors in generic shell/components resolve through semantic variables/utilities, not embedded palette literals or component-specific product themes.

Demonstrate that the same component source supports the default styling and two distinct scoped override fixtures. Computed browser styles must change for representative surface/text/action/focus tokens, while an adjacent scope keeps its own values. Include a portalled overlay in the override evidence so portal placement does not silently revert it to the global defaults. Changing an override requires no rewrite or rebuild of component source.

Fixtures are non-persisted test styling data, not Organizations, membership, authorization, or production hostname resolution. No runtime route, query parameter, header, local storage identifier, or development authenticator may claim to select an authorized Organization. A dedicated container/test fixture or browser-applied fixture may exercise the styling contract without exposing a production theme editor/demo endpoint.

Default and override fixtures must meet relevant WCAG 2.2 AA checks for the delivered surface: text contrast (4.5:1 ordinary, 3:1 qualifying large text), non-text control contrast (3:1 where required), keyboard operation, labels/landmarks, visible unobscured focus, and reflow. Validate actual computed colors; an automated audit alone does not prove focus or token isolation. Leave no unresolved applicable automated accessibility violations in the delivered surface; explain any tool false positive with evidence rather than disabling the rule globally. Do not claim full-product accessibility certification.

Apply these requirements in both resolved Light and Dark appearances, including default styling and both scoped override fixtures. Retain the seven existing semantic names; their values may be revised to map the supplied design language to accessible Light and Dark palettes under R9. Document changes from the inherited values and how mode palettes and scoped overrides compose so a mode change preserves scope isolation and portal styling.

### R4 — Preserve runtime identity and thin BFF authority

Product title and visible identity continue to come from `PRODUCT_NAME`; any displayed platform domain comes from `PLATFORM_BASE_DOMAIN`. Missing required configuration continues to fail closed. Preserve request-time resolution: restarting with renamed values using the same image must change title and UI without source edits or a rebuild. Do not bake identity into client bundles or use `NEXT_PUBLIC_*` for internal backend URLs.

Keep `web/lib/config.ts` and the core bridge server-only. Browser code may call same-origin web endpoints only; it must never call core/intelligence/datastores directly or receive internal URLs, backend credentials, or privileged error details. G002 needs no new backend fetch contract. Maintain `/api/health` liveness and `/api/ready` core-dependent readiness, including existing 200/503 and no-store behavior.

No domain logic, policy, tenant authority, datastore configuration, or AI provider calls belong in generic components. Preserve web/core/intelligence network separation and loopback/default ports. Validation requires both source/import boundary checks and real-browser request evidence; visible product/domain configuration is public display data, not an authentication secret.

### R5 — Meaningful, repeatable Docker validation

Add a committed browser-based test path against the built web application, runnable with documented repository commands in Docker and no host Node/Python/browser dependency. Install pinned browser/tool dependencies at image build/setup time; tests must not rely on live external services or installation downloads. A browser container may join an isolated test network to reach web; this does not justify expanding production web network access.

Required browser scenarios include shell landmarks/navigation, skip-link focus, keyboard traversal, overlay focus/close/return behavior, responsive/reflow behavior, token overrides including portals and adjacent-scope isolation, configured metadata/UI, and absence of hydration/uncaught browser errors. Record an automated accessibility audit plus explicit keyboard/computed-style assertions; a screenshot alone cannot satisfy these requirements.

Keep containerized web lint/type checks, production build, existing core/intelligence checks, and secret scanning. Document test commands, tested viewports/fixture values, tool/browser versions, and reports. Required scenarios cannot be skipped or reduced to static render snapshots. If a browser check is unavailable, the associated criterion remains unmet.

### R6 — Consumer provenance and bounded regression handling

Keep the accepted Skill, templates, fixtures, original manifest, and closed G000/G001 artifacts unchanged. New web files are G002 additions; changed rendered files are explicit consumer customizations. Update the project customization register with exact paths/reasons and demonstrate repeat plan/render preserving files with only expected conflicts/proposals.

Reconcile current project-harness assumptions that would falsely reject intended web changes (including the web package/lockfile identity check and disabled-variant composition). Preserve exact-path scope checking, protected content checks, dependency pin checks, source identity checks, secret canaries, isolation negative controls, and the meaning of existing checks. Do not suppress failures with broad directory exceptions or claim unmodified-template evidence for the evolved consumer.

The intelligence-disabled variant must carry the same applicable G002 web customizations/additions, preserve its own manifest, and still start/check without intelligence, Python service, or AI settings. Adaptations belong only to project-owned tooling and must be enumerated in the reviewed Plan. Historical G001 reports/allow-lists remain historical rather than being rewritten for new scope.

### R7 — Identity invariants and foundation handoff

Durable product/domain entities use UUID canonical identifiers. Slugs and names are alternate display/routing identifiers and convey no authority. Infrastructure/framework tables may use another justified strategy; existing goose metadata is not a product entity. G002 introduces **no** domain tables, migrations, Organization storage, identity API, or identifier-generation subsystem. Test CSS selectors and framework DOM IDs are not durable entity identity.

Document which web obligations are completed, inherited, and deferred. Actual Organization branding data and trusted resolution remain G016/G017 after their security dependencies. Preserve all other foundation stages; no G002 UI evidence establishes tenant isolation, authentication, MCP exposure, or product-feature permission. The G000–G021 barrier remains closed.

### R8 — Light, Dark, and System appearance

Provide a visible, keyboard-accessible shell control with named **Light**, **Dark**, and **System** choices and an identifiable selected choice. **System is the default** when no valid preference is saved. System resolves appearance from the browser/OS `prefers-color-scheme` preference and follows changes while the page is open. Explicit Light or Dark overrides the system preference and stays selected when the system preference changes; returning to System immediately follows the current preference.

Persist the selected mode across reloads in browser-local storage only. This is a display preference, with no account, Organization, server cookie, schema, or authorization meaning. Missing or invalid saved values resolve to System; unavailable storage must not crash or prevent in-session selection. Store the selected mode, preserving System as a choice rather than replacing it with its currently resolved palette.

Apply the resolved palette consistently to the shell, consumed components, focus indicators, overlays, and native controls via an appropriate `color-scheme`. Avoid a visibly incorrect initial palette on normal supported loads and introduce no hydration errors. The Plan must explain first-paint initialization, server/client responsibilities, and storage-failure behavior; broad hydration-warning suppression is not evidence of correctness. Keep runtime product identity and the server-only BFF unchanged in behavior. Any necessary layout edits or new dependencies must be explicitly enumerated and reviewed in the revised Plan, with pins, provenance, customization registration, and boundary checks updated accordingly.

Use the same component source for both appearances, with semantic token values supplying the difference. This is generic foundation appearance support, not an Organization branding editor or a detailed product design language. Required Docker browser evidence covers fresh System sessions under both emulated system preferences, live system changes, explicit choices overriding the system, returning to System, reload persistence, invalid/unavailable storage, correct initial appearance, native color scheme, and zero hydration/uncaught errors. Both resolved appearances must satisfy R2/R3 keyboard, contrast, accessibility, responsive, scoped-override, and portal obligations; document the test matrix rather than assuming Light evidence establishes Dark behavior.

### R9 — Use the supplied Vestive design language within foundation scope

Use the skill for visual guidance only. Ignore its marketing copy and content fundamentals, including voice/campaign prescriptions, commercial vocabulary, navigation labels, CTAs, and example text. Ignore its dark-only rule: dark ground is not mandatory, and the skill must not force Dark mode or override the user's appearance choice. Write only the generic foundation copy required by R2; retain configured product identity under R4 and Light/Dark/System behavior with System default under R8.

Claude must read and use skills/vestive-design-skill/SKILL.md (not included), its README, `styles.css`, `motion.js`, and `examples/reference.html` when planning the shell's visual treatment. Apply the supplied flat geometry, visible structural rules, flush-left hierarchy, typography roles, and restrained color roles where relevant to the delivered shell. Describe actual applied choices, not merely that the skill was read. Keep shadcn/Base UI behavior and reusable semantic-token components; map source colors and typography/layout roles into the web foundation rather than hardcoding brand values throughout component source.

The supplied skill is a dark marketing reference, not a complete accessible two-mode application theme. Derive a compatible Light appearance and document both palettes, contrast adjustments, and omitted or adapted treatments. Existing semantic token names are preserved; inherited placeholder values and the preset's radius may change where justified by this supplied reference. All contrast, reflow, focus, reduced-motion, and Light/Dark/System requirements remain mandatory. Angled controls must not clip labels or focus indicators; narrow shell content must reflow rather than adopt the marketing example's fixed-width scrolling grids. Use system/local font fallbacks for the supplied typography roles; no font binaries are currently included, and the Google Fonts import in the skill is not authorized for the application.

G002's explicit requirements take precedence over conflicting skill instructions. Adapt selected source styles into repository-owned semantic styling; do not import the skill stylesheet wholesale with its remote-font import, global dark-only overrides, raw component colors, or unrelated marketing primitives. Document this adaptation of the skill's instruction to link its CSS/load motion. Signature ribbon/node-field effects are not required for this bounded shell; if proposed, enumerate their minimal source/assets and validate reduced motion, accessibility, scope, and both appearances. Existing generated icon choices need not introduce another icon dependency solely to match the marketing reference; record any such adaptation.

The shell continues to display configured product identity and generic foundation content. Do not copy Vestive business names, logos/subsidiary marks, pricing, testimonials, service claims, marketing CTAs/navigation, or product terminology into the application as fixed identity/content. No marketing site, Organization branding storage, tenant selector, or product workflow is authorized. Keep the design skill's source files unchanged during implementation; enumerate any selected copied/adapted resources and their provenance in the Plan/Result and project docs. The newly relocated design skill is a separate product-owner input; do not confuse it with the protected accepted project-archetype Skill or use it to authorize edits to that Skill.

Provide a source-to-implementation mapping and adaptation table in WEB-SHELL documentation/Result, plus built-app screenshots of the actual shell and open overlay in Light and Dark. Visual review must establish observable use of the supplied language; screenshots complement the required computed-style and browser behavior evidence rather than replacing them. Include this review in the revised Plan and reassess B5 feasibility before implementation.

## Architecture and security invariants

- Core remains authoritative for facts, permission, tenancy, MCP, and public AI streaming; web stays UI/thin BFF and intelligence stays internal inference.
- No new public API/MCP contract, identity bypass, trusted tenant selector, schema, seed records, datastore authority, or external account is introduced.
- Secrets stay in scoped container environments and ignored local files. Tests, browser bundles, screenshots, traces, logs, manifests, reports, and committed examples contain no credentials. Keep secret scanning redacted and artifact capture bounded.
- Styling fixtures cannot establish security context. Never treat a slug, name, DOM ID, theme label, or browser-selected value as canonical domain identity or authorization.
- Preserve pre-existing local data/resources; disposable validation copies contain the current accepted-plus-G002 implementation, not pristine templates. Runtime rename/destructive cleanup never targets existing developer state.

## Expected artifacts and allowed areas

- Web shell, minimum reusable UI/layout components, shadcn project configuration, semantic styles and documentation, dependency/lockfile updates, and browser tests under `web/`.
- Bounded Docker/check/test configuration and project validation/provenance-register changes under `infra/`, as justified in the Plan. No production datastore/network/port changes.
- Focused setup/foundation documentation updates under README.md and `docs/foundation/`.
- Goal-specific Claude Plan and Result under `management/v0/plans/G002-web-shell-plan.md` and `management/v0/results/G002-web-shell-result.md`, plus a goal-specific exact-path scope artifact if needed. These are future deliverables, not created by this Spec.
- Supervisor-owned progress updates at handoffs/review. Preserve canonical governance/architecture, Skill content, original manifest, and all closed-goal artifacts. The reviewed Plan must name exact intended files; these areas are not blanket authorization for unrelated edits.

## Acceptance criteria

| ID | Required outcome | Required evidence |
|---|---|---|
| AC1 | Existing framework is extended with repository-owned shadcn/Base UI primitives and attributable dependencies. | Plan/Result enumerate component and CLI/registry provenance, exact pins, lockfile resolution, component imports/config, and production build. No re-scaffold or unrelated framework change. |
| AC2 | Root shell is usable and generic. | Built-app browser checks for configured title/identity, working navigation, landmarks/headings, usable 320px and desktop layouts, 200% text/zoom, and no hydration/uncaught errors; no domain/authentication/tenant-selection surface. |
| AC3 | Keyboard and accessibility behavior is correct. | Real-browser skip-link focus and keyboard traversal; overlay open/focus containment/Escape/close/return focus; visible unobscured focus; scoped automated accessibility results and computed contrast checks for default and override fixtures; reduced-motion evidence if animations exist. |
| AC4 | Semantic overrides work without component rewrites or scope leakage. | Documented token contract; default plus two overrides on identical components; computed surface/text/action/focus styles and visual evidence; adjacent scope unaffected; portalled overlay retains the intended styling scope. |
| AC5 | Runtime identity, server-only BFF, and health behavior are preserved. | Disposable runtime rename on unchanged image IDs; browser title/visible identity update; boundary import checks and network capture show no direct internal-service requests or leaked internal config/secrets; `/api/health` 200 and `/api/ready` 200/503 with no-store under core availability/outage. |
| AC6 | Tests and project checks run repeatably in Docker. | Exact documented commands pass for web lint/types, production build, browser scenarios, core/intelligence existing checks, and secret scan. Record pinned tooling, zero host-runtime dependency, no skipped required checks, and execution-owned cleanup. |
| AC7 | Evolved consumer retains provenance and bootstrap regressions. | Exact customization register, unchanged original manifest/Skill/protected files, repeat plan/render preserving content, passing project isolation/default-port/migration regressions, and disabled-variant web parity/start/check. Goal-specific scope checks catch unexpected files. |
| AC8 | Scope, identity invariants, handoff, and budgets support acceptance. | Full tracked/untracked diff against the recorded execution baseline; no migrations/domain schema or early product features; UUID/slugs/infrastructure-key distinctions documented; Result maps AC1–AC10, inherited/new/deferred work, limitations, cycle counts/checkpoint, and Codex acceptance. |
| AC9 | Light, Dark, and System modes work; System is the default. | Accessible choice control; real-browser R8 scenarios for system tracking, explicit override, return to System, reload persistence, storage fallback, initial appearance/native color scheme, and no hydration errors. R2/R3 evidence covers both resolved appearances, default plus both scoped fixtures, and portalled overlays; document palette values and composition. |
| AC10 | The delivered shell visibly uses the supplied Vestive design language within G002 constraints. | Plan/Result and WEB-SHELL docs identify source mappings, applied structure/type/color choices, Light adaptation, accessibility adjustments, resource provenance, and deliberate omissions. Actual-shell/overlay screenshots in both appearances support visual review; semantic-token, contrast, reflow, focus, no-external-assets, configured-identity, and generic-content checks remain passing. Design skill source is unchanged. |

## Required validation and evidence handling

Claude's Plan must map each criterion to commands/assertions, distinguish browser behavior from source checks, identify disposable fixtures and networks, and show how existing harness assertions remain meaningful. Reuse bootstrap coverage; do not rebuild migration/intelligence/isolation machinery to test a shell.

Run a final integrated project validation across applicable static, provenance, root, migrations, and variants coverage after the consumer assertions are reconciled, together with the committed G002 browser suite. The current harness invocation and G001 allow-list are historical inputs, not suitable unchanged scope authorization for G002. Use a G002 base and exact allow-list, with supervisor changes handled only as explicitly inspected exact-path exceptions.

Documented root start/check commands must still work as written. Where an actual-root trial would touch existing developer resources, preserve them and report the unmet gate rather than deleting them; seek a reviewed resolution. Overrides/outages use disposable owned projects. Run no live AI/provider calls. Reports stay outside the repository unless explicitly listed as safe deliverables; sanitize captures and record cleanup. Successful validation does not itself authorize completion.

## Explicit non-goals and review gate

No detailed G003 Spec or CI implementation, new brand-design program beyond applying the supplied skill, marketing site, tenant/theme administration, authentication, Organization schema, hostname resolution, policy/RLS, product workflows, production deployment, or accepted-Skill revision is included. No change to framework/infrastructure keys is required.

The product owner's submission of Plan r1 for Codex review released this Goal and Work Spec for planning. Codex approved Plan r6 on 2026-10-05 against `/tmp/g002-r6-approved-inputs.sha256`, with its own digest supplied in the supervisor approval. The pinned Plan and external S0a script hashes and all S0a/S0b/S0c gates are part of that approval. Changes outside those reviewed inputs or scope require a further handoff; Plan approval does not accept implementation.

## Revision record

- 2026-10-05 — Initial draft based on committed accepted G001; no implementation authorized.
- 2026-10-05 — Planning release recorded in response to the product owner's r1 review request; no independently verified earlier approval is asserted. Requirements, acceptance criteria, and non-goals are unchanged. Plan r1 requires correction before implementation.
- 2026-10-05 — Product-owner scope addition: R8/AC9 require Light, Dark, and System modes with System default, browser-local persistence, accessible selection, and validation in both resolved appearances. R3 clarifies palette/scoped-override coverage; all earlier criteria remain required. Claude must revise the Plan and feasibility forecast before implementation approval.
- 2026-10-05 — Product owner requires Vestive skill use: R9/AC10 add observable visual integration and documented adaptations/provenance. R2/R3 and non-goals reconcile the prior neutral-only direction; existing token names remain stable, while placeholder values may change for the supplied language and accessibility. No implementation or design-skill source edits are authorized by this planning revision.
- 2026-10-05 — Product owner clarifies: ignore the skill's marketing copy. R9 explicitly limits use to visual guidance; marketing content fundamentals are not shell-copy requirements.
- 2026-10-05 — Product owner also excludes the dark-only rule. R9 explicitly preserves both resolved appearances and the R8 default/selection behavior; the skill cannot force Dark mode.
- 2026-10-05 — Codex reviews r2 and retains all R1–R9/AC1–AC10 requirements. Goal budget is revised to B7 with checkpoint after cycle 4; two bounded Plan corrections remain before implementation authorization. No acceptance criterion is weakened.
- 2026-10-05 — Codex r3 review retains R1–R9/AC1–AC10 and B7. First-paint planning now meets the earlier correction; dependency reason-set consistency and fail-closed snapshot/final-run control flow remain Plan corrections, without changing requirements or authorizing implementation.
- 2026-10-05 — Codex r4 review retains R1–R9/AC1–AC10 and B7. Dependency control sets and first-paint timing are coherent at planning level. The starting gate must be executable before the currently scheduled wrapper creation; this is a bounded Plan correction under R5/R6, not a requirements change or implementation authorization.
- 2026-10-05 — Codex r5 review retains R1–R9/AC1–AC10 and B7. The bootstrap ordering now resolves the prior missing-wrapper issue. The external gate must validate evidence containment before creating directories and suppress Git's optional index writes; these are bounded Plan corrections, with no application changes or implementation authorization.
- 2026-10-05 — Codex approves Plan r6: the evidence-path and optional-index-write defects are resolved in its exact pinned external S0a script. R1–R9/AC1–AC10, B7/checkpoint 4, architecture/security constraints, UUID identity and the G021 barrier remain unchanged. Claude may implement only against the approved snapshot, with all gates and required evidence retained.
