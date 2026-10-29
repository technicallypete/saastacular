-- G013 ledger sync read (infra/recovery/recovery.sh ledger sync; Plan r3 §5.2): the currently restrictive states the
-- tool records, one line each: "kind<TAB>refs". Read only, as the cluster superuser (row-level security does not apply).
SELECT 'object.delete' || E'\t' || format('organization=%s,object=%s,deleter=%s', organization_id, id, deleted_by)
FROM app.stored_objects WHERE state IN ('deleting', 'deleted') AND deleted_by IS NOT NULL ORDER BY id;
SELECT 'user.disable' || E'\t' || format('user=%s,at=%s', id, to_char(disabled_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'))
FROM app.users WHERE status = 'disabled' ORDER BY id;
SELECT 'policy.disable' || E'\t' || format('organization=%s', id)
FROM app.organizations WHERE NOT group_admins_edit_business_profiles ORDER BY id;
