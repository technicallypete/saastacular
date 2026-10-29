-- G013 fence step X2 (infra/recovery/recovery.sh): revoke CONNECT on the current database from exactly the recorded
-- non-superuser grantees (:'grantees', comma-separated role names), then terminate their client backends (their open
-- transactions roll back). Output: "terminated<TAB>n". As the cluster superuser.
SELECT format('REVOKE CONNECT ON DATABASE %I FROM %I', current_database(), g)
FROM unnest(string_to_array(nullif(:'grantees', '-'), ',')) g
\gexec
SELECT 'terminated' || E'\t' || count(*) FILTER (WHERE pg_terminate_backend(s.pid))
FROM pg_stat_activity s WHERE s.datname = current_database() AND s.backend_type = 'client backend'
  AND s.usename = ANY (string_to_array(nullif(:'grantees', '-'), ','));
