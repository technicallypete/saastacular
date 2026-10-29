-- Durable audit events (G011; ADR-006; docs/foundation/AUDIT.md).
--
-- One append-only relation, audit.events, in the owner-owned audit schema that provisioning creates empty
-- (infra/postgres/init/10-roles.sh, infra/postgres/provision-roles.sql): a closed catalog of 29 actions, no foreign
-- keys, FORCE row-level security, and server-owned identity, time, actor and correlation fields.
--
-- Two recording mechanisms, no new role and no SECURITY DEFINER function:
--   - Eighteen capture triggers on eight app tables (17 AFTER ROW, 1 AFTER STATEMENT with transition tables for
--     session revocations) call owner-owned SECURITY INVOKER functions that copy allow-listed fields of the changed row
--     into audit.events in the same transaction. Only the local revocation transitions (a session revoked for logout,
--     logout_all, provider_logout or disabled, and a user active -> disabled) wrap their insert in a bounded exception
--     block: an audit-only fault rolls back just that insert, records the transaction-local gap marker
--     audit.revocation_gap and lets the revocation commit (owner decision; Spec R4/AC5). Every other change, sign-in
--     rotation (superseded) and enable included, is strictly atomic with its event.
--   - core appends the non-row events (authentication decisions, security denials, signed-URL issuance, external
--     storage outcomes and revocation gaps) with one INSERT of the 12 descriptive columns runtime may write.
-- A BEFORE INSERT trigger sets every server-owned field from the transaction context, enforces source separation
-- (trigger-owned actions only from a capture trigger, core-owned actions only from a direct insert, by
-- pg_trigger_depth()), validates a direct append's Organization against the established context and its Group against
-- that Organization under row-level security, and enforces each action's shape. UPDATE, DELETE and TRUNCATE are refused
-- for every role.
--
-- Reads: the current org_owner of the active Organization (RLS events_select; the guarded ReadAuditEvents). Global rows
-- (no Organization) are visible to no runtime or owner context. <db>_retention holds no audit privilege (G013 adds its
-- maintenance contract).
--
-- Catalog-only: CREATE TRIGGER takes brief SHARE ROW EXCLUSIVE locks on the eight tables; no rewrite, scan, backfill
-- or app change. History starts at version 8. Down (development and test only; the runner refuses down/reapply with
-- APP_ENV=production) refuses while any row exists. goose runs this file in one transaction.

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

  -- The 00006/00007 role checks, reused.
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

  -- The provisioned audit schema: present, owned by the owner, empty, and with no ACL entry for anyone else.
  IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'audit' AND nspowner = ow) THEN
    RAISE EXCEPTION 'schema audit is missing or not owned by %: run infra/postgres/provision-roles.sql', ow;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_class WHERE relnamespace = 'audit'::regnamespace)
     OR EXISTS (SELECT 1 FROM pg_proc WHERE pronamespace = 'audit'::regnamespace)
     OR EXISTS (SELECT 1 FROM pg_type WHERE typnamespace = 'audit'::regnamespace)
     OR EXISTS (SELECT 1 FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
                WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) THEN
    RAISE EXCEPTION 'schema audit is not empty and owner-only: run infra/postgres/provision-roles.sql on a clean schema';
  END IF;
  IF to_regclass('app.stored_objects') IS NULL THEN
    RAISE EXCEPTION 'version-7 precondition: app.stored_objects is missing';
  END IF;

  -- The exact version-7 inventories of 00007 postconditions 1-4.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  expected := 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-7 precondition: row-level security tables are %, expected %', actual, expected;
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
    || 'stored_objects_insert:stored_objects:a:true:ow+rt;stored_objects_select:stored_objects:r:true:ow+rt;'
    || 'stored_objects_update:stored_objects:w:true:ow+rt;'
    || 'users_auth_reader:users:r:true:ar;users_insert:users:a:true:ow+rt;users_select:users:r:true:ow+rt;'
    || 'users_update:users:w:true:ow+rt';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-7 precondition: policy inventory is %, expected %', actual, expected;
  END IF;
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                          WHEN ar THEN 'ar' ELSE a.grantee::regrole::text END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname || '.' || at.attname, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                          WHEN ar THEN 'ar' ELSE a.grantee::regrole::text END,
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
    RAISE EXCEPTION 'version-7 precondition: ACLs are %, expected %', actual, expected;
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
    || 'role_bindings_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;'
    || 'stored_objects_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-7 precondition: function allowlist is %, expected %', actual, expected;
  END IF;
END
$$;
-- +goose StatementEnd

-- The relation. No foreign keys: deleting a source row never deletes history, and attribution stays historical.
CREATE TABLE audit.events (
  id              uuid        NOT NULL DEFAULT gen_random_uuid(),
  schema_version  smallint    NOT NULL DEFAULT 1 CONSTRAINT events_schema_version CHECK (schema_version = 1),
  occurred_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
  action          text        NOT NULL CONSTRAINT events_action_valid CHECK (action IN (
                    'organization.created', 'organization.renamed', 'organization.policy_changed',
                    'group.created', 'group.renamed', 'group.slug_changed',
                    'membership.granted', 'membership.role_changed', 'membership.removed',
                    'business_profile.created', 'business_profile.updated', 'business_profile.linked',
                    'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled',
                    'session.issued', 'session.revoked', 'session.provider_revocation_recorded',
                    'object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted',
                    'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                    'object.deletion_pending', 'revocation.unaudited')),
  outcome         text        NOT NULL CONSTRAINT events_outcome_valid CHECK (outcome IN ('success', 'denied', 'failed', 'unknown')),
  reason          text        CONSTRAINT events_reason_valid CHECK (reason IS NULL OR reason IN (
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
  operation       text        CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read')),
  actor_kind      text        NOT NULL CONSTRAINT events_actor_kind_valid CHECK (actor_kind IN ('user', 'system', 'unknown')),
  actor_user_id   uuid,
  organization_id uuid,
  group_id        uuid,
  resource_type   text        CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object')),
  resource_id     uuid,
  subject_user_id uuid,
  role            text        CONSTRAINT events_role_valid CHECK (role IS NULL OR role IN
                    ('org_owner', 'org_admin', 'org_member', 'group_owner', 'group_admin', 'group_member')),
  previous_role   text        CONSTRAINT events_previous_role_valid CHECK (previous_role IS NULL OR previous_role IN
                    ('org_owner', 'org_admin', 'org_member', 'group_owner', 'group_admin', 'group_member')),
  ttl_seconds     integer     CONSTRAINT events_ttl_valid CHECK (ttl_seconds IS NULL OR ttl_seconds BETWEEN 1 AND 300),
  request_id      text        CONSTRAINT events_request_id_valid CHECK (request_id IS NULL OR request_id ~ '^[A-Za-z0-9._-]{8,64}$'),
  trace_id        text        CONSTRAINT events_trace_id_valid CHECK (trace_id IS NULL OR trace_id ~ '^[0-9a-f]{32}$'),
  CONSTRAINT events_pkey PRIMARY KEY (id),
  CONSTRAINT events_actor_consistent CHECK ((actor_kind = 'user') = (actor_user_id IS NOT NULL)),
  CONSTRAINT events_group_needs_org CHECK (group_id IS NULL OR organization_id IS NOT NULL),
  CONSTRAINT events_resource_pair CHECK ((resource_type IS NULL) = (resource_id IS NULL)),
  CONSTRAINT events_ids_nonzero CHECK (
    id <> '00000000-0000-0000-0000-000000000000'
    AND actor_user_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'
    AND organization_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'
    AND group_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'
    AND resource_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'
    AND subject_user_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000')
);
CREATE INDEX events_tenant_time_idx ON audit.events (organization_id, occurred_at, id) WHERE organization_id IS NOT NULL;

-- Attribution and shape (BEFORE INSERT). Depth 1 is a direct INSERT (core); depth 2 or more is an INSERT issued by a
-- capture trigger. The parsers raise on a malformed or illegal context (G008).
-- +goose StatementBegin
CREATE FUNCTION audit.events_attribute() RETURNS trigger
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

-- Append-only: UPDATE, DELETE and TRUNCATE are refused for every role, the owner and superusers included, unless a
-- privileged administrator disables the trigger.
-- +goose StatementBegin
CREATE FUNCTION audit.events_refuse() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RAISE EXCEPTION 'audit events are append-only' USING ERRCODE = 'insufficient_privilege';
END $$;
-- +goose StatementEnd

CREATE TRIGGER events_attribute BEFORE INSERT ON audit.events FOR EACH ROW EXECUTE FUNCTION audit.events_attribute();
CREATE TRIGGER events_refuse BEFORE UPDATE OR DELETE ON audit.events FOR EACH ROW EXECUTE FUNCTION audit.events_refuse();
CREATE TRIGGER events_refuse_truncate BEFORE TRUNCATE ON audit.events FOR EACH STATEMENT EXECUTE FUNCTION audit.events_refuse();

-- Capture functions (owner-owned SECURITY INVOKER; the INSERT runs as the invoking role). Each copies named columns of
-- the changed row only.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_organizations() RETURNS trigger
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

-- +goose StatementBegin
CREATE FUNCTION audit.capture_groups() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('group.created', 'success', NEW.organization_id, NEW.id, 'group', NEW.id);
    RETURN NULL;
  END IF;
  IF NEW.name IS DISTINCT FROM OLD.name THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('group.renamed', 'success', NEW.organization_id, NEW.id, 'group', NEW.id);
  END IF;
  IF NEW.slug IS DISTINCT FROM OLD.slug THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('group.slug_changed', 'success', NEW.organization_id, NEW.id, 'group', NEW.id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION audit.capture_role_bindings() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role)
    VALUES ('membership.granted', 'success', NEW.organization_id, NEW.group_id, 'role_binding', NEW.id, NEW.subject_user_id, NEW.role);
  ELSIF TG_OP = 'UPDATE' THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role,
                              previous_role)
    VALUES ('membership.role_changed', 'success', NEW.organization_id, NEW.group_id, 'role_binding', NEW.id,
            NEW.subject_user_id, NEW.role, OLD.role);
  ELSE
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role)
    VALUES ('membership.removed', 'success', OLD.organization_id, OLD.group_id, 'role_binding', OLD.id, OLD.subject_user_id, OLD.role);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION audit.capture_business_profiles() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id)
  VALUES (CASE WHEN TG_OP = 'INSERT' THEN 'business_profile.created' ELSE 'business_profile.updated' END, 'success',
          NEW.organization_id, 'business_profile', NEW.id);
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION audit.capture_group_business_profiles() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('business_profile.linked', 'success', NEW.organization_id, NEW.group_id, 'business_profile', NEW.business_profile_id);
  ELSE
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
    VALUES ('business_profile.unlinked', 'success', OLD.organization_id, OLD.group_id, 'business_profile', OLD.business_profile_id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- The users branch active -> disabled is a revocation-priority path: an audit-only fault rolls back only this insert
-- and records the gap marker; disabled -> active (enable) stays strict.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_users() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, resource_type, resource_id, subject_user_id)
    VALUES ('user.created', 'success', 'user', NEW.id, NEW.id);
  ELSIF OLD.status = 'active' AND NEW.status = 'disabled' THEN
    BEGIN
      INSERT INTO audit.events (action, outcome, resource_type, resource_id, subject_user_id)
      VALUES ('user.disabled', 'success', 'user', NEW.id, NEW.id);
    EXCEPTION WHEN OTHERS THEN
      PERFORM pg_catalog.set_config('audit.revocation_gap',
        concat_ws(',', nullif(pg_catalog.current_setting('audit.revocation_gap', true), ''), 'user.disabled'), true);
    END;
  ELSIF OLD.status = 'disabled' AND NEW.status = 'active' THEN
    INSERT INTO audit.events (action, outcome, resource_type, resource_id, subject_user_id)
    VALUES ('user.enabled', 'success', 'user', NEW.id, NEW.id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION audit.capture_sessions() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, resource_type, resource_id, subject_user_id)
    VALUES ('session.issued', 'success', 'session', NEW.id, NEW.user_id);
  ELSE
    INSERT INTO audit.events (action, outcome, resource_type, resource_id, subject_user_id)
    VALUES ('session.provider_revocation_recorded', CASE WHEN NEW.provider_revocation = 'confirmed' THEN 'success' ELSE 'failed' END,
            'session', NEW.id, NEW.user_id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- Session revocations (one call per sessions UPDATE statement). The flags read the transition tables only, so a
-- statement without a revocation transition (renewal, provider bookkeeping, logout-reference return) executes no audit
-- statement. A sign-in rotation (superseded) inserts strictly; the local revocation reasons insert in one statement
-- inside one bounded exception block, whose handler records the gap marker.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_session_revocations() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  has_strict boolean;
  has_priority boolean;
BEGIN
  SELECT coalesce(bool_or(n.revoke_reason = 'superseded'), false),
         coalesce(bool_or(n.revoke_reason IN ('logout', 'logout_all', 'provider_logout', 'disabled')), false)
    INTO has_strict, has_priority
  FROM new_rows n JOIN old_rows o ON o.id = n.id
  WHERE o.revoked_at IS NULL AND n.revoked_at IS NOT NULL;
  IF has_strict THEN
    INSERT INTO audit.events (action, outcome, reason, resource_type, resource_id, subject_user_id)
    SELECT 'session.revoked', 'success', n.revoke_reason, 'session', n.id, n.user_id
    FROM new_rows n JOIN old_rows o ON o.id = n.id
    WHERE o.revoked_at IS NULL AND n.revoked_at IS NOT NULL AND n.revoke_reason = 'superseded';
  END IF;
  IF has_priority THEN
    BEGIN
      INSERT INTO audit.events (action, outcome, reason, resource_type, resource_id, subject_user_id)
      SELECT 'session.revoked', 'success', n.revoke_reason, 'session', n.id, n.user_id
      FROM new_rows n JOIN old_rows o ON o.id = n.id
      WHERE o.revoked_at IS NULL AND n.revoked_at IS NOT NULL
        AND n.revoke_reason IN ('logout', 'logout_all', 'provider_logout', 'disabled');
    EXCEPTION WHEN OTHERS THEN
      PERFORM pg_catalog.set_config('audit.revocation_gap',
        concat_ws(',', nullif(pg_catalog.current_setting('audit.revocation_gap', true), ''), 'session.revoked'), true);
    END;
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION audit.capture_stored_objects() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  a text;
BEGIN
  IF TG_OP = 'INSERT' THEN
    a := 'object.upload_started';
  ELSIF OLD.state = 'pending' AND NEW.state = 'ready' THEN
    a := 'object.ready';
  ELSIF OLD.state IN ('pending', 'ready') AND NEW.state = 'deleting' THEN
    a := 'object.delete_started';
  ELSIF OLD.state = 'deleting' AND NEW.state = 'deleted' THEN
    a := 'object.deleted';
  ELSE
    RETURN NULL;
  END IF;
  INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id)
  VALUES (a, 'success', NEW.organization_id, NEW.group_id, 'stored_object', NEW.id);
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- The eighteen capture triggers. WHEN clauses and UPDATE OF lists skip no-op updates and unaudited columns.
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.organizations FOR EACH ROW EXECUTE FUNCTION audit.capture_organizations();
CREATE TRIGGER audit_capture_update AFTER UPDATE OF name, group_admins_edit_business_profiles ON app.organizations FOR EACH ROW
  WHEN (OLD.name IS DISTINCT FROM NEW.name OR OLD.group_admins_edit_business_profiles IS DISTINCT FROM NEW.group_admins_edit_business_profiles)
  EXECUTE FUNCTION audit.capture_organizations();
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.groups FOR EACH ROW EXECUTE FUNCTION audit.capture_groups();
CREATE TRIGGER audit_capture_update AFTER UPDATE OF name, slug ON app.groups FOR EACH ROW
  WHEN (OLD.name IS DISTINCT FROM NEW.name OR OLD.slug IS DISTINCT FROM NEW.slug)
  EXECUTE FUNCTION audit.capture_groups();
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.role_bindings FOR EACH ROW EXECUTE FUNCTION audit.capture_role_bindings();
CREATE TRIGGER audit_capture_update AFTER UPDATE OF role ON app.role_bindings FOR EACH ROW
  WHEN (OLD.role IS DISTINCT FROM NEW.role) EXECUTE FUNCTION audit.capture_role_bindings();
CREATE TRIGGER audit_capture_delete AFTER DELETE ON app.role_bindings FOR EACH ROW EXECUTE FUNCTION audit.capture_role_bindings();
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.business_profiles FOR EACH ROW EXECUTE FUNCTION audit.capture_business_profiles();
CREATE TRIGGER audit_capture_update AFTER UPDATE OF display_name, legal_name, description, industry, website ON app.business_profiles
  FOR EACH ROW
  WHEN (ROW(OLD.display_name, OLD.legal_name, OLD.description, OLD.industry, OLD.website)
        IS DISTINCT FROM ROW(NEW.display_name, NEW.legal_name, NEW.description, NEW.industry, NEW.website))
  EXECUTE FUNCTION audit.capture_business_profiles();
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.group_business_profiles FOR EACH ROW
  EXECUTE FUNCTION audit.capture_group_business_profiles();
CREATE TRIGGER audit_capture_delete AFTER DELETE ON app.group_business_profiles FOR EACH ROW
  EXECUTE FUNCTION audit.capture_group_business_profiles();
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.users FOR EACH ROW EXECUTE FUNCTION audit.capture_users();
CREATE TRIGGER audit_capture_update AFTER UPDATE OF status ON app.users FOR EACH ROW
  WHEN (OLD.status IS DISTINCT FROM NEW.status) EXECUTE FUNCTION audit.capture_users();
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.sessions FOR EACH ROW EXECUTE FUNCTION audit.capture_sessions();
CREATE TRIGGER audit_capture_update AFTER UPDATE OF provider_revocation ON app.sessions FOR EACH ROW
  WHEN (OLD.provider_revocation IS DISTINCT FROM NEW.provider_revocation AND NEW.provider_revocation IS NOT NULL)
  EXECUTE FUNCTION audit.capture_sessions();
CREATE TRIGGER audit_capture_revocations AFTER UPDATE ON app.sessions REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows
  FOR EACH STATEMENT EXECUTE FUNCTION audit.capture_session_revocations();
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.stored_objects FOR EACH ROW EXECUTE FUNCTION audit.capture_stored_objects();
CREATE TRIGGER audit_capture_update AFTER UPDATE OF state ON app.stored_objects FOR EACH ROW
  WHEN (OLD.state IS DISTINCT FROM NEW.state) EXECUTE FUNCTION audit.capture_stored_objects();

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
  REVOKE ALL ON FUNCTION audit.events_attribute(), audit.events_refuse(), audit.capture_organizations(), audit.capture_groups(),
    audit.capture_role_bindings(), audit.capture_business_profiles(), audit.capture_group_business_profiles(),
    audit.capture_users(), audit.capture_sessions(), audit.capture_session_revocations(), audit.capture_stored_objects()
    FROM PUBLIC;

  -- Runtime: schema usage, SELECT (filtered by events_select) and INSERT of the 12 descriptive columns. Nothing else, and
  -- nothing for the retention role (G013).
  EXECUTE format('GRANT USAGE ON SCHEMA audit TO %I', rt);
  EXECUTE format('GRANT SELECT ON audit.events TO %I', rt);
  EXECUTE format('GRANT INSERT (action, outcome, reason, operation, resource_type, resource_id, subject_user_id, organization_id, group_id, role, previous_role, ttl_seconds) ON audit.events TO %I', rt);

  ALTER TABLE audit.events ENABLE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;
  EXECUTE format('CREATE POLICY events_select ON audit.events FOR SELECT TO %I, %I USING ('
    || 'organization_id = (SELECT app.ctx_org()) AND (SELECT app.rls_org_role()) = ''org_owner'')', rt, ow);
  EXECUTE format('CREATE POLICY events_insert ON audit.events FOR INSERT TO %I, %I WITH CHECK ('
    || '(actor_kind <> ''user'' OR actor_user_id = (SELECT app.ctx_user())) AND '
    || '(organization_id IS NULL OR organization_id = (SELECT app.ctx_org())))', rt, ow);

  -- Postcondition 1: the version-7 app inventories, unchanged.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  expected := 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'app row-level security tables are %, expected %', actual, expected;
  END IF;
  SELECT count(*)::text INTO actual FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid WHERE c.relnamespace = 'app'::regnamespace;
  IF actual <> '29' THEN
    RAISE EXCEPTION 'app policy count is %, expected 29', actual;
  END IF;
  SELECT count(*)::text INTO actual FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace;
  IF actual <> '15' THEN
    RAISE EXCEPTION 'app function count is %, expected 15', actual;
  END IF;
  SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") INTO actual FROM (
    SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' WHEN pr::regrole THEN 'pr'
                          WHEN ar::regrole THEN 'ar' WHEN (db || '_retention')::regrole THEN 'rn'
                          ELSE a.grantee::regrole::text END || '=' || a.privilege_type AS v
    FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
    WHERE n.nspname = 'app' AND a.grantee <> n.nspowner) l;
  IF actual IS DISTINCT FROM 'ar=USAGE,pr=USAGE,rn=USAGE,rt=USAGE' THEN
    RAISE EXCEPTION 'schema app privileges are %, expected ar=USAGE,pr=USAGE,rn=USAGE,rt=USAGE', actual;
  END IF;

  -- Postcondition 2: the app trigger inventory: the 18 capture triggers plus the existing guards.
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
    RAISE EXCEPTION 'app trigger inventory is %, expected %', actual, expected;
  END IF;

  -- Postcondition 3: the audit relations, RLS flags, policies, triggers, ACLs and functions.
  SELECT string_agg(c.relname || ':' || c.relkind::text || ':' || CASE c.relowner WHEN ow::regrole THEN 'ow' ELSE 'other' END
           || CASE WHEN c.relrowsecurity THEN '+rls' ELSE '' END || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END,
           ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'audit'::regnamespace;
  IF actual IS DISTINCT FROM 'events:r:ow+rls+force,events_pkey:i:ow,events_tenant_time_idx:i:ow' THEN
    RAISE EXCEPTION 'audit relations are %', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s', p.polname, p.polcmd, p.polpermissive::text,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt::regrole THEN 'rt' WHEN ow::regrole THEN 'ow' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p WHERE p.polrelid = 'audit.events'::regclass;
  IF actual IS DISTINCT FROM 'events_insert:a:true:ow+rt;events_select:r:true:ow+rt' THEN
    RAISE EXCEPTION 'audit policies are %', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s', t.tgname, t.tgtype, t.tgenabled, p.proname), ';' ORDER BY t.tgname) INTO actual
  FROM pg_trigger t JOIN pg_proc p ON p.oid = t.tgfoid WHERE t.tgrelid = 'audit.events'::regclass AND NOT t.tgisinternal;
  IF actual IS DISTINCT FROM 'events_attribute:7:O:events_attribute;events_refuse:27:O:events_refuse;'
                             || 'events_refuse_truncate:34:O:events_refuse' THEN
    RAISE EXCEPTION 'audit.events triggers are %', actual;
  END IF;
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' ELSE a.grantee::regrole::text END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'audit'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname || '.' || at.attname, CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' ELSE a.grantee::regrole::text END,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
    FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.relnamespace = 'audit'::regnamespace GROUP BY c.relname, at.attname, a.grantee) t;
  expected := 'events:rt:SELECT;events.action:rt:INSERT;events.group_id:rt:INSERT;events.operation:rt:INSERT;'
    || 'events.organization_id:rt:INSERT;events.outcome:rt:INSERT;events.previous_role:rt:INSERT;events.reason:rt:INSERT;'
    || 'events.resource_id:rt:INSERT;events.resource_type:rt:INSERT;events.role:rt:INSERT;events.subject_user_id:rt:INSERT;'
    || 'events.ttl_seconds:rt:INSERT';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'audit ACLs are %, expected %', actual, expected;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ow::regrole THEN 'ow' ELSE 'other' END,
           p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' ELSE 'other' END
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

  -- Postcondition 4: schema audit privileges: runtime USAGE only.
  SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") INTO actual FROM (
    SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt::regrole THEN 'rt' ELSE a.grantee::regrole::text END
           || '=' || a.privilege_type AS v
    FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
    WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) l;
  IF actual IS DISTINCT FROM 'rt=USAGE' THEN
    RAISE EXCEPTION 'schema audit privileges are %, expected rt=USAGE', actual;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  rt text := db || '_runtime';
  actual text;
BEGIN
  -- Without FORCE the owner's count is not filtered by row-level security. A populated history is never silently
  -- discarded: refuse, and the refusal rolls back the NO FORCE with everything else.
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  IF EXISTS (SELECT 1 FROM audit.events) THEN
    RAISE EXCEPTION 'refusing to roll back 00008: audit.events holds rows (history is never discarded)';
  END IF;
  DROP TRIGGER audit_capture_insert ON app.organizations;
  DROP TRIGGER audit_capture_update ON app.organizations;
  DROP TRIGGER audit_capture_insert ON app.groups;
  DROP TRIGGER audit_capture_update ON app.groups;
  DROP TRIGGER audit_capture_insert ON app.role_bindings;
  DROP TRIGGER audit_capture_update ON app.role_bindings;
  DROP TRIGGER audit_capture_delete ON app.role_bindings;
  DROP TRIGGER audit_capture_insert ON app.business_profiles;
  DROP TRIGGER audit_capture_update ON app.business_profiles;
  DROP TRIGGER audit_capture_insert ON app.group_business_profiles;
  DROP TRIGGER audit_capture_delete ON app.group_business_profiles;
  DROP TRIGGER audit_capture_insert ON app.users;
  DROP TRIGGER audit_capture_update ON app.users;
  DROP TRIGGER audit_capture_insert ON app.sessions;
  DROP TRIGGER audit_capture_update ON app.sessions;
  DROP TRIGGER audit_capture_revocations ON app.sessions;
  DROP TRIGGER audit_capture_insert ON app.stored_objects;
  DROP TRIGGER audit_capture_update ON app.stored_objects;
  DROP TABLE audit.events;
  DROP FUNCTION audit.events_attribute();
  DROP FUNCTION audit.events_refuse();
  DROP FUNCTION audit.capture_organizations();
  DROP FUNCTION audit.capture_groups();
  DROP FUNCTION audit.capture_role_bindings();
  DROP FUNCTION audit.capture_business_profiles();
  DROP FUNCTION audit.capture_group_business_profiles();
  DROP FUNCTION audit.capture_users();
  DROP FUNCTION audit.capture_sessions();
  DROP FUNCTION audit.capture_session_revocations();
  DROP FUNCTION audit.capture_stored_objects();
  EXECUTE format('REVOKE ALL ON SCHEMA audit FROM %I', rt);

  -- The exact version-7 state: no capture trigger in app, an empty, owner-only audit schema.
  SELECT count(*)::text INTO actual FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
  WHERE c.relnamespace = 'app'::regnamespace AND NOT t.tgisinternal AND t.tgname LIKE 'audit\_%';
  IF actual <> '0' THEN
    RAISE EXCEPTION 'version-8 capture triggers survived the rollback';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_class WHERE relnamespace = 'audit'::regnamespace)
     OR EXISTS (SELECT 1 FROM pg_proc WHERE pronamespace = 'audit'::regnamespace)
     OR EXISTS (SELECT 1 FROM pg_type WHERE typnamespace = 'audit'::regnamespace)
     OR EXISTS (SELECT 1 FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a
                WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) THEN
    RAISE EXCEPTION 'version-8 audit objects or grants survived the rollback';
  END IF;
END
$$;
-- +goose StatementEnd
