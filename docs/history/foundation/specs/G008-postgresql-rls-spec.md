> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G008 — PostgreSQL RLS and least-privilege roles work specification

Goal: [G008-postgresql-rls.md](../goals/G008-postgresql-rls.md)
Epic: EPIC-003 — Authorization and isolation
State: r3 — Accepted-predecessor reconciliation complete; specification only, Claude Plan ACCEPT gates implementation

## Goal, Rationale and Authority

Add database enforcement beneath G007's core authorization so an omitted application query predicate does not expose another Organization or an unauthorized Group. Protect existing facts without creating new domain workflows. Context must be trusted, current and local to the database unit that uses it; pooled connections cannot carry one request's authority into another.

Follow ARCHITECTURE's canonical tenancy/foundation sections, accepted ADRs, AGENTS/CLAUDE, BUDGET_POLICY, ROADMAP/PROGRESS and the accepted [archetype RLS contract](../../../../skills/saas-foundation/references/database-rls.md) / [tenancy contract](../../../../skills/saas-foundation/references/tenancy-authorization.md). They require global identity protection as well as tenant-table RLS. SECURITY.md is currently empty; its presence is not proof and filling it is outside this Goal.

## Inspected State, Dependencies and Inherited Constraints

Inspection baseline: clean pushed staging `b6306e626759769d570ff07e49e0454bf93d9f25`.
This is a source-inspection reference, not an execution snapshot or a fresh live
catalog test. G004–G007 are accepted and integrated. [G007 final ACCEPT](../results/G007-implementation-codex-review.md#final-implementation-acceptance--2026-10-11--accept)
closes its fourteen criteria; its qualified staging source and historical failed/
canceled evidence remain in that record. G054 is independent and held for its
later live test; it is not a G008 dependency.

Accepted migration history is `core/migrations/00001_baseline.sql` through
`00005_organization_policy.sql` (version 5). No accepted application relation
currently has RLS or a SECURITY DEFINER policy reader. The source inventory is:

| Class | Existing `app` objects | G008 obligation |
|---|---|---|
| Tenant tables | `organizations`, `groups`, `role_bindings`, `business_profiles`, `group_business_profiles` | ENABLE/FORCE; pinned per-command and column permissions, including Organization-root protection. |
| Global identity | `users` | ENABLE/FORCE; self/shared-permitted-scope visibility and bounded pre-tenant identity paths. |
| Global authentication infrastructure | `identity_provider_binding`, `identity_principals`, `sessions`, `login_transactions` | Exact bounded authentication/privilege/privacy contract; no invented tenant ownership or blanket exemption. |
| Invoker projections | `organization_memberships`, `group_memberships`, `business_profile_access_paths` | Preserve SELECT-only `security_invoker` and the underlying scope guarantees. |
| Existing trigger functions | `organizations_keys_immutable`, `groups_keys_immutable`, `role_bindings_keys_immutable`, `business_profiles_keys_immutable`, `group_business_profiles_keys_immutable` | Preserve effective key immutability, exact ownership/grants and fixed paths. |
| System metadata | `app.goose_db_version` and its identity sequence | Remain owner-only; exact identity exclusion from tenant policies. |

The G007 flag is `organizations.group_admins_edit_business_profiles`, NOT NULL
DEFAULT true; there is no separate policy table or fallback after a read error.
This inventory anchors planning; the Plan must verify the actual catalog and
all functions/sequences/grants rather than treating this table as a complete ACL
specification. Add new migrations after version 5 without changing its bytes.

Accepted core boundaries are `identity.Service.Session` for opaque-credential
validation and `core/internal/tenancy/access` for `BeginRequest`, `Resolve` and
the guarded facade. The request scope and TenantContext are opaque bindings,
not cached authority. Each operation rechecks user/session/current scope facts.
G007 exposes eight guarded operations: Organization read/policy read/policy set,
Group read/name/slug updates, and BusinessProfile read/context update. The eleven
capability decisions do not authorize adding the missing public workflows.
Predecessor model/setup functions remain internal; their runtime grants are not
proof that untrusted callers may use them. G008 must inventory and reconcile each
existing trusted model/authentication/test path without turning it into a new
request administration policy or leaving a raw production bypass.

Read the actual [G004](../results/G004-implementation-codex-review.md),
[G005](../results/G005-implementation-codex-review.md),
[G006](../results/G006-implementation-codex-review.md) and G007 Results/reviews/
deviations, the [authorization handoff](../../../foundation/AUTHORIZATION.md),
[TENANCY](../../../foundation/TENANCY.md) and [PG-001/PG-002](../PERFORMANCE.md).
Bind then-current committed inputs afresh at Plan/execution review. Historical
migration tests may assert the exact version-5 grants/catalog: keep their lower-
version contract and add final-version expectations, never edit old migration
bytes or weaken a detector to hide the new enforcement.

Inherit G001's migrator/NOLOGIN-owner/runtime separation, owner-only migration metadata, core-only datastore access, configurable naming/ports, actual-root commands and disposable destructive tests. Preserve accepted UI appearance/first paint/scoped tokens/accessibility, G003/ADR-002 CI and redaction, ADR-003 global identity/opaque sessions/unknown assurance and optional MFA, and G005–G007 UUID identity, composite ownership constraints, single RoleBinding authority, capability-specific inheritance and revocation ordering. Slugs/names/hosts remain selectors, never canonical identity or authority. Infrastructure/framework keys may retain their justified accepted strategy.

## Scope and Explicit Non-goals

In scope: enforce existing database facts and guarded operations, bounded context/helper integration, additive RLS/grant migrations, reproducible role provisioning without reliance on image init, meaningful real-runtime tests and evidence. Minimum indexes/constraints may support existing ownership/policy queries; justify them in the Plan.

No new product/domain tables, writable membership mirrors, copied effective grants, new capabilities, public tenant/provisioning routes, MCP tools, hostname switching, onboarding, invitations, billing, ownership-transfer workflow, object storage, durable audit, new factor/provider integration, cloud deployment, pooler, generic authorization framework or CI redesign. G009/G010 and later Goals retain their scope; G021 remains closed.

## Requirements

### R1 — Complete Relation and Operation Inventory

Before Plan approval enumerate every accepted application table, view, function and relevant sequence. Classify tables as tenant-owned, global identity, other global authentication infrastructure or system metadata, with exact reasons and access paths. Tenant inventory includes Organization roots (their UUID is the boundary), Groups, BusinessProfiles, GroupBusinessProfile links, RoleBinding and G007 policy storage if separate. Include every actual addition, not just these expected names. No naming-prefix or broad schema exemption.

Every tenant-owned table and global identity table has ENABLE and FORCE RLS. Each SELECT/INSERT/UPDATE/DELETE command has a pinned role/column/row permission contract, an explicit command-specific policy where permitted, or documented default-deny plus privilege denial where not permitted. No FOR ALL shortcut, unconditional runtime policy or accidentally permissive additional policy. Classify supported privileged operations separately; no runtime flag can select an administrative policy. Catalog assertions cover both policies and effective grants. PostgreSQL combines permissive policies with OR, so review their combined effect, not individual expressions only. [PostgreSQL row security](https://www.postgresql.org/docs/17/ddl-rowsecurity.html)

Global provider mappings, sessions and auth transaction infrastructure need exact bounded G004 access, not artificial Organization ownership. System migration metadata remains owner-only and excluded by exact identity. Exemption from tenant RLS is not exemption from privilege/privacy tests. New domain schema to make the inventory convenient is forbidden.

### R2 — Runtime Identity and Least Privilege

Core uses the configured runtime login directly. It is never superuser, BYPASSRLS, owner, CREATEROLE/CREATEDB or a member able to inherit/assume owner, migrator, retention or helper roles. It cannot change policies, disable RLS, create/replace application objects or read migration metadata. Web/intelligence receive no database credentials. Retention remains reserved with no new tenant access; G011 owns its future purpose.

Preserve migrator-to-NOLOGIN-owner execution and configuration-derived role names. Table/function/sequence grants and column grants expose only reviewed operations; PUBLIC receives no new access. Deny runtime TRUNCATE and other whole-table privileges that RLS cannot filter. Owner-path row queries must demonstrate FORCE with an authorized positive and foreign-scope denial; do not install an unconditional owner row policy merely to make the tests pass. Document legitimate migration DDL/provisioning separately from row-access proof. [PostgreSQL row security exclusions](https://www.postgresql.org/docs/17/ddl-rowsecurity.html)

Verify both session_user and current_user/role attributes on every isolation-test connection. A superuser connection followed by SET ROLE runtime is not runtime-login evidence. Privileged fixture setup and controls use separately identified owned connections and cannot accidentally become the assertion path.

### R3 — Trusted Context and Transaction Boundary

Use accepted core authentication and G007 resolution to establish user UUID, active Organization UUID and optional Group UUID. Validate current enabled user/session, expiry/revocation, Organization membership, selected Group and exact target/action. Caller/model/header/provider assertions, role lists, names/slugs and arbitrary constructed contexts cannot supply authority. No default Organization, tenant-zero/global sentinel or bypass flag.

All protected queries/mutations execute on the same transaction-bound connection after complete parameterized transaction-local context initialization. Use local settings such as `set_config(..., true)`, never session-level authority settings. Pin exact fields, parsers, initialization ordering and transaction API in the Plan. Explicitly clear an absent optional Group. Missing/empty required fields, zero UUIDs, malformed/inconsistent settings and partial initialization fail closed with no protected mutation or data returned; errors are not allow fallbacks. Calling local initialization outside a transaction is insufficient. [PostgreSQL configuration functions](https://www.postgresql.org/docs/17/functions-admin.html)

Tenant work cannot escape via pool-level queries, another connection or a raw predecessor method reachable from untrusted production callers. Preserve G007's guarded operation boundary and explicit Organization query predicates. A returned allow decision or previously resolved context cannot authorize a later request. Context/session revalidation and database policy resolution must compose without an unprotected circular bootstrap.

**Trust limit:** settings transport core-established identity; they do not authenticate a database client. A holder of the shared runtime credential with arbitrary SQL may impersonate a different valid user by changing settings. Do not claim protection against credential compromise/arbitrary valid-subject SQL impersonation from RLS alone. Direct-SQL negatives keep the trusted actor fixed while substituting tenant/scope/row IDs; server-boundary tests separately reject subject forgery. No new signing/proxy/DB-auth framework is authorized to overstate this guarantee. No exposed application path permits arbitrary SQL or caller-selected context settings.

### R4 — Organization, Group and BusinessProfile Enforcement

For fixed authenticated identity/context, raw runtime queries lacking application filters return only the active Organization's permitted rows. Reads, aggregates, joins, subqueries, RETURNING and writes cannot reveal/mutate foreign rows. INSERT checks new ownership; UPDATE checks both existing visibility and proposed row ownership/scope. Existing UUID/Organization immutability, typed bindings, same-Organization composite foreign keys and uniqueness remain enforced independently of RLS. Referential constraints can bypass RLS, so retain those constraints and translate driver details into accepted non-disclosing application errors. [PostgreSQL policy command semantics](https://www.postgresql.org/docs/17/sql-createpolicy.html), [constraint interaction](https://www.postgresql.org/docs/17/ddl-rowsecurity.html)

Group visibility/update follows G007 own/all rules and selected-Group restriction. BusinessProfile profile access follows its exact Organization/capability-specific linked-Group paths and owner-controlled edit flag; selected Group limits inherited paths while independent Organization authority remain usable as G007 specifies. A shared profile exposes neither sibling Groups nor their links/memberships. Association visibility requires both the exact Group's visibility and the profile's visibility; derived profile access alone cannot reveal an unrelated Group/link. Never flatten BusinessProfile into one Group parent.

Separate protected profile/name/slug/policy changes from authority/structural mutations. An otherwise allowed row update cannot change identity, parent, scope, role/source, subject or unrelated protected columns. The runtime request path exposes only the existing G007 facade; raw SQL cannot self-promote, mint Organization/Group memberships, retarget links or bootstrap a tenant through an unconditional request policy. Trusted predecessor fixture/model setup may use explicitly bounded administrative/test paths, not a caller-controlled mode switch. Pin every command's allowed/denied outcome; broader administration semantics require scope review rather than invented capabilities or new workflows.

### R5 — Authority Lookups, Projections and Helper Safety

RLS and G007 must agree for the protected operations; maintain independent expected results, not an oracle generated from the same SQL/helper implementation. RoleBinding remains the sole writable authority and inheritance is computed from current facts. Membership projections are SELECT-only `security_invoker` views; direct base-table, view and join access all enforce the current caller's permitted scope. No writable or materialized authority mirror. [PostgreSQL invoker views](https://www.postgresql.org/docs/17/sql-createview.html)

Pin minimum readable authority facts: a caller can inspect its own current bindings in the active Organization within the selected scope; broader membership visibility needs the accepted exact scope-management capability. Internal policy lookup may require additional facts but cannot export them as an unrestricted directory. Resolve recursive authority-policy queries using the accepted archetype pattern or a reviewed equivalent preserving its guarantees.

If SECURITY DEFINER helpers are necessary, enumerate exact signatures, callers, referenced relations and return shapes. Use a dedicated NOLOGIN/NOBYPASSRLS role, SELECT-only role-specific policies/grants, fixed trusted search_path and no runtime membership in that role. Helpers are bounded current-context predicates or narrowly defined authentication operations, not arbitrary SQL, table/column names, arbitrary-subject lookup or bulk credential/role export. Revoke PUBLIC EXECUTE; grant only exact necessary callers. No custom LEAKPROOF, owner/superuser/BYPASSRLS definer or broad dynamic-execution helper. Assert the exact allowlist, privileges and helper role's restricted access; helpers must not weaken FORCE. [PostgreSQL definer safety](https://www.postgresql.org/docs/17/sql-createfunction.html)

### R6 — Global Identity and Pre-tenant Authentication

Apply ENABLE/FORCE to the accepted global identity table without making a user tenant-owned or adding tenant IDs to identity. Ordinary tenant-context reads admit self and only identities sharing a scope the caller is permitted to inspect; no directory access from mere global login or a profile link. Pin the minimum columns and shared-scope predicate against G007's membership visibility before approval. Foreign identities and their auth/session/provider material remain undisclosed. No people-directory feature is introduced.

Current `core/internal/identity/store.go` uses runtime pool-level SQL for credential-hash session lookup/renewal, provider-principal issuance, one-use callback state and trusted internal enable/disable. G008 must protect these paths as well as tenant `access.inTx`; initializing only the latter is insufficient. Preserve all accepted G004 flows before any Organization exists/is selected: verified provider-principal lookup/create, session creation/authentication, account self-read, logout/revocation, local disable and pending callback state. Their trusted credential/provider checks establish identity; a supplied user UUID cannot select that bootstrap path. Define the exact pre-tenant database access and any minimum helper exception in the Plan; no runtime owner connection, all-user policy, email auto-link or remotely selectable synthetic identity. A co-member identity row never grants read/write access to another user's sessions, provider mappings, login transactions or assurance fields. Broad auth redesign returns to Codex.

MFA remains optional for all human roles; RLS adds no enrollment requirement, provider setting or inferred assurance. Existing explicit verified-MFA-required denial and opted-in challenge cancellation guarantees remain unchanged. Staging qualification remains G004's separate requirement.

### R7 — Connection Reuse, Failure and Cancellation

Context ends on commit/rollback and cannot remain effective when a connection is released. Every newly acquired request transaction initializes all relevant fields; reused settings or stale cached authority cannot substitute. Reuse the same observed physical backend with a pool constrained to one connection for: Organization A → context-free access → Organization B; Group-selected → Organization-only; successful commit, rollback, statement error, context-init error and cancellation. Test failed commit/connection loss where the actual driver permits deterministic injection; uncertain/unusable connections must be discarded, never reused with unknown state.

Distinguish safe API behavior from PostgreSQL semantics: local settings revert to a prior session value if one was already present. Prove the production boundary rejects/clears unexpected prior session-level authority or discards the connection before protected work; local SET alone is not cleanup proof. Include a deliberately unsafe session-level setting control in an owned disposable connection that demonstrably leaks, and show the actual boundary prevents that leak. No sleep-based identity assumptions; identify backend reuse without leaking credentials. [PostgreSQL local-setting behavior](https://www.postgresql.org/docs/17/sql-set.html)

### R8 — Current Facts and Authorization Ordering

The accepted implementation uses REPEATABLE READ READ ONLY snapshots for guarded
reads and READ COMMITTED READ WRITE for guarded writes. Writes lock user, session,
Organization, subject memberships and relevant Groups in the declared order;
policy updates use Organization FOR NO KEY UPDATE. A changed slug may upgrade
the Group lock to FOR UPDATE. Preserve the same-statement/snapshot and ordering
contracts while adding RLS. FOR SHARE authority/liveness queries require compatible
privileges and row policies; tightening UPDATE cannot silently make those locks
return no rows or force a weaker read. Pin the least-privilege lock/read design
before Plan ACCEPT and prove that any privilege needed for locking does not allow
unauthorized request-path role/status/structural mutation; preserve the bounded
trusted authentication/lifecycle paths of R6. A broader lock or security-model change
requires explicit review rather than an incidental migration adjustment.

Retain G007's declared authorization linearization and transaction/locking contract with actual session/model writers. A new operation after a committed revoke/downgrade, unlink, policy change, account disable, session revoke or expiry cannot retain a removed allow path through an old context, pooled connection or prepared query. Alternative independent Organization or qualifying attached-Group authority still works when valid.

RLS authority lookups cannot open a new unchecked check-to-write gap. Use deterministic two-connection barriers to verify both orderings of a protected write and each relied-on mutable authority source. Either the authorized write is serialized before the conflicting change or the changed authority is observed and the write denies/rolls back. Pin supported isolation/locking behavior; do not claim PostgreSQL automatically makes cross-table policy checks race-free. Reuse G007 scenarios and add the changed database path, rather than build another orchestration system. Broad incompatible predecessor changes require review.

### R9 — Migrations and Portable Provisioning

Add migrations after accepted history; never edit prior migration bytes. Pin policy/role/helper/view/grant changes, dependencies and lock/transaction behavior. Upgrade populated predecessors without changing UUIDs, entity data, Organization/Group memberships, associations, session mappings or policy values. Verify empty full migration, populated upgrade, repeat up/status, restart persistence and down/reapply only in owned disposable databases. Down restores the exact predecessor catalog/grants without deleting predecessor data; it temporarily removes G008 protection and is never an operational production fallback. Preserve production down/reapply refusal and owner-only migration metadata.

Document the required roles/grants/ownership and a bounded executable provisioning procedure for PostgreSQL without image-init hooks. Exercise it on a fresh owned instance/database lacking the normal init setup, then migrate and repeat role/representative RLS checks. Verify repeat behavior and failure handling as pinned in the Plan. Do not provision AWS/GCP, select a managed vendor, purchase infrastructure, rotate live credentials or create a generalized provisioning framework. If adding the dedicated helper role, reconcile first-start and existing-volume upgrade paths without giving core administrative credentials.

### R10 — Validation, Privacy and Evidence

Preserve the accepted commit-outcome contract: success only after confirmed
COMMIT; definitive abort only when established. FATAL/PANIC, unknown severity or
lost replies can mean `commit_outcome_unknown`, public unavailable, no automatic
retry and no assertion that rollback erased a possibly committed write. RLS or
pool cleanup errors must not turn uncertainty into successful authorization or
connection reuse. Retain the covering classifier and real transaction evidence.

Use real accepted-version PostgreSQL with the actual runtime login and synthetic owned fixtures: two Organizations, at least three Groups, shared and unlinked BusinessProfiles, multiple global identities with differing Organization/Group roles and sessions. Each isolation denial has an in-scope positive control and exact expected row/state/SQLSTATE mechanism. Filtering may correctly return zero rows; not every RLS denial is an SQL error. Cannot pass merely because privilege denial prevents both positive and negative operations.

Required bounded negative controls demonstrate detection of missing FORCE, a permissive/bypassing policy or privileged assertion connection, a projection/helper privilege defect, and session-level context leakage. Changes occur only in owned disposable fixtures/rolled-back catalog controls; production RLS is never disabled to make tests pass. Reuse accepted catalog/connection guards; deterministic violations and intended failures are recorded, without a new generic validation framework or expanded hosted-failure campaign.

Logs, database errors and retained evidence contain no credentials, session/provider tokens, auth hashes, email/private profile/membership payloads or sensitive row/parameter details. Use existing safe correlation/reason codes and inherited canary/planted-leak controls. Record runtime identity/attributes, exact catalog/policy/command inventory, commands/totals/exit codes, baseline/input/scope/tree hashes, connection-reuse/concurrency outcomes and owned cleanup. Preserve failures, interruptions, skipped/unmet evidence and shared Goal-level attempt accounting.

Reuse accepted container commands, wrapper/drift/status rules, intelligence-disabled variant, customizations and fast/affected/full CI matrix. Reconcile version-specific pre-RLS catalog assertions at predecessor versions and assert the exact new RLS inventory at the new version; no broad exemptions or deletion of old meaning. Shared-check edits rerun their covering controls in the same covering batch. Required final local full, automatic MR fast and the applicable qualifying post-integration staging disposition cover the final submitted content under current governance. A canceled staging run is recorded as canceled, not passing tests; any failed job or known failing required check blocks acceptance. A forced local RLS-test failure plus accepted CI failure controls proves gating; no additional hosted failure campaign absent a demonstrated gap. Access/quota failures remain unmet evidence. Feature integration uses the accepted glab MR-to-staging/source-deletion route.

## Acceptance Criteria and Verification Strategy

| ID | Required outcome | Objective verification/evidence |
|---|---|---|
| AC1 | Exact relation/command inventory covers all accepted facts. | Catalog versus pinned Plan inventory; ENABLE/FORCE on every tenant and global identity table, exact per-command policies/grants and explicitly justified non-tenant exclusions. Missing/unexpected relation or broad policy controls fail. |
| AC2 | Runtime and privileged roles retain R2 separation. | Actual runtime login/current-role attestation; ownership, membership, PUBLIC/column/function/sequence grants; DDL, SET ROLE, TRUNCATE, metadata and row_security-off negatives. Owner FORCE positive/foreign denial and privileged-test-connection detector pass. |
| AC3 | Context is trusted and local to protected transactions. | Valid accepted sessions resolve context; fabricated subject/context, absent/expired/revoked/disabled identity and malformed/partial/zero/foreign scope cases deny for pinned reasons. Production query-boundary inspection proves no pool/raw-method bypass. |
| AC4 | Direct runtime SQL enforces Organization isolation for every permitted command. | Same fixed actor with missing application WHERE predicates: exact SELECT/aggregate/join/RETURNING rows; foreign INSERT/UPDATE/DELETE outcomes and unchanged foreign rows, paired with authorized positives. Include missing/malformed context. |
| AC5 | Group, shared BusinessProfile and policy enforcement match G007. | Own/all/selected scope, two linked Groups, Organization authority and Group-derived access, unlinked profiles and both policy values; exact allowed reads/ordinary updates, sibling Group/link denial and sensitive/authority escalation negatives. Independent expected matrix outcomes.  One user manages profiles through Groups in two Organizations; active-tenant queries exclude the other tenant and member-only management paths. Removing the last owner/admin path removes management eligibility without a profile-role row update; a second qualifying attached Group preserves only its allowed path. |
| AC6 | Immutable ownership and constrained authority survive database writes. | Direct runtime attempts to reparent/retarget/change UUID/subject/role/source or self-grant fail; existing composite FKs, uniqueness and typed-scope controls pass at their intended layer with unchanged data. Pair accepted FOR SHARE liveness/authority locking positives with self-role/status/session manipulation negatives, so required lock privileges cannot become a mutation grant. No copied authority appears. |
| AC7 | Projections and authority helpers cannot bypass the boundary. | Base/view/join comparisons; SELECT-only invoker projections, exact helper/role/search_path/grant allowlist and bounded-return checks. Deliberate privilege/projection defect fails its covering control; no recursion on valid or denied queries.  A derived Business Profile access query/view is read-only and cannot export sibling Group membership or provide a copied authority store. |
| AC8 | Global identity is protected and pre-tenant authentication still works. | Self/shared-permitted-scope/foreign identity row cases; no other-user auth material. Full accepted G004 callback/session/self/logout/disable cases before Organization creation/selection, forged bootstrap UUID and challenge cancellation negatives. Ordinary optional-MFA behavior unchanged. |
| AC9 | Physical pooled reuse carries no prior request authority. | One-connection backend-verified commit/rollback/error/init-failure/cancel reuse across tenants and optional Group reset; context-free access denies. Poisoned prior session setting is rejected/cleared or connection discarded. Unsafe control exposes the intended leak; connection-loss/failed-commit outcomes follow pinned driver semantics. |
| AC10 | Current authority and concurrency guarantees remain effective under RLS. | Reuse old context/connection/query after each committed authority/session change; correct denial or surviving independent path. Deterministic two-connection ordering cases retain G007's association-read linearization, declared user/session/Organization/membership/Group lock order and the unlink Group-lock surrogate, with no unauthorized mutation. Retain changed-slug lock-upgrade and uncertain-COMMIT behavior; no unavailable response is assumed to prove a rollback. |
| AC11 | Migration and non-init provisioning preserve accepted data and privileges. | Empty/populated/repeat/restart/disposable down/reapply; predecessor rows/migrations unchanged and restored catalog exact. Provision on owned no-init database, repeat/failure controls, migrate and rerun actual-role/RLS positives and negatives. |
| AC12 | Privacy, scope and inherited regressions remain effective. | Real success/deny/error canaries with intentional leak detection; applicable auth, model/policy, browser/theme/scopes, contracts, variants, dependencies and versioned catalog controls pass. Route/schema/import comparison shows no public/storage/audit/product or vendor expansion. |
| AC13 | Actual final implementation is gated locally and on GitLab. | Documented commands verbatim against submitted root/hash-identical disposable copies; approved-wrapper required local full succeeds with exact content/input evidence and no required skips; automatic MR fast and post-integration staging disposition follow current governance. Staging success or owner-canceled satisfies its status gate only with no failed job or known failing required check; canceled is never passing-test evidence. Pending/running/failed does not qualify. Forced local policy-test failure plus inherited CI controls cannot pass. |
| AC14 | Provenance, budget and review gates are complete. | Exact accepted baseline/inputs/scope, pre/post drift/NUL-safe status, sanitized evidence, failed/interrupted history, owned cleanup, B7/attempt-4 checkpoint and Result. Codex actual-diff ACCEPT; handoff lists G009/G010 obligations without claiming G010/G021 completion. |

## Expected Artifacts and Required Validation

Minimum additive database migrations/policies/grants/helpers, core transaction/context integration and containerized runtime-SQL/pool/concurrency/migration tests. Preserve exact accepted lower-version catalog/ACL checks and add final-version RLS expectations; joined-view DML may reject with SQLSTATE 55000 before ACL evaluation, so independently assert SELECT-only grants rather than forcing every negative to be 42501. Extend existing tenancy/authorization/migration documentation and add focused `docs/foundation/DATABASE-ISOLATION.md` for exact table/command policies, privileged/pre-tenant paths, settings trust limits, provisioning, pool lifecycle and G009/G010 handoff. Keep source skill and accepted historical artifacts unchanged; enumerate exact consumer customizations. No Plan or Result is created in this drafting step.

Claude's Plan maps AC1–AC14 to named scenarios, exact expected rows/denials/state and evidence; pins connection identity checks, catalog inventory, command/column privileges, context API, authentication/authority-reader exceptions, migration/provisioning and concurrency contracts. These tests are implemented/run during authorized execution, not as a parallel preapproval framework. The Result distinguishes fixture setup, actual runtime assertions, owner/control paths, root-stack checks and hosted evidence. Full G010 multi-surface attacks and G009 metadata/storage remain later; every database guarantee above is completed in G008.

## Permitted Deviations

Module/helper names, exact context settings, reviewed query/locking/index strategy and minimal provisioning representation belong to the Plan. They must preserve accepted policy semantics and all protection/cleanup guarantees. A denied command may use explicit privilege denial or default-deny RLS as pinned; claim only the demonstrated mechanism. No deviation permits runtime owner/BYPASSRLS, broad definer/admin flags, session-level authority, unprotected tables/views, copied grants, authentication regression, weakened tests or deferred G008 evidence.

## Risks and Decisions Before Implementation

1. **Accepted predecessor integration:** verify the above version-5 inventory, current eight-operation facade, authentication access and grants against the actual execution baseline; no provisional representation may override the accepted implementation.
2. **Authority recursion and pre-tenant identity:** pin the smallest protected lookup/fixture/provisioning paths and global identity shared-scope/column visibility. Any new broader administrative capability or changed privacy/identity policy returns to Codex/owner.
3. **Command mapping:** G007 exposes bounded reads/ordinary updates, not workflows for every matrix capability. Deny unsupported request mutations, explicitly preserve trusted setup, and review actual permitted DML without implementing new administration/onboarding.
4. **Pool and races:** prove actual driver reset/discard and accepted-writer compatibility; settings do not authenticate arbitrary SQL clients and RLS is not automatic revocation serialization. Escalate a gap rather than quietly promising more or weakening ordering.
5. **Migration/provisioning feasibility:** review helper-role creation, populated upgrade/lock impact, FORCE owner row work and no-init setup. Major infrastructure/provider changes or destructive live actions are excluded.
6. **Query/index/locking cost:** assess the RLS authority-lookup/join and recursive-reader footprint against current subject-scoped queries and existing indexes under PG-001/PG-002. Identify workload assumptions and bounded representative query-plan/lock evidence needed during execution. Flag unexpected broad scans, fan-out, lock waits, migration locks or a need to tune schema/indexes to the owner and performance register before implementing a mitigation. No invented latency target, full benchmark campaign or security-lock relaxation.
7. **B7 and evidence access:** forecast against the reconciled scope. Propose a bounded split before execution if needed; no required database or hosted acceptance evidence is deferred to G010. No extra live WorkOS accounts/factor qualification are prerequisites solely for RLS fixtures.

## Completion and Gate Conditions

Accepted-predecessor reconciliation is complete at the inspection baseline; G007/transitive dependency gates are satisfied. This revision authorizes specification work only. Resolved security/migration/lock-read pins, explicit Claude Plan ACCEPT, fresh committed input/scope binding and applicable hosted-action authorization precede implementation. G054 manual testing/publication is independent and creates no dependency. Complete AC1–AC14, Result and actual-diff Codex ACCEPT before G009 implementation. No product-specific work before G021 acceptance.

## Revision Record

- 2026-10-07 — r1 drafted at owner request after checking the active G004 session. Fourteen criteria, B7/checkpoint four and explicit provisional predecessor dependencies. Management artifacts only; no application changes, Claude dispatch, migration/build/test/provider/CI work, commit or push.

## Inherited owner clarification — 2026-10-08

Accepted [ADR-004](../../../adr/ADR-004-group-business-profile-cardinality.md) and [the cardinality/visibility addendum](G006-business-profile-cardinality-addendum.md) govern predecessor reconciliation and relevant existing acceptance cases. One completed Group has one Business Profile; multiple Groups may share it. A global user can manage multiple profiles, but only active-Organization profiles with a legitimate policy allow path are visible. Verify one user with grants in two Organizations cannot carry profile visibility or authority across the active-tenant boundary; shared profiles expose no sibling-Group links/memberships. BusinessEntity is profile/context, not a third tenancy tier or a required identifier on all Group resources. Existing capability matrix, budgets, criterion counts, predecessor gates and G021 remain. No new UI/public listing, onboarding workflow or implementation is authorized by this clarification.

## Owner authority revision — 2026-10-08

r2 follows ADR-004 and G007 r2: profile access derives from current Organization/Group bindings and attachments, never direct BE grants. The r1 Spec (not included) preserves prior design. The earlier inherited clarification is superseded where it assumes independent profile grants. Existing RLS/connection/transaction/privacy guarantees, fourteen criteria and B7 remain; add relevant last-path-removal/alternate-path and active-tenant cases to AC5/AC7. Predecessor/Plan/execution/G021 gates remain.

## Owner naming clarification — 2026-10-08

Use canonical `BusinessProfile` (UI: **Business Profile**) instead of `BusinessEntity`; the shared association becomes `GroupBusinessProfile`, future identifier spelling `business_profile_id`, and capability prefix `business_profile.*`. UI placement is **Group settings → Business Profile**; `GroupProfile` is not the reusable record name. Earlier names in archived/detailed planning records are historical. Stable Goal IDs, management filenames/links, UUID identity and all cardinality/derived-role/active-tenant guarantees remain. Claude must reconcile affected provisional Plan/schema/interface/fixture pins before execution; this does not authorize edits to closed artifacts, source Skills or active G004 inputs. Criteria, budgets, checkpoints and G021 are unchanged.

## Accepted-state revision — 2026-10-13

r3 supersedes r2's provisional baseline and obsolete paths/qualification/budget
wording. Exact r2 archive (not included) preserves history.
Fourteen criteria and B7 remain; its unit is now shared Goal-level execution
attempts under current BUDGET_POLICY, checkpoint after attempt 4. Update the
actual schema/guarded-operation inventory, row-lock compatibility requirement,
commit-outcome invariants, performance assessment and current solo-operator CI
status semantics. Historical approvals/failures are unchanged. No application,
SQL migration, provider configuration, test execution or G021 release occurs.
