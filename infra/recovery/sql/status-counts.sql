-- G013 status, maintenance.pending_review (Plan r3 §6.4, §7.1): counts only, never row content. Pending age never proves
-- an upload inert, so pending and deleting rows are reported for review, never selected by maintenance. Variable
-- :'pending_after' (seconds). Output: "deleting<TAB>n", "pending_old<TAB>n" and one "tombstone<TAB>id" per tombstone (the
-- tool checks them against the ledger and prints only the count). Read only, as the cluster superuser.
SELECT 'deleting' || E'\t' || count(*) FROM app.stored_objects WHERE state = 'deleting';
SELECT 'pending_old' || E'\t' || count(*) FROM app.stored_objects
WHERE state = 'pending' AND created_at < now() - make_interval(secs => :'pending_after'::int);
SELECT 'tombstone' || E'\t' || id FROM app.stored_objects WHERE state = 'deleted' ORDER BY id;
