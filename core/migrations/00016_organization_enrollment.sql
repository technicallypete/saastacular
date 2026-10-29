-- Generic Organization enrollment (G058 Plan r3 §4 as bound by the selected-design addendum r0 §A.1/§A.3, its
-- correction 01 and the Codex selected-design review dispositions F1/F4/F5; docs/foundation/TENANCY.md, AUTHORIZATION.md,
-- DATABASE-ISOLATION.md, AUDIT.md).
--
-- SLICE 1a (admission subset). This unreleased file is extended in place by the later G058 slices (request, policy and
-- readiness); what 1a contains:
--   - Precondition: owner, the three NOLOGIN NOINHERIT helper roles, the exact version-15 shapes, no 00016 object.
--   - app.organizations: enrollment_mode (invite_only default; every existing Organization stays closed) and the two
--     readiness columns of addendum §A.1 with their three constraints, added without a default (existing rows stay
--     unresolved) and defaulted afterwards (§A.3 order); the policy-reader and claim-writer column grants and the three
--     helper policies (organizations_policy_reader, organizations_claim_select, organizations_claim_lock WITH CHECK
--     false). Not in 1a: the runtime mode UPDATE grant/policy, organizations_enrollment_guard, the capture triggers and
--     the readiness setter/backfill functions (policy/readiness slice).
--   - app.login_transactions: enrollment_intent, invitation_id, request_on_arrival; return_path admits /join/status
--     (Plan r0 §4.1). The runtime's table-level SELECT/INSERT/DELETE already cover the new columns.
--   - app.enrollment_sessions (13 columns) and app.enrollment_restrictions (12 columns; the ledger), both FORCE RLS, with
--     their runtime/owner, auth-reader, claim-writer and owner-insert policies. No session role may INSERT into or
--     DELETE from the ledger; its only INSERT path is the owner policy usable inside a trigger.
--   - Six SECURITY DEFINER functions (the full 00016 allowance is 8; enrollment_request and
--     enrollment_restrict_on_decision arrive with the request slice): policy reader rls_enrollment_mode; claim writer
--     rls_invitation_pending (Codex F1: claim-writer owned, reading through its existing 00015 invitation SELECT),
--     enrollment_lock_organization (EXECUTE owner only) and enrollment_admit (kinds lift and open; request refused
--     until the request slice); auth reader auth_resolve_enrollment; owner trigger body
--     enrollment_restrict_on_departure on app.role_bindings. Owned-function counts after 1a: claim writer 3 -> 6
--     (final 7), policy reader +1 (final +1), auth reader +1 (final +1), owner DEFINER trigger bodies +1 (final +2).
--   - D-15(i): the runtime may EXECUTE app.invitation_claim_organization.
--   - audit (Codex F5 and round-2 audit decision): reasons 58 -> 65 and access.denied operations 53 -> 64 in the catalog
--     constraints (operations incl. enrollment.start, round-3 decision 1); events_attribute admits
--     recipient_proof_unavailable on auth.sign_in denied and the enrollment refusal
--     reasons (with the reused host_changed) on access.denied, each narrowed to its reviewed producer (the 00015 body
--     otherwise unchanged). Trigger actions and resource types are unchanged in 1a.
--
-- SLICE 1b (policy and readiness; Plan r3 §1.5/§4.4/§4.5, addendum §A.2/§A.5, correction 01 §C.1/§C.2/§C.5):
--   - the runtime UPDATE (enrollment_mode) grant with organizations_enrollment_update (MGR) and the owner-owned invoker
--     guard organizations_enrollment_guard (a session-role mode change only by the Organization owner; readiness only
--     through the superuser channel inside the audited procedure, marker enrollment_readiness_set | replay-restrict);
--   - audit captures audit_capture_enrollment (enrollment.policy_changed / enrollment.readiness_changed) and
--     audit_capture_backfill (enrollment.restriction_backfilled), actions 39 -> 42 (trigger-owned 33 -> 36) and
--     reasons 65 -> 68;
--   - the owner-owned SECURITY INVOKER operator functions app.enrollment_readiness_set and
--     app.enrollment_history_backfill (EXECUTE held by no role; only a superuser session completes them; a backfill
--     closes readiness). The SECURITY DEFINER inventory is unchanged.
--
-- REQUEST SLICE (Plan r3 §1.4/§4.2-§4.5, r3 corrections 1-2):
--   - app.enrollment_requests (FORCE RLS; one pending request per subject and Organization) with the runtime/owner
--     select and decide policies (no runtime INSERT, no privilege on authority_seq) and the claim writer's insert,
--     select and consumption policies; the ledger's source_request_id foreign key;
--   - the claim writer's app.enrollment_request (the 7th claim-writer body) and enrollment_admit kind 'request';
--   - the owner-owned DEFINER decision body app.enrollment_restrict_on_decision (denied/revoked -> a ledger row), the
--     owner-owned invoker state machine app.enrollment_requests_immutable (authority_seq derived under the Organization
--     lock), the capture audit.capture_enrollment_requests and the claim writer's events_insert_enrollment policy;
--   - actions 42 -> 45 (trigger-owned 36 -> 39: enrollment.requested, enrollment.decided, enrollment.revoked) and resource
--     types 10 -> 11 (enrollment_request). The SECURITY DEFINER inventory reaches the 00016 allowance of eight.
--
-- Locks: brief ACCESS EXCLUSIVE locks for the column, policy, grant and trigger DDL on organizations, login_transactions,
-- role_bindings and audit.events; the audit constraint swap scans audit.events once. ADD COLUMN ... DEFAULT is
-- catalog-only. No data migration, no rewrite.
--
-- Down (development and test only) refuses while app.enrollment_sessions, app.enrollment_restrictions or
-- app.enrollment_requests holds any row, app.login_transactions holds an enrollment-shaped row, or audit.events holds a
-- row with an 00016 action, reason, operation or resource type
-- (inspected with FORCE lifted inside the transaction; a refusal rolls the lift back), then restores version 15.

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
  r record;
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regrole(db || '_runtime') IS NULL OR to_regrole(db || '_policy_reader') IS NULL OR to_regrole(db || '_auth_reader') IS NULL
     OR to_regrole(db || '_claim_writer') IS NULL THEN
    RAISE EXCEPTION 'roles for database % are missing: run infra/postgres/provision-roles.sql (or a fresh 10-roles.sh) first', db;
  END IF;
  ow := current_user::regrole; rt := (db || '_runtime')::regrole; pr := (db || '_policy_reader')::regrole;
  ar := (db || '_auth_reader')::regrole; cw := (db || '_claim_writer')::regrole;

  -- The 00015 helper checks, unchanged.
  FOR r IN SELECT rolname, rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolinherit, rolcanlogin, rolreplication
           FROM pg_roles WHERE oid IN (pr, ar, cw) LOOP
    IF r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication THEN
      RAISE EXCEPTION 'role % has an unsafe attribute: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
    IF r.rolcanlogin OR r.rolinherit THEN
      RAISE EXCEPTION 'helper role % must be NOLOGIN NOINHERIT: run infra/postgres/provision-roles.sql', r.rolname;
    END IF;
  END LOOP;
  IF (SELECT count(*) FROM pg_auth_members m WHERE m.roleid IN (pr, ar, cw) AND m.member = ow
        AND NOT m.inherit_option AND m.set_option AND NOT m.admin_option) <> 3 THEN
    RAISE EXCEPTION 'the owner must be a member of the three helper roles WITH INHERIT FALSE, SET TRUE: run infra/postgres/provision-roles.sql';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member IN (rt, cw)) THEN
    RAISE EXCEPTION 'the runtime and claim-writer roles must not be members of any role';
  END IF;

  -- Version 15: the migration history ends at 15, the 00015 objects exist, none of the 00016 objects exist.
  IF (SELECT max(version_id) FROM app.goose_db_version) IS DISTINCT FROM 15 THEN
    RAISE EXCEPTION 'version-15 precondition: the current version is %', (SELECT max(version_id) FROM app.goose_db_version);
  END IF;
  IF to_regclass('app.invitations') IS NULL
     OR to_regprocedure('app.role_bindings_authority_guard()') IS NULL
     OR to_regprocedure('app.rls_user_enabled_lock(uuid, text, uuid)') IS NULL
     OR to_regprocedure('app.invitation_claim_group(uuid, bytea)') IS NULL
     OR to_regprocedure('app.invitation_claim_organization(uuid, bytea, text, text, text)') IS NULL
     OR NOT EXISTS (SELECT 1 FROM pg_policy WHERE polrelid = 'app.role_bindings'::regclass AND polname = 'role_bindings_insert_invited')
     OR NOT EXISTS (SELECT 1 FROM pg_policy WHERE polrelid = 'app.invitations'::regclass AND polname = 'invitations_claim_select')
     OR to_regclass('app.enrollment_sessions') IS NOT NULL OR to_regclass('app.enrollment_restrictions') IS NOT NULL
     OR to_regclass('app.enrollment_requests') IS NOT NULL
     OR to_regprocedure('app.rls_enrollment_mode(uuid)') IS NOT NULL
     OR to_regprocedure('app.rls_invitation_pending(uuid, uuid)') IS NOT NULL
     OR to_regprocedure('app.auth_resolve_enrollment(bytea, timestamptz)') IS NOT NULL
     OR to_regprocedure('app.enrollment_lock_organization(uuid)') IS NOT NULL
     OR to_regprocedure('app.enrollment_admit(bytea, uuid, text, uuid)') IS NOT NULL
     OR to_regprocedure('app.enrollment_restrict_on_departure()') IS NOT NULL
     OR to_regprocedure('app.enrollment_request(bytea, uuid)') IS NOT NULL
     OR to_regprocedure('app.enrollment_restrict_on_decision()') IS NOT NULL
     OR EXISTS (SELECT 1 FROM pg_policy WHERE polrelid = 'audit.events'::regclass AND polname = 'events_insert_enrollment')
     OR EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.organizations'::regclass AND NOT attisdropped
                  AND attname IN ('enrollment_mode', 'enrollment_history_resolved_at', 'enrollment_history_basis'))
     OR EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'app.login_transactions'::regclass AND NOT attisdropped
                  AND attname IN ('enrollment_intent', 'invitation_id', 'request_on_arrival'))
     OR EXISTS (SELECT 1 FROM pg_policy WHERE polname IN ('organizations_policy_reader', 'organizations_claim_select',
                  'organizations_claim_lock')) THEN
    RAISE EXCEPTION 'version-15 precondition: enrollment objects already exist or the version-15 shapes are missing';
  END IF;
  -- The claim writer owns exactly its three 00015 bodies; the Organization claim is still owner-only (D-15 pending).
  IF (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 3
     OR has_function_privilege(rt, 'app.invitation_claim_organization(uuid, bytea, text, text, text)', 'EXECUTE')
     OR NOT has_function_privilege(ow, 'app.invitation_claim_organization(uuid, bytea, text, text, text)', 'EXECUTE') THEN
    RAISE EXCEPTION 'version-15 precondition: claim-writer inventory or invitation_claim_organization EXECUTE differs';
  END IF;
  -- Codex F1: rls_invitation_pending reads through the claim writer's existing 00015 invitation SELECT; nothing is
  -- granted on app.invitations here.
  IF NOT has_table_privilege(cw, 'app.invitations', 'SELECT') THEN
    RAISE EXCEPTION 'version-15 precondition: the claim writer cannot read app.invitations';
  END IF;
END
$$;
-- +goose StatementEnd

-- The enrollment mode (Plan r0 §4.1): every existing Organization becomes invite-only (closed default; no data decision).
ALTER TABLE app.organizations ADD COLUMN enrollment_mode text NOT NULL DEFAULT 'invite_only'
  CONSTRAINT organizations_enrollment_mode_valid CHECK (enrollment_mode IN ('invite_only', 'request_access', 'open'));

-- The readiness fact (addendum §A.1/§A.3 steps 2-4): added WITHOUT a default, so every Organization present at 00016 is
-- left unresolved (closed) — nothing is fabricated or resolved; the constraints are valid immediately; the defaults then
-- apply only to Organizations inserted afterwards (ready by construction: the ledger has recorded every departure since).
ALTER TABLE app.organizations ADD COLUMN enrollment_history_resolved_at timestamptz NULL,
  ADD COLUMN enrollment_history_basis text NULL;
ALTER TABLE app.organizations
  ADD CONSTRAINT organizations_enrollment_history_basis_valid CHECK (enrollment_history_basis IS NULL
    OR enrollment_history_basis IN ('created_after_ledger', 'operator_verified_absent', 'operator_backfilled')),
  ADD CONSTRAINT organizations_enrollment_history_shape CHECK ((enrollment_history_resolved_at IS NULL) = (enrollment_history_basis IS NULL)),
  ADD CONSTRAINT organizations_open_requires_resolved_history CHECK (enrollment_mode <> 'open' OR enrollment_history_resolved_at IS NOT NULL);
ALTER TABLE app.organizations ALTER COLUMN enrollment_history_resolved_at SET DEFAULT now(),
  ALTER COLUMN enrollment_history_basis SET DEFAULT 'created_after_ledger';

-- The enrollment-intent login (Plan r0 §4.1, r3 §4.1): the callback's branch facts; /join/status is the one new return
-- path.
ALTER TABLE app.login_transactions ADD COLUMN enrollment_intent boolean NOT NULL DEFAULT false,
  ADD COLUMN invitation_id uuid NULL,
  ADD COLUMN request_on_arrival boolean NOT NULL DEFAULT false;
ALTER TABLE app.login_transactions DROP CONSTRAINT login_transactions_return_path_check,
  ADD CONSTRAINT login_transactions_return_path_check CHECK (return_path IN ('/', '/account', '/join/status'));

-- The pending enrollment credential (Plan r3 §4.2): proof that an authenticated subject may complete one admission for
-- one Organization within 30 minutes; never exchanged for a session (D-17). admitted_* is the durable T2 proof (no
-- foreign key: a later G019 removal deletes the binding while the proof stays). provider_session_id is the logout-scope
-- key only.
CREATE TABLE app.enrollment_sessions (
  id                  uuid        NOT NULL,
  user_id             uuid        NOT NULL,
  organization_id     uuid        NOT NULL,
  origin              text        NOT NULL,
  credential_hash     bytea       NOT NULL,
  csrf_token          bytea       NOT NULL,
  provider_session_id text        NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  expires_at          timestamptz NOT NULL,
  admitted_at         timestamptz,
  admitted_binding_id uuid,
  revoked_at          timestamptz,
  revoke_reason       text,
  CONSTRAINT enrollment_sessions_pkey PRIMARY KEY (id),
  CONSTRAINT enrollment_sessions_user_fkey FOREIGN KEY (user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_sessions_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_sessions_credential_hash_key UNIQUE (credential_hash),
  CONSTRAINT enrollment_sessions_credential_hash_length CHECK (octet_length(credential_hash) = 32),
  CONSTRAINT enrollment_sessions_csrf_token_length CHECK (octet_length(csrf_token) = 32),
  CONSTRAINT enrollment_sessions_origin_length CHECK (char_length(origin) BETWEEN 1 AND 300),
  CONSTRAINT enrollment_sessions_provider_session_id_length CHECK (char_length(provider_session_id) BETWEEN 1 AND 255),
  CONSTRAINT enrollment_sessions_lifetime CHECK (expires_at > created_at AND expires_at <= created_at + interval '30 minutes'),
  CONSTRAINT enrollment_sessions_admission_shape CHECK ((admitted_at IS NULL) = (admitted_binding_id IS NULL)),
  CONSTRAINT enrollment_sessions_revocation_shape CHECK ((revoked_at IS NULL) = (revoke_reason IS NULL))
);
CREATE INDEX enrollment_sessions_user_idx ON app.enrollment_sessions (user_id, organization_id) WHERE revoked_at IS NULL;
CREATE INDEX enrollment_sessions_provider_idx ON app.enrollment_sessions (user_id, provider_session_id) WHERE revoked_at IS NULL;
ALTER TABLE app.enrollment_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.enrollment_sessions FORCE ROW LEVEL SECURITY;

-- The access requests (request slice; Plan r3 §1.4/§4.2, review r2 correction 4): one pending request per subject and
-- Organization; an Organization manager decides it (pending -> approved | denied) or revokes an unconsumed approval
-- (approved -> revoked); the claim writer consumes an approval once inside enrollment_admit. authority_seq is the
-- greatest restriction seq covered by the approval, derived by enrollment_requests_immutable under the Organization lock
-- (never supplied). Plan r3 §4.2 enumerates thirteen columns (its "15" count does not match its own list; the list is
-- bound, as for the ledger). The decider and revoker carry foreign keys (Plan r3 §4.2); consumed_binding_id does not (a
-- later G019 removal deletes the binding while the consumption proof stays).
CREATE TABLE app.enrollment_requests (
  id                  uuid        NOT NULL,
  organization_id     uuid        NOT NULL,
  user_id             uuid        NOT NULL,
  status              text        NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  expires_at          timestamptz NOT NULL,
  decided_at          timestamptz,
  decided_by_user_id  uuid,
  revoked_at          timestamptz,
  revoked_by_user_id  uuid,
  consumed_at         timestamptz,
  consumed_binding_id uuid,
  authority_seq       bigint      NOT NULL DEFAULT 0,
  CONSTRAINT enrollment_requests_pkey PRIMARY KEY (id),
  CONSTRAINT enrollment_requests_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_requests_user_fkey FOREIGN KEY (user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_requests_decided_by_fkey FOREIGN KEY (decided_by_user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_requests_revoked_by_fkey FOREIGN KEY (revoked_by_user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_requests_status_valid CHECK (status IN ('pending', 'approved', 'denied', 'revoked')),
  CONSTRAINT enrollment_requests_lifetime CHECK (expires_at > created_at AND expires_at <= created_at + interval '30 days'),
  CONSTRAINT enrollment_requests_decision_shape CHECK ((status = 'pending') = (decided_at IS NULL)
    AND (decided_at IS NULL) = (decided_by_user_id IS NULL)),
  CONSTRAINT enrollment_requests_revocation_shape CHECK ((status = 'revoked') = (revoked_at IS NOT NULL)
    AND (revoked_at IS NULL) = (revoked_by_user_id IS NULL)),
  CONSTRAINT enrollment_requests_consumption_shape CHECK ((consumed_at IS NULL) = (consumed_binding_id IS NULL)
    AND (consumed_at IS NULL OR status = 'approved')),
  CONSTRAINT enrollment_requests_authority_shape CHECK (authority_seq >= 0 AND (authority_seq = 0 OR status IN ('approved', 'revoked')))
);
CREATE UNIQUE INDEX enrollment_requests_one_pending ON app.enrollment_requests (organization_id, user_id) WHERE status = 'pending';
CREATE INDEX enrollment_requests_scope_idx ON app.enrollment_requests (organization_id, status, id);
CREATE INDEX enrollment_requests_user_idx ON app.enrollment_requests (user_id, organization_id, created_at);
ALTER TABLE app.enrollment_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.enrollment_requests FORCE ROW LEVEL SECURITY;

-- The restriction ledger (Plan r3 §1.4/§4.2, D-4r1; addendum §A.2 / correction 01: recorded_via admits 'backfill'). A
-- durable authority record: departures (1a), denials and revocations (request slice) restrict open signup until an
-- explicit authority that is not older than the row lifts it inside enrollment_admit. seq is the authority version.
-- source_request_id names the denied or revoked request (request slice). created_by_user_id and lifted_by_user_id are
-- attribution only (no foreign key, as Plan r3 §4.2).
CREATE TABLE app.enrollment_restrictions (
  id                 uuid        NOT NULL DEFAULT gen_random_uuid(),
  seq                bigint      NOT NULL GENERATED ALWAYS AS IDENTITY,
  organization_id    uuid        NOT NULL,
  user_id            uuid        NOT NULL,
  kind               text        NOT NULL,
  recorded_via       text        NOT NULL,
  created_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
  created_by_user_id uuid,
  source_request_id  uuid,
  lifted_at          timestamptz,
  lifted_by_user_id  uuid,
  lifted_via         text,
  CONSTRAINT enrollment_restrictions_pkey PRIMARY KEY (id),
  CONSTRAINT enrollment_restrictions_seq_key UNIQUE (seq),
  CONSTRAINT enrollment_restrictions_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_restrictions_user_fkey FOREIGN KEY (user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_restrictions_source_request_fkey FOREIGN KEY (source_request_id)
    REFERENCES app.enrollment_requests (id) ON DELETE RESTRICT,
  CONSTRAINT enrollment_restrictions_kind_valid CHECK (kind IN ('departed', 'denied', 'revoked')),
  CONSTRAINT enrollment_restrictions_recorded_via_valid CHECK (recorded_via IN ('trigger', 'replay', 'backfill')),
  CONSTRAINT enrollment_restrictions_lifted_via_valid CHECK (lifted_via IS NULL OR lifted_via IN ('invitation', 'approved_request')),
  CONSTRAINT enrollment_restrictions_lift_shape CHECK ((lifted_at IS NULL) = (lifted_via IS NULL)),
  CONSTRAINT enrollment_restrictions_replay_shape CHECK (recorded_via = 'trigger'
    OR (kind = 'departed' AND source_request_id IS NULL AND created_by_user_id IS NULL))
);
CREATE INDEX enrollment_restrictions_open_idx ON app.enrollment_restrictions (organization_id, user_id) WHERE lifted_at IS NULL;
ALTER TABLE app.enrollment_restrictions ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.enrollment_restrictions FORCE ROW LEVEL SECURITY;

-- The widened catalog constraints (Codex F5, round-2 audit decision, round-3 decision 1): reasons +7 (58 -> 65),
-- operations +11 (53 -> 64; enrollment.start is the callback T1 enrollment-policy/intent denial). Slice 1b (policy and
-- readiness, addendum §A.2/§A.7, correction 01 §C.2): actions +3 trigger-owned (39 -> 42 incl. the core actions:
-- enrollment.policy_changed, enrollment.readiness_changed, enrollment.restriction_backfilled) and reasons +3 (65 -> 68:
-- the two readiness bases and history_unresolved); the readiness events name the Organization. Request slice (Plan r3
-- §4.5): actions +3 trigger-owned (42 -> 45: enrollment.requested, enrollment.decided, enrollment.revoked) and resource
-- types +1 (10 -> 11: enrollment_request).
ALTER TABLE audit.events DROP CONSTRAINT events_reason_valid, DROP CONSTRAINT events_operation_valid,
  DROP CONSTRAINT events_action_valid, DROP CONSTRAINT events_resource_type_valid;
ALTER TABLE audit.events
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain', 'invitation', 'enrollment_request')),
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
                    'object.deletion_pending', 'revocation.unaudited')),
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
                    'operator_verified_absent', 'operator_backfilled', 'history_unresolved')),
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
                    'enrollment.requests.decide', 'enrollment.requests.revoke'));

-- The attribution and shape rules (the 00015 body otherwise unchanged; CREATE OR REPLACE keeps identity, owner, ACL and
-- trigger), Codex F5 plus the round-2 audit decision and the round-3 operation-level scope rule, each combination
-- narrowed to its reviewed producer:
--   - auth.sign_in denied admits recipient_proof_unavailable (the consumed callback's recipient-proof failure);
--   - access.denied admits the enrollment refusal reasons (enrollment_closed, enrollment_restricted, approval_revoked,
--     invitation_mismatch, recipient_ineligible, session_invalid and the reused host_changed) only on an enrollment
--     operation, without an Organization, Group or resource;
--   - enrollment.start (the callback T1 policy/intent denial) and enrollment.admit (the IDENTITY-shape T2 admission)
--     never carry an Organization, Group or resource, whatever the reason: enrollment.start only with enrollment_closed
--     or invitation_mismatch (attributed only when the callback established the subject); enrollment.admit only with
--     its reasons or target_not_found and only with an established actor;
--   - slice 1b: the three trigger-owned enrollment captures name their Organization as the resource:
--     enrollment.policy_changed (the mode changed; no reason), enrollment.readiness_changed (the basis or
--     history_unresolved; the operator channel only, so actor unknown) and enrollment.restriction_backfilled (the
--     backfilled subject, operator_backfilled; actor unknown);
--   - request slice: the three trigger-owned request captures name the request as the resource and the applicant as the
--     subject, attributed to a user: enrollment.requested (the applicant itself), enrollment.decided (approved or
--     denied; the decider) and enrollment.revoked (the revoker); no reason.
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
               THEN NEW.reason IN ('group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled')
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
  db text := current_database();
  ow text := current_user;
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
  ar text := current_database() || '_auth_reader';
  cw text := current_database() || '_claim_writer';
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  B text := '(SELECT app.ctx_login_binding())';
  R text := '(SELECT app.rls_org_role())';
  MGR text := '(SELECT app.rls_org_role()) IN (''org_owner'', ''org_admin'')';
BEGIN
  -- app.organizations (Plan r3 §4.4, review r2 correction 1). The policy reader reads the public mode for
  -- rls_enrollment_mode. The claim writer reads (id, enrollment_mode) and holds UPDATE (enrollment_mode) only so that
  -- FOR SHARE / FOR NO KEY UPDATE are permitted (the users_claim_lock precedent, 00015:661/676): the lock-only policy's
  -- WITH CHECK (false) makes every claim-writer UPDATE a policy violation.
  EXECUTE format('GRANT SELECT (id, enrollment_mode) ON app.organizations TO %I', pr);
  EXECUTE format('CREATE POLICY organizations_policy_reader ON app.organizations FOR SELECT TO %I USING (true)', pr);
  EXECUTE format('GRANT SELECT (id, enrollment_mode), UPDATE (enrollment_mode) ON app.organizations TO %I', cw);
  EXECUTE format('CREATE POLICY organizations_claim_select ON app.organizations FOR SELECT TO %I USING (true)', cw);
  EXECUTE format('CREATE POLICY organizations_claim_lock ON app.organizations FOR UPDATE TO %I USING (true) WITH CHECK (false)', cw);

  -- app.enrollment_sessions (Plan r3 §4.4). Runtime/owner: self-row SELECT; INSERT of the issuing columns only (never
  -- the admission or revocation pair) in the IDENTITY shape; UPDATE of the revocation pair only; DELETE only of expired
  -- rows in the EMPTY shape (the 00006 login_transactions_expire shape: Service.Login's cleanup is the sweep). The
  -- bootstrap lookup by digest is app.auth_resolve_enrollment, never these policies.
  EXECUTE format('GRANT SELECT, DELETE ON app.enrollment_sessions TO %I', rt);
  EXECUTE format('GRANT INSERT (id, user_id, organization_id, origin, credential_hash, csrf_token, provider_session_id, created_at, '
              || 'expires_at) ON app.enrollment_sessions TO %I', rt);
  EXECUTE format('GRANT UPDATE (revoked_at, revoke_reason) ON app.enrollment_sessions TO %I', rt);
  EXECUTE format('CREATE POLICY enrollment_sessions_select ON app.enrollment_sessions FOR SELECT TO %I, %I USING (user_id = %s)',
    rt, ow, U);
  EXECUTE format('CREATE POLICY enrollment_sessions_insert ON app.enrollment_sessions FOR INSERT TO %I, %I WITH CHECK ('
    || 'user_id = %s AND %s IS NULL AND revoked_at IS NULL AND admitted_at IS NULL)', rt, ow, U, O);
  EXECUTE format('CREATE POLICY enrollment_sessions_update ON app.enrollment_sessions FOR UPDATE TO %I, %I USING (user_id = %s) '
    || 'WITH CHECK (user_id = %s AND %s IS NULL)', rt, ow, U, U, O);
  EXECUTE format('CREATE POLICY enrollment_sessions_expire ON app.enrollment_sessions FOR DELETE TO %I, %I USING ('
    || 'expires_at <= now() AND %s IS NULL AND %s IS NULL)', rt, ow, U, B);
  -- Auth reader (the auth_resolve_enrollment owner) and claim writer (enrollment_admit: own-row lock and the T2 pair).
  EXECUTE format('GRANT SELECT ON app.enrollment_sessions TO %I', ar);
  EXECUTE format('CREATE POLICY enrollment_sessions_auth_reader ON app.enrollment_sessions FOR SELECT TO %I USING (true)', ar);
  EXECUTE format('GRANT SELECT, UPDATE (admitted_at, admitted_binding_id) ON app.enrollment_sessions TO %I', cw);
  EXECUTE format('CREATE POLICY enrollment_sessions_claim_select ON app.enrollment_sessions FOR SELECT TO %I USING (true)', cw);
  EXECUTE format('CREATE POLICY enrollment_sessions_claim_update ON app.enrollment_sessions FOR UPDATE TO %I USING (true) '
    || 'WITH CHECK (user_id = app.ctx_user())', cw);

  -- app.enrollment_restrictions (Plan r3 §4.4, review r1 correction 2). Runtime/owner read the subject's own rows or, as
  -- an Organization manager, the context Organization's rows. The claim writer reads and lifts (enrollment_admit). The
  -- owner's only INSERT path is the trigger-bound policy below (the owner is bound by FORCE RLS): unusable at depth 0,
  -- trigger-recorded departures/denials/revocations only, never lifted. No INSERT grant or policy for rt/cw/pr/ar; no
  -- DELETE for any role; no owner UPDATE/DELETE policy. The identity sequence stays owner-only (the schema's default
  -- sequence grant to the runtime is revoked: no session role inserts, so none needs it).
  EXECUTE format('GRANT SELECT ON app.enrollment_restrictions TO %I', rt);
  EXECUTE format('CREATE POLICY enrollment_restrictions_select ON app.enrollment_restrictions FOR SELECT TO %I, %I USING ('
    || 'user_id = %s OR (organization_id = %s AND %s))', rt, ow, U, O, MGR);
  EXECUTE format('GRANT SELECT, UPDATE (lifted_at, lifted_by_user_id, lifted_via) ON app.enrollment_restrictions TO %I', cw);
  EXECUTE format('CREATE POLICY enrollment_restrictions_claim_select ON app.enrollment_restrictions FOR SELECT TO %I USING (true)', cw);
  EXECUTE format('CREATE POLICY enrollment_restrictions_claim_lift ON app.enrollment_restrictions FOR UPDATE TO %I USING (true) '
    || 'WITH CHECK (user_id = app.ctx_user() AND lifted_at IS NOT NULL)', cw);
  EXECUTE format('CREATE POLICY enrollment_restrictions_owner_insert ON app.enrollment_restrictions FOR INSERT TO %I WITH CHECK ('
    || 'pg_trigger_depth() >= 1 AND kind IN (''departed'', ''denied'', ''revoked'') AND recorded_via = ''trigger'' '
    || 'AND lifted_at IS NULL AND lifted_by_user_id IS NULL AND lifted_via IS NULL)', ow);
  EXECUTE format('REVOKE ALL ON SEQUENCE %s FROM %I', pg_get_serial_sequence('app.enrollment_restrictions', 'seq'), rt);

  -- app.enrollment_requests (request slice; Plan r3 §4.4, r3 correction 2). Runtime/owner read the subject's own requests
  -- or, as an Organization manager, the context Organization's; an Organization manager decides a pending request or
  -- revokes an approval (the five decision/revocation columns only: no privilege on authority_seq, which the BEFORE
  -- guard derives; no INSERT, no DELETE). The claim writer inserts the applicant's pending request (enrollment_request)
  -- and consumes an approval (enrollment_admit; the consumption pair only).
  EXECUTE format('GRANT SELECT ON app.enrollment_requests TO %I', rt);
  EXECUTE format('GRANT UPDATE (status, decided_at, decided_by_user_id, revoked_at, revoked_by_user_id) ON app.enrollment_requests TO %I', rt);
  EXECUTE format('CREATE POLICY enrollment_requests_select ON app.enrollment_requests FOR SELECT TO %I, %I USING ('
    || 'user_id = %s OR (organization_id = %s AND %s))', rt, ow, U, O, MGR);
  EXECUTE format('CREATE POLICY enrollment_requests_decide ON app.enrollment_requests FOR UPDATE TO %I, %I USING ('
    || 'organization_id = %s AND %s AND status IN (''pending'', ''approved'')) WITH CHECK ('
    || 'organization_id = %s AND %s AND status IN (''approved'', ''denied'', ''revoked''))', rt, ow, O, MGR, O, MGR);
  EXECUTE format('GRANT SELECT, INSERT (id, organization_id, user_id, status, expires_at), UPDATE (consumed_at, consumed_binding_id) '
              || 'ON app.enrollment_requests TO %I', cw);
  EXECUTE format('CREATE POLICY enrollment_requests_claim_insert ON app.enrollment_requests FOR INSERT TO %I WITH CHECK ('
    || 'user_id = app.ctx_user() AND status = ''pending'')', cw);
  EXECUTE format('CREATE POLICY enrollment_requests_claim_select ON app.enrollment_requests FOR SELECT TO %I USING (true)', cw);
  EXECUTE format('CREATE POLICY enrollment_requests_claim_update ON app.enrollment_requests FOR UPDATE TO %I USING (true) '
    || 'WITH CHECK (user_id = app.ctx_user())', cw);
  -- The applicant's request capture (Plan r3 §4.4): the claim writer appends enrollment.requested only, inside the capture
  -- trigger, attributed to itself, with the request's Organization. events_insert_claim and the runtime events_insert are
  -- unchanged; decision and revocation rows are appended by the deciding runtime session under events_insert.
  EXECUTE format('CREATE POLICY events_insert_enrollment ON audit.events FOR INSERT TO %I WITH CHECK ('
    || 'pg_trigger_depth() >= 1 AND actor_kind = ''user'' AND actor_user_id = app.ctx_user() AND organization_id IS NOT NULL '
    || 'AND action = ''enrollment.requested'')', cw);

  -- The policy reader's public mode reader (Plan r0 §4.3): readable with no tenant context; an absent row is
  -- 'invite_only'; discloses nothing but the mode.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pr);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  EXECUTE $r$
CREATE FUNCTION app.rls_enrollment_mode(p_org uuid) RETURNS text
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN coalesce((SELECT o.enrollment_mode FROM app.organizations o WHERE o.id = p_org), 'invite_only');
END $$
$r$;
  REVOKE ALL ON FUNCTION app.rls_enrollment_mode(uuid) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_enrollment_mode(uuid) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pr);

  -- The auth reader's bootstrap resolver (the auth_resolve_session pattern, 00010:596-609): the only way the runtime finds
  -- an enrollment row before ctx_user() exists; a live, unrevoked row of an active user.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $r$
CREATE FUNCTION app.auth_resolve_enrollment(p_hash bytea, p_at timestamptz)
  RETURNS TABLE (session_id uuid, user_id uuid, organization_id uuid, origin text, csrf_token bytea, expires_at timestamptz,
                 admitted_at timestamptz, admitted_binding_id uuid)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
BEGIN
  RETURN QUERY
  SELECT es.id, es.user_id, es.organization_id, es.origin, es.csrf_token, es.expires_at, es.admitted_at, es.admitted_binding_id
  FROM app.enrollment_sessions es JOIN app.users u ON u.id = es.user_id
  WHERE es.credential_hash = p_hash AND es.revoked_at IS NULL AND es.expires_at > p_at AND u.status = 'active';
END $$
$r$;
  REVOKE ALL ON FUNCTION app.auth_resolve_enrollment(bytea, timestamptz) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.auth_resolve_enrollment(bytea, timestamptz) TO %I, %I', rt, ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);

  -- The claim writer's three 1a bodies, created as the claim writer (the 00015 SET LOCAL ROLE pattern).
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', cw);
  EXECUTE format('SET LOCAL ROLE %I', cw);
  -- Invitation-path admission predicate (Plan r0 §4.3 D-5; Codex F1: claim-writer owned through its existing 00015
  -- SELECT on app.invitations and invitations_claim_select). Fixed SQL, boolean only: a pending (unconsumed, unrevoked,
  -- unexpired) Organization invitation p_id of p_org; no address, role, issuer or digest leaves the body.
  EXECUTE $r$
CREATE FUNCTION app.rls_invitation_pending(p_id uuid, p_org uuid) RETURNS boolean
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF p_id IS NULL OR p_org IS NULL THEN
    RETURN false;
  END IF;
  RETURN EXISTS (SELECT 1 FROM app.invitations i
                 WHERE i.id = p_id AND i.organization_id = p_org AND i.scope_type = 'organization'
                   AND i.consumed_at IS NULL AND i.revoked_at IS NULL AND i.expires_at > now());
END $$
$r$;
  -- The Organization lock for the owner-owned trigger bodies (Plan r3 correction 1): FOR NO KEY UPDATE under
  -- organizations_claim_lock, valid after the caller's own membership binding is gone and under the superuser replay
  -- alike; reads no column, returns no row data. EXECUTE for the owner only.
  EXECUTE $r$
CREATE FUNCTION app.enrollment_lock_organization(p_org uuid) RETURNS boolean
  LANGUAGE plpgsql VOLATILE STRICT SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  PERFORM 1 FROM app.organizations o WHERE o.id = p_org FOR NO KEY UPDATE;
  RETURN FOUND;
END $$
$r$;
  -- The access request (request slice; Plan r3 §4.3): IDENTITY shape; each step a separate command in the executable
  -- lock order (1) user FOR SHARE, (2) own enrollment session FOR UPDATE, (3) Organization FOR SHARE, then the mode
  -- (request_access only), an existing membership (already_member), an existing pending request (pending, its id, no
  -- write) and the one INSERT under enrollment_requests_claim_insert (its capture appends enrollment.requested under
  -- events_insert_enrollment). A prior denial or revocation does not refuse a new request (an administrator re-decides).
  -- An expired pending request keeps occupying one_pending and is returned as pending: never renewed, never consumable
  -- (enrollment_admit refuses an expired approval); an administrator resolves it (Plan r3 §4.3 step 6; Codex round-8
  -- disposition 3). A concurrent first request from a second credential of the same subject resolves to the one
  -- pending row (ON CONFLICT on one_pending, then the committed row) instead of a unique violation.
  EXECUTE $r$
CREATE FUNCTION app.enrollment_request(p_hash bytea, p_org uuid) RETURNS TABLE (outcome text, request_id uuid)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  org_mode text;
  rid uuid;
BEGIN
  IF u IS NULL OR app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL THEN
    RAISE EXCEPTION 'enrollment request context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'enrollment_context';
  END IF;
  IF p_hash IS NULL OR p_org IS NULL THEN
    RAISE EXCEPTION 'enrollment_request_invalid' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  -- (1) the subject, enabled, under a lock consistent with DisableUser.
  IF NOT coalesce(app.rls_user_enabled_lock(u, 'organization', p_org), false) THEN
    RETURN QUERY SELECT 'recipient_ineligible'::text, NULL::uuid;
    RETURN;
  END IF;
  -- (2) the subject's own live enrollment credential for this Organization.
  PERFORM 1 FROM app.enrollment_sessions es
  WHERE es.credential_hash = p_hash AND es.user_id = u AND es.organization_id = p_org AND es.revoked_at IS NULL
    AND es.expires_at > now()
  FOR UPDATE;
  IF NOT FOUND THEN
    RETURN QUERY SELECT 'enrollment_expired'::text, NULL::uuid;
    RETURN;
  END IF;
  -- (3) the Organization, shared; (4) the mode, read after the lock.
  PERFORM 1 FROM app.organizations o WHERE o.id = p_org FOR SHARE;
  IF NOT FOUND THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid;
    RETURN;
  END IF;
  SELECT o.enrollment_mode INTO org_mode FROM app.organizations o WHERE o.id = p_org;
  IF org_mode IS DISTINCT FROM 'request_access' THEN
    RETURN QUERY SELECT 'enrollment_closed'::text, NULL::uuid;
    RETURN;
  END IF;
  -- (5) a current Organization membership.
  IF EXISTS (SELECT 1 FROM app.role_bindings rb
             WHERE rb.subject_user_id = u AND rb.organization_id = p_org AND rb.scope_type = 'organization'
               AND rb.scope_id = p_org AND rb.source = 'membership') THEN
    RETURN QUERY SELECT 'already_member'::text, NULL::uuid;
    RETURN;
  END IF;
  -- (6) the existing pending request (idempotent while pending).
  SELECT r.id INTO rid FROM app.enrollment_requests r
  WHERE r.organization_id = p_org AND r.user_id = u AND r.status = 'pending';
  IF FOUND THEN
    RETURN QUERY SELECT 'pending'::text, rid;
    RETURN;
  END IF;
  -- (7) the one write.
  rid := gen_random_uuid();
  INSERT INTO app.enrollment_requests (id, organization_id, user_id, status, expires_at)
  VALUES (rid, p_org, u, 'pending', now() + interval '30 days')
  ON CONFLICT (organization_id, user_id) WHERE status = 'pending' DO NOTHING;
  IF NOT FOUND THEN
    SELECT r.id INTO rid FROM app.enrollment_requests r
    WHERE r.organization_id = p_org AND r.user_id = u AND r.status = 'pending';
    IF NOT FOUND THEN
      RAISE EXCEPTION 'enrollment_request_conflict' USING ERRCODE = 'serialization_failure';
    END IF;
  END IF;
  RETURN QUERY SELECT 'pending'::text, rid;
END $$
$r$;
  -- Admission (Plan r3 §1.4/§4.3): IDENTITY shape; each step a separate command in the executable lock order (1) user
  -- FOR SHARE, (2) own enrollment session FOR UPDATE, (3) Organization FOR SHARE, (4) the request row FOR UPDATE
  -- (request), (5) restriction rows FOR UPDATE (lift, request), (6) binding INSERT (open, request), (7) request
  -- consumption (request), lift of the covered restrictions, (8) admitted_* UPDATE. Kinds: 'lift' (p_request = the
  -- invitation the G019 claim consumed earlier in this transaction; an unlifted restriction not older than it RAISEs
  -- P0001 enrollment_restricted so the claim rolls back), 'open' and 'request' (p_request = the subject's approved,
  -- unconsumed, unexpired request whose decider still manages the Organization; an unlifted restriction newer than its
  -- authority_seq refuses with zero writes; the rows it covers are lifted).
  EXECUTE $r$
CREATE FUNCTION app.enrollment_admit(p_hash bytea, p_org uuid, p_kind text, p_request uuid)
  RETURNS TABLE (outcome text, binding_id uuid)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  sess_id uuid;
  sess_admitted timestamptz;
  sess_binding uuid;
  org_mode text;
  inv_created timestamptz;
  req_status text;
  req_expires timestamptz;
  req_decider uuid;
  req_consumed timestamptz;
  req_seq bigint;
  ex_id uuid;
  bid uuid;
BEGIN
  IF u IS NULL OR app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL THEN
    RAISE EXCEPTION 'enrollment admission context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'enrollment_context';
  END IF;
  IF p_hash IS NULL OR p_org IS NULL OR p_kind IS NULL OR p_kind NOT IN ('lift', 'open', 'request')
     OR (p_kind IN ('lift', 'request')) <> (p_request IS NOT NULL) THEN
    RAISE EXCEPTION 'enrollment_kind_invalid' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  -- (1) the subject, enabled, under a lock consistent with DisableUser.
  IF NOT coalesce(app.rls_user_enabled_lock(u, 'organization', p_org), false) THEN
    RETURN QUERY SELECT 'recipient_ineligible'::text, NULL::uuid;
    RETURN;
  END IF;
  -- (2) the subject's own live enrollment credential for this Organization; an admitted one reconciles by read. The
  -- historical admitted_* proof is not current membership (Codex round-3 decision 3): already_member only while the
  -- recorded binding is still the subject's current Organization membership binding; a removed or replaced binding is
  -- enrollment_restricted (nothing granted, the proof not rewritten; reinstatement needs a fresh authorized enrollment).
  SELECT es.id, es.admitted_at, es.admitted_binding_id INTO sess_id, sess_admitted, sess_binding
  FROM app.enrollment_sessions es
  WHERE es.credential_hash = p_hash AND es.user_id = u AND es.organization_id = p_org AND es.revoked_at IS NULL
    AND es.expires_at > now()
  FOR UPDATE;
  IF NOT FOUND THEN
    RETURN QUERY SELECT 'enrollment_expired'::text, NULL::uuid;
    RETURN;
  END IF;
  IF sess_admitted IS NOT NULL THEN
    IF EXISTS (SELECT 1 FROM app.role_bindings rb
               WHERE rb.id = sess_binding AND rb.subject_user_id = u AND rb.organization_id = p_org
                 AND rb.scope_type = 'organization' AND rb.scope_id = p_org AND rb.source = 'membership') THEN
      RETURN QUERY SELECT 'already_member'::text, sess_binding;
    ELSE
      RETURN QUERY SELECT 'enrollment_restricted'::text, NULL::uuid;
    END IF;
    RETURN;
  END IF;
  -- (3) the Organization, shared: serializes with mode changes, departures and decisions (FOR NO KEY UPDATE holders).
  PERFORM 1 FROM app.organizations o WHERE o.id = p_org FOR SHARE;
  IF NOT FOUND THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid;
    RETURN;
  END IF;
  -- (4) the mode, read after the lock: open requires 'open'; request requires request_access or open; lift requires
  -- nothing.
  SELECT o.enrollment_mode INTO org_mode FROM app.organizations o WHERE o.id = p_org;
  IF (p_kind = 'open' AND org_mode IS DISTINCT FROM 'open')
     OR (p_kind = 'request' AND (org_mode IS NULL OR org_mode NOT IN ('request_access', 'open'))) THEN
    RETURN QUERY SELECT 'enrollment_closed'::text, NULL::uuid;
    RETURN;
  END IF;
  IF p_kind = 'request' THEN
    -- (5) the subject's own request of this Organization, locked (serializes with revocation and a second consumer).
    SELECT r.status, r.expires_at, r.decided_by_user_id, r.consumed_at, r.authority_seq
      INTO req_status, req_expires, req_decider, req_consumed, req_seq
    FROM app.enrollment_requests r WHERE r.id = p_request AND r.organization_id = p_org AND r.user_id = u
    FOR UPDATE;
    IF NOT FOUND THEN
      RETURN QUERY SELECT 'not_found'::text, NULL::uuid;
      RETURN;
    END IF;
    IF req_status = 'pending' THEN
      RETURN QUERY SELECT 'pending'::text, NULL::uuid;
      RETURN;
    ELSIF req_status = 'denied' THEN
      RETURN QUERY SELECT 'denied'::text, NULL::uuid;
      RETURN;
    END IF;
    -- A consumed approval admits nothing again: the current membership reconciles; without one, the authority is spent.
    IF req_consumed IS NOT NULL THEN
      SELECT rb.id INTO ex_id FROM app.role_bindings rb
      WHERE rb.subject_user_id = u AND rb.organization_id = p_org AND rb.scope_type = 'organization' AND rb.scope_id = p_org
        AND rb.source = 'membership';
      IF ex_id IS NOT NULL THEN
        RETURN QUERY SELECT 'already_member'::text, ex_id;
      ELSE
        RETURN QUERY SELECT 'approval_revoked'::text, NULL::uuid;
      END IF;
      RETURN;
    END IF;
    -- Revoked, expired, or a decider who no longer manages the Organization: the approval is not current authority.
    IF req_status IS DISTINCT FROM 'approved' OR req_expires <= now()
       OR NOT EXISTS (SELECT 1 FROM app.role_bindings rb
                      WHERE rb.subject_user_id = req_decider AND rb.organization_id = p_org AND rb.scope_type = 'organization'
                        AND rb.scope_id = p_org AND rb.source = 'membership' AND rb.role IN ('org_owner', 'org_admin')) THEN
      RETURN QUERY SELECT 'approval_revoked'::text, NULL::uuid;
      RETURN;
    END IF;
    PERFORM 1 FROM app.enrollment_restrictions r
    WHERE r.organization_id = p_org AND r.user_id = u AND r.lifted_at IS NULL
    ORDER BY r.seq FOR UPDATE;
    -- (6') a restriction newer than the approval's authority version refuses, zero writes.
    IF EXISTS (SELECT 1 FROM app.enrollment_restrictions r
               WHERE r.organization_id = p_org AND r.user_id = u AND r.lifted_at IS NULL AND r.seq > req_seq) THEN
      RETURN QUERY SELECT 'enrollment_restricted'::text, NULL::uuid;
      RETURN;
    END IF;
  ELSIF p_kind = 'lift' THEN
    -- (5) the invitation the same transaction's claim consumed for this subject.
    SELECT i.created_at INTO inv_created FROM app.invitations i
    WHERE i.id = p_request AND i.organization_id = p_org AND i.scope_type = 'organization' AND i.consumed_by_user_id = u;
    IF NOT FOUND THEN
      RETURN QUERY SELECT 'not_found'::text, NULL::uuid;
      RETURN;
    END IF;
    PERFORM 1 FROM app.enrollment_restrictions r
    WHERE r.organization_id = p_org AND r.user_id = u AND r.lifted_at IS NULL
    ORDER BY r.seq FOR UPDATE;
    -- (6) a restriction not older than the invitation: the claim must roll back.
    IF EXISTS (SELECT 1 FROM app.enrollment_restrictions r
               WHERE r.organization_id = p_org AND r.user_id = u AND r.lifted_at IS NULL AND r.created_at >= inv_created) THEN
      RAISE EXCEPTION 'enrollment_restricted' USING ERRCODE = 'P0001';
    END IF;
  ELSIF EXISTS (SELECT 1 FROM app.enrollment_restrictions r
                WHERE r.organization_id = p_org AND r.user_id = u AND r.lifted_at IS NULL) THEN
    -- (6) open: any unlifted restriction refuses.
    RETURN QUERY SELECT 'enrollment_restricted'::text, NULL::uuid;
    RETURN;
  END IF;
  -- (7) the current Organization binding.
  SELECT rb.id INTO ex_id FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.organization_id = p_org AND rb.scope_type = 'organization' AND rb.scope_id = p_org
    AND rb.source = 'membership';
  IF p_kind IN ('open', 'request') THEN
    IF ex_id IS NOT NULL THEN
      RETURN QUERY SELECT 'already_member'::text, ex_id;
      RETURN;
    END IF;
    -- (8) the ordinary member binding under role_bindings_insert_invited (its capture appends membership.granted under
    -- events_insert_claim).
    bid := gen_random_uuid();
    INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
    VALUES (bid, p_org, u, 'organization', p_org, NULL, 'org_member', 'membership');
    IF p_kind = 'request' THEN
      -- the one consumption under enrollment_requests_claim_update, then (9) lift exactly the rows the approval covers.
      UPDATE app.enrollment_requests r SET consumed_at = now(), consumed_binding_id = bid WHERE r.id = p_request;
      UPDATE app.enrollment_restrictions r SET lifted_at = now(), lifted_by_user_id = req_decider, lifted_via = 'approved_request'
      WHERE r.organization_id = p_org AND r.user_id = u AND r.lifted_at IS NULL AND r.seq <= req_seq;
    END IF;
  ELSE
    IF ex_id IS NULL THEN
      RETURN QUERY SELECT 'not_found'::text, NULL::uuid;
      RETURN;
    END IF;
    bid := ex_id;
    -- (9) lift exactly the rows older than the invitation.
    UPDATE app.enrollment_restrictions r SET lifted_at = now(), lifted_by_user_id = NULL, lifted_via = 'invitation'
    WHERE r.organization_id = p_org AND r.user_id = u AND r.lifted_at IS NULL AND r.created_at < inv_created;
  END IF;
  -- (10) the durable T2 proof.
  UPDATE app.enrollment_sessions es SET admitted_at = now(), admitted_binding_id = bid WHERE es.id = sess_id;
  RETURN QUERY SELECT 'admitted'::text, bid;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.rls_invitation_pending(uuid, uuid), app.enrollment_lock_organization(uuid),
    app.enrollment_request(bytea, uuid), app.enrollment_admit(bytea, uuid, text, uuid) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_invitation_pending(uuid, uuid), app.enrollment_request(bytea, uuid), '
              || 'app.enrollment_admit(bytea, uuid, text, uuid) TO %I, %I', rt, ow);
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.enrollment_lock_organization(uuid) TO %I', ow);
  -- D-15 (i) (owner-selected): the runtime consumes Organization invitations directly.
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.invitation_claim_organization(uuid, bytea, text, text, text) TO %I', rt);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', cw);

  -- The departure body (owner-owned; Plan r3 §4.3/§4.5, corrections 1-2): refuses outside its AFTER ROW DELETE trigger
  -- on app.role_bindings, takes the Organization lock through the claim writer's helper, then records the departure
  -- through enrollment_restrictions_owner_insert (pg_trigger_depth() >= 1). A refused insert raises inside the deleting
  -- statement, so the removal rolls back with it. ctx_user() is NULL under the superuser replay.
  EXECUTE $r$
CREATE FUNCTION app.enrollment_restrict_on_departure() RETURNS trigger
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF pg_trigger_depth() < 1 OR TG_WHEN IS DISTINCT FROM 'AFTER' OR TG_LEVEL IS DISTINCT FROM 'ROW'
     OR TG_TABLE_SCHEMA IS DISTINCT FROM 'app' OR TG_TABLE_NAME IS DISTINCT FROM 'role_bindings' OR TG_OP IS DISTINCT FROM 'DELETE' THEN
    RAISE EXCEPTION 'enrollment restriction source refused'
      USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'enrollment_restriction_source';
  END IF;
  IF OLD.scope_type IS DISTINCT FROM 'organization' OR OLD.source IS DISTINCT FROM 'membership' THEN
    RETURN NULL;
  END IF;
  IF NOT coalesce(app.enrollment_lock_organization(OLD.organization_id), false) THEN
    RAISE EXCEPTION 'organization_missing' USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'organization_missing';
  END IF;
  INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via, created_by_user_id)
  VALUES (OLD.organization_id, OLD.subject_user_id, 'departed', 'trigger', app.ctx_user());
  RETURN NULL;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.enrollment_restrict_on_departure() FROM PUBLIC;
  -- AFTER every G019 BEFORE guard (role_bindings_authority_guard) and alongside the AFTER audit capture.
  CREATE TRIGGER role_bindings_enrollment_departure AFTER DELETE ON app.role_bindings
    FOR EACH ROW WHEN (OLD.scope_type = 'organization' AND OLD.source = 'membership')
    EXECUTE FUNCTION app.enrollment_restrict_on_departure();

  -- Request slice: the decision body (owner-owned; Plan r3 §4.3, corrections 1-2): refuses outside its AFTER ROW UPDATE
  -- trigger on app.enrollment_requests, takes the Organization lock through the helper (a no-op re-lock when the
  -- deciding facade already holds it), then records the denial or revocation through
  -- enrollment_restrictions_owner_insert. A refused insert raises inside the deciding statement, so the decision rolls back.
  EXECUTE $r$
CREATE FUNCTION app.enrollment_restrict_on_decision() RETURNS trigger
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF pg_trigger_depth() < 1 OR TG_WHEN IS DISTINCT FROM 'AFTER' OR TG_LEVEL IS DISTINCT FROM 'ROW'
     OR TG_TABLE_SCHEMA IS DISTINCT FROM 'app' OR TG_TABLE_NAME IS DISTINCT FROM 'enrollment_requests' OR TG_OP IS DISTINCT FROM 'UPDATE' THEN
    RAISE EXCEPTION 'enrollment restriction source refused'
      USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'enrollment_restriction_source';
  END IF;
  IF NEW.status NOT IN ('denied', 'revoked') OR OLD.status IS NOT DISTINCT FROM NEW.status THEN
    RETURN NULL;
  END IF;
  IF NOT coalesce(app.enrollment_lock_organization(NEW.organization_id), false) THEN
    RAISE EXCEPTION 'organization_missing' USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'organization_missing';
  END IF;
  INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via, created_by_user_id, source_request_id)
  VALUES (NEW.organization_id, NEW.user_id, NEW.status, 'trigger', coalesce(NEW.revoked_by_user_id, NEW.decided_by_user_id), NEW.id);
  RETURN NULL;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.enrollment_restrict_on_decision() FROM PUBLIC;
  CREATE TRIGGER enrollment_requests_restrict AFTER UPDATE OF status ON app.enrollment_requests
    FOR EACH ROW WHEN (NEW.status IN ('denied', 'revoked') AND OLD.status IS DISTINCT FROM NEW.status)
    EXECUTE FUNCTION app.enrollment_restrict_on_decision();

  -- The request state machine (owner-owned SECURITY INVOKER; Plan r3 §4.5, r3 correction 2): the identity, scope and
  -- timestamps of a request never change; pending -> approved | denied once (with decided_*); on approved the guard
  -- itself locks the Organization FOR NO KEY UPDATE (a same-transaction no-op under the deciding facade's lock) and
  -- derives authority_seq from the subject's restrictions, ignoring any supplied value; approved -> revoked once (with
  -- revoked_*) while unconsumed; the consumption pair once on an unconsumed approval; authority_seq never changes after
  -- the approval; anything else is refused. The recorded decider/revoker is the acting user (ctx_user()), never a
  -- supplied third party: enrollment_admit's "decider still manages the Organization" check and the decision event's
  -- actor then name the same person.
  EXECUTE $r$
CREATE FUNCTION app.enrollment_requests_immutable() RETURNS trigger
  LANGUAGE plpgsql SECURITY INVOKER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  ok boolean;
BEGIN
  ok := NEW.id = OLD.id AND NEW.organization_id = OLD.organization_id AND NEW.user_id = OLD.user_id
    AND NEW.created_at = OLD.created_at AND NEW.expires_at = OLD.expires_at;
  IF ok AND OLD.status = 'pending' AND NEW.status IN ('approved', 'denied') THEN
    ok := NEW.revoked_at IS NULL AND NEW.revoked_by_user_id IS NULL AND NEW.consumed_at IS NULL AND NEW.consumed_binding_id IS NULL
      AND NEW.decided_by_user_id = app.ctx_user();
    IF ok AND NEW.status = 'approved' THEN
      PERFORM 1 FROM app.organizations o WHERE o.id = NEW.organization_id FOR NO KEY UPDATE;
      IF NOT FOUND THEN
        RAISE EXCEPTION 'enrollment_request_transition' USING ERRCODE = 'insufficient_privilege',
          CONSTRAINT = 'enrollment_request_transition';
      END IF;
      NEW.authority_seq := coalesce((SELECT max(r.seq) FROM app.enrollment_restrictions r
                                     WHERE r.organization_id = NEW.organization_id AND r.user_id = NEW.user_id), 0);
    ELSE
      ok := ok AND NEW.authority_seq IS NOT DISTINCT FROM OLD.authority_seq;
    END IF;
  ELSIF ok AND OLD.status = 'approved' AND NEW.status = 'revoked' THEN
    ok := OLD.consumed_at IS NULL AND NEW.consumed_at IS NULL AND NEW.consumed_binding_id IS NULL
      AND NEW.revoked_by_user_id = app.ctx_user() AND NEW.decided_at = OLD.decided_at AND NEW.decided_by_user_id = OLD.decided_by_user_id
      AND NEW.authority_seq IS NOT DISTINCT FROM OLD.authority_seq;
  ELSIF ok AND OLD.status = 'approved' AND NEW.status = 'approved' THEN
    ok := OLD.consumed_at IS NULL AND NEW.consumed_at IS NOT NULL AND NEW.consumed_binding_id IS NOT NULL
      AND NEW.decided_at = OLD.decided_at AND NEW.decided_by_user_id = OLD.decided_by_user_id
      AND NEW.revoked_at IS NULL AND NEW.revoked_by_user_id IS NULL
      AND NEW.authority_seq IS NOT DISTINCT FROM OLD.authority_seq;
  ELSE
    ok := false;
  END IF;
  IF NOT coalesce(ok, false) THEN
    RAISE EXCEPTION 'enrollment_request_transition' USING ERRCODE = 'insufficient_privilege',
      CONSTRAINT = 'enrollment_request_transition';
  END IF;
  RETURN NEW;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.enrollment_requests_immutable() FROM PUBLIC;
  CREATE TRIGGER enrollment_requests_immutable BEFORE UPDATE ON app.enrollment_requests
    FOR EACH ROW EXECUTE FUNCTION app.enrollment_requests_immutable();

  -- The request captures (Plan r3 §4.5): enrollment.requested on insert (the applicant, appended by the claim writer
  -- under events_insert_enrollment), enrollment.decided on approved/denied and enrollment.revoked on revoked (the
  -- deciding runtime session under events_insert); the consumption UPDATE is not captured (membership.granted is the
  -- admission's one success event). Mandatory atomic append: a refused row raises.
  EXECUTE $r$
CREATE FUNCTION audit.capture_enrollment_requests() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  INSERT INTO audit.events (action, outcome, organization_id, subject_user_id, resource_type, resource_id)
  VALUES (CASE WHEN TG_OP = 'INSERT' THEN 'enrollment.requested'
               WHEN NEW.status = 'revoked' THEN 'enrollment.revoked' ELSE 'enrollment.decided' END,
          'success', NEW.organization_id, NEW.user_id, 'enrollment_request', NEW.id);
  RETURN NULL;
END $$
$r$;
  REVOKE ALL ON FUNCTION audit.capture_enrollment_requests() FROM PUBLIC;
  CREATE TRIGGER audit_capture_request_insert AFTER INSERT ON app.enrollment_requests FOR EACH ROW
    EXECUTE FUNCTION audit.capture_enrollment_requests();
  CREATE TRIGGER audit_capture_request_status AFTER UPDATE OF status ON app.enrollment_requests FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status) EXECUTE FUNCTION audit.capture_enrollment_requests();

  -- Slice 1b: the enrollment policy (Plan r3 §1.5/§4.4/§4.5). The runtime may UPDATE the mode column only; the policy
  -- admits an Organization manager in its context Organization, and the guard below confines the mode change to the
  -- Organization owner (an administrator also passes organizations_update_slug, so the guard, not a policy, is the
  -- owner rule). The runtime's table-level SELECT (00003) already covers the mode and the two readiness columns.
  EXECUTE format('GRANT UPDATE (enrollment_mode) ON app.organizations TO %I', rt);
  EXECUTE format('CREATE POLICY organizations_enrollment_update ON app.organizations FOR UPDATE TO %I, %I USING (id = %s AND %s) '
    || 'WITH CHECK (id = %s AND %s)', rt, ow, O, MGR, O, MGR);

  -- The guard (owner-owned SECURITY INVOKER; Plan r3 §4.5 re-declared by addendum §A.2 and correction 01 §C.3): a
  -- readiness change only through the superuser operations channel inside the audited procedure (the transaction-local
  -- marker enrollment_readiness_set or replay-restrict); a mode change by a session role only by the Organization owner.
  -- The CHECK organizations_open_requires_resolved_history, not this guard, is the single open-requires-ready rule.
  EXECUTE $r$
CREATE FUNCTION app.organizations_enrollment_guard() RETURNS trigger
  LANGUAGE plpgsql SECURITY INVOKER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  su boolean := current_setting('is_superuser') = 'on';
BEGIN
  IF (NEW.enrollment_history_resolved_at IS DISTINCT FROM OLD.enrollment_history_resolved_at
      OR NEW.enrollment_history_basis IS DISTINCT FROM OLD.enrollment_history_basis)
     AND NOT (su AND coalesce(current_setting('app.enrollment_readiness_setter', true), '')
                     IN ('enrollment_readiness_set', 'replay-restrict')) THEN
    RAISE EXCEPTION 'enrollment_readiness_setter' USING ERRCODE = 'insufficient_privilege',
      HINT = 'readiness changes only through app.enrollment_readiness_set or app.enrollment_history_backfill';
  END IF;
  IF NEW.enrollment_mode IS DISTINCT FROM OLD.enrollment_mode AND NOT su
     AND app.rls_org_role() IS DISTINCT FROM 'org_owner' THEN
    RAISE EXCEPTION 'enrollment_policy_owner' USING ERRCODE = 'insufficient_privilege';
  END IF;
  RETURN NEW;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.organizations_enrollment_guard() FROM PUBLIC;
  CREATE TRIGGER organizations_enrollment_guard BEFORE UPDATE OF enrollment_mode, enrollment_history_resolved_at,
    enrollment_history_basis ON app.organizations FOR EACH ROW EXECUTE FUNCTION app.organizations_enrollment_guard();

  -- The policy/readiness capture (addendum §A.2): enrollment.policy_changed when the mode changed and/or
  -- enrollment.readiness_changed (reason: the basis, or history_unresolved on a close) when readiness changed; a close of
  -- an open Organization therefore appends both in its one transaction. Mandatory atomic append: a refused row raises.
  EXECUTE $r$
CREATE FUNCTION audit.capture_organization_enrollment() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF NEW.enrollment_mode IS DISTINCT FROM OLD.enrollment_mode THEN
    INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id)
    VALUES ('enrollment.policy_changed', 'success', NEW.id, 'organization', NEW.id);
  END IF;
  IF NEW.enrollment_history_resolved_at IS DISTINCT FROM OLD.enrollment_history_resolved_at
     OR NEW.enrollment_history_basis IS DISTINCT FROM OLD.enrollment_history_basis THEN
    INSERT INTO audit.events (action, outcome, reason, organization_id, resource_type, resource_id)
    VALUES ('enrollment.readiness_changed', 'success', coalesce(NEW.enrollment_history_basis, 'history_unresolved'),
            NEW.id, 'organization', NEW.id);
  END IF;
  RETURN NULL;
END $$
$r$;
  REVOKE ALL ON FUNCTION audit.capture_organization_enrollment() FROM PUBLIC;
  CREATE TRIGGER audit_capture_enrollment AFTER UPDATE OF enrollment_mode, enrollment_history_resolved_at,
    enrollment_history_basis ON app.organizations FOR EACH ROW
    WHEN (OLD.enrollment_mode IS DISTINCT FROM NEW.enrollment_mode
          OR OLD.enrollment_history_resolved_at IS DISTINCT FROM NEW.enrollment_history_resolved_at
          OR OLD.enrollment_history_basis IS DISTINCT FROM NEW.enrollment_history_basis)
    EXECUTE FUNCTION audit.capture_organization_enrollment();

  -- The backfill capture (addendum §A.2): only operator backfill rows; trigger rows are audited as membership.removed and
  -- replay rows carry G013's receipts. The operator reference travels in request_id (review R-1).
  EXECUTE $r$
CREATE FUNCTION audit.capture_enrollment_restrictions() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  INSERT INTO audit.events (action, outcome, reason, organization_id, subject_user_id, resource_type, resource_id)
  VALUES ('enrollment.restriction_backfilled', 'success', 'operator_backfilled', NEW.organization_id, NEW.user_id,
          'organization', NEW.organization_id);
  RETURN NULL;
END $$
$r$;
  REVOKE ALL ON FUNCTION audit.capture_enrollment_restrictions() FROM PUBLIC;
  CREATE TRIGGER audit_capture_backfill AFTER INSERT ON app.enrollment_restrictions FOR EACH ROW
    WHEN (NEW.recorded_via = 'backfill') EXECUTE FUNCTION audit.capture_enrollment_restrictions();

  -- The trusted readiness setter (addendum §A.2, correction 01 §C.1/§C.5): owner-owned SECURITY INVOKER, EXECUTE held by
  -- no session role; only a superuser session completes it (the owner is refused at step 1 and, bound by FORCE RLS, by
  -- the guard). Outcomes resolved | closed | unchanged | absent | no_backfill. Not STRICT: NULL arguments raise 22004.
  EXECUTE $r$
CREATE FUNCTION app.enrollment_readiness_set(p_org uuid, p_basis text, p_by text) RETURNS text
  LANGUAGE plpgsql VOLATILE SECURITY INVOKER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  cur_at timestamptz;
  cur_basis text;
BEGIN
  IF current_setting('is_superuser') IS DISTINCT FROM 'on' THEN
    RAISE EXCEPTION 'enrollment_readiness_setter' USING ERRCODE = 'insufficient_privilege';
  END IF;
  IF p_org IS NULL OR p_by IS NULL THEN
    RAISE EXCEPTION 'enrollment_readiness_argument' USING ERRCODE = 'null_value_not_allowed';
  END IF;
  IF p_basis IS NOT NULL AND p_basis NOT IN ('operator_verified_absent', 'operator_backfilled') THEN
    RAISE EXCEPTION 'enrollment_readiness_basis' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF (CASE WHEN p_basis = 'operator_backfilled'
           THEN NOT (p_by ~ '^l[0-9]{1,12}[.][A-Za-z0-9._-]{1,50}$' AND char_length(p_by) <= 64)
           ELSE NOT (p_by ~ '^[A-Za-z0-9._-]{8,64}$') END) THEN
    RAISE EXCEPTION 'enrollment_readiness_reference' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  -- (3) the Organization lock (Plan r3 §1.4): serializes with admissions, mode changes, departures and decisions.
  SELECT o.enrollment_history_resolved_at, o.enrollment_history_basis INTO cur_at, cur_basis
  FROM app.organizations o WHERE o.id = p_org FOR NO KEY UPDATE;
  IF NOT FOUND THEN
    RETURN 'absent';
  END IF;
  -- (4) idempotence: no write, no audit.
  IF (p_basis IS NULL AND cur_at IS NULL) OR (p_basis IS NOT NULL AND cur_basis IS NOT DISTINCT FROM p_basis) THEN
    RETURN 'unchanged';
  END IF;
  -- (5) a backfill attestation needs at least one departed row of this Organization recorded by the operator backfill or
  -- reconstructed by a replay (re-attestation after restoring a checkpoint taken before the attestation; Codex
  -- re-attestation correction). Provenance is read, never rewritten; trigger rows and other Organizations' rows do not count.
  IF p_basis = 'operator_backfilled' AND NOT EXISTS (SELECT 1 FROM app.enrollment_restrictions r
                                                     WHERE r.organization_id = p_org AND r.kind = 'departed'
                                                       AND r.recorded_via IN ('backfill', 'replay')) THEN
    RETURN 'no_backfill';
  END IF;
  -- (6)-(8) one audited UPDATE under the transaction-local marker; a close of an open Organization flips the mode in the
  -- same statement (the CHECK requires it).
  PERFORM set_config('audit.request_id', p_by, true);
  PERFORM set_config('app.enrollment_readiness_setter', 'enrollment_readiness_set', true);
  UPDATE app.organizations o
  SET enrollment_history_resolved_at = CASE WHEN p_basis IS NULL THEN NULL ELSE clock_timestamp() END,
      enrollment_history_basis = p_basis,
      enrollment_mode = CASE WHEN p_basis IS NULL AND o.enrollment_mode = 'open' THEN 'invite_only' ELSE o.enrollment_mode END
  WHERE o.id = p_org;
  PERFORM set_config('app.enrollment_readiness_setter', '', true);
  RETURN CASE WHEN p_basis IS NULL THEN 'closed' ELSE 'resolved' END;
END $$
$r$;
  -- The operator backfill (addendum §A.2, correction 01 §C.1/§C.2/§C.5): one reviewed legacy departure, after its G013
  -- membership.remove ledger entry (p_by = 'l<seq>.<ref>'). Outcomes inserted | present | member | absent | not_legacy.
  -- An inserted row is stamped under the Organization lock (newer than every existing approval/invitation) and closes
  -- the Organization's readiness in the same transaction (the operator re-attests once per batch).
  EXECUTE $r$
CREATE FUNCTION app.enrollment_history_backfill(p_org uuid, p_user uuid, p_by text) RETURNS text
  LANGUAGE plpgsql VOLATILE SECURITY INVOKER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  cur_basis text;
BEGIN
  IF current_setting('is_superuser') IS DISTINCT FROM 'on' THEN
    RAISE EXCEPTION 'enrollment_readiness_setter' USING ERRCODE = 'insufficient_privilege';
  END IF;
  IF p_org IS NULL OR p_user IS NULL OR p_by IS NULL THEN
    RAISE EXCEPTION 'enrollment_backfill_argument' USING ERRCODE = 'null_value_not_allowed';
  END IF;
  IF NOT (p_by ~ '^l[0-9]{1,12}[.][A-Za-z0-9._-]{1,50}$' AND char_length(p_by) <= 64) THEN
    RAISE EXCEPTION 'enrollment_backfill_reference' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  SELECT o.enrollment_history_basis INTO cur_basis FROM app.organizations o WHERE o.id = p_org FOR NO KEY UPDATE;
  IF NOT FOUND OR NOT EXISTS (SELECT 1 FROM app.users u WHERE u.id = p_user) THEN
    RETURN 'absent';
  END IF;
  IF cur_basis IS NOT DISTINCT FROM 'created_after_ledger' THEN
    RETURN 'not_legacy';
  END IF;
  IF EXISTS (SELECT 1 FROM app.role_bindings rb
             WHERE rb.subject_user_id = p_user AND rb.organization_id = p_org AND rb.scope_type = 'organization'
               AND rb.scope_id = p_org AND rb.source = 'membership') THEN
    RETURN 'member';
  END IF;
  IF EXISTS (SELECT 1 FROM app.enrollment_restrictions r
             WHERE r.organization_id = p_org AND r.user_id = p_user AND r.kind = 'departed' AND r.lifted_at IS NULL) THEN
    RETURN 'present';
  END IF;
  PERFORM set_config('audit.request_id', p_by, true);
  INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via)
  VALUES (p_org, p_user, 'departed', 'backfill');
  PERFORM set_config('app.enrollment_readiness_setter', 'enrollment_readiness_set', true);
  UPDATE app.organizations o
  SET enrollment_history_resolved_at = NULL, enrollment_history_basis = NULL,
      enrollment_mode = CASE WHEN o.enrollment_mode = 'open' THEN 'invite_only' ELSE o.enrollment_mode END
  WHERE o.id = p_org AND o.enrollment_history_resolved_at IS NOT NULL;
  PERFORM set_config('app.enrollment_readiness_setter', '', true);
  RETURN 'inserted';
END $$
$r$;
  REVOKE ALL ON FUNCTION app.enrollment_readiness_set(uuid, text, text), app.enrollment_history_backfill(uuid, uuid, text) FROM PUBLIC;
END
$do$;
-- +goose StatementEnd

-- Postconditions (CT-ENR-1 repeats and extends them).
-- +goose StatementBegin
DO $$
DECLARE
  ow regrole := current_user::regrole;
  rt regrole := (current_database() || '_runtime')::regrole;
  pr regrole := (current_database() || '_policy_reader')::regrole;
  ar regrole := (current_database() || '_auth_reader')::regrole;
  cw regrole := (current_database() || '_claim_writer')::regrole;
  actual text;
  n_resolved bigint;
  n_open bigint;
BEGIN
  -- Postcondition 1: the new functions and the D-15(i) grant (owner, DEFINER, volatility, configuration, ACL): the eight
  -- new DEFINER functions of 00016 (slice 1a's six, the request slice's enrollment_request and decision body).
  SELECT string_agg(p.proname || ':' || CASE p.proowner WHEN ow THEN 'ow' WHEN pr THEN 'pr' WHEN ar THEN 'ar' WHEN cw THEN 'cw'
                                                         ELSE 'other' END
           || ':' || p.prosecdef::text || ':' || p.provolatile::text || ':' || coalesce(array_to_string(p.proconfig, '|'), '-')
           || ':' || coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                          SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN ow THEN 'ow' WHEN pr THEN 'pr'
                                                WHEN ar THEN 'ar' WHEN cw THEN 'cw' ELSE 'other' END || '=' || a.privilege_type AS v
                          FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'), ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.oid IN ('app.rls_enrollment_mode(uuid)'::regprocedure, 'app.rls_invitation_pending(uuid, uuid)'::regprocedure,
    'app.auth_resolve_enrollment(bytea, timestamptz)'::regprocedure, 'app.enrollment_lock_organization(uuid)'::regprocedure,
    'app.enrollment_admit(bytea, uuid, text, uuid)'::regprocedure, 'app.enrollment_restrict_on_departure()'::regprocedure,
    'app.enrollment_request(bytea, uuid)'::regprocedure, 'app.enrollment_restrict_on_decision()'::regprocedure,
    'app.invitation_claim_organization(uuid, bytea, text, text, text)'::regprocedure);
  IF actual IS DISTINCT FROM 'auth_resolve_enrollment:ar:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'enrollment_admit:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'enrollment_lock_organization:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE;'
    || 'enrollment_request:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'enrollment_restrict_on_decision:ow:true:v:search_path=pg_catalog, pg_temp:-;'
    || 'enrollment_restrict_on_departure:ow:true:v:search_path=pg_catalog, pg_temp:-;'
    || 'invitation_claim_organization:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_enrollment_mode:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_invitation_pending:cw:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE' THEN
    RAISE EXCEPTION 'enrollment functions are %', actual;
  END IF;

  -- Postcondition 2: the claim writer owns exactly its seven bodies (G019's three + rls_invitation_pending,
  -- enrollment_lock_organization, enrollment_admit and enrollment_request).
  IF (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 7 THEN
    RAISE EXCEPTION 'the claim writer owns % functions, expected 7', (SELECT count(*) FROM pg_proc WHERE proowner = cw);
  END IF;

  -- Postcondition 3: the new policies (name, command, roles).
  SELECT string_agg(format('%s:%s:%s', p.polname, p.polcmd,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt THEN 'rt' WHEN ow THEN 'ow' WHEN pr THEN 'pr' WHEN ar THEN 'ar' WHEN cw THEN 'cw' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p WHERE p.polrelid IN ('app.enrollment_sessions'::regclass, 'app.enrollment_restrictions'::regclass,
                                        'app.enrollment_requests'::regclass)
     OR p.polname IN ('organizations_policy_reader', 'organizations_claim_select', 'organizations_claim_lock',
                      'organizations_enrollment_update', 'events_insert_enrollment');
  IF actual IS DISTINCT FROM 'enrollment_requests_claim_insert:a:cw;enrollment_requests_claim_select:r:cw;'
    || 'enrollment_requests_claim_update:w:cw;enrollment_requests_decide:w:ow+rt;enrollment_requests_select:r:ow+rt;'
    || 'enrollment_restrictions_claim_lift:w:cw;enrollment_restrictions_claim_select:r:cw;'
    || 'enrollment_restrictions_owner_insert:a:ow;enrollment_restrictions_select:r:ow+rt;'
    || 'enrollment_sessions_auth_reader:r:ar;enrollment_sessions_claim_select:r:cw;enrollment_sessions_claim_update:w:cw;'
    || 'enrollment_sessions_expire:d:ow+rt;enrollment_sessions_insert:a:ow+rt;enrollment_sessions_select:r:ow+rt;'
    || 'enrollment_sessions_update:w:ow+rt;events_insert_enrollment:a:cw;organizations_claim_lock:w:cw;organizations_claim_select:r:cw;'
    || 'organizations_enrollment_update:w:ow+rt;organizations_policy_reader:r:pr' THEN
    RAISE EXCEPTION 'enrollment policies are %', actual;
  END IF;

  -- Postcondition 3b (slice 1b, request slice): the two readiness functions are owner-owned SECURITY INVOKER with no
  -- grantee at all (PUBLIC revoked; no session role may EXECUTE them); the two guards and the three captures are
  -- owner-owned invoker bodies without grantees.
  SELECT string_agg(p.pronamespace::regnamespace::text || '.' || p.proname || ':' || (p.proowner = ow)::text || ':' || p.prosecdef::text || ':'
           || coalesce((SELECT string_agg(a.grantee::regrole::text, '+') FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner), '-'),
           ';' ORDER BY p.pronamespace::regnamespace::text, p.proname) INTO actual
  FROM pg_proc p WHERE p.oid IN ('app.enrollment_readiness_set(uuid, text, text)'::regprocedure,
    'app.enrollment_history_backfill(uuid, uuid, text)'::regprocedure, 'app.organizations_enrollment_guard()'::regprocedure,
    'audit.capture_organization_enrollment()'::regprocedure, 'audit.capture_enrollment_restrictions()'::regprocedure,
    'app.enrollment_requests_immutable()'::regprocedure, 'audit.capture_enrollment_requests()'::regprocedure);
  IF actual IS DISTINCT FROM 'app.enrollment_history_backfill:true:false:-;'
    || 'app.enrollment_readiness_set:true:false:-;app.enrollment_requests_immutable:true:false:-;'
    || 'app.organizations_enrollment_guard:true:false:-;audit.capture_enrollment_requests:true:false:-;'
    || 'audit.capture_enrollment_restrictions:true:false:-;audit.capture_organization_enrollment:true:false:-' THEN
    RAISE EXCEPTION 'enrollment invoker functions are %', actual;
  END IF;

  -- Postcondition 4: no session role inserts into or deletes from the ledger; the runtime cannot lock an Organization
  -- through the helper; the claim writer cannot rename an Organization nor read its name.
  IF has_table_privilege(rt, 'app.enrollment_restrictions', 'INSERT') OR has_table_privilege(cw, 'app.enrollment_restrictions', 'INSERT')
     OR has_table_privilege(pr, 'app.enrollment_restrictions', 'INSERT') OR has_table_privilege(ar, 'app.enrollment_restrictions', 'INSERT')
     OR has_any_column_privilege(rt, 'app.enrollment_restrictions', 'INSERT') OR has_any_column_privilege(cw, 'app.enrollment_restrictions', 'INSERT')
     OR has_table_privilege(rt, 'app.enrollment_restrictions', 'DELETE') OR has_table_privilege(cw, 'app.enrollment_restrictions', 'DELETE')
     OR has_table_privilege(pr, 'app.enrollment_restrictions', 'DELETE') OR has_table_privilege(ar, 'app.enrollment_restrictions', 'DELETE')
     OR has_function_privilege(rt, 'app.enrollment_lock_organization(uuid)', 'EXECUTE')
     OR has_function_privilege(cw, 'app.enrollment_restrict_on_departure()', 'EXECUTE')
     OR has_column_privilege(cw, 'app.organizations', 'name', 'SELECT') OR has_column_privilege(cw, 'app.organizations', 'name', 'UPDATE')
     OR has_column_privilege(rt, 'app.enrollment_sessions', 'admitted_at', 'INSERT')
     OR has_column_privilege(rt, 'app.enrollment_sessions', 'admitted_at', 'UPDATE')
     OR NOT has_function_privilege(rt, 'app.invitation_claim_organization(uuid, bytea, text, text, text)', 'EXECUTE')
     OR NOT has_column_privilege(rt, 'app.organizations', 'enrollment_mode', 'UPDATE')
     OR has_column_privilege(rt, 'app.organizations', 'enrollment_history_resolved_at', 'UPDATE')
     OR has_column_privilege(rt, 'app.organizations', 'enrollment_history_basis', 'UPDATE')
     OR has_function_privilege(rt, 'app.enrollment_readiness_set(uuid, text, text)', 'EXECUTE')
     OR has_function_privilege(cw, 'app.enrollment_history_backfill(uuid, uuid, text)', 'EXECUTE')
     -- Request slice: no runtime INSERT or DELETE on requests, no runtime privilege on the authority version or the
     -- consumption pair, no claim-writer privilege on the decision columns, and no session role may run the decision body.
     OR has_table_privilege(rt, 'app.enrollment_requests', 'INSERT') OR has_any_column_privilege(rt, 'app.enrollment_requests', 'INSERT')
     OR has_table_privilege(rt, 'app.enrollment_requests', 'DELETE') OR has_table_privilege(cw, 'app.enrollment_requests', 'DELETE')
     OR has_column_privilege(rt, 'app.enrollment_requests', 'authority_seq', 'UPDATE')
     OR has_column_privilege(cw, 'app.enrollment_requests', 'authority_seq', 'UPDATE')
     OR has_column_privilege(cw, 'app.enrollment_requests', 'authority_seq', 'INSERT')
     OR has_column_privilege(rt, 'app.enrollment_requests', 'consumed_at', 'UPDATE')
     OR has_column_privilege(cw, 'app.enrollment_requests', 'status', 'UPDATE')
     OR has_function_privilege(rt, 'app.enrollment_restrict_on_decision()', 'EXECUTE')
     OR has_function_privilege(cw, 'app.enrollment_restrict_on_decision()', 'EXECUTE') THEN
    RAISE EXCEPTION 'enrollment privilege boundary violated';
  END IF;

  -- Postcondition 5: the triggers of app.role_bindings.
  SELECT string_agg(tgname, ',' ORDER BY tgname) INTO actual FROM pg_trigger
  WHERE tgrelid = 'app.role_bindings'::regclass AND NOT tgisinternal;
  IF actual IS DISTINCT FROM 'audit_capture_delete,audit_capture_insert,audit_capture_update,role_bindings_authority_guard,'
    || 'role_bindings_enrollment_departure,role_bindings_keys_immutable' THEN
    RAISE EXCEPTION 'app.role_bindings triggers are %', actual;
  END IF;

  -- Postcondition 6 (addendum §A.3 step 6): nothing resolved and nothing open at 00016. The owner is bound by FORCE, so
  -- FORCE is lifted only inside this transaction for the count (a refusal rolls the lift back) and restored at once.
  ALTER TABLE app.organizations NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) FILTER (WHERE enrollment_history_resolved_at IS NOT NULL), count(*) FILTER (WHERE enrollment_mode = 'open')
    INTO n_resolved, n_open FROM app.organizations;
  ALTER TABLE app.organizations FORCE ROW LEVEL SECURITY;
  IF n_resolved <> 0 OR n_open <> 0 THEN
    RAISE EXCEPTION 'readiness postcondition: % resolved / % open organizations at 00016', n_resolved, n_open;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  rt text := current_database() || '_runtime';
  pr text := current_database() || '_policy_reader';
  ar text := current_database() || '_auth_reader';
  cw text := current_database() || '_claim_writer';
  ow text := current_user;
  n_ses bigint;
  n_res bigint;
  n_req bigint;
  n_lt bigint;
  n_aud bigint;
BEGIN
  -- History is never discarded: refuse while any enrollment credential, ledger row, access request, enrollment-shaped
  -- login transaction or audit row with an 00016 action/reason/operation exists. FORCE is lifted only inside this
  -- transaction; a refusal rolls the lift back with everything else.
  ALTER TABLE app.enrollment_sessions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.enrollment_restrictions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.enrollment_requests NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.login_transactions NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_ses FROM app.enrollment_sessions;
  SELECT count(*) INTO n_res FROM app.enrollment_restrictions;
  SELECT count(*) INTO n_req FROM app.enrollment_requests;
  SELECT count(*) INTO n_lt FROM app.login_transactions
  WHERE enrollment_intent OR invitation_id IS NOT NULL OR request_on_arrival OR return_path = '/join/status';
  SELECT count(*) INTO n_aud FROM audit.events
  WHERE reason IN ('enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                   'recipient_proof_unavailable', 'recipient_ineligible', 'session_invalid',
                   'operator_verified_absent', 'operator_backfilled', 'history_unresolved')
     OR operation IN ('enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request', 'enrollment.admit', 'enrollment.abandon',
                      'enrollment.policy.get', 'enrollment.policy.update', 'enrollment.requests.list',
                      'enrollment.requests.decide', 'enrollment.requests.revoke')
     OR action IN ('enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled',
                   'enrollment.requested', 'enrollment.decided', 'enrollment.revoked')
     OR resource_type = 'enrollment_request';
  IF n_ses > 0 OR n_res > 0 OR n_req > 0 OR n_lt > 0 OR n_aud > 0 THEN
    RAISE EXCEPTION 'refusing to roll back 00016: % enrollment sessions / % restrictions / % requests / % login transactions / % audit rows (history is never discarded)',
      n_ses, n_res, n_req, n_lt, n_aud;
  END IF;
  ALTER TABLE app.login_transactions FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  DROP TRIGGER role_bindings_enrollment_departure ON app.role_bindings;
  DROP FUNCTION app.enrollment_restrict_on_departure();
  -- Request slice: the decision body, the state machine, the captures, the applicant's audit policy.
  DROP TRIGGER enrollment_requests_restrict ON app.enrollment_requests;
  DROP FUNCTION app.enrollment_restrict_on_decision();
  DROP TRIGGER enrollment_requests_immutable ON app.enrollment_requests;
  DROP FUNCTION app.enrollment_requests_immutable();
  DROP TRIGGER audit_capture_request_insert ON app.enrollment_requests;
  DROP TRIGGER audit_capture_request_status ON app.enrollment_requests;
  DROP FUNCTION audit.capture_enrollment_requests();
  DROP POLICY events_insert_enrollment ON audit.events;
  -- Slice 1b: the readiness functions, the guard, the two captures and the runtime mode write.
  DROP FUNCTION app.enrollment_history_backfill(uuid, uuid, text);
  DROP FUNCTION app.enrollment_readiness_set(uuid, text, text);
  DROP TRIGGER audit_capture_backfill ON app.enrollment_restrictions;
  DROP FUNCTION audit.capture_enrollment_restrictions();
  DROP TRIGGER audit_capture_enrollment ON app.organizations;
  DROP FUNCTION audit.capture_organization_enrollment();
  DROP TRIGGER organizations_enrollment_guard ON app.organizations;
  DROP FUNCTION app.organizations_enrollment_guard();
  DROP POLICY organizations_enrollment_update ON app.organizations;
  EXECUTE format('REVOKE UPDATE (enrollment_mode) ON app.organizations FROM %I', rt);

  EXECUTE format('SET LOCAL ROLE %I', cw);
  EXECUTE format('REVOKE EXECUTE ON FUNCTION app.invitation_claim_organization(uuid, bytea, text, text, text) FROM %I', rt);
  DROP FUNCTION app.enrollment_admit(bytea, uuid, text, uuid);
  DROP FUNCTION app.enrollment_request(bytea, uuid);
  DROP FUNCTION app.enrollment_lock_organization(uuid);
  DROP FUNCTION app.rls_invitation_pending(uuid, uuid);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.auth_resolve_enrollment(bytea, timestamptz);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  DROP FUNCTION app.rls_enrollment_mode(uuid);
  EXECUTE format('SET LOCAL ROLE %I', ow);

  DROP TABLE app.enrollment_restrictions;
  DROP TABLE app.enrollment_requests;
  DROP TABLE app.enrollment_sessions;

  DROP POLICY organizations_policy_reader ON app.organizations;
  DROP POLICY organizations_claim_select ON app.organizations;
  DROP POLICY organizations_claim_lock ON app.organizations;
  EXECUTE format('REVOKE SELECT (id, enrollment_mode) ON app.organizations FROM %I', pr);
  EXECUTE format('REVOKE SELECT (id, enrollment_mode), UPDATE (enrollment_mode) ON app.organizations FROM %I', cw);
  ALTER TABLE app.organizations DROP CONSTRAINT organizations_open_requires_resolved_history,
    DROP CONSTRAINT organizations_enrollment_history_shape, DROP CONSTRAINT organizations_enrollment_history_basis_valid,
    DROP COLUMN enrollment_history_basis, DROP COLUMN enrollment_history_resolved_at, DROP COLUMN enrollment_mode;

  ALTER TABLE app.login_transactions DROP CONSTRAINT login_transactions_return_path_check,
    ADD CONSTRAINT login_transactions_return_path_check CHECK (return_path IN ('/', '/account')),
    DROP COLUMN request_on_arrival, DROP COLUMN invitation_id, DROP COLUMN enrollment_intent;
END
$do$;
-- +goose StatementEnd

-- The version-15 attribution body (00015), restored.
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
    'invitation.issued', 'invitation.revoked', 'invitation.consumed');
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
               THEN NEW.reason IN ('group_admin_profile_edit_enabled', 'group_admin_profile_edit_disabled')
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
    WHEN NEW.action = 'auth.sign_in' THEN
      plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.organization_id IS NULL AND NEW.group_id IS NULL
      AND NEW.resource_type IS NULL
      AND ((NEW.outcome = 'success' AND NEW.reason IN ('signed_in', 'session_rotated')
            AND NEW.subject_user_id IS NOT NULL AND NEW.subject_user_id = NEW.actor_user_id)
        OR (NEW.outcome = 'denied' AND NEW.reason IN ('transaction_expired', 'login_canceled', 'code_rejected',
            'impersonation_refused', 'mfa_unverified', 'stale_authentication', 'identity_conflict', 'user_disabled',
            'organization_not_member', 'host_changed'))
        OR (NEW.outcome = 'failed' AND NEW.reason IN ('provider_error', 'provider_timeout', 'identity_namespace_mismatch',
            'identity_unavailable')))
    WHEN NEW.action = 'access.denied' THEN
      NEW.outcome = 'denied' AND NEW.reason IN ('organization_not_permitted', 'group_not_selectable', 'target_not_found',
        'target_kind_mismatch', 'target_outside_context', 'role_not_permitted', 'no_linked_path', 'edit_policy_disabled',
        'not_creator', 'object_not_ready', 'capability_unknown', 'fact_inconsistent', 'resource_mismatch')
      AND NEW.operation IS NOT NULL AND NEW.ttl_seconds IS NULL AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.subject_user_id IS NULL
      AND (NEW.resource_type IS NULL OR (NEW.organization_id IS NOT NULL
           AND NEW.resource_type IN ('organization', 'group', 'business_profile', 'stored_object', 'organization_domain')))
      AND (NEW.reason IS DISTINCT FROM 'resource_mismatch'
           OR (NEW.operation = 'mcp.request' AND NEW.organization_id IS NULL AND NEW.resource_type IS NULL))
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

-- The version-15 catalog constraints (reasons as 00013, actions, operations and resource types as 00015).
ALTER TABLE audit.events DROP CONSTRAINT events_reason_valid, DROP CONSTRAINT events_operation_valid,
  DROP CONSTRAINT events_action_valid, DROP CONSTRAINT events_resource_type_valid;
ALTER TABLE audit.events
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain', 'invitation')),
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
                    'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                    'object.deletion_pending', 'revocation.unaudited')),
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
                    'updated', 'reset', 'logo_set', 'logo_removed')),
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
                    'invitation.get', 'invitation.accept', 'invitation.consume', 'retirement.eligibility'));
