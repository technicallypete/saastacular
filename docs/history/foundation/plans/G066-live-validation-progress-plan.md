> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G066 — Live validation check progress — Implementation Plan

Revision: r1, 2026-10-23, plus the correction addendum r1 appended at the end (r1 §1–§9 preserved unchanged;
the addendum states what it replaces). Planning only; nothing implemented, built, run or cleaned (stack-free: static
reading, `bash -n`, `git`).
Goal: [G066](../goals/G066-live-validation-progress.md) at staging `4afca83752ddb660652166b73cc0d1c956684156`
(sha256 `e17804dd30e13509ece46ee756beabd9c50a3af20e9e23e1d40fb6d58c8bd65c`; the relayed hash
`9e28c145…` is the same Goal at `deed0a02` before the appended "Owner RC milestone" section, which changes no
scope). Spec: [r1](../specs/G066-live-validation-progress-spec.md) (sha256
`04e6598a0b328531eb23b3233c990eb2189b43406c992d1dfb9dbab91040d5ea`, identical at both SHAs).
Source base: `4afca837` (G065 merged, MR !87; all line numbers below are at this SHA).
**Implementation does not start until G065 final ACCEPT, this Plan's ACCEPT and the readiness/release
reconciliation are recorded.** Budget: B3, zero attempts spent; checkpoint after attempt 2.

## 1. Settled decisions (design summary)

1. **Writer = the existing record path.** `infra/validation/lib.sh::record()` (26–31) additionally appends one
   line `ID<TAB>STATUS<TAB>EPOCH_SECONDS` to the file named by `PLATFORM_PROGRESS_FILE` when that variable is
   non-empty (one `printf >>`, `|| :`; never the message, never RESULTS/TIMINGS/log changes). Background
   children inherit the variable in their subshell (`bg_start`, 2403–2411) and their `exec >"$dir/log"` (2407)
   does not touch a path append, so a child's row reaches the file at record time, before `bg_join`.
   `bg_join` merges child rows with `printf` (2439), not `record`, so merged rows produce no replay; its own
   `BACKGROUND` row (2448–2456) is a genuine event.
2. **Transport = one regular append-only file, no helper process, no FIFO, no Docker resource.**
   `infra/ci/validate.sh::run_harness()` creates `$EV/state/progress.tsv` (mode 600; `$EV/state` exists from
   `locate`, 305) and exports `PLATFORM_PROGRESS_FILE` for the `run.sh` launch (799), exactly as
   `PLATFORM_LEDGER_BOOTSTRAP_FILE` (794–795); a caller-supplied value is unset at startup beside line 58.
   `run.sh` accepts it only as an absolute path outside `$REPO` naming an existing regular non-symlink file
   (preflight, beside 122–127); otherwise it unsets it and logs one line (the run continues). No new `run.sh`
   argument: the stub harness parser (validate-controls.sh 1072–1081) and argument assertions stay untouched.
   `$EV/state` is never scanned (`do_raw_scan` 1371 covers `$EV/raw`), never staged (`STAGE_REPORT` 122–131) and
   is removed by `do_scrub` (2013) on every path, so no new cleanup step.
3. **Reader = the parent wrapper during its existing wait.** The blocking `wait "$HPID"` (812–817) becomes a
   bounded poll: every 0.5 s drain newline-terminated lines (one fd opened once; a trailing partial line is
   carried to the next read), while `kill -0 "$HPID"` and `DEADLINE == 0` (bash reaps the ended child
   asynchronously and keeps its status, the pattern of validate-controls.sh `await()` 670–684); then the
   existing `wait "$HPID"` for `rc`, the existing deadline/straggler/`on_signal` (837–849)/`finalize` (2079)
   code unchanged, one final drain after the harness ended. Reading is capped at 1 MiB (then one fixed notice).
4. **Trusted fields only.** A line is emitted iff it matches
   `^([A-Z][A-Z0-9-]{0,63})\t(PASS|FAIL|UNMET|INFO)\t([0-9]{10})$` and the ID is in the set loaded once from the
   bound `infra/validation/required-checks.tsv` (the `do_complete` row rule, 1515; 130 IDs) plus the fixed
   auxiliary inventory `PREFLIGHT PROBE PREBUILD BACKGROUND BUILD-LEDGER RESOURCES IMAGES B-ROOT
   B-SCOPE-EXCEPTION DEPSCAN-PINS` (every non-catalog `record`/`check` ID in run.sh/lib.sh at `4afca837`; the
   variable-ID sites run.sh 431–460, 1750, 3064–3070, 5105 and lib.sh 180 enumerate catalog IDs). Output:
   `validate: progress <ID> <STATUS> t=<epoch>` — distinct from the authoritative `validate: check <ID> <STATUS>`
   lines of `do_report` (1996), which stay as they are. Anything else (unknown ID, extra field, control
   characters, lowercase, unknown status, over-long line) is counted and never echoed; at the end one line
   `validate: progress: N event(s) suppressed`. A missing/unreadable file or catalog gives the fixed notice
   `validate: progress: unavailable` once; `rc`, phases, artifacts and exit codes are untouched.
5. **Everything else unchanged.** `results.tsv`, timings, `junit.sh`, scanners, staging, `--profile none`,
   recovery markers, direct `run.sh` invocations (their console already streams `record` lines through the tee at
   run.sh 117), `.gitlab-ci.yml`, `hosted-context.sh` (its attached `docker run`, 267/414, already relays
   `validate.sh` stdout to the job log live), `CI_SUITES_*`, `required-checks.tsv`.

## 2. Unproven assumptions and their proofs

| # | Assumption | Proof (smallest) |
|---|---|---|
| U1 **riskiest** | The 0.5 s poll in place of `wait` keeps deadline (USR1 after a `sleep`, ≤0.5 s late), signal, straggler and finalizer semantics within the existing controls' bounds. | **Early proof, first run after coding §1.2–1.4 and VG1:** `--only VG1,V18d,V18` (new live-before-deadline control plus the reused deadline and TERM controls). A failure here stops the task (Goal stop: lifecycle redesign). |
| U2 | A child's `record` row is visible in the file before its `bg_join`; merged rows do not replay. | VG3 (real `lib.sh`, VB driver). |
| U3 | No payload can reach the console before the scans. | VG2 (planted canaries in every rejected shape; `expect_no_canary`). |
| U4 | The hosted job log shows the lines while the harness runs. | The MR's own `validate:checks` fast job log (line timestamps before `harness: exit status`); no local full run. |
| U5 | No existing control parses the whole job log strictly enough to break on ≤ ~140 extra lines. | Verified statically: controls use `expect_line`/`grep` on named lines (1489, 1525); confirmed by the S3 reused controls. |

## 3. Files and exact changes

| File | Change |
|---|---|
| `infra/validation/lib.sh` 26–31 | `record()`: the guarded progress append (≈6 lines). |
| `infra/validation/run.sh` ~122 | preflight validation of `PLATFORM_PROGRESS_FILE` (≈6 lines; `log` only, never a row). |
| `infra/ci/validate.sh` 58, 776–835 | unset the caller value; `progress_open` (file + catalog), `progress_drain` (validate/emit/count), the poll loop, final drain and notices inside `run_harness` (≈70 lines; no new phase, `PHASES` 87 unchanged; `!`-free, shellcheck `-S warning` and `SC2251` clean). |
| `infra/ci/validate-controls.sh` | roster line 110: `"D VG1" "D VG2" "D VG3"`; header comment (group D); stub harness modes `progress-hold`, `progress-plant`, `progress-missing` beside 1327–1367 (write to `$PLATFORM_PROGRESS_FILE`, which validate.sh sets before the `env -i` launch reaches the stub); VB driver mode `progress` beside 3570; `ctl_VG1/2/3` beside 3756. |
| `docs/foundation/OPERATIONS.md` 302, 432, 753 | the operator note (fields, informational status, local/hosted applicability, where messages live: `harness.log`/`diagnostics/`); count 529 → 532. |
| `management/v0/results/G066-live-validation-progress-result.md` | new Result. |

None: schemas/migrations, public API/MCP, application code, dependencies, `.gitlab-ci.yml`, `profile-paths.tsv`,
`required-checks.tsv`, archetype (manifest `2ea491555634dde6` and `skills/saas-foundation/templates` list none of
the four files: no export-copy reconciliation).

**Control-count pins (repo grep at `4afca837`):** only `docs/foundation/OPERATIONS.md:753` ("prints 529
controls"); `run.sh` `B-CI-CONTROLS` (1518–1534) compares `--list` with the passing set and computes `$n`;
`required-checks.tsv` keeps one `B-CI-CONTROLS ci-controls` row and its 130-row comment; management prose is
historical. Dependent callers/pins: `record()` 3-arg signature unchanged for every `record`/`check` site;
`HPID` users (`stop_harness_group` 719, `on_signal` 842, `finalize` 2086, `harness_group_lingers` 746)
unchanged; `V1` asserts `^validate: check B-SHELL PASS$` (1525) — unaffected by the `progress` prefix.

## 4. New controls (group D; each justified by behaviour no control covers)

- **VG1 — live before a blocked step (U1, AC1).** Stub mode `progress-hold`: writes `PREFLIGHT INFO`, then
  `B-SHELL PASS` from a forked subshell (a child before any join), then `B-DEPS-CONTROLS FAIL`, marks
  `harness-running`, holds. The control waits ≤30 s (the R1 observation bound) for the three
  `validate: progress` lines in `$C/out` while the stub still holds (no `harness: exit status` line yet), then
  ends the watchdog sleep as V18d does. Expect: the three lines precede `harness: internal deadline reached`,
  `expect_rc 124`, `expect_stop deadline`, `expect_group_gone`, `expect_trap yes`, no `progress.tsv` left under
  `$C/tmp`. No daemon.
- **VG2 — confidentiality and outcomes (U3, AC2, AC3).** (a) stub mode `progress-plant,fail`: the file gets
  `B-SHELL PASS`, `B-SHELL FAIL` (genuine repeated transition), `X-UNKNOWN-<canary> PASS` (valid syntax, unknown
  ID), `PREFLIGHT INFO <ts> <canary>` (extra field), an ID with `\e[31m`, a raw `leak <canary>` line, a
  2000-byte line, a lowercase ID, `B-SHELL SKIPPED`. Expect exactly the two `B-SHELL` lines, one
  `progress: 7 event(s) suppressed`, `expect_no_canary` on `$C/out` and the published set, `expect_rc 1`,
  `expect_line has harness=1`, `expect_pub full`, the published `FAIL` row (as V2). (b) `progress-missing,fail`:
  the stub removes the file first → one `progress: unavailable`, still exit 1 (transport failure cannot mask a
  failed run). No daemon.
- **VG3 — background record before join (U2, AC1, AC3).** VB driver mode `progress` with
  `PLATFORM_PROGRESS_FILE=$WORK/progress.tsv`: `bg_start controls ctl_hold`; after `ready`, the file holds
  exactly the two `ctl_rows` rows with 10-digit epochs while the child is held; release; `bg_join`; the file
  gains exactly one `BACKGROUND INFO` row (no replay of the merged rows); a foreground `record B-SHELL FAIL`
  adds the fourth; driver exit 1; `results.tsv` unchanged in shape. Case (c): the variable unset → no file
  created. Daemon only through `vb_gone` (as VB5).

Reused, not duplicated: V18 (TERM), V18a (ignore-TERM/KILL), V18d (deadline), V18j (stray group), V18k
(blocked result), V1/V2 (stub pass/fail through the full finalizer), VB1/VB3b (real `run.sh` under the real
wrapper, kill/deadline), VB5 (join/foreign/lost), B-SHELL (shellcheck over the three edited scripts,
run.sh 314–322), hosted `validate:checks`.

## 5. Validation: selectors, order, estimates

Private scratch of the implementer's isolated worktree; one test slot of the shared pool; no image build (no
build slot); fetch/rebase reconciliation before each batch. Order and coding/elapsed/wait estimates:

| Step | Command (from the task worktree root) | Elapsed | Covers |
|---|---|---|---|
| S0 code §1.2–1.4 + VG1 | — | coding 2–3 h | — |
| S1 static | `bash -n infra/ci/validate.sh infra/validation/lib.sh infra/validation/run.sh infra/ci/validate-controls.sh`; `bash infra/ci/validate-controls.sh --list \| wc -l` = 532 | <1 min | syntax, count pin |
| S2 **early proof** | `bash infra/ci/validate-controls.sh --only VG1,V18d,V18` | ≈5 min | U1, AC1, AC4 |
| S0b code VG2, VG3, driver/stub modes, OPERATIONS note | — | coding 1–1.5 h | — |
| S3 focused | `bash infra/ci/validate-controls.sh --only VG2,VG3,V18a,V18j,V18k,V1,V2,VB5,VB3b,VB1` | ≈30–45 min (V18a 150 s grace; VB1 bound 900 s; VB3b bound 300 s) | AC1–AC4, U2, U3, U5 |
| S4 lint | `bash infra/validation/run.sh --base <staging SHA> --allow-file <task allow-file listing the five changed paths> lint lint-base` (canary rule: `lint-base` with `lint`) | ≈8 min | AC5 B-SHELL, B-SECRET-CANARY |
| S5 hosted | MR `validate:checks --profile fast` (automatic; `validate:integration` stays the optional manual job: `infra/` is path class `foundation`, profile-paths.tsv:61) | ≈25 min + queue wait unknown | AC1 real fast run (U4), AC5 |
| S6 post-merge | merged tree equal to the tested tree → attributable reuse, no run; an `infra/ci`/`infra/validation` delta → S2+S3 selectors on the merged SHA (≈45 min) | — | AC5 affected |

Not proposed: the 532-control `ci-controls` suite, a local `--profile fast|full` wrapper run, a full suite,
new fixtures beyond the three stub/driver modes. Review/queue waits (Plan ACCEPT, Codex source review, Duo,
merge authorization): unknown. Total coding ≈3–4.5 h; total local elapsed ≈45–60 min; hosted ≈25 min.

## 6. Sequence, delegation, budget

P0 Plan ACCEPT; G065 final ACCEPT and release reconciliation recorded → P1 isolated worktree from the current
staging tip (`4afca837` or later; fetch/rebase first) → S0, S1, S2 (early proof) → S0b, S3, S4 → Result
(AC mapping, exact outcomes, inventory, cleanup receipt), safekeeping push → Codex source review → MR to
`staging` (S5 monitored by the coordinator: MR/pipeline ID, SHA) → authorized merge → S6 → cleanup receipt → final
ACCEPT (the owner's RC9 cut follows separately). One Claude implementation owner (Opus 5.5 high), no subagent;
optional Codex read-only confidentiality review of the `progress_drain` diff. Attempt 1 = P1–S5 with in-scope
focused corrections; a reused-control failure needing more than one bounded fix, or any change to deadline,
signal, cleanup or scan semantics, ends the attempt with a compact Result checkpoint and escalation. Checkpoint
after attempt 2.

## 7. Acceptance mapping

| AC | Evidence |
|---|---|
| AC1 | VG1 (three events before the blocked step, 30 s bound), VG3 (child row before join), VB3b/VB1 (real run.sh under the wrapper), the S5 job log bound to the MR SHA |
| AC2 | VG2(a): every rejected shape suppressed without echo, `expect_no_canary`; the catalog+inventory rule in §1.4; `$EV/state` neither scanned nor staged |
| AC3 | VG2 (PASS then FAIL both shown; harness=1 authoritative; `check` lines and published `results.tsv` unchanged), VG2(b) transport failure, VG3 (no replay), V1/V2 unchanged |
| AC4 | V18, V18a, V18d, V18j, V18k, VB3b reused under the poll loop; VG1's no-leftover check; `do_scrub` removal of `$EV/state` |
| AC5 | S1, S4 (B-SHELL, canary), S5 hosted fast, S6 attributable affected verification; scanners untouched |
| AC6 | OPERATIONS.md note, Result with source/merge attribution, inventory and cleanup receipt, Codex ACCEPT before RC9 |

## 8. Cleanup and inventory

Expected persistent resources: none. Transient: the controls runner's scratch root (auto-removed; no
`--keep-scratch`), its control-labelled busybox containers (VB1/VB3b/VB5/VG3; removed and verified by the
runner's per-control cleanup), `run.sh`'s `$WORK` (its EXIT cleanup; shellcheck containers `--rm`), the task
allow-file (private evidence dir), the implementer's worktree (removed at acceptance). No image built, no
volume, no network, no port. The Result records every path/ID at allocation and the absence receipt at each
batch end and at acceptance.

## 9. Risks, questions, escalation

1. U1 fails or needs more than a bounded fix → stop; the Goal forbids a lifecycle redesign.
2. A control found to read the whole job log → fix only that control's assertion if it concerns the new lines;
   otherwise escalate.
3. `set -e`/`inherit_errexit` in `record()`: the append is `|| :`; a failed append never changes a row or status.
4. Scope stop: percentages/ETA, raw-log streaming, a dashboard, `run.sh` argument changes, JUnit changes, any
   RC action.

Questions for Codex: (Q1) confirm binding to `4afca837` and the Goal hash `e17804dd…`; (Q2) confirm the line
format `validate: progress <ID> <STATUS> t=<epoch>` (drop `t=` if an epoch is unwanted); (Q3) confirm that the
hosted fast job plus VB3b/VB1 satisfy "real fast-profile wrapper run" without a local `--profile fast` run;
(Q4) VG3 as a separate control (count 532) rather than folded into VB5 (531).

## Correction addendum r1 (Codex CORRECTION LOOP, 2026-10-23)

Answers the Codex review (not included) (sha256 `dfe65545…`). r1 §1–§9 stand
except where a paragraph says "replaces". Q1–Q4 adopted as settled: binding `4afca837`, Goal `e17804dd…`, Spec
`04e6598a…`; `t=<epoch>`; hosted `validate:checks` is the real fast-wrapper proof (VB3b real wrapper+harness, VB1
harness then replay), no local fast run; standalone VG3, 532. Planning only; B3 at zero attempts. (§1.1's
`bg_start`/`bg_join` lines are `infra/validation/lib.sh` 2395–2456.)

### C1 — bounded reads and child identity (replaces the reader bounds of §1.3)

- **Bounds before accumulation.** The reader takes at most 256 bytes per `read -r -n 256` chunk (a valid record is
  ≤ 81 bytes: 64+1+5+1+10). A full 256-byte chunk without a newline is an *oversized* record: dropped, the reader
  discards to the next newline (those chunks still count against the cap), one suppressed event. A short chunk at
  end-of-file without a newline is a *partial* record: carried (≤ 255 bytes) and completed by the next poll; one
  still present at the final drain is one suppressed event, never emitted. Nothing is buffered beyond 256 bytes.
- **Work per poll.** Each 0.5 s poll handles at most 512 complete records or 64 KiB; a poll that hits its quota is
  followed by the next poll without the sleep (identity and `DEADLINE` checked between polls). The final drain is
  one such series bounded to 2 s wall and the remaining cap, on the normal-exit path and on the deadline path
  (after `stop_harness_group`) only; `on_signal` and `finalize` never drain.
- **1 MiB cap** (every byte read counts, discarded included): fd closed, one fixed line `validate: progress: cap
  reached; later events are not shown`, no further reads (liveness polling only); the suppressed-count line still prints.
- **Identity and completion.** The poll condition is `[[ "$(proc_start_time "$HPID")" == "$hs" ]]` with the `hs`
  captured at launch (validate.sh 806; `proc_start_time` fails for a zombie or gone PID and differs for a reused
  one), not `kill -0`. Completion stays the existing `wait "$HPID"` (812–817, bash's saved status, unchanged); a
  reused PID only ends polling early. If `hs == none` the poll loop is skipped: the existing blocking `wait` runs
  and `validate: progress: unavailable` prints once.
- **Tests, inside the existing controls.** VG1: the stub writes `B-DEPS-CONTROLS\tFAIL\t<ts>` as two appends 1.2 s
  apart (`B-DEPS-CONT`, then `ROLS\tFAIL\t<ts>\n`: the fragment spans ≥ 2 polls); expect exactly one
  `validate: progress B-DEPS-CONTROLS FAIL t=` line and no fragment line. VG2(a): the 2000-byte line carries the
  canary after byte 1500 (beyond the first chunk: discard mode never echoes) and the file ends with an unterminated
  `B-SHELL\tPASS\t<ts>`; expect `8 event(s) suppressed`. New case VG2(c) `progress-cap,fail`: ≥ 1.5 MiB of valid
  `PREFLIGHT\tINFO\t<ts>` rows (one awk, scratch) then `B-SHELL\tFAIL`; expect the cap line once, no `progress
  B-SHELL FAIL` line, 0 < emitted `PREFLIGHT INFO` lines ≤ 52429, `expect_rc 1`, `expect_line has harness=1`,
  `expect_pub full`, the published FAIL row, `harness: exit status` within the existing launch bound (drain ≤ 2 s).

### C2 — producer append failures are surfaced (replaces §9.3 and the `|| :` of §1.1)

- `record()`, after its results/timings rows (so the row already exists): `if printf … >>"$PLATFORM_PROGRESS_FILE"
  2>/dev/null; then :; else : >>"$PLATFORM_PROGRESS_FILE.failed" 2>/dev/null || :; log "progress: append failed";
  PLATFORM_PROGRESS_FILE=""; fi`. Fixed literal, no payload; the variable is cleared in the failing process only
  (a child disables itself; no retry, no back-pressure); row, status and `run.sh` exit are unchanged.
- Wrapper, at the final drain on both paths: if `$EV/state/progress.tsv.failed` exists or
  `grep -qF 'progress: append failed' "$EV/raw/harness-console.log"` (literal test of a private file it already stages,
  1720; nothing echoed), one fixed line `validate: progress: incomplete (a producer append failed)`. `rc`, phases,
  scans, artifacts untouched; `do_scrub` removes the marker with `$EV/state`.
- Test: new case VG2(d) `progress-append-fail,fail`: the stub replaces `$PLATFORM_PROGRESS_FILE` with a directory
  (`EISDIR` for any user, root included), sources the real `lib.sh` as its `standalone-kill` mode does (1118; copied
  into the scratch tree, 1017), sets `RESULTS=$S/producer.tsv TIMINGS=$S/producer-timings.tsv`, calls
  `record B-SHELL FAIL producer`; expect that row in `$S/producer.tsv`, the `.failed` marker, the literal in the
  staged `harness.log`, the `incomplete` line once, `expect_rc 1`, `expect_line has harness=1`, `expect_pub full`,
  `expect_no_canary`, published results unchanged in shape. VG2(b) (reader-side missing file) stays distinct.

### C3 — affected verification on the merged SHA (replaces §5 row S6)

| Step | Command (frozen merged-SHA checkout; no rebase) | Elapsed | Covers |
|---|---|---|---|
| S6a | `bash -n` the four scripts; `bash infra/ci/validate-controls.sh --list \| wc -l` = 532 | <1 min | syntax, count pin |
| S6b | `bash infra/ci/validate-controls.sh --only VG1,VG2,VG3,V18d,V18,V1` | ≈10–12 min | progress output, split/cap/producer-failure, pre-join, deadline, TERM, finalizer on the merged bytes |
| S6c | attributable reuse of V18a, V18j, V18k, VB5, VB3b, VB1, S4 lint and S5 hosted, valid while the merged `infra/ci/` and `infra/validation/` trees equal the tested source trees (hashes recorded) | — | unchanged inputs |

If the merge delta touches `infra/ci/` or `infra/validation/` beyond this MR's own diff, S6b becomes S2+S3
(≈45 min); management-only target commits change nothing. No local `--profile fast|full`, full suite or 532-control
run. G065's no-rerun disposition was that task's own. The Result records merged SHA, selectors, outcomes, reuse (AC5).

### Counts and estimates

VG1–VG3 unchanged in number (VG2 gains (c), (d)); OPERATIONS.md 529 → 532 unchanged. S3 ≈ 32–48 min (+≈ 2 min);
S6 ≈ 11–13 min; local elapsed ≈ 60–75 min; coding +≈ 45 min (≈ 3.75–5.25 h); hosted ≈ 25 min; review/queue waits
unknown. Residual for Codex: confirm VG2(d) (real `record()` in the stub tree) satisfies "actual producer-append-failure".
