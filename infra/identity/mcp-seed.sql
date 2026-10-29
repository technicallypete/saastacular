-- MCP and hostname seed for the identity test overlay (infra/compose/identity.yml; G014 Plan r8 §10.5, G016 Plan r4
-- §1, D-2(a); test only, never production). The harness runs it once per identity run, after B-IDENTITY-IT and before
-- the browser run, as the PostgreSQL superuser with these psql variables (none is a secret):
--   issuer                 the configured G004 issuer (IDENTITY_WORKOS_ISSUER)
--   alice, bob, carol,     the emulator's user IDs of the seeded sign-in users (infra/identity/emulator-seed.yaml),
--   dave, frank, hostile   looked up through the emulator's API
--   org_a, org_b, org_c    the Organization UUIDs of infra/identity/mcp-hosts.tsv (org-a, org-b and org-c.localhost:543)
-- G016 D-2(a): a tenant-host sign-in completes only for a current Organization member, and a principal without an
-- account cannot become one there. So every browser sign-in user gets an account with its G004 principal (its first
-- sign-in maps to it) and an Organization membership of A, B and C; erin is deliberately absent (no account, refused).
-- The three Organizations carry the platform labels org-a, org-b and org-c (<label>.localhost under the overlay's
-- platform domain; G057 D-2: fixed synthetic test labels, never a production pattern) and the friendly slugs
-- alpha-agency, bravo-agency and charlie-agency (metadata only: alpha-agency.localhost routes nowhere, the slug-host
-- negative). Organization A also owns the routing custom host portal.acme.test (verified, TLS ready) and the
-- removed host old.acme.test, on which dave once created the connection the setup page lists with its own address
-- (D-4(a)). dave is in addition a Group member of research (A) and operations (B); design (A) is a Group he does not
-- belong to. The audit capture triggers record every insert as usual.
-- G017 (Plan r2 §6.4, E-BRD-1..6): carol is Organization A's owner (organization.manage_branding; her memberships of B
-- and C stay plain), and Organization A carries a branding row — a label with markup characters (rendered escaped) and
-- a complete passing palette in both appearances (the token-scope fixture's branded palette), no logo. Organizations
-- B and C stay unbranded (the platform-default fallback).
-- G018 (Plan r5 §3, §6.2 E-ONB-1..7; R0 addendum §B P-10): frank is Organization A's admin (the E-ONB-1 creator; alice
-- stays a plain member of A, the non-manager of G017's E-BRD-5 and G018's plain-member creator of E-ONB-4); bob is the
-- owner of research (A), which is onboarded with the Business Profile "Research Labs"; design (A) is onboarded with
-- "Design Studio"; Organization B holds one unattached profile ("Operations Partner"). dave additionally holds a Group
-- binding in planning (C) without any Organization C binding: the Group-role-only subject (its tenant sign-in on C is
-- refused under D-2(a)). The superuser inserts below the onboarding trigger's role condition, so every binding here is
-- explicit, as before.
-- G019 (Plan r1 §3, §6.3 E-MEM-1..7; current-target addendum P-11 (a)): every Organization and Group has an owner
-- (+2 org_owner: Bianca B, Cyrus C; +3 group_owner: Dana design, Bianca operations, Cyrus planning), and Organization A
-- has a second admin who is also an admin of research (Grace: +1 org_admin, +1 group_admin) for the promotion, demotion
-- and transfer scenarios. These four users have no principal and never sign in, so no G014–G018 sign-in user's
-- authority changes (carol/frank/bob rows kept; carol's B/C memberships stay plain). Browser scenarios that change
-- ownership use a Group the test itself creates, so the Light and Dark projects never share mutated seed state.
-- G058 (Plan r3 §6.2 E-ENR-1..5, §6.4 fixture eligibility): Organization B is open and Organization C request access;
-- Organization A stays invite-only and its membership history is closed (unresolved) through the operator procedure
-- app.enrollment_readiness_set (the owner's page offers open only once resolved). On C, carol is an owner and frank an
-- administrator (the policy owner and the request decider of the browser rows; both stay as before on A and B), alice
-- stays a plain member (the wrong-authority member), and dave and bob hold no Organization C binding: the request-mode
-- applicants of the two enrollment projects, asserted nonmembers of C below (dave keeps his planning Group binding, the
-- Group-role-only subject of G018). The first-time recipients erin and grace (infra/identity/emulator-seed.yaml) have no
-- account here.
\set ON_ERROR_STOP on
BEGIN;
INSERT INTO app.organizations (id, name, slug, platform_label) VALUES (:'org_a', 'Org A', 'alpha-agency', 'org-a'),
  (:'org_b', 'Org B', 'bravo-agency', 'org-b'), (:'org_c', 'Org C', 'charlie-agency', 'org-c');
INSERT INTO app.users (id, display_name) VALUES
  ('0f14d000-0000-4000-8000-00000000000d', 'Dave Example'),
  ('0f16e000-0000-4000-8000-0000000000a1', 'Alice Example'),
  ('0f16e000-0000-4000-8000-0000000000b1', 'Bob Example'),
  ('0f16e000-0000-4000-8000-0000000000c1', 'Carol Example'),
  ('0f16e000-0000-4000-8000-0000000000f1', 'Frank Example'),
  ('0f16e000-0000-4000-8000-0000000000e1', 'Hostile Example'),
  ('0f19e000-0000-4000-8000-0000000000b0', 'Bianca Owner'),
  ('0f19e000-0000-4000-8000-0000000000c0', 'Cyrus Owner'),
  ('0f19e000-0000-4000-8000-0000000000d0', 'Dana Owner'),
  ('0f19e000-0000-4000-8000-0000000000a3', 'Grace Admin');
INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES
  (:'issuer', :'dave', '0f14d000-0000-4000-8000-00000000000d'),
  (:'issuer', :'alice', '0f16e000-0000-4000-8000-0000000000a1'),
  (:'issuer', :'bob', '0f16e000-0000-4000-8000-0000000000b1'),
  (:'issuer', :'carol', '0f16e000-0000-4000-8000-0000000000c1'),
  (:'issuer', :'frank', '0f16e000-0000-4000-8000-0000000000f1'),
  (:'issuer', :'hostile', '0f16e000-0000-4000-8000-0000000000e1');
INSERT INTO app.groups (id, organization_id, name, slug) VALUES
  ('0f14a000-0000-4000-8000-0000000000a1', :'org_a', 'Research', 'research'),
  ('0f14a000-0000-4000-8000-0000000000a2', :'org_a', 'Design', 'design'),
  ('0f14b000-0000-4000-8000-0000000000b1', :'org_b', 'Operations', 'operations'),
  ('0f14c000-0000-4000-8000-0000000000c1', :'org_c', 'Planning', 'planning');
INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  (:'org_a', '0f14d000-0000-4000-8000-00000000000d', 'organization', :'org_a', NULL, 'org_member', 'membership'),
  (:'org_a', '0f14d000-0000-4000-8000-00000000000d', 'group', '0f14a000-0000-4000-8000-0000000000a1', '0f14a000-0000-4000-8000-0000000000a1', 'group_member', 'membership'),
  (:'org_b', '0f14d000-0000-4000-8000-00000000000d', 'organization', :'org_b', NULL, 'org_member', 'membership'),
  (:'org_b', '0f14d000-0000-4000-8000-00000000000d', 'group', '0f14b000-0000-4000-8000-0000000000b1', '0f14b000-0000-4000-8000-0000000000b1', 'group_member', 'membership'),
  (:'org_c', '0f14d000-0000-4000-8000-00000000000d', 'group', '0f14c000-0000-4000-8000-0000000000c1', '0f14c000-0000-4000-8000-0000000000c1', 'group_member', 'membership');
INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
  SELECT o.id, u.id, 'organization', o.id, NULL, 'org_member', 'membership'
  FROM (VALUES (:'org_a'::uuid), (:'org_b'::uuid), (:'org_c'::uuid)) AS o (id),
       (VALUES ('0f16e000-0000-4000-8000-0000000000a1'::uuid), ('0f16e000-0000-4000-8000-0000000000b1'::uuid),
               ('0f16e000-0000-4000-8000-0000000000c1'::uuid), ('0f16e000-0000-4000-8000-0000000000f1'::uuid),
               ('0f16e000-0000-4000-8000-0000000000e1'::uuid)) AS u (id)
  WHERE NOT (o.id = :'org_a'::uuid AND u.id IN ('0f16e000-0000-4000-8000-0000000000c1'::uuid, '0f16e000-0000-4000-8000-0000000000f1'::uuid))
    AND NOT (o.id = :'org_c'::uuid AND u.id IN ('0f16e000-0000-4000-8000-0000000000b1'::uuid, '0f16e000-0000-4000-8000-0000000000c1'::uuid,
                                                '0f16e000-0000-4000-8000-0000000000f1'::uuid));
INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  (:'org_a', '0f16e000-0000-4000-8000-0000000000c1', 'organization', :'org_a', NULL, 'org_owner', 'membership'),
  (:'org_a', '0f16e000-0000-4000-8000-0000000000f1', 'organization', :'org_a', NULL, 'org_admin', 'membership'),
  (:'org_a', '0f16e000-0000-4000-8000-0000000000b1', 'group', '0f14a000-0000-4000-8000-0000000000a1', '0f14a000-0000-4000-8000-0000000000a1', 'group_owner', 'membership'),
  (:'org_c', '0f16e000-0000-4000-8000-0000000000c1', 'organization', :'org_c', NULL, 'org_owner', 'membership'),
  (:'org_c', '0f16e000-0000-4000-8000-0000000000f1', 'organization', :'org_c', NULL, 'org_admin', 'membership');
INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  (:'org_b', '0f19e000-0000-4000-8000-0000000000b0', 'organization', :'org_b', NULL, 'org_owner', 'membership'),
  (:'org_b', '0f19e000-0000-4000-8000-0000000000b0', 'group', '0f14b000-0000-4000-8000-0000000000b1', '0f14b000-0000-4000-8000-0000000000b1', 'group_owner', 'membership'),
  (:'org_c', '0f19e000-0000-4000-8000-0000000000c0', 'organization', :'org_c', NULL, 'org_owner', 'membership'),
  (:'org_c', '0f19e000-0000-4000-8000-0000000000c0', 'group', '0f14c000-0000-4000-8000-0000000000c1', '0f14c000-0000-4000-8000-0000000000c1', 'group_owner', 'membership'),
  (:'org_a', '0f19e000-0000-4000-8000-0000000000d0', 'organization', :'org_a', NULL, 'org_member', 'membership'),
  (:'org_a', '0f19e000-0000-4000-8000-0000000000d0', 'group', '0f14a000-0000-4000-8000-0000000000a2', '0f14a000-0000-4000-8000-0000000000a2', 'group_owner', 'membership'),
  (:'org_a', '0f19e000-0000-4000-8000-0000000000a3', 'organization', :'org_a', NULL, 'org_admin', 'membership'),
  (:'org_a', '0f19e000-0000-4000-8000-0000000000a3', 'group', '0f14a000-0000-4000-8000-0000000000a1', '0f14a000-0000-4000-8000-0000000000a1', 'group_admin', 'membership');
INSERT INTO app.business_profiles (id, organization_id, display_name, legal_name, description, industry, website) VALUES
  ('0f18b000-0000-4000-8000-0000000000a1', :'org_a', 'Research Labs', 'Research Labs Ltd', 'The research team''s business context.', 'Research', 'https://research.example.com'),
  ('0f18b000-0000-4000-8000-0000000000a2', :'org_a', 'Design Studio', NULL, NULL, 'Design', NULL),
  ('0f18b000-0000-4000-8000-0000000000b1', :'org_b', 'Operations Partner', NULL, NULL, 'Logistics', NULL);
INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES
  (:'org_a', '0f14a000-0000-4000-8000-0000000000a1', '0f18b000-0000-4000-8000-0000000000a1'),
  (:'org_a', '0f14a000-0000-4000-8000-0000000000a2', '0f18b000-0000-4000-8000-0000000000a2');
-- G069 S23 / D-M3a (i) (Codex P0 disposition 2026-10-26; test input only): the palette passes every r3 pair of the kit's
-- rendered states. Light primary #0a5a4d replaces #0f6b5c (white on its kit 80 % hover was about 4.27), Light
-- destructive #9a2c25 replaces #a8322a (as text on its 20 % tint about 4.42), Dark destructive #ff9c8f replaces
-- #ff7a6b (as text on its 30 % tint about 4.35 / 4.04). The ring keeps #0f6b5c.
INSERT INTO app.organization_branding (organization_id, updated_by, display_label, light, dark) VALUES
  (:'org_a', '0f16e000-0000-4000-8000-0000000000c1', 'Acme <Works> & "Sons"',
   '{"background":"#f1f5f4","foreground":"#13201e","primary":"#0a5a4d","primary-foreground":"#ffffff","accent":"#7a4a00","muted-foreground":"#4a5d59","border":"#c9d6d3","popover":"#ffffff","popover-foreground":"#13201e","muted":"#e6eeec","secondary":"#3b5a7a","secondary-foreground":"#ffffff","destructive":"#9a2c25","input":"#4a5d59","ring":"#0f6b5c"}',
   '{"background":"#0b1514","foreground":"#e2ecea","primary":"#3fb8a3","primary-foreground":"#06201b","accent":"#f0c674","muted-foreground":"#93a8a4","border":"#24332f","popover":"#101d1b","popover-foreground":"#e2ecea","muted":"#0f1b19","secondary":"#8fb3d9","secondary-foreground":"#0b1514","destructive":"#ff9c8f","input":"#93a8a4","ring":"#f0c674"}');
INSERT INTO app.organization_domains (id, organization_id, hostname, generation, state, challenge_value, challenge_expires_at,
    verified_at, verified_until, last_check_at, last_check_outcome, tls_state, tls_checked_at, tls_not_after, tls_reference,
    removed_at, created_by) VALUES
  ('0f16d000-0000-4000-8000-0000000000a1', :'org_a', 'portal.acme.test', 1, 'active', repeat('a', 52), now() + interval '7 days',
    now(), now() + interval '30 days', now(), 'ok', 'ready', now(), now() + interval '60 days', 'overlay-fixture',
    NULL, '0f14d000-0000-4000-8000-00000000000d'),
  ('0f16d000-0000-4000-8000-0000000000a2', :'org_a', 'old.acme.test', 1, 'removed', repeat('b', 52), now() - interval '1 day',
    now() - interval '10 days', now() - interval '1 day', now() - interval '1 day', 'withdrawn', 'withdrawn', now() - interval '1 day', NULL, NULL,
    now() - interval '1 day', '0f14d000-0000-4000-8000-00000000000d');
-- The connection dave created on old.acme.test before its removal, with its own MCP session (never a browser credential:
-- the hashes are of random values nobody holds).
INSERT INTO app.sessions (id, user_id, kind, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at) VALUES
  ('0f16c000-0000-4000-8000-0000000000c2', '0f14d000-0000-4000-8000-00000000000d', 'mcp',
    sha256(convert_to(gen_random_uuid()::text, 'UTF8')), sha256(convert_to(gen_random_uuid()::text, 'UTF8')),
    now(), now() + interval '8 hours', now() + interval '8 hours');
INSERT INTO app.mcp_connections (id, session_id, user_id, organization_id, host, resource, state, pending_expires_at, client_id, activated_at) VALUES
  ('0f16c000-0000-4000-8000-0000000000c1', '0f16c000-0000-4000-8000-0000000000c2', '0f14d000-0000-4000-8000-00000000000d', :'org_a',
    'old.acme.test:543', 'https://old.acme.test:543/mcp/c/0f16c000-0000-4000-8000-0000000000c1', 'active', now() + interval '10 minutes',
    'seed-client', now());
-- G058: the enrollment modes (the superuser passes the owner guard; B was created after 00016, so its history is resolved
-- by construction), A's history closed through the audited operator procedure, and the applicants' eligibility.
UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = :'org_b';
UPDATE app.organizations SET enrollment_mode = 'request_access' WHERE id = :'org_c';
SELECT set_config('g058.org_a', :'org_a', true) AS a, set_config('g058.org_c', :'org_c', true) AS c \gset g058_
DO $$
DECLARE
  a uuid := current_setting('g058.org_a')::uuid;
  c uuid := current_setting('g058.org_c')::uuid;
BEGIN
  IF app.enrollment_readiness_set(a, NULL, 'mcp-seed.g058') IS DISTINCT FROM 'closed' THEN
    RAISE EXCEPTION 'G058 seed: Organization A history was not closed';
  END IF;
  IF app.is_organization_member('0f14d000-0000-4000-8000-00000000000d', c) OR app.is_organization_member('0f16e000-0000-4000-8000-0000000000b1', c) THEN
    RAISE EXCEPTION 'G058 seed: a request-mode applicant is already a member of Organization C';
  END IF;
  IF NOT (app.is_organization_member('0f16e000-0000-4000-8000-0000000000a1', c) AND app.is_organization_member('0f16e000-0000-4000-8000-0000000000c1', c)
          AND app.is_organization_member('0f16e000-0000-4000-8000-0000000000f1', c)) THEN
    RAISE EXCEPTION 'G058 seed: an Organization C manager or member binding is missing';
  END IF;
  IF (SELECT array_agg(enrollment_mode ORDER BY platform_label) FROM app.organizations WHERE platform_label IN ('org-a', 'org-b', 'org-c'))
       IS DISTINCT FROM ARRAY['invite_only', 'open', 'request_access'] THEN
    RAISE EXCEPTION 'G058 seed: unexpected enrollment modes';
  END IF;
END $$;
COMMIT;
