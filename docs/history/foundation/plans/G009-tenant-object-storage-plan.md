> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G009 — Tenant-aware object storage boundary: Implementation Plan

Revision: r1, 2026-10-14 UTC. Status: **review-ready for Codex Plan re-review (provisional)**. This is not approval,
not an execution binding and not implementation. Implementation is held until G008 has final ACCEPT, the integrated
G008 state is reconciled, Codex has accepted this Plan and a fresh execution binding exists (§0).

r1 is the bounded amendment for the r0 Codex review (not included) (CORRECTION LOOP).
r0 is archived byte-identically as G009-tenant-object-storage-plan-r0.md (not included)
(SHA-256 `e2843c781042a18fb4d8623620d3405e0982ac7020c41f2a9c8ba5dd26afc6f7`). B5 and the fourteen ACs are unchanged.

**r1 correction map:**

| Review item | Change | Where |
|---|---|---|
| 1 | Interruptible, owned upload source; final-byte holdback and zero-length probe; honest completion classification and compensation; integrity claims narrowed | §4.1, §4.3, §4.4, §4.6, §5.1, §5.2, §11 |
| 2 | Production HTTPS for the data endpoint as well as signing; mixed HTTP-data/HTTPS-signing refused; loader plus constructor enforcement | §4.6, §4.7, §5.1, §5.6 |
| 3 | Read order: context, then live session, then RLS scope discovery, then bounded exact-Group facts, then check and decision, all before any provider call; no all-binding fallback | §4.4, §5.2, §10 |
| 4 | Automated fetch of the unchanged signed URL through the real published host port | §4.5, §5.2, §5.5 |
| 5 | Finite decision and cleanup budgets; scoped `tx.go` helper edit preserving G008 pool semantics; stall test | §2, §4.3, §4.4, §5.1, §5.2, §11 |
| 6 | FORCE-owner foreign-scope denial | §5.2 |

The Q1–Q8 rulings are recorded in §14.

- **Goal / Spec:** [Goal r1](../goals/G009-tenant-object-storage.md), [Spec r1](../specs/G009-tenant-object-storage-spec.md).
  Inputs: write-policy r1 (not included) and planning authorization (not included).
  The r0 Goal/Spec/review are history only.
- **Planner:**
  - Task G009, role plan, native worktree `.claude/worktrees/agent-<id>` under the G009 naming exception.
  - Branch `g009-plan`. Effort xhigh as instructed; the runtime setting is unverifiable from inside the worker (unknown).
- **Baseline (Git-first):** `origin/staging` `bbb2173e13f776024099786b7935e41ea13e14b7`.
  - Its application, infrastructure and foundation-doc bytes equal G008 merge `81b186c`; correction candidate `d0e8d40`.
  - The four later commits change only management and skills files.
  - All inputs are bound by this commit and path. Git-first needs no per-file planning inventory. No snapshot, wrapper
    or planning probe was used, and no test, build, migration or Docker command was run.
  - **Pre-push reconciliation:** staging advanced to `0a893194135d5d473983fbc0c692315ccdca6b21`. That commit adds only
    one PROGRESS entry: the owner's unattended G008/G009 continuation, with G008 still awaiting pipeline `2945531965`.
    - The delta does not affect any Plan input or validation input, and it adds no authority beyond the existing
      delivery loop.
    - The Plan commit is rebased onto it (rebase only).
  - **r1 base:** staging `c944ef7797ab1b50e8801158598cbbc8bc090378`, which adds the r0 review, the PG-003 r0
    assessment in `docs/PERFORMANCE.md` and a PROGRESS entry. No application, harness or G008 source changed. No
    test, build or probe ran for r1.

## 0. Gates and dependency reconciliation

| Gate | State at planning | Effect |
|---|---|---|
| G008 final ACCEPT | Pending: staging pipeline `2945531965` was running at release; correction/source review ACCEPT on `d0e8d40` | Implementation gate. The Plan is drafted against the integrated code. |
| Actual G008 reconciliation | Done for `81b186c` (below) | Repeat as an exact-delta check at execution binding. |
| Codex Plan ACCEPT and execution binding | Pending | Implementation starts only after both. |
| G021 product barrier | Closed | Unchanged. No product route, UI or workflow. |

**Integrated G008, as relied on here:**
- **00006:** nine forced RLS tables, 26 command-specific policies, 4 SECURITY INVOKER parsers and 2 policy readers.
  The readers are `app.rls_org_role()` and `app.rls_group_role(uuid)`, owned by NOLOGIN NOBYPASSRLS
  `<db>_policy_reader`. There are 3 auth readers.
- **Runtime:** the runtime role holds no tenant INSERT or DELETE and is a member of nothing. The owner is bound by FORCE.
- **Context:** the TENANT-ORG and TENANT-GROUP shapes are set with transaction-local `set_config` and hold canonical
  lowercase nonzero UUIDs only.
- **Transactions:** `access.inContextTx` is the only production transaction path, with the accepted pool-discard
  deviation (pgx v5.11.0).
- **Commit outcomes:** unknown-COMMIT classification, as in [AUTHORIZATION.md](../../../foundation/AUTHORIZATION.md).
- **PG-002:** an accepted, open risk, measured by same-transaction `pg_stat_get_xact_function_calls` deltas.

**If G008 needs a correction:** re-reconcile before binding the affected item.

| G008 change | G009 items it affects |
|---|---|
| `00006` catalog, reader names or semantics | §3.2 preconditions and postconditions, §3.3 policies, `MIGRATION_00006_SHA` |
| `tx.go`, `inContextTx` or the discard lifecycle | §4.4 transaction use; the r1 cleanup-budget helper edit (§4.4) |
| identity `LockLiveSession` / `CheckLiveSession` | the storage guard |
| `run.sh` RLS pins | §5.6 |

A change outside these paths only needs an exact-delta note.

**Service guidance:** core has no service-specific CLAUDE.md or skill at `bbb2173`. The applicable guidance is the
root governance plus `skills/saas-foundation/references/storage-audit.md` and `database-rls.md`. `docs/SECURITY.md` is
an empty file at baseline (observation only; this Plan does not change it).

## 1. Implementation approach

A route-less, core-internal storage increment on the accepted G007/G008 boundary:

1. **Owned provider adapter** `core/internal/objectstore`.
   - A provider-neutral interface: Put, Get, Delete, SignedGetURL, Metadata (Head).
   - An S3-compatible implementation on the already-pinned aws-sdk-go-v2.
   - It accepts only server-derived keys. Importers are pinned to `access` only.
2. **Metadata relation** `app.stored_objects` (migration `00007_stored_objects.sql`, additive).
   - UUID identity, required Organization, optional same-Organization Group (composite FK), immutable creator UUID.
   - A database-generated storage key bound to the UUID.
   - A four-state lifecycle.
   - FORCE RLS with three command policies on the 00006 pattern, a SECURITY INVOKER transition trigger, column-exact
     runtime grants and no new role or definer.
3. **Five guarded operations on `access.Service`:** `PutObject`, `GetObject`, `ObjectMetadata`, `SignedGetURL`,
   `DeleteObject`.
   - Each runs the existing binding check, then session liveness, then RLS scope discovery for an object selector, then
     bounded exact-scope facts, then the context check, then a separate closed storage-action decision. All of this
     happens before any provider call or disclosure.
   - Provider I/O happens only between short database transactions. Each transaction has finite decision and cleanup
     budgets.
   - Upload sources are owned and interruptible.
   - Uploads publish only after re-authorization, and compensation runs only after an authoritative state is established.
4. **Logging:** one catalogued `storage.operation` record per call, with UUID attribution (R9) and no URL, secret,
   payload or provider text.
5. **Harness reconciliation:** the minimum covering harness changes.
   - Version-7 pins.
   - New migrations-suite checks B-OBJ-CATALOG, -MIGRATION, -IT, -ORDERING and -FAULTS.
   - The event catalog.
   - One profile-path row.
6. **Docs:** `docs/foundation/OBJECT-STORAGE.md` and updates to the existing foundation docs.

Nothing is wired into a binary. `cmd/server/main.go`, readiness, identity production code, web, intelligence,
contracts, compose, `.env.example`, `infra/minio/init.sh`, provisioning and 00001–00006 stay byte-identical.

## 2. Files and modules

**Create:**

| Path | Content |
|---|---|
| `core/migrations/00007_stored_objects.sql` | §3 |
| `core/internal/objectstore/objectstore.go` | `Store` interface, `Key` (validated `objects/<uuid>/<uuid>`), `Binding` metadata, `Info`, closed errors (`ErrNotFound`, `ErrDenied`, `ErrUnavailable`, `ErrOutcomeUnknown`, `ErrInvalid`) |
| `core/internal/objectstore/s3.go` | S3-compatible adapter (§4.6): one data client plus one presign client, bounded transport, no retries. The constructor `New(cfg, appEnv)` re-enforces the §4.7 endpoint transport rules. |
| `core/internal/objectstore/objectstore_test.go` | U-OBJ-KEY and U-OBJ-ADAPTER (httptest classification and request shape only; never provider evidence) |
| `core/internal/tenancy/objects.go` | `ObjectID` and its parser, `StoredObject`, `ObjectState`, `ValidateContentType`, statements, transaction-scoped primitives (`InsertObjectTx`, `ReadObjectTx`, `LockObjectTx`, `FinalizeObjectTx`, `MarkObjectDeletingTx`, `MarkObjectDeletedTx`); every statement names its Organization |
| `core/internal/tenancy/objects_test.go` | unit validators |
| `core/internal/tenancy/access/objects.go` | the five operations, the storage guard, the Put/Delete state machines, reconcile and compensation, inert test seams |
| `core/internal/tenancy/access/objects_policy.go` | `decideObject`, the closed storage-action mapping (§4.2), input and TTL validation |
| `core/internal/tenancy/access/objects_io.go` | owned upload source (close-once, close on cancel or deadline, in-flight Read tracking for reaping), exact reader with final-byte holdback, bounded and verifying download reader, the `SignedURL` redacting type |
| `core/internal/tenancy/access/objects_test.go` | U-OBJ-POLICY, U-OBJ-INPUT, U-OBJ-READER, U-OBJ-EVENT, U-OBJ-SEAM |
| `core/internal/tenancy/access/objects_expected_test.go` | **independent oracle** (delegate-owned, untagged pure data, §7) |
| `core/internal/tenancy/access/objects_integration_test.go` | `TestObjects` (B-OBJ-IT) |
| `core/internal/tenancy/access/objects_ordering_integration_test.go` | `TestObjectsOrdering` (B-OBJ-ORDERING) |
| `core/internal/tenancy/access/objects_faults_integration_test.go` | `TestObjectsFaults` (B-OBJ-FAULTS) |
| `core/internal/tenancy/access/objects_footprint_integration_test.go` | `TestObjectsFootprint` (E-OBJ-1..4 in B-OBJ-IT) |
| `core/internal/migrate/objects_integration_test.go` | `TestObjectsMigration` (B-OBJ-MIGRATION) |
| `core/internal/config/objects_config_test.go` | U-OBJ-CONFIG (a new file, because the rendered `config_test.go` is not a registered customization) |
| `docs/foundation/OBJECT-STORAGE.md` | §4.9 |
| `management/v0/results/G009-scope-allowlist.txt` | B-SCOPE allow-list, created and committed at execution binding from §2 |
| `management/v0/results/G009-tenant-object-storage-result.md` | the Result |

**Modify:**

| Path | Change |
|---|---|
| `core/internal/config/config.go` | optional object settings with defaults and bounds (§4.7). Already a registered customization. |
| `core/internal/tenancy/access/access.go` | `NewServiceWithObjects` and unexported object fields. `NewService` is unchanged; without objects every storage op is `unavailable` (`objects_unconfigured`). |
| `core/internal/tenancy/access/tx.go` (**scoped helper edit, r1**) | `inTx` and `inContextTx` gain `cleanup time.Duration`. `0` keeps today's deferred `Rollback(context.WithoutCancel(ctx))` byte-for-byte in behavior. A positive value bounds the deferred rollback with `context.WithTimeout(context.WithoutCancel(ctx), cleanup)`. The single `BeginTx` site, the `inContextTx` → `InitContext`-first order, discard while held and the commit classification are unchanged. |
| `core/internal/tenancy/access/{facade,context}.go`, `commit_test.go` | existing call sites pass `0` (`cleanupUnbounded`), so existing operations are unchanged; the U-RLS-BOUNDARY pins still hold |
| `core/internal/tenancy/access/errors.go` | storage failure mapping |
| `core/internal/tenancy/access/log.go` | `storage.operation` record. This is the only file that may read correlation. |
| `core/internal/tenancy/access/boundary_test.go` | pins for files, imports and exported surface. Production-importer rules unchanged: nothing imports `access`, only `access` imports `tenancy`. Adds an objectstore importer rule (only `access`) and an aws-sdk importer rule (`objectstore`, `health`, `cmd/server`). |
| `core/internal/tenancy/errors.go`, `boundary_test.go` | input codes; statement-inventory pins |
| `core/migrations/migrations_test.go`, `core/internal/migrate/runner_test.go` | set 00001–00007; 00006 bytes pinned |
| version-6 pins in G005–G008 Go tests (§5.6) | version-7 reconciliation only. **`rls_expected_test.go` (G008 oracle) unchanged.** |
| `infra/validation/run.sh` | version-7 pins, `MIGRATION_00006_SHA`, B-OBJ-* functions including the published-host signed-URL fetch, MinIO start in the migrations project (§5.5); the B-IDENTITY-CONFIG production probe uses an https data endpoint (§5.6) |
| `infra/validation/required-checks.tsv` | +5 IDs (suite `migrations`) |
| `infra/validation/tools/internal/logrecord/events.json` | `storage.operation`; five new `core.exit` config keys |
| `infra/ci/profile-paths.tsv` | one row, `foundation dir core/internal/objectstore/`. Otherwise an objectstore-only change would select integration, which never runs the migrations suite that holds the storage checks. |
| `docs/foundation/{DATABASE-ISOLATION,AUTHORIZATION,MIGRATIONS,VENDORS,FOLLOW-UP}.md` | version 7, storage operations, the R2 row notes, G010/G011/G013 handoffs |

**Not modified:**
- `core/cmd/**`, `core/internal/health/**`, identity production code, `core/go.mod`/`go.sum` (expected; see risk E9);
- 00001–00006, `infra/postgres/**`, `infra/minio/**`, `infra/compose/**`, `.env.example`;
- `web/`, `intelligence/`, `contracts/`, `.gitlab-ci.yml`, other `infra/ci/*`, `skills/`;
- `docs/PERFORMANCE.md` (Codex-owned; evidence goes to the Result).

## 3. Database: migration `00007_stored_objects.sql`

### 3.1 Relation (DDL)

```sql
CREATE TABLE app.stored_objects (
  id              uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  organization_id uuid NOT NULL,
  group_id        uuid,                                   -- NULL: Organization-scoped object
  creator_user_id uuid NOT NULL,                          -- immutable uploader attribution (R2/R3), never a role
  storage_key     text GENERATED ALWAYS AS ('objects/' || organization_id::text || '/' || id::text) STORED,
  state           text NOT NULL DEFAULT 'pending' CHECK (state IN ('pending', 'ready', 'deleting', 'deleted')),
  byte_size       bigint NOT NULL CHECK (byte_size BETWEEN 0 AND 1073741824),
  content_type    text NOT NULL CHECK (content_type ~ '^[a-z0-9][a-z0-9!#$&^_.+-]{0,62}/[a-z0-9][a-z0-9!#$&^_.+-]{0,62}$'),
  sha256          bytea CHECK (sha256 IS NULL OR octet_length(sha256) = 32),
  created_at      timestamptz NOT NULL DEFAULT now(),
  ready_at        timestamptz,
  deleting_at     timestamptz,
  deleted_at      timestamptz,
  deleted_by      uuid,
  CONSTRAINT stored_objects_organization_fkey FOREIGN KEY (organization_id) REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT stored_objects_group_fkey FOREIGN KEY (organization_id, group_id) REFERENCES app.groups (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT stored_objects_creator_fkey FOREIGN KEY (creator_user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT stored_objects_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT stored_objects_lifecycle_check CHECK (
       (state = 'pending'  AND sha256 IS NULL AND ready_at IS NULL AND deleting_at IS NULL AND deleted_at IS NULL AND deleted_by IS NULL)
    OR (state = 'ready'    AND sha256 IS NOT NULL AND ready_at IS NOT NULL AND deleting_at IS NULL AND deleted_at IS NULL AND deleted_by IS NULL)
    OR (state = 'deleting' AND deleting_at IS NOT NULL AND deleted_by IS NOT NULL AND deleted_at IS NULL)
    OR (state = 'deleted'  AND deleting_at IS NOT NULL AND deleted_by IS NOT NULL AND deleted_at IS NOT NULL))
);
CREATE INDEX stored_objects_organization_group_idx ON app.stored_objects (organization_id, group_id);
```

**Key decisions:**
- **Server-generated identity.**
  - `id` is a v4 UUID that core generates from `crypto/rand` before the insert, so a lost-reply reconciliation knows it.
  - The key is a database-generated column. Runtime, the owner and callers cannot write it.
  - Because `id` is unique and the key derives from `(organization_id, id)`, two rows can never share a provider key.
  - A forged row cannot redirect an operation to another object's bytes. Core also derives the key in Go and asserts
    equality.
- **Provider binding.** The uploaded object carries the user metadata `object-id`, `organization-id` and `group-id`.
  Get and Put verification compare them, together with the length, to the row.
- **No filename or display name.** This keeps private metadata minimal; product Goals own naming (question Q7).
- **Tombstones.** Rows are never hard-deleted by runtime (no DELETE grant). `deleted` rows keep attribution for G011
  and repair. Retention belongs to G013.
- **Relations added:** table, `stored_objects_pkey` and `stored_objects_organization_group_idx`. App relations go from
  35 to 38.

### 3.2 Up: preconditions, objects, grants, policies, postconditions (one transaction)

**Preconditions** (fail closed, before any change):
- the current user owns schema `app`;
- `<db>_runtime`, `_policy_reader` and `_auth_reader` exist with the 00006 safe attributes and memberships (the 00006
  checks, reused);
- `to_regclass('app.stored_objects') IS NULL`;
- three inventories equal the **exact version-6 strings** of 00006 postconditions 1, 2 and 4, copied verbatim: the RLS
  table set, the policy inventory and the function allowlist.

**Trigger** (owner-owned SECURITY INVOKER, `search_path = pg_catalog, pg_temp`, PUBLIC EXECUTE revoked):

```sql
CREATE FUNCTION app.stored_objects_guard() RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE actor uuid := app.ctx_user();      -- parser: malformed/zero/illegal context raises
BEGIN
  IF (NEW.id, NEW.organization_id, NEW.group_id, NEW.creator_user_id, NEW.byte_size, NEW.content_type, NEW.created_at)
     IS DISTINCT FROM (OLD.id, OLD.organization_id, OLD.group_id, OLD.creator_user_id, OLD.byte_size, OLD.content_type, OLD.created_at)
  THEN RAISE EXCEPTION 'stored object identity is immutable' USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'stored_objects_immutable'; END IF;
  IF actor IS NULL THEN RAISE EXCEPTION 'stored object update needs an actor' USING ERRCODE = 'insufficient_privilege'; END IF;
  NEW.ready_at := OLD.ready_at; NEW.deleting_at := OLD.deleting_at; NEW.deleted_at := OLD.deleted_at; NEW.deleted_by := OLD.deleted_by;
  IF OLD.state = 'pending' AND NEW.state = 'ready' AND OLD.creator_user_id = actor AND NEW.sha256 IS NOT NULL THEN
    NEW.ready_at := now();                                              -- only the creator finalizes (Put T2)
  ELSIF OLD.state IN ('pending', 'ready') AND NEW.state = 'deleting' AND NEW.sha256 IS NOT DISTINCT FROM OLD.sha256 THEN
    NEW.deleting_at := now(); NEW.deleted_by := actor;                  -- logical deletion
  ELSIF OLD.state = 'deleting' AND NEW.state = 'deleted' AND NEW.sha256 IS NOT DISTINCT FROM OLD.sha256 THEN
    NEW.deleted_at := now();                                            -- tombstone after confirmed physical removal
  ELSE RAISE EXCEPTION 'illegal stored object transition' USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'stored_objects_transition';
  END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER stored_objects_guard BEFORE UPDATE ON app.stored_objects FOR EACH ROW EXECUTE FUNCTION app.stored_objects_guard();
```

The trigger refuses same-state updates, any update of a `deleted` row and every unlisted transition.

**Grants** (runtime only; helpers and PUBLIC get nothing):
- `GRANT SELECT ON app.stored_objects TO <db>_runtime`;
- `GRANT INSERT (id, organization_id, group_id, creator_user_id, byte_size, content_type) ON app.stored_objects TO <db>_runtime`.
  `state`, `created_at` and the key come from the database;
- `GRANT UPDATE (state, sha256) ON app.stored_objects TO <db>_runtime`. The trigger sets timestamps and `deleted_by`;
- no DELETE, TRUNCATE, REFERENCES or TRIGGER privilege.

**RLS:** `ENABLE` and `FORCE ROW LEVEL SECURITY`. Three permissive, command-specific policies `TO <db>_runtime,
<db>_owner`, built with `format()` as in 00006. U, O and G are the parsers, R is `(SELECT app.rls_org_role())`,
`MGR := R IN ('org_owner','org_admin')`, and `GR(g) := app.rls_group_role(g)`. `SCOPE` is the read scope:

```
SCOPE := organization_id = O AND R IS NOT NULL AND (group_id IS NULL OR ((G IS NULL OR group_id = G) AND (MGR OR GR(group_id) IS NOT NULL)))
stored_objects_select  FOR SELECT  USING (SCOPE)                                                           -- organization.read / group.read
stored_objects_insert  FOR INSERT  WITH CHECK (organization_id = O AND R IS NOT NULL AND creator_user_id = U AND state = 'pending'
                         AND ((group_id IS NULL AND MGR)
                           OR (group_id IS NOT NULL AND (G IS NULL OR group_id = G) AND (MGR OR GR(group_id) IS NOT NULL))))   -- Put matrix
stored_objects_update  FOR UPDATE  USING (W) WITH CHECK (W), where
  W := organization_id = O AND R IS NOT NULL
       AND ((group_id IS NULL AND MGR)
         OR (group_id IS NOT NULL AND (G IS NULL OR group_id = G)
             AND (MGR OR GR(group_id) IN ('group_owner','group_admin') OR (creator_user_id = U AND GR(group_id) IS NOT NULL))))
```

`W` together with the trigger gives the database form of R3:
- Delete-any: Organization owner/admin, or the Group's owner/admin.
- Own delete: the creator with a current role in that exact Group.
- Finalize: the trigger allows only the creator. RLS additionally requires the creator to still hold upload scope (any
  current Group role, or Organization owner/admin for Organization objects).
- Organization objects: owner/admin only, including objects a former admin uploaded.
- `W` has no state predicate, so a `deleted` row stays lockable for an idempotent re-delete. The trigger forbids
  changing it.

**Postconditions** (exact strings, in the 00006 style with symbolic labels):

| Inventory | Version 7 |
|---|---|
| (1) RLS tables | the v6 set plus `stored_objects+force` (10) |
| (2) policy inventory | v6 plus `stored_objects_insert:stored_objects:a:true:ow+rt`, `stored_objects_select:…:r:true:ow+rt` and `stored_objects_update:…:w:true:ow+rt` (29) |
| (3) ACLs | v6 plus `stored_objects:rt:SELECT`, the six `stored_objects.<col>:rt:INSERT` and `stored_objects.sha256:rt:UPDATE`, `stored_objects.state:rt:UPDATE`. No helper or PUBLIC entry. |
| (4) function allowlist | v6 plus `stored_objects_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-` |
| (5) schema privileges | unchanged: `ar=USAGE,pr=USAGE,rn=USAGE,rt=USAGE` |

No new role, membership, SECURITY DEFINER, LEAKPROOF function or helper grant. The access boundary rule "SECURITY
DEFINER appears only in 00006" stays true.

### 3.3 Down (development and test only; the runner refuses it in production)

One transaction:
1. `ALTER TABLE app.stored_objects NO FORCE ROW LEVEL SECURITY`, so the owner's count is not filtered.
2. **Refuse** if any row exists. Dropping metadata would orphan provider bytes; the refusal rolls back the NO FORCE.
3. `DROP TABLE app.stored_objects`, which removes the policies, trigger, index and grants.
4. `DROP FUNCTION app.stored_objects_guard()`.
5. Postconditions: the exact version-6 inventories.

The helper roles and runtime are untouched.

### 3.4 Migration risk handling

- **Locks.** Additive and catalog-only. CREATE TABLE with FKs takes brief `SHARE ROW EXCLUSIVE` locks on
  `organizations`, `groups` and `users`.
  - No rewrite, no scan and no data change.
  - Run in a quiescent window, like 00006.
- **Compatibility.**
  - Deploy order: migrate to 7, then deploy the code.
  - G008 code at version 7 is unaffected.
  - G009 code at version 6 fails closed: storage calls are `unavailable`, since the relation is missing.
- **History.** 00001–00006 stay byte-identical (digest pins, §5.6).
- **Irreversibility.** The key layout `objects/<organization>/<id>` becomes durable provider layout once real data exists
  (escalation E3).

### 3.5 Inventory of added database objects (Spec R2)

| Kind | Added |
|---|---|
| Relations | `app.stored_objects`, `stored_objects_pkey`, `stored_objects_organization_group_idx` |
| Functions | `app.stored_objects_guard()` (trigger, invoker) |
| Triggers | `stored_objects_guard` BEFORE UPDATE |
| Policies | `stored_objects_select`, `stored_objects_insert`, `stored_objects_update` |
| Constraints | pkey/id-nonzero, state, size, content type, sha256 length, lifecycle, 2 Organization/Group FKs, 2 user FKs |
| Roles / memberships | none |
| Runtime privileges | SELECT; INSERT on 6 columns; UPDATE(state, sha256) |
| Lifecycle privileges | pending→ready: creator, with RLS upload scope. pending/ready→deleting: delete scope. deleting→deleted: delete scope. Nothing else. |

## 4. Contracts and interfaces

**Public API/MCP: none.** There is no route, handler, contract file, BFF or MCP tool. The internal Go API on
`access.Service` is new and its exported surface is pinned by U-ACCESS-BOUNDARY. Nothing imports `access` in production.
B-MCP-SOURCE and the existing TestProductionImporters rules stay unchanged; only the objectstore and aws-sdk importer
rules are added.

### 4.1 Guarded operations (internal Go API)

```go
func NewServiceWithObjects(db tenancy.DB, id *identity.Service, logger *slog.Logger, store objectstore.Store, cfg ObjectConfig) *Service
func (s *Service) PutObject(ctx context.Context, tc *TenantContext, req PutObjectRequest) (ObjectInfo, error)
func (s *Service) ObjectMetadata(ctx context.Context, tc *TenantContext, id tenancy.ObjectID) (ObjectInfo, error)
func (s *Service) GetObject(ctx context.Context, tc *TenantContext, id tenancy.ObjectID) (ObjectInfo, io.ReadCloser, error)
func (s *Service) SignedGetURL(ctx context.Context, tc *TenantContext, req SignedGetURLRequest) (SignedURL, error)
func (s *Service) DeleteObject(ctx context.Context, tc *TenantContext, id tenancy.ObjectID) (DeleteResult, error)

type PutObjectRequest struct { Group tenancy.GroupID; ContentType string; Size int64; Body io.ReadCloser } // zero Group: Organization object
type SignedGetURLRequest struct { Object tenancy.ObjectID; TTLSeconds int64; DefaultTTL bool }       // exactly one form
type ObjectInfo struct { ID tenancy.ObjectID; Organization tenancy.OrganizationID; Group tenancy.GroupID
                         ContentType string; Size int64; SHA256 [32]byte; CreatedAt, ReadyAt time.Time }
type SignedURL struct { /* url unexported */ ExpiresAt time.Time; TTL time.Duration }  // URL() string; String()/LogValue() redact
type DeleteResult struct { Outcome DeleteOutcome }                                    // DeleteOutcomeDeleted | DeleteOutcomeDeletionPending
```

- **Inputs.** Only the bound `TenantContext` and canonical UUID selectors. No role, allow result, creator, bucket,
  endpoint, key or path parameter exists (R1).
- **Upload source contract (r1).** `PutObject` owns `Body` from entry and closes it exactly once on every path,
  including validation failures, denials, cancellation and success.
  - `Close` must be safe to call concurrently with `Read` and must make a blocked `Read` return promptly. Conforming
    sources include `io.Pipe` readers, `net.Conn`-backed bodies and files. This is part of the documented API
    contract for any future adapter.
  - `Body` is never read after `PutObject` returns.
- **Errors.** The existing closed `access.Error` codes. `Field` codes for `invalid_input`: `invalid_id`,
  `invalid_content_type`, `object_size_required` (Size < 0, meaning an unknown length), `object_too_large`,
  `object_size_mismatch`, `missing_body` and `invalid_ttl`.
- **Missing or inaccessible.** Missing, foreign, inaccessible, outside-context and non-`ready` objects (for reads) are
  all `not_found`. A readable object whose action is denied is `forbidden`, decided by G007's same-facts read rule.
  Provider details never appear in errors.

### 4.2 Storage-action mapping (explicit review: escalation E1)

Reads reuse G007 capabilities unchanged. Writes use a **separate closed mapping** in `objects_policy.go`. The eleven
G007 capabilities, `decide`, `capabilityKind` and U-POL-MATRIX stay unchanged, and `organization.manage_*` is not
borrowed.

| Action | Organization object | Group object (owning Group g) |
|---|---|---|
| read (Metadata, Get, Sign) | `organization.read` (decide) | `group.read` on g (decide) |
| put | org_owner, org_admin | org_owner, org_admin, or any current role in g |
| delete | org_owner, org_admin | org_owner, org_admin, or group_owner/group_admin of g; **group_member of g only when stored creator = subject** |

The gates run in G007 order:
1. the current Organization binding;
2. the owning Group belongs to the active Organization (a foreign or unknown Group is `not_found`);
3. the selected Group restricts Group targets (`target_outside_context` maps to `not_found`);
4. the mapping.

Organization objects stay accessible in selected-Group context, as the context contract allows Organization
authority. Creator equality is evaluated only after the role gate and never grants scope. RLS enforces the same table
independently (§3.2). The oracle checks both layers (§7).

### 4.3 Bounds (R4, R5)

| Setting | Default | Allowed range | Error |
|---|---|---|---|
| Max object bytes | 26214400 (25 MiB) | 1..1073741824 (the hard cap, also enforced in the database) | Size < 0 → `object_size_required`; Size > max → `object_too_large`, both before any database or provider work. A stream shorter or longer than Size → `object_size_mismatch` (§4.4). |
| Size 0 | allowed (an empty object) | — | — |
| Content type | `application/octet-stream` when empty | input lowercased; must match the database pattern, ≤ 127 bytes, no parameters | `invalid_content_type` |
| Signed-URL TTL | 60 s (`DefaultTTL: true`) | `TTLSeconds` 1..config max, where config max ≤ 300 (the hard cap) | 0, negative, over max, both forms or neither → `invalid_ttl`. Checked before the int-to-duration conversion, so it cannot overflow. |
| Transfer deadline (Put body, Get stream) | 120 s | 1..900 s | — |
| Database **decision budget** (each storage transaction: BEGIN through COMMIT) | 5 s fixed | — | — |
| Database **cleanup budget** (each storage transaction's deferred rollback, r1) | 5 s fixed | — | — |
| Provider control calls (Head, Delete) | 10 s fixed | — | — |
| Source reap grace (after Close, wait for an in-flight `Read` to return) | 5 s fixed | — | a source that does not return becomes `unavailable` (`source_not_released`), residual X6 |

The HTTP transport is fixed: dial 5 s, TLS 5 s, response header 15 s, idle 90 s, 16 idle connections per host. The SDK
retryer is set to one attempt (no automatic retries): a retry is the caller's new, re-authorized operation.

**Resulting finite bounds** (limits, not latency promises):
- A storage transaction lasts at most 5 s decision plus 5 s cleanup. A poisoned connection adds G008's existing 5 s
  discard close.
- A PutObject call is bounded by the sum of its stages: T1, the transfer deadline plus the reap grace, Head, T2, R and C.
- A Get stream is bounded by the transfer deadline.

### 4.3a Transaction budgets and the scoped `tx.go` edit (r1, correction 5)

Today `inTx` defers `raw.Rollback(context.WithoutCancel(ctx))`, which strips any deadline, so a stalled server or
transport could hold the cleanup indefinitely.

Storage transactions run through one unexported wrapper:

```go
func (s *Service) storageTx(ctx context.Context, write bool, dbc dbContext, fn func(tx *trackedTx) *failure) *failure {
	ctx, cancel := context.WithTimeout(ctx, storageDecisionBudget) // 5 s: BEGIN, InitContext, statements, COMMIT
	defer cancel()
	return s.inContextTx(ctx, write, dbc, storageCleanupBudget, fn) // 5 s bound on the deferred rollback
}
```

`inTx(ctx, write, cleanup, fn)` defers:
- `cleanup == 0` (every existing operation): `raw.Rollback(context.WithoutCancel(ctx))`, unchanged;
- `cleanup > 0`: `rctx, c := context.WithTimeout(context.WithoutCancel(ctx), cleanup)`, then `raw.Rollback(rctx)`,
  then `c()`.

**Disposal under the accepted G008 lifecycle (pgx v5.11.0):**
- An expired cleanup context interrupts the rollback. pgx treats the failed rollback as a real rollback error and
  kills the connection while it is still held. pgxpool destroys the closed connection on release.
- This adds no post-release Close, so it stays within the accepted pool-discard deviation.
- A decision-budget expiry mid-statement or mid-COMMIT follows the existing classification (`canceled`, or
  `commit_outcome_unknown` for COMMIT). The connection is then not reused.
- After a successful COMMIT the deferred rollback is a no-op without I/O.

Nothing else changes: G007/G008 operations, `discard`, commit classification, the single `BeginTx` site and the
`inContextTx` → `InitContext`-first pin. The covering tests are U-OBJ-SEAM and FX-STALL (§5).

### 4.4 Operation sequences, linearization and outcomes

Every transaction is `storageTx`: `access.inContextTx` with the context statement first, the 5 s decision budget and
the 5 s cleanup budget (§4.3a). Writes are READ COMMITTED and use G007's ordered `FOR SHARE` facts:
- user → session → Organization → Organization binding → Group bindings → Group;
- the owning Group is read `FOR SHARE` (no `TargetGroup` exclusive lock);
- then the object row `FOR UPDATE`.

Reads are a REPEATABLE READ READ ONLY snapshot. **No provider call happens inside a transaction**, and no connection
is held during a transfer.

**PutObject:**

| Step | Action | On failure |
|---|---|---|
| P0 | take ownership of `Body` (close-once wrapper); binding check; validate selector, size, type and body presence | the error, with `Body` closed; no database or provider effect |
| P1 | for Size 0 only: a one-byte probe of `Body` under the transfer deadline (cancellation or the deadline closes `Body`) | data → `object_size_mismatch`; a blocked probe → `unavailable` (`canceled`). Neither makes a database or provider call. EOF → continue; the upload later sends an empty body. |
| T1 | facts (lock), then contextCheck, then decide(put), then INSERT pending (creator = subject, also enforced by RLS), then COMMIT | Deny: the code, nothing written. Abort or statement error: `unavailable`, nothing written. **Unknown COMMIT:** `unavailable`, no provider call; a committed pending row has no bytes, is inaccessible and is attributable (residual X2). |
| P2 | provider Put of the exact key: `Content-Length` = Size, the exact reader with holdback, SHA-256 tee, binding metadata, `Content-Disposition: attachment`. `context.AfterFunc(transferCtx, Body.Close)` unblocks a blocked `Read` on cancellation or deadline. After the SDK call returns (on every path) `Body` is closed and PutObject waits up to the reap grace for any in-flight `Read` to return, so there is no leaked transfer worker. | Classified by the completion table below; every non-success → C |
| V | provider Head: length = Size and binding equal | → C; `unavailable` or `internal` |
| T2 | facts (lock), then contextCheck, then **decide(put) again**, then lock row (must be pending with creator = subject), then UPDATE ready, sha256, then COMMIT | Deny, or the row is no longer pending: definitive rollback → C → the code (`not_found` when it is gone). Definitive abort → C → `unavailable`. **Unknown COMMIT → R.** |
| R | a reconcile transaction that locks the row `FOR UPDATE`, which waits out an in-flight T2 | `ready` → success, bytes kept. `pending`, `deleting` or `deleted` → C, then `unavailable`. Lock unavailable, denied or timed out → stop with `unavailable` and **no compensation** (the row may be ready). |
| C | compensation: provider Delete of the exact key, then Head confirms it absent. Metadata is not changed. | Failure: bytes may remain under a non-ready row. They stay inaccessible and attributable; the repair is an authorized Delete (residual X3). |

**Invariant I-FIN:** only the PutObject call that inserted a pending row can finalize it, inside that call. There is
no resume API, and the trigger allows finalization only by the creator. Compensation therefore runs only when the
state is authoritatively non-ready: a definitive T2 outcome, or R's locked read. It never runs after an unknown COMMIT
without R, so possibly committed bytes are never destroyed. A late in-flight PUT that lands after C leaves inaccessible
bytes under a non-ready row (residual X3).

**Exact reader with final-byte holdback (r1).** It releases bytes 1..Size−1 as they arrive. It releases the final
byte only after a lookahead `Read` returns EOF. Extra data or an early EOF makes `Read` return an error instead, so the
transport has delivered at most Size−1 body bytes against `Content-Length` = Size.

The reader records a verdict: `exact`, `short`, `long`, `source_error` or `not_released`. **This Plan does not claim
that a mismatch cannot create a provider object.** Completion is classified only from the verdict, the provider result
and V:

| Reader verdict | Provider Put result | Classification and action |
|---|---|---|
| `exact` | success, V passes | T2 (publish path) |
| `exact` | success, V fails (length or binding) | non-ready; C; `internal` (binding) or `unavailable` |
| `exact` | error or timeout (outcome unknown) | non-ready; C; `unavailable` (`provider_outcome_unknown`) |
| `short`, `long` | error (expected) **or success** (the provider completed regardless) | non-ready; C (also deletes any completed object); `invalid_input` (`object_size_mismatch`), with a `compensated` or `compensation_failed` reason |
| `source_error`, `not_released` | any | non-ready; C; `unavailable` (`canceled`, or `source_not_released`) |
| Size 0 (P1 EOF) | success or unknown | an empty body is sent; V must show length 0 and the binding; otherwise as the `exact` rows |

Metadata never becomes `ready` unless the verdict is `exact`, the provider reports success and V passes. Compensation
follows I-FIN.

**What the digest proves (r1, narrowed).**
- The stored `sha256` is the digest of the bytes core delivered to the transport.
- V proves only the stored length and the binding metadata. **Neither independently proves that the stored payload
  equals those bytes**: the payload is unsigned and no provider checksum is sent.
- The core `GetObject` path verifies the stored content against `sha256` as it streams. A mismatch is reported at EOF,
  after any delivered bytes.
- **Signed downloads are served by the provider directly and are not verified by core.**
- Production transit integrity relies on TLS (§4.7).

Unknown-length streams are impossible, since Size is required.

**GetObject, ObjectMetadata, SignedGetURL** (r1 read order, correction 3). One REPEATABLE READ READ ONLY snapshot,
all before any provider call or disclosure:
1. Binding check (no SQL).
2. Context initialization (the first statement).
3. `CheckLiveSession`.
4. **RLS scope discovery.**
   - `ReadObjectTx` selects the immutable `organization_id`, `group_id`, `creator_user_id` and the state for
     `organization_id = <active> AND id = <selector>` under RLS. A row comes back only if it is in the caller's read
     scope.
   - No row → scope unknown. Facts become the Organization-only query (`Groups: []`, plus the selected Group), for
     `contextCheck` only. The result is `context_denied` if the context fails, otherwise `not_found`.
5. **Bounded facts:**
   - a Group object reads `FactQuery{Groups: [owning Group] (+ selected Group)}`;
   - an Organization object reads the Organization-only query;
   - **never `Groups: nil`, the all-binding read.**
6. `contextCheck`, then decide(read) on the discovered scope; the row must be `ready`. The discovered row is not
   returned before the allow.
7. COMMIT, then the provider step:
   - **Get:** provider Get under the transfer deadline, verifying the binding metadata and Content-Length before
     returning. The returned reader is bounded to Size, verifies SHA-256 at EOF (an integrity error is reported at EOF),
     and Close cancels the transfer.
   - **Metadata:** none (no provider call).
   - **Sign:** an offline presign (§4.5).
8. Provider failures map as follows. A missing provider object for a `ready` row is `not_found`, with the error-level
   reason `object_bytes_missing`. Unavailable is `unavailable`. A binding mismatch is `internal`.

DeleteObject D1 and Put T2/R use the same discovery-then-bounded-facts order: D1's plain read is the discovery, and
Put already knows its scope from the request and T1. No storage path uses the all-binding fact read, so fact work is
constant in the subject's binding count.

**DeleteObject:**

| Step | Action | On failure |
|---|---|---|
| D0 | binding check; selector | the error |
| D1 | plain read of the row (its immutable scope and creator), then facts (lock) for the owning scope, then contextCheck, then decide(delete) on the immutable scope and creator, then lock the row `FOR UPDATE` (the state is read under the lock), then: pending/ready → UPDATE deleting; deleting/deleted → no update. Then COMMIT. The decision precedes the lock, so the RLS update scope never hides a readable row and turns `forbidden` into `not_found`. | Deny: `forbidden` or `not_found`; both systems unchanged. Unknown COMMIT: `unavailable` (the row may be deleting; a retry completes it). |
| D2 | provider Delete (idempotent: absent counts as success) | → result `deletion_pending` |
| D3 | provider Head must be NotFound | present or unknown → `deletion_pending` |
| D4 | facts (lock), then lock the row, then decide(delete), then UPDATE deleted (for a deleting row), then COMMIT. Skipped for an already-deleted row. | any failure, denial or unknown → `deletion_pending` |

The result is `deleted` only when the row is deleted **and** absence is confirmed. After D1 commits, the call returns
`(DeleteResult, nil)`. A repeat Delete by an authorized deleter is idempotent: it re-issues D2–D3 and returns `deleted`.
A missing or foreign object is `not_found`. **A failed provider delete is never reported as confirmed.**

**Linearization points:**

| Operation | Point |
|---|---|
| Put | completion of the T1 facts locks (insert), and the T2 facts locks (publish) |
| Delete | completion of the D1 facts locks (decision), then the row lock (state); logical deletion at D1 COMMIT |
| Read and Sign | the snapshot |

- **Revocation.** A revocation committed before a point denies every new step. A revocation after it may let an
  in-flight transfer, publish or signed issuance finish.
- **Logical deletion.** After D1 commits, no new read or sign succeeds.
- **Earlier URLs** stay usable until expiry (≤ 300 s) or physical removal. In-flight downloads may finish.
- **Not claimed:** immediate URL invalidation and transfer recall (R5/R8).

### 4.5 Signed GET and endpoint identity (R5)

A dedicated presign client uses the trusted `OBJECT_STORE_SIGNING_ENDPOINT` (default `OBJECT_STORE_ENDPOINT`), the
same credential and region, and path-style addressing.
- **Request.** Method GET only. The exact key comes from the row. `X-Amz-Expires` is the TTL. The signed response
  overrides are `response-content-disposition=attachment`, `response-content-type=<stored type>` and
  `response-cache-control=no-store`. The caller controls none of these.
- **No request input.** Presigning is offline. No inbound Host or forwarded header is read, and the URL is never
  rewritten.
- **Local host.** Set `OBJECT_STORE_SIGNING_ENDPOINT=http://127.0.0.1:9100` (BIND_ADDRESS plus `MINIO_API_PORT`;
  configurable). SigV4 binds the `Host` value, including its port, and Docker port publishing forwards the client's
  `Host` unchanged.
  - **r1 proof:** an automated fetch of the **unchanged** URL through the real published host port (OBJ-SIGN-HOST,
    §5.2 and §5.5), plus the internal altered-Host negatives.
- **In-network.** The default `http://minio:9000` is valid only for consumers on the data network.
- **Production.** Non-https signing **and** data endpoints are refused (§4.7).
- **Secrecy.** The URL and its query are a bearer secret: not logged, not stored, not placed in evidence. `SignedURL`
  redacts it in `String()` and `LogValue()`. The URL path reveals the Organization and object UUIDs to its holder, an
  authorized reader.

### 4.6 Adapter (`objectstore`) and provider security (R1, R6)

- **Keys.** The interface takes only `Key` values that match `^objects/<uuid>/<uuid>$`. Traversal, encoded separators,
  uppercase and anything else are refused before any request. Errors are closed categories; raw provider or SDK text
  never escapes. Classification uses `aws/transport/http.ResponseError` status and the s3 `types` NotFound/NoSuchKey
  from already-required modules.
- **Upload.** Bodies stream unseekable with `UNSIGNED-PAYLOAD` (`v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware`),
  `RequestChecksumCalculation=WhenRequired` and an explicit `Content-Length`. There is no buffering.
  - The upload checks are the exact length, a core-computed local digest, and Head verification of the stored length
    and binding. These do **not** independently prove the stored payload (§4.4).
  - Core `GetObject` verifies content against the digest; signed downloads are unverified by core.
  - TLS is mandatory for production data transport (§4.7; escalation E4, conditionally accepted).
- **Constructor (r1).** `objectstore.New(cfg, appEnv)` refuses any configuration that breaks the §4.7 transport rules,
  even when built without `config.Load` (tests, future adapters). `access.NewServiceWithObjects` re-checks the
  `ObjectConfig` bounds and refuses invalid values.
- **Credentials.** Only the existing scoped `OBJECT_STORE_*` credential, read through `config.SecretProvider`. Core never
  receives MinIO root credentials. Web and intelligence remain without storage settings (B-BOUNDARIES-ROOT,
  unchanged). The existing `core-objects` policy already limits Get/Put/Delete to the bucket, plus List/Location, so
  `infra/minio/init.sh` is unchanged.
- **Importers.** Only `access`. `health` keeps its own HeadBucket readiness check (unchanged, separately classified).
  Test setup uses the adapter under the scoped credential.

### 4.7 Configuration (`config.Load`, optional; not in `.env.example`)

| Variable | Default and bound |
|---|---|
| `OBJECT_STORE_SIGNING_ENDPOINT` | default `OBJECT_STORE_ENDPOINT`; an absolute http(s) URL with host and optional port, without userinfo, path, query or fragment |
| `OBJECT_STORE_MAX_OBJECT_BYTES` | 26214400, range 1..1073741824 |
| `OBJECT_STORE_SIGNED_URL_DEFAULT_TTL` | 60, range 1..300, ≤ max |
| `OBJECT_STORE_SIGNED_URL_MAX_TTL` | 300, range 1..300 |
| `OBJECT_STORE_TRANSFER_TIMEOUT` | 120, range 1..900 (seconds) |

**Endpoint transport rules (r1, correction 2):**

| Rule | Applies to |
|---|---|
| The existing required `OBJECT_STORE_ENDPOINT` must also be an absolute http(s) URL of the same shape. Every existing valid value passes (`http://minio:9000`, `http://minio.invalid:9000`). | all environments |
| `APP_ENV=production`: **both** the data endpoint and the signing endpoint must be `https`. | production |
| Mixed configuration (an `http` data endpoint with an `https` signing endpoint) is refused. | all environments |
| `http` for both endpoints is permitted. | development and test (local MinIO) |

- **Loader/constructor boundary.** `config.Load` enforces the rules at startup. Problems name only the variable;
  existing defaults and error secrecy are unchanged. `objectstore.New` re-enforces them for every storage
  construction (§4.6).
- **Existing tests unchanged:** `config_test.go` and `exit_test.go` use test or development with `http` endpoints.
- **The one inherited probe affected:** B-IDENTITY-CONFIG's production base environment. It moves to an https data
  endpoint so case cc8 still starts in production (RC in §5.6).

Invalid values fail startup and name only the variable. The five new keys join the `core.exit` `config_keys` enum.
Server startup validates them but uses no storage operation until a later reviewed adapter wires `access`; `main.go`
is unchanged. Later binary wiring is documented in OBJECT-STORAGE.md. Tests construct `ObjectConfig` from the same
loader (Q5 ruling: in scope).

### 4.8 `storage.operation` record (R9; operational evidence, not G011 audit)

Exactly one record per guarded storage call, catalogued in `events.json`.

**Fields:**
- `operation`: `object.put`, `object.get`, `object.metadata`, `object.sign` or `object.delete`.
- `outcome`: `allow`, `deny` or `failed`. Level: info, warn or error.
- `reason`, from a closed set:
  - the G007 reasons;
  - `not_creator`, `object_not_ready`, `objects_unconfigured`;
  - `provider_unavailable`, `provider_outcome_unknown`, `object_bytes_missing`;
  - `binding_mismatch`, `integrity_mismatch`, `compensated`, `compensation_failed`, `reconciled_ready`;
  - `source_not_released` (r1).
- `stage` (Put and Delete): `t1`, `provider`, `verify`, `t2`, `reconcile`, `compensate`.
- UUIDs: `actor_id`, `organization_id`, optional `group_id`, and `object_id` (absent before the insert).
- `ttl_seconds` (sign only); `physical` (delete only): `confirmed` or `pending`.
- `request_id` and `trace_id` when present.

**Never:** URL, signature, credential, key, payload, content type, size, filename, raw provider or driver text.

The deliberate difference from `access.decision`, which has no UUIDs, follows R9. Storage calls do not also emit
`access.decision`. The G011 handoff maps these fields to the future `audit` schema (ADR-006); no audit table is created.

### 4.9 Documentation

**OBJECT-STORAGE.md:**
- the operations and the §4.2 matrix;
- the state diagram and I-FIN;
- settings and bounds;
- the local-host and in-network signing model;
- TTL and revocation residuals;
- the failure, repair and retry procedure (authorized Delete of a pending, ready or deleting object; no daemon, queue
  or bucket scan);
- R2 production configuration: endpoint `https://<account>.r2.cloudflarestorage.com`, region `auto`, a bucket-scoped
  API token, path-style and checksum settings, presign and response-override support. Each item is checked against the
  [R2 S3 API reference](https://developers.cloudflare.com/r2/api/s3/api/) during implementation, with exact citations
  and unsupported items listed. There is no live R2 and no parity claim;
- the harness commands;
- the G010, G011 and G013 handoffs.

**Updates:**
- DATABASE-ISOLATION: version 7, 29 policies, the table, its grants and its footprint.
- AUTHORIZATION: storage guarded operations and the mapping.
- MIGRATIONS: version 7, the docs blocks and the down refusal.
- VENDORS: R2 row notes; local-only MinIO unchanged.
- FOLLOW-UP: the handoffs.

## 5. Tests and validation design

Every denial or fault case is paired with a working authorized positive, and both systems' owned state is inspected.
Spy and fault stores provide determinism. They never substitute for real provider success, expiry or isolation.

### 5.1 Unit (core-check image; B-UNIT-CHECKS)

| ID | Covers |
|---|---|
| U-OBJ-KEY | key derivation and adapter refusal of every malformed key form (traversal, `%2F`, `..`, uppercase, other prefixes) |
| U-OBJ-INPUT | the §4.3 table, including overflow and both TTL forms; denials make **zero** database or provider calls (spy) |
| U-OBJ-READER | exact reader: short, long, exact, zero, and an infinite source reads exactly Size+1 bytes. **Holdback:** the final byte is withheld until lookahead EOF, and on a long source the reader errors with Size−1 bytes released. Owned source (r1), each asserting `Body` closed exactly once, the call returning within the transfer deadline plus the reap grace, and the goroutine count back at baseline: (a) blocked first `Read` (an `io.Pipe` with no writer), unblocked by cancellation and by the deadline; (b) blocked final lookahead; (c) Size-0 probe blocked, and with extra data; (d) a non-conforming source that ignores Close → `source_not_released` within the grace; the test then releases its fake source and asserts the baseline. Also the bounded verifying download reader, and heap growth while streaming 25 MiB through a discarding store stays below a pinned bound. |
| U-OBJ-POLICY | `decideObject` on synthetic facts against the oracle cases; U-POL-MATRIX unchanged |
| U-OBJ-ADAPTER | httptest classification (404, 403, 5xx, timeout, reset) and presign shape (GET, expiry, overrides, signing host). No raw text escapes. |
| U-OBJ-CONFIG | defaults, bounds, variable-name-only problems. **Transport (r1):** production https data plus https signing passes; production http data is refused; production http signing is refused; mixed http-data/https-signing is refused in development, test and production; http for both passes in development and test. The same table runs against `objectstore.New` (constructor re-enforcement). The existing `config_test.go` is unchanged. |
| U-OBJ-EVENT | records conform to `events.json`; forbidden fields absent |
| U-OBJ-SEAM | production fault seams nil and inert. **Budgets (r1),** with a fake `pgx.Tx` in the `commit_test.go` style: a storage transaction's statements carry the 5 s decision deadline; its deferred rollback context carries the 5 s cleanup deadline and is not canceled by the caller; a rollback blocking until its context ends returns within the budget. The `cleanup == 0` path (existing operations) still uses an undeadlined `WithoutCancel` context, so existing operations are unchanged. `TestCommitFailureTransaction` still passes (one BEGIN, one COMMIT, the deferred rollback). |
| U-ACCESS-BOUNDARY (updated) | files, imports and exported surface; objectstore and aws-sdk importers; no SQL in `access` |

### 5.2 Integration: real PostgreSQL (runtime login) and real private MinIO (scoped credential)

These run in the core-check image on the migrations project's data network, with the G008 attestation pattern:
- the runtime login is not the owner, not a superuser and has no BYPASSRLS;
- the MinIO credential is the scoped user;
- `MINIO_ROOT_*` is absent from the test environment.

Privileged setup is separately identified (the DB superuser fixture, as in G008). Assertions never use owner, root or
superuser credentials. Each test records the database rows and provider keys it creates, and `t.Cleanup` removes
exactly those (superuser fixture for rows, scoped adapter for keys).

| Test (check) | Cases |
|---|---|
| OBJ-ATTEST (IT) | runtime and scoped-credential attestation, plus a privileged-assertion negative |
| OBJ-RT (IT) | Put → Metadata → Get for Organization and Group objects; zero-byte; exactly max; max+1 rejected before the provider. Compares bytes, length, type, SHA-256, the provider binding metadata and the derived key. |
| OBJ-READ (IT) | oracle read matrix (Metadata/Get/Sign); selected, foreign, sibling and shared-BusinessProfile cases; the same global UUID across Organizations; forged, stale, expired, revoked and other-scope `TenantContext` (B1–B6); non-ready objects. The spy proves no provider call on denial. **r1 scope discovery:** Organization-context reads of an exact Group object by the s1 and s50 subjects (1 and 50 Group bindings) allow with identical statement counts and fact rows. The FactQuery is asserted to name only the owning Group (plus the selected one), never nil. Selected-other-Group, foreign-Organization and shared-profile sibling objects are `not_found` with zero provider calls and no row disclosed. An invalid context with an unknown object is `context_denied`. |
| OBJ-WRITE (IT) | oracle Put/Delete matrix; member Put and own Delete; another uploader's object in the same Group; foreign, sibling or unselected Group; revoked membership or session; Organization mutation by non-admins and by a former-admin uploader; admin-to-member own/other split. Both systems unchanged on denial. |
| OBJ-RLS (IT) | direct runtime SQL without app filters per oracle actor and context: SELECT sets; INSERT (forged creator, wrong Organization or Group, wrong state); UPDATE transitions and immutable columns (privilege 42501 vs trigger); DELETE always 42501. Empty, malformed (22P02/22023) and poisoned pooled context, with discard. **FORCE owner (r1, correction 6):** the owner session in a valid Organization A context sees exactly A's in-scope objects (positive). In the same session it sees none of Organization B's objects; INSERT of a B-scoped row fails WITH CHECK; UPDATE of B's rows affects 0 rows. With selected Group G1 it does not see G2 objects. With no context it sees nothing. Same setup as the runtime cases. |
| OBJ-SCHEMA (IT) | zero id; cross-Organization Group FK; lifecycle check; the generated key cannot be written; a duplicate id is a unique violation with the original row and bytes unchanged; forged binding (provider object tampered with the scoped credential) → Get `internal`, nothing returned |
| OBJ-DEL (IT) | real physical deletion (Head 404) and tombstone; repeat → `deleted`; missing or foreign → `not_found`; Get/Metadata/Sign afterwards → `not_found`; denied delete leaves both systems unchanged |
| OBJ-SIGN (IT) | real fetch with the default 60 s and with 300 s; 0, −1, 301 and both-forms rejected. Altered key, method, host, signature or expiry → 403. Anonymous unsigned → 403. Real expiry with TTL 2 s (403 after it); reuse within the TTL works. Response overrides present. **Loopback binding:** a URL signed for `http://127.0.0.1:<port>` and sent to MinIO with that exact `Host` → 200; with any other `Host` → 403. **OBJ-SIGN-HOST (r1, correction 4):** the unchanged URL, signed for the migrations project's real published MinIO API endpoint `http://127.0.0.1:<published port>`, is fetched by the harness through that host port (§5.5). It returns 200 with the expected body digest and the response overrides. The same path without the query returns 403. URLs are handled in memory only and never logged, stored or placed in evidence. |
| OBJ-ADMIN (IT) | the scoped credential cannot CreateBucket, DeleteBucket, Put/GetBucketPolicy or use the MinIO admin API (SigV4-signed raw request → 403); anonymous list, put and get are denied |
| OBJ-LOG (IT) | records catalogued. Canaries absent from captured logs, test output and the PostgreSQL log: payload bytes, a content-type canary, the secret key, the signed-query `X-Amz-Signature` and `X-Amz-Credential`. A planted canary is detected (G008 logcheck pattern). |
| E-OBJ-1..4 (IT, footprint) | §10 |
| OOR-1..8 (ORDERING) | deterministic, sleep-free interleavings with conflicting writers on the attested setup pool and stage hooks (G008 pattern). (1) Group removal committed before Put T2: denied, compensated, never readable. (2) Removal racing T2 that holds the facts locks: the remover waits; then the next read is denied. (3) Delete D1 vs Put T2, both orders. (4) Admin→member downgrade: own delete allowed, others' denied, reads kept. (5) Session revoke between T1 and T2. (6) DisableUser during an in-flight Get: the transfer finishes, the next call is denied. (7) Sign, then revoke: the URL works until expiry, new issuance is denied, the URL fails after expiry. (8) Delete committed while a Get stream is open: no new read; bounded residual recorded. |
| FT-*/FP-*/FD-*/FG-*/FX-* (FAULTS) | every §4.4 transition, with real database and MinIO state inspected afterwards. T1: statement error, acknowledged-then-unknown COMMIT, FATAL before COMMIT. P2: error before send, timeout after body, short/long body. **r1 real-MinIO source cases:** (FP-LONG) an overlong source: the verdict is `long`, the row stays non-ready, C runs, Head shows the key absent; a forced provider-success variant (seam) is still compensated and reported `object_size_mismatch`. (FP-BLOCK1) blocked first `Read` with a 1 s transfer deadline. (FP-BLOCKN) blocked final lookahead. (FP-ZERO) Size 0 with an empty source succeeds with stored length 0; Size 0 with extra data → mismatch and no provider call. Each case checks termination within its bound, `Body` closed once, the goroutine baseline, row state and the provider key state. V: mismatch or unavailable. T2: deny, abort, unknown-but-committed (**bytes kept; R reports ready**), unknown-and-reconcile-unavailable (bytes kept), unknown-not-committed (R pending → C). Cancellation at each stage. Simulated process interruption after each stage, then an authorized Delete repairs. Compensation failure. D1 unknown, D2/D3 failure, D4 failure → `deletion_pending`, then a retry completes. Get: NoSuchKey, unavailable, binding and integrity. **FX-STALL (r1, correction 5):** a test-local freezable TCP forwarder (stdlib only, inside the test) sits between a runtime pool and PostgreSQL. It freezes traffic during a storage write transaction after its first statement. The call returns `unavailable` within decision plus cleanup budgets plus a small slack. The stalled connection is destroyed: the next acquire gets a new backend PID. T1 committed nothing, there are no leaked goroutines, and an existing G007 guarded read on an unfrozen pool is unaffected. **Pending and deleting objects never serve new reads or signs.** |
| TestObjectsMigration (MIGRATION) | empty → 7; populated G008 world 6 → 7; repeat; down refused with rows (version and rows intact); down and up on empty; exact v6 ↔ v7 catalogs; restart persistence (the harness writes an object, restarts postgres and minio, and the guarded Get returns identical bytes); old migrations unchanged |

**Fault seams.** Unexported, nil-by-default stage hooks in `access` and a fault-injecting `objectstore.Store` wrapper
in test files. The acknowledged-then-unknown COMMIT follows B-RLS-IT P8. FATAL-before-COMMIT uses
`pg_terminate_backend` from the setup pool.

### 5.3 Catalog and negative controls (B-OBJ-CATALOG; harness SQL, the `rls_catalog` pattern)

The exact version-7 delta (§3.2, §3.5) and the unchanged version-6 objects.

**Deliberate missing-protection detectors** (scratch database, each must make the check FAIL, then is discarded):

| ID | Mutation |
|---|---|
| N-OBJ-1 | NO FORCE |
| N-OBJ-2 | extra runtime DELETE |
| N-OBJ-3 | SELECT policy `USING (true)` |
| N-OBJ-4 | trigger disabled |
| N-OBJ-5 | a SECURITY DEFINER guard |
| N-OBJ-6 | a PUBLIC grant |
| N-OBJ-7 | key column not generated |

### 5.4 Independent oracle

See §7. Glue failures report `ORACLE-DISAGREEMENT`, never auto-resolved. A disagreement that reveals Spec ambiguity
goes to Codex.

### 5.5 Harness changes (minimum, existing structure)

- **`run.sh`, suite `migrations`:** after the existing RLS functions, it starts `minio` from local.yml in the
  disposable project.
  - It initializes the bucket and scoped user by running the built init image with `docker run -i`, the init script on
    stdin and an `--env-file` written under `$WORK` with umask 077 and deleted after use.
  - This adds **no new per-run host bind mount** (Docker Desktop rule).
  - Then it runs `obj_catalog`, `obj_migration`, `obj_it`, `obj_ordering` and `obj_faults`. The test env-file adds the
    six `OBJECT_STORE_*` values (scoped) and never MinIO root.
  - Every container and project is registered through the existing `docker_run_owned` / `register_project` inventory
    at creation.
  - **OBJ-SIGN-HOST (r1)**, inside `obj_it` and recorded under B-OBJ-IT, with no new suite:
    1. A detached core-check container runs the issuing subtest with
       `OBJECT_STORE_SIGNING_ENDPOINT=http://127.0.0.1:<the project's published MinIO API port>`.
    2. It writes the URL and the expected digest to a `--tmpfs` path (not a host bind mount) with mode 0600, then waits
       up to 60 s.
    3. The harness reads them with `docker exec … cat` into shell variables. Exec output is not kept in container logs.
    4. It fetches the unchanged URL through the published port with the existing host-probe mechanism (`PROBE_MODE`
       devtcp or curl, as B-PORTS uses). The request is `GET <path?query>` with `Host: 127.0.0.1:<port>`, exactly what
       a browser sends. Curl reads the URL from stdin (`-K -`), so it never appears in an argument list.
    5. It compares the status, the body digest and the override headers, unsets the variables and removes the container.
  - With `PROBE_MODE=none` the step is recorded **UNMET**, never PASS. Nothing secret reaches disk, logs or the report
    directory. B-SECRET-CANARY scans the run as today.
- **`required-checks.tsv`:** B-OBJ-CATALOG, B-OBJ-MIGRATION, B-OBJ-IT, B-OBJ-ORDERING and B-OBJ-FAULTS (suite
  `migrations`); totals updated.
- **`events.json`:** §4.7 and §4.8. **`profile-paths.tsv`:** §2. No runner, selector logic, `.gitlab-ci.yml` or new
  framework change.

### 5.6 Version-7 reconciliation of accepted checks (RC list; guarantees preserved, never weakened)

| Location | Change |
|---|---|
| `run.sh` lint `static_migration_set` (~985–994) | add `MIGRATION_00006_SHA` (00006 byte pin) and 00007 to the exact set |
| root migrate block (~1455–1468) | add 00007 and `current_version 0/7/7` |
| migrations B-MIGRATE-CLI/CATALOG (~1996–2007) | reapply/rollback now act on 00007; app relations 35 → 38 at version 7 |
| ops B-PERSIST (~3029) | version 7 |
| rename B-DBRENAME (~3080) | version 7 |
| off B-TOGGLE-OFF (~3196) | version 7 |
| `migrations_test.go`, `runner_test.go` | set through 00007; 00006 pinned |
| `migrate/integration_test.go`, `migrate/rls_integration_test.go` | outputs for the latest version 7. MG-RLS tests that exercise 00006 itself first step down from 7 on an empty object table. |
| `tenancy/integration_test.go` | rolls back two steps to version 5 |
| access/identity RLS tests asserting exact catalogs | "nine forced tables" becomes the v6 set plus `stored_objects` |
| `run.sh` identity B-IDENTITY-CONFIG (`identity_config`, ~3568) | r1: the production base environment's data endpoint changes from `http://minio.invalid:9000` to `https://minio.invalid:9000`. cc1–cc7 still each exit with one `core.exit config_invalid` line, and cc8 still starts in production (`--network none`, no connection). The guarantee is unchanged. |
| `access/commit_test.go` | the `inTx` call passes `cleanupUnbounded`; assertions unchanged |

The implementer inventories every remaining pin with `grep -nE '00006|version 6|current_version|nine|26 policies'`
before editing and records each RC in the Result. Object tests clean their rows so later rollbacks are not blocked;
the `obj_*` functions run after the RLS ones.

## 6. Validation batches, CI selection, provenance and cleanup

**Before every batch:**
1. `git fetch origin`.
2. Apply the **exact-delta rule**: assess the staging delta against validation inputs (code, harness, fixtures, images,
   docs blocks consumed by checks) and against authority (Goal/Spec/Plan/governance/bindings). If it is demonstrably
   irrelevant to both, record both SHAs and the reason in the Result and defer the rebase. Otherwise rebase and
   reconcile.
3. `bash infra/ci/docker-desktop-mounts.sh`; stop on failure.
4. Take the harness `flock`.
5. Record the target and candidate SHAs.

Reconciliation is mandatory before final qualification and again before the MR. The standard harness form is
`bash infra/validation/run.sh --base <target> --allow-file management/v0/results/G009-scope-allowlist.txt --report-dir
<EVID>/<batch> <suites>` (no snapshot wrapper; Git-first new work).

| Batch | Content | Purpose |
|---|---|---|
| b1 focus-go | gofmt/vet/unit for `config`, `objectstore`, `tenancy` and `access` in the pinned Go image | inner loop; not acceptance |
| b2 focus-db | selected `-tags integration` tests on a focus stack (pinned postgres image plus the built MinIO image) | inner loop; container, volume and network IDs recorded in `<EVID>/focus-inventory.tsv` at creation |
| b3 covering | `lint lint-base lint-core unit migrations` | B-MIGRATE-SET, digests, all migrations-suite checks at version 7 including B-RLS-* and B-OBJ-*; finish → fix → rerun failed |
| S-review | read-only security review of the candidate diff (§7) | findings fixed, then the affected b1–b3 rerun |
| b4 final | reconcile, then `all unit` | `all` is the full profile and includes **ci-controls**, because `infra/validation/*`, `events.json` and `profile-paths.tsv` change (`ci_controls_required`). `unit` is added because full has none. B-REQUIRED-CHECKS: every required ID PASS exactly once. |

**CI selection** (actual paths): `core/migrations/`, `core/internal/{config,tenancy}/` and `infra/` select **full**;
`infra/*` also selects **ci-controls**; docs alone would be fast. There is no feature-core area. **No manual or
operator stage** is planned (no UI). The Q3 ruling needs no human step; the published-host proof is automated
(OBJ-SIGN-HOST).

**Hosted:**
1. After Codex's pre-integration review: reconcile, push, then `glab` MR to `staging` with source-branch deletion.
2. Auto-merge is armed per the 2026-10-14 default once review and evidence holds clear.
3. The MR **fast** job runs automatically. The manual gate is optional and not requested by default; b4 is the local
   full evidence.
4. The post-merge staging pipeline (path-selected full plus ci-controls) is **required**. Success or owner-canceled
   satisfies it (cancellation recorded as canceled); a failed job fails; pending or running does not qualify.

**Provenance (Git-first):**
- The Result records, per batch: candidate and target commits, exact commands, outcomes (failed, interrupted, skipped
  explicit), image IDs/digests (core-check, pgvector, built MinIO with release tags), the Go version and CI job links.
- Dirty or untracked relevant changes are identified; scoped checkpoints are committed before b4. Extra hashes only for
  non-Git inputs.
- Evidence is reused only for exact unchanged content and inputs. A changed SHA alone triggers no rerun.
- Never retained: secrets, signed URLs, payloads, session records.

**Bounded cleanup:**
- The harness already registers projects and owned containers. Focus resources are written to the run inventory at
  creation: container/image IDs and volume names before container removal.
- Closeout: one scoped preflight, one removal pass and a compact receipt.
- MinIO data exists only in disposable project volumes (`down -v`). There is no bucket-wide deletion and the developer
  stack is never touched.
- Uncertain items are held with a reason; native refusals are reported once.

**Safekeeping:** commit and push `g009-tenant-object-storage` after each implementation or correction round (no MR).

## 7. Delegation

| Role | Owner of | Effort | Notes |
|---|---|---|---|
| Implementer (integration owner, single stack owner) | every §2 path except the oracle file | **high** (owner default; the operator may choose xhigh) | Worktree `g009-agent-<id>-impl`, or a recorded Codex exception if tooling cannot. Integrates and validates everything. |
| Oracle delegate | `objects_expected_test.go` only | high | Inputs: Spec R2/R3/AC3/AC4, write-policy r1, ARCHITECTURE §40 and §53.7, AUTHORIZATION.md (context, gates, forbidden/not_found), DATABASE-ISOLATION.md context validity, ADR-004 and a Plan-supplied fixture/case schema. **No access** to 00007, `objects*.go` or Plan §3.2/§4.2. The header pins its sources by SHA-256. Outputs `objectsExpectedFixture()` and `objectsExpectedCases()` (actor × context × object × action: read, put, delete and the RLS select/insert/transition forms, each with allow/deny, external code and paired positives). Dispatched early, in parallel. |
| Security reviewer | none (read-only) | high | RLS, grants, trigger, guard and scope-discovery ordering, I-FIN, completion classification and compensation, source ownership and reaping, transaction budgets and the `tx.go` edit, transport rules, URL secrecy (including OBJ-SIGN-HOST handling), redaction, credentials, harness secret handling; before b4. Cannot grant acceptance. |

No separate adapter worker: the implementer keeps the overlap risk at zero. Delegated opinions never replace Codex
acceptance.

## 8. Sequence

1. **Binding** (after G008 final ACCEPT and Plan ACCEPT): fetch; exact-delta check of G008 (§0); create the
   implementation worktree and branch from staging; commit the scope allow-list; record the binding in the Result.
   Dispatch the oracle.
2. Migration 00007, tenancy primitives, the migration tests and the version-7 RC pins (§5.6).
3. Objectstore adapter, config, unit tests.
4. Access operations, `decideObject`, readers, seams, logging, boundary pins, unit tests.
5. Integration, ordering, fault and footprint tests; oracle glue; harness `obj_*` functions, required checks,
   `events.json`, the profile row.
6. Docs.
7. b1 → b2 → b3, with finish → fix → rerun → advance and checkpoint pushes.
8. Security review, fixes and affected reruns.
9. Reconcile, b4, the Result, push, and hand back to Codex as review-ready.
10. After Codex pre-integration ACCEPT: reconcile, MR, fast, auto-merge after the holds, the staging disposition, and
    the G010/G011 handoffs. Codex alone marks final ACCEPT.

## 9. Acceptance-criteria mapping

| AC | Tests and checks | Evidence |
|---|---|---|
| AC1 | U-ACCESS-BOUNDARY, U-OBJ-KEY, U-OBJ-INPUT (no adapter call), B-MCP-SOURCE, B-TENANCY-HTTP, `contracts/` and routes unchanged (B-SCOPE) | unit and lint-core PASS; allow-list diff |
| AC2 | OBJ-SCHEMA, OBJ-RLS (creator and key forgery), B-OBJ-CATALOG | IT and catalog PASS |
| AC3 | OBJ-READ with the oracle (incl. r1 scope discovery at 1/50 bindings and the negatives), U-OBJ-POLICY, spy zero calls | ORACLE 0 disagreements |
| AC4 | OBJ-WRITE with the oracle (incl. admin→member, former-admin Organization object), OBJ-RLS transitions, unchanged-state checks | IT PASS, oracle 0 disagreements |
| AC5 | OBJ-RLS (incl. the FORCE-owner positive, **foreign-Organization/scope denial** and no-context), OBJ-ATTEST, B-OBJ-CATALOG N-OBJ-1..7, unchanged B-RLS-* | catalog plus IT |
| AC6 | OBJ-RT (real MinIO, scoped credential) | IT |
| AC7 | OBJ-DEL, FD-* | IT, FAULTS |
| AC8 | OBJ-SIGN (real fetch, expiry, reuse, alteration, anonymous, loopback binding) and **OBJ-SIGN-HOST** (unchanged URL through the published host port) | IT |
| AC9 | OOR-1..8 | B-OBJ-ORDERING |
| AC10 | U-OBJ-INPUT, U-OBJ-READER (blocked first Read, blocked lookahead, zero length, non-conforming source), U-OBJ-SEAM budgets, OBJ-RT boundaries | unit plus IT |
| AC11 | FT/FP/FD/FG-* incl. FP-LONG/BLOCK1/BLOCKN/ZERO and FX-STALL | B-OBJ-FAULTS |
| AC12 | TestObjectsMigration, the §5.6 RC pins, full root/ops/rename/off, B-RLS-*, B-POLICY-* | b3/b4 full PASS |
| AC13 | OBJ-LOG, OBJ-ADMIN, U-OBJ-EVENT, U-OBJ-CONFIG transport rules, B-SECRET-CANARY, B-IDENTITY-CONFIG (RC), OBJECT-STORAGE.md (R2 citations, TTL, revocation, failures, the narrowed integrity claims; no live-cloud claim) | IT plus docs review |
| AC14 | the Result (commits, commands, real vs injected evidence, failures, B5 ledger, cleanup receipt), b4, the staging disposition, the G010/G011 handoff without G021 release | Codex review |

## 10. PG-003 assessment (performance; no numeric promise)

- **Status.** The [PG-003 r0 assessment](../PERFORMANCE.md#pg-003-plan-r0-assessment--2026-10-14) approves
  bounded E-OBJ-1..4 collection during authorized execution. There is no benchmark, no numeric acceptance and no
  inherited waiver; the reader bounds below are hypotheses.
- **Queries.** Metadata access is a point lookup by primary key plus `organization_id`. Every storage path discovers
  the immutable object scope first. Facts are then read only for the owning Group (plus the selected Group), or the
  Organization-only set (r1, §4.4). There is never an all-binding read and no Organization-only profile fan-out.
- **Locks.** Writes take only `FOR SHARE` on the G007 rows and `FOR UPDATE` on the one object row.
  - There is no exclusive Organization or Group lock, so there is less contention than G007 renames.
  - FK inserts take `FOR KEY SHARE`, which waits only on the accepted slug-change `FOR UPDATE` and on DisableUser.
  - The lock order is acyclic with G007 writers.
  - The known sign-in rotation 40P01 is a definitive abort, compensated.
- **Lifetimes.** Each transaction runs about 8–10 statements under the 5 s decision budget plus the 5 s cleanup budget
  (§4.3a). Every provider transfer happens outside any transaction or connection, and the upload source is
  interruptible. E-OBJ-4 shows no runtime backend sitting `idle in transaction` while a stub transfer is blocked.
  FX-STALL shows the cleanup bound.
- **Expected reader calls per object statement** (one candidate row; same-transaction deltas of `app.rls_org_role` and
  `app.rls_group_role`):

  | Statement | `rls_org_role` | `rls_group_role` |
  |---|---|---|
  | SELECT | ≤ 2 | ≤ 1 |
  | INSERT | ≤ 2 | ≤ 1 |
  | SELECT … FOR UPDATE | ≤ 4 | ≤ 3 |
  | UPDATE | ≤ 6 | ≤ 5 |

  These are hypotheses, constant with respect to the subject's binding count and the object count.
- **Evidence:**
  - E-OBJ-1: EXPLAIN summaries on a populated fixture of about 200 objects across 2 Organizations. Small-table
    sequential scans are recorded as observations.
  - E-OBJ-2: the locked-row inventory equals the expected set.
  - E-OBJ-3: reader-call deltas and fact-row counts are equal for the s1 and s50 subjects, for writes and for
    Organization-context reads of a Group object.
  - E-OBJ-4: observations of transaction and transfer separation.
- **Prompt escalation (G008 lesson):** growth with bindings or objects, a statement above its bound, an unexpected lock
  or scan on `stored_objects`, or any provider I/O under a transaction. No benchmark, index tuning or threshold
  relaxation.
- **Not provided:** per-subject admission, rate limits and request deadlines stay future-adapter obligations
  (PG-001/PG-003).

## 11. Escalation items for explicit review; risks; residuals

| ID | Item | Class |
|---|---|---|
| E1 | New closed storage-action mapping beside G007's eleven capabilities (§4.2) | authorization/security model |
| E2 | Runtime INSERT (6 columns) and UPDATE(state, sha256) on a new tenant table, 3 policies, invoker trigger; no role, definer or helper grant (§3) | security model / RLS |
| E3 | Durable data-model choices: generated key `objects/<org>/<id>`, tombstones (no hard delete), down refused with rows, no filename | irreversible data model |
| E4 | `UNSIGNED-PAYLOAD` + `WhenRequired` checksums for unseekable streaming over local HTTP. Upload checks prove local digest, length and binding only (narrowed, r1). Core Get verifies content. TLS is mandatory for production data and signing (r1). Conditionally accepted, subject to corrections 1 and 2. | security trade-off / provider semantics |
| E5 | Internal exported API on `access` (r1: `Body io.ReadCloser` ownership contract) and new `objectstore` package; boundary pins updated | interface (no public API/MCP) |
| E6 | Five optional config keys, plus r1 shape and production-https validation of the existing data endpoint and the mixed-scheme refusal, at loader and constructor | configuration (Q5: in scope) |
| E7 | Harness: version-7 pins, 5 check IDs, OBJ-SIGN-HOST inside B-OBJ-IT, `events.json`, one `profile-paths.tsv` row (selects ci-controls) | bounded CI change (Q6: accepted) |
| E8 | Version-7 edits to accepted G005–G008 tests (§5.6), plus the B-IDENTITY-CONFIG production-probe https data endpoint and the `commit_test.go` call update, each preserving its guarantee; the G008 oracle is untouched | inherited checks |
| E9 | If the SDK forces a direct `smithy-go` import, `go.mod` moves it from indirect to direct (no new module, version or vendor); otherwise unchanged | dependency classification |
| E10 | Storage-only 5 s decision and 5 s cleanup budgets (existing operations unchanged) | behavior |
| E11 | Scoped `tx.go` helper edit: a `cleanup` parameter, with `0` preserving existing behavior (§4.3a) | G008 boundary helper |

**Implementation risks:**
- R-a: SDK streaming over HTTP may still refuse the body. Fallback: escalate before any buffering design. No silent
  spool.
- R-b: many harness pins (§5.6).
- R-c: real-expiry and timing tests must stay deterministic; there are no sleeps except the bounded TTL wait.
- R-d: hosted ci-controls drift.
- R-e: a G008 correction (§0).
- R-f (r1): if the SDK or `net/http` keeps a `Read` blocked beyond the source's `Close` (a non-conforming path), the
  call still returns within the reap grace (X6). If the conforming-source tests show a leak, stop and escalate before
  adding any worker framework.

**Residuals** (documented, not prevented):
- X1: bearer URLs stay valid ≤ TTL (≤ 300 s) after revocation or logical deletion until physical removal; in-flight
  transfers complete.
- X2: an unknown T1 COMMIT can leave an inert pending row without bytes.
- X3: a failed compensation or a late in-flight PUT can leave inaccessible bytes under a non-ready row. Repaired by
  Delete.
- X4: a holder of the runtime credential with arbitrary SQL can finalize its own pending row without bytes (the
  existing G008 trust limit; a later read is `not_found`).
- X5: local MinIO is unmaintained, and R2 is not cloud-qualified.
- X6 (r1): a source that violates the Close-unblocks-Read contract can leave its own blocked `Read` behind. PutObject
  returns `unavailable` (`source_not_released`) within the grace, never publishes and compensates.
- X7 (r1): signed downloads are served directly by the provider; core does not verify their content against the
  digest.

## 12. Budget, forecast and scope

**B5**, shared attempts, with a checkpoint after attempt 3. Planning spends none.

**Forecast: 2 (likely) to 3**:
- attempt 1: implementation through b4 and review-ready;
- attempt 2: one bounded Codex correction;
- attempt 3: only for hosted or ci-controls drift or a G008 reconciliation;
- 4: only if R-a forces a design change.

**Context:** one bounded storage increment. It fits a single implementation cycle with the oracle delegate. No split
is recommended. If the checkpoint after attempt 3 shows B5 at risk, propose to Codex separating the harness RC
reconciliation from the operations, never weaker tests.

## 13. Closeout

G009 is not the last Goal of EPIC-003 (G010 closes it), so the epic-closeout PROGRESS compaction checkpoint does not
apply. The ordinary post-ACCEPT housekeeping assessment is Codex's. Local bulky evidence is purged under the closeout
policy after final ACCEPT, keeping the compact record.

**Handoffs:**
- **G010:** the storage world, the oracle and the attack surfaces.
- **G011:** the `storage.operation` mapping and issuance/failure categories for durable audit.
- **G013:** tombstone and pending-row retention.

## 14. Questions and rulings

r0 questions, as ruled in the r0 review (not included):

| # | Ruling | Applied in |
|---|---|---|
| Q1 | E1–E3 accepted in principle (mapping, scoped RLS and grants, generated key, immutable attribution, tombstones). E4 accepted subject to corrections 1 and 2. | §3, §4.2, §4.4, §4.7 |
| Q2 | An authorized, idempotent repeat Delete returns `deleted` when absence is confirmed; pending outcomes stay truthful | §4.4 |
| Q3 | No human stage; an automated published-host proof is required | §4.5, §5.2, §5.5 (OBJ-SIGN-HOST) |
| Q4 | The pins are reasonable once the bounds are real; they are limits, not latency or concurrency guarantees | §4.3, §4.3a |
| Q5 | Optional defaulted settings are in scope; the active constructor enforces them and the production transport rules | §4.6, §4.7 |
| Q6 | The minimal version-7, catalog, event and profile reconciliation is accepted; the smithy direct-import classification is allowed without a version change | §5.5, §5.6, E9 |
| Q7 | No filename or display name | §3.1 |
| Q8 | Bounded E-OBJ evidence during authorized execution; no benchmark or inherited waiver; PG-003 stays distinct | §10 |

**No new open questions in r1.** One item to confirm in re-review: §5.6's B-IDENTITY-CONFIG production probe moves to
an https data endpoint. This is the minimum reconciliation forced by correction 2, and it preserves that check's
guarantee.

## Revision history

- r0, 2026-10-14 UTC: initial provisional Plan on `bbb2173` (G008 integrated at `81b186c`). It records final G008
  ACCEPT, Codex Plan ACCEPT and execution binding as implementation gates. Nothing was executed.
  Exact archive (not included); reviewed commit `011450f`.
- r1, 2026-10-14 UTC: bounded amendment for the r0 review (not included) CORRECTION LOOP,
  on staging `c944ef7`. It makes six corrections (map at top) and records the Q1–Q8 rulings. There is no new Spec,
  scope, framework or test run. B5, the fourteen ACs, G021 and every gate are unchanged.
