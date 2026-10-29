-- Generic labels and visibility (G020 Plan r2 §1.2, §1.4, §3; Spec §1-§2; docs/foundation/TENANCY.md, AUDIT.md,
-- DATABASE-ISOLATION.md).
--
-- Additive, one goose transaction:
--   - Precondition: the owner, the five roles with the 00018/00019 attribute rules, version 19, no label object, the
--     version-19 shapes this migration touches or relies on (the 00019 audit.events_attribute body by md5(prosrc), the
--     four catalog constraints at 47 actions / 79 reasons / 73 operations / 11 resource types, 79 relations in schema
--     app, FORCE on app.organizations, app.groups and audit.events) and the helper inventories (policy reader 9, claim
--     writer 7, auth reader 8, provision writer 3).
--   - app.label_definitions (Organization- or Group-owned; visibility owner_scope | target_members, the latter only for
--     Group-owned definitions; name 1..80 code points, no C0/DEL/C1 control, no U+2028/U+2029, no ASCII edge space;
--     ASCII-case-insensitive scoped uniqueness through the generated name_key) and app.label_assignments (production
--     storage is closed to an Organization-owned definition on a same-Organization Group target: the total composite
--     foreign key to the definition, the Group foreign key and the closed target-shape CHECK). Nine relations (79 -> 88).
--   - Five triggers: the two owner-owned SECURITY INVOKER keys-immutable guards (SQLSTATE 23000), the 200-per-owner-scope
--     bound (scope row FOR NO KEY UPDATE, then count; the invitations_bound pattern) and the two strict audit captures.
--   - Runtime: column-limited INSERT/UPDATE on definitions, column-limited INSERT on assignments, SELECT and DELETE on
--     both, no assignment UPDATE. No helper, PUBLIC, retention or migrator grant, no new role, no SECURITY DEFINER.
--   - Seven command-specific policies TO runtime, owner: Organization-owned definitions and their assignments only for
--     the Organization owner/admin; Group-owned definitions only for actual roles in the exact owning Group (managers;
--     members only for target_members). No policy reads the Group-management override or pg_trigger_depth().
--   - Audit (D-4): five trigger-owned actions on the new resource type label (the definition id; the owning Group of a
--     Group-owned definition or the assignment's target Group): label.created, label.updated (reason renamed or
--     visibility_changed), label.deleted, label.assigned, label.unassigned. Never a name, a target type or a non-Group
--     target id; a no-op update records nothing. Actions 47 -> 52, reasons 79 -> 81, operations 73 -> 80, resource types
--     11 -> 12; events_attribute admits the five actions as trigger-owned under their shape rule and label as an
--     access.denied resource type.
--   - Postconditions.
--
-- Locks: ACCESS EXCLUSIVE on the two new tables, SHARE ROW EXCLUSIVE on app.organizations and app.groups (foreign keys)
-- and one validating scan of audit.events per swapped catalog constraint. Migrate, then deploy: version-19 code never
-- names the tables; version-20 code requires version 20.
--
-- Down (development and test only; the runner refuses down/reapply with APP_ENV=production) refuses while any label row
-- exists or audit.events holds a label action, operation, reason or resource (inspected with FORCE lifted inside the
-- transaction; a refusal rolls the lift back), then restores the exact version-19 shapes.

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
  pw regrole;
  r record;
  n bigint;
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regrole(db || '_runtime') IS NULL OR to_regrole(db || '_policy_reader') IS NULL OR to_regrole(db || '_auth_reader') IS NULL
     OR to_regrole(db || '_claim_writer') IS NULL OR to_regrole(db || '_provision_writer') IS NULL THEN
    RAISE EXCEPTION 'roles for database % are missing: run infra/postgres/provision-roles.sql (or a fresh 10-roles.sh) first', db;
  END IF;
  ow := current_user::regrole; rt := (db || '_runtime')::regrole; pr := (db || '_policy_reader')::regrole;
  ar := (db || '_auth_reader')::regrole; cw := (db || '_claim_writer')::regrole; pw := (db || '_provision_writer')::regrole;

  -- The 00018/00019 role checks, unchanged.
  FOR r IN SELECT rolname, rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolinherit, rolcanlogin, rolreplication
           FROM pg_roles WHERE oid IN (pr, ar, cw, pw) LOOP
    IF r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication THEN
      RAISE EXCEPTION 'role % has an unsafe attribute: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
    IF r.rolcanlogin OR r.rolinherit THEN
      RAISE EXCEPTION 'helper role % must be NOLOGIN NOINHERIT: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
  END LOOP;
  IF EXISTS (SELECT 1 FROM pg_roles WHERE oid = rt AND (rolsuper OR rolbypassrls OR rolcreaterole OR rolcreatedb OR rolreplication))
     OR EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member IN (rt, cw, pw)) THEN
    RAISE EXCEPTION 'the runtime role has an unsafe attribute or the runtime, claim-writer or provision-writer role is a member of a role';
  END IF;

  -- Version 19: the migration history ends at 19, no label object exists, the touched shapes are the version-19 ones.
  IF (SELECT max(version_id) FROM app.goose_db_version) IS DISTINCT FROM 19 THEN
    RAISE EXCEPTION 'version-19 precondition: the current version is %', (SELECT max(version_id) FROM app.goose_db_version);
  END IF;
  IF to_regclass('app.label_definitions') IS NOT NULL OR to_regclass('app.label_assignments') IS NOT NULL
     OR to_regprocedure('app.label_definitions_keys_immutable()') IS NOT NULL
     OR to_regprocedure('app.label_assignments_keys_immutable()') IS NOT NULL
     OR to_regprocedure('app.label_definitions_bound()') IS NOT NULL
     OR to_regprocedure('audit.capture_label_definitions()') IS NOT NULL
     OR to_regprocedure('audit.capture_label_assignments()') IS NOT NULL
     OR EXISTS (SELECT 1 FROM pg_policy WHERE polname IN ('label_definitions_select', 'label_definitions_insert',
                  'label_definitions_update', 'label_definitions_delete', 'label_assignments_select', 'label_assignments_insert',
                  'label_assignments_delete')) THEN
    RAISE EXCEPTION 'version-19 precondition: label objects already exist';
  END IF;
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 79 THEN
    RAISE EXCEPTION 'version-19 precondition: schema app holds % relations, expected 79', n;
  END IF;
  -- The four catalog constraints name exactly the version-19 values (one '::text' per value) and no label value.
  n := 0;
  FOR r IN SELECT c.conname::text AS conname, pg_get_constraintdef(c.oid) AS def FROM pg_constraint c
           WHERE c.conrelid = 'audit.events'::regclass
             AND c.conname IN ('events_action_valid', 'events_reason_valid', 'events_operation_valid', 'events_resource_type_valid') LOOP
    n := n + 1;
    -- The CASE is parenthesized: a PL/pgSQL IF condition ends at the first THEN outside parentheses (00017 precedent).
    IF (length(r.def) - length(replace(r.def, '::text', ''))) / length('::text') IS DISTINCT FROM
         (CASE r.conname WHEN 'events_action_valid' THEN 47 WHEN 'events_reason_valid' THEN 79 WHEN 'events_operation_valid' THEN 73 ELSE 11 END)
       OR position('''label' IN r.def) > 0 OR position('''renamed''' IN r.def) > 0 OR position('''visibility_changed''' IN r.def) > 0 THEN
      RAISE EXCEPTION 'version-19 precondition: the catalog constraint % differs', r.conname;
    END IF;
  END LOOP;
  IF n <> 4 THEN
    RAISE EXCEPTION 'version-19 precondition: % of the four catalog constraints exist', n;
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = pr) <> 9 OR (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7
     OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 8 OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 3 THEN
    RAISE EXCEPTION 'version-19 precondition: the helper inventories differ';
  END IF;
  -- The body this migration replaces: 00019's attribution.
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM 'ce481e2677b791e5834a5f3da1af8956'
     OR EXISTS (SELECT 1 FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure AND (prosecdef OR proowner <> ow)) THEN
    RAISE EXCEPTION 'version-19 precondition: the attribution body differs';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_class WHERE oid IN ('app.organizations'::regclass, 'app.groups'::regclass, 'audit.events'::regclass)
               AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION 'version-19 precondition: FORCE on app.organizations, app.groups or audit.events differs';
  END IF;
END
$$;
-- +goose StatementEnd

-- LabelDefinition (Plan r2 §1.2). The owner facts are non-nullable in both scopes, so both foreign keys check every row.
CREATE TABLE app.label_definitions (
  id               uuid        NOT NULL,
  organization_id  uuid        NOT NULL,
  owner_scope_type text        NOT NULL,
  owner_scope_id   uuid        NOT NULL,
  group_id         uuid,
  name             text        NOT NULL,
  name_key         text        GENERATED ALWAYS AS (lower(name COLLATE "C")) STORED,
  visibility       text        NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT label_definitions_pkey PRIMARY KEY (id),
  CONSTRAINT label_definitions_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT label_definitions_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  -- MATCH SIMPLE: unchecked only when group_id IS NULL, which label_definitions_scope_consistent admits only for the
  -- Organization scope (whose owner is checked by label_definitions_organization_fkey).
  CONSTRAINT label_definitions_group_fkey FOREIGN KEY (organization_id, group_id)
    REFERENCES app.groups (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT label_definitions_scope_consistent CHECK (
       (owner_scope_type = 'organization' AND owner_scope_id = organization_id AND group_id IS NULL)
    OR (owner_scope_type = 'group' AND group_id IS NOT NULL AND owner_scope_id = group_id)),
  -- Organization-owned definitions can never opt into target-member visibility (Spec §2).
  CONSTRAINT label_definitions_visibility_valid CHECK (visibility IN ('owner_scope', 'target_members')
    AND (visibility = 'owner_scope' OR owner_scope_type = 'group')),
  -- The G063 stored-shape grammar with the D-5 bound: 1..80 code points, no ASCII edge space, no C0/DEL/C1 control and no
  -- U+2028/U+2029. Validated, never normalized.
  CONSTRAINT label_definitions_name_valid CHECK (char_length(name) BETWEEN 1 AND 80 AND name = btrim(name)
    AND (name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f\u2028\u2029]'),
  -- ASCII-case-insensitive, otherwise byte-exact uniqueness within the owner scope.
  CONSTRAINT label_definitions_scope_name_key UNIQUE (organization_id, owner_scope_type, owner_scope_id, name_key),
  -- The composite target of label_assignments_label_fkey (the business_profiles_organization_id_key precedent).
  CONSTRAINT label_definitions_owner_key UNIQUE (organization_id, id, owner_scope_type, owner_scope_id)
);
ALTER TABLE app.label_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.label_definitions FORCE ROW LEVEL SECURITY;

-- LabelAssignment (Plan r2 §1.2, §1.4). Every referencing column of the definition key is NOT NULL, so the composite
-- foreign key is total; the closed CHECK admits exactly the production target kind (an Organization-owned definition on
-- a same-Organization Group, whose existence label_assignments_target_group_fkey enforces). A future target type widens
-- the CHECK, adds its own per-type key column and foreign key and a SELECT disjunct in its own reviewed migration.
CREATE TABLE app.label_assignments (
  id               uuid        NOT NULL,
  organization_id  uuid        NOT NULL,
  label_id         uuid        NOT NULL,
  owner_scope_type text        NOT NULL,
  owner_scope_id   uuid        NOT NULL,
  target_type      text        NOT NULL,
  target_id        uuid        NOT NULL,
  target_group_id  uuid        NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT label_assignments_pkey PRIMARY KEY (id),
  CONSTRAINT label_assignments_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT label_assignments_label_fkey FOREIGN KEY (organization_id, label_id, owner_scope_type, owner_scope_id)
    REFERENCES app.label_definitions (organization_id, id, owner_scope_type, owner_scope_id) ON DELETE RESTRICT,
  CONSTRAINT label_assignments_target_group_fkey FOREIGN KEY (organization_id, target_group_id)
    REFERENCES app.groups (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT label_assignments_target_shape CHECK (owner_scope_type = 'organization' AND owner_scope_id = organization_id
    AND target_type = 'group' AND target_group_id = target_id),
  CONSTRAINT label_assignments_unique UNIQUE (label_id, target_type, target_id)
);
CREATE INDEX label_assignments_target_idx ON app.label_assignments (organization_id, target_type, target_id);
CREATE INDEX label_assignments_group_idx ON app.label_assignments (organization_id, target_group_id);
ALTER TABLE app.label_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.label_assignments FORCE ROW LEVEL SECURITY;

-- Immutability (owner-owned invoker; no early return for any role): a definition's owner facts and identity never change
-- (only name and visibility do); an assignment is never updated, only inserted or deleted.
-- +goose StatementBegin
CREATE FUNCTION app.label_definitions_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF (NEW.id, NEW.organization_id, NEW.owner_scope_type, NEW.owner_scope_id, NEW.group_id, NEW.created_at)
     IS DISTINCT FROM (OLD.id, OLD.organization_id, OLD.owner_scope_type, OLD.owner_scope_id, OLD.group_id, OLD.created_at) THEN
    RAISE EXCEPTION 'label definition keys are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'label_definitions_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.label_definitions_keys_immutable() FROM PUBLIC;
CREATE TRIGGER label_definitions_keys_immutable BEFORE UPDATE ON app.label_definitions
  FOR EACH ROW EXECUTE FUNCTION app.label_definitions_keys_immutable();

-- +goose StatementBegin
CREATE FUNCTION app.label_assignments_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF (NEW.id, NEW.organization_id, NEW.label_id, NEW.owner_scope_type, NEW.owner_scope_id, NEW.target_type, NEW.target_id,
      NEW.target_group_id, NEW.created_at)
     IS DISTINCT FROM (OLD.id, OLD.organization_id, OLD.label_id, OLD.owner_scope_type, OLD.owner_scope_id, OLD.target_type,
      OLD.target_id, OLD.target_group_id, OLD.created_at) THEN
    RAISE EXCEPTION 'label assignments are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'label_assignments_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.label_assignments_keys_immutable() FROM PUBLIC;
CREATE TRIGGER label_assignments_keys_immutable BEFORE UPDATE ON app.label_assignments
  FOR EACH ROW EXECUTE FUNCTION app.label_assignments_keys_immutable();

-- The audit captures (owner-owned SECURITY INVOKER; an audit fault refuses the write). The definition id is the
-- resource; the owning Group (Group-owned definitions) or the target Group (assignments) the group_id. Never a name, a
-- target type or a non-Group target id; a no-op update records nothing.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_label_definitions() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('label.created', 'success', NEW.organization_id, NEW.group_id, 'label', NEW.id);
  ELSIF TG_OP = 'UPDATE' THEN
    IF NEW.name IS DISTINCT FROM OLD.name THEN
      INSERT INTO audit.events (action, outcome, reason, organization_id, group_id, resource_type, resource_id)
      VALUES ('label.updated', 'success', 'renamed', NEW.organization_id, NEW.group_id, 'label', NEW.id);
    END IF;
    IF NEW.visibility IS DISTINCT FROM OLD.visibility THEN
      INSERT INTO audit.events (action, outcome, reason, organization_id, group_id, resource_type, resource_id)
      VALUES ('label.updated', 'success', 'visibility_changed', NEW.organization_id, NEW.group_id, 'label', NEW.id);
    END IF;
  ELSE
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('label.deleted', 'success', OLD.organization_id, OLD.group_id, 'label', OLD.id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION audit.capture_label_definitions() FROM PUBLIC;
CREATE TRIGGER audit_capture_label_definitions AFTER INSERT OR UPDATE OF name, visibility OR DELETE ON app.label_definitions
  FOR EACH ROW EXECUTE FUNCTION audit.capture_label_definitions();

-- +goose StatementBegin
CREATE FUNCTION audit.capture_label_assignments() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('label.assigned', 'success', NEW.organization_id, NEW.target_group_id, 'label', NEW.label_id);
  ELSE
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('label.unassigned', 'success', OLD.organization_id, OLD.target_group_id, 'label', OLD.label_id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION audit.capture_label_assignments() FROM PUBLIC;
CREATE TRIGGER audit_capture_label_assignments AFTER INSERT OR DELETE ON app.label_assignments
  FOR EACH ROW EXECUTE FUNCTION audit.capture_label_assignments();

-- The bound (owner-owned invoker, VOLATILE; runtime/owner), the runtime grants and the policies.
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  MGR text := '(SELECT app.rls_org_role()) IN (''org_owner'', ''org_admin'')';
  GMGR text := 'app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'')';
  manage text;
  amanage text;
BEGIN
  -- At most 200 definitions per owner scope: the scope row lock is one command, the count a separate following command
  -- (fresh READ COMMITTED snapshot after any wait); the facade already holds the same lock at the same strength.
  EXECUTE format($f$
CREATE FUNCTION app.label_definitions_bound() RETURNS trigger
  LANGUAGE plpgsql VOLATILE SET search_path = pg_catalog, pg_temp AS $b$
DECLARE
  n bigint;
BEGIN
  IF current_user::text NOT IN (%L, %L) THEN
    RETURN NEW;
  END IF;
  IF NEW.owner_scope_type = 'group' THEN
    PERFORM 1 FROM app.groups g WHERE g.id = NEW.group_id AND g.organization_id = NEW.organization_id FOR NO KEY UPDATE;
  ELSE
    PERFORM 1 FROM app.organizations o WHERE o.id = NEW.organization_id FOR NO KEY UPDATE;
  END IF;
  SELECT count(*) INTO n FROM app.label_definitions d
  WHERE d.organization_id = NEW.organization_id AND d.owner_scope_type = NEW.owner_scope_type
    AND d.owner_scope_id = NEW.owner_scope_id;
  IF n >= 200 THEN
    RAISE EXCEPTION 'label definition limit reached'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'label_definitions_bound';
  END IF;
  RETURN NEW;
END $b$
$f$, rt, ow);
  REVOKE ALL ON FUNCTION app.label_definitions_bound() FROM PUBLIC;
  CREATE TRIGGER label_definitions_bound BEFORE INSERT ON app.label_definitions
    FOR EACH ROW EXECUTE FUNCTION app.label_definitions_bound();

  -- Runtime grants (column-limited; Plan r2 §3 step 3).
  EXECUTE format('GRANT SELECT, DELETE ON app.label_definitions TO %I', rt);
  EXECUTE format('GRANT INSERT (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility), '
              || 'UPDATE (name, visibility) ON app.label_definitions TO %I', rt);
  EXECUTE format('GRANT SELECT, DELETE ON app.label_assignments TO %I', rt);
  EXECUTE format('GRANT INSERT (id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id, '
              || 'target_group_id) ON app.label_assignments TO %I', rt);

  -- Definitions: Organization-owned only for the Organization owner/admin; Group-owned only for actual roles in the exact
  -- owning Group (managers read and write; members read target_members definitions). Organization authority and the
  -- Group-management override confer nothing on a Group-owned definition.
  manage := format('organization_id = %s AND %s IS NOT NULL AND ((owner_scope_type = ''organization'' AND %s) OR '
    || '(owner_scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND %s))', O, R, MGR, G, G, GMGR);
  EXECUTE format('CREATE POLICY label_definitions_select ON app.label_definitions FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND %s IS NOT NULL AND ((owner_scope_type = ''organization'' AND %s) OR '
    || '(owner_scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR (visibility = ''target_members'' '
    || 'AND app.rls_group_role(group_id) IS NOT NULL)))))', rt, ow, O, R, MGR, G, G, GMGR);
  EXECUTE format('CREATE POLICY label_definitions_insert ON app.label_definitions FOR INSERT TO %I, %I WITH CHECK (%s)',
    rt, ow, manage);
  EXECUTE format('CREATE POLICY label_definitions_update ON app.label_definitions FOR UPDATE TO %I, %I USING (%s) WITH CHECK (%s)',
    rt, ow, manage, manage);
  EXECUTE format('CREATE POLICY label_definitions_delete ON app.label_definitions FOR DELETE TO %I, %I USING (%s)',
    rt, ow, manage);

  -- Assignments: readable only with the definition (the definitions policy applies inside the EXISTS under the invoker's
  -- row-level security; acyclic) and only for the production target kind; written by the Organization owner/admin for an
  -- Organization-owned definition on a Group, or by an owner/admin of the exact target Group for a Group-owned definition
  -- (type-independent; no production row of that shape exists while label_assignments_target_shape is closed).
  EXECUTE format('CREATE POLICY label_assignments_select ON app.label_assignments FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR target_group_id = %s) AND EXISTS (SELECT 1 FROM '
    || 'app.label_definitions d WHERE d.id = label_assignments.label_id AND d.organization_id = label_assignments.organization_id) '
    || 'AND target_type = ''group'')', rt, ow, O, R, G, G);
  amanage := format('organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR target_group_id = %s) AND ('
    || '(owner_scope_type = ''organization'' AND %s AND target_type = ''group'') OR (owner_scope_type = ''group'' AND '
    || 'app.rls_group_role(target_group_id) IN (''group_owner'', ''group_admin'')))', O, R, G, G, MGR);
  EXECUTE format('CREATE POLICY label_assignments_insert ON app.label_assignments FOR INSERT TO %I, %I WITH CHECK (%s)',
    rt, ow, amanage);
  EXECUTE format('CREATE POLICY label_assignments_delete ON app.label_assignments FOR DELETE TO %I, %I USING (%s)',
    rt, ow, amanage);
END
$do$;
-- +goose StatementEnd

-- The widened catalog: actions + the five label.* captures (47 -> 52), reasons + renamed, visibility_changed (79 -> 81),
-- access.denied operations + the seven label operations (73 -> 80), resource types + label (11 -> 12). One validating
-- scan of audit.events.
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
                    'mcp.connection_established', 'mcp.context_changed', 'mcp.connection_renamed',
                    'organization.slug_changed', 'domain.claimed', 'domain.state_changed', 'domain.checked',
                    'organization.branding_changed',
                    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
                    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
                    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked',
                    'label.created', 'label.updated', 'label.deleted', 'label.assigned', 'label.unassigned',
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
                    'group_management_override_enabled', 'group_management_override_disabled',
                    'signup_authenticated', 'tenant_created', 'signup_origin_rejected', 'signup_user_mismatch',
                    'signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid', 'signup_context_mismatch',
                    'renamed', 'visibility_changed')),
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
                    'enrollment.requests.decide', 'enrollment.requests.revoke', 'organization.set_group_override',
                    'signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                    'signup.complete', 'signup.abandon',
                    'label.definition.list', 'label.definition.create', 'label.definition.update', 'label.definition.delete', 'label.assignment.list', 'label.assign', 'label.unassign')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain', 'invitation', 'enrollment_request', 'label'));

-- The attribution and shape rules (the 00019 body with the five label.* actions added to the trigger-owned list and their
-- shape rule, and label added to the access.denied resource types; CREATE OR REPLACE keeps identity, owner, ACL and
-- trigger).
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
  ovr boolean;
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
    'object.deleted', 'mcp.connection_established', 'mcp.context_changed', 'mcp.connection_renamed', 'organization.slug_changed',
    'domain.claimed',
    'domain.state_changed', 'domain.checked', 'organization.branding_changed',
    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked',
    'label.created', 'label.updated', 'label.deleted', 'label.assigned', 'label.unassigned');
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

  -- The access.override_used authority (G060: the appender's current Organization role and the Organization's Group-
  -- management override) is read only for that action, in its own statement. Inside the CASE below, a generic plan
  -- (always used for a PL/pgSQL simple expression) checks the table and function privileges of every branch, taken or
  -- not, and the provision writer's captures hold neither app.organizations.group_management_override nor
  -- app.rls_org_role().
  ovr := false;
  IF NEW.action = 'access.override_used' THEN
    ovr := app.rls_org_role() IN ('org_owner', 'org_admin')
      AND coalesce((SELECT og.group_management_override FROM app.organizations og WHERE og.id = NEW.organization_id), false);
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
    WHEN NEW.action IN ('mcp.connection_established', 'mcp.context_changed', 'mcp.connection_renamed') THEN
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
    WHEN NEW.action IN ('label.created', 'label.updated', 'label.deleted', 'label.assigned', 'label.unassigned') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'label' AND NEW.resource_id IS NOT NULL
      AND CASE WHEN NEW.action = 'label.updated' THEN NEW.reason IN ('renamed', 'visibility_changed')
               ELSE NEW.reason IS NULL END
      AND (NEW.action NOT IN ('label.assigned', 'label.unassigned') OR NEW.group_id IS NOT NULL)
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated', 'signup_authenticated', 'tenant_created')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
            'organization_not_member', 'host_changed', 'recipient_proof_unavailable', 'signup_origin_rejected',
            'signup_user_mismatch'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
        'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch', 'recipient_ineligible',
        'session_invalid', 'host_changed', 'signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid',
        'signup_context_mismatch')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object', 'organization_domain', 'label')))
      AND (NEW.reason IS DISTINCT FROM 'resource_mismatch'
           OR (NEW.operation = 'mcp.request' AND NEW.organization_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.reason NOT IN ('enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                              'recipient_ineligible', 'session_invalid', 'host_changed')
           OR (NEW.operation IN ('enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request',
                                 'enrollment.admit', 'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update',
                                 'enrollment.requests.list', 'enrollment.requests.decide', 'enrollment.requests.revoke')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL)
           OR (NEW.reason = 'host_changed' AND NEW.operation IN ('signup.challenge', 'signup.complete')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.reason NOT IN ('signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid', 'signup_context_mismatch')
           OR (NEW.operation IN ('signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                                 'signup.complete', 'signup.abandon')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.operation NOT IN ('signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                                 'signup.complete', 'signup.abandon')
           OR (NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL
               AND NEW.reason IN ('signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid',
                                  'signup_context_mismatch', 'host_changed')))
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
      AND ovr
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

-- Postconditions (TestLabelsCatalog repeats and extends them).
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow regrole := current_user::regrole;
  rt regrole := (db || '_runtime')::regrole;
  pr regrole := (db || '_policy_reader')::regrole;
  ar regrole := (db || '_auth_reader')::regrole;
  cw regrole := (db || '_claim_writer')::regrole;
  pw regrole := (db || '_provision_writer')::regrole;
  actual text;
  n bigint;
  r record;
BEGIN
  SELECT string_agg(format('%s:%s:%s:%s', attname, format_type(atttypid, atttypmod), attnotnull, attgenerated), ',' ORDER BY attnum)
    INTO actual FROM pg_attribute WHERE attrelid = 'app.label_definitions'::regclass AND attnum > 0 AND NOT attisdropped;
  IF actual IS DISTINCT FROM 'id:uuid:t:,organization_id:uuid:t:,owner_scope_type:text:t:,owner_scope_id:uuid:t:,group_id:uuid:f:,'
                          || 'name:text:t:,name_key:text:f:s,visibility:text:t:,created_at:timestamp with time zone:t:' THEN
    RAISE EXCEPTION '00020 postcondition: app.label_definitions columns are %', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s', attname, format_type(atttypid, atttypmod), attnotnull, attgenerated), ',' ORDER BY attnum)
    INTO actual FROM pg_attribute WHERE attrelid = 'app.label_assignments'::regclass AND attnum > 0 AND NOT attisdropped;
  IF actual IS DISTINCT FROM 'id:uuid:t:,organization_id:uuid:t:,label_id:uuid:t:,owner_scope_type:text:t:,owner_scope_id:uuid:t:,'
                          || 'target_type:text:t:,target_id:uuid:t:,target_group_id:uuid:t:,created_at:timestamp with time zone:t:' THEN
    RAISE EXCEPTION '00020 postcondition: app.label_assignments columns are %', actual;
  END IF;
  SELECT string_agg(table_name::text || '.' || column_name::text || ':' || privilege_type::text, ','
                    ORDER BY table_name::text COLLATE "C", column_name::text COLLATE "C", privilege_type::text COLLATE "C") INTO actual
  FROM information_schema.column_privileges
  WHERE grantee = rt::text AND table_schema = 'app' AND table_name IN ('label_definitions', 'label_assignments')
    AND privilege_type <> 'SELECT';
  IF actual IS DISTINCT FROM 'label_assignments.id:INSERT,label_assignments.label_id:INSERT,label_assignments.organization_id:INSERT,'
                          || 'label_assignments.owner_scope_id:INSERT,label_assignments.owner_scope_type:INSERT,'
                          || 'label_assignments.target_group_id:INSERT,label_assignments.target_id:INSERT,'
                          || 'label_assignments.target_type:INSERT,label_definitions.group_id:INSERT,label_definitions.id:INSERT,'
                          || 'label_definitions.name:INSERT,label_definitions.name:UPDATE,label_definitions.organization_id:INSERT,'
                          || 'label_definitions.owner_scope_id:INSERT,label_definitions.owner_scope_type:INSERT,'
                          || 'label_definitions.visibility:INSERT,label_definitions.visibility:UPDATE' THEN
    RAISE EXCEPTION '00020 postcondition: runtime column privileges on the label tables are %', actual;
  END IF;
  IF NOT has_table_privilege(rt, 'app.label_definitions', 'SELECT') OR NOT has_table_privilege(rt, 'app.label_definitions', 'DELETE')
     OR NOT has_table_privilege(rt, 'app.label_assignments', 'SELECT') OR NOT has_table_privilege(rt, 'app.label_assignments', 'DELETE')
     OR has_table_privilege(rt, 'app.label_definitions', 'INSERT, UPDATE, TRUNCATE, REFERENCES, TRIGGER')
     OR has_table_privilege(rt, 'app.label_assignments', 'INSERT, UPDATE, TRUNCATE, REFERENCES, TRIGGER')
     OR EXISTS (SELECT 1 FROM pg_class c, aclexplode(c.relacl) a
                WHERE c.oid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass) AND a.grantee NOT IN (ow, rt))
     OR EXISTS (SELECT 1 FROM pg_attribute at, aclexplode(at.attacl) a
                WHERE at.attrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass) AND a.grantee <> rt)
     OR EXISTS (SELECT 1 FROM pg_class WHERE oid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass)
                  AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION '00020 postcondition: the label table grants or FORCE differ';
  END IF;
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 88 THEN
    RAISE EXCEPTION '00020 postcondition: schema app holds % relations, expected 88', n;
  END IF;
  SELECT string_agg(conname::text || ':' || contype::text || ':' || convalidated::text, ',' ORDER BY conname::text COLLATE "C") INTO actual
  FROM pg_constraint WHERE conrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass)
    AND contype IN ('c', 'f', 'p', 'u');
  IF actual IS DISTINCT FROM 'label_assignments_id_nonzero:c:true,label_assignments_label_fkey:f:true,label_assignments_pkey:p:true,'
                          || 'label_assignments_target_group_fkey:f:true,label_assignments_target_shape:c:true,'
                          || 'label_assignments_unique:u:true,label_definitions_group_fkey:f:true,label_definitions_id_nonzero:c:true,'
                          || 'label_definitions_name_valid:c:true,label_definitions_organization_fkey:f:true,'
                          || 'label_definitions_owner_key:u:true,label_definitions_pkey:p:true,label_definitions_scope_consistent:c:true,'
                          || 'label_definitions_scope_name_key:u:true,label_definitions_visibility_valid:c:true' THEN
    RAISE EXCEPTION '00020 postcondition: the label constraints are %', actual;
  END IF;
  -- The four foreign keys are RESTRICT, and every referencing column of the definition key is NOT NULL (a total check).
  IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass)
               AND contype = 'f' AND (confdeltype <> 'r' OR NOT convalidated))
     -- IN (SELECT unnest(conkey)): a parenthesized SELECT after ANY is the subquery form (smallint = smallint[]).
     OR EXISTS (SELECT 1 FROM pg_attribute a WHERE a.attrelid = 'app.label_assignments'::regclass AND NOT a.attnotnull
                  AND a.attnum IN (SELECT unnest(c.conkey) FROM pg_constraint c WHERE c.conrelid = 'app.label_assignments'::regclass
                                     AND c.conname = 'label_assignments_label_fkey'))
     OR position('(target_type = ''group''::text)' IN pg_get_constraintdef((SELECT oid FROM pg_constraint
          WHERE conrelid = 'app.label_assignments'::regclass AND conname = 'label_assignments_target_shape'))) = 0
     OR position('fixture' IN pg_get_constraintdef((SELECT oid FROM pg_constraint
          WHERE conrelid = 'app.label_assignments'::regclass AND conname = 'label_assignments_target_shape'))) > 0 THEN
    RAISE EXCEPTION '00020 postcondition: the label foreign keys or the closed target shape differ';
  END IF;
  SELECT string_agg(p.polname::text || ':' || p.polcmd::text || ':' || p.polpermissive::text || ':'
                    || (SELECT string_agg(s.k, '+' ORDER BY s.k COLLATE "C") FROM (SELECT CASE WHEN x = rt THEN 'rt' WHEN x = ow THEN 'ow'
                        ELSE x::regrole::text END AS k FROM unnest(p.polroles) x) s), ',' ORDER BY p.polname::text COLLATE "C") INTO actual
  FROM pg_policy p WHERE p.polrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass);
  IF actual IS DISTINCT FROM 'label_assignments_delete:d:true:ow+rt,label_assignments_insert:a:true:ow+rt,'
                          || 'label_assignments_select:r:true:ow+rt,label_definitions_delete:d:true:ow+rt,'
                          || 'label_definitions_insert:a:true:ow+rt,label_definitions_select:r:true:ow+rt,'
                          || 'label_definitions_update:w:true:ow+rt' THEN
    RAISE EXCEPTION '00020 postcondition: the label policies are %', actual;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_policy p WHERE p.polrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass)
               AND (position('group_management_override' IN coalesce(pg_get_expr(p.polqual, p.polrelid), '')
                                                    || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) > 0
                    OR position('pg_trigger_depth' IN coalesce(pg_get_expr(p.polqual, p.polrelid), '')
                                                    || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) > 0)) THEN
    RAISE EXCEPTION '00020 postcondition: a label policy reads the Group-management override or the trigger depth';
  END IF;
  SELECT string_agg(c.relname::text || '.' || t.tgname::text, ',' ORDER BY c.relname::text COLLATE "C", t.tgname::text COLLATE "C")
    INTO actual FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
  WHERE t.tgrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass) AND NOT t.tgisinternal AND t.tgenabled = 'O';
  IF actual IS DISTINCT FROM 'label_assignments.audit_capture_label_assignments,label_assignments.label_assignments_keys_immutable,'
                          || 'label_definitions.audit_capture_label_definitions,label_definitions.label_definitions_bound,'
                          || 'label_definitions.label_definitions_keys_immutable' THEN
    RAISE EXCEPTION '00020 postcondition: the label triggers are %', actual;
  END IF;
  FOR r IN SELECT p.oid, p.prosecdef, p.proowner, p.proconfig FROM pg_proc p
           WHERE p.oid IN ('app.label_definitions_keys_immutable()'::regprocedure, 'app.label_assignments_keys_immutable()'::regprocedure,
                           'app.label_definitions_bound()'::regprocedure, 'audit.capture_label_definitions()'::regprocedure,
                           'audit.capture_label_assignments()'::regprocedure, 'audit.events_attribute()'::regprocedure) LOOP
    IF r.prosecdef OR r.proowner <> ow OR r.proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, pg_temp']
       OR has_function_privilege('public', r.oid, 'EXECUTE')
       OR EXISTS (SELECT 1 FROM pg_proc q, aclexplode(q.proacl) a WHERE q.oid = r.oid AND a.grantee <> q.proowner) THEN
      RAISE EXCEPTION '00020 postcondition: function % is not an owner-only invoker body', r.oid::regprocedure;
    END IF;
  END LOOP;
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM 'cff3688235276166e16d3ba9fcdf24b8' THEN
    RAISE EXCEPTION '00020 postcondition: the attribution body differs';
  END IF;
  n := 0;
  FOR r IN SELECT c.conname::text AS conname, pg_get_constraintdef(c.oid) AS def FROM pg_constraint c
           WHERE c.conrelid = 'audit.events'::regclass
             AND c.conname IN ('events_action_valid', 'events_reason_valid', 'events_operation_valid', 'events_resource_type_valid') LOOP
    n := n + 1;
    IF (length(r.def) - length(replace(r.def, '::text', ''))) / length('::text') IS DISTINCT FROM
         (CASE r.conname WHEN 'events_action_valid' THEN 52 WHEN 'events_reason_valid' THEN 81 WHEN 'events_operation_valid' THEN 80 ELSE 12 END)
       OR position(CASE r.conname WHEN 'events_action_valid' THEN '''label.unassigned''' WHEN 'events_reason_valid' THEN '''visibility_changed'''
                     WHEN 'events_operation_valid' THEN '''label.unassign''' ELSE '''label''' END IN r.def) = 0 THEN
      RAISE EXCEPTION '00020 postcondition: the catalog constraint % differs', r.conname;
    END IF;
  END LOOP;
  IF n <> 4 OR EXISTS (SELECT 1 FROM pg_class WHERE oid = 'audit.events'::regclass AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION '00020 postcondition: the catalog constraints or FORCE on audit.events differ';
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = pr) <> 9 OR (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7
     OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 8 OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 3 THEN
    RAISE EXCEPTION '00020 postcondition: the helper inventories differ';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  rt text := current_database() || '_runtime';
  n_def bigint;
  n_asg bigint;
  n_aud bigint;
BEGIN
  -- Labels and their history are never silently discarded: refuse while any label row or any label audit row exists.
  -- FORCE is lifted only inside this transaction; a refusal rolls the lift back.
  ALTER TABLE app.label_definitions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.label_assignments NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_def FROM app.label_definitions;
  SELECT count(*) INTO n_asg FROM app.label_assignments;
  SELECT count(*) INTO n_aud FROM audit.events
  WHERE action IN ('label.created', 'label.updated', 'label.deleted', 'label.assigned', 'label.unassigned')
     OR operation IN ('label.definition.list', 'label.definition.create', 'label.definition.update', 'label.definition.delete',
                      'label.assignment.list', 'label.assign', 'label.unassign')
     OR reason IN ('renamed', 'visibility_changed') OR resource_type = 'label';
  IF n_def > 0 OR n_asg > 0 OR n_aud > 0 THEN
    RAISE EXCEPTION 'refusing to roll back 00020: % label definitions / % label assignments / % label audit rows (history is never discarded)',
      n_def, n_asg, n_aud;
  END IF;
  ALTER TABLE app.label_definitions FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.label_assignments FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  EXECUTE format('REVOKE ALL ON app.label_assignments, app.label_definitions FROM %I', rt);
  DROP TABLE app.label_assignments;
  DROP TABLE app.label_definitions;
  DROP FUNCTION app.label_definitions_bound();
  DROP FUNCTION app.label_definitions_keys_immutable();
  DROP FUNCTION app.label_assignments_keys_immutable();
  DROP FUNCTION audit.capture_label_definitions();
  DROP FUNCTION audit.capture_label_assignments();
END
$do$;
-- +goose StatementEnd

-- The version-19 attribution body (00019), restored.
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
  ovr boolean;
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
    'object.deleted', 'mcp.connection_established', 'mcp.context_changed', 'mcp.connection_renamed', 'organization.slug_changed',
    'domain.claimed',
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

  -- The access.override_used authority (G060: the appender's current Organization role and the Organization's Group-
  -- management override) is read only for that action, in its own statement. Inside the CASE below, a generic plan
  -- (always used for a PL/pgSQL simple expression) checks the table and function privileges of every branch, taken or
  -- not, and the provision writer's captures hold neither app.organizations.group_management_override nor
  -- app.rls_org_role().
  ovr := false;
  IF NEW.action = 'access.override_used' THEN
    ovr := app.rls_org_role() IN ('org_owner', 'org_admin')
      AND coalesce((SELECT og.group_management_override FROM app.organizations og WHERE og.id = NEW.organization_id), false);
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
    WHEN NEW.action IN ('mcp.connection_established', 'mcp.context_changed', 'mcp.connection_renamed') THEN
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
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated', 'signup_authenticated', 'tenant_created')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
            'organization_not_member', 'host_changed', 'recipient_proof_unavailable', 'signup_origin_rejected',
            'signup_user_mismatch'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
        'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch', 'recipient_ineligible',
        'session_invalid', 'host_changed', 'signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid',
        'signup_context_mismatch')
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
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL)
           OR (NEW.reason = 'host_changed' AND NEW.operation IN ('signup.challenge', 'signup.complete')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.reason NOT IN ('signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid', 'signup_context_mismatch')
           OR (NEW.operation IN ('signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                                 'signup.complete', 'signup.abandon')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.operation NOT IN ('signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                                 'signup.complete', 'signup.abandon')
           OR (NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL
               AND NEW.reason IN ('signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid',
                                  'signup_context_mismatch', 'host_changed')))
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
      AND ovr
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

-- The version-19 catalog (actions 00019, reasons and operations 00018, resource types 00016), restored.
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
                    'mcp.connection_established', 'mcp.context_changed', 'mcp.connection_renamed',
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
                    'group_management_override_enabled', 'group_management_override_disabled',
                    'signup_authenticated', 'tenant_created', 'signup_origin_rejected', 'signup_user_mismatch',
                    'signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid', 'signup_context_mismatch')),
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
                    'enrollment.requests.decide', 'enrollment.requests.revoke', 'organization.set_group_override',
                    'signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                    'signup.complete', 'signup.abandon')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain', 'invitation', 'enrollment_request'));

-- +goose StatementBegin
DO $$
BEGIN
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM 'ce481e2677b791e5834a5f3da1af8956'
     OR to_regclass('app.label_definitions') IS NOT NULL OR to_regclass('app.label_assignments') IS NOT NULL THEN
    RAISE EXCEPTION '00020 down postcondition: the version-19 attribution body or the label tables differ';
  END IF;
END
$$;
-- +goose StatementEnd
