-- G013 catalog fingerprint (infra/recovery/recovery.sh): schemas app and audit with every relation, column, policy,
-- function body, trigger, constraint, index and default privilege, and the project's roles (attributes and memberships
-- among roles named :'db'_*). ACLs are normalized: a NULL ACL is its default (acldefault) and items are sorted
-- (grantee, grantor, privilege, grantable), because a restore replays the same privileges in another item order and
-- leaves a default ACL NULL (b0, b3). Expression texts (defaults, policies, triggers, constraints, indexes) are compared
-- without parentheses and whitespace, because a restore re-deparses nested AND/OR without changing their meaning; a
-- difference only in grouping is therefore not detected here. Output: "catalog<TAB>md5" and "roles<TAB>md5".
-- Read only, as the cluster superuser.
SELECT 'catalog' || E'\t' || md5(string_agg(x, E'\n' ORDER BY x COLLATE "C")) FROM (
  SELECT 'schema ' || n.nspname || ' ' || n.nspowner::regrole::text || ' ' ||
         coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE "C") FROM (SELECT format('%s/%s/%s/%s', CASE a.grantee WHEN 0 THEN 'public' ELSE a.grantee::regrole::text END,
                  a.grantor::regrole, a.privilege_type, a.is_grantable) AS v FROM aclexplode(coalesce(n.nspacl, acldefault('n', n.nspowner))) a) q), '-') AS x
  FROM pg_namespace n WHERE n.nspname IN ('app', 'audit')
  UNION ALL
  SELECT 'rel ' || c.oid::regclass::text || ' ' || c.relkind::text || ' ' || c.relowner::regrole::text || ' ' ||
         c.relrowsecurity::text || c.relforcerowsecurity::text || ' ' || coalesce(array_to_string(c.reloptions, ','), '-') || ' ' ||
         coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE "C") FROM (SELECT format('%s/%s/%s/%s', CASE a.grantee WHEN 0 THEN 'public' ELSE a.grantee::regrole::text END,
                  a.grantor::regrole, a.privilege_type, a.is_grantable) AS v FROM aclexplode(coalesce(c.relacl, acldefault(CASE c.relkind WHEN 'S' THEN 's'::"char" ELSE 'r'::"char" END, c.relowner))) a) q), '-')
  FROM pg_class c WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace)
  UNION ALL
  SELECT 'col ' || c.oid::regclass::text || '.' || at.attname || ' ' || format_type(at.atttypid, at.atttypmod) || ' ' ||
         at.attnotnull::text || ' ' || at.attgenerated::text || ' ' || coalesce(regexp_replace(pg_get_expr(d.adbin, d.adrelid), '[()[:space:]]', '', 'g'), '-') || ' ' ||
         coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE "C") FROM (SELECT format('%s/%s/%s/%s', CASE a.grantee WHEN 0 THEN 'public' ELSE a.grantee::regrole::text END,
                  a.grantor::regrole, a.privilege_type, a.is_grantable) AS v FROM aclexplode(at.attacl) a) q), '-')
  FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
  LEFT JOIN pg_attrdef d ON d.adrelid = c.oid AND d.adnum = at.attnum
  WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace)
  UNION ALL
  SELECT 'pol ' || p.polrelid::regclass::text || ' ' || p.polname || ' ' || p.polcmd::text || ' ' || p.polpermissive::text || ' ' ||
         coalesce((SELECT string_agg(v, '+' ORDER BY v COLLATE "C") FROM (SELECT CASE r WHEN 0 THEN 'public' ELSE r::regrole::text END AS v FROM unnest(p.polroles) r) q), '-') || ' ' ||
         coalesce(regexp_replace(pg_get_expr(p.polqual, p.polrelid), '[()[:space:]]', '', 'g'), '-') || ' ' || coalesce(regexp_replace(pg_get_expr(p.polwithcheck, p.polrelid), '[()[:space:]]', '', 'g'), '-')
  FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace)
  UNION ALL
  SELECT 'fn ' || p.pronamespace::regnamespace::text || '.' || p.proname || '(' || pg_get_function_identity_arguments(p.oid) || ') ' ||
         p.proowner::regrole::text || ' ' || p.prosecdef::text || p.proleakproof::text || p.provolatile::text || ' ' ||
         coalesce(array_to_string(p.proconfig, '|'), '-') || ' ' || md5(p.prosrc) || ' ' ||
         coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE "C") FROM (SELECT format('%s/%s/%s/%s', CASE a.grantee WHEN 0 THEN 'public' ELSE a.grantee::regrole::text END,
                  a.grantor::regrole, a.privilege_type, a.is_grantable) AS v FROM aclexplode(coalesce(p.proacl, acldefault('f', p.proowner))) a) q), '-')
  FROM pg_proc p WHERE p.pronamespace IN ('app'::regnamespace, 'audit'::regnamespace)
  UNION ALL
  SELECT 'trg ' || t.tgrelid::regclass::text || ' ' || t.tgname || ' ' || t.tgenabled::text || ' ' || regexp_replace(pg_get_triggerdef(t.oid), '[()[:space:]]', '', 'g')
  FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
  WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND NOT t.tgisinternal
  UNION ALL
  SELECT 'con ' || co.conrelid::regclass::text || ' ' || co.conname || ' ' || regexp_replace(pg_get_constraintdef(co.oid), '[()[:space:]]', '', 'g')
  FROM pg_constraint co WHERE co.connamespace IN ('app'::regnamespace, 'audit'::regnamespace)
  UNION ALL
  SELECT 'idx ' || regexp_replace(pg_get_indexdef(i.indexrelid), '[()[:space:]]', '', 'g')
  FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace)
  UNION ALL
  SELECT 'defacl ' || d.defaclrole::regrole::text || ' ' || coalesce(d.defaclnamespace::regnamespace::text, '-') || ' ' ||
         d.defaclobjtype::text || ' ' ||
         coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE "C") FROM (SELECT format('%s/%s/%s/%s', CASE a.grantee WHEN 0 THEN 'public' ELSE a.grantee::regrole::text END,
                  a.grantor::regrole, a.privilege_type, a.is_grantable) AS v FROM aclexplode(d.defaclacl) a) q), '-')
  FROM pg_default_acl d
) s
UNION ALL
SELECT 'roles' || E'\t' || md5(coalesce(string_agg(x, E'\n' ORDER BY x COLLATE "C"), '')) FROM (
  SELECT 'role ' || r.rolname || ' ' || r.rolsuper::text || r.rolinherit::text || r.rolcreaterole::text || r.rolcreatedb::text ||
         r.rolcanlogin::text || r.rolreplication::text || r.rolbypassrls::text || ' ' || coalesce(array_to_string(r.rolconfig, '|'), '-') AS x
  FROM pg_roles r WHERE r.rolname LIKE :'db' || '\_%'
  UNION ALL
  SELECT 'member ' || m.member::regrole::text || '>' || m.roleid::regrole::text || ' ' || m.inherit_option::text ||
         m.set_option::text || m.admin_option::text
  FROM pg_auth_members m JOIN pg_roles r ON r.oid = m.roleid JOIN pg_roles s ON s.oid = m.member
  WHERE r.rolname LIKE :'db' || '\_%' OR s.rolname LIKE :'db' || '\_%'  -- memberships crossing the project boundary too (S-review F12)
) s;
