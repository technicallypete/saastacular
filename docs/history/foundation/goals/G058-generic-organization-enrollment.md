> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G058 — Generic Organization enrollment and self-signup

Epic: EPIC-007 — Group and Business Profile onboarding

Status: Definition authorized, 2026-10-17; planning/execution held under existing
owner stop boundaries and normal dependency/review/release gates

Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation

## Objective

Complete reusable tenant self-service onboarding before G021: a user authenticates,
joins an existing Organization only under its current enrollment policy, then
uses G018 to create their own Group and create/select an eligible Business Profile.
Admission gives ordinary membership, never staff/admin or other-Group authority.

## In Scope

- Organization enrollment modes: invite-only, request-access and open self-signup;
  safe closed default, guarded policy changes and current authorized request decisions.
- Verified identity/browser/origin/Organization-bound enrollment for nonmembers,
  distinct from ordinary tenant sessions; own pending/denied/approved status.
- Reuse G019 invitations, membership grants/revocation and authoritative RoleBindings;
  minimal request/decision/idempotency state where needed, not another invitation system.
- Owner decision, 2026-10-17: complete first-time invitees' sign-in, trusted recipient
  proof and admission here through G019's consumer contract; no prior account needed
  for invitation issuance, no email-only identity merging or membership grant.
- Atomic admission/membership/consumption/audit, current-policy race checks and
  honest retry/unknown-outcome behavior; normal tenant-session issuance only after admission.
- Thin generic public entry/status, bounded authorized request administration and
  return to accepted G018 Group/Business Profile onboarding; accessibility/privacy.
- Focused isolation, callback/admission, race/replay and changed-UI evidence;
  generic operating/adopter documentation and G021/SaaStacular acceptance linkage.

## Out of Scope

Public Organization creation, automatic owner/admin assignment, existing Group
membership grants through signup, another identity/issuer/vendor/email service,
commercial offerings/agreements/payment/trial activation, consultant recruitment,
marketing/page-builder work, new onboarding/workflow/notification framework,
cloud provisioning, live provider configuration or public deployment.

## Dependencies

- Final accepted G018 and G019, including current guarded creation, invitations,
  membership/ownership and recovery contracts; reconcile their accepted interfaces.
- Accepted G004/G007/G008/G011–G017 identity, context/RLS, audit/recovery,
  tenant-host binding and generic shell. G057 owns required live-host qualification.
- [Work Spec](../specs/G058-generic-organization-enrollment-spec.md), Architecture
  §53.11 and the dated ADR-007 enrollment amendment. G021 requires G058 ACCEPT.

## Completion Definition

Spec AC1–AC8 are met against reviewed source and attributable focused evidence;
manual changed-flow feedback, applicable MR/staging disposition and scoped cleanup
are complete. G021 can exercise a synthetic allowed self-signup→ordinary member→
own Group→Business Profile flow and the required denial paths. Codex final ACCEPT
is distinct from merge or a passing emulator. No production/public deployment claim.

## Execution Policy

**Architecture/security/data tier.** Owner “Yes make these updates” authorizes
this bounded definition and related architecture/roadmap/export corrections.
It does not authorize planning, implementation, deployment or changes to active
G017/G018 bindings. G018 Plan-acceptance owner check-in is now complete; the
owner separately authorized G019 Spec finalization. G019/G058 planning and
implementation remain held. Insertion of G058 does not enlarge
the pinned automatic-delivery mandate. Obtain later phase authority before dispatch.

After release, Claude proposes a concise Plan: finite modes/authority matrix,
ordinary-sign-in versus enrollment trust boundary, minimal persistence/transactions,
token/session lifecycle, migrations/RLS/audit, bounded routes/UI and focused checks.
Explicit Codex Plan acceptance, dependency ACCEPTs, exact bindings, resources and
recorded execution release precede writes. Material architecture decisions escalate.

## Iteration Budget

**Provisional B7; checkpoint after attempt4.** Reassess at Plan review against
the accepted G019 primitives and actual auth delta; reuse can justify a lower
budget. Planning consumes no execution attempt; workers share one Goal ledger.

## Context / Scope Budget

One existing tenant, three finite admission modes and one ordinary membership
outcome; two synthetic tenants suffice for isolation. Reuse auth/invite/role and
Group onboarding implementations. No role×mode×host×theme Cartesian campaign,
new evidence system, benchmark campaign or generic enrollment engine.

## Suggested Subagents / Delegation

One implementation owner. A bounded read-only auth/admission security review
before UI composition, plus changed-UX review if needed. Use current model/effort,
context-handoff and native-worktree defaults; no extra persistent coordinator.

## Stop / Escalation Conditions

Owner phase stop; missing accepted dependencies; ordinary-sign-in membership
bypass; enrollment credentials accepted as normal sessions; stale/revoked policy
grant; automatic reinstatement or staff elevation; foreign/sibling data exposure;
duplicate membership/invitation state; new vendor/trust channel/infrastructure;
unreviewed destructive migration; unsafe external action; native refusal; budget
or context exhaustion. G021 and OS001 remain blocked until required acceptance.

## Owner autopilot endpoint — stop before G020 planning (2026-10-17)

Owner “Go autopilot”, then “Stop before 20 planning”, changes the earlier phase
endpoint. Roadmap order at `584398064c06dbb8becab35b83bb41922f152b7a` places G058 between G019 and G020.
Recorded mandate: continue existing G017/G018 delivery and the normal planning,
review, implementation, integration, acceptance and closeout loop for the already
defined G019 and G058; **stop before G020 planning**, including speculative or
advance planning at a predecessor MR. This supersedes the earlier G019
implementation/G058 planning-execution stops for those two tasks only.

Use the normal first-implementation-MR planning trigger, reviewed predecessor
integration plus current Plan/input/scope/readiness and supervisor release for
isolated implementation overlap. Final predecessor ACCEPT still gates successor
integration/final ACCEPT. Preserve per-Goal budgets, focused checks, predecessor-owned
parallel repairs and all owner security/architecture/vendor/live-action decisions.
No automatic G020 specification or other missing Goal/Spec authoring is authorized,
no G020 execution, no later/independent task dispatch, and no later roadmap insertion
expands this pinned set. Existing public-provider/deployment/native permission
holds remain. This mandate is not Plan acceptance or immediate execution release.


## G019 consumer execution boundary — design dependency, 2026-10-17

G019 r2 review (not included) accepts owner-only
Organization-consumption execution for synthetic contract evidence; it does not
qualify the production adapter. G058's Plan must resolve and review the actual
provider-proof and execution-credential boundary before release. No runtime claim
EXECUTE grant, application owner credential, new identity login/pool or weakened
verified-recipient requirement is preapproved. A new credential/pool requires the
owner's explicit architecture decision. Preserve the G019-MR planning trigger;
this note creates no early planner or additional authorization.

## Owner-selected G058 design package — 2026-10-18 UTC

The owner answered “I'll take your recommendation” to the three-item package in
review r3 (not included), then confirmed the intended
client self-signup flow. This selects D-4r1 durable restrictions, D-15 option(i)
existing trusted core/runtime execution, and D-16 option(c) fail-closed readiness
for open signup where legacy departure history is unresolved. It supersedes the
prior unselected hold for these choices, not technical validation/release gates.

- Persist removal/denial/revocation restrictions independently of audit retention;
  only fresh explicit authorized reinstatement lifts them. Preserve reviewed
  restore/replay materialization and atomic audit/membership rules.
- The core-owned provider adapter supplies fresh verified address evidence bound
  to the canonical principal. Reviewed runtime EXECUTE may invoke the bounded
  Organization invitation consumer; SQL checks principal/recipient/token/current
  authority, not independent provider provenance. No new credential/pool/issuer.
  SDK projection qualification remains mandatory before affected execution.
- Add one minimal explicit Organization readiness fact, trusted setter/procedure
  and authoritative open-policy guard. Unresolved legacy history refuses open
  mode until operator-reviewed backfill or verified absence of prior departures.
  Ordinary tenant/applicant writes cannot forge this fact. Verified fresh installs
  and newly created Organizations without such history need no fabricated
  backfill. Pre-ledger restore must re-close unresolved readiness after migration;
  neither a missing ledger nor a mode change means history was cleared.
  Fresh admin-approved request-access and valid invitations remain available;
  this choice does not add a legacy hold on request_access.

**Required end-to-end outcome:** authenticate under the configured tenant host →
policy-authorized ordinary Organization membership → fresh ordinary sign-in →
G018 self-service creation of a new Group with the creator as initial owner →
create/select an eligible Business Profile and complete onboarding. In the
owner's product illustration, a client is that Group. Signup does not create a
public Organization, assign Organization staff privileges, or join another
client's existing Group. Tenant policy must permit open mode for admission to
be automatic; invite/request modes keep their explicit prerequisites.

Claude now appends a bounded selected-design/readiness Plan addendum with exact
storage, trusted setter, migration/fresh-install/restore behavior, RLS/direct-SQL
guards, route/status behavior, pins/scope and focused readiness negatives.
Preserve r3 and all history; do not restart the Plan or G019 work. Codex reviews
the selected delta before full Plan ACCEPT/current-target binding and release.
No application execution from this owner choice alone. Existing dependency,
SDK/native/live-action/budget gates and stop before G020 planning remain.


## Owner sequencing override — merge before M1, 2026-10-19 UTC

Owner: “Can we merge and then do m1 review?” Approve conditional staging integration
before manual M1, once current source review, dependency, CI and Duo gates clear.
M1 moves to a **post-merge final-acceptance gate**, not waived/passed. Supersedes
the prior M1-before-merge hold only. Manual review remains scheduled by owner for
the morning; final G058 Goal ACCEPT requires its real outcome, required hosted
qualification and cleanup. G055/MR48 remain held until that final ACCEPT; G020
phase stop/G021 barrier unchanged. No public deployment or native bypass.
