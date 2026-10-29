> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G047 — Successor Vestive shell visuals: implementation result

Goal: [G047-vestive-shell-refresh.md](../goals/G047-vestive-shell-refresh.md) · Spec: [G047-vestive-shell-refresh-spec.md](../specs/G047-vestive-shell-refresh-spec.md) · Plan: [r2](../plans/G047-vestive-shell-refresh-plan.md) (SHA-256 `8aa4a9763e0bb64349a4ae208dca0e6199848090630c7675f9a55d1da564fa65`)
Owner: Claude Code · Reviewer: Codex · Budget: **B3**

## 1. Status

**Submitted for Codex review, together with the G003 Result, under the owner-authorized combined execution (not included). Not accepted.** Only Codex ACCEPT completes G047; this Result does not mark itself accepted and does not edit PROGRESS, the Goal, the Spec or the Plan.

- All seven acceptance criteria have attributable local evidence (§7). The final complete wrapper-driven harness run on the delivered tree (S8, run `c67a4b`) has **wrapper 0, `tee` 0, postcheck passed, overall 0: 50 PASS, 3 INFO, 0 FAIL, 0 UNMET**, with 103/103 browser tests in both the root and the intelligence-disabled runs and 2/2 in the identity rename.
- **B3: 2 of 3 cycles consumed**, checkpoint after cycle 2 recorded (§10).
- No hosted CI evidence is required for G047 and none was produced. No G003 work is part of this Result.
- The foundation barrier stays closed through G021, including G047.

## 2. Execution binding and starting state

| Item | Value |
|---|---|
| Binding | Supervisor-created `/tmp/g047-g003-19f945a-approved-binding.json`, SHA-256 `ffb8b7c656324f3a8bedb9ca3f70bbfe22e09e1846a8af965e685082df13ead9` (verified before S0). Verified byte-identical copies retained outside both checkouts at `~/.local/state/platform-evidence/batch-19f945a/` (`RETAINED.sha256`). |
| Initial BASE | `19f945af7b7bfd1b27cb395fd56b0c2cec3b163b`, tree `b2105bb21ea3769f72d935e67ecafe39a5f4c389` |
| Execution checkout / branch | `.claude/worktrees/g047-exec` (linked worktree of the main checkout) / `worktree-g047-exec` |
| G047 snapshot | `/tmp/g047-19f945a-approved-inputs.sha256`, SHA-256 `0fd996fda72c8b13b2845146b8a50fcd26cfecafe01aca3f9adade8f3c226cbb`, 23 entries, `--path-re` and `--plan` as in the binding |
| Authorization addendum | `management/v0/plans/G047-G003-sequential-execution-authorization.md`, SHA-256 `d80043f76a850a3a8fe13afbf3deb4748ea7cbea955908aeffc3b03afd108ea7` (equal to the binding) |
| Evidence root | `EXEC_ID=361c5816`; `EV=~/.local/state/platform-evidence/g047-361c5816` (absolute, normalized, no symlinked component, outside both checkouts and `/tmp`, created `mkdir -m 700` only after those checks) |

### 2.1 S0 startup checks (read-only; all passed before the first repository write)

1. **Evidence identity first:** `EXEC_ID` from `/dev/urandom`, then `$EV`, containment checks, then `mkdir -m 700`.
2. **Read-only Git** (`GIT_OPTIONAL_LOCKS=0 git -c core.fsmonitor=false`, `GIT_DIR`/`GIT_WORK_TREE`/`GIT_INDEX_FILE` unset): top level = the bound checkout; `HEAD` = BASE; branch `worktree-g047-exec`; `status --porcelain=v1 -z --untracked-files=all --ignored` empty (0 bytes; digest `e3b0c442…b855`). All 23 G047 and 5 G003 snapshot entries are committed at BASE with blob hashes equal to the snapshot; 63 kit files committed; `sha256sum -c references/shadcn-kit.sha256` 63 OK. Evidence `$EV/s0-git.txt`. (A first attempt of this check under zsh clobbered `PATH` through zsh's special `path` variable and produced only command-not-found noise; it was discarded and rerun under `bash --noprofile --norc`. It wrote nothing to the repository.)
3. **Approved-wrapper precheck:** `run-approved.sh … --check-only` → `precheck passed (23 inputs, snapshot 0fd996fd…)`, wrapper 0, tee 0 (`$EV/s0-wrapper-check.txt`).
4. **Resources:** no `.env`; the 34 default and harness-variant ports free (`ss -ltnH`); `docker ps -a`/`network ls`/`volume ls`/`image ls` recorded with IDs (`$EV/s0-docker-inventory.txt`, SHA-256 `176507d5…05d1`): no container, network or volume of the root Compose project `beginify`. Unrelated developer resources (`payments-assistant-*`, `needknow-*`, `outreacher-*`, `summit-net`, the shared `project-archetype-render-cache` volume) were present and were not touched.

**S0b:** the allow-list (§6.2 of the Plan) was the first repository write; it equals the Plan block byte-for-byte; SHA-256 `b5d582788bb78a3583d73703e4c5f3afe6e5f95361e92f59c34dd6a15380b1e9`.

## 3. Implementation summary and files changed

Exactly the Plan §6.1 inventory (18 paths). Values and roles only; Base UI, the button/dialog APIs, routes, dependencies, fonts, assets, the appearance engine, scoped portals, `--angle` and radius tokens are unchanged.

| # | Path | Change |
|---|---|---|
| 1 | `web/app/globals.css` | Palette block comment and values exactly as Plan §4.1 (12 Light, 5 Dark values). Lines 1–640 (`GLOBALS_EJECT_SHA`) and everything else byte-unchanged. |
| 2 | `web/app/page.tsx` | Kicker `leading-[1.2]`; H3 `leading-[1.2]`; cell body and About paragraph `text-[1.04rem]`; About items `leading-[1.35]` |
| 3 | `web/components/layout/app-shell.tsx` | Footer `leading-[1.35]`, appended |
| 4 | `web/components/ui/button.tsx` | Base `cursor-pointer` after `select-none`; destructive, outline, ghost and link strings exactly as Plan §4.4 |
| 5 | `web/components/ui/dialog.tsx` | Popup, Close (`ghost`, `size="icon"`, `absolute top-5 right-5 text-muted-foreground`), header, title and description exactly as Plan §4.5; `DialogFooter`, Base UI wiring and the portal unchanged |
| 6 | `web/tests/browser/support/modes.ts` | Independent `SOURCES` (16 tokens × 2, typed from Plan §4.1); `PALETTE` derived from it (same six keys) |
| 7 | `web/tests/browser/support/spacing.ts` (new) | Text-spacing stylesheet, expected-element inventory checks, focused negative cases (Plan §8.4) |
| 8 | `web/tests/browser/design.spec.ts` | S-DESIGN additions (A8) and M2; new S-PALETTE (A1) |
| 9 | `web/tests/browser/accessibility.spec.ts` | S-CONTRAST additions (A5); new S-TEXT-SPACING shell (A3) |
| 10 | `web/tests/browser/shell.spec.ts` | S-RESPONSIVE: the open dialog also fits vertically (A7) |
| 11 | `web/tests/browser/tokens.spec.ts` | Independent source chain (M4), nesting-aware locators (M5), S-PORTAL (M3, A9), S-RUNTIME (M6), S-A11Y/S-SCREENSHOTS (M7), S-CONTRAST (A6), S-TOKENS (A10); new S-STATES (A2) and S-TEXT-SPACING fixtures (A4) |
| 12 | `web/tests/fixtures/token-scopes/fixtures.ts` | `FixtureScope` with optional `parent`; `light-only` and `dark-only-in-A` (Plan §8.2 values) |
| 13 | `web/tests/fixtures/token-scopes/page.tsx` | Shared scope content; nested scope inside A; a disabled button row per scope |
| 14 | `docs/foundation/WEB-SHELL.md` | Plan §6.3: token table with source roles; "Successor design input (G047)" (inventory with full hashes, provenance, mapping summary, K1–K22, unused roles); G002 sections labelled historical with superseded rows annotated; browser-test and status updates. The `appearance-init` hash line is unchanged. |
| 15 | `infra/validation/lib.sh` | `BROWSER_MATRIX_FULL` gains S-PALETTE and S-TEXT-SPACING (shell) and S-STATES and S-TEXT-SPACING (tokens). No other line. |
| 16 | `infra/validation/render-customizations.tsv` | Reason text of the `globals.css` and `page.tsx` rows gains "; …(G047)". Paths and the 8 rows unchanged. |
| 17 | `management/v0/results/G047-vestive-shell-refresh-result.md` (new) | This Result |
| 18 | `management/v0/results/G047-scope-allowlist.txt` (new) | Plan §6.2, verbatim |

Pre-change hashes of rows 1–16 equal the Plan §6.1 values. Final hashes: §11.

## 4. Migrations and schema

None. No migration, seed, table or domain identity; B-MIGRATE-SET unchanged (`00001_baseline.sql`).

## 5. Contracts, API and MCP

None. Routes stay `/`, `/api/health`, `/api/ready` (B-WEB-BOUNDARY). `Button`, `Dialog*`, `TokenScope`, `tokenOverrideStyle`, `TokenOverrides`, `SEMANTIC_COLOR_TOKENS`, `NON_COLOR_TOKENS` and the appearance API are unchanged. No dependency, font, asset or remote request (B-DEPS, C1–C12, B-WEB-DESIGN; the package files and `import` lines of `web/app` and `web/components` are unchanged).

## 6. Tests added and modified

Per run: **103 browser tests** (95 + 8), in the root and the intelligence-disabled runs alike; identity rename unchanged at 2. No existing assertion removed; no tolerance loosened.

| ID | Assertion |
|---|---|
| M1 | `PALETTE` = the Plan §4.1 values, derived from the independent `SOURCES` |
| M2 | S-DESIGN dialog: `box-shadow` is `none` (zero layers; stricter than G002's single hairline) and a 2px solid `--border` rule on all four sides |
| M3 | S-PORTAL: the Close icon equals the scope's `muted-foreground` |
| M4 | S-TOKENS expected value: own override → parent chain → `SOURCES` (never read from the page) |
| M5 | Fixture content located through `[data-token-scope="X"] > section` |
| M6 | S-RUNTIME: `light-only` unchanged; `dark-only-in-A` Light fill follows A's runtime change; its Dark fill stays its own |
| M7 | S-A11Y and S-SCREENSHOTS fixtures: dialogs in A, B and `dark-only-in-A` |
| A1 | **S-PALETTE** (shell-light/dark): all 16 × 2 `:root` sources parse to `SOURCES` (≤1 per RGBA channel); rendered/computed consumers (header ground and rule, band, H1, kicker, footer, trigger fill and label, selected option, appearance frame; panel, panel rule, title, scrim over the ground via the test's own `over()`) equal `SOURCES` (≤2); `fonts.json` |
| A2 | **S-STATES** (tokens-light/dark): per scope and variant — fills, labels, `::before` hover `brightness(1.12)` and pressed `brightness(0.9)`, clip-path = `--angle`; outline boundary/inside/hover; ghost transparent/hover; link ink, 2px underline, 4px offset, hover ink; `cursor: pointer`, radius 0; keyboard focus outline colour (`ring`, destructive `destructive`), 2px solid, offset 2px; disabled `:disabled`, opacity 0.5, `pointer-events: none`, focus does not move |
| A3, A4 | **S-TEXT-SPACING** shell (1280×800 and 320×640: closed page, focused skip link, open modal) and fixtures (1280×800: six scopes × 16 elements, then modals in A, B and `dark-only-in-A`); expected-element inventory with exact texts, visibility, covering, clipping, real scroll reachability, overlap, labels inside controls and inside the angle, usable controls, no page overflow, modal containment; focused negative cases first (§7, AC5) |
| A5 | S-CONTRAST shell: skip link focused (text ≥4.5, ring ≥3), dialog tips (≥4.5), Close icon on its hover fill (≥3); recorded without threshold: panel vs scrimmed page, dialog rule vs panel |
| A6 | S-CONTRAST fixtures: pressed labels for default, secondary and destructive in every scope (≥4.5); Close icon on hover (≥3); disabled pairs recorded as "exempt: inactive component" |
| A7 | S-RESPONSIVE: the open dialog is inside the viewport vertically |
| A8 | S-DESIGN computed type and dialog geometry at 1280×800 (±0.01px) and the overlay = `SOURCES[mode].overlay` |
| A9 | S-PORTAL: overlay = `SOURCES[m].overlay` (alpha included); the nested scope's popup and backdrop mount in `dark-only-in-A` |
| A10 | S-TOKENS: `light-only` ≠ default in Light and = default in Dark; `dark-only-in-A` = A in Light and ≠ A in Dark; scope count 6 |

## 7. Validation commands and outcomes

Every validation run used the Plan §8.7.2 invocation exactly (explicit `bash --noprofile --norc -c`, immediate `PIPESTATUS` capture, precedence wrapper → `tee` 70 → status record 71), with the binding values. The launcher that exports the binding values and runs that snippet is `$EV/g047-vw.sh` (SHA-256 `7687e6a51f2dea1baad6b5e48a2007ceb4456b1dbcbc7d634b79f09f4802a2f0`). No repository write occurred during any run.

| Step | Suites | Run ID | wrapper | tee | Postcheck | Overall | Checks |
|---|---|---|---|---|---|---|---|
| s1 (baseline, unmodified tree + allow-list) | all | b05ef4 | 0 | 0 | passed | **0** | 50 PASS, 3 INFO, 0 FAIL/UNMET; browser root 95/95, off 95/95, rename 2/2 |
| s3 | static provenance | cfae8d | 0 | 0 | passed | **0** | 24 PASS, 3 INFO |
| s4a | static provenance root | d16a00 | 1 | 0 | passed | 1 | B-BROWSER-ROOT FAIL: 99 passed, 4 failed (cycle 1) |
| s4b | static provenance root | 7e2f15 | 1 | 0 | passed | 1 | B-BROWSER-ROOT FAIL: 101 passed, 2 failed (cycle 2) |
| s4c | static provenance root | a9a895 | 0 | 0 | passed | **0** | 36 PASS, 3 INFO; browser root 103/103 |
| **s8 (final, delivered tree)** | all | **c67a4b** | 0 | 0 | passed | **0** | **50 PASS, 3 INFO, 0 FAIL/UNMET**; browser root 103/103, off 103/103, rename 2/2 |

Source rows: `$EV/runs.tsv`. Each run's report directory is `$EV/<step>/report/` with `harness.out` beside it.

### 7.1 Final run (S8) details

- **Integrity around S8** (`$EV/integrity-pre-s8.txt`, `$EV/integrity-post-s8.txt`; identical apart from their timestamps): tree digest `387977431a2c42d9a637057521fb5677317cec9362798715eddbff9cb5e1506e`; NUL status digest `4f90d77f4f33f54e3f755b5c20a148fc634a40910178630fa216ad0462443195`; change inventory (`git diff -z --name-only BASE` plus untracked, excluding standard ignores) **equals the 18 Plan §6.1 paths exactly**; ignored inventory empty; kit manifest 63 OK; no package or `import` diff in `web/app`/`web/components`; no `.env`. No repository write occurred during any run. This Result was completed after S8 (Plan S10); it is the only path written after the final run, and its final hash is in §11.
- **Every check passed:** the 50 required IDs, including B-SHELL, B-DEPS/B-DEPS-CONTROLS, B-WRAPPER-CONTROLS (W1–W9), B-WEB-BOUNDARY/-TOKENS/-COMPONENTS/-DESIGN, B-RENDER-ID/-PLAN/-REPEAT (8 registered conflicts), B-PROTECTED, B-SKILL-UNCHANGED, B-BROWSER-ROOT (103), B-CHECKS-ROOT, the migration suite, B-OUTAGE, B-PERSIST, B-RENAME, B-BROWSER-RENAME (2), B-DBRENAME, B-TOGGLE-OFF-ADDITIONS, B-BROWSER-OFF (103), B-TOGGLE-OFF, B-SECRET-CANARY (48 generated secrets absent) and B-SCOPE (18 changed paths since `19f945a`, all within the allow-list). INFO: PREFLIGHT (Docker 27.5.1, Compose 2.32.4), PROBE (devtcp), IMAGES.
- **Leftovers:** `inventory-leftover.txt` is empty for every run (s1, s3, s4a, s4b, s4c, s8). A final read-only inventory (`$EV/final-docker-inventory.txt`) shows the same 15 containers, networks and volumes, by ID and name, as the S0 inventory; no root-project or harness image remains. G047 deleted nothing itself.

### 7.2 Screenshots (before: s1 run b05ef4, baseline tree; after: s8 run c67a4b, delivered tree)

Viewport as named, `deviceScaleFactor` 1; Chromium 153.0.8010.12 in the unchanged runner image (`web-browser.Dockerfile` `83979271…2705d`, Playwright `v1.63.0-noble@sha256:eff16c30…a4a27`); fonts: Liberation Sans and Liberation Mono (`fonts.json`, after; the same runner definition before); open dialogs captured after their animations settle (opacity 1). Files: `$EV/s1/before/`, `$EV/s8/after/` (`SHA256SUMS` in each).

| Image | Before SHA-256 | After SHA-256 |
|---|---|---|
| shell-light 1280 | `b4f1c16f…9717` | `56492197…b954` |
| shell-light 1280 dialog | `77eab2bf…91b0` | `1790a759…26f1` |
| shell-light 320 | `b1949aee…76ff` | `06d517ad…3399` |
| shell-light 320 dialog | `76c9e113…806c` | `0e3c3936…d651` |
| shell-dark 1280 | `69919006…cebb` | `b32d95ff…a0e1` |
| shell-dark 1280 dialog | `95dcfaa0…4c36` | `dcc20336…6c5a` |
| shell-dark 320 | `69b3aa72…1709` | `5dae8321…beef` |
| shell-dark 320 dialog | `2c486aa8…9ac1` | `0a4ed40d…8934` |
| tokens-light fixtures | `6f0fb45d…54d1` | `04c9406b…e237` |
| tokens-light fixtures-dialog-A | `63f826cf…dad9` | `ebbbc1dd…fe25` |
| tokens-light fixtures-dialog-B | `0ba74c04…bcee` | `215a3c4b…bc8e` |
| tokens-light fixtures-dialog-dark-only-in-A | — | `762d4142…3149` |
| tokens-dark fixtures | `752d7cd3…df57` | `fc57ffab…b4c3` |
| tokens-dark fixtures-dialog-A | `8a4a344a…1702` | `02f630af…7d0f` |
| tokens-dark fixtures-dialog-B | `59edf0ad…74a5` | `d5aac7c4…854c` |
| tokens-dark fixtures-dialog-dark-only-in-A | — | `4eaa58f4…30ec` |

The after images show the warm-paper Light ground `#f4f2ee`, the `#ebe8e3` band, `#d4cfc8` rules, the `#0062d6` primary and selected option, the `#615b55`/`#8f8a84` appearance frame, gold accent ink, the white (Dark `#131313`) 2px-ruled dialog with the muted Close, solid destructive fills, 2px link underlines, disabled rows and the nested scope (pixel-sampled by the reviewer). **Note:** in every closed-shell image, before and after, the "Keyboard help" trigger is in its hover shade (`#645cff` before, `#006ef0` after), because `waitForHydration` clicks it and S-SCREENSHOTS does not move the pointer away; the resting `#0062d6` is asserted by S-PALETTE with the pointer at (0,0). The s4c and s8 after images are byte-identical (deterministic rendering). Spacing-trial images are in `$EV/s8/after/*spacing-*.png`. Screenshots supplement the measurements; they are not first-paint evidence.

### 7.3 Text-spacing trial (S-TEXT-SPACING; final run)

| Project | Real states (items, findings) | Negative cases |
|---|---|---|
| shell-light, shell-dark | 1280×800 closed (27, 0); skip link focused (1, 0); modal (7, 0); 320×640 closed (27, 0); skip link focused (1, 0); modal (7, 0) | missing, `visibility: hidden`, `display: none`, covered, clipped, unreachable (scrolling ancestor and viewport), angled cut, overlap, and covered Close in each modal: each reported by its matching check; the below-the-fold positive case clean |
| tokens-light, tokens-dark | fixtures closed (96 = 6 scopes × 16, 0); modals in A, B and `dark-only-in-A` (7 each, 0) | as above (closed-page set once, covered Close per modal) |

Reports: `$EV/s8/after/*__spacing-report.json`. Item rects are stored per item after its own scroll into view; the overlap assertion uses a separate single-pass measurement. Limits: the `text`, `label-outside`, `unusable` and `page-overflow` checks have no dedicated negative case (not required by Plan §8.4), and "covered" samples the centre of each text rect.


## 8. Acceptance criteria

| AC | Outcome | Evidence |
|---|---|---|
| **AC1** Scoped consumers implement the reviewed mapping | **Met (local)** | Palette block equals Plan §4.1 (string-checked by the reviewer). S-PALETTE: 16 × 2 sources = `SOURCES` (alpha included), and rendered/computed consumers equal `SOURCES` in each project (header ground and rule, band, H1, kicker, footer, trigger fill/label, selected option, appearance frame, panel, panel rule, title, scrim composite). S-DESIGN (A8, M2), S-STATES (all variants × 6 scopes × states), S-CONTRAST. Before/after screenshots (§7.2). Retained values are justified by K5–K22 and the unused-role list in `WEB-SHELL.md`. |
| **AC2** Interaction, configuration and dependency contracts preserved | **Met (local)** | S-LANDMARKS, S-NAV, S-SKIP, S-TAB, S-OVERLAY, S-MOTION, S-BOUNDARY pass in root and disabled runs; B-RENAME and B-BROWSER-RENAME (configuration-only rename, 2/2); B-DEPS and B-DEPS-CONTROLS (C1–C12, C8b, C8c); empty package and import diffs; B-WEB-BOUNDARY (routes exactly `/`, `/api/health`, `/api/ready`); B-WEB-DESIGN (no remote font/asset, no byte-identical skill asset of 63, no marketing literal); B-WEB-COMPONENTS (E1–E7). |
| **AC3** Appearance and first paint | **Met (local)** | All S-APPEARANCE-* scenarios, S-FIRSTPAINT (8), S-FIRSTPAINT-DELAYED (2), S-HTML-ATTRS (6), S-HYDRATION-DEV (6) and S-HEAD-ORDER (served `appearance-init` hash = `WEB-SHELL.md`) pass with the updated independent `PALETTE`; appearance code unchanged. |
| **AC4** Overrides separate and isolated | **Met (local)** | S-TOKENS (16 tokens × 6 scopes via the independent chain; Light-only and nested Dark-only assertions; count 6), S-PORTAL (popup and backdrop in every scope incl. the nested one, across live mode changes; overlay = `SOURCES` with alpha), S-RUNTIME (no neighbour leakage; nested inheritance follows A's runtime Light change; nested Dark stays its own), S-STATES; S-NOFIXTURE (fixture route 404 in production). |
| **AC5** Rendered accessibility | **Met (local)**, bounded as stated | S-CONTRAST: shell 34 thresholded pairs per appearance, fixtures 228, **0 below threshold**; pressed states for default/secondary/destructive in all scopes; smallest margins 4.68 (primary hover), 4.73 (Dark secondary pressed), 4.80 (Light destructive hover), 4.86, 4.88; Dark primary fill vs ground 3.44 (≥3); 36 disabled pairs recorded as exempt; scrim/panel 3.40 Light / 1.11 Dark and rule 1.55 / 1.23 recorded without threshold. S-TAB focus ≥3:1; S-ANGLE; S-MOTION; S-RESPONSIVE (320, 768, 1280, 640×400, 200% text; dialog fits both axes); S-TEXT-SPACING (§7.3); S-A11Y (axe, no violations). Not a full WCAG certification. |
| **AC6** Full stack and validation/provenance regressions | **Met (local)** | S8 (run `c67a4b`): wrapper 0, `tee` 0, postcheck passed, overall 0; 50 PASS, 0 FAIL/UNMET; root and disabled browser matrices 103/103 each, rename 2/2; inherited datastore, migration, isolation, outage, persistence and identity scenarios; all static/dependency/wrapper controls; B-PROTECTED and B-SKILL-UNCHANGED; the NUL-safe change inventory equals the 18 reviewed paths; pre/post integrity digests identical (§7.1). |
| **AC7** Attributable, bounded handoff | **Met** | This Result: AC map, exact commands and run rows, manifest digest `fcda5a27…9163` (63 OK), snapshot digest `0fd996fd…6cbb`, input hashes (`WEB-SHELL.md` successor inventory), provenance statement and line check (§11), cycle log with checkpoint (§10), leftover reports. No schema, product or CI implementation; no closed-record rewrite; PROGRESS/Goal/Spec/Plan unchanged. Codex ACCEPT is recorded separately. |


## 9. Architecture deviations

None. No architecture, security-model, tenancy, authority or contract change. `web` remains a thin UI/BFF; `core` and `intelligence` are untouched.

Implementation-level notes (within the approved Plan):
- **Test-only timing helpers.** S-STATES waits for animations on the control and its pseudo-elements (`getAnimations({ subtree: true })`) plus two frames before reading computed styles; `inspectSpacing` waits for running document animations (bounded at 2 s) and two frames, and can re-focus and assert `:focus` for a focused-state measurement. Both were added in cycles 1–2 (§10). They change no expectation and loosen no tolerance.
- **S-DESIGN viewport.** S-DESIGN and S-PALETTE set 1280×800 explicitly (the projects' default) so the A8 values are measured at the Plan's viewport.
- **WEB-SHELL.md inventory hashes** are written in full rather than abbreviated.
- **Execution convenience (disclosed by the review).** `$EV/g047-vw.sh` is a launcher outside the repository that exports the P7 binding values (`BASE`, `SNAP`, `SNAP_SHA`, `N`, `RE`, `PLAN`, `ALLOW`, `EV`) and runs the Plan §8.7.2 snippet verbatim; the integrity, evidence-extraction and provenance-check helpers are read-only scripts in the session job directory. None is a harness, wrapper, CI or repository change.

## 10. Execution budget, cycles, checkpoint and unresolved issues

**Budget class B3. Maximum 3 cycles. Consumed: 2** (main agent 2; read-only reviewer 0). No extension requested or used.

| Cycle | Opened by | Material failure | Diagnosis | Corrective implementation | Retest |
|---|---|---|---|---|---|
| 1 | s4a (run d16a00): B-BROWSER-ROOT FAIL, 99/103 | (a) S-STATES: the `::before` hover filter read `brightness(1)`; (b) S-TEXT-SPACING shell: the focused skip link measured translated off-screen (`y = -31`), "covered by nothing" and "unreachable" | (a) `settled()` awaited only the control's own animations, not the `::before` filter transition; (b) suspected focus loss | (a) `settledDeep` (subtree animations plus two frames) in S-STATES; (b) a `focused` option in `inspectSpacing` (re-focus and assert `:focus`, no mouse move) and the skip-link state measured before the full-page screenshot | s4b (run 7e2f15): S-STATES pass; 101/103 |
| 2 | s4b: S-TEXT-SPACING shell still failed only in the skip-link state, with **no** `focus` finding | The link matched `:focus` but was still at its translate start value | The test emulates reduced motion; its global rule gives every element a 0.01 ms transition, which only advances on the next frames, and the measurement ran in the same frame as the focus (S-SKIP does not emulate reduced motion; S-CONTRAST's skip check awaited `settled()`) | `inspectSpacing` awaits running document animations (bounded 2 s) and two frames before measuring | s4c (run a9a895): overall 0, 103/103 |

No product value, class or measurement failed in any run; both defects were test-timing defects of the same class ("measured before a transition settled"). No assertion was removed or loosened.

**Checkpoint after cycle 2** (recorded in `$EV/cycle-log.md` before any cycle 3): AC1, AC3, AC4 and AC5 evidenced on the root stack; AC2, AC6 and AC7 pending the final run and Result; the failures converged (99 → 101 → 103), with no regression of earlier-passing criteria; the remaining uncertainty was the full `all` run on the changed tree; the work was bounded. Decision: continue; cycle 3 kept in reserve and **not used**.

**Read-only review (S6)** by one subagent (no edits, 0 cycles): no blocking finding. Dispositions:
- *Should-fix (disclosure):* the V-W launcher `$EV/g047-vw.sh` exports the P7 binding values and runs the §8.7.2 snippet verbatim; it is execution convenience outside the repository, not a harness or wrapper change (§7).
- *Notes, recorded here, no change:* the closed-shell screenshots (before and after) show the trigger in its hover shade (`#006ef0` after, `#645cff` before) because `waitForHydration` clicks it and S-SCREENSHOTS does not move the pointer away; the resting `#0062d6` is proven by S-PALETTE with the pointer at (0,0). `spacing-report.json` stores each item's rects after its own `scrollIntoView` (per-item frames); the overlap assertion uses a separate single-pass measurement. The fixtures spacing test asserts each modal state before attaching the report, so a modal failure would lose the attachment (the shell test attaches first). Four oracles intentionally compare with page tokens as Plan M2/A8/A2 specify, while S-PALETTE/S-TOKENS bind those tokens to `SOURCES`. The `text`, `label-outside`, `unusable` and `page-overflow` spacing checks have no dedicated negative case (Plan §8.4 does not require one).

**Unresolved issues / blockers:** none for G047. Limits: Chromium only; local/system fonts (the runner renders Liberation Sans/Mono, `fonts.json`); bounded text-spacing trial; disabled states exempt by WCAG definition.


## 11. Provenance, integrity and cleanup

**Provenance.** Nothing from the kit is copied (Plan §5; `WEB-SHELL.md` "Successor design input (G047)"). The S9 line check over the actual `git diff -U0 19f945a -- web ':(exclude)web/tests'` added lines (31 lines in `globals.css`, `page.tsx`, `app-shell.tsx`, `button.tsx`, `dialog.tsx`) against the 1037 kit lines of 40 or more characters from all 63 kit files: **zero exact matches**; one shared 60-character fragment (`focus-visible:outline-2 focus-visible:outline-offset-2 focu…`) is **already present at BASE** in G002's `button.tsx`, so it is not a stop item. Output `$EV/provenance-line-check-final.txt`. This string check is supplementary evidence, not a licence determination. B-WEB-DESIGN found no `web/` file byte-identical to any of the 63 kit assets. Kit manifest `fcda5a273daee8117598ab32a244a36d898112c27bd94b3908fbe7a362c79163`: 63 OK at S0 and after S8.

**Approved inputs.** All 23 snapshot entries matched before and after every run (wrapper precheck and postcheck); the binding, both snapshots and the addendum are byte-identical to their pinned digests.

**Final hashes of the 18 paths** (the delivered tree; this Result's own hash is recorded in the G003 transition evidence, `~/.local/state/platform-evidence/batch-19f945a/`, because a file cannot contain its own hash):

| Path | SHA-256 |
|---|---|
| `web/app/globals.css` | `c76ea2dc1ca39d8ab2b3b723e8d0e1103c786689dec1551900fb1d0526bba8f7` |
| `web/app/page.tsx` | `25b0f6bd47524181b32291595bc8bb314464cde035d2ff0416a2128a78f85efa` |
| `web/components/layout/app-shell.tsx` | `ccfad26b0f28d4b0ca04f4d6b80ddeaeeccf241fff89ff0b7dba6107e482b976` |
| `web/components/ui/button.tsx` | `0315d51bff6c58e0ac859dbaa8cce15eff11d7d09a35340eecc424dea9a1a9e3` |
| `web/components/ui/dialog.tsx` | `44282c53b7914b6a1daf0524709f0600e2914937e67e0d0f12d7df6be9ec9fd6` |
| `web/tests/browser/support/modes.ts` | `1588769a2f05d71da3585fb9f43b4e8e126f28e64f6747f0f372cebd78b12e38` |
| `web/tests/browser/support/spacing.ts` | `321d353f57f6ae588f6d67def3b8a8b577514382746391241b466a8cde732540` |
| `web/tests/browser/design.spec.ts` | `e0190fbef0d24fb6a9a15c3ccb6270d0e94c606bbd988d564aec106f0b013c51` |
| `web/tests/browser/accessibility.spec.ts` | `fdbe2e9de6bf29bc65f12f3f7a6e4e549e43d012c675bdfcc0e891062372ec17` |
| `web/tests/browser/shell.spec.ts` | `61eaab1d2cdeddfa6356c2d33b813404f9aca049dd87e9e4a8300beefa380380` |
| `web/tests/browser/tokens.spec.ts` | `3967b3b0d6109af92795caade22f28014e7823d94f43a7848b54c68826dd9d4b` |
| `web/tests/fixtures/token-scopes/fixtures.ts` | `6142919d9832b2a2257324448cd52001ab133d6a495b68adb61cb41b318ae24f` |
| `web/tests/fixtures/token-scopes/page.tsx` | `f039326077cf0fb58f0b1e6021f4f487ed783834588981dd2a5b0572ef855bd8` |
| `docs/foundation/WEB-SHELL.md` | `1ce630c31e2a29ab0f8cfa7e07bcf5efcb13cf374d4b643d94b41337a4cae9c0` |
| `infra/validation/lib.sh` | `89c6970fe74edc3ee3a0f1faacd06e3974dcfaf40b1a4f813d79920cc95180bd` |
| `infra/validation/render-customizations.tsv` | `71b9241ed41670eb47b3180ee313cf7f0dc332e7199764c054918cb321003792` |
| `management/v0/results/G047-scope-allowlist.txt` | `b5d582788bb78a3583d73703e4c5f3afe6e5f95361e92f59c34dd6a15380b1e9` |

**Evidence locations** (outside both checkouts; nothing under `/tmp` is relied on): `$EV=~/.local/state/platform-evidence/g047-361c5816` (S0 evidence, `runs.tsv` SHA-256 `89d89f0f…0207`, each step's `harness.out` and `report/`, before/after evidence, integrity, inventory and provenance outputs, `cycle-log.md`); retained binding and snapshots in `~/.local/state/platform-evidence/batch-19f945a/`.

**Cleanup.** Only the accepted harness cleanup ran (every Docker-creating step was a wrapper-driven harness run); each run's leftover report is empty and the final inventory equals the S0 inventory (§7.1). Execution scratch files (the edited-file staging copy, launcher and helper scripts) live in the session's job directory outside both checkouts; they are not repository content.


## 12. G003 handoff (informational)

The phase transition follows the authorization addendum: the G047 checkpoint commit (exactly these 18 paths, parent = initial BASE) becomes G003's BASE, and the external mechanical reconciliation record carries the changed pins (`lib.sh`, `render-customizations.tsv`), the 103-test browser matrices with three added IDs, the `WEB-SHELL.md` content and the two new G047 management files. G047 does not edit the G003 Plan and does not approve G003.

## 13. Suggested follow-up (no authorization)

- The unconsumed `DialogFooter` keeps offsets that assume the old popup padding (residual R12); its first consumer restyles it under its own Goal.
- Optional owner choices recorded in the Plan remain open: K16 (gold-rule selected state) and the Light scrim (K8).
- Full WCAG 1.4.12 certification remains out of scope; S-TEXT-SPACING is a bounded trial for touched shell and control text.
