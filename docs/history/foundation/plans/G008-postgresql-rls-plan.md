> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G008 — PostgreSQL RLS and least-privilege roles: implementation plan (r3)

- **Plan revision:** r3, 2026-10-13 UTC, correcting r2 under the
  r2 Codex review (not included) (CORRECTION LOOP: one cleanup/pool safety case).
  - The owner's PG-002 disposition (not included) stands.
  - The exact r2 (not included), r1 (not included) and
    r0 (not included) bytes are preserved.
  - Earlier reviews: r1 (not included), r0 (not included).
  - Revision history: §14.
  - Author: Claude Code (the single G008 planning worker). Reviewer: Codex. Major decisions: product owner.
- **Status: PLAN DRAFT for Codex review.**
  - Nothing has been implemented, migrated, built, tested, run in Docker, pushed or merged.
  - **Zero of seven B7 attempts consumed.**
  - Q7 (PG-002) is resolved by the owner ("Proceed with measurements"; §12).
- **Tier:** architecture/security/data. Explicit Plan ACCEPT and a fresh execution binding gate implementation.
- **Planning provenance:**
  - Worktree: native `.claude/worktrees/agent-<id>` (agent `agent-<id>`). The generated
    name is allowed by the one-worker exception in the
    start authorization (not included).
  - Branch: `g008-plan`. r0 was committed at `0a2ca3c4a8653360d8da1e73558a244258189eaa` and r1 at
    `05eefe30512d22107a8e81a409ee2bc437beff62`, and r2 at `2fc98c19171d40d2fd96afb10bbfda4afbb3bb01`. r3 is committed
    on top of r2, unrebased, as instructed.
  - Effort: dispatched at xhigh. The effective setting cannot be seen from inside this worker, so it is reported as
    unknown.
  - Working scope: root governance first, then `core/`. The harness resets the shell to the worktree root on every
    call. No core-scoped skill exists.

## 1. Input binding (planning)

Source and schema inspection uses `9129a695812ca99690cbfbb79a840c80f4e700cf`. Planning-management inputs come from
`origin/staging` = `3068d984ed16aaadc8ddc16df5f1a568b3f240ea`.

The delta `9129a69..3068d98` touches only management and documentation files:
- the r0, r1 and r2 reviews;
- the owner disposition record;
- PG-002 notes in `docs/PERFORMANCE.md`;
- PROGRESS.

No code, schema or harness changed, so the source inspection still holds. `a316dd3..3068d98` adds only the r2
review and a PROGRESS update, so PERFORMANCE.md is unchanged since `a316dd3`.

| Input | SHA-256 |
|---|---|
| r2 Codex review `management/v0/results/G008-plan-r2-codex-review.md` (at `3068d98`) | `18d35611a69e8e6014c44b8b4033671958677aac03534c8f4e462428d47b2a2a` |
| `management/v0/PROGRESS.md` at `3068d98` | `0fe7afa3d93b4f5c02fd9a7234b942cdd6961389d9146f3dbbe59a4935c3e143` |
| r1 Codex review `management/v0/results/G008-plan-r1-codex-review.md` (at `590f003`) | `d958a43c6e4fb61d61c33b1663f917a588f67c7e643f348b846d9487e9493961` |
| Owner PG-002 disposition `management/v0/results/G008-performance-owner-disposition.md` (at `a316dd3`) | `22d6cbfa6e0e977892386f0274731e775744c91f14d08b672898972b5b35ad5b` |
| `docs/PERFORMANCE.md` (unchanged `a316dd3`→`3068d98`), incl. the PG-002 r0/r1 notes and the owner disposition (reconciled input, not a waiver) | `49fedd5cd4311fa21078ddcc0341ef41111ea980c6bdf7973d952ec2f692ec3b` |
| Superseded PROGRESS used by r2 (history; at `a316dd3`) | `7b7598a4e1199aa79ac69d54def9af763c53bb120e940031352a944dcf45c991` |
| r0 Codex review `management/v0/results/G008-plan-r0-codex-review.md` | `70ae630e1e4633ef60337829a115274932c1e998adf660fe0c725ade8e0021a2` |
| r2 Plan (archived as `G008-postgresql-rls-plan-r2.md`) | `cfdd024222fe781512dee34683c1ab9da2a8711ad66bfa93b0593bceb133fec5` |
| r1 Plan (archived as `G008-postgresql-rls-plan-r1.md`) | `9f3bc5982a25004582cc591c0d6fc0ce93767f5bd83cdb19db15b6b5d9ccfdd8` |
| r0 Plan (archived as `G008-postgresql-rls-plan-r0.md`) | `2a18da9ccd9593643ef507b8c0122c68ad6bc7c2ad519f2e8b59f719b385907f` |
| Superseded PERFORMANCE/PROGRESS used by r1 (history; at `32f3473`) | `37dfaa93…dc6d` / `f2c63c01…bb4c` |
| Goal r3 `management/v0/goals/G008-postgresql-rls.md` | `6508eb973c87a3b3fc4db067feaeaad2e9b3083b4c3d6e98728c6c73780851e2` |
| Spec r3 `management/v0/specs/G008-postgresql-rls-spec.md` | `d3cb3c51f69bdf5475204aea6597a3de00e37b057df77b01241200ae8cedac6a` |
| Reassessment `management/v0/results/G008-accepted-baseline-reassessment.md` | `cfbd45b9e33f9731e6b270d5bb3815b091565141740587dce36291925ef20040` |
| Start authorization `management/v0/results/G008-planning-start-authorization.md` | `15879e0b22fab0fff574532d115ef5e2487cfb71656df5099237deb7976e60f0` |
| Goal r2 / Spec r2 archives (history only) | `7cefde5c8b1ca2b917801b29cb5c685e0200ef614e5a1b7dfd07ada15a743941` / `af64016d1db003455a698d40b9ce3c3166ad47844341febac0ef4441652ed097` |
| `AGENTS.md` | `d548953dd67d828dfbbac8f7563e80520e3e7a1e8321913300a7bc4279a8ae0c` |
| `CLAUDE.md` | `88ef1bd590f71678d14836261088c98288b041d06e089d1109bcaa7cec81756d` |
| `docs/ARCHITECTURE.md` | `365a998ed4485e32a06b5fe572eab5446bada5c671261f43996310887cdaa975` |
| `docs/DECISIONS.md` | `b23522208c3f8aece05a6f5c930b688b6d44ad181dd632cfeb7ce5c7f40492f8` |
| `docs/SECURITY.md` (empty file) | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `management/BUDGET_POLICY.md` | `eb9d458ed5bfda365c6ca4dcef215dbe8b38679dede0d4d53c13d3cd7e210988` |
| `management/v0/ROADMAP.md` | `b92b6ae4e18b43e8d082ebe9380b9709e5f225b1b4686543c0756e59c1c322cb` |
| `docs/adr/ADR-003-workos-identity-provider.md` | `537522dca62e57cd291ffdad581ea000a1b04c45b378d5f505af061fdb952a62` |
| `docs/adr/ADR-004-group-business-profile-cardinality.md` | `51a487844d74e913d195c963095c2a381deeb293d1829085f2f026f060955a68` |
| `docs/foundation/AUTHORIZATION.md` | `0363d9d177e06395e3f1db528081791cbeeb6a928d2b48c27eb5627bda884f22` |
| `docs/foundation/TENANCY.md` | `03de23b4e7b333f3a9ac68b4f05b6f4cf5504caac783b6a60ed6b9cb61756f92` |
| `docs/foundation/MIGRATIONS.md` | `95cd689df447aa50ae07205c89e60504ca826af0bf6d142fa0023bbbf745394e` |
| `skills/saas-foundation/references/database-rls.md` | `d132268945bc4e91290d36c7d8abca37683907a50ca7ceb0dace566a0b9d4266` |
| `skills/saas-foundation/references/tenancy-authorization.md` | `2bb6171145631a5a917aa150c924ddb831cca1eb25625c0aed7acfbf95109a99` |
| `core/migrations/00005_organization_policy.sql` | `743f161d15221ffe400c341627bdd4d8e737b3a8523e2864f87abc525469cfdd` |
| `core/internal/identity/store.go` | `7fd548325fc878e4d4578c00c2a4692b71771d57defcb6bdb20c30298d439f06` |
| `core/internal/tenancy/access/tx.go` | `6a5aa28685e27931818c8e6f93e7c97ce3b6df51ee1779541e68fcda446a1256` |
| `core/internal/tenancy/store.go` | `640385b6da0af2197dd39c40cd109b6b16118fee2934122cf72089668ae64908` |
| `infra/postgres/init/10-roles.sh` | `a4617f869b4926e5aab1cb756a1c6b1f56ac258591b8de2ed68da9444ff23360` |
| Git tree `core/migrations` / `core/internal/identity` | `bcc8c4209a1028208c3ebaaa0992850eb712e43b` / `89369083ae05ab5017ead8f33323a5cdc90fe853` |
| Git tree `core/internal/tenancy` / `infra/validation` / `infra/postgres` | `ebbab38a253684f5bbebedf765c8801407eeaed2` / `887ed6520db7e99fea4a1d94844aaa3724b6a1f1` / `795480838004b0c8dabe25d0af74cc1fd722a094` |

**Execution gates (all pending):**
- **E1:** Codex ACCEPT of r3, including the D1–D4 interpretations. The owner's Q7 (PG-002) disposition is
  **satisfied** ("Proceed with measurements"; §12).
- **E2:** a fresh committed execution baseline, an approved-input snapshot and digest, the exact §10 allow-list, and
  a task-prefixed isolated checkout (`g008-agent-<id>-impl`).
- **E3:** reconciliation against that baseline:
  - equal inputs need nothing;
  - mechanical drift is re-derived and recorded;
  - semantic identity, tenancy, migration or harness changes return to Codex;
  - if another migration lands first, 00006 is renumbered mechanically.
- **S0:** a read-only start gate:
  - HEAD equals the bound base;
  - NUL-safe status is clean and root `.env` is absent;
  - `run-approved.sh --check-only` and `bash infra/ci/docker-desktop-mounts.sh` pass;
  - the ports and `local-harness.lock` are free;
  - no foreign owned Docker resources exist;
  - GitLab compute-minute usage has been read.

G054 remains independent.

## 2. Approach (smallest r3-compliant design)

1. **One additive migration, `core/migrations/00006_row_level_security.sql`,** with these parts:
   - four context parsers that validate the whole context shape (§4.1);
   - ENABLE and FORCE RLS on **nine** tables: five tenant, `users`, `sessions`, `identity_principals`,
     `login_transactions`;
   - command-specific policies on those tables;
   - five SELECT-only SECURITY DEFINER readers owned by two dedicated NOLOGIN, NOBYPASSRLS helper roles;
   - least-privilege grants, exact postconditions and an exact down.
2. **Authority reader (`<db>_policy_reader`)** has two helpers, `app.rls_org_role()` and `app.rls_group_role(uuid)`.
   They read only the caller's own bindings in the active Organization, plus the selected Group's row. They give:
   - the Organization gate;
   - selected-Group validity, as G007 `Resolve` defines it;
   - the caller's Group role.

   This keeps the policy graph acyclic and lets `role_bindings` admit the manager-scoped membership visibility C1
   requires (§4.3).
3. **Authentication reader (`<db>_auth_reader`)** has three narrowly defined lookups:
   - session resolution by credential hash (E1) and logout reference to session (E3). Both are reads gated by
     possession of a 256-bit secret, and both ignore ambient settings;
   - provider principal to user (E2), keyed by an **opaque provider-subject selector**. The selector is not a
     secret and not proof of authentication; core verifies the provider before calling E2, and E2 returns only a
     UUID.

   Every other authentication read or write runs under self-scoped policies, using the subject a lookup returned.
   Login transactions are keyed by the browser's login binding hash. Expired-login cleanup is allowed only in the
   EMPTY shape. Production establishes that shape with a guarded `init(EMPTY)` on the same transaction and
   connection before deleting anything (§4.5 E5, F3; r3).
4. **One context statement per protected transaction** sets four transaction-local settings in one statement. It
   finds nothing when prior authority is present, and core then discards the connection.
5. **Lock-only authority stays as in r0.** The UPDATE USING read set keeps every accepted lock; WITH CHECK `false`
   or a write predicate denies mutation.
6. **Everything else is reuse:** the G007 facade, ordering and commit classifier, the harness, wrapper and full CI
   profile, and the subsetFS pinning of predecessor suites.

## 3. Pinned inventory (R1, R2)

Version-5 catalog, verified from source and re-verified live by CT-RLS-1:
- 10 tables, 3 `security_invoker` views and 22 indexes (35 relations beyond metadata);
- 5 trigger functions with 5 BEFORE UPDATE triggers;
- the goose metadata table and its sequence;
- no policy, RLS flag or definer.

Version 6 adds no relation, index, trigger, column or constraint.

### 3.1 Relation, command and column contract at version 6

Notes for this table:
- "Runtime" is `<db>_runtime`. Every runtime policy is `TO <db>_runtime, <db>_owner`, so FORCE binds the owner to
  the same contexts.
- Denied commands are denied twice: by privilege (42501) and by default-deny RLS.

| Relation (class) | Runtime SELECT | INSERT | UPDATE | DELETE |
|---|---|---|---|---|
| `organizations` (tenant root) | table; `organizations_select` | **denied**: INSERT(name) revoked | lock USING read set; writes only `group_admins_edit_business_profiles` (UPDATE(name) revoked); WITH CHECK `org_owner` | **denied** |
| `groups` (tenant) | table; `groups_select` | **denied**: INSERT revoked | lock USING read set; writes name, slug; WITH CHECK group.update | **denied** |
| `role_bindings` (tenant authority) | **columns** organization_id, subject_user_id, scope_type, scope_id, group_id, role, source (`id` and `created_at` revoked); `role_bindings_select` (own + manager scope, §4.2) | **denied**: INSERT revoked | **lock only**: UPDATE(role); `role_bindings_lock` USING own rows, WITH CHECK `false` | **denied**: DELETE revoked |
| `business_profiles` (tenant) | table; `business_profiles_select` | **denied**: INSERT revoked | lock USING read set; writes the five context columns; WITH CHECK profile.update incl. the flag | **denied** |
| `group_business_profiles` (tenant) | table; `group_business_profiles_select` | **denied**: INSERT revoked | **denied**: no privilege (the Group row is the lock surrogate) | **denied**: DELETE revoked |
| `users` (global identity) | **columns** id, status, display_name, disabled_at; `users_select` (self + shared permitted scope, §4.2) | id, display_name, created_at, updated_at; `users_insert` | display_name, updated_at, status, disabled_at; `users_update` | **denied** |
| `sessions` (authentication) | table; `sessions_select` (own) | table; `sessions_insert` (own, identity shape) | the 8 G004-written columns (§4.5); `sessions_update` (own; writes in identity shape only) | **denied** |
| `identity_principals` (authentication) | table; `identity_principals_select` (own) | table; `identity_principals_insert` (own, identity shape) | **revoked** (unused by G004) | **denied** |
| `login_transactions` (authentication) | table; `login_transactions_select` (own binding) | table; `login_transactions_insert` (own binding) | none | table; `login_transactions_delete` (own binding) plus `login_transactions_expire` (expired rows only, **EMPTY shape only**, §4.5 E5) |
| `identity_provider_binding` (public namespace configuration; **no RLS**) | table | table | none | none |
| `organization_memberships`, `group_memberships`, `business_profile_access_paths` | SELECT; the invoker's base RLS applies | none | none | none |
| `app.goose_db_version` and its sequence | owner only, by exact identity | — | — | — |

- **Provider binding (separate class):** the single row holds `issuer` and `client_id`. Both are public (they
  appear in every authorization redirect), and the binding is needed before any identity exists. Its exact G004
  grants are the contract; it carries no user or secret.
- **Never held by runtime:** TRUNCATE, REFERENCES and TRIGGER, on all tables.
- **Function allowlist:**
  - Owner-owned SECURITY INVOKER: 5 trigger functions (no EXECUTE) and `app.ctx_user()`, `app.ctx_org()`,
    `app.ctx_group()`, `app.ctx_login_binding()`. EXECUTE on these goes to runtime and `<db>_policy_reader`.
  - Definers owned by `<db>_policy_reader`: `app.rls_org_role()` and `app.rls_group_role(uuid)`. EXECUTE goes to
    runtime and owner.
  - Definers owned by `<db>_auth_reader`: `app.auth_resolve_session(bytea, timestamptz)`,
    `app.auth_principal_subject(text, text)` and `app.auth_logout_ref(bytea, timestamptz)`. EXECUTE goes to
    runtime only.
  - All new functions are `plpgsql STABLE`, `SET search_path = pg_catalog, pg_temp`, not LEAKPROOF, with PUBLIC
    EXECUTE revoked. No dynamic SQL.
- **Roles:**
  - runtime, migrator, owner and retention are unchanged;
  - new helper roles `<db>_policy_reader` and `<db>_auth_reader`: NOLOGIN, NOSUPERUSER, NOBYPASSRLS, NOCREATEDB,
    NOCREATEROLE, NOINHERIT, no password;
  - each helper's only member is `<db>_owner`, `WITH INHERIT FALSE, SET TRUE` (needed to own the functions the
    migration creates);
  - runtime is a member of nothing.
- **Helper grants and policies** (SELECT-only, role-specific):

| Helper | Grants | Policy (`TO` that helper only) |
|---|---|---|
| policy_reader | USAGE on `app`; SELECT(organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) on `role_bindings`; SELECT(id, organization_id) on `groups`; EXECUTE on the four `ctx_*` | `role_bindings_policy_reader` USING `subject_user_id = app.ctx_user() AND organization_id = app.ctx_org()`; `groups_policy_reader` USING `id = app.ctx_group() AND organization_id = app.ctx_org()` |
| auth_reader | USAGE on `app`; the exact column grants listed below | `sessions_auth_reader`, `users_auth_reader`, `identity_principals_auth_reader`, each FOR SELECT USING `true` (D3) |

**auth_reader column grants (literal union; nothing else, no table-level privilege):**

| Table | Granted columns (SELECT) | Needed by |
|---|---|---|
| `sessions` | `id`, `user_id`, `credential_hash`, `csrf_token`, `provider_session_id`, `mfa_assurance`, `auth_time`, `idle_expires_at`, `absolute_expires_at`, `renewed_at`, `revoked_at`, `provider_logout_ref_hash`, `provider_logout_ref_expires_at`, `provider_logout_returned_at` (14 columns) | E1 returns `id` through `provider_session_id` and filters on `credential_hash`, `revoked_at` and both expiries. E3 returns `id`, `user_id` and filters on the three `provider_logout_*` columns |
| `users` | `id`, `status`, `display_name` | E1: the join, the active-user filter and the display name |
| `identity_principals` | `issuer`, `subject`, `user_id` | E2 |

Columns **not** granted to auth_reader: `sessions.created_at`, `sessions.revoke_reason` and
`sessions.provider_revocation`, plus every other table and column. CT-RLS-3 asserts the exact helper table- and
column-ACL strings with its existing ACL check; NEG N12 is a rolled-back grant of `sessions.revoke_reason`.

### 3.2 Privileged and trusted path inventory

| # | Path | Database identity | RLS position |
|---|---|---|---|
| T1 | Migrations | migrator, `SET ROLE` owner | DDL only. 00006 has no DML. A transient CREATE on `app` for the helpers lets the owner hand them their functions; it is revoked in the same transaction |
| T2 | Provisioning: `10-roles.sh` (first start) and the new portable `infra/postgres/provision-roles.sql` | image superuser, or a CREATEROLE admin that owns the database | Not reachable by core. Creates roles and the schema only |
| T3 | Guarded facade: `Resolve` and 8 operations, through `access.inTx` | runtime | TENANT shape |
| T4 | Identity flows F1–F10 (§4.5) | runtime | IDENTITY or LOGIN shape, or the E1–E3 definers |
| T5 | `DisableUser`/`EnableUser` (trusted, no route) | runtime | IDENTITY shape set to the target by the trusted in-process caller (E7) |
| T6 | Raw tenancy model (29 methods, `ReadFacts`, `*Tx`) | runtime | `ReadFacts`/`*Tx` only inside `inTx`. Structural mutations denied at version 6 (decided Q5). Tests use them through T7 |
| T7 | Test setup: fixtures, snapshots, conflicting writers, controls | separately attested superuser `setupConn`, disposable instances only (decided Q4); owner for rolled-back catalog controls and TRUNCATE resets | Never an assertion path |
| T8 | Harness catalog reads (`psql_super`) | image superuser | Validation only |
| T9 | Retention role | none | Reserved |
| T10 | Health probe | runtime `Ping` | No relation |

Core holds only the runtime credential. Web and intelligence hold none (B-MIGRATE-ENV, B-MCP-SOURCE unchanged).

## 4. Design

### 4.1 Context fields, shapes, parsers, statement and API (R3, C3)

**Fields:**
- `app.user_id`, `app.organization_id`, `app.group_id`: canonical UUID text;
- `app.login_binding`: 64 lowercase hex characters, the SHA-256 of the browser's login binding;
- `''` means absent.

**Valid shapes (exhaustive):**

| Shape | user | org | group | binding | Used by |
|---|---|---|---|---|---|
| EMPTY | – | – | – | – | **the only shape in which expired-login cleanup (E5) deletes anything**. Production enters it explicitly with a guarded `init(EMPTY)` (F3). Also the provider binding (no RLS). E1–E3 ignore settings and are not shape-gated |
| LOGIN | – | – | – | ✓ | login start insert, callback consume |
| IDENTITY | ✓ | – | – | – | sign-in, renewal, logout writes, provider-logout consume, Disable/Enable |
| TENANT-ORG | ✓ | ✓ | – | – | `Resolve`/guarded operations without a Group |
| TENANT-GROUP | ✓ | ✓ | ✓ | – | `Resolve`/guarded operations with a selected Group |

**Parsers.** `app.ctx_user/org/group/login_binding()` each parse and validate **all four** settings, then return
their own field. Any policy that references any parser therefore validates the whole context:
- a malformed UUID raises 22P02;
- an all-zero UUID raises 22023 (`zero context identifier`);
- a malformed binding raises 22023 (`malformed context binding`);
- any shape outside the table raises 22023 (`invalid context shape`), e.g. Group without Organization,
  Organization without user, or binding with any other field.

Every runtime policy references a parser, so structural validation always fires once a protected row reaches policy
evaluation. This now includes `login_transactions_expire` (r2). With at least one candidate row, the outcome is the
pinned SQLSTATE. With none, it is zero rows. Neither case returns data or performs a mutation.

For E5, a candidate row is an **expired** row: the planner may skip parser evaluation for rows that fail
`expires_at <= now()`.

**Cleanup-shape rule (E5).** `login_transactions_expire` admits a row only when it is expired and both `U` and `B`
parse to NULL. Because the parsers enforce the whole-shape table, `U` and `B` both NULL is exactly the EMPTY shape:
Organization or Group without a user already raises 22023. In every other valid shape, the expire policy deletes
nothing; that includes values left on a poisoned connection.

The expire policy alone is not enough, because it is ORed with `login_transactions_delete`. A no-WHERE DELETE run
on a connection carrying a leftover session-level `app.login_binding = b` would delete binding `b`'s own rows,
including unexpired ones. That is why **production never runs the cleanup without first establishing a verified
EMPTY context on the same transaction** (`init(EMPTY)`, F3, r3). Prior authority of any field, the binding
included, is therefore refused and the connection discarded before any mutation.

**Data-dependent validity.** `app.rls_org_role()` is NULL, and every tenant policy therefore denies, unless all of
these hold, which is G007 `Resolve` in SQL:
- user and Organization are present;
- the caller holds an Organization binding;
- if a Group is selected, it belongs to the Organization and the caller is `org_owner`/`org_admin` or holds a
  binding in that exact Group.

Unknown, foreign and unselectable selections are therefore indistinguishable: zero rows, never a distinct error.

**`users` self row.** It is visible in IDENTITY and both TENANT shapes once structural validation passes, but it
does **not** require a valid tenant selection. This keeps G007's external codes: liveness inside `Resolve` for a
non-member Organization stays `context_denied`, not `unauthenticated`. Shared-scope identities additionally
require a non-NULL `rls_org_role()`.

**Statement** (`sqlInitContext`; identical text in tenancy and identity, asserted by U-RLS-BOUNDARY):

```sql
SELECT pg_catalog.set_config('app.user_id', $1, true), pg_catalog.set_config('app.organization_id', $2, true),
       pg_catalog.set_config('app.group_id', $3, true), pg_catalog.set_config('app.login_binding', $4, true)
WHERE coalesce(pg_catalog.current_setting('app.user_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.organization_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.group_id', true), '') = ''
  AND coalesce(pg_catalog.current_setting('app.login_binding', true), '') = ''
```

- One row means initialized. No row means prior authority is present: the caller closes the connection
  (`tx.Conn().Close`; pgxpool destroys it on release) and fails closed.
- Absent fields are always passed as `''`.
- Order: `BEGIN` (isolation unchanged) → context statement → protected statements → `COMMIT`/`ROLLBACK`.
- Both APIs take a `pgx.Tx`.
- **Pinned re-initializations (exhaustive).** The statement succeeds again only while all four fields are still
  `''`. Exactly two production paths rely on that:
  - **F2 sign-in race:** a rolled-back savepoint reverts the settings to all-empty, then `init(IDENTITY winner)`.
  - **F3 login start (r3):** `init(EMPTY)` (all `''`, guarded) → the E5 cleanup → `init(LOGIN b)` →
    `INSERT`. EMPTY → LOGIN is the only shape transition inside one transaction. It goes from verified-empty to a
    single non-empty shape and never back.

  Any initialization after a non-empty shape returns no row. U-RLS-BOUNDARY enumerates these two call sites and
  fails on any other double initialization.

**API:**
- `tenancy.InitContext(ctx, tx, user, org, group) error`: `invalid_id` for a zero user or Organization; a new
  closed code `context_present` for the no-row case.
- `access.inTx(ctx, write, dbContext, fn)`, called from `resolve`, `guarded` and `manageableProfiles`.
  `context_present` closes the connection and returns `unavailable`/`store_unavailable`.
- Identity has its own unexported transaction helper and initializers for EMPTY, IDENTITY and LOGIN, with the same
  discard rule.
- A failed deferred `Rollback` triggers an explicit close.
- Identity cannot import tenancy (B-MCP-SOURCE), hence the duplicated constant.

### 4.2 Policies (exact predicates, one per command except E5; all permissive, `TO rt, ow`)

Abbreviations (expanded verbatim in the migration):
- `U`, `O`, `G`, `B` = `(SELECT app.ctx_user())`, `(SELECT app.ctx_org())`, `(SELECT app.ctx_group())`,
  `(SELECT app.ctx_login_binding())`.
- `R` = `(SELECT app.rls_org_role())`.
- `MGR` = `R IN ('org_owner','org_admin')`.
- `GR(g)` = `app.rls_group_role(g)`: the caller's Group role in `g`; NULL if none, or if a Group is selected and
  `g ≠ G`.
- `LINKED(roles)` = `EXISTS (SELECT 1 FROM app.group_business_profiles a WHERE a.organization_id =
  business_profiles.organization_id AND a.business_profile_id = business_profiles.id AND (G IS NULL OR a.group_id = G)
  AND app.rls_group_role(a.group_id) IN roles)`.

| Policy | Cmd | USING | WITH CHECK |
|---|---|---|---|
| `role_bindings_select` | SELECT | `R IS NOT NULL AND organization_id = O AND source = 'membership' AND ((subject_user_id = U AND (scope_type = 'organization' OR G IS NULL OR group_id = G)) OR (scope_type = 'organization' AND MGR) OR (scope_type = 'group' AND (G IS NULL OR group_id = G) AND (MGR OR GR(group_id) IN ('group_owner','group_admin'))))` | — |
| `role_bindings_lock` | UPDATE | `R IS NOT NULL AND organization_id = O AND subject_user_id = U AND source = 'membership' AND (scope_type = 'organization' OR G IS NULL OR group_id = G)` | `false` |
| `organizations_select` | SELECT | `id = O AND R IS NOT NULL` | — |
| `organizations_update` | UPDATE | `id = O AND R IS NOT NULL` | `id = O AND R = 'org_owner'` |
| `groups_select` | SELECT | `organization_id = O AND R IS NOT NULL AND (G IS NULL OR id = G) AND (MGR OR GR(id) IS NOT NULL)` | — |
| `groups_update` | UPDATE | same as select | `organization_id = O AND R IS NOT NULL AND (G IS NULL OR id = G) AND (MGR OR GR(id) IN ('group_owner','group_admin'))` |
| `group_business_profiles_select` | SELECT | `organization_id = O AND R IS NOT NULL AND (G IS NULL OR group_id = G) AND (MGR OR GR(group_id) IS NOT NULL)` | — |
| `business_profiles_select` | SELECT | `organization_id = O AND R IS NOT NULL AND (MGR OR LINKED(('group_owner','group_admin','group_member')))` | — |
| `business_profiles_update` | UPDATE | same as select | `organization_id = O AND R IS NOT NULL AND (MGR OR ((SELECT o.group_admins_edit_business_profiles FROM app.organizations o WHERE o.id = O) AND LINKED(('group_owner','group_admin'))))` |
| `users_select` | SELECT | `id = U OR (R IS NOT NULL AND EXISTS (SELECT 1 FROM app.role_bindings rb WHERE rb.subject_user_id = users.id AND rb.organization_id = O))` | — |
| `users_insert` | INSERT | — | `id = U AND O IS NULL AND status = 'active' AND disabled_at IS NULL` |
| `users_update` | UPDATE | `id = U` | `id = U AND O IS NULL` |
| `sessions_select` | SELECT | `user_id = U` | — |
| `sessions_insert` | INSERT | — | `user_id = U AND O IS NULL AND revoked_at IS NULL` |
| `sessions_update` | UPDATE | `user_id = U` | `user_id = U AND O IS NULL` |
| `identity_principals_select` | SELECT | `user_id = U` | — |
| `identity_principals_insert` | INSERT | — | `user_id = U AND O IS NULL` |
| `login_transactions_select` | SELECT | `binding_hash = B` | — |
| `login_transactions_insert` | INSERT | — | `binding_hash = B` |
| `login_transactions_delete` | DELETE | `binding_hash = B` | — |
| `login_transactions_expire` | DELETE | `expires_at <= now() AND U IS NULL AND B IS NULL` (EMPTY shape only; r2) | — |

There are 21 runtime/owner policies and 5 helper policies (§3.1). The mappings below are tested against an
independent oracle (§8.3).

**Combined DELETE effect on `login_transactions`** (two permissive DELETE policies, ORed; the SELECT policy is also
applied whenever the statement reads columns through WHERE or RETURNING):

| Statement form | Shape | Rows deleted |
|---|---|---|
| `DELETE … WHERE … [RETURNING …]` (the callback consume) | LOGIN `b` | at most the rows of binding `b` that match the WHERE |
| `DELETE … WHERE …` | any other valid shape | none (the SELECT policy needs `binding_hash = B`, and `B` is NULL) |
| `DELETE FROM app.login_transactions` (no WHERE, no RETURNING: the E5 cleanup) | EMPTY | exactly the rows expired by the database clock |
| the same | LOGIN `b`, whether set locally or left as a session-level value (a poisoned connection) | binding `b`'s own rows, **including unexpired ones**, through `login_transactions_delete`; the expire policy is false. Production never issues the cleanup in this state: F3's `init(EMPTY)` refuses and discards the connection first |
| the same | IDENTITY, TENANT-ORG, TENANT-GROUP | none |
| any | malformed, zero, or mixed or illegal shape, with an expired candidate row present | 22P02 or 22023; nothing deleted |

**Production cleanup safety (r3).** The production cleanup runs only as F3 step 3, after F3 step 2's `init(EMPTY)`
returned one row on the same transaction and connection.
- That guard verifies all four fields were empty before it set them to `''`. Any prior authority, including a
  session-level binding, means no row: the connection is closed and the login start fails closed **before**
  step 3.
- So with production code, the "LOGIN `b`" row of this table is unreachable for the cleanup. The poisoned-connection
  claims in E5 and I4 are claims about the production path, and RLS-AUTH E5 (f) pins them.

- **Organization gate and C3:** `R IS NOT NULL` on every tenant policy, including `role_bindings`. A selected Group
  must be valid even for Organization-level reads.
  - The Organization row stays visible in a valid selected-Group context.
  - Organization owner/admin profile paths stay independent of the selected Group (`business_profiles` does not
    restrict `MGR` by `G`).
  - No capability is widened or narrowed.
- **Groups:** owner/admin see `all` (pin b: limited to the selected Group); Group roles see `own`.
- **Profiles:** `linked` covers any Group role; `linked+flag` covers owner/admin with the flag.
- **Associations:** visible exactly when their Group is. Group visibility implies visibility of that Group's
  profile, which satisfies "requires both" without a cycle.
- **Memberships (C1):** matrix `organization.manage_members` and `group.manage_members`.
  - A caller always sees its own bindings within the selected scope.
  - `org_owner`/`org_admin` also see every Organization-scope binding (an Organization capability, not limited by
    the context gate) and the Group-scope bindings of every Group, or of the selected Group only.
  - `group_owner`/`group_admin` also see the Group-scope bindings of the Groups they manage, within the selected
    scope.
  - Ordinary members see only their own.
  - Readable binding columns are exactly organization_id, subject_user_id, scope_type, scope_id, group_id, role and
    source. `id` and `created_at` are not readable.
- **Identities (C1, R6):** self, plus identities with at least one binding the caller may see under the membership
  rule. That is exactly "identities sharing a scope the caller is permitted to inspect".
  - A profile link, a co-membership without manage capability, or a mere global login reveals nothing.
  - A foreign tenant reveals nothing.
  - **Exact minimum columns** on every visible row: id, status, display_name and disabled_at.
    - The self flows need status and disabled_at (liveness, sign-in, disable).
    - PostgreSQL column privileges are role-wide, and no masking relation is added.
    - `created_at`/`updated_at` are unreadable, and `users` holds no authentication column.
- **Authentication (C2):** self only, by subject. Login transactions are visible only by own binding. In tenant
  shape, identity and session rows can be locked but never written.

### 4.3 Reader-recursion solution (R5)

Policy graph:
- `role_bindings`, `organizations`, `groups` and `group_business_profiles` call the policy_reader definers.
- `business_profiles` reads `group_business_profiles` and `organizations` (and the definers).
- `users` reads `role_bindings`.
- `sessions`, `identity_principals` and `login_transactions` read context only.

Inside a definer the reads run as policy_reader. Its own policies reference only `ctx_*` and never call a definer.
No policy reads its own relation through runtime policies, so the graph is acyclic with no 42P17 path.

This is the archetype pattern:
- narrowly scoped STABLE definers;
- a dedicated NOLOGIN, NOBYPASSRLS owner;
- SELECT-only role-specific policies;
- a pinned `search_path`;
- PUBLIC EXECUTE revoked and an exact allowlist;
- no runtime membership and no LEAKPROOF.

Bounded returns: `rls_org_role()` returns one role code or NULL; `rls_group_role(g)` returns the caller's own role
in `g` or NULL. Neither returns another subject's facts.

### 4.4 Row-lock compatibility (R8)

A lock needs one UPDATE column privilege and the SELECT plus UPDATE USING expressions. WITH CHECK applies only to
an actual UPDATE's new row.

| G007 locking statement | Lock source | Rows under RLS |
|---|---|---|
| `sqlLockLiveUser` | `users_update` USING `id = U` | self, as before |
| `sqlLockLiveSession` | `sessions_update` USING `user_id = U` | own session, as before |
| `sqlFactOrganizationShare`/`Exclusive` | `organizations_update` USING | the active Organization in a valid context |
| `sqlFact*BindingLock` (3) | `role_bindings_lock` USING (own) | own rows, as before |
| `sqlFactTargetGroupLock`, `sqlFactGroupsLock` | `groups_update` USING (read set) | readable candidates; an unreadable target is absent (decided Q8 cases, §8.4) |

- **Mutation negatives** (lock privilege never becomes a write):
  - RLS 42501: `role` (WITH CHECK false); identity or session rows in tenant shape (WITH CHECK `O IS NULL`); a
    readable-but-unauthorized name, slug, profile or flag write;
  - privilege 42501: every non-granted column, INSERT or DELETE;
  - authority columns are immutable through privileges and the 00003/00004 triggers.

### 4.5 Global identity, authentication and pre-tenant bootstrap (R6, C2)

**Bounded exceptions (exhaustive):**

| ID | Exception | Exact contract |
|---|---|---|
| E1 | Credential resolution | `app.auth_resolve_session(p_hash bytea, p_at timestamptz)` returns at most one row: session id, user id, display name, assurance, auth time, idle and absolute expiry, `renewed_at`, CSRF token and provider session id. The row exists only for an unrevoked, unexpired session with exactly `credential_hash = p_hash` whose user is active (G004's join, verbatim). Gated by possession of the secret; ignores ambient settings |
| E2 | Provider principal | `app.auth_principal_subject(p_issuer text, p_subject text)` returns the mapped user UUID or NULL, and nothing else. `(issuer, subject)` is an **opaque selector**: not a secret and not proof of authentication. The only legitimate caller is sign-in, after core has verified the provider exchange (code, PKCE, token, issuer and client checks). Ignores ambient settings (D2) |
| E3 | Logout reference | `app.auth_logout_ref(p_ref bytea, p_at timestamptz)` returns at most one (session id, user id) for an unreturned, unexpired `provider_logout_ref_hash = p_ref`. Gated by possession of the secret; ignores ambient settings |
| E4 | Login transactions | LOGIN shape with `B` = the hash of the browser's binding cookie; SELECT, INSERT and DELETE of that binding's rows only |
| E5 | Expired-login global cleanup (the one legitimate cross-binding delete) | Exactly `DELETE FROM app.login_transactions`: **no WHERE, no RETURNING**, in the **EMPTY shape**. Policy: `login_transactions_expire` USING `expires_at <= now() AND U IS NULL AND B IS NULL`. No SELECT policy applies because the statement reads no column, so it exposes nothing. It deletes only rows expired by the database clock, which are unusable by definition. **r3:** it runs only inside the F3 login-start transaction, immediately after a guarded `init(EMPTY)` on the same connection. Any prior authority (user, Organization, Group or binding, including a leftover session-level `app.login_binding`) makes that `init` return no row, and the connection is closed before the DELETE. So on the production path a poisoned connection causes **no mutation**, of expired rows or of the leaked binding's unexpired rows. Malformed prior values are refused the same way. The clock moves from the service clock to `now()` (D4) |
| E6 | Provider binding | public namespace configuration; no RLS; SELECT, INSERT (§3.1) |
| E7 | Trusted lifecycle | `DisableUser`/`EnableUser` initialize IDENTITY to the target UUID given by the trusted in-process caller. No route exists (B-MCP-SOURCE and route pins), and this is not a bootstrap path |

**Flows** (`init(X)` = `sqlInitContext` with shape X):

| Flow | Statements |
|---|---|
| F1 `Session` (`BeginRequest`, Login, Callback, Logout) | autocommit `SELECT * FROM app.auth_resolve_session($1, $2)` (E1). **If renewal is due** (≥ 60 s): `BEGIN`; `init(IDENTITY u)`; the existing renewal `UPDATE` (self); `COMMIT` |
| F2 sign-in (`Callback`) | `BEGIN` RC. Binding upsert and select (E6). `SELECT app.auth_principal_subject(…)` (E2). **Mapped or continuity:** `init(IDENTITY mapped)` → supersede (own) → `issue`. **Unmapped:** `SAVEPOINT`; `id := crypto/rand v4` (decided Q3); `init(IDENTITY id)`; `INSERT users (id, …) RETURNING id`; `INSERT identity_principals … ON CONFLICT DO NOTHING RETURNING user_id`. **On conflict:** `ROLLBACK TO SAVEPOINT` (the row and settings revert) → E2 → `init(IDENTITY winner)`. Then `issue`, `COMMIT` |
| F3 login start (**r3: one guarded transaction**) | (1) `BEGIN` RC RW. (2) `init(EMPTY)` (`sqlInitContext` with all four `''`): no row means prior authority, including a session-level binding, so close the connection and return `identity_unavailable`; **nothing has been deleted**. (3) The E5 cleanup `DELETE FROM app.login_transactions` (exact, no WHERE, no RETURNING; effective shape verified EMPTY, so it deletes expired rows only). (4) `init(LOGIN b)`, the only EMPTY → LOGIN transition (§4.1). (5) `INSERT login_transactions` (WITH CHECK `binding_hash = B`). (6) `COMMIT`. Any failure rolls back steps 3–5 together, so the cleanup and insert commit atomically, as G004's error contract requires |
| F4 callback consume | `BEGIN`; `init(LOGIN b)`; the existing `DELETE … WHERE state_hash … RETURNING`; `COMMIT` |
| F5 logout current, all, provider record | after F1: `BEGIN`; `init(IDENTITY info.UserID)`; the existing revoke or record `UPDATE`; `COMMIT`. The provider call stays between the committed revoke and the record transaction, as today |
| F6 provider-logout return | E3 lookup. None: the generic outcome. Found: `BEGIN`; `init(IDENTITY user)`; `UPDATE sessions SET provider_logout_returned_at … WHERE id = … AND provider_logout_ref_hash = … AND provider_logout_returned_at IS NULL AND provider_logout_ref_expires_at > … RETURNING provider_revocation`; `COMMIT` (exactly-once, as today) |
| F7 `DisableUser` / F8 `EnableUser` | `BEGIN`; `init(IDENTITY target)` (E7); the existing statements (lock order unchanged); `COMMIT` |
| F9 namespace check | unchanged (E6) |
| F10 `Lock/CheckLiveSession` | unchanged, inside the tenant transaction (TENANT shape; self rows) |

`lookupUser` (no caller) is re-implemented over E2. Store methods that write sessions take the subject the F1
resolution established; that is an in-process value, never input. `logout.go` passes it.

**Self-scope writes that remain** (pinned explicitly, not as authority; decided Q6):
- **Own `users` row, IDENTITY shape only:**
  - `display_name`/`updated_at`: sign-in refresh;
  - `status`/`disabled_at`: only for the trusted, route-less E7 primitives.
- **Own `sessions` rows, IDENTITY shape only:**
  - INSERT: sign-in;
  - the 8 G004 columns: renewal, revoke, supersede, provider-logout record and return.
- **Own principal INSERT:** first sign-in.
- **Own binding's login transactions:** LOGIN shape.
- **Shape restricts, it does not authorize:** an empty Organization grants nothing; it restricts these writes and
  keeps tenant-shaped transactions away from identity and authentication rows.
- **Not database-enforced (D1, corrected in r2):** two self-scope residuals remain, open to any holder of the shared
  runtime credential who initializes IDENTITY as a subject (the R3 trust limit):
  - **Self re-enable:** updating the subject's own `users.status` back to `active`.
  - **Self un-revocation:** clearing `revoked_at` on the subject's own sessions.

  These are **distinct** from self-session insertion. A disabled account's sessions, inserted or un-revoked, are
  still rejected by the active-user conditions in E1 and the G007 liveness check, so un-revocation alone restores
  nothing while the account is disabled. A self re-enable plus un-revocation or insertion would restore access.
  That is exactly the residual.

  The legitimate boundary for status changes is core's trusted, route-less lifecycle primitives (`DisableUser`,
  `EnableUser`; E7). No request path issues them. SQL cannot distinguish a trusted `EnableUser` from a raw
  self-update made with the same role and shape.

  Cross-subject reads and writes are denied (§8.3 RLS-AUTH). No mitigation or trigger is added (D1).

### 4.6 Transaction, pool and discard (R7; identity and facade alike)

The same rules cover every identity and facade transaction:
- Each transaction re-initializes all four fields, and they revert at its end.
- Prior session-level authority is rejected by the guard, and that connection is closed.
- A failed rollback is followed by an explicit close.
- pgxpool's release destroys closed and non-idle connections.

**COMMIT outcomes:**
- Facade: G007's classifier, unchanged.
- Identity: success only after `Commit` returns nil. Any other outcome is `identity_unavailable`, and when the
  outcome is not established, the connection is closed.

`unavailable` alone is never taken as proof of rollback or of safe reuse. §8.3 RLS-POOL proves rollback, or the
unknown outcome, together with discard or clean reuse, through backend identity and setup-side state.

**E1–E3 and the poison guard.** E1 and E3 are autocommit, secret-gated reads that neither read nor set context.
E2 runs inside the sign-in transaction before `init`.
- Their autocommit path does **not** run the transaction poison guard, and the Plan does not claim it does.
- Their result does not depend on ambient settings. A poisoned connection therefore cannot change what they return.
- The poison is detected, and the connection discarded, by the next initialized transaction on that connection:
  the renewal, logout, provider-return, sign-in or login-start `init`.
- An F1 resolution needing no renewal can complete on a poisoned connection without that connection being discarded
  in the same call. It returns only the row the presented secret already entitles. RLS-POOL I4 pins this.

**The E5 cleanup is not guard-free (r3).** It is a write, and it runs inside F3's transaction after the guarded
`init(EMPTY)`. The poison guard therefore always precedes it on the same connection, and a poisoned connection is
discarded before any cleanup mutation. Only the read-only E1/E3 lookups (and E2, inside sign-in) run outside the
guard. `main.go` is unchanged.

### 4.7 Ordering and commit outcome (R8, R10)

- Lock order, the linearization point and the uncertain-COMMIT contract are unchanged.
- Context statements, policy subqueries and definers take no row locks.
- Facts read by policies at the step-9 UPDATE are already row-locked by steps 3–6; the Group row is the association
  surrogate.
- Identity writers keep their order: users before sessions in `DisableUser`; supersede before `issue` in rotation.
  O-15 is unchanged.

### 4.8 Trust limit (without vacuous denials)

Settings and E1–E3 inputs transport the identity core has established; they do not authenticate a database client.

A holder of the runtime credential with arbitrary SQL can:
- read through E1/E3 what any secret it already holds entitles;
- map any provider subject it knows to that subject's user UUID through E2. The subject is an opaque selector, not
  a secret, and the provider verification happens only in core;
- initialize IDENTITY as any UUID, and so perform that subject's self-scope writes (§4.5, D1).

The required **fixed-actor** denials are nevertheless real. With a fixed subject and shape, these all fail:
- reading, joining or aggregating other subjects' authentication or tenant rows without filters;
- substituting their row IDs;
- updating other subjects' sessions or identities;
- reading login transactions of other bindings.

Each case is tested with the other user's rows present and paired positives (§8.3).

## 5. Migration and provisioning (R9)

**`00006_row_level_security.sql`** runs in one goose transaction as the owner:

1. **Preconditions** (refuse otherwise, naming the provisioning step):
   - the current user is the schema owner;
   - runtime and both helper roles exist with the exact §3.1 attributes;
   - each helper's only member is the owner, `SET TRUE, INHERIT FALSE`, no ADMIN;
   - the version-5 ACL strings are exact.
2. **Parsers:** create the four `ctx_*` functions; revoke PUBLIC EXECUTE; grant EXECUTE to runtime and
   policy_reader.
3. **Helper grants and functions:**
   - USAGE on `app` and the §3.1 column grants to the helpers;
   - a transient `GRANT CREATE ON SCHEMA app` to the helpers;
   - create the five definers, then `ALTER FUNCTION … OWNER TO` the helper;
   - `REVOKE CREATE`; revoke PUBLIC EXECUTE; grant EXECUTE as §3.1 says.
4. **Runtime grants:** the revokes and column grants of §3.1.
5. **RLS:** ENABLE and FORCE on the nine tables.
6. **Policies:** the 21 runtime/owner and 5 helper policies, with role names from `current_database()` through
   `format('%I')`.
7. **Postcondition** (refuse to commit otherwise):
   - the exact RLS flag set;
   - the policy inventory: name, table, command, permissive, roles;
   - exact ACL strings for every role and PUBLIC;
   - exact function owner, `prosecdef`, `proleakproof`, `provolatile`, `proconfig` and ACL;
   - helper role attributes and memberships;
   - **policy expression text is not checked here.** CT-RLS-2 and the behaviour tests own it (this matters for
     F-1, §9).

**Down:**
- drop all policies, the five definers and the four parsers;
- NO FORCE and DISABLE on the nine tables;
- revoke every helper and version-6 grant, including USAGE on `app`;
- restore the version-5 grants exactly.

The helper roles survive as cluster objects holding nothing in `app`. MG-RLS-5 asserts the version-5 catalog is
exact. Down is development and test only; production down stays refused.

**Data, locks and deploy order:**
- No table, column, constraint, index or data change.
- Catalog-only, with no rewrite or scan, but `ACCESS EXCLUSIVE` on each of the nine tables for the migration
  transaction. Run it in a quiescent window; MG-RLS-3 records the observed duration.
- **Deploy order:** (1) upgrade provisioning (helper roles); (2) migrate to version 6 and deploy the G008 code as one
  step, before serving traffic. Rollback is not a production path.
- **Why one step:** the G008 identity code calls E1–E3, which exist only at version 6. This replaces r0's claim that
  the code could be deployed first.
- **Mismatches fail closed:**
  - MG-RLS-5b proves that G008 code against version 5 fails closed: the E1–E3 calls fail, so no session resolves and
    no data is returned.
  - Pre-G008 code against version 6 also fails closed: its raw identity and tenant SQL sees no rows under RLS.
  - Neither case can expose data.

**Provisioning:**
- **`10-roles.sh` (first start):** adds the two helper roles and the owner's `SET`-only memberships. It is a
  rendered file, so `render-customizations.tsv` gains a row and `run.sh` APPLICABLE lists it (the G007/G051 O7
  precedent).
- **`provision-roles.sql`** (portable, idempotent, `\getenv` passwords; the harness streams it over stdin):
  - **Equivalence:** the same roles, attributes, memberships, database and `public` revokes, `app` schema, default
    privileges and search paths as `10-roles.sh`.
  - **Existing volumes:** run as superuser or admin before migrating. It adds only the missing helper roles and
    memberships when every other role already matches; drift fails.
  - **Repeat:** no change when the catalog is exact.
  - **Failure** exits non-zero before any change: a missing password variable, or an existing role with an unsafe
    attribute (SUPERUSER, BYPASSRLS, CREATEROLE, LOGIN on a helper, extra members).
  - **Non-superuser admin (bounded difference):** the admin is a CREATEROLE database owner. It temporarily grants
    itself `SET` on `<db>_owner` for `CREATE SCHEMA … AUTHORIZATION` and owner default privileges, and the script
    **revokes that grant before it exits**. The remaining difference is PostgreSQL's automatic ADMIN-only grant
    (INHERIT false, SET false) to the creator of each role.
    - PROV-1 compares every membership among the application and helper roles exactly, excluding grantor.
    - PROV-1 also asserts that the admin's residual is ADMIN-only.
    - Runtime, retention and migrator restrictions are identical.

## 6. Query, index and lock cost (PG-001, PG-002; the basis of the owner's Q7 disposition)

**r3 changes one cost: login start rises from +3 to +4 round trips,** because F3 becomes a single guarded
transaction with an explicit `init(EMPTY)` before the cleanup.
- The cleanup's expired-row locks are now held until that transaction commits.
- Every other row of the table is unchanged; the r2 precision fixes changed wording and ACL inventory only.
- The owner's "Proceed with measurements" disposition (§12) covers renewal, login, logout, consume and Enable at
  "generally +3 round trips". One extra round trip on the rare login-start path does not materially change that
  accepted assumption, so it is recorded here and in the Result, with no escalation.

| Operation | Today | r1/r2 increment |
|---|---|---|
| `Session` resolve, every authenticated request (F1, no renewal due) | 1 autocommit statement (join) | **still 1 round trip:** one call to a plpgsql definer over the same join. Plus function-call overhead; no transaction |
| Renewal (≤ once per 60 s per session) | 1 autocommit `UPDATE` | `BEGIN` + context + `UPDATE` + `COMMIT`: **+3 round trips**. The renewal row lock is held one round trip longer |
| Guarded read / `Resolve` (≤ 8 statements) | — | +1 context statement. Per statement: 4 parser InitPlans, plus the `R` InitPlan (one definer call, 1–3 index probes) |
| Guarded write (≤ 9 + BEGIN/COMMIT) | — | +1 statement. Per candidate Group or association row: one `rls_group_role` definer call (1 index probe). Profile write: `LINKED` per association row, plus the flag lookup in WITH CHECK |
| Organization-only profile op (n = subject's Group bindings) | reads 5 + 3n rows, locks 4 + 2n | ≤ 2n extra definer calls; no extra lock |
| Sign-in | 1 transaction | +1 statement (+1 E2 call; +2 statements on the race path) |
| Login start (r3) | 2 autocommit statements (cleanup, insert) | one guarded transaction: `BEGIN`, `init(EMPTY)`, cleanup `DELETE`, `init(LOGIN)`, `INSERT`, `COMMIT` = 6 round trips, **+4** (r1/r2 said +3). Locks on the expired rows being deleted are held until `COMMIT`, about 3 round trips longer than the old autocommit cleanup. A concurrent login start's cleanup, or a consume of the same expired row, may wait that long, then finds the row gone |
| Callback consume | 1 autocommit statement | one transaction: +3 round trips |
| Logout writes / provider-logout return | 1 autocommit each | +3 round trips each (+1 E3 call) |
| `DisableUser` / `EnableUser` | 1 transaction / 1 autocommit | +1 / +3 round trips |

**Indexes:** existing indexes serve every probe, so **no new index**.
- `role_bindings_subject_scope_key`: the definers.
- `groups_pkey`: selected-Group validity.
- `group_business_profiles_business_profile_idx` and `_group_key`.
- The sessions `credential_hash` unique index: E1. The `provider_logout_ref_hash` unique index: E3.
- The principals primary key `(issuer, subject)`: E2.
- The sessions `user_id` index: logout-all and disable.

**Definers are not inlined:** each call runs a cached plpgsql plan.

**Locks:** none added. RLS removes the exclusive Group lock that a subject unable to read the Group used to take.
PG-001's remaining denied-exclusive-write contention (readable-but-unauthorized Group writes, non-owner policy
writes) is unchanged. No new deadlock edge.

**Pool:** each round trip added above extends how long the operation holds a connection:
- guarded transactions: +1 round trip;
- renewal, logout, consume and Enable: ~+3;
- login start: ~+4.

**No pool-pressure guarantee is claimed.**

**Unknown workloads:**
- Group counts per Organization and per subject;
- members per scope;
- request rates and the renewal share;
- concurrency;
- latency targets.

**Bounded evidence** (TestAccessFootprint's fixture; no benchmark, target or tuning):
- **E-RLS-1:** `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` summaries for every G007 statement (both forms), the four
  writes, F1 and the F5 writes.
- **E-RLS-2:** locked-row inventory = E-PG-2 minus the enumerated RLS-hidden rows.
- **E-RLS-3:** statement and definer-call counts per operation. Calls are counted through `pg_stat_user_functions`
  with `track_functions = 'all'`, which `setupConn` sets on the disposable database.

**Escalation triggers** (stop and report to Codex and the owner before any mitigation):
- an uncorrelated parser or `R` expression running per row (a SubPlan);
- definer calls exceeding candidate rows × 2;
- a policy subplan scanning beyond the subject's own or managed rows where an index exists;
- any 42P17;
- E-RLS-2 drift;
- a lock wait outside the declared order.

Small-fixture sequential scans already disposed under PG-001 are "known". The Result links PG-002, and Codex
appends to the register.

## 7. Contracts and interfaces

- **Public:** no HTTP, BFF, MCP, route or response change (the five auth routes keep their behaviour). `main.go`,
  its imports and routes are unchanged.
- **Internal Go:**
  - `tenancy.InitContext` and the code `context_present`;
  - the `access.inTx` signature (unexported);
  - identity store transactions, E1–E3 call sites and subject parameters (unexported);
  - an unexported minimal DB interface for the identity store (`*pgxpool.Pool` satisfies it; `NewStore` is
    unchanged), so identity transactions get G007's hook-based fault injection;
  - first-sign-in UUID generated in core (decided Q3).
- **Unchanged:** facade signatures, error codes, the internal reason set and `events.json`.
- **Documentation:**
  - new `docs/foundation/DATABASE-ISOLATION.md`: shapes, policies, helpers, E1–E7, residual self writes, trust
    limit, provisioning and existing-volume upgrade, pool lifecycle, G009/G010 handoff;
  - annotations in AUTHORIZATION, TENANCY, MIGRATIONS (version 6, roles, upgrade, down), IDENTITY and FOLLOW-UP.

## 8. Tests and evidence (designed now; implemented and run only after ACCEPT)

### 8.1 Connections (R2; decided Q4)

- **`runtimeConn`/`runtimePool`:** direct runtime login, attested before any assertion:
  - `session_user = current_user = <db>_runtime`;
  - super, bypass, createrole and createdb all false;
  - no membership in owner, migrator, retention or either helper.
- **`setupConn`:** attested `rolsuper`; used only for fixtures, conflicting writers, snapshots and controls. Every
  snapshot asserts its fixture rows exist (anti-vacuity).
- **`ownerConn`:** rolled-back controls, TRUNCATE resets and CT-RLS-6 only.
- **Privileged-assertion negative:** an isolation assertion on `setupConn` must be refused by attestation.

### 8.2 Unit (`unit` suite)

**U-RLS-BOUNDARY** (`access/boundary_test.go`, over core's non-test Go):
- `set_config` appears only in the two identical `sqlInitContext` constants, always with `true`;
- `inTx` calls `InitContext` before `fn`;
- every identity statement on an RLS table runs inside the identity transaction helper after an initializer, or is
  one of the read-only E1–E3 calls (an enumerated allowlist);
- the E5 cleanup statement appears only in F3, immediately after `init(EMPTY)` in the same transaction (r3);
- double initialization appears only at the two pinned sites: F2's savepoint path and F3's EMPTY → LOGIN;
- no production SQL contains `SET app.`, `RESET`, `row_security`, `SET ROLE` or `BYPASSRLS`;
- `SECURITY DEFINER` appears only in 00006;
- no bootstrap API accepts a user UUID.

Tenancy U-BOUNDARY: +1 statement.

### 8.3 New database suites (migrations suite, version 6)

The independent oracle for C1, C3 and the AC5 matrix is the delegate-authored `rls_expected_test.go`, derived from
Spec R3–R6, ADR-004 and the AUTHORIZATION.md matrix. It is not derived from Plan prose or 00006. Every denial has a
paired positive, the pinned mechanism (SQLSTATE or zero rows) and an unchanged setup hash.

**B-RLS-CATALOG** — `TestRLSCatalog`:
- CT-RLS-1: relations unchanged (35); RLS flags on exactly the nine tables; invoker views.
- CT-RLS-2: the 26-policy inventory with normalized expression text; the §4.3 dependency graph.
- CT-RLS-3: ACLs for every role and PUBLIC, including exact **column** sets for:
  - runtime on `users` and `role_bindings`;
  - policy_reader on `role_bindings` and `groups`;
  - auth_reader: the literal 14-column `sessions` set, `users` (id, status, display_name) and
    `identity_principals` (issuer, subject, user_id), with no table-level privilege.

  These use the existing exact ACL-string check. Also covered: the function allowlist and attributes, and the
  sequence set.
- CT-RLS-4: role attributes and memberships, including the helpers.
- CT-RLS-5: runtime denials (42501):
  - DDL, including `DISABLE`/`NO FORCE` and `CREATE`/`ALTER`/`DROP POLICY`;
  - `SET ROLE` to owner, migrator or a helper;
  - TRUNCATE on all 10 tables;
  - metadata reads;
  - `row_security = off` followed by a protected read;
  - every denied INSERT, DELETE and column UPDATE.
- CT-RLS-6, owner FORCE: positive in context A; foreign denied; no context gives zero rows while setup shows rows;
  INSERT is an RLS violation; DELETE affects 0 rows.
- CT-RLS-NEG (rolled-back owner controls, each yielding exactly its detector code, catalog unchanged after):
  - N1 `NO FORCE` on `groups`; N2 `DISABLE` on `sessions`;
  - N3 an extra `USING (true)` runtime policy; N4 a `FOR ALL` policy;
  - N5 `GRANT INSERT` on `role_bindings`; N6 PUBLIC EXECUTE on `rls_org_role`;
  - N7 a helper with LOGIN; N8 runtime granted membership in policy_reader;
  - N9 a definer without `search_path`; N10 a view with `security_invoker = false`;
  - N11 runtime SELECT on `users.created_at`; N12 auth_reader SELECT on `sessions.revoke_reason`;
  - N13 `login_transactions_expire` recreated without its `U IS NULL AND B IS NULL` shape terms (the r1 defect).

**B-RLS-MIGRATION** — `TestRLSMigration`, persistence seed/verify, and the harness `rls_provision` with
`TestRLSProvisioned`:
- MG-RLS-1..4: empty to 6; repeat; populated version 5 to 6 with all-table row hashes identical; reapply.
- MG-RLS-5: down to 5 with data restores the exact version-5 catalog with rows unchanged; up again.
- MG-RLS-5b: G008 code at version 5 fails closed (identity lookups unavailable, no data), proving the pinned
  same-step deploy order.
- MG-RLS-6: restart persistence.
- PROV-1: a no-init instance and a non-superuser admin give catalog equivalence plus the ADMIN-only residual.
- PROV-2/3: migrate; then representative runtime positives and negatives per protected table, plus owner FORCE.
- PROV-4: a repeat run is a no-op.
- PROV-5: a missing password, an unsafe runtime role and a helper with LOGIN each exit non-zero with no change.
- PROV-6, existing-volume upgrade: an instance initialized by the version-5 `10-roles.sh` (base bytes over stdin)
  gets only the helper roles added; then migrate.

**B-RLS-IT** — `TestRLSIsolation`, `TestRLSPool`, then logcheck on the access log and a canary scan of the
PostgreSQL log.

The world:
- Organizations A and B; Groups A1, A2, A3 and B1; profile PA shared by A1 and A2; PA3 on A3; PAu unlinked; PB.
- `oOwnerA`; `oAdminA` (also member of B and B1); `oMemberA`.
- `gOwnA1` (group manager of A1); `gAdmA2` (admin of A2 and B1); `gMemA3`.
- `groupOnly` (an A1 binding without an Organization binding); `foreign`.
- Each actor holds a real G004 session, a principal and a pending login transaction, so other users' rows are
  always present.

Scenarios:
- **RLS-CTX (C3, AC3).** Each field varied independently (malformed, zero, foreign, absent) across:
  - reads of all nine tables and three views;
  - every runtime-permitted write: Group name and slug, the profile, the flag, `users` self, `sessions` self, login
    transaction insert and delete.

  Also covered: every invalid shape, the partial cases and an inconsistent Group (foreign or unselectable).
  - Expected (pinned per case): 22P02 or 22023 when a candidate row exists, otherwise zero rows; a data-dependent
    invalid Group gives zero tenant rows. Rows are always unchanged.
  - The E5 cleanup is one of the covered writes; its per-field and per-shape expectations are the RLS-AUTH E5
    controls.
  - Positives: IDENTITY self; TENANT-ORG; TENANT-GROUP with Organization-row access and owner/admin profile authority
    under a valid Group.
  - Also: autocommit `set_config(…, true)` is ineffective; G007 IT-CTX-1..10 at version 6.
- **RLS-ORG (AC4):** fixed `oAdminA` in context A with no WHERE:
  - exact sets for SELECT, aggregates, joins (`groups ⋈ group_business_profiles ⋈ business_profiles`,
    `role_bindings ⋈ users`), subqueries and `UPDATE … RETURNING`;
  - foreign UPDATE affects 0 rows; INSERT and DELETE give 42501; paired positives.
- **RLS-SCOPE (C1, AC5).** Oracle exact sets per actor × context (Organization-only A; selected A1, A2, A3):
  - visible rows of `role_bindings`, the three views and `users` (minimum columns asserted via
    `has_column_privilege`, plus `created_at` 42501);
  - visible rows and raw UPDATE outcome (ok, 0 rows, 42501) for Groups, associations, profiles and Organizations;
  - flag on and off.

  Mandatory cases:
  - `oOwnerA` and `oAdminA` see all A bindings and identities. In selected A2 they see all Organization-scope
    bindings plus A2's Group bindings, but not `groupOnly`.
  - `gOwnA1` sees A1's Group bindings, including `groupOnly`, and those identities; not A2 or A3.
  - `oMemberA` and `gMemA3` see self only.
  - `foreign` sees nothing in A.
  - Shared PA: visible to `gOwnA1` through A1 only, revealing no A2 link, member or identity.
  - `gAdmA2`: two-Organization isolation.
  - Last-path removal: PA disappears with no profile write. Second path: PA is restored through A1 only.
  - IT-MATRIX-1 at version 6 agrees with the raw outcomes.
- **RLS-AUTH (C2, AC8).** Fixed actors in EMPTY, LOGIN, IDENTITY and TENANT shapes, with every other user's rows
  present.
  - Unfiltered reads, joins (`sessions ⋈ users`, `identity_principals ⋈ users`) and `count(*)` on `sessions`,
    `identity_principals` and `login_transactions` return own rows only (own binding for login transactions), or
    none.
  - Other users' rows give 0 rows on UPDATE, including un-revoke and expiry extension, and on DELETE.
  - In tenant shape, own session or identity UPDATE and INSERT give 42501 (WITH CHECK).
  - E1–E3 return the exact rows for exact keys and nothing for wrong keys. E1/E3 results are identical under
    EMPTY, IDENTITY, TENANT and poisoned settings, because they ignore ambient context.
  - **E5 cleanup controls (r2, extended r3).** (a)–(e) are raw-statement policy controls; (f) is the production-path
    control. Fixture: expired login transactions of two other browsers and an unexpired pending transaction of
    another user are present as candidates.
    - (a) The exact cleanup statement in EMPTY shape deletes exactly the expired rows. The other user's unexpired
      pending transaction survives (setup hash), and the statement returns and reads nothing.
    - (b) The same statement under malformed user, Organization, Group or binding values raises 22P02/22023
      (pinned per field). Under zero UUIDs it raises 22023. Under mixed or illegal shapes (Group without
      Organization, Organization without user, binding plus user) it raises 22023. Nothing is deleted.
    - (c) The raw statement in IDENTITY, TENANT-ORG or TENANT-GROUP shape, or with a valid-format session-level
      user value, deletes nothing.
      - In LOGIN shape, or with a valid-format **session-level binding `b`**, it deletes binding `b`'s own rows,
        **including unexpired ones** (the §4.2 combined table); the expire policy contributes nothing.
      - This raw-statement result is recorded as the reason production must run the guard first. It is not a
        production behaviour.
    - (d) The cleanup with a WHERE clause in EMPTY shape deletes nothing.
    - (e) After cleanup, a valid LOGIN-shape consume (F4) of an unexpired transaction still performs exactly one
      authorized consume: one row returned and deleted. A second consume returns none, and a consume under another
      binding returns none.
    - (f) **Production F3 against a poisoned connection (r3; the r2-review case).**
      - Setup: a runtime pool constrained to **one** observed backend (`pg_backend_pid()` recorded). Acquire it,
        run a session-level `SET app.login_binding = '<hex b>'` (valid format) and release it back to the pool.
      - Rows present, with their setup hash recorded:
        - an **unexpired** row whose `binding_hash = b` (the previous browser's pending login);
        - expired rows of other bindings;
        - an unexpired row of another binding.
      - Act: `identity.Service.Login` (production F3) on that pool.
      - Expected: `identity_unavailable`. **Every row is unchanged** (setup hash identical, the unexpired `b` row
        included, and no expired row deleted). The connection is closed, so the next acquisition reports a
        different backend.
      - Then a clean login start on the new backend succeeds: expired rows are deleted, unexpired rows are kept,
        and one new row exists for the new binding. Its callback consume (F4) succeeds exactly once.
      - The same case repeats with a session-level user value, and with a malformed binding value: identical
        no-mutation, discard and recovery outcomes.
  - Positives: every G004 flow before any Organization exists (first sign-in, repeat sign-in, renewal, Session
    self-read, logout current/all/provider and return, Disable/Enable, pending callback consumption). Also the
    forged-bootstrap negative, and the cancellation and optional-MFA cases at version 6 (existing G004 and IT-MFA-1).
- **RLS-IMMUT (AC6):**
  - privilege 42501 on key, parent, subject, scope and source columns; association writes; binding INSERT; the
    session columns outside the 8;
  - RLS 42501 on `role`, and on self status or session writes in tenant shape;
  - **lock positives** for every §4.4 statement;
  - constraint controls through setup with data unchanged; the CT-6 detector (no copied authority).
- **RLS-PROJ (AC7):** views against their base tables per actor; view writes give 42501 or 55000; `FOR SHARE`
  through views gives 42501; definers return only the caller's own facts; no 42P17 anywhere (tracer).
- **RLS-POOL (AC9, identity included).** A one-connection pool with `pg_backend_pid()` recorded:
  - facade P1–P10 as in r0: tenant switches, Group clear, commit, rollback, statement error, context-init failure,
    cancellation, lost COMMIT reply, poisoned session setting (the unsafe control leaks; the boundary refuses and
    discards), release inside a transaction;
  - **identity I1–I5**, through the new store DB interface and G007's hook:
    - I1 statement error in sign-in or renewal: no session issued, rows unchanged, clean reuse;
    - I2 cancellation in sign-in, consume and disable: no success; discard or clean reuse with empty settings,
      recorded;
    - I3 lost COMMIT reply in sign-in and disable: `identity_unavailable`, setup-side state recorded as possibly
      committed, next backend new;
    - I4 poisoned connection:
      - a non-renewing E1 on it returns only the presented key's row and does **not** discard it (no guard runs on
        that autocommit path);
      - the next initialized identity transaction on it (renewal, logout, sign-in or login start) refuses at
        `init`, the connection is discarded, and the next backend is new;
      - **production login start on it refuses at `init(EMPTY)` before the cleanup DELETE.** Every
        `login_transactions` row stays unchanged, including the poisoned binding's unexpired rows (RLS-AUTH
        E5 (f));
    - I5 a rollback failure is followed by a close.
- **I-MAP-CONC extension (decided Q3):** the existing 32-way first sign-in additionally asserts:
  - exactly one user and one principal (setup);
  - every loser's settings are back to `''` after `ROLLBACK TO SAVEPOINT`, read by a hook before the winner init;
  - every loser's issued session belongs to the winner subject.

**B-RLS-ORDERING** — `TestRLSOrdering`, with G007's deterministic barriers, orders A, B and C:
- OR-1: profile write against last-path removal.
- OR-2: profile write against unlink.
- OR-3: profile write against the flag being turned off.
- OR-4: old context, connection or prepared statement after each committed change: downgrade, removal, unlink,
  flag, disable, session revoke, expiry, manager demotion (C1 visibility shrinks immediately).
- OR-5: real slug-change lock upgrade.

O-1..O-15 and IT-FACADE/IT-REVOKE rerun at version 6.

### 8.4 Mechanical reconciliation and decided Q8 cases

| # | Site | Change |
|---|---|---|
| RC-1 | B-MIGRATE-SET | add the 00005 digest pin; set includes 00006 |
| RC-2 | B-MIGRATE-CLI/CATALOG | version 6; reapply and rollback of 00006; 35 relations |
| RC-3 | root migrate, B-PERSIST, B-DBRENAME, B-TOGGLE-OFF | version-6 texts. DBRENAME and OFF get the helper roles from the customized `10-roles.sh` |
| RC-4 | B-IDENTITY-CATALOG / `TestIdentityCatalog` | final identity and authentication contract; "version 6" |
| RC-5 | `TestMigrationLifecycle` | includes 00006 |
| RC-6 | `TestPolicyCatalog`/`Migration`/`Persist*` | pinned to subsetFS 5 |
| RC-7 | `TestTenancyModel`, `TestBusinessProfileModel` | run at version 5 after a reset (meaning preserved) |
| RC-8 | B-POLICY-IT/ORDERING (`TestAccessPolicy`/`Footprint`/`Ordering`) | fixtures, writers and owner reads move to `setupConn`; statement counts +1; the Q8 cases below; external codes, E-PG baselines and O-cases unchanged |
| RC-9 | identity tests (B-IDENTITY-IT, version 6) | owner counts move to `setupConn`; raw auth-table fixtures go through identity flows or setup; anti-vacuity |
| RC-10 | harness env (`itenv`, `identity_it`) | `DB_SUPERUSER`/`DB_SUPERUSER_PASSWORD` (0600, removed after use) |
| RC-11 | `core/migrations/migrations_test.go` | includes 00006 |
| RC-12 | `manual_qualification_test.go` (`workos_manual`; not run in G008) | its observer reads other users' `users` and `sessions` through raw runtime SQL, which version 6 correctly denies. Add a startup guard so it **fails closed** when those tables have RLS, rather than recording empty evidence. Reconciling the observer for future identity qualification needs its own authorization, and is listed as a handoff item. Historical G004 evidence is unchanged |
| RC-13 | `run.sh` APPLICABLE; `required-checks.tsv`; `render-customizations.tsv` | new paths and `10-roles.sh`; 4 IDs (95→99; full 88→92); the customization row |

**Decided Q8, exhaustive pinned cases** (external code unchanged, log-only reason shifts):
1. Group write by a subject who cannot read the target: `role_not_permitted` → `target_not_found`.
2. Group operation in selected-Group context against another Group: `target_outside_context` → `target_not_found`.
3. Group read by an `org_member` without a Group binding: `role_not_permitted` → `target_not_found`.
4. Profile operation on a profile invisible to the subject: `no_linked_path`/`role_not_permitted` →
   `target_not_found`.
5. `Resolve` or a guarded operation whose selected Group is foreign, unknown or no longer selectable:
   `group_not_selectable` → `organization_not_permitted`.

Any other observed reason change is a finding.

## 9. Validation sequence, CI and hosted evidence

- **Before every batch:**
  - fetch, then ancestor check or rebase-and-reconcile;
  - `bash infra/ci/docker-desktop-mounts.sh`;
  - `flock` on `local-harness.lock`;
  - record the target and candidate SHAs;
  - never edit during a run.
- **Wrapper, unchanged:** `bash infra/validation/run-approved.sh --snapshot <S> --snapshot-sha256 <D> --count <N>
  --path-re '<RE>' --plan management/v0/plans/G008-postgresql-rls-plan.md -- --base <BASE> --allow-file
  management/v0/results/G008-scope-allowlist.txt --report-dir <EVID>/<step> <suites>`.
- **Batches:** finish, fix, rerun, advance. Focused SQL, pool and identity reproduction precedes broad reruns.
- **CI profile: full.** All changed paths are already foundation; no `profile-paths.tsv` change.
- **No manual or operator-feedback stage.** No WorkOS or protected action.
- **F-1 (corrected), forced local failure:**
  - In a scratch clone, `groups_select` loses `(G IS NULL OR id = G)`. The migration's postcondition does not check
    expression text, so 00006 applies.
  - The expected result is **B-RLS-CATALOG FAIL** (CT-RLS-2 expression mismatch for `groups_select`) and **B-RLS-IT
    FAIL** (RLS-CTX and RLS-SCOPE selected-Group cases), then B-REQUIRED-CHECKS FAIL, with a non-zero harness exit.
  - No check is required to run against a failed migration.
- **Final:**
  - `W(all)` and `W(unit)` on the rebased candidate: every full ID plus B-UNIT-CHECKS;
  - MR `fast`;
  - staging `gate` under the owner status rule.

## 10. Files and scope allow-list

**Create:**
- `core/migrations/00006_row_level_security.sql`;
- `core/internal/migrate/rls_integration_test.go`;
- `core/internal/tenancy/access/{rls_integration,rls_ordering_integration,rls_expected}_test.go`;
- `infra/postgres/provision-roles.sql`;
- `docs/foundation/DATABASE-ISOLATION.md`;
- `management/v0/results/G008-{scope-allowlist.txt,postgresql-rls-result.md}`.

**Modify:**
- `core/internal/identity/{store,service,logout}.go` and identity `*_test.go`;
- `core/internal/tenancy/{store,facts,errors}.go`;
- `core/internal/tenancy/{boundary,integration,business_profile_integration,errors}_test.go`;
- `core/internal/tenancy/access/{tx,context,facade}.go`;
- `core/internal/tenancy/access/{boundary,integration,footprint_integration,ordering_integration}_test.go`;
- `core/internal/migrate/{integration,policy_integration}_test.go`;
- `core/migrations/migrations_test.go`;
- `infra/postgres/init/10-roles.sh`;
- `infra/validation/{run.sh,required-checks.tsv,render-customizations.tsv}`;
- `docs/foundation/{AUTHORIZATION,TENANCY,MIGRATIONS,IDENTITY,FOLLOW-UP}.md`.

**Untouched:** migrations 00001–00005, `core/cmd/**`, web, intelligence, contracts, `infra/ci/**`,
`infra/compose/**`, `events.json`, `go.mod`/`go.sum`, `skills/**`, governance, ADRs, Goal/Spec,
ROADMAP/PROGRESS/PERFORMANCE and SECURITY.

```text
^core/migrations/(00006_row_level_security\.sql|migrations_test\.go)$
^core/internal/identity/(store|service|logout)\.go$
^core/internal/identity/[a-z_]+_test\.go$
^core/internal/tenancy/(store|facts|errors)\.go$
^core/internal/tenancy/(boundary|integration|business_profile_integration|errors)_test\.go$
^core/internal/tenancy/access/(tx|context|facade)\.go$
^core/internal/tenancy/access/(boundary|integration|footprint_integration|ordering_integration|rls_integration|rls_ordering_integration|rls_expected)_test\.go$
^core/internal/migrate/(integration|policy_integration|rls_integration)_test\.go$
^infra/postgres/(provision-roles\.sql|init/10-roles\.sh)$
^infra/validation/(run\.sh|required-checks\.tsv|render-customizations\.tsv)$
^docs/foundation/(AUTHORIZATION|TENANCY|MIGRATIONS|IDENTITY|FOLLOW-UP|DATABASE-ISOLATION)\.md$
^management/v0/results/G008-(scope-allowlist\.txt|postgresql-rls-result\.md)$
```

## 11. Acceptance-criteria mapping

| AC | Scenarios → expected outcome → evidence |
|---|---|
| AC1 | CT-RLS-1/2/3, NEG N1–N4, N10, N11 → exactly §3.1 and §4.2 (26 policies, 9 RLS tables, column ACLs) → B-RLS-CATALOG |
| AC2 | CT-RLS-4/5/6, §8.1 attestation and the privileged-assertion negative, helper role checks, NEG N7/N8 → runtime never super, bypass, owner or helper member; owner FORCE → B-RLS-CATALOG |
| AC3 | RLS-CTX (every field and shape), G007 IT-CTX, U-RLS-BOUNDARY → pinned SQLSTATE or zero rows, rows unchanged, valid-shape positives, no bypass → B-RLS-IT, unit |
| AC4 | RLS-ORG → exact A-only sets; foreign writes 0 rows or 42501; paired positives → B-RLS-IT |
| AC5 | RLS-SCOPE (oracle) and IT-MATRIX-1 → Group, profile and policy outcomes; two Organizations; last and second path; sibling isolation → B-RLS-IT, B-POLICY-IT |
| AC6 | RLS-IMMUT → lock positives; role, status and session manipulation 42501; constraints; no copied authority → B-RLS-IT |
| AC7 | RLS-PROJ, CT-RLS-2/3, NEG N6/N9/N10 → views equal base; definer allowlist, `search_path`, bounded returns; 0 × 42P17 → B-RLS-CATALOG, B-RLS-IT |
| AC8 | RLS-SCOPE identity sets and minimum columns, RLS-AUTH, B-IDENTITY-* at version 6 → scoped identities; fixed-actor authentication privacy; every G004 flow before an Organization; forged bootstrap, cancellation and MFA unchanged → B-RLS-IT, ops |
| AC9 | RLS-POOL P1–P10 and I1–I5 → no carried authority; unsafe control detected; rollback, discard or unknown outcome proven → B-RLS-IT |
| AC10 | OR-1..5, O-1..15, IT-FACADE/IT-REVOKE → serialized-before or observed-and-denied; surviving paths; slug upgrade; uncertain COMMIT → B-RLS-ORDERING, B-POLICY-ORDERING |
| AC11 | MG-RLS-1..6 and 5b, PROV-1..6, B-MIGRATE-* → data unchanged; exact down; no-init equivalence with the ADMIN-only residual; existing-volume upgrade → B-RLS-MIGRATION |
| AC12 | log and canary scans, B-SECRET-CANARY, inherited full suites, B-TENANCY-HTTP, B-MCP-SOURCE → no public, storage, audit, product or vendor expansion → final W(all) |
| AC13 | final W(all) + W(unit); corrected F-1; B-CI-CONTROLS; MR fast; staging rule → Result |
| AC14 | binding, drift, allow-list, cleanup, attempt ledger and checkpoint, G009/G010 handoff, Codex ACCEPT → Result |

## 12. Decisions and the owner gate

**Settled by the r0 review, encoded here:**
- **Q3:** the first-sign-in user UUID is a core crypto v4. UUIDs stay canonical, no identity is caller-supplied, and
  the savepoint conflict behaviour holds; the I-MAP-CONC extension proves it.
- **Q4:** a separately attested disposable superuser handles setup; runtime assertions use a direct login, and the
  privileged-assertion negative is kept.
- **Q5:** unsupported runtime structural mutations are denied. The eight guarded operations and trusted setup stay;
  no G018/G019 workflow is invented.
- **Q6:** the identity-only shape is legitimate but is neither an authentication nor an admin boundary. Its writes
  are pinned in §4.5, and an empty Organization implies no grant.
- **Q8:** external codes are unchanged; the §8.4 list of five cases is exhaustive.
- **Q1/Q2:** resolved by the C1/C2 design in §4.

**Bounded interpretations D1–D4.** The r1 review accepted these subject to its corrections, which r2 applies. None
is a Spec or security-model change, a waiver of cross-user material protection, or an administrative bypass.

| ID | Point | Pinned interpretation (r2) |
|---|---|---|
| D1 | Self-scope residuals are not database-enforced (§4.5) | Two **distinct** residuals remain for a holder of the shared runtime credential acting as a subject in IDENTITY shape: (a) **self re-enable** of `users.status`; (b) **self un-revocation** of that subject's own sessions. Neither is equivalent to self-session insertion: E1 and the G007 liveness check reject every session of a disabled account, so (b) or an insertion alone restores nothing while the account is disabled, and (a) is the step that would restore access. The legitimate boundary for status changes is core's trusted, route-less `DisableUser`/`EnableUser` (E7). No request path issues them, and SQL cannot distinguish them from a raw self-update by the same role and shape. Cross-subject access stays denied. No trigger or other mitigation is added |
| D2 | E2 maps a provider subject to a user UUID for any caller with runtime access | `(issuer, subject)` is an **opaque selector**: not a secret, and not proof of authentication. Provider verification (code exchange, PKCE, signed token, issuer and client checks) happens in core before the sign-in path calls E2, which only returns the mapped UUID or NULL, with no authentication material. It ignores ambient settings. This is R6's pre-identity principal bootstrap, the narrowly defined authentication operation R5 permits. The alternatives are database-side token verification (an excluded framework) or unprotected principals (which fail C2) |
| D3 | The auth_reader helper policies are `USING (true)` | Scoped TO a NOLOGIN, NOINHERIT role whose only member is the owner, with the literal column grants of §3.1 and no table-level privilege. It is reachable only through E1–E3's fixed SQL. E1 and E3 are secret-possession reads returning ≤ 1 row; E2 returns ≤ 1 UUID for an exact selector |
| D4 | Expired-login global cleanup (E5) | The only cross-binding delete. Exactly `DELETE FROM app.login_transactions` with no WHERE and no RETURNING, **in the EMPTY shape only**, through `login_transactions_expire` USING `expires_at <= now() AND U IS NULL AND B IS NULL`. Whole-shape parsers reject malformed, zero and illegal shapes. The expire policy deletes nothing in any valid non-EMPTY shape. **r3:** production runs the cleanup only inside F3, after a guarded `init(EMPTY)` on the same transaction and connection. Any prior authority, including a leftover session-level binding whose own-binding DELETE policy would otherwise apply, is refused and discarded before the DELETE. It deletes only rows expired by the database clock and reads nothing. The clock moves from the service clock to `now()`; consumption expiry stays on the service clock |

**Q7 (PG-002) — decided by the owner: "Proceed with measurements."** Recorded in
the owner disposition (not included) (SHA-256
`22d6cbfa6e0e977892386f0274731e775744c91f14d08b672898972b5b35ad5b`, relayed by Codex) and PERFORMANCE.md.

- **What was accepted:** a risk disposition for the r1 design costs, not measured performance acceptance. The
  costs, unchanged in r2 and corrected in r3 for login start only (§6):
  - non-renewing `Session` resolution stays at 1 round trip;
  - guarded tenant transactions add +1 round trip plus policy and helper work;
  - renewal, logout, callback consume and Enable add their pinned costs, generally +3 round trips (Disable +1;
    provider-logout return +4);
  - **login start: +4 (r3; was +3).** This comes from the guarded `init(EMPTY)` the r2 review required. It does
    not materially change the accepted assumption, so it is recorded here without escalation;
  - connections and locks may be held longer;
  - workload, fairness, latency and pool limits remain unknown.
- **Required measurements:** the bounded E-RLS-1..3 query, index, lock, statement and helper-call measurements, with
  recorded parameters, environment and evidence, in the Result and linked from PG-002.
- **Escalation:** unexpected scans, fan-out, lock behaviour, helper call counts, or any feasibility-threatening cost
  (§6 triggers) stops work for review.
- **Not authorized:** a numeric target, a benchmark campaign, a new optimization, weakened security or extra scope.
- **Changed assumptions:** a materially changed cost or design assumption returns for review, rather than silently
  extending this decision.
- **Still gating implementation:** Codex Plan ACCEPT and a fresh committed execution and input binding.

**Remaining genuine conflicts: none.** No Spec or security-model change is proposed.

## 13. Sequence, delegation, budget and stops

| Step | Work (tests land with their change) | Batch |
|---|---|---|
| S0 | Start gate (§1) | — |
| S1 | `InitContext`, `inTx`, identity store transactions, DB interface, E1–E3 call sites, UUID; U-RLS-BOUNDARY | `W(unit)` |
| S2 | 00006; `10-roles.sh` and provisioning; B-RLS-CATALOG/MIGRATION; RC-1..RC-13 | `W(unit lint controls provenance migrations)`, then `W(ops)` |
| S3 | B-RLS-IT, integrating the oracle; E-RLS-1..3 | `W(migrations)`; trigger check |
| S4 | B-RLS-ORDERING | `W(migrations)`, then `W(lint controls contracts provenance root ops rename off)` |
| — | **Checkpoint after attempt 4,** if reached | below |
| S5 | Read-only security reviewer; bounded corrections | covering suites |
| S6 | Documentation | `W(lint provenance migrations off)` |
| S7 | Corrected F-1 in a scratch clone | the expected FAILs recorded |
| S8 | Rebase and reconcile; final `W(all)` + `W(unit)` | all full IDs pass |
| S9 | Target re-check; X1 push and MR; `fast`; Result | `fast` PASS |
| S10 | Codex review → X2 conditional integration → staging → Result | owner status rule; Codex ACCEPT |

**Delegation:**
- **Oracle delegate (high):** owns only `rls_expected_test.go`. It works from Spec, ADR-004 and AUTHORIZATION.md
  only, with no access to 00006 or Plan predicates.
- **Read-only security/SQL reviewer (xhigh)** after S4.
- **Declined:** a separate SQL implementation delegate.
- **Effort:** implementation defaults to high, with owner-selected xhigh suggested for S2–S4. A reminder precedes
  each phase, and actual effort is reported or recorded as unknown.

**B7: at most seven shared attempts, checkpoint after attempt 4.** The reforecast after C1/C2 is **3–5 attempts**
(r0 forecast 2–4). The added scope:
- two helper roles and five definers;
- three more RLS tables;
- identity flow transactions;
- the `10-roles.sh` customization and the upgrade path;
- larger oracles.

All of it stays within B7, with no weakening or deferral.

**Checkpoint:** if identity and authentication (C2, AC8) is not converging, propose **G008a** (tenant RLS, C1, C3,
pool, ordering) and **G008b** (authentication RLS, E1–E7, provisioning). No split now, because R6 couples them.

**Stop and escalate:**
- any need for BYPASSRLS, a write-capable definer, an owner or superuser connection in core, or a session-level
  setting;
- 42P17;
- a predecessor test changing meaning beyond §8.4;
- a writer or lock-order change;
- an E-RLS trigger, or any index or schema tuning;
- a D1–D4 disposition requiring a Spec or security-model change;
- harness work beyond §8.4;
- quota or access stops;
- B7 exhaustion or thrashing.

**External actions:**
- **X1:** push and MR after Codex's final pre-integration review (standing authority).
- **X2:** conditional integration, owner-authorized only.
- No other hosted or provider action.

**G009/G010 handoff:**
- storage metadata reuses the parsers and shapes, `R`/`GR`, the `TO rt, ow` pattern and the DAG;
- G010 reuses the B-RLS-IT world;
- G018/G019 own membership administration and provisioning flows;
- D1 is recorded;
- the RC-12 manual-qualification observer reconciliation is listed.

G010 and G021 remain open.

## 14. Revision history

- **r0** — 2026-10-13, commit `0a2ca3c4a8653360d8da1e73558a244258189eaa`, SHA-256
  `2a18da9ccd9593643ef507b8c0122c68ad6bc7c2ad519f2e8b59f719b385907f`. Archived byte-identically as
  G008-postgresql-rls-plan-r0.md (not included).
- **r1** — 2026-10-13, commit `05eefe30512d22107a8e81a409ee2bc437beff62`, SHA-256
  `9f3bc5982a25004582cc591c0d6fc0ce93767f5bd83cdb19db15b6b5d9ccfdd8`. Archived byte-identically as
  G008-postgresql-rls-plan-r1.md (not included).

  Responds to the r0 Codex review (not included) (SHA-256
  `70ae630e1e4633ef60337829a115274932c1e998adf660fe0c725ade8e0021a2`, CORRECTION LOOP, on `origin/staging`
  `32f3473`).
  - **C1:** replaces self-only design A with manager-scoped membership and identity visibility from
    `organization.manage_members` and `group.manage_members`, through the policy_reader definers. Pins the minimum
    columns: `users` (id, status, display_name, disabled_at), with `role_bindings` narrowed to 7 columns. Adds the
    independent-oracle verification (§4.2, §4.3, §8.3 RLS-SCOPE).
  - **C2:** replaces S1+ with RLS on `sessions`, `identity_principals` and `login_transactions`. Adds:
    - three SELECT-only auth_reader definers (E1–E3);
    - binding-keyed login transactions (E4);
    - the expired-only cleanup (E5);
    - the provider binding as public configuration (E6);
    - pinned self-scope writes;
    - fixed-actor RLS-AUTH verification (§4.5, §8.3).
  - **C3:** four-field shape validation in every parser, and selected-Group validity in `rls_org_role()` gating
    every tenant policy. `users` self visibility is structurally validated, with per-field RLS-CTX verification
    (§4.1, §8.3).
  - **Settled Q3–Q6 and Q8** encoded (§12): I-MAP-CONC extension; exhaustive Q8 cases.
  - **Small reconciliations:**
    - identity transaction safety evidence I1–I5;
    - the provisioning ADMIN-only residual, with the temporary owner `SET` revoked;
    - corrected F-1;
    - the deploy order corrected to same-step code and migration;
    - identity store DB interface for fault injection.
  - **Costs:** recomputed for Q7. `Session` stays at 1 round trip; PG-002's review note is recorded as an input,
    not a waiver.
  - **Forecast:** revised to 3–5 attempts.
- **r2** — 2026-10-13. Responds to the r1 Codex review (not included) (SHA-256
  `d958a43c6e4fb61d61c33b1663f917a588f67c7e643f348b846d9487e9493961`, CORRECTION LOOP, on `origin/staging`
  `590f003`). The resolved areas (C1, C2, E1–E3, C3, lock-only UPDATE, provisioning, rollback, and self reads in
  non-selectable contexts) are not reopened.
  - **Cleanup-context gap:**
    - `login_transactions_expire` becomes `expires_at <= now() AND U IS NULL AND B IS NULL`. Through the whole-shape
      parsers, this permits the cleanup only in the EMPTY shape.
    - E5 and D4 are restated as the single declared cross-binding exception: exact statement, no WHERE, no
      RETURNING, no read exposure, not covered by the poison guard.
    - Added: the §4.2 combined DELETE-effect table, the RLS-AUTH E5 controls (a)–(e) and NEG N13.
  - **Precision 1:** the literal 14-column `sessions` auth_reader grant union, plus `users` and `identity_principals`,
    with exact CT-RLS-3 ACL assertions and NEG N12.
  - **Precision 2:** D1 corrected. Self re-enable and self un-revocation are distinct residuals and are not
    equivalent to self-session insertion, because active-user checks reject a disabled account's sessions. The
    route-less trusted lifecycle boundary is stated.
  - **Precision 3:** the E2 subject is an opaque selector, not a secret or proof, with provider verification in core.
    E1/E3 are secret-possession reads that ignore settings, and their autocommit path does not run the poison
    guard. §4.6, §4.8 and RLS-POOL I4 are corrected accordingly.
  - **Q7:** recorded as the owner's decision "Proceed with measurements"
    (disposition (not included), SHA-256
    `22d6cbfa6e0e977892386f0274731e775744c91f14d08b672898972b5b35ad5b`, on `a316dd3`). Planning inputs are rebound
    to `a316dd3` (PERFORMANCE.md `49fedd5c…2ec3b`, PROGRESS `7b7598a4…c991`).
  - **No cost change.** Forecast 3–5 attempts; B7 and checkpoint 4 unchanged.
- **r3** — 2026-10-13. Responds to the r2 Codex review (not included) (SHA-256
  `18d35611a69e8e6014c44b8b4033671958677aac03534c8f4e462428d47b2a2a`, CORRECTION LOOP, on `origin/staging`
  `3068d98`): one cleanup/pool safety case.
  - r2 is archived byte-identically as G008-postgresql-rls-plan-r2.md (not included) (commit
    `2fc98c19171d40d2fd96afb10bbfda4afbb3bb01`, SHA-256
    `cfdd024222fe781512dee34683c1ab9da2a8711ad66bfa93b0593bceb133fec5`).
  - **Case fixed.** The EMPTY-only expire policy is ORed with the own-binding DELETE policy. Running the r2 cleanup
    autocommit, before any guard, would therefore let a reused connection carrying a session-level
    `app.login_binding = b` delete `b`'s unexpired pending rows.
  - **F3 is now one guarded transaction:** `BEGIN` → `init(EMPTY)` (the existing guard: any prior value means close
    and fail closed, before any mutation) → exact E5 cleanup → `init(LOGIN b)` (the single pinned EMPTY → LOGIN
    transition) → `INSERT` → `COMMIT`.
  - **Aligned:**
    - §2 item 3;
    - the §4.1 shape table, cleanup-shape rule and pinned re-initializations;
    - the §4.2 combined DELETE table and production-safety note;
    - E5, F3, §4.6, D4, U-RLS-BOUNDARY and I4;
    - RLS-AUTH E5 (c), corrected to show the raw own-binding effect, and the new production-path control (f).
  - **Cost:** login start goes from +3 to +4 round trips, and the expired-row locks are held until `COMMIT`. Not a
    material change to the Q7 assumption; recorded without escalation. No other cost changes.
  - **Bindings:** planning inputs rebound to `3068d98` (PROGRESS `0fe7afa3…e143`; PERFORMANCE.md unchanged).
  - **Settled items untouched:** the auth_reader ACL, D1/D2, E1/E3 independence, Q7, earlier decisions and the r0/r1
    archives. B7, checkpoint 4 and the 3–5 forecast are unchanged.
