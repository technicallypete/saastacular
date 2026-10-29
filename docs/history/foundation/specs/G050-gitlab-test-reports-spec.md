> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G050 — GitLab test-report visibility work specification

Goal: [G050-gitlab-test-reports.md](../goals/G050-gitlab-test-reports.md)
State: Defined for Claude planning; implementation requires separate Plan ACCEPT

## Rationale and Scope

Current CI runs checks and publishes scanned TSV/log evidence, but registers no JUnit XML reports. GitLab therefore shows zero reported tests. Publish a supplemental view of existing harness checks and CI phase failures. The reported count is the number of exported harness checks, not the total underlying unit/browser assertions.

Use the accepted integrated G004 baseline, existing job/container tools and privacy-controlled publication path. Include fast and gate, and identity-staging's configuration/adapter compatibility without requiring another real WorkOS qualification. No schema, application/public API/MCP behavior, dependency, runner, workflow/profile selection or G049 relocation changes. Existing governance, accepted ADRs, artifact access/retention, exact-input protection and G021 remain binding.

## Requirements and Architecture/Security Constraints

**R1 — Supplemental deterministic reports.** Produce valid UTF-8 JUnit XML from existing attributable outcomes. Stable suite/class/case identifiers distinguish job/profile and existing check IDs without embedding timestamps, secret values or volatile paths. Each executed check appears once in its suite; never expand aggregate checks into invented individual tests. Repeated check IDs across separate jobs remain distinguishable. Totals must equal emitted cases/statuses. Durations are optional and may be omitted if not already measured; never fabricate them.

**R2 — Honest status semantics.** PASS is a passing case; FAIL and UNMET are failing cases (UNMET is an unmet required prerequisite, not an acceptable skip). INFO does not become a passing test. A genuinely unexecuted optional/profile-excluded check is omitted, not reported as passed. Explicit supported skips, if any exist, require a distinct skipped representation and do not change existing gate policy. Missing, malformed, duplicate or unknown-status required data cannot yield an apparently successful report. Partial failures may include validated completed outcomes plus a bounded report/phase error, but must not invent outcomes for unexecuted checks. A failed CI phase without a failing check must remain visible through a separately named phase-error case when publication is safe.

**R3 — Preserve gating and failure publication.** Register report paths through existing jobs' `artifacts:reports:junit` and retain `when: always`. Keep original evidence and test commands. Successful reporting cannot mask a failing harness/phase or alter its established exit status. Report generation/validation/publication failure on an otherwise successful job must fail the job; on an already failing job retain its original failure status and add bounded diagnostics. Preserve pre/post input checks, cleanup and after-script behavior. Canceled/timed-out jobs may lack reports if GitLab cannot run/upload them; document this platform limitation without manufacturing success or adding retries/test executions.

**R4 — Privacy before publication.** Reports join the accepted candidate scan/publication boundary; never append unscanned XML after publication. Export allow-listed IDs/statuses and fixed diagnostic categories, not TSV free-text details, raw exceptions/stacks, stdout/stderr, URLs, provider identifiers, cookies, keys, environment dumps or filesystem paths. Preserve developer-only artifact access and retention. If existing privacy controls block artifact publication, no raw JUnit fallback is allowed. XML escaping and bounded input validation must prevent malformed documents or arbitrary content injection; readers/tests do not enable external entities/network resolution.

**R5 — Minimal compatible integration.** Use existing image/tool capabilities with no added dependency. Adjust only concrete artifact allow-lists/configuration guards needed for the report. Reuse current controls and profile selection unchanged. Reporting performs no network/provider requests, application startup, migrations or additional test runs. Do not create a second test-result authority or replace the complete-required-check validator.

## Expected Artifacts

One bounded adapter (or equivalently small existing reporting extension), minimal existing CI/artifact/guard changes, focused tests, operations documentation stating count/status/limitations, concise Claude Plan and Result. Concrete paths and changed guards belong to the Plan.

## Acceptance Criteria and Required Verification

| ID | Required outcome | Verification / evidence |
|---|---|---|
| AC1 | Reports are deterministic, valid JUnit XML with unique stable identifiers and accurate counts. | Parse with an independent existing XML reader using no external resolution; fixtures cover PASS, FAIL, UNMET, INFO and XML-sensitive input. Identical input produces identical XML; verify aggregate-case labels/counts and job/profile distinction. Do not require GitLab case counts to equal underlying browser/unit totals. |
| AC2 | Failure, exclusion and incomplete-input semantics cannot create false passing evidence. | Focused fixtures cover mixed outcomes, excluded/unexecuted checks, missing input, malformed row, duplicate ID, unknown status and a failed phase with no failed check. Assert exact outcomes and fail-closed behavior; retain every existing completeness check. |
| AC3 | Existing job status survives reporting; report defects fail an otherwise successful job. | Exercise combinations of successful/failing harness with successful/failing conversion/publication using existing bounded controls. Assert original nonzero status is preserved, success-plus-report-error is nonzero, post-checks/cleanup still run, and no tests are rerun by conversion. |
| AC4 | Safe failure reports are published without exposing sensitive content or bypassing scans. | Add focused canary fixtures containing secret-like/raw-detail values and XML metacharacters; assert no canary/raw content reaches XML/logs. Existing artifact privacy controls cover the new report and a rejected scan leaves no unsafe publication. Validate failed-job safe-report publication, artifact allow-lists and access/retention. |
| AC5 | Existing jobs register JUnit artifacts without changing workflow, profiles or required execution. | Lint actual GitLab configuration and covering guards; inspect fast/gate/identity-staging configuration and final diff. Run existing affected CI controls. Prove profile-excluded/fast-only gate behavior does not become a report failure or extra test execution; no live provider run required for unchanged identity behavior. |
| AC6 | GitLab consumes reports from the final candidate and the actual changed-path validation passes. | After focused checks, run required existing final local/hosted validation selected by actual CI changes (full where the accepted map requires it). On that authorized candidate pipeline, inspect uploaded report artifact and GitLab test-report API/UI; names/status/counts match source evidence. Reuse this pipeline for visibility evidence, not an additional benchmark or suite. A controlled failed-report fixture can prove failure mapping locally; no deliberately failed hosted full run is required. Record pipeline/job/commit IDs. |
| AC7 | Result and operations guidance preserve reproducible evidence and completion boundaries. | Record base/candidate commit/tree, approved inputs, changed files, exact commands/outcomes, fixtures and hosted artifacts/report counts, timing observations, failures/limitations, cleanup and B3 accounting. Document supplemental harness-level counts, UNMET semantics and cancellation limits. Codex reviews the actual diff; no product work or unrelated changes. |

## Provenance, Permitted Deviations and Completion Gates

Reuse accepted baseline/input wrappers and owned-resource procedures. Preserve historical evidence and protected predecessor checkouts. Pin actual final code/configuration/evidence, distinguish inspected Claude evidence from independent verification, and retain failed/interrupted runs. Minor adapter organization and suite labels belong to the Plan; status mapping, privacy, gating and scope changes require Codex review.

Before implementation, the Plan must identify existing XML tooling in the container, actual sanitized source/publication paths (including early-failure paths), minimum artifact/guard changes, and how safe partial failures are represented. It must state final path-selected CI requirements honestly: reporting itself adds no application runs, but changing CI/harness files still requires their existing regression gates. No automatic live WorkOS requalification or duplicate full campaign is requested. B3/checkpoint 2 and AC1–AC7 remain mandatory. This reporting follow-up is not a prerequisite to independent approved G005 work and does not change G049 scope or the G021 barrier.

Reference: [GitLab unit test reports](https://docs.gitlab.com/ci/testing/unit_test_reports/).

## Revision Record

- 2026-10-08 — Initial bounded specification; owner authorizes definition only. Claude Plan review and exact execution binding remain required.
