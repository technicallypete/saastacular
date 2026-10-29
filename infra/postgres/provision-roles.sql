-- Portable role provisioning (G008; docs/foundation/DATABASE-ISOLATION.md, "Provisioning").
--
-- Creates, or verifies and completes, the database roles, memberships, database and schema privileges, default
-- privileges and search paths that infra/postgres/init/10-roles.sh creates on a fresh data volume, for PostgreSQL
-- without Docker init hooks and for volumes initialized before G008 (it then adds only the two helper roles and their
-- memberships), before G011 (it then adds only the empty, owner-owned audit schema), before G019 (it then adds only
-- the third helper role, <db>_claim_writer, and its membership) or before G062 (it then adds only the fourth helper
-- role, <db>_provision_writer, and its membership). Run it before migrating, as a
-- superuser or as a CREATEROLE role that owns the database, with psql:
--
--   DB_MIGRATOR_PASSWORD=... DB_RUNTIME_PASSWORD=... DB_RETENTION_PASSWORD=... \
--     psql -X -v ON_ERROR_STOP=1 -v db=<DATABASE_NAME> -d <DATABASE_NAME> -f infra/postgres/provision-roles.sql
--
-- (or stream it on stdin with -f -). Passwords come from the environment (\getenv), never from the command line, and
-- are used only to create a missing LOGIN role; an existing role's password is never changed. Everything runs in one
-- transaction after a read-only check: a missing password variable, an existing role with an unsafe attribute
-- (SUPERUSER, BYPASSRLS, CREATEROLE, CREATEDB, REPLICATION, LOGIN or INHERIT on a NOLOGIN role) or any unexpected
-- membership refuses with a non-zero exit before any change. A repeat run on an exact catalog changes nothing.
--
-- A CREATEROLE (non-superuser) administrator temporarily grants itself SET on <db>_owner (for CREATE SCHEMA ...
-- AUTHORIZATION and the owner's default privileges) and revokes that grant before it commits. PostgreSQL keeps its
-- automatic ADMIN-only grant (no INHERIT, no SET) on each role it created; that residual is the one difference from a
-- superuser run. ADMIN lets that administrator grant itself (or any role) membership in any of those roles again at any
-- time, as this script does for the owner, which reaches every row through the helpers: treat the provisioning
-- administrator as owner-equivalent, and retire or lock its login after use.

\set ON_ERROR_STOP on
\if :{?db}
\else
  \warn 'provision-roles: set the psql variable db (-v db=<DATABASE_NAME>)'
  SELECT 'provision-roles: db is required'::int;
\endif
\getenv migrator_pw DB_MIGRATOR_PASSWORD
\getenv runtime_pw DB_RUNTIME_PASSWORD
\getenv retention_pw DB_RETENTION_PASSWORD
\if :{?migrator_pw}
\else
  \warn 'provision-roles: DB_MIGRATOR_PASSWORD is not set; nothing was changed'
  SELECT 'provision-roles: DB_MIGRATOR_PASSWORD is required'::int;
\endif
\if :{?runtime_pw}
\else
  \warn 'provision-roles: DB_RUNTIME_PASSWORD is not set; nothing was changed'
  SELECT 'provision-roles: DB_RUNTIME_PASSWORD is required'::int;
\endif
\if :{?retention_pw}
\else
  \warn 'provision-roles: DB_RETENTION_PASSWORD is not set; nothing was changed'
  SELECT 'provision-roles: DB_RETENTION_PASSWORD is required'::int;
\endif

\set owner_role :db _owner
\set migrator_role :db _migrator
\set runtime_role :db _runtime
\set retention_role :db _retention
\set policy_reader_role :db _policy_reader
\set auth_reader_role :db _auth_reader
\set claim_writer_role :db _claim_writer
\set provision_writer_role :db _provision_writer

BEGIN;

-- Read-only check: every problem refuses before any change.
SELECT string_agg(problem, '; ' ORDER BY problem) AS problems FROM (
  SELECT 'connected to database ' || current_database() || ', not ' || :'db' AS problem WHERE current_database() <> :'db'
  UNION ALL
  SELECT 'the provisioning role must be a superuser or a CREATEROLE owner of the database'
  WHERE NOT (SELECT rolsuper OR (rolcreaterole AND pg_has_role(current_user,
                    (SELECT datdba FROM pg_database WHERE datname = current_database()), 'USAGE'))
             FROM pg_roles WHERE rolname = current_user)
  UNION ALL
  SELECT format('role %s has an unsafe attribute', r.rolname)
  FROM pg_roles r
  WHERE r.rolname IN (:'owner_role', :'migrator_role', :'runtime_role', :'retention_role', :'policy_reader_role', :'auth_reader_role',
                      :'claim_writer_role', :'provision_writer_role')
    AND (r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication
         OR (r.rolname IN (:'owner_role', :'policy_reader_role', :'auth_reader_role', :'claim_writer_role',
                           :'provision_writer_role') AND (r.rolcanlogin OR r.rolinherit))
         OR (r.rolname IN (:'migrator_role', :'runtime_role', :'retention_role') AND NOT r.rolcanlogin)
         OR (r.rolname IN (:'runtime_role', :'retention_role') AND r.rolinherit))
  UNION ALL
  -- Memberships among the application roles must be exactly the expected ones (each with INHERIT FALSE, SET TRUE,
  -- no ADMIN); any other membership of an application role, or any member of one that is not an ADMIN-only grant of
  -- a non-application role, refuses.
  SELECT format('unexpected membership: %s in %s (inherit=%s, set=%s, admin=%s)', m.member::regrole, m.roleid::regrole,
                m.inherit_option, m.set_option, m.admin_option)
  FROM pg_auth_members m
  WHERE (m.roleid::regrole::text IN (:'owner_role', :'migrator_role', :'runtime_role', :'retention_role', :'policy_reader_role', :'auth_reader_role',
                                     :'claim_writer_role', :'provision_writer_role')
         OR m.member::regrole::text IN (:'owner_role', :'migrator_role', :'runtime_role', :'retention_role', :'policy_reader_role', :'auth_reader_role',
                                        :'claim_writer_role', :'provision_writer_role'))
    AND NOT (
      ((m.member::regrole::text, m.roleid::regrole::text) IN ((:'migrator_role', :'owner_role'), (:'owner_role', :'policy_reader_role'),
                                                             (:'owner_role', :'auth_reader_role'), (:'owner_role', :'claim_writer_role'),
                                                             (:'owner_role', :'provision_writer_role'))
       AND NOT m.inherit_option AND m.set_option AND NOT m.admin_option)
      OR (m.member::regrole::text NOT IN (:'owner_role', :'migrator_role', :'runtime_role', :'retention_role', :'policy_reader_role', :'auth_reader_role',
                                          :'claim_writer_role', :'provision_writer_role')
          AND NOT m.inherit_option AND NOT m.set_option))
  UNION ALL
  SELECT 'schema app exists but is not owned by ' || :'owner_role'
  FROM pg_namespace n WHERE n.nspname = 'app' AND n.nspowner::regrole::text <> :'owner_role'
  UNION ALL
  SELECT 'schema audit exists but is not owned by ' || :'owner_role'
  FROM pg_namespace n WHERE n.nspname = 'audit' AND n.nspowner::regrole::text <> :'owner_role'
) p \gset
\if :{?problems}
  \warn 'provision-roles: refused, nothing was changed:'
  \warn :problems
  SELECT 'provision-roles: refused'::int;
\endif

-- Roles: create what is missing, with exactly the attributes of 10-roles.sh.
SELECT NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'owner_role') AS mk_owner,
       NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'migrator_role') AS mk_migrator,
       NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'runtime_role') AS mk_runtime,
       NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'retention_role') AS mk_retention,
       NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'policy_reader_role') AS mk_policy_reader,
       NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'auth_reader_role') AS mk_auth_reader,
       NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'claim_writer_role') AS mk_claim_writer,
       NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'provision_writer_role') AS mk_provision_writer,
       NOT (SELECT rolsuper FROM pg_roles WHERE rolname = current_user) AS admin_path \gset
\if :mk_owner
CREATE ROLE :"owner_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
\endif
\if :mk_migrator
CREATE ROLE :"migrator_role" LOGIN PASSWORD :'migrator_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE;
\endif
\if :mk_runtime
CREATE ROLE :"runtime_role" LOGIN PASSWORD :'runtime_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
\endif
\if :mk_retention
CREATE ROLE :"retention_role" LOGIN PASSWORD :'retention_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
\endif
\if :mk_policy_reader
CREATE ROLE :"policy_reader_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
\endif
\if :mk_auth_reader
CREATE ROLE :"auth_reader_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
\endif
\if :mk_claim_writer
CREATE ROLE :"claim_writer_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION;
\endif
\if :mk_provision_writer
CREATE ROLE :"provision_writer_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION;
\endif

-- Memberships: the migrator may SET ROLE to the owner; the owner may SET ROLE to the helpers; none inherits.
SELECT NOT EXISTS (SELECT 1 FROM pg_auth_members WHERE roleid = :'owner_role'::regrole AND member = :'migrator_role'::regrole
                     AND set_option) AS g_migrator,
       NOT EXISTS (SELECT 1 FROM pg_auth_members WHERE roleid = :'policy_reader_role'::regrole AND member = :'owner_role'::regrole
                     AND set_option) AS g_policy_reader,
       NOT EXISTS (SELECT 1 FROM pg_auth_members WHERE roleid = :'auth_reader_role'::regrole AND member = :'owner_role'::regrole
                     AND set_option) AS g_auth_reader,
       NOT EXISTS (SELECT 1 FROM pg_auth_members WHERE roleid = :'claim_writer_role'::regrole AND member = :'owner_role'::regrole
                     AND set_option) AS g_claim_writer,
       NOT EXISTS (SELECT 1 FROM pg_auth_members WHERE roleid = :'provision_writer_role'::regrole AND member = :'owner_role'::regrole
                     AND set_option) AS g_provision_writer \gset
\if :g_migrator
GRANT :"owner_role" TO :"migrator_role" WITH INHERIT FALSE, SET TRUE;
\endif
\if :g_policy_reader
GRANT :"policy_reader_role" TO :"owner_role" WITH INHERIT FALSE, SET TRUE;
\endif
\if :g_auth_reader
GRANT :"auth_reader_role" TO :"owner_role" WITH INHERIT FALSE, SET TRUE;
\endif
\if :g_claim_writer
GRANT :"claim_writer_role" TO :"owner_role" WITH INHERIT FALSE, SET TRUE;
\endif
\if :g_provision_writer
GRANT :"provision_writer_role" TO :"owner_role" WITH INHERIT FALSE, SET TRUE;
\endif

-- A non-superuser administrator needs SET on the owner for the schema and default privileges below (PostgreSQL 16+
-- grants an object's privileges only through an inheriting membership or SET ROLE); revoked again before commit.
\if :admin_path
GRANT :"owner_role" TO CURRENT_USER WITH INHERIT FALSE, SET TRUE;
\endif

-- Database: no PUBLIC access (CONNECT, TEMPORARY, CREATE); explicit CONNECT only.
REVOKE ALL ON DATABASE :"db" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"db" TO :"migrator_role", :"runtime_role", :"retention_role";

-- Schemas: nothing may be created in public; application objects live in app.
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
CREATE SCHEMA IF NOT EXISTS app AUTHORIZATION :"owner_role";
-- G011 (ADR-006): the empty audit schema, owned by the owner; migration 00008 creates every object in it.
CREATE SCHEMA IF NOT EXISTS audit AUTHORIZATION :"owner_role";
-- The schema grants and the owner's default privileges are made as the owner (the same grantor either path).
SET LOCAL ROLE :"owner_role";
REVOKE ALL ON SCHEMA app FROM PUBLIC;
GRANT USAGE ON SCHEMA app TO :"runtime_role", :"retention_role";
REVOKE ALL ON SCHEMA audit FROM PUBLIC;

-- Defaults for objects the owner creates: functions are not executable by PUBLIC; tables receive no grants until a
-- migration grants them explicitly.
ALTER DEFAULT PRIVILEGES FOR ROLE :"owner_role" REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE :"owner_role" IN SCHEMA app GRANT USAGE, SELECT ON SEQUENCES TO :"runtime_role";
RESET ROLE;

ALTER ROLE :"runtime_role" SET search_path = app;
ALTER ROLE :"retention_role" SET search_path = app;
ALTER ROLE :"migrator_role" SET search_path = app;

\if :admin_path
REVOKE :"owner_role" FROM CURRENT_USER;
\endif

COMMIT;
