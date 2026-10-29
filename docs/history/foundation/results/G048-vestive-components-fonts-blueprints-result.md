> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G048 — Vestive components, fonts and blueprints: implementation result

Goal: [G048-vestive-components-fonts-blueprints.md](../goals/G048-vestive-components-fonts-blueprints.md) · Spec: [G048-vestive-components-fonts-blueprints-spec.md](../specs/G048-vestive-components-fonts-blueprints-spec.md) · Plan: [r2](../plans/G048-vestive-components-fonts-blueprints-plan.md) (SHA-256 `a3500b7e64d4c7282bfdacc990eddf5646aa6bad35933c7e3542cd5d1708d46e`) · Reconciliation: G048-execution-reconciliation.md (not included) (`664eacc1…eb7e`) · Source rights: G048-source-rights.md (not included) (`685dd4f5…0ca9`)
Owner: Claude Code · Reviewer: Codex · Budget: **B5, extended to six total cycles** (§10)

## 1. Status

**SUBMITTED FOR CODEX ACCEPTANCE REVIEW — not accepted. The CORRECTION LOOP (AC9 hosted evidence only) is complete: hosted pipeline 2922479202 passed fast and the full gate on the exact delivery commit (§7.3, correction loop 1).** The final local validation passed on the final tree.

The first hosted attempt is retained as history. API pipeline 2922044054 failed both jobs with `ci_quota_exceeded` before any runner started. After the owner purchased minutes, Codex's review (`G048-implementation-codex-review.md`, `4c86246d…1bec`) directed one rerun with no code change. Only Codex ACCEPT completes G048. This Result does not edit PROGRESS, the Goal, the Spec or the Plan.

- The six source modules run on Radix Dialog and Slot, and Base UI is removed. The internal text roles, the three bundled font families and an original blueprint drawing are in place. The guards are reconciled.
- **Final local validation:** s8b `all` (run `3d9d8a`, final tree) passed with overall 0 (§7.1).
  - 70 PASS and 3 INFO, with all 69 required IDs PASS.
  - Browser: root 133/133, disabled 133/133, rename 2/2.
  - B-CI-CONTROLS 203/203.
  - The §11.8 pre/post checks reported `fails=0`.
- **Hosted:** delivery commit `40bfebf9` (tree `2c21faab`) was pushed to `g048-vestive-adoption` (§7.3).
  - Attempt 1, API pipeline 2922044054: both jobs failed with **`ci_quota_exceeded`** before any runner started.
  - Correction loop 1, API pipeline **2922479202**: **success**.
    - fast job 17001434264: 32/32 required, pipeline full, executed fast;
    - gate job 17001434266: **full**, 70/70 required (70 PASS, 3 INFO), browser 133/133/2, B-CI-CONTROLS 203;
    - both jobs: attribution `api`/`branch-api`, final status 0, credential scan clean, `guard=ok`.
- **Cycles: 6 of 6 consumed** (B5, plus one owner-added cycle). The checkpoint after cycle 3 is recorded (§10). No budget remains.
- **Scope additions:**
  - the VP1 count, owner- and supervisor-authorized (§9.1);
  - the API-pipeline CI correction, owner-requested and Codex-authorized (§9.4);
  - the reviewed exceptions X19 (widened) and X23 (new) (§9.2).
- G004, G005, G006 and OS001 work is not part of this Result. The G021 foundation barrier stays closed.

## 2. Execution binding and starting state

| Item | Value |
|---|---|
| Binding | `~/.local/state/platform-evidence/g048-execution-3bb388f/binding.json`, SHA-256 `b2dd66dc22ae77bf66121a24f567bb1d0bb3de6e48f2ff2fbdd51aa3820f0224` (status "ACCEPT; G048 implementation and named hosted evidence actions authorized by owner on 2026-10-07 UTC") |
| BASE | `3bb388fc68d3a339238a30156e721f97c15f6044` |
| Checkout / branch | `.claude/worktrees/g048-exec` / `worktree-g048-exec` |
| Snapshot | `approved-inputs.sha256`, SHA-256 `4d2963b909201b00f93d1f317e4154f33822d0b51dc95cd7eca932259fff3555`, 30 entries (binding `RE`, `PLAN`, `ALLOW`) |
| Evidence root | `EXEC_ID=876a2392`; `EV=~/.local/state/platform-evidence/g048-876a2392` (outside the checkouts and `/tmp`, mode 700) |
| Wrapper | `infra/validation/run-approved.sh` (G047 Plan r2 §8.7.2 status precedence, reused through `g048-vw.sh`); every run row is in `$EV/runs.tsv` |

**S0 (read-only, all passed before the first write; `$EV/s0.txt`):**
- binding hash verified;
- top level, HEAD = BASE, branch, clean status;
- all 30 snapshot entries committed at BASE with matching hashes, plus the Plan r2 hash and the P3 rights record;
- kit manifest 63 OK, all committed;
- `run-approved.sh --check-only`: precheck passed (30 inputs);
- no `.env`; the 34 default/variant ports free; no root-project resources (Docker inventory recorded).

**S0b:** the allow-list (Plan §9.2, lines 1330–1350 byte for byte) was the first repository write, SHA-256 `6f2559292606015177e713f0225fafa742a15a949f56179b11b437432114b89f`. Two authorized additions followed:
- the VP1 line (§9.1), giving `d1a4669fc361f7d5fd802711eae10177b9605fe65f4db5bca77ad0b6baede05b`;
- the API-pipeline supplement's comment and three lines (§9.4), giving the final `6b36af4231dc186774061e8c57d226f045ba9eb5316e947903529c269e9ed990`.

**S1 baseline** (`s1 all`, run `e8b82a`, unmodified BASE plus the allow-list):
- wrapper 0, tee 0, postcheck passed, overall 0;
- 69 PASS and 3 INFO;
- browser 103 root, 103 disabled, 2 rename;
- duration 3307 s; no leftover resources.

The before screenshots are embedded in `$EV/s1/report/browser-{root,off}-results.json`.

## 3. Implementation summary and files changed

**56 paths:**
- the Plan §9.1 inventory (51);
- `infra/ci/validate-controls.sh` (§9.1 VP1);
- `.gitlab-ci.yml`, `infra/ci/validate.sh`, `infra/ci/select-profile.sh` and `docs/foundation/OPERATIONS.md` (§9.4, the API-pipeline supplement). Its other two paths, `validate-controls.sh` and `infra/validation/run.sh`, were already in the inventory.

| Area | Paths | Change |
|---|---|---|
| Source modules (6) | `web/components/ui/{typography,layout,bullet-list,site-header,button,dialog}.tsx` | Adapted from the kit with lineage headers. Translations T1–T14 and exceptions X1–X23 (WEB-SHELL.md). `button.tsx` uses Radix Slot; `dialog.tsx` is a client module on Radix Dialog with the scoped `DialogPortal` adapter. Module diffs: `$EV/module-diffs/*.diff` (§12, hunk map). |
| Portal and scope | `web/components/ui/portal-container.tsx`, `web/components/layout/token-scope.tsx` | `{ element }` context value. TokenScope keeps its container in callback-ref state (Plan §5.4). |
| Shell | `web/components/layout/{app-shell,shell-help-dialog,blueprint-backdrop}.tsx`, `web/app/page.tsx` | SiteHeader composition; TextLink skip link; the help trigger as `DialogTrigger asChild` → `Button size="sm"`; BulletList tips; overview, ruled grid and About from the adopted modules; the `BlueprintBackdrop`; `overflow-wrap:anywhere` at the configured-identity call sites (X19). |
| Tokens and styles | `web/app/globals.css` (lines 1–641 unchanged, digest pinned), `web/lib/tokens.ts`, `web/components.json` | `--light/--dark` muted-foreground (`#615b55`/`#8f8a84`). Internal roles `ink`, `text-2`, `text-3` resolved at every scope; `INTERNAL_ROLE_TOKENS`, with `tokenOverrideStyle()` deriving them from a foreground override. The source `--angle`; `--spacing-section` 88px and `--spacing-cell` 36px; the source font stacks. Four `@font-face` rules (section F); the `[data-blueprint]` layer (section E). The `angled` utility is removed. `components.json` style `radix-nova`. |
| Dependencies | `web/package.json`, `web/package-lock.json` | `@base-ui/react` removed; `@radix-ui/react-dialog` 1.1.23 and `@radix-ui/react-slot` 1.3.3 added. The lockfile delta equals Plan §5.1 exactly (§7.4). |
| Assets and notices | `web/public/fonts/**` (7), `web/public/blueprint/overview.svg`, `web/THIRD_PARTY_NOTICES.md` | Intake as Plan §7.1 (§7.4); the original drawing (§8); the shadcn/ui MIT notice, Radix (MIT) and the three OFL notices. |
| Images | `infra/docker/web.Dockerfile`, `infra/docker/web-browser.Dockerfile` | `COPY` of `public/` into the runtime image and the fixture build. |
| Browser tests | `web/playwright.config.ts`, `web/tests/browser/{accessibility,appearance,design,overlay,shell,tokens}.spec.ts`, `web/tests/browser/support/{fonts,source-reference,assets,png,contrast,angle,modes,spacing}.ts`, `web/tests/fixtures/token-scopes/page.tsx` | §6. |
| Harness | `infra/validation/{lib.sh,run.sh,render-customizations.tsv,required-checks.tsv}`, `infra/ci/validate-controls.sh` | §6.2. |
| CI (API supplement) | `.gitlab-ci.yml`, `infra/ci/{validate,select-profile,validate-controls}.sh`, the `run.sh` B-CI-CONFIG lines | API branch pipelines run fast and the full gate (§9.4). |
| Documentation | `docs/foundation/WEB-SHELL.md`, `docs/foundation/OPERATIONS.md` | Plan §11.10; the API pipeline source (§9.4). |
| Management | `management/v0/results/G048-scope-allowlist.txt`, this Result | — |

## 4. Migrations and schema

None. No migration, table, tenant data or Organization branding storage.

## 5. Contracts, API and MCP

**No public HTTP, MCP or BFF change.** Routes stay exactly `/`, `/api/health` and `/api/ready` (B-WEB-BOUNDARY). The test-only `/token-fixtures` route and its `?open=` parameter exist only in the browser image (S-NOFIXTURE). The token override input contract (16 names) is unchanged.

**Internal interfaces changed** (Plan §10):
- the Button and Dialog part props;
- `PortalContainerValue`;
- `INTERNAL_ROLE_TOKENS`;
- the `tokenOverrideStyle()` output (derived role sources);
- the static assets `/fonts/**` and `/blueprint/overview.svg`.

## 6. Tests added and modified

### 6.1 Browser (Playwright, Chromium only, no retries)

Scenario matrix as Plan §11.2. `BROWSER_MATRIX_FULL` adds:
- S-FIDELITY, S-FONTS, S-BLUEPRINT and S-LONG-IDENTITY to the shell projects;
- S-PORTAL-INIT, S-FONTS and S-BLUEPRINT to the token projects.

There are 133 tests per full run, up from 103.

**New oracles:**
- **S-FIDELITY:** the §4.7 table, typed independently with citations. Lengths ±1px, letter-spacing ±0.05px; weights, stacks, borders, radius, shadows and case exact. It also checks the 320 and 640×400 expectations and the module `data-slot` presence. Evidence: `source-fidelity.json`.
- **S-FONTS:**
  - phase B behind the bounded readiness gate: bytes and pins; `name`/`fvar`/OS/2; CDP `getPlatformFontsForNode` custom verdicts on 15 production nodes; weights; synthesis twins;
  - control 6: held fonts, phase A only, detected as fallback while the page stays readable, operable and reflowing;
  - control 7: synthesis detected;
  - two swap-stability cases;
  - the fixture probes in every scope.
- **S-BLUEPRINT:**
  - semantics and layout neutrality;
  - the pinned same-origin response;
  - on/off paint against the structure-stroke composite;
  - per-group motif captures;
  - a held asset;
  - forced colours;
  - per-scope stroke colour by line fit.
- **S-PORTAL-INIT:** the insertion history recorded from MutationRecords before any page script; first-frame and live-mode values; controls PI-C1–C4.
- **S-LONG-IDENTITY:** a 64-character name on a fourth fixture server.
- **Local text contrast over the imagery,** with an injected-bar negative control.

**Replacements** (Plan §11.3), each guarantee kept:
- S-OVERLAY allows the Radix focus guard, then checks focus is back inside.
- S-MOTION expects no dialog motion; the button-transition control is kept.
- S-PALETTE covers 16 plus 3 sources and H1/title in `ink`.
- S-STATES covers the source variants, the outline's angled frame, sizes, `block` and `asChild`.
- S-ANGLE checks every angled variant and size, with the ring at 3:1 on all four sides.
- S-TEXT-SPACING runs with the bundled fonts loaded and a 45-item per-scope fixture inventory.
- S-NAV adds a heading-not-covered check.
- S-RESPONSIVE uses ruled-grid columns.
- First paint expects H1 in the independent `ink`, with the distinctness precondition (§11.3a).

**Test-support discipline added during execution** (§10; test code only, no assertion weakened):
- `resize()`: wait for the new viewport and two frames;
- `parkPointer()`: the pointer away from the skip link's position;
- `resetFocus()`: let a pending Radix focus restore run first;
- `waitForHydration()`: wait for focus to return to the trigger;
- `setSource()` and `setScheme()`: two frames after a direct source edit or appearance switch;
- `surfaceOf()`: scroll a scope's top edge into view before sampling;
- `expectAngleIntact()`: centre the focused button in the viewport before sampling its ring (cycle 6). All four sides are still required.

### 6.2 Harness and CI controls

- **B-DEPS (`lib.sh`):**
  - 8 reviewed additions: the Radix pair replaces Base UI;
  - `DEPS_REMOVED_PREFIXES=("@base-ui/")` emits `DEPS_REMOVED_PRESENT` for any package.json or lockfile key at any depth;
  - `DEPS_FORBIDDEN=(shadcn radix-ui)`.
- **B-DEPS-CONTROLS:** C5 is retargeted to the dialog. New controls: C13 (leftover Base UI), C14 (lock-only `@base-ui/utils`), C15 (`radix-ui`), C16 (slot 1.4.0), C17 (dialog as dev), C18 (lock 1.1.22) and C19 (`tailwind-merge` 3.7.0).
- **B-WEB-TOKENS:** the 16 names plus `INTERNAL_ROLE_TOKENS`, giving 19 equal across sources, resolution and `@theme inline`; `TokenOverrides` stays keyed by `SemanticColorToken`; `--spacing-cell`; untranslated kit colour names are rejected.
- **B-WEB-COMPONENTS:** refactored into `web_components_check ROOT` with reason codes. It checks:
  - `radix-nova`;
  - the seven `components/ui` files;
  - the Radix imports;
  - no `@base-ui` import anywhere, tests included;
  - no other `@radix-ui` specifier;
  - the reviewed `data-slot` set;
  - the X9 focus outline;
  - X7, meaning `--angle` only on `::before`/`::after`.
- **B-WEB-DESIGN:** refactored into `web_design_check ROOT`. It checks:
  - remote URLs outside `web/public`;
  - `@font-face` only in `globals.css`, exactly 4;
  - the web CSS `url()` set;
  - kit copies and literals;
  - the exact `web/public` inventory with pins;
  - the SVG rules: parse/root, raw-byte denial, element and attribute allow-lists, size and motifs.
- **B-WEB-ASSET-CONTROLS (new, lint):** A1–A13, each an exact code set.
- **B-NAMING:** `web/public` is allowed.
- **Register:** 20 rows (`web.Dockerfile` added, with G048 reason suffixes); B-RENDER-PLAN expects 20.
- **APPLICABLE:** 38 entries (`web/public`, the notices, `web.Dockerfile`).
- **`required-checks.tsv`:** 71 rows; fast 32, integration 64, full 70.
- **`infra/ci/validate-controls.sh`:**
  - VP1 expects 32 (§9.1);
  - the API supplement (§9.4) adds P9, P10, N19–N23 and F-api-intent, and N17 now uses a trigger pipeline. That gives 203 controls, up from 195.
- **B-CI-CONFIG (`run.sh`):** the seven workflow rules and five gate rules, with the API rule in each; the message is updated (§9.4).

## 7. Validation commands and outcomes

Every run used `g048-vw.sh <step> <suites>` → `infra/validation/run-approved.sh` against the bound inputs. `static` = `lint controls`; `variants` = `ops rename off`.

| Step | Run | Suites | wrapper / tee / postcheck | Overall | Notes |
|---|---|---|---|---|---|
| s1 | `e8b82a` | all | 0 / 0 / passed | 0 | Baseline: 69 PASS, 3 INFO |
| s3 | `0f93b9` | lint controls provenance unit | 1 / 0 / passed | 1 | First validation. FAIL: B-SHELL, B-WEB-BOUNDARY, B-WEB-TOKENS, B-UNIT-CHECKS (tsc ×2) and B-CI-CONTROLS (VP1, scope). B-WEB-ASSET-CONTROLS and B-DEPS-CONTROLS passed on the first run. `unit` was added to S3 for early typecheck feedback (§9.3). |
| s4a | `0a6e55` | lint provenance root | 1 / 0 / passed | 1 | Cycle-1 fixes confirmed; browser 114/133 |
| s4b | `9cfc7e` | static provenance root | 1 / 0 / passed | 1 | B-CI-CONTROLS 195/195 (VP1 at 32); browser 129/133 |
| s4c | `d5fbac` | lint root variants | 1 / 0 / passed | 1 | Browser root 131/133, off 132/133 (S-RUNTIME only), rename 2/2 |
| s4d | `41a37c` | root variants | 0 / 0 / passed | **0** | Root 133/133, off 133/133, rename 2/2; ops, outage and persist PASS |
| s8 | `75c635` | all | 1 / 0 / passed | 1 | FAIL: B-BROWSER-ROOT and B-BROWSER-OFF (S-ANGLE, tokens-light and tokens-dark; 131/133 each), B-BUILD-CONTEXT (Docker Hub 502), and therefore B-REQUIRED-CHECKS; §7.1 |
| s8b | `3d9d8a` | all | 0 / 0 / passed | **0** | Final tree after cycle 6: 70 PASS, 3 INFO; root 133/133, off 133/133, rename 2/2; B-CI-CONTROLS 203/203; §7.1 |

The per-cycle diagnosis is in §10.

### 7.1 Final run (S8, S8b) and §11.8 checks

**S8 (`75c635`, after cycle 5): overall 1.** B-REQUIRED-CHECKS lists exactly the three FAILs below.

- **B-BROWSER-ROOT and B-BROWSER-OFF: one test failed, in both token projects.**
  - The failing test was [S-ANGLE] with "default-adjacent size sm: outline sampled on all sides". It sampled left, right and top only.
  - `ringSamples()` skips a side whose ring or adjacent point (2.5–8px outside the edge) lies outside the viewport.
  - X23 made the `sm` button shorter (line-height 1). That moved the default-adjacent `sm` button so that, after the focus scroll, its bottom sample points fell below the viewport.
  - This is a sample-position defect in the test helper. The outline was not lost, and s4d had passed this test.
  - Cycle 6 centres the button before sampling. The four-side, colour and 3:1 assertions are unchanged.
- **B-BUILD-CONTEXT: an external registry error.** The error was `failed to resolve source metadata for docker.io/docker/dockerfile:1: … 502 Bad Gateway` (the syntax-frontend lookup for `migrate.Dockerfile`).
  - There is no repository cause: it passed in s1 and s4d, and nothing changed for it.
  - s8b rechecks it.
- Every other check passed: lint, controls (B-CI-CONTROLS 195/195), provenance, contracts, unit, integration, ops, rename (2/2) and the root and off variants apart from S-ANGLE.
- **§11.8 checks:**
  - pre (`$EV/s8-checks-pre/`) and post (`$EV/s8-checks-post/`) both reported `fails=0`;
  - the post tree digest was `896d3935…4667d` (625 files);
  - 52 paths; 80 Docker objects recorded.

**S8b (`3d9d8a`, final tree after cycle 6): overall 0.**
- **Run:** wrapper 0, tee 0, postcheck passed, `RESULT: PASSED`. Suites: lint, controls, provenance, contracts, root, migrations, ops, rename and off. It ran 06:12:55–07:16:07 (3792 s).
- **Summary table:** 70 PASS and 3 INFO (PREFLIGHT, PROBE, IMAGES), with no FAIL or UNMET. B-REQUIRED-CHECKS: the 69 required IDs of the executed suites each recorded exactly once as PASS.
- **Browser:** B-BROWSER-ROOT 133/133, B-BROWSER-OFF 133/133 and B-BROWSER-RENAME 2/2, all `expected`. S-ANGLE passed in both token projects on both stacks.
- **B-BUILD-CONTEXT:** PASS (the S8 registry 502 did not recur).
- **CI checks:** B-CI-CONFIG passed with the seven workflow rules and the API gate rule. B-CI-CONTROLS: 203 controls pass, including P9, P10, N17 (trigger), N19–N23 and F-api-intent.
- **Cleanup:** `inventory-leftover.txt` is empty.
- **§11.8 checks:**
  - pre (`$EV/s8-checks-s8b-pre/`): `fails=0`, tree digest `8fd23253…50ce` (625 files), 56 paths;
  - post (`$EV/s8-checks-s8b-post/`): `fails=0`, the same tree and status digests, so nothing was written during the run;
  - 56 paths and 0 ignored;
  - the Docker inventory before and after is identical (80 objects; unrelated resources untouched).
- **The delivered tree** differs from the s8b tree only in this Result file (§7.1–§11 completed after the run). The hosted pipeline (§7.3) runs on the delivered tree, Result included.

### 7.2 Screenshots

- **Before:** s1, embedded in `$EV/s1/report/browser-{root,off}-results.json`.
- **After:** s8b. Every attachment is extracted to `$EV/s8b/attachments-{root,off}/`: 120 files each (64 PNG, 52 JSON, plus text and hash files). The manifest is `$EV/s8b/attachments.sha256`: 240 entries, SHA-256 `72d650e5018920f738cb2ae134e31a60ac3b0494c68b7d13e4afcd987cbcb40a`.
  - The shell set includes `shell-1280.png`, `shell-320.png`, `shell-1280-dialog.png`, `spacing-shell-320*.png` and `blueprint-{overview-320,motifs,without-dimensions}.png`.
  - Also included: `source-fidelity.json`, the font, swap and blueprint JSON, and the fixture and portal records.
- **Kit reference:** `skills/vestive-design-skill/examples/Screenshot 2026-10-05 10{3447,3502,3548}.png`.

### 7.3 Hosted CI (S8h)

**Named actions** (reconciliation P9; the API supplement):
1. the local delivery commit on `worktree-g048-exec` (parent BASE, exactly the 56 paths);
2. a push of that exact commit to the dedicated evidence branch `g048-vestive-adoption`, with `-o ci.skip` so the push creates no redundant feature-push pipeline;
3. one API branch pipeline on that branch (`POST /projects/:id/pipeline?ref=g048-vestive-adoption`, the same API `glab ci run` uses).

The pipeline runs `fast`, then the `gate` with the full profile (`branch-api`, `api-explicit`).

**Not done:** no merge request, staging or main push, settings change, schedule, runner change, purchase or deployment.

**Identifiers and outcome:** the pipeline cannot be recorded inside the tree it tests. They are therefore recorded below in a Result-only local follow-up commit, which is not pushed, and in the Codex handoff.

**Outcome (2026-10-07, recorded after the delivery commit). Evidence is in `$EV/hosted/`.**

| Item | Value |
|---|---|
| Delivery commit | `40bfebf9399aef686d50ea3162e765e220b6f8d0`; tree `2c21faabf06725e65e5e5a2bf53ace647b3c38c0`; parent BASE; exactly the 56 expected paths (`delivery.tsv`, `staged-paths.txt`) |
| Push | New branch `g048-vestive-adoption` = `40bfebf9…` (verified with `ls-remote`). The push pipeline 2922043924 was `skipped` (`ci.skip`), as intended. |
| API pipeline | **2922044054**: source `api`, ref `g048-vestive-adoption`, sha `40bfebf9…`. `yaml_errors` was null. The workflow admitted the `api` source, and both `fast` (job 16998085056) and `gate` (job 16998085057) were created. This confirms the corrected workflow and gate rules on GitLab itself. |
| Result | **failed: both jobs `failure_reason=ci_quota_exceeded`**, `started_at` null, no runner, no trace and no artifacts. The pipeline finished at 2026-10-07T11:18:52Z. |
| Quota (read-only) | Namespace `<namespace>` (group <gitlab-group-id>) is on plan `free`. October 2026 usage: 1827 compute minutes (54788 s shared-runner duration), mostly the G003 hosted campaign (`ci-minutes-usage.json`). The monthly quota resets on 2026-11-01. |

**Hosted gate: BLOCKED (resource).** This is attempt 1, kept as history; correction loop 1 below supersedes it.
- The full gate has not run on the delivered tree, so AC9's hosted part is unmet.
- This is not a code or CI failure, and it did not consume a cycle.
- Following the batch's stop rule for resource conditions, it was not retried and no minutes were purchased. The evidence branch and the failed pipeline are left in place for review.

**To complete:** rerun one API pipeline on `g048-vestive-adoption` at `40bfebf9`, using the existing named authorization (`g048-hosted.sh run`, then `g048-pipeline-status.sh <id> final`). It needs compute minutes, from one of:
- (a) the 2026-11-01 reset;
- (b) owner-purchased additional minutes, which needs a purchase authorization;
- (c) a different runner, which is a runner and architecture change under ADR-002 and is not authorized.

A full API pipeline uses roughly 100 compute minutes on the medium runner. G003's H1″ measured about 5 runner minutes for fast and 41 for the full gate. The medium runner's cost factor is 2, as October's 54788 s → 1827 minutes shows.

#### Correction loop 1 (2026-10-07): the hosted rerun — PASSED

**Trigger.** Codex review `management/v0/results/G048-implementation-codex-review.md` (SHA-256 `4c86246d7adc5cbc0af62cd9f6aa381c9e9a3e558836854afebb528648331bec`) returned CORRECTION LOOP for AC9 hosted evidence only, with no code correction. The owner reported that compute minutes had been purchased and directed the resume.

**Actions.** One API pipeline, under the existing named authorization: `g048-hosted.sh run`, which is `POST /projects/<gitlab-project-id>/pipeline?ref=g048-vestive-adoption`.

**Not done:**
- no code, test, config or asset change, and no local full run, as Codex's review directs;
- no new implementation cycle (still 6 of 6);
- no push, merge request, staging or main change;
- no settings or runner change;
- no purchase by Claude;
- no G004 work.

**Before the run:**
- the review hash was verified;
- `ls-remote` showed `g048-vestive-adoption` = `40bfebf9399aef686d50ea3162e765e220b6f8d0`;
- `rev-parse 40bfebf9^{tree}` = `2c21faabf06725e65e5e5a2bf53ace647b3c38c0`;
- attempt 1's records were preserved (`pipeline-2922044054*.json`, `pipeline-2922044054-create.json`, `delivery-attempt1.tsv`).

**After the run:** the branch had not moved, and `main` and `staging` were unchanged (`99e2c794`, `b5e0154d`). The branch's pipelines are exactly 2922479202 (api, success), 2922044054 (api, failed: quota) and 2922043924 (push, skipped: `ci.skip`).

| Item | Fast | Gate |
|---|---|---|
| Pipeline | **2922479202**, source `api`, ref `g048-vestive-adoption`, sha `40bfebf9…`, **success**, 3913 s | (same) |
| Job | 17001434264, **success**, 336 s | 17001434266, **success**, 3577 s |
| Runner | `saas-linux-medium-amd64` (GitLab hosted; ADR-002 contract) | `saas-linux-medium-amd64`; RP1–RP12 PASS (cpus 4, 15 GiB, 37 GiB free; budget 8659 s) |
| Attribution | `result: accepted`, `CI_PIPELINE_SOURCE: api`, tested content `branch-api`, head `40bfebf9…`, tree `2c21faab…` | The same, for job 17001434266 |
| Profile | pipeline **full**, executed **fast**, reason `api-explicit`, base none, no selection table | pipeline **full**, executed **full**, reason `api-explicit`, base none, no selection table |
| Required set | 32 (`0acf2beb…cd34`); `results.tsv` 32 PASS, 3 INFO, nothing else | **70** (`57e9bcac…bb82`); `results.tsv` 70 PASS, 3 INFO, nothing else; B-REQUIRED-CHECKS PASS; completeness phase 0 |
| Key checks | lint, provenance, contracts and unit set | B-CI-CONFIG (seven workflow rules incl. api); B-CI-CONTROLS 203; B-DEPS-CONTROLS; B-WEB-ASSET-CONTROLS; B-BUILD-CONTEXT; B-SECRET-CANARY (73 secrets); B-SCOPE |
| Browser | — | root 133/133, off 133/133, rename 2/2, all `expected` |
| Status and phases | `validate: final status 0; published=full; registry=removed; harness_stop=none`; every phase 0 (location … scrub) | The same |
| Credentials and privacy | `remotes=1 sanitized=1 … password_urls=0 token_files=0 scan=clean`; after_script `scan=clean`; `final: context=0 guard=ok` | The same |
| Cleanup | `cleanup-inventory.tsv` empty; `inventory-leftover.txt` only the permitted `project-archetype-render-cache` volume (as in G003's accepted hosted runs) | 2 captured images removed by ID; leftover only the permitted render-cache volume |
| Capacity | — | baseline 40022769664 bytes free, minimum 27778105344, incremental peak 12244664320 bytes; durable floor 35 GiB met |
| Artifacts (SHA-256 of archive) | `0822422192e849d2df9c2316637c02a92d7a6677d162cbf7599f3ee4380b06db` | `f697e522aa780f48d083cdac26bc496da8be80bc1da43a82bf13e6bdc63840c4` |

- **Gate suite durations (s):** lint 132, controls 897, provenance 16, contracts 57, root 1360, migrations 43, ops 239, rename 65, off 675.
- **Sanitized evidence:** `$EV/hosted/` holds `pipeline-2922479202{,-jobs}.json`, `job-1700143426{4,6}-trace.log` (URL userinfo redacted; zero credential-pattern matches in traces and artifacts), `artifacts-*/artifacts.zip` and `extracted-*/ci-evidence/`. Gate `results.tsv` `d525caf3…3cd2`, `summary.json` `05a25784…aa7c`, `attribution.json` `54b674a4…2b63`.

**Required-count note.** The full profile's required set is 70: the 69 executed-suite IDs plus B-REQUIRED-CHECKS itself (`required-checks.tsv`: fast 32, integration 64, full 70). G003's accepted full count was 69. G048 adds B-WEB-ASSET-CONTROLS.

**Outcome.** The fast and full gates succeeded on the exact delivered tree. AC9's hosted part is met. G048 returns to Codex for the acceptance decision.

### 7.4 Intake records

- **Fonts:** see the WEB-SHELL.md "Fonts" table and `$EV/intake/fonts.tsv`, which hold the URL, size, blob SHA-1 and SHA-256 at `google/fonts@7085eb89`. All seven equal Plan §7.1.
- **Lockgen:**
  - command: `node:24-slim` (`sha256:0e0ff40c…f9b6`), npm 11.19.0, `npm install --package-lock-only --save-exact --ignore-scripts` (Plan §5.2);
  - delta (`$EV/lockgen/delta.diff`): exactly the 24 additions and 9 removals of Plan §5.1, all MIT, with the direct integrities equal to §5.1;
  - one flag change, as planned: `@types/react-dom` from dev to devOptional;
  - `package.json` differs only in the three keys.
- **SVG:** `overview.svg`, 2474 bytes, SHA-256 `942a2a242594773f6be7e638d6e07f9f723a0ff6b6c6e8334427702e52f98c65`; groups construction 11, dimensions 15, structure 18. Authored during implementation. One structure path (`M800 104V56M880 104V24`) was added before the first application, so the 320 band carries enough structure strokes. Pinned at authoring.

## 8. Acceptance criteria

| AC | Status | Evidence |
|---|---|---|
| AC1 Six modules adopted with Radix, rights and traceable migration | Met (s8b `3d9d8a`) | Lineage headers; module diffs and the hunk map (§12); P3 rights record; THIRD_PARTY_NOTICES; B-WEB-COMPONENTS; B-DEPS (Radix pins, `DEPS_REMOVED_PRESENT` absent); S-FIDELITY slot presence |
| AC2 Source geometry, type and states within reviewed exceptions | Met (s8b `3d9d8a`) | S-FIDELITY (`source-fidelity.json`, §4.7 values; X23 restores the line-height); S-STATES; S-ANGLE; screenshots against the kit |
| AC3 Bundled fonts render | Met (s8b `3d9d8a`) | S-FONTS phase B plus controls 6/7 and both swaps (root and off); fixture probes; intake pins |
| AC4 Blueprint present, safe, decorative, both appearances | Met (s8b `3d9d8a`) | B-WEB-DESIGN SVG rules; S-BLUEPRINT (shell and scopes); screenshots; local contrast over the imagery |
| AC5 Appearance, first paint and hydration intact | Met (s8b `3d9d8a`) | The appearance project (11 IDs, H1 in `ink`); dev hydration ×6 |
| AC6 Scoped overrides compose | Met (s8b `3d9d8a`) | S-TOKENS (16 plus 3 roles); S-PORTAL; S-PORTAL-INIT (PI-C1–C4); S-RUNTIME; S-BLUEPRINT scopes |
| AC7 Accessibility with real fonts, imagery and geometry | Met (s8b `3d9d8a`) | S-A11Y, S-CONTRAST (local minima, negative control), S-ANGLE, S-OVERLAY, S-TAB, S-SKIP, S-NAV, S-RESPONSIVE, S-TEXT-SPACING, S-MOTION, S-LONG-IDENTITY, forced colours, control 6 |
| AC8 Guards reconciled without broad paths | Met | B-DEPS-CONTROLS C1–C19 and B-WEB-ASSET-CONTROLS A1–A13 exact (§6.2) |
| AC9 Full inherited validation passes | **Met.** Local: s8b `all`, overall 0. Hosted: pipeline 2922479202, fast and full gate success on `40bfebf9`/`2c21faab` (§7.3, correction loop 1). Attempt 1 (2922044054, `ci_quota_exceeded`) is kept as history. | s8b: wrapper, `tee` and overall 0; postcheck passed; root, disabled and rename browser runs; README and docs blocks; migrations, isolation, ports, outage and persist; B-DEPS on the real pair and B-DEPS-CONTROLS complete; B-CI-CONTROLS 203. Hosted full gate on the delivery commit: §7.3 |
| AC10 Provenance, scope, cleanup and record | Met | B-SCOPE (56 paths within the allow-list); §11.8 pre/post `fails=0`; no leftover resources; Docker inventory unchanged; this Result |

## 9. Deviations, reviewed exceptions and architecture

**Architecture deviations: none.** `web` stays a thin UI/BFF. There is no schema, API, MCP, tenant, authority, provider or product change.

### 9.1 Authorized scope addition: VP1 count (owner and supervisor)

`infra/ci/validate-controls.sh` (G003's accepted CI controls) hardcoded the fast-profile required count as 31. Reconciliation PG-7 places B-WEB-ASSET-CONTROLS in `lint`, which gives fast 32, so B-CI-CONTROLS failed in s3. `infra/ci/` was outside the allow-list, and I stopped that work and escalated.

- **Owner:** answered "Edit VP1 to 32".
- **Codex:** authorized exactly this change (`$EV/codex-vp1-count-authorization.txt`): add only that path to the allow-list, and change only `ctl_VP1`'s expected count and detail text.
- **Diff** (`$EV/vp1.diff`):

```diff
-  sum_has profile '"fast"'; sum_has executed_profile '"fast"'; sum_has required_count 31
+  sum_has profile '"fast"'; sum_has executed_profile '"fast"'; sum_has required_count 32
-  detail "exit $RC; fast: suites lint provenance contracts unit, 31 required IDs, executed_profile fast; published the fast set ($(published_list | wc -w) files)"
+  detail "exit $RC; fast: suites lint provenance contracts unit, 32 required IDs, executed_profile fast; published the fast set ($(published_list | wc -w) files)"
```

- **Allow-list:** one added line, `^infra/ci/validate-controls\.sh$`.
- **Unchanged:** no other CI behaviour, no negative control, and nothing made dynamic.
- **Result:** B-CI-CONTROLS passes 195/195 (s4b onward). This addition did not consume a cycle.

### 9.2 Reviewed exceptions added or widened during execution (for Codex review)

- **X19, widened.** S-LONG-IDENTITY at 320 found the lead and the header identity overflowing (scrollWidth 480). `overflow-wrap:anywhere` is now at all three configured-identity call sites: H1, the lead (`page.tsx`) and the header brand (`app-shell.tsx`, which also gets `min-w-0 max-w-full`). These are call sites in project files; no kit module is changed.
- **X23, new.** The `sm`, default and `lg` sizes repeat `leading-none` after their text size.
  - The class merge the source relies on (`cn`, tailwind-merge semantics) drops the base's `leading-none` when a later `text-[…]` follows, so labels inherited the document's 1.5. The help trigger measured 49.1px tall instead of 42.08px.
  - The kit's own `.vx-btn` font shorthand (line-height 1), its reference CTA (`reference.html`: 14px/30px padding, .88rem) and the kit screenshots (header CTA 42px, hero CTA about 51px) confirm Plan §4.7.
  - Plan §4.4's "merge reliance … verified, not assumed" check found the gap; X23 restores the source value.
  - `block` keeps `leading-[1.2]`; `icon` has no text size.
- **X22, clarified.** Neutral comments use the translated token names.

### 9.3 Procedural deviations

1. **S3 also ran `unit`** (`static provenance unit`) for early lint/typecheck feedback. Nothing was weakened.
2. **S-FIDELITY expectation, corrected and then restored.**
   - Cycle 2 changed the help-trigger and size line-height expectations to the merged value (21.12/49.12).
   - The S6 reviewer flagged that this reads the expected value from the implementation (Plan §11.1).
   - Cycle 5 restored the Plan's §4.7 values and fixed the implementation instead (X23).
3. **B-WEB-BOUNDARY input.** `web_sources` excludes `web/public` (static font binaries, licences and the drawing). Concatenating binaries switched grep to binary mode and broke the single-`suppressHydrationWarning` count. The boundary assertions are unchanged; `web/public` is bounded by B-WEB-DESIGN's pinned inventory.
4. **B-WEB-DESIGN remote-URL and `@font-face` rules** apply to code and stylesheets, not to `web/public`. The pinned inventory and the SVG `href`/`url(` denial cover `web/public`. This matches the A8 expected code set.
5. **`asChild` in the server-rendered fixture.** Plan §11.2 places an `asChild` anchor in the shared fixture `ScopeContent`; §5.5 says `asChild` is used only in client components. The fixture builds and renders: Slot 1.3.3 under RSC passes no function ref when none is forwarded. The pre-declared fallback (`"use client"` on `button.tsx`) was **not** needed. `button.tsx` stays directive-free.
6. **S6 timing.** The read-only review started on s4c evidence: every shell scenario was green there, and only S-RUNTIME failed, with a test-only fix. It started before s4d, the first fully green root, rather than strictly after it.
7. **D-ANGLE:** not used.

### 9.4 Authorized scope addition: API branch pipelines (owner and supervisor)

**Authorization.** The owner corrected G003's CI so that it does not reject API pipelines. Codex authorized a bounded correction in `management/v0/results/G048-api-pipeline-authorization-2026-10-07.md`:
- original binding SHA-256 `b857b22ab9cf212dc620604b655cab00cc943b6b0482cf7a4f5369796495ceab`;
- the dated follow-up approving `infra/validation/run.sh` (B-CI-CONFIG workflow and gate expectations, count and message only) is appended to the same file. Its current SHA-256 is `dda5c8898726c4a23ae205d5f57fd1f259f86be63750b06f0aaf9f4d8627d9a9`, verified before the edits.

The correction is part of cycle 6. It was drafted outside the repository while S8 ran and applied only after S8 finished. The pre-change tree inventory is `$EV/s8-checks-post/` (tree digest `896d3935…4667d`, 625 files, 52 paths).

**Behaviour.**
- **`.gitlab-ci.yml`:**
  - the workflow runs `$CI_PIPELINE_SOURCE == "api" && $CI_COMMIT_BRANCH`, before the final `when: never`;
  - `gate` runs for `api`;
  - `fast` runs as in every supported pipeline.
- **`infra/ci/validate.sh`:**
  - an `api)` attribution case mirrors `web`. It rejects merge-request metadata ("api pipeline carries …"), a missing branch, and a ref that differs from the branch;
  - tested content is `branch-api`;
  - both tested-content patterns accept `branch-api`;
  - the source is recorded as `api` in `attribution.json`;
  - every other source rejection and the failure precedence are unchanged.
- **`infra/ci/select-profile.sh`:** `branch-api` selects pipeline profile full, with reason `api-explicit`, without a diff.

**Controls (`infra/ci/validate-controls.sh`, 203 in total, 195 before).**

| Control | Case | Expected |
|---|---|---|
| P9 | API gate job | full, `api-explicit`, `api` recorded |
| P10 | API fast job | pipeline full, executed fast |
| N17 | Trigger pipeline (the genuinely unsupported source) | "unsupported pipeline source"; trigger recorded. This replaces the obsolete api negative. |
| N19 | Wrong commit | Rejected |
| N20 | No branch | Rejected |
| N21 | Ref mismatch | Rejected |
| N22 | Tag | Rejected |
| N23 | Merge-request IID | Rejected |
| F-api-intent | Docs-only change | full (`api-explicit`); no diff table |

**`infra/validation/run.sh` (B-CI-CONFIG only).**
- Both exact rule expectations include the api rule.
- The message reads "the seven workflow rules (…; web; api branch pipeline; never)" and "source-only rules (MR, staging/default push, schedule, web, api)".
- Nothing else in `run.sh` changed for this supplement.

**`docs/foundation/OPERATIONS.md`.** It describes:
- api pipelines in the workflow, the gate and profile selection;
- the `branch-api` tested content;
- "other sources (for example trigger pipelines) fail with 16";
- the cost line.

**Allow-list.** One comment and three lines: `^\.gitlab-ci\.yml$`, `^infra/ci/(validate|select-profile)\.sh$` and `^docs/foundation/OPERATIONS\.md$`.

**Not done (not authorized):** a staging or main push, any settings change, a merge, a deployment, a purchase, or a runner change.

## 10. Execution budget, cycles, checkpoint

- **Budget:** B5 (Plan §14). Owner and Codex extension: maximum **six** total shared cycles, recorded in `management/v0/results/G048-budget-extension-2026-10-07.md` (SHA-256 `b272afe5ae5fcc55be40d4cc5816f8a4a2383014a97b3a861876af8132a9143d`) when four were consumed. No reset.
- **Consumed: 6 of 6.** No budget remains. The full log is in `$EV/cycle-log.md`.

| Cycle | Trigger | Fixes | Retest |
|---|---|---|---|
| 1 | s3 | Shellcheck SC2178/2128 (array name collision); `web_sources` binary scan; role regex `[a-z0-9-]`; two TypeScript casts | s4a: lint, provenance and root green except the browser (114/133) |
| 2 | s4a | S-FIDELITY timing (active filter, stale layout after resize); help-trigger expectation (later reverted, §9.3); Radix focus-restore race; X19 widened; fixture top-edge sampling; spacing count 45; S-RUNTIME diagnostics | s4b: 129/133 |
| 3 | s4b | Pointer parked away from the skip link (S-FONTS twin, S-FIDELITY, S-CONTRAST); settle after direct source edits | s4c: 131/133 and 132/133 |
| 4 | s4c | Settle after appearance switches (S-RUNTIME, S-PORTAL/-INIT); forced-colours settle | s4d: overall 0 |
| 5 | S6 review | X23; restored §4.7 values; S-FIDELITY tightening (exact borders and radius, letter-spacing ±0.05, tone, border, Close-bar and `ch` max-width rows); evidence fields; strict `verdictOf` and font content type; docs (X19, X22, X23) | S8 `all` (`75c635`): overall 1. S-ANGLE failed in tokens-light/-dark (root and off), and B-BUILD-CONTEXT failed on an external Docker Hub 502 (§7.1) |
| 6 | S8 | `expectAngleIntact()` centres the button before ring sampling; the authorized API-pipeline CI correction (§9.4) | s8b `all` (`3d9d8a`): overall 0 |

**Checkpoint after cycle 3** (BUDGET_POLICY §7, recorded before cycle 4):
- **ACs:** AC1, AC3–AC8 had passing evidence; AC2 was green; AC9 and AC10 remained.
- **Failures:** cycle 1 was harness defects; cycles 2–3 were measurement-timing artefacts in the new oracles; one product fix (X19).
- **Repeated pattern:** reading state before it settles. Each instance was distinct and resolved systematically.
- **Convergence:** strong (browser failures 19 → 4 → 3); every non-browser check green since s4b.
- **Boundedness:** the Goal remains bounded. Decision: continue.

## 11. Provenance, integrity and cleanup

`$EV/../g048-bin/g048-s8-checks.sh pre|post` runs the §11.8 checks around each final run (read-only). They cover:
1. the tree and NUL-status digests;
2. the exact path inventory against `$EV/paths-expected.txt` (56) and zero ignored paths;
3. the kit manifest (63 OK; `skills/` unchanged);
4. the package pair equal to the lockgen output;
5. the added import sources;
6. fonts, notices and the SVG equal to their pins;
7. provenance: no kit line outside the adopted modules and tests that is not already at BASE;
8. no `.env` in the evidence;
9. the Docker inventory.

**Results:**
- S8 pre/post, and s8b pre/post, all reported `fails=0`.
- s8b: the tree and status digests are identical before and after.
- Kit manifest: 63 OK.
- The package pair equals the lockgen output (`3e0c86502d40`, `8abae14eeb85`).
- The added import sources are only the adopted modules, Radix, `class-variance-authority` and React.
- Fonts, notices and the SVG equal their pins.
- No kit line outside the modules and tests.
- No `.env` in the evidence.
- The Docker inventory is unchanged (80 objects).
- The harness's `inventory-leftover.txt` is empty.

**S9 (evidence extraction):**
- the browser attachments (§7.2);
- the run summaries (`$EV/runs.tsv`);
- `$EV/cycle-log.md`;
- the §11.8 directories.

Evidence stays in the owned external `$EV`, as reconciliation PG-14 requires; none of it is committed.

## 12. S6 read-only review: findings and triage

The reviewer was a read-only subagent, with no edits and no Docker. It reviewed the module diffs, the screenshots, all JSON evidence and the oracles against Plan §4–§8 and §11.

| # | Severity | Finding | Disposition |
|---|---|---|---|
| 1 | Should-fix (Codex decision) | The help-trigger and size line-height oracle had been changed to the implementation (21.12/49.12); the kit reference shows 42px | Cycle 5: X23 fixes the implementation; §4.7 values restored (§9.2, §9.3) |
| 2 | Should-fix | The X19 extension was undocumented | WEB-SHELL.md X19 widened; §9.2 |
| 3 | Should-fix | S-FIDELITY looser and narrower than §4.7 (borders and radius at ±1, letter-spacing ±1; tone, max-width, Close bars, skip border colour missing) | Cycle 5: all added or tightened |
| 4 | Should-fix | Control-6 durations, swap steps and faces, SVG hash and portal live values missing from the evidence | Cycle 5: all recorded |
| 5 | Note | `verdictOf` empty list gives "fallback"; octet-stream accepted | Cycle 5: "none" verdict; font content types only |
| 6 | Note | Behaviour-neutral hunks with no T/X entry | Hunk map below; X22 clarified |
| 7 | Note | axe `aria-hidden-focus` is "incomplete", not "pass", with the dialog open | Recorded: zero violations; "incomplete" unchanged from G047. Plan §5.4's "passes" wording is corrected here. |
| 8 | Note | Blueprint outside the tested widths: a hard right edge above 1360px; lead contrast at 1024–1279 is estimated, not measured (about 5:1); cropping at 1280; the fixture panel is a sliver | Follow-up (§13); behaviour as Plan §8.4 |
| 9 | Note | The missing-asset case holds the request (a 404 is not exercised) | Meets §8.5 item 5 wording; follow-up |
| 10 | Note | S-RUNTIME evidence was pending at review time | s4d passed S-RUNTIME (root and off) |
| 11 | Note | The skip-link focus ring is partly off-screen at top-left (unchanged from G047) | Follow-up |
| 12 | Note | The header at 320 is tall (about 276px), faithful to the source's gaps | Recorded |

**Hunk map additions** (behaviour-neutral, reviewer finding 6):
- `site-header.tsx`: the nav item key is `href`, not `label`; both are unique.
- `layout.tsx`: the custom property `--vx-cell-bg` is renamed `--ruled-cell-bg` (T-family naming; same value source).
- `typography.tsx`, `bullet-list.tsx` and `dialog.tsx`: doc comments use the translated names ("gold" → "accent", "dark scrim" → "scrim") (X22).

Every other hunk maps to a T or X entry or the lineage header (reviewer verified).

## 13. Suggested follow-up (no authorization implied)

- Blueprint at wide viewports: a right-edge fade or extension above 1360px, and local-contrast sampling at 1024–1279 (reviewer 8).
- A failed-load (404) case for the drawing (reviewer 9).
- The skip-link ring offset at the viewport edge (reviewer 11; pre-existing).
- A follow-up that removes the now fixture-only `lucide-react` dependency (Plan Q10).
