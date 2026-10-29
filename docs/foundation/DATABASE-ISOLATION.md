# Database isolation

PostgreSQL row-level security beneath core's authorization. Migration `core/migrations/00006_row_level_security.sql`
(version 6) adds no table, column, index, constraint or data: only parsers, readers, grants and policies. Core's
guarded facade ([AUTHORIZATION.md](AUTHORIZATION.md)) still decides every operation. This layer makes sure that an
omitted application predicate cannot expose another Organization, an unauthorized Group, or another user's identity
or authentication rows.

**Version 10 (MCP connections).** Migration `00010_mcp_connections.sql` adds `app.sessions.kind` (`browser` or `mcp`) and one
forced table, `app.mcp_connections`, with three Organization-bound policies (11 forced `app` tables, 42
relations; [MCP.md](MCP.md)):
- `mcp_connections_select`: the context user's own rows in the context Organization, with no role required, so the
  exact owner can still terminate a connection after losing membership;
- `mcp_connections_insert`: a pending row of the context user in the context Organization, only with a current
  Organization role (`app.rls_org_role()` not NULL);
- `mcp_connections_update`: the same rows, with a current Organization role; the SECURITY INVOKER guard trigger keeps
  the binding immutable, allows one pending-to-active transition and requires every selection change to advance the
  context version, to run in the Organization-only shape and to name a Group (and leave a previous selection) visible
  to the subject under row-level security (managers see every Group, other members their own); a clear needs nothing
  more.

`sessions_insert` keeps browser rows exactly as before (IDENTITY shape) and admits `mcp` rows only in the TENANT-ORG
shape with a current Organization role. E1 (`app.auth_resolve_session`) resolves browser sessions only, so an MCP
connection's session can never authenticate a browser request; the auth reader gains SELECT on `sessions.kind`. There
is no DELETE grant, no definer reader on the table and no token or hash column.

**Version 11 (domains).** Migration `00011_organization_domains.sql` adds `app.organizations.slug` and one forced table,
`app.organization_domains` (12 forced `app` tables, 48 relations; [DOMAINS.md](DOMAINS.md)), with:
- `organization_domains_select`, `_insert` and `_update`: rows of the context Organization for a current Organization
  owner or admin only; an insert must be a `pending` row at version 0; no DELETE policy or grant. The owner-owned
  SECURITY INVOKER guard implements the lifecycle table (allowed transitions and exactly which columns each may change);
- `organizations_update_slug`: an Organization owner or admin may update the slug; the owner-owned slug guard confines
  an admin to the slug, and `organizations_update` and the column grants are otherwise unchanged;
- two SECURITY DEFINER readers owned by `<db>_auth_reader`, EXECUTE to runtime only, ignoring ambient settings:
  `app.hostname_state(kind, key)` (one unique-index probe: routing and issuance predicates, no row data) and
  `app.is_organization_member(user, org)`; they read through the auth-reader SELECT policies
  `organization_domains_auth_reader`, `organizations_auth_reader` and `role_bindings_auth_reader`, limited by column
  grants to the columns those predicates need (from version 21, the platform branch reads
  `organizations.platform_label` instead of `slug`, and the auth reader gains `SELECT (platform_label)`).

`app.login_transactions` gains `organization_id` and `callback_origin`; its row-level security (admission by the
binding hash) is unchanged. Hostnames are not tenant secrets, but no reader returns another Organization's rows: the
resolver answers only the Organization a host routes to.

**Version 9 (recovery and retention).** Migration `00009_retention_maintenance.sql` adds privileges, four `TO <db>_retention` policies
(two on `app.stored_objects`: deleted tombstones only) and the owner-owned invoker function `app.retention_eligible`;
no relation, column, index or role changes, and runtime's and the owner's privileges are unchanged ([RECOVERY.md](RECOVERY.md)).

**Version 8 (audit).** Migration `00008_audit_events.sql` adds the append-only `audit.events` relation in the owner-owned
`audit` schema, with FORCE row-level security and two policies (`events_select`: the current `org_owner` of the context
Organization; `events_insert`: a user actor equals the context user and the Organization is NULL or the context
Organization), eighteen capture triggers on eight `app` tables and no `app` policy, function or grant change. The
context statement also sets the transaction-local `audit.request_id`/`audit.trace_id` correlation and resets the
`audit.revocation_gap` marker, outside the prior-context guard ([AUDIT.md](AUDIT.md)).

**Version 7 (object storage).** Migration `00007_stored_objects.sql` adds one forced tenant table, `stored_objects`, with three
command-specific policies on the same pattern (29 policies in all, 10 forced tables, 38 relations). See the table row
below and [OBJECT-STORAGE.md](OBJECT-STORAGE.md).

## Roles

| Role | Login | Purpose |
|---|---|---|
| `<db>_owner` | no | Owns schema `app` and every relation in it. FORCE row-level security binds the owner to the same policies as runtime. |
| `<db>_migrator` | yes | Runs migrations through `SET ROLE <db>_owner`. |
| `<db>_runtime` | yes | Core's only credential. It is never superuser or BYPASSRLS and is a member of nothing. |
| `<db>_retention` | yes | The operator's bounded retention login (version 9): `USAGE` on `app` and `audit`, SELECT of four `stored_objects` and three `audit.events` columns, DELETE on both tables limited by retention-only policies to declared-eligible rows, EXECUTE on `app.retention_eligible`; no INSERT, UPDATE, TRUNCATE, DDL or other table ([AUDIT.md](AUDIT.md), [RECOVERY.md](RECOVERY.md)). |
| `<db>_policy_reader` | no | Owns the two authority readers. Holds only SELECT on seven `role_bindings` columns and on `groups.id`/`organization_id`. Version 14 (Group onboarding) adds the eligibility reader and SELECT of four `business_profiles` and three `group_business_profiles` columns (below). |
| `<db>_auth_reader` | no | Owns the three authentication readers. Holds only SELECT on 14 `sessions` columns (15 with `kind` from version 10), `users` (id, status, display_name) and `identity_principals` (issuer, subject, user_id). |
| `<db>_claim_writer` | no | Version 15 (membership and invitations): owns exactly three SECURITY DEFINER bodies — `invitation_claim_group`, `invitation_claim_organization` and the scoped `rls_user_enabled_lock` — and holds only what they need (below): USAGE on `app`/`audit`, EXECUTE on the three context parsers, `rls_org_role`, `rls_group_role` and `is_organization_member`, SELECT on `invitations` and UPDATE of its four consumption columns, SELECT/INSERT of eight `role_bindings` columns and UPDATE (role), SELECT (id, status) and lock-only UPDATE (status) on `users`, SELECT (issuer, subject, user_id) on `identity_principals`, and INSERT of nine `audit.events` columns; no DELETE, no audit SELECT, no pool or credential. |
| `<db>_provision_writer` | no | Version 18 (platform root signup): owns exactly three SECURITY DEFINER bodies — `signup_provision_organization`, `signup_challenge_issue` and `signup_challenge_complete` — and holds only USAGE on `app`/`audit`, EXECUTE on the four context parsers, SELECT (id, status) and lock-only UPDATE (status) on `users`, SELECT and UPDATE of exactly 14 continuation/result columns on `signup_sessions`, INSERT (id, name, slug) and SELECT (id, slug) on `organizations` (version 21: plus INSERT (platform_label), no SELECT or UPDATE of it), INSERT of the eight binding columns on `role_bindings` and INSERT of nine `audit.events` columns. Not granted: EXECUTE on `app.hostname_state` or `app.is_organization_member` (runtime-only), any privilege on `login_transactions`, DELETE anywhere, audit SELECT; no pool or credential (below). |

The helpers are NOLOGIN, NOINHERIT, NOSUPERUSER, NOBYPASSRLS, with no password. Their only member is the owner, with
`INHERIT FALSE, SET TRUE`. The owner never inherits their policies, which would weaken FORCE. It may only `SET ROLE`
to them, and only migration 00006 does that: it hands the helpers their functions, grants on those functions, and
drops them on down. That runs through the migrator → owner → helper SET chain, and the owner is restored explicitly
afterwards.

No SECURITY DEFINER function and no application SQL switches roles. Runtime cannot `SET ROLE` to any of these roles.

## Context

Core sets four transaction-local settings in one statement, the first statement of every protected transaction:
`app.user_id`, `app.organization_id`, `app.group_id` and `app.login_binding`. The empty string means absent.

```sql
SELECT pg_catalog.set_config('app.user_id', $1, true), ... ('app.login_binding', $4, true)
WHERE <all four settings are currently empty>
```

- **No row returned** means the connection already carried context: a session-level value left by unsafe code. Core
  closes that connection while the transaction still holds it, so pgxpool destroys it on release, and the call fails
  closed (`unavailable` / `identity_unavailable`) before any protected statement.
- **Statement text:** the tenancy and identity packages hold byte-identical copies of the statement (U-RLS-BOUNDARY).
- **APIs:** `tenancy.InitContext` serves the facade (`access.inContextTx`, the only production transaction path).
  Identity has its own unexported initializers: EMPTY, IDENTITY and LOGIN.

Valid shapes:

| Shape | user | Organization | Group | binding | Used by |
|---|---|---|---|---|---|
| EMPTY | – | – | – | – | the expired-login cleanup only |
| LOGIN | – | – | – | ✓ | login start insert, callback consume |
| IDENTITY | ✓ | – | – | – | sign-in, renewal, logout, provider-logout return, Disable/Enable |
| TENANT-ORG | ✓ | ✓ | – | – | `Resolve` and guarded operations |
| TENANT-GROUP | ✓ | ✓ | ✓ | – | the same, with a selected Group |

**Parsers.** The four parsers `app.ctx_user()`, `ctx_org()`, `ctx_group()` and `ctx_login_binding()` validate all
four settings on every call:
- a malformed UUID raises 22P02;
- a zero UUID raises 22023;
- a malformed binding (anything but 64 lowercase hex characters) raises 22023;
- any other shape raises 22023.

Every runtime policy calls a parser, so an invalid context fails whenever a candidate row is evaluated. With no
candidate row, it returns zero rows.

**Data-dependent validity.**
`app.rls_org_role()` is NULL unless the selection is a valid TenantContext as the shared policy's `Resolve` defines it
([AUTHORIZATION.md](AUTHORIZATION.md)):
- the user holds an Organization binding;
- a selected Group belongs to that Organization, and the user is `org_owner`/`org_admin` or holds a binding in that
  exact Group.

Every tenant policy requires a non-NULL `rls_org_role()`. An unknown, foreign or unselectable selection therefore
denies every tenant row and projection, including the Organization row and the caller's own membership. A
capability cannot override this.

The caller's own `users` row stays readable in any structurally valid shape, so a live user stays live and `Resolve`
answers `context_denied`, not `unauthenticated`. Shared-scope identities are not readable in such a context.

## Policies (26 at version 6; 29 at version 7)

All policies are permissive and command-specific; there is no `FOR ALL`. The 21 runtime policies (24 at version 7)
are `TO <db>_runtime, <db>_owner`, and the 5 helper policies are `TO` their helper only.

Denied commands are denied twice: by privilege (42501) and by default-deny. Abbreviations: U, O, G and B are the
parsers; R is `rls_org_role()`; MGR means R is `org_owner` or `org_admin`; GR(g) is `rls_group_role(g)`.

| Table | Readable rows | Writable |
|---|---|---|
| `organizations` | the active Organization in a valid context | UPDATE of the policy flag by `org_owner` (lock set: readable rows) |
| `groups` | Organization owner/admin: all Groups; Group roles: their own; a selected Group restricts both | UPDATE of name and slug by owner/admin and Group owner/admin |
| `role_bindings` (7 columns) | own bindings in scope; owner/admin: every Organization-scope binding and the in-scope Group bindings; Group owner/admin: their Groups' bindings | none. UPDATE(role) exists only to lock: WITH CHECK `false` |
| `business_profiles` | owner/admin: all; Group roles: the profile of an in-scope Group they belong to | context columns: owner/admin, or Group owner/admin with the Organization flag |
| `group_business_profiles` | visible exactly when its Group is | none |
| `users` (id, status, display_name, disabled_at) | self; plus, in a valid context, the subjects of the bindings the caller may see | self only, in IDENTITY shape: display name at sign-in, status through the trusted lifecycle |
| `sessions` | own | own; INSERT and the 8 identity-written columns in IDENTITY shape only |
| `identity_principals` | own | INSERT of own in IDENTITY shape |
| `login_transactions` | own binding (LOGIN) | INSERT and DELETE of own binding; `login_transactions_expire`: expired rows in EMPTY shape only |
| `stored_objects` (version 7) | Organization objects: any current Organization role in a valid context; Group objects: the selected Group (when one is selected) and Organization owner/admin or a current role in that exact Group | INSERT of a pending row with the subject as creator: Organization objects by owner/admin, Group objects by owner/admin or any current role in that Group. UPDATE(state, sha256): Organization objects by owner/admin; Group objects by owner/admin, the Group's owner/admin, or the creator with a current role in that Group. No DELETE (42501); the SECURITY INVOKER guard trigger enforces immutable identity and the closed lifecycle (finalize only by the creator) |

- **Not under row-level security:** `identity_provider_binding` (issuer and client ID only: public namespace
  configuration needed before any identity exists) and `goose_db_version` (owner-only).
- **Views:** the three membership/access views are `security_invoker`, SELECT-only projections. The base-table
  policies apply to them.

## Readers (SECURITY DEFINER, STABLE, `search_path = pg_catalog, pg_temp`, no LEAKPROOF, PUBLIC EXECUTE revoked)

| Function | Owner | Callers | Returns |
|---|---|---|---|
| `app.rls_org_role()` | policy reader | runtime, owner | the caller's Organization role in a valid context, or NULL |
| `app.rls_group_role(uuid)` | policy reader | runtime, owner | the caller's own role in that Group (NULL outside the selected Group) |
| `app.auth_resolve_session(bytea, timestamptz)` (E1) | auth reader | runtime | ≤ 1 live session for an exact credential hash, joined to its active user (browser sessions only from version 10) |
| `app.auth_principal_subject(text, text)` (E2) | auth reader | runtime | the user mapped to an exact (issuer, subject), or NULL |
| `app.auth_logout_ref(bytea, timestamptz)` (E3) | auth reader | runtime | ≤ 1 (session, user) for an exact, unreturned, unexpired logout reference |
| `app.rls_owner_count(text, uuid)` (version 15) | policy reader | runtime, owner | the owner count of a scope the caller can currently see, else 0 |
| `app.rls_dependent_group_bindings(uuid)` (version 15) | policy reader | runtime, owner | the subject's Group bindings in the context Organization, over every Group (the departure guard) |
| `app.rls_sole_owner_scopes()` (version 15) | policy reader | runtime, owner | the caller's own sole-owned scopes (retirement eligibility) |
| `app.rls_own_organizations(uuid, integer)` (version 15) | policy reader | runtime, owner | one keyset page of the caller's own Organizations (trusted retirement) |
| `app.rls_group_invitation_candidates(uuid, uuid, integer)` (version 15) | policy reader | runtime, owner | one page of current Organization members without a binding in a Group the caller manages |
| `app.rls_user_enabled_lock(uuid, text, uuid)` (version 15, VOLATILE) | claim writer | runtime, owner | admission first with the invoker's context, then a share lock on an enabled user (NULL, no lock, when refused) |
| `app.invitation_claim_group(uuid, bytea)` (version 15, VOLATILE) | claim writer | runtime, owner | the Group consumption outcome for the context subject's own invitation |
| `app.invitation_claim_organization(uuid, bytea, text, text, text)` (version 15, VOLATILE) | claim writer | owner only | the Organization consumption outcome for verified-address evidence |
| `app.auth_resolve_signup(bytea, timestamptz)` (version 18) | auth reader | runtime, owner | ≤ 1 live, unrevoked signup credential of an active user for an exact credential hash, with its status/recovery projection; no hash is returned |
| `app.signup_provision_organization(bytea, text, text, text, bytea, bytea, text, timestamptz)` (version 18; 8 arguments with the platform label from version 21; VOLATILE) | provision writer | runtime, owner | the creation, identical-retry re-mint (with the stored target origin) or refusal outcome (`slug_taken`, `label_taken`) for the context user's own signup credential (IDENTITY shape only) |
| `app.signup_slug_taken(text)` (version 21) | auth reader | runtime, owner | whether an Organization holds that exact slug (slug availability; never routing) |
| `app.organization_id_by_slug(text)` (version 22) | auth reader | runtime, owner | the id of the Organization holding that exact slug (`C` collation), or NULL; called once per process at startup for `DEFAULT_TENANT_SLUG`, never from an HTTP route |
| `app.signup_challenge_issue(bytea, text, uuid, bytea, bytea, timestamptz)` (version 18, VOLATILE) | provision writer | runtime, owner | consumes a continuation reference for an exact origin and Organization and stores the challenge and nonce digests (LOGIN shape only) |
| `app.signup_challenge_complete(bytea, bytea, text, uuid, timestamptz)` (version 18, VOLATILE) | provision writer | runtime, owner | completes a generation whose challenge and proof digests belong to the same row (LOGIN shape only) |

Version 15 thus has eight membership DEFINER helpers: five policy-reader readers (STABLE, never locking) and three
claim-writer bodies; `TestMembershipCatalog` (CT-MEM-1) pins owner, volatility, `search_path`, ACL and body.

The policy graph is acyclic:
- the helper policies call only the parsers;
- `users` reads `role_bindings`;
- `business_profiles` reads `group_business_profiles` and `organizations`.

No 42P17 path exists.

E1 and E3 are reads gated by possession of a 256-bit secret. E2's (issuer, subject) is an opaque selector, not a
secret or proof of authentication: core verifies the provider exchange before it calls E2. All three ignore the
context settings.

## Identity and authentication flows

| Flow | Database access |
|---|---|
| Session resolution (every request) | E1, autocommit, one round trip. If renewal is due: an IDENTITY transaction runs the renewal UPDATE |
| Login start | one transaction: `init(EMPTY)` (fails closed on prior context, before anything is deleted), then `DELETE FROM app.login_transactions` (exactly: expired rows by the database clock), then `init(LOGIN b)`, then INSERT, then COMMIT |
| Callback consume | LOGIN transaction: the binding-matched DELETE … RETURNING |
| Sign-in | the binding upsert (no RLS), then E2. Mapped: IDENTITY. Unmapped: a savepoint with a core-generated v4 UUID, users and principal INSERT, and on a lost race ROLLBACK TO SAVEPOINT, then E2, then IDENTITY of the winner |
| Logout and provider-logout record | IDENTITY transactions of the resolved subject |
| Provider-logout return | E3, then an IDENTITY transaction marks the reference returned (exactly once) |
| `DisableUser` / `EnableUser` (trusted, route-less) | IDENTITY of the target the in-process caller names |

## Trust limit and residuals

The settings carry an identity core established; they do not authenticate a database client. A holder of the runtime
credential with arbitrary SQL can:
- initialize IDENTITY as any UUID and perform that subject's self-scope writes;
- read through E1/E3 what a secret it holds entitles;
- map a known provider subject to its user UUID through E2.

Two self-scope residuals are not database-enforced (an accepted design residual):
- self re-enable of `users.status`;
- self un-revocation of the subject's own sessions.

E1 and the shared policy's liveness check reject every session of a disabled account, so the second residual alone restores
nothing. The legitimate status boundary is core's trusted, route-less lifecycle.

With a fixed actor, every cross-subject and cross-tenant read, join, aggregate and write is denied. B-RLS-IT proves
this with the other subjects' rows present.

### Explicit Group authority (migration 00017; supersedes the `MGR` Group branches above)

`MGR` (`app.rls_org_role() IN ('org_owner','org_admin')`) keeps only the metadata branches: `groups_select`,
`business_profiles_select`, `group_business_profiles_select` and selected-Group validity. With `OVR` the inline
`(SELECT o.group_management_override FROM app.organizations o WHERE o.id = app.ctx_org())` and `GMGR` the caller's own
Group owner/admin role, the Group branches of `role_bindings_select`/`_update`/`_delete`, `groups_update`,
`business_profiles_update`/`_insert`, `group_business_profiles_insert` and `invitations_select`/`_insert`/`_update` are
`GMGR OR (MGR AND OVR)`; the override branch of `role_bindings_update` WITH CHECK and of `invitations_insert` never
admits the actor's own binding or a self-addressed invitation. Reviewed exceptions (`DEP` =
`pg_trigger_depth() >= 1`, true only while the schema's own trigger bodies run, i.e. the authority guard's
Organization-departure cascade): the Group branches of `role_bindings_select` and `role_bindings_delete` are
`GMGR OR (MGR AND (OVR OR DEP))`, so the cascade reads and removes the departing subject's dependent Group bindings, and
`role_bindings_update` USING is the same, so the cascade can take its `FOR NO KEY UPDATE` lock on each Group binding
before reading its role. `role_bindings_update` WITH CHECK has no `DEP` term, so the exception grants no standalone
Group-binding change at any depth, and at depth zero a raw roster read, Group-binding lock, UPDATE or DELETE by staff
without the override still sees no row. `stored_objects_select`/`_insert`/`_update` Group branches are
explicit `rls_group_role` only, under either setting. The policy-reader bodies `rls_owner_count` and
`rls_group_invitation_candidates` and the claim-writer bodies `rls_user_enabled_lock` and `invitation_claim_group` are
replaced (counts unchanged: claim-writer inventory 7, no new reader); the column is `SELECT`-granted to both helpers and
`UPDATE`-granted to runtime, and the owner-owned `organizations_group_override_guard` refuses every session but an
enabled `org_owner` (`group_override_owner`). B-RLS-IT RLS-SCOPE (40 cases) and RLS-SCOPE-MUTATE cover the staff rows
with the override off and on; B-RLS-ORDERING OR-6/OR-7 the guard and cascade races.

## Pool and transaction lifecycle

- **Per transaction:** each transaction re-initializes all four fields, and they revert at its end.
- **Prior context:** session-level authority on a connection is refused at initialization, and that connection is
  closed while held.
- **Outcomes:** success is reported only for a successful COMMIT; a failed or unknown COMMIT or ROLLBACK is
  classified conservatively as unavailable (after a lost COMMIT reply the write may have committed) and never
  retried automatically.
- **Broken connections (design note):** core does not close a connection after the
  pooled transaction's Commit or Rollback returns, because pgxpool releases the connection before returning and a
  later close could hit a new borrower. Instead, pgx kills the connection after a real rollback error and closes a
  non-idle connection after a commit error, and pgxpool destroys any closed, busy or non-idle connection on release,
  so a connection with transport loss or an undefined transaction state is never reused. A poisoned connection (prior
  context) is still closed while the transaction holds it. This relies on the pinned pgx v5.11.0 and pgxpool; another
  driver or pool must prove equivalent behaviour first.
- **Not every error means a new backend:** an application-level error manufactured after an acknowledged COMMIT on a
  clean connection is classified unknown, and that clean connection may be reused with empty settings (B-RLS-IT P8);
  the identity I3/I5 and poison cases are the ones that show a broken connection replaced by a new backend.
- **E1/E3:** these autocommit lookups do not run the guard. Their result does not depend on settings, and the next
  initialized transaction on that connection discards it.

## Provisioning

`infra/postgres/init/10-roles.sh` creates every role on a fresh data volume. For PostgreSQL without Docker init hooks,
and for volumes initialized before row-level security (version 6), run `infra/postgres/provision-roles.sql` before migrating. Run it as a
superuser, or as a CREATEROLE role that owns the database:

```sh
DB_MIGRATOR_PASSWORD=... DB_RUNTIME_PASSWORD=... DB_RETENTION_PASSWORD=... \
  psql -X -v ON_ERROR_STOP=1 -v db=<DATABASE_NAME> -d <DATABASE_NAME> -f infra/postgres/provision-roles.sql
```

- **Missing roles:** it creates missing roles with exactly the 10-roles.sh attributes, and on an existing volume adds
  only the two helpers and their memberships. A repeat run changes nothing. Since migration 00018 both paths provision four
  helpers (policy reader, auth reader, claim writer and provision writer); on a volume initialized before that the
  script adds only `<db>_provision_writer` and its owner membership, and migration 00018 refuses until it exists
  ("run infra/postgres/provision-roles.sql").
- **Refusal:** it refuses with a non-zero exit, before any change, on a missing password, an unsafe attribute or any
  unexpected membership.
- **No rotation:** provisioning never changes an existing role's password. Rotation and withdrawal of the login roles,
  including terminating pre-rotation connections, are in [SECURITY.md](../SECURITY.md) section 6.
- **Administrator path:** an administrator grants itself SET on the owner and revokes it before commit. PostgreSQL's
  automatic ADMIN-only grant to the creator remains on every role it created (owner, migrator, runtime, retention and
  each helper it created). ADMIN lets that login grant itself, or any role, membership in any of them again at any time (the
  script does so itself for the owner), which reaches every row through the helpers. Treat the provisioning
  administrator as owner-equivalent, and retire or lock its login after use.

**Deploy order:** (1) provision; (2) migrate to version 6 and deploy the row-level-security core code in one quiescent step. 00006
takes ACCESS EXCLUSIVE on the nine tables for its transaction, catalog-only. Both mismatches fail closed:
- row-level-security code against version 5: the readers do not exist, so no session resolves;
- older code against version 6: its raw SQL sees no rows.

Down restores the exact version-5 catalog (development and test only).

## Evidence

| Check | What it proves |
|---|---|
| B-RLS-CATALOG | the exact catalog and 13 negative controls |
| B-RLS-MIGRATION | empty, populated, repeat, down, restart and C4 role-switch migrations, and PROV-1..6 |
| B-RLS-IT | context, Organization, scope (the independent oracle), authentication, immutability, projections, the pool (P1–P10, I1–I5), the first-sign-in race and E-RLS-1..3 |
| B-RLS-ORDERING | OR-1..5 |
| B-OBJ-CATALOG, B-OBJ-IT (version 7) | the exact stored-object delta with seven negative controls; direct runtime and FORCE-owner SQL against the independent oracle's RLS cases ([OBJECT-STORAGE.md](OBJECT-STORAGE.md)) |
| B-IT-ISOLATION (`migrations`) | the integration tests' run-private template and fresh copies (below), T-1..T-8 |

### Test databases

Integration tests do not get their own role family or cluster. Each run's private PostgreSQL instance holds one
write-closed template `<db>_tpl`, copied from the migrated, verified-pristine `<db>`. Before each independent
invocation `<db>` is dropped and recreated from it **under the canonical name**, so the roles derived from the database
name, the `current_database()` literals and every grant inside the database are the provisioned ones. Cloning does not
copy the database ACL, so the recreation restores exactly the `10-roles.sh` grants (PUBLIC revoked; CONNECT for
`<db>_migrator`, `<db>_runtime`, `<db>_retention`). Roles, memberships and settings are cluster-global and never cloned:
a shared-catalog snapshot (family roles and `postgres` with their attributes, memberships crossing the family boundary in
either direction, family or database-qualified settings, owner/flags/ACL of `<db>` and the template) is compared before
every recreation, and any role, membership, setting or template delta refuses. A delta on `<db>` itself is recorded as
residue that the drop removes.

B-IT-ISOLATION runs right after the template in `migrations` and keeps every change reversible: **T-1** the template
record (source binding, goose history, catalog and role facts, fingerprint, empty baseline, closed flags); **T-1a**
history and table-set negatives on scratch copies; **T-1b** a stale record refused before any drop; **T-1c** a planted
provider-binding row makes the pristine check refuse; **T-2** a marker Organization absent after recreation (repeated
after the role-mutating signup catalog); **T-3** a held session refused, neither dropped nor terminated; **T-4** the
three login roles refused by the template, and `CREATE DATABASE ... TEMPLATE` refused to the migrator; **T-5a-c** an
attribute, a cross-family membership and a database-qualified setting each reported exactly, refused and reverted;
**T-6** the per-test reporter's fail-closed fixtures; **T-7** the template size and timings; **T-8** the per-test
reports present. `TestRLSProvisioned` and PROV-1..6 keep their own instances.

## Downstream consumers

- **Storage metadata (version 7):** reused the parsers and shapes, `R`/`GR`, the `TO <runtime>, <owner>` pattern and the
  acyclic graph (`00007_stored_objects.sql`, [OBJECT-STORAGE.md](OBJECT-STORAGE.md)).
- **Storage attack suite:** reuses the B-RLS-IT world and oracle for its multi-surface attacks.
- **Group onboarding and membership administration:** own membership administration and provisioning flows. Runtime has no structural INSERT or DELETE on
  tenant tables at version 6; version 14 (below) restores INSERT only, under four policies, and still no DELETE.
  Version 15 (below) grants runtime DELETE on `role_bindings` only.
- **Self-scope residual:** recorded above.
- **Manual qualification observer:** `manual_qualification_test.go` fails closed at version 6. Reconciling it
  for future identity qualification is separate work.

**Version 13 (branding).** Migration `00013_organization_branding.sql` adds one forced table, `app.organization_branding`:
runtime select/insert/update policies for the context Organization (writes only through the guarded facade with the
`organization.manage_branding` capability: owner or admin), no runtime delete, and two `<db>_auth_reader` policies — the
current logo row of `app.stored_objects` and the branding row itself — used only by the definer public projection, which
answers exactly the public fields (label, the 15 colours, the logo `v` and type) for the Organization a trusted host
resolves to. Another Organization's objects stay invisible under the runtime context as before. `TestBrandingCatalog`
pins the delta and its negatives.

**Version 14 (Group onboarding).** Migration `00014_group_onboarding.sql` adds no relation or index (still 50 relations; the
forced-table set is unchanged). It restores runtime **INSERT only**, column-limited, on four tenancy tables — `groups`
(`id`, `organization_id`, `name`, `slug`), `role_bindings` (the seven binding columns), `business_profiles` (`id`,
`organization_id` and the five context columns) and `group_business_profiles` (`organization_id`, `group_id`,
`business_profile_id`) — with no DELETE, no UPDATE change and no TRIGGER privilege, and four `FOR INSERT` policies `TO
<db>_runtime, <db>_owner` (FORCE binds the owner to the same contexts) ([TENANCY.md](TENANCY.md#group-onboarding)):
- `groups_insert`: the context Organization, Organization-only context (`G IS NULL`) and a current Organization
  membership (`R IS NOT NULL`; a Group role alone never qualifies);
- `role_bindings_insert_bootstrap`: only from inside a trigger (`pg_trigger_depth() > 0`) and only the fixed shape
  (context user, `group`, `group_id = scope_id`, `group_owner`, `membership`) in Organization-only context. The only
  trigger that inserts bindings is the owner-owned SECURITY INVOKER AFTER INSERT trigger `groups_bootstrap_owner` on
  `app.groups` (a no-op for roles other than runtime and owner or without a context user), so the creator's binding is
  written in the Group's own statement, an existing Group can never be claimed, and a direct runtime binding INSERT of
  any content is 42501;
- `business_profiles_insert`: the context Organization, a valid context, and MGR or a `group_owner`/`group_admin`
  binding of the subject (in the selected Group, when one is selected);
- `group_business_profiles_insert`: the context Organization, a valid context, the selected Group when one is
  selected, and MGR or (`GR(group_id)` is `group_owner`/`group_admin` and `app.rls_profile_eligible(business_profile_id)`).

`app.rls_profile_eligible(uuid)` is one new STABLE SECURITY DEFINER boolean reader owned by `<db>_policy_reader`
(`search_path = pg_catalog, pg_temp`, EXECUTE to runtime and owner only), created through the 00006 SET chain. It first
requires a valid current context (`ctx_user`, `ctx_org` and `rls_org_role()` non-NULL, else false before any row is
read) and returns true only when (a) the profile was **inserted by this transaction in this cluster** —
`(created_xid, created_sysid) = (pg_current_xact_id(), (pg_control_system()).system_identifier)` — or (b) it is attached
to a Group in which the context subject holds `group_owner`/`group_admin` (the selected Group only, when one is
selected). It returns no data. The policy reader gains SELECT of `business_profiles (id, organization_id, created_xid,
created_sysid)` and `group_business_profiles (organization_id, group_id, business_profile_id)` through two helper
policies limited to `organization_id = ctx_org()`; the policy graph stays acyclic. The two new `business_profiles`
columns (`created_xid xid8`, `created_sysid bigint`) are assigned once by their defaults, are in no runtime grant and are
made immutable for every role (owner included) by the replaced `business_profiles_keys_immutable()` body, so an UPDATE
cannot turn a committed profile into an attachable one, and a logically restored row carries its source cluster's
identifier ([RECOVERY.md](RECOVERY.md)). The migration refuses unless `pg_control_system()` and `pg_current_xact_id()`
are executable by the runtime and policy-reader roles (PostgreSQL's default PUBLIC grant). `TestOnboardingCatalog`
(B-ONBOARDING-CATALOG) pins the delta, its negatives and the restore-boundary cases; oracle r3 of
`rls_expected_test.go` (`RLS-SCOPE-INSERT`, 31 insert attempts) and `TestIsolationAttacks` B2/C3 exercise the policies
with fixed actors.

**Version 15 (membership and invitations).** Migration `00015_membership_invitations.sql` (model in
[TENANCY.md](TENANCY.md#membership-ownership-and-invitations)) adds the forced `app.invitations` table, the third
helper role `<db>_claim_writer` (provisioned by `provision-roles.sql` and `10-roles.sh` exactly like the version-6 helpers;
the precondition refuses unless it is NOLOGIN NOINHERIT with the owner its only member) and these policies:
`invitations_select`/`_insert`/`_update` for runtime and owner (the context Organization; managers of the scope, the
intended recipient or the consumer; an insert names the context subject as issuer and, for a Group, a current
Organization member as recipient; an update can only revoke, by column grant); `role_bindings_update` and
`role_bindings_delete` replacing the lock-only `role_bindings_lock` (the own-row branch keeps 00006's lock predicate, so
`ReadFacts`' `FOR SHARE` on own bindings is unchanged; otherwise managers of the scope); the policy reader's
`role_bindings_scope_reader`, `role_bindings_retirement_reader` and `users_policy_reader` (SELECT only); and the claim
writer's `invitations_claim_select`/`_update`, `role_bindings_claim_select`, `role_bindings_insert_invited` and
`role_bindings_claim_upgrade` (the context subject's own non-owner `membership` binding only), `users_claim_select`,
`users_claim_lock` (lock-only, `WITH CHECK (false)`), `identity_principals_claim_select` and `events_insert_claim`
(trigger-issued captures of `membership.granted`, `membership.role_changed` and `invitation.consumed` only, attributed
to the context subject, at trigger depth ≥ 1, without a context-Organization term). Runtime gains SELECT of
`invitations` without `token_digest`, INSERT of the intent with the digest, UPDATE (`revoked_at`), DELETE on
`role_bindings` and SELECT (`role_bindings.id`); the owner gains EXECUTE on `is_organization_member`. The legacy owner
preflight lifts FORCE on `organizations`, `groups` and `role_bindings` only inside the migration transaction;
runtime RLS is never disabled. `TestMembershipCatalog` (B-MEMBERSHIP-CATALOG) pins the delta, the preflight rollback
and the down restoration (NOT EXECUTED until DB qualification).

**Version 18 (platform root signup).** Migration `00018_platform_signup.sql` ([IDENTITY.md](IDENTITY.md#platform-root-signup))
adds the fourth helper role's authority and one forced table, `app.signup_sessions` (with its primary key, five unique
digest keys and one partial user index: 79 relations), holding the pending root signup credential (30 minutes), the
create-intent result and one continuation generation at a time; every secret is a 32-byte SHA-256 digest, and no
issuer, subject or `auth_time` is stored. The precondition refuses unless all four helpers are NOLOGIN NOINHERIT with
safe attributes and the owner as their only member (`INHERIT FALSE, SET TRUE`).
- **Runtime and owner** (`TO <db>_runtime, <db>_owner`): `signup_sessions_select` (own rows), `signup_sessions_insert`
  (the callback's issuing columns, IDENTITY shape, never provisioned or revoked), `signup_sessions_revoke` (revocation
  pair, own rows), `signup_sessions_prove` (exactly `nonce_consumed_at` and `proof_hash` by column grant, own rows: the
  runtime is the only writer of a non-null proof) and `signup_sessions_expire` (DELETE of expired rows in the EMPTY
  shape, the login-start sweep). Runtime and owner keep **no** INSERT path on `organizations` or `org_owner` bindings.
- **Auth reader:** `signup_sessions_auth_reader` behind `app.auth_resolve_signup`, the only lookup by credential digest
  before a context exists.
- **Provision writer** (policies `TO` it only): `signup_sessions_provision_select`/`_update` (rows of an active user;
  WITH CHECK `proof_hash IS NULL OR completed_at IS NOT NULL`, so it can reset a proof to NULL at a re-mint or keep one
  while completing, but never issue one), `organizations_provision_insert` (trigger depth 0, a context user and no context Organization, slug required; from version 21, `platform_label` required too),
  `role_bindings_insert_owner` (only the context user's own `org_owner` `membership` binding of the new Organization),
  `users_provision_select`/`users_provision_lock` (lock-only, `WITH CHECK (false)`) and `events_insert_provision`
  ([AUDIT.md](AUDIT.md)). Its bodies run the lock order users → own signup row → organizations → role_bindings (the
  membership lock order) and accept only their context shape (`signup_context`, `signup_continuation_context` refusals, 42501).
- **Login transactions:** `signup_intent`, `signup_handoff` and `expected_user_id` with three shape CHECKs (a signup
  intent has no Organization; a pinned handoff has an Organization and an expected user; only a handoff has an expected
  user) and the return-path CHECK widened to `/`, `/account`, `/join/status`, `/signup/details`. The pinned handoff row is
  a runtime INSERT under the unchanged `login_transactions_insert` policy (`binding_hash = ctx_login_binding()`); no new
  `login_transactions` policy or helper grant exists. Pinning is a restriction (which user may complete), never authority;
  the trust limit above is unchanged.
- `TestSignupCatalog` (B-SIGNUP-CATALOG) pins the delta, the 14-column provision-writer grant, the non-grants and the
  policy negatives; B-RLS-MIGRATION PROV-1..6 and B-DBRENAME cover the fourth helper in both bootstrap paths.

**Version 19 (MCP connection names).** Migration `00019_mcp_connection_names.sql` ([MCP.md](MCP.md)) adds no relation (79) and no
policy. `app.mcp_connections` gains the nullable `display_name` (no default, no backfill: existing rows stay unnamed)
with the defensive stored-shape CHECK `mcp_connections_display_name_valid`; core's Unicode validation is authoritative.
The runtime role gains exactly `INSERT (display_name)` and `UPDATE (display_name)` on the table (its INSERT column set
becomes nine, its UPDATE set six); nothing else, no DELETE, no new role, no BYPASSRLS, no definer function. The three
version-10 policies are unchanged: `mcp_connections_update` already requires the context user's own row in the context
Organization **with a current Organization role**, so a rename can reach no foreign user's or tenant's row and no
non-member's row at the database either; `mcp_connections_select` (own rows, no role) and the membership-independent
exact-own revoke are unchanged. The `app.mcp_connections_guard()` body is unchanged and never inspects the name, so a
name-only UPDATE passes while every state, client, binding and selection rule holds. The strict capture trigger
`audit_capture_name` records `mcp.connection_renamed` with no name ([AUDIT.md](AUDIT.md)). `TestMCPNamesCatalog` and
`TestMCPNamesMigration` (B-MCP-CATALOG) pin the delta, the column ACL, the negatives and the down restoration.

**Version 20 (labels).** Migration `00020_labels.sql` ([TENANCY.md](TENANCY.md#labels)) adds two forced tables,
`app.label_definitions` and `app.label_assignments` (nine relations, 79 → 88), with ENABLE and FORCE row-level security
and **seven** command-specific policies `TO` runtime, owner (assignments have no UPDATE policy because they have no
UPDATE grant):

- `label_definitions_select`: the context Organization with a current Organization role, and either an
  Organization-owned row with `MGR` (Organization owner/admin) or a Group-owned row of the selected Group (or any Group
  in Organization-only context) with `GMGR` on its owning Group or, for `target_members` rows, any own role there
  (`app.rls_group_role(group_id) IS NOT NULL`).
- `label_definitions_insert`/`_update`/`_delete`: the same predicate with the manager branches only (`MGR` for the
  Organization scope, `GMGR` for the Group scope).
- `label_assignments_select`: the context Organization with a current role, the selected-Group filter on
  `target_group_id`, `target_type = 'group'`, and an `EXISTS` on `app.label_definitions` qualified by the outer
  `label_assignments.label_id`/`.organization_id`, so the definitions policy applies inside it under the invoker's
  row-level security (acyclic; no 42P17).
- `label_assignments_insert`/`_delete`: the selected-Group filter and either an Organization-owned definition with `MGR`
  on a `group` target or a Group-owned definition with an own `group_owner`/`group_admin` role in `target_group_id`
  (type-independent; it admits no production row while `label_assignments_target_shape` is closed to `group`).

No policy reads `group_management_override` or `pg_trigger_depth()`, so neither Organization authority nor the Group-management
override reaches a Group-owned label. The runtime role gains column-limited INSERT (`id, organization_id,
owner_scope_type, owner_scope_id, group_id, name, visibility`) and UPDATE (`name, visibility`) plus SELECT and DELETE on
definitions, and column-limited INSERT (`id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type,
target_id, target_group_id`) plus SELECT and DELETE on assignments; no assignment UPDATE, no helper, PUBLIC, retention or
migrator grant, no new role, no SECURITY DEFINER body. The four helpers and their reader inventories are unchanged
(functions owned by the policy reader 9, claim writer 7, auth reader 8, provision writer 3; asserted by the migration's
precondition and postcondition). `TestLabelsCatalog` CT-LAB-1..4 and `TestLabelsMigration` MG-LAB-1..4 pin the delta
and the down restoration; `TestRLSIsolation` (RLS-LABEL) and `TestIsolationAttacks` (G1) prove the policies as the
runtime role.

**Version 21 (platform tenant addresses).** Migration `00021_platform_labels.sql` ([DOMAINS.md](DOMAINS.md#platform-addresses-slugs-and-renames))
adds `app.organizations.platform_label` (nullable, `C`-collated, the slug grammar and the exact 00011 reserved floor as
CHECKs) and its partial unique index `organizations_platform_label_key` (one relation, 88 → 89); existing rows receive a
random 40-character hex label in an owner-side backfill bracketed by `NO FORCE … FORCE ROW LEVEL SECURITY` inside the
migration transaction (the update capture does not fire: no audit row). The privilege delta is exactly two column
grants: the auth reader `SELECT (platform_label)` (the one column the replaced `app.hostname_state` platform branch reads
under its definer role) and the provision writer `INSERT (platform_label)`. No role holds `UPDATE (platform_label)`:
labels are immutable at runtime, and the slug edit route's `UPDATE (slug)` and slug guard cannot change them. One policy
changes: the writer-only `organizations_provision_insert` also requires `platform_label IS NOT NULL`, so production signup
cannot create an Organization without a label; fixture and owner paths outside that policy keep inserting NULL. Three
definer bodies, all `search_path = pg_catalog, pg_temp`, PUBLIC EXECUTE revoked: `app.hostname_state` (auth reader,
replaced in place, EXECUTE to runtime only), the 8-argument `app.signup_provision_organization` (provision writer,
replacing the 7-argument body; EXECUTE to runtime and owner; `label_taken` on the label index, the stored
`target_origin` on `already_created`, so the writer never reads `organizations`) and the new `app.signup_slug_taken`
(auth reader; EXECUTE to runtime and owner). No new role or table policy; reader inventories: policy reader 9, claim
writer 7, auth reader **9**, provision writer 3. The down (development and test only) refuses while a provisioned,
unrevoked signup row has a label-derived (or NULL) target origin and a live root credential or an in-flight (or NULL)
continuation, then restores the version-20 shapes. `TestPlatformLabelsCatalog` CT-PLB-1..6 and
`TestPlatformLabelsMigration` MG-PLB-1..4 (B-PLATFORM-ADDRESS-CATALOG) pin the delta, the bodies under their real roles,
the `42501` negatives and the down state matrix.

**Version 22 (default-tenant root alias).** Migration `00022_default_tenant.sql`
([DOMAINS.md](DOMAINS.md#default-tenant-root-alias-and-root-presentation)) adds one definer body,
`app.organization_id_by_slug(text) RETURNS uuid` (auth reader, STABLE, `search_path = pg_catalog, pg_temp`, PUBLIC
EXECUTE revoked, EXECUTE to runtime and owner): the id of the Organization whose slug equals the argument exactly
(`COLLATE "C"`), or NULL. It reads through the existing 00011 auth-reader `SELECT (id, slug)` grant and `USING (true)`
policy, so its privilege class is that of `app.signup_slug_taken`, one column wider in its result. Core calls it once
per process at startup, when `DEFAULT_TENANT_SLUG` is set, on the runtime pool; no HTTP route reaches it, and the
runtime's own `organizations_select` policy is unchanged. No table, column, policy, role or runtime table grant
changes (still 89 relations); reader inventories: policy reader 9, claim writer 7, auth reader **10**, provision writer
3. The down (development and test only) drops the body as its owner and restores the version-21 state.
`TestDefaultTenantCatalog` CT-DT-1..3 pins the delta, the body under the runtime role (exact slug, unknown and
differently cased slug, platform label never a key, nothing written, `42501` for the other helpers) and the down/up
round trip.
