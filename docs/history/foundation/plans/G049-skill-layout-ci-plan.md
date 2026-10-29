> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G049 — Agent skill discovery, foundation naming and CI classification: Plan (r2)

Goal: [G049-skill-layout-ci.md](../goals/G049-skill-layout-ci.md) · Spec: [G049-skill-layout-ci-spec.md](../specs/G049-skill-layout-ci-spec.md) (r2 plus the 2026-10-08 root-skills and subtree-skill clarifications)
Author: Claude Code · Date: 2026-10-09 UTC · Status: **Plan r2 for Codex review. No implementation is authorized.**
Review answered: `G049-plan-r1-codex-review.md` (SHA-256 `28390966fb6dd87c46f277ce175721a51057ce68e6a6cf5a8f018f6faa3ed820`; CORRECTION LOOP). The reviewed r1 is kept byte-exact as G049-skill-layout-ci-plan-r1.md (not included) (SHA-256 `13818145…f212e1`).
Budget: **B3**, at most three shared meaningful cycles, checkpoint after cycle 2. Planning used none.
Planning inputs: G049-planning-inputs.json (not included) (r2). The r1 inventory is kept as G049-planning-inputs-r1.json (not included).

## 0. Baseline and the two bindings

| Item | Value |
|---|---|
| **Planning baseline (provisional)** | `150544058d152fd72f87955477b84ef374b16d46`, tree `ef05f4110a65a7266ef97633bc669d0ace6c3f59`. This is the G004 MR !7 reconciliation merge (`7dc4c5c` + `c1412355`). MR !7 is open and not merged. Staging has since moved to `e761206`, adding only the G050 Goal and Spec plus PROGRESS and ROADMAP records; no CI, harness or skill input changed |
| **Execution baseline (not yet set)** | The actual staging commit that integrates G004. At S0 the 38 planning inputs are re-hashed there. The three skill tree IDs (`d8939ab9…`, `7c7f65e3…`, `68a6ba80…`) must match, otherwise STOP and report. Changes to other inputs are re-read and reported before execution |

**(a) Immutable approved-input snapshot** (existing `run-approved.sh`; regular files, byte-identical before and after every run). It holds only inputs that G049 does **not** modify:
- the G049 Goal, the Spec, this Plan (r2 or its accepted successor) and Codex's acceptance record;
- AGENTS.md, CLAUDE.md, `docs/SECURITY.md`, `docs/DECISIONS.md`, ADR-002 and `management/BUDGET_POLICY.md`;
- the G004 Result, the implementation review and the integration-reconciliation record;
- the documentation-bypass record;
- `infra/validation/required-checks.tsv`, the render manifest, `web/package-lock.json`, `web/app/globals.css`, `infra/docker/web.Dockerfile` and `web/tsconfig.json`. If the §3.3 contingency is used, `web/tsconfig.json` leaves the snapshot only through explicit Codex rebinding.

Codex fixes the exact list, count and path regex as usual.

**Excluded from the snapshot:**
- every path in G049's scope: the three source trees, the destination trees, the symlink, and the edited harness, CI, ESLint, register and documentation files;
- `docs/ARCHITECTURE.md`, whose reviewed edit is bound through B-PROTECTED `--expect-hash`;
- Codex's live PROGRESS and ROADMAP, unless Codex chooses to pin them.

No deleted source path and no symlink is ever a snapshot entry. The wrapper is unchanged.

**(b) Relocation binding**, commit-pinned and checked by the reviewed guards (§3.1):
- **Source inventory.** `git ls-tree -r` of the **bound BASE commit** under the three source directories: 248 entries (path, mode, blob). It is read from the commit, so it cannot drift with the work tree.
- **Post-relocation checks.** Each destination path, mode and content is checked against that inventory, apart from §2's three reviewed after-hashes. The link is checked by `git ls-files -s` mode 120000, `readlink` and `realpath`. No source may remain, and no unbound file may appear under a destination.
- **Planning-only entries.** The planning inventory's four skill-file entries (archetype `SKILL.md`, `README.md`, `validation/run.sh`, `shadcn-kit.sha256`) are planning references, not wrapper inputs.

## 1. Approach

Seven parts, the same as r1:
1. three `git mv` relocations;
2. one committed relative symlink;
3. existing guards reconciled through one exact inline relocation table;
4. web-application scans kept at their pre-G049 scope;
5. CI map, selector and workflow changes;
6. one narrowly scoped ESLint exclusion, registered as a rendered-file customization (§3.3);
7. live documentation.

Reused unchanged: the approved-input wrapper, the harness suites, the table-driven group-F controls, B-CI-CONFIG, the accepted Ruby admission oracle, disposable copies and cleanup. No new framework, dependency, runner or job.

## 2. Layout and exact byte deltas (unchanged from r1)

| Source (baseline tree ID) | Destination | Files and modes | Allowed content change |
|---|---|---|---|
| `skills/claude-session-handoff/` (`d8939ab9…`) | `.agents/skills/claude-session-handoff/` | 3, all 100644 | none |
| `skills/vestive-design-skill/` (`7c7f65e3…`) | `web/.claude/skills/vestive-design/` | 81, all 100644 | none; `sha256sum -c references/shadcn-kit.sha256` passes from the new directory |
| none | `web/.agents/skills/vestive-design` → `../../.claude/skills/vestive-design` | 1, mode 120000 | the link target string only |
| `skills/project-archetype/` (`68a6ba80…`) | `skills/saas-foundation/` | 164: 159 at 100644, 5 at 100755 | exactly the lines below |

| Archetype file | Edits (baseline line numbers) | SHA-256 before → after |
|---|---|---|
| `SKILL.md` | L2 `name: project-archetype` → `name: saas-foundation`; L9 `# Project archetype: multi-tenant SaaS foundation` → `# SaaS foundation: multi-tenant platform archetype` | `9bebaa0a…219a5a` → `672c13c4…928aa` |
| `README.md` | L1 `# Project archetype` → `# SaaS foundation`; L24 `` `project-archetype/` `` → `` `saas-foundation/` `` and "use `project-archetype`" → "use `saas-foundation`"; L53 and L58 `skills/project-archetype/` → `skills/saas-foundation/` | `9b47ce24…362bb2` → `fe61b119…52f8d0` |
| `validation/run.sh` | L20 `REPO_ALLOW="^skills/project-archetype/"` → `"^skills/saas-foundation/"` | `a54c27e5…646ddc` → `256fc0bf…dbe25` |

**Retained with no change (R3):**
- the manifest schema `project-archetype/manifest/v1` and the managed-block markers;
- the render-cache and validation volume names;
- the Go and shell header comments and the internal mount path (`validation/run.sh` L2, L80);
- `SkillVersion` 0.1.0, the licence and the description;
- generator source, render output and `output_id`;
- templates, the manifest and historical hashes.

Personal installs are not touched.

## 3. Files to modify

### 3.1 Harness: `infra/validation/run.sh` (and the `infra/validation/lib.sh` L8 comment)

- **Path consumers.** `RENDER` (L99) → `skills/saas-foundation/generator/render.sh`. The Vestive asset root in `web_design_check` (L629) and `static_web_asset_controls` (L709) → `web/.claude/skills/vestive-design/assets`.
- **Web-application scans** (`web_sources` L442, `web_files` L559, the asset-control copy L711, the dependency-control copy L419). These exclude, through one shared predicate, exactly the Vestive tree `web/.claude/skills/vestive-design/` and the link `web/.agents/skills/vestive-design`. Every other `web/.claude/**` and `web/.agents/**` path stays in the scans.
- **Rendered-customization reconciliation** (correction 1):
  - the B-RENDER-PLAN count pin at L910 changes from `23` to `24`;
  - its message becomes "3 G001 + 5 G002 + 11 G003 + 1 G048 + 3 G004 + 1 G049";
  - `web/eslint.config.mjs` joins `APPLICABLE` (L132–), so the disabled and rename variants carry the customized file. B-TOGGLE-OFF-ADDITIONS requires each registered customization in a variant to equal the root file, and its rendered base to equal the manifest.

  B-RENDER-REPEAT then expects exactly 24 proposals. With the TypeScript contingency (§3.3), the pin is 25, the message ends "+ 2 G049", and `web/tsconfig.json` joins `APPLICABLE`.
- **Relocation table and guards** (as r1, now tied to binding (b)):
  - three inline source → destination pairs, three delta after-hashes and the link path and target;
  - active only while BASE contains a source path;
  - **B-SKILL-UNCHANGED** covers `skills/`, `.agents/skills/`, `web/.claude/skills/` and `web/.agents/skills/`, and verifies the relocation;
  - **B-PROTECTED** treats a relocated BASE path as satisfied only through that verified move.
- **B-CI-CONFIG.** The exact workflow `want` and the PASS text are updated (§3.2). Rules stay ten and jobs three; the profile counts are unchanged.

### 3.2 CI: map, selector, workflow, controls

**Map, selector and workflow.**
- **Map** (`profile-paths.tsv`):
  - three `docs file` rows for the exact handoff files;
  - `agent dir .agents/` and `agent dir .claude/`;
  - `foundation dir web/.claude/` and `foundation dir web/.agents/`.
- **Selector** (`select-profile.sh`):
  - **Class `agent`.** It selects fast and shares the docs path-wise add, delete and rename branch, with reason `agent-only`.
  - **Subtree-skill rule.** A path containing `/.claude/skills/` that does not start with `.claude/` or `.agents/` is foundation. It applies after exact file rows and before the Markdown rule.
- **Workflow** (the shared `&non_documentation` list only):
  - in the first-component complement, `[^dm]*` becomes `[^.dm]*` plus the exact `.agents` complement;
  - per-level complements under `.agents/` admit everything except the three handoff files;
  - `**/.claude/skills/**/*` admits nested Claude skills.

  MR and push rules share this anchor, so a single admission evaluation covers both (the oracle asserts that the lists are identical).

**Controls** (correction 3). Admission and profile are separate questions:
- **"Automatic admission"** is the workflow rule, evaluated by the reused Ruby oracle on the candidate `.gitlab-ci.yml`. **Omitted** means no automatic pipeline and no job.
- **"Selector if invoked"** is `select-profile.sh` (pipeline / gate / reason). For an omitted case it only says what an explicit or mixed invocation would select; it does **not** mean an automatic fast job runs.

The table rows use the existing table-driven `F_ROWS` and `f_row`, which evaluate the **detached MR-source context** (`gate`, `mr-source`, a valid diff base). The push contexts follow the table.

| Case | Change (fixture ops) | Automatic admission | Selector if invoked (MR source) |
|---|---|---|---|
| H1–H3 | modify each handoff file alone | omitted | fast / none / docs-only |
| H4 | all three plus root, `docs/` and `management/` Markdown | omitted | fast / none / docs-only |
| H5–H7 | delete each handoff file (seeded in the base) | omitted | fast / none / docs-only |
| H8 | exempt → exempt: delete `references/local-session-handoff.md`, rename `SKILL.md` to it | omitted | fast / none / docs-only |
| H9 | handoff → documentation (`docs/handoff.md`) | omitted | fast / none / docs-only |
| H10 | handoff → non-exempt agent path (`.agents/skills/claude-session-handoff/references/notes.md`) | **admitted** | fast / none / agent-only |
| H11 | handoff → `skills/example/SKILL.md` | **admitted** | full / full / deletion-or-rename |
| H12 | new `.agents/skills/claude-session-handoff/scripts/helper.sh` | **admitted** | fast / none / agent-only |
| A1–A4 | `.claude/settings.json`; `.claude/skills/example/SKILL.md`; `.agents/skills/example/SKILL.md`; `.agents/config.toml` | **admitted** | fast / none / agent-only |
| A5 | delete the seeded `.agents/skills/example/SKILL.md` | **admitted** | fast / none / agent-only |
| A6 | rename `.claude/settings.json` → `.agents/settings.json` | **admitted** | fast / none / agent-only |
| A7 | rename the root agent skill → `skills/example/SKILL.md` | **admitted** | full / full / deletion-or-rename |
| N1–N3 | new `web/`, `core/` and `docs/.claude/skills/example/SKILL.md` | **admitted** (N3 only through the nested-skill glob) | full / full / foundation |
| V1–V4 | Vestive `styles.css`; a kit `.tsx`; link retarget (new `ln:` op); `web/.claude/settings.json` | **admitted** | full / full / foundation |
| S1–S2 | `skills/saas-foundation/generator/render.go`; a template file; plus the existing `F-doc-skill` | **admitted** | full / full / foundation |
| M1 | `.claude/settings.json` + `web/app/page.tsx` | **admitted** | integration / integration / runtime |
| M2 | `.claude/settings.json` + `web/package-lock.json` | **admitted** | full / full / foundation |
| M3 | a handoff file + `web/app/page.tsx` | **admitted** | integration / integration / runtime |
| M4 | a handoff file + N3 | **admitted** | full / full / foundation |

**Push contexts**, using the existing `ctl_F_push_*` pattern with `branch-push` and a valid before-SHA:
- staging push and default-branch push for A1 (fast / none), H1 (fast / none, if invoked), N3 (full) and M1 (integration);
- the feature-branch push (fast job only; gate refused) and merged-result (→ full) controls are unchanged;
- explicit web, API and schedule contexts with H1 → full / full.

Existing missing-base, degenerate-base and other fail-closed controls are unchanged.

**Fixture changes.** `mk_fbase` seeds `.claude/settings.json`, `.agents/skills/example/SKILL.md`, the three handoff files, `web/.claude/skills/vestive-design/styles.css` and the Vestive link, so the deletion and rename cases have bases. `f_apply` gains `ln:PATH:TARGET`.

**Oracle cases** (one evaluation for MR and push):
- the existing 206;
- H1–H12, A1–A7, N1–N3 (plus `management/.claude/skills/x/SKILL.md`), V1–V4, M1–M4;
- near-misses, all **admitted**: `SKILL.mdx`, `SKILL.md.sh`, `skill.md`, `claude-session-handoffs/`, `.agent/`, `.agentsx/`, `openai.yml`, `references/other.md`, `references/x/local-session-handoff.md`, `agents/sub/openai.yaml`;
- the broadened-pattern negative control.

### 3.3 Web toolchain (correction 2)

**ESLint: exact exclusion only.**
- `web/eslint.config.mjs` keeps its existing ignore object and gains one global-ignores object: `[".claude/skills/vestive-design/**", ".agents/skills/vestive-design/**"]`, with a one-line G049 comment. It is a rendered file, so it becomes a registered customization (§3.1) with this register row: `web/eslint.config.mjs<TAB>ignores exactly the relocated Vestive agent skill and its Codex discovery alias (G049)`. The row is appended after the G004 rows.
- No other agent path is excluded. ESLint 9.39 flat config lints other `web/.claude/**` and `web/.agents/**` code as at baseline.

**TypeScript: baseline behaviour kept.**
- No change is planned, and the include set is not broadened. S2 records `tsc --listFilesOnly`; at baseline, TypeScript's wildcard discovery omits dot-directories.
- **Contingency, pre-reviewed and used only if Vestive files appear in that list:** `"exclude": ["node_modules", ".claude/skills/vestive-design", ".agents/skills/vestive-design"]`, plus its register row, `APPLICABLE` entry and the count 25.

**Tailwind and images.** Unchanged from r1:
- production CSS assets from the baseline and the candidate must be byte-identical (any difference → STOP);
- the final web image has no `.claude` or `.agents` path.

**Exact deltas** (SHA-256 at the planning baseline; recomputed with the same replacement at S0 if a base blob differs):

| File | Before → after |
|---|---|
| `web/eslint.config.mjs` | `8330fba8…9fb92a` → `75dff6f7…4132b7b` |
| `infra/validation/render-customizations.tsv` (24 rows) | `cde2fe61…837500` → `35a52fa5…629582` |
| `web/tsconfig.json` (contingency only) | `22207916…298ea4` → `376d01e7…0790c7` |

### 3.4 Documentation (live references only)

- **`docs/foundation/OPERATIONS.md`.** The CI policy and the "Agent skills" subsection. It covers layout, discovery, personal-copy precedence, the Windows symlink limitation and the historical mapping.
- **Path updates with "formerly" notes:**
  - `docs/foundation/BOOTSTRAP.md`;
  - `docs/foundation/WEB-SHELL.md` L108;
  - `README.md` L18.
- **`docs/ARCHITECTURE.md`** (Q2 decision; one sentence only).
  - **Edit at L4871:** "The accepted G002 shell applies `skills/vestive-design-skill` as visual guidance" becomes "…applies the Vestive design skill (now `web/.claude/skills/vestive-design`; formerly `skills/vestive-design-skill`) as visual guidance". SHA-256 `c0357d1c…5e1dfe` → `62e1c53c…ca18cb024`, bound by `--expect-hash`.
  - **Not changed:** the §58 tree (`skills/` is still correct), the G002/G048 history and the semantic decisions.
- **Not changed:**
  - AGENTS.md, CLAUDE.md and `docs/agents/**`;
  - management history;
  - the render manifest and templates;
  - every application file.

### 3.5 Schema and contracts

None: no database, API, MCP, identity or permission change. The CI contract changes are §3.2. G049's own diff selects **full**.

## 4. Validation and sequence

| Step | Work | Covering evidence |
|---|---|---|
| **S0** | Bind the execution baseline; re-hash the planning inputs and verify the three skill tree IDs; Codex fixes snapshot (a); scope allow-list | Commit and tree recorded; wrapper `--check-only` precheck |
| **S1** | Moves, link and archetype edits | **F1** inventory against binding (b) (path, mode, SHA-256; only the three after-hashes differ); `sha256sum -c` in the Vestive directory; frontmatter names; mode 120000 and `readlink`. **F3** `run.sh provenance`: B-RENDER-ID unchanged; B-RENDER-PLAN conflicts exactly the 24 registered paths and everything else unchanged; B-RENDER-REPEAT writes exactly 24 proposals with no overwrite; the link is present in the destination |
| **S2** | Harness, CI, ESLint, register and docs | **F4** provenance guards pass through the wrapper (precheck and postcheck), plus the four disposable negatives against the real guards: a changed Vestive byte, a removed archetype file, an extra file under the handoff destination, and a retargeted link. Each must fail. **F5** group F (§3.2 rows and push contexts) with B setup; the Ruby oracle; `run.sh lint` (B-CI-CONFIG, shellcheck, web boundary, design and assets); `glab ci lint` if authorized. **F6** in an owned disposable copy of `web/`: an ESLint error file inside `.claude/skills/vestive-design/` is **not** reported, while equivalent error files at `web/.claude/notes/zz.ts` and `web/.agents/zz.mjs` **are** reported; `npm run check`; `tsc --listFilesOnly` recorded; production-CSS identity; image path scan. B-TOGGLE-OFF-ADDITIONS and the rename variant carry the customized ESLint file. **F7** `git diff --check`; shellcheck |
| **S3** | Discovery in a disposable clone of the candidate | **Codex.** `codex debug prompt-input`; each entry names its root path, so the clone's `.agents/skills` (root) and `web/.agents/skills` (link) entries are distinguishable from the personal `~/.agents/skills` duplicate. **Claude (Q1).** An isolated, owned `CLAUDE_CONFIG_DIR` that excludes personal skills. Isolation is proven by the absence of a personal-only skill (`nextjs-setup`). A personal-copy listing never satisfies AC2. If `CLAUDE_CONFIG_DIR` does not isolate personal skills, STOP and report |
| **M1** (operator, before final CI) | In the isolated configuration, the operator signs in, runs `/skills` from the clone's `web/` (expecting `vestive-design` as a project skill) and from the root, and attests the result | Attestation bound to the candidate; the disposable configuration and its sign-in are removed afterwards; findings resolved before S4 |
| **S4** | Final qualification on the frozen candidate | One `all` through the wrapper. Then the MR pipeline after an owner-authorized push: fast plus a gate selecting **full** |
| **S5** | Result and cleanup | Owned clones, configurations, containers and images removed; B3 history |

## 5. Delegation

No implementation subagents. One read-only reviewer after F1–F7 recomputes the relocation inventory against binding (b) and §2/§3 hashes.

## 6. Cycles

Cycle 1: S1–S2. Cycle 2: fixes and the S3/M1 findings, then the checkpoint. Cycle 3: reserve. Stop conditions are unchanged from r1, plus the S0 tree-ID mismatch and the S3 isolation failure.

## 7. Acceptance-criteria mapping

| AC | Evidence |
|---|---|
| AC1 | S0 bindings (a) and (b); wrapper pre- and post-checks; no writes to G004 checkouts |
| AC2 | F1 modes and links; S3 Codex root-path entries; S3/M1 Claude isolated-configuration listing |
| AC3 | F1 against binding (b); `sha256sum -c`; B-PROTECTED with the reviewed ARCHITECTURE hash; history diffs |
| AC4 | Frontmatter; F3 render checks with exactly 24 conflicts and proposals; manifest unchanged; variants carry the customization |
| AC5 | F4 negatives; F6 exact ESLint scope (Vestive excluded, unrelated hidden code still linted) |
| AC6 | H1–H9 omitted by the oracle, with selector-if-invoked rows; explicit contexts → full |
| AC7 | §3.2 table in MR source; push contexts; oracle positives, near-misses and mixed cases |
| AC8 | Focused checks, then S4 local `all` and MR fast plus full |
| AC9 | Docs (§3.4); a Result with the inventories, controls, hosted IDs, no application delta, cleanup and B3 |

## 8. Questions and coordination

- **Q1–Q4 are resolved by the review** (above, and in §3.4 and §4): isolated discovery; the exact ARCHITECTURE literal; the reused oracle plus the exact pin; scoped complement globs.
- **New, for Codex.** Snapshot (a) proposes leaving PROGRESS and ROADMAP unpinned, because they are Codex's live records. Confirm, or pin them.
- **G005 and G050 coordination.** The shared files are `run.sh`, `validate-controls.sh`, `.gitlab-ci.yml` (G050 changes job `artifacts`; G049 changes only the workflow list), `profile-paths.tsv` (G049 only) and OPERATIONS.md. The hunks are disjoint. Whichever Goal integrates second rebases and reruns its covering checks, and no Goal blocks another. G005's register or `APPLICABLE` edits, if any, also reconcile the count pin.

## Revision record

- 2026-10-09 UTC, r1 (archived `-r1.md`, `13818145…`). Initial Plan.
- 2026-10-09 UTC, r2. Answers the Codex CORRECTION LOOP (`28390966…`):
  1. rendered ESLint customization: register 24 and `APPLICABLE`, with the TypeScript contingency reconciled;
  2. exact Vestive and alias ESLint exclusion only, with disposable proof that unrelated hidden code stays linted;
  3. MR-source table, push contexts, deletions and exempt-to-exempt and non-exempt renames, with admission versus selector made explicit;
  4. two bindings: an immutable snapshot excluding mutable relocation paths, and the commit-pinned relocation inventory.

  Also applies the Q1 and Q2 decisions. B3 and checkpoint 2 unchanged; planning only.
