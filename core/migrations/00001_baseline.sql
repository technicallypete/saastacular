-- Baseline: marks the start of the migration history. It creates no schema
-- objects and changes no privileges; application schema arrives with later
-- foundation stages. It only asserts the conditions every migration relies on.

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  -- Migrations run as the owner of schema app, reached through the migrator's
  -- SET ROLE path; any other identity fails here as well as in the runner.
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;

  -- Migration metadata is system infrastructure: the table and its identity
  -- sequence grant nothing to any role other than their owner.
  IF EXISTS (
    SELECT 1
    FROM pg_class c
    CROSS JOIN LATERAL aclexplode(coalesce(c.relacl, acldefault(
      CASE WHEN c.relkind = 'S' THEN 's'::"char" ELSE 'r'::"char" END, c.relowner))) a
    WHERE c.oid IN ('app.goose_db_version'::regclass,
                    pg_get_serial_sequence('app.goose_db_version', 'id')::regclass)
      AND a.grantee <> c.relowner
  ) THEN
    RAISE EXCEPTION 'migration metadata grants a privilege to a role other than its owner';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- The baseline created nothing, so there is nothing to remove.
SELECT 1;
