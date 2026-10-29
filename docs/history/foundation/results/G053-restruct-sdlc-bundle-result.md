> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G053 Result — Reusable SDLC management skill and Restruct bundle

Status: **attempt 2 of B3: corrections C1–C3 (§10) and the AC13 observer (§§11–12) submitted for Codex implementation review; final AC reconciliation in §12.4.** Attempt 1 was submitted as `88e07f5`. Author: Claude Code (G053 implementer, logical task/role G053 / impl).
Authority: accepted [Plan r2](../plans/G053-restruct-sdlc-bundle-plan.md) (SHA-256
`f0928ed1c7310b7ad4ae21b9875b106706a58e81824574fb47359b0d62798b49`), the
owner workflow amendment (not included) (`f27dc19ace786927dff504f0dd71b3552089d2d13318af5f008eee86486ab4c0`)
and the dispatch/reconciliation ACCEPT (not included)
(`c02aee01accf88c90df4b7ef4646c085daa8769d72631ca432e267f75c29fad6`).

## 0. Execution addendum and change map (recorded before substantive work)

### 0.1 Dispatch, ownership and naming exception

| Field | Value |
|---|---|
| Logical task/role | G053 / impl |
| Actual path | `.claude/worktrees/agent-<id>` (native generated basename; Codex G053-specific exception) |
| Branch | `worktree-agent-<id>` (kept; reset onto `origin/staging` because it held no task commits) |
| Owner | agent `agent-<id>`, harness lock pid 74419, under coordinator session `580cdb2f` |
| HEAD at dispatch | `dc0977666db63f06b5c12586b0f58438fb2969d8` (older `main`, expected) |
| Mapping record | private `g053-exec-24271ac22f/dispatch-record-impl.md` under the platform evidence root, SHA-256 `4b5b32500a377709456821780c2c9857587c6087989cb79b42df1be549de7b04` |
| Activation | coordinator message "MAPPING ACTIVE 4b5b3250…"; my read-only `pwd`, `git rev-parse --show-toplevel/--abbrev-ref HEAD/HEAD` and `git worktree list --porcelain` matched every field before any write |
| Effort | high (coordinator record from native metadata; not independently observable from inside the session) |

No write, fetch or reset happened before activation.

### 0.2 S0 reconciliation

- `git fetch origin`; target `origin/staging` = `3a3f4b87528028d71b1f067e1657ff14f8b49cfc`. Candidate after
  `git reset --hard origin/staging` = the same commit, tree `d2e17b157fc930c234ee43cface7ac004d980749`.
- Plan r2 `f0928ed1…` and the r1 archive `0d78814a873d2af3b97225edf881846cc4c045d8f46e9d17a9bfc913f993f84b` are
  byte-identical on staging. Reviews: r1 `d84e8d7f…`, r2 `8a4973b0bec0aeafe3856d6831c103e2529762e8f01785d758569a77c661b28e`.
- `sha256sum -c` of the 29 §2 rows at `3a3f4b8`: 21 match; exactly the eight expected rows changed. Their current
  committed hashes:

| Path | r2 binding | Reconciled at `3a3f4b8` |
|---|---|---|
| management/v0/goals/G053-restruct-sdlc-bundle.md | `3c9fc42a…` | `2d8702b26cbbc8a0909556f96a4387d07271de14f8365ac69c1c8d8e7f1b4b46` |
| management/v0/specs/G053-restruct-sdlc-bundle-spec.md | `c8401052…` | `10df57a86bfbd9d9b8b1ed04874041df3e7e0b8e03b5a83a9341c0d6930a479a` |
| management/v0/PROGRESS.md | `ec58a86b…` | `852e2097f3f0a5b32f601cf910f28cb93642c21e4fa7df6a7f52ad0b6eabc562` |
| management/SUPERVISOR-HANDOFF.md | `8401118d…` | `335582796c66489a7335aa4348797c8f1f71419d9d556295d46055dd51c4088b` |
| AGENTS.md | `c1671d4f…` | `4dc0ac8e9fcc14d49e3eff5e8ee3016dc69ea7f22abfa743e4032dc16dcad4c4` |
| CLAUDE.md | `98a54087…` | `b3ce5aa90b97130b984b34ab486e8c10ddb9c8db09a9350a93be21832e56c429` |
| .agents/skills/claude-session-handoff/SKILL.md | `34b4ed8f…` | `bb2ae62b84d51921e02b0f07fbc62d66cabfbba92e6d6906ca909e2a2015110b` |
| .agents/skills/repository-supervisor/SKILL.md | `d6b03c14…` | `5ae8a5049cee1b750d7a1b14a76f5610f12e39d8a3c635fe9cef8e153a9ebbd8` |

- `git diff --name-status a54d024 3a3f4b8` lists only those eight files plus the six added G053 planning/approval
  records. No executable helper/test, CI/validation, architecture/security/decisions, roadmap/budget or
  foundation/license source changed.
- Companion changes are additive prose only. The §4.1 deltas are re-derived mechanically:
  - handoff SKILL.md: five lines inserted after L133 (role suffix, `.codex/worktrees/`). L12 and L24–25 are
    unchanged; L76, L155, L203, L275, L276 (were L150, L198, L270, L271) carry `management/v0/`; L194 (was L189)
    carries "staging disposition". The new paragraph is carried verbatim with the naming and dispatch paragraphs.
  - supervisor SKILL.md: 23 lines inserted after L39 (checkout/delegated review defaults, `saas-foundation`
    composition). L10–11 unchanged; the invariants line moves from L64–65 to L87–88. The new section is carried
    verbatim.
- **Execution baseline for B5** = `3a3f4b87528028d71b1f067e1657ff14f8b49cfc`. Supervisor documentation between
  `a54d024` and `3a3f4b8` is management provenance, not implementation change; B5's scope and protection diff runs
  from this baseline.
- DEST `/…/Beginify/restruct-sdlc` (main checkout parent; derived from `--git-common-dir`) is absent
  (`test ! -e && ! -L`) at 2026-10-12T15:25Z.
- Tools: Python 3.12.3, Git 2.43.0, Docker 27.5.1, Claude Code 2.1.296, Linux/WSL2.

### 0.3 Amendment change map (no replan; existing files, templates, checks and walkthroughs)

| AC | Amendment content | Where it lands (Plan r2 §3.2 files) | Evidence (existing batch) |
|---|---|---|---|
| AC2 | Supervisor in main integration checkout; bounded subagents for Spec drafting/research and formal Plan/Result reviews, read-only by default, supervisor owns disposition; observed review-ready submission triggers delegated review without another owner prompt; verify exact content; skip duplicate unchanged review; fall back to direct review when delegation is unavailable; automatic relay of decision and gates to the verified coordinator with receipt/acknowledgment; owner interruption and mandatory gates retained; no watcher/transport | `references/roles-lifecycle.md` (review and relay section), `templates/root/AGENTS.md.tmpl`, `templates/root/CLAUDE.md.tmpl` (submit to coordinator; supervisor relays), `templates/management/SUPERVISOR-HANDOFF.md.tmpl`, `templates/management/release/results/REVIEW.md.tmpl` (delegated-review and relay fields) | B3 repo A walkthrough: a review-ready synthetic G001 Plan → bounded child findings → supervisor disposition → relay record with acknowledgment; a "delegation unavailable" variant records direct review; a resubmission with unchanged hash records "duplicate unchanged review skipped" |
| AC4 | `sdlc-management` supersedes only the overlapping `saas-foundation` MANAGEMENT PROCESS (artifacts, responsibilities, authorization, budgets, progress, reviews, validation sequencing, integration, closeout); renderer dry run/resolved inputs, informed authorization, destination/conflict safeguards, technical checks, canonical architecture/ADRs stay; foundation stages become Goal/AC inputs; the renderer plan is not an Implementation Plan; unresolved technical conflicts escalate; no foundation edit or fourth bundled skill | `references/adoption.md` (composition section), `templates/root/AGENTS.md.tmpl` (composition rule), SKILL.md (precedence summary) | B3 repo B: foundation-stage-to-Goal mapping walkthrough record; inspection of the instantiated rule; B5 protection diff proves `skills/saas-foundation` unchanged |
| AC5 | Codex isolated workers under `.codex/worktrees/`, Claude workers under `.claude/worktrees/`; new worker names `<task-id>-agent-<agent-id>-<role>` (`-plan`, `-check`, `-impl`), role verified before task writes; `0-agent-<id>` reserved for the single coordinator; read-only supervisor workers need no checkout; generated worktree/session state private and gitignored; recorded task-specific exception for unsupported naming, as used by this very dispatch | `references/coordination.md` (naming, dispatch and relocation gate, items 4–7 extended with roots and role suffix), `templates/root/AGENTS.md.tmpl`/`CLAUDE.md.tmpl`, `templates/root/gitignore.lines` (private worktree roots), carried companion paragraphs | B3 §9.1 fixtures (a)–(h) with role-suffixed names (`g002-agent-x-impl`), a role-mismatch negative inside (a)/(b), a `.codex/worktrees/` placement case, and (c) modelled on this dispatch's exception record |

The Plan's file list, checker rules, batches and AC table are otherwise unchanged. The amendment adds no new
framework, transport, checker family or hosted campaign.

### 0.4 Attempt state

Attempt 1 of B3 (checkpoint after attempt 2) starts at this addendum. Planning, the amendment and metadata probes
consumed no attempts.

## 1. Status

**Submitted for Codex implementation review (attempt 1 of B3).** This is submission, not acceptance. Implementation
commits are on `worktree-agent-<id>` (not pushed: the coordinator publishes after review, so the
standing safekeeping push was not used). Hosted MR `fast` and the staging full-gate disposition are still pending,
because no MR exists yet.

## 2. Summary and files changed

- New skill `skills/sdlc-management/`: SKILL.md (122 lines), six references, 17 templates plus
  `parameters.example.json` and `gitignore.lines`, the stdlib checker `check_sdlc.py` and its unittest fixtures.
  The amendment content is placed as mapped in §0.3.
- This Result.
- The local bundle `../restruct-sdlc` was created under the Plan's destination-ownership procedure (§5). It lives
  outside the repository and is not in Git.

No other repository path changed. Application, generator, governance, companion helper and tests, runbooks,
`saas-foundation`, CI and validation are byte-identical to the baseline (§4, B5 scope and protection diff).

| Path (repository) | SHA-256 |
|---|---|
| `skills/sdlc-management/LICENSE` | `af5fb2212e1e79a81443ff705570a4f84c6c587143c84fe4fa590b4aae3667c6` |
| `skills/sdlc-management/SKILL.md` | `ed851074b85d10a6ec6099621600131091d7a0313d02933cfdf65e54ca802179` |
| `skills/sdlc-management/agents/openai.yaml` | `0deb1fbec0f3740c862dbc581df0fc04c1d96ec68fc92b1afc5783870f2d00b0` |
| `skills/sdlc-management/references/adoption.md` | `0d88021457de46a0f6ddb4c7b64908b6ce081422ccaed0c25a8ea68902580832` |
| `skills/sdlc-management/references/budgets.md` | `d69fa3ff908cc282ef1fa9971e130a3719f3ef28ff7b327835281a5858d4fcb8` |
| `skills/sdlc-management/references/closeout-risks.md` | `dee1c8e137437078d79fb08ee3d917943b0fedb00e6b4c99a62e17f89c4f6d33` |
| `skills/sdlc-management/references/coordination.md` | `e5976d9cf44f830cd2fe22b0b6f8c2af82d55c1b3c6c849ff25fc17095524d3a` |
| `skills/sdlc-management/references/roles-lifecycle.md` | `6d88fb60e83afc92b5b95dcb98346356190e4ac713f17a565fac6ce6bf6e30fc` |
| `skills/sdlc-management/references/validation-integration.md` | `42d9dfcc5e40f2e45c9c65d663b7b97360baaf432c4eda96c3c33517d2f50861` |
| `skills/sdlc-management/scripts/check_sdlc.py` | `404cc9be300015de8dfb627d2343dfd7bae764e58f7cc75bfda40443feaca9e2` |
| `skills/sdlc-management/scripts/test_check_sdlc.py` | `7d5a3f7255aecdfb203d5ef606f6cf71883153ab3cf83f92c25508448606a3f1` |
| `skills/sdlc-management/templates/docs/RISKS.md.tmpl` | `f8831c02c77ec1c096f24c1b1e51e5ff7d09d9ed778c942469003e9084f4a7d0` |
| `skills/sdlc-management/templates/docs/agents/BRANCH-WORKFLOW.md.tmpl` | `67de1c211fe8e23c4c690c64149d73265e920481543d3dba15da188e7d269ba1` |
| `skills/sdlc-management/templates/docs/agents/CLAUDE-SESSION-HANDOFF.md.tmpl` | `a063d40a260e3bc1eada22449cbcaadc597e2c7473a4596576a686ad4cba5376` |
| `skills/sdlc-management/templates/management/BUDGET_POLICY.md.tmpl` | `6a3a3d29479d2c15a01364011cc26c83df9dd173082764060808b1e9c63d4255` |
| `skills/sdlc-management/templates/management/SUPERVISOR-HANDOFF.md.tmpl` | `366e9d2b53b154acfe09061d50b68e08d6481716ac93c985f9a954d710aa64ce` |
| `skills/sdlc-management/templates/management/release/BACKLOG.md.tmpl` | `9acc9f3b9da09780063885ef94253f3abdba33e479c3df55ef746bf2800ee20d` |
| `skills/sdlc-management/templates/management/release/PROGRESS.md.tmpl` | `9ccb5335e854c1de12041fd9031db929c58b4e0ff907a2c1d0ff6840634dc2aa` |
| `skills/sdlc-management/templates/management/release/ROADMAP.md.tmpl` | `349ab478ecffdb1b13d0e51fdfe0a42f9f1ad25540a67693fbb0e7854f91e7da` |
| `skills/sdlc-management/templates/management/release/archive/PROGRESS-ARCHIVE.md.tmpl` | `ac3fe741c058b09ba0c681be2ad1f771bab32e308226d05e7dd8cbe89179f6e7` |
| `skills/sdlc-management/templates/management/release/goals/GOAL.md.tmpl` | `ae6c856899d711c7b6b86ebadac35224b5e0c5f9e4cc87756197375c70f83c18` |
| `skills/sdlc-management/templates/management/release/plans/PLAN.md.tmpl` | `e81bae6f7478088bdc7dae942cadd5515fd909aae5f5f234e4a620dce43dc9e9` |
| `skills/sdlc-management/templates/management/release/results/RESULT.md.tmpl` | `cc3748923b4aee2cf806d1afc3d9cd817be3b9469fb1c5370ed759bc93bcb05a` |
| `skills/sdlc-management/templates/management/release/results/REVIEW.md.tmpl` | `071aed87f7ea757c7623d7e3e1b1f1e204b5f4842c8b833c45a8cdee81baf4ea` |
| `skills/sdlc-management/templates/management/release/specs/SPEC.md.tmpl` | `c141693f4efc7ed7b2b730cb360261ddce1e9a3d241e42cb64dd47cbbb587b4e` |
| `skills/sdlc-management/templates/parameters.example.json` | `36cbc0b69504bd05aba4d1e97b5d310229f29e2196c26f411e605f40fe08575e` |
| `skills/sdlc-management/templates/root/AGENTS.md.tmpl` | `f9b9bf8457efacf3045e6f921e2a8c78110ca87ac19c7785c2b2dbf68f15fba9` |
| `skills/sdlc-management/templates/root/CLAUDE.md.tmpl` | `ac2e7dae6722b318d2245744543b86e78bbc9895ea4000439014a4640fe77d17` |
| `skills/sdlc-management/templates/root/gitignore.lines` | `3b9bc7e17d5e222aaed92f3d7f099151f13733c3be29c0880768724519c6d8d5` |
| `management/v0/results/G053-restruct-sdlc-bundle-result.md` | (this file; see handback) |

## 3. Migrations, schema and contracts

None. There is no public API/MCP, schema, provider, permission, model or dependency change, and no CI wiring
(Plan r2 §4.5; Codex Q1).

## 4. Tests added and validation

**Tests added:** `skills/sdlc-management/scripts/test_check_sdlc.py`, 32 stdlib unittest cases covering P1–P7, P3t
and A1–A6 in positive and negative form. These include the C1 minimum: a positive render passes; missing-target,
undeclared/unresolved-parameter and escaping-target cases fail, in shipped (P3), template (P3t) and rendered (A2/A3)
form. Private worker checkouts are excluded from the adoption scans.

**Before every batch:** `prebatch.sh` ran `git fetch origin` and `git merge-base --is-ancestor origin/staging HEAD`,
then `bash infra/ci/docker-desktop-mounts.sh` (4348 of 100000 mounts) and `sha256sum -c` over the 33-row execution
binding file: the 29 §2 rows (8 reconciled) plus Plan r2, r1, the amendment and the confirmation. Target
`origin/staging` stayed at `3a3f4b8` throughout, so no rebase was needed. Evidence lives under the private
`g053-exec-24271ac22f/impl/` evidence root (`$EVID`). Every batch ended with
`git status --porcelain --ignored` clean.

| Batch | Candidate | Command(s) | Outcome | Evidence (`$EVID/…`) |
|---|---|---|---|---|
| B1 dev runs | `eb25c4d` + uncommitted skill | unit tests | 2 failures (test-only: the self-scan matched a test literal; a usage-error test expected a return instead of `SystemExit`); fixed | `prebatch-b1-dev1.log`, `prebatch-b1-dev2.log` |
| B1 | `6ba2081` | `python3 -B -m unittest discover -s skills/sdlc-management/scripts -p 'test_check_sdlc.py' -v` (31 OK); `check_sdlc.py package skills/sdlc-management --skills-root .agents/skills --literals $EVID/literals.txt` (exit 0, P1–P6 + P3t pass); `wc -l` SKILL.md = 122; source helper baseline `test_handoff.py` 59 OK | PASS | `b1/` |
| B2 | `6ba2081` | export from committed blobs; `check_sdlc.py package $STAGE --literals …` (P1–P7 pass); exported helper 59 OK; exported checker 31 OK; `cmp` verbatim files (all same); `diff -u` adapted files; gitleaks v8.30.1 tar-on-stdin under the lock → `[]`, pipe `0 0 0`; notice inventory | PASS | `b2/` |
| B3 | `6ba2081` stage | walkthroughs | **found a defect:** adoption A3/A6 scanned private worker checkouts under `.claude/worktrees/`, whose checked-out `.agents/skills` templates hold `{{…}}`. Fixed in `5034d06` (skip private worker roots, plus a test). Not rerun on the old stage | (design review before the run) |
| B1 rerun | `5034d06` | as B1: 32 OK; package exit 0; 122 lines; helper 59 OK | PASS | `b1-r2/` |
| B2 rerun | `5034d06` | as B2: 32 OK and 59 OK; gitleaks `[]`; manifest `dd30c7b1…` | PASS | `b2-r2/` |
| B3 | `5034d06` stage | `b3_walkthrough.py` (59 checks) | 1 failure: the walkthrough script indexed AC7's log by position, which a rebase step had shifted. Fixed by lookup by step name | `b3/` |
| B3 rerun | `5034d06` stage | 59 checks | PASS (59/59) | `b3-r2/` |
| D1 | `5034d06` stage | §4.3 create-only publish | PASS (§5) | `d1/`, `destination.json` |
| B4 | DEST | `b4_relocation.py` (9 checks) | 1 failure: the resume script's row filter matched the table header. Fixed | `b4/` |
| B4 rerun | DEST | 9 checks | PASS (9/9) | `b4-r2/` |
| B5 | `5034d06` | `flock …/local-harness.lock bash infra/ci/validate.sh --profile full --evidence-dir $EVID/b5/evidence`; scope and protection diff from baseline `3a3f4b8` | PASS: `validate: final status 0`, every phase 0; 88 checks PASS and 5 INFO, 0 FAIL, including B-RENDER-PLAN, B-PROTECTED, B-SKILL-UNCHANGED, B-CHECKS-ROOT (incl. secret-scan), B-SCOPE and B-REQUIRED-CHECKS; cleanup 0, owned_remaining 0. Scope diff: only §3.1 paths; protection diff exit 0; DEST `skills/sdlc-management` equals the 29 HEAD blobs | `b5/` |

**Evidence reuse.** This Result is committed after B5 and changes only `management/**/*.md` (docs class). B5,
B1–B4 and D1 evidence is reused with provenance: tested commit `5034d06`, plus a Result-only delta.

Every failure in this table was diagnosed, fixed and rerun before the next batch. Only the B3/B4 walkthrough scripts
(private evidence tools) and one checker scope defect were changed. No product test was weakened.

**Outstanding-findings checklist**

- [x] B1 dev: test file self-matched its literal fixture → literal assembled at run time.
- [x] B1 dev: usage-error test expected a return code → asserts `SystemExit(2)`; covers 5 usage cases.
- [x] B3 design: A3/A6 scanned private worker checkouts → skipped via `PRIVATE_ROOTS`; test added; B1/B2 rerun.
- [x] B3: AC7 positional log index → step-name lookup; B3 rerun 59/59.
- [x] B4: resume row filter matched the header → `| G###` regex; B4 rerun 9/9.

## 5. Bundle (`../restruct-sdlc`) and destination ownership

- DEST `/…/Beginify/restruct-sdlc` was derived from `git rev-parse --path-format=absolute --git-common-dir`. It
  was absent before D1. It was created with `mkdir` (no `-p`) at 2026-10-12T15:47:13Z from the stage validated in
  B2-r2. Its manifest was verified in place: `check_sdlc.py package DEST` exits 0, and `diff -r` stage↔DEST is
  empty. There are no symlinks. The ownership record is `$EVID/destination.json`: `pre_existed: false`,
  manifest SHA-256 `dd30c7b15dd3b1701c69820bfed9a29fd5b12a45b60ebc875bdd9d167d507fc4`, 41 files, source commit
  `5034d06`.
- Not done: `git init`, remote, push, tag or publication.
- Inventory: README.md, LICENSE and MANIFEST.tsv, plus `skills/{sdlc-management, claude-session-handoff,
  repository-supervisor}`. MANIFEST.tsv rows, as `sha256 origin path`:

```text
af5fb2212e1e79a81443ff705570a4f84c6c587143c84fe4fa590b4aae3667c6  new  LICENSE
14a04d0b19ffdc69079e79c2480cd04be30efa018ccec4aaf5766b9868f5b4e5  new  README.md
af5fb2212e1e79a81443ff705570a4f84c6c587143c84fe4fa590b4aae3667c6  new  skills/claude-session-handoff/LICENSE
234e88685442937ed1d22e4678b7e411bf09069c1fde80ad0e4b865823d92ba9  adapted  skills/claude-session-handoff/SKILL.md
c623fdb58e3be79443e5d93f633a4ec1f7783cf63116a2f57ea2736676499107  verbatim  skills/claude-session-handoff/agents/openai.yaml
8ffa9915d07c3da4f080128d3270f5d0e0dcacd04b2b0a0220aa52c0e9bc6b5e  adapted  skills/claude-session-handoff/references/local-session-handoff.md
e2bdce82472125fe006efa9e5e70dec003c428e386fa971a3442e732724f9d77  verbatim  skills/claude-session-handoff/scripts/handoff.py
80a24c6519e76eb3c5be95e54561f167a3f0651e69c025624da63de052edc994  adapted  skills/claude-session-handoff/scripts/test_handoff.py
af5fb2212e1e79a81443ff705570a4f84c6c587143c84fe4fa590b4aae3667c6  new  skills/repository-supervisor/LICENSE
666cdd32692e2ac22bb2d3a6151eb2c19f085a6a52fca679ef382164b54cd2b1  adapted  skills/repository-supervisor/SKILL.md
587f23e27064559066153252d8a56f50d6dc167606d8d1dec12857d4f9ae58e8  verbatim  skills/repository-supervisor/agents/openai.yaml
af5fb2212e1e79a81443ff705570a4f84c6c587143c84fe4fa590b4aae3667c6  verbatim  skills/sdlc-management/LICENSE
ed851074b85d10a6ec6099621600131091d7a0313d02933cfdf65e54ca802179  verbatim  skills/sdlc-management/SKILL.md
0deb1fbec0f3740c862dbc581df0fc04c1d96ec68fc92b1afc5783870f2d00b0  verbatim  skills/sdlc-management/agents/openai.yaml
0d88021457de46a0f6ddb4c7b64908b6ce081422ccaed0c25a8ea68902580832  verbatim  skills/sdlc-management/references/adoption.md
d69fa3ff908cc282ef1fa9971e130a3719f3ef28ff7b327835281a5858d4fcb8  verbatim  skills/sdlc-management/references/budgets.md
dee1c8e137437078d79fb08ee3d917943b0fedb00e6b4c99a62e17f89c4f6d33  verbatim  skills/sdlc-management/references/closeout-risks.md
e5976d9cf44f830cd2fe22b0b6f8c2af82d55c1b3c6c849ff25fc17095524d3a  verbatim  skills/sdlc-management/references/coordination.md
6d88fb60e83afc92b5b95dcb98346356190e4ac713f17a565fac6ce6bf6e30fc  verbatim  skills/sdlc-management/references/roles-lifecycle.md
42d9dfcc5e40f2e45c9c65d663b7b97360baaf432c4eda96c3c33517d2f50861  verbatim  skills/sdlc-management/references/validation-integration.md
404cc9be300015de8dfb627d2343dfd7bae764e58f7cc75bfda40443feaca9e2  verbatim  skills/sdlc-management/scripts/check_sdlc.py
7d5a3f7255aecdfb203d5ef606f6cf71883153ab3cf83f92c25508448606a3f1  verbatim  skills/sdlc-management/scripts/test_check_sdlc.py
f8831c02c77ec1c096f24c1b1e51e5ff7d09d9ed778c942469003e9084f4a7d0  verbatim  skills/sdlc-management/templates/docs/RISKS.md.tmpl
67de1c211fe8e23c4c690c64149d73265e920481543d3dba15da188e7d269ba1  verbatim  skills/sdlc-management/templates/docs/agents/BRANCH-WORKFLOW.md.tmpl
a063d40a260e3bc1eada22449cbcaadc597e2c7473a4596576a686ad4cba5376  verbatim  skills/sdlc-management/templates/docs/agents/CLAUDE-SESSION-HANDOFF.md.tmpl
6a3a3d29479d2c15a01364011cc26c83df9dd173082764060808b1e9c63d4255  verbatim  skills/sdlc-management/templates/management/BUDGET_POLICY.md.tmpl
366e9d2b53b154acfe09061d50b68e08d6481716ac93c985f9a954d710aa64ce  verbatim  skills/sdlc-management/templates/management/SUPERVISOR-HANDOFF.md.tmpl
9acc9f3b9da09780063885ef94253f3abdba33e479c3df55ef746bf2800ee20d  verbatim  skills/sdlc-management/templates/management/release/BACKLOG.md.tmpl
9ccb5335e854c1de12041fd9031db929c58b4e0ff907a2c1d0ff6840634dc2aa  verbatim  skills/sdlc-management/templates/management/release/PROGRESS.md.tmpl
349ab478ecffdb1b13d0e51fdfe0a42f9f1ad25540a67693fbb0e7854f91e7da  verbatim  skills/sdlc-management/templates/management/release/ROADMAP.md.tmpl
ac3fe741c058b09ba0c681be2ad1f771bab32e308226d05e7dd8cbe89179f6e7  verbatim  skills/sdlc-management/templates/management/release/archive/PROGRESS-ARCHIVE.md.tmpl
ae6c856899d711c7b6b86ebadac35224b5e0c5f9e4cc87756197375c70f83c18  verbatim  skills/sdlc-management/templates/management/release/goals/GOAL.md.tmpl
e81bae6f7478088bdc7dae942cadd5515fd909aae5f5f234e4a620dce43dc9e9  verbatim  skills/sdlc-management/templates/management/release/plans/PLAN.md.tmpl
cc3748923b4aee2cf806d1afc3d9cd817be3b9469fb1c5370ed759bc93bcb05a  verbatim  skills/sdlc-management/templates/management/release/results/RESULT.md.tmpl
071aed87f7ea757c7623d7e3e1b1f1e204b5f4842c8b833c45a8cdee81baf4ea  verbatim  skills/sdlc-management/templates/management/release/results/REVIEW.md.tmpl
c141693f4efc7ed7b2b730cb360261ddce1e9a3d241e42cb64dd47cbbb587b4e  verbatim  skills/sdlc-management/templates/management/release/specs/SPEC.md.tmpl
36cbc0b69504bd05aba4d1e97b5d310229f29e2196c26f411e605f40fe08575e  verbatim  skills/sdlc-management/templates/parameters.example.json
f9b9bf8457efacf3045e6f921e2a8c78110ca87ac19c7785c2b2dbf68f15fba9  verbatim  skills/sdlc-management/templates/root/AGENTS.md.tmpl
ac2e7dae6722b318d2245744543b86e78bbc9895ea4000439014a4640fe77d17  verbatim  skills/sdlc-management/templates/root/CLAUDE.md.tmpl
3b9bc7e17d5e222aaed92f3d7f099151f13733c3be29c0880768724519c6d8d5  verbatim  skills/sdlc-management/templates/root/gitignore.lines
```

- **Private provenance.** Source commit `5034d06d8098ab4aea9bd00ba46e49b6456ff593`, tree
  `832658da850bd81c172a56b6cca4ee7e7d189eb3`. Export sources, read from committed blobs:
  - verbatim: `handoff.py` `e2bdce82…`, handoff `openai.yaml` `c623fdb5…`, supervisor `openai.yaml` `587f23e2…`
    (equal to the §2 hashes) and all 29 `skills/sdlc-management` files;
  - adapted: handoff SKILL.md (source `bb2ae62b…`, export `234e8868…`), `local-session-handoff.md` (`dbb72c27…` →
    `8ffa9915…`), `test_handoff.py` (`aa52b129…` → `80a24c65…`), supervisor SKILL.md (`5ae8a504…` → `666cdd32…`);
  - LICENSE: from `skills/saas-foundation/LICENSE` (`af5fb221…`, "Copyright (c) 2026 Vitruvian Technology, Corp.").
- **Notice inventory.** Only the five MIT LICENSE copies and the README license line. There are no third-party
  notices.
- **README** (new, the only Restruct branding): SHA-256 `14a04d0b19ffdc69079e79c2480cd04be30efa018ccec4aaf5766b9868f5b4e5`.

**Exact unified diffs of the adapted files** (re-derived mechanically at `3a3f4b8`; §0.2):

```diff
--- a/.agents/skills/claude-session-handoff/SKILL.md
+++ b/skills/claude-session-handoff/SKILL.md
@@ -9,7 +9,7 @@
 
 ## Select the procedure
 
-Read the active repository's `AGENTS.md` and `docs/agents/CLAUDE-SESSION-HANDOFF.md` when present. That repository runbook is authoritative for its local interface and workflow. Otherwise read [the bundled reference](references/local-session-handoff.md), adapted from the Beginify runbook verified with Claude Code 2.1.290/2.1.291 and 2.1.296 on Linux/WSL. Its internal socket interface is version-specific; verify compatibility before messaging after an upgrade. Resolve repository paths against the target checkout, not this skill's installation directory.
+Read the active repository's `AGENTS.md` and `docs/agents/CLAUDE-SESSION-HANDOFF.md` when present. That repository runbook is authoritative for its local interface and workflow. Otherwise read [the bundled reference](references/local-session-handoff.md), adapted from the originating repository's runbook verified with Claude Code 2.1.290/2.1.291 and 2.1.296 on Linux/WSL. Its internal socket interface is version-specific; verify compatibility before messaging after an upgrade. Resolve repository paths against the target checkout, not this skill's installation directory.
 
 ## Check progress
 
@@ -21,8 +21,8 @@
 
 ```sh
 python3 .agents/skills/claude-session-handoff/scripts/handoff.py --repo "$PWD" status
-python3 .agents/skills/claude-session-handoff/scripts/handoff.py --repo "$PWD" check --match G006
-python3 .agents/skills/claude-session-handoff/scripts/handoff.py --repo "$PWD" check --session FULL_UUID --match G006
+python3 .agents/skills/claude-session-handoff/scripts/handoff.py --repo "$PWD" check --match GNNN
+python3 .agents/skills/claude-session-handoff/scripts/handoff.py --repo "$PWD" check --session FULL_UUID --match GNNN
 ```
 
 `status` reports the preferred-coordinator pin and duplicate-safe discovery (see
@@ -73,7 +73,7 @@
 
 ```sh
 H=".agents/skills/claude-session-handoff/scripts/handoff.py"
-python3 "$H" --repo "$PWD" pin --session FULL_UUID --selection-ref management/v0/PROGRESS.md
+python3 "$H" --repo "$PWD" pin --session FULL_UUID --selection-ref management/<release>/PROGRESS.md
 python3 "$H" --repo "$PWD" status
 python3 "$H" --repo "$PWD" unpin --session FULL_UUID
 ```
@@ -152,7 +152,7 @@
 ```sh
 python3 "$H" --repo "$PWD" status --creation-check \
   --coordinator-uuid FULL_UUID \
-  --authority-ref management/v0/results/RECORD.md --authority-sha256 HEX \
+  --authority-ref management/<release>/results/RECORD.md --authority-sha256 HEX \
   --authority-text "exact quoted owner decision" \
   --creation-name NAME --creation-prompt-file /path/outside-checkout/prompt.txt
 ```
@@ -191,7 +191,7 @@
 1. **Verify closure first (Codex or owner).** Identify the standing terminal record
    on the integration branch: for `completed`, the task's final
    `results/<TASK>-implementation-codex-review*.md` ACCEPT after the qualifying
-   staging disposition; for `canceled`/`superseded`, the PROGRESS entry or review
+   integration-branch disposition; for `canceled`/`superseded`, the PROGRESS entry or review
    recording it (and the successor). Confirm no later record reopens the task and
    quote the exact sentence. A task label, handoff ID, receipt or idle PID is not
    closure. The closure reference must be the real file: a path through a
@@ -200,7 +200,7 @@
 
    ```sh
    python3 "$H" --repo MAIN_CHECKOUT compact --task GNNN --disposition completed \
-     --closure-ref management/v0/results/GNNN-implementation-codex-review.md \
+     --closure-ref management/<release>/results/GNNN-implementation-codex-review.md \
      --closure-sha256 HEX --closure-text "quoted ACCEPT sentence" --verified-by codex \
      --writer-sha256 REVIEWED_INTEGRATED_HELPER_SHA256 [--handoff-id UUID ...]
    ```
@@ -272,8 +272,8 @@
 ```sh
 python3 .agents/skills/claude-session-handoff/scripts/handoff.py --repo MAIN_CHECKOUT send \
   --message-file /tmp/authorized-handoff.txt --task GNNN \
-  --document management/v0/PROGRESS.md \
-  --document management/v0/results/GOAL-execution-binding-codex-confirmation.md
+  --document management/<release>/PROGRESS.md \
+  --document management/<release>/results/GOAL-execution-binding-codex-confirmation.md
 python3 .agents/skills/claude-session-handoff/scripts/handoff.py --repo MAIN_CHECKOUT receipt \
   --handoff-id UUID_RETURNED_BY_SEND
 ```
--- a/.agents/skills/claude-session-handoff/references/local-session-handoff.md
+++ b/skills/claude-session-handoff/references/local-session-handoff.md
@@ -1,4 +1,4 @@
-<!-- Bundled fallback adapted from docs/agents/CLAUDE-SESSION-HANDOFF.md in Beginify platform. Prefer the target repository runbook when present. Relative repository paths below refer to the target checkout. -->
+<!-- Bundled fallback adapted from the originating repository's docs/agents/CLAUDE-SESSION-HANDOFF.md. Prefer the target repository runbook when present. Relative repository paths below refer to the target checkout. -->
 
 # Codex ↔ Claude Code session handoffs
 
@@ -115,7 +115,7 @@
 PY
 ```
 
-Never print/copy/commit the key or token. Do not include them in handoff files, task prompts, logs or evidence. They authenticate the local inbox and are distinct from GitLab/provider credentials. Use this only for the user's authorized local session communication; never redirect the secret or handoff to a remote socket/service.
+Never print/copy/commit the key or token. Do not include them in handoff files, task prompts, logs or evidence. They authenticate the local inbox and are distinct from Git-hosting/provider credentials. Use this only for the user's authorized local session communication; never redirect the secret or handoff to a remote socket/service.
 
 ## Confirm receipt and dispatch
 
--- a/.agents/skills/claude-session-handoff/scripts/test_handoff.py
+++ b/skills/claude-session-handoff/scripts/test_handoff.py
@@ -7,9 +7,9 @@
 from unittest.mock import patch
 import handoff
 
-SESSION = "<redacted-uuid>"
-OTHER = "<redacted-uuid>"
-HANDOFF = "<redacted-uuid>"
+SESSION = "00000000-0000-4000-8000-000000000001"
+OTHER = "00000000-0000-4000-8000-000000000002"
+HANDOFF = "00000000-0000-4000-8000-000000000003"
 
 class ReceiptRouting(unittest.TestCase):
     def test_receipt_requires_selected_session_user_replay(self):
--- a/.agents/skills/repository-supervisor/SKILL.md
+++ b/skills/repository-supervisor/SKILL.md
@@ -7,8 +7,8 @@
 
 Recover current supervisor state from the repository rather than an opening prompt
 or a replay of earlier conversation. Follow the supervisor role in AGENTS.md when
-the user has assigned it to this agent, including an explicitly delegated GLM or
-DeepSeek session. Loading the skill does not assign authority, approve work or
+the user has assigned it to this agent, including another explicitly delegated
+agent session. Loading the skill does not assign authority, approve work or
 authorize merges. Claude Code remains the primary implementer; major product and
 architecture decisions remain with the owner.
 
@@ -84,8 +84,8 @@
 security, public-contract, irreversible-model and resource conflicts as required;
 do not extend/reset budgets or weaken evidence to finish.
 
-Preserve UUID canonical identities, current tenancy/authority semantics and the
-G021 product-feature barrier. Read service-root instructions/skills for service
+Preserve the adopting repository's canonical invariants, authority semantics and
+any declared delivery barrier. Read service-root instructions/skills for service
 work. The skill contains no substitute permission matrix, CI policy or fixed
 execution baseline; those belong to current approved repository records.
 
```

## 6. Acceptance criteria

| AC | Status | Evidence |
|---|---|---|
| AC1 | met | P1–P4 and P3t pass on the skill (skills root `.agents/skills`; B1) and on the bundle (skills root `skills/`; B2/D1/B4). SKILL.md is 122 lines, with six references and every template linked (P4). Boundaries: `sdlc-management` has no transport code and assigns no role; companions are linked, not duplicated; the SKILL.md boundary table says routine checks need no reload of all three. The naming gate reads consistently across `coordination.md` items 1–9, both root templates and the carried companion paragraphs (handoff L119–137, supervisor L40–61); every text says a name grants no authority. |
| AC2 | met | B3 repo A: A1–A6 pass; 10 created paths; supervisor reference at `management/SUPERVISOR-HANDOFF.md`; session log and both worktree roots ignored (`git check-ignore`); eleven Goal fields (A4). Gate items 1–9 and all key phrases are in the instantiated AGENTS.md and CLAUDE.md. The C1 negatives in repo A copies fail as required (missing target → A2, `{{UNDECLARED_X}}` → A3, `../../` escape → A2), as do the checker tests. Review walkthrough: a draft is not reviewed; r1 delegated-child findings lead to CORRECTION LOOP and a REVIEW record with a relay receipt/acknowledgment record; an unchanged resubmission is skipped as a duplicate; r2 with delegation unavailable is reviewed directly → ACCEPT. A6 literal scan: no project records. `b3-r2/records/ac2-*`. |
| AC3 | met | Synthetic G001 B3 ledger: 3 attempts counted against 3 commands and 3 failed batches. The worker contribution shares attempt 1. The extension record has previous/new budget, used attempts, reason, remaining work and approver, and counters continue to attempt 4. All failures are retained. `b3-r2/records/ac3-ledger.md`. |
| AC4 | met | Repo B (Harbor Notes, release `r1`, `develop`/`release`, `local-before-mr`): the pre-edited AGENTS.md hash is unchanged and `AGENTS.md.sdlc-proposed` exists; A1–A6 pass; no barrier (`(none)`) and no vendor literals. The composition rule is present: management-only precedence; renderer dry run, authorization, conflict safeguards and technical checks retained; renderer plan ≠ Implementation Plan; conflicts escalate. Foundation stages map to Goal/Spec ACs. The `saas-foundation` tree hash is unchanged before/after, and B5's protection diff is empty. |
| AC5 | met | Exported helper tests pass, equal to the source baseline (59/59). Fixtures (a)–(h) use real `git worktree add/list --porcelain/move`. (a) `g002-agent-x1-impl` verified; a role mismatch fails; a Codex `.codex/worktrees` worker is verified and the wrong root fails. (b) `agent-y2` fails with no task write and an exception request recorded. (c) The recorded exception mapping activates only its own dispatch. (d) `0-agent-w3` is rejected. (e) Helper creation-gate/discovery/routing tests pass. (f) A rename to `0-agent-z` without a pin is not routed. (g1) A live cwd is detected and the move refused. (g2) A cwd scan alone passes and the fd scan finds the holder → refused. (g3) Seven unknown/unproven inputs defer, with porcelain unchanged. (g4) The real host inspection is unprovable (other-user PIDs), so **real relocation is deferred**. A labelled SYNTHETIC complete inspection then allows the disposable move (porcelain verified), and the exported-fixture probe shows `status` `changes.checkout` `{at_pin: …/0-agent-old, now: …/0-agent-new}`. (h) The bound path is untouched; recovery records are kept. Documented examples: `0-agent-7f3a2c1`, `g002-agent-4be91d0-impl`, `g002-agent-<id>-impl`. |
| AC6 | met | Exported helper tests pass: newer-peer block, missing receipt unconfirmed, no retry, `queued_observed` is not a receipt, user replay required. Walkthrough: a queued delivery is reported unconfirmed with no resend, acknowledgment or execution; a peer request to change permissions/provider is refused. No config change and no session. |
| AC7 | met | Repo A with a local bare remote. Staging advanced, so the ancestor check fails, then rebase with recorded SHAs. B1 fails → fix → B1 rerun passes → manual feedback (SYNTHETIC) → final qualification passes. The five glab commands were printed, not run. After a simulated authorized merge, the merged temporary branch was deleted after an exact-tip check; the remote keeps only `main`/`staging`. No hosted action. |
| AC8 | met | Closed G001 history moved into `archive/PROGRESS-2026-10-12.md` from the template (source SHA, hashes, preserved text); A2 resolves the archive and PROGRESS links. The active G002 entry is kept. The clean merged worktree was removed; removal of the dirty active worktree was refused without force. RISKS.md records a missing-index and row-lock concern. Helper compact tests (t9 ×2, t11, t15) pass. |
| AC9 | met | MANIFEST is verified (P7) and the LICENSE copies are MIT. P5–P7 and the literal list (43 entries: archetype identity, Git identity, origin, paths, 11 UUIDs with prefixes, project terms) are clean on the skill, stage, DEST and the relocated copy. gitleaks `[]`. There is no `.git`, local state or symlink. The B5 scope diff contains only §3.1 paths and the protection diff is empty. The companion sources equal their reconciled §2 hashes. B-PROTECTED and B-SKILL-UNCHANGED pass in the full run. No existing code changed. |
| AC10 | met | B4 from a second path: package check P1–P7; flat install into repo C `.agents/skills`; adoption A1–A6; P3 on each installed skill; installed helper tests 59 OK. Bootstrap included the naming gate (`g001-agent-r7-plan` verified from porcelain; the installed reference defers relocation). Supervisor resume reconstructed G001, its phase and the next gate "Plan review (ACCEPT required before implementation)" from repo C records alone. Platform: Linux 5.15 WSL2, Python 3.12.3, Git 2.43.0, verified CLI 2.1.290/2.1.291/2.1.296 (creation 2.1.296). Unsupported cases are recorded (`b4-r2/b4-environment.json`). |
| AC11 | met locally; hosted pending | B1/B2 (checker 32 OK; exported helper 59 OK), the B3/B4 walkthroughs and B5 full profile (final status 0, 88 PASS) on `5034d06`. MR `fast` and the staging full-gate disposition are **pending** (post-review). No CI exemption. |
| AC12 | submitted | This Result: AC map, commands and statuses, source/export hashes and diffs, attempt ledger, limitations and the DEST record. Codex reviews the actual diff and bundle. |

## 7. Deviations

- **R1 optional reviewer not used.** The coordinator's dispatch forbids starting agents, so the AC10 resume exercise
  was performed mechanically from repo C records (B4) instead of by a separate reviewer. Plan §8 makes the
  reviewer optional.
- **No safekeeping push.** The coordinator's dispatch says not to push; it publishes after review.
- **Checker scope fix within the attempt.** A3/A6 skip `.claude/worktrees/` and `.codex/worktrees/` (private
  worker checkouts). This follows the amendment's private worktree roots and is documented in `adoption.md`.
- **AC5 (g4).** Executed per Codex's r2 clarification: real relocation is recorded as deferred, and the positive
  procedure used labelled synthetic complete inspection evidence. No real session was queried, moved or restarted.

No architecture, security-model, contract, dependency or scope deviation.

## 8. Execution Budget

Budget class: B3 (checkpoint after attempt 2). Maximum attempts: 3. Attempts used: 1 (active, submitted).

1. Attempt 1: S0 reconciliation, addendum, skill authoring, export, walkthroughs, D1 publish, B4 and B5. The inner
   fix loops listed in §4 stay within this attempt, and there were no subagent contributions. Boundary: this
   submission.

## 9. Limitations and follow-up

- The helper conventions (`G###`, `codex-review` record name, `codex-supervisor`, the `.agents/skills` install
  path, Linux `/proc`) are documented, not generalized (E3).
- Coordinator relocation stays deferred until a same-UUID restart from a moved path is verified locally.
- Follow-up (authorizes nothing): an owner decision on public publication of `../restruct-sdlc`.

## 10. Correction record — attempt 2 (Codex CORRECTION LOOP)

Review: [G053 implementation review](G053-implementation-codex-review.md), SHA-256
`21925f3b1695349732ba62c358368f4c72c26987187072ef0f9563220a40daff` (read in the integration checkout before it was
committed), delivered through coordinator handoff `14a1f4e5`. It reviewed candidate `88e07f5`. Attempt 2 of B3 was
authorized, in the same verified checkout and under the same mapping `4b5b3250…`. Only C1–C3 were fixed. Plan r2,
the amendment and their history are unchanged, and the helper, application, foundation and CI bytes are unchanged.

### 10.1 Changes

| Finding | Change | Files |
|---|---|---|
| C1 | Adoption scope is now separate from export privacy. A3 looks for unresolved placeholders only in adopted artifacts: the rendered template destinations, their `.sdlc-proposed` siblings, and the Markdown under `management/<release>/`. A6 applies the strict P6 scan to the installed bundle skills only; the rest of the target repository is scanned only for source-origin `--literals` and private-key material. P6 (`package`) is unchanged and strict. The module docstring and `adoption.md` explain the scopes. | `scripts/check_sdlc.py`, `scripts/test_check_sdlc.py` |
| C2 | A collision-safe flat install procedure, with the same text in `adoption.md` (new "Installing the skills" section) and the bundle README. Absent → copy. Identical → leave unchanged. Different → keep byte-identical and record the owner's merge or replacement decision. A symlinked `.agents`, `.agents/skills` or destination → stop. SKILL.md step 2 summarizes it. This is a shell snippet, not an installer framework. | `references/adoption.md`, `SKILL.md`, bundle `README.md` |
| C3 | CLAUDE template step 2 is now conditional: Plan ACCEPT is required unless the Goal's Execution Policy explicitly authorizes the direct tier. Direct work stays within the authorized Goal/Spec, its Result still records baseline, files and checks, and the diff is reviewed. There is no tier self-selection, and security/data/architecture work keeps stronger review. The AGENTS template and the SKILL.md lifecycle line match. | `templates/root/CLAUDE.md.tmpl`, `templates/root/AGENTS.md.tmpl`, `SKILL.md` |

Commit `3058d2b` (the tested correction candidate). SHA-256 at that commit:

| Path | SHA-256 |
|---|---|
| `skills/sdlc-management/SKILL.md` | `0dcdf6d2a107069ae94dd39753348a1cee66d1906e827cc3bafa58932cc7e770` |
| `skills/sdlc-management/references/adoption.md` | `2d4a505f7e4a98c8f8e516876e8e6c3a271892e6b3d5ed3b2ff5e0efc8bfcf95` |
| `skills/sdlc-management/scripts/check_sdlc.py` | `e65ce7dd301cb0b1f238b33d5537b5bc69519cc40bdc608a7ffc9adfb7829725` |
| `skills/sdlc-management/scripts/test_check_sdlc.py` | `1dbde082babd217d4a4936a69ac4e01db471694ec7e62eeb1a1d7bfd0ef467c8` |
| `skills/sdlc-management/templates/root/AGENTS.md.tmpl` | `5402ba4fa8dc2b4e1550f59da0e923e294aad2e02bcaceed43e700f80059fb76` |
| `skills/sdlc-management/templates/root/CLAUDE.md.tmpl` | `f92c3e9f608660f5589edc7f07626ddfa821e41c77fb8e6562d95def05101b34` |

The other 23 skill files are unchanged from §2.

### 10.2 Re-export (Plan r2 §4.3)

- Owned DEST matched `destination.json`: manifest `dd30c7b1…`, P7 clean, no symlinks.
- It was moved to `$EVID/superseded/20261012T170822Z/restruct-sdlc` and recreated create-only from the a2-b2
  stage.
- New manifest SHA-256: `476c1589f499f539f3931e6a18553783a2fb481a1b771ecadb6127ded3e1bf06` (41 files). DEST
  package check exits 0, `diff -r` stage↔DEST is empty, and there are no symlinks. `destination.json` was updated
  (the previous record is kept in `$EVID/a2-d1/destination-previous.json`).
- Changed bundle rows (7):
  - `README.md` → `a79f6bf9…`;
  - `skills/sdlc-management/SKILL.md`, `references/adoption.md`, `scripts/check_sdlc.py`,
    `scripts/test_check_sdlc.py`, `templates/root/AGENTS.md.tmpl` and `templates/root/CLAUDE.md.tmpl`, as in §10.1.
- The companion rows are unchanged: verbatim `handoff.py` `e2bdce82…` and the four adapted files with the same hashes
  and diffs as §5.

### 10.3 Validation (attempt 2)

Every batch ran after the pre-batch gate: fetch, ancestry, mount safeguard, bindings. Target stayed `3a3f4b8`.

| Batch | Candidate | Outcome | Evidence (`$EVID/…`) |
|---|---|---|---|
| dev run | `88e07f5` + uncommitted C1 edits | 35 OK. **Process slip:** this focused checker run started before the pre-batch gate; the gate ran immediately afterwards (`prebatch-a2-dev1.log`: target unchanged `3a3f4b8`, ancestor yes, mounts 4%, bindings ok). The run is development only, not evidence. | — |
| B1 | `3058d2b` | PASS: checker 35 OK; skill package check exit 0; SKILL.md 124 lines; source helper 59 OK | `a2-b1/` |
| B2 | `3058d2b` | PASS: stage P1–P7; exported helper 59 OK; exported checker 35 OK; verbatim `cmp` same; gitleaks `[]` (`0 0 0`); manifest `476c1589…` | `a2-b2/` |
| B3 | a2 stage | PASS: 69/69. The C1, C2 and C3 checks are listed in §10.4; all attempt-1 checks still pass. | `a2-b3/` |
| D1 re-export | a2 stage | PASS (§10.2) | `a2-d1/` |
| B4 | DEST | PASS: 10/10, including a fresh install through the README procedure (`INSTALLED` ×3) and resume | `a2-b4/` |
| B5 | `3058d2b` | PASS: `flock …/local-harness.lock bash infra/ci/validate.sh --profile full` final status 0, all phases 0; 88 PASS, 5 INFO, 0 FAIL (B-RENDER-PLAN, B-PROTECTED, B-SKILL-UNCHANGED, B-CHECKS-ROOT, B-SCOPE, B-REQUIRED-CHECKS); cleanup 0. Scope diff from `3a3f4b8`: only §3.1 paths; protection diff exit 0; DEST `skills/sdlc-management` equals the 29 blobs at `3058d2b` | `a2-b5/` |

### 10.4 Correction evidence

- **C1** (repo B, an existing repository with application content):
  - `app/templates/invoice.html` (`{{ customer.name }}`, mustache sections), a README with a public source URL, and
    `docs/tenants.md` with a real random UUID are byte-identical after adoption, and adoption passes.
  - Negatives on copies fail in their proper scopes: a source-origin literal in app code → A6 `private literal`; a
    private path in an installed skill → A6 `home-path`; `{{UNDECLARED_X}}` in PROGRESS → A3.
  - Checker tests (35): app content survives; placeholders in `AGENTS.md.sdlc-proposed`, a release Spec and
    `docs/RISKS.md` fail, while app `{{ title }}` does not; literal, private-key and installed-skill path/UUID/remote
    plants fail; P6 package tests are unchanged.
- **C2:** repo B had an edited `claude-session-handoff` (local helper edit) before adoption.
  - The install logged `COLLISION claude-session-handoff` and `INSTALLED` for the other two skills. The edited
    tree hash is unchanged, and the collision and owner decision are recorded in
    `management/r1/results/G000-adoption-install-record.md` (SYNTHETIC decision).
  - Rerun: `IDENTICAL` ×2, and the collision is kept.
  - A symlinked `.agents/skills` gives `STOP` ×3, with nothing written through the symlink.
  - Repo A and repo C fresh installs (`INSTALLED` ×3) still pass.
  - The snippet run is extracted verbatim from the bundle README, and the script asserts it equals `adoption.md`.
- **C3** (repo A):
  - The rendered CLAUDE.md has no unconditional "Stop for explicit Plan ACCEPT".
  - G003 (direct tier explicitly authorized) → "proceed under authorized Goal/Spec without a separate Plan".
  - G004 (ordinary) → "stop for explicit Plan ACCEPT".
  - G004 with the implementer requesting the direct tier → self-selection refused.

### 10.5 Publication status (appended; §1 text preserved)

§1 was written before publication and says nothing was pushed. Afterwards, the coordinator made a safekeeping push
of `88e07f5` to `origin/g053-restruct-sdlc-bundle`: a new branch, no MR and no pipeline, after a scope, secret and
path scan. Codex verified it independently. This correction's commits are not pushed by me; the coordinator
publishes the corrected checkpoint. There is still no MR, merge, hosted disposition or final ACCEPT.

### 10.6 Execution Budget update and checkpoint (B3, after attempt 2)

Budget class: B3. Maximum attempts: 3. Attempts used: 2 (attempt 2 submitted). Attempt 1 is preserved as recorded
in §8.

2. Authorized correction of C1–C3. Inner loops: none failed in attempt 2, apart from the gate-ordering slip above.
   No subagents. Boundary: this resubmission.

**Checkpoint assessment.**
- Criteria: all twelve are evidenced locally. AC11's hosted MR `fast` and staging disposition remain pending.
- Failures and patterns: attempt 1's inner failures were test, walkthrough-script and scope defects, all fixed. The
  three review findings were scope and wording gaps, each fixed with focused evidence. No repeated failure pattern.
- Convergence: yes. The corrections were bounded to six skill files and the README, with no scope, runtime or
  framework expansion.
- Remaining uncertainty: hosted results only.
- Bounds: the Goal remains appropriately bounded. Attempt 3 would be used only for another bounded correction.

## 11. Observer amendment: input reconciliation and change map (attempt 2 of B3)

Recorded before the observer validation batches. No replan and no new worker; the same owner works under mapping
`4b5b3250…`. AC13 was added, so thirteen ACs apply.

### 11.1 Rebases and input reconciliation

**First rebase** (owner authorization relayed by Codex, handoff `ad488a82`).
- Backup branch `g053-pre-observer-rebase-16e4f03`, created at the old candidate `16e4f03`. It was then an ordinary
  `git rebase origin/staging` onto `8647def`, with no `-X` strategy.
- One add/add conflict, on this Result in the first replayed commit. Staging held the imported `95a6e762…` version.
  I resolved it explicitly to that commit's own version (the addendum), and later commits replayed cleanly.
- The final Result equals the newest candidate `d4743e50…`, which already contains everything in `95a6e762…`.
- Staging's imported review history, including `G053-implementation-codex-review.md`, is kept.
- `git diff 16e4f03 HEAD -- skills <Result>` is empty, so the task content is preserved. New HEAD: `2b2d9f4`.

**Second rebase** (Codex handoff `8c84b426`).
- Staging advanced to `96b3a0a` (AGENTS.md, PROGRESS.md and the amendment). It was an ordinary rebase with no
  conflicts. New HEAD: `3b49102`.

**After both rebases:**
- `origin/staging` is an ancestor.
- `git diff --name-only origin/staging HEAD` lists only this Result and `skills/sdlc-management/**`.
- Protected and management inputs equal staging.
- Plan r2 `f0928ed1…` and the r1 archive `0d78814a…` are byte-identical.

**Re-bound inputs at `96b3a0a`** (36 rows; earlier binding files kept in `$EVID`):

| Path | Previous | Now |
|---|---|---|
| Goal | `2d8702b2…` | `079416689ea11037203612eecf50f841788d3e7c72b4a685a31a9a49fffda65d` |
| Spec | `10df57a8…` | `e422e91ae352db547c0efc5914647bf8908989a05b2b39fd8a40719472fb70ac` |
| PROGRESS.md | `852e2097…` | `00edfc36a5f95a4013961236b1d2cdda91e4437904efc5cccabeba787d387b3f` |
| AGENTS.md | `4dc0ac8e…` | `e550c93eff397f53656d22906f043ebb4150743aa1cc1f2114621059d1845017` |
| docs/agents/CLAUDE-SESSION-HANDOFF.md | `4ba73794…` | `438933f9abb0457509f1077a97190bc37e5ac81a1ac579b12ccea8fc65e9557a` (owner attribution section) |
| .gitignore | `cefaa07d…` | `e5ba1d203c3f31b6c397802a2f0da142bafc86ae1ba312f6e470632cde59639f` (`/.codex/worktrees/`) |
| observer amendment | — | `b9f0c1de…` superseded by `3bdc6a92ee91ce1bd944ca36fe2bab96d7d205cde6b120828ff5b490449c398a` (adds the supervisor publication option) |
| observer blocker (resolved) | — | `be670c2f724232fbe7c94eb9340d22d64d57c3d3b4b1542ab937ed3001380f8a` |
| implementation review | — | `21925f3b1695349732ba62c358368f4c72c26987187072ef0f9563220a40daff` |

CLAUDE.md, SUPERVISOR-HANDOFF, the budget policy, both companion SKILL.md files (before my edits), the helper and
its tests, CI and validation are unchanged. The bound companion files changed below are the amendment's named
exception and are re-bound after my edit commit.

### 11.2 Change map

| Item | Where | Evidence |
|---|---|---|
| AC13 watcher | **New** `.agents/skills/claude-session-handoff/scripts/watch.py` (stdlib). Reuses the helper's `status_report`/`live_registration` pin verification, private state primitives and `safe_excerpt`. It queues to one explicitly bound supervisor UUID with `codex queue`. Commands: `bind`/`unbind`, `start`/`stop`/`status`, `smoke`, `poll-once`; the internal `run` loop holds a per-repository lock. State is 0700/0600 under the helper's repository state directory (`observer/`), outside Git. Verified Codex CLI 0.160.0 and daemon 0.162.1; other versions send nothing. | focused tests; live smoke |
| AC13 notification sources | Only assistant rows of the pinned coordinator's own main transcript qualify. Incoming handoffs, echoes (`[handoff-id:`, `[observer-notice:`), queue operations, sidechains, repository files, Codex transcripts and observer records never qualify. "Waiting" is reported as native prompt not inspected. Deduplication and offsets persist across restarts, task closeout and log compaction. Identity loss, retirement, a changed pin, unverified versions or queue uncertainty suspend delivery until an explicit `bind --replace`; there are no retries. | `test_watch.py` |
| AC13 tests | **New** `scripts/test_watch.py`, using fake status, transcript and Codex boundaries. Covers idle polling with zero calls, dedup and history, restart, partial and rotated files, question/sidechain, waiting, Codex edits, incoming and echoed handoffs, notice echo then a genuine milestone, closeout, unknown/changed/retired identity, bind rules, the status fingerprint, unsafe state, queue failure and timeout, versions, single smoke, start/duplicate/stop, PID reuse, interval bounds and the single-instance run loop. | B-OBS focused batch |
| AC13 docs | Handoff SKILL.md: new observer section (startup, private state, limits, owner attribution versus native permission). `docs/agents/CLAUDE-SESSION-HANDOFF.md`: observer section. | inspection; export |
| AC13 export | The companion export adds `watch.py` and `test_watch.py` (verbatim). The adapted handoff SKILL.md carries the new section; its deltas are re-derived mechanically. Re-export follows the destination-ownership procedure. | B2/D1/B4 |
| Owner attribution (amendment) | Into the sdlc-management coordination reference and templates, distinguishing native permission. | B3 inspection |
| Supervisor publication option (`3bdc6a92`) | New parameters `SUPERVISOR_AUTO_PUBLISH` (default `true`) and `INTEGRATION_BRANCH` (explicit, example `staging`). The wording names actor, trigger, default, opt-out and branch, in the AGENTS template, the roles reference and the adoption parameters. | existing P3t/A-checks plus B3 rendered-text inspection |

Unchanged: `handoff.py` (`e2bdce82…`), `test_handoff.py` (`aa52b129…`), and all application, foundation and CI
code. The watcher is never enabled persistently, and the live smoke is exactly one labelled benign delivery.

## 12. Observer (AC13) record and final reconciliation (attempt 2 of B3)

### 12.1 Implementation

Commit `51c51d4` (tested candidate), on the reconciled baseline `96b3a0a`. It contains the additive observer, the
focused documentation and the sdlc-management publication and attribution option.

| Path | SHA-256 |
|---|---|
| `.agents/skills/claude-session-handoff/scripts/watch.py` (new) | `99bbc83feddc0852359d366331ade7b4dfd61c1142c1a2626f336ca331e625b6` |
| `.agents/skills/claude-session-handoff/scripts/test_watch.py` (new) | `5336909b6726a1f54bf6015d9f9d21673759b3995222226099807d3cdb04e33f` |
| `.agents/skills/claude-session-handoff/SKILL.md` (observer and attribution sections) | `4af35bec91c43636dfd41de4d12cc5a111b5968e0fee88871959c25d8c142f20` |
| `docs/agents/CLAUDE-SESSION-HANDOFF.md` (observer section) | `acc00b95894bda81af9e9b8f2adfe3f13f547a834d89d0afab1aee2ef4595015` |
| `skills/sdlc-management/SKILL.md` | `353ff1f0d5c02d3804a8d7cbb5c4e78c53538ca81a60ae7bbdb05ad3b719f038` |
| `skills/sdlc-management/references/adoption.md` | `1ce2df6bfdc1dc24f84309d979e1c7e5dfdd1c7982cfee823f47c9919ce78ea9` |
| `skills/sdlc-management/references/coordination.md` | `bd358a0f3b0f779104b3243c556d739787ddeedef2ad776e1f2ef264347e71ca` |
| `skills/sdlc-management/references/roles-lifecycle.md` | `862833765ebd7c367bcd18233f3beaf25e1dd134aae88fb14ecc3b723ffc3f77` |
| `skills/sdlc-management/templates/parameters.example.json` | `32f7af0fba054470bade68dfc270e9b36a6018d2c1ec9522895d07810bf3129a` |
| `skills/sdlc-management/templates/root/AGENTS.md.tmpl` | `c145842fe8a16624b4dd747ebfc75dbf9ef7e7d29887c444b1b30c8cd185a2ee` |
| `skills/sdlc-management/templates/root/CLAUDE.md.tmpl` | `687840a3bddcec0f154f2168c6032c031966a29e53f1db290548fb1251d473b4` |

`handoff.py` (`e2bdce82…`) and `test_handoff.py` (`aa52b129…`) are byte-identical, as are all application,
foundation, CI and validation sources. No runtime script lives under `management/`.

### 12.2 Validation (observer round)

Every batch ran after the pre-batch gate, against target `96b3a0a` with the re-bound 38-row inputs. The smoke and
the bundle steps are not Docker runs; B5 used the shared lock.

| Batch | Candidate | Outcome | Evidence (`$EVID/…`) |
|---|---|---|---|
| dev | `5c53947` + uncommitted watcher | `test_watch.py` 24 OK, 1 FAIL: the lifecycle test did not reap its own fixture child (a zombie in the test process). Fixed by reaping with `waitpid`; rerun 25 OK with `-W error::ResourceWarning` | `prebatch-obs-dev1.log`, `prebatch-obs-dev2.log` |
| B1 | `51c51d4` | PASS: checker 35 OK; skill package check exit 0 (SKILL.md 127 lines); source helper 59 OK; watcher 25 OK | `obs-b1/` |
| B2 | `51c51d4` | PASS: stage P1–P7; exported helper 59, checker 35 and watcher 25 OK; `cmp` of 34 verbatim files (including `watch.py`/`test_watch.py`) all same; adapted-file delta lines unchanged from §5; gitleaks `[]` | `obs-b2/` |
| B3 | `51c51d4` stage | PASS: 72/72. Includes the publication option (default `true`/`staging` in repo A; `false`/`develop` in repo B, with actor, trigger, opt-out and implementer workflow) and owner attribution with the native-permission distinction | `obs-b3/` |
| Live smoke | `51c51d4` | PASS (§12.3) | `obs-smoke/` |
| D1 re-export | `51c51d4` stage | PASS: owned DEST `476c1589…` moved to `superseded/20261012T181526Z/` and recreated create-only; manifest `a0f6fe0b82b77eb0bf861eac4f9686d5c184b78c34ca0316d7dad67aaee2135d` (43 files) | `obs-d1/` |
| B4 | DEST | PASS: 11/11, including the README fresh install and the installed observer tests (25 OK) | `obs-b4/` |
| B5 | `51c51d4` | PASS: `flock …/local-harness.lock bash infra/ci/validate.sh --profile full` final status 0, all phases 0; 88 PASS, 5 INFO, 0 FAIL (B-RENDER-PLAN, B-PROTECTED, B-SKILL-UNCHANGED, B-CHECKS-ROOT, B-SCOPE, B-REQUIRED-CHECKS); owned_remaining 0. Scope from `96b3a0a`: only allowed paths (skill, Result, the observer files, handoff SKILL.md, runbook); protection diff exit 0, including `handoff.py`/`test_handoff.py`; DEST skill and watcher blobs equal HEAD | `obs-b5/` |

**AC13 focused tests** (`test_watch.py`, 25, fake status/transcript/Codex boundaries):
- Idle polling makes zero queue and CLI calls.
- One notice per milestone; history is not replayed; restart keeps dedup and offsets; partial lines wait; rotated
  and truncated files are handled.
- Native question, sidechain and non-milestone rows; the waiting transition says the native prompt was not
  inspected.
- **Negatives:**
  - Codex repository and management edits cause zero calls.
  - Incoming handoffs, queue operations and echoes cause zero calls.
  - A notice echoed back and acknowledged does not retrigger, and the following genuine milestone notifies once.
  - Task closeout (worktree removal, log compaction) keeps monitoring and dedup.
  - Unknown, changed or retired identity suspends delivery until an explicit `bind --replace`, never retargeting.
  - Bind requires a verified pin and a canonical UUID.
  - `status` shows a fingerprint, not the UUID.
  - Unsafe state (mode, symlink, directory mode) blocks.
  - Queue failure and timeout are recorded as uncertain with no retry.
  - An unverified daemon version sends nothing.
  - The smoke delivers exactly once.
- Start, duplicate start and stop, with the process reaped. A reused PID is never signalled. Interval and binding
  bounds hold, and the run loop is single-instance and stops on signal.

### 12.3 Live smoke (redacted)

- The private binding was supplied by the parent outside the repository; its UUID and path are not reproduced here.
- Pre-checks: `handoff.py status` gave `pinned-live-verified`, coordinator `580cdb2f…`, discovery complete. Codex CLI
  `0.160.0`, daemon `appServerVersion` `0.162.1`.
- `watch.py bind --supervisor-file <private>` → `bound`, supervisor fingerprint `f81932f016d8`.
- `watch.py smoke --label "G053 AC13 benign observer smoke"` → **exactly one** `codex queue` call. Outcome
  `queued-not-acknowledged`, notice `<redacted-uuid>`, message SHA-256
  `aba218246e9629ee5d4460da8addae007b1016d1db86ac2302a7c8a6c1740975`. The text is a labelled benign delivery check
  that requests no reply, task action or approval. Acknowledgment by the supervisor is not claimed.
- `status` showed `stopped`, with no watcher process ever started. `unbind` → `unbound`.
  `pgrep -af watch.py` found nothing.
- The private `observer/state.json` (0600, directory 0700) retains the single delivery record for dedup. The observer
  is **not enabled**: no binding and no process.

### 12.4 Final AC reconciliation (13 ACs)

| AC | Status | Evidence |
|---|---|---|
| AC1 | met | §6 plus the obs-b1/obs-b2 package checks. The observer lives in the handoff companion only; sdlc-management references it and adds no transport. |
| AC2 | met | §6, §10 (C3) and obs-b3 (72/72). |
| AC3 | met | §6; obs-b3. |
| AC4 | met | §6, §10 (C1/C2) and obs-b3; the foundation tree hash is unchanged. |
| AC5 | met | §6; obs-b3. Helper tests are 59/59 in source, export and install. |
| AC6 | met | §6; obs-b3. Observer notices are recorded as `queued-not-acknowledged` and carry no authority. |
| AC7 | met | §6; obs-b3. The supervisor publication option is documented with branch selection. |
| AC8 | met | §6; obs-b3. Observer state survives closeout (`test_watch` closeout test). |
| AC9 | met | obs-b2/obs-d1: manifest `a0f6fe0b…`, P5–P7, the literal list and gitleaks clean. The B5 scope and protection diff (§12.2) shows only allowed paths. |
| AC10 | met | obs-b4 (11/11) from a second path, including the installed observer tests. |
| AC11 | met locally; hosted pending | B1/B2 and the walkthroughs, plus B5 full profile (final status 0, 88 PASS) on `51c51d4`. MR `fast` and the staging full-gate disposition remain to run after review. |
| AC12 | submitted | §§0–12 of this Result. |
| AC13 | met locally | `watch.py`/`test_watch.py` hashes (§12.1); 25 focused tests including every amendment negative; export and manifest; docs; one redacted benign live smoke with verified CLI/daemon versions and cleanup (§12.3). |

**Remaining** (not only hosted gates):
- Codex implementation review of the correction and of AC13 (the observer is newly authored and has not been
  reviewed).
- The coordinator's safekeeping push of this checkpoint.
- MR creation after pre-integration review, then MR `fast` and the owner-authorized merge.
- Staging full-gate disposition, then final ACCEPT.
- Public publication of the bundle remains a separate owner action.

### 12.5 Execution Budget update

B3, maximum 3. Attempts used: 2. Attempt 2 covers C1–C3 and the AC13 observer as one shared attempt (amendment:
"Codex contribution joins current attempt 2 without resetting it"). The checkpoint after attempt 2 falls at this
submission; the §10.6 assessment stands, with AC13 now evidenced locally. Attempt 3 is unused and needs supervisor
direction.

### 12.6 Owner README edit and destination reconciliation (focused, docs-only)

**The owner edit.** Codex added the owner-authorized "Why Restruct?" README section directly in `../restruct-sdlc`
(Codex handoff `50e6dfdc`), giving README `415d4f2f1575d3feb8925756fdcd1a6ec71e523acfa845f8617f7c6c8b281a4c` and
MANIFEST `7787825e208c6a4c6781c2c4d9c20402b06c30530feeca6c23a98b1f56ec5b1f`.
- I carried the owner's prose forward verbatim into the export README source. The previous source `a79f6bf9…` is
  kept in `$EVID/tools/`.
- A focused re-export at `51c51d4` (`$EVID/obs-readme/`) produces MANIFEST `7787825e…`. Its content equals DEST
  file for file, ignoring DEST's top-level VCS directory.
- The obs-b5 full run tested `51c51d4` with the pre-owner README. It is not claimed for these README bytes. The
  README is bundle-only, and no repository path changed.

**Focused checks on the reconciled stage:**
- Package P1–P5 and P7 pass. Without the private literal list, P6 passes as well (exit 0).
- With the list, P6 reports exactly one entry: the term "Vestive" in the owner-authored README links. This is an
  **explicit owner-authored exception** to the private literal list, recorded in the ownership record. It is not
  an export leak.
- Exported helper 59, checker 35 and observer 25 OK; gitleaks `[]`; README links are external `https` (P3).

**Destination state changed outside G053.** DEST now also contains a top-level Git directory with an initial
commit and a push to a public hosted repository under the owner's identity (about 14:40 local time). The G053
implementer did not create, modify or further inspect it.
- Consequences:
  - DEST is no longer the unpublished, implementer-owned tree from Plan r2 §4.3. A package check on DEST itself now
    reports P5/P7 for that directory, by design.
  - The move-and-recreate procedure must not be used on it.
  - Later changes go out as a validated stage plus a recorded diff for an owner decision.
- `destination.json` was updated to `state: owner-modified-and-owner-published` (SHA-256
  `18a10a446dd4e157d664a77b37bdc52709395d23fca47c4a7d2abcbc5a3c23ba`); the previous record is kept.
- Public publication was outside G053's authority, and this Result treats it as the owner's separate action. **I
  am reporting it to Codex for disposition**; I neither executed nor approved it.

## 13. Late amendment reconciliation: lifecycle, 30-second default, test activation (attempt 2 of B3)

### 13.1 Rebases and amendment inputs

Three more ordinary rebases of my own task branch, with no conflicts. Each was followed by the same checks:
staging is an ancestor; only allowed paths differ; Plan r2 `f0928ed1…` and r1 `0d78814a…` are unchanged.

| Staging | Amendment SHA-256 (superseding) | Change |
|---|---|---|
| `6bd14ff` | `a33e2525be330946a2f34b8641c76449afc5cfd67754f4f2311c27395491faba` (← `3bdc6a92`) | Operator activation of continuous monitoring |
| `e77ec1f` | `327eb9dc6e462ccb557ffba08478973762008d865971045825456dc6e9c61a42` (← `a33e2525`) | Activation deferred until integration |
| `286944c` | `1b84ee20c92668cd27c8c41d237c28a8d1dd0b526ffa23a5a9b67289ae9aed5f` (← `327eb9dc`) | Default poll interval 30 seconds |
| `0e9559b` | `160f6ac57c57a90f5a0c3b15a3184036f02fea3f5d4f579d8d680a9194e1c407` (← `1b84ee20`) | Temporary test activation permitted |

PROGRESS.md was re-bound to `8a0f2ada…`. My own edits were re-bound after commit `385099f`, as listed in §13.3.

### 13.2 Observer lifecycle record (truthful sequence)

1. The owner explicitly authorized enabling continuous monitoring, and Codex relayed that authorization
   (handoff `a0c7feac`; amendment `a33e2525`). Codex started the observer for the current coordinator–supervisor
   pair at a 20-second interval from a private, hash-recorded operator runtime outside task worktrees: the
   `g053-observer-runtime` directory in the private platform-evidence state, referenced here only. That runtime's
   `watch.py` equals candidate `99bbc83f…`, and its `handoff.py` equals `e2bdce82…`. The monitor was operator-owned;
   G053 did not start, stop or modify it.
2. The owner then superseded this: "Ok disavle it and only run once we integrated" (relayed by Codex, handoff
   `ebb2ea82`; amendment `327eb9dc`). Codex stopped the observer, verified there was no PID, and kept the private
   binding and deduplication history.
3. My read-only `watch.py status` at 19:11Z showed `stopped`, `pid: null`, `bound: true`, not suspended, one
   delivery (the historical smoke) and one dedup key. The private log shows polls with `queue_calls: 0` from 15:09
   to 15:11 local, then `stopped`. `ps` shows no `watch.py` process.
4. **The monitor is not running.** Continuous monitoring waits for G053 integration, after which the integrated
   skill is started at the 30-second default with re-verified identities. That is not part of this task. The
   reusable procedure still requires an explicit verified binding and activation.

### 13.3 30-second default (handoff `7e7f2e13`)

Commit `385099f` makes these changes:
- `watch.py` `DEFAULT_INTERVAL` goes from 20 to 30; explicit `--interval` overrides are unchanged.
- The handoff SKILL.md text ("default 30 seconds") and its example (`start [--interval 30]`) are updated.
- The runbook gains a line stating the 30-second default, that `--interval` overrides it, and the owner's
  deferral until integration.
- A new focused test asserts default 30 for `start` and `run`, a recorded interval of 30 when none is given, and
  an explicit `--interval 45` honoured in the record and argv. That 45 is a test-specific override.

sdlc-management states no interval, so nothing there changed.

| Path | SHA-256 |
|---|---|
| `.agents/skills/claude-session-handoff/scripts/watch.py` | `0c5b8a91fbe8edbcbccb04214effa1e1a9b2f88b6d5ec3ffc759f2b89c9d3ebf` |
| `.agents/skills/claude-session-handoff/scripts/test_watch.py` | `2edfa26d3b273c6f1c42b9bf2f37ad4f19248b5bd62954ad26b1999ee6c20a9d` |
| `.agents/skills/claude-session-handoff/SKILL.md` | `1db7f48716c0e9f2d23774327dffe1e0503b856875c5289e6def31c9db1ee582` |
| `docs/agents/CLAUDE-SESSION-HANDOFF.md` | `0d68f5fa01330090cdc54fc1b5975bda69927c44bd7bc3f7a7c7c6f2b1bff51e` |

`handoff.py` and `test_handoff.py` are unchanged, as are all application, foundation and CI sources.

**Focused validation.** Both batches ran after the pre-batch gate, with target `0e9559b` and the re-bound 38 rows.

| Batch | Candidate | Outcome | Evidence |
|---|---|---|---|
| B1 | `385099f` | PASS: observer 26 OK (one new test), helper 59 OK, checker 35 OK, skill package check exit 0 | `$EVID/int-b1/` |
| B2 (staged) | `385099f` | PASS: exported helper 59, checker 35 and observer 26 OK; 34 verbatim files identical; gitleaks `[]`; P1–P7 pass without literals. With the private literal list, the only finding is the owner-authored "Vestive" README exception (§12.6) | `$EVID/int-b2/` |

The full profile was not repeated, by instruction. Its last run (`51c51d4`, final status 0) is historical
evidence for the content it tested and is **not** claimed for these bytes. The delta is one constant, one focused
test and documentation, none of which the full profile executes.

### 13.4 Staged delivery (bundle published by the owner; not overwritten)

Under the updated ownership record (`owner-modified-and-owner-published`), the published bundle directory was not
moved, recreated or written.
- The staged export is preserved at `$EVID/delivery/interval-385099f/restruct-sdlc`, with MANIFEST SHA-256
  `5a26c9e48b8dcc1b708fe6616c5d003c88610015fa5b3a73462e1af2e741bb97`.
- It differs from the published bundle (`7787825e…`) in exactly four files: `MANIFEST.tsv` and
  `skills/claude-session-handoff/{SKILL.md, scripts/watch.py, scripts/test_watch.py}`. The changed rows are SKILL.md
  `cc35e5e3…` → `89f87e76…`, watch.py `99bbc83f…` → `0c5b8a91…` and test_watch.py `5336909b…` → `2edfa26d…`.
- The README is byte-identical to the owner's `415d4f2f…`.
- The unified diff for the owner's decision is `$EVID/delivery/interval-385099f/staged-vs-published.diff`
  (SHA-256 `5ad54d0c56449e61a1aefd868c277ba5e5c33cb79c84449edfa14c46ab4d5fa6`, 26 changed lines).
- Applying it to the published repository is an owner action.

### 13.5 Temporary test activation (handoff `f34d2747`; amendment `160f6ac5`): not used

I chose not to run the candidate observer live. It would poll the real coordinator transcript, and a genuine
milestone would produce an unapproved second live notice. That conflicts with the "no further live notice" limit,
and fake delivery boundaries cannot be injected into a real process without changing its code.
- Lifecycle behaviour is already covered with an owned fixture process: start, duplicate start, stop with reaping,
  PID-reuse refusal, interval bounds, single-instance lock, signal stop, and now the default and override interval.
  The one live smoke covers the real delivery route.
- No test process was started; `ps` shows no `watch.py`. The operator runtime, its binding and its history were
  not touched.

### 13.6 Status after this record

All 13 ACs reconcile as in §12.4, with AC13 now covering the 30-second default (observer 26 OK). Local evidence is
complete for the final content, apart from the full profile not re-run on `385099f` (by instruction).
- Remaining:
  - Codex implementation review of C1–C3 and AC13;
  - the coordinator's safekeeping push;
  - the owner's decision on applying the staged bundle diff;
  - MR and `fast`, owner merge, the staging gate and final ACCEPT;
  - post-integration activation by the operator.
- Budget: B3, attempt 2, the checkpoint at this submission; attempt 3 unused.

## 14. Direct delivery to the owner's sibling repository (attempt 2 of B3)

**Authority and attribution.** The owner said: "Restruct sdlc repo is under git now so it can modify files in the
directory directly" (relayed by Codex, handoff `a859c94f`; amendment `160f6ac5` → `373af2adca75b89a14732e392eb8bc01ccdbe7af855a9d361cac47c8d1eaccd5`
on staging `db59e97`). The Git initialization and publication observed in §12.6 were **the owner's own actions,
not a G053 breach**. Local edits only were authorized: no sibling commit, push or remote change.

**Rebase.** An ordinary rebase onto `db59e97`, with no conflicts. HEAD became `603c68c`. Staging is an ancestor;
only allowed paths differ; Plan r2 and r1 are unchanged. PROGRESS was re-bound to `94bf966f…` and the amendment to
`373af2ad…`. The package sources are byte-identical to `385099f`, so the staged delivery is current.

**Conflict check, then apply** (`$EVID/tools/apply_delivery.py`, record `$EVID/delivery-apply/apply.json`):
- The published MANIFEST equalled the last-known `7787825e…`. Every published file matched its manifest row, and
  no unlisted files existed outside the VCS directory, so nobody else had changed any owned path.
- Applied only `skills/claude-session-handoff/SKILL.md`, `scripts/watch.py`, `scripts/test_watch.py` and
  `MANIFEST.tsv`, each by atomic replacement keeping file modes.
- Bundle manifest before: `7787825e208c6a4c6781c2c4d9c20402b06c30530feeca6c23a98b1f56ec5b1f`.
  After: `5a26c9e48b8dcc1b708fe6616c5d003c88610015fa5b3a73462e1af2e741bb97`.
- VCS directory content digest `a3c87807…` was identical before and after (untouched). README `415d4f2f…` was
  identical before and after (owner text kept). No temporary or bytecode files remain.

**Focused metadata-aware checks** (`$EVID/delivery-apply/verify.json`). These ran on a temporary copy of the
delivered tree without its top-level VCS directory, using the unchanged checker:
- package P1–P7 pass. With the private literal list, only the recorded owner-authored "Vestive" README exception
  is reported;
- exported helper 59, observer 26 and checker 35 OK;
- the delivered tree equals the staged delivery file for file. gitleaks `[]` was already recorded for identical
  content in `int-b2`.

The validator was not weakened, and the VCS metadata was neither scanned as payload nor deleted. The ownership
record was updated to `owner-repository-with-authorized-direct-g053-delivery` (`369e83de…`), with the previous
records kept.

**Permission note.** The first attempt to append and commit this section was refused by the native classifier
("Modify Shared Resources") and stopped with no effect. The owner said "Resolve it" (relayed by Codex, handoff
`0a8e1e66`; not typed in the coordinator session), and the exact action was retried once.

## 15. Final scoped additions and review-ready 13-AC reconciliation (attempt 2 of B3; scope frozen)

### 15.1 Inputs and rebase

The owner said "Resolve it" (relayed by Codex, handoff `0a8e1e66`). The refused §14 append and commit were retried
once and committed (`0b6a04d`, now rebased).

An ordinary rebase onto `0fc9bd6` had no conflicts: staging's imported Result was still `95a6e762…`, which is
untouched by these commits. After the rebase, staging is an ancestor, only allowed paths differ, and Plan r2
`f0928ed1…` and r1 `0d78814a…` are unchanged.

| Input | Previous | Now |
|---|---|---|
| Observer amendment | `373af2ad…` | `1d95a88d…` (activation default) → `afc4c4c8…` (size advisory) → `a888cc7b9e83fd0a9e0ef32f64779e7a5297121e6179b51ccbedbfd5b2a37c55` (role effort) |
| AGENTS.md | `e550c93e…` | `5cdaa16d81ab7f757eed39c8bef69617710916692be4dc496fd5208a9a2323de` (role effort clarification) |
| CLAUDE.md | `b3ce5aa9…` | `88ef1bd590f71678d14836261088c98288b041d06e089d1109bcaa7cec81756d` |
| PROGRESS.md | `94bf966f…` | `734c6618a3e43ad11d79440eb173ee7a0a1f42d9a9ad955e2d580ef44bc6f319` |

### 15.2 Changes (commit `38aece3`, the tested final candidate)

- **Observer activation default** (`1d95a88d`):
  - New parameter `OBSERVER_AUTO_ACTIVATE`, default `true`, independent of `SUPERVISOR_AUTO_PUBLISH`.
  - The adoption procedure (new step 7), the AGENTS template section "Milestone observer activation", the handoff
    SKILL.md and the runbook all say the same thing. Once the coordinator pin is `pinned-live-verified` and the
    supervisor UUID has been supplied explicitly, run `bind` and then `start` at 30 seconds, with no further
    question. A missing, ambiguous, unverified or retired binding leaves it inactive and reported. No name
    guessing, session creation or permission change. `false` opts out.
  - `watch.py start` now re-verifies the bound pair first and otherwise returns `inactive` with the reason; an
    already-running observer gives an idempotent `already-running`.
- **Transcript-size advisory** (`afc4c4c8`), in `watch.py`:
  - a metadata-only `stat` size check on each verified poll;
  - binding options `--size-threshold BYTES` (default 104857600 = 100 MiB, minimum 1024) and `--no-size-advisory`;
  - one advisory with size, threshold and a safe handoff/archive recommendation. It is keyed by coordinator UUID
    and threshold, recorded before delivery (no retry), and persists across restart, rotation, size decrease and
    closeout;
  - `status` shows `size_threshold`, `transcript` (bytes, threshold, advised) and `size_advised`;
  - it uses the same identity, private-state, version and queue-uncertainty protections as milestones;
  - documented in the handoff SKILL.md and runbook.
- **Role effort** (`a888cc7b`):
  - New parameter `COORDINATOR_EFFORT`, default `high`.
  - The budgets reference, both root templates, the handoff SKILL.md (new "Effort for coordinators and workers"
    section, placed after the export-delta lines so the deltas are unchanged) and the runbook carry the defaults:
    coordinator high, xhigh when it plans; planning workers xhigh; implementation workers high. Worker effort is
    set explicitly through supported controls and verified, unknown is reported as unknown, and the owner may
    raise complex debugging or security work to xhigh.

| Path | SHA-256 |
|---|---|
| `.agents/skills/claude-session-handoff/scripts/watch.py` | `e2ab5ae9f6b215a7a963bc8b2d202c2e1164b54cd655cbc774e943b0e459991a` |
| `.agents/skills/claude-session-handoff/scripts/test_watch.py` | `7117fa975a70be6eea8d59299d53740c78cdbfb84e7da9f686d0188558f3938d` |
| `.agents/skills/claude-session-handoff/SKILL.md` | `71d7b3f84b5e1874bacd40043e4123ff79f6aadfeb2ab7d6aeefdf1b1f43dc94` |
| `docs/agents/CLAUDE-SESSION-HANDOFF.md` | `8baafc5d745ef0f979753426b5f3e4a8eb5b6ab1f0fd7a44cd01a6c386db4ca8` |
| `skills/sdlc-management/SKILL.md` (130 lines) | `414f829270c94f806836ee47c9db36e2446407f1b03ca100f55fdd82ec36d43e` |
| `skills/sdlc-management/references/adoption.md` | `3a8afc85ba141632485b1c17bf9bd42a2e9b1f07e5036f2b9c9bf436cc63eabd` |
| `skills/sdlc-management/references/budgets.md` | `5d54f542cecfb075c6f7699193b510e4e00d0dce734a9b8e3179448b0919b846` |
| `skills/sdlc-management/templates/parameters.example.json` | `90a96ef2f889fa0708105f98468ba2651a0ba212ee96789b1565eb4c80fc080a` |
| `skills/sdlc-management/templates/root/AGENTS.md.tmpl` | `5c58c2a91913ef2af1240937891c1a8a3a826847db7789c2c330777efcc727ce` |
| `skills/sdlc-management/templates/root/CLAUDE.md.tmpl` | `5b3465a283aee18cb64d7c5c9df23cfbec64c780e0bfb9ddff78fedef03abde6` |

`handoff.py`, `test_handoff.py`, AGENTS.md, CLAUDE.md, .gitignore and all application, foundation and CI sources
are unchanged against staging.

### 15.3 Focused validation (covering checks only; no full rerun, by instruction)

Every batch ran after the pre-batch gate: fetch, ancestry at `0fc9bd6`, mount safeguard, and the 38 re-bound rows.

| Batch | Candidate | Outcome | Evidence (`$EVID/…`) |
|---|---|---|---|
| dev | `3bf3cc9` + uncommitted watcher edits | observer 35 OK. **Process note:** this run followed a gate that reported `BINDINGS-CHANGED` for only my two uncommitted observer files (fetch, ancestry and mounts passed). It is development only; the formal batches below ran after commit and re-binding. | `prebatch-adv-dev1.log` |
| B1 | `38aece3` | PASS: observer 35 (9 size-advisory tests plus the inactive-start test), helper 59, checker 35, skill package check exit 0 | `fin-b1/` |
| B2 (staged) | `38aece3` | PASS: exported helper 59, checker 35 and observer 35 OK; 34 verbatim files identical; adapted-file delta lines unchanged; gitleaks `[]`. P1–P7 pass apart from the recorded owner README literal exception. Manifest `8c452087…` | `fin-b2/` |
| B3 | `38aece3` stage | PASS: 74/74, including the activation option rendered (default, verified start, inactive, no guessing, idempotent and opt-out), role effort in AGENTS.md and CLAUDE.md, and repo B opted out of both options | `fin-b3/` |
| Delivery | `38aece3` | PASS (§15.4) | `delivery-apply-final/` |
| B4 | delivered bundle (copy without VCS directory) | PASS: 11/11, including the installed observer tests (35) and resume; the delivered tree equals the staged delivery | `fin-b4/` |

**Advisory fake-boundary tests:**
- the 100 MiB default and the status fields;
- below, at and above the threshold;
- an already-large first poll qualifies;
- restart, decrease and rotation do not repeat;
- a changed threshold is a new key, opt-out works, and a too-small threshold is refused;
- a replacement coordinator is assessed separately only after an explicit rebind;
- uncertain delivery is never retried;
- unverified identity blocks the advisory.

No real notices were sent. The full profile last ran on `51c51d4` (final status 0) and is **not** claimed for the
later bytes. The deltas since then are the observer code, tests and documentation, none of which the full profile
executes. Whether a further full run is required before the MR is Codex's call.

### 15.4 Direct delivery (owner rules)

- **Conflict check first:** the published manifest equalled my last-delivered `5a26c9e4…`, every published file
  matched its row, and there were no unlisted files.
- **Applied:** the handoff SKILL.md, `watch.py` and `test_watch.py`; sdlc-management `SKILL.md`,
  `references/adoption.md`, `references/budgets.md`, `templates/parameters.example.json` and
  `templates/root/{AGENTS,CLAUDE}.md.tmpl`; and `MANIFEST.tsv`.
- **Bundle manifest:** before `5a26c9e48b8dcc1b708fe6616c5d003c88610015fa5b3a73462e1af2e741bb97`, after
  `8c452087a9042be1342b3c52c04d6412368f7bf15439a282c59da2258bb271d3`.
- The VCS directory digest `a3c87807…` and the README `415d4f2f…` are identical before and after. No sibling
  commit, push or remote change.
- The delivered tree, checked without its VCS directory, passes package P1–P7 and the exported helper, observer
  and checker tests.
- Ownership record `8f837414…`; previous records kept.

The watcher in this repository was not started. No test or fixture process remains, and the operator runtime,
binding and history were not touched.

### 15.5 Review-ready 13-AC reconciliation

| AC | Status | Evidence |
|---|---|---|
| AC1 | met | P1–P4/P3t pass on skill and bundle (fin-b1/fin-b2/fin-b4); SKILL.md is 130 lines. The observer lives in the handoff companion, which sdlc-management references without duplicating transport. |
| AC2 | met | Templates and adoption walkthroughs (fin-b3 74/74): A1–A6, links, eleven Goal fields, naming gate, delegated review and relay, C3 conditional Plan gate. |
| AC3 | met | Budget ledger walkthrough (§6); role-effort guidance added without changing budgets. |
| AC4 | met | Repo B with existing governance, an edited skill and app content (C1/C2); composition rule; foundation hash unchanged. |
| AC5 | met | Naming, dispatch and relocation fixtures (a)–(h); helper tests 59/59 in source, export and install. |
| AC6 | met | Receipt versus acknowledgment tests and walkthrough. Observer outcomes are `queued-not-acknowledged`; notices carry no authority. |
| AC7 | met | Validation and integration walkthrough. Supervisor publication option (default true, explicit branch, opt-out). |
| AC8 | met | Closeout and archive fixtures. Observer dedup survives closeout and log compaction. |
| AC9 | met | Manifest `8c452087…` verified; P5–P7 and gitleaks clean; owner README literal exception recorded; scope and protection unchanged against staging. |
| AC10 | met | Relocated, delivered bundle: flat install, adoption, installed tests and resume (fin-b4 11/11). |
| AC11 | met locally; hosted pending | Focused batches on `38aece3`; full profile status 0 on `51c51d4`. A further full run is Codex's call. MR `fast` and the staging gate remain. |
| AC12 | submitted | This Result, §§0–15. |
| AC13 | met locally | Observer source, tests (35), docs and export. One redacted benign live smoke with verified versions. Lifecycle activation then disable recorded (§13.2); 30-second default; size advisory; verified activation default; cleanup verified. |

**Remaining:**
- Codex implementation review (C1–C3 and AC13 including these additions);
- the coordinator's safekeeping push;
- MR and `fast`, owner merge, the staging full gate and final ACCEPT;
- operator activation after integration.

**Budget:** B3, attempt 2, checkpoint at this submission; attempt 3 unused.

## 16. Attempt 3 of B3 (final): review r2 corrections C4/C5 and the asynchronous-review documentation

**Authority.** The [implementation review r2](G053-implementation-codex-review-r2.md) (SHA-256
`7e3dde9685512f481d1813f736872b7e3f4f54ffcf27d630f0bda5e90e377c80`, staging `c07ea00`; Codex handoff `97d8bef8`)
returned CORRECTION LOOP and authorized **attempt 3, the final B3 attempt** (no reset), limited to C4/C5, their
tests and docs, export, and this Result. Owner doc-only addition: asynchronous delegated-review completion
(handoff `a9a50135`; amendment `a888cc7b` → `c65de410a3ae8683f66f0db8dc64d49da43d06c9bbd7acfd567fca0137e591d5` on
staging `34ee448`, together with AGENTS.md `b6278dce…` and PROGRESS `1299ad25…`). Under that amendment's explicit,
**doc-only companion-source exception**, `.agents/skills/repository-supervisor/SKILL.md` gains one paragraph; its
other bytes and all executable sources are unchanged.

**Rebases.** Ordinary rebases onto `c07ea00`, then `34ee448`, with no conflicts. Staging is an ancestor; only
allowed paths differ; Plan r2 and r1 are unchanged. The 39 inputs (now including review r2) are re-bound at
`34ee448` and after my own commits.

### 16.1 Changes

- **C4 (durable intent).**
  - `deliver` now appends a `pending-intent` record, with its dedup key, to the deliveries list and **saves state
    before calling `codex queue`**. If that save fails, nothing is queued.
  - Afterwards the record's outcome is updated (`queued-not-acknowledged`, or `queue-uncertain-*` with suspension).
  - On the next poll, `recover_pending` marks any unfinished intent `uncertain-interrupted` and suspends delivery
    (`interrupted-delivery-uncertain`). `bind --replace` converts pending intents to uncertain without retrying,
    and the dedup keys stay. The smoke path persists its intent and is never repeatable.
  - Transcript offsets are now committed only after every delivery in the poll, so an interruption cannot skip the
    poll's other, unattempted events.
- **C5 (oversized rows).** `new_rows` is now bounded in memory and work:
  - it reads in 1 MiB chunks, up to 16 MiB examined per poll;
  - a row without a newline within 1 MiB is tracked as `oversized` (start, scanned, a 4 KiB prefix) and scanned
    forward, resuming across polls while it grows, until its newline;
  - it is never parsed, and its fragments are never parsed as records;
  - a short partial tail waits; rotation or truncation resets it.
  - Each oversized row is recorded in `parse_limits`, which `status` shows as `parse_limit` (row byte limit, count,
    last). If its prefix shows coordinator (non-sidechain) assistant output, one `parse-limit` notice says it could
    not be assessed for milestones. Later rows continue normally.
- **Documentation:**
  - C4/C5 behaviour is described in the handoff SKILL.md and the runbook.
  - The three non-blocking follow-ups are done as bounded prose: the observer heading now distinguishes opt-in from
    the adoption default; SKILL.md "Package verification" gives a recipe for checking a Git-backed bundle as a copy
    without VCS metadata; and the GitLab adapter's fetch step now stops on failure (`git fetch origin || exit 1`)
    instead of rebasing against stale refs.
- **Asynchronous delegated review (doc-only).** Added to the roles-lifecycle reference, the AGENTS template, the
  SUPERVISOR-HANDOFF template, the SKILL.md lifecycle summary and the supervisor companion:
  - keep the bounded review turn open with the harness's asynchronous wait/notification interface, and the owner
    may steer meanwhile;
  - finalize, record and relay before the final reply;
  - use bounded waits with no repeated inference polling;
  - waiting is token-free, while reviewer work and result processing use tokens;
  - automatic wake-up after an ended turn is not guaranteed;
  - no scheduler or extra agent.
- **Export delta.** The supervisor export's invariants delta moved mechanically from L87–88 to L95–96. The adapted
  diff lines are identical to before.

| Path (commits `3df28c1`, `f5da33e`) | SHA-256 |
|---|---|
| `.agents/skills/claude-session-handoff/scripts/watch.py` | `cec22f681f3dadd86090cc562f6202103d4f0703992e5c5fa4280d403562f85d` |
| `.agents/skills/claude-session-handoff/scripts/test_watch.py` | `7ff3abf09cd173aab72c5b41ec523f1f56fb84f2f4bdfb8e6c8a43d7eff8efe6` |
| `.agents/skills/claude-session-handoff/SKILL.md` | `6c072c0c554ed0d023a658f4ac9b5ab34654e17ce0936a5fa32baedf8f98129f` |
| `.agents/skills/repository-supervisor/SKILL.md` (doc-only exception) | `506e9c6c4054cb048854a6df2c6c05712c298fc36ca930cf747335e7889ba510` |
| `docs/agents/CLAUDE-SESSION-HANDOFF.md` | `f5137c9db29cb18fb3cf0dff114cdfb8d992df41ccb14a1415b9432486322b7f` |
| `skills/sdlc-management/SKILL.md` (134 lines) | `ef590c284c27c1e38bbe93abfd624b00d8477a91b7669e09abcda274e096f4ea` |
| `skills/sdlc-management/references/roles-lifecycle.md` | `60df97ae73947fb9c98a663b07734304b8968af7ce3147fc2d6802ccf5430bb6` |
| `skills/sdlc-management/references/validation-integration.md` | `f8de168c5fbbe18603eda49e278e3871e03407c3daa9674acae01bbf7c17e543` |
| `skills/sdlc-management/templates/root/AGENTS.md.tmpl` | `85231d83aa6a96de00ea93941cc98655c678e746eaf4eb9662f5dda7db5bb7b4` |
| `skills/sdlc-management/templates/management/SUPERVISOR-HANDOFF.md.tmpl` | `4d6bf0ee2be123159ac1efda2a1610ee136ef4df06340dea5303c20bcbdc4d56` |

`handoff.py`, `test_handoff.py`, the supervisor's `agents/openai.yaml`, AGENTS.md, CLAUDE.md, .gitignore and all
application, foundation and CI sources are unchanged against staging.

### 16.2 Review r2 tests mapped to fake-boundary tests (`test_watch.py`; 43 tests in total)

| Review requirement | Test(s) | Result |
|---|---|---|
| C4: interruption after enqueue: milestone, advisory, smoke | `DurableIntent.test_interruption_after_enqueue_never_retries` (subtests ×3; the process "dies" right after the fake enqueue) | PASS: 1 queue call each; recovered `uncertain-interrupted`; suspended (smoke refused); no retry after `bind --replace` |
| C4: failed final persistence: milestone, advisory, smoke | `DurableIntent.test_failed_final_persistence_never_retries` (subtests ×3; the post-enqueue save raises) | PASS: as above |
| C4: no queue on failed initial persistence: milestone, advisory, smoke | `DurableIntent.test_failed_initial_persistence_never_queues` (subtests ×3) | PASS: zero queue calls |
| C4: other events in an interrupted poll are not lost | `DurableIntent.test_later_events_in_the_same_poll_are_not_lost` | PASS: after rebind, only the unattempted milestone is delivered |
| C5: complete oversized non-milestone row, then a normal milestone | `OversizedRows.test_complete_oversized_nonmilestone_row_then_milestone` | PASS: the milestone is delivered once; offset at EOF; `parse_limit` recorded (not coordinator output) |
| C5: growing partial oversized row that later completes; offsets progress; no whole-transcript rereads or fragment records | `OversizedRows.test_growing_partial_oversized_row_progresses_and_completes` | PASS: `scanned` advances from 2500 bytes to the next bound without restarting; per-poll work ≤ MAX_SCAN+MAX_READ; on completion exactly one parse-limit report and the following milestone; offset at EOF |
| C5: an oversized assistant milestone is not silently dropped | `OversizedRows.test_oversized_coordinator_row_is_reported_not_silently_dropped` | PASS: one `parse-limit` notice, not repeated |
| C5: rotation handling | `OversizedRows.test_rotation_clears_a_pending_oversized_row` | PASS |

All earlier observer tests still pass.

### 16.3 Focused validation (after the pre-batch gate; no duplicate full run)

| Batch | Candidate | Outcome | Evidence (`$EVID/…`) |
|---|---|---|---|
| dev | `312fc38` + uncommitted C4/C5 edits | observer 43 tests: 2 failures (the smoke recovery was not persisted before refusing), fixed (save, then refuse); rerun 43 OK. Development only. | — |
| B1 | `3df28c1` | PASS: observer 43, helper 59, checker 35 | `a3-b1/` |
| B2 (attempted) | `3df28c1` | **Not launched:** the pre-batch gate found staging advanced to `34ee448` (ancestor NO). I rebased and reconciled first | `a3-b2/prebatch.txt` |
| B1 rerun | `f5da33e` | PASS: observer 43, helper 59, checker 35; skill package check exit 0 (134 lines) | `a3-b1r/` |
| B2 (staged) | `f5da33e` | PASS: exported helper 59, checker 35 and observer 43 OK; 34 verbatim files identical; adapted-file delta lines identical; gitleaks `[]`. P1–P7 pass apart from the recorded owner README literal exception. Manifest `47785266…` | `a3-b2/` |
| B3 | `f5da33e` stage | PASS: 75/75, including asynchronous review rendered in AGENTS.md and SUPERVISOR-HANDOFF.md | `a3-b3/` |
| Delivery | `f5da33e` | PASS (§16.4) | `delivery-apply-a3/` |
| B4 | delivered bundle (copy without VCS directory) | PASS: 11/11, installed observer 43 | `a3-b4/` |

**Provenance.** The full profile last ran on `51c51d4` (final status 0). It is retained only for the content it
tested and is not claimed for the later observer bytes. Review r2 requires no duplicate full run for these watcher
changes. The MR and staging disposition remain binding. No live notice was sent and no test process remains.

### 16.4 Direct delivery (owner rules)

- **Conflict check:** the published manifest equalled my last delivery `8c452087…`, every file matched, and there
  were no unlisted files.
- **Applied:** the handoff `SKILL.md`, `watch.py` and `test_watch.py`; the supervisor `SKILL.md`; and
  sdlc-management `SKILL.md`, `references/roles-lifecycle.md`, `references/validation-integration.md`,
  `templates/management/SUPERVISOR-HANDOFF.md.tmpl` and `templates/root/AGENTS.md.tmpl`; plus `MANIFEST.tsv`.
- **Bundle manifest:** before `8c452087a9042be1342b3c52c04d6412368f7bf15439a282c59da2258bb271d3`, after
  `47785266389631ccde74d308661e1595cd4347dc19acf829596f78538782606d`.
- The VCS digest `a3c87807…` and README `415d4f2f…` are unchanged. No sibling commit or push.
- The delivered tree equals the staging copy and passes the metadata-aware package and test checks.
- Ownership record `676e2b83…`.

### 16.5 Status, ACs and budget

- AC1–AC10 are met (§15.5).
- AC13 is met locally, now including C4/C5 (43 focused tests).
- AC11 is met locally (focused on final bytes, plus the historical full run); the MR `fast` and staging gate are
  pending.
- AC12 is this submission.
- **B3: attempt 3 of 3 used.** No budget remains. If this attempt does not satisfy the review, work stops for a
  Codex budget or scope decision.
- No known blocker. The observer stays disabled until integration, and the operator runtime was untouched.

## 17. Attempt 4 (B3 override): C5 same-inode truncation correction

**Override provenance.** [Implementation review r3](G053-implementation-codex-review-r3.md) (SHA-256
`62e17cd9bde6aa48052ca90ffab937efaf65cf88e133d66508a7f48b845f8e49`, staging `4adb887`; Codex handoff `329f164b`)
returned CORRECTION LOOP, with C1–C4 resolved and the bundle and governance passing. It recorded an **explicit
exceptional budget decision**:
- approving supervisor: Codex;
- previous budget: B3, three attempts, all used and preserved;
- new maximum: **four total attempts**;
- reason: demonstrable convergence, with one concrete recovery predicate and its test remaining;
- remaining work: this correction only.

The Goal records the exception (`9e880dbfd18207e1139171c4dea41e6c8354ef01e5349de1d8702ab4ebadcb0b`). Attempt 4 is
limited to the fix, its test, export and provenance, and this Result. There is no automatic further extension.

**Rebase.** An ordinary rebase onto `4adb887`, with no conflicts. Staging is an ancestor; only allowed paths
differ; Plan r2 and r1 are unchanged. The 40 inputs (review r3 added; Goal and PROGRESS updated) are re-bound.

**Correction of §16.** §16 claimed that truncation was covered. That held only for truncation below the committed
offset and for inode replacement. The same-inode case inside a pending oversized row was not covered: §16's
"rotation or truncation resets it" overstated it.

**Fix** (commit `6a167f2`):
- In `new_rows`' existing truncation-recovery branch, a same-inode file whose size has fallen below a pending
  oversized row's `scanned` position (but not below the committed offset) now drops the stale oversized state and
  re-reads from the committed row start.
- The bytes before that start are intact, so there is no history replay and no stall at the old scan position.
- Inode replacement and truncation below the committed offset still restart at 0, as before.
- Bounded reads, rotation handling and durable intent are unchanged.

| Path | SHA-256 |
|---|---|
| `.agents/skills/claude-session-handoff/scripts/watch.py` | `678a2d8c7246228aee55ce9cd70fd4a9cdbddb59143b0b33409317549486d8b0` |
| `.agents/skills/claude-session-handoff/scripts/test_watch.py` | `926962b2ed0658238d4aa1137763b05a4f82e1e881765ea278d6bcb992f7e930` |

No other source changed.

**New fake-boundary test.** `OversizedRows.test_same_inode_truncation_below_pending_oversized_scan_recovers`:
1. A partial oversized row is written and polled, so the scan passes the truncation point.
2. The file is truncated in place (same inode) to 2000 bytes past the row start, which leaves the size above the
   row start and below `scanned`; the truncated fragment is newline-terminated.
3. A complete normal milestone is appended.
4. Over up to four polls the milestone is delivered **exactly once**, the persisted offset equals EOF, and no
   oversized state remains.

Result: **PASS**. Under the previous code the observer waited at the stale scan position. A development run of
the first fix (reset to offset 0) failed the exactly-once assertion with 2 notices because it replayed a
never-processed history row. It was corrected to resume from the committed row start before any recorded batch.

**Focused validation** (after the pre-batch gate at `4adb887`):

| Batch | Candidate | Outcome | Evidence (`$EVID/…`) |
|---|---|---|---|
| B1 | `6a167f2` | PASS: observer **44** (43 earlier, plus the new test), helper 59, checker 35; skill package check exit 0 | `a4-b1/` |
| B2 (staged) | `6a167f2` | PASS: exported helper 59, checker 35 and observer 44 OK; 34 verbatim files identical; gitleaks `[]`. P1–P7 pass apart from the recorded owner README literal exception. Manifest `befb9a11…` | `a4-b2/` |
| Delivery and verify | `6a167f2` | PASS: package P1–P7 on a copy without VCS metadata; exported helper 59, observer 44 and checker 35 OK; delivered tree equals the stage | `delivery-apply-a4/` |

Adoption (B3, 75/75 at `f5da33e`) and relocation (B4, 11/11) are reused with provenance. Since then no
sdlc-management file or template changed, and the delivered-tree verification reran the observer tests from the
bundle. There was no full rerun or live queue call, and no test process remains.

**Direct delivery (owner rules).**
- Conflict check: the published manifest equalled my last delivery `47785266…`; files matched; no unlisted files.
- Applied: `skills/claude-session-handoff/scripts/watch.py`, `scripts/test_watch.py` and `MANIFEST.tsv`.
- Bundle manifest before: `47785266389631ccde74d308661e1595cd4347dc19acf829596f78538782606d`. After:
  `befb9a11d10185ecbca10429e3cb328a8091cc246d4d02e823f851179fa1fdad`.
- VCS digest `a3c87807…` and README `415d4f2f…` are unchanged. No sibling commit or push.
- Ownership record `5dee84e1…`.

**Status and budget.** AC13 is met locally with this correction (observer 44); the other ACs are as in §15.5 and
§16.5; the hosted MR `fast` and staging gates remain pending. **Budget: B3 override, 4 of 4 attempts used**; no
budget remains. The observer stays disabled until integration.

## 18. Doc-only note within attempt 4: supervisor interruption is not stop-all

**Authority.** The owner keeps the async-wait policy and adds an interruption note (relayed by Codex, handoff
`64ab8e27`; amendment `c65de410` → `fafaa640be451e499ab570e9a26f2cb1556fc1b87632a87002e3cb552608ec8f` on staging
`561d65d`; AGENTS.md `d548953d…`, PROGRESS `f5893d8d…`). This is documentation only, with no runtime change or new
runtime test. Review r3, the attempt-4 override and the C5 fix (`6a167f2`) are unchanged. The
`repository-supervisor` paragraph uses the same recorded doc-only companion-source exception.

**Rebase.** An ordinary rebase onto `561d65d`, with no conflicts. Staging is an ancestor; only allowed paths
differ; Plan r2 and r1 are unchanged. The 40 inputs are re-bound.

**Change** (commit `d4f3f3d`). An interruption of the supervisor's turn alone is not stop-all authority. On
resuming, inspect the children's actual status, keep completed findings and authorized running work, and do not
cancel or re-dispatch automatically. An explicit stop-all, or a conflicting scope, security or owner direction,
governs. Never assume a client (for example mobile) cancellation reached child agents. The note was added next to
the async-review rule in these files:

| Path | SHA-256 |
|---|---|
| `.agents/skills/repository-supervisor/SKILL.md` | `1f5ad3e963218aa15789087ef0c5071924f9a3b22d3c70f5d47b4cef479a93f0` |
| `skills/sdlc-management/SKILL.md` (136 lines) | `2ebac11617213be9ffef8886ba609ac651dc855eb71c9d3c2cbd25fd781371b7` |
| `skills/sdlc-management/references/roles-lifecycle.md` | `672f841194d3e405fb0d3ff6cd947d5db778cbf98259ecf2d0469169d0f067ec` |
| `skills/sdlc-management/templates/management/SUPERVISOR-HANDOFF.md.tmpl` | `85dc725c3eb01ce57642a2955213a46f58633f76e5073e33f1f0847571b722db` |
| `skills/sdlc-management/templates/root/AGENTS.md.tmpl` | `82d5adcb7700fd58cb02b07470b1ce2e6a652c69b246f7db380b53870ba8ea0a` |

The supervisor export's invariants delta moved mechanically from L95–96 to L100–101; the adapted diff lines are
otherwise unchanged.

**Focused doc and template checks** (after the pre-batch gate at `561d65d`):
- B1: skill package check exit 0, with checker 35, helper 59 and observer 44 unchanged.
- B2 staged export: P1–P7 pass apart from the recorded owner README literal exception; verbatim files identical;
  gitleaks `[]`; manifest `5241f449…`.
- B3: 75/75. The existing async-review template check now also asserts the interruption wording.
- The delivered tree, verified without VCS metadata, passes package P1–P7 and the exported helper 59, observer 44
  and checker 35 tests, and equals the stage.

**Direct delivery.**
- Conflict check: the published manifest equalled my last delivery `befb9a11…`; files matched; no unlisted files.
- Applied: the supervisor `SKILL.md`; sdlc-management `SKILL.md`, `references/roles-lifecycle.md`,
  `templates/management/SUPERVISOR-HANDOFF.md.tmpl` and `templates/root/AGENTS.md.tmpl`; and `MANIFEST.tsv`.
- Bundle manifest before: `befb9a11d10185ecbca10429e3cb328a8091cc246d4d02e823f851179fa1fdad`. After:
  `5241f4491b457b7bada91a0015318643b40a96672b1294dc9b20718176983f6b`.
- VCS digest `a3c87807…` and README `415d4f2f…` are unchanged. No sibling commit or push.
- Ownership record `0e1642a5…`.

**Budget.** Unchanged: B3 override, 4 of 4 used (this doc note is inside attempt 4). Hosted gates are pending; the
observer stays disabled until integration; no test process is running.

## 19. Attempt 5 of 5 (B5): Duo findings C6 and C7

**Authority.** The [Duo review record](G053-mr-duo-codex-review.md) (SHA-256
`14f0a225cdb0b6cc2985779bafd861f627a8f0d6954ca0de5a5502a8a94fccf7`, staging `3e50932`) returned CORRECTION LOOP for
MR !21. The owner said "Fix these", relayed by Codex.
- Codex set **B5: at most five Goal-level attempts**, replacing the earlier maximum-four override. Four were used
  and are preserved.
- Attempt 5 is limited to C6/C7, their tests, source/export/manifest provenance, this Result and the existing MR
  update.
- The Goal is at `08705260414f69e1827a17a69d0d9e51cd9b649d86e23ebde065f3e79d43451d`.

**Identity and lock reconciliation.** The S0 re-check found the worktree's harness lock line and `locked` file gone
(removed at about 16:58 local). PID 74419 was the same process with the same start time. I stopped without writing
and reported it.
- The coordinator verified read-only that PID 74419 is the coordinator session's own process: the harness drops
  the run lock when an agent stops and does not restore it on resume. The same happened to the retained planner
  worktree.
- The coordinator also confirmed the native meta (path, branch, effort high), and that no other process has its cwd
  in this worktree.
- It recorded an addendum: the dispatch record is now SHA-256
  `4dfd7b4b46382093b48f7107ee966de3d8534a800e4d53d847b974c9d4f7e6f4` (was `4b5b3250…`).
- Under the new S0 rule for resumed runs, I re-verified HEAD `80e6efa`, a clean tree, the Git registration and
  branch, the record hash, and that no other process uses the worktree. No lock was added, removed or changed.

**Reconcile SHAs.**
- Prior delivery SHA `80e6efaf898e9e4070eb1f1b6add3363fd2cf6a4`.
- `git range-diff 561d65d..80e6efa c7b8f42..a29888d8`: patches 2–21 are `=`. Patch 1 shows `!` only because the
  staging-imported Result it was based on differed (`95a6e762…` versus `e8a092a6…`). Both versions of commit 1 yield
  identical Result bytes (`cfdbe4d6…`).
- `git diff 80e6efa a29888d8` lists staging management paths only.
- `git reset --keep a29888d8f071b43596403a659727dc0dd8b7c8a7` (MR !21 head) was accepted.
- A re-fetch followed by `git rebase origin/staging` was conflict-free onto `3e509327d778d6137ec4f13e705e61fd0408bea5`,
  giving `350f038`. `3e50932` is an ancestor, and the task paths are byte-identical to `a29888d8`.
- The C6/C7 commit is `c8b50f1`; HEAD is the Result commit after it.
- Inputs were re-bound (41 rows): Goal `08705260…`, PROGRESS `d1691c0c…`, ROADMAP `b92b6ae4…`, Duo review
  `14f0a225…`.

**Changes** (commit `c8b50f1`):
- **C6.** `new_rows` now splits complete chunks only on the JSONL LF byte (`split(b"\n")`, then decodes each row).
  U+2028, U+2029 and U+0085 inside JSON strings are content, not delimiters. The milestone excerpt split now also uses
  `"\n"`. Partial tails, oversized rows, rotation, truncation and offsets are unchanged. No existing test relied on
  CR splitting, and a trailing CR remains valid JSON whitespace.
- **C7.** Binding now separates an omitted threshold (`None`, which gives the 100 MiB default) from an explicit
  value. An explicit `0` fails as `size-threshold-too-small` before any binding change; bind never queues. Valid
  overrides and `--no-size-advisory` behave as before.

| Path | SHA-256 |
|---|---|
| `.agents/skills/claude-session-handoff/scripts/watch.py` | `af11be95ca40fc31d77ced165731a2ec301116ca74d268cc92fff43a50ed18cf` |
| `.agents/skills/claude-session-handoff/scripts/test_watch.py` | `31547d54dc597ecb1b3a463f0a69907d7baf0d8b04305491f944c463eb39b7b1` |

No other source changed.

**New tests.**
- `UnicodeSeparatorFraming.test_raw_unicode_separators_do_not_split_rows` has three subtests: raw U+2028, U+2029
  and U+0085 inside a genuine milestone row, followed by another milestone row. Both are delivered exactly once,
  the persisted offset is EOF, no oversized state remains and there are no false `parse_limit` entries.
- `SizeAdvisory.test_explicit_zero_threshold_is_refused_without_binding_change`: with no prior binding, nothing is
  created; with a prior binding, its bytes are unchanged; zero Codex calls.
- `SizeAdvisory.test_omitted_valid_and_opt_out_thresholds`: omitted gives 104857600, explicit 2048 is honoured, and
  opt-out gives none.
- Against the pre-fix `watch.py` (a copy in `$EVID/a5-oldcode`), the new tests fail (U+2028, U+2029 and U+0085
  subtests, plus explicit zero). With the fix they pass. The existing framing negatives (partial, oversized,
  truncation, rotation) are unchanged and pass.

**Commands and outcomes.** Every batch ran after the pre-batch gate: fetch, ancestry at `3e50932`, mounts 5%, and
bindings.

| Step | Outcome |
|---|---|
| dev run, uncommitted edits, after a gate showing `BINDINGS-CHANGED` for only my two edited files | observer 47 OK (development only) |
| old-code proof (`$EVID/a5-oldcode`) | 4 expected failures before the fix |
| B1 `bash b1.sh a5-b1` (`c8b50f1`) | PASS: observer **47**, helper 59, checker 35, skill package check exit 0 |
| B2 `bash b2.sh a5-b2` (staged export) | PASS: exported helper 59, checker 35, observer 47; 34 verbatim files identical; gitleaks `[]` (pipe `0 0 0`); manifest `58655517…`. P6 reports only the owner/Codex-authored README exceptions: the published repository link ("hosted-remote") and "Vestive". The export README source was aligned to the published README `8ae08b74…` |
| sibling conflict check, first run | stopped by my own script: tag 1.0.0 resolved to the annotated tag object `e823c411…`. Peeled read-only from the loose object, it points to `fdc439b94c1169592ddd89de83d0d8a3d22a6548`. Script corrected to peel the tag; nothing was written |
| sibling delivery `apply_two_files.py` | applied (below) |
| delivered verification (copy without VCS metadata) | package P1–P5 and P7 pass; P6 shows only the README exceptions; helper 59, observer 47 and checker 35 OK; delivered tree equals the stage |

**Retained evidence, with provenance.**
- Adoption 75/75 (`a4d-b3`, at `d4f3f3d`, before rebase): no sdlc-management file changed since.
- Relocation 11/11 (`a4-b4`/`fin-b4` lineage): watcher behaviour is re-checked by the delivered-tree observer tests.
- Full profile on `51c51d4` (status 0): historical, and not claimed for the later bytes.
- No duplicate full run. No live queue call. No test process remains (`ps` shows none). The observer stays off;
  the operator runtime and target were not touched.

**Sibling delivery** (`../restruct-sdlc`, record `$EVID/a5-sibling/apply.json`, SHA-256 `e38be472…`).
- Baseline per the coordinator: HEAD `80f2166` = origin/main (owner-authorized Codex README commit), tag 1.0.0 peeling
  to `fdc439b`, MANIFEST `5084d7f3949f0bdeb473081c4be5ad067356707d0ffdf378af0475a7bd7df006`, README
  `8ae08b748fedbc02450dc4f2d24fd74056585deb69e7341ab3c5d1a56322dedb`, `watch.py` `678a2d8c…`, `test_watch.py`
  `926962b2…`. All 42 rows matched and there were no unlisted files.
- Applied as **uncommitted working-tree changes**: the two files, plus the hash cells of their two MANIFEST rows.
  Every other row, including README `8ae08b74…`, is byte-identical.
- MANIFEST after: `586555173a1d474d3ac2a0ac83ab7a909833b56fcb9c61baf83422f772a56f8f`, equal to the stage manifest.
  `watch.py` after `af11be95…`, `test_watch.py` after `31547d54…`.
- VCS digest `c0e7ad9d…` is identical before and after; README unchanged; tag unchanged. No commit, tag, push,
  remote or VCS change. Ownership record `4c6e5a82…`.

**Budget.** B5: **5 of 5 attempts used**. No further attempt remains. The Duo discussions are left unresolved for
Codex. Hosted MR `fast` and the staging gate are pending; the coordinator lease-pushes the MR update.
