-- Hostname resolution and verified domain lifecycle (G016; ADR-007; docs/foundation/DOMAINS.md).
--
-- Additive, one goose transaction:
--   - app.organizations.slug: the optional platform hostname label of an Organization (<slug>.<platform domain>). NULL
--     means no platform hostname. Grammar and the reserved first-label floor are CHECK constraints (the floor is pinned
--     identically in core/internal/hostname); one Organization per slug (partial unique index). A slug is a routing
--     alias, never authority. Runtime may UPDATE (slug) under the unchanged owner-only organizations_update policy or the
--     added owner/admin organizations_update_slug policy; an owner-owned guard (organizations_slug_guard) lets an
--     Organization admin change the slug and nothing else (2026-10-16 deviation #1 closure). Changes are captured as
--     organization.slug_changed.
--   - app.organization_domains: one row per custom-hostname claim generation (core-minted UUID, immutable Organization,
--     hostname and generation), the finite state pending -> verified -> active <-> suspended -> removed, the ownership
--     challenge (pending only), proof freshness, TLS readiness and the removed row's terminal cleanup outcome, under an
--     optimistic version. At most one live (non-removed) row per hostname; a reassignment is a new row. The
--     generation is an opaque positive int63 claim fence minted by core from a CSPRNG and compared only for equality
--     with the row id (2026-10-16 deviation #3 closure: no per-hostname history counter, no cross-Organization reader). FORCE row-level security: Organization owners and admins of the context Organization read, insert a
--     pending row and update; no DELETE path. An owner-owned SECURITY INVOKER guard implements the lifecycle table of
--     Plan r4 §1.4 literally.
--   - Readers (SECURITY DEFINER, owned by <db>_auth_reader, EXECUTE to runtime only; they ignore ambient settings):
--     app.hostname_state(kind, key) returns the routing and issuance predicates of one platform slug or custom
--     hostname (one unique-index probe); app.is_organization_member(user, org) returns whether a current Organization membership binding exists (the
--     tenant-host session-issuance guard). New auth-reader SELECT policies on organizations, organization_domains and
--     role_bindings admit only these readers' column grants.
--   - app.login_transactions gains organization_id (the Organization pinned at a tenant-host login) and callback_origin
--     (the shared callback relay origin, B2). Row-level security is unchanged (admission by the binding hash).
--   - Audit (G011): four trigger-owned actions (domain.claimed, domain.state_changed, domain.checked,
--     organization.slug_changed), resource type organization_domain, access.denied operations for the domain
--     operations and organization.set_slug, auth.sign_in reasons organization_not_member and host_changed.
--
-- Locks: ALTER TABLE takes a brief ACCESS EXCLUSIVE lock on app.organizations and app.login_transactions (nullable
-- columns, no rewrite) and on audit.events (the four widened catalog CHECK constraints are validated by one scan); the
-- foreign keys take SHARE ROW EXCLUSIVE on app.organizations and app.users. Migrate in a quiescent window: migrate,
-- then deploy the G016 code (docs/foundation/MIGRATIONS.md).
--
-- Down (development and test only; the runner refuses down/reapply with APP_ENV=production) refuses while any domain
-- row, slug, tenant-host login transaction or domain audit event exists, then restores the exact version-10 catalog.

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

  IF to_regclass('app.organization_domains') IS NOT NULL
     OR EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.organizations'::regclass AND attname = 'slug' AND NOT attisdropped)
     OR EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.login_transactions'::regclass
                  AND attname IN ('organization_id', 'callback_origin') AND NOT attisdropped) THEN
    RAISE EXCEPTION 'version-10 precondition: domain objects already exist';
  END IF;

  -- Preconditions: the exact version-10 inventories (00010 postconditions).
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  IF actual IS DISTINCT FROM 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,mcp_connections+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force' THEN
    RAISE EXCEPTION 'version-10 precondition: app row-level security tables are %', actual;
  END IF;
  -- The 42 application relations (38 plus mcp_connections, its primary key and its two unique indexes) plus the
  -- migration ledger's table, identity sequence and primary key.
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace;
  IF actual <> '45' THEN
    RAISE EXCEPTION 'version-10 precondition: app relation count is %, expected 45 (42 plus the migration ledger)', actual;
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
    RAISE EXCEPTION 'version-10 precondition: policy inventory is %, expected %', actual, expected;
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
    RAISE EXCEPTION 'version-10 precondition: app ACLs are %, expected %', actual, expected;
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
    RAISE EXCEPTION 'version-10 precondition: app function allowlist is %, expected %', actual, expected;
  END IF;
  SELECT string_agg(c.relname || ':' || c.relkind::text || ':' || CASE c.relowner WHEN ow THEN 'ow' ELSE 'other' END
           || CASE WHEN c.relrowsecurity THEN '+rls' ELSE '' END || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END,
           ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'audit'::regnamespace;
  IF actual IS DISTINCT FROM 'events:r:ow+rls+force,events_pkey:i:ow,events_tenant_time_idx:i:ow' THEN
    RAISE EXCEPTION 'version-10 precondition: audit relations are %', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s', t.tgname, t.tgtype, t.tgenabled, p.proname), ';' ORDER BY t.tgname) INTO actual
  FROM pg_trigger t JOIN pg_proc p ON p.oid = t.tgfoid WHERE t.tgrelid = 'audit.events'::regclass AND NOT t.tgisinternal;
  IF actual IS DISTINCT FROM 'events_attribute:7:O:events_attribute;events_refuse:27:O:events_refuse;'
                             || 'events_refuse_truncate:34:O:events_refuse' THEN
    RAISE EXCEPTION 'version-10 precondition: audit.events triggers are %', actual;
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
    RAISE EXCEPTION 'version-10 precondition: audit ACLs are %, expected %', actual, expected;
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
    RAISE EXCEPTION 'version-10 precondition: audit function allowlist is %, expected %', actual, expected;
  END IF;
  -- The two reviewed bodies (the refusal body with the embedded role name normalized to <db>_retention).
  SELECT md5(replace(p.prosrc, rn::text, '<db>_retention')) INTO actual FROM pg_proc p WHERE p.oid = 'audit.events_refuse()'::regprocedure;
  IF actual IS DISTINCT FROM 'c6c0b51ae2876ba67658dc4041eec5e9' THEN
    RAISE EXCEPTION 'version-10 precondition: audit.events_refuse body digest is %', actual;
  END IF;
  SELECT md5(p.prosrc) INTO actual FROM pg_proc p WHERE p.oid = 'app.retention_eligible(uuid, timestamptz, uuid[])'::regprocedure;
  IF actual IS DISTINCT FROM '89c969a9ec364eae864aab3e9fda0d58' THEN
    RAISE EXCEPTION 'version-10 precondition: app.retention_eligible body digest is %', actual;
  END IF;
  SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") INTO actual FROM (
    SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN rn THEN 'rn' ELSE a.grantee::regrole::text END
           || '=' || a.privilege_type AS v
    FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
    WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) l;
  IF actual IS DISTINCT FROM 'rn=USAGE,rt=USAGE' THEN
    RAISE EXCEPTION 'version-10 precondition: schema audit privileges are %, expected rn=USAGE,rt=USAGE', actual;
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
    RAISE EXCEPTION 'version-10 precondition: app trigger inventory is %, expected %', actual, expected;
  END IF;
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'audit'::regnamespace;
  IF actual <> '3' THEN
    RAISE EXCEPTION 'version-10 precondition: audit relation count is %, expected 3', actual;
  END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE app.organizations ADD COLUMN slug text COLLATE "C" NULL
  CONSTRAINT organizations_slug_format CHECK (char_length(slug) BETWEEN 1 AND 63 AND slug ~ '^[0-9a-z]+(-[0-9a-z]+)*$')
  CONSTRAINT organizations_slug_reserved CHECK (slug NOT IN (
    'account', 'admin', 'api', 'app', 'auth', 'billing', 'blueprint', 'cdn', 'custom', 'fonts', 'health', 'healthz',
    'internal', 'login', 'logout', 'mail', 'mcp', 'members', 'message', 'messages', 'new', 'ns', 'ns1', 'ns2',
    'onboarding', 'readyz', 'rpc', 'settings', 'sign-in', 'smtp', 'sse', 'static', 'status', 'system', 'www'));
CREATE UNIQUE INDEX organizations_slug_key ON app.organizations (slug) WHERE slug IS NOT NULL;

ALTER TABLE app.login_transactions
  ADD COLUMN organization_id uuid NULL,
  ADD COLUMN callback_origin text NULL
    CONSTRAINT login_transactions_callback_origin_valid CHECK (char_length(callback_origin) BETWEEN 1 AND 2048);

CREATE TABLE app.organization_domains (
  id                   uuid        NOT NULL CONSTRAINT organization_domains_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  organization_id      uuid        NOT NULL,
  hostname             text        COLLATE "C" NOT NULL CONSTRAINT organization_domains_hostname_valid CHECK (
                                     char_length(hostname) BETWEEN 3 AND 253
                                     AND hostname ~ '^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$'
                                     AND hostname !~ '(^|\.)xn--' AND hostname !~ '\.[0-9]+$'),
  generation           bigint      NOT NULL DEFAULT 1 CONSTRAINT organization_domains_generation_valid CHECK (generation >= 1),
  state                text        NOT NULL DEFAULT 'pending' CONSTRAINT organization_domains_state_valid
                                     CHECK (state IN ('pending', 'verified', 'active', 'suspended', 'removed')),
  challenge_value      text        NOT NULL CONSTRAINT organization_domains_challenge_valid CHECK (challenge_value ~ '^[a-z2-7]{52}$'),
  challenge_expires_at timestamptz NOT NULL,
  verified_at          timestamptz NULL,
  verified_until       timestamptz NULL,
  last_check_at        timestamptz NULL,
  last_check_outcome   text        NULL CONSTRAINT organization_domains_outcome_valid CHECK (last_check_outcome IN (
                                     'ok', 'mismatch', 'nxdomain', 'transient', 'tls_error', 'withdrawn', 'withdraw_error',
                                     'withdraw_unknown')),
  check_failures       integer     NOT NULL DEFAULT 0 CONSTRAINT organization_domains_failures_valid CHECK (check_failures >= 0),
  tls_state            text        NOT NULL DEFAULT 'none' CONSTRAINT organization_domains_tls_state_valid CHECK (tls_state IN (
                                     'none', 'requested', 'pending', 'ready', 'error', 'expired', 'withdrawing', 'withdrawn',
                                     'unknown')),
  tls_checked_at       timestamptz NULL,
  tls_not_after        timestamptz NULL,
  tls_reference        text        NULL CONSTRAINT organization_domains_tls_reference_valid CHECK (char_length(tls_reference) BETWEEN 1 AND 256),
  removed_at           timestamptz NULL,
  created_by           uuid        NOT NULL,
  created_at           timestamptz NOT NULL DEFAULT now(),
  version              bigint      NOT NULL DEFAULT 0 CONSTRAINT organization_domains_version_valid CHECK (version >= 0),
  CONSTRAINT organization_domains_pkey PRIMARY KEY (id),
  CONSTRAINT organization_domains_hostname_generation_key UNIQUE (hostname, generation),
  CONSTRAINT organization_domains_organization_fkey FOREIGN KEY (organization_id) REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT organization_domains_created_by_fkey FOREIGN KEY (created_by) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT organization_domains_removed_shape CHECK ((state = 'removed') = (removed_at IS NOT NULL)),
  CONSTRAINT organization_domains_verified_shape CHECK (state NOT IN ('verified', 'active', 'suspended')
                                                        OR (verified_at IS NOT NULL AND verified_until IS NOT NULL)),
  CONSTRAINT organization_domains_active_shape CHECK (state <> 'active' OR (tls_state = 'ready' AND tls_not_after IS NOT NULL)),
  CONSTRAINT organization_domains_tls_shape CHECK (tls_state <> 'ready' OR tls_not_after IS NOT NULL)
);
CREATE UNIQUE INDEX organization_domains_hostname_live ON app.organization_domains (hostname) WHERE state <> 'removed';
CREATE INDEX organization_domains_organization ON app.organization_domains (organization_id, created_at);

-- The guard (owner-owned SECURITY INVOKER) implements the lifecycle table of Plan r4 §1.4: the binding columns never
-- change, every update advances the version by one, live rows move only along the listed edges, the challenge changes
-- only on a pending row that stays pending, verified_at is set once, freshness moves forward on success and drops to
-- exactly now() on a confirmed proof loss (from verified, active or an already suspended row), removal starts the
-- withdrawal, and a removed row records only its terminal cleanup outcome. Anything else raises.
-- +goose StatementBegin
CREATE FUNCTION app.organization_domains_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF NEW.id IS DISTINCT FROM OLD.id OR NEW.organization_id IS DISTINCT FROM OLD.organization_id
     OR NEW.hostname IS DISTINCT FROM OLD.hostname OR NEW.generation IS DISTINCT FROM OLD.generation
     OR NEW.created_by IS DISTINCT FROM OLD.created_by OR NEW.created_at IS DISTINCT FROM OLD.created_at
     OR NEW.version IS DISTINCT FROM OLD.version + 1 THEN
    RAISE EXCEPTION 'organization domain binding or version refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_domains_guard';
  END IF;
  IF OLD.state = 'removed' THEN
    IF NEW.state <> 'removed' OR NEW.challenge_value IS DISTINCT FROM OLD.challenge_value
       OR NEW.challenge_expires_at IS DISTINCT FROM OLD.challenge_expires_at OR NEW.verified_at IS DISTINCT FROM OLD.verified_at
       OR NEW.verified_until IS DISTINCT FROM OLD.verified_until OR NEW.tls_not_after IS DISTINCT FROM OLD.tls_not_after
       OR NEW.tls_reference IS DISTINCT FROM OLD.tls_reference OR NEW.removed_at IS DISTINCT FROM OLD.removed_at
       OR NEW.check_failures < OLD.check_failures
       OR (NEW.tls_state IS DISTINCT FROM OLD.tls_state
           AND (OLD.tls_state NOT IN ('withdrawing', 'unknown', 'error') OR NEW.tls_state NOT IN ('withdrawn', 'unknown', 'error'))) THEN
      RAISE EXCEPTION 'removed organization domain update refused'
        USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_domains_guard';
    END IF;
    RETURN NEW;
  END IF;
  IF NOT ((OLD.state || '>' || NEW.state) = ANY (ARRAY[
       'pending>pending', 'pending>verified', 'pending>removed',
       'verified>verified', 'verified>active', 'verified>suspended', 'verified>removed',
       'active>active', 'active>verified', 'active>suspended', 'active>removed',
       'suspended>suspended', 'suspended>verified', 'suspended>active', 'suspended>removed'])) THEN
    RAISE EXCEPTION 'organization domain transition refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_domains_guard';
  END IF;
  IF (NEW.challenge_value IS DISTINCT FROM OLD.challenge_value OR NEW.challenge_expires_at IS DISTINCT FROM OLD.challenge_expires_at)
     AND NOT (OLD.state = 'pending' AND NEW.state = 'pending') THEN
    RAISE EXCEPTION 'organization domain challenge refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_domains_guard';
  END IF;
  IF NEW.verified_at IS DISTINCT FROM OLD.verified_at AND NOT (OLD.state = 'pending' AND NEW.state = 'verified') THEN
    RAISE EXCEPTION 'organization domain verification time refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_domains_guard';
  END IF;
  IF NEW.verified_until IS DISTINCT FROM OLD.verified_until AND NOT (
       (NEW.state IN ('verified', 'active') AND NEW.verified_until > now()
        AND (OLD.verified_until IS NULL OR NEW.verified_until > OLD.verified_until))
       OR (NEW.state = 'suspended' AND OLD.state IN ('verified', 'active', 'suspended') AND NEW.verified_until = now())) THEN
    RAISE EXCEPTION 'organization domain freshness refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_domains_guard';
  END IF;
  IF NEW.state = 'removed' AND (NEW.tls_state <> 'withdrawing' OR NEW.tls_checked_at IS DISTINCT FROM OLD.tls_checked_at
       OR NEW.tls_not_after IS DISTINCT FROM OLD.tls_not_after OR NEW.tls_reference IS DISTINCT FROM OLD.tls_reference
       OR NEW.last_check_at IS DISTINCT FROM OLD.last_check_at OR NEW.last_check_outcome IS DISTINCT FROM OLD.last_check_outcome
       OR NEW.check_failures IS DISTINCT FROM OLD.check_failures) THEN
    RAISE EXCEPTION 'organization domain removal refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_domains_guard';
  END IF;
  IF NEW.state = 'pending' AND (NEW.tls_state IS DISTINCT FROM OLD.tls_state OR NEW.tls_checked_at IS DISTINCT FROM OLD.tls_checked_at
       OR NEW.tls_not_after IS DISTINCT FROM OLD.tls_not_after OR NEW.tls_reference IS DISTINCT FROM OLD.tls_reference) THEN
    RAISE EXCEPTION 'pending organization domain TLS update refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_domains_guard';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

CREATE TRIGGER organization_domains_guard BEFORE UPDATE ON app.organization_domains FOR EACH ROW
  EXECUTE FUNCTION app.organization_domains_guard();

-- The slug guard (owner-owned SECURITY INVOKER): an update made in an org_admin context (which can pass only the
-- organizations_update_slug policy) may change the slug and nothing else; the whole row minus slug is compared, so
-- current and future Organization fields stay owner-only. Owners and context-free paths are unaffected.
-- +goose StatementBegin
CREATE FUNCTION app.organizations_slug_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF (SELECT app.rls_org_role()) = 'org_admin' AND (to_jsonb(NEW) - 'slug') IS DISTINCT FROM (to_jsonb(OLD) - 'slug') THEN
    RAISE EXCEPTION 'organization admins may change only the slug' USING ERRCODE = 'insufficient_privilege';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

CREATE TRIGGER organizations_slug_guard BEFORE UPDATE ON app.organizations FOR EACH ROW
  EXECUTE FUNCTION app.organizations_slug_guard();

-- The capture function (owner-owned SECURITY INVOKER): claims, state changes and checks, strictly atomic.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_organization_domains() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id)
    VALUES ('domain.claimed', 'success', NEW.organization_id, 'organization_domain', NEW.id);
  ELSIF TG_NAME = 'audit_capture_state' THEN
    INSERT INTO audit.events (action, outcome, reason, organization_id, resource_type, resource_id)
    VALUES ('domain.state_changed', 'success', NEW.state, NEW.organization_id, 'organization_domain', NEW.id);
  ELSE
    INSERT INTO audit.events (action, outcome, reason, organization_id, resource_type, resource_id)
    VALUES ('domain.checked', CASE WHEN NEW.last_check_outcome IN ('ok', 'withdrawn') THEN 'success' ELSE 'failed' END,
            NEW.last_check_outcome, NEW.organization_id, 'organization_domain', NEW.id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.organization_domains FOR EACH ROW
  EXECUTE FUNCTION audit.capture_organization_domains();
CREATE TRIGGER audit_capture_state AFTER UPDATE OF state ON app.organization_domains FOR EACH ROW
  WHEN (OLD.state IS DISTINCT FROM NEW.state) EXECUTE FUNCTION audit.capture_organization_domains();
CREATE TRIGGER audit_capture_check AFTER UPDATE OF last_check_at ON app.organization_domains FOR EACH ROW
  WHEN (OLD.last_check_at IS DISTINCT FROM NEW.last_check_at AND NEW.last_check_outcome IS NOT NULL)
  EXECUTE FUNCTION audit.capture_organization_domains();

-- Slug changes are captured with the existing Organization actions (the version-8 body otherwise unchanged).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit.capture_organizations() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id)
    VALUES ('organization.created', 'success', NEW.id, 'organization', NEW.id);
    RETURN NULL;
  END IF;
  IF NEW.name IS DISTINCT FROM OLD.name THEN
    INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id)
    VALUES ('organization.renamed', 'success', NEW.id, 'organization', NEW.id);
  END IF;
  IF NEW.slug IS DISTINCT FROM OLD.slug THEN
    INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id)
    VALUES ('organization.slug_changed', 'success', NEW.id, 'organization', NEW.id);
  END IF;
  IF NEW.group_admins_edit_business_profiles IS DISTINCT FROM OLD.group_admins_edit_business_profiles THEN
    INSERT INTO audit.events (action, outcome, reason, organization_id, resource_type, resource_id)
    VALUES ('organization.policy_changed', 'success',
            CASE WHEN NEW.group_admins_edit_business_profiles THEN 'group_admin_profile_edit_enabled'
                 ELSE 'group_admin_profile_edit_disabled' END,
            NEW.id, 'organization', NEW.id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

DROP TRIGGER audit_capture_update ON app.organizations;
CREATE TRIGGER audit_capture_update AFTER UPDATE OF name, slug, group_admins_edit_business_profiles ON app.organizations
  FOR EACH ROW
  WHEN (OLD.name IS DISTINCT FROM NEW.name OR OLD.slug IS DISTINCT FROM NEW.slug
        OR OLD.group_admins_edit_business_profiles IS DISTINCT FROM NEW.group_admins_edit_business_profiles)
  EXECUTE FUNCTION audit.capture_organizations();

-- The attribution and shape rules with the four domain actions, the domain denial operations and resource type and the
-- two sign-in reasons (the version-10 body otherwise unchanged; CREATE OR REPLACE keeps identity, owner, ACL, trigger).
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
    'object.deleted', 'mcp.connection_established', 'mcp.context_changed', 'organization.slug_changed', 'domain.claimed',
    'domain.state_changed', 'domain.checked');
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
    WHEN NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed',
                        'organization.slug_changed') THEN
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
    WHEN NEW.action IN ('domain.claimed', 'domain.state_changed', 'domain.checked') THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL AND NEW.group_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization_domain' AND NEW.resource_id IS NOT NULL
      AND CASE NEW.action
            WHEN 'domain.claimed' THEN NEW.outcome = 'success' AND NEW.reason IS NULL
            WHEN 'domain.state_changed' THEN NEW.outcome = 'success' AND NEW.reason IN ('verified', 'active', 'suspended', 'removed')
            ELSE (NEW.outcome = 'success' AND NEW.reason IN ('ok', 'withdrawn'))
              OR (NEW.outcome = 'failed' AND NEW.reason IN ('mismatch', 'nxdomain', 'transient', 'tls_error', 'withdraw_error',
                                                           'withdraw_unknown')) END
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
            'organization_not_member', 'host_changed'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object', 'organization_domain')))
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

-- The widened catalog constraints (the version-10 lists plus the G016 values).
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
                    'organization.slug_changed', 'domain.claimed', 'domain.state_changed', 'domain.checked',
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
                    'provider_unavailable', 'tombstone_failed', 'tombstone_unknown',
                    'verified', 'active', 'suspended', 'removed', 'ok', 'mismatch', 'nxdomain', 'transient', 'tls_error',
                    'withdrawn', 'withdraw_error', 'withdraw_unknown', 'organization_not_member', 'host_changed')),
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain'));


-- The readers, created by their owner (the auth reader, through the session's SET chain).
-- +goose StatementBegin
DO $do$
DECLARE
  db text := current_database();
  ow text := current_user;
  rt text := db || '_runtime';
  ar text := db || '_auth_reader';
BEGIN
  EXECUTE format('GRANT SELECT (id, slug) ON app.organizations TO %I', ar);
  EXECUTE format('GRANT SELECT (organization_id, hostname, generation, state, verified_until, tls_state, tls_not_after) '
              || 'ON app.organization_domains TO %I', ar);
  EXECUTE format('GRANT SELECT (organization_id, subject_user_id, scope_type, source) ON app.role_bindings TO %I', ar);
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $r$
CREATE FUNCTION app.hostname_state(p_kind text, p_key text)
  RETURNS TABLE (organization_id uuid, kind text, routing boolean, issuance boolean)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF p_kind = 'platform' THEN
    RETURN QUERY SELECT o.id, 'platform'::text, true, false FROM app.organizations o WHERE o.slug = (p_key COLLATE "C");
  ELSIF p_kind = 'custom' THEN
    RETURN QUERY
    SELECT d.organization_id, 'custom'::text,
           d.state = 'active' AND d.verified_until > now() AND d.tls_state = 'ready' AND d.tls_not_after > now(),
           d.state IN ('verified', 'active', 'suspended') AND d.verified_until > now()
    FROM app.organization_domains d WHERE d.hostname = (p_key COLLATE "C") AND d.state <> 'removed';
  END IF;
END $$
$r$;
  EXECUTE $r$
CREATE FUNCTION app.is_organization_member(p_user uuid, p_org uuid) RETURNS boolean
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN EXISTS (SELECT 1 FROM app.role_bindings b WHERE b.subject_user_id = p_user AND b.organization_id = p_org
                   AND b.scope_type = 'organization' AND b.source = 'membership');
END $$
$r$;
  REVOKE ALL ON FUNCTION app.hostname_state(text, text), app.is_organization_member(uuid, uuid) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.hostname_state(text, text), app.is_organization_member(uuid, uuid) TO %I', rt);
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
  O text := '(SELECT app.ctx_org())';
  R text := '(SELECT app.rls_org_role())';
  actual text;
  expected text;
BEGIN
  REVOKE ALL ON FUNCTION app.organization_domains_guard(), app.organizations_slug_guard(), audit.capture_organization_domains()
    FROM PUBLIC;

  -- Runtime: the slug column; domain rows read, a pending row inserted, the lifecycle columns updated. No DELETE.
  EXECUTE format('GRANT UPDATE (slug) ON app.organizations TO %I', rt::text);
  EXECUTE format('GRANT SELECT ON app.organization_domains TO %I', rt::text);
  EXECUTE format('GRANT INSERT (id, organization_id, hostname, generation, challenge_value, challenge_expires_at, created_by, '
              || 'created_at) ON app.organization_domains TO %I', rt::text);
  EXECUTE format('GRANT UPDATE (state, challenge_value, challenge_expires_at, verified_at, verified_until, last_check_at, '
              || 'last_check_outcome, check_failures, tls_state, tls_checked_at, tls_not_after, tls_reference, removed_at, version) '
              || 'ON app.organization_domains TO %I', rt::text);

  ALTER TABLE app.organization_domains ENABLE ROW LEVEL SECURITY;
  ALTER TABLE app.organization_domains FORCE ROW LEVEL SECURITY;
  EXECUTE format('CREATE POLICY organization_domains_select ON app.organization_domains FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND %s IN (''org_owner'', ''org_admin''))', rt::text, ow::text, O, R);
  EXECUTE format('CREATE POLICY organization_domains_insert ON app.organization_domains FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IN (''org_owner'', ''org_admin'') AND state = ''pending'' AND version = 0 '
    || 'AND generation >= 1)', rt::text, ow::text, O, R);
  EXECUTE format('CREATE POLICY organization_domains_update ON app.organization_domains FOR UPDATE TO %I, %I USING ('
    || 'organization_id = %s AND %s IN (''org_owner'', ''org_admin'')) WITH CHECK ('
    || 'organization_id = %s AND %s IN (''org_owner'', ''org_admin''))', rt::text, ow::text, O, R, O, R);
  EXECUTE format('CREATE POLICY organization_domains_auth_reader ON app.organization_domains FOR SELECT TO %I USING (true)', ar::text);
  EXECUTE format('CREATE POLICY organizations_auth_reader ON app.organizations FOR SELECT TO %I USING (true)', ar::text);
  -- Deviation #1 closure: the owner/admin slug write; organizations_update and the column grants are unchanged and the
  -- slug guard confines an admin to the slug.
  EXECUTE format('CREATE POLICY organizations_update_slug ON app.organizations FOR UPDATE TO %I, %I '
    || 'USING (id = %s AND %s IN (''org_owner'', ''org_admin'')) WITH CHECK (id = %s AND %s IN (''org_owner'', ''org_admin''))',
    rt::text, ow::text, O, R, O, R);
  EXECUTE format('CREATE POLICY role_bindings_auth_reader ON app.role_bindings FOR SELECT TO %I USING (true)', ar::text);

  -- Postconditions: the exact version-11 inventories.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  IF actual IS DISTINCT FROM 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,mcp_connections+force,organization_domains+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force' THEN
    RAISE EXCEPTION 'app row-level security tables are %', actual;
  END IF;
  -- The 48 application relations (the version-10 42 plus organization_domains, its primary key, its hostname/generation
  -- key, its live-hostname and Organization indexes, and organizations_slug_key) plus the migration ledger's table,
  -- identity sequence and primary key.
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace;
  IF actual <> '51' THEN
    RAISE EXCEPTION 'app relation count is %, expected 51 (48 plus the migration ledger)', actual;
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
    || 'organization_domains_auth_reader:organization_domains:r:true:ar;organization_domains_insert:organization_domains:a:true:ow+rt;'
    || 'organization_domains_select:organization_domains:r:true:ow+rt;organization_domains_update:organization_domains:w:true:ow+rt;'
    || 'organizations_auth_reader:organizations:r:true:ar;'
    || 'organizations_select:organizations:r:true:ow+rt;organizations_update:organizations:w:true:ow+rt;'
    || 'organizations_update_slug:organizations:w:true:ow+rt;'
    || 'role_bindings_auth_reader:role_bindings:r:true:ar;role_bindings_lock:role_bindings:w:true:ow+rt;role_bindings_policy_reader:role_bindings:r:true:pr;'
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
    || 'organization_domains:rt:SELECT;organization_domains.challenge_expires_at:rt:INSERT,UPDATE;'
    || 'organization_domains.challenge_value:rt:INSERT,UPDATE;organization_domains.check_failures:rt:UPDATE;'
    || 'organization_domains.created_at:rt:INSERT;organization_domains.created_by:rt:INSERT;organization_domains.generation:ar:SELECT;organization_domains.generation:rt:INSERT;'
    || 'organization_domains.hostname:ar:SELECT;organization_domains.hostname:rt:INSERT;organization_domains.id:rt:INSERT;'
    || 'organization_domains.last_check_at:rt:UPDATE;organization_domains.last_check_outcome:rt:UPDATE;'
    || 'organization_domains.organization_id:ar:SELECT;organization_domains.organization_id:rt:INSERT;'
    || 'organization_domains.removed_at:rt:UPDATE;organization_domains.state:ar:SELECT;organization_domains.state:rt:UPDATE;'
    || 'organization_domains.tls_checked_at:rt:UPDATE;organization_domains.tls_not_after:ar:SELECT;'
    || 'organization_domains.tls_not_after:rt:UPDATE;organization_domains.tls_reference:rt:UPDATE;'
    || 'organization_domains.tls_state:ar:SELECT;organization_domains.tls_state:rt:UPDATE;organization_domains.verified_at:rt:UPDATE;'
    || 'organization_domains.verified_until:ar:SELECT;organization_domains.verified_until:rt:UPDATE;organization_domains.version:rt:UPDATE;'
    || 'organization_memberships:rt:SELECT;organizations:rt:SELECT;organizations.group_admins_edit_business_profiles:rt:UPDATE;'
    || 'organizations.id:ar:SELECT;organizations.slug:ar:SELECT;organizations.slug:rt:UPDATE;'
    || 'role_bindings.group_id:pr:SELECT;role_bindings.group_id:rt:SELECT;role_bindings.organization_id:ar:SELECT;role_bindings.organization_id:pr:SELECT;'
    || 'role_bindings.organization_id:rt:SELECT;role_bindings.role:pr:SELECT;role_bindings.role:rt:SELECT,UPDATE;'
    || 'role_bindings.scope_id:pr:SELECT;role_bindings.scope_id:rt:SELECT;role_bindings.scope_type:ar:SELECT;role_bindings.scope_type:pr:SELECT;'
    || 'role_bindings.scope_type:rt:SELECT;role_bindings.source:ar:SELECT;role_bindings.source:pr:SELECT;role_bindings.source:rt:SELECT;'
    || 'role_bindings.subject_user_id:ar:SELECT;role_bindings.subject_user_id:pr:SELECT;role_bindings.subject_user_id:rt:SELECT;sessions:rt:INSERT,SELECT;'
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
    || 'hostname_state:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;'
    || 'is_organization_member:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;'
    || 'mcp_connections_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'organization_domains_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'organizations_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'organizations_slug_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
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
    || 'capture_organization_domains:ow:false:false:v:search_path=pg_catalog, pg_temp:-;'
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
  -- The app trigger inventory: the 23 capture triggers plus the existing guards, the connection guard and the domain guard.
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
    || 'organization_domains.audit_capture_check:17:O:audit.capture_organization_domains:when;'
    || 'organization_domains.audit_capture_insert:5:O:audit.capture_organization_domains:-;'
    || 'organization_domains.audit_capture_state:17:O:audit.capture_organization_domains:when;'
    || 'organization_domains.organization_domains_guard:19:O:app.organization_domains_guard:-;'
    || 'organizations.audit_capture_insert:5:O:audit.capture_organizations:-;'
    || 'organizations.audit_capture_update:17:O:audit.capture_organizations:when;'
    || 'organizations.organizations_keys_immutable:19:O:app.organizations_keys_immutable:-;'
    || 'organizations.organizations_slug_guard:19:O:app.organizations_slug_guard:-;'
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
  rt text := db || '_runtime';
  ar text := db || '_auth_reader';
BEGIN
  -- Without FORCE the owner's reads are not filtered by row-level security. Domain state and history are never silently
  -- discarded: refuse while any exists (the refusal rolls back the NO FORCE with everything else).
  ALTER TABLE app.organization_domains NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.organizations NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.login_transactions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  IF EXISTS (SELECT 1 FROM app.organization_domains) OR EXISTS (SELECT 1 FROM app.organizations WHERE slug IS NOT NULL)
     OR EXISTS (SELECT 1 FROM app.login_transactions WHERE organization_id IS NOT NULL OR callback_origin IS NOT NULL) THEN
    RAISE EXCEPTION 'refusing to roll back 00011: domains, slugs or tenant-host login transactions exist';
  END IF;
  IF EXISTS (SELECT 1 FROM audit.events WHERE action IN ('organization.slug_changed', 'domain.claimed', 'domain.state_changed',
                                                         'domain.checked')
               OR reason IN ('verified', 'active', 'suspended', 'removed', 'ok', 'mismatch', 'nxdomain', 'transient',
                             'tls_error', 'withdrawn', 'withdraw_error', 'withdraw_unknown', 'organization_not_member',
                             'host_changed')
               OR operation IN ('domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck',
                                'domain.remove', 'domain.retry_withdrawal', 'organization.set_slug')
               OR resource_type = 'organization_domain') THEN
    RAISE EXCEPTION 'refusing to roll back 00011: audit.events holds domain history (history is never discarded)';
  END IF;
  ALTER TABLE app.organizations FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.login_transactions FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  -- The readers, dropped by their owner.
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.hostname_state(text, text);
  DROP FUNCTION app.is_organization_member(uuid, uuid);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  DROP POLICY organizations_auth_reader ON app.organizations;
  DROP POLICY organizations_update_slug ON app.organizations;
  DROP TRIGGER organizations_slug_guard ON app.organizations;
  DROP FUNCTION app.organizations_slug_guard();
  DROP POLICY role_bindings_auth_reader ON app.role_bindings;
  EXECUTE format('REVOKE SELECT (id, slug) ON app.organizations FROM %I', ar);
  EXECUTE format('REVOKE SELECT (organization_id, subject_user_id, scope_type, source) ON app.role_bindings FROM %I', ar);
  EXECUTE format('REVOKE UPDATE (slug) ON app.organizations FROM %I', rt);
  DROP TABLE app.organization_domains;
  DROP FUNCTION app.organization_domains_guard();
  DROP FUNCTION audit.capture_organization_domains();
  DROP TRIGGER audit_capture_update ON app.organizations;
  ALTER TABLE app.organizations DROP COLUMN slug;
  ALTER TABLE app.login_transactions DROP COLUMN organization_id, DROP COLUMN callback_origin;
END
$do$;
-- +goose StatementEnd

-- The exact version-8 capture body and update trigger of app.organizations (00008).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit.capture_organizations() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id)
    VALUES ('organization.created', 'success', NEW.id, 'organization', NEW.id);
    RETURN NULL;
  END IF;
  IF NEW.name IS DISTINCT FROM OLD.name THEN
    INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id)
    VALUES ('organization.renamed', 'success', NEW.id, 'organization', NEW.id);
  END IF;
  IF NEW.group_admins_edit_business_profiles IS DISTINCT FROM OLD.group_admins_edit_business_profiles THEN
    INSERT INTO audit.events (action, outcome, reason, organization_id, resource_type, resource_id)
    VALUES ('organization.policy_changed', 'success',
            CASE WHEN NEW.group_admins_edit_business_profiles THEN 'group_admin_profile_edit_enabled'
                 ELSE 'group_admin_profile_edit_disabled' END,
            NEW.id, 'organization', NEW.id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

CREATE TRIGGER audit_capture_update AFTER UPDATE OF name, group_admins_edit_business_profiles ON app.organizations FOR EACH ROW
  WHEN (OLD.name IS DISTINCT FROM NEW.name OR OLD.group_admins_edit_business_profiles IS DISTINCT FROM NEW.group_admins_edit_business_profiles)
  EXECUTE FUNCTION audit.capture_organizations();

-- The exact version-10 catalog constraints (00010).
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


-- The exact version-10 attribution body (00010).
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
