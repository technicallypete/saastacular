> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G019 — Scoped invitations, membership and ownership administration

Epic: EPIC-007 — Group and Business Profile onboarding

Status: Definition finalized after owner check-in, 2026-10-17; Claude planning and execution held

Owner: Codex Goal/Spec/review; Claude Plan/implementation/validation; product owner escalated policy

Date: 2026-10-16 UTC; reconciled after G018 Plan acceptance, 2026-10-17 UTC

History: original provisional definition (not included).
Owner “Do that” resumes specification finalization only after the recorded G018
Plan check-in. Earlier stops are preserved in r0; planning/implementation remain held.

## Objective

Provide scoped invitations and membership administration while guaranteeing one
or more owners for every existing Organization and Group. Creator ownership,
additional owners and atomic succession use the existing RoleBinding authority.

## In Scope

- Scoped invitations for known and first-time recipients: expiry/revocation, intended
  verified recipient/scope/role, single-use acceptance and authorized membership
  operations. G058 completes first-time sign-in/proof and public admission.
- Finite ordinary-versus-owner authority matrix; bounded membership/invitation
  administration UI and server-side current eligibility, without a global user directory.
- Organization and Group owner promotion, atomic transfer and last-owner
  protection, enforced by guarded core operations and database invariants.
- Minimal trusted Organization-creation/bootstrap reconciliation so a new tenant
  and its verified initial owner's binding commit together; G018 already owns
  authenticated Group creation and its initial creator-owner binding.
- Normal account retirement/departure checks across owned scopes, truthful UI
  eligibility, audit and bounded concurrent-operation/rollback evidence.
- Existing-data invariant preflight and restricted emergency recovery procedure.
- Organization departure atomically removes its dependent Group grants, after
  required ownership succession, so later readmission cannot revive dormant access.
- Stable invitation/admission consumption primitives for G058; G058 owns public
  nonmember authentication/enrollment and ordinary tenant-session issuance.

## Out of Scope

Public Organization signup, new membership/owner role store, BusinessProfile owner
roles, ownership inferred from email/domain/branding, commercial/client enrollment
or consultant application policies, new email vendor or generalized workflow
engine, production-data repair, and live emergency account recovery. No automatic
admin promotion, suspension bypass or early product work before G021.
No public global account-delete/disable API, Group-role-based Organization admission,
new applicant session/issuer, implicit Organization-owner override of Group ownership,
automatic restoration of revoked bindings, or forced production ownership repair.

## Dependencies

- Final ACCEPT/integration of G018 and accepted identity, policy/RLS, audit and
  recovery foundations. Reconcile G018's actual creation and permission interfaces.
- [Work Spec](../specs/G019-scoped-invitations-membership-ownership-spec.md),
  Architecture §53.2/§53.7, SECURITY, AUTHORIZATION and accepted tenancy ADRs.
- G018 Plan r5 design ACCEPT (not included),
  sourcef60e052c / SHA4fe4642f, settles the creation/profile/lock interfaces for
  this definition. G018 implementation is not accepted yet; reconcile actual
  final integrated source and relevant deltas before G019 execution.
- G058 follows G019 and consumes its primitives before G021. G058 is not a
  prerequisite for G019 implementation; public enrollment adapters are separate.
- No migration number, route naming or new infrastructure is reserved here.

## Completion Definition

Spec AC1–AC8 have reviewed source and attributable focused evidence, required
MR/staging disposition, scoped cleanup and Codex final ACCEPT. Existing tenants/
Groups cannot become ownerless through a supported ordinary mutation path.
Emergency security suspension remains immediate and does not grant a replacement
owner or erase the ownership record. G021 verifies the integrated invariant.

## Execution Policy

**Architecture/security/data tier.** Owner's 2026-10-17 “Do that” authorizes
reassessment/finalization of the existing definition after the G018 Plan check-in;
it supersedes the earlier stop before Spec reassessment for this phase only.
**Stop before G019 Claude planning or implementation.** Later explicit phase authority, Plan ACCEPT,
dependency/input/budget/resource readiness and recorded release are required.
G017/G018 work continues under its own mandate; no active code is changed. Any
recipient-identity policy requiring new trusted evidence must be resolved explicitly.

## Iteration Budget

Provisional **B7, checkpoint after attempt4**, under BUDGET_POLICY; reassess toward
a smaller class at Plan review. Concurrent membership/ownership, database guards
and existing-data transitions require a bounded reviewed design. Definition spends
no attempt; no budget is reset for earlier Goals or delegated workers.

## Context / Scope Budget

Two Organizations, a small set of Groups and current owner/admin/member roles.
Reuse existing membership, audit, session/revocation and database fixtures. Focus
on meaningful recipient/role denials, last-owner races and rollback; no role×scope
permutation campaign, new oracle framework or broad suite by default.

## Suggested Subagents / Delegation

One mutation/invariant owner; optional bounded read-only security/data review.
Use the verified coordinator and native isolated task mapping; no duplicate
session or background ownership-repair agent.

## Stop / Escalation Conditions

Material drift from accepted G018 interfaces; unreviewed owner-management authority; ownerless legacy
records without an authoritative resolution; unsupported recovery/retirement
semantics; security suspension delayed by transfer; foreign recipient/scope,
escalating role or replay acceptance; unsafe migration, new vendor/public surface,
native refusal, unexplained locking/query cost or budget exhaustion. Preserve
history and return the smallest concrete gap rather than invent ownership.


## Owner finalization and remaining stop — 2026-10-17

G018 Plan r5 design is accepted, exact reviewed artifacts imported at0568d610.
The owner answered “Do that” to finalizing this existing Spec, then explicitly
selected support for first-time invitees with sign-in/admission completed in G058.
The former stop before Spec reassessment is lifted only for this definition.
Earlier stop/check-in records remain in linked r0 history. G058 owns the public
nonmember authentication adapter; G019/G058 planning and implementation remain
held. Reconcile actual accepted G018 before later release.

## Owner lifts planning hold at G018 MR trigger — 2026-10-17

After Codex recommended lifting the G019 planning hold so its normal trigger can
start planning, the owner replied “Great ok.” G019 provisional Claude planning
is authorized at G018's first actual implementation MR against staging, including
a draft. Verify the MR/candidate and reuse one planner; do not create a placeholder
MR just to trigger planning. Reconcile against G018's actual implementation and
record pending dependencies. This supersedes the earlier planning stop only;
G019 implementation remains held for later phase authority and normal Plan/
binding/dependency/release gates. G058 planning/execution remains held. No
automatic specification or wider delivery mandate is added.

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


## Owner amendment — G019 manual review waived, 2026-10-18 UTC

At06:45:20 UTC the owner directly instructed the verified Claude coordinator:
“Waive my manual review for 019”. Codex independently verified the direct user
message, rather than treating an observer notice or coordinator summary as
approval. This waives G019's owner manual changed-flow review/M1 only, including
that prerequisite in the accepted Plan's §6.4/slice5 and AC6/AC8 evidence mapping.
Record M1 as **owner-waived, not performed**, never as passed or positive feedback.
Immutable Plan/history remains unchanged; this explicit amendment supersedes
only its manual-review requirement for G019.

Automated unit/web/browser/accessibility/security/isolation/database qualification,
Codex source review, native permissions, DB/Q1, applicable CI/Duo/integration,
cleanup and final Goal ACCEPT remain required. No native test execution authority,
security exception, scope/budget reset, G058 release or G020 planning is granted.
Unchanged automated evidence is reused under existing attribution rules; this
paperwork alone requires no test rerun. Claude appends the waiver to its Result
at its safe handback; active source ownership and frozen batches are preserved.
