-- G013 data fingerprint (infra/recovery/recovery.sh): one line per table of schemas app and audit, the migration ledger
-- included: "table<TAB>schema.name<TAB>row count<TAB>md5 of the rows' text in C order". Run as the cluster superuser
-- (row-level security does not apply), read only.
SELECT format($q$SELECT 'table' || E'\t' || %L || E'\t' || count(*) || E'\t' || coalesce(md5(string_agg(t::text, E'\n' ORDER BY t::text COLLATE "C")), '-') FROM %I.%I t$q$,
  n.nspname || '.' || c.relname, n.nspname, c.relname)
FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname IN ('app', 'audit') AND c.relkind IN ('r', 'p')
ORDER BY n.nspname, c.relname
\gexec
