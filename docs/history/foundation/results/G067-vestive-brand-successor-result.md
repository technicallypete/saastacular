> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G067 — Vestive Brand successor (Base UI/Rhea overhaul): Result (handback for Codex source review)

**Current supervisor disposition: FINAL ACCEPT at implementation merge `bb605d5b`, 2026-10-26 UTC.** See [final review](G067-final-acceptance-codex-review-20261026.md) and §15. The original handback status below is historical, superseded by §14 and §15.

Status: **attempt 1 of B5; correction round 1 (Codex review of `c34deb51`, 2026-10-25) applied and re-verified; Codex source ACCEPT of
executable tree `dbf0d98e` (review published on staging `f4f47bca`); not merged; hosted row 5 and post-merge row 6 pending.** Tested candidate
`g067-int` `c036c297ca5a138f5134720f9bc7bcad0d737d12`; corrected source `dbf0d98e9588be29d6f612bad0c89058a36875ac` (§10); source heads in §8.
Nothing here is accepted: only Codex records acceptance. Earlier status (kept): "attempt 1 active, not review-ready."
Budget B5, checkpoint after attempt 3 (BUDGET_POLICY §3, §7); zero
implementation attempts were spent before execution release (Plan r1 ACCEPT, execution released at staging `d882addc`).
Plan amendments in force: Codex-accepted parallelism amendment (§1) and Core palette delta r2 (`origin/g067-core-delta2`
`bc890b8c`, SHA-256 `3aea1880d5a9fbd6c9706976b548943e4444f6f262964c4ed7b1e1a0288c3845`; Codex review
`management/v0/results/G067-core-palette-delta-r2-codex-review-20261025.md`, ACCEPT with four binding clarifications, §5).
Plan: `management/v0/plans/G067-vestive-brand-successor-plan.md` r1 (SHA-256 `d8a33cd791b6f21d0b4c5ecd06ca340c7edf0bcf4992923d1b789347e21ac1b1`).
Integration owner and Result owner: owner A (web/harness); the coordinator runs every Docker batch and records the inventory
in the main checkout's `.local/evidence/g067-a1/inventory.tsv`.

## 1. Writer ledger (owner amendment 2026-10-25: parallelism beyond Plan §8.3's two writers; one writer per exact file)

Codex ACCEPTED the amendment on the verified native owner message `<redacted-uuid>` (2026-10-25T17:19:36Z):
"Yep, split the work as much as you can and tell Claude that you're doing that. I mean, Codex." Accepted Plan bytes are unchanged.
Every writer is an isolated Claude worktree agent, Opus 5.5 high, stack-free (no Docker, npm or build; the coordinator runs all
checks), authored as <owner>; none pushes — the coordinator pushes and reclaims finished writers with the source preserved.

| Owner | Agent | Branch (base) | Writable paths (exclusive) | Bound inputs / APIs | Status |
|---|---|---|---|---|---|
| B (intake/docs/infra-ci) | `a592cd5027167e1f2` | `g067-intake` (staging `d882addc`) → `1634476c` | `.agents/skills/vestive-brand/**`, `.claude/skills/vestive-brand`, `docs/**`, `AGENTS.md`, `CLAUDE.md`, `infra/ci/**` | home skill v0.2.1 (319 files, tree `67f92e83…`, SKILL.md `188ab915…`); V1 319 OK, V2 321 OK | handed back, reclaimed |
| A phase 1 (web/harness: P0.2, P0.3) | `afb6868d538a0dcdb` | `g067-impl` (`1b90c383`) → `213c9721` | `web/components.json`, `web/components/ui/{button,dialog,input,portal-container}.tsx`, the 16 link-site consumers, `shell-help-dialog`, `pending-submit`, token-scopes fixture, spec/support behaviour edits, fonts, `THIRD_PARTY_NOTICES.md`, `infra/validation/{run,lib}.sh`, `web/eslint.config.mjs` | `@base-ui/react` 1.9.0 (resolved in P0.1 `1b90c383`); `ButtonLink`; focus-guard attribute `data-base-ui-focus-guard` (confirm in P0.5) | handed back, reclaimed |
| C (remaining ui modules) | `a3e1343a01dbe5c36` | `g067-ui` (`1b90c383`) → `75ef4e41` | `web/components/ui/{bullet-list,layout,portal-container,select,site-header,table,textarea,typography}.tsx` | kit `ui/` + `vestive/` brand-layer sources; neutral lineage headers | handed back, reclaimed |
| F (foundation: palette, type, geometry) | `a5e4989bcfa174c1c` | `g067-found` (`g067-int` tip) → `6764f696` | `web/app/globals.css` (sections after line 640 only; lines 1–640 stay byte-identical, `GLOBALS_EJECT_SHA`) | Plan §5.1 palette (typed from `kit/src/styles/theme.css`), §5.2 type/geometry, the existing `--light-*`/`--dark-*` contract | handed back, merged |
| C1 (app routes, shell, signup) | `a9afd72ea7707e8d2` | `g067-cons1` (`g067-int` tip) → `de90b8ba` | `web/app/**` pages, `web/components/layout/{app-shell,appearance-control,brand-mark,context-indicator,shell-help-dialog}.tsx`, `web/components/signup/**` | `ButtonLink`/`Button` API (A), `brand` prop on `Heading`/`Kicker` (C); D-2/D-3 surfaces per Plan §5.3 | handed back, merged |
| C2 (feature components) | `ad868bf1ff38075e4` | `g067-cons2` (`g067-int` tip) → `cf436c75` | `web/components/{account,branding,enrollment,labels,membership,onboarding}/**` | `Select` wrapper API (`native-select-*` slots), `Textarea` auto-grow vs `rows`, `Table` slots (C) | handed back, merged |
| T1 (design oracles) | `adb8b1ea9b8ecd050` | `g067-t1` (`g067-int` tip) | `web/tests/browser/design.spec.ts`, `web/tests/browser/support/{source-reference,spacing}.ts` | measured kit values (Plan §0.1, §5.2): Button 24/28/32/36, icon 32/28; fields 32; 18px control radius, 24px dialog; Geist; 1px rules | **running** at A2's start (not yet merged; A2 does not touch its three files). Later: handed back → `be0285da`, merged into `g067-int` before `5260170b`; reclaimed |
| T2 (token/overlay/a11y oracles) | `a991254ae697adb9d` | `g067-t2` (`g067-int` tip) → `ab8e92d2` | `web/tests/browser/{tokens,overlay,accessibility,appearance,shell}.spec.ts`, `web/tests/browser/support/{modes,assets,fonts}.ts`, `web/tests/fixtures/token-scopes/page.tsx` | same measured values; Plan §7.2 | handed back, merged |
| A2 (integration, final phase G) | `a4468345789726af9` | `g067-final` (`origin/g067-int` `d6d44e02` + `origin/g067-result`) | product defects of T2's oracles in `web/components/{ui/button,ui/typography,layout/appearance-control,layout/app-shell,branding/branding-form}.tsx` and the matching `tokens.spec.ts` expectations; `infra/validation/run.sh`, `web/eslint.config.mjs`, `web/THIRD_PARTY_NOTICES.md`, `docs/{ARCHITECTURE.md,foundation/WEB-SHELL.md}`; this Result; the old-skill deletion (last commit) | all merged writers' outputs; Plan §2 item 6, §6, §8.1 | handed back (§4) → `g067-final` `e52ce172` (14 commits, last = the 81-file deletion of the old skill + alias); merged; writer branch tip and worktree reclaimed after push |
| T1b (T1 oracle follow-up) | `a2e4dae17182dc6a3` | `g067-t1b` (`g067-int` tip) → `bdee777b` | `web/tests/browser/design.spec.ts`, `web/tests/browser/support/source-reference.ts` (the §4.6 T1 lines) | §4.3 footer meta (Barlow Condensed 400, 17.28/23.328px); hover state at that time | handed back, merged, reclaimed |
| W1 (Button states, Light destructive) | `afbd6b64a4066373f` | `g067-w1` (`g067-int` tip) → `f8c846cc` | `web/components/ui/button.tsx`, `web/app/globals.css` (destructive), `web/tests/browser/support/modes.ts`, `docs/foundation/WEB-SHELL.md` | delta r2 §2.1/§2.3 + Codex clarification 1 | handed back, merged (`36541d1c`), reclaimed |
| Core (platform palette sync) | `a68461be6e15a1cc7` | `g067-core` (`g067-int` tip) → `2f4e2fae` | `core/internal/branding/colors.go`, `core/internal/branding/branding_test.go` | delta r1 §1–§4 as amended by r2; pins computed with Core's own math | handed back, merged (`6578540c`), reclaimed |
| W3 (Dark `text-2`) | `a02af78f910e86ef1` | `g067-w3` (`g067-int` tip) → `fe731203` | `web/app/globals.css`, `web/tests/browser/support/modes.ts`, `core/internal/branding/colors.go`, `docs/foundation/WEB-SHELL.md` (the `text-2` value only) | p07 S-CONTRAST worst-case finding (§2) | handed back, merged (`ad7319ba`), reclaimed |
| W2 (oracles per delta r2) | `aad3bfc3cfd143910` | `g067-w2` (`g067-int` tip) → `f6e91065` | `web/tests/browser/{tokens,design}.spec.ts`, `web/tests/browser/support/{contrast,color}.ts`, `web/tests/fixtures/token-scopes/fixtures.ts` | delta r2 §3.2 + Codex clarifications 1–4 | handed back, merged (`c036c297`), reclaimed |
| Core delta r1 planner (Fable) | `a51c49d6fe7207721` | `g067-core-delta` → `00927899` (SHA-256 `b551f131…`) | the delta file only | Plan r1; Core sources | handed back; Codex CORRECTION LOOP |
| Core delta r2 planner (Fable) | `a9154a03731be9e05` | `g067-core-delta2` → `bc890b8c` (SHA-256 `3aea1880…3845`) | the delta file only | Codex r1 review | handed back; Codex ACCEPT with clarifications |

All writers above: isolated, stack-free, authored as <owner>; Opus 5.5 high except the two Fable planners. The coordinator
pushed each branch; finished writer worktrees and writer-branch tips were reclaimed after push; the harness auto-branches
`worktree-agent-*` are held (inventory). Writer exclusivity held for every lane (one writer per exact file at a time; W1 and W3
touched the same files sequentially, W3 after W1's merge; W3 edited `colors.go` after the Core merge). Integration merges on
`g067-int` (first parent): T1b `bdee777b` and A2 `e52ce172` fast-forward; W1 `36541d1c`, Core `6578540c`, W3 `ad7319ba`, W2
`c036c297` (the tested candidate); staging `cb034355` merged in `83f66f0c`.

Rules for every later writer: exactly the files above; an edit outside them is a handback finding, not an edit. A integrates all
branches (disjoint paths), owns this Result and the final commit; extra writers do not authorize extra stacks, a third run,
parallel builds or any change to the owner's preview (`m1-preview-20261024`).

## 2. Coordinator-run batches (attempt 1)

| When (UTC) | What | Outcome |
|---|---|---|
| 2026-10-25 17:3x | P0.1 dependency resolution in a pinned, labelled, disposable `node:24-slim` container (image `sha256:0e0ff40c…`, run `p01`, volume removed) | lockfile +9 (exactly the expected `@base-ui/react` 1.9.0 closure, all MIT), −24 (16 `@radix-ui/*` + 8 non-scoped), 0 changed; `@radix-ui`/`radix-ui` in `package.json` 0, lock 0; `npm ls --all` radix count 0. Commit `1b90c383`. |
| 17:43–17:51 | Row 1 `lint lint-web controls` on integrated tree `2c5d7a8d` | FAIL: B-WEB-COMPONENTS, B-WEB-ASSET-CONTROLS A1–A19 and the B-REQUIRED-CHECKS rollup, one cause: `UI_SLOTS` lacked the slots owner C's table/select emit. All other rows PASS, including B-DEPS, B-DEPS-CONTROLS C1–C19, B-SECRET-SOURCE (real gitleaks) and B-SECRET-CANARY. |
| 17:51–17:52 | Row 2 `unit-web` on `2c5d7a8d` | PASS: `web-check` (eslint --max-warnings=0, `tsc --noEmit`, unit tests) + secret-scan. First whole-app type check clean. |
| (coordinator) | Pin fix `df6301ff` (UI_SLOTS); row 1 rerun on `df6301ff` | PASS (19 required IDs), coordinator-reported. |
| (coordinator) | Row 2 on `2c5d7a8d` (above) | PASS, coordinator-reported. |
| (coordinator) | Run p06 on the integrated tree `5260170b` (all writers incl. T1; before A2's fixes) | rows 1–2 PASS; P0.5 browser 65 passed / 37 failed, coordinator-reported. A2's follow-up in §4.6. |
| (coordinator) | P0.5 restricted browser proof, first run on `df6301ff` (before F/C1/C2/T2 were merged) | 62 passed / 40 failed, coordinator-reported; the failures are the old oracles and the old palette, e.g. S-FONTS B0 (4 of 5 font files requested), S-A11Y `color-contrast` on `#base-ui-_R_…` nodes. Not yet rerun on the integrated tree. A2 ran no check (§4.7). |
| 2026-10-25 (row 1 on `5260170b`) | Row 1 `lint lint-web controls`, base `d882addc`, allow-file `E/allow.txt` (`static-3`, `row1c.log`) | PASS. (Row-1 chain so far: FAIL `2c5d7a8d`, PASS `df6301ff`, PASS `5260170b`.) |
| (row 2 on `5260170b`) | Row 2 `unit-web` (`unit-web-3`) | PASS. |
| 18:32–18:40 | Row 1 on A2's `e52ce172` (`static-4`, `row1d.log`) | PASS. |
| (row 2 on `e52ce172`) | Row 2 `unit-web` (`unit-web-4`) | PASS. |
| (p07, after `bdee777b`) | Row 3 browser p07 on `bdee777b`, restricted spec/project set (private project `g067br-<run>`, no published port, `COMPOSE_PARALLEL_LIMIT=1`; project removed) | 99 passed / 3 failed: Dark S-CONTRAST local worst-case lead `#cbcac9` over `#5b5552` = 4.48 (< 4.5); two fixture-A Light destructive findings (`#eef5fa` on `#e7000b` 4.333; axe). Resolved by W3 (§6 item 3) and W1 (§6 item 2). Log `p05-browser-p07.log`. |
| (S1a, after `2f4e2fae`) | Core S1 (delta r2 §4.1; private, `network_mode: none`, project `g067core-<run>`; worktree, env and image removed) on Core commit `2f4e2fae`, run `s1a` | PASS: gofmt failing guard silent, `go vet ./internal/branding/` clean, `TestOverrideGrammar`, `TestDefaultsPass`, `TestContrastEdges`, `TestValidate` PASS (`s1-s1a.log`). |
| 18:58 | **Row 1 on the candidate `c036c297`** (`static-6`, `row1f.log`, driver `g067-row12f.out`) | **FAIL**: only `B-SCOPE` ("outside the allow-list: core/internal/branding/branding_test.go core/internal/branding/colors.go") and the `B-REQUIRED-CHECKS` rollup that follows; every other row PASS. Cause: the Core paths of accepted delta r2 §5 were not yet in the allow-file. Failed evidence preserved. |
| 18:59 | **Row 2 on `c036c297`** (`unit-web-6`, `row2f.log`) | **FAIL** of the same two rows only (B-SCOPE + rollup; `unit-web` itself passed). Failed evidence preserved. |
| 18:59 | Core S1 on the candidate `c036c297`, run `s1b` | PASS (same four tests, guard silent, vet clean; `s1-s1b.log`). Core pins: T8 ×1.12 4.6933, T10 ×1.12 on `#d2000a` 4.6213, N3 3.7135, grey-primary ×1.12 4.5682 (window 4.52–4.62). |
| 18:59–19:05 | Row 3 browser p08 on `c036c297` (same restricted set as p07) | 102 / 102 passed (`p05-browser-p08.log`). |
| (correction) | Allow-file: one line added, `^core/internal/branding/(colors\|branding_test)\.go$` (pipe escaped for the table) — exactly delta r2 §5; no other allow-list change | — |
| 19:12:11–19:19:37 | **Row 1 rerun on `c036c297`** (`static-7`, `row1g.log`) | **PASS** (every row incl. B-SCOPE, B-SECRET-CANARY, B-REQUIRED-CHECKS). |
| ended 19:20:15 | **Row 2 rerun on `c036c297`** (`unit-web-7`, `row2g.log`) | **PASS.** Rows 1–2 are qualified on the candidate only by these reruns. |
| 19:20:50–19:25:11 | **Row 3 final p09 on `c036c297`: full row 3** (all six projects shell-light, shell-dark, tokens-light, tokens-dark, appearance, dev-hydration; every spec; no `--grep`), private project, no published port, removed after the run | **146 passed, 0 failed** (`p09-browser-p09.log`). |
| 19:25:11–19:25:30 | **Row 3r renamed identity** on the same private project: private env `PRODUCT_NAME="Renamed Example"`, `PLATFORM_BASE_DOMAIN=renamed.example.net`, `CUSTOM_DOMAIN_CNAME_TARGET=custom.renamed.example.net`; `up -d --no-build --wait`; `--grep @identity` on shell-light + shell-dark | **2 passed**; the 7 image IDs unchanged before/after (`p09-images-before-p09.txt` = `p09-images-after-p09.txt`). Evidence line only, **not** a B-BROWSER-RENAME/B-RENAME record (`p09-browser-3r-p09.log`). |
| 19:34:19–19:35:31 | **Run p10 (evidence export)**: focused `tokens.spec.ts --grep S-CONTRAST`, tokens-light + tokens-dark, JSON reporter, on `83f66f0c`; private project, removed after the run | **2 passed**; the grey-primary `contrast.json` rows are exported (`p10-browser-p10.json`, `p10-grey-primary-contrast.txt`; numbers in §5). |
| 19:25:55–19:26:22 | **Row 4** `validate-controls.sh --only F-vestive-css,F-vestive-kit,F-vestive-md,F-vestive-link,F-agent-claude-skill,F-agent-agents-skill` on `c036c297` | **6 passed, 0 failed** of 6 selected (529 defined) (`f-controls.log`, `row4.log`). Needs the Docker daemon (not host-only as the Plan stated). |

Evidence directory `E` = main checkout `.local/evidence/g067-a1/` (ignored files); driver stdout in the coordinator's job
scratch. After every browser run the host resource counts returned to the pre-run baseline (26 containers, 14 volumes, 9
networks, 60 images). p06 (`5260170b`) causes are in §4.6; p05 is the earlier restricted proof row above.

## 3. Escalations and open items carried

E1 (Base UI closure) resolved by P0.1; E2 accepted by Codex (actual owner copy/adaptation authority, no G048 extension); E5 secret-scan of the
vendored bundle resolved by row 1 B-SECRET-SOURCE PASS; the Plan's row 4 `validate-controls.sh` needs the Docker daemon even for the F-only
selection (not host-only as the Plan stated) and runs after integration. Duo status-wording clarification for `docs/ARCHITECTURE.md` is
carried to the integration step.

Resolution (2026-10-25): row 4 ran after integration on `c036c297`, 6/6 PASS (§2). Duo status wording applied by A2 in
`docs/ARCHITECTURE.md` §61 (§4.5). E6 (Base UI guarantee gap) and E7 (scope): no stop was needed; every retained browser
scenario passes on the candidate (p09 146/146) and no consumer needed a behavioural change. The P0.5 "not yet rerun on the
integrated tree" row is superseded by p06–p09 (§2).
Codex decisions recorded during attempt 1: parallelism amendment (owner message `e7e40bb6`, §1); bounded Core palette sync
inside G067; delta r1 CORRECTION LOOP (tenant-override `#696969` hover 4.4151 vs Core 4.5682; S1 gofmt guard must fail);
delta r2 ACCEPT with four binding clarifications — all four applied (§5).

Resources: inventory in `.local/evidence/g067-a1/inventory.tsv`; cleanup receipt follows at closeout (not yet produced).

## 4. Integration owner A2 (final phase G): fixes, adaptations and open items

A2 (`a4468345789726af9`) wrote on `g067-final` = `origin/g067-int` `d6d44e027d21b141d8f7b1eb47f455cd4fddd16e` merged with
`origin/g067-result` (the Plan on `origin/staging` verified at SHA-256 `d8a33cd7…`). Stack-free: no Docker, npm, build or test;
every ratio below is computed (WCAG 2.x relative luminance, sRGB compositing of `bg-<token>/<n>` over the scope ground) from the typed
hex tokens of `web/app/globals.css` and `web/tests/fixtures/token-scopes/fixtures.ts`, for the six distinct fixture scopes
(default, A, B, light-only, dark-only-in-A, branded; default-adjacent equals default) in both appearances. **None of these fixes has
been run yet; the coordinator's next rows 1–3 are the evidence.**

### 4.1 Product defects (Button, appearance control)

| # | Change | Before | After |
|---|---|---|---|
| 1a | **PROVISIONAL** (coordinator relay of Codex, 2026-10-25: an alpha hover can render a tenant override, e.g. `#696969` white-on-primary, below 4.5; kept isolated in its own commit "provisional: hover states" so a later correction can replace it with whole-element brightness states, as core's ×1.12/×0.90 state model assumes). `default` hover `hover:bg-primary/80` → **`/90`** (press shares the hover fill: the variant has no separate pressed colour, only `translate-y-px`; focus keeps the rest fill and the 2px `ring` outline) | label on hover fill, Light: default **3.94**, A 4.63, B 4.88, light-only **4.25**, branded **4.25**; Dark: default 7.63, A 5.04, B 5.97, dark-only-in-A 5.98, branded 4.81 | Light: default **4.71**, A 5.77, B 5.98, light-only 5.23, branded 5.21; Dark: default 6.54, A 6.02, B 7.23, dark-only-in-A 7.21, branded 5.86 (min 4.71). `/85` would fail (4.29 Light default). Rest 5.65–8.70 unchanged; ring vs ground ≥ 5.83 |
| 1b | `link` `text-primary` → `text-primary … dark:text-accent` (Light unchanged) | Dark: default and light-only **3.50** (A 6.93, B 8.70, dark-only-in-A 8.65, branded 7.60) | Dark accent: default 13.35, A 10.05, B 10.57, light-only 13.35, dark-only-in-A 11.40, branded 11.52; Light primary 5.65 / 6.48 / 6.65 / 5.92 / 6.48 / 5.83. Accent is the app's link ink (`TextLink`) and core's T5 text pair; primary-as-text has no core pair |
| 1c | `outline` `border-border` → **`border-input`** | boundary vs ground 1.13–1.40 | Light 3.71 / 3.37 / 3.38 / 3.40 / 3.37 / 6.37; Dark 4.23 / 3.91 / 3.94 / 4.23 / 3.83 / 7.40 (min 3.37; core N3 pair) |
| 1d | `secondary` keeps the kit's neutral fill and adds **`border-input`** (1px; `ButtonLink` shares `buttonVariants`, so its secondary path is the same class) | fill vs ground 1.02–1.30 (branded 6.52 / 8.50) | boundary = 1c's values (min 3.37); label on fill unchanged 15.62 / 14.62 (branded 7.17 / 8.50) |
| 1e | `cta`: compound variant `px-4` for `xs`/`sm`/default (was the size's `px-3`) | at 32px height the label line box (≈ 20.7px, top ≈ 5.7px) must satisfy p ≥ 0.107·w, i.e. p ≥ 0.136·L for label width L: `px-3` covers L ≤ ≈ 88px, about T2's 0.2px shortfall for "cta action" | `px-4` covers L ≤ ≈ 118px (the fixture label, "Sign in", "Signing in…", "Sign Up"). Contrast unchanged: label on `::before` fill at rest/hover (`brightness-110`)/press (`brightness-90`) min 4.83 |
| 1f | `appearance-control.tsx` `px-2.5text-foreground` → `px-2.5 text-foreground` | invalid class (no padding, no colour) | intended |
| 1g | destructive as text on `muted`: **no change** | Light 4.24 (every non-branded Light scope), Dark 5.28 | the only rendered `text-destructive` text is `onboarding/form-field.tsx`'s field error, rendered on the page ground (New Group, profile fields, signup details: `ground` sections) = 4.77 Light; no rendered destructive text sits on a muted fill (the muted `existing-profile` section has no field error). Invalid fields' `border-destructive` on the muted field fill is non-text 4.24 ≥ 3. Recorded for Codex (§4.4 Q3) |

Resolution (2026-10-25): 1a was replaced as planned. Under delta r2 §2.1 and Codex clarification 1, W1 (`f8c846cc`) set every filled
variant (`default`, `secondary`, `destructive`, `cta`) to whole-element `hover:brightness-[1.12] active:brightness-90` (cta
1.10 → 1.12); no alpha/color-mix hover remains on a filled variant. 1d's added `border-input` on `secondary` was reverted to
the kit form by W1 once both filled-button boundary oracles were removed (§6 item 4). 1b, 1c, 1e, 1f stand. 1g unchanged
(§6 item 11). Rendered outcome: p09 S-CONTRAST/S-STATES PASS in every project (§2).

### 4.2 Plan-vs-oracle adaptations (for Codex review)

- **Secondary boundary (1d).** Plan §5.1 makes `secondary` the kit's neutral fill (1.13:1 Light, 1.30:1 Dark against the
  ground); the retained S-CONTRAST oracle requires a ≥ 3:1 fill or boundary for a button. The smallest fix keeping the kit fill
  is a 1px `border-input` boundary. `tokens.spec.ts` S-CONTRAST now measures `"<scope> secondary boundary"` (the
  `border-top-color`, threshold 3, the same form as the outline boundary) instead of `"<scope> secondary fill against the scope"`;
  the `default` fill check is unchanged.
- **S-STATES expectations follow the corrected variants** (`tokens.spec.ts` `lookOf`): default hover `primary/90` (provisional, in the same isolated commit as 1a); outline and
  secondary border `input`; link label `accent` in Dark. No assertion was removed or loosened.
- The outline (1c), link (1b) and hover (1a) changes are component fixes against unchanged oracles.

Resolution (2026-10-25): the secondary-boundary adaptation is withdrawn. Codex clarification 3 (G017 N1) removed both the `default` and the
`secondary` filled-button fill/boundary assertions in `tokens.spec.ts` (W2 `f6e91065`); S-STATES now expects the uniform
1.12/0.90 filters on all four filled variants, outline border `input`, Dark link `accent`. One assertion of the same type
remained in the shell spec (`accessibility.spec.ts:105`, §6 item 11) until the correction round removed it (§10).

### 4.3 Fonts (S-FONTS: four of five files requested)

The kit's brand layer uses Barlow Condensed **400** in exactly one brand-layer style: `kit/src/components/vestive/typography.tsx`
`textVariants.meta` = `font-label text-[1.08rem] leading-[1.35] text-muted-foreground` (also in the marketing-only `content.tsx`
Stat/Quote/PlanCard, not adopted). Owner C moved `meta` to the sans stack for everyday hints, so nothing on `/` requested
`barlow-condensed-400.ttf`. A2 restored that use as a brand-layer option: `Text` gains `brand` (meta only →
the kit's style), and the shell footer identity line (`app-shell.tsx`: the platform domain, or the platform attribution on a tenant
host) uses `<Text variant="meta" brand>` — the G048 consumer, and the existing `"footer meta"` row of `support/source-reference.ts`
already expects exactly that style (17.28px, label stack, 400). All five pinned faces now have a consumer on `/`; no asset removed.
This extends D-2's brand-moment list by the footer identity line (recorded in WEB-SHELL.md; Codex may redirect: the alternative
is to drop the 400 face and re-pin to four faces, which contradicts the Plan's five-face S-FONTS/AC2).

Resolution (2026-10-25): T1b (`bdee777b`) aligned the `"footer meta"` oracle rows; S-FONTS (five pinned responses) and B0 pass in p09.
The brand-moment extension itself stays **open for Codex** (§6 item 5).

### 4.4 Open questions and findings for Codex

- **Q1 — Light `destructive` (new finding, blocks a green S-CONTRAST/S-STATES in `tokens-light`).** The destructive variant's
  label is `text-background` (core's T10 pair "background on destructive"). With the kit's Light `#e7000b`, partial-override
  fixture scopes that set an off-white ground but inherit the platform destructive measure **A 4.33, B 4.34, light-only 4.37,
  dark-only-in-A 4.33** at rest (hover `/90`: 4.16 / 4.14 / 4.19 / 4.16); the default scope is 4.77 (hover 4.54). Core's own
  T10 under its ×1.12 state factor is **3.99** on the new platform palette (`core/internal/branding/colors.go` CheckContrast
  model; the only failing core pair, computed for every T/N pair in both appearances). No component-only token swap is
  core-consistent, so A2 did not change it. Smallest proposed fix (a Plan §5.1 derivation like `muted-foreground`): Light
  destructive **`#d4000a`** — rest ≥ 5.02 on every Light fixture ground, `/90` hover ≥ 4.69, core T10 at 1, 1.12, 0.90 = 4.56;
  it touches `globals.css`, `support/modes.ts`, the WEB-SHELL palette row, and core's palette (the pending Codex sync delta).
  Resolution (2026-10-25): Codex accepted Light destructive **`#d2000a`** (delta r2 §2.3; T10 5.6079 / 4.6213 / 5.2872), applied by W1 in
  `globals.css`, `support/modes.ts`, WEB-SHELL and by Core in `colors.go`; fixture A tinted label ≥ 4.63 at every state (§6 item 2).
- **Q2 — Core palette sync.** `core/internal/branding/colors.go` `platform` still holds the G048 palette (pending Codex review;
  not A2's to edit). Its sync must carry Q1's decision.
  Resolution (2026-10-25): Core writer (`2f4e2fae`) synced `colors.go`/`branding_test.go` under delta r2; S1 PASS (§2, §5).
- **Q3 — destructive on muted** (1g): 4.24 Light, not in the Plan table; no rendered pair below 4.5 today; a future muted-band
  error message would need the Q1 derivation (`#d4000a` on `#f3f1f1` ≈ 4.71).
  Update (2026-10-25): the 4.24 figure is for the kit `#e7000b`; with the accepted `#d2000a` destructive on `muted` is 4.98
  (delta r2 §2.3). Still not rendered anywhere; left open for Codex (§6 item 11).
- **Q4 — fonts** (§4.3): footer identity line as a brand moment vs. four faces. Open for Codex (§6 item 5).
- **Q5 — copy.** `TOKEN_LABELS.border` is now "Borders and dividers" (the `muted` button variant is gone; no test or core pin).
  The branding page's core-pair description `T7: "text on muted actions (the border colour)"` (`app/organization/branding/page.tsx`)
  is left: it describes core's T7 (`ink` on `border`), whose disposition belongs to the core sync.
  `infra/validation/render-customizations.tsv:27` still describes the ESLint ignore as "the relocated … skill (G049)" (description
  column, outside the G067 allow-file; no path).
  Still open at the candidate (both texts unchanged; §6 item 11).
- **For T1 / the coordinator:** T1's files (`design.spec.ts`, `support/source-reference.ts`, `support/spacing.ts`) were not
  touched, except the one amended line `design.spec.ts:292` (§4.6). Their final values must agree with: default hover `/90`; outline and secondary 1px `border-input`; link Dark `accent`;
  `cta` padding 16px at the default size (`px-4`); footer meta in Barlow Condensed 400 (`font-label`, 17.28px / 23.328px).
  Resolution (2026-10-25): T1 merged (`be0285da`); T1b and W2 aligned the T1 lines; the default hover is now brightness 1.12, not `/90`,
  and secondary has no added border (§4.1 resolution).

### 4.5 Guards and documentation (Plan §2 item 6, §6)

- `infra/validation/run.sh` (owner B's list; line numbers on this tree): 979 `VESTIVE_SKILL`/`VESTIVE_ALIAS` → `vestive-brand`;
  981–985 `G049_ESLINT_LINES` → `G067_ESLINT_LINES` (the two new `web/eslint.config.mjs` lines 8–9, each exactly once) and its uses
  in `eslint_literal_text`, A15, A17; 1190 DESIGN_KIT_COPY set = `kit/public/brand` + `design-system/assets`; 1275 `kit=` the
  `kit/` root; 1287 A1 copies `kit/src/components/vestive/ribbon.tsx` (still `data-slot="ribbon"`); 1289 A2 `kit/public/brand/logo.png`
  (exists); 1322 A16 glob `vestive-brand`; 1644 `SKILL_LINK`/`SKILL_LINK_TARGET` → `vestive-brand`. `SKILL_MOVES` (1630–1634) and
  the 1684 reversal clause stay as inert history. `bash -n` clean.
- `web/THIRD_PARTY_NOTICES.md`: Radix section → Base UI section with the resolved MIT closure (9 packages; copyright lines from
  each package's published `LICENSE`, read-only lookup); shadcn/ui paragraph names the registry `base-rhea` origin.
- `docs/foundation/WEB-SHELL.md`: current non-colour token values (G048 values noted), D-2 footer identity line, the button
  contrast adaptations, the 10-entry `WEB_PUBLIC_PINS` list and font consumers, the final `UI_SLOTS` list and the
  `UI_BUTTON_BASEUI`/`UI_DIALOG_BASEUI`/`UI_DIALOG_PORTAL` codes, the resolved closure (+9/−24) and the `devOptional` note
  (`@types/react` is an optional peer of `@base-ui/react`). Run outcomes (rendered palette verification, pin/S-FONTS outcomes,
  zero-Radix proof outcomes) stay "filled from the G067 Result at closeout".
- `docs/ARCHITECTURE.md` §61: the G067 record states "implemented in the G067 change and under Codex review; acceptance is
  recorded only by Codex in the G067 Result and `PROGRESS.md`" (Duo status-wording clarification).
- **Last commit:** `git rm -r .agents/skills/vestive-design` (81 files) and the alias `.claude/skills/vestive-design` (Plan §8.1).
  Remaining `vestive-design` / `components/ui/ribbon` mentions outside management records, all history or controls:
  `ADOPTION.md:274` (immutable commit citation), the upstream `SKILL.md:22` (the owner's archive file name, upstream byte),
  `docs/ARCHITECTURE.md` §61 (two history citations), `docs/foundation/OPERATIONS.md:820` (historical names),
  `docs/foundation/WEB-SHELL.md:150,152` (pre-G067 record), `run.sh` 1288 (A1 plants a copy at `web/components/ui/ribbon.tsx`
  as a negative control), `run.sh` 1632–1634 and 1684 (inert G049 relocation history). Codex-owned records listed in Plan §3.3 are
  untouched.

### 4.6 Run p06 findings (coordinator relay; A2's ledger amended to these exact lines)

- **Duplicate fixture testid.** `tests/fixtures/token-scopes/page.tsx:97` `ButtonLink data-testid="button-link"` collided with the
  link-variant Button (`button-${variant}`), giving strict-mode "found 2" in S-ANGLE/S-STATES/S-TEXT-SPACING (tokens-light/dark).
  Renamed to `button-as-link`; `tokens.spec.ts` 945 (`getByTestId`) and 972 (the ButtonLink spacing item) follow; variant-link
  usages keep `button-link`.
- **S-PALETTE "appearance frame"** (`design.spec.ts:292`): the left-edge midpoint of the 18px pill is the curve apex and
  anti-aliased (measured `#9a8d85` vs `#908279` Light, `#726863` vs `#7d726d` Dark). Now sampled on the straight top rule at
  mid-width, `Math.round(frame.y) + 0.5` (the device pixel the 1px border is painted on); expectation `src("input")` and
  tolerance unchanged. The only `design.spec.ts` edit.
- **Fonts** (B0 "exactly 5 pinned font responses" and the hold-capture S-FONTS tests): resolved by §4.3 (footer identity line
  in the kit's brand meta requests `barlow-condensed-400.ttf`); no asset removed.
- **axe `color-contrast` on `#base-ui-_R_…` (shell, Light).** That node is the help trigger (`DialogTrigger render={<Button
  size="sm"/>}`); S-A11Y clicks it to open the dialog and audits with the pointer still over it, so axe measured the hover fill
  `primary/80` (label 3.94:1). The provisional `/90` hover (1a) gives 4.71:1; a later whole-element brightness state must keep
  ≥ 4.5 (×1.12 on `#0062d6` = 4.68:1, core's T8 model).
- **Dark `default outline boundary #242222 against #0c0a09`**: fixed by 1c (`border-input`, 4.23:1).
- **HELD (not touched):** the Light destructive fill token (axe flags `.bg-destructive.text-background` on the fixtures; Q1) and
  any hover/pressed work beyond the isolated commit "provisional: hover states".
  Resolution (2026-10-25): both released by delta r2 and applied by W1 (`#d2000a`, brightness states); p08/p09 axe and S-CONTRAST PASS.
- **T1 lines that now disagree with the components (A2 may not edit them; coordinator/T1 to apply or redirect):**
  1. `support/source-reference.ts` "footer meta" row (157–164): the footer identity line is the kit's brand meta (§4.3):
     `cite` "kit brand-layer typography.tsx:textVariants meta font-label text-[1.08rem] leading-[1.35] text-muted-foreground",
     `px: { "font-size": 17.28, "line-height": 23.328 }`, `stack: "label"` (weight 400, colour unchanged).
  2. `design.spec.ts:537` `FONT_NODES` "footer meta": `family: "Barlow Condensed", weight: "400"`.
  3. `design.spec.ts` 409, 429–446 S-FIDELITY "help trigger (hover)": expects the kit's `primary/80`; with the provisional hover
     it is `primary/90` (`color-mix(in oklab, … 90%, transparent)`, cite `hover:bg-primary/90`). If Codex replaces the hover
     with brightness states, this row changes again with that correction.

Resolution (2026-10-25): lines 1–2 applied by T1b (`bdee777b`); line 3 first aligned by T1b to the provisional hover and then to the
brightness state by W2 (`f6e91065`, `design.spec.ts`). The duplicate testid, appearance-frame sampling and Dark outline
boundary fixes held (p09 PASS).

### 4.7 What A2 did not do

No Docker, compose, npm/npx, build or test; nothing pushed, no MR; the owner's preview `m1-preview-20261024` and every other
worktree untouched; T1's three files untouched except the amended `design.spec.ts:292` (§4.6); no check is claimed as passing.

## 5. Core palette sync and G017 states (delta r2 §6 Result section)

Authority: bounded Core palette sync inside G067 (Codex); delta r1 CORRECTION LOOP; delta r2 ACCEPT with binding clarifications
(`management/v0/results/G067-core-palette-delta-r2-codex-review-20261025.md`). Application of the four clarifications:

1. **Uniform 1.12/0.90** on every filled variant incl. `secondary` (r2 §2.1's Light-only secondary 0.90 hover is in neither
   implementation nor oracle): `web/components/ui/button.tsx` (W1 `f8c846cc`), S-STATES/S-CONTRAST expectations (W2 `f6e91065`).
   Applied.
2. **Filter-aware contrast oracle**: `web/tests/browser/support/contrast.ts` gains `brightnessFactors` and derives the effective
   label colour under the element's computed `brightness()`; a separate assertion keeps the computed `color` equal to the
   unscaled semantic token; hover and pressed labels both measured (W2). Applied. Known limitation: the helper throws on any
   non-brightness filter on the measured element or its ancestors.
3. **Both filled-button boundary checks removed** (`default` and `secondary` fill-versus-surface in `tokens.spec.ts`, G017 N1);
   outline/input boundary and focus-ring checks kept (W2). Applied. One shell assertion of the same type remains (§6 item 11).
4. **Fixture-coverage wording**: the Light-only `grey-primary` scope is iterated by S-TOKENS, S-CONTRAST, S-STATES (and also by
   S-PORTAL, S-ANGLE, S-TEXT-SPACING, S-FONTS, S-BLUEPRINT); S-A11Y covers it only as part of the whole fixture page, **not** in
   its dialog-open loop. Applied (this wording).

Core map diff (`core/internal/branding/colors.go` `platform`, Core `2f4e2fae` + W3's Dark `text-2`): every token synced to the
successor `--light-*`/`--dark-*` values of `web/app/globals.css`, e.g. Light `background #f4f2ee → #ffffff`, `secondary #526e26 →
#f3f1f1`, `destructive #af3028 → #d2000a`, `input #615b55 → #908279`; Dark `background #0d0d0d → #0c0a09`, `destructive #ff6b5e →
#ff6467`, `input #8f8a84 → #7d726d`, `text-2 #c9c5c0 → #d1d0cf`; `primary #0062d6`, `ring` and `accent` unchanged. Core pair
semantics, factor set, pair ids, thresholds, error fields, API and schema unchanged. `branding_test.go` additions: T10 ×1.12 pin,
N3 pin, accepted rows grey-primary `#696969`, Light `input #949494`, Dark sparse `popover-foreground #bdbdbd`, new refusal Dark
`accent #8a8a8a` (`contrast_failed/T5:dark`), the grey-primary ×1.12 window, updated defaults (`ink`; N2 `ring` equal to `muted`).
All pins computed by python3 with Core's own `rgb`/`scale`/`luminance`/`Contrast` math.

Pins: T8 ×1.12 **4.6933**, T10 ×1.12 on `#d2000a` **4.6213**, N3 **3.7135**, grey-primary ×1.12 **4.5682** (window 4.52–4.62).
S1 PASS on `2f4e2fae` (s1a) and on the candidate `c036c297` (s1b) (§2).

S-STATES filter expectations: `default`, `secondary`, `destructive`, `cta` hover `brightness(1.12)`, pressed `brightness(0.9)`.
Row 3 p09 S-STATES and S-CONTRAST PASS in tokens-light, tokens-dark, shell-light and shell-dark.

U1/U3 (rounding residual, Core float vs 8-bit canvas ≤ ~0.03): measured. The `contrast.json` attachment of the passing
S-CONTRAST row-3 test is exported by the focused run p10 (below; JSON reporter, rows in `p10-grey-primary-contrast.txt`).
`grey-primary` (Light `{primary #696969, primary-foreground #ffffff}`), painted values from the browser:

| State | Label on fill | Measured | Core model (float) | Note |
|---|---|---|---|---|
| default at rest | `#ffffff` on `#696969` | **5.49** | 5.4898 | equal |
| default hover (×1.12) | `#ffffff` on `#767676` | **4.54** | 4.5682 | 8-bit residual −0.03: the computed `(117.6 → 118)` fill gives 4.5422; above 4.5, the U1 stop condition did not occur |
| default pressed (×0.90) | `#e6e6e6` on `#5e5e5e` | **5.20** | 5.1318 | label and fill both scaled ×0.90 and rounded to 8 bits; measured +0.07 above the float model (higher, so no threshold risk) |
| destructive rest / hover / pressed (`#d2000a`) | white-ish on `#d2000a` / `#eb000b` / `#bd0009` | **5.61 / 4.63 / 5.31** | 5.6079 / 4.6213 / 5.2872 | all ≥ 4.5 |
| secondary rest / hover / pressed | `#1d1816` on `#f3f1f1`; `#201b19` on `#ffffff`; `#1a1614` on `#dbd9d9` | **15.62 / 17.04 / 12.78** | — | uniform ×1.12/×0.90; the Light hover fill clamps to `#ffffff` (the page), accepted by Codex's uniform-factor clarification |

The counterexample behaves as r2 §2.2 predicted: the whole-element ×1.12 state paints 4.54 (an alpha `/90` hover would have painted
4.4151). The hover residual is within one 8-bit step per channel (U1/U3); the pressed difference is positive (safe side). The p10 run: `tokens.spec.ts --grep S-CONTRAST` on
tokens-light and tokens-dark, **2 passed**, on integration tip `83f66f0c` (source identical to the tested candidate `c036c297`;
management-only merge), 19:34–19:35 UTC, private project `g067br-p10`, no published port, removed after the run, resource counts back
to 26/14/9/60 (`p10-browser-p10.json`, `p10-grey-primary-contrast.txt`).

## 6. Adaptations and decisions for Codex

| # | Adaptation / decision | Reason | Files | Status |
|---|---|---|---|---|
| 1 | Hover/pressed: whole-element `brightness` 1.12/0.90 on `default`, `secondary`, `destructive`, `cta` (cta 1.10 → 1.12) | delta r2 §2.1 + Codex clarification 1; the rendered state is the Core-checked state for every tenant palette | `web/components/ui/button.tsx`; oracles in `tokens.spec.ts`, `design.spec.ts` | adopted |
| 2 | Light `destructive` kit `#e7000b` → `#d2000a` | T10 ×1.12 3.99 on `#e7000b`; `#d2000a` 5.6079 / 4.6213 / 5.2872; tinted-label fixture A 4.333 → ≥ 4.63 | `web/app/globals.css`, `web/tests/browser/support/modes.ts`, `core/internal/branding/colors.go`, `docs/foundation/WEB-SHELL.md` | adopted (accepted in the r2 review) |
| 3 | Dark `text-2` kit `#cbcac9` → `#d1d0cf` | p07 S-CONTRAST local worst-case lead over `#5b5552` 4.48 → 4.757; 12.83 on `background`; Core T2/T3 unaffected | `web/app/globals.css`, `web/tests/browser/support/modes.ts`, `core/internal/branding/colors.go`, `docs/foundation/WEB-SHELL.md` | **accepted** by Codex (review of `c34deb51`, 2026-10-25) |
| 4 | `secondary` boundary: A2 added 1px `border-input` for a 3:1 boundary oracle (§4.1 1d); with G017 N1 withdrawn for filled buttons both fill-boundary oracles were removed and the border reverted to the kit form (W1) | Codex clarification 3 | `button.tsx`, `tokens.spec.ts` | adopted (both facts recorded) |
| 5 | Barlow Condensed 400 restored via the `Text` `brand` option for the shell footer identity line (`textVariants.meta`), adding the footer line to D-2 brand moments; footer-meta oracles aligned by T1b | each of the five pinned faces needs a consumer (S-FONTS/AC2) | `web/components/ui/typography.tsx`, `web/components/layout/app-shell.tsx`, `docs/foundation/WEB-SHELL.md`, `support/source-reference.ts`, `design.spec.ts` | **accepted** by Codex (review of `c34deb51`, 2026-10-25) |
| 6 | Fixture testid collision `button-link` → `button-as-link` | strict-mode "found 2" | `web/tests/fixtures/token-scopes/page.tsx`, `tokens.spec.ts` | adopted |
| 7 | S-PALETTE appearance frame sampled on the straight top border | anti-aliased curve apex | `design.spec.ts:292` | adopted |
| 8 | Filter-aware `contrast.ts` helper (`brightnessFactors`); throws on any non-brightness filter on the measured element/ancestors | Codex clarification 2 | `web/tests/browser/support/{contrast,color}.ts` | adopted; limitation recorded |
| 9 | New Light-only `grey-primary` fixture scope (`#696969` / `#ffffff`) | r2 D4: the tenant-override counterexample, rendered | `web/tests/fixtures/token-scopes/fixtures.ts` | adopted; coverage wording per §5 item 4 |
| 10 | Core platform palette sync (`colors.go`, `branding_test.go` only; accepted rows grey-primary, `input #949494`, Dark `popover-foreground #bdbdbd`; refusal Dark `accent #8a8a8a`) | delta r2; pins computed with Core's own math | `core/internal/branding/{colors,branding_test}.go` | adopted |
| 11 | Open items (below) | — | — | resolved or recorded (below); the two dispositions above are accepted |

Item 11:
- Resolved in the correction round: `web/tests/browser/accessibility.spec.ts:105` shell S-CONTRAST "primary fill against the
  header surface" was removed on Codex's requirement (§10); the label (rest, hover, focus) and focus-ring checks remain.
- No product blocker: destructive text on `muted` (4.24 with the kit value; 4.98 with `#d2000a`) is not rendered anywhere; the
  only rendered destructive text is the field error on the page ground (4.77, A2's figure with the kit value). Kept as a follow-up.
- Branding page T7 copy still reads "text on muted actions (the border colour)" (`web/app/organization/branding/page.tsx:56`);
  `TOKEN_LABELS.border` is now "Borders and dividers".
- `infra/validation/render-customizations.tsv:27` description still names the old skill (outside the allow-file).
- 11 remaining `vestive-design` mentions outside `management/` (`git grep`), all history or control code (§4.5 list).
- Rounding residual U3 (§5): exported by run p10 (numbers in §5); no open item.

## 7. Validation on the candidate `c036c297` and unexecuted checks

| Plan §7.1 row | Outcome on `c036c297` |
|---|---|
| 1 `lint lint-web controls` | first run FAIL (B-SCOPE + rollup only; allow-file lacked the r2 §5 Core line) → rerun **PASS** (`static-7`) |
| 2 `unit-web` | first run FAIL (same two rows) → rerun **PASS** (`unit-web-7`) |
| Core S1 (delta r2 §4.1) | **PASS** (s1b) |
| 3 browser, full (six projects, every spec, no `--grep`) | **146 passed, 0 failed** (p09) |
| 3r renamed identity (evidence line) | **2 passed**; image IDs unchanged |
| 4 F-controls selection | **6 / 6 PASS** |
| 5 hosted MR fast + Duo | **not executed** (after Codex source ACCEPT and MR open) |
| 6 post-merge identity stack on the frozen merged SHA (+ rows 1–4/3r only if the merged tree differs) | **not executed** |
| Cleanup receipt | **not produced** (closeout) |

Not executed by policy and not claimed: the real `rename` suite (B-RENAME host-port probes, disposable-copy `platform:start`,
B-SOURCE-HASH, B-DBRENAME — inputs unchanged; the allow-file excludes env/compose/migration/db), full `ci-controls`,
`migrations`, `ops`, `off`, the full profile. The `83f66f0c` staging merge is management-only (no source path), so candidate
coverage is not repeated for it.

## 8. Source heads for review

- Corrected source: `g067-int` **`dbf0d98e9588be29d6f612bad0c89058a36875ac`** = reviewed Result head `c34deb51ec8ed6d72097855bab51c4012df9fcd1`
  + merge of the correction writer's `g067-w4` `d13325766d670c73d33eb1791e41460f9a64c546` (§10). The correction touches
  `web/components/ui/input.tsx`, `web/tests/browser/{tokens,accessibility}.spec.ts` and `web/tests/fixtures/token-scopes/{fixtures.ts,page.tsx}` only.
- Reviewed Result head: `g067-result3` `c34deb51` (parent `g067-result2` `71ac5ac3`, parent `83f66f0c`); `g067-int` `83f66f0c` = tested
  candidate **`c036c297ca5a138f5134720f9bc7bcad0d737d12`** + merge of staging `cb0343552bf06319581c81721d2ee4dd16a4a5d4`
  (management-only: `management/SUPERVISOR-HANDOFF.md`, `management/v0/{BACKLOG,PROGRESS,ROADMAP}.md`, the G057 Goal, the RC13
  one-off record and the Codex review records of G067 Plan r1, delta r1 and delta r2; no source path).
- Result commit for this correction round: `g067-result4` `0fd43c08883a4e7fb08d9056aa23c6fd49a1c83d` (adds §10; parent `dbf0d98e`); Codex's
  source ACCEPT covers executable tree `dbf0d98e` with this Result head (review on staging `f4f47bca`).
- Motion delta (§12): `g067-w5` `56a55b15c00a07ef892526852ce38c576ee1c810` merged into `g067-int` as `b9c01618f561fcfc85fb47a522c3abc331b5d1ed`, plus this Result commit; it changes executable source beyond Codex's accepted `dbf0d98e` and returns for bounded review.
- MR source: `g067-int` after merging staging `f4f47bca3d79897fa55ad9aaa963ec03cd4dd6dd` (management-only: `management/SUPERVISOR-HANDOFF.md`,
  `management/v0/PROGRESS.md` and the Codex review record) plus this Result-only timestamp/binding fix. The merge was verified against the accepted
  tree: no path outside `management/` differs from `dbf0d98e`, so no code retest or re-review applies. History was merged, not rebased, so the
  accepted SHAs stay ancestors. A second management-only sync merged staging `059f8ef32db9bb58c5373fab2a4cfaa79a83a0f1` (Codex's Duo disposition,
  §11); the MR head became that merge plus the Result commit that adds §11.
- Staging base `cb034355` (merged in `83f66f0c`); execution-release base `d882addc`.
- Accepted delta r2 on `g067-core-delta2` `bc890b8c`; delta r1 on `g067-core-delta` `00927899`.

## 9. Result-format summary (AGENTS.md)

- **Implementation summary / files** (`d882addc..c036c297`: 474 files, +20983/−4431): successor skill `.agents/skills/vestive-brand/`
  (322 tracked files) and alias `.claude/skills/vestive-brand` → `../../.agents/skills/vestive-brand`; old skill and alias removed
  (81 files); `web/` Base UI/Rhea components, palette, type and geometry (`web/components/**` 37 files, `web/app/**` 11, fonts 5,
  `package.json`/lock, `components.json`, `eslint.config.mjs`, `THIRD_PARTY_NOTICES.md`); tests (`web/tests/**` 16 files, incl.
  `labels.spec.ts`, `support/color.ts`, `unit/host.test.mjs`); harness `infra/validation/{run,lib}.sh`,
  `infra/ci/{lib.sh,profile-paths.tsv,validate-controls.sh}`; docs `ARCHITECTURE.md`, `foundation/{WEB-SHELL,OPERATIONS}.md`,
  `AGENTS.md`, `CLAUDE.md`; Core `core/internal/branding/{colors,branding_test}.go` (delta r2).
- **Migrations/schema:** none. **Contracts/API/MCP:** none (Core default palette values change; pair semantics, API and error
  fields unchanged).
- **Tests added/changed:** existing suites migrated, no scenario deleted; new `grey-primary` fixture scope; filter-aware contrast
  helper; Core test rows (§5). Assertions removed: the two filled-button fill-boundary checks (Codex clarification 3) and the shell "primary fill against the header surface" check (§10).
- **Acceptance criteria (evidence only; Codex decides):** AC1 — 322 tracked successor files, alias target above, old paths empty,
  row 4 PASS, intake V1/V2 by owner B (§1); hosted B-SKILL-UNCHANGED/B-PROTECTED pending (row 5). AC2 — P0.1 closure (+9/−24,
  Radix 0), row 1 B-DEPS/B-WEB-COMPONENTS PASS, row 3 S-DESIGN/S-FIDELITY/S-PALETTE/S-FONTS PASS; the identity-surface screenshot
  belongs to row 6 (pending). AC3 — row 3 tokens/appearance/dev-hydration/S-IDENTITY PASS; 3r PASS. AC4 — row 3 S-OVERLAY,
  S-PORTAL-INIT, S-A11Y, S-CONTRAST, S-STATES and the rest PASS; row 6 identity-stack scenarios pending. AC5 — §4.5 docs
  (WEB-SHELL run outcomes still "filled from the G067 Result at closeout"). AC6 — rows 1–4/3r on the candidate done; row 5,
  row 6 and the cleanup receipt pending.
- **Architecture deviations:** none intended; adaptations in §6 (items 3 and 5 accepted by Codex).
- **History:** attempt 1 of B5. Failures preserved: row 1 FAIL on `2c5d7a8d` (UI_SLOTS), p05 62/40, p06 65/37, p07 99/3, rows 1–2
  FAIL on `c036c297` (B-SCOPE); rows 1–2 FAIL on `dbf0d98e` with base `d882addc` (B-SCOPE, staging management files; run
  `static-8`/`unit-web-8`); each fixed and rerun as recorded. Plan correction cycles: Core delta r1 → r2 (one); implementation correction loop: 1 (§10).
- **Follow-ups (suggestions; no scope authorized):** branding page T7 copy; `render-customizations.tsv:27` description. The
  grey-primary `contrast.json` rows were exported by run p10 (§5).

## 10. Correction round 1 (Codex implementation review of `c34deb51`, 2026-10-25; disposition CORRECTION LOOP)

Review: `management/v0/results/G067-implementation-codex-review-20261025.md` (staging `8c8d749e`, SHA-256 `794e4a11…97c4`). No source
ACCEPT and no MR authorization at the time. Required corrections and what was done (the coordinator ran all Docker work; the writer did none):

1. **Invalid input boundary.** `web/components/ui/input.tsx`: the kit's Dark half-alpha invalid border and translucent invalid ring
   (`aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40`)
   are removed; the invalid state is the solid `aria-invalid:border-destructive` in Light and Dark, as `textarea.tsx` and `select.tsx`
   already were (no edit there). New fixture scope `dark-sparse-destructive` (`{dark:{destructive:"#c07070"}}`, Core T10 5.4496 / 6.6711 /
   4.5554), S-TOKENS count 8 → 9, a blurred invalid `Input` (`data-testid="input-invalid"`) in every scope, and test
   `[S-FIELD-INVALID]` (computed border = solid `destructive` over the rendered fill; ground beside the field unhaloed; border ≥ 3:1 against
   the ground and against the field fill in every scope). Rendered rows (`field-invalid.json`, run p11): Dark default 6.84 / 5.28;
   **Dark sparse override 5.45 against the page and 4.20 against the field**; Light default and the Light side of the sparse scope 5.61 / 4.98;
   every scope ≥ 3 (Codex's blurred-border figures for the old half-alpha border: 2.67 / 2.06, ring 1.78).
2. **Withdrawn button boundary removed.** `accessibility.spec.ts` "primary fill against the header surface" is deleted (Core palette
   delta r2 clarification 3 / G017 N1); label and focus-ring checks stay.
3. **Screenshot attachments recovered.** Run p11 (below) exported every attachment of the JSON reports into `screenshots-p11/` (88 files):
   S-SCREENSHOTS shell-light/dark (`shell-1280*.png`, `shell-320*.png`, `sign-in-1280/320.png`), tokens-light/dark (38 token-screenshot
   files: fixtures, spacing, blueprint, shell) and the `contrast*.json`, `field-invalid.json`, `axe-summary.json`, `spacing-report.json` attachments.
   The post-merge identity-surface capture stays in row 6.

Source: writer W4 (Opus 5.5, isolated, stack-free) `g067-w4` `d13325766d670c73d33eb1791e41460f9a64c546` (5 files, +59/−8), merged into `g067-int` as
`dbf0d98e9588be29d6f612bad0c89058a36875ac`. Other alpha invalid usage found, not edited (outside the review's scope, not a field):
`web/components/ui/button.tsx:18` carries the same `aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50
dark:aria-invalid:ring-destructive/40` classes; a suggested follow-up (no scope authorized).

| Run | Scope | Outcome |
|---|---|---|
| Rows 1–2 `static-8`/`unit-web-8` on `dbf0d98e`, base `d882addc` | `lint lint-web`, `unit-web` | **FAIL, preserved:** B-SCOPE + B-REQUIRED-CHECKS rollup only, 19:58–19:59Z (the harness log lines show host-local EDT 15:58–15:59; the staging management files merged into the integration branch are outside the G067 allow-list when diffed against the pre-merge base) |
| Rows 1–2 `static-9`/`unit-web-9` on `dbf0d98e`, base = merged staging `cb034355` (allow-file unchanged) | `lint lint-web` (17 required IDs), `unit-web` (4 required IDs) | **PASS / PASS**, 20:11 UTC; `controls` and full `ci-controls` not run (policy) |
| Run p11 (browser A) on `dbf0d98e`, private no-port project `g067br-p11`, 20:00–20:05 UTC | `tokens.spec.ts` + `accessibility.spec.ts`, tokens-light/dark + shell-light/dark, JSON reporter | **38 passed, 0 unexpected, 0 flaky, 0 skipped** (includes S-TOKENS 9 scopes, S-FIELD-INVALID, S-CONTRAST rows for the new scope, shell S-CONTRAST) |
| Run p11 (browser B) | `design.spec.ts --grep S-SCREENSHOTS`, shell-light/dark | **2 passed** |
| Teardown | `down -v --rmi local --remove-orphans` | removed; leftovers by project label 0/0/0/0 |

Not rerun, reused from the candidate evidence (inputs unchanged for them): row 3 full p09 146/146 on `c036c297`, 3r, row 4, Core S1 (no Core
or harness path changed in the correction); the full row 3 was not repeated (Codex instruction: affected selectors only). Evidence:
`.local/evidence/g067-a1/{row1h,row2h,row1i,row2i}.log`, `static-8`, `unit-web-8`, `static-9`, `unit-web-9`, `p11-browser-{a,b}-p11.json`,
`screenshots-p11/`.

**Resource accounting.** Before any further Docker use the G067 inventory was reconciled against exact owned object IDs
(`reconcile-pre-corr/`): G067-owned containers/volumes/networks/images 0/0/0/0 (every run `g067br-p05…p11`, `g067core-s1a/s1b`, `g067-a1-nm-p01`
ended removed). After the runs (`reconcile-post-corr/`, `reconcile-post-corr2/`): totals 26 containers / 14 volumes / 9 networks / 60 images, unchanged
from before the correction round; G067-owned 0/0/0/0. The 26/14/9/60 other objects belong to the retained owner preview
`m1-preview-20261024` (11 containers/3 volumes/3 networks, retained by the owner until the walkthrough), `outreacher`, `payments-assistant` and
unlabelled `needknow-*`/`outreacher-pg` objects (not G067's), plus 5 dangling images weeks old (not G067's). Worktrees: only the coordinator's
and `g067-int` (plus the temporary Result worktrees, removed after push); writer worktrees reclaimed; auto-branches `worktree-agent-*` and the 21+
local `g067*` branches remain held. Caveat: the rows 1–2 harness reported `G064-RES accounting=unavailable (no registry)`; resource ownership
is therefore evidenced by the coordinator's inventory and the reconciliation reports, not by the harness registry.

Pending: Codex source re-review; row 5 hosted MR fast + Duo after any authorized MR; row 6 post-merge identity stack on the frozen merged SHA;
cleanup receipt; final Goal ACCEPT.

## 11. MR !102 hosted gates and the Duo disposition (Codex, staging `059f8ef3`)

Hosted MR pipeline 2987107645 on `0482301e` (MR !102, owner of the watch: the coordinator): `validate:checks` **success**; `validate:integration` is the
manual job and was **not run** (policy: no manual integration job for this no-code disposition). Duo reviewed with a truncated diff (about 15%
shown, by its own note) and reported three actionable discussions, all inside the owner-supplied Vestive Brand v0.2.1 bundle
(`.agents/skills/vestive-brand/**`, byte-pinned by intake and the hosted B-SKILL-UNCHANGED/B-PROTECTED checks); none touches G067 web/Core code.
Codex's disposition (review `management/v0/results/G067-implementation-codex-review-20261025.md`, staging `059f8ef3`): **no change in this MR; the
pinned bundle is preserved byte for byte.** Verified by the coordinator on the integration tree:

| Duo discussion | Disposition | Evidence |
|---|---|---|
| `design-system/components/Badge/README.md:5` — `bg-orange text-white` fails 4.5:1 in dark (`#fc6600`, about 3:1) | genuine defect in the bundled source; **follow-up for the source-kit owner** (no change here) | no `bg-orange`/`vx-orange` use in `web/app`, `web/components` or `web/lib` (`grep` empty); the app's own palette is tested at 4.5:1 (§5, S-CONTRAST) |
| `design-system/components/src/ui/sidebar.tsx:486` — `hsl(var(--sidebar-*))` around OKLCH values drops the outline ring | genuine defect in the bundled source; **follow-up for the source-kit owner** (no change here) | no `hsl(var(--sidebar` use in the app; `sidebar.tsx` entered only by the pristine intake commit `ddee020d` (G067 I1) and is not consumed |
| `design-system/components/src/ui/alert-dialog.tsx:7` — `@/components/ui/*` aliases | answered, no defect | `kit/README.md` step 2 copies the kit into an app "(the `@/` alias points at your source root)"; `web/tsconfig.json` has `"paths": { "@/*": ["./*"] }`; nothing in `web` imports `design-system/components` (grep empty); the commented `design-system/components` copy is a reference bundle, not app source |

The non-resolvable Duo summary note is acknowledged, not a discussion. The three discussions are replied to and resolved on the MR with the same
evidence; no further Duo review is requested. Follow-ups (suggestions; no scope authorized, owner of the kit decides): dark-mode badge foreground
guidance in the kit README/preview/`PlanCard`; `sidebar.tsx` outline variant colour wrapper.

## 12. Motion delta (owner preview feedback, 2026-10-25)

Owner direction (relayed by the supervisor): some form fields and overlays lack the subtle shadcn/Base UI style animations;
fix inside G067. Supersedes Plan §5.3 D-6 for subtle interaction/open-close motion only (D-4, D-5, the solid 2px focus
outline, the solid invalid border and every contrast oracle stand). Plan delta: `management/v0/plans/G067-motion-delta-plan.md`.
Writer W5 (Opus 5.5, isolated, stack-free; no Docker, build or test run by the writer), branch `g067-w5` from `g067-int` `be6467ef`.

What changed:

- **Dialog** (`web/components/ui/dialog.tsx`): the Backdrop takes the kit's `duration-100 data-open:animate-in data-open:fade-in-0
  data-closed:animate-out data-closed:fade-out-0`; the Popup the kit's `duration-100 data-open:animate-in data-open:fade-in-0
  data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95` (`tw-animate-css` 1.4.0, already pinned
  and imported). `bg-overlay`, no blur, the centring, scroll region, ring and shadow are kept.
- **Fields** (`input.tsx`, `textarea.tsx`, `select.tsx`): `transition-[color,box-shadow]` becomes
  `transition-[color,box-shadow,outline-color,outline-offset,outline-width]` (same `duration-200`), and the resting state gains
  `outline-0 outline-transparent`, so the focus outline eases in and out; the focused classes are unchanged. The resting outline is
  zero-width (never an indicator at rest; nothing drawn in forced colours).
- **Reduced motion:** the existing global rule (`web/app/globals.css`) shortens both to 0.01ms; no CSS change.
- Docs: `docs/foundation/WEB-SHELL.md` (successor summary D-6 sentence, the focus-adaptation paragraph, K17, the dialog lineage row);
  component header comments. `web/THIRD_PARTY_NOTICES.md` unchanged (it lists no dialog difference).

Tests affected (outcomes in the table below):

- `overlay.spec.ts` **[S-MOTION]**: assertions replaced (popup and backdrop animate, 0 < duration ≤ 0.15s; reduced motion ≤ 0.01s for
  popup/backdrop animation, popup transition and the button-transition control); S-OVERLAY settles before the Close box.
- `tokens.spec.ts` **[S-FIELD-MOTION]** (new): Input, Textarea, Select in the default scope; transition properties and 0.2s / ≤ 0.01s
  durations, 0px resting outline, settled keyboard focus solid 2px offset 2px in the ring colour. Fixture
  `web/tests/fixtures/token-scopes/page.tsx` adds a Textarea and a Select to the default scope only.
- Settle before mid-animation reads (new helper `dialogSettled`, `support/contrast.ts`): shell S-RESPONSIVE / S-LONG-IDENTITY,
  accessibility S-CONTRAST, design `openDialog` users (S-DESIGN, S-PALETTE, S-FIDELITY, S-FONTS, S-SCREENSHOTS) and the S-FONTS
  control-6 320px check, tokens S-PORTAL, S-A11Y, S-CONTRAST, S-TEXT-SPACING. Comment-only: accessibility S-A11Y, shell close path.

Focused selectors: delta Plan §6. Outcomes (coordinator-run, private no-port stacks, integration tree `b9c01618f561fcfc85fb47a522c3abc331b5d1ed` =
`be6467ef` + merge of `g067-w5` `56a55b15c00a07ef892526852ce38c576ee1c810`; executable delta vs the accepted `dbf0d98e` is exactly the W5 diff):

| Run | Scope | Outcome |
|---|---|---|
| Rows 1–2 `static-10`/`unit-web-10`, base = merged staging `a8fd9623` (allow-file unchanged) | `lint lint-web`, `unit-web` | **PASS / PASS**, 21:38 UTC; `controls` and full `ci-controls` not run |
| Run p12 browser A, project `g067br-p12`, 21:39–21:45 UTC | `tokens.spec.ts` + `accessibility.spec.ts` + `overlay.spec.ts` on tokens-light/dark + shell-light/dark | **44 passed, 0 unexpected, 0 flaky, 0 skipped**; includes S-MOTION and S-OVERLAY (shell-light/dark), S-FIELD-MOTION and S-FIELD-INVALID (tokens-light/dark) |
| Run p12 browser B | `design.spec.ts --grep "S-DESIGN|S-PALETTE|S-FIDELITY|S-FONTS|S-SCREENSHOTS"` on shell-light/dark | **18 passed, 0 unexpected** |
| Run p12 browser C | `shell.spec.ts --grep S-RESPONSIVE` on shell-light/dark | **10 passed, 0 unexpected** |
| Teardown | `down -v --rmi local --remove-orphans` | removed; leftovers by project label 0/0/0/0; totals 26/14/9/60 |

Evidence: `.local/evidence/g067-a1/{row1j,row2j}.log`, `static-10`, `unit-web-10`, `p12-browser-{a,b,c}-p12.json`, `screenshots-p12/` (102 exported
attachments, including the S-SCREENSHOTS images), `reconcile-post-motion/`. Not run, by scope: the long-identity project (S-LONG-IDENTITY uses its
own 64-character-identity build and only gained the `dialogSettled` call), full row 3, `ci-controls`, the identity/signup stacks. The M1 preview's
web image is not changed by this section (the preview still serves `c5e562f8`); updating it is a separate step.

Not changed, reported: `button.tsx` (`transition-all` already animates its outline), TextLink /
nav / site-header links (`transition-colors` includes `outline-color`), the table scroll region and the appearance control (no
transition).

**§12 addendum — S-MOTION close assertion (Codex motion review; writer W6, isolated, stack-free, branch `g067-w6` from `ea9f1037`).**
Only `web/tests/browser/overlay.spec.ts` changed. `popupTiming` now awaits `dialogSettled`, then, before Escape, installs a
MutationObserver on the popup and backdrop (`data-closed`, Base UI's closed-state attribute, the one the kit's `data-closed:animate-out`
classes key on) that records the computed animation name and duration at the closed-state moment, plus a capture-phase `animationstart`
listener recording the first animation each part starts and whether it was closed then; it resolves when both parts have both records or
after 1.5s. [S-MOTION] asserts without reduced motion, for popup and backdrop: closed state reached, close animation name not `none` and
different from the open name (`exit` vs `enter`), 0 < duration ≤ 0.15s, the animation that started is that close animation and the part
carried `data-closed` when it started; with reduced motion: closed state reached and close duration ≤ 0.01s (no `animationstart`
required). Open and reduced-motion open assertions and `toBeHidden` unchanged. Selectors unchanged (delta Plan §6 runs the whole file).
Not run by the writer: the browser test and a TypeScript check (no host `node_modules`); outcomes are coordinator-run.

Outcomes (coordinator-run, integration tree `f2e5b3e302b9f49f290511b6977315732380945d` = W6 `420752698ca4358722647ccc890374b6bdf2fe7d` + merge of
staging `c788183a3b1e3a6ea2a7c6988e7d1bc2375934f1`, management-only: 0 executable paths changed versus W6):

| Run | Scope | Outcome |
|---|---|---|
| Browser p13, project `g067br-p13`, 22:02-22:04 UTC, private no-port stack | `npx playwright test tests/browser/overlay.spec.ts --grep S-MOTION --project=shell-light --project=shell-dark` | **2 passed** (shell-light, shell-dark), 0 unexpected, 0 flaky, 0 skipped |
| Row 1 `static-11`, base = `a8fd9623` | `lint lint-web` | **FAILED B-SCOPE only** (22:01 UTC): `management/SUPERVISOR-HANDOFF.md` and `management/v0/PROGRESS.md` arrived with the merged staging records and lie outside the allow-list relative to the older base; every other check PASS. Preserved. |
| Row 1 `static-12`, base = merged staging `c788183a` (same allow-file) | `lint lint-web` | **PASS** (22:10 UTC): B-SCOPE 476 paths within the allow-list; B-REQUIRED-CHECKS 17/17 |
| Teardown | `down -v --rmi local --remove-orphans` | leftovers by project label 0/0/0/0; totals 26/14/9/60 |

Not rerun, within Codex's scope for this correction: `unit-web`, browsers A/B/C beyond the S-MOTION selector, full row 3, `ci-controls`.
Evidence: `.local/evidence/g067-a1/{corr3-rows.out,corr3b-rows.out,corr3-p13.out,p13-browser-a-p13.json,static-11,static-12}`. The M1 preview serves
`ea9f1037`; W6 changes only a test file, so its served code is unchanged.

**§12 integration (coordinator).** Codex accepted the motion delta at `g067-motion-review` `1fbda969fc1b9ef5dff17a593638ca7de51c7579` (W5 `56a55b15c00a07ef892526852ce38c576ee1c810` +
W6 `420752698ca4358722647ccc890374b6bdf2fe7d`; re-review published on staging as `5c9627bb`). It is reconciled into MR !102 by pushing `g067-int`
(previous MR head `be6467efd59b478100e5b9fedbbc804628aaa0d0`) after merging staging `5c9627bba09a3156b0976e6c70fe6e6e92f15349` (management-only; new staging base). That merge and this
paragraph change no executable file: the tested executable tree is the one the motion delta checks above ran on (`f2e5b3e3` and its W5-tree predecessors).
Hosted fast checks, Duo and the merge gates apply to the new MR head; no checks are repeated solely for the management-only staging merge.

## 13. Post-merge row 6 driver

- Driver: `management/v0/results/G067-row6-identity-driver.sh`, a minimal derivative of `management/v0/results/G064-s6-b4-identity-driver.sh` (same derived-copy mechanism, refusals, registry, redacted export, secret scan, cleanup receipt and exit codes). Writer W7, stack-free.
- What it runs on a frozen checkout of the merged SHA: the private identity overlay (configure, `identity:start`/`identity:migrate`, identity-setup secrets, `identity_it`, `mcp_seed`, as in `variants_identity`), then the full documented `identity:browser` line (16 projects, identity-light through labels-dark). Only `identity_e2e`'s one-off `run --rm` substitution is applied. The check ID is `B-IDENTITY-E2E-ROW6`, with the real matrices, and the driver requires it as PASS. The labels screenshots are copied as evidence.
- What it deliberately omits: `identity_probes`, `mcp_e2e`, `identity_privacy`, `identity_config` and the rest of `ops`, because Plan row 6 does not select them. They are reported as unexecuted.
- Invocation: `env G067_HARNESS_LOCK=held bash <driver> <frozen checkout> <evidence dir>`, launched by the coordinator that holds the harness lock (no flock wrapper). The driver refuses to start if Compose variables such as `IDENTITY_PROXY_PUBLISH` or `COMPOSE_PROJECT_NAME` are inherited from the environment (see the driver header on preview port 543).
- Outcomes (coordinator, executed 2026-10-26 UTC on a clean frozen detached checkout of merged SHA `bb605d5b19b97dcca23b59e7044bea56e6cfef89`, tree `495894e2c0c15d98237feab9bb5c8877971c190a`; driver `2ad3cf34e70e5b2ecf43aaba50a05c7bf07d5453` run by `env G067_HARNESS_LOCK=held bash <driver> <frozen checkout> <evidence dir>`, the harness lock held by the preview holder, none of the denied Compose variables inherited): driver exit 0 after run.sh 589 s. **B-IDENTITY-E2E-ROW6 PASS**: the full documented `identity:browser` line, 16 projects, **171 tests, 171 expected, 0 unexpected/flaky/skipped** (identity-light/dark 28/28, domain-platform 7, domain-custom 4, mcp-light/dark 8/8, onboarding 10/10, membership 12/12, enrollment 5/5, signup 9/9, labels 8/8), no `--grep`; M-LIST expected in mcp-light and mcp-dark with the seeded old-host row (1 before the run). **B-IDENTITY-IT PASS** (95/95 top-level tests, packages identity 74, workos 21, identitytest 0). B-DOCKER-ONLY, B-SECRET-CANARY (311 generated secrets, planted control detected), B-SCOPE (0 changed paths) and B-REQUIRED-CHECKS PASS; the evidence secret scan found none of the 311 registered secrets in 47 files. Cleanup receipt **CLEAN** (76 owned resources gone, 0 leftover, 0 unknown); host counts before and after 26 containers, 14 volumes, 9 networks, 60 images (the preview's 11 containers/3 volumes/3 networks/8 images untouched; its served commit was not changed).
- Unexecuted and not reported as passing: `identity_probes`, `mcp_e2e`, `identity_privacy`, `identity_config` and the rest of `ops`, `rename`, `ci-controls`, `migrations`, `off`; no full suite; rows 1-4/3r not repeated (the merged tree equals the tested tree).
- Labels attachments (Plan §7.2, AC2): the driver's `browser-identity-row6-attachments/` is empty (INFO row, count 0) because Playwright embeds the four `labels-org-*.png` as base64 `body` fields of `browser-identity-row6-results.json` rather than as files under `out/`; the driver does not fail on it. The coordinator extracted all four (1280 and 320 px, light and dark) from that report into `row6/screenshots/` with the same exporter as the earlier rows. The 320 px dark capture was viewed; it shows the Organization labels page. The wrapped-label evidence for the radios is G068's.
- Evidence (private, `.local/evidence/g067-a1/row6/`): `row6-summary.tsv` SHA-256 `2b0c5c63b5864064bab5573a7b2188c89dff4ae46bedfe13d8dccc35b549d5c4`, `browser-identity-row6-results.json` SHA-256 `345e6f92dc2c26941c095d73634e8d6c579067e00f460f0e4b6f737f87121cb7`, `results.tsv`, `cleanup.log`, `driver-logs/`, `report/browser-identity-row6-tests.tsv`, `screenshots/`.
- Cleanup (coordinator): the frozen checkout worktree and the integration worktree `g067-int` (with its merged local branch) are removed; writer worktrees were reclaimed at each handback; no G067-owned container, volume, network or image remains. Held, not deleted (no in-session owner approval): the remote `g067-*` writer/review branches and the older local `g067*`, `g064-*`, `worktree-agent-*`, `maint-*` branches and `refs/cleanup/20261022/*`.

## 14. Current disposition and closeout (coordinator; append-only correction, 2026-10-26)

The header, §8 and §9 above are historical: they were written before merge and say "not merged" and "pending". Nothing earlier is rewritten. Current facts: MR !102 merged as `bb605d5b19b97dcca23b59e7044bea56e6cfef89` (head `7d9f751e7cf35dce3b27cb0e89a857467cf9543a`, hosted pipeline 2987259478 success, manual `validate:integration` not run, §11; merge tree equals the tested executable tree); Codex source ACCEPT covers executable tree `dbf0d98e` and the accepted motion delta (§12); row 6 ran on the merged SHA (§13). Final Goal ACCEPT is Codex's.

| AC | Current evidence (what actually completed) | Still open or not claimed |
|---|---|---|
| AC1 | 322 tracked successor files and the alias, old paths empty, row 4 PASS, intake V1/V2 (§1, §9); hosted MR checks (row 5) success on `7d9f751e`; merged `bb605d5b` | Duo's three findings are in the byte-pinned kit and are follow-ups for the kit owner (§11) |
| AC2 | P0.1 closure, row 1 B-DEPS/B-WEB-COMPONENTS, row 3 S-DESIGN/S-FIDELITY/S-PALETTE/S-FONTS PASS (§9); S-FONTS four of five files requested (§4.3, Codex disposition); identity-surface screenshots: the four `labels-org-*.png` (1280/320 px, light/dark) extracted from the row 6 report (§13) | none claimed beyond §4.3 |
| AC3 | row 3 tokens/appearance/dev-hydration/S-IDENTITY PASS, 3r PASS (§9); motion delta checks (§12: S-MOTION shell-light/dark, S-FIELD-MOTION, browser A/B/C 44/18/10); row 6 identity-light/dark 28/28 | S-LONG-IDENTITY was not rerun after the motion delta and is not claimed beyond what the row 6 project rows above cover |
| AC4 | row 3 S-OVERLAY, S-PORTAL-INIT, S-A11Y, S-CONTRAST, S-STATES and the rest PASS (§9); row 6 on `bb605d5b`: B-IDENTITY-E2E-ROW6 PASS 171/171 across the 16 documented `identity:browser` projects (domain, mcp, onboarding, membership, enrollment, signup, labels included), B-IDENTITY-IT PASS 95/95 (§13) | `identity_probes`, `mcp_e2e`, `identity_privacy`, `identity_config`, the rest of `ops`, `rename`, `ci-controls`, `migrations`, `off`: unexecuted by selection, not passing |
| AC5 | §4.5 documentation and guards (§9); the WEB-SHELL run-outcome placeholder ("filled from the G067 Result at closeout") is satisfied by §9, §12 and §13 of this Result; no document was edited in this correction | whether WEB-SHELL needs a text touch is Codex's call |
| AC6 | rows 1-4/3r on the candidate and the correction rounds (§2, §10, §12); row 5 hosted success on `7d9f751e`; row 6 on `bb605d5b`; cleanup receipt CLEAN (§13) | the merged tree equals the tested tree, so rows 1-4/3r were not repeated (Plan row 6); full suite and full `ci-controls` were never selected |

**History kept (§9 and §12):** row 1 FAIL on `2c5d7a8d`; p05 62/40, p06 65/37, p07 99/3; rows 1-2 FAIL on `c036c297` and on `dbf0d98e` (B-SCOPE); lint-web `static-11` FAIL (B-SCOPE only, old base) with `static-12` PASS at base `c788183a`. No row 6 failure occurred (one run, exit 0).

**Immutable driver reference:** `management/v0/results/G067-row6-identity-driver.sh` is git blob `919c106c2b099915a98a76d333807c993673de27` at commit `2ad3cf34e70e5b2ecf43aaba50a05c7bf07d5453` (branch `g067-w7`, parent `bb605d5b`); the outcomes commit is `472cee428b663183d6a260dab2fb8c5c37129639` and this correction a child of it. Run evidence: `.local/evidence/g067-a1/row6/` (`row6-summary.tsv`, SHA-256 in §13).

**Cleanup disposition (task-owned, event-driven under the mandatory task cleanup; M1 and G068 untouched):**
- Containers, volumes, networks, images: row 6 receipt CLEAN (76 owned resources gone, 0 leftover); host counts 26/14/9/60 equal the pre-run counts; 0 objects labeled or named G067.
- Worktrees: frozen `g067-merged-bb605d5b`, integration `g067-int` and the writer W1-W7 worktrees removed. Remaining: the coordinator checkout, the M1 preview worktree (retained: the preview serves it until Codex or the owner retires it) and the G068 writer's worktree (separate task).
- Branches: every G067 branch fully reachable from staging was deleted: 22 remote (`g067-cons1/2`, `core`, `final`, `found`, `impl`, `intake`, `motion-review`, `result`/`result2`/`result3`/`result4`, `t1`/`t1b`/`t2`, `ui`, `w1`-`w6`) and 21 local (the same minus `motion-review`, plus the merged `g067-int`); each tip is recoverable from staging. Kept, not deleted, local and remote: `g067-w7` (the driver and these Results, unmerged until Codex imports them), `g067-plan`, `g067-core-delta` and `g067-core-delta2` (unmerged by SHA; the accepted Plan and Core delta were imported separately). These four are saved in the private bundle `.local/evidence/g067-a1/g067-unique-refs-20261026.bundle` (`git bundle verify` ok, four heads listed, `g067-w7` at `472cee42`, SHA-256 `4c510596a03982342d17b80f34b98751cef4a08820f60584cbac1b6cf41f705c`; this §14 commit is its child and lives on the remote branch); their release condition is Codex's confirmation that the driver and Results are imported, after which they can go.
- Not touched (other tasks, no G067 ownership): the `g064-*` branches, 142 `worktree-agent-*` branches, `maint-*` branches, `refs/cleanup/20261022/*`, the G068 writer branch.
- Evidence and temp: `.local/evidence/g067-a1/` (36 MB: inventory, row logs, screenshots, row 6 export, the bundle) retained until Codex's final ACCEPT, then removed except what Codex names; `/tmp` scratch directories of the run removed by the driver (0 left); the M1 preview stack and its lock-holding background task are retained (owner preview, release by owner or Codex).
- Native refusals during cleanup: none.

## 15. Supervisor import, final acceptance and final cleanup — 2026-10-26 UTC

Codex imported coordinator Result `2f4ca3aeb3a4dabefb2f3e89681fb48ba3f2776a`
(byte SHA-256 `7f876d1c1492be59f763b0836db18def774c6a7b3a55b7bbb16b14a209591a81`),
then added the current status and this disposition. The mistyped §14 W7 outcome
commit reference is mechanically corrected to `472cee428b663183d6a260dab2fb8c5c37129639`.
Prior reports, failed runs, scope exclusions and the earlier cleanup holds remain
historical. [Final review](G067-final-acceptance-codex-review-20261026.md) records
FINAL ACCEPT of the integrated source and attributable affected qualification;
this does not claim a full-suite run.

All six criteria in §14 are satisfied with the accepted prior source/motion
reviews, final hosted fast pipeline, merged-SHA row 6 and cleanup below. Codex
filled WEB-SHELL's two closeout outcome placeholders. The two unused source-kit
findings remain kit-owner follow-ups; the alias question was answered without a
defect. The owner's nonblocking form-polish observations are separately released
as G068. No additional application change or test run was made for this import.

Final cleanup supersedes the retention statements in §13–§14:

- Runtime objects: all 76 receipt IDs independently checked against Docker and
  absent (58 containers, 12 images, 3 volumes, 3 networks); zero unknowns.
  Frozen/integration/G067 worker worktrees are gone. M1, the coordinator and
  active G068 retain their existing ownership and release conditions.
- Refs: zero local and remote `g067-*` branches remain after the coordinator
  satisfied the four-ref preservation/import hold. The updated four-head bundle
  includes `2f4ca3ae`, verifies successfully, and has SHA-256
  `4393eed7b5b6b25b2ccefef35b3b0ab95d37e5b70d39d880553a9167bb7d0f7c`.
- Automatic worker refs: §14's claim that all 142 `worktree-agent-*` branches
  were unrelated was overbroad. Codex matched 23 to finished G067 native
  dispatches (including W6/W7). Every tip is reachable from long-lived remote
  main `4b15e244c77833da31f37d635b1a9e216915fba1`; none has an active worktree.
  Codex removed exactly these 23 using expected-tip checks and verified absence.
  Private proof: `.local/evidence/g067-a1/auto-ref-cleanup-20261026.tsv`.
  Other generic refs, G064/maintenance/cleanup refs and G068 remain untouched.
- Temporary output: run scratch was removed by its driver. Supervisor handoff
  scratch is removed after delivery/execution verification. Codex removed 525
  disposable evidence files after review, leaving 23 proof files totaling
  1,377,995 bytes plus a compact retention receipt. The retained set contains
  the verified unique-source bundle, ownership/ref inventory, row-6 summary,
  result/cleanup/browser JSON, four labels PNGs, small Go/browser/lifecycle/
  derivation records and selected earlier failure logs. Exact paths, owners,
  reason and release condition are in `retention-20261026.tsv` in the existing
  G067 evidence directory. Owner: Codex/G067 coordinator; release/reassessment:
  G021 closeout or an explicit earlier owner disposition. No disposable bulk
  or finished G067 worktree is held for fresh permission; no native refusal
  occurred in the supervisor cleanup.

Six superseded G067 progress checkpoints are archived with their original text
and adjusted links. G068's predecessor acceptance gate is cleared; its own source
review, hosted gates, affected qualification and cleanup remain. G057 execution
and external actions retain their separate readiness/owner holds.
