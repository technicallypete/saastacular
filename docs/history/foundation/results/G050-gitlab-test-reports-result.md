> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G050 — GitLab test-report visibility: Result

Goal: [G050-gitlab-test-reports.md](../goals/G050-gitlab-test-reports.md) · Spec: [G050-gitlab-test-reports-spec.md](../specs/G050-gitlab-test-reports-spec.md) · Plan r2: [G050-gitlab-test-reports-plan.md](../plans/G050-gitlab-test-reports-plan.md) (`0d1d6d92…`, review `537ad6c8…`)
Author: Claude Code · Date: 2026-10-09 UTC

## Status

**Submitted for Codex implementation review. Not accepted.** Only Codex accepts; PROGRESS is not touched here.
- Implementation and focused local evidence are complete.
- Final automated validation is the MR's hosted fast/full pipeline (owner policy, below): MR !9 pipeline **2930980752** on `b0ac030` **succeeded** (fast 33/33, gate full 77/77; test-report API 110/110). See [Hosted MR pipeline](#hosted-mr-pipeline-s4).
- Budget **B3: 1 of 3 meaningful cycles used.** The cycle-2 checkpoint was not reached.

## Binding

- **Execution release:** `G050-G005-execution-release-codex-confirmation.md` (`02efd1e8…1df5a`; root supplemental authorization, not copied).
- **BASE:** staging `0d524212f185d9157719bc92671c8d098bc249eb`, tree `706ad4de7eebdd67d9bdc8c811ba921cc2370396`.
- **Approved bundle:** `<home>/.local/state/platform-evidence/g050-execution-binding-proposed/`, bytes unchanged:
  - proposal `9b2640e4…b737c8`;
  - snapshot `aa3ae9e7…2168cf` (27 inputs);
  - scope `60ba724d…4503e` (8 expressions), copied byte-identically to G050-scope-allowlist.txt (not included).
- **Checkout:** `.claude/worktrees/g050-exec`, branch `g050-gitlab-test-reports`.
- **Evidence:** `<home>/.local/state/platform-evidence/g050-exec-0d52421/`:
  - `runs.tsv`: one row per run (step, start, end, exit, HEAD, post-state, arguments);
  - per-step `controls.log` and `harness.out`;
  - `bin/` launchers.

## S0

- **Branch:** `g050-gitlab-test-reports` created at BASE.
- **Docker, ports and disk:** unrelated containers (`payments-assistant-postgres-1` on 5433, `outreacher-pg` on 5432, `needknow-drafter-*` on 7777 and 7778) were observed and never touched. No harness port was in use. 915 GiB free.
- **GitLab quota (read-only):** October 2026 usage was 3478 compute minutes. Purchased minutes are not shown in this figure.
- **Shared Docker harness with the parallel G005 work:** every local harness and controls run took the lock `<home>/.local/state/platform-evidence/local-harness.lock`, because the harness uses fixed host ports and a fixed root project.

## Implementation

| File | Change |
|---|---|
| `infra/ci/junit.sh` (new) | `junit_report` builds `ci-evidence/junit.xml` from three validated sources only: the sanitized status columns (`ID<TAB>STATUS`), the recorded phase statuses, and the `JOB` and `PROFILE` tokens. `junit_attr` escapes `& < > " '`, using quoted replacements so Bash 5.2's `patsub_replacement` cannot expand them. Every input is re-validated; any defect returns 19 and writes nothing. Output is deterministic, written via `OUT.tmp` then renamed. |
| `infra/ci/validate.sh` | Sources `junit.sh`. New phases: `junit` after `summary`, and `junit_withheld` after `summary_withheld`, both in `PHASES`. `do_summary` takes the report phase as a third argument and adds it to the `pending` list of `summary.json` and `phase-status.tsv`. `do_junit` decides what goes in: check cases only for a full or STAGING-FAILED candidate whose completeness step ran; phase cases otherwise. `junit_check` is the self-check: the exact fixed shape (`JUNIT_SHAPE`), then a yq parse (pinned `CI_YQ_IMAGE`, `--pull never --network none`) whose counts, suite name and class names must agree. `junit_discard` (parent) ensures no `junit.xml` or `.tmp` survives a failed report phase. `aggregate` and the order of every other step are unchanged. |
| `.gitlab-ci.yml` | `artifacts:reports:junit: ci-evidence/junit.xml` on `fast`, `gate` and `identity-staging`. `when`, `expire_in`, `access` and `paths` are unchanged; no rule, job, trigger or workflow change. |
| `infra/validation/run.sh` | The two B-CI-CONFIG `static_ci_config` artifact pins (fast/gate loop and identity-staging) gain `"reports":{"junit":"ci-evidence/junit.xml"}`. The B-CI-CONFIG PASS text is extended in place. `infra/ci/junit.sh` is added to B-SHELL's two lists; see Deviations. |
| `infra/ci/validate-controls.sh` | The scratch template copies `junit.sh`. Published sets gain `junit.xml`: `FULL_SET`, `FAST_SET`, `expect_pub` minimal, and group B's `b_negative` allow list. New stub modes `unknown-status`, `dup-extra` and `rm-results`, and docker-shim fault `SHIM_YQ_XML_FAIL`. Controls J1–J6 are registered in group D. |
| `docs/foundation/OPERATIONS.md` | Finalizer list; artifacts line; a "Test reports (G050)" section covering scope, format, status semantics, content and privacy, supplementary status, defect behaviour and limits. |
| `management/v0/results/G050-scope-allowlist.txt` | The approved scope copy. |

There are no migrations or schema changes, and no contract, public API or MCP changes. There are no dependency, image, profile-map, required-check or selection changes: `required-checks.tsv` and `profile-paths.tsv` are untouched.

**Commits.** Over BASE:
- `2a46dae`: implementation;
- `c472680`: cycle-1 corrections.

## Tests added

All are registered in group D, alongside V14–V18n, and reuse its stub harness and fixtures.
- **J1 (AC1):**
  - a passing fast run gives a `fast:fast` suite, with check cases in `results.tsv` order and no phase case;
  - an independent yq XML parse agrees on counts and finds no `+content` or `+directive`;
  - no `<!`;
  - the job-log line is present;
  - a second identical run gives byte-identical XML.
- **J2 (AC1, AC2, AC4):**
  - replayed rows: PASS, FAIL, UNMET, plus INFO rows (PREFLIGHT, IMAGES, and B-SCOPE-EXCEPTION twice);
  - every detail cell carries XML-sensitive text and a random canary;
  - exact cases and the fixed messages; INFO omitted;
  - the canary is in neither the XML nor the job log.
- **J3 (AC2):**
  - a docs-only MR gate gives an empty `gate:none` suite;
  - in a fast run, none of the 46 required IDs outside the profile is invented;
  - a reported out-of-profile row is shown as reported, together with `phase:complete`.
- **J4a (AC2):** raw defects on harness exit 0 (malformed line, unknown status, missing `results.tsv`).
  - Each gives `complete=14`, overall 14 and `junit=0`, with `phase:complete` plus only the well-formed rows.
  - With harness exit 1, overall stays 1.
- **J4b (AC2):** the converter alone.
  - Eight sanitized defects each return 19 with nothing written: bad line, unknown status, duplicate, missing columns file, job label, profile label, phase name, phase status.
  - Executed profile `none` gives an empty `gate:none` suite.
  - Harness `not-run` gives `phase:preflight` only.
- **J4c (AC3):** on an otherwise successful job, two report defects: a duplicate non-required PASS ID, and a failed parse self-check (shim).
  - Each gives `junit=19`, overall 19 and no `junit.xml`.
  - The fast set is otherwise published unchanged.
- **J5 (AC3):**
  - Harness 1 with a valid report gives exit 1, with the FAIL row, `phase:harness` and `phase:complete`.
  - Harness 1 with a report defect still gives exit 1, `junit=19` and no XML.
  - Postcheck, cleanup and scrub ran; the harness ran once each time.
- **J6 (AC2, AC4):**
  - `phase:complete` without a failing check (exit 14);
  - STAGING-FAILED: the check cases plus `phase:stage` (exit 19);
  - WITHHELD from the raw scan (V5a plant): phase cases only; the canary is in neither the published files nor the job log;
  - WITHHELD from gitleaks (V5c plant): phase cases only;
  - WITHHELD from the candidate scan, using a gitleaks token in a tracked comment of the case tree outside the raw evidence:
    - `junit_withheld` publishes `phase:candidate_scan` only;
    - `raw_scan=0`, `candidate_scan=15`, `summary_withheld`, `junit_withheld` and `rescan` all 0;
    - exit 15.

## Validation (exact commands and outcomes)

| Step | Command | HEAD | Exit | Outcome |
|---|---|---|---|---|
| static | `bin/g050-static.sh`: `bash -n`; shellcheck 0.11.0 with B-SHELL's flags (`-S warning -x -P SCRIPTDIR`, then `-S info --include=SC2251`) on `run.sh`, `validate.sh`, `lib.sh`, `junit.sh` and `validate-controls.sh`; `git diff --check BASE`; changed paths against the approved scope | `2a46dae`, `c472680` | 0 | clean; all 7 changed paths in scope |
| direct unit | the job-dir script `g050/xml/unit.sh`: the generator on a mixed fixture, escaping, the empty suite and 8 error paths | working tree | 0 | as designed; deterministic |
| shape | `g050/xml/shape.sh`: `JUNIT_SHAPE` against generator output and 4 tampered variants (extra attribute, `system-out`, free message, comment) | working tree | 0 | 2 match, 4 rejected |
| `c1-focused` | `bin/g050-controls.sh c1-focused J1,J2,J3,J4a,J4b,J4c,J5,J6,V1,VP1,VP9,V5a,V5c,V14,V16,V18c,N1,C-SEC-10` | `2a46dae` | 1 | 17 passed, 1 failed: J4a, a fixture defect (see Cycles) |
| `c1-retest` | `bin/g050-controls.sh c1-retest J1,J2,J3,J4a,J4b,J4c,J5,J6,V1,VP1,VP5,VP9,V5a,V5c,V14,V16,V18c,N1,C-SEC-10` | `c472680` | 0 | 19 passed, 0 failed |
| precheck | `bin/g050-vw.sh check` | `c472680` | 0 | `precheck passed (27 inputs, snapshot aa3ae9e7…)` |
| `c1-lint-provenance` | `bin/g050-vw.sh c1-lint-provenance lint provenance`, i.e. `run-approved.sh … -- --base 0d52421… --allow-file management/v0/results/G050-scope-allowlist.txt --expect-hash <scope>=60ba724d… --report-dir … lint provenance` | `c472680` | 0 | `RESULT: PASSED`; 26 PASS: lint 17 (incl. B-SHELL, B-CI-CONFIG), provenance 6 (incl. B-PROTECTED, B-SKILL-UNCHANGED, B-RENDER-PLAN), final 3 (B-SECRET-CANARY, B-SCOPE, B-REQUIRED-CHECKS); `run-approved: harness exit 0; postcheck passed`; post-state ok |
| CI lint | `glab ci lint .gitlab-ci.yml` (read-only API) | `c472680` | 0 | "CI/CD YAML is valid!" |
| read-only privacy review | one subagent per Plan §6, on `2a46dae` | — | — | no critical, high or medium finding. Five low findings and six nits, all addressed in cycle 1 except N6 (ID length, not a new exposure) |

**Not run locally, by owner policy (execution release):**
- the Plan's S3 final local `all`;
- the full group A–G controls (B-CI-CONTROLS).

The MR's hosted fast/full pipeline supplies them. The distinct local-only evidence is kept: the BASE-aware B-SCOPE, B-PROTECTED and wrapper pre/post checks, and the focused controls.

## Pre-MR rebase and binding reconciliation

- **Rebase.** `git fetch origin`, then the branch was rebased onto `origin/staging` `e59bc405b6dc257a5acec6e5b10a348b02422515` (tree `27fa6c50caefab55b41bade8261f28ac43b5d975`) without conflicts.
- **Target changes since BASE:** management records only: `management/v0/PROGRESS.md` and the new `management/v0/results/G050-G005-execution-release-codex-confirmation.md`.
- **Commits after the rebase:**
  - `2a46dae` became `1242987`;
  - `c472680` became `b31c03e`.
- **Reconciled binding:** `<home>/.local/state/platform-evidence/g050-rebase-e59bc40/binding-reconciled.json` (`bb4fcfc1…9ec81`; produced by `bin/g050-rebase-reconcile.py`).
  - Same 27 paths, `path_re`, Plan, review and 8-expression scope (`60ba724d…`).
  - One changed input: `management/v0/PROGRESS.md` `c208e195…10ba8` → `7120f74a…4c824`.
  - Snapshot `aa3ae9e7…` → `d095df9a7bc1cf90a17491007ae75656bef9ac8972a8c5acbb8d9f5446e14c35`.
- **The G050 diff is identical over both bases:** the same 7 paths, the same post-image blobs and the same `git patch-id --stable` `b3b7ab23…`.
- **Precheck.** `bin/g050-precheck-reconciled.sh` reports `precheck passed (27 inputs, snapshot d095df9a…)`.
- **Local runs.** The BASE-aware local evidence above (`c1-lint-provenance`) was run on `c472680` against BASE `0d52421`. Its content is unchanged by the rebase, so the scope result is identical. The hosted pipeline runs on the rebased source.

## Hosted MR pipeline (S4)

**MR !9.** https://gitlab.example/<namespace>/<project>/-/merge_requests/9
- Source `g050-gitlab-test-reports` at `b0ac030affd2a7d094b8c58e72ffa31cdf2e6099`, target `staging`, source-branch deletion set.
- Pushed after the rebase onto `e59bc40`. Quota read-only just before the push: October 2026 at 3478 minutes.

**Pipelines.** Both ran on `b0ac030`. None was started, retried or cancelled by this session.

| Pipeline | Source | Job | Status | Duration | Selection (job log) | Report |
|---|---|---|---|---|---|---|
| **2930980752** (evidence of record) | `merge_request_event` (`mr-source`) | `fast` 17063444297 | success | 360 s | `pipeline=full executed=fast reason=foundation base=e59bc40…` | `junit: junit.xml with 33 cases, 0 failing`; `final status 0; published=full` |
| 2930980752 | | `gate` 17063444298 | success | 3807 s | `pipeline=full executed=full reason=foundation base=e59bc40…`; `run.sh all` | `junit: junit.xml with 77 cases, 0 failing`; `final status 0; published=full` |
| 2930980440 | `push` (first push of a new branch; automatic under the existing workflow rules) | `fast` 17063441798 | success | 370 s | `pipeline=fast executed=fast reason=feature-push` | 33 cases, 0 failing |

**Hosted controls.** In the gate's full run, B-CI-CONTROLS passed all 281 controls, including J1–J6 (`ci-controls: 281 passed, 0 failed (of 281 selected; 281 defined)`). Every phase was 0, including `junit`.

**GitLab consumption.** Read-only: the job artifacts, `GET /projects/<gitlab-project-id>/pipelines/2930980752/test_report` and `…/test_report_summary`.

| Job | Artifact `junit.xml` suite | Cases | `results.tsv` non-INFO rows | `summary.json` | API suite | API total / success / failed / skipped / error | Names and order |
|---|---|---|---|---|---|---|---|
| `fast` | `fast:fast` | 33 | 33 | `executed_profile` fast, `required_count` 33 | `fast` | 33 / 33 / 0 / 0 / 0 | the junit check order equals the `results.tsv` order; the API name set equals the `results.tsv` IDs; class `validate.fast.check` |
| `gate` | `gate:full` | 77 | 77 | `executed_profile` full, `required_count` 77 | `gate` | 77 / 77 / 0 / 0 / 0 | the same; class `validate.gate.check` |

- **Pipeline totals:** 110 / 110 / 0 / 0 / 0 (the summary API agrees). Push pipeline 2930980440: 33 / 33, the same match.
- **Expected counts** come from the final candidate's `summary.json` `required_count`, not an assumed figure.
- **Suite identity.** GitLab names each test suite after its job (`fast`, `gate`). The `JOB:PROFILE` label and the `validate.JOB.*` class distinguish the fast and full suites inside the XML.

**Access behaviour.** Recorded without any setting change.
- The project is `private` (`public_jobs: true`, `builds_access_level` and `merge_requests_access_level` enabled). The jobs' artifacts keep `access: developer`.
- The test-report and summary APIs, and the artifact download, returned data for this session's existing authenticated project credential.
- Not tested: a lower role's or a non-member's view of the Tests tab and API. No other credential was used.
- By design the report holds only check IDs, statuses, fixed messages and phase codes (OPERATIONS.md "Test reports").

**Evidence:** `<home>/.local/state/platform-evidence/g050-exec-0d52421/hosted/`, per pipeline:
- `p<id>/jobs.json`, `test_report.json`, `test_report_summary.json`, `project-access.json`;
- per job: `trace.log`, `key-lines.txt`, `artifacts.zip` and the extracted `a/ci-evidence/`.

Scripts: `bin/g050-pipeline-wait.sh` and `bin/g050-hosted-evidence.sh`.

**This record is committed locally on top of `b0ac030` and not pushed.** For merge-request pipelines GitLab evaluates `rules:changes` against the whole MR diff, so pushing it would start a new full pipeline, which is not authorized. The pipelined, reviewed source is `b0ac030`.

## Acceptance criteria

| AC | Evidence |
|---|---|
| AC1 | J1 (independent yq parse, no external resolution; stable `JOB:PROFILE` suite and `validate.JOB.check`/`validate.JOB.phase` classes; exact counts; byte-identical regeneration); J2. In-job: the fixed-shape and yq self-check on every write. |
| AC2 | J2 (INFO omitted, B-SCOPE-EXCEPTION repeated); J3 (none suite, nothing invented, reported out-of-profile row with `phase:complete`); J4a (raw defects keep 14 with `phase:complete`); J4b (sanitized defects are 19 with nothing written; intentional phase-only cases); J6 (phase failure without a failing check). |
| AC3 | J4c (success plus report defect, or failed self-check: 19, no report, rest published); J5 (harness failure keeps 1 with or without a report defect; finalizer intact; one harness run). `aggregate` unchanged. |
| AC4 | J2 (detail canaries absent from XML and job log); J6 (WITHHELD phase-only on raw, gitleaks and candidate-scan paths; generated-secret canary absent from published files and job log); the report sits inside the candidate before both scans; C-SEC-10 and V5a/V5c pass with the report present; the read-only privacy review. Artifact access and retention unchanged. |
| AC5 | B-CI-CONFIG pins (lint suite in `c1-lint-provenance`); `glab ci lint`; affected A/B/C/D/G controls (V1, VP1, VP5, VP9, N1, V5a, V5c, V14, V16, V18c, C-SEC-10); diff touches no rule, profile, required check or selection; `gate:none` is not an error (J3, J4b). |
| AC6 | MR !9 pipeline 2930980752 on the final candidate `b0ac030`: `fast` (fast) and `gate` (full, the profile the accepted map selects for these CI/infra paths) both succeeded. Both jobs uploaded `junit.xml`. The test-report API (110/110; fast 33, gate 77) matches each job's `results.tsv` non-INFO rows and `required_count`. Access recorded. Distinct local-only evidence: `c1-lint-provenance` (BASE-aware B-SCOPE/B-PROTECTED, wrapper pre/post) and the focused controls. |
| AC7 | OPERATIONS.md "Test reports (G050)"; this Result. |

## Deviations (all mechanical; no design, security or scope change)

1. **Hosted CI replaces the S3 final local `all`.** This is the owner policy in the execution release, and AC6's local-`all` reference is replaced by the MR's hosted fast/full evidence. Distinct local-only evidence is kept.
2. **`junit.sh` is added to B-SHELL's two shellcheck lists** (`run.sh` `suite_lint`), beyond the Plan's "static_ci_config only" for `run.sh`. B-SHELL does not report findings inside a sourced file it was not given by name; a probe confirmed this. Without the addition, the new project script would not be linted.
3. **The `JOB` label also accepts `full`.** `validate.sh` sets `JOB_KIND=full` for a CI run without `--profile`. The workflow never uses that route, but existing controls such as V1 do.
4. **The withheld rebuild's pending list names `junit_withheld`.** Plan: "Both are added to … the summary's `pending` list". Each summary names the report phase that follows it.
5. **Group B's `b_negative` allow list also gains `junit.xml`.** An attribution refusal now publishes a phase-only report. This is the same published-set update as groups A, C and D.

Architecture deviations: none.

## Cycles (B3)

- **Cycle 1:**
  - Implementation `2a46dae`, then `c1-focused`. **J4a failed** in its missing-`results.tsv` case: exit 2, not 14.
  - Diagnosis: the stub's `STUB_RM_REPORT=results.tsv` removes the file before the stub's own `awk` over it, so the stub exits 2 under `set -e`. The product behaved correctly for harness=2, reporting `phase:harness` and `phase:complete`.
  - Correction `c472680`: an end-of-run `rm-results` stub mode, plus the privacy-review hardening (L1–L5, N1–N5).
  - `c1-retest`: 19/19.
- **Total:** 1 of 3 used; the checkpoint after cycle 2 was not reached.

## Unresolved issues and follow-up (no new scope)

- **N6:** check IDs have no length bound (`[A-Z][A-Z0-9-]*`). This is not a new exposure, since the same IDs already appear in the job log.
- **Shared files with G005:** `run.sh` and `validate-controls.sh` are edited by both Goals, in disjoint regions. G005 edits `required-checks.tsv` and `profile-paths.tsv`, which are in G050's snapshot. Whichever integrates second rebases onto the other's accepted bytes and reconciles its binding. G005 grows the full profile from 77 to 81 required IDs; the report counts are dynamic (computed from the rows).
