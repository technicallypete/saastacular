-- Tenant-host session binding and proof-time hostname exclusivity (G016 security amendment; ADR-007 "Owner-approved
-- security amendment"; Plan amendments A1/A2; docs/foundation/IDENTITY.md, DOMAINS.md).
--
-- One goose transaction:
--   - app.sessions gains origin (the exact issuing origin) and organization_id (the canonical Organization of that tenant
--     host). Both are NULL together (sessions_binding_shape); 'mcp' rows stay unbound (sessions_mcp_unbound). The
--     sessions_insert policy (FORCE row-level security, so the owner is confined too) admits a new 'browser' row only
--     with both set, so nothing can create a usable unbound browser row; the runtime UPDATE column list (00006) excludes
--     both columns, so a binding never changes after insert. Pre-12 browser rows are not touched: they keep NULL, stay
--     revocable (revoked_at UPDATE) and recoverable (infra/recovery/sql/invalidate.sql) and are refused by the reader.
--   - app.auth_resolve_session(p_hash, p_at, p_origin) replaces the 2-argument reader (dropped; no unbound path remains):
--     a browser row resolves only when it is bound and its origin equals p_origin; it also returns the binding.
--   - Hostname exclusivity moves from claim time to proof time: the partial unique index organization_domains_hostname_live
--     (state <> 'removed') is replaced by organization_domains_hostname_owned (state IN verified/active/suspended), so
--     unverified pending claims of one hostname coexist and the first valid proof to commit wins. app.hostname_state's
--     custom branch reads proven rows only (one row per hostname by that index).
--
-- Locks: ALTER TABLE takes a brief ACCESS EXCLUSIVE lock on app.sessions (nullable columns, no rewrite; the CHECK
-- constraints are validated by one scan) and SHARE ROW EXCLUSIVE on app.organizations (foreign key); the index swap
-- takes SHARE on app.organization_domains. Migrate in a quiescent window: migrate version 12, then deploy the matching
-- code (docs/foundation/MIGRATIONS.md); from that deploy legacy browser sessions are refused.
--
-- Down (development and test only; the runner refuses down/reapply with APP_ENV=production) refuses while any bound
-- session exists (binding history is never discarded), while any live legacy browser session exists (the restored global
-- reader would revive it) or while a hostname has more than one non-removed row, then restores the version-11 shapes.

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  -- Preconditions: the version-11 shapes this migration replaces.
  IF to_regclass('app.organization_domains_hostname_live') IS NULL OR to_regclass('app.organization_domains_hostname_owned') IS NOT NULL
     OR to_regprocedure('app.auth_resolve_session(bytea, timestamptz)') IS NULL
     OR to_regprocedure('app.auth_resolve_session(bytea, timestamptz, text)') IS NOT NULL
     OR to_regprocedure('app.hostname_state(text, text)') IS NULL
     OR EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.sessions'::regclass AND attname IN ('origin', 'organization_id')
                  AND NOT attisdropped) THEN
    RAISE EXCEPTION 'version-11 precondition: session binding or proof-time exclusivity objects already exist';
  END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE app.sessions
  ADD COLUMN origin text NULL CONSTRAINT sessions_origin_valid
    CHECK (char_length(origin) BETWEEN 8 AND 300 AND origin ~ '^https?://[^/?#]+$'),
  ADD COLUMN organization_id uuid NULL CONSTRAINT sessions_organization_fkey
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  ADD CONSTRAINT sessions_binding_shape CHECK ((origin IS NULL) = (organization_id IS NULL)),
  ADD CONSTRAINT sessions_mcp_unbound CHECK (kind <> 'mcp' OR origin IS NULL);

-- The readers, replaced by their owner (the auth reader, through the session's SET chain).
-- +goose StatementBegin
DO $do$
DECLARE
  db text := current_database();
  ow text := current_user;
  rt text := db || '_runtime';
  ar text := db || '_auth_reader';
BEGIN
  EXECUTE format('GRANT SELECT (origin, organization_id) ON app.sessions TO %I', ar);
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.auth_resolve_session(bytea, timestamptz);
  EXECUTE $e1$
CREATE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz, p_origin text)
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
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.hostname_state(p_kind text, p_key text)
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
    FROM app.organization_domains d WHERE d.hostname = (p_key COLLATE "C") AND d.state IN ('verified', 'active', 'suspended');
  END IF;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.auth_resolve_session(bytea, timestamptz, text) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.auth_resolve_session(bytea, timestamptz, text) TO %I', rt);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);
END
$do$;
-- +goose StatementEnd

DROP INDEX app.organization_domains_hostname_live;
CREATE UNIQUE INDEX organization_domains_hostname_owned ON app.organization_domains (hostname)
  WHERE state IN ('verified', 'active', 'suspended');

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
  -- A new browser row must carry its binding; 'mcp' rows exactly as in version 11 and never bound.
  EXECUTE format('ALTER POLICY sessions_insert ON app.sessions WITH CHECK ('
    || 'user_id = %s AND revoked_at IS NULL AND ((kind = ''browser'' AND %s IS NULL AND origin IS NOT NULL '
    || 'AND organization_id IS NOT NULL) '
    || 'OR (kind = ''mcp'' AND %s IS NOT NULL AND %s IS NULL AND %s IS NOT NULL AND origin IS NULL)))', U, O, O, G, R);

  -- Postconditions: the version-12 deltas on the unchanged version-11 inventories.
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace;
  IF actual <> '51' THEN
    RAISE EXCEPTION 'app relation count is %, expected 51 (48 plus the migration ledger)', actual;
  END IF;
  SELECT string_agg(c.relname, ',' ORDER BY c.relname) INTO actual
  FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE i.indrelid = 'app.organization_domains'::regclass;
  IF actual IS DISTINCT FROM 'organization_domains_hostname_generation_key,organization_domains_hostname_owned,'
                             || 'organization_domains_organization,organization_domains_pkey' THEN
    RAISE EXCEPTION 'organization_domains indexes are %', actual;
  END IF;
  SELECT string_agg(p.proname || '(' || oidvectortypes(p.proargtypes) || ')', ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace AND p.proname IN ('auth_resolve_session', 'hostname_state');
  IF actual IS DISTINCT FROM 'auth_resolve_session(bytea, timestamp with time zone, text);hostname_state(text, text)' THEN
    RAISE EXCEPTION 'reader signatures are %', actual;
  END IF;
  SELECT string_agg(conname, ',' ORDER BY conname) INTO actual FROM pg_constraint
  WHERE conrelid = 'app.sessions'::regclass AND conname IN ('sessions_origin_valid', 'sessions_organization_fkey',
                                                          'sessions_binding_shape', 'sessions_mcp_unbound') AND convalidated;
  IF actual IS DISTINCT FROM 'sessions_binding_shape,sessions_mcp_unbound,sessions_organization_fkey,sessions_origin_valid' THEN
    RAISE EXCEPTION 'session binding constraints are %', actual;
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
    || 'sessions.mfa_assurance:ar:SELECT;sessions.organization_id:ar:SELECT;sessions.origin:ar:SELECT;'
    || 'sessions.provider_logout_ref_expires_at:ar:SELECT;'
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
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
BEGIN
  -- Without FORCE the owner's reads are not filtered by row-level security. A refusal rolls back the NO FORCE with
  -- everything else; on success FORCE is restored before any shape changes. Operators clear a refusal only through the
  -- existing paths (revocation, recovery.sh invalidate-auth, removing a claim), never by deleting rows.
  ALTER TABLE app.sessions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.organization_domains NO FORCE ROW LEVEL SECURITY;
  IF EXISTS (SELECT 1 FROM app.sessions WHERE origin IS NOT NULL) THEN
    RAISE EXCEPTION 'refusing to roll back 00012: bound sessions exist (binding history is never discarded)';
  END IF;
  IF EXISTS (SELECT 1 FROM app.sessions WHERE kind = 'browser' AND origin IS NULL AND revoked_at IS NULL
               AND idle_expires_at > now() AND absolute_expires_at > now()) THEN
    RAISE EXCEPTION 'refusing to roll back 00012: live legacy browser sessions would become usable by the global reader';
  END IF;
  IF EXISTS (SELECT 1 FROM app.organization_domains WHERE state <> 'removed' GROUP BY hostname HAVING count(*) > 1) THEN
    RAISE EXCEPTION 'refusing to roll back 00012: a hostname has more than one non-removed row';
  END IF;
  ALTER TABLE app.sessions FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.organization_domains FORCE ROW LEVEL SECURITY;

  -- The exact version-11 sessions_insert predicate (00010).
  EXECUTE format('ALTER POLICY sessions_insert ON app.sessions WITH CHECK ('
    || 'user_id = %s AND revoked_at IS NULL AND ((kind = ''browser'' AND %s IS NULL) '
    || 'OR (kind = ''mcp'' AND %s IS NOT NULL AND %s IS NULL AND %s IS NOT NULL)))', U, O, O, G, R);

  -- The version-11 readers, restored by their owner.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.auth_resolve_session(bytea, timestamptz, text);
  EXECUTE $e1$
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
    AND u.status = 'active' AND s.kind = 'browser';
END $$
$e1$;
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.hostname_state(p_kind text, p_key text)
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
  REVOKE ALL ON FUNCTION app.auth_resolve_session(bytea, timestamptz) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.auth_resolve_session(bytea, timestamptz) TO %I', rt);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);
  EXECUTE format('REVOKE SELECT (origin, organization_id) ON app.sessions FROM %I', ar);

  DROP INDEX app.organization_domains_hostname_owned;
  CREATE UNIQUE INDEX organization_domains_hostname_live ON app.organization_domains (hostname) WHERE state <> 'removed';
  ALTER TABLE app.sessions DROP CONSTRAINT sessions_binding_shape, DROP CONSTRAINT sessions_mcp_unbound,
    DROP COLUMN origin, DROP COLUMN organization_id;
END
$do$;
-- +goose StatementEnd
