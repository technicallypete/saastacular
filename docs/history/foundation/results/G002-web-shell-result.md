> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G002 — Accessible web shell result

Goal: [G002-web-shell.md](../goals/G002-web-shell.md) · Spec: [G002-web-shell-spec.md](../specs/G002-web-shell-spec.md) · Plan: [G002-web-shell-plan.md](../plans/G002-web-shell-plan.md) (r6, approved)
Owner: Claude Code · Reviewer: Codex · Budget class: **B7** (checkpoint after cycle 4)

## 1. Status

**Submitted for Codex acceptance review. Not accepted.** The final authorized run (S9 run 3, §6.1) exits `overall_rc=0`: the wrapper precheck passed, the harness reported 50 PASS, 0 FAIL and 0 UNMET, and the postcheck passed. The repository tree and status were unchanged across the run. B7 is fully used (7 of 7 cycles); see §10. An earlier S9 run failed and is recorded (§6.2). Only Codex can accept G002 or update PROGRESS.md. Nothing is committed: every G002 change is uncommitted on `staging` on top of the authorized baseline. No G003 artifact was created.

## 2. Execution record

### 2.1 Authorized binding and startup gates

| Item | Value |
|---|---|
| Baseline commit (`--base`) | `24002129e0ae07918264845eb08387fd276d36f0` ("Approve G002 web shell plan r6 and add visual design skill"), as authorized by Codex for the committed baseline |
| Approved-inputs snapshot | `/tmp/g002-r6-committed-baseline-inputs.sha256`, SHA-256 `010242e6df1c60f1cbec1234eb491fc86b1130a930218a352489f7d26b0c6881`, 20 entries (Goal, Spec, PROGRESS.md, Plan, the 16 design-skill files). The 20-entry binding supersedes Plan §17.2's one-entry alternative (Codex authorization). |
| Approved Plan | r6, `30501d82…7f41`, byte-identical to the committed Plan and bound by `--expect-hash` on every run |
| External S0a gate | `g002-s0-gate.sh` (refreshed), SHA-256 `26a24f1fe153e441042c2054f449b1d8cea4464b1f87899ab6468bc158ad606a`, kept outside the repository |
| Execution ID | `41ff07b7` |

- **S0a** exit 0: "HEAD 24002129…; snapshot 010242e6…; 20 inputs verified in the worktree and committed at HEAD; nothing differs from HEAD; no .env, no beginify resources, default ports free". No write of any kind preceded it.
- **S0b** created exactly `infra/validation/run-approved.sh` and `infra/validation/run-approved-controls.sh`; the status set equalled those two paths; shellcheck clean.
- **S0c** `run-approved-controls.sh` (W1–W9, 13 cases including the W5 sub-cases) exit 0; `run-approved.sh --check-only` exit 0; status unchanged.

### 2.2 S1 CLI and provenance

```sh
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -v "$PWD/web:/app" -v g002-web-nm-41ff07b7:/app/node_modules -w /app node:24-slim \
  sh -c 'npx -y shadcn@4.21.1 init -b base -p nova -y && npm install --save-exact shadcn@4.21.1 && npx -y shadcn@4.21.1 add button dialog -y && npx -y shadcn@4.21.1 eject -y'
```

- **First attempt (cycle 1):** `init` wrote a caret `shadcn` dependency that resolved to **4.21.2**, published the same day, so `eject` would have inlined unreviewed 4.21.2 CSS. The tree was restored to its S1 starting state and the CLI re-run with `shadcn@4.21.1` pinned exactly before `add` and `eject`.
- The written paths equalled the §16.2 S1 rows exactly. The ejected CSS block (`globals.css` lines 1–640) is byte-identical to the CLI's 4.21.1 output; its digest `16693ac5…c1e8` is pinned in B-WEB-TOKENS. `components.json` is kept as written. Raw output, the diff and the generated files are kept outside the repository (`g002-exec/s1-*`).
- `node_modules` lived only in the execution-owned volume `g002-web-nm-41ff07b7` (removed at cleanup, §7).

## 3. Implementation summary and files

The shell, components, tokens, appearance modes, browser suite and harness follow Plan r6 §3–§15. The four concerns are separate:
- **A, design language:** palette sources (`--light-*`, `--dark-*`), `--radius`, `--angle` and the type, container and spacing roles in `globals.css`;
- **B, theming:** resolution at `:root, [data-token-scope]` with the appearance-aware `dark` variant;
- **C, Organization overrides:** `TokenScope` with per-appearance sources and a scope portal container;
- **D, user appearance:** `localStorage["appearance-mode"]`, a classic inline head script and the `AppearanceControl` radio group.

`docs/foundation/WEB-SHELL.md` documents the contract, palettes, mapping, adaptations, provenance and tests.

### 3.1 Files changed (relative to `2400212`; equals the allow-list exactly)

- **Customized rendered files (5, registered):** `web/package.json`, `web/package-lock.json`, `web/app/globals.css`, `web/app/page.tsx`, `web/app/layout.tsx`.
- **Registered rendered file, changed content:** `docs/foundation/FOLLOW-UP.md` (STAGE-WEB-SHELL labels).
- **New web files (27):**
  - `web/components.json`; `web/lib/{utils,tokens,appearance}.ts`;
  - `web/components/ui/{button,dialog,portal-container}.tsx`;
  - `web/components/layout/{app-shell,token-scope,shell-help-dialog,appearance-control}.tsx`;
  - `web/playwright.config.ts`;
  - `web/tests/browser/{shell,overlay,appearance,accessibility,tokens,boundary,design}.spec.ts`;
  - `web/tests/browser/support/{color,png,page-errors,modes,contrast,angle}.ts`;
  - `web/tests/fixtures/token-scopes/{page.tsx,fixtures.ts}`.
- **New infrastructure (3):** `infra/docker/web-browser.Dockerfile`, its `.dockerignore`, `infra/compose/browser.yml`.
- **Validation tools:** new `infra/validation/run-approved.sh` and `run-approved-controls.sh`; modified `run.sh`, `lib.sh`, `render-customizations.tsv` (+5 rows).
- **Docs:** `README.md` (`platform:browser` block and WEB-SHELL link), new `docs/foundation/WEB-SHELL.md`.
- **Management:** this Result and `management/v0/results/G002-scope-allowlist.txt`. The Plan is unchanged.
- **Untouched (verified by B-SCOPE, B-PROTECTED and B-RENDER-PLAN):** `web/lib/{config,core-client}.ts`, `web/app/api/**`, web configs, the web Dockerfile, `local.yml`, `migrate.yml`, `core/**`, `intelligence/**`, `contracts/**`, `.env.example`, the manifest, both skills, governance and architecture docs.

### 3.2 Dependencies (exact pins; transitive attribution)

| Package | Pin | Category |
|---|---|---|
| `@base-ui/react` | 1.8.0 | dependencies |
| `class-variance-authority` | 0.7.1 | dependencies |
| `cn` | 0.4.0 | dependencies |
| `lucide-react` | 1.52.0 | dependencies |
| `tw-animate-css` | 1.4.0 | dependencies |
| `@playwright/test` | 1.63.0 | devDependencies |
| `@axe-core/playwright` | 4.13.0 | devDependencies |

- The 12 existing direct dependencies are unchanged (enforced by B-DEPS). `shadcn` is absent (removed by `eject`). There is no theming library and no font package.
- Transitive additions come from `@base-ui/react` (`@base-ui/utils`, `@floating-ui/*`, `@babel/runtime`, `reselect`, `use-sync-external-store`), `cn` and `class-variance-authority` (`clsx`), and the Playwright/axe test trees (`playwright`, `playwright-core`, `axe-core`).
- **Images:** `mcr.microsoft.com/playwright:v1.63.0-noble@sha256:eff16c30e6f3f4af0a03fa4b706120d5e9b0891c344a27d64559aff5900a4a27` (test-only) and the existing `node:24-slim`. Chromium in the runner: **153.0.8010.12**.

## 4. Migrations and schema

None. B-MIGRATE-SET is unchanged ({`00001_baseline.sql`}). No domain or tenant schema.

## 5. Contracts, API and MCP

None changed. Web routes remain exactly `/`, `/api/health` and `/api/ready` (B-WEB-BOUNDARY); core and intelligence are untouched; MCP is unchanged (B-MCP-SOURCE, B-MCP-HTTP).

## 6. Validation

### 6.1 Final acceptance run (S9)

```sh
bash infra/validation/run-approved.sh \
  --snapshot /tmp/g002-r6-committed-baseline-inputs.sha256 \
  --snapshot-sha256 010242e6df1c60f1cbec1234eb491fc86b1130a930218a352489f7d26b0c6881 \
  --count 20 \
  --path-re '^(management/v0/(goals/G002-web-shell\.md|specs/G002-web-shell-spec\.md|PROGRESS\.md|plans/G002-web-shell-plan\.md)|skills/vestive-design-skill/.+)$' \
  --plan management/v0/plans/G002-web-shell-plan.md \
  -- --base 24002129e0ae07918264845eb08387fd276d36f0 \
     --allow-file management/v0/results/G002-scope-allowlist.txt \
     --report-dir <scratchpad>/g002-exec/s9-final/report all
```

Evidence: `<scratchpad>/g002-exec/s9-final3/` (outside the repository): `harness.out`, `report/` (`results.tsv`, `deps-controls.tsv`, `wrapper-controls.log`, `browser-{root,off,rename}-results.json` and `-tests.tsv`, `plan.out`, `repeat-render.out`, `catalog-p.log`, `migrate-it.log`, `images.txt`, `inventory-leftover.txt`).

- **`overall_rc=0`.** The wrapper reported "precheck passed (20 inputs, snapshot 010242e6…)", then "harness exit 0; postcheck passed".
- **Harness: RESULT PASSED. 50 PASS, 0 FAIL, 0 UNMET**, plus INFO lines (PREFLIGHT docker 27.5.1 and compose 2.32.4; PROBE devtcp; IMAGES).
- **Every check:**
  - **Static:** B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-LITERALS, B-NAMING, B-DEPS, B-DEPS-CONTROLS, B-WRAPPER-CONTROLS, B-WEB-BOUNDARY, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, B-MCP-SOURCE, B-MIGRATE-SET, B-DOCKERIGNORE.
  - **Provenance:** B-RENDER-ID, B-RENDER-PLAN and B-RENDER-REPEAT (exactly 8 registered conflicts and proposals), B-PROTECTED (203 protected files; 20 expected-hash files), B-SKILL-UNCHANGED.
  - **Root:** B-CONFIGURE, B-DEFAULTS, B-PORTS-ROOT, B-BOUNDARIES-ROOT, B-BOUNDARY-CONTROLS, B-MCP-HTTP, B-WEB-HTTP, B-BROWSER-ROOT (95 tests), B-CHECKS-ROOT, B-DOCS-EXEC-GS, B-BUILD-CONTEXT, B-DOCS-EXEC-ROOT.
  - **Migrations:** B-MIGRATE-CLI, B-MIGRATE-CATALOG, B-MIGRATE-ENV, B-MIGRATE-IT (10/10).
  - **Variants:** B-OUTAGE, B-PERSIST, B-SOURCE-HASH, B-PORTS-OVERRIDE, B-RENAME, B-BROWSER-RENAME (2 `@identity` tests), B-DBRENAME, B-TOGGLE-OFF-ADDITIONS (28 applicable additions), B-BROWSER-OFF (95 tests), B-TOGGLE-OFF.
  - **Final:** B-DOCKER-ONLY, B-SECRET-CANARY (48 generated secrets; planted-secret control detected), B-SCOPE (44 changed paths, all allow-listed).
- **Browser runs inside the harness.** Root and disabled variant each ran 95 tests with 0 skipped, flaky or unexpected. Every scenario ID passed in each required project, S-ERRORS was annotated on every test, and the served `appearance-init` SHA-256 equals WEB-SHELL.md (`d5054b00…3e42`).
- **No repository edits during the run.** The tree digest (excluding `.git`) and the NUL status digest were identical before and after.

### 6.2 Run history (every run, including failures)

| Run | Scope | Outcome |
|---|---|---|
| S0a | external gate | exit 0 |
| S0c | W1–W9 controls; `--check-only` | exit 0; exit 0 |
| S2/S3 | `web-check`, production build | pass (after two lint/type fixes in S4: an unused import and a duplicate object key; not counted) |
| Browser run 1 (S4) | full suite | 19 failures → cycle 2 |
| Browser run 2 | full suite | 85 passed, 10 failures → cycle 3 |
| Browser run 3 | full suite | 91 passed, 4 failures (same cycle 3; two further test-sequencing fixes) |
| Browser run 4 | full suite | 95/95 passed |
| S5 gate 1 | static + provenance via the wrapper | 23 PASS, 1 FAIL (B-WEB-TOKENS: a sort-order defect in the new check) → cycle 4; `overall_rc=1` |
| S5 gate 2 | static + provenance via the wrapper | 24 PASS, 0 FAIL; `overall_rc=0` |
| S7 | root + migrations + variants via the wrapper | all PASS except B-BROWSER-ROOT and B-BROWSER-OFF: S-OVERLAY initial-focus race (1 test, then 2) → cycle 5; `overall_rc=1` |
| Cycle 5 retest | S-OVERLAY ×30 (`--repeat-each 15`, both shell projects); full suite with the harness matrix | 30/30; 95/95, matrix and init hash pass |
| Cycle 6 retest (after S8 review) | `web-check`; full suite with the harness matrix | pass; 95/95, matrix and init hash pass |
| S9 run 1 | `all` via the wrapper | 48 PASS, 2 FAIL (B-DEPS, B-DEPS-CONTROLS: `DEPS_LOCK_MISMATCH @types/react` in every case); `overall_rc=1`; postcheck passed; no leftover resources → cycle 7 |
| Cycle 7 retest | static + provenance via the wrapper | 24 PASS, 0 FAIL; C1–C12, C8b and C8c exact; `overall_rc=0` |
| S9 run 2 (aborted) | `all` via the wrapper | Stopped by me about 1 minute in, during static, to correct a WEB-SHELL.md sentence about the `devOptional` rule that cycle 7 had made inaccurate, so that the final run validates the exact handed-over tree. The only leftover was its own work directory, which I removed; no Docker resources were left. |
| S9 run 3 (final) | `all` via the wrapper | **50 PASS, 0 FAIL, 0 UNMET; `overall_rc=0`**; precheck and postcheck passed; tree and status unchanged; no leftovers |

### 6.3 Tests added

- **Browser suite** (95 tests; Chromium; `retries: 0`; one worker), in six projects. Every title carries a scenario ID, and the harness requires each ID to pass in each listed project:
  - **`shell-light`, `shell-dark`** (the running web): S-IDENTITY, S-LANDMARKS, S-NAV, S-SKIP, S-TAB, S-OVERLAY, S-MOTION, S-RESPONSIVE (5 viewports), S-A11Y, S-CONTRAST, S-BOUNDARY, S-DESIGN, S-ANGLE, S-SCREENSHOTS; `shell-light` also S-HEAD-ORDER, S-HEALTH, S-NOFIXTURE, S-FAILCLOSED.
  - **`tokens-light`, `tokens-dark`** (fixture build): S-TOKENS (all 16 tokens in four scopes), S-PORTAL (including a live mode change), S-RUNTIME, S-A11Y, S-CONTRAST, S-ANGLE, S-SCREENSHOTS.
  - **`appearance`**: S-APPEARANCE-DEFAULT, -LIVE, -OVERRIDE, -RETURN, -PERSIST, -INVALID, -NOSTORAGE; S-FIRSTPAINT (8 cases); S-FIRSTPAINT-DELAYED (2); S-HTML-ATTRS (6).
  - **`dev-hydration`** (`next dev`): S-HYDRATION-DEV (6).
  - **S-ERRORS** applies to every test: any `pageerror` or `console.error` fails it, and a passing test carries an `S-ERRORS` annotation that the harness requires on every test of the five non-dev projects.
- **Harness checks:** B-DEPS (reason-coded), B-DEPS-CONTROLS (C1–C12 plus C8b and C8c), B-WRAPPER-CONTROLS (W1–W9), B-WEB-BOUNDARY, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, B-WEB-HTTP, B-BROWSER-ROOT, B-BROWSER-RENAME, B-BROWSER-OFF; extended B-SHELL, B-HADOLINT, B-COMPOSE, B-NAMING, B-DOCKERIGNORE (unchanged list), B-BUILD-CONTEXT, B-OUTAGE, B-SECRET-CANARY (with a planted-secret positive control), B-SKILL-UNCHANGED and B-SCOPE (NUL-safe), B-RENDER-PLAN (exactly 8 registered conflicts).

### 6.4 Dependency controls (B-DEPS-CONTROLS; final run)

Expected and actual code sets (`;`-joined), from `deps-controls.tsv`:

| Control | Expected (must fail) | Actual | Verdict |
|---|---|---|---|
| C1 | DEPS_EXISTING_CHANGED next | DEPS_EXISTING_CHANGED next | pass |
| C2 | DEPS_EXTRA_DIRECT left-pad | DEPS_EXTRA_DIRECT left-pad | pass |
| C3 | DEPS_ADDITION_WRONG @axe-core/playwright | DEPS_ADDITION_WRONG @axe-core/playwright | pass |
| C4 | DEPS_EXISTING_MISSING server-only | DEPS_EXISTING_MISSING server-only | pass |
| C5 | DEPS_ADDITION_WRONG @base-ui/react; DEPS_NOT_EXACT @base-ui/react | same | pass |
| C6 | DEPS_LOCK_MISMATCH next | DEPS_LOCK_MISMATCH next | pass |
| C7 | DEPS_LOCK_MISMATCH next | DEPS_LOCK_MISMATCH next | pass |
| C8 | DEPS_LOCK_MISMATCH @playwright/test | DEPS_LOCK_MISMATCH @playwright/test | pass |
| C8b (added) | DEPS_LOCK_MISMATCH typescript | DEPS_LOCK_MISMATCH typescript | pass |
| C8c (added) | DEPS_LOCK_MISMATCH typescript | DEPS_LOCK_MISMATCH typescript | pass |
| C9 | DEPS_EXISTING_RECATEGORIZED typescript | DEPS_EXISTING_RECATEGORIZED typescript | pass |
| C10 | DEPS_EXTRA_DIRECT shadcn; DEPS_FORBIDDEN shadcn | same | pass |
| C11 | DEPS_IMAGE_MISMATCH | DEPS_IMAGE_MISMATCH | pass |
| C12 (positive) | ∅ (passes) | ∅ (passed) | pass |

### 6.5 Accessibility notes

- axe (`wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa`, `wcag22aa`): zero violations everywhere. The `incomplete` results are `color-contrast` (always; axe cannot resolve the angled pseudo-element fills and translucent layers) and `aria-hidden-focus` (with the dialog open; Base UI marks the background `inert`). Both are covered by measurement instead: S-CONTRAST measures rendered pixels, and S-OVERLAY proves Tab containment and that header, main and footer are inert.
- **WCAG mapping** (Plan §12): 1.3.1, 1.4.1, 1.4.3, 1.4.4 and 1.4.10 (S-RESPONSIVE: 320 px, 200% zoom and text), 1.4.11 (S-CONTRAST non-text pairs), 1.4.13 (no hover content), 2.1.1, 2.1.2, 2.4.1 (S-SKIP), 2.4.3 (S-TAB, S-OVERLAY), 2.4.7, 2.4.11 (S-TAB unobscured), 2.5.8 (S-TAB and the dialog's Close), 2.3.3 via reduced motion (S-MOTION), 2.4.2 (S-IDENTITY title), 3.1.1 (S-HTML-ATTRS `lang`), 4.1.2 (S-LANDMARKS, S-OVERLAY name and role).
- **Not automated:** 1.4.12 text spacing. Buttons expand horizontally and allow taller line boxes, but no scenario applies the 1.4.12 spacing override. It is listed as follow-up (§12).

## 7. Cleanup and Docker inventory

- **Harness:** each run took its own start inventory and removed only the projects and images it created; `inventory-leftover.txt` is empty in the final run.
- **Execution-owned resources, removed after S9:**
  - the `beginify-*` images built by my S4–S8 root-stack runs: `core`, `intelligence`, `minio`, `minio-init`, `web`, `web-check`, `web-browser`;
  - `g002-41ff07b7-web` and `g002-41ff07b7-web-check`;
  - the labelled volume `g002-web-nm-41ff07b7` (label `g002.exec=41ff07b7`, verified before removal);
  - the empty root-owned `web/node_modules` mountpoint that the S1 volume mount created (removed through a busybox container).
- **Root `.env`:** the two I created for S4 and the cycle-5/6 debugging were removed only after their SHA-256 matched the value recorded at creation. No `.env` exists now. The harness's root suite creates and removes its own `.env` by the same rule.
- **Inventory diff against S0a:** the container, volume, network and image inventory equals the S0a inventory exactly (no addition, no removal). Unrelated developer resources were untouched.
- **Repository:** no ignored files; no `.env`; the working tree holds exactly the 44 allow-listed changes plus this Result.

## 8. Acceptance criteria

| AC | Result | Evidence |
|---|---|---|
| AC1 shadcn/Base UI, attributable dependencies | Met (evidence for Codex review) | §2.2 CLI provenance (4.21.1, ejected block pinned), E1–E7 (B-WEB-COMPONENTS), §3.2 pins, B-DEPS with independent reason codes, B-DEPS-CONTROLS 14/14 exact sets, `npm ci` in every image build, production build, B-NAMING, B-PROTECTED |
| AC2 usable generic shell | Met (evidence for Codex review) | S-IDENTITY, S-LANDMARKS (no login/account/tenant names), S-NAV, S-RESPONSIVE (320×640, 768×1024, 1280×800, 640×400, 200% text) in both appearances, S-ERRORS; B-BROWSER-ROOT |
| AC3 keyboard and accessibility | Met (evidence for Codex review) | S-SKIP, S-TAB (order, unobscured, ≥24 px, ring ≥3:1), S-OVERLAY (Enter/Space, `aria-modal`, name, focus in, ×10 containment, inert background, Escape/Close, focus return), S-MOTION, S-A11Y, S-CONTRAST (rendered pixels, rest/hover/focus), S-ANGLE, all in both appearances |
| AC4 overrides without rewrite or leakage | Met (evidence for Codex review) | S-TOKENS (16 tokens × 4 scopes; adjacent default equals `:root`), S-PORTAL (popup and backdrop mount in A/B, A's values, live mode change), S-RUNTIME (immediate change; B and both defaults unchanged in both appearances), S-SCREENSHOTS (fixtures, dialogs in A and B), B-WEB-TOKENS, WEB-SHELL contract |
| AC5 identity, BFF, health | Met (evidence for Codex review) | B-RENAME and B-BROWSER-RENAME (image IDs unchanged); B-WEB-BOUNDARY; S-BOUNDARY (same-origin only; no internal names in HTML/JS/CSS; bodies saved for the secret scan); S-HEALTH; B-WEB-HTTP; B-OUTAGE (503 and 200 with no-store); S-FAILCLOSED; S-NOFIXTURE; B-SECRET-CANARY |
| AC6 repeatable Docker checks | Met (evidence for Codex review) | README blocks executed verbatim (B-DOCS-EXEC-ROOT, B-CHECKS-ROOT, B-BROWSER-ROOT); pinned images; S0a; S0c; B-WRAPPER-CONTROLS; final `overall_rc` (§6.1); B-DOCKER-ONLY; cleanup inventory (§7); IMAGES |
| AC7 provenance and regressions | Met (evidence for Codex review) | Register +5 (8 total); B-RENDER-ID, B-RENDER-PLAN and B-RENDER-REPEAT (exactly 8); B-PROTECTED; B-SKILL-UNCHANGED (both skills); root, ports, boundaries, migrations, outage, persistence and rename regressions; B-TOGGLE-OFF-ADDITIONS (28 applicable additions), B-TOGGLE-OFF, B-BROWSER-OFF; B-SCOPE with the exact allow-list |
| AC8 scope, invariants, handoff, budget | Met (evidence for Codex review) | §3.1 equals the allow-list; no migrations or schema; WEB-SHELL identity invariants (UUID versus slugs, names and infrastructure keys); FOLLOW-UP labels; this Result; §10 cycles and the cycle-4 checkpoint |
| AC9 Light, Dark and System; System default | Met (evidence for Codex review) | S-APPEARANCE-* (7 scenarios), S-HEAD-ORDER (classic inline script, hash equal to WEB-SHELL.md, checked by the harness), S-FIRSTPAINT (8 cases: paint-timed frames, LCP, palette and H1 ink per frame), S-FIRSTPAINT-DELAYED (framework JS held; stored appearance opposite the OS in painted ground and H1 ink), S-HTML-ATTRS, S-HYDRATION-DEV, S-ERRORS; R2/R3 scenarios in both resolved appearances |
| AC10 design language within constraints | Met (evidence for Codex review) | WEB-SHELL inventory, mapping, T1–T13 and the further adaptations; S-DESIGN (radius, 2px rules, display type, kicker, angle, hairline-only dialog, no blur, bullets, muted band, container, alignment), S-ANGLE, S-SCREENSHOTS (Light and Dark, with the dialog open), B-WEB-DESIGN, B-SKILL-UNCHANGED; SA-V review (§9.3) |

## 9. Deviations, residuals and review findings

### 9.1 Diff inspection

The changed set relative to `2400212` equals the allow-list exactly (B-SCOPE). The Plan, both skills, the manifest and all protected files are unchanged (B-PROTECTED, B-SKILL-UNCHANGED, the wrapper postcheck).

### 9.2 Deviations from the Plan (none architectural; Codex decides)

1. **`--angle` value.** Plan §7.2 gives `polygon(15% 0%,100% 0%,100% 35%,85% 100%,0% 100%,0% 65%)`. The implementation uses a constant 12px cut, `polygon(12px 0%, 100% 0%, 100% 35%, calc(100% - 12px) 100%, 0% 100%, 0% 65%)`, because the 15% cut clipped label corners on wider buttons (S-ANGLE, cycle 2). The token name and mechanism are unchanged. S-DESIGN compares the computed `clip-path` with the token, so the value itself is documented rather than pinned.
2. **Destructive variant (E1/E3 as written).** It uses a tinted angled fill (`destructive` at 10%/15% Light, 20%/25% Dark) with destructive ink and a destructive focus outline (`focus-visible:outline-destructive`), so that the label stays ≥4.5:1 at rest and on hover (cycle 2). It uses the same outline mechanism as E1, with the destructive colour.
3. **Two additional test-support modules,** `web/tests/browser/support/contrast.ts` and `angle.ts` (27 new web files rather than §18's 25). They hold the rendered-pixel contrast and angle-geometry helpers. Both are in the allow-list.
4. **B-DEPS rule (c) and `devOptional`.** Plan §15 says dev-only dependencies carry `dev: true`. npm instead marks a dev dependency `devOptional` when a runtime package also declares it as an optional peer. In this lockfile that applies to exactly two direct dev dependencies:
   - `@playwright/test`, an optional peer of `next`;
   - `@types/react`, an optional peer of `@base-ui/react`, `@base-ui/utils` and `lucide-react`. It was `dev: true` at the baseline.

   Rule (c) accepts `devOptional` as dev-only **only when a non-dev package in the same lockfile declares that name in `peerDependenciesMeta` as optional**; otherwise it is a mismatch.
   - C8 therefore removes that marking from `@playwright/test`; removing an absent `dev` would be a no-op.
   - Two controls are added: **C8b** removes `dev` from `typescript`, and **C8c** turns `typescript`'s `dev` into an unjustified `devOptional`. Both must fail, and both do.
5. **B-WEB-TOKENS literal scope.** It scans `components/**`, `lib/**`, `app/**` (`.ts`/`.tsx`) and the customized region of `globals.css`, where literals may appear only in the palette block. Two areas are explicit exceptions:
   - the verbatim ejected registry block, which contains `#000` mask stops and `oklch(from currentColor …)`; it is pinned by digest instead;
   - `tests/**`, which restate expected palette values as independent assertion oracles, so that tests cannot pass by reading the implementation.

   `tests/fixtures/**` holds the override values, as planned.
6. **B-WEB-BOUNDARY comment lines.** As in B-MCP-SOURCE, the source rules ignore lines that are entirely comment. `lib/config.ts`, which is untouched rendered code, mentions `NEXT_PUBLIC_*` in a comment. A comment followed by code on the same line is still matched.
7. **S-DESIGN dialog shadow.** §9.2 approves the `ring-1 ring-foreground/10` hairline; §11.2 says "no box-shadow". Tailwind draws the ring as a box-shadow, so S-DESIGN allows only zero layers plus exactly one `0 0 0 1px` layer: no offset, blur or spread beyond 1px.
8. **Test sequencing against Base UI.** These are waits; the final assertions are unchanged.
   - Dialog Tab containment waits two animation frames after each key. In between, focus may rest only on this dialog's own focus guard, which Base UI moves back inside on the next frame.
   - Initial focus and background inertness are polled.
   - Focus-ring reads wait for the element's transitions.
   - The responsive test waits for the closing backdrop to detach.
   - Screenshots of the open dialog wait for its fade-in.
9. **Browser-runner fonts.** The Playwright image's fontconfig resolved the generic families to a CJK face without bold, so headings rendered at regular weight in evidence screenshots. The runner image maps `sans-serif`, `system-ui` and `monospace` to its own Liberation faces (test environment only; no download). Product CSS is unchanged (T3 updated).
10. **Shell additions within the approved design** (from the SA-V review):
    - the "Keyboard help" trigger uses `h-9 px-5`;
    - the appearance labels use `px-2` below `sm`, so the group aligns with the 320 px content edge;
    - the "Getting around" section is the `muted` band that §7.1/§9.2 map. S-DESIGN asserts it.
11. **Skip link, `aria-modal`.** The skip link is `fixed` and translated off-screen until focused; `not-sr-only` reset its padding, and an absolutely positioned link scrolled out of view. Base UI does not set `aria-modal`, so `ShellHelpDialog` sets it explicitly; true modality is proven by containment and inertness.
12. **Final `--base`.** It is the committed baseline `2400212` with the 20-entry snapshot, per Codex's authorization, instead of §17.2's `cc157ed`.

### 9.3 Read-only reviews (S8: SA-R and SA-V; 0 cycles each)

All findings were triaged; the fixes were integrated and validated in cycle 6.

**SA-R (accessibility, boundary, harness, provenance):** no blockers.

| # | Finding | Disposition |
|---|---|---|
| R1 (major) | S-BOUNDARY bodies existed only base64-encoded in the JSON report, so B-SECRET-CANARY could not see them | Fixed: bodies are written as `responses.txt` in the output directory; B-BROWSER-ROOT requires them for both shell projects; B-SECRET-CANARY gained a planted-secret positive control |
| R2 (major) | S-RUNTIME checked the defaults only with `toBeTruthy` | Fixed: the default, adjacent-default and B fills are compared before and after, in both appearances |
| R3 (major) | `--angle` and the destructive variant differ from §7.2/E1–E3 as written | Recorded as deviations 1–2 |
| R4 | `devOptional` accepted globally | Fixed. The cycle-6 fix (a name allow-list containing only `@playwright/test`) missed `@types/react`, and S9 run 1 caught it. Cycle 7 replaced the allow-list with the lockfile's own optional-peer evidence, plus control C8c (deviation 4). |
| R5 | Comment-line exclusion bypassable (`/* x */ code`) | Fixed: only lines that are entirely comment are dropped (exercised with bypass samples) |
| R6 | H1 colour recorded but not asserted; the delayed case could not discriminate by primary | Fixed: every content frame asserts H1 ink; the delayed case samples painted H1 ink against the stored and OS palettes |
| R7 | S-MOTION backdrop fallback; button transitions unchecked | Fixed: the backdrop is required; the trigger's transition is checked under both motion settings |
| R8 | No count guards (nav links; body types) | Fixed: exactly 3 nav links; document, script and stylesheet bodies are required |
| R9 | `cssToRgba` returned black for an invalid colour | Fixed: it throws on an empty or unparsable colour |
| R10 | `expectFocusRing` passes on one side | Kept. The ring must be present and ≥3:1 on at least one side whose neighbour is the surrounding surface; segmented-group neighbours legitimately paint over a side. Recorded here. |
| R11 | S-ANGLE used the border box | Fixed: padding box (`clientLeft`/`clientTop`/`clientWidth`/`clientHeight`) |
| R12 | S-TOKENS checked 9 of 16 tokens | Fixed: all `SEMANTIC_COLOR_TOKENS` |
| WCAG mapping | 1.4.12 not automated; 2.4.2/3.1.1 implicit; 2.5.8 Close | 2.5.8 asserted for Close; mapping in §6.5; 1.4.12 is follow-up |
| Notes | `[suppresshydrationwarning]` DOM count always 0; ejected-block exception; B-SCOPE newline paths; final base | Acknowledged. B-WEB-BOUNDARY is the real source check. The exception is recorded (deviation 5). A newline-in-path case is contrived (paths are NUL-enumerated; matching is per line). The base is as authorized (deviation 12). |

**SA-V (design-language fidelity):** no blockers.

| # | Finding | Disposition |
|---|---|---|
| V-M1 (major) | The `-dialog` screenshots captured the first fade-in frame, so no dialog was visible | Fixed: wait for animations and opacity 1; capture with animations disabled; verified visually |
| V-m1 | Display type at regular weight | Root cause was the runner's CJK fallback face; fixed in the runner (deviation 9); verified visually |
| V-m2 | Angled CTA small, label near the cut | Trigger `h-9 px-5` |
| V-m3 | Appearance group past the 320 px content edge | `px-2` below `sm` |
| V-m4 | Two primary fills side by side in the header | Kept: §9.2 assigns `primary` to the filled action and the selected appearance; the selected state also carries the 14×2 bar |
| V-m5 | Band surface unused | `#getting-around` is the `muted` band (S-DESIGN asserts it) |
| V-n1–n3 | Hairline versus 2px rule; stray focus ring in the fixture screenshot; Dark secondary label not documented | Hairline kept (E6/§9.2); focus blurred before full-page captures; WEB-SHELL documents the Dark label |
| Clean | §9.2 mapping in both appearances, T1–T13 and the further adaptations, no identity/marketing/assets, nothing forces Dark, tokens rather than literals | — |

## 10. Execution budget (B7)

### 10.1 Cycle log

| Cycle | Trigger | Work |
|---|---|---|
| 1 | S1 | `shadcn` caret drift to 4.21.2 → restore and re-run with 4.21.1 pinned (counted conservatively) |
| 2 | Browser run 1, 19 failures | 12px angle cut; skip-link geometry; destructive hover tints; `aria-modal`; trigger locators; axe after fade-in; PERSIST sequencing |
| 3 | Browser run 2, 10 failures | Focus-guard frame (containment); transition-settled ring reads; skip link `fixed`; hairline-only shadow check; `allSettled`; closing-backdrop wait → run 4 95/95 |
| 4 | S5 gate, B-WEB-TOKENS | Sort-order defect in the new check (harness only) |
| 5 | S7, S-OVERLAY | Initial dialog focus polled (Base UI focuses on a later frame); 30/30 repeats |
| 6 | S8 review findings | The SA-R and SA-V fixes in §9.3. The retest was `web-check` and the full browser suite with the harness matrix; the static suite was not re-run (see the process error below). |
| 7 | S9 run 1: B-DEPS and B-DEPS-CONTROLS | Cycle 6's `devOptional` allow-list was too narrow (`@types/react` is also `devOptional`). Replaced with optional-peer evidence from the lockfile, plus control C8c. Static and provenance retested (24 PASS, 14/14 controls exact), then S9 re-run. |

**Used: 7 of 7. B7 is fully used at handoff; no further cycle is available.**

- **Convergence.** The failures converged: browser 19 → 10 → 0, harness 1 → 0, S7 2 → 0. The only issue that needed a second fix was the `devOptional` handling (cycles 6 and 7).
- **Process error, disclosed.** After changing `lib.sh` in cycle 6, I retested only `web-check` and the browser suite, not the static suite that exercises `lib.sh`. That is how the cycle-6 regression reached the final run, and it cost cycle 7.

### 10.2 Cycle-4 checkpoint (BUDGET_POLICY §7; recorded at the opening of cycle 4)

- **Completed:** S0a–S0c; S1–S4; browser suite 95/95 (run 4); S5 static and provenance PASS except B-WEB-TOKENS (a harness comparison defect, not product).
- **Remaining:** S5 retest; S6 docs; S7 root, migrations and variants with the new browser, HTTP, outage and build-context checks; S8 reviews; S9.
- **Repeated failures:** none. **Uncertainty:** the first execution of the root and variant browser integration (S7).
- **Forecast:** 5–6 of 7. No thrashing. Decision: continue.

## 11. Unresolved issues and blockers

- **No blockers.**
- **Recorded deviations for Codex to decide:** §9.2 items 1 (`--angle`), 2 (destructive variant), 3 (two extra support modules), 4 (`devOptional` evidence rule) and 5 (literal-scan scope).
- **Limitations:**
  - WCAG 1.4.12 is not automated.
  - The condensed display faces depend on locally installed fonts (T3).
  - `expectFocusRing` requires the ring on at least one side, not all four (R10).
  - Only Chromium is tested (as planned).
- **Budget:** B7 is fully used. A correction loop would need a budget decision from Codex.

## 12. Suggested follow-up (does not authorize scope)

- Automate WCAG 1.4.12 text spacing (a spacing-override stylesheet plus the S-RESPONSIVE overlap and clipping assertions).
- CSP with the `appearance-init` hash allow-listed (G012).
- Organization branding data by Organization UUID, with validation of untrusted values, feeding `TokenScope` (G016/G017).
- A product-owner decision on self-hosting the OFL display faces (Oswald, Barlow Condensed). It would be a new asset and licence decision; until then the condensed display type depends on locally installed fonts.
- Optional: an inverted selected state for the appearance control, if the product owner prefers a single primary block per header (SA-V m4).
