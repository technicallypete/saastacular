# Test and validation index

Use this page to translate a live `N/TOTAL ✅/❌ STATUS CHECK-ID` line into the
suite that produced it and the code or scenario documentation behind that check.
It is an index, not a second test runner or a replacement for the executable
checks.

## Sources of truth

| Question | Source |
|---|---|
| Which suite owns a check ID? | [`infra/validation/required-checks.tsv`](../../infra/validation/required-checks.tsv) |
| Which suites a profile selects, and how registered feature areas are added | [`infra/ci/lib.sh`](../../infra/ci/lib.sh), [`infra/ci/select-profile.sh`](../../infra/ci/select-profile.sh), and the profile table in [`OPERATIONS.md`](OPERATIONS.md#ci) |
| What a check executes and what counts as PASS | [`infra/validation/run.sh`](../../infra/validation/run.sh) and shared helpers in [`infra/validation/lib.sh`](../../infra/validation/lib.sh); search the check ID or suite function |
| Browser scenarios and their assertions | The `web/tests/browser/*.spec.ts` files and the linked foundation guides below |
| How to synchronize browser form interactions and investigate intermittent failures | [Browser test authoring guidance in `WEB-SHELL.md`](WEB-SHELL.md#authoring-reliable-form-interactions) |
| Service unit-check commands | The `platform:check` blocks in the root [`README.md`](../../README.md) |

The TSV is the authoritative ID-to-suite map. Keep this suite-purpose index in
sync when a suite's scope changes; do not duplicate the TSV's complete ID list
here. The PASS message emitted by `run.sh` is the compact executable description
for an individual ID.

## Reading progress and failures

The live `N/TOTAL` counter advances once for each required check recorded by
`record()` in `infra/validation/lib.sh`. A check can contain many lower-level
tests. For example, `B-BROWSER-ROOT` wraps the root Playwright matrix, while
`B-IDENTITY-E2E` wraps many identity browser projects and scenarios. The check
line shows suite progress; its report and diagnostics show which underlying
assertion failed.

`B-REQUIRED-CHECKS` is a final accounting check. It fails when an executed
required ID is missing, duplicated, or not PASS, so it can fail as a consequence
of an earlier check failure or an unexecuted suite.

## Profile selection

The stable base suites and counts are listed in [`OPERATIONS.md`](OPERATIONS.md#ci).
Feature suites are selected from registered areas and may add IDs to the base
profile count. `affected` is computed from changed paths, dependencies, and
registered area metadata; its size varies.

`ci-controls` exercises the validation harness itself. It is not a routine
application-feature suite: the `all` selector omits it when no relevant
CI/harness/control dependency changed. A forced or scheduled full run and an RC
promotion qualification retain it. Ordinary feature qualification selects the
affected product suites unless its change reaches that harness dependency
surface. The selection rule is implemented by `ci_controls_required()` in
[`infra/ci/lib.sh`](../../infra/ci/lib.sh) and applied by
[`infra/validation/run.sh`](../../infra/validation/run.sh) and
[`infra/ci/select-profile.sh`](../../infra/ci/select-profile.sh).

## Suite index

| Suite | What it covers | Start here |
|---|---|---|
| `lint` | Shell, Dockerfile, Compose, environment examples, dependencies, migration-set, Docker ignore, CI configuration, and submitted-source secret guards. | [`suite_lint`](../../infra/validation/run.sh#L346); [`SECURITY.md`](../SECURITY.md#21-static-checks) |
| `lint-base` | Repository-wide literals, naming, identity boundaries, and dependency vulnerability scanning. | [`suite_lint_base`](../../infra/validation/run.sh#L360); [`SECURITY.md`](../SECURITY.md#21-static-checks) |
| `lint-web` | Web server/client boundaries, theme tokens, component and design constraints, and asset integrity. | [`suite_lint_web`](../../infra/validation/run.sh#L367); `web/tests/browser/` |
| `lint-core` | Core MCP source-boundary guard. | [`suite_lint_core`](../../infra/validation/run.sh#L375) |
| `controls` | Pinned dependency and validation-wrapper negative controls. | [`suite_controls`](../../infra/validation/run.sh#L389) |
| `ci-controls` | CI/control-harness behavior, including its resource, lifecycle, runner, and failure-recovery controls. | [`suite_ci_controls`](../../infra/validation/run.sh#L396) |
| `provenance` | Protected paths and unchanged skill roots (the bundled `distributions/skills/` root and its discovery aliases) against the base commit; there are no renderer checks. | [`suite_provenance`](../../infra/validation/run.sh#L1561) |
| `contracts` | Contract tooling, Go and Python validators, validator agreement, peer compatibility, and negative cases. | [`suite_contracts`](../../infra/validation/run.sh#L1756); [`README.md`](../../README.md#quick-start) |
| `unit` | Root `platform:check` block across services, without a stack. | [`suite_unit`](../../infra/validation/run.sh#L404); [`README.md`](../../README.md#quick-start) |
| `unit-web`, `unit-core`, `unit-intel` | One service’s README check block plus `secret-scan`, without a stack. | [`suite_unit_web`, `suite_unit_core`, `suite_unit_intel`](../../infra/validation/run.sh#L442) |
| `root` | Root Compose configuration, ports, network boundaries, HTTP/health behavior, browser matrix, error output, build context, tenancy, and documented start/check/stop commands. | [`suite_root`](../../infra/validation/run.sh#L1876); [`WEB-SHELL.md`](WEB-SHELL.md); `web/tests/browser/` |
| `migrations` | Migration CLI/catalog/environment and PostgreSQL integration boundaries for tenancy, identity, profiles, policies, RLS, objects, audit, retention/recovery, MCP, domains, branding, onboarding, membership, enrollment, signup, isolation, labels, and platform address. | [`suite_migrations`](../../infra/validation/run.sh#L2936); [`RECOVERY.md`](RECOVERY.md); [`MCP.md`](MCP.md); [`IDENTITY.md`](IDENTITY.md) |
| `ops` | Observability, outage and persistence behavior; identity emulator, browser, privacy, probes and configuration; MCP browser coverage; secret rotation. | [`suite_ops`](../../infra/validation/run.sh#L5446); [`IDENTITY.md`](IDENTITY.md#evidence-layers); `web/tests/browser/identity.spec.ts` |
| `rename` | Renamed Compose project, port overrides, source hash, browser behavior, and database rename. | [`suite_rename`](../../infra/validation/run.sh#L5447) |
| `off` | Optional intelligence service disabled, including observability, browser, and toggle behavior. | [`suite_off`](../../infra/validation/run.sh#L5448) |
| `staging` | Real WorkOS staging contract; requires the separate identity-staging job and its authorized environment. | [`suite_staging`](../../infra/validation/run.sh#L6458); [`IDENTITY.md`](IDENTITY.md) |
| `feature-web` | Browser specs for the selected registered web feature areas. | [`suite_feature_web`](../../infra/validation/run.sh#L6517); `web/tests/browser/` |
| `feature-core` | PostgreSQL integration tests for selected registered core areas that contain integration-tagged Go tests. | [`suite_feature_core`](../../infra/validation/run.sh#L6618); `core/` test files |
| `final` (always appended) | Docker-only execution, secret-canary scan, change-scope check, and required-check accounting. | The `final` rows in [`required-checks.tsv`](../../infra/validation/required-checks.tsv) and the finalization path in [`run.sh`](../../infra/validation/run.sh) |

## Examples

- `B-BROWSER-ROOT` belongs to `root` and runs the documented root browser matrix
  across shell, token, appearance, and development-hydration projects.
- `B-IDENTITY-E2E` belongs to `ops` and runs
  [`IDENTITY.md`’s `identity:browser` block](IDENTITY.md#evidence-layers)
  against disposable HTTPS identity, domain, MCP, onboarding, membership,
  enrollment, signup, and labels projects.
- `B-IDENTITY-PRIVACY` also belongs to `ops`; it checks catalogued identity
  logs and identity-related stored values with the existing scanners.
- `B-CI-CONTROLS` belongs to `ci-controls`; it qualifies the validation harness,
  not the application's feature behavior.
