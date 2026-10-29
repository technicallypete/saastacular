-- The one Organization policy flag of G007 (Spec R4): group_admins_edit_business_profiles.
--
-- The flag is a NOT NULL column on the Organization row, so there is no missing-row or malformed state: every existing
-- Organization reads true after this migration (a catalog-only fast default, no table rewrite) and every new
-- Organization receives true from the column default. True is the configured default, never a fallback after a read
-- failure. The runtime role may UPDATE the column (the guarded internal facade of core/internal/tenancy/access lets only
-- a current org_owner do so) and may not INSERT it, so a creator cannot pre-set the flag. No new relation, function,
-- trigger, index, row-level security or SECURITY DEFINER object; organizations_keys_immutable is unaffected.
--
-- Row-level security over the Organization row, including this column, is G008's (docs/foundation/AUTHORIZATION.md).
--
-- Down drops the column: it destroys only the flag values and restores the version-4 behaviour and grants;
-- identity, sessions, Organizations, Groups, role bindings, BusinessProfiles and associations are untouched. The runner
-- refuses down/reapply with APP_ENV=production. goose runs this file in one transaction.

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regrole(current_database() || '_runtime') IS NULL THEN
    RAISE EXCEPTION 'runtime role for database % is missing', current_database();
  END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE app.organizations ADD COLUMN group_admins_edit_business_profiles boolean NOT NULL DEFAULT true;

-- +goose StatementBegin
DO $$
DECLARE
  rt text := current_database() || '_runtime';
  expected_columns text := 'group_admins_edit_business_profiles:UPDATE;name:INSERT,UPDATE';
  actual text;
BEGIN
  EXECUTE format('GRANT UPDATE (group_admins_edit_business_profiles) ON app.organizations TO %I', rt);

  -- Postcondition (as 00003): on organizations the runtime role holds exactly SELECT at table level and exactly these
  -- column privileges, and no role other than the owner and the runtime role holds any privilege (no PUBLIC entry).
  SELECT string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) INTO actual
  FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
  WHERE c.oid = 'app.organizations'::regclass AND a.grantee = to_regrole(rt);
  IF actual IS DISTINCT FROM 'SELECT' THEN
    RAISE EXCEPTION 'organizations runtime table grants are %, expected SELECT', actual;
  END IF;
  SELECT string_agg(t.col || ':' || t.privs, ';' ORDER BY t.col) INTO actual
  FROM (
    SELECT at.attname AS col, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_attribute at CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE at.attrelid = 'app.organizations'::regclass AND at.attnum > 0 AND NOT at.attisdropped
      AND a.grantee = to_regrole(rt)
    GROUP BY at.attname
  ) t;
  IF actual IS DISTINCT FROM expected_columns THEN
    RAISE EXCEPTION 'organizations runtime column grants are %, expected %', actual, expected_columns;
  END IF;
  IF EXISTS (
    SELECT 1
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.oid = 'app.organizations'::regclass AND a.grantee NOT IN (c.relowner, to_regrole(rt))
  ) OR EXISTS (
    SELECT 1
    FROM pg_attribute at CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE at.attrelid = 'app.organizations'::regclass AND at.attnum > 0 AND NOT at.attisdropped
      AND a.grantee <> to_regrole(rt)
  ) THEN
    RAISE EXCEPTION 'organizations grants a privilege to a role other than its owner and the runtime role';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE app.organizations DROP COLUMN group_admins_edit_business_profiles;
