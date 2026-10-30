# Web shell

`web` renders a generic, accessible application frame: a skip link, a ruled header with the configured product name, primary navigation, an appearance choice (Light, Dark or System), a keyboard-help dialog, main content and a footer showing the configured domain. It has no domain, account, tenant or Organization surface of its own. The frame is composed from a pinned design kit whose complete source is kept **unaltered** under `web/src` (shadcn/ui Rhea-style components on Base UI plus the kit's brand-layer modules); the application is adapted around it ([Unaltered kit adoption](#unaltered-kit-adoption)), styled only through semantic design tokens, with bundled fonts and an original blueprint drawing. Radix is not a direct dependency: it is present only as `cmdk`'s transitive closure ([below](#dependencies-and-the-cmdk-radix-closure-s2-s3)).

`web` stays a thin BFF: it holds no authority, policy or domain rules, it calls `core` only from server code, and it exposes no backend URL to the browser.

## Component and theming guidance

No agent skill or design tool is needed to work on the UI. In short:

- **Kit source.** The kit's components, hooks, utilities and stylesheets live unaltered in `web/src`. Compose them from app code outside `web/src`, adapting only through props, `className` and app-owned variables; never edit, regenerate or shadow a kit file. B-WEB-COMPONENTS verifies `web/src` byte for byte against the reference data `web/tests/reference/kit-source.sha256`.
- **Tokens and themes.** `web/app/globals.css` imports the kit chain and adds only the app-owned sections (the dark variant, the bundled `@font-face` rules and the blueprint layer). The palette is the kit's own `web/src/styles/theme.css`.
- **Appearance.** next-themes manages Light, Dark and System through one provider (`components/layout/theme-provider.tsx`).
- **Tenant style.** An Organization's saved branding is server-rendered into `<head>` by `components/branding/tenant-style.tsx`.
- **Checks.** The browser suite (accessibility, keyboard, appearance modes, design tokens, the BFF boundary) runs with the README `platform:browser` block; the harness `root` suite runs it as B-BROWSER-ROOT, and the `lint` suite runs B-WEB-COMPONENTS, B-WEB-TOKENS, B-WEB-BOUNDARY, B-WEB-DESIGN and B-DEPS.

## Four separate concerns

| Concern | Where | Inputs | Limits |
|---|---|---|---|
| **A. Design language** | the unaltered kit source under `web/src` (69 files: `components/ui/*`, the kit's brand-layer modules under `components/`, `hooks/use-mobile.ts`, `lib/utils.ts`, `globals.css`, `styles/*.css`); `app/globals.css` adds only the app-owned sections; `public/fonts` and `public/blueprint` | The kit's own palette (`src/styles/theme.css`), Tailwind theme mapping and components, byte for byte | No kit byte is edited, excluded from typechecking or shadowed (B-WEB-COMPONENTS fidelity); the app adapts around the kit |
| **B. Theming** | the kit's `:root`/`.light` and `.dark` variable rules, unaltered, and the app's dark variant in `app/globals.css` | The resolved `light`/`dark` class on `<html>` (or on an isolated preview) | Components use the kit's semantic utilities; app-owned code adds no palette utility or colour literal (B-WEB-TOKENS) |
| **C. Organization overrides** | `components/branding/tenant-style.tsx` (`<style data-tenant-branding>` in `<head>`), `lib/tenant-css.ts` (the fan-out), `components/branding/brand-preview.tsx` | An Organization's saved, re-checked branding (15 names × Light/Dark) | Styling only; never identity, authority or a selector of Organizations |
| **D. User appearance** | `components/layout/theme-provider.tsx` (next-themes), `components/layout/appearance-control.tsx`, the constants in `lib/appearance.ts` | `localStorage["appearance-mode"]` (written by next-themes) and the OS preference | A browser display preference: no cookie, account, Organization, schema or server meaning |

Each concern changes through its own files. An appearance change re-resolves every variable, so Organization overrides, the blueprint layer and portals keep working in both appearances.

## Unaltered kit adoption

The complete, unaltered kit `kit/src` tree is the application's UI source, and the application is adapted around it.
This section records the installed mechanism.

### Source and fidelity

- `web/src` holds exactly the 69 files listed in the reference data `web/tests/reference/kit-source.sha256` (the
  `kit/src/` rows of the pinned kit manifest) at their kit-relative paths, each byte-equal to its row; the
  source-manifest digest over them is
  `597f2aac251987be26050108dc6affe3a39a5594e7bfc7e2744de2fb5829a6ee` (B-WEB-COMPONENTS `UI_FIDELITY`/`UI_SET`;
  control A1). Nothing else lives under `web/src`, and `web/components/ui` and the transitional
  `web/components/legacy-ui` do not exist (`UI_SET`, `UI_LEGACY`).
- `web/tsconfig.json` resolves `@/*` to `./src/*` first, then `./*`, so the kit's own `@/components/ui/*`,
  brand-layer `@/components/…`, `@/lib/utils` and `@/hooks/*` imports resolve inside the kit, and app modules keep their
  paths (no specifier resolves under both roots).
- The kit tree stays typechecked and linted. One ESLint override for `src/**` turns off exactly the two rule IDs it
  trips (`@next/next/no-img-element`, `react-hooks/set-state-in-effect`); there is no `ignores` entry for it.
- Kit identity literals are allowed only inside those 69 manifest-bound files and in app import lines of the kit's
  brand-layer modules (S17; controls A20, A21). Runtime product identity stays configured.
- The kit's `SiteHeader`, `ribbon`, `node-field`, the compound `Select` and the rest of the catalogue stay compiled and
  unused where the app does not need them; the shell keeps its own header composition (S13).

### Stylesheet chain and cascade order

`app/globals.css` stays the app entry. Its first statement is `@import "../src/globals.css";` (the kit chain,
unchanged: `tailwindcss`, `tw-animate-css`, `shadcn/tailwind.css`, `styles/theme.css` and the kit's own stylesheet under
`web/src/styles/`). After it come only three app-owned sections, none of which changes a kit colour or state: the dark
variant with the Light-preview boundary (functional, below), the five bundled `@font-face` rules (self-hosted copies of
the kit's fonts) and the blueprint layer (hidden under `forced-colors: active`). The kit's colours, focus treatment,
cursor and motion render unaltered (the kit exactly; its contrast shortfalls are accepted and listed below); earlier
app-owned colour corrections, a base-layer pointer-cursor rule, a global reduced-motion rule and a forced-colours
`:focus-visible` outline are withdrawn, and the kit's own reduced-motion rule (in the kit's own stylesheet under
`web/src/styles/`) is the only one.

Cascade, lowest to highest:

1. the kit's `:root, .light` and `.dark` rules in `src/styles/theme.css` (specificity 0,1,0);
2. an Organization's saved branding, `html:root:not(.dark) { … }` (Light) and `html.dark:root { … }` (Dark) in the
   `<style data-tenant-branding>` element (0,2,1, so it wins regardless of stylesheet order);
3. an isolated preview element's inline custom properties.

The `@theme inline` mapping resolves `--color-*` at use time, so every level re-resolves the utilities.

### Dark variant and the Light-preview boundary

The kit's own definition, `@custom-variant dark (&:where(.dark, .dark *, [data-theme="dark"], [data-theme="dark"] *))`
(in the kit's own stylesheet under `web/src/styles/`), matches a Light brand preview on a Dark page, so its content
would take Dark utilities (for example the Dark destructive tint strengths) over Light values.
`app/globals.css` redefines the same variant after the kit import (Tailwind registers the last `@custom-variant` of a
name): it keeps the kit's four selectors and adds
`:not(:where([data-brand-preview].light, [data-brand-preview].light *))`. The exclusion names preview elements only, so
`html.light` never suppresses a Dark preview, and a Dark preview on a Light page still matches `.dark`. Specificity is
unchanged.

**Single-level limit.** The boundary covers exactly one level of preview inside the page (a `BrandPreview` with its own
`light` or `dark` class), not a nested-theme system: a Dark preview inside a Light preview, or any deeper nesting, is
unsupported. Previews render no portaled component (S8 rule, below).

### Kit colours and the accepted contrast shortfalls

No app-owned `--vx-blue`, `--muted-foreground`, `--destructive` or `--input` value is set (S10, S22): the kit's own
`src/styles/theme.css` values render, and the contrast shortfalls below are accepted. Core's platform map
(`core/internal/branding/colors.go`) holds the same kit values as `#rrggbb` (the Dark translucent `--border` and
`--input` as they paint over the Dark background: `#242322`, `#302f2e`), so blank saved colours inherit what renders.

| Pair (kit values) | Ratio | Previous requirement |
|---|---|---|
| White label on the `#0075ff` primary, both appearances (T8; the platform CtaButton at rest) | 4.21 | 4.5 |
| Same label on the Light 80 % hover composite (T8) | 3.18 | 4.5 |
| CtaButton label at `hover:brightness-110` | 3.77 | 4.5 |
| Field boundary (`--input`) on background, Light / Dark; Dark on popover (N3) | 1.26 / 1.48; 1.32 | 3 |
| Light `muted-foreground` on the `muted` band (T4) | 4.40 | 4.5 |
| Light `destructive` text on its 10 % / 20 % tint (T10) | 3.99 / 3.32 | 4.5 |
| Dark `destructive` text on its 30 % tint over background / popover (T10) | 4.34 / 3.73 | 4.5 |

Core's save-time contrast check evaluates its contrast pairs on the effective palette with a kit-baseline floor
(`floors`/`kitFloors()` in `core/internal/branding/colors.go`): for each pair entry and appearance the required ratio is
the lower of the existing threshold and the same entry's ratio on the platform defaults (`Effective(a, nil)`), so the
accepted kit shortfalls above are neither refused on save nor allowed to worsen, and every other pair keeps its
threshold. The browser specs record the accepted pairs (marked `OWNER-WAIVED` in the specs) at the kit's own ratio.

### Appearance (S5, S6)

- **One manager.** `components/layout/theme-provider.tsx` is the only app module that imports `next-themes`
  (B-WEB-BOUNDARY): `attribute="class"`, `defaultTheme="system"`, `enableSystem`, `storageKey="appearance-mode"`
  (preferences stored by the earlier appearance mechanism survive), `enableColorScheme={false}` (the kit's `theme.css`
  sets `color-scheme` itself). It renders once, as the first child of `<body>`; the kit's `Toaster` reads the same
  state. No app code touches browser storage; the only storage access in `web/src` is the kit sidebar's cookie (S19:
  unreachable, no app module imports `@/components/ui/sidebar`).
- **First paint.** next-themes' classic inline script runs before any body content and writes the resolved `light` or
  `dark` class on `<html>`; the palette is CSS keyed on that class. `<html lang="en" suppressHydrationWarning>` is the
  only hydration exemption (S-HTML-ATTRS: served `<html>` carries exactly `lang`; the class is written in the browser).
- **Blocked storage.** next-themes' script writes no class when storage access throws. Right after it the provider
  renders one short fallback script that, only when no class was written, adds the OS-resolved class; it reads no
  storage and sets nothing else (S-FIRSTPAINT blocked cases).
- **Stored values outside the three modes.** next-themes validates nothing: any stored string becomes the `<html>`
  class, by its script before the first paint and by its effect after hydration, and the document would keep no
  palette class. A guard rendered inside the provider (`StoredPreferenceGuard`, a layout effect) resets such a
  preference to System through the provider's own setter, which rewrites the stored value, and removes the stray
  class; it touches no storage itself, and the fallback script above has already painted the OS appearance
  (S-APPEARANCE-INVALID: `purple`, `DARK`, a JSON object; an empty value is next-themes' own default, System).
- **Init script hash.** For a future Content Security Policy hash allow-list (not yet implemented), S-HEAD-ORDER
  attaches the SHA-256 of the served next-themes script and the harness compares it with this line
  (`init_hash_documented`):

  `appearance-init` SHA-256: `73d697069cbaaabbb9ef46957cf7746bf26fc284701843523252b616b017e879`

  (it changes with the next-themes version, the S5 props or the production bundler's output, and must be updated here
  in the same change). The blocked-storage fallback's content is a fixed source literal, SHA-256
  `986a0a87852ecdf9961f8dcdc6f58c685d41906085818c61e208b32418c79016` (computed from the source, not checked by the
  harness).
- **Control (S6).** The kit's compound `Select` (`components/ui/select`, Base UI) is driven by `useTheme()`. Its
  trigger is the combobox named "Appearance" (`aria-label`), icon-only: the selected preference's symbol (Sun, Moon or
  Monitor) and the kit's chevron, inside the `border-input` frame on the page ground, mirroring the value in
  `data-appearance-value` (an icon-only combobox has no text value); `[data-appearance-control]` marks its wrapper. Its
  popup is the kit's `[data-slot=select-content]`, portalled into `document.body` and right-aligned under the trigger,
  listing Light, Dark and System with their symbols inside the kit's padded group (`p-1`), on the palette's popover
  ground and ink in both appearances. Enter, Space or an arrow key opens it, the arrow keys move the highlight, Enter
  chooses, Escape closes and returns focus to the trigger; the browser's own option list is not used. The server
  cannot know the stored preference, so until hydration the closed trigger shows the symbol of the palette already
  resolved on `<html>` (Sun or Moon, through the kit's dark variant), also for a stored System; after hydration it
  shows the preference's symbol (Monitor for System). Every palette guarantee of S-FIRSTPAINT-DELAYED is kept; only its
  "symbol while held" assertion follows the resolved palette.

### Tenant theme host and the saved-branding fan-out (S7, S9)

The root layout renders `<TenantStyle>` in `<head>` from the request's re-checked branding (`tenantRequest`, cached
per request with React `cache`), so the first HTML byte carries the tenant's values in both appearances and every
overlay the kit portals into `document.body` (dialog, sheet, popover, select, tooltip, menus, toasts) inherits them
from its first frame; nothing repairs styles after render. A loopback or unknown host, no branding, or any value that
fails the emitter's re-check (saved names only, lowercase `#rrggbb`) renders the element empty: kit and app defaults.
The persisted contract is unchanged (columns, the 15 names, hex grammar, API, audit payloads, error codes).

| Saved token | Kit variables set (`lib/tenant-css.ts` `TENANT_FAN_OUT`) |
|---|---|
| `background` | `--background`, `--sidebar` |
| `foreground` | `--foreground`, `--accent-foreground`, `--sidebar-foreground` |
| `primary` / `primary-foreground` | `--primary`, `--vx-blue`, `--sidebar-primary` / `--primary-foreground`, `--sidebar-primary-foreground` |
| `accent` | `--vx-gold` only (emphasis ink, Kicker, TextLink, selection, default ring); never the kit's neutral `--accent` |
| `muted` / `muted-foreground` | `--muted`, `--accent`, `--sidebar-accent` / `--muted-foreground` |
| `secondary` / `secondary-foreground` | `--secondary` / `--secondary-foreground` |
| `popover` / `popover-foreground` | `--popover`, `--card` / `--popover-foreground`, `--card-foreground` |
| `border` | `--border`, `--sidebar-border` |
| `input` | `--input` (boundary, 50 % field fill, 90 % radio fill) |
| `destructive` | `--destructive` (text on its tint, error text) |
| `ring` | `--ring`, `--sidebar-ring` |

Not overridable: `--radius`, `--vx-angle`, the charts, the other brand colours and the ribbon. The retired `overlay`
token is gone: the kit's dialog scrim is its theme-fixed `bg-black/30` with a backdrop blur (S16, E2). Saved rows are
re-judged by Core's current contrast pairs on their next save (S23): a row accepted under an earlier pair set may then
be refused with the same pair ID or T12; nothing is migrated.

### Isolated previews (S8)

`components/branding/brand-preview.tsx` renders `<div data-brand-preview="<label>" class="light|dark">` with one
appearance's values as inline custom properties (the same fan-out). The kit's palette rules match any element with the
`light` or `dark` class, so the preview's subtree takes its own appearance and values; the document root and every
other surface are unaffected. **A preview renders no portaled component**: portals land in `document.body`, outside the
preview, with the live tenant's values. Production has no live preview today; the token fixture page is the proof
surface.

### App compositions over the kit (`web/components/controls`, S12, S22, S24)

| Module | Use |
|---|---|
| `button-link.tsx` `ButtonLink` | A link styled as a button: a native `<a>` with the kit's `buttonVariants()` and `data-slot="button"`; links are never rendered as buttons. `variant="cta"` applies the kit's `ctaButtonVariants()`; it is the only app module importing the CTA classes |
| `pending-submit.tsx` `PendingSubmit` | The submit of a native POST form on the kit `Button`, with the accessible busy state (`aria-disabled`, pending label, polite status); a destructive submit keeps the kit's own focus treatment |
| `form-select.tsx` `FormSelect` | Every form dropdown (never the browser's option list): the kit `Select`/`SelectTrigger`/`SelectValue`/`SelectContent`/`SelectGroup`/`SelectItem` composed for a native POST form. Props `id` (on the trigger, so the site's `<label htmlFor>` names the combobox), `name`, `options: {value,label}[]` (also the Root's `items`, so the closed trigger shows the chosen label), `defaultValue` (the first option unless given, as the native select chose), `required`, `className`. Trigger `w-full`, otherwise the kit trigger as shipped; popup `alignItemWithTrigger={false}` with the items in the kit's padded group; the kit's hidden input posts `name=value`. Sites: the members table role selects, the ownership transfer form, the invitation form and the Group labels assign form |
| `radio-choice.tsx` `RadioChoice` | One option of a kit `RadioGroup` (`name`, `defaultValue`, `required` on the group: one hidden input per group), with the kit `Label` for layout |
| `scroll-region.tsx` `ScrollRegion` | A labelled, focusable horizontal scroll region (`role="region"`, `tabIndex=0`) around a kit `Table` |

**CtaButton placement (S24).** The kit `CtaButton` has a fixed white label on `--vx-blue` (`hover:brightness-110
active:brightness-90`), which cannot follow a tenant's `primary-foreground`. It renders only where the platform palette
is guaranteed: the root-host landing's Sign Up link (`components/signup/landing.tsx`, through `ButtonLink`, with the
S15 inset focus outline `focus-visible:outline-offset-[-3px] focus-visible:outline-white`) and the test-only token
fixture under its platform palette. The tenant-served sign-in submit is a kit `Button size="lg"`. B-WEB-COMPONENTS
`UI_CTA_TENANT` (control A22) and Core's platform-only CTA assertion in `TestDefaultsPass` enforce it. **No placeholder
text** in app pages and components (S22, `UI_PLACEHOLDER`, control A23): the kit placeholder is `muted-foreground` on
the 50 % input fill, below 4.5:1, so fields carry visible descriptions instead.

### Dependencies and the cmdk Radix closure (S2, S3)

Exact pins (B-DEPS): runtime `cmdk` 1.1.1, `clsx` 2.1.1, `tailwind-merge` 3.7.0, `next-themes` 0.4.6, `sonner` 2.0.8,
`react-day-picker` 9.14.0, `embla-carousel-react` 8.6.0, `input-otp` 1.5.0, `react-resizable-panels` 4.15.0,
`recharts` 3.10.1 (the packages the kit source imports); dev-only `shadcn` 4.22.1 (only for the kit's
`@import "shadcn/tailwind.css"`; no CLI is run); no `cn` package (the kit's `lib/utils.ts` provides `cn`).
`@base-ui/react` 1.9.0, `class-variance-authority`, `lucide-react` and `tw-animate-css` are pinned as well. Base UI is
the primary primitive library; `@base-ui/` is imported only inside `web/src` (`UI_IMPORT_BASEUI_APP`, control A12b),
and no web file, tests included, imports `@radix-ui/` (`UI_IMPORT_RADIX`).

**Radix only as cmdk's transitive closure.** The kit's `components/ui/command.tsx` imports `cmdk`, whose module graph
imports `@radix-ui/react-dialog`; that is the only reason any Radix package is present.
`DEPS_RADIX_VIA_CMDK` (`infra/validation/lib.sh`) lists the exact 16 entries; any other `@radix-ui/` lockfile entry at
any depth is `DEPS_RADIX_UNEXPECTED`, a direct `@radix-ui/` dependency is `DEPS_RADIX_DIRECT`, and `cmdk` must stay
direct while the list is non-empty (controls C13, C14). All are MIT.

| Package | Version | Reason it is present |
|---|---|---|
| `@radix-ui/react-dialog` | 1.3.1 | imported by `cmdk` 1.1.1 (its only dependent) |
| `@radix-ui/primitive` | 1.2.0 | event-handler composition used by the dialog |
| `@radix-ui/react-compose-refs` | 1.1.5 | ref composition used by the dialog and its parts |
| `@radix-ui/react-context` | 1.3.0 | scoped context used by the dialog |
| `@radix-ui/react-dismissable-layer` | 1.2.1 | outside-click and Escape dismissal of the dialog |
| `@radix-ui/react-focus-guards` | 1.2.0 | focus sentinels around the modal dialog |
| `@radix-ui/react-focus-scope` | 1.3.1 | focus trapping inside the dialog |
| `@radix-ui/react-id` | 1.1.4 | stable ids for the dialog's title and description |
| `@radix-ui/react-portal` | 1.2.0 | the dialog's portal |
| `@radix-ui/react-presence` | 1.1.11 | mount/unmount around the dialog's animations |
| `@radix-ui/react-primitive` | 2.2.0 | the element primitives of the dialog parts |
| `@radix-ui/react-slot` | 1.5.0 | `asChild` composition inside the dialog parts |
| `@radix-ui/react-use-callback-ref` | 1.1.5 | stable callbacks used by the layers above |
| `@radix-ui/react-use-controllable-state` | 1.3.0 | the dialog's open state |
| `@radix-ui/react-use-effect-event` | 0.0.5 | effect events used by the controllable state |
| `@radix-ui/react-use-layout-effect` | 1.1.4 | SSR-safe layout effects used by the parts |

Non-Radix helpers in the same closure (also MIT): `aria-hidden` 1.2.6 and `react-remove-scroll` 2.7.2 (imported by
`@radix-ui/react-dialog`: inert siblings and scroll lock), `react-remove-scroll-bar` 2.3.8, `react-style-singleton`
2.2.3, `use-callback-ref` 1.3.3 and `use-sidecar` 1.1.3 (imported by `react-remove-scroll`), `get-nonce` 1.0.1 (by
`react-style-singleton`) and `detect-node-es` 1.1.0 (by `use-sidecar`). No app module renders `Command`; the closure is
compiled with the kit and never reached by an app import.

### Configuration

`components.json` keeps RSC, TSX, `app/globals.css`, CSS variables, lucide icons and the aliases, with `style`
`base-rhea` (the registry's Base UI Rhea style; the name is `<base>-<style>`) and `baseColor` `taupe` (B-WEB-COMPONENTS
checks both). These strings only steer future `shadcn add` calls; no CLI is run.

### Harness rules

B-WEB-COMPONENTS (fidelity against `web/tests/reference/kit-source.sha256`, `UI_LEGACY`, imports, kit `data-slot` set
plus `button`, `UI_CTA_TENANT`, `UI_PLACEHOLDER`), B-WEB-TOKENS (kit palette source, the import statement, the pinned
correction names, 15 `SEMANTIC_COLOR_TOKENS`, colour literals only in the correction rules, palette-utility and
accent-surface bans in app-owned code with the one S15 `focus-visible:outline-white` exception), B-WEB-BOUNDARY (the
next-themes provider contract, no app browser storage, the S19 sidebar allowance), B-WEB-DESIGN (the S17
manifest-bound literal exemption), B-DEPS (the pins and the Radix closure), B-NAMING (`web/src` is the kit source root,
not a redundant path) and the render register (`web/tsconfig.json` registered; 32 customizations). `APPLICABLE` copies
`web/src` and `web/tsconfig.json` into every variant render, so the intelligence-disabled `off` copy resolves the kit.

### Known limitations

- The kit's outline Button keeps its `border-border` hairline (about 1.15–1.41:1); every outline site has a text label,
  which identifies the control (WCAG 1.4.11), so the oracle records the ratio instead of requiring 3:1.
- The kit `RadioGroup` changes its selection only with JavaScript; without it the hidden input still submits the
  default value and the server validates as before (S16).
- The dialog scrim is the kit's theme-fixed `bg-black/30` with blur, the same in both appearances (S16, E2).
- The Light-preview boundary is single-level and previews render no portal (S8).
- Saved rows are re-judged on their next save under the current contrast pairs (S23); no migration.
- The blocked-storage fallback script (above) is a second, storage-free inline script.
- The accepted kit contrast shortfalls listed above.

## Token contract

The palette is the kit's own `theme.css`, unaltered, with the saved-branding fan-out of
[Tenant theme host](#tenant-theme-host-and-the-saved-branding-fan-out-s7-s9). `SEMANTIC_COLOR_TOKENS` in
`lib/tokens.ts` holds the 15 overridable saved names (`background`, `foreground`, `primary`, `primary-foreground`,
`accent`, `muted-foreground`, `border`, `popover`, `popover-foreground`, `muted`, `secondary`, `secondary-foreground`,
`destructive`, `input`, `ring`; no `overlay`), matching core's branding tokens; B-WEB-TOKENS keeps them consistent.

**`accent` is ink.** A saved `accent` sets the kit's gold emphasis ink (`--vx-gold`), never the kit's neutral
`--accent` surface; B-WEB-TOKENS bans accent surfaces in app-owned code.

## Fonts

Application UI uses Geist and Geist Mono (the kit's app type); Oswald and Barlow Condensed are used for the brand
moments (header identity, landing H1, `cta`, brand kicker and the footer identity line). All are unmodified upstream
files from the `google/fonts` repository, served same-origin from `web/public/fonts` (the standalone web image copies
`public/`; the browser image's fixture build does the same):

| Served path | Upstream | Size | SHA-256 |
|---|---|---|---|
| `/fonts/geist/geist-wght.ttf` | `ofl/geist/Geist[wght].ttf` | 169056 | `73894e0448cae90a92b6c2f8732b7bb9acb7b94c418bff559dad4a18e1de9659` |
| `/fonts/geist/OFL.txt` | `ofl/geist/OFL.txt` | 4387 | `1781d2806a07d91c4edf4740b88449fab7d0eadad53f7c351b94cd4d4eb8c00f` |
| `/fonts/geist-mono/geist-mono-wght.ttf` | `ofl/geistmono/GeistMono[wght].ttf` | 171948 | `d00e590b8eb3a59acc329b2d044fd143ae935090b7da33199ebee27cc7de8196` |
| `/fonts/geist-mono/OFL.txt` | `ofl/geistmono/OFL.txt` | 4387 | `1781d2806a07d91c4edf4740b88449fab7d0eadad53f7c351b94cd4d4eb8c00f` |
| `/fonts/oswald/oswald-wght.ttf` | `ofl/oswald/Oswald[wght].ttf` | 172088 | `5b38c246e255a12f5712d640d56bcced0472466fc68983d2d0410ec0457c2817` |
| `/fonts/oswald/OFL.txt` | `ofl/oswald/OFL.txt` | 4483 | `0fd731a904b729a4e02eaf5e8ebd06783edd9abe400e8882760160230675b652` |
| `/fonts/barlow-condensed/barlow-condensed-400.ttf` | `ofl/barlowcondensed/BarlowCondensed-Regular.ttf` | 102480 | `583cec5da3b84bc4dc7c9c72e2a565c94d34e431518b19d7e250b7830ad5f996` |
| `/fonts/barlow-condensed/barlow-condensed-600.ttf` | `ofl/barlowcondensed/BarlowCondensed-SemiBold.ttf` | 109428 | `7b619d14bc2327509a9ef32b0890f709626f7ecc9ff61191c2a4314c5499d2d9` |
| `/fonts/barlow-condensed/OFL.txt` | `ofl/barlowcondensed/OFL.txt` | 4377 | `186d750eb496a4c17a76385f82be6aea2ac1cf2de074a811d63786cf374ea73f` |

- **Licences.** SIL Open Font License 1.1. Each `OFL.txt` ships byte-exact beside its font; `web/THIRD_PARTY_NOTICES.md`
  lists them. Copyright lines, reproduced: "Copyright 2024 The Geist Project Authors (https://github.com/vercel/geist-font)";
  "Copyright 2016 The Oswald Project Authors (https://github.com/googlefonts/OswaldFont)"; "Copyright 2017 The Barlow
  Project Authors (https://github.com/jpt/barlow)". The files are unmodified (no Reserved Font Name issue), not sold by
  themselves, and distributed under the OFL only. Italic Geist faces exist upstream but have no consumer and are not
  bundled.
- **Loading.** Five `@font-face` rules in `app/globals.css` (Oswald, Barlow Condensed 400 and 600, Geist, Geist Mono),
  `url()` sources only (never `local()`, so an installed font can never satisfy a face), `font-display: swap`. No font
  package, `next/font`, preload, postinstall or CDN. First paint and the appearance proof never wait for fonts.
  B-WEB-DESIGN expects exactly five, and the web CSS `url()` targets are the five fonts and `/blueprint/overview.svg`.
- **Pins.** `WEB_PUBLIC_PINS` lists 12 files: the 11 files of `web/public` (the Oswald, Barlow Condensed (400, 600),
  Geist and Geist Mono binaries, their four `OFL.txt` files, `/blueprint/overview.svg` and the product mark
  `/brand/mark.png`) and the favicon `web/app/icon.png` (the same bytes as the mark); B-WEB-DESIGN rejects any unpinned
  file in `web/public`.
- **Oracle (S-FONTS).** Phase A reads what is rendered now (Chromium `CSS.getPlatformFontsForNode` and `document.fonts`
  status) without waiting for a font; phase B runs after a bounded readiness gate and checks five pinned responses, the
  bytes against the pins, the binaries' own family names, axes and weight classes, `custom` render verdicts for Geist,
  Geist Mono, Oswald and Barlow Condensed, computed weights and a synthesis twin check. Controls: held fonts must be
  detected as fallback while the page stays readable, operable and reflowing; a 400-only face at 700 must be detected
  as synthesized; swapping fonts in must not move focus or break the open dialog.

## Imagery

`web/public/blueprint/overview.svg` is original, hand-authored geometry (no traced, imported or generated third-party drawing): an unfinished floor plan with dimensions and a partial axonometric volume. SHA-256 `942a2a242594773f6be7e638d6e07f9f723a0ff6b6c6e8334427702e52f98c65`, 2474 bytes; groups `construction` (11 drawable elements, stroke opacity 0.35), `dimensions` (15, 0.7) and `structure` (18, 1).

- **Safety (B-WEB-DESIGN).** Well-formed SVG with root `viewBox="0 0 960 640"`; elements `svg g path line polyline polygon rect circle` and the listed presentation attributes only; no `<!`, `<?`, `href`, `url(`, `script`, `style`, event attribute, `foreignObject`, `<image`, `<use`, `<animate`, `<set`, `@import` or `data:`; at most 40960 bytes; the three motif groups present.
- **Theme mapping.** The drawing only supplies alpha: the `[data-blueprint]` layer in `app/globals.css` paints `var(--muted-foreground)` at opacity 0.5 through `mask-image`, so the strokes follow the resolved tokens in both appearances (low emphasis).
- **Placement.** `BlueprintBackdrop` (`components/layout/blueprint-backdrop.tsx`) renders an empty `aria-hidden` element as the first child of the overview container (`isolate`). From 1024px it covers the right 46% at 0.75 scale with a 96px fade, beside the kicker, H1 and lead; below 1024px it sits in the 88px bottom padding band, never behind text. It is absolutely positioned (no layout change), has no animation, is not focusable, and is hidden under forced colours. A missing image paints nothing.
- **Verification (S-BLUEPRINT).** Decorative semantics, the same-origin pinned response, painted strokes against the expected composite (and the ground with the layer hidden), each motif's visible contribution, the missing-asset and forced-colours cases, and per-scope stroke colour in the fixture.

**Product mark and favicon.** The template ships a default product mark, the SaaStacular "S!" mark (a first-party Vitruvian Technology mark, not licensed by Apache-2.0; see `THIRD_PARTY_NOTICES.md`), as two byte-identical 192 px PNG rasters of the approved master: `web/public/brand/mark.png` and the favicon `web/app/icon.png` (Next's metadata-file convention). The header draws the mark decoratively (`alt=""`, 28 px, inside the identity paragraph before the name) only where the identity text is the configured product name: the generic frame (including the shared auth host's sign-in page), the public entry frame (the platform root and `www`) and a tenant frame without a saved label or logo. A tenant with a label or logo renders as before. Besides tenant hosts, the host gate serves the mark and favicon on the platform root, `www` and the auth host for a GET or HEAD of exactly `/brand/mark.png` and `/icon.png`, and nowhere else (`lib/host.ts` `productAssetAllows`). The mark is a file convention, not configuration: when you rename the product, replace both files with your own square PNGs, or delete `web/public/brand/mark.png` for a text-only header (`lib/product-mark.ts` shows the mark only while that file exists in the served `public/` directory) and `web/app/icon.png` for no favicon; then update or remove the file's `WEB_PUBLIC_PINS` row in `infra/validation/lib.sh`.

## Adding a component

The whole kit catalogue is already present, unaltered, under `web/src`: compose a kit component from app code (outside
`web/src`), adapting through props, `className` and app-owned variables only, and never edit, regenerate or shadow a
kit file. A component outside the kit is an architecture change: it needs B-DEPS review of any new package
(`infra/validation/lib.sh`), the reviewed import and `data-slot` sets in B-WEB-COMPONENTS, the semantic tokens instead
of palette literals, and browser-suite coverage. A new component library, theming package or font package is an
architecture change.

## Browser tests

`README.md` (`platform:browser`) runs the suite with the stack running. The runner is `infra/docker/web-browser.Dockerfile` (Playwright 1.63.0 image pinned by digest) through `infra/compose/browser.yml`, in the `web` container's network namespace with no published port. Its `fixture` stage builds the same source with two test-only routes, `/token-fixtures` and `/account-fixtures` (the real My Account composition with a scripted session; the fixture build's unreachable core makes its connection manager render the unavailable state), which exist only in that image (production returns 404).

| Server (inside the runner) | Purpose |
|---|---|
| `WEB_BASE_URL` (the running `web`) | Shell, overlay, accessibility, boundary, design, font, imagery and appearance scenarios |
| `127.0.0.1:3001` fixture build | Token scopes: default, fixture A, an adjacent default scope, fixture B, a Light-only scope (`light-only`) and a Dark-only scope nested in A (`dark-only-in-A`). Each scope shows text, a link, the button variants enabled and disabled, a size row (`sm`, default, `lg`, icon, a `ButtonLink` link), a `block` button, a typography sampler, font probes, a raised panel, a 320×160 blueprint panel and the help dialog; `?open=<scope>` renders a dialog open at mount in that scope |
| `127.0.0.1:3002` fixture build with empty `PRODUCT_NAME` | Fail-closed configuration |
| `127.0.0.1:3003` `next dev` | Dev-mode hydration warnings |
| `127.0.0.1:3004` fixture build with a 64-character `PRODUCT_NAME` | Long configured identity (S-LONG-IDENTITY) |

| Project | Scenarios |
|---|---|
| `shell-light`, `shell-dark` | S-IDENTITY, S-LANDMARKS, S-NAV, S-SKIP, S-TAB, S-OVERLAY, S-MOTION, S-RESPONSIVE, S-A11Y, S-CONTRAST, S-BOUNDARY, S-DESIGN, S-ANGLE, S-SCREENSHOTS, S-PALETTE, S-TEXT-SPACING, S-FIDELITY, S-FONTS, S-BLUEPRINT, S-LONG-IDENTITY, S-ACCOUNT-ANON, S-ACCOUNT-UNAVAILABLE, S-ACCOUNT-LINK, S-ACCOUNT-A11Y (the account and sign-in surface on a stack without a configured identity provider), S-ACCOUNT-MCP-UNAVAILABLE (`/account-fixtures` on the fixture build: Session and Sign-out intact with the connection manager unavailable, no MCP form, axe at 1280 and 320, production 404); `shell-light` also S-HEAD-ORDER, S-HEALTH, S-NOFIXTURE, S-FAILCLOSED |
| `tokens-light`, `tokens-dark` | S-A11Y, S-CONTRAST, S-TOKENS, S-PORTAL, S-PORTAL-INIT, S-RUNTIME, S-ANGLE, S-SCREENSHOTS, S-STATES, S-TEXT-SPACING, S-FONTS, S-BLUEPRINT |
| `appearance` | S-APPEARANCE-DEFAULT, -LIVE, -OVERRIDE, -RETURN, -PERSIST, -INVALID, -NOSTORAGE, S-FIRSTPAINT, S-FIRSTPAINT-DELAYED, S-HTML-ATTRS |
| `dev-hydration` | S-HYDRATION-DEV |
| `identity-light`, `identity-dark` (only with `IDENTITY_E2E=1`, against the HTTPS identity environment; `IDENTITY.md` `identity:browser`) | The L2 identity scenarios (E-LOGIN … E-PRIVACY, A-1/A-3/A-7); no inherited project runs there, and no fixture server starts |
| `onboarding-light`, `onboarding-dark` (likewise only with `IDENTITY_E2E=1`, on the same `IDENTITY.md` `identity:browser` line) | E-ONB-1..8 (`tests/browser/onboarding.spec.ts`; E-ONB-8 covers explicit Group authority) and the member-page S-LANDMARKS/S-NAV: banner, Primary and Workspace areas navigation, main and contentinfo; `aria-current` only on the page's own area and none on `/account`; Primary anchors and both area links by keyboard. The `shell-*` projects are fixture-only, so they keep S-LANDMARKS on the generic frame, which has no areas navigation |
| `membership-light`, `membership-dark` (likewise only with `IDENTITY_E2E=1`, on the same `IDENTITY.md` `identity:browser` line) | E-MEM-1..10 (`tests/browser/membership.spec.ts`; E-MEM-8..10 cover explicit Group authority) and the members-page S-LANDMARKS/S-NAV on `/organization/members`, `/invitations` (area Members) and a Group's members page (area Groups): the three Workspace areas Groups, Business Profiles and Members, `aria-current` on the page's own area, Primary anchors and the Members area link by keyboard |
| `enrollment-light`, `enrollment-dark` (likewise only with `IDENTITY_E2E=1`, on the same `IDENTITY.md` `identity:browser` line) | E-ENR-1..5 (`tests/browser/enrollment.spec.ts`): the join entry of each mode, open signup and the Organization invitation entry through the ordinary fresh sign-in, the applicant status page (pending, approved, withdrawn, denied, closed-after-start; no tenant data), the owner's policy form and the administrators' request list with the wrong-authority refusals, keyboard-only entry and policy change, axe and reflow at 1280 and 320 |
| `signup-light`, `signup-dark` (likewise only with `IDENTITY_E2E=1`, on the same `IDENTITY.md` `identity:browser` line; base `https://localhost:543`) | E-SUP-1..6 (`tests/browser/signup.spec.ts`): root and `www` landing with one Sign Up, fresh and existing identity into a new tenant host through the automatic continuation, field errors, keyboard and axe at 1280 and 320, the captured-secret, recovery and no-re-mint-on-GET cases, and first use of the new tenant without onboarding |
| `labels-light`, `labels-dark` (likewise only with `IDENTITY_E2E=1`, on the same `IDENTITY.md` `identity:browser` line) | E-LAB-1..6 (`tests/browser/labels.spec.ts`): Organization label create/rename/delete on `/organization/labels` by keyboard (another Organization admin sees it, a member gets 404); assign/unassign on a Group page, with no Labels section and no label markup for the Group's owner and member; Group-owned labels on `/groups/<id>/labels` (the "cannot be applied" sentence, the `target_members`-only member list, `labels-not-administered` for Organization staff with the override on or off, 404 for others); immediate visibility narrowing and role removal; forged POSTs answered "Not permitted"; axe WCAG 2.2 AA at 1280/320 in both appearances and S-ERRORS; plus S-LANDMARKS/S-NAV on `/organization/labels` |

Expected values are typed independently and never read from the implementation: `tests/browser/support/modes.ts`, `support/source-reference.ts` (the source geometry, type and colour roles of each element, cited by kit file and class) and `support/assets.ts` (the asset pins).

- **S-FIDELITY** measures the source reference table at 1280×800 (±1px; weights, font stacks, borders, radius, shadows and case exact) plus the 320 and 640×400 expectations, and the presence of each module's `data-slot` on the production page.
- **S-PALETTE** compares the palette sources in both appearances and the mapped consumers as rendered.
- **S-STATES** checks every variant's rest, hover, press, focus and disabled states, the sizes and `block` in every fixture scope.
- **S-CONTRAST** adds local worst-case text contrast over the imagery (kicker, H1 and lead at 1280, 320 and 200% text), with an injected bar that must be detected.
- **S-PORTAL-INIT** proves, from the DOM mutation history recorded before any page script, that a dialog open at mount is inserted only where expected and first paints with the expected values; injected controls (a body insertion, a same-task move, a moved subtree and an extracted node) must each be detected.
- **S-TEXT-SPACING** is a bounded spacing trial for touched shell and control text (line height 1.5, paragraph spacing 2em, letter spacing 0.12em, word spacing 0.16em; not full WCAG 1.4.12 certification) with the bundled fonts loaded: an expected-element inventory with exact texts, visibility, no covering, clipping or overlap, real scroll reachability and usable controls, after focused negative cases that each must be detected.
- **First paint.** S-FIRSTPAINT and S-FIRSTPAINT-DELAYED expect the first contentful shell element in the resolved palette; a run without a content sample at or before the first contentful paint is rejected, never passed. The raw timeline is attached as `first-paint-timeline.json` before any assertion.

S-ERRORS applies to every test outside `dev-hydration`: any `pageerror` or `console.error` fails the test. Viewports: 320×640, 768×1024, 1280×800, 640×400 (1280×800 at 200% zoom) and 200% text at 1280 wide. Contrast is measured on rendered pixels, so translucent layers, hover brightening, angled fills and imagery are measured as painted. Every wait in the font oracle has a named bound that fails on expiry. Runs use Chromium only, `retries: 0` and one worker; the harness requires every scenario to pass in each listed project with nothing skipped or flaky.

The harness retains a run-owned one-off browser container long enough to copy internal `/tmp/report` output to host
ownership, then removes it. The README browser command and all assertions stay unchanged; the harness's collection
adaptation replaces `--rm` with a unique name/run label and internal report paths. See
[operations](OPERATIONS.md#browser-report-ownership).

### Authoring reliable form interactions

Browser rendering, hydration, focus transfer and controlled form updates can finish at different times. Before an
action, wait for the specific observable state it requires, using Playwright's locator assertions or a bounded
`expect.poll`. A visible control or highlighted option alone does not establish keyboard readiness.

- **Text inputs:** establish application hydration and that the field is editable before filling it. Use keyboard
  events when the behavior depends on individual keystrokes; assert the resulting value and relevant application
  state rather than assigning DOM properties directly.
- **Validation:** trigger the event the form actually uses (input, blur or submit), then assert its visible result.
  Filling a field does not itself prove that blur or submit validation ran.
- **Selects and comboboxes:** distinguish selected value, highlighted option and DOM focus. Before sending navigation
  keys, assert the focus state required by that component's interaction model. In `appearance.spec.ts`'s `choose`
  helper, the highlighted option must also hold focus before the next arrow press; the prior highlight-only wait
  intermittently sent a key before focus arrived. Keep checks for the chosen value, closing, returned focus and
  persistence where those behaviors apply.
- **Checkboxes, radios and switches:** assert checked state separately from focus. Cover label activation and
  keyboard interaction when relevant to the behavior under test; clicking a label alone is not proof of selection.
- **Animated and portaled controls:** locate the live popup, wait for the required visibility and focus, and let
  relevant layout animation finish before measuring geometry. For tests of motion or first paint itself, install
  observation before the triggering action and retain the transition assertions; do not wait past the behavior
  being tested. An animation event alone can be missed and is insufficient proof of the complete interaction.
- **Submission:** assert the intended completion or error outcome, including busy/disabled behavior where relevant.
  A click or dispatched request alone does not prove that the operation completed.

Avoid fixed sleeps, forced interactions or automatic retries as repairs for unexplained failures. Preserve the
failed assertion and use focused repeated reproduction to investigate timing. When attributing a failure to a
change, compare with the unchanged source under comparable conditions. Fix synchronization without weakening
behavioral assertions, preserve failure history and rerun the affected checks. These precautions do not classify
every intermittent failure as a test defect: component behavior and application errors still require diagnosis.


## Organization branding

- **Source of truth.** Core alone decides what a host is and what its public branding is. `lib/branding.ts`
  (server-only) asks core's public projection (`GET /internal/v1/branding/public?origin=`) with the request's diagnostic
  headers and no credential, and re-checks every answer: exactly the 15 overridable names (the contract names without
  `overlay`), lowercase `#rrggbb`, the label's bounds, the logo `v` (12 hex). A malformed answer is no branding.
- **Tenant frame.** On a platform or custom tenant host, `AppShell` receives `brand`: the overrides are the root
  layout's server-rendered `<style data-tenant-branding>` in `<head>` (both appearances, so first paint is correct; see
  [Tenant theme host](#tenant-theme-host-and-the-saved-branding-fan-out-s7-s9)), the header shows the brand mark (logo
  or label, else the product name) and the current-Organization indicator (anonymous, member, not a member, or nothing
  when the session read fails), and the footer shows `PRODUCT_NAME · PLATFORM_BASE_DOMAIN`. A branding outage keeps the
  tenant frame with platform defaults; the loopback profile and unrouted hosts keep the generic frame unchanged.
  Branding never sets or forces an appearance. A tenant page's branding read (and the MCP page's indicator session
  read) precede the page's own core call and carry lookup request IDs.
- **Logged-out entry.** A tenant host's `/` shows only the brand mark, the workspace name, one Sign in action, the
  appearance control, keyboard help and the attribution; the header omits its own Sign in on the entry and sign-in pages.
  The Sign in action is the sign-in page's own guarded `POST /auth/login` form (a busy-state submit), as is the
  onboarding sign-in-first notice; `/sign-in` stays for its closed notices and the provider's logout return. The
  platform root and `www` are not tenant hosts: their `/` is the platform landing below, never this entry.
- **Logo.** `GET /branding/logo?v=<sha12>` serves the current normalized PNG/JPEG for this request's own origin with
  `public, max-age=86400, immutable`, an ETag and `nosniff`; anything else is a `no-store` 404. Already-cached bytes stay
  usable for up to 24 hours after a replacement or removal; every new page references only the current `v`.
- **Management.** `/organization/branding` (owner or admin, core-authorized; everyone else and unrouted hosts get the
  fixed 404, anonymous visitors a sign-in-first page) holds native POST forms following `/connections/mcp`: the label and
  15×2 colour fields (blank inherits the platform default), the logo upload (multipart reduced to the raw image, at most
  512 KiB), logo removal, deletion of unreferenced uploads, and reset. Each form carries the session's CSRF token and the
  view's version; outcomes return to the page as one closed status (including the failing contrast pair and a stale
  `version_conflict`), while session, CSRF and origin refusals use the form layer's fixed pages. `/account` links to the
  page only when the management read succeeds.
- **Tests.** The `branded` fixture (a complete passing palette) joins S-TOKENS, S-PORTAL, S-PORTAL-INIT,
  S-STATES, S-CONTRAST and S-TEXT-SPACING; E-BRD-1..6 in `identity.spec.ts` (`domain-platform`, `domain-custom`) cover
  the branded and unbranded hosts, the indicator, the management flow and per-host isolation.

## My Account and AI connections

`/account` renders, for a signed-in session, the shared `AccountSections` composition
(`web/components/account/account-sections.tsx`): the Session section, the AI connections sections of the shared
server-rendered `ConnectionManager` (`web/components/account/connection-manager.tsx`) and the Sign-out section with its
three forms. `/connections/mcp` remains a compatibility entrypoint rendering the same manager with its existing ids,
whole-page sign-in-first, not-available and invalid-page branches and `cursor`/`status` parameters; on `/account` the
manager uses `mcp_cursor` and `mcp_status` (closed tables, disjoint from the account notices; unknown or repeated values
show nothing) and the anchors `#mcp-connections`/`#mcp-create`. The manager makes one core call (the list) and reduces
every failure to a closed state inside its own section (unavailable, not available, invalid page), so Session and
Sign-out always render; no `cookies()` call, no client component of its own and no new primitive.

| Route | Purpose |
|---|---|
| `POST /connections/mcp/create` | create a connection with an optional `display_name`; omitted or blank keeps the unnamed create |
| `POST /connections/mcp/rename` | set, change or clear (empty) one own connection's name |
| `POST /connections/mcp/revoke` | revoke one own connection (unchanged) |

All three take the closed `return_to` (`account` or `connections`; omitted: `connections`; any other value or a repeat:
400) and redirect to `/connections/mcp?status=<code>` or `/account?mcp_status=<code>#mcp-connections` with the closed
codes `created`, `renamed`, `revoked` and `invalid_name`. The name field alone is read strictly from the raw body
(`request.clone()`): invalid UTF-8 or an encoded surrogate, a control or line-separator character, or more than 80 code
points after trimming redirects with `invalid_name` without calling core; a repeated name is 400; every other field
keeps `readForm`'s decoding. The raw name never appears in a URL, redirect, status page or log; names render as
escaped text with safe wrapping, the input has no `maxlength` (code points, not UTF-16 units, are the limit), and the
row shows the bound client separately. Evidence: `mcp-connections` and `mcp-forms` unit tests, M-SETUP, M-RENAME,
M-ACCOUNT and M-A11Y (identity environment) and S-ACCOUNT-MCP-UNAVAILABLE (fixture build).

## Group onboarding pages

Server-rendered, never cached, native forms with the session's CSRF token and exact-Origin checks; every read and write
goes to core's internal onboarding routes through `lib/onboarding-client.ts` (strict answer shapes and per-endpoint
size bounds; anything unexpected is `malformed`). Pages exist on tenant hosts only (elsewhere the fixed 404); anonymous
visitors get a sign-in-first page. The model is in [TENANCY.md](TENANCY.md#group-onboarding).

| Route | Purpose |
|---|---|
| `GET /groups?after=` | the Groups the member may read, in core's order with a Next link: name, slug, state (Incomplete / Complete), Open or Resume onboarding; New Group |
| `GET /groups/new?name=` | step 1 the name, step 2 core's suggested slug as an editable value |
| `POST /groups/create` | create; then `/groups/<uuid>/onboarding`. Core's definitive conflict on the slug returns to step 2 (`?error=slug_taken&name=…&slug=…`, the slug only when grammar-valid) with the field error and "Nothing was created by this request"; the page's fresh `GET /internal/v1/groups?slug=` read adds the holder's name and Open link only when the viewer can read it — never "created" or "recovered"; a possibly committed failure shows the separate honest unknown-state page and never retries |
| `GET /groups/address-check?slug=` | advisory, same-origin and authenticated: one grammar-checked slug against core's guarded exact-slug list; closed JSON `{state, slug}` with `in_use` (a readable Group holds it), `not_seen` (none the viewer can read — never "available") or `invalid`, `slug` echoing the request; fixed `401`/`403`/`400`/`503`/`405` errors, no names, ids or counts, `no-store`, no CORS. Step 2's slug field asks it on blur (one request at a time, 5 s lifetime, cleared on every edit, applied only to the exact current value); the hint never blocks submit and the create stays authoritative |
| `GET /groups/<uuid>` | the Group: incomplete → Resume onboarding; complete → profile summary with Edit or View Business Profile per core's `can_update` hint |
| `GET /groups/<uuid>/onboarding?after=` | two explicit choices: create a new Business Profile (five fields) or select an existing one (radio list from one `ListBusinessProfiles` page, with Next); a completed Group shows its state without forms |
| `POST /groups/<uuid>/onboarding/complete` | `mode=new` or `mode=existing`; then `/groups/<uuid>?status=completed`, or `status=already_complete` on a conflict |
| `GET /business-profiles?after=` | My Business Profiles (summaries) |
| `GET /business-profiles/<uuid>`, `POST …/update` | the shared profile page: the editor or a read-only view per `can_update`, re-decided by core on submit (a refused update shows "Not permitted"); last save wins, disclosed on the page; no associations, sibling Groups or people |

- **Completion notices follow core state.** The Group page shows the `completed` and `already_complete` notices only
  when core reports the Group complete; the URL marker alone never presents an incomplete Group as complete (E-ONB-2).
- **Form budgets.** `readForm` keeps its 16 KiB default; the two profile forms pass `PROFILE_FORM_BYTES` (96 KiB),
  derived from the field maxima at 12 encoded bytes per code point. Descriptions are normalized from CRLF to LF before
  they are forwarded.
- **Refusals.** Unknown and foreign ids and a rejected origin are the fixed 404; `not_available`/`forbidden` reads show
  "Not available" (every Organization member may onboard, so the page exists for them); store failures "unavailable".
- **Workspace areas navigation.** `AppShell` renders a "Workspace areas" navigation (Groups, Business Profiles) after
  the header only when the tenant frame's indicator says the viewer is a member of the host's Organization, on every
  tenant page including branding and the tenant home. `aria-current="page"` marks only the area the page belongs to
  (none on `/account`). It is absent for anonymous visitors (M-NOAUTH) and on the generic frame of the shell projects
  (S-LANDMARKS). The links are navigation, never authority; the Primary anchor navigation is unchanged.
- **Explicit Group authority.** Core's Group answers carry `self_role`, `can_manage` and `management_source`, and
  profile summaries `can_select`/`can_update` (display hints; core re-decides every request). Organization owners and
  admins read every Group's metadata, so a readable Group the viewer does not manage shows its state and "Onboarding is
  administered by this Group's owners and admins" (`onboarding-not-administered`) instead of Resume onboarding or the
  forms (that onboarding view's Primary navigation offers only its one section, "Group onboarding"; the manager view
  keeps "Create new" and "Select existing"), and its Members link appears only for an actual Group role or a manager. The onboarding picker offers only
  `can_select` rows while the cursor and First/Next links stay the unfiltered page's. The Business Profiles index lists
  every readable profile and marks rows the viewer cannot change "Read-only". The members page of a readable Group
  without management shows a truthful notice (`members-not-administered`) instead of the fixed 404; an administrator
  through the override sees `override-admin-notice` ("… you are not a member of this Group") and no self card, Leave or
  owner controls; an actual member who also administers through it keeps the self card and sees
  `override-admin-member-notice`. `/organization/group-authority` (owner: a native radio form Off/On with the scope
  text referenced by `aria-describedby`, posting to `/organization/group-authority/action`; admin: the read-only
  effective state; anyone else: the fixed 404) is linked from `/organization/members` and `/account` for Organization
  owners and admins. A Group invitation page whose id is malformed or that core answers not_found (nonexistent, foreign
  or another recipient's) shows the one fixed sentence "Invalid invitation link. Switch to a different account, or ask
  for a new invitation link." in its frame instead of the fixed 404, and the Organization callback's
  `invitation_mismatch` notice is the same sentence; neither names the intended account or says whether the invitation
  exists. Sign-in, the accept route, the callback redirect and every other refusal or failure are unchanged. Tests:
  E-ONB-8 (`onboarding-*`), E-MEM-8/E-MEM-9/E-MEM-10 and E-MEM-3/E-MEM-5 (`membership-*`); unit
  `group-authority-forms.test.mjs`, `enrollment-client.test.mjs` and the extended client and form tests.
- **Tests.** E-ONB-1 admin creation with the suggested then edited slug, a refused field and new-profile completion;
  E-ONB-2 owner resume without false completion and existing-profile selection; E-ONB-3 both entry points to one
  shared record and the read-only member; E-ONB-4 plain-member self-service and the cross-Group, forged-request and
  Group-role-only refusals; E-ONB-5 keyboard-only and axe WCAG 2.2 AA at 1280 and 320 px; E-ONB-6 two-tab double
  submit; E-ONB-7 MCP connection rows unchanged by navigation. Unit: `onboarding-client.test.mjs`,
  `onboarding-forms.test.mjs`.

## Platform root landing and signup pages

Server-rendered, never cached, native forms with exact-Origin checks; every core call goes through
`lib/signup-client.ts` (strict answer shapes, closed refusals). The flow, cookies and limits are in
[IDENTITY.md](IDENTITY.md#platform-root-signup); which host serves which path in [DOMAINS.md](DOMAINS.md).
Pages use the entry frame with platform defaults (no tenant branding, no header Account or Sign in action) and the
existing components (`AppShell`, `Section`/`Kicker`/`Heading`/`Text`, `FormField`, `PendingSubmit`, `StatusMessage`,
`TextLink`, `Button`); no new token or component.

| Route (host) | Purpose |
|---|---|
| `GET /` (root and `www`) | the landing: product identity and one **Sign Up** link (`data-testid="root-sign-up"`) to `${signup_origin}/signup`; no sign-in, picker, directory or lookup |
| `GET /signup` (root) | the start page with one `POST /signup/start` form; closed `?status=` notices |
| `GET /signup/details` (root) | step 1 the Organization name, step 2 the editable suggested slug; for an already provisioned signup the read-only recovery view (Try again only for a matching restoration record); no credential → `/signup` |
| `POST /signup/start`, `/signup/create`, `/signup/abandon`, `GET /signup/abandon` (root) | signup login start; suggestion, creation and Try again; abandon (revokes); the cookie-clearing GET for a dead credential |
| `GET /signup/continue`, `/signup/complete` (platform tenant host), `GET /signup/prove` (root) | the three automatic continuation hops (303 only) |
| `GET /?status=tenant_created` (new tenant host) | the tenant entry with the closed "Your organization is ready." notice; no onboarding redirect |

Notices come from closed maps and never echo a submitted value; names and slugs never enter a URL. Unit:
`signup-client.test.mjs`, `signup-forms.test.mjs`, `host.test.mjs`, `session.test.mjs`; browser: `signup-*` above.
`WEB_ROUTES` gains the eight signup files.

## Labels pages

Minimal server-rendered pages in the existing shell; authority and data model in
[AUTHORIZATION.md](AUTHORIZATION.md#guarded-label-operations) and [TENANCY.md](TENANCY.md#labels). Every core
call goes through `lib/labels-client.ts` (the membership-client `call`/`readBounded`/`hasExactly`/`failure` pattern,
closed key sets `LABEL_KEYS` = `id, scope, group_id, name, visibility` and `GROUP_LABEL_KEYS` = `label_id, name`,
bounded answers, closed `?status=` maps). Forms are native POST forms (`<fieldset>`/radio, the kit `Select` through
`FormSelect`, `PendingSubmit`, hidden `csrf_token`, exact-Origin `readForm`, 303 to the closed status map); "Not permitted" on
`forbidden`, the honest unknown-state page on `unavailable`; names render as escaped text, with no colours. No new
dependency, kit module, token, `UI_FILES`/`UI_SLOTS` or B-DEPS change; chips are StateBadge-style `<span>`s.

| Route | Purpose |
|---|---|
| `GET /organization/labels` (+ `POST …/action`) | Organization-owned definitions for Organization owners/admins: list with rename and delete, one create form (name only; visibility fixed to "Organization administrators" and stated); anyone else the fixed 404. Linked from `/organization/members` and `/account` behind the existing owner/admin gates. `data-testid`: `label-list`, `label-create-form`, `label-row`, `label-rename-form`, `label-delete-form`, `labels-status`; the status `organization_context_required` reads "Deleting an Organization label requires the Organization-wide context" (never a partial deletion) |
| `GET /groups/[id]` Labels section (+ `POST /groups/[id]/labels/action`) | rendered **only** when core answers the Group's label list `ok` (Organization owner/admin): chips (`data-testid="group-label"`), an assign form (a `FormSelect` of unassigned Organization-owned definitions, the first offered by default) and per-chip remove forms. A `forbidden`/`not_found` answer renders nothing (no hidden markup, no empty heading); `unavailable` renders the existing notice without the section |
| `GET /groups/[id]/labels` (+ `POST …/action`) | Group-owned definitions. Owners/admins of the exact Group: list, create (name and visibility "Group administrators"/"Group members"), rename, visibility change and delete, with the fixed sentence "Group labels cannot be applied to anything yet: no Group-owned object type exists in this foundation" (`data-testid="labels-no-targets"`). A Group member: the read-only list of `target_members` definitions (`labels-read-only`). A readable Group without a Group role (Organization staff, override on or off): the notice `labels-not-administered` and no list. Absent, foreign or unreadable Group: the fixed 404. Linked from `/groups/[id]` only when the caller has an actual Group role, never from the override hint |

The BFF calls seven internal method/path pairs on six paths (`/internal/v1`, web BFF only, no public API;
`core/internal/labels/api`, origin/session-binding/CSRF, strict JSON, 8 KiB bodies): `GET`/`POST /internal/v1/labels`,
`POST /internal/v1/labels/{id}/update`, `POST /internal/v1/labels/{id}/delete`, `GET /internal/v1/groups/{id}/labels`,
`POST /internal/v1/groups/{id}/labels/assign` and `POST /internal/v1/groups/{id}/labels/unassign`. **No route exposes a
Group-owned assignment**: production storage has no Group-owned target type, so the Group-owned pages manage
definitions only. Unit: `labels-client.test.mjs`, `labels-forms.test.mjs`; browser: `labels-*` above. `WEB_ROUTES` gains
`groups/[id]/labels/action/route.ts`, `groups/[id]/labels/page.tsx`, `organization/labels/action/route.ts` and
`organization/labels/page.tsx`.

## Identity invariants

- The product name and domain come only from `PRODUCT_NAME` and `PLATFORM_BASE_DOMAIN` through server configuration; renaming is an `.env` change with no rebuild.
- No `NEXT_PUBLIC_` variable, `document.cookie` or `cookies()`; client components import no server configuration.
- Previews, labels and fixtures carry no identity: Organizations are addressed by UUID through `core`, never by a styling label, slug or display name. Slugs and names are display and routing conveniences, not authority.
- Infrastructure keys (`COMPOSE_PROJECT_NAME`, `DATABASE_NAME`, image and volume names) are deployment configuration, never tenant or user identity.
- The appearance preference is per browser and carries no authorization meaning.

## Status

- **Implemented:** the accessible generic shell and keyboard operation; Light, Dark and System appearance with first-paint correctness through next-themes; the unaltered kit source with fidelity checks; the saved-branding fan-out and isolated previews; bundled fonts and the blueprint drawing; Organization branding (core-validated overrides, the tenant entry, indicator, logo route and management page); My Account and AI connections; the Group onboarding pages and the member Workspace areas navigation; the platform root landing and signup pages; the Organization and Group labels pages; the browser suite.
- **Inherited and verified:** the thin BFF (`lib/config.ts`, `lib/core-client.ts`, `/api/health`, `/api/ready`, all `no-store`).
- **Known limits:** browser evidence is Chromium only; layout metrics depend on the bundled TrueType files (other faces would change measured geometry); the text-spacing trial is bounded, not a full 1.4.12 certification; the accepted kit contrast shortfalls above.
- **Deferred:** a Content Security Policy with the init-script hash; cross-tab appearance synchronization (not required).
