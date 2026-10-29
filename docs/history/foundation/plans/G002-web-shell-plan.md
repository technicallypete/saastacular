> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G002 — Accessible web shell implementation plan

Goal: [G002-web-shell.md](../goals/G002-web-shell.md)
Spec: [G002-web-shell-spec.md](../specs/G002-web-shell-spec.md)
Owner: Claude Code · Reviewer: Codex · Budget class: **B7**: 7 shared meaningful cycles, checkpoint after cycle 4. Codex authorized B7 in the r2 review; it is not Plan approval.

| Revision | Date | Change |
|---|---|---|
| r1 | 2026-10-05 | Initial Plan (SHA-256 `851d003f…5b28`). Codex review: **CORRECTION LOOP**, with three corrections and Q0–Q5 resolved (§0). |
| r2 | 2026-10-05 | Addresses all three r1 corrections, and adds product-owner R8/AC9 (Light/Dark/System, System default) and R9/AC10 (Vestive visual language, including the three example screenshots as visual references; marketing copy and the dark-only rule ignored). Keeps the four concerns separate (§3). Recommends B7 (§19). SHA-256 `fbb910e1e2d08cfbedd687abeb43d2a010c8dc4d4cbdfd93d04e845f985730c9`. Codex review: **CORRECTION LOOP**, with two corrections, B7 authorized and Q6–Q10 resolved (§0). |
| r3 | 2026-10-05 | Addresses both r2 corrections. **(1) First paint:** a parser-blocking inline `<head>` script is now the mandatory mechanism (`next/script` `beforeInteractive` is dropped). It is evidenced by a paint-timed, visible-content-synchronized probe, a delayed-framework-JS scenario with stored appearance opposite the OS, head-order checks and `<html>` attribute checks. The import order is aligned with the build gates. **(2) Dependency controls:** each forbidden mutation is internally consistent with its lockfile and asserts its intended reason code, with a separate lock-mismatch control. Records B7 and the Q6–Q10 resolutions, refreshes the reviewed-input hashes and selects the exact-hash route. SHA-256 `1689f7384a0c61946933733bd6583ddc3825ee99bb727509a0823b864a0165a6`. Codex review: **CORRECTIONS REQUIRED**, two corrections (§0). |
| r4 | 2026-10-05 | Addresses both r3 corrections. **(1) Reason codes:** independent rules, and each control asserts its complete expected code set; the lock-mismatch controls mutate only lock metadata; one control per rule (C1–C12). **(2) Fail-closed validation:** a committed `infra/validation/run-approved.sh` verifies the snapshot's pinned digest on the bytes it parses, requires exactly 20 approved unique relative paths, verifies every entry before invoking the harness, preserves the harness exit status through an always-run postcheck, and fails the run on any failure. Wrapper controls W1–W9 use a stub harness. Supervisor hashes are refreshed. SHA-256 `db4373ab9242f746c66371962cdce2f0337e60577b4c345e4be8388dee645e62`. Codex review: **CORRECTIONS REQUIRED**, one correction: the bootstrap dependency (§0). |
| r5 | 2026-10-05 | Removes the bootstrap dependency. **S0a** is a self-contained, read-only, fail-closed Bash gate needing only bash, coreutils, git and docker. Its exact text and SHA-256 are in §16.1. It verifies HEAD, the pinned snapshot digest on the bytes it parses, exactly 20 approved unique relative paths, the input hashes, the NUL-safe starting state and local resources, before any write. Only after S0a passes and the Plan is approved: **S0b** creates `run-approved.sh` and `run-approved-controls.sh` (the only bootstrap writes, gated), and **S0c** runs W1–W9 and `--check-only` before S1. The S1 comparison and the S5 entry are reconciled. A read-only planning demonstration of S0a against the actual baseline (wrapper absent) is recorded (§16.4). Supervisor hashes are refreshed. SHA-256 `b7d7f44b988800da78768ab488053bed807ce712a19bb6d1de48f22bcb967305`. Codex review: **CORRECTIONS REQUIRED**, two S0a read-only defects (§0). |
| r6 | 2026-10-05 | Fixes the two S0a read-only defects; everything else is unchanged. **(1)** A new step 0 resolves the evidence destination with `realpath -m` and rejects (exit 26) the repository, its descendants, traversal, and external or dangling symlinks resolving into it, **before any directory or file is created**. Creation uses the resolved path, which is re-verified before every evidence write. **(2)** All Git inspection runs with `git --no-optional-locks -c core.fsmonitor=false`, with `GIT_OPTIONAL_LOCKS=0` exported and inherited Git redirection variables unset, so `.git/index` is never refreshed. The new exact S0a SHA-256 is `853b20df…b428` (the r5 script was `aa5aacf1…32ac`). P and N1–N11 are repeated, and controls A1–A4 (containment), B1–B4 (symlink containment), C0–C4 (index immutability, with a stale-index fixture and a non-vacuous detector) and D (whole-tree inventory, including `.git`) are added. All 28 results are as expected (§16.4). |

## 0. Review record

### r1 → Codex: CORRECTION LOOP (2026-10-05)

| # | Codex correction | Where r2 addresses it |
|---|---|---|
| 1 | B-DEPS must keep the existing dependency constraints: every existing runtime and dev dependency keeps its name, category and version; the only additions are the reviewed entries in their stated categories. Keep exact-pin, lockfile and image-version checks. Disposable negative controls must show that a framework-version change and an extra exact-pinned package fail, while the intended additions pass. | §5, §15 (B-DEPS rewritten; B-DEPS-CONTROLS added). R8 adds **no** dependency, so the reviewed seven additions are unchanged. |
| 2 | Token and contrast assertions must be executable and complete. Separate semantic colour tokens from non-colour tokens and derived mappings. Correct the S-CONTRAST pairs: real surfaces, hover and focus states, 3:1 control, icon and focus boundaries, and translucent backgrounds composited against their real backing. Keep the fixtures, portal checks and adjacent-scope isolation. | §7 (contract split into colour, non-colour and derived), §11 (S-CONTRAST rewritten around rendered-pixel sampling), §15 (B-WEB-TOKENS split). The “retain seven existing defaults” instruction is superseded by the later product-owner R9 decision; the names are retained and the value changes are documented (§7.4). |
| 3 | Reconcile the execution gates with the known state and CLI behaviour. Refresh S0 to the current HEAD and permit only the reviewed exact paths, hash-bound. Supply valid `--expect-hash` values. S1 must explicitly permit the previewed temporary `layout.tsx` font mutation. Enumerate transient writes versus final modifications, stop on any other write, and verify protected paths against the refreshed baseline. | §16 (S0 and S1 gates, transient-write table), §17 (exact final invocation for the committed and uncommitted baselines), §15 (B-SKILL-UNCHANGED handling of the owner-input design skill) |

**Question resolutions:**
- **Q0:** the planning release is recorded.
- **Q1:** the fixture build is accepted.
- **Q2:** the shared web network namespace is accepted. It is not an isolation claim; there are no published fixture ports and no secret environment.
- **Q3:** `base-nova`, `eject` and `cn@0.4.0` are accepted, with recorded provenance and no silent registry refresh.
- **Q4:** the pinned Playwright image is a test tool, not infrastructure needing an ADR. Record and verify the full digest and the version agreement.
- **Q5:** Chromium-only coverage is accepted. Record the browser version and limitations.

### r2 → Codex: CORRECTION LOOP (2026-10-05)

| # | Codex correction | Where r3 addresses it |
|---|---|---|
| 1 | `beforeInteractive` does not establish a first-paint guarantee. Choose an initialization that sets the validated stored mode before theme-bearing content can paint, or make the inline-head alternative mandatory. S-FIRSTPAINT must synchronize on visible theme-bearing content, distinguish actual paint timing from first-rAF or pre-hydration assumptions, include a delayed framework-JS scenario (stored Light vs OS Dark, stored Dark vs OS Light), record probe and paint timing with DOM and palette evidence, and reject runs with no valid visible-content sample. Keep the storage fallback, dev and production hydration checks and the narrow suppression. Create modules before the build gate that imports them. | §8.2 (inline parser-blocking head script, mandatory); §11.2 (S-FIRSTPAINT rewritten; S-FIRSTPAINT-DELAYED, S-HEAD-ORDER and S-HTML-ATTRS added); §11.4 (probe design); §16.3 (S2 creates every module that S2 files import); §21 (R4, R12) |
| 2 | Each disposable forbidden manifest mutation must be internally consistent with its lockfile (root category and specifier, resolved node-module version and category metadata), and must assert the intended rejection reason, so that a lock mismatch alone cannot make a control pass. Keep the real tree as the positive control. | §15 (B-DEPS emits reason codes; B-DEPS-CONTROLS C1–C7 with lockfile-consistent mutations, an asserted reason, absence of `DEPS_LOCK_MISMATCH`, and a separate lock-mismatch control) |

**Budget (Q8):** **B7 is authorized**, with the checkpoint after cycle 4. The combined G002 scope is kept and the split is not selected. The two read-only reviewers are approved.

**Question resolutions:**
- **Q6:** the `#5952ff` primary and the documented secondary adaptations are approved as a bounded reconciliation. Rendered rest, hover and focus contrast remains authoritative.
- **Q7:** narrow `<html>`-only suppression is accepted, with explicit checks that html's language and other attributes stay correct and no descendant suppression exists. Dev and production checks stay. See S-HTML-ATTRS and B-WEB-BOUNDARY.
- **Q9:** the exact-hash uncommitted-input route is selected, with no commit. The §2.1 hashes are refreshed. Stop on any unexpected state or hash change, and never bind freshly observed content as approved (§16.1, §17).
- **Q10:** the emphasis-ink `accent` role is approved. No `bg-accent` consumption anywhere in the delivered set (B-WEB-TOKENS). A future surface use needs a reviewed, contrast-compatible foreground pairing.

### r3 → Codex: CORRECTIONS REQUIRED (2026-10-05)

| # | Codex correction | Where r4 addresses it |
|---|---|---|
| 1 | The checker prints every violation. C5 triggers `DEPS_ADDITION_WRONG` and `DEPS_NOT_EXACT`, and C6 triggers `DEPS_EXISTING_CHANGED` and `DEPS_LOCK_MISMATCH`, so the written single-code assertions cannot hold. Specify exact expected code sets with no unexpected codes, or a reviewed priority scheme preserving every rule. Make C6 mutate only lock metadata. Keep C1–C5 internally consistent and C7 valid. Weaken no rule. | §15 B-DEPS (rules independent, with no priority scheme; output is the sorted unique set of `CODE name`) and B-DEPS-CONTROLS (C1–C12, each with its complete expected set; C6–C8 are lock-only, one per lock sub-rule) |
| 2 | §17 does not fail closed: a failed precheck need not stop execution, and a successful trailing hash check can mask a failed harness. Use a stated Bash context; stop before any build or test if snapshot-digest or entry verification fails; pin the snapshot's own digest from the approval; require exactly the approved 20 unique relative paths; build arguments from the verified bytes; preserve a nonzero harness status; any failed precheck, harness run or postcheck fails the run and stops the handoff. Add scratch controls for entry mismatch (harness not invoked), changed snapshot, harness failure with a successful postcheck, and postcheck drift; all-valid passes. Keep NUL-safe paths and quoted arrays. | §17 (new committed wrapper `infra/validation/run-approved.sh`, exact control flow and exit codes), §15 B-WRAPPER-CONTROLS (W1–W9 with a stub harness), §16.1 (S0 uses `--check-only`), §16.3, §18, §23 |

Also from the r3 review:
- The first-paint, appearance, design and override planning is adequate at planning level. Implementation must record actual observation times for computed-style samples, distinguish them from rAF callback arguments and observer delivery times when correlating with paint, and must not use later screenshots alone as first-frame proof. §11.4 is amended accordingly.
- B7 remains authorized, with the checkpoint after cycle 4.
- `/tmp/g002-r3-reviewed-inputs.sha256` is a continuity snapshot only, **not** an approved execution snapshot.

### r4 → Codex: CORRECTIONS REQUIRED (2026-10-05)

| # | Codex correction | Where r5 addresses it |
|---|---|---|
| 1 | S0 invokes `run-approved.sh --check-only`, but S5 creates it, so valid inputs stop with a missing command, while skipping or moving S0 would violate the approved-input gate. Specify a fail-closed, read-only initial Bash precheck using baseline tools. It verifies the pinned snapshot digest, the exact approved unique relative path set, the input hashes and the NUL-safe starting state before any write, CLI, install or build. After it passes, and only after Plan approval, create the wrapper as the first validation-tool step and run `--check-only` and W1–W9 before S1. Reconcile the initial and subsequent inventories, the allowed bootstrap writes, the S1 comparison and the S5 entry. Preserve every exit-status and postcheck rule. Demonstrate the ordering against the baseline, where the wrapper is absent: a malformed snapshot or a mismatched input must prevent bootstrap and application work. | §16.1 (S0a gate: full script, SHA-256 and exit codes), §16.2 (S0b bootstrap-write rows), §16.3 (S0a → S0b → S0c → S1; S5 wires the existing tools), §16.4 (read-only demonstration N1–N11), §15 (B-WRAPPER-CONTROLS implemented by the committed controls script), §17, §18, §23 |

Also from the r4 review: C1–C12 and the B-DEPS semantics are coherent at planning level. The broad design-path regex is not authorization for new inputs: the supervisor-pinned snapshot binds the exact 20 paths, and the regex is only an additional syntax constraint. B7 remains authorized with the checkpoint after cycle 4. `/tmp/g002-r4-reviewed-inputs.sha256` is continuity only, not an approved execution snapshot.

### r5 → Codex: CORRECTIONS REQUIRED (2026-10-05)

| # | Codex correction | Where r6 addresses it |
|---|---|---|
| 1 | The S0a script runs `mkdir -p -- "$EV"` before canonicalizing and rejecting an in-repository evidence directory. Codex reproduced exit 26, but only after `new-evidence` had been created inside a disposable repository, both directly and through an external symlink. Resolve and check the destination without writes before any mkdir or output; reject the repository itself, its descendants, and paths resolving into it through symlinks or traversal; recheck before writing evidence. | §16.1 step 0 and script (`realpath -m`, `inside_repo`, `mkdir` only afterwards, on the resolved path; `ev_ok` before every evidence write); controls A1–A4 and B1–B4 (§16.4) |
| 2 | Ordinary `git status` can refresh `.git/index` even with empty porcelain output, which Codex reproduced with a stat-only change in a disposable fixture. Run S0a's Git inspections with optional locks disabled. An unchanged status digest alone does not prove this. | §16.1 script: `g()` = `git --no-optional-locks -c core.fsmonitor=false -C "$REPO"`, with `GIT_OPTIONAL_LOCKS=0` exported and Git redirection variables unset; controls C0–C4 (stale-index fixture, non-vacuous detector) and D (whole-tree inventory, including `.git`) in §16.4 |

Also from the r5 review: the S0b/S0c ordering, S1 before-snapshot, S5 control reuse, final harness-status and drift behaviour, root `.env` failing with 24, B7 with the checkpoint after cycle 4, and the 5–6 cycle forecast are coherent and unchanged. `/tmp/g002-r5-reviewed-inputs.sha256` is continuity only, not an approved execution snapshot.

**Product-owner additions after the r1 review** (PROGRESS.md 2026-10-05):
- R8/AC9: Light, Dark and System, with System the default. This supersedes r1's “no dark theme”.
- R9/AC10: use `skills/vestive-design-skill`. This supersedes the neutral-only values and preset radius.
- Clarifications: ignore marketing copy, content fundamentals and the dark-only rule. The skill's example screenshots are visual references; no new Goal or acceptance criterion.

## 1. Summary

G002 extends the accepted web skeleton in place. It touches only `web/`, project-owned validation and Docker test tooling, and docs.

- **Components.** Pinned shadcn 4.21.1 with Base UI (`base-nova`): only `button` and `dialog`, then `eject`, so that no CLI ships at runtime.
- **Shell.** A generic, server-rendered shell at `/`: skip link, ruled header with configured identity, in-page navigation, an appearance control, a Base UI help dialog, three ruled sections and a footer.
- **Four separate concerns** (§3):
  1. **Design language:** Vestive visual rules, expressed as palette values and typography and layout roles.
  2. **Theming infrastructure:** mode-aware semantic tokens.
  3. **Organization-level overrides:** `TokenScope`, with per-mode source values and scope-local portals.
  4. **User appearance preference:** Light/Dark/System in `localStorage`, System by default, with System tracked live in CSS.
- **Browser evidence.** Playwright 1.63.0 with axe-core 4.13.0, in a pinned Playwright container. The matrix covers both resolved appearances, default plus two scoped fixtures, portals, appearance behaviour, first paint, hydration (production and dev), responsive layout, design-language computed styles, and screenshots for visual review.
- **Harness.** Reconciled narrowly. B-DEPS becomes stricter, with negative controls.
- **Not changed:** no schema, API or MCP change; no Organization storage; no product feature; no dependency beyond r1's seven. The accepted project-archetype Skill and the design skill stay unchanged.

## 2. Current state (verified 2026-10-05)

### 2.1 Repository

- **HEAD** is `cc157ed` (“License project archetype under MIT and add package README”). That separately authorized Skill packaging is part of the protected baseline, not a G002 change. Earlier commits are `701693b` (G002 Goal and Spec) and `b405077` (G001 acceptance).
- **Working tree.** It differs from HEAD only by:

  | Path | State | Owner | SHA-256 (Codex r5-review continuity snapshot; verified at r6 writing) |
  |---|---|---|---|
  | `management/v0/goals/G002-web-shell.md` | modified | Codex | `a7d516a92411f6304089655dc0f975c7cbf25cf7ad609c165d1707fa654f9885` |
  | `management/v0/specs/G002-web-shell-spec.md` | modified | Codex | `7f7e52368c1104300bb1d4bac9c40a0c4d2a2073a20f1a33ff83a4fbb9cad1e3` |
  | `management/v0/PROGRESS.md` | modified | Codex | `1918c739d55f2097f26313501b38534af107ee8863fe7c825fa81ffd5dee663a` |
  | `management/v0/plans/G002-web-shell-plan.md` | untracked | Claude | this file (r2 `fbb910e1…30c9`, r3 `1689f738…65a6`, r4 `db4373ab…5e62`, r5 `b7d7f44b…7305`); the approved hash is fixed only by Codex's approval |
  | `skills/vestive-design-skill/**` (16 files, §9.1) | untracked | product owner | listed in §9.1; all 16 match the snapshot |

  Codex's continuity snapshot is `/tmp/g002-r5-reviewed-inputs.sha256` (its own SHA-256 is `23d5894435e323daaf8a597e34465f0ae1a543aa784b85367ed6b3be0fc7606a`), outside the repository. At r6 writing, `sha256sum -c` reported all 20 entries OK, and the revised S0a gate passed against it (§16.4). The Plan entry is r5 and changes with this revision. It is **not an approved execution snapshot**. Execution uses only the snapshot and digest named in Codex's approval (§16.1, §17).

  Nothing is ignored, there is no `.env`, and there are no Docker resources for this project.
- **Governance documents.** AGENTS.md, CLAUDE.md, ARCHITECTURE.md and BUDGET_POLICY.md are unchanged since r1. SECURITY.md and DECISIONS.md are still empty.
- **Architecture.** §37A, §50, §58 and §59 still govern. §37A's “design language skill” is now supplied as `skills/vestive-design-skill`.

### 2.2 Web implementation (unchanged since r1; rendered by G001, byte-identical to manifest `55bb52881a31a589`)

As inspected in r1:
- Next.js 16.3.8, React 19.3.0, TypeScript 5.9.3 and npm. Tailwind 4.3.3 with seven semantic tokens.
- `layout.tsx` and `page.tsx` are minimal. The BFF is server-only, and `/api/health` and `/api/ready` send no-store.
- The Docker check runs lint and types. There are no components and no tests.
- The harness assumptions needing reconciliation remain: B-DEPS manifest equality, the G001-only `APPLICABLE` list, and the three-entry register.
- All web templates live in the archetype `base` layer, so they are identical in the intelligence-disabled render.

### 2.3 Research retained from r1 (containerized, read-only; nothing written to the repository)

- **Pins:** `shadcn` 4.21.1, `@base-ui/react` 1.8.0, `cn` 0.4.0, `class-variance-authority` 0.7.1, `lucide-react` 1.52.0, `tw-animate-css` 1.4.0, `@playwright/test` 1.63.0 and `@axe-core/playwright` 4.13.0.
- **Playwright image:** `mcr.microsoft.com/playwright:v1.63.0-noble`, index digest `sha256:eff16c30e6f3f4af0a03fa4b706120d5e9b0891c344a27d64559aff5900a4a27`.
- **Disposable CLI preview** (`git archive` copy in a throwaway container). `init -b base -p nova -y`, `add button dialog -y` and `eject -y`:
  - wrote `components.json`, `lib/utils.ts`, `components/ui/{button,dialog}.tsx`, `package.json`, `package-lock.json` and `app/globals.css`;
  - **temporarily added a Geist `next/font/google` import to `app/layout.tsx`**;
  - left 14 new lockfile entries after `eject`;
  - produced a production build that succeeded.
- **Base UI** `Dialog.Portal` accepts `container` (an element or `RefObject`).
- **Contrast arithmetic** for the source and candidate palettes (WCAG relative luminance) is summarized in §7.3. The rendered-pixel browser tests are authoritative.

## 3. Four separate concerns (they must not collapse into one mechanism)

| Concern | What it decides | Mechanism (owner files) | Inputs | Must not |
|---|---|---|---|---|
| **A. Product design language** (Vestive skill) | Palette *values*, typography roles, flat geometry, 2px rules, flush-left hierarchy, angled filled buttons, focus style | Palette source variables `--light-<token>` and `--dark-<token>` on `:root`, plus non-colour design tokens (`--radius`, `--angle`, fonts, container and spacing) in `web/app/globals.css`. Shell layout classes in `components/layout/*`. The `angled` utility used by button variants. | §9 mapping from `SKILL.md`, `README.md`, `styles.css`, `reference.html` and the 3 screenshots | Carry marketing copy or identity; force Dark; load remote fonts or assets; scroll grids instead of reflowing |
| **B. Theming infrastructure** | How a semantic token resolves for the active appearance, at `:root` and every scope boundary | `globals.css`: a resolution block `:root, [data-token-scope] { --<token>: var(--light-<token>) }`, with `@variant dark` resolving to `var(--dark-<token>)`; `color-scheme`; the appearance-aware `@custom-variant dark`; `@theme inline` mapping to utilities. `lib/tokens.ts` names the contract. | Concern A values; the `data-appearance` attribute from D; `prefers-color-scheme` | Know about Organizations, storage or identity |
| **C. Organization-level token overrides** | Different source values for one subtree, for each appearance | `components/layout/token-scope.tsx`: inline `--light-*` and `--dark-*` overrides only, never resolved tokens. `components/ui/portal-container.tsx`: a scope-local overlay container. Fixtures only in tests. | `tokens: { light?, dark? }` props (test fixtures in G002; authorized Organization data in G017) | Persist, resolve an Organization, select by route, query, header or storage, or decide the appearance mode |
| **D. User appearance preference** | Light, Dark or System; System by default; persisted in the browser | `lib/appearance.ts` (modes, storage key, validation, init-script source), `components/layout/appearance-control.tsx` (native radio group). A parser-blocking inline `<head>` script in `layout.tsx` sets `<html data-appearance>` before any body content exists (§8.2). System tracking is pure CSS (`prefers-color-scheme`). | `localStorage` value; user choice; OS preference | Store a resolved palette instead of the mode; use a server cookie, schema or account data; carry authorization meaning |

**How they compose**, in one direction only:
1. **D** sets the mode attribute.
2. **B** picks the light or dark source for every semantic token at `:root` and at every `[data-token-scope]` element.
3. **C** replaces sources within a subtree.
4. **A** supplies the default sources.

A mode change therefore re-resolves every scope, which keeps scope isolation and portal styling (§7.5). An Organization override never changes the mode, and the mode never changes override values.

## 4. shadcn/ui and Base UI integration

### 4.1 CLI steps (pinned; in Docker; on the real `web/`)

```sh
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp \
  -v "$PWD/web:/app" -v g002-web-nm-<exec-id>:/app/node_modules -w /app node:24-slim \
  sh -c 'npx -y shadcn@4.21.1 init -b base -p nova -y && npx -y shadcn@4.21.1 add button dialog -y && npx -y shadcn@4.21.1 eject -y'
```

`node_modules` goes into an execution-owned, labelled volume (`g002.exec=<exec-id>`), so nothing lands on the host tree. The raw output and the diff are saved outside the repository as provenance. The S1 write gate is in §16.

### 4.2 `components.json`

Kept exactly as the CLI writes it. The style is `base-nova` (accepted, Q3). Further settings: `rsc`, `tsx`, `tailwind.css: "app/globals.css"`, `cssVariables`, `baseColor: "neutral"` (with values replaced by §7), `iconLibrary: "lucide"`, and aliases `@/components`, `@/components/ui`, `@/lib/utils`, `@/lib` and `@/hooks` (the last is not created).

### 4.3 Reviewed reconciliation of CLI output

| CLI output | r2 decision | Reason |
|---|---|---|
| `shadcn` runtime dependency (removed by `eject`) | Absent | No CLI or MCP tooling at runtime |
| Caret specifiers | Exact pins (§5), set with `npm install --save-exact` in the container | R1 |
| `layout.tsx`: Geist from `next/font/google` (transient) | **Removed at S2.** `layout.tsx` is instead customized only for the appearance init script and `suppressHydrationWarning` (§8). This is now a registered customization. | No external font; the R8 first paint needs it |
| `@import "tw-animate-css"`; ejected `shadcn/tailwind.css` block (marker `ejected from shadcn@4.21.1`) | Kept verbatim | Registry variants and animations |
| `@custom-variant dark (&:is(.dark *))` | **Replaced** by the appearance-aware dark variant (§7.5) | Generated `dark:` classes must follow the R8 mode, and System must work without JS |
| shadcn token set, `.dark` block, `card`, `chart`, `sidebar`, `accent-foreground`, `--font-sans` self-reference, base `outline-ring/50` | **Replaced** by §7 (16 colour tokens with light and dark sources; non-colour design tokens) and a base layer of `* { @apply border-border outline-ring; }` | R3 minimal contract; R9 values; opaque focus |
| `lib/utils.ts` (`export { cn } from "cn"`) | Kept | Accepted (Q3) |

Generated-source edits are the only intended differences from the registry. They are recorded in WEB-SHELL.md and the Result, and are visible through `shadcn add <x> --diff`:

| # | File | Edit | Reason |
|---|---|---|---|
| E1 | `ui/button.tsx` base classes | Focus: `focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring` replaces `focus-visible:border-ring ring-3 ring-ring/50`; the destructive variant uses the same outline. | Skill: “focus ring 2px gold, 2px offset”. A 50% alpha ring cannot reach 3:1. An outline is not clipped by the angled fill. |
| E2 | `ui/button.tsx` base classes | Add `font-display uppercase tracking-[1.5px]` (radius comes from the token: 0) | Skill: buttons in display type, uppercase, tracked |
| E3 | `ui/button.tsx` filled variants (`default`, `secondary`, `destructive`) | Fill moves to an `angled` pseudo-element: `angled before:bg-primary` and so on. Hover is `hover:before:brightness-[1.12]`; press is `active:before:brightness-90`. `ghost`, `link`, `outline` and the icon sizes stay rectangular. | Skill: angled silhouette and brighten or dim. Because the element itself is unclipped, labels and focus outlines are never clipped (R9). |
| E4 | `ui/button.tsx` `link` variant | `text-primary` becomes `text-accent` | Skill: links are gold. `primary` as text on the dark ground is 3.77:1, below 4.5:1. |
| E5 | `ui/button.tsx` `outline` variant | `border-input` in both appearances (it was `border-border` in light) | 3:1 boundary for an outline control (the light `border` is decorative, 1.39:1) |
| E6 | `ui/dialog.tsx` backdrop | `bg-black/10 supports-backdrop-filter:backdrop-blur-xs` becomes `bg-overlay` | No palette literal; skill: no blur |
| E7 | `ui/dialog.tsx` `DialogPortal` | Default `container` from `usePortalContainer()`; an explicit prop still wins | The overlay stays in its token scope |

## 5. Dependencies (unchanged from r1; theming adds none)

| Package | Pin | Category | Purpose |
|---|---|---|---|
| `@base-ui/react` | 1.8.0 | dependencies | Base UI Button and Dialog |
| `class-variance-authority` | 0.7.1 | dependencies | Generated variants |
| `cn` | 0.4.0 | dependencies | Generated `cn()` |
| `lucide-react` | 1.52.0 | dependencies | Dialog close icon (the configured icon library; see adaptation T9) |
| `tw-animate-css` | 1.4.0 | dependencies | Generated animations |
| `@playwright/test` | 1.63.0 | devDependencies | Browser runner; equals the image version |
| `@axe-core/playwright` | 4.13.0 | devDependencies | Accessibility audit (`axe-core ~4.13.0`) |

- **Unchanged and enforced by B-DEPS:** dependencies `next 16.3.8`, `react 19.3.0`, `react-dom 19.3.0`, `server-only 0.0.1`; devDependencies `@tailwindcss/postcss 4.3.3`, `@types/node 24.19.1`, `@types/react 19.3.0`, `@types/react-dom 19.3.0`, `eslint 9.39.5`, `eslint-config-next 16.3.8`, `tailwindcss 4.3.3`, `typescript 5.9.3`.
- **Transitive** (attributed in the Result): `@base-ui/utils`, `@floating-ui/*`, `@babel/runtime`, `clsx`, `reselect`, `use-sync-external-store`, and the Playwright and axe trees.
- **No theming library** (no `next-themes`), **no font package**, and nothing else.
- **Images:** the test-only `mcr.microsoft.com/playwright:v1.63.0-noble@sha256:eff16c30e6f3f4af0a03fa4b706120d5e9b0891c344a27d64559aff5900a4a27`, plus the existing `node:24-slim`.

## 6. Shell and component structure

### 6.1 Files

| Path | Kind | Concern | Responsibility |
|---|---|---|---|
| `web/app/layout.tsx` | rendered, **customized** | D (bootstrap) | Adds `suppressHydrationWarning` to `<html>` only. Adds a `<head>` whose children include a **classic, synchronous inline `<script id="appearance-init" dangerouslySetInnerHTML={{ __html: appearanceInitScript }}>`** (no `src`, `async`, `defer` or `type`), with source from `lib/appearance.ts` (§8.2). Metadata (title from `PRODUCT_NAME`), `dynamic`, body classes and `globals.css` are unchanged. |
| `web/app/page.tsx` | rendered, **customized** | — | Server: reads `runtimeConfig()` as today and renders `<AppShell productName platformBaseDomain>`. Keeps `data-testid` `product-name` and `platform-domain`. |
| `web/app/globals.css` | rendered, **customized** | A + B | Imports, the ejected block, palette sources (A), resolution and mode blocks (B), `@theme inline` mappings, non-colour tokens, the `angled` utility, the base layer, the existing `body` rule, the appearance-control selected-state rules, and reduced motion |
| `web/lib/tokens.ts` | new | B | `SEMANTIC_COLOR_TOKENS` (16), `NON_COLOR_TOKENS`, and the types `TokenOverrides = { light?: Partial<Record<…,string>>; dark?: … }` |
| `web/lib/appearance.ts` | new (shared, not server-only) | D | `APPEARANCE_MODES = ["light","dark","system"]`, `APPEARANCE_STORAGE_KEY = "appearance-mode"`, `parseAppearanceMode()`, and `appearanceInitScript` (string) |
| `web/lib/utils.ts` | new (generated) | — | `cn` |
| `web/components.json` | new (generated) | — | shadcn configuration |
| `web/components/ui/button.tsx`, `dialog.tsx` | new (generated + E1–E7) | — | Primitives |
| `web/components/ui/portal-container.tsx` | new (client) | C | `PortalContainerContext` and `usePortalContainer()` |
| `web/components/layout/token-scope.tsx` | new (client) | C | `TokenScope({ tokens?, label?, className, children })` (§7.6) |
| `web/components/layout/appearance-control.tsx` | new (client) | D | `useAppearance()` (via `useSyncExternalStore`) and `AppearanceControl` (§8.4) |
| `web/components/layout/shell-help-dialog.tsx` | new (client) | — | Generic keyboard-help dialog (trigger, title, description, close) |
| `web/components/layout/app-shell.tsx` | new (server) | A (layout) | Generic frame (§6.2) |

No `components/domain`, `features`, `hooks` or `public` directory is created. No new route or API handler.

### 6.2 Shell anatomy (applies §9; all copy generic; identity only from configuration)

```text
<TokenScope label="default">                                   ← root scope; min-h-screen; bg-background
  <a href="#main">Skip to main content</a>                     ← first focusable; visible on focus (accent outline)
  <header>  border-b-2 rule · container 1360px · gutter clamp(20px,4vw,40px) · flex-wrap
    {productName}          font-display uppercase tracking (text identity; no logo)
    <nav aria-label="Primary"> Overview · Getting around · About    font-display uppercase .94rem tracking 1.2px
    <AppearanceControl/>   fieldset “Appearance”: Light | Dark | System (2px ruled segmented group)
    <ShellHelpDialog/>     angled primary button “Keyboard help”
  </header>
  <main id="main" tabIndex={-1}>
    <section id="overview">  border-b-2 · padding clamp(3rem,8vw,5.5rem)
      kicker “Platform foundation”   (font-label 600 uppercase tracking 2.5px · accent)
      <h1>{productName}</h1>         (font-display 700 uppercase · clamp(2rem,5.6vw,5rem) · lh 1 · −1px · wraps)
      lead: generic sentence naming {productName} and {platformBaseDomain}   (max 52ch)
    <section id="getting-around">  <h2> + a ruled 3-cell grid (2px vertical rules ≥ md; stacked with 2px horizontal rules below)
      cell: index “01” (mono, muted-foreground) · <h3> in accent · body (muted-foreground)   ×3: skip link, keyboard help, appearance
    <section id="about">  <h2> + a short generic paragraph + list with 14×2px accent rule bullets
  </main>
  <footer> border-t-2 rule · {platformBaseDomain} (font-label meta, muted-foreground)
</TokenScope>
```

**Adaptations in this layout:**
- No sticky header, so focus is never obscured.
- The grid reflows to one column, where the skill's grids scroll horizontally.
- Inline body links are underlined, so meaning is not carried by colour alone.
- No orange emphasis word, no ribbon and no node field (T6).

## 7. Semantic tokens, palettes and overrides

### 7.1 Semantic colour tokens: exact contract `SEMANTIC_COLOR_TOKENS` (16)

- **Retained (7):** `background`, `foreground`, `primary`, `primary-foreground`, `accent`, `muted-foreground`, `border`.
- **Added (9), each needed by an installed component or the shell:** `popover` and `popover-foreground` (dialog popup); `muted` (ghost and outline hover, dialog footer, band surface); `secondary` and `secondary-foreground` (secondary variant); `destructive` (destructive variant, `aria-invalid`); `input` (outline variant boundary, E5); `ring` (focus outline); `overlay` (dialog backdrop, E6).

The same 16-name set appears in five places, which B-WEB-TOKENS checks for equality:

1. the `SEMANTIC_COLOR_TOKENS` list;
2. `--light-<t>` on `:root`;
3. `--dark-<t>` on `:root`;
4. the resolution declarations in both modes;
5. `@theme inline --color-<t>`.

### 7.2 Non-colour tokens and derived mappings (separate exact coverage)

| Kind | Name | Value | Source |
|---|---|---|---|
| `:root` design token | `--radius` | `0` | Skill: zero radius everywhere |
| `:root` design token | `--angle` | `polygon(15% 0%,100% 0%,100% 35%,85% 100%,0% 100%,0% 65%)` | `--vx-angle` |
| derived `@theme inline` | `--radius-sm`, `--radius-md`, `--radius-lg`, `--radius-xl`, `--radius-2xl`, `--radius-3xl`, `--radius-4xl` | the CLI's `calc(var(--radius) * k)` multiples (all 0) | Registry mapping |
| `@theme` | `--font-display` | `"Oswald", "Arial Narrow", "Roboto Condensed", ui-sans-serif, system-ui, sans-serif` | `--vx-font-display` (local fallbacks only) |
| `@theme` | `--font-label` | `"Barlow Condensed", "Arial Narrow", "Roboto Condensed", ui-sans-serif, system-ui, sans-serif` | `--vx-font-label` |
| `@theme` | `--font-sans` | `"Roboto", ui-sans-serif, system-ui, sans-serif` | `--vx-font-body` |
| `@theme` | `--font-mono` | `ui-monospace, "SFMono-Regular", Menlo, monospace` | `--vx-font-mono` |
| derived `@theme inline` | `--font-heading` | `var(--font-display)` | Used by the generated `DialogTitle` |
| `@theme` | `--container-shell` | `1360px` | `--vx-container` |
| `@theme` | `--spacing-gutter`, `--spacing-section` | `clamp(20px,4vw,40px)`, `clamp(3rem,8vw,5.5rem)` | `--vx-gutter`; `--vx-section-y` adapted (88px is reached only on wide viewports) |

The rule width (2px) uses Tailwind `border-2`, `border-b-2` and similar utilities: it is structural, not a token. There is no `@font-face`, no font binary and no remote import.

### 7.3 Palettes (concern A values; hex sRGB; ratios computed in planning, verified by rendered-pixel tests)

| Token | Light (derived) | Dark (from skill) | Source mapping |
|---|---|---|---|
| background | `#fafaf9` | `#0d0d0d` | `--vx-bg` (screenshot pixel `#0d0d0d`) |
| foreground | `#1a1918` | `#e8e6e3` | `--vx-text` |
| muted | `#f0efed` | `#101010` | `--vx-surface` band (screenshot `#101010`) |
| muted-foreground | `#5f5a54` | `#a8a39d` | `--vx-text-3` |
| popover / popover-foreground | `#ffffff` / `#1a1918` | `#131313` / `#e8e6e3` | `--vx-raised` (screenshot `#131313`) |
| primary / primary-foreground | `#5952ff` / `#ffffff` | `#5952ff` / `#ffffff` | Rendered CTA in all three screenshots (`#5952ff`); see Q6 |
| secondary / secondary-foreground | `#4d6800` / `#ffffff` | `#6a8e00` / `#0d0d0d` | Rendered alternate CTA (`#6a8e00`); label colour adapted |
| accent | `#8a5a12` | `#f8ce8a` | `--vx-gold` (kickers, links, bullets); light bronze derived |
| ring | `#8a5a12` | `#f8ce8a` | Skill focus: 2px gold |
| border | `#d9d6d2` | `#262626` | `--vx-rule` (screenshot `#262626`); decorative separators, not control boundaries |
| input | `#8f8a84` | `#6f6a64` | `--vx-muted` / `--vx-faint`, reused as a control boundary |
| destructive | `#b3261e` | `#ff8a7a` | Not in the skill; an accessible error red |
| overlay | `rgb(13 13 13 / 0.45)` | `rgb(0 0 0 / 0.7)` | Dims the ground; no blur |

**Planning-time ratios**, light then dark:
- Body text on background: 16.81 / 15.60. On muted: 15.28 / 15.28.
- Muted-foreground on background: 6.53 / 7.77. On muted: 5.94 / 7.60. On popover: 6.82 / 7.42.
- Popover text: 17.56 / 14.92.
- Primary-foreground on primary: 5.15 (both). Secondary label: 6.37 / 5.07.
- Accent on background: 5.66 / 13.13. On popover: 5.91 / 12.55.
- Ring against background, popover and muted: at least 5.15 / 12.55.
- Primary fill against background: 4.93 / 3.77. Input boundary: 3.28 / 3.63.
- Destructive: 6.26 / 8.49.

**Rejected source values:**
- `#0075FF` (declared in `styles.css`): white text on it is 4.21:1, below 4.5:1.
- `--vx-faint` `#6f6a64` for small index text: 3.63:1. Indices use `muted-foreground` instead.

### 7.4 Changes from the inherited values (R3)

| Token | Inherited | Now | Note |
|---|---|---|---|
| background | `oklch(1 0 0)` | L `#fafaf9` / D `#0d0d0d` | Warm-neutral ground, as in the skill |
| foreground | `oklch(0.21 0.01 260)` | `#1a1918` / `#e8e6e3` | Warm ink |
| primary | `oklch(0.45 0.12 260)` | `#5952ff` | Skill CTA as rendered |
| primary-foreground | `oklch(0.98 0 0)` | `#ffffff` | |
| accent | `oklch(0.94 0.02 260)` (a light tint) | `#8a5a12` / `#f8ce8a` | **Role change**: from a subtle surface tint to the emphasis ink (kickers, links, bullets). No installed component uses `bg-accent`; a future component that does must add `accent-foreground` (documented). |
| muted-foreground | `oklch(0.5 0.02 260)` | `#5f5a54` / `#a8a39d` | |
| border | `oklch(0.9 0.01 260)` | `#d9d6d2` / `#262626` | Skill rules |

The preset radius (0.625rem) becomes 0.

### 7.5 Resolution and mode (concern B)

```css
:root {                               /* concern A: palette sources (both appearances) */
  --light-background: #fafaf9;  --dark-background: #0d0d0d;   /* … all 16 tokens … */
  --radius: 0;  --angle: polygon(…);
}
:root, [data-token-scope] {           /* concern B: resolve at the root and at every scope boundary */
  --background: var(--light-background);   /* … all 16 … */
  color-scheme: light;
  @variant dark {
    --background: var(--dark-background);  /* … all 16 … */
    color-scheme: dark;
  }
}
@custom-variant dark {                /* appearance-aware; System works without JS */
  &:where([data-appearance="dark"], [data-appearance="dark"] *) { @slot; }
  @media (prefers-color-scheme: dark) {
    &:where(:root:not([data-appearance="light"]), :root:not([data-appearance="light"]) *) { @slot; }
  }
}
@theme inline { --color-background: var(--background); /* … all 16 … */ }
```

- **Effective appearance:** Dark when `data-appearance="dark"`, or when the OS prefers dark and the attribute is not `light` (including `system` and absent). Otherwise Light.
- **Live System tracking is CSS-only.** No `matchMedia` listener is needed for the palette.
- **Generated `dark:` classes follow the same variant.**

### 7.6 Organization-level overrides (concern C)

- **`TokenScope`.** `TokenScope tokens={{ light: { primary: "…" }, dark: { primary: "…" } }}` renders `<div data-token-scope={label} style="--light-primary:…; --dark-primary:…">`, plus an empty `<div data-portal-container>` provided through `PortalContainerContext`.
- **Resolution follows the mode.** Overrides set only *source* variables. Because each `[data-token-scope]` re-resolves the semantic tokens (§7.5), its subtree picks up its own sources for the active appearance. A missing mode value falls back to the inherited source; for example, a light-only override leaves the platform dark source in place.
- **Portals.** The dialog portal renders inside the scope (E7), so backdrop, popup, text, close icon and focus outline follow the scope in both appearances and across a live mode change.
- **Labels.** The scope label is a styling and test label only: never identity, authority or a selector of Organizations.
- **Fixtures** (`web/tests/fixtures/token-scopes/fixtures.ts`) are test data only:
  - `fixtureA` (cool) and `fixtureB` (warm) each define light **and** dark sources for surface, text, primary, primary-foreground, muted-foreground, accent, ring, popover and popover-foreground.
  - Each meets the §11 contrast pairs in both appearances and differs from the defaults and from the other fixture.
- **Deferred to G017:** authorized Organization data keyed by Organization UUID, and validation of untrusted values (for example rejecting `url(…)`).

## 8. Appearance mode architecture (concern D, R8)

### 8.1 Storage and values

- The key `appearance-mode` holds `"light" | "dark" | "system"` in `localStorage`. The selected *mode* is stored, never the resolved palette.
- A missing, invalid or unreadable value means System.
- The preference has no account, Organization, cookie, schema or server meaning. B-WEB-BOUNDARY forbids `document.cookie` and `cookies()` in web.

### 8.2 First paint (server and client responsibilities; r3 correction 1)

- **Mechanism (mandatory).** The root layout renders a classic, synchronous inline `<script>` in `<head>`: no `src`, `async`, `defer` or `type=module`. Under HTML parsing rules, a parser-blocking inline script runs when the parser reaches it, **before the parser creates any `<body>` content**. So no theme-bearing node exists, and none can be painted, before `data-appearance` is set.
  - The palette is pure CSS keyed on that attribute (§7.5), and the stylesheet is render-blocking. The first frame that contains body content therefore already uses the stored mode.
  - Framework JS (Next's `async` chunks) and hydration play no part in the initial palette.
- **Not used:** `next/script` `beforeInteractive`. Next documents ordering relative to framework code and hydration, not parser-time execution before paint. There is no weaker fallback: if the inline head script cannot be emitted as specified, stop and return to Codex (§21).
- **Init script** (about 300 bytes, from `lib/appearance.ts`; every storage access in `try`/`catch`): it reads `localStorage["appearance-mode"]`. If the value is valid it sets `document.documentElement.dataset.appearance` to it; otherwise to `system`. Unavailable or throwing storage gives `system` silently. Its SHA-256 is documented in WEB-SHELL.md for future CSP hash-allowlisting (G012).
- **Server.** It renders no mode-dependent markup: `<html lang="en" suppressHydrationWarning>`, where React serializes only `lang`. Without JS, the CSS media rule gives System behaviour.
- **Client after hydration.** `AppearanceControl` reads the attribute through `useSyncExternalStore` (server snapshot `system`) and corrects the radio `checked` state. The palette never depends on hydration.
- **Hydration (narrow, Q7).** React never renders `data-appearance`. `suppressHydrationWarning` appears **exactly once** in web source, on `<html>`; it is checked by B-WEB-BOUNDARY, and no descendant uses it. Because that suppression can hide any `<html>` attribute mismatch, S-HTML-ATTRS separately asserts `<html>`'s exact attribute set and values. S-HYDRATION-DEV (dev) and S-ERRORS (production) still apply.
- **Verification layers:**
  - **S-HEAD-ORDER:** the served HTML contains the script as specified, before `</head>`, with the documented hash.
  - **S-FIRSTPAINT:** paint-timed, visible-content-synchronized frame evidence.
  - **S-FIRSTPAINT-DELAYED:** with framework JS held, so hydration cannot occur, the painted pixels show the stored appearance opposite the OS.

### 8.3 System changes

No JavaScript is needed: the CSS media query applies live. An explicit Light or Dark overrides it, because the attribute wins in the variant (§7.5). Returning to System sets the attribute to `system`, so the palette follows the current OS preference immediately.

### 8.4 Control (`AppearanceControl`)

- **Markup.** A native `<fieldset>` with the legend “Appearance” (visually hidden but announced). It contains three `<input type="radio" name="appearance">`, labelled Light, Dark and System. They are visually styled as a 2px-ruled segmented group, using an sr-only input inside each label.
- **Keyboard.** Tab enters the group; the arrow keys move and select (native behaviour).
- **Focus.** The label shows an `outline-ring` outline when the input has `:focus-visible`.
- **Selected state.**
  - The visual is driven by CSS from `html[data-appearance]`, so it is correct at first paint: a `primary` fill plus a leading 14×2px bar, which is a shape cue as well as colour (Vestive bullet motif).
  - The `checked` state comes from `useAppearance()`. It uses `useSyncExternalStore`, with server snapshot `system` and client snapshot from the attribute, so there is no hydration mismatch; React corrects `checked` immediately after hydration.
- **Change.** Choosing a mode sets the attribute, then calls `localStorage.setItem` inside `try`/`catch`, so an unavailable store still allows in-session selection without errors.
- **Not included:** no cross-tab sync, which is not required (documented).

## 9. Design-language application (concern A, R9/AC10)

### 9.1 Design-input inventory (`skills/vestive-design-skill`, product-owner input; unchanged during G002)

| File | SHA-256 | Use in G002 |
|---|---|---|
| `SKILL.md` | `46ed33b87d6fcbaa9dd012f2da2717652841450d5f717fe5b0bae049bd4ba77b` | Entry point. Read; its instruction to link `styles.css` and `motion.js` is **adapted**: values are mapped into tokens (T1) |
| `README.md` | `80d514a96d99540b0791dd5d2f79ef7e6c5d36bca4845b4e631ed0343e843231` | **Visual foundations, signature motion (omitted), iconography and Do/Don't**: applied (§9.2). The **content fundamentals, pricing, voice, nav labels and CTAs are ignored**, as is the dark-only ground rule. |
| `styles.css` | `b6d41de3df552af95e90599ec28c320e3c88a4c5486d63828414efa7fbfff73a` | Token values, type sizes and tracking, angle polygon, focus rule, and button hover and press. Not imported (it has a remote Google Fonts `@import` and global dark overrides). |
| `motion.js` | `6ad3c9876312524f94231f3d955243bb168c8daae1c9a7c75b3f61149f0e1adf` | Read; ribbon and node field **omitted** (T6) |
| `examples/reference.html` | `d0dc04398ec296def6caea94419d39b4f77265764122fa97c51fc12b2512b3f3` | Structural reference: ruled sticky header (made non-sticky), kicker → H1 → lead, ruled cell grid with indices, angled CTA and gold link. Marketing text ignored. |
| `examples/Screenshot 2026-10-05 103447.png` | `4de5ad8e3ffca75b69906081e29672aca2a178f3d4b762c70ce92264ddfd8e38` | **Visual reference**: risk grid plus plans grid (2205×1300) |
| `examples/Screenshot 2026-10-05 103502.png` | `fb140500b42267cf18fa28257aad14f065695aee494c877455dfda9bc2dc9589` | **Visual reference**: header plus hero plus risk grid (2212×1282) |
| `examples/Screenshot 2026-10-05 103548.png` | `6574fb356426b3d70958fcf20c1a62d2325deddd1a0a6d9cc1891892619710f2` | **Visual reference**: plans grid plus closing band (2210×1291) |
| `assets/logo.png` | `b47c48709ca4e22cab989614fc76f8465e6a4dfe3b1fb54f4e77bb27694d4445` | **Not used**: Vestive mark; identity stays configured |
| `assets/icon.png` | `49b579a4c12badca9d53aa2b30b27f4cdb314460b3769457d261ea795b74386a` | **Not used** |
| `assets/subsidiary.png` | `cdfb8e47c21940e5f9dd88f33623ea6e0c20ed0ee41cc18d5df6adc0dd83537b` | **Not used**: VitruvianTech mark |
| `assets/icons8-instagram-50.svg` | `a0c0bacbdda14b4bbe3a58e5003d0939d7594c26d13cf4b901f0bfd22ee566b7` | **Not used**: there are no social links in a generic shell |
| `assets/icons8-linkedin-50.svg` | `73953b384aacff945715d51a8df8e0892b4a1a57a898dfa74cac9db2027e9f10` | **Not used** |
| `assets/icons8-tiktok.svg` | `80910a86cad3c46d1bc37f9722f959e3859472ef98449372a40367c357ece932` | **Not used** |
| `assets/icons8-twitterx-50.svg` | `98b61c17327180fa52e4c756edacf3ada0149d5f8efbe12d36f3c3ab42b2b936` | **Not used** |
| `assets/icons8-youtube-50.svg` | `f08f9e2b77656b329c1c79ecef17ff5594038c517d1fb5c207ef7c168991253c` | **Not used** |

These hashes are as of r2 writing. S0 re-records them; if the files are still uncommitted, they are bound with `--expect-hash` (§16–§17).

**No skill file is copied into `web/`.** Only values (colours, sizes, tracking, polygon) are adapted, and each is attributed in §7 and §9.2. B-WEB-DESIGN asserts that no web file is byte-identical to any skill asset.

### 9.2 What the screenshots show, and how the shell applies it

| Observed in the screenshots (and `README.md`/`styles.css`) | Applied in the shell | Light adaptation |
|---|---|---|
| Near-black ground `#0d0d0d`, bands `#101010`, raised `#131313` (pixel-sampled) | `background`, `muted` (band) and `popover` (dialog) dark sources | Warm off-white ground, very light band, white popover |
| Thin 2px rules under the header, between sections and between cells (`#262626`) | `border-b-2`, `border-t-2` and `border-2` with `border-border`, on the header, every section, the footer, grid cells and the appearance group | `#d9d6d2` rules |
| Flush-left hierarchy: gold uppercase kicker, then a huge uppercase condensed H1, then a lead paragraph at about 52ch | Overview section: kicker “Platform foundation”, H1 = configured product name, generic lead | Same structure, bronze kicker |
| Header: mark left, uppercase condensed nav, CTA right | Product-name text left (no logo), uppercase display nav, appearance control, and an angled “Keyboard help” button right; wraps at narrow widths | Same |
| Ruled grid cells with small mono indices `01`/`02`/`03` and gold uppercase cell headings | The “Getting around” grid: three cells, indices in mono (`muted-foreground` for contrast), H3 in `accent` | Same with bronze; it stacks below `md` instead of scrolling |
| Short gold dash bullets (14×2px) | The About list's bullet bars, drawn as `border-t-2 border-accent` rules, never `bg-accent` (Q10); also the selected-state cue in the appearance control (`currentColor`) | Bronze bars |
| Angled CTA silhouettes in solid violet-blue (`#5952ff`) and green (`#6a8e00`) | The `angled` pseudo-element fill on filled Button variants (E3): primary `#5952ff`, secondary green | Same blue; darker green with white label |
| Gold underlined uppercase text link | The `link` variant uses `accent` (E4); standalone links use a 2px accent bottom rule | Bronze |
| Zero radius, no shadows, no blur or glows | `--radius: 0`; the dialog has no shadow (`ring-1 ring-foreground/10` hairline only) and no backdrop blur (E6) | Same |
| Restrained colour: mostly ink on ground | Accent is used only for kicker, cell headings, bullets, links and focus. Primary is used only for the filled action and the selected appearance. | Same |

### 9.3 Adaptations and omissions (with reasons)

| # | Skill instruction | G002 treatment | Reason |
|---|---|---|---|
| T1 | “Link `styles.css` from every page; load `motion.js`” | Values mapped into semantic tokens and theme; neither file is linked or copied | R9: no wholesale import (remote font, global dark overrides, raw colours) |
| T2 | Dark ground only | Light/Dark/System with System the default; the dark palette is derived from the skill and the light palette from its roles | Product-owner clarification, R8 |
| T3 | Oswald, Barlow Condensed and Roboto from Google Fonts | Same family names first in local-only stacks, with system fallbacks; no download | No external fonts or assets. Screenshots in the test image show fallback faces (recorded as a limitation). |
| T4 | Grids keep their column count and scroll horizontally | Grid reflows to one column below `md`, with horizontal 2px rules | R2 reflow at 320 px |
| T5 | Primary CTA `#0075FF` (`styles.css`) | `#5952ff`, as rendered in the screenshots | White on `#0075FF` is 4.21:1; `#5952ff` reaches 5.15:1. Q6. |
| T6 | Logo ribbon and node-field motion | Omitted | Not required (R9); bounded scope; no motion beyond button transitions and dialog fades |
| T7 | Orange emphasis word in headings | Omitted | The H1 is configured identity; there is no marketing emphasis |
| T8 | Faint index colour `#6f6a64` | `muted-foreground` | 3.63:1 is below 4.5:1 for small text |
| T9 | MDI icons | Generated lucide `X` kept | R9: do not add an icon dependency only to match |
| T10 | Angled silhouette on every button | Filled variants only. Ghost, icon, link and outline stay rectangular, and the angle is never applied to the clickable element itself (pseudo-element) | Square icon buttons would clip the glyph; R9 requires no clipped labels or focus outlines |
| T11 | Sticky header in the reference | Not sticky | Focus not obscured (2.4.11) |
| T12 | Uppercase via CSS | Kept (source text in sentence case) | Screen readers read the source text |
| T13 | Logo, social icons, marks, all marketing copy, pricing, testimonials, CTAs and nav labels | Not used | R9 generic content and configured identity. B-WEB-DESIGN checks for Vestive and marketing literals. |

## 10. Docker and runtime implications

- **Unchanged:** the `web` Dockerfile and its `.dockerignore`, the runtime image, `local.yml`, and the core and intelligence images.
- **`web-check`** (`npm run check`, `network_mode: none`) now lints and type-checks the new source, tests and fixtures.
- **New `infra/docker/web-browser.Dockerfile`** (context `web/`). Its `.dockerignore` excludes `.git`, `node_modules`, `.next`, `.env`, `.env.*`, `test-results` and `playwright-report`. Stages:
  1. **`fixture`** (`node:24-slim`): `npm ci`, copy the source, copy `tests/fixtures/token-scopes/page.tsx` into `app/token-fixtures/page.tsx`, and `npm run build`. The test-only route exists only in this stage.
  2. **`browser`** (the pinned Playwright image): `npm ci`, copy the source, and copy the fixture standalone output to `/fixture`. Then `chown -R pwuser /app` so the dev hydration server can write `/app/.next`, and switch to `USER pwuser`. The command is `npx playwright test`.
- **New overlay `infra/compose/browser.yml`.** It defines `web-browser` with profile `browser`, `network_mode: "service:web"`, `init`, `shm_size: 1gb`, no ports, and environment exactly `WEB_BASE_URL=http://127.0.0.1:3000`, `EXPECTED_PRODUCT_NAME=${PRODUCT_NAME:?}` and `EXPECTED_PLATFORM_BASE_DOMAIN=${PLATFORM_BASE_DOMAIN:?}`.
- **Command** (README block `platform:browser`, with the stack running): `docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/browser.yml --profile browser run --rm --build --no-deps web-browser`.
- **Servers inside the runner.** Playwright `webServer` starts three servers on loopback in the shared namespace; none is published:

  | Port | Server | Environment |
  |---|---|---|
  | `127.0.0.1:3001` | Fixture production server | Neutral `PRODUCT_NAME=Fixture Product`, `PLATFORM_BASE_DOMAIN=fixture.invalid`, `CUSTOM_DOMAIN_CNAME_TARGET=custom.fixture.invalid`, `CORE_INTERNAL_URL=http://127.0.0.1:9` |
  | `127.0.0.1:3002` | Fail-closed fixture server | As 3001, but empty `PRODUCT_NAME` |
  | `127.0.0.1:3003` | Dev hydration server | `next dev` on `/app` source, with the fixture identity |

  No test needs egress. Any cross-origin request fails S-BOUNDARY.

## 11. Browser validation matrix

### 11.1 Configuration

`web/playwright.config.ts`:
- **Runner:** Chromium (version recorded in the Result), `forbidOnly`, `retries: 0`, `workers: 1`, `deviceScaleFactor: 1`.
- **Reporters:** `list` plus `json`, writing to `PLAYWRIGHT_JSON_OUTPUT_FILE`, with `outputDir` from `PLAYWRIGHT_OUTPUT_DIR` (defaults under `/tmp`).
- **Artefacts:** screenshots are attachments; there are no traces.
- **Projects:**

  | Project | baseURL | Emulation | Specs |
  |---|---|---|---|
  | `shell-light`, `shell-dark` | `WEB_BASE_URL` (the real running web) | `colorScheme` light / dark; empty storage, so System | `shell`, `overlay`, `accessibility` (shell), `boundary`, `design` |
  | `tokens-light`, `tokens-dark` | `http://127.0.0.1:3001` (fixture build) | `colorScheme` light / dark | `tokens`, `accessibility` (fixture) |
  | `appearance` | `WEB_BASE_URL` | Per test | `appearance` |
  | `dev-hydration` | `http://127.0.0.1:3003` | Per test | `hydration` (part of `appearance.spec.ts`) |

Every test title carries a scenario ID. The harness requires each ID to **pass in each listed project**, with 0 skipped, 0 flaky and 0 unexpected (§15). Viewports: 320×640, 768×1024, 1280×800, 640×400 (equivalent to 1280×800 at 200% zoom), and text at 200% (`html { font-size: 200% }` at 1280 wide).

### 11.2 Scenarios

| ID | Projects | Assertions |
|---|---|---|
| S-IDENTITY `@identity` | shell-* | `document.title`, the header identity and the H1 equal `EXPECTED_PRODUCT_NAME`; the visible domain equals `EXPECTED_PLATFORM_BASE_DOMAIN` |
| S-LANDMARKS | shell-* | One banner, navigation “Primary”, main and contentinfo; one H1; H2/H3 order with no skipped levels; the fieldset “Appearance” has its legend; no element named login, sign in, account, organization, tenant or switch |
| S-NAV | shell-* | Each nav link, activated by keyboard, updates the hash and brings an existing section into view |
| S-SKIP | shell-* | The first Tab focuses the visible skip link (not clipped, in the viewport); Enter moves focus to `main#main`; the next Tab continues from main |
| S-TAB | shell-* | Exact Tab and Shift+Tab order: skip, three nav links, the appearance radio group (one stop), help trigger. Each stop has an unobscured, visible focus outline (`elementFromPoint`); target at least 24 px |
| S-OVERLAY | shell-* | Enter and Space open `role=dialog`, `aria-modal`, named; focus inside; Tab and Shift+Tab ×10 stay inside; background inert; Escape and the “Close” control close it and return focus to the trigger |
| S-MOTION | shell-* | Under reduced motion, popup, backdrop and button transitions compute ≤ 0.01 s; without it, the animations exist |
| S-ERRORS | all except dev-hydration | Zero `pageerror` and `console.error` across every scenario |
| S-RESPONSIVE | shell-* | At every §11.1 viewport: `scrollWidth ≤ clientWidth`; no clipped or overlapped controls; the grid is one column below `md` and three columns at or above `lg`; the H1 wraps with no overflow; the dialog fits with Close reachable; S-SKIP and S-TAB visibility rerun at 320 |
| S-A11Y | shell-*, tokens-* | axe (`wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa`, `wcag22aa`) reports zero violations: the shell closed and with the dialog open, at 1280 and 320; the fixture page in all scopes, with dialogs open in A and B. Only a selector-scoped exclusion with evidence; never a global disable. |
| S-CONTRAST | shell-*, tokens-* | Rendered-pixel method (§11.3), for each resolved appearance and each scope (default, A, B). Text pairs at ≥ 4.5 (or 3.0 only for qualifying large text): body and lead text on page and band; muted text on its actual surface (page, band, cell, popup, footer); kicker, accent headings and links on their surfaces, at rest and on hover; nav links at rest, hover and focus; each Button variant's label on its rendered fill at rest, hover and focus; the appearance labels on selected and unselected fills; dialog title and description on the popup. Non-text pairs at ≥ 3.0: every focus outline against the adjacent surface; the close icon against the popup; the outline-variant boundary against its surface; the selected appearance indicator against the unselected and the page; the primary fill against the page. |
| S-TOKENS | tokens-* | For default, A and B, the computed resolved token values (`getPropertyValue('--primary')`, …) equal each scope's source for the active appearance. Rendered surface, text, action and focus colours equal the expected values. A differs from B and from default; the adjacent default scope equals `:root`. |
| S-PORTAL | tokens-* | A dialog opened in A mounts inside A (`closest('[data-token-scope]')`), and its backdrop, popup, text, close icon and focus outline use A's values; likewise B; default shows default. **Live mode change while open:** emulating the other `colorScheme` re-resolves the popup to A's other-mode source, and the popup stays in A. |
| S-RUNTIME | tokens-* | Setting `--light-primary` on A's element changes A's button in Light immediately (no rebuild), leaves A's Dark source unchanged after switching mode, and leaves B and the defaults unchanged |
| S-APPEARANCE-DEFAULT | appearance | With no storage, each `colorScheme` gives `data-appearance="system"`, System checked, and computed `color-scheme` and palette matching the OS preference |
| S-APPEARANCE-LIVE | appearance | In System, emulating dark → light → dark while open changes the palette and `color-scheme` each time, with no reload and no errors |
| S-APPEARANCE-OVERRIDE | appearance | Choosing Light (or Dark) by keyboard with the OS on the opposite preference applies it; a later OS change does not alter it; storage holds `light` (`dark`) |
| S-APPEARANCE-RETURN | appearance | Choosing System applies the current OS preference immediately; storage holds `system` |
| S-APPEARANCE-PERSIST | appearance | For each mode, reload keeps the mode, the checked state and the palette; storage holds the mode, never the resolved palette |
| S-APPEARANCE-INVALID | appearance | Stored `purple`, empty or JSON values resolve to System with no error |
| S-APPEARANCE-NOSTORAGE | appearance | Through `addInitScript`, `localStorage` access throws. The page loads in System with no error; Dark selection works in-session; reload returns to System. |
| S-HEAD-ORDER | shell-light | The raw HTML response for `/` contains exactly one `script#appearance-init` inside `<head>`. It is classic (no `src`, `async`, `defer`, `type`, `nonce` or `type=module`), appears before `</head>` and therefore before any `<body>` content, and its content SHA-256 equals the hash documented in WEB-SHELL.md. |
| S-FIRSTPAINT | appearance | Probe of §11.4. Matrix: stored `light`, `dark` and `system` × OS `light` and `dark`, plus storage-unavailable × both. **Valid-sample requirement:** at least one frame sample with theme-bearing content present and laid out (header, H1 and root scope with non-zero boxes), timestamped at or before the first-contentful-paint time; `paint` (`first-contentful-paint`) and `largest-contentful-paint` entries exist; and the LCP element is theme-bearing content. A run without these **fails**; it is never skipped or passed. **Assertions:** `data-appearance` already equals the expected mode at the content-insertion record and in the first content frame. Every content frame from that point until hydration completes shows the expected resolved palette (computed root-scope background, H1 colour and `color-scheme`). A painted-pixel sample of the ground and header at the first stable frame equals the expected palette. The timeline (frames, FCP, LCP, insertion, hydration-complete time and pixel values) is attached as evidence. |
| S-FIRSTPAINT-DELAYED | appearance | Stored `light` with OS `dark`, and stored `dark` with OS `light`. `page.route` **holds every `/_next/static/**/*.js` response** (CSS is not held) until the test releases it. Navigation uses `waitUntil: "commit"`; the test waits until header and H1 are visible, then holds at least 2 s. **Hydration impossible:** the held-request log shows at least one framework chunk requested and none released. **Paint before hydration:** the FCP time is earlier than the release time. **Correct palette while delayed:** every probe frame and repeated 1×1 painted-pixel samples (ground, header, H1 ink, primary button fill) equal the *stored* appearance, and differ from the OS-preference palette. **After release:** hydration completes (the help dialog opens), the palette is unchanged, `checked` equals the stored mode, and there are no console errors. The hold, release, FCP, frame and pixel evidence is attached. |
| S-HTML-ATTRS | appearance | (Q7) In each stored mode × OS preference, after hydration, `<html>` has exactly the attributes `lang` and `data-appearance`: `lang="en"`, and `data-appearance` equals the expected mode. The served HTML's `<html>` has exactly `lang="en"`. No other element in the served or hydrated DOM carries a suppression marker (B-WEB-BOUNDARY covers the source). |
| S-HYDRATION-DEV | dev-hydration | Against `next dev`, `/` in each stored mode × each OS preference shows zero React hydration warnings and errors in the console (dev reports any mismatch outside `<html>`'s own attributes) |
| S-BOUNDARY | shell-* | All requests are same-origin; no internal host or port. HTML, JS and CSS bodies contain no internal URL or secret-variable names (`core:8180`, `intelligence:8000`, `CORE_INTERNAL_URL`, `DB_`, `REDIS_`, `OBJECT_STORE_`, `MINIO_`, `INTELLIGENCE_`) and no `fonts.googleapis`. Bodies are saved for the harness secret scan. |
| S-HEALTH | shell-light | `/api/health` returns 200 and `/api/ready` returns 200 `{"checks":{"core":"ok"}}`, both with no-store (503 under outage: B-OUTAGE) |
| S-NOFIXTURE | shell-light | Production `/token-fixtures` returns 404 |
| S-FAILCLOSED | shell-light | The 3002 server returns 500 for `/` with no identity or environment content; `/api/health` returns 200 |
| S-DESIGN | shell-* | Design language observable through computed styles (AC10): radius 0 on buttons and popup; 2px `border` rules on header, sections, footer and cells; H1, H2, H3, nav and buttons have `text-transform: uppercase` and a `font-family` starting with the display stack; the kicker uses the label stack with 2.5px tracking and `accent` colour; filled buttons' `::before` uses `clip-path` equal to `--angle`; the popup has no `box-shadow` and the backdrop `backdrop-filter: none`; bullets are 14×2px `accent` bars; container max-width 1360px; headings align to the start |
| S-ANGLE | shell-*, tokens-* | For every filled button, every corner of the label's bounding box lies inside the `--angle` polygon mapped to the button rect, and the focus outline's sampled pixels are present on all four sides |
| S-SCREENSHOTS | shell-*, tokens-* | Evidence attachments: the shell at 1280 and 320, with the dialog open, in Light and Dark; the fixture page with dialogs in A and B in both appearances |

### 11.3 Rendered-colour method (S-CONTRAST, S-TOKENS, S-ANGLE)

- **Backgrounds** are sampled from rendered pixels. A 1×1 `page.screenshot({ clip })` at points inside the element's padding area, away from text, is decoded in Node with `zlib` and no new dependency (`support/png.ts`). Pixels already composite every translucent layer, pseudo-element fill, filter (hover brightness), backdrop and real backing surface.
- **Text and icon colours** come from computed style, converted to sRGB through a canvas readback. Text and icon tokens must be opaque; if alpha is below 1, the value is composited over the sampled background.
- **Focus outline** pixels are sampled in the outline band (offset 2px plus width 2px) and compared with a sample just outside it.
- **Timing.** Hover and focus states are captured after their transition completes, or under reduced motion.
- **Ratios** use WCAG relative luminance (`support/color.ts`).

### 11.4 First-paint probe design (S-FIRSTPAINT, S-FIRSTPAINT-DELAYED)

`addInitScript` installs the probe at document creation, before any page script, including the inline init.

- **Paint timing.** `PerformanceObserver` (buffered) for `paint` (`first-paint`, `first-contentful-paint`) and `largest-contentful-paint`, which records `renderTime` and the LCP element. These are the browser's actual paint timestamps, not assumptions.
- **Content insertion.** A `MutationObserver` on the document records `performance.now()` and `data-appearance` when the theme-bearing elements (root scope, header, H1) are inserted.
- **Frame samples.** A `requestAnimationFrame` chain runs from document creation until hydration completes plus 1 s. Each callback runs in the rendering update for the frame about to be painted. It records the timestamp, `data-appearance`, whether theme-bearing content is present with non-zero boxes, and the computed `background-color`, `color` and `color-scheme` of the root scope and H1.
- **Correlation (r3 review note).** Each computed-style sample records its **actual observation time** (`performance.now()` taken inside the sampling code, at the moment `getComputedStyle` is read). That time is kept separate from the rAF callback's timestamp argument, and from PerformanceObserver *delivery* time; paint correlation uses the entries' own `startTime` and `renderTime`. The frame whose observation precedes FCP is identified explicitly; a frame merely being first is never treated as proof. Pixel samples (`page.screenshot` with 1×1 clips, decoded as in §11.3) add the painted result **during** the scenario. A later screenshot alone is never accepted as first-frame proof.
- **Hydration-complete marker.** The first moment the help-dialog trigger responds (the test clicks after release) or an `AppearanceControl` effect flag, whichever the test observes. It is used only to bound the sample window.

## 12. Accessibility validation (both resolved appearances)

| WCAG 2.2 AA area | Evidence (shell-light/dark, tokens-light/dark, appearance) |
|---|---|
| 1.3.1, 2.4.1, 2.4.6 structure, bypass, labels | S-LANDMARKS, S-SKIP, axe; fieldset and legend naming |
| 1.4.1 not colour alone | Underlined inline links; shape cue on the selected appearance; focus geometry (S-TAB, S-CONTRAST) |
| 1.4.3, 1.4.11 contrast | S-CONTRAST (rendered pixels, rest, hover and focus, default and two fixtures, both appearances); axe `color-contrast` |
| 1.4.4, 1.4.10 resize and reflow | S-RESPONSIVE (text 200%, zoom equivalent, 320 px) |
| 2.1.1, 2.1.2 keyboard, traps | S-TAB, S-NAV, S-OVERLAY, S-APPEARANCE-OVERRIDE (radio arrow keys) |
| 2.4.3, 2.4.7, 2.4.11 focus order, visible, not obscured | S-TAB, S-SKIP, S-OVERLAY (return focus), S-ANGLE (outline not clipped) |
| 2.5.8 target size | S-TAB bounding boxes ≥ 24 px |
| 4.1.2 name, role, value | S-OVERLAY, radio `checked` state after hydration, axe |
| Reduced motion | S-MOTION |
| Native colour scheme | S-APPEARANCE-*, computed `color-scheme` |

**Limitations stated in the Result:** Chromium only, automated and computed; no manual screen-reader test; no full-product certification (Q5).

## 13. Responsive validation

- **Viewports:** 320×640, 768×1024, 1280×800, 640×400 (zoom equivalent) and text 200%, in **both** appearances (S-RESPONSIVE).
- **Specific checks:**
  - header items wrap without overlap;
  - nav, appearance group and help button stay reachable;
  - the H1 wraps (`overflow-wrap: anywhere`, minimum 2rem);
  - the grid stacks below `md` and has three columns at `lg`;
  - section padding scales (clamp);
  - the dialog stays within `100% − 2rem`.
- **Controls:** the same layout runs on the fixture page at 320 for scope screenshots.

## 14. Organization override behaviour versus user preference (summary of separation)

| Question | Answer | Test |
|---|---|---|
| Does choosing Dark change an Organization's override values? | No. It switches every scope to its dark sources. | S-PORTAL (live mode change), S-RUNTIME |
| Does an Organization override change the mode? | No. Scopes cannot set `data-appearance`. | B-WEB-BOUNDARY (TokenScope never writes the attribute) and S-TOKENS |
| Is the preference stored per Organization or account? | No. Browser `localStorage` only, mode value only. | S-APPEARANCE-PERSIST, B-WEB-BOUNDARY (no cookie or server write) |
| Do design-language values live in components? | No. Palette sources and design tokens live in `globals.css`; components use semantic utilities. | B-WEB-TOKENS (no palette literals outside the palette block and fixtures) |

## 15. Project harness changes (`infra/validation/`)

Each change preserves or strengthens the check's meaning.

| Check | Change |
|---|---|
| **B-DEPS** (r1 correction 1; reason codes r3; independent rules r4) | **Two functions, no install or download** (`yq` on JSON):<br>**`deps_pair_check PKG LOCK`** evaluates every rule below **independently**. There is no priority scheme and no rule suppresses another. It prints the **sorted unique set** of `CODE name` lines, and fails if the set is non-empty.<br>`DEPS_EXISTING_MISSING <n>`: an existing entry (the 4 runtime and 8 dev dependencies in §5) is in neither category.<br>`DEPS_EXISTING_RECATEGORIZED <n>`: it is present in the other category.<br>`DEPS_EXISTING_CHANGED <n>`: it is in its category with a specifier other than its original exact version.<br>`DEPS_EXTRA_DIRECT <n>`: a direct dependency that is neither existing nor one of the 7 reviewed additions.<br>`DEPS_ADDITION_WRONG <n>`: a reviewed addition is absent from its category, present in the other category, or has a specifier other than its reviewed exact version.<br>`DEPS_NOT_EXACT <n>`: any direct specifier that is not an exact `x.y.z`.<br>`DEPS_LOCK_MISMATCH <n>`: one of three conditions: (a) the lockfile root `packages[""].dependencies`/`.devDependencies` specifier for `<n>`, or its category, differs from `package.json`; (b) for an exact specifier, `packages["node_modules/<n>"].version` differs from it, or the entry is missing; (c) the entry's `dev` flag disagrees with its category (dev-only dependencies `dev: true`, runtime dependencies without it).<br>`DEPS_FORBIDDEN shadcn`: a `shadcn` direct dependency.<br>**`deps_repo_check REPO`** runs `deps_pair_check` on the real pair, plus three repository-level rules:<br>`DEPS_NOT_REGISTERED <path>`: `web/package.json` or `web/package-lock.json` is missing from the customization register.<br>`DEPS_FORBIDDEN next/font`: a `next/font` import in `web/`.<br>`DEPS_IMAGE_MISMATCH`: the Playwright image tag differs from `@playwright/test`'s version, or the full `@sha256:` digest is missing.<br>Transitive packages are not restricted. Intelligence lockfile equality and the goose and MCP pins are unchanged. |
| **B-DEPS-CONTROLS** (r3 and r4 corrections) | Disposable inputs in `$WORK`, built from copies of the real files, with no repository edit, install or download. Each control asserts that the function **fails** and that its printed set **equals exactly** the expected set: no missing and no unexpected code.<br>**Manifest mutations (C1–C5, C9–C10)** are mirrored in the lockfile root (specifier and category) and in `packages["node_modules/<n>"]` (`version`, `resolved` and `dev`), so each pair is internally consistent. `integrity` is not part of the rules; real builds verify it through `npm ci`.<br>**Lock-only mutations (C6–C8)** leave `package.json` at the approved content.<br>• C1, existing version changed: `next` 16.3.8→16.3.9 in both files. Expect {`DEPS_EXISTING_CHANGED next`}.<br>• C2, extra direct package: `"left-pad": "1.3.0"` in `dependencies` in both files, plus the lock entry (`version` 1.3.0, no `dev`). Expect {`DEPS_EXTRA_DIRECT left-pad`}.<br>• C3, addition in the wrong category: `@axe-core/playwright` moved to `dependencies` in both files, with `dev` removed from its lock entry and from `axe-core`. Expect {`DEPS_ADDITION_WRONG @axe-core/playwright`}.<br>• C4, existing removed: `server-only` removed from both files. Expect {`DEPS_EXISTING_MISSING server-only`}.<br>• C5, range on an addition: `@base-ui/react` set to `^1.8.0` in both files, lock version unchanged. Expect {`DEPS_ADDITION_WRONG @base-ui/react`, `DEPS_NOT_EXACT @base-ui/react`}. The lock rule (a) compares the equal specifier strings, and (b) applies only to exact specifiers.<br>• C6, lock version only: `packages["node_modules/next"].version` (and `resolved`) set to 16.3.9. Expect {`DEPS_LOCK_MISMATCH next`}.<br>• C7, lock root specifier only: `packages[""].dependencies.next` set to `16.3.9`. Expect {`DEPS_LOCK_MISMATCH next`}.<br>• C8, lock dev flag only: `dev` removed from `packages["node_modules/@playwright/test"]`. Expect {`DEPS_LOCK_MISMATCH @playwright/test`}.<br>• C9, existing recategorized: `typescript` moved to `dependencies` in both files, with `dev` removed from its lock entry. Expect {`DEPS_EXISTING_RECATEGORIZED typescript`}.<br>• C10, forbidden package: `"shadcn": "4.21.1"` in `dependencies` in both files, plus a consistent lock entry. Expect {`DEPS_EXTRA_DIRECT shadcn`, `DEPS_FORBIDDEN shadcn`}.<br>• C11, image tag mismatch: `deps_repo_check` on a scratch copy of the repository-level inputs, with the Playwright `FROM` tag set to `v1.62.0`. Expect {`DEPS_IMAGE_MISMATCH`}.<br>• C12, positive control: the real pair and the real repository-level inputs. Expect the empty set; the function passes.<br>Each case's expected and actual sets are recorded. A missing or unexpected code fails B-DEPS-CONTROLS, and no rule may be relaxed to make a control pass. |
| B-SHELL | Also shellchecks `infra/validation/run-approved.sh` and `infra/validation/run-approved-controls.sh` |
| B-HADOLINT, B-DOCKERIGNORE | Add `web-browser.Dockerfile` and its `.dockerignore` |
| B-COMPOSE | Also validates `local.yml` + `browser.yml`. `web-browser` has exactly the three environment keys, `network_mode: service:web`, no ports and no secret value. `web`'s networks remain exactly `edge`. |
| B-NAMING | Additionally forbids `web/src`, `web/components/domain` and `web/public` |
| **B-WEB-BOUNDARY** (new) | Client files and `components/**` import none of `@/lib/config`, `@/lib/core-client`, `server-only` or `process.env`. `lib/config.ts` and `lib/core-client.ts` keep `import "server-only"`. There is no `NEXT_PUBLIC_`, `document.cookie` or `cookies()` in web. There is no absolute fetch in components or pages. Routes are exactly `/`, `/api/health` and `/api/ready`. `TokenScope` never references `data-appearance`. `lib/appearance.ts` uses only `localStorage`. `suppressHydrationWarning` occurs **exactly once** in `web/` source, on `<html>` in `app/layout.tsx` (Q7). `layout.tsx` imports neither `next/script` nor `next/font`, and renders `appearance-init` as a classic inline script (S-HEAD-ORDER is the runtime counterpart). |
| **B-WEB-TOKENS** (correction 2) | **Colour:** the 16-name set is equal across the five places in §7.1. **Non-colour:** `:root` holds exactly `--radius` and `--angle` among non-colour design properties. `@theme` holds exactly the §7.2 font, container and spacing names, and the derived radius and heading names map as specified. **Literals:** no palette utilities and no `#hex`, `rgb()`, `hsl()` or `oklch()` in `components/**` or `app/**/*.tsx`; colour literals appear only in the `globals.css` palette block and `tests/fixtures/**`. **Accent role (Q10):** no `bg-accent`, `from-`/`via-`/`to-accent`, `fill-accent` or `accent-foreground` in `components/**` or `app/**` (generated sources included). Accent appears only as text, border and outline ink. |
| **B-WEB-COMPONENTS** (new) | `components.json` matches §4.2. `components/ui` contains exactly `button.tsx`, `dialog.tsx` and `portal-container.tsx`. The eject marker is present. The generated sources import `@base-ui/react`. Edits E1–E7 are present (pattern checks). |
| **B-WEB-DESIGN** (new) | No remote fonts or assets in web: no `fonts.googleapis`, `@import url(`, remote `@font-face` or `http(s)://` asset URL. No web file is byte-identical to any `skills/vestive-design-skill/assets/*` file. No Vestive identity or marketing literals in `web/`, case-insensitive: `vestive`, `vitruvian`, `interlock`, `restruct`, `strategic advisor`, `principal engineer`, `associate architect`, `book a free assessment`, `$2,000`, `$5,000`, `$8,000`. |
| **B-WRAPPER-CONTROLS** (r4 correction 2; r5 bootstrap order) | Implemented by the committed `infra/validation/run-approved-controls.sh`. It is created at S0b, runs standalone at S0c before S1, and runs again from `run.sh`'s static suite (wired at S5). It exercises `infra/validation/run-approved.sh` (§17) in a scratch tree outside the repository: a copy of the wrapper, a **stub** `infra/validation/run.sh`, scratch input files (including names with spaces) and a scratch snapshot. The stub records each invocation to a marker file, exits with `STUB_RC`, and can optionally modify one input file or the snapshot during its run (`STUB_MUTATE`). Each case asserts the wrapper's exit status and whether the stub ran:<br>• W1, all valid: exit 0; the stub ran once; pre- and postcheck passed.<br>• W2, snapshot digest differs from the pinned `--snapshot-sha256`: exit 10; the stub did **not** run.<br>• W3, an input file differs from its entry: exit 11; the stub did not run.<br>• W4, an entry count other than `--count`: exit 12; the stub did not run.<br>• W5, a duplicate path, an absolute path, a `..` segment, an escaped (`\`-prefixed) entry or a path outside `--path-re`: exit 12 for each sub-case; the stub did not run.<br>• W6, the harness fails with a clean postcheck (`STUB_RC=1`): exit 1; the stub ran; the postcheck passed and its result is reported; the overall run is failed.<br>• W7, postcheck drift (the stub modifies an input, `STUB_RC=0`): exit 13; the stub ran.<br>• W8, the snapshot changes during the run: exit 13.<br>• W9, the harness fails **and** an input drifts: exit 1 (the harness status is preserved), with drift also reported.<br>No case can turn a failure into exit 0. |
| B-SKILL-UNCHANGED (r1 correction 3; route selected by Q9) | `git diff --quiet <base> -- skills/` is still required. Each untracked `skills/**` path, enumerated NUL-separated (`git ls-files -z -o --exclude-standard`) because three screenshot names contain spaces, must be named exactly by an `--expect-hash` and match it. Any other untracked or changed `skills/` path FAILS. There is no directory-wide exception. B-SCOPE's changed-path enumeration and `--exception` matching likewise switch to NUL-separated exact-path comparison. |
| provenance | Driven by the register: B-RENDER-PLAN and B-RENDER-REPEAT expect exactly 3 + 5 conflicts and proposals |
| **B-BROWSER-ROOT** (new) | Runs README `platform:browser` verbatim, then the same command with an execution-owned report mount and the two `PLAYWRIGHT_*` variables. From the JSON: every required scenario ID passes in each required project (§11.2); 0 skipped, flaky or unexpected; the Chromium version is recorded. |
| **B-WEB-HTTP** (new) | Host probes: `/api/health` 200 and `/api/ready` 200, both with `Cache-Control: no-store` (new `http_get_headers` helper in `lib.sh`) |
| B-OUTAGE | Adds no-store on web `/api/ready` 503 during the core outage, and on `/api/health` |
| B-RENAME | Adds **B-BROWSER-RENAME**: the `@identity` scenarios run with the renamed values after the `.env`-only rename and `up --no-build`, with image IDs compared as today |
| B-TOGGLE-OFF-ADDITIONS | `APPLICABLE` adds every G002 web addition and customized web file, `browser.yml`, `web-browser.Dockerfile` and its `.dockerignore`, and `WEB-SHELL.md` |
| B-TOGGLE-OFF | Adds **B-BROWSER-OFF**: the full browser suite against the disabled variant |
| B-BUILD-CONTEXT | Adds the `web-browser` fixture stage, with canary `web/.env` and `.env.local` absent from `/app` and the standalone output |
| B-SECRET-CANARY | Also scans the browser report directories |
| IMAGES | Includes the Playwright image |

**Unchanged:** B-SCOPE, B-PROTECTED, B-LITERALS (now also covering `web/tests`), B-MCP-SOURCE, the migrations suite, ports and boundaries with their controls, B-PERSIST, B-DBRENAME, B-SOURCE-HASH and B-DOCKER-ONLY. The historical G001 allow-list and reports are not edited.

## 16. Execution sequence, gates and provenance handling

### 16.1 S0a: read-only, fail-closed starting-state gate (r5; read-only defects fixed in r6)

**Bootstrap rule.** S0a needs nothing that G002 creates. It uses only bash, coreutils, git and docker, all present at the baseline. It runs from a file **outside the repository** (`<scratch>/g002-s0-gate.sh`), extracted byte-exactly from the block below. The executor verifies the extracted file's SHA-256 equals **`853b20dfe26b974479d1a794cd4ccd024f620edcfe2571e9d8ad5b252916b428`** before running it; a different hash means stop.

**Read-only contract (r6).** S0a never writes inside the repository, including `.git`:

- **Evidence path.** The destination is resolved (`realpath -m`: symlinks, including dangling ones, and `..` are followed for every component, existing or not) and **rejected with exit 26 before anything is created** if it is the repository, a descendant, or resolves into it. Only then is the resolved directory created. It is re-verified (`ev_ok`: it still resolves to the same path, is a directory, and is outside the repository) before every evidence write.
- **Git.** Every Git call is `git --no-optional-locks -c core.fsmonitor=false -C "$REPO"`. `GIT_OPTIONAL_LOCKS=0` is exported and inherited Git redirection variables (`GIT_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE` and others) are unset, so status never refreshes `.git/index` and no fsmonitor daemon starts.
- **Proof.** §16.4 shows unchanged whole-tree inventory (including `.git`) and unchanged index bytes and stat data across every invocation, including a stale-index fixture.

```bash
bash <scratch>/g002-s0-gate.sh <repo> <approved snapshot> <approved snapshot SHA-256 from Codex's approval> <evidence dir outside repo>
```

| Step | Verifies | Exit on failure |
|---|---|---|
| 0 | Usage. The repository resolves to a directory. The evidence destination resolves (`realpath -m`) **outside** the repository: not the repository itself, not a descendant, and not reached by traversal or an external or dangling symlink. This check happens **before any write**. Only afterwards is the resolved directory created, and it is re-verified before every evidence write. | 26 |
| 1 | `HEAD` = `cc157ed1b6132aa9795940d399aa3e7ff8e24a39` (read with `--no-optional-locks`) | 20 |
| 2 | The snapshot is read **once**, and the SHA-256 of those bytes equals the supervisor-pinned digest | 21 |
| 3 | Only those bytes are parsed. Each line is `<64 hex>  <path>` with no escaped entries. Each path is relative, with no `.` or `..` segment, unique, and matches the approved syntax. There are exactly 20 entries, and the Plan is one of them. **The exact set is the pinned snapshot's content; the regex is only an extra syntax constraint.** | 22 |
| 4 | Every entry is a regular, non-symlink file whose SHA-256 equals its entry | 23 |
| 5 | `git --no-optional-locks -c core.fsmonitor=false status --porcelain=v1 -z --untracked-files=all --ignored`: every record is `??` or ` M`, and the set of differing paths **equals** the 20 snapshot paths. Anything staged, deleted, renamed, ignored (for example a root `.env`) or extra fails. | 24 |
| 6 | No root `.env`; no containers, volumes or networks for the Compose project; default ports 3100, 8180, 8100, 5532, 6479, 9100 and 9101 free. The Docker inventory is recorded to evidence. | 25 |

**Exit 0 is the only way to proceed.** Every other status stops the execution before S0b: no file is created, and no CLI, install, build or test runs.

Exact S0a script (SHA-256 `853b20dfe26b974479d1a794cd4ccd024f620edcfe2571e9d8ad5b252916b428`; extract the lines between the fences, including the final newline):

```bash
#!/usr/bin/env bash
# G002 S0a: read-only, fail-closed starting-state gate (Plan §16.1, r6).
# Needs only bash, GNU coreutils, git and docker. Writes nothing inside REPO,
# including .git: Git runs without optional locks, so the index is never refreshed.
# The evidence destination is resolved and proven outside REPO before anything is created.
# Usage: bash g002-s0-gate.sh REPO SNAPSHOT SNAPSHOT_SHA256 EVIDENCE_DIR
# Exit: 0 pass; 20 HEAD; 21 snapshot digest; 22 snapshot format/paths/count;
#       23 input hash; 24 starting state; 25 local resources; 26 usage or evidence path.
set -euo pipefail
shopt -s inherit_errexit

fail() { printf 'S0 FAIL %s: %s\n' "$1" "$2" >&2; exit "$1"; }
[[ $# -eq 4 ]] || fail 26 "usage: REPO SNAPSHOT SNAPSHOT_SHA256 EVIDENCE_DIR"

# Git inspection must not write repository metadata (no index refresh, no fsmonitor),
# and must not be redirected by the caller's environment.
export GIT_OPTIONAL_LOCKS=0
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES \
  GIT_COMMON_DIR GIT_NAMESPACE GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT
g() { git --no-optional-locks -c core.fsmonitor=false -C "$REPO" "$@"; }

# 0. Resolve every location before any write.
REPO="$(realpath -e -- "$1" 2>/dev/null)" || fail 26 "repository path does not resolve"
[[ -d "$REPO" ]] || fail 26 "repository path is not a directory"
SNAP="$2" SNAP_SHA="$3"
EV="$(realpath -m -- "$4" 2>/dev/null)" || fail 26 "evidence path does not resolve"
inside_repo() { [[ "$1" == "$REPO" || "$1" == "$REPO"/* ]]; }
inside_repo "$EV" && fail 26 "evidence destination resolves inside the repository: $EV"
# Only now may anything be created; the resolved path is re-verified before every write.
mkdir -p -- "$EV"
ev_ok() {
  local now
  now="$(realpath -e -- "$EV" 2>/dev/null)" || fail 26 "evidence directory vanished"
  if [[ "$now" != "$EV" || ! -d "$now" ]] || inside_repo "$now"; then fail 26 "evidence directory changed: $now"; fi
}
ev_ok

BASE=cc157ed1b6132aa9795940d399aa3e7ff8e24a39
COUNT=20
PLAN=management/v0/plans/G002-web-shell-plan.md
PATH_RE='^(management/v0/(goals/G002-web-shell\.md|specs/G002-web-shell-spec\.md|PROGRESS\.md|plans/G002-web-shell-plan\.md)|skills/vestive-design-skill/.+)$'
LINE_RE='^([0-9a-f]{64})  ([^\].*)$'
cd -- "$REPO"

# 1. Baseline commit.
[[ "$(g rev-parse HEAD)" == "$BASE" ]] || fail 20 "HEAD is not $BASE"

# 2. Snapshot: read once; verify the digest of exactly those bytes.
bytes="$(cat -- "$SNAP"; printf x)"; bytes="${bytes%x}"
[[ "$(printf '%s' "$bytes" | sha256sum | cut -d' ' -f1)" == "$SNAP_SHA" ]] || fail 21 "snapshot digest mismatch"

# 3. Parse only those bytes: format, relative unique approved paths, exact count, Plan present.
[[ -n "$bytes" && "$bytes" == *$'\n' ]] || fail 22 "snapshot is empty or lacks a final newline"
declare -A want=()
n=0
while IFS= read -r line; do
  [[ "$line" =~ $LINE_RE ]] || fail 22 "malformed entry: $line"
  sha="${BASH_REMATCH[1]}" path="${BASH_REMATCH[2]}"
  [[ "$path" != /* && "/$path/" != */./* && "/$path/" != */../* ]] || fail 22 "non-relative path: $path"
  [[ "$path" =~ $PATH_RE ]] || fail 22 "path outside the approved set: $path"
  [[ -z "${want[$path]+x}" ]] || fail 22 "duplicate path: $path"
  want[$path]="$sha"; n=$((n + 1))
done <<<"${bytes%$'\n'}"
(( n == COUNT )) || fail 22 "expected $COUNT entries, found $n"
[[ -n "${want[$PLAN]+x}" ]] || fail 22 "the Plan is not an entry"

# 4. Every input matches its approved hash.
for path in "${!want[@]}"; do
  [[ -f "$path" && ! -L "$path" ]] || fail 23 "missing or non-regular input: $path"
  [[ "$(sha256sum -- "$path" | cut -d' ' -f1)" == "${want[$path]}" ]] || fail 23 "hash mismatch: $path"
done

# 5. Starting state (NUL-safe, no optional locks): the paths differing from HEAD are exactly
#    the snapshot paths, each untracked or unstaged-modified; nothing staged, deleted, renamed or ignored.
ev_ok
g status --porcelain=v1 -z --untracked-files=all --ignored >"$EV/s0-status.z" || fail 24 "git status failed"
declare -A seen=()
while IFS= read -r -d '' rec; do
  xy="${rec:0:2}" path="${rec:3}"
  [[ "$xy" == "??" || "$xy" == " M" ]] || fail 24 "unexpected status '$xy': $path"
  [[ -n "${want[$path]+x}" ]] || fail 24 "differing path not in the snapshot: $path"
  seen[$path]=1
done <"$EV/s0-status.z"
for path in "${!want[@]}"; do
  [[ -n "${seen[$path]+x}" ]] || fail 24 "snapshot path does not differ from HEAD: $path"
done

# 6. Local resources the execution must never touch: no root .env, no project resources, free ports.
project="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env.example)"
[[ -n "$project" ]] || fail 25 "cannot read COMPOSE_PROJECT_NAME from .env.example"
[[ ! -e .env && ! -L .env ]] || fail 25 "a root .env exists"
for kind in "ps -a -q" "volume ls -q" "network ls -q"; do
  # shellcheck disable=SC2086
  out="$(docker $kind --filter "label=com.docker.compose.project=$project")" || fail 25 "docker query failed"
  [[ -z "$out" ]] || fail 25 "existing $project resources ($kind)"
done
for port in 3100 8180 8100 5532 6479 9100 9101; do
  if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then fail 25 "port $port is busy"; fi
done
ev_ok
{ docker ps -a --format 'container {{.Names}}'; docker volume ls --format 'volume {{.Name}}'
  docker network ls --format 'network {{.Name}}'; docker image ls --format 'image {{.Repository}}:{{.Tag}}'; } |
  LC_ALL=C sort >"$EV/s0-docker-inventory.txt"

ev_ok
printf 'S0 PASS: HEAD %s; snapshot %s; %d inputs verified; differing paths = snapshot paths; no .env, no %s resources, default ports free\n' \
  "$BASE" "$SNAP_SHA" "$n" "$project" | tee "$EV/s0-gate.txt"
```

### 16.1b S0b: bootstrap writes (only after S0a passes and the Plan is approved)

1. **Create** `infra/validation/run-approved.sh` (§17.1) and `infra/validation/run-approved-controls.sh` (§15, B-WRAPPER-CONTROLS). These are G002's first writes, and its only bootstrap writes.
2. **Bootstrap-write gate.** The NUL-safe status set must equal exactly the 20 S0a paths plus these two new untracked paths. Nothing else may be written. Shellcheck both scripts in Docker (the pinned image, read-only mount). Any difference or finding means stop.

### 16.1c S0c: wrapper validation before S1

1. Run `bash infra/validation/run-approved-controls.sh` (W1–W9, stub harness, scratch tree outside the repository). It must exit 0, with every case's expected status and stub-invocation outcome met.
2. Run `bash infra/validation/run-approved.sh --check-only` with the same approved snapshot, pinned digest, `--count 20`, `--path-re` and `--plan` as in §17.2. It must exit 0.
3. **Status check.** The status set must still equal the S0b set (20 + 2), because S0c writes only to scratch.
4. Any failure stops before S1. A wrapper defect found here is diagnosed, fixed and retested as a counted meaningful cycle (BUDGET_POLICY §3).

### 16.2 Transient versus final writes

| Step | Path | Transient (allowed mid-execution) | Final state |
|---|---|---|---|
| S0b | `infra/validation/run-approved.sh`, `infra/validation/run-approved-controls.sh` | created (the only bootstrap writes; §16.1b gate) | kept; S5 wires the controls into `run.sh` without changing either file's interface |
| S1 | `web/components.json`, `web/lib/utils.ts`, `web/components/ui/button.tsx`, `web/components/ui/dialog.tsx` | created by the CLI | kept (button and dialog with E1–E7) |
| S1 | `web/package.json`, `web/package-lock.json` | caret dependencies, then `eject` removes `shadcn` | exact pins (S2), registered customizations |
| S1 | `web/app/globals.css` | CLI token block and `.dark` block, then the ejected CSS | §7 content (S2), registered customization |
| S1 | `web/app/layout.tsx` | **Geist `next/font/google` import and `cn(...)` className (explicitly permitted)** | S2 removes both; the final file is §8.2 only (verified: no `next/font`, no `next/script`, classic inline init present; S-HEAD-ORDER at runtime), registered customization |
| S1 | `/app/node_modules` | only in execution-owned volume `g002-web-nm-<exec-id>` | removed at S9 (label-verified) |

**S1 gate.** The "before" tree hash (excluding `.git`) is taken **after S0c**, so it already includes the two S0b files. The difference after S1 must equal exactly the S1 rows above; any other path means stop and return to Codex. After S2, `layout.tsx` must contain no `next/font` and no `geist`.

### 16.3 Sequence (r3: every gate builds only after the modules it imports exist)

| Step | Work | Import closure created in this step | Gate |
|---|---|---|---|
| S0a | Read-only gate (§16.1) | — | Exit 0, or stop (no write of any kind) |
| S0b | Create `run-approved.sh` and `run-approved-controls.sh` (§16.1b) | — | Status = 20 S0a paths + 2; shellcheck clean |
| S0c | Run W1–W9 and `--check-only` (§16.1c) | — | Both exit 0; status unchanged from S0b |
| S1 | CLI init, add and eject (§4.1) | `lib/utils.ts`, `components/ui/{button,dialog}.tsx` (from the CLI) | §16.2 S1 write gate (no build) |
| S2 | Reconcile: pins (`npm install --save-exact` in the container with the volume), add the two dev dependencies, `globals.css` §7, `layout.tsx` §8.2, E1–E7. **Create the modules imported by these edits:** `lib/appearance.ts` (imported by `layout.tsx`), `components/ui/portal-container.tsx` (imported by `dialog.tsx`, E7) and `lib/tokens.ts`. | `layout.tsx` → `@/lib/config`, `@/lib/appearance`; `dialog.tsx` → `./button`, `./portal-container`, `@base-ui/react`, `lucide-react`; `button.tsx` → `cn`, `class-variance-authority`, `@base-ui/react` | `web-check` and the production build pass in Docker |
| S3 | `components/layout/{token-scope,appearance-control,shell-help-dialog,app-shell}.tsx`, then `app/page.tsx` | `token-scope` → `ui/portal-container`, `lib/tokens`; `appearance-control` → `lib/appearance`; `shell-help-dialog` → `ui/dialog`, `ui/button`; `app-shell` → the three; `page.tsx` → `app-shell`, `lib/config` | `web-check` and the build pass |
| S4 | Playwright config, specs, support, fixtures; `web-browser.Dockerfile` and its `.dockerignore`; `browser.yml` | Fixtures → `components/layout/*`, `lib/tokens`; specs → `support/*`, fixtures data | Browser suite green against the root started with README blocks |
| S5 | Harness and register (§15): `run.sh` gains B-WRAPPER-CONTROLS, which invokes the **existing** `run-approved-controls.sh` (created at S0b, validated at S0c); `G002-scope-allowlist.txt` (exact paths, §18) | — | shellcheck; static (including B-DEPS-CONTROLS C1–C12 and B-WRAPPER-CONTROLS W1–W9) and provenance green |
| S6 | Docs: README, WEB-SHELL.md (including the init-script hash), FOLLOW-UP labels | — | — |
| S7 | Incremental static → provenance → root → migrations → variants | — | All green |
| S8 | Two read-only reviewers (§20); triage and fixes | — | Findings resolved or recorded |
| S9 | Final run **only through `run-approved.sh`** (§17), with **no repository edits during the run**; full diff; execution-owned cleanup with an inventory diff; Result; stop for Codex | — | The wrapper exits 0, which requires the precheck to pass, the harness to exit 0 with 0 FAIL and 0 UNMET, and the postcheck to pass. **Any other exit stops the handoff as a failed run.** |

### 16.4 Planning-time demonstration of S0a (r6 script; read-only; actual baseline, wrapper absent)

- **Setup:** run on 2026-10-05 at r6 writing, with the §16.1 script (SHA-256 `853b20df…b428`) and Codex's r5 continuity snapshot (digest `23d58944…606a`). That snapshot is not an approved execution snapshot; it exercises the gate only while the Plan entry still matched r5.
- **Measurements around every invocation:**
  - **tree:** a whole-tree inventory **including `.git`** (`find -printf '%P %y %s %T@ %i'`: every path, type, size, mtime to the nanosecond and inode);
  - **index:** the `.git/index` SHA-256 and its stat data (size, mtime, ctime, inode);
  - **status:** the NUL-safe status digest;
  - for containment cases, an explicit check that each probed in-repository path does not exist afterwards.
- **Isolation:**
  - The real repository hosted P, N1–N6, A1–A4 and B1–B4.
  - Disposable `git clone`s carrying the same 20 inputs hosted N7–N11 and C0–C4. They were removed afterwards.
  - Final real-repository state: status digest `7e1d605e…` (as before r6 work), all 20 snapshot entries OK, the wrapper absent, and none of the probed paths present.
  - No CLI, install, build or wrapper ran.

| Case | Input | Expected | Actual | tree / index / status |
|---|---|---|---|---|
| P | Real repository, correct snapshot and digest, external evidence directory | 0 | 0: `S0 PASS … 20 inputs verified …`. Evidence (`s0-gate.txt`, `s0-status.z`, `s0-docker-inventory.txt`) is written only to the external directory. | same / same / same |
| N1 | Wrong pinned digest | 21 | 21 | same / same / same |
| N2 | Wrong count (19 entries; own digest pinned) | 22 | 22: expected 20, found 19 | same / same / same |
| N3 | Duplicate path | 22 | 22 | same / same / same |
| N4 | Absolute path | 22 | 22 | same / same / same |
| N5 | Escaped (`\`-prefixed) entry | 22 | 22 | same / same / same |
| N6 | Altered input hash (digest re-pinned) | 23 | 23: hash mismatch (the Plan) | same / same / same |
| **A1** | Evidence `<repo>/s0-new-evidence` (nonexistent, inside) | 26 | 26; **not created** | same / same / same |
| **A2** | Evidence = the repository root | 26 | 26 | same / same / same |
| **A3** | Evidence `<repo>/infra/../s0-traversal` | 26 | 26; not created | same / same / same |
| **A4** | Evidence `<repo>/.git/s0-evidence` | 26 | 26; not created | same / same / same |
| **B1** | External symlink pointing at the repository, then a child (`<ext>/link-to-repo/new-evidence`) | 26 | 26; `<repo>/new-evidence` not created | same / same / same |
| **B2** | External **dangling** symlink pointing at a nonexistent in-repository path | 26 | 26; target not created | same / same / same |
| **B3** | External symlink pointing at existing `<repo>/infra`, then a child | 26 | 26; `<repo>/infra/s0-x` not created | same / same / same |
| **B4** | External symlink pointing at `<repo>/infra`, then `..` | 26 | 26; `<repo>/s0-up` not created | same / same / same |
| N7 | Clone with the same 20 inputs | 0 | 0 | same / same / same |
| N8 | Clone plus an extra untracked file | 24 | 24: not in the snapshot | same / same / same |
| N9 | Clone with PROGRESS.md staged | 24 | 24: status `M ` | same / same / same |
| N10 | Clone plus a root `.env` | 24 | 24: status `!!` (rejected by step 5 before step 6, unchanged from r5) | same / same / same |
| N11 | Clone with HEAD moved by an empty commit | 20 | 20 | same / same / same |
| **C0** | Stale-index fixture: two tracked files' mtimes changed, content unchanged; a copy runs an **ordinary** `git status` | index rewritten | rewritten. The detector is non-vacuous: without the fix, status writes the index. | — |
| **C1** | Stale-index fixture, positive path | 0 | 0 | same / **same** / same |
| **C2** | Stale-index fixture, failing path that runs `git status` (extra untracked file) | 24 | 24 | same / **same** / same |
| **C3** | Stale-index fixture, failing path before Git (wrong digest) | 21 | 21 | same / **same** / same |
| **C4** | After C1–C3, an ordinary `git status` on the fixture | index rewritten | rewritten. The index was still stale, so the gate never refreshed it. | — |

**Control D (repository immutability):** in every row above, including all rejected invocations (N1–N6, A1–A4, B1–B4, N8–N11, C2, C3), the whole-tree inventory including `.git` was byte-for-byte identical in path, type, size, mtime and inode before and after the gate. The gate added or modified no tracked, untracked, ignored or Git-metadata path.

**Ordering consequence (unchanged):** every failing case exits before S0b, so no bootstrap or application write can follow a malformed snapshot, a mismatched input, a bad starting state or an in-repository evidence destination.

## 17. Final validation: fail-closed wrapper (r4 correction 2)

### 17.1 `infra/validation/run-approved.sh` (new, project-owned, committed)

A Bash script (`#!/usr/bin/env bash`, `set -euo pipefail`, `shopt -s inherit_errexit`). It never relies on `errexit` for the harness status, which is captured explicitly.

```text
run-approved.sh --snapshot FILE --snapshot-sha256 HEX64 --count N --path-re ERE --plan PATH
                [--check-only] -- <run.sh arguments…>
```

**1. Read once, verify those bytes.**
- The snapshot is read into memory exactly once (`bytes="$(cat -- "$FILE"; printf x)"; bytes="${bytes%x}"`).
- The SHA-256 of **those bytes** (`printf '%s' "$bytes" | sha256sum`) must equal `--snapshot-sha256`, which is pinned from Codex's approval. Otherwise exit **10**.
- Every later step parses only `$bytes`; the file is never re-read for arguments, so it cannot be substituted or regenerated.

**2. Parse strictly.** Every line must match `^[0-9a-f]{64}  [^\\].*$`: two spaces, and no escaped (`\`-prefixed) entries. Each path must:
- be relative, with no leading `/`;
- contain no `.` or `..` segment;
- be unique;
- match `--path-re`.

The line count must equal `--count`, and `--plan` must be one of the entries. Otherwise exit **12**.

For G002:

```text
--count 20
--path-re '^(management/v0/(goals/G002-web-shell\.md|specs/G002-web-shell-spec\.md|PROGRESS\.md|plans/G002-web-shell-plan\.md)|skills/vestive-design-skill/.+)$'
```

**3. Precheck the entries.** For every parsed entry, `sha256sum -- "$path"` must equal the entry's hash. Paths are held in quoted arrays and handle the spaced screenshot names. Any mismatch or missing file exits **11**. With `--check-only`, the wrapper exits 0 here.

**4. Build the arguments from the parsed entries only.**
- One `--expect-hash "$path=$sha"` per entry.
- One `--exception "$path"` per entry except `--plan`. The Plan is a G002 artifact and is in the G002 allow-list.

**5. Run the harness and capture its status explicitly:**

```bash
set +e
"$REPO/infra/validation/run.sh" "${passthrough[@]}" "${args[@]}"
harness_rc=$?
set -e
```

**6. Postcheck, always.** It recomputes the snapshot-bytes digest from the file and each entry's hash. A mismatch sets `post_rc=13`. The result is printed for evidence in every case.

**7. Exit:**
- `harness_rc` if it is nonzero. **A successful postcheck can never replace a failing harness**, and drift is still reported.
- Otherwise 13 if the postcheck failed.
- Otherwise 0.

There is no `|| true`, `; true` or trailing command that could mask the status.

### 17.2 Final invocation (S9)

```bash
bash infra/validation/run-approved.sh \
  --snapshot <approved snapshot named by Codex> \
  --snapshot-sha256 <its SHA-256 from Codex's approval> \
  --count 20 \
  --path-re '^(management/v0/(goals/G002-web-shell\.md|specs/G002-web-shell-spec\.md|PROGRESS\.md|plans/G002-web-shell-plan\.md)|skills/vestive-design-skill/.+)$' \
  --plan management/v0/plans/G002-web-shell-plan.md \
  -- --base cc157ed1b6132aa9795940d399aa3e7ff8e24a39 \
     --allow-file management/v0/results/G002-scope-allowlist.txt \
     --report-dir <outside-repo> all
echo "overall_rc=$?"     # recorded; the handoff proceeds only when it is 0
```

- **Exit 0** means a passing run: the precheck passed, the harness passed with 0 FAIL and 0 UNMET, and the postcheck passed.
- **Any other status** (10, 11, 12, 13, or the harness's own nonzero status) is a failed overall run. The Result records the status and evidence, and the handoff stops as a failure; no success is claimed.
- **Committed route:** if Codex or the owner later commits the inputs instead, `--base` becomes that commit and only the Plan needs binding. The same wrapper is used with a one-entry approved snapshot (`--count 1`).

### 17.3 Controls

B-WRAPPER-CONTROLS W1–W9 (§15) prove the control flow: entry mismatch and a changed snapshot never invoke the harness; harness failure survives a clean postcheck; and drift fails a passing harness.

## 18. Files expected to change (exact; `G002-scope-allowlist.txt` mirrors this section)

- **Customized rendered files (5, registered):** `web/package.json`, `web/package-lock.json`, `web/app/globals.css`, `web/app/page.tsx`, `web/app/layout.tsx`.
- **Already-registered rendered file, changed content:** `docs/foundation/FOLLOW-UP.md`. The STAGE-WEB-SHELL labels:
  - shadcn/Base UI: done;
  - accessible layout and keyboard: done;
  - token overrides: “mechanism done; Organization branding data in G016/G017”;
  - thin BFF: verified;
  - appearance modes and design language: noted.
- **New web files (25):**
  - `web/components.json`;
  - `web/lib/{utils,tokens,appearance}.ts`;
  - `web/components/ui/{button,dialog,portal-container}.tsx`;
  - `web/components/layout/{app-shell,token-scope,shell-help-dialog,appearance-control}.tsx`;
  - `web/playwright.config.ts`;
  - `web/tests/browser/{shell,overlay,appearance,accessibility,tokens,boundary,design}.spec.ts`;
  - `web/tests/browser/support/{color,png,page-errors,modes}.ts`;
  - `web/tests/fixtures/token-scopes/{page.tsx,fixtures.ts}`.
- **New infrastructure (3):** `infra/docker/web-browser.Dockerfile`, `infra/docker/web-browser.Dockerfile.dockerignore`, `infra/compose/browser.yml`.
- **New project validation tools (2):** `infra/validation/run-approved.sh` and `infra/validation/run-approved-controls.sh`. Both are created at S0b. Neither is in `APPLICABLE`, because they are root-only validation tooling, like `run.sh`.
- **Not in the repository:** the S0a gate script (§16.1) is execution tooling kept outside the repository, verified by its recorded SHA-256.
- **Modified project-owned files:** `infra/validation/run.sh`, `infra/validation/lib.sh`, `infra/validation/render-customizations.tsv` (+5 rows), `README.md` (`platform:browser` block and WEB-SHELL link).
- **New docs:** `docs/foundation/WEB-SHELL.md`, covering:
  - the four concerns;
  - the colour and non-colour token contract, with both palettes and the inherited-value changes;
  - scope composition with modes, and the portal rule;
  - appearance architecture, storage and fallback, and the init-script hash;
  - the design-input inventory, the screenshot-to-shell mapping and the adaptations T1–T13;
  - component and CLI provenance, with the reconciliation and E1–E7;
  - adding components in Docker;
  - the test matrix, viewports, fixtures and tool versions;
  - the identity invariants;
  - what is completed, inherited and deferred.
- **G002 management artifacts:** this Plan; `management/v0/results/G002-web-shell-result.md`; `management/v0/results/G002-scope-allowlist.txt`.
- **Not touched:**
  - `web/lib/{config,core-client}.ts`, `web/app/api/**`, `web/{next.config.ts,tsconfig.json,eslint.config.mjs,postcss.config.mjs,next-env.d.ts}`;
  - the web Dockerfile and its `.dockerignore`, `infra/compose/{local,migrate}.yml`;
  - `core/**`, `intelligence/**`, `contracts/**`;
  - `docs/foundation/{BOOTSTRAP,MIGRATIONS,GETTING-STARTED,VENDORS}.md`, `.env.example`, `.gitignore`, the manifest;
  - **`skills/**` (both the project-archetype Skill and the design skill)**, governance and architecture docs, and closed-goal artifacts.

## 19. Budget and feasibility (B7 authorized; checkpoint after cycle 4)

Codex authorized **B7** in the r2 review: 7 shared meaningful cycles, with the checkpoint after cycle 4. The combined scope is kept, and the split proposed in r2 was not selected.

The uncertain surface comprises:
- a mode-aware CSS architecture (custom variant, `@variant` inside `:root`, angled pseudo-element fills);
- the parser-time init, with the head-order, paint-timed and delayed-JS first-paint evidence;
- dev-mode hydration evidence;
- a rendered-pixel contrast harness across 2 appearances × 3 scopes, including hover and focus;
- angled-label geometry;
- the strict B-DEPS with independent reason codes and controls C1–C12;
- the fail-closed wrapper with controls W1–W9;
- design-language calibration.

**Forecast of meaningful cycles:**

| Source | Likelihood |
|---|---|
| First build after the CSS and token architecture (S2/S3) | likely 1 |
| Appearance first paint (head order, paint probe, delayed-JS routing) and dev hydration | likely 1 |
| Browser-suite calibration (contrast and focus sampling, angles, axe in both appearances) | 1–2 |
| Harness integration (B-DEPS-CONTROLS, B-WRAPPER-CONTROLS, scenario-matrix parsing, variants) | likely 1 |
| Read-only review findings batch | likely 1 |

**Expected 5–6 of 7.** The r3 and r4 corrections add test design and a small validation wrapper, not product surface, so the forecast is unchanged from r2.

- **Checkpoint after cycle 4** (BUDGET_POLICY §7): criteria completed and remaining, suite status, repeated failures, uncertainty, convergence, and the forecast against 7. Stop and return to Codex on thrashing (§8) or a forecast above 7.
- **Counting:** BUDGET_POLICY §3–§5; doubtful cases count. Planning revisions count no cycles. The sequence in §16.3 creates every module before the gate that imports it, so no cycle arises from step order.

## 20. Delegation

- **No implementation subagents.** The work is tightly coupled across `globals.css`, components, tests and harness.
- **Two read-only reviewers after S7** (approved by Codex in the r2 review), run in parallel. Corrective implementation is mine and counts against B7. Neither edits files; each reports findings with severity and file:line, and cycles used (expected 0). I validate and integrate every finding, and their opinions do not replace Codex review.
  - **SA-R (accessibility, boundary and provenance).** The WCAG mapping; whether any browser assertion could pass vacuously; the rendered-pixel method; token isolation, portals and mode composition; appearance storage, first-paint (S-HEAD-ORDER, S-FIRSTPAINT, S-FIRSTPAINT-DELAYED) and hydration evidence (S-HTML-ATTRS); the BFF boundary; harness weakening; B-DEPS-CONTROLS; and scope against the allow-list.
  - **SA-V (design-language fidelity).** The built-shell screenshots in Light and Dark, compared with the skill's three screenshots and the `README.md` visual rules. It checks the §9.2 mapping and the T1–T13 adaptations, and the absence of marketing copy, identity and assets. It confirms that no rule forced Dark and that semantic tokens rather than literals carry the language.

## 21. Risks and escalation

| # | Risk | Handling |
|---|---|---|
| R1 | Very recent package releases (r1) | Exact pins, lockfile, B-DEPS strict |
| R2 | The CLI writes differently on the real tree | S1 gate (§16.2); any extra path means stop |
| R3 | Tailwind 4.3 `@custom-variant` with nested `@media`, or `@variant` inside `:root`/selectors, behaves unexpectedly | Verified at S2 by the build plus S-APPEARANCE-*. Fallback: explicit duplicated dark blocks guarded by B-WEB-TOKENS set equality (same semantics). |
| R4 | React 19 or Next moves, hoists or warns about the inline head script, or adds `async`/`defer` to it | S-HEAD-ORDER (served HTML) and S-HYDRATION-DEV detect it. There is no weaker fallback: stop and return to Codex (§8.2). |
| R5 | Pixel sampling is flaky (anti-aliasing, transitions) | Sample padding areas, at device scale factor 1, after transitions or under reduced motion; `retries: 0` keeps failures visible |
| R6 | Palette or fixture values fail rendered contrast in one appearance | Values adjusted within the design mapping; adjustments recorded; no rule weakened |
| R7 | Fallback fonts make the screenshots look less like the skill | T3 recorded; SA-V judges structure, type roles and colour roles; no font download |
| R8 | The dev server in the runner (write permissions, compile time) | `chown` in the image; `webServer` timeout; the server only serves `/` |
| R9 | An axe false positive on `color-mix` or translucent surfaces | Selector-scoped exclusion only with rendered-pixel evidence |
| R10 | Harness edits weaken G001 checks | §15 lists each change; SA-R reviews; scope, protected and provenance checks are untouched |
| R11 | Budget exhaustion | B7 authorized; checkpoint after cycle 4; thrashing stop |
| R12 | Delayed-JS routing holds the wrong resources (for example CSS emitted under `/_next/static/chunks/`), so the scenario tests a stylesheet delay rather than a framework delay | The route matches only `*.js` and logs every held URL. The test asserts at least one held chunk, that the stylesheet loaded (the computed palette differs from the browser default), and the FCP-before-release ordering. If the URL layout differs, the pattern is corrected within the scenario; the assertions are never weakened. |
| R13 | `addInitScript` probe timing differs from parser timing | The probe installs at document creation (before the inline init), and its records are correlated with the browser's own paint entries (§11.4). Runs without a valid visible-content sample fail. |

**Stop and escalate:**
- architecture contradictions, scope expansion, security-model change, a public API or MCP change, new infrastructure or vendors, schema or migrations, or conflict with an accepted ADR;
- if Base UI or shadcn requires replacing Next.js, Tailwind or npm, or a `web/src` relocation;
- if token or appearance work needs tenant schema, resolution, cookies or server state;
- if either skill needs an edit;
- if a required check cannot run without weakening it;
- if the CLI writes unexpected paths, or a non-registered rendered file changes;
- if developer `.env` or Docker resources would be touched;
- on thrashing or budget exhaustion.

No product-specific work.

## 22. Questions for Codex

**None open.** Q0–Q10 are resolved (§0). Implementation waits for an explicit Codex ACCEPT of this Plan, and for the approved-inputs snapshot named in that approval (§16.1).

## 23. Acceptance-criteria mapping

| AC | Implementation | Evidence |
|---|---|---|
| AC1 shadcn/Base UI, attributable dependencies | §4 CLI and reconciliation, E1–E7, §5 pins | S1 output and diff; B-DEPS (independent reason-coded rules) and B-DEPS-CONTROLS C1–C12 (consistent manifest mutations, lock-only mutations, exact expected code sets, positive control); B-WEB-COMPONENTS; generated-versus-committed diff; `npm ci` in builds; production build; B-NAMING; B-PROTECTED |
| AC2 usable generic shell | §6 shell, generic copy | S-IDENTITY, S-LANDMARKS, S-NAV, S-RESPONSIVE (all viewports, both appearances), S-ERRORS, negative name checks; B-BROWSER-ROOT |
| AC3 keyboard and accessibility | Skip link, outlines, Base UI modal, radio group, reduced motion | S-SKIP, S-TAB, S-OVERLAY, S-MOTION, S-A11Y, S-CONTRAST (rendered, rest, hover and focus), S-ANGLE, in both appearances |
| AC4 overrides without rewrite or leakage | §7.1–§7.6 contract, `TokenScope` with per-mode sources, scope portals | S-TOKENS, S-PORTAL (including live mode change), S-RUNTIME, S-SCREENSHOTS (fixtures), B-WEB-TOKENS, WEB-SHELL contract |
| AC5 identity, BFF, health | BFF untouched; identity from config; no `NEXT_PUBLIC_` or cookies | B-RENAME plus B-BROWSER-RENAME; B-WEB-BOUNDARY; S-BOUNDARY; S-HEALTH; B-WEB-HTTP; B-OUTAGE (no-store); S-FAILCLOSED; S-NOFIXTURE; B-SECRET-CANARY |
| AC6 repeatable Docker checks | README blocks; pinned image; read-only S0a gate (§16.1); fail-closed wrapper (§17), created and validated before S1 | S0a exit 0; S0c (W1–W9 and `--check-only`); B-WRAPPER-CONTROLS W1–W9; the final run through `run-approved.sh` with `overall_rc=0`; B-DOCS-EXEC-ROOT, B-CHECKS-ROOT, B-BROWSER-ROOT (no skips; ID × project matrix), B-DOCKER-ONLY, inventory-diff cleanup, IMAGES |
| AC7 provenance and regressions | Register +5; §15 reconciliation; disabled-variant additions; exact-hash approved inputs with a pinned snapshot digest (§16.1, §17) | S0a steps 2–5; the S0b bootstrap-write gate; B-DEPS-CONTROLS C1–C12 (exact code sets); the wrapper precheck and postcheck; B-RENDER-ID, B-RENDER-PLAN, B-RENDER-REPEAT (exactly 8), B-PROTECTED, B-SKILL-UNCHANGED (both skills), full root, migrations and variants regression; B-TOGGLE-OFF-ADDITIONS, B-TOGGLE-OFF, B-BROWSER-OFF; B-SCOPE with the exact allow-list |
| AC8 scope, invariants, handoff, budget | §18 files; UUID invariants (WEB-SHELL); FOLLOW-UP; the handoff gated on `overall_rc=0` | Full diff against the S0 base; no schema (B-MIGRATE-SET unchanged); Result maps AC1–AC10, cycles and the checkpoint |
| AC9 Light, Dark and System; System default | §8 storage, parser-blocking inline init (§8.2), CSS System tracking, control; §7.5 resolution | S-APPEARANCE-DEFAULT, -LIVE, -OVERRIDE, -RETURN, -PERSIST, -INVALID and -NOSTORAGE; S-HEAD-ORDER; S-FIRSTPAINT (paint-timed, visible-content-synchronized); S-FIRSTPAINT-DELAYED (framework JS held; stored appearance opposite the OS); S-HTML-ATTRS; S-HYDRATION-DEV; S-ERRORS; the R2/R3 scenarios in shell-light and shell-dark, and tokens-light and tokens-dark; palette tables in WEB-SHELL |
| AC10 Vestive visual language within constraints | §9 inventory (including the 3 screenshots), §9.2 mapping, T1–T13, §7.2–§7.3 tokens, E2–E4, `angled` | S-DESIGN (computed design rules), S-ANGLE, S-SCREENSHOTS (shell and dialog, Light and Dark; fixtures), S-CONTRAST, S-RESPONSIVE, B-WEB-DESIGN (no remote fonts, assets or marketing literals), B-SKILL-UNCHANGED; SA-V review; WEB-SHELL mapping and adaptation tables |

## 24. Readiness

**r6: ready for Codex plan review.**
- **r5 corrections:** the evidence destination is resolved and proven outside the repository before any write, then re-verified before every evidence write. All Git inspection runs without optional locks. The new exact S0a SHA-256 is `853b20dfe26b974479d1a794cd4ccd024f620edcfe2571e9d8ad5b252916b428`.
- **Demonstration:** P, N1–N11, A1–A4, B1–B4, C0–C4 and D, all as expected, with the real tree (including `.git`) unchanged (§16.4).
- **Unchanged:** S0b/S0c, S1, S5, the final wrapper status and drift rules, B7 with the checkpoint after cycle 4, and the forecast.
- No question is open.

Nothing has been implemented, installed or initialized. Implementation waits for an explicit Codex ACCEPT naming the approved snapshot and its SHA-256.
