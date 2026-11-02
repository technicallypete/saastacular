-- G013 ledger sync read (infra/recovery/recovery.sh ledger sync; Plan r3 §5.2): the currently restrictive states the
-- tool records, one line each: "kind<TAB>refs". Read only, as the cluster superuser (row-level security does not apply).
-- G074 (Plan r3.1 §3.12, C10): a withdrawn published avatar (deleting or deleted, ready_at kept by the guard) is the
-- terminal avatar.delete restriction; a compensated upload that never became ready is not a restriction.
SELECT 'object.delete' || E'\t' || format('organization=%s,object=%s,deleter=%s', organization_id, id, deleted_by)
FROM app.stored_objects WHERE state IN ('deleting', 'deleted') AND deleted_by IS NOT NULL ORDER BY id;
SELECT 'user.disable' || E'\t' || format('user=%s,at=%s', id, to_char(disabled_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'))
FROM app.users WHERE status = 'disabled' ORDER BY id;
SELECT 'policy.disable' || E'\t' || format('organization=%s', id)
FROM app.organizations WHERE NOT group_admins_edit_business_profiles ORDER BY id;
SELECT 'avatar.delete' || E'\t' || format('user=%s,object=%s', user_id, id)
FROM app.account_avatars WHERE state IN ('deleting', 'deleted') AND ready_at IS NOT NULL ORDER BY id;
