# Authorization: trusted context and policy

This page describes the trusted request context and the shared policy evaluation over the tenancy model. Core
resolves an immutable request-scoped `TenantContext` from an authenticated session and evaluates a closed capability
matrix over the current RoleBinding facts, the Group–BusinessProfile associations and one Organization policy flag. A
small guarded facade authorizes and acts in one database transaction. The model it reads is described in
[TENANCY.md](TENANCY.md), and sessions in [IDENTITY.md](IDENTITY.md).

Nothing here is wired. No binary, route, handler, CLI, environment variable or bootstrap reaches the package, and no
public tenant operation exists. From version 6, PostgreSQL row-level security enforces the same boundary beneath this
facade ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)): every facade transaction starts with the context
statement (`access.inContextTx`), and an omitted predicate cannot reach another Organization or an unauthorized Group.

| Part | Where |
|---|---|
| Request scope, context, evaluator and guarded facade | `core/internal/tenancy/access` (package `access`); no SQL |
| Ordered fact reads and transaction-scoped writes | `core/internal/tenancy/facts.go`, statements in `core/internal/tenancy/store.go` |
| Session liveness inside a tenant transaction | `identity.Service.LockLiveSession` and `CheckLiveSession` (`core/internal/identity`) |
| Policy flag | `core/migrations/00005_organization_policy.sql` (version 5); see [MIGRATIONS.md](MIGRATIONS.md) |
| Real-PostgreSQL tests | `core/internal/migrate/policy_integration_test.go`; `core/internal/tenancy/access/{integration,footprint_integration,ordering_integration}_test.go` |

## Request scope and TenantContext

**One authenticated request, one scope.** `BeginRequest(ctx, credential)` authenticates through
`identity.Service.Session`, the accepted identity path including its throttled renewal. It returns a derived `ctx` that
carries a new, unexported request scope (the session UUID and its subject), plus an `end` function. The contract for
every future adapter:

- Call `BeginRequest` exactly once per request, and defer `end`. `end` is idempotent and safe for concurrent use. After
  it, the scope and every context resolved in it are refused.
- A `ctx` that already carries a scope, from this or another `Service`, is refused (`context_denied`) before identity is
  called. One request chain has exactly one authenticated subject.
- Derived contexts (deadline, cancellation, values) keep the scope, so legitimate calls in the same request work.
- The scope is identified by pointer identity, so two requests never share one, even with the same credential. It holds
  no credential, CSRF token, display name, provider session, assurance, role or allow result. It is a binding, not an
  authority snapshot: every operation rechecks the session, the user and the facts.
- Request IDs and traces are **never** a binding. Only `access/log.go` reads them, to annotate log records; they are
  never compared, stored or used in a decision. Hostnames, headers, cookies, provider claims, email, names, slugs and
  model data have no input path. Transport checks such as origin and CSRF remain the adapter's obligation.

**Resolving a context.** `Resolve(ctx, ResolveRequest{OrganizationID, GroupID})` takes selectors, never authority:
canonical UUIDs of an Organization and, optionally, a Group of it (empty means Organization-only).

1. The `ctx` scope must exist, must have been issued by this `Service` and must not have ended. This check is pure and
   runs before any SQL.
2. Malformed, braced, unhyphenated, empty or all-zero selectors are `invalid_input`.
3. In one REPEATABLE READ READ ONLY transaction: the scope's session must still be live (`CheckLiveSession`); the
   Organization must exist and the subject must hold a current Organization membership in it; with a Group selector,
   the Group must belong to that Organization and the subject must be `org_owner` or `org_admin`, or hold a binding in
   that exact Group.

An unknown or foreign Organization or Group, no Organization membership, and a Group binding without an Organization
binding all give the same `context_denied`. Nothing is written and there is no default Organization. Several contexts
may be resolved in one scope (for example Organization-only, then a selected Group); each is bound to that scope.

**TenantContext** has unexported fields only: issuer, scope, subject, session, Organization and optional Group. It
stores no request ID or other correlation value, and no role, capability, assurance or allow result. Its read-only
accessors (`Subject`, `Organization`, `Group`) confer nothing.

**Binding check.** Every guarded operation runs this check before any transaction. The first failing condition decides
the internal reason; every failure is external `context_denied`.

| # | Rejected when | Reason |
|---|---|---|
| B1 | the context is nil, zero-valued or caller-built | `context_forged` |
| B2 | the context was issued by another `Service` instance | `context_forged` |
| B3 | `ctx` carries no scope, or one issued by another `Service` | `context_scope_missing` |
| B4 | the `ctx` scope has ended | `context_scope_ended` |
| B5 | the context's scope is not the `ctx` scope (pointer identity) | `context_scope_mismatch` |
| B6 | the context's subject or session differs from the scope's | `context_subject_mismatch` |

Cross-request substitution fails B5, even with the same credential, equal inbound `X-Request-Id` values or no request ID.
A context of user A under user B's scope fails B5 and B6. Use outside an authenticated request fails B3, and use after
`end` fails B4, even through `context.WithoutCancel`. B6 cannot fail when B5 passes; it is kept as defence in depth.

The check is in process. A component inside the trusted core process can always misuse its own request's pair, which is
the same trust boundary as the raw model primitives. Even then, every operation applies only that subject's current
authority.

## Capabilities and the matrix

`decide(capability, context, facts, target)` is a pure, closed evaluator. Unknown capabilities, roles, sources or target
kinds, and inconsistent facts, deny (`capability_unknown`, `fact_inconsistent`). There is no ranking, prefix matching or
global authority, and no allow result, capability check or evaluator is exported. Gates apply in order:

1. **Organization gate:** the facts carry the active Organization and exactly one current Organization binding with
   source `membership`. A Group binding never creates Organization membership.
2. **Target gate:** Organization capabilities and `audit.read` target the active Organization, `group.*` an exact Group of
   it, and `business_profile.*` an exact profile of it. A foreign, unknown or wrong-kind target denies, even for a
   subject holding the named role.
3. **Context gate:** in selected-Group context, a `group.*` target must be the selected Group.
4. **Matrix.**

| Capability | org_owner | org_admin | org_member | group_owner | group_admin | group_member |
|---|---|---|---|---|---|---|
| organization.read | Y | Y | Y | — | — | — |
| organization.manage_members | Y | Y | — | — | — | — |
| organization.manage_policy | Y | — | — | — | — | — |
| group.read | all | all | — | own | own | own |
| group.update | all | all | — | own | own | — |
| group.manage_members | all | all | — | own | own | — |
| group.create | Y | Y | Y | — | — | — |
| group.complete_onboarding | all | all | — | own | own | — |
| business_profile.profile.read | Y | Y | — | linked | linked | linked |
| business_profile.profile.update | Y | Y | — | linked+flag | linked+flag | — |
| business_profile.select_for_onboarding | Y | Y | — | linked | linked | — |
| business_profile.privileged_settings.manage | Y | — | — | — | — | — |
| audit.read | Y | — | — | — | — | — |

- `all`: an Organization owner or admin path to any Group of the active Organization. `own`: the subject's binding in the
  exact target Group.
- `linked`: a candidate Group with the required Group role whose current association is the profile. The candidates are
  the selected Group in selected-Group context, otherwise every Group of the subject in the active Organization.
  `linked+flag` also requires the Organization's `group_admins_edit_business_profiles` to be true.
- Organization owner and admin profile paths are independent of the selected Group and include unlinked profiles.
  `select_for_onboarding` ignores the flag. Member-only paths allow reading only.
- No Group path confers privileged settings, Organization policy or audit authority. There is no BusinessProfile role or
  grant: profile authority is computed from Organization roles and Group associations at evaluation time.
- `group.manage_members` is not authority to grant, change or remove a `group_owner` binding; ownership transfer and
  last-owner rules belong to membership administration (version 15).
- `group.create` is an Organization capability held by every current
  Organization membership role, in Organization-only context only: a selected-Group context denies it
  (`role_not_permitted`). A Group role alone never satisfies the Organization gate. `group.complete_onboarding` follows
  the `group.update` rule. Fifteen capabilities in all (thirteen before Group onboarding).

**Interpretation pins.**

- (a) A Group-role subject is evaluated with an `org_member` binding holding the Organization gate. The result is the
  combination of independent paths: `group_owner` plus `org_member` may `organization.read` through `org_member` only.
- (b) An Organization owner's or admin's `all` path is restricted by a selected-Group context to that Group.
- (c) Policy read is `organization.read`; policy update is `organization.manage_policy`.
- (d) The manageable-profile set is every profile of the active Organization for which `select_for_onboarding` allows,
  under the current context. Organization-only context therefore deduplicates across attached Groups; selected-Group
  context restricts only the Group-derived paths.

**Policy flag.** `app.organizations.group_admins_edit_business_profiles` is `boolean NOT NULL DEFAULT true`: every
existing and new Organization reads `true` as the configured default, never as a fallback. A read error is
`unavailable`, never an allow. The runtime role may only `UPDATE` the column, so a creator cannot preset it. The flag
affects inherited ordinary profile updates only, never reads or selection. Request data, environment, claims and other
Organizations cannot override it. There is no policy map, rule editor, MFA policy or public endpoint.

**MFA.** Assurance stays `unknown` (IDENTITY.md). Two otherwise identical subjects with and without enrolled factors have
identical capabilities, and no capability requires MFA.

### Explicit Group authority (supersedes the `all` management cells above)

The matrix above keeps its historical (pre-version-17) meaning for
`group.read` and `business_profile.profile.read` only. Since migration 00017, an Organization owner's or admin's `all`
path is **metadata only**: it lists and reads the active Organization's Groups and ordinary Business Profiles (including
unlinked ones) and selects a Group (MCP `context.list_groups`/`switch_group`, `rls_org_role` selected-Group validity),
which confers nothing by itself. Ordinary Group management — `group.update`, `group.manage_members`,
`group.complete_onboarding`, `business_profile.profile.update`, `business_profile.select_for_onboarding` and invitation
issue/list/revoke/candidates — requires an explicit `group_owner`/`group_admin` binding in the target Group (`own`,
`linked`, `linked+flag` unchanged) **or** the Organization-path `orgMgr` = Organization owner/admin **and**
`app.organizations.group_management_override` true. Group content never follows from an Organization role:

| Caller without a Group binding | Metadata reads | Ordinary Group management | Group objects | Group owner powers | Change the override |
|---|---|---|---|---|---|
| Organization owner, override off | Yes | No | No | No | Yes |
| Organization admin, override off | Yes | No | No | No | No |
| Organization owner/admin, override on | Yes | Yes (bounded) | No | No | owner only |
| Organization member | own Groups only | No | No | No | No |

- A caller with a Group binding gets exactly that role's powers unioned with the row above: an Organization admin who is a
  Group member keeps member-level content and administers the Group only through the override.
- The explicit Group path is decided first; an allow through `orgMgr` alone carries reason `allowed_override` and, on a
  write, appends `access.override_used` in the mutating transaction (an append failure refuses the write;
  [AUDIT.md](AUDIT.md)).
- The override never grants `group.manage_owners`, owner-role invitations, a last-owner bypass, Leave, a self membership
  or Group objects, and it cannot be used on oneself: a Group invitation to its own issuer is refused (`self_target`),
  the claim body never consumes one, and the authority guard refuses a runtime/owner session raising its own Group role
  without that Group's owner role (`self_escalation`).
- **The override.** `group_management_override boolean NOT NULL DEFAULT false`; existing Organizations read `false` (no
  inference from earlier behaviour). Only a current, enabled Organization owner sets it: `SetGroupManagementOverride`
  (operation `organization.set_group_override`, capability `organization.manage_policy`), and in SQL the owner-owned
  guard refuses any other session (`group_override_owner` → `forbidden`), so a direct runtime UPDATE by an admin fails.
  `SetOrganizationPolicy` cannot set it. `GetGroupAuthority` returns `{override_enabled, can_change}` to every member.
- **Liveness.** Facts are read under the Organization row lock every guarded write already takes (`FOR SHARE`) and the
  write takes `FOR NO KEY UPDATE`, so a committed disable denies every later override-based call; the invitation claim
  re-checks an override-only issuer under the Organization row and the issuer's enabled user row
  (`issuer_authority_lost`). No session, MCP catalog or web state caches the setting.
- **Objects.** Group object read/metadata/sign/put require an explicit Group role; delete a Group owner/admin or the
  member creator; Organization objects are unchanged ([OBJECT-STORAGE.md](OBJECT-STORAGE.md)).

## Guarded operations

| Method | Capability | Target | Transaction | Write |
|---|---|---|---|---|
| `BeginRequest(credential)` | none (authentication) | — | identity's accepted resolve and renewal | — |
| `Resolve(selectors)` | none (scope check and context gate) | Organization (and Group) | REPEATABLE READ READ ONLY | — |
| `GetOrganization(tc)` | organization.read | active Organization | REPEATABLE READ READ ONLY | — |
| `GetOrganizationPolicy(tc)` | organization.read | active Organization | REPEATABLE READ READ ONLY | — |
| `SetOrganizationPolicy(tc, p)` | organization.manage_policy | active Organization | READ COMMITTED, ordered locks | the flag (absolute set) |
| `GetGroup(tc, g)` | group.read | exact Group | REPEATABLE READ READ ONLY | — |
| `RenameGroup(tc, g, name)` | group.update | exact Group | READ COMMITTED, ordered locks | accepted rename |
| `ChangeGroupSlug(tc, g, slug)` | group.update | exact Group | READ COMMITTED, ordered locks | accepted slug change |
| `GetBusinessProfile(tc, p)` | business_profile.profile.read | exact profile | REPEATABLE READ READ ONLY | — |
| `UpdateBusinessProfile(tc, p, c)` | business_profile.profile.update | exact profile | READ COMMITTED, ordered locks | accepted full replace |

- Every operation and `Resolve` runs the scope or binding check first, before any transaction. `BeginRequest` and
  `Resolve` are not policy decisions and carry no capability.
- Payloads keep the accepted tenancy contracts. `ValidateName`, `ValidateSlug` and `ValidateBusinessContext` run before
  authorization; they are pure and reveal nothing about targets.
- A shared-profile read returns only the profile, never its associations, access paths or sibling memberships.
- `group.manage_members` and `privileged_settings.manage` are evaluated (unit and database-loaded facts) but have no
  operation; `audit.read` guards `ReadAuditEvents` and `select_for_onboarding` the Group onboarding operations
  below. The manageable-profile set stays unexported; `ListBusinessProfiles` is its paged form.

### Guarded domain operations

`organization.manage_domains` is the twelfth capability: an Organization-scope, role-level capability held by the
current Organization owner or admin (the same rule as `organization.manage_members`). `core/internal/domains/api` is the
only caller ([DOMAINS.md](DOMAINS.md)); every method runs in the facade's guarded TENANT-ORG transaction with the lock
order user → session → Organization (FOR SHARE) → domain row (`FOR NO KEY UPDATE`), and DNS/TLS work runs between two
calls, outside every transaction:

| Method | Write |
|---|---|
| `ListDomains(tc)` | — (snapshot) |
| `ClaimDomain(tc, host, …)`, `RotateChallenge(tc, id, …)` | a pending row; a new challenge of a pending row |
| `BeginDomainCheck(tc, id, check)` / `RecordDomainCheck(tc, read, check, obs, …)` | — / the guarded state, freshness and TLS update (version guard) |
| `RemoveDomain(tc, id)` / `RecordWithdrawal(tc, removed, res, …)` | removal at commit / the removed row's cleanup outcome |
| `RetryWithdrawal(tc, id, max)` | the bounded retry precondition |
| `SetOrganizationSlug(tc, slug)` | the slug (Organization row `FOR NO KEY UPDATE`) |

Denials write `access.denied` with the `domain.*` and `organization.set_slug` operations; a raced change is
`domain_state_invalid` with zero effects. `domains` imports nothing from `access` or `identity`; the facade importers
are `core/internal/mcp` and `core/internal/domains/api` (`TestProductionImporters`).

### Guarded Group onboarding operations

`core/internal/onboarding/api` is the only caller: eight method/path pairs on six internal paths under
`/internal/v1/groups*` and `/internal/v1/business-profiles*`, for the web BFF only (not public API; the session-scope
helper is a byte-pinned copy of `branding/api`'s). Model semantics are in
[TENANCY.md](TENANCY.md#group-onboarding).

| Method | Capability | Target | Transaction | Write |
|---|---|---|---|---|
| `SuggestGroupSlug(tc, name)` | group.create | active Organization | REPEATABLE READ READ ONLY | — (advisory; reserves nothing) |
| `CreateGroup(tc, name, slug)` | group.create | active Organization, Organization-only context | READ COMMITTED, ordered locks | one keyed Group INSERT; its trigger writes the creator's `group_owner` binding |
| `ListGroups(tc, after, limit, slug)` | group.read, per row | one page of the Organization's Groups | REPEATABLE READ READ ONLY | — |
| `GetGroupOnboarding(tc, g)` | group.read | exact Group | REPEATABLE READ READ ONLY | — |
| `CompleteGroupOnboardingWithNewProfile(tc, g, c)` | group.complete_onboarding | exact Group without an association | READ COMMITTED, ordered locks | profile, then association |
| `CompleteGroupOnboardingWithProfile(tc, g, p)` | group.complete_onboarding on g, then business_profile.select_for_onboarding on p, on the same locked facts | exact Group without an association, exact profile | READ COMMITTED, ordered locks | association |
| `ListBusinessProfiles(tc, after, limit)` | business_profile.select_for_onboarding, per scanned row | one page of the Organization's profiles | REPEATABLE READ READ ONLY | — |
| `GetBusinessProfileView(tc, p)` | business_profile.profile.read (operation `business_profile.get`) | exact profile | REPEATABLE READ READ ONLY | — |

- **Completion facts.** The target Group is locked `FOR NO KEY UPDATE` and its association is read after the lock
  (step 7); for an existing profile the subject's other Groups are the candidate paths, locked `FOR SHARE`, with their
  associations and the profile. An existing association is `conflict` with `Field` `association_exists` and zero
  writes; the association unique key is the backstop. Either capability denial is recorded with the established
  target; an unreadable profile is `not_found`, a readable one without a path `forbidden` (`no_linked_path`).
- **`can_update` is a display hint.** `GetBusinessProfileView` runs a second `decide` of
  `business_profile.profile.update` on the same snapshot; the page renders the editor or the read-only view from it,
  and `UpdateBusinessProfile` always re-decides on fresh locked facts. The hint implies nothing about sibling Groups.
- **Lists** decide every row on the snapshot's facts (no per-row round trip); a list holds no authority.
- **Records.** The seven operations `group.suggest_slug`, `group.create`, `group.list`, `group.onboarding.get`,
  `group.onboarding.complete_new`, `group.onboarding.complete_existing` and `business_profile.list` join the
  `access.decision` record and the `access.denied` operations (29 → 36; [AUDIT.md](AUDIT.md)). The HTTP adapter adds
  one operational `onboarding.request` event per request (operation, outcome, closed reason, host tag; never UUIDs,
  names or payloads).
- **Defense in depth.** The version-14 INSERT policies admit exactly the statements these operations execute
  ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)), so a facade bypass with the runtime role meets the same refusal
  (42501).

**Mutation paths for a Group's association:**

| Path | Can it attach or remove a completed Group's profile? |
|---|---|
| guarded `CompleteGroupOnboarding*` | attach only to a Group without an association; `conflict` otherwise; never replaces |
| guarded `UpdateBusinessProfile` | edits the context; never touches associations |
| raw `tenancy.Service` primitives | trusted-only (import boundary), not a privilege boundary: under the runtime role with a context the same policies apply. `UnlinkGroupBusinessProfile` fails 42501 (no DELETE grant); `LinkGroupBusinessProfile` is admitted exactly when the guarded association policy admits it; `CreateBusinessProfile` (RETURNING) is refused for a Group-derived actor by the RETURNING visibility check; `CreateGroup`'s explicit binding statement is refused (only the trigger may insert a binding), so its transaction fails |
| runtime SQL with a fixed context | INSERT only through the four policies: never a second association (unique key), a direct binding (trigger-only policy) or an ineligible profile (eligibility reader); no DELETE |
| owner pool, migrator, DBA | FORCE-bound to the same policies with a context; the superuser is the documented restricted repair path, not an application path |
| MCP tools, browser | none touches associations; a Group page re-reads its state on every load |

### Guarded membership, ownership and invitation operations

`core/internal/membership/api` is the only HTTP caller: 16 method/path pairs on 15 internal paths for the web BFF
(not public API). Model semantics, the authority guard and the invitation lifecycle are in
[TENANCY.md](TENANCY.md#membership-ownership-and-invitations). Four capabilities are added (15 → 19):
`organization.manage_owners` (`org_owner` only), `group.manage_owners` (the exact Group's `group_owner` only; no
Organization-owner override), `organization.leave` (every Organization role) and `group.leave` (every role of the exact
Group). Ordinary member administration reuses `organization.manage_members` / `group.manage_members`, so an
Organization admin keeps ordinary management of the Organization's Groups but never owner controls.

| Method | Capability | Write |
|---|---|---|
| `ListMembers(tc, sc, after, limit, filter)` | manage_members of the scope (owner candidates: manage_owners) | — |
| `SelfMembership(tc, sc)` | the caller's own binding (not_found without one); `can_manage` is a display hint | — |
| `ChangeMemberRole(tc, sc, subject, role)` | manage_members; never an owner, never the actor's own binding | role UPDATE |
| `RemoveMember(tc, sc, subject)` | manage_members; Organization removal from an Organization-only context (explicit Group authority, version 17: the authority guard removes the subject's Group bindings inside the Organization DELETE; another subject's Group-owner binding refuses `owner_subject`) | one binding DELETE (plus the guard's nested Group DELETEs) |
| `LeaveScope(tc, sc)` | organization.leave / group.leave | own binding DELETE (plus the guard's nested Group DELETEs) |
| `PromoteToOwner(tc, sc, subject)` | manage_owners; an enabled same-scope admin | role UPDATE |
| `DemoteOwner(tc, sc, subject, role)` | manage_owners; the last owner is refused | role UPDATE |
| `TransferOwnership(tc, sc, recipient, formerRole)` | manage_owners; recipient promoted first, then the actor's own role or departure | role UPDATEs (and DELETEs for `none`) |
| `ListInvitations`, `GetInvitation`, `ListInvitationCandidates` | manage_members of the scope (a recipient reads its own invitation) | — |
| `ListMyInvitations(tc, after, limit)` | the caller's own invitations | — |
| `IssueInvitation(tc, sc, role, recipient, lifetime)` | manage_members; invitable roles only | invitation INSERT |
| `RevokeInvitation(tc, id)` | manage_members of the invitation's scope | `revoked_at` UPDATE |
| `AcceptGroupInvitation(tc, id, token)` | the intended recipient's context | one `invitation_claim_group` call |
| `ConsumeOrganizationInvitation(ev, id, token)` | verified-address evidence; owner-role pool only (public admission is the enrollment model's) | one `invitation_claim_organization` call |
| `RetirementEligibility(credential, origin)` | the session subject, IDENTITY-shape context | — |

- **Failures.** `last_owner`, `dependent_bindings`, `recipient_ineligible` and `issuer_authority_lost` are `conflict`
  with that `Field`; `owner_authority`, `self_escalation` and (version 17) `owner_subject` are `forbidden`; a gone or invisible binding or invitation is `not_found`; an
  Organization departure or removal outside an Organization-only context is `invalid_input` (`context`). Every class-40
  abort is `unavailable` (reason `aborted`), never retried.
- **Records.** The seventeen operations `membership.list`, `membership.self`, `membership.change_role`,
  `membership.remove`, `membership.leave`, `ownership.promote`, `ownership.demote`, `ownership.transfer`,
  `invitation.list`, `invitation.mine`, `invitation.candidates`, `invitation.issue`, `invitation.revoke`,
  `invitation.get`, `invitation.accept`, `invitation.consume` and `retirement.eligibility` join the `access.decision`
  record and the `access.denied` operations (36 → 53; [AUDIT.md](AUDIT.md)); the HTTP adapter adds one operational
  `membership.request` event per request.
- **Defense in depth.** The version-15 policies admit exactly these statements and the authority guard repeats the
  owner rules in the database, so a facade bypass with the runtime role meets the same refusal; the tokens' digest is
  unreadable by runtime.
- **Evidence (NOT EXECUTED until DB qualification).** The written MEM scenarios (`TestMembership`: MEM-10, MEM-11,
  MEM-14), O-MEM-1..3(d) and O-MEM-DUP
  (`TestAccessOrdering`), CT-MEM-1 (`TestMembershipCatalog`) and E-PG-MEM (`TestAccessFootprint`).

### Platform root signup (outside the facade)

Root signup ([IDENTITY.md](IDENTITY.md#platform-root-signup)) creates a tenant before any TenantContext exists, so
it is **not** a guarded facade operation and carries no capability: `core/internal/signup` imports neither `tenancy` nor
`access` (`TestProductionImporters` and the tenancy boundary tests are unchanged). Its authority is the caller's
own root signup credential (IDENTITY shape of that user) plus its CSRF token, and the database decides the write: only
the provision writer's `app.signup_provision_organization` can insert the Organization and the creator's `org_owner`
binding, atomically and only for the context user ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)). No actor, role or
tenant parameter is accepted. From the new owner's first tenant session
on, every request is the ordinary trusted context and policy above (an `org_owner` with no Group binding; the
default-off Group-management override). Refusals write `access.denied` with a `signup.*` operation and no Organization ([AUDIT.md](AUDIT.md)).

### Guarded label operations

Migration `00020_labels.sql`; the model, constraints and registration contract are in
[TENANCY.md](TENANCY.md#labels). Three closed capabilities are added (19 → 22), evaluated by the same `decide`
switch: `label.organization.manage` (Organization kind; Organization owner/admin), `label.group.manage` (Group kind;
the caller's own `group_owner`/`group_admin` binding in the exact Group) and `label.group.read` (Group kind; any own
role in the exact Group). `orgAdmin`, `orgMgr` and the Group-management override are never consulted for `label.group.*`:
Group-owned labels are Group content (the `decideObject` rule).

| Capability (kind) | org_owner | org_admin | org_member | group_owner(g) | group_admin(g) | group_member(g) | Override effect |
|---|---|---|---|---|---|---|---|
| `label.organization.manage` (organization) | Y | Y | — | — | — | — | none |
| `label.group.manage` (group g) | — | — | — | own | own | — | none (never consulted) |
| `label.group.read` (group g) | — | — | — | own | own | own | none |

- Organization-owned definitions and their Group assignments are read and written only through
  `label.organization.manage`; a Group role never reads them, and `target_members` visibility is refused for them
  (`visibility_not_allowed`, before SQL and by CHECK). Organization-owned assign, unassign and assignment listing also
  require `group.read` on the exact target Group in the same transaction (existence, same Organization, selected-Group
  context).
- Group-owned definitions need an actual role in the exact owning Group: managers read and write, members read only
  `target_members` definitions. Organization staff without a Group role, with the override on or off, get no list.
  `readCapability` maps both Group label kinds to `label.group.read`, so such a caller's refusal is `not_found`.
- A subject holding both roles independently (Organization admin and Group admin) reads each owner scope through its
  own path; the paths never combine into a third.

| Method | Operation | Capability | Write |
|---|---|---|---|
| `ListLabelDefinitions(tc, sc, after, limit)` | `label.definition.list` | Organization scope: `label.organization.manage`; Group scope: `label.group.manage`, else `label.group.read` for the `target_members` subset | — |
| `CreateLabelDefinition(tc, sc, name, visibility)` | `label.definition.create` | manage of the scope | definition INSERT |
| `UpdateLabelDefinition(tc, id, name, visibility)` | `label.definition.update` | manage of the definition's owner scope (row read first, then decided) | name/visibility UPDATE (absolute set) |
| `DeleteLabelDefinition(tc, id)` | `label.definition.delete` | manage of the definition's owner scope (row read first) | assignment DELETE, then definition DELETE |
| `ListLabelAssignments(tc, target)` | `label.assignment.list` | `group` target: `label.organization.manage` and `group.read` on the Group (a future registered Group-owned kind: `label.group.read`/`manage` on the owning Group plus the adapter's `readable`) | — |
| `AssignLabel(tc, id, target)` | `label.assign` | Organization-owned definition on a `group` target: `label.organization.manage` and `group.read` on the Group (a future Group-owned kind: `label.group.manage` on the owning Group) | assignment INSERT |
| `UnassignLabel(tc, id, target)` | `label.unassign` | as `AssignLabel` (definition row read first) | assignment DELETE |

- **Failures.** A taken scoped name is `conflict` (`name_conflict`), a full owner scope (200 definitions) `conflict`
  (`scope_full`), a duplicate assignment `conflict` (`already_assigned`), an absent one `conflict` (`not_assigned`);
  an unknown or unregistered target kind, or a Group-owned definition aimed at a Group, is `invalid_input`
  (`target_unsupported`) and more than one target field `invalid_input` (`target_ambiguous`), both before any SQL; an
  unknown, foreign or unreadable definition is `not_found`. Deleting an Organization-owned definition with a selected
  Group in the trusted context is `invalid_input` (`organization_context_required`) with zero writes and no audit row;
  the refusal does not depend on whether any assignment exists (lock order below).
- **Records.** The seven operations join the `access.decision` record and the `access.denied` operations (73 → 80;
  [AUDIT.md](AUDIT.md)); no record carries a label name, target id or payload, and no label operation appends
  `access.override_used` (no override path exists). `core/internal/labels/api` is the only HTTP caller (seven
  method/path pairs on six internal paths for the web BFF; [WEB-SHELL.md](WEB-SHELL.md#labels-pages)) and adds one
  operational `labels.request` event per request.
- **MCP.** No production label tool exists; `eligible` follows `decide` for the three capabilities (Organization staff
  for `label.organization.manage` only; `label.group.*` only with an explicit Group role; the override is irrelevant).
  Test-only fixture tools prove eligibility and hidden direct-call behaviour only ([MCP.md](MCP.md)).
- **Defense in depth.** The seven version-20 policies admit the same scopes at the database
  ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)), so a facade bypass with the runtime role meets the same refusal.

### Guarded MCP connection operations

`core/internal/mcp` is the only caller ([MCP.md](MCP.md)). Its request scope comes from a verified Connect token, never
from a browser credential: `BeginVerifiedRequest` takes the connection identity already read and checked by
`identity` (binding, client, liveness, renewal) and establishes a request scope bound to that connection's user and
Organization. No caller-supplied Organization exists. The operations carry no capability of the matrix; their rule is
current Organization membership, plus the Group rules of `context.switch_group`:

| Method | Rule | Transaction | Write |
|---|---|---|---|
| `CreateMCPConnection(tc, …)` | current Organization role of the browser session's user on the Organization the request host currently resolves to (formerly the registry host's) | TENANT-ORG | the pending row and its `mcp` session |
| `ListMCPConnections(tc, after)` | current Organization membership; the caller's own live connections only, newest first (≤ 100 per page; `after` is an untrusted keyset position) | snapshot | — |
| `ActivateConnection(ctx, client)` | current Organization role; one pending-to-active winner | TENANT-ORG, row lock | state, client, activation time |
| `AuthorizeConnection(ctx)` | current Organization role and, when a Group is selected, its current validity | snapshot | — |
| `GetConnectionContext(ctx)`, `ListSelectableGroups(ctx, after, limit)` | as above; Organization owners and admins list every Group, other members only their own (≤ 100 per page) | snapshot | — |
| `SelectConnectionGroup(ctx, g)` | a select needs the target selectable (as listed) and any current selection still selectable; a clear needs membership only. A decision snapshot, then one guarded write in the Organization-only shape | READ COMMITTED, ordered locks, connection row last | selected Group and context version |

The guarded write runs in the Organization-only shape for a select and a clear alike, where the subject's own bindings
in every Group are visible (row-level security). Lock order: the user and the connection's session, then the Organization, the
Organization binding, the subject's bindings in the current and target Groups (group_id order) and those Group rows
(`ReadFacts` lock mode), then the connection row. It decides on the locked facts: membership; for a select, the current
selection still valid (after the selected Group is lost only a clear or logout succeeds, `context_denied`
otherwise) and the target selectable. It stores the selection only if the connection still holds the one the decision
snapshot saw (any change in between is a `write_conflict`, reported to the tool as `unavailable`). The 00010 guard
re-validates on the same facts: a selection change only in the Organization-only shape, naming a Group (and leaving a
previous selection) visible to the subject under row-level security; a clear needs nothing more.

Termination is not a facade operation: the logout tool and web revoke go through `identity.RevokeMCPConnection` with
exact ownership only, so a former member can still end their own connection. A refused protected call is
`access.denied` with operation `mcp.request`; a binding mismatch of a verified subject is `resource_mismatch`
([AUDIT.md](AUDIT.md)). The operations write `access.decision` with the `access.begin_verified_request` and `mcp.*`
operation values and no capability.

### MCP discovery

Each MCP tool declares one opaque, comparable `ToolRequirement`: `RequireBinding()` (logout), `RequireMembership()`
(`context.switch_group`), `RequireContext()` (the read tools) or `RequireCapability(name)`. `RequireCapability` accepts
only the canonical capability values (eleven when discovery was introduced, fifteen since Group onboarding); any other name is `invalid_input`. These four constructors and
`AuthorizeCatalog(ctx, reqs) ([]string, error)` are the only additions to the exported surface. `AuthorizeCatalog` is
discovery, not authorization. It runs one snapshot in `AuthorizeConnection`'s shape and writes nothing. It returns
eligible names only: no context, facts, decision or allow result for a target. No operation accepts its result.
`decide`, `Capability` handling and the manageable-profile set stay unexported, and there is no per-tool check method.
Every tool still authorizes its own execution: the MCP connection handlers above for the control tools, and its own guarded
operation for any future capability-mapped tool.

| Requirement | Eligible when (after `contextCheck` holds) | Extra facts read |
|---|---|---|
| binding, membership, context | always; the tools keep their own connection checks, and nothing runs in front of them | none |
| Organization capability (`organization.*`, `audit.read`) | `decide` allows the active Organization | none |
| Group capability (`group.*`) | an Organization owner or admin (the `all` cell: any Group, from the Organization binding alone, restricted to the selected Group in selected context); otherwise `decide` allows some candidate Group | Organization-only: every Group of the subject (`Groups: nil`), with every profile, and the associations of those Groups; selected context: associations and profiles only |
| BusinessProfile capability | `privileged_settings.manage`: Organization owner. Otherwise an Organization owner or admin (the `Y` cell: any profile, unlinked ones included, independent of the selection), or `decide` allows the current association of some candidate Group (`own`, `linked`, `linked+flag`) | as for a Group capability |

The fact options depend on the catalog's shape, never on the subject's role, and add at most four statements to the
snapshot (E-MCP-4). Candidate Groups are the selected Group in selected context, otherwise every Group of the subject.
The Organization owner/admin path is role-level. It asserts no existing Group or BusinessProfile target, so its names
are listed even in an otherwise empty Organization. Privileged settings stay owner-only. The Group-derived paths need
their current candidate, link and policy facts. Eligibility never proves that a target exists, grants an effect or
replaces the exact-target check: an unknown or foreign target keeps its non-disclosing refusal at call time.
When the context check fails, the snapshot records `access.denied` (operation `mcp.request`) and the caller gets
`context_denied`. A store failure is `unavailable`. Production declares no capability-mapped tool, so its eligibility
snapshot reads exactly what `AuthorizeConnection` reads. Evidence: `TestToolEligibility` compares every capability,
subject, context, flag and link state with `decide` over the oracle world. `TestMCPOrdering` O-MCP-7 shows a
snapshot taken during a held switch, or spanning its commit, sees one (selection, facts) pair. O-MCP-8 shows a change
committed after the names are computed governs the next request and the tools' own checks.

### Guarded storage operations

`NewServiceWithObjects` adds five storage operations over the owned object store
([OBJECT-STORAGE.md](OBJECT-STORAGE.md)): `PutObject`, `ObjectMetadata`, `GetObject`, `SignedGetURL` and `DeleteObject`.
They take only the bound `TenantContext` and canonical UUID selectors. Reads reuse `organization.read` and `group.read`
unchanged; writes use a separate closed storage-action mapping (the eleven capabilities and the matrix above are
unchanged, and `organization.manage_*` is not borrowed):

| Action | Organization object | Group object (owning Group g) |
|---|---|---|
| read (Metadata, Get, Sign) | organization.read | group.read on g |
| put | org_owner, org_admin | org_owner, org_admin, or any current role in g |
| delete | org_owner, org_admin | org_owner, org_admin, group_owner/group_admin of g; group_member of g only as the stored creator |

The gates run in the order above: the Organization binding, the owning Group in the active Organization, the selected
Group restricting Group targets (Organization objects stay reachable in selected-Group context), then the mapping.
Creator equality is evaluated after the role gate and never grants scope. Every storage transaction discovers the
object's immutable scope under row-level security first and then reads only the owning (and selected) Group's facts,
never every binding. A missing, foreign, inaccessible, outside-context or non-ready object is `not_found`; a readable
object whose action is denied is `forbidden`. Storage calls write `storage.operation`, not `access.decision`.

## Ordering and failure contract

**Guarded write.** One READ COMMITTED read-write transaction on one connection, in this order:

1. `users` row, `FOR SHARE` with `status = 'active'` (first, matching `DisableUser`'s user-then-sessions order).
2. `sessions` row, `FOR SHARE` with the id, user, revocation and expiry predicates at the service clock.
3. `organizations` row, `FOR SHARE`; `FOR NO KEY UPDATE` for `SetOrganizationPolicy`, taken up front so that a shared lock
   is never upgraded.
4. The subject's Organization binding, `FOR SHARE`.
5. The subject's Group bindings for the candidate set, `FOR SHARE`, in `group_id` order. The candidates are the target
   Group for Group writes, the selected Group for any write in selected-Group context, and every subject Group in the
   Organization for an Organization-only profile write.
6. Group rows, in ascending `id` order, each with its own strength (since version 14): the other candidates below the update
   target `FOR SHARE`, the target `FOR NO KEY UPDATE`, then the other candidates above it `FOR SHARE`. Without a target
   one ascending statement; shapes with only a target are unchanged. Every multi-Group lock therefore follows one order
   and two completions whose target and candidate Groups cross cannot deadlock (O-ONB-D); the combined shape costs at
   most one extra statement.
7. The associations of those Groups: a plain read in a new statement.
8. The target read and the decision.
9. The write.
10. `COMMIT`.

Steps 4 and 5 read `app.role_bindings` directly, with exactly the projections' filters (`scope_type` and
`source = 'membership'`), because a row lock needs `UPDATE` on a column of the relation it names and the projections are
SELECT-only (design option A). They return the same columns and rows as the projections through the existing
`UPDATE (role)` grant. Snapshot reads stay on the projections. These three locking statements are the only direct reads
of `role_bindings`, pinned by name and exact text in `core/internal/tenancy/boundary_test.go`.

**Linearization point:** the completion of step 7. Every relied-on mutable fact is row-locked by its own statement until
the transaction ends. A statement that waited on a concurrent writer re-evaluates its predicates against the committed
row, so a revoked or expired session, a disabled user, or a deleted or changed binding is excluded, and an unlink
committed before step 6 is invisible to step 7. The decision uses only rows returned by steps 1–7; a concurrent grant or
link is never relied on. A conflicting writer therefore either commits after the guarded write, or the guarded write
observes the change and denies. No write commits after a revocation it did not see.

**Guarded read and Resolve.** One REPEATABLE READ READ ONLY transaction without row locks. All facts and the returned row
come from one snapshot, so a read that starts after a revocation commits observes it.

**Accepted writers.** No writer adopts a lock convention, and every signature is unchanged. Logout, `DisableUser`,
sign-in rotation, renewal and provider-logout bookkeeping conflict with steps 1–2; Organization and Group role changes
and removals with steps 4–5; policy updates with step 3. `UnlinkGroupBusinessProfile` now locks its Group row
(`FOR NO KEY UPDATE`) before its DELETE, so it conflicts with step 6. Grants, `LinkGroupBusinessProfile`, `CreateGroup`
and `CreateBusinessProfile` take only foreign-key `FOR KEY SHARE` locks, which are compatible with every lock of steps
1–7: new authority is never relied on. The exception is a slug change that changes the slug (below): once its step-9
UPDATE has run, `LinkGroupBusinessProfile` and `GrantGroupRole` inserts that reference that Group wait until the guarded
write ends.

**Locks.** Steps 1–7 take shared locks plus at most two exclusive ones, the Organization row for a policy update (step 3)
and one target Group row (step 6), in one global order, and a shared lock is never upgraded. Step 9 then locks the row
it writes:

- A policy update writes the Organization row it already holds `FOR NO KEY UPDATE`.
- A Group rename, and a slug change that keeps the slug, write the Group row already held `FOR NO KEY UPDATE`.
- A slug change that changes the slug takes `FOR UPDATE` on the Group row, a stronger mode than step 6's. `slug` belongs
  to the unique key `groups_organization_slug_key`, and PostgreSQL treats the columns of such a unique index as key
  columns. `FOR UPDATE` conflicts with the foreign-key `FOR KEY SHARE` locks that `LinkGroupBusinessProfile` and
  `GrantGroupRole` inserts take on that Group: step 9 waits for such an insert in flight, and later inserts wait for the
  guarded write until it ends. This is exactly the behaviour of the accepted raw `ChangeGroupSlug`. Those inserts take
  that lock as their last contended lock and never wait for a lock the guarded write holds, and a trace of these writers
  found no new wait cycle. The footprint evidence (E-PG-2) changes a slug to its current value, so it did not measure
  this upgrade; the statement rests on PostgreSQL's documented lock semantics for updates of unique-key columns.
- A profile update takes an exclusive (`FOR NO KEY UPDATE`) row lock on the profile.

**Deadlocks.** One known cycle remains: a sign-in rotation that supersedes the guard's own session while the provider
display name changed takes the sessions row and then updates the users row, the reverse of steps 1–2. PostgreSQL aborts
one side with `40P01`; an aborted guard returns `unavailable` with state unchanged, and an aborted rotation issues no
session (B-POLICY-ORDERING O-15). An ordinary `40P01` or `40001` error (severity ERROR) is a definitive abort, never
retried.

**Explicit Group authority (version 17): the raw Organization departure is not in the facade's lock order.** The authority guard's step (3)
deletes the departing subject's Group bindings (ascending `group_id`) inside the Organization-binding DELETE, so the
facade departure (`RemoveMember`, `LeaveScope`, `TransferOwnership(…, none)`, retirement) keeps the order above: users,
session, Organization row `FOR SHARE`, the subject's Organization binding, then the guard's nested Group rows, and the
enrollment departure trigger's (version 16) Organization `FOR NO KEY UPDATE` last. A **raw** runtime DELETE of an Organization binding has no
facade steps: it locks that binding (and its Group rows) first and only then requests the Organization row in the enrollment
departure trigger (for a raw departure of an Organization owner already in guard step (4); a self-departing Group
owner also locks that `groups` row `FOR NO KEY UPDATE` in the nested step (4)). Any concurrent facade operation that
holds the Organization row `FOR SHARE` (fact step) and next locks that subject's Organization binding — the subject's
own claim, a claim the subject issued (the claim body locks the issuer's bindings), or a `ChangeMemberRole`,
`RemoveMember` or `TransferOwnership` naming the subject — can therefore form a wait cycle with it; PostgreSQL aborts one
side with `40P01` (or `40001`). The raw binding-then-Organization edge already existed with enrollment (version 16) for a subject without
Group bindings; version 17 extends it to subjects with Group bindings and adds the nested Group row locks, so it is
recorded as a new wait edge here.
The cascade locks each of the subject's Group bindings `FOR NO KEY UPDATE`, ascending `group_id`, **before** it reads
the binding's role: a concurrent promotion of that binding to Group owner either commits
first, so the locked read returns `group_owner` and the departure refuses `owner_subject` with nothing removed, or it
waits for the departure and then finds no row (B-RLS-ORDERING OR-7). These are the same rows the nested DELETEs already
locked in the same order, so no further wait edge is added (the locks are only taken somewhat earlier, up to the
cursor's prefetch, and an `owner_subject` refusal can now wait for an in-flight promotion before refusing);
`role_bindings_update` admits that row lock (its `USING`) at
trigger depth ≥ 1 only, and its `WITH CHECK` still requires the override, so no Group binding can be changed through it.
The abort is the existing definitive class-40 outcome: the aborted transaction leaves no partial Group or Organization
binding and no partial audit row, the facade answers once with `unavailable` (reason `aborted`), nothing retries it, and
the facade's own step order is unchanged (B-POLICY-ORDERING O-MEM-4c forces the interleaving; its first run observed the
raw DELETE as the `40P01` victim and the claim committing).

**Label rows (version 20).** Label operations keep steps 1–7 and add one lock after step 7, before the write: the
label definition row, `FOR NO KEY UPDATE` for definition update and delete and `FOR SHARE` for assign and unassign. No
existing writer locks label rows and every label operation locks Group rows before label rows, so no new wait cycle is
added; two assigns of one label to two Groups hold the definition `FOR SHARE` (compatible). Two exceptions to the
ordinary strengths:

- `CreateLabelDefinition` in the Organization scope takes the Organization row `FOR NO KEY UPDATE` **up front at step 3**
  (the `SetOrganizationPolicy` rule, through `policyWriteQuery`), so a shared lock is never upgraded; in the Group scope
  it locks the owning Group `FOR NO KEY UPDATE` at step 6. The `label_definitions_bound` trigger then re-locks the same
  scope row at the same strength before counting (a re-lock, not a conversion). Two Organization-scope creates serialize
  at step 3; Organization-scope creation therefore also serializes with policy writes and enrollment approvals of the
  same Organization.
- `DeleteLabelDefinition` reads the definition under the same facts, then refuses an Organization-owned definition when
  the trusted context has a selected Group (`organization_context_required`, before any write: the selected-Group
  assignment DELETE policy could not remove other Groups' assignments and the foreign-key refusal would reveal them).
  Otherwise it takes the definition `FOR NO KEY UPDATE` without locking Group rows, deletes its assignments and then the
  definition in two statements; an in-flight assign holding the definition `FOR SHARE` commits first and its row is
  removed by the fresh DELETE statement, and a later assign waits and then finds no definition (`not_found`).

**What a caller can conclude.** The internal reason distinguishes these cases in the log only; a caller sees the external
code.

| Failure | Returned (internal reason) | Effect |
|---|---|---|
| Validation, deny, statement error, cancellation before `COMMIT` is sent | the pinned code; `unavailable` for failures | nothing changed |
| `COMMIT` never sent (the driver proves it unsent) | `unavailable` (`canceled` or `store_unavailable`) | nothing changed |
| `COMMIT` answered with ROLLBACK, or rejected by an ordinary command-severity ERROR | `unavailable` (`aborted`) | nothing changed: the server rolled the transaction back |
| After `COMMIT` may have been sent: a session-ending FATAL (for example 57P01) or a PANIC (57P02), a server error with a missing or unrecognized severity, or a lost reply | `unavailable` (`commit_outcome_unknown`) | unknown: the write may be durable |

A server error alone is never taken as a confirmed rollback. Only the nonlocalized severity ERROR (or, when the server
sends none, the localized severity ERROR) establishes one.

A caller cannot tell these `unavailable` cases apart. Adapters must therefore treat **any** `unavailable` from a write as
possibly durable, and reconcile through a guarded read, which is subject to current authority. Every other code is
returned before `COMMIT` and means nothing changed.

There is no automatic retry and no idempotency key. The four writes are absolute assignments (name, slug, full context,
boolean), so re-applying a request cannot accumulate effects, and setting the current value is a successful no-op. A
caller's retry is a new operation: it is authorized at retry time and may be denied after a revocation even if the first
attempt committed, and under last-commit-wins it may overwrite an intervening write.

**Not claimed:** database-wide isolation or row-level security (those are [DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)'s); protection against holders of the runtime
credential running arbitrary SQL; serializable cross-table guarantees for unguarded raw model calls; ordering of reads
beyond their snapshot; any performance target, query plan or bounded lock-wait time.

## Query and lock footprint

Workload limits are unknown and no latency is promised; cardinalities, transaction deadlines and revocation latency
need revisiting before public tenant adapters.

- **Statements.** Each call runs a constant number of statements; every set is read by one `= ANY` statement, never one
  per Group. A guarded read runs at most 8 statements; a guarded write at most 9 plus `BEGIN` and `COMMIT`.
- **Growth.** Organization-only profile operations read 5 + 3·n rows and, for a write, lock 4 + 2·n rows, where n is the
  subject's Group bindings in the Organization. Every other operation is constant in Group count.
- **Hot rows.** Every guarded write shares the Organization row; a subject's writes share its user and session rows; and
  Organization-only profile writes share the rows of the subject's Groups. Row locks dirty the heap page and are
  WAL-logged, and concurrent shared lockers use a MultiXact. Guarded reads take no row locks.
- **Waiting.** A writer that conflicts with in-flight guarded writes waits for them. No `lock_timeout`, `statement_timeout`
  or idle-transaction timeout is configured; the bounds are the caller's deadline and connection loss.
- **Evidence (B-POLICY-IT and B-POLICY-ORDERING; environment-labelled observations, not benchmarks):**
  - E-PG-1, plan summaries: the locking forms of statements 4 and 5 (the three direct `role_bindings` locks) use the
    subject's index path, with no Organization-wide membership scan. The snapshot forms through the projections, used by
    `Resolve` and guarded reads, are not measured; the same paths are inferred, because the projections are simple
    filters over `role_bindings`.
  - On the small fixture tables (6 to 264 rows), PostgreSQL chose sequential scans for statements 2 (sessions),
    3 (organizations), 6a and 6b (groups), 7 (associations), 8 (profiles) and both UPDATEs of statement 9 (groups and
    profiles), each with one to four shared buffer hits and no reads. These are observations, not evidence of efficient
    plans or of production scalability, and no index, query or lock change follows from them.
  - E-PG-2, the locked-row inventory: the rows locked are exactly the rows the locking statements return.
  - E-PG-3, lock queues: a compatible shared locker can proceed past a writer already queued for the same row (a policy
    write behind the Organization row, a removal behind the subject's own guard). Ordering safety holds, but there is no
    fairness or maximum-wait guarantee.

No index, query or lock-strategy change follows from these observations. New broad membership scans, unexpected
locking behaviour or a latency requirement return to review.

## Errors, reasons and the event

`*access.Error{Code, Field}` prints `access: <code>` and never IDs, names, roles, policy values or driver text.

| Code | When |
|---|---|
| `unauthenticated` | absent, invalid, expired or revoked session; disabled user |
| `context_denied` | the scope or binding check, a nested `BeginRequest`, the context gate, or the context's Organization membership or Group selectability no longer holds |
| `not_found` | an unknown, foreign or wrong-kind target, a target outside the selected Group, or a target the subject cannot read |
| `forbidden` | the target is readable by the subject but the capability is denied (including the flag off) |
| `invalid_input` | a malformed selector or payload; `Field` carries the tenancy code (for example `invalid_slug`) |
| `conflict` | `slug_conflict`; Group onboarding: Group creation with `Field` `slug`, a completed Group with `Field` `association_exists` |
| `unavailable` | identity or the database unavailable, a definitive abort, cancellation, or a lost commit reply |
| `internal` | an unexpected tenancy `internal` error |

`forbidden` versus `not_found` uses the read capability of the same kind on the same facts, so an inaccessible target is
indistinguishable from an unknown one.

**Internal reasons** (closed; logged only): `allowed`, `context_established`, `session_invalid`, `selector_invalid`;
`context_forged`, `context_scope_missing`, `context_scope_ended`, `context_scope_mismatch`, `context_subject_mismatch`,
`context_scope_nested`; `organization_not_permitted`, `group_not_selectable`, `target_not_found`,
`target_kind_mismatch`, `target_outside_context`; `capability_unknown`, `role_not_permitted`, `no_linked_path`,
`edit_policy_disabled`, `fact_inconsistent`; `input_invalid`, `write_conflict`, `store_unavailable`, `aborted`,
`canceled`, `commit_outcome_unknown`.

**Event.** Each call writes one `access.decision` record, catalogued in `events.json`:

- `operation`: `access.begin_request` and `access.resolve` (the context operations), or one of `organization.get`,
  `organization.get_policy`, `organization.set_policy`, `group.get`, `group.rename`, `group.change_slug`,
  `business_profile.get` and `business_profile.update`;
- `capability`: forbidden for a context operation; required for a guarded operation, and exactly its mapping above;
- `outcome`: `established` (info) for a context operation's success, `allow` (info) for a guarded operation's, `deny`
  (warn), or `failed` (error: `unavailable` or `internal`);
- `reason`: one internal reason; `allowed` only with `allow`, `context_established` only with `established`;
- `request_id` and `trace_id` when the request context carries them.

Records never contain UUIDs, names, slugs, roles, profile or policy values, tokens, claims or driver text. Guarded writes
do not also emit `tenancy.operation`. This is an operational log, not the audit ledger: the durable record is
`audit.events` ([AUDIT.md](AUDIT.md)). A recordable denial (a policy refusal after a valid binding check:
`context_denied` with `organization_not_permitted` or `group_not_selectable`, or `not_found`/`forbidden`) is also written
as one `access.denied` event in a short audit transaction after the denied transaction released its connection and
before this record; its target only when `forbidden`, its Organization only when the Organization gate passed. A
recording failure leaves the decision unchanged and writes one `audit.record_failed`. `ReadAuditEvents` is the guarded
`audit.read` operation (the current `org_owner` only). Request and trace correlation is stored only as bounded
diagnostic audit correlation; it is never compared or used in a decision.

## Raw primitives and the import boundary

The tenancy `Service` methods, `ReadFacts` and the transaction-scoped writers (`RenameGroupTx`, `ChangeGroupSlugTx`,
`UpdateBusinessProfileTx`, `SetOrganizationPolicyTx`) authorize nothing. They remain **trusted-only**, for in-process
bootstrap, tests and future integration. The guarantee is the production import boundary, checked by
U-ACCESS-BOUNDARY over every non-test Go file under `core/`: only `core/internal/tenancy/access` may import
`core/internal/tenancy`, and only `core/internal/mcp` imports `access`. B-MCP-SOURCE pins that importer set and
keeps both packages out of the migrate binary; the server links them only through the MCP handler, and B-TENANCY-HTTP
proves that no tenancy route exists. Every adapter calls the guarded facade, never the raw primitives. Later importers
of `access`, each imported only by `cmd/server` (`TestProductionImporters`): `core/internal/domains/api`,
`core/internal/branding/api` and `core/internal/onboarding/api`. B-TENANCY-HTTP now pins the two served
onboarding core paths as closed refusals with no model payload and keeps probing the absence of every other tenancy route.

## Row-level security handoff

*Delivered by version 6 ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)): every locking statement below is admitted by the
UPDATE policies' existing-row sets, the role lock grants no write (WITH CHECK `false`), and five log-only reasons shift as
pinned there (an unreadable or out-of-context target becomes `target_not_found`; an unselectable Group context becomes
`organization_not_permitted`). External codes are unchanged.*

Version 6 enables and forces row-level security. This is the inventory it started from.

- **Tenant relations:** `app.organizations` (including `group_admins_edit_business_profiles`), `app.groups`,
  `app.role_bindings` and its projections `app.organization_memberships` and `app.group_memberships`,
  `app.business_profiles`, `app.group_business_profiles` and `app.business_profile_access_paths`. Their tenant columns
  and later obligations are listed in [TENANCY.md](TENANCY.md#what-later-stages-must-secure).
- **Global identity tables read inside tenant transactions:** `app.users` and `app.sessions` (session liveness, locked
  `FOR SHARE` in writes). They have no tenant column; [DATABASE-ISOLATION.md](DATABASE-ISOLATION.md) classifies them.
- **Statements** (named constants):
  - `ReadFacts` in `core/internal/tenancy/store.go`: `sqlFactOrganization` with `sqlFactOrganizationShare` and
    `sqlFactOrganizationExclusive`; `sqlFactOrganizationBinding` and `sqlFactOrganizationBindingLock`;
    `sqlFactGroupBindings` and `sqlFactGroupBindingsLock`; `sqlFactAllGroupBindings` and `sqlFactAllGroupBindingsLock`;
    `sqlSelectGroup` and `sqlFactTargetGroupLock`; `sqlFactGroups` and `sqlFactGroupsLock`; `sqlFactAssociations`;
    `sqlFactAllProfiles`; `sqlSelectBusinessProfile`.
  - The transaction-scoped writers: `sqlSetOrganizationPolicy`, `sqlRenameGroup`, `sqlChangeSlug` and
    `sqlUpdateBusinessProfile`; and the raw `UnlinkGroupBusinessProfile`'s new `sqlLockGroupForUnlink`.
  - The identity checks in `core/internal/identity/store.go`: `sqlLiveUser` and `sqlLiveSession`, and their locking forms
    `sqlLockLiveUser` and `sqlLockLiveSession`. `BeginRequest` uses identity's accepted session resolve and renewal in
    their own transactions, outside the tenant transaction.

  Every tenant statement names its Organization (`organization_id = $1`, or `id = $1` for `organizations`).
- **Row locks under row-level security.** PostgreSQL applies UPDATE policies as well as SELECT policies to rows read with
  `FOR SHARE` or `FOR NO KEY UPDATE`. The row-level security policies must admit every locking statement above (including the three
  direct `role_bindings` locks) in the active context, or guarded writes will deny.
- **One transaction-owning call site.** `inTx` in `core/internal/tenancy/access/tx.go` begins, commits and rolls back
  every guarded operation and `Resolve`. Row-level security's transaction-local context is initialized there (`set_config(…, true)`),
  from the bound `TenantContext`, before the first statement.
- **Raw Service methods** stay trusted-only and need their own transaction-local context if they are ever used by an
  untrusted path.

## Validation

The policy checks run in the `migrations` suite of the validation harness, which only full runs select:

```sh
infra/validation/run.sh --base <commit> --allow-file <scope-allow-list> migrations
```

- B-POLICY-CATALOG: the version-5 catalog, the column and its exact grants, the option A locks and unchanged projection
  ACLs, and the database refusal of malformed flag values.
- B-POLICY-MIGRATION: empty, populated, repeated, restarted and down/up migrations to version 5.
- B-POLICY-IT: `TestAccessPolicy` (context binding and revalidation, the matrix on database-loaded facts, profiles,
  the flag, the facade's failure modes, revocation, MFA and the log record) and `TestAccessFootprint` (E-PG-1, E-PG-2).
- B-POLICY-ORDERING: `TestAccessOrdering`, deterministic interleavings without sleeps (O-1..O-15, E-PG-3; Group onboarding adds
  O-ONB-1..3 with 3s/3p for revocations racing a completion, O-ONB-D for the ascending Group lock order and O-ONB-E
  for an UPDATEd profile after a concurrent revocation).
- Group onboarding: B-ONBOARDING-CATALOG and B-ONBOARDING-IT ([TENANCY.md](TENANCY.md#commands)); `TestAccessFootprint` gains the
  "E-PG-2 onboarding writes" observations (Group creation, onboarding completion and list cost).

The unit tests (`core/internal/tenancy/access`: U-POL-MATRIX with an independently written matrix, U-POL-GATE,
U-POL-SCOPE, U-POL-MFA, U-CTX-1..9 and U-ACCESS-BOUNDARY; and the tenancy U-BOUNDARY) run in the core check image.
`bash infra/ci/validate.sh --profile full` (README `platform:ci`) includes all of them.

## Isolation attack suite

`TestIsolationAttacks` (`core/internal/tenancy/access/attack_integration_test.go`) composes the single-boundary
guarantees above, [DATABASE-ISOLATION.md](DATABASE-ISOLATION.md) and [OBJECT-STORAGE.md](OBJECT-STORAGE.md) into
integrated attack chains. Each chain crosses at least two of four surfaces: the session/context boundary, the guarded
facade, raw runtime SQL on the attested runtime login, and the real provider (private MinIO, scoped credential).
Sessions are real identity sign-ins. The literal expectations come from an independent oracle,
`attack_expected_test.go`, written only from the contracts with the outcomes withheld. A disagreement is reported
for a ruling and the oracle is never edited to match. Single-boundary cases that already exist (the policy, row-level security and storage tests and
the two accepted oracles) are cited in the oracle's `Inherited` fields, not repeated. The suite adds ten composition
steps, and Group onboarding two more (B2, C3; oracle r2).

**World.** Organizations A (Group-inherited profile management on) and B. In A, sibling Groups g1 and g2 share
profile p1, g3 has p2, and p3 is unlinked. In B, gB reuses g1's slug, and pB is its profile. Roles exercised by a
step: aOwner, aAdmin and admin2 (org_admin), gAdmin, gMember, gMember2 and gMember3 (roles in g1), dual
(group_admin of g1 and g2), multi (one global identity: group_admin of g3 in A, org_owner of B), and outsider (a
session with no binding). aMember and gOwner are in the world, and their behaviour is cited from the inherited tests.

| Step | Composition | Surfaces |
|---|---|---|
| A1 | one request scope, two tenant contexts (A/g3 and B): each context's targets in the other are `not_found`; the two runtime SQL shapes see exactly their own rows; a cross-tenant UPDATE affects 0 rows | context, facade, provider, SQL |
| A2 | storage writes with an ended scope (B4), another scope of the same credential (B5) or another subject's scope: `context_denied` before any transaction | context, provider |
| B1 | the shared profile is readable and editable through g1, and gives no path to sibling g2 (Group, its objects, a Put, its link and bindings in SQL) | facade, provider, SQL |
| B2 (onboarding) | a member creates its own Group and reads its onboarding state, but cannot complete it with a profile it has no owner/admin path to (`forbidden` for the member-only path, `not_found` otherwise); the runtime association INSERT and direct or forged owner-binding INSERTs are 42501 | facade, SQL |
| C1 | guarded member uploads (the stored creator is the subject): own delete `deleted`, another member's `forbidden`, the Group admin's delete `deleted` | facade, provider |
| C2 | Organization removal leaves a Group-only binding: the retained context is `context_denied` for storage, and the runtime shape sees no tenant row, its own object included | context, provider, SQL |
| C3 (onboarding) | a Group owner's retained context after its Organization membership is removed: both completions `context_denied`, the eligibility reader false and the association INSERT 42501; a completed Group's unlink DELETE is 42501 | context, facade, SQL |
| D1 | one retained scope across three committed changes (unlink, downgrade, removal): profile and storage authority change together | context, facade, provider |
| D2 | one pooled backend serves storage and facade calls, including an injected statement failure: the same PID, empty settings, exact results | facade, provider, SQL |
| E1 | an Organization role downgrade between the landed bytes and the publish: `forbidden`, compensated, never readable; the identical staged control publishes | context, provider |
| E2 | an issued URL across a paused delete: 200 while logically deleted, new issuance `not_found`, non-2xx after the physical removal and before expiry | facade, provider |
| E3 | a Group removal between the byte removal and the tombstone: `deletion_pending`; an authorized repeat delete completes it | context, provider |

**Denial evidence.** A pre-provider denial counts only when all four hold: the oracle's exact external code; its
paired authorized control checked earlier in agreement; zero provider calls; and unchanged Head presence,
stored-object rows and tenancy rows. E1 and E3 are staged. They assert the accepted staged effects instead (the
pending row, compensation and absence; the deleting row and absent bytes) and a working authorized control. A
missing prerequisite fails the test, so infrastructure failure never passes as a denial.

**Trust limits.** The suite proves no new guarantee beyond these documents. The recorded residuals stay as they are
and are not new defects:
- the DATABASE-ISOLATION self-scope residual and the trust limit (a holder of the runtime credential can set any context);
- OBJECT-STORAGE X1–X7 (for example, an issued URL works until expiry or byte removal, and provider-served downloads
  are not verified by core);
- the provider credential, which is scoped to the bucket, not to an Organization.

It is not a penetration test or a compliance attestation.

**Run.** B-OBJ-IT runs it as a second invocation after `TestObjects`, with its own log. The check pins 17 passing
subtests and `ISO-ORACLE {"cases":82,"disagreements":0}`, and runs the access-log canary checks on both logs.

## Availability: denied exclusive-write requests

Two kinds of guarded write request an exclusive row lock before the capability decision:

- `SetOrganizationPolicy` locks the Organization row `FOR NO KEY UPDATE` at step 3, before the subject's
  Organization binding is read and the decision is made.
- `RenameGroup` and `ChangeGroupSlug` lock the target Group row `FOR NO KEY UPDATE` at step 6, before the decision,
  when the context is Organization-only or the target is the selected Group.

A live subject whose context resolves in the Organization but who lacks the capability (for example an `org_member`)
therefore holds such a lock until its request is denied and the transaction ends, and can contend with authorized
writes:

- a denied policy write delays every guarded write in that Organization, because they all share the Organization row,
  and Organization renames;
- a denied Group write delays writes that lock that Group row: guarded writes whose candidate set includes it,
  unlinks, renames and slug changes.

The denial changes nothing; no unauthorized mutation is possible. Ordinary guarded reads take no row locks and are not
blocked by these locks.

The reviewed locked-facts decision and the lock order stay as they are. Before any adapter exposes these operations,
it must provide per-subject admission or rate limits and a bounded request and transaction lifetime. Neither exists
for these operations: they have no adapter, and no `lock_timeout`, `statement_timeout` or idle-transaction timeout is
configured. No numeric bound is claimed.

### Onboarding completions (the first adapter-exposed case)

`CompleteGroupOnboardingWithNewProfile` and `CompleteGroupOnboardingWithProfile` lock the target Group
`FOR NO KEY UPDATE` before the decision, and the onboarding HTTP adapter (`POST /internal/v1/groups/{id}/business-profile`,
reached through the BFF) exposes them. They carry the required bounds; the decision and lock order are unchanged:

- **Admission (adapter, process-local).** After the unchanged authenticated `begin` (which already does its own
  session and context database work) and before either completion's transaction or Group lock, the adapter admits at
  most one completion per verified canonical subject (two sessions of one user share the entry) and at most 256 in
  all. Admission never waits; a refused request is the existing closed `onboarding_unavailable` 503. The entry is
  released on every outcome. Admission grants nothing: the facade still decides on the locked facts. The state is per
  core process, with no shared limiter, so with N serving core processes one subject may have up to N completions in
  flight, each bounded as below.
- **Request and body read (adapter).** The completion route reads its body and runs under a 10-second request
  deadline (the connection read deadline and the request context; an earlier caller deadline is kept). Other routes
  are unchanged.
- **Transaction (facade).** Each completion's transaction, BEGIN through COMMIT, runs under a 5-second decision
  budget that is also the context of the completion's own writes and their failure classification, and its deferred
  rollback under a 5-second cleanup budget. A recordable denial keeps the existing detached audit budgets (2 seconds to
  record, 2 seconds of rollback cleanup). The worst-case server lifetime of one admitted completion is therefore
  about 19 seconds (10 seconds of read and decision, at most 5 of them the decision, then up to 5 seconds of rollback
  and 2 + 2 seconds of denial audit), all while its admission entry is held.
- **Outcomes unchanged.** A deadline before COMMIT rolls back and is the closed unavailable failure. A deadline or
  failure during COMMIT keeps the existing commit classification: the outcome may be durable, so the answer is
  unavailable and the web page states that the result is unknown. There is no retry and no new error code.

The other exclusive-lock writes above (`SetOrganizationPolicy`, `RenameGroup`, `ChangeGroupSlug`) have no adapter and
keep the rule as stated. No database-wide `lock_timeout`, `statement_timeout` or idle-transaction timeout is added.


### Branding capability

Current `org_owner`/`org_admin` in the trusted Organization may update/reset branding and manage its logo through the
guarded paths; Organization membership/Group roles alone do not qualify. Existing scope/RLS/session/CSRF/audit and
object-delete restrictions remain.

The thirteenth capability
`organization.manage_branding` (owner and admin of the request's Organization) guards
GetBranding, UpdateBranding, ResetBranding, SetBrandingLogo, RemoveBrandingLogo and the
unreferenced-upload delete; denials are audited `access.denied` as for the other
capabilities. The public projection and the logo route need no capability: they answer
only the public fields of the Organization a trusted host resolves to.


### Ownership continuity

Every existing Organization/Group needs at least one owner; multiple owners are
allowed. Creator ownership is atomic and uses its single owner RoleBinding,
which includes admin powers. Last-owner departure/demotion/removal or ordinary
account retirement requires promotion of an eligible admin or atomic transfer;
a server-side count/eligibility result is rechecked under guarded transaction
and database protection. Ordinary admin authority cannot grant ownership to self.
Emergency security suspension remains immediate and keeps the assigned ownership
binding without usable access; all-owner unavailability needs restricted verified
recovery. Group onboarding owns initial Group creation; membership administration (version 15) owns owner
administration and comprehensive enforcement. The raw primitive role grants alone do not enforce this policy.
The guarded operations and the version-15 authority guard implement it
([Guarded membership, ownership and invitation operations](#guarded-membership-ownership-and-invitation-operations));
the database-qualification caveat of that section's evidence applies.
