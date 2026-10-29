-- G013 audit retention (infra/recovery/recovery.sh retention audit; Plan r3 §6.3), run as the retention login. The
-- transaction declares its policy in the three transaction-local settings; app.retention_eligible (00009) re-checks
-- scope, cutoff and holds on every row and is the security boundary. The explicit predicates (bound to the same
-- validated values) only shape the plan and can never widen deletion. psql variables: :count and :global (true|false),
-- :'scope', :'cutoff', :'holds', :'org' ("-" for global), :batch, :'stmt_ms', :'lock_ms'.
\if :count
BEGIN READ ONLY;
\else
BEGIN;
\endif
SET LOCAL statement_timeout = :'stmt_ms';
SET LOCAL lock_timeout = :'lock_ms';
SELECT set_config('retention.scope', :'scope', true), set_config('retention.cutoff', :'cutoff', true),
       set_config('retention.holds', :'holds', true) \g /dev/null
\if :global
\if :count
SELECT 'eligible' || E'\t' || count(*) FROM audit.events WHERE organization_id IS NULL AND occurred_at < :'cutoff'::timestamptz;
\else
WITH d AS (DELETE FROM audit.events WHERE id IN (
  SELECT id FROM audit.events
  WHERE organization_id IS NULL AND occurred_at < :'cutoff'::timestamptz
  ORDER BY occurred_at, id LIMIT :batch) RETURNING 1)
SELECT 'deleted' || E'\t' || count(*) FROM d;
\endif
\else
\if :count
SELECT 'eligible' || E'\t' || count(*) FROM audit.events WHERE organization_id = :'org'::uuid AND occurred_at < :'cutoff'::timestamptz;
\else
WITH d AS (DELETE FROM audit.events WHERE id IN (
  SELECT id FROM audit.events
  WHERE organization_id = :'org'::uuid AND occurred_at < :'cutoff'::timestamptz
  ORDER BY occurred_at, id LIMIT :batch) RETURNING 1)
SELECT 'deleted' || E'\t' || count(*) FROM d;
\endif
\endif
COMMIT;
