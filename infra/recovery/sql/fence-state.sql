-- G013 fence state (infra/recovery/recovery.sh): the facts a fence records and its window checks read. Variables:
-- :'holder' the holder's application_name. Output lines "key<TAB>value". Read only, as the cluster superuser.
SELECT 'datacl' || E'\t' || coalesce(datacl::text, '-') FROM pg_database WHERE datname = current_database();
SELECT 'datacl_norm' || E'\t' || coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE "C") FROM (SELECT format('%s/%s/%s/%s', CASE a.grantee WHEN 0 THEN 'public' ELSE a.grantee::regrole::text END,
  a.grantor::regrole, a.privilege_type, a.is_grantable) AS v FROM pg_database d, aclexplode(d.datacl) a WHERE d.datname = current_database()) q), '-');
-- Non-superuser CONNECT grantees (the fence revokes exactly these).
SELECT 'grantees' || E'\t' || coalesce(string_agg(DISTINCT a.grantee::regrole::text, ',' ORDER BY a.grantee::regrole::text), '-')
FROM pg_database d, aclexplode(d.datacl) a JOIN pg_roles r ON r.oid = a.grantee
WHERE d.datname = current_database() AND a.privilege_type = 'CONNECT' AND NOT r.rolsuper;
-- PUBLIC's CONNECT (a NULL ACL is the type default, which grants it) and the login roles that can still connect after
-- the revoke, directly, through PUBLIC or through a role they inherit (S-review F6).
SELECT 'public_connect' || E'\t' || count(*) FROM pg_database d, aclexplode(coalesce(d.datacl, acldefault('d', d.datdba))) a
WHERE d.datname = current_database() AND a.grantee = 0 AND a.privilege_type = 'CONNECT';
SELECT 'connectable_logins' || E'\t' || count(*) FROM pg_roles r
WHERE r.rolcanlogin AND NOT r.rolsuper AND has_database_privilege(r.oid, current_database(), 'CONNECT');
SELECT 'nonsuper_backends' || E'\t' || count(*) FROM pg_stat_activity s JOIN pg_roles r ON r.oid = s.usesysid
WHERE s.datname = current_database() AND s.backend_type = 'client backend' AND NOT r.rolsuper;
-- Other superuser client backends: not this session and not the fence's holder (a visible window contradiction).
SELECT 'other_superusers' || E'\t' || count(*) FROM pg_stat_activity s JOIN pg_roles r ON r.oid = s.usesysid
WHERE s.backend_type = 'client backend' AND r.rolsuper AND s.pid <> pg_backend_pid() AND coalesce(s.application_name, '') <> :'holder';
SELECT 'holder' || E'\t' || coalesce((SELECT s.pid || ' ' || to_char(s.backend_start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"') || ' ' ||
  coalesce(to_char(s.xact_start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'), '-')
  FROM pg_stat_activity s WHERE s.application_name = :'holder' ORDER BY s.backend_start LIMIT 1), '-');
SELECT 'holder_locks' || E'\t' || count(*) FROM pg_locks l JOIN pg_stat_activity s ON s.pid = l.pid
JOIN pg_class c ON c.oid = l.relation
WHERE s.application_name = :'holder' AND l.locktype = 'relation' AND l.mode = 'ShareLock' AND l.granted
  AND c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND c.relkind IN ('r', 'p');
SELECT 'tables' || E'\t' || count(*) FROM pg_class c WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND c.relkind IN ('r', 'p');
SELECT 'table_list' || E'\t' || string_agg(format('%I.%I', c.relnamespace::regnamespace::text, c.relname), ', ' ORDER BY c.relnamespace::regnamespace::text, c.relname)
FROM pg_class c WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND c.relkind IN ('r', 'p');
SELECT 'sysid' || E'\t' || system_identifier FROM pg_control_system();
SELECT 'clock' || E'\t' || to_char(clock_timestamp() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"');
