# Tenancy model

This page describes the Organization/Group tenancy model and its single RoleBinding authority store (migration 00003), and the
Organization-owned BusinessProfile with its Group associations and derived access facts (migration 00004). It is a model
foundation, not an authorization system. The trusted request context and the policy evaluated over these facts are
described in [AUTHORIZATION.md](AUTHORIZATION.md); row-level security is in place from version 6
([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)), and every public surface arrives with later
stages (see [What later stages must secure](#what-later-stages-must-secure)). The original canonical contract was
the generator Skill's tenancy reference (retired with the generator and not part of this distribution); this foundation's accepted design supersedes that reference's
BusinessEntity grant examples (no BusinessProfile grant exists; access is derived, below).

| Part | Where |
|---|---|
| Schema, grants and projections | `core/migrations/00003_tenancy.sql` (version 3) and `core/migrations/00004_business_profiles.sql` (version 4); see [MIGRATIONS.md](MIGRATIONS.md) |
| Internal model primitives | `core/internal/tenancy` (`Service`), for trusted in-process callers and tests only |
| Policy flag, ordered fact reads and guarded access | `core/migrations/00005_organization_policy.sql` (version 5), `core/internal/tenancy/facts.go` and `core/internal/tenancy/access`; see [AUTHORIZATION.md](AUTHORIZATION.md) |
| Real-PostgreSQL tests | `core/internal/migrate/{tenancy,business_profile}_integration_test.go` and `core/internal/tenancy/{integration,business_profile_integration}_test.go` |
| Group creation and onboarding | `core/migrations/00014_group_onboarding.sql` (version 14), `core/internal/tenancy/onboarding.go`, `core/internal/tenancy/access/onboarding.go`, `core/internal/onboarding/api`; see [Group onboarding](#group-onboarding) |

## Model

- **Organization** (`app.organizations`) is the tenant and security boundary. It has a database-generated, nonzero
  UUID `id`, a display `name` (1–200 code points, trimmed, no control characters; not unique) and `created_at`.
  Since version 5 it also carries the one policy flag, `group_admins_edit_business_profiles` (`boolean NOT NULL DEFAULT
  true`; [AUTHORIZATION.md](AUTHORIZATION.md)). `Tenant` is a security concept, not a second persisted entity.
- **Group** (`app.groups`) belongs to exactly one Organization for its whole life (`organization_id`, non-null). It has
  a UUID, a display `name` and an editable `slug`. Renaming a Group never rewrites its slug; editing the slug changes no
  UUID, reference, binding or owning Organization. Moving a Group to another Organization is not supported.
- **RoleBinding** (`app.role_bindings`) is the only place a role is written: `subject_user_id` (an existing global user,
  `app.users.id`), `scope_type`, `scope_id`, `group_id`, `role`, `source` and its owning `organization_id`.
- **Memberships** are read-only projections of the bindings: `app.organization_memberships(user_id, organization_id,
  role)` and `app.group_memberships(user_id, organization_id, group_id, role)`. `organization_id` is the Group view's
  tenant anchor. Both are `security_invoker` views, so later row-level security on `role_bindings` also applies to them.
- **BusinessProfile** (`app.business_profiles`) is reusable business context owned by exactly one Organization for
  its whole life (`organization_id`, non-null, immutable). It has a database-generated, nonzero UUID and editable
  context: `display_name` (required), `legal_name`, `description`, `industry` and `website` (each optional; absent is
  NULL). It has no Group parent, no owner, creator or role column and no grant. A business present in two Organizations
  is two rows; equal names never merge.
- **GroupBusinessProfile** (`app.group_business_profiles`) associates a Group with a BusinessProfile of the **same**
  Organization (composite foreign keys to both). A Group has **at most one** association
  (`group_business_profiles_group_key UNIQUE (organization_id, group_id)`); one profile may be shared by several Groups
  of its Organization, and an update of the profile is seen through every one of them.
- **Access paths** (`app.business_profile_access_paths(user_id, organization_id, business_profile_id, group_id,
  group_role)`) are a read-only `security_invoker` projection: one row per current Group membership binding of a
  Group associated with the profile, with its raw Group role. They are facts, not capabilities (the shared policy classifies roles);
  a shared profile appears once per Group path, and a profile without associations has none. Organization authority is
  read from `organization_memberships`, never copied here.

Names, slugs, provider subjects, websites and hostnames are never keys or authority. A Group may have no BusinessProfile
association: the model allows at most one, not exactly one, and such a Group is **not** a completed onboarding. Since version 14 a
Group is complete exactly when its single association exists ([Group onboarding](#group-onboarding)).

### Roles and sources

| Scope | Roles | Source |
|---|---|---|
| Organization (`scope_type = 'organization'`, `scope_id = organization_id`, no `group_id`) | `org_owner`, `org_admin`, `org_member` | `membership` |
| Group (`scope_type = 'group'`, `scope_id = group_id`, same-Organization Group) | `group_owner`, `group_admin`, `group_member` | `membership` |

These are internal semantic codes, not client-facing labels; presentation names resolve through configuration in later
stages. One closed CHECK (`role_bindings_scope_consistent`) admits exactly these six combinations and nothing else. A
later scope or source (for example `invitation` or `system`) needs its own reviewed migration of that CHECK.

- There is at most one membership binding per subject and scope (`role_bindings_subject_scope_key`).
- An Organization role never synthesizes a Group membership, and a Group role never synthesizes an Organization role.
  No role is global.
- **Business Profile access is derived, not stored.** There is no BusinessProfile scope, role, grant, owner column or
  writable mirror, and `role_bindings` is unchanged by the BusinessProfile model. `business_profile_access_paths` derives the facts from
  current Group bindings and the Group's single association; linking, unlinking or editing a profile writes no binding,
  and a Group role change writes no profile row. The shared policy evaluates them with Organization authority.
- **Advisor, Principal and Associate** are future engagement functions, not roles. The CHECK rejects them like any other
  unsupported role.
- **Creator ownership.** `CreateGroup` takes an explicit, existing creator UUID from the trusted caller and, in one
  transaction, creates the Group and that creator's `group_owner` binding. Either both rows commit or neither does. It
  mints no Organization membership and consults no provider claim. There is no owner or `created_by` column anywhere:
  ownership exists only as that binding. Afterwards the binding is an ordinary row for the model primitives, which
  carry no actor and implement no ownership transfer, last-owner rule or "admin cannot change owners" policy (those belong to the shared policy and membership
  administration). The guarded creation path (version 14) has no creator parameter: the creator is always the request's verified context user,
  bound by a database trigger in the Group's own INSERT statement ([Group onboarding](#group-onboarding)).

## Constraint and privilege inventory

| Rule | Enforcement |
|---|---|
| Keys | `gen_random_uuid()` defaults; `*_id_nonzero` and `role_bindings_subject_nonzero` CHECKs refuse the all-zero UUID |
| Key immutability | Runtime column grants (first line) and BEFORE UPDATE triggers `*_keys_immutable` raising SQLSTATE 23000 (second line, for every role) |
| Group parent | `groups_organization_fkey`, `ON DELETE RESTRICT` |
| Binding subject | `role_bindings_subject_fkey` → `app.users (id)`, `ON DELETE RESTRICT` |
| Binding Organization | `role_bindings_organization_fkey`, `ON DELETE RESTRICT` |
| Same-Organization Group scope | composite `role_bindings_group_fkey (organization_id, group_id)` → `groups (organization_id, id)` |
| Typed scope, role and source | `role_bindings_scope_consistent` (two branches, above) |
| Slug grammar | `groups_slug_format`: 1–63 ASCII bytes, `^[0-9a-z]+(-[0-9a-z]+)*$`, column collation `C` |
| Reserved slugs | `groups_slug_reserved` (the set below) |
| Slug uniqueness | `groups_organization_slug_key UNIQUE (organization_id, slug)`; the same slug in two Organizations is allowed |
| One binding per subject and scope | `role_bindings_subject_scope_key` |
| BusinessProfile Organization | `business_profiles_organization_fkey`, `ON DELETE RESTRICT`; `business_profiles_organization_id_key UNIQUE (organization_id, id)` is the composite target |
| Same-Organization association | composite `group_business_profiles_group_fkey (organization_id, group_id)` → `groups (organization_id, id)` and `group_business_profiles_business_profile_fkey (organization_id, business_profile_id)` → `business_profiles (organization_id, id)`, plus `group_business_profiles_organization_fkey`; all `ON DELETE RESTRICT` |
| At most one profile per Group | `group_business_profiles_group_key UNIQUE (organization_id, group_id)`; nothing is unique on a profile alone, so a profile is shareable |
| BusinessProfile context | `business_profiles_{display_name,legal_name,description,industry}_valid` (bounds, trimmed, no control characters; TAB and LF allowed in the description) and `business_profiles_website_valid` (printable ASCII without space or backslash after `http://` or `https://`) |

No foreign key cascades: deleting a referenced user, Organization, Group or BusinessProfile directly fails while a
reference exists.

The runtime role (`<db>_runtime`) holds exactly:

| Object | Privileges |
|---|---|
| `organizations` | SELECT; INSERT (`name`); UPDATE (`name`, and since version 5 `group_admins_edit_business_profiles`) |
| `groups` | SELECT; INSERT (`organization_id`, `name`, `slug`); UPDATE (`name`, `slug`) |
| `role_bindings` | SELECT; INSERT (`organization_id`, `subject_user_id`, `scope_type`, `scope_id`, `group_id`, `role`, `source`); UPDATE (`role`); DELETE |
| `organization_memberships`, `group_memberships` | SELECT only |
| `business_profiles` | SELECT; INSERT (`organization_id` and the five context columns); UPDATE (the five context columns) |
| `group_business_profiles` | SELECT; INSERT (`organization_id`, `group_id`, `business_profile_id`); DELETE |
| `business_profile_access_paths` | SELECT only (a join view, so not updatable either) |

It has no DELETE on Organizations, Groups or BusinessProfiles, no UPDATE on associations, cannot supply or change any
key or `created_at`, has no TRUNCATE, REFERENCES or TRIGGER privilege and no EXECUTE on the trigger functions
(owner-only; triggers still fire). All objects are owned by `<db>_owner`; retention, the migrator without `SET ROLE` and
PUBLIC hold nothing. Each migration's postcondition verifies these exact table and column ACLs and refuses to commit
otherwise. The shared policy's three membership-fact locks (`FOR SHARE` on `role_bindings`) use the existing SELECT and UPDATE
(`role`) grants; no grant changed for them. These are the version-3/4 grants: version 6 narrowed them and revoked
every runtime INSERT and DELETE on the tenancy tables, and version 14 restores column-limited INSERT only on
`groups`, `role_bindings`, `business_profiles` and `group_business_profiles` under four INSERT policies
([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)).

## Group slugs

- **Stored and explicit slugs** are validated, never normalized: no case folding, percent-decoding, Unicode folding or
  transliteration. A non-canonical value is `invalid_slug`; a reserved value is `slug_reserved`.
- **Reserved set (24):** `account`, `admin`, `api`, `auth`, `billing`, `blueprint`, `fonts`, `health`, `healthz`,
  `internal`, `login`, `logout`, `mcp`, `members`, `message`, `messages`, `new`, `onboarding`, `readyz`, `rpc`,
  `settings`, `sign-in`, `sse`, `system`. Other top-level segments (for example `_next`, `favicon.ico`) cannot pass the
  grammar. Adding a reservation later is a reviewed migration.
- **Suggestion base** from a valid name: ASCII letters are lowercased; ASCII letters and digits are kept; every other
  character (non-ASCII included) separates; separator runs become one `-`; leading and trailing `-` are trimmed; the
  result is cut to 63 bytes and a trailing `-` is trimmed. `Acme Revenue` gives `acme-revenue`; `Café Ops` gives
  `caf-ops`; a name with no ASCII letter or digit has no base.
- **`SuggestGroupSlug`** is advisory and read-only: the base if it is neither reserved nor used in that Organization, else
  the first free `base-2` … `base-99` (each cut to fit 63 bytes), else `slug_unavailable`; a name without a base gives
  `slug_required`. It reserves nothing.
- **`CreateGroup` with an empty slug** uses the base only, with no suffix search, so a collision is reported
  deterministically (`slug_conflict`, `slug_reserved` or `slug_required`).
- Slug lookups always name the owning Organization. Neither a slug nor an availability result grants access.

## Internal operations

`Service` has exactly 29 methods: `CreateOrganization`, `GetOrganization`, `RenameOrganization`; `SuggestGroupSlug`,
`CreateGroup`, `GetGroup`, `GetGroupBySlug`, `ListGroups`, `RenameGroup`, `ChangeGroupSlug`; `GrantOrganizationRole`,
`ChangeOrganizationRole`, `RemoveOrganizationMembership`, `GrantGroupRole`, `ChangeGroupRole`, `RemoveGroupMembership`,
`GetOrganizationMembership`, `ListOrganizationMemberships`, `GetGroupMembership`, `ListGroupMemberships` (version 3); and
`CreateBusinessProfile`, `GetBusinessProfile`, `UpdateBusinessProfile`, `LinkGroupBusinessProfile`,
`GetGroupBusinessProfile`, `ListGroupsForBusinessProfile`, `UnlinkGroupBusinessProfile`,
`ListBusinessProfileAccessPaths`, `ListUserBusinessProfileAccessPaths` (version 4). There is no delete, move, merge, transfer
or replace operation, no cross-Organization listing, no combined create-and-link (that is Group onboarding's), no grant, bootstrap grant or
actor parameter.

- **Boundary.** The package reaches the server only through the access package, whose only production importer is
  `core/internal/mcp` (B-MCP-SOURCE); the migrate binary links neither. No HTTP, BFF, CLI, environment hook or
  bootstrap reaches it directly, and no tenancy route exists (B-TENANCY-HTTP); MCP tools reach it only through the access
  facade's guarded operations. It reads no configuration; the caller supplies a pool logged in as the runtime role. Only
  the access package imports it in production code (U-ACCESS-BOUNDARY). The internal Group onboarding routes
  (`core/internal/onboarding/api`, web BFF only) likewise call only guarded facade operations; B-TENANCY-HTTP keeps
  probing the absence of every other tenancy route.
- **Transaction-scoped helpers.** Besides the 29 methods, the package exports `ReadFacts` and four writers that
  run inside the caller's transaction: `RenameGroupTx`, `ChangeGroupSlugTx`, `UpdateBusinessProfileTx` and
  `SetOrganizationPolicyTx`. The first three share their statements, validation and error mapping with the
  corresponding `Service` methods. Like the methods, they authorize nothing: the access package is their only caller.
  Group onboarding adds `CreateGroupTx`, `CreateBusinessProfileTx`, `LinkGroupBusinessProfileTx`, `SuggestGroupSlugTx` and the
  page readers `GroupsPage` and `ProfilesPage` under the same rule.
- **Scoping.** Every Group and membership call names its Organization, and every read or change of a tenant-owned row
  carries that Organization predicate. An unknown and a foreign selector produce the identical error.
- **Transactions.** Each call is one READ COMMITTED transaction (read-only for reads) and returns success only after the
  commit. A mutation writes one row after scoped existence reads; `CreateGroup` writes exactly two (Group, then owner
  binding). Any failure rolls back everything.
- **Role changes** update the binding in place and keep its UUID. **Removal** deletes the binding; a regrant creates a
  new binding with a new UUID and the regranted role, so an obsolete role cannot return.
- **Concurrency.** Unique constraints decide duplicate grants, slugs and Group creations (exactly one winner; the
  others get `membership_exists` or `slug_conflict` and change nothing). Concurrent role changes on one binding are
  serialized by its row lock and the **last committed change wins**; there is no optimistic-concurrency or lost-update
  claim. A change that waits on a committed removal finds nothing (`membership_not_found`). A grant racing a direct
  parent or subject delete ends with exactly one side failing on the foreign key, never a dangling row.
- **Retries.** None. A transient failure is `unavailable` and the caller may retry. A commit whose reply is lost is
  also `unavailable` although the server may have committed it; a retry then reports `membership_exists`,
  `slug_conflict` or `association_exists`, never a duplicate.

### BusinessProfile operations

- **Context contract** (validated before any SQL, fields in table order; the first failing field is reported and nothing
  is written): valid UTF-8; Unicode whitespace trimmed at both ends, with no other normalization; `DisplayName`
  required, and an optional field that is nil, empty or whitespace-only is absent (NULL); no C0, DEL or C1 control
  character, except TAB and LF inside the description (CR is refused everywhere); 1–255, 1–255, 1–4,000, 1–255 and
  1–2,048 code points. Codes: `invalid_display_name`, `invalid_legal_name`, `invalid_description`, `invalid_industry`,
  `invalid_website`. The database CHECKs are defense in depth and never stricter.
- **Website contract**: printable ASCII without space or backslash; an exact lowercase `http://` or `https://` prefix; a
  URL with no opaque part, no userinfo and a non-empty host without `%`; an LDH host name (which admits dotted IPv4) or a
  bracketed IPv6 literal; a port, when present, of 1–65535. The value is stored verbatim and is **never fetched,
  resolved, rendered or treated as domain ownership**; later work that fetches it must apply its own SSRF policy.
- **Updates** replace all five fields in one statement (an omitted optional field is cleared); concurrent updates
  resolve as last committed wins, and every associated Group reads the same committed context. The UUID, Organization,
  `created_at`, associations and bindings never change.
- **Prechecks** run in order (Organization, Group in that Organization, profile in that Organization, and for linking
  that the Group has no association); unknown and foreign selectors give identical codes and error text.
- **Linking** a Group that already has an association, with the same or another profile, is `association_exists` (from
  the precheck or the unique constraint) and leaves the existing association unchanged; there is no replace operation.
  Concurrent links of one Group produce exactly one association.
- **Unlinking** names the profile: `UnlinkGroupBusinessProfile(org, group, profile)` deletes exactly that association,
  or is `association_not_found` (also when the Group is associated with another profile, which is unchanged). Since
  version 5 its Group precheck locks the Group row (`FOR NO KEY UPDATE`) before the DELETE, so an unlink conflicts with every
  guarded write relying on that Group's association; codes and results are unchanged. It is a foundation primitive:
  Group onboarding and membership administration must guard any exposure of it, so that a completed Group never silently loses its profile.
- **Reads**: `GetGroupBusinessProfile` returns a Group's single association and its profile; the access-path reads list
  committed facts on the next statement snapshot, with nothing cached.

### Error codes

`*tenancy.Error` prints `tenancy: <code>` and never IDs, names, slugs or driver text; it unwraps only to a context
cancellation or deadline.

| Code | Cause |
|---|---|
| `invalid_id`, `invalid_name`, `invalid_slug`, `slug_reserved`, `slug_required`, `invalid_role` | Input validation before any SQL; the database CHECKs map to the same codes |
| `organization_not_found`, `group_not_found`, `user_not_found`, `membership_not_found` | Precheck miss, zero-row change or the matching foreign key |
| `slug_conflict`, `membership_exists` | `groups_organization_slug_key`, `role_bindings_subject_scope_key` |
| `invalid_display_name`, `invalid_legal_name`, `invalid_description`, `invalid_industry`, `invalid_website` | BusinessProfile context validation before any SQL; the `business_profiles_*_valid` CHECKs map to the same codes |
| `business_profile_not_found`, `association_not_found` | Profile precheck miss or foreign key; unlink or Group lookup without that association |
| `association_exists` | The Group already has an association (`group_business_profiles_group_key` or its precheck) |
| `slug_unavailable` | The suggestion search is exhausted |
| `unavailable` | Connection, transaction-begin or commit failure; SQLSTATE class 08, 40, 53 or 57; cancellation or deadline |
| `internal` | Anything else, including a privilege or immutability violation |

### Logging

Each mutation writes one `tenancy.operation` record (catalogued in `events.json`): `operation` (including
`business_profile.create`, `business_profile.update`, `association.link` and `association.unlink`), `scope_type`
(membership operations only), `outcome` (`ok` at info, `rejected` at warn, `failed` at error), `error_code` exactly when
the outcome is not `ok`, and `request_id`/`trace_id` when the request context carries them. Records never contain UUIDs,
names, slugs, roles, context values, driver errors or payloads; reads are not logged. Invalid names and context values
are refused before any SQL, and database-reached failures report key columns only, so names do not reach the PostgreSQL
log either. Those key columns are UUIDs and slugs: a slug conflict or a missing reference puts the Organization UUID and
slug, or the referenced UUID, into the server log's DETAIL line (bind parameters are not logged). A slug derived from a
name therefore can reach the server log on a conflict. The transaction-scoped helpers do not log; their caller writes
one `access.decision` record instead ([AUTHORIZATION.md](AUTHORIZATION.md)). From version 8 every committed change of
Organizations, Groups, bindings, profiles and associations writes its durable audit event in the same transaction
([AUDIT.md](AUDIT.md)); privileged primitives without a request context are recorded with actor `unknown`.

## Group onboarding

Self-service Group creation and create-or-select BusinessProfile onboarding, through the guarded facade
(`core/internal/tenancy/access/onboarding.go`) and the version-14 INSERT policies. Capabilities, lock order, errors and
the mutation-path inventory are in [AUTHORIZATION.md](AUTHORIZATION.md#guarded-group-onboarding-operations); the
policies in [DATABASE-ISOLATION.md](DATABASE-ISOLATION.md); the pages in [WEB-SHELL.md](WEB-SHELL.md#group-onboarding-pages).

- **Who may create.** Every current Organization membership role (`org_owner`, `org_admin`, `org_member`),
  in Organization-only context (`group.create`). A Group role without a current Organization binding,
  a selected-Group context and a foreign Organization never create.
- **Creator becomes `group_owner`.** `CreateGroupTx` runs one keyed INSERT (a core-generated UUID v4, no RETURNING); the
  AFTER INSERT trigger `app.groups_bootstrap_owner` writes the context user's `group_owner` binding (`source =
  'membership'`) inside that statement, and the new row is then read back. There is no creator parameter, no
  Organization binding is written, and there is no second owner, owner column or `created_by`. The Group confers no
  authority in any other Group and no eligibility for any existing profile. Either both rows commit or neither does.
- **Slugs.** `SuggestGroupSlug` (transaction-scoped as `SuggestGroupSlugTx`) is advisory over the slugs the context can
  read and reserves nothing; the web form prefills it as an editable value. `groups_organization_slug_key` still decides:
  a taken slug is `conflict` (field `slug`) and nothing is overwritten; an empty slug uses the base only. A collision of
  the generated id is `unavailable`, never a slug conflict.
- **Completeness is derived.** A Group is complete exactly when its single `group_business_profiles` row exists
  (`GroupOnboarding{BusinessProfileID, Complete}`); there is no status column and no browser marker. A new Group is
  incomplete.
- **Completion** (`group.complete_onboarding`: Organization owner/admin on any Group, `group_owner`/`group_admin` on
  their own) locks the target Group and reads its association after the lock. A Group that already has a profile is
  `conflict` (field `association_exists`) with zero writes; the unique key is the backstop. Two branches:
  - *create new*: a keyed profile INSERT (no RETURNING) and the association in the same transaction, so a profile never
    commits without its association;
  - *select existing*: also `business_profile.select_for_onboarding` on the same locked facts. The profile must already
    be attached to a Group in which the subject holds `group_owner`/`group_admin`, or the subject is an Organization
    owner/admin (unlinked profiles included). The new Group never qualifies a profile. An unreadable profile is
    `not_found`; a readable one without such a path is `forbidden`.

  There is no unlink, replace or standalone profile-creation path; runtime holds no DELETE on associations.
- **Shared profile.** A completed Group's profile is edited only through the existing `UpdateBusinessProfile` (full
  replace, last commit wins), and every Group sharing it sees the change; there is no Group-owned copy.
- **Lists** (keyset): `after` UUID cursor, `limit` 1..100 (default 50), ascending id, `limit+1` rows read to
  compute `next`. `ListGroups` returns the Groups the context may read (`group.read` per row: Organization owner/admin
  all, others their own; a selected Group only that Group) with their onboarding state, plus an exact `slug` filter
  (at most one row) used for conflict recovery. `ListBusinessProfiles` ("My Business Profiles") returns summaries
  `{id, display_name, industry}` of the profiles `select_for_onboarding` admits; member-only readable profiles are
  neither listed nor selectable, and a page may hold fewer than `limit` rows while `next` (the last scanned id) is
  present. Nothing listed is authority: every write re-decides on fresh facts.
- **Retries.** No idempotency key: a repeated create with the same slug and a repeated completion are `conflict`, and
  the current committed state is shown; an `unavailable` after `COMMIT` may be durable and is reconciled by a guarded
  read.

## Membership, ownership and invitations

Migration `00015_membership_invitations.sql`. Guarded operations and capabilities are in
[AUTHORIZATION.md](AUTHORIZATION.md#guarded-membership-ownership-and-invitation-operations); roles, policies and
readers in [DATABASE-ISOLATION.md](DATABASE-ISOLATION.md); audit rows in [AUDIT.md](AUDIT.md).

- **Owners.** Every Organization has at least one `org_owner` and every Group at least one `group_owner` binding
  (`source = 'membership'`); several owners are allowed and an owner binding carries the admin powers. Ownership is
  assigned only by promoting an enabled same-scope admin (a Group recipient must also be a current Organization member)
  or by an atomic transfer; ordinary admin authority never grants ownership, and an Organization role never overrides
  Group ownership. The creator's owner binding (the Group bootstrap trigger, `CreateOrganizationWithOwner`) is the only other source.
- **Authority guard.** `app.role_bindings_authority_guard` (owner-owned SECURITY INVOKER, VOLATILE, `BEFORE UPDATE OF
  role OR DELETE`; it returns at once for any role other than runtime, owner and claim writer, so the superuser repair
  path, the seed and the recovery replay are untouched) enforces, each as its own command under READ COMMITTED: owner
  authority for any owner-role change (`owner_authority`); promotion eligibility under `rls_user_enabled_lock`, a share
  lock consistent with `DisableUser`'s `FOR UPDATE` (`recipient_ineligible`); an Organization departure only from an
  Organization-only context and only without dependent Group bindings (`context_required`, `dependent_bindings`); and
  the last-owner rule — command A locks the scope row `FOR NO KEY UPDATE`, command B counts owners with the snapshot
  taken after A's wait (`last_owner`). Each refusal rolls back the whole transaction.
- **Lock order and class-40 aborts.** Facade operations keep Group onboarding's facts order (actor bindings before Groups). Two
  facades (two owners demoting each other) or a facade and a direct runtime statement can still form a wait-for cycle;
  PostgreSQL then aborts exactly one side (`40P01`, or `40001`), always as a whole-transaction rollback with zero
  partial writes, answered `unavailable` (reason `aborted`) and never retried automatically. Because a 503 can also
  hide a committed outcome, the web page says the request may or may not have completed and links to the current list.
- **Invitations.** `app.invitations` holds the immutable intent (scope, invitable role — never an owner role —, issuer,
  intended user or normalized address, expiry at most 30 days), a 32-byte token digest (unique; the token itself is
  shown once and never stored), revocation and one-use consumption (`consumed_at`, `consumed_by_user_id`,
  `consumed_binding_id`). Runtime reads every column except the digest, inserts the intent with the digest, may set
  only `revoked_at`, and deletes nothing; `invitations_keys_immutable` keeps the intent fixed and
  `invitations_bound` limits a scope to 100 pending unexpired invitations (lock the scope row, then count). Every
  foreign key is `ON DELETE RESTRICT`.
- **Consumption.** Only the claim writer's SECURITY DEFINER bodies write a consumption: `invitation_claim_group`
  (runtime and owner; the context subject's own Group invitation) and `invitation_claim_organization` (owner role only;
  public admission is the enrollment model's, [IDENTITY.md](IDENTITY.md)). Each locks the invitation row first (duplicate claims serialize), refuses
  `not_found` unless the presented digest matches an unconsumed row, rechecks the recipient's enabled status and the
  issuer's current authority (`recipient_ineligible`, `issuer_authority_lost`), then inserts or upgrades the subject's
  own non-owner binding and records the consumption in the same transaction. A replay by the same recipient answers
  `already_member` while a qualifying binding is current, otherwise `consumed`; nothing is ever restored.
- **Retirement.** `RetirementEligibility` reads the session subject's sole-owned scopes; trusted `RetireUser` pages
  through every own Organization (`rls_own_organizations`), departs each, and identity-owned `FinalizeRetirement`
  ([IDENTITY.md](IDENTITY.md)) disables the user only when no sole ownership or new membership appeared meanwhile;
  progress is resumable.
- **Legacy owner preflight.** The 00015 up refuses, with counts only (constraint `legacy_owner_preflight`), while any
  Organization or Group lacks an owner binding; the owner lifts FORCE on `organizations`, `groups` and `role_bindings`
  only inside the migration transaction to count and restores it before any change (a refusal rolls the lift back).
  Resolution is the restricted procedure in [RECOVERY.md](RECOVERY.md#legacy-ownerless-records).
- **SQL inventory.** The tenancy writers and readers (`membership.go`, `invitations.go`; fixture-only primitives in
  `fixture.go`) are the only statements on the new objects; `token_digest` is named only by the invitation INSERT and
  the claim calls (`TestSQLBoundary`). The down migration refuses while any invitation or invitation-era audit row
  exists (history is never discarded).

## Explicit Group authority

Migration `00017_group_authority.sql`. It supersedes the membership statements above
that an Organization owner or admin administers every Group: the matrix and the override are in
[AUTHORIZATION.md](AUTHORIZATION.md#explicit-group-authority-supersedes-the-all-management-cells-above).

- **Three boundaries.** Organization owners/admins read every Group's and ordinary Business Profile's metadata (names,
  slugs, onboarding state, profile fields; never rosters, invitations, tokens or objects). Ordinary Group management
  needs an explicit Group owner/admin binding or the Organization's `group_management_override` (default false; only a
  current enabled Organization owner changes it). Group content needs an explicit Group role under either setting.
- **Override administration.** With the override on, Organization owners and admins equally administer ordinary
  members, invitations, names, slugs, onboarding and profile edits of any Group, without a membership: no self binding,
  self-invitation or own-role change, no owner role, owner invitation, last-owner bypass, Leave or Group objects. They
  can grant access to other people, a trusted delegation (the settings page says so). Each such write appends
  `access.override_used`.
- **Revocation.** Turning the override off, demoting or removing the staff member, or disabling the issuer denies the
  next request and every outstanding invitation issued solely under the override (`issuer_authority_lost`); genuine
  Group bindings, including those granted by the override, survive.
- **Creation and joining.** Creator-owner bootstrap, invitation-only Group joining and the membership owner, departure,
  single-use and audit guarantees are unchanged.

## Platform root signup

Migration `00018_platform_signup.sql`; flow in [IDENTITY.md](IDENTITY.md#platform-root-signup), authority in
[AUTHORIZATION.md](AUTHORIZATION.md#platform-root-signup-outside-the-facade).

- **New Organization with its owner.** An authenticated person at the platform root creates an Organization (name by the
  tenancy rule; slug by the hostname grammar and reserved floor, unique platform-wide, [DOMAINS.md](DOMAINS.md)) and
  becomes its only `org_owner` in the same statement sequence of one transaction, through the provision writer's
  SECURITY DEFINER body. This adds a third creator-owner source beside the two named in the membership **Owners** bullet
  above (the binding still has `source = 'membership'`; the role/source CHECK is unchanged). Defaults apply as for every
  Organization (`invite_only` enrollment, `group_management_override = false`); no Group, BusinessProfile, billing or
  other membership is created, and the creator's existing memberships elsewhere are unchanged.
- **Idempotency.** The create intent (name and slug digest) binds the creation: an identical repeat returns the same
  Organization (`already_created`) and a changed one is `intent_used`; a slug race has one winner and the loser retains
  no write (`slug_taken`). Organization authority after creation is the ordinary model above; runtime still has no
  Organization or `org_owner` INSERT path.
- **Opaque platform address (migration 00021).** The new Organization also receives a core-generated random
  platform label, its public host `<label>.<domain>`, in the same transaction; the label is not part of the create intent,
  never changes at runtime and, like every hostname, is a routing name, never identity or authority (UUIDs stay
  identity); it is not an audit attribute. The slug stays friendly Organization
  metadata (still editable by owners and admins) and no longer routes; a slug change leaves the label and its host
  unchanged ([DOMAINS.md](DOMAINS.md#platform-addresses-slugs-and-renames)).

## Labels

Migration `00020_labels.sql`; authority and lock order in
[AUTHORIZATION.md](AUTHORIZATION.md#guarded-label-operations), policies in
[DATABASE-ISOLATION.md](DATABASE-ISOLATION.md), audit in [AUDIT.md](AUDIT.md). Labels are private administrative
metadata; they grant nothing and no access decision reads them.

- **Model.** Two forced tables. `app.label_definitions` (`id`, `organization_id`, `owner_scope_type` `organization` or
  `group`, `owner_scope_id`, `group_id` NULL for the Organization scope else the owning Group, `name`, the generated
  `name_key = lower(name COLLATE "C")`, `visibility` `owner_scope` or `target_members`, `created_at`) and
  `app.label_assignments` (`id`, `organization_id`, `label_id`, the copied `owner_scope_type`/`owner_scope_id`,
  `target_type`, `target_id`, `target_group_id` = the Group that owns the target, `created_at`). Every assignment column
  is `NOT NULL`; ids are core-generated and nonzero (`*_id_nonzero`).
- **Definition constraints.** `label_definitions_scope_consistent` (an Organization scope has `owner_scope_id =
  organization_id` and no Group; a Group scope has `owner_scope_id = group_id`); `label_definitions_visibility_valid`
  (the closed value set, and `target_members` only for Group-owned definitions); `label_definitions_name_valid` (1–80
  code points, equal to its trimmed form, no C0/DEL/C1 control and no U+2028/U+2029; validated, never normalized; core's
  `ValidateLabelName` is never weaker); `label_definitions_scope_name_key UNIQUE (organization_id, owner_scope_type,
  owner_scope_id, name_key)` (ASCII-case-insensitive, otherwise byte-exact; collision `name_conflict`);
  `label_definitions_owner_key UNIQUE (organization_id, id, owner_scope_type, owner_scope_id)` (the composite target
  below); the Organization and `(organization_id, group_id)` foreign keys (RESTRICT). Triggers:
  `label_definitions_keys_immutable` (only `name` and `visibility` change; SQLSTATE 23000) and `label_definitions_bound`
  (at most 200 definitions per owner scope, counted after locking the scope row `FOR NO KEY UPDATE`; `scope_full`).
- **Assignment constraints.** `label_assignments_label_fkey (organization_id, label_id, owner_scope_type,
  owner_scope_id) → label_definitions_owner_key` ON DELETE RESTRICT is **total**: every referencing column is `NOT
  NULL`, so under MATCH SIMPLE an absent, foreign-Organization or owner-mismatched label id is refused (23503) in both
  scopes, and a definition with assignments cannot be deleted without removing them first (no hidden orphans).
  `label_assignments_target_group_fkey (organization_id, target_group_id) → groups` RESTRICT. The closed
  `label_assignments_target_shape` CHECK admits exactly `owner_scope_type = 'organization' AND owner_scope_id =
  organization_id AND target_type = 'group' AND target_group_id = target_id`: production storage holds only
  Organization-owned labels on same-Organization Groups and no Group-owned assignment row.
  `label_assignments_unique UNIQUE (label_id, target_type, target_id)` (`already_assigned`); indexes
  `label_assignments_target_idx (organization_id, target_type, target_id)` and `label_assignments_group_idx
  (organization_id, target_group_id)`; `label_assignments_keys_immutable` (an assignment is never updated).
- **Runtime grants.** Definitions: SELECT, DELETE, column-limited INSERT (`id, organization_id, owner_scope_type,
  owner_scope_id, group_id, name, visibility`) and UPDATE (`name, visibility`). Assignments: SELECT, DELETE and
  column-limited INSERT (`id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id,
  target_group_id`); no UPDATE. No helper, PUBLIC, retention or migrator grant; no new role or SECURITY DEFINER body.
- **Closed target registry.** `core/internal/tenancy/access/labels_targets.go` registers exactly one kind, `group`,
  admitted only for Organization-owned definitions; unknown, unregistered or ambiguous selectors and a Group-owned
  definition aimed at a Group are refused before any SQL. Production startup registers nothing else (`labelHooks` is
  nil); a test-only fixture kind and its schema widening exist only in the disposable databases of the integration
  suites, as template evidence, never as a production Group-owned target. No HTTP route or MCP tool reaches a
  Group-owned assignment.
- **Registration contract for a future target type `T`** (owned and reviewed by the change that introduces `T`):
  1. a core adapter registered in `labels_targets.go`: `resolve` reads the target from authoritative storage inside the
     operation's transaction and returns its owning Group (the caller supplies nothing but the id); `readable` decides
     current read authority on the resolved target from the operation's facts; `admits` names the definition owner
     scopes it accepts (a Group-owned target: the `group` scope only);
  2. a migration that widens `label_assignments_target_shape` with `(owner_scope_type = 'group' AND target_type = 'T'
     AND target_group_id = owner_scope_id)`;
  3. a per-type, NULL-skipping existence-and-ownership key: a stored generated column `target_T_id uuid GENERATED ALWAYS
     AS (CASE WHEN target_type = 'T' THEN target_id END) STORED` and the foreign key `(organization_id, target_T_id,
     target_group_id) → T (organization_id, id, group_id)`, ON UPDATE RESTRICT and ON DELETE RESTRICT or CASCADE (`T`
     carries `UNIQUE (organization_id, id, group_id)`); under MATCH SIMPLE it constrains only `T` rows and binds the
     target's existence and its actual owning Group to the definition's owning Group;
  4. a disjunct of `label_assignments_select` requiring the target row to be visible to the invoker under `T`'s own
     row-level security with the row's ownership facts (`EXISTS (SELECT 1 FROM T t WHERE t.organization_id =
     organization_id AND t.id = target_id AND t.group_id = target_group_id)`), so that `readable` and storage agree.

  Core statements never name a per-type key column. Until `T` exists, production SQL carries no per-type
  target-visibility predicate; `TestLabelTargetsClosed` pins the core map at `{group}` and `TestLabelsCatalog` CT-LAB-2
  pins the production CHECK, SELECT policy and assignment column list.
- **Operations.** The seven guarded label operations, their failures (`name_conflict`, `scope_full`,
  `already_assigned`, `not_assigned`, `target_unsupported`, `target_ambiguous`, `visibility_not_allowed`,
  `organization_context_required`) and the up-front Organization lock for Organization-scope creation are in
  [AUTHORIZATION.md](AUTHORIZATION.md#guarded-label-operations). Label SQL lives in `tenancy/labels.go`; constraint
  names map to codes in `tenancy/errors.go`.

## What later stages must secure

Historical record from versions 3–5, kept because it states each table's obligations: at that stage row-level
security was **not** enabled on the tenancy tables, no binary linked the model and nothing exposed tenant data. Version 6
enables and forces row-level security with per-command policies ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)). The
shared policy's trusted TenantContext and guarded facade authorize in process; they are not isolation. Its
[row-level security handoff](AUTHORIZATION.md#row-level-security-handoff) lists the statements, row locks and the single
transaction-owning call site.

| Object | Tenant column | Later obligation |
|---|---|---|
| `app.organizations` | `id` | Row-level security policies: visible only within the active Organization context, including the policy flag (guarded update requires `organization.manage_policy`) |
| `app.groups` | `organization_id` | Row-level security policies on every command; the guarded `group.read` and `group.update` are in place |
| `app.role_bindings` | `organization_id` | Row-level security policies, which must admit the shared policy's three `FOR SHARE` membership-fact locks; the shared policy evaluates the bindings; membership administration and ownership lifecycle (version 15) |
| `app.organization_memberships`, `app.group_memberships` | `organization_id` | Inherit `role_bindings` policies through `security_invoker` |
| `app.business_profiles` | `organization_id` | Row-level security policies on every command; the shared policy evaluates who may read or edit a profile (Organization authority and the derived Group paths) |
| `app.group_business_profiles` | `organization_id` | Row-level security policies on every command; Group onboarding (version 14) delivered the create-or-select flow (completion derived from the single association, no runtime unlink); membership administration (version 15) owns the reassignment and removal lifecycle of completed Groups |
| `app.business_profile_access_paths` | `organization_id` | Inherits the policies of both base tables through `security_invoker`; the shared policy's evaluator reads the same facts from the bindings and associations, not through this view |

Queries to secure are the `Service` statements in `core/internal/tenancy/store.go`, including the shared policy's fact and
transaction-scoped statements; each already names the owning Organization. Disabled users may keep bindings: identity
refuses their authentication, and the guarded operations require an active user row.
**Future BillingProfiles** are 1:N children of a BusinessProfile keyed by `(organization_id, business_profile_id)`
through `business_profiles_organization_id_key`, with no singleton on the profile; the BusinessProfile model adds no billing, payment, tax,
onboarding, completion or policy object.

## Migration and test-database effects

- **Down** (development and test only; refused with `APP_ENV=production`): from version 5 it drops only the
  policy flag column, destroying its values and leaving every other row unchanged at version 4; from version 4 it
  drops the BusinessProfile objects and **destroys every profile and association**, leaving identity, sessions,
  Organizations, Groups, slugs and every binding unchanged at version 3; from version 3 it drops every tenancy object
  and **destroys all tenancy data**. Identity and session tables and rows are untouched.
- **Identity test fixtures.** The identity tests reset their database with `TRUNCATE app.login_transactions,
  app.sessions, app.identity_principals, app.users, app.identity_provider_binding CASCADE`. Because `role_bindings`
  references `app.users`, the reset also empties every membership binding; Organizations, Groups, BusinessProfiles and
  associations remain (no BusinessProfile table references `app.users`), and the access paths become empty with the bindings. This
  affects disposable test databases and whatever database the `identity:staging-local` block in `IDENTITY.md` migrates
  (the Compose database that `.env` names): run that block only against a disposable qualification database.

## Commands

The checks run in the `migrations` suite of the validation harness, which only full runs select:

```sh
infra/validation/run.sh --base <commit> --allow-file <scope-allow-list> migrations
```

It runs B-TENANCY-CATALOG (exact catalog, ACLs, constraints, the single-authority detector and its negative controls,
pinned to version 3), B-TENANCY-MIGRATION (empty, populated, repeated, restarted and down/up migrations, pinned to
version 3) and B-TENANCY-IT (the model as the runtime role at version 5, concurrency, injected failures and the
operation log); and for BusinessProfiles B-BUSINESS-PROFILE-CATALOG (the additive version-4 catalog, exact ACLs and constraints, the
detector with the version-4 projection inventory and its negative controls), B-BUSINESS-PROFILE-MIGRATION (empty,
populated, repeated, restarted and down/up migrations) and B-BUSINESS-PROFILE-IT (contexts, shared profiles, single
association, cross-Organization rejection, concurrency, injected failures, derived access paths and the operation log);
and for the shared policy B-POLICY-CATALOG, B-POLICY-MIGRATION, B-POLICY-IT and B-POLICY-ORDERING
([AUTHORIZATION.md](AUTHORIZATION.md#validation)); and for Group onboarding B-ONBOARDING-CATALOG (`TestOnboardingCatalog`: the
exact version-14 delta, the restore-boundary cases, the widened audit operation constraint and down/up) and
B-ONBOARDING-IT (`TestOnboarding` ONB-1..13 on the attested runtime login). `bash infra/ci/validate.sh --profile full` (README `platform:ci`) includes them; the unit tests run in the core check
image.
