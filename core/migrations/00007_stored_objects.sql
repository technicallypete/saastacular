-- Tenant-aware object metadata (G009; docs/foundation/OBJECT-STORAGE.md).
--
-- One additive relation, app.stored_objects: a server-generated UUID identity, a required Organization, an optional
-- Group of that Organization (composite foreign key), an immutable creator, a database-generated provider key bound to
-- the Organization and the UUID, and a four-state lifecycle (pending, ready, deleting, deleted). Rows are never
-- hard-deleted by the runtime role: a deleted row is a tombstone that keeps its attribution.
--
-- FORCE row-level security with three permissive, command-specific policies on the 00006 pattern (TO the runtime role
-- and the owner, so FORCE binds the owner to the same contexts), built on the 00006 parsers and the two policy readers.
-- An owner-owned SECURITY INVOKER trigger enforces identity immutability and the closed lifecycle; it sets the
-- timestamps and deleted_by itself. Runtime receives SELECT, INSERT on six columns and UPDATE(state, sha256) only. No
-- role, membership, SECURITY DEFINER function, LEAKPROOF function or helper grant is added.
--
-- Additive and catalog-only: CREATE TABLE with its foreign keys takes brief SHARE ROW EXCLUSIVE locks on
-- organizations, groups and users; no rewrite, scan or data change. Down (development and test only; the runner
-- refuses down/reapply with APP_ENV=production) refuses while any row exists, because dropping the metadata would
-- orphan provider bytes. goose runs this file in one transaction.

-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow regrole; rt regrole; pr regrole; ar regrole;
  r record;
  actual text;
  expected text;
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regrole(db || '_runtime') IS NULL OR to_regrole(db || '_policy_reader') IS NULL OR to_regrole(db || '_auth_reader') IS NULL THEN
    RAISE EXCEPTION 'roles for database % are missing: run infra/postgres/provision-roles.sql (or a fresh 10-roles.sh) first', db;
  END IF;
  ow := current_user::regrole; rt := (db || '_runtime')::regrole;
  pr := (db || '_policy_reader')::regrole; ar := (db || '_auth_reader')::regrole;

  -- The 00006 role checks, reused: runtime and both helpers can never bypass row-level security or administer
  -- anything; the helpers cannot log in or inherit; each helper's only usable member is the owner; runtime is a member
  -- of nothing.
  FOR r IN SELECT rolname, rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolinherit, rolcanlogin, rolreplication
           FROM pg_roles WHERE oid IN (rt, pr, ar) LOOP
    IF r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication THEN
      RAISE EXCEPTION 'role % has an unsafe attribute: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
    IF r.rolname <> rt::text AND (r.rolcanlogin OR r.rolinherit) THEN
      RAISE EXCEPTION 'helper role % must be NOLOGIN NOINHERIT: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
  END LOOP;
  IF (SELECT count(*) FROM pg_auth_members m WHERE m.roleid IN (pr, ar) AND m.member = ow
        AND NOT m.inherit_option AND m.set_option AND NOT m.admin_option) <> 2 THEN
    RAISE EXCEPTION 'the owner must be a member of both helper roles WITH INHERIT FALSE, SET TRUE: run infra/postgres/provision-roles.sql';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.roleid IN (pr, ar) AND m.member <> ow
               AND (m.inherit_option OR m.set_option OR m.member::regrole::text LIKE db || '\_%')) THEN
    RAISE EXCEPTION 'a helper role has an unexpected member: run infra/postgres/provision-roles.sql';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member = rt) THEN
    RAISE EXCEPTION 'the runtime role must not be a member of any role';
  END IF;
  IF to_regclass('app.stored_objects') IS NOT NULL THEN
    RAISE EXCEPTION 'version-6 precondition: app.stored_objects already exists';
  END IF;

  -- The exact version-6 inventories of 00006 postconditions 1, 2 and 4.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  expected := 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,users+force';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-6 precondition: row-level security tables are %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s', p.polname, c.relname, p.polcmd, p.polpermissive::text,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt THEN 'rt' WHEN ow THEN 'ow' WHEN pr THEN 'pr'
                            WHEN ar THEN 'ar' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid;
  expected := 'business_profiles_select:business_profiles:r:true:ow+rt;business_profiles_update:business_profiles:w:true:ow+rt;'
    || 'group_business_profiles_select:group_business_profiles:r:true:ow+rt;groups_policy_reader:groups:r:true:pr;'
    || 'groups_select:groups:r:true:ow+rt;groups_update:groups:w:true:ow+rt;'
    || 'identity_principals_auth_reader:identity_principals:r:true:ar;'
    || 'identity_principals_insert:identity_principals:a:true:ow+rt;identity_principals_select:identity_principals:r:true:ow+rt;'
    || 'login_transactions_delete:login_transactions:d:true:ow+rt;login_transactions_expire:login_transactions:d:true:ow+rt;'
    || 'login_transactions_insert:login_transactions:a:true:ow+rt;login_transactions_select:login_transactions:r:true:ow+rt;'
    || 'organizations_select:organizations:r:true:ow+rt;organizations_update:organizations:w:true:ow+rt;'
    || 'role_bindings_lock:role_bindings:w:true:ow+rt;role_bindings_policy_reader:role_bindings:r:true:pr;'
    || 'role_bindings_select:role_bindings:r:true:ow+rt;sessions_auth_reader:sessions:r:true:ar;'
    || 'sessions_insert:sessions:a:true:ow+rt;sessions_select:sessions:r:true:ow+rt;sessions_update:sessions:w:true:ow+rt;'
    || 'users_auth_reader:users:r:true:ar;users_insert:users:a:true:ow+rt;users_select:users:r:true:ow+rt;'
    || 'users_update:users:w:true:ow+rt';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-6 precondition: policy inventory is %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ow THEN 'ow' WHEN pr THEN 'pr' WHEN ar THEN 'ar' ELSE 'other' END,
           p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN ow THEN 'ow'
                                             WHEN pr THEN 'pr' WHEN ar THEN 'ar' ELSE 'other' END
                              || '=' || a.privilege_type AS v
                       FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-')), ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace;
  expected := 'auth_logout_ref:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;'
    || 'auth_principal_subject:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;'
    || 'auth_resolve_session:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;'
    || 'business_profiles_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'ctx_group:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;'
    || 'ctx_login_binding:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;'
    || 'ctx_org:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;'
    || 'ctx_user:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;'
    || 'group_business_profiles_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'groups_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'organizations_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'rls_group_role:pr:true:false:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_org_role:pr:true:false:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'role_bindings_keys_immutable:ow:false:false:v:search_path=pg_catalog:-';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-6 precondition: function allowlist is %, expected %', actual, expected;
  END IF;
END
$$;
-- +goose StatementEnd

-- The relation. The provider key is generated from the Organization and the UUID; neither the runtime role, the owner
-- nor a caller can write it, and two rows can never share one (id is unique).
CREATE TABLE app.stored_objects (
  id              uuid PRIMARY KEY CONSTRAINT stored_objects_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  organization_id uuid NOT NULL,
  group_id        uuid,
  creator_user_id uuid NOT NULL,
  storage_key     text GENERATED ALWAYS AS ('objects/' || organization_id::text || '/' || id::text) STORED,
  state           text NOT NULL DEFAULT 'pending'
                  CONSTRAINT stored_objects_state_valid CHECK (state IN ('pending', 'ready', 'deleting', 'deleted')),
  byte_size       bigint NOT NULL CONSTRAINT stored_objects_size_valid CHECK (byte_size BETWEEN 0 AND 1073741824),
  content_type    text NOT NULL
                  CONSTRAINT stored_objects_content_type_valid
                  CHECK (content_type ~ '^[a-z0-9][a-z0-9!#$&^_.+-]{0,62}/[a-z0-9][a-z0-9!#$&^_.+-]{0,62}$'),
  sha256          bytea CONSTRAINT stored_objects_sha256_valid CHECK (sha256 IS NULL OR octet_length(sha256) = 32),
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

-- The lifecycle guard (owner-owned SECURITY INVOKER). Identity columns are immutable; only the creator finalizes
-- (pending -> ready, with a digest); pending or ready -> deleting records the actor; deleting -> deleted is the
-- tombstone. Same-state updates, any update of a deleted row and every other transition are refused. The parser
-- raises on a malformed, zero or illegal context.
-- +goose StatementBegin
CREATE FUNCTION app.stored_objects_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  actor uuid := app.ctx_user();
BEGIN
  IF (NEW.id, NEW.organization_id, NEW.group_id, NEW.creator_user_id, NEW.byte_size, NEW.content_type, NEW.created_at)
     IS DISTINCT FROM (OLD.id, OLD.organization_id, OLD.group_id, OLD.creator_user_id, OLD.byte_size, OLD.content_type, OLD.created_at)
  THEN
    RAISE EXCEPTION 'stored object identity is immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'stored_objects_immutable';
  END IF;
  IF actor IS NULL THEN
    RAISE EXCEPTION 'stored object update needs an actor' USING ERRCODE = 'insufficient_privilege';
  END IF;
  NEW.ready_at := OLD.ready_at; NEW.deleting_at := OLD.deleting_at; NEW.deleted_at := OLD.deleted_at; NEW.deleted_by := OLD.deleted_by;
  IF OLD.state = 'pending' AND NEW.state = 'ready' AND OLD.creator_user_id = actor AND NEW.sha256 IS NOT NULL THEN
    NEW.ready_at := now();
  ELSIF OLD.state IN ('pending', 'ready') AND NEW.state = 'deleting' AND NEW.sha256 IS NOT DISTINCT FROM OLD.sha256 THEN
    NEW.deleting_at := now(); NEW.deleted_by := actor;
  ELSIF OLD.state = 'deleting' AND NEW.state = 'deleted' AND NEW.sha256 IS NOT DISTINCT FROM OLD.sha256 THEN
    NEW.deleted_at := now();
  ELSE
    RAISE EXCEPTION 'illegal stored object transition'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'stored_objects_transition';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd
CREATE TRIGGER stored_objects_guard BEFORE UPDATE ON app.stored_objects FOR EACH ROW EXECUTE FUNCTION app.stored_objects_guard();

-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow text := current_user;
  rt text := db || '_runtime';
  pr text := db || '_policy_reader';
  ar text := db || '_auth_reader';
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  MGR text;
  scope text;
  w text;
  actual text;
  expected text;
BEGIN
  MGR := R || ' IN (''org_owner'', ''org_admin'')';

  REVOKE ALL ON FUNCTION app.stored_objects_guard() FROM PUBLIC;

  -- Runtime grants: read, insert of the six caller-supplied columns, and the two lifecycle columns. The state default,
  -- created_at and the key come from the database; the trigger sets the timestamps and deleted_by.
  EXECUTE format('GRANT SELECT ON app.stored_objects TO %I', rt);
  EXECUTE format('GRANT INSERT (id, organization_id, group_id, creator_user_id, byte_size, content_type) ON app.stored_objects TO %I', rt);
  EXECUTE format('GRANT UPDATE (state, sha256) ON app.stored_objects TO %I', rt);

  ALTER TABLE app.stored_objects ENABLE ROW LEVEL SECURITY;
  ALTER TABLE app.stored_objects FORCE ROW LEVEL SECURITY;

  -- Read scope: an Organization object is readable with any current Organization role in a valid context; a Group
  -- object with the selected Group (when one is selected) and an Organization owner/admin role or a current role in
  -- that exact Group.
  scope := format('organization_id = %s AND %s IS NOT NULL AND (group_id IS NULL OR ((%s IS NULL OR group_id = %s) AND '
    || '(%s OR app.rls_group_role(group_id) IS NOT NULL)))', O, R, G, G, MGR);
  EXECUTE format('CREATE POLICY stored_objects_select ON app.stored_objects FOR SELECT TO %I, %I USING (%s)', rt, ow, scope);
  -- Upload scope: the subject as creator of a pending row; an Organization object needs Organization owner/admin, a
  -- Group object any current role in that exact Group (or Organization owner/admin).
  EXECUTE format('CREATE POLICY stored_objects_insert ON app.stored_objects FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND creator_user_id = %s AND state = ''pending'' AND ('
    || '(group_id IS NULL AND %s) OR (group_id IS NOT NULL AND (%s IS NULL OR group_id = %s) AND '
    || '(%s OR app.rls_group_role(group_id) IS NOT NULL))))', rt, ow, O, R, U, MGR, G, G, MGR);
  -- Lifecycle scope: Organization objects by Organization owner/admin only; Group objects by Organization owner/admin,
  -- the Group's owner/admin, or the creator while holding a current role in that exact Group. No state predicate: a
  -- deleted row stays lockable for an idempotent repeat delete (the trigger forbids changing it).
  w := format('organization_id = %s AND %s IS NOT NULL AND ((group_id IS NULL AND %s) OR (group_id IS NOT NULL AND '
    || '(%s IS NULL OR group_id = %s) AND (%s OR app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'') OR '
    || '(creator_user_id = %s AND app.rls_group_role(group_id) IS NOT NULL))))', O, R, MGR, G, G, MGR, U);
  EXECUTE format('CREATE POLICY stored_objects_update ON app.stored_objects FOR UPDATE TO %I, %I USING (%s) WITH CHECK (%s)',
    rt, ow, w, w);

  -- Postcondition 1: the version-6 tables plus stored_objects, all forced.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  expected := 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'row-level security tables are %, expected %', actual, expected;
  END IF;

  -- Postcondition 2: the policy inventory (name, table, command, permissive, roles), with symbolic role labels.
  SELECT string_agg(format('%s:%s:%s:%s:%s', p.polname, c.relname, p.polcmd, p.polpermissive::text,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt::regrole THEN 'rt' WHEN ow::regrole THEN 'ow' WHEN pr::regrole THEN 'pr'
                            WHEN ar::regrole THEN 'ar' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid;
  expected := 'business_profiles_select:business_profiles:r:true:ow+rt;business_profiles_update:business_profiles:w:true:ow+rt;'
    || 'group_business_profiles_select:group_business_profiles:r:true:ow+rt;groups_policy_reader:groups:r:true:pr;'
    || 'groups_select:groups:r:true:ow+rt;groups_update:groups:w:true:ow+rt;'
    || 'identity_principals_auth_reader:identity_principals:r:true:ar;'
    || 'identity_principals_insert:identity_principals:a:true:ow+rt;identity_principals_select:identity_principals:r:true:ow+rt;'
    || 'login_transactions_delete:login_transactions:d:true:ow+rt;login_transactions_expire:login_transactions:d:true:ow+rt;'
    || 'login_transactions_insert:login_transactions:a:true:ow+rt;login_transactions_select:login_transactions:r:true:ow+rt;'
    || 'organizations_select:organizations:r:true:ow+rt;organizations_update:organizations:w:true:ow+rt;'
    || 'role_bindings_lock:role_bindings:w:true:ow+rt;role_bindings_policy_reader:role_bindings:r:true:pr;'
    || 'role_bindings_select:role_bindings:r:true:ow+rt;sessions_auth_reader:sessions:r:true:ar;'
    || 'sessions_insert:sessions:a:true:ow+rt;sessions_select:sessions:r:true:ow+rt;sessions_update:sessions:w:true:ow+rt;'
    || 'stored_objects_insert:stored_objects:a:true:ow+rt;stored_objects_select:stored_objects:r:true:ow+rt;'
    || 'stored_objects_update:stored_objects:w:true:ow+rt;'
    || 'users_auth_reader:users:r:true:ar;users_insert:users:a:true:ow+rt;users_select:users:r:true:ow+rt;'
    || 'users_update:users:w:true:ow+rt';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'policy inventory is %, expected %', actual, expected;
  END IF;

  -- Postcondition 3: every table/view/sequence and column ACL entry of schema app other than the owner's own, for every
  -- grantee including PUBLIC, with symbolic labels: version 6 plus the stored_objects runtime grants.
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' WHEN pr::regrole THEN 'pr'
                                          WHEN ar::regrole THEN 'ar' ELSE a.grantee::regrole::text END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname || '.' || at.attname, CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' WHEN pr::regrole THEN 'pr'
                                          WHEN ar::regrole THEN 'ar' ELSE a.grantee::regrole::text END,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
    FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.relnamespace = 'app'::regnamespace GROUP BY c.relname, at.attname, a.grantee) t;
  expected := 'business_profile_access_paths:rt:SELECT;business_profiles:rt:SELECT;business_profiles.description:rt:UPDATE;'
    || 'business_profiles.display_name:rt:UPDATE;business_profiles.industry:rt:UPDATE;business_profiles.legal_name:rt:UPDATE;'
    || 'business_profiles.website:rt:UPDATE;group_business_profiles:rt:SELECT;group_memberships:rt:SELECT;groups:rt:SELECT;'
    || 'groups.id:pr:SELECT;groups.name:rt:UPDATE;groups.organization_id:pr:SELECT;groups.slug:rt:UPDATE;'
    || 'identity_principals:rt:INSERT,SELECT;identity_principals.issuer:ar:SELECT;identity_principals.subject:ar:SELECT;'
    || 'identity_principals.user_id:ar:SELECT;identity_provider_binding:rt:INSERT,SELECT;login_transactions:rt:DELETE,INSERT,SELECT;'
    || 'organization_memberships:rt:SELECT;organizations:rt:SELECT;organizations.group_admins_edit_business_profiles:rt:UPDATE;'
    || 'role_bindings.group_id:pr:SELECT;role_bindings.group_id:rt:SELECT;role_bindings.organization_id:pr:SELECT;'
    || 'role_bindings.organization_id:rt:SELECT;role_bindings.role:pr:SELECT;role_bindings.role:rt:SELECT,UPDATE;'
    || 'role_bindings.scope_id:pr:SELECT;role_bindings.scope_id:rt:SELECT;role_bindings.scope_type:pr:SELECT;'
    || 'role_bindings.scope_type:rt:SELECT;role_bindings.source:pr:SELECT;role_bindings.source:rt:SELECT;'
    || 'role_bindings.subject_user_id:pr:SELECT;role_bindings.subject_user_id:rt:SELECT;sessions:rt:INSERT,SELECT;'
    || 'sessions.absolute_expires_at:ar:SELECT;sessions.auth_time:ar:SELECT;sessions.credential_hash:ar:SELECT;'
    || 'sessions.csrf_token:ar:SELECT;sessions.id:ar:SELECT;sessions.idle_expires_at:ar:SELECT;sessions.idle_expires_at:rt:UPDATE;'
    || 'sessions.mfa_assurance:ar:SELECT;sessions.provider_logout_ref_expires_at:ar:SELECT;'
    || 'sessions.provider_logout_ref_expires_at:rt:UPDATE;sessions.provider_logout_ref_hash:ar:SELECT;'
    || 'sessions.provider_logout_ref_hash:rt:UPDATE;sessions.provider_logout_returned_at:ar:SELECT;'
    || 'sessions.provider_logout_returned_at:rt:UPDATE;sessions.provider_revocation:rt:UPDATE;'
    || 'sessions.provider_session_id:ar:SELECT;sessions.renewed_at:ar:SELECT;sessions.renewed_at:rt:UPDATE;'
    || 'sessions.revoke_reason:rt:UPDATE;sessions.revoked_at:ar:SELECT;sessions.revoked_at:rt:UPDATE;sessions.user_id:ar:SELECT;'
    || 'stored_objects:rt:SELECT;stored_objects.byte_size:rt:INSERT;stored_objects.content_type:rt:INSERT;'
    || 'stored_objects.creator_user_id:rt:INSERT;stored_objects.group_id:rt:INSERT;stored_objects.id:rt:INSERT;'
    || 'stored_objects.organization_id:rt:INSERT;stored_objects.sha256:rt:UPDATE;stored_objects.state:rt:UPDATE;'
    || 'users.created_at:rt:INSERT;users.disabled_at:rt:SELECT,UPDATE;users.display_name:ar:SELECT;'
    || 'users.display_name:rt:INSERT,SELECT,UPDATE;users.id:ar:SELECT;users.id:rt:INSERT,SELECT;users.status:ar:SELECT;'
    || 'users.status:rt:SELECT,UPDATE;users.updated_at:rt:INSERT,UPDATE';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-7 ACLs are %, expected %', actual, expected;
  END IF;

  -- Postcondition 4: the function allowlist: version 6 plus the invoker guard, with no grant.
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ow::regrole THEN 'ow' WHEN pr::regrole THEN 'pr' WHEN ar::regrole THEN 'ar' ELSE 'other' END,
           p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' WHEN ow::regrole THEN 'ow'
                                             WHEN pr::regrole THEN 'pr' WHEN ar::regrole THEN 'ar' ELSE 'other' END
                              || '=' || a.privilege_type AS v
                       FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-')), ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace;
  expected := 'auth_logout_ref:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;'
    || 'auth_principal_subject:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;'
    || 'auth_resolve_session:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;'
    || 'business_profiles_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'ctx_group:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;'
    || 'ctx_login_binding:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;'
    || 'ctx_org:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;'
    || 'ctx_user:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;'
    || 'group_business_profiles_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'groups_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'organizations_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'rls_group_role:pr:true:false:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_org_role:pr:true:false:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'role_bindings_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'stored_objects_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'function allowlist is %, expected %', actual, expected;
  END IF;

  -- Postcondition 5: schema app privileges are unchanged.
  SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") INTO actual FROM (
    SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' WHEN pr::regrole THEN 'pr'
                          WHEN ar::regrole THEN 'ar' WHEN (db || '_retention')::regrole THEN 'rn'
                          ELSE a.grantee::regrole::text END || '=' || a.privilege_type AS v
    FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
    WHERE n.nspname = 'app' AND a.grantee <> n.nspowner) l;
  expected := 'ar=USAGE,pr=USAGE,rn=USAGE,rt=USAGE';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'schema app privileges are %, expected %', actual, expected;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow text := current_user;
  rt text := db || '_runtime';
  pr text := db || '_policy_reader';
  ar text := db || '_auth_reader';
  actual text;
  expected text;
BEGIN
  -- Without FORCE the owner's count is not filtered by row-level security. Dropping metadata while rows exist would
  -- orphan provider bytes: refuse, and the refusal rolls back the NO FORCE with everything else.
  ALTER TABLE app.stored_objects NO FORCE ROW LEVEL SECURITY;
  IF EXISTS (SELECT 1 FROM app.stored_objects) THEN
    RAISE EXCEPTION 'refusing to roll back 00007: app.stored_objects holds rows (their provider bytes would be orphaned)';
  END IF;
  DROP TABLE app.stored_objects;
  DROP FUNCTION app.stored_objects_guard();

  -- The exact version-6 inventories.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  expected := 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,users+force';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'row-level security tables are %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s', p.polname, c.relname, p.polcmd, p.polpermissive::text,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt::regrole THEN 'rt' WHEN ow::regrole THEN 'ow' WHEN pr::regrole THEN 'pr'
                            WHEN ar::regrole THEN 'ar' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid;
  expected := 'business_profiles_select:business_profiles:r:true:ow+rt;business_profiles_update:business_profiles:w:true:ow+rt;'
    || 'group_business_profiles_select:group_business_profiles:r:true:ow+rt;groups_policy_reader:groups:r:true:pr;'
    || 'groups_select:groups:r:true:ow+rt;groups_update:groups:w:true:ow+rt;'
    || 'identity_principals_auth_reader:identity_principals:r:true:ar;'
    || 'identity_principals_insert:identity_principals:a:true:ow+rt;identity_principals_select:identity_principals:r:true:ow+rt;'
    || 'login_transactions_delete:login_transactions:d:true:ow+rt;login_transactions_expire:login_transactions:d:true:ow+rt;'
    || 'login_transactions_insert:login_transactions:a:true:ow+rt;login_transactions_select:login_transactions:r:true:ow+rt;'
    || 'organizations_select:organizations:r:true:ow+rt;organizations_update:organizations:w:true:ow+rt;'
    || 'role_bindings_lock:role_bindings:w:true:ow+rt;role_bindings_policy_reader:role_bindings:r:true:pr;'
    || 'role_bindings_select:role_bindings:r:true:ow+rt;sessions_auth_reader:sessions:r:true:ar;'
    || 'sessions_insert:sessions:a:true:ow+rt;sessions_select:sessions:r:true:ow+rt;sessions_update:sessions:w:true:ow+rt;'
    || 'users_auth_reader:users:r:true:ar;users_insert:users:a:true:ow+rt;users_select:users:r:true:ow+rt;'
    || 'users_update:users:w:true:ow+rt';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'policy inventory is %, expected %', actual, expected;
  END IF;
  SELECT count(*)::text INTO actual FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.relname IN ('stored_objects', 'stored_objects_pkey', 'stored_objects_organization_group_idx');
  IF actual <> '0' OR to_regprocedure('app.stored_objects_guard()') IS NOT NULL THEN
    RAISE EXCEPTION 'version-7 objects survived the rollback';
  END IF;
END
$$;
-- +goose StatementEnd
