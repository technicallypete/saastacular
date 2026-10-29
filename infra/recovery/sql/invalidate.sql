-- G013 restore-time invalidation (infra/recovery/recovery.sh invalidate-auth; Plan r3 §5.1). One transaction on the
-- closed target, as the cluster superuser:
--   - every unrevoked session expires now (idle and absolute expiry set to the least of their value and the transaction
--     time; E1 refuses an expired session and renewal is unreachable) and gets a fresh CSRF value (minimization; the
--     backup keeps the old ones and stays private);
--   - every unreturned provider-logout reference expires now (E3 refuses it);
--   - every login transaction (recoverable PKCE verifiers) is deleted;
--   - G062: every unrevoked root signup credential and its current continuation generation expire now (the credential
--     and every one-use reference/challenge/nonce/proof stop matching; app.auth_resolve_signup and the continuation
--     bodies compare against the transaction time); a NULL continuation expiry (generation 0) stays NULL.
-- No revoked_at changes, so no session.revoked event is forged; expiry and login bookkeeping are unaudited by design.
-- Postconditions in the same transaction, else everything rolls back. Output: "key<TAB>count" lines.
BEGIN;
CREATE TEMP TABLE g013_counts (k text, n bigint) ON COMMIT DROP;
WITH s AS (
  UPDATE app.sessions SET idle_expires_at = least(idle_expires_at, now()), absolute_expires_at = least(absolute_expires_at, now()),
         csrf_token = uuid_send(gen_random_uuid()) || uuid_send(gen_random_uuid())
  WHERE revoked_at IS NULL RETURNING 1)
INSERT INTO g013_counts SELECT 'sessions_expired', count(*) FROM s;
WITH r AS (
  UPDATE app.sessions SET provider_logout_ref_expires_at = least(provider_logout_ref_expires_at, now())
  WHERE provider_logout_ref_hash IS NOT NULL AND provider_logout_returned_at IS NULL RETURNING 1)
INSERT INTO g013_counts SELECT 'logout_refs_expired', count(*) FROM r;
WITH l AS (DELETE FROM app.login_transactions RETURNING 1)
INSERT INTO g013_counts SELECT 'login_transactions_deleted', count(*) FROM l;
WITH g AS (
  UPDATE app.signup_sessions SET expires_at = least(expires_at, now()),
         continuation_expires_at = CASE WHEN continuation_expires_at IS NULL THEN NULL ELSE least(continuation_expires_at, now()) END
  WHERE revoked_at IS NULL RETURNING 1)
INSERT INTO g013_counts SELECT 'signup_sessions_expired', count(*) FROM g;
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM app.sessions WHERE revoked_at IS NULL AND idle_expires_at > now() AND absolute_expires_at > now())
     OR EXISTS (SELECT 1 FROM app.login_transactions)
     OR EXISTS (SELECT 1 FROM app.sessions WHERE provider_logout_ref_hash IS NOT NULL AND provider_logout_returned_at IS NULL
                AND provider_logout_ref_expires_at > now())
     OR EXISTS (SELECT 1 FROM app.signup_sessions WHERE revoked_at IS NULL
                AND (expires_at > now() OR continuation_expires_at > now())) THEN
    RAISE EXCEPTION 'invalidation postconditions failed';
  END IF;
END $$;
SELECT k || E'\t' || n FROM g013_counts ORDER BY k;
SELECT 'live_sessions' || E'\t' || count(*) FROM app.sessions WHERE revoked_at IS NULL AND idle_expires_at > now() AND absolute_expires_at > now();
COMMIT;
