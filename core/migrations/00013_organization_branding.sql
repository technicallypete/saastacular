-- Organization branding (G017 Plan r2 §4; R0 addendum; docs/foundation/WEB-SHELL.md, OBJECT-STORAGE.md, AUDIT.md).
--
-- One goose transaction:
--   - app.organization_branding: one optional row per Organization (display label, Light/Dark colour overrides, an
--     optional logo reference, an optimistic version). No foreign key into app.stored_objects: the guard validates a
--     changed reference at write time (same Organization, ready, PNG/JPEG, at most 512 KiB) and every reader joins the
--     live object row, so G009 deletion and G013 purge proceed unchanged and a deleted logo simply stops being shown.
--   - FORCE row-level security with the Organization owner/admin predicate of organization_domains (00011); runtime
--     SELECT, column INSERT/UPDATE, no DELETE.
--   - Audit: one trigger-owned action organization.branding_changed (reason updated, reset, logo_set or logo_removed;
--     no values), the five branding.* denial operations, and the version-13 events_attribute body.
--   - The auth reader (not BYPASSRLS) gets column grants and exact policies: every branding row (its content is the
--     intentionally public projection) and only those stored objects that are some Organization's current logo. The
--     SECURITY DEFINER reader app.public_branding(org), owned by the auth reader and executable by the runtime only,
--     returns the projection with the logo joined only while ready.
--
-- Locks: one CREATE TABLE, policy and grant DDL on app.stored_objects (brief), CREATE OR REPLACE of
-- audit.events_attribute and the audit.events constraint swap (one validating scan, as 00011). No data migration.
--
-- Down (development and test only) refuses while any branding row or branding history exists, then restores the
-- version-12 shapes.

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regclass('app.organization_branding') IS NOT NULL OR to_regprocedure('app.public_branding(uuid)') IS NOT NULL
     OR to_regprocedure('app.auth_resolve_session(bytea, timestamptz, text)') IS NULL
     OR to_regclass('app.organization_domains_hostname_owned') IS NULL THEN
    RAISE EXCEPTION 'version-12 precondition: branding objects already exist or the version-12 shapes are missing';
  END IF;
END
$$;
-- +goose StatementEnd

-- The override validator: at most the 15 token names, each a lowercase #rrggbb.
-- +goose StatementBegin
CREATE FUNCTION app.branding_overrides_valid(o jsonb) RETURNS boolean
  LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF o IS NULL OR jsonb_typeof(o) <> 'object' THEN
    RETURN false;
  END IF;
  RETURN NOT EXISTS (
    SELECT 1 FROM jsonb_each(o) e
    WHERE e.key NOT IN ('background', 'foreground', 'primary', 'primary-foreground', 'accent', 'muted-foreground', 'border',
                        'popover', 'popover-foreground', 'muted', 'secondary', 'secondary-foreground', 'destructive',
                        'input', 'ring')
       OR jsonb_typeof(e.value) <> 'string' OR (e.value #>> '{}') !~ '^#[0-9a-f]{6}$');
END $$;
-- +goose StatementEnd

CREATE TABLE app.organization_branding (
  organization_id uuid        NOT NULL,
  display_label   text        NULL CONSTRAINT organization_branding_label_valid CHECK (
                                char_length(display_label) BETWEEN 1 AND 80 AND display_label !~ '[[:cntrl:]]'
                                AND display_label IS NFC NORMALIZED),
  logo_object_id  uuid        NULL,
  light           jsonb       NOT NULL DEFAULT '{}' CONSTRAINT organization_branding_light_valid
                                CHECK (app.branding_overrides_valid(light)),
  dark            jsonb       NOT NULL DEFAULT '{}' CONSTRAINT organization_branding_dark_valid
                                CHECK (app.branding_overrides_valid(dark)),
  version         bigint      NOT NULL DEFAULT 1 CONSTRAINT organization_branding_version_valid CHECK (version >= 1),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  updated_by      uuid        NOT NULL,
  CONSTRAINT organization_branding_pkey PRIMARY KEY (organization_id),
  CONSTRAINT organization_branding_organization_fkey FOREIGN KEY (organization_id) REFERENCES app.organizations (id)
    ON DELETE RESTRICT,
  CONSTRAINT organization_branding_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app.users (id) ON DELETE RESTRICT
);

-- The guard (owner-owned SECURITY INVOKER, so the reference check runs under the writer's row-level security): the
-- Organization never changes, a new row starts at version 1, every update advances the version by one, and a new or
-- changed logo reference must name a ready PNG/JPEG object of at most 512 KiB in the same Organization.
-- +goose StatementBegin
CREATE FUNCTION app.organization_branding_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  IF TG_OP = 'INSERT' AND NEW.version <> 1 THEN
    RAISE EXCEPTION 'organization branding version refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_branding_guard';
  END IF;
  IF TG_OP = 'UPDATE' AND (NEW.organization_id IS DISTINCT FROM OLD.organization_id
                           OR NEW.version IS DISTINCT FROM OLD.version + 1) THEN
    RAISE EXCEPTION 'organization branding binding or version refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_branding_guard';
  END IF;
  IF NEW.logo_object_id IS NOT NULL AND (TG_OP = 'INSERT' OR NEW.logo_object_id IS DISTINCT FROM OLD.logo_object_id)
     AND NOT EXISTS (SELECT 1 FROM app.stored_objects s
                     WHERE s.id = NEW.logo_object_id AND s.organization_id = NEW.organization_id AND s.state = 'ready'
                       AND s.content_type IN ('image/png', 'image/jpeg') AND s.byte_size <= 524288) THEN
    RAISE EXCEPTION 'organization branding logo refused'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'organization_branding_guard';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

CREATE TRIGGER organization_branding_guard BEFORE INSERT OR UPDATE ON app.organization_branding FOR EACH ROW
  EXECUTE FUNCTION app.organization_branding_guard();

-- The capture function (owner-owned SECURITY INVOKER): one organization.branding_changed per write, strictly atomic,
-- its reason derived from the old and new rows (a logo-only change is logo_set/logo_removed, an update to the defaults
-- reset, anything else updated). No value is recorded.
-- +goose StatementBegin
CREATE FUNCTION audit.capture_organization_branding() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
DECLARE
  r text;
BEGIN
  IF TG_OP = 'INSERT' THEN
    r := CASE WHEN NEW.logo_object_id IS NOT NULL AND NEW.display_label IS NULL AND NEW.light = '{}' AND NEW.dark = '{}'
              THEN 'logo_set' ELSE 'updated' END;
  ELSIF NEW.logo_object_id IS DISTINCT FROM OLD.logo_object_id AND NEW.display_label IS NOT DISTINCT FROM OLD.display_label
        AND NEW.light = OLD.light AND NEW.dark = OLD.dark THEN
    r := CASE WHEN NEW.logo_object_id IS NULL THEN 'logo_removed' ELSE 'logo_set' END;
  ELSIF NEW.display_label IS NULL AND NEW.logo_object_id IS NULL AND NEW.light = '{}' AND NEW.dark = '{}' THEN
    r := 'reset';
  ELSE
    r := 'updated';
  END IF;
  INSERT INTO audit.events (action, outcome, reason, organization_id, resource_type, resource_id)
  VALUES ('organization.branding_changed', 'success', r, NEW.organization_id, 'organization', NEW.organization_id);
  RETURN NULL;
END $$;
-- +goose StatementEnd

CREATE TRIGGER audit_capture_branding AFTER INSERT OR UPDATE ON app.organization_branding FOR EACH ROW
  EXECUTE FUNCTION audit.capture_organization_branding();

-- The attribution and shape rules with organization.branding_changed (the version-11/12 body otherwise unchanged;
-- CREATE OR REPLACE keeps identity, owner, ACL and trigger).
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

-- The widened catalog constraints (the version-12 lists plus the G017 values).
ALTER TABLE audit.events DROP CONSTRAINT events_action_valid, DROP CONSTRAINT events_reason_valid,
  DROP CONSTRAINT events_operation_valid, DROP CONSTRAINT events_resource_type_valid;
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
                    'branding.get', 'branding.update', 'branding.reset', 'branding.set_logo', 'branding.remove_logo')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain'));

-- The public reader, created by its owner (the auth reader, through the session's SET chain).
-- +goose StatementBegin
DO $do$
DECLARE
  db text := current_database();
  ow text := current_user;
  rt text := db || '_runtime';
  ar text := db || '_auth_reader';
BEGIN
  EXECUTE format('GRANT SELECT (organization_id, display_label, logo_object_id, light, dark, version) '
              || 'ON app.organization_branding TO %I', ar);
  EXECUTE format('GRANT SELECT (id, organization_id, state, content_type, byte_size, sha256, storage_key) '
              || 'ON app.stored_objects TO %I', ar);
  EXECUTE format('GRANT CREATE ON SCHEMA app TO %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  EXECUTE $r$
CREATE FUNCTION app.public_branding(p_org uuid)
  RETURNS TABLE (revision bigint, display_label text, light jsonb, dark jsonb, logo_sha256 bytea, logo_content_type text,
                 logo_size bigint, logo_key text)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  RETURN QUERY
  SELECT b.version, b.display_label, b.light, b.dark, s.sha256, s.content_type, s.byte_size, s.storage_key
  FROM app.organization_branding b
  LEFT JOIN app.stored_objects s ON s.id = b.logo_object_id AND s.organization_id = b.organization_id AND s.state = 'ready'
  WHERE b.organization_id = p_org;
END $$
$r$;
  REVOKE ALL ON FUNCTION app.public_branding(uuid) FROM PUBLIC;
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.public_branding(uuid) TO %I', rt);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  EXECUTE format('REVOKE CREATE ON SCHEMA app FROM %I', ar);
END
$do$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
DECLARE
  db text := current_database();
  ow regrole := current_user::regrole;
  rt regrole := (current_database() || '_runtime')::regrole;
  ar regrole := (current_database() || '_auth_reader')::regrole;
  O text := '(SELECT app.ctx_org())';
  R text := '(SELECT app.rls_org_role())';
  actual text;
BEGIN
  REVOKE ALL ON FUNCTION app.branding_overrides_valid(jsonb), app.organization_branding_guard(),
    audit.capture_organization_branding() FROM PUBLIC;
  -- The two override checks call the validator as the writing role: the runtime executes it (pure, IMMUTABLE).
  EXECUTE format('GRANT EXECUTE ON FUNCTION app.branding_overrides_valid(jsonb) TO %I', rt::text);

  -- Runtime: rows read, inserted at version 1 and updated through the listed columns. No DELETE.
  EXECUTE format('GRANT SELECT ON app.organization_branding TO %I', rt::text);
  EXECUTE format('GRANT INSERT (organization_id, display_label, logo_object_id, light, dark, version, updated_by) '
              || 'ON app.organization_branding TO %I', rt::text);
  EXECUTE format('GRANT UPDATE (display_label, logo_object_id, light, dark, version, updated_at, updated_by) '
              || 'ON app.organization_branding TO %I', rt::text);

  ALTER TABLE app.organization_branding ENABLE ROW LEVEL SECURITY;
  ALTER TABLE app.organization_branding FORCE ROW LEVEL SECURITY;
  EXECUTE format('CREATE POLICY organization_branding_select ON app.organization_branding FOR SELECT TO %I, %I USING ('
    || 'organization_id = %s AND %s IN (''org_owner'', ''org_admin''))', rt::text, ow::text, O, R);
  EXECUTE format('CREATE POLICY organization_branding_insert ON app.organization_branding FOR INSERT TO %I, %I WITH CHECK ('
    || 'organization_id = %s AND %s IN (''org_owner'', ''org_admin'') AND version = 1)', rt::text, ow::text, O, R);
  EXECUTE format('CREATE POLICY organization_branding_update ON app.organization_branding FOR UPDATE TO %I, %I USING ('
    || 'organization_id = %s AND %s IN (''org_owner'', ''org_admin'')) WITH CHECK ('
    || 'organization_id = %s AND %s IN (''org_owner'', ''org_admin''))', rt::text, ow::text, O, R, O, R);
  -- The auth reader: every branding row (the intentionally public projection; the HTTP layer answers only for the
  -- Organization the request's host routes to) and only the stored objects that are some Organization's current logo.
  EXECUTE format('CREATE POLICY organization_branding_auth_reader ON app.organization_branding FOR SELECT TO %I USING (true)',
    ar::text);
  EXECUTE format('CREATE POLICY stored_objects_auth_reader ON app.stored_objects FOR SELECT TO %I USING (EXISTS ('
    || 'SELECT 1 FROM app.organization_branding b WHERE b.logo_object_id = stored_objects.id '
    || 'AND b.organization_id = stored_objects.organization_id))', ar::text);

  -- Postconditions: the version-13 deltas.
  SELECT string_agg(c.relname || CASE WHEN c.relforcerowsecurity THEN '+force' ELSE '' END, ',' ORDER BY c.relname) INTO actual
  FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND (c.relrowsecurity OR c.relforcerowsecurity);
  IF actual IS DISTINCT FROM 'business_profiles+force,group_business_profiles+force,groups+force,identity_principals+force,'
    || 'login_transactions+force,mcp_connections+force,organization_branding+force,organization_domains+force,'
    || 'organizations+force,role_bindings+force,sessions+force,stored_objects+force,users+force' THEN
    RAISE EXCEPTION 'app row-level security tables are %', actual;
  END IF;
  -- The 50 application relations (version 12's 48 plus organization_branding and its primary key) plus the migration
  -- ledger's table, identity sequence and primary key.
  SELECT count(*)::text INTO actual FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace;
  IF actual <> '53' THEN
    RAISE EXCEPTION 'app relation count is %, expected 53 (50 plus the migration ledger)', actual;
  END IF;
  SELECT string_agg(p.polname || ':' || p.polcmd::text, ',' ORDER BY p.polname) INTO actual
  FROM pg_policy p WHERE p.polrelid = 'app.organization_branding'::regclass OR p.polname = 'stored_objects_auth_reader';
  IF actual IS DISTINCT FROM 'organization_branding_auth_reader:r,organization_branding_insert:a,organization_branding_select:r,'
                             || 'organization_branding_update:w,stored_objects_auth_reader:r' THEN
    RAISE EXCEPTION 'branding policies are %', actual;
  END IF;
  SELECT string_agg(p.proname || ':' || CASE p.proowner WHEN ow THEN 'ow' WHEN ar THEN 'ar' ELSE 'other' END || ':'
           || p.prosecdef::text || ':' || coalesce(array_to_string(p.proconfig, '|'), '-'), ';' ORDER BY p.proname) INTO actual
  FROM pg_proc p WHERE p.oid IN ('app.branding_overrides_valid(jsonb)'::regprocedure, 'app.organization_branding_guard()'::regprocedure,
                                 'app.public_branding(uuid)'::regprocedure, 'audit.capture_organization_branding()'::regprocedure);
  IF actual IS DISTINCT FROM 'branding_overrides_valid:ow:false:search_path=pg_catalog, pg_temp;'
                             || 'capture_organization_branding:ow:false:search_path=pg_catalog, pg_temp;'
                             || 'organization_branding_guard:ow:false:search_path=pg_catalog, pg_temp;'
                             || 'public_branding:ar:true:search_path=pg_catalog, pg_temp' THEN
    RAISE EXCEPTION 'branding functions are %', actual;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $do$
DECLARE
  db text := current_database();
  ow text := current_user;
  ar text := db || '_auth_reader';
BEGIN
  -- Without FORCE the owner's reads are not filtered by row-level security. Branding state and history are never
  -- silently discarded: refuse while any exists (the refusal rolls back the NO FORCE with everything else).
  ALTER TABLE app.organization_branding NO FORCE ROW LEVEL SECURITY;
  ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY;
  IF EXISTS (SELECT 1 FROM app.organization_branding) THEN
    RAISE EXCEPTION 'refusing to roll back 00013: branding rows exist';
  END IF;
  IF EXISTS (SELECT 1 FROM audit.events WHERE action = 'organization.branding_changed'
               OR reason IN ('updated', 'reset', 'logo_set', 'logo_removed')
               OR operation IN ('branding.get', 'branding.update', 'branding.reset', 'branding.set_logo',
                                'branding.remove_logo')) THEN
    RAISE EXCEPTION 'refusing to roll back 00013: audit.events holds branding history (history is never discarded)';
  END IF;
  ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;

  DROP POLICY stored_objects_auth_reader ON app.stored_objects;
  EXECUTE format('REVOKE SELECT (id, organization_id, state, content_type, byte_size, sha256, storage_key) '
              || 'ON app.stored_objects FROM %I', ar);
  EXECUTE format('SET LOCAL ROLE %I', ar);
  DROP FUNCTION app.public_branding(uuid);
  EXECUTE format('SET LOCAL ROLE %I', ow);
  DROP TABLE app.organization_branding;
  DROP FUNCTION app.organization_branding_guard();
  DROP FUNCTION audit.capture_organization_branding();
  DROP FUNCTION app.branding_overrides_valid(jsonb);
END
$do$;
-- +goose StatementEnd

-- The version-12 attribution body (00011's, unchanged by 00012).
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
    'domain.state_changed', 'domain.checked');
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

-- The version-12 catalog constraints.
ALTER TABLE audit.events DROP CONSTRAINT events_action_valid, DROP CONSTRAINT events_reason_valid,
  DROP CONSTRAINT events_operation_valid, DROP CONSTRAINT events_resource_type_valid;
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
                    'withdrawn', 'withdraw_error', 'withdraw_unknown', 'organization_not_member', 'host_changed')),
  ADD CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (
                    'access.resolve', 'organization.get', 'organization.get_policy', 'organization.set_policy', 'group.get',
                    'group.rename', 'group.change_slug', 'business_profile.get', 'business_profile.update', 'object.put',
                    'object.get', 'object.metadata', 'object.sign', 'object.delete', 'audit.read', 'mcp.request',
                    'domain.list', 'domain.claim', 'domain.rotate', 'domain.verify', 'domain.recheck', 'domain.remove',
                    'domain.retry_withdrawal', 'organization.set_slug')),
  ADD CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization', 'group', 'role_binding', 'business_profile', 'user', 'session', 'stored_object',
                     'mcp_connection', 'organization_domain'));
