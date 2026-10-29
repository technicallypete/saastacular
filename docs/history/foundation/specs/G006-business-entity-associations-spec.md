> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G006 — BusinessProfile and Group associations work specification

Goal: [G006-business-entity-associations.md](../goals/G006-business-entity-associations.md)
Epic: EPIC-002 — Identity and tenant foundation
State: r3 — owner-approved derived profile authority; G005 acceptance and revised Claude Plan approval required

## Goal, Rationale and Authority

Establish reusable BusinessProfile context and one-profile-per-Group links within one Organization. Use current Organization/Group RoleBindings and profile attachments to derive access facts; create no separate BusinessProfile grants, owner/admin rows or writable authority projection. Preserve future 1:N BillingProfile ownership without commerce schema or features.

Follow canonical governance, BUDGET_POLICY, roadmap/progress, accepted ADRs and ARCHITECTURE's **Canonical Multi-Tenant Semantics (v0.1 Foundation Override)** and **Revised v0.1 Epic Sequence**. These titles disambiguate repeated section numbers. The accepted archetype tenancy contract (not included) supplies the shared-authority/reuse contract. Its historical permission for multiple profiles per Group is superseded by accepted ADR-004. Architecture and owner decisions prevail; Claude chooses the concrete implementation in its Plan.

## Inspected State and Prerequisites

Drafting observes main HEAD `19f945af7b7bfd1b27cb395fd56b0c2cec3b163b`. G001/G002 are accepted; G047/G003 execute separately and remain unaccepted; G004 and G005 are management drafts only. The main implementation has no production BusinessProfile/association schema. Fixtures in the accepted Skill are contract demonstrations, not this Goal's runtime evidence.

Require G005 ACCEPT and reassess its actual user/Organization/Group/binding representation, constraints, projections, migration/grant contracts, Result/deviations and CI before Claude planning. G004's identity trust boundary and G003/G047 operational/visual acceptance are transitive prerequisites. A fresh accepted baseline and exact reviewed-input binding are required at approval; pending batch inputs and this observed HEAD are not an execution binding.

Inherit G001's owner/migrator/runtime separation, actual-root Docker commands, disposable destructive copies and core-only datastore access; G002/G047's browser/appearance/first-paint/scoped tokens; G003's logging/CI/contracts/evidence controls; G004's global UUID/session mapping; and G005's tenant ownership, Group slugs, the sole Organization/Group binding store and immediate projections. Preserve recorded residuals/deviations without unrelated repair. No new provider/vendor is selected here.

## Scope and Explicit Non-goals

In scope: minimal BusinessProfile context, same-Organization single-profile Group links, read-only derived profile-access facts, internal core operations, migrations/runtime privileges, SQL/concurrency/failure tests and focused docs/provenance/CI.

Out of scope: capability evaluation/inheritance/request TenantContext (G007), RLS (G008), storage/isolation completion (G009/G010), durable audit (G011), MCP (G014/G015), domains/branding (G016/G017), onboarding/member administration (G018/G019), billing/payment/legal-operation schemas, AI context workflows, public profile UI/API, business merge/search catalog and product features. No production seed or new general metadata/authorization framework.

## Requirements

### R1 — Organization-owned Canonical BusinessProfile

BusinessProfile has an immutable nonzero UUID and non-null `organization_id` referencing an existing Organization. Organization is its tenant/security owner; Group is not its parent. No required `group_id`, single-Group ownership column or per-Group copied profile row. Tenant ownership and UUID cannot be changed by supported operations or direct runtime writes.

Canonical references use UUIDs. Business names, website, legal name, labels and future hostname/slug are attributes, not identity or authority. The same display/legal name may exist on different BusinessProfiles or in different Organizations; do not merge/link them by name, email, website or provider claim. A shared legal business in two Organizations is represented by distinct tenant-owned records, not a cross-tenant reference.

### R2 — Minimal Bounded Business Context

Provide a required display name (1–255 characters, not all whitespace), and optional legal name (up to 255), description (up to 4,000), industry (up to 255) and website (up to 2,048). Text is valid Unicode; the Plan pins normalization, character-count, null/empty and whitespace behavior before approval. Reject invalid/over-limit inputs without partial updates. Context fields are untrusted inert data, never SQL/HTML/script or logging templates.

If present, website is an absolute HTTP/HTTPS URL under a reviewed parser/validation contract; unsafe schemes or credentials in the URL are rejected. Saving/reading it never fetches, verifies, renders remote content or claims domain ownership. Legal name is passive profile context; it is not a tax/legal-settings workflow or a privileged ownership action. No arbitrary executable metadata, files, payment secrets or AI invocation.

An update changes the one BusinessProfile context row. All Groups linked to it read that same updated context while unrelated BusinessProfiles/Organizations remain unchanged. Profile updates never change its UUID, owning Organization, links or role bindings. Additional addresses/contact/typed metadata are not mandatory in this bounded increment; adding a context subsystem requires a separately reviewed scope change.

### R3 — One Profile per Group, Reusable within its Organization

GroupBusinessProfile is a durable UUID-identified association carrying non-null `organization_id`, `group_id` and `business_profile_id`. Both endpoints exist in that same Organization. Enforce same-Organization membership with **composite foreign keys** to Group and BusinessProfile and service validation before writes; service-only checks or independent UUID foreign keys without tenant alignment are insufficient.

Enforce at most one BusinessProfile association for each Group through a database uniqueness constraint and service validation. Many Groups may share one BusinessProfile; do not impose uniqueness on `business_profile_id` alone. Linking another Group to an existing profile reuses that BusinessProfile UUID/context row. A duplicate/retried same-profile association is either an idempotent return or a reviewed conflict, never a second row or new profile. Attempting to attach a different profile to an already linked Group fails without replacing its current link. Concurrent attempts with different profiles persist at most one link for the Group; the losing operation returns the pinned conflict. No profile-replacement workflow is authorized in G006.

Reject unknown/mismatched/foreign Organization, Group or BusinessProfile IDs in the service and through direct SQL as the actual runtime role. Attempts to retarget an existing link across Organizations or reparent either endpoint must also fail. Unlink removes only that link, preserving the BusinessProfile, other Group links and memberships. It cannot cascade into profile or Group deletion. Referenced parents/endpoints cannot be deleted into dangling links or memberships. No cross-Organization transfer/cascade lifecycle workflow is introduced.

### R4 — Derived Profile Access, No Independent BusinessProfile Roles

RoleBinding supports Organization and Group authority only. Do not add a BusinessProfile scope, `be_owner`/`be_admin` roles, direct-grant source, profile owner column or independently writable BusinessProfileMembership. A creator marker, if later introduced as audit metadata, cannot confer access.

Expose read-only current facts joining the existing Group memberships to each Group's single profile attachment within its Organization. The Plan pins the projection/query representation, columns and runtime read-only privileges; no role or capability is copied onto the profile/association. G006 supplies these relational facts, not authenticated policy decisions or a public list. G007 applies Organization membership, Group-owner/admin/member capability rules, selected scope and the policy flag; G008 enforces the database boundary.

A global user can manage multiple profiles through ownership/administration of any Group attached to each profile, or qualifying Organization authority. Losing an owner/admin role removes that management path immediately after commit, with no profile-role update. Another qualifying attached Group can independently preserve access. Downgrade to Group member may retain ordinary profile read but cannot preserve management through the removed role. Unlink similarly removes only that path. Group/Organization role changes touch only their authoritative bindings, not BusinessProfile rows or copied access records.

A profile attachment or derived access fact never grants Organization membership or access to sibling Groups. Profiles without attachments have no Group-derived access; any access is evaluated through Organization authority. Future "My Business Profiles" means distinct manageable profiles in the active tenant, not every profile visible for reading or a global directory.

### R5 — Profile Context, Onboarding and Future Billing

BusinessProfile is reusable profile/context, not another tenancy/container level. Group-owned resources retain Organization/Group ownership without a mandatory BusinessProfile foreign key or three-part scope. The same global user can manage profiles in multiple Organizations, but future lists/selection expose only profiles in the active Organization for which current policy grants access.

Preserve future BusinessProfile 1:N BillingProfile ownership: no singleton billing field, payment/tax/provider schema, duplicated profile or uniqueness limiting one profile to one Group. Future BillingProfile children use canonical BusinessProfile UUID and Organization ownership. G007 pins Organization-only sensitive-operation policy; Group ownership/administration alone cannot imply billing/payment/tax privileges.

G018 owns the client-facing flow: create the Group with its creator as Group owner, then create a new profile or select an existing manageable profile in that Organization. Selection is independently authorized through a pre-existing attached Group owner/admin path or Organization authority; the new Group's owner role alone does not authorize an unrelated profile. Onboarding completes only with exactly one attachment. New-profile creation and attachment commit together; failure leaves no orphan profile from that attempt and the Group remains unfinished. G006 has internal model primitives only, not an onboarding state/flag or public creation/selection workflow. A foundation Group may temporarily have no link. Replacement/removal of an onboarded Group's profile remains a separately reviewed lifecycle, not an implicit G006 operation.

### R6 — Internal Core Contracts and Transaction Behavior

Core owns internal create/read/update BusinessProfile, link/read/unlink association and read-only derived access-fact primitives. The Plan pins inputs, bounded fields, explicit Organization predicates, deterministic conflict/not-found/error categories and transaction/retry semantics. Do not add HTTP/BFF/MCP/profile-admin/bootstrap endpoints, a public provisioning CLI, auth/provider changes or browser datastore access.

Service operations validate both association endpoints and owning Organization before a write attempt. Unknown and foreign selectors have the same non-disclosing failure category; they do not produce a business name, role or existence oracle. Access-fact reads preserve global subject and Organization/Group scope integrity and cannot mint a membership. Service prechecks complement the race-proof database constraints and cannot be replaced by catching every database failure as success/conflict.

Rejected inputs, constraint errors, injected store/transaction failures and concurrency outcomes leave the documented prior state unchanged or a complete committed operation, with no partial profile/link. Acknowledged success is durable. Concurrent duplicate links yield one valid record; concurrent profile edits follow the Plan's reviewed conflict or last-committed contract. Do not claim stronger concurrency semantics without tests. Rollbacks do not corrupt accepted Group slugs or other bindings.

### R7 — Additive Migrations and Least Privilege

Extend accepted G005 migration history without editing predecessor migrations. Use the accepted migrator/NOLOGIN-owner path, actual runtime CRUD grants and read-only projection privileges. If composite keys require additional parent uniqueness/indexes, justify exactly those additive changes in the Plan; parent UUID/slug/role behavior stays unchanged. Runtime cannot own objects, perform DDL, assume privileged roles, read migration metadata or use a broad security-definer bypass. No web/intelligence database credentials.

Test full empty-database migration, populated post-G005 upgrade, repeated up/status, real runtime model CRUD, restart persistence and disposable down/reapply. Down restores the predecessor schema and preserves global identity/session, Organization/Group and their bindings/slugs; only G006 fixture data/scope additions may be removed, explicitly documented. No live/developer rollback/reset. Inspect the actual constraint catalog and role attributes/grants; owner/superuser tests cannot substitute for runtime negatives.

### R8 — Non-vacuous Tests, Redaction and Regressions

Use two Organizations, multiple Groups, at least two BusinessProfiles in one Organization and one in the other, multiple global users with Organization and Group roles. Demonstrate two Groups sharing one BusinessProfile, a third Group using another profile, rejection of a second profile for an already linked Group (including concurrent distinct-profile attempts), one shared-context update, unlink preserving the other link/context/membership, independent Group-owner/admin/member changes and immediate derived-access updates and all foreign/unknown-ID cases. Assert exact row sets/counts and intended service/SQL constraint outcomes rather than any failure or a mock-only pass.

Synthetic profile/auth canaries never enter actual correlated model logs, error bodies or retained evidence; structured bounded outcome/trace codes remain observable. The accepted planted-secret detector must still catch a deliberate test leak. No raw business profile, provider claims, credential, private contact data or complete grant payload is recorded. G011 later adds the durable transactional audit ledger; no audit scope is pulled forward.

Retain inherited global identity/session, Organization/Group/binding/slug, health/boundary/port, browser appearance/first-paint/hydration/token, optional-intelligence, dependency/provenance and wrapper drift/status controls. Reconcile exact new model paths/constraints into current guards and consumer customizations; do not broadly remove old empty-schema/role/scope checks. Shared-harness changes rerun applicable static/control suites in the same cycle.

### R9 — Actual Repository, CI and Evidence Binding

Run documented container commands against the submitted root and actual PostgreSQL, including model/migration/concurrency tests under the runtime role. Destructive tests use hash-matched owned disposable copies/databases. Add gating G006 model/migration checks to the accepted GitLab pipeline; retain identity/provider CI safeguards and accepted regression profiles. New model fixtures need no new provider/vendor/account.

Required evidence includes a separately authorized hosted run on the final implementation with commit/tree/job IDs and required results. Unavailable access or required evidence remains unmet; a local pass or Skill fixture is not hosted/runtime proof. Prior Goals' named hosted-action grants do not authorize new actions here.

Keep Skills/source kits/original manifests unchanged; record exact consumer customizations and applicable intelligence-disabled additions. Reuse approved input binding, exact scope allow-list, NUL-safe status, pre/post verification, safe external evidence and positively owned cleanup. Final full validation and evidence must cover the actual final tree after all model/harness changes. Do not rebuild unchanged approval/execution controls.

## Expected Artifacts and Architectural Constraints

- Core BusinessProfile/association/internal access-fact modules, read-only derived profile-access facts and minimal additive migrations/grants; no new web/public/MCP contract.
- Real-runtime SQL/composite-FK/concurrency/transaction/migration tests and exact container/harness/CI integration, plus an extension of G005's `docs/foundation/TENANCY.md` documenting context, link, derived access and future billing dimensions.
- Updated schema/constraint inventory covering tenant ownership, parent composite keys, link cardinality/uniqueness, absence of independent profile grants, projection privileges and future policy/RLS owners.
- Claude-owned `management/v0/plans/G006-business-entity-associations-plan.md` before implementation and `management/v0/results/G006-business-entity-associations-result.md` with scope/evidence afterward. Neither is created here.
- Sanitized evidence outside the repository; no production fixtures, new provider, billing/product schema or compliance claim. Named areas are not broad write permission; the reviewed Plan lists exact files/dependencies.

## Acceptance Criteria and Verification Strategy

| ID | Required outcome | Required verification/evidence |
|---|---|---|
| AC1 | R1 BusinessProfile is UUID-identified and owned by one Organization independently of Group. | Schema/catalog plus real-runtime create/read/rename tests; UUID/parent mutation, nonexistent parent and zero-ID negatives. Equal business names remain separate across/within Organizations; no Group parent or global business identity is substituted. |
| AC2 | R2 bounded context is inert, validated and shared rather than copied. | Boundary/null/Unicode/whitespace/unsafe-URL tests for each pinned field; invalid updates preserve row state. Two linked Groups observe one committed context update by the same UUID, while unrelated profiles/links/grants stay unchanged. No URL fetch or HTML/SQL execution occurs. |
| AC3 | R3 association cardinality and uniqueness permit actual reuse. | Link two Groups to one BusinessProfile and a third Group to a different profile; inspect shared profile identity and exact association UUID/counts. A second profile for an already linked Group is rejected by service and runtime SQL, leaving the original link unchanged. Duplicate/retry and concurrent distinct-profile attempts persist at most one link for the Group with the intended conflict; the database constraint permits multiple Groups for one profile. |
| AC4 | R3 same-Organization links fail closed at service and database layers. | Service prevalidation test proves foreign/unknown endpoints are rejected before writes; direct SQL via the actual runtime role rejects all Organization/Group/BusinessProfile mismatch combinations. Composite-FK catalog, intended error/constraint and before/after rows prove both defenses. Retarget/reparent negatives also fail. |
| AC5 | R3/R6 unlink and rejected operations cannot remove unrelated context/authority. | Unlink one shared link and retain the other link, context, Groups and membership. Failed link/profile/grant transactions leave no partial rows; referenced-parent deletion cannot produce dangling links/bindings. No cross-tenant cascade/transfer is introduced. |
| AC6 | R4 profile-access facts derive from current Group roles and attachments without a second authority store. | Inspect exact projection/query and read-only grants; one global user owns/administers Groups for two profiles with another membership in a different tenant. Grant/change/remove/downgrade and unlink immediately change only applicable facts. Profile rows remain unchanged; another qualifying attached Group preserves its path. Projection writes fail; no BusinessProfile RoleBinding, owner column or independent role row exists. |
| AC7 | R4/R5 association and Group-role changes need no profile-authority synchronization. | Snapshot profile and unrelated binding rows around each mutation. Removing the last qualifying owner/admin path removes management facts; removing one of two paths preserves the other; member-only differs from management. No copied capability store, automatic Organization role or sibling-Group authority appears. |
| AC8 | R5/R6 no public/onboarding/billing/policy/product surface appears. | Compare/probe accepted route/contract inventory and inspect wiring/schema. No profile/grant/provisioning endpoint, TenantContext/policy/RLS/onboarding flag, BillingProfile/payment schema or product feature. Document future 1:N child relationship and G007/G018 owners without restricting one profile to one Group or adding early implementation. |
| AC9 | R7 migration/runtime/persistence evidence preserves the predecessor. | Empty/full and populated-G005 up/status, runtime CRUD/grants, restart, disposable down/reapply; exact predecessor identity/session/Group/slugs/bindings survive. Inspect ownership/composite keys/projection grants and runtime DDL/role/metadata negatives; accepted migrations are byte-identical. |
| AC10 | R8 inherited checks and privacy remain effective. | Accepted G005 model and global-auth/browser/boundary/variant/provenance/dependency regressions pass. Actual safe correlated logs exclude seeded sensitive context in success/failure paths; planted-leak negative is detected. Shared-check changes rerun their static/control coverage with no relaxed reason codes or broad exclusions. |
| AC11 | R9 actual container commands and hosted gates validate the final model. | Execute documented commands verbatim on actual root and hash-matched copies; final approved-wrapper full suite, G006 runtime tests and disabled variant pass with no required skips. Authorized final hosted job binds commit/tree and model/migration results; forced model-test failure fails its gate. |
| AC12 | Provenance, Result and completion gates are complete within B5. | Exact baseline/input/scope/evidence hashes, pre/post drift/status, cleanup inventory, failure/deviation history and B5/cycle-3 checkpoint accompany AC1–AC11. Codex reviews actual diff/evidence and returns ACCEPT; no inheritance/RLS/onboarding/billing/product completion is claimed. |

## Required Validation and Evidence Handling

Before approval the Plan maps every criterion to named scenarios, intended service/SQL outcomes, state assertions and artifacts. Build/run tests during approved execution using the accepted root/migration/variant/static/provenance/browser and CI controls. Unchanged wrappers need references and required regressions, not another validation implementation or pre-approval probe series.

Bind the actual accepted post-G005 commit, reviewed Goal/Spec/Plan, exact scope/constraint/internal contract inventories and final tree. Record actual runtime connection identity/attributes, catalog/SQL constraints, versions/grants, safe commands, totals/exit codes, NUL-safe status and start/end owned resources. Retain failures/aborted runs and distinguish fixture setup, runtime SQL, actual stack and hosted results. No secrets or business context in source/artifacts/evidence; hashes/IDs must remain sufficient to attribute the run.

All required link integrity, shared-record, derived-access, migration and CI evidence belongs in G006. G007's capability tests add authorization/inheritance guarantees; they are not a fallback for an invalid G006 association/grant. Preserve accepted G004 provider regression arrangements without procuring new accounts solely for synthetic model fixtures.

## Permitted Deviations

The reviewed Plan chooses module/table naming, profile normalization/optional-value behavior, exact internal errors, idempotent-link versus duplicate-conflict semantics, concurrency contract and additive index/constraint layout. It must satisfy the pinned field bounds, same-Organization composite foreign keys, UUID stability, single authority/projection rules and accepted predecessor behavior.

No deviation permits a Group-owned/one-Group-only/cross-tenant BusinessProfile or multiple simultaneous profiles for one Group, copied role store, derived-access integrity bypass, public profile/onboarding surface, premature policy/RLS/billing or deferred mandatory evidence. New lifecycle/transfer/merge semantics or major dependencies require a separate reviewed scope/owner decision.

## Risks and Decisions Before Implementation

1. **Accepted predecessor representation.** G005 is a draft; inspect its accepted typed-scope/projection/schema design before planning. Extend it once, preserving Organization/Group rules and no unconstrained polymorphic path. A needed redesign beyond this increment returns to Codex.
2. **Profile contract pins.** The Plan fixes exact Unicode/count/normalization/null semantics, URL parsing/no-fetch guarantee, errors and update/concurrency behavior within R2/R6. No tax/legal-state machine, arbitrary metadata or hidden per-Group profile copy.
3. **Constraint and rollback risk.** Review additive composite parent keys, association/source-role constraints, referential deletion behavior and removal of only G006 profile/link/projection additions on disposable down. Populated G005 data must survive; cross-tenant moves/cascades and ownership transfer are not authorized workarounds.
4. **Future authorization boundary.** Direct grants and links are facts only; G007 determines effective permissions, G008 RLS and G018 authorized selection/onboarding. If implementation needs those policy decisions to expose data, stop rather than exposing a partially protected service.
5. **Budget and hosted access.** B5 is conditional on reuse of accepted G005 controls and no policy/UI expansion. If full required evidence cannot fit, recommend a bounded revision/split before execution. Missing hosted permission/access blocks completion; do not silently skip it or introduce a new runner framework.

## Completion and Gate Conditions

This draft is ready for owner review only. G005 ACCEPT, actual-tree reassessment, resolved Plan pins, separate Codex Plan/input/budget approval and all AC1–AC12 evidence precede implementation acceptance. Only Codex ACCEPT releases the next G007 handoff. No G007 Spec is created here; the foundation/product barrier through G021 remains closed.

## Revision Record

- 2026-10-05 — Initial G006 definition with 12 measurable criteria and B5/checkpoint 3. Drafted with G005 at the owner's request; no Plan, implementation, dependency/build/migration work, commit or push.


## Parallel provisional planning — 2026-10-07

Codex authorizes drafting a provisional G006 Plan in an isolated clone while G004/G005 remain unimplemented/unaccepted. This supersedes only the requirement to wait for G005 ACCEPT before planning. Use G004 Goal/Spec r3 and design-accepted Plan r4, and G005 Goal/Spec and provisional design-accepted Plan r2 with its Codex review. Label exact schema/module/migration/grant/harness assumptions provisional. Reconcile the actual accepted G005 Result and predecessor state before final approval and execution binding.

Planning only at xhigh, optionally one isolated Claude planning subagent. No implementation, installations, builds, Docker/provider/CI actions, active G048 checkout/input/resource edits, commit or push. G005 ACCEPT/integration remains an implementation prerequisite; B5/checkpoint three, all ACs, UUID/shared RoleBinding/same-Organization invariants and G021 remain. Return substantive ambiguities to Codex rather than invent accepted facts or expand scope.

## Revision Record — 2026-10-08

r2 records the owner-approved single-profile cardinality in ADR-004 and [the scope addendum](G006-business-profile-cardinality-addendum.md). The original r1 Spec (not included) is preserved. R3/R5/R8 and AC3/AC7 now require one profile at most per Group, shared reuse, distinct-profile conflict/concurrency coverage and multiple scoped direct grants for a global user. Exactly one on completed onboarding remains G018; no onboarding state, client creation UI, third tenancy tier or cross-tenant profile is introduced in G006. AC count remains twelve; B5, predecessor gates and G021 remain. Provisional Plan r1 needs bounded reconciliation; no implementation is authorized.

## Owner authority revision — 2026-10-08

r3 supersedes r2's independent BusinessEntity grants and r1's many-to-many cardinality. Preserved r2 (not included) and r1 (not included) retain history. Access derives from current Organization/Group roles and attachments; Group creator is owner, with no second membership record to synchronize. R4/R5 and AC6/AC7 now require derived-fact and role-removal coverage. Twelve criteria and B5/checkpoint three remain; predecessor acceptance and a bounded Claude Plan amendment are required. No implementation or active G004 input change.

## Owner naming clarification — 2026-10-08

Use canonical `BusinessProfile` (UI: **Business Profile**) instead of `BusinessEntity`; the shared association becomes `GroupBusinessProfile`, future identifier spelling `business_profile_id`, and capability prefix `business_profile.*`. UI placement is **Group settings → Business Profile**; `GroupProfile` is not the reusable record name. Earlier names in archived/detailed planning records are historical. Stable Goal IDs, management filenames/links, UUID identity and all cardinality/derived-role/active-tenant guarantees remain. Claude must reconcile affected provisional Plan/schema/interface/fixture pins before execution; this does not authorize edits to closed artifacts, source Skills or active G004 inputs. Criteria, budgets, checkpoints and G021 are unchanged.
