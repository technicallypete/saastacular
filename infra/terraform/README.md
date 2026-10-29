# Staging infrastructure (Terraform)

Generic infrastructure for one on-demand staging environment on Google Cloud: one ordinary (non-Spot) VM with a
retained data disk, a reserved static IPv4, a dedicated VPC, and three DNS-only Cloudflare records. Every value
is an input; examples use `staging.example.net`, `replace-me` and `PROJECT_ID`. The operator runbook is
`docs/foundation/STAGING.md`; the lifecycle commands (`provision`, `deploy`, `start`, `stop`, `reset`, `retire`,
`status`) are `infra/staging/staging.sh`. This directory never deploys the application.

```
bootstrap/bootstrap.sh          one-time gcloud setup (owner ADC); the retained bootstrap set
bootstrap/ci-federation.sh      GitLab CI federation per environment project (owner; G072)
bootstrap/ci-meta-permissions.tsv  the metadata bucket's bindings (single source for the script and its fixtures)
modules/gcp-staging-vm/         VPC, subnet, 2 firewall rules, address, data disk, VM, startup.sh
modules/cloudflare-dns/         B, *.B, custom.B: A records, proxied = false, TTL dns_ttl (300)
environments/example/           the only committed root (partial "gcs" backend); generic values only
```

## Inputs and private files

The adopting environment's `terraform.tfvars` and `backend.hcl` (copied from the `.example` files) live with its
`staging.conf` in the private inputs directory (`<main checkout>/.local/secrets/<task>/<env>/`, mode 0600,
ignored by Git). They hold identifiers only. Secret values exist only in Secret Manager: no Terraform variable,
output, state entry, plan file or instance metadata carries one (a `sensitive` flag would not keep a value out of
state, so none is ever passed). Credentials are never written to these files:

- GCP: the operator's short-lived ADC (`gcloud auth application-default login`); Terraform and the backend
  impersonate `<env>-deployer` (`impersonate_service_account`). No service-account key exists anywhere.
- Cloudflare: `CLOUDFLARE_API_TOKEN` in the operator environment only, a zone-scoped token (`Zone:Read`,
  `DNS:Edit`) distinct from the edge's DNS-01 token.

## One-time bootstrap (owner)

```
infra/terraform/bootstrap/bootstrap.sh --project PROJECT_ID --region REGION --env staging \
  --operator user:ops@example.net [--dry-run]
```

Idempotent; prints one `bootstrap<TAB>category<TAB>name<TAB>state` receipt line per item (`existing`, `created`,
`enabled`, `generated`, `present`, `awaiting-owner`). It creates, when absent: the APIs; the state bucket
`<project>-tfstate` (versioning, uniform access, public-access prevention, noncurrent versions expire after 90
days; prefix `staging/<env>` per environment); service accounts `<env>-vm` and `<env>-deployer` with the bindings
below; the Docker repository `<region>-docker.pkg.dev/<project>/<env>-images`; and 13 secrets `<env>-<NAME>`. The
8 application secrets get one generated version (never printed); the owner adds `CF_API_TOKEN` and, optionally,
`OPENROUTER_API_KEY` from their own terminal (`gcloud secrets versions add <env>-CF_API_TOKEN --data-file=-`);
the three `IDENTITY_WORKOS_*` secrets stay empty for the identity owner. Environment names for bootstrap are 3-21
characters (service-account ID length).

These items are the retained bootstrap set (owner: environment owner; reason: state history, secret versions,
images; release: no environment state object remains). `bootstrap.sh --teardown --confirm PROJECT_ID ...` removes
them and refuses while the environment's state object or any labelled VM, disk or address exists.

## CI federation (owner, G072)

`bootstrap/ci-federation.sh` lets the manual GitLab CI start/deploy and stop jobs (`docs/foundation/STAGING.md`)
federate into one environment project without any stored credential. The environment owner runs it from their own
terminal with their short-lived `gcloud auth login`; no agent and no CI job runs it. Each environment has its own
project (staging and every preview slot), so everything below is per environment and no identity of one
environment holds a role in another's project.

```
infra/terraform/bootstrap/ci-federation.sh --project PROJECT_ID --project-number PROJECT_NUMBER --region REGION \
  --zone ZONE --env ENV --gitlab-project-path GROUP/PROJECT --gitlab-namespace-id NAMESPACE_ID [--dry-run]
infra/terraform/bootstrap/ci-federation.sh --verify (the same arguments) [--dry-run]
infra/terraform/bootstrap/ci-federation.sh --teardown --confirm PROJECT_ID --project PROJECT_ID \
  --project-number PROJECT_NUMBER --region REGION --zone ZONE --env ENV [--confirm-meta ENV] [--dry-run]
```

Inputs are checked before any `gcloud` call: `ENV` must be a line of `infra/staging/ci-environments.txt` (exact,
so `preview-1` is not `preview-10`); `--project-number` is a pin compared with the project's number; the GitLab
project ID is read from `infra/ci/hosted-context.sh` (`STG_PROJECT_ID`) and `--gitlab-project-path` must hash to its
`STG_PROJECT_PATH_SHA` (the path itself is never committed); `--gitlab-namespace-id` is the numeric ID of the
project's namespace, read by the owner from GitLab.

**Order.** 1) `bootstrap.sh` for the environment (deployer, VM account, repository, state bucket); 2) the owner creates
`<env>-TERRAFORM_CLOUDFLARE_API_TOKEN` (the operator Terraform token, never readable by the VM or the deployer);
3) `staging.sh provision` so `<env>-vm` exists; 4) `ci-federation.sh --dry-run`, reviewed; 5) `ci-federation.sh`;
6) `ci-federation.sh --verify`; 7) `ci.sh lineage init` (STAGING.md). A missing secret or VM is reported
(`awaiting-owner`, `awaiting-provision`) and its binding is applied by a rerun; `--verify` fails until then.
Re-provisioning recreates `<env>-vm` and loses its instance-level binding: rerun the script, then `--verify`.

**What it creates** (idempotent; one `ci-federation<TAB>category<TAB>name<TAB>state` receipt per item):

- `sts.googleapis.com` enabled.
- Workload identity pool `<env>-ci` and OIDC provider `gitlab`: issuer `https://gitlab.com`; allowed audiences exactly
  `staging-ci:<env>:start`, `staging-ci:<env>:stop`, `staging-ci:<env>:read`; mapping `google.subject=assertion.sub`
  and `attribute.action` = `start`/`stop`/`read` for those exact audiences, else `none`; attribute condition: the
  pinned `project_id` and `namespace_id`, `ref_type == "branch"`, `ref == "staging"`, `ref_protected == "true"`,
  `ci_config_ref_uri == "gitlab.com/<path>//.gitlab-ci.yml@refs/heads/staging"`, `pipeline_source in ["web", "api"]`,
  `environment == "<env>"`, `environment_protected == "true"`, `runner_environment == "gitlab-hosted"` and `aud` in the
  three audiences. Rerunning re-applies the provider, so a hand edit converges back.
- Service accounts `<env>-ci-stop` and `<env>-ci-read` (`<env>-deployer` stays bootstrap's, grants unchanged).
- Custom roles `ciStop<env>` (hyphens become underscores, e.g. `ciStoppreview_1`; role IDs allow no hyphen) and
  `ciMetaMutable` = {`storage.objects.delete`}; permissions re-applied on every run.
- Metadata bucket `gs://<project>-<env>-ci-meta` (at most 60 characters; versioned, uniform bucket-level access,
  public-access prevention, noncurrent versions expire after 90 days, label `managed-by=ci-federation`). Its bindings
  are rendered only from `bootstrap/ci-meta-permissions.tsv` (`principal<TAB>grant<TAB>keys`; `keys` is `*` or the
  exact object names a `ciMetaMutable` grant is conditioned on, here `current.tsv,lease`); the same file is the IAM
  model of the adapter's fake storage. The Terraform state bucket gets no binding from this script.
- The principal bindings of the CI rows in the table below.

`--dry-run` prints every mutating command (`+ `) without calling `gcloud`; review it first, in particular the
attribute condition and the exact-key condition. `--verify` uses describe, get-iam-policy and list calls only and
compares the provider fields, both roles, the bucket settings, the absence of user-managed keys and, for every policy
a CI principal could appear in (project, the four service accounts, the repository, the VM, both buckets, every
secret), exactly the rendered bindings; any extra or missing binding, including one on the state bucket, fails it
(exit 1). `--verify --dry-run` prints those reads (`? `).

**Rollback.** `--teardown --confirm PROJECT_ID` removes exactly what setup created: the CI bindings, both service
accounts, both custom roles, the provider, the pool and the metadata bucket; the API stays enabled and no bootstrap
item, secret or the VM is touched. It refuses, before any change, while the metadata bucket holds a live object
(`current.tsv`, `lease`, receipts, logs) unless `--confirm-meta <env>` is given after the environment's retirement.
The pool, provider and roles are soft-deleted by Google Cloud (pools and providers for 30 days, custom roles for 7
days); a later setup in that window undeletes them. After a role's purge its ID stays reserved for some weeks, during
which a setup fails at the role create until the ID is free. Revoking CI quickly without teardown: disable the provider in the
console or remove the principal-set bindings.

Unverified until the first live setup (Plan U-items; the offline fixture `infra/staging/tests/ci-federation.sh` checks
the rendering, not Google's parser): gcloud's dictionary parsing of `--attribute-mapping` and of
`--condition=expression=...,title=...` whose values contain `==` (no commas are used inside them); `--condition=None`
on bucket bindings; the `value(...)` projections `--verify` reads; and the IAP tunnel resource name. The stop
identity's `iap.tunnelResourceAccessor` is scoped to the instance (Plan §2.4) by a project binding with the IAM
condition `resource.name == "projects/<project-number>/iap_tunnel/zones/<zone>/instances/<env>-vm"` (title
`ci-stop-iap-instance`); the script refuses any unconditioned grant of that role, `--verify` fails on a project-wide
one and `--teardown` removes the conditioned binding. Whether IAP evaluates the instance name (rather than only its
numeric ID) in that resource name is unverified: a mismatch fails closed (the stop job's tunnel is denied, exit 4 at
F-1a) and is corrected by changing the condition, never by widening it.

## IAM and trust boundaries

| Principal | Roles | Scope |
|---|---|---|
| Human operator (owner ADC) | existing project roles (recorded, not expanded); `iam.serviceAccountTokenCreator` | `<env>-deployer` only |
| `<env>-deployer` (impersonated) | `compute.instanceAdmin.v1`, `compute.networkAdmin`, `compute.securityAdmin`, `iap.tunnelResourceAccessor`, `compute.osAdminLogin` | project |
| | `iam.serviceAccountUser` | `<env>-vm` only |
| | `storage.objectAdmin` | state bucket only |
| | `artifactregistry.writer` | the repository only |
| `<env>-vm` (attached, scope `cloud-platform`) | `secretmanager.secretAccessor` | each `<env>-*` secret |
| | `artifactregistry.reader` | the repository only |
| `operator_principals` (optional) | `compute.osAdminLogin` | this VM only (Terraform) |
| CI start principal set (`attribute.action/start`, `ci-federation.sh`) | `iam.workloadIdentityUser` | `<env>-deployer` only |
| | `secretmanager.secretAccessor` | the one secret `<env>-TERRAFORM_CLOUDFLARE_API_TOKEN` |
| CI stop principal set (`attribute.action/stop`) | `iam.workloadIdentityUser` | `<env>-ci-stop` only |
| CI read principal set (`attribute.action/read`) | `iam.workloadIdentityUser` | `<env>-ci-read` only |
| `<env>-ci-stop` | custom `ciStop<env>` (`compute.instances.get/list/stop`, `compute.disks.list`, `compute.addresses.list`, `compute.zoneOperations.get`, `compute.projects.get`) | project |
| | `iap.tunnelResourceAccessor` (project binding conditioned on the `<env>-vm` tunnel resource) | `<env>-vm` instance only |
| | `compute.osAdminLogin` | `<env>-vm` instance only |
| | `iam.serviceAccountUser` | `<env>-vm` account only |
| | `storage.objectViewer`, `storage.objectCreator`, `storage.legacyBucketReader`; custom `ciMetaMutable` (`storage.objects.delete`) only on the objects `current.tsv` and `lease` | metadata bucket only |
| `<env>-ci-read` | `storage.objectViewer`, `storage.legacyBucketReader` | metadata bucket only |
| | `artifactregistry.reader` | the repository only |
| `<env>-deployer` (CI start, or the operator by impersonation) | `storage.objectAdmin`, `storage.legacyBucketReader` | metadata bucket (in addition to the rows above) |

Never granted: the Owner or Editor basic roles, project IAM administration, service-account key administration,
any Secret Manager role on the deployer, any compute or storage role on the VM account. The VM account cannot
stop, start, snapshot or modify the VM; expiry is Compute Engine's. For CI additionally: no role on the Terraform
state bucket for `<env>-ci-stop` or `<env>-ci-read`, no Secret Manager role on any service account, no
service-account key anywhere, and no CI principal in another environment's project.

Actual access boundaries, recorded rather than narrowed: injected secret values also exist in the containers'
configuration on the encrypted data disk (Google-managed keys) for as long as those containers exist, stopped
ones included; the env file `/run/<env>/app.env` is on tmpfs and removed after each run. Whoever can reach the
Docker socket (root over OS Login with sudo through IAP) or administer the disk can read them: the deployer
account, listed operators, and the project's and organization's administrators (Owner/Editor, Compute Admin,
OS Login administrators, anyone able to impersonate the deployer) - a wider set than the Secret Manager readers.
No command snapshots or images the data disk.

## Provision, plan and apply

`infra/staging/staging.sh provision` runs these from `environments/example` (shown for review):

```
terraform init -backend-config=<private>/backend.hcl -input=false
terraform plan -var-file=<private>/terraform.tfvars -var session_max_seconds=<n> -out=<private>/plan.bin   # always shown and kept private
terraform apply <private>/plan.bin                                                                        # only with --yes
```

`<n>` is the session length: on first allocation `--duration` or the configured default, on an existing VM the length
the instance reports (the module ignores it there: provision never changes it).

Every labelled resource carries `environment=<env>` and `managed-by=terraform` (VM, both disks, address);
names are deterministic (`<env>-vm`, `<env>-data`, `<env>-ip`, `<env>-vpc`, `<env>-subnet`, `<env>-edge`,
`<env>-iap`), so a rerun converges or fails on "already exists" and is never a second resource. Outputs:
`instance_name`, `instance_zone`, `instance_self_link`, `external_ip`, `service_account_email`, `data_disk_name`,
`network_name`, `subnetwork_name`, `firewall_names`, `max_run_duration_seconds`, `dns_record_ids`.

**Partial apply.** State is written per resource: rerun `plan` (it shows the remainder) and `apply`. An orphan
(created but not in state) is found by comparing `gcloud compute instances|disks|addresses list
--filter=labels.environment=<env>` (and the Cloudflare records by name) with `terraform state list`, then
`terraform import`. A stale lock is released with `terraform force-unlock <id>` only after the inventory shows no
running apply.

## Stop, start and the session limit

`desired_status` is not managed: `stop`/`start` use `gcloud compute instances stop|start`, and `terraform plan`
afterwards is expected to show no change. `scheduling` sets `provisioning_model = "STANDARD"`,
`instance_termination_action = "STOP"` (the only accepted value) and `max_run_duration` = `session_max_seconds`
(default 7200 s, validated to 600..28800 whole seconds) when the VM is created. Compute Engine clears the deadline at
every stop and re-arms it at every start. hashicorp/google 8.8.1 replaces the VM when `max_run_duration` changes, so
the module carries `lifecycle { ignore_changes = [scheduling[0].max_run_duration] }` and `-var session_max_seconds`
no longer changes an existing VM. To change the length: `staging.sh stop`, then `staging.sh start --duration D`,
which sets it on the stopped VM with `gcloud compute instances set-scheduling`, reads the whole scheduling back, saves
a `terraform plan -refresh-only` whose only drift must be that length of `module.vm.google_compute_instance.vm`
(`scheduling[0].max_run_duration[0].seconds`, plus `current_status`/`cpu_platform` and the recorded-state lag leaves of a
never-refreshed VM), applies exactly that plan,
verifies the instance and the `max_run_duration_seconds` output, and then starts the VM (docs/foundation/STAGING.md
"Session length", including the recovery of a partial change).

## Startup script (`modules/gcp-staging-vm/startup.sh`)

Stored in the plain `startup-script` metadata key (updated in place) behind one rendered line
`export STAGING_ENV=<env>`; metadata keys are exactly `enable-oslogin` and `startup-script`. At every boot it
formats the data disk once (only when it has no filesystem or partition table; ext4 label `staging-data`, device
`google-staging-data`), mounts it at `/var/lib/docker` through `/etc/fstab`, makes `docker.service` require that
mount, installs Docker Engine 29.8.2 and Compose 5.5.1 from Docker's apt repository (key fingerprint checked,
packages held, classic image store so images stay on the data disk), creates `/srv/<env>` (0750) and
`/run/<env>` (0700) and refuses unless `findmnt` reports `/run/<env>` on tmpfs. It creates no secret and never
pulls, deploys or starts the application.

## Retirement

`staging.sh retire --yes`: `stop`, `terraform plan -destroy` (shown, kept private), `terraform destroy`, then
absence checks (label queries empty, the three Cloudflare names absent, `terraform state list` empty, address
released). The bootstrap set stays, named in the receipt with its owner, reason and release condition.

## Offline checks (coordinator; no credentials, `--network none`)

Pins: Terraform 1.16.7 (`hashicorp/terraform:1.16.7@sha256:<digest>`), `hashicorp/google` 8.8.1,
`cloudflare/cloudflare` 5.28.1 (`environments/example/.terraform.lock.hcl`, written once with network by
`terraform providers lock -platform=linux_amd64 -platform=linux_arm64 -platform=darwin_arm64`). With a prepared
read-only provider mirror, a disposable copy `<T>/tf-work/<run>/src` of this tree, `cli.tfrc` =
`provider_installation { filesystem_mirror { path = "/mirror" } direct { exclude = ["*/*"] } }` and
`RUN='docker run --rm --network none --entrypoint /bin/sh -v <T>/tf-work/<run>:/work -v <T>/tf-mirror:/mirror:ro
-e TF_CLI_CONFIG_FILE=/work/cli.tfrc -e TF_IN_AUTOMATION=1 -e TF_INPUT=0 hashicorp/terraform:1.16.7@sha256:<digest>'`:

```
$RUN -c 'terraform fmt -check -recursive -diff /work/src'
$RUN -c 'export TF_DATA_DIR=/work/data/root; cd /work/src/environments/example && terraform init -backend=false -lockfile=readonly && terraform validate -json'
$RUN -c 'export TF_DATA_DIR=/work/data/gcp-staging-vm; cd /work/src/modules/gcp-staging-vm && terraform init -backend=false && terraform validate -json && terraform test -json -no-color'
$RUN -c 'export TF_DATA_DIR=/work/data/cloudflare-dns; cd /work/src/modules/cloudflare-dns && terraform init -backend=false && terraform validate -json && terraform test -json -no-color'
```

Each module's tests (`tests/plan.tftest.hcl`, mocked providers, `command = plan`) run only from that module's
directory. A module passes when its `test_summary` reports `failed`, `errored` and `skipped` 0 and `passed` at
least 1; a missing summary or zero runs is a failure. Startup-script package pins, without disk or tmpfs guard:
`docker run --rm -i debian:12@sha256:<digest> bash -s -- --no-disk < modules/gcp-staging-vm/startup.sh`.

Copyright 2026 Vitruvian Technology, Corp. Licensed under the Apache License, Version 2.0; see the repository `LICENSE` and `NOTICE`.
