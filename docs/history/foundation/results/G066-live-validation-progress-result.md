> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G066 — Simple live test progress — Result

Status: implementation done, pending Codex source review. B3 attempt 1, no budget reset.

## Restart and supersession

- The earlier expanded approach (old head `6fae883a`, two commits) is superseded by
  [Spec r2](../specs/G066-live-validation-progress-spec.md) (sha256 `0eeb28fc…4dc480`).
  - It had an ID allowlist, caps, identity checks, a failure marker, VG1–VG3 and a count-pin change.
  - It is preserved in `<home>/.local/state/platform-evidence/g066-a1/g066-impl-old-6fae883a.bundle`
    (verified, complete history).
- This is a clean restart from `origin/staging` `c7dd47591ff4883c948d7c937d317654fef38308`. Implementation commit:
  `99db8f3b`.

## Method (shared record()/validate.sh path)

- `infra/validation/lib.sh` `record()`:
  - When `PLATFORM_PROGRESS_FILE` is set, it also appends `ID<TAB>STATUS` to that file.
  - It never writes the message there.
  - Append errors are ignored.
- `infra/ci/validate.sh` `run_harness`:
  - It creates `$EV/state/progress.tsv` and exports the variable to the run.sh launch. `$EV/state` is never scanned or
    staged, and `do_scrub` removes it.
  - It loads `required_ids "$EXECUTED_SUITES"` once into the set `PG_REQ` and sets `TOTAL` to its size. These are the
    required checks of the executed suites plus `final`, as completeness counts them. The fast profile gives 35. If
    loading fails, `TOTAL` stays `?` and no progress line is printed.
  - The blocking `wait` becomes a 0.5 s poll (`kill -0`, `DEADLINE`). After the poll come the unchanged wait, deadline,
    straggler and finalize code, then one final drain.
  - `progress_drain` matches complete lines `^([A-Z][A-Z0-9-]*)<TAB>(PASS|FAIL|UNMET)$`. It counts and prints only when
    the ID is in `PG_REQ` and not yet seen, then marks it seen. So each required check counts once, and N never exceeds
    TOTAL. Output: `validate: progress N/TOTAL STATUS`, for example `validate: progress 12/35 PASS`.
  - `INFO`, non-required IDs (such as `PREFLIGHT` or `BACKGROUND`), repeats and anything else are ignored silently.
    The ID is never printed: only numbers and the status word reach the console.
- `docs/foundation/OPERATIONS.md`: one operator sentence in Harness step 3. The control count did not change.
- No changes to controls, run.sh, CI files or results/scanner behaviour.
- Diff stat of the source files versus base `c7dd4759`: `OPERATIONS.md` 2 +-, `validate.sh` 32 +/1 -, `lib.sh` 2 +.
  In total: 3 files, 34 insertions, 2 deletions. Staging has since gained management-only commits (`49bfa205`), which
  touch none of these files.

## Correction: N could exceed TOTAL (owner request)

- In the first version (`475bf505`), N counted every recorded PASS/FAIL/UNMET row, including non-required IDs. So N
  could exceed TOTAL in a failing run. The owner asked for this to be fixed.
- Fix (one correction commit on top of `475bf505`): `record()` now writes the ID with the status, and `progress_drain` counts only the first row of each required
  ID. The output format and the poll, wait, deadline, signal and finalize code are unchanged.

## Reconcile and owner format change (2026-10-23)

- Single reconcile: the three G066 commits were rebased onto `origin/staging` `8d07e5b8` (the zod 4.6.6 repair merge
  `747e46af` plus one management-only commit). That rebase was the only reconcile for the dependency repair that failed
  the earlier hosted fast run. No conflicts. The `infra/` and `docs/` diff versus staging was byte-identical before and
  after.
- Owner UI choice: the live line now shows an icon and the check ID on PASS, FAIL and UNMET lines:
  `validate: progress N/TOTAL ICON STATUS CHECK-ID`, for example `validate: progress 1/35 ✅ PASS B-SHELL` and
  `validate: progress 13/35 ❌ FAIL B-DEPSCAN`.
  - ✅ for PASS. ❌ for FAIL and UNMET, and the literal status word stays.
  - CHECK-ID is the required ID that the regex already matched and found in `PG_REQ`. No message or other text is
    printed. This replaces the earlier rule that the ID is never printed.
  - The emoji are written as byte escapes (`$'\xe2\x9c\x85'`, `$'\xe2\x9d\x8c'`), so the script stays ASCII.
- Otherwise the counter code is unchanged: counting, TOTAL, once-per-required-ID dedup, polling, wait, deadline,
  results, exit and cleanup. Only the `say` line, one `local` and the comment changed in `validate.sh`, plus the
  OPERATIONS.md sentence.
- Format re-check (extracted `say` and `progress_drain`, real `record()` in a child that exits 3, fake `docker`):
  - Output was exactly `1/3 ✅ PASS B-SHELL`, `2/3 ❌ FAIL B-DEPSCAN` (a half-written line across polls, printed once)
    and `3/3 ❌ UNMET C-THREE`.
  - Non-required IDs, a repeated required ID, `INFO` and junk rows printed nothing.
  - The planted canary never reached the console. Exit code 3 was kept.
  - The output bytes under `LANG=C.UTF-8` and `LC_ALL=C` were identical.
  - `bash -n` and `git diff --check` passed.

## Validation run (stack-free: fake `docker` first in PATH, `DOCKER_HOST=tcp://127.0.0.1:1`)

- `bash -n infra/ci/validate.sh infra/validation/lib.sh`: pass.
- `git diff --check`: pass.
- Correction re-check (extracted code from the corrected validate.sh, the real `record()` in a child that exits 3,
  under `LANG=C.UTF-8`):
  - The TOTAL load (the extracted `run_harness` lines with `suite_set`/`required_ids`) gave 35 for the fast suites and
    13 for `lint`, matching the catalog. An invalid suite list left TOTAL at `?`.
  - With three required IDs, the output was exactly `1/3 PASS`, `2/3 FAIL`, `3/3 UNMET`. The first line was printed
    while the child was still running.
  - Non-required `PREFLIGHT FAIL` and `BACKGROUND UNMET` rows printed nothing. A repeated required ID (`A-ONE FAIL`
    after `A-ONE PASS`) printed nothing, and an `INFO` row printed nothing.
  - A required line written half across two polls (`B-TW`, then `O<TAB>FAIL`) printed once, as `2/3 FAIL`.
  - Junk lines were ignored: a bare `PASS`, a trailing message, a lowercase ID, an empty ID and an unknown ID.
  - No ID and no planted canary reached the console. The rows `record()` wrote held only `ID<TAB>STATUS`, and the
    result rows were unchanged.
  - The saved exit code 3 was kept.
- The first version's check (before the correction) passed the same way with status-word-only rows.

## Not run

- Shellcheck (`B-SHELL`): it needs its container and there is no host shellcheck.
- Any CI control, the harness, Docker or Compose, a local fast/full run, or the hosted pipeline.
- The only proof is the hosted `validate:checks` fast job on the MR. Its log should show
  `validate: progress N/35 ICON STATUS CHECK-ID` lines before `harness: exit status`, and `B-SHELL` passing.


## Supervisor-imported final qualification and cleanup — 2026-10-23

MR !89 merged as `79097127` from accepted `19df51f0`. Hosted fast pipeline
`2982880122` / job `17428285915` succeeded: 35 required checks PASS, with all
35 live `N/35 ✅ PASS CHECK-ID` lines emitted before harness exit. Tested/merged
infra, foundation docs, skills, archetype, web lock and Result bytes match.
The earlier Zod-blocked job remains FAILED. Broader integration/full/local
coverage was not run under the owner's fast-only exception.

Finished G066 refs/worktrees removed; coordinator reports local Docker baseline
15/11/6/52. Supervisor verifies named refs/tree absence and the retained old-WIP
bundle. Existing held native refs/G064 resources remain protected; hosted shared
renderer-cache inventory is disclosed separately from local cleanup. AC1–AC3,
artifact hash and all retention/limitations are in
[the FINAL ACCEPT](G066-source-codex-review-20261023.md#final-g066-accept--2026-10-23).
Status: **G066 finally ACCEPTED**, B3 attempt 1. Immediate RC9 cut follows.
