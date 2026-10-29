-- G013 replay, object.delete final transition: after the provider confirmed the bytes absent, the deleting row becomes
-- deleted in the recorded deleter's context (object.deleted is recorded by the capture trigger). Variables :'org',
-- :'obj', :'deleter'. Output: "updated<TAB>n" and the resulting state. One transaction.
BEGIN;
SELECT set_config('app.user_id', :'deleter', true) \g /dev/null
WITH upd AS (UPDATE app.stored_objects SET state = 'deleted'
             WHERE id = :'obj'::uuid AND organization_id = :'org'::uuid AND state = 'deleting' RETURNING 1)
SELECT 'updated' || E'\t' || (SELECT count(*) FROM upd);
SELECT 'state' || E'\t' || coalesce((SELECT state FROM app.stored_objects WHERE id = :'obj'::uuid), 'absent');
COMMIT;
