-- G013 replay gap record (Plan r3 §5.2, Q7 accepted): after a committed disablement whose transaction carried the gap
-- marker, one independent attempt, bounded to 2 s, of the existing revocation.unaudited action in its existing G6 shape
-- (outcome success, reason disabled, resource user = subject, no Organization or Group; the BEFORE INSERT trigger sets
-- actor unknown, since no lifecycle context exists). Variable :'user'. Output: "recorded<TAB>1".
BEGIN;
SET LOCAL statement_timeout = '2s';
SET LOCAL lock_timeout = '2s';
WITH i AS (INSERT INTO audit.events (action, outcome, reason, resource_type, resource_id, subject_user_id)
           VALUES ('revocation.unaudited', 'success', 'disabled', 'user', :'user'::uuid, :'user'::uuid) RETURNING 1)
SELECT 'recorded' || E'\t' || (SELECT count(*) FROM i);
COMMIT;
