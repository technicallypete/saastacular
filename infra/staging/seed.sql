-- Synthetic staging seed (docs/foundation/STAGING.md). Generic and idempotent: no product literal, and no principal,
-- session, domain, MCP connection, branding or enrollment change. infra/staging/remote/lib.sh runs it only when the
-- run found no PostgreSQL data volume before its first `up` (the first deploy, or reset --confirm), after
-- `migrate up`, as the cluster superuser over the PostgreSQL container's own socket (the privilege class of
-- infra/identity/mcp-seed.sql and infra/recovery), never over a network and never as the runtime role:
--   docker compose ... exec -T postgres psql --no-psqlrc --set=ON_ERROR_STOP=1 --username postgres --dbname <db> <seed.sql
-- A repeat run changes nothing: fixed ids, ON CONFLICT DO NOTHING, labels generated only on the first insert.
--
-- Contents (three synthetic Organizations, invite-only by default):
--   - the dataset marker (slug synthetic-staging-marker): its platform label is the dataset marker the lifecycle
--     scripts read (unchanged across stop, start and deploy; new after reset);
--   - two synthetic tenants of the same shape: a generated-label host under the wildcard certificate serves a tenant
--     page without new DNS records or certificates;
--   - per Organization one owner user without any identity_principals row (every Organization keeps an owner; nobody
--     can sign in as these users) and its org_owner binding.
-- Labels: 32 lowercase hex digits (128 random bits): 16 random digits from each of two gen_random_uuid() values, the
-- version and variant nibbles skipped as in migration 00021's backfill (pgcrypto is not installed, so there is no
-- gen_random_bytes). They satisfy the 00021 format and reserved-label CHECKs. The audit capture triggers record the
-- inserts as usual.
\set ON_ERROR_STOP on
BEGIN;

INSERT INTO app.organizations (id, name, slug, platform_label)
SELECT s.id, s.name, s.slug,
  (SELECT substr(g.a, 1, 12) || substr(g.a, 14, 3) || substr(g.a, 18, 1)
       || substr(g.b, 1, 12) || substr(g.b, 14, 3) || substr(g.b, 18, 1)
   FROM (SELECT replace(gen_random_uuid()::text, '-', '') AS a, replace(gen_random_uuid()::text, '-', '') AS b) AS g
   WHERE s.id IS NOT NULL) -- correlated: one fresh label per row
FROM (VALUES
  ('5ea70000-0000-4000-8000-000000000001'::uuid, 'Synthetic staging marker', 'synthetic-staging-marker'),
  ('5ea70000-0000-4000-8000-000000000002'::uuid, 'Synthetic staging tenant A', 'synthetic-staging-a'),
  ('5ea70000-0000-4000-8000-000000000003'::uuid, 'Synthetic staging tenant B', 'synthetic-staging-b')) AS s (id, name, slug)
ON CONFLICT (id) DO NOTHING;

INSERT INTO app.users (id, display_name) VALUES
  ('5ea70000-0000-4000-8000-0000000000a1', 'Synthetic staging owner M'),
  ('5ea70000-0000-4000-8000-0000000000a2', 'Synthetic staging owner A'),
  ('5ea70000-0000-4000-8000-0000000000a3', 'Synthetic staging owner B')
ON CONFLICT (id) DO NOTHING;

INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('5ea70000-0000-4000-8000-000000000001', '5ea70000-0000-4000-8000-0000000000a1', 'organization', '5ea70000-0000-4000-8000-000000000001', NULL, 'org_owner', 'membership'),
  ('5ea70000-0000-4000-8000-000000000002', '5ea70000-0000-4000-8000-0000000000a2', 'organization', '5ea70000-0000-4000-8000-000000000002', NULL, 'org_owner', 'membership'),
  ('5ea70000-0000-4000-8000-000000000003', '5ea70000-0000-4000-8000-0000000000a3', 'organization', '5ea70000-0000-4000-8000-000000000003', NULL, 'org_owner', 'membership')
ON CONFLICT ON CONSTRAINT role_bindings_subject_scope_key DO NOTHING;

-- Verification: the exact synthetic shape, or the whole seed rolls back.
DO $$
DECLARE
  orgs uuid[] := ARRAY['5ea70000-0000-4000-8000-000000000001', '5ea70000-0000-4000-8000-000000000002',
                       '5ea70000-0000-4000-8000-000000000003']::uuid[];
  owners uuid[] := ARRAY['5ea70000-0000-4000-8000-0000000000a1', '5ea70000-0000-4000-8000-0000000000a2',
                         '5ea70000-0000-4000-8000-0000000000a3']::uuid[];
BEGIN
  IF (SELECT count(*) FROM app.organizations
      WHERE id = ANY (orgs) AND platform_label ~ '^[0-9a-f]{32}$' AND enrollment_mode = 'invite_only') <> 3 THEN
    RAISE EXCEPTION 'staging seed: the three synthetic Organizations are not all present with a 32-digit hex label';
  END IF;
  IF (SELECT count(*) FROM app.role_bindings
      WHERE organization_id = ANY (orgs) AND scope_type = 'organization' AND role = 'org_owner'
        AND subject_user_id = ANY (owners)) <> 3
     OR EXISTS (SELECT 1 FROM app.role_bindings WHERE organization_id = ANY (orgs) AND NOT (subject_user_id = ANY (owners))) THEN
    RAISE EXCEPTION 'staging seed: the synthetic Organizations do not hold exactly their synthetic owners';
  END IF;
  IF EXISTS (SELECT 1 FROM app.identity_principals WHERE user_id = ANY (owners)) THEN
    RAISE EXCEPTION 'staging seed: a synthetic owner has an identity principal';
  END IF;
END
$$;

COMMIT;

-- Counts and names only (no label, no identifier beyond the synthetic slugs).
SELECT o.slug, (SELECT count(*) FROM app.role_bindings b WHERE b.organization_id = o.id AND b.role = 'org_owner') AS owners
FROM app.organizations o
WHERE o.id IN ('5ea70000-0000-4000-8000-000000000001', '5ea70000-0000-4000-8000-000000000002', '5ea70000-0000-4000-8000-000000000003')
ORDER BY o.slug;
