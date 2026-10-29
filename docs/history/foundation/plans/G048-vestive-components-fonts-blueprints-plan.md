> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G048 — Vestive source components, bundled fonts and blueprint imagery: implementation plan

- **Plan revision:** r2, 2026-10-06. It corrects r1 under Codex's CORRECTION LOOP of 2026-10-06 (§0.1).
  - r1 is preserved byte-identical at `management/v0/plans/G048-vestive-components-fonts-blueprints-plan-r1.md` (SHA-256 `2f0203dc41b30b7feef14ff85b76fcac94b1d6a9b6e7285d15881cd211e19a53`).
- **Author:** Claude Code, as the planning-only agent. **Reviewer:** Codex. **Major decisions:** product owner.
- **Status:** resubmitted for Codex plan review.
  - This Plan is not approved.
  - Nothing has been implemented, installed, built or run in Docker.
  - Nothing has been committed or pushed, and no hosted action has been taken. None of these is authorized by this document.
  - **Zero implementation cycles are consumed** by drafting r1 or r2.
- **Governing inputs (r2):**
  - Goal `management/v0/goals/G048-vestive-components-fonts-blueprints.md` (`ac7e0a07…65f2`)
  - Spec `management/v0/specs/G048-vestive-components-fonts-blueprints-spec.md` (`481fb423…e63b`)
  - ADR-001 `docs/adr/ADR-001-radix-vestive-components.md` (`4fb60db7…0756`)
  - AGENTS.md, CLAUDE.md and BUDGET_POLICY.md
- **Preserved history:** the r1 Goal and Spec (`2520959…fb9d`, `ac30cf7a…4df5`) are read for history. Wherever r1 and r2 differ, the r2 documents govern.
- **Budget:** B5. At most five shared, meaningful implementation/debug cycles, with a checkpoint after cycle 3.
  - Codex kept B5 on 2026-10-06 and authorized no extension or split.
  - Codex called the four-cycle forecast plausible but tight. It is reassessed against the accepted post-G003 tree (§14).

> **PENDING banner: post-G003 facts.**
>
> This Plan is drafted against two sources:
> - the planning checkout `dc1ff7c315e3f2c753b5281a93d6d989f7d04e16` (governance, ADR, Goal and Spec);
> - the accepted G047 checkpoint `cbffd6da5d9936c90fdfd9e51d84d36c0f17ad81` (application, tests and harness). I read it from the clone's object store with `git show`. It is **not** an ancestor of `dc1ff7c`.
>
> G003 is **not accepted**, and its implementation is in validation elsewhere. Every statement below about the post-G003 baseline is marked **PENDING** and listed in §2.4. This covers the harness, CI, required checks, register and `APPLICABLE` counts, browser counts, package scripts and file contents.
>
> Before implementation, the actual accepted and integrated post-G003 tree must be re-inspected, and this Plan mechanically reconciled to it and re-reviewed. No G003 evidence is claimed or assumed accepted.

---

## 0. Revision record

| Rev | Date | Change |
|---|---|---|
| r1 | 2026-10-06 | Initial Plan against Goal and Spec r2 and ADR-001. Planning used one read-only web-research subagent for font provenance (zero cycles). |
| r2 | 2026-10-06 | Bounded correction of the four findings in Codex's r1 CORRECTION LOOP (§0.1):<br>• AC5 first-paint `ink` expectations (§11.3a; `appearance.spec.ts` added as §9.1 row 51);<br>• historical portal-insertion evidence (§5.4.1);<br>• exact C19 semantics (§11.5);<br>• a bounded blocked-font control (§7.4).<br>Q2 is recorded as answered and Q8 records Codex's budget direction. No other scope change. Planning used only read-only registry metadata lookups for `tailwind-merge` (no downloads, no subagent, zero cycles). |

### 0.1 Review history

**2026-10-06 — Codex review of r1: CORRECTION LOOP.**

- **Record.** "G048 Plan r1 review: CORRECTION LOOP" in the main checkout's `management/v0/PROGRESS.md`.
  - Main HEAD is `3d022ef5e5e1d4c901a80a84b78c85e6f8b0084b`.
  - The section was uncommitted in that working tree when read; the file hash was `61210afe…34d8`.
- **Reviewed artifact.** r1: SHA-256 `2f0203dc41b30b7feef14ff85b76fcac94b1d6a9b6e7285d15881cd211e19a53`, 1450 lines. It is archived byte-identical at `management/v0/plans/G048-vestive-components-fonts-blueprints-plan-r1.md`.
- **Confirmed by the review.**
  - Still in scope:
    - the selected source adoption;
    - the Radix Dialog/Slot direction;
    - bundled fonts;
    - the original decorative blueprint;
    - the separation of internal roles from the 16 override tokens.
  - Dialog 1.1.23 and Slot 1.3.3 are valid registry versions, and no upgrade is required (Q2 answered).
  - B5 stays, with a checkpoint after cycle 3. The four-cycle forecast is plausible but tight and is reassessed against the accepted post-G003 tree. No extension or split is authorized.
- **Still open after the review** (r2 does not resolve these):
  - G003 acceptance and integration, and the PG-1–PG-15 reconciliation;
  - the selected-code rights and upstream notices (Q1/Q1b);
  - final Plan approval;
  - the supervisor-created committed execution binding.

| # | Finding (Codex label) | r2 resolution | r2 sections |
|---|---|---|---|
| 1 | AC5 — first-paint expectations | H1 is now checked against independent `ink` values: `PALETTE[m].ink` from `ROLE_SOURCES`. The semantic `foreground` values stay distinct, with a precondition that the two differ by more than twice the painted tolerance. Only the two H1 expectations in `appearance.spec.ts` change. These are all kept: the sample/FCP correlation, the held-JS proof, both opposing-OS cases, hydration continuity and the tolerances. `appearance.spec.ts` joins the inventory and allow-list. The tests rerun on the root and disabled variants. | §1, §2.3, §2.4 (PG-13), §6.2, §9.1 (rows 34, 51), §9.2, §11.2, §11.3, §11.3a, §11.6, §11.7, §13, §15 AC5 |
| 2 | AC4 — insertion-history evidence | Evidence comes only from `MutationRecord.target` and the added portal subtrees, plus removal records for nodes taken out of a foreign subtree; it never uses delivery-time parents. Same-task body→scope relocation controls must be detected: PI-C2 is the required one, and PI-C1, PI-C3 and PI-C4 cover the variants. The initially open scoped cases add first-frame and live-mode checks. The adapter contract is unchanged. | §1, §5.4, §5.4.1, §9.1 (row 40), §11.4, §11.6, §11.7, §15 AC6, §16 (j) |
| 3 | AC9 — C19 exact dependency semantics | Option (a): `tailwind-merge` exactly **3.7.0**, consistent across package.json, the lock root and the lock entry. The expected sorted set is exactly `DEPS_EXTRA_DIRECT tailwind-merge`, and every rule is shown evaluated independently. No rule is relaxed. The complete controls and the real pair rerun. | §11.5, §11.6, §11.7, §15 AC8/AC9, §18 |
| 4 | AC2/AC7 — blocked-font control termination | The checks are split in two:<br>• phase A: bounded pre-load checks that never wait for fonts;<br>• phase B: the positive loaded-font oracle, behind a bounded readiness gate.<br>The held-font control uses `domcontentloaded` readiness, explicit timeouts and release in `finally`. It must return the `fallback` findings and complete its readability, keyboard and reflow steps. A timeout is always a failure, never a detection. | §1, §7.2, §7.4, §9.1 (row 35), §11.4, §11.6, §14, §15 AC2/AC3/AC7, §16 (i) |

**AC label note.**
- The table keeps Codex's finding labels.
- In the Spec r2 criteria table:
  - scoped-portal evidence sits under AC6 (R4);
  - dependency controls sit under AC8, and their final rerun under AC9;
  - blocked-font evidence sits under AC3 and AC7.
- §15 therefore records:
  - finding 2 under AC6;
  - finding 3 under AC8 and AC9;
  - finding 4 under AC2, AC3 and AC7.
- AC4 (blueprint) is unaffected.
- Q13 asks Codex to confirm or redirect this mapping.

---

## 1. Summary of the key decisions

1. **Adopt six source modules into `web/components/ui/`.** These are the actual `typography`, `layout`, `bullet-list`, `site-header`, `button` and `dialog` modules from `skills/vestive-design-skill/assets/shadcn/components/ui/` (hashes in §4.1).
   - Composition is kept: Radix Dialog, Slot/`asChild`, CVA, and `cn` through the existing `@/lib/utils`.
   - Every source-to-consumer alteration is listed (§4.4).
   - Each alteration is either:
     - a value-identical **token translation** (T, §4.2); or
     - an enumerated, measured **exception** (X, §4.3) within the Spec R2 categories.
   - Nothing is a silent substitution.
2. **Exact-pinned dependency delta** (§5.1):
   - add `@radix-ui/react-dialog` **1.1.23** and `@radix-ui/react-slot` **1.3.3**, both `dependencies`, both MIT;
   - remove `@base-ui/react` 1.8.0.

   Dialog 1.1.23 depends on exactly slot 1.3.3, so there is one copy of Slot. This pair has been the current 1.1 line since about 2026-07-25. The newer 1.2.0 and 1.4.0 were published on 2026-10-05. Codex confirmed the 1.1.23/1.3.3 pins on 2026-10-06 and requires no upgrade (Q2 answered). The lockfile is regenerated in Docker with `--package-lock-only --ignore-scripts` outside the repository, and the delta is verified against the planned transitive set (24 added, 9 removed).
3. **Complete Base UI migration** (§5.3). The following all move to Radix or native equivalents:
   - every production and test import, render prop, compound-part prop and selector;
   - the scoped portal adapter;
   - the `@base-ui` lockfile entries.

   The result is enforced by import and dependency guards with negative controls (§11.5).
4. **Scoped portals without a body fallback** (§5.4).
   - The portal context now carries the scope's container **element** (callback-ref state), not a `RefObject`.
   - Inside a scope, the adapted `DialogPortal` renders nothing until that element exists. This is because Radix `Portal` falls back to `document.body` whenever its `container` is null.
   - A new S-PORTAL-INIT scenario covers dialogs that are open at mount.
     - It proves their placement history from the mutation records' own insertion targets (`MutationRecord.target`) and the added subtrees, never from a node's parent at callback delivery.
     - Same-task body-to-scope relocation controls must be detected.
     - First-frame and live-mode values must be the scope's own (§5.4.1).
5. **Fonts** (§7):
   - **Source.** Unmodified upstream binaries from `google/fonts` at `7085eb89a950e85db5b166b7a58d414544b4140c`:
     - Oswald variable (wght 200–700);
     - Barlow Condensed static 400 and 600;
     - Roboto variable (wdth 75–100, wght 100–900).
   - **Notices.** Each family's OFL-1.1 text is shipped beside it. Files are renamed, never modified, so original SHA-256 equals delivered SHA-256.
   - **Delivery.** They are served same-origin from `web/public/fonts/` through `@font-face` with `url()` sources only and `font-display: swap`.
   - **Proof.** A new S-FONTS oracle checks:
     - response bytes;
     - the binary's own name, `fvar` and OS/2 tables;
     - CDP `CSS.getPlatformFontsForNode` (`isCustomFont`, family);
     - a synthesis twin-render check.

     It has two separated phases:
     - bounded pre-load checks that never wait for font loading;
     - the positive loaded-font oracle, behind an explicit, bounded readiness gate (§7.4).

     It has two negative controls:
     - **Held font requests.** Fallback must be detected, and the page must stay readable and operable. Every hold is released in `finally`, and a timeout is never a pass.
     - **A deliberately missing weight.** Synthesis must be detected.
6. **Blueprint imagery** (§8). One original, hand-authored, static SVG (`web/public/blueprint/overview.svg`, viewBox `0 0 960 640`, at most 40 KiB, no references at all).
   - **Motifs.** Three id'd motif groups: `construction`, `dimensions` and `structure`.
   - **Delivery.** It is used only as a CSS **mask** on an empty, `aria-hidden`, pointer-events-none layer, coloured by the scope-resolved `--muted-foreground` at 0.5 opacity. That makes it theme- and scope-aware, and it cannot execute anything.
   - **Placement.**
     - At 1024px and above: the right 46% of the overview container, with a 96px fade.
     - Below 1024px: the overview's 88px bottom padding band, so it is never behind text.
   - **Static guard.** The structure and safety are checked by the harness (yq XML parse, allow-lists).
   - **Browser checks.** S-BLUEPRINT proves painted strokes with imagery-on/off pixel comparison.
7. **Tokens** (§6). The 16-name colour contract and `TokenOverrides` input are unchanged.
   - Three internal role sources (`ink`, `text-2`, `text-3`) keep the source's heading-ink and body distinctions.
   - In a scope, they follow that scope's `foreground` override for the same appearance, derived inside `tokenOverrideStyle`. The input type is unchanged.
   - `muted-foreground` returns to the kit's own slot value (vx-muted), so meta and body stay distinct (Q6).
   - First-paint evidence now expects H1 to be the independent `ink` source. The semantic `foreground` values stay distinct (§11.3a).
8. **Narrow guard reconciliation and reuse** (§11.5, §11.7). Only these guards change:
   - B-DEPS (and its controls);
   - B-WEB-TOKENS;
   - B-WEB-COMPONENTS;
   - B-WEB-DESIGN (now also the public asset inventory, the @font-face policy and SVG safety);
   - B-NAMING (`web/public` allowed, with exact contents);
   - `APPLICABLE` and the register;
   - the browser matrix.

   One new control check, B-WEB-ASSET-CONTROLS, mirrors the accepted B-DEPS-CONTROLS pattern. Everything else reuses the accepted G047 execution controls: the approved-input wrapper, the fail-closed invocation, harness cleanup and the integrity checks. **No new execution framework.**
9. **Prerequisites and budget.**
   - **Blocking prerequisite:** an owner statement of rights and notice for the selected kit code (Q1, §4.6).
   - **Budget:** B5 is kept. Forecast 4 cycles, range 3–5. Codex judged this plausible but tight; it is reassessed against the accepted post-G003 tree. The fallback split point is defined in §14 (Q8); Codex has authorized no split or extension.

---

## 2. Baseline, inputs and pending post-G003 facts

### 2.1 Observed planning state (observation only, not a binding)

- **Planning checkout.** `<home>/.local/state/platform-evidence/g048-planning/platform`:
  - detached at `dc1ff7c` ("Record Radix architecture decision and specify G048 Vestive adoption");
  - no remote;
  - clean before this Plan.
- **History.** `dc1ff7c` descends from `19f945a`, as does the G047 checkpoint `cbffd6d`. G047 changed exactly 18 paths. `dc1ff7c` changes none of them (`git diff 19f945a dc1ff7c -- web infra` is empty). **The G048 base therefore needs the G047 checkpoint integrated**, which is not yet the case on `staging` (PROGRESS, "Most recently accepted goal").
- **Kit integrity.** The kit manifest `skills/vestive-design-skill/references/shadcn-kit.sha256` (`fcda5a27…9163`) verifies 63/63 OK.
- **`docs/SECURITY.md`.** Empty (`e3b0c442…b855`), as G047's review recorded. DECISIONS.md indexes ADR-001 only.

### 2.2 Inputs read

| Input | SHA-256 (at `dc1ff7c` unless noted) |
|---|---|
| AGENTS.md | `dcc607a1730efe57609f28ae6b8c5c20fa1e0e9cb57e958f4bfc90ab87512764` |
| CLAUDE.md | `cf1cbde9748502ec55f2dfe8243752d34313db781fad06dd7147da76a44f8eac` |
| management/BUDGET_POLICY.md | `b183179b4b34dbe29c024f25ed7cd133c0a51835b1db58381d4509e3bce8f437` |
| docs/ARCHITECTURE.md (§36, §37A, §61) | `a9fd1bde1deec4c9a89ca1d81cbf0f083536bbb28baedcf2614149adbae322d0` |
| docs/DECISIONS.md | `e22b7407c1e4e93a27255576ca9a7fa35659d056b7ecc669bdae23e3c5b1727e` |
| docs/adr/ADR-001-radix-vestive-components.md | `4fb60db722452042ffad8598370687f1d9e1154375a9cb6cc68f1b82aec70756` |
| management/v0/ROADMAP.md | `3048301475e80314e30b1bc90045ba6e557082a70355433cd5ef26304f7b3421` |
| management/v0/PROGRESS.md (planning observation) | `924129cbfaa5b0d4bdd723fe0c2af362f90c6aaaad71b8ca36b135a3bf90e0df` |
| G048 Goal r2 / r1 | `ac7e0a07b7dbc66533393d957bc6f4b76487496bc8eb7fad22bd41169c8765f2` / `252095931b63a5cb962dac414325908642fd2c17bc7404cd34834b8a329fb59d` |
| G048 Spec r2 / r1 | `481fb423c63d2ce92149340bf09ca018e8ed3cc85601e377fddbae432f38e63b` / `ac30cf7ad01d78dfffe1129d5899b7086724e69c10da9ed7529444811ecc4df5` |
| G047 Plan r2 (accepted) | `8aa4a9763e0bb64349a4ae208dca0e6199848090630c7675f9a55d1da564fa65` |
| G003 Plan r4 (read for PENDING items only) | `10a827884bb811f808fc82bc230df9d12b95e237b84d93a36f3067348d00ccb4` |
| management/v0/PROGRESS.md, main-checkout working tree (the r1 CORRECTION LOOP section; read for r2) | `61210afe098480749127167d42b73847b805a3d5269211cbb6682aa568f634d8` (its `3d022ef` HEAD blob is `924129cb…e0df`) |
| docs/agents/CLAUDE-SESSION-HANDOFF.md (main `3d022ef`; read for r2) | `93d186ba80b47f118f2d6018642a801921df52943bd7c9573920f925efe700d1` |

The Goal r2, Spec r2, their r1 histories and ADR-001 are byte-identical between `dc1ff7c` and main HEAD `3d022ef` (empty read-only `git diff`).

**Also read:**
- the G001, G002 and G047 Goals, Specs, Plans and Results, and their acceptance records in PROGRESS (the G047 Result is read from `cbffd6d`);
- `docs/foundation/WEB-SHELL.md` at `cbffd6d`;
- the web application, its tests and fixtures at `cbffd6d`;
- `infra/validation/{run.sh,lib.sh,run-approved.sh,run-approved-controls.sh,render-customizations.tsv}`;
- both web Dockerfiles, `infra/compose/browser.yml` and the archetype templates list;
- **for r2**, from `cbffd6d` via `git show`:
  - `web/tests/browser/appearance.spec.ts`;
  - `support/{color,contrast,modes,page-errors,png}.ts` and `tokens.spec.ts`;
  - `web/playwright.config.ts`;
  - `shell-help-dialog.tsx`, `token-scope.tsx` and `portal-container.tsx`;
  - the token-scope fixture page and `fixtures.ts`;
  - the B-DEPS code in `lib.sh` (`deps_pair_check`, `deps_repo_check`) and `run.sh` (`static_deps_controls`);
  - the browser matrix and the suites that run B-BROWSER-ROOT (`root`) and B-BROWSER-OFF (`variants`).
- The kit `styles/tokens.css` (`d3bcb426…a0ca`), for the `--vx-ink` values.

**Selected kit inputs.** These are the sources this Plan adapts or cites. They are paths under `skills/vestive-design-skill/` and are unchanged by G048.

| Path | SHA-256 |
|---|---|
| `assets/shadcn/components/ui/typography.tsx` | `3cc7d9aad7bae82f6980d5138e2c3cc1f8ea5925c1825d37f1ecefe6e92b0610` |
| `assets/shadcn/components/ui/layout.tsx` | `dada4ef1b5f5365ea90a6ceeea80a9c681942e3e529e9a5b2c3dfd63a8a55935` |
| `assets/shadcn/components/ui/bullet-list.tsx` | `613b3ddee0908893b402240fcd37d6389a14a2b227d0e104d1cc69acc3dd8a55` |
| `assets/shadcn/components/ui/site-header.tsx` | `4ec103f1455830866c90fd3bee800350afbf67434371d164c29dd50002f22fda` |
| `assets/shadcn/components/ui/button.tsx` | `0f7fe0fecd140ebea85994746fb55bb9e000f34d9c3957fb938696a37cb87011` |
| `assets/shadcn/components/ui/dialog.tsx` | `c4d04cb9b490c0f2b661bd0f23e882466bae0fa3ddde1e8b63982b3268484abf` |
| `assets/shadcn/lib/utils.ts` (reference; not copied) | `d1f1e0d62cb8d8d1e04c26e14de842d8a151f75812d81b046c65b5d1fe8e4b27` |
| `assets/shadcn/styles/tokens.css` | `d3bcb4264b41196681096c7b8ea25db39ebc2146f4a5e6eb4cbbf7f97dcba0ca` |
| `assets/shadcn/styles/vestive.css` | `93fab6c1c096776f903ac30ea8a445f114e1ab34e7eb07d3cebcd075b7985484` |
| `assets/shadcn/globals.css` | `a915d08d929c3aef985f76b098fa29073cc4f48740f83bf7fa3a509cfb110cf2` |
| `assets/shadcn/BRAND.md` | `b7c7491ce672107d48e2395b98f39a8bf210d77af3de1a453f4d0f2f613209ef` |
| `assets/shadcn/README.md` | `988944ef8b2abd066891ffeb09748f9f462e11bc5922dc9078e9d8964e379321` |
| `assets/shadcn/tokens.json` | `91009c56ea6657de6ce22356d892ec84d2d5a2d3c401d9ab4150bc3c11853bfd` |
| `references/shadcn-kit.md` | `b4e8cd92c89294bc17057c09d5814956d1108fa452c7a02ceec39c4b104df37e` |
| `references/shadcn-kit.sha256` (manifest) | `fcda5a273daee8117598ab32a244a36d898112c27bd94b3908fbe7a362c79163` |
| `SKILL.md` | `7dfac5faecadb0b297a1be5641b1b86140dd28e6b4d32c78b140bc1cb65c670e` |
| `README.md` (skill) | `80898087d1afc7324840f930748cd045f7e0d5173453be70aa587ec65e3598b3` |
| `styles.css` | `b6d41de3df552af95e90599ec28c320e3c88a4c5486d63828414efa7fbfff73a` |
| `examples/reference.html` (page-composition reference) | `d0dc04398ec296def6caea94419d39b4f77265764122fa97c51fc12b2512b3f3` |
| `examples/Screenshot 2026-10-05 103447.png`, `…103502.png`, `…103548.png` | `4de5ad8e…8e38`, `fb140500…c9589`, `6574fb35…10f2` |

### 2.3 Application baseline used for planning (G047 checkpoint `cbffd6d`; PENDING re-derivation)

These are the hashes of the files G048 changes, as accepted in G047. The post-G003 integrated versions may differ (§2.4), and the reconciliation compares them there.

| Path | SHA-256 at `cbffd6d` |
|---|---|
| web/app/globals.css | `c76ea2dc1ca39d8ab2b3b723e8d0e1103c786689dec1551900fb1d0526bba8f7` |
| web/app/page.tsx | `25b0f6bd47524181b32291595bc8bb314464cde035d2ff0416a2128a78f85efa` |
| web/components/layout/app-shell.tsx | `ccfad26b0f28d4b0ca04f4d6b80ddeaeeccf241fff89ff0b7dba6107e482b976` |
| web/components/layout/token-scope.tsx | `cdb567036581a0e13c36c40ed67512f326b5e6b23b5101dd7371090e7f2dfb8f` |
| web/components/layout/shell-help-dialog.tsx | `f082344df011f09640ace9841ed8cb4fc931021853250090fb2e41d17dee8b1f` |
| web/components/ui/button.tsx | `0315d51bff6c58e0ac859dbaa8cce15eff11d7d09a35340eecc424dea9a1a9e3` |
| web/components/ui/dialog.tsx | `44282c53b7914b6a1daf0524709f0600e2914937e67e0d0f12d7df6be9ec9fd6` |
| web/components/ui/portal-container.tsx | `8310f2f6de911111c727fe5604822be03004075ef92774998787d305d5b19186` |
| web/components.json | `9928283696d86836422959a2c520df83a718d31ffbcab533a14ca8dca2c8d475` |
| web/lib/tokens.ts | `f5035d50851a06d1620eb50d6a98519c3b43bf488c63003ff94c81c1742b9ec7` |
| web/package.json | `2b3ab094b7d951a24419876f37dd2ba4a20cdee26a90b2abdb9baf5183df9fa0` (G003 changes scripts; PENDING) |
| web/package-lock.json | `73a28c09182d72d59157c9f47c1c7e442ab858baeab79d376603654cdc0b05f5` |
| web/playwright.config.ts | `b0624cdddb6814469df5d8c505af9e1de12d01a8a8dc1d704b68275ad9dbfca1` |
| web/tests/browser/{accessibility,design,overlay,shell,tokens}.spec.ts | `fdbe2e9d…72ec17`, `e0190fbe…3c013c51`, `73f9cd01…e297`, `61eaab1d…80380`, `3967b3b0…d4b` |
| web/tests/browser/appearance.spec.ts (changed from r2) | `ba06f0f4dd96ea8eabfc858631ffd955531c917864d2cf01281b83265048477e` |
| web/tests/browser/support/{angle,contrast,modes,png,spacing}.ts | `9ad28e7b…7336`, `33040828…dd27`, `1588769a…2e38`, `c4ab2a8e…300e`, `321d353f…732540` |
| web/tests/fixtures/token-scopes/page.tsx | `f039326077cf0fb58f0b1e6021f4f487ed783834588981dd2a5b0572ef855bd8` |
| infra/validation/lib.sh | `89c6970fe74edc3ee3a0f1faacd06e3974dcfaf40b1a4f813d79920cc95180bd` (G003 changes; PENDING) |
| infra/validation/run.sh | `d5e7e5be4a20ada0df3bbc261ff8d3c0316b03f241e0ec1f30b4f502be941ac8` (G003 changes; PENDING) |
| infra/validation/render-customizations.tsv | `71b9241ed41670eb47b3180ee313cf7f0dc332e7199764c054918cb321003792` (G003 +11 rows; PENDING) |
| infra/validation/run-approved.sh | `620a3e84a68494f9eca9e1bc2ca92a5bf057ebb316c44a31eefd837b4c53c695` (unchanged by G048) |
| infra/docker/web.Dockerfile (rendered template file) | `cb5841802f543c0fce6cacd9f10b97a716533476dcce3108207f0f11086fc37b` |
| infra/docker/web-browser.Dockerfile | `83979271c7ee4ebe884fd6e16ff38d817688757f62fbc543e79cfac686d2705d` |
| docs/foundation/WEB-SHELL.md | `1ce630c31e2a29ab0f8cfa7e07bcf5efcb13cf374d4b643d94b41337a4cae9c0` |

### 2.4 PENDING post-G003 reconciliation list (re-inspect, reconcile, re-review before any implementation)

None of these is assumed. Each needs a recorded value from the accepted, integrated tree, and any difference is reflected in a later revision of this Plan (r3 or later; r2 is the review correction) or in a Codex-approved mechanical reconciliation note.

| ID | Fact to re-inspect | Where it affects this Plan |
|---|---|---|
| PG-1 | The accepted G003 commit and its integration with the G047 checkpoint on the base branch. The integrated hashes of every §2.3 path. | Base for every edit; §9 inventory |
| PG-2 | `web/package.json` after G003. G003 r4 plans `test:unit` and a `check` that includes it, with no dependency change. | §5.1/§5.2: only dependency keys change; B-DEPS |
| PG-3 | `web/package-lock.json` after G003. It should equal `73a28c09…05f5`. | §5.2 planned delta |
| PG-4 | The harness after G003. In `run.sh` and `lib.sh`: the new suites (`contracts`), the B-DEPS and B-WEB-BOUNDARY edits, `--owned-registry` and `--secrets-file`, the `obs_*` checks. | §11.5 guard edits are applied on top of it; static_* function shapes |
| PG-5 | `render-customizations.tsv` row count and the B-RENDER-PLAN count assertion. G003 r4 states +11 rows. | §9: +1 row (web.Dockerfile); the count assertion becomes post-G003 + 1 |
| PG-6 | The `APPLICABLE` list. G003 r4 states 28 → 34. | §11.5: +3 entries |
| PG-7 | `infra/validation/required-checks.tsv` (G003 r4: 69 IDs) and the B-REQUIRED-CHECKS semantics | §11.5: +1 ID (B-WEB-ASSET-CONTROLS) |
| PG-8 | CI: `.gitlab-ci.yml`, `infra/ci/validate.sh`, any `unchanged.allow` constraining harness files, the job timeout, the artifact staging allow-list, and whether later Goals must produce **hosted** evidence | §11.9; Q7 |
| PG-9 | Browser matrix contents and per-run test counts after G003. G047: 103 root, 103 disabled, 2 rename. | §11.2 additions; Result counts |
| PG-10 | Base images and pins (`node:24-slim`, the Playwright image digest, the validation-tools image) and any G003 image policy | §5.2 lockfile container; B-DEPS image rule |
| PG-11 | The approved-input wrapper interface and the S0 gate. G003 r4 says `run-approved.sh` is unchanged. | §11.7 commands |
| PG-12 | README and OPERATIONS blocks after G003 (`platform:contracts`, `platform:ci`) | Docs execution regressions |
| PG-13 | Any G003 change to `web/tests/**` (including `appearance.spec.ts`, which G048 changes from r2), `web/components/**`, the fixtures or `web/playwright.config.ts`. G003 r4 lists them as unchanged. | §11.3 and §11.3a mapping |
| PG-14 | The evidence and artifact conventions G003 adds (B-ARTIFACT-SCAN, staged artifact names) | §11.6 new attachments |
| PG-15 | Whether G003 adds hadolint or other checks to `infra/docker/web.Dockerfile` | §9 Dockerfile edit |

### 2.5 Prerequisites and execution binding (performed by their owners; documented, not performed)

| # | Prerequisite |
|---|---|
| P1 | G003 implementation ACCEPT, and integration of G003 and the G047 checkpoint on the base branch |
| P2 | The PG-1…PG-15 reconciliation recorded. This Plan is revised (r3 or later) or mechanically reconciled, then re-reviewed. |
| P3 | **Owner rights and notice statement** for the six selected modules (Q1, §4.6), committed where Codex directs |
| P4 | Codex Plan ACCEPT of the reconciled revision, including the remaining dispositions (Q3–Q7, Q9–Q13; Q1 is P3). Q2 was answered and Q8 directed on 2026-10-06 (§0.1, §17). |
| P5 | Byte-exact integration of the accepted Plan; a committed base `BASE` containing every bound input |
| P6 | The execution worktree at `BASE` (for example `.claude/worktrees/g048-exec` on `worktree-g048-exec`), created through the supported workflow |
| P7 | A supervisor-created snapshot and binding (`SNAP`, `SNAP_SHA`, `N`, `RE`) through the existing mechanism. The proposal is below; Codex decides. |
| P8 | Network access at implementation time to `registry.npmjs.org` (lockfile) and `raw.githubusercontent.com` (font intake), from the host and the Node container |
| P9 | If Q7 requires hosted evidence: the separately authorized push and pipeline actions and runner availability |

**Proposed snapshot** (Codex decides):
- **Entries.** 28:
  - the G048 Goal, Spec and accepted Plan;
  - PROGRESS.md;
  - ADR-001;
  - the 22 skill inputs of §2.2: the six modules, `lib/utils.ts`, the two `styles/*.css`, the kit `globals.css`, `BRAND.md`, the kit `README.md`, `tokens.json`, `shadcn-kit.md`, `shadcn-kit.sha256`, `SKILL.md`, the skill `README.md`, `styles.css`, `reference.html` and the three screenshots;
  - the P3 rights record. Drop it if the record lives in PROGRESS, which leaves 27.
- **Path regex.**
  ```
  ^(management/v0/(goals/G048-vestive-components-fonts-blueprints\.md|specs/G048-vestive-components-fonts-blueprints-spec\.md|plans/G048-vestive-components-fonts-blueprints-plan\.md|PROGRESS\.md)|docs/adr/ADR-001-radix-vestive-components\.md|skills/vestive-design-skill/.+)$
  ```
  Extend it for P3 if needed.
- **Plan.** `--plan management/v0/plans/G048-vestive-components-fonts-blueprints-plan.md`.
- Historical snapshots and `dc1ff7c` are not execution authority.

---

## 3. Implementation approach

1. **Intake.** Done in `$EV`, outside the repository:
   - download the seven font files at the pinned commit;
   - verify each git blob SHA-1 and size (§7.1);
   - generate the lockfile delta in a throwaway Node container (§5.2);
   - verify both against this Plan before any copy into the repository.
2. **Theme layer** (`globals.css`, `lib/tokens.ts`):
   - the palette change (muted-foreground) and the three internal role sources;
   - the structural tokens (`--spacing-section: 88px`, `--spacing-cell: 36px`, `--angle` = the source polygon);
   - the source font stacks;
   - the delimited bundled-font `@font-face` block;
   - the blueprint layer block;
   - removal of the G047 `angled` utility.
3. **Components.** The six adopted modules, the portal contract (`portal-container.tsx`, `token-scope.tsx`), the blueprint layer component.
4. **Consumers.** `app-shell.tsx`, `shell-help-dialog.tsx`, `page.tsx`, then the test fixture.
5. **Tests.** Support modules, then the specs (§11).
6. **Infrastructure and guards.** Dockerfiles (2 lines), then `lib.sh`, `run.sh`, the register and required checks (§11.5).
7. **Documentation.** `WEB-SHELL.md`, then `THIRD_PARTY_NOTICES.md`.
8. **Validation.** Only through the approved wrapper and harness (§11.7, §13).

**Out of scope, unchanged:**
- the appearance engine (`lib/appearance.ts`, the inline init, `appearance-control.tsx`, `layout.tsx`);
- config and the BFF;
- core and intelligence;
- schemas;
- skills;
- the archetype templates and manifest.

---

## 4. Source adoption (R1, R2)

### 4.1 Lineage, exports and production consumers

| Source module (kit path, SHA-256) | Target | Exports kept | Production consumers (after G048) | Fixture-only use |
|---|---|---|---|---|
| `typography.tsx` `3cc7d9aa…0610` | `web/components/ui/typography.tsx` | Kicker, headingVariants, Heading, textVariants, Text, TextLink (Emph omitted, X21) | `page.tsx`: Kicker (overview), Heading h1/h2/h3, Text lead/body/index. `app-shell.tsx`: TextLink (skip link), Text meta (footer). | Heading `cta`, every Text variant, TextLink in each scope |
| `layout.tsx` `dada4ef1…5935` | `web/components/ui/layout.tsx` | Container, sectionVariants, Section, RuledGrid | `page.tsx`: three Sections (ground, surface, ground) and RuledGrid (Getting around). `app-shell.tsx`: Container (footer). `site-header.tsx`: Container. | Section `raised` tone panel |
| `bullet-list.tsx` `613b3ddee…a55` | `web/components/ui/bullet-list.tsx` | BulletList, BulletItem | `page.tsx` About list; `shell-help-dialog.tsx` tips | — |
| `site-header.tsx` `4ec103f1…2fda` | `web/components/ui/site-header.tsx` | SiteHeader | `app-shell.tsx` header (configured identity, Primary nav, appearance control and help) | — |
| `button.tsx` `0f7fe0fe…b87011` | `web/components/ui/button.tsx` (replaces the G002/G047 Base UI file) | buttonVariants, Button (`asChild`) | `shell-help-dialog.tsx` trigger (`DialogTrigger asChild` → `Button size="sm"`) | Every variant (default, secondary, outline, muted, destructive, ghost, link) enabled and disabled; sizes sm, default, lg, icon; `block`; `asChild` link |
| `dialog.tsx` `c4d04cb9…4abf` | `web/components/ui/dialog.tsx` (replaces the Base UI file) | Dialog, DialogTrigger, DialogPortal (adapter), DialogClose, DialogOverlay, DialogContent, DialogHeader, DialogFooter, DialogTitle, DialogDescription | `shell-help-dialog.tsx` | The `?open=` fixture dialog: DialogFooter, `DialogClose asChild` |

**Rendered-structure evidence.** S-FIDELITY asserts that each module's `data-slot` is present in the served production page:
- typography: `kicker`, `heading`, `text`, `text-link`;
- layout: `container`, `section`, `ruled-grid`;
- bullet-list: `bullet-list`, `bullet-item`;
- site-header: `site-header`;
- button: `button`;
- dialog, when open: `dialog-overlay`, `dialog-content`, `dialog-header`, `dialog-title`, `dialog-description`.

Copying only these modules satisfies the "no unused-file" rule:
- **Not copied:** `components/index.ts`, `lib/utils.ts`, `globals.css`, `styles/*` and every other kit file.
- **Imports:** the source `@/lib/utils` resolves unchanged to the existing `web/lib/utils.ts` (`export { cn } from "cn"`).

### 4.2 Translations (T): value-identical role and name mappings

Source class and variable names are translated to the application's existing scope-resolved tokens. These rows change names only; each target resolves to the same value as the source in both appearances (tested by S-PALETTE and S-FIDELITY).

| T# | Source | Application | Value identity |
|---|---|---|---|
| T1 | `text-gold`, `border-gold` | `text-accent`, `border-accent` | accent = vx-gold (`#8a5a00` / `#f8ce8a`) |
| T2 | `border-rule`, `bg-rule` | `border-border`, `bg-border` | border = vx-rule (`#d4cfc8` / `#262626`) |
| T3 | `bg-surface`; `var(--vx-surface)` | `bg-muted`; `var(--muted)` | muted = vx-surface (`#ebe8e3` / `#101010`) |
| T4 | `bg-raised`; `var(--vx-raised)` | `bg-popover`; `var(--popover)` | popover = vx-raised (`#ffffff` / `#131313`) |
| T5 | `var(--vx-bg)` | `var(--background)` | identical |
| T6 | ghost `hover:bg-accent` (the kit's accent slot is the rule colour) | `hover:bg-border` | the kit's `--accent` → vx-rule; matches the K11 role mapping |
| T7 | destructive `text-destructive-foreground` | `text-background` | the kit's destructive-foreground = vx-bg (K10) |
| T8 | `font-body` | `font-sans` | the source body stack, adopted verbatim as `--font-sans` |
| T9 | `max-w-[var(--vx-container)]`, `px-[var(--vx-gutter)]` | `max-w-shell`, `px-gutter` | 1360px; `clamp(20px, 4vw, 40px)` |
| T10 | `py-[var(--vx-section-y)]` | `py-section` | `--spacing-section` set to the source **88px** (supersedes G047's clamp) |
| T11 | `px-[var(--vx-cell-pad)]` (RuledGrid cells, DialogContent) | `px-cell` | new `--spacing-cell: 36px` |
| T12 | `[clip-path:var(--vx-angle)]` | `var(--angle)` on the visual layer (X7) | `--angle` set to the source `polygon(15% 0%,100% 0%,100% 35%,85% 100%,0% 100%,0% 65%)` (supersedes G047's 12px cut) |
| T13 | BulletItem marker `bg-gold` | `border-t-2 border-accent` with the same `h-[2px] w-[14px]` | Identical 14×2 painted rule. The accent token stays ink and border only (B-WEB-TOKENS Q10). |
| T14 | `text-ink`, `text-text-2`, `text-text-3`, `hover:text-ink`, `hover:border-ink` | Kept verbatim. Backed by the internal roles of §6.2 (`ink`, `text-2`, `text-3`). | The source values |

### 4.3 Exceptions (X): every deviation from a source declaration

Each row gives the source declaration, the replacement, the reason (Spec R2 category), the affected states and widths, and the verification.

| X# | Source declaration | Replacement | Reason / category | Affected | Verification |
|---|---|---|---|---|---|
| X1 | Text `index`: `text-faint` | `text-muted-foreground` | Measured contrast: faint is 3.5–4.4:1, below 4.5 for normal text (kit §5) | All widths, both appearances | S-CONTRAST (index ≥ 4.5); S-FIDELITY colour |
| X2 | `primary` vx-blue `#0075ff` with a white label | Retain the `#0062d6` palette value (G047 K1) | Measured contrast: white on `#0075ff` is 4.21:1. Labels are 14–15px at weight 500, so normal text. | Filled default variant: rest, hover, press, focus | S-CONTRAST, S-STATES |
| X3 | `secondary` vx-green `#66892f` with a white label | Retain `#526e26` (Light) and `#709734` with a `#0d0d0d` label (Dark) (K2, K3) | Measured contrast 4.05:1 | Secondary variant | Same |
| X4 | Light `destructive` `#b8322a` | Retain `#af3028` (K4) | Measured contrast: hover 4.44:1 | Destructive, Light | Same |
| X5 | — (reserved; T7 is value-identical) | — | — | — | — |
| X6 | Overlay `bg-[rgba(0,0,0,.72)]` (theme-fixed) | `bg-overlay`: Dark `rgb(0 0 0 / .72)` (identical); Light `rgb(13 13 13 / .45)` (K8) | Semantic-token translation. The source scrim is tuned for Dark; no colour literals in components. | Open dialog, Light | S-PALETTE, S-PORTAL overlay oracle |
| X7 | Button: `[clip-path:…]` on the **element**; the outline inner shape on `::before` | Fills on pseudo-element visual layers: `before:absolute before:inset-0 before:-z-10 before:content-[''] before:[clip-path:var(--angle)] before:bg-…` on the angled variants (default, secondary, destructive, muted, outline). The element is unclipped with a transparent background. | Measured focus correction (kit §9.7): the element clip removes the 2px offset focus outline. Labels and focus stay whole. | All angled variants, all states | S-ANGLE (label corners inside the polygon; ring present **and ≥ 3:1 on all four sides**); S-STATES clip-path on `::before`; element clip-path `none` |
| X8 | Outline frame `bg-rule` (1.27–1.55:1) | Frame layer `before:bg-input` (vx-muted, ≥ 3:1). Inner `after:inset-[2px] after:bg-background after:[clip-path:var(--angle)]`. **The angled 2px frame is kept.** | Measured contrast: 1.4.11 control boundary (K7, K12 reasoning; the rectangular G047 frame is **not** reused) | Outline, all states | S-STATES (frame pixel = input; inner = background; corner pixel = scope surface); S-CONTRAST outline boundary ≥ 3:1 |
| X9 | Button: no focus style (relies on the kit's global `:focus-visible` base, which is not imported) | Base `focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring focus-visible:outline-solid` | Focus correction (no global reset imported) | Every variant, focused | S-STATES focus outline; S-ANGLE |
| X10 | — (reserved) | — | — | — | — |
| X11 | Dialog Close: no focus style | Same focus classes on Close | Focus correction | Open dialog | S-PORTAL focus-ring colour; S-OVERLAY |
| X12 | DialogContent: no height limit | `max-h-[calc(100dvh-2rem)] overflow-y-auto` | Reflow, zoom and text spacing: content and Close stay reachable at 640×400, at 200% and under spacing. No visual change at 1280×800. | Short viewports | S-RESPONSIVE (dialog in viewport); S-TEXT-SPACING modal |
| X13 | DialogHeader `flex flex-col gap-3` | Adds `pr-9` | Text-spacing and reflow: the title must clear the 32px Close. At 320px the content box is 204px and the Close overlaps its right 14px. | All widths; visible only when the title wraps | S-TEXT-SPACING modal overlap checks |
| X14 | Dialog has no `"use client"`; `DialogPortal = DialogPrimitive.Portal` (body by default) | Adds `"use client"`. `DialogPortal` becomes the scope adapter (§5.4). | Radix, App Router and portal compatibility; scope composition (ADR-001 §3–4) | Every dialog | S-PORTAL, S-PORTAL-INIT, S-HYDRATION-DEV |
| X15 | SiteHeader: `logoSrc` `<img alt="Vestive">`; marketing `DEFAULT_LINKS`; unlabelled `<nav>` | Props `brand: ReactNode` (configured text identity), required `links`, `navLabel` (rendered as `aria-label`) and `action`. No image or default links. | Configured neutral content and semantic HTML (Spec R2) | Header | S-IDENTITY, S-LANDMARKS, B-WEB-DESIGN literal scan |
| X16 | SiteHeader `sticky top-0 z-50` | Static (`relative`) | Accessibility: in-page anchor targets (nav and skip) would scroll under a sticky bar (2.4.11). The wrapped header at 320 and 640×400 covers a large share of the viewport. | All widths (consistent behaviour) | S-NAV (target heading in the viewport **and not covered**, via `elementFromPoint`); S-SKIP |
| X17 | Nav link `text-foreground no-underline transition-colors hover:text-gold` (no target size, no focus style) | Adds `inline-flex min-h-8 items-center` and the focus classes (hover per T1) | Target size 2.5.8 (S-TAB ≥ 24px); focus correction | Nav links | S-TAB; S-FIDELITY (font metrics unchanged) |
| X18 | RuledGrid: `minWidth = 860`, inline `gridTemplateColumns`, wrapper `overflow-x-auto` | `minWidth` prop removed. Columns from `--ruled-columns` with `grid-cols-1 md:[grid-template-columns:repeat(var(--ruled-columns),minmax(0,1fr))]`. Wrapper `overflow-x-auto` removed; `-mt-[14px] pt-[14px]` kept. | Reflow 1.4.10 (Spec R6: "Reflow the source grid") | Below 768px: 1 column (2px gaps become horizontal rules). 768px and above: N columns. | S-RESPONSIVE (columns 1/3/3/1/3); no horizontal scroll |
| X19 | Heading h1: no wrapping rule | Adds `[overflow-wrap:anywhere]` at the call site | Reflow of long configured identity | H1 | S-LONG-IDENTITY, S-RESPONSIVE |
| X20 | TextLink: no focus style | Adds the focus classes | Focus correction (skip link) | Focused | S-SKIP, S-TAB, S-CONTRAST skip-link ring |
| X21 | `Emph` export (`text-orange`) | Omitted from the copied module | Unconsumed marketing emphasis; avoids adding an unreviewed orange role (T7 precedent). Q4 offers retention. | — | B-WEB-COMPONENTS `data-slot` inventory excludes `emph` |
| X22 | Source comments naming the brand or marketing CTA (button.tsx, site-header.tsx) | Neutral comments plus a lineage header | B-WEB-DESIGN literal scan; neutral content (comment-only) | — | B-WEB-DESIGN |

**Retained without exception** (verified against the source table, §4.7):
- the source 88px section padding and 36px cell padding at every width;
- the source H1 and H2 minimum clamps (2.6rem and 1.9rem);
- the source 15% angle;
- the help trigger at source size `sm`;
- the source dialog widths, paddings, gap, frame, close geometry and type;
- ghost/link gold resting ink and the link border underline;
- no dialog motion.

**Superseded G047 choices, explicitly not reused:**
- compact h-8/h-9 buttons (K19);
- the rectangular 1px outline (K12);
- the 12px angle (K18);
- the dialog title and padding clamps (K14);
- ghost resting ink inherited (K13);
- the 100ms dialog fade (K17).

**Conditional pre-declared fallback D-ANGLE** (applied only if measured; reported for Codex review):
- **Trigger:** S-ANGLE or S-TEXT-SPACING (`angled` items) report a label-box corner outside the 15% polygon in a required state.
- **Change:** `--angle` becomes `polygon(min(15%, 30px) 0%, 100% 0%, 100% 35%, max(85%, calc(100% - 30px)) 100%, 0% 100%, 0% 65%)`. This is identical to the source for buttons up to 200px wide, capping the cut at the `sm` padding.
- **Test change:** `support/angle.ts` `polygonPoints` gains `min()`/`max()` parsing.
- **Analytic margins** at 1280×800 (planning computation, not evidence):
  - help trigger: label corner at 30px against a cut of 12.5px at the label top;
  - at 200% text: 30px against 26px.

  Both pass, so D-ANGLE is not expected.

### 4.4 Per-module alteration list (source → target)

Every adopted file gets a two-line lineage header (no brand literal):

```tsx
// Adapted from the design kit components/ui/<name>.tsx (sha256 <full>); changes per docs/foundation/WEB-SHELL.md (G048).
```

`button.tsx` and `dialog.tsx` add a third line: `// Derived from shadcn/ui (MIT); see web/THIRD_PARTY_NOTICES.md.`

- **typography.tsx**
  - T1 (Kicker, TextLink); T8 (lead, body); X1 (index); X20 (TextLink focus); X21 (Emph removed).
  - Everything else verbatim: `forwardRef`, the `as` props, the CVA variants including `cta`, `m-0`, the clamp sizes, weights, tracking, max-widths and `text-ink`/`text-text-2`/`text-text-3`.
- **layout.tsx**
  - Container T9. `sectionVariants`: T2, T3, T4. Section inner T10. CELL_BG T3–T5.
  - RuledGrid: X18 and T2/T11; the doc comment loses "scrolls rather than orphaning".
  - Retained: the `contained` and `containerClassName` API and `data-slot` values.
- **bullet-list.tsx**
  - T13. Everything else verbatim: `gap-3`, `gap-3.5`, `font-label text-[1.08rem] leading-[1.35] text-text-2`, `mt-[11px]`, `flex-[0_0_14px]`.
- **site-header.tsx**
  - X15, X16, X17, T2, X22.
  - Retained: `Container` with `flex min-h-[76px] flex-wrap items-center justify-between gap-x-10 gap-y-6 py-3`; nav `flex flex-wrap gap-9 font-display text-[.94rem] uppercase tracking-[1.2px]`; `bg-background`.
- **button.tsx**
  - X7, X8, X9, T2/T6/T7/T1 (per variant), X22.
  - Retained verbatim:
    - the base `relative isolate inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 whitespace-nowrap border-0 font-display font-medium uppercase leading-none tracking-[1.5px] transition-[filter,color,background-color] duration-150 hover:brightness-[1.12] active:brightness-90 disabled:pointer-events-none disabled:opacity-50 [&_svg]:…`;
    - all four sizes, `block` and `compoundVariants` (`link` → `px-0 py-0 pb-1`);
    - `Slot` and `asChild`, and `forwardRef`.
  - Ghost and link keep `[clip-path:none]`. It is a no-op after X7, kept for lineage.
  - **Merge reliance.** The source relies on Tailwind class merging (link versus size padding; `block` versus nowrap/leading). The existing `cn` 0.4.0 states tailwind-merge v3 parity for Tailwind v4. This is **verified, not assumed**: S-STATES asserts the computed padding, white-space and line-height of `link`, `block` and every size.
- **dialog.tsx**
  - X14, X6, T2, T11, X12, X11, X13, T8.
  - Retained:
    - Content `fixed top-1/2 left-1/2 z-50 grid w-[calc(100%-40px)] max-w-[560px] -translate-x-1/2 -translate-y-1/2 gap-5 border-2 … bg-popover pt-[34px] pb-10 text-popover-foreground outline-none`;
    - the Close `aria-label="Close"`, `absolute top-5 right-5 grid size-8 … text-muted-foreground hover:text-ink` with the two crossed 2px rules (no icon library);
    - Title `font-display text-[1.9rem] font-semibold uppercase leading-[1.1] tracking-[-.5px] text-ink`;
    - Description `text-[1.04rem] leading-[1.65] text-text-3`;
    - Footer verbatim.
  - `React.ElementRef` is kept. It is deprecated in `@types/react` 19 but still exported, so no change is needed.

### 4.5 Unadopted source content

- **Kit modules.** The other 40 `components/ui` modules, `icons.tsx`, `menu-styles.ts` and `index.ts`.
- **Styling.** Kit `globals.css`, `styles/*` (the global resets, `.light`/`.dark`/`[data-theme]` selectors, `::selection`, the base `:focus-visible`, ribbon/node keyframes, `.vx-hatch`); `tokens.json` samples.
- **Brand and marketing.** `public/brand/*`, all marketing copy, `next-themes`, Google Fonts links, `clsx` and `tailwind-merge`.

The unadopted kit roles are recorded in WEB-SHELL.md:
- `vx-faint` (X1);
- `vx-orange`/Emph (X21);
- `vx-star`;
- the chart colours;
- the ribbon and node field;
- the `card`/`accent-foreground` slots.

### 4.6 Source rights and notices (blocking prerequisite P3)

**Known facts (kit guide §10):**
- The kit has no licence, copyright, author or origin statement.
- BRAND.md says the "brand pieces" (`typography`, `layout`, `bullet-list`, `site-header`) come "straight from the marketing site".
- It says the "shadcn set" (`button`, `dialog`) is "restyled".
- upstream shadcn/ui is MIT, "Copyright (c) 2023 shadcn" (https://raw.githubusercontent.com/shadcn-ui/ui/main/LICENSE.md, read 2026-10-06).
- The owner's instruction authorizes adoption but does not name the rights holder or any notice obligations.

**Plan position (Q1).** No module is copied until P3 is satisfied.
- **P3 requested content.** An owner statement:
  1. that the owner (or the owner's company) holds rights to the four brand modules and the restyled `button`/`dialog` and permits their use and modification in this repository;
  2. the attribution or notice text, if any, the owner requires;
  3. confirmation that `button.tsx` and `dialog.tsx` derive from shadcn/ui (MIT), so the MIT notice must be retained.
- **Placement.**
  - The owner statement and lineage table go in `docs/foundation/WEB-SHELL.md`. That file is not scanned by B-WEB-DESIGN, and B-LITERALS checks only manifest identity values.
  - `web/THIRD_PARTY_NOTICES.md` carries the full shadcn/ui MIT text for `components/ui/button.tsx` and `dialog.tsx`. It also states that the G002-generated predecessors had the same origin, and points to each bundled font's `OFL.txt` with its copyright line.
- **Literal-scan conflict.** If the owner requires an in-file notice naming the brand or company, B-WEB-DESIGN's literal scan would reject it. That needs an exact-line exception, which is an escalation (Q1b), never a blanket allowance.
- **Radix.** The packages are MIT, "Copyright (c) 2022 WorkOS". They are carried as npm dependencies with their own LICENSE, and recorded in the Result licence inventory, consistent with G002's dependency practice.
- **Fonts and imagery.** Fonts: §7.3. The blueprint SVG is original, project-authored geometry with no third-party source; its provenance is in §8.6.

### 4.7 Source reference expectations (R2; typed independently into `web/tests/browser/support/source-reference.ts`)

**Measurement conditions.** 1280×800, deviceScaleFactor 1, default text, fonts loaded, reduced motion.
- **Tolerance:** ±1 CSS px for lengths.
- **Exact match** for font weight, the family list (compared with quotes normalized), border widths, radius, box-shadow, text-transform and variant selection.
- **Typing:** every value is typed from the source declaration and cited by `file:class`, never read from application CSS.

| Element | Expected |
|---|---|
| Container | `max-width` 1360px; padding-inline 40px (`clamp(20px, 4vw, 40px)` with 4vw = 51.2); content 1200px |
| Section | `border-bottom` 2px solid border; container padding-block 88px; tones: overview and about = background, getting-around = muted |
| Kicker | Barlow Condensed stack; 16px; line-height 19.2px; weight 600; letter-spacing 2.5px; uppercase; accent |
| H1 | Oswald stack; 71.68px (`clamp(2.6rem,5.6vw,5rem)`); line-height 71.68px; 700; −1px; uppercase; ink |
| H2 | 43.2px (`clamp(1.9rem,3.4vw,2.7rem)` → max); line-height 47.52px; 600; −0.5px; ink |
| H3 | 21.6px; line-height 25.92px; 500; uppercase; accent (reference.html composition: `color: var(--vx-gold)`, `margin-top: 14px`) |
| Lead | Roboto stack; 19.04px; line-height 30.8448px; 400; `max-width: 52ch`; text-2 |
| Body (cells, About, dialog description) | 16.64px; line-height 27.456px; 400; `max-width: 62ch` (Text body); text-3 |
| Meta (footer) | Barlow; 17.28px; line-height 23.328px; 400; muted-foreground |
| Index | mono stack; 11.52px; letter-spacing 1px; muted-foreground (X1) |
| RuledGrid | border-top and border-bottom 2px border; gap 2px; 3 columns at 768px and above; cells padding 34 / 36 / 40 / 36; cell background = band |
| BulletList / BulletItem | list gap 12px; item gap 14px; Barlow 17.28/23.328, 400, text-2; marker 14×2, margin-top 11px, accent |
| Header | `border-bottom` 2px; container min-height 76px; padding-block 12px; column gap 40px; row gap 24px; nav gap 36px; nav links Oswald 15.04px, 400, 1.2px, uppercase, foreground; link min-height 32px (X17) |
| Help trigger (Button default, `sm`) | padding 14px 30px; Oswald 14.08px; weight 500; letter-spacing 1.5px; uppercase; line-height 14.08px; height 42.08px; radius 0; element `clip-path: none` (X7); `::before` clip = `--angle` = source polygon; `::before` background primary; label primary-foreground; hover `filter: brightness(1.12)`; active `brightness(0.9)` |
| Button sizes (fixture) | sm 14/30 and 14.08px; default 18/40 and 15.2px; lg 22/48 and 15.2px; icon 44×44; link padding 0/0/4px with border-bottom 2px accent |
| Dialog content | width 560px; border 2px solid border (all sides); padding 34 / 36 / 40 / 36; gap 20px; background popover; box-shadow none |
| Dialog title | Oswald 30.4px; line-height 33.44px; 600; −0.5px; uppercase; ink |
| Dialog description | Roboto 16.64px; line-height 27.456px; text-3 |
| Dialog header | gap 12px |
| Close | 32×32 at 20px from the padding-box top and right; mark 14×14; two 2px bars at ±45°; colour muted-foreground |
| Overlay | `bg-overlay`; `backdrop-filter: none` |
| TextLink (skip link, focused) | Oswald 16px; 1.5px; uppercase; accent; border-bottom 2px accent; padding-bottom 4px |

**Responsive expectations** (stated breakpoints):

| Range | Expectation |
|---|---|
| Below 768px | RuledGrid 1 column (X18) |
| All widths | Header static (X16); wraps at narrow widths |
| Below 1024px | Blueprint in the 88px bottom band (§8.4) |
| 320×640 | Gutter 20px; H1 41.6px with `overflow-wrap:anywhere`; H2 30.4px; dialog width 280px with a 204px content box; header `pr-9` (X13) |
| 640×400 (1280 at 200% zoom) | Gutter 25.6px; dialog `max-height: calc(100dvh - 2rem)` = 368px, scrollable (X12) |

**Reference comparison evidence:**
- `source-fidelity.json` (expected, measured, delta, verdict, source citation);
- the Result shows each after-screenshot beside the supplied kit screenshots (cited by path and hash; not copied into the repository).

---

## 5. Radix migration

### 5.1 Exact dependency delta

| Change | Package | Exact version | Category | Licence | Registry integrity (re-verified in the lockfile at intake) |
|---|---|---|---|---|---|
| add | `@radix-ui/react-dialog` | 1.1.23 | dependencies | MIT | `sha512-Ksw4WeROkO4rC9k/onilX/Ao2Cr1ku1unMNH+XSCcP4jSXYu7HDsg9n4ojMjVb22XpYjAQ9qfrFlVbru1vXDUA==` |
| add | `@radix-ui/react-slot` | 1.3.3 | dependencies | MIT | `sha512-qx7oqnYbxnK9kYI9m317qmFmEgo6ywqWvbTogdj7cL9p3/yx4M48p7Rnw5z3H890cL/ow/EeWJsuTykeZVXP5Q==` |
| remove | `@base-ui/react` | 1.8.0 | dependencies | MIT | — |

**Compatibility.**
- Peer ranges: React and React DOM `^16.8 || ^17 || ^18 || ^19 || ^19.0.0-rc`. The installed React 19.3.0 is in range.
- `@types/react` and `@types/react-dom` are optional peers.
- Dialog 1.1.23's own Slot dependency is exactly 1.3.3, so the top-level pin dedupes.

**Version rationale and choice.**
- The 1.1.23/1.3.3 pair carries:
  - CI provenance attestations (from 1.1.21/1.3.1);
  - the React 19 Slot ref-loop fix (slot 1.2.5);
  - the removal of dev-time console warnings for a missing title or description (dialog 1.1.17). S-ERRORS forbids console errors.
- Dialog 1.2.0 and Slot 1.4.0 were published 2026-10-05, from radix-ui/primitives commit `d5767dcb`. They add multiple `aria-describedby` and a nested portalled-layer fix, which G048 does not need.
- **Q2 answered (Codex, 2026-10-06).** 1.1.23 and 1.3.3 are valid registry versions, and no upgrade to the newer pair is required. A different pair would be a different reviewed delta.

**Planned transitive delta** (planning evidence from registry metadata).
- **24 additions:**
  - `@radix-ui/`:
    - `primitive` 1.1.7;
    - `react-compose-refs` 1.1.5;
    - `react-context` 1.2.2;
    - `react-dismissable-layer` 1.1.19;
    - `react-focus-guards` 1.1.6;
    - `react-focus-scope` 1.1.16;
    - `react-id` 1.1.4;
    - `react-portal` 1.1.17;
    - `react-presence` 1.1.10;
    - `react-primitive` 2.1.10;
    - `react-use-callback-ref` 1.1.4;
    - `react-use-controllable-state` 1.2.6;
    - `react-use-effect-event` 0.0.5;
    - `react-use-layout-effect` 1.1.4;
    - plus the two direct packages.
  - Other packages:
    - `aria-hidden` ^1.2.4 → 1.2.6;
    - `react-remove-scroll` ^2.7.2 → 2.7.2;
    - `react-remove-scroll-bar` 2.3.8;
    - `react-style-singleton` 2.2.3;
    - `use-callback-ref` 1.3.3;
    - `use-sidecar` 1.1.3;
    - `get-nonce` 1.0.1;
    - `detect-node-es` 1.1.0.
  - Licences: all MIT. `tslib` 2.8.1 (0BSD) is already present as a runtime entry and is reused.
- **9 removals** (computed from the current lockfile graph as reachable only through `@base-ui/react`): `@base-ui/react` 1.8.0, `@base-ui/utils` 0.4.0, `@floating-ui/core` 1.8.0, `@floating-ui/dom` 1.8.0, `@floating-ui/react-dom` 2.1.9, `@floating-ui/utils` 0.2.12, `@babel/runtime` 7.29.7, `reselect` 5.3.0 and `use-sync-external-store` 1.7.0. All MIT.
- **Flag change:** `@types/react-dom` goes from `dev` to `devOptional`. Radix declares it as an optional peer, and the existing B-DEPS rule accepts `devOptional` only when such a declaration exists, so it passes. `@types/react` stays `devOptional` (lucide-react and Radix declare it).
- **Deviation handling:**
  - A caret-ranged transitive package resolving to a newer patch within range is **recorded** in the Result.
  - Any package outside this name set, a licence other than MIT/0BSD/ISC, or any change to an unrelated lock entry **stops for Codex review**.

**Unchanged direct dependencies:**
- `class-variance-authority` 0.7.1;
- `cn` 0.4.0;
- `lucide-react` 1.52.0. After G048 it is consumed only by the fixture's `icon` size button, because the source Close uses rules rather than a glyph. Removing it is outside the authorized delta (Q10).
- `tw-animate-css` 1.4.0 (the ejected block still imports it);
- `next`, `react` and the dev tools.

No umbrella `radix-ui`, `clsx`/`tailwind-merge` or other Radix package is added.

### 5.2 Lockfile procedure (implementation time; no host Node)

**Container command.** In `$EV/lockgen/`, outside the repository, holding copies of the two package files:

```sh
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -e npm_config_update_notifier=false \
  -v "$EV/lockgen:/work" -w /work node:24-slim sh -ceu '
  npm pkg delete "dependencies.@base-ui/react"
  npm install --package-lock-only --save-exact --ignore-scripts --no-audit --no-fund \
    @radix-ui/react-dialog@1.1.23 @radix-ui/react-slot@1.3.3
  npm --version'
```

- No volume is created, no install scripts run, the container is removed and nothing is written to the repository.
- The image digest and npm version are recorded.

**Delta check** (read-only, using the harness's yq image to list `packages` keys, versions, dev flags and integrities):
- the before/after difference must equal §5.1 exactly;
- package.json may differ only in the three dependency keys;
- the two direct integrities must equal §5.1.

Only then are both files copied into `web/`. `npm ci` in the web and browser image builds then proves lockfile consistency, which B-DEPS also checks independently.

### 5.3 Consumer and API migration inventory

The old Base UI usage at `cbffd6d`, and its replacement:

| Old usage | Location | Replacement |
|---|---|---|
| `import { Button as ButtonPrimitive } from "@base-ui/react/button"`; `ButtonPrimitive.Props` | ui/button.tsx | Source `forwardRef` `<button>`/`Slot`; `ButtonProps extends ButtonHTMLAttributes & VariantProps & { asChild?: boolean }` |
| Button sizes `xs`, `icon-xs`, `icon-sm`, `icon-lg` (and the G047 `default` h-8) | ui/button.tsx (no consumer) | Removed; the source sizes are sm, default, lg and icon, plus `block`. No call site uses the removed sizes. |
| Variant set (default, outline, secondary, ghost, destructive, link) | fixture page | Source set; `muted` added (fixture) |
| `import { Dialog as DialogPrimitive } from "@base-ui/react/dialog"`; `Root`, `Trigger`, `Portal`, `Close`, `Backdrop`, `Popup`, `Title`, `Description` `.Props` types | ui/dialog.tsx | `import * as DialogPrimitive from "@radix-ui/react-dialog"`: Root, Trigger, Portal (via the adapter), Close, Overlay, Content, Title, Description |
| `render={<Button … />}` composition | shell-help-dialog.tsx (trigger); ui/dialog.tsx (Close, Footer close) | `asChild` with a child `<Button size="sm">`. The source Close is a plain `DialogPrimitive.Close` with crossed rules, so no Button is used. |
| `DialogContent showCloseButton`; `DialogFooter showCloseButton` | ui/dialog.tsx (no consumer passed them) | Removed (source API): Close always rendered; Footer is a plain container |
| `data-slot="dialog-trigger"`, `"dialog-portal"`, `"dialog-close"` (wrapper-added) | ui/dialog.tsx; test selectors | Not present in the source. Tests use role or `button[aria-haspopup="dialog"]` (§11.3). |
| `PortalContainerContext` value `RefObject<HTMLElement \| null>`; `usePortalContainer()` → `RefObject \| undefined` | ui/portal-container.tsx; token-scope.tsx (`useRef`) | Context value `{ element: HTMLElement \| null }`. `usePortalContainer()` → that value, or `null` outside any scope. TokenScope keeps the element with `useState` and a callback ref. |
| `XIcon` from lucide-react in the Close | ui/dialog.tsx | The source crossed 2px rules |
| `aria-modal="true"` passed by ShellHelpDialog | shell-help-dialog.tsx | Kept: Radix Content spreads it onto the `role="dialog"` element. The comment is updated. |
| Comments naming Base UI | shell-help-dialog.tsx, portal-container.tsx, overlay.spec.ts | Updated |
| `data-base-ui-focus-guard` tolerance in `expectContained` | tests/browser/overlay.spec.ts | `data-radix-focus-guard`. Same guarantee: the end state after two frames is inside the dialog; nothing else is allowed. |
| B-DEPS `DEPS_ADDITIONS` `@base-ui/react\|1.8.0`; the C5 control on `@base-ui/react`; the comment naming `@base-ui/react` | infra/validation/lib.sh, run.sh | §11.5 |
| B-WEB-COMPONENTS `@base-ui/react/button` and `/dialog` import assertion; E1–E7 | infra/validation/run.sh | §11.5 |
| `components.json` `"style": "base-nova"` | web/components.json | `"radix-nova"` (Q5), so a future CLI addition follows ADR-001. No CLI is run. |
| WEB-SHELL.md "shadcn/ui on Base UI" | docs | Rewritten. Historical G002/G047 statements are kept as history. |

**Not edited:**
- the preserved skills (`skills/project-archetype/**`, `skills/vestive-design-skill/**`);
- historical artifacts;
- `docs/foundation/FOLLOW-UP.md`, which records the historical G001 checklist.

These mentions of Base UI remain historical.

**Public contracts.** No public HTTP, MCP or token-override contract changes. Internal component APIs change only as listed above, and every call site is migrated in the same change.

### 5.4 Scoped portal contract and lifecycle

**`TokenScope`:**

```tsx
const [container, setContainer] = useState<HTMLDivElement | null>(null)
const value = useMemo(() => ({ element: container }), [container])
<div data-portal-container="" ref={setContainer} />
```

The scope props and markup are unchanged, so the SSR HTML is identical.

**`DialogPortal` adapter** (`ui/dialog.tsx`):
- With an explicit `container` prop → `DialogPrimitive.Portal container={container}`.
- Else, inside a scope:
  - while the scope element is `null`, render **nothing**;
  - once it is set, render with `container={scope.element}`.
- Else, no provider → Radix default (body), as G002 documented.

**Guarantees:**
1. **Never into body inside a scope.** Radix `Portal` falls back to `document.body` only when `container` is falsy, and the adapter never passes a falsy container inside a scope.
2. **SSR returns null.** A closed dialog renders no portal; Radix Presence is false. The adapter returns null on the server, so there is no hydration mismatch (S-HYDRATION-DEV).
3. **Open at mount (`defaultOpen` or controlled).** The first client render returns null. After the commit, TokenScope's callback ref sets the element, and the next render portals into the scope. There is no transient root-token paint.
4. **Overlay and content portal separately into the same container.** Radix maps each `DialogPortal` child through its own Presence and Portal.
5. **Unmount.** Closing unmounts through Presence and focus returns to the trigger (Radix `onCloseAutoFocus` → `triggerRef`). Unmounting the scope removes the container with it.
6. **Modal semantics:**
   - `role="dialog"`, `aria-labelledby`/`aria-describedby` from Title and Description, and `aria-modal="true"` (passed);
   - FocusScope `loop` and `trapped`;
   - `hideOthers(content)` sets `aria-hidden` on every sibling up the ancestor chain, including the other scopes' content and the parent scope sections;
   - `pointer-events: none` on body while open;
   - RemoveScroll scroll lock.

   Axe's `aria-hidden-focus` rule passes when a modal is open (its `focusable-not-tabbable` check returns true while `isModalOpen()`), so S-A11Y with the dialog open stays at zero violations. This is verified, not assumed.

**Verification:**
- S-PORTAL: kept. Popup and overlay `closest('[data-token-scope]')` equal the scope, across live appearance changes, for every scope including the nested one.
- **S-PORTAL-INIT** (new; tokens-light and tokens-dark): dialogs open at mount (`/token-fixtures?open=A` and `?open=dark-only-in-A`). They are verified from insertion history, the first frame and live appearance changes (§5.4.1).
- S-OVERLAY and S-TEXT-SPACING: inertness or aria-hidden of background regions (the existing accepted predicate `closest('[inert], [aria-hidden=true]')`).

#### 5.4.1 S-PORTAL-INIT: insertion history, first frame and live mode (r2)

**Why history, not the current parent.**
- A MutationObserver callback runs at a microtask checkpoint, after the task that changed the DOM.
- Reading a node's `parentNode` or `closest()` there gives its parent at delivery. Suppose a node is inserted into `body` and moved into the scope container in the same task. Its delivery-time parent is then the container, so the transient body insertion vanishes from the evidence. That was the gap in r1's wording.
- Each `MutationRecord` keeps its own `target`, the parent at the moment of that mutation. r2 builds the evidence from records only.

**Recorder** (in `tokens.spec.ts`; no new file):
- **Installation.** Installed with `page.addInitScript` before navigation, so it precedes every page script and the parser's insertions. It observes `document` with `{ childList: true, subtree: true }`.
- **Selector.** `PORTAL = '[data-slot="dialog-overlay"], [data-slot="dialog-content"]'`.
- **R-a, added nodes.** For each record and each added element `n`, the portal elements are:
  - `n` itself, if it matches `PORTAL`;
  - plus `n.querySelectorAll(PORTAL)`, the added subtree.

  Each one is logged as `{ seq, kind: "added", target: record.target, node }`.
- **R-b, removed nodes.** For each removed element that itself matches `PORTAL`, the recorder logs `{ seq, kind: "removed", target: record.target, node }`. This catches a portal node taken out of a subtree that was inserted elsewhere, when the extraction happens before delivery: the extraction is a removal whose target is the old parent.
- **First-frame samples.** For each newly logged real (non-control) node, the callback synchronously samples three values:
  - the content's `background-color`;
  - the overlay's `background-color`, alpha included;
  - the title's `color`.

  The callback runs before the rendering step that first paints the node, so these are the first painted frame's values. Per-frame `requestAnimationFrame` samples of the same three values continue until the test stops the probe, as in the first-paint probe.
- **Verdict.** Entries hold live references in page memory, and the verdict is computed in the page by **identity**. Only a serialized summary leaves the page:
  - the kind;
  - the target class: `body`, `expected-container`, `other-container` or `other`;
  - the target's scope label;
  - the node's slot;
  - the control ID.

**Completeness.** Suppose a portal element is ever attached under a parent other than the expected container. Then some record has that parent as its target, in one of two ways:
- the element is added there, directly or inside an added subtree that still contains it at delivery (R-a);
- or it leaves such a subtree before delivery, and the removal's target is that parent (R-b).

Portal identity is the static `data-slot`, so attribute mutations are not involved.

**Cases.** Each case runs in tokens-light and tokens-dark, so in both appearances:
- `?open=A`: an overridden top-level scope;
- `?open=dark-only-in-A`: a nested, partially overridden scope.

1. **Navigation readiness.**
   - `goto` with `waitUntil: "load"`.
   - Then `[data-slot=dialog-content]` must be visible inside the expected scope within the expect timeout.
   - `waitForHydration` is not used before the assertions, because the dialog is already open and modal.
2. **SSR.** The navigation response body contains no `data-slot="dialog-content"` and no `data-slot="dialog-overlay"`.
3. **Scoped insertion.**
   - The real entries hold exactly one `added` entry for the overlay and one for the content.
   - Each target is **identical** to the expected scope's own container, `[data-token-scope="<label>"] > [data-portal-container]`.
   - No real entry has a `body`, `other-container` or `other` target.
4. **First frame.** Every delivery and frame sample equals the scope's expected values:
   - popover from `expectedSource`;
   - title ink from `expectedRole` (§6.2), within 2 per channel;
   - the scrim against `SOURCES[mode].overlay` within 1, alpha included.

   A precondition, from the independent tables, makes the check discriminating: the expected title ink differs from the root ink by more than the tolerance in all four case and appearance combinations. The distances are 33 (A, Light), 36 (A, Dark), 33 (dark-only-in-A, Light, inherited from A) and 25 (dark-only-in-A, Dark). A root-token first frame therefore cannot pass.
5. **Live mode.** The S-PORTAL loop runs on the initially open dialog: `mode`, then `other`, then `mode`, through `emulateMedia`.
   - In each appearance, the popup surface pixel, title ink, overlay and focus ring equal the scope's expected values.
   - The log gains no new `added` entry, so the dialog is not re-portalled.
6. **Modal and close.**
   - Focus is inside the content, and Tab stays inside.
   - Escape closes, and the two `removed` entries target the same container.
7. **Controls.** These run after step 6, one `page.evaluate` per control, so each control's mutations happen in a single task.
   - Every control node carries `data-g048-control="<id>"`, which excludes it from the real verdict.
   - All control nodes are removed in `finally`.

| ID | Same-task mutation | Must be detected as |
|---|---|---|
| PI-C1 (r1, kept) | `body.appendChild(overlayNode)` | an `added` entry with target `body` |
| PI-C2 (r2, required) | `body.appendChild(overlayNode)`, then `container.appendChild(overlayNode)` | An `added` entry with target `body`, even though the node's parent at delivery is the expected container. The test asserts that delivery-time parent, which proves that a current-parent recorder would have missed the insertion. The later container entry is also logged. |
| PI-C3 (r2) | `body.appendChild(wrapper)` where `wrapper` contains a content node, then `container.appendChild(wrapper)` | an `added` entry with target `body`, found through the added subtree (R-a) |
| PI-C4 (r2) | `body.appendChild(wrapper)` with a content node inside, then `container.appendChild(contentNode)` | A `removed` entry whose target is `wrapper`, class `other` (R-b). The wrapper's own body record holds no portal node at delivery. |

A control passes only when its exact expected entries appear. An undetected control fails S-PORTAL-INIT.

**Adapter contract unchanged.**
- The recorder and its controls test the evidence, not the adapter.
- The adapter is exactly as specified above:
  - it renders nothing until the scope element exists;
  - it never passes a falsy container inside a scope.
- r2 adds no body fallback, retry or relocation step.
- B-WEB-COMPONENTS still requires `dialog.tsx` to use `usePortalContainer` (§11.5).

**Evidence.** One `portal-init.json` per project and case, holding:
- the serialized entries;
- the first-frame and per-frame samples;
- the live-mode values;
- each control's outcome, including PI-C2's delivery-time parent.

**Execution-time validation.**
- S-PORTAL-INIT passes in tokens-light and tokens-dark on both the root and disabled variants (§11.7).
- Its attachments show all four controls detected.
- The real entries all target the expected container, and the first-frame and live-mode values match.
- The Result reviews the adapter source against the contract above.

### 5.5 Radix compatibility notes (adopted versions)

- **`"use client"` boundaries.**
  - `@radix-ui/react-dialog` 1.1.23's dist starts with `"use client"`.
  - `@radix-ui/react-slot` 1.3.3's dist does **not**, and `Slot` calls `useComposedRefs` (a hook). So `asChild` is used only inside client components: the shell help dialog and the fixture's client dialog.
  - `button.tsx` stays directive-free like the source, so server components can render `asChild={false}` Buttons. If the production build or dev hydration shows otherwise, adding `"use client"` to `button.tsx` is the pre-declared compatibility adaptation (category: Radix compatibility). It is recorded in the Result and costs a cycle only if it is a corrective fix.
- **React 19.3.** `forwardRef` remains supported, and Slot 1.2.5+ handles React 19 element refs. Typecheck covers the types.
- **`data-state` replaces `data-open`/`data-closed`.** The ejected `data-open`/`data-closed` custom variants already match `[data-state=open|closed]`. No dialog animation is used (source; G047 K17 superseded).
- **Body side effects while open:**
  - focus guards (`span[data-radix-focus-guard]` at the body edges);
  - `body[data-scroll-locked]` and an injected `<style>` from react-style-singleton;
  - `pointer-events: none`.

  Effect on tests:
  - S-HTML-ATTRS checks `<html>` only, so it is unaffected.
  - S-HEAD-ORDER inspects the served HTML only, so it is unaffected.
  - Tests that click outside the dialog must use Escape or Close. No current test clicks outside.

---

## 6. Tokens, roles, appearance and scopes (R4)

### 6.1 Palette and structure changes (`web/app/globals.css`, foundation section only)

The ejected block and `GLOBALS_EJECT_SHA` are untouched.

| Change | Light | Dark | Note |
|---|---|---|---|
| `--light/--dark-muted-foreground` | `#615b55` | `#8f8a84` | The kit's own slot (vx-muted). Was the text-3 values (G047 K5). Q6. |
| New internal role `ink` | `#0d0d0d` | `#ffffff` | vx-ink: headings, dialog title, hover ink |
| New internal role `text-2` | `#36322e` | `#c9c5c0` | vx-text-2: lead and bullet items |
| New internal role `text-3` | `#4d4843` | `#a8a39d` | vx-text-3: body and dialog description |
| `--angle` | the source polygon | — | T12 |
| `@theme` `--font-display` / `--font-label` / `--font-sans` / `--font-mono` | `"Oswald", "Arial Narrow", sans-serif` / `"Barlow Condensed", "Arial Narrow", sans-serif` / `"Roboto", system-ui, sans-serif` / `ui-monospace, "SFMono-Regular", Menlo, monospace` | — | Source stacks (`tokens.css`) |
| `@theme` `--spacing-section` | `88px` | — | T10 |
| `@theme` `--spacing-cell` (new) | `36px` | — | T11 |
| `@utility angled` | removed | — | Replaced by X7 per-variant layers |
| `@font-face` block (new, delimited) | §7.2 | — | |
| Blueprint layer block (new) | §8.4 | — | |

**Contrast of the new roles** (planning computation; the rendered checks are authoritative):

| Role | Light | Dark |
|---|---|---|
| ink on ground | ≥ 18:1 | ≥ 18:1 |
| text-2 | ≥ 10.4:1 | ≥ 10.8:1 |
| text-3 | ≥ 7.4:1 | ≥ 7.4:1 |
| muted-foreground (vx-muted) | ≥ 5.4:1 | ≥ 5.4:1 (5.43 on the popover) |

The three grounds are bg, surface and raised.

### 6.2 Internal role resolution and override propagation

**Declaration.** The roles are declared exactly like the 16 contract names:
- `--light-<role>` and `--dark-<role>` palette sources on `:root`;
- `--<role>: var(--light-<role>)`, with the dark equivalent under `@variant dark`, in the `:root, [data-token-scope]` resolution block;
- `--color-<role>: var(--<role>)` in `@theme inline`.

They therefore resolve per appearance at the root and at every scope. None is root-only.

**`web/lib/tokens.ts`:**
- `export const INTERNAL_ROLE_TOKENS = ["ink", "text-2", "text-3"] as const`;
- `SEMANTIC_COLOR_TOKENS` (16) and the `TokenOverrides` type are unchanged, so the override **input** contract is unchanged;
- `tokenOverrideStyle()` also emits, for each appearance `m` in which an override sets `foreground`, `--m-ink`, `--m-text-2` and `--m-text-3` equal to that value.

**Effects:**
- An Organization-style foreground override applies to every text-ink role of its scope, in that appearance only.
- A scope without a foreground override inherits its parent's role sources, through CSS custom-property inheritance from the parent's inline style.
- Light-only and Dark-only overrides stay partial.

These are internal visual roles, **not** an Organization branding API or schema.

**Oracle.** S-TOKENS/S-RUNTIME (§11.3):
- `expectedRole(label, mode, role)` walks scope → parent, taking an own `foreground` override for that mode; otherwise it uses the independent platform role source in `support/modes.ts`;
- a runtime change of a role source (`--light-ink` set on A) applies immediately to A and its nested scope only, in that appearance only;
- the first-paint tests use the same independent source for H1: `PALETTE[m].ink` comes from `ROLE_SOURCES`, while `PALETTE[m].foreground` stays the semantic value (§11.3a).

**Documented behaviour.** Derivation happens where overrides enter: `TokenScope` → `tokenOverrideStyle`. A changed override passed through `TokenScope` therefore re-derives the roles. Editing a single source variable directly changes exactly that variable, as it does today. S-RUNTIME asserts both behaviours explicitly.

### 6.3 Appearance unchanged

The following are unchanged:
- Light, Dark and System, with System as the default;
- localStorage-only persistence and the invalid/unavailable-storage fallback;
- CSS-only live OS following;
- the classic inline `appearance-init` in `<head>` (hash `d5054b00…e143e42` unchanged);
- `suppressHydrationWarning` only on `<html>`;
- the dark variant.

No `.dark`/`.light`/`[data-theme]` selectors, `next-themes` or new preference store are added. `layout.tsx` is unchanged: no font preload is added and the head order is unchanged.

---

## 7. Bundled fonts (R3)

### 7.1 Pinned artifacts

**Upstream.** `google/fonts` at `7085eb89a950e85db5b166b7a58d414544b4140c`, main HEAD observed 2026-10-06. None of the three family directories changed after its latest family commit:
- `oswald` `8b0a1d0f…`;
- `barlowcondensed` `60824dce…`;
- `roboto` `1c627bfa…`.

**URL pattern.** `https://raw.githubusercontent.com/google/fonts/7085eb89a950e85db5b166b7a58d414544b4140c/<path>`, with `[`, `]` and `,` percent-encoded in requests.

**Rules:**
- Original bytes are copied unmodified, so delivered SHA-256 = original SHA-256.
- At intake, each file's `blob` SHA-1 (`printf 'blob %d\0' $size | cat - f | sha1sum`) and size must equal the values below before the SHA-256 is computed and pinned (in `lib.sh` and `support/assets.ts`).
- Any mismatch stops the work, with no cycle consumed.

| Upstream path | Size | Git blob SHA-1 | Delivered path |
|---|---|---|---|
| `ofl/oswald/Oswald[wght].ttf` | 172088 | `d1a3b9cb1325bf20b3a06ef9849d21c411212a3b` | `web/public/fonts/oswald/oswald-wght.ttf` |
| `ofl/oswald/OFL.txt` | 4483 | `4d4e319b6114be0174e6de5d8090408f78dc65b8` | `web/public/fonts/oswald/OFL.txt` |
| `ofl/barlowcondensed/BarlowCondensed-Regular.ttf` | 102480 | `ecbe0cc84118decc6a16216ad53d83d39888b175` | `web/public/fonts/barlow-condensed/barlow-condensed-400.ttf` |
| `ofl/barlowcondensed/BarlowCondensed-SemiBold.ttf` | 109428 | `9df66576492bee4bd0559066825472808fa98adc` | `web/public/fonts/barlow-condensed/barlow-condensed-600.ttf` |
| `ofl/barlowcondensed/OFL.txt` | 4377 | `2f22ba6ca4dbc12145ae40b14c3368f387cba0e2` | `web/public/fonts/barlow-condensed/OFL.txt` |
| `ofl/roboto/Roboto[wdth,wght].ttf` | 488584 | `5522a368d9072fd88c299916e61fcff369949061` | `web/public/fonts/roboto/roboto-wdth-wght.ttf` |
| `ofl/roboto/OFL.txt` | 4394 | `9c48e05a24fa2267e207ecf17560291b5621c197` | `web/public/fonts/roboto/OFL.txt` |

- **Font payload:** 872,580 bytes (four binaries).
- **Metadata** (recorded, not shipped): the `METADATA.pb` blobs `ed3535d8…` (Oswald), `16a06d65…` (Barlow Condensed) and `9ae795d6…` (Roboto).

**Upstream lineage** (identical blobs):
- Oswald: `googlefonts/OswaldFont` `fonts/variable/Oswald[wght].ttf` at `89795261`, packaged as v4.103.
- Barlow: `jpt/barlow` v1.408 (`b4726ddf`).
- Roboto: `googlefonts/roboto-3-classic` v3.015 (`91d5d3e5`, archive `Roboto_v3.015.zip`). Roboto moved from Apache to OFL in google/fonts (commits `113bac09`, `06afa5c5`), and the shipped `ofl/roboto` files are OFL-1.1.

**Coverage and axes** (METADATA):

| Family | Axes or weights | Subsets |
|---|---|---|
| Oswald | `wght` 200–700 | latin, latin-ext, cyrillic(-ext), vietnamese |
| Barlow Condensed | static 400 and 600 | latin, latin-ext, vietnamese |
| Roboto | `wdth` 75–100, `wght` 100–900 | latin, latin-ext, cyrillic(-ext), greek(-ext), vietnamese, math, symbols |

**Weights exercised:**
- Oswald 400 (nav, skip link), 500 (button, H3), 600 (H2, dialog title, header identity) and 700 (H1);
- Barlow 400 (meta, bullets) and 600 (kicker, appearance labels);
- Roboto 400 in production, plus 500 and 700 in fixture probes.

Every required weight is inside a declared face range. The mono role stays system mono.

**Why TTF.** There is no official Roboto WOFF2. Converting or subsetting would add a tool and modified-version notices, so the unmodified upstream TTFs are shipped. None of the three declares a Reserved Font Name.

### 7.2 Delivery

**`@font-face` block.** A delimited block in `globals.css` (`/* F. Bundled fonts (G048): same-origin, unmodified upstream files; see WEB-SHELL.md */`):

```css
@font-face { font-family: "Oswald"; src: url("/fonts/oswald/oswald-wght.ttf") format("truetype"); font-weight: 200 700; font-style: normal; font-display: swap; }
@font-face { font-family: "Barlow Condensed"; src: url("/fonts/barlow-condensed/barlow-condensed-400.ttf") format("truetype"); font-weight: 400; font-style: normal; font-display: swap; }
@font-face { font-family: "Barlow Condensed"; src: url("/fonts/barlow-condensed/barlow-condensed-600.ttf") format("truetype"); font-weight: 600; font-style: normal; font-display: swap; }
@font-face { font-family: "Roboto"; src: url("/fonts/roboto/roboto-wdth-wght.ttf") format("truetype"); font-weight: 100 900; font-stretch: 75% 100%; font-style: normal; font-display: swap; }
```

- Sources are `url()` only, never `local()`, so a font installed on the host or in the browser image can never satisfy a face. A web font of the same family name also shadows any installed font.
- **Serving.** Next serves `web/public/**` same-origin. The standalone image needs `public/` copied:
  - `infra/docker/web.Dockerfile` runtime stage: `COPY --from=build --chown=node:node /app/public ./public`;
  - `infra/docker/web-browser.Dockerfile`: `COPY --from=fixture /app/public /fixture/public`;
  - the `next dev` server in the browser stage reads `/app/public` from the copied source.

  The production web, the fixture builds and dev all serve the same bytes.
- **No external fetch.** Nothing fetches fonts at build time: no font package, no `next/font`, no postinstall, no CDN.
- **First paint.** `font-display: swap` keeps text visible with the fallback. FCP and the first-paint palette proof do not wait for fonts, which is consistent with the Spec. Screenshots wait for fonts through the bounded phase-B readiness gate (§7.4 B0), which includes `document.fonts.ready`.
- **Rendering consistency.** The browser image's fontconfig alias (Liberation for the generic families) stays. It now affects only the fallback and the system mono.

### 7.3 Licences and notices

- **Font licences.** Each family's upstream `OFL.txt` ships byte-exact beside its font, and is served at `/fonts/<family>/OFL.txt`. The copyright lines are:
  - "Copyright 2016 The Oswald Project Authors (https://github.com/googlefonts/OswaldFont)";
  - "Copyright 2017 The Barlow Project Authors (https://github.com/jpt/barlow)";
  - "Copyright 2011 The Roboto Project Authors (https://github.com/googlefonts/roboto-classic)".

  They are reproduced, not relabelled, in `web/THIRD_PARTY_NOTICES.md` and WEB-SHELL.md.
- **OFL conditions honoured:**
  - each copy accompanies the notice;
  - no sale by itself;
  - unmodified, so no Reserved Font Name issue;
  - distributed under the OFL only.

### 7.4 Actual-font oracle (S-FONTS): bounded pre-load checks and the positive loaded-font oracle

Implemented in `web/tests/browser/support/fonts.ts`, and used by the shell projects on production nodes and by the tokens projects on fixture probes.

r1 mixed two things: the full oracle waited for `document.fonts.ready` and for successful font responses, while its fallback control held every font request indefinitely. r2 separates them:
- **Phase A** reads what is rendered now and never waits for a font to load. It is the only phase used while fonts are held.
- **Phase B** is the positive oracle. It runs only on normal delivery and starts with an explicit, bounded readiness gate.

A held request never reaches a phase-B wait.

**Timeouts.** These are constants in `fonts.ts`. Each expiry fails the test with a named message; it is never a finding.

| Wait | Bound |
|---|---|
| Held-request capture (the four pinned paths observed and held) | 10 s poll |
| `FontFace` status poll | 10 s |
| One CDP call (`DOM.*`, `CSS.getPlatformFontsForNode`) | 5 s, by `Promise.race` |
| One font response completing (`response.finished()`) | 15 s |
| `document.fonts.ready` | 15 s. It is raced inside the page against a timer, and the result must be `"ready"`; `"timeout"` fails. |
| Hydration | the existing `waitForHydration` bound (60 s) |

Each bound is well below the 120 s Playwright test timeout (`web/playwright.config.ts`), and the response waits run concurrently.
- A single hang therefore surfaces as a named step failure.
- If slow steps together exceeded 120 s, the global timeout would fail the test too.
- No timeout of either kind can pass a test.

**Phase A — bounded pre-load checks.** These use the DOM and CDP only, with no screenshots: Playwright's screenshot path can wait for web fonts, so a screenshot taken while fonts are held could hang.
- **A1 Rendered fonts.** `renderedFonts(page, targets)` calls `CSS.getPlatformFontsForNode` for each target and returns one verdict per node:
  - `custom`: exactly one platform font, with `isCustomFont: true`, the family equal to the binary's own name, and `glyphCount > 0`;
  - `fallback`: no custom font;
  - `mixed`: a custom font plus another.

  It returns verdicts and does not throw for a fallback. Detection is therefore always an asserted value, never an exception.
- **A2 Face status.** `faceStatuses(page)` reads `document.fonts` immediately: family, weight range and `status` for the four bundled faces.
- **A3 Geometry.** DOM-only reflow and overlap checks, shared with S-RESPONSIVE.

**Phase B — positive loaded-font oracle.** It runs on normal delivery only. The S-FONTS positive tests use it, and S-FIDELITY, S-STATES, S-TEXT-SPACING's precondition and S-SCREENSHOTS use its readiness gate B0.
- **B0 Readiness.**
  - `goto` with the default `load`, then `waitForHydration`.
  - Exactly the four pinned `/fonts/` responses are recorded, each `finished()` within its bound.
  - `document.fonts.ready` resolves within its bound, and all four bundled faces are `loaded`.
  - Any expiry fails the test.
- Steps 1–5 below are r1's positive checks, unchanged in content, and run only after B0.

1. **Bytes.**
   - Record every response whose path starts with `/fonts/`: same origin as the page, status 200, a font content type, and a SHA-256 of `response.body()` equal to the independent pin in `support/assets.ts`.
   - Exactly the four binaries are requested on `/`, and no request goes to any other origin (S-BOUNDARY, unchanged).
2. **Binary metadata.** From those same bytes, parse:
   - `name` (IDs 1, 2, 4, 6, 16);
   - `fvar` (axes and ranges);
   - `OS/2.usWeightClass`.

   The expected family is the binary's own name, not the CSS alias. Variable ranges must include every required weight, and the static faces' weight class must equal 400 and 600.
3. **Loaded faces.** After B0's bounded `document.fonts.ready`, each of the four `FontFace` entries has status `loaded`.
4. **Rendered fonts.** A1 (CDP `CSS.getPlatformFontsForNode`) on the representative nodes must give the verdict `custom`: exactly one font, with `isCustomFont: true`, the family equal to the binary's name, and `glyphCount > 0`. A node rendered partly in a fallback (`mixed`) fails.
   - **Production nodes:**
     - H1 (700);
     - header identity (600);
     - nav link (400);
     - help trigger (500);
     - H2 (600);
     - H3 (500);
     - dialog title (600);
     - focused skip link (400);
     - kicker (600);
     - appearance label (600);
     - About bullet (400);
     - footer meta (400);
     - lead, cell body and dialog description (Roboto 400).
   - **Index:** expected `fallback`, not custom (system mono role).
5. **Weights and synthesis.**
   - The computed `font-weight` equals §4.7.
   - **Synthesis twin check:** take an element screenshot, set `font-synthesis: none` on that element, and take another; they must be byte-identical. If they differ, synthetic bold or oblique was in use.

**Controls and the swap case.** Each is its own test in shell-light and shell-dark, so its requests and S-ERRORS stay attributable. They therefore run on the root and disabled variants.

6. **Negative control: held fonts → fallback detected, readable and operable** (phase A only; r2).
   - **Hold.** Before navigation, `holdFonts(page)` routes same-origin requests whose path starts with `/fonts/` and holds each one. It neither fulfils nor aborts them: an aborted or failed font load would log a console error under S-ERRORS.
   - **Structure.** Everything after the hold runs in `try { … } finally { hold.release() }`.
   - **Readiness without the load event.** Held font requests can delay `load`, so the test navigates with `waitUntil: "domcontentloaded"`. Three bounded checks then prove that the control is active:
     - the held set equals the four pinned paths;
     - hydration completes through `waitForHydration`;
     - all four faces are `loading`.

     If the build had rewritten the font URLs away from `/fonts/`, this capture would fail closed (§16 risk d).
   - **Detection.** A1 on H1, kicker and lead must return exactly `[{H1: fallback}, {kicker: fallback}, {lead: fallback}]`. These are the intended non-custom-font findings. Phase B is never invoked.
   - **Readability.** Each of the three has:
     - a non-empty rendered text box;
     - `visibility: visible`;
     - a computed colour equal to its §4.7 role: ink, accent and text-2.

     `font-display: swap` gives no invisible period.
   - **Keyboard.**
     - From the document start (`resetFocus`), Tab reaches the help trigger within at most 20 presses.
     - Enter opens the dialog, and focus is inside.
     - Tab and Shift+Tab stay inside.
     - Escape closes, and focus returns to the trigger.
   - **Reflow.** At 1280×800 and at 320×640 (`setViewportSize`):
     - there is no horizontal overflow (`scrollWidth ≤ clientWidth`);
     - header controls do not overlap one another;
     - at 320, the opened dialog lies inside the viewport and Close is reachable.
   - **Still held.** At the end, all four faces are still `loading`, the held set is unchanged and no `/fonts/` response has completed. Every check above therefore ran under fallback, not after a late load.
   - **Steps recorded.** The attachment records the ordered list of completed steps. The test passes only when it equals the expected list. A timeout throws before the list is complete, so it can never pass the control.
   - **`finally`.** `release()` lets the route pass requests through, calls `route.continue()` for each held request and unroutes. It is idempotent.
     - A release error is recorded, and it fails a test whose body otherwise passed.
     - It never replaces the body's own failure.
7. **Negative control: synthesis.** As in r1: an injected test-only face `"Synthesis Probe"` uses only `barlow-condensed-400.ttf`.
   - It is loaded with `FontFace.load()` within the response bound, after B0.
   - An element in it at weight 700 must be **detected** by check 5.
8. **Swap stability** (r2 structure), at 1280×800 and at 320×640:
   - **Hold and readiness.** `holdFonts` before navigation, then control 6's phase-A readiness: `domcontentloaded`, four requests held, hydration, faces `loading`.
   - **Before release.**
     - Open the dialog and focus Close.
     - A1 on the dialog title must be `fallback`, which proves a swap really happens during the test.
     - Record the state: dialog open, Close focused, dialog inside the viewport, header controls not overlapping.
   - **Release.** `hold.release()` runs inside `try`, followed by B0's bounded readiness:
     - the four responses finish with status 200 and matching hashes;
     - `fonts.ready` returns `"ready"`;
     - the faces are `loaded`.
   - **After release.**
     - A1 on the dialog title must now be `custom` Oswald.
     - The dialog stays open, and `document.activeElement` is still Close.
     - The dialog is still inside the viewport, header controls do not overlap, and there is no horizontal overflow.
   - **`finally`.** Releases anything still held and unroutes.

**Evidence.** `fonts.json` holds:
- the requests, hashes, name and `fvar` data;
- phase-A and phase-B verdicts per node, and the twin results;
- the readiness timings;
- for each control and swap case: its step list, held paths, status snapshots and release record.

**Execution-time validation.**
- Normal delivery passes phase B:
  - in shell-light and shell-dark, on the root and disabled variants;
  - on the fixture probes in tokens-light and tokens-dark.
- Controls 6 and 7 and both swap cases complete as `expected` tests, with full step lists and per-step durations inside their bounds.
- Control 6 reports exactly the three `fallback` findings.

---

## 8. Blueprint imagery (R5)

### 8.1 Composition (original, hand-authored)

**File.** `web/public/blueprint/overview.svg`.

**Root.** `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 960 640" fill="none" stroke="#000" stroke-linecap="square">`.
- No XML declaration, DOCTYPE, comments, `<style>`, `<defs>`, `<use>`, `href`, `<text>`, `<image>`, raster, filter, animation or script.
- No internal references either.
- Target size at most 16 KiB; hard limit 40 KiB.

**Three motif groups:**
- `<g id="construction" stroke-opacity=".35" stroke-width="1">`:
  - a 40-unit construction grid as compact multi-segment paths;
  - three extended guide lines crossing the whole field, dashed `8 6`.
- `<g id="dimensions" stroke-opacity=".7" stroke-width="1.25">`:
  - four dimension lines (two horizontal, two vertical) with extension lines and 45° architectural ticks at both ends;
  - one chained dimension with intermediate ticks.
- `<g id="structure" stroke-opacity="1" stroke-width="2">`:
  - an unfinished floor plan: double-line outer walls with open ends;
  - interior partitions;
  - door openings with quarter-circle swing arcs;
  - window marks (three thin parallels);
  - column markers (r=10) at grid intersections;
  - a partial axonometric volume with dashed hidden edges, conveying a beginning.

**Allowed vocabulary:**
- **Elements:** `svg`, `g`, `path`, `line`, `polyline`, `polygon`, `rect`, `circle`.
- **Attributes:** `xmlns`, `viewBox`, `fill`, `stroke`, `stroke-width`, `stroke-opacity`, `stroke-dasharray`, `stroke-linecap`, `stroke-linejoin`, `id` (exactly those three values), `d`, `x1`, `y1`, `x2`, `y2`, `points`, `x`, `y`, `width`, `height`, `cx`, `cy`, `r`, `transform`.

There are no labels and nothing that claims a feature, product, entity map, logo or customer data.

### 8.2 Safety, enforced statically (B-WEB-DESIGN, §11.5)

**Parse.** Well-formed XML through the harness yq image (`-p xml`).
- **Root:** the root element is `svg`, `viewBox` is `0 0 960 640`, and `xmlns` is the SVG namespace.
- **Allow-lists:** every element and attribute is in the allow-lists above.
- **Raw-byte denial:** `<!`, `<?`, `href`, `url(`, `script`, `style`, `on[a-z]+=`, `foreignObject`, `<image`, `<use`, `<animate`, `<set`, `@import`, `data:`.

**Size.** At most 40960 bytes.

**Motifs.** Exactly the three groups, with at least 10, 8 and 10 drawable descendants.

**Pin.** The file SHA-256 is pinned at authoring time. Any later edit changes the pin, which is reported as drift.

### 8.3 Theme roles and opacity

- **Colour.** The layer's colour is `background-color: var(--muted-foreground)` with `opacity: .5`. The SVG supplies only alpha through `mask-image`, so the stroke `#000` value is irrelevant.
- **Effective stroke colour** = muted-foreground × (0.5 × group opacity) over the section ground. For structure lines:
  - Light ≈ `#aaa6a1` (≈ 2.2:1 against `#f4f2ee`);
  - Dark ≈ `#4e4b48` (≈ 2.2:1 against `#0d0d0d`).
- **Hierarchy.** Dimensions are at 0.7 and construction at 0.35 of that, so the treatment is low-emphasis but visible.
- **Scopes.** The layer uses the scope-resolved token, so overridden scopes change its colour without new inputs.
- **No preference.** There is no separate background preference or persistence.

### 8.4 Placement, masking and layout

**Component.** `web/components/layout/blueprint-backdrop.tsx` renders `<div aria-hidden="true" data-blueprint="" className={className} />`. It is empty, with no tabindex, role or text.

**Overview usage.** It is the first child of the overview `Section` content, with `containerClassName="isolate"`. The Container is already `relative`.

**CSS** (`globals.css` block `/* E. Decorative imagery (G048) */`):

```css
[data-blueprint] {
  position: absolute; inset: 0 0 0 auto; width: 46%; z-index: -1; pointer-events: none;
  background-color: var(--muted-foreground); opacity: .5;
  mask-image: url("/blueprint/overview.svg"), linear-gradient(to right, transparent, var(--foreground) 96px);
  mask-size: 720px auto, 100% 100%; mask-position: left center, 0 0;
  mask-repeat: no-repeat; mask-composite: intersect;
}
@media (width < 64rem) {
  [data-blueprint] { inset: auto 0 0 0; width: auto; height: var(--spacing-section);
    mask-image: url("/blueprint/overview.svg"); mask-size: 640px auto; mask-position: right top; }
}
@media (forced-colors: active) { [data-blueprint] { display: none; } }
```

- **1024px and above.** The right 46% of the container, at 0.75 scale with a 96px left fade.
  - The kicker, lead (`max-width: 52ch`, right edge ≈ 535px at 1280) and normal H1 lie left of 691px.
  - A long configured H1 may cross into it. That is measured locally (§8.5) for large-text contrast; the planning estimate is ≥ 8:1.
- **Below 1024px.** The 88px bottom padding band of the overview, at 0.667 scale, right-aligned. This is padding, so it is never behind text.
- **Paint order.** The `-z-10`-equivalent `z-index: -1` inside the isolated Container paints above the section ground and below the in-flow content.
- **Layout.** The layer is absolutely positioned and does not change document order, sizing or overflow.
- **Behaviour.** No animation, so reduced motion has nothing to stop. Under forced colours it is hidden. A missing or unloaded mask image paints nothing, leaving the same content and behaviour.

### 8.5 Verification (S-BLUEPRINT; shell projects = both appearances; tokens projects = scopes)

**At 1280×800 and 320×640:**
1. **Decorative semantics.**
   - The layer exists in `#overview`, has `aria-hidden="true"`, is not focusable (S-TAB order unchanged), has computed `pointer-events: none`, and runs no animations.
   - `elementFromPoint` at its centre is not the layer.
   - Overview child boxes and section height are identical with the layer `display: none`.
2. **Response.** The SVG is same-origin, status 200, `image/svg+xml`, with SHA-256 equal to the pin.
3. **Imagery on versus off.** Over the same clip: the layer box ∩ viewport, minus any text-node rectangles. With the configured default identity none intersect; text over imagery is covered by §11.4.
   - visible versus `visibility: hidden` via an injected test style;
   - pixels differing by more than 8 ≥ max(400, 0.5% of the area);
   - with the layer hidden, every pixel is within 2 of the section ground (this also proves no text is in the region);
   - with it visible, at least 200 pixels are ≥ 50% of the way toward the expected structure-stroke composite. That composite is computed from the independent `SOURCES[mode]` muted-foreground and background.
4. **Motif evidence.** Load `/blueprint/overview.svg` as a document and screenshot it whole. Hide each group in turn by its id and screenshot again; each group's removal must change at least 300 pixels. Both captures are attached.
5. **Missing asset.** A new page holds `/blueprint/` requests. The clip equals the ground, content boxes are unchanged and the help dialog works.
6. **Forced colours.** Under `emulateMedia({ forcedColors: "active" })` the layer computes to `display: none`, overview text is visible, and Tab to the trigger shows an outline.
7. **Scopes (fixture).** Each scope's 320×160 `BlueprintBackdrop` panel's structure-stroke composite equals that scope's resolved muted-foreground over its background. A runtime change of `--light-muted-foreground` on A recolours A and its nested scope only (S-RUNTIME).
8. **Screenshots.** Overview clips at desktop and 320, in both appearances; also in the full-page S-SCREENSHOTS.

**Local text contrast over imagery** (S-CONTRAST; §11.4):
- kicker, H1 and lead at 1280, 320 and 1280 with 200% text;
- the long-identity H1 (S-LONG-IDENTITY).

### 8.6 Provenance

The SVG is written by hand during implementation. Its geometry is original: no traced, imported or generated third-party drawing. Its SHA-256, size and group element counts are recorded in WEB-SHELL.md and the Result.

---

## 9. Files and exact changes

### 9.1 Inventory (exactly 51 paths; this equals the G048 allow-list; row 50 is PENDING PG-7; row 51 added in r2)

| # | Path | Action | Content |
|---|---|---|---|
| 1 | `web/components/ui/typography.tsx` | create | §4.4 |
| 2 | `web/components/ui/layout.tsx` | create | §4.4 |
| 3 | `web/components/ui/bullet-list.tsx` | create | §4.4 |
| 4 | `web/components/ui/site-header.tsx` | create | §4.4 |
| 5 | `web/components/ui/button.tsx` | replace | source-derived (§4.4) |
| 6 | `web/components/ui/dialog.tsx` | replace | source-derived plus the adapter (§4.4, §5.4) |
| 7 | `web/components/ui/portal-container.tsx` | modify | element-state contract (§5.4) |
| 8 | `web/components/layout/token-scope.tsx` | modify | callback-ref state (§5.4) |
| 9 | `web/components/layout/app-shell.tsx` | modify | TextLink skip link; SiteHeader (`brand` = the existing product-name `<p>`, links, `navLabel="Primary"`, action = appearance control plus help); footer Container and Text meta |
| 10 | `web/components/layout/shell-help-dialog.tsx` | modify | `DialogTrigger asChild` with `Button size="sm"`; BulletList tips; comment |
| 11 | `web/components/layout/blueprint-backdrop.tsx` | create | §8.4 |
| 12 | `web/app/page.tsx` | modify | Section, Kicker, Heading, Text, RuledGrid (cells: Text index, Heading 3 `mt-3.5 text-accent`, Text body `mt-3.5`), BulletList, BlueprintBackdrop. Same IDs, anchors, texts, `data-testid`s and heading order. |
| 13 | `web/app/globals.css` | modify | §6.1, §7.2, §8.4. The ejected block is untouched. |
| 14 | `web/lib/tokens.ts` | modify | §6.2 |
| 15 | `web/components.json` | modify | `style` → `radix-nova` (Q5) |
| 16 | `web/package.json` | modify | three dependency keys (§5.1) |
| 17 | `web/package-lock.json` | modify | §5.2 |
| 18 | `web/playwright.config.ts` | modify | webServer 4: the fixture build at `127.0.0.1:3004` with a 64-character `PRODUCT_NAME` (S-LONG-IDENTITY) |
| 19–25 | `web/public/fonts/oswald/{oswald-wght.ttf,OFL.txt}`, `web/public/fonts/barlow-condensed/{barlow-condensed-400.ttf,barlow-condensed-600.ttf,OFL.txt}`, `web/public/fonts/roboto/{roboto-wdth-wght.ttf,OFL.txt}` | create | §7.1, byte-exact |
| 26 | `web/public/blueprint/overview.svg` | create | §8.1 |
| 27 | `web/THIRD_PARTY_NOTICES.md` | create | §4.6, §7.3 |
| 28 | `web/tests/browser/support/fonts.ts` | create | §7.4 |
| 29 | `web/tests/browser/support/source-reference.ts` | create | §4.7 |
| 30 | `web/tests/browser/support/assets.ts` | create | the expected SHA-256 of the 4 fonts, 3 notices and SVG |
| 31 | `web/tests/browser/support/png.ts` | modify | adds `decodePng` (8-bit RGB/RGBA, filters 0–4) for region analysis |
| 32 | `web/tests/browser/support/contrast.ts` | modify | adds `localTextContrast` (§11.4) |
| 33 | `web/tests/browser/support/angle.ts` | modify | all four ring sides ≥ 3:1 for angled controls; `min()`/`max()` parsing only under D-ANGLE |
| 34 | `web/tests/browser/support/modes.ts` | modify | SOURCES muted-foreground; `ROLE_SOURCES` (ink, text-2, text-3); `PALETTE[m].ink` from `ROLE_SOURCES`, with `PALETTE[m].foreground` unchanged (§11.3a) |
| 35 | `web/tests/browser/support/spacing.ts` | modify | MODAL_ITEMS Close → `${POPUP} > button[aria-label="Close"]`, texts `[""]`, kind icon; precondition that the bundled faces are loaded (phase-B readiness, §7.4 B0) |
| 36 | `web/tests/browser/accessibility.spec.ts` | modify | §11.3 |
| 37 | `web/tests/browser/design.spec.ts` | modify | S-DESIGN, S-PALETTE, S-FIDELITY, S-FONTS, S-BLUEPRINT, S-ANGLE, S-SCREENSHOTS |
| 38 | `web/tests/browser/overlay.spec.ts` | modify | §11.3 |
| 39 | `web/tests/browser/shell.spec.ts` | modify | S-NAV, S-RESPONSIVE, S-LONG-IDENTITY |
| 40 | `web/tests/browser/tokens.spec.ts` | modify | S-TOKENS, S-PORTAL, S-PORTAL-INIT (with the §5.4.1 history recorder and PI-C1–PI-C4), S-RUNTIME, S-CONTRAST, S-ANGLE, S-STATES, S-FONTS, S-BLUEPRINT, S-TEXT-SPACING |
| 41 | `web/tests/fixtures/token-scopes/page.tsx` | modify | §11.2 fixture content |
| 42 | `infra/docker/web.Dockerfile` | modify | one `COPY` of `public` (rendered file → registered) |
| 43 | `infra/docker/web-browser.Dockerfile` | modify | one `COPY` of the fixture `public` |
| 44 | `infra/validation/lib.sh` | modify | §11.5 (B-DEPS lists and the removal rule; browser matrix) |
| 45 | `infra/validation/run.sh` | modify | §11.5 |
| 46 | `infra/validation/render-customizations.tsv` | modify | + the `infra/docker/web.Dockerfile` row; G048 reason suffixes on the package files, `globals.css` and `page.tsx` |
| 47 | `docs/foundation/WEB-SHELL.md` | modify | §11.10 |
| 48 | `management/v0/results/G048-scope-allowlist.txt` | create | the first execution write (§9.2) |
| 49 | `management/v0/results/G048-vestive-components-fonts-blueprints-result.md` | create | the Result |
| 50 | `infra/validation/required-checks.tsv` | modify | + `B-WEB-ASSET-CONTROLS` (PENDING PG-7 format) |
| 51 | `web/tests/browser/appearance.spec.ts` | modify | Added in r2 (§11.3a). S-FIRSTPAINT and S-FIRSTPAINT-DELAYED expect H1 = `PALETTE[m].ink` instead of `.foreground`, with the distinctness precondition and the expected values in the attachments. Nothing else changes. |

- **Count.** Rows 19–25 are seven paths, so the table lists exactly 51 paths:
  - 19 created (rows 1–4, 11, 19–30, 48, 49);
  - 32 modified, including the two in-place replacements (rows 5–6);
  - 0 deleted.

  If PG-7 shows that `required-checks.tsv` does not exist after G003, row 50 is dropped and the count is 50.
- **Unchanged.**
  - `web/app/layout.tsx`, `web/components/layout/appearance-control.tsx` and `web/lib/{appearance,utils,config,core-client}.ts`;
  - `web/tests/browser/boundary.spec.ts`, `support/{color,page-errors}.ts` and `fixtures/token-scopes/fixtures.ts`;
  - README.md, `skills/**`, `.archetype/**` and every other path.
- **Deletions.** None. The old button and dialog are replaced in place, and the `angled` utility is removed inside `globals.css`.

### 9.2 `management/v0/results/G048-scope-allowlist.txt` (exact content)

```
# G048 scope allow-list (input to infra/validation/run.sh B-SCOPE).
# Extended regular expressions, one exact anchored path per line. Mirrors G048 Plan §9.1.
^web/components/ui/(typography|layout|bullet-list|site-header|button|dialog|portal-container)\.tsx$
^web/components/layout/(token-scope|app-shell|shell-help-dialog|blueprint-backdrop)\.tsx$
^web/app/(page\.tsx|globals\.css)$
^web/lib/tokens\.ts$
^web/components\.json$
^web/package(-lock)?\.json$
^web/playwright\.config\.ts$
^web/public/fonts/oswald/(oswald-wght\.ttf|OFL\.txt)$
^web/public/fonts/barlow-condensed/(barlow-condensed-400\.ttf|barlow-condensed-600\.ttf|OFL\.txt)$
^web/public/fonts/roboto/(roboto-wdth-wght\.ttf|OFL\.txt)$
^web/public/blueprint/overview\.svg$
^web/THIRD_PARTY_NOTICES\.md$
^web/tests/browser/support/(fonts|source-reference|assets|png|contrast|angle|modes|spacing)\.ts$
^web/tests/browser/(accessibility|appearance|design|overlay|shell|tokens)\.spec\.ts$
^web/tests/fixtures/token-scopes/page\.tsx$
^infra/docker/web(-browser)?\.Dockerfile$
^infra/validation/(lib\.sh|run\.sh|render-customizations\.tsv|required-checks\.tsv)$
^docs/foundation/WEB-SHELL\.md$
^management/v0/results/G048-(scope-allowlist\.txt|vestive-components-fonts-blueprints-result\.md)$
```

The exact-inventory check (§11.8 item 2) additionally requires the realized change set to equal the §9.1 list with no missing or extra path. The regexes alone are not the exactness proof.

---

## 10. Schema, contracts and interfaces

- **Database and schema:** none. No migrations, no domain or tenant data, no Organization branding storage.
- **Public HTTP, MCP and BFF:** none. The routes stay exactly `/`, `/api/health` and `/api/ready` (B-WEB-BOUNDARY). The test-only `/token-fixtures` route exists only in the browser image, and the new `?open=` parameter is fixture-only.
- **Token override input contract:** unchanged (`TokenOverrides`, the 16 names).
- **Internal interfaces that change:** Button and Dialog part props (§5.3); the `PortalContainerContext` value (§5.4); new `INTERNAL_ROLE_TOKENS`; `tokenOverrideStyle` output (derived role sources).
- **Static assets added:** the same-origin `/fonts/**` and `/blueprint/overview.svg`.
- **Authority:** no authentication, authority or tenant selection. web stays a thin BFF.

---

## 11. Tests and validation

### 11.1 Principles

- Everything runs through the **accepted** harness and the approved-input wrapper, Docker-only (G047 Plan r2 §8.7, `8aa4a976…fa65`).
- No new execution framework, debug stack route or resource registry.
- The new oracles live in the existing Playwright surface; the new static guards and controls live in the existing harness functions, using the B-DEPS-CONTROLS pattern.
- Expected values are typed independently (SOURCES, ROLE_SOURCES, source-reference, asset pins) and never read from the implementation.
- Every negative control must be **detected** for its test to pass.
- No required scenario may be skipped, flaky or retried.

### 11.2 Scenario matrix (lib.sh `BROWSER_MATRIX_FULL`; existing IDs kept in the same projects)

| Projects | Existing (unchanged requirement) | Added |
|---|---|---|
| shell-light, shell-dark | S-IDENTITY, S-LANDMARKS, S-NAV, S-SKIP, S-TAB, S-OVERLAY, S-MOTION, S-ERRORS, S-RESPONSIVE, S-A11Y, S-CONTRAST, S-BOUNDARY, S-DESIGN, S-ANGLE, S-SCREENSHOTS, S-PALETTE, S-TEXT-SPACING | **S-FIDELITY, S-FONTS, S-BLUEPRINT, S-LONG-IDENTITY** |
| shell-light | S-HEAD-ORDER, S-HEALTH, S-NOFIXTURE, S-FAILCLOSED | — |
| tokens-light, tokens-dark | S-ERRORS, S-A11Y, S-CONTRAST, S-TOKENS, S-PORTAL, S-RUNTIME, S-ANGLE, S-SCREENSHOTS, S-STATES, S-TEXT-SPACING | **S-PORTAL-INIT, S-FONTS, S-BLUEPRINT** |
| appearance | All eleven IDs unchanged. S-FIRSTPAINT and S-FIRSTPAINT-DELAYED now expect H1 = `ink` (§11.3a). | — |
| dev-hydration | S-HYDRATION-DEV (unchanged code, rerun on the changed page) | — |

`BROWSER_MATRIX_IDENTITY` (rename) is unchanged.

**Fixture page** (`/token-fixtures`; the same `ScopeContent` in all six scopes, plus the `?open=` dialog):
- body and muted text and a link (existing);
- the seven variants, enabled and disabled;
- a size row (sm, default, lg, icon with an `aria-label` and a lucide icon, `block`, and an `asChild` anchor);
- a typography sampler: Kicker, Heading 1/2/3 and `cta`, Text lead/body/meta/index, TextLink, BulletList;
- font probes (Oswald 400–700, Barlow 400/600, Roboto 400/500/700);
- a `raised` Section panel;
- a BlueprintBackdrop panel;
- the ShellHelpDialog.

`?open=<label>` renders that scope's fixture-only client dialog with `defaultOpen`, DialogFooter, `DialogClose asChild` and `Button`.

### 11.3 Existing assertions → G048 replacement (no scenario dropped)

| Scenario / assertion (G047) | G048 replacement | Guarantee kept |
|---|---|---|
| S-OVERLAY: transient focus allowed on `[data-base-ui-focus-guard]` | Allowed on `[data-radix-focus-guard]` only; after two frames focus is inside | Containment |
| S-OVERLAY: name, `aria-modal`, Enter and Space open, Escape and Close, focus return, background inert or aria-hidden, Close ≥ 24px | Unchanged (role and name based) | All |
| S-MOTION: popup and backdrop animate (> 0.05s) without reduced motion; ≤ 0.01s with it; button transitions | Popup and backdrop animation-duration **= 0 in both modes** (the source has no dialog motion; K17 superseded). The button transition control stays: > 0.05s normally, ≤ 0.01s reduced. | Reduced motion removes non-essential motion; a real transition still proves the control |
| S-DESIGN selectors `#overview p`, `#getting-around .grid`, `[data-slot=dialog-trigger]`, Close `svg` colour | `[data-slot=kicker]`, `[data-slot=ruled-grid] > div`, `header button[aria-haspopup="dialog"]`, Close `color` plus the painted centre pixel of the cross | Roles, rules, angled fill, bullets, band, container, dialog panel |
| S-DESIGN: trigger `::before` clip = `--angle`, element unclipped | Unchanged (X7) with the source polygon | Unclipped focus and label |
| S-DESIGN: G047 type metrics (kicker, H3, cell body, About, footer) and dialog geometry (560, 34/40/36) | Moved into **S-FIDELITY** with the full source table (§4.7) | Superset |
| S-PALETTE: 16 sources; H1 = foreground; title = popover-foreground | 16 sources (muted-foreground updated) plus 3 role sources. H1 = **ink**; dialog title = **ink**; lead = text-2; cell body = text-3; footer = muted-foreground (vx-muted) | Independent expected values |
| S-PALETTE `fonts.json`: "at least one platform font" | **S-FONTS** (§7.4): exact custom font, bytes, metadata and synthesis | Strengthened |
| S-CONTRAST selectors (`#overview p`, `.grid p.font-mono`, `.grid p.leading-[1.65]`, `#about li span:last-child`, trigger, Close `svg`) | `data-slot`-based selectors; the trigger by role; Close icon = the cross centre pixel. **Adds local worst-case contrast** over imagery (§11.4). | All pairs kept, plus imagery |
| S-TEXT-SPACING shell items: `header p`, `#overview p:first-child`, `#overview h1 + p`, `#getting-around .grid p:first-child`, `.grid h3 + p`, `header [data-slot=dialog-trigger]` | `[data-slot=site-header] [data-slot=container] > p` (the SiteHeader element is itself the `<header>`), `#overview [data-slot=kicker]`, `#overview h1 + [data-slot=text]`, `[data-slot=ruled-grid] [data-slot=text]:first-child`, `[data-slot=ruled-grid] h3 + [data-slot=text]`, `header button[aria-haspopup="dialog"]`. Same exact texts and kinds. Bundled fonts loaded first, through the bounded phase-B readiness gate (§7.4 B0). | Inventory, clipping, overlap, scroll reachability, negatives |
| S-TEXT-SPACING fixture: "16 expected elements per scope" | The new per-scope inventory (existing 16 plus sampler, size row and probes); count typed in the test | Same checks over more items |
| S-RESPONSIVE: `#getting-around .grid` columns 1/3/3/1/3; dialog fits; Close in viewport | `[data-slot=ruled-grid] > div` columns 1/3/3/1/3 (X18); otherwise unchanged | Reflow |
| S-NAV: target `toBeInViewport` | Plus target heading not covered (`elementFromPoint`), for X16 | Strengthened |
| S-STATES: G047 variants (filled `::before` filter; outline 1px input border; ghost resting ink inherited) | Source variants: filled and muted `::before` fill, element filter 1.12/0.9 (source), label tokens; outline angled 2px `input` frame, inner background, corner = scope surface; ghost resting **accent** ink with border hover fill; link accent ink, 2px accent border-bottom, padding 0/0/4; sizes and `block` computed metrics; focus ring = ring for every variant (source global ring; the G047 destructive ring is superseded); disabled: opacity 0.5, no pointer events, not focusable | Every state, every scope |
| S-ANGLE: FILLED variants, label corners inside the polygon, ring on all sides | Plus outline and muted, the sizes, and the production trigger at 1280 and 320; ring ≥ 3:1 against the adjacent surface on **all four** sides (R6) | Strengthened |
| S-PORTAL: popup `svg` colour = muted-foreground | Cross centre-pixel colour = muted-foreground; otherwise unchanged | Scope tokens |
| S-TOKENS: 16 resolved per scope | Plus 3 roles per scope via `expectedRole` (§6.2); sampler ink/text-2/text-3 per scope | Override propagation |
| S-SCREENSHOTS: shell, dialog, fixtures | Plus overview blueprint clips. The pointer moves to (0,0) before capture (G047 hover-shade note). Wait for fonts through the bounded phase-B readiness gate (§7.4 B0, which includes `document.fonts.ready`). | Visual evidence |
| S-FIRSTPAINT (8 cases) and S-FIRSTPAINT-DELAYED (2 opposing-OS cases): H1 = `PALETTE[m].foreground` | H1 = the independent `PALETTE[m].ink`, with a distinctness precondition against the unchanged `foreground` (§11.3a, r2). Every other expectation, sample, correlation and tolerance is unchanged. `fillOf` samples (0.6w, 3), inside the 15% polygon. Fonts swap after FCP. The held-JS route matches `.js` only, so fonts and the SVG load normally. | First paint and hydration |
| S-HTML-ATTRS, S-HYDRATION-DEV, S-HEAD-ORDER, S-BOUNDARY and the other appearance scenarios | Unchanged code, rerun against the changed page on the root and disabled variants | Appearance, head order and hydration |

### 11.3a First-paint expectation mapping (AC5; added in r2)

G048 renders H1 with the internal `ink` role (§4.4, §6.1). The accepted first-paint assertions in `appearance.spec.ts` (`cbffd6d`, `ba06f0f4…477e`) compare H1 with `PALETTE[mode].foreground`, which is no longer H1's colour. Only the expected H1 colour changes.

**Independent expectations** (`support/modes.ts`). The `ink` values are typed from the kit's `styles/tokens.css` `--vx-ink` (Light line 31, Dark line 10), never read from `globals.css`.

| Appearance | `PALETTE[m].ink` (new) | `PALETTE[m].foreground` (unchanged semantic token) | Distance (`support/color.ts` `distance`, max channel) |
|---|---|---|---|
| Light | `#0d0d0d` | `#1f1d1b` | 18 |
| Dark | `#ffffff` | `#e8e6e3` | 28 |

- `PALETTE` keeps `background`, `foreground`, `primary`, `popover`, `ring` and `accent`, and gains `ink` from `ROLE_SOURCES`.
- `SOURCES[m].foreground` is unchanged and still governs body text, fixture text and the tokens oracles.

**Assertion changes.** These are the only edits to `appearance.spec.ts`.

| Location | G047 expectation | r2 expectation | Tolerance (unchanged) |
|---|---|---|---|
| `evaluateTimeline`: every content frame from the first through hydration, in the ordinary and delayed tests | computed H1 `color` = `PALETTE[want].foreground` | = `PALETTE[want].ink` | ≤ 1 |
| S-FIRSTPAINT-DELAYED: three painted samples while framework JS is held | painted H1 ink (`inkOf`) = `PALETTE[stored].foreground` and ≠ `PALETTE[os].foreground` | = `PALETTE[stored].ink` and ≠ `PALETTE[os].ink` | ≤ 8; > 20 |
| Both tests: new precondition, from the typed tables only | — | `distance(ink, foreground) > 16`, twice the painted tolerance, in both appearances | — |
| Attachments `first-paint.json` and `first-paint-delayed.json` | — | also record the expected `ink` and `foreground` hex values | — |

The precondition keeps the expectations distinct. With a gap above 16, the ≤ 8 acceptance regions around `ink` and around `foreground` are disjoint. An H1 painted in the semantic foreground therefore cannot pass as ink, under either the computed (≤ 1) or the painted (≤ 8) check.

**Preserved without change:**
- **The probe.** It is installed by `addInitScript` before any page script. Samples are timed by their own `performance.now()`, apart from the rAF timestamp and observer delivery.
- **The sample-read/FCP correlation.**
  - At least one theme-bearing content sample is observed at or before FCP; otherwise the run is rejected.
  - The LCP entry is in scope.
  - `data-appearance` is correct when theme-bearing content is inserted.
- **Per-frame checks.** For every frame up to hydration: `data-appearance`, `color-scheme` and background (≤ 1). The painted ground is within 2.
- **The held-framework-JS proof.**
  - The route holds `/_next/static/…js` only.
  - Held chunks are more than 0, and only scripts are held.
  - FCP happens before release, and hydration after release.
  - The primary fill is within 2.
  - The ground equals the stored palette (≤ 2) and is more than 20 from the OS palette.
- **Both opposing-OS cases:** stored Light against OS Dark, and stored Dark against OS Light. Also all eight ordinary cases: three stored modes × two OS settings, plus blocked storage × two OS settings.
- **Hydration continuity.** Every frame through `hydratedAt` is checked, then `expectResolved` and the checked radio after hydration.
- **The tolerances.**
- **No shortcuts.** There is no skip, retry or `fixme`. No screenshot taken after the fonts load stands in for first-paint evidence.

**Fonts and imagery.**
- No first-paint test waits for fonts.
- Fonts and the blueprint SVG are not framework JS, so the delayed test never holds them.
- H1's colour does not depend on the face.
- `inkOf` keeps the sample farthest from the ground. Blueprint strokes (about 2.2:1) are nearer the ground than ink, and the normal H1 lies left of the imagery (§8.4).

**Execution-time validation.**
- **What runs.** The appearance project: S-FIRSTPAINT ×8, S-FIRSTPAINT-DELAYED ×2 and the other appearance scenarios. The dev-hydration project: S-HYDRATION-DEV ×6.
- **Where.** In both B-BROWSER-ROOT (suite `root`) and B-BROWSER-OFF (suite `variants`):
  - first in the retest that turns `root` green, followed by a `variants` run (§11.7);
  - again in S8 `all`.
- **What the Result reports, per variant:**
  - 8 `first-paint.json` and 2 `first-paint-delayed.json` attachments, with the expected and observed ink;
  - FCP before release;
  - held URLs that are `.js` only;
  - `hydratedAt > releasedAt`;
  - 0 skipped, flaky or unexpected tests.

### 11.4 New oracles and their negative controls

| Oracle | Positive | Negative control (must be detected) |
|---|---|---|
| Actual font (S-FONTS) | Phase B (§7.4 B0 and items 1–5) on normal delivery; swap stability (item 8) at 1280 and 320 | Held fonts (phase A only, bounded, released in `finally`) → exactly the `fallback` verdicts for H1, kicker and lead, with the page readable, keyboard-operable and reflowing while still held. Injected 400-only face at 700 → synthesis reported. |
| Local text contrast (`localTextContrast`) | Hide the text (`color` and `-webkit-text-fill-color: transparent`), capture its text-rect clip, decode the PNG, then take the minimum over every pixel of contrast(text colour composited over the pixel, pixel). Must be ≥ 4.5 normal, ≥ 3 large. | Inject a test-only 3px `background: var(--foreground)` bar behind the kicker (inside the isolated container, `z-index: -1`) → minimum < 4.5 reported |
| Imagery painted (S-BLUEPRINT) | Visible versus hidden difference above threshold; stroke pixels toward the expected composite | Hidden-layer capture equals the ground (zero stroke pixels), proving the comparator measures paint |
| Portal insertion (S-PORTAL-INIT, §5.4.1) | Insertion history from `MutationRecord.target` and the added subtrees: exactly one overlay entry and one content entry, both added to the expected scope's container. First-frame and live-mode values equal the scope's. | In the same test, one task per control:<br>• PI-C1 body append;<br>• PI-C2 same-task body→scope relocation (detected even though the delivery-time parent is the container);<br>• PI-C3 same-task subtree relocation;<br>• PI-C4 extraction from a body-inserted wrapper.<br>Each must be detected with its exact entries; control nodes are removed in `finally`. |
| Long identity (S-LONG-IDENTITY, `127.0.0.1:3004`) | 320 and 1280, both appearances: no horizontal scroll; H1 wraps without overflow; header controls inside and non-overlapping; dialog operable; H1 local contrast over imagery ≥ 3 | Shares the responsive overlap detector with S-RESPONSIVE |

### 11.5 Guard reconciliation: old → new → independent control

Edits apply on the **post-G003** versions of these functions (PG-4).

| Guard | Old assertion | New assertion (exact, for this delta only) | Control |
|---|---|---|---|
| B-DEPS (`lib.sh`) | `DEPS_ADDITIONS` includes `dependencies\|@base-ui/react\|1.8.0` (7 additions) | `dependencies\|@radix-ui/react-dialog\|1.1.23` and `dependencies\|@radix-ui/react-slot\|1.3.3` replace it (8 additions). **New** `DEPS_REMOVED_PREFIXES=("@base-ui/")`: any package.json key or **any** lockfile `packages` key matching `(^\|/)node_modules/@base-ui/` → `DEPS_REMOVED_PRESENT <name>`. `DEPS_FORBIDDEN` gains `radix-ui` (umbrella). Existing rules, the `devOptional` justification and the image rule are unchanged. | B-DEPS-CONTROLS |
| B-DEPS-CONTROLS (`run.sh`) | C1–C12, C8b, C8c; C5 mutates `@base-ui/react` | C5 mutates `@radix-ui/react-dialog` to `^1.1.23` → {DEPS_ADDITION_WRONG, DEPS_NOT_EXACT} for it. **New:** C13 leftover direct `@base-ui/react` 1.8.0 (pkg, lock root and entry) → {DEPS_EXTRA_DIRECT @base-ui/react, DEPS_REMOVED_PRESENT @base-ui/react}; C14 lockfile-only `node_modules/@base-ui/utils` → {DEPS_REMOVED_PRESENT @base-ui/utils}; C15 `radix-ui` 1.6.7 → {DEPS_EXTRA_DIRECT radix-ui, DEPS_FORBIDDEN radix-ui}; C16 slot 1.4.0 consistently in pkg and lock → {DEPS_ADDITION_WRONG @radix-ui/react-slot}; C17 dialog recategorized to dev consistently → {DEPS_ADDITION_WRONG @radix-ui/react-dialog}; C18 lock entry 1.1.22 against pkg 1.1.23 → {DEPS_LOCK_MISMATCH @radix-ui/react-dialog}; C19 (r2) `tailwind-merge` exactly `3.7.0`, consistent in package.json, the lock root and a `node_modules/tailwind-merge` entry with no dev flag → {DEPS_EXTRA_DIRECT tailwind-merge} (see "C19 exact semantics" below). C12, the real pair, passes. | Each expected set exact (no missing or extra code); all evaluated independently |
| B-WEB-TOKENS | 16 names equal in five places; palette `--light/--dark` names exactly 16; `@theme` names `--container-shell --font-display --font-label --font-mono --font-sans --spacing-gutter --spacing-section`; no `bg-accent`/`accent-foreground` | 16 names unchanged in `SEMANTIC_COLOR_TOKENS` **and** a separately parsed `INTERNAL_ROLE_TOKENS` (exactly `ink text-2 text-3`). Palette, resolution and `@theme inline` sets each equal 16 ∪ 3 = 19. `TokenOverrides` still keyed by `SemanticColorToken` (grep). `@theme` names plus `--spacing-cell`. `:root` non-colour still exactly `--radius --angle`. Ejected digest, literal, palette-utility and accent rules unchanged. A new rule rejects untranslated kit colour names, `(bg\|text\|border\|from\|via\|to\|fill\|stroke)-(gold\|rule\|surface\|raised\|ground\|orange\|faint\|star)\b`, in `web/{components,app,lib}`. The `ink`, `text-2` and `text-3` roles are allowed. | Static; a typo or extra name fails the equality |
| B-WEB-COMPONENTS | `components.json` base-nova; `components/ui` exactly button, dialog, portal-container; `@base-ui` imports; E1–E7 | `components.json` identical except `style radix-nova`. `components/ui` exactly the seven files (§9.1). button imports `@radix-ui/react-slot`; dialog imports `@radix-ui/react-dialog` and uses `usePortalContainer`. **No `@base-ui` import anywhere in web sources, tests included** (`UI_IMPORT_BASEUI`). The only `@radix-ui/*` specifiers are those two (`UI_IMPORT_RADIX`). Every `data-slot="…"` literal in `web/{app,components,lib}` is in the reviewed set: kicker, heading, text, text-link, container, section, ruled-grid, bullet-list, bullet-item, site-header, button, dialog-overlay, dialog-content, dialog-header, dialog-footer, dialog-title, dialog-description (`UI_SLOT <slot>`; every kit module's distinctive slot is outside it). Button base has the X9 focus outline. No element-level `[clip-path:var(--angle)]` without a `before:`/`after:` prefix. Eject marker and `lib/utils` re-export unchanged. Refactored into `web_components_check ROOT` emitting sorted codes. | B-WEB-ASSET-CONTROLS |
| B-WEB-DESIGN | No remote font, **no `@font-face`**, no remote asset URL; no web file byte-identical to a kit asset; no brand or marketing literal | Remote-font, remote-URL, kit-copy and literal rules unchanged. `@font-face` allowed **only** in `web/app/globals.css`, exactly 4 blocks, and the set of `url(…)` targets in web CSS equals the 4 font paths plus `/blueprint/overview.svg` (`FONT_FACE`, `DESIGN_REMOTE`). **`web/public` exact inventory**: exactly the 8 files of §9.1 rows 19–26, each SHA-256 equal to its pin (`ASSET_EXTRA`, `ASSET_MISSING`, `ASSET_HASH`). **SVG safety and motifs** per §8.2 (`SVG_PARSE`, `SVG_UNSAFE <token>`, `SVG_SIZE`, `SVG_MOTIF <group>`). Refactored into `web_design_check ROOT`. | B-WEB-ASSET-CONTROLS |
| B-NAMING | `web/public` forbidden as a redundant path | `web/public` removed from the forbidden list. Its contents are bounded by B-WEB-DESIGN's exact inventory. The other names are unchanged. | A3 and A2 (extra public files fail through the inventory) |
| **B-WEB-ASSET-CONTROLS** (new, static) | — | Scratch copies of `web/` with exact expected code sets, all A1–A13 evaluated independently (list below). | Self |
| B-RENDER-PLAN, -REPEAT, -ID | Register exactly 8 (3 G001 + 5 G002) | Register = post-G003 count + 1 (`infra/docker/web.Dockerfile`, a rendered path). The count assertion and message are updated (PENDING PG-5). | Repeat plan and render conflicts = register |
| B-TOGGLE-OFF-ADDITIONS | `APPLICABLE` (28; post-G003 34, PENDING PG-6) | Plus `web/public`, `web/THIRD_PARTY_NOTICES.md` and `infra/docker/web.Dockerfile` (customized). The disabled variant gets fonts, imagery and notices byte-identical. | Byte comparison in the check |
| B-BROWSER-ROOT, -OFF, -RENAME | Matrix §11.2 old IDs | Plus the §11.2 IDs | `browser_matrix` |
| B-HADOLINT | migrate and web-browser Dockerfiles | Unchanged list. The one added `COPY` must pass. `web.Dockerfile` hadolint status follows PG-15. | — |
| B-REQUIRED-CHECKS (G003) | PENDING PG-7 | `B-WEB-ASSET-CONTROLS` added once, suite `static` | G003's own controls |
| B-SCOPE | G047 allow-list | §9.2 | Exact inventory (§11.8) |
| Unchanged | B-WEB-BOUNDARY, B-LITERALS, B-DOCKERIGNORE, B-PROTECTED, B-SKILL-UNCHANGED, B-SECRET-CANARY, B-WRAPPER-CONTROLS, `init_hash_documented`, provenance, migrations, isolation, ports, rename, outage, persist and every G003 check | — | Full S8 run |

**C19 exact semantics (r2).** This is option (a) of the review: a concrete exact version with a consistent package/lock pair, so the control isolates one code.
- **Version.** `tailwind-merge` **3.7.0**. It was the registry's `latest` dist-tag on 2026-10-06, is MIT and has no dependencies or peers (read-only registry metadata, §18). The control never installs it; the version only needs to be a real, exact specifier.
- **Mutation** (the accepted C2/C10 pattern):
  - package: `.dependencies["tailwind-merge"] = "3.7.0"`;
  - lock: `.packages[""].dependencies["tailwind-merge"] = "3.7.0" | .packages["node_modules/tailwind-merge"] = {"version": "3.7.0", "resolved": "$reg/tailwind-merge/-/tailwind-merge-3.7.0.tgz"}`.

  The real lockfile has no `tailwind-merge` entry (`cbffd6d`; `cn` 0.4.0 has no dependencies), so the entry is new and nothing else changes.
- **Every rule, evaluated independently on this pair** (`deps_pair_check`, with r1's §11.5 additions):

| Rule | Outcome for C19 |
|---|---|
| Existing dependency missing, recategorized or changed | none: unchanged entries |
| Reviewed additions (8, including the Radix pins) | none |
| Unapproved direct key | `DEPS_EXTRA_DIRECT tailwind-merge` |
| Forbidden name (`shadcn`, `radix-ui`) | none: `tailwind-merge` is not on the forbidden list; the extra-direct rule rejects it |
| Exact specifier (`DEPS_NOT_EXACT`) | none: `3.7.0` is exact |
| Locked version equals the specifier | none: the entry is 3.7.0 |
| Dev flag against category | none: with no `dev` flag the entry is runtime, which matches `dependencies` |
| Lock-root specifier and category | none: identical to package.json |
| Removed prefix `@base-ui/` (`DEPS_REMOVED_PRESENT`) | none |

- **Expected result.** The sorted set is exactly `DEPS_EXTRA_DIRECT tailwind-merge`, and `deps_pair_check` exits non-zero. `deps_expect` compares the `LC_ALL=C sort -u` output exactly, so a missing or extra code fails the control.
- **Nothing relaxed.** No rule changes for C19. r1's range form is withdrawn. A range would add `DEPS_NOT_EXACT`, a code C5 already exercises.
- **Execution-time validation.** The static suite runs in the same retest as any `lib.sh` or `run.sh` edit, and again in S8 `all`. It covers:
  - B-DEPS on the real repository pair (`deps_repo_check`);
  - B-DEPS-CONTROLS: C1–C19 with C8b and C8c, the C11 image case, and C12 (the real pair and repository inputs) passing.

  Every row of `deps-controls.tsv` must read `pass`, with its exact want and got sets.

**B-WEB-ASSET-CONTROLS cases** (each a disposable copy under `$WORK`, compared with an exact expected code set):

| # | Mutation | Expected codes |
|---|---|---|
| A1 | Kit `ribbon.tsx` copied to `web/components/ui/ribbon.tsx` with one byte appended | {UI_SET, UI_SLOT ribbon} |
| A2 | Kit `public/brand/logo.png` copied byte-identical to `web/public/brand/logo.png` | {ASSET_EXTRA web/public/brand/logo.png, DESIGN_KIT_COPY web/public/brand/logo.png} |
| A3 | `web/public/fonts/extra/x.ttf` added; and separately `oswald-wght.ttf` with one byte changed | {ASSET_EXTRA …}; {ASSET_HASH web/public/fonts/oswald/oswald-wght.ttf} |
| A4 | `globals.css` gains an `@font-face` with `src:url("https://fonts.gstatic.com/x.woff2")` | {DESIGN_REMOTE web/app/globals.css, FONT_FACE web/app/globals.css} |
| A5 | An `@font-face` string in a component | {FONT_FACE <file>} |
| A6 | SVG gains `<script/>` | {ASSET_HASH <svg>, SVG_UNSAFE script} |
| A7 | SVG gains an `onload=` attribute | {ASSET_HASH, SVG_UNSAFE onload} |
| A8 | SVG gains `<image href="https://x/y.png"/>` | {ASSET_HASH, SVG_UNSAFE image, SVG_UNSAFE href} |
| A9 | SVG gains `<foreignObject/>` | {ASSET_HASH, SVG_UNSAFE foreignObject} |
| A10 | SVG grown past 40960 bytes with path data | {ASSET_HASH, SVG_SIZE} |
| A11 | SVG `dimensions` group removed | {ASSET_HASH, SVG_MOTIF dimensions} |
| A12 | `import { Dialog } from "@base-ui/react/dialog"` added to `overlay.spec.ts` | {UI_IMPORT_BASEUI web/tests/browser/overlay.spec.ts} |
| A13 | The real inputs | {} |

**Rule.** Shared harness edits trigger the static and provenance suites in the same cycle (G002 lesson; G047 §10 S4).

### 11.6 Evidence

**Evidence root.** `$EV = ~/.local/state/platform-evidence/g048-<EXEC_ID>` (G047 convention: created only after the containment checks, outside checkouts and `/tmp`, `mkdir -m 700`).

**Recorded:**
- step logs and reports;
- `runs.tsv`;
- intake records: font URLs, sizes, blob SHA-1, SHA-256; the lockgen image digest and npm version; the lock delta report;
- the source → target `diff -u` for each of the six modules (the alteration table's evidence);
- before (S1) and after (S8) screenshots: shell and dialog at 1280 and 320 in Light and Dark; fixtures; overview blueprint clips; the SVG document and per-group captures;
- the report attachments:
  - `fonts.json`, with the phase verdicts and the control step lists (§7.4);
  - `blueprint.json`;
  - `source-fidelity.json`;
  - `contrast.json`, with local imagery minima;
  - `spacing-report.json`;
  - `first-paint*.json`, with the expected `ink` and `foreground` (§11.3a);
  - `portal-init.json`, with the insertion history and control outcomes (§5.4.1);
  - `axe-summary.json`;
- `deps-controls.tsv` (C1–C19, C8b, C8c, C11, C12);
- a manifest with SHA-256 per artifact, plus commit, tree and run IDs.

No secrets. Destructive and negative mutations happen only in harness-owned disposable copies or in-test DOM.

### 11.7 Commands (reuse, no redesign)

- **Startup and invocation.** These are G047 Plan r2 §8.7.1 and §8.7.2, reused verbatim except for:
  - the step label `g048`;
  - the G048 binding values (`BASE`, `SNAP`, `SNAP_SHA`, `N`, `RE`, `PLAN`, `ALLOW=management/v0/results/G048-scope-allowlist.txt`, `EV`);
  - adjustments to any post-G003 interface change (PG-11).

  That is: explicit Bash, immediate `PIPESTATUS` capture, wrapper-status precedence, 70/71 for logging failures, and per-run rows.
- **Steps:**
  - `s1 all`;
  - `s3 static provenance`;
  - `s4a…` `static provenance root`;
  - **once `root` is green, one `static provenance root variants` retest before S6** (r2). This puts the disabled variant's appearance, first-paint, dev-hydration, S-PORTAL-INIT and font-control results in hand before S8. A suite added to a retest is not a cycle (BUDGET_POLICY §3–4);
  - `s8 all`, plus any suite G003 adds to the required final set (PG-4, PG-7).
- **Dependency controls.** Any retest that follows a `lib.sh` or `run.sh` edit includes `static`. That reruns B-DEPS on the real pair and the complete B-DEPS-CONTROLS set (§11.5).
- **Wrapper, harness and cleanup.** Unchanged procedures. Cleanup is the harness's own owned-resource cleanup and leftover report. The only other Docker use is the single `--rm` lockgen container (§5.2), which creates no volume, network or project.

### 11.8 Integrity, provenance and cleanup checks around S8 (G047 §8.8 reused)

These run with no repository write during S8:
1. Pre/post tree digest and NUL status digest are equal.
2. The full change inventory equals the §9.1 path list exactly (none missing, none extra). The ignored inventory is empty, or only `.claude/settings.local.json`.
3. Kit manifest 63 OK at S0 and after S8. `skills/**` is unchanged.
4. The package diff is exactly §5.1. The lockfile delta equals the §5.2 report.
5. Added `^+import` lines under `web/{app,components,lib}` are limited to `@radix-ui/react-dialog`, `@radix-ui/react-slot`, `class-variance-authority`, `react`, `@/lib/utils`, `@/components/**` and the existing modules.
6. The font, notice and SVG hashes equal their pins and the intake records.
7. Provenance line check (supplementary): added lines outside the six adopted modules and the tests, compared with every kit line of 40 or more characters, give zero matches, or matches already at `BASE`. The six module diffs map hunk by hunk to T/X entries.
8. No `.env` content in evidence.
9. Cleanup: the per-run `inventory-leftover.txt` plus a final read-only inventory comparison. G048 deletes nothing by itself.

### 11.9 Hosted CI (PENDING PG-8; Q7)

If G003's accepted contract requires hosted evidence for subsequent Goals, the **accepted G003 job** is run unchanged on the G048 commit after S8. It is not a new workflow, and it requires separately authorized push and pipeline actions (P9).

- Its outcome is recorded with its pipeline ID.
- Missing access is a completion blocker, never a pass.
- If no hosted gate applies, the Result says so with the G003 reference.

### 11.10 Documentation (`docs/foundation/WEB-SHELL.md`)

Updated sections:
- the four concerns (A now includes source-derived components, bundled fonts and imagery);
- the token contract (16 plus 3 internal roles, propagation, the muted-foreground change);
- non-colour tokens (the source angle, spacing);
- a **Fonts** section (artifacts, hashes, licences, loading, oracle);
- an **Imagery** section (composition, safety, placement, theme mapping);
- **Components and provenance**: lineage table, T and X tables, the owner rights record (P3), the shadcn MIT notice reference, the Radix pins, the portal contract;
- dependencies;
- "Adding a component" (`radix-nova`, Radix only);
- browser tests (the new scenarios and the fixture);
- status and known limits (Chromium-only; local font metrics depend on the bundled TTFs; the bounded text-spacing trial).

The historical G002 and G047 subsections stay labelled as history.

---

## 12. Delegation

- **Planning.** One bounded read-only web-research subagent (font provenance: commits, blobs, sizes, metadata, OFL terms). It made no writes and consumed zero cycles. I verified its key values against the requirements; its blob values are re-verified at intake.
- **Implementation.** By me only. Every path has one owner, and shared harness and test files make split ownership unsafe.
- **Review.** One read-only visual and accessibility reviewer after the first fully green `root` run and before S8. It reviews:
  - the diff against §4–§8;
  - the before and after screenshots;
  - `fonts.json`, `blueprint.json`, `source-fidelity.json`, `contrast.json` and `spacing-report.json`;
  - the six module diffs against the T and X tables;
  - the run rows.

  It reports findings only. I triage and integrate them, and any corrective implementation counts toward B5.
- **Not delegated:** architecture, scope, rights decisions, exceptions, integration, the Result.

---

## 13. Sequence and gates

| Step | Work | Gate / outcome | Cycles |
|---|---|---|---|
| P1–P9 | Prerequisites (§2.5) by their owners | All complete; PG reconciliation re-reviewed | — |
| S0 | Read-only startup (§11.7):<br>• `EXEC_ID` then `$EV`;<br>• Git, status and inputs;<br>• kit manifest 63 OK;<br>• **the P3 rights record present at `BASE`**;<br>• wrapper `--check-only`;<br>• resource checks | Any failure is a startup failure: stop, no write, no cycle | 0 |
| S0b | Write the allow-list (§9.2); record its hash | — | 0 |
| S1 | `s1 all` on the unmodified base; before screenshots; counts | Overall 0; otherwise a baseline failure (stop) | 0 |
| S2a | Intake in `$EV`: fonts (§7.1) and lockgen (§5.2), each verified against the Plan | A mismatch stops the work and is reported (not a cycle) | 0 |
| S2b | Implement §9.1 in the §3 order; author the SVG; pin the asset hashes | — | Initial implementation |
| S3 | `s3 static provenance`, then `s4a static provenance root` | The first validation result; a material failure opens cycle 1 | — |
| S4 | Diagnose from the harness reports; fix; retest (`s4b…`). An edited harness file means static runs in the same cycle. Once `root` is green, one `root variants` retest (§11.7). | Green `root`, then green `variants` (disabled-variant browser matrix, including §11.3a, §5.4.1 and §7.4) | Cycles 1–3 |
| S5 | **Checkpoint after cycle 3** (BUDGET_POLICY §7), recorded before any cycle 4: ACs done and remaining, failure patterns, convergence, boundedness | Continue only if converging; otherwise stop and return with the §14 split proposal | — |
| S6 | Read-only review (§12) and triage | Findings integrated or recorded | Fixes count |
| S7 | Finalize WEB-SHELL.md and the notices; draft the Result | — | 0 |
| S8 | §11.8 pre-checks; `s8 all` (plus G003-required suites); §11.8 post-checks | Overall 0; no FAIL, UNMET or skip; exact inventory | A fix uses the next cycle; rerun as `s8b` |
| S8h | Hosted CI (§11.9) if required and authorized | Recorded | — |
| S9 | Provenance line check; evidence extraction; leftover report | — | 0 |
| S10 | Result (AGENTS.md format, AC1–AC10, every run row, cycle log, checkpoint, deviations including any D-ANGLE use); return to Codex | Stop | — |

---

## 14. Budget, forecast and checkpoint (B5)

**Forecast: 4 cycles (range 3–5).**

| Risk source | Probability of costing a cycle | Mitigation |
|---|---|---|
| Radix modal and portal behaviour versus the inherited assertions | Medium | §5.4 adapter design; existing inert/aria-hidden predicate; axe modal semantics analysed |
| Source geometry under reflow, text spacing and angle | Medium | X12, X13, X16–X18 pre-declared; D-ANGLE pre-declared |
| Font oracle details (CDP names, twin determinism, swap timing) | Medium | Names come from the binary itself. Fonts are held, not aborted. r2 separates bounded phase-A checks, which never wait for fonts, from the phase-B oracle behind a bounded readiness gate, and every hold is released in `finally` (§7.4). |
| Harness guard refactor and controls (bash, exact code sets) | Medium | The B-DEPS-CONTROLS pattern; static runs in the same cycle |
| Post-G003 integration surprises (counts, required checks, CI) | Low to medium | PG list resolved before approval |

- **Assessment.** B5 is plausible but **tight**: about 20–25% probability of needing more than 5. I do not unilaterally change it.
- **Codex direction (2026-10-06).**
  - B5 stays, with a checkpoint after cycle 3.
  - The four-cycle forecast is plausible but tight, and is reassessed against the accepted post-G003 tree (P2).
  - No budget extension or split is authorized.
- **Effect of r2 on the forecast.** r2 keeps the forecast at 4 (range 3–5).
  - Its corrections remove two failure modes: a hang or false pass in the held-font control, and missed transient body portalling.
  - They add test code: the history recorder with four in-page controls, a phased font helper, and two expectation edits in `appearance.spec.ts`.
  - The "Radix modal and portal" and "Font oracle details" rows above already cover that risk.
- **Q8 options:**
  - (a) keep B5 (my recommendation, with the checkpoint);
  - (b) a reviewed B7;
  - (c) the pre-agreed split below.
- **Split point**, if the S5 checkpoint shows non-convergence:
  - **G048a:** the Radix migration plus the source button and dialog, the portal adapter, the B-DEPS and B-WEB-COMPONENTS reconciliation and their tests;
  - **G048b:** typography, layout, header and bullets, fonts and blueprint, with their guards and tests.

  This split preserves validation at each step. It is proposed to Codex, never self-applied. r2 adds no new split or extension proposal; Codex has authorized neither.
- **Accounting.** Zero cycles consumed by planning.

---

## 15. Acceptance-criteria mapping

| AC | Evidence (where produced) |
|---|---|
| **AC1** Six modules adopted with Radix, rights and traceable migration | §4.1 lineage and §4.4 alterations; the six `diff -u` files; P3 record and notices (§4.6); B-WEB-COMPONENTS (file set, Radix imports, no Base UI, `data-slot` inventory); S-FIDELITY rendered `data-slot` presence; B-DEPS exact pins and `DEPS_REMOVED_PRESENT` absent; §11.8 items 4–5; §5.3 inventory with every call site migrated |
| **AC2** Source geometry, type and states within reviewed exceptions | §4.7 table → S-FIDELITY `source-fidelity.json` (±1px, exact categories); S-STATES (all variants, sizes, `block`, outline 2px angled frame, production `sm` trigger); S-ANGLE; screenshots against the kit screenshots; T and X tables; D-ANGLE reported if used. **r2:** S-FIDELITY, S-STATES and the screenshots measure only after the bounded phase-B font readiness gate (§7.4 B0). A font that never loads fails there by name; it never yields fallback metrics or a hang. |
| **AC3** Bundled fonts render | §7.1 intake (URL, size, blob SHA-1, SHA-256, metadata, notices); B-WEB-DESIGN pins; S-FONTS on root **and** disabled variants plus fixture (bytes, name/`fvar`/OS/2, CDP custom font, weights, twin synthesis). **r2 (§7.4):**<br>• the positive phase-B oracle on normal delivery;<br>• the held-font control, run in phase A only, bounded and released in `finally`: exactly the `fallback` findings for H1, kicker and lead, readable and operable fallback, and a still-held proof;<br>• the synthesis control detected;<br>• swap stability completed at 1280 and 320. |
| **AC4** Blueprint present, safe, decorative, both appearances | §8.2 static checks (B-WEB-DESIGN); S-BLUEPRINT (semantics, on/off paint, motif captures, missing asset, forced colours, layout neutrality); desktop and 320 screenshots in Light and Dark; local text contrast over imagery |
| **AC5** Appearance, first paint and hydration intact | The appearance project's 11 scenario IDs. **r2 (§11.3a):** S-FIRSTPAINT ×8 and S-FIRSTPAINT-DELAYED ×2 on the changed page now expect H1 = the independent `ink` (`#0d0d0d`/`#ffffff`), while `foreground` stays distinct (`#1f1d1b`/`#e8e6e3`, precondition gap > 16). Changed files: `appearance.spec.ts` (§9.1 row 51, allow-list §9.2) and `modes.ts` (row 34). Unchanged:<br>• the sample/FCP correlation;<br>• the held-framework-JS proof;<br>• both opposing-OS cases;<br>• hydration continuity;<br>• the tolerances.<br>Also S-HTML-ATTRS, S-HEAD-ORDER (init hash unchanged) and S-HYDRATION-DEV ×6. Everything runs on the root and disabled variants (§11.7, S8). No skip, and screenshots never substitute for first-paint evidence. |
| **AC6** Scoped overrides compose | S-TOKENS (16 plus 3 roles via `expectedRole`); S-PORTAL; S-RUNTIME (roles, imagery colour); S-BLUEPRINT fixture panels; partial Light-only and Dark-only, nested and sibling; no Organization storage or API (§10). **S-PORTAL-INIT (r2, §5.4.1):**<br>• insertion history from `MutationRecord.target` and the added subtrees;<br>• exactly the overlay and content added to the expected container;<br>• first-frame and live-mode values;<br>• PI-C1–PI-C4 detected, including the same-task body→scope relocation;<br>• the adapter's no-body-fallback contract unchanged.<br>Codex labelled this finding AC4 (§0.1, Q13). |
| **AC7** Accessibility with real fonts, imagery and geometry | S-A11Y (axe closed and open, fixtures); S-CONTRAST (states, local imagery minima, negative control); S-ANGLE (all-sides ≥ 3:1); S-OVERLAY; S-TAB; S-SKIP; S-NAV (uncovered); S-RESPONSIVE (320, 768, 1280, 640×400, 200% text); S-TEXT-SPACING with bundled fonts; S-MOTION; S-LONG-IDENTITY; forced colours. **r2 (§7.4):**<br>• the held-font control completes its readability, keyboard (trigger, open, trap, Escape, focus return) and reflow (320 and 1280) checks while the fonts are still held, within explicit bounds;<br>• swap stability: no focus shift, overlap, overflow or dialog loss when the fonts arrive. |
| **AC8** Guards reconciled without broad paths | §11.5 table; B-DEPS-CONTROLS C1–C19 exact (C19, r2: `tailwind-merge` 3.7.0 as a consistent pair → exactly `DEPS_EXTRA_DIRECT tailwind-merge`, every rule evaluated independently); B-WEB-ASSET-CONTROLS A1–A13 exact; B-WEB-TOKENS, -COMPONENTS, -DESIGN, B-NAMING; browser negative controls |
| **AC9** Full inherited validation passes | S8 `all` (plus G003-required suites): wrapper, `tee` and overall 0; pre/post postcheck; root, disabled and rename browser runs; README and docs blocks; migrations, isolation, ports, outage and persist; hosted CI per §11.9 and Q7. **r2:** the S8 static suite reruns B-DEPS on the real repository pair and the complete B-DEPS-CONTROLS set (C1–C19, C8b, C8c, C11, C12), every row `pass` with exact sets and no rule relaxed. |
| **AC10** Provenance, scope, cleanup and record | B-SCOPE plus the §11.8 exact inventory; kit manifest; source, font, licence and notice records; artifact SHA-256 manifest; Result with the AC checklist, cycle log and checkpoint; no schema, API, MCP or product change; leftover report; Codex ACCEPT |

---

## 16. Risks, stop conditions and escalations

**Stop and return to Codex for any of these:**
- G003 is not accepted or integrated, a PG fact contradicts this Plan, inputs drift, or resources conflict;
- P3 rights are unresolved, or the owner requires notices the guards cannot accept without an exception;
- a need for any package outside §5.1 or §5.2 (including `clsx`, `tailwind-merge` or another Radix package), a token-contract break, an appearance-engine change or body portalling inside a scope;
- an accessibility conflict not resolvable by the X list or D-ANGLE;
- whole-library or product expansion;
- unsafe or dynamic SVG;
- remote assets;
- any weakening of an inherited check;
- non-convergence at the checkpoint, or B5 exhaustion.

**Specific risks:**

| Risk | Handling |
|---|---|
| (a) Radix `hideOthers` uses `aria-hidden`, not `inert` | The accepted predicate covers both, and axe passes with a modal open. Verified in S-A11Y and S-OVERLAY. |
| (b) Slot lacks `"use client"` | Mitigated (§5.5); the pre-declared adaptation is used if needed |
| (c) `cn` merge parity | Verified by computed styles. If it fails, the fix stays within module classes: restructure the conflicting variants. No new package. |
| (d) Turbopack treatment of root-relative CSS `url()` | Proven by the build and by S-FONTS and S-BLUEPRINT responses. If it is rewritten or rejected, stop and report. |
| (e) Chromium mask support | `mask-image`/`mask-composite` are unprefixed in current Chromium (Playwright 1.63). Verified by S-BLUEPRINT. |
| (f) Test runtime growth | May interact with G003 CI timeouts (PG-8) |
| (g) Synthesis twin check sensitivity | Byte equality on the same element in the same frame. A false positive would show on all nodes; diagnose before changing the method. |
| (h) Security | No new authority or routes. Assets are static and same-origin. The SVG is static and allow-listed. Downloaded inputs are verified by hash and never executed. Lockgen runs with `--ignore-scripts`. There is no CSP today (G012 owns it); the init-script hash is unchanged. |
| (i) Held-font tests hanging or passing on a timeout (r2) | Phase A never awaits `fonts.ready`, a font response or the `load` event, and takes no screenshots while fonts are held. Every wait has a named bound that fails on expiry. Detection is an asserted verdict value, never a caught exception. Holds are released in `finally` (§7.4). |
| (j) Transient body portalling hidden by delivery-time parents (r2) | The evidence uses only `MutationRecord.target`, the added subtrees and the removal targets. The PI-C2–PI-C4 relocation controls must be detected (§5.4.1). |

---

## 17. Questions for Codex

**Blocking:**
- **Q1 — Source rights (P3).** The owner must confirm:
  - (a) ownership of, and permission to use and modify, the four brand modules and the restyled button and dialog;
  - (b) any required attribution text.

  If an in-file notice must name the owner's company, B-WEB-DESIGN's literal scan needs an exact-line exception (Q1b). That requires Codex approval.
- **Q7 — Hosted CI.** Does G003's accepted contract require hosted pipeline evidence for G048? If so, the specific push and pipeline actions need authorization (P9).
- **Q9 — Binding.** The snapshot composition (§2.5, 27–28 entries) and the base branch and commit route after G003 integration.

**Confirmations** (the Plan proceeds on the stated default unless Codex directs otherwise):
- **Q2.** Radix 1.1.23/1.3.3 (default) or 1.2.0/1.4.0 (newer, one day old)?
  - **Answered by Codex, 2026-10-06:** the exact pins 1.1.23/1.3.3 are valid registry versions, and no upgrade is required. The Plan keeps them.
- **Q3.** `web/public` with the B-NAMING relaxation (contents exact-pinned), two Dockerfile `COPY` lines and a register row (default). The alternative — CSS-emitted `/_next/static/media` assets — avoids the Dockerfile change but cannot place the OFL notices in the image.
- **Q4.** Omit Emph (default) or retain it with an `orange` internal role?
- **Q5.** `components.json` `style` → `radix-nova` (default) or leave `base-nova` with a documentation note?
- **Q6.** Return `muted-foreground` to the kit's vx-muted (default; it supersedes G047 K5) so meta and body stay distinct.
- **Q8.** Budget: B5 (default), a reviewed B7, or the §14 split.
  - **Codex direction, 2026-10-06:**
    - B5 stays, with a checkpoint after cycle 3.
    - No extension or split is authorized.
    - The four-cycle forecast is plausible but tight, and is reassessed against the accepted post-G003 tree.
- **Q10.** `lucide-react` stays declared, used only by the fixture `icon` button. Its removal is outside the authorized delta.
- **Q11.** S-MOTION mapping (§11.3): there is no dialog motion, matching the source. The button-transition control is kept.
- **Q12.** D-ANGLE conditional fallback pre-approval (§4.3).
- **Q13 (r2).** The AC mapping of the review findings (§0.1).
  - The review labels the findings AC4 (insertion history), AC9 (C19) and AC2/AC7 (blocked fonts).
  - The Spec r2 criteria place insertion history under AC6, the C19 control under AC8 (with its final rerun under AC9), and blocked-font evidence under AC3 and AC7.
  - §15 records them under the Spec's criteria, plus AC2 for the font-readiness gate. AC4 (blueprint) is unaffected.
  - Default: this mapping, unless Codex directs another.

**Still open after r2.** r2 does not resolve any of these, and claims no owner permission or accepted predecessor state:
- **Rights and notices.** The selected-code rights and notices: Q1, Q1b and P3.
- **G003.** G003 implementation acceptance and integration with the G047 checkpoint (P1), and the PG-1–PG-15 reconciliation with re-review (P2).
- **Approval.** Final Codex Plan approval (P4).
- **Binding.** The supervisor-created committed execution binding: P5–P7 (`BASE`, the execution worktree, `SNAP`, `SNAP_SHA`, `N`, `RE`).
- **Hosted CI.** The hosted-CI requirement (Q7, P9).
- **Binding composition.** The binding composition (Q9).

---

## 18. Planning research record (read-only; no downloads to disk)

- **npm registry metadata.**
  - `@radix-ui/react-dialog` 1.1.23 and 1.2.0 dependencies, integrities and peers;
  - `@radix-ui/react-slot` 1.3.3 and 1.4.0;
  - every transitive package listed in §5.1;
  - versions and licences of `aria-hidden`, `react-remove-scroll(-bar)`, `react-style-singleton`, `use-callback-ref`, `use-sidecar`, `get-nonce` and `detect-node-es`.
- **radix-ui/primitives.**
  - `packages/react/{dialog,slot}/CHANGELOG.md`;
  - `packages/react/portal/src/portal.tsx` at `579c5b84` (container fallback to `document.body`);
  - `dialog.tsx` at `579c5b84`; release commits `579c5b84` (2026-07-25, dialog 1.1.23) and `d5767dcb` (2026-10-05, dialog 1.2.0);
  - unpkg dist heads: dialog 1.1.23 has `"use client"`; slot 1.3.3 does not and uses `useComposedRefs`.
- **Licences.** shadcn/ui `LICENSE.md` (MIT, 2023 shadcn). radix-ui/primitives `LICENSE` (MIT, 2022 WorkOS).
- **`cn` package.** `github.com/shadcn-ui/cn` README: tailwind-merge v3 parity on Tailwind v4.
- **axe-core** (develop): `aria-hidden-focus.json`, `aria-hidden-focus-matches.js`, `focusable-not-tabbable-evaluate.js` (passes when `isModalOpen()`) and `is-modal-open.js`.
- **google/fonts** at `7085eb89…140c` (via the subagent): directory listings, blob SHA-1s, sizes, `METADATA.pb` and `OFL.txt`. Also OswaldFont, jpt/barlow and roboto-3-classic release and asset facts.
- **shadcn component styles** (`radix-nova`/`base-nova` naming), from a web search.

**Local actions.**
- Read-only inspection of the planning checkout and its object store (`git show cbffd6d:…`, `sha256sum`).
- A read-only lockfile graph analysis via `python3 -I` over the checked-in `web/package-lock.json`. That file is in the checkout itself, not downloaded.
- The only file written is this Plan.
- No package, font or image was downloaded to disk.
- No Docker, npm or build command was run.
- The main checkout and the G003 execution worktree were not accessed.

**r2 additions** (read-only; no downloads to disk; no subagent):
- **npm registry metadata for `tailwind-merge`.**
  - dist-tags: `latest` was 3.7.0 on 2026-10-06.
  - The 3.7.0 manifest:
    - MIT;
    - tarball `https://registry.npmjs.org/tailwind-merge/-/tailwind-merge-3.7.0.tgz`;
    - integrity `sha512-XPPUyAc+cvspz3lHTcR/QgPfW2A0lv/xQNIjX3HGhLR+Nq2lHaLq5MtTesHn8GUr3W3DguT2KT5x3NVgRtYwmA==`;
    - no dependencies or peers.
  - The 3.3.1 manifest was also read but not chosen.
- **Local reads.**
  - In the main checkout, read only: the PROGRESS.md review section and `docs/agents/CLAUDE-SESSION-HANDOFF.md`.
  - A read-only `git diff --stat dc1ff7c 3d022ef` of the Goal, Spec (with r1 histories) and ADR-001, which was empty.
  - The `cbffd6d` checkpoint files listed in §2.2, through `git show` in the planning clone.
- **Writes.** The only file written for r2 is this Plan, revised in place.
  - The r1 archive (`…-plan-r1.md`, `2f0203dc…9a53`) was already present when r2 drafting began, and is unchanged.
  - No Docker, npm, build, install, commit or hosted action.
  - The G003 execution worktree was not accessed.
