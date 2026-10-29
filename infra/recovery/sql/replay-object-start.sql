-- G013 replay, object.delete first transition (Plan r3 §5.2): in the IDENTITY-shape context of the recorded deleter (the
-- G009 guard requires an actor; the capture trigger records object.delete_started with that actor), a ready or pending
-- row of the recorded Organization becomes deleting. Variables :'org', :'obj', :'deleter'. Output: the prior state
-- ("absent" when no such row), whether the deleter exists in the checkpoint, and the rows changed. One transaction.
BEGIN;
SELECT set_config('app.user_id', :'deleter', true) \g /dev/null
WITH cur AS (SELECT state FROM app.stored_objects WHERE id = :'obj'::uuid AND organization_id = :'org'::uuid FOR UPDATE),
     upd AS (UPDATE app.stored_objects SET state = 'deleting'
             WHERE id = :'obj'::uuid AND organization_id = :'org'::uuid AND state IN ('ready', 'pending')
               AND EXISTS (SELECT 1 FROM app.users WHERE id = :'deleter'::uuid) RETURNING 1)
SELECT 'prior' || E'\t' || coalesce((SELECT state FROM cur), 'absent') || E'\n' ||
       'deleter' || E'\t' || EXISTS (SELECT 1 FROM app.users WHERE id = :'deleter'::uuid)::text || E'\n' ||
       'updated' || E'\t' || (SELECT count(*) FROM upd);
COMMIT;
