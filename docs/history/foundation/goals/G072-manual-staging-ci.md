> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G072 — Manual staging and preview CI start/deploy and stop

Final status: ACCEPTED 2026-10-29; see the dated final disposition in the existing Result. Earlier status text below is historical.

Date: 2026-10-27 UTC
Status: Spec r2, including numbered previews and the retention-policy boundary below; queued following G070's integrated runtime/readiness checkpoint. Specification only; no planner, implementation or provider action is released by this publication.
Owner: Codex Goal/Spec/review; existing Claude coordinator Plan/implementation/validation/Result.
Priority: high — supervisor default for the owner's requested operational follow-up; independent implementation lanes follow the approved Plan.

## Objective

Operate the reusable SaaStacular staging environment from GitLab through two
explicit manual actions: start/deploy an exact selected staging candidate with a
bounded session duration, and stop the environment on demand. Reuse G070 rather
than create another deployment system or require the owner's workstation to stay
online. Ordinary CI qualification remains independent.

## In Scope

The [Work Spec](../specs/G072-manual-staging-ci-spec.md): deployment-only pipeline
routing; trusted candidate selection; GitLab OIDC/GCP workload federation; small
adapters for G070's ephemeral-runner configuration and durable deployment metadata;
manual start/deploy and independent stop; shared-environment serialization;
two-hour default/custom duration/native auto-stop; focused validation and operator
documentation. One existing staging environment and its bounded deployment identity.

## Out of Scope

Automatic deployments on merge/push/tag, production deployment, arbitrary
unmerged branch execution with credentials, a new cloud scheduler, CI reset or
retirement buttons, general image-publishing infrastructure, Kubernetes, new
application behavior, and repeated application/full-suite qualification per action.
G070's initial infrastructure/secret bootstrap and G057's real authentication and
custom-domain qualification keep their own ownership and live-action gates.

## Dependencies

Planning inputs: G070 Goal/Spec r2, accepted Plan including runtime addendum r1.9
and its Codex implementation clarifications, current MR !113/source, ADR-009,
existing GitLab CI/branch/security policies. Reconcile the actual integrated G070
interface before implementation; preserve its running checkouts and evidence.

The follow-up uses G070's merged lifecycle tooling and a specifically released
ready environment for live verification. Do not wait circularly for G070's final
retirement, which can follow G057 Part B; readiness and final ACCEPT are distinct.
G057 may use its ready environment independently of G072. This follow-up does not
add a new G021 prerequisite or release the foundation barrier.

## Completion Definition

The Spec ACs have attributable source, routing/authentication and bounded hosted
start/deploy/stop evidence; the exact candidate/digests and native stop deadline
are recorded, stop is usable independently, and resources/retention are reconciled.
Codex accepts the Result after applicable integration and selected qualification.

## Execution Policy

Security/infrastructure tier: Claude writes a compact Plan; Codex Plan ACCEPT and
release precede implementation. The owner's instruction, "make a follow up spec
for the ci jobs", authorizes this Goal/Spec and scheduling only. Preserve that
phase boundary: no automatic planner or worker launch from this publication.
The later Plan identifies exact files, CI UI, credential trust conditions, G070
adapters, selectors, resource reservations, and separate code/setup/build/live
time estimates. Account/IAM/GitLab-setting writes and live jobs require their
concrete reviewed action release; reuse existing authority only within its scope.

## Iteration Budget

B3: three shared execution attempts, checkpoint after attempt 2, under the
budget policy (not included). Zero spent at specification. Planning and
readiness waits spend no execution attempt; worker changes do not reset budgets.

## Context / Scope Budget

Two manual actions for one environment, existing G070 lifecycle/manifest formats,
one federation integration and bounded docs/tests. No general deployment platform.
If this requires broad lifecycle/state redesign, stop with the smallest adapter
or split proposal. The earlier 3–6-hour conversational estimate is provisional;
the Plan must size actual remaining work and live/setup dependencies.

## Suggested Subagents / Delegation

Plan targets 3–5 implementation workers when useful: identity/bootstrap,
pipeline/lifecycle adapter, and routing/proof/docs lanes with exclusive paths.
Record smaller allocations where shared files or resource capacity force
sequencing. One coordinator integrates; no simultaneous writers on G070's active
files, no duplicate monitoring owner, and no parallel image builds. Existing
private-run/resource limits remain binding.

## Stop / Escalation Conditions

Untrusted refs obtaining deployment credentials; long-lived Google account keys;
secrets/state in logs or public artifacts; widened production/account access;
reset/destroy during routine start/stop; unsafe retained-data/schema downgrade;
unowned provider cost; inability to stop after a failed/canceled deployment;
conflicting lifecycle owners; G070 contract/security changes; native refusal or
budget exhaustion. Preserve failures and unaffected safe work.

## History

- 2026-10-27 — Owner requests a follow-up Spec after discussing two manual CI
  start/deploy and stop actions, SHA/branch selection and short-lived federation.
  Latest owner duration preference is 2 hours by default, configurable per session,
  with on-demand manual stop. No implementation or cloud release is inferred.

## Owner amendment — reusable preview slots, 2026-10-27 UTC

The owner approves adding concurrent preview environments to G072, using stable
numbered slot names. This supersedes the single-environment limits in the original
In Scope and Context sections. The initial registered targets are `staging` and
`preview-1`; the adopting hostnames are `staging.env.foundation.test` and
`preview-1.env.foundation.test`, each with its own tenant wildcard. Reusable examples stay
generic. Later slots use `preview-2`, `preview-3`, etc., added by configuration and
their separate resource release; no extra slot is provisioned by this Spec.

The same two manual actions accept an allowlisted environment selector, candidate
and duration. Staging keeps integrated-candidate restrictions. Preview may deploy
an explicitly selected same-repository implementation/MR SHA, pinned before use,
through trusted protected-staging control code. Preview build/runtime boundaries
must prevent that candidate from obtaining deployment authority or staging/
production credentials. Each environment has independent VM/data, state/config,
identities/secrets, DNS/certificates, metadata, ownership lock and deadline; both
may run concurrently under existing shared build/resource limits.

The [Spec r1 amendment](../specs/G072-manual-staging-ci-spec.md#8-r1--numbered-preview-environments-2026-10-27-utc)
defines isolation and proof. Bootstrap of the second environment remains an
explicit one-time G070-style operation, outside routine CI jobs. The earlier
3–6-hour estimate predates this amendment and must be reassessed in the Plan.
The current authority remains specification only; G070 is not reopened and no
planning, execution, provisioning or new G021 prerequisite is released.

## Owner amendment — retention policy boundary, 2026-10-27 UTC

Use the retention defaults in [G070 Spec r3](../specs/G070-on-demand-staging-spec.md#r3--stopped-resource-retention-and-retirement-follow-up-2026-10-27-utc):
14 stopped days for shared staging, protected while G057 Part B depends on it,
and 48 stopped hours for numbered previews, with explicit owner extension
available. G072 remains limited to manual start/deploy and stop. Automatic
retirement requires a separate G070 lifecycle follow-up and is not authorized by
this Goal or its Plan release.

## Plan r1 acceptance and source implementation release — 2026-10-28

Codex accepted Plan r1 (`4ba65b06`, Plan SHA-256
`ef530b36b81d8575d067123369e7dd33889c23292a14211d5d96e4bb6dbcf509`) and
released its four-worker source implementation and focused offline proofs under
the owner's earlier direction to continue G072 planning/implementation after
G057 Plan §18.R1 publication. The release does not include GitLab setting
changes, WIF/IAM, GCP/DNS changes, credentials, or live staging/preview actions;
those retain §10.14's separate gates. G057 continues to own staging identity-pin
changes until its hold is released.
