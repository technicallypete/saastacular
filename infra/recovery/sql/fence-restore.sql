-- G013 fence release (infra/recovery/recovery.sh): end the holder's transaction by terminating its backend
-- (application_name :'holder'; removing its client container alone does not end the server backend, b0), then re-grant
-- CONNECT only to the recorded prior grantees (:'grantees') that lack it. Never adds a grantee. Idempotent.
SELECT 'holder_terminated' || E'\t' || count(*) FILTER (WHERE pg_terminate_backend(s.pid))
FROM pg_stat_activity s WHERE s.application_name = :'holder';
SELECT format('GRANT CONNECT ON DATABASE %I TO %I', current_database(), g)
FROM unnest(string_to_array(nullif(:'grantees', '-'), ',')) g
WHERE NOT EXISTS (SELECT 1 FROM pg_database d, aclexplode(d.datacl) a
                  WHERE d.datname = current_database() AND a.grantee = g::regrole AND a.privilege_type = 'CONNECT')
\gexec
