-- Global account profiles (G074 Plan r3.1 §3.1-§3.6, §3.9a; Spec G074; EPIC-009; docs/foundation/DATABASE-ISOLATION.md,
-- AUDIT.md, OBJECT-STORAGE.md, IDENTITY.md).
--
-- Additive, one goose transaction:
--   - Precondition: the owner, the six roles with the 00018/00020 attribute rules, version 22, 89 relations in schema app,
--     no 00023 object, the 00012 E1 reader, the 00017 invitation-candidate body and the 00020 audit.events_attribute body
--     (owner, definer flag and md5(prosrc)), the four catalog constraints at 52 actions / 81 reasons / 80 operations / 12
--     resource types, FORCE on app.users, app.role_bindings and audit.events, the policy reader's SELECT (id,
--     display_name, status) on app.users and the helper inventories (policy reader 9, claim writer 7, auth reader 10,
--     provision writer 3).
--   - app.account_avatars: identity-owned avatar object metadata (no Organization): a core-generated UUID, the owning
--     user, the generated provider key avatars/<user>/<object>, the G009 four-state lifecycle (pending, ready, deleting,
--     deleted; tombstones are kept), the bounded size (1..512 KiB) and the PNG/JPEG type. app.account_profiles: one row
--     per canonical user UUID with the optional explicit presentation name (NULL = the provider name in
--     app.users.display_name), the current avatar reference and its version, and the optimistic row version. Relations
--     89 -> 94 (two tables, two primary keys, one index).
--   - Two owner-owned SECURITY INVOKER guards: account_avatars_guard (identity immutable; only the owning user changes a
--     row; the 00007 transitions with database-set timestamps and ready_at preserved; I-CUR: a row still referenced as
--     a current avatar never leaves ready) and account_profiles_guard (user immutable; version +1 per change;
--     avatar_version +1 exactly when the reference changes; a new reference names a ready avatar of the same user).
--   - Forced row-level security with self-only writes: nine permissive command-specific policies. Profiles are readable
--     by their user and, through the new policy-reader predicate app.rls_profile_visible, by the current owner/admin of
--     the session's Organization when the subject holds a binding there, or by a current member of a Group of that
--     Organization the subject is also a member of. Avatar rows are readable by their user only. Runtime: SELECT and
--     column-limited INSERT/UPDATE on both tables; no DELETE, TRUNCATE, REFERENCES or TRIGGER; no retention grant.
--   - The policy reader: app.rls_profile_visible (the predicate) and app.rls_account_profile (the gated projection:
--     effective name, avatar fields only while the referenced avatar is ready) are new; the 00017 body of
--     app.rls_group_invitation_candidates is replaced in place: its admission checks and its page query are kept verbatim,
--     the page query becoming a subquery under an outer projection that shows the effective name only for subjects
--     app.rls_profile_visible admits, and never an avatar field. Inventory 9 -> 11.
--   - The auth reader: E1 app.auth_resolve_session is replaced in place (same identity, result columns and predicates):
--     display_name is the effective presentation name coalesce(account_profiles.name, users.display_name). It gains
--     SELECT (user_id, name) on app.account_profiles only. Inventory stays 10.
--   - Audit (Plan r3.1 §3.6): the trigger-owned global action profile.updated (resource type account_profile, resource
--     and subject the user, actor the context user, which the shape rule requires to be the subject; reason
--     name_changed, avatar_set or avatar_removed, one event per changed aspect; never a name, digest, object id or key;
--     a no-op writes nothing), and the five guarded profile operations as access.denied operations. Actions 52 -> 53,
--     reasons 81 -> 84, operations 80 -> 85, resource types 12 -> 13; events_attribute admits the action as trigger-owned
--     under its shape rule.
--   - Postconditions.
--
-- Locks: ACCESS EXCLUSIVE on the two new tables, SHARE ROW EXCLUSIVE on app.users (foreign keys), catalog-only body
-- replacements, and one validating scan of audit.events per swapped catalog constraint. Migrate, then deploy: version-22
-- code never names the tables and E1 keeps its result columns; version-23 code against version 22 fails closed on its
-- first profile statement (42P01 is a store error).
--
-- Down (development and test only; the runner refuses down/reapply with APP_ENV=production) refuses while either table
-- holds a row (dropping avatar metadata would orphan provider bytes, the 00007 rule) or audit.events holds a profile
-- action, operation, reason or resource (inspected with FORCE lifted inside the transaction; a refusal rolls the lift
-- back), then restores the exact version-22 shapes: the 00012 E1 body, the 00017 invitation-candidate body, the 00020
-- attribution body and catalog constraints.

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

  -- Version 22: the migration history ends at 22, no 00023 object exists, the replaced bodies are the version-22 ones.
  IF (SELECT max(version_id) FROM app.goose_db_version) IS DISTINCT FROM 22 THEN
    RAISE EXCEPTION 'version-22 precondition: the current version is %', (SELECT max(version_id) FROM app.goose_db_version);
  END IF;
  IF to_regclass('app.account_avatars') IS NOT NULL OR to_regclass('app.account_profiles') IS NOT NULL
     OR to_regprocedure('app.account_avatars_guard()') IS NOT NULL OR to_regprocedure('app.account_profiles_guard()') IS NOT NULL
     OR to_regprocedure('audit.capture_account_profiles()') IS NOT NULL
     OR to_regprocedure('app.rls_profile_visible(uuid)') IS NOT NULL OR to_regprocedure('app.rls_account_profile(uuid)') IS NOT NULL
     OR to_regprocedure('app.auth_resolve_session(bytea, timestamptz, text)') IS NULL
     OR to_regprocedure('app.rls_group_invitation_candidates(uuid, uuid, integer)') IS NULL
     OR EXISTS (SELECT 1 FROM pg_policy WHERE polname IN ('account_avatars_select', 'account_avatars_insert', 'account_avatars_update',
                  'account_avatars_policy_reader', 'account_profiles_select', 'account_profiles_insert', 'account_profiles_update',
                  'account_profiles_policy_reader', 'account_profiles_auth_reader')) THEN
    RAISE EXCEPTION 'version-22 precondition: account profile objects already exist or a replaced reader is missing';
  END IF;
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 89 THEN
    RAISE EXCEPTION 'version-22 precondition: schema app holds % relations, expected 89', n;
  END IF;
  -- The four catalog constraints name exactly the version-22 values (one '::text' per value) and no profile value.
  n := 0;
  FOR r IN SELECT c.conname::text AS conname, pg_get_constraintdef(c.oid) AS def FROM pg_constraint c
           WHERE c.conrelid = 'audit.events'::regclass
             AND c.conname IN ('events_action_valid', 'events_reason_valid', 'events_operation_valid', 'events_resource_type_valid') LOOP
    n := n + 1;
    -- The CASE is parenthesized: a PL/pgSQL IF condition ends at the first THEN outside parentheses (00017 precedent).
    IF (length(r.def) - length(replace(r.def, '::text', ''))) / length('::text') IS DISTINCT FROM
         (CASE r.conname WHEN 'events_action_valid' THEN 52 WHEN 'events_reason_valid' THEN 81 WHEN 'events_operation_valid' THEN 80 ELSE 12 END)
       OR position('''profile.' IN r.def) > 0 OR position('''account_profile''' IN r.def) > 0
       OR position('''name_changed''' IN r.def) > 0 OR position('''avatar_set''' IN r.def) > 0
       OR position('''avatar_removed''' IN r.def) > 0 THEN
      RAISE EXCEPTION 'version-22 precondition: the catalog constraint % differs', r.conname;
    END IF;
  END LOOP;
  IF n <> 4 THEN
    RAISE EXCEPTION 'version-22 precondition: % of the four catalog constraints exist', n;
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = pr) <> 9 OR (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7
     OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 10 OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 3 THEN
    RAISE EXCEPTION 'version-22 precondition: the helper inventories differ';
  END IF;
  -- The three bodies this migration replaces: 00012's E1, 00017's invitation candidates and 00020's attribution.
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.auth_resolve_session(bytea, timestamptz, text)'::regprocedure) IS DISTINCT FROM '2676fc77baa1904d2ec5a4039069c49f'
     OR (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.rls_group_invitation_candidates(uuid, uuid, integer)'::regprocedure) IS DISTINCT FROM '7aa71bd36d41595cfaf734dcd43857c0'
     OR (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM 'cff3688235276166e16d3ba9fcdf24b8'
     OR EXISTS (SELECT 1 FROM pg_proc WHERE oid = 'app.auth_resolve_session(bytea, timestamptz, text)'::regprocedure AND (NOT prosecdef OR proowner <> ar))
     OR EXISTS (SELECT 1 FROM pg_proc WHERE oid = 'app.rls_group_invitation_candidates(uuid, uuid, integer)'::regprocedure AND (NOT prosecdef OR proowner <> pr))
     OR EXISTS (SELECT 1 FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure AND (prosecdef OR proowner <> ow)) THEN
    RAISE EXCEPTION 'version-22 precondition: the session reader, the invitation-candidate reader or the attribution body differs';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_class WHERE oid IN ('app.users'::regclass, 'app.role_bindings'::regclass, 'audit.events'::regclass)
               AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION 'version-22 precondition: FORCE on app.users, app.role_bindings or audit.events differs';
  END IF;
  IF NOT has_column_privilege(pr, 'app.users', 'id', 'SELECT') OR NOT has_column_privilege(pr, 'app.users', 'display_name', 'SELECT')
     OR NOT has_column_privilege(ar, 'app.users', 'display_name', 'SELECT') THEN
    RAISE EXCEPTION 'version-22 precondition: the helpers lack SELECT on app.users';
  END IF;
END
$$;
-- +goose StatementEnd

-- Avatar object metadata (Plan r3.1 §3.1): identity-owned (no Organization), the 00007 lifecycle shape without
-- deleted_by (the actor is always the owning user). The provider key is generated from the user and the UUID; neither
-- the runtime role, the owner nor a caller can write it. upload_started_at is the write-ahead marker of the one
-- provider Put of the row (Plan r3.1 §3.8 correction, G074 review P1): the writer sets it, while the row is still
-- pending, before it issues the Put, and the guard stamps it. A pending row that never got the marker can never have
-- a Put in flight (a sweep that moved it to deleting makes the marker update fail), so its bytes cannot appear later; a
-- marked row's outcome is unproven until its bytes are observed and deleted or the Put is known to be refused.
CREATE TABLE app.account_avatars (
  id           uuid        PRIMARY KEY CONSTRAINT account_avatars_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  user_id      uuid        NOT NULL,
  storage_key  text        GENERATED ALWAYS AS ('avatars/' || user_id::text || '/' || id::text) STORED,
  state        text        NOT NULL DEFAULT 'pending'
                           CONSTRAINT account_avatars_state_valid CHECK (state IN ('pending', 'ready', 'deleting', 'deleted')),
  byte_size    bigint      NOT NULL CONSTRAINT account_avatars_size_valid CHECK (byte_size BETWEEN 1 AND 524288),
  content_type text        NOT NULL CONSTRAINT account_avatars_type_valid CHECK (content_type IN ('image/png', 'image/jpeg')),
  sha256       bytea       CONSTRAINT account_avatars_sha256_valid CHECK (sha256 IS NULL OR octet_length(sha256) = 32),
  created_at   timestamptz NOT NULL DEFAULT now(),
  ready_at     timestamptz,
  deleting_at  timestamptz,
  deleted_at   timestamptz,
  upload_started_at timestamptz,
  CONSTRAINT account_avatars_user_fkey FOREIGN KEY (user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT account_avatars_lifecycle_check CHECK (
       (state = 'pending'  AND sha256 IS NULL AND ready_at IS NULL AND deleting_at IS NULL AND deleted_at IS NULL)
    OR (state = 'ready'    AND sha256 IS NOT NULL AND ready_at IS NOT NULL AND deleting_at IS NULL AND deleted_at IS NULL)
    OR (state = 'deleting' AND deleting_at IS NOT NULL AND deleted_at IS NULL)
    OR (state = 'deleted'  AND deleting_at IS NOT NULL AND deleted_at IS NOT NULL))
);
CREATE INDEX account_avatars_user_state_idx ON app.account_avatars (user_id, state);
ALTER TABLE app.account_avatars ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.account_avatars FORCE ROW LEVEL SECURITY;

-- The global profile (Plan r3.1 §3.1): name NULL means the current provider name (app.users.display_name, resolved P3).
-- Core's Unicode validation (1..200 code points, trimmed, no control, line or paragraph separator) is authoritative; the
-- CHECK is the G063 stored-shape grammar as defense in depth.
CREATE TABLE app.account_profiles (
  user_id          uuid        PRIMARY KEY,
  name             text        NULL,
  avatar_object_id uuid        NULL,
  avatar_version   integer     NOT NULL DEFAULT 0 CONSTRAINT account_profiles_avatar_version_valid CHECK (avatar_version >= 0),
  version          integer     NOT NULL DEFAULT 1 CONSTRAINT account_profiles_version_valid CHECK (version >= 1),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT account_profiles_user_fkey FOREIGN KEY (user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT account_profiles_avatar_fkey FOREIGN KEY (avatar_object_id) REFERENCES app.account_avatars (id) ON DELETE RESTRICT,
  CONSTRAINT account_profiles_name_valid CHECK (name IS NULL OR (char_length(name) BETWEEN 1 AND 200 AND name = btrim(name)
    AND (name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f\u2028\u2029]')),
  CONSTRAINT account_profiles_avatar_shape CHECK (avatar_object_id IS NULL OR avatar_version >= 1)
);
ALTER TABLE app.account_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.account_profiles FORCE ROW LEVEL SECURITY;

-- The avatar lifecycle guard (owner-owned SECURITY INVOKER, so the I-CUR read runs under the writer's row-level
-- security). Identity columns are immutable; only the owning user changes a row; pending -> ready (with a digest),
-- pending or ready -> deleting, deleting -> deleted are the only transitions, with the timestamps set here and ready_at
-- preserved (a withdrawn published avatar stays distinguishable from a compensated upload). I-CUR: a row still
-- referenced as a profile's current avatar never becomes deleting; the reference is cleared or replaced first, in the
-- same transaction. The one same-state update is the pending upload marker (upload_started_at NULL -> now(), once, the
-- digest unchanged); the column is otherwise preserved, except that the move of a pending row to deleting may clear it (a
-- Put the provider definitively refused, or that was never issued, has no object). Every other same-state update, any update of a deleted row and
-- every other transition are refused.
-- +goose StatementBegin
CREATE FUNCTION app.account_avatars_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  actor uuid := app.ctx_user();
  cleared boolean := OLD.upload_started_at IS NOT NULL AND NEW.upload_started_at IS NULL;
BEGIN
  IF (NEW.id, NEW.user_id, NEW.byte_size, NEW.content_type, NEW.created_at)
     IS DISTINCT FROM (OLD.id, OLD.user_id, OLD.byte_size, OLD.content_type, OLD.created_at) THEN
    RAISE EXCEPTION 'account avatar identity is immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'account_avatars_immutable';
  END IF;
  IF actor IS NULL OR actor IS DISTINCT FROM OLD.user_id THEN
    RAISE EXCEPTION 'account avatar update needs its owner' USING ERRCODE = 'insufficient_privilege';
  END IF;
  NEW.ready_at := OLD.ready_at; NEW.deleting_at := OLD.deleting_at; NEW.deleted_at := OLD.deleted_at;
  NEW.upload_started_at := OLD.upload_started_at;
  IF OLD.state = 'pending' AND NEW.state = 'pending' AND OLD.upload_started_at IS NULL
     AND NEW.sha256 IS NOT DISTINCT FROM OLD.sha256 THEN
    NEW.upload_started_at := now();
  ELSIF OLD.state = 'pending' AND NEW.state = 'ready' AND NEW.sha256 IS NOT NULL THEN
    NEW.ready_at := now();
  ELSIF OLD.state IN ('pending', 'ready') AND NEW.state = 'deleting' AND NEW.sha256 IS NOT DISTINCT FROM OLD.sha256 THEN
    IF EXISTS (SELECT 1 FROM app.account_profiles p WHERE p.avatar_object_id = OLD.id) THEN
      RAISE EXCEPTION 'current account avatar refused'
        USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'account_avatars_current';
    END IF;
    NEW.deleting_at := now();
    -- the one clearing of the marker: a pending row whose Put was definitively refused (or never issued)
    IF cleared AND OLD.state = 'pending' THEN
      NEW.upload_started_at := NULL;
    END IF;
  ELSIF OLD.state = 'deleting' AND NEW.state = 'deleted' AND NEW.sha256 IS NOT DISTINCT FROM OLD.sha256 THEN
    NEW.deleted_at := now();
  ELSE
    RAISE EXCEPTION 'illegal account avatar transition'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'account_avatars_transition';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd
CREATE TRIGGER account_avatars_guard BEFORE UPDATE ON app.account_avatars FOR EACH ROW EXECUTE FUNCTION app.account_avatars_guard();

-- The profile guard (owner-owned SECURITY INVOKER, so the reference check runs under the writer's row-level security;
-- the 00013 branding logo-reference pattern): the user never changes; a new row starts at version 1 (avatar_version 1
-- with an avatar, else 0); every update advances version by exactly one and avatar_version by exactly one when, and only
-- when, the reference changes; a new or changed reference must name a ready avatar of the same user. updated_at is set
-- here.
-- +goose StatementBegin
CREATE FUNCTION app.account_profiles_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.version <> 1 OR NEW.avatar_version <> (CASE WHEN NEW.avatar_object_id IS NULL THEN 0 ELSE 1 END) THEN
      RAISE EXCEPTION 'account profile version refused'
        USING ERRCODE = 'check_violation', CONSTRAINT = 'account_profiles_guard';
    END IF;
  ELSIF NEW.user_id IS DISTINCT FROM OLD.user_id OR NEW.version IS DISTINCT FROM OLD.version + 1
        OR NEW.avatar_version IS DISTINCT FROM
           OLD.avatar_version + (CASE WHEN NEW.avatar_object_id IS DISTINCT FROM OLD.avatar_object_id THEN 1 ELSE 0 END) THEN
    RAISE EXCEPTION 'account profile binding or version refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'account_profiles_guard';
  END IF;
  IF NEW.avatar_object_id IS NOT NULL AND (TG_OP = 'INSERT' OR NEW.avatar_object_id IS DISTINCT FROM OLD.avatar_object_id)
     AND NOT EXISTS (SELECT 1 FROM app.account_avatars a
                     WHERE a.id = NEW.avatar_object_id AND a.user_id = NEW.user_id AND a.state = 'ready') THEN
    RAISE EXCEPTION 'account profile avatar refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'account_profiles_guard';
  END IF;
  NEW.updated_at := now();
  RETURN NEW;
END $$;
-- +goose StatementEnd
CREATE TRIGGER account_profiles_guard BEFORE INSERT OR UPDATE ON app.account_profiles FOR EACH ROW
  EXECUTE FUNCTION app.account_profiles_guard();

-- The capture (owner-owned SECURITY INVOKER; an audit fault refuses the write): one global profile.updated per changed
-- aspect, the user as resource and subject. Never the name, the object id, the digest or the key; a write that changes
-- neither aspect (an INSERT of an empty row included) records nothing.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_account_profiles() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF (TG_OP = 'INSERT' AND NEW.name IS NOT NULL) OR (TG_OP = 'UPDATE' AND NEW.name IS DISTINCT FROM OLD.name) THEN
    INSERT INTO audit.events (action, outcome, reason, subject_user_id, resource_type, resource_id)
    VALUES ('profile.updated', 'success', 'name_changed', NEW.user_id, 'account_profile', NEW.user_id);
  END IF;
  IF (TG_OP = 'INSERT' AND NEW.avatar_object_id IS NOT NULL)
     OR (TG_OP = 'UPDATE' AND NEW.avatar_object_id IS DISTINCT FROM OLD.avatar_object_id) THEN
    INSERT INTO audit.events (action, outcome, reason, subject_user_id, resource_type, resource_id)
    VALUES ('profile.updated', 'success', CASE WHEN NEW.avatar_object_id IS NULL THEN 'avatar_removed' ELSE 'avatar_set' END,
            NEW.user_id, 'account_profile', NEW.user_id);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd
CREATE TRIGGER audit_capture_account_profiles AFTER INSERT OR UPDATE OF name, avatar_object_id ON app.account_profiles
  FOR EACH ROW EXECUTE FUNCTION audit.capture_account_profiles();

-- Grants, helper policies, the policy-reader bodies, the self-only policies and E1 (Plan r3.1 §3.2).
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
  ar text := current_database() || '_auth_reader';
  U text := '(SELECT app.ctx_user())';
BEGIN
  REVOKE ALL ON FUNCTION app.account_avatars_guard(), app.account_profiles_guard(), audit.capture_account_profiles() FROM PUBLIC;

  -- Runtime grants (column-limited): the state default, created_at, the key and the lifecycle timestamps come from the
  -- database and the guards. No DELETE: avatar tombstones and profile rows are kept.
  EXECUTE format('GRANT SELECT ON app.account_avatars TO %I', rt);
  EXECUTE format('GRANT INSERT (id, user_id, byte_size, content_type), UPDATE (state, sha256, upload_started_at) ON app.account_avatars TO %I', rt);
  EXECUTE format('GRANT SELECT ON app.account_profiles TO %I', rt);
  EXECUTE format('GRANT INSERT (user_id, name, avatar_object_id, avatar_version, version, updated_at), '
              || 'UPDATE (name, avatar_object_id, avatar_version, version, updated_at) ON app.account_profiles TO %I', rt);

  -- Helper grants and policies (SELECT only, read only inside the helpers' own bodies): the policy reader reads both
  -- tables for the projection and the invitation candidates; the auth reader reads the explicit name for E1, never an
  -- avatar column.
  EXECUTE format('GRANT SELECT (id, user_id, state, content_type, byte_size, sha256) ON app.account_avatars TO %I', pr);
  EXECUTE format('GRANT SELECT (user_id, name, avatar_object_id, avatar_version) ON app.account_profiles TO %I', pr);
  EXECUTE format('GRANT SELECT (user_id, name) ON app.account_profiles TO %I', ar);
  EXECUTE format('CREATE POLICY account_avatars_policy_reader ON app.account_avatars FOR SELECT TO %I USING (true)', pr);
  EXECUTE format('CREATE POLICY account_profiles_policy_reader ON app.account_profiles FOR SELECT TO %I USING (true)', pr);
  EXECUTE format('CREATE POLICY account_profiles_auth_reader ON app.account_profiles FOR SELECT TO %I USING (true)', ar);

  -- The policy-reader bodies, created and replaced as the policy reader (the 00006 SET chain, 00014/00017 pattern).
  -- app.rls_profile_visible reads app.role_bindings only (under the reader's own policies), so the profile policies
  -- below stay acyclic. Context first: no valid current context, false before any row is read; the selected Group does
  -- not narrow the relationship.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pr);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  EXECUTE $r$
CREATE FUNCTION app.rls_profile_visible(p_subject uuid) RETURNS boolean
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  r text := app.rls_org_role();
BEGIN
  IF u IS NULL OR o IS NULL OR r IS NULL OR p_subject IS NULL THEN
    RETURN false;
  END IF;
  IF p_subject = u THEN
    RETURN true;
  END IF;
  -- The current owner/admin of the session's Organization, when the subject holds a binding there.
  IF r IN ('org_owner', 'org_admin') AND EXISTS (
       SELECT 1 FROM app.role_bindings rb
       WHERE rb.organization_id = o AND rb.subject_user_id = p_subject AND rb.source = 'membership') THEN
    RETURN true;
  END IF;
  -- A current shared Group of the session's Organization.
  RETURN EXISTS (
    SELECT 1 FROM app.role_bindings mine
    JOIN app.role_bindings theirs ON theirs.organization_id = mine.organization_id AND theirs.scope_type = 'group'
      AND theirs.scope_id = mine.scope_id AND theirs.source = 'membership'
    WHERE mine.organization_id = o AND mine.subject_user_id = u AND mine.scope_type = 'group' AND mine.source = 'membership'
      AND theirs.subject_user_id = p_subject);
END $$
$r$;
  EXECUTE $r$
CREATE FUNCTION app.rls_account_profile(p_subject uuid)
  RETURNS TABLE (user_id uuid, name text, avatar_version integer, avatar_object_id uuid, avatar_type text, avatar_size bigint,
                 avatar_sha256 bytea)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
BEGIN
  IF p_subject IS NULL OR NOT app.rls_profile_visible(p_subject) THEN
    RETURN;
  END IF;
  -- The effective name; the avatar fields only while the referenced avatar is ready. users.status is not a filter.
  RETURN QUERY
  SELECT u.id, coalesce(ap.name, u.display_name), CASE WHEN aa.id IS NOT NULL THEN ap.avatar_version END, aa.id,
         aa.content_type, aa.byte_size, aa.sha256
  FROM app.users u
  LEFT JOIN app.account_profiles ap ON ap.user_id = u.id
  LEFT JOIN app.account_avatars aa ON aa.id = ap.avatar_object_id AND aa.user_id = u.id AND aa.state = 'ready'
  WHERE u.id = p_subject;
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
  -- G074: the effective name only for a subject the relationship predicate admits (at most one call per emitted row);
  -- the page query below is the 00017 one, verbatim: candidate set, order and cursor are unchanged. No avatar field.
  RETURN QUERY
  SELECT c.id, CASE WHEN app.rls_profile_visible(c.id) THEN coalesce(ap.name, c.display_name) ELSE c.display_name END
  FROM (
  SELECT us.id, us.display_name FROM app.role_bindings rb JOIN app.users us ON us.id = rb.subject_user_id
  WHERE rb.organization_id = o AND rb.scope_type = 'organization' AND rb.scope_id = o AND rb.source = 'membership'
    AND us.status = 'active' AND (p_after IS NULL OR us.id > p_after)
    AND NOT EXISTS (SELECT 1 FROM app.role_bindings gb WHERE gb.subject_user_id = rb.subject_user_id
                      AND gb.organization_id = o AND gb.scope_type = 'group' AND gb.scope_id = p_group)
  ORDER BY us.id
  LIMIT least(p_limit, 101)
  ) c LEFT JOIN app.account_profiles ap ON ap.user_id = c.id
  ORDER BY c.id;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.rls_profile_visible(uuid), app.rls_account_profile(uuid) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_profile_visible(uuid), app.rls_account_profile(uuid) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pr);

  -- The self-only policies (TO runtime and owner; FORCE binds the owner to the same contexts). Avatar rows are readable
  -- by their user only: other viewers reach avatar metadata through app.rls_account_profile alone.
  EXECUTE format('CREATE POLICY account_avatars_select ON app.account_avatars FOR SELECT TO %I, %I USING (user_id = %s)', rt, ow, U);
  EXECUTE format('CREATE POLICY account_avatars_insert ON app.account_avatars FOR INSERT TO %I, %I WITH CHECK ('
    || 'user_id = %s AND state = ''pending'')', rt, ow, U);
  EXECUTE format('CREATE POLICY account_avatars_update ON app.account_avatars FOR UPDATE TO %I, %I USING (user_id = %s) '
    || 'WITH CHECK (user_id = %s)', rt, ow, U, U);
  EXECUTE format('CREATE POLICY account_profiles_select ON app.account_profiles FOR SELECT TO %I, %I USING ('
    || 'user_id = %s OR app.rls_profile_visible(user_id))', rt, ow, U);
  EXECUTE format('CREATE POLICY account_profiles_insert ON app.account_profiles FOR INSERT TO %I, %I WITH CHECK (user_id = %s)',
    rt, ow, U);
  EXECUTE format('CREATE POLICY account_profiles_update ON app.account_profiles FOR UPDATE TO %I, %I USING (user_id = %s) '
    || 'WITH CHECK (user_id = %s)', rt, ow, U, U);

  -- E1, replaced in place by its owner (the 00012/00022 SET LOCAL ROLE pattern): the effective presentation name; result
  -- columns, the canonical UUID and every predicate unchanged.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $e1$
CREATE OR REPLACE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz, p_origin text)
  RETURNS TABLE (session_id uuid, user_id uuid, display_name text, mfa_assurance text, auth_time timestamptz,
                 idle_expires_at timestamptz, absolute_expires_at timestamptz, renewed_at timestamptz, csrf_token bytea,
                 provider_session_id text, origin text, organization_id uuid)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN QUERY
  SELECT s.id, s.user_id, coalesce(ap.name, u.display_name), s.mfa_assurance, s.auth_time, s.idle_expires_at, s.absolute_expires_at,
         s.renewed_at, s.csrf_token, s.provider_session_id, s.origin, s.organization_id
  FROM app.sessions s JOIN app.users u ON u.id = s.user_id LEFT JOIN app.account_profiles ap ON ap.user_id = u.id
  WHERE s.credential_hash = p_hash AND s.revoked_at IS NULL AND s.idle_expires_at > p_at AND s.absolute_expires_at > p_at
    AND u.status = 'active' AND s.kind = 'browser' AND s.origin IS NOT NULL AND s.organization_id IS NOT NULL
    AND s.origin = p_origin;
END $$
$e1$;
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);
END
$do$;
-- +goose StatementEnd

-- The widened catalog: actions + profile.updated (52 -> 53), reasons + name_changed, avatar_set, avatar_removed
-- (81 -> 84), access.denied operations + the five profile operations (80 -> 85), resource types + account_profile
-- (12 -> 13). One validating scan of audit.events.
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

-- The attribution and shape rules (the 00020 body with profile.updated added to the trigger-owned list and its global
-- shape rule: no Organization or Group, the user as resource and subject, the context user as actor; CREATE OR REPLACE
-- keeps identity, owner, ACL and trigger).
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
    INTO actual FROM pg_attribute WHERE attrelid = 'app.account_avatars'::regclass AND attnum > 0 AND NOT attisdropped;
  IF actual IS DISTINCT FROM 'id:uuid:t:,user_id:uuid:t:,storage_key:text:f:s,state:text:t:,byte_size:bigint:t:,content_type:text:t:,'
                          || 'sha256:bytea:f:,created_at:timestamp with time zone:t:,ready_at:timestamp with time zone:f:,'
                          || 'deleting_at:timestamp with time zone:f:,deleted_at:timestamp with time zone:f:,'
                          || 'upload_started_at:timestamp with time zone:f:' THEN
    RAISE EXCEPTION '00023 postcondition: app.account_avatars columns are %', actual;
  END IF;
  SELECT string_agg(format('%s:%s:%s:%s', attname, format_type(atttypid, atttypmod), attnotnull, attgenerated), ',' ORDER BY attnum)
    INTO actual FROM pg_attribute WHERE attrelid = 'app.account_profiles'::regclass AND attnum > 0 AND NOT attisdropped;
  IF actual IS DISTINCT FROM 'user_id:uuid:t:,name:text:f:,avatar_object_id:uuid:f:,avatar_version:integer:t:,version:integer:t:,'
                          || 'updated_at:timestamp with time zone:t:' THEN
    RAISE EXCEPTION '00023 postcondition: app.account_profiles columns are %', actual;
  END IF;
  -- Every table and column ACL entry of the two tables other than the owner's own, with symbolic labels: runtime SELECT
  -- and its column-limited writes, the two helpers' column SELECT, nothing else (no DELETE, no retention, no PUBLIC).
  SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") INTO actual FROM (
    SELECT c.relname::text AS k, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr' WHEN ar THEN 'ar'
                                                ELSE 'other' END AS g,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.oid IN ('app.account_avatars'::regclass, 'app.account_profiles'::regclass) AND a.grantee <> c.relowner
    GROUP BY c.relname, a.grantee
    UNION ALL
    SELECT c.relname::text || '.' || at.attname::text, CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN pr THEN 'pr'
                                                                      WHEN ar THEN 'ar' ELSE 'other' END,
           string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
    FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.oid IN ('app.account_avatars'::regclass, 'app.account_profiles'::regclass)
    GROUP BY c.relname, at.attname, a.grantee) t;
  IF actual IS DISTINCT FROM 'account_avatars:rt:SELECT;account_avatars.byte_size:pr:SELECT;account_avatars.byte_size:rt:INSERT;'
      || 'account_avatars.content_type:pr:SELECT;account_avatars.content_type:rt:INSERT;account_avatars.id:pr:SELECT;'
      || 'account_avatars.id:rt:INSERT;account_avatars.sha256:pr:SELECT;account_avatars.sha256:rt:UPDATE;'
      || 'account_avatars.state:pr:SELECT;account_avatars.state:rt:UPDATE;account_avatars.upload_started_at:rt:UPDATE;'
      || 'account_avatars.user_id:pr:SELECT;'
      || 'account_avatars.user_id:rt:INSERT;account_profiles:rt:SELECT;account_profiles.avatar_object_id:pr:SELECT;'
      || 'account_profiles.avatar_object_id:rt:INSERT,UPDATE;account_profiles.avatar_version:pr:SELECT;'
      || 'account_profiles.avatar_version:rt:INSERT,UPDATE;account_profiles.name:ar:SELECT;account_profiles.name:pr:SELECT;'
      || 'account_profiles.name:rt:INSERT,UPDATE;account_profiles.updated_at:rt:INSERT,UPDATE;account_profiles.user_id:ar:SELECT;'
      || 'account_profiles.user_id:pr:SELECT;account_profiles.user_id:rt:INSERT;account_profiles.version:rt:INSERT,UPDATE' THEN
    RAISE EXCEPTION '00023 postcondition: the account profile grants are %', actual;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_class WHERE oid IN ('app.account_avatars'::regclass, 'app.account_profiles'::regclass)
               AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION '00023 postcondition: row-level security is not forced on the account profile tables';
  END IF;
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 94 THEN
    RAISE EXCEPTION '00023 postcondition: schema app holds % relations, expected 94', n;
  END IF;
  SELECT string_agg(conname::text || ':' || contype::text || ':' || convalidated::text, ',' ORDER BY conname::text COLLATE "C") INTO actual
  FROM pg_constraint WHERE conrelid IN ('app.account_avatars'::regclass, 'app.account_profiles'::regclass)
    AND contype IN ('c', 'f', 'p', 'u');
  IF actual IS DISTINCT FROM 'account_avatars_id_nonzero:c:true,account_avatars_lifecycle_check:c:true,account_avatars_pkey:p:true,'
                          || 'account_avatars_sha256_valid:c:true,account_avatars_size_valid:c:true,account_avatars_state_valid:c:true,'
                          || 'account_avatars_type_valid:c:true,account_avatars_user_fkey:f:true,account_profiles_avatar_fkey:f:true,'
                          || 'account_profiles_avatar_shape:c:true,account_profiles_avatar_version_valid:c:true,'
                          || 'account_profiles_name_valid:c:true,account_profiles_pkey:p:true,account_profiles_user_fkey:f:true,'
                          || 'account_profiles_version_valid:c:true' THEN
    RAISE EXCEPTION '00023 postcondition: the account profile constraints are %', actual;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid IN ('app.account_avatars'::regclass, 'app.account_profiles'::regclass)
               AND contype = 'f' AND (confdeltype <> 'r' OR NOT convalidated)) THEN
    RAISE EXCEPTION '00023 postcondition: an account profile foreign key is not RESTRICT';
  END IF;
  SELECT string_agg(p.polname::text || ':' || p.polcmd::text || ':' || p.polpermissive::text || ':'
                    || (SELECT string_agg(s.k, '+' ORDER BY s.k COLLATE "C") FROM (SELECT CASE WHEN x = rt THEN 'rt' WHEN x = ow THEN 'ow'
                        WHEN x = pr THEN 'pr' WHEN x = ar THEN 'ar' ELSE x::regrole::text END AS k FROM unnest(p.polroles) x) s),
                    ',' ORDER BY p.polname::text COLLATE "C") INTO actual
  FROM pg_policy p WHERE p.polrelid IN ('app.account_avatars'::regclass, 'app.account_profiles'::regclass);
  IF actual IS DISTINCT FROM 'account_avatars_insert:a:true:ow+rt,account_avatars_policy_reader:r:true:pr,'
                          || 'account_avatars_select:r:true:ow+rt,account_avatars_update:w:true:ow+rt,'
                          || 'account_profiles_auth_reader:r:true:ar,account_profiles_insert:a:true:ow+rt,'
                          || 'account_profiles_policy_reader:r:true:pr,account_profiles_select:r:true:ow+rt,'
                          || 'account_profiles_update:w:true:ow+rt' THEN
    RAISE EXCEPTION '00023 postcondition: the account profile policies are %', actual;
  END IF;
  -- Only the profile read policy consults the relationship predicate; no avatar policy reads another user's row.
  SELECT string_agg(p.polname::text, ',' ORDER BY p.polname::text COLLATE "C") INTO actual FROM pg_policy p
  WHERE p.polrelid IN ('app.account_avatars'::regclass, 'app.account_profiles'::regclass)
    AND position('rls_profile_visible' IN coalesce(pg_get_expr(p.polqual, p.polrelid), '')
                                          || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) > 0;
  IF actual IS DISTINCT FROM 'account_profiles_select' THEN
    RAISE EXCEPTION '00023 postcondition: the policies reading the predicate are %', actual;
  END IF;
  SELECT string_agg(c.relname::text || '.' || t.tgname::text, ',' ORDER BY c.relname::text COLLATE "C", t.tgname::text COLLATE "C")
    INTO actual FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
  WHERE t.tgrelid IN ('app.account_avatars'::regclass, 'app.account_profiles'::regclass) AND NOT t.tgisinternal AND t.tgenabled = 'O';
  IF actual IS DISTINCT FROM 'account_avatars.account_avatars_guard,account_profiles.account_profiles_guard,'
                          || 'account_profiles.audit_capture_account_profiles' THEN
    RAISE EXCEPTION '00023 postcondition: the account profile triggers are %', actual;
  END IF;
  -- The owner-owned invoker bodies: guards, capture and the attribution.
  FOR r IN SELECT p.oid, p.prosecdef, p.proowner, p.proconfig FROM pg_proc p
           WHERE p.oid IN ('app.account_avatars_guard()'::regprocedure, 'app.account_profiles_guard()'::regprocedure,
                           'audit.capture_account_profiles()'::regprocedure, 'audit.events_attribute()'::regprocedure) LOOP
    IF r.prosecdef OR r.proowner <> ow OR r.proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, pg_temp']
       OR has_function_privilege('public', r.oid, 'EXECUTE')
       OR EXISTS (SELECT 1 FROM pg_proc q, aclexplode(q.proacl) a WHERE q.oid = r.oid AND a.grantee <> q.proowner) THEN
      RAISE EXCEPTION '00023 postcondition: function % is not an owner-only invoker body', r.oid::regprocedure;
    END IF;
  END LOOP;
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM '1500a5fca99e38a0fc222285b9e085b7' THEN
    RAISE EXCEPTION '00023 postcondition: the attribution body differs';
  END IF;
  -- The helper bodies: owner, definer flag, volatility, configuration, EXECUTE holders and the reviewed text.
  SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname, CASE p.proowner WHEN pr THEN 'pr' WHEN ar THEN 'ar' ELSE 'other' END,
           p.prosecdef::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
           coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                       SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN ow THEN 'ow' ELSE 'other' END
                              || '=' || a.privilege_type AS v
                       FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'), md5(p.prosrc)),
           ';' ORDER BY p.proname COLLATE "C") INTO actual
  FROM pg_proc p WHERE p.oid IN ('app.rls_profile_visible(uuid)'::regprocedure, 'app.rls_account_profile(uuid)'::regprocedure,
                                 'app.rls_group_invitation_candidates(uuid, uuid, integer)'::regprocedure,
                                 'app.auth_resolve_session(bytea, timestamptz, text)'::regprocedure);
  IF actual IS DISTINCT FROM 'auth_resolve_session:ar:true:s:search_path=pg_catalog, pg_temp:rt=EXECUTE:1308290fd397a6613a6fe47bcdd30ab1;'
      || 'rls_account_profile:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE:6e9632805eb10aee1b63e426a41544d9;'
      || 'rls_group_invitation_candidates:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE:71d009452af524e1c92b78c7cb79ca9e;'
      || 'rls_profile_visible:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE:209a1104526e462836f464237babae7e' THEN
    RAISE EXCEPTION '00023 postcondition: the helper bodies are %', actual;
  END IF;
  IF (SELECT pg_get_function_result('app.auth_resolve_session(bytea, timestamptz, text)'::regprocedure)) IS DISTINCT FROM
     'TABLE(session_id uuid, user_id uuid, display_name text, mfa_assurance text, auth_time timestamp with time zone, '
     || 'idle_expires_at timestamp with time zone, absolute_expires_at timestamp with time zone, renewed_at timestamp with time zone, '
     || 'csrf_token bytea, provider_session_id text, origin text, organization_id uuid)' THEN
    RAISE EXCEPTION '00023 postcondition: the session reader result columns differ';
  END IF;
  IF has_schema_privilege(pr, 'app', 'CREATE') OR has_schema_privilege(ar, 'app', 'CREATE') THEN
    RAISE EXCEPTION '00023 postcondition: a helper kept CREATE on schema app';
  END IF;
  SELECT string_agg(proname, ',' ORDER BY proname COLLATE "C") INTO actual FROM pg_proc WHERE proowner = pr;
  IF actual IS DISTINCT FROM 'rls_account_profile,rls_dependent_group_bindings,rls_enrollment_mode,rls_group_invitation_candidates,'
                          || 'rls_group_role,rls_org_role,rls_own_organizations,rls_owner_count,rls_profile_eligible,'
                          || 'rls_profile_visible,rls_sole_owner_scopes' THEN
    RAISE EXCEPTION '00023 postcondition: the policy reader owns %', actual;
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7 OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 10
     OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 3 THEN
    RAISE EXCEPTION '00023 postcondition: the helper inventories differ';
  END IF;
  n := 0;
  FOR r IN SELECT c.conname::text AS conname, pg_get_constraintdef(c.oid) AS def FROM pg_constraint c
           WHERE c.conrelid = 'audit.events'::regclass
             AND c.conname IN ('events_action_valid', 'events_reason_valid', 'events_operation_valid', 'events_resource_type_valid') LOOP
    n := n + 1;
    IF (length(r.def) - length(replace(r.def, '::text', ''))) / length('::text') IS DISTINCT FROM
         (CASE r.conname WHEN 'events_action_valid' THEN 53 WHEN 'events_reason_valid' THEN 84 WHEN 'events_operation_valid' THEN 85 ELSE 13 END)
       OR position(CASE r.conname WHEN 'events_action_valid' THEN '''profile.updated''' WHEN 'events_reason_valid' THEN '''avatar_removed'''
                     WHEN 'events_operation_valid' THEN '''profile.avatar.remove''' ELSE '''account_profile''' END IN r.def) = 0 THEN
      RAISE EXCEPTION '00023 postcondition: the catalog constraint % differs', r.conname;
    END IF;
  END LOOP;
  IF n <> 4 OR EXISTS (SELECT 1 FROM pg_class WHERE oid = 'audit.events'::regclass AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION '00023 postcondition: the catalog constraints or FORCE on audit.events differ';
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
  ar text := current_database() || '_auth_reader';
  n_prof bigint;
  n_av bigint;
  n_aud bigint;
BEGIN
  -- Profiles, avatar metadata and their history are never silently discarded: refuse while any row or any profile audit
  -- row exists. FORCE is lifted only inside this transaction; a refusal rolls the lift back.
  ALTER TABLE app.account_profiles NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.account_avatars NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_prof FROM app.account_profiles;
  SELECT count(*) INTO n_av FROM app.account_avatars;
  SELECT count(*) INTO n_aud FROM audit.events
  WHERE action = 'profile.updated'
     OR operation IN ('profile.get', 'profile.update', 'profile.avatar.get', 'profile.avatar.set', 'profile.avatar.remove')
     OR reason IN ('name_changed', 'avatar_set', 'avatar_removed') OR resource_type = 'account_profile';
  IF n_prof > 0 OR n_av > 0 OR n_aud > 0 THEN
    RAISE EXCEPTION 'refusing to roll back 00023: % profiles / % avatar rows / % profile audit rows (avatar metadata and history are never discarded)',
      n_prof, n_av, n_aud;
  END IF;
  ALTER TABLE app.account_profiles FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.account_avatars FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  -- The version-22 E1 (00012), restored by its owner.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $e1$
CREATE OR REPLACE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz, p_origin text)
  RETURNS TABLE (session_id uuid, user_id uuid, display_name text, mfa_assurance text, auth_time timestamptz,
                 idle_expires_at timestamptz, absolute_expires_at timestamptz, renewed_at timestamptz, csrf_token bytea,
                 provider_session_id text, origin text, organization_id uuid)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN QUERY
  SELECT s.id, s.user_id, u.display_name, s.mfa_assurance, s.auth_time, s.idle_expires_at, s.absolute_expires_at,
         s.renewed_at, s.csrf_token, s.provider_session_id, s.origin, s.organization_id
  FROM app.sessions s JOIN app.users u ON u.id = s.user_id
  WHERE s.credential_hash = p_hash AND s.revoked_at IS NULL AND s.idle_expires_at > p_at AND s.absolute_expires_at > p_at
    AND u.status = 'active' AND s.kind = 'browser' AND s.origin IS NOT NULL AND s.organization_id IS NOT NULL
    AND s.origin = p_origin;
END $$
$e1$;
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);

  -- The version-22 invitation-candidate body (00017), restored by its owner.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pr);
  EXECUTE format('SET LOCAL ROLE %I', pr);
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

  -- The tables (their policies, guard and capture triggers, keys, index and grants go with them), then the owner-owned
  -- functions, then the two readers as their owner once no policy calls them.
  DROP TABLE app.account_profiles;
  DROP TABLE app.account_avatars;
  DROP FUNCTION app.account_profiles_guard();
  DROP FUNCTION app.account_avatars_guard();
  DROP FUNCTION audit.capture_account_profiles();
  EXECUTE format('SET LOCAL ROLE %I', pr);
  DROP FUNCTION app.rls_account_profile(uuid);
  DROP FUNCTION app.rls_profile_visible(uuid);
  EXECUTE format('SET LOCAL ROLE %I', ow);
END
$do$;
-- +goose StatementEnd

-- The version-22 attribution body (00020), restored.
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

-- The version-22 catalog (00020), restored.
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
  IF n <> 89 OR to_regclass('app.account_profiles') IS NOT NULL OR to_regclass('app.account_avatars') IS NOT NULL
     OR to_regprocedure('app.rls_profile_visible(uuid)') IS NOT NULL OR to_regprocedure('app.rls_account_profile(uuid)') IS NOT NULL
     OR to_regprocedure('audit.capture_account_profiles()') IS NOT NULL THEN
    RAISE EXCEPTION '00023 down postcondition: % relations (expected 89) or an account profile object remains', n;
  END IF;
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.auth_resolve_session(bytea, timestamptz, text)'::regprocedure) IS DISTINCT FROM '2676fc77baa1904d2ec5a4039069c49f'
     OR (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.rls_group_invitation_candidates(uuid, uuid, integer)'::regprocedure) IS DISTINCT FROM '7aa71bd36d41595cfaf734dcd43857c0'
     OR (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure) IS DISTINCT FROM 'cff3688235276166e16d3ba9fcdf24b8'
     OR (SELECT count(*) FROM pg_proc WHERE proowner = pr) <> 9 OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 10
     OR has_schema_privilege(pr, 'app', 'CREATE') OR has_schema_privilege(ar, 'app', 'CREATE') THEN
    RAISE EXCEPTION '00023 down postcondition: the version-22 bodies or helper inventories differ';
  END IF;
END
$$;
-- +goose StatementEnd
