> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G050 — GitLab test-report visibility: Plan (r2)

Goal: [G050-gitlab-test-reports.md](../goals/G050-gitlab-test-reports.md) (SHA-256 `0e3d9965a8b2713d69633b13167c57f04482f38235e78eb9431a01e56459a986`) · Spec: [G050-gitlab-test-reports-spec.md](../specs/G050-gitlab-test-reports-spec.md) (SHA-256 `eb353f08f8f2d88b2ee22e3bddb587a4e57222a87aa4f7c54b6d0e242b659ca7`). Both are read from staging `e7612062ac09fa45bd85336518b24fe77aa6a445`.
Author: Claude Code · Date: 2026-10-09 UTC · Status: **Plan r2 for Codex review. No implementation is authorized.** Budget: **B3**, checkpoint after cycle 2. Planning used none.
Review answered: `G050-plan-r1-codex-review.md` (SHA-256 `c8c9919d4250e093b0f2adc707b5a4ab9dad858705e157baa2d9fd7508eb02f5`; CORRECTION LOOP). The reviewed r1 is kept as G050-gitlab-test-reports-plan-r1.md (not included) (`911a7d2d…bd1a1`).
Planning inputs: G050-planning-inputs.json (not included) (r2). The r1 inventory is kept as G050-planning-inputs-r1.json (not included).

## 0. Baseline

- **Implementation baseline (provisional).** `150544058d152fd72f87955477b84ef374b16d46` (tree `ef05f411…`), the G004 MR !7 reconciliation. At planning time MR !7 was open and **not merged**; its MR pipeline 2928632937 was running, and `1505440` is not in staging.
- **Staging at planning.** `e761206`. Since `c1412355` it differs only by the G050 Goal and Spec, plus PROGRESS and ROADMAP records; no CI or harness file.
- **Before execution approval.** Bind the actual staging commit that integrates G004. If G049 has integrated by then, reconcile its shared-CI changes (§6).

## 1. Design

The job publishes a supplemental `ci-evidence/junit.xml` built only from fields `validate.sh` has already validated. GitLab reads it through `artifacts:reports:junit`. Nothing reruns, and no test-selection or exit-status rule changes.

**Sources, all validated.**
- **Checks:** `$EV/final/status-columns.tsv`. `do_complete` writes it, one `ID<TAB>STATUS` per well-formed `results.tsv` row, before the candidate scan. Free-text TSV details are never read.
- **Phases:** the numeric phase statuses in `ST[]`.
- **Labels:** the validated `PROFILE_ARG` and `EXECUTED_PROFILE` tokens.

**Generator.** A new `infra/ci/junit.sh`: one sourced pure-bash function, writing only to the candidate directory.
- It re-validates each input with the same regexes (`^[A-Z][A-Z0-9-]*$`, `PASS|FAIL|UNMET|INFO`, phase-name list, numeric status).
- It escapes every attribute (`& < > " '`), emits no DOCTYPE, entity, timestamp, duration, path, URL or text node, and gives each failure a fixed message.
- **Output:** one `<testsuites>` containing one `<testsuite name="JOB:PROFILE">`. JOB is `fast`, `gate`, `staging` or `local`; PROFILE is the executed profile, or `none`.
- **Cases:**
  - one `<testcase classname="validate.JOB.check" name="ID">` per exported check, in `results.tsv` order;
  - one `<testcase classname="validate.JOB.phase" name="phase:NAME">` per **failed** phase.
- `tests` and `failures` are computed from the emitted cases; `errors` and `skipped` are `0`.
- Identical input gives identical bytes.

**Status semantics (R2).**

| Source | JUnit |
|---|---|
| PASS | passing case |
| FAIL | `<failure type="FAIL" message="check failed"/>` |
| UNMET | `<failure type="UNMET" message="required prerequisite unmet"/>` |
| INFO (PREFLIGHT, PROBE, IMAGES, B-SCOPE-EXCEPTION) | **omitted**. These are informational records, not checks, and B-SCOPE-EXCEPTION may repeat |
| Check not executed or outside the profile | absent from the source, so never invented |
| Phase with numeric status ≠ 0, recorded before generation | `<failure type="phase-error" message="phase NAME exited N"/>`. A failed phase stays visible with no failing check |
| Non-numeric phase token (`skip`, `n/a`, `not-run`, `not-required`) | omitted |
| **Raw** `results.tsv` malformed line, unknown status, or missing file | The generator never sees these, because it never reads raw TSV. `do_complete` handles them first: it keeps only well-formed rows in `status-columns.tsv` and records `complete=14`. `aggregate` keeps 14 ahead of 19, unless a higher-priority status, such as a nonzero harness, applies. A published report carries the `phase:complete` failure plus only the well-formed rows, so it never shows completeness success. Without a safe candidate there is no report (fail-closed) |
| **Sanitized-input** defect: a `status-columns.tsv` line that is not exactly an ID, a tab and one of PASS, FAIL, UNMET or INFO, or the file missing although the harness executed and `complete` ran; also an invalid label | **converter error.** No `junit.xml` is written and the phase returns 19 |
| Duplicate non-INFO ID (allowed by completeness for non-required checks) | **converter error**, as above |
| Intentional phase-only cases: executed profile `none` (a gate after a fast-only pipeline), or the harness `not-run` or `not-required`, so `complete` did not run and there is no source | a valid suite with no check cases (only phase errors, if any); not an error. It is distinguished from missing data by `ST[complete]` not being numeric |

**Integration in `infra/ci/validate.sh`.**
- **New phases.** `junit` runs right after `summary`, and `junit_withheld` right after `summary_withheld`. Both are added to `PHASES` and to the summary's `pending` list. Each runs only when a candidate exists, so the report is always inside the candidate before `candidate_scan` or `rescan`, and never added after publication.
- **Content by candidate kind:**
  - **full**, or **minimal STAGING-FAILED** (raw scan clean): checks plus phase errors;
  - **minimal WITHHELD** (raw or candidate scan found unsafe content): phase errors only, with no check cases. That rule includes `candidate_scan=15` when the report is rebuilt;
  - **no candidate** (scanner failure 17, summary failure): nothing is published, so no report.
- **Self-check.** The written file is parsed with the pinned `mikefarah/yq:4.48.1` (`--pull never --network none -p xml`). It is the existing tool that preflight already pulls; no external resolution is used. The counts in the attributes must equal the emitted cases and failures; otherwise the file is removed and the phase returns 19.
- **Gating (R3).** `aggregate` is unchanged. The harness, signal, scan, complete, postcheck and cleanup statuses keep precedence. A nonzero `junit` on an otherwise successful job becomes 19 through the existing no-masking rule. On an already failing job, the original status stands, and the report or its absence is supplementary.
- **Unchanged order and steps.** Postcheck, cleanup, scans, staging, report and scrub keep their order and behaviour.

**Early and partial failures.** Location or preflight failures before the yq pull produce no candidate and so no report. The job's 16 and its log line remain authoritative. This is documented.

**Canceled and timed-out jobs.** GitLab may not run the job's finalization or upload artifacts for a canceled job, so it may show no report. The internal deadline (124) normally lets the finalizer publish before the job timeout. No retry or synthetic report is added.

## 2. Files

| File | Change |
|---|---|
| `infra/ci/junit.sh` (new) | the generator and self-check function (§1) |
| `infra/ci/validate.sh` | source the generator; the `junit` and `junit_withheld` phases; `PHASES`; the `pending` list |
| `.gitlab-ci.yml` | `artifacts:reports:junit: ci-evidence/junit.xml` on `fast`, `gate` and `identity-staging`. `paths`, `when: always`, `access: developer` and 90-day expiry are unchanged; no rule, trigger or job change |
| `infra/validation/run.sh` (`static_ci_config`) | the exact artifact pins for the three jobs and the B-CI-CONFIG PASS text. Rules and job keys are unchanged |
| `infra/ci/validate-controls.sh` | the exact published-file sets (full and minimal) in the existing V-controls of groups A, C and D gain `junit.xml`; the focused controls in §4, registered in group D |
| `docs/foundation/OPERATIONS.md` | the Tests tab covers harness-level check IDs, not underlying unit or browser assertions; status semantics, with UNMET as failing and INFO omitted; supplemental, not authoritative; early-failure and canceled-job limits |

No change to `required-checks.tsv`, `profile-paths.tsv`, `select-profile.sh`, `hosted-context.sh`, images, dependencies or any application file. The artifact privacy boundary is reused unchanged: the report joins the candidate, so it gets the literal-secret scan, gitleaks and the forbidden-name scan. `junit.xml` does not match the forbidden names.

## 3. Sequence

1. **S0.** Bind the integrated G004 baseline and the inputs; reconcile G049 if it has integrated.
2. **S1.** Write `junit.sh`, the `validate.sh` phases, the controls and the docs, then run the focused checks (§4).
3. **S2.** Change `.gitlab-ci.yml` and the B-CI-CONFIG pins; run `run.sh lint` and the CI controls.
4. **S3.** One final local `all` through the approved wrapper.
5. **S4.** The final candidate's applicable MR pipeline. The paths are CI and infra files, so it runs `fast` plus `gate`, which selects **full**. Read-only: the uploaded `junit.xml` artifact and `GET /projects/:id/pipelines/:pid/test_report`. Names, statuses and counts must match `ci-evidence/results.tsv` minus INFO rows. Expected counts come from the final candidate's registry: `summary.json` `required_count` for each executed profile, not an assumed 33/77. The fast and full suites must be distinguishable by `JOB:PROFILE`. During this authorized evidence collection, also record how the existing project and report access behaves for the Tests tab and API, with no setting change.
6. **S5.** The Result.

`identity-staging` runs only on the protected qualification ref, so it is covered by configuration and control evidence rather than a live WorkOS run.

## 4. Validation

Focused controls are registered in validate-controls group D (staging and publication), next to V14–V18. They reuse its stub harness and fixtures:
- **J1 (AC1).** PASS only. The yq parse is valid, the counts match, and the suite is `JOB:PROFILE`. A second identical run gives byte-identical XML.
- **J2 (AC1, AC2).** Mixed PASS, FAIL, UNMET and INFO, with XML-sensitive and secret-like canaries in the TSV detail column. Exact cases and failures; INFO absent; no canary in the XML or the job log.
- **J3 (AC2).** Executed profile `none` gives an empty suite. A profile-excluded check is absent.
- **J4a (AC2): raw-input path, existing completeness.** Each case uses the stub harness with exit 0:
  - a malformed `results.tsv` line;
  - an unknown raw status;
  - `results.tsv` missing.

  Each gives `complete=14` and overall **14**, not 19. The published report shows `phase:complete` as failing, with only the well-formed rows. The completeness rules are unchanged. A variant with stub exit 1 keeps the harness status 1.
- **J4b (AC2): converter only.** The adapter function is called directly on sanitized fixtures. Each must return 19 and write no XML:
  - a bad `status-columns.tsv` line;
  - an unknown status;
  - the file missing while `complete` is numeric;
  - an invalid label.

  Two intentional phase-only cases must succeed with no check cases: `EXECUTED_PROFILE=none` and harness `not-run`.
- **J4c (AC3): end-to-end reporting defect on an otherwise successful job.** A duplicate non-required, non-INFO PASS ID, which completeness permits, gives `junit=19` and overall **19**. No `junit.xml` is published; the rest of the candidate publishes unchanged.
- **J5 (AC3).** A failing harness (stub exit 1) plus a valid report: the original 1 is preserved and the report shows the failure. A failing harness plus a generation error: still 1. Postcheck, cleanup and scrub run, and the stub's invocation count shows nothing reran.
- **J6 (AC2, AC4).** A complete-phase failure with no failing check gives a `phase:complete` failure case. A STAGING-FAILED path gives checks plus `phase:stage`. A WITHHELD path (the existing V5a and V5c plants) gives phase errors only, including the candidate-scan case, and no check case.
- **Existing V-control expectations** (groups A, C, D) are updated only to include `junit.xml` in the published sets. No assertion is removed.
- **Also:** the B-CI-CONFIG artifact pins (`run.sh lint`); C-SEC-* and the staging-profile controls (group G) with the report present; `glab ci lint` if Codex authorizes the read-only call; shellcheck; `git diff --check`.

**Final.** S3 local `all` (required: CI and harness inputs changed). S4 MR pipeline with GitLab consumption evidence. No extra application run, no duplicate pipeline and no WorkOS requalification.

## 5. AC mapping

| AC | Evidence |
|---|---|
| AC1 | J1, J2: yq parse with no external resolution, stable `JOB:PROFILE` and class names, exact counts, byte-identical regeneration |
| AC2 | J2, J3, J4a, J4b, J6: INFO omitted, excluded checks absent; raw defects stay `complete=14` with a `phase:complete` failure; sanitized defects fail closed with 19 and no XML; intentional phase-only cases distinguished; phase-error visibility |
| AC3 | J4c, J5: success plus a reporting defect is 19 with no report published; a harness failure keeps its status; finalizer steps intact; no rerun |
| AC4 | J2, J6: canaries absent; the report inside the candidate scan; WITHHELD gives phase errors only; existing artifact access and retention unchanged |
| AC5 | B-CI-CONFIG pins, `glab ci lint` and the A, C, D, F and G controls; the diff touches no rule, profile or required check; executed profile `none` is not an error |
| AC6 | S3 local `all`; S4 MR pipeline IDs, artifacts and the test-report API counts matched to `results.tsv` |
| AC7 | OPERATIONS.md text; a Result with bindings, commands, controls, pipeline and job IDs, limitations, cleanup and B3 |

## 6. Delegation, budget, coordination and questions

- **Delegation.** No implementation subagent. One read-only privacy reviewer checks the generator and fixtures after S1.
- **B3.** Cycle 1: S1–S2. Cycle 2: fixes, then the checkpoint. Cycle 3: reserve.
- **G049 overlap.** The shared files are `.gitlab-ci.yml` (G049: the workflow `changes` list; G050: job `artifacts` only), `run.sh` `static_ci_config` (adjacent pins and the PASS text), `validate-controls.sh` (G049: group F and `mk_fbase`; G050: groups A, C and D) and OPERATIONS.md. The hunks are disjoint. Whichever Goal integrates second rebases onto the other's accepted bytes and reruns B-CI-CONFIG and the controls. Neither Goal blocks G005.
- **Decisions from the r1 review (recorded):**
  - **Q1:** INFO is omitted, never shown as skipped.
  - **Q2:** only the sanitized check IDs, statuses and fixed diagnostic categories may appear in the Tests tab. `artifacts:access: developer` and the project visibility are unchanged. OPERATIONS.md states that reports contain no sensitive or raw detail. S4 records the actual access behaviour, without assuming that download restrictions equal UI/API visibility. No setting or privacy-model change.
  - **Q3:** failed-job publication is proven locally (J4a, J4c, J5, J6); there is no deliberately failing hosted run.
- **Open questions:** none.

## Revision record

- 2026-10-09 UTC: r1 (archived `-r1.md`, `911a7d2d…`). Short Plan against the provisional baseline `1505440` and the exact G050 inputs from `e761206`. Planning only. B3 and checkpoint 2 unchanged.
- 2026-10-09 UTC: r2. Answers the Codex CORRECTION LOOP (`c8c9919d…`):
  - raw-input defects stay with the existing `complete=14` precedence and are shown as `phase:complete`;
  - sanitized-input and duplicate defects are converter errors (19, no XML);
  - intentional phase-only cases are distinguished from missing data;
  - J4 is split into J4a, J4b and J4c;
  - S4 counts come from the candidate registry, and access behaviour is recorded;
  - the Q1–Q3 decisions are recorded.

  No change to completeness, scope, the generator's sources or the budget. Planning only.
