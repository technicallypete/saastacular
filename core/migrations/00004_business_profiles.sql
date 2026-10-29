-- Organization-owned BusinessProfile context, one-profile-per-Group same-Organization links and read-only access facts
-- derived from the single RoleBinding store (G006).
--
-- A BusinessProfile belongs to exactly one Organization for its whole life and is identified by a database-generated,
-- nonzero UUID; names and websites are never identifiers. A Group has at most one BusinessProfile association
-- (group_business_profiles_group_key), and one profile may be shared by several Groups of the same Organization. Both
-- association endpoints are bound to the Organization through composite foreign keys. No role, scope, grant, owner
-- column or writable mirror is added and role_bindings is not altered: business_profile_access_paths is a read-only
-- projection joining current Group membership bindings to the Group's single association.
--
-- Row-level security and policy are later obligations (G007/G008), and exactly-one-at-completion onboarding is G018's
-- (docs/foundation/TENANCY.md). No billing, onboarding, completion, policy or audit object.
--
-- Down drops every object of this migration and destroys all BusinessProfile and association data; identity, sessions,
-- Organizations, Groups and role bindings are untouched. The runner refuses down/reapply with APP_ENV=production. goose
-- runs this file in one transaction.

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

CREATE TABLE app.business_profiles (
  id              uuid        NOT NULL DEFAULT gen_random_uuid(),
  organization_id uuid        NOT NULL,
  display_name    text        NOT NULL,
  legal_name      text,
  description     text,
  industry        text,
  website         text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT business_profiles_pkey PRIMARY KEY (id),
  CONSTRAINT business_profiles_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT business_profiles_organization_id_key UNIQUE (organization_id, id),
  CONSTRAINT business_profiles_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT business_profiles_display_name_valid CHECK (char_length(display_name) BETWEEN 1 AND 255
    AND display_name = btrim(display_name) AND (display_name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'),
  CONSTRAINT business_profiles_legal_name_valid CHECK (legal_name IS NULL OR (char_length(legal_name) BETWEEN 1 AND 255
    AND legal_name = btrim(legal_name) AND (legal_name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]')),
  CONSTRAINT business_profiles_description_valid CHECK (description IS NULL OR (char_length(description) BETWEEN 1 AND 4000
    AND description = btrim(description) AND (description COLLATE "C") !~ '[\u0001-\u0008\u000b-\u001f\u007f-\u009f]')),
  CONSTRAINT business_profiles_industry_valid CHECK (industry IS NULL OR (char_length(industry) BETWEEN 1 AND 255
    AND industry = btrim(industry) AND (industry COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]')),
  CONSTRAINT business_profiles_website_valid CHECK (website IS NULL OR (char_length(website) BETWEEN 1 AND 2048
    AND (website COLLATE "C") ~ '^https?://[\u0021-\u005b\u005d-\u007e]+$'))
);

CREATE TABLE app.group_business_profiles (
  id                  uuid        NOT NULL DEFAULT gen_random_uuid(),
  organization_id     uuid        NOT NULL,
  group_id            uuid        NOT NULL,
  business_profile_id uuid        NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT group_business_profiles_pkey PRIMARY KEY (id),
  CONSTRAINT group_business_profiles_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT group_business_profiles_group_fkey FOREIGN KEY (organization_id, group_id)
    REFERENCES app.groups (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT group_business_profiles_business_profile_fkey FOREIGN KEY (organization_id, business_profile_id)
    REFERENCES app.business_profiles (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT group_business_profiles_group_key UNIQUE (organization_id, group_id),
  CONSTRAINT group_business_profiles_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000')
);
CREATE INDEX group_business_profiles_business_profile_idx
  ON app.group_business_profiles (organization_id, business_profile_id);

-- Key immutability, second line (the first is the runtime column grants below): any change of a key column raises
-- SQLSTATE 23000 with a named constraint, for every role including the owner.
-- +goose StatementBegin
CREATE FUNCTION app.business_profiles_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
  IF (NEW.id, NEW.organization_id) IS DISTINCT FROM (OLD.id, OLD.organization_id) THEN
    RAISE EXCEPTION 'business_profiles key columns are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'business_profiles_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION app.group_business_profiles_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
  IF (NEW.id, NEW.organization_id, NEW.group_id, NEW.business_profile_id)
     IS DISTINCT FROM (OLD.id, OLD.organization_id, OLD.group_id, OLD.business_profile_id) THEN
    RAISE EXCEPTION 'group_business_profiles key columns are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'group_business_profiles_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

CREATE TRIGGER business_profiles_keys_immutable BEFORE UPDATE ON app.business_profiles
  FOR EACH ROW EXECUTE FUNCTION app.business_profiles_keys_immutable();
CREATE TRIGGER group_business_profiles_keys_immutable BEFORE UPDATE ON app.group_business_profiles
  FOR EACH ROW EXECUTE FUNCTION app.group_business_profiles_keys_immutable();

-- Read-only access facts derived from the single authority store: one row per current Group membership binding of a
-- Group that has an association, carrying the raw Group role (a fact, not a capability). Organization authority stays
-- in organization_memberships. security_invoker applies the reader's privileges (and, from G008, row-level security)
-- on both base tables; no write privilege on the view is ever granted.
CREATE VIEW app.business_profile_access_paths WITH (security_invoker = true) AS
  SELECT rb.subject_user_id AS user_id, gbp.organization_id, gbp.business_profile_id, gbp.group_id, rb.role AS group_role
  FROM app.group_business_profiles gbp
  JOIN app.role_bindings rb ON rb.organization_id = gbp.organization_id AND rb.group_id = gbp.group_id
  WHERE rb.scope_type = 'group' AND rb.source = 'membership';

-- +goose StatementBegin
DO $$
DECLARE
  rt text := current_database() || '_runtime';
  rels text[] := ARRAY['business_profiles', 'group_business_profiles', 'business_profile_access_paths'];
  expected_tables text := 'business_profile_access_paths:SELECT;business_profiles:SELECT;group_business_profiles:DELETE,SELECT';
  expected_columns text := 'business_profiles.description:INSERT,UPDATE;business_profiles.display_name:INSERT,UPDATE;'
                        || 'business_profiles.industry:INSERT,UPDATE;business_profiles.legal_name:INSERT,UPDATE;'
                        || 'business_profiles.organization_id:INSERT;business_profiles.website:INSERT,UPDATE;'
                        || 'group_business_profiles.business_profile_id:INSERT;group_business_profiles.group_id:INSERT;'
                        || 'group_business_profiles.organization_id:INSERT';
  actual text;
BEGIN
  REVOKE ALL ON app.business_profiles, app.group_business_profiles, app.business_profile_access_paths FROM PUBLIC;
  EXECUTE format('GRANT SELECT ON app.business_profiles, app.group_business_profiles, '
              || 'app.business_profile_access_paths TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, display_name, legal_name, description, industry, website), '
              || 'UPDATE (display_name, legal_name, description, industry, website) ON app.business_profiles TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, group_id, business_profile_id), DELETE ON app.group_business_profiles TO %I', rt);

  -- Postcondition (as 00003): the runtime role holds exactly these table and column privileges on the three relations,
  -- and no role other than the owner and the runtime role holds any.
  SELECT string_agg(t.relname || ':' || t.privs, ';' ORDER BY t.relname) INTO actual
  FROM (
    SELECT c.relname, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
    FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
    WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY (rels) AND a.grantee = to_regrole(rt)
    GROUP BY c.relname
  ) t;
  IF actual IS DISTINCT FROM expected_tables THEN
    RAISE EXCEPTION 'business profile runtime table grants are %, expected %', actual, expected_tables;
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
    RAISE EXCEPTION 'business profile runtime column grants are %, expected %', actual, expected_columns;
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
    RAISE EXCEPTION 'a business profile relation grants a privilege to a role other than its owner and the runtime role';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
DROP VIEW app.business_profile_access_paths;
DROP TABLE app.group_business_profiles;
DROP TABLE app.business_profiles;
DROP FUNCTION app.group_business_profiles_keys_immutable();
DROP FUNCTION app.business_profiles_keys_immutable();
