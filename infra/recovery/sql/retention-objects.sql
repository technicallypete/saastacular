-- G013 tombstone purge (infra/recovery/recovery.sh retention objects; Plan r3 §6.5), run as the retention login with the
-- same settings contract as audit retention. Only deleted tombstones the declaration makes eligible are visible.
-- psql variables: :mode (count, candidates or delete), :'scope', :'cutoff', :'holds', :'org', :batch, :'ids'
-- (comma-separated confirmed IDs for delete), :'stmt_ms', :'lock_ms'. Output lines "eligible<TAB>n", "candidate<TAB>id"
-- or "deleted<TAB>id".
SELECT :'mode' = 'delete' AS is_delete, :'mode' = 'count' AS is_count \gset
\if :is_delete
BEGIN;
\else
BEGIN READ ONLY;
\endif
SET LOCAL statement_timeout = :'stmt_ms';
SET LOCAL lock_timeout = :'lock_ms';
SELECT set_config('retention.scope', :'scope', true), set_config('retention.cutoff', :'cutoff', true),
       set_config('retention.holds', :'holds', true) \g /dev/null
\if :is_count
SELECT 'eligible' || E'\t' || count(*) FROM app.stored_objects
WHERE organization_id = :'org'::uuid AND state = 'deleted' AND deleted_at < :'cutoff'::timestamptz;
\elif :is_delete
WITH d AS (DELETE FROM app.stored_objects WHERE id = ANY (string_to_array(:'ids', ',')::uuid[]) RETURNING id)
SELECT 'deleted' || E'\t' || id FROM d ORDER BY id;
\else
SELECT 'candidate' || E'\t' || id FROM app.stored_objects
WHERE organization_id = :'org'::uuid AND state = 'deleted' AND deleted_at < :'cutoff'::timestamptz
ORDER BY deleted_at, id LIMIT :batch;
\endif
COMMIT;
