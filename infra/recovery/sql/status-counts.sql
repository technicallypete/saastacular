-- G013 status, maintenance.pending_review (Plan r3 §6.4, §7.1): counts only, never row content. Pending age never proves
-- an upload inert, so pending and deleting rows are reported for review, never selected by maintenance. Variable
-- :'pending_after' (seconds). Output: "deleting<TAB>n", "pending_old<TAB>n" and one "tombstone<TAB>id" per tombstone (the
-- tool checks them against the ledger and prints only the count). Read only, as the cluster superuser.
-- G074 (Plan r3.1 §3.8, §3.12): the same three lines for identity-owned avatar objects: "avatar_deleting<TAB>n",
-- "avatar_pending_old<TAB>n" and one "avatar_tombstone<TAB>id" per withdrawn published avatar (deleted with ready_at; the
-- tool checks them against avatar.delete entries). A compensated upload that never became ready is not a restriction.
SELECT 'deleting' || E'\t' || count(*) FROM app.stored_objects WHERE state = 'deleting';
SELECT 'pending_old' || E'\t' || count(*) FROM app.stored_objects
WHERE state = 'pending' AND created_at < now() - make_interval(secs => :'pending_after'::int);
SELECT 'tombstone' || E'\t' || id FROM app.stored_objects WHERE state = 'deleted' ORDER BY id;
SELECT 'avatar_deleting' || E'\t' || count(*) FROM app.account_avatars WHERE state = 'deleting';
SELECT 'avatar_pending_old' || E'\t' || count(*) FROM app.account_avatars
WHERE state = 'pending' AND created_at < now() - make_interval(secs => :'pending_after'::int);
SELECT 'avatar_tombstone' || E'\t' || id FROM app.account_avatars WHERE state = 'deleted' AND ready_at IS NOT NULL ORDER BY id;
