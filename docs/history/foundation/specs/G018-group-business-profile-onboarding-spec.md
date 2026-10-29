> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G018 — Group and Business Profile onboarding work specification

Goal: [G018](../goals/G018-group-business-profile-onboarding.md)

Epic: EPIC-007 — Group and Business Profile onboarding

State: Defined; dated owner mandate below authorizes planning/delivery under
normal predecessor, Plan, binding and supervisor-release gates

## Purpose, authority and inspected baseline

Expose a reusable authenticated Group onboarding flow over accepted UUID tenancy
and derived Business Profile access. A Group's creator starts as its owner;
onboarding remains unfinished until exactly one profile is attached. Shared
profiles are Organization-owned context, never a third tenant container or role.

Owner “Spec” (2026-10-15) authorizes this definition only. EPIC-005/006 autopilot
still ends at G017: a ready Spec or G017 implementation MR does not authorize
G018 planning/execution. New explicit owner phase authority is required. No G019
definition/work is included; the G021 product-feature barrier remains closed.

Inspected `staging` at `7d067493391bb65eec04d8bfb227c91ee3d4fad0`:
G005 `tenancy.Service.CreateGroup` already creates Group plus creator owner in one
transaction, but is a trusted primitive, not an authenticated creation capability.
G006 supplies profiles/at-most-one association and separate link/unlink primitives;
new-profile creation and link currently have independent transaction ownership.
G008 migration 00006 revokes runtime Group/profile/association INSERT and
association DELETE; its creation grants cannot be presumed available. The current
guarded facade has no Group-create/onboarding-completion operation. Completion is
not yet enforced; unlink explicitly delegates completed-Group rules to G018/G019.
Reconcile actual accepted G017/predecessors before a later Plan; reserve no migration.

Canonical sources: AGENTS/CLAUDE, Architecture §53.4/§53.5 and current identity/
onboarding/UI boundaries, SECURITY, DECISIONS, ADR-001/003/004/005/006,
ROADMAP/PROGRESS, G005–G008 and G017 accepted artifacts, AUTHORIZATION,
DATABASE-ISOLATION, AUDIT, WEB-SHELL and PERFORMANCE. UI work explicitly reads the
repository Vestive skill; accepted Radix/bundled-font/appearance contracts override
its obsolete Base UI, marketing and Dark-only source guidance.

## Scope, prerequisites and non-goals

In scope: guarded Group create/list/read/slug handling; create-new or eligible
existing profile attachment/completion; resumable incomplete state; bounded
duplicate/retry handling; authorized profile list/read/ordinary edit; thin sibling
management pages and shared profile shortcut; minimal schema/grant/RLS/audit
changes and focused verification. Execution depends on final G017 ACCEPT and all
transitive dependencies, a later owner phase release and normal Plan/input gates.

Exclude public signup, Organization creation or membership grants, invitations/
member administration/ownership transfer (G019), completed-profile replacement/
unlink, deletion/merge, billing and privileged settings, richer business fields,
AI discovery, commercial service enrollment, consultants/advisors/workshops, product
features, new auth/domain/branding behavior, generic workflow/idempotency framework
or CI redesign. A same-name business in another tenant remains a separate record.

## Requirements and invariants

### R1 — Current authority and authenticated Group creation

The creator is the verified current global user in the trusted active Organization,
with live Organization membership and the reviewed Group-create authority. Neither
raw Organization/user IDs nor hostname/login success grants membership. Pin the
finite authenticated creation/attachment/list operation and capability matrix in
the Plan; G005 trusted primitives and `organization.read` are not implicit mutation
authority. Any required capability/RLS/grant addition needs explicit review using
existing roles, never new profile roles or unrestricted provisioning access.

Superseding owner D-1, 2026-10-17: `group.create` permits current `org_owner`,
`org_admin` and `org_member` membership in the active Organization to create a
new Group of their own. This is generic authenticated self-onboarding; no
Organization elevation or authority over existing/other Groups is granted.
A Group role without valid Organization membership is still insufficient.

Create Group and its verified creator's initial `group_owner` binding atomically.
Every existing Group requires at least one owner; multiple owners are permitted
after later authorized promotion. This is not an exactly-one-owner invariant. Failure
leaves neither; creation grants no Organization role and writes no second owner
column. Preserve name/slug validation, suggested editable kebab-case slug, reserved
names and Organization-local uniqueness. Suggestions are advisory; simultaneous
create/edit collisions cannot overwrite a Group. UUID remains canonical through
name/slug changes; another Organization may use the same permitted Group slug.

### R2 — Exactly-one completion with atomic profile creation/attachment

An authorized newly created Group may remain unfinished/resumable with no profile.
It cannot be rendered/acknowledged as completed until exactly one same-Organization
BusinessProfile association exists. Preserve database at-most-one constraints;
derive completion from authoritative facts or keep any explicit status transactionally
consistent, never a browser step/localStorage flag or independently stale marker.

Creating a new profile and attaching it to the target incomplete Group must commit
together, including completion state if persisted and required audit. Failure or
a committed revocation that wins the accepted authorization/commit ordering
leaves the existing Group unfinished and no orphan profile/association from that
attempt. Preserve the existing guard linearization; a revocation waiting behind
a legitimately held mutation guard is not falsely reported as having won. Existing selection attaches only
after fresh eligibility and target-Group authority are proved in the committing
transaction; no check-then-link TOCTOU. Concurrent finish attempts yield exactly
one agreed completion or a safe conflict, never a second profile or overwritten link.

Incomplete state belongs to the target Group, not the global user or all of the
Organization's Groups. Resume only currently authorized incomplete Groups;
unfinished work in one must not block access to another completed, authorized
Group or create a global onboarding flag. Existing resource capabilities still
govern what is available in each Group.

The Plan inventories every reachable application/runtime completion and association
mutation path, including legacy trusted primitives and proposed endpoints, and
shows how no supported path removes/reassigns a completed Group's profile. Do not
expose unlink/replace as an onboarding retry. Explicit DBA/migration/repair paths
remain restricted and documented; broad runtime creation/deletion privileges or
RLS bypass are not acceptable shortcuts. No completed-profile lifecycle is added.

### R3 — Existing-profile eligibility and context correctness

Existing profile selection requires current active-Organization authority through
an eligible **pre-existing attached Group** owner/admin path or qualifying
Organization role under `business_profile.select_for_onboarding`. The new unattached
Group cannot authorize selection of an unrelated profile; submitting/guessing a
UUID or sharing a legal/display name proves nothing. Ordinary member/read-only
access alone is insufficient. A valid second management path may survive removal
of another, but removing the last path revokes eligibility without mirrored grants.

Keep existing context semantics: Group-derived profile checks consider the selected
Group only when one is selected, and all the user's eligible Groups in legitimate
Organization-only context. Organization owner/admin profile paths follow accepted
policy independently. “My Business Profiles” deduplicates manageable profiles in
the active Organization; it is not a cross-tenant directory and does not expose
all readable/member-only profiles as selectable. If the sibling management flow
uses Organization-only context, establish it through the reviewed current request
binding, not by ignoring an existing selected-Group restriction or trusting a raw
browser selector. Never clear/switch an MCP connection as a web-navigation side effect.

List results are not authority for later selection/write. Re-evaluate membership,
role, existing association, policy, session and resource ownership at the actual
operation, with inherited locking/ordering and non-disclosing errors. Profile
sharing never reveals sibling Groups, people, associations or private work.

### R4 — Shared profile editing, finite fields and UX

Use existing validated fields: Group name/slug and BusinessContext display name
(required), legal name, description, industry and website (optional). Existing
length/format bounds apply; website data is not fetched, a verified domain or
authority. No billing/contact/address expansions, external enrichment or arbitrary
HTML. Empty/invalid input cannot persist a partial create/complete/edit operation.

Expose Groups and Business Profiles as sibling areas within the current tenant.
A Group's **Edit Business Profile** action opens the same canonical profile page
as the sibling area; editing one shared record changes the context read by every
authorized attached Group. Do not create a separate Group-owned editor/copy.
Show incomplete state and an authorized resume action; finish only on core-confirmed
completion. Existing profile selection and new-profile creation are explicit choices.

Profile reads/edits always use current capability-specific policy. Group owner/
admin ordinary edits honor `group_admins_edit_business_profiles`; members have
read-only access and Organization paths retain their existing independent authority.
The flag does **not** disable `select_for_onboarding`. Losing a management path or
flag change cannot be bypassed through a shortcut or a previously rendered form.
No new privileged/billing/ownership action follows from ordinary edit permission.
Preserve existing concurrent-edit semantics unless a scoped reviewed change is
necessary; disclose that behavior rather than claiming unimplemented protection.

Use accepted Vestive/Radix components, bundled fonts, semantic tokens, tenant shell
and Light/Dark/System behavior. Render inputs as escaped data, use current CSRF/
same-origin session protections and preserve visible focus, labels/error association,
keyboard access, reflow and reduced motion. Routing/branding is presentation;
G016/G017 identity/MCP bindings remain unchanged.

### R5 — Resume, retries, audit and bounded query work

Pin double-submit, interruption/unknown commit and retry behavior for Group create
and both completion branches. Return/recover the same scoped committed result for
a duplicate logical operation, or a truthful conflict with an authorized resume
path; do not create duplicate Groups/profiles or silently select another Group.
An intent/retry key is not authority and must bind actor/Organization/target/payload
as applicable. Reuse existing transactions/constraints and the smallest necessary
state; no generalized idempotency service. Distinct intentional creations remain
possible; names are not identity or automatic deduplication keys.

Use existing audit ownership rules for Group/role/profile/association changes and
denials, with no duplicated trigger/app events or private profile payloads in logs.
Runtime grants remain the minimum needed, parameterized SQL and RLS remain defense
in depth, and unavailable policy/session state denies rather than inventing defaults.

Assess create/finish transaction length, eligibility-list query count/deduplication,
association/slug indexes and revocation/contention ordering under PERFORMANCE.md.
Use bounded fixture query/plan observations for new paths; escalate unexpected
scans/fan-out/helper work or deadlocks. Do not query each candidate profile through
an unbounded per-row round-trip loop or begin a benchmark/index campaign.

## Acceptance criteria and verification

| ID | Required outcome | Objective verification |
|---|---|---|
| AC1 | Group creation requires current reviewed Organization authority and atomically assigns the verified creator as its initial owner under the one-or-more-owner model. | Guarded API/service and real runtime-role tests cover permitted actor, nonmember/foreign/disabled/revoked/session failure and forged creator selectors. Inject failure between Group/binding writes; assert neither row, no Organization grant and no extra owner state. |
| AC2 | Slug suggestion/editing preserves UUID identity and race-proof scoped uniqueness. | Reuse existing validation/reserved cases with actual UI/server submission; same-tenant concurrent collision yields one valid record/conflict, same slug in another tenant is independent, and rename changes no UUID/profile/role association. |
| AC3 | New-profile completion is one authorized atomic outcome. | Create+attach+completion succeeds as one coherent state; validation, audit/store fault and deterministic precommit revocation that wins the accepted guard ordering leave the Group incomplete and no new orphan profile/link. Two completion attempts cannot attach two profiles or overwrite the winner. |
| AC4 | Existing selection uses pre-existing current management paths without manufacturing access. | Cases cover qualifying attached Group owner/admin, qualifying Organization authority, member-only/foreign/unrelated profile and newly owned unattached Group. Last qualifying path removal that wins the guarded ordering denies with zero link/completion effect; an independent remaining path still works. |
| AC5 | Completed state cannot drift from exactly one profile through supported application/runtime paths. | Source/path inventory and database-backed negatives cover direct completion attempts, stale browser markers, forbidden completed unlink/replacement and raced association mutation. Existing incomplete Groups resume without being reported complete; no broad RLS/grant bypass is added. |
| AC6 | Lists and shared profile reads/edits preserve active tenant, selected scope and edit-policy semantics. | Use one shared profile/two Groups and a foreign profile: manageable lists deduplicate only permitted records; selected-context restrictions and legitimate org-context flow are distinct. Flag off denies Group-derived edits but preserves eligible selection/read; Organization paths and membership changes behave as accepted, with no sibling details leaked. |
| AC7 | Interrupted/double-submit flows resume without duplicate logical results or false success. | Repeat Group-create and both finish branches, including lost-response/unknown-commit and changed-payload cases. Assert pinned scoped retry/conflict outcomes, same authorized UUIDs where committed, no orphan/duplicate records and honest unfinished/unknown presentation. |
| AC8 | Generic onboarding and sibling pages use one canonical profile editor and preserve accepted accessibility/security. | Focused actual-host browser flow covers Group→new/existing profile→complete, resume and both paths to the shared editor. Required manual feedback precedes final qualification; keyboard, error labels, narrow reflow and existing appearance/portal tests cover changed UI. No signup/invite/product flow or web-triggered MCP context change appears. |
| AC9 | New guards/schema/contracts/audit and lookup footprint remain bounded and least privileged. | Review actual migration/grants/RLS and closed operation mapping; real runtime positives/negatives and focused audit tests prove intended effects and privacy. Record new query/locking observations and any risk disposition; no independent profile roles, third tenant tier or infrastructure/framework expansion. |
| AC10 | Evidence, integration and cleanup satisfy normal acceptance gates without duplicate qualification. | Result maps AC1–AC9 to reviewed/tested content, relevant inputs, commands/outcomes, manual feedback and deviations; record actual selected local/MR/staging checks and scoped cleanup. Failed/pending evidence stays explicit and Codex final ACCEPT is separate from merge. |

## Verification strategy and expected artifacts

Five scenario groups: (1) authenticated create/owner/slug; (2) new/existing profile
completion with concurrency; (3) manageable lists/shared edits/context/flag;
(4) retry/resume and minimal accessible UX; (5) RLS/audit/queries/integration/cleanup.
Reuse existing two-Organization G007/G008 fixtures with one profile shared by two
Groups, one unrelated profile and focused role/revocation barriers. Do not take
every role×profile×context×theme permutation or run a full stack/suite per branch.

Pure validation/state mapping uses unit tests; actual runtime PostgreSQL proves
authorization/uniqueness/atomicity/completion, and selected HTTP/browser cases
prove thin-adapter composition and UX. Reuse accepted unchanged RLS/identity/MCP
proofs rather than rebuilding their full oracle. Required manual feedback comes
before final qualification. Plan identifies focused commands and actual CI
profile selected by changed paths; no broad runner/harness/control changes.

Expected artifacts: reviewed Claude Plan; bounded guarded core creation/completion/
list operations and necessary schema/grant/RLS migration; thin BFF/pages/shared
profile editor; focused tests and finite audit/contract documentation; Result and
cleanup receipt. Git-first inputs and actual tested/reviewed content bind evidence;
record failed/interrupted outcomes and reuse unchanged passes explicitly. Keep
compact durable Results/CI artifacts, with local raw evidence temporary. No new
evidence registry or mandatory upfront benchmark campaign.

## Decisions, deviations and completion gates

Later Plan review must pin: authenticated Group-create/new-profile/association
capability matrix; minimum runtime grants/RLS and creator-owner bootstrap;
completion representation and every reachable mutation guard; actual transaction/
revocation ordering; scoped retry/unknown-commit contract; legitimate Organization-
only management context and public HTTP/MCP surface where required by canonical
contract conventions. Existing trusted primitives alone cannot resolve these
authorization gaps. Escalate owner-owned role-policy/security choices rather than
silently borrowing unrelated membership/ownership privileges.

No blanket deviations. Internal composition, bounded UI route naming and completion
representation may vary if all outcomes and accepted authority remain identical.
New role hierarchy, broader profile grants, completed-profile reassignment/removal,
weaker RLS or deferred completion evidence requires explicit scope/owner review.
No early implementation while a decision is pending.

G018 planning and implementation remain explicitly held for new owner authority,
even when G017 opens an MR or receives final ACCEPT. Following release, execution
also requires all dependency ACCEPTs, Plan ACCEPT/reconciliation, decisions,
Git-first bindings, budget/resources and supervisor release. Final completion
requires AC1–AC10, reviewed diff/Result, applicable staging disposition and scoped
cleanup, then Codex final ACCEPT. No G019 definition/planning/execution or G021
product-feature waiver is authorized by this Spec.

## Query tooling note — 2026-10-16

Follow Architecture §27.1 (not included):
Goose handles migrations; incremental sqlc adoption for ordinary queries is a
separate bounded follow-up, not a required deliverable or new acceptance criterion.
Query generation, if separately adopted before this Goal is released, must fit
inside the existing caller-supplied guarded transactions. Group/owner creation,
profile attachment, completion, retry and audit effects keep their atomic and
current-authority guarantees; generated methods are not a second mutation path.
No inherited auth/RLS rewrite or sqlc adoption prerequisite is added. G018's
explicit planning/implementation owner hold remains unchanged.

## Owner mandate extension — 2026-10-16

Owner **“I think we can authorize autopilot up to 19 spec”** supersedes this
artifact’s earlier definition-only/G018 planning and execution hold. Under the
recorded exact mandate (not included),
G018 may plan provisionally at G017’s first actual implementation MR, then proceed
through normal reviewed delivery and closeout after dependency ACCEPTs, Plan
ACCEPT, exact inputs, decisions/budget/resources and supervisor execution release.
No additional routine owner proceed prompt is required. Existing functional
requirements, provisional B7/checkpoint4 and escalation/validation rules remain.

Codex may define G019 after G018’s Plan settles the relevant interfaces, but must
**stop before G019 Claude planning/implementation**. G020+ is not authorized.
This is a bounded phase mandate, not live provider/deployment authority or a
waiver of the G021 barrier. Earlier authority text remains historical.


## Dependency clarification — G016/G057 qualification split (2026-10-16)

Under the owner-approved [split](../results/G016-G057-qualification-split.md),
G016 final ACCEPT covers the reviewed implementation/local security boundary;
P-1/P-2 actual-provider/deployed HTTPS proof moves to G057 before G021. This Goal
consumes accepted predecessor routing/session contracts and does not wait for
G057. Its own manual visual/UX feedback, security, Plan, dependency, input,
budget and supervisor-release gates remain. No live custom-host readiness claim,
new infrastructure authority or waiver of implementation defects follows.


## Owner amendment — ownership cardinality and succession (2026-10-16)

The owner approved **one or more owners, at least one required**, and requested
this amendment for G018/G019. Initial Group creation and its creator-owner binding
commit together; failure leaves neither, and an unfinished Group is still owned.
No maximum-one-owner index/validation, duplicate admin binding, ownership column
or BusinessProfile role is introduced. Exactly one Business Profile per completed
Group remains a separate, unchanged rule.

G018 AC1 must demonstrate initial ownership/rollback and avoid an API/schema
contract preventing later additional owners. [G019](G019-scoped-invitations-membership-ownership-spec.md)
owns the additional-owner, atomic transfer, normal last-owner departure/demotion/
removal and database/runtime enforcement paths for both Organizations and Groups.
Its UI uses current server eligibility; it never authorizes from a cached count.
Emergency suspension remains immediate, preserving the binding but denying the
suspended user access; restricted verified recovery is separate from normal
retirement. G018 adds no transfer/member UI and does not wait for G019 to execute.

This owner instruction authorizes provisional G019 Goal/Spec drafting now,
superseding the earlier wait-for-G018-Plan condition only for definition; interfaces
remain provisional until reconciled with G018. Stop before G019 planning or
implementation. Current G018 B7/checkpoint4, dependencies and release gates remain.

## Owner decision D-1 — Group creation roles, 2026-10-17

**Historical initial choice, superseded by the later self-service amendment below.**

The owner answered “Yes” to Codex's recommendation of **Organization owners and
admins**, rather than all Organization members. Apply that finite mapping to
Group creation, its UI eligibility, guarded capability and database policies.
The authenticated creator becomes the initial Group owner atomically; no new
Organization role or separate owner store is introduced. This resolves D-1 only.
The r0 correction loop, final G017 acceptance/reconciliation, revised Plan review,
binding, B7/checkpoint4 and explicit execution release still apply. No earlier
source is retroactively accepted, and the G018 Plan-acceptance owner check-in
and G019 stop boundary remain.

Scope clarification: D-1 governs this Goal's administrative creation in an existing
Organization. It does not prohibit future authorized client self-onboarding into
the client's own Group without Organization-admin elevation. G042/G043 govern
that separate enrollment policy/flow; no public signup or wider creation role is
added to G018 by this clarification.

## Superseding D-1 — Generic self-service Group onboarding, 2026-10-17

After questioning the owner/admin-only restriction, the owner confirmed that
generic Group self-onboarding belongs in the foundation: “Right. I'm beginify
case a group of a client that a user can sign up and setup.” This follows Codex's
explicit recommendation to permit existing Organization members to create their
own Groups without becoming Organization admins. It supersedes the earlier
same-day D-1 answer and the administrative-only scope interpretation above.

- `group.create` admits current `org_owner`/`org_admin`/`org_member` membership
  in the trusted active Organization, in Organization-only context. The verified
  creator becomes the new Group's initial `group_owner` atomically. No actor ID
  selector, Organization-role grant, extra owner store or other-Group privilege.
- Generic foundation UI supports that authorized self-service create→new/eligible
  Business Profile→completion flow. Ordinary members see the create action and
  can complete the Group they created. Existing profile selection still requires
  a pre-existing qualifying path or independent Organization authority; the new
  Group never authorizes an unrelated profile.
- Same-user global identity is not membership. Public signup, client enrollment,
  invite/open/request-access policies and their membership provisioning remain
  separately reviewed in G058 after G019 and before G021. G042/G043 compose
  later product/commercial enrollment with these accepted primitives; no public
  endpoint or implicit membership grant is added
  to G018. Agency/client terminology stays illustrative, not hard-coded generic UI.
- Revise only the affected creation capability/RLS/SQL/HTTP/UI/tests and binding
  assumptions in the pending Plan. In particular, a member cannot RETURNING-read
  a Group before its creator-owner trigger has run; pin the member-safe atomic
  insertion/read sequence under actual grants. Preserve settled profile-insert,
  bootstrap, locking, paging/recovery design except where this role change has
  a demonstrated effect; do not restart the Plan or reset B7/checkpoint4.

Remaining r1 review corrections, final G017 acceptance/current reconciliation,
revised Plan review, input binding/resources and explicit execution release still
apply. The G018 Plan-acceptance owner check-in and G019 stop remain unchanged.

## Owner amendment — overlap G018 implementation with G017 staging CI (2026-10-17)

Owner instruction: “We should start 018 impl before the hour long staging tests.
Anything broken, 018 workers can fix it.” This supersedes this Goal/Spec's earlier
requirement for final G017 ACCEPT **before G018 implementation** and the preparation-only
hold recorded in the G018 r5 review. G017 source is reviewed and integrated at
`3cc1822826754ff70cf6346809f62b417080832c`; staging pipeline `2958742766` is
still running. It is not recorded as passed or finally accepted.

G018 may execute its accepted r5 design on an isolated task branch once the
current-target R0 reconciliation, exact scope/input binding, initial Result,
resource readiness and Codex release are confirmed. Do not wait for G017 staging
CI/final ACCEPT solely to start that implementation. Other accepted dependencies,
security/architecture requirements, B7/checkpoint4, focused validation and the G021
barrier remain. G018 integration and final ACCEPT still require final G017 ACCEPT.

If G017's gate exposes defects, the same available workers may perform bounded
repairs after Codex records the failing revision/evidence, exact repair scope,
branch/writer and validation selection. Keep predecessor repair and G018 changes
attributable and their existing attempt ledgers separate; no budget reset, broad
rewrite or silent expansion of the G018 allow-list. Pause only work affected by
a changed contract or unsafe shared resource; preserve unaffected completed work.
No need to restart G018 or repeat passing checks for unrelated management changes.
G019/G058 planning/execution holds are unchanged.

## Owner clarification — predecessor retains staging repairs (2026-10-17)

The owner clarifies that staging pipeline failures remain the predecessor task's
implementation responsibility, as before. Its assigned owner handles scoped
repairs using the predecessor branch and attempt ledger, in parallel with the
successor's isolated implementation. Do not assign those failures to successor
workers by default or wait for predecessor qualification solely to start successor
work. Pause only affected/unsafe work; predecessor final ACCEPT still gates
successor integration/final acceptance. This supersedes the earlier wording about
the same available/successor workers fixing predecessor failures. Any worker
reassignment must be explicit and retain predecessor attribution, scope and budget.
No implementation restart, new test run, Plan rewrite or budget reset is required.

## Mechanical correction autonomy — owner direction, 2026-10-17

Owner “Make that change” authorizes the current mechanical-correction policy (not included):
a released implementer handles in-scope selectors/fixtures, approved-value catalog/
count/version pins and harness/import bookkeeping directly, preserving assertions
and recording focused evidence/Plan wording reconciliation in the Result. No
separate approval or full replan for each mechanical inconsistency. Exact scope,
security/contracts, unapproved schema/constraints, budget/native stops, required
validation and final Codex review remain. Apply at the current safe boundary; no
worker restart, new test campaign, frozen-input rewrite or budget reset. For held
Goals this changes guidance only, not execution/live/publication authority.


## Completion-adapter availability — 2026-10-17 security review

The existing AUTHORIZATION.md rule for denied exclusive-write requests applies to
both new/existing-profile completion modes now exposed by the adapter. Before merge,
provide nonqueueing admission by verified canonical subject across sessions before
the completion guard (at most one in flight per subject, bounded active state,
released on every outcome), plus bounded core request/transaction/cleanup lifetimes.
A BFF timeout or server header timeout alone does not establish these bounds.

Use existing request-context/read-deadline and transaction-cleanup mechanisms;
reviewed starting bounds are10s core request,5s decision and5s rollback cleanup,
with the existing separate denial-audit2s record/2s cleanup explicitly accounted.
Retain an earlier caller deadline. Admission exhaustion can use the existing closed
onboarding_unavailable503 response; no new public error enum is required. Preserve
unknown-COMMIT semantics and all session/CSRF/tenant/locked-facts decisions. No
pre-decision authorization shortcut, SQL lock-order rewrite, global DB timeout,
shared limiter framework or unrelated adapter campaign is included.

Claude supplies a short exact-file/entry-point addendum before this security code
change. Reuse the existing scoped files/dependencies where possible, identify any
actual scope addition for review, and extend only the relevant admission/timeout
cases. Prior UI correction evidence remains valid; this does not authorize a full
local test run or reset the current B7/checkpoint4 ledger.


### Accepted availability implementation clarification — 2026-10-17 UTC

The reviewed addendum (not included)
uses bounded process-local admission: at most one completion per canonical subject
per serving core process, maximum 256 active subjects per process. With N serving
processes the per-subject bound is N; no cluster-wide singleton is claimed or new
shared infrastructure required. The decision context must reach every completion
SQL callback as well as BEGIN/COMMIT; earlier cancellation, bounded cleanup and
truthful unknown-COMMIT handling remain required. Other Spec guarantees stand.

## Owner M1 UX amendment — 2026-10-18 UTC

Owner M1 feedback requests direct entry sign-in, inline Group-address conflict
feedback and dynamic blur-time slug hints. Approved bounded implementation uses
the existing guarded `/auth/login` POST for home and onboarding sign-in actions;
keeps `/sign-in` and all existing error/logout-return flows; moves definitive
create conflicts into the form with fresh authorized holder lookup; retains the
separate unknown-COMMIT recovery page without auto-retry or false success.

Add one authenticated same-origin GET BFF route `/groups/address-check?slug=…`,
reusing the current guarded exact-slug list. Closed advisory states are invalid,
in_use (a readable holder), or not_seen (filtered absence, never availability).
No holder names/IDs in JSON, no new core endpoint/capability, no schema/vendor or
global uniqueness oracle. Backend create remains authoritative under races.
Clear hints on edit; generation/value checks, finite request lifetime and closed
bounded responses prevent stale or malformed hints. Preserve accessible field
association, native submit and script-free behavior.

See the exact Claude proposal (not included) and
binding review/release (not included).
This dated amendment supersedes only the prior separate-conflict-page presentation
and adds the reviewed BFF surface. Original approved Plan/history remains intact;
no new standalone Goal/Spec or general authentication redesign. Remaining owner
M1 questions and mandatory MR/Duo/staging acceptance gates stay in force.
