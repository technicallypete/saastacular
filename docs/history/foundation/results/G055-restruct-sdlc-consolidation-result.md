> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G055 Result — Single-skill Restruct SDLC consolidation

> Historical implementation handback follows. Its interim status is superseded by the final acceptance and cleanup record in [G055 final review](G055-final-codex-review-20261020.md).

## Final disposition — 2026-10-20

Codex **FINAL ACCEPTS and closes G055** under the owner's task-specific qualification exception. The exact merged-SHA full run remains failed on B-DEPSCAN and its derived B-REQUIRED-CHECKS; see [the retained failure inventory](G055-fullrun-disposition-20261020.md) and [final review/cleanup receipt](G055-final-codex-review-20261020.md). The independent certifi maintenance proceeds on Beginify staging and does not reopen G055 or change the published Restruct `2.0.0` release.

## Status

**In progress; attempt 1. Slices 1, 2a, 2b and correction 01 reviewed; V1–V4, the compat run, V3 classes and the
governance step done (below, "Slices 3–6"); V5 stage, M1 and F1 remain. Not Codex-accepted, not ready for F1/MR.**
Budget B5, checkpoint after attempt 3; attempt 1 running (no reset). G055 merge held until G058 final Goal
ACCEPT (owner hold, staging `a608145d`/`30890b16`). No live cutover, install, export, publication or
settings change. G054 deferred and G020 held: untouched.
Update 2026-10-18: correction 02 (P6/A2) and V5 stage-only (outcome S) recorded below; M1 and F1 remain.
Update 2026-10-18 (final writer handoff): correction 03 (P6 continuation residual) and the V5 stage refresh recorded
below; source work complete for Codex review. Pending items are listed under "Transfer state" (Codex link
reconciliation, owner README wording, M1, F1, live gates G8a/G8b, merge hold).
Update 2026-10-19 (Plan r3 resume, attempt 1 continues): S0′ rebind onto staging `c29ff777`, S1′ Q1–Q12 fold
plus the owner RC auto-merge and optional RC Duo amendments, focused checks and the V5 stage refresh recorded below
("Plan r3 resume"). Still pending: Codex links, owner README wording, M1, F1, G8a/G8b, merge hold.

- Writer: single G055 implementation writer, task G055, role implement, native isolated worktree
  `.claude/worktrees/agent-<id>` (generated name; one writer verified by branch/HEAD assertion). Retired
  after slice 1; replacement single writer from correction 01: `.claude/worktrees/agent-<id>`; fresh
  single writer from V1 (checkpoint handoff): `.claude/worktrees/agent-<id>`.
- Branch `g055-restruct-sdlc-consolidation`, created from origin/staging `db236f03`, fast-forwarded before any
  write to `30890b161818ce47db5d12968c3a1f81c71b452b` (two coordinator-relayed owner deltas, §S0.3).
- Model/effort: dispatched Opus 5.5 high; runtime value not verifiable inside the worker (unknown).
- Private evidence: `$HOME/.local/state/platform-evidence/g055-impl-a1/` (0700); secrets
  `$HOME/.local/state/platform-secrets/g055-impl-a1/` (0700, empty). Repo `.local` stays suspended.

## Authority and approved package

Correction01 ACCEPT and isolated execution release (not included)
plus its open-decision dispositions (OD1–OD4) and the later owner integration hold. Package = Plan r1
(`7ac56da5…24bf1`) + r2 addendum (`b1c52ae4…fa4d1`) + correction01 (`6d150552…48e1`) + that disposition.
Inputs `G055-planning-inputs-r2c1.json` (`ebff5890…c6c4`); scope `G055-scope-allowlist.txt` (27 patterns,
`f3f3ad3e…8895`).

## S0 — exact binding (before any write)

### S0.1 Inventory verification at db236f03

`sha256sum -c` of all 56 hashed r2c1 rows (`s0-expected.sha256`): **53 OK, 3 drift**, all status-only and
exactly as the release review predicted: `management/v0/PROGRESS.md`, `management/SUPERVISOR-HANDOFF.md`,
`management/v0/results/ONEOFF-vestive-shared-root-codex-review-20261018.md` (G055 release row, Vestive
Duo/owner-hold text; `git diff d914440b db236f03`). Null rows confirmed absent: `.claude/skills/restruct-sdlc`,
`.agents/skills/claude-session-handoff/scripts/test_claude_only.py`. The r2 review itself grew by its appended
release section (`b07d902f…` 60 lines → `a501d17c…` 96 lines at db236f03; first 60 lines are the bound text).
Binding release/correction/inputs/scope hashes matched the dispatch exactly.

### S0.2 Status reconciliation

Codex-owned status files were not edited; newer supervisor text is the current state. The one active legacy
link in `management/SUPERVISOR-HANDOFF.md` (L22, `../skills/sdlc-management/references/coordination.md#enforce-existing-usage-limits`)
is unchanged and is retargeted only at the safe merge point under correction01 §C3 (later slice).

### S0.3 Coordinator-relayed owner deltas (applied before writes; no replan/restart)

Staging advanced `db236f03 → 8798f701 → 30890b16`; the empty branch was fast-forwarded (`git merge --ff-only`)
and the inventory re-verified (`s0-sha256-check.log`, SHA-256 `2ea4085a…9931`): 50 OK, 6 drift:

| Path | Bound (r2c1) | Now (30890b16) | Disposition |
|---|---|---|---|
| management/v0/goals/G055-restruct-sdlc-consolidation.md | `4d680824…cb21` | `25ee7d3aa221393316000883a5fd81579e66fc61445e165392cdf8f5b15c6b32` | **rebound**; appended "Event-driven status updates" section (owner direction); carried below |
| management/v0/specs/G055-restruct-sdlc-consolidation-spec.md | `1ba9829c…9dc2` | `fcd97476a2e648a410b0c0f8c51f010e837525fa11ac437be3ad1bca050079c7` | **rebound**; same appended section |
| skills/sdlc-management/references/coordination.md | `455f56e0…9750` | `4c07434924f6ac8315de56a74901bd7741e283746f8747ba4cebda1f447f3306` | owner "Update frequency" subsection (+15 lines); moved byte-for-byte to `skills/restruct-sdlc/references/coordination.md` (same hash) |
| management/v0/PROGRESS.md | `7c2b69b3…` | `3535dbbddd9f6874c3cbbaceda84e1081df5faafd143e8060d78e4fe3637afd2` | status only |
| management/SUPERVISOR-HANDOFF.md | `2c7984ef…` | `edc8d9a2…3925` | status only (G058 row, G055 owner merge hold); L22 link unchanged |
| ONEOFF-vestive-shared-root review | `00a5ea47…` | `7829e396…d385` | status only (MR !48 still owner-held, not integrated) |

Also new: `G055-plan-r2-codex-review-20261018.md` now `5b42a6f0…3d54` (appended owner integration hold),
`PROCESS-status-prose-20261018.md` `fc9350a7…e373`. No helper/observer/checker/template/CI/settings input
drifted. Event-driven status policy: the canonical coordination reference carries it now (moved); the generated
guidance/templates carry it with the governance step (Plan §9 step 5).

### S0.4 Baseline P2 (`fail-baseline-P2`, history)

`python3 -B -m unittest discover -s skills/sdlc-management/scripts -p test_check_sdlc.py -k test_real_skill_passes_all_checks -v`
on untouched db236f03: exit 1, Ran 1, FAILED (failures=1); sole finding `P2 … SKILL.md 218 lines exceeds 200`.
Log `s0-p2-baseline.log` SHA-256 `a3d34e4d7fd2b8b49251a934c6ec6962ba291d4559de3f626ba3084b8184f7f6`. Recorded as
expected baseline failure; not fixed by trimming the retired file; the canonical entrypoint must pass P2 at V4.

### S0.5 Sibling (read-only; no git there — worktree guard)

| Item | Value |
|---|---|
| Public remote policy tip (Codex-reported, latest) | `f16d8af6be1c8a1af677e7b0ae55455767ba6b03`, parent `f6ed5995` (loose object read read-only: "Reduce routine status reporting to meaningful changes") |
| Earlier Codex-reported tip | `f6ed5995f3a88a71c11892020ac9901c0e07c97a`, parent `568721b0` |
| Local sibling `refs/heads/main` (HEAD) | `cb998602e10767507c96dec57ef7e7f6319230f7` — 12 commits behind `f16d8af6` (chain read from loose objects) |
| Local `refs/remotes/origin/main` | `c3148cf0c52a1aae2384eb4ac36b7ed998818459` (stale tracking ref) |
| Local `FETCH_HEAD` main | `568721b0` at first read, `f6ed5995` at slice end (fetched by another actor; not by this writer) |
| Tags | `1.0.0`, `1.0.1` |
| Working tree | `MANIFEST.tsv` 43 rows `99f153fb…1b51`, `README.md` `ec39ef94…`, `LICENSE` `af5fb221…` (all = r2c1); dirty G054 overlay present; `status --porcelain` not run (guard) |
| `.git/index` | `6c3444d8268a08a8061a1a98627ff7b1802867bfed79682689f90a6695c5f9b0`, identical at S0 and slice end |

OD2 outcome **S (stage-only)**; nothing written to the sibling/public repository.

### S0.6 Runtime, observer, storage

- Tools: Claude Code 2.1.302, codex-cli 0.160.0, Python 3.12.3, Git 2.43.0.
- Live observer (`ps`, read-only; no state/binding read): one process, PID 3867, started 2026-10-17 08:03:32,
  `python3 -B <main>/.agents/skills/claude-session-handoff/scripts/watch.py --repo <main> run --interval 30`.
  Identical at slice end. Not stopped, rebound or replaced (main checkout untouched).
- OD3 one-line resolution: isolated workers are natively refused main-checkout writes, so Beginify keeps the
  home override; the portable repo-local `.local/` default is documented with that limitation (installation.md, later).

## Slice 1 — canonical package relocation, legacy forwarders, repository-local trust chain

Earliest risky migration/trust slice (Plan r1 §9 steps 1–2, repository-local part only).

### Implementation summary

1. `git mv skills/sdlc-management → skills/restruct-sdlc` (whole tree, history preserved; references incl. the
   new Update-frequency subsection byte-identical; `SKILL.md` moved unchanged — rewrite is the next slice).
2. `git mv` `.agents/skills/claude-session-handoff/scripts/{handoff,watch,test_handoff,test_watch}.py →
   skills/restruct-sdlc/scripts/`; `watch.py` byte-unchanged (`6a00ee02…`).
3. Forwarders: `skills/restruct-sdlc/compat/legacy/claude-session-handoff/scripts/{handoff,watch}.py` and the
   installed copies at `.agents/skills/claude-session-handoff/scripts/{handoff,watch}.py` — one 16-line stdlib
   source (SHA-256 `30dea4085d11bd684ce0f9e6f8a1eab5cada76e1bcc7442aeb84bba64d26c506`, four identical files).
   Root = real path's Git root only when the real path ends in the legacy layout; target = canonical
   `skills/restruct-sdlc/scripts/<name>`, else flat `.agents/skills/restruct-sdlc/scripts/<name>`, each a regular
   file with no symlink component; `os.execv` of the same interpreter (keeps `-B`) with identical argv; otherwise
   exit 2 `legacy-forwarder-target-missing`. No home/other-checkout fallback, no argument parsing.
4. `handoff.py` (`c6e3319e…2a93`), enumerated sites only: `HELPER_RELATIVE` → canonical, `FLAT_HELPER_RELATIVE`,
   `helper_relative()` selects the regular blob in `HEAD` (`ls-tree` mode 100644/100755; an alias path is never a
   blob); `helper_dirs` adds per-worktree canonical/flat/both-alias/legacy-forwarder dirs and the three home roots
   (resolved); `helper_argument` also matches `watch.py`; `prerequisite_evidence` compares self/committed/working/
   writer on the selected canonical file (working must resolve to itself), adds `executing-path-outside-canonical`
   (executing file ≠ the selected repository's canonical file) and `legacy-forwarder-modified` (each forwarder in
   any worktree must equal its HEAD blob and the package compat bytes; symlinked forwarder refused). Evidence adds
   `helper_relative`, `legacy_forwarders_modified`. No check removed or weakened.
5. Not in this slice (deferred, unchanged design): home binding/`--repo` refusal set, both discovery aliases (staged
   outside discovery; root `.claude/skills` would be listed by live sessions scanning worktrees), legacy
   `SKILL.md`/references retirement, entrypoint/`handoff.md`/`supervision.md`/`installation.md`, checker formats,
   governance links, SUPERVISOR-HANDOFF retarget, export.

Intermediate state (expected, WIP branch; branch pushes run no pipeline per `.gitlab-ci.yml` workflow rules):
active governance links to `skills/sdlc-management/…` dangle until the governance step; the legacy
`claude-session-handoff` SKILL.md/references remain and work through the forwarders.

### Tests added/extended

- `test_handoff.py` (`cea7cf0c…3c09`): fixture `make_repo` executes as the first fixture repository's canonical
  copy (same bytes; `SELF_PATH` patched); **extend** T14 with 5 cases: executing path outside canonical; uncommitted
  or symlinked canonical; flat real install accepted and both aliases never a committed helper; equal forwarders
  accepted, modified/symlinked/linked-worktree-stale forwarders refused (log and archive unchanged via `refused()`);
  `watch.py` and forwarder-directory invocations refuse.
- `test_consolidation.py` (`b4800754…137e`) **new**, class `TestInstallMigrate` (forwarder part only, 4 cases):
  identical/≤20-line/stdlib-only sources; legacy command parity (exit/stdout/stderr/argv/`sys.path[0]` import/`-B`)
  for both scripts; flat fallback; refusals (no target, symlinked source-maintainer and flat paths, stray copy,
  non-Git root) all exit 2 with the exact message.

### Commands and outcomes (host `python3 -B`, temp HOME/dirs, no lock, no network)

Pre-batch: `git fetch origin staging` → `30890b16`, ancestor of HEAD (exit 0).

| Command (all `python3 -B -m unittest discover -s skills/restruct-sdlc/scripts …`) | Exit | Count | Log SHA-256 |
|---|---|---|---|
| `-p test_handoff.py -k test_t14_ -k test_t15_ -v` (run1) | 0 | Ran 26, OK | `57fd649f19267bf816cfeca086644d178d801a357f28588e62063c2f4385588a` |
| `-p test_consolidation.py -k TestInstallMigrate -v` (run1) | 1 | Ran 4, 1 FAIL | `47b6eaf85d344056622a97336e24c60e432daca44b4e8777610f0200d883e1a9` |
| same (run2, after forwarder fix) | 0 | Ran 4, OK | `11ebab5f5c3a0a5559f0bf451953b79c1270afaaf67736725ba528957b5f6048` |
| `-p test_handoff.py -k test_t14_ -k test_t15_ -v` (run2, forwarder bytes changed) | 0 | Ran 26, OK | `200beefa2de455da1501852d4083cbddb6eb702f9b0c4935e478430379a74a1a` |
| `-p test_watch.py -k test_start_ -k test_stop_ -k test_fresh_start_ -k test_known_unsent -k KnownUnsent -v` | 0 | Ran 12, OK | `1c12bc3280018f05a5fe5f7351435ce72bed261dc19649c2b49a3c265e34156b` |
| `-p test_handoff.py -v` (whole file once: shared compaction prerequisite callers T09–T15) | 0 | Ran 69 (64+5), OK | `6adf96f1ee71a08345044c57e85592a869e38175f48b8117a2382b851b97462f` |

Run1 failure (real defect, kept as history): a forwarder copy outside the legacy layout raised `IndexError` at
`HERE.parents[4]` (exit 1, still non-executing) instead of exit 2. Fix: compute the root only when the real path
ends in the legacy layout. Note: `test_t14_real_process_holding_the_log_refuses` (existing) reads real `/proc`
entries read-only; the live observer is outside the fixture's helper directories; no observer state touched.

Not run (by design, later batches): `test_check_sdlc.py` (V4; its fixtures still pin the legacy names, r2 §6 row 2),
whole `test_watch.py`, compat run, V3 remainder, V5, M1, F1.

### Scope and protection

`git diff --cached --name-status -M 30890b16` (`slice1-name-status.txt`, `4904ef0cb5a936e96582ff19d0ea2c06c49d57489f465c01a4b0775cae324c34`): 38 entries, 69 unique
paths, **0 outside the 27 patterns**. Protection diff over `.gitlab-ci.yml infra core web intelligence contracts
skills/saas-foundation .agents/skills/vestive-design .agents/skills/repository-supervisor management .gitignore
.claude README.md agents AGENTS.md CLAUDE.md docs`: empty (exit 0).

## Correction 01 — forwarder genuine Git root (slice-1 review CORRECTION LOOP)

Authority: slice-1 review (not included) (SHA-256 `e90e1ebf…367a`): `.git` existence is not
proof of a genuine checkout root. Fresh writer (predecessor retired), native isolated worktree
`.claude/worktrees/agent-<id>`; branch/HEAD asserted `ccba8499` before any write. Attempt 1/B5/checkpoint 3
unchanged; scope now `results/G055-scope-allowlist.txt` (28 patterns, `e7c2bebc…b64d`).

Reconciliation: origin/staging `7df1cd1a` (not an ancestor of `ccba8499`; merge base `30890b16`) → rebased the one slice
commit (`ccba8499` → `8d5c0ba3`, clean; the Result was byte-identical on both sides). Staging delta 9 paths; bound-input
drift status-only: ROADMAP `2645b8b8…9320` (G020 Goal/Spec authority note), PROGRESS `cbb4040e…e438`, SUPERVISOR-HANDOFF
`be2997fd…9ff6` (status rows; the active legacy link moved L22→L23 byte-unchanged). Unbound: ARCHITECTURE (label privacy),
G020 Goal/Spec, slice-1 review, scope28. No helper/observer/checker/template/CI/settings drift.

Change (forwarder SHA-256 `30dea408…c506` → `2a664bc77e969f47e5cfb7a3aae3c145a37aeab735afc2cabb9d654c20e18386`, four
copies byte-identical, 20 lines, stdlib `os`/`subprocess`/`sys`/`pathlib`): before forwarding, run
`git -C ROOT rev-parse --show-toplevel --path-format=absolute --git-dir` with every `GIT_*` variable removed (discovery
from ROOT only), then require Git's toplevel == ROOT **and** Git's git dir == `ROOT/.git` (main checkout) or the linked
worktree git dir's `gitdir` back-pointer == `ROOT/.git`. Any Git failure, missing `git`, unreadable back-pointer or
unexpected output → exit 2 `legacy-forwarder-target-missing`. Probe (disposable, `c1/probe_gitroot.py`): Git itself
reports a fake gitfile pointing at the main git dir or at another worktree's git dir as toplevel = the fake directory, so
the back-pointer check is required in addition to the toplevel match. Unchanged: legacy-layout containment, canonical then
flat target, symlink-component refusal, same interpreter/`-B`/argv `execv`, no home/other-checkout fallback, no argument
parsing. Fail-closed residual: a main checkout whose `.git` is a gitfile to a separate git dir (`--separate-git-dir`) is
refused (not a supported layout).

Tests (`test_consolidation.py` **extend** `TestInstallMigrate`): import-set assertion updated to include `subprocess`;
new `test_dot_git_existence_is_not_a_checkout_root` — positive genuine linked worktree (`git worktree add`, gitfile) forwards
to its own canonical; refusals for fake `.git` files inside a genuine repository (gitfile → main git dir, gitfile → linked
worktree git dir, garbage) and fake `.git` dirs (empty, inside the repository and outside any repository), both scripts.
Existing main-checkout parity, flat fallback and refusal cases unchanged.

| Command (`python3 -B -m unittest discover -s skills/restruct-sdlc/scripts …`) | Exit | Count | Log SHA-256 |
|---|---|---|---|
| `-p test_consolidation.py -k TestInstallMigrate -v` (corrected forwarder) | 0 | Ran 5, OK | `b5e3182c8ce881c5bdb781e44a0620f4cdf96527f8dfa1c7ee2637a0b59cecd5` |
| `-p test_consolidation.py -k test_dot_ -v` negative control: new test against the pre-correction forwarder bytes (temporarily restored, then re-restored to `2a664bc7…`) | 1 | Ran 1, FAIL (fake gitfile → main git dir executed the neighbouring script, exit 3) | `5d2a119e389f96336baf3840c9af1eb4fcb48f10c32874d6f8d2e1ae38e2a227` |
| `-p test_handoff.py -k test_t14_ -k test_t15_ -v` (forwarder bytes changed) | 0 | Ran 26, OK | `654f58bbd0b9b6ef7c498b5d3beb45fda9ff2728205a09a1bd2774039fde4a9c` |

Logs under `$HOME/.local/state/platform-evidence/g055-impl-a1/c1/`. Pre-batch: fetched staging `7df1cd1a`, ancestor of HEAD.
Slice-1 evidence (incl. the run1 failure) retained; observer-lifecycle and whole-`test_handoff.py` runs not repeated
(`watch.py`/`handoff.py` bytes unchanged by this correction). Live observer PID 3867 unchanged (`ps` only).

## Slice 2a — single entrypoint and module references (Plan §5.1, §9 step 1; no retirement, no alias)

Pre-batch: staging `7df1cd1a` unchanged and ancestor of HEAD.

| File | Change | SHA-256 |
|---|---|---|
| `skills/restruct-sdlc/SKILL.md` | rewritten fresh as routing: `name: restruct-sdlc`, version 2.0.0; boundaries, capability table, task-to-resource table, installation summary, precedence/composition, limits, verification, policy pointers; **118 lines** (P2 baseline was 218) | `fd995fa0f5ba91a6ed13600811be830f7a594715a76fb6fe8a30cc8da5f04e4c` |
| `references/handoff.md` | `claude-session-handoff` SKILL.md body without frontmatter; mechanical only: 10 management links → same-directory references, 3 runbook/script links → `local-session-handoff.md`/`../scripts/*`, 7 command paths `.agents/skills/claude-session-handoff/scripts/` → `.agents/skills/restruct-sdlc/scripts/`; private-state paths unchanged | `34f4818f7707c2cadd31181962e553845dd63dd6c02b9e77ee0dc0631170b892` |
| `references/supervision.md` | `repository-supervisor` SKILL.md body without frontmatter; 14 management links → same-directory, 1 → `handoff.md`; composition name `sdlc-management` → `restruct-sdlc` | `d8a21e282c87cae3cf01640d957981d19f4881c377a2f07d9783a89ed3cbbb19` |
| `references/installation.md` | new: layouts/aliases, discovery evidence rule, install loop, selective adoption, executing-code binding (repository + home refusal set), migration order, legacy commands incl. the correction-01 Git-root rule, cutover and **C1 rollback order (7 steps)**, shared home roots, templates table (moved from the old entrypoint), distribution route A; supersedes adoption.md's two three-skill install sections | `08212d17ba870d757a170d5fb40743bdc3b0f227ed2375d64054a1977022ca09` |
| `references/local-session-handoff.md` | byte copy of the legacy runbook (legacy copy removed at retirement) | `58620c03…a8d9` |
| `references/{coordination,roles-lifecycle,closeout-risks}.md` | intra-package link prefixes only: `../../claude-session-handoff/SKILL.md` → `handoff.md` (4), `../../repository-supervisor/SKILL.md` → `supervision.md` (1) | `e6fd270f…b6d7`, `d2ca6ae4…c10a`, `d6964f11…ef9` |
| `agents/openai.yaml` | one UI metadata: "Restruct SDLC" | `89af81da…515a` |

handoff.md/supervision.md are derived copies while the legacy packages still exist (deriver
`s2/derive_modules.py`, ledger `s2/derive-ledger.tsv`); the legacy bodies are deleted in the retirement commit.

**Moved-sentence ledger (old entrypoint → destination; nothing dropped).** Intro/no-authority → SKILL.md intro.
Boundaries table → Capabilities table. When-to-use/Not-for → frontmatter + "Not for" line. Native worker naming →
coordination.md#naming-dispatch-and-relocation-gate. Inputs → adoption.md#parameters + templates table. Authority
precedence → SKILL.md verbatim (+ relayed-authorization sentence from the lifecycle summary). Adoption summary →
adoption.md#procedure; three-skill install → installation.md#install-loop. Lifecycle summary: flow, delegated
review/relay, async wait, interruption → roles-lifecycle.md; `SUPERVISOR_AUTO_PUBLISH` → roles-lifecycle.md L241 +
adoption.md#parameters; `OBSERVER_AUTO_ACTIVATE` → adoption.md#parameters/procedure step 7; effort by role →
budgets.md#effort; mechanical corrections, mandates/stops, automatic planning, execution release → their
roles-lifecycle anchors; evidence/bounded cleanup → closeout-risks anchors; Git-first proof →
validation-integration.md. Validation paragraph → validation-integration.md. PROGRESS compaction →
closeout-risks.md#progress-compaction. References table → Task-to-resource table. Templates table →
installation.md#templates. Supported limits → rewritten for one package (old "native `.claude/skills` discovery and
personal installation unsupported" superseded by installation.md layouts). Package verification → Verification +
adoption.md#procedure. Model-aware effort, Fable fallback, mandatory cleanup, private files → budgets.md /
closeout-risks.md / adoption.md anchors (also present in handoff.md and supervision.md). Operational defaults and
targeted reruns → validation-integration.md, roles-lifecycle.md, coordination.md, closeout-risks.md anchors.

V4 focused (P1–P4 + line count), runner `s2/v4_p1_p4.py` calling the unmodified checker functions on
`skills/restruct-sdlc` (the checker still gates P2/P4 on the legacy management name until the checker-format slice):

| Command | Exit | Result | Log SHA-256 |
|---|---|---|---|
| `python3 -B …/s2/v4_p1_p4.py skills/restruct-sdlc` | 0 | P1 P2 P3 P4 pass, 0 findings, SKILL.md 118 lines | `cd4257d79a5cbc4ce65a4d6bc5a29c54447aa2e83f45b3456f34ee6bcd8546c8` |

Findings for later slices (not fixed here): `check_sdlc.py package skills/restruct-sdlc` also reports **P6
non-synthetic full UUID** in `scripts/test_handoff.py` L10–12 (three fixture constants inherited from the moved
legacy test; the legacy skill was never package-checked) — checker/test-format slice. Templates still link
`.agents/skills/sdlc-management|claude-session-handoff/...` (P3t; `SKILL_LOCATION` rendering, governance slice).
Reference text still names the old skills in prose (`adoption.md` L191/L201/L228/L262, `coordination.md` L4–5/
L44/L134 labels) — mechanical name retarget in the governance slice; installation.md already supersedes the
install sections.

### Slice 2b — retirement + both aliases: STOPPED before aliases (M1 not verifiable by the writer)

Per the dispatch caveat, aliases are created only with retirement and M1 verification. M1 (OD1) is a genuine
read-only harness listing; neither CLI here offers a session-free listing command the writer can run
(codex-cli 0.160.0 `--help`: no skills listing; a Claude listing needs a session), and OD1/Plan §8 make it an
owner-performed row with no session creation authorized. Creating the commit in this checkout would also place
`.claude/skills/restruct-sdlc` inside `.claude/worktrees/agent-<id>`, which new Claude sessions started
from the main checkout list as a worktree-scoped skill (as they list the worktrees' `web/.claude/skills/vestive-design`).
Nothing retired, no alias, no legacy file removed. Smallest remedy for Codex disposition: authorize the writer to
make the retirement+alias commit here (accepting the worktree-scoped listing until cleanup), with the owner's M1
listings taken later from a disposable clone of the pushed candidate outside `.claude/worktrees`, before F1.

### Slice 2b — executed under the recorded staging route (2026-10-18)

Authority: slice-2b disposition and c01+2a partial ACCEPT (not included) (SHA-256
`76d4bf03…a166`, staging `3a6d7a43`). Pre-slice: fetched staging `3a6d7a43`; asserted branch/HEAD `e58319da`; rebased
(one Result conflict at the correction commit resolved to staging's imported Result `658b1d03…8a60`, which equals the
pre-rebase final; source diff vs `e58319da` empty) → `9344e941`/`8f1216e9`/`8f7f83fa`, lease-pushed. Rebind:
status-only drift (SUPERVISOR-HANDOFF `498f5152…64a2`, legacy link still 1 occurrence; PROGRESS `537e95e2…98f8`;
ROADMAP unchanged); no helper/observer/checker/template/CI/settings drift.

One commit:
- **Retired** (pre-retirement blobs for the path map): `.agents/skills/claude-session-handoff/SKILL.md` `9a0bf543`,
  `…/agents/openai.yaml` `614ff597`, `…/references/local-session-handoff.md` `3abef5aa`,
  `.agents/skills/repository-supervisor/SKILL.md` `37742397`, `…/agents/openai.yaml` `20c7344f` (pattern 28). The
  legacy directory now holds only the two forwarders; `skills/sdlc-management/` was already fully moved in slice 1.
- **Aliases**: `.agents/skills/restruct-sdlc` and `.claude/skills/restruct-sdlc`, relative symlinks, `readlink`
  exactly `../../skills/restruct-sdlc`, one realpath (the canonical package). `ln -s` succeeded (no native refusal).
  Worktree-scoped discovery is the accepted temporary artifact; no main-checkout link, home install, observer or
  session change. **M1 pending** (owner native listings on the exact pushed candidate in a disposable clone outside
  the discovery tree, at full-candidate readiness); discovery is not green.
- **Bounded doc fixes** (review, no new tests): SKILL.md limits now say CLI versions are compatibility baselines
  that warn while protocol/peer-identity/key-shape/code-binding gates fail closed, and that live discovery refresh
  is not assumed and is verified by a harness listing (no platform-wide unsupported claim); installation.md drops
  the unsupported `pin --role supervisor` and uses `handoff.py --repo PATH binding propose`. SKILL.md
  `b835244a…6593` (119 lines), installation.md `71f62393…ad03`.

| Command | Exit | Result | Log SHA-256 |
|---|---|---|---|
| `python3 -B …/s2/v4_p1_p4.py skills/restruct-sdlc` | 0 | P1–P4 pass, 0 findings, SKILL.md 119 lines | `4b2a0a796cd121dde017f912b470ca366f719821525422872e078ba1af7ba6d1` |

`check_sdlc.py package skills/restruct-sdlc`: P1/P3/P5 pass; P6 fixture-UUID finding unchanged (checker slice).

## Slices 3–6 — home binding (V1), stopped cutover and compat run (V2), checker formats and templates (V4/P3t), V3 classes, governance (2026-10-18)

Writer: fresh single writer `.claude/worktrees/agent-<id>` (checkpoint handoff; attempt 1/B5/checkpoint 3
unchanged). Branch/HEAD asserted `7796e347` before any write. Reconciliation: staging advanced `3a6d7a43 → 8ede7dbf →
58dad972 → 6a494e24`, status-only each time (SUPERVISOR-HANDOFF, PROGRESS, G058/G019 records, slice review now
`9e6d0711…140c` with the slice-2b partial ACCEPT); rebased cleanly (Result identical to staging's import), ancestry
verified before every batch and push; no helper/observer/checker/template/CI/settings drift.

**Changes (current SHA-256):**
- `scripts/handoff.py` (`1398cee2…7efe84`), Plan §5.4 home mode: package detected only when the executing file is
  under exactly one supported home root (`~/.agents|.claude/skills/restruct-sdlc`, resolved; distinct differing
  copies → `ambiguous-home-copies`); repository precedence (`repository-install-present` + the local command);
  explicit `--repo` required in home mode (`home-repo-argument-required`; repository mode keeps the `.` default
  so existing callers are unchanged); committed, tracked, clean `management/RESTRUCT-SDLC-CODE-BINDING.json`
  (`home-code-binding-missing|dirty|mismatch`: schema, exact inventory of `scripts/{handoff,watch,check_sdlc}.py`,
  loaded bytes, package version, review record = a different tracked clean management file with matching hash);
  read-only commands (`discover`, `check`, `status`, `watch.py status`) attach an explanation, everything else
  refuses before acting; `watch.py stop` is also allowed unbound (PID-verified quiescence step — deliberate choice,
  flagged for review); `receipt` requires the binding (Plan list taken literally). Read-only `binding propose
  [--review-ref]` and `binding quiescence` (`shared-code-in-use`: process scan over home roots, invocations filtered
  to scripts resolving into them, file holders by inode, uninspectable processes fail closed). `prerequisite_evidence`
  in home mode uses the verified binding as the committed side and the home compat bytes for forwarders. Repository
  mode chain unchanged. `process_scan` gained an optional `within` and inode-set form (default behaviour identical).
- `scripts/watch.py` (`84b225cd…eef3761`): one `code_binding_gate` call (None in repository mode); status carries it.
- `scripts/check_sdlc.py` (`419a6340…3dbf`), Plan §5.5/§5.8: `--format canonical|legacy`, `mixed-format` under either
  format, exact relative aliases (both paths, one realpath, COLLISION/dangling/escape), scripts-only compatibility
  directory verified (`compat-forwarders: verified (2 files)` / `compat-forwarder-modified|extra-file`),
  `--capabilities` (`not-adopted`, `dependency-missing`), `adoption --package` with binding states, `package
  --public-remote` (exact URL, bundle-root README only), two MIT licences, P2/P3t/P4 gated on `restruct-sdlc`,
  SKILL_ROOT code-span resources checked (P3t templates, A2 rendered docs). Baseline parity defect fixed:
  `IGNORE_LINES` lacked `/.claude/settings.local.json` and `/.local/` already present in `templates/root/
  gitignore.lines` at `30890b16` (by inspection; a one-test baseline probe was invalid — its fixture needs the whole
  package — log `9a643c51…621c`, history).
- Templates (P3t; step 6): 82 legacy skill links → literal `{{SKILL_ROOT}}` code-span references (no conditionals),
  legacy labels → module names, `parameters.example.json` (`7393aa4d…d5b3`) gains `SKILL_ROOT` only. Schema id and
  `gitignore.lines` comment kept (compatibility ids).
- `test_handoff.py` L10–12 fixture UUIDs now synthetic (P6 finding closed).
- `references/installation.md` (`6869650b…6f38`): live-refresh wording aligned with SKILL.md (review residual), home
  command contract, SKILL_ROOT rendering note. Package references: legacy prose names → module names (adoption.md
  superseded install section now one-skill, compact path, check command; coordination/closeout/roles labels).
- Governance (step 5): mechanical active-link retarget in AGENTS.md, docs/agents/{CLAUDE-SESSION-HANDOFF,
  GITLAB-BRANCH-WORKFLOW,CLAUDE-PERMISSIONS}.md (links only), docs/SECURITY.md, docs/foundation/OPERATIONS.md (L555
  layout row; L203/L563 kept as history); AGENTS.md service-root explicit-read sentence; dated runbook section
  (G8a/G8b pending, rollback mirrors the order). `G055-legacy-path-map.md` (`48338353…5935`): 38 rows old → new with
  pre-migration blobs at staging `6a494e24` (includes `9a0bf543`, `614ff597`, `3abef5aa`, `37742397`, `20c7344f`).

**Tests added:** test_handoff T14 +6 (escaping alias target over both aliases; home binding approves compaction;
binding states refuse before any write incl. untracked, dirty, version drift, inventory/loaded-file mismatch,
self-approval, malformed, package-held manifest; ambiguous copies/identical copy/alias, repository precedence,
`--repo` required; shared-code quiescence; watch commands need the binding). test_watch `TestStoppedCutover` (forward
legacy→canonical and C1 reverse: seeded pin/binding/history/uncertainty hashed, live relocation refused by the
verified process and the run lock, interrupted start repeated, one process with the target argv, no replay, no
retry). test_check_sdlc +10 (canonical/legacy fixtures; legacy/mixed/incomplete; public remote; licences; SKILL_ROOT
resources; single entrypoint; aliases; compat directory; legacy adoption; capabilities; home binding states).
test_consolidation: `TestInstallMigrate` +4 (install loop run verbatim from installation.md; migration order with the
partial refusal and parity; `test_rollback_requires_stop_then_legacy_passes`), `TestRouting` (lifecycle walkthrough
with both disabled variants and saas-foundation unchanged; routing + read-only status), `TestHomeTwoRepositories`,
`TestCapabilities`.

| Command (`python3 -B -m unittest discover -s skills/restruct-sdlc/scripts …` unless noted) | Exit | Count | Log SHA-256 |
|---|---|---|---|
| `-p test_handoff.py -k test_t14_ -k test_t15_ -v` | 0 | Ran 32, OK | `65a3026d1ed7b1197ed816b762e2716680769dd9c15f6bf2986b8d03eab799c3` |
| `-p test_handoff.py -v` (whole file; main() refactor affects every command) | 0 | Ran 75, OK | `412298c16e7c3844698f0b9cd74db11661192dfb9bed4bee075dfa99c84509ba` |
| `-p test_handoff.py -k ReceiptRouting -k test_t14_home -v` (UUID constants changed) | 0 | Ran 9, OK | `016b9baad7faf88e64845ad7c47c391618a58c7d76f0fe25b27c24b9f6b8383b` |
| `-p test_handoff.py -k TransportCompatibility -v` (same) | 0 | Ran 5, OK | `f50260766a95008ca48acd83b45e073ea0a7e8f57811d7fb5b7216789fdef755` |
| `-p test_watch.py -k TestStoppedCutover -k test_start_ -k test_stop_ -k test_fresh_start_ -k test_known_unsent -k KnownUnsent -v` | 0 | Ran 14, OK | `3403c164fbc7653c31649cfb335a7a9b3e7d3bafc112b91e121dfd7c60a6fbb1` |
| `-p test_watch.py -v` (whole file once) | 0 | Ran 86, OK | `c283028b6d11e5c04d68b25ac01a2070400cd023caa776618dfc6a1e71aee901` |
| compat: base `30890b16` `test_handoff.py` vs candidate scripts (expected list `expected.txt` `6c2d2407…a5da`, written first) | 1 | Ran 64, 17 fail-path-only exactly as listed, 47 pass | `d75c1e6e558d48807c1156662dab24c46a4ac9731244c9d4d974156837264959` |
| compat: base `30890b16` `test_watch.py` vs candidate scripts | 0 | Ran 84, OK | `4fcd6541da1da68e6019717572d2bfe0973377579b9445ae16960952a3c7d423` |
| `-p test_check_sdlc.py -v` run1 | 1 | Ran 45, 1 FAIL (baseline IGNORE_LINES parity) | `44e3101374827bca9bb3865d363392164901b337f0454e4bc5b6e1abba3d0542` |
| same, run2 (constant aligned) | 0 | Ran 45, OK | `e9fbc2d715ea974e07aa28f94a16fb1f8d8e46a2590d7484e7d3a92ef5873f92` |
| same, run3 (after reference retargets) | 0 | Ran 45, OK | `ddc9f4f6e2bf3907739a11e04ccb7cc7787b6d59334bf2d52f5b575fe463bf90` |
| `check_sdlc.py package skills/restruct-sdlc` (after all package edits); `wc -l SKILL.md` = 119 | 0 | P1–P6 pass | `860c9a90ab973fb77731d82194ecdd1442bc91c91526b74c1f60397462af09d9` |
| `-p test_consolidation.py -k TestInstallMigrate -k TestRouting -v` | 0 | Ran 10, OK | `1f6b0e156fb38d45132591560e2c5d7ca5ff5c70c2874f82ad855315a4dd6e6e` |
| `-p test_consolidation.py -k TestHomeTwoRepositories -k TestCapabilities -v` | 0 | Ran 2, OK | `e71f6244aa1a6fa88e2cb0ff854c5bd93c60d9bf1617594b83a2e65335fb066d` |
| `-p test_consolidation.py -v` (whole file once) | 0 | Ran 12, OK | `7363d3c16380019b8d688da1a4ee3515990582957bb278d72c84220af2024de1` |
| `check_sdlc.py adoption . --release v0 --capabilities full` (Beginify) | 1 | see below | `c55ca023fc562c7986a35a38af728ef5ae1ed6496cc4bef4ef7e225ecfb4a959` |

Compat replacement check for the 17 path-only cases: the same-named candidate cases (fixture executes as the
repository's canonical copy), whole-file log `412298c1…09ba`. Cause verified in the log: the only prerequisite
failure is `executing-path-outside-canonical`; no behaviour assertion removed.

Beginify adoption (expected, not a regression): compat forwarders verified; active findings only CLAUDE.md (16
dangling `skills/sdlc-management/` links, held: dispatch forbids CLAUDE.md edits), SUPERVISOR-HANDOFF.md (1; prefix
retarget at the merge point per correction01 §C3) and PROGRESS.md (1; Codex-owned, outside scope); 157 findings in
59 historical records (closed evidence keeps old paths and pre-existing dangles, e.g. project-archetype).

Logs under `$HOME/.local/state/platform-evidence/g055-impl-a1/s3/`. Fixture child processes reaped (`ps` clean after
each batch). Live observer PID 3867 untouched (existing real-/proc scan cases only read process metadata).

## Correction 02 — P6 exact URL and A2 containment before resolution (slices 3–6 review CORRECTION LOOP, 2026-10-18)

Writer: fresh single writer `.claude/worktrees/agent-<id>`, branch `g055-restruct-sdlc-consolidation`
(attempt 1/B5/checkpoint 3 unchanged; correction pass 1 of 2 for these two findings). Asserted remote HEAD `8ca14836`
before any write. Reconciliation: staging `8f584d01` (review "six dispositions", `8c270979…5f23`) then `594979b3`
(G058 records, PROGRESS/SUPERVISOR-HANDOFF status lines, review addendum "CLAUDE.md editor ownership fallback",
`cbd5b2f3…a4b8`): status-only; rebased cleanly (the slices 3–6 Result commit dropped as already upstream; Result
`71ec1be7…b3e5` identical), ancestry verified before the batch; no helper/observer/checker/template/CI/settings drift.

**Changes:**
- `scripts/check_sdlc.py` (`ce968c5b…0add`):
  - P6: `--public-remote` now removes only complete URL occurrences (`public_url_re`): not preceded by a URL
    character, not followed by one; trailing sentence punctuation is a boundary unless more URL characters follow
    (`repo.git` is a different URL). Bundle-root README-only restriction and all literal/private-host rules unchanged.
  - A2: `symlinked_ancestor` checks every ancestor lexically before anything is resolved. `installed_package` reports
    `install-ancestor-symlink` for a symlinked `.agents`, `.agents/skills` or `.claude/skills`; an exact alias whose
    target components include a symlink (for example a symlinked `skills/`), or whose target is an undeclared
    symlink, fails "alias escapes the repository: target components must be real directories"; the flat install
    also requires containment. `check_compat_directory` reports `compat-forwarder-escape` for any symlinked
    component (`.agents`, `.agents/skills`, the compatibility directory, `scripts`) before comparing bytes. Declared
    contained aliases (`../../skills/restruct-sdlc`, `../../.agents/skills/restruct-sdlc`) and the genuine flat or
    source-maintainer layouts keep passing. No symlink allowance, privacy exemption or fixture-only path.
- `scripts/test_check_sdlc.py` (`d6b71775…c4a8`): public-remote test extended (accepted exact forms: Markdown
  link, angle brackets, trailing period, two occurrences; rejected: `-evil`, `/other`, `.git`, `?x=1`, `%2F..`,
  prefixed `x…`, extended URL beside an exact one). New `test_a2_ancestor_escapes_fail_before_resolution` (flat
  package with `.agents` / `.agents/skills` relocated outside; compatibility directory and its `scripts/` relocated
  outside with byte-equal forwarders; contained declared aliases pass; symlinked `.claude/skills` and symlinked
  `skills/` alias target fail; restored layouts pass).
- Disposition 4: `scripts/handoff.py` (`27be6466…14b9`) comment only — `watch:stop` is the reviewed quiescence step;
  it signals (SIGTERM) the verified observer process and removes that process record after exit; it does not
  repin, send, resume, replay or write other state. No behaviour change.
- Dispositions 2, 3, 5, 6 need no source change (CLAUDE-PERMISSIONS three link paths kept; PROGRESS and
  SUPERVISOR-HANDOFF untouched, Codex reconciles at integration; `receipt` stays binding-required; A5 five entries
  kept). Disposition 1 / addendum: CLAUDE.md not edited; Codex owns the 17 link-target changes at the final
  source handoff — recorded as **pending active-link reconciliation**, to close before final adoption/F1.

| Command (`python3 -B -m unittest discover -s skills/restruct-sdlc/scripts …` unless noted) | Exit | Count | Log SHA-256 |
|---|---|---|---|
| `-p test_check_sdlc.py -k test_public_remote -k test_a2_ -k test_compatibility_directory -k test_fresh_install -k test_home_package -k test_bundle_passes -k test_p6_ -k test_rendered_adoption -v` | 0 | Ran 12, OK | `76bc033f7c081316d206337efc30921e2a9f8b777c2f669504dde85114d3999e` |
| red proof: same new tests against the pre-correction checker (temporary package copy, `HEAD` checker) | 1 | Ran 2, 2 FAIL (`('.agents', [])`; `…repo-evil` accepted) | `38e889cd8ab641921486a528f15c5198e6c8ca67994c7bd988d2074b6407829f` |
| `-p test_check_sdlc.py -v` (whole checker suite: `installed_package`/compat run in every adoption case) | 0 | Ran 46, OK | `b659914fc8739322fd20959164188c1b1bf791d381ab59d484fe6102108adab2` |
| `-p test_consolidation.py -k TestInstallMigrate -k TestRouting -v` (the classes that call the checker) | 0 | Ran 10, OK | `16fdc77ad3dcf5f629e1260cdb890d4074ec5f1c1055a48a8f6fadabacf7f3cf` |
| `-p test_handoff.py -k test_t14_home -v` (after the handoff comment edit) | 0 | Ran 5, OK | `b58e4e5b7fbf8649177c416afd58087e95af46d4c9a39519e3330b287ca6064c` |
| `check_sdlc.py package skills/restruct-sdlc` (final bytes) | 0 | P1–P6 pass | `2999a2318f9aa36b682fe5362c08e9c25a52caf572a3079e36fc0519b49db89a` |
| `check_sdlc.py adoption . --release v0 --capabilities full` (Beginify link proof) | 1 | expected; see below | `fbfe1fdfe4faaa305d1c4c0af46b4a5a72338b095729708f85a488c11a6210f2` |

Beginify adoption: compat forwarders verified; no `install-ancestor-symlink`/`compat-forwarder-escape` on the real
layout; active findings unchanged — CLAUDE.md 16 (Codex-owned reconciliation), SUPERVISOR-HANDOFF 1 and PROGRESS 1
(Codex at integration); 158 historical (157 + one in the newly imported G058 records; closed records keep old paths).
Not rerun (unchanged inputs): whole `test_handoff`/`test_watch`, compat runs, other `test_consolidation` classes.
Scope: 87 changed paths vs `594979b3`, 0 outside the 28 patterns; CLAUDE.md, PROGRESS, SUPERVISOR-HANDOFF and
settings unchanged. Logs `$HOME/.local/state/platform-evidence/g055-impl-a1/c2/` (0700). No fixture processes left.

## V5 — stage-only export (Plan §5.9, OD2 S; 2026-10-18)

Candidate `e2916bff8dec95ff0a3757dbd773cfc4ca829a40` on staging `594979b3` (worktree clean; ancestry verified).
Evidence `$HOME/.local/state/platform-evidence/g055-impl-a1/v5/` (0700); tools in `…/v5tools/` (outside both
checkouts). **Outcome S: stage only.** Nothing was written to the sibling or any public repository; no Git command
ran there; its dirty overlay was only read.

**Sibling (read-only), before and after V5 — identical:** `refs/heads/main` `cb998602…f7f7` (HEAD → main), tags
`1.0.0`, `1.0.1`; `MANIFEST.tsv` `99f153fb…1b51` (43 rows), `README.md` `ec39ef94…0763`, `LICENSE` `af5fb221…67c6`,
`.git/index` `6c3444d8…c5f9b0` — all equal to S0.5. W preconditions not met (overlay dirty; `status --porcelain`
not runnable here) → S, as defaulted.

**Stage build** (`build_stage.py`): `git archive e2916bff skills/restruct-sdlc` (tar `source-e2916bff.tar`), extracted
with the `data` filter (regular files/directories only); bundle-root `LICENSE` = the package MIT file (byte-equal to
the sibling's); `README.md` = sibling owner prose with only the six listed sections replaced (below); `MANIFEST.tsv`
rebuilt: 42 rows, unique, `151a2a5ac5c9b6c6ae61eb697035f50c582905f1997666a18641a5f00a7af973` (2 `new`, 5 `adapted`,
35 `verbatim`). Stage: 43 files; no symlink, bytecode or VCS metadata. Mapping recorded by the comparison: 40
stage-only paths (`skills/restruct-sdlc/**`), 41 sibling-only paths (the three legacy skill trees a W export would
remove), `LICENSE` equal, `README.md`/`MANIFEST.tsv` differ.

**Private literal list** (`make_literals.py`, G053 §4.2, generated 0600 in the evidence directory, never in the
bundle; values not printed): 58 entries — archetype identity 4, commit identity 4, origin 4, checkout path/HOME 3,
19 non-synthetic UUIDs under `management/` plus their 19 prefixes, terms `G021`, `WorkOS`, `Vestive`,
`identity-qualification`, `OS001` (`Beginify` already present via the archetype values).

**Deviation (export adaptations; for Codex disposition).** The first stage (all package files verbatim) failed P6
with the literal list in five package files: the project name in `references/{handoff,installation,
local-session-handoff}.md`, `G021` in `references/supervision.md`, `OS001` in `scripts/test_watch.py`. Plan §5.9
predicted only the README and the `test_handoff.py` UUID lines as adapted (the UUIDs are now synthetic at source, so
`test_handoff.py` exports verbatim). 1.x adapted the counterparts of the first four at export, so the stage applies
the same public-identity invariant — only the lines naming the originating project change, reusing the 1.x wording
where a counterpart exists — and the source keeps Beginify's own text and policy (no source change). Exact diffs
(`adapted.diff` `61d86d2a…2f2b`; 7 lines removed, 7 added):

| File | Line(s) | Source → stage |
|---|---|---|
| references/handoff.md | 7 | "adapted from the Beginify runbook verified" → "adapted from the originating repository's runbook verified" (1.x wording) |
| references/installation.md | 14, 71 | "Source maintainer (Beginify)" → "Source maintainer"; "Repository-local (default; Beginify)." → "Repository-local (default)." |
| references/local-session-handoff.md | 1 | "adapted from docs/agents/CLAUDE-SESSION-HANDOFF.md in Beginify platform." → "adapted from the originating repository's docs/agents/CLAUDE-SESSION-HANDOFF.md." (1.x wording) |
| references/supervision.md | 132–133 | "Preserve UUID canonical identities, current tenancy/authority semantics and the / G021 product-feature barrier." → "Preserve the adopting repository's canonical invariants, authority semantics and / any declared delivery barrier." (1.x wording) |
| scripts/test_watch.py | 281 | status-line task label `OS001` → `AB001` (still a non-`G` label; assertion unchanged) |

Alternative for review: make the installation.md and test_watch.py lines generic at source (both in scope) so they
export verbatim; the handoff/local-session-handoff/supervision lines carry Beginify provenance/policy and stay
export-adapted as in 1.x.

**README (owner prose preserved; proposed text applied in the stage).** Lines 1–31 (title, intro, status line,
"Why Restruct?", the three continuation paragraphs) and the License section are byte-equal to the sibling; only
Skills, Requirements, Installation, Verify, Adopt and Limits changed (70 sibling lines → 91; `readme.diff`
`cd844add…e6ce`; stage README `4fac1b68…864b`): Skills — one `restruct-sdlc` row with its three modules, "no role, no
authority", 2.0.0 replaces the three 1.x skills; Requirements — Linux/WSL2, Git, Python 3, `/proc`, 2.1.290/291/296
baselines with fail-closed gates; Installation — the one-package loop from installation.md (SRC/DEST form), collision
rules, the optional Claude alias, home placement only with `--repo` and a committed code binding, and a new
"Upgrading from 1.x" subsection (preflight `--format legacy`, live-code gate, two forwarders, rollback by revert);
Verify — `check_sdlc.py package .` plus `unittest discover -p 'test_*.py'`, check a copy without `.git`; Adopt —
parameters incl. `SKILL_ROOT`, `--capabilities`; Limits — one-package paths, aliases, refresh, binding, platforms.
**Owner wording decisions requested (not applied):** intro line 3 still says "a bundle of three agent skills" (Plan
keeps the intro untouched; suggested "a single agent skill with three modules"); the status line is unchanged.

| Check | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `check_sdlc.py package STAGE --literals literals.txt --public-remote <public URL>` | 1 | P1–P5, P7 pass; P6 only `README.md` literal #56 (`Vestive`, the recorded owner-prose exception, G053 §12.6: stage README L23–24 `vestive.com` links in the kept "Why Restruct?" text) | `baccd9dda36aaa95db062d1dcdbfe0f247751a549d03b37daa1f7f5c84145e2c` |
| same without `--literals` | 0 | P1–P7 pass | `c1eb1aff66dae19fb36ccba168c503f2e2f1b7849c64586225fb85f2ddae47cd` |
| same without `--public-remote` (negative) | 1 | only `README.md: hosted-remote` (the exact public URL is required) | `8ac042f9da9ef48d659668ed44334c591b0872bb9676e5a56f87300bd7ff2492` |
| `v5checks.py` (cmp, README, hygiene, notices, sibling comparison) | 0 | 35 verbatim files `cmp`-identical to the committed source; 5 adapted (diffs above); README kept lines equal; no symlink/bytecode/VCS; notices: MIT `LICENSE` ×2 and `Copyright (c) 2026 Vitruvian Technology, Corp.` only | `0e7647245a4adcd1e1733ede15001165f9f21b2d60baeed97551715b47f1f23d` |
| exported `test_check_sdlc.py` on the stage | 0 | Ran 46, OK | `da2c8442cfd9dba6e640633eca6c7d9aebe84b60afda4c5d663b917af3bdec69` |
| exported `test_consolidation.py` on the stage | 0 | Ran 12, OK | `bfe6daa88285cd1690ead40174eb24a23591c91bf08bb8abc5193091b665050a` |
| exported `test_handoff.py` on the stage | 0 | Ran 75, OK | `b6650acb4900634668c4dcf9d611f01de409eaac99e3c19f5ec25796d9e31875` |
| exported `test_watch.py` on the stage (incl. the adapted line) | 0 | Ran 86, OK | `80b6c723fc6550c7b932c8b9238e7fe8b4c3c3199065713fffdeb26c417ddb12` |
| package check again after the suites (stage unchanged, MANIFEST verifies) | 0 | identical result | `c1eb1aff…47cd` |
| `collisions.py` — the stage README loop verbatim except `TARGET` | 0 | absent → INSTALLED (tree equal); rerun → IDENTICAL; 1.x `sdlc-management` copy → COLLISION, tree unchanged; edited helper → COLLISION, unchanged; symlinked `.agents/skills` → STOP, nothing written; symlinked destination → STOP, nothing written | `7edd62ed20920b4426c096fe2116ba7126bcd4dc41e1ff5b2774873e98470676` |
| gitleaks `zricethezav/gitleaks:v8.30.1` (`sha256:c00b6bd0…`), validate.sh tar-on-stdin form, `--network none`, `--pull never`, `infra/security/gitleaks.toml`, under `local-harness.lock` | 0 | report `[]`, pipe `0 0 0`, "no leaks found" | report `37517e5f…b570`, log `588d1050…c808` |

Resources: `bash infra/ci/docker-desktop-mounts.sh` first (14% of mounts); lock taken by a background
`flock -n … sleep 600` (holder verified), released by ending that holder after the scan (exit 143); no container
left; no fixture process left. The exported suites ran from the disposable stage with temporary HOMEs, so the whole
`test_handoff`/`test_watch` files ran once on the exported layout (Plan V5 "four exported suites"); source-tree
evidence for those files is unchanged and reused. The earlier compat run (47 pass, 8 fail + 9 errors = the 17
predeclared path-only cases, `d75c1e6e…4959`) stays historical, not green.

## Correction 03 — P6 continuation after the approved URL (Correction02/V5 partial disposition, 2026-10-18)

Authority: Correction02/V5 partial disposition (not included) (SHA-256
`e88c29ac…2162`, staging `8b38d03e`): one P6 residual. Writer: fresh single writer
`.claude/worktrees/agent-<id>`; attempt 1/B5/checkpoint 3 unchanged; correction pass 1 for this residual.
Pre-batch: asserted branch/HEAD `93ce5dc0` (= remote); fetched staging `8b38d03e` (adds only the review section;
Result `bad1e635…3cc2` identical on both sides); rebased (one Result conflict at the correction 02 commit resolved to
the replayed version; tree vs `93ce5dc0` differs only by the imported review file), ancestry verified, inputs rebound
(review `e88c29ac…2162`, Result `bad1e635…3cc2`, scope `e7c2bebc…b64d`); no helper/observer/checker/template/CI/
settings drift.

**Finding.** `(`, `)`, `[`, `]` are legal URL characters, but the correction 02 boundary only refused URL-core
characters and sentence punctuation followed by URL-core characters, so `<URL>(private)`, `<URL>)(x)`, `<URL>).git`
and similar longer URLs were stripped as the approved one.

**Change** (`scripts/check_sdlc.py` `825f5e80…f48a`, commit `d46af97a`): `public_url_re` now exempts the approved URL
only when the whole run of sentence or closing punctuation after it (`.,:;!?'` and `)`, `]`, `>`, `"`) is followed by
no URL character, delimiter or further punctuation (new constants `URL_DELIM`, `URL_CLOSE`). The preceding-character
rule, the bundle-root README-only restriction, literal/private-host rules and the hosted-remote pattern are
unchanged; no exemption was broadened. `scripts/test_check_sdlc.py` (`c1f441d9…96ea`): new
`test_public_remote_continuation_after_delimiter_is_a_different_url` — 11 negatives (`(private)`, `)(private)`,
`)private`, `]/x`, `[1]`, `'(x)`, `.(x)`, `).git`, Markdown `source (not included)`, angle `<<URL>(private)>`,
sentence `Source: <URL>(private).`) and 7 positives (Markdown link, angle brackets, trailing period, `(see <URL>)`,
Markdown link + period, double quotes, mixed sentence). Existing public-remote test unchanged.

| Command (`python3 -B -m unittest discover -s skills/restruct-sdlc/scripts …` unless noted) | Exit | Count | Log SHA-256 |
|---|---|---|---|
| red proof: `-p test_check_sdlc.py -k test_public_remote -v`, new test against the unchanged checker | 1 | Ran 2, 1 FAIL (`<URL>(private)` exempted) | `ab057f93e6f98c9d837bfab3377264cd69cf41a102b0858667abf3901b143f44` |
| per-case probe `c3/probe.py` (pre-correction blob `48cf6363` vs corrected checker; 22 cases) | 0 | all 11 new negatives exempted before, refused after; 3 earlier negatives refused both; 8 positives exempt both; wrong_post=0 | `4481bc4a2494a7c53f3f716073932387d4759529033bbd351214b8cb4b6cc145` |
| `-p test_check_sdlc.py -v` (whole checker file, the affected unit) | 0 | Ran 47, OK | `1754c60ff7ceee2bd2e1bf9432d48dc91be90684e329e59dafb8873aaa161a3b` |
| `check_sdlc.py package skills/restruct-sdlc` (final source bytes) | 0 | P1–P6, P3t pass (P7 is bundle-only) | `41e78686dd2c2f630292859fb17efd2ba6ef3e6430d3d4d82cfa95cd52d9fd57` |

Not rerun: `test_consolidation` (no caller passes `--public-remote`; `public_url_re` is reached only through it),
`test_handoff`, `test_watch`, adoption check (P6 public matching is package-only). Logs
`$HOME/.local/state/platform-evidence/g055-impl-a1/c3/` (0700).

### V5 refresh (stage only, after correction 03)

Candidate `d46af97ac236d7be13f27bc17ff03efc2a0e9b67` on staging `8b38d03e`, worktree clean. Same `build_stage.py`,
sibling README and section text; new tar `source-d46af97a.tar` (`eefbb99a…fc8f`); stage
`…/v5r/export-stage`. `diff -rq` against the V5 stage: exactly three files differ — `check_sdlc.py` and
`test_check_sdlc.py` (both still `verbatim`; the test file is exported verbatim, so it changes with the checker) and
`MANIFEST.tsv`. MANIFEST: 42 rows, two rows changed (`check_sdlc.py` `ce968c5b…0add` → `825f5e80…f48a`,
`test_check_sdlc.py` `d6b71775…c4a8` → `c1f441d9…96ea`), SHA-256
`22cd57e06a9175f45616875290fe5713317520de568a64b61f392ef5a7626a1b` (was `151a2a5a…a973`). Unchanged and reused:
README `4fac1b68…864b`, `readme.diff` `cd844add…e6ce`, the five approved adaptations (`adapted.diff`
`61d86d2a…2f2b`, identical), `v5checks.json` `0e764724…3d4d` (identical: 35 verbatim `cmp`-equal to `d46af97a`, 5
adapted, README kept lines equal, no symlink/bytecode/VCS, same notices; 43 files), the exported
`test_consolidation`/`test_handoff`/`test_watch` evidence, `collisions.py` evidence. Sibling read-only before and after:
`refs/heads/main` `cb998602`, `README.md` `ec39ef94…0763`, `MANIFEST.tsv` `99f153fb…1b51`, `LICENSE` `af5fb221…67c6`,
`.git/index` `6c3444d8…c5f9b0` — unchanged; no Git command there.

| Check (stage's own `check_sdlc.py`) | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `package STAGE --literals literals.txt --public-remote <public URL>` | 1 | P1–P5, P3t, P7 pass; P6 only `README.md` literal #56 — the existing exact G053 owner-prose exception (preserved "Why Restruct?" wording), not an allowlist and not a full private-literal pass | `bdd0ee24bd538b777811e793f60e3f36096289865180bc96939204cd2153bfbe` |
| same without `--literals` | 0 | P1–P7 pass (no-literals scan only) | `4d1e22bb23ebb416368dd602cda1fc4bfb010fc35667dc2b130c1314af09a025` |
| same without `--public-remote` (negative) | 1 | only `README.md: hosted-remote` | `a6510035086375351b7cb2e4077241b2bbbe76b714058a6188829e96380c1c93` |
| exported `test_check_sdlc.py` on the stage (the one changed exported suite) | 0 | Ran 47, OK; stage unchanged afterwards (`diff -rq`) | `6a00bbecc792e9fd9324e584ded2a4f25edd714289d602d1b491d34a3759b420` |
| package scan again after the suite (no literals) | 0 | identical output | `4d1e22bb…a025` |

The stage README's own public link is the Markdown form `GitHub (not included).` (closing parenthesis then period) and is
still exempted. gitleaks not rerun: the stage delta is the two checker files (regex constants and synthetic test
URLs built from fragments; no credential-shaped strings); the V5 gitleaks evidence (`37517e5f…b570`/`588d1050…c808`)
is retained for the other 41 unchanged files and stays bound to `e2916bff`. No lock taken, no container started.

## Acceptance-criteria progress (partial; nothing claimed complete)

- AC1 package: relocation, entrypoint (119 lines, P1–P4 pass), module references, legacy retirement and both
  aliases done; M1 discovery pending. AC4: forwarder parity/refusals and genuine-Git-root rule proven; install
  loop, migration and rollback cases pending (rollback procedure documented in installation.md). AC5:
  repository-local trust chain extended and proven; home binding pending.
- AC2, AC3, AC6–AC14: not started.
- Update after slices 3–6: AC4 install loop/migration/rollback and AC5 home binding proven (focused + whole files);
  AC7 stopped cutover both directions; AC9 checker formats; AC2/AC3/AC13/AC14 fixture evidence present (M1 real
  discovery pending); AC10 (V5 stage) not started; AC11 compat run recorded; AC12 at F1.
- Update after correction 02 and V5: AC9 P6 exact public URL and A2 containment negatives added and passing; AC10
  stage-only export evidence recorded (S; five export adaptations pending Codex disposition; owner README wording
  decision open); CLAUDE.md, PROGRESS and SUPERVISOR-HANDOFF active-link reconciliation pending with Codex, to close
  before final adoption/F1.
- Update after correction 03 and the V5 refresh (final writer handoff; nothing claimed complete): AC9 P6 now refuses
  delimiter/closing-punctuation continuations (red proven, whole checker file green); AC10 stage refreshed (five
  adaptations ACCEPTed by Codex; strict scan = sole owner-prose exception). Open per AC: AC1/AC2/AC3/AC13 real
  discovery (M1); AC12 and final local qualification (F1); live/cutover gates G8a/G8b; active-link reconciliation
  (Codex) before F1; README intro wording (owner). No AC is self-accepted.

## Deviations, refusals, issues

- Architecture deviations: none.
- Native/guard refusals: **none** (Write/Edit under `.agents/**`, `git mv`, `cp`, `chmod`, `python3 -B -m unittest`,
  `sha256sum` all succeeded in this session; no settings edited).
- Scope gap for a later slice: `.agents/skills/repository-supervisor/agents/openai.yaml` exists on staging but no
  remove pattern covers it (only its `SKILL.md`). Retiring the legacy supervisor package needs one pattern
  `^[.]agents/skills/repository-supervisor/agents/openai[.]yaml$` (remove) — request for Codex scope disposition;
  not touched here.
- Discovery staging note: root `.claude/skills/restruct-sdlc` inside a `.claude/worktrees/*` checkout would likely be
  listed by running Claude sessions (this session lists nested worktree `.claude/skills` entries), so the aliases are
  created only at the retirement step and verified at M1 (C4: no live-refresh assumption).

## Cycle history

- Attempt 1, slice 1: one inner fix (forwarder shallow-path refusal). Correction passes: 0.
- Correction 01 (forwarder genuine Git root): pass 1 of max 2; first focused run passed; same attempt.
- Slice 2a (entrypoint/modules): first V4 run passed. Slice 2b stopped before aliases (M1 disposition requested).
- Slices 3–6: one inner fix (checker suite run1: baseline IGNORE_LINES parity); every other first run passed.
  Correction passes 0. Native/guard refusals: the worktree guard refused three compound shell forms (a `cd &&
  python -m unittest … > log; tail` chain, an `awk` program, a heredoc python edit followed by `grep`); each was
  rerun as plain single commands or a script file. No native permission refusal.
- Correction 02 (P6 exact URL, A2 containment): pass 1 of max 2; first focused run passed; red proof confirmed the
  new negatives fail on the previous checker. Worktree guard refused two forms: `git merge-base --is-ancestor …;
  echo` (rerun as plain `git merge-base`) and a `grep` whose patterns named `.gitignore`/`.gitlab-ci.yml` (rerun
  as a `python3 -I` set check). No native permission refusal.
- V5: first stage P6 run found five package files with private literals (deviation above; adapted at export); every
  other first run passed. The worktree guard refused one form: a checker command naming the public URL followed by
  `; echo exit=$?` (rerun as a plain single command with a file redirect). No native permission refusal.
- Correction 03 (P6 continuation): pass 1 of max 2; first post-fix run passed; red proof by test and per-case probe.
  No guard or native refusal. One tool quirk: `git show <blob> --output=FILE` printed to stdout and left an empty file;
  the pre-correction copy was taken from the saved output and verified by blob id `48cf6363`.

## Transfer state / next in Plan order

Slice 2b done. Next, in Plan order: home binding/`--repo` (V1), `TestStoppedCutover` (V2) and compat run, checker formats incl. the P6 fixture finding
(V4), remaining `test_consolidation` classes (V3), governance links + SUPERVISOR-HANDOFF prefix retarget + path map,
V5 stage-only, M1, F1. Merge held until G058 final ACCEPT; G8a/G8b live gates unchanged.

Update after slices 3–6: V1, V2 + compat, V4/P3t, V3 and governance done. Open decisions for Codex: CLAUDE.md
active-link retarget (17 hits, scope pattern present but dispatch forbids CLAUDE.md edits); PROGRESS.md 1 active
legacy link (Codex-owned); `watch.py stop` unbound and `receipt` bound in home mode (choices above). Next: V5
stage-only (export stage, P1–P7 with literals, gitleaks under the shared lock, sibling comparison read-only, OD2 S),
optional R1, SUPERVISOR-HANDOFF prefix retarget at the merge point, M1 (owner listings from a disposable clone of
the pushed candidate outside the discovery tree, requested once at full readiness), then F1.

Update after correction 02 and V5 (writer `agent-<id>`): correction 02 and V5 stage-only done. Open
for Codex: the five export adaptations (or the source alternative for installation.md/test_watch.py) and the owner
README intro wording. Codex-owned: the 17 CLAUDE.md link targets, the PROGRESS active link and the SUPERVISOR-HANDOFF
prefix at the safe final handoff/integration point. Next: optional R1, M1 (requested once at readiness), F1.
Merge and armed auto-merge held until G058 final ACCEPT; no live cutover.

Final writer handoff after correction 03 (writer `agent-<id>`, 2026-10-18). Source work in Plan order is
done through V5 (stage only); R1 is optional in the Plan (§9 step 7) and was not run. **Pending, none performed by the
writer:**

1. **Codex at integration (Codex-owned files, not edited here):** `CLAUDE.md` 17 active link targets
   `skills/sdlc-management/` → `skills/restruct-sdlc/`; `management/v0/PROGRESS.md` 1 active link; the
   `management/SUPERVISOR-HANDOFF.md` prefix retarget under r2c1 §C3 (prefix-only diff, re-grepped at the merge point).
   To close before final adoption/F1.
2. **Owner README wording:** intro line 3 "a bundle of three agent skills" kept (Plan §5.9 protected prose); proposed
   "a single agent skill with three modules". Stage-only; no public/sibling write.
3. **M1 (owner, OD1; requested once, at full-candidate readiness after item 1):** genuine native listings; the writer
   cannot run them (no session-free listing command; session creation not authorized). Exact owner procedure:
   > G055 M1 (read-only discovery evidence). Use the final pushed candidate SHA named in the request.
   > 1. `git clone --branch g055-restruct-sdlc-consolidation --single-branch <platform origin URL> /tmp/g055-m1/repo`
   >    (outside the Beginify checkout and any `.claude/worktrees` tree); confirm
   >    `git -C /tmp/g055-m1/repo rev-parse HEAD` equals the candidate SHA.
   > 2. Repository mode: in `/tmp/g055-m1/repo` start Claude Code and Codex in turn, open each harness's own skill
   >    listing (for example `/skills`) without sending any prompt, and copy the listing verbatim with
   >    `claude --version` / `codex --version`. Expected: `restruct-sdlc` (Claude via `.claude/skills/restruct-sdlc`,
   >    Codex via `.agents/skills/restruct-sdlc`); none of `sdlc-management`, `claude-session-handoff`,
   >    `repository-supervisor`.
   > 3. Home mode (AC13): with a disposable `HOME=/tmp/g055-m1/home`, install the package from the clone with the
   >    `references/installation.md` home loop (`~/.agents/skills/restruct-sdlc` plus the optional Claude alias
   >    `~/.claude/skills/restruct-sdlc`), start each harness in an empty non-repository directory
   >    `/tmp/g055-m1/empty`, and record the listing the same way. If a harness requires login on the disposable HOME,
   >    stop and record that; do not copy credentials or settings.
   > 4. No settings, permission or Beginify change; exit each session without a prompt; delete `/tmp/g055-m1`
   >    afterwards. Return the verbatim listings, versions and the candidate SHA.
   Absent these listings, the AC13 real-discovery row stays blocked, never substituted by fixtures.
4. **F1** (final local qualification) after items 1 and 3, on a fresh fetch/rebase/reconcile.
5. **Live gates G8a/G8b** (stopped cutover with the observer) and code-binding approval: unchanged, not started.
6. **Merge/MR:** held until G058 final ACCEPT; no MR opened, no auto-merge.

## Plan r3 resume — S0′ rebind, S1′ fold and V5 refresh (2026-10-19)

Authority: r3 review/release (not included) (SHA-256 at `c29ff777`
`bd7fdfe96ef7a731bed31134c53e076d68f6d0627f66af7d9ab7f0abdb561719`; released at `63bb72bc` as `c724ceade6f7…6238`, then
extended by its two carry-forward sections) + Plan r3 refresh (`113fdef1…0dba`) + inputs r3 (`7f151858…7f1c`).
Writer: single G055 implementation writer, native isolated worktree `.claude/worktrees/agent-<id>`
(generated name), branch `g055-restruct-sdlc-consolidation`; effort dispatched Opus 5.5 high, runtime unknown.
Budget: attempt 1 of B5 continues (checkpoint after attempt 3); correction passes unchanged; no reset.
Evidence: `$HOME/.local/state/platform-evidence/g055-impl-a1/r3/` (0700).

### S0′ — binding and rebase

- Branch asserted before any rebase: `g055-restruct-sdlc-consolidation` at `e91875fc1e26fe944470d23e1c33a2a475b63a56`
  (= origin). Rebased onto origin/staging `13c9b503`, pushed with `--force-with-lease` bound to `e91875fc` (→ `ec854d19`).
  Staging then advanced with relevant owner policy (coordinator relay), so the S1′ commits were rebased again onto
  **`c29ff7774a002f28c4eaba1ad1b14230aa039f1d`** and pushed with the lease bound to my own `ec854d19` (→ `fab89d1b`).
  Final base `c29ff777`; candidate `fab89d1b2dd0f6bacb52501046538b44a44ecc4c`. No branch pipeline was created.
- r3 inventory verified at its snapshot `825c7fb8`: 68 input entries (65 regular-file hashes, the
  `.agents/skills/vestive-design` symlink target, 2 absent paths) + 13 authority/history hashes = 81/81 match
  (`r3/verify_inputs.py`). Not compared against later mutable status files.
- Later drift `825c7fb8..c29ff777` inspected: E1 default-role paragraphs restored by Codex in AGENTS.md/CLAUDE.md
  (preserved; AGENTS.md carries them unchanged, CLAUDE.md untouched) and the optional Claude-supervisor mode kept;
  owner RC auto-merge option (`32b0daf9`, `13c9b503`) and optional RC Duo follow-ups (`72e7d37b`, `c29ff777`) in
  AGENTS/CLAUDE, GITLAB-BRANCH-WORKFLOW, the legacy `skills/sdlc-management` references/templates, Spec and the r3
  review; PROGRESS/SUPERVISOR-HANDOFF/BACKLOG/RC/G058 records are status-only. Current inputs at `c29ff777`: Goal
  `d0ba9c34…129d` (= r3), Spec `a64a219a…09dc` (r3 + two owner amendments), scope list `e7c2bebc…b64d` (28 patterns).
- Conflict inventory resolved toward staging text: AGENTS.md (auto-merged), CLAUDE-SESSION-HANDOFF.md (staging
  footer/short-report/post-merge sections kept, links retargeted, G055 cutover record appended),
  GITLAB-BRANCH-WORKFLOW.md (staging text, links retargeted), OPERATIONS.md (staging root-Vestive rows kept; the handoff
  row replaced by the Restruct SDLC row; no `web/.claude/skills/vestive-design` reintroduced — Q8), coordination.md
  (staging post-merge section + module transport paragraph), root/branch templates (staging text in the
  `{{SKILL_ROOT}}` link form). Rename/modify pairs: every staging-added line under `skills/sdlc-management/**`
  (`8b38d03e..c29ff777`) is present in `skills/restruct-sdlc/**` except four lines intentionally converted
  (`r3/carry_check.py`). Delete/modify: the two legacy SKILL.md files stay retired; their staging deltas
  (`r3/legacy-skill-staging-delta.diff`) are folded below. The three Result-only commits resolved to the staging Result
  (byte-equal to `e91875fc`); the last one dropped as already upstream. Scripts, aliases and Result equal `e91875fc`.
- Harness, read-only: `infra/validation/run.sh` L1484 `SKILL_ROOTS` includes root `.claude/skills/`; L1498 `SKILL_LINK`
  → `../../.claude/skills/vestive-design`; `profile-paths.tsv` rows 44–52 not changed by G055. The added
  `.claude/skills/restruct-sdlc` alias is committed content, so `B-SKILL-UNCHANGED`/`skill_relocation` with
  `--base` = the merged SHA are inert; proof remains the post-merge local full run (E2). No harness edit.
- Sibling, read-only (no Git command there): `refs/heads/main` `cb998602`, tags `1.0.0`/`1.0.1`, remote-tracking
  `origin/main` `ab7a6e9d` (stale); `.git/index` `6c3444d8…`, `MANIFEST.tsv` `99f153fb…`, `README.md` `ec39ef94…`,
  `LICENSE` `af5fb221…` — unchanged before and after. Fresh public main (`git ls-remote`, per the coordinator's
  "fresh main, not a pinned tip"): **`ebecfbd4f313d7378a17d0da4f9abc7410d2dc2a`** (newer than `ec98d4a2`/`862959a4`);
  its README equals the sibling README (`ec39ef94…`), LICENSE equal, MANIFEST `702c0181…` (42 rows, 1.x layout).

### S1′ — fold (commits `02769cc3`, `fab89d1b`)

| Row | Canonical location (evidence) |
|---|---|
| Q1 repair gate, Q2 supervisor verification | carried by rename into `validation-integration.md`, RESULT/REVIEW/BRANCH-WORKFLOW templates; legacy supervisor delta folded into `supervision.md` (repair-completeness link now intra-package) |
| Q3 short reports | `coordination.md`, templates (carried); the legacy "Mandatory short owner reports" sections folded into `handoff.md`/`supervision.md` |
| Q4 (history), Q7/Q9 local-after-merge, no push pipelines | carried into `validation-integration.md`, `adoption.md`, root/branch templates; legacy post-merge handoff and supervisor sections folded; `parameters.example.json` `FINAL_VALIDATION_PLACEMENT` names the exact merged SHA and no automatic post-merge/release-branch/tag-push pipeline; `installation.md` states the package installs no CI rule and carries these defaults |
| Q5 optional Claude supervisor + E1 | `roles-lifecycle.md`, templates (carried); legacy supervisor paragraph folded into `supervision.md`; template default-role paragraphs labelled "Default role assignment (unless the owner explicitly selects the optional Claude supervisor mode below)" |
| Q6 public tips | public main `ebecfbd4` policy files compared line by line with the package (`r3/public-policy-check.txt`): absent reference/template lines are only 1.x link/layout forms (three-skill names, `.agents/skills/sdlc-management` links) and the two intentional reconciliations below; no policy text missing |
| Q8 MR48 Vestive root | governance resolved to root `.claude/skills/vestive-design`; the package has no Vestive reference; protection set includes `.claude/skills/vestive-design` |
| Q10 RC promotion + owner RC auto-merge + optional RC Duo | carried sections in `validation-integration.md`, `adoption.md`, AGENTS/CLAUDE/BRANCH-WORKFLOW templates; older wording reconciled: BRANCH-WORKFLOW "wait for explicit owner promotion instruction (the manual default) unless the owner explicitly selected auto-merge"; the promotion Duo-waiver sentences (validation-integration, template) now say optional advisory review, planned follow-ups, a new fully qualified RC for shipped changes, never editing/moving an RC tag |
| Q11 direct owner replies | `coordination.md`, CLAUDE/handoff templates (carried); legacy handoff contract and supervisor section folded into `handoff.md`/`supervision.md`; `watch.py` unchanged |
| Q12 merge rule record | `adoption.md` permissions: a machine-local merge-command allow rule is a personal `settings.local.json` exception and grants no merge authority; feature auto-merge default, owner-selected RC auto-merge (manual by default) and owner holds stay authority decisions |

Also converted two staging-added BRANCH-WORKFLOW template links and one CLAUDE template link to the
`{{SKILL_ROOT}}` code-span form. `G055-legacy-path-map.md`: no path changed.

### Focused checks (candidate `fab89d1b`, base `c29ff777`; host `python3 -B`)

| Command | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `python3 -B -m unittest skills/restruct-sdlc/scripts/test_check_sdlc.py` | 1 | invocation-form error (module not importable by path), no test ran; rerun in the established form | — |
| `python3 -B -m unittest discover -s skills/restruct-sdlc/scripts -p test_check_sdlc.py` | 0 | Ran 47, OK | `a5894ef0…f5c8a` |
| `check_sdlc.py package skills/restruct-sdlc` | 0 | P1–P6, P3t pass | `860c9a90…d09d9` |
| `check_sdlc.py adoption . --release v0 --capabilities full` | 1 | expected; vs the V5 run 0 findings removed, 4 added, all in staging-imported Codex-owned/historical records (G055 Goal/Spec and G054 Spec legacy `complete-repair-inventory` link; G058 Result `{{.Names}}`); active: CLAUDE.md 16, SUPERVISOR-HANDOFF 1, PROGRESS 1 (Codex) | `e43a0002…be43a1` |
| `wc -l skills/restruct-sdlc/SKILL.md` | 0 | 119 (unchanged) | — |
| `… -p test_consolidation.py -k TestInstallMigrate -k TestRouting -v` | 0 | Ran 10, OK | `de746ef1…5c8a` |
| scope: changed paths vs `c29ff777` against the 28 patterns | 0 | 56 paths, 0 outside | — |
| protection diff (r3 §4 set + CLAUDE.md, SUPERVISOR-HANDOFF, PROGRESS, ROADMAP, Goal/Spec, settings, README) vs `c29ff777` | 0 | empty | — |

Not rerun (unchanged inputs): `test_handoff.py`, `test_watch.py` (handoff.py/watch.py unchanged), other
`test_consolidation` classes, compat runs. Worktree clean after each run.

### V5 stage refresh (stage only, outcome S)

Source `git archive fab89d1b skills/restruct-sdlc` (`7be86b01…c74a`); README base = fresh public main README
(`ec39ef94…`, equal to the sibling); same sections text. Stage `r3/export-stage`: 42 MANIFEST rows,
`da269a267a1e8eddaec0d60f437eb3164db9d99827040027c4b262cb3ab59944`; README `4fac1b68…864b` and `readme.diff`
`cd844add…e6ce` unchanged; the five approved adaptations unchanged in content (`adapted.diff` `d27969d3…90e7`; only the
supervision.md hunk offset moved from 132 to 136); `v5checks.json` `0e764724…3d4d` (35 verbatim `cmp`-equal to
`fab89d1b`, 5 adapted, no unexpected differences, hygiene clean, same notices, 43 files). Differences from the V5r
stage: MANIFEST plus 15 prose/template files of the fold; scripts unchanged. Private literal list regenerated
(58 entries, same composition; values not printed).

| Check (stage's own `check_sdlc.py`) | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `package STAGE --literals literals.txt --public-remote <public URL>` | 1 | P1–P5, P3t, P7 pass; P6 only `README.md` literal #56 — the existing G053 owner-prose exception | `0c0497fa…ce28` |
| same without `--literals` | 0 | P1–P7 pass | `c306dd0d…2878` |
| same without `--public-remote` (negative) | 1 | only `README.md: hosted-remote` | `201aeffb…c2ef` |
| gitleaks v8.30.1, validate.sh tar-on-stdin form, `--network none`, `--pull never`, under `local-harness.lock` | 0 | report `[]`, pipe `0 0 0`, "no leaks found" | report `37517e5f…b570`, log `7df664c2…9b3b9a` |

Resources: `bash infra/ci/docker-desktop-mounts.sh` first (15%); lock held by a background `flock -n … sleep 900`
(holder verified) and released after the scan; the orphaned `sleep` child kept the lock fd after the flock process
ended, so it was ended too and the lock verified free. No container left. Exported suites not rerun: scripts are
byte-unchanged since `d46af97a` (V5r evidence reused); template rendering is covered by TestRouting on the source.
No sibling or public write, no install, publish or cutover.

### Cycle and refusals

Attempt 1 continues; one invocation-form rerun (not a product failure); every other first run passed; correction
passes unchanged. The worktree guard refused compound forms (a `cd && git` chain, a git loop, a `$HOME`-computed
python argument, a curl chain to the public raw host, a checker command with the public URL followed by `; echo`, a
heredoc Result edit); each was rerun as plain single commands or with the Edit tool. One mistyped lease SHA was
rejected by the server (stale info) and retried with the verified full SHA. No native permission refusal.

### Transfer state

Pending, none performed by the writer: (1) Codex-owned CLAUDE.md 16 active links, PROGRESS 1 and the SUPERVISOR-HANDOFF
prefix retarget at the safe handoff; (2) owner README wording (intro line 3); (3) M1 genuine discovery (procedure
above); (4) F1 on a fresh fetch/rebase/reconcile; (5) G8a/G8b and code-binding approval; (6) merge held for G058
final ACCEPT; no MR. No escalation.

### Addendum — owner RC cancellation waiver (staging `a810f7b9`)

Coordinator relay: staging advanced to `2e860f00`/`a810f7b98445388abdb590d5d4aa18ee47fa3007` (r3 review SHA-256 now
`bab3473fb28b3a885bf360b267e1f0eca1efa42ffd85cf495e37dbac042dbbc6`; Spec +6 lines; public main `1b641907`). Branch rebased
cleanly onto `a810f7b9` (no conflicts). The per-RC owner cancellation waiver section is carried by rename into
`validation-integration.md`, `adoption.md` and the BRANCH-WORKFLOW/AGENTS/CLAUDE templates (carry check: all
staging-added lines present except the five intentional conversions/reconciliations above). Net reconciliation
(commit `003345b2`): `CANCELED_PIPELINE_SATISFIES` in `parameters.example.json`, the adoption parameter list and the
validation table now state that the default stays `false` and a waiver for one named RC promotion pipeline is
separate and changes no default (the old "post-merge pipeline" description was also obsolete under Q9).
Candidate **`003345b2c8ef062fb525ee44f650f07a7e77bd9b`** on base **`a810f7b9`**.

| Affected check (rerun only) | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `check_sdlc.py package skills/restruct-sdlc` | 0 | P1–P6, P3t pass | log `package-2.log` |
| `… -p test_consolidation.py -k TestInstallMigrate -k TestRouting` (parameters/template rendering) | 0 | Ran 10, OK | log `test_consolidation-2.log` |
| `check_sdlc.py adoption . --release v0 --capabilities full` | 1 | expected; 0 added, 0 removed vs the run above | log `adoption-2.log` |
| stage `r3/export-stage-2` from `003345b2`, README base = public main `1b641907` README (`ec39ef94…`, unchanged) | 0 | MANIFEST 42 rows `fba537eb5c819dacc9da177c504f36487aff0b3053933e47e83340cd003ae3e9`; vs the first refresh only MANIFEST + 6 prose/template files differ; `adapted.diff` `d27969d3…90e7`, `readme.diff` `cd844add…e6ce`, `v5checks.json` `0e764724…3d4d` unchanged | — |
| stage `package --literals --public-remote` | 1 | P6 only `README.md` literal #56 (owner-prose exception) | `32d7057b…434c7d` |
| stage `package --public-remote` | 0 | P1–P7 pass | `6f7eb6a6…bd747` |
| stage `package` (negative) | 1 | only `README.md: hosted-remote` | `980f6791…c165` |
| gitleaks, `flock -n local-harness.lock bash gitleaks.sh …` (lock held for exactly the scan) | 0 | report `[]`, pipe `0 0 0`, "no leaks found"; lock free afterwards, no container | report `37517e5f…b570`, log `5deda216…260e` |
| scope / protection vs `a810f7b9` | 0 | 57 paths (Result now included), 0 outside 28 patterns; protection diff empty | — |

`test_check_sdlc.py` not rerun: checker and tests unchanged; prose links are covered by the package check. Sibling
unchanged (`cb998602`, index `6c3444d8…`, MANIFEST `99f153fb…`, README `ec39ef94…`). Same transfer state; no escalation.

### Addendum — bounded reconciliation onto staging `e34d6d4a` (2026-10-19)

Release: coordinator relay of Codex handoff `d851d7a3` (review `G060-plan-codex-review.md`, staging `e34d6d4a`); same
owner, worktree, Plan, scope and budget (attempt 1 of B5 continues; no reset). Asserted branch
`g055-restruct-sdlc-consolidation` and HEAD `d7139b79` = origin before rebasing onto
**`e34d6d4aa0ca3e5d6c4923e6fafd4672129d278d`**. Constraint honoured: a G060 full local run holds the harness lock and the
single build slot, so no image build, DB/Compose/browser run, Docker container or lock was used.

Staging deltas since `a810f7b9` carried by rename into the canonical package: reviewed-merge classifier guidance
(`88e92146`), private per-run DB test scheduling (`4ce22d4e`, `7f4fb59e`), private browser/Compose runs sharing the
pool of 2 with sequential builds (`26a0a720`), reusable project test permission rules with label `platform.test_run`
and prefix `platform-test-` as adopter examples (`b539e615`), RC source selection at cut time; affected:
`references/adoption.md`, `references/validation-integration.md`, BRANCH-WORKFLOW/AGENTS/CLAUDE templates. Carry
check: every staging-added line under `skills/sdlc-management/**` is present except the five intentional
conversions/reconciliations already recorded. Conflict: `docs/agents/CLAUDE-PERMISSIONS.md` resolved to the staging
text with its four new links retargeted to `skills/restruct-sdlc/`; two new AGENTS.md links retargeted the same way;
the two new root-template reviewed-merge links converted to the `{{SKILL_ROOT}}` form (commit `fa4d4fb2`). The two
Result-only commits resolved to the staging Result (byte-equal to `d7139b79`; Codex import `c827e4b1`); the last
dropped as already upstream. All link targets (`#reviewed-merges-and-classifier-false-positives`,
`#private-per-run-database-test-scheduling`, `#private-browser-runs-and-sequential-image-builds`,
`#reusable-private-test-permission-rules`, `#rc-source-selection-at-cut-time`) exist as headings in the package.

| Check (lock-free, no Docker/build) | Exit | Outcome |
|---|---|---|
| `python3 -B -m unittest discover -s skills/restruct-sdlc/scripts -p test_check_sdlc.py` | 0 | Ran 47, OK |
| `check_sdlc.py package skills/restruct-sdlc` | 0 | P1–P6, P3t pass |
| `… -p test_consolidation.py -k TestInstallMigrate -k TestRouting` | 0 | Ran 10, OK |
| `check_sdlc.py adoption . --release v0 --capabilities full` | 1 | expected; 0 removed, 11 added, all legacy `skills/sdlc-management` links in Codex-owned files: CLAUDE.md +2 (now 18 active), G055 Goal +4, G055 Spec +5 |
| stage `r3/export-stage-3` from `fa4d4fb2`, README base = fresh public main `a2223861` README (`ec39ef94…`, unchanged) | 0 | MANIFEST 42 rows `9d4b1c4ce11f1cc64b790b0046f09f236506c1546d353d39bf9bbf66de23d951`; `adapted.diff` `d27969d3…`, `readme.diff` `cd844add…`, `v5checks.json` `0e764724…` unchanged |
| stage `package --literals --public-remote` / `--public-remote` / negative | 1 / 0 / 1 | P6 only README literal #56 (owner-prose exception) / P1–P7 pass / only `hosted-remote` (`0d39b53d…`, `9d3ac8f6…`, `bd2f6100…`) |
| scope / protection vs `e34d6d4a` | 0 | 56 paths, 0 outside the 28 patterns; protection diff (incl. CLAUDE.md, SUPERVISOR-HANDOFF, PROGRESS, Goal/Spec, settings) empty |

**Pending (needs Docker or the lock, deferred while G060 holds them):** gitleaks on `export-stage-3` under
`local-harness.lock`. Prior gitleaks evidence stays bound to `003345b2`'s stage. Then M1, F1, G8a/G8b, Codex links
(CLAUDE.md now 18, PROGRESS 1, SUPERVISOR-HANDOFF prefix), README wording; merge held (G055 completion waits for G060
final ACCEPT and G058 holds as recorded). G020 planning held. No deviation beyond the gitleaks deferral.

### Addendum — amendment 01 (Q13) reconciliation onto staging `7afee895` (2026-10-20)

Authority: accepted amendment 01 (not included)
(SHA-256 `badb13eb…bb0e3`) and its Codex review (not included) (`3f6b5be7…6db`),
imported on staging; the planning branch was not merged. Same task, attempt 1 of B5, correction passes and 28-pattern
scope; no reset. Writer: fresh continuation (context handoff, single write owner) in native isolated worktree
`.claude/worktrees/agent-<id>`, local branch `g055-restruct-w2` cut from `origin/g055-restruct-sdlc-consolidation`
because the previous writer's worktree `agent-<id>` still holds the branch name (left untouched, stale at
`4170e8f9`). Effort dispatched Opus 5.5 high; runtime unknown. No Docker, image build, lock, DB/Compose/browser run.

**Bindings and rebase.** Asserted HEAD `4170e8f9` = origin before rebasing onto
**`7afee895b45f3df0e51418a56ac7b3d7d083dfad`** (20 staging commits since `e34d6d4a`, incl. G060 scope amendment
`7d2d6221` and the amendment import). `origin/staging` is an ancestor of the candidate (`git merge-base --is-ancestor`,
exit 0). Inputs at `7afee895`, unchanged by G055: allow-list `e7c2bebc…b64d` (blob `7e34da75`, 28 patterns), inputs r3
`7f151858…7f1c` (blob `ab2ae207`), Plan r3 `113fdef1…0dba` (blob `ed388070`), r3 review `bab3473f…bbc6`, Goal
`3746bb19…ab9f` (blob `ecdfc683`), Spec `7c459d68…a197` (blob `21ddff22`), `.codex/config.toml` blob `56617a06`
(Goal/Spec/config blobs equal the amendment table). Binding change: none beyond the accepted Goal/Spec owner amendments.

Conflicts (resolved toward staging text, G055 structure kept): `SKILL.md` (entrypoint kept; the staging coordinator-model
lines live in `budgets.md`, carried by rename); `.agents/skills/repository-supervisor/SKILL.md` delete/modify (stays
retired; delta folded below); handoff/root templates and `AGENTS.md`/`docs/agents/CLAUDE-SESSION-HANDOFF.md` (150k/200k
staging numbers with the `{{SKILL_ROOT}}` or `skills/restruct-sdlc/` link forms; staging's 2026-10-20 context amendment
kept beside the G055 cutover record); `docs/foundation/OPERATIONS.md` (Restruct row plus staging's D24 Vestive rows:
`.agents/skills/vestive-design/` canonical, `.claude/skills/vestive-design` Claude alias). The Result-only commit
`4170e8f9` dropped as already upstream (Result byte-equal to staging before this addendum).

**Fold commit `9c17dd158598da75bb4e645e67b8586088a4fdd6`** (rebased base `28cd2129`):

| Item | Change |
|---|---|
| T1 `coordination.md` | (a)–(c) exact; TOML block = the four checked-in keys; (d) applied as "The model IDs above are this repository's checked-in defaults (`15243c09`); portable adopters choose their supported models explicitly." — the leading "Portable adopters choose their supported child model explicitly;" was dropped to avoid duplicating the new clause (wording deviation, same meaning) |
| T2 `adoption.md` | exact replacement plus the OS001 export-boundary sentence |
| T3 root `AGENTS.md.tmpl`/`CLAUDE.md.tmpl` | exact replacement plus "See model/delegation defaults (`{{SKILL_ROOT}}/references/coordination.md#codex-supervisor-model-and-delegation-defaults--owner-direction-2026-10-20`)" |
| T4 `supervision.md` | legacy supervisor staging deltas folded: §Coordination economy (verbatim + two intra-package links) and §Codex routine work and Sol delegates (T3 text, intra-package link) |
| T6 | root `AGENTS.md`/`CLAUDE.md` Luna/Sol paragraph verified as Codex's staging text, not edited; `.codex/config.toml`, Goal, Spec untouched |
| Coordinator default | `handoff.md` §Model-aware effort and §Fable fallback, `supervision.md` §Fable fallback: "coordination remain high" → Sonnet 5.5 high default, Opus 5.5 high selectable (staging wording) |
| Self-compaction | `coordination.md` §Enforce existing usage limits: "If compaction is owner-only…" → agents self-compact under the authorized section; report once only when the sole control is operator-only |
| Link forms | staging-added template links `coordination economy (not included)` and `safe self-compaction (not included)` converted to the code-span form (the markdown form broke legacy-format rendering: A2 target missing); two new `AGENTS.md` links retargeted to `skills/restruct-sdlc/` |

Carry check (`r3/carry_check.py . e34d6d4a origin/staging skills/sdlc-management/=skills/restruct-sdlc/
.agents/skills/repository-supervisor/SKILL.md=skills/restruct-sdlc/references/supervision.md`): 16 absent lines, all
intentional — 5 `SKILL.md` model lines (entrypoint; carried in `budgets.md`), T1 ×3, T2 ×1, template link-form ×4 and T3
×2, T4 ×1. Context economy 150k/200k, authorized self-compaction and coordinator default agree across package, root
templates and the handoff runbook (grep for `300k|350k|cheaper main|owner-controlled|selected example`: only superseding
history phrases and unrelated RC/final-tag text remain). Vestive: the package has no Vestive reference; both Vestive
paths are protected and unchanged.

| Check (lock-free, host `python3 -B`) | Exit | Outcome |
|---|---|---|
| `check_sdlc.py package skills/restruct-sdlc` | 0 | P1–P6, P3t pass |
| `… -p test_consolidation.py -k TestInstallMigrate -k TestRouting` (first run, before link-form fix) | 1 | 2 failures: migration/rollback legacy preflight, A2 on the two staging `{{SKILL_ROOT}}` markdown links (+ the new T3 link); reproduced with `checks-a01/debug_legacy.py` |
| same, after the code-span conversion | 0 | Ran 10, OK |
| `… -p test_consolidation.py` (all classes; rendered template content changed) | 0 | Ran 12, OK |
| `… -p test_check_sdlc.py` (fixtures render templates) | 0 | Ran 47, OK |
| `wc -l skills/restruct-sdlc/SKILL.md` | 0 | 119 |
| `check_sdlc.py adoption . --release v0 --capabilities full` | 1 | expected; vs `adoption-3.log`: 0 removed, 5 added, all Codex-owned/historical: CLAUDE.md +2 (now 20), G055 Goal +1, Spec +1, amendment 01 quoted `{{SKILL_ROOT}}` (A3) |
| scope: `git diff --name-only origin/staging HEAD` vs 28 patterns (`checks-a01/scope_check.py`) | 0 | 56 paths, 0 outside (57 with this Result) |
| protection diff `git diff --exit-code origin/staging HEAD -- .gitlab-ci.yml infra core web intelligence contracts skills/saas-foundation .agents/skills/vestive-design .claude/skills/vestive-design management/BUDGET_POLICY.md .gitignore .codex/config.toml CLAUDE.md management/SUPERVISOR-HANDOFF.md management/v0/PROGRESS.md management/v0/ROADMAP.md management/v0/goals management/v0/specs management/v0/plans .claude/settings.json README.md` | 0 | empty (incl. `.codex/config.toml`, T6) |
| public main read-only: `git ls-remote <public URL> refs/heads/main` + shallow clone | 0 | **`fb5bdabbfb23987aaee75d20c948dade3b11b3f5`** (the coordinator-model publication; ancestry to `4df892b6` not checked in the depth-1 clone); README `ec39ef94…` and LICENSE `af5fb221…` unchanged; MANIFEST `c75ab964…` |
| Q6 public policy comparison (`r3/public_policy_check.py fb5bdabb …`) vs the `ebecfbd4` run | 0 | +31 absent lines, all intentional: public's older/alternative coordinator-model phrasing (package carries staging wording), T1–T4 conversions, the `003345b2` `CANCELED_PIPELINE_SATISFIES` reconciliation, link-form conversions; no policy text missing |

**Native refusal (reported once, not retried):** building `r3/export-stage-4` with
`python3 -I v5tools/build_stage.py source-9c17dd15.tar public-fb5bdabb-clone/README.md v5tools/readme-sections.md
r3/export-stage-4` was denied by the auto-mode classifier (no explanation). `source-9c17dd15.tar` exists; no stage was
written.

**Pending:** (1) T5 stage refresh `export-stage-4` and its stage checks (`package --literals --public-remote`,
`--public-remote`, negative, `v5checks`) — blocked by the refusal above; needs the coordinator's permission
disposition. (2) gitleaks on the stage under `local-harness.lock` (coordinator; not run). (3) Deferred to the atomic
cutover, record only (Codex-owned, not executed): legacy `skills/sdlc-management` links now CLAUDE.md 20, G055 Goal +5,
Spec +6 since `a810f7b9` (previously recorded 11 → 15), PROGRESS 1, SUPERVISOR-HANDOFF prefix. (4) M1, F1, G8a/G8b,
code-binding approval, owner README wording — unchanged. (5) Integration waits for G060 final ACCEPT; no MR.

### Addendum — w3 resume after RC3 MR !62: staging `5df34735` reconciliation, stage-4 and gitleaks (2026-10-20)

Release: coordinator dispatch after RC3 promotion MR !62 opened (owner sequence; PROGRESS at `5df34735`). Same task,
attempt 1 of B5 (checkpoint after attempt 3), correction passes and 28-pattern scope; no reset, no Plan change.
`release/0.1.0-rc.3`, tag `0.1.0-rc.3` and MR !62 not touched. Writer: fresh continuation (single write owner) in native
isolated worktree `.claude/worktrees/agent-<id>`, local branch `g055-restruct-w3` (the name
`g055-restruct-sdlc-consolidation` is held by stale worktree `agent-<id>`; `agent-<id>` also
stale; both untouched). Effort dispatched Opus 5.5 high; runtime unknown. Evidence `…/g055-impl-a1/r3/w3/` (0700).

**Staging reconciliation.** Asserted HEAD `7662fdf3f83b3634aa37c0fbe6a65f4dbaaf1ee4` = origin before rebasing onto
`23a323c833fdee908219e6dc37dd0931b130efa5` (19 commits since `7afee895`: G060 tests/logrecord schema, G060 records,
PROGRESS/ROADMAP; no path G055 changes) → `9d3dff4a6011224b73597a8f6d582800727a4734`, pushed with
`--force-with-lease=g055-restruct-sdlc-consolidation:7662fdf3…`. Staging then advanced to
**`5df347356740b0e7f0804538b8028a92b67ce3c4`** (PROGRESS + G060 review status only); rebased again →
**`fce062db83d9dbd8a736a65646f9e6495b053f81`**, pushed with the lease bound to `9d3dff4a`. Both rebases conflict-free;
`git diff 7662fdf3 9d3dff4a` equals the staging delta exactly (no G055 content change); package tree
`f1369cfb` = `9c17dd15`'s. `git merge-base --is-ancestor origin/staging HEAD` exit 0. No branch pipeline expected/observed
beyond the push itself (no MR).

**Bindings (unchanged; no delta).** Allow-list `e7c2bebc…b64d` (blob `7e34da75`, 28 patterns), inputs r3 `7f151858…7f1c`
(blob `ab2ae207`), Plan r3 `113fdef1…0dba` (blob `ed388070`), r3 review `bab3473f…bbc6`, amendment 01 `badb13eb…bb0e3`,
its review `3f6b5be7…36db`, Goal `3746bb19…ab9f` (blob `ecdfc683`), Spec `7c459d68…a197` (blob `21ddff22`),
`.codex/config.toml` blob `56617a06` (= `7afee895`).

| Check (candidate `fce062db`, base `5df34735`) | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| scope: `git diff --name-only origin/staging HEAD` vs 28 patterns (`checks-a01/scope_check.py`) | 0 | 57 paths, 0 outside, 0 protected (Result included) | — |
| protection diff (amendment 01 list incl. `.codex/config.toml`, CLAUDE.md, SUPERVISOR-HANDOFF, PROGRESS, ROADMAP, goals/specs/plans, settings, README) | 0 | empty | — |
| `check_sdlc.py adoption . --release v0 --capabilities full` (staging added management records) | 1 | expected; vs `checks-a01/adoption.json`: 196 → 196 findings, 0 added, 0 removed | `24d5e2cc…75a6` |
| other source suites (`test_check_sdlc`, `test_consolidation`, `package`) | — | not rerun: package tree byte-identical to the a01 candidate | — |

**T5 stage refresh `export-stage-4` (stage only, outcome S).** `git archive --output=w3/source-fce062db.tar fce062db
skills/restruct-sdlc` (`8c0a74df…249d`); `python3 -I v5tools/build_stage.py w3/source-fce062db.tar
public-fb5bdabb-clone/README.md v5tools/readme-sections.md r3/export-stage-4` — **ran this time (no refusal)**. Public main
re-read: `git ls-remote <public URL> refs/heads/main` = `fb5bdabb…` (unchanged); README base `ec39ef94…` unchanged.
MANIFEST 42 rows **`55adfdf810c0f2c592b7603002f626d1007362a23860164ab490855ed97a65b0`**; vs stage-3 only MANIFEST and nine
prose/template files differ (adoption, budgets, coordination, handoff*, supervision* references; CLAUDE-SESSION-HANDOFF,
BUDGET_POLICY, root AGENTS/CLAUDE templates; * = adapted); scripts byte-unchanged. README `4fac1b68…864b`.

| Stage check | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `v5checks.py STAGE <worktree> <sibling> v5checks-4` | 0 | 35 verbatim `cmp`-equal to `fce062db`, 5 adapted, 0 unexpected, README head/tail equal, hygiene clean, same notices, 43 files; `adapted.diff` `d27969d3…`, `readme.diff` `cd844add…`, `v5checks.json` `0e764724…` = stage-3 | as stated |
| stage `package --literals --public-remote` | 1 | P1–P5, P3t, P7 pass; P6: README literal #56 (owner-prose exception) **and new: `references/adoption.md` literal #58** | `d827928c…c16d` |
| stage `package --public-remote` | 0 | P1–P7 pass | `452a17e4…b7a` |
| stage `package` (negative) | 1 | only `README.md: hosted-remote` | `ec030b48…7b3e` |
| gitleaks v8.30.1 (`sha256:c00b6bd0…`), `flock -n local-harness.lock bash v5tools/gitleaks.sh export-stage-4 infra/security/gitleaks.toml r3/gitleaks-4` (`--network none`, `--pull never`) | 0 | pipe `0 0 0`, report `[]`, "no leaks found" (~797 KB) | report `37517e5f…b570`, log `cc48b396…d2ce` |

Resources: `bash infra/ci/docker-desktop-mounts.sh` first (17%); lock held by `flock -n` for exactly the scan and verified
free afterwards; no container left (`--rm`). This scan supersedes the stage-3 deferral (stage-4 contains every stage-3
file plus the a01 prose). Sibling read-only before/after identical: `refs/heads/main` `cb998602`, `.git/index`
`6c3444d8…`, MANIFEST `99f153fb…`, README `ec39ef94…`, LICENSE `af5fb221…`; no Git command there.

**New finding (question for Codex; not self-resolved).** Amendment 01 T2's exact accepted sentence ends "…the exporting
Goal's responsibility (here OS001); …" (`references/adoption.md` L376). `OS001` is private literal #58 (G053 list,
same term that `test_watch.py` already adapts to `AB001` at export), so the stage fails P6 beyond the owner-prose
exception. Not a source or adaptation change by the writer (T2 is exact Plan text; the five adaptations are a Codex
disposition). Options: **(A)** sixth export adaptation `(here OS001)` dropped at export (source keeps T2 verbatim) —
preview `w3/export-stage-4a-proposal` built with `w3/tools/build_stage_a6.py` (`7d5d8c31…4fbb`, copy of `build_stage.py` +
that one pair): exactly one line differs, MANIFEST row `adoption.md` → `adapted`, MANIFEST `00f11c4c…c697`; stage
`package --literals --public-remote` leaves only README #56 (`a6b3b7f0…39fa`); or **(B)** remove "(here OS001)" at source
(T2 wording deviation; candidate SHA changes; affected checks: package, stage refresh). Proposal stage is evidence only,
not the T5 stage of record.

**Holds / pending (unchanged gates).** (1) Codex disposition of the #58 finding (A or B), then the stage of record and
its P6 row. (2) M1 owner listings: procedure above (`### Final writer handoff`), to be requested once at full-candidate
readiness — candidate SHA depends on (1). (3) F1 after (1)–(2) on a fresh fetch/rebase. (4) G8a/G8b and code-binding
approval after Codex ACCEPT/MR (Plan r3 §5 step 6). (5) Owner README intro wording (line 3). (6) Codex-owned cutover
link retargets, record only: adoption findings unchanged (CLAUDE.md 20 legacy links, Goal/Spec +5/+6 since `a810f7b9`,
PROGRESS 1, SUPERVISOR-HANDOFF prefix). (7) G060 now finally ACCEPTed (staging `23a323c8`); G055 integration still waits
for Codex release; no MR. G020 held.

Native refusals: none (the stage-4 build refused in the a01 round succeeded unchanged). Cleanup: no container, image or
lock left; task-owned evidence kept for review (`r3/export-stage-4`, `r3/v5checks-4`, `r3/gitleaks-4`, `r3/w3/**` incl.
`export-stage-4a-proposal`); stale worktrees `agent-<id>`/`agent-<id>` left for closeout cleanup.

### Addendum — w4: T2 correction 01 (`(here OS001)` removed), staging `2e9c78a1` reconciliation, stage-5 and gitleaks (2026-10-20)

Authority: accepted Plan correction 01 (not included)
(SHA-256 `a788749b536276b5ea3d9711a947672e97b96cad73b20b0518e5f3327517dafe`) and its
Codex review (not included) (ACCEPT, SHA-256
`c562bab7b6e6f8c119d06036cdb1176516b273fff058e364ac1840b1815c5ea8`), both imported on staging; the planning branch was
not merged. Same task, attempt 1 of B5, correction passes and 28-pattern scope; no reset, no other Plan change.
`release/0.1.0-rc.3`, tag `0.1.0-rc.3` and MR !62 not touched. Writer: fresh continuation (single write owner) in native
isolated worktree `.claude/worktrees/agent-<id>`, local branch `g055-restruct-w4` (the stale w2/w3 worktrees
and branches left untouched). Effort dispatched Opus 5.5 high; runtime unknown. Evidence `…/g055-impl-a1/r3/w4/` (0700).

**Staging reconciliation.** Asserted HEAD `cfb1a2d925a3eab28e9c69657d8bf84f8a6e76bb` = origin, then rebased onto
**`2e9c78a18369dce295a2d728188e641dd935d6ad`** (2 commits since `5df34735`: `699e0bbd` export privacy disposition,
`2e9c78a1` correction import; paths PROGRESS + the correction + its review, none G055 changes) →
**`8c7ee3f19b1471d52daf2749dce3efa894dac04f`**, conflict-free; `git diff cfb1a2d9 8c7ee3f1` equals the staging delta
exactly (3 files, +98/−1). Pushed with `--force-with-lease=g055-restruct-sdlc-consolidation:cfb1a2d9…`.
`git merge-base --is-ancestor origin/staging HEAD` exit 0 (before the checks and again before the checkpoint push).

**Bindings (unchanged at HEAD and staging).** Allow-list
`e7c2bebc7d9859f692dd1bd67544c31a7b3906789ed0d9482a9631a0c125b64d` (blob `7e34da75`, 28 patterns), inputs r3
`7f151858…7f1c` (blob `ab2ae207`), Plan r3 `113fdef1…0dba` (blob `ed388070`), amendment 01 `badb13eb…bb0e3`,
`.codex/config.toml` blob `56617a06` (untouched). This round's only changed path is
`skills/restruct-sdlc/references/adoption.md` (`git diff --name-only 8c7ee3f1 HEAD`); root `AGENTS.md`/`CLAUDE.md` not
touched.

**Deviation (authorized).** T2 correction 01 (Codex option B, source wording): amendment 01 T2's sentence loses only
"(here OS001)" and its preceding space; the exporting-Goal ownership rule is unchanged; no sixth export adaptation (the
five accepted adaptations and the README #56 owner-prose exception unchanged). Commit
**`fbd4179826abcfc1b5a9061def950ef537fc5a1a`**, exact one-line diff (`references/adoption.md` L376, word diff):

```
-…is the exporting Goal's responsibility (here OS001); this package documents the pattern and installs no `.codex/config.toml`.
+…is the exporting Goal's responsibility; this package documents the pattern and installs no `.codex/config.toml`.
```

The package now names `OS001` only in `scripts/test_watch.py` L281, the existing export adaptation (→ `AB001`).

| Check (candidate `fbd41798`, base `2e9c78a1`) | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `python3 -B skills/restruct-sdlc/scripts/check_sdlc.py package skills/restruct-sdlc` | 0 | P1–P6, P3t pass, 0 findings | — |
| source P6 literal scan: same `--literals r3/literals.txt` | 1 | expected: 8 P6 hits, all in the five export-adapted files (handoff, installation, local-session-handoff #1/#2; supervision #54; `test_watch.py` #58); **no `adoption.md` hit** (the #58 hit is gone) | `29baa1a7…1ee0` |
| `git archive --output=w4/source-fbd41798.tar fbd41798 skills/restruct-sdlc` | 0 | source tar | `17732ba5…64cb` |
| public main re-read: `git ls-remote <public URL> refs/heads/main` | 0 | `fb5bdabb…` unchanged; README base `ec39ef94…` unchanged | — |
| `python3 -I v5tools/build_stage.py w4/source-fbd41798.tar r3/public-fb5bdabb-clone/README.md v5tools/readme-sections.md r3/export-stage-5` (builder `f8a00c23…`, sections `276b14b1…`) | 0 | 42 MANIFEST rows **`a75669f5675e0b9d6874732ef47f4a85ee087577c5e4911d145b9ef7a9d8d89b`**; `diff -r export-stage-4 export-stage-5`: only `adoption.md` L376 and its MANIFEST row (still `verbatim`) differ; README `4fac1b68…864b` unchanged | as stated |
| `python3 -I v5tools/v5checks.py r3/export-stage-5 <worktree> <sibling> r3/v5checks-5` | 0 | 35 verbatim `cmp`-equal to `fbd41798`, 5 adapted, 0 unexpected, 0 missing, README head/tail equal, hygiene clean, same notices, 43 files; sibling compare 1/2/40/41 | `v5checks.json` `0e764724…1f23d`, `adapted.diff` `d27969d3…90e7`, `readme.diff` `cd844add…e6ce` (all = stage-4) |
| stage `package --literals r3/literals.txt --public-remote <public URL>` | 1 | P1–P5, P3t, P7 pass; **P6 only `README.md` literal #56** (accepted owner-prose exception); #58 gone | `8368e3ee…6c2e` |
| stage `package --public-remote <public URL>` | 0 | P1–P7 pass, 0 findings | `da070572…1ed31` |
| stage `package` (negative) | 1 | only `README.md: forbidden content: hosted-remote` | `e3543ed7…8aa7` |
| scope: `git diff --name-only origin/staging HEAD` vs 28 patterns (`checks-a01/scope_check.py`) | 0 | 57 paths, 0 outside, 0 protected (before this Result edit; the Result is already among them) | — |
| protection diff (amendment 01 list incl. `.codex/config.toml`, CLAUDE.md, SUPERVISOR-HANDOFF, PROGRESS, ROADMAP, goals/specs/plans, settings, README; `git diff --exit-code`) | 0 | empty | — |
| `check_sdlc.py adoption . --release v0 --capabilities full` vs `w3/adoption.json` (`r3/adopt_diff.py`) | 1 | expected; **196 → 196 findings, 0 added, 0 removed** (A2–A4 fail as before) | `24d5e2cc…75a6` (byte-equal to w3) |
| gitleaks v8.30.1 (`sha256:c00b6bd0…`), `flock -n local-harness.lock bash v5tools/gitleaks.sh r3/export-stage-5 infra/security/gitleaks.toml r3/gitleaks-5` (`--network none`, `--pull never`; config `df25add6…`) | 0 | **rerun** (no reuse): pipe `0 0 0`, report `[]`, "no leaks found" (~796.5 KB) | report `37517e5f…b570`, log `a5e90330…3e52` |

Not rerun (per correction §4/review): `test_check_sdlc`, `test_handoff`, `test_watch`, `test_consolidation` (no script
or template change). Harness bookkeeping: the first `v5checks.py` call exited 1 (`FileNotFoundError`: the output
directory must pre-exist, as in earlier rounds); created `r3/v5checks-5` (0700) and reran unchanged → exit 0.

**Stage of record:** `r3/export-stage-5` (supersedes stage-4 for T5). **History preserved:** `r3/export-stage-4`,
`r3/v5checks-4`, `r3/gitleaks-4` and `w3/stage4-*.json` (incl. the #58 P6 failure `d827928c…c16d`) and
`w3/export-stage-4a-proposal` (option-A evidence, not adopted) are kept unmodified.

Resources: `bash infra/ci/docker-desktop-mounts.sh` first (17%); lock held by `flock -n` for exactly the scan and
verified free afterwards (`flock -n … true` exit 0); no gitleaks container left (`--rm`); no image built. Sibling
read-only before/after identical: `refs/heads/main` `cb998602`, `.git/index` `6c3444d8…`, MANIFEST `99f153fb…`, README
`ec39ef94…`, LICENSE `af5fb221…`; no Git command there.

**Holds / pending (unchanged; not attempted).** (1) M1 owner skill listings — candidate SHA now `fbd41798` plus this
Result commit; request once at full-candidate readiness. (2) F1 after M1 on a fresh fetch/rebase. (3) G8a/G8b and
code-binding approval after Codex ACCEPT/MR. (4) Owner README intro wording (line 3). (5) Codex-owned cutover link
retargets, record only: adoption findings unchanged (CLAUDE.md 20 legacy links, Goal/Spec +5/+6 since `a810f7b9`,
PROGRESS 1, SUPERVISOR-HANDOFF prefix). (6) Integration waits for Codex release; no MR. G020 held.

Native refusals: none. Cleanup: no container, image or lock left; task-owned evidence kept for review
(`r3/export-stage-5`, `r3/v5checks-5`, `r3/gitleaks-5`, `r3/w4/**`); stale worktrees `agent-<id>`,
`agent-<id>`, `agent-<id>` (w3) left for closeout cleanup.

## Attempt 2 (r4 bundle) — implementation round 1, 2026-10-20

Authority: Plan r4 (not included) (SHA-256
`8f88fe27b93aeec5a131774757ba190d5411bfa632e1ef8acdeb8886b7b495fa`), inputs r4 (not included)
(`55dee354de859c5389f6e48e1959b02495c72d9c86678257e852ea6683527f43`), executable scope
G055-scope-allowlist.txt (not included) (`114d6b7b99abd8741b0cf234d5e10945bccad3a015ccb78165da8dc0e6533d93`,
26 r4 patterns, verified) and the Codex review (not included) §"Correction 03 — ACCEPT and
execution release" with its three mandatory supplements (now SHA-256
`e89258121ef7d069433f318c5339bb408e0a549b5e2e186b09d24757be438b03` after the §"Communication clarification preserves
0.1.0 behavior" addendum). **Attempt 2 of B5** (checkpoint after attempt 3); no reset or extension; r4 correction
passes 0. Implementation only: this round claims no acceptance.

**Ownership (recorded, not asserted beyond what is verifiable).** Single implementation writer, native isolated
worktree `.claude/worktrees/agent-<id>` (agent `a8487839233c9e98b`), local branch `g055-bundle-impl`,
dispatched by the coordinator as Opus 5.5 high. Running model reported by the harness: Opus 5.5 (`claude-opus-5-5`);
runtime effort: unknown (not verifiable from inside the session). Worktree initially at `main` `5a40cdc1`; the branch
was created from origin/staging before any write (fail-closed branch/HEAD assertion).

**Staging reconciliation.** Dispatch base `6c4d22dae7d2148f4bdbb0168ae58f6dc0674d81` (verified by fetch). Before any
write the coordinator relayed staging `d6f2e8136308709af18a759e423a6b9425825052` (owner communication clarification):
its in-scope change is `skills/sdlc-management/references/coordination.md` (heading "owner direction" → "owner
clarification" plus a two-line "clarifies the existing communication workflow" paragraph), so that one execution
baseline hash is re-derived as **`be271b86352b65fab21e45ec98d08511c15f6f17377e9e2cd6b3d4141c84d46e`**; the other six
review baseline hashes (repository-supervisor SKILL, SUPERVISOR-HANDOFF/AGENTS/CLAUDE templates, root AGENTS/CLAUDE)
matched exactly. The branch was moved to `d6f2e813` (`git reset --keep`, no local changes). Before the validation
batch staging advanced to **`b8804f387ce853c7d5ab2463a868927d9313d80d`** (G062 docs); its only in-scope path is
`docs/SECURITY.md` (new paragraphs, not the retargeted L446 link). Rebased conflict-free: S1 `6b67bc60`, S2/S3
`147d1644`, S4 `fba48763`; `merge-base --is-ancestor origin/staging HEAD` holds. All validation below ran on
`fba48763` against base `b8804f38`. The inputs-JSON baseline of the 38 files at `4aeeb213` differs only in the five
owner-workflow files (expected).

### S0 attributable baseline (untouched skills at `d6f2e813`, host `python3 -B`, no Docker, no lock)

| Command | Exit | Outcome |
|---|---|---|
| `check_sdlc.py package skills/sdlc-management` | 1 | P1/P4/P5/P6 pass; P2 `220 lines exceeds 200`; P3 ×6 (siblings not adjacent); P3t ×13 (5 adopter paths, 4 undeclared `SKILL_ROOT`, 4 `{{SKILL_ROOT}}` links) — as Plan §4.2 |
| `unittest … test_check_sdlc.py` | 1 | **35 run, 13 failures** (not only the 2 the Plan anticipated): D5 P2 and D6 `SKILL_ROOT` break every exit-0 package/adoption assertion, plus `test_ignore_lines_constant_matches_template` (see finding F1) |
| `unittest … test_handoff.py` | 0 | 64 OK |
| `unittest … test_watch.py` | 0 | 84 OK |
| `check_sdlc.py package .agents/skills/claude-session-handoff` (baseline tree) | 1 | P3 ×9 and **P6 `non-synthetic full UUID` in `scripts/test_handoff.py`** (pre-existing; finding F2) |

### Implementation (Plan §2–§4.5; one writer, no delegation)

- **S1 `6b67bc60`:** `git mv` of the three skills whole into `skills/restruct-sdlc/` (old standalone paths retired in
  the same commit); thin parent `SKILL.md` (54 lines, routing table to the three children, no role/policy/helper),
  parent `LICENSE` (byte copy), parent `agents/openai.yaml`, aliases `.agents/skills/restruct-sdlc` and
  `.claude/skills/restruct-sdlc` → `../../skills/restruct-sdlc` (mode 120000; `realpath -e` inside the checkout).
- **S2/S3 `147d1644`:** Rule A/B/D6 prefix substitutions (8 files, proven prefix-only); two `sdlc-management/SKILL.md`
  sentences (still 220 lines); `adoption.md` install/layout rewrite, step 4a, fallback sentence; `check_sdlc.py`
  bundle constants; `handoff.py` `HELPER_CANDIDATES`/`helper_relative()`; `.codex/config.toml.tmpl` and 5 parameters;
  tests (D5 oracle, fixtures, +4 checker and +2 handoff cases).
- **S4 `fba48763`:** governance/runbook link retargets (Plan §4.4) and one AGENTS.md G055 relocation sentence.
- Management: this section and the path map (not included).

### Validation (candidate `fba48763`, base `b8804f38`; host `python3 -B`, temp dirs, no Docker/lock/network)

Selectors verified before running (unittest discovery of the named file; `-k` patterns listed by `-v`). Evidence root
`~/.local/state/platform-evidence/g055-r4/` (0700); hashes below.

| AC | Command | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|---|
| AC3/AC4(iii) | `python3 -B -m unittest discover -s skills/restruct-sdlc/sdlc-management/scripts -p test_check_sdlc.py -v` | 1 | **39 run, 38 pass, 1 failure = pre-existing `test_ignore_lines_constant_matches_template` (F1), unchanged from baseline**; the 4 new cases and all D5-oracle cases pass | `14e88b41…37f9` |
| AC1 (D5) | same with `-k test_real_skill_passes_all_checks -k test_bundle_passes` | 0 | 2 OK (red at S0); each asserts exit 1 with exactly the P2 finding | — |
| AC4(iii)/AC6 | `… -s skills/restruct-sdlc/claude-session-handoff/scripts -p test_handoff.py -v` | 0 | 66 OK (64 + `test_helper_relative_prefers_tracked_canonical_path`, `test_alias_invocation_resolves_same_repository`); all T14/T15 cases inside | `0bb9cbff…eff4` |
| AC4(iii) | `… -p test_watch.py -v` | 0 | 84 OK (no code change; attributable run after the import path moved) | `6bbe29e4…fe3d6e` |
| AC1/AC3(c) | `check_sdlc.py package skills/restruct-sdlc/sdlc-management --skills-root skills/restruct-sdlc` | **1** | **expected D5 result, not a pass:** exactly one finding `P2 …/sdlc-management/SKILL.md "220 lines exceeds 200"`; P1/P3/P3t/P4/P5/P6 pass (P3t now includes `.codex/config.toml`; the 13 baseline P3t and 6 P3 findings are gone) | `f3569aea…8929` |
| AC1/AC2 | `check_sdlc.py package skills/restruct-sdlc` | **1** | P1/P3/P5 pass over parent + children (parent routing links resolve); **P6: one finding, `claude-session-handoff/scripts/test_handoff.py` non-synthetic full UUID** — the pre-existing F2 content, so the Plan's "expected exit 0" is not met | `fd4157fc…be81` |
| AC4(i) | `ac4_evidence.py` (git plumbing only) | 0 | expected tree = observed tree, 44 entries each (both files `b891d1a2…f9edd`); no path left under the old roots | summary `6771d504…816d` |
| AC4(ii) | pinned-hash partition, old roots at `b8804f38` vs new root at `fba48763` | 0 | **23 hash/mode-identical = the Plan set; 15 non-identical = the classified set; 0 unmapped**; the 4 new files (parent ×3, Codex template) reported separately; also `git diff -M100% --name-status` over both old roots and the new root | partition `5a7c4201…6d9`, rename list `6e27e1fc…1b32` |
| AC4(ii) | `prefix_only.py` | 0 | the 8 Rule A/B/D6 files are reproduced byte-for-byte by applying only those substitutions; the other 7 diffs are in the patch | `b0c798cc…2f45`, patch `7361703f…b57c` |
| AC4(iv) | rule-phrase counts at actual locations (coordination, repository-supervisor, AGENTS/CLAUDE/SUPERVISOR-HANDOFF templates) | 0 | every count equal baseline vs candidate; `coordination.md` and `watch.py` R100; CLAUDE template carries no-poll + immediate nonblocking messaging, AGENTS template the live-guidance rule (and its no-poll paragraph); rendered-file assertions in `test_bootstrap_creates_missing_files` | `5b364d86…fcff` |
| AC5 | `check_sdlc.py adoption . --release v0` vs the same on the `b8804f38` tree (`adopt_delta.py`) | 1 / 1 | 68 → 173 findings at `fba48763` (172 once the path map below exists; rerun on the working tree): −25 (old installed-skill link failures, "not installed", moved A6 row) and +130: 126 in closed/historical records (goals 27, specs 79, archive 10, plans 5, results 5) that link old paths, 1 moved A6 row (F2), the path-map link (resolved by this commit) and the two Codex-owned links (SUPERVISOR-HANDOFF L92, PROGRESS L94) — finding F5 | candidate `d0440e2a…00f7`, base `73fe2975…e900` |
| AC5 | `git grep` active old-path references outside closed records | 0 | remaining only: protected inert rows (`.gitlab-ci.yml`, `infra/ci/*`, `infra/validation/run.sh`), OPERATIONS historical/inert lines, Codex-owned SUPERVISOR-HANDOFF/PROGRESS/ROADMAP, the deliberate retired-path fixture in `test_handoff.py` | — |
| scope | `scope_check.py` vs the 26 patterns (`git diff --name-status -M origin/staging HEAD` + untracked) | 0 | 51 rows, 90 paths, 0 outside | — |
| protection | Plan §6.2 `git diff --exit-code` list + SUPERVISOR-HANDOFF + plans | 0 | empty (SUPERVISOR-HANDOFF untouched) | — |
| integrity | `infra/validation/run.sh … provenance` | not run | Docker-bound (Compose preflight, docker inventory) → deferred to the coordinator's mandatory post-merge local full, as the Plan allows | — |

### Findings and deviations for Codex disposition

- **F1 (pre-existing, not fixed): `test_ignore_lines_constant_matches_template` is red on staging.** `gitignore.lines`
  gained `/.claude/settings.local.json` and `/.local/` without `check_sdlc.IGNORE_LINES`. Fixing either side changes A5
  semantics or a test outside the Plan's classified hunks, so per supplement 2 it is reported, not adapted. **Question:**
  classify a parity fix (constant += the two lines; `test_a5_ignore_lines` 3 → 5) or accept as a recorded baseline
  failure.
- **F2 (pre-existing, not fixed): strict P6 flags real-looking UUIDs in `test_handoff.py`** (`SESSION`/`OTHER`/`HANDOFF`
  constants). `check_sdlc.py package skills/restruct-sdlc` therefore exits 1 (Plan AC1 expected 0), and the AC7
  outside-bundle check will show this P6 finding beside the D5 P2. **Question:** classify a fixture-only substitution to
  synthetic `00000000-0000-4000-8000-…` values (the reviewed v2 candidate did this), an export adaptation, or a
  recorded baseline finding.
- **D-a (mechanical oracle parity):** the D5 helper is applied not only to the two named tests but also to the four
  other exit-0 package assertions on the same real child (`test_p3_links_inside_code_are_ignored`,
  `test_p3t_rendered_template_in_documented_layout_passes`, two in `test_p6_forbidden_content_and_literals`); without
  it they stay red for the same documented P2. Assertion strength is unchanged or stronger (exact single finding).
- **D-b (AC10 placement):** the root templates carry the `.local/{evidence,secrets,tmp}` and existing-task retention
  sentences unchanged; the isolated-worker fallback sentence is added only to `adoption.md` (§4.5), which the rendered
  files link. The test asserts each at its actual location; no unclassified template hunk was added.
- **D-c:** `helper_dirs()` scans both `HELPER_CANDIDATES` in every worktree (a superset that includes the resolved
  path) rather than only `helper_relative()`; refusal coverage is never narrower. `committed`/`working`/worktree
  comparisons use `helper_relative()`.
- **D-d:** `management/SUPERVISOR-HANDOFF.md` L92 is left for Codex (Plan §4.4 lists it with the Codex-owned
  retargets), although the allow-list would permit the one prefix change.
- **F5 (AC5 expectation):** the Plan's "A2 count non-increasing" cannot hold while closed records keep their original
  links (Spec: closed records keep original path text); no closed record was edited.
- Plan §4.2 attributes the handoff SKILL hits as A ×8 / B ×6; observed A ×7 / B ×7 (same 14 lines).

### Gates not performed (coordinator/owner, after merge) and holds

No MR, merge, outside-repository write, export, publication, settings/permission change, observer stop/start or live
discovery cutover was performed. Read-only live inventory: observer PID 3867 (`watch.py … run --interval 30`, started
2026-10-17) runs from the main checkout's retired path `.agents/skills/claude-session-handoff/scripts/watch.py`; it
keeps its loaded code until the §7.1 safe-point stop/pull/start (H2). Pending: Codex review of this round; H1 genuine
discovery listings; H2 cutover and the three Codex-owned link retargets; post-merge local full (incl. `provenance`);
§7.2 outside export (expected package result: exit 1 with exactly the D5 P2 finding, plus F2 unless dispositioned);
final ACCEPT; AC9 cleanup. Native refusals: none. Resources: no container, image, lock or build; evidence and tools
kept under `g055-r4/` for review.

### Attempt 2 (r4 bundle) — correction 01 (Codex round-1 CORRECTION LOOP), 2026-10-20

Authority: round-1 source review (not included) (SHA-256
`b45b21d1d1d4004325d567d4a62fb6f146dcc9952fc2ef21dc49c6e1c79f7cd4`), items 1–4 only. Same writer, worktree and branch;
still **attempt 2 of B5**; r4 correction pass **1 of 2**; no Plan, budget or test-policy change. Staging `4c40b5cc`
(review import + PROGRESS; no in-scope source change) reconciled by a conflict-free rebase (round-1 `abf67e3e` → S1 `ac0b0f65`, S2/S3 `52a1a59a`,
S4 `ece66f57`, S5 `94246e1e`), correction commit **`bfe3e891`**; all checks below ran on `bfe3e891`.

**Corrections.** (1) `check_sdlc.py` `IGNORE_LINES` now carries all five template exclusions (adds
`/.claude/settings.local.json`, `/.local/`); `test_a5_ignore_lines` expects 5, new
`test_a5_each_required_ignore_line_is_enforced` removes each required line in turn and asserts exactly that A5
finding, and the parity test also asserts five entries. F1 is resolved. (2) `test_handoff.py` `SESSION`/`OTHER`/
`HANDOFF` replaced (one occurrence each) by distinct canonical synthetic UUIDs `00000000-0000-4000-8000-0000000000a1/a2/a3`
using a value-blind script; the previous values are not reproduced in this Result, the path map or new evidence (a
names-only scan of the bundle and the new evidence finds 0 files). P6 unchanged (strict). F2 is resolved. (3) the
test-only `adopt()` now appends only missing `gitignore.lines` entries and preserves existing content (adoption.md
step 5); new `test_bootstrap_appends_missing_ignore_lines` proves preservation of existing lines, exactly one
occurrence of each required line, and an unchanged file on repeated adoption. (4) AC5 disposition below.

| Check (candidate `bfe3e891`) | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `python3 -B -m unittest discover -s skills/restruct-sdlc/sdlc-management/scripts -p test_check_sdlc.py -v` | 0 | **41 run, 41 OK** (39 + 2 new) | `02adba9c…6a0d` |
| `python3 -B -m unittest discover -s skills/restruct-sdlc/claude-session-handoff/scripts -p test_handoff.py -v` | 0 | 66 OK | `1ab436f9…8de4` |
| `check_sdlc.py package skills/restruct-sdlc` (source, parent + children) | 0 | P1/P3/P5/P6 pass, 0 findings | `5c40f913…3d12` |
| `check_sdlc.py package skills/restruct-sdlc/sdlc-management --skills-root skills/restruct-sdlc` | **1** | expected D5 only: exactly `P2 … "220 lines exceeds 200"`; every other check passes (not a pass) | `f3569aea…8929` |
| bundle-root (export layout) `check_sdlc.py package <dist-sim>` on a **local disposable simulation** built from `git archive bfe3e891 skills/restruct-sdlc` (+ neutral README, LICENSE, generated MANIFEST; 45 files; not the outside repository) | **1** | expected D5 only: exactly the P2 finding on `restruct-sdlc/sdlc-management/SKILL.md`; P1/P3/P3t/P4/P5/P6/P7 pass | `d0147390…98e3`, MANIFEST `5fe82f8d…d5e` |
| AC4 partition re-derived (base `4c40b5cc`) | 0 | tree 44/44 match; 23 identical / 15 classified / 0 unmapped; rule counts equal | summary `670f77d6…82fa`, partition `6dd4bd1d…5f7a` |

`test_watch.py` not rerun (no change; round-1 84 OK stands). The real outside export, its gitleaks scan and
publication remain the post-merge coordinator/owner gate.

**AC5 disposition (explicit validation deviation).** The Plan's AC5 expectation that the global `adoption` A2 count is
non-increasing **did not pass and is not claimed**: at round 1 the repository-wide scan went 68 → 172 findings.
Separated by kind:
- *Closed/historical records* — **126 added A2 findings** (goals 27, specs 79, archive 10, plans 5, results 5) are
  links in superseded/closed Goals, Specs, Plans, Results and PROGRESS archives that keep the retired
  `skills/sdlc-management/…` paths. They are preserved unchanged as history (Spec: closed records keep original path
  text; Git history and the path map explain the move). Disposition: retained; no edit.
- *Active references* — after the S4 retargets, active governance/runbooks/templates have **no** retired-path link;
  the active-reference inventory leaves only protected, now-inert CI/harness rows (`.gitlab-ci.yml`,
  `infra/ci/profile-paths.tsv`, `infra/ci/validate-controls.sh`, `infra/validation/run.sh`), the OPERATIONS
  historical/inert lines and the deliberate retired-path negative fixture in `test_handoff.py`.
- *Codex-owned cutover links* — `management/SUPERVISOR-HANDOFF.md` L92 and `management/v0/PROGRESS.md` (2 of the added
  findings) plus `management/v0/ROADMAP.md` L230 (no A2 finding): planned Codex retargets at the merge/cutover point
  (H2), not edited by the writer.
- The remaining delta is bookkeeping: −25 (old installed-skill link failures, "not installed") and the moved A6 row,
  which F2's correction now removes.

Scope: `git diff --name-status -M origin/staging HEAD` checked against the 26 patterns after this commit (see
handback); protection diff empty. Pending gates unchanged from round 1.

Reconciliation addendum: before the push, staging advanced to `dbb421b7` (G062-before-G020 scheduling; PROGRESS,
ROADMAP, G020/G062 Goal/Spec only — no in-scope or protected-by-this-task change). Conflict-free rebase:
`bfe3e891` → `1da96c77` (correction source) with byte-identical skill/governance blobs, so the checks above stay
attributable without a rerun. Scope after rebase: 53 diff rows, 91 paths, 0 outside the 26 patterns; protection
diff empty.

### Attempt 2 (r4 bundle) — correction 02 (MR !63 Duo thread `da4f3d09`), 2026-10-20

Authority: coordinator-relayed bounded correction; Codex confirmed it is inside approved G055 scope. Same writer,
worktree and branch; **attempt 2 of B5, r4 correction pass 2 of 2**; no Plan, budget or test-policy change. Resync:
the coordinator had rebased the branch onto staging `d022656162395ace0460a79bc6404d3906db0572`; local branch reset
(`git reset --keep`, clean tree) to the remote head `9de56517609948280ef80be9bdffda3be8e95484` (bundle tree unchanged),
staging verified as an ancestor. Correction commit **`1fba7b2d`**.

**Finding (valid).** Round 1 dropped the retired standalone scripts directory from `helper_dirs()`. During the
migration/cutover window a helper or observer can still run from `.agents/skills/claude-session-handoff/scripts/`
(the live observer does, see round 1), so `process_scan` would miss it and compact's
`other-helper-invocation-in-flight` guard would stop refusing. **Fix:** new scan-only constant
`LEGACY_HELPER_RELATIVE` (commented as covering the migration window); `helper_dirs()` now also adds that directory
in every worktree and the home copy `~/.agents/skills/claude-session-handoff/scripts`, with missing directories still
ignored by the existing `suppress`/`is_dir` guard. `HELPER_CANDIDATES` and `helper_relative()` (writer/committed-path
selection) are unchanged. Limitation, unchanged from the baseline design: a directory is added only while it exists
on disk; after a pull removes the retired directory, a still-running watcher is matched only through the existing
`handoff.py` basename rule — hence the Plan §7.1 safe-point stop before the pull remains required.

**Regression test** `T14LegacyWriter.test_retired_standalone_helper_path_is_still_scanned` (synthetic temp tree and
fake `/proc`, temporary HOME, no real processes, synthetic UUIDs only): asserts both retired directories are in
`helper_dirs()`, the writer path is still `HELPER_CANDIDATES[0]`, and three invocations — `watch.py … run` by absolute
retired path, `watch.py status` relative to a cwd in the retired directory, and a `python3.12` script in the home
retired copy — are each detected (`helper_invocations` = that PID) and compact refuses with
`other-helper-invocation-in-flight` (dry run and apply, log unchanged).

| Check (candidate `1fba7b2d`) | Exit | Outcome | Evidence SHA-256 |
|---|---|---|---|
| `python3 -B -m unittest discover -s skills/restruct-sdlc/claude-session-handoff/scripts -p test_handoff.py -v` | 0 | **67 OK** (66 + 1) | `75599a4d…ffae2` |
| `… -p test_watch.py -v` (rerun: `watch.py` imports `handoff`) | 0 | 84 OK | `140aac80…b10a` |
| `check_sdlc.py package skills/restruct-sdlc` (source bundle root) | 0 | P1/P3/P5/P6 pass, 0 findings | `5c40f913…3d12` |
| `check_sdlc.py package skills/restruct-sdlc/sdlc-management --skills-root skills/restruct-sdlc` | **1** | expected D5 only: exactly `P2 "220 lines exceeds 200"` (not a pass) | `f3569aea…8929` |
| export-layout `check_sdlc.py package <dist-sim-1fba7b2d>` (local disposable simulation from `git archive 1fba7b2d skills/restruct-sdlc`, 45 files; not the outside repository) | **1** | expected D5 only: exactly the same P2 finding on `restruct-sdlc/sdlc-management/SKILL.md`; P1/P3/P3t/P4/P5/P6/P7 pass | `6ba1c831…c730`, MANIFEST `d15b2d32…6339` |

`test_check_sdlc` not rerun (untouched; correction 01's 41/41 stands). Path map: `handoff.py` and `test_handoff.py`
rows updated (new hashes `456c04e0a280`, `256050d1b408`; classification extended); the 23/15/0 partition is
unchanged. No MR, Duo-thread, merge, outside-repository, cutover or observer action.
