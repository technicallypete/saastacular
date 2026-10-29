> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G005 — Organization and Group membership model work specification

Goal: [G005-organization-group-memberships.md](../goals/G005-organization-group-memberships.md)
Epic: EPIC-002 — Identity and tenant foundation
State: r2 — owner-approved Group creator/owner role; provisional Plan amendment and G004 acceptance required

## Goal, Rationale and Authority

Implement the first Organization/Group tenancy schema and scoped membership model using the accepted global identity. One RoleBinding store supplies canonical membership projections. Different scoped roles for the same user remain independent; Group names/slugs are alternate identifiers and never authority.

Follow AGENTS.md, CLAUDE.md, BUDGET_POLICY, the roadmap and the architecture sections titled **Canonical Multi-Tenant Semantics (v0.1 Foundation Override)** and **Revised v0.1 Epic Sequence**. Section numbers repeat in ARCHITECTURE; these titles identify the applicable override. The accepted archetype's tenancy contract (not included) clarifies RoleBinding/projection semantics. Architecture and accepted owner decisions prevail over examples. Claude's Plan chooses storage/module details without changing these guarantees.

## Inspected State and Prerequisites

Drafting observes main HEAD `19f945af7b7bfd1b27cb395fd56b0c2cec3b163b`. G001/G002 are accepted. G047/G003 execute in a separate checkout; G004 has draft [Goal](../goals/G004-global-identity-sessions.md)/[Spec](G004-global-identity-sessions-spec.md) only. No production identity/Organization/Group/RoleBinding model exists in the inspected main tree. Skill fixtures demonstrate the intended contract but are not the production implementation.

G004 ACCEPT must supply the actual global UUID identity/session model and accepted post-G003/G047 operations. Before Claude planning, read its Result, deviations, migrations, grants, route inventory and validation commands. At G005 approval bind the actual accepted baseline and current reviewed inputs; do not use this planning observation or an active batch snapshot as execution authority.

Inherit G001's migrator/NOLOGIN-owner separation, core-only datastore access, configurable identity/ports, actual-root Docker commands and disposable destructive tests; G002/G047's appearance/first-paint/accessibility/scoped-token guarantees; G003's accepted contracts, structured redaction, CI and evidence controls; and G004's global UUID/session trust boundary. Accepted residuals are preserved, not silently fixed or declared resolved. No provider choice or infrastructure change is made here.

## Scope and Explicit Non-goals

In scope: minimal Organizations/Groups; Group slugs; scoped RoleBinding persistence and membership projections; internal core model operations; migrations/grants; real PostgreSQL/concurrency tests; exact consumer provenance, docs and applicable CI gates.

Out of scope: BusinessProfile/associations/derived profile-access facts (G006); request TenantContext/policy/inheritance (G007); RLS (G008); storage and attack-suite completion (G009/G010); durable audit (G011); public MCP (G014/G015); domains/branding (G016/G017); onboarding/invitations/member administration (G018/G019); account/provider changes and all product features. No web screen, public tenancy endpoint or domain seed/demo is required or authorized.

## Requirements

### R1 — Canonical Entities and Ownership

Organization is the single persisted tenant/security-boundary entity. Group belongs to exactly one Organization and has a non-null durable `organization_id`. Use immutable, nonzero UUID canonical identifiers for Organization, Group and RoleBinding; reference the accepted G004 global user UUID. Names, slugs, provider subjects and hostnames cannot become canonical keys. No persisted Tenant, Agency or ClientOrganization replacement model.

Persist only bounded display identity and structural metadata needed by this Goal. Group name and slug are distinct attributes; Organization display names need not be globally unique. No Organization-hostname routing slug/domain table is introduced. Renaming an entity or editing a Group slug preserves its UUID, owning Organization, memberships and references. Moving a Group or a binding to another Organization is outside scope and rejected by supported operations and database protections, not treated as a rename.

### R2 — Typed, Referentially Valid RoleBinding

RoleBinding retains `subject`, `scope_type`, `scope_id`, `role`, `source` semantics and carries its owning `organization_id`. Its subject is an existing global user, not an email/provider subject/group name. Supported G005 scopes are Organization and Group only. An Organization binding's scope UUID equals its `organization_id`; a Group binding refers to an existing Group in that Organization. Nonexistent/foreign/type-mismatched scopes, zero IDs and invalid subjects cannot become bindings.

Database enforcement must prove subject existence, scope existence/type and same-Organization ownership, including direct writes by the actual runtime role and concurrent writes. A polymorphic UUID with only service validation or an unchecked generic scope string is insufficient. The Plan chooses the constraint representation and documents how every supported scope is constrained; it must not create a duplicate canonical tenant entity or an unconstrained future scope path.

Use the accepted vocabulary: Organization roles `org_owner`, `org_admin`, `org_member`; Group roles `group_owner`, `group_admin`, `group_member`; source `membership`. Unsupported scope/role/source combinations fail closed in service and database validation. BusinessProfile is not a RoleBinding scope: G006 derives profile-access facts from Group memberships and attachments. Independent BusinessProfile grants/roles are forbidden. Invitation/system grant sources have later owners. These identifiers are internal semantic codes, not hardcoded client-facing product role labels. No role granted at one scope becomes a global role.

### R3 — One Writable Authority and Immediate Projections

RoleBinding is the only writable role store. Expose `OrganizationMembership(user_id, organization_id, role)` and `GroupMembership(user_id, group_id, role)` as read-only projections of current bindings. No role on the user, Organization, Group, mutable membership mirror, separate ACL or copied permission cache can contradict it. If projections are database views, runtime grants must make them read-only; any service projection has no independent role-writing path.

At most one membership binding exists per subject and scope, with database uniqueness under concurrency. A role change updates that binding rather than inserting another role or synchronizing copies. Preserve binding identity across role changes. Removal deletes/revokes the single authority record as specified by the reviewed storage contract and immediately removes the membership projection; regrant cannot resurrect an obsolete role. No read projection may return stale authority after a committed change.

An Organization role does not synthesize Group membership; a Group role does not synthesize an Organization role. Removing/changing one binding does not change other scoped bindings, the global UUID or application sessions. G007 determines effective access from current scoped facts, including consequences of absent Organization membership; G005 must not invent a global administrator, last-owner rule, inherited capability or scope-switching policy. No Organization prerequisite is silently implemented as copied Group grants.

### R4 — Suggested and Editable Group Slugs

Stored slugs are lowercase ASCII kebab-case, 1–63 characters, matching `[a-z0-9]+(?:-[a-z0-9]+)*`. Reject empty, overlong, noncanonical or reserved stored values. On creation provide a deterministic name-derived suggestion: `Acme Revenue` yields `acme-revenue`. The reviewed Plan pins punctuation/Unicode/empty-result behavior and availability/collision responses. An unsluggable name requires a valid explicit choice or a documented bounded suggestion rule, not an invalid stored slug.

An explicit valid slug can replace the suggestion. Editing a slug changes no UUID, role, foreign key or owning Organization; a name edit alone does not silently rewrite an already stored slug. Enforce `UNIQUE (organization_id, slug)` in PostgreSQL. The same slug in two Organizations is permitted; two Groups in one Organization cannot share it. An availability suggestion is advisory: concurrent create/edit must still use the unique constraint and report a deterministic conflict without overwriting a Group.

Reserve at least `api`, `auth`, `account`, `login`, `logout`, `health`, `healthz`, `readyz`, `mcp`, `sse`, `rpc`, `system`, plus valid-slug top-level application/system route segments in the accepted starting tree. The Plan inventories and pins the exact set; tests cover every reserved entry. Encoded separators, dots, traversal, leading/trailing/repeated hyphens and non-ASCII text cannot bypass the stored grammar. Suggestion normalization applies the same reserved/uniqueness checks. Slug/name lookup always includes an explicit owning Organization; neither a slug nor an availability result grants access.

### R5 — Internal Operations and Failure Atomicity

Core owns internal Organization/Group creation, UUID-based reads, display-name/Group-slug edits and membership grant/read/change/remove operations. The Plan pins exact internal inputs, bounded fields, conflict/not-found/error categories, transaction boundaries and retry behavior. These are model primitives for trusted in-process foundation use/tests, not authenticated-user administration claims. Group creation accepts an explicit existing global creator UUID supplied by the trusted caller and atomically creates that Group plus one `group_owner` membership binding for the creator. Failure in either write leaves neither row. It does not mint Organization membership, grant provider-derived authority or add a second owner column on Group; later authenticated onboarding supplies the creator from its verified context, never a browser-asserted identity. Existing internal grant/change/remove primitives support owner/admin/member without implementing public ownership transfer or last-owner policy.

No HTTP, BFF, MCP, startup env hook, public CLI or remotely selectable test/provisioning route can invoke these writes. Tests may seed synthetic accepted-user records through isolated fixtures; no development authenticator or automatic provider-claim/hostname membership bootstrap is introduced. Unknown/foreign selector outcomes reveal no cross-Organization details. Explicit Organization predicates apply to scope-sensitive reads and updates, even though full request policy and RLS arrive later.

Validate related identifiers and slug/role constraints before mutation; database constraints remain race-proof defense. A rejected mutation or injected transaction/store failure commits no partial entity/binding/projection change and cannot acknowledge success. Concurrent duplicate grants cannot create two membership rows; concurrent role changes leave one valid binding/projection, with reviewed conflict or last-committed semantics. Do not claim optimistic concurrency or lost-update protection without implementing/testing the chosen contract.

No cross-Organization reparenting or cascade removal of identities/scopes is authorized. Database references prevent dangling bindings/Groups if a referenced subject/parent is deleted directly. Public deletion, ownership transfer and last-owner lifecycle policy remain later work; they are not prerequisites to this model Goal.

### R6 — Scope Boundaries for Later Goals

Group rows may temporarily have no BusinessProfile association because G006 has not run. They must not be presented as completed onboarding; G018 owns the create-or-select workflow and its completion invariant. Do not add an onboarding-complete flag, invitation state machine, Organization policy/preferences/branding or a BillingProfile placeholder.

G005 must not claim end-user cross-tenant authorization or RLS isolation. Its guarantee is tenant-anchored referential/model integrity plus absence of a public model surface. G007/G008 remain mandatory before public tenant data operations. Record which model queries and tables they will secure, without implementing their policy or Specs. MCP remains dormant and the foundation barrier remains closed through G021.

### R7 — Migrations and Actual Runtime Privileges

Add reviewed forward migrations after the accepted G004 history; never edit accepted baseline/identity migrations. Use the existing migrator-to-NOLOGIN-owner path. The actual runtime role has only necessary model CRUD/projection reads and no ownership, DDL, role assumption, migration-metadata access or security-definer bypass. Keep table/function/sequence ownership and PUBLIC privileges consistent with the accepted role model. No web/intelligence datastore credentials.

Test empty-database full migration, populated post-G004 upgrade, repeated up/status, restart persistence and down/reapply on owned disposable databases. Down must restore the predecessor schema without changing accepted identity/session data; it may destroy only G005 fixture data, explicitly documented. Validate real runtime CRUD and negative SQL constraints/grants, not only owner/test-superuser operations. No production/developer rollback or reset is authorized.

### R8 — Tests, Privacy and Inherited Controls

Use multiple synthetic users, two Organizations, at least two Groups in one Organization and a Group in the other. Include different roles for one global UUID, equal Group slugs in different Organizations, unknown/foreign IDs, duplicate/concurrent writes, committed role change/removal and injected failure. Inspect persisted rows/projections and intended constraint/error names, not merely nonzero exits or existence of tests.

Use accepted structured correlation/redaction without logging raw display names, authentication material, provider claims or complete membership payloads. Bounded event/error codes and reviewed UUID attribution may remain. Canary inputs exercise success/failure logging; the accepted planted-leak detector must still detect a deliberate test leak. No durable audit ledger is implemented; G011 owns transactional AuditEvent integration.

Preserve accepted health/auth/session contracts, appearance/first paint/hydration/scoped tokens, browser/service boundaries, intelligence-disabled operation, configurable identity/ports, dependency pins and wrapper reason-code/status/drift semantics. Exact new model additions may reconcile prior empty-schema/provenance guards; old checks cannot be removed or broadly excluded to accommodate tenant schema. Shared harness edits rerun affected controls/static checks in the same cycle.

### R9 — Container, CI and Provenance Evidence

Model/migration/concurrency checks run through documented container commands against actual PostgreSQL and the submitted root implementation. Destructive and collision/outage copies must be byte-identical to that tree and positively owned. SQLite, mocks, archetype fixtures and an owner connection cannot substitute for required real-runtime evidence.

Integrate these checks into the accepted gating GitLab pipeline. Record at least one authorized hosted run on the final implementation with its commit/tree, job identities and required model/migration results; missing runner access is unmet evidence, not an allowed skip. Retain accepted identity/provider CI secret protections and regression profiles; model fixtures require no additional live provider/vendor. Earlier G003 hosted authorization does not grant new hosted actions for this Goal.

Keep both Skills/source kits and original `.archetype` manifests unchanged. Register exact consumer customizations and applicable disabled-variant additions. At approval, bind the actual accepted predecessor commit and Goal/Spec/Plan inputs using accepted controls; final pre/post checks, exact path allow-list, source hashes and cleanup record prove which implementation was validated. No new approval-wrapper framework.

## Expected Artifacts and Architectural Constraints

- Core Organization/Group/RoleBinding modules, read-only membership projections and reviewed model migrations/grants; no handler, browser or MCP contract additions.
- Containerized model/SQL/concurrency/migration tests and narrowly reviewed harness/CI integration, plus focused `docs/foundation/TENANCY.md` and exact consumer customization records.
- A schema/constraint inventory naming global identity references, immutable IDs, tenant-owned tables, scope validity, role/source checks, uniqueness, projection grants and later RLS coverage. No duplicate tenant or authority store.
- Claude-owned `management/v0/plans/G005-organization-group-memberships-plan.md` before execution and `management/v0/results/G005-organization-group-memberships-result.md` with exact scope/evidence afterward. Neither is created by this drafting task.
- Sanitized evidence outside the repository; no synthetic data shipped as production seed, credentials, product schema, new vendor/service or SOC 2 compliance claim. Artifact areas are not blanket write permissions; the Plan enumerates paths.

## Acceptance Criteria and Verification Strategy

| ID | Required outcome | Required verification/evidence |
|---|---|---|
| AC1 | R1 canonical Organization/Group identity and ownership are implemented. | Catalog/schema and internal CRUD tests show nonzero UUIDs, non-null Group Organization reference and no duplicate tenant model. Rename/slug edits preserve UUID/parent/bindings; service and runtime-SQL parent/identity mutation controls enforce the reviewed immutability contract. |
| AC2 | R2 bindings reference valid global users and correctly typed same-Organization scopes. | Positive Organization/Group grants via actual runtime connection; direct SQL and service negatives for nonexistent user/scope, scope-type mismatch, foreign Organization/Group and zero IDs. Constraint catalog and post-failure rows prove no invalid binding persisted. |
| AC3 | R2/R3 only supported role/source combinations and one membership per subject/scope exist. | Table-driven valid/invalid role/scope/source cases, including valid `group_owner` creation/grant/change cases and premature BusinessProfile/invitation sources; concurrent duplicate-grant attempts yield exactly one binding. Unsupported combinations fail for their intended rule and cannot use an unconstrained future branch. |
| AC4 | R3 membership projections are read-only and follow one binding immediately. | Grant/change/remove/regrant sequence checks exact authority-row count and both membership projection shapes. Attempt projection role writes; inspect schema/grants and service API for a second writable role store. Committed updates/removal change no copied membership or unrelated binding. |
| AC5 | The same accepted global user has independent roles in multiple scopes. | One real persisted G004 user UUID holds different roles in two Organizations and two Groups; change/remove one and assert other roles, global mapping and session rows remain unchanged. No provider-claim grant, global admin or implicit membership is created. |
| AC6 | R4 Group suggestion/edit/uniqueness/reserved rules are complete. | `Acme Revenue` suggestion, exact reviewed normalization/length matrix and every reserved name; same slug across Organizations succeeds, same-Organization create/edit collides. Concurrent create/edit cannot overwrite rows. UUID/reference checks survive rename and explicit slug edit. |
| AC7 | R5 internal operations, including creator ownership, are scoped and transactionally fail closed. | Equal-slug/foreign UUID reads and writes use explicit Organization predicates; rejected/input/store/transaction failures preserve before/after row sets. Concurrent role changes leave one consistent projection with documented semantics. Negative assertions identify the intended error rather than any failure.  Creator creation yields exactly one Group and its `group_owner` binding; unknown creator and injected failure between the two writes leave neither. No Organization role is created. |
| AC8 | R5/R6 no public tenancy surface or future-goal implementation is introduced. | Compare accepted core/web/MCP route and contract inventories; probe reviewed candidate tenancy/provisioning routes for absence and inspect server wiring. Schema/source diff has no BusinessProfile, policy/RLS, onboarding, routing/branding, audit or product implementation; no externally writable model bootstrap. |
| AC9 | R7 migrations, persistence and least privilege work on real PostgreSQL. | Empty/full and populated-G004 up/status, runtime CRUD, restart, disposable down/reapply; identity/session preservation on upgrade/down. Catalog ownership/projection grants and runtime DDL/role/metadata/invalid-reference controls pass. Accepted migrations remain byte-identical. |
| AC10 | R8 inherited contracts and redaction remain effective. | Accepted auth/session/browser/first-paint/token/boundary/variant regressions pass; actual success/failure model logs contain safe correlated codes and no planted sensitive values. Planted-leak detection still fails as intended; dependency/protected/provenance controls retain their semantics. |
| AC11 | R9 documented container and hosted CI checks gate the submitted model. | Commands run verbatim against the actual root and hash-matched copies; final approved-wrapper full suite plus G005 tests and disabled variant pass with no required skips. Final hosted job binds tree/commit and model/migration results; a forced model-test failure cannot produce a passing gating job. |
| AC12 | Provenance, bounded Result and all completion gates are satisfied. | Exact baseline/input/scope/evidence hashes, pre/post drift/status checks, start/end owned-resource inventory, failures/deviations and B7/cycle-4 checkpoint accompany AC1–AC11 evidence. Codex reviews actual diff and returns ACCEPT; no policy/RLS/onboarding/product completion is asserted. |

## Required Validation and Evidence Handling

The Plan maps every criterion to named test cases, intended constraint/error outcomes, before/after state assertions and evidence artifacts. Build these tests during approved execution; do not require a second pre-approval demonstration of unchanged wrappers or construct a generalized validation framework. Reuse current root/migrations/variants/browser/provenance/static and hosted CI controls.

Final evidence binds the accepted post-G004 baseline, reviewed inputs, schema/internal contract inventory, exact new dependencies if any and submitted implementation. Record NUL-safe tracked/untracked/ignored status, pre/post checks, actual runtime role identity/attributes, migration versions/catalog/grants, test totals/exit codes and positively owned cleanup. Keep failed/aborted results and cycle accounting; no evidence-directory loss or cache reuse is presented as a fresh uncached success.

New model tests require no provider account creation or live MFA demonstration solely to create database fixtures. Preserve the accepted G004 regression arrangement and secrets safeguards. Required final model/runtime/hosted checks cannot move to G007/G008; those Goals own additional policy/RLS guarantees, not remediation of invalid G005 rows.

## Permitted Deviations

The Plan may choose module/table names, projection representation, UUID generation strategy, typed-scope constraint implementation, internal conflict/retry semantics and deterministic slug suggestion details within R1–R5. It must pin these before execution and prove the guarantees, including actual runtime-SQL integrity and read-only projections. Optional lifecycle tables are not needed; adding one requires a demonstrated scoped requirement and review, with no second role column.

No deviation permits a writable membership mirror, raw polymorphic reference without database integrity, global/implicit grants, public provisioning, cross-Organization move, premature BusinessProfile/policy/RLS or skipped mandatory evidence. Alternative tenant hierarchy, ownership policy or canonical authority semantics need owner resolution, not an implementation workaround.

## Risks and Decisions Before Implementation

1. **Predecessor model is not yet accepted.** Reconcile exact user keys, migrations, sessions, grants and CI from G004's actual accepted Result. If earlier drafts conflict, return the difference to Codex before implementation; do not rewrite accepted history.
2. **Typed-scope persistence and projections.** The Plan must show how PostgreSQL enforces existing typed scopes and ownership, how projection writes are denied and how one-row membership changes remain atomic. An unconstrained generic foreign ID is a substantive defect.
3. **Role vocabulary versus lifecycle policy.** Semantic roles/source are pinned above. Bootstrap ownership, minimum/last-owner rules, membership invitations and the policy consequences of losing Organization membership are not silently selected here; G007/G018/G019 retain those decisions. If the model requires such a decision to exist, stop with the concrete conflict rather than expanding scope.
4. **Slug and migration pins.** Before approval enumerate reserved routes, exact suggestion normalization/collision behavior, bounded display fields, error categories and rollback effects. No unreviewed cascade or parent transfer; required race tests must run under the actual runtime role.
5. **Budget/access.** B7 includes all model/constraint/migration/hosted evidence. Missing runner permissions blocks required evidence. Additional UI, policy, RLS, provider work or a new process framework requires a separate scope decision rather than use of remaining cycles.

## Completion and Gate Conditions

This draft is ready for owner review, not implementation. G004 must be accepted and the actual tree reassessed; Claude then prepares a bounded Plan; Codex approves exact scope/input binding, migrations/internal contracts and budget feasibility. Current evidence must satisfy AC1–AC12 and receive Codex ACCEPT. Only then is G006 dependency-ready; G007/G008 and the G021 foundation barrier remain mandatory.

## Revision Record

- 2026-10-05 — Initial G005 specification with 12 measurable criteria, B7/checkpoint 4 and no public tenancy surface. Drafted with G006 at the owner's request; no implementation, Plan, test execution, commit or push.


## Owner-authorized parallel planning — 2026-10-07

The owner authorizes a provisional G005 Plan while G004 is not implemented/accepted. This supersedes only the earlier requirement to wait for G004 ACCEPT before planning. Read G004 Goal/Spec r3, accepted Plan r4 and its Codex review as provisional integration inputs; label proposed table/module/migration dependencies as assumptions, not accepted implementation facts. Reconcile against the actual accepted G004 Result, schema, grants and operational evidence before final Plan approval and execution binding.

Planning only, xhigh, in an isolated clone outside the active G048 checkout. No application/schema/test implementation, dependency installation, builds, Docker/provider/CI changes, commit or push. G004 ACCEPT/integration still gates G005 implementation. Scope, ACs, B7/checkpoint four, Codex Plan approval, exact execution binding and G021 remain unchanged. If provisional dependencies prevent a concrete decision, name it as an implementation prerequisite rather than expanding scope or inventing accepted facts.

## Owner revision — 2026-10-08

r2 adds `group_owner` and atomic creator ownership under ADR-004. The original r1 Spec (not included) and Plan r2/review remain historical; Claude must amend the role checks, creation contract and tests before execution. G006 adds no BusinessEntity role scope. AC count twelve, B7/checkpoint four, G004 dependency and G021 remain. No implementation is authorized.

## Owner naming clarification — 2026-10-08

Use canonical `BusinessProfile` (UI: **Business Profile**) instead of `BusinessEntity`; the shared association becomes `GroupBusinessProfile`, future identifier spelling `business_profile_id`, and capability prefix `business_profile.*`. UI placement is **Group settings → Business Profile**; `GroupProfile` is not the reusable record name. Earlier names in archived/detailed planning records are historical. Stable Goal IDs, management filenames/links, UUID identity and all cardinality/derived-role/active-tenant guarantees remain. Claude must reconcile affected provisional Plan/schema/interface/fixture pins before execution; this does not authorize edits to closed artifacts, source Skills or active G004 inputs. Criteria, budgets, checkpoints and G021 are unchanged.
