> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G019 — Scoped invitations, membership and ownership work specification

Goal: [G019](../goals/G019-scoped-invitations-membership-ownership.md)

State: Definition finalized after owner check-in, 2026-10-17; Claude planning/execution held

History: original provisional Spec (not included).
The original owner decisions/stops remain historical; the dated authority below governs.

## Authority and accepted interfaces

Owner “Do that” resumes Spec reassessment/finalization after [G018 Plan r5 design
ACCEPT](../results/G018-plan-codex-review-r5.md), sourcef60e052c, SHA4fe4642f.
The owner also selected **“Yes—support first-time invitees; complete their sign-in
and admission in G058.”** This authorizes the definition and G058 boundary update,
not G019/G058 planning, execution, live provider changes or a wider delivery mandate.

G018’s accepted design permits any current Organization owner/admin/member to
create their own Group in Organization-only context, with atomic creator-owner
binding, member-safe keyed INSERT then read, no other-Group/Organization elevation.
It also pins derived Business Profile eligibility, completion and ascending Group
lock order. Preserve those contracts, bootstrap trigger/grants and provenance
unless a specific necessary change is proposed and reviewed. G018 implementation
is not accepted yet: reconcile its final integrated source before G019 execution;
do not reserve a migration number or infer implemented guarantees from its Plan.

Read governance, architecture §53.2/§53.7/§53.11, SECURITY/DECISIONS, accepted ADRs
and existing membership/RLS/audit/recovery paths. Canonical UUIDs and RoleBindings
remain the only membership/role authority. No profile roles or mirrored owner store.

## R1 — Finite authority and scope

Administrative and existing-member operations require a current enabled subject and fresh trusted context, exact
same-Organization scope, current membership and the applicable existing capability.
Selected-Group context cannot target another Group. Server decisions, not UI or
client-supplied actor/Organization/role claims, are authoritative.

The sole recipient-membership exception is G058's bounded verified-nonmember
Organization-invitation consumer, which establishes that first membership. It
still requires the enabled verified recipient, exact tenant binding, current
issuer/grant authority and all R4 checks. It is not an ordinary nonmember
TenantContext/session or an administrative-capability exception.

| Operation | Organization scope | Group scope |
|---|---|---|
| List/manage ordinary admin/member bindings and invitations | Current Organization owner/admin | Current Organization owner/admin across the active Organization, or target Group owner/admin; selected-context restrictions remain |
| Promote an existing eligible admin to owner; change/remove/transfer an owner | Current Organization owner | Current owner of that Group; Organization-wide ordinary member management does not implicitly grant owner administration |
| Leave one’s own membership | Current member acting on self, subject to scope/dependent-ownership protection | Current member acting on self, subject to last-owner protection |
| First-time/nonmember admission through an invitation | G058 verifies the intended identity and invokes G019’s bounded consumption contract | Must already have current Organization membership; a Group invitation alone cannot admit someone to the Organization |

Ordinary administration handles admin/member roles only; any operation touching
an existing owner uses the owner-only boundary, even if its desired final role
is admin/member or absent. Invitations never mint owner roles. Group authority
never grants Organization authority. No new Organization-owner override of Group
ownership is inferred; restricted emergency recovery is separate. A same-scope
owner can promote/transfer only to an enabled, current same-scope admin; Group
recipients also require current Organization membership. A recipient lacking that
role must first receive a separately authorized ordinary role change.

## R2 — One-or-more owners, bootstrap and succession

Every committed existing Organization and Group has at least one assigned owner;
multiple owners are valid. The verified creator receives the initial owner binding
atomically with creation. One membership role per subject/scope, with owner including
admin powers plus owner-only powers; no extra admin binding or unique-owner index.
G018 owns Group creation. G019 reconciles trusted Organization bootstrap so tenant
and initial creator ownership commit together, with no ownerless intermediate commit.

Enforce owner continuity in core and PostgreSQL through every supported mutation:
creation, promotion, transfer, demotion, binding removal, ordinary retirement and
relevant FK/cascade effects. Count authoritative bindings, never an actor-filtered
RLS projection. Serialize competing owner changes on the same scope and pin lock
order with existing user/session/Organization/Group guards; no count-then-write race.
A direct runtime statement must not bypass the invariant or owner-only authority.
Do not broaden runtime DDL, role control, trigger creation or tenant visibility.

A last owner must first promote a replacement or use one atomic transfer that
promotes the recipient and demotes/removes the former owner together. Recheck
actor, recipient, membership and eligibility at the authoritative mutation; audit
failure rolls back ordinary succession. A UI count or earlier eligibility response
cannot authorize a stale transfer. Normal membership admin cannot relabel a victim
owner as a member to bypass this boundary.

Preflight existing data and trusted fixtures before enforcing new constraints.
Ownerless records require an explicit attributable resolution, not arbitrary
promotion/deletion. Stop on unresolved production/legacy ownership. Use a forward
reviewed migration, reversible where safe; never edit accepted historical migrations
or misrepresent the current CreateOrganization/role-change gaps as already fixed.

## R3 — Departure, retirement, suspension and recovery

Organization departure/removal checks every Group grant of the departing subject
in that Organization. Arrange required owner succession first, then atomically
remove the Organization binding and dependent Group bindings, invalidating related
pending grants as appropriate. Fail before partial removal if any continuing Group
would lose its last owner. A Group owner lacking Organization authority can leave
one Group but cannot remove another user’s Organization membership. A tenant admin
cannot silently strip a Group owner through an Organization-removal shortcut:
explicit authorized Group succession/removal must occur first.

No dormant Group grants may revive when the user later rejoins. A later permitted
readmission is a fresh Organization membership; previous Group ownership/access
requires a new explicit authorized grant. Preserve authoritative revocation and
restriction replay so pending tokens or restores cannot silently reinstate access.
Do not turn the audit log into a second role store.

Ordinary account retirement must address all owned Organizations/Groups, not just
the selected tenant, without exposing other-tenant details to an administrator.
Provide the bounded eligibility/guarded trusted-retirement reconciliation; no new
public global account-delete/disable endpoint or physical user deletion is required.
Tenant administrators cannot retire a global account or change its other tenants.

Emergency security disable/suspension/session revocation remains immediate even
for the last owner. Retain assigned ownership while denying usable authority;
last-owner protection cannot delay that security action. Assigned owner cardinality
and available operators are distinct. If all owners are unavailable, use a restricted,
verified, audited recovery procedure with explicit operator authority, not automatic
admin promotion or revived credentials. Document readiness and exercise implemented
pieces synthetically; no live recovery or production-data repair is authorized.

## R4 — One invitation lifecycle, including first-time recipients

Invitation identity/scope/role, issuer, intended recipient, expiry, revocation and
one-use consumption are immutable grant intent; changing intent requires explicit
revocation/reissue. Use cryptographically random opaque tokens, persist only their
safe digest, and bound lifetime/input/rates and outstanding/list traversal using
existing mechanisms. Raw tokens are shown only through the reviewed delivery/
redemption flow, never logs, analytics, audit payloads or cacheable responses.
Manual copy/delivery is sufficient; no email vendor/automated email service.

Support known canonical users and invitations addressed to first-time people.
An address may identify the intended recipient for delivery, but is never a global
identity key or membership proof. Current VerifiedIdentity intentionally exposes no
trusted email: a typed string, display name, unverified claim, domain match or token
possession cannot stand in for proof of the recipient. Existing accounts use canonical
UUID/provider namespace+subject mapping. G058’s reviewed identity adapter must obtain
and validate the trusted provider recipient proof required for first-time redemption,
then bind the intended recipient to the canonical identity without email-based account
merging. Pin exact verified-address/subject evidence and normalization at Plan review;
if the selected provider cannot establish it, stop for a reviewed resolution, never
silently accept an unverified email or defer first-time support as completed.

G019 owns issuing/revoking/status and the narrowly bounded verified-subject consumption
primitive. G058 owns public signup/enrollment, WorkOS proof acquisition/callback intent,
minimal first-time identity mapping and post-admission session issuance. No circular
dependency: G019 defines/tests its trusted-consumer contract with synthetic evidence;
G058 implements and qualifies the public adapter before G021. G019’s existing-member
browser acceptance covers Group invitations. Do not add an ordinary nonmember tenant
session, another applicant session/issuer or callback membership bypass to G019 just
to demonstrate Organization invitation redemption. An ordinary client cannot construct
or submit a trusted recipient-proof object or pick the consuming subject UUID.

At consumption recheck enabled intended subject, exact tenant/scope/role, current
issuer grant authority, expiry/revocation/restrictions and existing membership. Group
invitations require Organization membership; where a first-time person needs both,
G058 first completes separately authorized Organization admission, then Group acceptance.
One Group-manager token cannot double as an Organization invitation. A token intended
for one identity/tenant can never authorize another. Revoke/refuse grants invalidated
by issuer/recipient authority loss; define races against membership removal explicitly.

Membership effects, single-use consumption and mandatory audit commit atomically.
Concurrent duplicate acceptance yields one effect and a truthful consumed/already-member
outcome. Existing equal/higher authority is preserved, never silently downgraded;
any requested upgrade must be a currently authorized explicit invitation grant.
A lost response/unknown commit is recovered through a guarded status/idempotency read,
not another binding or fabricated success. Expired/revoked/used tokens cannot restore
removed membership. Session issuance belongs after committed admission, through G058.

## R5 — Thin bounded UI and failure behavior

Use the existing shell, core policy facade and thin HTTP/BFF conventions, including
exact-origin/session/CSRF, no-store and safe redirects. Provide only scoped member/
invitation lists, permitted issue/revoke/accept actions, ordinary role/removal controls,
owner promotion/transfer/leave and truthful last-owner/replacement feedback. Bound
pagination, candidates and response bodies; no cross-tenant/global email directory,
private invitation enumeration or unrelated profile/group disclosure. Never show a
global retirement detail list to a tenant administrator.

Reuse existing components and current appearance dropdown; no framework or new
notification/email system. Keep current keyboard/focus/error association, Light/Dark/
System and narrow reflow. Manual changed-flow feedback precedes final qualification.
Recheck every action server-side, including stale list/recipient choices. Preserve
existing strict atomic audit for ordinary mutations and the accepted security-revocation
failure exception; no duplicate application/trigger success or raw subject payload logs.

## Acceptance criteria

| ID | Required outcome | Focused evidence |
|---|---|---|
| AC1 | Atomic creator ownership and one-or-more-owner invariant | G018 creator path retained; Organization bootstrap atomic; two owners valid; one role per subject/scope; explicit legacy preflight; runtime/cascade/rollback cannot leave an ownerless existing scope. |
| AC2 | Finite ordinary/owner authority and atomic succession | Legitimate same-scope admin promotion/transfer; ordinary admin self-promotion/owner removal, inherited Group-owner override, disabled/foreign/stale recipient denied; audit fault preserves original ownership. |
| AC3 | Last-owner protection and departure races | Single last-owner refusal, competing two-owner departures leave one, transfer-vs-revocation ordering; Organization departure handles dependent Group bindings or refuses before partial writes; later readmission revives no Group grants. |
| AC4 | Retirement versus immediate security suspension | Ordinary eligibility spans all owned scopes privately; tenant admin cannot retire global account; last-owner disable/revocation immediate with ownership retained; recovery/restriction semantics explicit and implemented pieces tested synthetically. |
| AC5 | Scoped invitation lifecycle and first-time boundary | Intended canonical/verified subject consumption, scope/role/issuer checks, expiry/revoke/replay/foreign/escalation negatives and acceptance/removal races; atomic membership/consumption/audit and truthful lost-response status. First-time issuance/consumer contract covered synthetically; G058 public proof/admission remains explicitly pending. |
| AC6 | Authorized bounded member/invitation/owner UX | Existing-member Group invitation acceptance, role/owner controls, eligible-admin paging and last-owner/stale conflict feedback; privacy/CSRF/keyboard/appearance/reflow and manual changed-flow feedback, no global directory. |
| AC7 | Schema/guards/audit/restore remain coherent | Actual migration/grants/RLS/helper trust/lock order and affected FK/cascade paths reviewed; no mirrored roles; authoritative restrictions and owner cardinality survive supported restore/replay; focused query observations only. |
| AC8 | Reviewed delivery and reusable downstream contract | Result maps source/inputs/commands/outcomes/deviations, manual findings, applicable MR/staging disposition and cleanup; Codex final ACCEPT before dependent G058/G020 execution and G021. Public first-time invitation journey cannot be claimed complete until G058 qualification. |

## Validation and later Plan

Five bounded scenario groups: bootstrap/invariant; permission/succession/concurrency;
invitation/recipient/atomicity; departure/suspension/recovery; changed UX/delivery.
Use two Organizations, few Groups and existing owner/admin/member, RLS, audit,
identity, recovery and browser fixtures. Extend existing meaningful cases; no
role×scope×invite-state Cartesian campaign, duplicate oracle or full local gate.
Pin executable affected selectors from the first batch, exact module/CWD and changed
input provenance; missing existing diagnostics are not grounds for blind retries.
Use required hosted profiles honestly, temporary local evidence and normal cleanup.

Later Claude Plan inventories actual accepted G018 mutations/triggers/lock order,
remaining trusted bootstrap/member/retirement/cascade consumers, finite capability
and invitation-proof boundary, smallest database enforcement, legacy transitions,
minimal invitation/consumption/restriction persistence, audit ownership, route/field/
lifetime/rate/list limits and focused selectors. No migration/route number or new
vendor is prescribed. Identify unavoidable major policy/trust deviations before
implementation; routine filenames/SQL/lock design remain Claude’s planning work.

Architecture/security/data tier; provisional B7/checkpoint4, zero attempts. Owner
has authorized this definition only. Stop before G019 Claude planning/implementation;
final G018 dependency ACCEPT, later phase authority, Plan review, current bindings,
resources and supervisor execution release are still required. G017/G018 continue
undisturbed; G058 stays definition-only and no other roadmap mandate expands.

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

## Later Group authority replacement — G060, 2026-10-19

Owner selects [G060](G060-explicit-group-authority-spec.md) as a new reviewed increment:
ordinary Group management defaults to actual Group owner/admin, with a default-off
Organization-owner-controlled administration override; Organization staff metadata
visibility is separate from content. This supersedes R1's unconditional Organization
Group administration in G060, not retroactive alteration of accepted G019 evidence.
Group-owner-only powers, last-owner protection, invitation recipient proof and
current issuer-authority rechecks remain. Combined Organization+Group admission
is not included in this amendment.
