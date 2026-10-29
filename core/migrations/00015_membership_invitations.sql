-- Scoped invitations, membership and ownership administration (G019 Plan r1 §4 as amended by the r2 addendum §1/§3/§5/§6,
-- the r3 addendum C1, the r4 addendum and the current-target/correction-01 addenda; docs/foundation/TENANCY.md,
-- AUTHORIZATION.md, DATABASE-ISOLATION.md, AUDIT.md).
--
-- One goose transaction:
--   - Precondition: owner, the three NOLOGIN NOINHERIT helper roles (the G019 <db>_claim_writer included), the exact
--     version-14 shapes, and the legacy owner preflight (every Organization has an org_owner binding and every Group a
--     group_owner binding; counts only, never ids or names; inspected with FORCE lifted inside the transaction).
--   - app.invitations: immutable intent, one-use consumption, revocation, a 32-byte token digest; runtime SELECT without
--     token_digest, column-limited INSERT, UPDATE (revoked_at) only, no DELETE; triggers invitations_keys_immutable,
--     invitations_bound (100 pending unexpired per scope, lock then count) and audit.capture_invitations.
--   - role_bindings: runtime DELETE and SELECT (id); role_bindings_lock replaced by role_bindings_update and
--     role_bindings_delete; the SECURITY INVOKER guard app.role_bindings_authority_guard() (owner authority, promotion
--     eligibility, departure context and dependent grants, last owner after a separate scope-row lock command).
--   - Five <db>_policy_reader STABLE readers (owner count, dependent Group bindings, sole-owned scopes, the subject's own
--     Organizations for trusted retirement (Result addendum A1-01), Group invitation candidates) and three <db>_claim_writer bodies (invitation_claim_group, invitation_claim_organization,
--     rls_user_enabled_lock); invitation_claim_organization is executable by the owner role only (D-15 pending, G058).
--   - audit: three invitation.* actions, the invitation resource type, seventeen access.denied operations, the
--     events_attribute body with the invitation shape (00013 body otherwise unchanged) and events_insert_claim.
--
-- Locks: policy, grant and trigger DDL take brief ACCESS EXCLUSIVE locks on role_bindings, users, identity_principals
-- and audit.events; the audit constraint swap scans audit.events once. No data migration, no rewrite.
--
-- Down (development and test only) refuses while app.invitations holds any row or audit.events holds any invitation.*
-- action (inspected with FORCE lifted inside the transaction; a refusal rolls the lift back), then restores version 14.

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
  n_org bigint;
  n_grp bigint;
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

  -- The 00008/00011 helper checks, extended to the third helper.
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
  IF EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.roleid IN (pr, ar, cw) AND m.member <> ow
               AND (m.inherit_option OR m.set_option OR m.member::regrole::text LIKE db || '\_%')) THEN
    RAISE EXCEPTION 'a helper role has an unexpected member: run infra/postgres/provision-roles.sql';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member IN (rt, cw)) THEN
    RAISE EXCEPTION 'the runtime and claim-writer roles must not be members of any role';
  END IF;

  -- Version-14 shapes: the onboarding objects exist, none of the 00015 objects exist.
  IF to_regprocedure('app.rls_profile_eligible(uuid)') IS NULL OR to_regprocedure('app.groups_bootstrap_owner()') IS NULL
     OR to_regclass('app.invitations') IS NOT NULL
     OR to_regprocedure('app.role_bindings_authority_guard()') IS NOT NULL
     OR to_regprocedure('app.rls_owner_count(text, uuid)') IS NOT NULL
     OR to_regprocedure('app.rls_dependent_group_bindings(uuid)') IS NOT NULL
     OR to_regprocedure('app.rls_sole_owner_scopes()') IS NOT NULL
     OR to_regprocedure('app.rls_own_organizations(uuid, integer)') IS NOT NULL
     OR to_regprocedure('app.rls_group_invitation_candidates(uuid, uuid, integer)') IS NOT NULL
     OR to_regprocedure('app.rls_user_enabled_lock(uuid, text, uuid)') IS NOT NULL
     OR to_regprocedure('app.invitation_claim_group(uuid, bytea)') IS NOT NULL
     OR to_regprocedure('app.invitation_claim_organization(uuid, bytea, text, text, text)') IS NOT NULL
     OR to_regprocedure('audit.capture_invitations()') IS NOT NULL
     OR EXISTS (SELECT 1 FROM pg_policy WHERE polname IN ('role_bindings_update', 'role_bindings_delete',
                  'role_bindings_scope_reader', 'role_bindings_retirement_reader', 'users_policy_reader',
                  'role_bindings_claim_select', 'role_bindings_insert_invited', 'role_bindings_claim_upgrade',
                  'users_claim_select', 'users_claim_lock', 'identity_principals_claim_select', 'events_insert_claim')) THEN
    RAISE EXCEPTION 'version-14 precondition: membership objects already exist or the version-14 shapes are missing';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_policy WHERE polrelid = 'app.role_bindings'::regclass AND polname = 'role_bindings_lock'
                   AND polcmd = 'w' AND pg_get_expr(polwithcheck, polrelid) = 'false') THEN
    RAISE EXCEPTION 'version-14 precondition: role_bindings_lock (WITH CHECK false) is missing';
  END IF;
  IF has_table_privilege(rt, 'app.role_bindings', 'DELETE') OR has_column_privilege(rt, 'app.role_bindings', 'id', 'SELECT') THEN
    RAISE EXCEPTION 'version-14 precondition: runtime already holds DELETE or SELECT (id) on app.role_bindings';
  END IF;
  IF NOT has_function_privilege(rt, 'app.is_organization_member(uuid, uuid)', 'EXECUTE')
     OR has_function_privilege(ow, 'app.is_organization_member(uuid, uuid)', 'EXECUTE') THEN
    RAISE EXCEPTION 'version-14 precondition: app.is_organization_member EXECUTE is not runtime-only';
  END IF;
  -- The role_bindings captures run as the invoker (00008): the claim writer appends them under events_insert_claim.
  IF (SELECT prosecdef FROM pg_proc WHERE oid = 'audit.capture_role_bindings()'::regprocedure) THEN
    RAISE EXCEPTION 'version-14 precondition: audit.capture_role_bindings is not SECURITY INVOKER';
  END IF;

  -- Legacy owner preflight (Plan r1 §4.1/§4.7). The owner is bound by FORCE ROW LEVEL SECURITY, so without lifting it
  -- the count would read nothing; FORCE is lifted only inside this transaction and restored before any change (a
  -- refusal rolls the lift back with everything else). Counts only; an operator resolves each through the documented
  -- restricted procedure before re-running (no automatic promotion or deletion).
  ALTER TABLE app.organizations NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.groups NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.role_bindings NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_org FROM app.organizations o
  WHERE NOT EXISTS (SELECT 1 FROM app.role_bindings rb WHERE rb.organization_id = o.id AND rb.scope_type = 'organization'
                      AND rb.scope_id = o.id AND rb.role = 'org_owner' AND rb.source = 'membership');
  SELECT count(*) INTO n_grp FROM app.groups g
  WHERE NOT EXISTS (SELECT 1 FROM app.role_bindings rb WHERE rb.organization_id = g.organization_id AND rb.scope_type = 'group'
                      AND rb.scope_id = g.id AND rb.role = 'group_owner' AND rb.source = 'membership');
  IF n_org > 0 OR n_grp > 0 THEN
    RAISE EXCEPTION 'legacy owner preflight: % organizations / % groups without an owner binding (resolve through the restricted owner procedure, docs/foundation/RECOVERY.md, then re-run)',
      n_org, n_grp USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'legacy_owner_preflight';
  END IF;
  ALTER TABLE app.organizations FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.groups FORCE ROW LEVEL SECURITY;
  ALTER TABLE app.role_bindings FORCE ROW LEVEL SECURITY;
END
$$;
-- +goose StatementEnd

-- The invitation lifecycle (Plan r1 §4.2).
CREATE TABLE app.invitations (
  id                  uuid        NOT NULL,
  organization_id     uuid        NOT NULL,
  scope_type          text        NOT NULL,
  scope_id            uuid        NOT NULL,
  group_id            uuid,
  role                text        NOT NULL,
  issuer_user_id      uuid        NOT NULL,
  intended_user_id    uuid,
  intended_address    text,
  token_digest        bytea       NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  expires_at          timestamptz NOT NULL,
  revoked_at          timestamptz,
  consumed_at         timestamptz,
  consumed_by_user_id uuid,
  consumed_binding_id uuid,
  CONSTRAINT invitations_pkey PRIMARY KEY (id),
  CONSTRAINT invitations_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT invitations_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  -- MATCH SIMPLE: unchecked when group_id IS NULL (Organization scope).
  CONSTRAINT invitations_group_fkey FOREIGN KEY (organization_id, group_id)
    REFERENCES app.groups (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT invitations_issuer_fkey FOREIGN KEY (issuer_user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT invitations_intended_user_fkey FOREIGN KEY (intended_user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT invitations_consumed_by_fkey FOREIGN KEY (consumed_by_user_id) REFERENCES app.users (id) ON DELETE RESTRICT,
  CONSTRAINT invitations_token_digest_key UNIQUE (token_digest),
  CONSTRAINT invitations_token_digest_length CHECK (octet_length(token_digest) = 32),
  CONSTRAINT invitations_scope_consistent CHECK (
       (scope_type = 'organization' AND scope_id = organization_id AND group_id IS NULL
        AND role IN ('org_admin', 'org_member') AND intended_address IS NOT NULL)
    OR (scope_type = 'group' AND group_id IS NOT NULL AND scope_id = group_id
        AND role IN ('group_admin', 'group_member') AND intended_user_id IS NOT NULL AND intended_address IS NULL)),
  -- The local part is byte-exact (no case folding, NFC or rewriting); only the domain is pinned lowercase.
  CONSTRAINT invitations_address_normalized CHECK (intended_address IS NULL OR (
        octet_length(intended_address) BETWEEN 3 AND 254
    AND octet_length(split_part(intended_address, '@', 1)) <= 64
    AND intended_address ~ '^[^@[:space:][:cntrl:]]+@[^@[:space:][:cntrl:]]+$'
    AND split_part(intended_address, '@', 2) = lower(split_part(intended_address, '@', 2)))),
  CONSTRAINT invitations_lifetime CHECK (expires_at > created_at AND expires_at <= created_at + interval '30 days'),
  CONSTRAINT invitations_consumption_shape CHECK ((consumed_at IS NULL) = (consumed_by_user_id IS NULL)
    AND (consumed_at IS NULL OR revoked_at IS NULL) AND (consumed_at IS NOT NULL OR consumed_binding_id IS NULL)),
  CONSTRAINT invitations_consumed_is_intended CHECK (consumed_by_user_id IS NULL OR intended_user_id = consumed_by_user_id)
);
CREATE INDEX invitations_scope_idx ON app.invitations (organization_id, scope_type, scope_id, id);
CREATE INDEX invitations_pending_idx ON app.invitations (organization_id, scope_type, scope_id, expires_at)
  WHERE consumed_at IS NULL AND revoked_at IS NULL;
CREATE INDEX invitations_recipient_idx ON app.invitations (intended_user_id, id)
  WHERE consumed_at IS NULL AND revoked_at IS NULL;
ALTER TABLE app.invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.invitations FORCE ROW LEVEL SECURITY;

-- Audit capture for invitations (owner-owned SECURITY INVOKER, the audit.capture_* pattern; strict and atomic). Never
-- the digest or the address.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_invitations() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role)
    VALUES ('invitation.issued', 'success', NEW.organization_id, NEW.group_id, 'invitation', NEW.id, NEW.intended_user_id, NEW.role);
  ELSIF OLD.consumed_at IS NULL AND NEW.consumed_at IS NOT NULL THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role)
    VALUES ('invitation.consumed', 'success', NEW.organization_id, NEW.group_id, 'invitation', NEW.id, NEW.consumed_by_user_id,
            NEW.role);
  ELSIF OLD.revoked_at IS NULL AND NEW.revoked_at IS NOT NULL THEN
    INSERT INTO audit.events (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role)
    VALUES ('invitation.revoked', 'success', NEW.organization_id, NEW.group_id, 'invitation', NEW.id, NEW.intended_user_id, NEW.role);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION audit.capture_invitations() FROM PUBLIC;
CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.invitations FOR EACH ROW EXECUTE FUNCTION audit.capture_invitations();
CREATE TRIGGER audit_capture_update AFTER UPDATE ON app.invitations FOR EACH ROW EXECUTE FUNCTION audit.capture_invitations();

-- The attribution and shape rules with the invitation.* actions (the version-13 body otherwise unchanged; CREATE OR
-- REPLACE keeps identity, owner, ACL and trigger).
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

-- The widened catalog constraints: actions +3, resource types +1, operations +17 (36 -> 53); reasons unchanged (58).
ALTER TABLE audit.events DROP CONSTRAINT events_action_valid, DROP CONSTRAINT events_operation_valid,
  DROP CONSTRAINT events_resource_type_valid;
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
                    'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                    'object.deletion_pending', 'revocation.unaudited')),
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
                    'invitation.get', 'invitation.accept', 'invitation.consume', 'retirement.eligibility')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain', 'invitation'));

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
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  MGR text := '(SELECT app.rls_org_role()) IN (''org_owner'', ''org_admin'')';
  GMGR text := 'app.rls_group_role(group_id) IN (''group_owner'', ''group_admin'')';
  manage text;
BEGIN
  -- Immutability (owner-owned invoker; no early return for any role): runtime/owner may only set revoked_at once on an
  -- unconsumed row; the claim writer may only perform the consumption transition on an unrevoked row, claiming
  -- intended_user_id only for an address-addressed Organization invitation in that same update.
  EXECUTE format($f$
CREATE FUNCTION app.invitations_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $b$
DECLARE
  keys_same boolean;
BEGIN
  keys_same := (NEW.id, NEW.organization_id, NEW.scope_type, NEW.scope_id, NEW.group_id, NEW.role, NEW.issuer_user_id,
                NEW.intended_address, NEW.token_digest, NEW.created_at, NEW.expires_at)
               IS NOT DISTINCT FROM
               (OLD.id, OLD.organization_id, OLD.scope_type, OLD.scope_id, OLD.group_id, OLD.role, OLD.issuer_user_id,
                OLD.intended_address, OLD.token_digest, OLD.created_at, OLD.expires_at);
  IF keys_same AND current_user::text IN (%L, %L)
     AND OLD.revoked_at IS NULL AND NEW.revoked_at IS NOT NULL AND OLD.consumed_at IS NULL
     AND (NEW.intended_user_id, NEW.consumed_at, NEW.consumed_by_user_id, NEW.consumed_binding_id)
         IS NOT DISTINCT FROM (OLD.intended_user_id, OLD.consumed_at, OLD.consumed_by_user_id, OLD.consumed_binding_id) THEN
    RETURN NEW;
  END IF;
  IF keys_same AND current_user::text = %L
     AND OLD.revoked_at IS NULL AND NEW.revoked_at IS NULL
     AND OLD.consumed_at IS NULL AND OLD.consumed_by_user_id IS NULL AND OLD.consumed_binding_id IS NULL
     AND NEW.consumed_at IS NOT NULL AND NEW.consumed_by_user_id IS NOT NULL AND NEW.consumed_binding_id IS NOT NULL
     AND (NEW.intended_user_id IS NOT DISTINCT FROM OLD.intended_user_id
          OR (OLD.intended_user_id IS NULL AND OLD.scope_type = 'organization')) THEN
    RETURN NEW;
  END IF;
  RAISE EXCEPTION 'invitation columns are immutable'
    USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'invitations_keys_immutable';
END $b$
$f$, rt, ow, cw);
  REVOKE ALL ON FUNCTION app.invitations_keys_immutable() FROM PUBLIC;
  CREATE TRIGGER invitations_keys_immutable BEFORE UPDATE ON app.invitations
    FOR EACH ROW EXECUTE FUNCTION app.invitations_keys_immutable();

  -- Issuance bound (owner-owned invoker, VOLATILE; runtime/owner): the scope row lock is one command, the pending
  -- unexpired count a separate following command (fresh READ COMMITTED snapshot after any wait).
  EXECUTE format($f$
CREATE FUNCTION app.invitations_bound() RETURNS trigger
  LANGUAGE plpgsql VOLATILE SET search_path = pg_catalog AS $b$
DECLARE
  n bigint;
BEGIN
  IF current_user::text NOT IN (%L, %L) THEN
    RETURN NEW;
  END IF;
  IF NEW.scope_type = 'group' THEN
    PERFORM 1 FROM app.groups g WHERE g.id = NEW.group_id AND g.organization_id = NEW.organization_id FOR NO KEY UPDATE;
  ELSE
    PERFORM 1 FROM app.organizations o WHERE o.id = NEW.organization_id FOR NO KEY UPDATE;
  END IF;
  SELECT count(*) INTO n FROM app.invitations i
  WHERE i.organization_id = NEW.organization_id AND i.scope_type = NEW.scope_type AND i.scope_id = NEW.scope_id
    AND i.consumed_at IS NULL AND i.revoked_at IS NULL AND i.expires_at > now();
  IF n >= 100 THEN
    RAISE EXCEPTION 'pending invitation limit reached'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'invitation_limit';
  END IF;
  RETURN NEW;
END $b$
$f$, rt, ow);
  REVOKE ALL ON FUNCTION app.invitations_bound() FROM PUBLIC;
  CREATE TRIGGER invitations_bound BEFORE INSERT ON app.invitations FOR EACH ROW EXECUTE FUNCTION app.invitations_bound();

  -- Runtime grants.
  EXECUTE format('GRANT SELECT (id, organization_id, scope_type, scope_id, group_id, role, issuer_user_id, intended_user_id, '
              || 'intended_address, created_at, expires_at, revoked_at, consumed_at, consumed_by_user_id, consumed_binding_id) '
              || 'ON app.invitations TO %I', rt);
  EXECUTE format('GRANT INSERT (id, organization_id, scope_type, scope_id, group_id, role, issuer_user_id, intended_user_id, '
              || 'intended_address, token_digest, expires_at) ON app.invitations TO %I', rt);
  EXECUTE format('GRANT UPDATE (revoked_at) ON app.invitations TO %I', rt);
  EXECUTE format('GRANT DELETE ON app.role_bindings TO %I', rt);
  EXECUTE format('GRANT SELECT (id) ON app.role_bindings TO %I', rt);

  -- Policy-reader grants and helper policies (SELECT only).
  EXECUTE format('GRANT SELECT (id, display_name, status) ON app.users TO %I', pr);
  EXECUTE format('CREATE POLICY role_bindings_scope_reader ON app.role_bindings FOR SELECT TO %I USING ('
    || 'organization_id = app.ctx_org())', pr);
  EXECUTE format('CREATE POLICY role_bindings_retirement_reader ON app.role_bindings FOR SELECT TO %I USING ('
    || 'subject_user_id = app.ctx_user() OR role IN (''org_owner'', ''group_owner''))', pr);
  EXECUTE format('CREATE POLICY users_policy_reader ON app.users FOR SELECT TO %I USING (true)', pr);

  -- The is_organization_member reader (auth-reader owned): EXECUTE for the owner (the FORCE-bound owner reaches the
  -- guard and the invitation insert policy) and the claim writer (Group-claim recipient check).
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.is_organization_member(uuid, uuid) TO %I, %I', ow, cw);
  EXECUTE format('SET LOCAL ROLE %I', ow);

  -- The five policy-reader functions, created as the policy reader (00006 SET chain). Each answers only for a scope
  -- the caller can currently see, or only for the caller itself; otherwise 0 / no rows.
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', pr);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  EXECUTE $r$
CREATE FUNCTION app.rls_owner_count(p_scope_type text, p_scope_id uuid) RETURNS integer
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text := app.rls_org_role();
BEGIN
  IF app.ctx_user() IS NULL OR o IS NULL OR r IS NULL OR p_scope_id IS NULL THEN
    RETURN 0;
  END IF;
  IF p_scope_type = 'organization' THEN
    IF p_scope_id <> o THEN
      RETURN 0;
    END IF;
    RETURN (SELECT count(*) FROM app.role_bindings rb
            WHERE rb.organization_id = o AND rb.scope_type = 'organization' AND rb.scope_id = o
              AND rb.role = 'org_owner' AND rb.source = 'membership');
  ELSIF p_scope_type = 'group' THEN
    IF (g IS NOT NULL AND p_scope_id <> g) OR NOT (r IN ('org_owner', 'org_admin') OR app.rls_group_role(p_scope_id) IS NOT NULL) THEN
      RETURN 0;
    END IF;
    RETURN (SELECT count(*) FROM app.role_bindings rb
            WHERE rb.organization_id = o AND rb.scope_type = 'group' AND rb.scope_id = p_scope_id
              AND rb.role = 'group_owner' AND rb.source = 'membership');
  END IF;
  RETURN 0;
END $$
$r$;
  EXECUTE $r$
CREATE FUNCTION app.rls_dependent_group_bindings(p_subject uuid) RETURNS integer
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
BEGIN
  IF u IS NULL OR o IS NULL OR p_subject IS NULL
     OR NOT (p_subject = u OR app.rls_org_role() IN ('org_owner', 'org_admin')) THEN
    RETURN 0;
  END IF;
  RETURN (SELECT count(*) FROM app.role_bindings rb
          WHERE rb.organization_id = o AND rb.subject_user_id = p_subject AND rb.scope_type = 'group'
            AND rb.source = 'membership');
END $$
$r$;
  EXECUTE $r$
CREATE FUNCTION app.rls_sole_owner_scopes() RETURNS TABLE (organization_id uuid, scope_type text, scope_id uuid)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
BEGIN
  IF u IS NULL THEN
    RETURN;
  END IF;
  RETURN QUERY
  SELECT rb.organization_id, rb.scope_type, rb.scope_id FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.source = 'membership' AND rb.role IN ('org_owner', 'group_owner')
    AND (SELECT count(*) FROM app.role_bindings x
         WHERE x.organization_id = rb.organization_id AND x.scope_type = rb.scope_type AND x.scope_id = rb.scope_id
           AND x.source = 'membership' AND x.role IN ('org_owner', 'group_owner')) = 1
  ORDER BY rb.organization_id, rb.scope_type, rb.scope_id;
END $$
$r$;
  -- Trusted retirement enumeration (Result addendum A1-01, D-A1): only the context subject's own current Organization
  -- memberships, identifiers only, in ascending keyset pages of at most 100; nothing outside the IDENTITY shape.
  EXECUTE $r$
CREATE FUNCTION app.rls_own_organizations(p_after uuid, p_limit integer) RETURNS TABLE (organization_id uuid)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
BEGIN
  IF u IS NULL OR app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL
     OR p_limit IS NULL OR p_limit < 1 OR p_limit > 100 THEN
    RETURN;
  END IF;
  RETURN QUERY
  SELECT rb.organization_id FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.scope_type = 'organization' AND rb.scope_id = rb.organization_id
    AND rb.source = 'membership' AND (p_after IS NULL OR rb.organization_id > p_after)
  ORDER BY rb.organization_id
  LIMIT p_limit;
END $$
$r$;
  EXECUTE $r$
CREATE FUNCTION app.rls_group_invitation_candidates(p_group uuid, p_after uuid, p_limit integer)
  RETURNS TABLE (user_id uuid, display_name text)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  r text := app.rls_org_role();
BEGIN
  IF app.ctx_user() IS NULL OR o IS NULL OR r IS NULL OR p_group IS NULL OR p_limit IS NULL OR p_limit < 1
     OR (g IS NOT NULL AND g <> p_group)
     OR NOT (r IN ('org_owner', 'org_admin') OR app.rls_group_role(p_group) IN ('group_owner', 'group_admin'))
     OR NOT EXISTS (SELECT 1 FROM app.role_bindings x WHERE x.organization_id = o AND x.scope_type = 'group'
                      AND x.scope_id = p_group AND x.source = 'membership') THEN
    RETURN;
  END IF;
  RETURN QUERY
  SELECT us.id, us.display_name FROM app.role_bindings rb JOIN app.users us ON us.id = rb.subject_user_id
  WHERE rb.organization_id = o AND rb.scope_type = 'organization' AND rb.scope_id = o AND rb.source = 'membership'
    AND us.status = 'active' AND (p_after IS NULL OR us.id > p_after)
    AND NOT EXISTS (SELECT 1 FROM app.role_bindings gb WHERE gb.subject_user_id = rb.subject_user_id
                      AND gb.organization_id = o AND gb.scope_type = 'group' AND gb.scope_id = p_group)
  ORDER BY us.id
  LIMIT least(p_limit, 101);
END $$
$r$;
  REVOKE ALL ON FUNCTION app.rls_owner_count(text, uuid), app.rls_dependent_group_bindings(uuid), app.rls_sole_owner_scopes(),
    app.rls_own_organizations(uuid, integer), app.rls_group_invitation_candidates(uuid, uuid, integer) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_owner_count(text, uuid), app.rls_dependent_group_bindings(uuid), '
              || 'app.rls_sole_owner_scopes(), app.rls_own_organizations(uuid, integer), '
              || 'app.rls_group_invitation_candidates(uuid, uuid, integer) TO %I, %I', rt, ow);
  -- The claim writer evaluates MGR/GMGR (lock admission, users policies).
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_org_role(), app.rls_group_role(uuid) TO %I', cw);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', pr);

  -- Claim-writer authority (r3 C1 as corrected by r4): schema usage, the three context parsers, the exact columns,
  -- policies TO the claim writer only, and the trigger-issued audit captures of the three claim actions.
  EXECUTE format('GRANT USAGE ON SCHEMA app TO %I', cw);
  EXECUTE format('GRANT USAGE ON SCHEMA audit TO %I', cw);
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.ctx_user(), app.ctx_org(), app.ctx_group() TO %I', cw);
  EXECUTE format('GRANT SELECT ON app.invitations TO %I', cw);
  EXECUTE format('GRANT UPDATE (consumed_at, consumed_by_user_id, intended_user_id, consumed_binding_id) ON app.invitations TO %I', cw);
  EXECUTE format('GRANT SELECT (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source), '
              || 'INSERT (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source), '
              || 'UPDATE (role) ON app.role_bindings TO %I', cw);
  EXECUTE format('GRANT SELECT (id, status), UPDATE (status) ON app.users TO %I', cw);
  EXECUTE format('GRANT SELECT (issuer, subject, user_id) ON app.identity_principals TO %I', cw);
  EXECUTE format('GRANT INSERT (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role, '
              || 'previous_role) ON audit.events TO %I', cw);
  EXECUTE format('CREATE POLICY invitations_claim_select ON app.invitations FOR SELECT TO %I USING (true)', cw);
  EXECUTE format('CREATE POLICY invitations_claim_update ON app.invitations FOR UPDATE TO %I USING (true) '
    || 'WITH CHECK (consumed_by_user_id = app.ctx_user())', cw);
  EXECUTE format('CREATE POLICY role_bindings_claim_select ON app.role_bindings FOR SELECT TO %I USING (true)', cw);
  EXECUTE format('CREATE POLICY role_bindings_insert_invited ON app.role_bindings FOR INSERT TO %I WITH CHECK ('
    || 'subject_user_id = app.ctx_user() AND source = ''membership'' AND role NOT IN (''org_owner'', ''group_owner''))', cw);
  EXECUTE format('CREATE POLICY role_bindings_claim_upgrade ON app.role_bindings FOR UPDATE TO %I USING (true) WITH CHECK ('
    || 'subject_user_id = app.ctx_user() AND source = ''membership'' AND role NOT IN (''org_owner'', ''group_owner''))', cw);
  manage := 'id = app.ctx_user() OR (app.rls_org_role() IS NOT NULL AND EXISTS (SELECT 1 FROM app.role_bindings rb '
    || 'WHERE rb.subject_user_id = users.id AND rb.organization_id = app.ctx_org()))';
  EXECUTE format('CREATE POLICY users_claim_select ON app.users FOR SELECT TO %I USING (%s)', cw, manage);
  EXECUTE format('CREATE POLICY users_claim_lock ON app.users FOR UPDATE TO %I USING (%s) WITH CHECK (false)', cw, manage);
  EXECUTE format('CREATE POLICY identity_principals_claim_select ON app.identity_principals FOR SELECT TO %I USING (true)', cw);
  EXECUTE format('CREATE POLICY events_insert_claim ON audit.events FOR INSERT TO %I WITH CHECK ('
    || 'pg_trigger_depth() >= 1 AND actor_kind = ''user'' AND actor_user_id = app.ctx_user() AND organization_id IS NOT NULL '
    || 'AND action IN (''membership.granted'', ''membership.role_changed'', ''invitation.consumed''))', cw);

  -- The three claim-writer bodies, created as the claim writer (00013 SET LOCAL ROLE pattern).
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', cw);
  EXECUTE format('SET LOCAL ROLE %I', cw);
  -- Scoped user lock (r2 §5): admission first, with the invoker's context; refused -> NULL without taking any lock.
  EXECUTE $r$
CREATE FUNCTION app.rls_user_enabled_lock(p_user uuid, p_scope_type text, p_scope_id uuid) RETURNS boolean
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  admitted boolean := false;
  st text;
BEGIN
  IF u IS NULL OR p_user IS NULL THEN
    RETURN NULL;
  END IF;
  IF p_user = u THEN
    admitted := true;
  ELSIF o IS NOT NULL AND p_scope_id IS NOT NULL AND EXISTS (
          SELECT 1 FROM app.role_bindings rb WHERE rb.subject_user_id = p_user AND rb.organization_id = o
            AND rb.scope_type = p_scope_type AND rb.scope_id = p_scope_id AND rb.source = 'membership') THEN
    IF p_scope_type = 'organization' THEN
      admitted := p_scope_id = o AND coalesce(app.rls_org_role() IN ('org_owner', 'org_admin'), false);
    ELSIF p_scope_type = 'group' THEN
      admitted := (g IS NULL OR g = p_scope_id)
        AND coalesce(app.rls_org_role() IN ('org_owner', 'org_admin')
                     OR app.rls_group_role(p_scope_id) IN ('group_owner', 'group_admin'), false);
    END IF;
  END IF;
  IF NOT admitted THEN
    RETURN NULL;
  END IF;
  PERFORM 1 FROM app.users us WHERE us.id = p_user FOR SHARE;
  SELECT us.status INTO st FROM app.users us WHERE us.id = p_user;
  RETURN st = 'active';
END $$
$r$;
  -- Group consumption: the context subject's own Group invitation in the active Organization (and selected Group).
  EXECUTE $r$
CREATE FUNCTION app.invitation_claim_group(p_id uuid, p_digest bytea)
  RETURNS TABLE (outcome text, organization_id uuid, scope_type text, scope_id uuid, group_id uuid, role text, binding_id uuid)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  o uuid := app.ctx_org();
  g uuid := app.ctx_group();
  inv app.invitations%ROWTYPE;
  ex_id uuid;
  ex_role text;
  bid uuid;
  res text;
BEGIN
  IF u IS NULL OR o IS NULL THEN
    RAISE EXCEPTION 'invitation claim context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'invitation_context';
  END IF;
  -- (1) serialize duplicate claims on the invitation row.
  SELECT i.* INTO inv FROM app.invitations i WHERE i.id = p_id FOR NO KEY UPDATE;
  -- (2) digest, kind, tenant, selected scope and recipient intent; any mismatch is the one closed not_found.
  IF NOT FOUND OR p_digest IS NULL OR inv.token_digest <> p_digest OR inv.scope_type <> 'group'
     OR inv.organization_id <> o OR (g IS NOT NULL AND g <> inv.group_id) OR inv.intended_user_id IS DISTINCT FROM u THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF inv.consumed_at IS NOT NULL THEN
    -- Replay by the same recipient: already_member only with a current qualifying binding; never restored.
    IF inv.consumed_by_user_id = u THEN
      SELECT rb.id, rb.role INTO ex_id, ex_role FROM app.role_bindings rb
      WHERE rb.subject_user_id = u AND rb.organization_id = inv.organization_id AND rb.scope_type = inv.scope_type
        AND rb.scope_id = inv.scope_id AND rb.source = 'membership';
      IF ex_id IS NOT NULL AND (CASE ex_role WHEN 'group_owner' THEN 3 WHEN 'group_admin' THEN 2 ELSE 1 END)
                               >= (CASE inv.role WHEN 'group_admin' THEN 2 ELSE 1 END) THEN
        RETURN QUERY SELECT 'already_member'::text, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, ex_role, ex_id;
      ELSE
        RETURN QUERY SELECT 'consumed'::text, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, inv.role, NULL::uuid;
      END IF;
      RETURN;
    END IF;
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF inv.revoked_at IS NOT NULL OR inv.expires_at <= now() THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  -- (3) the enabled recipient under a lock consistent with DisableUser.
  IF NOT coalesce(app.rls_user_enabled_lock(u, inv.scope_type, inv.scope_id), false) THEN
    RAISE EXCEPTION 'invitation recipient ineligible'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
  END IF;
  -- (4) lock the issuer's and the recipient's bindings for the scope (ascending id), then check with the post-lock
  -- snapshot: the issuer's current authority and the recipient's current Organization membership.
  PERFORM 1 FROM app.role_bindings rb
  WHERE rb.organization_id = inv.organization_id AND rb.source = 'membership'
    AND rb.subject_user_id IN (inv.issuer_user_id, u)
    AND (rb.scope_type = 'organization' OR rb.scope_id = inv.scope_id)
  ORDER BY rb.id FOR SHARE;
  IF NOT EXISTS (SELECT 1 FROM app.role_bindings rb
                 WHERE rb.subject_user_id = inv.issuer_user_id AND rb.organization_id = inv.organization_id
                   AND rb.source = 'membership'
                   AND ((rb.scope_type = 'organization' AND rb.role IN ('org_owner', 'org_admin'))
                     OR (rb.scope_type = 'group' AND rb.scope_id = inv.scope_id AND rb.role IN ('group_owner', 'group_admin')))) THEN
    RAISE EXCEPTION 'invitation issuer authority lost'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'issuer_authority_lost';
  END IF;
  IF NOT app.is_organization_member(u, inv.organization_id) THEN
    RAISE EXCEPTION 'invitation recipient ineligible'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
  END IF;
  -- (5) existing scope binding: equal/higher -> already_member; lower -> upgrade; none -> insert from the intent.
  SELECT rb.id, rb.role INTO ex_id, ex_role FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.organization_id = inv.organization_id AND rb.scope_type = inv.scope_type
    AND rb.scope_id = inv.scope_id AND rb.source = 'membership';
  IF ex_id IS NOT NULL THEN
    IF (CASE ex_role WHEN 'group_owner' THEN 3 WHEN 'group_admin' THEN 2 ELSE 1 END)
       >= (CASE inv.role WHEN 'group_admin' THEN 2 ELSE 1 END) THEN
      res := 'already_member';
    ELSE
      UPDATE app.role_bindings rb SET role = inv.role WHERE rb.id = ex_id;
      res := 'upgraded'; ex_role := inv.role;
    END IF;
    bid := ex_id;
  ELSE
    bid := gen_random_uuid();
    INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
    VALUES (bid, inv.organization_id, u, inv.scope_type, inv.scope_id, inv.group_id, inv.role, 'membership');
    res := 'joined'; ex_role := inv.role;
  END IF;
  -- (6) consume.
  UPDATE app.invitations i SET consumed_at = now(), consumed_by_user_id = u, consumed_binding_id = bid WHERE i.id = p_id;
  RETURN QUERY SELECT res, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, ex_role, bid;
END $$
$r$;
  -- Organization consumption (the G058 contract; EXECUTE for the owner role only in G019, D-15 pending): IDENTITY
  -- shape; the canonical (issuer, subject) principal must resolve to the context subject, in SQL.
  EXECUTE $r$
CREATE FUNCTION app.invitation_claim_organization(p_id uuid, p_digest bytea, p_issuer text, p_subject text, p_address text)
  RETURNS TABLE (outcome text, organization_id uuid, scope_type text, scope_id uuid, group_id uuid, role text, binding_id uuid)
  LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
#variable_conflict use_column
DECLARE
  u uuid := app.ctx_user();
  inv app.invitations%ROWTYPE;
  principal uuid;
  ex_id uuid;
  ex_role text;
  bid uuid;
  res text;
BEGIN
  IF u IS NULL OR app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL THEN
    RAISE EXCEPTION 'invitation claim context refused' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'invitation_context';
  END IF;
  SELECT i.* INTO inv FROM app.invitations i WHERE i.id = p_id FOR NO KEY UPDATE;
  SELECT ip.user_id INTO principal FROM app.identity_principals ip WHERE ip.issuer = p_issuer AND ip.subject = p_subject;
  IF NOT FOUND OR inv.id IS NULL OR p_digest IS NULL OR inv.token_digest <> p_digest OR inv.scope_type <> 'organization'
     OR principal IS DISTINCT FROM u OR p_address IS NULL OR inv.intended_address IS DISTINCT FROM p_address
     OR (inv.intended_user_id IS NOT NULL AND inv.intended_user_id <> u) THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF inv.consumed_at IS NOT NULL THEN
    IF inv.consumed_by_user_id = u THEN
      SELECT rb.id, rb.role INTO ex_id, ex_role FROM app.role_bindings rb
      WHERE rb.subject_user_id = u AND rb.organization_id = inv.organization_id AND rb.scope_type = 'organization'
        AND rb.scope_id = inv.organization_id AND rb.source = 'membership';
      IF ex_id IS NOT NULL AND (CASE ex_role WHEN 'org_owner' THEN 3 WHEN 'org_admin' THEN 2 ELSE 1 END)
                               >= (CASE inv.role WHEN 'org_admin' THEN 2 ELSE 1 END) THEN
        RETURN QUERY SELECT 'already_member'::text, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, ex_role, ex_id;
      ELSE
        RETURN QUERY SELECT 'consumed'::text, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, inv.role, NULL::uuid;
      END IF;
      RETURN;
    END IF;
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF inv.revoked_at IS NOT NULL OR inv.expires_at <= now() THEN
    RETURN QUERY SELECT 'not_found'::text, NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::text, NULL::uuid;
    RETURN;
  END IF;
  IF NOT coalesce(app.rls_user_enabled_lock(u, inv.scope_type, inv.scope_id), false) THEN
    RAISE EXCEPTION 'invitation recipient ineligible'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
  END IF;
  PERFORM 1 FROM app.role_bindings rb
  WHERE rb.organization_id = inv.organization_id AND rb.source = 'membership' AND rb.scope_type = 'organization'
    AND rb.subject_user_id IN (inv.issuer_user_id, u)
  ORDER BY rb.id FOR SHARE;
  IF NOT EXISTS (SELECT 1 FROM app.role_bindings rb
                 WHERE rb.subject_user_id = inv.issuer_user_id AND rb.organization_id = inv.organization_id
                   AND rb.scope_type = 'organization' AND rb.source = 'membership' AND rb.role IN ('org_owner', 'org_admin')) THEN
    RAISE EXCEPTION 'invitation issuer authority lost'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'issuer_authority_lost';
  END IF;
  SELECT rb.id, rb.role INTO ex_id, ex_role FROM app.role_bindings rb
  WHERE rb.subject_user_id = u AND rb.organization_id = inv.organization_id AND rb.scope_type = 'organization'
    AND rb.scope_id = inv.organization_id AND rb.source = 'membership';
  IF ex_id IS NOT NULL THEN
    IF (CASE ex_role WHEN 'org_owner' THEN 3 WHEN 'org_admin' THEN 2 ELSE 1 END)
       >= (CASE inv.role WHEN 'org_admin' THEN 2 ELSE 1 END) THEN
      res := 'already_member';
    ELSE
      UPDATE app.role_bindings rb SET role = inv.role WHERE rb.id = ex_id;
      res := 'upgraded'; ex_role := inv.role;
    END IF;
    bid := ex_id;
  ELSE
    bid := gen_random_uuid();
    INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
    VALUES (bid, inv.organization_id, u, 'organization', inv.organization_id, NULL, inv.role, 'membership');
    res := 'joined'; ex_role := inv.role;
  END IF;
  UPDATE app.invitations i SET consumed_at = now(), consumed_by_user_id = u, intended_user_id = u, consumed_binding_id = bid
  WHERE i.id = p_id;
  RETURN QUERY SELECT res, inv.organization_id, inv.scope_type, inv.scope_id, inv.group_id, ex_role, bid;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.rls_user_enabled_lock(uuid, text, uuid), app.invitation_claim_group(uuid, bytea),
    app.invitation_claim_organization(uuid, bytea, text, text, text) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.rls_user_enabled_lock(uuid, text, uuid), app.invitation_claim_group(uuid, bytea) '
              || 'TO %I, %I', rt, ow);
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.invitation_claim_organization(uuid, bytea, text, text, text) TO %I', ow);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', cw);

  -- Runtime/owner invitation policies (FORCE binds the owner to the same contexts).
  manage := format('((scope_type = ''organization'' AND %s) OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) '
    || 'AND (%s OR %s)))', MGR, G, G, MGR, GMGR);
  EXECUTE format('CREATE POLICY invitations_select ON app.invitations FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND (%s OR intended_user_id = %s OR consumed_by_user_id = %s))', rt, ow, O, manage, U, U);
  EXECUTE format('CREATE POLICY invitations_insert ON app.invitations FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND issuer_user_id = %s AND ((scope_type = ''organization'' AND %s IS NULL AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR %s) '
    || 'AND app.is_organization_member(intended_user_id, %s))))', rt, ow, O, U, G, MGR, G, G, MGR, GMGR, O);
  EXECUTE format('CREATE POLICY invitations_update ON app.invitations FOR UPDATE TO %I, %I USING ('
    || 'organization_id = %s AND %s) WITH CHECK (organization_id = %s AND %s)', rt, ow, O, manage, O, manage);

  -- role_bindings: the lock-only policy becomes a real UPDATE policy; DELETE is admitted under the same USING. The
  -- own-row branch keeps 00006's role_bindings_lock predicate so ReadFacts' FOR SHARE on own bindings is unchanged.
  -- Owner-role transitions, eligibility and last-owner protection are the guard's (a policy cannot see OLD).
  DROP POLICY role_bindings_lock ON app.role_bindings;
  manage := format('organization_id = %s AND source = ''membership'' AND ('
    || '(subject_user_id = %s AND %s IS NOT NULL AND (scope_type = ''organization'' OR %s IS NULL OR group_id = %s)) '
    || 'OR (scope_type = ''organization'' AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR %s)))', O, U, R, G, G, MGR, G, G, MGR, GMGR);
  EXECUTE format('CREATE POLICY role_bindings_update ON app.role_bindings FOR UPDATE TO %I, %I USING (%s) WITH CHECK ('
    || 'organization_id = %s AND source = ''membership'' AND ((scope_type = ''organization'' AND %s) '
    || 'OR (scope_type = ''group'' AND (%s IS NULL OR group_id = %s) AND (%s OR %s))))', rt, ow, manage, O, MGR, G, G, MGR, GMGR);
  EXECUTE format('CREATE POLICY role_bindings_delete ON app.role_bindings FOR DELETE TO %I, %I USING (%s)', rt, ow, manage);

  -- The authority and cardinality guard (SECURITY INVOKER, owner-owned, VOLATILE; each step a separate command with its
  -- own READ COMMITTED snapshot). Other roles (the superuser repair path, the seed, the recovery replay) return early.
  EXECUTE format($f$
CREATE FUNCTION app.role_bindings_authority_guard() RETURNS trigger
  LANGUAGE plpgsql VOLATILE SET search_path = pg_catalog AS $b$
DECLARE
  old_owner boolean;
  new_owner boolean;
BEGIN
  IF current_user::text NOT IN (%L, %L, %L) THEN
    IF TG_OP = 'DELETE' THEN
      RETURN OLD;
    END IF;
    RETURN NEW;
  END IF;
  old_owner := OLD.role IN ('org_owner', 'group_owner');
  new_owner := TG_OP = 'UPDATE' AND NEW.role IN ('org_owner', 'group_owner');
  -- (1) owner authority: no Organization-role override of Group ownership.
  IF old_owner OR new_owner THEN
    IF OLD.scope_type = 'organization' THEN
      IF app.rls_org_role() IS DISTINCT FROM 'org_owner' THEN
        RAISE EXCEPTION 'owner authority required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'owner_authority';
      END IF;
    ELSIF app.rls_group_role(OLD.group_id) IS DISTINCT FROM 'group_owner' THEN
      RAISE EXCEPTION 'owner authority required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'owner_authority';
    END IF;
  END IF;
  -- (2) promotion eligibility: same-scope admin, enabled under a lock consistent with DisableUser, and (Group) a
  -- current Organization member.
  IF NOT old_owner AND new_owner THEN
    IF OLD.role IS DISTINCT FROM (CASE OLD.scope_type WHEN 'organization' THEN 'org_admin' ELSE 'group_admin' END)
       OR NOT coalesce(app.rls_user_enabled_lock(OLD.subject_user_id, OLD.scope_type, OLD.scope_id), false)
       OR (OLD.scope_type = 'group' AND NOT app.is_organization_member(OLD.subject_user_id, OLD.organization_id)) THEN
      RAISE EXCEPTION 'promotion recipient ineligible'
        USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'recipient_ineligible';
    END IF;
  END IF;
  -- (3) Organization departure: an Organization-only context, and no dependent Group binding (the definer count sees
  -- every sibling Group regardless of the selected context).
  IF TG_OP = 'DELETE' AND OLD.scope_type = 'organization' THEN
    IF app.ctx_group() IS NOT NULL THEN
      RAISE EXCEPTION 'organization-only context required' USING ERRCODE = 'insufficient_privilege', CONSTRAINT = 'context_required';
    END IF;
    IF app.rls_dependent_group_bindings(OLD.subject_user_id) > 0 THEN
      RAISE EXCEPTION 'dependent group bindings remain'
        USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'dependent_bindings';
    END IF;
  END IF;
  -- (4) last owner: command A locks the scope row; command B counts with the snapshot taken after A's wait.
  IF old_owner AND NOT new_owner THEN
    IF OLD.scope_type = 'organization' THEN
      PERFORM 1 FROM app.organizations o WHERE o.id = OLD.scope_id FOR NO KEY UPDATE;
    ELSE
      PERFORM 1 FROM app.groups g WHERE g.id = OLD.scope_id FOR NO KEY UPDATE;
    END IF;
    IF app.rls_owner_count(OLD.scope_type, OLD.scope_id) - 1 < 1 THEN
      RAISE EXCEPTION 'last owner' USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'last_owner';
    END IF;
  END IF;
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END $b$
$f$, rt, ow, cw);
  REVOKE ALL ON FUNCTION app.role_bindings_authority_guard() FROM PUBLIC;
  CREATE TRIGGER role_bindings_authority_guard BEFORE UPDATE OF role OR DELETE ON app.role_bindings
    FOR EACH ROW EXECUTE FUNCTION app.role_bindings_authority_guard();
END
$do$;
-- +goose StatementEnd

-- Postconditions (CT-MEM-1 repeats and extends them).
-- +goose StatementBegin
DO $$
DECLARE
  ow regrole := current_user::regrole;
  rt regrole := (current_database() || '_runtime')::regrole;
  pr regrole := (current_database() || '_policy_reader')::regrole;
  ar regrole := (current_database() || '_auth_reader')::regrole;
  cw regrole := (current_database() || '_claim_writer')::regrole;
  actual text;
BEGIN
  -- Postcondition 1: the new and replaced functions (owner, SECURITY DEFINER, volatility, configuration, ACL).
  SELECT string_agg(p.proname || ':' || CASE p.proowner WHEN ow THEN 'ow' WHEN pr THEN 'pr' WHEN cw THEN 'cw' ELSE 'other' END
           || ':' || p.prosecdef::text || ':' || p.provolatile::text || ':' || coalesce(array_to_string(p.proconfig, '|'), '-')
           || ':' || coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
                          SELECT CASE a.grantee WHEN 0 THEN 'public' WHEN rt THEN 'rt' WHEN ow THEN 'ow' WHEN cw THEN 'cw'
                                                ELSE 'other' END || '=' || a.privilege_type AS v
                          FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'), ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.oid IN ('app.invitations_keys_immutable()'::regprocedure, 'app.invitations_bound()'::regprocedure,
    'audit.capture_invitations()'::regprocedure, 'app.role_bindings_authority_guard()'::regprocedure,
    'app.rls_owner_count(text, uuid)'::regprocedure, 'app.rls_dependent_group_bindings(uuid)'::regprocedure,
    'app.rls_sole_owner_scopes()'::regprocedure, 'app.rls_own_organizations(uuid, integer)'::regprocedure,
    'app.rls_group_invitation_candidates(uuid, uuid, integer)'::regprocedure,
    'app.rls_user_enabled_lock(uuid, text, uuid)'::regprocedure, 'app.invitation_claim_group(uuid, bytea)'::regprocedure,
    'app.invitation_claim_organization(uuid, bytea, text, text, text)'::regprocedure);
  IF actual IS DISTINCT FROM 'capture_invitations:ow:false:v:search_path=pg_catalog, pg_temp:-;'
    || 'invitation_claim_group:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'invitation_claim_organization:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE;'
    || 'invitations_bound:ow:false:v:search_path=pg_catalog:-;'
    || 'invitations_keys_immutable:ow:false:v:search_path=pg_catalog:-;'
    || 'rls_dependent_group_bindings:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_group_invitation_candidates:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_own_organizations:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_owner_count:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_sole_owner_scopes:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'rls_user_enabled_lock:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;'
    || 'role_bindings_authority_guard:ow:false:v:search_path=pg_catalog:-' THEN
    RAISE EXCEPTION 'membership functions are %', actual;
  END IF;

  -- Postcondition 2: the claim writer owns exactly its three bodies (a fourth fails the catalog).
  IF (SELECT count(*) FROM pg_proc WHERE proowner = cw) <> 3 THEN
    RAISE EXCEPTION 'the claim writer owns % functions, expected 3', (SELECT count(*) FROM pg_proc WHERE proowner = cw);
  END IF;

  -- Postcondition 3: the new policies (name, command, roles) and the dropped lock policy.
  SELECT string_agg(format('%s:%s:%s', p.polname, p.polcmd,
           (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
              SELECT CASE x WHEN rt THEN 'rt' WHEN ow THEN 'ow' WHEN pr THEN 'pr' WHEN cw THEN 'cw' ELSE 'other' END AS v
              FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname) INTO actual
  FROM pg_policy p WHERE p.polname IN ('invitations_select', 'invitations_insert', 'invitations_update',
    'invitations_claim_select', 'invitations_claim_update', 'role_bindings_update', 'role_bindings_delete',
    'role_bindings_scope_reader', 'role_bindings_retirement_reader', 'role_bindings_claim_select',
    'role_bindings_insert_invited', 'role_bindings_claim_upgrade', 'users_policy_reader', 'users_claim_select',
    'users_claim_lock', 'identity_principals_claim_select', 'events_insert_claim', 'role_bindings_lock');
  IF actual IS DISTINCT FROM 'events_insert_claim:a:cw;identity_principals_claim_select:r:cw;'
    || 'invitations_claim_select:r:cw;invitations_claim_update:w:cw;invitations_insert:a:ow+rt;'
    || 'invitations_select:r:ow+rt;invitations_update:w:ow+rt;role_bindings_claim_select:r:cw;'
    || 'role_bindings_claim_upgrade:w:cw;role_bindings_delete:d:ow+rt;role_bindings_insert_invited:a:cw;'
    || 'role_bindings_retirement_reader:r:pr;role_bindings_scope_reader:r:pr;role_bindings_update:w:ow+rt;'
    || 'users_claim_lock:w:cw;users_claim_select:r:cw;users_policy_reader:r:pr' THEN
    RAISE EXCEPTION 'membership policies are %', actual;
  END IF;

  -- Postcondition 4: runtime never reads the digest nor writes the consumption columns; the claim writer has no DELETE.
  IF has_column_privilege(rt, 'app.invitations', 'token_digest', 'SELECT')
     OR has_table_privilege(rt, 'app.invitations', 'DELETE')
     OR has_column_privilege(rt, 'app.invitations', 'consumed_at', 'UPDATE')
     OR has_column_privilege(rt, 'app.invitations', 'consumed_by_user_id', 'UPDATE')
     OR has_column_privilege(rt, 'app.invitations', 'intended_user_id', 'UPDATE')
     OR has_function_privilege(rt, 'app.invitation_claim_organization(uuid, bytea, text, text, text)', 'EXECUTE')
     OR has_table_privilege(cw, 'app.role_bindings', 'DELETE') OR has_table_privilege(cw, 'audit.events', 'SELECT')
     OR NOT has_table_privilege(rt, 'app.role_bindings', 'DELETE') THEN
    RAISE EXCEPTION 'membership privilege boundary violated';
  END IF;

  -- Postcondition 5: the triggers of app.invitations and app.role_bindings.
  SELECT string_agg(tgname, ',' ORDER BY tgname) INTO actual FROM pg_trigger
  WHERE tgrelid = 'app.invitations'::regclass AND NOT tgisinternal;
  IF actual IS DISTINCT FROM 'audit_capture_insert,audit_capture_update,invitations_bound,invitations_keys_immutable' THEN
    RAISE EXCEPTION 'app.invitations triggers are %', actual;
  END IF;
  SELECT string_agg(tgname, ',' ORDER BY tgname) INTO actual FROM pg_trigger
  WHERE tgrelid = 'app.role_bindings'::regclass AND NOT tgisinternal;
  IF actual IS DISTINCT FROM 'audit_capture_delete,audit_capture_insert,audit_capture_update,role_bindings_authority_guard,'
    || 'role_bindings_keys_immutable' THEN
    RAISE EXCEPTION 'app.role_bindings triggers are %', actual;
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
  U text := '(SELECT app.ctx_user())';
  O text := '(SELECT app.ctx_org())';
  G text := '(SELECT app.ctx_group())';
  R text := '(SELECT app.rls_org_role())';
  n_inv bigint;
  n_aud bigint;
BEGIN
  -- History is never discarded: refuse while any invitation (any tenant) or invitation.* audit row exists. FORCE is
  -- lifted only inside this transaction; a refusal rolls the lift back with everything else.
  ALTER TABLE app.invitations NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  SELECT count(*) INTO n_inv FROM app.invitations;
  SELECT count(*) INTO n_aud FROM audit.events
  WHERE action IN ('invitation.issued', 'invitation.revoked', 'invitation.consumed')
     OR operation IN ('membership.list', 'membership.self', 'membership.change_role', 'membership.remove', 'membership.leave',
                      'ownership.promote', 'ownership.demote', 'ownership.transfer', 'invitation.list', 'invitation.mine',
                      'invitation.candidates', 'invitation.issue', 'invitation.revoke', 'invitation.get', 'invitation.accept',
                      'invitation.consume', 'retirement.eligibility');
  IF n_inv > 0 OR n_aud > 0 THEN
    RAISE EXCEPTION 'refusing to roll back 00015: % invitations / % audit rows (history is never discarded)', n_inv, n_aud;
  END IF;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  DROP TRIGGER role_bindings_authority_guard ON app.role_bindings;
  DROP FUNCTION app.role_bindings_authority_guard();
  DROP POLICY role_bindings_update ON app.role_bindings;
  DROP POLICY role_bindings_delete ON app.role_bindings;
  EXECUTE format('CREATE POLICY role_bindings_lock ON app.role_bindings FOR UPDATE TO %I, %I USING ('
    || '%s IS NOT NULL AND organization_id = %s AND subject_user_id = %s AND source = ''membership'' AND '
    || '(scope_type = ''organization'' OR %s IS NULL OR group_id = %s)) WITH CHECK (false)',
    rt, ow, R, O, U, G, G);
  EXECUTE format('REVOKE DELETE ON app.role_bindings FROM %I', rt);
  EXECUTE format('REVOKE SELECT (id) ON app.role_bindings FROM %I', rt);

  DROP POLICY role_bindings_claim_select ON app.role_bindings;
  DROP POLICY role_bindings_insert_invited ON app.role_bindings;
  DROP POLICY role_bindings_claim_upgrade ON app.role_bindings;
  DROP POLICY role_bindings_scope_reader ON app.role_bindings;
  DROP POLICY role_bindings_retirement_reader ON app.role_bindings;
  DROP POLICY users_policy_reader ON app.users;
  DROP POLICY users_claim_select ON app.users;
  DROP POLICY users_claim_lock ON app.users;
  DROP POLICY identity_principals_claim_select ON app.identity_principals;
  DROP POLICY events_insert_claim ON audit.events;

  EXECUTE format('SET LOCAL ROLE %I', cw);
  DROP FUNCTION app.invitation_claim_organization(uuid, bytea, text, text, text);
  DROP FUNCTION app.invitation_claim_group(uuid, bytea);
  DROP FUNCTION app.rls_user_enabled_lock(uuid, text, uuid);
  EXECUTE format('SET LOCAL ROLE %I', pr);
  DROP FUNCTION app.rls_group_invitation_candidates(uuid, uuid, integer);
  DROP FUNCTION app.rls_own_organizations(uuid, integer);
  DROP FUNCTION app.rls_sole_owner_scopes();
  DROP FUNCTION app.rls_dependent_group_bindings(uuid);
  DROP FUNCTION app.rls_owner_count(text, uuid);
  EXECUTE format('REVOKE EXECUTE ON FUNCTION app.rls_org_role(), app.rls_group_role(uuid) FROM %I', cw);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE format('REVOKE EXECUTE ON FUNCTION app.is_organization_member(uuid, uuid) FROM %I, %I', ow, cw);
  EXECUTE format('SET LOCAL ROLE %I', ow);

  DROP TABLE app.invitations;
  DROP FUNCTION app.invitations_keys_immutable();
  DROP FUNCTION app.invitations_bound();
  DROP FUNCTION audit.capture_invitations();

  EXECUTE format('REVOKE SELECT (id, display_name, status) ON app.users FROM %I', pr);
  EXECUTE format('REVOKE SELECT (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source), '
              || 'INSERT (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source), '
              || 'UPDATE (role) ON app.role_bindings FROM %I', cw);
  EXECUTE format('REVOKE SELECT (id, status), UPDATE (status) ON app.users FROM %I', cw);
  EXECUTE format('REVOKE SELECT (issuer, subject, user_id) ON app.identity_principals FROM %I', cw);
  EXECUTE format('REVOKE INSERT (action, outcome, organization_id, group_id, resource_type, resource_id, subject_user_id, role, '
              || 'previous_role) ON audit.events FROM %I', cw);
  EXECUTE format('REVOKE EXECUTE ON FUNCTION app.ctx_user(), app.ctx_org(), app.ctx_group() FROM %I', cw);
  EXECUTE format('REVOKE USAGE ON SCHEMA app FROM %I', cw);
  EXECUTE format('REVOKE USAGE ON SCHEMA audit FROM %I', cw);
END
$do$;
-- +goose StatementEnd

-- The version-13 attribution body (00013), restored.
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
    'domain.state_changed', 'domain.checked', 'organization.branding_changed');
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

-- The version-14 catalog constraints.
ALTER TABLE audit.events DROP CONSTRAINT events_action_valid, DROP CONSTRAINT events_operation_valid,
  DROP CONSTRAINT events_resource_type_valid;
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
                    'auth.sign_in', 'access.denied', 'object.url_issued', 'object.upload_failed',
                    'object.deletion_pending', 'revocation.unaudited')),
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug',
                    'branding.get', 'branding.update', 'branding.reset', 'branding.set_logo', 'branding.remove_logo',
                    'group.suggest_slug', 'group.create', 'group.list', 'group.onboarding.get',
                    'group.onboarding.complete_new', 'group.onboarding.complete_existing', 'business_profile.list')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain'));
