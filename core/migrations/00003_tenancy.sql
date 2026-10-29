-- Organization/Group tenancy model and the single RoleBinding authority store (G005).
--
-- Organization is the tenant boundary; a Group belongs to exactly one Organization and carries an editable,
-- Organization-unique slug. role_bindings is the only writable role store: organization_memberships and
-- group_memberships are read-only projections of it. The typed scope is enforced declaratively: an Organization binding
-- names its own Organization, a Group binding names a Group of the same Organization through a composite foreign key,
-- and one closed CHECK ties scope, shape, role and source together. Keys are database-generated, nonzero UUIDs; the
-- runtime role can neither supply nor change them (column grants first, BEFORE UPDATE triggers second).
--
-- Tenant-owned rows carry organization_id, but row-level security, policy and request context are later obligations
-- (G007/G008, docs/foundation/TENANCY.md). No BusinessProfile, onboarding, routing, branding, audit or owner column.
--
-- Down drops every object of this migration and destroys all tenancy data; identity and session tables and rows are
-- untouched. The runner refuses down/reapply with APP_ENV=production. goose runs this file in one transaction.

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

CREATE TABLE app.organizations (
  id         uuid        NOT NULL DEFAULT gen_random_uuid(),
  name       text        NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT organizations_pkey PRIMARY KEY (id),
  CONSTRAINT organizations_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT organizations_name_valid CHECK (char_length(name) BETWEEN 1 AND 200
    AND name = btrim(name) AND (name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]')
);

CREATE TABLE app.groups (
  id              uuid             NOT NULL DEFAULT gen_random_uuid(),
  organization_id uuid             NOT NULL,
  name            text             NOT NULL,
  slug            text COLLATE "C" NOT NULL,
  created_at      timestamptz      NOT NULL DEFAULT now(),
  CONSTRAINT groups_pkey PRIMARY KEY (id),
  CONSTRAINT groups_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT groups_organization_id_key UNIQUE (organization_id, id),
  CONSTRAINT groups_organization_slug_key UNIQUE (organization_id, slug),
  CONSTRAINT groups_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT groups_name_valid CHECK (char_length(name) BETWEEN 1 AND 200
    AND name = btrim(name) AND (name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'),
  CONSTRAINT groups_slug_format CHECK (char_length(slug) BETWEEN 1 AND 63
    AND octet_length(slug) = char_length(slug)
    AND slug ~ '^[0-9a-z]+(-[0-9a-z]+)*$'),
  CONSTRAINT groups_slug_reserved CHECK (slug NOT IN (
    'account', 'admin', 'api', 'auth', 'billing', 'blueprint', 'fonts', 'health',
    'healthz', 'internal', 'login', 'logout', 'mcp', 'members', 'message', 'messages',
    'new', 'onboarding', 'readyz', 'rpc', 'settings', 'sign-in', 'sse', 'system'))
);

CREATE TABLE app.role_bindings (
  id              uuid        NOT NULL DEFAULT gen_random_uuid(),
  organization_id uuid        NOT NULL,
  subject_user_id uuid        NOT NULL,
  scope_type      text        NOT NULL,
  scope_id        uuid        NOT NULL,
  group_id        uuid,
  role            text        NOT NULL,
  source          text        NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT role_bindings_pkey PRIMARY KEY (id),
  CONSTRAINT role_bindings_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT role_bindings_subject_fkey FOREIGN KEY (subject_user_id)
    REFERENCES app.users (id) ON DELETE RESTRICT,
  -- MATCH SIMPLE: unchecked when group_id IS NULL (Organization scope).
  CONSTRAINT role_bindings_group_fkey FOREIGN KEY (organization_id, group_id)
    REFERENCES app.groups (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT role_bindings_subject_scope_key UNIQUE (subject_user_id, scope_type, scope_id),
  CONSTRAINT role_bindings_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT role_bindings_subject_nonzero CHECK (subject_user_id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT role_bindings_scope_consistent CHECK (
       (scope_type = 'organization' AND scope_id = organization_id AND group_id IS NULL
        AND role IN ('org_owner', 'org_admin', 'org_member') AND source = 'membership')
    OR (scope_type = 'group' AND group_id IS NOT NULL AND scope_id = group_id
        AND role IN ('group_owner', 'group_admin', 'group_member') AND source = 'membership'))
);
CREATE INDEX role_bindings_scope_idx ON app.role_bindings (organization_id, scope_type, scope_id);
CREATE INDEX role_bindings_group_idx ON app.role_bindings (organization_id, group_id) WHERE group_id IS NOT NULL;

-- Key immutability, second line (the first is the runtime column grants below): any change of a key column raises
-- SQLSTATE 23000 with a named constraint, for every role including the owner.
-- +goose StatementBegin
CREATE FUNCTION app.organizations_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
  IF NEW.id IS DISTINCT FROM OLD.id THEN
    RAISE EXCEPTION 'organizations key columns are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'organizations_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION app.groups_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
  IF (NEW.id, NEW.organization_id) IS DISTINCT FROM (OLD.id, OLD.organization_id) THEN
    RAISE EXCEPTION 'groups key columns are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'groups_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION app.role_bindings_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
  IF (NEW.id, NEW.organization_id, NEW.subject_user_id, NEW.scope_type, NEW.scope_id, NEW.group_id, NEW.source)
     IS DISTINCT FROM
     (OLD.id, OLD.organization_id, OLD.subject_user_id, OLD.scope_type, OLD.scope_id, OLD.group_id, OLD.source) THEN
    RAISE EXCEPTION 'role_bindings key columns are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'role_bindings_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

CREATE TRIGGER organizations_keys_immutable BEFORE UPDATE ON app.organizations
  FOR EACH ROW EXECUTE FUNCTION app.organizations_keys_immutable();
CREATE TRIGGER groups_keys_immutable BEFORE UPDATE ON app.groups
  FOR EACH ROW EXECUTE FUNCTION app.groups_keys_immutable();
CREATE TRIGGER role_bindings_keys_immutable BEFORE UPDATE ON app.role_bindings
  FOR EACH ROW EXECUTE FUNCTION app.role_bindings_keys_immutable();

-- Read-only membership projections of the single authority store. security_invoker applies the reader's privileges
-- (and, from G008, row-level security) on role_bindings; no write privilege on either view is ever granted.
CREATE VIEW app.organization_memberships WITH (security_invoker = true) AS
  SELECT subject_user_id AS user_id, organization_id, role
  FROM app.role_bindings WHERE scope_type = 'organization' AND source = 'membership';
CREATE VIEW app.group_memberships WITH (security_invoker = true) AS
  SELECT subject_user_id AS user_id, organization_id, group_id, role
  FROM app.role_bindings WHERE scope_type = 'group' AND source = 'membership';

-- +goose StatementBegin
DO $$
DECLARE
  rt text := current_database() || '_runtime';
  rels text[] := ARRAY['organizations', 'groups', 'role_bindings', 'organization_memberships', 'group_memberships'];
  expected_tables text := 'group_memberships:SELECT;groups:SELECT;organization_memberships:SELECT;organizations:SELECT;'
                       || 'role_bindings:DELETE,SELECT';
  expected_columns text := 'groups.name:INSERT,UPDATE;groups.organization_id:INSERT;groups.slug:INSERT,UPDATE;'
                        || 'organizations.name:INSERT,UPDATE;role_bindings.group_id:INSERT;'
                        || 'role_bindings.organization_id:INSERT;role_bindings.role:INSERT,UPDATE;'
                        || 'role_bindings.scope_id:INSERT;role_bindings.scope_type:INSERT;role_bindings.source:INSERT;'
                        || 'role_bindings.subject_user_id:INSERT';
  actual text;
BEGIN
  REVOKE ALL ON app.organizations, app.groups, app.role_bindings,
                app.organization_memberships, app.group_memberships FROM PUBLIC;
  EXECUTE format('GRANT SELECT ON app.organizations, app.groups, app.role_bindings, '
              || 'app.organization_memberships, app.group_memberships TO %I', rt);
  EXECUTE format('GRANT INSERT (name), UPDATE (name) ON app.organizations TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, name, slug), UPDATE (name, slug) ON app.groups TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source), '
              || 'UPDATE (role), DELETE ON app.role_bindings TO %I', rt);

  -- Postcondition (the G004 pattern, extended to column ACLs): the runtime role holds exactly these table and column
  -- privileges on the five relations, and no role other than the owner and the runtime role holds any.
  SELECT string_agg(t.relname || ':' || t.privs, ';' ORDER BY t.relname) INTO actual
  FROM (
    SELECT c.relname, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY (rels) AND a.grantee = to_regrole(rt)
    GROUP BY c.relname
  ) t;
  IF actual IS DISTINCT FROM expected_tables THEN
    RAISE EXCEPTION 'tenancy runtime table grants are %, expected %', actual, expected_tables;
  END IF;
  SELECT string_agg(t.col || ':' || t.privs, ';' ORDER BY t.col) INTO actual
  FROM (
    SELECT c.relname || '.' || at.attname AS col, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c
    JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY (rels) AND a.grantee = to_regrole(rt)
    GROUP BY c.relname, at.attname
  ) t;
  IF actual IS DISTINCT FROM expected_columns THEN
    RAISE EXCEPTION 'tenancy runtime column grants are %, expected %', actual, expected_columns;
  END IF;
  IF EXISTS (
    SELECT 1
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY (rels)
      AND a.grantee NOT IN (c.relowner, to_regrole(rt))
  ) OR EXISTS (
    SELECT 1
    FROM pg_class c
    JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
    CROSS JOIN LATERAL aclexplode(at.attacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY (rels)
      AND a.grantee <> to_regrole(rt)
  ) THEN
    RAISE EXCEPTION 'a tenancy relation grants a privilege to a role other than its owner and the runtime role';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
DROP VIEW app.group_memberships;
DROP VIEW app.organization_memberships;
DROP TABLE app.role_bindings;
DROP TABLE app.groups;
DROP TABLE app.organizations;
DROP FUNCTION app.role_bindings_keys_immutable();
DROP FUNCTION app.groups_keys_immutable();
DROP FUNCTION app.organizations_keys_immutable();
