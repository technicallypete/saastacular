-- Tenant-bound MCP connections (G014; ADR-005; contracts/mcp/control-plane.md; docs/foundation/MCP.md).
--
-- Additive, one goose transaction:
--   - app.sessions.kind ('browser' | 'mcp', default 'browser'). An MCP connection owns exactly one 'mcp' session row
--     that carries its liveness (idle and absolute expiry), revocation and G013 restore invalidation. It has no issued
--     bearer: its credential hash is the hash of a discarded random value and its CSRF value is never returned.
--     sessions_insert keeps browser rows exactly as before (IDENTITY shape) and admits 'mcp' rows only in the TENANT-ORG
--     shape with a current Organization role (app.rls_org_role() IS NOT NULL). The E1 reader
--     (app.auth_resolve_session) resolves 'browser' sessions only; its helper gains SELECT on the kind column.
--   - app.mcp_connections: one row per connection (core-minted UUID, immutable user, Organization, host and resource,
--     pending -> active once, the client bound at activation, the optional selected Group and its context version).
--     FORCE row-level security: SELECT of the context user's own rows in the context Organization (no role needed, so
--     exact-owner termination still works after membership loss); INSERT of a pending row and UPDATE only with a
--     current Organization role. A SECURITY INVOKER guard trigger keeps the binding immutable, allows exactly one
--     pending -> active transition before the pending expiry and requires every selection change to advance the
--     context version, to run in the Organization-only context, and to name a Group (and leave a previous selection)
--     visible to the subject under row-level security; a clear needs nothing more. No DELETE path, no definer
--     reader, no token, hash or refresh value, no foreign key on the selected Group (rechecked per call).
--   - Audit (G011): two trigger-owned actions, mcp.connection_established (pending -> active) and mcp.context_changed
--     (selected Group changed), resource type mcp_connection, access.denied operation mcp.request with reason
--     resource_mismatch; audit.events_attribute is replaced with the matching attribution and shape rules and two
--     capture triggers record the actions in the same transaction (strict). Termination keeps using session.revoked.
--
-- Locks: ALTER TABLE takes a brief ACCESS EXCLUSIVE lock on app.sessions (column with a constant default: no rewrite)
-- and on audit.events (the four widened catalog CHECK constraints are validated by one scan). Migrate in a quiescent
-- window: provision, migrate, then code. G013's restore invalidation needs no change: it already expires every
-- unrevoked session regardless of kind.
--
-- Down (development and test only; the runner refuses down/reapply with APP_ENV=production) refuses while any MCP
-- connection, 'mcp' session or MCP audit event exists, then restores the exact version-9 catalog.

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

  IF to_regclass('app.mcp_connections') IS NOT NULL
     OR EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.sessions'::regclass AND attname = 'kind' AND NOT attisdropped) THEN
    RAISE EXCEPTION 'version-9 precondition: MCP connection objects already exist';
  END IF;

  -- The exact version-9 inventories (00009 postconditions, and the unchanged 00008 app trigger inventory).
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  IF actual IS DISTINCT FROM 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force' THEN
    RAISE EXCEPTION 'version-9 precondition: app row-level security tables are %', actual;
  END IF;
  -- The 38 application relations plus the migration ledger's table, identity sequence and primary key: no relation added.
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace;
  IF actual <> '41' THEN
    RAISE EXCEPTION 'version-9 precondition: app relation count is %, expected 41 (38 plus the migration ledger)', actual;
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
    RAISE EXCEPTION 'version-9 precondition: policy inventory is %, expected %', actual, expected;
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
    RAISE EXCEPTION 'version-9 precondition: app ACLs are %, expected %', actual, expected;
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
    RAISE EXCEPTION 'version-9 precondition: app function allowlist is %, expected %', actual, expected;
  END IF;
  SELECT string_agg(c.relname || ':' || c.relkind::text || ':' || CASE c.relowner WHEN ow THEN 'ow' ELSE 'other' END
           || CASE WHEN c.relrowsecurity THEN '+rls' ELSE '' END || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END,
           ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'audit'::regnamespace;
  IF actual IS DISTINCT FROM 'events:r:ow+rls+force,events_pkey:i:ow,events_tenant_time_idx:i:ow' THEN
    RAISE EXCEPTION 'version-9 precondition: audit relations are %', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s', t.tgname, t.tgtype, t.tgenabled, p.proname), ';' ORDER BY t.tgname) INTO actual
  FROM pg_trigger t JOIN pg_proc p ON p.oid = t.tgfoid WHERE t.tgrelid = 'audit.events'::regclass AND NOT t.tgisinternal;
  IF actual IS DISTINCT FROM 'events_attribute:7:O:events_attribute;events_refuse:27:O:events_refuse;'
                             || 'events_refuse_truncate:34:O:events_refuse' THEN
    RAISE EXCEPTION 'version-9 precondition: audit.events triggers are %', actual;
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
    RAISE EXCEPTION 'version-9 precondition: audit ACLs are %, expected %', actual, expected;
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
    RAISE EXCEPTION 'version-9 precondition: audit function allowlist is %, expected %', actual, expected;
  END IF;
  -- The two reviewed bodies (the refusal body with the embedded role name normalized to <db>_retention).
  SELECT md5(replace(p.prosrc, rn::text, '<db>_retention')) INTO actual FROM pg_proc p WHERE p.oid = 'audit.events_refuse()'::regprocedure;
  IF actual IS DISTINCT FROM 'c6c0b51ae2876ba67658dc4041eec5e9' THEN
    RAISE EXCEPTION 'version-9 precondition: audit.events_refuse body digest is %', actual;
  END IF;
  SELECT md5(p.prosrc) INTO actual FROM pg_proc p WHERE p.oid = 'app.retention_eligible(uuid, timestamptz, uuid[])'::regprocedure;
  IF actual IS DISTINCT FROM '89c969a9ec364eae864aab3e9fda0d58' THEN
    RAISE EXCEPTION 'version-9 precondition: app.retention_eligible body digest is %', actual;
  END IF;
  SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") INTO actual FROM (
    SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END
           || '=' || a.privilege_type AS v
    FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
    WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) l;
  IF actual IS DISTINCT FROM 'rn=USAGE,rt=USAGE' THEN
    RAISE EXCEPTION 'version-9 precondition: schema audit privileges are %, expected rn=USAGE,rt=USAGE', actual;
  END IF;
  SELECT string_agg(format('%s.%s:%s:%s:%s:%s', c.relname, t.tgname, t.tgtype, t.tgenabled,
           p.pronamespace::regnamespace::text || '.' || p.proname,
           CASE WHEN t.tgnewtable IS NOT NULL THEN 'tt' WHEN t.tgqual IS NOT NULL THEN 'when' ELSE '-' END),
           ';' ORDER BY c.relname, t.tgname) INTO actual
  FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid JOIN pg_proc p ON p.oid = t.tgfoid
  WHERE c.relnamespace = 'app'::regnamespace AND NOT t.tgisinternal;
  expected := 'business_profiles.audit_capture_insert:5:O:audit.capture_business_profiles:-;'
    || 'business_profiles.audit_capture_update:17:O:audit.capture_business_profiles:when;'
    || 'business_profiles.business_profiles_keys_immutable:19:O:app.business_profiles_keys_immutable:-;'
    || 'group_business_profiles.audit_capture_delete:9:O:audit.capture_group_business_profiles:-;'
    || 'group_business_profiles.audit_capture_insert:5:O:audit.capture_group_business_profiles:-;'
    || 'group_business_profiles.group_business_profiles_keys_immutable:19:O:app.group_business_profiles_keys_immutable:-;'
    || 'groups.audit_capture_insert:5:O:audit.capture_groups:-;'
    || 'groups.audit_capture_update:17:O:audit.capture_groups:when;'
    || 'groups.groups_keys_immutable:19:O:app.groups_keys_immutable:-;'
    || 'organizations.audit_capture_insert:5:O:audit.capture_organizations:-;'
    || 'organizations.audit_capture_update:17:O:audit.capture_organizations:when;'
    || 'organizations.organizations_keys_immutable:19:O:app.organizations_keys_immutable:-;'
    || 'role_bindings.audit_capture_delete:9:O:audit.capture_role_bindings:-;'
    || 'role_bindings.audit_capture_insert:5:O:audit.capture_role_bindings:-;'
    || 'role_bindings.audit_capture_update:17:O:audit.capture_role_bindings:when;'
    || 'role_bindings.role_bindings_keys_immutable:19:O:app.role_bindings_keys_immutable:-;'
    || 'sessions.audit_capture_insert:5:O:audit.capture_sessions:-;'
    || 'sessions.audit_capture_revocations:16:O:audit.capture_session_revocations:tt;'
    || 'sessions.audit_capture_update:17:O:audit.capture_sessions:when;'
    || 'stored_objects.audit_capture_insert:5:O:audit.capture_stored_objects:-;'
    || 'stored_objects.audit_capture_update:17:O:audit.capture_stored_objects:when;'
    || 'stored_objects.stored_objects_guard:19:O:app.stored_objects_guard:-;'
    || 'users.audit_capture_insert:5:O:audit.capture_users:-;'
    || 'users.audit_capture_update:17:O:audit.capture_users:when';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-9 precondition: app trigger inventory is %, expected %', actual, expected;
  END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE app.sessions ADD COLUMN kind text NOT NULL DEFAULT 'browser' CONSTRAINT sessions_kind_valid CHECK (kind IN ('browser', 'mcp'));

CREATE TABLE app.mcp_connections (
  id                 uuid        NOT NULL CONSTRAINT mcp_connections_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  session_id         uuid        NOT NULL,
  user_id            uuid        NOT NULL,
  organization_id    uuid        NOT NULL,
  host               text        NOT NULL CONSTRAINT mcp_connections_host_valid CHECK (char_length(host) BETWEEN 1 AND 253),
  resource           text        NOT NULL CONSTRAINT mcp_connections_resource_valid CHECK (char_length(resource) BETWEEN 1 AND 2048),
  state              text        NOT NULL DEFAULT 'pending' CONSTRAINT mcp_connections_state_valid CHECK (state IN ('pending', 'active')),
  pending_expires_at timestamptz NOT NULL,
  client_id          text        NULL CONSTRAINT mcp_connections_client_valid CHECK (char_length(client_id) BETWEEN 1 AND 2048),
  activated_at       timestamptz NULL,
  selected_group_id  uuid        NULL,
  context_version    bigint      NOT NULL DEFAULT 0 CONSTRAINT mcp_connections_version_valid CHECK (context_version >= 0),
  created_at         timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT mcp_connections_pkey PRIMARY KEY (id),
  CONSTRAINT mcp_connections_session_key UNIQUE (session_id),
  CONSTRAINT mcp_connections_resource_key UNIQUE (resource),
  CONSTRAINT mcp_connections_session_fkey FOREIGN KEY (session_id) REFERENCES app.sessions (id) ON DELETE RESTRICT,
  CONSTRAINT mcp_connections_user_fkey FOREIGN KEY (user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT mcp_connections_organization_fkey FOREIGN KEY (organization_id) REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT mcp_connections_active_shape CHECK ((state = 'active') = (client_id IS NOT NULL)
                                                 AND (client_id IS NULL) = (activated_at IS NULL)),
  CONSTRAINT mcp_connections_selection_shape CHECK (selected_group_id IS NULL OR state = 'active')
);

-- The guard (owner-owned SECURITY INVOKER). The binding columns never change; state moves pending -> active exactly
-- once, binding the client, before the pending expiry; a selection change advances the context version by one and
-- equals the context Group (NULL in Organization-only context). Anything else raises.
-- +goose StatementBegin
CREATE FUNCTION app.mcp_connections_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  g uuid := app.ctx_group();
BEGIN
  IF NEW.id IS DISTINCT FROM OLD.id OR NEW.session_id IS DISTINCT FROM OLD.session_id OR NEW.user_id IS DISTINCT FROM OLD.user_id
     OR NEW.organization_id IS DISTINCT FROM OLD.organization_id OR NEW.host IS DISTINCT FROM OLD.host
     OR NEW.resource IS DISTINCT FROM OLD.resource OR NEW.pending_expires_at IS DISTINCT FROM OLD.pending_expires_at
     OR NEW.created_at IS DISTINCT FROM OLD.created_at THEN
    RAISE EXCEPTION 'mcp connection binding is immutable' USING ERRCODE = 'check_violation', CONSTRAINT = 'mcp_connections_guard';
  END IF;
  IF NEW.state IS DISTINCT FROM OLD.state THEN
    IF NOT (OLD.state = 'pending' AND NEW.state = 'active' AND OLD.client_id IS NULL AND NEW.client_id IS NOT NULL
            AND OLD.activated_at IS NULL AND NEW.activated_at IS NOT NULL AND OLD.pending_expires_at > now()
            AND NEW.selected_group_id IS NOT DISTINCT FROM OLD.selected_group_id
            AND NEW.context_version = OLD.context_version) THEN
      RAISE EXCEPTION 'mcp connection activation refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'mcp_connections_guard';
    END IF;
    RETURN NEW;
  END IF;
  IF NEW.client_id IS DISTINCT FROM OLD.client_id OR NEW.activated_at IS DISTINCT FROM OLD.activated_at THEN
    RAISE EXCEPTION 'mcp connection client binding is immutable' USING ERRCODE = 'check_violation', CONSTRAINT = 'mcp_connections_guard';
  END IF;
  -- Selection writes run only in the Organization-only shape. A select needs the target and any previous selection to be
  -- visible to the subject under row-level security (groups_select: an Organization owner or admin sees every Group,
  -- any other member exactly its own-role Groups); a clear needs nothing more (Spec R5).
  IF NEW.selected_group_id IS DISTINCT FROM OLD.selected_group_id OR NEW.context_version IS DISTINCT FROM OLD.context_version THEN
    IF NOT (OLD.state = 'active' AND NEW.context_version = OLD.context_version + 1 AND g IS NULL
            AND (NEW.selected_group_id IS NULL
                 OR (EXISTS (SELECT 1 FROM app.groups x
                             WHERE x.id = NEW.selected_group_id AND x.organization_id = NEW.organization_id)
                     AND (OLD.selected_group_id IS NULL
                          OR EXISTS (SELECT 1 FROM app.groups y
                                     WHERE y.id = OLD.selected_group_id AND y.organization_id = NEW.organization_id))))) THEN
      RAISE EXCEPTION 'mcp connection selection refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'mcp_connections_guard';
    END IF;
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

CREATE TRIGGER mcp_connections_guard BEFORE UPDATE ON app.mcp_connections FOR EACH ROW EXECUTE FUNCTION app.mcp_connections_guard();

-- The attribution and shape rules with the two MCP actions and the mcp.request / resource_mismatch denial (the
-- version-8 body otherwise unchanged; CREATE OR REPLACE keeps the function's identity, owner, ACL and trigger).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit.events_attribute() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  d integer := pg_trigger_depth();
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  rid text := coalesce(current_setting('audit.request_id', true), '');
  tid text := coalesce(current_setting('audit.trace_id', true), '');
  trig boolean;
  lifecycle boolean;
  plain boolean;
  ok boolean;
BEGIN
  NEW.id := gen_random_uuid();
  NEW.schema_version := 1;
  NEW.occurred_at := clock_timestamp();
  NEW.request_id := CASE WHEN rid ~ '^[A-Za-z0-9._-]{8,64}$' THEN rid END;
  NEW.trace_id := CASE WHEN tid ~ '^[0-9a-f]{32}$' THEN tid END;

  trig := NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed',
    'group.created', 'group.renamed', 'group.slug_changed', 'membership.granted', 'membership.role_changed',
    'membership.removed', 'business_profile.created', 'business_profile.updated', 'business_profile.linked',
    'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled', 'session.issued', 'session.revoked',
    'session.provider_revocation_recorded', 'object.upload_started', 'object.ready', 'object.delete_started',
    'object.deleted', 'mcp.connection_established', 'mcp.context_changed');
  IF (d >= 2) IS DISTINCT FROM coalesce(trig, false) THEN
    RAISE EXCEPTION 'audit event source refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_source';
  END IF;

  -- Actor: never supplied. Lifecycle actions are system only inside the lifecycle producer (the target's IDENTITY
  -- shape); otherwise the context user, or unknown without one.
  lifecycle := NEW.action IN ('user.disabled', 'user.enabled')
    OR (NEW.action IN ('session.revoked', 'revocation.unaudited') AND NEW.reason IS NOT DISTINCT FROM 'disabled');
  IF lifecycle AND u IS NOT NULL AND u = NEW.subject_user_id THEN
    NEW.actor_kind := 'system'; NEW.actor_user_id := NULL;
  ELSIF u IS NOT NULL THEN
    NEW.actor_kind := 'user'; NEW.actor_user_id := u;
  ELSE
    NEW.actor_kind := 'unknown'; NEW.actor_user_id := NULL;
  END IF;

  -- A direct append names an Organization only when it is the established context (the context Organization, with a
  -- current Organization role now), and a Group only when that Group belongs to the Organization and is visible to the
  -- appender under row-level security. With a Group, the groups_select policy itself requires the context
  -- Organization and a non-NULL app.rls_org_role(), so the separate reader call is made only without a Group.
  IF d = 1 THEN
    IF NEW.organization_id IS NOT NULL AND (o IS NULL OR NEW.organization_id <> o) THEN
      RAISE EXCEPTION 'audit event scope refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_scope';
    END IF;
    IF NEW.group_id IS NOT NULL THEN
      IF NOT EXISTS (SELECT 1 FROM app.groups g WHERE g.id = NEW.group_id AND g.organization_id = NEW.organization_id) THEN
        RAISE EXCEPTION 'audit event group refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_group';
      END IF;
    ELSIF NEW.organization_id IS NOT NULL AND app.rls_org_role() IS NULL THEN
      RAISE EXCEPTION 'audit event scope refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_scope';
    END IF;
  END IF;

  plain := NEW.operation IS NULL AND NEW.ttl_seconds IS NULL;
  ok := CASE
    WHEN NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization'
      AND NEW.resource_id = NEW.organization_id
      AND CASE WHEN NEW.action = 'organization.policy_changed'
               THEN NEW.reason IN ('group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled')
               ELSE NEW.reason IS NULL END
    WHEN NEW.action IN ('group.created', 'group.renamed', 'group.slug_changed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.group_id IS NOT NULL
      AND NEW.resource_type = 'group' AND NEW.resource_id = NEW.group_id
    WHEN NEW.action IN ('membership.granted', 'membership.role_changed', 'membership.removed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.organization_id IS NOT NULL
      AND NEW.resource_type = 'role_binding' AND NEW.subject_user_id IS NOT NULL
      AND CASE WHEN NEW.group_id IS NULL THEN NEW.role IN ('org_owner', 'org_admin', 'org_member')
               ELSE NEW.role IN ('group_owner', 'group_admin', 'group_member') END
      AND CASE WHEN NEW.action <> 'membership.role_changed' THEN NEW.previous_role IS NULL
               WHEN NEW.group_id IS NULL THEN NEW.previous_role IN ('org_owner', 'org_admin', 'org_member')
               ELSE NEW.previous_role IN ('group_owner', 'group_admin', 'group_member') END
    WHEN NEW.action IN ('business_profile.created', 'business_profile.updated', 'business_profile.linked',
                        'business_profile.unlinked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'business_profile'
      AND (NEW.group_id IS NULL) = (NEW.action IN ('business_profile.created', 'business_profile.updated'))
    WHEN NEW.action IN ('user.created', 'user.disabled', 'user.enabled') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'user' AND NEW.resource_id = NEW.subject_user_id
    WHEN NEW.action IN ('session.issued', 'session.revoked', 'session.provider_revocation_recorded') THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.subject_user_id IS NOT NULL AND NEW.resource_type = 'session'
      AND CASE NEW.action
            WHEN 'session.issued' THEN NEW.outcome = 'success' AND NEW.reason IS NULL
            WHEN 'session.revoked' THEN NEW.outcome = 'success'
              AND NEW.reason IN ('logout', 'logout_all', 'provider_logout', 'disabled', 'superseded')
            ELSE NEW.outcome IN ('success', 'failed') AND NEW.reason IS NULL END
    WHEN NEW.action IN ('object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'stored_object'
    WHEN NEW.action IN ('mcp.connection_established', 'mcp.context_changed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'mcp_connection' AND NEW.resource_id IS NOT NULL
      AND (NEW.action = 'mcp.context_changed' OR NEW.group_id IS NULL)
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object')))
      AND (NEW.reason IS DISTINCT FROM 'resource_mismatch'
           OR (NEW.operation = 'mcp.request' AND NEW.organization_id IS NULL AND NEW.resource_type IS NULL))
    WHEN NEW.action = 'object.url_issued' THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND NEW.operation IS NULL AND NEW.ttl_seconds IS NOT NULL
      AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'stored_object'
    WHEN NEW.action IN ('object.upload_failed', 'object.deletion_pending') THEN
      NEW.outcome IN ('failed', 'unknown') AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.resource_type = 'stored_object'
      AND CASE WHEN NEW.action = 'object.upload_failed'
               THEN NEW.reason IN ('compensated', 'compensation_failed', 'not_compensated')
               ELSE NEW.reason IN ('provider_unavailable', 'tombstone_failed', 'tombstone_unknown') END
    WHEN NEW.action = 'revocation.unaudited' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL
      AND ((NEW.reason IN ('logout', 'provider_logout') AND NEW.resource_type = 'session')
        OR (NEW.reason IN ('logout_all', 'disabled') AND NEW.resource_type = 'user' AND NEW.resource_id = NEW.subject_user_id))
    ELSE false
  END;
  IF NOT coalesce(ok, false) THEN
    RAISE EXCEPTION 'audit event shape refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'audit_event_shape';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- The capture function (owner-owned SECURITY INVOKER): activation and selection changes, strictly atomic.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_mcp_connections() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_NAME = 'audit_capture_state' THEN
    IF OLD.state = 'pending' AND NEW.state = 'active' THEN
      INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id, subject_user_id)
      VALUES ('mcp.connection_established', 'success', NEW.organization_id, 'mcp_connection', NEW.id, NEW.user_id);
    END IF;
  ELSIF NEW.selected_group_id IS DISTINCT FROM OLD.selected_group_id THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id)
    VALUES ('mcp.context_changed', 'success', NEW.organization_id, NEW.selected_group_id, 'mcp_connection', NEW.id, NEW.user_id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

CREATE TRIGGER audit_capture_state AFTER UPDATE OF state ON app.mcp_connections FOR EACH ROW
  WHEN (OLD.state IS DISTINCT FROM NEW.state) EXECUTE FUNCTION audit.capture_mcp_connections();
CREATE TRIGGER audit_capture_selection AFTER UPDATE OF selected_group_id ON app.mcp_connections FOR EACH ROW
  WHEN (OLD.selected_group_id IS DISTINCT FROM NEW.selected_group_id) EXECUTE FUNCTION audit.capture_mcp_connections();

-- The widened catalog constraints (the version-8 lists plus the G014 values).
ALTER TABLE audit.events DROP CONSTRAINT events_action_valid, DROP CONSTRAINT events_reason_valid,
  DROP CONSTRAINT events_operation_valid, DROP CONSTRAINT events_resource_type_valid;
ALTER TABLE audit.events
  ADD CONSTRAINT events_action_valid CHECK (action IN (
                    'organization.created', 'organization.renamed', 'organization.policy_changed',
                    'group.created', 'group.renamed', 'group.slug_changed',
                    'membership.granted', 'membership.role_changed', 'membership.removed',
                    'business_profile.created', 'business_profile.updated', 'business_profile.linked',
                    'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled',
                    'session.issued', 'session.revoked', 'session.provider_revocation_recorded',
                    'object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted',
                    'mcp.connection_established', 'mcp.context_changed',
                    'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                    'object.deletion_pending', 'revocation.unaudited')),
  ADD CONSTRAINT events_reason_valid CHECK (reason IS NULL OR reason IN (
                    'group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled',
                    'logout', 'logout_all', 'provider_logout', 'disabled', 'superseded',
                    'signed_in', 'session_rotated', 'transaction_expired', 'login_canceled', 'code_rejected',
                    'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
                    'provider_error', 'provider_timeout', 'identity_namespace_mismatch', 'identity_unavailable',
                    'organization_not_permitted', 'group_not_selectable', 'target_not_found', 'target_kind_mismatch',
                    'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled', 'not_creator',
                    'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
                    'compensated', 'compensation_failed', 'not_compensated',
                    'provider_unavailable', 'tombstone_failed', 'tombstone_unknown')),
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection'));

-- The E1 reader resolves browser sessions only (replaced by its owner, the auth reader, through the session's SET chain).
-- +goose StatementBegin
DO $do$
DECLARE
  db text := current_database();
  ow text := current_user;
  ar text := db || '_auth_reader';
BEGIN
  EXECUTE format('GRANT SELECT (kind) ON app.sessions TO %I', ar);
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $e1$
CREATE OR REPLACE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz)
  RETURNS TABLE (session_id uuid, user_id uuid, display_name text, mfa_assurance text, auth_time timestamptz,
                 idle_expires_at timestamptz, absolute_expires_at timestamptz, renewed_at timestamptz, csrf_token bytea,
                 provider_session_id text)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN QUERY
  SELECT s.id, s.user_id, u.display_name, s.mfa_assurance, s.auth_time, s.idle_expires_at, s.absolute_expires_at,
         s.renewed_at, s.csrf_token, s.provider_session_id
  FROM app.sessions s JOIN app.users u ON u.id = s.user_id
  WHERE s.credential_hash = p_hash AND s.revoked_at IS NULL AND s.idle_expires_at > p_at AND s.absolute_expires_at > p_at
    AND u.status = 'active' AND s.kind = 'browser';
END $$
$e1$;
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);
END
$do$;
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
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  actual text;
  expected text;
BEGIN
  REVOKE ALL ON FUNCTION app.mcp_connections_guard(), audit.capture_mcp_connections() FROM PUBLIC;

  -- Runtime: read own rows, insert a pending row, and update only the lifecycle and selection columns. No DELETE.
  EXECUTE format('GRANT SELECT ON app.mcp_connections TO %I', rt::text);
  EXECUTE format('GRANT INSERT (id, session_id, user_id, organization_id, host, resource, pending_expires_at, created_at) '
              || 'ON app.mcp_connections TO %I', rt::text);
  EXECUTE format('GRANT UPDATE (state, client_id, activated_at, selected_group_id, context_version) ON app.mcp_connections TO %I', rt::text);

  ALTER TABLE app.mcp_connections ENABLE ROW LEVEL SECURITY;
  ALTER TABLE app.mcp_connections FORCE ROW LEVEL SECURITY;
  EXECUTE format('CREATE POLICY mcp_connections_select ON app.mcp_connections FOR SELECT TO %I, %I USING ('
    || 'user_id = %s AND organization_id = %s)', rt::text, ow::text, U, O);
  EXECUTE format('CREATE POLICY mcp_connections_insert ON app.mcp_connections FOR INSERT TO %I, %I WITH CHECK ('
    || 'user_id = %s AND organization_id = %s AND %s IS NULL AND %s IS NOT NULL AND state = ''pending'' '
    || 'AND selected_group_id IS NULL AND context_version = 0)', rt::text, ow::text, U, O, G, R);
  EXECUTE format('CREATE POLICY mcp_connections_update ON app.mcp_connections FOR UPDATE TO %I, %I USING ('
    || 'user_id = %s AND organization_id = %s AND %s IS NOT NULL) WITH CHECK ('
    || 'user_id = %s AND organization_id = %s AND %s IS NOT NULL)', rt::text, ow::text, U, O, R, U, O, R);
  -- Browser rows exactly as before; 'mcp' rows only in the TENANT-ORG shape with a current Organization role.
  EXECUTE format('ALTER POLICY sessions_insert ON app.sessions WITH CHECK ('
    || 'user_id = %s AND revoked_at IS NULL AND ((kind = ''browser'' AND %s IS NULL) '
    || 'OR (kind = ''mcp'' AND %s IS NOT NULL AND %s IS NULL AND %s IS NOT NULL)))', U, O, O, G, R);

  -- Postconditions: the exact version-10 inventories.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  IF actual IS DISTINCT FROM 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,mcp_connections+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force' THEN
    RAISE EXCEPTION 'app row-level security tables are %', actual;
  END IF;
  -- The 42 application relations (38 plus mcp_connections, its primary key and its two unique indexes) plus the
  -- migration ledger's table, identity sequence and primary key.
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace;
  IF actual <> '45' THEN
    RAISE EXCEPTION 'app relation count is %, expected 45 (42 plus the migration ledger)', actual;
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
    || 'mcp_connections_insert:mcp_connections:a:true:ow+rt;mcp_connections_select:mcp_connections:r:true:ow+rt;'
    || 'mcp_connections_update:mcp_connections:w:true:ow+rt;'
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
    || 'mcp_connections:rt:SELECT;mcp_connections.activated_at:rt:UPDATE;mcp_connections.client_id:rt:UPDATE;'
    || 'mcp_connections.context_version:rt:UPDATE;mcp_connections.created_at:rt:INSERT;mcp_connections.host:rt:INSERT;'
    || 'mcp_connections.id:rt:INSERT;mcp_connections.organization_id:rt:INSERT;mcp_connections.pending_expires_at:rt:INSERT;'
    || 'mcp_connections.resource:rt:INSERT;mcp_connections.selected_group_id:rt:UPDATE;mcp_connections.session_id:rt:INSERT;'
    || 'mcp_connections.state:rt:UPDATE;mcp_connections.user_id:rt:INSERT;'
    || 'organization_memberships:rt:SELECT;organizations:rt:SELECT;organizations.group_admins_edit_business_profiles:rt:UPDATE;'
    || 'role_bindings.group_id:pr:SELECT;role_bindings.group_id:rt:SELECT;role_bindings.organization_id:pr:SELECT;'
    || 'role_bindings.organization_id:rt:SELECT;role_bindings.role:pr:SELECT;role_bindings.role:rt:SELECT,UPDATE;'
    || 'role_bindings.scope_id:pr:SELECT;role_bindings.scope_id:rt:SELECT;role_bindings.scope_type:pr:SELECT;'
    || 'role_bindings.scope_type:rt:SELECT;role_bindings.source:pr:SELECT;role_bindings.source:rt:SELECT;'
    || 'role_bindings.subject_user_id:pr:SELECT;role_bindings.subject_user_id:rt:SELECT;sessions:rt:INSERT,SELECT;'
    || 'sessions.absolute_expires_at:ar:SELECT;sessions.auth_time:ar:SELECT;sessions.credential_hash:ar:SELECT;'
    || 'sessions.csrf_token:ar:SELECT;sessions.id:ar:SELECT;sessions.idle_expires_at:ar:SELECT;sessions.idle_expires_at:rt:UPDATE;sessions.kind:ar:SELECT;'
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
    || 'mcp_connections_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
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
    || 'capture_mcp_connections:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
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
  -- The app trigger inventory: the 20 capture triggers plus the existing guards and the connection guard.
  SELECT string_agg(format('%s.%s:%s:%s:%s:%s', c.relname, t.tgname, t.tgtype, t.tgenabled,
           p.pronamespace::regnamespace::text || '.' || p.proname,
           CASE WHEN t.tgnewtable IS NOT NULL THEN 'tt' WHEN t.tgqual IS NOT NULL THEN 'when' ELSE '-' END),
           ';' ORDER BY c.relname, t.tgname) INTO actual
  FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid JOIN pg_proc p ON p.oid = t.tgfoid
  WHERE c.relnamespace = 'app'::regnamespace AND NOT t.tgisinternal;
  expected := 'business_profiles.audit_capture_insert:5:O:audit.capture_business_profiles:-;'
    || 'business_profiles.audit_capture_update:17:O:audit.capture_business_profiles:when;'
    || 'business_profiles.business_profiles_keys_immutable:19:O:app.business_profiles_keys_immutable:-;'
    || 'group_business_profiles.audit_capture_delete:9:O:audit.capture_group_business_profiles:-;'
    || 'group_business_profiles.audit_capture_insert:5:O:audit.capture_group_business_profiles:-;'
    || 'group_business_profiles.group_business_profiles_keys_immutable:19:O:app.group_business_profiles_keys_immutable:-;'
    || 'groups.audit_capture_insert:5:O:audit.capture_groups:-;'
    || 'groups.audit_capture_update:17:O:audit.capture_groups:when;'
    || 'groups.groups_keys_immutable:19:O:app.groups_keys_immutable:-;'
    || 'mcp_connections.audit_capture_selection:17:O:audit.capture_mcp_connections:when;'
    || 'mcp_connections.audit_capture_state:17:O:audit.capture_mcp_connections:when;'
    || 'mcp_connections.mcp_connections_guard:19:O:app.mcp_connections_guard:-;'
    || 'organizations.audit_capture_insert:5:O:audit.capture_organizations:-;'
    || 'organizations.audit_capture_update:17:O:audit.capture_organizations:when;'
    || 'organizations.organizations_keys_immutable:19:O:app.organizations_keys_immutable:-;'
    || 'role_bindings.audit_capture_delete:9:O:audit.capture_role_bindings:-;'
    || 'role_bindings.audit_capture_insert:5:O:audit.capture_role_bindings:-;'
    || 'role_bindings.audit_capture_update:17:O:audit.capture_role_bindings:when;'
    || 'role_bindings.role_bindings_keys_immutable:19:O:app.role_bindings_keys_immutable:-;'
    || 'sessions.audit_capture_insert:5:O:audit.capture_sessions:-;'
    || 'sessions.audit_capture_revocations:16:O:audit.capture_session_revocations:tt;'
    || 'sessions.audit_capture_update:17:O:audit.capture_sessions:when;'
    || 'stored_objects.audit_capture_insert:5:O:audit.capture_stored_objects:-;'
    || 'stored_objects.audit_capture_update:17:O:audit.capture_stored_objects:when;'
    || 'stored_objects.stored_objects_guard:19:O:app.stored_objects_guard:-;'
    || 'users.audit_capture_insert:5:O:audit.capture_users:-;'
    || 'users.audit_capture_update:17:O:audit.capture_users:when';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'app trigger inventory is %, expected %', actual, expected;
  END IF;
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'audit'::regnamespace;
  IF actual <> '3' THEN
    RAISE EXCEPTION 'audit relation count is %, expected 3', actual;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  db text := current_database();
  ow text := current_user;
  ar text := db || '_auth_reader';
BEGIN
  -- Without FORCE the owner's reads are not filtered by row-level security. MCP state and history are never silently
  -- discarded: refuse while any exists (the refusal rolls back the NO FORCE with everything else).
  ALTER TABLE app.mcp_connections NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.sessions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  IF EXISTS (SELECT 1 FROM app.mcp_connections) OR EXISTS (SELECT 1 FROM app.sessions WHERE kind = 'mcp') THEN
    RAISE EXCEPTION 'refusing to roll back 00010: MCP connections or mcp sessions exist';
  END IF;
  IF EXISTS (SELECT 1 FROM audit.events WHERE action IN ('mcp.connection_established', 'mcp.context_changed')
               OR reason = 'resource_mismatch' OR operation = 'mcp.request' OR resource_type = 'mcp_connection') THEN
    RAISE EXCEPTION 'refusing to roll back 00010: audit.events holds MCP history (history is never discarded)';
  END IF;
  ALTER TABLE app.sessions FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  -- The version-9 E1 reader, replaced by its owner.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $e1$
CREATE OR REPLACE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz)
  RETURNS TABLE (session_id uuid, user_id uuid, display_name text, mfa_assurance text, auth_time timestamptz,
                 idle_expires_at timestamptz, absolute_expires_at timestamptz, renewed_at timestamptz, csrf_token bytea,
                 provider_session_id text)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN QUERY
  SELECT s.id, s.user_id, u.display_name, s.mfa_assurance, s.auth_time, s.idle_expires_at, s.absolute_expires_at,
         s.renewed_at, s.csrf_token, s.provider_session_id
  FROM app.sessions s JOIN app.users u ON u.id = s.user_id
  WHERE s.credential_hash = p_hash AND s.revoked_at IS NULL AND s.idle_expires_at > p_at AND s.absolute_expires_at > p_at
    AND u.status = 'active';
END $$
$e1$;
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);

  -- The version-9 sessions_insert policy (00006).
  ALTER POLICY sessions_insert ON app.sessions WITH CHECK (
    user_id = (SELECT app.ctx_user()) AND (SELECT app.ctx_org()) IS NULL AND revoked_at IS NULL);

  DROP TABLE app.mcp_connections;
  DROP FUNCTION app.mcp_connections_guard();
  DROP FUNCTION audit.capture_mcp_connections();
  ALTER TABLE app.sessions DROP COLUMN kind;

  ALTER TABLE audit.events DROP CONSTRAINT events_action_valid, DROP CONSTRAINT events_reason_valid,
    DROP CONSTRAINT events_operation_valid, DROP CONSTRAINT events_resource_type_valid;
  ALTER TABLE audit.events
    ADD CONSTRAINT events_action_valid CHECK (action IN (
                      'organization.created', 'organization.renamed', 'organization.policy_changed',
                      'group.created', 'group.renamed', 'group.slug_changed',
                      'membership.granted', 'membership.role_changed', 'membership.removed',
                      'business_profile.created', 'business_profile.updated', 'business_profile.linked',
                      'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled',
                      'session.issued', 'session.revoked', 'session.provider_revocation_recorded',
                      'object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted',
                      'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                      'object.deletion_pending', 'revocation.unaudited')),
    ADD CONSTRAINT events_reason_valid CHECK (reason IS NULL OR reason IN (
                      'group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled',
                      'logout', 'logout_all', 'provider_logout', 'disabled', 'superseded',
                      'signed_in', 'session_rotated', 'transaction_expired', 'login_canceled', 'code_rejected',
                      'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
                      'provider_error', 'provider_timeout', 'identity_namespace_mismatch', 'identity_unavailable',
                      'organization_not_permitted', 'group_not_selectable', 'target_not_found', 'target_kind_mismatch',
                      'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled', 'not_creator',
                      'object_not_ready', 'capability_unknown', 'fact_inconsistent',
                      'compensated', 'compensation_failed', 'not_compensated',
                      'provider_unavailable', 'tombstone_failed', 'tombstone_unknown')),
    ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                      'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                      'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                      'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read')),
    ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                      ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object'));
END
$do$;
-- +goose StatementEnd

-- The exact version-8 attribution body (00008).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit.events_attribute() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  d integer := pg_trigger_depth();
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  rid text := coalesce(current_setting('audit.request_id', true), '');
  tid text := coalesce(current_setting('audit.trace_id', true), '');
  trig boolean;
  lifecycle boolean;
  plain boolean;
  ok boolean;
BEGIN
  NEW.id := gen_random_uuid();
  NEW.schema_version := 1;
  NEW.occurred_at := clock_timestamp();
  NEW.request_id := CASE WHEN rid ~ '^[A-Za-z0-9._-]{8,64}$' THEN rid END;
  NEW.trace_id := CASE WHEN tid ~ '^[0-9a-f]{32}$' THEN tid END;

  trig := NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed',
    'group.created', 'group.renamed', 'group.slug_changed', 'membership.granted', 'membership.role_changed',
    'membership.removed', 'business_profile.created', 'business_profile.updated', 'business_profile.linked',
    'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled', 'session.issued', 'session.revoked',
    'session.provider_revocation_recorded', 'object.upload_started', 'object.ready', 'object.delete_started',
    'object.deleted');
  IF (d >= 2) IS DISTINCT FROM coalesce(trig, false) THEN
    RAISE EXCEPTION 'audit event source refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_source';
  END IF;

  -- Actor: never supplied. Lifecycle actions are system only inside the lifecycle producer (the target's IDENTITY
  -- shape); otherwise the context user, or unknown without one.
  lifecycle := NEW.action IN ('user.disabled', 'user.enabled')
    OR (NEW.action IN ('session.revoked', 'revocation.unaudited') AND NEW.reason IS NOT DISTINCT FROM 'disabled');
  IF lifecycle AND u IS NOT NULL AND u = NEW.subject_user_id THEN
    NEW.actor_kind := 'system'; NEW.actor_user_id := NULL;
  ELSIF u IS NOT NULL THEN
    NEW.actor_kind := 'user'; NEW.actor_user_id := u;
  ELSE
    NEW.actor_kind := 'unknown'; NEW.actor_user_id := NULL;
  END IF;

  -- A direct append names an Organization only when it is the established context (the context Organization, with a
  -- current Organization role now), and a Group only when that Group belongs to the Organization and is visible to the
  -- appender under row-level security. With a Group, the groups_select policy itself requires the context
  -- Organization and a non-NULL app.rls_org_role(), so the separate reader call is made only without a Group.
  IF d = 1 THEN
    IF NEW.organization_id IS NOT NULL AND (o IS NULL OR NEW.organization_id <> o) THEN
      RAISE EXCEPTION 'audit event scope refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_scope';
    END IF;
    IF NEW.group_id IS NOT NULL THEN
      IF NOT EXISTS (SELECT 1 FROM app.groups g WHERE g.id = NEW.group_id AND g.organization_id = NEW.organization_id) THEN
        RAISE EXCEPTION 'audit event group refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_group';
      END IF;
    ELSIF NEW.organization_id IS NOT NULL AND app.rls_org_role() IS NULL THEN
      RAISE EXCEPTION 'audit event scope refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_scope';
    END IF;
  END IF;

  plain := NEW.operation IS NULL AND NEW.ttl_seconds IS NULL;
  ok := CASE
    WHEN NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization'
      AND NEW.resource_id = NEW.organization_id
      AND CASE WHEN NEW.action = 'organization.policy_changed'
               THEN NEW.reason IN ('group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled')
               ELSE NEW.reason IS NULL END
    WHEN NEW.action IN ('group.created', 'group.renamed', 'group.slug_changed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.group_id IS NOT NULL
      AND NEW.resource_type = 'group' AND NEW.resource_id = NEW.group_id
    WHEN NEW.action IN ('membership.granted', 'membership.role_changed', 'membership.removed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.organization_id IS NOT NULL
      AND NEW.resource_type = 'role_binding' AND NEW.subject_user_id IS NOT NULL
      AND CASE WHEN NEW.group_id IS NULL THEN NEW.role IN ('org_owner', 'org_admin', 'org_member')
               ELSE NEW.role IN ('group_owner', 'group_admin', 'group_member') END
      AND CASE WHEN NEW.action <> 'membership.role_changed' THEN NEW.previous_role IS NULL
               WHEN NEW.group_id IS NULL THEN NEW.previous_role IN ('org_owner', 'org_admin', 'org_member')
               ELSE NEW.previous_role IN ('group_owner', 'group_admin', 'group_member') END
    WHEN NEW.action IN ('business_profile.created', 'business_profile.updated', 'business_profile.linked',
                        'business_profile.unlinked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'business_profile'
      AND (NEW.group_id IS NULL) = (NEW.action IN ('business_profile.created', 'business_profile.updated'))
    WHEN NEW.action IN ('user.created', 'user.disabled', 'user.enabled') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'user' AND NEW.resource_id = NEW.subject_user_id
    WHEN NEW.action IN ('session.issued', 'session.revoked', 'session.provider_revocation_recorded') THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.subject_user_id IS NOT NULL AND NEW.resource_type = 'session'
      AND CASE NEW.action
            WHEN 'session.issued' THEN NEW.outcome = 'success' AND NEW.reason IS NULL
            WHEN 'session.revoked' THEN NEW.outcome = 'success'
              AND NEW.reason IN ('logout', 'logout_all', 'provider_logout', 'disabled', 'superseded')
            ELSE NEW.outcome IN ('success', 'failed') AND NEW.reason IS NULL END
    WHEN NEW.action IN ('object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'stored_object'
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object')))
    WHEN NEW.action = 'object.url_issued' THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND NEW.operation IS NULL AND NEW.ttl_seconds IS NOT NULL
      AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'stored_object'
    WHEN NEW.action IN ('object.upload_failed', 'object.deletion_pending') THEN
      NEW.outcome IN ('failed', 'unknown') AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.resource_type = 'stored_object'
      AND CASE WHEN NEW.action = 'object.upload_failed'
               THEN NEW.reason IN ('compensated', 'compensation_failed', 'not_compensated')
               ELSE NEW.reason IN ('provider_unavailable', 'tombstone_failed', 'tombstone_unknown') END
    WHEN NEW.action = 'revocation.unaudited' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL
      AND ((NEW.reason IN ('logout', 'provider_logout') AND NEW.resource_type = 'session')
        OR (NEW.reason IN ('logout_all', 'disabled') AND NEW.resource_type = 'user' AND NEW.resource_id = NEW.subject_user_id))
    ELSE false
  END;
  IF NOT coalesce(ok, false) THEN
    RAISE EXCEPTION 'audit event shape refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'audit_event_shape';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd
