> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G058 — Generic Organization enrollment and self-signup work specification

Goal: [G058](../goals/G058-generic-organization-enrollment.md)

State: Definition authorized, 2026-10-17; no planning/execution or external-action release

## Purpose and authority

The owner confirmed that generic public enrollment belongs in the reusable
foundation, then instructed “Yes make these updates” to a bounded Goal after G019
and before G021. G042/G043 retain commercial/client acquisition composition;
they no longer own the initial generic admission capability. G018 continues its
current authenticated Group/Business Profile work; do not fold enrollment into
its active Plan or restart it.

Canonical identity is global, Organization is the tenant, and RoleBindings remain
the sole stored membership/role authority. In Beginify, an agency is an Organization
and a client account/team can be a Group; these labels are illustrative only.
SaaStacular must retain the same generic behavior from accepted G021.

Read governance, Architecture §53.4–53.11, SECURITY, DECISIONS, ADR-003/004/005/007
and accepted G018/G019 Plans/Results before later planning. ADR-007's ordinary
sign-in membership refusal stays in force; its dated G058 amendment permits only
the separate, policy-authorized enrollment path defined here. Detailed interfaces
are reconciled after G019 implementation. G019 is now definition-finalized under
separate owner authority; neither Goal has a planning/implementation release.

## R1 — Finite tenant policy and authority

Support exactly these generic modes for an existing, server-resolved Organization:

| Mode | Admission outcome |
|---|---|
| Invite-only | Reuse a valid G019 invitation and its intended subject, scope, role, expiry/revocation and one-use semantics. No invitation means no automatic membership. |
| Request-access | An authenticated applicant can create/reuse one bounded pending request and inspect their own state. Current authorized approval is required before membership. |
| Open self-signup | An eligible authenticated applicant is admitted as an ordinary Organization member without an extra administrator approval step. |

Missing/invalid mode is closed/invite-only, never open. Enrollment-mode changes
are privileged Organization settings under existing owner-only authority. Request
decisions use the final accepted G019 Organization membership-grant authority;
pin exact roles/capabilities in the Plan, deny absent authority, and escalate a
conflict rather than inventing powers. Group roles alone never manage Organization
admission. Applicants cannot supply the grant role, approval or destination tenant.

Recheck current mode, applicant eligibility, approver authority where applicable
and host→Organization binding at admission. Old pages, in-flight callbacks and
queued approvals cannot override policy closure, revoked approval or host reassignment.
No automatic admission to another Organization or existing Group.

### First-time invitation recipients — owner decision, 2026-10-17

The owner explicitly selected first-time invitee support, with sign-in/admission
completed here. Reuse finalized G019 invitation intent and its verified-subject
consumption contract; do not require a pre-existing platform account to issue an
invitation. G019 may record an intended delivery address, not membership proof.
Current VerifiedIdentity contains no trusted email. G058's reviewed core-owned
identity adapter must obtain/validate fresh trusted provider evidence of the
invited recipient, bind it to the canonical namespace/subject and local UUID,
then invoke G019 atomically. Never merge identities or admit users merely because
typed/unverified emails match, or let clients construct a trusted proof object.
Pin provider support/evidence and normalization before implementation; fail closed
and escalate an unsupported proof instead of claiming first-time redemption works.
Retain current auth/session and optional-assurance rules; no new vendor or issuer.

## R2 — Authentication is distinct from admission

Reuse WorkOS through the existing core-owned adapter, global canonical UUID mapping,
PKCE, browser-bound one-use state and exact-origin/Organization validation. Identity
claims, hostname and email/domain similarity alone are not membership authority.
Preserve account-disable, MFA/assurance and provider trust rules; no new issuer.

Define an explicit enrollment intent/boundary for a verified nonmember. It may
create only the global identity mapping and applicant state necessary for admission.
Any credential/reference for pending enrollment is short-lived and limited to
the applicant's enrollment operations/status; ordinary HTTP/TenantContext/MCP
operations must reject it. Pending/denied applicants receive no ordinary tenant
session or tenant-data access. Do not globally relax G016's sign-in membership check.

After committed admission, issue/rotate an ordinary exact-origin and Organization-
bound session only through fresh G016 issuance checks. A direct or staged transition
is acceptable if its atomicity and failure outcomes are pinned; provider/network
calls cannot hold database locks. Existing valid members keep their normal login
flow; policy-gated open signup may complete admission during the authenticated
signup return without an extra manual enrollment step. Do not equate provider
account creation, local identity persistence and committed tenant admission.

## R3 — One membership transaction and truthful retries

Reuse G019's guarded invitation/membership service and constraints. Consume the
relevant invitation/approval and create or recognize exactly one authoritative
ordinary membership with required audit effects atomically. No duplicate role
store, second invitation lifecycle or member-creation endpoint that bypasses G019.
An external provider account can remain after local admission fails; do not
report tenant membership or destructively compensate an uncertain external outcome.

Concurrent open signup, invitation redemption, approval, policy change, membership
revocation and duplicate submissions must preserve current authority and one
binding. Pin lock/commit order, retry/conflict/unknown outcomes and a guarded status
read. A lost response must not create another request/grant or falsely claim success.
Session failure after membership commits is reported/recoverable without inventing
rollback of durable membership. Recovery/restriction replay must preserve admission
denials/revocations alongside G019/G013; scope the actual necessary changes.

Open signup must not silently recreate revoked/blocked membership, override a
denied/revoked admission decision, re-enable a disabled account or bypass an
applicable suspension. Distinguish first enrollment from reinstatement; use
accepted G019 authority for explicit reinstatement or escalate unresolved policy.
Changing signup mode is not automatic reinstatement authority.

## R4 — Minimal grant and foundation handoff

Self-enrollment grants ordinary Organization membership only, never Organization
admin/owner, invitation/policy management, existing Group membership or access to
another client's Group/profile. Invitation acceptance retains G019's separately
authorized finite grants; it is not widened by open signup.

An admitted member can use G018 to create their own Group, become its initial
owner atomically, and create a new Business Profile or select an already eligible
one. Completion remains exactly one same-Organization attached profile, Group-local
and resumable. Reuse that flow and its constraints/UI; do not create a second
Group/profile engine. New Group ownership never authorizes an unrelated profile.

## R5 — Thin public/applicant/admin surfaces

Expose only intentional public policy/entry information for a currently valid
tenant host, own applicant status and currently authorized tenant request decisions.
Use generic configured labels and the existing tenant shell. Bound inputs, bodies,
request/list traversal, lifetimes and request rates with existing infrastructure;
no separate abuse-control platform. Keep redirects server-derived and same-origin,
CSRF/browser intent protected, private results no-store and role decisions server-side.

No account/membership/request enumeration, private reasons about another applicant,
tenant data in pending pages, or invitation/callback/session material in URLs/logs
beyond the existing reviewed callback protocol. Audit policy/admission/decision
outcomes with minimal trusted attribution; no duplicate success events or raw
identity/profile payloads. Reuse existing audit ownership/rollback rules.

Required manual feedback covers the changed public/enrollment/status transition
and return to G018, with keyboard/error association, narrow reflow and existing
Light/Dark/System behavior. No new brand asset, vendor or broad UI matrix.

## Acceptance criteria

| ID | Required outcome | Focused evidence |
|---|---|---|
| AC1 | Closed default, three finite modes and guarded controls | Current role/mode mapping, missing/invalid policy denial, unauthorized policy/approval negatives; no Group-role administration. |
| AC2 | Verified nonmember enrollment has no normal tenant authority | Callback/browser/origin/Organization binding positives and replay/foreign/expired/disabled negatives; pending credential refused by ordinary HTTP and MCP; normal nonmember login still refused. |
| AC3 | Invitation, request and open admission reuse one membership contract | Valid known-user and first-time intended-recipient invites via G019; first-time WorkOS/recipient proof cannot be forged by a client or replaced by unverified email; request remains pending until allowed approval; eligible open signup creates one ordinary binding; existing member retry remains idempotent. |
| AC4 | Current authority and atomicity survive races/failures | One deterministic case each for relevant mode/approval closure, duplicate admission/overlap and host or membership revocation; injected transaction fault leaves no partial mandatory grant/consumption/audit; truthful lost-response recovery. |
| AC5 | Enrollment cannot bypass prior restrictions | Removed/blocked/suspended/disabled/rejected/revoked cases deny automatic reinstatement; authorized explicit recovery is distinct; recovery/replay retains restrictions. |
| AC6 | Admitted user can complete generic own-Group onboarding | Actual configured-host synthetic flow authenticate→policy admission→ordinary member→G018 Group/new-or-eligible profile→complete, with no admin/other-Group/unrelated-profile access; reuse unchanged G018 cases. |
| AC7 | Public/applicant/admin UI and evidence preserve privacy/accessibility | Own-state and authorized request list/decision tests, bounded inputs/traversal and safe errors/logs; focused browser/manual feedback, no provider secret/production data. |
| AC8 | Reviewed foundation delivery and export linkage | Result maps source/inputs/commands/outcomes and manual findings, required MR/staging disposition and cleanup; G021 requires G058 final ACCEPT and OS001 retains accepted generic behavior. No deployment claim. |

## Validation, Plan and stop boundaries

Five scenario groups: mode/authority; auth-versus-admission; atomicity/races/restrictions;
G018 handoff; privacy/changed UX/closeout. Reuse G019 invitation and membership
fixtures, G016 callback/session negatives, current RLS/audit checks and two synthetic
tenants. Extend only distinct missing cases. Focused executable selectors from
the first local batch through final coverage; no default full local gate, provider
campaign or role×mode×host×theme cross product. Required hosted profile remains
honest; justify any broader batch technically before dispatch. Preserve failures,
compact Results and CI artifacts; raw local evidence is temporary.

Later Claude Plan pins the finite permission matrix, minimal applicant/admission
state, nonmember mapping/callback/session transition, transactions/RLS/audit and
restriction replay, exact routes/files/count pins, bounded lifetimes/rates and
test selectors. A Plan/definition does not grant external WorkOS/DNS/TLS changes.
G057's required live qualification must reconcile any relevant accepted auth delta;
do not repeat unaffected live proof or claim emulator evidence proves provider behavior.

G018 Plan-acceptance owner check-in is complete; G019 Spec finalization was
separately authorized. G019/G058 planning/implementation stops remain. G058
definition does not extend the automatic delivery range. Final dependencies, later phase
authority, Plan ACCEPT/current binding, resource readiness and explicit supervisor
release precede implementation. Provisional B7/checkpoint4; no attempt reset.

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
