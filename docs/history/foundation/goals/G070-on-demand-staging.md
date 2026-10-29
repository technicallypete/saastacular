> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G070 — Reusable on-demand SaaStacular staging

Final status: ACCEPTED 2026-10-29; see the dated final disposition in the existing Result. Earlier status text below is historical.

Date: 2026-10-26 UTC
Status: source MR !113 merged; owner live release granted; pilot in progress. Spec r4 adds operator-token storage to this G070; final acceptance remains pending.
Owner: Codex Goal/Spec/architecture/review; existing Claude coordinator Plan/implementation/validation/Result.

## Objective

Provide a repeatable, configurable GCP staging environment: Terraform provisions
one ordinary Compute Engine VM, Docker Compose serves an exact application
candidate with seeded local PostgreSQL and object storage, Cloudflare provides
DNS, and Caddy/Let's Encrypt provides trusted root/wildcard HTTPS. Start for a
testing session and stop afterwards, preserving explicitly retained state. Support
an intentional reset and full retirement as separate operations.

The owner approves this named Spec/planning task with "ok start on 070 spec etc".
The owner confirms `staging.env.foundation.test` but directs that the implementation stay
generic to SaaStacular; actual names are deployment inputs, never source defaults.

## In Scope

The [Work Spec](../specs/G070-on-demand-staging-spec.md): infrastructure/configuration,
bounded lifecycle commands, reproducible application deployment and safe seeding,
DNS-validated wildcard certificates, secret handling, resource/cost accountability,
automatic run expiry, focused validation, and operator documentation. Preserve a
production extension path using managed PostgreSQL and private R2, documented only.

## Out of Scope

Production provisioning, a managed database/R2 rollout now, Kubernetes, a new
platform scheduler, application feature changes, automatic release promotion,
per-tenant infrastructure, or a second implementation of G057 live qualification.
No Spot/preemptible capacity initially. Protect local M1 and frozen qualification.

## Dependencies

Initial planning input: staging `8e02a6be2b690fe5a6a19f041c49a96d983725cd`; fetch/reconcile actual later inputs
without modifying G069 workers or frozen G057 `c840d714`. Planning overlaps G069.
Development can proceed independently after Plan/release; the real pilot pins an
accepted staging candidate and reconciles any later integrated G069 UI changes.
No dependency on final G057 acceptance: G070 supplies its Part B environment.
G057 Part B remains deferred until environment/readiness/live-action gates clear.
G070 and G069 delivery plus final G057 acceptance precede G021; the foundation
barrier and independent RC promotion authority remain.

## Completion Definition

All Spec ACs have attributable source and lifecycle evidence, including one bounded
real staging provision/deploy/stop/start/retire proof when account/action readiness
is recorded. Codex reviews the implementation, Result, selected merged-SHA coverage
and cleanup/intentional retention. Offline rehearsal alone is not live acceptance.
G057's actual WorkOS/custom-domain qualification is a separate handback.

## Execution Policy

Architecture/infrastructure tier. Claude writes one Plan with actual paths,
resource identities, lifecycle/secret/state design, exact selectors and costs.
Codex Plan ACCEPT/release precedes implementation. Planning starts now with no
cloud/DNS/WorkOS writes or paid allocation. Prepare a concrete account/region/
resource/cost plan before the later live-action release; no repeated permission
for actions already covered by that exact release. This Spec authorizes neither
account purchases nor production actions. Use current model/effort policy.

## Iteration Budget

B5: five shared Goal execution attempts; checkpoint after attempt 3, under
BUDGET_POLICY (not included). Planning spends zero attempts. One
attempt includes its implementation/validation inner loops; parallel workers
share the allowance. Do not replenish G057's separate budget or rewrite its runs.

## Context / Scope Budget

One staging environment and one GCP project/account scope, reusing existing service
images/Compose/bootstrap code. Avoid unrelated cloud product comparisons or a
production rollout. Planner estimates coding, sequential builds, DNS/certificate
waits and live operator time separately; recommend a split only for actual added
scope such as an unavoidable application/provider adapter change.

## Suggested Subagents / Delegation

Parallel planning research: (1) GCP Terraform/IAM/state/lifecycle/cost; (2) Compose,
images/seeds/secrets and current production-mode fit; (3) DNS/Caddy certificates
and the G057 handoff. One lead integrates one Plan and one coordinator remains.
Implementation uses dependency-ready disjoint infrastructure, deployment and
edge/docs lanes with one owner per artifact. Reconcile shared docs with G069 at
a safe boundary; do not duplicate Caddy ownership with G057. Keep local build/
test reservations, named remote ownership and continuous cleanup.

## Stop / Escalation Conditions

Missing exact account/action authority, unsafe Terraform/state/destroy changes,
credential leakage, production/unrelated resource impact, unowned retained cost,
a bootstrap that weakens DB RLS/identity/network boundaries, native refusal,
actual code/adapter gap outside scope, or budget exhaustion. Stop affected work
and return concrete smallest options. Preserve unaffected safe work and evidence.

## History

- 2026-10-26 — Owner selects GCP on-demand staging, Terraform, Cloudflare DNS and
  Caddy/Let's Encrypt; confirms staging.env.foundation.test as a configurable deployment
  input; authorizes G070 Spec and planning alongside G069. G057 Part B is deferred
  to the resulting environment, not waived or marked passed.
- 2026-10-26 — Plan r1 ACCEPT/source release (not included):
  zero planning attempts; execution starts attempt 1 under B5/checkpoint3.
  Independent lanes may run alongside G069, with its shared-file ownership held
  through the explicit handoff. Cloud/DNS/provider actions await the later exact
  account/resource/cost release; this is not final Goal ACCEPT.
- 2026-10-26 — Owner subsequently queues [G071](G071-default-tenant-root-alias.md)
  before G070. Preserve already-dispatched worker checkpoints and frozen proofs;
  hold new G070 phase dispatch/integration until G071's applicable predecessor
  gates clear. Accepted Plan/source evidence and budgets are not restarted. Later
  reconcile the optional config setting through the existing deployment path.

## Owner amendment — stopped-resource retention follow-up, 2026-10-27 UTC

The owner sets a 14-consecutive-stopped-day default retirement for shared
`staging`, with owner review within 7 days after the pilot stops and a hard hold
until G057 Part B acceptance/handoff. Numbered previews default to retirement
after 48 consecutive stopped hours; an owner may extend either retention period.
This supersedes the earlier no-auto-delete preference, not the accepted
implementation's stop-versus-retire contract or 2-hour VM-stop behavior. Record
automatic enforcement as a separate G070 lifecycle follow-up, with G072 retaining
the manual CI actions only. See the [G070 Spec r3 amendment](../specs/G070-on-demand-staging-spec.md#r3--stopped-resource-retention-and-retirement-follow-up-2026-10-27-utc).
No planning, implementation, live action or resource-retirement release is made.


## Owner amendment — current-task operator token storage and native rule, 2026-10-27 UTC

The owner explicitly directs storing the existing Terraform Cloudflare token in
GCP now, within G070, for reuse on future deployments. It is one additional
operator-only Secret Manager secret with access scoped to the existing operator;
no VM/application binding or project-wide grant. The private operator helper may
fetch it through existing Google authentication and pass it to the existing CLI
environment interface, keeping the frozen candidate and images unchanged. This
is not a new follow-up or a requirement to rewrite the source deployment tool.
See [Spec r4](../specs/G070-on-demand-staging-spec.md#r4--operator-terraform-credential-in-secret-manager-2026-10-27-utc).

The owner separately instructed Codex in the coordinator session to add the
narrow native Bash allow rule for that reviewed helper. Preserve all other
permission settings and verify loading and actual operation separately; native
refusals and all substantive resource/action limits remain binding. The owner
live-release decision is recorded in the current review (not included).
