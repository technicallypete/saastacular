> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G070 Work Spec — Reusable on-demand SaaStacular staging

Revision r0 — 2026-10-26 UTC. Initial status: ready for parallel Claude planning.
Current revision r4 — 2026-10-27 UTC. Implementation merged; the owner has released
the live pilot and added operator-token storage to this same G070. Live proof is
in progress. Earlier revisions and holds remain historical below.
[Goal](../goals/G070-on-demand-staging.md) /
[ADR-009](../../../adr/ADR-009-on-demand-gcp-staging.md).

## 1. Outcome, authority and boundaries

Deliver a small repeatable staging deployment on GCP Compute Engine, created by
Terraform and operated through documented commands. One ordinary, non-Spot Linux
VM runs the existing application services with seeded local PostgreSQL, Redis and
MinIO as appropriate. Cloudflare DNS points a configured base/wildcard domain to
the VM; Caddy obtains trusted Let's Encrypt certificates and reverse-proxies web.
A session ends by stopping the VM; resetting data and retiring infrastructure are
separate deliberate operations. No production deployment is included.

The owner's confirmed deployment hostname is `staging.env.foundation.test`. Record it
as this adopting project's input only: all reusable files, module defaults,
commands and examples are generic SaaStacular, using variables and documentation
examples such as `staging.example.net`. Do not bake Beginify names, cloud account
IDs, DNS zone IDs, region, credentials or product identity into reusable tooling.
Domain availability/registration is not assumed or purchased by this task.

Planning is authorized now. Code requires a reviewed Claude Plan and recorded
execution release. Actual paid allocation/DNS/provider changes require the later
concrete environment/action release; preparation should make that decision
reviewable as the last step, not ask the owner to design infrastructure.

## 2. Environment and Terraform ownership

Use one bounded GCP staging project/account scope, one region/zone and one VM.
No Kubernetes, always-running load balancer/NAT gateway or additional managed
services without a concrete requirement in the Plan. Ordinary VM capacity is the
default; Spot is a later option requiring explicit handling of interruption.
Select adequate size from the actual running Compose footprint and image-build
strategy; a lightweight runtime VM must not be assumed capable of large builds.

Terraform owns persistent resource definitions: VM/boot or data disk, network and
firewall, least-privilege service identity, required DNS records and other proven
necessities. Use provider/version locks and a reproducible remote-state backend
with appropriate access/locking, documented one-time bootstrap and retirement.
Separate environment state and identities; staging operations cannot address
production. Lifecycle tooling must not fight Terraform by creating duplicate VMs
or untracked disks/IPs; document how stop/start and refresh reconcile actual state.

Reuse existing run inventory/receipts for task/project/owner, exact provider
resource IDs, labels where supported, DNS record IDs, local Docker/worktree/tmp
allocations and retained state. No new general cleanup framework or mandatory
manifest service. Externally required shared resources are explicitly retained,
not falsely claimed absent. No host/account-wide prune or deletion by label alone.

## 3. Operator lifecycle

Expose a small documented interface; the Plan selects command names and exact
paths. Require these distinct behaviors:

| Operation | Required behavior |
|---|---|
| Provision | Show the exact target/diff, create only the authorized environment, bootstrap dependencies and deploy a pinned candidate. Safely retry partial completion without duplicates or reseeding existing data. |
| Start/resume | Start the retained VM, reconcile its address/DNS, start services, apply safe pending migrations, load existing state and verify readiness. Never silently reset the database, object store or certificate state. |
| Deploy | Select an immutable source/image candidate, reconcile configuration/migrations, perform health checks and report actual running versions. No automatic moving-branch deployment on VM boot. |
| Stop | Gracefully stop the application/database and the VM, verify provider stopped state and record retained resources/cost. Preserve database/object bytes and Caddy certificate/account storage. |
| Reset/reseed | Explicitly scoped destructive reset of this synthetic staging dataset, with clear target identification and secret-safe seed result. Distinct from start/deploy; no production/unrelated DB may be selected. |
| Retire/destroy | Remove the task/environment's VM, owned disks, IPs, DNS and other scoped infrastructure; verify actual absence. Preserve only specifically named shared/bootstrap state or credentials under a reason/release condition. |

A session has an explicit maximum run time and expiry action. Prefer Compute
Engine's native run-duration/stop mechanism when compatible, keeping it independent
of the VM's own process staying healthy. The Plan proves start/resume re-arms the
limit and selects a safe default. Do not introduce a bespoke polling daemon or
new scheduler. Record retained disk/IP/state costs even while compute is stopped;
stop is not claimed to be a zero-cost or zero-resource teardown.

Default database and certificate volumes are intentionally retained for repeat
sessions under the environment owner until reset/retirement. Crash/reboot/partial
apply recovery reconciles actual provider and Docker state before new allocation.
Failed cleanup of disposable items holds new consuming work; intentional staging
retention has named ownership, cost and a release condition. Keep this distinction
in the ordinary all-category receipt rather than adding an approval loop per stop.

## 4. Application, database and seed configuration

Reuse existing service Dockerfiles, migrations, DB role bootstrap and Compose
network boundaries. Prefer bounded staging overlays/files instead of copying a
second application stack. Publish only required public edge ports (80/443); keep
PostgreSQL, Redis, MinIO administration and core/internal endpoints private.
Preserve web/core/intelligence network separation and Host/Origin forwarding
semantics. Do not expose the local development ports wholesale to the Internet.

The Plan must verify the actual deployable-image path. Existing Dockerfiles do
not prove that a production image registry already exists. Choose digest-pinned
published images or a bounded reproducible build of an exact source commit using
existing tooling; record source/config/image identities. Avoid an unrelated CI
publishing platform or compilation embedded inside uncontrolled Terraform
provisioner retries. Existing MR/RC pipeline triggers remain unchanged.

Run supported production-mode configuration for the public environment: real DNS,
trusted HTTPS and the real WorkOS staging adapter, with separate staging secrets.
Never use the test-only identity overlay/emulator, fixed credentials, DNS/TLS
fixtures or test signers as public authentication. First inventory actual startup/
adapter requirements and report any production-mode gap before claiming readiness;
G057's retained P-4 production-adapter boundary is not silently cleared.

PostgreSQL remains on the VM for staging. Initialize roles and current migrations
using existing bootstrap; the runtime role is not owner/superuser/BYPASSRLS.
Seed synthetic, clearly identified Organizations and representative data through
an idempotent supported path. No real customer data, public test-auth bypass or
hardcoded real-provider identities. Separate one-time bootstrap from later boot;
existing data survives stop/start and deploy. Brand/domain label data follow the
current accepted schema. Provider users/sessions and actual signup remain G057
qualification activities, with explicit identities and cleanup attribution.

Redis retains its existing ephemeral semantics unless a documented workload
requires durable state; MinIO data remains local/private for staging. External
storage endpoints remain configurable so later production can use managed
PostgreSQL and private R2 without application provider names or state assumptions.

## 5. DNS and HTTPS

Parameterize the platform base `B` (example `staging.example.net`), shared auth
`auth.B`, custom-domain target `custom.B` and wildcard `*.B`. Configure both base
and wildcard DNS toward the actual serving address; specific records take
precedence. Reserve infrastructure labels under the existing hostname rules.
No per-tenant DNS record or certificate is required for generated platform labels.
The application itself supplies opaque labels; deployment does not replace them
with friendly slugs or expose tenant lookup lists.

Choose address retention versus DNS update on restart explicitly, with its cost
and readiness implications. Do not claim a stable ephemeral VM IP across stop/start.
The configured DNS/edge path must preserve actual Host values, secure cookies,
exact-origin sessions and G057's shared callback relay.

Caddy terminates trusted TLS, redirects HTTP appropriately and obtains/renews
certificates covering **B and *.B** through Let's Encrypt DNS-01 validation.
Pin the Caddy image and Cloudflare DNS module; document its small build/publishing
path and provenance. Credentials are zone-scoped and privately supplied; no
provider-wide key or committed token. Keep Caddy data/account/key storage persistent
across normal stops. On resume, renew/recover as needed before reporting HTTPS
ready. Test ACME wiring with the CA's testing endpoint when useful, then use real
browser-trusted certificates for the actual staging session and live proof.
Cloudflare can remain DNS-only; do not assume its standard zone certificate
covers nested `*.staging.example.net` hostnames.

Custom-host certificates follow the existing G016/Caddy authorization design;
`tls-ask` is internal and accepts only currently verified allowed hosts. No open
on-demand issuance for arbitrary names. G070 supplies the reusable edge and its
minimal readiness smoke; G057 owns actual custom-host claim/provider/browser
qualification and its privacy-first sequence. Agree one owner for shared Caddy/
Compose files rather than implementing two overlays independently.

## 6. Identity, secrets and operational evidence

One scoped deployment identity and a separately scoped runtime identity where
needed; avoid Owner/Editor blanket roles or long-lived cloud service-account keys.
Prefer operator authenticated short-lived credentials initially; if GitLab-driven
operations are proposed, use scoped OIDC/workload federation and manual authorized
triggers, not a new automatic deployment on every merge/tag. Record actual IAM
permissions and the supported secret source in the Plan.

No secret values, DB passwords, DNS tokens, WorkOS keys, session/callback values,
private certificate keys or raw sensitive state in Git, Terraform outputs/plan
artifacts, instance metadata/startup scripts, console output or screenshots.
A Terraform `sensitive` flag alone does not prevent a value entering state. Use
secret references and a private runtime delivery mechanism; keep state and local
artifacts access-controlled. Pin readiness metadata without dumping environments.

Apply G057's callback-query privacy requirements at the edge before any real
provider browser flow, including access/error/issuance logs. Preserve the required
sentinel proofs and privacy-first order; don't count a hosted login redirect as
a completed real login. Minimal health/DNS/certificate smoke belongs to G070;
provider application changes and real G057 flows keep their existing readiness gates.

## 7. Relationship to G057 and future production

G070 owns reusable hosting/lifecycle; G057 Part B is deferred to this environment.
Its P-1/P-2 requirements, observer/security checks, manual provider/browser
participation and failure history remain in G057. Narrowly reconcile the accepted
Part B Plan to consume G070's exact deployed candidate/config and account for
retained infrastructure; do not replan/retest completed Part A just for this change.

Environment readiness can unblock G057 proof while G070 closeout proceeds under
one coordinated resource handoff; neither final acceptance depends on the other's
final report circularly. G057 releases its own sessions/users/test claims/temporary
settings and records any named retained baseline owned by G070. Do not destroy the
reusable environment merely to reproduce the old disposable-host cleanup receipt,
or misreport retained resources as removed. Both delivery outcomes still gate G021.

Production direction is documented only: the same configurable application/edge
approach on always-on compute, managed PostgreSQL (e.g. Cloud SQL in the same
region, with backups/recovery configured), and private Cloudflare R2 for bytes.
Preserve DB role/RLS and tenant storage boundaries; actual Cloud SQL/R2 compatibility,
backups/restore, monitoring and availability requirements need later production
qualification. No production Terraform apply, managed DB/R2 provisioning, migration
of production data or HA claim is part of G070.

## 8. Required validation and acceptance

Claude's Plan supplies executable selectors, exact input/resource bindings,
minimal operator readiness and costs. Reuse existing checks; add only focused
coverage of changed lifecycle/configuration behavior. No new harness framework,
mandatory full local suite, full ci-controls, or mandatory UI screenshot campaign.

| AC | Required evidence |
|---|---|
| AC1 — Generic reproducible infrastructure | Parameterized domain/project/region/identity inputs, provider locks, Terraform formatting/validation and a reviewed scoped plan; a reproducible pinned app/Caddy build or image path. No Beginify literals in reusable defaults/code. |
| AC2 — Real scoped deployment | One authorized GCP pilot of the exact candidate; expected services healthy, synthetic data/migrations/roles correct, public base/tenant HTTPS ready, private ports closed and no production/test-auth crossover. Honest distinction between observed and unexecuted G057 flows. |
| AC3 — Safe lifecycle | Focused provision/retry, deploy, stop/start and reset proofs demonstrate no duplicate allocation or accidental reseed; a benign dataset marker and certificate storage survive normal stop/start; native expiry re-arms and stops the VM as documented. |
| AC4 — Trusted DNS/TLS | Base plus wildcard trusted HTTPS, an additional generated-label host without new DNS/cert allocation, certificate state reuse/recovery, exact host routing and internal-only issuance permission. DNS-only/proxy choice and renewal behavior documented. |
| AC5 — Identity/secret/privacy boundaries | Scoped IAM and secret delivery, existing DB roles/network separation, no credentials or callback material in public outputs/state/metadata; pre-flow sentinel/privacy proof reused with exact attribution when G057 runs. |
| AC6 — Costs, cleanup and handoff | Actual running/stopped/retired cost model with dated assumptions; live stop and isolated retirement/partial-failure cleanup proof plus all-category inventory. Reusable retained baseline has owner/reason/release condition; G057 consumes one exact deployment receipt without duplicate provisioning. |
| AC7 — Delivery | Applicable hosted fast/source/Duo, standing authorized staging merge, affected exact-merged-SHA checks and reusable operator docs; Result AC mapping, failures, resource receipt and Codex final ACCEPT. Production is a documented target only. |

Offline Terraform validation/Caddy validation/rehearsals precede any live pilot;
paid/API operations are not smuggled into routine MR checks. A small pilot may
serve both lifecycle evidence and G057 environment readiness when inputs match;
reuse unchanged evidence rather than provision twice. Preserve a deterministic
contract failure as failure even if a subsequent attempt passes. Final G070
ACCEPT requires the selected live lifecycle evidence, not a fictional dry-run PASS.

## 9. Sources and revision history

Primary technical references checked 2026-10-26:

- [Compute Engine start/stop and retained charges](https://docs.cloud.google.com/compute/docs/instances/stop-start-instance).
- [Native VM run-time limits](https://docs.cloud.google.com/compute/docs/instances/limit-vm-runtime).
- [Caddy automatic HTTPS, storage and DNS challenges](https://caddyserver.com/docs/automatic-https).
- [Cloudflare DNS module](https://github.com/caddy-dns/cloudflare).
- [Let's Encrypt DNS-01](https://letsencrypt.org/docs/challenge-types/#dns-01-challenge).
- [GitLab/GCP workload identity](https://docs.gitlab.com/ci/cloud_services/google_cloud/).
- [Existing object-store qualification limits](../../../foundation/OBJECT-STORAGE.md#production-configuration-cloudflare-r2-notes-not-qualified).

2026-10-26 r0: owner authorizes the named G070 Spec and planning, confirms
staging.env.foundation.test while requiring generic SaaStacular tooling, and defers
G057 Part B to the reusable environment. No live resource or implementation
release is made by this Spec publication.

## r1 — Bounded staging profile refinement, 2026-10-26 UTC

Plan r0 review (not included) found that the
current production loader rejects the required private HTTP MinIO endpoint.
This dated refinement supersedes §4's requirement to run the public synthetic
staging environment in production loader mode. It does not change production
requirements or release implementation/live actions by itself.

G070 may use `APP_ENV=development` with real public DNS/trusted HTTPS, explicit
`IDENTITY_COOKIE_PROFILE=https` on both web and core, the default real WorkOS
adapter, private backend endpoints and the real edge-probe path. No identity
emulator, fixed test credentials, DNS/TLS fixtures, test signer or public auth
bypass is allowed. Lifecycle commands must not expose destructive migration
down/reapply operations. Keep existing production refusals and readiness evaluator
unchanged; disclose unresolved internal transport/at-rest qualification. Staging
readiness is not production readiness, and HTTPS R2 alone would not resolve the
production gaps in SECURITY §7.

An identity-unconfigured pilot may prove infrastructure/lifecycle only, with
authentication unavailable rather than bypassed. G057's actual real-provider
configuration, privacy-first sequence and interactive evidence remain required
before claiming authenticated staging qualification; its P-4 boundary remains.
The private MinIO signing endpoint does not supply externally usable download
URLs. External signed-download delivery remains unavailable/unqualified for this
pilot; no public MinIO endpoint or TLS/private-CA expansion is required here.

Runtime secret delivery must accurately disclose private Docker container
configuration retained on the encrypted, access-controlled data disk. A tmpfs
env file does not prevent Docker persisting injected environment values or erase
them when the file is deleted. Record access, version/rotation and container/disk
removal semantics without environment-bearing evidence dumps. Existing bans on
secret values in Git, Terraform state/outputs, metadata and public evidence remain.

Stop/expiry recovery never silently reseeds preserved data. Reusable stopped
baseline retention is the default; real scoped retirement/partial-failure proof
still gates AC6. Its authorized lifecycle timing may follow G057's use of the
environment, so readiness does not require premature destruction or a second VM.
Source/offline work waits for Plan approval, while bootstrap, registry writes,
DNS and provision follow the separate concrete live-action release. Account
inputs are required for that live phase, not for completing the Plan correction.

## r2 — Per-session VM runtime selection, 2026-10-27 UTC

The owner sets the default VM runtime to **2 hours**, superseding the Plan's
previous 4-hour default and the owner's earlier same-day direction to keep 4
hours. The operator can choose a different VM runtime for each staging session
when provisioning or starting the VM. Keep Compute Engine's native
`max_run_duration` with termination action `STOP`; do not add Cloud Scheduler or
a separate shutdown job. Apply the selected limit before that VM start and show
the scheduled stop time in operator status/receipt. The Plan defines the
input/CLI, validation bounds, and how Terraform remains the source of truth when
changing the scheduling property on a stopped VM. An application-only deploy
does not silently reset a running session's timer. The existing on-demand
`staging.sh stop` operation remains available; both manual stop and expiry stop
preserve the retained baseline and do not destroy disks, DNS, secrets or
application data.

This owner refinement does not release paid allocation, DNS/provider writes or
the live pilot; those still require the concrete live-action release.

## r3 — Stopped-resource retention and retirement follow-up, 2026-10-27 UTC

The owner selects automatic retirement defaults for stopped environments,
superseding the earlier no-auto-delete preference while preserving the accepted
G070 implementation's existing distinction between stopping and destroying:

- **Shared `staging`:** review the stopped pilot within 7 days. The default is
  retirement after 14 consecutive days stopped. Do not retire while G057 Part B
  still depends on this baseline; that is a hard retention hold through Part B
  acceptance and handoff. If the 14-day threshold passes during the hold, retire
  at the first safe expiry check after the hold is released. The owner may retire
  earlier or explicitly extend retention.
- **Numbered previews:** default retirement after 48 consecutive hours stopped.
  Starting a preview suspends its stopped-time clock; the next stop starts a new
  interval. The owner may explicitly extend retention before expiry.
- **Retirement scope:** remove only the exact environment's owned VM, disks,
  addresses, DNS and other environment resources, with identity verification,
  locking, a failure-safe receipt, and preservation of shared/bootstrap-owned
  resources. The existing 2-hour native VM expiry still only stops the VM; it
  does not retire these resources.

This policy requires a separate **G070 lifecycle follow-up** to define and
validate the automatic expiry controller, schedule/trigger, retention holds and
extensions, locking, exact-owner cleanup, retries and receipts. G072 remains the
manual start/deploy and stop interface; it does not add or implement that
controller. This amendment records policy and queued follow-up scope only; it
does not authorize planning, implementation, cloud/DNS changes or retirement.


## r4 — Operator Terraform credential in Secret Manager, 2026-10-27 UTC

The owner's instruction is: "it shouldn't be an additional setup followup thing.
let's just add it now to gcp, and you have it from now on." The owner repeated in
the coordinator session that the Terraform secret should be in GCP "so you always
have it", and explicitly asked Codex to add the task's native Bash allow rule.
This adds the following bounded operational scope to the current G070 release;
it is not a separate Goal, follow-up, source rewrite or new planning campaign.

- Store the existing Terraform Cloudflare operator token in one distinct Secret
  Manager secret in the selected staging project/region, outside Terraform
  variables, data sources, state, metadata and application environment files.
  This is **one additional operator secret**, separate from the up-to-13 existing
  runtime/bootstrap secret containers. For this adopting environment its ID is
  `staging-TERRAFORM_CLOUDFLARE_API_TOKEN`; reusable source remains generic.
- Give the existing authenticated operator access to that one secret. Grant no
  new runtime-VM or application access and no project-wide Secret Accessor role.
  Keep it out of bootstrap's runtime-secret loop; Caddy's `CF_API_TOKEN` remains
  separate. Existing inherited administrator authority is not removed or hidden.
- Upload once from the private owner file without printing the value or placing
  it in arguments/logs. Reuse the same owned secret/version on later runs; fail
  on uncertain existence, failed version reads, ownership/access mismatch or
  invalid/disabled/inaccessible versions. Do not silently duplicate or overwrite
  unrelated credentials.
- The operator helper fetches the token with the existing Google authentication
  and supplies it only to the command requiring `CLOUDFLARE_API_TOKEN`. This can
  use the current private operator driver with the source CLI's existing env
  interface; modifying `staging.sh` or rebuilding application images is not
  required merely to centralize this credential. Pin and account for the helper.
- Verify the remote value equals the supplied value without printing either,
  verify the secret/version and scoped explicit IAM, and record ownership and
  retention in the existing inventory/Result. Keep the local owner file until
  the owner chooses its removal or rotation. Retain the operator secret for
  repeated deployments until explicit environment teardown or owner removal.
- The owner-requested native allow rule is limited to the reviewed G070 operator
  helper. Install, loaded configuration, and successful native execution are
  separate facts. Preserve all refusals; change no general mode, hooks, deny
  rules or managed policy, and do not use another executor to clear a refusal.
  Remove the temporary task rule at closeout unless the owner retains/replaces it.

The current frozen deployed candidate `74f9acd9` and existing seven images remain
bound. Any later reusable source change still needs bounded review and impact
reconciliation; this amendment does not authorize quiet candidate substitution.
A value-free upload/access proof and focused helper checks suffice for this
operator-only addition. All other pilot, budget, security, retention, G057 Part B
and final acceptance requirements continue.
