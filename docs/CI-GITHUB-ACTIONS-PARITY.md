# GitHub Actions parity guide

GitLab CI is the supported CI provider for this repository: [`.gitlab-ci.yml`](../.gitlab-ci.yml) and its scripts
under [`infra/ci/`](../infra/ci) are shipped and documented in [OPERATIONS.md](foundation/OPERATIONS.md#ci). **No
GitHub Actions workflow is shipped**, and no GitHub implementation, adapter or hosted run exists for this repository.
This page maps the existing GitLab jobs and profiles to the work needed to reproduce them on GitHub Actions. It is
migration guidance, not a ready-to-use workflow, and it makes no parity, timing or cost claim.

No hosted pipeline (GitLab or GitHub) has been run for this copy of the repository.

## What is shipped and reusable

The validation logic is provider-neutral and runs in Docker:

| Reusable piece | What it does |
|---|---|
| `bash infra/ci/validate.sh --profile fast\|integration\|full\|affected` | the single validation entry point: profile selection, capacity checks, evidence scanning against generated secrets and canaries, owned-resource cleanup ([README](../README.md#quick-start)) |
| `infra/ci/select-profile.sh`, `infra/ci/lib.sh`, `infra/ci/profile-paths.tsv` | changed-path classification into profiles and registered feature areas |
| `infra/validation/run.sh` and `infra/validation/required-checks.tsv` | the suites and the required check IDs each profile must record as PASS ([TEST-INDEX.md](foundation/TEST-INDEX.md)) |
| `infra/ci/validate.sh` evidence (`ci-evidence/`, `junit.xml`) | the published, secret-scanned evidence directory and the supplemental JUnit report |

The GitLab-specific layer is [`infra/ci/hosted-context.sh`](../infra/ci/hosted-context.sh): it runs only inside a
GitLab CI job beside a job-private TLS Docker daemon, removes credentials from the job-created Git configuration and
proves no job credential remains before any nested container mounts the build directory, runs `validate.sh` in an
unprivileged context container, guards free disk space, and forwards identity-staging variables only in the approved
project, branch, pipeline source and environment. A GitHub port needs an equivalent adapter; this one is not reusable
as is.

## Job-by-job mapping

| GitLab job or rule | Behaviour in `.gitlab-ci.yml` | GitHub Actions work needed |
|---|---|---|
| `workflow` rules | no pipeline for pushes to feature branches, the integration branch or the default branch, or for tags; merge-request pipelines only when a non-documentation path changed; scheduled, web and API pipelines admitted; MR checks auto-cancel on a newer commit | choose `on:` triggers (`pull_request`, `workflow_dispatch`, `schedule`) and `paths`/`paths-ignore` filters to the same effect; use `concurrency` with `cancel-in-progress` for pull requests only; avoid duplicate `push` plus `pull_request` runs |
| `validate:checks` (stage `checks`) | every admitted pipeline; runs `hosted-context.sh -- --profile fast` (or the narrow affected profile for registered areas); 40 min timeout; artifacts `ci-evidence/` kept 90 days with developer access and a JUnit report | one required job running the fast profile through a GitHub adapter; upload `ci-evidence/` with restricted retention; publish the JUnit report |
| `validate:integration` (stage `integration`) | after `validate:checks` (ordering only, without its artifacts); runs `--profile gate`, the remainder of the integration or full coverage the changed paths require; optional manual job on ordinary merge requests, automatic for schedules, web and API pipelines, and automatic and required for a release-candidate promotion merge request; 3 h timeout | a second job with `needs:` on the fast job but no artifact download; decide which events make it required versus manual (`workflow_dispatch` or an environment approval); keep its clean checkout |
| release-candidate promotion rule | a merge request from `release/X.Y.Z-rc.N` to the default branch runs both jobs automatically and requires them; extra fetch of the integration branch and tags so `validate.sh` can prove the candidate | reproduce with a branch-name condition on `pull_request` and an explicit fetch of the needed refs; verify the merge-ref semantics GitHub uses before relying on it |
| `identity-staging` | only on a protected `identity-qualification` branch for web or scheduled pipelines, in the protected environment `identity-staging`, the only job that receives the environment-scoped WorkOS staging variables | a GitHub environment with required reviewers and environment-scoped secrets, restricted to a protected branch and to `workflow_dispatch`/`schedule`; an adapter that refuses the staging profile anywhere else |
| `staging:prepare`, `staging:check`, `staging:build`, `staging:deploy`, `staging:stop` | manual deployment mode selected by typed pipeline inputs (`action`, `environment`, `candidate`, `duration`), only for web/API pipelines on the protected integration branch; replaces the validation jobs; prepare and build hold no ID token; check, deploy and stop use OIDC ID tokens with per-action audiences; deploy and stop serialize per environment through a resource group; runs `infra/ci/staging-action.sh` | `workflow_dispatch` inputs with the same option lists and regexes; GitHub OIDC (`id-token: write`) federated to the cloud provider with per-action audiences or subject conditions; `concurrency` groups per environment; protected environments for deploy and stop. See [STAGING.md](foundation/STAGING.md) and [DEPLOYMENT.md](DEPLOYMENT.md) |
| runner tag `saas-linux-medium-amd64`, `docker:29.8.2-cli` image with a pinned `docker:29.8.2-dind` service, TLS on port 2376 | every job runs the CLI image beside a job-private, TLS-verified Docker daemon | a runner (hosted or self-hosted) where Docker Compose is available and capacity is qualified; keep an isolated daemon per job rather than a shared host daemon |

## Parity areas

**Triggers and changed-path selection.** Map feature pull requests, integration-branch and default-branch events,
manual runs and scheduled diagnostics to the fast, affected/integration and full profiles. Avoid duplicate branch and
pull-request runs. Select full validation when comparison inputs are missing or ambiguous; routine feature changes use
the relevant checks, while foundation changes require full coverage (`infra/ci/select-profile.sh` already encodes
this given a reliable base).

**Reuse versus provider integration.** Reuse `infra/ci/validate.sh` and the harness unchanged. Replace the GitLab
event/ref plumbing (`CI_PIPELINE_SOURCE`, merge-request refs, the comparison base, job variables, artifact paths) with
a GitHub adapter. GitLab merge-request and GitHub pull-request checkout/merge behaviour must be verified, not assumed
identical.

**Runner and Docker requirements.** Qualify Docker Compose availability, CPU, memory and disk (the GitLab adapter uses
a provisional 35 GiB start floor and a 4 GiB runtime reserve), timeouts and image-build/cache behaviour on the chosen
runner. Keep the isolated-daemon model with TLS and credential boundaries: privileged outer infrastructure, unprivileged
application and test containers. Never mount a production or developer Docker socket, and do not assume GitLab runner
capacity transfers unchanged.

**Gate behaviour and evidence.** Preserve the fast-job prerequisite, required-check enforcement through branch
protection, failure and timeout propagation, cleanup on failure, secret-safe artifacts and their retention. A skipped,
cancelled or failed required check must never become acceptance evidence.

**Secrets and trust.** Set least-privilege `permissions:` for the workflow token; never expose secrets to fork pull
requests or untrusted inputs; keep checkout credentials and deployment secrets out of test containers and published
evidence (the GitLab adapter removes credentials before any nested container starts). Restrict manual and scheduled
execution and cache access to those trust boundaries.

**Qualification effort.** On the chosen provider, run a passing fast, integration and full case as applicable and an
intentional-failure case, and measure actual runtime, storage use and billing before claiming parity. Treat every
sizing or timing figure in this repository as an estimate for GitLab, not a GitHub measurement. Do not claim a
ready-to-use workflow, equal cost or completed parity without that work.

## Project-specific values to re-pin

Whichever provider you use, the shipped configuration names project-specific values that you must replace with your
own before any hosted run: the approved project identity checked by `infra/ci/hosted-context.sh` and
`infra/ci/staging-action.sh`, the protected branch names (`staging`, `identity-qualification`), the runner tag, the
environment names in `infra/staging/ci-environments.txt`, and the cloud federation settings described in
[STAGING.md](foundation/STAGING.md). Refusal logic stays in place: a job in an unapproved project or context refuses to
forward credentials or deploy.
