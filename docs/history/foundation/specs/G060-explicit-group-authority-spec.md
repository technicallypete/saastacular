> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G060 — Explicit Group authority work specification

Goal: [G060](../goals/G060-explicit-group-authority.md).
State: Owner-selected definition, 2026-10-19; Claude Plan required before execution.
Decision: [ADR-008](../../../adr/ADR-008-explicit-group-authority.md).

## R1 — Three separate authority boundaries

Current Organization owners/admins may list Groups and read Group and ordinary
Business Profile metadata in their active Organization, regardless of Group
membership. Metadata means existing Group identity/name/slug/onboarding/profile
association and ordinary profile fields; it excludes Group objects, files, signed
URLs, object counts/previews, work data and protected object labels. Membership/
invitation rosters and tokens are administration, not general metadata. Billing,
payment/tax/legal settings retain their distinct existing privileged contracts.
Foreign tenants remain invisible. A shared Business Profile never opens sibling
Group content or membership rosters. Plan must inventory exact DTO fields, reads
and consumers rather than reclassify a content-bearing endpoint as “metadata”.

Default management requires current explicit Group owner/admin membership.
Organization owners/admins with a Group role receive exactly that Group role's
ordinary powers; their Organization role cannot upgrade a Group member to admin.
Creator-owner atomic bootstrap is preserved. Group joining remains invitation-only
through the existing authorized issuer/verified recipient consumer; recipient
consumption is a bounded exception to requiring membership before joining.

## R2 — Owner-controlled override, default off

One persisted Organization boolean permits current Organization owners/admins
bounded ordinary Group administration without joining. Default FALSE on fresh and
existing Organizations; missing/invalid/unavailable policy cannot grant the override.
Only a current enabled Organization owner may change it, including when the caller
has no Group role. An Organization admin who is a Group owner cannot set it.
Do not broaden the existing general policy-write capability to grant this setting
to admins; use a distinct checked operation/field or equivalent narrow boundary.

When enabled, permit ordinary Group metadata/onboarding/profile management and
ordinary admin/member invitation/role/removal administration under the existing
same-tenant/selected-context/ordinary-profile-edit rules. Reuse current approved
contracts; no destructive lifecycle or profile reassignment feature is introduced.
Group ownership administration, owner-role invitations, last-owner bypass and
Group content/object access never follow from this setting. Do not create a
synthetic Group membership, self card, Leave action or owner controls.

The override cannot issue or consume a self-targeted grant or change the actor's
own Group role to obtain content access through Organization authority. Joining
or elevating the actor still requires an independently authorized Group owner/admin
and existing verified invitation/role rules. Ordinary Group-role administration
continues under its existing restrictions. Delegated membership administrators
can grant access to other people: this is an explicit trusted administration
power, not a guarantee against collusion. Explain its scope in the setting UI.

| Caller, no Group membership | Metadata reads | Ordinary Group management | Group content/objects | Group owner powers | Change override |
|---|---|---|---|---|---|
| Organization owner, override off | Yes | No | No | No | Yes |
| Organization admin, override off | Yes | No | No | No | No |
| Organization owner/admin, override on | Yes | Yes, bounded above | No | No | Owner only |
| Organization ordinary member | Existing explicit authorized paths only | No | No | No | No |

For every caller with explicit Group membership, union only authorized metadata/
management paths with the actual Group role's existing content powers. An explicit
Group member who is also Organization admin still has member-level content powers;
turning the override off removes only the additional administration path.

## R3 — Content and current authority at every boundary

Remove implicit Organization owner/admin shortcuts for Group-scoped object reads,
uploads, deletes, signing and other content operations. Read-only metadata access
or selecting a Group never creates content authority. For Group objects, evaluate
actual Group membership, correct Group role, uploader rules and exact tenant/context.
Keep Organization-scoped resources and their existing Organization policies distinct.

Inventory policy/capability discovery, execution, RLS/helpers/security-definer guards,
HTTP/BFF, MCP Group listing/selection/tools, onboarding/profile paths and storage.
Organization-only and selected-Group contexts must both enforce the matrix; a
foreign or selected-other Group remains refused. Lists, pagination/aggregates,
error details and signed URLs must not reveal content through metadata visibility.
Read SelfMembership truthfully; a management override is not a self membership.

Evaluate current setting, enabled user, Organization/Group roles and context at
the authoritative transaction, with a documented lock/order consistent with existing
membership/policy mutations. A disabling setting change/demotion/removal that has
committed cannot leave stale session/MCP-catalog/client flags granting later calls.
Test concurrent policy/role changes and invitation consumption: an outstanding
invitation issued solely under an override cannot bypass a later loss of issuer
authority. Existing G019 owner, departure, single-use and audit guarantees remain.

## R4 — Persistence, audit and UI

Use a reviewed forward migration and bounded current storage/API extension; do not
edit accepted historical migrations or invent a new role store. Plan identifies
default/backfill, rollback limits, runtime grants/RLS and recovery/restore behavior.
Only trusted owner-authorized writes set the override; direct runtime SQL cannot
bypass that restriction. Existing data receives false; never infer opt-in from
old broad administration behavior. Preserve owner bindings and content grants.

Audit successful policy changes (actor, canonical Organization, old/new boolean)
and privileged override-based administration using existing conventions. Bind audit
and mutation atomically under existing mandatory-write rules; preserve unknown
commit outcomes and normal refusal privacy. No token/credential payload logging.

Use existing Organization settings UI: owner-only switch with clear scope text,
read-only effective status for admins where useful. Group/profile metadata pages
remain read-only without management authority; membership pages distinguish actual
Group roles from bounded override administration. Do not show content navigation or
owner actions merely because metadata is visible. Keep keyboard/form/error/light/
dark conventions. Fold paused 5f94a705 only after design review; it currently exposes
old unconditional administration and has no validation. No wholesale cherry-pick.

## R5 — Validation and acceptance

Claude Plan maps the nine criteria below to existing checks and exact executable
selectors. Reuse/extend current policy, runtime-RLS, storage, membership/invitation,
HTTP/MCP, browser and attack tests; explain distinct guarantees without duplicating
suites. Trace actual changed migrations/policy/audit/catalog counts to their callers
and pins. Early risky slice proves metadata-vs-content and owner-only setting writes
before broader UI work. Preserve every failure and correction with source attribution.

| Criterion | Required proof |
|---|---|
| AC1 Metadata visibility | Org owner/admin without Group binding lists/reads same-tenant Group/profile metadata; ordinary unauthorized, foreign and selected-other cases preserve privacy. No content/roster leakage. |
| AC2 Default separation | Fresh/upgraded default false; Org staff no-binding cannot manage Group or access content. Actual Group roles retain their bounded powers. |
| AC3 Owner-only setting | Owner can enable/disable; Org admin, Group owner/admin and ordinary caller cannot via UI, forged API, generic policy patch or direct runtime SQL. Invalid/stale policy fails closed. |
| AC4 Bounded enabled mode | Staff can manage ordinary Group metadata/members/invitations when enabled, with no membership creation for actor, self-escalation, owner powers or content access. |
| AC5 Content isolation | Real runtime SQL, core/HTTP/MCP and storage deny Group object read/upload/delete/sign absent explicit Group role under either setting, with paired genuine Group-member positives. Shared profile and context selection confer none. |
| AC6 Live revocation/concurrency | Disable/demotion/removal denies later management and stale invite consumption; serialized races have coherent audit/no partial grants. Genuine Group membership survives turning override off. |
| AC7 Data/audit continuity | Forward upgrade/backfill, rollback/recovery disposition, runtime grants/catalog checks and mandatory audit success/failure/unknown-commit proofs. Last-owner, creation and enrollment regressions remain valid. |
| AC8 UI and machine surfaces | Automated focused settings/metadata/membership light/dark/keyboard/refusal tests; owner/admin/member catalog and execution parity; truthful nonmember presentation. Manual review owner-waived/not performed. |
| AC9 Integration/closeout | Current-source review/fast/Duo, exact merged-SHA local full qualification owned by existing coordinator, complete failure disposition, cleanup receipt and adopter docs. No automatic staging cloud full run or duplicate campaign. |

Use current workflow: focused proof + source/Duo/feature-MR fast → staging merge →
coordinator-owned full local qualification. Final ACCEPT follows qualification and
cleanup; no separate manual owner preview/sign-off required. Hosted full remains
available for separately authorized RC promotion, not a mandatory duplicate here.

## Scope and continuity

This task supersedes the old automatic Organization-admin Group-management and
Group-content shortcuts only in the reviewed G060 implementation. ADR-004 cardinality,
explicit-role derivation and G019 owner continuity remain. Old Specs/Plans/Results
retain their historical evidence; append supersession links rather than rewriting
accepted contracts. G058 closes against original enrollment requirements, with this
tracked policy/UI follow-up outstanding. G060 final ACCEPT gates G055 completion,
G021 and OS001/SaaStacular. G020 planning and OS001 execution remain held.

Combined admission to both Organization and Group for a first-time/nonmember invitee
remains separately scoped; no new admission permission is implied by this task.


## Owner amendment — invalid-invitation presentation, 2026-10-19

Owner approved Option B after the 1–1.5h combined incremental estimate: fold the
minimal Organization and Group invitation message fix into G060's current UI work.
For the Organization callback's existing invitation_mismatch status, show a fixed
“Invalid invitation link” message, optionally suggesting switching accounts or
requesting a new link. For the Group invitation page, malformed IDs and core's
closed not_found (wrong recipient/nonexistent/foreign) render the same generic
message locally. Never disclose the intended account/email or distinguish whether
the invitation exists. Preserve sign-in/recipient/verified-identity checks and all
successful-invite behavior. Other authentication, unavailable and context errors
keep their existing handling; do not mislabel infrastructure failures as invalid
invitations. Accept POST refusal, callback redirect/routes and core remain unchanged.

Exact added paths: web/lib/enrollment-client.ts,
web/tests/unit/enrollment-client.test.mjs, web/app/invitations/[id]/page.tsx.
web/tests/browser/membership.spec.ts is already approved. Required evidence:
message unit assertion and a focused Group wrong-account browser case, with the
same generic presentation for nonexistent/malformed IDs and existing intended-
recipient success retained. Reuse current browser/build setup and relevant checks;
no duplicate campaign. No new role, endpoint, provider or authorization/error-code
contract. Direct presentation amendment, no separate Plan cycle. Preserve accepted
Plan history; same in-flight attempt4 and budget, no reset. A context handoff alone
spends no new attempt; a later post-submission correction follows normal counting.
