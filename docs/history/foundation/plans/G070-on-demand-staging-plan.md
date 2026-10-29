> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G070 — Reusable on-demand SaaStacular staging on GCP: Implementation Plan (r0)

Revision: **r0, 2026-10-26 UTC.** Author: Claude (Fable 5.1, high), single lead planner integrating three read-only
research lanes (GCP Terraform/IAM/state/lifecycle/cost; runtime Compose/images/seeds/secrets; DNS/Caddy/G057 handoff).
**This revision releases nothing: no install, build, Docker object, Terraform/gcloud/provider/DNS call, cloud allocation,
application or infrastructure edit, commit to any other file, push or MR.** Implementation begins only after Codex Plan
ACCEPT and recorded execution release; paid/DNS/provider actions begin only after the separate live release of §4.3.
Budget **B5, checkpoint after attempt 3; zero attempts spent** (§8.4).

Goal: [G070](../goals/G070-on-demand-staging.md) (sha256 `5e4dcc62fc6c8294ea3df391e321b18ca868416fda65d6b01aee7eb8fde7778e`).
Spec: [r0](../specs/G070-on-demand-staging-spec.md) (sha256 `e4e6cbc4d86885bc159949226f7ea8a8a3d50cc89b20c2a7437a899a2bab3a02`).
ADR: [ADR-009](../../../adr/ADR-009-on-demand-gcp-staging.md) (accepted direction; no live release).

## 0. Bindings, verified facts and corrected lane claims

| Binding | Value (verified in this planning checkout, read-only) |
|---|---|
| Staging base | `a5aa46ba80e0ede2837b3673dea1cb4d62c67af6` ("docs(G070): specify reusable GCP staging and release parallel planning") |
| Planning checkout | isolated worktree `agent-<id>`, created at `main` `b40d4a4d`, moved by `git checkout --detach a5aa46ba` (HEAD verified, status clean before and after); this file is its only change |
| Lane reports (evidence, not authority) | `lane1-gcp-terraform.md` sha256 `cd9e7033…c63c`, `lane2-runtime.md` `0593bff5…eea6`, `lane3-dns-caddy-g057.md` `acfb00af…b075` (job directory `…/580cdb2f/tmp/g070/`) |
| G057 inputs | Goal §"G070 reusable environment handoff" (2026-10-26), Spec §"G070 reusable staging dependency" (2026-10-26), accepted Part B Plan r5 (§4, §10–§12, §15.4–§15.6, §16); frozen Part A source `c840d714` (MR !107) untouched |
| Adopting deployment input | `staging.env.foundation.test` — **appears in this Plan, the Goal/Spec and the Result only**; never in any reusable file (§1.1 S0; machine-enforced, item 12 below) |

**Verified facts that drive the design (file:line at `a5aa46ba`):**

| # | Fact | Where |
|---|---|---|
| F1 | No Terraform, Caddyfile, Caddy image, `infra/edge/` or `infra/staging/` exists; no registry/publish/deploy job; no image digest exists for any application image | `find` (none); `.gitlab-ci.yml:6-7`; G057 Result item 7 "left unwritten by design" |
| F2 | Every runtime image is built from source by Compose (`build:` on minio, minio-init, core, web, intelligence, migrate); only postgres `pgvector/pgvector:0.8.7-pg17` and redis `redis:8.10.2-alpine` are pulled; MinIO is compiled from Go source (heaviest build) | `infra/compose/local.yml:10,32,46-50,67-71,87-91,132-136,155-161`; `migrate.yml:10-13`; `infra/docker/minio.Dockerfile:6-9` |
| F3 | Every service publishes a `${BIND_ADDRESS}:<port>` mapping today (postgres 5432, redis 6379, minio 9000/9001, core 8180, web 3000, intelligence 8000); `ports: !reset []` is the accepted override form | `local.yml:22-23,37-38,57-59,114-115,143-144,177-178`; `identity.yml:116` |
| F4 | Networks: web `edge` only, core `edge data inference`, intelligence `inference`, migrate `data`; no service has a restart policy except one-shots (`restart: "no"`); no CPU/memory limits anywhere | `local.yml:29,44,65,85,130,153,184`; `migrate.yml:23-24` |
| F5 | `APP_ENV=production` refuses an `http` object-store endpoint ("OBJECT_STORE_ENDPOINT must be https with APP_ENV=production") and unconfigured identity ("IDENTITY_WORKOS_CLIENT_ID is required in production"); `loopback-http` is refused in production; `PLATFORM_HOST_PORT`, DNS/TLS fixtures, a non-default WorkOS base and the Connect JWKS file are accepted only with `APP_ENV=test`; web's cookie profile defaults to `loopback-http` outside production unless `IDENTITY_COOKIE_PROFILE=https` is set | `core/internal/config/config.go:181-183,268-274,390-396,430-432,441-443`; `domains.go:96-101,175-214`; `web/lib/session.ts:86-93` |
| F6 | The readiness evaluator pins core→PostgreSQL, migrate→PostgreSQL and core→Redis as unsupported for TLS; "Current production result: blocked" — no configuration of this stack is production-ready, so staging never claims it | `core/internal/config/readiness.go:79-89`; `docs/SECURITY.md:326-330` |
| F7 | `migrate down`/`reapply` are refused only with `APP_ENV=production`; core's `/readyz` requires intelligence when `INTELLIGENCE_URL` is set | `core/internal/migrate/runner.go:313-316`; `core/cmd/server/main.go:148-150` |
| F8 | DB roles are created once at first volume init (`<db>_runtime` `NOSUPERUSER NOBYPASSRLS NOINHERIT`); `provision-roles.sql` is a safe repeat verification; the test seed is superuser-over-socket, non-idempotent and unsafe for public use (fixed labels, verified `portal.acme.test` row, principals bound to the emulator issuer) | `infra/postgres/init/10-roles.sh:1-12,36-40`; `infra/postgres/provision-roles.sql:1-18`; `infra/identity/mcp-seed.sql:44-70,111-116` |
| F9 | The edge contract: `GET /internal/v1/domains/tls-ask?domain=` → 200/403/503, no body; the test proxy passes `Host` unchanged and routes only `/mcp/c/` and `/.well-known/oauth-protected-resource/mcp/c/` to core, everything else to web; no component reads `X-Forwarded-*`; the access log must redact the **entire** query on `/auth/callback` and `/signup/{continue,prove,complete}` | `core/internal/domains/api/http.go:34,221-240`; `infra/identity/proxy.conf:34-57`; `docs/foundation/DOMAINS.md:180-196`; `docs/SECURITY.md:316` |
| F10 | B-COMPOSE asserts compose combinations, the migrate env/network/ports, a per-secret service matrix and web's `edge`-only network; it deletes the secret-bearing config JSON | `infra/validation/run.sh:700-745` |
| F11 | `lint_shellcheck` and `lint_hadolint` scan **explicit file lists** (`run.sh:331-345`); `B-PROTECTED` requires every `infra/docker/*.Dockerfile` `FROM` to carry a tag and flags `image:` lines without a colon in `infra/compose/*.yml` | `run.sh:331-345,866-868` |
| F12 | **B-LITERALS scans `web core intelligence contracts infra docs/foundation README.md .gitignore` for the manifest's identity literals** (product name, `foundation.test`, …): the genericity rule is machine-enforced for every reusable path; `management/` is outside the scan | `run.sh:838-843` |
| F13 | `infra/` is a `foundation` path class (hosted optional gate selects `full`); hosted MR `validate:checks` runs the fast profile; no push starts a pipeline | `infra/ci/profile-paths.tsv` (`foundation dir infra/`); `.gitlab-ci.yml:3-12` |
| F14 | `docs/foundation/VENDORS.md` and `infra/compose/local.yml` are rendered archetype paths (VENDORS.md registered as a customization, register row 26); README.md, DOMAINS.md, OPERATIONS.md are not rendered | `.archetype/outputs/14af57049c4d995d/manifest.json`; `infra/validation/render-customizations.tsv:26` |
| F15 | Task-private files: `<main checkout>/.local/{evidence,secrets,tmp}/` (ignored `/.local/`; excluded from every build context) | `.gitignore:14`; `.dockerignore:3`; G057 Plan §10, G069 Plan §6.1 |
| F16 | Provider doc (fetched 2026-10-26): `google_compute_instance.scheduling` supports `provisioning_model` (STANDARD…), `max_run_duration`, `instance_termination_action` (STOP/DELETE), `on_instance_stop_action`; **`graceful_shutdown` is Beta only**; `desired_status` ∈ RUNNING/SUSPENDED/TERMINATED is acted on at apply; `allow_stopping_for_update` lists machine_type/min_cpu_platform/service_account/shielded config/enable_display, **not** scheduling; `metadata_startup_script` recreates on change, the `metadata["startup-script"]` key does not | hashicorp/terraform-provider-google `compute_instance.html.markdown` (main) |
| F17 | Compute Engine (fetched 2026-10-26): standard VMs need `--instance-termination-action` (STOP/DELETE) with a run duration; range 30 s–120 days; "The termination timestamp is automatically cleared whenever a VM is stopped or suspended" and on restart "recalculated by adding that duration to the VM's latest start time"; changing the limit on an existing VM: "first stop your VM, then update its properties, then restart it" (`set-scheduling`/`setScheduling`); the automatic stop is the `compute.instances.deferredStop` operation, up to 30 s late; after a stop "you continue to incur charges for any disks, disk snapshots, or external IP addresses" | docs.cloud.google.com limit-vm-runtime; stop-start-instance |
| F18 | Caddy (fetched 2026-10-26): `caddy adapt --validate` "will load and provision the adapted configuration to check for validity"; `caddy validate` provisions all modules "but the config is not actually started"; neither page states whether provisioning performs network calls (→ early proof P0.2); caddy-dns/cloudflare requires a token with `Zone.Zone:Read` + `Zone.DNS:Edit`, syntax `dns cloudflare {env.CF_API_TOKEN}`, no file-backed token documented | caddyserver.com/docs/command-line; github.com/caddy-dns/cloudflare |

**Lane claims checked and corrected (the rest of the cited line numbers were spot-checked and hold at `a5aa46ba`):**

1. Lane 2 §4 item 5 defines reset as `docker compose … down -v`. Wrong: `-v` also removes the edge certificate volume. Reset removes exactly `<project>_postgres-data` and `<project>_minio-data` (lanes 1 and 3 agree; §2.6).
2. Lane 1 marks provider graceful-shutdown support "unknown". It is known: the provider exposes `scheduling.graceful_shutdown` only in the Beta provider (F16). v0 does not use it (S8); the normal `stop` path shuts Compose down first and the deferred stop relies on PostgreSQL crash recovery of a re-seedable dataset.
3. Lane 2 §7 states the static suite's shell/Dockerfile checks "cover `caddy.Dockerfile` and any lifecycle script" automatically. Wrong: both checks use explicit lists (F11). The Plan extends the lists (§3.1) and places the Caddy build at `infra/docker/edge.Dockerfile` (lane 3 proposed `infra/edge/Dockerfile`, lane 2 `infra/docker/caddy.Dockerfile`) so `lint_hadolint`'s `infra/docker/$f.Dockerfile` loop covers it.
4. Lane 3 §5.1 wires `APP_ENV=production` into the overlay; lane 2 §2.4 shows that cannot start against local MinIO (F5). Resolved as D-C1 (§1.3): `APP_ENV=development` with every production-equivalent setting explicit, disclosed as not production-mode.
5. Lane 1 §1 leaves the private-file root open ("the Plan must pick one"). Settled: `<main>/.local/{evidence,secrets,tmp}/g070-a1/` (F15; the G057/G069 precedent).
6. Lane 1 §3.1 gives the Terraform deployer `roles/secretmanager.admin` and lane 2 has Terraform create secret containers. Neither is needed: the bootstrap script (owner ADC) creates the VM service account, the secrets and their accessor bindings; Terraform never touches Secret Manager, so the deployer identity holds no secret or IAM-admin role (§2.1.4).
7. Lane 1 §5 assumes no service has a restart policy and that nothing auto-starts on boot. True today (F4); the overlay deliberately adds `restart: unless-stopped` to long-running services so a VM restart resumes the **same pinned images** (not a moving branch) — S13 explains why this stays within Spec §3.
8. Lane 2 §2.5 and lane 3 §2.2 agree on routing; the Caddyfile mirrors `proxy.conf` exactly (F9); lane 3's "`handle_path`-free `handle` blocks" is kept.
9. Lane 1/3 cost figures for the idle static IP (≈$0.01/h) and `pd-balanced` (≈$0.10/GB-month) are third-party/unverified; they stay marked **[unverified]** in §8.3 and must be confirmed in the owner's console before the live release.
10. Lane 3 §6 item 2 proves the log filter with `curl`; the official `caddy` image is Alpine-based without curl — the proof uses busybox `wget` inside the container (P0.2).
11. Lane 2 §1.4 wants the overlay to carry literal `image: <registry>/…@sha256:…` pins. A registry path contains the adopter's project ID (F12 forbids adopter literals in `infra/`), so the overlay carries `image: ${IMAGE_<SERVICE>:?…}` and the per-deploy manifest (private, non-secret, published in the Result) carries the digests (§2.2).
12. Lane 3 §1.1 recommends TTL 60 s "during pilots"; settled 300 s (`dns_ttl` variable, default 300): the address is static (S3), so TTL does not affect readiness.
13. Lane 1 §4 estimates 1.6–3.1 GiB runtime RSS, lane 2 §6 0.7–1.3 GB; both unmeasured. `e2-standard-2` (8 GB) covers either; the pilot measures (U-1).

## 1. Settled decisions, unproven assumptions, decisions for owner/Codex, stop conditions

### 1.1 Settled (within Spec/ADR authority; Codex may redirect at Plan review)

| # | Decision | Reason / authority |
|---|---|---|
| S0 | **Genericity, enforced:** every reusable file (`infra/terraform/**`, `infra/staging/**`, `infra/edge/**`, `infra/compose/staging.yml`, `infra/docker/edge.Dockerfile`, `docs/foundation/STAGING.md`, README) uses variables and `staging.example.net`/`example.net` examples; no project ID, zone ID, region, credential or product literal. The adopting project's inputs live in **`<main>/.local/secrets/g070-a1/<env>/staging.conf`** (0600, ignored, non-secret identifiers only: environment name, GCP project/region/zone, base domain `staging.env.foundation.test`, Cloudflare zone ID, ACME contact, machine type, session length, registry path, state bucket, product/database names) plus `terraform.tfvars`/`backend.hcl` in the same directory; secret **values** live only in Secret Manager (S6). B-LITERALS (F12) fails the MR if an identity literal lands in `infra/`, `docs/foundation/` or README. | Spec §1; Goal "generic to SaaStacular"; F12 |
| S1 | **Image path = digest-pinned images in one private Artifact Registry Docker repository, built off-VM from the exact candidate commit by the existing Compose build stanzas, pulled by the VM's service account (no pull secret, no source/toolchain/Git credential on the VM).** Concrete requirement for the one added service: content-addressed `@sha256` identity for every running service (AC1 "digest-pinned published images"), a runtime-only VM (Spec §2 "a lightweight runtime VM must not be assumed capable of large builds") and no secret delivery for pulls. Cost ≈ $0 (0.5 GB free, then ≈$0.10/GB-month [unverified]). Build-on-VM is rejected (source + Git credential on the public host, 4 vCPU/16 GB forced); `docker save/load` over IAP is the recorded fallback only if Artifact Registry is refused by an org policy (§1.4 stop → bounded delta). | Spec §4 image-path paragraph; §2 sizing |
| S2 | **VM:** `e2-standard-2` (2 vCPU / 8 GB, `provisioning_model = STANDARD`, non-Spot), variable `machine_type`; `e2-medium` rejected (shared-core, 4 GB). Boot 20 GB `pd-balanced` Debian 12 (`debian-12` family, `debian-cloud`); **data disk 50 GB `pd-balanced` mounted at `/var/lib/docker`** so images, build cache and the named volumes `postgres-data`, `minio-data`, `edge-data`, `edge-config` live on one retained disk with no Compose volume change; boot disk replaceable. Shielded VM, OS Login, no project SSH keys, no external SSH: port 22 only from the IAP range `35.235.240.0/20`. Dedicated VPC + subnet (no default network). | Spec §2; lane 1 §4 |
| S3 | **Reserved static external IPv4** (`google_compute_address`) attached to the VM: DNS records never change, resume readiness does not wait on TTL, the G057 `T → custom.B` CNAME and `edge-probe` target stay stable; cost ≤ ≈$7.30/month while stopped [unverified rate]. DNS-update-on-restart rejected (second manager of the same records per Spec/ARCHITECTURE "one manager per resource", TTL wait). The Spec's "do not claim a stable ephemeral IP" is honoured by reserving. | Spec §5; `docs/ARCHITECTURE.md:3865-3866` |
| S4 | **DNS in Terraform** (`cloudflare/cloudflare` v5 `cloudflare_dns_record`, separate module `modules/cloudflare-dns`): `B` A, `*.B` A, `custom.B` A (explicit — it is the CNAME target the core checker compares by name), all **DNS-only (`proxied = false`)**, TTL `dns_ttl` (default 300). No AAAA in v0. Caddy alone writes `_acme-challenge.*` TXT; Terraform never manages that name; G057 owns `T` CNAME/`_platform-verify.T` TXT. DNS-only is required: Universal SSL does not cover second-level wildcards and proxying would insert a second TLS terminator and rewritten client headers (F9). Terraform's token comes from `CLOUDFLARE_API_TOKEN` in the operator environment only (zone-scoped `Zone:Read` + `DNS:Edit`); never a variable (a `sensitive` flag would not keep it out of state, Spec §6). | Spec §2 "required DNS records", §5 |
| S5 | **Remote state:** GCS backend, bucket `<project>-tfstate` created once by `infra/terraform/bootstrap/bootstrap.sh` (gcloud, operator ADC; create-if-absent; versioning on, uniform access, public-access prevention enforced, 90-day noncurrent lifecycle), prefix `staging/<env>`; backend config from a private `backend.hcl` (`-backend-config`), credentials from ADC (never written to `.terraform`/plan files). Native GCS locking; stale-lock procedure documented (force-unlock only after the inventory shows no live apply). Production would use a different project and bucket. The bucket is a **named retained baseline item** (owner: environment owner; reason: state history; release: no state objects remain). | Spec §2 "reproducible remote-state backend with appropriate access/locking, documented one-time bootstrap and retirement" |
| S6 | **Secrets = GCP Secret Manager, one secret per name**, IDs `<env>-<VAR>`; containers and the VM SA's per-secret `roles/secretmanager.secretAccessor` bindings created by `bootstrap.sh`; the 8 application secrets (`POSTGRES_SUPERUSER_PASSWORD`, `DB_MIGRATOR_PASSWORD`, `DB_RUNTIME_PASSWORD`, `DB_RETENTION_PASSWORD`, `REDIS_PASSWORD`, `MINIO_ROOT_PASSWORD`, `OBJECT_STORE_SECRET_KEY`, `INTELLIGENCE_SERVICE_TOKEN`) are **generated by bootstrap** (48 hex, like `configure-env.sh`) and added as versions without being printed; `CF_API_TOKEN` (Caddy's zone-scoped token, distinct from Terraform's) and the optional `OPENROUTER_API_KEY` are owner-entered (`gcloud secrets versions add --data-file=-` from the owner's terminal); the optional **G057-owned** `IDENTITY_WORKOS_CLIENT_ID`, `IDENTITY_WORKOS_ISSUER`, `IDENTITY_WORKOS_API_KEY` containers exist from bootstrap and stay **empty until G057 Part B readiness**. **Delivery:** `deploy`/`start` run on the VM a fetch step that uses the instance metadata token and the Secret Manager REST `…/versions/<n>:access` endpoint (curl + jq; no gcloud dependency assumed) to write one mode-0600 file **`/run/<env>/app.env` on tmpfs** (never on a disk; gone at stop/expiry), passed as `docker compose --env-file`; the deploy manifest records the resolved **version numbers**, never values. Nothing is stored in Terraform variables/state/outputs, instance metadata or startup scripts; `docker compose config`/`docker inspect` output is never captured as evidence. Rotation = new version + `deploy`; DB role passwords additionally need the documented `ALTER ROLE` procedure (existing rule: provisioning never rewrites a password). | Spec §6 "secret references and a private runtime delivery mechanism"; `docs/SECURITY.md:74-85` |
| S7 | **Identities:** (a) VM runtime SA `<env>-vm@…` (bootstrap-created, never the default SA; scope `cloud-platform`; roles: `secretmanager.secretAccessor` per staging secret, `artifactregistry.reader` on the one repository; nothing else — no compute/storage/state access); (b) deployer SA `<env>-deployer@…` (bootstrap-created) with exactly `roles/compute.instanceAdmin.v1`, `roles/compute.networkAdmin`, `roles/compute.securityAdmin` (project), `roles/iam.serviceAccountUser` on the VM SA only, `roles/storage.objectAdmin` on the state bucket only, `roles/artifactregistry.writer` on the repository only, `roles/iap.tunnelResourceAccessor` and `roles/compute.osAdminLogin` (project); the human operator (owner ADC, short-lived) holds `roles/iam.serviceAccountTokenCreator` on the deployer SA and runs Terraform/gcloud with **`impersonate_service_account`**, so every apply runs with the recorded least-privilege set even when the human is project Owner. No SA keys anywhere; no GitLab OIDC in v0 (D-11 deferred — manual-trigger design documented only). Role sufficiency is a live precondition (`gcloud iam roles describe`, read-only, first account call of the pilot). | Spec §6 identities paragraph |
| S8 | **Native expiry:** `scheduling { provisioning_model = "STANDARD"; max_run_duration { seconds = var.session_max_seconds }; instance_termination_action = "STOP"; automatic_restart = true; on_host_maintenance = "MIGRATE" }`, variable validation rejects `DELETE`; **default 4 h (14400 s)** (G057's operator window is ≈115 min; 4 h covers one session plus diagnosis; cost of forgetting ≤ 4 h × VM rate). Re-arm is Compute Engine's documented behaviour (F17); `start` records the observed `terminationTimestamp` in the inventory; `status` prints it. Changing the length = `stop` → `terraform apply -var session_max_seconds=…` while TERMINATED → `start` (single manager; F17 requires a stopped VM). Provider in-place support for that update is an unproven assumption (A2). No bespoke poller; provider graceful shutdown not used (Beta, F16). | Spec §3 expiry paragraph |
| S9 | **Terraform vs stop/start:** `desired_status` is **unset** in the module; `start`/`stop` use `gcloud compute instances start|stop` (never create/delete); `provision`/`status`/`retire` are the only commands that run Terraform (`plan`/`apply`/`destroy`). Expected: `terraform plan` after a gcloud stop/start reports no changes (A1, live proof #1; fallback in §1.4). Deterministic names (`<env>-vm`, `<env>-data`, `<env>-ip`, `<env>-vpc`, `<env>-edge`/`<env>-iap` firewall rules) make a re-run converge or fail on "already exists" → `terraform import`, never a second resource. No Terraform provisioners; the startup script (`metadata["startup-script"]`, not `metadata_startup_script`) only formats/mounts the data disk by label, installs the pinned Docker Engine + Compose plugin from Docker's apt repository if absent, enables the OS Login guest agent defaults, and **never deploys the application**. | Spec §2 "must not fight Terraform", §3 provision/deploy rows |
| S10 | **Firewall:** ingress tcp 80/443 from `0.0.0.0/0` to target tag `<env>-edge`; tcp 22 from `35.235.240.0/20` only; everything else denied (dedicated VPC implied deny); egress open (apt, ACME, Cloudflare API, Artifact Registry, Secret Manager). Compose publishes only the edge's `80:80` and `443:443`; every other service gets `ports: !reset []` (defense in depth: the staging env still sets `BIND_ADDRESS=127.0.0.1` so a forgotten reset binds loopback). No HTTP/3 (443/udp) in v0. | Spec §4 ports paragraph |
| S11 | **Compose overlay `infra/compose/staging.yml`** over `local.yml` + `migrate.yml` (one owner, G070): `ports: !reset []` on postgres/redis/minio/core/web/intelligence; `build: !reset null` + `image: ${IMAGE_<SVC>:?…}` + `pull_policy: missing` on core/web/intelligence/minio/minio-init/migrate (the `:?` colon also satisfies B-PROTECTED's image regex, F11); service **`edge`** (`image: ${IMAGE_EDGE:?…}`, networks `[edge]`, ports `80:80`, `443:443`, volumes `edge-data:/data`, `edge-config:/config`, `../edge/Caddyfile:/etc/caddy/Caddyfile:ro`, env `EDGE_BASE_DOMAIN`, `ACME_EMAIL`, `EDGE_ACME_CA`, `CF_API_TOKEN`, `depends_on: web healthy`); `restart: unless-stopped` on postgres/redis/minio/core/web/intelligence/edge (S13); core env additions `PLATFORM_AUTH_HOST: ${PLATFORM_AUTH_HOST:?}`, `IDENTITY_COOKIE_PROFILE: https`, `DOMAIN_TLS_EDGE_ADDR: edge:443`, `IDENTITY_PROVIDER: workos`, and the optional `IDENTITY_WORKOS_CLIENT_ID/ISSUER/API_KEY: ${…-}` and `IDENTITY_PUBLIC_ORIGINS: ${IDENTITY_PUBLIC_ORIGINS-}` (empty = unconfigured, F5); web env `IDENTITY_COOKIE_PROFILE: https`; labels `<env>.environment`, `<env>.managed-by=compose`. `identity.yml`, `browser.yml`, `secret-scan.yml` and the `checks` profile are never composed on the VM (asserted by the B-COMPOSE extension and the early proof). | Spec §4 reuse/overlay paragraphs; F3/F4/F5 |
| S12 | **Edge image `infra/docker/edge.Dockerfile`** (context `infra/docker`, ~12 lines): `FROM caddy:<X.Y.Z>-builder@sha256:<digest> AS builder` → `RUN xcaddy build v<X.Y.Z> --with github.com/caddy-dns/cloudflare@<tag>` → `FROM caddy:<X.Y.Z>@sha256:<digest>` → `COPY --from=builder /usr/bin/caddy /usr/bin/caddy`; lane-reported current pins `2.11.7` / `v0.2.4` are **verified and resolved to digests at implementation** (recorded in the Result with `caddy list-modules`, Go version and image digest). Built off-VM in the single build slot, pushed like the app images. **`infra/edge/Caddyfile`** (≤ 60 lines, env-driven, lane 3 §2.2 shape): global `email {$ACME_EMAIL}`, `acme_ca {$EDGE_ACME_CA}` (production directory by default, Let's Encrypt staging directory for the rehearsal), `admin off`, `on_demand_tls { ask http://core:8180/internal/v1/domains/tls-ask }`, filtered JSON access log (whole query redacted on every path, `Cookie`/`Authorization`/`Set-Cookie`/`Location` deleted, no `log_credentials`); site `{$EDGE_BASE_DOMAIN}, *.{$EDGE_BASE_DOMAIN}` with `tls { dns cloudflare {env.CF_API_TOKEN}; resolvers 1.1.1.1 1.0.0.1; propagation_delay 20s }`, `header Strict-Transport-Security "max-age=300"` (no `includeSubDomains`/`preload`), `@mcp path /mcp/c/* /.well-known/oauth-protected-resource/mcp/c/*` → `reverse_proxy core:8180`, else `reverse_proxy web:3000`, both with `header_up -X-Forwarded-For/-Proto/-Host`; catch-all `https://` site with `tls { on_demand }` and the same two routes (the only place on-demand is enabled; `ask` restricts it). Caddy's default issuer chain (Let's Encrypt, ZeroSSL failover) stays; the observed issuer is recorded. | Spec §5; ADR-007/DOMAINS edge inputs; F9 |
| S13 | **Restart semantics:** `restart: unless-stopped` resumes the already-pulled pinned images after a VM reboot (no pull, no new candidate, no migration); every candidate change goes through `deploy`. This is not the "automatic moving-branch deployment on VM boot" Spec §3 forbids; it keeps a mid-session reboot from leaving the edge dark. `stop` runs `docker compose stop` (which also clears the "unless stopped" state) before the VM stop. | Spec §3 deploy row |
| S14 | **Seed `infra/staging/seed.sql`** (generic, idempotent, superuser over the PostgreSQL container's own socket via `docker compose exec -T postgres psql`, the same privilege class as `mcp-seed.sql`/`recovery.sh`; never over a network): one **marker Organization** (fixed UUID, slug `synthetic-staging-marker`, name "Synthetic staging marker", `platform_label` generated once as 32-hex from `gen_random_bytes(16)` — satisfies the 00021 grammar/reserved checks) plus two synthetic label-host Organizations of the same shape, each with one principal-less owner user and its `org_owner` binding (the G019 "every Organization keeps an owner" invariant; nobody can sign in as them; precedent `mcp-seed.sql:29-33`); `INSERT … ON CONFLICT (id) DO NOTHING`; no sessions, principals, domains, MCP rows or `enrollment_mode=open`; a verification query prints counts/names only. The exact column set is copied from `mcp-seed.sql`/migration 00021 at the pinned candidate. The marker's label is the AC3 dataset marker (survives stop/start/deploy; changes after reset). | Spec §4 seed paragraph; F8 |
| S15 | **Lifecycle tooling** = one operator-side script `infra/staging/staging.sh` (bash, shellcheck'd) with subcommands (§2.3) + `infra/staging/lib.sh` + VM-side `infra/staging/remote/{deploy,start,stop,reset,status}.sh` executed over `gcloud compute ssh --tunnel-through-iap`; files reach the VM as `git archive <candidate> infra/compose infra/postgres infra/minio infra/edge infra/staging` (the exact candidate's subtree; no Git on the VM) under `/srv/<env>/`. `--dry-run` prints every gcloud/terraform/ssh/compose command without executing (review and P0.4 proof). `migrate down`/`reapply` are not exposed. | Spec §3 |
| S16 | **Offline check runner `infra/staging/check.sh`** (coordinator-run, attributable, **not a new harness ID**): `terraform fmt -check`, `terraform init -backend=false` + `validate` + `terraform test` against a prepared read-only provider mirror with `--network none`; `caddy adapt --validate` in the edge image with `--network none`; the log-filter sentinel proof; `staging.sh --dry-run` for every subcommand; grep guards (no `roles/owner|roles/editor|serviceAccountKey|DELETE` in `infra/terraform`, no `identity.yml` reference in `infra/staging`). Reason not to register it in `lint`: hosted `validate:checks` would then need a hermetic Terraform provider mirror and the edge image in DinD — substantial CI work AGENTS.md keeps separate; D-C3 offers registration as a follow-up. | Spec §8 "add only focused coverage"; AGENTS "Keep CI changes separate and bounded" |
| S17 | **Intelligence is deployed** (full existing service set; `OPENROUTER_API_KEY` optional/unset); Redis stays ephemeral; MinIO stays local/private (archived upstream — disclosed, F2); MinIO console 9001 unreachable. | Spec §4 |
| S18 | **Docs:** new generic runbook `docs/foundation/STAGING.md` (bootstrap, inputs, lifecycle, expiry, secrets, DNS/TLS, receipts/costs, retirement, production extension: always-on compute + Cloud SQL + private R2, documented only), one README index line, a cross-link paragraph in `DOMAINS.md` "Edge deployment inputs", four provider rows in `VENDORS.md` (GCP Compute/Artifact Registry/Secret Manager; Cloudflare DNS; Let's Encrypt; Caddy + cloudflare module) with its register reason updated (F14). ARCHITECTURE/SECURITY/ADR text unchanged (Codex owns ADRs; ADR-009 already records the direction). | Spec §1 operator documentation; §7 production direction |
| S19 | **Validation stays focused** (§6): extend B-COMPOSE/B-SHELL/B-HADOLINT, one new stack-free Go test, the offline check runner, one private DB seed proof and one private no-DNS routing rehearsal; hosted fast + Duo on the MR; post-merge affected checks on the frozen merged SHA; the live pilot is evidence, never a routine MR check. No full local suite, no `ci-controls`, no browser campaign, no `migrations` suite (no schema change). | Spec §8; CLAUDE.md focused-selection rules |

### 1.2 Unproven assumptions (proof/fallback named; P0 = offline early proof §4.1, L = live pilot §4.3)

| # | Assumption | Proof / fallback |
|---|---|---|
| A1 | With `desired_status` unset, `terraform plan` after a gcloud stop or start shows **no changes** (S9). | **L-1** (first `plan -refresh-only`/`plan` after the first stop and after the first start). Fallback: manage `desired_status` as a tfvar — `start`/`stop` become `apply` runs under the state lock (bounded correction to `staging.sh`; no other design change). |
| A2 | The provider updates `scheduling.max_run_duration` **in place** on a TERMINATED instance (no recreate, no `allow_stopping_for_update`). | **L-5** (apply `session_max_seconds=600` while stopped for the expiry proof, then restore 14400). Fallback: `gcloud compute instances set-scheduling` as the documented manual step **with `lifecycle { ignore_changes = [scheduling[0].max_run_duration] }`** so Terraform does not fight it (recorded exception to single-manager; Codex reviews). |
| A3 | `caddy adapt --validate` with the cloudflare DNS module and the ACME issuer makes **no network call** at provisioning time (F18 is silent). | **P0.2** under `--network none`; a failure is a finding: the offline check falls back to `caddy adapt` (syntax + JSON artefact assertions) and `--validate` moves to the private routing rehearsal (P0.7) with network. |
| A4 | Caddy's `log` `filter` can redact the whole query (`request>uri` `regexp` replace `\?.*$` → `?REDACTED`, or per-key `query replace`) and the result shows host/path/status. | **P0.2** sentinel run (`G070SENTINEL-*` + `?& (not included)=` → zero hits). Fallback: `log { output discard }` for the HTTP server (equivalent to the test proxy's `access_log off`; hop evidence then comes from web/core logs only). |
| A5 | `docker compose config` accepts the overlay with `build: !reset null` and digest-pinned `image:` values while the build contexts (`core/`, `web/`, …) are **absent** (the VM holds only the archived `infra/` subtree). | **P0.1** run from a directory holding only the archive. Fallback: keep `build:` reset and add empty placeholder context directories in the archive (no semantic change). |
| A6 | Terraform `mock_provider "google"`/`"cloudflare"` plan-mode tests can assert the firewall ports, scheduling values, `STANDARD`/`STOP`, labels, the three DNS-only records and the absence of secret-shaped outputs. | **P0.3**. Fallback: assert the same facts on `terraform plan -out` + `terraform show -json` with placeholder variables (needs the provider mirror only, still offline). |
| A7 | Debian 12 GCE image + Docker's apt repository: the pinned engine/compose-plugin versions are installable by the startup script within the first boot; `/run` is tmpfs. | **P0.5** rehearses the script body in a `debian:12` container (network to apt, disk-mount step skipped by flag); **L-2** proves it on GCE. Fallback: pin to the versions apt actually serves (recorded). |
| A8 | `e2-standard-2` holds the full service set with margin (U-1). | **L-3** measures `docker stats`, cgroup `memory.peak`, `free -m`, `docker system df`. Fallback: `machine_type = e2-standard-4` (variable; needs a stopped VM — F16). |
| A9 | The VM can fetch Secret Manager versions through the metadata token with curl + jq and pull from Artifact Registry with its attached SA (no `gcloud auth configure-docker` needed when using the SA token: `docker login -u oauth2accesstoken --password-stdin`). | **L-3**. Fallback: install `google-cloud-cli` in the startup script (pinned) and use `gcloud secrets versions access` / `gcloud auth configure-docker`. |
| A10 | The role set of S7 is sufficient (`compute.instanceAdmin.v1` covers addresses/disks/metadata; `networkAdmin`+`securityAdmin` cover VPC/subnet/firewall). | **L-0** `gcloud iam roles describe` (read-only) before any write; a missing permission adds the narrowest role, recorded. |
| A11 | Cloudflare DNS-01 through the zone-scoped token issues `{B, *.B}` within Caddy's propagation window (20 s delay, 2 min timeout). | **L-4a** against the Let's Encrypt **staging** CA on a separate `edge-data-rehearsal` volume first; then **L-4b** production on the retained `edge-data` volume. Fallback: raise `propagation_delay`/`propagation_timeout` (config only). |
| A12 | `DOMAIN_TLS_EDGE_ADDR=edge:443` (internal) lets core's `edge-probe` verify a custom host's chain with default roots from the distroless image (U-2). | **L-6** smoke: probe an unknown host → 403 from `tls-ask` and a refused handshake; the real custom-host probe is G057's. Fallback: `custom.B:443` (hairpin through the external IP; GCE supports it, unverified). |
| A13 | PostgreSQL recovers cleanly from the deferred (abrupt) stop (WAL replay) and the marker survives. | **L-5** (expiry proof stops the VM without `compose stop`; the next `start` verifies the marker). Fallback: accept reseed for that case (documented) and keep `stop` as the normal path. |

### 1.3 Decisions that need a bounded owner/Codex choice (smallest options; recommendation first)

| ID | Decision | Options | Owner/Codex |
|---|---|---|---|
| **D-C1** | **Production-mode gap G-1** (F5): local MinIO over `http://minio:9000` cannot run under `APP_ENV=production`. | (i) **recommended:** `APP_ENV=development` with every production-equivalent setting explicit (`IDENTITY_COOKIE_PROFILE=https` on web and core, no `PLATFORM_HOST_PORT`, no fixtures, default WorkOS base, `edge-probe` adapter, real DNS/HTTPS) — smallest surface; **disclosed in the Result as "staging runs the development loader mode; not production-mode"**; G057's P-4 stays held; the readiness evaluator stays "blocked" in either mode (F6); production itself has no gap because R2 is https. (ii) `APP_ENV=production` + MinIO TLS: a private CA, certificate rotation, a CA bundle in the distroless core image and an https browser-reachable signing endpoint (a public `objects.B` host) — new surface, recorded as a bounded follow-up only. | Codex (Spec §4 wording "production-mode configuration" vs. the actual loader) |
| **D-C2** | **G057 Part B amendment** (§10): one exact deployment receipt, retained baseline instead of destroy-every-host, no `down -v`, file-ownership move (`infra/compose/edge.yml`, `infra/edge/Caddyfile` leave G057's allow-list), identity values delivered through G057-owned Secret Manager entries + a G070 `deploy`. | Accept as proposed, or keep G057's separate `--env-file` (then G070's `deploy` accepts an optional second env file — one script flag). | Codex (owns G057 documents) |
| **D-C3** | **Harness registration of the offline staging checks** (S16). | (i) **recommended:** unregistered coordinator-run `infra/staging/check.sh` with ledger lines in the Result (local and post-merge); (ii) register `B-STAGING-OFFLINE` in a new suite outside every profile (touches `run.sh` suite dispatch, `SUITES_ALL/KNOWN`, B-CI-CONFIG consistency, count pins) as a separate bounded CI follow-up. | Codex |
| **D-O1** | **Owner inputs (never assumed):** GCP project ID with billing linked; region/zone; environment name (`^[a-z][a-z0-9-]{1,20}$`); Cloudflare account/zone holding `staging.env.foundation.test` and two zone-scoped tokens (Terraform; Caddy); ACME contact e-mail; confirmation that `staging.env.foundation.test`/`*.staging.env.foundation.test` are free for this use (no registration/purchase assumed); WorkOS staging application (G057 input, not needed for the G070 pilot); pilot cost ceiling. | Supplied in writing before L-0 (§4.3). | Owner |
| **D-O2** | **Retention after the pilot.** | (i) **recommended:** `retire` at the end of the pilot (proves retirement; zero running cost) and `provision` again (≈15 min, same inputs, one more `{B,*.B}` issuance within the 5/week duplicate budget) when G057 Part B readiness is recorded; (ii) keep the stopped baseline (≈$11–15/month [unverified]) as the G057 environment immediately. | Owner |
| **D-O3** | **Session length and HSTS defaults** for the adopting environment: 4 h session (`session_max_seconds`), `Strict-Transport-Security: max-age=300` without `includeSubDomains`. | Keep the generic defaults, or set other values in `staging.conf` (reversible configuration; no code change). | Owner (optional) |
| **D-O4** | **Operator machine authority:** the live phase runs from the owner's machine (WSL) after `gcloud auth login` + `gcloud auth application-default login` and `export CLOUDFLARE_API_TOKEN=…`; the coordinator executes the released commands. Native permission rules for `gcloud`, `terraform` (in Docker) and `infra/staging/staging.sh` are requested by Claude and installed by Codex per the standing permission handoff before L-0. | Confirm the operator machine and session; or the owner runs the listed commands personally. | Owner |

### 1.4 Stop conditions (Goal "Stop / Escalation"; stop the affected lane, report the exact conflict and smallest options)

- A production-mode gap that cannot be closed within D-C1 (e.g., a loader rule that refuses the development-mode staging template) → stop; options: loader change (actual code/adapter gap, outside scope) or Spec amendment.
- IAM beyond S7's set is required, or an org policy blocks SA impersonation/Artifact Registry/Secret Manager → stop; options: narrowest additional role, `docker save/load` fallback (S1), scp delivery fallback (S6).
- Any path where a secret value would enter Git, Terraform state/plan/outputs, metadata, logs, console output or evidence → stop before the write; no "redact later".
- A1/A2 failures that would require Terraform to manage stop/start or scheduling outside the single-manager rule (fallbacks in §1.2 are bounded corrections needing Codex review, not silent changes).
- Terraform plan touching anything not labelled `environment=<env>`/`managed-by=terraform` in the staging project, or any production/unrelated resource → stop (`plan` is always shown and saved before `apply`; `--yes` is explicit).
- Duplicate-certificate budget (5 per 7 days) or a `_acme-challenge` litter condition that cannot be cleaned → stop issuance attempts; record.
- A third concurrent private run, a second concurrent build, an unowned Docker object, or a collision with the owner's M1 preview / the frozen G057 checkout → stop.
- Any need to edit `core/**` (beyond the one test file), `web/**`, migrations, MCP/HTTP contracts, `.env.example`, `local.yml`, the identity overlay or G057 documents → scope stop.

## 2. Approach

### 2.1 Terraform (`infra/terraform/`; owner W1)

```
infra/terraform/
  README.md                          # bootstrap → provision → retire; generic; links STAGING.md
  bootstrap/bootstrap.sh             # one-time, idempotent gcloud (owner ADC): enable APIs (compute, iam, iap,
                                     # storage, secretmanager, artifactregistry); state bucket; VM SA + deployer SA
                                     # + bindings (S7); Artifact Registry repo; secret containers (S6) + generated
                                     # app secret versions (never printed); receipt of what existed/was created
  modules/gcp-staging-vm/            # provider-pure google module
    versions.tf variables.tf network.tf identity.tf instance.tf outputs.tf startup.sh
    tests/plan.tftest.hcl            # mock_provider assertions (A6)
  modules/cloudflare-dns/            # B, *.B, custom.B A records, proxied=false, ttl=var.dns_ttl
    versions.tf variables.tf main.tf outputs.tf tests/plan.tftest.hcl
  environments/example/              # the only committed root; generic values only
    main.tf  (module calls; backend "gcs" {} partial config)
    backend.hcl.example  terraform.tfvars.example  .terraform.lock.hcl
```

- **Variables (module):** `project_id`, `region`, `zone`, `environment_name` (validated), `machine_type` (`e2-standard-2`), `boot_image` (`debian-12`), `boot_disk_size_gb` (20), `data_disk_size_gb` (50), `disk_type` (`pd-balanced`), `session_max_seconds` (14400; 30 ≤ n ≤ 10368000), `instance_termination_action` (validation: only `STOP`), `iap_ssh` (true), `vm_service_account_email` (from bootstrap), `operator_principals` (OS Login/IAP members), `labels` (map, merged with `environment`, `managed-by=terraform`). No secret-shaped variable exists. **Outputs:** instance name/zone/self-link, `external_ip`, SA e-mail, data disk name, network/subnet/firewall names, `max_run_duration_seconds` — nothing else.
- **DNS module variables:** `zone_id`, `base_domain`, `address`, `dns_ttl` (300), `custom_label` (`custom`). Outputs: the three record IDs (inventory input).
- **Versions/locks:** Terraform `~> 1.16` in `hashicorp/terraform:<1.16.x>@sha256:<digest>`; `hashicorp/google` and `cloudflare/cloudflare` (v5) pinned exactly in `required_providers` and `.terraform.lock.hcl` written with `terraform providers lock -platform=linux_amd64 -platform=linux_arm64 -platform=darwin_arm64`. Lane-reported current versions (Terraform 1.16.7, google 8.8.1, cloudflare 5.28.1) are **verified at implementation**; `google 7.46.x` is the recorded fallback if 8.x breaks the module.
- **Offline commands** (coordinator; `<T>` = `<main>/.local/tmp/g070-a1`):
  - mirror preparation (network, once, under the existing cache-preparation lock): `docker run --rm -v "$PWD/infra/terraform:/src:ro" -v "<T>/tf-mirror:/mirror" -w /src/environments/example hashicorp/terraform:<pin> providers mirror -platform=linux_amd64 /mirror`
  - offline: `docker run --rm --network none -v "$PWD/infra/terraform:/src:ro" -v "<T>/tf-mirror:/mirror:ro" -v "<T>/tf-work:/work" -e TF_DATA_DIR=/work/.terraform -e TF_CLI_CONFIG_FILE=/work/cli.tfrc -w /src/environments/example hashicorp/terraform:<pin> sh -c 'terraform fmt -check -recursive -diff /src && terraform init -backend=false -input=false && terraform validate -json && terraform test'` with `cli.tfrc` = `provider_installation { filesystem_mirror { path = "/mirror" } direct { exclude = ["*/*"] } }`.
- **Partial-apply recovery:** state is written per resource; re-run `plan` (shows the remainder) → `apply`. Orphans (API success without state write, stale lock) are found by the label query (`gcloud compute instances|disks|addresses list --filter=labels.environment=<env>`) versus `terraform state list`, then `terraform import`; never a differently named second resource. `status` performs this reconciliation and refuses new allocation while it finds an unattributed object.

### 2.1.4 IAM summary (recorded for AC5)

| Principal | Roles (exact) | Scope |
|---|---|---|
| Human operator (owner ADC) | project roles as they already exist (recorded, not expanded); `roles/iam.serviceAccountTokenCreator` on `<env>-deployer` | deployer SA |
| `<env>-deployer` SA (impersonated by Terraform/gcloud) | `compute.instanceAdmin.v1`, `compute.networkAdmin`, `compute.securityAdmin`, `iap.tunnelResourceAccessor`, `compute.osAdminLogin` | project (one staging project) |
| | `iam.serviceAccountUser` | `<env>-vm` SA only |
| | `storage.objectAdmin` | state bucket only |
| | `artifactregistry.writer` | the one repository |
| `<env>-vm` SA (attached) | `secretmanager.secretAccessor` | each `<env>-*` secret |
| | `artifactregistry.reader` | the one repository |
| Not granted anywhere | Owner/Editor, `resourcemanager.projectIamAdmin`, `iam.serviceAccountKeyAdmin`, any `secretmanager.*` on the deployer, any compute/storage role on the VM SA | — |

### 2.2 Runtime: images, overlay, env, seed (owner W2)

- **Build/push (operator side, coordinator, one build slot):** `staging.sh build --candidate <sha>` runs `git archive`-free builds from a detached checkout of `<sha>` with the existing stanzas: `docker compose --env-file <dummy> -f infra/compose/local.yml -f infra/compose/migrate.yml --profile migrate build core web intelligence minio minio-init migrate` (`COMPOSE_PARALLEL_LIMIT=1`) plus `docker build -f infra/docker/edge.Dockerfile infra/docker`; tags `<AR_REPOSITORY>/<service>:<sha>`; `staging.sh push` pushes and resolves `RepoDigests` into the **deploy manifest** `<main>/.local/evidence/g070-a1/deploy/<sha>.tsv` (candidate SHA, per-image digest, Dockerfile sha256, `staging.yml`/Caddyfile sha256, secret version numbers; no secret value — publishable in the Result).
- **Env on the VM:** `/run/<env>/app.env` = non-secret lines rendered from `staging.conf` (`PRODUCT_NAME`, `PLATFORM_BASE_DOMAIN=B`, `CUSTOM_DOMAIN_CNAME_TARGET=custom.B`, `PLATFORM_AUTH_HOST=auth.B`, `DATABASE_NAME`, `COMPOSE_PROJECT_NAME=<env>`, `APP_ENV=development` (D-C1), `BIND_ADDRESS=127.0.0.1`, the inert port numbers, internal URLs, `OBJECT_STORE_*` non-secrets, `MINIO_ROOT_USER`, `REDIS_HOST/PORT_INTERNAL`, `IMAGE_*` from the manifest, `EDGE_BASE_DOMAIN=B`, `ACME_EMAIL`, `EDGE_ACME_CA`) + the fetched secret values + `IDENTITY_PUBLIC_ORIGINS=https://auth.B` **only when** the three identity secrets have versions. The committed example is `infra/staging/examples/app.env.example` (`staging.example.net`; secrets empty).
- **Seed:** S14; executed by `remote/deploy.sh` after `migrate up` only when the marker is absent (`provision`/`reset`), never on `start`/`deploy` of an existing dataset.
- **Core config test** `core/internal/config/staging_config_test.go` (new, stack-free): the generic template values inline (comment points at the example file; the two are kept in step by review) → `Load`/`LoadIdentity`/`LoadDomains`/`LoadMCP` succeed in development with `IDENTITY_COOKIE_PROFILE=https`, `DOMAIN_TLS_EDGE_ADDR=edge:443`, identity unconfigured (`Configured()==false`) and configured (dummy triple); `APP_ENV=production` + `http://minio:9000` → the exact refusal text (G-1 executable); production + identity unset → its refusal. Uncovered today: the staging template as a whole through all four loaders.

### 2.3 Operator lifecycle interface (`infra/staging/staging.sh <command> --config <staging.conf> [flags]`; owner W2)

| Command | Terraform | What it does (exact behaviour) | Safety |
|---|---|---|---|
| `bootstrap` (separate `infra/terraform/bootstrap/bootstrap.sh`) | none (gcloud) | idempotent: APIs, state bucket, SAs + bindings, AR repository, secret containers + generated app secret versions; prints a receipt of created/existing names | owner ADC; never prints values; refuses without `--project` matching `staging.conf` |
| `build` / `push` | none | §2.2; writes/updates the deploy manifest | build slot; AR push needs the deployer impersonation |
| `provision --candidate <sha> [--yes]` | `init` (backend.hcl) → `plan -out <private>` → show → `apply` | after apply: wait RUNNING + IAP/OS Login ready, then `deploy` (which seeds only on an empty dataset) | plan always shown and saved privately; `--yes` required for apply; re-run converges (no changes → skip apply; marker present → no reseed); refuses if the plan touches an unlabelled resource |
| `deploy --candidate <sha>` | none | upload the candidate's `infra/` archive + manifest; fetch secrets → `/run/<env>/app.env`; `docker login` with the VM SA token; `compose pull`; `up -d --wait postgres redis minio minio-init`; `migrate status` → `migrate up`; stop app services if a candidate change is pending → `up -d --wait` core/web/intelligence/edge; seed if marker absent; readiness (`/readyz` via `https://B/api/ready` with `--resolve`, marker label, certificate not-after ≥ 7 d, running digests = manifest); prints running versions | never `down`, never `migrate down/reapply`; refuses a manifest whose migrate and core digests come from different candidates |
| `start` | none | `gcloud compute instances start` → RUNNING → verify `external_ip` unchanged (static) → `remote/start.sh`: fetch secrets, `up -d --wait` stores → `migrate status/up` (pending only) → `up -d --wait` app + edge → readiness incl. HTTPS (serial recorded) → record `terminationTimestamp` (re-armed) | never resets DB/object/certificate state; refuses if `status` finds unattributed resources |
| `stop` | none | `remote/stop.sh`: `docker compose stop` (PostgreSQL fast shutdown first), delete `/run/<env>/app.env` → `gcloud compute instances stop` → poll `TERMINATED` → receipt: retained disks (70 GB), static IP, DNS records, `edge-data`, state, secrets, AR images with their costs | preserves database/object bytes and Caddy account/certificates |
| `reset --confirm <env>` | none | VM must be RUNNING and `COMPOSE_PROJECT_NAME == <env>`: `compose down` (volumes kept) → `docker volume rm <env>_postgres-data <env>_minio-data` (exact names) → first-init path (roles) → `migrate up` → seed → new marker; `edge-data`/`edge-config` untouched | distinct from start/deploy; cannot address another project/VM (project and instance name bound in `staging.conf` and checked against `gcloud config`) |
| `retire [--yes] [--purge-images]` | `plan -destroy` → show → `destroy` | `stop` first → destroy VM, data disk, address, firewall, subnet/VPC, DNS records → verify absence (label query empty, Cloudflare API read empty, `terraform state list` empty) → receipt naming the retained set (state bucket/object, secrets, SAs, AR repo and, unless `--purge-images`, its images) with owner/reason/release condition | `--yes` required; never deletes bootstrap-owned items |
| `status` / `receipt` | `plan -refresh-only` (status only, read) | provider state (instance status, `terminationTimestamp`, IP), Docker state (`compose ps`, digests), marker, certificate serial/not-after, DNS answers vs state IDs, cost line (running/stopped), inventory reconciliation; `receipt` writes the all-category receipt to `<main>/.local/evidence/g070-a1/receipts/` | read-only |

Every command supports `--dry-run`; every mutating command appends to `<main>/.local/evidence/g070-a1/inventory.tsv` at the moment a resource is created/changed (provider IDs, DNS record IDs, Docker objects, worktrees, temp paths, slot, owner, lifecycle).

**Retained baseline and receipt (AC6).** Categories always listed with count/owner/reason/release condition: VM (stopped), boot disk, data disk, static address, 3 DNS records, `edge-data` (ACME account + `{B,*.B}` certificate), `edge-config`, Terraform state object + bucket, Secret Manager secrets (names + version counts), AR repository + image digests, service accounts, local Docker objects (task images incl. untagged IDs, containers, volumes, networks), worktrees, `.local/{evidence,secrets,tmp}/g070-a1` paths. Intentional retention (owner: G070 environment owner → G057 during Part B; reason: repeat sessions/G057 readiness; release: `retire` after G057 Part B acceptance or an owner decision) is listed as retained, never reported absent; disposable leftovers hold new consuming work until reconciled.

### 2.4 Edge (owner W3) — S12; provenance record (Dockerfile sha256, builder/base digests, plugin tag, `caddy list-modules`, Go version, image digest, build host/date) in `infra/edge/README.md` template and the Result.

### 2.5 DNS/TLS sequence (live): records applied by Terraform → `deploy` with `EDGE_ACME_CA=<LE staging directory>` on `edge-data-rehearsal` (observe `certificate obtained successfully`, issuer `(STAGING)`, TXT write/cleanup) → `deploy` with the production directory on the retained `edge-data` (one issuance of `{B,*.B}`; the only production issuance a normal pilot performs) → renewals are Caddy's (≈ 30 days before expiry; a certificate expired while stopped > 60 d is reissued at `start` before HTTPS-ready is reported).

### 2.6 Reset/retire boundaries: reset never touches `edge-data` (LE duplicate budget 5/7 d; ADR-009 §4); only `retire` deletes it (the ACME account is abandoned, no CA call).

### 2.7 Database/schema changes: **none.** No migration, role, grant or policy change; `10-roles.sh`, `provision-roles.sql` and the migrator path are reused unchanged (Goal stop condition respected).

### 2.8 Contracts/interfaces: no public API/MCP contract change. New operator interface = §2.3; new Compose variable contract = `IMAGE_*`, `PLATFORM_AUTH_HOST`, `EDGE_BASE_DOMAIN`, `ACME_EMAIL`, `EDGE_ACME_CA`, `CF_API_TOKEN`, optional `IDENTITY_*`.

## 3. Harness/bookkeeping inventory and the minimum mapping change for each (owner W4)

### 3.1 `infra/validation/run.sh` (lint suite; hosted fast runs it)
- `lint_shellcheck` list += `infra/staging/staging.sh infra/staging/lib.sh infra/staging/check.sh infra/staging/remote/*.sh infra/terraform/bootstrap/bootstrap.sh infra/terraform/modules/gcp-staging-vm/startup.sh` (B-SHELL PASS text updated).
- `lint_hadolint` loop += `edge` (B-HADOLINT PASS text updated).
- `static_compose`: a derived env `static-staging.env` = `static.env` + `IMAGE_<SVC>=registry.example.invalid/<svc>@sha256:<64 hex>`, `PLATFORM_AUTH_HOST=auth.<domain>`, `EDGE_BASE_DOMAIN`, `ACME_EMAIL=ops@example.net`, `EDGE_ACME_CA=https://acme-v02.api.letsencrypt.org/directory`, `CF_API_TOKEN=<random hex>`, `IDENTITY_WORKOS_API_KEY=<random hex>`; `docker compose --env-file static-staging.env -f local.yml -f migrate.yml -f staging.yml --profile migrate config --format json` → assertions: every service except `edge` has 0 ports; `edge` publishes exactly `80/tcp`, `443/tcp`; no `build` key on core/web/intelligence/minio/minio-init/migrate; every `image` of those and of `edge` matches `@sha256:[0-9a-f]{64}$`; networks `edge=[edge]`, `web=[edge]`, `core=[edge data inference]`, `intelligence=[inference]`, `migrate=[data]`; no service named `emulator|proxy|identity-setup|core-replica|web-browser|secret-scan`; no env key matching `^(IDENTITY_TEST_|CUSTOM_DOMAIN_DNS_FIXTURE_FILE|DOMAIN_TLS_FIXTURE_FILE|MCP_CONNECT_JWKS_FILE|PLATFORM_HOST_PORT|IDENTITY_WORKOS_API_BASE_URL)`; `APP_ENV != test`; secret matrix += `[CF_API_TOKEN]="edge"`, `[IDENTITY_WORKOS_API_KEY]="core"`; config JSON deleted. B-COMPOSE PASS text updated. (Controls: none exist for B-COMPOSE today; no new control matrix.)
- No change to `required-checks.tsv`, `infra/ci/lib.sh` profile counts, `validate-controls.sh`, `profile-paths.tsv` (`infra/` is already foundation) or `.gitlab-ci.yml`.

### 3.2 `infra/validation/render-customizations.tsv`: VENDORS.md reason row extended (F14). B-LITERALS, B-SECRET-SOURCE (gitleaks; example tokens obviously fake, e.g. `CF_API_TOKEN=replace-me`), B-PROTECTED (tagged FROM, colon in `image:`), B-DOCKERIGNORE (edge shares the MinIO context rule) cover the new files automatically — reuse.

### 3.3 `infra/staging/check.sh` (S16): the offline runner's ledger lines (`TF-FMT`, `TF-VALIDATE`, `TF-TEST`, `EDGE-ADAPT`, `EDGE-FILTER`, `LIFECYCLE-DRYRUN`, `IAM-GREP`) are recorded in the Result with command, exit code and input SHA.

## 4. Early proof (P0; riskiest first; inside attempt 1; coordinator-run; **no paid allocation**), then the single paid pilot

### 4.1 Offline P0 (order = risk; stop criteria stated)

| Step | Assumption | Exact command/selector (`<E>` = `<main>/.local/evidence/g070-a1`) | Stop criterion | Runs where |
|---|---|---|---|---|
| P0.1 | A5, S11 | `tar -C <T>/archive -xf <(git archive <head> infra/compose infra/postgres infra/minio infra/edge infra/staging)`; `cd <T>/archive && docker compose --env-file <T>/p0.env -f infra/compose/local.yml -f infra/compose/migrate.yml -f infra/compose/staging.yml --profile migrate config --format json > <T>/cfg.json` + the §3.1 `yq` assertions; delete `cfg.json` | any assertion fails → fix overlay; `config` refuses missing contexts → A5 fallback | coordinator, stack-free, network none |
| P0.2 | A3, A4 | build slot: `docker build -f infra/docker/edge.Dockerfile -t g070-edge:p0 infra/docker` (network for xcaddy; prepared once); `docker run --rm --network none -v $PWD/infra/edge:/edge:ro -e EDGE_BASE_DOMAIN=staging.example.net -e ACME_EMAIL=ops@example.net -e EDGE_ACME_CA=https://acme-staging-v02.api.letsencrypt.org/directory -e CF_API_TOKEN=replace-me g070-edge:p0 caddy adapt --config /edge/Caddyfile --validate --pretty > <E>/p0/edge-adapt.json`; assert `dns.providers.cloudflare` in `caddy list-modules`, the `ask` URL, `on_demand` only in the catch-all server, the filter chain and `header_up` deletions in the JSON; then the sentinel run: `docker run -d --network none --name g070-edge-p0 -v <T>/filter-test/Caddyfile:/etc/caddy/Caddyfile:ro g070-edge:p0` (site `http://localhost:8080`, same `log` block) → `docker exec g070-edge-p0 wget -qO- 'http://127.0.0.1:8080/auth/callback?code=G070SENTINEL-1&state=G070SENTINEL-2&error=G070SENTINEL-3&error_description=G070SENTINEL-4'` and `/signup/continue?ref=G070SENTINEL-5`, `/signup/prove?n=G070SENTINEL-6`, `/signup/complete?p=G070SENTINEL-7` → `docker logs` grep `G070SENTINEL` and `?& (not included)=` → **zero hits**, host/path/status present; grep the adapted JSON for `replace-me` → zero hits (token is an env reference) | validate needs network → A3 fallback; any sentinel hit → A4 fallback | coordinator; build slot then network none |
| P0.3 | A6, S8–S10 | mirror prep (network, once) then the §2.1 offline `fmt/init -backend=false/validate/test` command; tests assert: firewall allows only tcp 80/443 from `0.0.0.0/0` and tcp 22 from `35.235.240.0/20`; `scheduling.max_run_duration.seconds == var.session_max_seconds`; `instance_termination_action == "STOP"`; `provisioning_model == "STANDARD"`; `desired_status` null; no `metadata_startup_script`; metadata keys exactly `enable-oslogin`, `startup-script`; labels contain `environment`/`managed-by`; outputs contain no name matching `(?i)secret|token|password|key`; DNS module emits exactly 3 A records, `proxied=false`, no TXT | mock tests unsupported → A6 fallback (`plan -out` + `show -json` assertions, still offline) | coordinator, network none |
| P0.4 | S15 | `bash infra/staging/staging.sh <cmd> --config infra/staging/examples/staging.conf.example --dry-run` for every subcommand → printed commands compared with §2.3 (no `create`, no `down -v`, no `migrate down`, no `-v`); `shellcheck` via the extended B-SHELL list | a forbidden verb appears → fix | coordinator, stack-free |
| P0.5 | A7 | `docker run --rm debian:12@sha256:<pin> bash -s -- --no-disk < infra/terraform/modules/gcp-staging-vm/startup.sh` (apt network) → `docker --version`, `docker compose version` equal the pins; `mount | grep /run` tmpfs | pinned package unavailable → re-pin (recorded) | coordinator |
| P0.6 | S14 | private per-run database (existing private-run rules, one pool slot, network limited to the run): `migrate up` → `psql < infra/staging/seed.sql` → counts → seed again → counts unchanged, marker label unchanged; `provision-roles.sql` repeat = no change (existing PROV proof reused by reference) | counts change or a CHECK fails → fix seed | coordinator |
| P0.7 | S11/S12 routing without DNS | private no-port Compose stack `COMPOSE_PROJECT_NAME=g070rt-<run>` with the locally built candidate images (`IMAGE_*` = local tags), overlay + an override `tls internal` for the edge, `EDGE_BASE_DOMAIN=staging.example.net`; from a one-off container on the `edge` network: `curl --cacert <caddy internal root> --resolve staging.example.net:443:<edge ip> https://staging.example.net/api/ready` → 200; `https://<seeded label>.staging.example.net/` → 200; `https://unknownlabel.staging.example.net/` → the fixed 404; `https://staging.example.net/mcp/c/anything` reaches core (401/404 from core, not web); `https://staging.example.net/internal/v1/domains/tls-ask?domain=x` → web 404; `curl --resolve g070smoke.invalid-test.example:443:<edge ip>` → handshake refused and core logs the `tls-ask` 403 | routing or Host passthrough wrong → fix Caddyfile | coordinator; pool slot; builds in the build slot |
| P0.8 | §2.2 test | `bash infra/validation/run.sh --base <head> --allow-file <G070 allow-file> --report-dir <E>/p0-unit unit-core` (B-UNIT-CORE includes `staging_config_test.go`) | refusal text differs → fix test, never the loader | coordinator, stack-free |

P0 gate: P0.1–P0.4 and P0.8 pass before the MR is opened; P0.5–P0.7 pass before the live release is requested. Findings are consolidated in one checkpoint in the Result (owner direction 2026-10-23); a failed proof is corrected within attempt 1.

### 4.2 Which G070 checks reuse existing validation infrastructure
B-COMPOSE/B-SHELL/B-HADOLINT/B-LITERALS/B-SECRET-SOURCE/B-PROTECTED/B-DOCKERIGNORE (lint, hosted fast + local), B-UNIT-CORE (unit-core), the private per-run DB and private Compose run rules (two-run pool, build slot), the existing `run.sh --allow-file` B-SCOPE, the logrecord fixture for the `tls-ask` 403 line, G057's sentinel set (prefix changed to `G070SENTINEL-`).

### 4.3 Single paid pilot — exact preconditions (recorded in the Result before the first write) and sequence

**Preconditions (all in writing):** D-O1 inputs; D-O2/D-O3/D-O4 answers; Codex execution release **and** the separate live-action release naming: project `<id>`, billing account linked, region/zone, environment name, the resource list below, the cost model of §8.3 with the owner's confirmed prices, cost ceiling; owner ADC session active on the operator machine; permission rules installed and verified; P0.1–P0.8 PASS on the candidate; the candidate SHA pinned (an accepted staging commit, G069's integrated UI reconciled); images built and pushed (`build`/`push`); `bootstrap.sh` run once with its receipt; owner-entered `CF_API_TOKEN` (and optionally `OPENROUTER_API_KEY`) versions present; pool/build slot free; inventory row for the environment opened.

**Resources the release authorizes (exact):** 1 VPC + 1 subnet, 2 firewall rules, 1 static external IPv4, 1 `e2-standard-2` VM, 1 × 20 GB + 1 × 50 GB `pd-balanced`, 3 Cloudflare A records in the named zone, 1 GCS bucket (state), 1 Artifact Registry repository (≈ 1.5–2 GB), ≤ 12 Secret Manager secrets, 2 service accounts, 2 ACME issuances (`{B,*.B}` on the staging CA, then production), plus the `_acme-challenge` TXT records Caddy writes and removes.

| Step | Action | Proof collected |
|---|---|---|
| L-0 | `gcloud iam roles describe` for every role of §2.1.4 (read-only); `bootstrap.sh` receipt; token health `/user/tokens/verify` (status only) | A10; AC5 IAM record |
| L-1 | `provision --candidate <sha>` (plan reviewed and saved privately → apply → deploy with `EDGE_ACME_CA=staging` on `edge-data-rehearsal`) | VM RUNNING, `terminationTimestamp` set, IP = static, DNS authoritative answers = state IDs/`proxied=false`; staging-CA issuance observed (A11) |
| L-1n | **Deterministic partial-apply negative:** re-run `provision` with a deliberately invalid `zone_id` override → DNS module fails after the VM exists → re-run with the correct value → converges with no duplicate VM/disk/IP; label query = 1 of each | AC3/AC6 "safely retry partial completion" |
| L-2 | `deploy` with the production CA on the retained `edge-data` | `openssl s_client -servername B` and `-servername <random>.B` chain to ISRG roots (or the observed issuer), SAN = `B,*.B`, not-after ≈ +90 d; serial recorded |
| L-3 | AC2/AC4/AC5 smoke from the operator machine: `https://B/api/ready` 200; `https://<seeded label>.B/` 200 and `https://<random>.B/` 404 under the same wildcard (no new DNS/cert); SNI for an unknown custom host refused + core `tls-ask` 403 in the log; `/internal/*` unreachable (web 404), `curl http://<ip>:8180|5432|9001` refused; `ss -tlnp` on the VM shows only Caddy's 80/443 published; `G070SENTINEL-*` set against `B`, `auth.B`, `g070sentinel.B` → grep edge/web/core logs, VM journal, evidence dir → zero hits; running digests = manifest; `docker stats`/`memory.peak`/`free`/`docker system df` recorded (A8); grep Terraform state/plan JSON/metadata/console/evidence for the token's first 6 chars and `BEGIN PRIVATE KEY` → zero | AC2, AC4, AC5 |
| L-4 | `stop` → receipt → `start` | marker label unchanged, certificate serial unchanged, no new `obtain` in Caddy's log, `terminationTimestamp` re-armed, `terraform plan` empty before and after (A1) |
| L-5 | `stop` → `terraform apply -var session_max_seconds=600` (A2) → `start` → wait ≈ 10.5 min → `deferredStop` observed → `start` → marker present (A13) → `stop` → restore 14400 → `start` | AC3 native expiry re-arms and stops |
| L-6 | `reset --confirm <env>` | new marker label; certificate serial unchanged; roles re-created; migrations current |
| L-7 | `deploy --candidate <sha>` again (same candidate) | no reseed, no duplicate pull, versions reported |
| L-8 | D-O2: `retire --yes` (recommended) → absence verification → receipt; or `stop` and record the retained baseline with owner/reason/release condition | AC6; G057 receipt handoff |

Operator time ≈ 2.5–3.5 h including ≈ 1 h of waits (§8.1). Every step is recorded with command, time, outcome; failed runs stay in the Result.

## 5. Ownership, dependency graph, parallel schedule, workers

### 5.1 Writable paths (disjoint; one writer per file; the coordinator integrates)

| Owner | Writable paths (exclusive) | Content |
|---|---|---|
| **W1 infra** | `infra/terraform/**` | modules, example root, lock file, tests, bootstrap script, startup script, README |
| **W2 runtime** | `infra/compose/staging.yml`, `infra/staging/**` (except `check.sh`), `core/internal/config/staging_config_test.go` | overlay, lifecycle scripts (operator + remote), seed, examples (`staging.conf.example`, `app.env.example`), config test |
| **W3 edge** | `infra/docker/edge.Dockerfile`, `infra/edge/**` | edge image, Caddyfile, provenance README |
| **W4 harness+docs** | `infra/validation/run.sh`, `infra/validation/render-customizations.tsv`, `infra/staging/check.sh`, `docs/foundation/STAGING.md`, `docs/foundation/DOMAINS.md` (one paragraph), `docs/foundation/VENDORS.md`, `README.md` (one line) | B-COMPOSE/B-SHELL/B-HADOLINT extensions, offline runner, runbook, index/cross-links |
| **Coordinator** | task branch `g070`, `management/v0/results/G070-*`, allow-file, `.local/{evidence,secrets,tmp}/g070-a1`, deploy manifests, inventory/receipts | single integration owner; runs every Docker/Terraform/Caddy/gcloud command |

Shared contracts settled in this Plan so lanes code in parallel: the Compose variable contract (§2.8), the env rendering (§2.2), the secret names (S6), the Terraform outputs/labels (§2.1), the VM paths (`/srv/<env>`, `/run/<env>/app.env`, `/var/lib/docker` on the data disk), the receipt categories (§2.3), and the check names (§3.3).

### 5.2 Dependency graph and schedule (→ precedes; ∥ parallel)

```
A (parallel from release): W1 Terraform ∥ W2 overlay+seed+examples+config test ∥ W3 edge image+Caddyfile ∥ W4 harness extensions + STAGING.md skeleton
   coordinator: provider mirror prep (network, once) → P0.3 on W1's first handback; P0.1/P0.8 on W2's; edge build + P0.2 on W3's; P0.4 after W2's scripts
B (after A handbacks on g070): W2 remote scripts final (needs W1 outputs + W3 env names) ∥ W4 docs completion (reads A) ∥ W1 README
   coordinator: P0.5 → P0.6 (pool slot) → P0.7 (pool slot + build slot) → checkpoint in the Result
C: coordinator integrates → lint/unit-core rows (§6) → check.sh → safekeeping push → Codex source review → MR (hosted fast + Duo)
D: authorized merge → post-merge affected rows on the frozen merged SHA → live release request (D-O1…D-O4 answered) → pilot L-0…L-8 → Result → receipt → Codex final ACCEPT
```

True dependencies: W2's `remote/*.sh` need W1's output names and W3's env names (declared here, so only the final wiring waits); P0.7 needs W2+W3; the pilot needs P0 complete, the merged candidate and the live release. Reruns are driven by failures/affected checks only.

### 5.3 Workers and isolation
Four writers (W1–W4), **Opus 5.5 high**, each in a native isolated worktree (`isolation:"worktree"`), **stack-free, Docker-free and cloud-free**: they edit, self-review, hand back branches. The coordinator (Sonnet 5.5 high; Opus by owner choice) runs every Docker/offline Terraform/Caddy proof, every private run (two-run pool), every image build (one at a time), the harness rows and — after the live release, under the owner's ADC — every gcloud/terraform/ssh command. Fable 5.1 high for test-failure diagnosis/corrections (xhigh at a safe boundary if needed, recorded). Context 150k checkpoint / 200k handoff per worker; finished worker trees reclaimed after handback once commits are on `g070`. Codex: Plan review, D-C1–D-C3, G057 amendment (§10), source review, Result acceptance.

## 6. Validation (focused executable selectors; AC mapping)

### 6.1 Selectors (`<E>` = `<main>/.local/evidence/g070-a1`; `<A>` = `management/v0/results/G070-scope-allowlist.txt` once Codex imports §7)

| # | Selector (exact) | Checks / AC | Reuse·extend·new (uncovered behaviour for new) | Slot |
|---|---|---|---|---|
| 1 | `bash infra/validation/run.sh --base a5aa46ba80e0ede2837b3673dea1cb4d62c67af6 --allow-file <A> --report-dir <E>/static-<n> lint lint-base` | B-SHELL (+6 scripts), B-HADOLINT (+edge), **B-COMPOSE (+staging overlay assertions, +2 secret scopes)**, B-ENV-EXAMPLE (unchanged), B-MIGRATE-SET, B-DOCKERIGNORE, B-CI-CONFIG, B-SECRET-SOURCE, B-LITERALS (genericity), B-NAMING, B-IDENTITY-GUARDS, B-DEPSCAN, B-SCOPE, B-SECRET-CANARY — **AC1, AC5** | extend (B-COMPOSE: uncovered today = the public overlay's port/network/no-build/digest/no-test-overlay invariants; B-SHELL/B-HADOLINT lists) | none |
| 2 | same, `unit-core` | B-UNIT-CORE incl. `staging_config_test.go` — **AC1, AC2 (G-1 executable)** | new test file (uncovered: the staging template through all four loaders; the production+http refusal as a named pin) | build slot (core-check image, once) |
| 3 | same, `provenance` at `--base HEAD` | B-RENDER-PLAN/B-PROTECTED with the VENDORS register row and the tagged edge FROM lines — **AC1** | reuse | none |
| 4 | `bash infra/staging/check.sh --report <E>/offline-<n>` (S16: TF-FMT, TF-VALIDATE, TF-TEST, EDGE-ADAPT, EDGE-FILTER, LIFECYCLE-DRYRUN, IAM-GREP) | AC1 (Terraform formatting/validation/reviewed plan assertions, provider locks), AC4 (adapted config: ask URL, on-demand scope, filter), AC5 (filter sentinel offline, no literal token), AC3 (dry-run verbs) | new runner script (uncovered: no Terraform or Caddy exists today) — not a harness ID (D-C3) | build slot for the edge image; network none otherwise |
| 5 | P0.6 private DB seed proof (§4.1) | AC3 "no accidental reseed", idempotent seed; roles/RLS unchanged (PROV-1..6 by reference) | extend the private-run pattern; seed new | one pool slot |
| 6 | P0.7 private routing rehearsal (§4.1) | AC4 exact-host routing/404/MCP paths/internal-only `tls-ask` without DNS; AC2 "private ports closed" (no published port but 80/443 inside the run) | new rehearsal (uncovered: the Caddy edge in front of web/core) | one pool slot + build slot |
| 7 | hosted MR `validate:checks` (fast: `lint lint-base lint-web lint-core provenance contracts unit`) + Duo advisory; `validate:integration` (full, since `infra/` is foundation) not requested unless Codex asks | **AC7** | reuse | hosted |
| 8 | post-merge, exact merged SHA, frozen detached checkout: rows 1–4 (rows 5–6 only if the seed/overlay/Caddyfile differ from the tested tree); receipt | **AC7** attributable affected coverage | reuse | as above |
| 9 | live pilot L-0…L-8 (§4.3), after the separate live release | **AC2, AC3, AC4, AC5, AC6** live evidence; one deployment receipt for G057 | new live evidence (never a routine MR check) | owner ADC; pilot cost |

Not selected, reported as unexecuted: `root`, `rename`, `ops`, `off`, `migrations` (no schema/compose-base change), `ci-controls`, `feature-*`, browser suites, the full profile, `validate:integration`.

### 6.2 AC mapping

| AC | Evidence (rows) |
|---|---|
| AC1 generic reproducible infrastructure | rows 1 (B-LITERALS, B-COMPOSE, B-PROTECTED), 3, 4 (TF-*, lock file hashes); the deploy manifest format; pinned edge build provenance |
| AC2 real scoped deployment | row 9 L-1…L-3 (services healthy, migrations/roles/marker, public HTTPS, private ports closed, no test overlay — row 1 asserts the overlay cannot compose `identity.yml`); G057 flows explicitly **not executed** (disclosed); D-C1 disclosure (development loader mode) |
| AC3 safe lifecycle | rows 4 (dry-run verbs), 5 (idempotent seed), 9 L-1n (retry without duplicates), L-4 (marker + certificate survive stop/start), L-5 (expiry re-arms and stops), L-6 (reset scoped), L-7 (deploy without reseed) |
| AC4 trusted DNS/TLS | rows 4 (EDGE-ADAPT), 6 (routing), 9 L-2/L-3 (base + wildcard chain, extra label host without new DNS/cert, unknown-host refusal via internal `tls-ask`, certificate reuse across stop/start, DNS-only/proxied=false, renewal behaviour documented in STAGING.md) |
| AC5 identity/secret/privacy boundaries | §2.1.4 IAM record + L-0 role verification; S6 delivery; row 1 secret matrix; row 4 EDGE-FILTER + IAM-GREP; L-3 sentinel/secret greps; G057 consumes the sentinel evidence by path/sha, then re-runs Q-P1 as its own gate |
| AC6 costs, cleanup, handoff | §8.3 dated cost model confirmed in the console; L-4 stop receipt; L-1n partial-failure proof; L-8 retirement absence proof or retained-baseline record; all-category receipt; one deployment receipt ID for G057 |
| AC7 delivery | rows 7–8; Result AC mapping, failures, receipt; STAGING.md; Codex final ACCEPT |

## 7. Allow-list (B-SCOPE; `management/v0/results/G070-scope-allowlist.txt`, imported by Codex) and exclusions

```
^infra/terraform/
^infra/staging/
^infra/edge/
^infra/compose/staging\.yml$
^infra/docker/edge\.Dockerfile$
^infra/validation/run\.sh$
^infra/validation/render-customizations\.tsv$
^core/internal/config/staging_config_test\.go$
^docs/foundation/(STAGING|DOMAINS|VENDORS)\.md$
^README\.md$
^management/v0/(plans|results)/G070-.*
```

**Excluded (never touched):** `.env.example` (B-ENV-EXAMPLE hash binding), `infra/compose/(local|identity|migrate|browser|secret-scan)\.yml`, every other `infra/docker/*.Dockerfile`, `infra/ci/**`, `.gitlab-ci.yml`, `infra/validation/required-checks.tsv`, `core/**` beyond the one test, `web/**`, `intelligence/**`, `core/migrations/**`, every HTTP/MCP contract, `docs/ARCHITECTURE.md`, `docs/SECURITY.md`, `docs/adr/**`, AGENTS/CLAUDE, G057's Goal/Spec/Plan/Result/allow-list (Codex-owned; §10 is a proposal), the frozen G057 checkout `c840d714`, the M1 preview, the Vestive kit, any production resource. `.local/**` is ignored and never committed.

## 8. Estimates (evidence-based; unknowns marked), cost model, resources, attempt boundaries

### 8.1 Wall-clock by phase (parallel coding / sequential builds / offline validation / live waits / review waits)

| Phase | Parallel coding (per owner) | Sequential builds | Offline validation elapsed | Live waits | Review/queue waits |
|---|---|---|---|---|---|
| A | W1 10–14 h; W2 14–20 h; W3 4–6 h; W4 8–12 h (≈ 2–3 working days wall) | edge image 3–6 min (xcaddy; **unmeasured**); core-check image 2–4 min | P0.1/P0.3/P0.4/P0.8 ≈ 5–10 min total; mirror prep 1–3 min | — | — |
| B | W2 4–6 h; W4 3–5 h; W1 1–2 h | candidate images cold: web 5–10, minio 5–12, core/migrate 2–4, intelligence 2–4 min ≈ **20–35 min** (**unmeasured**; the hosted full job builds them inside 8 GB) | P0.5 3–5 min; P0.6 3–5 min; P0.7 5–10 min (+builds) | — | checkpoint write-up ≈ 1 h |
| C | integration 3–5 h | rebuilds on source change | rows 1–4 ≈ 10–15 min per iteration; expect 2–3 iterations ⇒ **0.5–1 h** | — | hosted fast ≈ 10 min/push; Duo 5–10 min; **Codex source review: unknown** |
| D post-merge | 1 h | 10–20 min on the frozen checkout | rows 1–4 ≈ 15 min (+5/6 if affected) | — | merge authorization: unknown |
| Pilot | 1 h prep | push ≈ 1.5–2 GB (**unknown uplink**, 5–15 min) | — | bootstrap 10–15 min; apply 2–4 min; VM boot + apt/Docker 3–6 min; IAP ready 1–2 min; pulls 2–5 min; DNS ≤ 5 min; staging + production ACME 2–4 min each; L-5 ≈ 11 min; stop/start cycles 3–5 min each; retire 3–5 min ⇒ **≈ 1 h of waits inside 2.5–3.5 h operator time** | live release decision: unknown; final ACCEPT: unknown |

**Totals:** coding ≈ **44–65 h** across four writers; realistic wall-clock to a review-ready MR ≈ **3–5 working days**, plus one pilot half-day after the live release, excluding review latency. **This is a planner estimate, not an accepted deadline.** No split is recommended: the only candidate for added scope (D-C1 option ii) is explicitly deferred.

### 8.2 Resources
Builds sequential (one slot); at most two private runs repository-wide (P0.6/P0.7 sequentially use one); no published port anywhere locally; project names `g070rt-<run>`, `g070db-<run>`; task images labelled `g070.task=G070`; M1 and the frozen G057 checkout never stopped/rebuilt/reclaimed. Inventory reconciled at every handback, batch end, failure and acceptance; compact receipt with every category (containers, images incl. untagged, volumes, networks, worktrees, temp/evidence/secret paths, provider resources, DNS records).

### 8.3 Cost model (dated 2026-10-26; example region `us-central1`; owner region unknown; **every unit price below is lane-reported from aggregators/announcements and marked [unverified] until the owner confirms it in the console/pricing calculator before the live release**)

| Item | Unit price | Running 4 h session | Stopped baseline / month | Retired / month |
|---|---|---|---|---|
| `e2-standard-2` on-demand | ≈ $0.067/h [unverified] | ≈ $0.27 | $0 | $0 |
| External IPv4 in use / idle reserved | $0.005/h (Google notice 2024) / ≈ $0.01/h [unverified] | ≈ $0.02 | ≤ $7.30 (≈ $3.65 if the idle rate is $0.005) | $0 |
| 70 GB `pd-balanced` | ≈ $0.10/GB-month [unverified] | — | ≈ $7.00 | $0 |
| Secret Manager (≤ 12 active versions) | ≈ $0.06/version-month + access ops [unverified] | — | ≈ $0.72 | ≈ $0.72 |
| Artifact Registry (≈ 2 GB) | 0.5 GB free, then ≈ $0.10/GB-month [unverified] | — | ≈ $0.15 | ≈ $0.15 (0 with `--purge-images`) |
| GCS state, Cloudflare DNS (Free plan), Let's Encrypt | ≈ $0 / $0 / $0 | — | ≈ $0 | ≈ $0 |
| **Totals** | | **≈ $0.30 per 4 h session** | **≈ $11–15/month stopped** | **≈ $1/month retired** |
| Forgotten running VM (why `max_run_duration` is mandatory) | | | ≈ $49 + $11 ≈ **$60/month** | |
| Pilot (provision, ≈ 6 VM-hours incl. proofs, one week of disks/IP if not retired) | | | **< $10** | |

### 8.4 Attempt boundaries (`management/BUDGET_POLICY.md` §3–§4, §7)
Attempt 1 = execution release through the review-ready Result (P0, all lanes, inner loops, reruns, MR, post-merge rows **and** the pilot). If the live release is delayed after the source merge, the source/evidence handback is an **interim checkpoint within attempt 1** (the G057 S5 ruling pattern), not a review-ready Goal Result; readiness waits are not attempts. A Codex-authorized bounded correction after the Result review is attempt 2. B5 maximum 5; checkpoint after attempt 3; this Plan spends nothing.

## 9. Risks, rollback/cleanup, integration and closeout

### 9.1 Risks (mitigation)
Duplicate allocation (deterministic names, GCS lock, no `create` in scripts, label+state reconciliation before allocation, L-1n proof); secrets in state/plan/metadata/logs (no secret-shaped variable, env-only provider credentials, tmpfs env file, grep guards in P0.2/L-3, `admin off`, no `docker inspect`/`compose config` in evidence); expiry hitting a live G057 session (`status` prints `terminationTimestamp`; the operator extends by stop → apply → start; 4 h default); data loss on the deferred stop (A13; re-seedable dataset; `stop` is the normal path); LE duplicate-certificate budget (retained `edge-data`, staging-CA rehearsal, `retire` is the only deletion path, D-O2 counts issuances); DNS-01 propagation races (`propagation_delay 20s`, Cloudflare resolvers); TXT litter after a crash (inventory line + manual cleanup note); Cloudflare token over-scope (two zone-scoped tokens, `/user/tokens/verify` status only); Caddy `X-Forwarded-*` (deleted in `header_up`); ZeroSSL failover (recorded issuer); HSTS leakage (300 s, no `includeSubDomains`); MinIO archived upstream (disclosed; R2 is the production target); plaintext hops inside the VM (F-1/F-2/F-3/F-4 remain disclosed local exceptions, not cleared); the owner's machine as the only operator path in v0 (documented; GitLab OIDC later); provider version drift (exact pins, lock file); unmeasured sizing (A8, variable `machine_type`).

### 9.2 Rollback and cleanup (all categories)
- Source: revert the MR (no migration, no data); `local.yml` untouched.
- Provider: `retire` (`terraform destroy` + absence verification); orphan path = label query → `import` → `destroy`; DNS records removed by the same destroy; the `_acme-challenge` TXT cleaned by Caddy or manually (recorded); Secret Manager/AR/bucket/SAs are bootstrap-owned retained items with a documented one-command removal (`bootstrap.sh --teardown`, refusing while any environment state object exists).
- Local: task images (incl. untagged), containers, volumes, networks removed at acceptance; `.local/evidence/g070-a1` compacted to the Result's necessary proof; `.local/secrets/g070-a1` keeps only `staging.conf`/`backend.hcl`/`terraform.tfvars` under named retention (owner, reason: repeat sessions, release: G057 Part B acceptance or retirement) — no secret values are stored there; worktrees reclaimed after handback.

### 9.3 Integration and closeout sequence
1. Execution release → phases A/B (§5.2) → P0 checkpoint → phase C rows 1–4 → safekeeping pushes to `g070` → Codex source review → MR to `staging` (source-branch deletion; hosted fast + Duo; Duo findings dispositioned in the Result).
2. Authorized merge → frozen detached checkout at the exact merged SHA → row 8 → receipt of local cleanup.
3. Live release request with D-O1…D-O4 answered and §4.3 preconditions recorded → pilot L-0…L-8 (every command/outcome in the Result; failures preserved) → deployment receipt ID for G057 → D-O2 disposition.
4. Result: AC mapping (§6.2), D-C1 disclosure, lane corrections carried, attempt log, all-category resource receipt (local + provider), retained baseline with owner/reason/release condition → review-ready handback → Codex final ACCEPT → mandatory acceptance cleanup (worktrees, images, evidence) verified in the receipt.

## 10. G057 Part B reconciliation — bounded proposal for Codex (Codex owns G057 documents; nothing here edits them)

Narrow amendments the accepted Part B Plan r5 needs so it consumes G070's environment (no Part A replan, no edit to frozen `c840d714` or M1, no change to P-1/P-2 steps, the observer, negatives, B3 or the P-3/P-4 holds):

1. **§4.1 readiness table:** "Serving target" → "the retained G070 environment identified by one deployment receipt `<id>`: VM name/zone, static IP, candidate SHA, image digests, edge image digest, `edge-data` certificate serial/not-after, `COMPOSE_PROJECT_NAME`"; "Wildcard certificate" → "Caddy DNS-01 by G070 (D-6 option (i) closed; option (ii) is now the design, with the module/image owned by G070)"; "Certificate for `T`" unchanged (on-demand via `tls-ask`, outbound ACME exists by construction).
2. **§4.2:** X-1 drops `B`/`*.B` (Terraform/G070), keeps `T` CNAME → `custom.B` and `_platform-verify.T` TXT; X-2 is replaced by "verify the receipt's certificate covers `B` and `*.B`"; X-3 becomes "the owner adds versions to the G057-owned `<env>-IDENTITY_WORKOS_{CLIENT_ID,ISSUER,API_KEY}` secrets (never printed) and G070's owner runs `deploy` on the retained environment; no `g057-qual-<run>` project, no `edge.yml`; G057 adds only its observer container (attaching to the resolved `edge` network by name) and its own users/claims"; X-9 becomes "remove the domain row, the six provider entries, per-run users, `T` records; clear the three identity secret versions (disable/destroy versions); hand the environment back to its G070 owner by name with the retained baseline recorded (VM, disks, IP, `B/*.B/custom.B`, `edge-data`) — **no `down -v`**, no host destruction".
3. **§10:** the remote host moves from "named retention … release condition X-9" to "G070-owned baseline; G057 owns only sessions/users/claims/settings/observer container"; `g057.lifecycle=retained` is not applied to G070 objects.
4. **§12 allow-list:** drop `^infra/compose/edge[.]yml$` and `^infra/edge/Caddyfile$`; an edge correction during Part B is a bounded correction request to G070's owner (Caddyfile change → `check.sh` EDGE-* → `deploy`), then G057 re-runs Q-P1.
5. **§4.3 Q-P1:** "reuse G070's AC5 sentinel evidence `<path/sha>` as attribution; re-run Q-P1 on the G057 candidate as the gate" — order unchanged (Q-P1 → observer → Q-A → … → Q-P2 → Q-2).
6. **Receipt consumption:** G057 cites the receipt ID/sha instead of provisioning; a re-deploy of a newer candidate during Part B is a G070 `deploy` producing a new receipt that G057 re-binds; evidence attributable to the old receipt stays distinct.
7. **D-C2 alternative** (if Codex prefers G057's separate env file): G070's `deploy` gains `--identity-env <file>` (second `--env-file`, 0600, G057-owned, never uploaded to Secret Manager); everything else above stands.

## 11. Handback

- Plan path (this file): `management/v0/plans/G070-on-demand-staging-plan.md` — sha256 recorded in the handback message (computed after the final write; the job copy `<home>/.claude/jobs/580cdb2f/tmp/g070/G070-plan-r0.md` is byte-identical).
- Inventories/inputs: Goal/Spec hashes (§0), base SHA `a5aa46ba`, lane report hashes (§0). No other file changed; no `.local/` path created by planning.
- Next action: Codex Plan review (D-C1–D-C3; §10 proposal), import of §7 as `G070-scope-allowlist.txt`, execution release; owner answers D-O1…D-O4 before the later live release.

## Revision history

- **r0 — 2026-10-26 UTC:** initial Plan from the three read-only lanes; base `a5aa46ba`; 13 lane corrections recorded (§0); zero attempts spent; no release claimed. Codex alone reviews and accepts.

## r1 correction — 2026-10-26 UTC (appended; r0 above is byte-unmodified and superseded only where this section says so)

Revision **r1**, author Claude (Fable 5.1, high), isolated planning worktree `agent-<id>`, placed by the harness
at `main` `b40d4a4d`, moved by `git checkout --detach 9178df0fe1b2db678a525c527d33078110fc0306` (HEAD and clean status
verified; the r0 file hashed to `b11cc3e8…b560e` before this append, and lines 1–409 of this file still do). **Still planning
only:** no code, Docker object, Terraform/gcloud/provider/DNS call, install, research, commit or push. Budget **B5, checkpoint
after attempt 3, zero attempts spent.** Input: the binding Codex r0 review (not included)
(CORRECTION LOOP; C1–C4, D-C1–D-C3). Where r1 and r0 disagree, r1 governs; r0 text is retained as history by reference.

### r1.0 Refreshed bindings (supersedes §0's "Staging base"/"Planning checkout" rows and the Spec line in the r0 header)

| Binding | Value |
|---|---|
| Base | `9178df0fe1b2db678a525c527d33078110fc0306` ("docs(G070): review staging plan and clarify bounded deployment profile"; contains `a5aa46ba`). The delta `a5aa46ba..9178df0f` touches `management/**` plus `docs/ARCHITECTURE.md`, `docs/SECURITY.md` and `docs/adr/ADR-009-on-demand-gcp-staging.md` (dated G070 refinement paragraphs only); no `core/`, `web/`, `infra/` file changed, so every `file:line` in §0 F1–F15 holds at the new base. |
| Goal | `management/v0/goals/G070-on-demand-staging.md` sha256 `5e4dcc62fc6c8294ea3df391e321b18ca868416fda65d6b01aee7eb8fde7778e` (unchanged) |
| Spec | `management/v0/specs/G070-on-demand-staging-spec.md` **with the dated r1 addendum**, sha256 `7cfba30d6ae3a277357db1f127bfba1ba618e587158c306c6f7be5b8d97dcaef` (supersedes the r0 binding `e4e6cbc4…ab02`) |
| Codex review | `management/v0/results/G070-plan-r0-codex-review-20261026.md` sha256 `31a601065abe2e113a58cc3aad4957f445d43a98c1f2dffdf4762a0b60ed7303` |
| r0 Plan prefix | lines 1–409 of this file = sha256 `b11cc3e80b8fcd7548546a48b67538079d25ee321136ea3367c01a89197b560e` (proof command in r1.8) |
| Lane reports, G057 inputs | unchanged (§0); frozen G057 Part A `c840d714` (MR !107) untouched; G069 attempt 1, M1 and RC15 authority untouched |
| G069 shared-file ownership (verified in G069 Plan §5 W5) | `infra/validation/run.sh`, `infra/validation/lib.sh`, `infra/validation/render-customizations.tsv`, `infra/ci/profile-paths.tsv`, `docs/foundation/WEB-SHELL.md`, `docs/ARCHITECTURE.md`, `docs/DECISIONS.md` |

### r1.1 Decisions resolved by the review (replace the D-C1–D-C3 rows of §1.3; the D-O rows are replaced in r1.4)

**D-C1 — resolved: bounded synthetic development-loader staging profile** (Spec r1 addendum). Settled profile: `APP_ENV=development`;
real public DNS and trusted HTTPS; `IDENTITY_COOKIE_PROFILE=https` explicit on **both** web and core; backend ports private
(S10/S11); no `PLATFORM_HOST_PORT`, no identity emulator/fixed credentials, no DNS/TLS fixture, no test signer, no
`IDENTITY_WORKOS_API_BASE_URL`; the default real WorkOS adapter and the real `edge-probe` path; `migrate down/reapply` never
exposed (S15). **Not in scope and not a recorded follow-up:** MinIO TLS, a private CA/bundle, a public object-store host or any
repair of production transport gaps — r0 §1.3 D-C1 option (ii) is **withdrawn**, not deferred. **Superseded r0 statements:**
(a) the §1.3 D-C1 sentence "production itself has no gap because R2 is https" is **deleted** — `docs/SECURITY.md` §7 lists other
unresolved transport/at-rest prerequisites (F6's three unsupported TLS hops among them), and HTTPS R2 alone resolves none of them;
(b) S17's disclosure is extended. **Disclosures the Result must carry verbatim in substance:** staging runs the development loader
and is not production readiness (F6 stays "blocked"); the private MinIO signing endpoint yields no externally usable URL — external
signed-download delivery is **unavailable/unqualified in this pilot** (source inspection at review found no non-test HTTP/MCP
caller of `SignedGetURL`; the pilot never exercises one; no public endpoint is created); an identity-unconfigured pilot is
**infrastructure smoke only** — authentication is unavailable, never bypassed; G057 keeps real-provider configuration, the
privacy-first sequence and interactive qualification; **P-4 stays held**. The 32-hex synthetic marker label (S14) is kept as is.

**D-C2 — resolved: G070-owned deployment + G057-owned identity secret versions, one exact deployment receipt.** The Secret
Manager path of S6 is the only identity delivery path; §10 item 7 (second `--identity-env` file) is **withdrawn**; `deploy` gains
no second env-file flag. Receipt identity: `<env>-<candidate sha12>-<n>` with the fields of §10 item 1, written by `receipt`,
published in the Result, cited by G057 by ID + sha256. G070 owns `infra/compose/staging.yml`, `infra/edge/**`, the edge image.

**D-C3 — resolved: unregistered focused offline checks** (`infra/staging/check.sh`, S16). No `B-STAGING*` ID, no count-pin
change, no `ci-controls`, no full local run, no new harness framework; r0's option (ii) is **not scheduled** (a later CI Goal may
propose registration under AGENTS.md's separate-CI rule). Ledger rows stay as §3.3 with the C1 changes below.

### r1.2 C1 — an offline Terraform proof that executes its tests (supersedes §2.1 "Offline commands" bullet 2, A6's fallback, A7's `/run` clause, P0.3, P0.5, and the TF-TEST row of §3.3)

Facts applied: the official `hashicorp/terraform` image has `terraform` as entrypoint, so a shell pipeline needs
`--entrypoint /bin/sh`; `terraform test` discovers only the current directory and the root's `tests/` directory, so module tests
under `modules/*/tests/` are **not** run from `environments/example`; `terraform init` writes `.terraform.lock.hcl` into the working
directory, which a read-only source mount refuses. A plain `debian:12` container does **not** mount `/run` as tmpfs, so r0's P0.5
`mount | grep /run` claim could not have proved anything about the GCE guest.

- **Layout (W1):** `modules/gcp-staging-vm/tests/plan.tftest.hcl` and `modules/cloudflare-dns/tests/plan.tftest.hcl` carry the
  assertions of P0.3 using `mock_provider "google" {}` / `mock_provider "cloudflare" {}` with `run` blocks (`command = plan`; where
  a value is unknown until apply, `command = apply` **against the mock provider** or `override_resource`/`override_data` — in-memory,
  no credentials, no network). The root `environments/example` has no backend-dependent test; a root `tests/` file is optional and
  does **not** count toward the minimum. **No fallback to a real-provider `plan` without credentials on a network-disabled runner**
  (A6's r0 fallback is replaced by the override/mock-apply variant above).
- **Runner (`check.sh`, coordinator):** `<T>` = `<main>/.local/tmp/g070-a1`, `<E>` = evidence dir. Mirror prep unchanged (network,
  once, cache-preparation lock). Per run: `rm -rf <T>/tf-work/<run> && mkdir -p <T>/tf-work/<run> && cp -R infra/terraform
  <T>/tf-work/<run>/src` (disposable writable copy of the **bound** tree; its sha256 over `git ls-files infra/terraform` is recorded),
  `cli.tfrc` as in r0 §2.1. Then, with `RUN='docker run --rm --network none --entrypoint /bin/sh -v <T>/tf-work/<run>:/work
  -v <T>/tf-mirror:/mirror:ro -e TF_CLI_CONFIG_FILE=/work/cli.tfrc -e TF_IN_AUTOMATION=1 -e TF_INPUT=0
  hashicorp/terraform:<pin>@sha256:<digest>'`:
  1. `TF-FMT`: `$RUN -c 'terraform fmt -check -recursive -diff /work/src'`.
  2. `TF-VALIDATE` (root): `$RUN -e TF_DATA_DIR=/work/data/root -c 'cd /work/src/environments/example && terraform init
     -backend=false -lockfile=readonly && terraform validate -json'` (the committed lock must match the mirror; a mismatch is a
     finding, not a reason to drop `-lockfile=readonly`).
  3. `TF-TEST` (per module, isolated data dir): for `m` in `gcp-staging-vm cloudflare-dns`: `$RUN -e TF_DATA_DIR=/work/data/$m -c
     "cd /work/src/modules/$m && terraform init -backend=false && terraform validate -json && terraform test -json -no-color"
     > <E>/offline-<n>/tf-test-$m.json`. Pass criterion per module: `jq -r 'select(.type=="test_summary")|.test_summary'` reports
     `failed == 0`, `errored == 0`, `skipped == 0` **and `passed ≥ 1`**; the ledger row records `modules=2 runs=<N> passed=<N>` with
     the per-module counts. **A zero-run or missing-summary exit is FAIL**, never an "offline PASS".
  4. `TF-GREP` (was IAM-GREP, extended): r0's role/key/DELETE guards plus `provisioner`, `local-exec`, `remote-exec`,
     `metadata_startup_script` absent from `infra/terraform/**`.
  Evidence kept: the two JSON streams, the command lines, the image digest and the copied-tree sha256; `<T>/tf-work/<run>` deleted
  after the ledger is written. **One run per candidate** (rerun only when `infra/terraform/**` changes); no additional campaign.
- **P0.3 (rewritten):** the three TF-* rows above with the P0.3 assertion list of r0; stop criterion unchanged except the fallback.
- **P0.5 (rewritten):** rehearses only the startup script's package pins in `debian:12@sha256:<pin>` (`docker --version`, `docker
  compose version`); the `/run` tmpfs assertion is **removed from the rehearsal**. The guarantee is proved live and enforced at
  runtime: `remote/deploy.sh` and `remote/start.sh` run `findmnt -no FSTYPE --target /run/<env>` and **refuse to write `app.env`
  unless the answer is `tmpfs`**; L-1 records that output. A7 is split accordingly (packages → P0.5; tmpfs → L-1).

### r1.3 C2 — retained-secret and recovery semantics (supersedes S6's "(never on a disk; gone at stop/expiry)", A13's fallback, the §9.1 "data loss on the deferred stop" and "tmpfs env file" phrases, and the seed clause of §2.2/§2.3 `deploy`)

**Where injected values actually live.** `docker compose --env-file` resolves `${VAR}` at `up`; Docker Engine then stores each
container's full configuration, including `Config.Env`, under `/var/lib/docker/containers/<id>/config.v2.json` — on the retained
`<env>-data` disk, for as long as the container object exists, stopped containers included. S13's reboot restart and the deferred
stop both rely on exactly that persistence. So secret values exist on the VM in three places: (a) the 0600 tmpfs `app.env` until it
is deleted; (b) the running processes' environment; (c) the container configuration on the data disk. **Removing `app.env` revokes
nothing in (b) or (c).** The r0 "never on a disk" claim is false and withdrawn; the correct statement is below and goes verbatim
into `STAGING.md` and the stop receipt.

**Controls (existing mechanisms only; no new secret framework):** the data disk is a Compute Engine persistent disk, encrypted at
rest by default with Google-managed keys (L-1 records the disk's encryption state as observed; CMEK is not required); it is
attached only to `<env>-vm`; no command creates a snapshot or image of it; the VM SA holds no disk/snapshot/storage role (S7);
the Docker socket is root/`docker`-group only; interactive access is OS Login over IAP for `operator_principals` only, no SSH keys,
no external port 22 (S2/S10). The principals able to read (c) are therefore the same operator principals who can read Secret
Manager — recorded as such in the AC5 record. `app.env` is now deleted **at the end of every `deploy`/`start` run** (Compose needs
it only during `up`), and additionally by `stop` as a guard.

**Lifecycle (recorded in STAGING.md and §2.3):**
- *Rotation:* add a new Secret Manager version → `deploy --candidate <same sha>` (the rendered env changes → Compose recreates the
  affected containers and removes the replaced objects) → disable the superseded version → verify by the changed
  `com.docker.compose.config-hash` label and container creation time (`docker ps --format '{{.ID}} {{.CreatedAt}}'`), **never** by
  reading `Env`. DB role passwords additionally follow the existing `ALTER ROLE` procedure.
- *Removal/revocation:* a value leaves the VM only when its container objects are removed — `deploy` after rotation (replaced
  containers), `reset` (`compose down` removes every container; volumes kept), `retire` (disk destroyed). Disabling/destroying a
  Secret Manager version stops new fetches but not already-injected values; `stop` **retains** them and its receipt says so
  ("container configuration with injected environment retained on `<env>-data`, encrypted, VM-only").
- *Evidence rule (hard):* no `docker inspect` on the VM without a `--format` that excludes `.Config`/`.Config.Env`; no
  `docker compose config` capture outside `--dry-run`; never `/proc/*/environ` or `config.v2.json`. `check.sh` `TF-GREP` gains a
  sibling `STAGING-GREP`: every `docker inspect` in `infra/staging/**` carries `--format`/`-f`, and no `compose config` appears
  outside the dry-run branch. P0.1's local `config --format json` keeps running only with placeholder values and is deleted after
  assertion (unchanged).

**Recovery after an abrupt stop (A13/§9.1 corrected; "accept reseed" withdrawn).** `start` brings up the stores, waits for
PostgreSQL readiness, runs `migrate status`, and verifies the marker label **before** starting app/edge services. If PostgreSQL
fails recovery, `migrate status` is inconsistent, or the marker is absent on an existing dataset, `start` **holds**: app/edge are
not started, a HOLD receipt is written, and the coordinator escalates to Codex; nothing is reseeded. **Seeding rule:** `remote/
deploy.sh` seeds only when the `<env>_postgres-data` volume **did not exist before this run's first `up`** (checked with
`docker volume ls -q --filter name=^<env>_postgres-data$` before any `up`) — true on first `provision` and after `reset`, false on
every `start`/`deploy` of an existing dataset, whether or not the marker is present. `reset --confirm <env>` remains the **only**
data-discarding path and is an explicit owner/Codex decision recorded in the Result. L-5 keeps its marker/certificate stop/start
and expiry proofs; its outcome after the deferred stop is either "marker present" (PASS) or a recorded HOLD (failure preserved),
never a silent reset.

### r1.4 C3 — authorized live sequence, retention default, AC6 proofs, D-O updates (supersedes §4.3 preconditions/resource list/L-1n/L-8, §5.2 phase D, §9.3 step 3, §1.3 D-O1–D-O4, §8.3 secret row)

**Sequence (strict order; nothing paid or written before step 2):**
1. **Offline readiness:** source merged to `staging`; post-merge rows (§6.1 rows 1–4, 5/6 only if affected) on the frozen merged
   SHA; P0.1–P0.8 PASS; candidate SHA pinned; candidate images **built locally** (build slot; no registry) and their local digests
   recorded. Evidence in the Result before any release request.
2. **Concrete live release** (coordinator proposal → Codex → owner decision): names the GCP project and linked billing account,
   region/zone, environment name, the Cloudflare zone and the two zone-scoped tokens' scopes, ACME contact, the exact resource
   list (r0 §4.3 with **≤ 13 Secret Manager secrets**), the cost model (r1.4 pricing rule below), the cost ceiling, the retention
   default and the retirement timing choice. **D-O1 inputs and D-O2 are supplied/decided here**, not before implementation.
3. **L-0 readiness record (D-O4):** `gcloud auth list`, ADC presence, installed permission rules proven by one allowed read-only
   call, A10 role describe; any native refusal is recorded with the narrow remedy via the standing permission handoff — never
   bypassed, never a blanket mode change.
4. **L-0a bootstrap** (first write): `bootstrap.sh` → receipt (APIs, state bucket, SAs/bindings, AR repository, secret containers
   and the 8 generated versions); owner enters the `CF_API_TOKEN` (and optional `OPENROUTER_API_KEY`) versions.
5. **L-0b publish:** `push` of the locally built images under deployer impersonation → deploy manifest with registry digests.
6. **L-1 … pilot** as r0 §4.3 with the r1 changes below.

**Pricing rule:** the step-2 proposal cites **primary pricing pages** (Compute Engine VM instance pricing, persistent disk
pricing, external IP pricing, Secret Manager pricing, Artifact Registry pricing) with dated figures for the named region; every
figure stays **[unverified]** until read there at proposal time (no research is performed now; §8.3 figures remain lane-reported
and marked). The owner confirms the account, billing and ceiling — pricing research is the coordinator's, not the owner's.

**Retention default (D-O2 corrected):** default = **retain the stopped reusable baseline** (owner intent; Spec r1). r0's
recommendation (i) "retire then provision again" is withdrawn; certificates in `edge-data` are handed to G057 intact — no
destroy/recreate. **L-8 becomes:** `stop` → all-category receipt with the retained baseline (owner: G070 environment owner → G057
during Part B; reason: repeat sessions/G057 readiness; release condition: observed `retire` after G057 Part B acceptance or an
earlier owner decision) → deployment receipt ID for G057. Environment readiness is handed over at this interim checkpoint (§8.4),
which **precedes** final G070 ACCEPT.

**AC6 proofs (minimal; same environment; no second VM, no cleanup framework, no mandatory duplicate provision):**
- **First-allocation partial-failure recovery (replaces L-1n):** the pilot's **first** `provision` runs with a deliberately invalid
  Cloudflare `zone_id` (an explicit `--override zone_id=<nonexistent>` flag, visible in `--dry-run` and in the saved plan). The
  compute module's resources are created; the DNS module fails on the unknown zone; `apply` exits nonzero with partial state. Then:
  `status` reconciliation (label query vs `terraform state list`; the actual partial set is recorded, not predicted) → `provision`
  with the correct `zone_id` → convergence: every already-created resource keeps its ID, only the missing ones are created, label
  query = 1 VM/1 data disk/1 address, 3 records present. Deterministic, and it is the first allocation — not a reapply to a finished
  environment. The orphan case (API success without state write) stays a documented reconciliation procedure, not a simulated proof.
- **Observed scoped retirement (stop is not retirement):** `retire --yes` → `plan -destroy` shown and saved → `destroy` →
  absence verification (instances/disks/addresses/firewalls/subnets/networks label queries empty; Cloudflare read of `B`,
  `*.B`, `custom.B` empty; `terraform state list` empty; address released) → receipt naming the bootstrap-retained set (state
  bucket/object, secrets, SAs, AR repository and images) with owner/reason/release condition. **Timing is selected by the live
  release** (default: after G057 Part B acceptance or an earlier owner decision; the owner may instead choose retirement at the
  end of the pilot, accepting a later re-provision for G057 — not the default). Until it is observed, the Result's AC6 row reads
  "retirement: pending, scheduled on `<condition>`"; **final G070 ACCEPT requires the observed retirement**, while environment
  readiness, the receipt handoff and the interim checkpoint do not wait for it.

**Owner-decision rows (replace §1.3 D-O1–D-O4):**

| ID | r1 status |
|---|---|
| D-O1 | Live-gate inputs supplied at sequence step 2 (project/billing, region/zone, environment name, Cloudflare zone + two scoped tokens, ACME contact, cost ceiling; domain availability confirmed, not purchased). **Not a blocker** to Plan acceptance, execution release or source implementation. |
| D-O2 | Default fixed: retain the stopped baseline; the live release only picks the retirement timing. Not an implementation blocker. |
| D-O3 | **No owner answer required.** 4 h expiry (`session_max_seconds=14400`) and `Strict-Transport-Security: max-age=300` without `includeSubDomains` are configurable defaults documented in `STAGING.md`'s configuration table; the adopter may change them in `staging.conf`/`terraform.tfvars`. |
| D-O4 | **No owner question.** L-0 records actual operator readiness (auth state, installed/loaded permission rules, successful read-only call) and any native refusal with its narrow remedy; no bypass, no permission-mode change. The operator machine is whatever runs the released commands under the owner's ADC. |

**Secret containers:** the upper bound is **13** = 8 generated app secrets + `CF_API_TOKEN` + optional `OPENROUTER_API_KEY` +
the 3 G057-owned `IDENTITY_WORKOS_*` containers (empty until Part B). r0's "≤ 12" in §4.3 and §8.3 is superseded; §8.3's row
becomes "Secret Manager (≤ 13 containers; ≤ 10 active versions at the pilot, ≈ $0.06/active version-month [unverified])".

### r1.5 C4 — G057 boundaries and the shared-file one-writer handoff (supersedes §10 items 2 and 7, §5.1's W4 row, §5.2 lane A for W4, §7's `render-customizations.tsv` line)

**§10 corrections (proposal to Codex, who owns G057 documents):**
- *Item 2, X-3:* the G057 observer attaches to the **resolved private `data` network** (`<env>_data`, resolved by Compose label and
  verified to carry the `postgres` container, as G057 Plan §3.7 already prescribes) — **not `edge`** — using its accepted database
  path (`<db>_migrator` login → `SET ROLE <db>_auth_reader`, read-only). It publishes no port; PostgreSQL publishes none; the
  firewall allows only 80/443 and IAP 22 — **no external DB exposure**. Its `APP_ENV` must equal the observed stack's value, i.e.
  `development` under D-C1 (a one-word G057 amendment). The accepted order Q-P1 → observer → Q-A → … → Q-P2 → Q-2 is unchanged.
- *Item 2, X-9:* "remove **only run-owned additions**: the domain row, the provider entries G057 added — restoring any preexisting
  provider setting it changed to its recorded prior value — per-run users, the `T` records; disable/destroy **only the identity
  secret versions G057 added during Part B**; a version that existed before Part B is left enabled (never destroy a preexisting
  baseline); hand the environment back by name with the retained baseline recorded; no `down -v`, no host destruction."
- *Item 7:* withdrawn (D-C2). G057 Part A evidence (`c840d714`, MR !107) stays immutable; this review released no Part B work.

**Shared files — exact G070 touch list.** Of the seven G069-owned files (r1.0), G070 edits **exactly one**:
`infra/validation/run.sh`, in three bounded hunks — the `lint_shellcheck` file list (`run.sh:331`), the `lint_hadolint` loop
(`run.sh:340`, `+ edge`) and the `static_compose` staging block (`run.sh:700`, §3.1) with their PASS strings. **The
`render-customizations.tsv` edit is dropped:** `docs/foundation/VENDORS.md` is already registered on row 26 and a repeat render
reports "conflict" for it regardless of content, so the reason-text update is not needed for B-RENDER-PLAN; §7's
`^infra/validation/render-customizations\.tsv$` line is removed from the allow-list. No G070 edit touches `lib.sh`,
`profile-paths.tsv`, `WEB-SHELL.md`, `ARCHITECTURE.md` or `DECISIONS.md` (all excluded in §7).

**One-writer handoff (recorded rule):**
1. From execution release, W1 (Terraform), W2 (runtime), W3 (edge) and W4's **G070-owned** files (`infra/staging/check.sh`,
   `docs/foundation/STAGING.md`, the `DOMAINS.md` paragraph, `VENDORS.md` rows, the README line) proceed in parallel, unchanged.
2. The `run.sh` hunks form a **named final sequential lane W4s** (same W4 worker, later phase). W4s starts **only after** the
   coordinator records the handoff: G069's actual integrated contract (the `staging` commit that contains G069's `run.sh`), the
   fetch/rebase of the `g070` task branch onto that tip — **never a rebase of a running proof checkout** — and the exact line
   ranges W4s may touch. Until then nobody in G070 edits `run.sh`.
3. W4s edits only its three hunks; no other G070 writer edits `run.sh` concurrently (one writer); G069's rules are not modified.
4. Validation after W4s = §6.1 rows 1–4 only (`lint lint-base`, `unit-core`, `provenance`, `check.sh`) — **no blanket
   full-suite run**; G069's own merged evidence stands, and G070 reruns none of G069's selectors (its hunks do not change G069
   rules; B-SCOPE runs on the G070 allow-file).
5. If G069 is still not integrated when G070 phases A–B are complete, the coordinator reports it and Codex chooses between
   waiting and a bounded delta (the overlay assertions carried by `check.sh` as a `COMPOSE-STATIC` row, with the two list hunks
   kept minimal at the eventual handoff). A later G069 correction MR triggers reconciliation of the three hunks only.

### r1.6 Estimate deltas (only where the corrections change r0 §8.1; unknowns marked)

| Item | r0 | r1 |
|---|---|---|
| W1 coding | 10–14 h | 12–16 h (per-module tests with mocks/overrides, `--override zone_id` flag, no provisioners) |
| W2 coding | 14–20 h | 16–22 h (volume-existence seeding rule, HOLD path in `start`, `findmnt` guard, rotation/removal text, inspect `--format` rule) |
| W4 | 8–12 h + 3–5 h | unchanged, but the `run.sh` part (≈ 1–2 h) moves to lane W4s; **queue wait for the G069 handoff: unknown** |
| P0.3 offline Terraform proof | inside "5–10 min total" | ≈ 3–6 min on its own (copy + 1 root init/validate + 2 module init/validate/test; **unmeasured**) |
| Pilot live waits | ≈ 1 h | + ≈ 5–10 min (deliberate partial first `provision` + reconciliation + converging re-run); retirement ≈ 10–15 min operator time **when its timing arrives** (outside the pilot by default) |
| Live-release proposal | — | ≈ 1 h coordinator time to read the primary pricing pages and write the proposal (no owner research) |

Totals: coding ≈ **48–71 h**; wall-clock to a review-ready MR still ≈ 3–5 working days **plus** the G069 handoff wait (unknown);
pilot half-day unchanged; final ACCEPT additionally waits for the observed retirement at the release-selected time.

### r1.7 r1 change log (finding → sections; superseded r0 statements)

| Finding | r1 section | Superseded r0 statements (by reference; r0 bytes unchanged) |
|---|---|---|
| Spec binding | r1.0 | r0 header Spec line (`e4e6cbc4…`); §0 "Staging base" `a5aa46ba` and "Planning checkout" rows |
| D-C1 | r1.1 | §1.3 D-C1 row incl. "production itself has no gap because R2 is https" (deleted) and option (ii) (withdrawn); S17 disclosure extended |
| D-C2 | r1.1, r1.5 | §1.3 D-C2 row; §10 item 7 (withdrawn) |
| D-C3 | r1.1 | §1.3 D-C3 row (option (ii) not scheduled) |
| C1 | r1.2 | §2.1 "Offline commands" bullet 2 (`sh -c` without `--entrypoint`, single root run, read-only `/src` for init); A6 fallback (real-provider `plan`); A7 `/run` clause; P0.3; P0.5 `/run` assertion; §3.3 `TF-TEST`/`IAM-GREP` definitions |
| C2 | r1.3 | S6 "(never on a disk; gone at stop/expiry)"; §2.3 `deploy` "seeds only when the marker is absent"/`stop` deletes-at-stop phrasing; A13 fallback "accept reseed"; §9.1 "data loss on the deferred stop … re-seedable dataset" and "tmpfs env file" mitigations |
| C3 | r1.4 | §4.3 preconditions ("images built and pushed", "`bootstrap.sh` run once", "D-O2/D-O3/D-O4 answers" as preconditions); resource list "≤ 12 secrets"; L-1n (reapply on an existing VM); L-8 ("`retire --yes` (recommended)"); §1.3 D-O1–D-O4; §5.2 phase D; §9.3 step 3; §8.3 secret row and "owner confirms every price" framing (owner confirms account/budget; pricing is the proposal's) |
| C4 | r1.5 | §10 item 2 X-3 "attaching to the resolved `edge` network" and X-9 "clear the three identity secret versions"; §5.1 W4 row (`render-customizations.tsv`; parallel `run.sh` edit); §5.2 lane A for W4; §3.2 tsv change; §7 `render-customizations.tsv` line |
| Estimates | r1.6 | §8.1 W1/W2 rows, P0.3 timing, pilot waits; §8.1 totals |

Unchanged and still binding from r0: S0–S5, S7–S16 (S6/S13/S14 as amended above), S17–S19, A1–A5, A8–A12, §1.4 stop conditions,
§2.1 layout/variables/IAM table, §2.2 (except the seed clause), §2.4–§2.8, §3.1, §4.1 P0.1/P0.2/P0.4/P0.6–P0.8, §4.2, §5.3,
§6 (rows and AC mapping, with `check.sh` rows renamed per r1.2/r1.3), §7 (minus the tsv line), §8.2, §8.4, §9.2.

### r1.8 Handback

- Plan path (this file): `management/v0/plans/G070-on-demand-staging-plan.md`; job copy
  `<home>/.claude/jobs/580cdb2f/tmp/g070/G070-plan-r1.md` (byte-identical; sha256 in the handback message).
- r0-prefix proof: `head -n 409 management/v0/plans/G070-on-demand-staging-plan.md | sha256sum` =
  `b11cc3e80b8fcd7548546a48b67538079d25ee321136ea3367c01a89197b560e`.
- No other file changed; no `.local/` path, Docker object, cloud/DNS/provider call, commit or push from this revision.
- Next action: Codex reviews the r1 delta and, if ready, records execution release (source/offline work only); the live
  release is the separate later gate of r1.4 step 2; §10 corrections go to G057's owner (Codex) as a proposal.
- **r1 — 2026-10-26 UTC:** bounded correction of r0 per the Codex review (C1–C4, D-C1–D-C3, Spec binding); base `9178df0f`;
  zero attempts spent; no release claimed.

## r1.9 addendum — per-session VM runtime (appended 2026-10-27 UTC; r0/r1 bytes above unmodified; Spec r2 binding)

Author: Claude coordinator (Sonnet 5.5). **Planning only:** no code for this interface is written until Codex reviews this
addendum. Binding input: Spec r2 at staging `68177597` (owner: default **2 h**, operator chooses a runtime per provision/start
session, native `max_run_duration` with `STOP`, **no Cloud Scheduler or separate shutdown job**, manual `stop` retained, scheduled
stop time shown in status/receipt, app-only deploy never resets a running session's timer, expiry and manual stop both preserve
the retained baseline). The Duo fixes for MR !113 are separate, already inside accepted r1 scope, and proceed in parallel.
Supersedes r1.4 D-O3's "4 h expiry", S8's "default 4 h (14400 s)", §8.3's 4 h cost column and L-5's `600 … restore 14400`.

### r1.9.1 Settled (follows from Spec r2 and the existing design)
| # | Decision |
|---|---|
| R1 | **Default = 7200 s (2 h).** One source: the module/root variable default `session_max_seconds = 7200`; `staging.sh` reads the same number from the optional `staging.conf` key `SESSION_MAX_SECONDS` (default 7200), and a `STAGING-GREP`/`TF-TEST` assertion pins the places (module default, root default, example tfvars/conf, STAGING.md table) to the same value. No second timer, scheduler or poller. |
| R2 | **Interface:** `staging.sh provision … [--duration D]` and `staging.sh start … [--duration D]`; `D` is `<n>m`, `<n>h` or whole seconds (`7200`). Absent `--duration` = the configured default (so a session never inherits a previous session's custom length). `--duration` is accepted only by `provision` and `start`; `deploy`, `stop`, `reset`, `status`, `receipt`, `retire` reject it. |
| R3 | **Bounds (policy, stricter than the provider's 30 s–120 d):** **600 s (10 min) to 28800 s (8 h)**, whole seconds. 600 keeps the L-5 expiry proof (10 min) inside the range; 8 h caps the cost of a forgotten VM at about one working day. Enforced twice: CLI parse (exit 2, before any Terraform/gcloud) and the Terraform variable validation (the module validation tightens from `30…10368000` to `600…28800`; plan tests cover both edges and reject 599/28801/1.5). No clamping. |
| R4 | **Terraform stays the single source of truth.** The only writer of `scheduling[0].max_run_duration[0].seconds` is `terraform apply` of a saved plan; `gcloud … set-scheduling` stays the A2 fallback only if the live proof shows the provider cannot update in place (then Codex reviews the `ignore_changes` exception, as in A2). |
| R5 | **Stopped-VM-only change, no silent reset.** The duration changes only inside `start`/`provision` while the VM is TERMINATED (or on first allocation). Compute Engine arms the deadline at each VM start and clears it at stop (F17), so the running session's timer is untouched by anything else. |
| R6 | Manual `stop` and expiry stop are unchanged and both leave disks, address, DNS, secrets and data (expiry stop is abrupt: A13 recovery as in r1.3). |

### r1.9.2 `start --duration` sequence (Terraform-safe update; each step stops on failure)
1. Parse and validate `D` (R3); resolve `want` = `D` or the configured default; `reconcile` and `tf_init` as today.
2. `vm_status`. Read `have` = the instance's `scheduling.maxRunDuration.seconds` from `gcloud compute instances describe` (the same read `expiry()` already makes) and cross-check it against `terraform output max_run_duration_seconds`; a mismatch is **HOLD** (drift between Terraform and the instance: no apply, no start; the operator reconciles first).
3. **RUNNING:** `want == have` → the existing explicit resume checks, no apply, no re-arm, message "already running; scheduled stop <time>". `want != have` → **refuse (exit 2)**: "the session is running with <have> s; stop first, then start with --duration" (no change, no stop issued implicitly).
4. **TERMINATED, `want == have`:** no Terraform write; `gcloud compute instances start` as today.
5. **TERMINATED, `want != have`:** `terraform plan -input=false -detailed-exitcode -var-file=<private tfvars> -var session_max_seconds=<want> -out=<saved plan>`; `label_check`; then a new **scope check on the saved plan** (`terraform show -json`): exactly one `resource_changes[]` entry with a non-no-op action, it is `google_compute_instance.vm`, its actions are exactly `["update"]` (never `delete`/`create`/replace), and the only differing path is `scheduling[0].max_run_duration[0].seconds` (before == have, after == want). Anything else (other resources, replacement, other attributes) → **HOLD, no apply**, plan kept for review. Only then `terraform apply <saved plan>` (part of the operator's `start --duration`, not a separate release; `--dry-run` prints plan/apply/start commands without running). Re-read the instance: `have == want` or HOLD. Then `gcloud compute instances start` and the usual wait/readiness/`start.sh`.
6. After the start, `expiry()` prints the armed `resourceStatus.scheduling.terminationTimestamp`; `start` records `session <want>s, scheduled stop <timestamp>` in the inventory and the run log.
7. Failure semantics: apply failure → no start, state/instance unchanged or recorded partial (existing `record_state`), the next `start --duration` converges (idempotent: step 4 once `have == want`). Start failure after a successful apply → the new `have` is already recorded; the next `start` re-derives `want` (without `--duration` it applies the default again: one more stopped-VM update), so a session length is always explicit.
8. `provision --duration`: on first allocation `-var session_max_seconds=<want>` rides the one apply (the VM is created RUNNING, so the deadline arms at creation; the `deploy` that follows in the same command happens inside that session). On an existing VM, RUNNING with `want != have` → refuse as in step 3; TERMINATED → the plan must pass the same scope check (an infrastructure change plus a duration change in one plan is refused: apply infrastructure first, then `start --duration`). Every Terraform call from `provision` (without `--duration` on an existing VM), `retire` and `status` passes `-var session_max_seconds=<have>` (derived from the live instance; the default only when no instance exists), so no other command ever plans a spurious scheduling update on a running VM. `deploy`: no Terraform, no `start`, no `set-scheduling`, no `--duration`, never touches the timer.

### r1.9.3 Status, receipt, docs
- `status` and `receipt` add `session`: configured `have` seconds, VM state, and **scheduled stop = `terminationTimestamp` when RUNNING** (`not armed (stopped)` otherwise); `receipt` records it as a `vm-session` row (value, owner "operator", reason "native expiry bound", release "stop/expiry/retire").
- `docs/foundation/STAGING.md`: configuration row `Session length` → default 7200, range 600–28800, per-session `--duration`; the lifecycle table, the expiry paragraph (deadline armed at each start, cleared at stop, changed only while stopped through Terraform), the cost table (2 h session ≈ half the 4 h figure; 8 h ceiling ≈ twice it; a forgotten VM is still bounded by one session length), examples (`terraform.tfvars.example`, `staging.conf.example`) and the default pin.
- `infra/terraform/README.md` input table/default and the "changing the length" sentence.

### r1.9.4 Files (all inside the accepted allow-list; writer after Codex review: one Opus 5.5 high, stack-free)
`infra/terraform/modules/gcp-staging-vm/variables.tf` (default 7200, validation 600…28800), `…/tests/plan.tftest.hcl` (default 7200; 600/7200/28800 accepted; 599, 28801, 1.5 rejected), `infra/terraform/environments/example/{variables.tf,terraform.tfvars.example}`, `infra/staging/staging.sh` (`--duration` parse for `provision`/`start`, the steps above, scope check, `have`/`want` helpers, status/receipt rows), `infra/staging/lib.sh` (duration parser/validator, plan scope-check function), `infra/staging/examples/staging.conf.example` (`SESSION_MAX_SECONDS`), `infra/staging/check.sh` (below), `docs/foundation/STAGING.md`, `infra/terraform/README.md`. No Cloud Scheduler resource, no new Terraform resource, no IAM change (the operator identity that applies the plan already owns the instance update; A2 live-proves it). No contract, schema, product or `run.sh` change; no new check row (the 16 rows are extended).

### r1.9.5 Validation (focused, offline first; reuse/extend)
| Selector (`check.sh` row, extended) | Proves (new cases) |
|---|---|
| `TF-TEST` (existing, mirrored provider) | default 7200 in plan and output; bounds 600/28800 accepted; 599, 28801 and non-integer rejected; `STOP`-only unchanged. |
| `LIFECYCLE-DRYRUN` | `start --duration` dry-runs: parser forms (`90m`, `2h`, `7200`); out-of-range/garbage exits 2 before any command; RUNNING + different duration refuses with no command; TERMINATED + same → no Terraform write, start only; TERMINATED + different → plan, label_check, scope check, apply, re-read, start in that order; `deploy` dry-run contains no `terraform`, no `instances start/set-scheduling/stop` and rejects `--duration`; `stop` rejects it. |
| New fake-Terraform plan fixtures (same row) | scope check accepts the single in-place scheduling update; refuses a replace/delete, a second resource, another attribute, a `before != have`; drift between instance and `terraform output` is HOLD. Mutation check (diagnostic, not committed): weakening the action test lets the replace fixture pass and the row fail. |
| `STAGING-GREP` | the default pin across the places; no `google_cloud_scheduler*` and no scheduler/cron/`at`-job text in the source; STAGING.md says what the code does. |
| Static run `lint lint-base provenance`, `TF-FMT/VALIDATE`, shellcheck/hadolint rows | unchanged coverage on the new head (B-SCOPE against current staging). |

**Live (inside the already gated pilot, no new release):** L-5 becomes: `stop` → `start --duration 600` (the plan scope check runs against the real provider: **A2**) → status shows `terminationTimestamp` ≈ start + 600 s → `deploy` of the same candidate while running, timer unchanged → wait ≈ 10.5 min → `deferredStop` observed → `start` (no flag: applies the 7200 default, one more in-place update) → `terraform plan -detailed-exitcode` = 0 and the timestamp ≈ start + 7200 s → marker present (A13). **A2 (provider in-place update on a TERMINATED instance) stays unproven until L-5**; if the scope check shows a replacement, stop and report: Codex decides between the A2 `set-scheduling`/`ignore_changes` fallback and another design (no silent exception).

### r1.9.6 Risks, estimates
Risks: (a) a duration-change plan includes unrelated drift → the scope check refuses, no apply; (b) a failed `have` read → HOLD, never default-and-apply; (c) the operator picks 8 h and forgets → cost bounded at one session; (d) a stale `-var` in a non-start command → every such command derives `have` from the instance. Estimates (unmeasured): coding 3–5 h after review; offline `check.sh` rerun ≈ 1 min plus static rows ≈ 6 min; review wait unknown; live L-5 about 2 min longer than r1.4.

### r1.9.7 Revision history and next action
- **r1.9 — 2026-10-27 UTC:** per-session runtime addendum per Spec r2; Plan-only; no code, no cloud/provider/DNS action; zero attempts spent by this addendum.
- Next action: Codex reviews this addendum; on ACCEPT one writer implements R1–R6 on the MR branch (same MR, Codex source re-review). The Duo fixes land first or in parallel; the MR stays open and unmerged until corrected code is source-reviewed, affected evidence is refreshed and the Duo threads are resolved.
