> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G007 — Trusted context and shared policy evaluation work specification

Goal: [G007-trusted-context-policy.md](../goals/G007-trusted-context-policy.md)
Epic: EPIC-003 — Authorization and isolation
State: r3 — Accepted-predecessor reassessment; ready for Claude planning, separate Plan approval gates execution

## Goal, Rationale and Authority

Turn the G005/G006 membership and association facts into enforceable core authorization. A global identity alone grants no tenant access; a Group link alone grants no identity or membership. One server-owned resolver/evaluator must support future HTTP and MCP callers without putting policy in web, intelligence or provider metadata.

Follow canonical ARCHITECTURE's **Canonical Multi-Tenant Semantics (v0.1 Foundation Override)**, accepted ADRs and the [archetype tenancy contract](../../../../skills/saas-foundation/references/tenancy-authorization.md). That Skill contains historical BusinessEntity/direct-grant examples; canonical ARCHITECTURE, ADR-004 and the current matrix below supersede those exact portions. Do not revive independent profile roles or rewrite the Skill as part of G007. Follow AGENTS/CLAUDE, BUDGET_POLICY, ROADMAP and PROGRESS. SECURITY.md is currently empty; its presence is not security evidence and filling it is not a new workstream here.

## Inspected State and Prerequisites

Reassessment observation: clean, pushed staging `6e38fdadd9fa1b2fe551d77314d0f1f5dd658c55`. G004–G006 are accepted and integrated. The [G006 final review](../results/G006-implementation-codex-review.md#final-goal-acceptance--2026-10-10) binds tested implementation `08d279f`, Result `071a20b` and integration `4486b41`; its canceled replacement staging pipeline is an accepted disposition, not hosted test success. No G007 Plan or implementation exists. This observation is not an execution snapshot.

Actual baseline contracts:

- `core/internal/identity.Service.Session` resolves opaque application credentials against current PostgreSQL session/user state without requiring a live WorkOS call. `SessionInfo` is an exported result, not proof supplied by an arbitrary caller. Preserve accepted credential hashing, expiry, disable/revocation, unknown MFA assurance and privacy.
- G005 supplies UUID Organization/Group identities and the sole `app.role_bindings` authority with membership source, six Organization/Group roles and read-only membership projections. Group creation atomically adds its creator-owner binding. Model primitives intentionally carry no authenticated actor or capability enforcement.
- G006 migration `00004_business_profiles.sql` adds Organization-owned `app.business_profiles`, Group-unique `app.group_business_profiles`, and SELECT-only `security_invoker` `app.business_profile_access_paths`. Access paths expose current raw Group-role facts; they are not complete capability decisions and do not establish Organization membership. Several Groups may share a profile; an unfinished Group may have none.
- The tenancy Service has 29 exported model methods, each with an independent READ COMMITTED transaction. Session resolution also does not span a later protected mutation. A permission check followed by an existing separate write is not proof of R6's atomic ordering. Pin the minimum transaction integration in the Plan; no broad service rewrite is authorized.

Before execution, bind the actual accepted G004–G006 Results/reviews, session/store interfaces, schema/grants and applicable regressions. Preserve G005 own-view/PUBLIC-ACL clarifications and G006 accepted deviations D1–D5; do not regress read-only projection evidence to an incorrect universal SQLSTATE. G006 website/context fields are data, never fetched or trusted as authority. Unknown/foreign identifiers remain indistinguishable. Accepted lost-commit-reply behavior may return unavailable after a durable commit; a failed response is not proof of rollback. Pin safe reconciliation/retry claims rather than silently adding idempotency.

Do not rewrite existing migrations or transplant another Goal's execution snapshot. Reconcile the actual CI/harness at the new committed execution baseline, including mount-safe evidence streaming and pre-batch checks. G051 is independent: its unaccepted proposals are not dependencies or guarantees. Later integration changes require explicit shared-input reconciliation, not a blanket repeat of unchanged tests or a wait for G051 acceptance.

## Scope and Non-goals

In scope: trusted request context, a closed capability matrix, the one policy flag in R4, minimum guarded internal core operations and meaningful tests/evidence. No new public tenant route, browser screen, MCP tool, transport context switch, hostname resolver, invitation/admin/onboarding workflow, RLS or storage implementation. No billing/tax/payment/ownership side effects, generic policy engine, copied roles, provider changes or product implementation. G008–G019 retain their named work; G021 remains closed.

## Requirements

### R1 — Trusted and Bounded TenantContext

Core resolves an immutable request-scoped security envelope containing the authenticated global user UUID, active Organization UUID and optional Group UUID. It is not a Tenant table, browser-owned context, durable session switch or second identity. Resolve the subject through the accepted G004 authenticated-session boundary; validate current enabled user/session state, expiry and revocation. Reject fabricated, expired, revoked, foreign-user or unauthenticated contexts.

Organization/Group IDs supplied for resolution are selectors only. Organization existence and a current Organization membership are required for every tenant operation, including users with Group-derived profile access. A Group binding without an Organization binding cannot establish tenant access. Org-owner/admin or a current binding in the selected Group can establish Group context; all selected objects must belong to the active Organization. Neither Group roles nor associations create Organization membership.

Only the resolver produces usable trusted contexts. Caller-constructed structs, reused contexts for another authenticated subject/request, browser headers, cookies carrying role/tenant assertions, provider Organization/role claims, email/domain, names/slugs and AI/model data cannot satisfy this contract. All-zero/malformed IDs and inconsistent context fail closed. No default Organization fallback or automatic grant. Actual hostname verification and MCP switching remain future adapters; G007 does not pretend an arbitrary Host header is a trusted selector.

### R2 — Closed Policy and Scope Gates

Evaluate capabilities from current RoleBinding facts, same-Organization GroupBusinessProfile links and the R4 policy. Before the role matrix, require a valid context and exact target ownership/type. Unknown capabilities, roles, sources, target kinds or inconsistent records deny; errors/cancellation/unavailable identity or database data cannot produce an allow. Global authority, numeric role ranking and substring/prefix matching are forbidden. The three Organization capabilities and audit.read target the active Organization; the three Group capabilities target an exact Group; the four BusinessProfile capabilities target an exact BusinessProfile. A mismatched target kind denies even when the caller holds the named role.

Group-target operations with selected Group context require that exact Group; without it the matrix's own/all semantics apply to the exact target. For BusinessProfile inheritance, selected Group context permits only the associated selected Group; Organization-only context may use the user's associated authorized Groups in that Organization. Organization authority remains independently evaluable for a same-Organization profile, including unlinked profiles. Group-derived authority never authorizes unrelated Group records. No independent BusinessProfile role/grant is recognized. A shared profile read reveals only the authorized profile, not sibling Group membership/association details.

### R3 — Capability Matrix

`Y` means a role at the target scope; `all` means any Group in the active Organization; `own` means the exact Group binding; `linked` means a current authorized Group plus its current BusinessProfile association; `linked+flag` also requires R4's flag. Dashes deny. Role grants never cross Organizations. The Organization-membership and context gates apply to every cell.

| Capability | org_owner | org_admin | org_member | group_owner | group_admin | group_member |
|---|---|---|---|---|---|---|
| organization.read | Y | Y | Y | — | — | — |
| organization.manage_members | Y | Y | — | — | — | — |
| organization.manage_policy | Y | — | — | — | — | — |
| group.read | all | all | — | own | own | own |
| group.update | all | all | — | own | own | — |
| group.manage_members | all | all | — | own | own | — |
| business_profile.profile.read | Y | Y | — | linked | linked | linked |
| business_profile.profile.update | Y | Y | — | linked+flag | linked+flag | — |
| business_profile.select_for_onboarding | Y | Y | — | linked | linked | — |
| business_profile.privileged_settings.manage | Y | — | — | — | — | — |
| audit.read | Y | — | — | — | — | — |

Combine legitimate allow paths for the exact capability, not whole roles. Profile management/selectability derives from Organization owner/admin or owner/admin of a Group already attached to that profile; no BusinessProfile ownership record exists. The edit flag affects inherited ordinary updates, not read or selection eligibility. In Organization-only context, a future "My Business Profiles" list returns distinct manageable profiles in the active Organization through any qualifying attached Group or Organization role. Member-only profile reading does not make a profile manageable/selectable. In selected-Group context, inherited access remains restricted to that exact Group; the Organization-only management list does not bypass this restriction. No public list/UI is implemented in G007.

The creator is `group_owner`. For the bounded ordinary capabilities above, owner and admin have the same Group paths. `group.manage_members` is not blanket authority to promote oneself/admins to owner or transfer ownership: Group admin cannot grant/change/remove a Group-owner binding through a future administration facade. Public ownership transfer, last-owner and owner-specific destructive lifecycle rules remain separately reviewed G019 contracts; G007 adds no transfer action or workflow. Unknown future actions deny.

No Group path confers privileged settings, Organization policy or audit authority. Billing/payment/tax changes are not ordinary profile edits; their eventual workflows require separately reviewed Organization authority and do not acquire a hidden BusinessProfile grant store. Profile ownership-transfer capability is removed because the profile has no independent role owner. G006 legal-name context remains ordinary data, not a legal-settings workflow. No new billing, audit or onboarding side effects are implemented merely to exercise these evaluable contracts.

### R4 — One Organization Policy Flag

Support only `group_admins_edit_business_profiles`, a boolean per Organization, default true as in the archetype; the existing key applies to both Group owners and admins. Core reads it from a reviewed minimal persistent representation owned by that Organization; it cannot be overridden by request data, environment role claims or another Organization's value. The Plan pins default materialization/missing-row semantics, input validation, migration/grants and internal read/write contract. Default true is an explicit configured default, never a fallback after a store failure or malformed policy.

Only current org_owner authority may update it through the guarded internal facade. Update another tenant's flag, org_admin/member/Group-role attempts and malformed values fail without changing any row. Committed false/true changes immediately affect subsequent inherited decisions, without mutating roles/links. No Organization MFA policy, arbitrary policy map/DSL, rule editor or additional preferences. No public policy endpoint. Any new tenant-owned policy row belongs in the G008 RLS handoff.

### R5 — Enforced Internal Operations, Not Advisory Checks

Provide a bounded core facade for Organization read, Group read/update (accepted name/slug operations), BusinessProfile profile read/update and policy read/update. Policy read uses organization.read; policy update uses organization.manage_policy. Group updates retain the accepted name/slug contract; profile updates retain the accepted bounded context-field contract. The Plan pins exact methods, target/payload contracts and their matrix mapping before approval. Every call resolves/revalidates trusted identity/context and checks its exact target/action; optional client-provided capability checks or an exported allow result reused as authority are insufficient.

Scoped query/write predicates always include Organization ownership. Protected mutations authorize and mutate within one reviewed database unit; denied/unavailable/failed transactions change no entity, binding, link or policy state and return no success. Queries must not disclose foreign rows in lists, returned errors or logs. Unknown and inaccessible targets use the same non-disclosing external-facing error category; the Plan pins bounded internal reason codes.

Existing raw model primitives may remain for explicitly trusted bootstrap/tests and future integration, but cannot be reachable from untrusted production callers as a bypass. Identify and test the production import/wiring boundary. No remotely selectable test identity, startup bootstrap, public provisioning or new public tenant endpoint. Core owns this logic; web/intelligence contain no policy copy or database access. Authenticated future adapters are not replaced by a test-only production shortcut.

### R6 — Current Authority, Revocation and Concurrency

Resolve current facts on each operation. No durable/cache/session snapshot of effective roles or inherited capabilities may substitute for evaluation. Previously resolved context, old allow decisions or an unchanged login session cannot preserve authority after committed membership removal/downgrade, link removal, policy change, user disable, session revocation or expiry. Restoring one fact does not revive a revoked session or synthesize a missing Organization membership.

An operation starting after a relevant revocation/change commits must observe it and deny if its allow path is gone. In-flight protected writes need a declared authorization linearization point and transaction/locking contract: the write must be ordered before a conflicting revocation or observe the revoked authority and deny; an unchecked authorization-to-write gap is not allowed. The Plan must show how this composes with the actual accepted G004–G006 writers, including policy defaults and link changes. Use deterministic synchronization barriers to test both orderings, not timing sleeps. No new distributed invalidation framework is a prerequisite; escalate if this needs broad predecessor changes.

An immutable context or prior `SessionInfo`/allow value is not a reusable authority snapshot. Define how the guarded mutation revalidates session/user, Organization membership, the exact role/attachment allow path and policy within its declared ordering boundary. Existing logout/disable, role-change/removal and association-unlink writers must conflict correctly without requiring callers to adopt a new optional lock convention. Shared-method/signature or exact-boundary-test changes necessary for this guarantee are narrowly reviewable Plan scope, not permission to weaken predecessor checks or expose raw primitives publicly. If safe ordering needs a broader redesign, stop at planning and propose the concrete gap/split.

### R7 — Optional MFA and Honest Assurance

Inherit ADR-003's optional enrollment decision. Ordinary authorized operations cannot require authenticator enrollment merely because the user is a client, consultant, agency user or owner/admin. Two otherwise identical eligible users with/without MFA enrollment or with unknown assurance have identical ordinary capabilities. Social sign-in, enrolled factor, provider Required setting or role never upgrades assurance or grants tenant authority.

Preserve the accepted G004 denial for a trusted request/test profile explicitly requiring verified MFA without reviewed session-specific proof, and its challenge cancellation behavior. G007 adds no product operation requiring MFA, local factor/step-up UI, passkey/SMS integration or WorkOS setting change. Required-MFA staging capability qualification remains separate from eventual optional production onboarding.

### R8 — Persistence, Privileges and Later Isolation

RoleBinding remains the single writable role authority; projections stay read-only and inherited capabilities are computed, not mirrored. Preserve UUID canonical identity, same-Organization composite constraints, accepted migrations and session/model data. Use the accepted migrator/NOLOGIN-owner/runtime separation for any minimal policy migration; no SECURITY DEFINER/owner/DDL/role-assumption shortcut or extra web/intelligence credentials.

Test migration from populated predecessor data, empty full migration, repeat status/up, restart and down/reapply only on owned disposable databases if a schema addition is used. Down removes only G007 additions and restores prior behavior without changing predecessor data. If no migration is needed, demonstrate that against the exact accepted catalog.

G007 proves application policy, not database-wide RLS or storage isolation. Record all tenant-owned policy/model relations and query/context contracts G008 must protect; do not enable RLS, introduce pooled SET context or declare the full G010 attack suite complete here. No public tenant data surface before its owning reviewed isolation/integration gates.

### R9 — Privacy, Existing Validation and Provenance

Use accepted structured correlation/redaction. Record bounded action/scope/outcome/reason codes; no session tokens, provider claims, factor information, email, names, raw membership/profile/policy payloads or driver row details in logs/evidence. Retain the predecessor's UUID-attribution restrictions. Successful/denied/error paths carry canaries, and the accepted planted-leak detector still fails for an intentional test leak.

Policy/model tests use synthetic accepted global-user/session fixtures and real PostgreSQL with the actual runtime role; mocks alone are insufficient. They need no new live provider account or MFA enrollment. Preserve applicable accepted auth/session, browser/appearance, contracts, variants, migration, dependency and secret controls; choose scenarios from changed/shared boundaries. Reuse the accepted wrapper/harness/CI profiles with the smallest path/registry changes. No runner migration, new workflow framework, unneeded browser UX or repeated unchanged-control qualification.

Use the current solo-operator contract: fetch/rebase/reconcile before every local validation batch; required local full evidence before opening the MR; automatic MR fast, optional/manual MR gate; applicable automatic staging gate after separately authorized conditional integration. The owner's terminal-success-or-canceled staging disposition applies only with no known failing required check; canceled evidence is never a passing test result. Main promotion and final ACCEPT retain the remaining applicable gates. No second hosted full pass or live WorkOS campaign is required solely for G007. Compose a local forced policy-test failure with accepted CI failure controls to prove the gate cannot pass. Missing mandatory local evidence is unmet, not a successful skip.

Preserve exact baseline/input/scope hashes, pre/post drift/status, failed/interrupted runs, cycle history and ownership-based cleanup. Use the integrated Docker Desktop mount safeguard before local batches where applicable. Reuse existing cleanup/recovery/storage reporting if integrated; do not copy or implement G051's pending amendment here. Follow the glab MR-to-staging/source-deletion workflow, refreshing remote ancestry again before MR/integration. Required manual feedback, if the Plan identifies any, precedes final qualification; no new browser workflow is specified by this Goal.

## Acceptance Criteria and Verification Strategy

| ID | Required outcome | Objective verification/evidence |
|---|---|---|
| AC1 | R1 authenticated tenant resolution is server-owned. | Real accepted-session positives plus absent/expired/revoked/disabled/forged subject/context negatives; header/provider/model assertions cannot change subject or grants. Empty/default/foreign IDs and caller-constructed contexts fail for pinned reasons. |
| AC2 | Organization membership and selected Group gates hold. | Current bindings for multiple users/two Organizations/three Groups; Group bindings without Organization membership deny. Org admin versus Group-role selection, foreign Group and mismatched Group-target cases match R1/R2; no membership row is minted. |
| AC3 | Every R3 matrix cell and scope modifier is implemented. | Table-driven positive/negative cases for all 11 capabilities × 6 roles, with required Organization gating held constant, and own/all/linked scope variants. Unknown action/role/source/kind and store errors deny. Explicit independent expected results; no oracle generated from evaluator output. |
| AC4 | Profile capabilities derive solely from current Organization/Group roles and attachments. | Shared profile linked to two Groups, distinct profile for a third; owner/admin allow listed linked profile paths, member read only, and no Group role allows privileged settings/policy/audit. No BE-role input is accepted. In active A, distinct manageable profiles exclude B and member-only profiles; selected Group restricts inheritance and profile access reveals no sibling membership. Selection cannot use the new unattached Group to authorize an existing unrelated profile. |
| AC5 | The policy flag is scoped, owner-writable and effective. | Default/materialized/missing/error semantics, both values and malformed values; owner-only updates, cross-tenant/other-role denials, restart persistence. Toggling affects Group-inherited update only, with Organization paths and binding/link row sets unchanged. |
| AC6 | Guarded operations enforce rather than merely advertise policy. | Invoke each facade method on real runtime-role fixtures; unauthorized operations and definitive precommit/aborted failures return pinned safe errors with unchanged persisted state. Separately inject a lost commit reply: return unavailable, inspect persisted state to reconcile the authorized outcome, and do not claim rollback or silently retry. Inspect exact method mappings, production call paths and bypass absence. |
| AC7 | Committed role/link changes immediately remove only their derived profile paths. | Reuse context/session/allow after owner/admin removal or downgrade, unlink, policy false, disable/revocation/expiry. Losing the last management path denies edits/selection and removes management-list eligibility without a profile-row update; another qualifying attached Group or Organization path still allows in the correct context. Member-only may retain read, not management. No role/permission synchronization occurs. |
| AC8 | Authorization-to-write races follow the pinned ordering. | Deterministic two-connection barriers for each relied-on mutable allow source and G004 lifecycle state. Both orderings produce either authorized-before-change commit or denial/rollback after change; no unauthorized row mutation. Result identifies linearization and accepted-writer compatibility. |
| AC9 | Optional MFA creates no ordinary role-based enrollment gate. | Equivalent valid sessions/users with enrolled/unenrolled/unknown assurance yield equal ordinary decisions across all roles. Accepted verified-MFA-required negative and challenged-cancellation controls remain failing closed; no assurance inferred from Google/roles/provider policy. |
| AC10 | Persistence/grants and predecessor data survive the increment. | Real runtime-role policy CRUD and negative ownership/DDL/PUBLIC/projection-write controls; populated/empty/restart/disposable down/reapply or verified unchanged catalog. No copied grants/capability store; accepted migrations and predecessor data stay unchanged. |
| AC11 | Future boundaries remain absent and G008 receives an exact inventory. | Route/MCP/schema/import/source comparisons: no public tenancy/switching/onboarding/billing/audit/storage/RLS/product work. Tenant relations/query predicates and trusted context contract are listed for G008; raw SQL isolation is not claimed. |
| AC12 | Privacy and inherited regressions retain their guarantees. | Actual allow/deny/failure logs and reports are scanned with canaries; planted-leak control fails as intended. Applicable auth/session/browser/theme/scoped-token/contracts/variant/dependency checks pass on the final tree without weakening guards. |
| AC13 | Existing container/CI gates validate the final policy implementation. | Documented commands run verbatim on real root and hash-identical disposable copies; final wrapper full suite actually passes on the rebased candidate. Bind MR automatic fast and applicable post-integration staging disposition to the integrated content under R9/current owner policy; canceled is not test success and any known required failure blocks acceptance. Local deliberate policy-test failure plus inherited gate controls prove fail-closed behavior; no required skips. |
| AC14 | Review/evidence/budget and completion gates are satisfied. | Exact predecessor baseline, inputs/scope/evidence, pre/post wrapper results, NUL-safe status, owned cleanup and failed/interrupted runs; B7/cycle-4 accounting, Result and actual-diff Codex ACCEPT. No G008/G010/G021 completion claim. |

## Expected Artifacts and Required Validation

Core resolver/evaluator/guarded facade, the minimum policy persistence/migration when needed, containerized unit/integration/concurrency tests and narrowly scoped harness/CI wiring. Focused `docs/foundation/AUTHORIZATION.md` must state the exact matrix, policy/context semantics, operation mappings, ordering, error/privacy contract and G008 handoff. Claude creates the canonical G007 Plan and Result only in their owning stages; this draft creates neither.

The Plan maps AC1–AC14 to named cases, exact action/state/error outcomes and evidence paths; tests are built/run during approved execution. Use two Organizations, at least three Groups and shared/unlinked BusinessProfiles, multiple users with differing scoped/Organization/Group grants, and accepted session fixtures. Assertions inspect exact returned data, state row sets and intended denials, not arbitrary failure or presence of tests. Resource ownership and destructive isolation use accepted controls.

No application dependency is specified by this Spec. Claude enumerates any proposed library/version/category and justified internal interface changes for approval; a major policy vendor or generic engine is outside scope. Final immutable inputs include applicable predecessor acceptance/deviation records and current canonical MFA/CI directions. Do not transplant active G004's snapshot or mutate it.

## Permitted Deviations

Module names, typed private context representation, query/locking strategy and minimal policy persistence are Plan choices if the above guarantees and accepted predecessor contracts hold. Pin method/action/error/default/transaction contracts before approval. No deviation changes matrix cells, broadens Group inheritance, bypasses Organization membership, infers MFA, copies grants, adds public routes, weakens tests or defers a G007 criterion to G008. Substantive predecessor incompatibility or a policy change returns to Codex/owner.

## Risks and Decisions Before Implementation

1. G004–G006 acceptance/integration is satisfied. Reconcile the new committed execution baseline and exact accepted session/model/grant contracts; old design-only observations and fixed migration/check counts are historical, not current execution assumptions.
2. Product owner reviews the concrete matrix and context interpretation (R1–R3). They preserve the archetype but distinguish Organization gating, selected-Group inheritance and derived BusinessProfile access; any contradictory product decision must be resolved before Plan approval.
3. Claude must pin policy storage/defaults and transactional authorization ordering with the existing lifecycle writers. If a guaranteed race-free facade needs broad predecessor changes, report the narrow gap/split, not a framework or weakened promise.
4. Sensitive capability names are evaluation contracts only. No real transfer, billing, policy editor, grant-administration or audit operation can be silently introduced to demonstrate a matrix cell.
5. B7 feasibility and required hosted access must be assessed during planning. No new provider qualification or process campaign is needed solely for policy fixtures. SECURITY.md's empty state and unrelated accepted residuals do not justify expanding this Goal.

## Completion and Gate Conditions

Ready for Claude Plan preparation against accepted G006 and transitive predecessors. No Claude dispatch or implementation is performed by this reassessment. A scoped Claude Plan receives Codex review of exact context/matrix/operation/default/ordering/migration contracts, budget and fresh committed binding; unresolved owner decisions remain explicit. All ACs must pass with Codex implementation ACCEPT and required integration disposition before G008 execution. G021 remains the product-feature barrier.

## Revision Record

- 2026-10-07 — r1 drafted from current architecture, optional-MFA owner decision, archetype policy contract and provisional G004–G006 designs/reviews. Goal/Spec/Progress/roadmap links only; no production code, Plan/Result, dependencies, builds/tests, provider/CI actions, commit or push.

## Inherited owner clarification — 2026-10-08

Accepted [ADR-004](../../../adr/ADR-004-group-business-profile-cardinality.md) and [the cardinality/visibility addendum](G006-business-profile-cardinality-addendum.md) govern predecessor reconciliation and relevant existing acceptance cases. One completed Group has one Business Profile; multiple Groups may share it. A global user can manage multiple profiles, but only active-Organization profiles with a legitimate policy allow path are visible. Verify one user with grants in two Organizations cannot carry profile visibility or authority across the active-tenant boundary; shared profiles expose no sibling-Group links/memberships. BusinessEntity is profile/context, not a third tenancy tier or a required identifier on all Group resources. Existing capability matrix, budgets, criterion counts, predecessor gates and G021 remain. No new UI/public listing, onboarding workflow or implementation is authorized by this clarification.

## Owner authority revision — 2026-10-08

r2 replaces independent BusinessEntity authority with Organization/Group-derived access and adds Group owner to the ordinary matrix. The r1 Spec (not included) retains historical scope. Matrix is now eleven capabilities by six roles; acceptance-criterion count remains fourteen. Inherited clarification above is superseded where it references independent profile grants or the former matrix. Current role/attachment data, active tenant and edit policy govern access with no synchronized BE ownership. B7/checkpoint four, predecessor acceptance and G021 remain; no implementation authorization.

## Owner naming clarification — 2026-10-08

Use canonical `BusinessProfile` (UI: **Business Profile**) instead of `BusinessEntity`; the shared association becomes `GroupBusinessProfile`, future identifier spelling `business_profile_id`, and capability prefix `business_profile.*`. UI placement is **Group settings → Business Profile**; `GroupProfile` is not the reusable record name. Earlier names in archived/detailed planning records are historical. Stable Goal IDs, management filenames/links, UUID identity and all cardinality/derived-role/active-tenant guarantees remain. Claude must reconcile affected provisional Plan/schema/interface/fixture pins before execution; this does not authorize edits to closed artifacts, source Skills or active G004 inputs. Criteria, budgets, checkpoints and G021 are unchanged.

## Accepted-baseline reassessment — 2026-10-10

r3 preserves the exact r2 Spec (not included) and
refreshes inspected state, the relocated Skill reference and validation policy.
See Codex reassessment (not included).
Authorization guarantees and all matrix cells remain unchanged. AC6 distinguishes
definitive aborts from the accepted predecessor's lost-commit-reply ambiguity;
neither permits an unauthorized mutation. Fourteen criteria
and B7/checkpoint four remain; no implementation, tests, hosted action or budget
cycle is consumed. R9/AC13 now explicitly apply already-established owner CI
policy rather than demand a duplicate hosted qualification.

ADR-004's later UX clarification supersedes the exclusive Group-settings
placement above: Groups and Business Profiles are sibling management areas,
sharing one authorized profile page. G018 owns that presentation and onboarding;
G007 has no UI or route implementation. The historical Skill's direct-grant
matrix cannot override canonical Organization/Group-only authority.

## Execution blocker clarification — 2026-10-10

The bounded Codex review (not included) authorizes exactly scoped direct RoleBinding locking reads while preserving snapshot projections, grants, fact predicates and ordering. It disposes the observed small-fixture E-PG-1 trigger without accepting production performance or weakening safety. The exact shared scanner correction may be carried independently of G051 completion with explicit scope/input reconciliation. Fourteen ACs and all required validation remain. The linked Goal explicitly adopts goal-level attempt accounting without rewriting historical Plan/snapshot bytes.
