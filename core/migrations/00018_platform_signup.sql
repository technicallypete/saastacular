-- Platform root signup: constrained tenant provisioning and automatic tenant continuation (G062 Plan r5 §3 with the
-- r5 ACCEPT pending-row policy clarification; docs/foundation/IDENTITY.md, TENANCY.md, DATABASE-ISOLATION.md, AUDIT.md).
--
-- Up, in one transaction:
--   - Precondition: owner, the FOUR NOLOGIN NOINHERIT helper roles (the fourth, <db>_provision_writer, is created by
--     infra/postgres/init/10-roles.sh on a fresh volume or infra/postgres/provision-roles.sql on an existing one; never
--     by a migration), the exact version-17 shapes, no 00018 object.
--   - app.login_transactions: signup_intent, signup_handoff, expected_user_id (FK users) with the three shape CHECKs;
--     return_path admits /signup/details (both closed sets: this CHECK and identity.returnPaths).
--   - app.signup_sessions (FORCE RLS): the pending root signup credential (30 minutes; SHA-256 digests only), the
--     create-intent result and one continuation generation at a time (initial reference R, tenant challenge C, root
--     nonce N, proof P; each a one-use SHA-256 digest, 120 s).
--   - Provision-writer authority: USAGE app/audit; EXECUTE on the four context parsers only; users SELECT (id, status) +
--     lock-only UPDATE (status); signup_sessions SELECT + UPDATE of exactly 14 columns; organizations INSERT (id, name,
--     slug) + SELECT (id, slug); role_bindings INSERT of the eight 00015 columns; audit.events INSERT of the nine
--     claim-writer columns. Policies TO the provision writer only. NOT granted: EXECUTE on app.hostname_state /
--     app.is_organization_member, any privilege on app.login_transactions, DELETE anywhere, audit SELECT.
--   - SECURITY DEFINER bodies (00018 allowance 4): provision writer app.signup_provision_organization,
--     app.signup_challenge_issue, app.signup_challenge_complete; auth reader app.auth_resolve_signup.
--   - Runtime/owner signup_sessions policies: own-row select, T1 insert, revoke, prove (nonce_consumed_at, proof_hash
--     only), EMPTY-shape expiry DELETE. No runtime/owner INSERT path on organizations or org_owner bindings is added.
--   - Audit (Plan r5 §3.6): no new action; reasons 70 -> 79, access.denied operations 65 -> 73; events_attribute admits
--     the signup auth.sign_in reasons and the signup refusals, each narrowed to its producer; events_insert_provision
--     (organization.created / membership.granted at trigger depth >= 1, actor = the context user).
--   - Postconditions.
--
-- Locks: brief ACCESS EXCLUSIVE locks for the column, policy, grant and constraint DDL on login_transactions,
-- organizations, role_bindings, users and audit.events; the audit constraint swap scans audit.events once and the
-- login_transactions CHECKs scan that (short-lived) table once. No data migration, no rewrite, no edit to an accepted
-- migration.
--
-- Down (development and test only) refuses while app.signup_sessions holds any row, app.login_transactions holds a
-- signup-shaped row or a /signup/details row, or audit.events holds a row with an 00018 reason or operation (inspected
-- with FORCE lifted inside the transaction; a refusal rolls the lift back), then restores version 17. The provision-writer
-- role is provisioning-owned and is not dropped.

-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow regrole;
  rt regrole;
  pr regrole;
  ar regrole;
  cw regrole;
  pw regrole;
  r record;
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regrole(db || '_runtime') IS NULL OR to_regrole(db || '_policy_reader') IS NULL OR to_regrole(db || '_auth_reader') IS NULL
     OR to_regrole(db || '_claim_writer') IS NULL OR to_regrole(db || '_provision_writer') IS NULL THEN
    RAISE EXCEPTION 'roles for database % are missing: run infra/postgres/provision-roles.sql (or a fresh 10-roles.sh) first', db;
  END IF;
  ow := current_user::regrole; rt := (db || '_runtime')::regrole; pr := (db || '_policy_reader')::regrole;
  ar := (db || '_auth_reader')::regrole; cw := (db || '_claim_writer')::regrole; pw := (db || '_provision_writer')::regrole;

  -- The 00015 helper checks, extended to the fourth helper.
  FOR r IN SELECT rolname, rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolinherit, rolcanlogin, rolreplication
           FROM pg_roles WHERE oid IN (pr, ar, cw, pw) LOOP
    IF r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication THEN
      RAISE EXCEPTION 'role % has an unsafe attribute: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
    IF r.rolcanlogin OR r.rolinherit THEN
      RAISE EXCEPTION 'helper role % must be NOLOGIN NOINHERIT: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
  END LOOP;
  IF (SELECT count(*) FROM pg_auth_members m WHERE m.roleid IN (pr, ar, cw, pw) AND m.member = ow
        AND NOT m.inherit_option AND m.set_option AND NOT m.admin_option) <> 4 THEN
    RAISE EXCEPTION 'the owner must be a member of the four helper roles WITH INHERIT FALSE, SET TRUE: run infra/postgres/provision-roles.sql';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member IN (rt, cw, pw))
     OR EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.roleid = pw AND m.member <> ow AND (m.inherit_option OR m.set_option)) THEN
    RAISE EXCEPTION 'the runtime, claim-writer and provision-writer roles must not be members of any role, and only the owner may use the provision writer';
  END IF;

  -- Version 17: the migration history ends at 17, the touched objects exist, no 00018 object exists, the provision
  -- writer owns nothing and holds nothing.
  IF (SELECT max(version_id) FROM app.goose_db_version) IS DISTINCT FROM 17 THEN
    RAISE EXCEPTION 'version-17 precondition: the current version is %', (SELECT max(version_id) FROM app.goose_db_version);
  END IF;
  IF to_regprocedure('app.organizations_group_override_guard()') IS NULL
     OR to_regprocedure('app.hostname_state(text, text)') IS NULL
     OR to_regprocedure('app.is_organization_member(uuid, uuid)') IS NULL
     OR NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'audit.events'::regclass AND conname = 'events_action_valid')
     OR NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'app.organizations'::regclass AND tgname = 'audit_capture_insert')
     OR NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'app.role_bindings'::regclass AND tgname = 'audit_capture_insert')
     OR to_regclass('app.signup_sessions') IS NOT NULL
     OR to_regprocedure('app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz)') IS NOT NULL
     OR to_regprocedure('app.signup_challenge_issue(bytea, text, uuid, bytea, bytea, timestamptz)') IS NOT NULL
     OR to_regprocedure('app.signup_challenge_complete(bytea, bytea, text, uuid, timestamptz)') IS NOT NULL
     OR to_regprocedure('app.auth_resolve_signup(bytea, timestamptz)') IS NOT NULL
     OR EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.login_transactions'::regclass AND NOT attisdropped
                  AND attname IN ('signup_intent', 'signup_handoff', 'expected_user_id'))
     OR EXISTS (SELECT 1 FROM pg_policy WHERE polname IN ('organizations_provision_insert', 'role_bindings_insert_owner',
                  'users_provision_select', 'users_provision_lock', 'events_insert_provision')) THEN
    RAISE EXCEPTION 'version-17 precondition: signup objects already exist or the version-17 shapes are missing';
  END IF;
  IF (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7 OR (SELECT count(*) FROM pg_proc WHERE proowner = pw) <> 0
     OR has_schema_privilege(pw, 'app', 'USAGE') OR has_schema_privilege(pw, 'audit', 'USAGE')
     OR EXISTS (SELECT 1 FROM pg_class WHERE oid IN ('app.organizations'::regclass, 'app.role_bindings'::regclass,
                  'app.users'::regclass, 'app.login_transactions'::regclass, 'audit.events'::regclass)
                  AND NOT (relrowsecurity AND relforcerowsecurity)) THEN
    RAISE EXCEPTION 'version-17 precondition: helper inventory, provision-writer privileges or FORCE differ';
  END IF;
END
$$;
-- +goose StatementEnd

-- The signup and pinned-handoff login transactions (Plan r5 §3.1): no provider-session column; the shapes keep the
-- three login kinds disjoint; /signup/details is the one new return path (the tenant completion keeps '/' and carries
-- its status in the closed callback outcome tenant_created).
ALTER TABLE app.login_transactions ADD COLUMN signup_intent boolean NOT NULL DEFAULT false,
  ADD COLUMN signup_handoff boolean NOT NULL DEFAULT false,
  ADD COLUMN expected_user_id uuid NULL
    CONSTRAINT login_transactions_expected_user_fkey REFERENCES app.users (id) ON DELETE CASCADE;
ALTER TABLE app.login_transactions
  ADD CONSTRAINT login_transactions_signup_intent_shape CHECK (NOT signup_intent
    OR (organization_id IS NULL AND NOT enrollment_intent)),
  ADD CONSTRAINT login_transactions_signup_handoff_shape CHECK (NOT signup_handoff
    OR (organization_id IS NOT NULL AND expected_user_id IS NOT NULL AND NOT signup_intent AND NOT enrollment_intent)),
  ADD CONSTRAINT login_transactions_expected_user_shape CHECK (signup_handoff OR expected_user_id IS NULL),
  DROP CONSTRAINT login_transactions_return_path_check,
  ADD CONSTRAINT login_transactions_return_path_check CHECK (return_path IN ('/', '/account', '/join/status', '/signup/details'));

-- The pending root signup credential, its create-intent result and the continuation generation (Plan r5 §3.2). Never
-- exchanged for a session; no issuer/subject/auth_time; every secret is stored as its SHA-256 digest only.
CREATE TABLE app.signup_sessions (
  id                      uuid        NOT NULL,
  user_id                 uuid        NOT NULL,
  origin                  text        NOT NULL,
  credential_hash         bytea       NOT NULL,
  csrf_token              bytea       NOT NULL,
  provider_session_id     text        NOT NULL,
  created_at              timestamptz NOT NULL DEFAULT now(),
  expires_at              timestamptz NOT NULL,
  revoked_at              timestamptz,
  revoke_reason           text,
  intent_hash             bytea,
  organization_id         uuid,
  provisioned_at          timestamptz,
  target_origin           text,
  continuation_generation integer     NOT NULL DEFAULT 0,
  continuation_expires_at timestamptz,
  ref_hash                bytea,
  ref_consumed_at         timestamptz,
  challenge_hash          bytea,
  nonce_hash              bytea,
  nonce_consumed_at       timestamptz,
  proof_hash              bytea,
  proof_consumed_at       timestamptz,
  completed_at            timestamptz,
  CONSTRAINT signup_sessions_pkey PRIMARY KEY (id),
  CONSTRAINT signup_sessions_user_fkey FOREIGN KEY (user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT signup_sessions_organization_fkey FOREIGN KEY (organization_id) REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT signup_sessions_credential_hash_key UNIQUE (credential_hash),
  CONSTRAINT signup_sessions_ref_hash_key UNIQUE (ref_hash),
  CONSTRAINT signup_sessions_challenge_hash_key UNIQUE (challenge_hash),
  CONSTRAINT signup_sessions_nonce_hash_key UNIQUE (nonce_hash),
  CONSTRAINT signup_sessions_proof_hash_key UNIQUE (proof_hash),
  CONSTRAINT signup_sessions_digest_lengths CHECK (octet_length(credential_hash) = 32 AND octet_length(csrf_token) = 32
    AND (intent_hash IS NULL OR octet_length(intent_hash) = 32) AND (ref_hash IS NULL OR octet_length(ref_hash) = 32)
    AND (challenge_hash IS NULL OR octet_length(challenge_hash) = 32) AND (nonce_hash IS NULL OR octet_length(nonce_hash) = 32)
    AND (proof_hash IS NULL OR octet_length(proof_hash) = 32)),
  CONSTRAINT signup_sessions_origin_length CHECK (char_length(origin) BETWEEN 1 AND 300),
  CONSTRAINT signup_sessions_target_origin_length CHECK (target_origin IS NULL OR char_length(target_origin) BETWEEN 1 AND 300),
  CONSTRAINT signup_sessions_provider_session_id_length CHECK (char_length(provider_session_id) BETWEEN 1 AND 255),
  CONSTRAINT signup_sessions_lifetime CHECK (expires_at > created_at AND expires_at <= created_at + interval '30 minutes'),
  CONSTRAINT signup_sessions_revocation_shape CHECK ((revoked_at IS NULL) = (revoke_reason IS NULL)),
  CONSTRAINT signup_sessions_provisioned_shape CHECK (num_nulls(intent_hash, organization_id, provisioned_at, target_origin) IN (0, 4)),
  CONSTRAINT signup_sessions_continuation_requires_provisioned CHECK (provisioned_at IS NOT NULL OR (
    continuation_generation = 0 AND num_nonnulls(continuation_expires_at, ref_hash, ref_consumed_at, challenge_hash, nonce_hash,
    nonce_consumed_at, proof_hash, proof_consumed_at, completed_at) = 0)),
  CONSTRAINT signup_sessions_generation_valid CHECK (continuation_generation >= 0
    AND (continuation_generation = 0) = (continuation_expires_at IS NULL)),
  CONSTRAINT signup_sessions_challenge_shape CHECK ((challenge_hash IS NULL) = (ref_consumed_at IS NULL)
    AND (nonce_hash IS NULL) = (ref_consumed_at IS NULL)),
  CONSTRAINT signup_sessions_nonce_shape CHECK (nonce_consumed_at IS NULL OR nonce_hash IS NOT NULL),
  CONSTRAINT signup_sessions_proof_shape CHECK ((proof_hash IS NULL OR nonce_consumed_at IS NOT NULL)
    AND (proof_consumed_at IS NULL OR proof_hash IS NOT NULL)),
  CONSTRAINT signup_sessions_completion_shape CHECK (completed_at IS NULL OR proof_consumed_at IS NOT NULL)
);
CREATE INDEX signup_sessions_user_idx ON app.signup_sessions (user_id, provider_session_id) WHERE revoked_at IS NULL;
ALTER TABLE app.signup_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.signup_sessions FORCE ROW LEVEL SECURITY;

-- The widened catalog constraints (Plan r5 §3.6): reasons +9 (70 -> 79), access.denied operations +8 (65 -> 73).
-- Actions (46) and resource types (11) unchanged. One validating scan of audit.events.
ALTER TABLE audit.events DROP CONSTRAINT events_reason_valid, DROP CONSTRAINT events_operation_valid,
  DROP CONSTRAINT events_action_valid;
ALTER TABLE audit.events
  ADD CONSTRAINT events_action_valid CHECK (action IN (
                    'organization.created', 'organization.renamed', 'organization.policy_changed',
                    'group.created', 'group.renamed', 'group.slug_changed',
                    'membership.granted', 'membership.role_changed', 'membership.removed',
                    'business_profile.created', 'business_profile.updated', 'business_profile.linked',
                    'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled',
                    'session.issued', 'session.revoked', 'session.provider_revocation_recorded',
                    'object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted',
                    'mcp.connection_established', 'mcp.context_changed',
                    'organization.slug_changed', 'domain.claimed', 'domain.state_changed', 'domain.checked',
                    'organization.branding_changed',
                    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
                    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
                    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked',
                    'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                    'object.deletion_pending', 'revocation.unaudited', 'access.override_used')),
  ADD CONSTRAINT events_reason_valid CHECK (reason IS NULL OR reason IN (
                    'group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled',
                    'logout', 'logout_all', 'provider_logout', 'disabled', 'superseded',
                    'signed_in', 'session_rotated', 'transaction_expired', 'login_canceled', 'code_rejected',
                    'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
                    'provider_error', 'provider_timeout', 'identity_namespace_mismatch', 'identity_unavailable',
                    'organization_not_permitted', 'group_not_selectable', 'target_not_found', 'target_kind_mismatch',
                    'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled', 'not_creator',
                    'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
                    'compensated', 'compensation_failed', 'not_compensated',
                    'provider_unavailable', 'tombstone_failed', 'tombstone_unknown',
                    'verified', 'active', 'suspended', 'removed', 'ok', 'mismatch', 'nxdomain', 'transient', 'tls_error',
                    'withdrawn', 'withdraw_error', 'withdraw_unknown', 'organization_not_member', 'host_changed',
                    'updated', 'reset', 'logo_set', 'logo_removed',
                    'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                    'recipient_proof_unavailable', 'recipient_ineligible', 'session_invalid',
                    'operator_verified_absent', 'operator_backfilled', 'history_unresolved',
                    'group_management_override_enabled', 'group_management_override_disabled',
                    'signup_authenticated', 'tenant_created', 'signup_origin_rejected', 'signup_user_mismatch',
                    'signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid', 'signup_context_mismatch')),
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug',
                    'branding.get', 'branding.update', 'branding.reset', 'branding.set_logo', 'branding.remove_logo',
                    'group.suggest_slug', 'group.create', 'group.list', 'group.onboarding.get',
                    'group.onboarding.complete_new', 'group.onboarding.complete_existing', 'business_profile.list',
                    'membership.list', 'membership.self', 'membership.change_role', 'membership.remove', 'membership.leave',
                    'ownership.promote', 'ownership.demote', 'ownership.transfer',
                    'invitation.list', 'invitation.mine', 'invitation.candidates', 'invitation.issue', 'invitation.revoke',
                    'invitation.get', 'invitation.accept', 'invitation.consume', 'retirement.eligibility',
                    'enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request', 'enrollment.admit',
                    'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update', 'enrollment.requests.list',
                    'enrollment.requests.decide', 'enrollment.requests.revoke', 'organization.set_group_override',
                    'signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                    'signup.complete', 'signup.abandon'));

-- The attribution and shape rules (the 00017 body otherwise unchanged; CREATE OR REPLACE keeps identity, owner, ACL and
-- trigger): auth.sign_in admits the T1 success signup_authenticated and the pinned-callback success tenant_created, and
-- the denials signup_origin_rejected and signup_user_mismatch; access.denied admits the signup refusals only on a signup
-- operation without an Organization, Group or resource, the reused host_changed also on signup.challenge /
-- signup.complete, and a signup operation carries only those reasons. The access.override_used authority reads (the
-- appender's Organization role and the Group-management override) move out of the shape CASE into a statement run only
-- for that action, with the same predicate, so the provision writer's captures never need those privileges.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit.events_attribute() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  d integer := pg_trigger_depth();
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  rid text := coalesce(current_setting('audit.request_id', true), '');
  tid text := coalesce(current_setting('audit.trace_id', true), '');
  trig boolean;
  lifecycle boolean;
  plain boolean;
  ovr boolean;
  ok boolean;
BEGIN
  NEW.id := gen_random_uuid();
  NEW.schema_version := 1;
  NEW.occurred_at := clock_timestamp();
  NEW.request_id := CASE WHEN rid ~ '^[A-Za-z0-9._-]{8,64}$' THEN rid END;
  NEW.trace_id := CASE WHEN tid ~ '^[0-9a-f]{32}$' THEN tid END;

  trig := NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed',
    'group.created', 'group.renamed', 'group.slug_changed', 'membership.granted', 'membership.role_changed',
    'membership.removed', 'business_profile.created', 'business_profile.updated', 'business_profile.linked',
    'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled', 'session.issued', 'session.revoked',
    'session.provider_revocation_recorded', 'object.upload_started', 'object.ready', 'object.delete_started',
    'object.deleted', 'mcp.connection_established', 'mcp.context_changed', 'organization.slug_changed', 'domain.claimed',
    'domain.state_changed', 'domain.checked', 'organization.branding_changed',
    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked');
  IF (d >= 2) IS DISTINCT FROM coalesce(trig, false) THEN
    RAISE EXCEPTION 'audit event source refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_source';
  END IF;

  -- Actor: never supplied. Lifecycle actions are system only inside the lifecycle producer (the target's IDENTITY
  -- shape); otherwise the context user, or unknown without one.
  lifecycle := NEW.action IN ('user.disabled', 'user.enabled')
    OR (NEW.action IN ('session.revoked', 'revocation.unaudited') AND NEW.reason IS NOT DISTINCT FROM 'disabled');
  IF lifecycle AND u IS NOT NULL AND u = NEW.subject_user_id THEN
    NEW.actor_kind := 'system'; NEW.actor_user_id := NULL;
  ELSIF u IS NOT NULL THEN
    NEW.actor_kind := 'user'; NEW.actor_user_id := u;
  ELSE
    NEW.actor_kind := 'unknown'; NEW.actor_user_id := NULL;
  END IF;

  -- A direct append names an Organization only when it is the established context (the context Organization, with a
  -- current Organization role now), and a Group only when that Group belongs to the Organization and is visible to the
  -- appender under row-level security. With a Group, the groups_select policy itself requires the context
  -- Organization and a non-NULL app.rls_org_role(), so the separate reader call is made only without a Group.
  IF d = 1 THEN
    IF NEW.organization_id IS NOT NULL AND (o IS NULL OR NEW.organization_id <> o) THEN
      RAISE EXCEPTION 'audit event scope refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_scope';
    END IF;
    IF NEW.group_id IS NOT NULL THEN
      IF NOT EXISTS (SELECT 1 FROM app.groups g WHERE g.id = NEW.group_id AND g.organization_id = NEW.organization_id) THEN
        RAISE EXCEPTION 'audit event group refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_group';
      END IF;
    ELSIF NEW.organization_id IS NOT NULL AND app.rls_org_role() IS NULL THEN
      RAISE EXCEPTION 'audit event scope refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_scope';
    END IF;
  END IF;

  -- The access.override_used authority (G060: the appender's current Organization role and the Organization's Group-
  -- management override) is read only for that action, in its own statement. Inside the CASE below, a generic plan
  -- (always used for a PL/pgSQL simple expression) checks the table and function privileges of every branch, taken or
  -- not, and the provision writer's captures hold neither app.organizations.group_management_override nor
  -- app.rls_org_role().
  ovr := false;
  IF NEW.action = 'access.override_used' THEN
    ovr := app.rls_org_role() IN ('org_owner', 'org_admin')
      AND coalesce((SELECT og.group_management_override FROM app.organizations og WHERE og.id = NEW.organization_id), false);
  END IF;

  plain := NEW.operation IS NULL AND NEW.ttl_seconds IS NULL;
  ok := CASE
    WHEN NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed',
                        'organization.slug_changed') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization'
      AND NEW.resource_id = NEW.organization_id
      AND CASE WHEN NEW.action = 'organization.policy_changed'
               THEN NEW.reason IN ('group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled',
                                   'group_management_override_enabled', 'group_management_override_disabled')
               ELSE NEW.reason IS NULL END
    WHEN NEW.action IN ('group.created', 'group.renamed', 'group.slug_changed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.group_id IS NOT NULL
      AND NEW.resource_type = 'group' AND NEW.resource_id = NEW.group_id
    WHEN NEW.action IN ('membership.granted', 'membership.role_changed', 'membership.removed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.organization_id IS NOT NULL
      AND NEW.resource_type = 'role_binding' AND NEW.subject_user_id IS NOT NULL
      AND CASE WHEN NEW.group_id IS NULL THEN NEW.role IN ('org_owner', 'org_admin', 'org_member')
               ELSE NEW.role IN ('group_owner', 'group_admin', 'group_member') END
      AND CASE WHEN NEW.action <> 'membership.role_changed' THEN NEW.previous_role IS NULL
               WHEN NEW.group_id IS NULL THEN NEW.previous_role IN ('org_owner', 'org_admin', 'org_member')
               ELSE NEW.previous_role IN ('group_owner', 'group_admin', 'group_member') END
    WHEN NEW.action IN ('business_profile.created', 'business_profile.updated', 'business_profile.linked',
                        'business_profile.unlinked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'business_profile'
      AND (NEW.group_id IS NULL) = (NEW.action IN ('business_profile.created', 'business_profile.updated'))
    WHEN NEW.action IN ('user.created', 'user.disabled', 'user.enabled') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'user' AND NEW.resource_id = NEW.subject_user_id
    WHEN NEW.action IN ('session.issued', 'session.revoked', 'session.provider_revocation_recorded') THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.subject_user_id IS NOT NULL AND NEW.resource_type = 'session'
      AND CASE NEW.action
            WHEN 'session.issued' THEN NEW.outcome = 'success' AND NEW.reason IS NULL
            WHEN 'session.revoked' THEN NEW.outcome = 'success'
              AND NEW.reason IN ('logout', 'logout_all', 'provider_logout', 'disabled', 'superseded')
            ELSE NEW.outcome IN ('success', 'failed') AND NEW.reason IS NULL END
    WHEN NEW.action IN ('object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'stored_object'
    WHEN NEW.action IN ('mcp.connection_established', 'mcp.context_changed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'mcp_connection' AND NEW.resource_id IS NOT NULL
      AND (NEW.action = 'mcp.context_changed' OR NEW.group_id IS NULL)
    WHEN NEW.action IN ('domain.claimed', 'domain.state_changed', 'domain.checked') THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL AND NEW.group_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization_domain' AND NEW.resource_id IS NOT NULL
      AND CASE NEW.action
            WHEN 'domain.claimed' THEN NEW.outcome = 'success' AND NEW.reason IS NULL
            WHEN 'domain.state_changed' THEN NEW.outcome = 'success' AND NEW.reason IN ('verified', 'active', 'suspended', 'removed')
            ELSE (NEW.outcome = 'success' AND NEW.reason IN ('ok', 'withdrawn'))
              OR (NEW.outcome = 'failed' AND NEW.reason IN ('mismatch', 'nxdomain', 'transient', 'tls_error', 'withdraw_error',
                                                           'withdraw_unknown')) END
    WHEN NEW.action = 'organization.branding_changed' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization'
      AND NEW.resource_id = NEW.organization_id AND NEW.reason IN ('updated', 'reset', 'logo_set', 'logo_removed')
    WHEN NEW.action IN ('invitation.issued', 'invitation.revoked', 'invitation.consumed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'invitation' AND NEW.resource_id IS NOT NULL
      AND CASE WHEN NEW.group_id IS NULL THEN NEW.role IN ('org_admin', 'org_member')
               ELSE NEW.role IN ('group_admin', 'group_member') END
      AND (NEW.action <> 'invitation.consumed' OR NEW.subject_user_id IS NOT NULL)
    WHEN NEW.action IN ('enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.group_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization' AND NEW.resource_id = NEW.organization_id
      AND CASE NEW.action
            WHEN 'enrollment.policy_changed' THEN NEW.reason IS NULL AND NEW.subject_user_id IS NULL
            WHEN 'enrollment.readiness_changed' THEN NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NULL
              AND NEW.reason IN ('operator_verified_absent', 'operator_backfilled', 'history_unresolved')
            ELSE NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NOT NULL AND NEW.reason = 'operator_backfilled' END
    WHEN NEW.action IN ('enrollment.requested', 'enrollment.decided', 'enrollment.revoked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'enrollment_request' AND NEW.resource_id IS NOT NULL AND NEW.actor_kind = 'user'
      AND (NEW.action <> 'enrollment.requested' OR NEW.actor_user_id = NEW.subject_user_id)
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated', 'signup_authenticated', 'tenant_created')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
            'organization_not_member', 'host_changed', 'recipient_proof_unavailable', 'signup_origin_rejected',
            'signup_user_mismatch'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
        'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch', 'recipient_ineligible',
        'session_invalid', 'host_changed', 'signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid',
        'signup_context_mismatch')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object', 'organization_domain')))
      AND (NEW.reason IS DISTINCT FROM 'resource_mismatch'
           OR (NEW.operation = 'mcp.request' AND NEW.organization_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.reason NOT IN ('enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                              'recipient_ineligible', 'session_invalid', 'host_changed')
           OR (NEW.operation IN ('enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request',
                                 'enrollment.admit', 'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update',
                                 'enrollment.requests.list', 'enrollment.requests.decide', 'enrollment.requests.revoke')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL)
           OR (NEW.reason = 'host_changed' AND NEW.operation IN ('signup.challenge', 'signup.complete')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.reason NOT IN ('signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid', 'signup_context_mismatch')
           OR (NEW.operation IN ('signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                                 'signup.complete', 'signup.abandon')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.operation NOT IN ('signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                                 'signup.complete', 'signup.abandon')
           OR (NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL
               AND NEW.reason IN ('signup_invalid', 'intent_used', 'slug_taken', 'continuation_invalid',
                                  'signup_context_mismatch', 'host_changed')))
      AND (NEW.operation NOT IN ('enrollment.start', 'enrollment.admit')
           OR (NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL
               AND CASE WHEN NEW.operation = 'enrollment.start'
                        THEN NEW.reason IN ('enrollment_closed', 'invitation_mismatch')
                        ELSE NEW.actor_user_id IS NOT NULL AND NEW.reason IN ('target_not_found', 'enrollment_closed',
                             'enrollment_restricted', 'approval_revoked', 'recipient_ineligible', 'session_invalid',
                             'host_changed') END))
    WHEN NEW.action = 'access.override_used' THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL
      AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL AND NEW.actor_kind = 'user'
      AND NEW.organization_id IS NOT NULL AND NEW.resource_id IS NOT NULL
      AND ovr
      AND CASE NEW.resource_type
            WHEN 'group' THEN NEW.group_id IS NOT NULL AND NEW.resource_id = NEW.group_id
              AND NEW.operation IN ('group.rename', 'group.change_slug', 'group.onboarding.complete_new',
                                    'group.onboarding.complete_existing', 'membership.change_role', 'membership.remove',
                                    'invitation.issue', 'invitation.revoke')
            WHEN 'business_profile' THEN
              (NEW.operation = 'business_profile.update' AND NEW.group_id IS NULL)
              OR (NEW.operation = 'group.onboarding.complete_existing' AND NEW.group_id IS NOT NULL)
            ELSE false END
    WHEN NEW.action = 'object.url_issued' THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND NEW.operation IS NULL AND NEW.ttl_seconds IS NOT NULL
      AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'stored_object'
    WHEN NEW.action IN ('object.upload_failed', 'object.deletion_pending') THEN
      NEW.outcome IN ('failed', 'unknown') AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.resource_type = 'stored_object'
      AND CASE WHEN NEW.action = 'object.upload_failed'
               THEN NEW.reason IN ('compensated', 'compensation_failed', 'not_compensated')
               ELSE NEW.reason IN ('provider_unavailable', 'tombstone_failed', 'tombstone_unknown') END
    WHEN NEW.action = 'revocation.unaudited' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL
      AND ((NEW.reason IN ('logout', 'provider_logout') AND NEW.resource_type = 'session')
        OR (NEW.reason IN ('logout_all', 'disabled') AND NEW.resource_type = 'user' AND NEW.resource_id = NEW.subject_user_id))
    ELSE false
  END;
  IF NOT coalesce(ok, false) THEN
    RAISE EXCEPTION 'audit event shape refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'audit_event_shape';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  ar text := current_database() || '_auth_reader';
  pw text := current_database() || '_provision_writer';
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  B text := '(SELECT app.ctx_login_binding())';
  -- r5 ACCEPT clarification: the provision writer's signup_sessions policies admit the rows of an ACTIVE user, pending
  -- (first create) and provisioned (retry, continuation) alike; the bodies keep the own-credential (IDENTITY) or exact
  -- reference/challenge + origin + Organization (LOGIN) checks.
  ACT text := 'EXISTS (SELECT 1 FROM app.users u WHERE u.id = signup_sessions.user_id AND u.status = ''active'')';
BEGIN
  -- Runtime/owner (Plan r5 §3.5): own-row SELECT; T1 INSERT of the issuing columns only in the IDENTITY shape; revocation
  -- pair and the prove pair (nonce_consumed_at, proof_hash) under own-row UPDATE policies; DELETE only of expired rows in
  -- the EMPTY shape (the login sweep). The bootstrap lookup by digest is app.auth_resolve_signup, never these policies.
  EXECUTE format('GRANT SELECT, DELETE ON app.signup_sessions TO %I', rt);
  EXECUTE format('GRANT INSERT (id, user_id, origin, credential_hash, csrf_token, provider_session_id, created_at, expires_at) '
              || 'ON app.signup_sessions TO %I', rt);
  EXECUTE format('GRANT UPDATE (revoked_at, revoke_reason, nonce_consumed_at, proof_hash) ON app.signup_sessions TO %I', rt);
  EXECUTE format('CREATE POLICY signup_sessions_select ON app.signup_sessions FOR SELECT TO %I, %I USING (user_id = %s)', rt, ow, U);
  EXECUTE format('CREATE POLICY signup_sessions_insert ON app.signup_sessions FOR INSERT TO %I, %I WITH CHECK ('
    || 'user_id = %s AND %s IS NULL AND revoked_at IS NULL AND provisioned_at IS NULL)', rt, ow, U, O);
  EXECUTE format('CREATE POLICY signup_sessions_revoke ON app.signup_sessions FOR UPDATE TO %I, %I USING (user_id = %s AND %s IS NULL) '
    || 'WITH CHECK (user_id = %s AND %s IS NULL)', rt, ow, U, O, U, O);
  EXECUTE format('CREATE POLICY signup_sessions_prove ON app.signup_sessions FOR UPDATE TO %I, %I USING (user_id = %s AND %s IS NULL) '
    || 'WITH CHECK (user_id = %s AND %s IS NULL)', rt, ow, U, O, U, O);
  EXECUTE format('CREATE POLICY signup_sessions_expire ON app.signup_sessions FOR DELETE TO %I, %I USING ('
    || 'expires_at <= now() AND %s IS NULL AND %s IS NULL)', rt, ow, U, B);
  -- Auth reader (the auth_resolve_signup owner).
  EXECUTE format('GRANT SELECT ON app.signup_sessions TO %I', ar);
  EXECUTE format('CREATE POLICY signup_sessions_auth_reader ON app.signup_sessions FOR SELECT TO %I USING (true)', ar);

  -- Provision-writer authority (Plan r5 §3.3/§3.5).
  EXECUTE format('GRANT USAGE ON SCHEMA app TO %I', pw);
  EXECUTE format('GRANT USAGE ON SCHEMA audit TO %I', pw);
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.ctx_user(), app.ctx_org(), app.ctx_group(), app.ctx_login_binding() TO %I', pw);
  EXECUTE format('GRANT SELECT (id, status), UPDATE (status) ON app.users TO %I', pw);
  EXECUTE format('CREATE POLICY users_provision_select ON app.users FOR SELECT TO %I USING (true)', pw);
  EXECUTE format('CREATE POLICY users_provision_lock ON app.users FOR UPDATE TO %I USING (id = app.ctx_user()) WITH CHECK (false)', pw);
  EXECUTE format('GRANT SELECT, UPDATE (intent_hash, organization_id, provisioned_at, target_origin, continuation_generation, '
              || 'continuation_expires_at, ref_hash, ref_consumed_at, challenge_hash, nonce_hash, nonce_consumed_at, proof_hash, '
              || 'proof_consumed_at, completed_at) ON app.signup_sessions TO %I', pw);
  EXECUTE format('CREATE POLICY signup_sessions_provision_select ON app.signup_sessions FOR SELECT TO %I USING (%s)', pw, ACT);
  EXECUTE format('CREATE POLICY signup_sessions_provision_update ON app.signup_sessions FOR UPDATE TO %I USING (%s) '
    || 'WITH CHECK (proof_hash IS NULL OR completed_at IS NOT NULL)', pw, ACT);
  EXECUTE format('GRANT INSERT (id, name, slug), SELECT (id, slug) ON app.organizations TO %I', pw);
  EXECUTE format('CREATE POLICY organizations_provision_insert ON app.organizations FOR INSERT TO %I WITH CHECK ('
    || 'pg_trigger_depth() = 0 AND app.ctx_user() IS NOT NULL AND app.ctx_org() IS NULL AND slug IS NOT NULL)', pw);
  EXECUTE format('GRANT INSERT (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) '
              || 'ON app.role_bindings TO %I', pw);
  EXECUTE format('CREATE POLICY role_bindings_insert_owner ON app.role_bindings FOR INSERT TO %I WITH CHECK ('
    || 'subject_user_id = app.ctx_user() AND app.ctx_org() IS NULL AND scope_type = ''organization'' AND scope_id = organization_id '
    || 'AND group_id IS NULL AND role = ''org_owner'' AND source = ''membership'')', pw);
  EXECUTE format('GRANT INSERT (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role, '
              || 'previous_role) ON audit.events TO %I', pw);
  EXECUTE format('CREATE POLICY events_insert_provision ON audit.events FOR INSERT TO %I WITH CHECK ('
    || 'pg_trigger_depth() >= 1 AND actor_kind = ''user'' AND actor_user_id = app.ctx_user() AND organization_id IS NOT NULL '
    || 'AND action IN (''organization.created'', ''membership.granted''))', pw);

  -- The auth reader's bootstrap resolver (the auth_resolve_enrollment pattern): the only way the runtime finds a signup
  -- row before ctx_user() exists; an unrevoked, unexpired row of an active user; no digest leaves the body.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $r$
CREATE FUNCTION app.auth_resolve_signup(p_hash bytea, p_at timestamptz)
  RETURNS TABLE (session_id uuid, user_id uuid, origin text, csrf_token bytea, expires_at timestamptz, provider_session_id text,
                 organization_id uuid, slug text, target_origin text, provisioned_at timestamptz,
                 continuation_generation integer, completed_at timestamptz)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
BEGIN
  RETURN QUERY
  SELECT ss.id, ss.user_id, ss.origin, ss.csrf_token, ss.expires_at, ss.provider_session_id, ss.organization_id, o.slug,
         ss.target_origin, ss.provisioned_at, ss.continuation_generation, ss.completed_at
  FROM app.signup_sessions ss JOIN app.users u ON u.id = ss.user_id
  LEFT JOIN app.organizations o ON o.id = ss.organization_id
  WHERE ss.credential_hash = p_hash AND ss.revoked_at IS NULL AND ss.expires_at > p_at AND u.status = 'active';
END $$
$r$;
  REVOKE ALL ON FUNCTION app.auth_resolve_signup(bytea, timestamptz) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.auth_resolve_signup(bytea, timestamptz) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);

  -- The three provision-writer bodies, created as the provision writer (the 00015 SET LOCAL ROLE pattern).
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pw);
  EXECUTE format('SET LOCAL ROLE %I', pw);
  -- Constrained provisioning (Plan r5 §3.4): IDENTITY shape; lock order users -> own signup row -> organizations ->
  -- role_bindings (G019 F-01). The Organization and its org_owner binding are inserted here only; the capture triggers
  -- append organization.created and membership.granted in the same statements under events_insert_provision.
  EXECUTE $r$
CREATE FUNCTION app.signup_provision_organization(p_hash bytea, p_name text, p_slug text, p_intent bytea, p_ref bytea,
                                                  p_target text, p_now timestamptz)
  RETURNS TABLE (outcome text, organization_id uuid, slug text, generation integer)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  st text;
  ss app.signup_sessions%ROWTYPE;
  oid_new uuid;
BEGIN
  IF u IS NULL OR app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL OR app.ctx_login_binding() IS NOT NULL THEN
    RAISE EXCEPTION 'signup context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'signup_context';
  END IF;
  IF p_hash IS NULL OR p_intent IS NULL OR p_ref IS NULL OR p_now IS NULL OR p_target IS NULL THEN
    RETURN QUERY SELECT 'signup_invalid'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END IF;
  -- (1) the creator, under a lock consistent with DisableUser (FOR UPDATE there, FOR SHARE here).
  PERFORM 1 FROM app.users us WHERE us.id = u FOR SHARE;
  SELECT us.status INTO st FROM app.users us WHERE us.id = u;
  IF st IS DISTINCT FROM 'active' THEN
    RETURN QUERY SELECT 'user_disabled'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END IF;
  -- (2) the caller's own live signup row, locked.
  SELECT s.* INTO ss FROM app.signup_sessions s
  WHERE s.credential_hash = p_hash AND s.user_id = u AND s.revoked_at IS NULL AND s.expires_at > p_now FOR UPDATE;
  IF NOT FOUND THEN
    RETURN QUERY SELECT 'signup_invalid'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END IF;
  -- (3) an already provisioned row: the same intent re-mints generation g+1 (every secret of g stops matching); a
  -- changed payload is intent_used.
  IF ss.provisioned_at IS NOT NULL THEN
    IF ss.intent_hash = p_intent THEN
      UPDATE app.signup_sessions s SET ref_hash = p_ref, continuation_expires_at = p_now + interval '120 seconds',
        ref_consumed_at = NULL, challenge_hash = NULL, nonce_hash = NULL, nonce_consumed_at = NULL, proof_hash = NULL,
        proof_consumed_at = NULL, completed_at = NULL, continuation_generation = s.continuation_generation + 1
      WHERE s.id = ss.id;
      RETURN QUERY SELECT 'already_created'::text, ss.organization_id, p_slug, ss.continuation_generation + 1;
      RETURN;
    END IF;
    RETURN QUERY SELECT 'intent_used'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END IF;
  -- (4) grammar and reserved labels are the 00011 CHECKs; (5) the Organization (defaults: invite_only,
  -- created_after_ledger, override false); the unique slug index decides one winner, a loser retains no write.
  oid_new := gen_random_uuid();
  BEGIN
    INSERT INTO app.organizations (id, name, slug) VALUES (oid_new, p_name, p_slug);
  EXCEPTION WHEN unique_violation THEN
    RETURN QUERY SELECT 'slug_taken'::text, NULL::uuid, NULL::text, NULL::integer;
    RETURN;
  END;
  -- (6) the creator's org_owner binding.
  INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
  VALUES (gen_random_uuid(), oid_new, u, 'organization', oid_new, NULL, 'org_owner', 'membership');
  -- (7) the intent, result, target and generation 1 with its initial reference.
  UPDATE app.signup_sessions s SET intent_hash = p_intent, organization_id = oid_new, provisioned_at = p_now,
    target_origin = p_target, continuation_generation = 1, continuation_expires_at = p_now + interval '120 seconds',
    ref_hash = p_ref
  WHERE s.id = ss.id;
  RETURN QUERY SELECT 'created'::text, oid_new, p_slug, 1;
END $$
$r$;
  -- Tenant challenge (Plan r5 §3.4 step 2): LOGIN shape; one statement consumes the initial reference R once and stores
  -- the challenge and nonce digests; p_org is the runtime-resolved request Organization. No reader is called.
  EXECUTE $r$
CREATE FUNCTION app.signup_challenge_issue(p_ref bytea, p_origin text, p_org uuid, p_challenge bytea, p_nonce bytea,
                                           p_now timestamptz)
  RETURNS TABLE (ok boolean)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  n integer;
BEGIN
  IF app.ctx_login_binding() IS NULL OR app.ctx_user() IS NOT NULL OR app.ctx_org() IS NOT NULL THEN
    RAISE EXCEPTION 'signup continuation context refused'
      USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'signup_continuation_context';
  END IF;
  UPDATE app.signup_sessions s SET ref_consumed_at = p_now, challenge_hash = p_challenge, nonce_hash = p_nonce
  WHERE s.ref_hash = p_ref AND s.ref_consumed_at IS NULL AND s.continuation_expires_at > p_now
    AND s.target_origin = p_origin AND s.organization_id = p_org AND s.revoked_at IS NULL
    AND EXISTS (SELECT 1 FROM app.users us WHERE us.id = s.user_id AND us.status = 'active');
  GET DIAGNOSTICS n = ROW_COUNT;
  RETURN QUERY SELECT n = 1;
END $$
$r$;
  -- Tenant completion (Plan r5 §3.4 step 4): LOGIN shape; matches only when the challenge AND the proof digests belong
  -- to the same row and generation; a wrong C or P changes nothing and spends nothing.
  EXECUTE $r$
CREATE FUNCTION app.signup_challenge_complete(p_challenge bytea, p_proof bytea, p_origin text, p_org uuid, p_now timestamptz)
  RETURNS TABLE (user_id uuid, organization_id uuid)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
BEGIN
  IF app.ctx_login_binding() IS NULL OR app.ctx_user() IS NOT NULL OR app.ctx_org() IS NOT NULL THEN
    RAISE EXCEPTION 'signup continuation context refused'
      USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'signup_continuation_context';
  END IF;
  RETURN QUERY
  UPDATE app.signup_sessions s SET proof_consumed_at = p_now, completed_at = p_now
  WHERE s.challenge_hash = p_challenge AND s.proof_hash = p_proof AND s.proof_consumed_at IS NULL
    AND s.nonce_consumed_at IS NOT NULL AND s.continuation_expires_at > p_now AND s.target_origin = p_origin
    AND s.organization_id = p_org AND s.revoked_at IS NULL
    AND EXISTS (SELECT 1 FROM app.users us WHERE us.id = s.user_id AND us.status = 'active')
  RETURNING s.user_id, s.organization_id;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz),
    app.signup_challenge_issue(bytea, text, uuid, bytea, bytea, timestamptz),
    app.signup_challenge_complete(bytea, bytea, text, uuid, timestamptz) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz), '
              || 'app.signup_challenge_issue(bytea, text, uuid, bytea, bytea, timestamptz), '
              || 'app.signup_challenge_complete(bytea, bytea, text, uuid, timestamptz) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pw);
END
$do$;
-- +goose StatementEnd

-- Postconditions (TestSignupCatalog repeats and extends them).
-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  rt regrole := (db || '_runtime')::regrole;
  ar regrole := (db || '_auth_reader')::regrole;
  cw regrole := (db || '_claim_writer')::regrole;
  pw regrole := (db || '_provision_writer')::regrole;
  cols text;
BEGIN
  IF (SELECT string_agg(p.proname, ',' ORDER BY p.proname) FROM pg_proc p WHERE p.proowner = pw)
       IS DISTINCT FROM 'signup_challenge_complete,signup_challenge_issue,signup_provision_organization'
     OR (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7
     OR (SELECT count(*) FROM pg_proc WHERE proowner = ar) <> 8
     OR EXISTS (SELECT 1 FROM pg_proc p WHERE p.proowner IN (pw, ar)
                  AND p.proname IN ('signup_challenge_complete', 'signup_challenge_issue', 'signup_provision_organization',
                                    'auth_resolve_signup')
                  AND (NOT p.prosecdef OR p.proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, pg_temp']
                       OR has_function_privilege('public', p.oid, 'EXECUTE')))
  THEN
    RAISE EXCEPTION '00018 postcondition: SECURITY DEFINER inventory differs';
  END IF;
  IF has_function_privilege(pw, 'app.hostname_state(text, text)', 'EXECUTE')
     OR has_function_privilege(pw, 'app.is_organization_member(uuid, uuid)', 'EXECUTE')
     OR has_table_privilege(pw, 'app.login_transactions', 'SELECT, INSERT, UPDATE, DELETE')
     OR has_table_privilege(pw, 'audit.events', 'SELECT')
     OR EXISTS (SELECT 1 FROM information_schema.table_privileges WHERE grantee = pw::text AND privilege_type IN ('DELETE', 'TRUNCATE'))
     OR has_schema_privilege(pw, 'app', 'CREATE')
     OR (has_table_privilege(rt, 'app.organizations', 'INSERT') AND EXISTS (SELECT 1 FROM pg_policy
          WHERE polrelid = 'app.organizations'::regclass AND polcmd = 'a' AND rt = ANY (polroles))) THEN
    RAISE EXCEPTION '00018 postcondition: provision-writer non-grants or runtime organization INSERT differ';
  END IF;
  SELECT string_agg(column_name, ',' ORDER BY column_name) INTO cols FROM information_schema.column_privileges
  WHERE grantee = pw::text AND table_schema = 'app' AND table_name = 'signup_sessions' AND privilege_type = 'UPDATE';
  IF cols IS DISTINCT FROM 'challenge_hash,completed_at,continuation_expires_at,continuation_generation,intent_hash,nonce_consumed_at,'
                        || 'nonce_hash,organization_id,proof_consumed_at,proof_hash,provisioned_at,ref_consumed_at,ref_hash,target_origin' THEN
    RAISE EXCEPTION '00018 postcondition: provision-writer signup_sessions UPDATE columns are %', cols;
  END IF;
  IF NOT (SELECT relrowsecurity AND relforcerowsecurity FROM pg_class WHERE oid = 'app.signup_sessions'::regclass)
     OR pg_get_constraintdef((SELECT oid FROM pg_constraint WHERE conrelid = 'app.login_transactions'::regclass
          AND conname = 'login_transactions_return_path_check'))
        IS DISTINCT FROM 'CHECK ((return_path = ANY (ARRAY[''/''::text, ''/account''::text, ''/join/status''::text, ''/signup/details''::text])))' THEN
    RAISE EXCEPTION '00018 postcondition: signup_sessions FORCE or the return-path CHECK differ';
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  ow text := current_user;
  rt text := current_database() || '_runtime';
  ar text := current_database() || '_auth_reader';
  pw text := current_database() || '_provision_writer';
  n_ses bigint;
  n_lt bigint;
  n_aud bigint;
BEGIN
  -- History is never discarded: refuse while any signup row, signup-shaped login transaction or audit row with an 00018
  -- reason or operation exists. FORCE is lifted only inside this transaction; a refusal rolls the lift back.
  ALTER TABLE app.signup_sessions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.login_transactions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_ses FROM app.signup_sessions;
  SELECT count(*) INTO n_lt FROM app.login_transactions WHERE signup_intent OR signup_handoff OR return_path = '/signup/details';
  SELECT count(*) INTO n_aud FROM audit.events
  WHERE reason IN ('signup_authenticated', 'tenant_created', 'signup_origin_rejected', 'signup_user_mismatch', 'signup_invalid',
                   'intent_used', 'slug_taken', 'continuation_invalid', 'signup_context_mismatch')
     OR operation IN ('signup.start', 'signup.status', 'signup.slug', 'signup.create', 'signup.challenge', 'signup.prove',
                      'signup.complete', 'signup.abandon');
  IF n_ses > 0 OR n_lt > 0 OR n_aud > 0 THEN
    RAISE EXCEPTION 'refusing to roll back 00018: % signup sessions / % login transactions / % audit rows (history is never discarded)',
      n_ses, n_lt, n_aud;
  END IF;
  ALTER TABLE app.login_transactions FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  EXECUTE format('SET LOCAL ROLE %I', pw);
  DROP FUNCTION app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz);
  DROP FUNCTION app.signup_challenge_issue(bytea, text, uuid, bytea, bytea, timestamptz);
  DROP FUNCTION app.signup_challenge_complete(bytea, bytea, text, uuid, timestamptz);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.auth_resolve_signup(bytea, timestamptz);
  EXECUTE format('SET LOCAL ROLE %I', ow);

  DROP TABLE app.signup_sessions;
  DROP POLICY events_insert_provision ON audit.events;
  DROP POLICY role_bindings_insert_owner ON app.role_bindings;
  DROP POLICY organizations_provision_insert ON app.organizations;
  DROP POLICY users_provision_select ON app.users;
  DROP POLICY users_provision_lock ON app.users;
  EXECUTE format('REVOKE INSERT (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role, '
              || 'previous_role) ON audit.events FROM %I', pw);
  EXECUTE format('REVOKE INSERT (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) '
              || 'ON app.role_bindings FROM %I', pw);
  EXECUTE format('REVOKE INSERT (id, name, slug), SELECT (id, slug) ON app.organizations FROM %I', pw);
  EXECUTE format('REVOKE SELECT (id, status), UPDATE (status) ON app.users FROM %I', pw);
  EXECUTE format('REVOKE EXECUTE ON FUNCTION app.ctx_user(), app.ctx_org(), app.ctx_group(), app.ctx_login_binding() FROM %I', pw);
  EXECUTE format('REVOKE USAGE ON SCHEMA audit FROM %I', pw);
  EXECUTE format('REVOKE USAGE ON SCHEMA app FROM %I', pw);

  ALTER TABLE app.login_transactions DROP CONSTRAINT login_transactions_return_path_check,
    ADD CONSTRAINT login_transactions_return_path_check CHECK (return_path IN ('/', '/account', '/join/status')),
    DROP CONSTRAINT login_transactions_expected_user_shape, DROP CONSTRAINT login_transactions_signup_handoff_shape,
    DROP CONSTRAINT login_transactions_signup_intent_shape,
    DROP COLUMN expected_user_id, DROP COLUMN signup_handoff, DROP COLUMN signup_intent;
END
$do$;
-- +goose StatementEnd

-- The version-17 attribution body (00017), restored.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit.events_attribute() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  d integer := pg_trigger_depth();
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  rid text := coalesce(current_setting('audit.request_id', true), '');
  tid text := coalesce(current_setting('audit.trace_id', true), '');
  trig boolean;
  lifecycle boolean;
  plain boolean;
  ok boolean;
BEGIN
  NEW.id := gen_random_uuid();
  NEW.schema_version := 1;
  NEW.occurred_at := clock_timestamp();
  NEW.request_id := CASE WHEN rid ~ '^[A-Za-z0-9._-]{8,64}$' THEN rid END;
  NEW.trace_id := CASE WHEN tid ~ '^[0-9a-f]{32}$' THEN tid END;

  trig := NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed',
    'group.created', 'group.renamed', 'group.slug_changed', 'membership.granted', 'membership.role_changed',
    'membership.removed', 'business_profile.created', 'business_profile.updated', 'business_profile.linked',
    'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled', 'session.issued', 'session.revoked',
    'session.provider_revocation_recorded', 'object.upload_started', 'object.ready', 'object.delete_started',
    'object.deleted', 'mcp.connection_established', 'mcp.context_changed', 'organization.slug_changed', 'domain.claimed',
    'domain.state_changed', 'domain.checked', 'organization.branding_changed',
    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked');
  IF (d >= 2) IS DISTINCT FROM coalesce(trig, false) THEN
    RAISE EXCEPTION 'audit event source refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_source';
  END IF;

  -- Actor: never supplied. Lifecycle actions are system only inside the lifecycle producer (the target's IDENTITY
  -- shape); otherwise the context user, or unknown without one.
  lifecycle := NEW.action IN ('user.disabled', 'user.enabled')
    OR (NEW.action IN ('session.revoked', 'revocation.unaudited') AND NEW.reason IS NOT DISTINCT FROM 'disabled');
  IF lifecycle AND u IS NOT NULL AND u = NEW.subject_user_id THEN
    NEW.actor_kind := 'system'; NEW.actor_user_id := NULL;
  ELSIF u IS NOT NULL THEN
    NEW.actor_kind := 'user'; NEW.actor_user_id := u;
  ELSE
    NEW.actor_kind := 'unknown'; NEW.actor_user_id := NULL;
  END IF;

  -- A direct append names an Organization only when it is the established context (the context Organization, with a
  -- current Organization role now), and a Group only when that Group belongs to the Organization and is visible to the
  -- appender under row-level security. With a Group, the groups_select policy itself requires the context
  -- Organization and a non-NULL app.rls_org_role(), so the separate reader call is made only without a Group.
  IF d = 1 THEN
    IF NEW.organization_id IS NOT NULL AND (o IS NULL OR NEW.organization_id <> o) THEN
      RAISE EXCEPTION 'audit event scope refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_scope';
    END IF;
    IF NEW.group_id IS NOT NULL THEN
      IF NOT EXISTS (SELECT 1 FROM app.groups g WHERE g.id = NEW.group_id AND g.organization_id = NEW.organization_id) THEN
        RAISE EXCEPTION 'audit event group refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_group';
      END IF;
    ELSIF NEW.organization_id IS NOT NULL AND app.rls_org_role() IS NULL THEN
      RAISE EXCEPTION 'audit event scope refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'audit_event_scope';
    END IF;
  END IF;

  plain := NEW.operation IS NULL AND NEW.ttl_seconds IS NULL;
  ok := CASE
    WHEN NEW.action IN ('organization.created', 'organization.renamed', 'organization.policy_changed',
                        'organization.slug_changed') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization'
      AND NEW.resource_id = NEW.organization_id
      AND CASE WHEN NEW.action = 'organization.policy_changed'
               THEN NEW.reason IN ('group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled',
                                   'group_management_override_enabled', 'group_management_override_disabled')
               ELSE NEW.reason IS NULL END
    WHEN NEW.action IN ('group.created', 'group.renamed', 'group.slug_changed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.group_id IS NOT NULL
      AND NEW.resource_type = 'group' AND NEW.resource_id = NEW.group_id
    WHEN NEW.action IN ('membership.granted', 'membership.role_changed', 'membership.removed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.organization_id IS NOT NULL
      AND NEW.resource_type = 'role_binding' AND NEW.subject_user_id IS NOT NULL
      AND CASE WHEN NEW.group_id IS NULL THEN NEW.role IN ('org_owner', 'org_admin', 'org_member')
               ELSE NEW.role IN ('group_owner', 'group_admin', 'group_member') END
      AND CASE WHEN NEW.action <> 'membership.role_changed' THEN NEW.previous_role IS NULL
               WHEN NEW.group_id IS NULL THEN NEW.previous_role IN ('org_owner', 'org_admin', 'org_member')
               ELSE NEW.previous_role IN ('group_owner', 'group_admin', 'group_member') END
    WHEN NEW.action IN ('business_profile.created', 'business_profile.updated', 'business_profile.linked',
                        'business_profile.unlinked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'business_profile'
      AND (NEW.group_id IS NULL) = (NEW.action IN ('business_profile.created', 'business_profile.updated'))
    WHEN NEW.action IN ('user.created', 'user.disabled', 'user.enabled') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'user' AND NEW.resource_id = NEW.subject_user_id
    WHEN NEW.action IN ('session.issued', 'session.revoked', 'session.provider_revocation_recorded') THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.subject_user_id IS NOT NULL AND NEW.resource_type = 'session'
      AND CASE NEW.action
            WHEN 'session.issued' THEN NEW.outcome = 'success' AND NEW.reason IS NULL
            WHEN 'session.revoked' THEN NEW.outcome = 'success'
              AND NEW.reason IN ('logout', 'logout_all', 'provider_logout', 'disabled', 'superseded')
            ELSE NEW.outcome IN ('success', 'failed') AND NEW.reason IS NULL END
    WHEN NEW.action IN ('object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'stored_object'
    WHEN NEW.action IN ('mcp.connection_established', 'mcp.context_changed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'mcp_connection' AND NEW.resource_id IS NOT NULL
      AND (NEW.action = 'mcp.context_changed' OR NEW.group_id IS NULL)
    WHEN NEW.action IN ('domain.claimed', 'domain.state_changed', 'domain.checked') THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL AND NEW.group_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization_domain' AND NEW.resource_id IS NOT NULL
      AND CASE NEW.action
            WHEN 'domain.claimed' THEN NEW.outcome = 'success' AND NEW.reason IS NULL
            WHEN 'domain.state_changed' THEN NEW.outcome = 'success' AND NEW.reason IN ('verified', 'active', 'suspended', 'removed')
            ELSE (NEW.outcome = 'success' AND NEW.reason IN ('ok', 'withdrawn'))
              OR (NEW.outcome = 'failed' AND NEW.reason IN ('mismatch', 'nxdomain', 'transient', 'tls_error', 'withdraw_error',
                                                           'withdraw_unknown')) END
    WHEN NEW.action = 'organization.branding_changed' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization'
      AND NEW.resource_id = NEW.organization_id AND NEW.reason IN ('updated', 'reset', 'logo_set', 'logo_removed')
    WHEN NEW.action IN ('invitation.issued', 'invitation.revoked', 'invitation.consumed') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'invitation' AND NEW.resource_id IS NOT NULL
      AND CASE WHEN NEW.group_id IS NULL THEN NEW.role IN ('org_admin', 'org_member')
               ELSE NEW.role IN ('group_admin', 'group_member') END
      AND (NEW.action <> 'invitation.consumed' OR NEW.subject_user_id IS NOT NULL)
    WHEN NEW.action IN ('enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.group_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization' AND NEW.resource_id = NEW.organization_id
      AND CASE NEW.action
            WHEN 'enrollment.policy_changed' THEN NEW.reason IS NULL AND NEW.subject_user_id IS NULL
            WHEN 'enrollment.readiness_changed' THEN NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NULL
              AND NEW.reason IN ('operator_verified_absent', 'operator_backfilled', 'history_unresolved')
            ELSE NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NOT NULL AND NEW.reason = 'operator_backfilled' END
    WHEN NEW.action IN ('enrollment.requested', 'enrollment.decided', 'enrollment.revoked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'enrollment_request' AND NEW.resource_id IS NOT NULL AND NEW.actor_kind = 'user'
      AND (NEW.action <> 'enrollment.requested' OR NEW.actor_user_id = NEW.subject_user_id)
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
            'organization_not_member', 'host_changed', 'recipient_proof_unavailable'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
        'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch', 'recipient_ineligible',
        'session_invalid', 'host_changed')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object', 'organization_domain')))
      AND (NEW.reason IS DISTINCT FROM 'resource_mismatch'
           OR (NEW.operation = 'mcp.request' AND NEW.organization_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.reason NOT IN ('enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                              'recipient_ineligible', 'session_invalid', 'host_changed')
           OR (NEW.operation IN ('enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request',
                                 'enrollment.admit', 'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update',
                                 'enrollment.requests.list', 'enrollment.requests.decide', 'enrollment.requests.revoke')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.operation NOT IN ('enrollment.start', 'enrollment.admit')
           OR (NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL
               AND CASE WHEN NEW.operation = 'enrollment.start'
                        THEN NEW.reason IN ('enrollment_closed', 'invitation_mismatch')
                        ELSE NEW.actor_user_id IS NOT NULL AND NEW.reason IN ('target_not_found', 'enrollment_closed',
                             'enrollment_restricted', 'approval_revoked', 'recipient_ineligible', 'session_invalid',
                             'host_changed') END))
    WHEN NEW.action = 'access.override_used' THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL
      AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL AND NEW.actor_kind = 'user'
      AND NEW.organization_id IS NOT NULL AND NEW.resource_id IS NOT NULL
      AND app.rls_org_role() IN ('org_owner', 'org_admin')
      AND coalesce((SELECT og.group_management_override FROM app.organizations og WHERE og.id = NEW.organization_id), false)
      AND CASE NEW.resource_type
            WHEN 'group' THEN NEW.group_id IS NOT NULL AND NEW.resource_id = NEW.group_id
              AND NEW.operation IN ('group.rename', 'group.change_slug', 'group.onboarding.complete_new',
                                    'group.onboarding.complete_existing', 'membership.change_role', 'membership.remove',
                                    'invitation.issue', 'invitation.revoke')
            WHEN 'business_profile' THEN
              (NEW.operation = 'business_profile.update' AND NEW.group_id IS NULL)
              OR (NEW.operation = 'group.onboarding.complete_existing' AND NEW.group_id IS NOT NULL)
            ELSE false END
    WHEN NEW.action = 'object.url_issued' THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND NEW.operation IS NULL AND NEW.ttl_seconds IS NOT NULL
      AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.subject_user_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'stored_object'
    WHEN NEW.action IN ('object.upload_failed', 'object.deletion_pending') THEN
      NEW.outcome IN ('failed', 'unknown') AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL AND NEW.resource_type = 'stored_object'
      AND CASE WHEN NEW.action = 'object.upload_failed'
               THEN NEW.reason IN ('compensated', 'compensation_failed', 'not_compensated')
               ELSE NEW.reason IN ('provider_unavailable', 'tombstone_failed', 'tombstone_unknown') END
    WHEN NEW.action = 'revocation.unaudited' THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.subject_user_id IS NOT NULL
      AND ((NEW.reason IN ('logout', 'provider_logout') AND NEW.resource_type = 'session')
        OR (NEW.reason IN ('logout_all', 'disabled') AND NEW.resource_type = 'user' AND NEW.resource_id = NEW.subject_user_id))
    ELSE false
  END;
  IF NOT coalesce(ok, false) THEN
    RAISE EXCEPTION 'audit event shape refused' USING ERRCODE = 'check_violation', CONSTRAINT = 'audit_event_shape';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- The version-17 catalog constraints (00017), restored.
ALTER TABLE audit.events DROP CONSTRAINT events_reason_valid, DROP CONSTRAINT events_operation_valid,
  DROP CONSTRAINT events_action_valid;
ALTER TABLE audit.events
  ADD CONSTRAINT events_action_valid CHECK (action IN (
                    'organization.created', 'organization.renamed', 'organization.policy_changed',
                    'group.created', 'group.renamed', 'group.slug_changed',
                    'membership.granted', 'membership.role_changed', 'membership.removed',
                    'business_profile.created', 'business_profile.updated', 'business_profile.linked',
                    'business_profile.unlinked', 'user.created', 'user.disabled', 'user.enabled',
                    'session.issued', 'session.revoked', 'session.provider_revocation_recorded',
                    'object.upload_started', 'object.ready', 'object.delete_started', 'object.deleted',
                    'mcp.connection_established', 'mcp.context_changed',
                    'organization.slug_changed', 'domain.claimed', 'domain.state_changed', 'domain.checked',
                    'organization.branding_changed',
                    'invitation.issued', 'invitation.revoked', 'invitation.consumed',
                    'enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
                    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked',
                    'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                    'object.deletion_pending', 'revocation.unaudited', 'access.override_used')),
  ADD CONSTRAINT events_reason_valid CHECK (reason IS NULL OR reason IN (
                    'group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled',
                    'logout', 'logout_all', 'provider_logout', 'disabled', 'superseded',
                    'signed_in', 'session_rotated', 'transaction_expired', 'login_canceled', 'code_rejected',
                    'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
                    'provider_error', 'provider_timeout', 'identity_namespace_mismatch', 'identity_unavailable',
                    'organization_not_permitted', 'group_not_selectable', 'target_not_found', 'target_kind_mismatch',
                    'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled', 'not_creator',
                    'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch',
                    'compensated', 'compensation_failed', 'not_compensated',
                    'provider_unavailable', 'tombstone_failed', 'tombstone_unknown',
                    'verified', 'active', 'suspended', 'removed', 'ok', 'mismatch', 'nxdomain', 'transient', 'tls_error',
                    'withdrawn', 'withdraw_error', 'withdraw_unknown', 'organization_not_member', 'host_changed',
                    'updated', 'reset', 'logo_set', 'logo_removed',
                    'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                    'recipient_proof_unavailable', 'recipient_ineligible', 'session_invalid',
                    'operator_verified_absent', 'operator_backfilled', 'history_unresolved',
                    'group_management_override_enabled', 'group_management_override_disabled')),
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug',
                    'branding.get', 'branding.update', 'branding.reset', 'branding.set_logo', 'branding.remove_logo',
                    'group.suggest_slug', 'group.create', 'group.list', 'group.onboarding.get',
                    'group.onboarding.complete_new', 'group.onboarding.complete_existing', 'business_profile.list',
                    'membership.list', 'membership.self', 'membership.change_role', 'membership.remove', 'membership.leave',
                    'ownership.promote', 'ownership.demote', 'ownership.transfer',
                    'invitation.list', 'invitation.mine', 'invitation.candidates', 'invitation.issue', 'invitation.revoke',
                    'invitation.get', 'invitation.accept', 'invitation.consume', 'retirement.eligibility',
                    'enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request', 'enrollment.admit',
                    'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update', 'enrollment.requests.list',
                    'enrollment.requests.decide', 'enrollment.requests.revoke', 'organization.set_group_override'));
