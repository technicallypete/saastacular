-- Global identity and application sessions (G004). Creates the single provider-namespace binding row, global
-- users, their exact provider principals, opaque application sessions and one-use login transactions.
--
-- These are global tables, not tenant-owned rows: no Organization, Group, membership, role or vendor-organization
-- column, and no row-level security (G008 classifies them). Session MFA assurance can only be 'unknown' until a
-- separately reviewed proof source and migration exist (ADR-003, M2). Runtime receives exactly the grants below:
-- no DELETE on the binding, users, principals or sessions, and no DDL.
--
-- Down drops all five tables and destroys identity data. The runner refuses down/reapply with APP_ENV=production.

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

CREATE TABLE app.identity_provider_binding (
  provider  text PRIMARY KEY CHECK (provider = 'workos'),
  issuer    text NOT NULL CHECK (char_length(issuer) BETWEEN 1 AND 2048),
  client_id text NOT NULL CHECK (char_length(client_id) BETWEEN 1 AND 255),
  bound_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE app.users (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  status       text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
  display_name text NULL CHECK (char_length(display_name) BETWEEN 1 AND 200),
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  disabled_at  timestamptz NULL,
  CHECK ((status = 'disabled') = (disabled_at IS NOT NULL))
);

CREATE TABLE app.identity_principals (
  issuer     text NOT NULL CHECK (char_length(issuer) BETWEEN 1 AND 2048),
  subject    text NOT NULL CHECK (char_length(subject) BETWEEN 1 AND 255 AND subject LIKE 'user\_%'),
  user_id    uuid NOT NULL UNIQUE REFERENCES app.users (id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (issuer, subject)
);

CREATE TABLE app.sessions (
  id                             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                        uuid NOT NULL REFERENCES app.users (id) ON DELETE RESTRICT,
  credential_hash                bytea NOT NULL UNIQUE CHECK (octet_length(credential_hash) = 32),
  csrf_token                     bytea NOT NULL CHECK (octet_length(csrf_token) = 32),
  provider_session_id            text NULL CHECK (char_length(provider_session_id) BETWEEN 1 AND 255),
  mfa_assurance                  text NOT NULL DEFAULT 'unknown' CHECK (mfa_assurance = 'unknown'),
  auth_time                      timestamptz NOT NULL,
  created_at                     timestamptz NOT NULL DEFAULT now(),
  renewed_at                     timestamptz NOT NULL DEFAULT now(),
  idle_expires_at                timestamptz NOT NULL,
  absolute_expires_at            timestamptz NOT NULL,
  revoked_at                     timestamptz NULL,
  revoke_reason                  text NULL
    CHECK (revoke_reason IN ('logout', 'logout_all', 'provider_logout', 'disabled', 'superseded')),
  provider_revocation            text NULL CHECK (provider_revocation IN ('confirmed', 'failed')),
  provider_logout_ref_hash       bytea NULL UNIQUE CHECK (octet_length(provider_logout_ref_hash) = 32),
  provider_logout_ref_expires_at timestamptz NULL,
  provider_logout_returned_at    timestamptz NULL,
  CHECK (idle_expires_at <= absolute_expires_at),
  CHECK ((revoked_at IS NULL) = (revoke_reason IS NULL)),
  CHECK ((provider_logout_ref_hash IS NULL) = (provider_logout_ref_expires_at IS NULL)),
  CHECK (provider_revocation IS NULL OR revoke_reason = 'provider_logout')
);
CREATE INDEX sessions_user_live_idx ON app.sessions (user_id) WHERE revoked_at IS NULL;

CREATE TABLE app.login_transactions (
  state_hash    bytea PRIMARY KEY CHECK (octet_length(state_hash) = 32),
  binding_hash  bytea NOT NULL CHECK (octet_length(binding_hash) = 32),
  pkce_verifier text NOT NULL CHECK (char_length(pkce_verifier) BETWEEN 43 AND 128),
  client_id     text NOT NULL CHECK (char_length(client_id) BETWEEN 1 AND 255),
  origin        text NOT NULL CHECK (char_length(origin) BETWEEN 1 AND 2048),
  return_path   text NOT NULL CHECK (return_path IN ('/', '/account')),
  max_age       integer NULL CHECK (max_age > 0),
  created_at    timestamptz NOT NULL DEFAULT now(),
  expires_at    timestamptz NOT NULL,
  CHECK (expires_at > created_at AND expires_at <= created_at + interval '10 minutes')
);

-- +goose StatementBegin
DO $$
DECLARE
  rt text := current_database() || '_runtime';
  expected text := 'identity_principals:INSERT,SELECT,UPDATE;identity_provider_binding:INSERT,SELECT;'
                || 'login_transactions:DELETE,INSERT,SELECT;sessions:INSERT,SELECT,UPDATE;users:INSERT,SELECT,UPDATE';
  actual text;
BEGIN
  EXECUTE format('GRANT SELECT, INSERT ON app.identity_provider_binding TO %I', rt);
  EXECUTE format('GRANT SELECT, INSERT, UPDATE ON app.users, app.identity_principals, app.sessions TO %I', rt);
  EXECUTE format('GRANT SELECT, INSERT, DELETE ON app.login_transactions TO %I', rt);

  -- Postcondition: the runtime role holds exactly these table privileges on the five tables, and no other role
  -- except the owner holds any.
  SELECT string_agg(t.relname || ':' || t.privs, ';' ORDER BY t.relname) INTO actual
  FROM (
    SELECT c.relname, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace
      AND c.relname IN ('identity_provider_binding', 'users', 'identity_principals', 'sessions', 'login_transactions')
      AND a.grantee = to_regrole(rt)
    GROUP BY c.relname
  ) t;
  IF actual IS DISTINCT FROM expected THEN
    RAISE EXCEPTION 'identity runtime grants are %, expected %', actual, expected;
  END IF;
  IF EXISTS (
    SELECT 1
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace
      AND c.relname IN ('identity_provider_binding', 'users', 'identity_principals', 'sessions', 'login_transactions')
      AND a.grantee NOT IN (c.relowner, to_regrole(rt))
  ) THEN
    RAISE EXCEPTION 'an identity table grants a privilege to a role other than its owner and the runtime role';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
DROP TABLE app.login_transactions;
DROP TABLE app.sessions;
DROP TABLE app.identity_principals;
DROP TABLE app.users;
DROP TABLE app.identity_provider_binding;
