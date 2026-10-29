# Third-party notices (web)

This file lists third-party material shipped in the web application's source and static assets. npm
dependencies carry their own licence files in their packages; their licences are summarized in the
repository's root `THIRD_PARTY_NOTICES.md`. Lineage and provenance are documented in `docs/foundation/WEB-SHELL.md`.

## shadcn/ui

The 54 modules under `src/components/ui/` are the shadcn/ui registry's output (Base UI, Rhea style) as shipped
in the adopted brand kit, copied verbatim with the rest of the kit source under `src/` (G069; the copied tree is
verified against the kit's source manifest and is never edited in place). `src/hooks/use-mobile.ts` and
`src/lib/utils.ts` (`cn`, on `clsx` and `tailwind-merge`) are the same registry's hook and utility, and
`src/styles/rhea-leftovers.css` holds the Rhea style rules the registry transform leaves to CSS. The registry
stylesheet `shadcn/tailwind.css`, which the kit's `src/globals.css` imports, comes from the `shadcn` package, a
development dependency installed only for that stylesheet (its command-line tool is never run). The kit's
owner-authored brand modules (the eight files under `src/components/` outside `ui/`) and its other stylesheets
under `src/styles/` are first-party material composed on these components. The G002-generated, G048 and G067
predecessors of the application's components had the same shadcn/ui origin. shadcn/ui is distributed under the
following licence:

```
MIT License

Copyright (c) 2023 shadcn

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Base UI

`@base-ui/react` (G067; the primitive foundation of the kit components under `src/components/ui/`) and its
resolved dependency closure are npm dependencies under the MIT licence. Each package carries its own `LICENSE`
file. The closure as resolved in `package-lock.json`:

| Package | Version | Licence | Copyright |
|---|---|---|---|
| `@base-ui/react` | 1.9.0 | MIT | Copyright (c) 2019 Material-UI SAS |
| `@base-ui/utils` | 0.5.0 | MIT | Copyright (c) 2019 Material-UI SAS |
| `reselect` | 5.3.0 | MIT | Copyright (c) 2015-2018 Reselect Contributors |
| `@babel/runtime` | 8.0.8 | MIT | Copyright (c) 2014-present Sebastian McKenzie and other contributors |
| `@floating-ui/react-dom` | 2.1.9 | MIT | Copyright (c) 2021-present Floating UI contributors |
| `@floating-ui/dom` | 1.8.0 | MIT | Copyright (c) 2021-present Floating UI contributors |
| `@floating-ui/core` | 1.8.0 | MIT | Copyright (c) 2021-present Floating UI contributors |
| `@floating-ui/utils` | 0.2.12 | MIT | Copyright (c) 2021-present Floating UI contributors |
| `use-sync-external-store` | 1.7.0 | MIT | Copyright (c) Meta Platforms, Inc. and affiliates. |

## Kit component dependencies

The kit components' remaining npm dependencies (G069) and their resolved runtime closures are npm dependencies.
Each package carries its own licence file, which holds its copyright notice. Versions and licences as resolved in
`package-lock.json`; React, React DOM and their scheduler, and the packages already listed under Base UI
(`reselect` 5.3.0, `use-sync-external-store` 1.7.0, which `recharts` also uses), are not repeated.

| Package | Version | Licence | Used by |
|---|---|---|---|
| `clsx` | 2.1.1 | MIT | `src/lib/utils.ts` |
| `tailwind-merge` | 3.7.0 | MIT | `src/lib/utils.ts` |
| `cmdk` | 1.1.1 | MIT | `src/components/ui/command.tsx` |
| `next-themes` | 0.4.6 | MIT | `src/components/ui/sonner.tsx` and the application's theme provider |
| `sonner` | 2.0.8 | MIT | `src/components/ui/sonner.tsx` |
| `react-day-picker` | 9.14.0 | MIT | `src/components/ui/calendar.tsx` |
| `embla-carousel-react` | 8.6.0 | MIT | `src/components/ui/carousel.tsx` |
| `input-otp` | 1.5.0 | MIT | `src/components/ui/input-otp.tsx` |
| `react-resizable-panels` | 4.15.0 | MIT | `src/components/ui/resizable.tsx` |
| `recharts` | 3.10.1 | MIT | `src/components/ui/chart.tsx` |
| `shadcn` | 4.22.1 | MIT | `src/globals.css` (`shadcn/tailwind.css`); development dependency, only the stylesheet is built into the application |

`cmdk` closure. The `@radix-ui/*` packages are reached only through `cmdk`; no application or kit module imports
them.

| Package | Version | Licence |
|---|---|---|
| `@radix-ui/primitive` | 1.2.0 | MIT |
| `@radix-ui/react-compose-refs` | 1.1.5 | MIT |
| `@radix-ui/react-context` | 1.3.0 | MIT |
| `@radix-ui/react-dialog` | 1.3.1 | MIT |
| `@radix-ui/react-dismissable-layer` | 1.2.1 | MIT |
| `@radix-ui/react-focus-guards` | 1.2.0 | MIT |
| `@radix-ui/react-focus-scope` | 1.3.1 | MIT |
| `@radix-ui/react-id` | 1.1.4 | MIT |
| `@radix-ui/react-portal` | 1.2.0 | MIT |
| `@radix-ui/react-presence` | 1.1.11 | MIT |
| `@radix-ui/react-primitive` | 2.2.0 | MIT |
| `@radix-ui/react-slot` | 1.5.0 | MIT |
| `@radix-ui/react-use-callback-ref` | 1.1.5 | MIT |
| `@radix-ui/react-use-controllable-state` | 1.3.0 | MIT |
| `@radix-ui/react-use-effect-event` | 0.0.5 | MIT |
| `@radix-ui/react-use-layout-effect` | 1.1.4 | MIT |
| `aria-hidden` | 1.2.6 | MIT |
| `react-remove-scroll` | 2.7.2 | MIT |
| `react-remove-scroll-bar` | 2.3.8 | MIT |
| `react-style-singleton` | 2.2.3 | MIT |
| `use-callback-ref` | 1.3.3 | MIT |
| `use-sidecar` | 1.1.3 | MIT |
| `get-nonce` | 1.0.1 | MIT |
| `detect-node-es` | 1.1.0 | MIT |
| `tslib` | 2.8.1 | 0BSD |

`react-day-picker` and `embla-carousel-react` closures:

| Package | Version | Licence | Dependency of |
|---|---|---|---|
| `date-fns` | 4.4.0 | MIT | `react-day-picker` |
| `date-fns-jalali` | 4.1.0-0 | MIT | `react-day-picker` |
| `@date-fns/tz` | 1.5.0 | MIT | `react-day-picker` |
| `@tabby_ai/hijri-converter` | 1.0.5 | MIT | `react-day-picker` |
| `embla-carousel` | 8.6.0 | MIT | `embla-carousel-react` |
| `embla-carousel-reactive-utils` | 8.6.0 | MIT | `embla-carousel-react` |

`recharts` closure. `victory-vendor` redistributes d3 modules and is recorded as `MIT AND ISC`; the `@types/*`
packages are type declarations only.

| Package | Version | Licence |
|---|---|---|
| `@reduxjs/toolkit` | 2.13.0 | MIT |
| `@standard-schema/spec` | 1.1.0 | MIT |
| `@standard-schema/utils` | 0.3.0 | MIT |
| `d3-array` | 3.2.4 | ISC |
| `d3-color` | 3.1.0 | ISC |
| `d3-ease` | 3.0.1 | BSD-3-Clause |
| `d3-format` | 3.1.2 | ISC |
| `d3-interpolate` | 3.0.1 | ISC |
| `d3-path` | 3.1.0 | ISC |
| `d3-scale` | 4.0.2 | ISC |
| `d3-shape` | 3.2.0 | ISC |
| `d3-time` | 3.1.0 | ISC |
| `d3-time-format` | 4.1.0 | ISC |
| `d3-timer` | 3.0.1 | ISC |
| `decimal.js-light` | 2.5.1 | MIT |
| `es-toolkit` | 1.53.0 | MIT |
| `eventemitter3` | 5.0.4 | MIT |
| `immer` | 11.2.9 | MIT |
| `internmap` | 2.0.3 | ISC |
| `react-is` | 16.13.1 | MIT |
| `react-redux` | 9.3.0 | MIT |
| `redux` | 5.0.1 | MIT |
| `redux-thunk` | 3.1.0 | MIT |
| `reselect` (nested under `recharts`) | 5.2.0 | MIT |
| `tiny-invariant` | 1.3.3 | MIT |
| `victory-vendor` | 37.3.6 | MIT AND ISC |
| `@types/d3-array` | 3.2.2 | MIT |
| `@types/d3-color` | 3.1.3 | MIT |
| `@types/d3-ease` | 3.0.2 | MIT |
| `@types/d3-interpolate` | 3.0.4 | MIT |
| `@types/d3-path` | 3.1.1 | MIT |
| `@types/d3-scale` | 4.0.9 | MIT |
| `@types/d3-shape` | 3.2.0 | MIT |
| `@types/d3-time` | 3.0.4 | MIT |
| `@types/d3-timer` | 3.0.2 | MIT |
| `@types/use-sync-external-store` | 0.0.6 | MIT |

## Bundled fonts (SIL Open Font License 1.1)

The font files under `public/fonts/` are unmodified upstream files from `google/fonts` (Oswald and
Barlow Condensed at commit `7085eb89a950e85db5b166b7a58d414544b4140c`; Geist and Geist Mono at commit
`9e25e2ba265e5298f70f6182dd4e8a3ebf1b9123`). Each family's licence is shipped beside it, byte-exact,
and served with it:

| Family | Files | Licence | Copyright |
|---|---|---|---|
| Oswald | `public/fonts/oswald/oswald-wght.ttf` | `public/fonts/oswald/OFL.txt` | Copyright 2016 The Oswald Project Authors (https://github.com/googlefonts/OswaldFont) |
| Barlow Condensed | `public/fonts/barlow-condensed/barlow-condensed-400.ttf`, `public/fonts/barlow-condensed/barlow-condensed-600.ttf` | `public/fonts/barlow-condensed/OFL.txt` | Copyright 2017 The Barlow Project Authors (https://github.com/jpt/barlow) |
| Geist | `public/fonts/geist/geist-wght.ttf` | `public/fonts/geist/OFL.txt` | Copyright 2024 The Geist Project Authors (https://github.com/vercel/geist-font.git) |
| Geist Mono | `public/fonts/geist-mono/geist-mono-wght.ttf` | `public/fonts/geist-mono/OFL.txt` | Copyright 2024 The Geist Project Authors (https://github.com/vercel/geist-font.git) |

The fonts are distributed under the SIL Open Font License, Version 1.1 only, are not sold by
themselves and are not modified.

## Imagery

`public/blueprint/overview.svg` is original, project-authored geometry with no third-party source.
