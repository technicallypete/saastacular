> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G069 — Adopt the complete, unaltered Vestive Brand kit: Implementation Plan (r1)

Revision: **r1, 2026-10-26 UTC — bounded correction of r0 after the Codex review
G069-plan-r0-codex-review-20261026 (not included) (CORRECTION LOOP: four items,
three returned decisions).** Author: Claude (Fable 5.1, high). r0 was a single lead planner integrating three read-only
research lanes (dependencies/build, themes/branding/portals, consumers/checks); r1 changes only the sections the
[r1 correction record](#r1-correction-record) lists and preserves everything else. **This revision releases nothing:
no install, build, Docker object, application edit, commit to any other file, push or MR.** Implementation begins only
after Codex Plan ACCEPT and recorded execution release. Budget **B5, checkpoint after attempt 3; zero attempts spent**
(§8.3 states the attempt boundaries).

Goal: [G069](../goals/G069-unaltered-brand-kit.md) (sha256 `3389b868b27aa61f32a50265ec45faf7bc5dd494edbd719dadf89a1166e8ed3b`).
Spec: [r0](../specs/G069-unaltered-brand-kit-spec.md) (sha256 `7289f465b136fb3c56e00422bff2efbd635a3984d661ccd26db9c14db8fd6d18`).
ADR: [ADR-001, G069 owner amendment 2026-10-26](../../../adr/ADR-001-radix-vestive-components.md#g069-owner-amendment--unaltered-kit-adoption-2026-10-26).

## 0. Bindings and verified inputs

| Binding | Value (verified in this planning checkout, read-only) |
|---|---|
| Staging base (r0) | `29dded769b14992e27de9150f6306790385c2b5f` ("docs: specify G069 unaltered brand kit adoption and release planning"); PROGRESS sha256 `cc296dfe6ebac77c7daa6672c24fcdc7be5d11632b67c5f78d8ccd44d13c330e` |
| Staging base (r1) | `de74ccbc312d350d4ef4f3bbeae24ffbe8d2e2ef` ("docs(G069): review plan r0 and return bounded corrections"); PROGRESS sha256 `d38488a43a287a9897982fbd96d89e04090273b6d7f95bf918509daec76279f6`. Delta `29dded76` → `de74ccbc`: one commit, **4 management files only** (`management/SUPERVISOR-HANDOFF.md`, `management/v0/PROGRESS.md`, the imported r0 Plan, the Codex review; 570 insertions, no executable change), so every line number, hash and tree fact below verified at `29dded76` holds at `de74ccbc` (`web`, `infra`, `core`, `.agents` trees unchanged). The r0 Plan at the base is sha256 `50991dafacc24ac4db078447e112e6df145f21d0f72ab33f57108f81594fd540`. |
| Goal baseline → staging delta | `57d40b8e` → `29dded76`: 12 files, management/docs/skill-record only; the skill tree differs only in `ADOPTION.md` (G069 note) and `LOCAL.sha256`; **`kit/src` is byte-identical at both** |
| Kit source | `.agents/skills/vestive-brand/kit/src`, Git subtree `b1be1af48bf92ae84a6ce077b0042592ca0b2700` (`git rev-parse HEAD:.agents/skills/vestive-brand/kit/src`) |
| Regular files / bytes | **69 / 241557** (`find -type f`; 54 `components/ui/*.tsx`, 8 `components/vestive/*`, `hooks/use-mobile.ts`, `lib/utils.ts`, `globals.css`, 4 `styles/*.css`) |
| Source-manifest SHA-256 | **`597f2aac251987be26050108dc6affe3a39a5594e7bfc7e2744de2fb5829a6ee`** recomputed with the Spec §2 algorithm (sorted relative POSIX paths, lines `<sha256>  <path>\n`); `UPSTREAM.sha256` and `LOCAL.sha256` each hold the same 69 `kit/src/` rows |
| Planning checkout | r0: isolated worktree `agent-<id>`, created at `main` `b40d4a4d`, moved by `git checkout --detach 29dded76` (clean before and after). r1: isolated worktree `agent-<id>`, created at `main` `b40d4a4d`, moved by `git checkout --detach de74ccbc` (HEAD verified, status clean); this file is its only change |

**Lane claims verified and corrected (all other cited line numbers were spot-checked and hold at `29dded76`):**

1. Lane 3 recommends the `web/` root to avoid the B-NAMING flip. The Spec prefers `web/src`; `web/src` is listed as a
   "redundant path" only by the harness (`run.sh:850-852`), while ARCHITECTURE §37A itself draws `web/src/` and line
   4805 accepts natural framework subdirectories. Decided for `web/src` (§1.1 S1) — one harness line, not a naming
   exception of the architecture.
2. Lane 1's "src must win so the kit's imports never land on old app modules" is moot: after cutover no `@/…` specifier
   resolves under both roots (`web/components/ui/*`, `web/lib/utils.ts` are deleted), so only multi-candidate `paths`
   support matters, not ordering.
3. Lane 3 says `web/src` would be an unmapped CI path; `profile-paths.tsv` row `runtime dir web/` classifies it as
   `runtime` (gate profile `integration`), not `unknown`. The MR still selects `full` for the optional gate through
   its foundation paths (`web/package.json`, `web/app/globals.css`, `web/components/layout/`, `web/lib/`, tests).
4. Lane 2's `RuledGrid` "inline `style` can override `minWidth` but not the column count" is right; lane 3's
   "`className` grid-cols override" is wrong — `gridTemplateColumns` is an inline style (`layout.tsx:41`), so a class
   cannot win. The consumer `style` prop can (§2.7).
5. Kit Light `--input` is the same value as `--border` (`oklch(0.922 0.005 34.3)` ≈ `#e8e4e3`), so the resting field
   fill `bg-input/50` on white is ≈1.06:1 (lanes cite 1.12:1 from G067's WEB-SHELL table; both are below 3:1 — the
   decision D-A2 is unchanged).
6. Lane 3's "B-BRANDING-IT/CATALOG via the migrations suite" has a narrower selector: `branding_checks`
   (`run.sh:4175-4191`) runs exactly `go test -tags integration ./internal/tenancy/access/ -run '^TestBranding$'`
   (23 subtest results) and `./internal/migrate/ -run '^TestBrandingCatalog$'` (3) on a private PostgreSQL+MinIO; the
   Plan uses those Go selectors in a private per-run database run, not the whole `migrations` suite (§6.1 row 5).
7. Lane 2's next-themes `enableColorScheme` default adds `style="color-scheme:…"` to `<html>`; the kit's `theme.css`
   already sets `color-scheme` per selector, so the Plan disables it (fewer `<html>` attributes to re-pin).
8. Lane 1's B8 ("`no-img-element` fires on `site-header.tsx:28`") also applies to `components/vestive/content.tsx:106`
   (lane 3). Both are kit files; one scoped override covers them.
9. `appearance-control.tsx` is the only app file besides `lib/appearance.ts` reading storage (identity guard
   `run.sh:5854`); with next-themes neither app file touches storage, so that allow-list becomes empty, not retargeted.

**r0 claims found wrong while correcting (r1; each verified at the base):**

10. r0 §6.1 row 3i said "all 14 identity projects". `web/playwright.config.ts:105-128` declares **16**: `identity-light`,
    `identity-dark`, `domain-platform`, `domain-custom`, `mcp-light`, `mcp-dark`, `onboarding-light`, `onboarding-dark`,
    `membership-light`, `membership-dark`, `enrollment-light`, `enrollment-dark`, `signup-light`, `signup-dark`,
    `labels-light`, `labels-dark` (the same 16 B-IDENTITY-E2E names at `run.sh:6255`). Lane 3's "14" was carried over.
11. r0 §4 P0.1 step 3 was not executable: after the S1 alias flip, `@/components/ui/{button,input,select,table,textarea,dialog}`
    and `@/lib/utils` resolve to the kit for every old consumer, so the whole-app `tsc`/`next build` (and therefore the
    P0.2 production build) cannot pass before L1–L3. r1 §4 stages the legacy modules explicitly (S21) instead.
12. Lane 3 / r0 §6.1 called `off` "unaffected". It is affected: `variants_off` (`run.sh:5668-5788`) renders the archetype
    and copies only the `APPLICABLE` paths (`run.sh:226-300`: `web/components`, `web/lib/*.ts`, `web/app/<routes>`,
    `web/tests`, …) — **neither `web/src` nor `web/tsconfig.json` is in that list**, so without a harness change the
    disabled copy cannot resolve the kit and B-BROWSER-OFF (`run.sh:5777-5782`, the full `BROWSER_MATRIX_FULL` suite
    on the copy) would fail. Its ports `PORTS_OFF=(3500 8580 - 5932 6879 9500 9501)` (`run.sh:173`) do not collide with
    M1, unlike `root`. r1 §3.4/§6.1 handle it.
13. r0 S11 wrote the destructive tints as "10 % and 20 %". The kit's used `destructive` Button variant tints are per
    appearance: Light `bg-destructive/10` → hover `/20`, **Dark `/20` → hover `/30`** (`button.tsx:13`).
14. r0 S11 withdrew every brightness factor and validated only `primary-foreground` on `primary`. The kit `CtaButton`
    (`components/vestive/cta-button.tsx:8,12`) is `bg-blue text-white` with `hover:brightness-110 active:brightness-90`:
    a **fixed white** label on `--vx-blue` (saved `primary` under D-M1), never `primary-foreground`. r1 S24 decides where
    the CTA may render and S11/§2.6 carry its actual rest/hover/active contract.
15. r0 treated the existing accepted saved-branding fixtures as compatible by hex. Computed under the kit's actual states
    (r3 model): the seeded custom-host palette `infra/identity/mcp-seed.sql:109-110` = `fixtureBranded`
    (`web/tests/fixtures/token-scopes/fixtures.ts:91-126`) paints white on `primary` `#0f6b5c` at the kit's `hover:bg-primary/80`
    over its `background` `#f1f5f4` at **≈4.27:1** (rest 6.4); `fixtureGreyPrimary` `#696969` paints **≈3.60:1** on hover
    (rest 5.49); and in Dark the seeded `primary` `#3fb8a3` would carry a fixed-white CTA label at **≈2.44:1**. The r1
    compatibility proof (§4 P0.3) and S23's resolution rule cover them; the computed values are confirmed by measurement.
16. r0 overlooked the only app placeholder, `web/components/branding/branding-form.tsx:70` `placeholder="Platform default"`:
    kit placeholders are `muted-foreground` on the `bg-input/50` fill, ≈3.35:1 on the platform Light default and ≈2.9:1
    on the seeded palette. S22 removes the placeholder (visible description instead) rather than adding a pair no palette
    can pass while N3 holds.
17. r0 S20 said "every commit the hosted fast job builds typechecks". Hosted fast runs once per MR pipeline on the MR
    head, not per commit; r1 S20 claims only what is executed: every commit integrated on `g069` passes the coordinator's
    stack-free `tsc`/`eslint`, and the MR head passes hosted fast.

## 1. Settled decisions, unproven assumptions, decisions for owner/Codex, stop conditions

### 1.1 Settled (within Spec/ADR authority; Codex may redirect at Plan review)

| # | Decision | Reason / authority |
|---|---|---|
| S1 | **Source root `web/src`**: exactly the 69 files at their kit-relative paths, nothing else under `web/src`. `web/tsconfig.json` `paths` → `{"@/*": ["./src/*", "./*"]}`; `include` already covers `**/*.ts(x)`. No app file moves. `web/src` leaves the B-NAMING redundant list with the recorded reason (Spec §2 root; ARCHITECTURE §37A tree). `web/tsconfig.json` becomes a registered customization (render register 31 → 32). | Spec §2 "prefer `web/src`"; keeps the kit a delimited tree for fidelity, literal and lint scoping |
| S2 | **Dependency delta** (exact pins, `--save-exact`; final versions taken from the proof's actual resolution, §4): add runtime `cmdk 1.1.1`, `clsx 2.1.1`, `tailwind-merge 3.7.0`, `next-themes 0.4.6`, `sonner 2.0.8`, `react-day-picker 9.14.0`, `embla-carousel-react 8.6.0`, `input-otp 1.5.0`, `react-resizable-panels 4.15.0`, `recharts 3.10.1`; add **dev** `shadcn 4.22.1` (only for `@import "shadcn/tailwind.css"`; no CLI run); **remove `cn 0.4.0`** (its only consumer `web/lib/utils.ts` is replaced by the kit's). Keep `@base-ui/react 1.9.0`, `class-variance-authority 0.7.1`, `lucide-react 1.52.0`, `tw-animate-css 1.4.0` and all Next/React/TS/ESLint/Tailwind/Playwright pins. No `date-fns`, no `@radix-ui/*` or `radix-ui` direct. The archetype template under `skills/` is **not** touched (G001 base; `web/package.json` is a registered customization; B-SKILL-UNCHANGED). | Spec §3 lists these candidates and allows the installed CSS package; registry metadata read by lane 1 |
| S3 | **cmdk Radix closure**: `cmdk` → `@radix-ui/react-dialog` → the 15–16 `@radix-ui/*` utilities plus `aria-hidden`, `react-remove-scroll(-bar)`, `react-style-singleton`, `use-callback-ref`, `use-sidecar`, `get-nonce`, `detect-node-es`. Recorded by exact name/version in `lib.sh` (`DEPS_RADIX_VIA_CMDK`) and in WEB-SHELL.md with the reason per package; no app file imports `@radix-ui/` (kit `command.tsx` composes the Base UI Dialog; Radix is reached only inside cmdk's module graph). | ADR amendment §2; Spec §3 |
| S4 | **CSS chain**: `web/app/globals.css` stays the app entry (`layout.tsx` import, `components.json` `tailwind.css`, render register) and becomes: `@import "../src/globals.css";` (the kit chain unchanged: tailwindcss, tw-animate-css, shadcn/tailwind.css, theme.css, vestive.css) followed by app-owned sections only: the five `@font-face` rules (unchanged bytes), the app-owned variable corrections (`:root{…}` / `.dark{…}` after the import, so source order wins at the kit's own specificity, §2.4), the blueprint backdrop rules, the global reduced-motion rule; **no field rule** (S22 uses `className` props, r1). The current ejected block (lines 1–640), `@custom-variant dark`, the `--light-*/--dark-*` sources, `[data-token-scope]` resolution and the `data-appearance` icon rules are deleted. | Spec §4 "consume the shipped stylesheet chain unchanged"; kit `globals.css:1-6` |
| S5 | **One appearance manager: next-themes `ThemeProvider`** in an app-owned client module `web/components/layout/theme-provider.tsx`: `attribute="class"`, `defaultTheme="system"`, `enableSystem`, `storageKey="appearance-mode"` (every stored `light|dark|system` preference survives), `enableColorScheme={false}` (theme.css sets `color-scheme`), no `value` map, no `nonce` (no CSP exists; the prop is noted for G012). Rendered as the first child of `<body>` in `app/layout.tsx`; the `appearance-init` script and `lib/appearance.ts`'s script export are deleted. `<html lang="en" suppressHydrationWarning>` stays (next-themes writes `class` before hydration). The control stays a native `<select aria-label="Appearance">` driven by `useTheme()`. Required by kit `sonner.tsx:3,8`; a second manager would make toasts follow the OS instead of the user's choice. | Spec §4 "one authoritative appearance state (including next-themes where required by the source)" |
| S6 | **D-A1 pre-hydration System symbol: option (a), settled by Codex (r1)** — with framework JS held, `<html>` carries only the resolved class, so the closed control shows Sun/Moon for a stored System until hydration, then Monitor. S-FIRSTPAINT-DELAYED keeps every palette assertion (first-frame ground/action/H1 ink equal the stored preference's palette) and retargets only the "symbol while held" line to the resolved symbol; a second inline marker script is not added. Codex: acceptable with all palette/first-paint guarantees retained; no second theme script. | Spec §4: correct first paint = palette; "do not retain legacy selectors merely to satisfy obsolete tests"; Codex review "Decisions" D-A1/S6 |
| S7 | **Tenant theme host at the document root**: a server component `web/components/branding/tenant-style.tsx` renders `<style data-tenant-branding>` in `<head>` from the re-checked `Brand.tokens` (`readTokens` guarantees `^#[0-9a-f]{6}$`; the emitter re-validates and emits only the names of the fan-out table §2.5), as `html:root:not(.dark){…}` (Light) and `html.dark:root{…}` (Dark) — specificity (0,2,1) beats the kit's `:root`/`.dark` (0,1,0) and the app corrections regardless of stylesheet order. Server-rendered → first HTML byte correct in both appearances; every body portal (dialog, sheet, popover, select, tooltip, menus, toasts) inherits from first insertion; no post-render repair. Tenant switching stays per request (a different host = a different `<style>`); reset = empty element = kit + app defaults. `TokenScope`, `PortalContainerProvider`, `usePortalContainer`, `tokenOverrideStyle`, `--light-*/--dark-*` and the scope-container adapter are deleted. | Spec §4 portals paragraph; ADR §3 |
| S8 | **Isolated previews** = an app-owned server component `web/components/branding/brand-preview.tsx`: `<div data-brand-preview class="light|dark" style="--primary:…;--vx-blue:…">` with one appearance per surface (a System-following preview carries no class). Rule: **previews render no portaled component** (E1 option a; any `*Content` would land in `document.body` with the live tenant's values). Production has no live preview today (the branding form's swatches stay); the fixture page is the proof surface. | Spec §4 "independent branding previews must remain isolated"; kit `theme.css:3,51`, `vestive.css:6` match any element |
| S9 | **Saved-branding mapping = the fan-out table §2.5** (D-M1 yes: `primary` also drives `--vx-blue`/`--sidebar-primary`; D-M2 yes: `accent` → `--vx-gold` only, never the kit's neutral `--accent`; `muted` → `--muted` + `--accent` + `--sidebar-accent`; `popover(-foreground)` → also `--card(-foreground)`). **No persisted-contract change**: columns, 15 names, hex grammar, JSON shapes, audit payloads, error codes and `"<pair>:<appearance>"` fields are unchanged; rows are interpreted at render. | Spec §4; `core/internal/branding/branding.go:42-45`; `lib/branding.ts:18-34` |
| S10 | **App-owned accessibility variable corrections** (documented in WEB-SHELL.md; exact values measured in P0): Light `--vx-blue` dark enough that white keeps ≥4.5:1 at rest **and** on the kit's `hover:bg-primary/80` composite over `background` and `popover` (arithmetic candidate `#004fb4`: 7.6 rest, 4.9 hover; `#0062d6` is 5.65 rest but ≈3.97 on hover; kit `#0075ff` is 4.21 at rest); Dark `--vx-blue` `#0062d6`; Light `--muted-foreground` `#6f615b` (kit `#7c6d67` is 4.40 on the muted band); Light `--destructive` `#d2000a` (kit `#e7000b` is 4.01 as text on its 10 % tint; `#d2000a` ≈4.7). Dark border/input translucency: `--input` takes the D-A2 values (S22) `#908279` / `#7d726d`; `--border` stays the kit's unless P0 measures a failing boundary. `CtaButton` (`bg-blue text-white`) and `text-blue` follow `--vx-blue`, so the single value serves both; with S24 the CTA renders only under the platform palette, so its fixed-white label is measured against these app-owned `--vx-blue` values at rest, ×1.10 (hover) and ×0.90 (active): Light `#004fb4` ≈ 7.6 / 6.6 / 9.0, Dark `#0062d6` ≈ 5.65 / 4.84 / 6.9 (computed; measured in P0.2 and pinned in `TestDefaultsPass`). | Spec §4 / ADR §4 "app-owned supported theme variables can address contrast without editing source" |
| S11 | **Core alignment, bounded ("Core palette delta r3", §2.6; r1)**: `core/internal/branding/colors.go` platform map = the app's effective rendered defaults after cutover; existing IDs T1–T10/N2–N4 keep their meaning class and are re-modelled to what the kit paints for the variants the app **actually uses** (`default` ×16 sites, `outline` ×23, `secondary` ×2, `ghost` ×1, `destructive` ×1; `link` unused — not tested): roles `ink/text-2/text-3` replaced by `foreground` and the `foreground` 80 % composite; **T8** = `primary-foreground` on `primary` at rest and on `primary` at 80 % over `background` and over `popover` (kit `hover:bg-primary/80`, the only hover model of the used Button); **T9** = `secondary-foreground` on `secondary` and on `secondary` mixed 5 % `foreground`; **T10** = `destructive` **as text** on its tint over `background`/`popover` — Light 10 % and 20 %, **Dark 20 % and 30 %** (`button.tsx:13`) — and as text on `background` (errors); `outline`/`ghost` hovers are `foreground` on `muted` (= T3). **New T12** = `foreground` on `input` at 50 % over `background` and over `popover` (the kit field fill of Input/Textarea/NativeSelect) and, Dark only, on `input` at 30 % over `background` (the outline hover ground); no placeholder pair because S22 removes the app's only placeholder. **No CTA pair**: under S24 the fixed-white `CtaButton` never renders under a tenant palette, so its contract is a platform-default assertion in `TestDefaultsPass` (white on platform `primary` ×1 / ×1.10 / ×0.90) plus the browser probe, not a tenant pair that would refuse legitimate light primaries (the seeded Dark `#3fb8a3` would fail at ≈2.44:1 for a button tenants never see). The whole-element brightness factors are withdrawn from T8–T10 (the used kit Button has none). N3 follows S22 (`input` as a solid border ≥3:1 on `background`/`popover`). `PAIR_FIELD` (`web/lib/branding.ts:410`) widens to `^(T1[02]|T[1-9]|N[2-4]):(light|dark)$` (W1) and the branding page `PAIRS` map gains T12 and the rewritten descriptions (W3 consuming W5's contract, §2.6). Reviewed as its own bounded delta (G067 precedent). Stored rows: no migration; re-judged on their next save; the rendered-state consequences and the resolution rule are S23. | Spec §4 "bounded Core branding defaults/validation alignment is in scope"; Codex review item 2 and D-M3 |
| S12 | **Consumers**: kit `Button` (variants default/outline/secondary/ghost/destructive; `link` unused); app-owned `web/components/controls/button-link.tsx` (`<a>` with kit `buttonVariants()` or, for the one landing CTA of S24, `ctaButtonVariants()`, `data-slot="button"`; links are never buttons); `web/components/controls/pending-submit.tsx` (moved from `account/`, renders kit `Button` only — the r0 `cta` prop is dropped under S24; the sign-in submit is kit `Button` `size="lg"`; `ref` passes as a prop on React 19); `web/components/controls/field-classes.ts` exporting the D-A2 class strings of S22 applied as `className` at every kit `Input`/`Textarea`/`NativeSelect`/`RadioGroupItem` site; `web/components/controls/radio-choice.tsx` composing kit `RadioGroupItem` (with the S22 radio class) + `Label` (layout only); consumers use kit `RadioGroup name defaultValue required` (the documented form API: one hidden input per group); kit `NativeSelect` (+`className="w-full"`) at the 6 native sites — never the compound `Select`; kit `Input`/`Textarea` incl. the branding form's hand-copied colour/file inputs; kit `Dialog` (body portal, `sr-only` Close keeps the name); kit `Table` parts with an app-owned `web/components/controls/scroll-region.tsx` (`role="region"` `aria-label` `tabIndex=0` wrapper) at the 4 list sites; `@/components/vestive/{typography,layout,content}` for Heading/Text/Kicker/TextLink, Container/Section (`surface`→`muted`, `raised`→`card`), BulletList. `Text variant="index"` (`muted-foreground/80` ≈3.8:1) is not used; the 4 sites use `meta`. The 11 old modules (at `web/components/legacy-ui/*` after S0.0, S21), `components/account/{pending-submit,radio-option}.tsx`, `components/layout/token-scope.tsx`, `lib/{appearance,tokens}.ts`'s old exports are deleted in the same commits that migrate their last consumer. | Spec §2–§3 |
| S13 | **D3 header**: the app keeps its own header composition inside `app-shell.tsx` on kit `Container`/`Button`/typography (skip link, brand mark, Primary nav with `aria-label`, appearance control, help dialog, account/indicator; not sticky). Kit `SiteHeader` requires a logo URL, defaults to a brand wordmark and marketing links, renders `<a href="#">` and is sticky (S-TAB unobscured focus) — it stays compiled, unused, as the Spec allows. `web/components/ui/site-header.tsx` is deleted. | Spec §2 "application-specific composition outside the copied tree, with honest names"; §3 no marketing defaults |
| S14 | **D4 RuledGrid**: consumers pass `style={{ gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 16rem), 1fr))", minWidth: 0 }}` (props, not source) so the grid reflows at 320 px; S-RESPONSIVE decides. Fallback if any site still scrolls: plain Tailwind grid utilities at that site. | Spec §4 reflow; `layout.tsx:41` spreads `style` last |
| S15 | **E3 CtaButton focus**: used only for the landing CTA link (S24), with `className="focus-visible:outline-offset-[-3px] focus-visible:outline-white"` (an inset 2 px outline inside the clip, white on the app-owned blue fill ≥6.6:1 at every brightness state; tailwind-merge resolves the offset/colour conflicts; `outline-white` is the kit's own fixed label colour, exempt from the palette-utility ban at exactly this site, §3.3). Measured by `expectFocusRing` in P0.2 on the landing fixture; fallback = kit `Button` for that action too. | ADOPTION.md:115-117; className is a prop |
| S16 | **E2 scrim**: the kit's theme-fixed `bg-black/30` + blur is accepted; the `overlay` token is retired from `lib/tokens.ts` (never persisted). **D1 destructive**: solved by S10's value (no composition). **D-RAD**: kit `RadioGroup` is the compound bridge the Spec names; its selection change requires JS (the hidden input still submits `defaultValue` without JS; server validation unchanged) — recorded as a known behavioural delta in the Result, not a stop. | Spec §3 "correctly bridge the compound control's documented form API" |
| S17 | **Literal guard**: B-WEB-DESIGN's identity/marketing scan exempts exactly the 69 fidelity-verified `web/src` paths (by manifest, not glob) and, in app-owned files, only import lines matching `^import .* from "@/components/vestive(/[a-z-]+)?";?$`; every other literal stays governed (runtime copy is still checked by S-IDENTITY/E-BRD). New controls A20 (a byte-altered kit file at its real path → `UI_FIDELITY`, not exempt) and A21 (a non-import brand literal in an app file → `DESIGN_LITERAL`). | Spec §3 "narrowly scoped literal-guard allowance; application-owned files and runtime identity remain governed" |
| S18 | **ESLint**: one override object in `web/eslint.config.mjs` for `files: ["src/**"]` turning off exactly the rule IDs the proof reports (expected: `@next/next/no-img-element`; possibly `react-hooks/set-state-in-effect` for `use-mobile.ts:14`, `react-hooks/static-components` for `calendar.tsx`); **no `ignores`** (the tree stays linted and typechecked); the two exact G067 skill-ignore lines stay (A14–A17 unchanged). | Spec §2 "no disabled type checking"; `package.json:9` `--max-warnings=0` |
| S19 | **`document.cookie` in kit `sidebar.tsx:86`**: B-WEB-BOUNDARY keeps the rule and allows exactly that file, paired with a new assertion that no app-owned file imports `@/components/ui/sidebar` (the cookie code is unreachable; it stores a sidebar UI preference, never a credential). | B-WEB-BOUNDARY `run.sh:1002` |
| S20 | Focused validation only (§6); no `ci-controls`, no `root` (M1 holds the default ports), no full local suite; post-merge affected coverage on the exact merged SHA (now including one `off` run, S25); one MR. **Check timing (r1):** every commit integrated on `g069` passes the coordinator's stack-free whole-app `tsc --noEmit` + `eslint --max-warnings=0 .` (the transitional `legacy-ui` state of S21 included); the legacy-absence rules (B-WEB-COMPONENTS `UI_LEGACY`/fidelity, B-WEB-TOKENS, B-WEB-DESIGN S17, B-NAMING, B-WEB-BOUNDARY) are coded by W5 in phase A but **first required in phase D**, after L1–L3 are integrated and `web/components/legacy-ui` is deleted; hosted fast (lint-web included) runs on the MR head only. The whole-app typecheck first becomes required at P0.1 step 3 (on the transitional state) and stays required at every integration. | CLAUDE.md policy; Spec §6; Codex review item 1 |
| S21 | **Transitional legacy staging (S0.0, W0, before the alias flip; r1)**: the 11 old `web/components/ui/*.tsx` modules are `git mv`-ed to `web/components/legacy-ui/*.tsx` and the old `web/lib/utils.ts` (`export { cn } from "cn"`) is deleted; one mechanical rewrite changes the 54 importers' specifiers `@/components/ui/<m>` → `@/components/legacy-ui/<m>` (53 app files + `tests/fixtures/token-scopes/page.tsx`) and the 14 app importers of `@/lib/utils` keep their specifier (the kit's `cn(...inputs)` has the same call shape; verified by the same `tsc`). Nothing else changes in those files (no API migration, no shadow implementation). Then S1 flips the alias and S0.1 copies the kit: the kit resolves under `web/src`, the app keeps compiling against its unchanged legacy modules, and `next build` works — which is what makes P0.1 step 3 and the P0.2 production build executable. Each consumer lane migrates its `legacy-ui` imports to the kit and deletes a legacy module with its last consumer; `web/components/legacy-ui` must be empty/absent at phase D (`UI_LEGACY`, S20). Transitional files carry no new literals (B-WEB-DESIGN governs them exactly as it governs them today). | Codex review item 1 ("move the minimum prerequisite mechanical consumer migration before the gate"); chosen over a kit-only gate because P0.2 needs a production build of the whole app |
| S22 | **D-A2 settled (Codex: option bc, source bytes unchanged)**: app-owned `--input` `#908279` (Light) / `#7d726d` (Dark) **plus consumer `className` props, not a CSS rule**: `FIELD_CLASSES = "border-input dark:aria-invalid:border-destructive"` on every kit `Input`/`Textarea`/`NativeSelect` site and `RADIO_CLASSES = "border-input"` on `RadioGroupItem` (both exported from `web/components/controls/field-classes.ts`, W0). Precedence is by construction, not by cascade order: tailwind-merge drops the kit's `border-transparent` for `border-input` (same utility group, later wins) and the kit's `dark:aria-invalid:border-destructive/50` for the solid twin; the kit's `focus-visible:border-ring`, `aria-invalid:border-destructive`, `focus-visible:ring-ring/30`, `aria-invalid:ring-3`, `disabled:opacity-50` are untouched variant utilities and keep their own precedence (variant utilities are emitted after plain ones in Tailwind v4). Resulting states (computed, Light on white): rest boundary `#908279` 3.71:1 (N3), radio rest (`bg-input/90` + solid border) ≥3.16:1 on `background` and ≥3.30:1 on `muted`, invalid `#d2000a` solid, focus = kit ring composite; `outline` buttons (`border-border`) are unaffected. The CSS-rule alternative is recorded as the fallback only: an unlayered `[data-slot=input]{border-color:…}` would beat every layered utility including focus/invalid (`vestive.css` puts the kit in Tailwind's `@layer base/components`, utilities in `@layer utilities`), so if it were ever needed it must live in `@layer utilities`, precede nothing the kit emits, and prove precedence in P0.2 — the className route needs no such proof. Measured in P0.2 (field rest/focus/invalid/disabled, radio rest/checked/disabled, textarea, native select, in both appearances and under `fixtureBranded`/the retyped grey-primary). No accessibility exception, no broad restyling, no owner gate. | Codex review D-A2; Spec §4 existing guarantee |
| S23 | **D-M3 settled (Codex: accept in principle, with a rule)**: no storage migration; columns, 15 names, hex grammar, API, audit payloads and error codes unchanged; Core re-judges a row only on its next save. **Documented rendering consequences:** saved `input` (formerly only the 3:1 boundary) now also paints the 50 % field fill and the 90 % radio fill; saved `destructive` (formerly a solid fill with a `background` label) now paints as text on its 10/20 % (Light) or 20/30 % (Dark) tint and as error text; saved `primary` now paints the `hover:bg-primary/80` composite instead of a brightness state; a row accepted under the G067 pairs may receive `contrast_failed` with the same pair ID (or T12) on its next save. **Rule (stop-style, §1.4):** every existing accepted saved-branding fixture is part of the P0.3 compatibility proof — the seeded custom-host row (`infra/identity/mcp-seed.sql:109-110`), `fixtureBranded`, `fixtureGreyPrimary`, `fixtureDarkSparseDestructive` (`fixtures.ts:91-166`), the accepted/refused U-BRD-4 cases (`branding_test.go:97-148`), BRD-1's `#1d4ed8` row and BRD-4's refusals (`branding_integration_test.go:79-213`), and E-BRD-1/E-BRD-5 (`identity.spec.ts:1286,1354`); unchanged hex is never taken as unchanged rendering — each is re-run through the r3 pairs **and** measured on the rendered states. Any fixture, seed or stored row that is incompatible (refused by r3) or inaccessible (a used rendered state < 4.5:1 text / < 3:1 non-text) gets a concrete bounded resolution proposal in the P0 checkpoint before the Core delta or the affected lane lands; nothing is dismissed. Pre-identified by computation (§0 item 15): the seed/`fixtureBranded` Light hover ≈4.27:1 and `fixtureGreyPrimary` hover ≈3.60:1 — proposed resolution: retype those **test** inputs to r3-valid edges (Light `primary` ≈ `#0a5a4d` for the seed/branded fixture, rest ≈8.1 / hover ≈5.1; grey edge ≈ `#525252`, rest ≈7.7 / hover ≈4.64; final values from P0 measurement), keeping E-BRD-5's refusal (`light.primary-foreground #0f6b5c` still fails T8 at rest) and the comments that explain each edge; no persisted data changes. Stored rows outside the repository's seeds are not known to exist at this pre-release stage; if the owner has any, the same rule applies (proposal before landing), see §1.3. | Codex review D-M3 |
| S24 | **CtaButton placement (r1)**: the kit `CtaButton`/`ctaButtonVariants` (fixed `text-white` on `bg-blue`, `hover:brightness-110 active:brightness-90`) renders only where the **platform palette** is guaranteed: the root-host landing CTA (`web/components/signup/landing.tsx:68` via `ButtonLink`). The sign-in submit (`web/app/sign-in/page.tsx:82`, served on tenant hosts, E-BRD-1) uses kit `Button` `size="lg"` so tenant `primary-foreground` on `primary` applies (T8). Reason: a fixed-white label cannot follow a tenant's `primary-foreground`; a Core pair "white on primary ×1/×1.10/×0.90" would refuse legitimate light primaries (the seeded Dark `#3fb8a3` ≈2.44:1) for a control tenants never see, while omitting it would let the yellow-primary/black-foreground case pass Button checks and ship an unreadable CTA. "Where appropriate" (Spec §3) is read with the kit's own note (`cta-button.tsx:5-6`: the one call to action of a view; everyday actions use `Button`). Executable: B-WEB-COMPONENTS (h) `UI_CTA_TENANT <path>` — `CtaButton|ctaButtonVariants` may be imported only by `web/components/controls/button-link.tsx` and used (`variant="cta"`) only in `web/components/signup/landing.tsx`; `TestDefaultsPass` pins white on platform `primary` at ×1/×1.10/×0.90 ≥4.5 in both appearances; the browser CTA probe (§6.2) measures the painted rest/hover/active label on the landing. | Codex review item 2; Spec §3 "use kit CtaButton where appropriate" |
| S25 | **`off` disposition (r1)**: `off` = `variants_off` (B-TOGGLE-OFF, B-TOGGLE-OFF-ADDITIONS, B-OBS-OFF, B-BROWSER-OFF): an archetype render with `intelligence_enabled=false` plus the `APPLICABLE` repository paths, migrations 1–21, the GETTING-STARTED blocks, the observability chain and **the full `BROWSER_MATRIX_FULL` browser suite on that copy** (`run.sh:5668-5788`). G069 touches it in two ways: (a) `APPLICABLE` must gain `web/src` and `web/tsconfig.json` (§3.4, W5) or the copy cannot resolve the kit; (b) the copied web bytes change. There is no narrower selector inside `off` (one function, no `--grep`/project knob), and no attributable current evidence: the last executed `off` PASS is G064's corrected source (`74091550`, 2026-10-23), whose web tree predates G067/G068, and the RC15 candidate's hosted qualification is recorded as pending. Rationale: B-BROWSER-OFF's UI assertions are the same specs rows 3 runs privately on the same candidate bytes; what only `off` proves is (a) — the disabled composition still builds and serves the kit. Selection: **no pre-merge `off` run; one `off` run post-merge on the frozen merged SHA (row 7)** under the shared-port harness lock, ports `PORTS_OFF` free of M1; its browser failures, if any, are attributed to the composition (the same specs passed on row 3) and repaired by a normal source MR. Not a whole-campaign default: `ops`, `rename`, `migrations`, `root`, `ci-controls` stay unselected (no env/compose/migration/db change; M1 ports). Disabled-**state** UI coverage (the other reading) is already selected by existing selectors: E-ENR-5 (disabled "open" radio option, `enrollment.spec.ts:483,507-519`), E-A11Y's pending `aria-disabled` submit (`identity.spec.ts:901,934-941`), the P0.2 `disabled` radio item and the kit `disabled:opacity-50` samples in S-STATES. | Codex review item 4; CLAUDE.md post-merge affected-coverage policy |

### 1.2 Unproven assumptions (the proof §4 settles each)

| # | Assumption | Proof / fallback |
|---|---|---|
| A1 | All 69 TS/TSX files typecheck under the app's `strict`/`isolatedModules`/`bundler` options with the pinned packages (UMD-global `React.*` type references in 9 files, Base UI member APIs in combobox/menubar/drawer/navigation-menu, `react-day-picker 9.14.0` types, `react-resizable-panels` v4 names, `recharts 3.10.1` type exports). | P0.1 `tsc --noEmit` on `src/**` only. A failure is an upstream incompatibility → §1.4 stop for that file class (compile blocker), reported with the exact error. |
| A2 | The actual npm resolution equals S2/S3 (Radix only under cmdk; `reselect` 5.2.0 nested beside 5.3.0; zod ≥ 4.6.6) and B-DEPSCAN reports no Critical/High/Unclassified finding in the new closure (shadcn's CLI closure included). | P0.1 lockfile inventory + `npm ls --all`; row 1 B-DEPSCAN. A blocking finding → §1.4 dependency/security stop (per-finding disposition under SECURITY §4.3, never a blanket waiver). |
| A3 | Multi-candidate tsconfig `paths` are honoured by TypeScript, `next build` (Turbopack), the ESLint TS resolver and Playwright's loader. | P0.1 `tsc` + `eslint`; P0.2 production build in the private browser image. Fallback: `web/src` as the only `@/*` root is impossible without moving app code → report. |
| A4 | next-themes 0.4.6 renders a classic inline script at the provider position (inside `<body>`, before children), sets the html class from `localStorage["appearance-mode"]`/`matchMedia` before first paint, and `useTheme` reads the provider. | P0.2 S-FIRSTPAINT/S-FIRSTPAINT-DELAYED/S-HYDRATION-DEV/S-HEAD-ORDER (retargeted). |
| A5 | React 19 renders a plain `<style>` child of `<head>` in place with no hydration mismatch under `next dev`. | P0.2 S-HYDRATION-DEV, S-HEAD-ORDER. Fallback: render the element at the top of `<body>` (same cascade result). |
| A6 | Whole-state focus (1 px `border-ring` + 3 px `ring-ring/30`) measures ≥3:1 against the kit field fill and the page in both appearances and under `fixtureBranded`/the retyped grey-primary; the S22 field and radio states (rest boundary, invalid, disabled, checked) and the composites of S10/S11 (T8 80 % hover, T10 tints, T12 fills) measure as computed. | P0.2 `expectFocusRing` composite rewrite, S-STATES-style samples, the S22 field/radio samples. A failure is corrected by an app-owned `--ring`/`--input` value or a `className` (allowed) before any consumer lane proceeds; a state no app-owned value or prop can fix is a §1.4 stop. |
| A11 | The S23 computations hold on the rendered page: the seed/`fixtureBranded` Light hover and `fixtureGreyPrimary` hover are below 4.5:1 under the kit's `bg-primary/80`, and no other listed fixture fails an r3 pair or a measured used state. | P0.3 (the `go test` pair run + the fixture-page samples). Either outcome produces the S23 proposal/record in the P0 checkpoint. |
| A12 | The S21 mechanical rewrite (54 importers, 14 `@/lib/utils` importers, the token-scopes fixture page) leaves the app typechecking and building with the alias flipped and the kit present, with no kit file excluded and no old module shadowed. | P0.1 step 3 (`tsc`, `eslint`, then the P0.2 `next build`). A residual collision is fixed by renaming the legacy module's path, never by an exclusion. |
| A7 | Base UI `RadioGroup`: `name`/`defaultValue`/`required`/per-item `disabled` reach the submitted `FormData`, arrow keys move selection, `getByRole("radio").check()` works. | P0.2 proof form; the identity-stack `.check()` sites listed in lane 3 §2.3. Failure → D-RAD escalation (app-owned native radio composition with kit `Label`/`Field`). |
| A8 | `ref` passes through kit `CtaButton`'s props spread to the `<button>` (React 19 ref-as-prop), so `PendingSubmit` keeps its submit interception. | P0.2 proof form. |
| A9 | The rendered kit `Heading` (Oswald uppercase, `clamp(2.6rem,5.6vw,5rem)` H1) keeps S-LONG-IDENTITY/S-TEXT-SPACING reflow at 320 px. | P0.2 / row 3. Fallback: `className` size overrides at the shell H1 site (prop). |
| A10 | `shadcn@4.22.1/dist/tailwind.css` still equals the vendored block (16,340 B, sha256 `4c371f7a…`), so the registry layer is visually unchanged. | P0.1 evidence line. |

### 1.3 Decisions that need a bounded owner/Codex choice (r1)

**None open.** r0's D-A2 and D-M3 were decided by Codex and are settled as S22 and S23; D-A1 is settled as S6. The r1
corrections create no new owner decision; one **conditional** item exists, which arises only from the P0.3 measurement
and is disposed inside the existing checkpoint:

| ID | Condition | Smallest options (recommendation first) |
|---|---|---|
| D-M3a | P0.3 confirms (or finds) an accepted saved-branding fixture, seed or stored row that r3 refuses or that paints a used state below threshold (§0 item 15 pre-identifies the seed/`fixtureBranded` Light hover and `fixtureGreyPrimary`). | (i) retype the affected **test** inputs to r3-valid edges with the recorded reason (S23 values; no persisted data) — recommended and expected to cover every known case; (ii) if the owner reports stored rows in a live environment, a bounded follow-up item "stored-row re-validation notice on the branding page" (web-only, reads the existing `contrast_failed` contract on next save; no Core API change) — proposed, not executed in G069; (iii) a reviewed data migration — not recommended (no correctness need). Codex disposes (i) at the P0 checkpoint under the existing review; only (ii)/(iii) would reach the owner. |

### 1.4 Stop conditions (Goal "Stop / Escalation"; stop the affected lane, report the exact conflict and smallest options)

- An upstream file does not compile/typecheck/lint under any app-side configuration (A1/A3/A4): no kit edit, no
  `ts-ignore`, no exclusion; report file, error and the smallest options (pin change, configuration, Spec amendment).
- A blocking B-DEPSCAN finding in the new closure without a fixed release (A2).
- An accessibility requirement that only a kit edit could satisfy after D-A2 (e.g., focus composite failing with every
  app-owned `--ring`): report the exact measured case.
- A persisted-branding semantic that cannot be preserved by the fan-out table (none identified; S23 covers the two
  reinterpretations).
- **S23 rule (r1):** an accepted saved-branding fixture, seed or stored row that the r3 pairs refuse or that paints a
  used rendered state below threshold stops the Core delta / affected lane until a concrete bounded resolution
  proposal (§1.3 D-M3a options) is recorded in the P0 checkpoint and disposed; it is never silently dismissed and no
  unreadable rendering or data loss is approved by default.
- Resource collision with the owner's preview or an unowned Docker object; a third concurrent run; a second builder.
- Scope creep: any consumer needing new behaviour (copy, routes, form semantics, authority) is reported, not done.

## 2. Approach

### 2.1 Source copy (S0.1, owner W0)

`git archive 29dded76 .agents/skills/vestive-brand/kit/src | tar -x` into `web/src` (strip the prefix), 69 files, Git
mode 100644. Verification at copy time and in every row 1: per-file `sha256sum -c` against the 69 `kit/src/` rows of
`LOCAL.sha256` with paths rewritten to `web/src/`, plus the aggregate digest `597f2aac…` recomputed over `web/src`
with the Spec algorithm; `find web/src -type f | wc -l` = 69; the upstream tree unchanged (`git diff 29dded76 --
.agents/skills/vestive-brand/kit` empty). The kit's `components/vestive/index.ts` re-exports stay; app consumers import
the concrete modules.

### 2.2 Aliases and configuration (W0)

`web/tsconfig.json` `paths` (S1); `web/components.json` unchanged (its aliases `@/components|@/components/ui|@/lib/utils|@/lib|@/hooks` already match the kit; `tailwind.css: app/globals.css` stays true);
`web/eslint.config.mjs` (S18); `web/next.config.ts`, `postcss.config.mjs`, `next-env.d.ts` unchanged.

### 2.3 Dependencies (S0.2, owner W0; the resolution container is run by the coordinator §4)

Pins per S2; `web/package.json` + `web/package-lock.json` (both registered customizations; reasons updated);
`web/THIRD_PARTY_NOTICES.md`: the shadcn/ui section rewritten for the verbatim 54-file registry output and the
owner-authored `components/vestive`, the Base UI table extended with the full new closure and licences from the
lockfile (`recharts` pulls ISC/BSD d3 packages through `victory-vendor`). Dependency rules in `infra/validation/lib.sh`
(§3.3) are W5's.

### 2.4 App stylesheet and theme host (S0.3, owner W1)

`web/app/globals.css` per S4. Cascade: kit `theme.css` `:root`/`.dark` (0,1,0) → app corrections `:root`/`.dark`
later in the same compiled sheet (same specificity, later wins) → tenant `<style>` `html:root:not(.dark)` /
`html.dark:root` (0,2,1) → preview elements' inline `style` (highest). The `@theme inline` map resolves `--color-*`
at use time, so every level re-resolves utilities. **Layers (r1, S22):** the kit's `vestive.css` declares
`@layer base`/`@layer components`; Tailwind v4 emits utilities in `@layer utilities`; an unlayered app rule would
outrank all of them, so the app stylesheet adds **no element/attribute rules** that touch kit-utility properties — the
only app-owned stylesheet content is variables (`:root`/`.dark`), `@font-face`, the blueprint backdrop and the
reduced-motion rule, exactly as today. Field boundaries are `className` props (S22). `web/app/layout.tsx`: `<html lang="en" suppressHydrationWarning>`,
`<head>` holds `<TenantStyle brand={…} />` (server component, reads the same `tenantRequest()` facts the frame uses;
loopback/unknown host → empty), `<body>` opens with `<ThemeProvider>` (S5). `lib/appearance.ts` keeps only the mode
constants (`APPEARANCE_MODES`, `APPEARANCE_STORAGE_KEY`, `DEFAULT_APPEARANCE`, `parseAppearanceMode`) consumed by the
control and the specs; no storage access anywhere in app code. `components/layout/appearance-control.tsx`: `useTheme`;
the symbol follows `theme` after hydration (S6). `components/branding/tenant-frame.tsx` keeps `shellBrand`;
`AppShell` loses `TokenScope` and `brand.tokens` (the style element carries them).

### 2.5 Saved-branding fan-out (the explicit mapping; emitter `web/lib/tenant-css.ts`, unit-tested)

| Saved token (`branding.go` `Tokens`) | Kit variables set (Light block / Dark block from `light`/`dark`) | Note |
|---|---|---|
| `background` | `--background`, `--sidebar` | lossless |
| `foreground` | `--foreground`, `--accent-foreground`, `--sidebar-foreground` | kit text on the neutral accent fill; T3 guarantees foreground on muted (= accent fill) |
| `primary`, `primary-foreground` | `--primary`, `--vx-blue`, `--sidebar-primary` / `--primary-foreground`, `--sidebar-primary-foreground` | D-M1: CtaButton/`text-blue` follow the tenant |
| `accent` | `--vx-gold` | emphasis ink (Kicker, TextLink, `::selection`, default `--ring`); **never** the kit `--accent` |
| `muted`, `muted-foreground` | `--muted`, `--accent`, `--sidebar-accent` / `--muted-foreground` | kit neutral fill = saved `muted` (G067 translation reversed) |
| `secondary`, `secondary-foreground` | `--secondary` / `--secondary-foreground` | lossless |
| `popover`, `popover-foreground` | `--popover`, `--card` / `--popover-foreground`, `--card-foreground` | kit card = popover |
| `border` | `--border`, `--sidebar-border` | an opaque tenant hairline replaces the kit's translucent Dark default |
| `input` | `--input` | S22/S23: boundary (solid, via `className`) **and** the 50 %/90 % fills; T12/N3 |
| `destructive` | `--destructive` | S23: text on its tint, error text; T10 |
| `ring` | `--ring`, `--sidebar-ring` | lossless |
| not saved: `overlay`, `--radius`, `--angle`, `--chart-*`, `--vx-{orange,green,lime,cyan,umber,star}`, `--vx-ribbon` | kit defaults, documented as non-overridable | `overlay` token retired |
| `display_label`, logo | unchanged (`BrandMark`) | — |

### 2.6 Core palette delta r3 (r1; Core owner W5; page consumers W3/W1; reviewed as its own delta)

**W5 (Core semantics and tests):** `core/internal/branding/colors.go`: platform map updated to the app's effective
rendered defaults (values from P0; unchanged names); pairs per S11 with a `blend(fg, bg, alpha)` helper replacing
`scale()`/`states` — T8 (rest; 80 % over `background`; 80 % over `popover`), T9 (rest; 5 % `foreground` mix), T10
(text on the Light 10/20 % and Dark 20/30 % tints over `background`/`popover`; text on `background`), **T12 new**
(`foreground` on `input` 50 % over `background`/`popover`; Dark `input` 30 % over `background`), N3 (`input` solid on
`background`/`popover` ≥3), the rest unchanged; roles `ink/text-2/text-3` removed from the map and `Effective` (the
`foreground` override no longer fans out to roles); `scale()` is kept **only** for the platform-default CTA assertion.
`branding_test.go`: U-BRD-3 `TestDefaultsPass` re-pinned (T8 80 % hover, T10 tint, N3, T12 ratios; **plus** white on
platform `primary` at ×1 / ×1.10 / ×0.90 ≥4.5 in both appearances = the S24 CTA contract); U-BRD-4 edges retyped to the
alpha model (the "primary hover loses the label" edge and the accepted grey primary move to r3-valid values, S23);
a new T12 refusal edge (mid-grey `foreground` with mid-grey `input`). The **contract W5 hands over** (recorded in the
Result's Core delta section): the pair IDs with one plain-language description each, evaluated in ID order so the
first failing pair is reported as today.

**W3 (page consuming the contract):** `web/app/organization/branding/page.tsx:49-63` `PAIRS` rewritten from W5's
descriptions (same IDs T1–T10/N2–N4 kept, T12 added) and `branding-form.tsx:70`'s placeholder replaced by a visible
description (S22). **W1:** `web/lib/branding.ts:410` `PAIR_FIELD` widened for T12; `tests/unit/branding-client.test.mjs`
"management answers" extended with one `T12:light` round-trip. Dependency recorded in §5.2: W5's contract text precedes
W3/W1's edits; the Core values follow P0 measurements. E-BRD-5 (`T8:light`) keeps working. No SQL, API, audit or
migration change. Outcomes: `unit-core` (U-BRD-*), private `TestBranding` 23/23 + `TestBrandingCatalog` 3/3 (§6.1 row
5), `unit-web` (the PAIR_FIELD round-trip), E-BRD-5 in row 3i.

### 2.7 Consumer migration map (from lane 3 §1.2, re-verified counts)

**S0.0 first (S21):** the 54 importers of `@/components/ui/*` (typography 49, layout 29, button 16, input 7, select 5,
table 4, bullet-list 3, textarea 2, portal-container 2, dialog 2, site-header 1; 53 app files + the token-scopes fixture
page) are rewritten to `@/components/legacy-ui/*` and the 11 modules moved there; the lanes then migrate from
`legacy-ui` to the kit. `RadioOption` 12 uses in 4 fieldsets; `variant="cta"` at `app/sign-in/page.tsx:82` (→ kit
`Button` `size="lg"`, S24) and `components/signup/landing.tsx:68` (→ `ButtonLink` with `ctaButtonVariants()`, the only
CTA); 52 legacy class uses (`text-text-2` → `text-foreground/80`,
`font-display` → `font-heading`, `px-gutter` → `px-[var(--vx-gutter)]`, `text-accent`/`border-accent`/`hover:text-accent`
→ `gold`, `tone="surface"` ×9 → `muted`, `variant="index"` ×4 → `meta`); `RuledGrid` 8 sites (S14); `Table` 4 sites (S12);
the branding form's hand-copied `FIELD/INPUT/COLOUR_INPUT` and the raw `<input type="file" id="logo-file">` → kit `Input`.
Every mutation stays a native `<form method="post">` with the CSRF hidden input; routes, handlers, copy and authority are
untouched. `tests/fixtures/token-scopes/{page,fixtures}.tsx` are re-modelled (§6.2) and `tests/fixtures/account-mcp/page.tsx`
is unchanged.

## 3. Inventory of contracts, bookkeeping and the minimum mapping change for each

### 3.1 Dependency rules — `infra/validation/lib.sh:1714-1760`, `run.sh:877-956` (W5)

- `DEPS_ADDITIONS`: − `dependencies|cn|0.4.0`; + the 10 runtime pins and `devDependencies|shadcn|4.22.1` (S2).
- `DEPS_REMOVED_PREFIXES=("@radix-ui/")` → removed; new `DEPS_RADIX_VIA_CMDK=(<exact name@version list from P0>)` with two
  rules: any `package.json` key under `@radix-ui/` → `DEPS_RADIX_DIRECT`; any lockfile key `node_modules/@radix-ui/<x>`
  (any depth) not in the list → `DEPS_RADIX_UNEXPECTED <x>`; `cmdk` must be a direct dependency while the list is non-empty.
- `DEPS_FORBIDDEN=(radix-ui)` (shadcn moves to the additions as dev-only; a runtime `shadcn` is `DEPS_ADDITION_WRONG`).
- Controls: C10 → `shadcn` recategorized to `dependencies` (`DEPS_ADDITION_WRONG shadcn`); C13 → direct
  `@radix-ui/react-dialog` (`DEPS_EXTRA_DIRECT` + `DEPS_RADIX_DIRECT`); C14 → nested unlisted `@radix-ui/react-popover`
  (`DEPS_RADIX_UNEXPECTED`); C19 → direct exact `date-fns` (`DEPS_EXTRA_DIRECT date-fns`: transitive-only packages never
  become direct); C1–C9, C11, C12, C15–C18 unchanged; B-DEPS PASS text updated.

### 3.2 Component/fidelity rules — `run.sh:1131-1180`, `lib.sh:1813-1821` (W5)

`web_components_check` becomes: (a) **fidelity**: `web/src` holds exactly the 69 manifest paths, each sha256 equal to
its `LOCAL.sha256` row, aggregate digest equal `597f2aac…` (`UI_FIDELITY <path>`, `UI_SET`); no `web/components/ui`
directory; (b) **no competing implementation**: `usePortalContainer|from "cn"|data-appearance|data-token-scope|--light-|--dark-|appearanceInitScript`
absent from `web/{app,components,lib,tests}` (`UI_LEGACY <path>`); (c) `UI_IMPORT_RADIX` kept over every web source
(tests included); (d) `@base-ui/` imports only inside `web/src` (`UI_IMPORT_BASEUI_APP <path>` elsewhere); (e) app
`data-slot` literals ⊆ the kit's set ∪ {`button`} (ButtonLink); (f) `components.json` expectation unchanged; (g) X9/X7
(`UI_FOCUS`, `UI_ELEMENT_CLIP`) dropped; **(h, r1)** `UI_CTA_TENANT <path>`: `CtaButton|ctaButtonVariants` imported only
by `web/components/controls/button-link.tsx` and `variant="cta"` used only in `web/components/signup/landing.tsx`
(S24); **(i, r1)** `UI_PLACEHOLDER <path>`: no `placeholder=` attribute in app-owned `web/{app,components}` (S22: no
placeholder text exists, so T12 needs no placeholder pair); (b)'s `UI_LEGACY` also matches `components/legacy-ui`
(S21). **Timing:** (a)–(i) are coded in phase A and first required in phase D (S20). Controls: A22 → a
`variant="cta"` appended to `web/app/sign-in/page.tsx` (`UI_CTA_TENANT`); A23 → a `placeholder="x"` appended to the
same file (`UI_PLACEHOLDER`). Controls: A1 → byte-altered `ribbon.tsx` at `web/src/components/vestive/ribbon.tsx`
(`UI_FIDELITY`); A5 → `@font-face` appended to `web/components/layout/brand-mark.tsx` (`FONT_FACE`); A12b → a Base UI
import appended to the same app file (`UI_IMPORT_BASEUI_APP`); A20/A21 per S17; A2–A4, A6–A12, A13–A19 unchanged.

### 3.3 Token rules — `run.sh:1050-1114` (W5)

Rewritten to the kit contract: palette source = `web/src/styles/theme.css` (bytes proven by (a) above); app entry's
first statement is `@import "../src/globals.css";`; the app-owned correction block declares exactly the documented
variable names per appearance (pinned list from P0, S10/D-A2); colour literals only in that block, in `web/lib/tenant-css.ts`'s
unit fixtures and `tests/fixtures`; palette-utility ban (`bg-black`, `text-white`, …) and the `bg-accent|accent-foreground`
ban over app-owned `app/`, `components/`, `lib/` only (kit files are exempt; one recorded exception: the literal
`focus-visible:outline-white` of S15 in `web/components/signup/landing.tsx`, the kit CTA's own fixed label colour); the untranslated-name ban (`gold|orange|…`)
is dropped (they are the theme's names now); `lib/tokens.ts` exports exactly the 15 `SEMANTIC_COLOR_TOKENS` (no `overlay`,
no roles, no `NON_COLOR_TOKENS`) and `TokenOverrides` keyed by them.

### 3.4 Boundary and design rules — `run.sh:985-1047`, `1184-1270`, `5854` (W5)

B-WEB-BOUNDARY: `TokenScope` line removed; `lib/appearance.ts` line → "no app file references
`localStorage|sessionStorage|document.cookie|indexedDB`" (the identity allow-list at 5854 becomes empty) except the exact
kit file of S19 plus "no app import of `@/components/ui/sidebar`"; `<script` count in `layout.tsx` → 0 and exactly one
`<ThemeProvider` element inside `<body>`; `suppressHydrationWarning` once on `<html lang="en" suppressHydrationWarning>`
unchanged; `components/layout/theme-provider.tsx` is the only file naming `next-themes` outside `web/src` and declares
`storageKey="appearance-mode"`, `attribute="class"`, `defaultTheme="system"`. B-WEB-DESIGN: literal exemption per S17;
`@font-face` rule unchanged (still only `web/app/globals.css`, exactly 5); `WEB_CSS_URLS`, `WEB_PUBLIC_PINS`,
`DESIGN_KIT_COPY` unchanged (the kit CSS has no `url(`, no `@font-face`, no remote reference; no `public/` asset is copied).
B-NAMING: `web/src` removed from the redundant list with the reason (S1). B-RENDER-PLAN: `web/tsconfig.json` row added,
31 → 32, reasons for `web/package.json`, `package-lock.json`, `app/globals.css`, `app/layout.tsx`, `eslint.config.mjs` updated.
`infra/ci/profile-paths.tsv`: one row `foundation dir web/src/` (UI primitives are foundation like `web/components/ui/`).
**`APPLICABLE` (`run.sh:226`, r1, S25):** `web/src` and `web/tsconfig.json` added with a G069 comment so the `off`
(and any other variant) copy can resolve the kit; `web/components` already covers `legacy-ui` during the transition and
`controls/` afterwards.
`docs/foundation/WEB-SHELL.md`: new `appearance-init` SHA-256 (the served next-themes script; `init_hash_documented` unchanged)
and the G069 cutover sections; `docs/ARCHITECTURE.md` §37A/§61 installed-behaviour records; `.agents/skills/vestive-brand/ADOPTION.md`
one cutover-record paragraph with `LOCAL.sha256` regenerated (its own row); `docs/DECISIONS.md` cutover line.

## 4. Early proof (P0; riskiest assumptions first; inside attempt 1; run by the coordinator) — r1

**Staging decision (Codex item 1).** The early gate cannot be "kit-only typecheck, whole app later" because P0.2 is a
representative integration proof on a **production build** of the app (`next build` in the private web image), and
`next build` compiles and typechecks every route; with the S1 alias flipped, the old consumers would fail that build
until L1–L3 migrate. So r1 moves the **minimum prerequisite mechanical step before the gate** (S21 = S0.0: relocate
the 11 legacy modules, rewrite 54 import specifiers, delete the old `lib/utils.ts`; no API migration) and keeps a
genuinely executable gate: (a) the entire 69-file kit typechecks and lints under a kit-only project that excludes
nothing; (b) the whole app (kit + unchanged legacy consumers) typechecks, lints and builds; (c) the representative
themed tenant overlay + form + preview proof; (d) the saved-branding compatibility proof (S23). The legacy-absence
rules (B-WEB-* etc.) are **first required at phase D**, after completed migration (S20). No kit edit, no hidden
typecheck exclusion, no shadow final implementation: `legacy-ui` is the current implementation at a transitional path,
deleted module by module as the lanes land.

Prerequisites: S0.0 + S0.1 + S0.2 (W0) for P0.1; additionally S0.3/S0.4 (W1, W0) and the P0 oracle subset (W4) for
P0.2; the W5 Core pair draft for P0.3.
Everything Docker runs on the coordinator's checkout of the integrated `g069` branch; writers never run Docker.
`bash infra/ci/docker-desktop-mounts.sh` first; fetch/rebase/reconcile (exact-delta rule) before each batch.
Storage: `<main>/.local/{evidence,secrets,tmp}/g069-a1/` (`<main>` = `<source-checkout>`);
every container/image/volume/network/worktree/temp path in `<main>/.local/evidence/g069-a1/inventory.tsv` at allocation.

**P0.1 — the full catalogue resolves, typechecks and lints with real pinned dependencies (A1–A3, A10).**

1. Resolution (the only networked step; node:24-slim = `web.Dockerfile` base, image ID/RepoDigest recorded):
   `docker run --rm --name g069-a1-npm-<run> --label g069.run=<run> -u <uid:gid> -e HOME=/tmp -v <W0 worktree>/web:/app -v g069-a1-nm-<run>:/app/node_modules -w /app node:24-slim sh -c 'npm ci --no-audit --no-fund && npm uninstall cn && npm install --save-exact --no-audit --no-fund cmdk@1.1.1 clsx@2.1.1 tailwind-merge@3.7.0 next-themes@0.4.6 sonner@2.0.8 react-day-picker@9.14.0 embla-carousel-react@8.6.0 input-otp@1.5.0 react-resizable-panels@4.15.0 recharts@3.10.1 && npm install --save-exact --save-dev --no-audit --no-fund shadcn@4.22.1 && npm ls --all > /tmp/ls.txt; grep -c "node_modules/@radix-ui/" package-lock.json; sha256sum node_modules/shadcn/dist/tailwind.css'`
   Evidence lines: the resolved `@radix-ui/*` name@version set (→ `DEPS_RADIX_VIA_CMDK`), every dependent of each Radix
   entry is `cmdk` or `@radix-ui/*`, `reselect` versions, `zod` ≥ 4.6.6, licences of new entries, the shadcn CSS digest.
   The volume is removed right after `package-lock.json` is copied out.
2. **Kit-only compile/lint** in the same image (network none, read-only source, disposable cache):
   `npx tsc --noEmit -p tsconfig.kit.json` with a scratch `tsconfig.kit.json` that extends `./tsconfig.json`, sets
   `paths: {"@/*": ["./src/*"]}` (the kit resolves only inside the kit, exactly the upstream layout) and
   `include: ["next-env.d.ts","src/**/*.ts","src/**/*.tsx"]` — **all 69 files, no `exclude`**; then
   `npx eslint --max-warnings=0 src` → the exact rule IDs for S18 (recorded; the override lists only those). The
   scratch file is evidence only (kept under `<E>`, never committed). **Stop criterion:** any `tsc` error in `src/**`
   → §1.4 compile stop with the file/error; consumer lanes do not dispatch.
3. **Whole-app check on the transitional state** (S21 applied, alias flipped, kit present, legacy consumers unchanged),
   stack-free in the same image: `npx tsc --noEmit` (the app `tsconfig.json`: kit + app + tests) and
   `npx eslint --max-warnings=0 .`, plus `node --test tests/unit/*` (route/CSRF/branding-client tests unchanged at this
   point). Exit: `tsc` clean, lint clean (A12). **This is the first point where the whole-app typecheck is required,
   and it stays required at every later integration.** Rows 1–2 of §6.1 (the harness with W5's flipped B-WEB-*,
   B-NAMING, B-DEPS rules) are **not** part of P0: they are first required in phase D after completed migration;
   B-DEPS/B-DEPSCAN alone may be run here as evidence (stack-free `lint lint-base` at `--base de74ccbc` with the
   `DEPS_*` delta) because the dependency delta is final at S0.2.

**P0.2 — a representative themed tenant overlay + form with the proposed integration (A4–A9).** The re-modelled
fixture page (`tests/fixtures/token-scopes/page.tsx`, test-only route) renders: `<TenantStyle>` from `fixtureBranded`
(root = tenant), a native `<form method="post">` with kit `Input` (one `aria-invalid`), `Textarea`, `NativeSelect`,
`RadioGroup` (`name`, `defaultValue`, one `disabled` item, `required`; S22 classes on every field and radio), kit
`Button` `size="lg"` as the pending submit (the sign-in shape, S24), kit `Button` default/destructive/outline/secondary,
kit `Table`; a kit `Dialog` open at mount (`?open=root`); one `Select` and one `Popover`; a `Toaster` with one toast;
two sibling previews (`fixtureA` Light-only, `fixtureB` Dark) that open no portal; and, **outside the tenant style's
effect** (a second fixture route or the root landing itself under the platform palette), the one `ButtonLink`
`variant="cta"` with the S15 classes so the CTA's painted white label is sampled at rest, hover (×1.10) and active
(×0.90) and its inset focus outline measured. Private browser stack (§6.1 row 3 command) restricted with
`--grep "S-PORTAL|S-PORTAL-INIT|S-FIRSTPAINT|S-HTML-ATTRS|S-HEAD-ORDER|S-HYDRATION-DEV|S-OVERLAY|S-MOTION|S-STATES|S-FIELD|S-CONTRAST|S-A11Y|S-RESPONSIVE|S-TAB"`
on the six inherited projects. Pass = body insertion with first-frame samples equal to the tenant's values across a live
theme change, no preview value in any portal sample, stored light/dark/system first paint, no hydration mismatch, focus
composites ≥3:1, **the S22 field/radio states** (rest boundary ≥3:1 on `background` and `muted`, invalid solid
`destructive`, disabled 50 %, checked dot) and **the S11 used-state composites** (T8 80 % hover, T9 mix, T10 tints,
T12 fills, CTA ×1.10/×0.90) measured within ±1 of the computed values, submitted `FormData` keys/values (with JS)
including `required`/`disabled` behaviour, Escape focus return, 320 px reflow. **Stop criterion:** a failed guarantee
→ compact Result checkpoint with the exact failing sample and the smallest option (variable value, prop, or §1.4
stop) before L2/L3 dispatch. Elapsed (§8): ≈ 30–40 min including the first image builds.

**P0.3 — saved-branding compatibility proof (S23; A11).** Stack-free: W5's draft r3 pairs run in `unit-core`
(`go test ./internal/branding/ -run 'TestDefaultsPass|TestContrastEdges'`) against every fixture of S23 as table
cases (the seed row's two JSON palettes, `fixtureBranded`, `fixtureGreyPrimary`, `fixtureDarkSparseDestructive`,
BRD-1's `#1d4ed8`, the U-BRD-4 accepted/refused cases), reporting each pair's ratio rather than only pass/fail; then
the rendered check is the P0.2 fixture page's `branded`/`grey-primary`/`dark-sparse-destructive` scopes sampled on
the used states. Output: the compatibility table in the P0 checkpoint (fixture × pair × computed × measured × verdict)
and, for every refused/inaccessible entry, the S23 resolution proposal (expected: the two pre-identified Light hover
cases → D-M3a option (i)). **Stop criterion:** a refused/inaccessible entry without a recorded proposal blocks the
Core delta r3 and L3; unchanged hex is not accepted as evidence.

## 5. Ownership, dependency graph, parallel schedule, workers

### 5.1 Writable paths (disjoint; one writer per file; the coordinator integrates)

| Owner | Writable paths (exclusive) | Content |
|---|---|---|
| **W0** | `web/src/**` (copy only), `web/tsconfig.json`, `web/package.json`, `web/package-lock.json`, `web/THIRD_PARTY_NOTICES.md`, `web/eslint.config.mjs`, `web/components/controls/{button-link,pending-submit,radio-choice,scroll-region}.tsx`, `web/components/controls/field-classes.ts`, `web/lib/utils.ts` (delete), `web/components/ui/**` → `web/components/legacy-ui/**` (S0.0 move; later deleted by the lanes), **and, for S0.0 only, the one-line import-specifier rewrite in the 54 importer files across every lane's paths** — performed before any other writer is dispatched (phase A, first commit on `g069`), so no overlap exists; afterwards those files belong to their lane owners | S0.0 (S21), S0.1, S0.2, S0.4, S18, S22 classes; shared compositions |
| **W1** | `web/app/globals.css`, `web/app/layout.tsx`, `web/components/layout/{theme-provider,appearance-control,app-shell,brand-mark,context-indicator,shell-help-dialog,blueprint-backdrop}.tsx`, `web/components/layout/token-scope.tsx` (delete), `web/components/branding/{tenant-style,brand-preview,tenant-frame}.tsx`, `web/lib/{appearance,tokens,tenant-css}.ts`, `web/app/page.tsx`, `web/app/sign-in/page.tsx`, `web/tests/fixtures/token-scopes/**`, `web/tests/unit/branding-client.test.mjs` (fan-out cases) | S0.3 theme/tenant host, proof page, L1 shell |
| **W2** | `web/components/{membership,enrollment,labels,onboarding}/**`, `web/app/{groups,invitations,join,business-profiles,organization/{members,enrollment,group-authority,labels}}/**` | L2 forms (NativeSelect ×6, RadioGroup ×4 fieldsets, Input/Textarea, Table ×4, typography/layout sweep) |
| **W3** | `web/components/{account,branding/(branding-form,logo-forms),signup}/**`, `web/app/{account,connections,organization/branding,signup}/**` — including `web/app/organization/branding/page.tsx` `PAIRS` (consuming W5's contract text, §2.6) and the branding-form placeholder → description (S22) | L3 account/branding/signup; the landing CTA (S24) |
| **W4** | `web/tests/browser/**` | L4 oracles (P0 subset first: S-PORTAL(-INIT), S-FIRSTPAINT*, S-HTML-ATTRS, S-HEAD-ORDER, `contrast.ts` composite, `modes.ts`) |
| **W5** | `infra/validation/{run,lib}.sh` (incl. `APPLICABLE`, S25), `infra/ci/{profile-paths.tsv,validate-controls.sh}`, `infra/validation/render-customizations.tsv`, `docs/ARCHITECTURE.md`, `docs/DECISIONS.md`, `docs/foundation/WEB-SHELL.md`, `.agents/skills/vestive-brand/{ADOPTION.md,LOCAL.sha256}`, `core/internal/branding/{colors,branding_test}.go`, `infra/identity/mcp-seed.sql` (only the S23 retyped seed values, if D-M3a (i) is disposed) | L5 harness/CI/docs + Core delta r3 (Core semantics/tests and the `PAIRS` contract text; **not** the page file, which is W3's) |
| **Coordinator** | integration branch `g069`, `management/v0/results/G069-*`, allow-file, evidence/inventory/receipt | single integration owner; runs every Docker batch; owns the Result |

Shared contracts settled first (before any dependent lane): the 69-file root and alias (S1), the pins (S2/S3), the CSS/theme/tenant host and the fan-out table (S4–S9), the control compositions' prop surfaces (S12), the harness contract (§3). They are declared here, so W4/W5 can code against them in parallel with W0/W1.

### 5.2 Dependency graph and schedule (→ precedes; ∥ parallel)

```
A: W0 S0.0 relocate+rewrite (first commit) → S1 alias → S0.1 copy → S0.2 pins → [coordinator P0.1 steps 1–3]
   ∥ W5 harness/CI retargets (§3) + draft r3 pairs + PAIRS contract text   ∥ W4 P0 oracle subset
B (after A's commits are on g069): W1 S0.3 theme/tenant host + proof page   ∥  W0 S0.4 compositions + field-classes + S18
P0: coordinator integrates A+B → P0.1 step 3 (rerun on B) → P0.2 → P0.3  (gate; checkpoint incl. the S23 table)
C (after P0 pass): W1 L1 shell + PAIR_FIELD ∥ W2 L2 forms ∥ W3 L3 account/branding/signup + PAIRS ∥ W4 remaining oracles ∥ W5 docs + Core delta r3 values
D: coordinator integrates C (order: W5 harness → W1 → W2/W3 → W4 → W5 docs/Core) → legacy-ui absent → rows 1–5 (first required here)
   → coordinator self-review of S-SCREENSHOTS recorded in the Result (§9.3; owner look optional) → safekeeping push → Codex source review → MR
```

True dependencies (r1): S0.0 before S1/S0.1 (the alias flip would shadow the old modules otherwise); S0.2 before
any `tsc`; S0.3 before P0.2; P0 (incl. P0.3's recorded proposals) before L2/L3 (Spec §5); **W5's pair IDs and
`PAIRS` contract text before W3's page edit and W1's `PAIR_FIELD` edit** (Codex item 4); Core delta values after P0
measurements; L4's spec re-typing after L1–L3 are integrated (support tables and the P0 subset earlier); rows 1–2
(B-WEB-* legacy-absence) only after `legacy-ui` is gone. Reruns are driven by failures/affected checks only.

### 5.3 Workers and isolation

Six writers (W0–W5), **Opus 5.5 high**, each in a native isolated worktree (`isolation:"worktree"`, task-prefixed name
when supported, else the generated `agent-<id>`), **stack-free and Docker-free**: writers edit, self-review and hand back
branches; the coordinator (Sonnet 5.5 high; Opus by owner choice) runs the stack-free `tsc`/`eslint` container on each
handback (no shared lock), every private DB/browser run (two-run pool), every image build (one at a time) and all
harness rows. Fable 5.1 high for test-failure diagnosis/corrections (xhigh at a safe boundary if needed, recorded).
Context: 150k checkpoint / 200k handoff per worker. Finished worker trees are reclaimed after handback once their
commits are on `g069`. Codex: Plan review, Core delta r3 review, source review, Result acceptance.

## 6. Validation (focused executable checks; AC mapping; stale-oracle inventory)

### 6.1 Selectors (`<E>` = `<main>/.local/evidence/g069-a1`; suite names verified: `suite_lint` 346, `suite_lint_base` 360, `suite_lint_web` 367, `suite_controls` 389, `suite_unit_web` 442, `suite_unit_core` 443)

| # | Selector (exact) | Checks / AC | Reuse·extend·new | Slot |
|---|---|---|---|---|
| 1 | `bash infra/validation/run.sh --base 29dded769b14992e27de9150f6306790385c2b5f --allow-file <G069 allow-file> --report-dir <E>/static-<n> lint lint-base lint-web controls` (one invocation; `lint` supplies the canary, `lint-base` holds B-NAMING/B-DEPSCAN/B-LITERALS) | B-DEPS (flipped), B-DEPSCAN (new closure), B-NAMING, B-LITERALS, B-WEB-BOUNDARY/TOKENS/COMPONENTS/DESIGN/ASSET-CONTROLS (A1–A21), B-DEPS-CONTROLS C1–C19, B-SCOPE, B-SECRET-CANARY — **AC1, AC2** | extend (rules re-pinned; fidelity verifier **new** inside B-WEB-COMPONENTS: uncovered today = byte fidelity at the app root + absence of the old runtime; Radix closure rule **new**: the transitive exception is explicit and bounded) | none |
| 2 | same, `unit-web` | B-UNIT-WEB: `eslint --max-warnings=0 .`, `tsc --noEmit` (all 69 + app), `node --test tests/unit/*` (incl. the fan-out cases in `branding-client.test.mjs`), secret-scan; evidence line `npm ls --all` Radix listing — **AC2, AC3** (route/CSRF unit tests unchanged) | reuse; extend one unit file | build slot (web-check image, once per lockfile change) |
| 3 | private browser stack: `COMPOSE_PROJECT_NAME=g069br-<run> docker compose --env-file <main>/.local/secrets/g069-a1/browser-<run>.env -f infra/compose/local.yml -f <override: ports: !reset []> up -d --build --wait`, then `… -f infra/compose/browser.yml --profile browser run --rm --build --no-deps -e PLAYWRIGHT_WORKERS=2 web-browser npx playwright test --project=shell-light --project=shell-dark --project=tokens-light --project=tokens-dark --project=appearance --project=dev-hydration` (P0 adds the §4 `--grep`) | shell/overlay/accessibility/boundary/design/account/tokens/appearance specs: S-APPEARANCE-*, S-FIRSTPAINT*, S-HTML-ATTRS, S-HEAD-ORDER, S-HYDRATION-DEV, S-TOKENS, S-PORTAL(-INIT) rewritten, S-RUNTIME, S-STATES, S-CONTRAST, S-FIELD-*, S-A11Y, S-TEXT-SPACING, S-FONTS, S-BLUEPRINT, S-ANGLE (CtaButton), S-DESIGN/S-FIDELITY/S-PALETTE, S-OVERLAY, S-MOTION, S-TAB, S-RESPONSIVE, S-LONG-IDENTITY, S-SCREENSHOTS — **AC4, AC5** | extend (values/selectors re-typed, scenarios kept); **S-PORTAL(-INIT) rewritten** (the only semantic change: body insertion + first-paint tenant values + no cross-scope leak) | one run slot; build slot during image builds |
| 3i | private identity stack (`infra/compose/identity.yml`, no ports), `IDENTITY_E2E=1 … npx playwright test` with **all 16 identity projects** (`web/playwright.config.ts:105-128`): `--project=identity-light --project=identity-dark --project=domain-platform --project=domain-custom --project=mcp-light --project=mcp-dark --project=onboarding-light --project=onboarding-dark --project=membership-light --project=membership-dark --project=enrollment-light --project=enrollment-dark --project=signup-light --project=signup-dark --project=labels-light --project=labels-dark`; correction loops use `--grep "E-BRD|E-A11Y|E-ONB-2|E-ONB-5|E-MEM-8|E-ENR-5|E-LAB-4"` (E-A11Y added: the pending `aria-disabled` submit, S25) | every form in membership/labels/onboarding/enrollment/signup/mcp changes (select, radio, button, table): submit/default/error/disabled semantics, dialog/select/radio interaction, authorization/CSRF/tenant/host boundaries; E-BRD-1/2/5/6 tenant palette, reset, interleaved hosts; the G068 radio/focus helpers — **AC3, AC4, AC5** | reuse; extend `support/form-controls.ts` for the kit radio | the second run slot |
| 4 | `bash infra/ci/validate-controls.sh --only F-lock-pkg,F-lock-pkglock,F-review-web,F-widen-ui,F-vestive-css,F-vestive-kit,F-vestive-md,F-vestive-link --log <E>/f-controls.log` | profile selection with the one added `web/src/` row — **AC6** | reuse | none (Docker daemon only) |
| 5 | `unit-core` (stack-free; U-BRD-*, incl. the P0.3 fixture table cases and the S24 CTA platform assertion), plus a private per-run database run using the harness's own selectors: `go test -tags integration ./internal/tenancy/access/ -run '^TestBranding$'` (23) and `./internal/migrate/ -run '^TestBrandingCatalog$'` (3) against private PostgreSQL+MinIO at migration 13 (G017/G067 private-run pattern, `network_mode` limited to the run's services) — only if the Core delta lands | accepted/refused palettes under the r3 pairs; the S23 compatibility proof — **AC4** | reuse; extend `branding_test.go` with the fixture table (uncovered today: the saved fixtures' ratios under the kit's actual states) | one run slot |
| 6 | hosted MR `validate:checks` (fast: `lint lint-base lint-web lint-core provenance contracts unit`, B-SKILL-UNCHANGED/B-PROTECTED at `--base HEAD`) + Duo advisory; `validate:integration` not requested | **AC6** | reuse | hosted |
| 7 | post-merge, exact merged SHA, frozen detached checkout: rows 1–2 and 4 only if the merged tree differs from the tested tree; rows 3 and 3i in full; row 5 if Core changed; **plus one `bash infra/validation/run.sh --base <merged SHA> --allow-file <G069 allow-file> --report-dir <E>/off-postmerge off`** (S25: B-TOGGLE-OFF, B-TOGGLE-OFF-ADDITIONS, B-OBS-OFF, B-BROWSER-OFF on the intelligence-disabled copy with the new `APPLICABLE` entries; shared-port harness lock; ports 3500/8580/5932/6879/9500/9501, none held by M1); receipt | **AC6** attributable affected coverage; the disabled-variant composition with the kit | reuse (`off` is an existing suite; no sub-selector exists) | two run slots, sequential builds; `off` takes the harness lock |

Not selected, reported as unexecuted: `root` (M1 holds 543/3100/8180/4100/5532/6479/8100/9100/9101; B-BROWSER-ROOT's
record is root-locked; its scenarios run as private rows 3/3i evidence lines), `rename`, `ops`, `migrations`,
full `ci-controls`, `feature-web`, the full profile; `off` is selected once post-merge only (row 7, S25), not
pre-merge. Upstream-unchanged evidence (AC1) = `git diff 29dded76 -- .agents/skills/vestive-brand/kit`
empty + `sha256sum -c` of the 69 rows + hosted B-SKILL-UNCHANGED; local `provenance` is run at `--base HEAD` like the
hosted job because `ADOPTION.md`/`LOCAL.sha256` legitimately change.

### 6.2 Stale-oracle inventory (oracle changes only where the contract changed; assertions preserved otherwise)

| Oracle | Disposition |
|---|---|
| `tokens.spec.ts` S-PORTAL `:175-191`, S-PORTAL-INIT `:384-410` | **rewrite** to the new guarantee: overlay+content inserted into `document.body` (under Base UI's portal host), every delivery/frame sample equals the root tenant's `popover`/`foreground`/scrim from first insertion and across a live theme change, never re-portalled; new negative: no preview surface's values in any sample; PI-C1..C4 controls keep their meaning; SSR renders no portal (kept) |
| `tokens.spec.ts` S-RUNTIME `:507-580` | retarget: edits `--primary`/`--vx-gold` on a preview element instead of `--light-*` on a scope |
| `tokens.spec.ts` VARIANTS/FILLED/ANGLED `:21-25`, `:1064`; `HOVER_FACTOR`/`PRESSED_FACTOR`/`filledLook` `:842-870` | − `cta` from the tenant fixture scopes (S24: no CTA under a tenant palette); `link` dropped (unused); the filled looks retyped from brightness factors to the kit's alpha model (default `primary/80` over the surface, secondary 5 % mix, destructive text on its tint) — the same rest/hover/pressed sampling kept; **CtaButton probe** retargeted to the landing/platform fixture: fixed white label on `--vx-blue` at rest, hover (`brightness(1.1)`) and active (`brightness(0.9)`), painted values ≥4.5:1, inset focus outline (S15); `ButtonLink` `["A","button"]` kept |
| `tokens.spec.ts` S-FIELD-MOTION `:724`, S-FIELD-INVALID `:691` | motion: kit `transition-[color,box-shadow] 200ms` (retype); invalid: **retained** — S22 keeps the solid `destructive` boundary ≥3:1 against ground and fill in both appearances (`dark:aria-invalid:border-destructive` via className), so the oracle's assertions stand; only the fill expectation moves from `muted` to `input` at 50 % over the surface |
| `fixtures.ts` `fixtureBranded` `:91-126`, `fixtureGreyPrimary` `:133-138`, `infra/identity/mcp-seed.sql:109-110`; `branding_test.go` U-BRD-4 `:97-148` | **retyped only if P0.3 confirms the computed failures** (S23/D-M3a (i)): the Light `primary` edge values move to r3-valid ones with the explaining comments rewritten (grey edge: passes T8 at rest and at 80 % hover; branded/seed: passes every r3 pair in both appearances); Dark values and every other token unchanged; E-BRD-1 reads the seeded values through its existing `tenantColour` probe (no literal to update), E-BRD-5's `#0f6b5c` refusal kept |
| `design.spec.ts` S-FIDELITY help-trigger hover `:411-425` | retype the hover from `brightness(1.12)` to the kit's `bg-primary/80` composite (the "core T8 checked hover" comment follows r3) |
| `appearance.spec.ts` S-HTML-ATTRS `:420-431` | retarget to exactly `[["class", resolved], ["lang","en"]]`, `suppressed` 0 |
| `appearance.spec.ts` S-FIRSTPAINT-DELAYED `:398` | one assertion retargeted (S6); palette assertions kept |
| `appearance.spec.ts` all `data-appearance` probes, `support/modes.ts:77-98` | read `html.classList` + `useTheme` value through the select; storage key unchanged |
| `boundary.spec.ts` S-HEAD-ORDER `:73-97` | retarget: exactly one classic inline theme script (no `src/async/defer/type/nonce`), served before any element below the provider, content equal to next-themes' script; the SHA-256 attachment name `appearance-init.sha256` is kept so `init_hash_documented` and WEB-SHELL.md keep working |
| `support/modes.ts` SOURCES/ROLE_SOURCES `:12-58` | SOURCES retyped from `theme.css` + S10/D-A2 corrections (15 names; `overlay` dropped); ROLE_SOURCES removed |
| `support/contrast.ts` `expectFocusRing` `:186-197` | **rewrite** to the whole-state composite (1 px ring border + 3 px 30 % halo) measured against the field fill and the page; CtaButton inset outline measured as an outline |
| `support/form-controls.ts` `expectRadioOptions` `:135-206`, `expectFocusMotion` `:24-45` | rewrite for the kit radio (`role="radio"`, `aria-checked`, hidden input, disabled 50 %) keeping label association, keyboard, the 3:1 whole-state boundary (S22: solid `input` border on the `bg-input/90` fill), 320 px wrap |
| `design.spec.ts` `source-reference.ts` 32 rows, `BUTTON_SIZES`, `SOURCE_ANGLE`, `FONT_NODES`, S-ANGLE, S-PALETTE roles | retype (cite `web/src/...` class strings; `--vx-angle`; H3 Oswald; roles → foreground/80 and muted-foreground); S-FONTS ×4, S-BLUEPRINT, S-SCREENSHOTS retained (+ extend screenshots with `/sign-in` and one form page) |
| `overlay.spec.ts` `:73-78` scrim probe | retype to `bg-black/30`; S-OVERLAY/S-MOTION logic retained |
| `shell.spec.ts` S-TAB, S-RESPONSIVE, S-LONG-IDENTITY; `accessibility.spec.ts` ×4; `account.spec.ts` | **retain unchanged** (the behavioural guards for S13/S14/A9) except `label[data-appearance-control]` selector if the control markup changes |
| `identity.spec.ts` `:1230-1234,1241,1301-1316,1361-1364,1454-1465` | retarget `[data-token-scope="tenant"]`/`--light-` probes to `html` computed values and the `<style data-tenant-branding>` text; `PLATFORM_PRIMARY` to the S10 value; G068 focus-motion helper per the kit transition |
| 19 unit test files | retain unchanged; `branding-client.test.mjs` **extended** with fan-out emission cases (uncovered: the mapping table) and one `T12:light` `PAIR_FIELD` round-trip (S11) |
| `run.sh`/`lib.sh` rules | §3 (re-pins); controls A1/A5/A12b/C10/C13/C14/C19 retargeted; A20/A21 new; **A22/A23 new** (CTA placement, placeholder; §3.2 h/i); `APPLICABLE` + `web/src`, `web/tsconfig.json` (S25) |

### 6.3 AC mapping

| AC | Evidence |
|---|---|
| AC1 | row 1 B-WEB-COMPONENTS fidelity (69 paths, per-file + aggregate `597f2aac…`), `UI_LEGACY` empty, `git ls-files web/components/ui` empty; upstream unchanged (§6.1 note); hosted B-SKILL-UNCHANGED |
| AC2 | P0.1 lines (resolution, `npm ls`, closure, licences); row 2 `tsc`/`eslint` over all 69 + app; production build = the private web and web-browser images (`npm run build`); row 1 B-DEPS/B-DEPS-CONTROLS (C10/C13/C14/C19 retargeted), B-DEPSCAN; the recorded reason per Radix package |
| AC3 | row 3i whole identity matrix; row 2 unit route/CSRF tests; P0.2 `FormData` evidence; the `.check()`/`selectOption` sites green |
| AC4 | row 3 appearance + dev-hydration + tokens (rewritten S-PORTAL(-INIT), S-RUNTIME), S-PALETTE, S-HEAD-ORDER; row 3i E-BRD-1/2/5/6; the fan-out table and Core delta r3 (row 5) **with the P0.3 compatibility table and every S23 disposition** |
| AC5 | row 3 S-STATES/S-CONTRAST/S-FIELD-*/S-A11Y/S-TEXT-SPACING/S-OVERLAY/S-MOTION/S-TAB/S-RESPONSIVE/S-ANGLE/S-FIDELITY (the retyped alpha-state looks and the CTA probe); row 3i E-A11Y/E-ONB-5/E-MEM-8/E-ENR-5/E-LAB-4; WEB-SHELL.md's documented variable corrections and the S22 field classes with measured ratios |
| AC6 | rows 1–5 on the tested candidate; hosted fast + Duo disposition; authorized merge; row 7 on the frozen merged SHA (incl. the one `off` run); Result AC table; all-category cleanup receipt; preserved failures and known limitations (S16 no-JS radio, E2 scrim, the S23 next-save re-judgement, unused-catalogue defects listed honestly) |

## 7. Allow-list (B-SCOPE; `management/v0/results/G069-scope-allowlist.txt`) and exclusions

Patterns: `^web/src/`, `^web/(app|components|lib|tests)/`, `^web/(package\.json|package-lock\.json|tsconfig\.json|eslint\.config\.mjs|THIRD_PARTY_NOTICES\.md)$`,
`^infra/validation/(run|lib)\.sh$`, `^infra/validation/render-customizations\.tsv$`, `^infra/ci/(profile-paths\.tsv|validate-controls\.sh)$`,
`^core/internal/branding/(colors|branding_test)\.go$`, `^docs/(ARCHITECTURE|DECISIONS)\.md$`, `^docs/foundation/WEB-SHELL\.md$`,
`^\.agents/skills/vestive-brand/(ADOPTION\.md|LOCAL\.sha256)$`, `^management/v0/(plans|results)/G069-.*`,
**r1 addition:** `^infra/identity/mcp-seed\.sql$` (only for the S23/D-M3a (i) seed retyping; `web/components/legacy-ui/`
and `controls/field-classes.ts` are already covered by `^web/(app|components|lib|tests)/`).
**Excluded (never touched):** `.agents/skills/vestive-brand/kit/**` and every other upstream skill byte, `.claude/skills/**`,
`skills/**` (archetype templates), `web/components.json`, `web/next.config.ts`, `web/public/**`, `infra/compose/**`,
`infra/docker/**`, `.gitlab-ci.yml`, `infra/validation/required-checks.tsv`, `core/migrations/**`, every HTTP/MCP
contract, `intelligence/**`, AGENTS/CLAUDE, the G057 frozen checkout and the M1 preview.

## 8. Estimates (evidence-based; unknowns marked), resources, attempt boundaries

### 8.1 Wall-clock by phase (coding ∥ / sequential builds / validation runs / hosted / waits)

| Phase | Parallel coding (per owner) | Sequential builds | Validation elapsed | Waits |
|---|---|---|---|---|
| A: S0.0 relocation (W0, **sequential first**) → S0.1–S0.2 (W0) ∥ harness + r3 draft (W5) ∥ P0 oracles (W4) | W0 1–2 h (S0.0, scripted `git mv` + `sed`, verified by `tsc`) + 2–3 h; W5 5–8 h (r3 draft and the two new controls added); W4 3–4 h | resolution container 3–8 min (registry-bound, **unmeasured**) | kit-only `tsc`+`eslint` 2–4 min; whole-app `tsc`+`eslint`+unit (P0.1 step 3) 3–6 min | W1 waits for A's commits (≈ 3–5 h after release) |
| B: S0.3/S0.4 (W1, W0) | W1 6–10 h; W0 2–3 h | web-check image cold `npm ci` 3–8 min (**unmeasured** with ~100 new entries) | P0.1 step 3 rerun 3–6 min (rows 1–2 **not** here, S20) | — |
| P0.2 + P0.3 | 1–2 h (proof page, W1) + 1 h (fixture table, W5) | web + web-browser images 8–15 min (**unmeasured**; G067 Result had no number) | restricted row 3 ≈ 5–10 min (G067 p09: 146 tests in 4 m 21 s); P0.3 `go test` < 1 min | checkpoint write-up incl. the S23 table ≈ 1 h |
| C: L1 ∥ L2 ∥ L3 ∥ L4 ∥ L5 | W1 4–6 h; W2 6–9 h; W3 4–7 h (PAIRS, placeholder, landing CTA); W4 6–10 h; W5 3–5 h (≈ 1–1.5 working days wall) | — | coordinator `tsc`/`eslint` per handback ≈ 2–3 min | W3/W1 pair edits wait for W5's contract text (phase A output, so no wait expected) |
| D: integration + rows | 4–6 h | image rebuilds per lockfile/source change | per iteration: rows 1–3 ≈ 20–35 min + row 3i ≈ 10–15 min (G067 row 6: 589 s / 171 tests; G068 focused runs minutes); row 5 ≈ 5–10 min; expect 3–4 iterations ⇒ **2–3.5 h** test wall-clock | coordinator screenshot self-review ≈ 0.5 h (no owner wait, §9.3) |
| MR | 1 h | — | hosted fast ≈ 10 min/push (measured, G067); Duo 5–10 min | **Codex source review: unknown** |
| Post-merge | 1 h | rebuilds on the frozen checkout 10–20 min | rows 3 + 3i ≈ 20–30 min; deltas only if the tree differs; **`off` ≈ 25–45 min (unmeasured: variant render, two image builds, migrations, 146 browser tests; G064 recorded only the whole run)** | final ACCEPT: unknown |

**Totals (r1, recomputed):** coding ≈ **55–80 h** across six writers (+1–2 h S0.0, +2–3 h r3 pairs/fixture table/new
controls, +1–2 h CTA/placeholder/PAIR_FIELD consumer edits). **Critical path changed:** it is now W0's sequential
S0.0 → S0.1 → S0.2 → P0.1 (≈ 4–6 h incl. the resolution container) **before** W1 can start S0.3 (previously W1 started
in parallel), then W1 ≈ 14–20 h, then P0.2/P0.3 and the checkpoint (≈ 2–3 h), then integration; i.e. ≈ 20–29 h of
dependent work plus phase C/D; sequential build+validation on the coordinator's path ≈ 4.5–7 h (the post-merge `off`
run added). Realistic wall-clock to a review-ready MR ≈ **5–7 working days**, excluding Codex review rounds (G067,
a smaller migration, took one day for its attempt-1 batches after a 3–4 day estimate; this task has ~6× the file count
and the Core delta). Unknowns: the resolution/image build times, Codex review latency, whether P0.3 finds more than the
two pre-identified cases. **This is a planner estimate, not an accepted deadline.**

### 8.2 Resources

Builds sequential (`COMPOSE_PARALLEL_LIMIT=1`, one build slot); at most two private runs repository-wide (rows 3/3i,
or 3/5); no published port; project names `g069br-<run>`/`g069id-<run>`; M1 (`m1-preview-20261024`, ports above) and
the G057 frozen checkout are never stopped, rebuilt, rebased or reclaimed. Resolution volume `g069-a1-nm-<run>` removed
after the lockfile copy. Inventory reconciled at every handback, batch end, failure and acceptance; compact receipt
with every category (containers, images incl. untagged, volumes, networks, worktrees, temp/evidence paths).

### 8.3 Attempt boundaries (`management/BUDGET_POLICY.md` §3–§4, §7)

Attempt 1 = execution release through the review-ready Result (P0, all lanes, inner loops, reruns, checkpoints,
rebases, handoffs). A bounded supervisor correction after that Result review is attempt 2. B5 maximum 5; checkpoint
after attempt 3; this Plan spends nothing.

## 9. Risks, escalation, screenshots/self-review, closeout

### 9.1 Risks (mitigation)

Upstream type/lint errors in unused files (A1; P0.1 first, stop with exact errors); osv finding in shadcn's CLI closure
(A2; per-finding disposition; a High without fix stops); Turbopack `paths` support (A3; P0 build); next-themes first-paint
semantics (A4; P0.2); focus/field composites (A6; S22 classes measured, app-owned values as fallback); the S23
compatibility findings (A11; proposals ready); the S0.0 rewrite (A12; `tsc` proves it); Base UI radio form API (A7; D-RAD fallback);
oracle re-typing volume (bounded: 32 source-reference rows, two rewrites, two helper rewrites); staging advancing
(exact-delta rule; affected rows only); six-writer merge drift (disjoint paths; the coordinator resolves import-map
residue only); the owner's preview (private projects, no ports, inventory); hover-state contrast (S10 value measured).

### 9.2 Escalation

§1.4 stops (incl. the S23 rule); D-A2/D-M3 are settled (S22/S23) — only the conditional D-M3a (§1.3) can arise, at
the P0 checkpoint; any newly discovered schema/API/authority change → Codex; unused-catalogue cosmetic defects are
listed in the Result as follow-ups, never whole-task blockers (Spec §6).

### 9.3 Screenshots and self-review (r1; no owner gate)

Spec §6 introduces no mandatory screenshot-review gate, and AGENTS' manual-feedback sequencing applies to stages a
Plan already requires — this Plan requires none. The existing S-SCREENSHOTS (landing, sign-in, one form page, the
branding page, an open dialog; 1280 light/dark and 320) are produced by the private row 3 as today, and the
**coordinator records a self-review of them in the Result** (what was looked at, any finding and its correction inside
attempt 1) before the final rows. **Owner feedback is optional**: the screenshots are linked in the review-ready
handback; the owner may comment at any time, and only an explicit owner hold makes it a gate. Nothing waits on it.

### 9.4 Integration and closeout sequence

Safekeeping pushes after each completed round → final rows 1–5 on the tested candidate (fetch/rebase/reconcile first;
target SHA an ancestor of the candidate; the §9.3 self-review recorded alongside, no feedback wait) → Codex source
review → MR to `staging` with source-branch deletion (hosted
fast + Duo; monitoring owner = coordinator, MR/pipeline ID and SHA recorded) → Duo findings triaged and recorded →
authorized merge → frozen detached checkout of the exact merged SHA: row 7 (incl. the one `off` run) → Result AC mapping, known limitations,
attempt log → all-category cleanup (worker trees unlocked and removed after unique commits are on staging, task images
incl. attributable untagged IDs, volumes, networks, evidence reduced to minimal proof) → receipt handed to Codex (no
Result-only MR). Rollback: the MR is one revertable merge; no migration or data change exists; the old skill bytes and
history are untouched.

## 10. Handback

- Plan file: `management/v0/plans/G069-unaltered-brand-kit-plan.md` (this file, r1; sha256 in the handback message),
  copied unchanged to the coordinator's scratch `<home>/.claude/jobs/580cdb2f/tmp/g069/G069-plan-r1.md`
  (r0 is preserved in Git history at `de74ccbc` and as `…/G069-plan-r0.md`, sha256 `50991daf…`).
- Codex: delta review of the r1 correction record (below) — §1 (S6, S10–S12, S15, S20–S25, A6, A11–A12, §1.3, §1.4),
  §2.4–§2.7, §3.2–§3.4, §4, §5.1–§5.2, §6, §7, §8.1, §9.2–§9.4; the Core delta r3 (§2.6) is still reviewed as its own
  bounded delta once P0 supplies the values.
- Nothing installed, built, run in Docker, pushed or opened; the planning worktree holds only this file. G057's
  frozen `c840d714`, M1 and the RC15 promotion hold remain untouched.

## Revision history

- r0, 2026-10-26 — initial Plan from three read-only research lanes (deps/build, themes/branding/portals,
  consumers/checks), with every load-bearing lane claim re-verified against `29dded76` (§0 lists the nine corrections);
  bindings recomputed (`sha256sum`, `git rev-parse`, manifest digest); WCAG ratios computed from hex; no implementation,
  install, build, test, Docker object, push or MR; the owner's preview and the G057 frozen checkout untouched. Zero
  implementation attempts spent.
- r1, 2026-10-26 — bounded correction after the Codex r0 review (see the record below); base `de74ccbc`; planning
  only; zero implementation attempts spent; no other file touched.

### r1 correction record

Each Codex item/decision → the exact sections changed. Everything not listed is **unchanged from r0**, including S1–S5,
S7–S9, S13–S14, S16–S19, A1–A5, A7–A10, §2.1–§2.3, §2.5 (table rows other than `input`/`destructive` notes), §3.1,
§3.3 (apart from the one recorded `outline-white` exception), §5.3, §6.1 rows 1–4 and 6, §6.2 rows not named below,
§7's other patterns and exclusions, §8.2–§8.3, §9.1, the Goal/Spec hashes, the 69-file manifest digest, the
dependency pins, the allow-list's protection of the G057 frozen `c840d714`, M1 and the RC15 hold.

| Codex item | Decision taken in r1 | Sections changed |
|---|---|---|
| 1 (P1) executable early proof | The minimum prerequisite mechanical step moves **before** the gate: S21/S0.0 relocates the 11 legacy modules to `web/components/legacy-ui`, rewrites 54 import specifiers, deletes the old `lib/utils.ts`; the gate = full-kit `tsconfig.kit.json` typecheck/lint (69 files, no `exclude`) + whole-app `tsc`/`eslint`/unit on the transitional state + the P0.2 production-build integration proof + P0.3. Chosen over a kit-only gate because P0.2 needs `next build` of the whole app. Whole-app typecheck first required at P0.1 step 3 and at every integration; legacy-absence rules (B-WEB-*, B-NAMING) first required at phase D; hosted fast on the MR head only. | §0 items 11, 17; S20, S21; A12; §2.7; §3.2 (`UI_LEGACY` incl. `legacy-ui`); §4 (staging decision, P0.1 steps 2–3); §5.1 W0; §5.2; §8.1 |
| 2 (P1) actual CTA colours/states | Verified `cta-button.tsx`: fixed white on `--vx-blue`, brightness 1.10/0.90. S24 confines the CTA to the platform-palette landing (sign-in submit → kit `Button`), so a fixed-white label never meets a tenant `primary`; its contract is pinned in `TestDefaultsPass` (white on platform `primary` ×1/×1.10/×0.90) and a retargeted browser probe, enforced by `UI_CTA_TENANT`. Regular Button alpha checks kept (T8 80 % hover, T9 mix, T10 per-appearance tints corrected, `outline`/`ghost` = T3); **T12** added for text on the kit field fills (used); `link` not tested (unused); pair IDs T1–T10/N2–N4 preserved, `PAIR_FIELD` widened for T12, W5 supplies the `PAIRS` contract and W3 edits the page. | §0 items 13–14; S10, S11, S12, S15, S24; §2.6; §2.7; §3.2 (h), (i), A22/A23; §3.3; §4 P0.2; §6.1 row 5; §6.2 (tokens looks/CTA probe, design hover); §6.3 AC5 |
| 3 (P2) invented owner gate | Removed: existing S-SCREENSHOTS + coordinator self-review recorded in the Result; owner feedback optional unless the owner sets a hold; no "pending feedback = not ready" wording; no wait in §8.1/§9.4. | §9.3 rewritten; §9.4; §5.2 line D; §8.1 D row |
| 4 (P2) ownership/selection | `PAIRS`: W5 owns Core semantics/tests and the contract text, W3 edits the page (dependency in §5.2); identity projects = the actual 16 selectors from `playwright.config.ts:105-128`; `off` analysed from `run.sh:5668-5788`: affected (`APPLICABLE` lacks `web/src`/`web/tsconfig.json`), no sub-selector, no attributable current evidence → one post-merge `off` run on the frozen merged SHA with the reason; disabled-state UI cases named by existing selectors (E-ENR-5, E-A11Y, P0.2 disabled radio, S-STATES); no whole-campaign default, no full ci-controls. | §0 items 10, 12; S25; §2.6; §3.4 (`APPLICABLE`); §5.1 W3/W5; §5.2; §6.1 rows 3i, 7 and "Not selected"; §6.3 AC6; §8.1 post-merge |
| D-A2 (settled: bc, bytes unchanged) | `className` props at the consumer sites (`field-classes.ts`: `border-input dark:aria-invalid:border-destructive` on fields, `border-input` on radio items) + app-owned `--input`; precedence by tailwind-merge within the same utility layer (focus/invalid/disabled variants untouched); the CSS-rule route recorded only as a fallback with its `@layer utilities` requirement; measured in P0.2 (field and radio states). Moved from §1.3 to settled. | S10, S12, S22; A6; §1.3; §2.4; §2.5 `input` row; §4 P0.2; §5.1 W0; §6.2 (S-FIELD-INVALID retained, radio helper) |
| D-M3 (accepted with rule) | No migration; stored values/schema/API/audit unchanged; the new `input`/`destructive`/`primary` rendering and next-save consequences documented; the existing accepted fixtures enumerated (seed `mcp-seed.sql:109-110`, `fixtureBranded`, `fixtureGreyPrimary`, `fixtureDarkSparseDestructive`, U-BRD-4 cases, BRD-1/BRD-4, E-BRD-1/5) with the P0.3 selector; hex ≠ rendering; the §1.4 stop-style rule; two pre-computed failures with a concrete proposal (retype test inputs) and the conditional D-M3a. | §0 item 15; S11, S23; A11; §1.3; §1.4; §4 P0.3; §5.1 W5 (seed path); §6.1 row 5; §6.2 fixtures row; §6.3 AC4/AC6; §7 allow-list |
| D-A1/S6 (accepted) | Moved to settled; no second theme script. | S6 |
| Estimate | Recomputed after the gate correction: coding 55–80 h; critical path now W0's sequential S0.0→P0.1 before W1, then W1, then P0.2/P0.3; 5–7 working days; unknowns marked; planner estimate, not a deadline. | §8.1 |
| Bindings | r1 staging base `de74ccbc` added with the management-only delta (4 files), PROGRESS sha256 re-pinned, r1 worktree recorded; Goal/Spec hashes unchanged. | §0 |


## P0 checkpoint 2 execution amendment — 2026-10-26 UTC

Supervisor disposition: checkpoint 2 review (not included).
The accepted r1 snapshot and frozen P0 inputs above remain historical and are not
rewritten or rebound. Attempt 1 continues under the existing release. Corrected
Core `6bd3bcd4` may land for the planned focused database checks. N4's exemption
is limited to decorative outlines on visibly text-labelled buttons; focus/text/
meaningful-state checks remain. Accept the missing-initial-class System fallback
and the low-priority enabled-button cursor rule. N5 uses a shared consumer solid
destructive focus-border class; the disabled-radio data-state fix needs its rerun.

S4/S8 receive one bounded app-owned addition: after the unchanged kit stylesheet
import, a supported Tailwind dark-variant boundary can exclude the declared Light
preview subtree while preserving Dark previews under Light root. Keep single-level
non-portaled previews, all kit bytes and platform fixture values unchanged. This
supersedes S4's blanket deletion of app-owned custom-variant syntax only for this
necessary boundary; it does not authorize broad component styling overrides. No
fixture-only contrast workaround or extra Core tint admission rule is selected.
Prove actual four-way page/preview appearance states through the existing P0
browser checks and return one corrected handback; no new planning/full suite.

## Owner execution amendment — exact kit visuals, 2026-10-26 UTC

Supervisor records the owner's direct instruction: “Omg waive the contrast
issues... I want it EXACTLY like the brand kit - no compromises”. The
[Spec amendment](../specs/G069-unaltered-brand-kit-spec.md#8-owner-amendment--exact-kit-visuals-and-contrast-waiver-2026-10-26-utc)
and decision record (not included) control
over conflicting S4/S5/S10–S12/S15/S22, P0.2, AC5 mappings and earlier contrast
review dispositions. Preserve the accepted snapshots and failed runs above.

Remove platform token corrections (`input`, `muted-foreground`, `destructive`,
`vx-blue`) and consumer restyling/substituted close controls introduced for
contrast. Use exact shipped visuals; no replacement visual correction. Use kit
compound Select for every application dropdown, including Appearance, superseding
the native/NativeSelect selections. The existing consumer sweep may use matching
kit Label/Field/FieldSet/Alert/Card while preserving semantics and app behavior.
All copied kit source bytes stay unchanged. Functional theme initialization,
tenant-variable mapping and preview isolation remain required.

Inherited contrast failures due to the exact kit treatment/defaults are explicitly
owner-waived. Keep ratios and provenance in the Result, distinguish waived checks
from threshold PASS, and retain unrelated behavior/security/accessibility checks.
Reconcile Core platform fallback/defaults with shipped values without globally
disabling validation of tenant-provided palettes or modifying persisted data.
Claude records the concrete file/test selector mapping with its existing Result
and returns the combined changed head for Codex delta review. My prior hold on
this visual change is lifted; existing workers/attempts continue.

The owner additionally selects full hosted `validate:integration` before merge
on the final MR candidate (alongside hosted fast/review), replacing the full
post-merge run when tested/merged executable trees match. Keep the previously
selected focused post-merge coverage; no local full run is requested. Reconcile
this management-only publication at a safe handoff without restarting active runs.

Later verified owner sequence: existing dropdown and consumer-fidelity writers
finish, then the retained M1 preview is rebuilt from the combined candidate for
the owner's manual check, then the subsequent focused test phase may run after
that check permits it. See the exact imported
coordinator record (not included), owner
instruction 5. Its earlier color HELD status was resolved by the Spec amendment
published as `4a7ced4b`; do not reopen that contrast decision. Preserve already
running bound work and do not infer manual-check completion from elapsed time.

## Core validator execution clarification — later 2026-10-26 UTC

The supervisor approves the bounded resolution to the held Core kit-default
commit `d840429b`: in `CheckContrast`, each applicable pair/state and appearance
uses `min(existing threshold, same pair's ratio on pinned platform kit defaults)`
as its floor. Use the same pair evaluator for candidate and baseline, with no
tenant overrides in the baseline. Keep former thresholds where defaults meet
them, and reject candidates below each applicable floor. No blanket bypass,
data edit or new visual correction. Baseline shortfalls stay owner-waived.

Claude updates the Core implementation and affected unit/branding regression
expectations and records concrete selectors/results in the existing Result.
Cover default/empty and explicit-default palettes, unrelated branding saves,
below-floor negatives and retained thresholds in both appearances. Existing
workers/attempt continue. The owner's manual M1 check precedes the subsequent
focused test phase; this clarification authorizes no early test launch. Combined
source review and full hosted pre-merge qualification still apply. See the
source diagnosis and disposition (not included).
