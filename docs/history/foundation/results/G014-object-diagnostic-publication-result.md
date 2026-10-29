> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G014 object diagnostic publication: Result (follow-up round, part (a) only)

Branch `g014-obj-diagnostic-codes` from staging 8cbbbc66613cc327455a0827d926a653240a558a (fetched and verified as an
ancestor before every batch). Authorized by
G014-object-diagnostic-publication-review.md (not included) (SHA-256 350acf7b…0e99,
"Execution boundaries"), implementing the accepted addendum (SHA-256 d06e8bd9…f8b3) with its two clarifications.
Writer: the released G014 diagnose/correction worker (agent-ac07fb18, isolated worktree). No MR opened; integration
waits for Codex's source review and for the disposition of gate retry 17225023104, which this round never touched.

## Budget, ownership and effort

- Attempt ledger: G014's existing pool, no new or reset pool; this round is **G014 attempt 3 of 7** as counted by the
  coordinator. Cycle count in this round: 1 (no correction loop yet).
- Effort/model: run as **Fable 5.1 high** at Codex's direction (reuse of the diagnosing worker), against the
  implementation default of Opus high; the actual effort was not verifiable from inside the session beyond the stated
  setting.

## Change (exactly the three files plus this Result)

1. `infra/validation/lib.sh` (+55): `obj_fail_codes LOG N` after `record`. Reads only the top-level and first-level
   `--- FAIL:`/`--- PASS:` header lines of the TestObjects/TestObjectsFootprint log (anchored, `LC_ALL=C grep -a -E`),
   maps each failed subtest's exact Go-printed name (the ten `t.Run` names of
   core/internal/tenancy/access/objects_integration_test.go:977-1490) through a literal table to the closed codes
   OBJ-ATTEST, OBJ-READ-WRITE, OBJ-RLS, OBJ-RT, OBJ-SCOPE, OBJ-SCHEMA, OBJ-DEL, OBJ-SIGN, OBJ-ADMIN, OBJ-LOG, plus PARENT
   (TestObjects failed with no matched subtest), FOOTPRINT (explicit TestObjectsFootprint failure), one trailing
   `unknown` (a failed subtest outside the table), `unknown` (nothing attributable) or `none`. Clarifications built in:
   `none` requires N == 10 **and** an explicit top-level PASS of both TestObjects and TestObjectsFootprint (missing,
   skipped, truncated or panic-ended evidence gives `unknown`; explicit footprint failure gives FOOTPRINT); the output
   grammar is known codes in canonical order with at most one trailing `unknown`. Missing, symlinked, unreadable, empty,
   NUL-bearing or over-64 MiB logs, CRLF headers and any helper error give `unknown`. A subtest name is looked up only
   when it consists of the table's own characters (never `@`, `*` or a bracket as a subscript). Always exits 0, never
   echoes input, prints exactly one token list.
2. `infra/validation/run.sh` (1 line, `obj_it` FAIL row): `…; logcheck: ${lc:-ok}; failed-subtests: $(obj_fail_codes
   "$log" "$n"))` appended to the existing B-OBJ-IT FAIL detail; every current field, the PASS text and the status
   decision (`ok`, taken before the helper runs) are unchanged. results.tsv already travels through scan, stage and
   publish; no new file, no STAGE_REPORT change, no timing-test change.
3. `infra/ci/validate-controls.sh` (+116/-6): the stub harness's FAIL-row detail literal becomes
   `stub failure; failed-subtests: OBJ-SIGN` (the `STUB_FAIL_ID` printf and the `fail`/`fail-row` sed); a shared
   `expect_fail_row_code` asserts that the published `ci-evidence/results.tsv` holds exactly one FAIL row and that it
   carries that literal verbatim; existing `A V3` and `A V7` call it; the `obj-codes` fixture (FIXTURES, selected only
   by `--only`, not listed, so the control count pinned by run.sh is unchanged; no new control ID) sources the
   template's tracked lib.sh in a strict-mode subshell and covers: single, multiple in reverse log order (canonical
   output), parent-only, footprint, unknown and mixed (`OBJ-SIGN,unknown`), private/spoofed text (a canary, a signed URL
   with X-Amz values, an assertion line before the header, the header text inside an RLS-ACCESSLOG line, an unindented
   header, a 1 MiB line) excluded from the output, a prefix spoof (`OBJ-SIGNATURE_leak` → `unknown`), NUL, CRLF,
   bare 1 MiB line, empty, missing, over-cap (sparse 64 MiB + 1) and (unprivileged) unreadable input → `unknown`,
   `none` with its truncated-footprint and skipped negatives, and the `record B-OBJ-IT FAIL …` expression evaluated
   under `set -euo pipefail` with a missing log, which stays the FAIL row with `failed-subtests: unknown`.

## Validation (focused only; no full controls, migrations or ops suite; no G014 test rerun)

Pre-batch reconciliation: `git fetch origin staging` → 8cbbbc6, `merge-base --is-ancestor origin/staging HEAD` ok,
tree clean apart from the three files. `bash infra/ci/docker-desktop-mounts.sh`: 12980 of 100000 mounts (12%).
Both batches ran under `flock <home>/.local/state/platform-evidence/local-harness.lock`; no edits were made
while either ran. Evidence under `<home>/.local/state/platform-evidence/g014-impl-b04ec59/`:

| batch | command | outcome | evidence (SHA-256) |
|---|---|---|---|
| controls | `oc-controls.sh oc-controls1` = `bash infra/ci/validate-controls.sh --only V3,V7,V5a,V5c,V6c,obj-codes --log …` | **6 passed, 0 failed** (of 6 selected; 441 defined), 51 s, exit 0 | oc-controls1/ci-controls-summary.log 520b56cd…e2be; meta.txt 0a9d10ac…a1b5; oc-controls.sh bda2eb8f…b815 |
| lint | `oc-run.sh oc-lint1 lint lint-base` = `run.sh --base 8cbbbc6 --allow-file obj-codes-allow.txt --owned-registry … --secrets-file … lint lint-base` | **all PASS**, 0 FAIL/UNMET: B-SHELL (shellcheck warning + SC2251 on infra/ci), B-SECRET-SOURCE, B-SECRET-CANARY (16 generated secrets absent, planted control detected), B-SCOPE (3 changed paths since 8cbbbc6, all allow-listed), B-REQUIRED-CHECKS (16/16); 255 s, exit 0; cleanup rc 0, leftover inventory empty | oc-lint1/report/results.tsv fdb85f33…ceeb; meta.txt 7e1d9966…3b00; oc-run.sh 227888cf…e702; obj-codes-allow.txt a9f6cffd…d674d |
| pre-check | `obj-codes-smoke.sh` (standalone, hostile names `@`, `*`, `]`, `[`; mixed; two codes + FOOTPRINT; none/unknown; missing) | smoke-ok, 9/9 | obj-codes-smoke.sh 29e0e341…081f1 |

Control results verbatim: `PASS V3 exit 14 (FAIL row under harness exit 0; failed-subtests code published)`,
`PASS V7 exit 1; postcheck passed; failed-subtests code published`, `PASS V5a exit 15; minimal WITHHELD; canary
absent`, `PASS V5c exit 15; gitleaks finding withheld`, `PASS V6c exit 0; both scans clean`, `PASS obj-codes …
(row check 1); 0 problems`. So the actual published closed-code content (V3, V7), withholding (V5a, V5c), clean
publication (V6c), invalid/private/spoofed/multiple/unknown inputs and the unchanged FAIL output on helper errors
(obj-codes) are each evidenced.

## Deviations, limits and what remains uncertain

- One mid-implementation syntax fix (a `;` inside an unquoted `=~` pattern, moved into a variable) before any batch;
  no batch was rerun.
- The lint suites and the selected controls are the only local coverage, per the release. The hosted pipeline selects
  ci-controls itself for the `infra/*` change; the staging gate runs the migrations suite after a reviewed, authorized
  integration. The real producer path (a genuine B-OBJ-IT failure) is exercised end to end only by the next hosted
  failure; locally it is covered by the fixture's `record` expression and the publisher controls.
- Part (b) (assertion-line publication) stays deferred/unapproved. No change to any expiry assertion; the original
  hosted failure's cause remains uncertain.
- Handback expects Codex's source review; no MR, no staging push, no change to retry 17225023104.
