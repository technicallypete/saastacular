> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G049 — Agent skill discovery, foundation naming and CI classification: Result (interim)

**Status: CYCLE-2 CHECKPOINT REACHED. The covering checks pass. M1 (the operator check) precedes the final local and hosted qualification. B3: 2 of 3 cycles used. Not submitted for acceptance.** The earlier status texts below are kept as history.

## Cycle 2 placement amendment (Codex `c323817e…`; binding `3c302941…`)

**The amendment.**
- Option B was approved. The two directives now sit after the complete foundation header comment, before the "B. Appearance-aware dark variant" comment, followed by a blank line.
- `web/app/globals.css` after-hash: **`577cfc66…`, exact**. The shadcn ejected block is verbatim; no guard exemption.
- The launcher now verifies binding `3c302941…`; the c2 launcher is kept as `g049-vw-c2.sh` (`3707979c…`).
- The 32-input snapshot, 23-expression scope and all other bindings are unchanged.

**Failed covering run (history).** `c2-lint-prov` (`08:32Z`): FAIL B-WEB-TOKENS, with the authorized cycle-2 placement inside the pinned shadcn block, and therefore B-REQUIRED-CHECKS. It also logged 85 "ignored null byte" warnings from my first correction-3 implementation; the fix is in `242a6f5`.

**Covering rerun.** `c2b-lint-prov` (`08:41Z`–`08:45Z`, placement binding; wrapper precheck 32 inputs, harness exit 0, postcheck passed): **PASSED**. 26 PASS, including:
- B-WEB-TOKENS, B-WEB-DESIGN and B-WEB-ASSET-CONTROLS (A1–A19);
- B-LITERALS, B-CI-CONFIG and B-RENDER-PLAN (24);
- B-PROTECTED and B-SKILL-UNCHANGED (relocation verified);
- B-SCOPE: 265 paths within the 23-expression scope;
- **B-REQUIRED-CHECKS: 25/25** for lint, provenance and final.

There were 0 null-byte warnings.

**Exact emitted CSS and toolchain** (`$EV/f6c2/`, the candidate with the Option B bytes, owned images):
- production CSS **byte-identical to BASE** (`907bcd99…`, 1 file);
- `tsc --listFilesOnly`: 0 `.claude`/`.agents` inputs;
- ESLint scope PASS: the Vestive tree is not linted, while unrelated hidden code at `web/.claude/notes` and `web/.agents` is reported;
- `npm run check` exit 0;
- runtime image: 0 agent paths.

**Cycle accounting** (BUDGET_POLICY §3: failure, diagnosis, corrective implementation and retest):
- **Cycle 1:** the implementation, whose F3 and F6 failures were corrected in cycle 2.
- **Cycle 2:** the authorized corrections. Their covering run failed on B-WEB-TOKENS; the placement correction was retested and passed.

**B3: 2 of 3 used.** Cycle 3 is unused and held in reserve.

## M1: operator Claude discovery (isolated configuration)

**Set-up.** The owner signed in to an owned, disposable `CLAUDE_CONFIG_DIR=$EV/s3/claude-config`, then ran `/skills` in the fresh clone of `242a6f5` (Claude Code 2.1.292).

**Operator feedback, verbatim:**
- "in location 1 only vestive-design, location 2 all default skills plus vestive-design".
- Clarification: "vestive design is also a global skill inhome directory, it did not have nextjs-setup for loc 2".

**Read-only attribution** (the isolated `history.jsonl`, file creation times and project records; no credentials read). The recorded transcripts do not contain the `/skills` output itself; they hold only the exit messages.

| Location | Where it actually ran | When | Skill sources present then | Conclusion |
|---|---|---|---|---|
| 1 | `…/s3/clone/web` | `/skills` 04:51:09 local | none in the isolated config: `skills/synced` was created at 04:51:59.87 and the account-synced `vestive-design` at 04:52:00.68 | The listed `vestive-design` can only be the clone's `web/.claude/skills/vestive-design`. `nextjs-setup` absent, so personal `~/.claude/skills` are isolated. **Claude web discovery: PASS** |
| 2 | `<home>/Code/VitruvianTech/Beginify` (the history's project path; not the clone; no `.claude` there) | `/skills` 04:52:18 local, after the sync | the claude.ai account skills synced into the isolated config at sign-in: docx, xlsx, pptx, pdf, skill-creator, deep-research and others, plus an **account-synced `vestive-design`** | The "default skills plus vestive-design" are account-synced. `nextjs-setup` absent, so personal skills are again isolated. Not a clone-root check, which AC2 does not require (Spec: Vestive discovery is web-scoped) |

**Finding.** The owner's claude.ai account carries a same-named `vestive-design` that syncs into any signed-in configuration. Together with the personal `~/.claude/skills/vestive-design`, this means that outside an isolated, not-yet-synced configuration, `/skills` cannot attribute the source by name. Repository discovery is unaffected.

**Kept until final cleanup:** the isolated configuration (it holds the sign-in) and the clone.

**Codex checkpoint review:** `G049-cycle-2-checkpoint-codex-review.md` (`82b3e706…`), ACCEPT for the checkpoint and discovery only. M1 is sufficient. The Plan's manual Claude root repetition is waived as redundant to AC2. The account-sync caveat was added to OPERATIONS.md (`c51aa22`). B3: 2 used, 1 available.

## Read-only reviewer (subagent; independent commands; no writes)

On candidate `a8a26f1`, all six items **PASS**:
1. **Relocation.** All 248 rows match; 245 identical plus 3 reviewed after-hashes; the diffs touch only the Plan §2 lines.
2. **Link.** Mode 120000 and the exact target; it is the only entry under `web/.agents/`.
3. **Reviewed hashes.** ESLint, the register (24 rows), ARCHITECTURE, globals.css and the allow-list all match; `web/tsconfig.json` is unchanged.
4. **Scope.** 513 changed paths, all within the 23 expressions; no snapshot path changed.
5. **Guard and CI diff.** No removed assertion and no broadening beyond the approved exclusions.
   - The reviewer's Python emulation of GitLab `File.fnmatch` found 0 mismatches over 391,225 spec paths.
   - Comparing BASE with the candidate over 4.29M paths: exactly the 3 handoff files lost automatic admission, and all 344 newly admitted paths are nested `.claude/skills`.
6. **Application files.** Only `globals.css`, `eslint.config.mjs` and the approved link changed.

**Low findings and their dispositions:**
1. **Fixed (`c2c`).** B-SKILL-UNCHANGED had lost BASE's fail-closed behaviour on a `git diff` error, because the list was read through process substitution and the exit status discarded. Its git listings, and those of `skill_relocation`, now go through files or report a problem, so a failing git command fails the check.
2. **Fixed.** The alias check now requires non-empty `realpath` results; two failing calls no longer compare equal.

   Covering run `c2c-prov`: B-RENDER-ID, B-RENDER-PLAN, B-RENDER-REPEAT, B-PROTECTED and B-SKILL-UNCHANGED all PASS; the wrapper postcheck passed. The exit is 1 only because B-SECRET-CANARY is UNMET in a provenance-only run (no secrets are generated), as in F4 P0.
3. **Recorded; no code change.** Plan §3.2 row H10 lists the reason `agent-only`. The control expects `docs-only`, which is the selector's deterministic tie-break: at equal rank, the first path in byte order wins, and `SKILL.md` sorts before `references/…`. Profile, gate and admission are as planned.
4. **Recorded; no code change.** There are 5 push-context controls, not 8: A1 on staging and main, H1 on staging, N3 on main, M1 on staging. Staging and default-branch pushes share one selector path (`select-profile.sh` `branch-push`).

**For Codex (no change made).**
- Nested Codex `.agents/skills/**/*.md` under `docs/` or `management/` still follow the documentation bypass, as at BASE. The approved rule (d) covers nested `.claude/skills` only.
- Hosted CI runs the harness with `--base HEAD`, so relocation evidence (AC3) comes from the local bound runs, by design.

## S4: final local `all`, CANCELED at the owner's instruction (no pass claimed)

- **Frozen candidate:** `42c785c7d6c03130666b2e06a03a4fa9733172ef`, tree `b4ed07149e18d90b4d7035324bcb1dfba9b26fdc`.
- **Run `s4-final`** (placement-bound wrapper; precheck passed, 32 inputs; harness run `453115`): started at `09:00:46Z`.
- **Owner instruction, verbatim, relayed by the Codex session:** "Just cancel the local runs and make the mr". I sent SIGTERM to the harness, which ran its own EXIT cleanup.
- **Outcome:**
  - harness exit **143** (terminated) at `09:04:23Z`, while still in the `lint` suite;
  - 16 PASS and 0 FAIL recorded up to that point;
  - wrapper postcheck passed.
- **Owned resources:** verified afterwards that no `g001.run`-labelled container, volume, network or image remains, and the run's work directory is gone.
- **The local full run did not complete and is NOT claimed as passing.**

**Owner waiver** (relayed; Codex will record it in the supervisor record): the successful local `all` (AC8 and Plan S4) is waived for G049 only. The focused checks, binding integrity, the read-only review and the required hosted fast/full evidence on the MR pipeline remain.

## MR !8 and the rebase onto remote staging (owner direction; Codex authorization `2792a762…`)

**MR and first pipeline.**
- MR !8 (`g049-skill-layout-ci` → `staging`, source-branch deletion enabled) was opened on `a68ba40`.
- Its pipeline 2929783487 (`merge_request_event`): `fast` **success** (326 s); `gate` **canceled** at about 796 s. I did not cancel it; the cancellation coincided with this rebase. It is historical, and **not** evidence for the rebased candidate.
- MR !8 then showed conflicts, because staging had moved.

**Rebase.**
- Binding `d706e995…`: 32-input snapshot `afbbea6d…` (only AGENTS.md, CLAUDE.md and PROGRESS.md hashes changed), the same 23-expression scope, inventory, Plan, review and after-hashes.
- Freshly fetched `origin/staging` = target `761a8dca5035a3c8d505d70325d5b7d4b9412b90` (tree `0857bd31…`). The old source `a68ba40` was rebased onto it: six commits, new tip before this record `0267e94`.
- Launcher switched to `d706e995…`; the placement launcher is kept as `g049-vw-c2p.sh`.

**Conflicts and their resolution:**
- `.gitlab-ci.yml`: the header comment only; both glob lists merged automatically. The result is 54 globs: the G049 first-component complement, the ten `.agents/` handoff carve-out globs and the nested-skill glob, the `docs/**` globs, and the target's `management/` planning-JSON globs.
- `infra/validation/run.sh`: the B-CI-CONFIG `want`. Resolved mechanically: the target's pin, proven to round-trip, plus exactly the G049 glob transformation. The result is asserted equal to the merged YAML list for both MR and push.
- `docs/foundation/OPERATIONS.md`: the docs bullet combines the target's planning-JSON wording with G049's handoff-files sentence. The agent and nested-skill bullets and the target's MR-validation paragraph are kept.
- `infra/ci/select-profile.sh` merged automatically: G049's subtree-skill rule runs ahead of the target's widened documentation rule.

**Checks after the rebase:**
- No conflict markers.
- `git diff a68ba40..0267e94` contains only the target's own changes.
- All reviewed after-hashes are intact: ESLint, register, ARCHITECTURE, `globals.css`, scope copy and the three archetype files.
- Wrapper precheck: 32 inputs, `afbbea6d…`, passed.

**Focused results recorded before the owner stopped further local checks:**
- **Admission oracle on the merged workflow** (Ruby `File.fnmatch`; the policy extended with the target's planning-JSON exemption; `$EV/oracle-rebase/`): **PASS**, 312 path cases. G049's handoff, nested-skill and root-agent rules and the target's planning-JSON exemption all hold.
- **Concrete pre-existing TARGET defect (not G049; reported, not changed).**
  - **The glob.** In the target's own workflow (`9577149`, unchanged in `761a8dc`; #33 there, #44 after the merge): `management/v0/plans/{[^G]*,…,G[0-9][0-9][0-9]-planning-input[^s]*}{d,md}`. Its alternatives end in `*`, so it admits **every direct `management/v0/plans/*md` file**, including the Plans themselves, to automatic pipelines.
  - **Evidence:** `$EV/oracle-rebase/diag/`. Four paths are explained solely by this glob.
  - **Effect:** it fails open (extra CI); nothing is bypassed.
  - **Probable one-line fix,** for the target's owner: `management/v0/plans/{d,md}`.
- **Group F** (`$EV/ctl-f-rebase/`): 127 of 128. `F-schedule-other` aborted in fixture setup, the known standalone limitation; rerun `--only N12q,F-schedule-qualification,F-schedule-other`: 3/3 PASS.
- **`rb-lint-prov`** (rebase binding): **stopped at the owner's instruction**, exit 143. Up to that point, 12 PASS and 0 FAIL, including B-LITERALS, B-WEB-TOKENS and B-WEB-DESIGN. B-CI-CONFIG, B-SCOPE and the provenance and relocation checks had **not** run, and **no pass is claimed** for them. Wrapper postcheck passed. No `g001.run`-labelled resources remain, and the work directory is gone.

**Owner direction.** The remaining local checks duplicate the hosted coverage. The hosted fast and full pipeline on the rebased candidate supplies them.

## Cycle 3 (the last B3 cycle): GitLab's 50-entry `rules:changes` limit (Codex correction `92903f59…`)

**Failure.** The rebased candidate `5a01fcd` was pushed to MR !8 with an exact lease on `a68ba40`. GitLab then created pipeline **2929888600**, which failed at creation with **zero jobs**. GraphQL `failureReason`: "The pipeline failed due to an error on the CI/CD configuration file". `glab ci lint`: "workflow:rules:rule:changes config has too many entries (maximum 50)", for the MR and the push rule.

**Cause.** The merged `&non_documentation` list had 54 entries: G049's 24 combined with the target's 43. Neither branch exceeded the limit alone. Codex records that its planning-JSON exemption missed this constraint.

**Correction** (B3 cycle 3; same binding `d706e995…`, BASE `761a8dc`, 32 inputs and 23 scope expressions):
- **`.gitlab-ci.yml`.** Each same-prefix non-Markdown quartet `P*[^d]`, `P*[^m]d`, `P*[^.]md`, `P{d,md}` collapses to `P{*[^d],*[^m]d,*[^.]md,d,md}` (FNM_EXTGLOB brace, no spaces).
  - Eight quartets were merged: root, `docs/**/`, `management/`, `management/{non-v0}/**/`, `management/v0/`, `management/v0/{non-plans}/**/`, `management/v0/plans/*/**/` and the plans-name complement.
  - **54 → 30 entries.** The G049 handoff, root-agent and nested-skill globs and the planning-JSON globs are byte-identical.
- **`run.sh` B-CI-CONFIG.** The workflow pin is regenerated from the YAML list. A new assertion requires every `rules:changes` list, aliases resolved through `explode(.)`, to have at most 50 entries.
  - Checked: the pinned yq reports 30 and 30 for the candidate, and 54 and 54 for the pre-fix configuration, so the check catches this failure.
  - The PASS text names the planning-JSON omission and the limit.

**Checks** (Codex minimum; no local full run and no group-F rerun):
- **Equivalence** (`$EV/glob-limit/equiv.log`; Ruby `File.fnmatch` with GitLab's flags): the 54- and 30-entry lists admit identically over 624 generated paths.
- **Rebase oracle:** PASS on 312 path cases with 30 patterns, MR and push identical. The same pre-existing target `plans/*md` defect is preserved, not introduced (4 paths, explained only by that glob).
- **`glab ci lint .gitlab-ci.yml`:** "CI/CD YAML is valid".
- `bash -n`, `git diff --check`, no conflict markers.
- Wrapper precheck: 32 inputs, `afbbea6d…`.

**B3: 3 of 3 cycles used (exhausted).** Any further failing correction is reported to Codex before more work; no reset.

**Pipelines after the cycle-3 push:**
- **2929905269** (`merge_request_event`, `bf6781a`): created normally. `fast` **success** (366 s); `gate` **canceled** at about 495 s, by someone outside this session (I did not cancel it). It is history only: no full-gate result exists for `bf6781a`.
- **2929888555** (a feature-branch push pipeline for `5a01fcd`) also failed, from the same 54-entry config error.

## Second rebase onto staging `0ba2054` (owner direction; Codex authorization `7e55e846…`)

**Why it was needed.**
- The separate owner-directed staging one-off `0ba2054` (`CI-plans-markdown-admission-fix-2026-10-09.md`) corrected the target's direct-plans globs.
  - Every direct `plans/*.md` file had been admitted; now it is omitted.
  - Short non-documentation names had been omitted; now they are admitted.
- It also compressed that list to 10 entries. MR !8 conflicted again.
- Its staging pipeline 2929943945 was canceled outside this session during `fast`.

**This pass.**
- One separately tracked integration reconciliation; **not** a B3 cycle and no reset.
- Binding `fc100c44…`: BASE `0ba20543aca973bfbbe9173482d4af7537fc5910` (tree `eff3de02…`). The 32-input snapshot `afbbea6d…`, 23-expression scope, 248-entry inventory and protected after-hashes are all unchanged.
- Old source `bf6781a` was rebased onto `0ba2054`.

**Resolution** (deterministic, `$EV/bin/g049-rebase2-resolve.py`, applied to both conflicting commits):
- **Workflow list.** The target's corrected 10-entry list, plus exactly the G049 transformation: the `.agents`-aware first-component complement in place of the target's, and the 11 G049 `.agents`/nested-skill globs after it. **21 entries**, the same for MR and push.
- **`.gitlab-ci.yml` header.** G049's combined comment.
- **`run.sh`.**
  - The B-CI-CONFIG pin is regenerated from that list.
  - Exactly **one** alias-aware ≤50 guard: the target's and G049's identical guards are de-duplicated.
  - One PASS phrase naming the documentation Markdown, planning-input JSON and exact handoff-file omissions, nested Claude-skill admission and the limit.
- **The replayed cycle-3 commit** now carries only its Result text; its CI content is subsumed by the target plus this resolution.

**Readiness checks** (the authorization's minimum; no full, group-F or oracle run):
- no conflict markers; `bash -n`;
- the target is an ancestor of the candidate;
- wrapper precheck: 32 inputs, `afbbea6d…`;
- protected and reviewed after-hashes intact (ARCHITECTURE `62e1c53c…`, `globals.css` `577cfc66…`, scope `616ecc5e…`, ESLint `75dff6f7…`, register `35a52fa5…`);
- pinned yq entry counts: **21 and 21**;
- `glab ci lint`: valid;
- 265 changed paths against the target, the same as against the previous base.

Hosted fast and full evidence must bind the new candidate. Earlier pipelines are history only.

**Next, in order:**
1. ~~M1~~ done (above).
2. ~~The read-only reviewer~~ done (above).
3. ~~Local `all`~~ canceled and waived (above).
3. S4: the final local `all` on the frozen candidate, then the push and MR pipeline.
4. The final Result.

---

**Earlier status (history): CYCLE-2 CHECKPOINT. One decision for Codex (B-WEB-TOKENS and the authorized globals.css placement); M1 operator check pending. Cycles 1 and 2 of B3 used. Not submitted for acceptance.**

## Cycle 2 (Codex authorization `b51e1354…`; replacement binding `c10e5ca7…`)

**Binding.**
- Replacement bundle `g049-execution-approved-c2/`: 32-input snapshot `4139c7eb…` (only `web/app/globals.css` removed), 23-expression scope `616ecc5e…` (only `^web/app/globals\.css$` added). The `globals.css` after-hash is `4418ef7d…`; ARCHITECTURE stays `62e1c53c85b0…`.
- The launcher now verifies `c10e5ca7…` and passes every `expect_hash_protected` entry. The cycle-1 launcher is kept as `g049-vw-c1.sh` (`dc8c50eb…`), and the original bundle is untouched.
- Wrapper precheck: 32 inputs passed. The scope copy is `616ecc5e…`.
- The supervisor record was pushed to `origin/staging` as `aef09c3`; this bound checkout was not pulled or updated.

**Corrections applied:**
1. **`web/app/globals.css`:** the two authorized `@source not` lines right after `@import "tailwindcss";`, after-hash `4418ef7d…` exactly. The emitted CSS of this exact edit was proven byte-identical to BASE in cycle 1's disposable remedy build (`907bcd99…`, 32,860 bytes).
2. **`static_literals`:** hits under exactly the Vestive subtree or alias are filtered with the existing `vestive_skill_path` predicate.
3. **`web_design_check`:** DESIGN_LITERAL scans `web/eslint.config.mjs` with exactly its two reviewed G049 lines removed (`eslint_literal_text`). Each line must occur exactly once; otherwise the whole file is scanned and `DESIGN_LITERAL_EXEMPT` is reported. Asset controls A14–A19 were added:
   - A14: another literal in the ESLint configuration;
   - A15–A17: a duplicated, changed and missing reviewed line;
   - A18: a literal in unrelated hidden web code;
   - A19, positive: a literal inside the excluded Vestive tree.

   A first implementation read every web file through command substitution, which raised 85 "ignored null byte" warnings on binary files. Fixed: only the ESLint file is read that way; every other file is grepped directly, as at baseline.

**Covering runs and focused evidence:**

| Run | Result |
|---|---|
| `c2-lint-prov` (c2 wrapper, `lint provenance`; pre- and post-check passed) | 24 PASS: **B-LITERALS**, **B-WEB-DESIGN**, **B-WEB-ASSET-CONTROLS** (A1–A19), B-RENDER-PLAN (24), B-PROTECTED, B-SKILL-UNCHANGED, B-CI-CONFIG and B-SCOPE (265 paths in scope). **FAIL: B-WEB-TOKENS** (below), and therefore B-REQUIRED-CHECKS. It ran before the null-byte fix |
| B-LITERALS scan negatives (`$EV/literals-neg/`; the real `static_literals` and `vestive_skill_path` extracted byte-for-byte) | L1, a literal in `web/.claude/notes/x.md`: **FAIL**, reporting only that file. L2, a literal only inside the Vestive tree plus the alias: **PASS** |
| F4 on disposable clones of local commit `242a6f5` (`$EV/f4/`; real `run.sh provenance`) | P0, unmutated: B-SKILL-UNCHANGED and B-PROTECTED PASS. Its exit is 1 only because B-SECRET-CANARY is UNMET in a provenance-only run. N1, a changed Vestive byte: **FAIL** (destination differs). N2, a removed archetype file: **FAIL** (missing destination; protected file removed). N3, an unbound handoff file: **FAIL**. N4, the alias retargeted to `../../../core`: **FAIL** |
| S3, Codex (`$EV/s3/`; fresh clone of `242a6f5`, no bootstrap; `codex debug prompt-input`, codex-cli 0.160.0) | Clone root: the project skill `claude-session-handoff` from the clone's `.agents/skills`. `web/`: additionally `vestive-design` through the clone's `web/.agents/skills` (link mode 120000 → `../../.claude/skills/vestive-design`). Personal roots are listed separately; source attribution is by path |

**Decision for Codex: B-WEB-TOKENS (a newly found conflict, so affected work is stopped).**
- **The conflict.** B-WEB-TOKENS pins the verbatim shadcn@4.21.1 ejected block of `globals.css`: lines 1 up to the foundation header, digest `16693ac5…` (`GLOBALS_EJECT_SHA`). The authorized placement, right after line 1, falls inside that block, so the pin fails ("globals.css ejected block differs from the shadcn@4.21.1 output"). Both remedies need Codex.
- **Option B (recommended): place the same two lines in the project-owned foundation section,** right after its header comment. No guard change is needed, and the shadcn block stays verbatim.

  Evidence (`$EV/css-placement/`):
  - after-hash `577cfc66487d71467cb82d6c7b7bf3bc7d6a0b104f90bb9b18c4e302b4b16646`;
  - the ejected block equals the pinned digest;
  - emitted production CSS is byte-identical to BASE (`907bcd99…`).

  It needs a new `globals.css` after-hash in the binding.
- **Option A:** keep the authorized placement. B-WEB-TOKENS would hash the ejected block with exactly those two lines removed, requiring each exactly once at lines 2–3. That is a guard exception analogous to correction 3.

**Pending:**
- **M1:** the Claude isolated-configuration discovery check. It needs the operator's readiness and sign-in in an owned, disposable `CLAUDE_CONFIG_DIR`.
- The read-only reviewer on the final candidate.
- The cycle-3 resolution of B-WEB-TOKENS.
- S4: the final local `all`, then the MR pipeline.

The toolchain evidence from cycle 1 (`tsc` inputs, ESLint scope, `npm run check`, runtime image) is unaffected by the two CSS directives and is re-covered by S4.

**Cycle accounting.** Cycles 1 and 2 used. Cycle 3, the last, is reserved for the Codex-decided B-WEB-TOKENS resolution and final qualification.

**Communication.** My acknowledgement and early stop notice to the platform-37 session was held undelivered; this transcript is the record.

---

**Earlier status (history): STOPPED for Codex. A Plan stop condition fired (production CSS differs). Cycle 1 of B3 used. Not submitted for acceptance.**

## Binding

| Item | Value |
|---|---|
| Base | `4062844361aae360653b978ace9acc3582d5cb09`, tree `c41c6c3ce71ceb32bd45fbe249945e23bd1357ee` |
| Checkout and branch | `.claude/worktrees/g049-exec`, `g049-skill-layout-ci` (uncommitted candidate) |
| Approval | `G049-execution-binding-codex-confirmation.md` `a89199de…`; bundle `execution-binding.json` `329670ca…` |
| Snapshot | 33 inputs, `6a2c43df…`; the wrapper precheck passed through the bound launcher `$EV/bin/g049-vw.sh` |
| Scope | 22 expressions, `a9ce83a8…`, copied byte-exactly to `management/v0/results/G049-scope-allowlist.txt` |
| Relocation inventory | 248 entries, `0ef9620e…` |
| Plan and review | `a1d8a313…` and `0a9f8d0b…`, unchanged |

Evidence: `$EV` = `<home>/.local/state/platform-evidence/g049-exec-4062844/`.

## Cycle 1: implementation (S1–S2) and focused checks

**S1:**
- the three `git mv` relocations;
- the relative link `web/.agents/skills/vestive-design` → `../../.claude/skills/vestive-design` (mode 120000);
- the seven archetype line edits;
- the scope copy.

**F1** (`$EV/f1/`): 248 inventory rows, 245 byte-identical plus 3 reviewed after-hashes (`672c13c4…`, `fe61b119…`, `256fc0bf…`), with modes preserved. No source remains and no destination file is unbound. The link resolves to the Vestive directory, the kit's `sha256sum -c` passes from the new location, and the frontmatter names equal the directory names.

**S2 changes** (all in scope):
- **`run.sh`:**
  - the renderer path and the Vestive asset root;
  - exact Vestive/alias exclusion in `web_sources`, `web_files`, the asset-control copy and the dependency-control copy;
  - register pin 24 and `APPLICABLE` + `web/eslint.config.mjs`;
  - the inline relocation table with B-SKILL-UNCHANGED over the four skill roots, and B-PROTECTED accepting a removed base path only through the verified move;
  - the B-CI-CONFIG workflow pin and PASS text.
- **`lib.sh`:** the comment.
- **`profile-paths.tsv`:** three `docs file` rows, `agent dir .agents/` and `.claude/`, and `foundation dir web/.claude/` and `web/.agents/`.
- **`select-profile.sh`:** the `agent` class (fast, `agent-only`) and the subtree-skill rule ahead of the Markdown rule.
- **`.gitlab-ci.yml`:** 24 globs, with MR and push identical. The first-component complement excludes exactly `.agents`, ten `.agents/` complements carve out the three handoff files, and `**/.claude/skills/**/*` is added.
- **`validate-controls.sh`:** fixture seeds, an `ln:` op, 32 table rows (MR source) and 8 push or explicit-context controls.
- **Exact reviewed bytes:**
  - `web/eslint.config.mjs` (`75dff6f7…`);
  - the register (`35a52fa5…`, 24 rows);
  - `docs/ARCHITECTURE.md` (`62e1c53c85b0…`, the approved after-hash).
- **Docs:** OPERATIONS.md (classification, cost and an Agent skills section), BOOTSTRAP.md, WEB-SHELL.md and README.md.

**Focused checks:**

| Check | Result |
|---|---|
| Syntax, `git diff --check` | pass |
| Admission oracle (reused Ruby `File.fnmatch` procedure, `ruby@sha256:50133d02…`, the same image as the documentation-bypass record; `$EV/oracle/`) | **PASS**: 294 path cases and 24 patterns, MR and push identical. Handoff-only sets, exempt renames and deletions are omitted; handoff → agent, handoff → skills, mixed, root-agent delete and rename are admitted. The broadened-pattern negative control fires |
| `validate-controls.sh --group F` (`$EV/ctl-f/`) | 127 of 128 pass, including all 40 new controls. `F-schedule-other` aborted in fixture setup ("a branch named 'staging' already exists"): the standalone-F limitation recorded in G004's integration record. Rerun `--only N12q,F-schedule-qualification,F-schedule-other`: **3/3 PASS** |
| F3: `g049-vw.sh f3-lint-prov lint provenance` (wrapper; pre- and post-check passed; `$EV/f3-lint-prov/`) | 22 PASS, including B-RENDER-ID, B-RENDER-PLAN (24 conflicts), B-RENDER-REPEAT (24 proposals), B-PROTECTED (248 relocated through the verified mapping), B-SKILL-UNCHANGED (relocation verified), B-CI-CONFIG, B-SHELL, B-WEB-BOUNDARY, B-SECRET-CANARY and B-SCOPE (263 paths, all in scope). **FAIL:** B-LITERALS, B-WEB-DESIGN, B-WEB-ASSET-CONTROLS, and therefore B-REQUIRED-CHECKS (findings 2 and 3) |
| F6 (owned disposable copies and images; `$EV/f6/`) | `tsc --listFilesOnly`: 0 `.claude`/`.agents` inputs, so the TypeScript contingency is **not** needed. ESLint scope **PASS**: the planted error in the Vestive tree is not linted (0 Vestive or alias files linted), while planted errors at `web/.claude/notes/zz.ts`, `web/.agents/zz.ts` and `web/.agents/zz.mjs` are reported. `npm run check` exit 0. The runtime image has 0 agent paths. **Production CSS DIFFERS** (finding 1) |

## Findings for Codex

1. **STOP: production CSS changes (Plan §3.3 stop condition; R6).**
   - **Cause.** Tailwind 4.3.3's automatic source detection scans the relocated kit under `web/.claude/skills/vestive-design`.
   - **Effect.** The candidate's production CSS grows from 32,860 to 45,169 bytes: 194 added selectors, none removed (`$EV/f6/css-diagnosis.txt`).
   - **Attribution.** The same candidate with `web/.claude` and `web/.agents` removed builds byte-identical CSS (`907bcd99…`).
   - **Tested remedy, on a disposable copy only:** two lines after `@import "tailwindcss";` in `web/app/globals.css`: `@source not "../.claude";` and `@source not "../.agents";`. They restore byte-identical CSS (`$EV/f6/css-remedy.txt`). They exclude agent guidance from CSS generation only; lint and type coverage are unchanged.
   - **What it needs from Codex:**
     - `web/app/globals.css` leaves the immutable snapshot (rebinding);
     - the scope gains `^web/app/globals\.css$`.

     globals.css is already a registered customization, so the register count is unchanged. Its register reason text could gain "(G049)"; Codex decides.
   - **Alternatives (not recommended):** move Vestive out of `web/` (contradicts the Spec layout); `source(none)` with explicit sources (a larger CSS-configuration change); accept the CSS change (not permitted).
2. **B-LITERALS: a missed web-scan site.** `static_literals` greps `web` recursively, and three Vestive files contain the product name. **Ready fix** under the approved §3.1 principle: filter the hits under exactly `web/.claude/skills/vestive-design/`, as in the other web scans.
3. **B-WEB-DESIGN and B-WEB-ASSET-CONTROLS: the reviewed ESLint lines contain "vestive"**, which the G048 `DESIGN_LITERAL` scan forbids in web files. Narrowing the ESLint pattern would break correction 2. **Proposed fix (new, needs review):** in `web_design_check`, test `web/eslint.config.mjs` for literals with exactly its two reviewed G049 lines removed, so that any other literal in that file still fails. The two remedy lines in finding 1 contain no literal and need no exemption.

**Cycle accounting.** Cycle 1 used; zero consumed before release. The fixes for findings 1–3 would be cycle 2, followed by the checkpoint report.

**Not yet run:**
- F4's disposable negatives;
- S3 discovery and M1;
- the read-only reviewer;
- S4 and the hosted run.

Owned resources are kept for reuse until cleanup: the images labelled `g049.run=4062844f6` and the pulled `ruby:3.4-alpine`.
