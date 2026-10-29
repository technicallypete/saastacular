-- G013 replay of the strict restriction kinds (Plan r3 §5.2): policy.disable, membership.remove, membership.downgrade and
-- profile.unlink, each with its primitive's exact predicate, in one transaction (an audit fault rolls the whole entry
-- back: M1). Variables :'kind', :'org', :'subject', :'scope_type', :'scope_id', :'role', :'from_role', :'to_role',
-- :'grp', :'profile' (unused ones are "-"). Output: "updated<TAB>n" and "state<TAB>..." for the idempotence decision, then
-- one "ledger<TAB>..." line (G058, below) that recovery.sh's kv() ignores: its updated/state decision is unchanged.
BEGIN;
WITH p AS (UPDATE app.organizations SET group_admins_edit_business_profiles = false
           WHERE :'kind' = 'policy.disable' AND id = nullif(:'org', '-')::uuid AND group_admins_edit_business_profiles RETURNING 1),
     r AS (DELETE FROM app.role_bindings
           WHERE :'kind' = 'membership.remove' AND organization_id = nullif(:'org', '-')::uuid AND subject_user_id = nullif(:'subject', '-')::uuid
             AND scope_type = :'scope_type' AND scope_id = nullif(:'scope_id', '-')::uuid AND source = 'membership' AND role = :'role'
           RETURNING 1),
     d AS (UPDATE app.role_bindings SET role = :'to_role'
           WHERE :'kind' = 'membership.downgrade' AND organization_id = nullif(:'org', '-')::uuid AND subject_user_id = nullif(:'subject', '-')::uuid
             AND scope_type = :'scope_type' AND scope_id = nullif(:'scope_id', '-')::uuid AND source = 'membership' AND role = :'from_role'
           RETURNING 1),
     u AS (DELETE FROM app.group_business_profiles
           WHERE :'kind' = 'profile.unlink' AND organization_id = nullif(:'org', '-')::uuid AND group_id = nullif(:'grp', '-')::uuid
             AND business_profile_id = nullif(:'profile', '-')::uuid
           RETURNING 1)
SELECT 'updated' || E'\t' || ((SELECT count(*) FROM p) + (SELECT count(*) FROM r) + (SELECT count(*) FROM d) + (SELECT count(*) FROM u));
-- "unlinked" and "none" (already applied) only when every referenced entity exists in the restored checkpoint and the
-- link or binding is absent; a wrong or missing reference is "absent", which replays as a conflict (S-review F2).
SELECT 'state' || E'\t' || CASE :'kind'
  WHEN 'policy.disable' THEN coalesce((SELECT CASE WHEN group_admins_edit_business_profiles THEN 'enabled' ELSE 'disabled' END
                                       FROM app.organizations WHERE id = nullif(:'org', '-')::uuid), 'absent')
  WHEN 'profile.unlink' THEN CASE
    WHEN NOT EXISTS (SELECT 1 FROM app.groups WHERE organization_id = nullif(:'org', '-')::uuid AND id = nullif(:'grp', '-')::uuid)
      OR NOT EXISTS (SELECT 1 FROM app.business_profiles WHERE organization_id = nullif(:'org', '-')::uuid AND id = nullif(:'profile', '-')::uuid)
      THEN 'absent'
    WHEN EXISTS (SELECT 1 FROM app.group_business_profiles WHERE organization_id = nullif(:'org', '-')::uuid
                   AND group_id = nullif(:'grp', '-')::uuid AND business_profile_id = nullif(:'profile', '-')::uuid) THEN 'linked'
    ELSE 'unlinked' END
  ELSE coalesce((SELECT role FROM app.role_bindings WHERE organization_id = nullif(:'org', '-')::uuid
                   AND subject_user_id = nullif(:'subject', '-')::uuid AND scope_type = :'scope_type'
                   AND scope_id = nullif(:'scope_id', '-')::uuid),
                CASE WHEN EXISTS (SELECT 1 FROM app.organizations WHERE id = nullif(:'org', '-')::uuid)
                      AND EXISTS (SELECT 1 FROM app.users WHERE id = nullif(:'subject', '-')::uuid)
                      AND ((:'scope_type' = 'organization' AND nullif(:'scope_id', '-')::uuid = nullif(:'org', '-')::uuid)
                           OR (:'scope_type' = 'group' AND EXISTS (SELECT 1 FROM app.groups
                                 WHERE organization_id = nullif(:'org', '-')::uuid AND id = nullif(:'scope_id', '-')::uuid)))
                     THEN 'none' ELSE 'absent' END) END;
-- G058 enrollment ledger (Plan r3 §1.4; selected-design correction 01 §C.3), same transaction, after the statements above:
-- an Organization-scope membership.remove whose subject no longer holds an Organization-scope binding is recorded as an
-- unlifted departed restriction (recorded_via 'replay', no creator, created_at = this replay's clock, so newer than every
-- restored approval or invitation), unless one is already present; when the Organization exists but the restriction
-- cannot be materialized (a binding remains, e.g. a role mismatch, the subject user is absent or the scope id is not the
-- Organization), that Organization's open-signup readiness is closed instead (open -> invite_only) under the
-- transaction-local marker replay-restrict, which the readiness guard accepts from a superuser only. psql variables reach
-- the DO block through transaction-local settings (psql does not interpolate inside dollar quotes). Output values:
--   materialized  the restriction was recorded by this replay
--   present       an unlifted departed restriction exists (restored, or recorded by the delete's trigger above)
--   closed        the restriction cannot be materialized; readiness is closed (idempotent: already closed stays closed)
--   absent        the restored checkpoint predates the ledger (pre-00016; the mandatory migrate closes readiness)
--   unavailable   no Organization-scope departure to record (another kind, Group scope or the Organization is absent)
-- Nothing is inferred from audit history and nothing is written for the other kinds.
SELECT set_config('replay.kind', :'kind', true), set_config('replay.org', :'org', true), set_config('replay.subject', :'subject', true),
       set_config('replay.scope_type', :'scope_type', true), set_config('replay.scope_id', :'scope_id', true) \g /dev/null
DO $$
DECLARE
  o uuid;
  u uuid;
  outcome text := 'unavailable';
BEGIN
  IF to_regclass('app.enrollment_restrictions') IS NULL THEN
    outcome := 'absent';
  ELSIF current_setting('replay.kind') = 'membership.remove' AND current_setting('replay.scope_type') = 'organization' THEN
    o := nullif(current_setting('replay.org'), '-')::uuid;
    u := nullif(current_setting('replay.subject'), '-')::uuid;
    -- The Organization lock (Plan r3 §1.4 chronology rule), as the departure body and the operator functions take it.
    PERFORM 1 FROM app.organizations WHERE id = o FOR NO KEY UPDATE;
    IF FOUND THEN
      IF nullif(current_setting('replay.scope_id'), '-')::uuid IS DISTINCT FROM o
         OR NOT EXISTS (SELECT 1 FROM app.users WHERE id = u)
         OR EXISTS (SELECT 1 FROM app.role_bindings WHERE organization_id = o AND subject_user_id = u
                      AND scope_type = 'organization' AND scope_id = o) THEN
        PERFORM set_config('app.enrollment_readiness_setter', 'replay-restrict', true);
        UPDATE app.organizations
        SET enrollment_history_resolved_at = NULL, enrollment_history_basis = NULL,
            enrollment_mode = CASE WHEN enrollment_mode = 'open' THEN 'invite_only' ELSE enrollment_mode END
        WHERE id = o AND enrollment_history_resolved_at IS NOT NULL;
        PERFORM set_config('app.enrollment_readiness_setter', '', true);
        outcome := 'closed';
      ELSIF EXISTS (SELECT 1 FROM app.enrollment_restrictions
                    WHERE organization_id = o AND user_id = u AND kind = 'departed' AND lifted_at IS NULL) THEN
        outcome := 'present';
      ELSE
        INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via) VALUES (o, u, 'departed', 'replay');
        outcome := 'materialized';
      END IF;
    END IF;
  END IF;
  PERFORM set_config('replay.ledger', outcome, true);
END
$$;
SELECT 'ledger' || E'\t' || current_setting('replay.ledger');
COMMIT;
