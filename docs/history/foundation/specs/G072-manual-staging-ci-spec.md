> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G072 Work Spec — Manual staging and preview CI start/deploy and stop

Revision r1 — 2026-10-27 UTC. Status: ready for later Claude planning; specification only.
Current revision r2 — 2026-10-27 UTC. Original r0 and r1 requirements are preserved; §8 controls environment/candidate scope and §9 records the owner retention policy. Status remains specification only.
Priority: high, matching the [Goal](../goals/G072-manual-staging-ci.md).
References: [G070 Spec](G070-on-demand-staging-spec.md),
runtime Plan review (not included),
[ADR-009](../../../adr/ADR-009-on-demand-gcp-staging.md),
branch workflow (not included).

## 1. Outcome and scope

Expose two deliberate GitLab staging actions: **start/deploy** and **stop**.
Reuse the integrated G070 scripts, Terraform ownership, deployment manifests,
retained VM/stores/certificates, readiness checks and resource receipts. The
operator can launch these actions without keeping a local shell running. The
pipeline controls the same staging environment, not a separate environment per
branch. All project/domain/identity inputs remain generic SaaStacular settings.

This is a follow-up to G070, not scope added to MR !113. Source specification
does not authorize planning, implementation, WIF/IAM setup, GitLab configuration
changes or a live deployment. The later reviewed release identifies actual
project, environment, identities, cloud cost/time scope and retained resources.

## 2. Manual interface and CI routing

**R1 — Two actions, deployment-only mode.** The Plan chooses manual jobs or an
explicit pipeline input selecting start/deploy versus stop. Both are manually
initiated from trusted protected `staging` CI configuration. Preserve ordinary
web-triggered validation when deployment mode is not selected. Deployment mode
excludes validation suites; pressing start/stop does not rerun the application
suite. MR/RC qualification rules, disabled tag/push/post-merge pipelines and
their credential isolation remain unchanged. Unselected manual actions do not
block unrelated pipeline completion; a selected action's failure must still fail
its job/pipeline rather than be hidden by optional manual-job settings. Stop is runnable in a fresh pipeline, not
only as a dependent job after a successful start/build in an older pipeline.

**R2 — Immutable candidate selection.** Start/deploy defaults to the pipeline's
frozen `staging` commit. Accept an optional full SHA or permitted branch reference;
resolve a branch once to a full commit before side effects, verify repository
identity and that the commit belongs to reviewed integrated staging history, then
record/use that SHA throughout. A later branch move cannot change an active run.
Credential-bearing control code comes from the trusted pipeline revision, never
from an arbitrary selected branch, fork, MR or user-provided shell command. The
Plan defines the small adapter needed if G070 currently assumes candidate and
control checkout are identical; no alternate lifecycle implementation is added.

## 3. Start/deploy and stop behavior

**R3 — Start/deploy the selected candidate.** Reconcile existing state and reuse
the approved G070 paths to start/resume the retained VM and deploy the selected
candidate. Reuse validated published digest manifests for that candidate when
available; otherwise build its actual source with existing Dockerfiles and
publish through G070, recording immutable digests. Builds remain sequential;
runner execution must not quietly build current control checkout while labelling
the images as a different SHA. Missing/mismatched artifacts fail before false
deployment success. Report ready only after the selected candidate/configuration
and expected services are observed. A resumed old deployment is not reported as
the requested new deployment. Preserve data and migrations; refuse incompatible
older candidates rather than downgrade schema or reset/reseed retained data.

**R4 — Session duration.** Default 7200 seconds (2 hours); allow an explicit
per-session duration through the approved G070 interface and bounds. Apply it
before VM start, display the actual scheduled stop timestamp, and preserve
G070's running-VM behavior: application-only deployment does not extend/reset a
session or silently power-cycle the VM to change its duration. A request that
needs a stopped VM gives the existing actionable refusal. Compute Engine's
native STOP remains the time limit even when GitLab/the runner is unavailable;
no Cloud Scheduler, cron worker or second shutdown timer is introduced.

**R5 — Independent on-demand stop.** Stop targets the configured environment and
actual deployed state; it requires no candidate input, successful build, previous
pipeline workspace, expiring artifact or selected-commit scripts. Reuse G070's
graceful stop and provider-level stop fallback. A stopped VM is an idempotent
success only after observed state confirms it. Failed, canceled or partly
completed start/deploy must leave stop available; stop does not build, push,
migrate, apply a Terraform change, reset data or destroy retained resources.
Report the observed stopped state and retained-resource receipt. Stop/expiry
retain disks, certificates, DNS, images and other named baseline items with costs.

## 4. Authentication, secrets and ownership

**R6 — Short-lived CI identity.** Use GitLab ID tokens with GCP Workload Identity
Federation and narrow service-account access/impersonation. Enforce trusted
namespace/project, protected staging ref and approved environment in provider
trust/identity conditions as well as CI routing, using supported documented
claims. Token audience/issuer must match the configured provider. No owner ADC
login, stored Google service-account private key, Owner/Editor role or bootstrap
administrator credentials are available to runtime CI jobs. The Plan names the
actual permissions and whether start/deploy and stop use distinct identities;
reuse G070's scoped accounts where suitable. Stop obtains only the credentials
its actual operation needs. WIF creation/IAM bindings remain one-time reviewed
operator/bootstrap actions, not a third routine deployment button.

**R7 — Runner configuration and durable metadata.** Adapt G070's ADC/config
inputs to runner-local short-lived federation files, including Terraform's
containerized credential access. Reuse Secret Manager/private scoped credential
delivery. The Terraform Cloudflare token is available only where its root-plan
refresh is required; Caddy/application/provider secrets remain VM-delivered and
are not made available to the CI build merely for convenience. Reconstruct
non-secret configuration and fetch/store deployment manifests/receipts through
bounded private durable storage using existing G070 formats and cloud resources
where practical. A runner filesystem or short-lived GitLab artifact cannot be
the sole copy required to stop or identify the current deployment. Verify metadata
identity/digests and keep Terraform state/locks separate from ordinary metadata.
No tokens, secret values, raw Terraform state/plan, runtime env dumps or private
keys enter logs, Git or public/job artifacts. Publish only reviewed sanitized
summaries; clear temporary credential/config files on completion/cancellation.

**R8 — One environment manager.** Serialize start/deploy and stop mutations with
the same GitLab resource group. State locking alone does not protect the entire
VM/app lifecycle. The Plan identifies how this excludes concurrent local operator
mutation and handles queued/canceled jobs so an earlier pending start does not
unexpectedly restart an environment after a later stop. Keep locking bounded;
do not add a general scheduler/coordination service. Avoid holding an environment
mutation lock through unrelated validation or image-build preparation. Preserve
G070 recovery on interrupted runs and reconcile before the next mutation. Never
reset the dataset as a cancellation handler. Use the existing run inventory for
pipeline/job/candidate/digests, duration/deadline, actual outcomes and every owned
resource; WIF bindings and persistent metadata have a named retention owner and
retirement procedure. Existing shared/local resource reservations still apply.

## 5. Expected artifacts

Small changes to the existing GitLab configuration, bounded `infra/ci` deployment
entrypoint, G070 credential/config/lifecycle adapter where necessary, and scoped
WIF/bootstrap configuration. Add operator documentation to the existing staging/
operations/Terraform runbooks: initial setup, supported inputs, Play/Stop workflow,
SHA and deadline display, failed-job recovery, local/CI ownership handoff, secret
rotation, residual stopped costs, and disabling/retiring the CI integration.
The Claude Plan provides exact paths and exclusive writers. No application,
tenant model, service contract or production topology change is included.

## 6. Acceptance and focused validation

| ID | Required evidence |
|---|---|
| AC1 — Manual routing | Deployment mode exposes both actions without validation suites; existing MR/RC/ordinary web rules still work; push/tag/schedule/untrusted ref/fork contexts cannot execute these actions or obtain credentials. Reuse/extend the existing CI configuration checks with a focused routing matrix. |
| AC2 — Candidate fidelity | Default pipeline SHA and explicit ref resolution are frozen; invalid/unintegrated/moving refs fail or remain pinned as specified; reused manifests and genuinely built images match the chosen SHA/digests. Malformed/mismatched input refuses before cloud mutation. |
| AC3 — Identity and secrets | Reviewed identity/permission/config diff and offline negative cases for wrong project/ref/audience; one separately released hosted job successfully federates. Prove a forbidden identity context is denied without mutating infrastructure. No secret canary/token/runtime env appears in inspected outputs/artifacts. |
| AC4 — Hosted start/deploy | On the approved retained G070 environment, the manual action deploys the pinned candidate and observes readiness, image identities, configured duration and native deadline; an unchanged published candidate reuses its valid images. Reuse G070's existing expiry/retained-data proof when unchanged rather than repeat a two-hour wait. |
| AC5 — Hosted stop and recovery | Independent manual stop observes VM stopped and retained data/certificates. Focused failure/cancellation/expired-artifact fixtures demonstrate stop independence, no overlapping creators, safe queued-operation order and reconciliation; observed cloud result is not inferred from a green script alone. |
| AC6 — Integration/closeout | Applicable hosted fast and Duo/source review, selected exact-merged-SHA checks, sanitized Result/AC mapping, preserved failures and all-category cleanup/intentional retention. G057 authentication qualification remains distinct. |

The Plan names exact existing selectors and explains any added fixtures. CI
workflow changes warrant their affected routing/auth/lifecycle controls, not an
automatic full ci-controls run or full application suite. MR checks and fixtures
must not provision resources or contact real cloud/DNS providers. Live WIF setup
and the bounded hosted start/deploy/stop proof occur only after their concrete
release and environment ownership window; reuse attributable unchanged G070
evidence. Qualifying source checks are separate from routine deployment actions.

## 7. Non-goals and history

No automatic merge-to-deploy, production release/promotion, arbitrary feature
preview stacks, CI bootstrap/reset/destroy, new hosted scheduler, general pipeline
dashboard or application/auth test campaign. Keep G070's local operator path
available under the same environment ownership rules.

- 2026-10-27 r0 — Owner asks for this follow-up Spec after discussing two manual
  CI start/deploy and stop actions. The latest 2-hour default and manual stop
  direction apply. This publication supplies no planning/execution/live release.

Primary implementation references: [GitLab manual jobs](https://docs.gitlab.com/ci/jobs/job_control/),
[environments and stop actions](https://docs.gitlab.com/ci/environments/),
[GitLab/GCP federation](https://docs.gitlab.com/ci/cloud_services/google_cloud/),
[Google deployment-pipeline federation](https://docs.cloud.google.com/iam/docs/workload-identity-federation-with-deployment-pipelines).

## 8. r1 — Numbered preview environments, 2026-10-27 UTC

Owner direction: after discussing two concurrent environments and work-in-progress
previews, the owner accepts `preview-1`, future numbered slots, and asks to add
this to G072. This amendment supersedes §1's single-environment scope, §2 R2's
integrated-history restriction for preview candidates only, the single-target
assumptions in R6–R8/AC1–AC5, and §7's exclusion of selected implementation
previews. Other boundaries, including specification-only authority, remain.

### R9 — Stable targets and one reusable job pair

| Environment selector | Adopting base hostname | Permitted candidate |
|---|---|---|
| `staging` | `staging.env.foundation.test` | Reviewed integrated staging history, as in R2 |
| `preview-1` | `preview-1.env.foundation.test` | Explicit same-repository implementation/MR branch or full SHA, resolved and pinned under R10 |

Keep `staging` unnumbered. Additional registered preview slots use lowercase
`preview-2`, `preview-3`, etc. Slot identity persists when its deployed branch
changes; no slot is named after a branch and no per-MR environment is created
automatically. G072 initially supports the two registered targets above and a
configuration-driven way to add later slots without copying CI job logic. Actual
new resources always require a separate bounded bootstrap/provision release.
Reuse the same two manual start/deploy and stop actions with a required allowlisted
environment input. Reject missing/unknown selectors before fetching credentials or
mutating resources. The selected trusted profile fixes project, resource names,
state prefix, secret references and hostname; arbitrary pipeline input cannot
substitute a project, state backend, credential scope or configuration path.
All reusable source/examples use generic domains such as `preview-1.example.net`;
the table's Beginify names are adopting inputs only.

### R10 — Selected implementation code without deployment authority

Protected `staging` remains the source of CI control code and environment
configuration for both targets. Preview selection permits an explicitly requested
same-repository implementation branch/MR commit before merge; freeze its full SHA
and record the selection/actor. Fork/external refs and automatic branch/MR-triggered
deployments remain excluded. An unmerged preview SHA is refused for `staging`.
Changing or deleting the source branch after selection does not change the run.

Build the selected application source in a credential-free build context; it
must not receive OIDC/deployer credentials, Terraform/Cloudflare credentials or
host mounts that expose them. The privileged lifecycle uses trusted control
scripts, Terraform, Compose/network/metadata-guard and edge policy, not arbitrary
candidate replacements. The Plan identifies the bounded candidate-build adapter
and compatibility contract; an incompatible candidate fails clearly instead of
running its control code with cloud permissions. Candidate runtime/migrations may
receive only the selected preview environment's required synthetic/nonproduction
application permissions and secrets. They cannot reach staging/production data,
credentials or service identities. Public test-auth bypass remains forbidden.

### R11 — Independent environments and per-environment ownership

Reuse G070's parameterized provisioning with independent VMs, application stores,
data/certificate volumes, network boundaries, Terraform state prefixes,
configuration, secrets/runtime identities, manifests and receipts. The base
hostname and its wildcard (for example `*.preview-1.env.foundation.test`), auth origin
and custom-domain target belong to that environment. A shared GCP project,
Cloudflare zone or immutable image registry is permissible only where the scoped
identity and resource boundary is maintained; no cross-environment data/secret
copy is implicit. Distinct account names or state prefixes alone are insufficient:
the Plan must account for G070's current project-wide compute and bucket-wide
state grants when narrowing preview access. Cloudflare DNS tokens are zone-scoped,
so a shared zone does not provide per-record IAM isolation. Trusted CI/Caddy may
hold the required zone-scoped token; exact environment/record checks protect
sibling names, and candidate builds/application code never receive that token.
Document this privileged DNS boundary accurately. Any authenticated preview
requires its explicitly configured provider app/origins and isolated nonproduction
credentials. Report unavailable
authentication honestly; do not infer G057 live qualification or reuse staging
provider secrets merely to make preview login work.

Each target has its own allowed CI identity scope and `resource_group`/operator
ownership handoff. Operations on one target cannot modify or stop the other;
both environments may run simultaneously. Same-environment mutations still
serialize and pending-start/cancellation rules remain. Shared image-build slots
stay sequential under the existing resource limits; immutable images may be
reused only when candidate and build/config identities match.

Each environment independently defaults to a two-hour session, accepts a duration
through G070's bounds and supports on-demand stop and native expiry. Stop needs
only the environment selector and its trusted current metadata. Switching preview
branches preserves retained-data compatibility checks: incompatible schema/data
requires an explicit separately authorized reset path, never an implicit reset
or migration downgrade in these jobs. Record running and stopped residual costs
per environment. Do not assume the two-hour compute limit removes disk/IP costs.

### R12 — Bounded additional verification and handback

Extend the existing AC matrix with focused cases: environment allowlist/config
substitution; integrated SHA accepted for staging; selected same-repository
unmerged SHA accepted only for preview; candidate branch move remains pinned;
untrusted pipeline/control contexts denied; preview GCP identity unable to access
staging compute/state/secrets; trusted DNS targeting refuses sibling-environment
records; exact resource identities distinguish `preview-1` from `preview-10`;
stopping either target leaves the other's identity/state intact. Verify
credential-free build boundaries and that selected candidate infrastructure
cannot replace trusted orchestration. Reuse
existing routing/federation/lifecycle fixtures rather than create a new framework
or full-suite campaign.

The separately released live proof uses the existing staging baseline and one
explicitly provisioned `preview-1`: both serve at once, preview identifies the
chosen implementation SHA/digests, and stopping preview leaves staging serving.
Observe independent native deadlines; reuse unchanged G070 expiry evidence rather
than wait for two full sessions. Clean/retain the proof resources by exact owner
and approved retention; no `preview-2` resource or per-branch stack is needed.
Document slot selection, branch/schema compatibility, provider-origin readiness,
independent stop, per-environment costs and later-slot registration. The Plan
re-estimates this expanded scope before release. G057/G021 dependencies and the
specification-only phase hold remain unchanged.

Additional reference: [Google pipeline identity practices](https://docs.cloud.google.com/iam/docs/best-practices-for-using-service-accounts-in-deployment-pipelines).
[Cloudflare DNS token scope](https://developers.cloudflare.com/fundamentals/api/reference/template/)
documents the shared-zone credential boundary.

## 9. r2 — Stopped-environment retention policy and G070 follow-up, 2026-10-27 UTC

The owner's retention defaults are defined by the [G070 Spec r3 lifecycle
amendment](G070-on-demand-staging-spec.md#r3--stopped-resource-retention-and-retirement-follow-up-2026-10-27-utc): shared
`staging` is eligible for retirement after 14 consecutive stopped days, with a
hard hold until G057 Part B no longer needs the baseline; numbered previews are
eligible after 48 consecutive stopped hours. The owner may explicitly extend
retention. Starting an environment suspends its stopped-time clock; a later stop
starts a new interval.

Automatic retirement is a separate queued G070 lifecycle follow-up, not part of
G072's two manual CI actions. G072 continues to expose manual start/deploy and
independent stop only; it does not add a scheduler, expiry controller or
destructive cleanup job. Its Plan must consume the reviewed G070 lifecycle
interface when available and must not claim automatic retirement before that
follow-up is implemented and qualified. This addendum changes no live
environment, resource, credential or execution authority.


## Owner amendment — isolated preview WorkOS configuration, 2026-10-28

Owner directs configuring WorkOS for both deployment environments before the
coordinated live validation session. Preview may replace the earlier all-none
identity pins with a complete numeric triple only from its own GCP project and
separate nonproduction WorkOS environment. Staging uses its own complete triple.
Both WorkOS applications are named Deploy; staging's original Local application
is preserved. Display names confer no identity or authority. No staging secret
or version is reused in the preview profile. All-or-none pinning, no latest,
private value handling and failure/permission/isolation proofs remain required.
WorkOS setup and issuer discovery are configuration evidence; live callback,
privacy and cross-project authorization evidence are still required at validation.
