> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G003 — Final Codex review of hosted CI evidence

Date: 2026-10-07 UTC
Supervisor: Codex
Outcome: **ACCEPT — implementation and hosted CI evidence accepted; owner-authorized project settings verified**

Initial review outcome: CORRECTION LOOP, limited to the administrative merge-gate disposition below. Final acceptance supersedes that pending gate; its review history is preserved.

## Reviewed delivery

Execution checkout HEAD remains `cbffd6da5d9936c90fdfd9e51d84d36c0f17ad81`. The delivered code/validation files match hosted H1 commit `5ec78dec07c07f913b896de8185b180b4fb8d393`, tree `4dcdc04438a57fa938ad8af55cd4e988e1f0fa54`, byte-for-byte. The final Result is `management/v0/results/G003-foundation-observability-ci-result.md` in that checkout, SHA-256 `74b4e2caa1b14fab3f08c8a258ea9346e3fc2aaf9b689c6c20b738c4a72565c8`; its sections 1–15 retain digest `f98dcf00a02a7cafcafc74a71b848588770983bf6e3f214e7a43cdff9afc4d28` before the appended section 16. All nine approved inputs match snapshot `acae154f6acc7b439a20b11f4e647e19dbf91a9f26eb0c97bafce4362638e5ab`.

This review reuses the prior G003 correction review for unchanged application observability, contracts and isolation. It inspects the CI revision, the final one-file B-TOGGLE-OFF correction, profile/gate wiring, retained evidence, attribution, cleanup and live hosted job statuses. Codex performed read-only inspection only; no application edits, builds or repeated validation campaign.

## Findings and evidence

- H1 full: pipeline `2920018355`, job `16985151413`, authoritative status success; 69 distinct required PASS checks, 0 FAIL/UNMET. Its fast job passed separately. Gate duration 2478 seconds, about 41.3 minutes.
- H-INT: job `16985164485`, authoritative status success; 63 required PASS checks, 0 FAIL/UNMET, integration selected and executed from the reviewed runtime-only comparison. Its tree equals H1; gate duration 1783 seconds, about 29.7 minutes. Fast also passed.
- H2: job `16985165048`, authoritative status failed as intended. The invalid EI39 expectation causes the independent Go/Python/agreement checks to fail; the README unit aggregate and required-check aggregate fail accordingly. Harness status 1 survives finalization and safe evidence publication. This is a real contract fault, not a synthetic unrelated failing job.
- Browser evidence: root/off each 103 expected, rename 2 expected; zero skipped, flaky or unexpected. Independent contract validators each report 120 cases, the same schema/corpus hashes and no skips/mismatches. Full includes all 195 CI controls, inherited wrapper/dependency checks and foundation regressions.
- Final Python-image correction obtains the control image if absent and retains the existing PYTHON_VERSION detection and every disabled-intelligence assertion. Fresh-daemon diagnosis proves the original image-store assumption failed. Local present/absent/unavailable controls and final hosted `pulled` results support the correction without relaxing coverage.
- Profile selection fails to full on missing/ambiguous comparison or failed diff/sort; completeness checks reject missing/failed/duplicate/out-of-profile required rows. YAML keeps required failures fatal, orders gate after fast without downloading its artifacts, and preserves isolated TLS DinD/context and credential scans.
- Final wrapper inspection records passing pre/postchecks and no delivered-input drift. Hosted cleanup leaves only the permitted render-cache volume; temporary evidence branches/MRs are removed except the retained H1 draft MR. Result accounting is 4 of 6 revision cycles, original B7 6/7 and first correction pass 0/3 unchanged.

Artifact SHA-256 bindings:

| Evidence | SHA-256 |
|---|---|
| H1 full archive | `493034155fd4bf6941e0e1c1395bbe44be7ea4f699a11af786d75709dcc945f1` |
| Integration archive | `7a731aec06fdd8981000fbc86ae747e1326888d827ab3456c3c3938e07cbd3d6` |
| H2 failure archive | `b74c2aa567db0fce916149d9d4bc4d5386f94677e39ef639f9bdc27c35bdaa34` |

## Accepted dispositions

1. Accept **35 GiB free at start for qualified hosted integration/full**, with the existing 4 GiB runtime reserve and fail-closed sampler. Measured incremental peaks approximately 11.4–11.5 GiB satisfy the reviewed `max(35, ceil(1.5 × peak + 4))` formula; approximately 37.3 GiB was available initially. Retain local 40 GiB and all CPU/RAM checks. Fast 15 GiB remains supported by measured approximately 2.5 GiB peak. No new runner qualification or code edit is needed to record this disposition.
2. Accept the observed cold timings: approximately 5 minutes fast and 29.7 minutes integration meet the respective 10/30-minute cold objectives; full is approximately 41 minutes. No warm-performance claim or additional benchmark is required. Earlier failed/slow attempts remain disclosed. Broader integration acceleration and live failure reporting can be later bounded follow-ups, not blockers here.
3. Accept the documented reuse semantics: require authoritative job success plus **`pipeline_profile=full` and `executed_profile=full`**, exact tree, required-set digest and image inventory. `profile=gate` identifies the requested job kind; the stricter two-field semantic check supersedes Plan r2's shorthand. No report-schema or Plan revision is required.

## Remaining gate: owner-authorized project settings

The implementation and required hosted evidence need no further correction. Read-only project inspection still finds `only_allow_merge_if_pipeline_succeeds=false` and `ci_pipeline_variables_minimum_override_role=developer`. Consequently a failed required pipeline does not actually block merging, and custom pipeline-variable overrides can spoof the predefined inputs trusted by profile selection. These are the already disclosed owner-setting prerequisites, not new application work.

Minimum proposed administrative correction, requiring explicit owner authorization:

1. Set `only_allow_merge_if_pipeline_succeeds=true`; keep merging on skipped pipelines disabled.
2. Set `ci_pipeline_variables_minimum_override_role=no_one_allowed` to prevent custom pipeline variables from narrowing validation through spoofed predefined values.

After authorization, update only those settings and verify them by read-only API inspection. Preserve before/after values and record final acceptance in PROGRESS; do not rerun full/integration merely for project-settings changes. Staging protection, weekly schedule provisioning and additional retention remain separate operational owner decisions and do not require another implementation cycle for this review. Existing workflow/rules support them; no schedule is created by this review.

G003 remains unaccepted pending that disposition. No main/staging implementation integration, G048 release, OS001 resume, purchase or deployment is authorized here. Historical input bytes in the execution checkout remain untouched.


## Final acceptance — 2026-10-07 UTC

**ACCEPT.** The owner explicitly authorized the proposed GitLab project settings and final acceptance. Codex applied them to project `<gitlab-project-id>` and independently re-read the project API:

| Setting | Before | Verified after |
|---|---|---|
| `only_allow_merge_if_pipeline_succeeds` | `false` | `true` |
| `allow_merge_on_skipped_pipeline` | `null` | `false` |
| `ci_pipeline_variables_minimum_override_role` | `developer` | `no_one_allowed` |

These close the remaining merge/profile-override gate. The reviewed implementation, all applicable G003 acceptance criteria and required hosted evidence are accepted against the exact delivery bindings above. No additional implementation cycle, pipeline or application edit was needed. Original B7 accounting remains 6/7, first correction pass 0/3, CI revision 4/6.

Acceptance does not merge or push the execution checkout. G048 still requires predecessor integration, source/font rights and a fresh exact execution binding before its implementation release. OS001 remains paused. The product-feature barrier remains closed through G021.
