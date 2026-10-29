-- Retention maintenance privileges (G013; docs/foundation/AUDIT.md "G013 contract", docs/foundation/RECOVERY.md).
--
-- The existing <db>_retention login (provisioned since G001, no consumer until now) gains exactly what bounded retention
-- needs, all limited by row-level security to rows its own transaction declares eligible:
--   - app.retention_eligible(organization, at, ids): owner-owned SECURITY INVOKER, STABLE. It parses three
--     transaction-local settings on every call and raises 22023 when any is absent or malformed: retention.scope
--     (organization:<uuid> or global), retention.cutoff (UTC timestamp strictly before transaction_timestamp()) and
--     retention.holds (none, or up to 256 comma-separated canonical UUIDs). A row is eligible exactly when its
--     Organization matches the scope (NULL for global), its time is before the cutoff and none of its ids is held.
--   - Four permissive, command-specific policies TO <db>_retention only: SELECT and DELETE of eligible audit.events rows
--     and of eligible deleted app.stored_objects tombstones.
--   - USAGE on schema audit, column SELECT on the columns its statements read, DELETE on both tables, EXECUTE on the
--     function. Nothing else: no INSERT, UPDATE, TRUNCATE, REFERENCES, TRIGGER, DDL, membership or BYPASSRLS.
--   - audit.events_refuse returns OLD only for a DELETE by <db>_retention of an eligible row; every other UPDATE, DELETE
--     and TRUNCATE by every role, the owner and superusers included, still raises append-only.
-- Runtime and the owner stay unable to delete audit history or tombstones. No table, column, index, data or role change.
--
-- CREATE POLICY takes a brief ACCESS EXCLUSIVE lock on audit.events and app.stored_objects: migrate in a quiescent
-- window (provision, migrate, then code). Down (development and test only; the runner refuses down/reapply with
-- APP_ENV=production) restores the exact version-8 catalog and is allowed with rows present (no data is involved).
-- goose runs this file in one transaction.

-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow regrole; rt regrole; pr regrole; ar regrole; rn regrole;
  r record;
  actual text;
  expected text;
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regrole(db || '_runtime') IS NULL OR to_regrole(db || '_policy_reader') IS NULL OR to_regrole(db || '_auth_reader') IS NULL
     OR to_regrole(db || '_retention') IS NULL THEN
    RAISE EXCEPTION 'roles for database % are missing: run infra/postgres/provision-roles.sql (or a fresh 10-roles.sh) first', db;
  END IF;
  ow := current_user::regrole; rt := (db || '_runtime')::regrole; rn := (db || '_retention')::regrole;
  pr := (db || '_policy_reader')::regrole; ar := (db || '_auth_reader')::regrole;

  -- The 00006-00008 role checks, reused, plus the retention login: LOGIN NOINHERIT, never able to bypass row-level
  -- security or administer anything, a member of no role and with no member.
  FOR r IN SELECT rolname, rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolinherit, rolcanlogin, rolreplication
           FROM pg_roles WHERE oid IN (rt, pr, ar, rn) LOOP
    IF r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication THEN
      RAISE EXCEPTION 'role % has an unsafe attribute: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
    IF r.rolname NOT IN (rt::text, rn::text) AND (r.rolcanlogin OR r.rolinherit) THEN
      RAISE EXCEPTION 'helper role % must be NOLOGIN NOINHERIT: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
    IF r.rolname = rn::text AND (NOT r.rolcanlogin OR r.rolinherit) THEN
      RAISE EXCEPTION 'retention role % must be LOGIN NOINHERIT: run infra/postgres/provision-roles.sql', r.rolname;
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
  IF EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member IN (rt, rn) OR m.roleid = rn) THEN
    RAISE EXCEPTION 'the runtime and retention roles must not be members of any role, and retention must have no member';
  END IF;

  -- The exact version-8 inventories (00008 preconditions and postconditions 1-4).
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  expected := 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-8 precondition: row-level security tables are %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s', p.polname, c.relname, p.polcmd, p.polpermissive::text,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt THEN 'rt' WHEN ow THEN 'ow' WHEN pr THEN 'pr'
                            WHEN ar THEN 'ar' WHEN rn THEN 'rn' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid;
  expected := 'business_profiles_select:business_profiles:r:true:ow+rt;business_profiles_update:business_profiles:w:true:ow+rt;'
    || 'events_insert:events:a:true:ow+rt;events_select:events:r:true:ow+rt;'
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
    RAISE EXCEPTION 'version-8 precondition: policy inventory is %, expected %', actual, expected;
  END IF;
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                          WHEN ar THEN 'ar' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname || '.' || at.attname, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                          WHEN ar THEN 'ar' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END,
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
    RAISE EXCEPTION 'version-8 precondition: app ACLs are %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ow THEN 'ow' WHEN pr THEN 'pr' WHEN ar THEN 'ar' ELSE 'other' END,
           p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN ow THEN 'ow'
                                             WHEN pr THEN 'pr' WHEN ar THEN 'ar' WHEN rn THEN 'rn' ELSE 'other' END
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
    RAISE EXCEPTION 'version-8 precondition: app function allowlist is %, expected %', actual, expected;
  END IF;
  SELECT string_agg(c.relname || ':' || c.relkind::text || ':' || CASE c.relowner WHEN ow THEN 'ow' ELSE 'other' END
           || CASE WHEN c.relrowsecurity THEN '+rls' ELSE '' END || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END,
           ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'audit'::regnamespace;
  IF actual IS DISTINCT FROM 'events:r:ow+rls+force,events_pkey:i:ow,events_tenant_time_idx:i:ow' THEN
    RAISE EXCEPTION 'version-8 precondition: audit relations are %', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s', t.tgname, t.tgtype, t.tgenabled, p.proname), ';' ORDER BY t.tgname) INTO actual
  FROM pg_trigger t JOIN pg_proc p ON p.oid = t.tgfoid WHERE t.tgrelid = 'audit.events'::regclass AND NOT t.tgisinternal;
  IF actual IS DISTINCT FROM 'events_attribute:7:O:events_attribute;events_refuse:27:O:events_refuse;'
                             || 'events_refuse_truncate:34:O:events_refuse' THEN
    RAISE EXCEPTION 'version-8 precondition: audit.events triggers are %', actual;
  END IF;
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'audit'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname || '.' || at.attname, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
    FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.relnamespace = 'audit'::regnamespace GROUP BY c.relname, at.attname, a.grantee) t;
  expected := 'events:rt:SELECT;events.action:rt:INSERT;events.group_id:rt:INSERT;events.operation:rt:INSERT;'
    || 'events.organization_id:rt:INSERT;events.outcome:rt:INSERT;events.previous_role:rt:INSERT;events.reason:rt:INSERT;'
    || 'events.resource_id:rt:INSERT;events.resource_type:rt:INSERT;events.role:rt:INSERT;events.subject_user_id:rt:INSERT;'
    || 'events.ttl_seconds:rt:INSERT';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-8 precondition: audit ACLs are %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ow THEN 'ow' ELSE 'other' END,
           p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE 'other' END
                              || '=' || a.privilege_type AS v
                       FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-')), ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.pronamespace = 'audit'::regnamespace;
  expected := 'capture_business_profiles:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_group_business_profiles:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_groups:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_organizations:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_role_bindings:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_session_revocations:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_sessions:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_stored_objects:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_users:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'events_attribute:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'events_refuse:ow:false:false:v:search_path=pg_catalog, pg_temp:-';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-8 precondition: audit function allowlist is %, expected %', actual, expected;
  END IF;
  IF (SELECT md5(p.prosrc) FROM pg_proc p WHERE p.oid = 'audit.events_refuse()'::regprocedure) IS DISTINCT FROM
     'ee436a151f6bcf4bf893747a8cf243f9' THEN
    RAISE EXCEPTION 'version-8 precondition: audit.events_refuse is not the reviewed version-8 body';
  END IF;
  SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") INTO actual FROM (
    SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END
           || '=' || a.privilege_type AS v
    FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
    WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) l;
  IF actual IS DISTINCT FROM 'rt=USAGE' THEN
    RAISE EXCEPTION 'version-8 precondition: schema audit privileges are %, expected rt=USAGE', actual;
  END IF;
END
$$;
-- +goose StatementEnd

-- The eligibility function: the single database-side definition of "declared eligible". No exception block (no
-- subtransaction per row); every malformed or absent setting raises 22023.
-- +goose StatementBegin
CREATE FUNCTION app.retention_eligible(p_organization uuid, p_at timestamptz, p_ids uuid[]) RETURNS boolean
  LANGUAGE plpgsql STABLE SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u constant text := '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}';
  s text := current_setting('retention.scope', true);
  c text := current_setting('retention.cutoff', true);
  h text := current_setting('retention.holds', true);
  cutoff timestamptz;
  holds uuid[];
BEGIN
  IF s IS NULL OR (s <> 'global' AND s !~ ('^organization:' || u || '$')) THEN
    RAISE EXCEPTION 'retention.scope is absent or malformed' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF c IS NULL OR c !~ '^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]{1,6})?Z$' THEN
    RAISE EXCEPTION 'retention.cutoff is absent or malformed' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  cutoff := c::timestamptz;
  IF cutoff >= transaction_timestamp() THEN
    RAISE EXCEPTION 'retention.cutoff must be before the transaction time' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF h IS NULL OR (h <> 'none' AND (h !~ ('^' || u || '(,' || u || ')*$') OR cardinality(string_to_array(h, ',')) > 256)) THEN
    RAISE EXCEPTION 'retention.holds is absent or malformed' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  holds := CASE WHEN h = 'none' THEN '{}'::uuid[] ELSE string_to_array(h, ',')::uuid[] END;
  RETURN coalesce(CASE WHEN s = 'global' THEN p_organization IS NULL ELSE p_organization = substr(s, 14)::uuid END, false)
    AND coalesce(p_at < cutoff, false)
    AND NOT coalesce(p_ids && holds, false);
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow regrole := current_user::regrole;
  rt regrole := (current_database() || '_runtime')::regrole;
  pr regrole := (current_database() || '_policy_reader')::regrole;
  ar regrole := (current_database() || '_auth_reader')::regrole;
  rn regrole := (current_database() || '_retention')::regrole;
  actual text;
  expected text;
BEGIN
  REVOKE ALL ON FUNCTION app.retention_eligible(uuid, timestamptz, uuid[]) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.retention_eligible(uuid, timestamptz, uuid[]) TO %I', rn::text);

  -- Grants: schema usage, the minimal columns and DELETE (each limited by the policies below).
  EXECUTE format('GRANT USAGE ON SCHEMA audit TO %I', rn::text);
  EXECUTE format('GRANT SELECT (id, occurred_at, organization_id), DELETE ON audit.events TO %I', rn::text);
  EXECUTE format('GRANT SELECT (id, organization_id, state, deleted_at), DELETE ON app.stored_objects TO %I', rn::text);

  -- Policies for the retention login only (permissive, command-specific).
  EXECUTE format('CREATE POLICY events_retention_select ON audit.events FOR SELECT TO %I USING ('
    || 'app.retention_eligible(organization_id, occurred_at, ARRAY[organization_id, group_id, actor_user_id, subject_user_id, resource_id]))', rn::text);
  EXECUTE format('CREATE POLICY events_retention_delete ON audit.events FOR DELETE TO %I USING ('
    || 'app.retention_eligible(organization_id, occurred_at, ARRAY[organization_id, group_id, actor_user_id, subject_user_id, resource_id]))', rn::text);
  EXECUTE format('CREATE POLICY stored_objects_retention_select ON app.stored_objects FOR SELECT TO %I USING ('
    || 'state = ''deleted'' AND app.retention_eligible(organization_id, deleted_at, ARRAY[organization_id, group_id, id, creator_user_id, deleted_by]))', rn::text);
  EXECUTE format('CREATE POLICY stored_objects_retention_delete ON app.stored_objects FOR DELETE TO %I USING ('
    || 'state = ''deleted'' AND app.retention_eligible(organization_id, deleted_at, ARRAY[organization_id, group_id, id, creator_user_id, deleted_by]))', rn::text);

  -- The append-only refusal, with one exception: a DELETE by the retention login of a row its declaration makes
  -- eligible. Same name, owner, ACL and triggers (CREATE OR REPLACE keeps the function's identity).
  EXECUTE format($f$CREATE OR REPLACE FUNCTION audit.events_refuse() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $b$
BEGIN
  IF TG_OP = 'DELETE' THEN
    IF current_user = %L THEN
      IF app.retention_eligible(OLD.organization_id, OLD.occurred_at,
           ARRAY[OLD.organization_id, OLD.group_id, OLD.actor_user_id, OLD.subject_user_id, OLD.resource_id]) THEN
        RETURN OLD;
      END IF;
    END IF;
  END IF;
  RAISE EXCEPTION 'audit events are append-only' USING ERRCODE = 'insufficient_privilege';
END $b$ $f$, rn::text);

  -- Postconditions: the exact version-9 inventories.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  IF actual IS DISTINCT FROM 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force' THEN
    RAISE EXCEPTION 'app row-level security tables are %', actual;
  END IF;
  -- The 38 application relations plus the migration ledger's table, identity sequence and primary key: no relation added.
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace;
  IF actual <> '41' THEN
    RAISE EXCEPTION 'app relation count is %, expected 41 (38 plus the migration ledger)', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s', p.polname, c.relname, p.polcmd, p.polpermissive::text,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt THEN 'rt' WHEN ow THEN 'ow' WHEN pr THEN 'pr'
                            WHEN ar THEN 'ar' WHEN rn THEN 'rn' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid;
  expected := 'business_profiles_select:business_profiles:r:true:ow+rt;business_profiles_update:business_profiles:w:true:ow+rt;'
    || 'events_insert:events:a:true:ow+rt;events_retention_delete:events:d:true:rn;events_retention_select:events:r:true:rn;'
    || 'events_select:events:r:true:ow+rt;'
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
    || 'stored_objects_insert:stored_objects:a:true:ow+rt;stored_objects_retention_delete:stored_objects:d:true:rn;'
    || 'stored_objects_retention_select:stored_objects:r:true:rn;stored_objects_select:stored_objects:r:true:ow+rt;'
    || 'stored_objects_update:stored_objects:w:true:ow+rt;'
    || 'users_auth_reader:users:r:true:ar;users_insert:users:a:true:ow+rt;users_select:users:r:true:ow+rt;'
    || 'users_update:users:w:true:ow+rt';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'policy inventory is %, expected %', actual, expected;
  END IF;
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                          WHEN ar THEN 'ar' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname || '.' || at.attname, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                          WHEN ar THEN 'ar' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END,
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
    || 'stored_objects:rn:DELETE;stored_objects:rt:SELECT;stored_objects.byte_size:rt:INSERT;stored_objects.content_type:rt:INSERT;'
    || 'stored_objects.creator_user_id:rt:INSERT;stored_objects.deleted_at:rn:SELECT;stored_objects.group_id:rt:INSERT;'
    || 'stored_objects.id:rn:SELECT;stored_objects.id:rt:INSERT;stored_objects.organization_id:rn:SELECT;'
    || 'stored_objects.organization_id:rt:INSERT;stored_objects.sha256:rt:UPDATE;stored_objects.state:rn:SELECT;'
    || 'stored_objects.state:rt:UPDATE;'
    || 'users.created_at:rt:INSERT;users.disabled_at:rt:SELECT,UPDATE;users.display_name:ar:SELECT;'
    || 'users.display_name:rt:INSERT,SELECT,UPDATE;users.id:ar:SELECT;users.id:rt:INSERT,SELECT;users.status:ar:SELECT;'
    || 'users.status:rt:SELECT,UPDATE;users.updated_at:rt:INSERT,UPDATE';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'app ACLs are %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ow THEN 'ow' WHEN pr THEN 'pr' WHEN ar THEN 'ar' ELSE 'other' END,
           p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN ow THEN 'ow'
                                             WHEN pr THEN 'pr' WHEN ar THEN 'ar' WHEN rn THEN 'rn' ELSE 'other' END
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
    || 'retention_eligible:ow:false:false:s:search_path=pg_catalog, pg_temp:rn=EXECUTE;'
    || 'rls_group_role:pr:true:false:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_org_role:pr:true:false:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'role_bindings_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'stored_objects_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'app function allowlist is %, expected %', actual, expected;
  END IF;
  SELECT string_agg(c.relname || ':' || c.relkind::text || ':' || CASE c.relowner WHEN ow THEN 'ow' ELSE 'other' END
           || CASE WHEN c.relrowsecurity THEN '+rls' ELSE '' END || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END,
           ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'audit'::regnamespace;
  IF actual IS DISTINCT FROM 'events:r:ow+rls+force,events_pkey:i:ow,events_tenant_time_idx:i:ow' THEN
    RAISE EXCEPTION 'audit relations are %', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s', t.tgname, t.tgtype, t.tgenabled, p.proname), ';' ORDER BY t.tgname) INTO actual
  FROM pg_trigger t JOIN pg_proc p ON p.oid = t.tgfoid WHERE t.tgrelid = 'audit.events'::regclass AND NOT t.tgisinternal;
  IF actual IS DISTINCT FROM 'events_attribute:7:O:events_attribute;events_refuse:27:O:events_refuse;'
                             || 'events_refuse_truncate:34:O:events_refuse' THEN
    RAISE EXCEPTION 'audit.events triggers are %', actual;
  END IF;
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'audit'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname || '.' || at.attname, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
    FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.relnamespace = 'audit'::regnamespace GROUP BY c.relname, at.attname, a.grantee) t;
  expected := 'events:rn:DELETE;events:rt:SELECT;events.action:rt:INSERT;events.group_id:rt:INSERT;events.id:rn:SELECT;'
    || 'events.occurred_at:rn:SELECT;events.operation:rt:INSERT;events.organization_id:rn:SELECT;'
    || 'events.organization_id:rt:INSERT;events.outcome:rt:INSERT;events.previous_role:rt:INSERT;events.reason:rt:INSERT;'
    || 'events.resource_id:rt:INSERT;events.resource_type:rt:INSERT;events.role:rt:INSERT;events.subject_user_id:rt:INSERT;'
    || 'events.ttl_seconds:rt:INSERT';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'audit ACLs are %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ow THEN 'ow' ELSE 'other' END,
           p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE 'other' END
                              || '=' || a.privilege_type AS v
                       FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-')), ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.pronamespace = 'audit'::regnamespace;
  expected := 'capture_business_profiles:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_group_business_profiles:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_groups:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_organizations:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_role_bindings:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_session_revocations:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_sessions:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_stored_objects:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'capture_users:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'events_attribute:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'events_refuse:ow:false:false:v:search_path=pg_catalog, pg_temp:-';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'audit function allowlist is %, expected %', actual, expected;
  END IF;
  -- The two reviewed bodies (the refusal body with the embedded role name normalized to <db>_retention).
  SELECT md5(replace(p.prosrc, rn::text, '<db>_retention')) INTO actual FROM pg_proc p WHERE p.oid = 'audit.events_refuse()'::regprocedure;
  IF actual IS DISTINCT FROM 'c6c0b51ae2876ba67658dc4041eec5e9' THEN
    RAISE EXCEPTION 'audit.events_refuse body digest is %', actual;
  END IF;
  SELECT md5(p.prosrc) INTO actual FROM pg_proc p WHERE p.oid = 'app.retention_eligible(uuid, timestamptz, uuid[])'::regprocedure;
  IF actual IS DISTINCT FROM '89c969a9ec364eae864aab3e9fda0d58' THEN
    RAISE EXCEPTION 'app.retention_eligible body digest is %', actual;
  END IF;
  SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") INTO actual FROM (
    SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END
           || '=' || a.privilege_type AS v
    FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
    WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) l;
  IF actual IS DISTINCT FROM 'rn=USAGE,rt=USAGE' THEN
    RAISE EXCEPTION 'schema audit privileges are %, expected rn=USAGE,rt=USAGE', actual;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
  rn regrole := (current_database() || '_retention')::regrole;
  rt regrole := (current_database() || '_runtime')::regrole;
  actual text;
BEGIN
  DROP POLICY events_retention_select ON audit.events;
  DROP POLICY events_retention_delete ON audit.events;
  DROP POLICY stored_objects_retention_select ON app.stored_objects;
  DROP POLICY stored_objects_retention_delete ON app.stored_objects;
  -- The exact version-8 refusal body (00008).
  CREATE OR REPLACE FUNCTION audit.events_refuse() RETURNS trigger
    LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $b$
BEGIN
  RAISE EXCEPTION 'audit events are append-only' USING ERRCODE = 'insufficient_privilege';
END $b$;
  DROP FUNCTION app.retention_eligible(uuid, timestamptz, uuid[]);
  EXECUTE format('REVOKE SELECT (id, occurred_at, organization_id), DELETE ON audit.events FROM %I', rn::text);
  EXECUTE format('REVOKE SELECT (id, organization_id, state, deleted_at), DELETE ON app.stored_objects FROM %I', rn::text);
  EXECUTE format('REVOKE USAGE ON SCHEMA audit FROM %I', rn::text);

  -- The exact version-8 state: no retention policy or privilege in either schema, the version-8 refusal body.
  IF EXISTS (SELECT 1 FROM pg_policy p WHERE rn = ANY (p.polroles))
     OR EXISTS (SELECT 1 FROM pg_proc p WHERE p.proname = 'retention_eligible')
     OR EXISTS (SELECT 1 FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
                WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND a.grantee = rn)
     OR EXISTS (SELECT 1 FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid CROSS JOIN LATERAL aclexplode(at.attacl) a
                WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND a.grantee = rn) THEN
    RAISE EXCEPTION 'version-9 retention objects or grants survived the rollback';
  END IF;
  SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") INTO actual FROM (
    SELECT CASE a.grantee WHEN rt THEN 'rt' ELSE a.grantee::regrole::text END || '=' || a.privilege_type AS v
    FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) l;
  IF actual IS DISTINCT FROM 'rt=USAGE' THEN
    RAISE EXCEPTION 'schema audit privileges after the rollback are %', actual;
  END IF;
  IF (SELECT md5(p.prosrc) FROM pg_proc p WHERE p.oid = 'audit.events_refuse()'::regprocedure) IS DISTINCT FROM
     'ee436a151f6bcf4bf893747a8cf243f9' THEN
    RAISE EXCEPTION 'audit.events_refuse is not the version-8 body after the rollback';
  END IF;
END
$$;
-- +goose StatementEnd
