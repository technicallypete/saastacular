-- Opaque platform tenant addresses (G057 Plan r5 §3.1, §3.5; Spec "Opaque platform tenant addresses — owner amendment,
-- 2026-10-23"; ADR-007 opaque-address amendment; docs/foundation/DOMAINS.md, DATABASE-ISOLATION.md, TENANCY.md).
--
-- One goose transaction:
--   - Precondition: the owner, the five roles with the 00018/00020 attribute rules, version 20, 88 relations in schema
--     app, no 00021 object, the 00012 app.hostname_state body and the 00018 7-argument provisioning body (md5(prosrc)),
--     FORCE on app.organizations and app.signup_sessions, and the helper inventories (policy reader 9, claim writer 7,
--     auth reader 8, provision writer 3).
--   - app.organizations.platform_label (text COLLATE "C" NULL): the slug grammar and the exact 00011 reserved list as
--     CHECKs, and a partial unique index (88 -> 89 relations). NULL = no platform host. The >= 128-bit rule for
--     production labels is the core generator's (Spec), not a CHECK, so fixed synthetic test labels stay legal.
--   - Backfill: every existing row receives a 40-character lowercase hex label (160 random bits from two
--     gen_random_uuid() values, version and variant nibbles excluded), owner-side inside a NO FORCE ... FORCE bracket
--     (the owner is bound by FORCE and every owner UPDATE policy needs a request context). No audit row: the update
--     capture fires only for name, slug and the two policy columns.
--   - app.hostname_state (auth reader, replaced in place): the platform branch reads platform_label; the slug never routes.
--     The auth reader gains SELECT (platform_label), the one column its body now reads under its own definer role.
--   - app.signup_provision_organization (provision writer): the 7-argument body is replaced by an 8-argument one taking
--     the core-generated label, inserting it with the Organization and returning target_origin (created: the given
--     target; already_created: the stored origin, so the writer never reads app.organizations). A unique violation on
--     the label index is the closed outcome label_taken (core's bounded retry), on the slug index slug_taken as before.
--     The writer gains INSERT (platform_label) only: no SELECT or UPDATE of the column, no new policy for it.
--   - organizations_provision_insert (writer-only) additionally requires platform_label IS NOT NULL.
--   - app.signup_slug_taken(p_slug) (auth reader, new): slug availability, so app.hostname_state stays a routing reader.
--   - Postconditions. No role, no runtime grant, no UPDATE (platform_label) for any role: labels are immutable at runtime.
--
-- Locks: brief ACCESS EXCLUSIVE on app.organizations (nullable column, no rewrite; the CHECKs are validated by one scan),
-- one backfill UPDATE of every Organization row and the index build. Migrate in a quiescent window, then deploy the
-- matching code: version-20 code calls the 7-argument provisioning body and reaches tenants by slug host.
--
-- Down (development and test only; the runner refuses down/reapply with APP_ENV=production) refuses while a live or
-- re-mintable signup continuation references a platform label (inspected with FORCE lifted inside the transaction; a
-- refusal rolls the lift back): a provisioned, unrevoked row whose target origin is label-derived (or NULL) and whose root
-- credential is live or whose continuation is in flight (or NULL). Dead signup history and slug-era origins never refuse.
-- Then it restores the exact version-20 shapes and drops the labels.

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

  -- Version 20: the migration history ends at 20, no 00021 object exists, the replaced bodies are the reviewed ones.
  IF (SELECT max(version_id) FROM app.goose_db_version) IS DISTINCT FROM 20 THEN
    RAISE EXCEPTION 'version-20 precondition: the current version is %', (SELECT max(version_id) FROM app.goose_db_version);
  END IF;
  IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.organizations'::regclass AND attname = 'platform_label' AND NOT attisdropped)
     OR to_regclass('app.organizations_platform_label_key') IS NOT NULL
     OR to_regprocedure('app.signup_slug_taken(text)') IS NOT NULL
     OR to_regprocedure('app.signup_provision_organization(bytea, text, text, text, bytea, bytea, text, timestamptz)') IS NOT NULL
     OR to_regprocedure('app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz)') IS NULL
     OR to_regprocedure('app.hostname_state(text, text)') IS NULL THEN
    RAISE EXCEPTION 'version-20 precondition: platform-label objects already exist or the version-20 readers are missing';
  END IF;
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 88 THEN
    RAISE EXCEPTION 'version-20 precondition: schema app holds % relations, expected 88', n;
  END IF;
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.hostname_state(text, text)'::regprocedure) IS DISTINCT FROM 'c084e0c8ef21c161937b918a37738500'
     OR (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz)'::regprocedure)
        IS DISTINCT FROM 'a46e8536f16795c164ae010460fbeb18'
     OR (SELECT proowner FROM pg_proc WHERE oid = 'app.hostname_state(text, text)'::regprocedure) <> ar
     OR (SELECT proowner FROM pg_proc WHERE oid = 'app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz)'::regprocedure) <> pw THEN
    RAISE EXCEPTION 'version-20 precondition: the hostname reader or the provisioning body differs';
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = pr) <> 9 OR (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7
     OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 8 OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 3 THEN
    RAISE EXCEPTION 'version-20 precondition: the helper inventories differ';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_class WHERE oid IN ('app.organizations'::regclass, 'app.signup_sessions'::regclass)
               AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION 'version-20 precondition: FORCE on app.organizations or app.signup_sessions differs';
  END IF;
END
$$;
-- +goose StatementEnd

-- The platform label (Plan r5 §3.1 steps 1-2): the slug's grammar and reserved floor, unique where present.
ALTER TABLE app.organizations ADD COLUMN platform_label text COLLATE "C" NULL
  CONSTRAINT organizations_platform_label_format
    CHECK (char_length(platform_label) BETWEEN 1 AND 63 AND platform_label ~ '^[0-9a-z]+(-[0-9a-z]+)*$')
  CONSTRAINT organizations_platform_label_reserved CHECK (platform_label NOT IN (
    'account', 'admin', 'api', 'app', 'auth', 'billing', 'blueprint', 'cdn', 'custom', 'fonts', 'health', 'healthz',
    'internal', 'login', 'logout', 'mail', 'mcp', 'members', 'message', 'messages', 'new', 'ns', 'ns1', 'ns2',
    'onboarding', 'readyz', 'rpc', 'settings', 'sign-in', 'smtp', 'sse', 'static', 'status', 'system', 'www'));
CREATE UNIQUE INDEX organizations_platform_label_key ON app.organizations (platform_label) WHERE platform_label IS NOT NULL;

-- The backfill (Plan r5 §3.1 step 3): the owner is bound by FORCE, so FORCE is lifted only for this statement, inside this
-- transaction. Each label is the first 20 random hex digits of each of two v4 UUIDs (the version nibble, position 13, and
-- the variant nibble, position 17, are skipped): 40 hex digits, 160 random bits. A collision fails the unique index and the
-- whole migration (rerun).
-- +goose StatementBegin
DO $$
DECLARE
  n bigint;
BEGIN
  ALTER TABLE app.organizations NO FORCE ROW LEVEL SECURITY;
  WITH g AS (
    SELECT o.id, replace(gen_random_uuid()::text, '-', '') AS a, replace(gen_random_uuid()::text, '-', '') AS b
    FROM app.organizations o WHERE o.platform_label IS NULL
  )
  UPDATE app.organizations o
  SET platform_label = substr(g.a, 1, 12) || substr(g.a, 14, 3) || substr(g.a, 18, 5)
                    || substr(g.b, 1, 12) || substr(g.b, 14, 3) || substr(g.b, 18, 5)
  FROM g WHERE g.id = o.id;
  SELECT count(*) INTO n FROM app.organizations WHERE platform_label IS NULL OR platform_label !~ '^[0-9a-f]{40}$';
  ALTER TABLE app.organizations FORCE ROW LEVEL SECURITY;
  IF n <> 0 THEN
    RAISE EXCEPTION '00021 backfill: % organizations without a 40-digit hex platform label', n;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  ar text := current_database() || '_auth_reader';
  pw text := current_database() || '_provision_writer';
BEGIN
  -- The two column grants (Plan r5 §3.1 steps 4-5): the reader's platform branch and the writer's INSERT.
  EXECUTE format('GRANT SELECT (platform_label) ON app.organizations TO %I', ar);
  EXECUTE format('GRANT INSERT (platform_label) ON app.organizations TO %I', pw);
  -- D-12: production signup, the only INSERT path of this writer-only policy, must carry a platform label.
  ALTER POLICY organizations_provision_insert ON app.organizations WITH CHECK (
    pg_trigger_depth() = 0 AND app.ctx_user() IS NOT NULL AND app.ctx_org() IS NULL AND slug IS NOT NULL AND platform_label IS NOT NULL);

  -- The auth reader's two bodies, replaced and created by their owner (the 00012 SET LOCAL ROLE pattern).
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $r$
CREATE OR REPLACE FUNCTION app.hostname_state(p_kind text, p_key text)
  RETURNS TABLE (organization_id uuid, kind text, routing boolean, issuance boolean)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF p_kind = 'platform' THEN
    RETURN QUERY SELECT o.id, 'platform'::text, true, false FROM app.organizations o WHERE o.platform_label = (p_key COLLATE "C");
  ELSIF p_kind = 'custom' THEN
    RETURN QUERY
    SELECT d.organization_id, 'custom'::text,
           d.state = 'active' AND d.verified_until > now() AND d.tls_state = 'ready' AND d.tls_not_after > now(),
           d.state IN ('verified', 'active', 'suspended') AND d.verified_until > now()
    FROM app.organization_domains d WHERE d.hostname = (p_key COLLATE "C") AND d.state IN ('verified', 'active', 'suspended');
  END IF;
END $$
$r$;
  -- Slug availability for signup (D-3): the slug is metadata, so its existence is no longer a routing question.
  EXECUTE $r$
CREATE FUNCTION app.signup_slug_taken(p_slug text) RETURNS boolean
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN EXISTS (SELECT 1 FROM app.organizations o WHERE o.slug = (p_slug COLLATE "C"));
END $$
$r$;
  REVOKE ALL ON FUNCTION app.signup_slug_taken(text) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.signup_slug_taken(text) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);

  -- The provisioning body, replaced by its owner (the 00018 pattern): one more argument, one more result column.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pw);
  EXECUTE format('SET LOCAL ROLE %I', pw);
  DROP FUNCTION app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz);
  -- Constrained provisioning (00018 Plan r5 §3.4, G057 Plan r5 §3.1 step 5): IDENTITY shape; lock order users -> own
  -- signup row -> organizations -> role_bindings. A NULL label reaches the INSERT and is refused there (D-12, 42501).
  EXECUTE $r$
CREATE FUNCTION app.signup_provision_organization(p_hash bytea, p_name text, p_slug text, p_label text, p_intent bytea,
                                                  p_ref bytea, p_target text, p_now timestamptz)
  RETURNS TABLE (outcome text, organization_id uuid, slug text, generation integer, target_origin text)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  st text;
  ss app.signup_sessions%ROWTYPE;
  oid_new uuid;
  cn text;
BEGIN
  IF u IS NULL OR app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL OR app.ctx_login_binding() IS NOT NULL THEN
    RAISE EXCEPTION 'signup context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'signup_context';
  END IF;
  IF p_hash IS NULL OR p_intent IS NULL OR p_ref IS NULL OR p_now IS NULL OR p_target IS NULL THEN
    RETURN QUERY SELECT 'signup_invalid'::text, NULL::uuid, NULL::text, NULL::integer, NULL::text;
    RETURN;
  END IF;
  -- (1) the creator, under a lock consistent with DisableUser (FOR UPDATE there, FOR SHARE here).
  PERFORM 1 FROM app.users us WHERE us.id = u FOR SHARE;
  SELECT us.status INTO st FROM app.users us WHERE us.id = u;
  IF st IS DISTINCT FROM 'active' THEN
    RETURN QUERY SELECT 'user_disabled'::text, NULL::uuid, NULL::text, NULL::integer, NULL::text;
    RETURN;
  END IF;
  -- (2) the caller's own live signup row, locked.
  SELECT s.* INTO ss FROM app.signup_sessions s
  WHERE s.credential_hash = p_hash AND s.user_id = u AND s.revoked_at IS NULL AND s.expires_at > p_now FOR UPDATE;
  IF NOT FOUND THEN
    RETURN QUERY SELECT 'signup_invalid'::text, NULL::uuid, NULL::text, NULL::integer, NULL::text;
    RETURN;
  END IF;
  -- (3) an already provisioned row: the same intent re-mints generation g+1 (every secret of g stops matching) and
  -- returns the stored target origin (the label of this call is discarded); a changed payload is intent_used.
  IF ss.provisioned_at IS NOT NULL THEN
    IF ss.intent_hash = p_intent THEN
      UPDATE app.signup_sessions s SET ref_hash = p_ref, continuation_expires_at = p_now + interval '120 seconds',
        ref_consumed_at = NULL, challenge_hash = NULL, nonce_hash = NULL, nonce_consumed_at = NULL, proof_hash = NULL,
        proof_consumed_at = NULL, completed_at = NULL, continuation_generation = s.continuation_generation + 1
      WHERE s.id = ss.id;
      RETURN QUERY SELECT 'already_created'::text, ss.organization_id, p_slug, ss.continuation_generation + 1, ss.target_origin;
      RETURN;
    END IF;
    RETURN QUERY SELECT 'intent_used'::text, NULL::uuid, NULL::text, NULL::integer, NULL::text;
    RETURN;
  END IF;
  -- (4) grammar and reserved labels are the 00011/00021 CHECKs; (5) the Organization with its platform label (defaults:
  -- invite_only, created_after_ledger, override false); each unique index decides one winner, a loser retains no write.
  oid_new := gen_random_uuid();
  BEGIN
    INSERT INTO app.organizations (id, name, slug, platform_label) VALUES (oid_new, p_name, p_slug, p_label);
  EXCEPTION WHEN unique_violation THEN
    GET STACKED DIAGNOSTICS cn = CONSTRAINT_NAME;
    IF cn = 'organizations_platform_label_key' THEN
      RETURN QUERY SELECT 'label_taken'::text, NULL::uuid, NULL::text, NULL::integer, NULL::text;
    ELSE
      RETURN QUERY SELECT 'slug_taken'::text, NULL::uuid, NULL::text, NULL::integer, NULL::text;
    END IF;
    RETURN;
  END;
  -- (6) the creator's org_owner binding.
  INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
  VALUES (gen_random_uuid(), oid_new, u, 'organization', oid_new, NULL, 'org_owner', 'membership');
  -- (7) the intent, result, target and generation 1 with its initial reference.
  UPDATE app.signup_sessions s SET intent_hash = p_intent, organization_id = oid_new, provisioned_at = p_now,
    target_origin = p_target, continuation_generation = 1, continuation_expires_at = p_now + interval '120 seconds',
    ref_hash = p_ref
  WHERE s.id = ss.id;
  RETURN QUERY SELECT 'created'::text, oid_new, p_slug, 1, p_target;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.signup_provision_organization(bytea, text, text, text, bytea, bytea, text, timestamptz) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.signup_provision_organization(bytea, text, text, text, bytea, bytea, text, timestamptz) '
              || 'TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pw);
END
$do$;
-- +goose StatementEnd

-- Postconditions (TestPlatformLabelsCatalog repeats and extends them).
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
  rn regrole := (db || '_retention')::regrole;
  actual text;
  n bigint;
BEGIN
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 89 THEN
    RAISE EXCEPTION '00021 postcondition: schema app holds % relations, expected 89', n;
  END IF;
  SELECT string_agg(conname::text || ':' || convalidated::text, ',' ORDER BY conname::text COLLATE "C") INTO actual FROM pg_constraint
  WHERE conrelid = 'app.organizations'::regclass AND conname::text LIKE 'organizations_platform_label%';
  IF actual IS DISTINCT FROM 'organizations_platform_label_format:true,organizations_platform_label_reserved:true'
     OR NOT EXISTS (SELECT 1 FROM pg_index i WHERE i.indexrelid = 'app.organizations_platform_label_key'::regclass AND i.indisunique
                      AND i.indrelid = 'app.organizations'::regclass) THEN
    RAISE EXCEPTION '00021 postcondition: the platform-label constraints or index differ (%)', actual;
  END IF;
  -- The three definer bodies: owners, hardened search_path, no PUBLIC EXECUTE, the reviewed text; the 7-argument body gone.
  SELECT string_agg(format('%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ar THEN 'ar' WHEN pw THEN 'pw' ELSE 'other' END, p.prosecdef::text, md5(p.prosrc)), ';' ORDER BY p.proname)
    INTO actual FROM pg_proc p
  WHERE p.oid IN ('app.hostname_state(text, text)'::regprocedure, 'app.signup_slug_taken(text)'::regprocedure,
                  'app.signup_provision_organization(bytea, text, text, text, bytea, bytea, text, timestamptz)'::regprocedure);
  IF actual IS DISTINCT FROM 'hostname_state:ar:true:82138e2efb8c18c12eaef4f84793333b;signup_provision_organization:pw:true:b92e32a8936f2b04b5b649a5a60613d9;'
                          || 'signup_slug_taken:ar:true:ec7ee37a2169822a22bf93aacebcd864'
     OR EXISTS (SELECT 1 FROM pg_proc p WHERE p.proname IN ('hostname_state', 'signup_slug_taken', 'signup_provision_organization')
                  AND p.pronamespace = 'app'::regnamespace
                  AND (p.proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, pg_temp'] OR has_function_privilege('public', p.oid, 'EXECUTE')))
     OR to_regprocedure('app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz)') IS NOT NULL THEN
    RAISE EXCEPTION '00021 postcondition: the definer bodies are %', actual;
  END IF;
  -- The exact privilege delta: ar SELECT and pw INSERT on the column, nothing else; no UPDATE of it for any role.
  IF NOT has_column_privilege(ar, 'app.organizations', 'platform_label', 'SELECT')
     OR NOT has_column_privilege(pw, 'app.organizations', 'platform_label', 'INSERT')
     OR has_column_privilege(pw, 'app.organizations', 'platform_label', 'SELECT')
     OR has_column_privilege(rt, 'app.organizations', 'platform_label', 'UPDATE')
     OR has_column_privilege(ar, 'app.organizations', 'platform_label', 'UPDATE')
     OR has_column_privilege(pw, 'app.organizations', 'platform_label', 'UPDATE')
     OR has_column_privilege(pr, 'app.organizations', 'platform_label', 'SELECT, INSERT, UPDATE, REFERENCES')
     OR has_column_privilege(cw, 'app.organizations', 'platform_label', 'SELECT, INSERT, UPDATE, REFERENCES')
     OR has_column_privilege(rn, 'app.organizations', 'platform_label', 'SELECT, INSERT, UPDATE, REFERENCES')
     OR EXISTS (SELECT 1 FROM pg_attribute at, aclexplode(at.attacl) a
                WHERE at.attrelid = 'app.organizations'::regclass AND at.attname = 'platform_label' AND a.grantee NOT IN (ar, pw)) THEN
    RAISE EXCEPTION '00021 postcondition: the platform_label privileges differ';
  END IF;
  IF position('platform_label IS NOT NULL' IN (SELECT pg_get_expr(polwithcheck, polrelid) FROM pg_policy
                                                 WHERE polrelid = 'app.organizations'::regclass AND polname = 'organizations_provision_insert')) = 0
     OR (SELECT polroles FROM pg_policy WHERE polrelid = 'app.organizations'::regclass AND polname = 'organizations_provision_insert')
        IS DISTINCT FROM ARRAY[pw::oid] THEN
    RAISE EXCEPTION '00021 postcondition: organizations_provision_insert differs';
  END IF;
  IF has_schema_privilege(ar, 'app', 'CREATE') OR has_schema_privilege(pw, 'app', 'CREATE')
     OR EXISTS (SELECT 1 FROM pg_class WHERE oid = 'app.organizations'::regclass AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION '00021 postcondition: helper CREATE on app or FORCE on app.organizations differs';
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = pr) <> 9 OR (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7
     OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 9 OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 3 THEN
    RAISE EXCEPTION '00021 postcondition: the helper inventories differ';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  ar text := current_database() || '_auth_reader';
  pw text := current_database() || '_provision_writer';
  n bigint;
BEGIN
  -- D-4: refuse while a live or re-mintable signup continuation references a platform label: (i) a provisioned,
  -- unrevoked row whose root credential is live (the already_created branch re-mints a continuation) or (ii) whose
  -- continuation is in flight. NULL target_origin / continuation_expires_at on a provisioned row is ambiguous and refuses.
  -- FORCE is lifted only inside this transaction; a refusal rolls the lift back. Operators clear a refusal only through
  -- the existing paths (revocation, expiry), never by deleting rows.
  ALTER TABLE app.signup_sessions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.organizations NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n FROM app.signup_sessions s JOIN app.organizations o ON o.id = s.organization_id
  WHERE s.provisioned_at IS NOT NULL
    AND s.revoked_at IS NULL
    AND (s.target_origin IS NULL OR s.target_origin LIKE ('https://' || o.platform_label || '.%'))
    AND (s.expires_at > now()
         OR COALESCE(s.continuation_expires_at > now() AND s.completed_at IS NULL, true));
  IF n > 0 THEN
    RAISE EXCEPTION 'refusing to roll back 00021: % live or re-mintable signup continuation(s) reference platform labels', n;
  END IF;
  ALTER TABLE app.signup_sessions FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.organizations FORCE ROW LEVEL SECURITY;

  -- The version-20 auth-reader bodies, restored by their owner.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.signup_slug_taken(text);
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
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);

  -- The version-20 (00018) provisioning body, restored by its owner with its exact grants.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pw);
  EXECUTE format('SET LOCAL ROLE %I', pw);
  DROP FUNCTION app.signup_provision_organization(bytea, text, text, text, bytea, bytea, text, timestamptz);
  EXECUTE $r$
CREATE FUNCTION app.signup_provision_organization(p_hash bytea, p_name text, p_slug text, p_intent bytea, p_ref bytea,
                                                  p_target text, p_now timestamptz)
  RETURNS TABLE (outcome text, organization_id uuid, slug text, generation integer)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  st text;
  ss app.signup_sessions%ROWTYPE;
  oid_new uuid;
BEGIN
  IF u IS NULL OR app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL OR app.ctx_login_binding() IS NOT NULL THEN
    RAISE EXCEPTION 'signup context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'signup_context';
  END IF;
  IF p_hash IS NULL OR p_intent IS NULL OR p_ref IS NULL OR p_now IS NULL OR p_target IS NULL THEN
    RETURN QUERY SELECT 'signup_invalid'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END IF;
  -- (1) the creator, under a lock consistent with DisableUser (FOR UPDATE there, FOR SHARE here).
  PERFORM 1 FROM app.users us WHERE us.id = u FOR SHARE;
  SELECT us.status INTO st FROM app.users us WHERE us.id = u;
  IF st IS DISTINCT FROM 'active' THEN
    RETURN QUERY SELECT 'user_disabled'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END IF;
  -- (2) the caller's own live signup row, locked.
  SELECT s.* INTO ss FROM app.signup_sessions s
  WHERE s.credential_hash = p_hash AND s.user_id = u AND s.revoked_at IS NULL AND s.expires_at > p_now FOR UPDATE;
  IF NOT FOUND THEN
    RETURN QUERY SELECT 'signup_invalid'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END IF;
  -- (3) an already provisioned row: the same intent re-mints generation g+1 (every secret of g stops matching); a
  -- changed payload is intent_used.
  IF ss.provisioned_at IS NOT NULL THEN
    IF ss.intent_hash = p_intent THEN
      UPDATE app.signup_sessions s SET ref_hash = p_ref, continuation_expires_at = p_now + interval '120 seconds',
        ref_consumed_at = NULL, challenge_hash = NULL, nonce_hash = NULL, nonce_consumed_at = NULL, proof_hash = NULL,
        proof_consumed_at = NULL, completed_at = NULL, continuation_generation = s.continuation_generation + 1
      WHERE s.id = ss.id;
      RETURN QUERY SELECT 'already_created'::text, ss.organization_id, p_slug, ss.continuation_generation + 1;
      RETURN;
    END IF;
    RETURN QUERY SELECT 'intent_used'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END IF;
  -- (4) grammar and reserved labels are the 00011 CHECKs; (5) the Organization (defaults: invite_only,
  -- created_after_ledger, override false); the unique slug index decides one winner, a loser retains no write.
  oid_new := gen_random_uuid();
  BEGIN
    INSERT INTO app.organizations (id, name, slug) VALUES (oid_new, p_name, p_slug);
  EXCEPTION WHEN unique_violation THEN
    RETURN QUERY SELECT 'slug_taken'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END;
  -- (6) the creator's org_owner binding.
  INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
  VALUES (gen_random_uuid(), oid_new, u, 'organization', oid_new, NULL, 'org_owner', 'membership');
  -- (7) the intent, result, target and generation 1 with its initial reference.
  UPDATE app.signup_sessions s SET intent_hash = p_intent, organization_id = oid_new, provisioned_at = p_now,
    target_origin = p_target, continuation_generation = 1, continuation_expires_at = p_now + interval '120 seconds',
    ref_hash = p_ref
  WHERE s.id = ss.id;
  RETURN QUERY SELECT 'created'::text, oid_new, p_slug, 1;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz) '
              || 'TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pw);

  -- The exact version-20 (00018) policy text, then the two column grants (explicitly, before the column goes), the index
  -- and the column with its CHECKs.
  ALTER POLICY organizations_provision_insert ON app.organizations WITH CHECK (
    pg_trigger_depth() = 0 AND app.ctx_user() IS NOT NULL AND app.ctx_org() IS NULL AND slug IS NOT NULL);
  EXECUTE format('REVOKE SELECT (platform_label) ON app.organizations FROM %I', ar);
  EXECUTE format('REVOKE INSERT (platform_label) ON app.organizations FROM %I', pw);
  DROP INDEX app.organizations_platform_label_key;
  ALTER TABLE app.organizations DROP COLUMN platform_label;
END
$do$;
-- +goose StatementEnd
