-- G013 replay, user.disable (Plan r3 §5.2, revocation gaps r1): the transaction resets the transaction-local G011 gap
-- marker first, applies the disablement with the recorded time, and reads the marker before COMMIT. The disablement
-- stands whatever the marker holds (the users capture trigger catches an audit-insert fault on this
-- revocation-priority path); it is never rolled back to obtain an event. Variables :'user', :'at'. Output: rows changed,
-- the resulting status, the marker ("-" when empty) and whether a user.disabled row was recorded in this transaction.
BEGIN;
SELECT set_config('audit.revocation_gap', '', true) \g /dev/null
WITH u AS (UPDATE app.users SET status = 'disabled', disabled_at = coalesce(disabled_at, :'at'::timestamptz), updated_at = :'at'::timestamptz
           WHERE id = :'user'::uuid AND status = 'active' RETURNING 1)
SELECT 'updated' || E'\t' || (SELECT count(*) FROM u);
SELECT 'state' || E'\t' || coalesce((SELECT status FROM app.users WHERE id = :'user'::uuid), 'absent');
SELECT 'gap' || E'\t' || coalesce(nullif(current_setting('audit.revocation_gap', true), ''), '-');
SELECT 'event' || E'\t' || count(*) FROM audit.events
WHERE action = 'user.disabled' AND subject_user_id = :'user'::uuid AND occurred_at >= transaction_timestamp();
COMMIT;
