> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G067 — Vestive Brand successor (Base UI/Rhea overhaul) for generic SaaStacular: Implementation Plan (r1)

Revision: **r1, 2026-10-25 UTC** — one consolidated correction of r0 under the Codex r0 review
(`management/v0/results/G067-plan-r0-codex-review-20261025.md`, sha256 `aa84dc3e9fa5c1d6016e26242cdf15d25ac1db19b7fc91b297dae1393d9f21db`):
(1) kit geometry measured from the real sources, (2) P0 reordered and the dependency/CLI/font inputs resolved by
read-only registry/documentation discovery, (3) two-step intake verification and an accurate rights/lineage record,
(4) `.local` task storage, accounted disposable creators, corrected selectors and a focused private rename proof,
(5) B5 attempt boundaries per `management/BUDGET_POLICY.md` §3–§4, (6) two named parallel owners with an explicit
dependency graph, the smallest source-faithful CSS/utility integration and a recomputed estimate. r0 text is kept
where the review did not correct it. **This revision releases nothing: no implementation, install, build, pull, test,
Docker object, push or MR; the owner's frozen preview `m1-preview-20261024` and its stack were not touched**
(r0's single read-only `docker ps` listing of its ports stands, §7.3). Author: Claude (Fable 5.1, high).
Goal: [G067](../goals/G067-vestive-brand-successor.md) (sha256 `e5fe85809d97af547c7ae4e9fcd92f8c7192a9e355d2ca3b886f3f9eb589b711`).
Spec: [r2](../specs/G067-vestive-brand-successor-spec.md) (sha256 `918514be044d43ee3d256656a70535a03cf756c52bfef5a29adf7e1a73356dba`;
r1 preserved as `G067-vestive-brand-successor-spec-r1.md`). ADR: ADR-001 with the owner amendment of 2026-10-25 (not included).
Bound to `origin/staging` **`0ff598f21bcfe2cfd5e319c69b7493f5fb6cf5e2`** (PROGRESS sha256
`1c4e1f9bbb71f3fde0acf42343cbce75b7360fee5501fb293620a653eeb068d4`). `0e957c69` → `0ff598f2` is one management commit
(SUPERVISOR-HANDOFF, PROGRESS, the r0 review); `b6bfffb0` → `0e957c69` was docs/management only (ARCHITECTURE
§37A/§38/§61, DECISIONS, ADR-001, ROADMAP, PROGRESS, SUPERVISOR-HANDOFF, Goal, Spec). No `web/`, `infra/` or skill file
changed across these three SHAs, so every source citation below holds at all of them. This r1 commit sits on the r0
commit `bb1175f3` (branch `g067-plan`), whose parent is `0e957c69`; the branch is not rebased onto `0ff598f2` (the
staging delta is management-only and Codex imports accepted Plan bytes into staging as usual). Budget: **B5,
checkpoint after attempt 3; zero implementation attempts spent** (§8.2 states the attempt boundaries).
Implementation begins only after Codex Plan ACCEPT and ordinary readiness.

## 0. Verified inputs and established facts

- **Supplied source re-verified (read-only).** `<home>/.agents/skills/vestive-brand`: 319 regular files, 0
  symlinks, 2,694,542 bytes; `SKILL.md` sha256 `188ab915be8d8a98a3563477306c54cfb8cd162eaf34a805d7b4699443904662`; tree
  digest (Spec algorithm) `67f92e83e2cc61d24691baf2eebe5893261eebdc7056fc863d69897aa602ac88` — all equal the Spec.
  Layout: `SKILL.md`; `kit/` (`README.md`, `BRAND.md`, `tokens.json`, `public/brand/` 11 images, `src/`: 54
  `components/ui/*.tsx` on **Base UI** subpaths (`@base-ui/react/{button,dialog,input,menu,select,…}`, plus
  `use-render`/`merge-props` in 5 files), 7 `components/vestive/*`, `globals.css` (imports `tailwindcss`,
  `tw-animate-css`, **`shadcn/tailwind.css`**, `styles/theme.css`, `styles/vestive.css`), `styles/{theme,vestive,
  ribbon-wave,rhea-leftovers}.css`, `hooks/use-mobile.ts`, `lib/utils.ts` (clsx + tailwind-merge)); `design-system/`
  (brand book, `tokens.json`, `design-system.json`, `assets/{Logos,Social}`, 78 component READMEs, a duplicate
  `components/src/` tree, prebuilt `bundle.js` 1.29 MB + `bundle.css` + `index.d.ts`, minified React). Third-party
  imports across the kit: `react`, `lucide-react` (20), `class-variance-authority` (16), and per-component `recharts`,
  `sonner`, `react-resizable-panels`, `react-day-picker`, `next-themes`, `input-otp`, `embla-carousel-react`, `cmdk`
  (one file each; none consumed here). **No licence or notice file in the tree.**
- **Old repository skill.** `.agents/skills/vestive-design/` (81 files) and alias `.claude/skills/vestive-design` (120000).
- **Current web primitive footprint (the migration inventory, §3).** Direct: `@radix-ui/react-dialog` 1.1.23,
  `@radix-ui/react-slot` 1.3.3. Lockfile closure of those two: **25 `packages` entries** — 16 `@radix-ui/*`
  (`primitive`, `react-compose-refs`, `react-context`, `react-dialog`, `react-dismissable-layer`, `react-focus-guards`,
  `react-focus-scope`, `react-id`, `react-portal`, `react-presence`, `react-primitive`, `react-slot`,
  `react-use-callback-ref`, `react-use-controllable-state`, `react-use-effect-event`, `react-use-layout-effect`) and 9
  non-scoped (`aria-hidden`, `detect-node-es`, `get-nonce`, `react-remove-scroll`, `react-remove-scroll-bar`,
  `react-style-singleton`, `tslib`, `use-callback-ref`, `use-sidecar`); only `tslib` is also required by non-Radix
  packages (`@swc/helpers`, `@emnapi/*`, `@tailwindcss/oxide-wasm32-wasi`, `@tybys/wasm-util`) and stays. No other
  dependency of web depends on anything in that closure (`cn`, `lucide-react`, `class-variance-authority`,
  `tw-animate-css` have no deps into it). Radix API consumers: `components/ui/button.tsx` (Slot/`asChild`),
  `components/ui/dialog.tsx` (`DialogPrimitive.*`), 17 production `asChild` sites in 11 files (r1 recount: 16 links, 1 `DialogTrigger`; r0 said 19), 2 fixture sites, 1 test
  block (`tokens.spec.ts:807–831`), 1 test assertion on `data-radix-focus-guard` (`overlay.spec.ts:19–21`), the pinned
  ejected registry block in `globals.css` (`--radix-accordion-content-height` fallbacks, `/* ejected from shadcn@4.21.1 */`,
  `components.json` `style: radix-nova`), and the harness rules that **require** Radix / **forbid** Base UI (§6).
- **Architecture now.** ADR-001 amendment: Base UI/Rhea is the successor foundation; Radix must leave application/test
  imports, active configuration and the resolved tree, direct and transitive, after consumers migrate; no parallel
  Radix foundation and no Radix restyle. §37A/§38/§61 and DECISIONS agree. **No schema, migration, HTTP/MCP contract,
  authority or tenancy change.**

### 0.1 Facts added in r1 (read-only discovery; nothing installed or run)

- **Kit geometry, measured from the sources** (`kit/src/components/ui/*.tsx`, `kit/src/styles/{theme,vestive}.css`;
  root 16px): Button `size` → `xs` `h-6` 24px, **`sm` `h-7` 28px, `default` `h-8` 32px, `lg` `h-9` 36px**, `icon`
  `size-8` 32px, `icon-xs` 24px, `icon-sm` 28px, `icon-lg` 36px; 1px transparent border (`border border-transparent
  bg-clip-padding`), `text-sm font-medium` (14px/500). `Input` and `NativeSelect` **`h-8` 32px** (`NativeSelect`
  `size="sm"` `h-7` 28px), `px-2.5`, `text-base md:text-sm` (16px, 14px from 768px), `Textarea` `min-h-16` 64px.
  Control radius **`rounded-2xl` = `calc(var(--radius) * 1.8)` = 1.125rem = 18px** with `--radius: 0.625rem`
  (`theme.css:40`; multiples `vestive.css:57–63`, identical to the repository's `globals.css:769–775`). Dialog
  `Popup` radius `min(var(--radius-4xl), 24px)` = min(26px, 24px) = **24px**, `p-6`, `gap-6`, `sm:max-w-md` (448px),
  `ring-1 ring-foreground/5`, `shadow-xl`; Close = `Button size="icon-sm"` (28px) at `top-4 right-4`; `Title`
  16px/500 `leading-none`; `Description` 14px `muted-foreground`. `Table`: head `h-10` 40px `px-2 font-medium`, cells
  `p-2`, 1px `border-b` rows. Rules are 1px `border-border` hairlines. Nothing adopted renders at the 10px base
  radius; r0's 32/36/40px buttons, 36px fields and "10px controls" were wrong and are corrected throughout (§5.2, §7.2).
- **`@base-ui/react` 1.9.0** is the registry `latest` (MIT; `https://registry.npmjs.org/@base-ui/react/latest`;
  integrity `sha512-S8CSYosASBKtpL8SWUfS78rFElfLijiCgA7oHOK22jCJ95FULsYyKwBNFOgudt0RA+/7yNc7F81bOvQnRgbBiQ==`;
  tarball `https://registry.npmjs.org/@base-ui/react/-/react-1.9.0.tgz`; subpath exports `./button`, `./dialog`,
  `./input` present). Declared closure (every manifest read from the registry): `@base-ui/utils` 0.5.0 (→ `reselect`
  ^5.3.0 → 5.3.0, `@babel/runtime`, `@floating-ui/utils`, `use-sync-external-store`), `@babel/runtime` ^7.29.7 || ^8.0.0
  (latest 8.0.8), `@floating-ui/utils` ^0.2.12 (0.2.12), `@floating-ui/react-dom` ^2.1.9 (2.1.9 → `@floating-ui/dom`
  ^1.8.0 → 1.8.0 → `@floating-ui/core` ^1.8.0 → 1.8.0), `use-sync-external-store` ^1.7.0 (1.7.0); all MIT; **no
  `@radix-ui/*` or `radix-ui` anywhere in the declared closure.** Peers: `react`/`react-dom` ^17 || ^18 || ^19 (19.3.0
  satisfies); optional `@types/react`, `date-fns`, `@date-fns/tz` (not installed). None of the nine packages is in the
  current lockfile, so the expected new `packages` entries are exactly these nine; `tslib` and `clsx` (already present
  for other packages) are unaffected.
- **Registry CSS and style.** `shadcn@4.22.1` (registry `latest`) and `shadcn@4.21.1` ship a byte-identical
  `dist/tailwind.css` (16,340 bytes, sha256 `4c371f7a1ff5d219ae2f7ff28bd256b4346fd546fe46fbae22092e57db2f0fae`), and
  the repository's pinned ejected block (`globals.css` lines 1–640 = the two `@import`s + the marker + that file; sha256
  `16693ac5674b28684fc8398f8442bc36cba1b2984b7ccf6cda2c6ebc5bc6c1e8` = `GLOBALS_EJECT_SHA`) **already contains exactly
  the file the kit imports as `shadcn/tailwind.css`**. The shadcn CLI's own `eject` is documented as "inline
  `shadcn/tailwind.css` into your global CSS file and remove the `shadcn` dependency". The Rhea style exists in the
  registry for Base UI as **`base-rhea`** (`https://ui.shadcn.com/r/styles/base-rhea/{button,input,dialog}.json` HTTP
  200; `radix-rhea` and `base-nova` also 200, so the name is `<base>-<style>`); its `button`/`input`/`dialog` differ
  from the kit's only by class order, the `cn` import path (**the registry imports `cn` from `"cn"`**, the package the
  repository already pins), the kit's `XIcon` in place of the registry's icon placeholder, and the registry's
  `cn-font-heading` title class. `baseColor` `taupe` is a documented allowed value. The preset code `b1Ha5oWoa` was
  not decoded (that needs the CLI); its composition is taken from `kit/README.md` ("Base UI · Rhea · Taupe · Amber
  charts · Geist · Lucide · medium radius") and is consistent with the registry evidence.
- **Base UI documentation (v1.9.0).** `Dialog.Portal` has `container` (`HTMLElement | ShadowRoot | RefObject`) and
  `keepMounted`; `Dialog.Root` `modal` defaults to `true` (focus trapped, scroll locked, outside pointer disabled);
  `Popup` has `initialFocus`/`finalFocus` (default: first tabbable; on close the trigger or previously focused element);
  parts expose `data-open`/`data-closed`/`data-starting-style`/`data-ending-style`; `Trigger`/`Close` take `render`
  and `nativeButton` (default `true`; set `false` when the rendered element is not a button). **The Button page states
  that links "have their own semantics and should not be rendered as buttons through the `render` prop"** — style an
  `<a>` directly. The focus-guard attribute name is not documented and stays P0-bound (A2).
- **Fonts.** `google/fonts` at commit **`9e25e2ba265e5298f70f6182dd4e8a3ebf1b9123`** (2026-05-20; the latest commit
  touching `ofl/geistmono`, after the latest `ofl/geist` commit `718e1db4`): `ofl/geist/Geist[wght].ttf` 169,056 B
  sha256 `73894e0448cae90a92b6c2f8732b7bb9acb7b94c418bff559dad4a18e1de9659` (blob `f63f0afc6390323715972b6645115485f37cc9f4`);
  `ofl/geistmono/GeistMono[wght].ttf` 171,948 B sha256 `d00e590b8eb3a59acc329b2d044fd143ae935090b7da33199ebee27cc7de8196`
  (blob `173867dce0580ea751e4b9c558740d34f20ce549`); `OFL.txt` 4,387 B sha256
  `1781d2806a07d91c4edf4740b88449fab7d0eadad53f7c351b94cd4d4eb8c00f` (same blob `61f6f4b7` in both directories);
  `METADATA.pb` copyright "Copyright 2024 The Geist Project Authors (https://github.com/vercel/geist-font)", licence
  OFL. Same mechanism as G048's Oswald/Barlow (`7085eb89`). Italic faces exist but have no consumer and are not bundled.
- **Harness facts.** `suite_lint_web` (`run.sh:367–373`) runs only the five `static_web_*` checks and generates no
  secrets, so a `lint-web`-only invocation ends with `B-SECRET-CANARY UNMET` (`run.sh:6668`); `suite_lint` creates
  `static.env` and the B-SECRET-SOURCE canaries, and `unit_service` creates its own environment file, so `lint …` and
  `unit-web` invocations meet the canary. `variants_rename` (`run.sh:5549–5603`) refuses busy `PORTS_REN` (3300 8380
  8300 5732 6679 9300 9301), starts README `platform:start` on a disposable copy with those ports, then runs
  `@identity` in shell-light/shell-dark and renames the database; it is neither port-free nor browser-only.
- **`asChild` consumers, re-read (exact grep).** 17 production sites in 11 files: 16 wrap an `<a href>` (links) and the
  17th is `shell-help-dialog.tsx:26` `DialogTrigger asChild` around a `Button`. The fixture has one `DialogClose asChild`
  (`Button`) and one `Button asChild` link; `tokens.spec.ts:807–831` asserts the link renders as `["A", "button"]`.
  `pending-submit.tsx:7` omits `asChild` from `ButtonProps`.

## 1. Settled decisions, unproven assumptions, early proof

### 1.1 Settled

| # | Decision | Basis |
|---|---|---|
| S1 | Copy all 319 files byte-exact to `.agents/skills/vestive-brand/` (`cp --no-preserve=mode`; source files are 0777 on disk, Git records 100644); alias `.claude/skills/vestive-brand -> ../../.agents/skills/vestive-brand`; nothing excluded (the prebuilt bundle and minified React are inputs, never imported; §6 exclusions keep them out of web scans). Local additions only: `ADOPTION.md`, `UPSTREAM.sha256` (the 319-row upstream manifest; its sha256 = the tree digest), `LOCAL.sha256` (the final-tree manifest) and a 4-line "Repository adoption" pointer appended to `SKILL.md`. Verification is two-step (§4.1): `sha256sum -c UPSTREAM.sha256` passes on the pristine copy **before** the `SKILL.md` edit; `sha256sum -c LOCAL.sha256` plus the bound local delta passes on the final tree. | Spec R1, R3 |
| S2 | **Base UI is the primitive foundation; Radix is retired completely.** `components/ui` modules are derived from the new kit's actual Base UI sources (`kit/src/components/ui/{button,dialog,input,textarea,native-select,table}.tsx`) and brand layer (`components/vestive/{typography,layout,content,site-header}.tsx`), not from the old Radix files. Consumers migrate from `asChild`/Slot: the 16 link sites become styled anchors (`ButtonLink`, an `<a>` carrying `buttonVariants()` and `data-slot="button"`, exported from `button.tsx` — Base UI's documented pattern for links, §0.1), the `DialogTrigger`/`DialogClose` sites use Base UI's `render` prop (`render={<Button … />}`); the Dialog maps `Overlay/Content` → `Backdrop/Popup` with the scope-portal adapter rebuilt on `Dialog.Portal`'s `container` prop. | ADR-001 amendment, Spec R2, Base UI Button docs |
| S3 | Dependency delta (exact pins only; §4.2): **remove** `@radix-ui/react-dialog`, `@radix-ui/react-slot`; **add `@base-ui/react@1.9.0`** (`--save-exact`; the nine-package closure of §0.1 is the expected lockfile delta); keep `class-variance-authority` 0.7.1, **`cn` 0.4.0** (the registry's own `base-rhea` output imports `cn` from `"cn"`; the kit's `lib/utils.ts` (clsx + tailwind-merge) is the kit-local variant of the same merge semantics and is not adopted — `lib/utils.ts` stays `export { cn } from "cn"`; C19 keeps `tailwind-merge` out), `lucide-react` 1.52.0, `tw-animate-css` 1.4.0, Next/React pins. **`shadcn` stays a never-dependency and no CLI is run:** the file the kit imports as `shadcn/tailwind.css` is byte-identical at 4.21.1 and 4.22.1 and is already the repository's pinned ejected block (§0.1), so `globals.css` lines 1–640 and `GLOBALS_EJECT_SHA` are **unchanged** — adding the 32-dependency CLI package as a runtime dependency for one static CSS file that the CLI's own `eject` inlines has no concrete reason. No `next-themes`, no `clsx`, no per-component catalogue packages. | ADR-001 amendment 2, Spec R2 "add no unused dependency catalogue", §0.1 |
| S4 | The 16-token contract, the 3 internal roles, `--radius`/`--angle` names, the five-place equality, `TokenScope`, `tokenOverrideStyle`, the `appearance-init` script and Light/Dark/System stay in **shape** (public override input contract preserved: Spec R3/R4). Source slot semantics that differ are translated at the integration boundary: the kit's neutral `accent` slot → the app's `muted`; the app's `accent` stays gold ink; kit `card` → `popover`; kit `ring` = gold = app `ring`. | Spec R3 "explicit translation/aliases at the integration boundary" |
| S5 | App UI = plain Rhea with brand colour: Geist/Geist Mono bundled locally (Roboto retired), Taupe neutrals, `--radius: 0.625rem` with the pinned registry multiples (= the kit's), hairline rules, Rhea button/field/dialog/table geometry; Oswald/Barlow only for brand moments (§5.2). | Spec R2, `BRAND.md` "Two layers" |
| S6 | White-on-blue/green source labels (4.21:1 / 4.05:1 at 14px medium) are **not** adopted: `primary` keeps the measured derivation `#0062d6` (white 5.65:1) in both appearances; `secondary` becomes the kit's neutral fill; the kit's tinted destructive variant (4.01:1 in Light) is replaced by a solid fill. Ratios in §5.1. | Spec R3 (explicit), WCAG 1.4.3 |
| S7 | Old treatments that conflict with the successor are retired (2px rules, zero radius, uppercase Oswald everyday controls/headings, Barlow labels, angled everyday buttons, the `muted` variant). The angled silhouette survives only as a `cta` Button variant (D-3). | Spec R2 |
| S8 | One MR; the old skill directory and alias are deleted in the **last** commit together with every live retarget, so no staging commit names an absent skill; until then every writer still reads the old skill as AGENTS.md directs. Radix packages are removed in the same commit that removes the last Radix import (no half state in any commit that the hosted fast job builds). | Spec R1/R2 |
| S9 | Focused validation only (§7); no `ci-controls`, no full local run, no `validate:integration` request; post-merge affected coverage on the exact merged SHA. | Spec, CLAUDE.md |
| S10 | Vestive logos, icons, social glyphs, marketing copy, prices, plan names and `SiteHeader` defaults never enter `web/` (B-WEB-DESIGN literal/byte-copy rules retargeted to the new image folders). | Spec R2 |

### 1.2 Unproven assumptions and their proofs

| # | Assumption | Proof / fallback |
|---|---|---|
| A1 | **Resolved in r1 (§0.1):** `@base-ui/react@1.9.0` with a declared nine-package closure that contains no `@radix-ui/*` or `radix-ui` package and whose peers accept React 19.3. Remaining assumption: npm's actual resolution equals the declared closure (no undeclared optional/bundled package). | **P0 step 1** (§1.3): resolve the lockfile in the accounted container and compare the new `packages` entries with the expected nine; any extra entry is reported (a Radix entry stops the work). |
| A2 | Base UI `Dialog` satisfies the retained guarantees: `Dialog.Portal` accepts `container` (scope element), a dialog open at mount can render nothing until the scope element exists and then portal into it (first-insertion scope, S-PORTAL-INIT), modal mode makes the outside inert/`aria-hidden`, initial focus moves into the popup, Escape/Close return focus to the trigger, `aria-modal` is set, SSR/hydration under `next dev` reports no mismatch, `render` composes `Button` for Trigger/Close (links use `ButtonLink`, not `render`). | **P0 step 5** (§1.3) runs S-OVERLAY, S-MOTION, tokens-* S-PORTAL-INIT/scope/inert cases, S-HYDRATION-DEV and S-A11Y on the migrated shell + fixture. A failed guarantee is adapted in the repository wrapper (allowed by Spec R2) or, if Base UI cannot provide it, reported as a blocker. |
| A3 | **Resolved in r1 (§0.1):** the registry CSS the kit imports is already the pinned ejected block (identical at `shadcn@4.21.1` and `@4.22.1`), so no eject/CLI run is needed; the Base UI Rhea style name is `base-rhea` and `baseColor` `taupe` is an allowed value. Remaining assumption: none affecting the build (the `style`/`baseColor` strings only steer future `npx shadcn add` calls, which this task does not make). | **P0 step 2** writes `components.json` (`style: "base-rhea"`, `baseColor: "taupe"`) and records the evidence line `sha256sum` of `globals.css` lines 1–640 = `GLOBALS_EJECT_SHA`; the marker stays `/* ejected from shadcn@4.21.1 */` because those are the bytes (WEB-SHELL.md notes the 4.22.1 identity). |
| A4 | **Resolved in r1 (§0.1):** Geist and Geist Mono variable TTFs and `OFL.txt` exist in `google/fonts` at `9e25e2ba` with the recorded sizes/sha256/blob SHA-1s. Remaining assumption: the bytes fetched during implementation equal those digests. | Intake step I2 (owner A): fetch by commit URL (`https://raw.githubusercontent.com/google/fonts/9e25e2ba265e5298f70f6182dd4e8a3ebf1b9123/ofl/...`), `sha256sum` must equal §0.1 or the intake stops and reports; `custom` render verdicts come from S-FONTS in P0 step 6. |
| A5 | The browser oracles (32 `SOURCE_REFERENCE` rows, S-DESIGN geometry, S-FONTS, `SOURCES`/`ROLE_SOURCES`, `BUTTON_SIZES`, `VARIANTS`) can be re-typed from the new kit without weakening any assertion. | P0 step 5 counts failing assertions on the shell before the visual sweep (budget ≤ ~60; the measured values of §0.1 keep the re-typing mechanical). |
| A6 | The secret scanner does not false-positive on the vendored `bundle.js`/minified React. | Owner B runs the README `secret-scan` line against its worktree right after I1 (one stack-free `docker compose run`, recorded); row 1's B-SECRET-SOURCE and hosted provenance repeat it. A finding gets an owner-reviewed exemption, never a silent removal. |
| A7 | The identity-stack browser specs (155+ tests) assert behaviour and per-page contrast, not old geometry literals. | Read-only grep in P0; hits listed in the Result. |

### 1.3 Early proof (P0, inside attempt 1, before the visual sweep)

Riskiest assumptions now: **A2 (Base UI covers the consumers' dialog/button/portal guarantees)** and A5 (oracle
re-typing stays bounded); A1/A3/A4 are resolved pins that P0 merely confirms. P0 runs in owner A's isolated worktree
(§8.3) and is ordered so that **the application typechecks before its first whole-app check**:

1. **Dependencies (A1 confirmation).** In the accounted dependency-resolution container (§7.3: `node:24-slim` as the
   repository's own `web.Dockerfile` base, image ID/digest recorded at allocation; `node_modules` in a run-labelled
   disposable volume; registry access for this step only): `npm uninstall @radix-ui/react-dialog @radix-ui/react-slot`
   and `npm install --save-exact @base-ui/react@1.9.0` → inventory `package-lock.json` (`grep -c '@radix-ui' = 0`, the
   24 removed entries listed, the new entries compared with the nine expected packages/versions/integrities of §0.1),
   then `npm ls --all | grep -ci radix` = 0. An unexpected entry is reported; a Radix entry stops the work.
2. **Configuration (A3 confirmation).** `components.json` `style: "base-rhea"`, `baseColor: "taupe"`; evidence line
   that `globals.css` lines 1–640 still hash to `GLOBALS_EJECT_SHA` (no eject run, §4.3).
3. **Complete mechanical API/type migration** (all in one coding pass, no whole-app check in between): `button.tsx`
   (kit `Button` on `ButtonPrimitive` + `ButtonLink` + `cta`), `dialog.tsx` (Base UI parts + rebuilt scope portal),
   `input.tsx`; **all 16 production link sites → `ButtonLink`**, `shell-help-dialog.tsx` `DialogTrigger` → `render`,
   `pending-submit.tsx` type (`Omit<…, "type" | "children">`), the fixture's `DialogClose`/link sites,
   `tokens.spec.ts:807–831` (anchor with `data-slot="button"`), `VARIANTS`/`FILLED`/`ANGLED`/`BUTTON_SIZES` (−`muted`,
   +`cta`), `overlay.spec.ts` focus-guard selector; Geist/Geist Mono files + `@font-face` + `--font-sans/--font-mono`
   (A4 confirmation); the §6 harness flips (`DEPS_*`, B-WEB-COMPONENTS import rules, `components.json` expectation,
   `UI_SLOTS`, `WEB_PUBLIC_PINS`/`WEB_CSS_URLS`). Other `components/ui` modules, palette values and the D-2/D-3 surfaces
   wait for F/V/C.
4. **First whole-app static/type/unit check:** row 1 (`lint lint-web controls`) and row 2 (`unit-web`: `eslint`,
   `tsc --noEmit`, unit tests, secret-scan, plus the `npm ls` evidence line). Exit: B-DEPS/B-DEPS-CONTROLS PASS with the
   flipped policy, B-WEB-COMPONENTS/B-WEB-TOKENS/B-WEB-DESIGN PASS, B-SECRET-CANARY PASS, `tsc` clean.
5. **Behaviour proof on the private browser stack** (row 3 restricted to `overlay|tokens|design|accessibility|
   appearance` specs): S-OVERLAY (name, initial focus, containment, outside inert, Escape/Close, focus return, Close
   target ≥ 24px — `icon-sm` is 28px), S-MOTION, S-PORTAL-INIT and the nested/partial/sibling scope cases on the
   rebuilt `container` adapter, S-HYDRATION-DEV, S-A11Y, S-FONTS (Geist/Geist Mono `custom` on at least one node).
   Exit: those PASS on Base UI; remaining failures are typed-value re-types (geometry, palette, fonts) within the A5
   budget (≤ ~60 failing assertions on the shell before T).
6. Anything else — a guarantee Base UI plus the repository wrapper cannot meet, a Radix entry, a font digest mismatch —
   is a compact Result checkpoint before continuing (E1/E4/E6).

## 2. Approach (two named owners, disjoint writable paths; §8.3)

Owner **B (intake/docs)** and owner **A (web)** start together after execution release; A integrates B's branch before
the last commit. The Goal authorizes exactly this parallelism ("independent intake/docs and web work may run in
parallel with named ownership and nonoverlapping writable paths after Plan approval").

1. **Intake (I, owner B).** Skill copy, alias, `UPSTREAM.sha256` + V1, `ADOPTION.md`, `LOCAL.sha256` + V2, `SKILL.md`
   tail; docs retargets (ARCHITECTURE §61, OPERATIONS, WEB-SHELL G067 sections, AGENTS/CLAUDE skill paths);
   `infra/ci` retargets (`lib.sh`, `profile-paths.tsv`, `validate-controls.sh` fixtures) and row 4.
2. **P0 (owner A).** §1.3: dependencies, configuration, the complete mechanical API/type migration incl. fonts and
   harness flips, then the first whole-app checks and the behaviour proof.
3. **Foundation + primitives (F/V, owner A).** `globals.css` palette/type/geometry re-sourced from
   `kit/src/styles/theme.css` into the `--light-*`/`--dark-*` contract; the remaining `components/ui` modules
   (textarea/native-select/table on native elements; typography/layout/bullet-list/site-header from
   `components/vestive/*`), each with a lineage header naming the tracked kit path + sha256.
4. **Consumers (C, owner A).** Class re-maps in the 33 files that use font/geometry utilities directly; `muted` →
   `secondary`; D-2/D-3 surfaces; copy, routes, forms, errors, conditional behaviour and authority untouched.
5. **Oracles (T, owner A).** Re-type expectations from the measured kit values (§0.1/§5.2); re-pin fonts/palette;
   extend the screenshot captures; no scenario deleted.
6. **Guards and docs (G, owner A after merging B).** `run.sh` skill-path retargets, `THIRD_PARTY_NOTICES.md`, old skill,
   alias and Roboto deleted in the last commit.

## 3. Migration inventory (at `0e957c69`)

### 3.1 Shared modules and consumers

| `web/components/ui/` module | New kit source | Consumer import lines (`app/`, `components/`, `tests/`) |
|---|---|---|
| `button.tsx` (Slot, `asChild`) | `kit/src/components/ui/button.tsx` (`@base-ui/react/button`) + `ButtonLink` (styled `<a>`, Base UI's documented link pattern) + a `cta` variant from `components/vestive/cta-button.tsx` | 15 (`<Button` 21; variants `outline` 22, `secondary` 2, `destructive` 1; `size="sm"` 14; **`asChild` 17 production sites, 16 of them links → `ButtonLink`**: `app/connections/mcp/page.tsx` ×2, `app/groups/page.tsx`, `app/page.tsx`, `app/organization/branding/page.tsx`, `app/groups/[id]/page.tsx`, `components/account/account-sections.tsx` ×3, `connection-manager.tsx` ×3, `layout/context-indicator.tsx` ×2, `layout/app-shell.tsx`, `signup/landing.tsx`; the 17th is `shell-help-dialog.tsx` `DialogTrigger asChild` → `render`; `account/pending-submit.tsx` omits the `asChild` type) |
| `dialog.tsx` (`DialogPrimitive.*`, portal adapter) + `portal-container.tsx` | `kit/src/components/ui/dialog.tsx` (`@base-ui/react/dialog`: Root/Trigger/Portal/Backdrop/Popup/Title/Description/Close) | 2 (+2): `layout/shell-help-dialog.tsx` (`DialogTrigger asChild`), `tests/fixtures/token-scopes/page.tsx` (`defaultOpen`, `DialogClose asChild`), `account/connection-manager.tsx` (Button only) |
| `input.tsx` / `textarea.tsx` / `select.tsx` | `input.tsx` (`@base-ui/react/input`) / `textarea.tsx` (native) / `native-select.tsx` (native, chevron icon) | 6 / 1 / 4 (`<Input` 7, `<Textarea` 1, `<Select` 7, mostly via `onboarding/form-field.tsx`) |
| `table.tsx` | `table.tsx` (native; our scroll region + `label` kept) | 4 |
| `typography.tsx` / `layout.tsx` / `bullet-list.tsx` / `site-header.tsx` | `components/vestive/typography.tsx`, `layout.tsx`, `content.tsx` (BulletList/BulletItem only), `site-header.tsx` (props kept: `brand`, `links`, `navLabel`, `action`; no logo/defaults) | 49 / 29 / 3 / 1 |

54 files import from `components/ui`; **33 files use `font-display|font-label|font-sans|font-mono` directly** (69
lines; list in the r0 inventory: `app/business-profiles/[id]/page.tsx`, `app/groups/new/page.tsx`,
`app/groups/[id]/onboarding/page.tsx`, `components/account/{account-sections,connection-manager}.tsx`,
`components/branding/{branding-form,logo-forms}.tsx`, `components/enrollment/{join-form,policy-form}.tsx`,
`components/labels/{group-labels-section,label-form,label-list}.tsx`, `components/layout/{app-shell,appearance-control,brand-mark}.tsx`,
`components/membership/{accept-form,group-authority-form,invitation-form,member-table,ownership-forms}.tsx`,
`components/onboarding/{address-check,form-field,state-badge}.tsx`, the 11 `ui` modules, `globals.css`).
Tests: `tokens.spec.ts` (`VARIANTS` incl. `muted`, `FILLED`/`ANGLED`, the `asChild` block 807–831, S-PORTAL-INIT),
`overlay.spec.ts` (`data-radix-focus-guard`), `design.spec.ts` (S-DESIGN zero-radius/2px/Oswald/`FONT_NODES`,
S-FIDELITY, S-FONTS "four", S-ANGLE), `support/{source-reference,modes,assets,fonts,spacing}.ts`; `data-state="pending"`
in mcp/membership/onboarding specs is the app's own attribute (unaffected).

### 3.2 Active configuration and dependency entries

`web/package.json` (2 removals, 1 addition), `web/package-lock.json` (24 removed entries — the closure minus `tslib`
— plus the Base UI closure), `web/components.json` (`style` → the Rhea/Base UI value from A3; aliases unchanged),
`web/app/globals.css` lines 1–640 **unchanged** (r1: the kit's `shadcn/tailwind.css` is already this block; only the foundation sections after line 640 change), `web/lib/utils.ts` unchanged.

### 3.3 Live references to the old skill and to Radix (outside closed history)

| Where | Lines | Treatment |
|---|---|---|
| `AGENTS.md` 70, 73; `CLAUDE.md` 524 | skill path | retarget (last commit) |
| `docs/ARCHITECTURE.md` §61 (5446) | old skill sentence | G067 successor record with an immutable `0e957c69:.agents/skills/vestive-design` citation |
| `docs/foundation/OPERATIONS.md` 813–814, 818, 820 | skills table / personal copies / historical names | retarget; old path appended to "Historical names" |
| `docs/foundation/WEB-SHELL.md` 1, 108, "Components and provenance" (incl. 309 deps, 326 "keep it on Radix … Base UI not allowed"), "Browser tests" | design-language, lineage, deps, adding-a-component | G067 sections; historical G002/G047/G048 tables kept with the commit citation; the component recipe flips to Base UI |
| `web/eslint.config.mjs` 8–9 | the two exact G049 lines | replaced by the G067 pair (exactly two lines, each once) |
| `web/components/layout/shell-help-dialog.tsx` 21 comment, `components/ui/portal-container.tsx` comment | "Radix" wording | reworded (no Radix mention remains in web outside lineage) |
| `infra/validation/run.sh` 978–983, 1319, ~1193, 1265, 1641; **1140–1153 (B-WEB-COMPONENTS import rules), 1174 PASS text, 1311 (A12 fixture), 879 (B-DEPS PASS text), 912–940 (C5, C13–C18)** | skill paths; Radix/Base UI rules | retarget / flip (§6) |
| `infra/validation/lib.sh` 1720 (`DEPS_ADDITIONS`), 1725 (`DEPS_REMOVED_PREFIXES`), 1733 comment, 1813–1828 (`UI_FILES`, `UI_SLOTS`, `WEB_PUBLIC_PINS`, `WEB_CSS_URLS`) | pins | re-pin (§6); `GLOBALS_EJECT_SHA` (run.sh 1050) **unchanged** (r1) |
| `infra/ci/lib.sh` 319; `infra/ci/profile-paths.tsv` 47–48; `infra/ci/validate-controls.sh` 5505/5511/5513/5688–5691 | skill paths | retarget |
| `web/THIRD_PARTY_NOTICES.md` | Radix section, fonts table | Base UI (MIT) section, Geist rows, Roboto row removed |
| Codex-owned, not edited here | `management/SUPERVISOR-HANDOFF.md`, `PROGRESS.md` (historical), `specs/OS001-foundation-public-export-spec.md:136,485` (names the old skill as an OS001 input) | reported to Codex; 29 Results and 54 Goals/Plans/Specs keep historical paths |

## 4. Intake, dependencies, configuration (R1, R2)

### 4.1 Skill intake and retirement

Canonical `.agents/skills/vestive-brand/` (tracked, real directory), alias `.claude/skills/vestive-brand`
(relative symlink). Intake integrity is proven in **two steps, both of which must pass** (owner B; an intentionally
failing manifest is never acceptance evidence):

- **V1 — upstream, before the local edit.** Re-run the Spec digest on `<home>/.agents/skills/vestive-brand`
  (319 files, `SKILL.md` `188ab915…`, tree `67f92e83…`; a changed byte is reported, not rebound), copy, write
  `UPSTREAM.sha256` (the 319 upstream rows; its own sha256 = the tree digest) and run `sha256sum -c UPSTREAM.sha256`
  inside the pristine tracked copy → **319 OK, 0 FAILED**, recorded in the Result intake section (command, exit 0,
  counts) and committed as B's first commit (`UPSTREAM.sha256` never changes afterwards).
- **Local delta.** Append the 4-line "Repository adoption" pointer to `SKILL.md` (the only upstream file touched), add
  `ADOPTION.md` and `LOCAL.sha256` = the 318 unchanged upstream rows + rows for the final `SKILL.md`, `ADOPTION.md` and
  `UPSTREAM.sha256`.
- **V2 — final tree.** `sha256sum -c LOCAL.sha256` → **321 OK, 0 FAILED**; `diff <(grep -v ' SKILL.md$'
  UPSTREAM.sha256) <(grep -v -E ' (SKILL\.md|ADOPTION\.md|UPSTREAM\.sha256)$' LOCAL.sha256)` empty (every upstream byte
  except `SKILL.md` unchanged); `head -n -4 SKILL.md | sha256sum` = `188ab915…` (the upstream `SKILL.md` is the final
  file minus exactly the appended tail). Recorded in the Result AC1 evidence and in `ADOPTION.md`; hosted
  B-SKILL-UNCHANGED/B-PROTECTED at `--base HEAD` cover later drift.

`ADOPTION.md` records: the owner-reported origin (Claude Design; Design System artifact v0.2.1
`BnvjycYZQwMXFQ4tzGsTAN` as the upstream `SKILL.md` cites), intake date, the three digests of §0, the local delta and
both verification commands with their recorded outcomes, the current copy/adaptation authority and lineage (below),
the statement that the tracked copy is the source of truth, and the §5.4 carry-forward rules.

**Copy/adaptation authority and lineage, recorded accurately (no new ownership declaration is requested).**
The authority to copy the supplied v0.2.1 skill into the repository and to adopt/adapt its components is the owner's
explicit G067 instruction (Goal "Scope"/"Owner scope amendment — 2026-10-25": "migrate and use the new Vestive brand
as is plus any additional useful carry over"; Spec r2 "Authority" and R1/R2), together with the owner's stated origin
(Goal: "the new skill was created from Claude Design"; Spec R1: document "owner-reported Claude Design origin").
The G048 selected-source rights record (`management/v0/results/G048-source-rights.md`, 2026-10-07) confirmed company
ownership and permission to copy/modify **six modules of the earlier kit** and is cited only as the historical record
for those modules; it is not extended to v0.2.1 and no such extension is asked of Codex. Third-party lineage, which
the Claude Design origin does not replace (Spec R3): the adopted `button`/`input`/`dialog` are the shadcn/ui registry
`base-rhea` output coloured with the Vestive tokens (§0.1; shadcn MIT notice already shipped, "restyled kit" wording
updated), `@base-ui/react` and its closure are MIT (new notice section), Geist/Geist Mono are OFL 1.1 with the
upstream copyright line (new font rows; Roboto row removed), Lucide unchanged. The v0.2.1 tree still carries no
licence/notice file of its own; its brand images (`kit/public/brand`, `design-system/assets`) are copied into the
skill directory as reference input and are never shipped by `web/` (S10), so no shipped-asset rights question arises.
What remains genuinely unconfirmed is listed once under E2 for Codex's disposition, not as a request to the owner.

Retirement (last commit): `git rm -r .agents/skills/vestive-design .claude/skills/vestive-design`; the old
`references/shadcn-kit.md` knowledge that still matters lives in `ADOPTION.md` (§5.4), not as a second live skill.

### 4.2 Dependency delta and the zero-Radix proof

| Package | Old | New | Notes |
|---|---|---|---|
| `@radix-ui/react-dialog` | 1.1.23 | **removed** | with its 15 scoped + 8 non-scoped closure entries (§0) |
| `@radix-ui/react-slot` | 1.3.3 | **removed** | `render` replaces Slot |
| `@base-ui/react` | — | **1.9.0** (`--save-exact`; integrity `sha512-S8CSYosASBKtpL8SWUfS78rFElfLijiCgA7oHOK22jCJ95FULsYyKwBNFOgudt0RA+/7yNc7F81bOvQnRgbBiQ==`) | the only direct addition; expected transitive entries (§0.1): `@base-ui/utils` 0.5.0, `reselect` 5.3.0, `@babel/runtime` 8.0.8, `@floating-ui/utils` 0.2.12, `@floating-ui/react-dom` 2.1.9, `@floating-ui/dom` 1.8.0, `@floating-ui/core` 1.8.0, `use-sync-external-store` 1.7.0 — all MIT; the actual lockfile closure is recorded in the Result and WEB-SHELL.md |
| `class-variance-authority`, `cn` 0.4.0, `lucide-react`, `tw-animate-css`, Next/React/TS/ESLint/Tailwind/Playwright/axe pins | unchanged | unchanged | `tailwind-merge`/`clsx` not added: the registry's own `base-rhea` modules import `cn` from `"cn"`, so keeping `cn` is the source-faithful choice, not a pin-preservation one; C19 unchanged |
| `shadcn`, `radix-ui`, `next-themes`, per-component catalogue | never | never | B-DEPS forbids `shadcn`/`radix-ui`; the kit's `shadcn/tailwind.css` import is satisfied by the existing vendored block (§4.3); nothing else requested |

**Proof of zero resolved Radix (executable, each recorded in the Result):**
1. `grep -c -E '@radix-ui|"radix-ui"' web/package.json web/package-lock.json` → `0` and `0`.
2. B-DEPS with `DEPS_REMOVED_PREFIXES=("@radix-ui/")` and `DEPS_FORBIDDEN=(shadcn radix-ui)`: `deps_pair_check` already
   scans **every lockfile `packages` key at any depth** for a removed prefix (`DEPS_REMOVED_PRESENT`), so a transitive
   Radix package fails it; C13/C14 are re-targeted to plant `@radix-ui/react-slot` (direct) and a nested
   `node_modules/@radix-ui/primitive` entry (transitive) and must fail with exactly those codes.
3. B-WEB-COMPONENTS: `code_lines "[\"'\`]@radix-ui/"` over every web source file **including tests** → any hit is
   `UI_IMPORT_RADIX <file>`; `@base-ui/react/<subpath>` imports allowed only in the reviewed module set (`button.tsx`:
   `button`; `dialog.tsx`: `dialog`; `input.tsx`: `input`), any other Base UI import `UI_IMPORT_BASEUI_UNREVIEWED`.
4. In the `web-check` container (P0 and the final run): `npm ls --all 2>/dev/null | grep -ci radix` → `0`
   (evidence line, not a new harness check).
5. `grep -rniE 'radix' web --exclude-dir=node_modules --exclude-dir=.next` → only lineage/history lines allowed (the
   Result lists each hit; the target is none in source, config or tests). One known CSS hit is declared in advance:
   the vendored registry block (`globals.css:11,20`) names the custom property `--radix-accordion-content-height` as a
   fallback inside shadcn's own `tailwind.css` (the same bytes at 4.21.1 and 4.22.1, library-neutral, emitted for
   every base). It is not an import, package, wrapper or active primitive configuration (Spec R2's four conditions)
   and the pinned block is kept verbatim; the Result records it as such.

### 4.3 shadcn configuration and the ejected registry block

`components.json` keeps `rsc`, `tsx`, `app/globals.css`, CSS variables, `lucide`, the five aliases, `menuColor`/
`menuAccent`; `style` → **`"base-rhea"`**, `baseColor` → **`"taupe"`** (§0.1). **Smallest source-faithful CSS
integration (review item 6):** the kit's `globals.css` is `tailwindcss` + `tw-animate-css` + `shadcn/tailwind.css` +
`theme.css` + `vestive.css`. The first three are already the repository's `globals.css` lines 1–640 byte-for-byte
(the vendored `shadcn@4.21.1` file equals `shadcn@4.22.1`'s), so **nothing is ejected, re-ejected or re-pinned**:
`GLOBALS_EJECT_SHA` and the marker stay; the custom variants (`data-open`, `data-closed`, `data-checked`,
`data-disabled`, …) Base UI needs are in that block. `theme.css` is consumed as **values** re-sourced into the existing
`--light-*`/`--dark-*` contract sections (A palette sources, B resolution, `@theme inline` colour/radius/font
mappings, `@theme` stacks, F faces, E blueprint, D appearance icons, the global reduced-motion rule) because the
override contract and `data-appearance` switching are retained guarantees (Spec R4). `vestive.css` is **not** imported:
its `@theme inline` radius multiples are already identical in `globals.css`; its `--font-sans`/`--font-mono` stacks
are adopted into the existing `@theme` names (`--font-heading`/`--font-label` map to the app's `--font-display`/
`--font-label`, a name alias only); its `@custom-variant dark` on class/`data-theme`, `body`/`h1–h4`/`::selection` base
rules and brand colour utilities are translated (the `dark` variant stays the `data-appearance` one; headings get
their family from the components; `::selection` derives from `accent` per appearance); `ribbon-wave.css`/
`rhea-leftovers.css` have no consumer. This replaces r0's planned eject container run and keeps no custom mechanism
for the sake of an old pin: the pins that no longer describe the source (`DEPS_REMOVED_PREFIXES`, the "never Base
UI" import rule, `radix-nova`, the old `UI_SLOTS`, the zero-radius/2px oracles) are flipped in §6, not worked around.

## 5. Source mapping (R2) and carry-forward (R3)

### 5.1 Palette (typed from `kit/src/styles/theme.css`; oklch converted; WCAG 2.x ratios computed from hex)

| Token | Old L / D | New L / D | Source slot / derivation | Verified ratio |
|---|---|---|---|---|
| background / foreground | `#f4f2ee`,`#1f1d1b` / `#0d0d0d`,`#e8e6e3` | `#ffffff`,`#0c0a09` / `#0c0a09`,`#fbfaf9` | Taupe `background`/`foreground` | 19.8 / 19.0 |
| muted | `#ebe8e3` / `#101010` | `#f3f1f1` / `#2b2422` | Taupe `muted` (band, field fill, hover fill; also receives the kit's `accent` slot role) | — |
| muted-foreground | `#615b55` / `#8f8a84` | **`#6f615b`** / `#aba09c` | kit L `#7c6d67` = 4.95 on white but **4.40 on the muted band** → derived 5.94 / 5.28; D kept 7.75 / 6.90 (card) / 5.98 (muted) | ≥ 4.5 |
| popover / -foreground | `#ffffff`,`#1f1d1b` / `#131313`,`#e8e6e3` | `#ffffff`,`#0c0a09` / `#1d1816`,`#fbfaf9` | Taupe `popover` (= `card`) | 19.8 / 17.6 |
| primary / -foreground | `#0062d6`,`#ffffff` | **unchanged** | kit `#0075ff` + white = **4.21** (S6); `#0062d6` non-text 5.65 / 3.50 on grounds | 5.65 |
| secondary / -foreground | green derivations | `#f3f1f1`,`#1d1816` / `#2b2422`,`#fbfaf9` | kit neutral `secondary` (green = CtaButton brand moment only) | 15.6 / 14.6 |
| accent | gold | **unchanged** `#8a5a00` / `#f8ce8a` | `vx-gold` ink; the kit's neutral `accent` slot is translated to `muted` (S4) | 5.93 (muted 5.27) / 13.4 |
| ring | gold | unchanged | kit `ring` = `vx-gold` | ≥ 3 |
| border | `#d4cfc8` / `#262626` | `#e8e4e3` / `#242222` (white 10 % over the ground, opaque) | Taupe `border`; decorative hairlines | 1.26 / 1.25 — never a control's only boundary |
| input | `#615b55` / `#8f8a84` | **`#908279`** / **`#7d726d`** | kit fields `bg-input/50 border-transparent` = **1.12:1** at rest → 3:1 boundary role kept: 3.71 (white) 3.30 (muted) / 4.23, 3.77 (card), 3.27 (muted) | ≥ 3 |
| destructive | `#af3028` / `#ff6b5e` | `#e7000b` / `#ff6467` | kit values; white label 4.77 (L), `background` label 6.84 (D); the kit's tinted variant = **4.01** in Light → solid fill | ≥ 4.5 |
| overlay | `rgb(13 13 13/.45)` / `rgb(0 0 0/.72)` | unchanged | kit `bg-black/30` theme-fixed scrim not imported | n/a |
| ink / text-2 / text-3 | old roles | `#0c0a09`,`#3d3b3a`,`#6f615b` / `#fbfaf9`,`#cbcac9`,`#aba09c` | kit typography uses `foreground`, `foreground/80` (opaque equivalents), `muted-foreground`; names unchanged | 11.1–19.8 / 7.8–19.0 |

Not mapped (no app consumer): `card`, `accent-foreground`, `sidebar-*`, `chart-1…5`, `vx-lime/cyan/umber/star`, ribbon
ramp. Kit utilities that would fail B-WEB-TOKENS are translated at adoption: `text-white` → `text-primary-foreground`,
`bg-black/30` → `bg-overlay`, `text-gold/orange/blue/green` → `text-accent`/`text-primary`; no colour literal outside the
palette block; `bg-accent`/`accent-foreground` stay forbidden.

### 5.2 Type, geometry and primitive API mapping

| Aspect | Old (Radix/G048) | New (kit source, Base UI) |
|---|---|---|
| `--font-sans` / `--font-mono` | Roboto / ui-monospace | `"Geist", ui-sans-serif, system-ui, sans-serif` / `"Geist Mono", ui-monospace, "SFMono-Regular", Menlo, monospace` (bundled) |
| `--font-display`/`--font-label` | Oswald/Barlow everywhere | stacks unchanged; use narrowed to brand moments: header identity (`app-shell.tsx`, `brand-mark.tsx`), landing H1 + kicker (`page.tsx`), sign-in H1 (D-2) |
| Everyday headings / Kicker | Oswald uppercase / Barlow gold | Geist `font-semibold tracking-tight` scale (h1 `text-2xl`, h2 `text-xl`, h3 `text-base`), sentence case; Kicker Geist `text-xs font-semibold uppercase tracking-[0.08em] text-accent`; `brand` boolean keeps the Oswald/Barlow display styles |
| Button | `React.forwardRef` + Slot `asChild`; Oswald uppercase angled `::before` fill; 42px `sm` | kit `Button` on `ButtonPrimitive` (`@base-ui/react/button`); links use **`ButtonLink`** (an `<a>` with `buttonVariants()` and `data-slot="button"`, per Base UI's Button docs — `render` is used only where a button composes a button, e.g. `DialogTrigger render={<Button/>}`); Geist `text-sm font-medium`, `rounded-2xl` (**18px**), **`h-7` 28px `sm` / `h-8` 32px default / `h-9` 36px `lg`** / `xs` 24px / icon 32px, `icon-sm` 28px, `icon-xs` 24px, `icon-lg` 36px (measured, §0.1); variants `default`, `outline`, `secondary` (neutral), `ghost`, `destructive` (solid, D-7), `link`, **`cta`** (angled `::before` fill + Oswald label, D-3); `muted` removed; **intentional adaptation:** the shell focus outline (solid 2px `ring`, offset 2) replaces the kit's `focus-visible:ring-ring/30` because a 30 %-alpha gold ring (`#8a5a00` at 0.3 over white ≈ `#dcceb3`, ≈ 1.6:1) fails the 3:1 non-text focus-indicator requirement (Spec R4 "visible unclipped focus"; WCAG 1.4.11); `pending-submit.tsx` omits `type`/`children` only |
| Dialog | `DialogPrimitive.Root/Trigger/Portal/Overlay/Content/Title/Description/Close`; scope adapter on Radix Portal `container` | kit `Dialog` on `@base-ui/react/dialog`: `Root` (`modal` default `true`), `Trigger` (`render`), `Portal` (**scope adapter rebuilt**: explicit `container` wins; inside a scope render nothing until `scope.element`, then `container={scope.element}`; outside a scope Base UI's default), `Backdrop` (`data-slot="dialog-overlay"`, `bg-overlay` for the kit's theme-fixed `bg-black/30`, no blur D-4), `Popup` (`data-slot="dialog-content"`, `rounded-[min(var(--radius-4xl),24px)]` = **24px**, `p-6`, `gap-6`, `ring-1 ring-foreground/5 dark:ring-foreground/10 shadow-xl sm:max-w-md`, `max-h`/`overflow-y-auto`, `pr-9` header, no open/close motion D-6), `Title` (Geist 16px/500 `leading-none`), `Description` (14px `text-muted-foreground`), `Close` (`render={<Button variant="ghost" size="icon-sm"/>}` = **28px** with lucide `XIcon` + `sr-only` "Close"); `aria-modal`/`modal` behaviour proven by S-OVERLAY (A2) |
| Fields | 2px `border-input`, 48px, Roboto | kit `Input` on `@base-ui/react/input` (native `<input>` rendering), `Textarea` (`min-h-16` 64px), `NativeSelect` (chevron icon, wrapper slot; `sm` 28px); `rounded-2xl` (**18px**), **`h-8` 32px**, `px-2.5`, Geist `text-base md:text-sm` (16px, 14px from 768px — kept: it is the kit's mobile zoom guard), shell focus outline (same reason as Button), `aria-invalid:border-destructive`; **intentional adaptation:** the kit's `bg-input/50 border-transparent` resting boundary measures 1.12:1 (§5.1), so the 1px `border-input` derived boundary (≥ 3:1) on the field fill is kept (Spec R3 "resting field identification"); `form-field.tsx` keeps native `<label>`/`aria-describedby` (kit `Field` not adopted) |
| Table | 2px rules, Barlow heads | kit `Table`: 1px hairline rows, Geist `text-sm`, heads **`h-10` 40px** `px-2 font-medium text-foreground`, `p-2` cells; our scroll region + `label` prop and wrapping kept; `whitespace-nowrap` and hover/selected states dropped (reflow, Spec R3) |
| Rules/sections/grid | 2px | 1px hairline; `RuledGrid` reflow below 768px kept (kit `minWidth 860 + overflow-x-auto` not adopted) |
| Radius | `0` | base `--radius: 0.625rem` with the already-pinned multiples; **rendered: controls and fields 18px (`rounded-2xl`), dialog 24px**; the appearance control follows the fields (18px) |
| Bullets / links / header / appearance control / blueprint | — | bullets unchanged (14×2 gold rule, Geist text); `TextLink` Geist `text-accent` hairline underline; header hairline, Geist nav, Oswald identity; appearance control `rounded-2xl border border-input`, Geist options; blueprint backdrop kept (D-5) |
| State selectors | Radix `data-state=open`, `data-radix-focus-guard` | Base UI `data-open`/`data-closed` (already in the custom-variant block), FloatingUI focus guards (`data-floating-ui-focus-guard`; the exact attribute is read from the installed package in P0 and cited in `overlay.spec.ts`) |

Representative comparisons (before = G048/preview screenshots in evidence; after = S-SCREENSHOTS and the two extended
captures of §7.2): root shell + help dialog, `/sign-in`, one members table and one form page (`/groups/new`), each at
1280 Light/Dark and 320 px.

### 5.3 Decisions with stated defaults (Codex may redirect; none blocks)

- **D-2** brand-moment type: header identity, landing H1/kicker and sign-in H1 only. Alternative: Barlow kickers everywhere.
- **D-3** angle: `--angle` stays (B-WEB-TOKENS names it); a `cta` variant uses it on at most one action per view
  (landing "Sign up", sign-in primary), fill on `::before` so label and focus are never clipped (the kit's own
  `CtaButton` clips the element and therefore its outline). Alternative: no angled control.
- **D-4** no backdrop blur. **D-5** blueprint backdrop kept. **D-6** no dialog open/close motion (S-MOTION unchanged).
- **D-7** destructive = solid fill (the kit's tint fails 4.5:1 in Light).

### 5.4 Carry-forward table: old skill guidance → disposition

| Old guidance (`SKILL.md` / `README.md` / `references/shadcn-kit.md`) | Disposition | Reason |
|---|---|---|
| Rule 1: product requirements, ARCHITECTURE/ADRs/Goals win over the skill | carry forward (`ADOPTION.md` §1) | new `SKILL.md` §7 says the same; made repository-specific |
| Rule 2: marketing copy, voice, pricing, trademarks, nav labels are reference only; identity from configuration | carry forward | new §2 "Voice" is still marketing; literal guard unchanged |
| Rule 3: Light/Dark/System, System default; "Dark is the brand default", `next-themes`, `.dark`/`[data-theme]` are source docs | carry forward | kit README steps 4–5 and `theme.css` selectors still default to Dark; translated to `data-appearance` |
| Rule 4: source stack is not permission to change the app; packages need review, major ones an ADR | carry forward, **now exercised**: the primitive change is reviewed through the ADR-001 amendment and this Plan; the catalogue packages are still not installed | ADR-001 amendment |
| Rule 5: keyboard, visible focus, modality, contrast, reduced motion, reflow, semantic tokens, scope isolation incl. portals, pre-paint/hydration | carry forward (Spec R4) | guarantees, not styling |
| Rule 6: logos/icons/marks are Vestive identity, not product identity | carry forward | v0.2 icon set likewise |
| Rule 7: preserved sources read-only; `sha256sum -c`; copy out and adapt | carry forward as `UPSTREAM.sha256` | same mechanism |
| §8 map by role, not slot name (kit `accent` ≠ app `accent`) | carry forward as the §5.1 translation table | Spec R3 names it |
| §8 translate primitive specifics (Radix ↔ Base UI attributes, `asChild`/Slot ↔ `render`) | carry forward, direction reversed | now the app adopts Base UI: link `asChild` → styled anchor (`ButtonLink`), button-composing `asChild` → `render`, `data-state` → `data-open`, `Overlay/Content` → `Backdrop/Popup` |
| `shadcn-kit.md` class-merge semantics (`cn`, tailwind-merge; G048 X23 `leading-none` repeat) | carry forward the fact, **drop** X23 | the registry's `base-rhea` output itself imports `cn` from `"cn"`; the kit's `lib/utils.ts` is its local clsx + tailwind-merge variant with the same conflict-resolution semantics; X23 belonged to the retired angled everyday button (sizes no longer set `text-[…]` after `leading-none`) |
| G002 ejected registry CSS pinned verbatim (`GLOBALS_EJECT_SHA`) | carry forward **unchanged** | the kit's `shadcn/tailwind.css` is byte-identical to the pinned block (§0.1); no eject run |
| §9.6 portals default to `document.body`, leaving the token scope | carry forward | Base UI `Dialog.Portal` does the same; adapter rebuilt on its `container` prop; S-PORTAL-INIT stays |
| §9.7 clipped focus on angled buttons | carry forward (D-3) | verified in the new `cta-button.tsx` |
| §9.8 white on blue/green 4.21/4.05 at 14–15px is below 4.5 | carry forward, explicit handling §5.1 | the new brand book repeats "kept exact … keep labels at their set size and weight", contradicting its own "body text ≥ 4.5:1" checklist; recorded for the owner (Q2) |
| §9.9 resting field boundary | carry forward, re-measured (1.12:1 now) | derived `input` boundary |
| §9.10 reflow (`RuledGrid` 860px, `whitespace-nowrap`, fixed `NodeField`) | carry forward | new `layout.tsx`/`table.tsx` repeat them |
| §9.11 reduced motion covers ribbon/node/`[data-motion]` only | carry forward | `tw-animate-css` dialog motion and `transition-all` not covered → global rule stays |
| §9.12 theme-fixed scrim/selection values | carry forward | per-appearance tokens |
| §9.5 no wholesale global stylesheet/reset | carry forward | `vestive.css`/`globals.css` of the kit not imported; registry CSS ejected and pinned instead |
| §9.3/§7 remote Google Fonts, no font files | carry forward | bundled OFL files |
| §10 unresolved provenance (no licence text, icons8 terms) | carry forward, updated | still no licence file; MDI paths gone; icons8 glyphs not shipped |
| T11/X16 static header, T12 uppercase via CSS, X17 nav target size, X18 grid reflow, X19 `overflow-wrap:anywhere` for identity, X12/X13 dialog reflow, X14 client dialog module | carry forward | repository rules independent of styling |
| G002 T4/T8/T10, G047 K1–K4/K7/K8/K10, G048 X1–X4/X6–X9/X11/X20 (derived fills, boundary role, outline focus, overlay token) | carry forward (values re-derived §5.1) | accessibility corrections the new source does not make either |
| Old palette values, 2px rules, zero radius, "no shadows/blur/glows", Roboto body, Oswald on every heading/button/nav, Barlow labels/table heads | **supersede** | Spec R3: old palette/typography/geometry are superseded |
| **Radix composition (ADR-001 original), the `asChild`/Slot API, Radix state selectors, the "no `@base-ui` import" rule, `radix-nova` style, the G048 "keep it on Radix" recipe** | **supersede** (ADR-001 amendment) | Base UI/Rhea is the successor foundation; G048's Radix migration stays historically correct |
| MDI icon guidance, `icons.tsx`/`menu-styles.ts`, `vx-faint`, ribbon/node motion, orange `Emph`, `PlanCard`/`Rating`/`Quote`/`Stat`, 1360/88/36 structure as mandatory, `styles.css`/`motion.js`/examples | **drop** (Git history keeps them) | marketing/brand-moment material; lucide is the icon library |

`ADOPTION.md` holds the "carry forward" rows as repository rules (reusable accessibility cautions in a separate
section), the §5.1 derivations, the lineage table (kit path → `components/ui` file → sha256), the dependency delta
and the open owner questions. It is linked from the appended `SKILL.md` tail.

## 6. Guard adaptation (source-path, fixture, pin and rule flips only; no new runner)

- **Skill paths:** `run.sh` 978 (`VESTIVE_SKILL`/`VESTIVE_ALIAS`), 981–983 (the two exact ESLint lines → the G067 pair),
  1319 (A16 glob), `web_design_check` byte-copy root → `kit/public/brand` and `design-system/assets`, 1641
  (`SKILL_LINK`/`_TARGET`); `infra/ci/lib.sh` 319; `profile-paths.tsv` 47–48; `validate-controls.sh` 5505/5511/5513/5688–5691
  (control IDs kept; `-kit` → `kit/src/components/ui/button.tsx`, `-css` → `kit/src/styles/theme.css`). `SKILL_MOVES` and
  the 1680 reversal clause stay as inert history (hosted/`validate.sh` runs use `--base HEAD`).
- **Dependency policy (`lib.sh` 1720–1727):** `DEPS_ADDITIONS` loses the two Radix rows and gains
  `dependencies|@base-ui/react|1.9.0`; `DEPS_REMOVED_PREFIXES=("@radix-ui/")`; `DEPS_FORBIDDEN=(shadcn radix-ui)`;
  comment 1733 reworded (the `devOptional` note now names `@base-ui/react`'s optional `@types/react` peer). `run.sh`
  C5/C16/C17/C18 re-targeted to the `@base-ui/react` 1.9.0 pin (wrong range / wrong version / recategorized / lock
  mismatch), C13 → direct `@radix-ui/react-slot` (`DEPS_EXTRA_DIRECT` + `DEPS_REMOVED_PRESENT`), C14 → nested
  `node_modules/@radix-ui/primitive` (`DEPS_REMOVED_PRESENT`); C10/C15/C19 unchanged; B-DEPS PASS text updated.
- **Component rules (`run.sh` 1130–1176):** `components.json` expectation → `style` `base-rhea`, `baseColor` `taupe`;
  `UI_FILES` unchanged (same 11 names); `UI_SLOTS` → the new source slot names actually emitted (recorded when final:
  e.g. `native-select`, `native-select-wrapper`, `native-select-icon`, `table-container`, `table-header`, `table-body`,
  `table-row`, `table-head`, `table-cell`, `table-caption`, `dialog-close`, `cta-button`; `ButtonLink` emits
  `data-slot="button"`); **eject marker and `GLOBALS_EJECT_SHA` unchanged**; `UI_BUTTON_SLOT` → `UI_BUTTON_BASEUI`
  (`import { Button as ButtonPrimitive } from "@base-ui/react/button"`), `UI_DIALOG_RADIX` → `UI_DIALOG_BASEUI`
  (`import { Dialog as DialogPrimitive } from "@base-ui/react/dialog"`), portal-adapter rule kept (`usePortalContainer`
  in `dialog.tsx`); import scan flipped (§4.2 item 3); X9 focus rule and X7 `--angle` rule kept; X23 rule dropped
  (§5.4); A12 fixture plants a `@radix-ui/react-dialog` import in a test and expects `UI_IMPORT_RADIX`.
- **Pins:** `WEB_PUBLIC_PINS` (10 entries: 8 current − 3 Roboto + 4 Geist/Geist Mono files + 1; exact count recorded
  when final), `WEB_CSS_URLS` (5 fonts + SVG), B-WEB-DESIGN `== 4` → `5` and PASS text; tests `assets.ts`, `fonts.ts`,
  `source-reference.ts`, `modes.ts`, `design.spec.ts` counts and geometry literals (0px → 18px/24px, 2px → 1px).
- **Unchanged:** B-WEB-TOKENS logic, B-WEB-BOUNDARY, `profile-paths` classes for `web/**`, `.gitlab-ci.yml`, `required-checks.tsv`.
- Any broader runner change found necessary in P0 gets its own recorded coverage rationale and supervisor disposition.

## 7. Validation (focused executable checks)

### 7.1 Selectors

`<E>` = `<source-checkout>/.local/evidence/g067-a1` (§7.3). Suite names verified
against `infra/validation/run.sh` (`suite_lint` 346, `suite_lint_web` 367, `suite_controls` 389, `suite_unit_web` 442,
`suite_rename` 5418) and control IDs against `infra/ci/validate-controls.sh` (152–154, 5679–5691).

| # | Selector (exact) | Checks | Reuse / extend / new | Runs in | Slot |
|---|---|---|---|---|---|
| 1 | `bash infra/validation/run.sh --base <staging SHA> --allow-file <G067 allow-file> --report-dir <E>/static-<n> lint lint-web controls` — **one invocation** (r0's `lint-web`-only row would end `B-SECRET-CANARY UNMET`, §0.1; `lint` supplies the generated secrets/canaries) | `lint`: B-SHELL on the edited scripts, B-HADOLINT, B-COMPOSE, **B-DEPS** (flipped policy, lockfile at any depth), B-SECRET-SOURCE, …; `lint-web`: B-WEB-BOUNDARY, B-WEB-TOKENS, B-WEB-COMPONENTS (flipped), B-WEB-DESIGN, B-WEB-ASSET-CONTROLS A1–A19; `controls`: **B-DEPS-CONTROLS** C1–C19 (incl. re-targeted C5/C13–C18), B-WRAPPER-CONTROLS; rollup B-SCOPE, **B-SECRET-CANARY PASS** | reuse (rules/pins updated) | host shell + tool image; no stack, no port | none |
| 2 | same, suite `unit-web` | B-UNIT-WEB = README `web-check` (`eslint --max-warnings=0`, `tsc --noEmit`, `node --test tests/unit/*`) + secret-scan (its own generated environment → canary met); plus the `npm ls --all` Radix count (evidence line) | reuse | Docker `web-check` image (build has registry access for `npm ci`; checks run without network) | build slot |
| 3 | private browser stack: `COMPOSE_PROJECT_NAME=g067br-<run>` `docker compose --env-file <main>/.local/secrets/g067-a1/browser-<run>.env -f infra/compose/local.yml -f <override: ports: !reset [] for postgres redis minio core web intelligence> up -d --build --wait`, then `… -f infra/compose/browser.yml --profile browser run --rm --build --no-deps -e PLAYWRIGHT_WORKERS=2 web-browser npx playwright test --project=shell-light --project=shell-dark --project=tokens-light --project=tokens-dark --project=appearance --project=dev-hydration` (P0 adds `--grep` to the `overlay|tokens|design|accessibility|appearance` specs) | overlay.spec (S-OVERLAY, S-MOTION), tokens.spec (scopes, S-PORTAL-INIT, inert, `render`/`ButtonLink` composition, S-STATES), design.spec (S-DESIGN, S-PALETTE, S-FIDELITY, S-FONTS ×4 + swap stability, S-BLUEPRINT, S-ANGLE, S-SCREENSHOTS), accessibility.spec (S-A11Y, S-CONTRAST ×2, S-TEXT-SPACING), shell.spec (S-IDENTITY … S-LONG-IDENTITY), appearance.spec (all incl. S-HYDRATION-DEV) | reuse; **extend** S-SCREENSHOTS (+`/sign-in` 1280 L/D, 320) | Docker, private project, `network_mode: service:web`, **no published port** (G064 driver pattern `results/G064-br-proof-driver.sh:265–268`) | one of the two run slots; build slot during image builds |
| 3r | **focused private renamed-identity UI proof**, appended to the final pre-MR run of row 3 on the same private project: set `PRODUCT_NAME="Renamed Example"`, `PLATFORM_BASE_DOMAIN=renamed.example.net`, `CUSTOM_DOMAIN_CNAME_TARGET=custom.renamed.example.net` in the private `.env`; record `docker compose … images -q` before; `docker compose … up -d --no-build --wait`; images unchanged after; `… --profile browser run --rm --no-deps -e PLAYWRIGHT_WORKERS=1 web-browser npx playwright test --grep @identity --project=shell-light --project=shell-dark` | **exactly `BROWSER_MATRIX_IDENTITY`** (`lib.sh:1879`: S-IDENTITY in shell-light and shell-dark — title and visible identity from configuration under the renamed name/domain), i.e. the browser half of B-BROWSER-RENAME, plus the image-ID-unchanged line; recorded as an evidence line, **not** as a B-BROWSER-RENAME/B-RENAME record. **Not covered and reported unexecuted:** B-RENAME host-port probes on `PORTS_REN`, README `platform:start` on a disposable copy, B-SOURCE-HASH, B-DBRENAME (database rename, roles, `migrate up`); G067 changes no `.env.example`, compose file, migration, core or database code (the allow-file excludes them), so their inputs are unchanged since their last attributable pass. Running the real `rename` suite would need the seven `PORTS_REN` loopback ports, a disposable repository copy and a second full stack — not reserved for this task | **new evidence line** (no harness change) | same private project, no published port | the row-3 slot |
| 4 | `bash infra/ci/validate-controls.sh --only F-vestive-css,F-vestive-kit,F-vestive-md,F-vestive-link,F-agent-claude-skill,F-agent-agents-skill --log <E>/f-controls.log` | profile selection on the retargeted map/fixture | reuse | host shell, fixture repos, no Docker | none |
| 5 | hosted MR `validate:checks` (fast: lint, lint-base, lint-web, lint-core, provenance incl. B-SKILL-UNCHANGED/B-PROTECTED at `--base HEAD`, contracts, unit) + Duo | automatic per push; `validate:integration` not requested | reuse | hosted | n/a |
| 6 | **post-merge, exact merged SHA, frozen checkout:** rows 1–4 and 3r only if the merged tree differs from the tested tree; plus `IDENTITY_E2E=1` private identity stack (`infra/compose/identity.yml`, no ports) `npx playwright test --project=identity-light --project=identity-dark --project=mcp-light --project=mcp-dark --project=onboarding-light --project=onboarding-dark --project=membership-light --project=membership-dark --project=enrollment-light --project=enrollment-dark --project=signup-light --project=signup-dark --project=labels-light --project=labels-dark` | **extend** one membership or labels scenario with a `testInfo.attach` screenshot (AC2 on an app surface) | Docker, private stack | one run slot, sequential builds |

Not selected, reported as unexecuted: `rename` (replaced by the attributable row 3r proof and the unchanged-input
statement above), `ci-controls`, `migrations`, `ops`, `off` (B-BROWSER-OFF repeats the same UI on the
intelligence-disabled variant; no disabled-variant branch in the UI change — a supervisor may add it with a
reason), the full profile. Allow-file (B-SCOPE) patterns: `^web/(app|components|lib|public|tests|package\.json|package-lock\.json|components\.json|THIRD_PARTY_NOTICES\.md|eslint\.config\.mjs)`,
`^\.agents/skills/vestive-(brand|design)/`, `^\.claude/skills/vestive-(brand|design)$`,
`^infra/(validation/(run|lib)\.sh|ci/(lib\.sh|profile-paths\.tsv|validate-controls\.sh))$`,
`^docs/(ARCHITECTURE\.md|foundation/(WEB-SHELL|OPERATIONS)\.md)$`, `^(AGENTS|CLAUDE)\.md$`, `^management/v0/(plans|results)/G067-.*`.

### 7.2 Test changes

- **Reuse unchanged:** appearance.spec, shell.spec, S-MOTION, tokens.spec scope/override/first-insertion logic,
  accessibility.spec logic, fonts oracle phases, unit tests (19 files; structure/text assertions only).
- **Extend (values/selectors re-typed, scenarios kept):** `overlay.spec.ts` focus-guard attribute (the installed
  Base UI value, read in P0), `tokens.spec.ts` 807–831 (`ButtonLink`: the anchor itself carries the variant classes
  and `data-slot="button"` — the same `["A", "button"]` assertion), `VARIANTS`/`FILLED`/`ANGLED` (−`muted`, +`cta`),
  `BUTTON_SIZES` (**28/32/36px**), `source-reference.ts` 32 rows (cite `kit/src/components/ui/<file>.tsx:<class
  string>` / `theme.css`), `design.spec.ts` S-DESIGN (**controls and fields `border-radius` 18px, dialog 24px, rules
  1px**, Geist families, `brand` kicker), S-FIDELITY (`h-8` 32px fields, `size-7` 28px dialog close), `FONT_NODES`,
  S-FONTS counts, S-ANGLE → `cta`, `modes.ts` sources, `assets.ts` pins, the token-scopes fixture (`ButtonLink`,
  `render`), S-SCREENSHOTS, one identity-stack screenshot attachment. Uncovered behaviour proven: Base UI
  composition/modality/portal-scope on the adopted version (AC4), successor geometry/type/roles rendered (AC2), five
  fonts delivered and rendered without synthesis (AC2), derived contrast pairs at rest/hover/focus (AC4), the `cta`
  angle clipping nothing (AC4).
- **New:** none (every guarantee already has an oracle).

### 7.3 Dependency graph, resource order, concurrency, storage, cleanup

**Dependency graph (→ = must precede; ∥ = independent, may overlap):**

```
B: I1 copy+alias+UPSTREAM.sha256+V1+secret-scan (A6) → I2 ADOPTION.md+LOCAL.sha256+SKILL.md tail+V2 → I3 docs retargets
   (ARCHITECTURE §61, OPERATIONS, WEB-SHELL G067 sections, AGENTS, CLAUDE) → I4 infra/ci retargets + row 4 → handback
A: P0.1 deps ∥ B.I1 → P0.2 config → P0.3 complete mechanical migration (+fonts, harness flips) → P0.4 rows 1–2
   → P0.5 row 3 restricted → F/V → C → T → rows 1–2 → row 3 (+3r on the final pre-MR run)
   → merge B's branch (disjoint paths) → G (run.sh skill paths, notices, deletions; last commit) → rows 1, 2, 4
   → safekeeping push → Codex source review → MR (row 5) → authorized merge → row 6 on the frozen merged SHA
```

True dependencies: P0.1 before P0.3 (types come from the installed package); P0.3 complete before P0.4 (the app must
typecheck); P0.5 before F/V/C (the proof gates the strategy); T after C (one re-typing pass against the final rendered
choices; the measured kit values of §0.1 are already known, so T is bounded); G after B's handback (the old skill is
deleted and `run.sh` paths flip only once the new skill is in the tree). B's I1–I4 depend on nothing in A; A's
lineage headers cite the tracked kit path + sha256, which equal the home-source bytes B verifies in V1 (A reads the
home source read-only until B.I1 lands). Reruns are driven by failures/affected checks only (no fixed iteration
counts): a failed row reruns after its fix together with the checks the fix affects; attributable passes are kept.

**Resources.** Builds are sequential (one build slot; `COMPOSE_PARALLEL_LIMIT=1`). At most one G067 run slot at a
time; no third run. The owner's preview holds project `m1-preview-20261024` with loopback ports 543, 3100, 8180, 4100,
5532, 6479, 8100, 9100, 9101; G067 runs publish **no port** and use distinct project names; its checkout, containers,
volumes and images are never stopped, rebuilt, rebased or reclaimed. **Task-private storage** is the main
integration checkout's ignored `.local/` (AGENTS.md "Private files for new tasks"; `.gitignore:14`):
`<main>/.local/evidence/g067-a1/` (harness `--report-dir`, browser reports, `inventory.tsv`, the compact receipt),
`<main>/.local/secrets/g067-a1/` (the private `.env` files, mode 0600 in a 0700 directory) and `<main>/.local/tmp/g067-a1/`
(scratch), where `<main>` = `<source-checkout>`; r0's home-state root is withdrawn.
**Disposable creators are pinned and accounted:** the P0.1 dependency-resolution container runs the repository's own
`web.Dockerfile` base `node:24-slim` (image ID and `RepoDigest` recorded at allocation, shared base image never
removed), as `docker run --rm --name g067-a1-npm-<run> --label g067.run=<run> -u <uid:gid> -e HOME=/tmp -v
<A worktree>/web:/app -v g067-a1-nm-<run>:/app/node_modules -w /app node:24-slim sh -c 'npm ci && npm uninstall … &&
npm install --save-exact @base-ui/react@1.9.0'`, the volume `g067-a1-nm-<run>` recorded at creation and removed
right after the lockfile is copied out; browser stacks use `COMPOSE_PROJECT_NAME=g067br-<run>`; every
container/image/volume/network/worktree/temp path is recorded at allocation in `<main>/.local/evidence/g067-a1/inventory.tsv`,
reconciled at each handback/batch end/failure, removed by exact ID; `docker ps -a --filter
label=com.docker.compose.project=g067br-<run>` empty after each browser run; compact receipt reports all categories.
`bash infra/ci/docker-desktop-mounts.sh` before each Docker batch; fetch/rebase/reconcile before every batch (exact-delta
rule for irrelevant staging changes); the frozen merged-SHA checkout is never rebased.

## 8. Files, sequence, delegation, estimates

### 8.1 Files

Create (B): `.agents/skills/vestive-brand/**` (319 + `ADOPTION.md`, `UPSTREAM.sha256`, `LOCAL.sha256`),
`.claude/skills/vestive-brand`. Create (A): `web/public/fonts/geist/{geist-wght.ttf,OFL.txt}`,
`web/public/fonts/geist-mono/{geist-mono-wght.ttf,OFL.txt}` (the §0.1 bytes), `ButtonLink` in `button.tsx`,
`management/v0/results/G067-vestive-brand-successor-result.md`.
Ownership: **B** writes only `.agents/skills/vestive-brand/**`, `.claude/skills/vestive-brand`, `docs/**`, `AGENTS.md`,
`CLAUDE.md`, `infra/ci/**`; **A** writes everything else below (incl. `infra/validation/{run,lib}.sh`,
`web/THIRD_PARTY_NOTICES.md`) and merges B's branch before G.
Modify: `web/package.json`, `web/package-lock.json`, `web/components.json`, `web/app/globals.css` (foundation sections only; eject block unchanged),
the 11 `web/components/ui/*`, `web/components/layout/{app-shell,appearance-control,brand-mark,context-indicator,shell-help-dialog}.tsx`,
the 11 `asChild` files and the other consumer files of §3.1 (class re-maps), `web/app/page.tsx`, `web/app/sign-in/page.tsx`,
`components/signup/landing.tsx`, `web/tests/browser/{design,tokens,overlay,accessibility}.spec.ts` + one identity-stack spec,
`web/tests/browser/support/{assets,fonts,modes,source-reference,spacing}.ts`, `web/tests/fixtures/token-scopes/page.tsx`,
`web/THIRD_PARTY_NOTICES.md`, `web/eslint.config.mjs`, `infra/validation/{run,lib}.sh`, `infra/ci/{lib.sh,profile-paths.tsv,validate-controls.sh}`,
`docs/ARCHITECTURE.md` (§61 record), `docs/foundation/{WEB-SHELL,OPERATIONS}.md`, `AGENTS.md`, `CLAUDE.md`, the new `SKILL.md` tail.
Delete (last commit): `.agents/skills/vestive-design/**` (81), `.claude/skills/vestive-design`, `web/public/fonts/roboto/**`.
Unchanged: `lib/tokens.ts` names, `lib/appearance.ts`, `app/layout.tsx`, `token-scope.tsx`, every route/handler,
`core/`, `intelligence/`, `.gitlab-ci.yml`, `required-checks.tsv`.

### 8.2 Sequence and estimates (coding / build-test elapsed / review-queue; unknowns marked), attempt boundaries

| Phase (owner) | Work | Coding | Build/test elapsed (serial on that owner's path) | Review/queue waits |
|---|---|---|---|---|
| I1–I4 (B, ∥ A) | copy, alias, V1, secret-scan, ADOPTION.md, LOCAL.sha256, V2, docs retargets, infra/ci retargets, row 4 | 3–4 h | secret-scan ≈ 1–2 min (stack-free); row 4 ≈ 1–2 min | — |
| P0 (A) | deps resolution + closure comparison (A1), `components.json` (A3), complete mechanical migration (3 primitives, `ButtonLink`, 16 link sites + 1 trigger + type + fixture + 2 specs), fonts (A4), harness flips, rows 1–2, row 3 restricted | 4–6 h | resolution container ≈ 3–8 min (**unmeasured**, registry-bound; the only networked step); row 1 ≈ 8–11 min (`lint` ≈ 6 min measured as `lint lint-base`; `lint-web` 1–3 min; `controls` **unmeasured** ≈ 2 min); row 2 ≈ 4–8 min (**unmeasured**; fast profile 9 min 10 s measured for seven suites); row 3 ≈ 8–15 min builds + 10–20 min tests (**unmeasured**; B-BROWSER-ROOT is 141 tests) ⇒ ≈ 35–65 min | — |
| F/V + C (A) | palette/type re-sourcing, remaining 8 `ui` modules, 33 consumer files' class re-maps, D-2/D-3 surfaces | 6–9 h | rows 1–2 ≈ 12–19 min per rerun; reruns only for failures (expect 2–3) ⇒ ≈ 25–60 min | — |
| T (A) | re-type 32 rows + specs from the measured values, fixture, screenshots | 5–8 h (A5) | row 3 ≈ 20–35 min per rerun; expect 2–3 ⇒ ≈ 40–105 min; final run + 3r ≈ 30–40 min | — |
| merge B + G (A) | merge B's branch, `run.sh` paths, notices, deletions, docs check, Result | 2–3 h | rows 1, 2, 4 ≈ 15 min | — |
| Review/MR | safekeeping pushes, Codex source review, MR, Duo triage | 1 h | hosted fast ≈ 10 min/push (measured); Duo 5–10 min | **Codex review rounds: unknown, not budgeted as a promise** |
| Post-merge | row 6: identity stack, 14 projects (155+ tests, **unmeasured**, est. 30–60 min) + delta reruns only if the merged tree differs; receipt | 1 h | 1–1.5 h | final ACCEPT: unknown |

**Totals.** Coding ≈ 19–31 h (A ≈ 17–27 h on the critical path, B ≈ 3–4 h in parallel). Serial validation elapsed on
A's path ≈ 3–5 h (sum of the expected runs above; unmeasured rows dominate the spread). Queue/review: hosted ≈ 10–20
min per push; Codex source-review and ACCEPT rounds are unknown and excluded. Realistic wall-clock for A's path to a
review-ready MR ≈ **3–4 working days**, excluding Codex waits; r0's 25–36 h figure is withdrawn (no eject container,
resolved pins, parallel intake), without reducing the authorized overhaul or its evidence.

**Attempt boundaries (`management/BUDGET_POLICY.md` §3–§4, §7).** Attempt 1 begins at execution release and covers
everything above through the review-ready Goal Result (P0, F/V/C/T, G, every inner implement/test/fix loop, both
owners, reruns, checkpoints, rebases, handoffs and compactions — none of these ends an attempt, §3/§4). The attempt
ends when the review-ready Result is submitted, when the attempt cannot satisfy the Goal, or when the material
strategy is abandoned (§3). **A bounded supervisor correction after that Result review is a new attempt and consumes
one increment (attempt 2)** — §3: "a new attempt consumes one increment when retrying after an unsuccessful attempt
or executing a bounded supervisor correction after review"; §4 names a "supervisor-authorized correction round" as
accounted under §3. r0's statement that a Codex correction round stays inside attempt 1 is withdrawn. This planning
correction (r1) spends nothing (Goal: "Planning/review currently spends zero implementation attempts"); B5 maximum 5,
checkpoint after attempt 3 (§7) unchanged; the Result keeps the attempt log of §14.

### 8.3 Delegation

Two Claude writers with disjoint writable paths (§8.1), both at Opus 5.5 high in isolated worktrees: **A** (`g067-impl`,
web/harness owner and the single integration owner: merges B's branch, runs rows 1–3/3r/6, owns the Result and the
receipt) and **B** (`g067-intake`, intake/docs/infra-ci owner; hands back one branch with V1/V2 evidence and row 4
output). B's handback names the `run.sh` skill-path lines A must flip in G. No other subagent writes; one optional
read-only Explore pass over the identity specs (A7). Codex: Plan review, source review, optional read-only review of
§4.1/§4.2/§5.1/§5.4. Fable test corrections at high (xhigh at a safe boundary if needed, recorded). Both writers read
the repository governance and the applicable skill: the old skill until B's handback is merged, then the new one.

## 9. Acceptance-criterion mapping

| AC | Evidence |
|---|---|
| AC1 local successor, alias, retired old directory/alias, working references | digests re-run at copy time; **V1** `sha256sum -c UPSTREAM.sha256` 319 OK on the pristine copy (before the `SKILL.md` edit) and **V2** `sha256sum -c LOCAL.sha256` 321 OK + the bound local delta on the final tree (§4.1); tracked-file count 322; `readlink` of the alias; `git ls-files` of the old paths empty at the final commit; grep for `vestive-design` outside history returns only §3.3 citations; row 4 PASS; hosted provenance (B-SKILL-UNCHANGED, B-PROTECTED at `--base HEAD`) PASS; A14–A19 PASS |
| AC2 actual Base UI/Rhea components, palette, typography, geometry; Radix completely retired from code/config/resolved tree | §3/§4.2/§5 mapping and `ADOPTION.md` lineage; the five zero-Radix proofs of §4.2 (lockfile greps, B-DEPS flipped with C13/C14, B-WEB-COMPONENTS import scan incl. tests, `npm ls --all`, repo grep with the declared vendored-CSS hit); the lockfile closure equal to the nine expected entries; `components.json` `base-rhea`/`taupe` and the unchanged eject pin; row 3 S-DESIGN/S-FIDELITY/S-PALETTE on oracles re-typed from the measured values (28/32/36px, 32px, 18px, 24px, 1px); S-FONTS phase B (five pinned responses, `custom` verdicts for Geist/Geist Mono/Oswald/Barlow); S-SCREENSHOTS + `/sign-in` + one identity-surface attachment at 1280 L/D and 320 |
| AC3 theme/branding scope, first paint, hydration, generic identity | row 3 tokens-light/dark (nested/partial/sibling scopes, portal first insertion on Base UI, invalid storage/System), appearance (first paint, delayed JS, html attrs), dev-hydration, S-IDENTITY/S-LONG-IDENTITY; row 3r renamed-identity S-IDENTITY ×2 |
| AC4 Base UI replacement preserves interactions and accessibility incl. scoped portal insertion, modality/focus/ref/event, SSR/hydration | row 3 S-OVERLAY (name, initial focus, containment, outside inert, Escape/Close, focus return, 28px Close), S-PORTAL-INIT, S-MOTION, S-HYDRATION-DEV, S-A11Y, S-CONTRAST ×2, S-TEXT-SPACING, S-TAB/S-SKIP/S-RESPONSIVE, S-ANGLE on `cta`, S-STATES, `render`/`ButtonLink` composition; row 6 identity-stack per-page a11y/forms/dialog scenarios; selectors migrated, no scenario deleted (test diffs reviewed line by line) |
| AC5 old cautions and new lineage documented, no stale mandates | `ADOPTION.md` (authority/lineage per §4.1), WEB-SHELL.md G067 sections (historical tables kept), THIRD_PARTY_NOTICES.md (Base UI closure MIT, Geist OFL with the upstream copyright line, Roboto removed, shadcn lineage reworded), ARCHITECTURE §61 record, OPERATIONS skills table; `DESIGN_REMOTE` empty |
| AC6 integration, affected qualification, cleanup | rows 1–4 and 3r on the tested candidate SHA/tree before the MR; hosted fast + Duo on the MR head; merged-SHA row 6 with tree comparison; inventory + compact receipt (all categories zero or named with owner/reason/release); Codex final ACCEPT |

## 10. Risks, escalation flags, open owner questions

**Flags (after r1's read-only discovery; E1/E3/E4 are demoted to confirmations):**
- **E1 `@base-ui/react` 1.9.0 closure** (A1, **resolved**): the declared closure is nine Radix-free MIT packages
  (§0.1); P0.1 confirms the actual lockfile equals it. Only an unexpected entry is reported; only a Radix entry stops.
- **E2 Rights (narrowed; for Codex's disposition, no owner declaration requested).** The copy/adaptation authority and
  lineage are recorded in §4.1 from the Goal/Spec and the existing notices. Genuinely unconfirmed and left so: the
  v0.2.1 tree carries no licence/notice file of its own (the owner's authorization and Claude Design origin are the
  recorded basis; nothing from it is shipped by `web/` except the adapted registry-derived modules whose shadcn/Base UI
  lineage is noticed). Codex may record a disposition; the Plan does not extend G048's six-module statement.
- **E3 Registry CSS / `components.json`** (A3, **resolved**): no CLI run; eject block and pin unchanged (identical
  bytes at 4.21.1/4.22.1); `style` `base-rhea`, `baseColor` `taupe` from the registry/docs evidence. Residual: the
  preset code was not decoded (composition from `kit/README.md`); it affects only future `shadcn add` calls.
- **E4 Font files** (A4, **resolved**): Geist/Geist Mono/OFL at `google/fonts` `9e25e2ba` with the §0.1 digests; the
  intake fetch must reproduce them or stop.
- **E5 Scanner false positive on the vendored bundle** (A6) → owner-reviewed exemption only.
- **E6 Base UI guarantee gap** (A2): if a retained guarantee cannot be met by Base UI plus a repository wrapper, stop
  and report with the smallest alternative (no silent assertion change, no Radix fallback).
- **E7 Scope.** A consumer needing a behavioural change (copy, route, form semantics) is out of scope and reported.

**Open owner questions (none blocks planning):** Q2 the brand book's self-contradicting contrast guidance (app uses
`#0062d6`; correct upstream?); Q3 D-2/D-3 defaults acceptable, or no brand-moment type/angle at all?; Q4 whether the
v0.2 icon should be considered for a later product-identity Goal (not used here). The Base UI question is **resolved**
by the owner (ADR-001 amendment) and no longer open.

**Risks:** Base UI behavioural differences surfacing only in the identity stack post-merge (mitigated by P0's
shell/fixture proof and row 6); test re-typing effort (A5, bounded by P0 and the measured values); Light derivations
missing a composite (S-CONTRAST measures pixels; fix the value); resource mistakes near the preview (private projects,
no ports, inventory); staging advancing (exact-delta rule, affected rows only); a two-writer merge conflict (prevented
by the disjoint path ownership of §8.1; A resolves any residue); Codex-owned references left stale if not relayed.

## 11. Transfer state and handback

- r0 planning worktree `.claude/worktrees/agent-<id>` was reclaimed after its push; r1 was written in
  `.claude/worktrees/agent-<id>` on branch `g067-plan` checked out from `origin/g067-plan` (`bb1175f3`,
  parent `0e957c69`); this file is the only change in both commits. No stack, image, container, volume, network or
  evidence directory created; `/tmp/g067r1-scratch` (verification script, registry/font discovery downloads in their
  own directory, never executed) is disposable.
- Implementation owners A and B: start from Codex's accepted Plan bytes in the integration branch; rebind to current
  staging (exact-delta rule); B verifies the home source digest before copying; A runs P0 first; hold one run slot and
  the build slot only while a batch runs; record the inventory from the first allocation under `<main>/.local/`.
- Codex: Plan ACCEPT / CORRECTION LOOP on §0.1, §1, §4.1–§4.3, §5.2, §6, §7, §8; disposition of E2 and E5–E7; relay of
  the Codex-owned references (§3.3) and Q2–Q4 to the owner when convenient.

## Revision history

- r0, 2026-10-25 — initial Plan, drafted against staging `b6bfffb0`/Spec r1 and reconciled in place to staging
  `0e957c69`, Spec r2 and the ADR-001 owner amendment (Base UI/Rhea overhaul, complete Radix retirement) before its
  first commit. Read-only planning: digest verification by `sha256sum`, lockfile closure inventory by reading
  `package-lock.json`, WCAG ratios computed from hex/oklch values, one read-only `docker ps` listing of the owner's
  preview ports; no implementation, install, build, test, pull, push or MR.
- r1, 2026-10-25 — consolidated correction under the Codex r0 review (CORRECTION LOOP, six items): §0.1 added
  (measured kit geometry; registry/docs/font discovery); §1.1 S1–S3, §1.2 A1/A3/A4 and §1.3 P0 rewritten (complete
  mechanical migration before the first whole-app check; `ButtonLink` for link consumers per the Base UI Button docs;
  no eject run); §2 and §8.3 two owners with disjoint paths; §4.1 two-step verification (V1 before the local edit,
  V2 on the final tree) and the accurate authority/lineage record; §4.2 exact pins and the declared vendored-CSS hit;
  §4.3 smallest CSS integration; §5.2 geometry corrected (28/32/36px, 32px fields, 18px/24px radii, 1px rules) with
  the two intentional adaptations and their concrete requirements; §5.4 three rows; §6 pins; §7.1 `lint lint-web
  controls` in one invocation, row 3r focused renamed-identity proof, `rename` reported unexecuted with the
  unchanged-input reason; §7.3 dependency graph, `.local` storage, accounted pinned creators; §8.2 recomputed estimate
  and BUDGET_POLICY §3/§4 attempt boundaries; §9/§10/§11 reconciled. Read-only discovery only: npm registry manifests
  (`@base-ui/react`, its closure, `shadcn`, `cn`), `unpkg` copies of `shadcn@4.21.1`/`@4.22.1` `dist/tailwind.css`,
  the shadcn registry style items and docs, the Base UI docs, GitHub API/raw files of `google/fonts` (hashed in a
  scratch directory, never executed), `sha256sum` of repository files; no implementation, install, build, test,
  Docker object, push or MR; the owner's preview untouched. Zero implementation attempts spent.
