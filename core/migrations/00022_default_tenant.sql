-- Default-tenant root alias: the one startup lookup of an Organization id by its slug (G071 Plan r4 §3, S2; Spec
-- "Default-tenant root alias"; ADR-010; docs/foundation/DOMAINS.md, DATABASE-ISOLATION.md).
--
-- One goose transaction, no DML:
--   - Precondition: the owner, the six roles, version 21, 89 relations in schema app, no 00022 object, the 00021
--     app.hostname_state and app.signup_slug_taken bodies (owner and md5(prosrc)), the auth reader's SELECT (id, slug)
--     on app.organizations (00011) and its inventory of 9 bodies.
--   - app.organization_id_by_slug(p_slug) (auth reader, new): the id of the Organization whose slug equals p_slug
--     exactly (COLLATE "C": no case folding), or NULL. Core calls it once per process at startup when
--     DEFAULT_TENANT_SLUG is set; no HTTP route reaches it. Its privilege class is that of app.signup_slug_taken (an
--     auth-reader body under the 00011 USING (true) reader policy, runtime EXECUTE), one column wider in its result.
--   - Postconditions. No table, column, policy, role or runtime table grant changes; still 89 relations; the auth
--     reader owns exactly 10 bodies.
--
-- Locks: none beyond the catalog. Migrate, then deploy the matching code: version-21 code never calls the function;
-- version-22 code with DEFAULT_TENANT_SLUG set against a version-21 database fails its startup lookup (42883 is a store
-- error) and exits default_tenant_unavailable.
--
-- Down (development and test only; the runner refuses down/reapply with APP_ENV=production) drops the function as its
-- owner and restores the exact version-21 state. No data depends on it.

-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ar regrole;
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
  ar := (db || '_auth_reader')::regrole;

  -- Version 21: the migration history ends at 21, no 00022 object exists, the 00021 readers are the reviewed ones.
  IF (SELECT max(version_id) FROM app.goose_db_version) IS DISTINCT FROM 21 THEN
    RAISE EXCEPTION 'version-21 precondition: the current version is %', (SELECT max(version_id) FROM app.goose_db_version);
  END IF;
  IF to_regprocedure('app.organization_id_by_slug(text)') IS NOT NULL
     OR to_regprocedure('app.signup_slug_taken(text)') IS NULL
     OR to_regprocedure('app.hostname_state(text, text)') IS NULL THEN
    RAISE EXCEPTION 'version-21 precondition: the default-tenant reader already exists or the version-21 readers are missing';
  END IF;
  IF (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.hostname_state(text, text)'::regprocedure) IS DISTINCT FROM '82138e2efb8c18c12eaef4f84793333b'
     OR (SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.signup_slug_taken(text)'::regprocedure) IS DISTINCT FROM 'ec7ee37a2169822a22bf93aacebcd864'
     OR (SELECT proowner FROM pg_proc WHERE oid = 'app.hostname_state(text, text)'::regprocedure) <> ar
     OR (SELECT proowner FROM pg_proc WHERE oid = 'app.signup_slug_taken(text)'::regprocedure) <> ar THEN
    RAISE EXCEPTION 'version-21 precondition: the hostname or slug-availability reader differs';
  END IF;
  IF NOT has_column_privilege(ar, 'app.organizations', 'id', 'SELECT')
     OR NOT has_column_privilege(ar, 'app.organizations', 'slug', 'SELECT') THEN
    RAISE EXCEPTION 'version-21 precondition: the auth reader lacks SELECT (id, slug) on app.organizations';
  END IF;
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 89 THEN
    RAISE EXCEPTION 'version-21 precondition: schema app holds % relations, expected 89', n;
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 9 THEN
    RAISE EXCEPTION 'version-21 precondition: the auth reader owns % bodies, expected 9', (SELECT count(*) FROM pg_proc WHERE proowner = ar);
  END IF;
END
$$;
-- +goose StatementEnd

-- The reader (Plan r4 §3 step 2), created by its owner (the 00012/00021 SET LOCAL ROLE pattern) with the temporary CREATE.
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  ar text := current_database() || '_auth_reader';
BEGIN
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $r$
CREATE FUNCTION app.organization_id_by_slug(p_slug text) RETURNS uuid
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN (SELECT o.id FROM app.organizations o WHERE o.slug = (p_slug COLLATE "C"));
END $$
$r$;
  REVOKE ALL ON FUNCTION app.organization_id_by_slug(text) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.organization_id_by_slug(text) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);
END
$do$;
-- +goose StatementEnd

-- Postconditions (TestDefaultTenantCatalog repeats and extends them).
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
  f oid;
  actual text;
  n bigint;
BEGIN
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 89 THEN
    RAISE EXCEPTION '00022 postcondition: schema app holds % relations, expected 89', n;
  END IF;
  f := to_regprocedure('app.organization_id_by_slug(text)');
  IF f IS NULL THEN
    RAISE EXCEPTION '00022 postcondition: app.organization_id_by_slug(text) is missing';
  END IF;
  -- The new body and the two 00021 auth-reader bodies: owners, definer flag, the reviewed text.
  SELECT string_agg(format('%s:%s:%s:%s', p.proname,
           CASE p.proowner WHEN ar THEN 'ar' ELSE 'other' END, p.prosecdef::text, md5(p.prosrc)), ';' ORDER BY p.proname)
    INTO actual FROM pg_proc p
  WHERE p.oid IN (f, 'app.hostname_state(text, text)'::regprocedure, 'app.signup_slug_taken(text)'::regprocedure);
  IF actual IS DISTINCT FROM 'hostname_state:ar:true:82138e2efb8c18c12eaef4f84793333b;'
                          || 'organization_id_by_slug:ar:true:4a63253c508682c39230b2fa6601a697;'
                          || 'signup_slug_taken:ar:true:ec7ee37a2169822a22bf93aacebcd864' THEN
    RAISE EXCEPTION '00022 postcondition: the auth-reader bodies are %', actual;
  END IF;
  IF (SELECT proconfig FROM pg_proc WHERE oid = f) IS DISTINCT FROM ARRAY['search_path=pg_catalog, pg_temp']
     OR (SELECT provolatile FROM pg_proc WHERE oid = f) <> 's'
     OR (SELECT prorettype FROM pg_proc WHERE oid = f) <> 'uuid'::regtype THEN
    RAISE EXCEPTION '00022 postcondition: the default-tenant reader configuration differs';
  END IF;
  -- EXECUTE: exactly the runtime and the owner (plus the definer itself), never PUBLIC or another helper.
  IF NOT has_function_privilege(rt, f, 'EXECUTE') OR NOT has_function_privilege(ow, f, 'EXECUTE')
     OR has_function_privilege('public', f, 'EXECUTE')
     OR has_function_privilege(pr, f, 'EXECUTE') OR has_function_privilege(cw, f, 'EXECUTE')
     OR has_function_privilege(pw, f, 'EXECUTE') OR has_function_privilege(rn, f, 'EXECUTE') THEN
    RAISE EXCEPTION '00022 postcondition: the default-tenant reader EXECUTE holders differ';
  END IF;
  IF has_schema_privilege(ar, 'app', 'CREATE') THEN
    RAISE EXCEPTION '00022 postcondition: the auth reader kept CREATE on schema app';
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 10 THEN
    RAISE EXCEPTION '00022 postcondition: the auth reader owns % bodies, expected 10', (SELECT count(*) FROM pg_proc WHERE proowner = ar);
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  ar text := current_database() || '_auth_reader';
BEGIN
  -- The reader, dropped by its owner (the 00018 down pattern); no object depends on it.
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.organization_id_by_slug(text);
  EXECUTE format('SET LOCAL ROLE %I', ow);
END
$do$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
DECLARE
  ar regrole := (current_database() || '_auth_reader')::regrole;
  actual text;
  n bigint;
BEGIN
  SELECT count(*) INTO n FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
    AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
    AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass);
  IF n <> 89 OR to_regprocedure('app.organization_id_by_slug(text)') IS NOT NULL THEN
    RAISE EXCEPTION '00022 down postcondition: % relations (expected 89) or the default-tenant reader remains', n;
  END IF;
  SELECT string_agg(format('%s:%s', p.proname, md5(p.prosrc)), ';' ORDER BY p.proname) INTO actual FROM pg_proc p
  WHERE p.oid IN ('app.hostname_state(text, text)'::regprocedure, 'app.signup_slug_taken(text)'::regprocedure) AND p.proowner = ar;
  IF actual IS DISTINCT FROM 'hostname_state:82138e2efb8c18c12eaef4f84793333b;signup_slug_taken:ec7ee37a2169822a22bf93aacebcd864'
     OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 9 THEN
    RAISE EXCEPTION '00022 down postcondition: the version-21 auth-reader bodies differ (%)', actual;
  END IF;
END
$$;
-- +goose StatementEnd
