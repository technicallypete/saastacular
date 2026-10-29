-- Row-level security and least-privilege grants (G008; docs/foundation/DATABASE-ISOLATION.md).
--
-- ENABLE and FORCE row-level security on the five tenant tables, the global users table and the three authentication
-- tables (sessions, identity_principals, login_transactions), with one permissive, command-specific policy per
-- permitted command (TO the runtime role and the owner, so FORCE binds the owner to the same contexts) and default
-- deny everywhere else. identity_provider_binding stays public namespace configuration without row-level security; the
-- goose metadata stays owner-only.
--
-- Context: four transaction-local settings (app.user_id, app.organization_id, app.group_id, app.login_binding) set by
-- core's one context statement. Four owner-owned SECURITY INVOKER parsers validate the whole context shape (EMPTY, LOGIN,
-- IDENTITY, TENANT-ORG, TENANT-GROUP) on every call: a malformed UUID raises 22P02, a zero UUID, a malformed binding or
-- an illegal shape raises 22023. Two STABLE SECURITY DEFINER authority readers owned by the NOLOGIN, NOBYPASSRLS role
-- <db>_policy_reader (the Organization gate with G007's selected-Group validity, and the caller's Group role) keep the
-- policy graph acyclic. Three STABLE SECURITY DEFINER authentication readers owned by the NOLOGIN, NOBYPASSRLS role
-- <db>_auth_reader resolve a session by credential hash (E1), a verified provider principal to its user (E2) and a
-- logout reference to its session (E3). Both helper roles hold only SELECT on the exact columns their readers need,
-- with policies TO that helper only; their sole member is the owner (SET, no INHERIT); runtime is a member of nothing.
--
-- Grants: runtime loses every privilege row-level security could not filter or no supported path uses: INSERT and
-- DELETE on tenant tables, INSERT(name)/UPDATE(name) on organizations, the role_bindings id and created_at columns, and
-- everything on users and sessions except the exact columns of §3.1 of the G008 Plan.
--
-- No table, column, constraint, index or data changes. The migration takes ACCESS EXCLUSIVE locks on the nine tables
-- for its transaction (catalog-only; no rewrite or scan): run it in a quiescent window, provisioning first
-- (infra/postgres/provision-roles.sql), then this migration and the G008 core code as one step.
--
-- Down restores the exact version-5 catalog and grants (development and test only; the runner refuses down/reapply
-- with APP_ENV=production). The helper roles survive as cluster objects holding nothing in schema app. goose runs this
-- file in one transaction.

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

  -- Role attributes: runtime and both helpers can never bypass row-level security or administer anything; the helpers
  -- cannot log in or inherit.
  FOR r IN SELECT rolname, rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolinherit, rolcanlogin, rolreplication
           FROM pg_roles WHERE oid IN (rt, pr, ar) LOOP
    IF r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication THEN
      RAISE EXCEPTION 'role % has an unsafe attribute: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
    IF r.rolname <> rt::text AND (r.rolcanlogin OR r.rolinherit) THEN
      RAISE EXCEPTION 'helper role % must be NOLOGIN NOINHERIT: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
  END LOOP;
  -- Memberships: each helper's only usable member is the owner (SET, no INHERIT, no ADMIN); any other grant on a helper
  -- is an ADMIN-only grant (no SET, no INHERIT) of a role that is not an application role (the provisioning creator).
  -- Runtime is a member of nothing.
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
  -- The helpers hold nothing yet.
  IF EXISTS (SELECT 1 FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
             WHERE c.relnamespace = 'app'::regnamespace AND a.grantee IN (pr, ar))
     OR EXISTS (SELECT 1 FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0
                CROSS JOIN LATERAL aclexplode(at.attacl) a
                WHERE c.relnamespace = 'app'::regnamespace AND a.grantee IN (pr, ar))
     OR EXISTS (SELECT 1 FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace AND p.proowner IN (pr, ar))
     OR EXISTS (SELECT 1 FROM pg_policy) THEN
    RAISE EXCEPTION 'version-5 precondition: a helper already holds a privilege or a policy exists';
  END IF;
  -- The exact version-5 runtime grants (00002-00005).
  SELECT string_agg(t.k || ':' || t.privs, ';' ORDER BY t.k COLLATE "C") INTO actual FROM (
    SELECT c.relname AS k, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND a.grantee = rt GROUP BY c.relname
    UNION ALL
    SELECT c.relname || '.' || at.attname, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
    FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND a.grantee = rt GROUP BY c.relname, at.attname) t;
  expected := 'business_profile_access_paths:SELECT;business_profiles:SELECT;business_profiles.description:INSERT,UPDATE;'
    || 'business_profiles.display_name:INSERT,UPDATE;business_profiles.industry:INSERT,UPDATE;'
    || 'business_profiles.legal_name:INSERT,UPDATE;business_profiles.organization_id:INSERT;'
    || 'business_profiles.website:INSERT,UPDATE;group_business_profiles:DELETE,SELECT;'
    || 'group_business_profiles.business_profile_id:INSERT;group_business_profiles.group_id:INSERT;'
    || 'group_business_profiles.organization_id:INSERT;group_memberships:SELECT;groups:SELECT;groups.name:INSERT,UPDATE;'
    || 'groups.organization_id:INSERT;groups.slug:INSERT,UPDATE;identity_principals:INSERT,SELECT,UPDATE;'
    || 'identity_provider_binding:INSERT,SELECT;login_transactions:DELETE,INSERT,SELECT;organization_memberships:SELECT;'
    || 'organizations:SELECT;organizations.group_admins_edit_business_profiles:UPDATE;organizations.name:INSERT,UPDATE;'
    || 'role_bindings:DELETE,SELECT;role_bindings.group_id:INSERT;role_bindings.organization_id:INSERT;'
    || 'role_bindings.role:INSERT,UPDATE;role_bindings.scope_id:INSERT;role_bindings.scope_type:INSERT;'
    || 'role_bindings.source:INSERT;role_bindings.subject_user_id:INSERT;sessions:INSERT,SELECT,UPDATE;'
    || 'users:INSERT,SELECT,UPDATE';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-5 precondition: runtime grants are %, expected %', actual, expected;
  END IF;
END
$$;
-- +goose StatementEnd

-- Context parsers (owner-owned, SECURITY INVOKER). Each validates all four settings, then returns its own field, so any
-- policy referencing any parser validates the whole context shape. '' (or unset) means absent.
-- +goose StatementBegin
CREATE FUNCTION app.ctx_user() RETURNS uuid
  LANGUAGE plpgsql STABLE SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  su text := coalesce(current_setting('app.user_id', true), '');
  so text := coalesce(current_setting('app.organization_id', true), '');
  sg text := coalesce(current_setting('app.group_id', true), '');
  sb text := coalesce(current_setting('app.login_binding', true), '');
  u uuid; o uuid; g uuid;
BEGIN
  IF (su <> '' AND su !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
     OR (so <> '' AND so !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
     OR (sg <> '' AND sg !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$') THEN
    RAISE EXCEPTION 'malformed context identifier' USING ERRCODE = 'invalid_text_representation';
  END IF;
  u := nullif(su, '')::uuid; o := nullif(so, '')::uuid; g := nullif(sg, '')::uuid;
  IF u = '00000000-0000-0000-0000-000000000000' OR o = '00000000-0000-0000-0000-000000000000'
     OR g = '00000000-0000-0000-0000-000000000000' THEN
    RAISE EXCEPTION 'zero context identifier' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF sb <> '' AND sb !~ '^[0-9a-f]{64}$' THEN
    RAISE EXCEPTION 'malformed context binding' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF NOT ((sb = '' AND ((u IS NULL AND o IS NULL AND g IS NULL) OR (u IS NOT NULL AND (g IS NULL OR o IS NOT NULL))))
          OR (sb <> '' AND u IS NULL AND o IS NULL AND g IS NULL)) THEN
    RAISE EXCEPTION 'invalid context shape' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  RETURN u;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION app.ctx_org() RETURNS uuid
  LANGUAGE plpgsql STABLE SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  su text := coalesce(current_setting('app.user_id', true), '');
  so text := coalesce(current_setting('app.organization_id', true), '');
  sg text := coalesce(current_setting('app.group_id', true), '');
  sb text := coalesce(current_setting('app.login_binding', true), '');
  u uuid; o uuid; g uuid;
BEGIN
  IF (su <> '' AND su !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
     OR (so <> '' AND so !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
     OR (sg <> '' AND sg !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$') THEN
    RAISE EXCEPTION 'malformed context identifier' USING ERRCODE = 'invalid_text_representation';
  END IF;
  u := nullif(su, '')::uuid; o := nullif(so, '')::uuid; g := nullif(sg, '')::uuid;
  IF u = '00000000-0000-0000-0000-000000000000' OR o = '00000000-0000-0000-0000-000000000000'
     OR g = '00000000-0000-0000-0000-000000000000' THEN
    RAISE EXCEPTION 'zero context identifier' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF sb <> '' AND sb !~ '^[0-9a-f]{64}$' THEN
    RAISE EXCEPTION 'malformed context binding' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF NOT ((sb = '' AND ((u IS NULL AND o IS NULL AND g IS NULL) OR (u IS NOT NULL AND (g IS NULL OR o IS NOT NULL))))
          OR (sb <> '' AND u IS NULL AND o IS NULL AND g IS NULL)) THEN
    RAISE EXCEPTION 'invalid context shape' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  RETURN o;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION app.ctx_group() RETURNS uuid
  LANGUAGE plpgsql STABLE SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  su text := coalesce(current_setting('app.user_id', true), '');
  so text := coalesce(current_setting('app.organization_id', true), '');
  sg text := coalesce(current_setting('app.group_id', true), '');
  sb text := coalesce(current_setting('app.login_binding', true), '');
  u uuid; o uuid; g uuid;
BEGIN
  IF (su <> '' AND su !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
     OR (so <> '' AND so !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
     OR (sg <> '' AND sg !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$') THEN
    RAISE EXCEPTION 'malformed context identifier' USING ERRCODE = 'invalid_text_representation';
  END IF;
  u := nullif(su, '')::uuid; o := nullif(so, '')::uuid; g := nullif(sg, '')::uuid;
  IF u = '00000000-0000-0000-0000-000000000000' OR o = '00000000-0000-0000-0000-000000000000'
     OR g = '00000000-0000-0000-0000-000000000000' THEN
    RAISE EXCEPTION 'zero context identifier' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF sb <> '' AND sb !~ '^[0-9a-f]{64}$' THEN
    RAISE EXCEPTION 'malformed context binding' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF NOT ((sb = '' AND ((u IS NULL AND o IS NULL AND g IS NULL) OR (u IS NOT NULL AND (g IS NULL OR o IS NOT NULL))))
          OR (sb <> '' AND u IS NULL AND o IS NULL AND g IS NULL)) THEN
    RAISE EXCEPTION 'invalid context shape' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  RETURN g;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION app.ctx_login_binding() RETURNS bytea
  LANGUAGE plpgsql STABLE SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  su text := coalesce(current_setting('app.user_id', true), '');
  so text := coalesce(current_setting('app.organization_id', true), '');
  sg text := coalesce(current_setting('app.group_id', true), '');
  sb text := coalesce(current_setting('app.login_binding', true), '');
  u uuid; o uuid; g uuid;
BEGIN
  IF (su <> '' AND su !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
     OR (so <> '' AND so !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
     OR (sg <> '' AND sg !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$') THEN
    RAISE EXCEPTION 'malformed context identifier' USING ERRCODE = 'invalid_text_representation';
  END IF;
  u := nullif(su, '')::uuid; o := nullif(so, '')::uuid; g := nullif(sg, '')::uuid;
  IF u = '00000000-0000-0000-0000-000000000000' OR o = '00000000-0000-0000-0000-000000000000'
     OR g = '00000000-0000-0000-0000-000000000000' THEN
    RAISE EXCEPTION 'zero context identifier' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF sb <> '' AND sb !~ '^[0-9a-f]{64}$' THEN
    RAISE EXCEPTION 'malformed context binding' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF NOT ((sb = '' AND ((u IS NULL AND o IS NULL AND g IS NULL) OR (u IS NOT NULL AND (g IS NULL OR o IS NOT NULL))))
          OR (sb <> '' AND u IS NULL AND o IS NULL AND g IS NULL)) THEN
    RAISE EXCEPTION 'invalid context shape' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  RETURN decode(nullif(sb, ''), 'hex');
END $$;
-- +goose StatementEnd

-- Authority readers (owned by <db>_policy_reader after creation). rls_org_role: the caller's Organization role in the
-- active Organization, or NULL unless the context is valid as G007 Resolve defines it (an Organization binding; a
-- selected Group of that Organization that the caller may select: Organization owner/admin, or a binding in that exact
-- Group). rls_group_role: the caller's own Group role in p_group, NULL if none or if a Group is selected and p_group is
-- another Group. Neither returns another subject's facts.
-- +goose StatementBegin
CREATE FUNCTION app.rls_org_role() RETURNS text
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text;
BEGIN
  IF u IS NULL OR o IS NULL THEN
    RETURN NULL;
  END IF;
  SELECT rb.role INTO r FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.organization_id = o AND rb.scope_type = 'organization' AND rb.scope_id = o
    AND rb.source = 'membership';
  IF r IS NULL THEN
    RETURN NULL;
  END IF;
  IF g IS NOT NULL THEN
    IF NOT EXISTS (SELECT 1 FROM app.groups gr WHERE gr.id = g AND gr.organization_id = o) THEN
      RETURN NULL;
    END IF;
    IF r NOT IN ('org_owner', 'org_admin') AND NOT EXISTS (
         SELECT 1 FROM app.role_bindings rb
         WHERE rb.subject_user_id = u AND rb.organization_id = o AND rb.scope_type = 'group' AND rb.scope_id = g
           AND rb.source = 'membership') THEN
      RETURN NULL;
    END IF;
  END IF;
  RETURN r;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION app.rls_group_role(p_group uuid) RETURNS text
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text;
BEGIN
  IF u IS NULL OR o IS NULL OR p_group IS NULL OR (g IS NOT NULL AND p_group <> g) THEN
    RETURN NULL;
  END IF;
  SELECT rb.role INTO r FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.organization_id = o AND rb.scope_type = 'group' AND rb.scope_id = p_group
    AND rb.source = 'membership';
  RETURN r;
END $$;
-- +goose StatementEnd

-- Authentication readers (owned by <db>_auth_reader after creation; they ignore ambient settings).
-- E1: the live session of an exact credential hash, joined to its active user (G004's resolution, verbatim).
-- +goose StatementBegin
CREATE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz)
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
END $$;
-- +goose StatementEnd

-- E2: the user mapped to an exact (issuer, subject), an opaque selector core calls only after verifying the provider.
-- +goose StatementBegin
CREATE FUNCTION app.auth_principal_subject(p_issuer text, p_subject text) RETURNS uuid
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  r uuid;
BEGIN
  SELECT p.user_id INTO r FROM app.identity_principals p WHERE p.issuer = p_issuer AND p.subject = p_subject;
  RETURN r;
END $$;
-- +goose StatementEnd

-- E3: the session and user of an exact, unreturned, unexpired provider-logout reference hash.
-- +goose StatementBegin
CREATE FUNCTION app.auth_logout_ref(p_ref bytea, p_at timestamptz) RETURNS TABLE (session_id uuid, user_id uuid)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN QUERY
  SELECT s.id, s.user_id FROM app.sessions s
  WHERE s.provider_logout_ref_hash = p_ref AND s.provider_logout_returned_at IS NULL AND s.provider_logout_ref_expires_at > p_at;
END $$;
-- +goose StatementEnd

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
  B text := '(SELECT app.ctx_login_binding())';
  R text := '(SELECT app.rls_org_role())';
  MGR text;
  vis text;
  linked text;
  tbl text;
  actual text;
  expected text;
BEGIN
  MGR := R || ' IN (''org_owner'', ''org_admin'')';

  -- Parsers: no PUBLIC EXECUTE; runtime and the policy reader may call them.
  REVOKE ALL ON FUNCTION app.ctx_user(), app.ctx_org(), app.ctx_group(), app.ctx_login_binding() FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.ctx_user(), app.ctx_org(), app.ctx_group(), app.ctx_login_binding() TO %I, %I', rt, pr);

  -- Helper grants: schema usage and the exact SELECT columns of each reader.
  EXECUTE format('GRANT USAGE ON SCHEMA app TO %I, %I', pr, ar);
  EXECUTE format('GRANT SELECT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) ON app.role_bindings TO %I', pr);
  EXECUTE format('GRANT SELECT (id, organization_id) ON app.groups TO %I', pr);
  EXECUTE format('GRANT SELECT (id, user_id, credential_hash, csrf_token, provider_session_id, mfa_assurance, auth_time, '
              || 'idle_expires_at, absolute_expires_at, renewed_at, revoked_at, provider_logout_ref_hash, '
              || 'provider_logout_ref_expires_at, provider_logout_returned_at) ON app.sessions TO %I', ar);
  EXECUTE format('GRANT SELECT (id, status, display_name) ON app.users TO %I', ar);
  EXECUTE format('GRANT SELECT (issuer, subject, user_id) ON app.identity_principals TO %I', ar);

  -- The readers pass to their helper; the transient CREATE lets the helper own a function in app and is revoked at once.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I, %I', pr, ar);
  EXECUTE format('ALTER FUNCTION app.rls_org_role() OWNER TO %I', pr);
  EXECUTE format('ALTER FUNCTION app.rls_group_role(uuid) OWNER TO %I', pr);
  EXECUTE format('ALTER FUNCTION app.auth_resolve_session(bytea, timestamptz) OWNER TO %I', ar);
  EXECUTE format('ALTER FUNCTION app.auth_principal_subject(text, text) OWNER TO %I', ar);
  EXECUTE format('ALTER FUNCTION app.auth_logout_ref(bytea, timestamptz) OWNER TO %I', ar);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I, %I', pr, ar);
  -- Only a reader's owner can grant on it: act as each helper (the session's SET chain, migrator -> owner -> helper),
  -- then return to the owner.
  EXECUTE format('SET LOCAL ROLE %I', pr);
  REVOKE ALL ON FUNCTION app.rls_org_role(), app.rls_group_role(uuid) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_org_role(), app.rls_group_role(uuid) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  REVOKE ALL ON FUNCTION app.auth_resolve_session(bytea, timestamptz), app.auth_principal_subject(text, text),
    app.auth_logout_ref(bytea, timestamptz) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.auth_resolve_session(bytea, timestamptz), app.auth_principal_subject(text, text), '
              || 'app.auth_logout_ref(bytea, timestamptz) TO %I', rt);
  EXECUTE format('SET LOCAL ROLE %I', ow);

  -- Runtime grants (§3.1): remove what row-level security cannot filter or no supported path uses.
  EXECUTE format('REVOKE INSERT (name), UPDATE (name) ON app.organizations FROM %I', rt);
  EXECUTE format('REVOKE INSERT (organization_id, name, slug) ON app.groups FROM %I', rt);
  EXECUTE format('REVOKE SELECT, DELETE ON app.role_bindings FROM %I', rt);
  EXECUTE format('REVOKE INSERT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) ON app.role_bindings FROM %I', rt);
  EXECUTE format('GRANT SELECT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) ON app.role_bindings TO %I', rt);
  EXECUTE format('REVOKE INSERT (organization_id, display_name, legal_name, description, industry, website) ON app.business_profiles FROM %I', rt);
  EXECUTE format('REVOKE DELETE ON app.group_business_profiles FROM %I', rt);
  EXECUTE format('REVOKE INSERT (organization_id, group_id, business_profile_id) ON app.group_business_profiles FROM %I', rt);
  EXECUTE format('REVOKE SELECT, INSERT, UPDATE ON app.users FROM %I', rt);
  EXECUTE format('GRANT SELECT (id, status, display_name, disabled_at), INSERT (id, display_name, created_at, updated_at), '
              || 'UPDATE (display_name, updated_at, status, disabled_at) ON app.users TO %I', rt);
  EXECUTE format('REVOKE UPDATE ON app.sessions FROM %I', rt);
  EXECUTE format('GRANT UPDATE (idle_expires_at, renewed_at, revoked_at, revoke_reason, provider_revocation, '
              || 'provider_logout_ref_hash, provider_logout_ref_expires_at, provider_logout_returned_at) ON app.sessions TO %I', rt);
  EXECUTE format('REVOKE UPDATE ON app.identity_principals FROM %I', rt);

  -- Row-level security on the nine tables.
  FOREACH tbl IN ARRAY ARRAY['organizations', 'groups', 'role_bindings', 'business_profiles', 'group_business_profiles',
                             'users', 'sessions', 'identity_principals', 'login_transactions'] LOOP
    EXECUTE format('ALTER TABLE app.%I ENABLE ROW LEVEL SECURITY', tbl);
    EXECUTE format('ALTER TABLE app.%I FORCE ROW LEVEL SECURITY', tbl);
  END LOOP;

  -- Runtime/owner policies (21). Permissive, one per permitted command except the two DELETE policies of
  -- login_transactions.
  EXECUTE format('CREATE POLICY role_bindings_select ON app.role_bindings FOR SELECT TO %I, %I USING ('
    || '%s IS NOT NULL AND organization_id = %s AND source = ''membership'' AND ('
    || '(subject_user_id = %s AND (scope_type = ''organization'' OR %s IS NULL OR group_id = %s))'
    || ' OR (scope_type = ''organization'' AND %s)'
    || ' OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'')))))',
    rt, ow, R, O, U, G, G, MGR, G, G, MGR);
  EXECUTE format('CREATE POLICY role_bindings_lock ON app.role_bindings FOR UPDATE TO %I, %I USING ('
    || '%s IS NOT NULL AND organization_id = %s AND subject_user_id = %s AND source = ''membership'' AND '
    || '(scope_type = ''organization'' OR %s IS NULL OR group_id = %s)) WITH CHECK (false)',
    rt, ow, R, O, U, G, G);
  EXECUTE format('CREATE POLICY organizations_select ON app.organizations FOR SELECT TO %I, %I USING (id = %s AND %s IS NOT NULL)',
    rt, ow, O, R);
  EXECUTE format('CREATE POLICY organizations_update ON app.organizations FOR UPDATE TO %I, %I USING (id = %s AND %s IS NOT NULL) '
    || 'WITH CHECK (id = %s AND %s = ''org_owner'')', rt, ow, O, R, O, R);
  vis := format('organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR id = %s) AND (%s OR app.rls_group_role(id) IS NOT NULL)',
    O, R, G, G, MGR);
  EXECUTE format('CREATE POLICY groups_select ON app.groups FOR SELECT TO %I, %I USING (%s)', rt, ow, vis);
  EXECUTE format('CREATE POLICY groups_update ON app.groups FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR id = %s) AND (%s OR app.rls_group_role(id) IN (''group_owner'', ''group_admin'')))',
    rt, ow, vis, O, R, G, G, MGR);
  EXECUTE format('CREATE POLICY group_business_profiles_select ON app.group_business_profiles FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s IS NULL OR group_id = %s) AND (%s OR app.rls_group_role(group_id) IS NOT NULL))',
    rt, ow, O, R, G, G, MGR);
  linked := 'EXISTS (SELECT 1 FROM app.group_business_profiles a WHERE a.organization_id = business_profiles.organization_id '
    || 'AND a.business_profile_id = business_profiles.id AND (' || G || ' IS NULL OR a.group_id = ' || G || ') '
    || 'AND app.rls_group_role(a.group_id) IN (%s))';
  vis := format('organization_id = %s AND %s IS NOT NULL AND (%s OR %s)', O, R, MGR,
    format(linked, '''group_owner'', ''group_admin'', ''group_member'''));
  EXECUTE format('CREATE POLICY business_profiles_select ON app.business_profiles FOR SELECT TO %I, %I USING (%s)', rt, ow, vis);
  EXECUTE format('CREATE POLICY business_profiles_update ON app.business_profiles FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND %s IS NOT NULL AND (%s OR ((SELECT o.group_admins_edit_business_profiles FROM app.organizations o '
    || 'WHERE o.id = %s) AND %s)))',
    rt, ow, vis, O, R, MGR, O, format(linked, '''group_owner'', ''group_admin'''));
  EXECUTE format('CREATE POLICY users_select ON app.users FOR SELECT TO %I, %I USING (id = %s OR (%s IS NOT NULL AND EXISTS ('
    || 'SELECT 1 FROM app.role_bindings rb WHERE rb.subject_user_id = users.id AND rb.organization_id = %s)))', rt, ow, U, R, O);
  EXECUTE format('CREATE POLICY users_insert ON app.users FOR INSERT TO %I, %I WITH CHECK ('
    || 'id = %s AND %s IS NULL AND status = ''active'' AND disabled_at IS NULL)', rt, ow, U, O);
  EXECUTE format('CREATE POLICY users_update ON app.users FOR UPDATE TO %I, %I USING (id = %s) WITH CHECK (id = %s AND %s IS NULL)',
    rt, ow, U, U, O);
  EXECUTE format('CREATE POLICY sessions_select ON app.sessions FOR SELECT TO %I, %I USING (user_id = %s)', rt, ow, U);
  EXECUTE format('CREATE POLICY sessions_insert ON app.sessions FOR INSERT TO %I, %I WITH CHECK ('
    || 'user_id = %s AND %s IS NULL AND revoked_at IS NULL)', rt, ow, U, O);
  EXECUTE format('CREATE POLICY sessions_update ON app.sessions FOR UPDATE TO %I, %I USING (user_id = %s) WITH CHECK (user_id = %s AND %s IS NULL)',
    rt, ow, U, U, O);
  EXECUTE format('CREATE POLICY identity_principals_select ON app.identity_principals FOR SELECT TO %I, %I USING (user_id = %s)', rt, ow, U);
  EXECUTE format('CREATE POLICY identity_principals_insert ON app.identity_principals FOR INSERT TO %I, %I WITH CHECK ('
    || 'user_id = %s AND %s IS NULL)', rt, ow, U, O);
  EXECUTE format('CREATE POLICY login_transactions_select ON app.login_transactions FOR SELECT TO %I, %I USING (binding_hash = %s)', rt, ow, B);
  EXECUTE format('CREATE POLICY login_transactions_insert ON app.login_transactions FOR INSERT TO %I, %I WITH CHECK (binding_hash = %s)', rt, ow, B);
  EXECUTE format('CREATE POLICY login_transactions_delete ON app.login_transactions FOR DELETE TO %I, %I USING (binding_hash = %s)', rt, ow, B);
  EXECUTE format('CREATE POLICY login_transactions_expire ON app.login_transactions FOR DELETE TO %I, %I USING ('
    || 'expires_at <= now() AND %s IS NULL AND %s IS NULL)', rt, ow, U, B);

  -- Helper policies (5), TO that helper only.
  EXECUTE format('CREATE POLICY role_bindings_policy_reader ON app.role_bindings FOR SELECT TO %I USING ('
    || 'subject_user_id = app.ctx_user() AND organization_id = app.ctx_org())', pr);
  EXECUTE format('CREATE POLICY groups_policy_reader ON app.groups FOR SELECT TO %I USING ('
    || 'id = app.ctx_group() AND organization_id = app.ctx_org())', pr);
  EXECUTE format('CREATE POLICY sessions_auth_reader ON app.sessions FOR SELECT TO %I USING (true)', ar);
  EXECUTE format('CREATE POLICY users_auth_reader ON app.users FOR SELECT TO %I USING (true)', ar);
  EXECUTE format('CREATE POLICY identity_principals_auth_reader ON app.identity_principals FOR SELECT TO %I USING (true)', ar);

  -- Postcondition 1: exactly the nine tables have row-level security, all forced.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  expected := 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,organizations+force,role_bindings+force,sessions+force,users+force';
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
    || 'users_auth_reader:users:r:true:ar;users_insert:users:a:true:ow+rt;users_select:users:r:true:ow+rt;'
    || 'users_update:users:w:true:ow+rt';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'policy inventory is %, expected %', actual, expected;
  END IF;

  -- Postcondition 3: every table/view/sequence and column ACL entry of schema app other than the owner's own, for every
  -- grantee including PUBLIC (label "public"), with symbolic labels.
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
    || 'users.created_at:rt:INSERT;users.disabled_at:rt:SELECT,UPDATE;users.display_name:ar:SELECT;'
    || 'users.display_name:rt:INSERT,SELECT,UPDATE;users.id:ar:SELECT;users.id:rt:INSERT,SELECT;users.status:ar:SELECT;'
    || 'users.status:rt:SELECT,UPDATE;users.updated_at:rt:INSERT,UPDATE';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'version-6 ACLs are %, expected %', actual, expected;
  END IF;

  -- Postcondition 4: the function allowlist: owner, SECURITY DEFINER, LEAKPROOF, volatility, configuration and ACL.
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
    || 'role_bindings_keys_immutable:ow:false:false:v:search_path=pg_catalog:-';
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'function allowlist is %, expected %', actual, expected;
  END IF;

  -- Postcondition 5: schema app privileges: runtime, retention and the helpers hold USAGE only; nobody else but the
  -- owner holds anything.
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
  p record;
  tbl text;
BEGIN
  FOR p IN SELECT pol.polname, c.relname FROM pg_policy pol JOIN pg_class c ON c.oid = pol.polrelid
           WHERE c.relnamespace = 'app'::regnamespace LOOP
    EXECUTE format('DROP POLICY %I ON app.%I', p.polname, p.relname);
  END LOOP;
  -- The readers belong to their helper, and the owner has no INHERIT on it: drop them as the helper (the session's
  -- SET chain, migrator -> owner -> helper), then return to the owner.
  EXECUTE format('SET LOCAL ROLE %I', pr);
  DROP FUNCTION app.rls_group_role(uuid);
  DROP FUNCTION app.rls_org_role();
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.auth_logout_ref(bytea, timestamptz);
  DROP FUNCTION app.auth_principal_subject(text, text);
  DROP FUNCTION app.auth_resolve_session(bytea, timestamptz);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  DROP FUNCTION app.ctx_login_binding();
  DROP FUNCTION app.ctx_group();
  DROP FUNCTION app.ctx_org();
  DROP FUNCTION app.ctx_user();
  FOREACH tbl IN ARRAY ARRAY['organizations', 'groups', 'role_bindings', 'business_profiles', 'group_business_profiles',
                             'users', 'sessions', 'identity_principals', 'login_transactions'] LOOP
    EXECUTE format('ALTER TABLE app.%I NO FORCE ROW LEVEL SECURITY', tbl);
    EXECUTE format('ALTER TABLE app.%I DISABLE ROW LEVEL SECURITY', tbl);
  END LOOP;
  -- Helper grants.
  EXECUTE format('REVOKE ALL (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) ON app.role_bindings FROM %I', pr);
  EXECUTE format('REVOKE ALL (id, organization_id) ON app.groups FROM %I', pr);
  EXECUTE format('REVOKE ALL (id, user_id, credential_hash, csrf_token, provider_session_id, mfa_assurance, auth_time, '
              || 'idle_expires_at, absolute_expires_at, renewed_at, revoked_at, provider_logout_ref_hash, '
              || 'provider_logout_ref_expires_at, provider_logout_returned_at) ON app.sessions FROM %I', ar);
  EXECUTE format('REVOKE ALL (id, status, display_name) ON app.users FROM %I', ar);
  EXECUTE format('REVOKE ALL (issuer, subject, user_id) ON app.identity_principals FROM %I', ar);
  EXECUTE format('REVOKE ALL ON SCHEMA app FROM %I, %I', pr, ar);
  -- Version-5 runtime grants, exactly.
  EXECUTE format('GRANT INSERT (name), UPDATE (name) ON app.organizations TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, name, slug) ON app.groups TO %I', rt);
  EXECUTE format('REVOKE SELECT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) ON app.role_bindings FROM %I', rt);
  EXECUTE format('GRANT SELECT, DELETE ON app.role_bindings TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) ON app.role_bindings TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, display_name, legal_name, description, industry, website) ON app.business_profiles TO %I', rt);
  EXECUTE format('GRANT DELETE ON app.group_business_profiles TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, group_id, business_profile_id) ON app.group_business_profiles TO %I', rt);
  EXECUTE format('REVOKE ALL (id, status, display_name, disabled_at, created_at, updated_at) ON app.users FROM %I', rt);
  EXECUTE format('GRANT SELECT, INSERT, UPDATE ON app.users TO %I', rt);
  EXECUTE format('REVOKE UPDATE (idle_expires_at, renewed_at, revoked_at, revoke_reason, provider_revocation, '
              || 'provider_logout_ref_hash, provider_logout_ref_expires_at, provider_logout_returned_at) ON app.sessions FROM %I', rt);
  EXECUTE format('GRANT UPDATE ON app.sessions TO %I', rt);
  EXECUTE format('GRANT UPDATE ON app.identity_principals TO %I', rt);
END
$$;
-- +goose StatementEnd
