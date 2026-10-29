-- Group onboarding (G018 Plan r5 §4; R0 addendum; audit-operation amendment
-- management/v0/results/G018-audit-operation-amendment-codex-review.md; docs/foundation/TENANCY.md,
-- AUTHORIZATION.md, DATABASE-ISOLATION.md, AUDIT.md).
--
-- One goose transaction:
--   - Runtime INSERT only, column-limited, on app.groups, app.role_bindings, app.business_profiles and
--     app.group_business_profiles. No DELETE, no UPDATE change, no TRIGGER privilege.
--   - Insertion provenance on app.business_profiles: created_xid (xid8, DEFAULT pg_current_xact_id()) and created_sysid
--     (bigint, DEFAULT the cluster's system identifier), assigned once by their defaults, in no runtime grant, immutable
--     by the replaced app.business_profiles_keys_immutable() (owner included).
--   - Creator bootstrap: the SECURITY INVOKER AFTER INSERT trigger groups_bootstrap_owner writes the context subject's
--     group_owner binding in the statement that creates the Group; role_bindings_insert_bootstrap admits a binding only
--     from inside a trigger (pg_trigger_depth() > 0) with that fixed shape.
--   - Four FOR INSERT policies (TO runtime and owner) and the boolean SECURITY DEFINER reader app.rls_profile_eligible
--     (owned by <db>_policy_reader, two helper policies limited to ctx_org()), which checks its own context first.
--   - audit.events.events_operation_valid gains exactly the seven G018 access.denied operations (all 29 version-13
--     operations kept). One ALTER TABLE drops and re-adds the constraint: ACCESS EXCLUSIVE on audit.events for the rest of
--     the migration transaction and one validating scan of audit.events; a failure rolls back everything, so no
--     permissive gap is visible outside the transaction. Every other audit list, the attribution trigger, grants and
--     failure behaviour are unchanged.
--
-- Locks: ADD COLUMN with two volatile-free defaults rewrites the small app.business_profiles table once (ACCESS
-- EXCLUSIVE); grant, policy and trigger DDL on the four tenancy tables take brief ACCESS EXCLUSIVE locks. No data
-- migration beyond the column defaults.
--
-- Down (development and test only) refuses while audit.events holds any row with a G018 operation (inspected across all
-- tenants with FORCE lifted inside the transaction; a refusal rolls the lift back), then restores the version-13 shapes.

-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regclass('app.organization_branding') IS NULL
     OR to_regprocedure('app.groups_bootstrap_owner()') IS NOT NULL
     OR to_regprocedure('app.rls_profile_eligible(uuid)') IS NOT NULL
     OR EXISTS (SELECT 1 FROM pg_policy WHERE polname IN ('groups_insert', 'role_bindings_insert_bootstrap',
                  'business_profiles_insert', 'group_business_profiles_insert', 'business_profiles_policy_reader',
                  'group_business_profiles_policy_reader'))
     OR EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'groups_bootstrap_owner') THEN
    RAISE EXCEPTION 'version-13 precondition: onboarding objects already exist or the version-13 shapes are missing';
  END IF;
  IF has_table_privilege(rt, 'app.groups', 'INSERT') OR has_table_privilege(rt, 'app.role_bindings', 'INSERT')
     OR has_table_privilege(rt, 'app.business_profiles', 'INSERT') OR has_table_privilege(rt, 'app.group_business_profiles', 'INSERT')
     OR has_any_column_privilege(rt, 'app.groups', 'INSERT') OR has_any_column_privilege(rt, 'app.role_bindings', 'INSERT')
     OR has_any_column_privilege(rt, 'app.business_profiles', 'INSERT')
     OR has_any_column_privilege(rt, 'app.group_business_profiles', 'INSERT')
     OR has_any_column_privilege(pr, 'app.business_profiles', 'SELECT')
     OR has_any_column_privilege(pr, 'app.group_business_profiles', 'SELECT') THEN
    RAISE EXCEPTION 'version-13 precondition: runtime INSERT or policy-reader SELECT on the onboarding tables already exists';
  END IF;
  IF (SELECT string_agg(attname, ',' ORDER BY attnum) FROM pg_attribute
      WHERE attrelid = 'app.business_profiles'::regclass AND attnum > 0 AND NOT attisdropped)
     IS DISTINCT FROM 'id,organization_id,display_name,legal_name,description,industry,website,created_at' THEN
    RAISE EXCEPTION 'version-13 precondition: app.business_profiles columns differ from the 00004 shape';
  END IF;
  IF NOT has_function_privilege(rt, 'pg_catalog.pg_control_system()', 'EXECUTE')
     OR NOT has_function_privilege(pr, 'pg_catalog.pg_control_system()', 'EXECUTE')
     OR NOT has_function_privilege(rt, 'pg_catalog.pg_current_xact_id()', 'EXECUTE')
     OR NOT has_function_privilege(pr, 'pg_catalog.pg_current_xact_id()', 'EXECUTE') THEN
    RAISE EXCEPTION 'precondition: pg_control_system() and pg_current_xact_id() must be executable by the runtime and policy-reader roles';
  END IF;
END
$$;
-- +goose StatementEnd

-- Insertion provenance (r3/r4). Existing rows receive this migration transaction's id (never current again) and the
-- migrating cluster's identifier.
ALTER TABLE app.business_profiles
  ADD COLUMN created_xid xid8 NOT NULL DEFAULT pg_current_xact_id(),
  ADD COLUMN created_sysid bigint NOT NULL DEFAULT (pg_control_system()).system_identifier;

-- Key and provenance immutability (replaces the 00004 body; same owner, attributes and trigger).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.business_profiles_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
  IF (NEW.id, NEW.organization_id, NEW.created_xid, NEW.created_sysid)
     IS DISTINCT FROM (OLD.id, OLD.organization_id, OLD.created_xid, OLD.created_sysid) THEN
    RAISE EXCEPTION 'business_profiles key columns are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'business_profiles_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- The G018 access.denied operations (amendment): the 29 version-13 operations plus exactly seven.
ALTER TABLE audit.events DROP CONSTRAINT events_operation_valid,
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug',
                    'branding.get', 'branding.update', 'branding.reset', 'branding.set_logo', 'branding.remove_logo',
                    'group.suggest_slug', 'group.create', 'group.list', 'group.onboarding.get',
                    'group.onboarding.complete_new', 'group.onboarding.complete_existing', 'business_profile.list'));

-- +goose StatementBegin
DO $do$
DECLARE
  db text := current_database();
  ow text := current_user;
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  MGR text := '(SELECT app.rls_org_role()) IN (''org_owner'', ''org_admin'')';
BEGIN
  -- Creator bootstrap trigger function: SECURITY INVOKER, owner-owned; the role names are fixed at migration time.
  -- The superuser seed and any other role keep their explicit binding inserts (the trigger does nothing for them).
  EXECUTE format($f$
CREATE FUNCTION app.groups_bootstrap_owner() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $b$
BEGIN
  IF current_user::text NOT IN (%L, %L) OR app.ctx_user() IS NULL THEN
    RETURN NULL;
  END IF;
  INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
  VALUES (NEW.organization_id, app.ctx_user(), 'group', NEW.id, NEW.id, 'group_owner', 'membership');
  RETURN NULL;
END $b$
$f$, rt, ow);
  REVOKE ALL ON FUNCTION app.groups_bootstrap_owner() FROM PUBLIC;
  CREATE TRIGGER groups_bootstrap_owner AFTER INSERT ON app.groups FOR EACH ROW EXECUTE FUNCTION app.groups_bootstrap_owner();

  -- Runtime: INSERT only, column-limited (ids of groups and business_profiles are client-generated: keyed inserts
  -- without RETURNING; binding and association keys and every created_at stay database-generated).
  EXECUTE format('GRANT INSERT (id, organization_id, name, slug) ON app.groups TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) '
              || 'ON app.role_bindings TO %I', rt);
  EXECUTE format('GRANT INSERT (id, organization_id, display_name, legal_name, description, industry, website) '
              || 'ON app.business_profiles TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, group_id, business_profile_id) ON app.group_business_profiles TO %I', rt);

  -- Policy reader: exactly the columns the eligibility reader needs.
  EXECUTE format('GRANT SELECT (id, organization_id, created_xid, created_sysid) ON app.business_profiles TO %I', pr);
  EXECUTE format('GRANT SELECT (organization_id, group_id, business_profile_id) ON app.group_business_profiles TO %I', pr);
  EXECUTE format('CREATE POLICY business_profiles_policy_reader ON app.business_profiles FOR SELECT TO %I USING ('
    || 'organization_id = app.ctx_org())', pr);
  EXECUTE format('CREATE POLICY group_business_profiles_policy_reader ON app.group_business_profiles FOR SELECT TO %I USING ('
    || 'organization_id = app.ctx_org())', pr);

  -- The eligibility reader, created as the policy reader (00006 SET chain). Context first: no valid current context,
  -- false before any row is read. Then (a) inserted by this transaction in this cluster, or (b) attached to a Group on
  -- which the context subject holds group_owner/group_admin (selected-Group context: that Group only).
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pr);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  EXECUTE $r$
CREATE FUNCTION app.rls_profile_eligible(p_profile uuid) RETURNS boolean
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text := app.rls_org_role();
BEGIN
  IF u IS NULL OR o IS NULL OR r IS NULL OR p_profile IS NULL THEN
    RETURN false;
  END IF;
  IF EXISTS (SELECT 1 FROM app.business_profiles bp
             WHERE bp.id = p_profile AND bp.organization_id = o AND bp.created_xid = pg_current_xact_id()
               AND bp.created_sysid = (pg_control_system()).system_identifier) THEN
    RETURN true;
  END IF;
  RETURN EXISTS (
    SELECT 1 FROM app.group_business_profiles a
    JOIN app.role_bindings rb ON rb.organization_id = a.organization_id AND rb.scope_type = 'group'
      AND rb.scope_id = a.group_id AND rb.source = 'membership'
    WHERE a.business_profile_id = p_profile AND a.organization_id = o AND rb.subject_user_id = u
      AND rb.role IN ('group_owner', 'group_admin') AND (g IS NULL OR a.group_id = g));
END $$
$r$;
  REVOKE ALL ON FUNCTION app.rls_profile_eligible(uuid) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_profile_eligible(uuid) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pr);

  -- The four INSERT policies (TO runtime and owner; FORCE binds the owner to the same contexts).
  EXECUTE format('CREATE POLICY groups_insert ON app.groups FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NULL AND %s IS NOT NULL)', rt, ow, O, G, R);
  EXECUTE format('CREATE POLICY role_bindings_insert_bootstrap ON app.role_bindings FOR INSERT TO %I, %I WITH CHECK ('
    || 'pg_trigger_depth() > 0 AND organization_id = %s AND %s IS NULL AND %s IS NOT NULL AND subject_user_id = %s '
    || 'AND scope_type = ''group'' AND group_id = scope_id AND role = ''group_owner'' AND source = ''membership'')',
    rt, ow, O, G, R, U);
  EXECUTE format('CREATE POLICY business_profiles_insert ON app.business_profiles FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s OR EXISTS (SELECT 1 FROM app.role_bindings rb '
    || 'WHERE rb.subject_user_id = %s AND rb.organization_id = %s AND rb.scope_type = ''group'' '
    || 'AND rb.source = ''membership'' AND rb.role IN (''group_owner'', ''group_admin'') AND (%s IS NULL OR rb.scope_id = %s))))',
    rt, ow, O, R, MGR, U, O, G, G);
  EXECUTE format('CREATE POLICY group_business_profiles_insert ON app.group_business_profiles FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR group_id = %s) AND (%s OR ('
    || 'app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'') AND app.rls_profile_eligible(business_profile_id))))',
    rt, ow, O, R, G, G, MGR);
END
$do$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
DECLARE
  ow regrole := current_user::regrole;
  rt regrole := (current_database() || '_runtime')::regrole;
  pr regrole := (current_database() || '_policy_reader')::regrole;
  ar regrole := (current_database() || '_auth_reader')::regrole;
  actual text;
BEGIN
  -- Postcondition 1: every ACL entry (other than the owner's) of the four tables and their columns.
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr' WHEN ar THEN 'ar'
                                          ELSE a.grantee::regrole::text END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.oid IN ('app.groups'::regclass, 'app.role_bindings'::regclass, 'app.business_profiles'::regclass,
                    'app.group_business_profiles'::regclass) AND a.grantee <> c.relowner
    GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname || '.' || at.attname, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                                          WHEN ar THEN 'ar' ELSE a.grantee::regrole::text END,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
    FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.oid IN ('app.groups'::regclass, 'app.role_bindings'::regclass, 'app.business_profiles'::regclass,
                    'app.group_business_profiles'::regclass)
    GROUP BY c.relname, at.attname, a.grantee) t;
  IF actual IS DISTINCT FROM
       'business_profiles:rt:SELECT;business_profiles.created_sysid:pr:SELECT;business_profiles.created_xid:pr:SELECT;'
    || 'business_profiles.description:rt:INSERT,UPDATE;business_profiles.display_name:rt:INSERT,UPDATE;'
    || 'business_profiles.id:pr:SELECT;business_profiles.id:rt:INSERT;business_profiles.industry:rt:INSERT,UPDATE;'
    || 'business_profiles.legal_name:rt:INSERT,UPDATE;business_profiles.organization_id:pr:SELECT;'
    || 'business_profiles.organization_id:rt:INSERT;business_profiles.website:rt:INSERT,UPDATE;'
    || 'group_business_profiles:rt:SELECT;group_business_profiles.business_profile_id:pr:SELECT;'
    || 'group_business_profiles.business_profile_id:rt:INSERT;group_business_profiles.group_id:pr:SELECT;'
    || 'group_business_profiles.group_id:rt:INSERT;group_business_profiles.organization_id:pr:SELECT;'
    || 'group_business_profiles.organization_id:rt:INSERT;'
    || 'groups:rt:SELECT;groups.id:pr:SELECT;groups.id:rt:INSERT;groups.name:rt:INSERT,UPDATE;'
    || 'groups.organization_id:pr:SELECT;groups.organization_id:rt:INSERT;groups.slug:rt:INSERT,UPDATE;'
    || 'role_bindings.group_id:pr:SELECT;role_bindings.group_id:rt:INSERT,SELECT;'
    || 'role_bindings.organization_id:ar:SELECT;role_bindings.organization_id:pr:SELECT;role_bindings.organization_id:rt:INSERT,SELECT;'
    || 'role_bindings.role:pr:SELECT;role_bindings.role:rt:INSERT,SELECT,UPDATE;'
    || 'role_bindings.scope_id:pr:SELECT;role_bindings.scope_id:rt:INSERT,SELECT;'
    || 'role_bindings.scope_type:ar:SELECT;role_bindings.scope_type:pr:SELECT;role_bindings.scope_type:rt:INSERT,SELECT;'
    || 'role_bindings.source:ar:SELECT;role_bindings.source:pr:SELECT;role_bindings.source:rt:INSERT,SELECT;'
    || 'role_bindings.subject_user_id:ar:SELECT;role_bindings.subject_user_id:pr:SELECT;role_bindings.subject_user_id:rt:INSERT,SELECT' THEN
    RAISE EXCEPTION 'version-14 onboarding ACLs are %', actual;
  END IF;

  -- Postcondition 2: the six new policies (name, command, roles).
  SELECT string_agg(format('%s:%s:%s', p.polname, p.polcmd,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt THEN 'rt' WHEN ow THEN 'ow' WHEN pr THEN 'pr' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p WHERE p.polname IN ('groups_insert', 'role_bindings_insert_bootstrap', 'business_profiles_insert',
    'group_business_profiles_insert', 'business_profiles_policy_reader', 'group_business_profiles_policy_reader');
  IF actual IS DISTINCT FROM 'business_profiles_insert:a:ow+rt;business_profiles_policy_reader:r:pr;'
    || 'group_business_profiles_insert:a:ow+rt;group_business_profiles_policy_reader:r:pr;groups_insert:a:ow+rt;'
    || 'role_bindings_insert_bootstrap:a:ow+rt' THEN
    RAISE EXCEPTION 'onboarding policies are %', actual;
  END IF;

  -- Postcondition 3: the two new functions (owner, SECURITY DEFINER, volatility, configuration, ACL).
  SELECT string_agg(p.proname || ':' || CASE p.proowner WHEN ow THEN 'ow' WHEN pr THEN 'pr' ELSE 'other' END || ':'
           || p.prosecdef::text || ':' || p.provolatile::text || ':' || coalesce(array_to_string(p.proconfig, '|'), '-') || ':'
           || coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                          SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN ow THEN 'ow' ELSE 'other' END
                                 || '=' || a.privilege_type AS v
                          FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'), ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.oid IN ('app.groups_bootstrap_owner()'::regprocedure, 'app.rls_profile_eligible(uuid)'::regprocedure);
  IF actual IS DISTINCT FROM 'groups_bootstrap_owner:ow:false:v:search_path=pg_catalog:-;'
    || 'rls_profile_eligible:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE' THEN
    RAISE EXCEPTION 'onboarding functions are %', actual;
  END IF;

  -- Postcondition 4: the triggers of app.groups.
  SELECT string_agg(tgname, ',' ORDER BY tgname) INTO actual FROM pg_trigger
  WHERE tgrelid = 'app.groups'::regclass AND NOT tgisinternal;
  IF actual IS DISTINCT FROM 'audit_capture_insert,audit_capture_update,groups_bootstrap_owner,groups_keys_immutable' THEN
    RAISE EXCEPTION 'app.groups triggers are %', actual;
  END IF;

  -- Postcondition 5: the ten business_profiles columns, the provenance defaults, every row carrying this cluster's id,
  -- and the replaced immutability body.
  SELECT string_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod) || CASE WHEN a.attnotnull THEN ' nn' ELSE '' END,
           ',' ORDER BY a.attnum) INTO actual
  FROM pg_attribute a WHERE a.attrelid = 'app.business_profiles'::regclass AND a.attnum > 0 AND NOT a.attisdropped
    AND a.attname IN ('created_xid', 'created_sysid');
  IF actual IS DISTINCT FROM 'created_xid xid8 nn,created_sysid bigint nn'
     OR (SELECT count(*) FROM pg_attribute WHERE attrelid = 'app.business_profiles'::regclass AND attnum > 0
         AND NOT attisdropped) <> 10
     OR (SELECT pg_get_expr(d.adbin, d.adrelid) FROM pg_attrdef d JOIN pg_attribute a ON a.attrelid = d.adrelid
         AND a.attnum = d.adnum WHERE d.adrelid = 'app.business_profiles'::regclass AND a.attname = 'created_xid')
        IS DISTINCT FROM 'pg_current_xact_id()'
     OR (SELECT pg_get_expr(d.adbin, d.adrelid) FROM pg_attrdef d JOIN pg_attribute a ON a.attrelid = d.adrelid
         AND a.attnum = d.adnum WHERE d.adrelid = 'app.business_profiles'::regclass AND a.attname = 'created_sysid')
        IS DISTINCT FROM '(pg_control_system()).system_identifier' THEN
    RAISE EXCEPTION 'business_profiles provenance columns are %', actual;
  END IF;
  IF position('created_sysid' IN (SELECT prosrc FROM pg_proc WHERE oid = 'app.business_profiles_keys_immutable()'::regprocedure)) = 0 THEN
    RAISE EXCEPTION 'business_profiles_keys_immutable body was not replaced';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
  ow text := current_user;
BEGIN
  -- History is never discarded: refuse while any audit row (any tenant) names a G018 operation. FORCE is lifted only
  -- inside this transaction; a refusal rolls the lift back with everything else.
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  IF EXISTS (SELECT 1 FROM audit.events WHERE operation IN ('group.suggest_slug', 'group.create', 'group.list',
               'group.onboarding.get', 'group.onboarding.complete_new', 'group.onboarding.complete_existing',
               'business_profile.list')) THEN
    RAISE EXCEPTION 'refusing to roll back 00014: audit.events holds onboarding history (history is never discarded)';
  END IF;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  DROP POLICY groups_insert ON app.groups;
  DROP POLICY role_bindings_insert_bootstrap ON app.role_bindings;
  DROP POLICY business_profiles_insert ON app.business_profiles;
  DROP POLICY group_business_profiles_insert ON app.group_business_profiles;
  DROP POLICY business_profiles_policy_reader ON app.business_profiles;
  DROP POLICY group_business_profiles_policy_reader ON app.group_business_profiles;
  EXECUTE format('SET LOCAL ROLE %I', pr);
  DROP FUNCTION app.rls_profile_eligible(uuid);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  DROP TRIGGER groups_bootstrap_owner ON app.groups;
  DROP FUNCTION app.groups_bootstrap_owner();
  EXECUTE format('REVOKE INSERT (id, organization_id, name, slug) ON app.groups FROM %I', rt);
  EXECUTE format('REVOKE INSERT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) '
              || 'ON app.role_bindings FROM %I', rt);
  EXECUTE format('REVOKE INSERT (id, organization_id, display_name, legal_name, description, industry, website) '
              || 'ON app.business_profiles FROM %I', rt);
  EXECUTE format('REVOKE INSERT (organization_id, group_id, business_profile_id) ON app.group_business_profiles FROM %I', rt);
  EXECUTE format('REVOKE SELECT (id, organization_id, created_xid, created_sysid) ON app.business_profiles FROM %I', pr);
  EXECUTE format('REVOKE SELECT (organization_id, group_id, business_profile_id) ON app.group_business_profiles FROM %I', pr);
END
$do$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.business_profiles_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
  IF (NEW.id, NEW.organization_id) IS DISTINCT FROM (OLD.id, OLD.organization_id) THEN
    RAISE EXCEPTION 'business_profiles key columns are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'business_profiles_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

ALTER TABLE app.business_profiles DROP COLUMN created_sysid, DROP COLUMN created_xid;

-- The version-13 operation constraint.
ALTER TABLE audit.events DROP CONSTRAINT events_operation_valid,
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug',
                    'branding.get', 'branding.update', 'branding.reset', 'branding.set_logo', 'branding.remove_logo'));
