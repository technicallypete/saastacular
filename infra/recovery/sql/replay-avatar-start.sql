-- G074 replay, avatar.delete first transition (Plan r3.1 §3.12, C10, Q-R4): in the IDENTITY-shape context of the owning
-- user (the account_avatars guard requires the owner as actor; the profile capture records profile.updated
-- avatar_removed with that actor, as object.delete_started is recorded for the deleter) the user's profile row is
-- locked first (the facade's order); its current reference is cleared when it still names the object (avatar_version
-- and version advance by one, as the profile guard requires); then a ready or pending row of that user becomes deleting
-- (the guard's I-CUR rule refuses that while the row is still referenced, so the clear is an earlier statement of the
-- same transaction). Variables :'user', :'obj'. Output: the prior state of the user's row ("absent" when the user has
-- no row with that id), whether the user exists in the checkpoint, whether another user's row carries the id
-- ("foreign": a wrong binding), whether the reference was cleared and the rows changed. One transaction: any SQL error
-- rolls everything back and the tool reports the entry failed.
BEGIN;
SELECT set_config('app.user_id', :'user', true) \g /dev/null
SELECT 1 FROM app.account_profiles WHERE user_id = :'user'::uuid FOR NO KEY UPDATE \g /dev/null
WITH cur AS (SELECT state FROM app.account_avatars WHERE id = :'obj'::uuid AND user_id = :'user'::uuid FOR UPDATE)
SELECT 'prior' || E'\t' || coalesce((SELECT state FROM cur), 'absent');
SELECT 'user' || E'\t' || EXISTS (SELECT 1 FROM app.users WHERE id = :'user'::uuid)::text;
SELECT 'foreign' || E'\t' || EXISTS (SELECT 1 FROM app.account_avatars WHERE id = :'obj'::uuid AND user_id <> :'user'::uuid)::text;
WITH clr AS (UPDATE app.account_profiles SET avatar_object_id = NULL, avatar_version = avatar_version + 1, version = version + 1
             WHERE user_id = :'user'::uuid AND avatar_object_id = :'obj'::uuid RETURNING 1)
SELECT 'was_current' || E'\t' || (SELECT count(*) FROM clr);
WITH upd AS (UPDATE app.account_avatars SET state = 'deleting'
             WHERE id = :'obj'::uuid AND user_id = :'user'::uuid AND state IN ('ready', 'pending') RETURNING 1)
SELECT 'updated' || E'\t' || (SELECT count(*) FROM upd);
COMMIT;
