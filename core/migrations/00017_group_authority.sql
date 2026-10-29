-- Explicit Group authority and the owner-controlled Group-management override (G060 Plan r0 §3 as amended by the r1
-- addendum C1/C3/C4/C5 and correction02 C02-1/C02-2; ADR-008; docs/foundation/AUTHORIZATION.md, DATABASE-ISOLATION.md,
-- AUDIT.md).
--
-- Up, in one transaction:
--   - Precondition: owner, the three NOLOGIN NOINHERIT helper roles, the exact version-16 shapes, no 00017 object.
--   - app.organizations.group_management_override boolean NOT NULL DEFAULT false: every existing and new Organization
--     reads false; nothing is inferred from earlier broad administration (no backfill). Column grants: runtime UPDATE,
--     claim writer SELECT, policy reader SELECT (correction02 C02-1). No new organizations policy: the runtime's
--     table-level SELECT and the existing permissive UPDATE policies admit the column, the claim writer reads it under
--     organizations_claim_select and locks under organizations_claim_lock, the policy reader reads it under
--     organizations_policy_reader (all 00016).
--   - The owner-owned SECURITY INVOKER guard app.organizations_group_override_guard (BEFORE UPDATE OF the column): a
--     session-role change only by the current, enabled Organization owner (the permissive slug/enrollment UPDATE policies
--     would otherwise admit an administrator; amendment01 A2: the actor's users row FOR SHARE, status 'active', coherent
--     with DisableUser), raising group_override_owner.
--   - OVR, the inline current setting of the context Organization (the business_profiles_update precedent; no new
--     reader function): Organization staff manage a Group only when OVR is true; Group content never follows from it.
--     Re-created policies: role_bindings_select/update/delete (amendment01 A1.3: the Group branch admits Organization
--     staff only with OVR, or for select/delete and update USING (attempt3: the cascade's row lock) while the schema's own
--     trigger bodies run, pg_trigger_depth() >= 1, i.e. the guard's Organization-departure cascade; WITH CHECK of update:
--     the override, and no role change of the actor's own binding through the Organization path), groups_update, group_business_profiles_insert, business_profiles_update,
--     business_profiles_insert, stored_objects_select/insert/update (Group objects: an explicit Group role only, under
--     either setting), invitations_select/insert/update.
--   - Replaced bodies (identity, owner and ACL kept): policy reader rls_owner_count and rls_group_invitation_candidates;
--     claim writer rls_user_enabled_lock and invitation_claim_group (r1 C3: an issuer never consumes its own Group
--     invitation; r1 C4: the Organization row FOR SHARE, the issuer's and recipient's user rows FOR SHARE in ascending
--     id, then the bindings; the issuer must be enabled with current Group owner/admin authority of the scope or
--     Organization owner/admin authority while OVR is true); the owner-owned role_bindings_authority_guard gains step
--     (3b) (correction02 C02-2: a runtime or owner session never changes the role of its own Group binding unless it is
--     that Group's owner; the claim writer's validated upgrade path is not affected) and its step (3) cascades an
--     Organization-binding DELETE to the subject's Group bindings in ascending group_id inside the same statement
--     (amendment01 A1.3; owner_subject for another subject's Group-owner binding), then keeps the dependent check.
--   - Audit: events_reason_valid +2 (group_management_override_enabled/disabled), events_operation_valid +1
--     (organization.set_group_override), events_action_valid +1 core action (access.override_used: the actual mutated
--     Group or Business Profile of an override-based administrative write, appended atomically by core);
--     audit.events_attribute admits both; audit.capture_organizations records organization.policy_changed when the
--     override changes and audit_capture_update watches the column.
--   - Postconditions.
--
-- Locks: brief ACCESS EXCLUSIVE locks for the column, grant, policy and trigger DDL on organizations, groups,
-- role_bindings, business_profiles, group_business_profiles, stored_objects, invitations and audit.events; ADD COLUMN
-- with a constant default is catalog-only (no rewrite); the audit constraint swap scans audit.events once. No data
-- migration, no destructive step, no edit to an accepted migration.
--
-- Down (development and test only) refuses while any Organization has the override enabled or audit.events holds a row
-- with an 00017 action, reason or operation (inspected with FORCE lifted inside the transaction; a refusal rolls the
-- lift back), then restores version 16 byte-identically.

-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow regrole;
  rt regrole;
  pr regrole;
  ar regrole;
  cw regrole;
  r record;
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regrole(db || '_runtime') IS NULL OR to_regrole(db || '_policy_reader') IS NULL OR to_regrole(db || '_auth_reader') IS NULL
     OR to_regrole(db || '_claim_writer') IS NULL THEN
    RAISE EXCEPTION 'roles for database % are missing: run infra/postgres/provision-roles.sql (or a fresh 10-roles.sh) first', db;
  END IF;
  ow := current_user::regrole; rt := (db || '_runtime')::regrole; pr := (db || '_policy_reader')::regrole;
  ar := (db || '_auth_reader')::regrole; cw := (db || '_claim_writer')::regrole;

  -- The 00015 helper checks, unchanged.
  FOR r IN SELECT rolname, rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolinherit, rolcanlogin, rolreplication
           FROM pg_roles WHERE oid IN (pr, ar, cw) LOOP
    IF r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication THEN
      RAISE EXCEPTION 'role % has an unsafe attribute: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
    IF r.rolcanlogin OR r.rolinherit THEN
      RAISE EXCEPTION 'helper role % must be NOLOGIN NOINHERIT: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
  END LOOP;
  IF (SELECT count(*) FROM pg_auth_members m WHERE m.roleid IN (pr, ar, cw) AND m.member = ow
        AND NOT m.inherit_option AND m.set_option AND NOT m.admin_option) <> 3 THEN
    RAISE EXCEPTION 'the owner must be a member of the three helper roles WITH INHERIT FALSE, SET TRUE: run infra/postgres/provision-roles.sql';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member IN (rt, cw)) THEN
    RAISE EXCEPTION 'the runtime and claim-writer roles must not be members of any role';
  END IF;

  -- Version 16: the migration history ends at 16, the touched 00006/00007/00014/00015/00016 objects exist, no 00017
  -- object exists.
  IF (SELECT max(version_id) FROM app.goose_db_version) IS DISTINCT FROM 16 THEN
    RAISE EXCEPTION 'version-16 precondition: the current version is %', (SELECT max(version_id) FROM app.goose_db_version);
  END IF;
  IF to_regprocedure('app.role_bindings_authority_guard()') IS NULL
     OR to_regprocedure('app.rls_user_enabled_lock(uuid, text, uuid)') IS NULL
     OR to_regprocedure('app.invitation_claim_group(uuid, bytea)') IS NULL
     OR to_regprocedure('app.rls_owner_count(text, uuid)') IS NULL
     OR to_regprocedure('app.rls_group_invitation_candidates(uuid, uuid, integer)') IS NULL
     OR to_regprocedure('app.organizations_enrollment_guard()') IS NULL
     OR to_regprocedure('app.organizations_group_override_guard()') IS NOT NULL
     OR EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.organizations'::regclass AND NOT attisdropped
                  AND attname = 'group_management_override')
     OR (SELECT count(*) FROM pg_policy WHERE polname IN ('organizations_policy_reader', 'organizations_claim_select',
           'organizations_claim_lock') AND polrelid = 'app.organizations'::regclass) <> 3
     OR (SELECT count(*) FROM pg_policy WHERE (polrelid, polname) IN (
           ('app.role_bindings'::regclass, 'role_bindings_update'), ('app.role_bindings'::regclass, 'role_bindings_select'),
           ('app.role_bindings'::regclass, 'role_bindings_delete'), ('app.groups'::regclass, 'groups_update'),
           ('app.group_business_profiles'::regclass, 'group_business_profiles_insert'),
           ('app.business_profiles'::regclass, 'business_profiles_update'),
           ('app.business_profiles'::regclass, 'business_profiles_insert'),
           ('app.stored_objects'::regclass, 'stored_objects_select'), ('app.stored_objects'::regclass, 'stored_objects_insert'),
           ('app.stored_objects'::regclass, 'stored_objects_update'), ('app.invitations'::regclass, 'invitations_select'),
           ('app.invitations'::regclass, 'invitations_insert'), ('app.invitations'::regclass, 'invitations_update'))) <> 13 THEN
    RAISE EXCEPTION 'version-16 precondition: group-authority objects already exist or the version-16 shapes are missing';
  END IF;
  -- The claim writer owns exactly its seven 00016 bodies; the touched tables are forced.
  IF (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7
     OR EXISTS (SELECT 1 FROM pg_class WHERE oid IN ('app.organizations'::regclass, 'app.groups'::regclass,
                  'app.role_bindings'::regclass, 'app.business_profiles'::regclass, 'app.group_business_profiles'::regclass,
                  'app.stored_objects'::regclass, 'app.invitations'::regclass) AND NOT (relrowsecurity AND relforcerowsecurity))
     OR has_column_privilege(cw, 'app.organizations', 'name', 'SELECT') THEN
    RAISE EXCEPTION 'version-16 precondition: claim-writer inventory, FORCE or organizations column grants differ';
  END IF;
END
$$;
-- +goose StatementEnd

-- The override (Plan r0 §3 item 2): default false for every existing and future Organization; no backfill.
ALTER TABLE app.organizations ADD COLUMN group_management_override boolean NOT NULL DEFAULT false;

-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
  cw text := current_database() || '_claim_writer';
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  MGR text := '(SELECT app.rls_org_role()) IN (''org_owner'', ''org_admin'')';
  GMGR text := 'app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'')';
  -- G060: the context Organization's current override, read under the invoker's row-level security; NULL is false.
  OVR text := 'COALESCE((SELECT o.group_management_override FROM app.organizations o WHERE o.id = (SELECT app.ctx_org())), false)';
  -- amendment01 A1.3: true only while the schema's own trigger bodies execute (the guard's departure cascade); the
  -- runtime defines no trigger and no body executes caller-supplied SQL (the 00015 events_insert_claim precedent).
  DEP text := 'pg_trigger_depth() >= 1';
  manage text;
  linked text;
  vis text;
  w text;
BEGIN
  -- Column grants (correction02 C02-1): runtime UPDATE (the guard confines it to the owner), claim writer and policy
  -- reader SELECT of the one boolean (their existing USING (true) policies admit the row). Nothing else is widened.
  EXECUTE format('GRANT UPDATE (group_management_override) ON app.organizations TO %I', rt);
  EXECUTE format('GRANT SELECT (group_management_override) ON app.organizations TO %I', cw);
  EXECUTE format('GRANT SELECT (group_management_override) ON app.organizations TO %I', pr);

  -- The owner-only guard (the 00016 enrollment-guard pattern): a session-role change of the override only by the
  -- current Organization owner of the context; the superuser operations channel is not a session role.
  EXECUTE $r$
CREATE FUNCTION app.organizations_group_override_guard() RETURNS trigger
  LANGUAGE plpgsql SECURITY INVOKER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  st text;
BEGIN
  IF NEW.group_management_override IS DISTINCT FROM OLD.group_management_override
     AND current_setting('is_superuser') IS DISTINCT FROM 'on' THEN
    IF app.rls_org_role() IS DISTINCT FROM 'org_owner' THEN
      RAISE EXCEPTION 'group management override owner required'
        USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'group_override_owner';
    END IF;
    -- amendment01 A2: the owner must be enabled now; FOR SHARE waits for an in-flight DisableUser (users FOR UPDATE)
    -- and returns the committed status (the rls_user_enabled_lock pattern).
    SELECT us.status INTO st FROM app.users us WHERE us.id = app.ctx_user() FOR SHARE;
    IF st IS DISTINCT FROM 'active' THEN
      RAISE EXCEPTION 'group management override owner required (enabled)'
        USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'group_override_owner';
    END IF;
  END IF;
  RETURN NEW;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.organizations_group_override_guard() FROM PUBLIC;
  CREATE TRIGGER organizations_group_override_guard BEFORE UPDATE OF group_management_override ON app.organizations
    FOR EACH ROW EXECUTE FUNCTION app.organizations_group_override_guard();

  -- role_bindings_select/update/delete (amendment01 A1.3): own-row visibility and Leave (the own-row branches) and
  -- genuine Group authority (GMGR) are kept verbatim; the Group branch admits Organization staff only with the override
  -- (OVR) or, for select/delete and the update USING (the cascade's FOR NO KEY UPDATE lock), inside the guard's
  -- Organization-departure cascade (DEP). At depth zero a raw roster read, Group-binding lock or ordinary Group-binding
  -- UPDATE/DELETE by staff without the override sees no row. The Group branch of
  -- update WITH CHECK admits the Organization path only with the override and never for the actor's own binding (r1 C3);
  -- step (3b) of the guard refuses the own-row role change before WITH CHECK runs (correction02 C02-2).
  DROP POLICY role_bindings_select ON app.role_bindings;
  EXECUTE format('CREATE POLICY role_bindings_select ON app.role_bindings FOR SELECT TO %I, %I USING ('
    || '%s IS NOT NULL AND organization_id = %s AND source = ''membership'' AND ('
    || '(subject_user_id = %s AND (scope_type = ''organization'' OR %s IS NULL OR group_id = %s))'
    || ' OR (scope_type = ''organization'' AND %s)'
    || ' OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR (%s AND (%s OR %s))))))',
    rt, ow, R, O, U, G, G, MGR, G, G, GMGR, MGR, OVR, DEP);
  DROP POLICY role_bindings_update ON app.role_bindings;
  -- attempt3 (SR2-3): USING also admits DEP so the cascade can lock the subject's Group bindings before reading their
  -- role; WITH CHECK below has no DEP, so staff without the override still cannot change a Group binding at any depth.
  manage := format('organization_id = %s AND source = ''membership'' AND ('
    || '(subject_user_id = %s AND %s IS NOT NULL AND (scope_type = ''organization'' OR %s IS NULL OR group_id = %s)) '
    || 'OR (scope_type = ''organization'' AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR (%s AND (%s OR %s)))))',
    O, U, R, G, G, MGR, G, G, GMGR, MGR, OVR, DEP);
  EXECUTE format('CREATE POLICY role_bindings_update ON app.role_bindings FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND source = ''membership'' AND ((scope_type = ''organization'' AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR (%s AND %s AND subject_user_id <> %s)))))',
    rt, ow, manage, O, MGR, G, G, GMGR, MGR, OVR, U);
  DROP POLICY role_bindings_delete ON app.role_bindings;
  manage := format('organization_id = %s AND source = ''membership'' AND ('
    || '(subject_user_id = %s AND %s IS NOT NULL AND (scope_type = ''organization'' OR %s IS NULL OR group_id = %s)) '
    || 'OR (scope_type = ''organization'' AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR (%s AND (%s OR %s)))))',
    O, U, R, G, G, MGR, G, G, GMGR, MGR, OVR, DEP);
  EXECUTE format('CREATE POLICY role_bindings_delete ON app.role_bindings FOR DELETE TO %I, %I USING (%s)', rt, ow, manage);

  -- groups_update: visibility unchanged (metadata); the write needs the Group's owner/admin or the override.
  DROP POLICY groups_update ON app.groups;
  vis := format('organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR id = %s) AND (%s OR app.rls_group_role(id) IS NOT NULL)',
    O, R, G, G, MGR);
  EXECUTE format('CREATE POLICY groups_update ON app.groups FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR id = %s) AND (app.rls_group_role(id) IN (''group_owner'', ''group_admin'') '
    || 'OR (%s AND %s)))', rt, ow, vis, O, R, G, G, MGR, OVR);

  -- group_business_profiles_insert: onboarding completion by the Group's owner/admin (eligible profile) or the override.
  DROP POLICY group_business_profiles_insert ON app.group_business_profiles;
  EXECUTE format('CREATE POLICY group_business_profiles_insert ON app.group_business_profiles FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR group_id = %s) AND ((%s AND %s) OR ('
    || 'app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'') AND app.rls_profile_eligible(business_profile_id))))',
    rt, ow, O, R, G, G, MGR, OVR);

  -- business_profiles_update: visibility unchanged (ordinary profile fields are metadata); the write needs the override
  -- or the existing linked Group path under group_admins_edit_business_profiles.
  DROP POLICY business_profiles_update ON app.business_profiles;
  linked := 'EXISTS (SELECT 1 FROM app.group_business_profiles a WHERE a.organization_id = business_profiles.organization_id '
    || 'AND a.business_profile_id = business_profiles.id AND (' || G || ' IS NULL OR a.group_id = ' || G || ') '
    || 'AND app.rls_group_role(a.group_id) IN (%s))';
  vis := format('organization_id = %s AND %s IS NOT NULL AND (%s OR %s)', O, R, MGR,
    format(linked, '''group_owner'', ''group_admin'', ''group_member'''));
  EXECUTE format('CREATE POLICY business_profiles_update ON app.business_profiles FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND ((%s AND %s) OR ((SELECT o.group_admins_edit_business_profiles FROM app.organizations o '
    || 'WHERE o.id = %s) AND %s)))',
    rt, ow, vis, O, R, MGR, OVR, O, format(linked, '''group_owner'', ''group_admin'''));

  -- business_profiles_insert: a new profile during onboarding by a Group owner/admin or under the override.
  DROP POLICY business_profiles_insert ON app.business_profiles;
  EXECUTE format('CREATE POLICY business_profiles_insert ON app.business_profiles FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND ((%s AND %s) OR EXISTS (SELECT 1 FROM app.role_bindings rb '
    || 'WHERE rb.subject_user_id = %s AND rb.organization_id = %s AND rb.scope_type = ''group'' '
    || 'AND rb.source = ''membership'' AND rb.role IN (''group_owner'', ''group_admin'') AND (%s IS NULL OR rb.scope_id = %s))))',
    rt, ow, O, R, MGR, OVR, U, O, G, G);

  -- stored_objects: a Group object needs an explicit role in that exact Group under either setting (Spec R3); Organization
  -- objects keep their Organization rules.
  DROP POLICY stored_objects_select ON app.stored_objects;
  DROP POLICY stored_objects_insert ON app.stored_objects;
  DROP POLICY stored_objects_update ON app.stored_objects;
  EXECUTE format('CREATE POLICY stored_objects_select ON app.stored_objects FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND %s IS NOT NULL AND (group_id IS NULL OR ((%s IS NULL OR group_id = %s) AND '
    || 'app.rls_group_role(group_id) IS NOT NULL)))', rt, ow, O, R, G, G);
  EXECUTE format('CREATE POLICY stored_objects_insert ON app.stored_objects FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND creator_user_id = %s AND state = ''pending'' AND ('
    || '(group_id IS NULL AND %s) OR (group_id IS NOT NULL AND (%s IS NULL OR group_id = %s) AND '
    || 'app.rls_group_role(group_id) IS NOT NULL)))', rt, ow, O, R, U, MGR, G, G);
  w := format('organization_id = %s AND %s IS NOT NULL AND ((group_id IS NULL AND %s) OR (group_id IS NOT NULL AND '
    || '(%s IS NULL OR group_id = %s) AND (app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'') OR '
    || '(creator_user_id = %s AND app.rls_group_role(group_id) IS NOT NULL))))', O, R, MGR, G, G, U);
  EXECUTE format('CREATE POLICY stored_objects_update ON app.stored_objects FOR UPDATE TO %I, %I USING (%s) WITH CHECK (%s)',
    rt, ow, w, w);

  -- invitations: Group invitations are administered by the Group's owner/admin or under the override (rosters and
  -- tokens are administration, Spec R1); Organization invitations unchanged.
  DROP POLICY invitations_select ON app.invitations;
  DROP POLICY invitations_insert ON app.invitations;
  DROP POLICY invitations_update ON app.invitations;
  manage := format('((scope_type = ''organization'' AND %s) OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) '
    || 'AND (%s OR (%s AND %s))))', MGR, G, G, GMGR, MGR, OVR);
  EXECUTE format('CREATE POLICY invitations_select ON app.invitations FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND (%s OR intended_user_id = %s OR consumed_by_user_id = %s))', rt, ow, O, manage, U, U);
  EXECUTE format('CREATE POLICY invitations_insert ON app.invitations FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND issuer_user_id = %s AND ((scope_type = ''organization'' AND %s IS NULL AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR (%s AND %s)) '
    || 'AND intended_user_id IS DISTINCT FROM issuer_user_id '
    || 'AND app.is_organization_member(intended_user_id, %s))))', rt, ow, O, U, G, MGR, G, G, GMGR, MGR, OVR, O);
  EXECUTE format('CREATE POLICY invitations_update ON app.invitations FOR UPDATE TO %I, %I USING ('
    || 'organization_id = %s AND %s) WITH CHECK (organization_id = %s AND %s)', rt, ow, O, manage, O, manage);

  -- The two policy-reader bodies, replaced as the policy reader (the 00006 SET chain): the Group branch admits the
  -- Organization path only with the override.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pr);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.rls_owner_count(p_scope_type text, p_scope_id uuid) RETURNS integer
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text := app.rls_org_role();
BEGIN
  IF app.ctx_user() IS NULL OR o IS NULL OR r IS NULL OR p_scope_id IS NULL THEN
    RETURN 0;
  END IF;
  IF p_scope_type = 'organization' THEN
    IF p_scope_id <> o THEN
      RETURN 0;
    END IF;
    RETURN (SELECT count(*) FROM app.role_bindings rb
            WHERE rb.organization_id = o AND rb.scope_type = 'organization' AND rb.scope_id = o
              AND rb.role = 'org_owner' AND rb.source = 'membership');
  ELSIF p_scope_type = 'group' THEN
    -- G060: the Organization path only with the override; any own role of the Group otherwise.
    IF (g IS NOT NULL AND p_scope_id <> g)
       OR NOT ((r IN ('org_owner', 'org_admin')
                AND coalesce((SELECT org.group_management_override FROM app.organizations org WHERE org.id = o), false))
               OR app.rls_group_role(p_scope_id) IS NOT NULL) THEN
      RETURN 0;
    END IF;
    RETURN (SELECT count(*) FROM app.role_bindings rb
            WHERE rb.organization_id = o AND rb.scope_type = 'group' AND rb.scope_id = p_scope_id
              AND rb.role = 'group_owner' AND rb.source = 'membership');
  END IF;
  RETURN 0;
END $$
$r$;
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.rls_group_invitation_candidates(p_group uuid, p_after uuid, p_limit integer)
  RETURNS TABLE (user_id uuid, display_name text)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text := app.rls_org_role();
BEGIN
  -- G060: the Organization path only with the override; the Group's owner/admin otherwise.
  IF app.ctx_user() IS NULL OR o IS NULL OR r IS NULL OR p_group IS NULL OR p_limit IS NULL OR p_limit < 1
     OR (g IS NOT NULL AND g <> p_group)
     OR NOT coalesce((r IN ('org_owner', 'org_admin')
                      AND coalesce((SELECT org.group_management_override FROM app.organizations org WHERE org.id = o), false))
                     OR app.rls_group_role(p_group) IN ('group_owner', 'group_admin'), false)
     OR NOT EXISTS (SELECT 1 FROM app.role_bindings x WHERE x.organization_id = o AND x.scope_type = 'group'
                      AND x.scope_id = p_group AND x.source = 'membership') THEN
    RETURN;
  END IF;
  RETURN QUERY
  SELECT us.id, us.display_name FROM app.role_bindings rb JOIN app.users us ON us.id = rb.subject_user_id
  WHERE rb.organization_id = o AND rb.scope_type = 'organization' AND rb.scope_id = o AND rb.source = 'membership'
    AND us.status = 'active' AND (p_after IS NULL OR us.id > p_after)
    AND NOT EXISTS (SELECT 1 FROM app.role_bindings gb WHERE gb.subject_user_id = rb.subject_user_id
                      AND gb.organization_id = o AND gb.scope_type = 'group' AND gb.scope_id = p_group)
  ORDER BY us.id
  LIMIT least(p_limit, 101);
END $$
$r$;
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pr);

  -- The two claim-writer bodies, replaced as the claim writer (the 00015 SET LOCAL ROLE pattern).
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', cw);
  EXECUTE format('SET LOCAL ROLE %I', cw);
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.rls_user_enabled_lock(p_user uuid, p_scope_type text, p_scope_id uuid) RETURNS boolean
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  admitted boolean := false;
  st text;
BEGIN
  IF u IS NULL OR p_user IS NULL THEN
    RETURN NULL;
  END IF;
  IF p_user = u THEN
    admitted := true;
  ELSIF o IS NOT NULL AND p_scope_id IS NOT NULL AND EXISTS (
          SELECT 1 FROM app.role_bindings rb WHERE rb.subject_user_id = p_user AND rb.organization_id = o
            AND rb.scope_type = p_scope_type AND rb.scope_id = p_scope_id AND rb.source = 'membership') THEN
    IF p_scope_type = 'organization' THEN
      admitted := p_scope_id = o AND coalesce(app.rls_org_role() IN ('org_owner', 'org_admin'), false);
    ELSIF p_scope_type = 'group' THEN
      -- G060: the Organization path only with the override.
      admitted := (g IS NULL OR g = p_scope_id)
        AND coalesce((app.rls_org_role() IN ('org_owner', 'org_admin')
                      AND coalesce((SELECT org.group_management_override FROM app.organizations org WHERE org.id = o), false))
                     OR app.rls_group_role(p_scope_id) IN ('group_owner', 'group_admin'), false);
    END IF;
  END IF;
  IF NOT admitted THEN
    RETURN NULL;
  END IF;
  PERFORM 1 FROM app.users us WHERE us.id = p_user FOR SHARE;
  SELECT us.status INTO st FROM app.users us WHERE us.id = p_user;
  RETURN st = 'active';
END $$
$r$;
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.invitation_claim_group(p_id uuid, p_digest bytea)
  RETURNS TABLE (outcome text, organization_id uuid, scope_type text, scope_id uuid, group_id uuid, role text, binding_id uuid)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  inv app.invitations%ROWTYPE;
  ex_id uuid;
  ex_role text;
  bid uuid;
  res text;
BEGIN
  IF u IS NULL OR o IS NULL THEN
    RAISE EXCEPTION 'invitation claim context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'invitation_context';
  END IF;
  -- (0) G060 (r1 C4, security review): the context Organization row FOR SHARE first, the facade's order (Organization,
  -- then invitation), so a committed or in-flight override change (FOR NO KEY UPDATE) is ordered before this
  -- consumption's authority check and no raw claim can invert the lock order; concurrent claims never serialize on it.
  PERFORM 1 FROM app.organizations og WHERE og.id = o FOR SHARE;
  -- (1) serialize duplicate claims on the invitation row.
  SELECT i.* INTO inv FROM app.invitations i WHERE i.id = p_id FOR NO KEY UPDATE;
  -- (2) digest, kind, tenant, selected scope and recipient intent; any mismatch is the one closed not_found. G060 (r1 C3):
  -- an issuer never consumes its own Group invitation, under any later authority.
  IF NOT FOUND OR p_digest IS NULL OR inv.token_digest <> p_digest OR inv.scope_type <> 'group'
     OR inv.organization_id <> o OR (g IS NOT NULL AND g <> inv.group_id) OR inv.intended_user_id IS DISTINCT FROM u
     OR inv.issuer_user_id = u THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF inv.consumed_at IS NOT NULL THEN
    -- Replay by the same recipient: already_member only with a current qualifying binding; never restored.
    IF inv.consumed_by_user_id = u THEN
      SELECT rb.id, rb.role INTO ex_id, ex_role FROM app.role_bindings rb
      WHERE rb.subject_user_id = u AND rb.organization_id = inv.organization_id AND rb.scope_type = inv.scope_type
        AND rb.scope_id = inv.scope_id AND rb.source = 'membership';
      IF ex_id IS NOT NULL AND (CASE ex_role WHEN 'group_owner' THEN 3 WHEN 'group_admin' THEN 2 ELSE 1 END)
                               >= (CASE inv.role WHEN 'group_admin' THEN 2 ELSE 1 END) THEN
        RETURN QUERY SELECT 'already_member'::text, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, ex_role, ex_id;
      ELSE
        RETURN QUERY SELECT 'consumed'::text, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, inv.role, NULL::uuid;
      END IF;
      RETURN;
    END IF;
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF inv.revoked_at IS NOT NULL OR inv.expires_at <= now() THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  -- (3a) G060 (r1 C4): the issuer's and the recipient's user rows FOR SHARE in ascending id (DisableUser locks users
  -- FOR UPDATE, then sessions, and never an invitation or Organization row).
  PERFORM 1 FROM app.users us WHERE us.id IN (inv.issuer_user_id, u) ORDER BY us.id FOR SHARE;
  -- (3) the enabled recipient under a lock consistent with DisableUser.
  IF NOT coalesce(app.rls_user_enabled_lock(u, inv.scope_type, inv.scope_id), false) THEN
    RAISE EXCEPTION 'invitation recipient ineligible'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
  END IF;
  -- (4) lock the issuer's and the recipient's bindings for the scope (ascending id), then check with the post-lock
  -- snapshot: the issuer's current authority (G060: Group owner/admin of the scope, or Organization owner/admin while the
  -- override is currently enabled), the recipient's current Organization membership, and the issuer still enabled.
  PERFORM 1 FROM app.role_bindings rb
  WHERE rb.organization_id = inv.organization_id AND rb.source = 'membership'
    AND rb.subject_user_id IN (inv.issuer_user_id, u)
    AND (rb.scope_type = 'organization' OR rb.scope_id = inv.scope_id)
  ORDER BY rb.id FOR SHARE;
  IF NOT EXISTS (SELECT 1 FROM app.role_bindings rb
                 WHERE rb.subject_user_id = inv.issuer_user_id AND rb.organization_id = inv.organization_id
                   AND rb.source = 'membership'
                   AND ((rb.scope_type = 'organization' AND rb.role IN ('org_owner', 'org_admin')
                         AND coalesce((SELECT og.group_management_override FROM app.organizations og
                                       WHERE og.id = inv.organization_id), false))
                     OR (rb.scope_type = 'group' AND rb.scope_id = inv.scope_id AND rb.role IN ('group_owner', 'group_admin')))) THEN
    RAISE EXCEPTION 'invitation issuer authority lost'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'issuer_authority_lost';
  END IF;
  IF NOT app.is_organization_member(u, inv.organization_id) THEN
    RAISE EXCEPTION 'invitation recipient ineligible'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM app.users us WHERE us.id = inv.issuer_user_id AND us.status = 'active') THEN
    RAISE EXCEPTION 'invitation issuer authority lost'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'issuer_authority_lost';
  END IF;
  -- (5) existing scope binding: equal/higher -> already_member; lower -> upgrade; none -> insert from the intent.
  SELECT rb.id, rb.role INTO ex_id, ex_role FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.organization_id = inv.organization_id AND rb.scope_type = inv.scope_type
    AND rb.scope_id = inv.scope_id AND rb.source = 'membership';
  IF ex_id IS NOT NULL THEN
    IF (CASE ex_role WHEN 'group_owner' THEN 3 WHEN 'group_admin' THEN 2 ELSE 1 END)
       >= (CASE inv.role WHEN 'group_admin' THEN 2 ELSE 1 END) THEN
      res := 'already_member';
    ELSE
      UPDATE app.role_bindings rb SET role = inv.role WHERE rb.id = ex_id;
      res := 'upgraded'; ex_role := inv.role;
    END IF;
    bid := ex_id;
  ELSE
    bid := gen_random_uuid();
    INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
    VALUES (bid, inv.organization_id, u, inv.scope_type, inv.scope_id, inv.group_id, inv.role, 'membership');
    res := 'joined'; ex_role := inv.role;
  END IF;
  -- (6) consume.
  UPDATE app.invitations i SET consumed_at = now(), consumed_by_user_id = u, consumed_binding_id = bid WHERE i.id = p_id;
  RETURN QUERY SELECT res, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, ex_role, bid;
END $$
$r$;
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', cw);

  -- The authority and cardinality guard, replaced (owner-owned SECURITY INVOKER; identity, trigger and ACL kept) with
  -- step (3b) (correction02 C02-2): a runtime or owner session never changes the role of its own Group binding unless it
  -- holds that Group's owner role (the owner's own demotion stays under steps (1) and (4)). The claim writer's validated
  -- upgrade (invitation_claim_group step (5)) is governed by its own checks and is not affected.
  EXECUTE format($f$
CREATE OR REPLACE FUNCTION app.role_bindings_authority_guard() RETURNS trigger
  LANGUAGE plpgsql VOLATILE SET search_path = pg_catalog AS $b$
DECLARE
  old_owner boolean;
  new_owner boolean;
  r record;
BEGIN
  IF current_user::text NOT IN (%L, %L, %L) THEN
    IF TG_OP = 'DELETE' THEN
      RETURN OLD;
    END IF;
    RETURN NEW;
  END IF;
  old_owner := OLD.role IN ('org_owner', 'group_owner');
  new_owner := TG_OP = 'UPDATE' AND NEW.role IN ('org_owner', 'group_owner');
  -- (1) owner authority: no Organization-role override of Group ownership.
  IF old_owner OR new_owner THEN
    IF OLD.scope_type = 'organization' THEN
      IF app.rls_org_role() IS DISTINCT FROM 'org_owner' THEN
        RAISE EXCEPTION 'owner authority required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'owner_authority';
      END IF;
    ELSIF app.rls_group_role(OLD.group_id) IS DISTINCT FROM 'group_owner' THEN
      RAISE EXCEPTION 'owner authority required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'owner_authority';
    END IF;
  END IF;
  -- (2) promotion eligibility: same-scope admin, enabled under a lock consistent with DisableUser, and (Group) a
  -- current Organization member.
  IF NOT old_owner AND new_owner THEN
    IF OLD.role IS DISTINCT FROM (CASE OLD.scope_type WHEN 'organization' THEN 'org_admin' ELSE 'group_admin' END)
       OR NOT coalesce(app.rls_user_enabled_lock(OLD.subject_user_id, OLD.scope_type, OLD.scope_id), false)
       OR (OLD.scope_type = 'group' AND NOT app.is_organization_member(OLD.subject_user_id, OLD.organization_id)) THEN
      RAISE EXCEPTION 'promotion recipient ineligible'
        USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
    END IF;
  END IF;
  -- (3) Organization departure (amendment01 A1.3): an Organization-only context; then, inside this same statement, the
  -- subject's Group bindings are deleted in ascending group_id (another subject's Group-owner binding refuses with
  -- owner_subject: succession first); each nested DELETE passes this guard (steps (1) and (4)) and the audit capture as
  -- the same invoker, and its row-level security admits staff only at trigger depth >= 1. The dependent count (the
  -- definer count sees every sibling Group regardless of the selected context) remains the invariant check.
  IF TG_OP = 'DELETE' AND OLD.scope_type = 'organization' THEN
    IF app.ctx_group() IS NOT NULL THEN
      RAISE EXCEPTION 'organization-only context required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'context_required';
    END IF;
    -- attempt3 (SR2-3): each Group binding is locked FOR NO KEY UPDATE, in ascending group_id, before its role is read,
    -- so a concurrent promotion either commits first (the locked read returns group_owner: owner_subject) or waits and
    -- then finds the row deleted. The lock needs role_bindings_update USING, whose Group branch admits staff at
    -- trigger depth >= 1 only for this lock; its WITH CHECK keeps the override requirement (no mutation).
    FOR r IN SELECT rb.id, rb.group_id, rb.role FROM app.role_bindings rb
             WHERE rb.organization_id = OLD.organization_id AND rb.subject_user_id = OLD.subject_user_id
               AND rb.scope_type = 'group' AND rb.source = 'membership' ORDER BY rb.group_id FOR NO KEY UPDATE LOOP
      IF r.role = 'group_owner' AND OLD.subject_user_id IS DISTINCT FROM app.ctx_user() THEN
        RAISE EXCEPTION 'group owner subject' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'owner_subject';
      END IF;
      DELETE FROM app.role_bindings rb2 WHERE rb2.id = r.id;
    END LOOP;
    IF app.rls_dependent_group_bindings(OLD.subject_user_id) > 0 THEN
      RAISE EXCEPTION 'dependent group bindings remain'
        USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'dependent_bindings';
    END IF;
  END IF;
  -- (3b) G060 self-escalation: a session role (runtime or owner, never the claim writer) changes the role of its own
  -- Group binding only as that Group's owner.
  IF TG_OP = 'UPDATE' AND current_user::text IN (%L, %L) AND OLD.scope_type = 'group'
     AND NEW.subject_user_id = app.ctx_user() AND app.rls_group_role(OLD.group_id) IS DISTINCT FROM 'group_owner' THEN
    RAISE EXCEPTION 'own group role change refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'self_escalation';
  END IF;
  -- (4) last owner: command A locks the scope row; command B counts with the snapshot taken after A's wait.
  IF old_owner AND NOT new_owner THEN
    IF OLD.scope_type = 'organization' THEN
      PERFORM 1 FROM app.organizations o WHERE o.id = OLD.scope_id FOR NO KEY UPDATE;
    ELSE
      PERFORM 1 FROM app.groups g WHERE g.id = OLD.scope_id FOR NO KEY UPDATE;
    END IF;
    IF app.rls_owner_count(OLD.scope_type, OLD.scope_id) - 1 < 1 THEN
      RAISE EXCEPTION 'last owner' USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'last_owner';
    END IF;
  END IF;
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END $b$
$f$, rt, ow, cw, rt, ow);
END
$do$;
-- +goose StatementEnd

-- The widened catalog constraints (Plan r0 §3 item 7, r1 C5): actions +1 core (45 -> 46: access.override_used), reasons
-- +2 (68 -> 70: the override's organization.policy_changed reasons), operations +1 (64 -> 65: the override write's
-- access.denied operation). Resource types unchanged (11). One validating scan of audit.events.
ALTER TABLE audit.events DROP CONSTRAINT events_reason_valid, DROP CONSTRAINT events_operation_valid,
  DROP CONSTRAINT events_action_valid;
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
                    'organization.branding_changed',
                    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
                    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
                    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked',
                    'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                    'object.deletion_pending', 'revocation.unaudited', 'access.override_used')),
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
                    'withdrawn', 'withdraw_error', 'withdraw_unknown', 'organization_not_member', 'host_changed',
                    'updated', 'reset', 'logo_set', 'logo_removed',
                    'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                    'recipient_proof_unavailable', 'recipient_ineligible', 'session_invalid',
                    'operator_verified_absent', 'operator_backfilled', 'history_unresolved',
                    'group_management_override_enabled', 'group_management_override_disabled')),
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug',
                    'branding.get', 'branding.update', 'branding.reset', 'branding.set_logo', 'branding.remove_logo',
                    'group.suggest_slug', 'group.create', 'group.list', 'group.onboarding.get',
                    'group.onboarding.complete_new', 'group.onboarding.complete_existing', 'business_profile.list',
                    'membership.list', 'membership.self', 'membership.change_role', 'membership.remove', 'membership.leave',
                    'ownership.promote', 'ownership.demote', 'ownership.transfer',
                    'invitation.list', 'invitation.mine', 'invitation.candidates', 'invitation.issue', 'invitation.revoke',
                    'invitation.get', 'invitation.accept', 'invitation.consume', 'retirement.eligibility',
                    'enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request', 'enrollment.admit',
                    'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update', 'enrollment.requests.list',
                    'enrollment.requests.decide', 'enrollment.requests.revoke', 'organization.set_group_override'));

-- The attribution and shape rules (the 00016 body otherwise unchanged; CREATE OR REPLACE keeps identity, owner, ACL and
-- trigger): organization.policy_changed admits the two override reasons; access.override_used is a direct (core) append
-- by an established user actor naming the actual mutated resource (r1 C5): a Group with its own id as group_id for the
-- Group operations, or a Business Profile for business_profile.update (no Group) and for an existing-profile onboarding
-- completion (with that Group); no reason, role, subject or TTL.
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
    'domain.state_changed', 'domain.checked', 'organization.branding_changed',
    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked');
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
               THEN NEW.reason IN ('group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled',
                                   'group_management_override_enabled', 'group_management_override_disabled')
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
    WHEN NEW.action = 'organization.branding_changed' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization'
      AND NEW.resource_id = NEW.organization_id AND NEW.reason IN ('updated', 'reset', 'logo_set', 'logo_removed')
    WHEN NEW.action IN ('invitation.issued', 'invitation.revoked', 'invitation.consumed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'invitation' AND NEW.resource_id IS NOT NULL
      AND CASE WHEN NEW.group_id IS NULL THEN NEW.role IN ('org_admin', 'org_member')
               ELSE NEW.role IN ('group_admin', 'group_member') END
      AND (NEW.action <> 'invitation.consumed' OR NEW.subject_user_id IS NOT NULL)
    WHEN NEW.action IN ('enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.group_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization' AND NEW.resource_id = NEW.organization_id
      AND CASE NEW.action
            WHEN 'enrollment.policy_changed' THEN NEW.reason IS NULL AND NEW.subject_user_id IS NULL
            WHEN 'enrollment.readiness_changed' THEN NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NULL
              AND NEW.reason IN ('operator_verified_absent', 'operator_backfilled', 'history_unresolved')
            ELSE NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NOT NULL AND NEW.reason = 'operator_backfilled' END
    WHEN NEW.action IN ('enrollment.requested', 'enrollment.decided', 'enrollment.revoked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'enrollment_request' AND NEW.resource_id IS NOT NULL AND NEW.actor_kind = 'user'
      AND (NEW.action <> 'enrollment.requested' OR NEW.actor_user_id = NEW.subject_user_id)
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
            'organization_not_member', 'host_changed', 'recipient_proof_unavailable'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
        'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch', 'recipient_ineligible',
        'session_invalid', 'host_changed')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object', 'organization_domain')))
      AND (NEW.reason IS DISTINCT FROM 'resource_mismatch'
           OR (NEW.operation = 'mcp.request' AND NEW.organization_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.reason NOT IN ('enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                              'recipient_ineligible', 'session_invalid', 'host_changed')
           OR (NEW.operation IN ('enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request',
                                 'enrollment.admit', 'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update',
                                 'enrollment.requests.list', 'enrollment.requests.decide', 'enrollment.requests.revoke')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.operation NOT IN ('enrollment.start', 'enrollment.admit')
           OR (NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL
               AND CASE WHEN NEW.operation = 'enrollment.start'
                        THEN NEW.reason IN ('enrollment_closed', 'invitation_mismatch')
                        ELSE NEW.actor_user_id IS NOT NULL AND NEW.reason IN ('target_not_found', 'enrollment_closed',
                             'enrollment_restricted', 'approval_revoked', 'recipient_ineligible', 'session_invalid',
                             'host_changed') END))
    WHEN NEW.action = 'access.override_used' THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL
      AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL AND NEW.actor_kind = 'user'
      AND NEW.organization_id IS NOT NULL AND NEW.resource_id IS NOT NULL
      AND app.rls_org_role() IN ('org_owner', 'org_admin')
      AND coalesce((SELECT og.group_management_override FROM app.organizations og WHERE og.id = NEW.organization_id), false)
      AND CASE NEW.resource_type
            WHEN 'group' THEN NEW.group_id IS NOT NULL AND NEW.resource_id = NEW.group_id
              AND NEW.operation IN ('group.rename', 'group.change_slug', 'group.onboarding.complete_new',
                                    'group.onboarding.complete_existing', 'membership.change_role', 'membership.remove',
                                    'invitation.issue', 'invitation.revoke')
            WHEN 'business_profile' THEN
              (NEW.operation = 'business_profile.update' AND NEW.group_id IS NULL)
              OR (NEW.operation = 'group.onboarding.complete_existing' AND NEW.group_id IS NOT NULL)
            ELSE false END
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

-- organization.policy_changed for the override (the 00011 body otherwise unchanged; CREATE OR REPLACE keeps identity,
-- owner and ACL), atomic with the change; the update trigger also watches the new column.
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
  IF NEW.group_management_override IS DISTINCT FROM OLD.group_management_override THEN
    INSERT INTO audit.events (action, outcome, reason, organization_id, resource_type, resource_id)
    VALUES ('organization.policy_changed', 'success',
            CASE WHEN NEW.group_management_override THEN 'group_management_override_enabled'
                 ELSE 'group_management_override_disabled' END,
            NEW.id, 'organization', NEW.id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

DROP TRIGGER audit_capture_update ON app.organizations;
CREATE TRIGGER audit_capture_update AFTER UPDATE OF name, slug, group_admins_edit_business_profiles, group_management_override
  ON app.organizations FOR EACH ROW
  WHEN (OLD.name IS DISTINCT FROM NEW.name OR OLD.slug IS DISTINCT FROM NEW.slug
        OR OLD.group_admins_edit_business_profiles IS DISTINCT FROM NEW.group_admins_edit_business_profiles
        OR OLD.group_management_override IS DISTINCT FROM NEW.group_management_override)
  EXECUTE FUNCTION audit.capture_organizations();

-- Postconditions (TestMigrationLifecycle repeats the version-17 column-grant and body pins).
-- +goose StatementBegin
DO $$
DECLARE
  ow regrole := current_user::regrole;
  rt regrole := (current_database() || '_runtime')::regrole;
  pr regrole := (current_database() || '_policy_reader')::regrole;
  ar regrole := (current_database() || '_auth_reader')::regrole;
  cw regrole := (current_database() || '_claim_writer')::regrole;
  actual text;
  body text;
  n_on bigint;
BEGIN
  -- Postcondition 1: the column (NOT NULL, default false) and its exact column ACL (correction02 C02-1).
  IF NOT EXISTS (SELECT 1 FROM pg_attribute a JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
                 WHERE a.attrelid = 'app.organizations'::regclass AND a.attname = 'group_management_override'
                   AND a.atttypid = 'boolean'::regtype AND a.attnotnull AND NOT a.attisdropped
                   AND pg_get_expr(d.adbin, d.adrelid) = 'false') THEN
    RAISE EXCEPTION 'app.organizations.group_management_override is missing or not boolean NOT NULL DEFAULT false';
  END IF;
  SELECT string_agg(y.v, ',' ORDER BY y.v COLLATE "C") INTO actual
  FROM (SELECT CASE x.grantee WHEN rt THEN 'rt' WHEN cw THEN 'cw' WHEN pr THEN 'pr' WHEN ar THEN 'ar' WHEN ow THEN 'ow'
                              ELSE 'other' END || '=' || x.privilege_type AS v
        FROM pg_attribute a CROSS JOIN LATERAL aclexplode(a.attacl) x
        WHERE a.attrelid = 'app.organizations'::regclass AND a.attname = 'group_management_override') y;
  IF actual IS DISTINCT FROM 'cw=SELECT,pr=SELECT,rt=UPDATE' THEN
    RAISE EXCEPTION 'group_management_override column ACL is %', actual;
  END IF;
  FOR actual, body IN SELECT g.who, coalesce((SELECT string_agg(s.x, ',' ORDER BY s.x COLLATE "C") FROM (
                        SELECT at.attname || ':' || a.privilege_type AS x FROM pg_attribute at CROSS JOIN LATERAL aclexplode(at.attacl) a
                        WHERE at.attrelid = 'app.organizations'::regclass AND at.attnum > 0 AND NOT at.attisdropped
                          AND a.grantee = g.oid) s), '')
                      FROM (VALUES ('cw', cw::oid), ('pr', pr::oid)) g(who, oid) LOOP
    IF body IS DISTINCT FROM (CASE actual
         WHEN 'cw' THEN 'enrollment_mode:SELECT,enrollment_mode:UPDATE,group_management_override:SELECT,id:SELECT'
         ELSE 'enrollment_mode:SELECT,group_management_override:SELECT,id:SELECT' END) THEN
      RAISE EXCEPTION '% column grants on app.organizations are %', actual, body;
    END IF;
  END LOOP;
  IF NOT has_column_privilege(rt, 'app.organizations', 'group_management_override', 'UPDATE')
     OR NOT has_column_privilege(cw, 'app.organizations', 'group_management_override', 'SELECT')
     OR NOT has_column_privilege(pr, 'app.organizations', 'group_management_override', 'SELECT')
     OR has_column_privilege(cw, 'app.organizations', 'group_management_override', 'UPDATE')
     OR has_column_privilege(pr, 'app.organizations', 'group_management_override', 'UPDATE')
     OR has_column_privilege(ar, 'app.organizations', 'group_management_override', 'SELECT')
     OR has_column_privilege(cw, 'app.organizations', 'name', 'SELECT')
     OR has_column_privilege(pr, 'app.organizations', 'name', 'SELECT') THEN
    RAISE EXCEPTION 'group_management_override privilege boundary violated';
  END IF;

  -- Postcondition 2: the guard is an owner-owned SECURITY INVOKER body with no grantee, fired BEFORE UPDATE OF the column;
  -- the claim writer still owns exactly seven bodies and the policy reader its version-16 inventory.
  SELECT p.proname || ':' || (p.proowner = ow)::text || ':' || p.prosecdef::text || ':'
         || coalesce((SELECT string_agg(a.grantee::regrole::text, '+') FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner), '-')
    INTO actual FROM pg_proc p WHERE p.oid = 'app.organizations_group_override_guard()'::regprocedure;
  IF actual IS DISTINCT FROM 'organizations_group_override_guard:true:false:-' THEN
    RAISE EXCEPTION 'override guard is %', actual;
  END IF;
  SELECT prosrc INTO body FROM pg_proc WHERE oid = 'app.organizations_group_override_guard()'::regprocedure;
  IF position('FROM app.users us WHERE us.id = app.ctx_user() FOR SHARE' IN body) = 0
     OR position('us.status' IN body) = 0
     OR position('IS DISTINCT FROM ''active''' IN body) = 0 THEN
    RAISE EXCEPTION 'override guard lacks the reviewed enabled-owner lock';
  END IF;
  SELECT string_agg(tgname, ',' ORDER BY tgname) INTO actual FROM pg_trigger
  WHERE tgrelid = 'app.organizations'::regclass AND NOT tgisinternal;
  IF actual IS DISTINCT FROM 'audit_capture_enrollment,audit_capture_insert,audit_capture_update,organizations_enrollment_guard,'
    || 'organizations_group_override_guard,organizations_keys_immutable,organizations_slug_guard' THEN
    RAISE EXCEPTION 'app.organizations triggers are %', actual;
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7 THEN
    RAISE EXCEPTION 'the claim writer owns % functions, expected 7', (SELECT count(*) FROM pg_proc WHERE proowner = cw);
  END IF;
  SELECT string_agg(proname, ',' ORDER BY proname) INTO actual FROM pg_proc WHERE proowner = pr;
  IF actual IS DISTINCT FROM 'rls_dependent_group_bindings,rls_enrollment_mode,rls_group_invitation_candidates,rls_group_role,'
    || 'rls_org_role,rls_own_organizations,rls_owner_count,rls_profile_eligible,rls_sole_owner_scopes' THEN
    RAISE EXCEPTION 'the policy reader owns %', actual;
  END IF;

  -- Postcondition 3: the re-created policies keep their commands and roles; exactly the ten administration policies
  -- read the override; only role_bindings_select, role_bindings_delete and (USING only) role_bindings_update admit the
  -- trigger-depth cascade branch; the Group object policies name no Organization role on their Group branch.
  SELECT string_agg(format('%s:%s:%s', p.polname, p.polcmd,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt THEN 'rt' WHEN ow THEN 'ow' ELSE 'other' END AS v FROM unnest(p.polroles) x) l)),
           ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p WHERE p.polrelid IN ('app.role_bindings'::regclass, 'app.groups'::regclass, 'app.group_business_profiles'::regclass,
                                        'app.business_profiles'::regclass, 'app.stored_objects'::regclass, 'app.invitations'::regclass)
    AND (coalesce(pg_get_expr(p.polqual, p.polrelid), '') || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), ''))
        LIKE '%group_management_override%';
  IF actual IS DISTINCT FROM 'business_profiles_insert:a:ow+rt;business_profiles_update:w:ow+rt;'
    || 'group_business_profiles_insert:a:ow+rt;groups_update:w:ow+rt;invitations_insert:a:ow+rt;invitations_select:r:ow+rt;'
    || 'invitations_update:w:ow+rt;role_bindings_delete:d:ow+rt;role_bindings_select:r:ow+rt;role_bindings_update:w:ow+rt' THEN
    RAISE EXCEPTION 'override-reading policies are %', actual;
  END IF;
  -- The runtime/owner policies naming pg_trigger_depth(): the three cascade branches plus the accepted 00014 bootstrap
  -- insert and the 00016 owner-only restriction insert; nothing else. role_bindings_update names it in USING only
  -- (attempt3: the cascade's lock), never in WITH CHECK.
  SELECT string_agg(p.polname, ',' ORDER BY p.polname) INTO actual
  FROM pg_policy p
  WHERE (rt::oid = ANY (p.polroles) OR ow::oid = ANY (p.polroles))
    AND (coalesce(pg_get_expr(p.polqual, p.polrelid), '') || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), ''))
        LIKE '%pg_trigger_depth()%';
  IF actual IS DISTINCT FROM 'enrollment_restrictions_owner_insert,role_bindings_delete,role_bindings_insert_bootstrap,'
    || 'role_bindings_select,role_bindings_update' THEN
    RAISE EXCEPTION 'trigger-depth runtime policies are %', actual;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_policy p WHERE p.polrelid = 'app.role_bindings'::regclass AND p.polname = 'role_bindings_update'
               AND (coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '') LIKE '%pg_trigger_depth()%'
                    OR coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '') NOT LIKE '%group_management_override%'
                    OR coalesce(pg_get_expr(p.polqual, p.polrelid), '') NOT LIKE '%pg_trigger_depth()%'
                    OR coalesce(pg_get_expr(p.polqual, p.polrelid), '') NOT LIKE '%group_management_override%')) THEN
    RAISE EXCEPTION 'role_bindings_update must admit the trigger-depth branch in USING only';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_policy p WHERE p.polrelid = 'app.stored_objects'::regclass
               AND p.polname IN ('stored_objects_select', 'stored_objects_insert', 'stored_objects_update')
               -- Each expression separately (USING then WITH CHECK concatenated would pair one's Group branch with the
               -- other's Organization branch).
               AND (coalesce(pg_get_expr(p.polqual, p.polrelid), '') LIKE '%group_id IS NOT NULL%org_owner%'
                    OR coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '') LIKE '%group_id IS NOT NULL%org_owner%'))
     OR (SELECT count(*) FROM pg_policy p WHERE p.polrelid = 'app.stored_objects'::regclass
           AND p.polname IN ('stored_objects_select', 'stored_objects_insert', 'stored_objects_update')) <> 3 THEN
    RAISE EXCEPTION 'stored object policies still admit an Organization role for Group objects';
  END IF;

  -- Postcondition 4: the replaced bodies carry the reviewed G060 fragments.
  SELECT prosrc INTO body FROM pg_proc WHERE oid = 'app.role_bindings_authority_guard()'::regprocedure;
  IF position('CONSTRAINT = ''self_escalation''' IN body) = 0
     OR position(format('current_user::text IN (%L, %L) AND OLD.scope_type = ''group''', rt::text, ow::text) IN body) = 0
     OR position(format('current_user::text NOT IN (%L, %L, %L)', rt::text, ow::text, cw::text) IN body) = 0
     OR position('(3b)' IN body) > position('(4) last owner' IN body) THEN
    RAISE EXCEPTION 'authority guard lacks the reviewed self-escalation step';
  END IF;
  IF position('CONSTRAINT = ''owner_subject''' IN body) = 0
     OR position('ORDER BY rb.group_id FOR NO KEY UPDATE LOOP' IN body) = 0
     OR position('ORDER BY rb.group_id FOR NO KEY UPDATE LOOP' IN body) > position('IF r.role = ''group_owner''' IN body)
     OR position('(3) Organization departure' IN body) = 0
     OR NOT (position('(3) Organization departure' IN body) < position('app.rls_dependent_group_bindings(OLD.subject_user_id) > 0' IN body)
             AND position('app.rls_dependent_group_bindings(OLD.subject_user_id) > 0' IN body) < position('(3b)' IN body))
     OR position('CONSTRAINT = ''context_required''' IN body) > position('ORDER BY rb.group_id' IN body) THEN
    RAISE EXCEPTION 'authority guard lacks the reviewed Organization-departure cascade';
  END IF;
  SELECT prosrc INTO body FROM pg_proc WHERE oid = 'app.invitation_claim_group(uuid, bytea)'::regprocedure;
  IF position('OR inv.issuer_user_id = u THEN' IN body) = 0
     OR position('PERFORM 1 FROM app.organizations og WHERE og.id = o FOR SHARE;' IN body) = 0
     OR position('FOR SHARE;' IN body) > position('FOR NO KEY UPDATE;' IN body)
     OR position('ORDER BY us.id FOR SHARE;' IN body) = 0
     OR position('FOR SHARE;' IN body) > position('app.rls_user_enabled_lock(u,' IN body)
     OR position('us.status = ''active''' IN body) = 0 THEN
    RAISE EXCEPTION 'invitation_claim_group lacks the reviewed issuer and lock steps';
  END IF;

  -- Postcondition 5: nothing is enabled at 00017. The owner is bound by FORCE, so FORCE is lifted only inside this
  -- transaction for the count and restored at once.
  ALTER TABLE app.organizations NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_on FROM app.organizations WHERE group_management_override;
  ALTER TABLE app.organizations FORCE ROW LEVEL SECURITY;
  IF n_on <> 0 THEN
    RAISE EXCEPTION 'override postcondition: % organizations enabled at 00017', n_on;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  n_on bigint;
  n_aud bigint;
BEGIN
  -- History is never discarded: refuse while any Organization has the override enabled or audit.events holds a row with
  -- an 00017 action, reason or operation. FORCE is lifted only inside this transaction; a refusal rolls the lift back.
  ALTER TABLE app.organizations NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_on FROM app.organizations WHERE group_management_override;
  SELECT count(*) INTO n_aud FROM audit.events
  WHERE action = 'access.override_used'
     OR reason IN ('group_management_override_enabled', 'group_management_override_disabled')
     OR operation = 'organization.set_group_override';
  IF n_on > 0 OR n_aud > 0 THEN
    RAISE EXCEPTION 'refusing to roll back 00017: % organizations with the override enabled / % audit rows (history is never discarded)',
      n_on, n_aud;
  END IF;
  ALTER TABLE app.organizations FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;
END
$do$;
-- +goose StatementEnd

-- The version-16 Organization capture (the 00011 body and trigger), restored.
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

-- The version-16 attribution body (00016), restored.
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
    'domain.state_changed', 'domain.checked', 'organization.branding_changed',
    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked');
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
    WHEN NEW.action = 'organization.branding_changed' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization'
      AND NEW.resource_id = NEW.organization_id AND NEW.reason IN ('updated', 'reset', 'logo_set', 'logo_removed')
    WHEN NEW.action IN ('invitation.issued', 'invitation.revoked', 'invitation.consumed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'invitation' AND NEW.resource_id IS NOT NULL
      AND CASE WHEN NEW.group_id IS NULL THEN NEW.role IN ('org_admin', 'org_member')
               ELSE NEW.role IN ('group_admin', 'group_member') END
      AND (NEW.action <> 'invitation.consumed' OR NEW.subject_user_id IS NOT NULL)
    WHEN NEW.action IN ('enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.group_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization' AND NEW.resource_id = NEW.organization_id
      AND CASE NEW.action
            WHEN 'enrollment.policy_changed' THEN NEW.reason IS NULL AND NEW.subject_user_id IS NULL
            WHEN 'enrollment.readiness_changed' THEN NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NULL
              AND NEW.reason IN ('operator_verified_absent', 'operator_backfilled', 'history_unresolved')
            ELSE NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NOT NULL AND NEW.reason = 'operator_backfilled' END
    WHEN NEW.action IN ('enrollment.requested', 'enrollment.decided', 'enrollment.revoked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'enrollment_request' AND NEW.resource_id IS NOT NULL AND NEW.actor_kind = 'user'
      AND (NEW.action <> 'enrollment.requested' OR NEW.actor_user_id = NEW.subject_user_id)
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
            'organization_not_member', 'host_changed', 'recipient_proof_unavailable'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
        'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch', 'recipient_ineligible',
        'session_invalid', 'host_changed')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object', 'organization_domain')))
      AND (NEW.reason IS DISTINCT FROM 'resource_mismatch'
           OR (NEW.operation = 'mcp.request' AND NEW.organization_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.reason NOT IN ('enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                              'recipient_ineligible', 'session_invalid', 'host_changed')
           OR (NEW.operation IN ('enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request',
                                 'enrollment.admit', 'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update',
                                 'enrollment.requests.list', 'enrollment.requests.decide', 'enrollment.requests.revoke')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.operation NOT IN ('enrollment.start', 'enrollment.admit')
           OR (NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL
               AND CASE WHEN NEW.operation = 'enrollment.start'
                        THEN NEW.reason IN ('enrollment_closed', 'invitation_mismatch')
                        ELSE NEW.actor_user_id IS NOT NULL AND NEW.reason IN ('target_not_found', 'enrollment_closed',
                             'enrollment_restricted', 'approval_revoked', 'recipient_ineligible', 'session_invalid',
                             'host_changed') END))
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

-- The version-16 catalog constraints (00016), restored.
ALTER TABLE audit.events DROP CONSTRAINT events_reason_valid, DROP CONSTRAINT events_operation_valid,
  DROP CONSTRAINT events_action_valid;
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
                    'organization.branding_changed',
                    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
                    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
                    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked',
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
                    'withdrawn', 'withdraw_error', 'withdraw_unknown', 'organization_not_member', 'host_changed',
                    'updated', 'reset', 'logo_set', 'logo_removed',
                    'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                    'recipient_proof_unavailable', 'recipient_ineligible', 'session_invalid',
                    'operator_verified_absent', 'operator_backfilled', 'history_unresolved')),
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug',
                    'branding.get', 'branding.update', 'branding.reset', 'branding.set_logo', 'branding.remove_logo',
                    'group.suggest_slug', 'group.create', 'group.list', 'group.onboarding.get',
                    'group.onboarding.complete_new', 'group.onboarding.complete_existing', 'business_profile.list',
                    'membership.list', 'membership.self', 'membership.change_role', 'membership.remove', 'membership.leave',
                    'ownership.promote', 'ownership.demote', 'ownership.transfer',
                    'invitation.list', 'invitation.mine', 'invitation.candidates', 'invitation.issue', 'invitation.revoke',
                    'invitation.get', 'invitation.accept', 'invitation.consume', 'retirement.eligibility',
                    'enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request', 'enrollment.admit',
                    'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update', 'enrollment.requests.list',
                    'enrollment.requests.decide', 'enrollment.requests.revoke'));

-- The version-16 policies, bodies, guard and grants, restored; then the column.
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
  cw text := current_database() || '_claim_writer';
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  MGR text := '(SELECT app.rls_org_role()) IN (''org_owner'', ''org_admin'')';
  GMGR text := 'app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'')';
  manage text;
  linked text;
  vis text;
  scope text;
  w text;
BEGIN
  DROP TRIGGER organizations_group_override_guard ON app.organizations;
  DROP FUNCTION app.organizations_group_override_guard();

  -- 00015 role_bindings_update.
  DROP POLICY role_bindings_update ON app.role_bindings;
  manage := format('organization_id = %s AND source = ''membership'' AND ('
    || '(subject_user_id = %s AND %s IS NOT NULL AND (scope_type = ''organization'' OR %s IS NULL OR group_id = %s)) '
    || 'OR (scope_type = ''organization'' AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR %s)))', O, U, R, G, G, MGR, G, G, MGR, GMGR);
  EXECUTE format('CREATE POLICY role_bindings_update ON app.role_bindings FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND source = ''membership'' AND ((scope_type = ''organization'' AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR %s))))', rt, ow, manage, O, MGR, G, G, MGR, GMGR);
  -- 00015 role_bindings_delete (the same manage USING) and 00006 role_bindings_select (amendment01 A1.3 item 5).
  DROP POLICY role_bindings_delete ON app.role_bindings;
  EXECUTE format('CREATE POLICY role_bindings_delete ON app.role_bindings FOR DELETE TO %I, %I USING (%s)', rt, ow, manage);
  DROP POLICY role_bindings_select ON app.role_bindings;
  EXECUTE format('CREATE POLICY role_bindings_select ON app.role_bindings FOR SELECT TO %I, %I USING ('
    || '%s IS NOT NULL AND organization_id = %s AND source = ''membership'' AND ('
    || '(subject_user_id = %s AND (scope_type = ''organization'' OR %s IS NULL OR group_id = %s))'
    || ' OR (scope_type = ''organization'' AND %s)'
    || ' OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'')))))',
    rt, ow, R, O, U, G, G, MGR, G, G, MGR);

  -- 00006 groups_update.
  DROP POLICY groups_update ON app.groups;
  vis := format('organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR id = %s) AND (%s OR app.rls_group_role(id) IS NOT NULL)',
    O, R, G, G, MGR);
  EXECUTE format('CREATE POLICY groups_update ON app.groups FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR id = %s) AND (%s OR app.rls_group_role(id) IN (''group_owner'', ''group_admin'')))',
    rt, ow, vis, O, R, G, G, MGR);

  -- 00014 group_business_profiles_insert and business_profiles_insert.
  DROP POLICY group_business_profiles_insert ON app.group_business_profiles;
  EXECUTE format('CREATE POLICY group_business_profiles_insert ON app.group_business_profiles FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR group_id = %s) AND (%s OR ('
    || 'app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'') AND app.rls_profile_eligible(business_profile_id))))',
    rt, ow, O, R, G, G, MGR);
  DROP POLICY business_profiles_insert ON app.business_profiles;
  EXECUTE format('CREATE POLICY business_profiles_insert ON app.business_profiles FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s OR EXISTS (SELECT 1 FROM app.role_bindings rb '
    || 'WHERE rb.subject_user_id = %s AND rb.organization_id = %s AND rb.scope_type = ''group'' '
    || 'AND rb.source = ''membership'' AND rb.role IN (''group_owner'', ''group_admin'') AND (%s IS NULL OR rb.scope_id = %s))))',
    rt, ow, O, R, MGR, U, O, G, G);

  -- 00006 business_profiles_update.
  DROP POLICY business_profiles_update ON app.business_profiles;
  linked := 'EXISTS (SELECT 1 FROM app.group_business_profiles a WHERE a.organization_id = business_profiles.organization_id '
    || 'AND a.business_profile_id = business_profiles.id AND (' || G || ' IS NULL OR a.group_id = ' || G || ') '
    || 'AND app.rls_group_role(a.group_id) IN (%s))';
  vis := format('organization_id = %s AND %s IS NOT NULL AND (%s OR %s)', O, R, MGR,
    format(linked, '''group_owner'', ''group_admin'', ''group_member'''));
  EXECUTE format('CREATE POLICY business_profiles_update ON app.business_profiles FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s OR ((SELECT o.group_admins_edit_business_profiles FROM app.organizations o '
    || 'WHERE o.id = %s) AND %s)))',
    rt, ow, vis, O, R, MGR, O, format(linked, '''group_owner'', ''group_admin'''));

  -- 00007 stored_objects_select/insert/update.
  DROP POLICY stored_objects_select ON app.stored_objects;
  DROP POLICY stored_objects_insert ON app.stored_objects;
  DROP POLICY stored_objects_update ON app.stored_objects;
  scope := format('organization_id = %s AND %s IS NOT NULL AND (group_id IS NULL OR ((%s IS NULL OR group_id = %s) AND '
    || '(%s OR app.rls_group_role(group_id) IS NOT NULL)))', O, R, G, G, MGR);
  EXECUTE format('CREATE POLICY stored_objects_select ON app.stored_objects FOR SELECT TO %I, %I USING (%s)', rt, ow, scope);
  EXECUTE format('CREATE POLICY stored_objects_insert ON app.stored_objects FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND creator_user_id = %s AND state = ''pending'' AND ('
    || '(group_id IS NULL AND %s) OR (group_id IS NOT NULL AND (%s IS NULL OR group_id = %s) AND '
    || '(%s OR app.rls_group_role(group_id) IS NOT NULL))))', rt, ow, O, R, U, MGR, G, G, MGR);
  w := format('organization_id = %s AND %s IS NOT NULL AND ((group_id IS NULL AND %s) OR (group_id IS NOT NULL AND '
    || '(%s IS NULL OR group_id = %s) AND (%s OR app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'') OR '
    || '(creator_user_id = %s AND app.rls_group_role(group_id) IS NOT NULL))))', O, R, MGR, G, G, MGR, U);
  EXECUTE format('CREATE POLICY stored_objects_update ON app.stored_objects FOR UPDATE TO %I, %I USING (%s) WITH CHECK (%s)',
    rt, ow, w, w);

  -- 00015 invitations_select/insert/update.
  DROP POLICY invitations_select ON app.invitations;
  DROP POLICY invitations_insert ON app.invitations;
  DROP POLICY invitations_update ON app.invitations;
  manage := format('((scope_type = ''organization'' AND %s) OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) '
    || 'AND (%s OR %s)))', MGR, G, G, MGR, GMGR);
  EXECUTE format('CREATE POLICY invitations_select ON app.invitations FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND (%s OR intended_user_id = %s OR consumed_by_user_id = %s))', rt, ow, O, manage, U, U);
  EXECUTE format('CREATE POLICY invitations_insert ON app.invitations FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND issuer_user_id = %s AND ((scope_type = ''organization'' AND %s IS NULL AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR %s) '
    || 'AND app.is_organization_member(intended_user_id, %s))))', rt, ow, O, U, G, MGR, G, G, MGR, GMGR, O);
  EXECUTE format('CREATE POLICY invitations_update ON app.invitations FOR UPDATE TO %I, %I USING ('
    || 'organization_id = %s AND %s) WITH CHECK (organization_id = %s AND %s)', rt, ow, O, manage, O, manage);

  -- The 00015 policy-reader bodies.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pr);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.rls_owner_count(p_scope_type text, p_scope_id uuid) RETURNS integer
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text := app.rls_org_role();
BEGIN
  IF app.ctx_user() IS NULL OR o IS NULL OR r IS NULL OR p_scope_id IS NULL THEN
    RETURN 0;
  END IF;
  IF p_scope_type = 'organization' THEN
    IF p_scope_id <> o THEN
      RETURN 0;
    END IF;
    RETURN (SELECT count(*) FROM app.role_bindings rb
            WHERE rb.organization_id = o AND rb.scope_type = 'organization' AND rb.scope_id = o
              AND rb.role = 'org_owner' AND rb.source = 'membership');
  ELSIF p_scope_type = 'group' THEN
    IF (g IS NOT NULL AND p_scope_id <> g) OR NOT (r IN ('org_owner', 'org_admin') OR app.rls_group_role(p_scope_id) IS NOT NULL) THEN
      RETURN 0;
    END IF;
    RETURN (SELECT count(*) FROM app.role_bindings rb
            WHERE rb.organization_id = o AND rb.scope_type = 'group' AND rb.scope_id = p_scope_id
              AND rb.role = 'group_owner' AND rb.source = 'membership');
  END IF;
  RETURN 0;
END $$
$r$;
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.rls_group_invitation_candidates(p_group uuid, p_after uuid, p_limit integer)
  RETURNS TABLE (user_id uuid, display_name text)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text := app.rls_org_role();
BEGIN
  IF app.ctx_user() IS NULL OR o IS NULL OR r IS NULL OR p_group IS NULL OR p_limit IS NULL OR p_limit < 1
     OR (g IS NOT NULL AND g <> p_group)
     OR NOT (r IN ('org_owner', 'org_admin') OR app.rls_group_role(p_group) IN ('group_owner', 'group_admin'))
     OR NOT EXISTS (SELECT 1 FROM app.role_bindings x WHERE x.organization_id = o AND x.scope_type = 'group'
                      AND x.scope_id = p_group AND x.source = 'membership') THEN
    RETURN;
  END IF;
  RETURN QUERY
  SELECT us.id, us.display_name FROM app.role_bindings rb JOIN app.users us ON us.id = rb.subject_user_id
  WHERE rb.organization_id = o AND rb.scope_type = 'organization' AND rb.scope_id = o AND rb.source = 'membership'
    AND us.status = 'active' AND (p_after IS NULL OR us.id > p_after)
    AND NOT EXISTS (SELECT 1 FROM app.role_bindings gb WHERE gb.subject_user_id = rb.subject_user_id
                      AND gb.organization_id = o AND gb.scope_type = 'group' AND gb.scope_id = p_group)
  ORDER BY us.id
  LIMIT least(p_limit, 101);
END $$
$r$;
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pr);

  -- The 00015 claim-writer bodies.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', cw);
  EXECUTE format('SET LOCAL ROLE %I', cw);
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.rls_user_enabled_lock(p_user uuid, p_scope_type text, p_scope_id uuid) RETURNS boolean
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  admitted boolean := false;
  st text;
BEGIN
  IF u IS NULL OR p_user IS NULL THEN
    RETURN NULL;
  END IF;
  IF p_user = u THEN
    admitted := true;
  ELSIF o IS NOT NULL AND p_scope_id IS NOT NULL AND EXISTS (
          SELECT 1 FROM app.role_bindings rb WHERE rb.subject_user_id = p_user AND rb.organization_id = o
            AND rb.scope_type = p_scope_type AND rb.scope_id = p_scope_id AND rb.source = 'membership') THEN
    IF p_scope_type = 'organization' THEN
      admitted := p_scope_id = o AND coalesce(app.rls_org_role() IN ('org_owner', 'org_admin'), false);
    ELSIF p_scope_type = 'group' THEN
      admitted := (g IS NULL OR g = p_scope_id)
        AND coalesce(app.rls_org_role() IN ('org_owner', 'org_admin')
                     OR app.rls_group_role(p_scope_id) IN ('group_owner', 'group_admin'), false);
    END IF;
  END IF;
  IF NOT admitted THEN
    RETURN NULL;
  END IF;
  PERFORM 1 FROM app.users us WHERE us.id = p_user FOR SHARE;
  SELECT us.status INTO st FROM app.users us WHERE us.id = p_user;
  RETURN st = 'active';
END $$
$r$;
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.invitation_claim_group(p_id uuid, p_digest bytea)
  RETURNS TABLE (outcome text, organization_id uuid, scope_type text, scope_id uuid, group_id uuid, role text, binding_id uuid)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  inv app.invitations%ROWTYPE;
  ex_id uuid;
  ex_role text;
  bid uuid;
  res text;
BEGIN
  IF u IS NULL OR o IS NULL THEN
    RAISE EXCEPTION 'invitation claim context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'invitation_context';
  END IF;
  -- (1) serialize duplicate claims on the invitation row.
  SELECT i.* INTO inv FROM app.invitations i WHERE i.id = p_id FOR NO KEY UPDATE;
  -- (2) digest, kind, tenant, selected scope and recipient intent; any mismatch is the one closed not_found.
  IF NOT FOUND OR p_digest IS NULL OR inv.token_digest <> p_digest OR inv.scope_type <> 'group'
     OR inv.organization_id <> o OR (g IS NOT NULL AND g <> inv.group_id) OR inv.intended_user_id IS DISTINCT FROM u THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF inv.consumed_at IS NOT NULL THEN
    -- Replay by the same recipient: already_member only with a current qualifying binding; never restored.
    IF inv.consumed_by_user_id = u THEN
      SELECT rb.id, rb.role INTO ex_id, ex_role FROM app.role_bindings rb
      WHERE rb.subject_user_id = u AND rb.organization_id = inv.organization_id AND rb.scope_type = inv.scope_type
        AND rb.scope_id = inv.scope_id AND rb.source = 'membership';
      IF ex_id IS NOT NULL AND (CASE ex_role WHEN 'group_owner' THEN 3 WHEN 'group_admin' THEN 2 ELSE 1 END)
                               >= (CASE inv.role WHEN 'group_admin' THEN 2 ELSE 1 END) THEN
        RETURN QUERY SELECT 'already_member'::text, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, ex_role, ex_id;
      ELSE
        RETURN QUERY SELECT 'consumed'::text, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, inv.role, NULL::uuid;
      END IF;
      RETURN;
    END IF;
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF inv.revoked_at IS NOT NULL OR inv.expires_at <= now() THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  -- (3) the enabled recipient under a lock consistent with DisableUser.
  IF NOT coalesce(app.rls_user_enabled_lock(u, inv.scope_type, inv.scope_id), false) THEN
    RAISE EXCEPTION 'invitation recipient ineligible'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
  END IF;
  -- (4) lock the issuer's and the recipient's bindings for the scope (ascending id), then check with the post-lock
  -- snapshot: the issuer's current authority and the recipient's current Organization membership.
  PERFORM 1 FROM app.role_bindings rb
  WHERE rb.organization_id = inv.organization_id AND rb.source = 'membership'
    AND rb.subject_user_id IN (inv.issuer_user_id, u)
    AND (rb.scope_type = 'organization' OR rb.scope_id = inv.scope_id)
  ORDER BY rb.id FOR SHARE;
  IF NOT EXISTS (SELECT 1 FROM app.role_bindings rb
                 WHERE rb.subject_user_id = inv.issuer_user_id AND rb.organization_id = inv.organization_id
                   AND rb.source = 'membership'
                   AND ((rb.scope_type = 'organization' AND rb.role IN ('org_owner', 'org_admin'))
                     OR (rb.scope_type = 'group' AND rb.scope_id = inv.scope_id AND rb.role IN ('group_owner', 'group_admin')))) THEN
    RAISE EXCEPTION 'invitation issuer authority lost'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'issuer_authority_lost';
  END IF;
  IF NOT app.is_organization_member(u, inv.organization_id) THEN
    RAISE EXCEPTION 'invitation recipient ineligible'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
  END IF;
  -- (5) existing scope binding: equal/higher -> already_member; lower -> upgrade; none -> insert from the intent.
  SELECT rb.id, rb.role INTO ex_id, ex_role FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.organization_id = inv.organization_id AND rb.scope_type = inv.scope_type
    AND rb.scope_id = inv.scope_id AND rb.source = 'membership';
  IF ex_id IS NOT NULL THEN
    IF (CASE ex_role WHEN 'group_owner' THEN 3 WHEN 'group_admin' THEN 2 ELSE 1 END)
       >= (CASE inv.role WHEN 'group_admin' THEN 2 ELSE 1 END) THEN
      res := 'already_member';
    ELSE
      UPDATE app.role_bindings rb SET role = inv.role WHERE rb.id = ex_id;
      res := 'upgraded'; ex_role := inv.role;
    END IF;
    bid := ex_id;
  ELSE
    bid := gen_random_uuid();
    INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
    VALUES (bid, inv.organization_id, u, inv.scope_type, inv.scope_id, inv.group_id, inv.role, 'membership');
    res := 'joined'; ex_role := inv.role;
  END IF;
  -- (6) consume.
  UPDATE app.invitations i SET consumed_at = now(), consumed_by_user_id = u, consumed_binding_id = bid WHERE i.id = p_id;
  RETURN QUERY SELECT res, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, ex_role, bid;
END $$
$r$;
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', cw);

  -- The 00015 authority guard.
  EXECUTE format($f$
CREATE OR REPLACE FUNCTION app.role_bindings_authority_guard() RETURNS trigger
  LANGUAGE plpgsql VOLATILE SET search_path = pg_catalog AS $b$
DECLARE
  old_owner boolean;
  new_owner boolean;
BEGIN
  IF current_user::text NOT IN (%L, %L, %L) THEN
    IF TG_OP = 'DELETE' THEN
      RETURN OLD;
    END IF;
    RETURN NEW;
  END IF;
  old_owner := OLD.role IN ('org_owner', 'group_owner');
  new_owner := TG_OP = 'UPDATE' AND NEW.role IN ('org_owner', 'group_owner');
  -- (1) owner authority: no Organization-role override of Group ownership.
  IF old_owner OR new_owner THEN
    IF OLD.scope_type = 'organization' THEN
      IF app.rls_org_role() IS DISTINCT FROM 'org_owner' THEN
        RAISE EXCEPTION 'owner authority required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'owner_authority';
      END IF;
    ELSIF app.rls_group_role(OLD.group_id) IS DISTINCT FROM 'group_owner' THEN
      RAISE EXCEPTION 'owner authority required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'owner_authority';
    END IF;
  END IF;
  -- (2) promotion eligibility: same-scope admin, enabled under a lock consistent with DisableUser, and (Group) a
  -- current Organization member.
  IF NOT old_owner AND new_owner THEN
    IF OLD.role IS DISTINCT FROM (CASE OLD.scope_type WHEN 'organization' THEN 'org_admin' ELSE 'group_admin' END)
       OR NOT coalesce(app.rls_user_enabled_lock(OLD.subject_user_id, OLD.scope_type, OLD.scope_id), false)
       OR (OLD.scope_type = 'group' AND NOT app.is_organization_member(OLD.subject_user_id, OLD.organization_id)) THEN
      RAISE EXCEPTION 'promotion recipient ineligible'
        USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
    END IF;
  END IF;
  -- (3) Organization departure: an Organization-only context, and no dependent Group binding (the definer count sees
  -- every sibling Group regardless of the selected context).
  IF TG_OP = 'DELETE' AND OLD.scope_type = 'organization' THEN
    IF app.ctx_group() IS NOT NULL THEN
      RAISE EXCEPTION 'organization-only context required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'context_required';
    END IF;
    IF app.rls_dependent_group_bindings(OLD.subject_user_id) > 0 THEN
      RAISE EXCEPTION 'dependent group bindings remain'
        USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'dependent_bindings';
    END IF;
  END IF;
  -- (4) last owner: command A locks the scope row; command B counts with the snapshot taken after A's wait.
  IF old_owner AND NOT new_owner THEN
    IF OLD.scope_type = 'organization' THEN
      PERFORM 1 FROM app.organizations o WHERE o.id = OLD.scope_id FOR NO KEY UPDATE;
    ELSE
      PERFORM 1 FROM app.groups g WHERE g.id = OLD.scope_id FOR NO KEY UPDATE;
    END IF;
    IF app.rls_owner_count(OLD.scope_type, OLD.scope_id) - 1 < 1 THEN
      RAISE EXCEPTION 'last owner' USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'last_owner';
    END IF;
  END IF;
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END $b$
$f$, rt, ow, cw);

  -- The three column grants, then the column (no policy, trigger or body references it any more).
  EXECUTE format('REVOKE UPDATE (group_management_override) ON app.organizations FROM %I', rt);
  EXECUTE format('REVOKE SELECT (group_management_override) ON app.organizations FROM %I', cw);
  EXECUTE format('REVOKE SELECT (group_management_override) ON app.organizations FROM %I', pr);
  ALTER TABLE app.organizations DROP COLUMN group_management_override;
END
$do$;
-- +goose StatementEnd
