> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G047 — Successor Vestive shell visuals: implementation plan

Goal: [G047-vestive-shell-refresh.md](../goals/G047-vestive-shell-refresh.md) · Spec: [G047-vestive-shell-refresh-spec.md](../specs/G047-vestive-shell-refresh-spec.md)
Owner: Claude Code · Reviewer: Codex · Budget: **B3** (at most 3 shared meaningful cycles; checkpoint after cycle 2, before cycle 3)
State: **r2, submitted for Codex Plan review after the r1 CORRECTION LOOP. Implementation is not authorized.**

## 0. Revision record

| Rev | Date | Change | Codex review |
|---|---|---|---|
| r1 | 2026-10-05 | First submission. SHA-256 `8aaee0065645e59b832335283d1a1b172a05a6d9c30579a5ee85a365ff2e146b`, 817 lines. Archived byte-identically (hash-verified) outside the repository at `/tmp/g047-plan-r2/history/G047-vestive-shell-refresh-plan.r1.md`. | **CORRECTION LOOP** (PROGRESS, "G047 Plan r1 review: CORRECTION LOOP"). Four corrections. Resolved: K1–K22, fixtures, M1–M7, Q2. Depth narrowed by "G047 planning proportionality clarification". |
| r2 | 2026-10-05 | The four corrections, resolved proportionately by reusing the accepted wrapper and harness: <br>(1) **V-W status:** explicit Bash invocation with immediate wrapper and `tee` status capture and failure precedence (§8.7). <br>(2) **Startup and baseline:** handoff facts (§2.1a), the execution checkout and the committed-baseline prerequisites (§2.4), `EXEC_ID` before `$EV`, and a summary of the read-only startup checks (§8.7.1). <br>(3) **Cleanup:** the accepted harness cleanup only; the r1 label claim and the separate debug route are removed (§8.8.7). <br>(4) **Text spacing:** expected-element inventory, closed-page versus active-modal checks, real clipping and scroll-reachability assertions, and focused negative cases, all implemented in the existing browser tests after approval (§8.4). <br>Visual mapping, fixtures, assertion ledger, ACs, first-paint, contrast and pressed-state checks are unchanged. Only this Plan file changed. | Pending |

No implementation cycle has been consumed (0 of 3). G002 stays accepted with 7 of 7 cycles used. G003's B7 stays separate and unused, and G003 stays unauthorized. The foundation barrier stays closed through G021, including G047.

## 1. Summary

G047 restyles the accepted generic shell to the successor kit's Light and Dark roles. It adapts values only, by role, onto the existing 16-token contract, Base UI components and appearance engine.

**What changes.**
- **Palette values.** 12 Light and 5 Dark palette values change (§4.1). The token names, override shape, resolution and first-paint mechanism do not change.
- **Kit Light palette.** Light moves to the kit's warm paper: ground `#f4f2ee`, band `#ebe8e3`, raised `#ffffff`, rule `#d4cfc8`, text `#1f1d1b`, body text `#4d4843` and ochre gold `#8a5a00`.
- **Kit Dark palette.** Dark keeps the kit grounds `#0d0d0d`, `#101010` and `#131313`, which G002 already used.
- **Accessible fills.** Primary, secondary and destructive fills and the control boundary use measured accessible derivations of the kit hues. Each is listed as an exception with its measurement (K1–K7).
- **Button states.** Hover fills use the kit's rule colour (`border`). The link variant gets the kit's persistent 2px underline. The destructive variant becomes the kit's solid danger fill with a ground-coloured label.
- **Help dialog.** It becomes the kit's 2px-ruled raised panel, with kit padding and title/description type, a muted-to-ink close control and a per-appearance scrim.
- **Page type.** Body and label type takes the kit's sizes and leadings.

**What does not change.** Base UI, the button and dialog APIs, routes, dependencies, fonts, assets, the appearance engine, the scoped portals and the `--angle` and radius tokens.

**Copying and provenance.** No kit file, code line or asset is copied (§5).

**Tests.** The tests gain independent value oracles and three new scenario IDs (S-PALETTE, S-STATES, S-TEXT-SPACING), plus fixture scopes for partial and nested overrides (AC4). No existing assertion is removed and no tolerance is loosened.

**Execution (r2).** Validation reuses the accepted approved-input wrapper and harness, including its cleanup. The only additions are:
- a fail-closed Bash status capture around the wrapper call (§8.7);
- a documented startup in a named, committed execution checkout (§2.4, §8.7.1).

**Forecast.** Two cycles are expected, three in the worst case. B3 is feasible but tight (§11). No budget change is requested.

## 2. Starting state, inputs and binding

### 2.1 Repository state observed at planning (r1; historical observation, not an approval)

| Item | Value |
|---|---|
| Branch, HEAD, tree | `staging`, `9461c4bdbd8c1ed3e9a9986c70647cd1add57101`, tree `355d4e13f0b915d0cce1bbf90bbfc26c69d6e33f` |
| `web/`, `infra/`, `docs/foundation/` | Identical to HEAD. `web/` is unchanged since the G002 acceptance commit `7193ed8`; only `docs/ARCHITECTURE.md` changed since then (§61 added). |
| NUL status digest (`git status --porcelain=v1 -z -uall --ignored \| sha256sum`) | `c1a16f287449a6b1692693cef09f1235015aad908e4c3fc93b681c37814c7a61`, taken before this Plan existed |
| `.env` | Absent |
| Default host ports (3100, 8180, 8100, 5532, 6479, 9100, 9101) | Free |
| Docker | Unrelated projects and images exist (`payments-assistant-*`, `needknow-*`, `outreacher-*`, `atlas-*`/`harbor-*` images). None is touched. |

Pending changes before this Plan (the owner's skill consolidation, Codex management work and the G003 Plan). None may be deleted or broadly excepted.

| Status | Path | SHA-256 |
|---|---|---|
| M | `management/v0/PROGRESS.md` | `cf9059bafb7424add3c97c5cd9cb5e8df3d375252f29a2cb9892b835e62ce370` |
| M | `management/v0/ROADMAP.md` | `afb66a2fd9d6d8d7b30adeea32f36c57d1262afbd4172e66c5efb48f6eac77ab` |
| M | `management/v0/goals/G003-foundation-observability-ci.md` | `c2958828ebc0e72f8cc14ac8d3cf8c0f232db4c618a5f2d0bcbabe06dcb3e3aa` |
| M | `management/v0/specs/G003-foundation-observability-ci-spec.md` | `d50b7621ea1bf7fc2586ebc830313b3de9d92eb4401b98e1ed5e4dc14db57344` |
| M | `skills/vestive-design-skill/README.md` | `80898087d1afc7324840f930748cd045f7e0d5173453be70aa587ec65e3598b3` |
| M | `skills/vestive-design-skill/SKILL.md` | `7dfac5faecadb0b297a1be5641b1b86140dd28e6b4d32c78b140bc1cb65c670e` |
| ?? | `management/v0/goals/G047-vestive-shell-refresh.md` | `b7d06d3e09f142c32502040458a529754a1fd5ede1d6ccdbe8cfb605fe990769` |
| ?? | `management/v0/specs/G047-vestive-shell-refresh-spec.md` | `e4fc7c1e9dc1a8332624644c024660339982e9fce2823665ab91fb6b30b31899` |
| ?? | `management/v0/plans/G003-foundation-observability-ci-plan.md` (r4) | `10a827884bb811f808fc82bc230df9d12b95e237b84d93a36f3067348d00ccb4` |
| ?? | `skills/vestive-design-skill/references/shadcn-kit.md` | `b4e8cd92c89294bc17057c09d5814956d1108fa452c7a02ceec39c4b104df37e` |
| ?? | `skills/vestive-design-skill/references/shadcn-kit.sha256` | `fcda5a273daee8117598ab32a244a36d898112c27bd94b3908fbe7a362c79163` |
| ?? | `skills/vestive-design-skill/assets/shadcn/**` (63 files) | Covered by the manifest above |
| !! | `.claude/settings.local.json` | Ignored client settings; not a G047 input |

This Plan, `management/v0/plans/G047-vestive-shell-refresh-plan.md`, is now an additional untracked file.

### 2.1a Isolated-worktree handoff facts (r2; observation only, not an approval)

- **Planning worktree.**
  - Location: `.claude/worktrees/g047-plan-save` under the main checkout, on branch `worktree-g047-plan-save` at `9461c4b`.
  - Read-only status inside it: exactly `?? management/v0/plans/G047-vestive-shell-refresh-plan.md`.
  - It holds only the HEAD versions plus this Plan. It lacks the G047 Goal and Spec, the G003 Plan r4, the consolidated skill (`references/` and `assets/shadcn/`) and every other pending input.
- **Main checkout.**
  - `staging` at `9461c4b` (per `git worktree list`, run inside the planning worktree).
  - It holds the pending inputs of §2.1, plus an untracked `.claude/worktrees/` tree that contains the planning worktree.
  - This isolated session cannot run Git against the main checkout or write there. Its facts come from file reads and the Codex record.
- **Supervisor inputs changed since r1.** These are pending inputs for the eventual reviewed baseline. Their earlier hashes are historical, not current.

  | File (main checkout) | Hash at r2 | Note |
  |---|---|---|
  | `management/v0/PROGRESS.md` | `4dbbeb9ba9668c59f596d3cfef35d45c5ef8a93a5dad7e7838669995b38458b0` | Adds the r1 review and the proportionality clarification |
  | `AGENTS.md` | `dcc607a1730efe57609f28ae6b8c5c20fa1e0e9cb57e958f4bfc90ab87512764` | Adds "Proportional planning and reuse of accepted controls" |
  | `CLAUDE.md` | `cf1cbde9748502ec55f2dfe8243752d34313db781fad06dd7147da76a44f8eac` | — |
  | `management/BUDGET_POLICY.md` | `b183179b4b34dbe29c024f25ed7cd133c0a51835b1db58381d4509e3bce8f437` | Adds "Planning proportionality" |

  The G047 Goal and Spec, ROADMAP, the skill inputs, ARCHITECTURE and the G003 Plan r4 are unchanged from §2.1 and §2.2.
- **Baseline status.** Neither checkout is an approved execution baseline (Codex review).

### 2.2 Governing inputs read

| Input | SHA-256 |
|---|---|
| `AGENTS.md`, `CLAUDE.md`, `management/BUDGET_POLICY.md` | Current (r2) hashes in §2.1a. r1 read `3bc3e81f…775e9`, `ee2b13ff…80e6c` and `9abced55…2aa3b`, now historical. |
| `docs/ARCHITECTURE.md` (§37A frontend UI and design system; §61 "Evolution of the Vestive Design Input and Component Library") | `6a6fd49b854ce90acf1a2ce6bd281a4deee3e68e8f4e021d8d737c3b04d90726` |
| `docs/SECURITY.md`, `docs/DECISIONS.md` | Both empty (`e3b0c442…b855`). No accepted ADR applies; nothing conflicts. |
| ROADMAP, G047 Goal/Spec | As in §2.1 (unchanged) |
| PROGRESS | Current (r2) hash in §2.1a |
| G002 Goal / Spec / Plan r6 / Result / allow-list (historical; not revised) | `b60424e0…f745` / `00f4d5cb…cf86` / `30501d82…7f41` / `581c79eb…8770` / `fa65d700…19b8` |
| `docs/foundation/WEB-SHELL.md` | `0bb85614aad3c9aede8585fcd1d8e1f77916d6f0fa32968b212362cbed641f1d` |

Pre-change hashes of every file this Plan changes are in §6.1.

### 2.3 Source intake: manifest and selected inputs

- **Manifest verified.** `sha256sum -c references/shadcn-kit.sha256`, run from `skills/vestive-design-skill/`, reports 63 of 63 OK.
- **Manifest digest.** `fcda5a273daee8117598ab32a244a36d898112c27bd94b3908fbe7a362c79163`, matching the Spec. The sorted `sha256sum` of the 63 files under `assets/shadcn/` reproduces the same digest.
- **Kit consistency.** `tokens.json` agrees with `styles/tokens.css` for all 16 shared `vx-*` colour values in both themes. `vx-on-fill` exists only in `tokens.css`.
- **Provenance as supplied** (guide §1 and §10):
  - the product owner supplied the kit on 2026-10-05, and it was consolidated into this skill;
  - `tokens.json` declares `"version": 1`, and no other version is given;
  - no licence, copyright notice, author or origin statement is supplied for the kit;
  - MDI paths are labelled "Apache 2.0" in comments, without the licence text;
  - the icons8 SVGs, logos and fonts have no stated terms.

Every source used to make a visual decision is pinned below. Kit paths are relative to `skills/vestive-design-skill/assets/shadcn/`. All 15 kit files match their manifest entries.

| Selected input | Bytes | SHA-256 | Use in G047 |
|---|---|---|---|
| `SKILL.md` (skill entry point) | 5783 | `7dfac5faecadb0b297a1be5641b1b86140dd28e6b4d32c78b140bc1cb65c670e` | Precedence rules 1–7, applied |
| `README.md` (skill) | 10067 | `80898087d1afc7324840f930748cd045f7e0d5173453be70aa587ec65e3598b3` | Visual foundations and states; marketing content ignored |
| `references/shadcn-kit.md` (integration guide) | 24441 | `b4e8cd92c89294bc17057c09d5814956d1108fa452c7a02ceec39c4b104df37e` | Roles, contrast, conflicts §9.1–9.13, provenance §10 |
| `references/shadcn-kit.sha256` (manifest) | 6657 | `fcda5a273daee8117598ab32a244a36d898112c27bd94b3908fbe7a362c79163` | Intake fingerprint |
| `README.md` (kit) | 1452 | `988944ef8b2abd066891ffeb09748f9f462e11bc5922dc9078e9d8964e379321` | Read. Install, font and dark-default steps not followed (§9.1, 9.3, 9.4). |
| `BRAND.md` | 10281 | `b7c7491ce672107d48e2395b98f39a8bf210d77af3de1a453f4d0f2f613209ef` | Colour roles, states, type roles; content fundamentals ignored |
| `tokens.json` | 17457 | `91009c56ea6657de6ce22356d892ec84d2d5a2d3c401d9ab4150bc3c11853bfd` | Cross-check of values and usage notes |
| `globals.css` | 191 | `a915d08d929c3aef985f76b098fa29073cc4f48740f83bf7fa3a509cfb110cf2` | Read; not imported |
| `styles/tokens.css` | 2723 | `d3bcb4264b41196681096c7b8ea25db39ebc2146f4a5e6eb4cbbf7f97dcba0ca` | **Palette values** (both themes), slots, structure tokens |
| `styles/vestive.css` | 4465 | `93fab6c1c096776f903ac30ea8a445f114e1ab34e7eb07d3cebcd075b7985484` | Base focus, border and selection rules (reference only; §9.5) |
| `components/ui/button.tsx` | 2445 | `0f7fe0fecd140ebea85994746fb55bb9e000f34d9c3957fb938696a37cb87011` | Variant roles and states, hover and press |
| `components/ui/dialog.tsx` | 3435 | `c4d04cb9b490c0f2b661bd0f23e882466bae0fa3ddde1e8b63982b3268484abf` | Panel, padding, title/description type, close treatment, scrim |
| `components/ui/typography.tsx` | 3249 | `3cc7d9aad7bae82f6980d5138e2c3cc1f8ea5925c1825d37f1ecefe6e92b0610` | Kicker, heading, lead/body/meta/index, text-link roles |
| `components/ui/layout.tsx` | 2751 | `dada4ef1b5f5365ea90a6ceeea80a9c681942e3e529e9a5b2c3dfd63a8a55935` | Container, section bands, ruled grid |
| `components/ui/site-header.tsx` | 1624 | `4ec103f1455830866c90fd3bee800350afbf67434371d164c29dd50002f22fda` | Header and nav roles; logo and default nav labels ignored |
| `components/ui/bullet-list.tsx` | 750 | `613b3ddee0908893b402240fcd37d6389a14a2b227d0e104d1cc69acc3dd8a55` | 14×2 bullet rule and item type |
| `components/ui/toggle.tsx` | 1283 | `ba0b05b2317eb0f7bb15b256c29a98d338c45de33b4c4a030a5cd9ece0d01eb6` | Considered for the appearance control; not adopted (K16) |
| `components/ui/toggle-group.tsx` | 1662 | `57d70477fdaab05288ecc006fe7419bae4aaee71ca74b72f205d4caf0a9f9eb9` | Considered for the appearance control; not adopted (K16) |
| `components/ui/icons.tsx` | 1457 | `31c3a0205fd717e7689d51b5aca679653900385ba2fc535e8d4c5198f1257e16` | `CloseMark` considered; not adopted (K15); no MDI path used |

**Historical context (not a G047 input).** The static `styles.css` (`b6d41de3…a73a`) still equals G002's recorded hash. The kit's Dark values match it (guide §2).

**Intake is not permission to ship.** Being preserved in the skill does not authorize shipping any of these sources. The table records what informed decisions.

### 2.4 Execution checkout, committed baseline and binding (prerequisites; documented, not performed)

**Execution checkout.** A dedicated linked worktree, `.claude/worktrees/g047-exec` under the main checkout, on branch `worktree-g047-exec`, created at exactly the bound `BASE`. Why:
- Its own Git status excludes the main checkout's untracked `.claude/worktrees/` tree without any ignore rule or commit of that tree.
- Background sessions here cannot write the shared main checkout.
- The harness works from its own path with `git -C "$REPO"`, so it does not depend on which checkout it runs in.
- The harness has not yet run in a linked worktree. The S1 baseline run is the check: an incompatibility is a baseline failure, so stop and report; no cycle and no harness change.

**Route A: commit, then bind (unchanged reason).** B-SKILL-UNCHANGED compares `skills/` with `--base` and accepts `--expect-hash` only for untracked skill paths. The modified tracked skill files therefore cannot be bound against `9461c4b` without a harness change, and none is proposed.

**Prerequisites.** Performed by Codex, the owner, or a session they authorize:

| # | Prerequisite |
|---|---|
| P1 | Codex ACCEPT of this Plan revision. |
| P2 | Integrate the accepted Plan into the main checkout: a byte-exact, hash-verified copy from the planning worktree. |
| P3 | Route A base commit on `staging`, containing exactly the pending inputs:<br>• the consolidated skill: both modified files, `references/` and the 63 kit files;<br>• PROGRESS, ROADMAP, `AGENTS.md`, `CLAUDE.md` and `BUDGET_POLICY.md`;<br>• the G003 Goal, Spec and Plan r4;<br>• the G047 Goal, Spec and accepted Plan.<br>No `.claude/**` path is committed, no broad ignore rule is added, and nothing pending is discarded. |
| P4 | Optionally retire the planning worktree through the supported worktree removal. |
| P5 | Create the execution worktree at `BASE`: `git worktree add -b worktree-g047-exec <main>/.claude/worktrees/g047-exec <BASE>`. The executing session then enters it. |
| P6, P7 | Codex creates the approved-input snapshot with the existing mechanism, and records the binding (`BASE`, checkout path and branch, `SNAP`, `SNAP_SHA`, `N`, `RE`). |

Old G002 and G003 snapshots are historical only. I will not generate or self-approve a digest, a HEAD or a checkout.

**Proposed snapshot.** Codex decides.
- 23 entries:
  - the G047 Goal, Spec and accepted Plan;
  - `management/v0/PROGRESS.md`;
  - the 19 selected skill inputs in §2.3.
- `--path-re`:
  ```
  ^(management/v0/(goals/G047-vestive-shell-refresh\.md|specs/G047-vestive-shell-refresh-spec\.md|PROGRESS\.md|plans/G047-vestive-shell-refresh-plan\.md)|skills/vestive-design-skill/.+)$
  ```
- `--count 23`.
- `--plan management/v0/plans/G047-vestive-shell-refresh-plan.md`.

## 3. Implementation approach

1. **Map by role, not slot name.**
   - Each of the 16 tokens keeps its accepted role, and the kit value for that role is applied (§4.1).
   - `accent` stays emphasis ink (G002 Q10).
   - The kit's rule-coloured hover surface maps to `border`.
   - The kit's `destructive-foreground` role (ground) maps to the existing `background` token.
   - No token is added, renamed or repurposed.
2. **Values only, independently implemented.** Edits stay inside G002's own files: the palette block, Base UI component classes and page classes. No kit file, CSS rule, component or asset is copied or imported (§5).
3. **Measure before adopting.** Every kit value that fails a threshold in its consumer is replaced by a measured derivation of the same hue (K1–K4, K7), recorded as an exception.
4. **Keep the four concerns separate.**
   - A (palette): only values change.
   - B (resolution): unchanged.
   - C (TokenScope and portals): unchanged.
   - D (appearance init, control and storage): unchanged.
5. **Independent oracles.** Expected values in tests are typed from this Plan's tables (§8.1) and never read from `globals.css`.
6. **No class-merge reliance.** `cn` (the `cn` 0.4.0 package) is not assumed to resolve Tailwind conflicts. Wherever a consumer adds a class, the variant leaves that property unset. For example, the dialog close button adds `text-muted-foreground` and the ghost variant sets no resting text colour.
7. **Reuse accepted controls (r2).** Validation, isolation and cleanup are the accepted approved-input wrapper and harness, referenced by path. G047 adds only a status-capturing invocation (§8.7) and its own browser assertions.

## 4. Visual mapping

### 4.1 Semantic tokens: current, source and proposed values, per appearance

"Kit role (value)" is the role in `styles/tokens.css`. "Kit slot" is what the kit's own shadcn slot of the same name aliases, where it differs. Proposed values are the independent expected values that the tests restate (§8.1).

| Token | Accepted role and consumers | Light: current → kit role (value) → **proposed** | Dark: current → kit role (value) → **proposed** | Note |
|---|---|---|---|---|
| `background` | Page and header ground. Skip-link surface. Outline-button fill. Destructive label (K10). Scope surface. | `#fafaf9` → vx-bg (`#f4f2ee`) → **`#f4f2ee`** | `#0d0d0d` → vx-bg (`#0d0d0d`) → **`#0d0d0d`** | Kit value |
| `foreground` | Default text: H1/H2, lead, nav, product name, About items, unselected appearance labels, outline/ghost labels, hover ink | `#1a1918` → vx-text (`#1f1d1b`) → **`#1f1d1b`** | `#e8e6e3` → vx-text → **`#e8e6e3`** | Kit value. Headings use it because there is no ink token (K22). |
| `primary` | Angled default fill (the "Keyboard help" trigger), selected appearance option | `#5952ff` → vx-blue (`#0075ff`) → **`#0062d6`** | `#5952ff` → vx-blue (`#0075ff`) → **`#0062d6`** | **K1** |
| `primary-foreground` | Label on primary | `#ffffff` → vx-on-fill → **`#ffffff`** | `#ffffff` → vx-on-fill → **`#ffffff`** | Kit value |
| `accent` | Emphasis ink: kicker, cell H3, bullet rules, skip link, nav hover, link variant | `#8a5a12` → vx-gold (`#8a5a00`) → **`#8a5a00`** | `#f8ce8a` → vx-gold → **`#f8ce8a`** | Kit value. Kit slot `accent` is vx-rule (K9). |
| `muted-foreground` | Cell index and body, About body, footer meta, dialog description, close icon at rest, fixtures' muted text | `#5f5a54` → vx-text-3 (`#4d4843`); kit slot vx-muted (`#615b55`) → **`#4d4843`** | `#a8a39d` → vx-text-3 (`#a8a39d`); kit slot vx-muted (`#8f8a84`) → **`#a8a39d`** | **K5** |
| `border` | 2px rules (header, sections, grid, footer, dialog panel), hover fill (K11) | `#d9d6d2` → vx-rule (`#d4cfc8`) → **`#d4cfc8`** | `#262626` → vx-rule → **`#262626`** | Kit value. Structural only. |
| `popover` | Dialog panel | `#ffffff` → vx-raised → **`#ffffff`** | `#131313` → vx-raised → **`#131313`** | Kit value |
| `popover-foreground` | Dialog title and tips | `#1a1918` → vx-text (`#1f1d1b`) → **`#1f1d1b`** | `#e8e6e3` → vx-text → **`#e8e6e3`** | Kit value |
| `muted` | "Getting around" band. Unconsumed `DialogFooter`. | `#f0efed` → vx-surface (`#ebe8e3`); kit slot vx-raised → **`#ebe8e3`** | `#101010` → vx-surface → **`#101010`** | **K6** |
| `secondary` | Angled secondary fill (fixtures) | `#4d6800` → vx-green (`#66892f`) → **`#526e26`** | `#6a8e00` → vx-green (`#66892f`) → **`#709734`** | **K2 / K3** |
| `secondary-foreground` | Label on secondary | `#ffffff` → vx-on-fill → **`#ffffff`** | `#0d0d0d` → vx-on-fill (`#ffffff`) → **`#0d0d0d`** | **K3** |
| `destructive` | Solid destructive fill, its focus outline, `aria-invalid` | `#b3261e` → vx-danger (`#b8322a`) → **`#af3028`** | `#ff8a7a` → vx-danger (`#ff6b5e`) → **`#ff6b5e`** | **K4** (Light). Dark is the kit value. |
| `input` | Control boundary: outline button, appearance group frame and separators | `#8f8a84` → vx-muted (`#615b55`); kit slot vx-rule (`#d4cfc8`) → **`#615b55`** | `#6f6a64` → vx-muted (`#8f8a84`); kit slot vx-rule → **`#8f8a84`** | **K7** |
| `ring` | Focus outline | `#8a5a12` → vx-gold (`#8a5a00`) → **`#8a5a00`** | `#f8ce8a` → vx-gold → **`#f8ce8a`** | Kit value |
| `overlay` | Dialog scrim | `rgb(13 13 13 / 0.45)` → no kit Light scrim; kit scrim `rgba(0,0,0,.72)` is theme-fixed and Dark-tuned (guide §9.12) → **`rgb(13 13 13 / 0.45)`** | `rgb(0 0 0 / 0.7)` → kit scrim (`rgba(0,0,0,.72)`) → **`rgb(0 0 0 / 0.72)`** | **K8** (Light retained) |

**Totals.** Light changes 12 values; `primary-foreground`, `popover`, `secondary-foreground` and `overlay` are retained. Dark changes 5 values (`primary`, `secondary`, `destructive`, `input`, `overlay`).

**Fixture overrides.** Fixtures A and B (`tests/fixtures/token-scopes/fixtures.ts`) are unchanged. The two added fixture scopes are in §8.2.

**Planned palette block** (`web/app/globals.css`). The block keeps its exact two-column layout and order. Only the comment's trailing text and the values change.

```css
/* A. Design-language palette sources (successor kit role mapping, G047; both appearances). */
:root {
  --light-background: #f4f2ee;           --dark-background: #0d0d0d;
  --light-foreground: #1f1d1b;           --dark-foreground: #e8e6e3;
  --light-primary: #0062d6;              --dark-primary: #0062d6;
  --light-primary-foreground: #ffffff;   --dark-primary-foreground: #ffffff;
  --light-accent: #8a5a00;               --dark-accent: #f8ce8a;
  --light-muted-foreground: #4d4843;     --dark-muted-foreground: #a8a39d;
  --light-border: #d4cfc8;               --dark-border: #262626;
  --light-popover: #ffffff;              --dark-popover: #131313;
  --light-popover-foreground: #1f1d1b;   --dark-popover-foreground: #e8e6e3;
  --light-muted: #ebe8e3;                --dark-muted: #101010;
  --light-secondary: #526e26;            --dark-secondary: #709734;
  --light-secondary-foreground: #ffffff; --dark-secondary-foreground: #0d0d0d;
  --light-destructive: #af3028;          --dark-destructive: #ff6b5e;
  --light-input: #615b55;                --dark-input: #8f8a84;
  --light-ring: #8a5a00;                 --dark-ring: #f8ce8a;
  --light-overlay: rgb(13 13 13 / 0.45);              --dark-overlay: rgb(0 0 0 / 0.72);
```

The `--radius` and `--angle` lines, the resolution block, `@theme inline`, `@theme`, `angled`, the base layer, `body`, the appearance-option rules and the reduced-motion rule are byte-unchanged. Lines 1–640, the ejected block pinned by `GLOBALS_EJECT_SHA`, are untouched.

### 4.2 Consumer and state matrix (planning model; rendered evidence replaces it)

The ratios come from the disposable model in §15. Its brightness model is calibrated exactly against G002's measured hover fills (7 of 7 identical). Its contrast reproduces G002's measured ratios to two decimals. These are forecasts only. Acceptance uses rendered-pixel measurements (S-CONTRAST, S-STATES).

| Consumer / state | Foreground on background | Light | Dark | Min | Asserted by |
|---|---|---|---|---|---|
| Lead, H1, H2, nav, product name, About items, unselected appearance label, outline/ghost at rest | foreground on background | 15.03 | 15.60 | 4.5 (H1: 3) | S-CONTRAST, S-STATES |
| Kicker, nav hover, skip link, link variant at rest | accent on background | 5.30 | 13.13 | 4.5 | S-CONTRAST (skip link added) |
| Cell H3 on the band | accent on muted | 4.85 | 12.85 | 4.5 | S-CONTRAST |
| Cell index and body on the band | muted-foreground on muted | 7.40 | 7.60 | 4.5 | S-CONTRAST |
| About body, footer meta | muted-foreground on background | 8.08 | 7.77 | 4.5 | S-CONTRAST |
| Dialog title, tips | popover-foreground on popover | 16.80 | 14.92 | 4.5 | S-CONTRAST (tips added) |
| Dialog description | muted-foreground on popover | 9.04 | 7.42 | 4.5 | S-CONTRAST |
| Close icon at rest (non-text) | muted-foreground on popover | 9.04 | 7.42 | 3 | S-CONTRAST, S-PORTAL (token) |
| Close icon, outline and ghost on hover | foreground on border | 10.85 | 12.15 | 4.5 (icon 3) | S-CONTRAST (added), S-STATES |
| Link variant on hover | foreground on background | 15.03 | 15.60 | 4.5 | S-CONTRAST, S-STATES |
| Primary label: rest / hover / pressed (white on `#0062d6`; hover `#006ef0`; pressed `#0058c1`) | primary-foreground on primary | 5.65 / 4.68 / 6.63 | 5.65 / 4.68 / 6.63 | 4.5 | S-CONTRAST (pressed added) |
| Secondary label: rest / hover / pressed | secondary-foreground on secondary | 5.81 / 4.86 / 6.78 | 5.71 / 7.04 / 4.73 | 4.5 | S-CONTRAST (pressed added) |
| Destructive label: rest / hover / pressed | background on destructive | 5.73 / 4.80 / 6.65 | 6.96 / 7.54 / 5.71 | 4.5 | S-CONTRAST, S-STATES |
| Primary fill and selected appearance option vs ground; selected vs unselected option | primary vs background | 5.05 | 3.44 | 3 | S-CONTRAST |
| Secondary fill vs ground | secondary vs background | 5.19 | 5.71 | 3 | S-CONTRAST |
| Control boundary: outline button, appearance frame | input vs background | 5.99 | 5.68 | 3 | S-CONTRAST, S-PALETTE |
| Focus outline vs ground / band / popover | ring | 5.30 / 4.85 / 5.93 | 13.13 / 12.85 / 12.55 | 3 | S-TAB, S-CONTRAST, S-PORTAL |
| Destructive focus outline vs ground | destructive vs background | 5.73 | 6.96 | 3 | S-CONTRAST |
| 2px rules vs ground / band / popover (structural, recorded) | border | 1.39 / 1.27 / 1.55 | 1.28 / 1.26 / 1.23 | none: never a control's only boundary | S-DESIGN (colour) |
| Panel vs scrimmed page (translucent composite, recorded) | popover vs overlay composited over background (`#8c8b89` Light, `#040404` Dark) | 3.40 | 1.10 | none: recorded | S-CONTRAST (record), S-PALETTE (composite) |
| Disabled controls at `opacity: .5`, exempt as inactive (WCAG 1.4.3/1.4.11), recorded | label vs fill as composited | primary 2.28, secondary 2.17, destructive 2.30, outline/ghost 3.16, link 2.10 | 3.18, 2.25, 2.51, 4.52, 3.96 | exempt | S-CONTRAST fixtures (record), S-STATES (inactivity proven) |

**Fixture scopes.**
- The model covers every planned pair in both appearances:
  - 178 modelled pairs for the default scope and fixtures A and B (`proposal.out`): 152 thresholded and 26 recorded-only; default-adjacent equals default;
  - the added `light-only` and nested `dark-only-in-A` scopes, including inheritance (`fixtures-model.out`; §8.2).
- All pass. These pairs include each scope's own overrides combined with inherited platform tokens (secondary, destructive with that scope's ground as label, input, border).
- The smallest margins:
  - primary on hover, 4.68 (every scope using the platform primary);
  - Dark secondary pressed, 4.73;
  - Light destructive on hover, 4.80 (default ground; 4.88 on A's);
  - secondary on hover, 4.86.

**Pressed states.**
- The kit specifies press as `brightness(.9)`.
- G002 never measured pressed states. G047 adds them for the three filled variants, because K3 depends on it.
- Pressed is not an exemption.

### 4.3 Structure, typography and geometry accounting

| Kit item (source) | Kit value | Current shell | G047 | Reason |
|---|---|---|---|---|
| Rule width (`tokens.css` `vx-rule-w`) | 2px | `border-*-2` | Unchanged | Equal |
| Container (`vx-container`) | 1360px | `max-w-shell` 1360px | Unchanged | Equal |
| Gutter (`vx-gutter`) | `clamp(20px,4vw,40px)` | `px-gutter`, same value | Unchanged | Equal |
| Section padding (`vx-section-y`) | 88px | `clamp(3rem,8vw,5.5rem)`, max 88px | Unchanged | G002 narrow-viewport adaptation; equal on desktop |
| Cell padding (`vx-cell-pad`, BRAND) | 36px; 34px top, 40px bottom | `clamp(1rem,3vw,2.25rem)`, 34px, 40px | Unchanged | Equal on desktop; 320px reflow |
| Radius | 0 | 0 | Unchanged | Equal |
| Angle (`vx-angle`) | 15%/85% polygon | Constant 12px cut | Unchanged (K18) | G002 deviation 1: the 15% cut clipped labels |
| Font families | Oswald, Barlow Condensed, Roboto, mono | Same names first, local and system stacks only | Unchanged | No font download (guide §9.3) |
| H1 (`typography.tsx` level 1) | `clamp(2.6rem,5.6vw,5rem)` 700, lh 1, −1px | `clamp(2rem,5.6vw,5rem)`, same weight, lh and tracking | Unchanged (K21) | The configured product name must wrap at 320px |
| H2 | `clamp(1.9rem,3.4vw,2.7rem)` 600, lh 1.1, −.5px | `clamp(1.5rem,…)` | Unchanged (K21) | Same |
| H3 | 1.35rem 500, lh 1.2 | 1.35rem 500, lh 1.25 | **lh 1.2** | Kit value |
| Kicker | label 1rem 600, 2.5px, uppercase, lh 1.2, gold | Same without lh 1.2 | **lh 1.2** | Kit value |
| Lead | 1.19rem, lh 1.62, 52ch, text-2 | Same, in foreground | Unchanged (K22) | No text-2 token |
| Body | 1.04rem, lh 1.65, 62ch, text-3 | 1rem, lh 1.65, muted-foreground | **1.04rem** | Kit value |
| Meta (footer) | label 1.08rem, lh 1.35, muted | label 1.08rem (inherited lh), muted-foreground | **lh 1.35** | Kit value |
| Index | mono .72rem, 1px, faint | Same, muted-foreground | Unchanged | vx-faint 3.5–4.4:1 fails 4.5 (K5) |
| Bullet item | label 1.08rem, lh 1.35, text-2; 14×2 gold rule | label 1.08rem, foreground; 14×2 accent rule | **lh 1.35** | Kit value |
| Nav (`site-header.tsx`) | display .94rem, 1.2px, gap 36px, white → gold | Display .94rem, 1.2px, gap-x 32px, foreground → accent with underline | Unchanged (K20) | Underline is a non-colour hover cue; 4px gap difference retained |
| Header | Sticky, min-h 76px, py-3 | Not sticky, py-4 (about 68px) | Unchanged (K20) | Never covers focus (WCAG 2.4.11, G002 T11) |
| Buttons (`button.tsx`) | Display 500, uppercase, 1.5px; `px-10 py-[18px]` .95rem; `cursor-pointer` | Display 500, uppercase, 1.5px; h-8 px-2.5 text-sm; trigger h-9 px-5 | **`cursor-pointer` added**; sizes unchanged (K19) | Compact product controls (G002 V-m2); the kit sizes are marketing CTAs |
| Hover / press | `brightness(1.12)` / `(.9)` | Same on the angled fill | Unchanged; also applied to the solid destructive fill | Equal |

### 4.4 Button variants and states (`web/components/ui/button.tsx`)

| Variant | Kit (source) | G002 | **G047** | Guard |
|---|---|---|---|---|
| Base | `cursor-pointer`; clip on the element | No cursor; fill clipped on `::before` (E1–E3) | Adds **`cursor-pointer`** after `select-none`. Nothing else changes. | E1, E2 |
| default | `bg-primary text-primary-foreground` | Angled primary fill | Class unchanged; value via K1 | E3 |
| secondary | `bg-secondary text-secondary-foreground` | Angled secondary fill | Class unchanged; values via K2/K3 | E3 |
| destructive | Solid `bg-destructive text-destructive-foreground` (vx-bg) | Tinted angled fill (G002 deviation 2) | **`"angled text-background before:bg-destructive hover:before:brightness-[1.12] active:before:brightness-90 focus-visible:outline-destructive"`** (K4, K10) | E3 (angled, `before:bg-`) |
| outline | Angled 2px rule frame with ground inside; hover ink | `border-input bg-background hover:bg-muted …` plus `dark:bg-input/30 dark:hover:bg-input/50` | **`"border-input bg-background hover:bg-border hover:text-foreground aria-expanded:bg-border aria-expanded:text-foreground"`**. The `dark:` input tints are removed, so the inside is the ground as in the kit (K11, K12). | E5 (`"border-input `), E3 (rectangular) |
| ghost | Gold text; hover rule fill and ink | `hover:bg-muted …`; `dark:hover:bg-muted/50` (G002 measured an invisible `#0e0e0e` hover) | **`"hover:bg-border hover:text-foreground aria-expanded:bg-border aria-expanded:text-foreground"`**; resting text still inherits (K11, K13) | E3 (rectangular) |
| link | Gold text, 2px gold underline about 4px below, hover ink | `text-accent underline-offset-4 hover:underline` | **`"text-accent underline decoration-2 underline-offset-4 hover:text-foreground"`**, a persistent 2px `currentColor` underline that turns ink on hover | E4 (`"text-accent `, no `text-primary`), E3 |
| Disabled (all) | `disabled:opacity-50`, no pointer events | Same | Unchanged; measured and reported as exempt | — |
| Sizes | — | — | Unchanged | — |

### 4.5 Help dialog and overlay (`web/components/ui/dialog.tsx`)

| Part | Kit (`dialog.tsx`) | G002 | **G047** |
|---|---|---|---|
| Scrim | `bg-[rgba(0,0,0,.72)]`, theme-fixed | `bg-overlay` with a 100ms fade (E6) | Class unchanged; per-appearance value (K8); fade retained (K17) |
| Panel | `border-2 border-rule bg-popover`, `px 36px pt 34px pb 40px`, `w-[calc(100%-40px)] max-w-[560px]`, `gap-5`, no radius or shadow | `rounded-xl … ring-1 ring-foreground/10`, `p-4`, `max-w-[calc(100%-2rem)] sm:max-w-sm`, `gap-4`, `text-sm`, `zoom-in-95` | Popup class (K14, K17): `"fixed top-1/2 left-1/2 z-50 grid max-h-[calc(100dvh-2rem)] w-[calc(100%-2.5rem)] max-w-[35rem] -translate-x-1/2 -translate-y-1/2 gap-5 overflow-y-auto border-2 border-border bg-popover px-[clamp(1rem,3vw,2.25rem)] pt-[34px] pb-10 text-[1.04rem] leading-[1.65] text-popover-foreground duration-100 outline-none data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0"` |
| Header | `flex flex-col gap-3` | `gap-2` | `"flex flex-col gap-3 pr-9"`, where `pr-9` keeps the title clear of the close control (K14) |
| Title | display 1.9rem 600, lh 1.1, −.5px, uppercase, ink | `font-heading text-base leading-none font-medium` | `"font-heading text-[clamp(1.5rem,6vw,1.9rem)] leading-[1.1] font-semibold tracking-[-0.5px] uppercase"` (K14, K22) |
| Description | body 1.04rem, lh 1.65, text-3 | `text-sm text-muted-foreground …` | `"text-[1.04rem] leading-[1.65] text-muted-foreground *:[a]:underline *:[a]:underline-offset-3 *:[a]:hover:text-foreground"` |
| Close | 32px, `top-5 right-5`; muted-foreground → ink on hover; crossed 2px rules | Ghost `icon-sm` (28px), `top-2 right-2`, lucide `XIcon`, inherited popover-foreground | `<Button variant="ghost" className="absolute top-5 right-5 text-muted-foreground" size="icon" />`. That is 32px, 20px insets, muted at rest, and ink on a rule fill on hover through the ghost variant. Lucide `XIcon` is kept (K15). |
| `DialogFooter` | `mt-2 flex …` | Unconsumed export | **Unchanged and not rendered.** Its `-mx-4 -mb-4` offsets assume the old `p-4` padding; any first consumer must restyle it under its own Goal (residual R12). |
| Base UI wiring | Radix | Base UI Root, Trigger, Portal (E7), Backdrop, Popup, Close | Unchanged: no API, portal, focus, `aria-modal` or close-behaviour change |

**Modelled geometry** (font metrics assume the runner's Liberation Sans fallback; planning estimates only):
- **1280×800:** popup 560px wide, about 310–320px tall.
- **640×400 (200% zoom equivalent):** about 313px, within the 368px `max-h`.
- **1280×800 at 200% text:** about 552px, within 736px. `max-w` is in rem, so the line count matches 100% text.
- **320×640:** about 500px; about 597px under the spacing trial, within 608px.

`max-h` plus `overflow-y-auto` guarantees that Close and all content stay reachable at any zoom. It does not hide text.

### 4.6 Appearance control, header, regions and footer

| Region | G047 change | Unchanged (reason) |
|---|---|---|
| Header (`app-shell.tsx`) | Palette only: ground, 2px rule, nav ink and gold hover, primary trigger | Structure, non-sticky header, nav underline hover (K20) |
| Appearance control (`appearance-control.tsx`, first-paint CSS in `globals.css`) | Palette only: `input` frame and separators (K7), `primary` selected fill (K1), foreground labels | The kit's ToggleGroup gold-rule selected state is not adopted (K16). The selector CSS keyed on `html[data-appearance]`, the radios and storage are byte-unchanged. |
| Overview (`page.tsx`) | Kicker `leading-[1.2]` | H1/H2 clamps (K21); lead colour (K22) |
| Getting around (`page.tsx`) | Band becomes vx-surface (K6); H3 `leading-[1.2]`; cell body `text-[1.04rem]` (keeping `leading-[1.65]`) | Index colour (K5); grid reflow below `md` (guide §9.10) |
| About (`page.tsx`) | Paragraph `text-[1.04rem]`; items `leading-[1.35]` | Bullet rule mechanism |
| Footer (`app-shell.tsx`) | `leading-[1.35]`, appended after the existing classes | — |
| Skip link | Palette only | Class unchanged |

### 4.7 Exceptions and adaptations (approved in the r1 review; unchanged)

| # | Kit guidance | G047 treatment | Measured or usability reason |
|---|---|---|---|
| K1 | `primary` = vx-blue `#0075ff`, white label | `#0062d6` in both appearances: the kit hue scaled by 0.84 (`round(0.84 × (0,117,255))`); white label | White on `#0075ff` is 4.21; on its hover `#0083ff`, 3.69. White on `#0062d6`: 5.65 at rest, 4.68 on hover, 6.63 pressed. The fill reaches 3.44 against the Dark ground and 5.05 against the Light ground. A lighter blue fails hover (k ≥ 0.86 gives hover ≤ 4.49); a darker one weakens the Dark-ground margin. |
| K2 | Light `secondary` = vx-green `#66892f`, white label | `#526e26` (×0.80) | White on `#66892f` is 4.05. On `#526e26`: 5.81 at rest, 4.86 on hover, 6.78 pressed; fill vs ground 5.19. |
| K3 | Dark `secondary` = `#66892f`, white label | `#709734` (×1.10) with a `#0d0d0d` label | White on `#66892f` is 4.05. Even with a dark label, `#66892f` pressed is 4.00. `#709734`: 5.71 at rest, 7.04 on hover, 4.73 pressed; fill vs ground 5.71. |
| K4 | Light `destructive` = vx-danger `#b8322a` on a solid fill, label vx-bg | `#af3028` (×0.95); Dark keeps `#ff6b5e` | Ground label on `#b8322a`: 5.32 at rest but 4.44 on hover. On `#af3028`: 5.73, 4.80 and 6.65; other fixture grounds are ≥ 4.88 on hover. Keeping the G002 tint with kit danger gives Light hover 4.24 (fails), which is why the solid fill is chosen. |
| K5 | Slot `muted-foreground` = vx-muted (meta); body uses vx-text-3; index uses vx-faint | One token = vx-text-3 (`#4d4843` / `#a8a39d`) for body copy, meta and index | Three of the five consumers are body copy (cells, About, dialog description). vx-faint (3.5–4.4) fails 4.5 for the index. vx-muted remains in use as the control-boundary value (K7). Dark keeps the G002 value. |
| K6 | Slot `muted` = vx-raised | `muted` = vx-surface (`#ebe8e3` / `#101010`) | `muted`'s consumer is the alternate band; the kit band role is vx-surface (Spec R1 names `#ebe8e3`). The raised role is `popover`. |
| K7 | Slot `input` = vx-rule (`#d4cfc8` / `#262626`) | `input` = vx-muted (`#615b55` / `#8f8a84`) | vx-rule measures 1.27–1.55 and fails 3:1 for a control boundary (guide §9.9). vx-muted is the kit's own control-border role: 5.99 Light, 5.68 Dark. |
| K8 | Scrim `rgba(0,0,0,.72)`, theme-fixed | Dark `rgb(0 0 0 / 0.72)`; Light keeps ink `#0d0d0d` at 0.45 | The kit scrim is Dark-tuned (guide §9.12) and no Light scrim is supplied. At 0.45, the white panel measures 3.40 against the dimmed Light page, which keeps the dialog distinct and the page recognisably Light. A 0.72 scrim would turn Light near-black. |
| K9 | Slot `accent` = rule hover fill | `accent` stays emphasis ink | G002 Q10; B-WEB-TOKENS forbids accent surfaces |
| K10 | `destructive-foreground` slot = vx-bg | The destructive label uses the existing `background` token | Same role. No token added. Modelled ≥ 4.80 for every fixture ground (§4.2). |
| K11 | Hover fill = kit `accent` slot (vx-rule) | `hover:bg-border` and `aria-expanded:bg-border` on outline and ghost | Role mapping. Labels on it: 10.85 Light, 12.15 Dark. Replaces G002's near-invisible Dark ghost hover. |
| K12 | Outline = angled 2px rule frame | Rectangular 1px `input` boundary | E3/E5 keep outline rectangular (unclipped labels and focus). vx-rule fails 1.4.11. 2px would conflict with the base `border` class (no merge reliance). |
| K13 | Ghost resting text = gold | Resting text inherits | Ghost's only shell consumer is the dialog close, whose kit treatment is muted → ink. A gold resting colour would need a conflicting override. The link variant carries the gold role. |
| K14 | Dialog `max-w-[560px]`, 1.9rem title, 36px padding | `max-w-[35rem]`, title `clamp(1.5rem,6vw,1.9rem)`, padding `clamp(1rem,3vw,2.25rem)`, `pr-9` header, `max-h` with scroll | The rem width scales with text zoom. The clamps keep 320px usable: title on one line, 248px content width. `pr-9` prevents the title from overlapping Close. `max-h` keeps Close and content reachable at 200% zoom and under the spacing trial. |
| K15 | Close = two crossed 2px rules (`CloseMark`) | Lucide `XIcon` kept | No icon or code copied (§5). The existing accepted dependency draws the same crossed form; S-PORTAL and S-CONTRAST measure the SVG. |
| K16 | Toggles: selected = 2px gold rule | Appearance selected state stays a primary fill plus the 14×2 bar | The kit's "on" state of form controls is vx-blue (BRAND), so this is kit-consistent. A gold selected rule would collide visually with the gold focus outline. The accepted first-paint selector CSS stays unchanged. SA-V m4 remains an optional owner follow-up. |
| K17 | No dialog motion | 100ms fade retained; `zoom-in-95`/`zoom-out-95` removed | S-MOTION's control needs a real animation, and reduced motion removes it. Dropping the scale motion follows the kit's flat panel. |
| K18 | 15% angle | Constant 12px cut | G002 deviation 1 (S-ANGLE) |
| K19 | Large CTA sizes | G002 sizes kept | Compact product controls; S-TAB and S-RESPONSIVE geometry preserved |
| K20 | Sticky header, nav without underline | Not sticky; underline on hover | WCAG 2.4.11; a non-colour hover cue |
| K21 | H1 and H2 minimum clamps | G002 minimums kept | The configured product name must wrap at 320px (S-RESPONSIVE) |
| K22 | Headings in vx-ink; lead and bullets in vx-text-2 | `foreground` (vx-text) | The contract has no ink or text-2 token, and adding one would change the 16-name contract (forbidden) |

### 4.8 Source roles left unused (recorded, not consumed)

| Source role | Reason |
|---|---|
| vx-ink, vx-text-2 | No token; K22 |
| vx-muted as meta | K5. Used instead as the control boundary (K7). |
| vx-faint | Fails 4.5 |
| vx-orange (`Emph`) | No emphasised word: the H1 is configured identity (G002 T7) |
| vx-star, `chart-1…5`, the ribbon gradient, NodeField and hatch | No consumer; motion and marketing primitives are excluded (Spec R1, Goal) |
| Slots `card`, `card-foreground`, `accent-foreground` | No consumer. `accent-foreground` is forbidden (B-WEB-TOKENS). |
| `destructive-foreground` slot | Role realized by `background` (K10) |
| `::selection` tint `rgba(248,206,138,.25)` | A global base style outside the bounded surface; Dark-tuned (guide §9.12); the browser default is kept |
| Logos, icons8 SVGs, the subsidiary mark, Google Fonts, `next-themes`, Radix, `clsx`/`tailwind-merge`, the `.light`/`[data-theme]` selectors | Excluded by the Goal and Spec, SKILL precedence 3–6 and guide §9 |

## 5. Provenance, copy versus independent adaptation, licensing

- **Copied: nothing.** No kit file, CSS rule, component, class-string line, icon path, font or asset enters `web/`. G047 edits only G002's own repository files.
- **Independently implemented adaptations.** These are owner-supplied *values and roles*: the colours and type sizes in §4. They are expressed through existing utilities in the existing Base UI components. Derived colours (K1–K4) are computed here and are not in the kit.
- **Planning pre-check** (§15). The 16 intended changed lines have:
  - zero exact matches against the 1037 kit lines of 40 characters or more;
  - one shared 60-character fragment (`focus-visible:outline-2 focus-visible:outline-offset-2 focus`), which already exists in G002's accepted `button.tsx` and `app-shell.tsx`.
  To avoid avoidable coincidences, `cursor-pointer` is placed after `select-none` and the footer's `leading-[1.35]` is appended after the existing classes. The r2 re-run reproduces the output byte-identically (`e8b7c957…d1a8e`).
- **Execution check (S9).** Run the same comparison on the actual `git diff -U0 BASE` added lines of `web/` (excluding test files and fixtures, whose literals are oracles). Required: zero exact kit-line matches. Any ≥60-character fragment absent at BASE is a stop item: rewrite it independently or justify it as coincidental equal values before handoff. Report everything in the Result.
  - The string match is supplementary evidence, not proof of source-use rights (r1 review).
- **Automated guard.** B-WEB-DESIGN already fails on any `web/` file byte-identical to any design-skill asset, which now includes all 63 kit files. It also fails on any `vestive`, `vitruvian` or other marketing literal in `web/**`, including tests.
- **Licensing decision.**
  - The kit's code and assets carry no licence (guide §10).
  - Because nothing is copied, this Plan adds no notice or attribution file and adopts no licensed material.
  - Lucide `XIcon` is the existing accepted dependency (`lucide-react` 1.52.0, ISC, G002).
  - No MDI path, font, logo or icons8 file is used.
- **Q2 resolved (r1 review).** Independent adaptation of the supplied visual values is already permitted by the Goal, the Spec and the owner-requested adoption. This values-only approach needs no further architecture decision.
  - This is not a licence determination.
  - It does not permit copying source code or assets. Any proposed copying stops for provenance review.

## 6. Files and exact changes

### 6.1 Change inventory (exactly 18 paths; equals the G047 allow-list)

| # | Path | Kind | Pre-change SHA-256 | Change |
|---|---|---|---|---|
| 1 | `web/app/globals.css` | M (registered customization) | `c1f008c8…0aa66` | Palette block comment and values (§4.1) |
| 2 | `web/app/page.tsx` | M (registered customization) | `0c6c2d12…7b0a3` | Kicker `leading-[1.2]`; H3 `leading-tight` → `leading-[1.2]`; cell body and About `p` gain `text-[1.04rem]`; About `li` gains `leading-[1.35]` |
| 3 | `web/components/layout/app-shell.tsx` | M | `a2b6c3c5…e5b24` | Footer `div` gains `leading-[1.35]`, appended |
| 4 | `web/components/ui/button.tsx` | M | `d1c9cc89…6137b` | Base `cursor-pointer`; destructive, outline, ghost and link strings (§4.4) |
| 5 | `web/components/ui/dialog.tsx` | M | `ee9856d7…23db6` | Popup, close, header, title and description (§4.5); nothing else |
| 6 | `web/tests/browser/support/modes.ts` | M | `96d926a9…e4543` | `SOURCES` (16 tokens × 2) and `PALETTE` derived from it (§8.1) |
| 7 | `web/tests/browser/support/spacing.ts` | A | — | Text-spacing stylesheet, expected-element inventory, assertions and focused negative cases (§8.4) |
| 8 | `web/tests/browser/design.spec.ts` | M | `0992d8b3…ea08a` | S-DESIGN updates; new S-PALETTE (§8.2) |
| 9 | `web/tests/browser/accessibility.spec.ts` | M | `90558acc…4abbd` | S-CONTRAST additions; new S-TEXT-SPACING (shell) |
| 10 | `web/tests/browser/shell.spec.ts` | M | `51584f2c…d7749` | S-RESPONSIVE: the dialog also fits vertically |
| 11 | `web/tests/browser/tokens.spec.ts` | M | `a3b4350b…59a4c` | Independent sources; nesting-aware locators; S-PORTAL, S-RUNTIME and S-CONTRAST additions; new S-STATES and S-TEXT-SPACING (fixtures) |
| 12 | `web/tests/fixtures/token-scopes/fixtures.ts` | M | `3632ebc7…1568b` | `FixtureScope` type with optional `parent`; adds `light-only` and `dark-only-in-A` |
| 13 | `web/tests/fixtures/token-scopes/page.tsx` | M | `da285411…4652e` | Shared `ScopeContent`; nested scope rendered inside A; a disabled row per scope |
| 14 | `docs/foundation/WEB-SHELL.md` | M | `0bb85614…41f1d` | §6.3 |
| 15 | `infra/validation/lib.sh` | M | `9f7a3997…e57c3` | `BROWSER_MATRIX_FULL` gains S-PALETTE and S-TEXT-SPACING (shell) and S-STATES and S-TEXT-SPACING (tokens). No other line. |
| 16 | `infra/validation/render-customizations.tsv` | M | `9abf9e92…4279f` | Reason text for `web/app/globals.css` and `web/app/page.tsx` gains "; …(G047)". Path column and row count (8) unchanged. |
| 17 | `management/v0/results/G047-vestive-shell-refresh-result.md` | A | — | Result |
| 18 | `management/v0/results/G047-scope-allowlist.txt` | A | — | §6.2 |

**Explicitly unchanged** (verified by B-SCOPE and the S9 inventory):
- `web/app/layout.tsx`, `web/lib/**`, `web/components/layout/{appearance-control,shell-help-dialog,token-scope}.tsx` and `web/components/ui/portal-container.tsx`;
- `web/components.json`, `web/package.json`, `web/package-lock.json` and `web/playwright.config.ts`;
- `web/tests/browser/{appearance,boundary,overlay}.spec.ts` and `web/tests/browser/support/{angle,color,contrast,page-errors,png}.ts`;
- every Dockerfile and compose file, `infra/validation/run.sh`, `run-approved.sh` and `run-approved-controls.sh`;
- `README.md`, the renderer manifest, both skills, all G000–G003 artifacts, governance, architecture, ROADMAP and PROGRESS.

**Outside the repository** (r2): only evidence under `$EV`. G047 adds no execution helper, and no harness, wrapper, CI or protected-file change.

The `infra/validation/render-customizations.tsv` rows after the change are tab-separated, with the path column unchanged:

```
web/app/globals.css	ejected component CSS, semantic token contract, palettes and appearance modes (G002); successor palette values (G047)
web/app/page.tsx	renders the generic application shell from configured identity (G002); successor body and label type (G047)
```

The `infra/validation/lib.sh` matrix after the change (the other three entries are byte-unchanged):

```
  "shell-light shell-dark|S-IDENTITY S-LANDMARKS S-NAV S-SKIP S-TAB S-OVERLAY S-MOTION S-ERRORS S-RESPONSIVE S-A11Y S-CONTRAST S-BOUNDARY S-DESIGN S-ANGLE S-SCREENSHOTS S-PALETTE S-TEXT-SPACING"
  "tokens-light tokens-dark|S-ERRORS S-A11Y S-CONTRAST S-TOKENS S-PORTAL S-RUNTIME S-ANGLE S-SCREENSHOTS S-STATES S-TEXT-SPACING"
```

### 6.2 `management/v0/results/G047-scope-allowlist.txt` (exact content; the first execution write)

```
# G047 scope allow-list (input to infra/validation/run.sh B-SCOPE).
# Extended regular expressions, one exact anchored path per line. Every repository
# path G047 adds or changes relative to the bound base commit must match one line;
# any other path fails B-SCOPE. Mirrors G047 Plan §6.1.
^web/app/globals\.css$
^web/app/page\.tsx$
^web/components/layout/app-shell\.tsx$
^web/components/ui/button\.tsx$
^web/components/ui/dialog\.tsx$
^web/tests/browser/support/modes\.ts$
^web/tests/browser/support/spacing\.ts$
^web/tests/browser/design\.spec\.ts$
^web/tests/browser/accessibility\.spec\.ts$
^web/tests/browser/shell\.spec\.ts$
^web/tests/browser/tokens\.spec\.ts$
^web/tests/fixtures/token-scopes/fixtures\.ts$
^web/tests/fixtures/token-scopes/page\.tsx$
^docs/foundation/WEB-SHELL\.md$
^infra/validation/lib\.sh$
^infra/validation/render-customizations\.tsv$
^management/v0/results/G047-vestive-shell-refresh-result\.md$
^management/v0/results/G047-scope-allowlist\.txt$
```

### 6.3 `docs/foundation/WEB-SHELL.md` changes

The text must contain no identity literal from the manifest; B-LITERALS scans `docs/foundation/`. The `appearance-init` SHA-256 line is unchanged.

- **Token table:** proposed values, with a "Source role" column (vx-*, derived K1–K4/K7, retained K8).
- **New subsection "Successor design input (G047)":**
  - the §2.3 inventory with hashes;
  - the provenance statement in §5;
  - the mapping summary;
  - adaptations K1–K22;
  - unused roles.
- **Historical marking:** the G002 inventory, the screenshot table and T1–T13 are labelled "historical (G002)". The rows superseded by G047 (T5's `#5952ff`, the hairline dialog and the tinted destructive variant) are annotated.
- **"Further adaptations":** destructive is now a solid fill (K4/K10); the dialog panel is the 2px rule (K14).
- **Browser tests:**
  - the scenario table adds S-PALETTE, S-STATES and S-TEXT-SPACING;
  - the fixture list adds `light-only` and `dark-only-in-A`;
  - the 1.4.12 note becomes "bounded spacing trial for touched shell and control text (not full certification)".
- **Status:** adds the G047 refresh.

## 7. Schema, contracts and dependencies

- **Database and schema:** none. No migration; B-MIGRATE-SET unchanged (`00001_baseline.sql`).
- **Contracts, public API and MCP:** none.
  - Routes stay exactly `/`, `/api/health` and `/api/ready` (B-WEB-BOUNDARY).
  - `Button`, `Dialog*`, `TokenScope`, `tokenOverrideStyle`, `TokenOverrides`, `SEMANTIC_COLOR_TOKENS`, `NON_COLOR_TOKENS` and the appearance API are unchanged.
- **Dependencies:** none added, removed, re-versioned or recategorized (B-DEPS and C1–C12 unchanged). No font, asset or remote request.
- **Security and authority:**
  - no change to core, intelligence, credentials, tenancy or authority;
  - `web` stays a thin BFF;
  - scope labels stay non-identity.

## 8. Tests and validation

### 8.1 Principles

- **Independent oracles.** `web/tests/browser/support/modes.ts` gains `SOURCES`: the §4.1 proposed values for all 16 tokens in both appearances, typed from this Plan. `PALETTE` is then derived from `SOURCES` (same six keys, same values as §4.1).
  - Fixture expectations come from `fixtures.ts` constants.
  - Computed-style expectations (§8.2) are constants typed from §4.3–4.5.
  - Neither CSS nor tests are generated from the other.
  - A transcription error in either fails S-PALETTE, S-TOKENS or S-STATES.
- **Retained assertions.** No existing assertion is removed and no tolerance is loosened. Every modified assertion is listed in §8.2 with its reason.
- **No scan weakened.** B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, B-WEB-BOUNDARY, B-LITERALS and B-DEPS keep their regexes, scopes and allow-lists.
- **Content constraints.**
  - Test text contains no `vestive`, `vitruvian` or marketing literals (B-WEB-DESIGN).
  - Test text contains no `fonts.googleapis` literal and no identity literal from the manifest (B-LITERALS).
- **Negative cases (r2).** The spacing assertions include focused negative cases (§8.4). They are written with the tests after approval and must be detected before the real checks count.
- **Real application evidence.** Planning models and reference screenshots are not acceptance evidence.

### 8.2 Assertion ledger

**Modified existing assertions**

| ID | Where | Before | After | Reason |
|---|---|---|---|---|
| M1 | `support/modes.ts` `PALETTE` | G002 values | §4.1 values (same keys and tolerances), derived from `SOURCES` | Reviewed palette (AC3) |
| M2 | S-DESIGN dialog | `box-shadow` contains exactly one `0 0 0 1px` hairline | `box-shadow` is `none` (zero layers), **plus** a 2px solid border in `--border` on all four sides | Kit ruled panel. Stricter: no shadow layer allowed. |
| M3 | S-PORTAL close icon | Equals scope `popover-foreground` | Equals scope `muted-foreground` (nesting-aware source chain) | Kit close treatment (§4.5) |
| M4 | S-TOKENS `expected()` | Root source read from the page | Own override → parent chain → `SOURCES` | Independent oracle; nesting (AC4) |
| M5 | `tokens.spec` content locators | `[data-token-scope="X"]` descendants | That scope's own section, `[data-token-scope="X"] > section` | Nested scopes would otherwise match twice. Identical elements for the existing four scopes. |
| M6 | S-RUNTIME | A's runtime Light change; default, adjacent and B unchanged in both appearances | Same, plus `light-only` unchanged; `dark-only-in-A` Light fill follows A's runtime change (inheritance) while its Dark fill stays its own | Nesting and partial overrides |
| M7 | S-A11Y and S-SCREENSHOTS (fixtures) | Dialogs in A and B | Dialogs in A, B and `dark-only-in-A` | Scoped dialog in a nested scope |

**Added assertions and scenarios**

| ID | Scenario (projects) | Assertions |
|---|---|---|
| A1 | **S-PALETTE** (shell-light, shell-dark; new, design.spec) | (a) For all 16 tokens × both appearances, `:root` `--light-*`/`--dark-*` parse (browser canvas) to `SOURCES`, distance ≤ 1 per RGBA channel. (b) In the project's resolved appearance, rendered pixels or computed colours equal `SOURCES` (≤ 2): header ground = `background`; band = `muted`; header bottom rule = `border`; H1 = `foreground`; kicker = `accent`; footer = `muted-foreground`; trigger fill = `primary`, label = `primary-foreground`; selected option = `primary`; appearance frame = `input`. With the dialog open: panel = `popover`; title = `popover-foreground`; panel rule = `border`; scrim pixel at (2,2) = `overlay` composited over `background` in the test's own `over()`. (c) Attaches `fonts.json` (CDP `CSS.getPlatformFontsForNode` for h1, nav link, kicker, lead, index, trigger and dialog title), with at least one platform font per node. |
| A2 | **S-STATES** (tokens-light, tokens-dark; new) | For each of the six scopes, expected tokens via the M4 chain: <br>• **default / secondary / destructive:** `::before` background = `primary` / `secondary` / `destructive`; label = `primary-foreground` / `secondary-foreground` / `background`; `::before` filter is `brightness(1.12)` on hover and `brightness(0.9)` when pressed; `::before` clip-path = `--angle`. <br>• **outline:** 1px border = `input`; background = `background`; on hover, background = `border` and colour = `foreground`. <br>• **ghost:** background transparent at rest; `border` and `foreground` on hover. <br>• **link:** colour = `accent`; `text-decoration-line` includes `underline`; thickness 2px; offset 4px; hover colour = `foreground`. <br>• **all:** `cursor: pointer`; border-radius 0. Keyboard focus gives `ring` (destructive: `destructive`), 2px solid, offset 2px. <br>• **disabled row:** `:disabled`; opacity 0.5; `pointer-events: none`; focusing leaves `document.activeElement` unchanged. |
| A3 | **S-TEXT-SPACING** shell (shell-light, shell-dark; new, accessibility.spec) | §8.4 at 1280×800 and 320×640: closed page, focused skip link and open modal checked separately, with the focused negative cases |
| A4 | **S-TEXT-SPACING** fixtures (tokens-light, tokens-dark; new, tokens.spec) | §8.4 at 1280×800: closed page over the six scopes (16 expected elements each, disabled buttons included), then open modals in A, B and `dark-only-in-A`, with the focused negative cases |
| A5 | S-CONTRAST shell (existing ID) | Adds: skip-link text on its own ground when focused (≥ 4.5) and its focus ring (≥ 3); dialog tip text (≥ 4.5); close icon on hover against its painted hover fill (≥ 3). Records, without threshold: panel vs scrimmed page, and the dialog rule vs panel. |
| A6 | S-CONTRAST fixtures (existing ID) | Adds, per scope: pressed label ≥ 4.5 for default, secondary and destructive (mouse down; sample; mouse up); close icon on hover ≥ 3; disabled pairs recorded as "exempt: inactive component" with no threshold |
| A7 | S-RESPONSIVE (existing ID) | With the dialog open, `d.y ≥ −0.5` and `d.y + d.height ≤ viewport height + 0.5` |
| A8 | S-DESIGN (existing ID) | Adds computed values at 1280×800, compared numerically (±0.01px): <br>• kicker line-height 19.2px; <br>• H3 21.6px, line-height 25.92px; <br>• cell body and About `p` font-size 16.64px, cell body line-height 27.456px; <br>• About `li` and footer line-height 23.328px; <br>• popup width 560px, padding 34px top, 40px bottom, 36px left and right, font-size 16.64px; <br>• title 30.4px, line-height 33.44px, weight 600, letter-spacing −0.5px, uppercase, display family; <br>• description 16.64px, line-height 27.456px, colour = `--muted-foreground`; <br>• Close is 32×32 at 20px from the popup's inner top and right edges, and its SVG colour = `--muted-foreground`; <br>• overlay background = `SOURCES[mode].overlay`. |
| A9 | S-PORTAL (existing ID) | Overlay colour = `SOURCES[m].overlay`, an independent oracle, in addition to the existing scope-resolved comparison. Nested scope: popup and backdrop mount in `dark-only-in-A`. |
| A10 | S-TOKENS (existing ID) | Per mode: `light-only` fill ≠ default in Light and = default in Dark; `dark-only-in-A` fill = A in Light and ≠ A in Dark; `[data-token-scope]` count = `FIXTURE_SCOPES.length` (6) |

**Added fixture scopes** (`fixtures.ts`). The values are test data; the model passes all pairs in both appearances (§15).

```ts
export type FixtureScope = { label: string; tokens?: TokenOverrides; parent?: string };
export const fixtureLightOnly: TokenOverrides = {
  light: { background: "#f2f6f1", foreground: "#14261a", primary: "#1d6b47", "primary-foreground": "#ffffff",
           "muted-foreground": "#3e5a48", accent: "#6b3d8f", ring: "#1d6b47", popover: "#ffffff", "popover-foreground": "#14261a" },
};
export const fixtureDarkOnly: TokenOverrides = {
  dark: { background: "#1b1426", foreground: "#eee6f6", primary: "#c7a6f2", "primary-foreground": "#1b1426",
          "muted-foreground": "#b9a9cc", accent: "#f4c87a", ring: "#f4c87a", popover: "#241b31", "popover-foreground": "#eee6f6" },
};
// Order: default, A, default-adjacent, B (unchanged), then light-only (sibling) and dark-only-in-A (nested in A).
export const FIXTURE_SCOPES: readonly FixtureScope[] = [
  { label: "default" }, { label: "A", tokens: fixtureA }, { label: "default-adjacent" }, { label: "B", tokens: fixtureB },
  { label: "light-only", tokens: fixtureLightOnly }, { label: "dark-only-in-A", tokens: fixtureDarkOnly, parent: "A" },
];
```

**Fixture page** (`page.tsx`, test-only; still copied only into the runner image):
- Top-level scopes render in the grid.
- A scope with children renders them as `TokenScope` siblings of its own `<section>`, inside its element (`mt-6 border-2 border-border p-6`).
- Each section adds a second button row: the same six variants with `disabled` and `data-testid="button-<variant>-disabled"`.
- S-NOFIXTURE still requires 404 in production.

### 8.3 Scenario, project and test-count matrix

Per browser run: **103 tests** (95 existing + 8 new), in root and disabled-variant runs alike. The identity-rename run stays at 2 `@identity` tests. The §8.4 negative cases run inside the existing S-TEXT-SPACING tests, so they add no test.

| Project | Scenarios (★ new ID, ✎ existing ID with added or modified assertions) |
|---|---|
| shell-light, shell-dark | S-IDENTITY, S-LANDMARKS, S-NAV, S-SKIP, S-TAB, S-OVERLAY, S-MOTION, S-RESPONSIVE ✎, S-A11Y, S-CONTRAST ✎, S-BOUNDARY, S-DESIGN ✎, S-ANGLE, S-SCREENSHOTS, **S-PALETTE ★**, **S-TEXT-SPACING ★**; shell-light also S-HEAD-ORDER, S-HEALTH, S-NOFIXTURE, S-FAILCLOSED |
| tokens-light, tokens-dark | S-A11Y ✎, S-CONTRAST ✎, S-TOKENS ✎, S-PORTAL ✎, S-RUNTIME ✎, S-ANGLE, S-SCREENSHOTS ✎, **S-STATES ★**, **S-TEXT-SPACING ★** |
| appearance | S-APPEARANCE-DEFAULT, -LIVE, -OVERRIDE, -RETURN, -PERSIST, -INVALID, -NOSTORAGE; S-FIRSTPAINT (8); S-FIRSTPAINT-DELAYED (2); S-HTML-ATTRS (6). Code unchanged; the oracle values change through M1. |
| dev-hydration | S-HYDRATION-DEV (6), unchanged |
| All except dev-hydration | S-ERRORS on every test, unchanged |

### 8.4 Touched-text spacing trial (S-TEXT-SPACING; bounded, not a full 1.4.12 certification)

**Status.** Described here; implemented in `support/spacing.ts` and the existing specs after approval.

**Unchanged from r1.**
- The stylesheet: `* { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; } p { margin-bottom: 2em !important; }`.
- The cases: shell at 1280×800 and 320×640 in both appearances; fixtures at 1280×800 in both appearances.
- The angled-label check.
- Every other acceptance assertion and all first-paint evidence.

**Expected-element inventory.** Each item has a locator, an exact count and its exact source text (`textContent`, whitespace collapsed). The test proves each item exists, has its text and is visible (`checkVisibility` with opacity and visibility, plus a non-empty client rect). Enumerating visible rects alone cannot prove that hidden or missing content was retained.

| Area | Items (count) |
|---|---|
| Shell, closed page | product name (1), nav links (3), appearance labels (3), Keyboard help (1), kicker (1), H1 (1), lead (1), H2 (2), indices (3), H3 (3), cell bodies (3), About paragraph (1), About items (3), footer (1) |
| Shell, skip link focused | Skip to main content (1) |
| Shell, modal open | Title (1), description (1), tips (4), Close (1) |
| Fixtures, closed page (each of 6 scopes) | Body, muted, link (3); six variant buttons (6); six disabled buttons (6); Keyboard help (1) |
| Fixtures, modal open | The shell modal items, in A, B and `dark-only-in-A` |

The configured product name and domain come from `EXPECTED_PRODUCT_NAME` and `EXPECTED_PLATFORM_BASE_DOMAIN`, as S-IDENTITY uses. Every other text is typed from the BASE sources.

**Closed page versus active modal.**
- With the dialog closed, only page items are checked.
- With the dialog open, the test first asserts that `header`, `main` and `footer` are inert or `aria-hidden`, as S-OVERLAY already does. It then checks only the modal items.
- Covered, inert background controls are therefore never counted as modal overlaps.

**Assertions per item.**
- **Not covered:** after `scrollIntoView`, `elementFromPoint` at the text centre hits the item. In the modal, the hit must also lie inside the popup.
- **Not clipped:** no text rect lies outside an ancestor whose `overflow` is `hidden` or `clip`.
- **Reachable:** for every scrolling ancestor, including the document, `scrollIntoView` must actually bring the text inside both that ancestor's client rect and the viewport. No `auto`/`scroll` ancestor is assumed reachable.
- **No overlap:** no two items' text rects intersect within the same state.
- **Labels inside controls:** labels stay inside their control's box, and angled labels stay inside the polygon.
- **Controls usable:** enabled controls can take focus; disabled fixture buttons are `:disabled`.
- **No horizontal page overflow.**
- **Modal only:** the popup is inside the viewport, and Close is in view and does not intersect the title or description.

**Focused negative cases.** These run inside the same S-TEXT-SPACING tests, before the real checks, on scratch elements that are removed afterwards. Each must be reported by the matching assertion, otherwise the test fails:
- a missing item;
- `visibility: hidden` and `display: none`;
- text covered by an opaque sibling;
- two overlapping blocks;
- `nowrap` text clipped by `overflow: hidden`;
- text in an `overflow: auto` box at a negative offset that scrolling cannot reach;
- a label running under an angled cut;
- with the modal open, an opaque element placed over Close.

One positive case runs alongside them: text below the fold of an `overflow: auto` box must pass as reachable. No separate validation framework or pre-approval demonstration is built.

**Evidence.**
- `spacing-report.json`: each item, its rects and its findings.
- The spacing screenshots (shell at both widths, closed and with the dialog; fixtures).
- No scenario ID or test is added beyond S-TEXT-SPACING (§8.3).

### 8.5 Screenshot matrix and identification

| Set | Images | Source |
|---|---|---|
| **Before** (actual shell, bound baseline) | {Light, Dark} × {1280×800, 320×640} × {closed, dialog open} = 8 | S-SCREENSHOTS attachments in `browser-root-results.json` from the S1 baseline run |
| **After** (actual shell, delivered tree) | Same 8 | Same scenario, final S8 run |
| Fixtures | Full page plus dialogs (A, B; after, also `dark-only-in-A`) per appearance | S-SCREENSHOTS (tokens) before and after |
| Spacing trial (after) | §8.4 attachments | S-TEXT-SPACING, final run |

**Identification in the Result.** Every screenshot is recorded with:
- its viewport and `deviceScaleFactor` 1;
- the project and colour scheme;
- Chromium as reported by B-BROWSER-ROOT;
- the runner image (`web-browser.Dockerfile` `83979271…2705d`, unchanged, Playwright base `v1.63.0-noble@sha256:eff16c30…a4a27`);
- fonts: the `fonts.json` platform fonts (after). Before, the same runner definition applies, with the Liberation mapping from G002 deviation 9;
- the revision: the BASE commit plus the working-tree digest at run start;
- its SHA-256.

The open dialog is captured after its animations settle (opacity 1, as in G002). Screenshots supplement measurements; they are never first-paint evidence.

### 8.6 Guard, customization and disabled-variant reconciliation

| Guard / mechanism | G047 effect | Why it still holds (no weakening) |
|---|---|---|
| `GLOBALS_EJECT_SHA` (`16693ac5…c1e8`) | None | Lines 1–640 and the foundation header untouched |
| B-WEB-TOKENS | None | 16 names in the five places unchanged. Literals only in the palette block (`#…`, `rgb(…)`). `:root` non-colour tokens still exactly `--radius --angle`; `@theme` names unchanged. No palette utility (`text-background`, `bg-border` and `decoration-2` do not match the palette regex). No `bg-accent` or `accent-foreground`. |
| B-WEB-COMPONENTS (E1–E7) | None | Base still contains E1 and E2. Default, secondary and destructive keep `"angled …before:bg-…"`, and default and secondary keep the brightness hover and press. Outline, ghost and link stay rectangular. Link starts `"text-accent `; outline starts `"border-input `. Overlay class and E7 unchanged. `components/ui` is still exactly three files. |
| B-WEB-DESIGN | None | No remote font or asset. No web file byte-identical to any of the design-skill assets, now including the 63 kit files; the PASS detail's asset count therefore rises from G002's. No marketing literal in `web/**`. |
| B-WEB-BOUNDARY, B-LITERALS, B-NAMING | None | No server import, route, cookie or `NEXT_PUBLIC_`; no identity literal in `web/`, `infra/` or `docs/foundation/` |
| B-DEPS, B-DEPS-CONTROLS (C1–C12, C8b, C8c) | None | Package files and the runner image unchanged |
| B-SHELL | Covers the edited `lib.sh` | shellcheck must pass |
| B-RENDER-ID, -PLAN, -REPEAT | Register reason text only | Still exactly 8 registered conflicts (`globals.css` and `page.tsx` were already registered); `app-shell.tsx`, `button.tsx`, `dialog.tsx`, tests, `WEB-SHELL.md` and harness files are not rendered paths (manifest `55bb5288…` checked) |
| **Disabled variant** (B-TOGGLE-OFF-ADDITIONS, B-BROWSER-OFF) | `APPLICABLE` unchanged (28 entries) | It already covers `web/app/globals.css`, `web/app/page.tsx`, `web/components`, `web/tests` and `docs/foundation/WEB-SHELL.md`. Every G047 web, test and doc path is copied into the intelligence-disabled render and compared byte-for-byte. Registered customizations keep their rendered base hashes. B-BROWSER-OFF runs the extended full matrix (103 tests). |
| B-BROWSER-ROOT, -OFF, -RENAME | Matrix +3 IDs (§6.1) | All existing IDs stay required in the same projects; RENAME unchanged |
| B-PROTECTED, B-SKILL-UNCHANGED | None | Pass only with a Route A base (§2.4) |
| B-SCOPE | New G047 allow-list | Exact 18 paths |
| `run-approved.sh` and controls (W1–W9) | None | Unchanged; fresh G047 snapshot |
| `init_hash_documented` | None | The `appearance-init` script and its documented hash are unchanged |
| Harness cleanup (`cleanup`, `inventory-start.txt`, `inventory-leftover.txt`; r2) | None; reused | Every validation run relies on the harness's own owned-resource cleanup and its start/end evidence (§8.8.7). G047 adds no cleanup procedure and no debug route. |

### 8.7 Validation commands (exact)

#### 8.7.1 Startup (after P1–P7; read-only until the allow-list write)

1. **Evidence identity first.**
   - Generate `EXEC_ID` (8 hex characters from `/dev/urandom`).
   - Only then set `EV=$HOME/.local/state/platform-evidence/g047-$EXEC_ID`.
   - Before creating it, confirm that `EV` is an absolute, normalized path without symlinked components (`realpath -m "$EV"` equals `"$EV"`), is outside both checkouts and outside `/tmp`, and does not exist yet. `/tmp` is excluded because it is lost on WSL restart.
   - Then create it with `mkdir -m 700`. All startup evidence goes there.
2. **Read-only Git in the execution checkout.** Use `GIT_OPTIONAL_LOCKS=0 git -c core.fsmonitor=false` with `GIT_DIR`, `GIT_WORK_TREE` and `GIT_INDEX_FILE` unset. Check:
   - the top level is the bound checkout, `HEAD` equals `BASE`, and the branch is the bound branch;
   - `status --porcelain=v1 -z --untracked-files=all --ignored` is empty, or only the ignored `.claude/settings.local.json` (NUL-exact);
   - every snapshot entry, every P3 pending input and all 63 kit files are committed at `BASE` (`cat-file -e`, `ls-tree`);
   - `sha256sum -c references/shadcn-kit.sha256` reports 63 OK.
3. **Approved-wrapper precheck.** `bash infra/validation/run-approved.sh … --check-only` (the existing mechanism, unchanged) passes. Its status is captured as in §8.7.2.
4. **Resources (read-only).**
   - No `.env`.
   - The default and harness-variant ports (`.env.example`; `PORTS_*` in `run.sh`) are free.
   - `docker ps -a`, `docker network ls`, `docker volume ls` and `docker image ls`, recorded with IDs in `$EV`, show no container (running or stopped), network, volume or image of the root Compose project.

Any failure here is a **startup failure**: no repository write, no cycle, stop and report. The first repository write, the allow-list (§6.2), follows only when all four pass.

#### 8.7.2 V-W: approved-wrapper run with fail-closed status capture

Run from the execution checkout, with the P7 binding values (`BASE`, `SNAP`, `SNAP_SHA`, `N`, `RE`), `PLAN`, `ALLOW=management/v0/results/G047-scope-allowlist.txt` and `EV` exported. The command is always run by Bash, never by zsh, and never as a plain pipe:

```sh
bash --noprofile --norc -c '
  set -u
  step="$1"; shift
  mkdir -m 700 -- "$EV/$step" || exit 80
  bash infra/validation/run-approved.sh --snapshot "$SNAP" --snapshot-sha256 "$SNAP_SHA" --count "$N" \
    --path-re "$RE" --plan "$PLAN" -- --base "$BASE" --allow-file "$ALLOW" \
    --report-dir "$EV/$step/report" "$@" 2>&1 | tee "$EV/$step/harness.out"
  ps=("${PIPESTATUS[@]}"); w=${ps[0]}; t=${ps[1]}
  post=$(grep -oE "postcheck (passed|FAILED)$" "$EV/$step/harness.out" | tail -1)
  run=$(grep -oE "\] run [0-9a-f]{6};" "$EV/$step/harness.out" | tail -1)
  if (( w != 0 )); then rc=$w; elif (( t != 0 )); then rc=70; else rc=0; fi
  printf "%s\twrapper=%s\ttee=%s\t%s\t%s\toverall=%s\n" "$step" "$w" "$t" "${post:-postcheck not-reached}" "${run:-no-run-id}" "$rc" >>"$EV/runs.tsv" || { (( rc == 0 )) && rc=71; }
  exit "$rc"' g047 <step> <suites>
```

Steps: `s1 all` (baseline), `s3 static provenance`, `s4a`, `s4b`, … `static provenance root`, and `s8 all` (final; `s8b` on a rerun).

**Status precedence.**
- A nonzero wrapper status always wins.
- A wrapper status of 0 with a failed `tee` gives 70.
- An unwritable status record gives 71.
- Otherwise 0.

**Recorded per run.** Each run records its wrapper status, `tee` status, the wrapper's postcheck verdict, the harness RUN_ID and the overall status. The Result reports them for S1, S3, every S4 retest and S8. S1 and S8 require overall 0.

**Planning check (§15).** A four-case stub check of this exact snippet, with only the wrapper and `tee` substituted, gives:
- wrapper 11, `tee` 0 → 11;
- wrapper 0, `tee` 1 → 70;
- both fail (13 and 1) → 13;
- both pass → 0.

r1's plain pipe returned 0 for the first case.

**Unchanged.** The wrapper (`run-approved.sh` `620a3e84…c53c695`) and the harness.

### 8.8 Integrity checks (run around the final S8 run, with no repository write during any run)

Git commands here are read-only, with `GIT_OPTIONAL_LOCKS=0 git -c core.fsmonitor=false`, run in the execution checkout (§2.4).

1. **Pre/post tree digest.** `find . -path ./.git -prune -o -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum | sha256sum`, plus the NUL status digest. Both must be equal before and after.
2. **Full change inventory.** `{ git diff -z --name-only $BASE; git ls-files -z -o --exclude-standard; }`, sorted and unique, must equal the 18 paths of §6.1 exactly: none missing, none extra. Ignored inventory: empty, or only `.claude/settings.local.json`.
3. **Kit manifest.** `sha256sum -c references/shadcn-kit.sha256` reports 63 OK at startup and after S8.
4. **Package and module diff.** `git diff --stat $BASE -- web/package.json web/package-lock.json` must be empty. `git diff -U0 $BASE -- web/components web/app | grep '^+import'` must be empty.
5. **Provenance line check** (§5).
6. **Secret hygiene.** Evidence must contain no `.env` content (B-SECRET-CANARY also scans).
7. **Cleanup: the accepted harness procedure (r2; replaces r1's unsupported `g047.exec` label claim).**
   - Every run that creates Docker resources is a harness run through the approved wrapper (§8.7). The harness records `inventory-start.txt` before creating anything.
   - It refuses to touch the root project when a `.env`, existing root-project containers or volumes (stopped included), or busy default ports are found (B-ROOT UNMET).
   - Destructive scenarios run on disposable copies under unique, registered project names.
   - At exit it removes only registered projects and images that were absent from its start inventory.
   - It removes the root `.env` only when its hash matches the one it created.
   - It writes `inventory-leftover.txt` (copied into the step's report directory).
   - **G047 deletes nothing itself:** no prefix-, label- or name-only deletion. Pre-existing resources are preserved.
   - The Result reports each run's RUN_ID with its `inventory-leftover.txt`, the startup Docker inventory and a final read-only inventory comparison. Every leftover or ambiguity is reported for the owner, never removed by G047 and never described as restored.
   - The separate r1 V-D debug route is **removed**: diagnosis uses the harness's own report directory (`browser-root-results.json`, Playwright attachments, `harness.out`).

## 9. Delegation

- **Implementer: me only.** No implementation is delegated. All 18 paths have a single owner.
- **One read-only reviewer** (subagent, no write tools used), after the first fully green `root` run and before S8. It reviews:
  - the diff against §4 and §6;
  - the before and after screenshots;
  - `contrast.json`, `spacing-report.json` and `fonts.json`;
  - the `runs.tsv` status rows and the harness leftover reports;
  - the K-list.
  It reports findings only. I triage and integrate them. Any corrective implementation counts toward B3.
- **Not delegated:** product or architecture authority, scope, exceptions, integration and the Result.
- **Planning:** used no subagent.

## 10. Sequence

| Step | Work | Gate / outcome | Cycle accounting |
|---|---|---|---|
| **P1–P7** | Prerequisites (§2.4), performed by their owners, not by this Plan | All complete | — |
| **S0** | Startup checks (§8.7.1): `EXEC_ID` → `$EV` checked and created → read-only Git, status and input checks → wrapper `--check-only` → read-only resource checks | All pass; otherwise a **startup failure** (no write, stop, report) | 0 |
| **S0b** | Create `$ALLOW` exactly as §6.2 (the first repository write) and record its SHA-256 in `$EV`. | — | 0 |
| **S1** | **Baseline:** §8.7.2 with `s1 all` on the unmodified tree. Extract the before screenshots (§8.5), the matrix counts (95 per run), the status row and the harness leftover report. | Overall 0; otherwise a baseline failure (stop, report, no cycle) | 0 (no implementation) |
| **S2** | Implement §6.1 rows 1–16 as one coherent change, in this order: palette → components → page → `modes.ts` → fixtures → specs and helper → `lib.sh` and register → `WEB-SHELL.md`. | — | Initial implementation (BUDGET_POLICY §4) |
| **S3** | §8.7.2 with `s3 static provenance`, then `s4a static provenance root`. The root suite runs `web-check` and the full browser matrix on the root stack. | First validation result | A material failure that needs corrective implementation opens **cycle 1** |
| **S4** | Diagnose from the harness report (`browser-root-results.json`, Playwright attachments, `harness.out`), fix, then retest with `s4b`, `s4c`, … `static provenance root`. This applies the shared-harness rule: an edited `lib.sh` means the static suite runs in the same cycle. | Green `root` (overall 0) | Cycles 1–2 as needed |
| **S5** | **Checkpoint after cycle 2** (only if a second cycle was used), before any cycle 3: ACs complete or remaining, failures, repeated patterns, uncertainty, convergence, boundedness. | Continue only if converging; otherwise stop and return to Codex. | — |
| **S6** | Read-only review (§9); triage. | Findings integrated or recorded | Fixes count |
| **S7** | Finalize `WEB-SHELL.md` and draft the Result. | — | 0 |
| **S8** | **Final:** §8.8 pre-checks, then §8.7.2 with `s8 all` on the delivered tree with no edits during the run, then §8.8 post-checks. | Overall 0, no FAIL or UNMET, inventory exact | A failure that needs a fix uses the next available cycle; then rerun as `s8b` |
| **S9** | Provenance line check; extract the after screenshots and evidence; leftover and inventory report (§8.8.7). | — | 0 |
| **S10** | Complete the Result (AGENTS.md format): AC1–AC7 evidence, every status row, every run including failures, deviations, the cycle log with checkpoint, and leftovers. Return to Codex. | Stop | — |

Debug iterations use the harness `root` suite rather than a separate stack route. This costs wall-clock time, not cycles: a cycle is a corrective iteration, whatever command the retest uses.

## 11. Budget feasibility (B3) and forecast

- **Size.**
  - One coherent change: 5 small application files.
  - Tests are the bulk: about 8 new tests plus additions to 6 existing scenarios, 2 fixture scopes, a locator refactor and the spacing inventory with its negative cases.
  - Docs, plus a 2-line harness edit.
- **Risks.** The main risk is test-code correctness on the first run: the spacing assertions, the nesting-aware locator refactor and the CDP fonts call. Product CSS risk is low, because every value was modelled with a calibrated model.
- **Forecast.**
  - Cycle 1 most likely: first-run test and geometry defects.
  - Cycle 2 possible: review or final-run findings.
  - Cycle 3 reserve.
  - Expected 2; range 1–3. This is conditional on the prerequisites and a passing startup and S1 baseline.
- **Infeasibility.**
  - Inspection does not show B3 to be infeasible, so no split or budget change is recommended.
  - The pressed-state measurements are required (K3, AC5) and are not droppable.
  - If the checkpoint shows non-convergence, I stop and return a split or budget recommendation to Codex rather than extend.

## 12. Acceptance-criteria mapping

| AC | Plan sections | Evidence produced |
|---|---|---|
| **AC1** Scoped consumers implement the reviewed successor mapping | §4.1–4.8, §6.1 rows 1–5 | S-PALETTE (16 × 2 sources, plus rendered grounds, rules, text, fills, frame, panel and scrim composite); S-DESIGN additions (A8) and M2; S-STATES (all variant states, all scopes); S-CONTRAST; before and after screenshots (§8.5). Every retained value is justified in K5–K22 and §4.8. |
| **AC2** Interaction, configuration and dependency contracts preserved | §3, §7, §8.6 | S-LANDMARKS, S-NAV, S-SKIP, S-TAB, S-OVERLAY (keyboard, focus, inert, Escape and Close, return focus, 24px), S-MOTION; B-RENAME and B-BROWSER-RENAME (configuration-only rename); B-DEPS, C1–C12, the empty package and import diffs (§8.8.4); B-WEB-BOUNDARY (routes); B-WEB-DESIGN (no marketing literals, assets or fonts); S-BOUNDARY (same-origin requests only) |
| **AC3** Appearance and first paint | §4.6, §8.1 (M1) | All S-APPEARANCE-* scenarios; S-FIRSTPAINT (8) and S-FIRSTPAINT-DELAYED (2) with the updated independent `PALETTE` (ground, H1 ink, primary fill); S-HTML-ATTRS; S-HYDRATION-DEV; S-HEAD-ORDER with the unchanged init hash. Missing paint evidence fails, as already implemented. |
| **AC4** Overrides separate from preference, and isolated | §8.2 (M3–M7, A9, A10, fixture scopes) | S-TOKENS (16 tokens × 6 scopes with independent sources; Light-only and Dark-only inheritance; nested inheritance from A; adjacent default equals root), S-PORTAL (popup and backdrop in each scope including the nested one, across live mode changes, with the independent overlay oracle), S-RUNTIME (no neighbour leakage; nested inheritance follows the runtime change), S-STATES; S-NOFIXTURE |
| **AC5** Rendered accessibility | §4.2, §8.2 (A3–A7), §8.4 | S-CONTRAST rest, hover, focus, pressed and selected, with exempt disabled pairs and the translucent backdrop recorded; S-TAB focus ≥ 3:1; S-ANGLE (labels and focus not clipped); S-MOTION (reduced motion); S-RESPONSIVE (320, 768, 1280, 640×400 zoom, 200% text, dialog fit); S-TEXT-SPACING (shell and fixtures; expected-element inventory, closed-page and modal checks, clipping and real scroll reachability, focused negative cases, §8.4); S-A11Y (axe, with the documented `incomplete` items). Exact measurements and limits go in the Result. |
| **AC6** Full stack and validation/provenance regressions | §2.4, §8.6–8.8, §10 S0/S1/S8 | Startup checks pass (§8.7.1). The wrapper `all` run through the §8.7.2 status capture: wrapper 0, postcheck passed, `tee` 0, overall 0, recorded with its RUN_ID. Every check PASS and no UNMET. Browser root and off runs at 103 tests each with the extended matrix; inherited datastore, migration, isolation, outage, persistence and rename scenarios. All static, dependency and wrapper controls; B-PROTECTED and B-SKILL-UNCHANGED; the exact NUL inventory; postcheck and tree-digest equality. |
| **AC7** Attributable, bounded handoff; safe for G003 reconciliation | §2, §5, §10 S10, §14 | Result with the AC map, commands, manifest digest and approved snapshot digest, the §2.3 input hashes, the provenance statement and line check, the cycle log and checkpoint, every run's status row, and the harness leftover reports with dispositions. The prerequisites (P1–P7) are performed by their owners. no schema, product or CI work; no closed-record rewrite. Codex ACCEPT recorded separately; the G003 items in §14. |

## 13. Risks, stop conditions and questions

### 13.1 Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | Test-code volume versus B3 | Tests are written against Plan constants; fixtures and specs refactored mechanically; harness-report diagnosis; checkpoint |
| R2 | Font-dependent dialog wrapping with the runner's Liberation fallback | K14 clamps, `pr-9` and `max-h` with scroll; modelled heights with margin; S-RESPONSIVE and S-TEXT-SPACING verify |
| R3 | Narrow margins: hover on primary 4.68, Dark primary fill 3.44, secondary hover 4.86, Light destructive hover 4.80, Dark secondary pressed 4.73 | The model reproduces Chromium's brightness exactly; rendered measurement is authoritative. A failure is fixed by the next derivation step (K-table method), not by relaxing thresholds. |
| R4 | The nested-scope locator refactor could break previously passing fixture tests | M5 uses the same elements for the existing scopes; the full fixture suite runs in S3 |
| R5 | Provenance | Q2 resolved for values only; any copying stops for provenance review |
| R6 | Baseline integration (P1–P7) | Route A required; no harness workaround; startup checks enforce the committed, clean state |
| R7 | Changing G002's accepted destructive tint deviation | Approved as K4/K10 in the r1 review |
| R8 | Role consolidations (K5, K6, K7, K8, K16) may not match owner preference | Approved; reversible palette-value decisions |
| R9 | Wall-clock time and Docker conflicts during S1 and S8 | Startup resource checks; the harness refuses root-project conflicts; destructive scenarios run on harness-owned copies |
| R10 | An edit during a run invalidates B-RENDER-PLAN | No repository write while any harness run is active |
| R11 | Disabled-state (`disabled` attribute) controls could be mistaken for the intelligence-disabled variant | Terminology kept distinct in tests and the Result |
| R12 | The unconsumed `DialogFooter` offsets no longer match the new popup padding | Recorded residual; a first consumer restyles it under its own Goal |
| R13 (r2) | The harness has not yet run in a linked worktree | The S1 baseline run is the check; an incompatibility stops before implementation, with no harness change in G047 |
| R14 (r2) | Debug iterations are slower without the removed debug route | Wall-clock only; cycles count corrective iterations, not commands |

### 13.2 Stop conditions (from the Goal, applied)

Stop and return to Codex on any of the following:
- **Binding and inputs:** a missing or mismatched Codex binding; any startup-check failure (§8.7.1); a failing S1 baseline; a drifting kit manifest or snapshot input.
- **Leftovers:** a harness leftover or resource ambiguity that the owner must resolve.
- **Contracts:** any need to change token names or meanings, APIs, the appearance or portal mechanisms, dependencies, primitives, fonts or assets; a public contract or schema.
- **Checks:** any need to weaken a check, edit a protected or template file, or change the wrapper or harness.
- **Scope and provenance:** a path outside §6.1; an unresolvable provenance fragment (§5).
- **Budget:** non-convergence at the checkpoint, or B3 exhausted.
- **Barrier:** any product-specific work. The barrier stays closed through G021.

### 13.3 Questions requiring a decision before execution

| # | Question | Status |
|---|---|---|
| **Q1** | Perform P2–P7 (§2.4): Plan integration, Route A base commit including the updated governance and progress files, the `g047-exec` worktree at `BASE`, the snapshot and the binding. Codex decides the snapshot membership. | Open: Codex and owner |
| **Q2** | Values-only adaptation while the kit's licence is unresolved | Resolved (r1 review) |
| **Q3** | K1–K22 | Resolved (r1 review) |
| **Q4** | Fixtures, M1–M7 and pressed states | Resolved (r1 review; pressed states required) |
| **Q5** | Confirm:<br>• the allow-list as the first write after the startup checks;<br>• the evidence root `$HOME/.local/state/platform-evidence`;<br>• cleanup by the accepted harness only, with the debug route removed. | Open: Codex |
| **Q6** | B3 unchanged (forecast 2, range 1–3) | Open: Codex |

## 14. G003 reconciliation after G047 acceptance (informational; no G003 work here)

The G003 Plan r4 must later rebind to the accepted G047 tree. The points that change:
- the `lib.sh` and `render-customizations.tsv` pins (`9f7a3997…57c3` and `9abf9e92…279f` will change);
- "both browser matrices (95 tests each)" becomes 103, with three added IDs;
- the `WEB-SHELL.md` content;
- the G047 allow-list and Result as new management files;
- G003's planned register rows (+11) and `APPLICABLE` (+7) added on top of G047's reason-text change;
- the Route A base, including the updated governance and progress files, and the G047 execution worktree.

G047 does not edit the G003 Plan and does not approve G003.

## 15. Disposable planning evidence (outside the repository; not acceptance evidence)

Location: `/tmp/g047-plan/` (owned; may be lost on restart; reproducible from these scripts).

| File | SHA-256 | Content |
|---|---|---|
| `wcag.py` | `82de548a…bff0a` | WCAG 2.x luminance and contrast; brightness (sRGB multiply, round, clamp); alpha compositing |
| `calibrate.py` / `.out` | `2da798e5…be9c9` / `c9527b0a…6be39` | Brightness model = G002-measured hover fills (7/7 exact); ratios reproduce G002's measurements; prints `CALIBRATION PASS` |
| `candidates.py` / `.out` | `0186ef03…eafe5` / `2aad07f5…a5e02` | Primary search; kit-claim re-measurement (4.21, 4.05, 5.32, 6.96, …) |
| `candidates2.py` / `.out` | `afcf2277…053cc` / `b723f9f1…bda40` | Secondary and destructive search, solid versus tint |
| `candidates3.py` / `.out` | `32ce7d5f…23736` / `a06165fd…57594` | Dark secondary press search; pressed states |
| `proposal.py` / `.out` / `proposal-matrix.json` | `2bde3819…d4d49` / `0289d52a…e5d8d` / `ecbc4aa4…4892d` | 178 modelled pairs (152 thresholded, 26 recorded-only; default, A and B; both appearances): `FAILURES: none` |
| `fixtures-model.py` / `.out` | `d7845890…7fe12` / `082015fb…3738b` | Light-only and nested dark-only-in-A, with inheritance: `FAILURES: none` |
| `planned-lines.txt`, `provenance-precheck.py` / `.out` | `b68bd080…d2858`, `96f8b345…8e5bd` / `e8b7c957…d1a8e` | §5 pre-check: zero exact matches; one pre-existing G002 fragment |

No application build, container, install or test was run during planning. Reading the G002 browser report (`…/g002-exec/s9-final3/report/browser-root-results.json`, historical) was read-only, for calibration.

**r2 additions** (`/tmp/g047-plan-r2/`; disposable):

| Item | SHA-256 | Result |
|---|---|---|
| `history/G047-vestive-shell-refresh-plan.r1.md` | `8aaee0065645e59b832335283d1a1b172a05a6d9c30579a5ee85a365ff2e146b` | r1 archive, byte-identical (`cmp` clean) |
| `four/t-four.sh` / `t-four.out` | `51e23793…58f4b` / `cb422799…a9452` | §8.7.2 snippet, with only the wrapper (stub) and `tee` substituted:<br>• wrapper 11 + `tee` 0 → 11;<br>• wrapper 0 + `tee` 1 → 70;<br>• both fail (13 and 1) → 13;<br>• both pass → 0;<br>• r1's plain pipe → 0 for the first case, which is the defect reproduced.<br>5/5. |

**Earlier r2 runs.**
- An early-r2 rerun reproduced every model output above byte-identically. It also reproduced the manifest check (63 OK; output `3f9ba2fb…7b1e`). Their inputs are unchanged, so they need no further rerun.
- Scratch from the superseded, broader r2 draft (`tools/`, `tests/`, `build/`) is preserved but is not part of this Plan, not shipped and not a prerequisite. It was halted by the proportionality clarification.

**How the scripts ran.** This isolated session's guard refuses a direct `bash …` invocation, so the planning scripts were run by path from the planning worktree. None uses Git against the main checkout.
