> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G072 — Manual staging and preview CI start/deploy and stop: Implementation Plan (r0)

Date: 2026-10-27 UTC. Author: G072 planner (Claude). Model note: Opus 5.5 xhigh under the standing Fable fallback
(the coordinator reported Fable's usage limit; not independently verified here). **Planning only.** No code, cloud,
IAM, GitLab-setting, DNS or live action is released by this Plan. Goal priority **high** (supervisor default);
execution tier **security/infrastructure** (Codex Plan ACCEPT and an execution release precede implementation; every
live item below needs its own named release). Budget **B3** (checkpoint after attempt 2); zero attempts spent.

## 0. Bound inputs

| Input | Value |
|---|---|
| Planning baseline | `origin/staging` **`fc2f84503f21eae2506346df202fdd3c400599f9`** (branch `g072-plan`) |
| G070 integrated source | merge `e8f5d9527f90a5cd34bb8ab786b41e1e95fee5e5` (MR !117); no non-management change between it and the baseline |
| G072 Spec r2 | `management/v0/specs/G072-manual-staging-ci-spec.md` sha256 `bbcbbc5e516686ea95ae43ae5659efd0f1707dad22b809544492df84e4aef726` |
| G072 Goal | sha256 `d42f40abdd87ba07d6340e08339f8cd4df3f887a21a5ec632f4fecf1dcda311b` |
| G070 Spec r4 (r2 runtime 280–297, r3 retention 299–325, r4 operator token 328–373) | sha256 `a0d9bbc97f0ad824545deda80217a1291c59e7032edae6dd189aa5eb910cc9d8` |
| G070 Plan incl. r1.9 | sha256 `4056ea2a3223f1648ede00a414977cfb34e35e21a84f8f0bb767cc971233f03a`; runtime review `105417be…7e783` (4 clarifications) |
| G070 Result | staging copy sha256 `932a1f4d…1b102` (ends at MR !117). **The L-5 rerun, L-6–L-8 and the A2 live proof are only on `origin/g070-closeout-4` `94f2e06c`** (commits `345533f3`, `34720f2c`, `8923ce0b`, `251101ad`; Codex-reviewed in `G070-interim-result-codex-review-20261027.md`, not yet published to staging). Read from there. |
| G057 Plan §18/§18.R1 (lines 1375–1548; hazard rule 1498–1501) | sha256 `8535114552dead832862967844dec1314844e45e6b8b8bcf86d17e0dc9aaa8c6`; review `052220f8…3cab` |
| ADR-009 | sha256 `0d504e10…c814` |
| `.gitlab-ci.yml` | sha256 `1cec903d…c3b5`: workflow 29–79, `validate:checks` 83–120, `validate:integration` 122–172, `identity-staging` 174–210 |
| `infra/staging/staging.sh` | sha256 `7c319382…3ee7`: usage 4–13, exit classes 45–48, `--duration` 80–86, `--pin-secret` 97–101, `cmd_build` 114–153 (control==candidate assumption at 120 and 133–138), `cmd_push` 156–186, `deploy_candidate` 250–284 (archive of the candidate's infra 271; VM log `tee` to stdout 280), `cmd_start` 368–419 (running-session refusal 391–394, `set_session` 405), `stop_vm`/`cmd_stop` 422–439, `cmd_status` 527–556 (Cloudflare token required 529), `write_receipt` 565–643 (`tf show` 607) |
| `infra/staging/lib.sh` | sha256 `9d11fe40…b06b`: `CONF_REQUIRED/OPTIONAL` 11–16, session bounds 20–22, `load_conf` 85–108 (ADC default 102, `GFLAGS` impersonation 107), `validate_conf` 110–139 (deployer name 116, AR host 117), `check_candidate` 164–167, `require_pushed` 192–202, `tf` 206–215 (ADC file mount, `CLOUDFLARE_API_TOKEN` by name), `need_cloudflare_token` 219–223, `vm_read`/`vm_status` 228–252, `vm_ssh` 399, `remote` 417–431 (runs `/srv/<env>/current` scripts), exact names 437–457, `label_check` 504–515, `reconcile` 603–624 |
| `infra/staging/remote/lib.sh` | sha256 `a98c5797…a796`: `IDENTITY_SECRETS` 42; `resolve_secrets` 256–306 (deploy default `latest` 278; identity all-or-none 306); `migrate_current` 396–403 (no "database ahead of candidate" refusal) |
| `infra/terraform/bootstrap/bootstrap.sh` | sha256 `92deadc6…672b`: roles 33–34, names 62–66 (state bucket `<project>-tfstate`), `bind_all` 167–187 (project-wide compute roles 169–172, bucket-wide `objectAdmin` 175), state prefix `staging/<env>` 202 |
| Terraform root | `environments/example/main.tf:28` provider impersonates the deployer; `modules/gcp-staging-vm/identity.tf:5–11` instance-level `osAdminLogin` for operator principals |
| `infra/ci/hosted-context.sh` | sha256 `6684b1d3…2886`: pinned project ID/path SHA 75–76, approved-context gate 86–96, private temp dir 118–125, context container 258–279 (pattern reused, file not edited) |
| `infra/validation/run.sh` | sha256 `f9a7ebe4…d53c`: `static_ci_config` (B-CI-CONFIG) 1587–1690 |
| `infra/staging/check.sh` | sha256 `d687a3c8…3c26`: usage 6–9, rows 19–110, `ROWS` 126 |
| `infra/ci/validate-controls.sh` | text parsers of `.gitlab-ci.yml`: 2172/2254 (`RR1`, `RR2`), 5932–5940 (`F-push-main`, `F-push-staging`, `F-agent-push-staging`, `F-agent-push-main`) |
| Not verified (no live read allowed) | GitLab tier/features (pipeline inputs in the root config, protected environments, `environment_protected`/`environment_action`/`ci_config_ref_uri`/`runner_environment` claims, `artifacts:access: none`, artifact size limit, pipeline-variable restriction value), `staging` branch protection settings, google-cloud-cli image contents/digest, gcloud/Terraform behaviour with an `external_account` file plus impersonation, direct federated Secret Manager access, OS Login handling of a pre-registered key with TTL. All are assumptions below with proofs. |

## 1. Settled decisions, unproven assumptions and the early proof

### 1.1 Settled (within Spec authority; Codex may redirect)

| # | Decision |
|---|---|
| S1 | **Pipeline inputs, not manual jobs.** `spec:inputs` in `.gitlab-ci.yml`: `action` ∈ {`validate` (default), `start-deploy`, `stop`}, `environment` ∈ {`none` (default), `staging`, `preview-1`}, `candidate` (string, regex `^([0-9a-f]{40}\|[A-Za-z0-9][A-Za-z0-9._/-]{0,199})?$`, default empty), `duration` (regex `^([1-9][0-9]{0,5}[mh]?)?$`, default empty), `images` ∈ {`build` (default), `published`}. Defaults keep MR, RC, schedule and ordinary web pipelines exactly as today. The selected action's jobs are ordinary (non-manual, no `allow_failure`), so a failure fails the pipeline and nothing unselected waits. |
| S2 | **Two actions only.** `start-deploy` = jobs `staging:build` → `staging:deploy`; `stop` = job `staging:stop` alone, in a fresh pipeline, no `needs`, no artifacts. No reset/retire/provision/status button; WIF setup is a bootstrap script. |
| S3 | **Deployment mode is admitted only** for `web`/`api` pipelines on protected `staging` (`$CI_COMMIT_REF_PROTECTED == "true"`); workflow `when: never` for a non-default action on any other branch or source (push, tag, schedule, MR, trigger, parent/child stay unchanged or refused). Validation jobs and `identity-staging` get a first rule `action != validate → never`; deploy jobs exist only for their action with `environment != none` (otherwise zero jobs: GitLab refuses the pipeline before any credential). |
| S4 | **Trusted profile per environment** = four GitLab **file** variables, protected and scoped to the exact environment name (`STAGING_CI_CONF` = the G070 `staging.conf` minus `EVIDENCE_DIR`/`PRIVATE_DIR`/`GOOGLE_ADC_FILE`; `STAGING_CI_BACKEND` = `backend.hcl`; `STAGING_CI_TFVARS` = `terraform.tfvars`, byte-equal to the local operator's so no plan drifts; `STAGING_CI_SETTINGS` = CI-only keys `CI_WIF_PROVIDER` and the required explicit `CI_IDENTITY_PINS`). Inputs never carry a project, state backend, credential scope or path. The project setting "minimum role to use pipeline variables" must be `no_one_allowed` (verified/set at L-G1) so no variable can override a profile. Committed registry `infra/staging/ci-environments.txt` (`staging`, `preview-1`) must equal the input options (B-CI-CONFIG asserts it); the adapter re-checks selector == registry entry == `CI_ENVIRONMENT_NAME` == profile `STAGING_ENV` exactly (so `preview-1` ≠ `preview-10`, `preview-01`, `Preview-1`). |
| S5 | **Candidate policy.** Staging: empty or `staging` = the pipeline's frozen `CI_COMMIT_SHA`; a full SHA only on `git rev-list --first-parent "$CI_COMMIT_SHA"` (actual integrated staging tips; an intermediate feature commit or unmerged SHA is refused). Preview: required; a full SHA reachable from a same-repository branch (`refs/heads/*` fetched; merge-request refs, which may come from forks, are never fetched or accepted) or a branch name (`git check-ref-format --branch`, must exist) resolved once. The resolved SHA, selection text, branch (if any) and `GITLAB_USER_LOGIN` are recorded and used throughout; later branch moves/deletion do not affect the run (the build job hands the deploy job a `git bundle` of the pinned commit). |
| S6 | **Control vs candidate.** All credential-bearing and operator-side code (`.gitlab-ci.yml`, launcher, adapter, `staging.sh`/`lib.sh`, Terraform) comes from the pipeline's protected `staging` revision (`ci_config_ref_uri` pins this in the provider condition). **Compatibility contract:** the candidate's deployment-infrastructure paths `infra/compose`, `infra/postgres`, `infra/minio`, `infra/edge`, `infra/staging`, `infra/docker/edge.Dockerfile`, `infra/docker/minio.Dockerfile` must be byte-identical (`git diff --quiet <control> <candidate> -- <paths>`) to control; otherwise refuse before any credential or cloud call ("incompatible candidate: deployment infrastructure differs from trusted control: <paths>; integrate it to staging first"). The VM release G070 archives from the candidate (`staging.sh:271`) and the edge/minio images are then control-equivalent, so candidate code cannot replace trusted orchestration (remote scripts, Compose networks, metadata guard, Caddy policy). Application code (`core`, `web`, `intelligence`, migrations) is the only candidate-controlled content. |
| S7 | **Schema lineage gate** (adapter, before mutation): candidate migration files (`git ls-tree <sha> core/migrations`, blob IDs) versus the deployed lineage recorded in durable metadata: every version up to the deployed maximum must have the same blob, and the candidate maximum must be ≥ the deployed maximum. Otherwise refuse ("retained schema incompatible: … an explicit, separately authorized `staging.sh reset --confirm <env>` is required; not available in CI"). Closes the gap that `migrate_current` (remote/lib.sh 396–403) has no "database ahead" refusal, without editing VM scripts. |
| S8 | **Credential-free build job.** `staging:build` has **no `id_tokens`, no `environment`, no protected variables**, `resource_group: staging-image-build` (repository-wide single build slot, both environments). It runs only control code: candidate fetch/pin, S6 gate, then control `staging.sh build --source <candidate worktree>` against a synthetic placeholder config (generic values; images tagged under a placeholder registry), under `env -i` (only `PATH`, `HOME`, `DOCKER_*`); outputs `docker save | gzip` of the seven images, the G070-format build manifest (image IDs, Dockerfile and input-file hashes), `candidate.tsv` and the bundle as artifacts (`expire_in: 1 day`, `access: none` if supported). Skipped when `images=published`. Candidate code runs only inside DinD build containers, which receive no job environment, no bind mount and no credential (none exists in the job). |
| S9 | **Deploy job** `staging:deploy`: `needs: staging:build` (optional when skipped), `environment: {name: <env>, action: start}`, `resource_group: staging-env-<env>`, `id_tokens: STAGING_CI_ID_TOKEN: aud: staging-ci:<env>:start`. **Stop job** `staging:stop`: `environment: {name: <env>, action: stop}`, same resource group, `aud: staging-ci:<env>:stop`. Both through the launcher (§2.2) on the pinned docker CLI image with the pinned TLS DinD service and runner tag of the existing jobs. Timeouts: build 45m (40m+4m), deploy 60m (55m+4m), stop 30m (25m+4m). |
| S10 | **Identities** (per environment; §2.4): start/deploy impersonates the existing G070 `<env>-deployer` (unchanged grants); stop impersonates a new `<env>-ci-stop` (no Terraform state, Cloudflare, Secret Manager API, registry, create or delete); the Terraform Cloudflare token secret `<env>-TERRAFORM_CLOUDFLARE_API_TOKEN` (G070 r4) is readable only by the federated start principal and fetched only when a stopped VM's session length must change (the only CI path that plans the full root). The deployer keeps **no** Secret Manager role (G070 README; G057 §18.1). |
| S11 | **Durable metadata** under the environment's existing state bucket, prefix `deploy-meta/<env>/` (never under the Terraform prefix `staging/<env>`; Terraform never reads it): `deploy/<sha>.tsv` (unchanged G070 manifest format), `current.tsv` (candidate, manifest sha256 and object generation, schema-lineage object, last status-log object, pipeline/job/actor, session length, scheduled stop, `started_at`/`stopped_at`), `schema/<sha>.tsv`, `logs/`, `receipts/` (G070 receipts), `runs/<pipeline>-<job>.tsv` (create-only run records), `stop-requests/<pipeline-id>` (create-only), `lease`. Downloads are verified (sha256, candidate field = filename, AR repository = profile, digest shape) before use; a mismatch HOLDs. |
| S12 | **Serialization:** GitLab `resource_group` per environment for deploy and stop (process mode `oldest_first`, set at L-G2); a GCS **lease** object (create with `if-generation-match=0`, holder `ci:<pipeline>/<job>` or `local:<name>`, expiry) excludes concurrent local G070 mutation; a stale CI lease (expired; CI jobs are already serialized) may be broken and is logged; a local lease is never broken by CI. **Pending-start rule:** the stop job writes `stop-requests/<CI_PIPELINE_ID>` before stopping; the deploy job, after acquiring group and lease, refuses (exit 2, no mutation) when any recorded stop pipeline ID is greater than its own, so an earlier queued start can never restart the environment after a later stop. |
| S13 | **Session/duration:** default 7200 s from the profile; `duration` passes through to G070 `start --duration` (600–28800, G070 parse). RUNNING VM: no `start` is issued (deploy never touches the timer); a `duration` that differs from the running session is refused before mutation (G070 semantics). TERMINATED VM: `staging.sh start [--duration D]`. Before deploy the adapter requires ≥ 300 s until the scheduled stop (as the L-5 rerun's 200 s guard, with margin), else refuses. Native `max_run_duration`/STOP remains the only timer. |
| S14 | **Exit classes** (job exit = adapter exit; every non-zero fails the job): 0 observed success; 2 refused before any change (input, gate, compat, lineage, supersession, lease, remaining time); 3 HOLD (from G070); 4 failed after a change; **5 session expired during the action** (G070 returned 255 or another non-zero and the instance is TERMINATED after its armed deadline — the L-5 first-run case; never success); 6 privacy canary hit (outputs withheld); 16 not an approved CI context (launcher; mirrors hosted-context). |
| S15 | **Logs/artifacts:** all `staging.sh` output (it prints saved Terraform plans and VM logs to stdout) goes to a private mode-0600 log, canary-scanned, then uploaded to `deploy-meta/<env>/logs/`. The job log and the only deploy/stop artifact (`staging-summary/summary.txt`, developer access, 90 days) carry a fixed-field summary: action, environment, candidate SHA and selection, actor, digests, session length, scheduled stop, readiness, receipt ID and sha256, exit class. No token, state, plan, env dump or key in Git, logs or artifacts. |
| S16 | **Identity pins (G057 hazard rule):** every CI deploy passes `--pin-secret IDENTITY_WORKOS_<NAME>=<n\|none>` for all three from `CI_IDENTITY_PINS`; a missing pin or `latest` refuses (exit 2). Other secrets keep G070's default. Preview-1 pins `none` (authentication unavailable, reported as such; no staging provider secret reused). |
| S17 | **Retention boundary (Spec §9):** G072 records `started_at`/`stopped_at` per environment (start suspends the stopped clock) but implements **no** scheduler, expiry controller, retire/reset job or automatic retirement, and claims none. |

### 1.2 Unproven assumptions (proof, fallback)

| # | Assumption | Proof | Fallback |
|---|---|---|---|
| A1 | GitLab root-config `spec:inputs` with `options`/`regex`, defaults for MR/schedule/push, and `$[[ inputs.* ]]` interpolation in `rules:if` (literal comparison form), `environment:name`, `resource_group` and `id_tokens:aud` | offline yq/B-CI-CONFIG structure; the G072 MR pipeline proves the file parses with defaults; F-1 proves a deployment-mode web pipeline | web pipeline variables (`STAGING_*`) with pipeline-variable role `maintainer`, adapter allowlist validation unchanged; residual (maintainers could override profile variables) recorded |
| A2 | ID tokens carry `project_id`, `namespace_id`, `ref`, `ref_type`, `ref_protected`, `ci_config_ref_uri`, `pipeline_source`, `environment`, `environment_protected`, `runner_environment`, `aud` as documented | F-1 records claim names and non-personal values (signature never printed) | drop an unavailable clause only by Codex decision; `aud` (always present) carries the action split, so `environment_action` is optional hardening |
| A3 | gcloud and the containerized Terraform accept an `external_account` file (no `service_account_impersonation_url`) and perform the single existing impersonation hop (`GFLAGS`, provider `impersonate_service_account`, `GOOGLE_IMPERSONATE_SERVICE_ACCOUNT`) | F-1 (stop path: gcloud) and L-H1 (deploy path: Terraform `show`/`refresh` through `tf()`) | put the SA hop into the credential file and pass `GCLOUD_IMPERSONATE_SA=` empty (needs the same small lib.sh key, D8) |
| A4 | Direct federated principal access to one Secret Manager secret | L-H1 duration change on preview-1 | dedicated `<env>-ci-token` SA with accessor on that one secret, impersonated only by the start principal |
| A5 | Seven-image `docker save | gzip` fits GitLab's artifact limit and transfers in minutes (G070 local images ≈1.3 GB uncompressed) | EP-1 measures the real compressed size offline | two-job variant with a publish-only `<env>-ci-publish` identity pushing after the build (Plan delta, D2) |
| A6 | gcloud compute ssh uses a pre-registered key (`gcloud compute os-login ssh-keys add --ttl`) without re-adding it untimed | F-1/L-H1 OS Login profile key count before/after | lib.sh optional `STAGING_SSH_KEY_EXPIRE_AFTER` → `--ssh-key-expire-after` on `vm_ssh`/`wait_vm_ready` (in W2's allow-list, used only if needed) |
| A7 | google-cloud-cli slim image (digest-pinned) has bash, git, tar, coreutils, openssh-client | W3 inspects at implementation (coordinator pulls once, cache-prep lock) | `buildpack-deps` base (as ci-context) + digest-pinned gcloud tarball |
| A8 | GCS conditional bindings by object-name prefix for the stop identity (get/create, no list) | L-W1p read-back + F-1 | stop identity reads `current.tsv` through the deploy job's published summary only (no metadata read); receipts kept locally in the job artifact |

### 1.3 Early proof (attempt 1, before broad integration; riskiest first)

- **EP-1 — credential-free build boundary (offline, coordinator, build slot; ~45–75 min elapsed after W2's build phase and W4's fixture land).**
  `bash infra/staging/tests/ci-build-boundary.sh --work <T>/boundary --report <E>/boundary-1`: (a) a synthetic hostile
  candidate repository (Dockerfile `RUN` steps that dump `env`, `/proc/*/environ`, `/builds`, `/certs`; a variant whose
  `infra/compose/local.yml` adds `build.args: ${CANARY}` / `additional_contexts`) is driven through the real build phase
  with canary values in `CI_JOB_TOKEN`, a fake `STAGING_CI_ID_TOKEN` and `CLOUDFLARE_API_TOKEN` in the job environment:
  the compose variant is refused by the S6 gate before any build; the Dockerfile variant builds and **no canary appears**
  in `docker export` of every image, the build log or the artifacts; (b) one real build of the baseline staging tip
  through the same path records the compressed artifact size (A5) and image-ID preservation across save/load/retag.
  Stop criterion: any canary hit or a size above the documented artifact limit → stop and report (D2 fallback).
- **EP-2 — first live federation (first released live action, ~30 min, no infrastructure mutation).** After merge,
  preview-1 WIF bootstrap and the preview-1 profile: run `stop` for `preview-1` (F-1). Proves A1–A3/A6/A8 on the
  real provider: token claims, condition, single-hop impersonation as `preview-1-ci-stop`, read-only compute calls,
  metadata writes; the denial probes of §2.4 run in the same job. If preview-1 is not provisioned yet, the same job
  ends at G070's `ABSENT` refusal (exit 2) after federating — still a pure federation proof.

## 2. Design

### 2.1 Routing matrix (B-CI-CONFIG pins the rules that implement it)

| Context | Jobs |
|---|---|
| push (any branch), tag | none (workflow, unchanged) |
| MR (feature, fork) and RC-promotion MR | validation exactly as today; no `staging:*`; no `id_tokens` anywhere in those jobs |
| schedule | validation as today; `action != validate` → workflow never |
| web/api on `staging` (protected), `action=validate` | validation as today |
| web/api on `staging`, `start-deploy`, environment ∈ registry | `staging:build` → `staging:deploy` only |
| web/api on `staging`, `stop`, environment ∈ registry | `staging:stop` only |
| web/api on `staging`, non-default action, `environment=none` | zero jobs → refused at creation |
| web/api on another branch (incl. `identity-qualification`, `main`), non-default action | workflow never |
| `identity-qualification` web/schedule, default action | `identity-staging` as today |

### 2.2 Launcher and context (`infra/ci/staging-action.sh`, POSIX sh, job script)

Pattern of `hosted-context.sh` (file not edited; its project-ID/path pins 75–76 are **read** as the single source):
G1-style context check (GitLab job, TLS DinD vars), approved-context gate (pinned project ID and path SHA, branch
`staging`, ref protected, source web|api, no MR IID, exact job name, and for deploy/stop `CI_ENVIRONMENT_NAME` == the
selected environment), refusing with exit 16 before any token is touched. For deploy/stop it moves the ID token from
the job environment into a mode-0600 file in a private directory `"$NS/.staging-ci-$CI_JOB_ID"` (outside the
project directory, so never an artifact), copies the four profile files there, unsets the variables, builds the
context image (`infra/docker/ci-context.Dockerfile` for build; new `infra/docker/staging-ci.Dockerfile` =
digest-pinned google-cloud-cli + the pinned docker CLI stage + the pinned jq 1.7.1 binary for deploy/stop) and runs
`bash infra/staging/ci.sh <action>` in it with `-v "$NS:$NS"`, host network to the job daemon and an explicit `-e`
allow-list. `after_script`: best-effort lease release, OS Login key removal, private-directory removal, summary.

**Launcher ↔ adapter contract (settled so W2 and W3 work in parallel):** env `STAGING_CI_ACTION`
(build|deploy|stop), `STAGING_CI_ENV`, `STAGING_CI_CANDIDATE_INPUT`, `STAGING_CI_DURATION_INPUT`,
`STAGING_CI_IMAGES`, `STAGING_CI_PRIVATE`, `STAGING_CI_ID_TOKEN_FILE` (deploy/stop), `STAGING_CI_PROFILE_DIR`
(deploy/stop), `STAGING_CI_OUT` (artifact directory), the public CI metadata the hosted context already forwards,
plus `CI_ENVIRONMENT_NAME`, `CI_JOB_NAME`, `GITLAB_USER_LOGIN`; exit classes S14.

### 2.3 Adapter (`infra/staging/ci.sh` + `infra/staging/ci-lib.sh`, bash)

- **build:** fetch `refs/heads/*` (runner credential helper; no tokens written), resolve/pin (S5), S6 gate, lineage
  file, bundle, synthetic config, `staging.sh build --source <worktree> --config <synthetic>` under `env -i`, save
  and checksum artifacts. No `gcloud`, no network beyond Git and base-image pulls inside DinD.
- **deploy:** profile load (G070 `load_conf` rules plus CI keys), registry/selector/profile identity checks; credential
  files (§2.5); `gcloud auth login --cred-file`; OS Login key with TTL; lease; supersession check; pins check;
  re-verify the build artifacts (candidate.tsv vs bundle commit vs policy for the recorded selection; S6 re-run;
  lineage re-computed, never trusted from the artifact); metadata pull + verify; lineage gate (S7); images: a valid
  published `deploy/<sha>.tsv` → reuse (no load/push; AC4), else `docker load`, verify IDs against the build manifest,
  ID-preserving retag to the profile repository (the G070 R-1/R-1b procedure), `staging.sh push` (re-verifies IDs,
  records digests), upload manifest; VM state read; TERMINATED → token only if the session length must change (§S10)
  → `staging.sh start [--duration D]`; RUNNING → duration check; remaining-time guard; `staging.sh deploy --candidate
  <sha> --pin-secret …`; success only when the deploy's readiness lines report READY and running digests equal the
  manifest; scheduled stop read with `gcloud … describe`; `staging.sh receipt --candidate <sha>`; metadata push
  (`current.tsv`, schema, logs, receipt, run record); lease release; summary.
- **stop:** profile/identity checks; `GCLOUD_IMPERSONATE_SA=<env>-ci-stop@<project>` (D8); credentials; write
  `stop-requests/<pipeline-id>`; lease (refuse while a live local lease exists); `staging.sh stop` (G070 graceful
  VM-side stop, then provider stop, waits for observed `TERMINATED`; already stopped = idempotent success after the
  describe confirms it); receipt (Terraform-derived rows are unavailable to the stop identity by design and are cited
  from the last deploy receipt); metadata push; summary. Needs no candidate, build, artifact or earlier workspace.
- **local subcommands** (operator ADC, same validation): `lease acquire|release|show` and `meta pull|push` for the
  documented CI↔local handoff (§2.7).
- Exit classification and canary scan (patterns: the ID token, any fetched Cloudflare token, `ya29.`, JWT prefix,
  `BEGIN … PRIVATE KEY`, `cfut_`/`cfat_` shapes) on the private log, summary and artifacts before publication.

### 2.4 Identity and IAM (`infra/terraform/bootstrap/ci-federation.sh`, owner-run, idempotent, `--dry-run`, `--teardown`)

Per environment **project** (D1 option A, recommended): workload identity pool and OIDC provider (issuer the GitLab
instance; allowed audiences exactly `staging-ci:<env>:start` and `staging-ci:<env>:stop`); `sts.googleapis.com`
enabled; mapping `google.subject=assertion.sub`, `attribute.action` = `start`/`stop` from the `aud` suffix (else
`none`). **Attribute condition:** `project_id` and `namespace_id` equal the pinned values, `ref_type=="branch"`,
`ref=="staging"`, `ref_protected=="true"`, `ci_config_ref_uri` equals the project's `//.gitlab-ci.yml@refs/heads/staging`,
`pipeline_source in ["web","api"]`, `environment=="<env>"`, `environment_protected=="true"`,
`runner_environment=="gitlab-hosted"`, `aud` in the two audiences.

| Principal | Grants | Scope |
|---|---|---|
| principalSet `attribute.action/start` | `roles/iam.workloadIdentityUser` | `<env>-deployer` (G070 SA; its existing grants unchanged) |
| principalSet `attribute.action/start` | `roles/secretmanager.secretAccessor` | the one secret `<env>-TERRAFORM_CLOUDFLARE_API_TOKEN` |
| principalSet `attribute.action/stop` | `roles/iam.workloadIdentityUser` | new `<env>-ci-stop` |
| `<env>-ci-stop` | custom role `ciStop<env>`: `compute.instances.get/list/stop`, `compute.disks.list`, `compute.addresses.list`, `compute.zoneOperations.get`, `compute.projects.get`; `roles/compute.osAdminLogin` and `roles/iap.tunnelResourceAccessor` on the instance; `roles/iam.serviceAccountUser` on `<env>-vm` | environment project / its VM |
| `<env>-ci-stop` | conditional `objectViewer` and `objectCreator` limited to `objects/deploy-meta/<env>/` | state bucket |
| never | Owner/Editor, project IAM admin, service-account keys (no `keys create` anywhere), any Secret Manager role on either SA, any Terraform-state access for stop | — |

Honest boundary: `osAdminLogin` (needed for G070's graceful `sudo` stop) makes the stop identity root on its own VM,
hence able to read that environment's VM-delivered secrets; it still cannot plan/apply, read state, change DNS,
touch the registry, create/delete resources or reach another environment. **Accounting for G070's broad grants:**
the deployer's project-wide compute/network/security roles and bucket-wide `objectAdmin` are confined to one
environment because each environment has its own project (A) and therefore its own state bucket; no preview identity
holds any role in the staging project. Option B (shared project) would need conditional, exact-name compute bindings,
a list-only custom role, a per-environment state bucket or prefix-conditioned binding and a `bootstrap.sh` change
(+3–5 h coding, live IAM-condition proof) — fallback only, by D1. **Cloudflare:** both the Terraform token
(operator/CI-side) and Caddy's `CF_API_TOKEN` (VM, edge container only) are zone-scoped; a shared zone gives no
per-record isolation. Protection is procedural plus exact checks: `label_check` (lib.sh 504–515) refuses plan changes
to records outside `B`/`*.B`, Caddy issues only for its `EDGE_BASE_DOMAIN`, the S6 gate keeps edge/Compose control-equal,
and no candidate build or application container receives either token. Recommend separate tokens per environment for
independent revocation (owner choice). **Denial probes** (non-mutating, every deploy/stop job; cross-environment
targets listed by the owner in `CI_DENIAL_PROBES` during proof windows): the stop principal's impersonation of the
deployer must be denied; a listed foreign VM describe / state object read / secret access must be denied; any success
fails the job (exit 6-class isolation finding).

### 2.5 Runner-local federation files

Under `STAGING_CI_PRIVATE` (`$NS/.staging-ci-<job>/`, mode 0700, on the shared builds volume the job daemon sees):
`id-token` (0600); `gcloud-adc.json` (`external_account`, `credential_source.file` = host path); `private/` = G070
`PRIVATE_DIR` holding the profile `backend.hcl`/`terraform.tfvars`, plans, `tf-data` and `ci/tf-adc.json` whose
`credential_source.file` is `/private/ci/id-token` (the token copy inside `PRIVATE_DIR`, which `tf()` already mounts at
`/private`); `GOOGLE_ADC_FILE` points at that file, so **`tf()` needs no change**; `evidence/` = G070 `EVIDENCE_DIR`;
`home/` = `HOME` (gcloud config, ephemeral SSH key). `CLOUDFLARE_API_TOKEN` exists only in the environment of the one
`staging.sh start` process that needs it, then is unset. Removal on exit/cancel (trap plus `after_script`), verified.

### 2.6 Rejected alternatives

| Alternative | Reason rejected |
|---|---|
| `when: manual` jobs in every web pipeline | R1 conflict: optional manual jobs hide failures (`allow_failure`), required ones block; selector would need job variables |
| GitLab `on_stop` paired stop job | ties stop to an old pipeline/workspace/artifacts (R5) |
| Single build+deploy job | holds the environment lock through the build (R8); OIDC token present in the job that builds candidate code (R10) |
| Build job with a publish identity | OIDC credential in the build job; kept only as the A5 fallback |
| Images via GitLab container registry | new publishing infrastructure (out of scope) |
| Running the candidate's `staging.sh`/remote scripts, or overlaying control infra on a different candidate | untrusted control code / silent incompatibility; S6 refuses instead |
| Terraform-managed WIF | G070 keeps bootstrap identities out of Terraform state |
| Service-account key in a CI variable | forbidden (R6) |
| GitLab artifacts as deployment metadata | expiring, not a durable sole copy (R7) |
| New metadata bucket | extra bootstrap resource; a separate prefix of the per-environment bucket suffices |
| Cloud Scheduler / cron / expiry controller | forbidden (R4, §9) |
| Cloudflare sub-zone per preview | plan-dependent (sub-domain zones need Cloudflare Enterprise); not assumed |

### 2.7 Local-operator handoff

CI→local: `ci.sh meta pull` into the operator's private `EVIDENCE_DIR`, then `ci.sh lease acquire --holder
local:<name>`, then G070 local commands, then `meta push` and `lease release`. CI refuses while a live local lease
exists; the local G070 path itself is not modified, so a local operator who skips the lease is not technically
blocked (residual; D4 offers enforcement in `staging.sh`). Interrupted CI runs: the next mutation runs G070
`reconcile`/`vm_status` and the remote recovery; cancellation never resets data.

## 3. Bounded G070 adapter changes (all additive; defaults unchanged; D8 asks Codex to confirm none is a contract change)

1. `staging.sh build --source DIR`: build from a clean detached checkout at the candidate (compose `-f` files, build
   contexts and edge context under `DIR`; git checks against `DIR`); default `REPO` (current behaviour). Needed because
   `cmd_build` assumes control == candidate (staging.sh 120, 133–138).
2. `lib.sh` optional `staging.conf` key `GCLOUD_IMPERSONATE_SA` (validated as exactly `<env>-deployer@…` or
   `<env>-ci-stop@<project>.iam.gserviceaccount.com`) used in `GFLAGS` (lib.sh 107); absent = `DEPLOYER_SA`.
3. Only if A6 fails: `lib.sh` optional `STAGING_SSH_KEY_EXPIRE_AFTER` (`^[0-9]+[mhd]$`) added to `vm_ssh`/`wait_vm_ready`.
No change to Terraform, remote scripts, manifests, receipts, exit codes or the local operator workflow.

Consumed unchanged: `start --duration` (A2 `set-scheduling` path), `deploy --candidate --pin-secret`, `stop`,
`receipt`, `push`, exact-name helpers, `label_check`, `reconcile`, HOLD semantics, retained-baseline receipts.

## 4. Files, writers, schedule

Four implementation workers (Opus 5.5 high, isolated worktrees, stack-free: no Docker/Terraform/cloud; pure-bash
fixture self-runs and `bash -n` only). One coordinator integrates and runs every Docker batch. Fewer than five because
the remaining work shares `run.sh`, `check.sh` and `staging.sh` and must stay single-writer.

| Worker | Exclusive writable paths | Depends on |
|---|---|---|
| **W1 identity/bootstrap** | NEW `infra/terraform/bootstrap/ci-federation.sh`; NEW `infra/staging/tests/ci-federation.sh` (fake-gcloud dry-run fixtures); MOD `infra/terraform/README.md` (CI federation section, IAM rows) | Plan §2.4 |
| **W2 lifecycle adapter** | NEW `infra/staging/ci.sh`, `infra/staging/ci-lib.sh`, `infra/staging/ci-environments.txt`, `infra/staging/examples/ci-settings.example`; MOD `infra/staging/staging.sh`, `infra/staging/lib.sh` (§3 only); NEW `infra/staging/tests/ci-adapter.sh` | §2.2 contract |
| **W3 pipeline/launcher** | MOD `.gitlab-ci.yml`; NEW `infra/ci/staging-action.sh`, `infra/docker/staging-ci.Dockerfile`; MOD `infra/validation/run.sh` (B-CI-CONFIG two-document parse, new rules/jobs, registry equality; B-SHELL list; B-HADOLINT loop); NEW `infra/staging/tests/ci-launcher.sh` | §2.1, §2.2 contract |
| **W4 docs/proof** | Phase 1: MOD `docs/foundation/STAGING.md` (CI section: setup, inputs, Play/Stop, SHA/deadline display, failed-job recovery, handoff, rotation, costs, branch/schema compatibility, provider-origin readiness, slot registration, disabling/retirement), `docs/foundation/OPERATIONS.md` (CI paragraph: deploy jobs now exist); NEW `infra/staging/tests/ci-build-boundary.sh` + hostile fixture tree. Phase 2 (after W1–W3 hand back): MOD `infra/staging/check.sh` (rows `CI-FEDERATION`, `CI-ADAPTER`, `CI-LAUNCHER`; LIFECYCLE-DRYRUN cases for §3; STAGING-GREP: `ci.sh` never lets `tf show`/VM logs reach stdout, no scheduler text; SHELLCHECK-DIRECT list) | W1–W3 test scripts |
| Coordinator | Result; proposed `management/v0/results/G072-scope-allowlist.txt` (Codex imports); integration, batches, MR | all |

Sequence: P0 dispatch W1–W4 together → W2 delivers the build phase first (EP-1 input) → coordinator runs EP-1 →
W4 phase 2 → integrate → S1–S4 → Codex source review → MR (hosted fast + Duo; coordinator monitors) → merge →
post-merge affected rows → live items (§5.3) under their releases. Shared-file risks: `staging.sh`/`lib.sh` (G070's
remaining work is Result/closeout only; verify no pending G070 source change at dispatch), `check.sh` and `run.sh`
(single writers W4/W3), `.gitlab-ci.yml` (W3 only).

## 5. Validation and acceptance mapping

### 5.1 Source checks (offline, coordinator; no cloud/DNS contact; no full ci-controls or application suite)

- **S1** `bash infra/validation/run.sh --base <origin/staging SHA> --allow-file management/v0/results/G072-scope-allowlist.txt --report-dir <E>/static-N lint lint-base provenance` (B-CI-CONFIG extended, B-SHELL, B-HADOLINT, B-LITERALS, B-SECRET-SOURCE, B-PROTECTED, B-SCOPE, B-SECRET-CANARY, aggregate).
- **S2** `bash infra/staging/check.sh run --work <T>/check --report <E>/offline-N --edge-image <retained edge ref> --only CI-FEDERATION,CI-ADAPTER,CI-LAUNCHER,LIFECYCLE-DRYRUN,STAGING-GREP,SHELLCHECK-DIRECT` (`prepare` once if pins changed).
- **S3** EP-1 `bash infra/staging/tests/ci-build-boundary.sh --work <T>/boundary --report <E>/boundary-N` (build slot).
- **S4** `bash infra/ci/validate-controls.sh --only F-push-main,F-push-staging,F-agent-push-staging,F-agent-push-main,RR1,RR2` — only because these controls parse `.gitlab-ci.yml` text (5932–5940, 2172/2254); IDs confirmed with `--list` first.
- **Hosted:** MR `validate:checks` (fast) on every push, monitored by the coordinator; `validate:integration` manual, not run.
- **Post-merge** on the frozen merged SHA: S2 rows; S1 only if the merged tree differs from the last S1 tree.
Docker resources: task labels `g072.task=G072`; build slot for S3; no DB/browser slot needed.

### 5.2 AC mapping (reuse / extend / new, with the uncovered boundary for each new case)

| AC | Source evidence | Live evidence |
|---|---|---|
| AC1 routing | **extend** B-CI-CONFIG: exact inputs spec, workflow rules (deployment refusals), guard rules on the three validation jobs, the three new jobs' exact keys/rules/environment/resource_group/id_tokens/images/timeouts/artifacts, no `id_tokens`/`environment` on validation or build jobs, registry = input options; **reuse** S4 controls; **new** CI-LAUNCHER cases (gate: MR IID, unprotected, other branch, push/schedule/trigger source, wrong job, env mismatch → exit 16, no token file) — new boundary: the launcher | F-1 pipeline; one refused creation (non-default action on a non-staging branch) |
| AC2 candidate fidelity | **new** CI-ADAPTER: default pipeline SHA; first-parent SHA accepted for staging; intermediate/unmerged SHA refused for staging; same-repo branch resolved once and pinned; moved/deleted branch uses the bundle; MR-ref/fork-only SHA refused; malformed inputs refused with zero fake-gcloud calls; manifest reuse vs mismatch (candidate, repository, digest shape, image ID) refused before push; S6 gate (infra change refused, app-only accepted); S7 lineage (older, divergent, equal, newer); **extend** LIFECYCLE-DRYRUN for `build --source`; S3 image-ID preservation | L-H1 digests = manifest of the pinned SHA; second run with `images=published` reuses without push |
| AC3 identity/secrets | **new** CI-FEDERATION: dry-run renders the exact pool/provider/audiences/mapping/condition and bindings; one mutation per clause fails the row; no forbidden role or key verb; teardown removes exactly the created set; **new** CI-ADAPTER: credential-file shapes and paths, token only for a session change, canary scan blocks publication (exit 6), `GCLOUD_IMPERSONATE_SA` validation; S3 canaries | F-1 federates; denial probes denied; L-W read-backs (sanitized policies) |
| AC4 start/deploy | **new** CI-ADAPTER: RUNNING/TERMINATED paths, duration refusal on a running session, remaining-time guard, pins required (`latest` refused), readiness parsing; **reuse** G070 L-5 rerun expiry/retained-data evidence (no two-hour wait) | L-H1 (preview-1); L-H2 (staging, gated) |
| AC5 stop/recovery | **new** CI-ADAPTER: stop with no candidate/artifacts; stop after a failed/canceled start; stale CI lease broken, local lease refused; supersession refusal; missing build artifact fails deploy only; rc 255 + TERMINATED-after-deadline → exit 5 | F-1 (idempotent stop), L-H1 stop, L-H2 stop-preview-leaves-staging |
| AC6 closeout | Result AC map, preserved failures, cleanup receipt | hosted fast, Duo, merged-SHA rows |
| R12 | CI-ADAPTER/CI-FEDERATION: allowlist/profile substitution refused; `preview-1` vs `preview-10` (selector, resource group, attribute, profile, names — the latter reusing the G070 fixture); **extend** LIFECYCLE-DRYRUN: `label_check` refuses a preview-1 plan touching staging or preview-10 records; stop targets only its environment's names | both serve at once; preview reports its SHA/digests; stopping preview leaves staging serving; independent deadlines |

### 5.3 Separately released live items (none released by this Plan)

| ID | Action | Performer | Precondition | Time / cost | Rollback | Release needed |
|---|---|---|---|---|---|---|
| L-G0 | Read-only GitLab/GCP readiness: tier/features, pipeline-variable setting, `staging` protection, project path/ID equality with the pins | coordinator | merge | 15 min / $0 | — | Codex |
| L-G1 | GitLab settings: protected environments `staging`, `preview-1` (deploy: maintainers/owner); 4 env-scoped protected file variables per environment; pipeline variables `no_one_allowed` | owner (or coordinator with in-session owner approval) | L-G0 | 30–45 min / $0 | delete settings | owner + Codex |
| L-P0 | Preview GCP project (D1=A), billing link, monthly alert budget | owner | D1 | 30–60 min / $0 | delete project | owner |
| L-P1 | Preview-1 G070 bootstrap, owner-entered Caddy and Terraform tokens, first `build/push/provision` of the staging tip through the G070 local path (one-time, outside routine CI), production certificate for `preview-1` B and `*.B`, then stop | owner (bootstrap, secrets); coordinator (G070 commands) | L-P0, G070-style live proposal | 1.5–2.5 h / running ≈ $0.072/h, stopped ≈ $11–12/month | G070 `retire --yes` + bootstrap teardown (own release) | owner live release + Codex |
| L-W1p | `ci-federation.sh` for preview-1 (dry-run reviewed first), read-back | owner | L-P1 (or L-P0 for the federation-only EP-2) | 20–30 min / $0 | `--teardown` | owner + Codex |
| F-1 (EP-2) | `stop` pipeline for preview-1 (idempotent) with denial probes | coordinator via web/API pipeline | L-G1, L-W1p | 15 min / CI ≈ 16 units | none (metadata retained) | Codex |
| L-G2 | Resource-group process mode `oldest_first` for `staging-env-*`, `staging-image-build` | owner/coordinator | groups exist (after F-1) | 10 min | revert mode | Codex |
| L-H1 | Preview-1: `start-deploy` of a same-repo unmerged branch SHA with a custom duration, observe; `images=published` reuse run; `stop`; optional live supersession check | coordinator | F-1 | 60–75 min / VM ≈ $0.10, CI ≈ 150 units | `stop` pipeline; G070 local stop | owner + Codex |
| L-W1s, F-3 | `ci-federation.sh` for staging; `stop` of the stopped staging baseline (idempotent) | owner; coordinator | **G057 Part B retention hold released, or an explicit owner release coordinated with G057**; staging identity pins set in the profile | 30–45 min / $0 | `--teardown` | owner + Codex |
| L-H2 | Two environments at once: staging `start-deploy` (pins from profile) and preview-1 `start-deploy`; both serve; stop preview, staging still serves; stop staging | coordinator | L-W1s/F-3, same G057 gate | 60 min / ≈ $0.15 | `stop` pipelines | owner + Codex |

No retire, reset or reprovision of the shared `staging` baseline is part of G072. Preview-1's retention after the
proof (48 stopped hours default, owner extension, or a separately released G070 retirement) is an owner decision;
automatic retirement does not exist.

## 6. Estimates (unmeasured; ranges)

| Category | Estimate |
|---|---|
| Coding (agent) | W1 3–5 h, W2 7–10 h, W3 4–6 h, W4 3–5 h ≈ **17–26 agent-hours**; wall clock ≈ **8–11 h** with four parallel workers, plus coordinator integration 2–4 h |
| Build/test elapsed (coordinator) | S1 ≈ 6 min/run, S2 ≈ 3–5 min/run, S4 ≈ 5–10 min, S3/EP-1 ≈ 45–75 min (one real build); with two or three correction reruns ≈ **2–3.5 h** |
| Review/queue waits | Plan review, 1–3 source-review rounds and Duo: **unknown** (G070 needed several rounds); hosted fast ≈ 13–17 min per push |
| Live/setup | preview path (L-G0…L-H1) ≈ **4–6 h elapsed** including owner steps; staging path (L-W1s, F-3, L-H2) ≈ **1.5–2 h**, date **unknown** (gated on G057, not before Thursday) |
| Total | ≈ **3–5 working days** elapsed, dominated by reviews and owner/live windows; the old 3–6 h estimate is superseded |

## 7. Resources, costs and cleanup

Run inventory (existing coordinator inventory plus `runs/` records): pipeline/job IDs, action, environment, actor,
candidate SHA and selection, manifest sha256, digests, requested/actual session length, scheduled stop, outcome
class, receipt ID/sha256, profile file digests, lease generation, CI units, cost window. Retained items with owner
and retirement: WIF pool/provider, `<env>-ci-stop`, custom role, bindings (owner: environment owner; retire with
`ci-federation.sh --teardown`); GitLab protected environments, variables, resource groups, input option (project
owner; remove with the environment); `deploy-meta/<env>/` (environment owner; delete after retirement, noncurrent
versions expire under the bucket's 90-day rule); preview-1 project resources (environment owner; G070 retire plus
bootstrap teardown). Costs per environment (G070 proposal r2 figures): running ≈ $0.072/h; stopped ≈ $11–12/month
(two disks ≈ $7, reserved IPv4 ≈ $3.65, secrets/registry small); the two-hour limit does not remove disk/IP cost.
Local: worker worktrees removed at each handback; S3 images (`g072-*` tags, labels) removed and verified after each
run; private paths `.local/{evidence,secrets,tmp}/g072-a1` (G070 precedent) cleaned at acceptance; no leftover
containers, volumes, networks or worktrees.

## 8. Risks, stop conditions and decisions

Stop and escalate on: an untrusted ref or context obtaining a token or impersonation; any Google key; a secret,
token, state, plan or env dump in a log/artifact/Git; a widened production/account grant; any reset/destroy in
start/stop; schema downgrade; unowned cost; stop unusable after a failed/canceled start; conflicting lifecycle owners
(lease); a G070 contract/security change beyond §3; a G057 hold conflict; a native refusal; budget exhaustion.
Main risks: GitLab feature gaps (A1/A2; fallbacks recorded); artifact size (A5); first-start latency on a VM whose
release predates the `--init` fix (mitigated: preview starts from the merged release); human error in profile
variables (adapter identity checks plus provider condition); a local operator bypassing the lease (D4).
Scope: about 2.5–3k lines including fixtures; a split (staging-only first, previews second) is possible but shares
the build boundary and federation work, so it is not recommended.

**Decisions requested.** D1 preview isolation: separate GCP project per preview slot (**recommended**) or shared
project with conditional IAM (+scope). D2 confirm the two-job build/deploy design (single job would hold the lock).
D3 owner sets pipeline variables to `no_one_allowed` and creates protected environments/variables (GitLab settings).
D4 local exclusion: CI-enforced lease plus procedure (planned) or also enforce in `staging.sh` (+1–2 h). D5 staging
candidates: first-parent history (planned) vs any ancestor. D6 accept the S6 rule that refuses older staging
candidates whose deployment infrastructure differs. D7 who updates staging `CI_IDENTITY_PINS` during G057 R-1…X-9.
D8 confirm §3 changes are small adapters, not G070 contract changes. D9 confirm GitLab Premium protected
environments are available. **Blockers:** the G057 retention hold for every staging live item; owner GitLab/GCP
setup for preview-1.

## 9. Schema, contracts and history

No database schema, migration, application, tenant, public API or MCP change. New interfaces: the pipeline inputs,
the launcher/adapter contract (§2.2), the four profile variables, `deploy-meta/<env>/` layout, and the `staging.sh`
additions in §3.

- 2026-10-27 r0 — Plan written against staging `fc2f8450` from Spec r2; planning only; zero attempts spent.

## 10. Revision r1 after Codex review (CORRECTION LOOP) — 2026-10-28 UTC

Author: G072 planner (Claude, Opus 5.5). **Planning only**: nothing in §§0–10 releases code, tests, CI, IAM/WIF,
GitLab settings, DNS or any live action. §§0–9 (r0) are preserved byte-for-byte above. **r1 controls wherever it
differs from r0**; every r0 item not superseded below stays in force. Unchanged: priority high, security/infrastructure
tier, budget **B3** (zero attempts spent), the four-worker target, and the G057 retention hold on every staging live item.

### 10.0 Rebinding

| Input | r1 value |
|---|---|
| Planning baseline | `origin/staging` **`0c66332d8914cf362bbfe3444d3557ee539b6eef`** (merged into this branch). r0's `fc2f84503f21eae2506346df202fdd3c400599f9` and the intermediate `1e29cd4bfefd5b749c2295f8898c8204f2b4c451` are historical bindings. `git diff --name-only fc2f8450 0c66332d` lists management/ paths only (PROGRESS, ROADMAP, SUPERVISOR-HANDOFF, the Codex review, a ONEOFF note), so every §0 source hash and line reference holds unchanged. |
| Review closed by r1 | `management/v0/results/G072-plan-r0-codex-review-20261028.md` (CORRECTION LOOP: 3×P1, 5×P2) of r0 commit `5ff64938…`, Plan sha256 `f9cd8a65…c645` (exact prefix of this file) |
| Spec / Goal | unchanged: Spec r2 `bbcbbc5e…f726`, Goal `d42f40ab…311b` |
| Sources added for r1 | `infra/staging/remote/deploy.sh` sha256 `afe76b84…bb71` (`promote` 34–63 runs before `migrate_current` 106, readiness 127); `remote/start.sh` `bffa03ae…c2e8` (`migrate_current` of the current release 49); `remote/status.sh` `77a92379…73be` (`candidate` line 27); `remote/lib.sh` 381–403 (migrate status/up), 421–455 (superuser `docker exec … psql` read pattern; goose table `app.goose_db_version`); `core/internal/migrate/runner.go` `2a7ba81b…05f9` 204–225 (status lines); `staging.sh` 63–65 (EXIT trap), 648–665 (`main`); `lib.sh` 54–63 (`forbid`: no ` create ` verb), 107 (`GFLAGS`); `bootstrap.sh` 175 (deployer bucket-wide `objectAdmin`), 236–242 (state bucket kept while live objects remain) |

### 10.1 Resolution map

| Codex finding / superseded r0 item | r0 location | r1 resolution |
|---|---|---|
| **1 P1** stop identity cannot overwrite `current.tsv` or release the lease | §2.4 stop rows, S11, S12, A8 | dedicated per-environment metadata bucket; bucket-wide view/create plus delete only on the exact keys `current.tsv` and `lease`; IAM-model fixture; repeated stop proof; A8 fallback deleted (§10.3) |
| **2 P1** stale schema lineage after a partial deploy | S7, §2.3 | fence entry written before upload/promote/migrate; gate against every possible lineage; observed-lineage reconciliation; named clearers (§10.4) |
| **3 P1** advisory local/CI lock | S12, §2.7, D4 | the same lease enforced by G070 `provision`/`deploy`/`start`/`stop`/`reset`/`retire`; recovery and break-glass; negative concurrency and legacy-parity fixtures (§10.5) |
| **4 P2** cloud writes before validation | §2.3 deploy/stop sequence | three phases: local checks, read-only cloud reads and gates, then writes; zero-write call-log negatives (§10.6) |
| **5 P2** raw Terraform plans in durable logs | S15, §2.5 | raw output only in the private job directory; allowlisted, canary-scanned publication (§10.7) |
| **6 P2** A1 pipeline-variable fallback | §1.2 A1 | deleted; unavailable typed inputs or variable restriction = stop and return to Codex (§10.2) |
| **7 P2** published-image mode lacks candidate preparation; extra `images` choice | S1 `images`, S8, S9, §2.3 | `images` input removed; credential-free `staging:prepare` always runs; read-only `staging:check` makes the automatic R3 reuse decision (§10.8) |
| **8 P2** preview candidate default | S5 | empty candidate = frozen staging pipeline SHA for both environments; explicit same-repository preview SHA/branch overrides; stop takes no candidate/duration (§10.2) |
| S1 inputs (`images`), S2 job set, S8 skip rule, S9 `needs` | §1.1 | §10.2, §10.8, §10.10 |
| S11 metadata location, S12 lease, S15 logs | §1.1 | §10.3, §10.5, §10.7 |
| A8 | §1.2 | replaced by A8′ (no fallback), §10.3, §10.16 |
| §1.3 early proof | §1.3 | §10.12 |
| §2.1 routing, §2.2 contract, §2.3 adapter order | §2 | §10.10, §10.6 |
| §2.4 stop-identity storage row; audiences | §2.4 | §10.3, §10.8, §10.10 (compute grants of `<env>-ci-stop` and the provider condition unchanged) |
| §2.5 runner files | §2.5 | adds `raw/` (§10.7) |
| §2.6 rejected "new metadata bucket" | §2.6 | reversed (§10.3); new rejected alternatives in §10.3/§10.8 |
| §2.7 local handoff | §2.7 | §10.5 |
| §3 "all additive … D8 asks … none is a contract change" | §3 | additive G070 interfaces, documented and parity-proven (§10.9) |
| §4 files/writers, §5.1 S2 selector, §5.2 AC rows, §5.3 live items, §6 estimates, §8 risks/decisions | §§4–8 | §10.11, §10.13, §10.14, §10.15, §10.16 |

### 10.2 Settled decisions, operator flow (findings 6 and 8)

- **D1 settled (owner):** separate GCP project per preview slot (r0 option A). Each environment's identities, state
  bucket, metadata bucket, registry and IAM live in its own project; no preview identity holds any role in the staging
  project. L-P0 is that preview project.
- **Settled by Codex:** D2 credential-free build/pin stages stay separate from credential-bearing jobs; D5 first-parent
  staging candidates; D6 the infrastructure-compatibility refusal (S6); D8 additive interfaces (§10.9); D9 verify the
  GitLab tier/features at L-G0 and **stop** if unavailable, never weakening an identity boundary. D4 is closed by
  finding 3 (technical enforcement, §10.5).

Operator flow: a new pipeline on trusted ref `staging` (web UI or API), choosing an action and an allowlisted
environment. Inputs (r0 S1 minus `images`):

| Input | Values | Empty / default behaviour (W4 documents each in STAGING.md) |
|---|---|---|
| `action` | `validate` (default), `start-deploy`, `stop` | `validate` = today's ordinary web validation pipeline; no staging job exists |
| `environment` | `none` (default), `staging`, `preview-1` (= `ci-environments.txt`) | `none` with `start-deploy`/`stop` → zero jobs → GitLab refuses to create the pipeline; nothing runs |
| `candidate` | empty (default), full SHA, `staging`, or (preview only) a same-repository branch name | empty or `staging` → the frozen `CI_COMMIT_SHA` of this staging pipeline, **for both environments**. Staging accepts an explicit SHA only on `--first-parent` staging history. `preview-1` additionally accepts a full SHA reachable from a same-repository `refs/heads/*` branch, or a branch name resolved once in `staging:prepare`; MR/fork refs are never fetched. The pinned SHA is used throughout; a later branch move/deletion has no effect. Must be empty for `stop`. |
| `duration` | empty (default), `<n>m`, `<n>h` or seconds, 600–28800 | empty: a stopped VM starts with **7200 s (2 h)** (CI profiles must keep `SESSION_MAX_SECONDS=7200`; the adapter refuses another value); a running VM keeps its session (no extension, no restart, the existing scheduled stop is reported). An explicit value differing from a running session is refused before any write (G070 semantics). Must be empty for `stop`. |

`stop` = action + environment only: the `staging:stop` rule requires empty `candidate` and `duration` (otherwise zero
jobs, creation refused) and the launcher re-checks (exit 2 before the token file is written). Job log and summary show
the effective SHA, selection text, branch (if any), actor, digests, session length and the scheduled stop read back
with `gcloud … describe` after the action.

**A1 fallback deleted (finding 6).** If root-config typed inputs (`options`/`regex`, interpolation in `rules`,
`environment`, `resource_group`, `id_tokens`) or the `no_one_allowed` pipeline-variable restriction are unavailable or
cannot be read back, G072 stops and returns to Codex for a reviewed design change. There is no pipeline-variable mode
and no recorded residual. Proof points: the MR fast pipeline parses the file with defaults; L-G0 reads back
`ci_pipeline_variables_minimum_override_role`; F-1a is the first deployment-mode pipeline.

### 10.3 Finding 1 — least-privilege metadata protocol

| | A: custom role on the state bucket, conditioned on `deploy-meta/<env>/` | **B: dedicated per-environment metadata bucket (recommended)** |
|---|---|---|
| Terraform-state isolation | the stop identity holds a mutating role on the bucket that holds `staging/<env>` state and locks; isolation rests on one condition expression; metadata shares the state bucket's versioning, lifecycle and teardown (bootstrap.sh 236–242 keeps the bucket while live objects remain) | stop/read identities hold **no** role on the state bucket: denial is the absence of a binding, verifiable with `testIamPermissions`; state/locks physically separate from ordinary metadata (Spec R7) |
| Least privilege | the condition must encode the prefix and the mutable keys | create bucket-wide on a bucket holding only this environment's metadata; delete (required to overwrite) only on two exact keys |
| Change surface | a binding on the existing state bucket (`bootstrap.sh` or a second script editing its policy) | a new bucket created and torn down by `ci-federation.sh`; `bootstrap.sh` unchanged |
| Cost | none | one bucket per environment holding KB objects, ≈ $0 |

**Recommendation B.** Bucket `gs://<GCP_PROJECT>-<STAGING_ENV>-ci-meta` (≤ 60 characters, in the environment's own
project), versioned, uniform bucket-level access, public-access prevention, 90-day noncurrent expiry, label
`managed-by=ci-federation`. Keys at the bucket root: `current.tsv`, `lease`, `lineage.tsv`, `fence.tsv`,
`schema/<sha>.tsv`, `deploy/<sha>.tsv` (G070 manifest format), `stop-requests/<pipeline-id>`, `runs/<pipeline>-<job>.tsv`,
`receipts/<id>.tsv`, `logs/<pipeline>-<job>.log` (sanitized only, §10.7), `lease-breaks/<utc>-<holder>.tsv`. Every write is
`gcloud storage cp … --if-generation-match=<0|observed>`, every delete `gcloud storage rm … --if-generation-match=<observed>`;
downloads are verified as in r0 S11. r0's "deploy-meta/<env>/ under the state bucket" (S11) is superseded.

Grants (single source NEW `infra/terraform/bootstrap/ci-meta-permissions.tsv`, W1: rendered into bindings by
`ci-federation.sh` and loaded as the IAM model by W2's fake GCS, so both fixtures test the same set):

| Principal | On `<env>-ci-meta`, bucket-wide | Exact-key condition | On the state bucket |
|---|---|---|---|
| `<env>-deployer` (CI start; local operator via ADC impersonation) | `roles/storage.objectAdmin`, `roles/storage.legacyBucketReader` (bucket get, for the presence check of §10.5) | — | G070 `objectAdmin` unchanged |
| `<env>-ci-stop` | `roles/storage.objectViewer`, `roles/storage.objectCreator`, `roles/storage.legacyBucketReader` | project custom role `ciMetaMutable` = {`storage.objects.delete`} with condition `resource.name == "projects/_/buckets/<B>/objects/current.tsv" \|\| resource.name == "projects/_/buckets/<B>/objects/lease"` | **none** |
| `<env>-ci-read` (new, §10.8) | `roles/storage.objectViewer`, `roles/storage.legacyBucketReader` | — | **none** |

Stop protocol under exactly those grants: read `lease` and `current.tsv`; create `stop-requests/<pipeline>` (gen 0);
create `lease` (gen 0), or break an expired CI lease (create `lease-breaks/…`, then delete `lease` at its observed
generation); control `staging.sh stop`; overwrite `current.tsv` (`stopped_at`, observed state, receipt ID) at its observed
generation (create + delete); create `runs/`, `receipts/`, `logs/` objects (gen 0); delete its own `lease` (observed
generation). It overwrites or deletes no other key, lists nothing in the state bucket and never reads `fence.tsv` or
`lineage.tsv`. **A8's fallback is deleted:** stop reads only durable metadata and observed VM state, never a GitLab
artifact, summary or earlier workspace (Spec R5/R7). If the exact-key condition cannot be bound or behaves otherwise at
L-W1p, G072 stops and returns to Codex; no wider grant is pre-approved here.

Proof:
- **ENV-LEASE (offline, W2):** fake GCS enforcing generation preconditions and the permission file per impersonated
  identity. Cycles start → stop → start → stop → stop run as `<env>-deployer`/`<env>-ci-stop`: every step permitted,
  `lease` absent after each job, `current.tsv` generation advances at each stop. With the delete condition removed from
  the model, the first stop fails at the `current.tsv` overwrite (the r0 defect, reproduced); any protocol step that
  needs an ungranted permission fails the row.
- **CI-FEDERATION (offline, W1):** dry-run renders the bucket settings, the three identities' bindings, the exact
  condition text and nothing on the state bucket for `ci-stop`/`ci-read`; one mutation per clause fails; `--teardown`
  removes exactly the created set and refuses while live metadata objects exist unless `--confirm-meta <env>` is given
  after retirement.
- **Live (F-1a/F-1b, L-H1):** two consecutive stops and a start/stop/start/stop cycle with read-backs (`lease` absent,
  `current.tsv` generation advanced per job). Non-mutating denial probes in every stop job: JSON-API
  `testIamPermissions` (token on a pipe, G070 `cf_records` pattern) on this environment's state bucket for objects
  get/list/create/delete → empty; the same on each owner-listed sibling environment's state and metadata buckets → empty;
  overwrite of the job's own just-created `runs/` record at its observed generation → 403 (a success fails the job with
  exit 6; the stray version is noncurrent and recoverable); impersonation of `<env>-deployer` → denied. Check and deploy
  jobs run the sibling probes for their own identities.

### 10.4 Finding 2 — schema-lineage fence

Records: `lineage.tsv` = `state` (`verified` | `assumed`), source candidate, writer, one `version<TAB>blob` row per
applied migration. `fence.tsv` = ordered entries, one per deploy started since the last verified lineage: time, lease
holder, candidate, sha256 of `schema/<candidate>.tsv`. `schema/<sha>.tsv` = that candidate's `core/migrations`
version→blob map (create-only; an existing object must be byte-equal).

**Write point (before migration):** inside G070 `deploy_candidate` (staging.sh 250–284), after the lease is held and
`require_pushed` and the RUNNING check pass, and **before** the release upload — the VM's `promote` and `migrate_current`
follow it (deploy.sh 100–106). It appends the candidate's entry to `fence.tsv` (read-modify-write at the observed
generation, under the lease); a failed fence write aborts before upload. It runs for CI deploys and for local
`staging.sh deploy`/`provision` whenever the environment's metadata bucket exists (§10.5). `start` and `reset` add no
entry: they migrate only the VM's current release, which is always the lineage source or an already fenced candidate,
because `promote` runs only inside a fenced deploy.

**Gate (replaces S7's comparison base; adapter; read-only; before any write, §10.6):** the possible-lineage set =
{`lineage.tsv`} ∪ {schema of every `fence.tsv` entry}. The candidate passes only if r0's S7 rule holds against **every**
member (same blob for each version up to that member's maximum; candidate maximum ≥ member maximum). Whatever subset of a
fenced candidate's migrations actually ran, a candidate that contains every member's migrations with identical blobs is
safe; an older or divergent candidate is refused with exit 2: "schema fence unreconciled since <holder> deployed <sha>;
deploy that candidate or a successor, or reconcile (STAGING.md)".

**Clearers** (nothing clears a fence on failure or cancellation; stop never touches lineage):
1. CI deploy success: after READY and running digests = manifest, the adapter reads the observed lineage; it must equal
   the candidate's schema exactly (every version applied, none extra). Then `lineage.tsv` := `verified` (observed
   generation) and `fence.tsv` is deleted (observed generation). A mismatch is exit 4 with the fence kept.
2. `ci.sh lineage reconcile --config FILE` (operator/owner with local G070 credentials, lease held, VM RUNNING): reads
   the observed lineage; each applied version must have one identical blob across every set member containing it →
   `verified` lineage, fence deleted. A version absent from all members, or with differing blobs, is HOLD (exit 3):
   an owner decision; data is discarded only by a separately authorized `reset --confirm`.
3. Setup only: `ci.sh lineage init --config FILE --candidate SHA` at L-W1p/L-W1s (refuses when `lineage.tsv` exists)
   records `assumed` lineage from the VM's current release candidate (the last G070 status log `candidate` line).
   `assumed` is an upper bound treated like a fence member until a CI deploy verifies it.

**Observed lineage read:** read-only, control code over `vm_ssh`, no VM script change — `sudo docker exec <postgres
container selected by com.docker.compose.project/service labels> psql --no-psqlrc --tuples-only --no-align --username
postgres --dbname <DATABASE_NAME> --command '<applied version IDs from app.goose_db_version, excluding 0, ordered>'`,
the `marker_label` pattern (remote/lib.sh 422–432: superuser socket, no secret). W2 confirms the protected goose store's
columns from `core/internal/migrate` (A13); the fixture pins the query text.

**Proof (CI-ADAPTER `--only fence`, W2; new boundary: partial-deploy lineage):** L0 = versions 1–24 verified. X = L0 +
25 deploys; the fake VM applies 25 and readiness fails (exit 4) → fence [X] kept, lineage still L0. Older W (= L0) →
**refused exit 2 with zero write calls** (no lease, OS Login key, push, compute, ssh or upload) — the Codex case. Same
after a cancellation injected between the fence write and readiness. Y = X + 26 → accepted; success verifies Y and
deletes the fence. Reconcile with observed 1–24 → verified L0, then W accepted; observed 1–25 → W still refused;
observed unknown 27, or a divergent blob → HOLD 3. A local `staging.sh deploy` of a divergent Z appends an entry, and
CI then refuses candidates incompatible with Z until reconcile.

### 10.5 Finding 3 — technically enforced environment lease

**Covered (mutating) G070 commands:** `provision`, `deploy`, `start`, `stop`, `reset`, `retire`. **Not covered:**
`build` (local Docker only), `push` (registry content only: per-candidate tags, digests recorded in the manifest; every
deploy pulls by digest and re-verifies; `retire --purge-images` is covered through `retire`), `status` and `receipt`
(read-only; `status` plans with `-lock=false`). `--dry-run` prints the lease commands and acquires nothing.

**Mechanism** (NEW `infra/staging/meta-lib.sh`, sourced by `staging.sh` and `ci-lib.sh`; writer W2): `main`
(staging.sh 648–665) calls `lease_acquire` before `cmd_<covered>`, and the existing EXIT trap (63–65) releases a lease
this process acquired (INT/TERM already exit through it). Presence check: `gcloud storage buckets describe
gs://<GCP_PROJECT>-<STAGING_ENV>-ci-meta` through `GFLAGS` — present and owned by `GCP_PROJECT` → lease mandatory;
positively absent (404) → the environment is not CI-enabled, one log line, then exactly today's behaviour; anything else
(403, error, foreign project) → refuse exit 2 (fail closed). CI requires the bucket, so wherever CI can act every covered
local command takes the lease. Acquire = create `lease` with `--if-generation-match=0` (TSV: holder
`local:<user>@<host>:<pid>` or `ci:<pipeline>/<job>`, command, candidate, acquired, expires); a precondition failure
reads the holder and refuses with exit 2 naming holder and times, before any other call. CI re-entrancy: the adapter
holds the lease and invokes control `staging.sh start|deploy|stop` with `STAGING_LEASE_HELD=<holder>@<generation>`; the
inner command reads `lease` and proceeds only on an exact holder and generation match (no acquire, no release). Lease
commands use `cp`/`rm` only, so G070 `forbid` (no ` create ` verb) is unchanged.

**Recovery:** CI lease TTL = job timeout + 5 min; a CI lease past its TTL belongs to an ended job (one environment's CI
jobs are serialized by `resource_group`), so the next CI job may break it: create `lease-breaks/…`, then delete at the
observed generation. A local lease is never broken by CI or by another command, expired or not. The holder releases
with `staging.sh lease --config FILE --release` (holder match); the environment owner's break-glass is `staging.sh
lease --config FILE --break --confirm <env> --reason TEXT` (create-only `lease-breaks/` record with operator, reason and
the broken lease's content, then delete at the observed generation); `--show` prints the lease. Interrupted runs keep
r0's recovery (G070 `reconcile`/`vm_status`/remote recovery before the next mutation); cancellation never resets data.

**Boundary stated honestly:** this prevents every accidental overlap of G070 mutators, local or CI. A principal with
deployer-equivalent rights who deliberately bypasses the tooling (raw `gcloud`, deleting `lease`) is not stopped by a
tooling control; bucket versioning, `lease-breaks/` records and Cloud Audit Logs make it attributable. No bypass is
documented as a procedure.

**Proof** (NEW `infra/staging/tests/env-lease.sh`, W2; NEW check.sh row `ENV-LEASE`, W4; new boundary: G070 mutator
exclusion): (a) local `staging.sh start` holds → CI deploy and CI stop refused exit 2 with zero write calls; (b) CI holds
→ each covered local command refused exit 2 before any non-lease call; (c) after holder release, or break-glass with its
record present, the waiting mutator proceeds; (d) an expired CI lease is broken with a record; an expired local lease
refuses CI; (e) `STAGING_LEASE_HELD` with a wrong generation or holder is refused; (f) **legacy parity:** for every
covered command, the baseline `staging.sh`/`lib.sh` (`git show <base>:…` into a temp dir) and the new ones run against
the same fakes with a free lease and with an absent bucket; after removing lease/presence calls and lease log lines,
call logs, stdout and exit codes are identical; (g) a 403 on the presence check refuses exit 2. LIFECYCLE-DRYRUN keeps
every existing assertion; only the printed lease lines are added (mechanical expectation update).

**Scope:** further additive G070 changes in `staging.sh` (`main`, cleanup, `deploy_candidate` fence call, `lease`
subcommand ≈ 60–90 lines) and new `meta-lib.sh` (≈ 250–350 lines); writer W2 only; ≈ +3–4 h coding and +100–150 fixture
lines; documented in §10.9.

### 10.6 Finding 4 — validate before any write

`staging:deploy` order:
- **Phase A — local only** (no cloud call, no token exchange): launcher context gate (16); input syntax; registry,
  selector, `CI_ENVIRONMENT_NAME` and profile identity (S4); profile parse plus CI keys and `SESSION_MAX_SECONDS=7200`;
  identity pins present and not `latest` (S16); duration bounds; `candidate.tsv` vs bundle commit vs policy for the
  recorded selection; S6 on the bundle; candidate schema from the bundle; build-artifact checksums when a build ran.
- **Phase B — authenticated, read-only:** credential files, token exchange and impersonation; reads of `current.tsv`,
  `lineage.tsv`, `fence.tsv`, `lease`, `stop-requests/` listing, `deploy/<sha>.tsv` with digest existence; VM describe
  and session. Gates: supersession, lease free (or an expired CI lease), lineage set (§10.4), a valid manifest or build
  artifacts (§10.8); RUNNING: duration rule and ≥ 300 s remaining; TERMINATED: the Cloudflare token is needed only for a
  length change.
- **Phase C — writes, only after every gate:** lease create → re-read the generations observed in B; any change re-runs
  B's gates (a refusal then deletes only the lease, exit 2) → OS Login key with TTL → images (reuse, or load, verify IDs,
  retag, push, upload manifest) → token fetch if needed → `staging.sh start` → `staging.sh deploy` (fence inside, §10.4)
  → readiness → observed-lineage verification → metadata → receipt → lease release.

`staging:stop`: phase A (gate, selector/profile, empty candidate and duration) → phase B (lease and VM state) →
create `stop-requests/<pipeline>` (an intent record, written also when the only refusal is a live local lease so an
older queued start cannot restart the environment after this stop request) → lease → OS Login key → `staging.sh stop`
→ metadata → release.

Proof (CI-ADAPTER call-log fixture; extends r0's zero-call negatives): the fakes classify each call as read or write.
Malformed candidate/duration/selector, profile mismatch, missing or `latest` pin, bundle/selection mismatch, S6
difference, lineage refusal, supersession, held lease, running-duration conflict, < 300 s remaining, and an invalid
manifest without build artifacts each exit 2 with **zero writes** (no OS Login key, lease, stop request, push, compute
start/stop/set-scheduling, ssh or upload); phase-A failures also show zero cloud calls and no token exchange.

### 10.7 Finding 5 — raw Terraform plans stay private

The adapter runs every `staging.sh` subprocess with stdout and stderr redirected to `STAGING_CI_PRIVATE/raw/<step>.log`
(0600, inside the private job directory that the trap and `after_script` remove, verified; never an artifact, never
uploaded). G070's `tf show -no-color` plan text (session refresh, staging.sh 348), Terraform init/plan/apply output and
`+ ` command echoes therefore stay there. Publication (NEW `infra/staging/ci-publish.sh`, W3; sourced by `ci.sh`):
`publish_filter RAW OUT` keeps only allowlisted line kinds — `staging: ` messages and the G070 key<TAB>value kinds
`step`, `result`, `session`, `address`, `reconcile`, `receipt`, `digest`, `cert`, `marker`, `service`, `secret` (version
numbers only); the exact list is pinned by a fixture from the `printf`/`say` keys of `staging.sh` and `remote/*.sh` — and
appends a dropped-line count. `canary_scan FILE…` scans raw and filtered files (r0 patterns plus the plan canary); a hit
in either is exit 6 with nothing published. `summary_render` writes the fixed-field summary. Durable `logs/`, the job log
and the summary artifact receive only filtered lines and the summary; readiness evidence is taken from the VM-side lines
of the filtered deploy log. S15 and §2.5 are superseded accordingly (`raw/` added to §2.5's private layout).

Checks: STAGING-GREP (extend, W4) — `ci.sh` never sends `staging.sh` output to its own stdout/stderr, never uploads from
`raw/`, publishes only through `publish_filter`. CI-LAUNCHER (W3) — a fake Terraform `show` prints a plan containing
`G072PLANCANARY` and resource attributes during a session change: the canary is absent from the job-log capture,
summary, artifacts and every fake-GCS upload, and `raw/` is removed; a canary inside an allowlisted line gives exit 6 with
nothing published. CF-TOKEN-CANARY and B-SECRET-CANARY are reused unchanged.

### 10.8 Finding 7 — automatic image reuse

The `images` input is removed. `start-deploy` runs four jobs:

| Job | Credentials | Role |
|---|---|---|
| `staging:prepare` | none (no `id_tokens`, `environment` or protected variables) | fetch `refs/heads/*`, resolve/pin (§10.2), S6, bundle, `candidate.tsv`, schema; **always runs**; artifacts (1 day, `access: none` if supported) |
| `staging:check` | `id_tokens` aud `staging-ci:<env>:read` → `<env>-ci-read` (§10.3 grants plus `roles/artifactregistry.reader` on the environment repository; nothing else); `environment: {name: <env>, action: prepare}`; no resource group | the automatic R3 decision **after authentication**: a valid `deploy/<sha>.tsv` (fields, candidate = filename, repository = profile, digest shape) whose seven digests exist → `STAGING_REUSE=yes`, else `no`; non-secret dotenv artifact |
| `staging:build` | none | `needs: [prepare, check]`; `STAGING_REUSE=no` → r0 S8 build; `yes` → writes a skipped marker, builds nothing |
| `staging:deploy` | start | `needs: [prepare, check, build]`; decides again authoritatively: valid manifest → reuse (no load/push); else build artifacts required → push; neither → exit 2 before any write |

Why: no credential-free job can read the private durable metadata; always building wastes a 30–60 min build per reuse;
an operator `images` choice is the unexplained extra input Codex rejected. The check identity reads nothing secret and
cannot mutate. Check/deploy disagreement fails closed: a manifest that disappeared → exit 2; one that appeared → reuse,
the build is discarded. Rejected: always-build; operator `images` input; a GitLab-side reuse hint (new publishing
surface writable by same-project MR pipelines).

Proof (CI-ADAPTER/CI-LAUNCHER): a fresh pipeline for an environment that already holds a valid published manifest and
**has no previous pipeline artifacts** → check `yes`, build skipped, deploy reuses with zero load/push calls and reports
the manifest's digests; invalid or mismatched manifests (candidate, repository, digest shape, missing digest) → check
`no`; check `yes` but invalid at deploy → exit 2 with zero writes; check `no` but valid at deploy → reuse. B-CI-CONFIG
pins the four jobs' credentials, `needs` and rules. Live: L-H1's second run reuses.

### 10.9 Additive G070 interfaces (D8)

These are **additive G070 interfaces**, not "no contract change"; defaults preserve today's behaviour, each is
documented in STAGING.md (W4) and proven against legacy invocations:
1. `staging.sh build --source DIR` (r0 §3.1).
2. Optional `staging.conf` key `GCLOUD_IMPERSONATE_SA`, validated as exactly `<env>-deployer`, `<env>-ci-stop` or
   `<env>-ci-read` `@<GCP_PROJECT>.iam.gserviceaccount.com`; absent → `DEPLOYER_SA` (r0 §3.2, extended by the read identity).
3. The environment lease on `provision|deploy|start|stop|reset|retire` when `<GCP_PROJECT>-<STAGING_ENV>-ci-meta` exists,
   and `staging.sh lease --config FILE --show|--release|--break --confirm ENV --reason TEXT` (§10.5).
4. The schema-fence entry written by `deploy_candidate` when that bucket exists (§10.4).
5. Only if A6 fails: optional `STAGING_SSH_KEY_EXPIRE_AFTER` (r0 §3.3).

Legacy proof: every existing LIFECYCLE-DRYRUN assertion unchanged (plus lease lines and `build --source` cases); ENV-LEASE
parity (f); a config without `GCLOUD_IMPERSONATE_SA` yields `GFLAGS` identical to today. No Terraform, remote-script,
manifest, receipt or exit-code change. Dispatch precondition: no pending G070 source change on staging.

### 10.10 Routing matrix and launcher/adapter contract (supersede §2.1/§2.2 where they differ)

| Context | Jobs |
|---|---|
| push (any branch), tag | none (unchanged) |
| MR (feature, fork), RC-promotion MR | validation as today; no `staging:*`; no `id_tokens` |
| schedule | validation as today; non-default action → workflow never |
| web/api on protected `staging`, `action=validate` | validation as today |
| web/api on `staging`, `start-deploy`, environment ∈ registry | `staging:prepare` → `staging:check` → `staging:build` → `staging:deploy` |
| web/api on `staging`, `stop`, environment ∈ registry, empty candidate and duration | `staging:stop` |
| web/api on `staging`, `stop` with candidate or duration set | zero jobs → creation refused |
| web/api on `staging`, non-default action, `environment=none` | zero jobs → creation refused |
| web/api on another branch, non-default action | workflow never |
| `identity-qualification` web/schedule, default action | `identity-staging` as today |

Contract deltas: `STAGING_CI_ACTION` ∈ prepare|check|build|deploy|stop; `STAGING_CI_IMAGES` removed; `STAGING_REUSE`
(check dotenv → build/deploy); `STAGING_CI_ID_TOKEN_FILE` and `STAGING_CI_PROFILE_DIR` for check, deploy and stop
(check receives only `STAGING_CI_CONF` and `STAGING_CI_SETTINGS`); `STAGING_LEASE_HELD` is adapter-internal (the
launcher unsets it from the job environment); `ci-publish.sh` exposes `publish_filter`, `canary_scan`, `summary_render`
(W3 → W2). The launcher refuses `stop` with candidate/duration and any action/environment pair outside the registry
before writing the token file. Exit classes S14 unchanged (lease and fence refusals are 2, reconcile ambiguity 3).
Timeouts: prepare 10m, check 10m, build 45m, deploy 60m, stop 30m. Provider audiences `staging-ci:<env>:start|stop|read`;
`attribute.action` ∈ start|stop|read|none; the §2.4 attribute condition is unchanged.

### 10.11 Files, writers, schedule (supersedes §4's table)

Four workers retained; no correction forces a fifth. The added lease/fence code shares `staging.sh`/`lib.sh` with W2's
existing changes and must stay single-writer; to balance load the log sanitizer moves to W3, which owns the launcher's
publication path.

| Worker | Exclusive writable paths | Delta vs r0 |
|---|---|---|
| **W1** | NEW `infra/terraform/bootstrap/ci-federation.sh`, NEW `infra/terraform/bootstrap/ci-meta-permissions.tsv`, NEW `infra/staging/tests/ci-federation.sh`; MOD `infra/terraform/README.md` | + metadata bucket, `<env>-ci-read`, `ciMetaMutable` exact-key binding, permissions file, bucket teardown |
| **W2** | NEW `infra/staging/ci.sh`, `ci-lib.sh`, `meta-lib.sh`, `ci-environments.txt`, `examples/ci-settings.example`, `tests/ci-adapter.sh`, `tests/env-lease.sh`, `tests/fakes/*`; MOD `infra/staging/staging.sh`, `infra/staging/lib.sh` (§10.9 only) | + lease, fence, lineage init/reconcile, phase order, prepare/check actions, parity fixture; − sanitizer |
| **W3** | MOD `.gitlab-ci.yml`; NEW `infra/ci/staging-action.sh`, `infra/docker/staging-ci.Dockerfile`, `infra/staging/ci-publish.sh`, `infra/staging/tests/ci-launcher.sh`; MOD `infra/validation/run.sh` (B-CI-CONFIG, B-SHELL list incl. `meta-lib.sh`/`ci-publish.sh`, B-HADOLINT) | + prepare/check jobs, stop rule, no `images` input, sanitizer |
| **W4** | phase 1: MOD `docs/foundation/STAGING.md` (+ empty-input/default table, lease/break-glass/handoff, fence/reconcile/init, metadata bucket, automatic reuse, additive interfaces), `docs/foundation/OPERATIONS.md`; NEW `infra/staging/tests/ci-build-boundary.sh` + hostile fixture. Phase 2: MOD `infra/staging/check.sh` (rows `CI-FEDERATION`, `CI-ADAPTER`, `ENV-LEASE`, `CI-LAUNCHER`; LIFECYCLE-DRYRUN lease lines and §10.9 cases; STAGING-GREP raw-log/lease-routing rules; SHELLCHECK-DIRECT list) | + `ENV-LEASE` row, docs |
| Coordinator | Result; proposed scope allowlist; integration, batches, MR | unchanged |

Sequence: P0 dispatch W1–W4 → W2 delivers `meta-lib.sh` + `staging.sh` lease/fence + `env-lease.sh` first (EP-1b), then
the prepare/build phase (EP-1 input), then the rest → coordinator EP-1b, then EP-1 → W4 phase 2 → integrate → S1–S4 →
Codex source review → MR (hosted fast + Duo; coordinator monitors) → merge → post-merge rows → live items.

### 10.12 Early proof (supersedes §1.3; riskiest first)

- **EP-1b — G070 interface change (offline, first, ≈ 10–20 min, no build slot):** `bash infra/staging/tests/env-lease.sh
  --work <T>/lease --report <E>/lease-1` and `bash infra/staging/tests/ci-adapter.sh --only fence --work <T>/fence
  --report <E>/fence-1`. Proves enforcement, legacy parity, the stop IAM model and the fence before the adapter is built
  on them. Stop criterion: a parity diff, an unrefused concurrent mutator or a refused-older-candidate case with a write
  → stop and report.
- **EP-1 — credential-free build boundary (offline):** r0 EP-1 unchanged; it remains the highest-impact security proof
  and measures A5.
- **EP-2 — first live action (F-1a, F-1b):** two consecutive `stop` pipelines for preview-1 after L-W1p: claims,
  single-hop `<env>-ci-stop`, exact-key overwrite of `current.tsv` and lease release twice, denial probes (§10.3). If
  preview-1 is not provisioned, G070 refuses ABSENT (exit 2) after the lease cycle: federation and lease create/delete
  are proven and the `current.tsv` overwrite proof moves to L-H1.

### 10.13 Validation and AC mapping (supersede the §5.1 S2 selector and §5.2 rows where they differ)

S2 becomes `bash infra/staging/check.sh run --work <T>/check --report <E>/offline-N --edge-image <retained edge ref>
--only CI-FEDERATION,CI-ADAPTER,ENV-LEASE,CI-LAUNCHER,LIFECYCLE-DRYRUN,STAGING-GREP,SHELLCHECK-DIRECT`. S1, S3, S4, hosted
and post-merge rules unchanged; no full ci-controls or application suite.

| AC | Source evidence (reuse / extend / new — uncovered boundary) | Live evidence |
|---|---|---|
| AC1 routing | **extend** B-CI-CONFIG: inputs without `images`; five jobs' exact keys (prepare/build without `id_tokens`/`environment`/protected variables; check with the `read` audience, `environment action: prepare`, no resource group; deploy/stop as r0); stop rule requiring empty candidate/duration; `needs` chain; registry = options. **reuse** S4 controls. **extend** CI-LAUNCHER: stop with inputs refused, `STAGING_LEASE_HELD` unset — boundary: launcher | F-1a; one refused creation (stop with a candidate) |
| AC2 candidate fidelity | **new** CI-ADAPTER: empty candidate = frozen pipeline SHA for staging **and** preview-1 (boundary: approved default); explicit preview SHA/branch resolved once, moved/deleted branch uses the bundle; staging first-parent rules; fork/MR refs refused; automatic reuse with no previous pipeline artifacts (boundary: R3 decision point); validate-before-write call-log negatives (boundary: ordering). **extend** LIFECYCLE-DRYRUN `build --source`; S3 image-ID preservation | L-H1 digests = manifest of the pinned SHA; reuse run without build/push |
| AC3 identity/secrets | **new** CI-FEDERATION: three audiences/identities, metadata bucket, exact-key condition, permissions file, teardown. **new** ENV-LEASE IAM model (boundary: protocol needs exactly the grants). **new** CI-LAUNCHER publication cases (boundary: durable-log sanitization). CI-ADAPTER credential files, token only for a length change, `GCLOUD_IMPERSONATE_SA` validation; S3 canaries; **reuse** CF-TOKEN-CANARY, B-SECRET-CANARY | F-1a federates; §10.3 denial probes; L-W1p sanitized policy read-backs |
| AC4 start/deploy | **new** CI-ADAPTER: RUNNING/TERMINATED, empty duration (2 h stopped / keep running), duration conflict, ≥ 300 s, pins, readiness, fence/lineage cases (boundary: partial deploy). **reuse** G070 L-5 rerun expiry/retained-data evidence | L-H1; L-H2 (gated) |
| AC5 stop/recovery | **new** ENV-LEASE: both-direction concurrency for every covered command, expired CI break, local never broken by CI, break-glass record, re-entrancy, legacy parity, absent/unreadable bucket (boundary: G070 mutator exclusion). **new** CI-ADAPTER: stop without candidate, artifacts or earlier summary; stop after a failed/canceled start; three start/stop cycles; supersession; rc 255 + TERMINATED after the deadline → exit 5 | F-1a/F-1b repeated stop; L-H1 cycle; L-H2 |
| AC6 closeout | unchanged (r0) | unchanged |
| R12 | r0 row plus the preview default candidate and the sibling metadata/state probes | r0 row |

### 10.14 Live items (supersede §5.3 rows where they differ; none released)

| ID | Action | Performer | Precondition | Time / cost | Rollback | Release |
|---|---|---|---|---|---|---|
| L-G0 | Read-only readiness: tier and D9 features (root typed inputs with options/regex and interpolation, protected environments, ID-token claims incl. `environment_protected`/`environment_action`, `artifacts:access`, artifact size limit, readable `ci_pipeline_variables_minimum_override_role`), `staging` protection, project path/ID = pins. **Any unavailable → stop and return to Codex.** | coordinator | merge | 15–20 min / $0 | — | Codex |
| L-G1 | as r0 | owner (or coordinator with in-session owner approval) | L-G0 | 30–45 min / $0 | delete settings | owner + Codex |
| L-P0 | **The preview-1 GCP project (D1-A)**, billing link, monthly alert budget | owner | Plan ACCEPT | 30–60 min / $0 | delete project | owner |
| L-P1 | as r0 (G070 local path; no metadata bucket yet, so no lease or fence) | owner; coordinator | L-P0, G070-style proposal | 1.5–2.5 h | as r0 | owner live release + Codex |
| L-W1p | `ci-federation.sh` for preview-1 (dry-run reviewed first): pool/provider with three audiences, `ci-stop`, `ci-read`, metadata bucket, bindings and condition; read-back; then `ci.sh lineage init --candidate <L-P1 candidate>` | owner (federation); coordinator (lineage init) | L-P1 (L-P0 for a federation-only EP-2) | 30–40 min / ≈ $0 | `--teardown` | owner + Codex |
| F-1a, F-1b (EP-2) | two consecutive `stop` pipelines for preview-1 with denial probes and read-backs | coordinator | L-G1, L-W1p | 25 min / CI ≈ 30 units | none (metadata retained) | Codex |
| L-G2 | as r0 (`oldest_first` for `staging-env-*`, `staging-image-build`; check has no group) | owner/coordinator | after F-1a | 10 min | revert mode | Codex |
| L-H1 | preview-1: start-deploy with a same-repository branch override and a custom duration (build) → stop → start-deploy of the same SHA (automatic reuse, no build/push) → stop; optional default-candidate run; optional supersession check | coordinator | F-1a/F-1b | 75–100 min / VM ≈ $0.12, CI ≈ 180 units | stop pipeline; G070 local stop | owner + Codex |
| L-W1s, F-3 | as r0, plus the staging metadata bucket and `lineage init` from the retained baseline's candidate | owner; coordinator | **G057 retention hold released, or an explicit owner release coordinated with G057**; staging identity pins set | 35–50 min / $0 | `--teardown` | owner + Codex |
| L-H2 | as r0 | coordinator | L-W1s/F-3, same G057 gate | 60 min / ≈ $0.15 | stop pipelines | owner + Codex |

Retention and retirement remain owner decisions (r0); each environment's metadata bucket is retired with
`ci-federation.sh --teardown` after the environment's retirement.

### 10.15 Estimates (supersede §6; unmeasured ranges)

| Category | Estimate |
|---|---|
| Coding (agent) | W1 4.5–7 h, W2 10–14 h, W3 5–7.5 h, W4 3.5–5.5 h ≈ **23–34 agent-hours** (r0 17–26); wall clock ≈ **11–15 h** on W2's critical path, plus coordinator integration 3–5 h |
| Build/test elapsed | S1 ≈ 6 min/run; S2 ≈ 5–8 min/run (ENV-LEASE adds ≈ 2–3); S4 5–10 min; EP-1b 10–20 min; S3/EP-1 45–75 min; with two or three correction reruns ≈ **2.5–4 h** |
| Review/queue waits | Plan r1 review, 1–3 source-review rounds and Duo: **unknown**; hosted fast ≈ 13–17 min per push |
| Live/setup | preview path (L-G0…L-H1) ≈ **4.5–7 h** elapsed incl. owner steps; staging path ≈ **1.5–2.5 h**, date **unknown** (G057) |
| Total | ≈ **3.5–6 working days** elapsed; scope ≈ 3.2–4k lines incl. fixtures (r0 2.5–3k); a staging-first split still shares the build boundary, federation and lease work and is not recommended |

Unknowns: real artifact size (A5), GitLab feature availability (D9), review rounds, owner windows.

### 10.16 Risks, decisions and unverifiable inputs (supersede §8 where they differ)

Stop conditions: r0 §8 list unchanged, plus: a lease-parity diff; any protocol step needing a grant outside §10.3;
unavailable typed inputs or variable restriction (§10.2); an unavailable exact-key condition (A8′).

Residual risks: deliberate lease bypass by a deployer-equivalent principal (attributable, not prevented, §10.5); the
presence check depends on 404 for a nonexistent bucket (A14), and a transient error fails closed and blocks a local
command until retried; the fence set can refuse a safe candidate until reconcile, which needs a RUNNING VM; local deploys
add unverified fence entries and `current.tsv` describes the last CI deployment until the next CI deploy verifies (the
summary reports open fence entries); check/deploy disagreement fails closed; r0's `osAdminLogin` and zone-scoped
Cloudflare boundaries unchanged; artifact size (A5).

| # | Decision | Status |
|---|---|---|
| D1 | preview isolation | **settled (owner):** separate GCP project per preview slot |
| D2, D5, D6 | build/deploy separation; first-parent; compatibility refusal | **settled (Codex)** |
| D3 | protected environments/variables, `no_one_allowed` | owner action at L-G1 (not a design choice) |
| D4 | local exclusion | **closed:** technical enforcement (§10.5) |
| D7 | who updates staging `CI_IDENTITY_PINS` during G057 R-1…X-9 | **open — owner** |
| D8 | §3 changes | **settled:** additive G070 interfaces (§10.9) |
| D9 | GitLab tier/features | **settled:** verify at L-G0; unavailable → stop |
| D10 | metadata protocol option B (§10.3) | **for Codex in this review** |
| D11 | read-only `staging:check` as the automatic reuse decision point (§10.8) | **for Codex in this review** |
| D12 | local `staging.sh deploy` keeps legacy semantics (lease + fence entry, no local S7 refusal; enforcing S7 locally would add a G070 refusal) | **for Codex in this review** |
| — | separate Cloudflare token per environment; preview-1 retention after the proof | owner choices, non-blocking (r0) |

Assumptions changed by r1: A1 — fallback deleted (stop). A8 — replaced by **A8′** (GCS IAM condition on exact object
names governs delete/overwrite under uniform access; proof L-W1p read-back, F-1a/F-1b; no fallback). New: **A12**
`gcloud storage cp/rm --if-generation-match` (0 and observed) semantics and a 412 distinct from 403; **A13** goose
version-table columns in the protected store; **A14** GCS answers 404 for a nonexistent bucket to an authenticated
caller without rights. A5's r0 fallback (a publish identity next to the build) now needs a Codex decision because D2 is settled.

| # | Unverifiable input | Why not verifiable now | Proof point | If false |
|---|---|---|---|---|
| U1 | GitLab tier/features (typed root inputs and interpolation, protected environments, ID-token claims incl. `environment_action`, `artifacts:access: none`, artifact size limit, pipeline-variable restriction) | no live read in planning | MR fast parse; L-G0; F-1a | stop (D9, §10.2) |
| U2 | `staging` branch protection | same | L-G0 | stop |
| U3 | google-cloud-cli image contents/digest (A7) | no pull in planning | W3 at implementation (cache-prep lock) | r0 A7 fallback |
| U4 | `external_account` file + single impersonation hop (A3) | live | F-1a, L-H1 | r0 A3 fallback |
| U5 | direct federated Secret Manager access (A4) | live | L-H1 | r0 A4 fallback |
| U6 | OS Login pre-registered TTL key (A6) | live | F-1a, L-H1 | §10.9 item 5 |
| U7 | exact-object-name IAM condition (A8′) | live | L-W1p, F-1a/F-1b | stop, return to Codex |
| U8 | generation preconditions in `gcloud storage` (A12) | tool behaviour | fixture model; F-1a | same semantics through the JSON API (Plan delta if needed) |
| U9 | goose table columns (A13) | code read at implementation | W2 fixture; L-H1 observed read | mechanical query correction |
| U10 | 404 for a nonexistent bucket (A14) | live | L-W1p read of a never-created derived name (read-only) | fail closed; Plan delta (explicit config key) |
| U11 | compressed seven-image artifact size (A5) | needs a real build | EP-1 | stop; Codex decision |
| U12 | G070 closeout evidence (L-5 rerun, L-6–L-8, A2 proof) remains only on `origin/g070-closeout-4` `94f2e06c`, not on staging `0c66332d` | publication pending | re-read at dispatch | cite from that branch |
| U13 | Fable allowance behind the model note | not independently verified | — | — |

### 10.17 History

- 2026-10-28 r1 — CORRECTION LOOP response to `G072-plan-r0-codex-review-20261028.md`: rebound to staging `0c66332d`
  (management-only delta), D1 recorded, Codex-settled D2/D5/D6/D8/D9 applied, findings 1–8 closed in §§10.2–10.8, four
  workers kept, estimates revised; planning only; zero attempts spent.


## Owner-approved configuration sequencing and identity pins — 2026-10-28

Owner selects configuration first, coordinated live validation afterward, and
Codex-only setup execution. This amends the preview-none setup default, not
the trust boundaries or required proofs: each environment receives its own
GCP Secret Manager CLIENT_ID/ISSUER/API_KEY numeric triple from its distinct
WorkOS environment. Current desired pins are 1/1/1 in both environment-scoped
SETTINGS variables. No key/client/issuer value is stored in GitLab. G072 live
CI proofs, preview's VM-dependent federation readback, lineage/fences, resource
groups, and G057 privacy/browser gates remain. No full-suite rerun is added for
configuration publication, and no live deploy PASS is inferred from pin setup.
