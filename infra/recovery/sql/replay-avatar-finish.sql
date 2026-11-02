-- G074 replay, avatar.delete final transition (Plan r3.1 §3.12): after the provider confirmed the bytes of
-- avatars/<user>/<obj> absent, the user's deleting row becomes deleted in that user's context (the guard sets
-- deleted_at and keeps ready_at). Variables :'user', :'obj'. Output: "updated<TAB>n" and the resulting state of the
-- user's row ("absent" when none). One transaction.
BEGIN;
SELECT set_config('app.user_id', :'user', true) \g /dev/null
WITH upd AS (UPDATE app.account_avatars SET state = 'deleted'
             WHERE id = :'obj'::uuid AND user_id = :'user'::uuid AND state = 'deleting' RETURNING 1)
SELECT 'updated' || E'\t' || (SELECT count(*) FROM upd);
SELECT 'state' || E'\t' || coalesce((SELECT state FROM app.account_avatars WHERE id = :'obj'::uuid AND user_id = :'user'::uuid), 'absent');
COMMIT;
