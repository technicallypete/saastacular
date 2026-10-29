#!/bin/sh
# Runs once, on first initialization of the data volume, as the image's init
# superuser. Creates separated roles; role names derive from DATABASE_NAME.
#   <db>_owner     NOLOGIN; owns schemas "app" and "audit" (G011) and every object in them
#   <db>_migrator  LOGIN; may SET ROLE <db>_owner (no inheritance) to migrate
#   <db>_runtime   LOGIN; application role: no superuser, no BYPASSRLS,
#                  no owner membership, so row-level security always applies
#   <db>_retention LOGIN; reserved for the audit-retention path
#   <db>_policy_reader, <db>_auth_reader  NOLOGIN, NOINHERIT helpers (G008): own the
#                  SELECT-only row-level security readers of migration 00006; their only
#                  member is <db>_owner (SET, no INHERIT), which hands them those functions
#   <db>_claim_writer  NOLOGIN, NOINHERIT helper (G019): owns the invitation claim bodies and the
#                  scoped user-lock helper of migration 00015; its only member is <db>_owner (SET, no INHERIT)
#   <db>_provision_writer  NOLOGIN, NOINHERIT helper (G062): owns the platform-signup provisioning and
#                  continuation bodies of migration 00018; its only member is <db>_owner (SET, no INHERIT)
# New tables are unreachable to runtime until explicitly granted. Existing volumes
# initialized before G008 add the helpers, before G011 the empty audit schema,
# before G019 the claim-writer helper, and before G062 only the fourth helper
# (provision-writer), with infra/postgres/provision-roles.sql.
set -eu
db="${DATABASE_NAME:?DATABASE_NAME is required}"
: "${DB_MIGRATOR_PASSWORD:?}" "${DB_RUNTIME_PASSWORD:?}" "${DB_RETENTION_PASSWORD:?}"

psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${db}" \
  -v db="${db}" \
  -v owner_role="${db}_owner" -v migrator_role="${db}_migrator" \
  -v runtime_role="${db}_runtime" -v retention_role="${db}_retention" \
  -v policy_reader_role="${db}_policy_reader" -v auth_reader_role="${db}_auth_reader" \
  -v claim_writer_role="${db}_claim_writer" -v provision_writer_role="${db}_provision_writer" \
  <<'SQL'
-- Passwords are read from the environment (\getenv), never passed on the
-- psql command line.
\getenv migrator_pw DB_MIGRATOR_PASSWORD
\getenv runtime_pw DB_RUNTIME_PASSWORD
\getenv retention_pw DB_RETENTION_PASSWORD
CREATE ROLE :"owner_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE :"migrator_role" LOGIN PASSWORD :'migrator_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE;
GRANT :"owner_role" TO :"migrator_role" WITH INHERIT FALSE, SET TRUE;
CREATE ROLE :"runtime_role" LOGIN PASSWORD :'runtime_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE :"retention_role" LOGIN PASSWORD :'retention_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
-- Row-level security reader helpers (G008, migration 00006): no login, no password,
-- no inheritance; the owner may only SET ROLE to them (to hand them their functions).
CREATE ROLE :"policy_reader_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE :"auth_reader_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
GRANT :"policy_reader_role" TO :"owner_role" WITH INHERIT FALSE, SET TRUE;
GRANT :"auth_reader_role" TO :"owner_role" WITH INHERIT FALSE, SET TRUE;
-- Invitation claim-writer helper (G019, migration 00015): same class as the readers.
CREATE ROLE :"claim_writer_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION;
GRANT :"claim_writer_role" TO :"owner_role" WITH INHERIT FALSE, SET TRUE;
-- Platform-signup provision-writer helper (G062, migration 00018): same class as the claim writer.
CREATE ROLE :"provision_writer_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION;
GRANT :"provision_writer_role" TO :"owner_role" WITH INHERIT FALSE, SET TRUE;

-- Database: no PUBLIC access (CONNECT, TEMPORARY, CREATE); explicit CONNECT only.
REVOKE ALL ON DATABASE :"db" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"db" TO :"migrator_role", :"runtime_role", :"retention_role";

-- Schemas: nothing may be created in public; application objects live in app.
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
CREATE SCHEMA app AUTHORIZATION :"owner_role";
REVOKE ALL ON SCHEMA app FROM PUBLIC;
GRANT USAGE ON SCHEMA app TO :"runtime_role", :"retention_role";
-- G011 (ADR-006): the empty audit schema; migration 00008 creates every object in it.
CREATE SCHEMA audit AUTHORIZATION :"owner_role";
REVOKE ALL ON SCHEMA audit FROM PUBLIC;

-- Defaults for objects the owner creates: functions are not executable by
-- PUBLIC; tables receive no grants until a migration grants them explicitly.
ALTER DEFAULT PRIVILEGES FOR ROLE :"owner_role" REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE :"owner_role" IN SCHEMA app GRANT USAGE, SELECT ON SEQUENCES TO :"runtime_role";

ALTER ROLE :"runtime_role" SET search_path = app;
ALTER ROLE :"retention_role" SET search_path = app;
ALTER ROLE :"migrator_role" SET search_path = app;
SQL
