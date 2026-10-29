# On-demand staging on Google Cloud

This template includes reusable tooling for one on-demand staging environment per project: a regular Compute Engine
VM that runs the existing Compose services with private local PostgreSQL, Redis and MinIO, a Caddy edge with trusted
Let's Encrypt certificates, Cloudflare DNS records managed by Terraform, and one operator lifecycle command. The
tooling is provider-specific (Google Cloud, Cloudflare, Let's Encrypt, GitLab); it is not provider-neutral. Every
project, region, zone, domain, account, token and product name is a configuration input; the examples use
`staging.example.net`, `example.net`, `PROJECT_ID` and `replace-me`.

**Status.** This page describes the tooling and its rules. It records no observed deployment, and the exported tooling
has not been exercised against any provider (Google Cloud, Cloudflare, Let's Encrypt or GitLab) by this export:
provisioning, issuance, lifecycle, sizing and cost are unverified until you run them in your own accounts. Nothing
here performs or authorizes a paid, DNS, registry or secret action; each live step is an explicit operator decision.

## Profile and disclosed gaps

Staging is a bounded synthetic profile, not production readiness
(see the staging and production prerequisites in [SECURITY.md](../SECURITY.md)):

- **Development loader.** `APP_ENV=development`, with `IDENTITY_COOKIE_PROFILE=https` explicit on web and core, real
  public DNS and trusted HTTPS, private backend ports, the default WorkOS adapter and the real `edge-probe` path. No
  identity emulator, fixed credentials, DNS/TLS fixture, test signer, `PLATFORM_HOST_PORT` or
  `IDENTITY_WORKOS_API_BASE_URL`; `migrate down`/`reapply` are never exposed.
- **Production result stays blocked.** The SECURITY.md production prerequisites remain open: no verified TLS from core and migrate
  to PostgreSQL or from core to Redis; plaintext internal hops (edge to web/core, web to core, core to intelligence) on
  private Docker networks of one VM; no at-rest evidence for PostgreSQL, object storage or backup sets; no operator
  deprovisioning entry point (F-6). Findings F-1 to F-4 stay disclosed local exceptions. The readiness evaluator
  result is unchanged.
- **Signed downloads.** MinIO stays private on the VM; its signing endpoint yields no externally usable URL. External
  signed-download delivery is unavailable and unqualified in this environment; no public object-store host exists.
- **Authentication.** With the identity secrets empty, a deployment is infrastructure smoke only: authentication is
  unavailable, never bypassed. Real-provider configuration, the privacy-first sequence and interactive qualification
  are separate identity setup steps ([IDENTITY.md](IDENTITY.md)); this tooling performs none of them.
- **MinIO** is archived upstream (see [VENDORS.md](VENDORS.md)); the planned production store is R2.

## Operator prerequisites

- **Inputs (supplied by the environment owner):** a GCP project with linked billing; region and zone; an environment
  name (`^[a-z][a-z0-9-]{1,20}$`); the Cloudflare zone holding the base domain `B` (not registered or purchased by this
  tooling) and two zone-scoped tokens with `Zone:Read` and `DNS:Edit` (one for Terraform, one for Caddy); an ACME
  contact address; a cost ceiling.
- **Operator machine:** Docker (Terraform runs in its pinned image), the gcloud CLI after `gcloud auth login` and
  `gcloud auth application-default login`, and permission to impersonate the deployer service account. No
  service-account key is created or used. Terraform's Cloudflare token comes only from `CLOUDFLARE_API_TOKEN` in the
  operator's environment, never from a variable or file. Every command that plans the whole root needs it: `provision`,
  `status`, `retire` and a `start` that changes the session length (below); `start` with an unchanged length does not.
- **Private configuration:** `staging.conf`, `terraform.tfvars` and `backend.hcl` live in an operator-private directory
  (mode 0600, outside Git) and hold non-secret identifiers only. Templates:
  `infra/staging/examples/staging.conf.example`, `infra/staging/examples/app.env.example`,
  `infra/terraform/environments/example/{terraform.tfvars,backend.hcl}.example`.

## Configuration defaults

| Setting | Default | Where | Notes |
|---|---|---|---|
| Session length (`session_max_seconds`, `SESSION_MAX_SECONDS`) | 7200 (2 h) | `staging.conf` (optional); per session `--duration` on `provision` (first allocation) and `start` | native `max_run_duration` with `STOP`; 600 to 28800 whole seconds (10 min to 8 h), never clamped; Terraform sets it at creation, then `start` changes it only while the VM is stopped (`gcloud ... set-scheduling`, then a refresh-only Terraform apply) |
| Termination action | `STOP` | Terraform variable validation | `DELETE` is refused |
| Machine type | `e2-standard-2` (non-Spot) | `terraform.tfvars` | sizing is unmeasured until a pilot records it |
| Disks | 20 GB boot + 50 GB data, `pd-balanced` | `terraform.tfvars` | the data disk holds `/var/lib/docker` (images and named volumes) |
| DNS TTL (`dns_ttl`) | 300 | `terraform.tfvars` | the address is static |
| HSTS | `max-age=300`, no `includeSubDomains`, no `preload` | edge configuration (`infra/edge/`) | short on purpose for a staging base domain |
| ACME directory (`EDGE_ACME_CA`) | Let's Encrypt production | rendered env | the Let's Encrypt staging directory for the rehearsal |
| Root host (`DEFAULT_TENANT_SLUG`, `PLATFORM_ROOT_PAGE`) | unset (core's defaults) | `staging.conf` (optional) | passed unchanged to core only; slug empty or `^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`, page empty, `signup_redirect` or `landing`; semantics in [DOMAINS.md](DOMAINS.md) |
| MCP (`MCP_CONNECT_ISSUER`, `MCP_ALLOWED_ORIGINS`) | unset (MCP disabled: no route, `/account` reports connections unavailable) | `staging.conf` (optional) | passed unchanged to core only; issuer empty or `https://<host>` (the environment's AuthKit Connect issuer, a public value; no path or trailing slash), origins empty or comma-separated exact https origins and only with an issuer; semantics in [MCP.md](MCP.md). Endpoint availability is not client enrollment: the WorkOS environment must also allow the client's registration |

## Lifecycle

`bash infra/staging/staging.sh <command> --config <staging.conf> [flags]`; every command accepts `--dry-run`, which
prints the commands without executing them. One-time setup is `infra/terraform/bootstrap/bootstrap.sh` (operator
credentials; idempotent; prints a receipt of created and existing names, never a value).

| Command | Effect | Safety |
|---|---|---|
| `bootstrap` (`bootstrap.sh`) | APIs, state bucket, VM and deployer service accounts with their bindings, Artifact Registry repository, Secret Manager containers and generated application secret versions | never prints values; refuses a project that differs from `staging.conf` |
| `build --candidate <sha>`, `push` | builds the candidate's images off the VM, pushes them and writes the deploy manifest (digests, file hashes, secret version numbers) | one build at a time; no value in the manifest |
| `provision --candidate <sha> [--duration D] [--yes]` | Terraform `plan` (saved privately, shown) then `apply`, then `deploy` | `--yes` required; a re-run converges; refuses a plan touching unlabelled resources; an existing VM keeps its session length and refuses `--duration` |
| `deploy --candidate <sha>` | uploads the candidate's `infra/` subtree and manifest, renders the env, pulls, migrates up, starts services, verifies readiness | never `down`, `migrate down` or `reapply`; seeds only a new data volume, or one an interrupted first deploy left without its marker and without application data; no Terraform, VM start or stop, or scheduling change: a running session's timer is untouched |
| `start [--duration D]` | sets the session length of the stopped VM when it differs (below), starts the VM, holds app and edge, checks stores and the dataset marker, then starts app and edge; prints and records the scheduled stop | holds on any recovery failure (below); refuses a different length while the VM runs |
| `stop` | `docker compose stop` (when it fails or the env cannot be rendered, every container is stopped by its Docker labels instead; the VM-side stop fails, without `STOPPED`, only when that fails too), then stops the VM; writes a receipt | keeps database, objects and certificates |
| `reset --confirm <env>` | removes only the `<env>_postgres-data` and `<env>_minio-data` volumes, re-initializes roles, migrates and reseeds | the only data-discarding path; never touches `edge-data` |
| `retire [--yes] [--purge-images]` | stops, then Terraform destroy of the VM, disks, address, firewall, network and DNS records; verifies absence | `--yes` required; bootstrap-owned items are kept and listed |
| `status`, `receipt` | read-only state: VM, session (configured length, state, scheduled stop), containers and digests, marker, certificate, DNS, cost line; `receipt` writes the all-category receipt with a `vm-session` row | read-only |

**CI-enabled environments.** When the environment's CI metadata bucket exists, `provision`, `deploy`, `start`,
`stop`, `reset` and `retire` take the environment lease and `deploy` writes a schema-fence entry before the release
upload; without the bucket every command behaves as described here. See
[Manual CI start/deploy and stop](#manual-ci-startdeploy-and-stop).

**Several environments in one project.** Environment names may extend one another (`staging`, `staging-eu`), so the
tooling never matches an `<env>-` prefix: reconciliation, the plan check and retirement's absence check use the
environment's exact object names (`<env>-vm`, `-data`, `-ip`, `-edge`, `-iap`, `-vpc`, `-subnet`) or its exact
`environment` label, and the receipt lists only Secret Manager IDs `<env>-<NAME>` with an upper-case `NAME`. A stray
labelled object of the environment is still reported; an unlabelled object under another name is not attributed to it.

**Session length.** The VM carries Compute Engine's native run-duration limit (`max_run_duration`) with action `STOP`;
it is the only timer (no Cloud Scheduler, cron or separate shutdown job). Compute Engine arms the deadline at each VM
start and clears it at stop, so manual `stop` remains available and both stops keep the retained baseline. The
operator picks a length per session with `--duration D`, where `D` is `<n>m`, `<n>h` or whole seconds (`90m`, `2h`,
`7200`) from 600 (10 min) to 28800 (8 h); anything else exits 2 before any command, and Terraform's variable
validation enforces the same range. Without `--duration` the session takes the configured default
(`SESSION_MAX_SECONDS` in `staging.conf`, else 7200), never a previous session's length. Only `provision` and `start`
accept `--duration`; `--override session_max_seconds=...` is refused.

- **First allocation** (`provision` with no VM): `--duration` or the default rides the one apply; the VM is created
  running, so its deadline is armed then. An **existing VM's** `provision` is infrastructure maintenance (partial
  allocation repair included): every plan passes the length the instance reports, it never starts a stopped VM, and
  it refuses `--duration` (use `start --duration`, after `stop` when the VM runs). `status` and `retire` also plan
  with the observed length, so no other command proposes a scheduling change. "No VM" is declared only on positive
  confirmation (the describe fails and an exact-name instance list succeeds empty); a VM whose state cannot be read,
  or an existing VM whose session length cannot be read or is missing or malformed (not whole positive seconds),
  HOLDs (exit 3) with nothing planned, applied or started and no default inferred (the receipt records an unreadable
  VM as unreadable, never absent). Only a confirmed-absent VM uses the configured default.
- **Who writes the length.** The provider (hashicorp/google 8.8.1) replaces the VM when `max_run_duration` changes,
  so the module sets it from `session_max_seconds` only when the VM is created and then ignores it
  (`lifecycle { ignore_changes = [scheduling[0].max_run_duration] }`): `-var session_max_seconds` no longer changes
  an existing VM. Its writer is `start`, through `gcloud compute instances set-scheduling` followed by a refresh-only
  Terraform apply that records the live value (a deliberate exception to the single-manager rule).
- **`start`** reads the instance's `scheduling.maxRunDuration.seconds` and Terraform's `max_run_duration_seconds`
  output (either unreadable or malformed: HOLD, exit 3). A running VM with the requested length gets the resume checks
  only (the deadline is not re-armed; the message shows the scheduled stop); a different length on a running VM is
  refused (exit 2; `start` never stops the VM). A stopped VM with the requested length is started with no Terraform
  write. A stopped VM with another length needs `CLOUDFLARE_API_TOKEN` (the refresh plans the whole root, the DNS
  records included; without it `start` exits 2 before any change) and a `TERMINATED` state read just before the
  change. Then, each step stopping on failure: (1) `gcloud compute instances set-scheduling <env>-vm
  --max-run-duration=<n>s --instance-termination-action=STOP --maintenance-policy=MIGRATE --restart-on-failure
  --provisioning-model=STANDARD` (every other scheduling attribute stated as provisioned; never `--clear-*` or
  `--termination-time`); (2) the instance's whole scheduling is read back: `TERMINATED`, the new length, `STOP`,
  `STANDARD`, automatic restart, `MIGRATE`, not preemptible; (3) `terraform plan -refresh-only -detailed-exitcode`
  saved privately (exit 2 expected; exit 0, no drift, or any failure HOLDs) and the environment check; (4) the saved
  plan must propose no resource change and its only drift must be `module.vm.google_compute_instance.vm` updating
  `scheduling[0].max_run_duration[0].seconds` from Terraform's to the requested value, besides the status values
  `current_status` and `cpu_platform`, and the recorded-state lag of a VM never refreshed since its first start (the
  attachment records `users` and `last_attach_timestamp` of the address and data disk, empty-collection defaults on
  the firewall rules and the instance, and the OS Login IAM member's `etag`; each only in the shapes `benign_drift`
  names, an attachment only to this configuration's exact project, zone and VM) (another resource or leaf, a replacement, a delete, another or an
  unknown attribute or a different value HOLDs, plan kept for review); (5) only that plan is applied; (6) the instance and Terraform must
  both report the new length; only then does the VM start. `start` prints and records
  `session <n>s, scheduled stop <terminationTimestamp>` (inventory and output).
- **Failures and recovery.** No failure starts the VM; every HOLD after step (1) prints the partial state (the
  instance's and Terraform's length as read then) and the recovery. Recovery: with the VM stopped, run the same
  `start --duration <n>` again. When the instance already reports `<n>` and only Terraform differs (the change landed
  but the refresh, apply or output did not), `start` runs only steps (2) to (6) and starts; when the instance still
  matches Terraform, it repeats the change. Any other disagreement (the instance at neither Terraform's nor the
  requested length, or a running VM) HOLDs with the manual reconcile steps: keep the instance's length with
  `start --duration <the instance's length>`, or set the stopped instance back with the same `set-scheduling`
  command and Terraform's length (`--max-run-duration=<Terraform's>s`), then `start` again. Nothing overwrites the
  state with a guess or infers a default.

`status` and the receipt show the configured length, the VM state and the scheduled stop (the armed
`terminationTimestamp` while running, `not armed (stopped)` otherwise). Stopping is not retirement: disks, the static
address, DNS records, state, secrets and images remain and keep costing (below).

## Secrets: delivery, retention, rotation and removal

Secret values live in Secret Manager, one secret per name (`<env>-<VAR>`, at most 13: the eight generated application
secrets, Caddy's `CF_API_TOKEN`, the optional `OPENROUTER_API_KEY` and the three identity secrets
`IDENTITY_WORKOS_{CLIENT_ID,ISSUER,API_KEY}`, empty until the identity setup adds versions). They never enter Git, Terraform
variables, state or outputs, instance metadata, startup scripts or published evidence. On the VM, `deploy` and `start`
fetch the pinned versions with the VM service account into a mode-0600 env file on tmpfs, run Compose, and delete the
file at the end of the run; any later Compose call reconstructs it the same way and deletes it afterwards.

**Token format.** `CF_API_TOKEN` must be a Cloudflare API token: either exactly 40 characters of `A-Z`, `a-z`, `0-9`,
`_` and `-` (the earlier form), or the current prefixed form, `cfut_` (user token) or `cfat_` (account token) then exactly 48
characters of `A-Z`, `a-z` and `0-9`; any other prefix is refused because the pinned Cloudflare DNS plugin rejects it and then repeats the value. Every fetch (`deploy` and the pinned versions of `start`, `stop` and `reset`) checks the format before
the env file is rendered and refuses otherwise with a fixed message that carries no value, length or prefix, so a
malformed token never reaches Caddy: the edge's Cloudflare DNS plugin repeats the token it was given in its
provisioning error, and the access-log filters do not redact configuration errors. The owner enters this version
directly with `gcloud secrets versions add`; no tool of this repository receives it before the VM fetches it.

**Where injected values remain.** Secret values exist on the VM in three places: the tmpfs env file while a run uses
it; the running processes' environment; and each container's configuration under
`/var/lib/docker/containers/<id>/config.v2.json` on the retained data disk, for as long as the container object
exists, stopped containers included. Deleting the env file revokes nothing in the last two. A stop receipt therefore
states: "container configuration with injected environment retained on `<env>-data`, encrypted, VM-only".

**Access.** The data disk is encrypted at rest with Google-managed keys, attached only to the VM, and never
snapshotted or imaged by this tooling; the VM service account holds no disk, snapshot or storage role. Interactive
access is OS Login over IAP for the configured operator principals only (no SSH keys, no external port 22). The
principals who can read retained values are not only the Secret Manager readers: they include those operator
principals and any project, organization or platform administrator able to grant themselves OS Login, attach or
snapshot the disk or read Secret Manager, and Google as the platform operator.

**Metadata server.** The VM service account can read every staging secret version and the Artifact Registry
repository, and anything that reaches the metadata server (`169.254.169.254`, `fd20:ce::254`,
`metadata.google.internal`) can obtain its token. The token is a property of the VM, not of a process, so narrowing
the account's roles in Terraform cannot keep it from the containers. The host scripts use it; no container needs it.
The startup script therefore installs `staging-metadata-guard.service` before Docker: a oneshot unit, ordered before
`docker.service`, which requires it (Docker never starts without it), and which inserts DROP rules for both metadata
addresses into the raw table's `PREROUTING` chain (TCP and UDP except port 53: Docker's embedded resolver forwards a
container's DNS queries to the host's resolver, which is the metadata address, from the container's namespace). That
chain sees only packets arriving on an interface, which is every container network's traffic, before routing, NAT and
Docker's chains; the host's own requests leave through `OUTPUT` and are untouched. On the VM, `metadata_denied`
then asks for the token from a throwaway container on each of the project's Compose networks, for all three
addresses, before app and edge start and again before READY. The probe fails closed: each network first runs a tool
control (the same `wget` against a closed loopback port must exit exactly 1, so a crash such as BusyBox 1.35's exit 139
on `wget -T` cannot read as a denial), and a probe is denied only with an empty answer and status 1, 124 or 143 (refused,
unreachable or timed out under `timeout 5`; the guard drops silently). Any answer, any other status, or a probe that
cannot run HOLDs with app and edge stopped; the answer is only tested for emptiness, never printed or kept. Each probe
container has a unique name, the labels `g070.task=G070`, `g070.env`, `g070.probe` and `g070.probe.run`, a `--cidfile`
and no `--rm`, so its exact container ID and the immutable image ID are in the inventory
`/srv/<env>/state/probe-containers.tsv` (append-only: `allocate`, `created`, `removed`, `reaped`, `remove-failed`,
`held`, `stale-closed` rows, repeated as `probe-container` and `probe-reap` receipt steps) before anything removes it,
and removal is an explicit, verified step. At the start of every check (so after an interruption, a reboot or a
`start`), at its end and from the EXIT trap, `probe-reap` finds this environment's probe containers by their labels and
removes one only when its name has an `allocate` record whose run label, image ID and container ID all match what the
daemon reports; anything else is `held`, left untouched and fails the check (HOLD), and a record whose container no
longer exists is closed as stale. A cidfile is deleted only after its container's absence is verified (a held container
keeps it as evidence), and a stale record is closed only after its cidfile is verified gone (a failed unlink keeps the
record open, the check unresolved and the recovery retryable). The check fails closed: while that reconciliation is unresolved, or when an inventory row cannot
be written, no probe container is started (a failed `created` row removes exactly the recorded container). A probe
container that cannot be proven created, recorded or removed stops the check at once: no later endpoint or network is
probed, the exact-owner reconciliation runs again, and readiness HOLDs. Other resources are never touched.

- **Rotation:** add a Secret Manager version, `deploy --candidate <same sha>` (Compose recreates the affected
  containers and removes the replaced ones), then disable the superseded version. Verify by the changed
  `com.docker.compose.config-hash` label and container creation time, never by reading the environment. Database role
  passwords also follow the existing `ALTER ROLE` procedure ([SECURITY.md](../SECURITY.md)).
- **Removal:** a value leaves the VM only when its container objects are removed: a `deploy` after rotation, `reset`
  (removes every container; volumes kept) or `retire` (disk destroyed). Disabling or destroying a version stops new
  fetches, not already injected values; restore the previous runtime configuration with `deploy` first.
- **Evidence:** no `docker inspect` without a `--format` that excludes `.Config`/`.Config.Env`, no captured
  `docker compose config` outside `--dry-run`, never `/proc/*/environ` or `config.v2.json`.

## Recovery after an abrupt stop

Services run with `restart: unless-stopped`, so after a native expiry, host maintenance or reboot Docker may resume
containers before an operator runs `start`. That autonomous restart is not recovery evidence. `start` first stops
the app and edge services by their Docker labels (no secret is needed for that) and only then fetches the pinned
secret versions, so a slow or failed fetch leaves them stopped; it then brings up the stores, waits for PostgreSQL,
runs `migrate status` and verifies the dataset marker; only then does it start app and edge and report ready. If PostgreSQL recovery fails, the migration
status is inconsistent or the marker is absent on an existing dataset, `start` leaves app and edge stopped, writes a
HOLD receipt and the operator escalates. Nothing is reseeded: `deploy` seeds when the `<env>_postgres-data` volume
did not exist before its first `up`, never touches a dataset that has its marker, and seeds an existing volume without
its marker again only while no table of the database holds a row (goose's version table aside): a first deploy
interrupted after the volume was created (`migrate up` or the seed failed) therefore converges when re-run with the
same candidate. Any application row, or a count that cannot be read, HOLDs instead, and only `reset --confirm <env>`,
the only data-discarding path, recovers it.

## DNS and TLS

- **Records** (Terraform, Cloudflare, DNS-only, no proxying): `B`, `*.B` and `custom.B` as A records to the reserved
  static address. Caddy alone writes and removes `_acme-challenge` TXT records. A tenant's `T` CNAME and
  `_platform-verify.T` TXT belong to the custom-domain flow ([DOMAINS.md](DOMAINS.md)).
- **Certificates:** Caddy obtains `{B, *.B}` by DNS-01 with the pinned Cloudflare module and its zone-scoped token
  (an environment reference in the Caddyfile, never a literal). Custom hosts use on-demand TLS only in the catch-all
  site, permitted by core's internal `tls-ask` endpoint ([DOMAINS.md](DOMAINS.md), edge deployment inputs).
  A first issuance is rehearsed against the Let's Encrypt staging directory on a separate volume. Readiness is proven
  per name, because Caddy may serve the exact and the wildcard name from distinct certificates: `B` and a random label
  host each complete a request through the edge with a verified chain, `B`'s certificate lists `B` and the label's
  lists `*.B` among its subjectAltNames, and each stays valid for at least 7 more days.
- **Storage and limits:** the ACME account and certificates live in the `edge-data` volume, kept across `stop`,
  `start`, `deploy` and `reset`; only `retire` deletes it. This protects the Let's Encrypt limit of 5 duplicate
  certificates per 7 days. Caddy renews about 30 days before expiry; a certificate that expired while stopped is
  reissued at `start` before HTTPS is reported ready.
- **Edge behaviour:** exact `Host` passthrough with `X-Forwarded-*` removed upstream; `/mcp/c/*` and
  `/.well-known/oauth-protected-resource/mcp/c/*` to core, everything else to web (internal routes unreachable); the
  admin API off; the access log redacts the whole query and filters `Cookie`, `Authorization`, `Set-Cookie` and
  `Location`.

## Trust boundaries

| Boundary | Rule |
|---|---|
| Internet to VM | tcp 80/443 to the edge only; tcp 22 only from the IAP range `35.235.240.0/20` with OS Login; every other service has no published port |
| Edge to web/core | plaintext HTTP on private Docker networks (disclosed above) |
| Human operator | short-lived credentials; `roles/iam.serviceAccountTokenCreator` on the deployer only |
| Deployer service account (impersonated) | `compute.instanceAdmin.v1`, `compute.networkAdmin`, `compute.securityAdmin`, `iap.tunnelResourceAccessor`, `compute.osAdminLogin` on the project; `iam.serviceAccountUser` on the VM account; `storage.objectAdmin` on the state bucket; `artifactregistry.writer` on the repository; no Secret Manager or IAM-admin role |
| VM service account | `secretmanager.secretAccessor` per staging secret; `artifactregistry.reader` on the repository; nothing else; its token is reachable from the host only (`staging-metadata-guard`) |
| Containers to the metadata server | dropped in raw `PREROUTING` (TCP/UDP except DNS) before Docker starts; proven from every Compose network before READY |
| Never granted | Owner, Editor, `resourcemanager.projectIamAdmin`, `iam.serviceAccountKeyAdmin`, service-account keys |
| Administrators | project, organization and platform administrators can reach secrets, disks and the VM regardless of these roles; they are inside the trust boundary |
| Terraform state | remote GCS bucket (versioned, uniform access, public access prevented, native locking); holds no secret value |
| Cloudflare | two zone-scoped tokens; Terraform's from the operator environment, Caddy's from Secret Manager |

## Receipts, retention and retirement

`receipt` lists every category with count, owner, reason and release condition: VM, boot and data disks, static
address, DNS records, `edge-data` and `edge-config`, Terraform state, Secret Manager secrets (names and version
counts), Artifact Registry images, service accounts, local Docker objects, worktrees and private paths. One deployment
receipt identity, `<env>-<candidate sha12>-<n>`, identifies a deployment in later records.

The default after a session is the **retained stopped baseline**, recorded as retained (owner: the environment owner;
reason: repeat sessions; release: an observed `retire` or an owner decision), never reported absent. `retire` verifies absence (label queries, Cloudflare
reads, empty state) and names the bootstrap-retained set (state bucket, secrets, service accounts, repository and,
unless `--purge-images`, images). Bootstrap-owned items have their own removal step in `bootstrap.sh`, which refuses
while any environment state object exists.

## Costs [unverified]

Planning figures for an example region, from third-party and announcement sources. **Every figure is [unverified]**
until you read the primary Google pricing pages for your region and confirm your account and cost ceiling.

| Item | Running 2 h session (default) | Stopped baseline / month | Retired / month |
|---|---|---|---|
| `e2-standard-2` on demand (≈ $0.067/h) | ≈ $0.13 | $0 | $0 |
| External IPv4, in use / reserved idle | ≈ $0.01 | ≤ ≈ $7.30 | $0 |
| 70 GB `pd-balanced` (≈ $0.10/GB-month) | — | ≈ $7.00 | $0 |
| Secret Manager (≤ 13 secrets, ≤ 10 active versions) | — | ≈ $0.60 | ≈ $0.60 |
| Artifact Registry (≈ 2 GB) | — | ≈ $0.15 | ≈ $0.15 (0 with `--purge-images`) |
| GCS state, Cloudflare DNS, Let's Encrypt | ≈ $0 | ≈ $0 | ≈ $0 |
| **Total** | **≈ $0.15** | **≈ $11–15** | **≈ $1** |

The running cost scales with the chosen length: about half the former 4 h figure at the 2 h default, about $0.60 at the
8 h ceiling. A VM left running would cost about $60 a month; the native expiry bounds a forgotten VM to one session
length (2 h by default, 8 h at most).

## Offline checks

`infra/staging/check.sh` (operator-run, not a harness suite) checks the tooling without a cloud account: Terraform
formatting, validation and mocked-provider module tests against a prepared provider mirror; `caddy adapt --validate`
and a log-redaction sentinel run in the edge image; `--dry-run` of every lifecycle command and the root host settings
through the real parsers; the exact-name matching with two environments in one project and the quoted name filters;
the session-length interface (parser, refusals, the `start`/`provision` decisions on faked instance and Terraform
state, the set-scheduling and refresh-only sequence with its failure and recovery cases, the refresh scope check on
fixtures and real `terraform show -json` documents) and its default and bounds pins; static guards; the Compose
overlay invariants; the `start` stop-before-fetch order, the `stop` label fallback and the `deploy` seeding rule; the
per-name certificate proof on fixtures; the `CF_API_TOKEN` format refusal with canaries; the metadata guard, statically
and packet-level in an isolated network namespace; ShellCheck and Hadolint on the new files. `prepare` is its only
networked mode; `run` writes a ledger with command, exit code and input digest per row. Its lease `forbid()` comparison
reads a fixed earlier revision of `infra/staging/lib.sh` from the hash-pinned fixture
`infra/staging/tests/fixtures/baseline/` (no Git history needed); a missing or altered fixture fails the row.

## Manual CI start/deploy and stop

The action dropdown uses the exact labels below (including spaces); API callers pass the same strings.

Two manual GitLab CI actions, **Deploy environment** and **Stop environment**, for registered environments, sit on top
of the lifecycle tooling above, which they run as trusted control code. The local operator path stays available under
the same environment ownership rules (lease below).

**Status: nothing here has been performed by this export.** The GitLab settings, workload identity federation and
IAM, the preview GCP project, the metadata buckets and every live deploy or stop are separate setup steps in your own
accounts. This section describes the interface and its rules; no step in it has been carried out against a provider.
The CI integration implements no scheduler, expiry controller, reset, reprovision or retirement job.

### Environments and slots

The committed registry `infra/staging/ci-environments.txt` lists the selectable environments, one per line: `staging`
and `preview-1`. It must equal the `environment` input's options in `.gitlab-ci.yml` (B-CI-CONFIG asserts it). Each
environment has its own GCP project (one project per preview slot), and in it its own bootstrap (state bucket,
deployer, registry, secrets), VM, stores, data and certificate volumes, DNS names, Terraform state prefix, metadata
bucket, federation provider and CI identities, plus its own GitLab protected environment, profile variables and
resource group. No preview identity holds any role in the staging project. The adapter acts only when the selector,
the registry entry, `CI_ENVIRONMENT_NAME` and the profile's `STAGING_ENV` are exactly equal (`preview-1` is never
`preview-10`, `preview-01` or `Preview-1`). Examples use generic names such as `preview-1.example.net`.

**Later slots** (`preview-2`, `preview-3`, lower case): one registry line and the matching input option, in one
merge request; no job logic is copied. The slot's resources (project, bootstrap and provision, federation,
GitLab environment and variables, lineage initialization) each need their own release; until they exist its jobs
refuse.

### Pinning your GitLab project

The launcher (`infra/ci/staging-action.sh`), the hosted adapter's identity-staging gate and the federation bootstrap
(`infra/terraform/bootstrap/ci-federation.sh`) accept exactly one GitLab project, pinned in `infra/ci/hosted-context.sh`
by two lines. The template ships synthetic placeholders (project `999999999` at `example-group/example-project/platform`),
so every real project is refused (exit 16) until you re-pin them:

1. Set `STG_PROJECT_ID=<your numeric project ID>` (Settings, General).
2. Set `STG_PROJECT_PATH_SHA=` to the output of `printf '%s' '<namespace>/<project>' | sha256sum` for your project's
   full path; the path itself is never committed.
3. Keep each line exactly in the form `STG_PROJECT_ID=<digits>` and `STG_PROJECT_PATH_SHA=<64 hex>`: the launcher,
   the bootstrap and the offline tests parse those exact forms. The refusal logic itself needs no change.
4. Update the identity-staging harness controls in `infra/ci/validate-controls.sh` to the same project (`g_stg_path`
   returns the project's full path; `g_stg_env` sets the project ID), or C-SEC-8 reports that the
   derived path does not match the pin.

### Setup order (each step an explicit operator action; none performed by this export)

| Step | Action |
|---|---|
| 1 | Read-only readiness: GitLab tier and features (typed root-config inputs with options/regex and interpolation in rules, `environment`, `resource_group` and `id_tokens`; protected environments; ID-token claims including `environment_protected`; `artifacts:access`; the artifact size limit; a readable pipeline-variable restriction), `staging` branch protection, and the project pins above. Anything unavailable needs a reviewed design change; there is no pipeline-variable fallback. |
| 2 | GitLab settings: protected environments `staging` and `preview-1` (deploy: maintainers/owner); four protected file variables scoped to each environment (below); "minimum role to use pipeline variables" set to `no_one_allowed`, so no variable can override a profile |
| 3 | The preview-1 GCP project, its billing link and a monthly budget alert |
| 4 | Preview-1 bootstrap, owner-entered Caddy and Terraform Cloudflare tokens, the first build/push/provision of the staging tip through the local lifecycle path (outside CI), its production certificate, then stop. No metadata bucket exists yet, so no lease or fence |
| 5 | `infra/terraform/bootstrap/ci-federation.sh` for preview-1 (its `--dry-run` reviewed first): pool and provider with three audiences, `preview-1-ci-stop`, `preview-1-ci-read`, the metadata bucket, bindings and condition; read-back; then `ci.sh lineage init` from the step 4 candidate |
| 6 | Two consecutive `stop` pipelines for preview-1: the first federation, with denial probes and read-backs |
| 7 | Resource-group process mode `oldest_first` for `staging-env-*` and `staging-image-build` |
| 8 | Preview-1: deploy of a same-repository branch with a custom duration, stop, deploy of the same SHA (automatic reuse), stop |
| 9 | Staging: federation, metadata bucket and `lineage init` from the retained baseline's candidate; an idempotent stop; both environments serving at once, stopping preview leaves staging serving |

### Trusted profile per environment

Four GitLab **file** variables, protected and scoped to exactly one environment name:

| Variable | Content |
|---|---|
| `STAGING_CI_CONF` | the `staging.conf` without `EVIDENCE_DIR`, `PRIVATE_DIR` and `GOOGLE_ADC_FILE`; `SESSION_MAX_SECONDS` must be 7200 (the adapter refuses another value) |
| `STAGING_CI_BACKEND` | `backend.hcl` |
| `STAGING_CI_TFVARS` | `terraform.tfvars`, byte-equal to the local operator's, so no plan drifts between CI and local runs |
| `STAGING_CI_SETTINGS` | CI-only keys: `CI_WIF_PROVIDER`, the required explicit `CI_IDENTITY_PINS`, and during proof windows the owner-listed sibling targets of the denial probes, `CI_DENIAL_PROBES` (template `infra/staging/examples/ci-settings.example`) |

Pipeline inputs never carry a project, state backend, credential scope or path. The check job receives only
`STAGING_CI_CONF` and `STAGING_CI_SETTINGS`.

### Running an action (Play and Stop)

Start a **new pipeline on `staging`** (the web UI's run-pipeline form or the pipelines API) and choose the inputs. Only
web and API pipelines of the protected `staging` branch admit a non-default action; any other branch or source
(push, tag, schedule, merge request, trigger, parent/child) never runs these jobs or obtains a credential.

| Input | Values | Empty / default behaviour |
|---|---|---|
| `action` | `Validate` (default), `Deploy environment`, `Stop environment` | `Validate` = today's ordinary web validation pipeline; no staging job exists |
| `environment` | `none` (default), `staging`, `preview-1` (= `ci-environments.txt`) | `none` with `Deploy environment` or `Stop environment` gives zero jobs, so GitLab refuses to create the pipeline; nothing runs |
| `candidate` | empty (default), full SHA, `staging`, or (preview only) a same-repository branch name | empty or `staging` = the frozen `CI_COMMIT_SHA` of this staging pipeline, **for both environments**. Staging accepts an explicit SHA only on its `--first-parent` history (an intermediate or unmerged commit is refused). `preview-1` also accepts a full SHA reachable from a same-repository `refs/heads/*` branch, or a branch name resolved once in `staging:prepare`; merge-request and fork refs are never fetched. The pinned SHA is used throughout; a later branch move or deletion has no effect. Must be empty for `Stop environment`. |
| `duration` | empty (default), `<n>m`, `<n>h` or whole seconds, 600 to 28800 | empty: a stopped VM starts with **7200 s (2 h)**; a running VM keeps its session (no extension, no restart; the existing scheduled stop is reported). An explicit value that differs from a running session is refused before any write (lifecycle semantics). Must be empty for `Stop environment`. |

**Stop environment** takes only the action and the environment: with a candidate or duration the stop job has no rule (zero
jobs, creation refused) and the launcher refuses again (exit 2) before any token file exists. Stop is always its own
new pipeline: GitLab's `on_stop` pairing is not used, so stop never depends on an earlier pipeline, workspace,
artifact or summary, and works after a failed or canceled start.

| Job | Credentials | Resource group | Timeout | Role |
|---|---|---|---|---|
| `staging:prepare` | none (no `id_tokens`, `environment` or protected variables) | — | 10 min | fetch `refs/heads/*`, resolve and pin the candidate, compatibility gate, bundle of the pinned commit, `candidate.tsv`, candidate schema; artifacts kept 1 day; always runs |
| `staging:check` | ID token, audience `staging-ci:<env>:read`, as `<env>-ci-read`; `environment` action `prepare` | none | 10 min | the automatic image-reuse decision (below); a non-secret dotenv artifact `STAGING_REUSE=yes\|no` |
| `staging:build` | none | `staging-image-build` (one build at a time, both environments) | 45 min | `STAGING_REUSE=no`: builds the seven images from the pinned candidate with control code; `yes`: writes a skipped marker |
| `staging:deploy` | ID token, audience `staging-ci:<env>:start`, as `<env>-deployer`; `environment` action `start` | `staging-env-<env>` | 60 min | the deploy sequence below |
| `staging:stop` | ID token, audience `staging-ci:<env>:stop`, as `<env>-ci-stop`; `environment` action `stop` | `staging-env-<env>` | 30 min | records the stop request, then lifecycle `stop` |

The jobs are ordinary (not manual, never allowed to fail): a failure fails the pipeline and nothing unselected waits.

**Deploy sequence.** Phase A, local only (no cloud call): launcher context gate, input syntax, registry/selector/profile
identity, profile and CI keys, identity pins present and not `latest`, duration bounds, the build artifacts against
the bundle and the recorded selection, compatibility gate and candidate schema from the bundle. Phase B,
authenticated and read-only: credential exchange, then reads of the metadata (current state, lineage, fence, lease,
stop requests, manifest) and the VM state; gates: supersession, lease free (or an expired CI lease), schema lineage,
a valid manifest or build artifacts, and for a running VM the duration rule and at least 300 s until the scheduled
stop. Phase C, writes only after every gate: lease, OS Login key with a time limit, images (reuse, or load, verify
image IDs, ID-preserving retag, push, upload the manifest), the Cloudflare token only if a stopped VM's length must
change, lifecycle `start`, lifecycle `deploy` (fence entry inside), readiness, observed-lineage verification, metadata, receipt,
lease release. A deploy succeeds only when readiness reports READY and the running digests equal the manifest.
**Pending-start rule:** a stop pipeline records `stop-requests/<pipeline-id>` before stopping, and a deploy refuses
(exit 2, no write) when a recorded stop has a higher pipeline ID than its own, so an earlier queued start never
restarts the environment after a later stop.

**Stop sequence.** Gate and profile checks with empty candidate and duration; lease and VM state reads; the stop
request record (written even when the only refusal is a live local lease); lease; OS Login key; lifecycle `stop` (graceful
VM-side stop, then the provider stop, waiting for an observed `TERMINATED`; an already stopped VM is an idempotent
success once the describe confirms it); receipt (the Terraform-derived rows the stop identity cannot read are cited
from the last deploy receipt); metadata; lease release. Database, objects and certificates are kept.

### Reading the effective SHA and the scheduled stop

The job log and the only deploy/stop artifact, `staging-summary/summary.txt` (developer access, 90 days), carry a
fixed-field summary: action, environment, candidate SHA and selection text, branch (if any), actor, image digests,
session length, the scheduled stop read back with `gcloud ... describe` after the action, readiness, receipt ID and
sha256, open schema-fence entries and the exit class. The raw output of the lifecycle commands (Terraform plans, VM logs,
command echoes) goes only to `raw/` in the private job directory, which is removed at exit; the job log, the summary
and the sanitized copy in the metadata bucket's `logs/` receive only allowlisted lines (`staging:` messages and the
lifecycle key/value kinds) and the summary. A canary hit in raw or filtered output is exit 6 with nothing published.

### Exit classes and failed-job recovery

| Exit | Meaning | Recovery |
|---|---|---|
| 0 | observed success | — |
| 2 | refused before any change (input, gate, compatibility, lineage, supersession, lease, remaining time) | read the refusal in the summary, correct the input or wait for the holder, and run a new pipeline |
| 3 | HOLD (lifecycle semantics, or ambiguous lineage reconciliation) | follow the printed recovery; lineage ambiguity is an owner decision |
| 4 | failed after a change started | the fence stays; deploy the same or a successor candidate, or stop; the next mutation runs lifecycle reconciliation and the remote recovery first |
| 5 | the session expired during the action (the VM is `TERMINATED` after its armed deadline); never success | deploy again (the stopped VM starts a new session) |
| 6 | privacy canary hit, or a denial probe that unexpectedly succeeded; outputs withheld | treat as a security finding and escalate; do not rerun to rediscover it |
| 16 | not an approved CI context (launcher) | run from a web/API pipeline of protected `staging` |

Stop remains available after any failed or canceled start. Cancellation never resets data. A crashed or canceled
job's CI lease expires after the job timeout plus 5 minutes; the next CI job of that environment breaks it with a
recorded `lease-breaks/` entry.

### Automatic image reuse

There is no `images` input. `staging:check` reads the environment's published `deploy/<sha>.tsv` (lifecycle manifest
format): valid fields, candidate equal to the file name, repository equal to the profile's, digest shape, and all seven
digests present in the registry give `STAGING_REUSE=yes` and the build is skipped; anything else gives `no`.
`staging:deploy` decides again authoritatively before any write: a valid manifest is reused with no load or push;
otherwise the build artifacts are required and pushed; neither is exit 2. A disagreement fails closed (a manifest that
disappeared is exit 2; one that appeared is reused and the build discarded). A reuse needs no earlier pipeline's
artifacts. The build lane holds no publishing credential; if the image archives exceed the GitLab artifact limit or
cannot transfer within the measured bound, the design calls for a reviewed change rather than adding a publisher.

### Candidate compatibility: branch and schema

- **Deployment infrastructure.** All credential-bearing and operator-side code (`.gitlab-ci.yml`, launcher, adapter,
  `staging.sh`/`lib.sh`, Terraform) comes from the pipeline's protected `staging` revision. The candidate's
  `infra/compose`, `infra/postgres`, `infra/minio`, `infra/edge`, `infra/staging`, `infra/docker/edge.Dockerfile`
  and `infra/docker/minio.Dockerfile` must be byte-identical to control; otherwise prepare refuses before any
  credential or cloud call: "incompatible candidate: deployment infrastructure differs from trusted control:
  <paths>; integrate it to staging first". Application code (`core`, `web`, `intelligence`, migrations) is the only
  candidate-controlled content, so a candidate cannot replace remote scripts, Compose networks, the metadata guard or
  the Caddy policy. An older staging candidate whose deployment infrastructure differs is refused the same way.
- **Schema lineage.** The metadata holds `lineage.tsv` (`verified` or `assumed`, source candidate, writer, one
  version/blob row per applied migration), `fence.tsv` (one entry per deploy started since the last verified
  lineage) and `schema/<sha>.tsv` (a candidate's migration version-to-blob map, create-only). A candidate passes
  only if, against the lineage and every fence entry, it carries the same blob for each version up to that member's
  maximum and its own maximum is not lower; otherwise exit 2: "schema fence unreconciled since <holder> deployed
  <sha>; deploy that candidate or a successor, or reconcile". The fence entry is written inside lifecycle `deploy` before
  the release upload (so before promote and migrate), by CI and by local `deploy`/`provision` whenever the metadata
  bucket exists; `start` and `reset` add none. Nothing clears a fence on failure or cancellation, and stop never
  touches lineage. Clearers: (1) a successful CI deploy whose observed applied versions equal the candidate's schema
  exactly (a mismatch is exit 4 with the fence kept); (2) `ci.sh lineage reconcile --config FILE` by an operator with
  local lifecycle credentials, the lease held and the VM running: each applied version must have one identical blob across
  every member containing it, else HOLD (exit 3) for an owner decision; data is discarded only by a separately
  authorized `staging.sh reset --confirm <env>`, which no CI job offers; (3) at setup only, `ci.sh lineage init
  --config FILE --candidate SHA` (refuses when `lineage.tsv` exists) records `assumed` lineage from the VM's current
  release candidate, treated like a fence entry until a CI deploy verifies it.

### Metadata bucket

Each environment has a dedicated bucket, `gs://<GCP_PROJECT>-<STAGING_ENV>-ci-meta`, in its own project, created and
torn down by `ci-federation.sh`: versioned, uniform bucket-level access, public-access prevention, 90-day noncurrent
expiry, label `managed-by=ci-federation`. It is the durable record of CI deployments (GitLab artifacts expire and are
never the sole copy), separate from the Terraform state bucket, which Terraform alone reads. Keys: `current.tsv`
(candidate, manifest sha256 and generation, lineage and last log objects, pipeline/job/actor, session length,
scheduled stop, `started_at`/`stopped_at`), `lease`, `lineage.tsv`, `fence.tsv`, `schema/<sha>.tsv`,
`deploy/<sha>.tsv`, `stop-requests/<pipeline-id>`, `runs/<pipeline>-<job>.tsv`, `receipts/<id>.tsv`,
`logs/<pipeline>-<job>.log` (sanitized only) and `lease-breaks/<utc>-<holder>.tsv`. Every write is conditional on the
object generation (`0` for create-only, else the observed one) and every download is verified before use (a mismatch
HOLDs). Grants (single source `infra/terraform/bootstrap/ci-meta-permissions.tsv`):

| Principal | On the metadata bucket | Delete limited to the exact keys `current.tsv`, `lease` | On the state bucket |
|---|---|---|---|
| `<env>-deployer` (CI start; local operator by impersonation) | `storage.objectAdmin`, `storage.legacyBucketReader` | — | the bootstrap `objectAdmin`, unchanged |
| `<env>-ci-stop` | `storage.objectViewer`, `storage.objectCreator`, `storage.legacyBucketReader` | custom role `ciMetaMutable` (`storage.objects.delete`) | none |
| `<env>-ci-read` | `storage.objectViewer`, `storage.legacyBucketReader` | — | none |

Stop therefore creates its records, overwrites `current.tsv` and deletes its own `lease` at the observed generation,
and nothing else; it never reads `fence.tsv` or `lineage.tsv` and lists nothing in the state bucket.

### Environment lease

When an environment's metadata bucket exists, the mutating lifecycle commands `provision`, `deploy`, `start`, `stop`,
`reset` and `retire` take the environment lease, CI and local alike; `build`, `push`, `status` and `receipt` do not
(`--dry-run` prints the lease commands and acquires nothing). The presence check reads the bucket through the
impersonated identity: present and owned by `GCP_PROJECT` makes the lease mandatory; a positive 404 means the
environment is not CI-enabled (one log line, then today's behaviour); anything else (403, error, foreign project)
refuses with exit 2. Acquire creates `lease` only if absent (holder `local:<user>@<host>:<pid>` or
`ci:<pipeline>/<job>`, command, candidate, acquired, expires); when it exists the command refuses with exit 2 naming
holder and times, before any other call. The CI adapter holds one lease across its lifecycle steps
(`STAGING_LEASE_HELD`, adapter-internal; the launcher unsets it from the job environment).

- **Release:** automatic when the acquiring command exits (its EXIT trap), or by the holder with `staging.sh lease
  --config FILE --release`; `--show` prints the lease.
- **Expired CI lease** (TTL = job timeout + 5 min): the next CI job of the environment breaks it, writing a
  `lease-breaks/` record first.
- **Local lease:** never broken by CI or by another command, expired or not.
- **Break-glass** (environment owner only): `staging.sh lease --config FILE --break --confirm <env> --reason TEXT`
  writes a create-only `lease-breaks/` record (operator, reason, the broken lease's content), then deletes `lease`
  at its observed generation.
- **After a crash:** the next mutation runs lifecycle reconciliation (`reconcile`, VM status, the remote recovery) before
  changing anything.

The lease prevents accidental overlap of lifecycle mutators. A principal with deployer-equivalent rights who deliberately
bypasses the tooling is not technically stopped; bucket versioning, `lease-breaks/` records and Cloud Audit Logs make
that attributable. No bypass is a procedure.

### Local and CI handoff

On a CI-enabled environment every covered local command takes and releases the lease itself, so a CI job refuses
(exit 2, no write) while a local command runs, and a local command refuses while a CI job holds the lease. Local
deploys keep their lifecycle semantics: they append a fence entry but are not refused by the CI schema gate, and
`current.tsv` keeps describing the last CI deployment until the next CI deploy verifies (the summary reports open
fence entries). The adapter's local subcommands `ci.sh meta pull|push` copy the metadata (for example a published
manifest) to and from the operator's private `EVIDENCE_DIR`. The local operator keeps lifecycle credentials (owner ADC
impersonating `<env>-deployer`); no key exists.

### Identity, secrets and the DNS boundary

- **Federation** (per environment project, by `ci-federation.sh`; details in `infra/terraform/README.md`): a workload
  identity pool and OIDC provider for the GitLab instance, audiences exactly `staging-ci:<env>:start`, `:stop` and
  `:read`, `attribute.action` mapped from the audience suffix. The attribute condition requires the pinned project and
  namespace IDs, branch `staging`, a protected ref, `ci_config_ref_uri` equal to the project's
  `.gitlab-ci.yml@refs/heads/staging`, source `web` or `api`, environment `<env>` and protected, a GitLab-hosted
  runner and one of the audiences.
- **Principals:** start impersonates the existing `<env>-deployer` (grants unchanged) and alone may read the one
  secret `<env>-TERRAFORM_CLOUDFLARE_API_TOKEN`; stop impersonates `<env>-ci-stop` (instance get/list/stop and the
  related read permissions, OS Login and IAP on the instance, service-account user on `<env>-vm`; no Terraform state,
  Cloudflare, Secret Manager, registry, create or delete); check impersonates `<env>-ci-read` (metadata read and
  registry read on the environment's repository). Never: Owner/Editor, project IAM admin, service-account keys, a
  Secret Manager role on either service account, state access for stop or read. Honest boundary: OS Login admin
  (needed for the lifecycle tooling's graceful `sudo` stop) makes the stop identity root on its own VM, so it can read that
  environment's VM-delivered secrets; it still cannot plan or apply, read state, change DNS, touch the registry,
  create or delete resources or reach another environment. The deployer's broad lifecycle grants are confined to one
  environment because each environment has its own project.
- **Runtime credentials:** the launcher moves the ID token into a mode-0600 file in a private job directory outside
  the project directory (never an artifact), the adapter uses `external_account` credential files and one
  impersonation hop, the Cloudflare Terraform token exists only in the environment of the one `start` that changes a
  stopped VM's length, and the private directory is removed on exit or cancellation. No candidate build or
  application container receives any credential.
- **Denial probes** (non-mutating, every stop job; check and deploy jobs for their own identities): permission tests on
  this environment's state bucket and each owner-listed sibling's state and metadata buckets return nothing; an
  overwrite of the job's own `runs/` record is denied; impersonating the deployer is denied. Any success fails the job
  (exit 6).
- **Identity pins:** every CI deploy passes `--pin-secret IDENTITY_WORKOS_<NAME>=<n|none>` for all three identity
  secrets from `CI_IDENTITY_PINS`; a missing pin or `latest` is exit 2. Other secrets keep the lifecycle tooling's
  default. The staging `CI_IDENTITY_PINS` belong to whoever configures the identity provider for that environment:
  the CI integration never changes them, and pins that cannot be verified keep a staging deployment held. `preview-1`
  uses `none`.
- **Provider-origin readiness:** with `none` pins a preview is infrastructure smoke only and reports authentication as
  unavailable. An authenticated preview needs its own explicitly configured provider application, origins and isolated
  nonproduction credentials; no staging provider secret is reused, and this tooling does not provide them.
- **DNS:** the Terraform and Caddy Cloudflare tokens are zone-scoped, so a shared zone gives no per-record isolation.
  The protection is procedural plus exact checks: `label_check` refuses plan changes to records outside `B`/`*.B`,
  Caddy issues only for its `EDGE_BASE_DOMAIN`, the compatibility gate keeps edge and Compose equal to control, and no
  candidate build or application container receives either token. Separate tokens per environment (independent
  revocation) are an owner choice.
- **Rotation:** there is no service-account key to rotate. Application and Caddy secrets rotate as in the lifecycle tooling (a new
  version, then a deploy of the same candidate); identity secrets change only with their pins; the Terraform
  Cloudflare token follows the lifecycle tooling's owner-entered procedure and is read only for a session-length change; profile
  variables are edited by the owner in GitLab.

### Costs and retention

The cost table above applies **per environment**: each one stopped keeps its disks, address, secrets and images and
keeps costing (the two-hour limit bounds compute only); the metadata bucket holds kilobytes. Planning estimates of
GitLab compute [unverified]: about 16 units for a stop pipeline and 150 to 180 for two deploys and two
stops. `current.tsv` records `started_at`/`stopped_at` per environment (a start suspends the stopped-time clock). Any
retention policy (for example a shared `staging` eligible after 14 consecutive stopped days, previews after 48
stopped hours) is yours to define: **no automatic retirement exists**, and retiring an environment is an owner
decision executed through lifecycle `retire`.

### Disabling and retiring CI for an environment

- **Disable:** remove the environment's protected environment or profile variables in GitLab (federation requires a
  protected environment and the adapter requires the profile), or remove the slot from the registry and the input
  options in one merge request.
- **Retire the CI integration:** `ci-federation.sh --teardown` removes exactly the federation set it created and
  refuses while live metadata objects exist unless `--confirm-meta <env>` is given after the environment's retirement;
  then remove the GitLab environment, variables, resource groups and input option. Without the metadata bucket the
  local lifecycle commands return to their legacy behaviour (no lease).
- **Retire the environment** (VM, disks, address, DNS): lifecycle `retire` and the bootstrap teardown, separately released;
  never a CI action.

### Additive lifecycle interfaces

The CI integration changes the local lifecycle tooling only through **additive interfaces whose defaults preserve today's behaviour**; each is proven
against legacy invocations (LIFECYCLE-DRYRUN unchanged apart from the printed lease lines, ENV-LEASE parity):

1. `staging.sh build --source DIR` builds from a clean detached checkout at the candidate (Compose files, build
   contexts and the edge context under `DIR`); without it, the current checkout as before.
2. The optional `staging.conf` key `GCLOUD_IMPERSONATE_SA`, exactly `<env>-deployer`, `<env>-ci-stop` or
   `<env>-ci-read` `@<GCP_PROJECT>.iam.gserviceaccount.com`; absent, `DEPLOYER_SA` (identical gcloud flags to today).
3. The environment lease on `provision|deploy|start|stop|reset|retire` when the metadata bucket exists, and
   `staging.sh lease --config FILE --show|--release|--break --confirm ENV --reason TEXT`.
4. The schema-fence entry written by `deploy` when that bucket exists.
5. Only if the pre-registered OS Login key proof fails: the optional `STAGING_SSH_KEY_EXPIRE_AFTER`.

No Terraform, remote-script, manifest, receipt or exit-code change.

### Offline proofs

Operator-run, no cloud contact: `infra/staging/tests/ci-federation.sh`, `ci-adapter.sh`, `env-lease.sh` and
`ci-launcher.sh` (check.sh rows `CI-FEDERATION`, `CI-ADAPTER`, `ENV-LEASE`, `CI-LAUNCHER`), and the credential-free
build-boundary proof `infra/staging/tests/ci-build-boundary.sh --work DIR --report DIR` in the build slot: a hostile
candidate whose build steps dump their environment, process environments, `/builds` and `/certs` is built through the
real prepare/build phases with canaries in the job environment, and no canary may appear in any image export, build
log or artifact; a hostile Compose variant must be refused before any build; the baseline build records the artifact
size and image-ID preservation. The `env-lease.sh` parity cases compare against fixed earlier revisions of
`infra/staging/staging.sh` and `lib.sh`, vendored byte-exact and hash-pinned under
`infra/staging/tests/fixtures/baseline/`; a missing or altered fixture fails the run before any case.

## Identity configuration on a retained environment

Adding real identity provider configuration to a retained environment means adding identity secret versions and
deploying; at cleanup, restore the previous runtime identity configuration with `deploy` before disabling the versions
you added (disabling alone does not remove injected values), and remove only what you added.

## Production extension (documented only)

Production would use a separate project and state bucket, always-on application compute, a managed PostgreSQL service
and private R2 object storage, and must first close the SECURITY.md production prerequisites. None of it is
provisioned by this tooling.

Copyright 2026 Vitruvian Technology, Corp.
