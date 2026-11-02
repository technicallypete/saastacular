-- Group member profiles (G075 Plan r2 §3.1-§3.3, §3.6; Spec G075; EPIC-009; docs/foundation/DATABASE-ISOLATION.md,
-- AUDIT.md, AUTHORIZATION.md).
--
-- Additive, one goose transaction:
--   - Precondition: the owner, the six roles with the 00018/00020 attribute rules, version 23, 94 relations in schema app,
--     no 00024 object, the 00023 relationship predicate app.rls_profile_visible and attribution body (owner, definer flag
--     and md5(prosrc)), the four catalog constraints at 53 actions / 84 reasons / 85 operations / 13 resource types, FORCE
--     on app.organizations, app.groups, app.role_bindings, app.users and audit.events, the policy reader's SELECT of the
--     role_bindings columns it reads, the runtime's SELECT (id) on app.role_bindings, the policy reader's Organization
--     scope reader on app.role_bindings and the helper inventories (policy reader 11, claim writer 7, auth reader 10,
--     provision writer 3).
--   - app.group_member_profiles: one optional descriptive role_label (1..100 code points, the G074 stored-shape grammar;
--     NULL = no label) for exactly one current Group membership binding. The primary key is that binding's
--     app.role_bindings.id with ON DELETE CASCADE, so every removal path (G019 RemoveMember and LeaveScope, the G060
--     amendment01 Organization-departure cascade, retirement) takes the label with the membership and a readmission (a
--     new binding id) starts without one; the referential action runs as the table owner outside row-level security, so
--     the runtime needs no DELETE grant. organization_id, group_id and user_id are denormalized from the binding and
--     checked against it by the guard. Relations 94 -> 97 (the table, its primary key, the one-per-membership key).
--   - The owner-owned SECURITY INVOKER guard group_member_profiles_guard: an insert names a current Group membership
--     binding the writer can see under its own row-level security, with the same Organization, Group and user, at
--     version 1 (group_member_profiles_binding otherwise); an update keeps the binding, Organization, Group and user,
--     advances the version by exactly one and changes the label; updated_at is set here.
--   - Forced row-level security with two writer paths, four permissive command-specific policies: the member itself
--     (any own role in that Group) and a current group_owner or group_admin of that exact Group, both in the context
--     Organization with a current Organization role and, in selected-Group context, the selected Group only. The
--     Organization owner/admin path and the G060 override confer no write. Every other read goes through the policy
--     reader's projection. Runtime: SELECT and column-limited INSERT/UPDATE; no DELETE, TRUNCATE, REFERENCES or
--     TRIGGER; no retention grant.
--   - The policy reader: app.rls_group_member_cards(p_subject), the one subject-targeted card projection. Zero rows unless
--     the context is valid and app.rls_profile_visible admits the subject; then one card (Group, label, edit hint,
--     version) per current Group membership of the subject in the session's Organization that the viewer may see: self,
--     a Group the viewer is also a member of, or any of them for a current Organization owner/admin (the override is
--     not read); in selected-Group context the selected Group only; at most 201 rows ordered by Group id (the 201st is
--     the facade's overflow signal). It reads app.role_bindings and the new table only, never app.groups: the Group
--     name is joined by the caller under its own groups_select policy. Inventory 11 -> 12.
--   - Audit (Plan r2 §3.6): the trigger-owned action group_member_profile.updated (Organization and Group scope,
--     resource group_member_profile = the binding id, subject the member, actor the context user; reason
--     role_label_changed; one row per changed label, never the text; a no-op writes nothing; the cascade writes nothing,
--     membership.removed is its record), and the two guarded operations as access.denied operations. Actions 53 -> 54,
--     reasons 84 -> 85, operations 85 -> 87, resource types 13 -> 14; events_attribute admits the action as trigger-owned
--     under its shape rule.
--   - Postconditions.
--
-- Locks: ACCESS EXCLUSIVE on the new table, SHARE ROW EXCLUSIVE on app.role_bindings, app.organizations, app.groups and
-- app.users (foreign keys), catalog-only body replacement, and one validating scan of audit.events per swapped catalog
-- constraint. Migrate, then deploy: version-23 code never names the table; version-24 code against version 23 fails
-- closed on its first card or label statement (42P01 is a store error).
--
-- Down (development and test only; the runner refuses down/reapply with APP_ENV=production) refuses while the table
-- holds a row or audit.events holds a group member profile action, operation, reason or resource (inspected with FORCE
-- lifted inside the transaction; a refusal rolls the lift back), then restores the exact version-23 attribution body and
-- catalog constraints (00023).

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
     OR to_regrole(db || '_claim_writer') IS NULL OR to_regrole(db || '_provision_writer') IS NULL
     OR to_regrole(db || '_retention') IS NULL THEN
    RAISE EXCEPTION 'roles for database % are missing: run infra/postgres/provision-roles.sql (or a fresh 10-roles.sh) first', db;
  END IF;
  ow := current_user::regrole; rt := (db || '_runtime')::regrole; pr := (db || '_policy_reader')::regrole;
  ar := (db || '_auth_reader')::regrole; cw := (db || '_claim_writer')::regrole; pw := (db || '_provision_writer')::regrole;

  -- The 00018/00020 role checks, unchanged.
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

  -- Version 23: the migration history ends at 23, no 00024 object exists, the predicate and the attribution are 00023's.
  IF (SELECT max(version_id) FROM app.goose_db_version) IS DISTINCT FROM 23 THEN
    RAISE EXCEPTION 'version-23 precondition: the current version is %', (SELECT max(version_id) FROM app.goose_db_version);
  END IF;
  IF to_regclass('app.group_member_profiles') IS NOT NULL
     OR to_regprocedure('app.group_member_profiles_guard()') IS NOT NULL
     OR to_regprocedure('audit.capture_group_member_profiles()') IS NOT NULL
     OR to_regprocedure('app.rls_group_member_cards(uuid)') IS NOT NULL
     OR to_regprocedure('app.rls_profile_visible(uuid)') IS NULL
     OR EXISTS (SELECT 1 FROM pg_policy WHERE polname IN ('group_member_profiles_select', 'group_member_profiles_insert',
                  'group_member_profiles_update', 'group_member_profiles_policy_reader')) THEN
    RAISE EXCEPTION 'version-23 precondition: group member profile objects already exist or the profile predicate is missing';
  END IF;
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 94 THEN
    RAISE EXCEPTION 'version-23 precondition: schema app holds % relations, expected 94', n;
  END IF;
  -- The four catalog constraints name exactly the version-23 values (one '::text' per value) and no group member
  -- profile value.
  n := 0;
  FOR r IN SELECT c.conname::text AS conname, pg_get_constraintdef(c.oid) AS def FROM pg_constraint c
           WHERE c.conrelid = 'audit.events'::regclass
             AND c.conname IN ('events_action_valid', 'events_reason_valid', 'events_operation_valid', 'events_resource_type_valid') LOOP
    n := n + 1;
    -- The CASE is parenthesized: a PL/pgSQL IF condition ends at the first THEN outside parentheses (00017 precedent).
    IF (length(r.def) - length(replace(r.def, '::text', ''))) / length('::text') IS DISTINCT FROM
         (CASE r.conname WHEN 'events_action_valid' THEN 53 WHEN 'events_reason_valid' THEN 84 WHEN 'events_operation_valid' THEN 85 ELSE 13 END)
       OR position('''group_member_profile' IN r.def) > 0 OR position('''role_label_changed''' IN r.def) > 0
       OR position('''profile.groups.list''' IN r.def) > 0 OR position('''profile.group_label.update''' IN r.def) > 0 THEN
      RAISE EXCEPTION 'version-23 precondition: the catalog constraint % differs', r.conname;
    END IF;
  END LOOP;
  IF n <> 4 THEN
    RAISE EXCEPTION 'version-23 precondition: % of the four catalog constraints exist', n;
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = pr) <> 11 OR (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7
     OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 10 OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 3 THEN
    RAISE EXCEPTION 'version-23 precondition: the helper inventories differ';
  END IF;
  -- The body this migration replaces (00023's attribution) and the predicate the projection calls (00023's).
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM '1500a5fca99e38a0fc222285b9e085b7'
     OR (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.rls_profile_visible(uuid)'::regprocedure) IS DISTINCT FROM '209a1104526e462836f464237babae7e'
     OR EXISTS (SELECT 1 FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure AND (prosecdef OR proowner <> ow))
     OR EXISTS (SELECT 1 FROM pg_proc WHERE oid = 'app.rls_profile_visible(uuid)'::regprocedure AND (NOT prosecdef OR proowner <> pr)) THEN
    RAISE EXCEPTION 'version-23 precondition: the attribution body or the profile predicate differs';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_class WHERE oid IN ('app.organizations'::regclass, 'app.groups'::regclass, 'app.role_bindings'::regclass,
                                                  'app.users'::regclass, 'audit.events'::regclass)
               AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION 'version-23 precondition: FORCE on app.organizations, app.groups, app.role_bindings, app.users or audit.events differs';
  END IF;
  IF NOT has_column_privilege(pr, 'app.role_bindings', 'organization_id', 'SELECT')
     OR NOT has_column_privilege(pr, 'app.role_bindings', 'subject_user_id', 'SELECT')
     OR NOT has_column_privilege(pr, 'app.role_bindings', 'scope_type', 'SELECT')
     OR NOT has_column_privilege(pr, 'app.role_bindings', 'scope_id', 'SELECT')
     OR NOT has_column_privilege(pr, 'app.role_bindings', 'source', 'SELECT')
     OR NOT has_column_privilege(rt, 'app.role_bindings', 'id', 'SELECT')
     OR NOT EXISTS (SELECT 1 FROM pg_policy WHERE polrelid = 'app.role_bindings'::regclass AND polname = 'role_bindings_scope_reader'
                      AND polroles = ARRAY[pr::oid]) THEN
    RAISE EXCEPTION 'version-23 precondition: the role_bindings reads of the guard or the projection are not granted';
  END IF;
END
$$;
-- +goose StatementEnd

-- The Group member profile (Plan r2 §3.1, C1, C5, C7): one optional label per current Group membership binding, keyed by
-- that binding (CASCADE: the label goes with the membership; a readmission is a new binding). Core's Unicode validation
-- (1..100 code points, trimmed, no control, line or paragraph separator) is authoritative; the CHECK is the G063/G074
-- stored-shape grammar as defense in depth. A cleared label keeps the row (role_label NULL; the version advances).
CREATE TABLE app.group_member_profiles (
  binding_id      uuid        PRIMARY KEY,
  organization_id uuid        NOT NULL,
  group_id        uuid        NOT NULL,
  user_id         uuid        NOT NULL,
  role_label      text        NULL,
  version         integer     NOT NULL DEFAULT 1 CONSTRAINT group_member_profiles_version_valid CHECK (version >= 1),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT group_member_profiles_binding_fkey FOREIGN KEY (binding_id) REFERENCES app.role_bindings (id) ON DELETE CASCADE,
  CONSTRAINT group_member_profiles_organization_fkey FOREIGN KEY (organization_id) REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT group_member_profiles_group_fkey FOREIGN KEY (organization_id, group_id) REFERENCES app.groups (organization_id, id)
    ON DELETE RESTRICT,
  CONSTRAINT group_member_profiles_user_fkey FOREIGN KEY (user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT group_member_profiles_label_valid CHECK (role_label IS NULL OR (char_length(role_label) BETWEEN 1 AND 100
    AND role_label = btrim(role_label) AND (role_label COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f\u2028\u2029]')),
  CONSTRAINT group_member_profiles_one_per_membership UNIQUE (organization_id, user_id, group_id)
);
ALTER TABLE app.group_member_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.group_member_profiles FORCE ROW LEVEL SECURITY;

-- The guard (owner-owned SECURITY INVOKER, so the binding check runs under the writer's row-level security: the member
-- sees its own binding, a Group owner/admin the Group's bindings). A new row names a current Group membership binding with
-- the row's Organization, Group and user and starts at version 1; an update keeps the binding, Organization, Group and
-- user, advances the version by exactly one and changes the label (an identical value is never written). updated_at is
-- set here.
-- +goose StatementBegin
CREATE FUNCTION app.group_member_profiles_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.version <> 1 THEN
      RAISE EXCEPTION 'group member profile version refused'
        USING ERRCODE = 'check_violation', CONSTRAINT = 'group_member_profiles_guard';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM app.role_bindings rb
                   WHERE rb.id = NEW.binding_id AND rb.organization_id = NEW.organization_id AND rb.scope_type = 'group'
                     AND rb.scope_id = NEW.group_id AND rb.group_id = NEW.group_id AND rb.subject_user_id = NEW.user_id
                     AND rb.source = 'membership') THEN
      RAISE EXCEPTION 'group member profile binding refused'
        USING ERRCODE = 'check_violation', CONSTRAINT = 'group_member_profiles_binding';
    END IF;
  ELSIF (NEW.binding_id, NEW.organization_id, NEW.group_id, NEW.user_id)
        IS DISTINCT FROM (OLD.binding_id, OLD.organization_id, OLD.group_id, OLD.user_id)
        OR NEW.version IS DISTINCT FROM OLD.version + 1 OR NEW.role_label IS NOT DISTINCT FROM OLD.role_label THEN
    RAISE EXCEPTION 'group member profile binding or version refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'group_member_profiles_guard';
  END IF;
  NEW.updated_at := now();
  RETURN NEW;
END $$;
-- +goose StatementEnd
CREATE TRIGGER group_member_profiles_guard BEFORE INSERT OR UPDATE ON app.group_member_profiles FOR EACH ROW
  EXECUTE FUNCTION app.group_member_profiles_guard();

-- The capture (owner-owned SECURITY INVOKER; an audit fault refuses the write): one group_member_profile.updated per
-- changed label (set, replace or clear), the binding as resource and the member as subject, the Organization and Group as
-- scope. Never the label; an insert without a label records nothing; there is no delete capture (the cascade is
-- membership.removed's).
-- +goose StatementBegin
CREATE FUNCTION audit.capture_group_member_profiles() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF (TG_OP = 'INSERT' AND NEW.role_label IS NOT NULL) OR (TG_OP = 'UPDATE' AND NEW.role_label IS DISTINCT FROM OLD.role_label) THEN
    INSERT INTO audit.events (action, outcome, reason, organization_id, group_id, subject_user_id, resource_type, resource_id)
    VALUES ('group_member_profile.updated', 'success', 'role_label_changed', NEW.organization_id, NEW.group_id, NEW.user_id,
            'group_member_profile', NEW.binding_id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd
CREATE TRIGGER audit_capture_group_member_profiles AFTER INSERT OR UPDATE OF role_label ON app.group_member_profiles
  FOR EACH ROW EXECUTE FUNCTION audit.capture_group_member_profiles();

-- Grants, the helper policy, the policy-reader projection and the two writer paths (Plan r2 §3.2).
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  scope text;
  writer text;
BEGIN
  REVOKE ALL ON FUNCTION app.group_member_profiles_guard(), audit.capture_group_member_profiles() FROM PUBLIC;

  -- Runtime grants (column-limited): updated_at comes from the guard, the version from the default and the statements.
  -- No DELETE: a cleared label keeps its row and the membership's removal cascades as the table owner.
  EXECUTE format('GRANT SELECT ON app.group_member_profiles TO %I', rt);
  EXECUTE format('GRANT INSERT (binding_id, organization_id, group_id, user_id, role_label, version, updated_at), '
              || 'UPDATE (role_label, version, updated_at) ON app.group_member_profiles TO %I', rt);

  -- The policy reader reads the table inside its projection body only (SELECT, USING (true)).
  EXECUTE format('GRANT SELECT (binding_id, organization_id, group_id, user_id, role_label, version) ON app.group_member_profiles TO %I', pr);
  EXECUTE format('CREATE POLICY group_member_profiles_policy_reader ON app.group_member_profiles FOR SELECT TO %I USING (true)', pr);

  -- The projection, created as the policy reader (the 00006 SET chain, 00014/00017/00023 pattern). It reads
  -- app.role_bindings (the reader's Organization scope reader) and the new table only, so no policy cycle arises: no
  -- policy of app.role_bindings, app.groups or app.account_profiles reads the new table.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pr);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  EXECUTE $r$
CREATE FUNCTION app.rls_group_member_cards(p_subject uuid)
  RETURNS TABLE (group_id uuid, role_label text, can_edit boolean, version integer)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text := app.rls_org_role();
BEGIN
  -- No valid current context, or a subject the profile predicate refuses: zero rows before any card is read.
  IF u IS NULL OR o IS NULL OR r IS NULL OR p_subject IS NULL OR NOT app.rls_profile_visible(p_subject) THEN
    RETURN;
  END IF;
  -- One card per current Group membership of the subject in the session's Organization (the selected Group only, in
  -- selected-Group context) that the viewer may see: self, a Group the viewer is also a member of, or any of them for the
  -- current Organization owner/admin (the Group-management override is not read). The edit hint is self or the viewer's
  -- own owner/admin role in that Group. At most 201 rows: the 201st is the overflow signal, never shown.
  RETURN QUERY
  SELECT sb.scope_id, gm.role_label,
         (p_subject = u OR coalesce(app.rls_group_role(sb.scope_id) IN ('group_owner', 'group_admin'), false)),
         coalesce(gm.version, 0)
  FROM app.role_bindings sb
  LEFT JOIN app.group_member_profiles gm ON gm.organization_id = sb.organization_id AND gm.user_id = sb.subject_user_id
    AND gm.group_id = sb.scope_id
  WHERE sb.organization_id = o AND sb.scope_type = 'group' AND sb.subject_user_id = p_subject AND sb.source = 'membership'
    AND (g IS NULL OR sb.scope_id = g)
    AND (p_subject = u OR r IN ('org_owner', 'org_admin') OR EXISTS (
          SELECT 1 FROM app.role_bindings mine
          WHERE mine.organization_id = o AND mine.scope_type = 'group' AND mine.scope_id = sb.scope_id
            AND mine.subject_user_id = u AND mine.source = 'membership'))
  ORDER BY sb.scope_id
  LIMIT 201;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.rls_group_member_cards(uuid) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_group_member_cards(uuid) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pr);

  -- The two writer paths (TO runtime and owner; FORCE binds the owner to the same contexts): the context Organization
  -- with a current Organization role, the selected Group only in selected-Group context; the member's own row while it
  -- holds any role in that Group, another member's row for a current group_owner or group_admin of that exact Group.
  -- app.rls_group_role returns NULL outside the selected Group, so a selected-Group context narrows writes here too.
  scope := format('organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR group_id = %s)', O, R, G, G);
  writer := format('((user_id = %s AND app.rls_group_role(group_id) IS NOT NULL) '
    || 'OR (user_id <> %s AND app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'')))', U, U);
  EXECUTE format('CREATE POLICY group_member_profiles_select ON app.group_member_profiles FOR SELECT TO %I, %I USING (%s AND '
    || '(user_id = %s OR app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'')))', rt, ow, scope, U);
  EXECUTE format('CREATE POLICY group_member_profiles_insert ON app.group_member_profiles FOR INSERT TO %I, %I WITH CHECK (%s AND %s)',
    rt, ow, scope, writer);
  EXECUTE format('CREATE POLICY group_member_profiles_update ON app.group_member_profiles FOR UPDATE TO %I, %I USING (%s AND %s) '
    || 'WITH CHECK (%s AND %s)', rt, ow, scope, writer, scope, writer);
END
$do$;
-- +goose StatementEnd

-- The widened catalog: actions + group_member_profile.updated (53 -> 54), reasons + role_label_changed (84 -> 85),
-- access.denied operations + profile.groups.list and profile.group_label.update (85 -> 87), resource types +
-- group_member_profile (13 -> 14). One validating scan of audit.events.
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
                    'profile.updated', 'group_member_profile.updated',
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
                    'renamed', 'visibility_changed',
                    'name_changed', 'avatar_set', 'avatar_removed',
                    'role_label_changed')),
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
                    'label.definition.list', 'label.definition.create', 'label.definition.update', 'label.definition.delete', 'label.assignment.list', 'label.assign', 'label.unassign',
                    'profile.get', 'profile.update', 'profile.avatar.get', 'profile.avatar.set', 'profile.avatar.remove',
                    'profile.groups.list', 'profile.group_label.update')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain', 'invitation', 'enrollment_request', 'label',
                     'account_profile', 'group_member_profile'));

-- The attribution and shape rules (the 00023 body with group_member_profile.updated added to the trigger-owned list and
-- its shape rule: the context Organization and a Group, the binding as resource, the member as subject, the context user
-- as actor, reason role_label_changed; CREATE OR REPLACE keeps identity, owner, ACL and trigger).
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
    'label.created', 'label.updated', 'label.deleted', 'label.assigned', 'label.unassigned',
    'profile.updated', 'group_member_profile.updated');
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
    WHEN NEW.action = 'profile.updated' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL
      AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL AND NEW.resource_type = 'account_profile'
      AND NEW.resource_id = NEW.subject_user_id AND NEW.actor_kind = 'user' AND NEW.actor_user_id = NEW.subject_user_id
      AND NEW.reason IN ('name_changed', 'avatar_set', 'avatar_removed')
    WHEN NEW.action = 'group_member_profile.updated' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NOT NULL
      AND NEW.organization_id = o AND NEW.group_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'group_member_profile' AND NEW.resource_id IS NOT NULL AND NEW.actor_kind = 'user'
      AND NEW.reason = 'role_label_changed'
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

-- Postconditions.
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
    INTO actual FROM pg_attribute WHERE attrelid = 'app.group_member_profiles'::regclass AND attnum > 0 AND NOT attisdropped;
  IF actual IS DISTINCT FROM 'binding_id:uuid:t:,organization_id:uuid:t:,group_id:uuid:t:,user_id:uuid:t:,role_label:text:f:,'
                          || 'version:integer:t:,updated_at:timestamp with time zone:t:' THEN
    RAISE EXCEPTION '00024 postcondition: app.group_member_profiles columns are %', actual;
  END IF;
  -- Every table and column ACL entry other than the owner's own, with symbolic labels: runtime SELECT and its
  -- column-limited writes, the policy reader's column SELECT, nothing else (no DELETE, no retention, no PUBLIC).
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname::text AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr' WHEN ar THEN 'ar'
                                                ELSE 'other' END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.oid = 'app.group_member_profiles'::regclass AND a.grantee <> c.relowner
    GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname::text || '.' || at.attname::text, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                                                      WHEN ar THEN 'ar' ELSE 'other' END,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
    FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.oid = 'app.group_member_profiles'::regclass
    GROUP BY c.relname, at.attname, a.grantee) t;
  IF actual IS DISTINCT FROM 'group_member_profiles:rt:SELECT;group_member_profiles.binding_id:pr:SELECT;'
      || 'group_member_profiles.binding_id:rt:INSERT;group_member_profiles.group_id:pr:SELECT;group_member_profiles.group_id:rt:INSERT;'
      || 'group_member_profiles.organization_id:pr:SELECT;group_member_profiles.organization_id:rt:INSERT;'
      || 'group_member_profiles.role_label:pr:SELECT;group_member_profiles.role_label:rt:INSERT,UPDATE;'
      || 'group_member_profiles.updated_at:rt:INSERT,UPDATE;group_member_profiles.user_id:pr:SELECT;'
      || 'group_member_profiles.user_id:rt:INSERT;group_member_profiles.version:pr:SELECT;'
      || 'group_member_profiles.version:rt:INSERT,UPDATE' THEN
    RAISE EXCEPTION '00024 postcondition: the group member profile grants are %', actual;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_class WHERE oid = 'app.group_member_profiles'::regclass AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION '00024 postcondition: row-level security is not forced on app.group_member_profiles';
  END IF;
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 97 THEN
    RAISE EXCEPTION '00024 postcondition: schema app holds % relations, expected 97', n;
  END IF;
  SELECT string_agg(conname::text || ':' || contype::text || ':' || convalidated::text, ',' ORDER BY conname::text COLLATE "C") INTO actual
  FROM pg_constraint WHERE conrelid = 'app.group_member_profiles'::regclass AND contype IN ('c', 'f', 'p', 'u');
  IF actual IS DISTINCT FROM 'group_member_profiles_binding_fkey:f:true,group_member_profiles_group_fkey:f:true,'
                          || 'group_member_profiles_label_valid:c:true,group_member_profiles_one_per_membership:u:true,'
                          || 'group_member_profiles_organization_fkey:f:true,group_member_profiles_pkey:p:true,'
                          || 'group_member_profiles_user_fkey:f:true,group_member_profiles_version_valid:c:true' THEN
    RAISE EXCEPTION '00024 postcondition: the group member profile constraints are %', actual;
  END IF;
  -- Exactly one cascading foreign key, from the binding to app.role_bindings; the others RESTRICT; and no other table
  -- cascades from app.role_bindings.
  SELECT string_agg(conname::text || ':' || confdeltype::text || ':' || (SELECT relname::text FROM pg_class WHERE oid = confrelid), ','
                    ORDER BY conname::text COLLATE "C") INTO actual
  FROM pg_constraint WHERE conrelid = 'app.group_member_profiles'::regclass AND contype = 'f';
  IF actual IS DISTINCT FROM 'group_member_profiles_binding_fkey:c:role_bindings,group_member_profiles_group_fkey:r:groups,'
                          || 'group_member_profiles_organization_fkey:r:organizations,group_member_profiles_user_fkey:r:users' THEN
    RAISE EXCEPTION '00024 postcondition: the group member profile foreign keys are %', actual;
  END IF;
  SELECT string_agg((SELECT relname::text FROM pg_class WHERE oid = conrelid), ',' ORDER BY conname::text COLLATE "C") INTO actual
  FROM pg_constraint WHERE confrelid = 'app.role_bindings'::regclass AND contype = 'f' AND confdeltype = 'c';
  IF actual IS DISTINCT FROM 'group_member_profiles' THEN
    RAISE EXCEPTION '00024 postcondition: the tables cascading from app.role_bindings are %', actual;
  END IF;
  SELECT string_agg(p.polname::text || ':' || p.polcmd::text || ':' || p.polpermissive::text || ':'
                    || (SELECT string_agg(s.k, '+' ORDER BY s.k COLLATE "C") FROM (SELECT CASE WHEN x = rt THEN 'rt' WHEN x = ow THEN 'ow'
                        WHEN x = pr THEN 'pr' WHEN x = ar THEN 'ar' ELSE x::regrole::text END AS k FROM unnest(p.polroles) x) s),
                    ',' ORDER BY p.polname::text COLLATE "C") INTO actual
  FROM pg_policy p WHERE p.polrelid = 'app.group_member_profiles'::regclass;
  IF actual IS DISTINCT FROM 'group_member_profiles_insert:a:true:ow+rt,group_member_profiles_policy_reader:r:true:pr,'
                          || 'group_member_profiles_select:r:true:ow+rt,group_member_profiles_update:w:true:ow+rt' THEN
    RAISE EXCEPTION '00024 postcondition: the group member profile policies are %', actual;
  END IF;
  -- Every runtime policy names the Group-role branch; none reads the profile predicate; no policy of another table
  -- reads the new table (the policy graph stays acyclic).
  SELECT string_agg(p.polname::text, ',' ORDER BY p.polname::text COLLATE "C") INTO actual FROM pg_policy p
  WHERE p.polrelid = 'app.group_member_profiles'::regclass
    AND position('rls_group_role' IN coalesce(pg_get_expr(p.polqual, p.polrelid), '')
                                     || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) > 0
    AND position('rls_profile_visible' IN coalesce(pg_get_expr(p.polqual, p.polrelid), '')
                                          || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) = 0;
  IF actual IS DISTINCT FROM 'group_member_profiles_insert,group_member_profiles_select,group_member_profiles_update' THEN
    RAISE EXCEPTION '00024 postcondition: the policies with the Group-role branch are %', actual;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_policy p WHERE p.polrelid <> 'app.group_member_profiles'::regclass
               AND position('group_member_profiles' IN coalesce(pg_get_expr(p.polqual, p.polrelid), '')
                                                     || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) > 0) THEN
    RAISE EXCEPTION '00024 postcondition: a policy of another table reads app.group_member_profiles';
  END IF;
  SELECT string_agg(c.relname::text || '.' || t.tgname::text, ',' ORDER BY c.relname::text COLLATE "C", t.tgname::text COLLATE "C")
    INTO actual FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
  WHERE t.tgrelid = 'app.group_member_profiles'::regclass AND NOT t.tgisinternal AND t.tgenabled = 'O';
  IF actual IS DISTINCT FROM 'group_member_profiles.audit_capture_group_member_profiles,group_member_profiles.group_member_profiles_guard' THEN
    RAISE EXCEPTION '00024 postcondition: the group member profile triggers are %', actual;
  END IF;
  -- The owner-owned invoker bodies: the guard, the capture and the attribution.
  FOR r IN SELECT p.oid, p.prosecdef, p.proowner, p.proconfig FROM pg_proc p
           WHERE p.oid IN ('app.group_member_profiles_guard()'::regprocedure, 'audit.capture_group_member_profiles()'::regprocedure,
                           'audit.events_attribute()'::regprocedure) LOOP
    IF r.prosecdef OR r.proowner <> ow OR r.proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, pg_temp']
       OR has_function_privilege('public', r.oid, 'EXECUTE')
       OR EXISTS (SELECT 1 FROM pg_proc q, aclexplode(q.proacl) a WHERE q.oid = r.oid AND a.grantee <> q.proowner) THEN
      RAISE EXCEPTION '00024 postcondition: function % is not an owner-only invoker body', r.oid::regprocedure;
    END IF;
  END LOOP;
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM '95e69e5708a6d2ce36be3ae24c9a8744' THEN
    RAISE EXCEPTION '00024 postcondition: the attribution body differs';
  END IF;
  -- The policy-reader bodies the cards depend on: owner, definer flag, volatility, configuration, EXECUTE holders and the
  -- reviewed text (the 00023 predicate unchanged).
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname, CASE p.proowner WHEN pr THEN 'pr' WHEN ar THEN 'ar' ELSE 'other' END,
           p.prosecdef::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN ow THEN 'ow' ELSE 'other' END
                              || '=' || a.privilege_type AS v
                       FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'), md5(p.prosrc)),
           ';' ORDER BY p.proname COLLATE "C") INTO actual
  FROM pg_proc p WHERE p.oid IN ('app.rls_group_member_cards(uuid)'::regprocedure, 'app.rls_profile_visible(uuid)'::regprocedure);
  IF actual IS DISTINCT FROM 'rls_group_member_cards:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE:42742b48f0a0b6154f0d6282a9d665b3;'
      || 'rls_profile_visible:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE:209a1104526e462836f464237babae7e' THEN
    RAISE EXCEPTION '00024 postcondition: the helper bodies are %', actual;
  END IF;
  IF (SELECT pg_get_function_result('app.rls_group_member_cards(uuid)'::regprocedure)) IS DISTINCT FROM
     'TABLE(group_id uuid, role_label text, can_edit boolean, version integer)' THEN
    RAISE EXCEPTION '00024 postcondition: the card projection result columns differ';
  END IF;
  IF has_schema_privilege(pr, 'app', 'CREATE') OR has_schema_privilege(ar, 'app', 'CREATE') THEN
    RAISE EXCEPTION '00024 postcondition: a helper kept CREATE on schema app';
  END IF;
  SELECT string_agg(proname, ',' ORDER BY proname COLLATE "C") INTO actual FROM pg_proc WHERE proowner = pr;
  IF actual IS DISTINCT FROM 'rls_account_profile,rls_dependent_group_bindings,rls_enrollment_mode,rls_group_invitation_candidates,'
                          || 'rls_group_member_cards,rls_group_role,rls_org_role,rls_own_organizations,rls_owner_count,'
                          || 'rls_profile_eligible,rls_profile_visible,rls_sole_owner_scopes' THEN
    RAISE EXCEPTION '00024 postcondition: the policy reader owns %', actual;
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7 OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 10
     OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 3 THEN
    RAISE EXCEPTION '00024 postcondition: the helper inventories differ';
  END IF;
  n := 0;
  FOR r IN SELECT c.conname::text AS conname, pg_get_constraintdef(c.oid) AS def FROM pg_constraint c
           WHERE c.conrelid = 'audit.events'::regclass
             AND c.conname IN ('events_action_valid', 'events_reason_valid', 'events_operation_valid', 'events_resource_type_valid') LOOP
    n := n + 1;
    IF (length(r.def) - length(replace(r.def, '::text', ''))) / length('::text') IS DISTINCT FROM
         (CASE r.conname WHEN 'events_action_valid' THEN 54 WHEN 'events_reason_valid' THEN 85 WHEN 'events_operation_valid' THEN 87 ELSE 14 END)
       OR position(CASE r.conname WHEN 'events_action_valid' THEN '''group_member_profile.updated''' WHEN 'events_reason_valid' THEN '''role_label_changed'''
                     WHEN 'events_operation_valid' THEN '''profile.group_label.update''' ELSE '''group_member_profile''' END IN r.def) = 0 THEN
      RAISE EXCEPTION '00024 postcondition: the catalog constraint % differs', r.conname;
    END IF;
  END LOOP;
  IF n <> 4 OR EXISTS (SELECT 1 FROM pg_class WHERE oid = 'audit.events'::regclass AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION '00024 postcondition: the catalog constraints or FORCE on audit.events differ';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  pr text := current_database() || '_policy_reader';
  n_rows bigint;
  n_aud bigint;
BEGIN
  -- Labels and their history are never silently discarded: refuse while any row or any group member profile audit row
  -- exists. FORCE is lifted only inside this transaction; a refusal rolls the lift back.
  ALTER TABLE app.group_member_profiles NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_rows FROM app.group_member_profiles;
  SELECT count(*) INTO n_aud FROM audit.events
  WHERE action = 'group_member_profile.updated' OR operation IN ('profile.groups.list', 'profile.group_label.update')
     OR reason = 'role_label_changed' OR resource_type = 'group_member_profile';
  IF n_rows > 0 OR n_aud > 0 THEN
    RAISE EXCEPTION 'refusing to roll back 00024: % group member profiles / % group member profile audit rows (labels and their history are never discarded)',
      n_rows, n_aud;
  END IF;
  ALTER TABLE app.group_member_profiles FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  -- The table (its policies, guard and capture triggers, keys and grants go with it), then the owner-owned functions,
  -- then the projection as its owner once nothing calls it.
  DROP TABLE app.group_member_profiles;
  DROP FUNCTION app.group_member_profiles_guard();
  DROP FUNCTION audit.capture_group_member_profiles();
  EXECUTE format('SET LOCAL ROLE %I', pr);
  DROP FUNCTION app.rls_group_member_cards(uuid);
  EXECUTE format('SET LOCAL ROLE %I', ow);
END
$do$;
-- +goose StatementEnd

-- The version-23 attribution body (00023), restored.
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
    'label.created', 'label.updated', 'label.deleted', 'label.assigned', 'label.unassigned',
    'profile.updated');
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
    WHEN NEW.action = 'profile.updated' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL
      AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL AND NEW.resource_type = 'account_profile'
      AND NEW.resource_id = NEW.subject_user_id AND NEW.actor_kind = 'user' AND NEW.actor_user_id = NEW.subject_user_id
      AND NEW.reason IN ('name_changed', 'avatar_set', 'avatar_removed')
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

-- The version-23 catalog (00023), restored.
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
                    'profile.updated',
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
                    'renamed', 'visibility_changed',
                    'name_changed', 'avatar_set', 'avatar_removed')),
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
                    'label.definition.list', 'label.definition.create', 'label.definition.update', 'label.definition.delete', 'label.assignment.list', 'label.assign', 'label.unassign',
                    'profile.get', 'profile.update', 'profile.avatar.get', 'profile.avatar.set', 'profile.avatar.remove')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain', 'invitation', 'enrollment_request', 'label',
                     'account_profile'));

-- +goose StatementBegin
DO $$
DECLARE
  pr regrole := (current_database() || '_policy_reader')::regrole;
  ar regrole := (current_database() || '_auth_reader')::regrole;
  n bigint;
BEGIN
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 94 OR to_regclass('app.group_member_profiles') IS NOT NULL OR to_regprocedure('app.rls_group_member_cards(uuid)') IS NOT NULL
     OR to_regprocedure('app.group_member_profiles_guard()') IS NOT NULL
     OR to_regprocedure('audit.capture_group_member_profiles()') IS NOT NULL THEN
    RAISE EXCEPTION '00024 down postcondition: % relations (expected 94) or a group member profile object remains', n;
  END IF;
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM '1500a5fca99e38a0fc222285b9e085b7'
     OR (SELECT count(*) FROM pg_proc WHERE proowner = pr) <> 11 OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 10
     OR has_schema_privilege(pr, 'app', 'CREATE') OR has_schema_privilege(ar, 'app', 'CREATE') THEN
    RAISE EXCEPTION '00024 down postcondition: the version-23 attribution body or helper inventories differ';
  END IF;
END
$$;
-- +goose StatementEnd
