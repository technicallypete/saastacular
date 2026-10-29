//go:build integration

// G011 B-AUDIT-CATALOG and B-AUDIT-MIGRATION (Plan r2 §5, §9) against the disposable PostgreSQL of the migrations
// suite. TestAuditCatalog pins the exact version-8 audit boundary on top of the unchanged version-7 app catalog (CT-AUD-1
// .. CT-AUD-6) and eight rolled-back missing-protection detectors (CT-AUD-NEG N1-N8); TestAuditMigration covers empty to
// 8, a populated version-7 world to 8 with every row and ACL identical, the refusal of a missing or unclean audit schema,
// the down refusal while history exists, down and up on an empty history, and a rolled-back up. TestAuditPersistSeed and
// TestAuditPersistVerify run around a PostgreSQL stop/start. Setup and rolled-back controls use the attested superuser;
// assertions use the attested runtime login, the retention role (assumed by the superuser inside a rolled-back
// transaction, attested) or the owner with FORCE.
package migrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// The version-8 app catalog: the version-7 pins, with the two audit policies in the (all-schema) policy inventory.
var auditWantAppPolicies = strings.Replace(objWantPolicies, "business_profiles_update:business_profiles:w:true:ow+rt;",
	"business_profiles_update:business_profiles:w:true:ow+rt;events_insert:events:a:true:ow+rt;events_select:events:r:true:ow+rt;", 1)

// The pinned version-8 audit boundary.
const (
	auditWantRelations = "events:r:ow:t:t;events_pkey:i:ow:f:f;events_tenant_time_idx:i:ow:f:f"
	auditWantColumns   = "id:uuid:t;schema_version:smallint:t;occurred_at:timestamp with time zone:t;action:text:t;outcome:text:t;" +
		"reason:text:f;operation:text:f;actor_kind:text:t;actor_user_id:uuid:f;organization_id:uuid:f;group_id:uuid:f;" +
		"resource_type:text:f;resource_id:uuid:f;subject_user_id:uuid:f;role:text:f;previous_role:text:f;ttl_seconds:integer:f;" +
		"request_id:text:f;trace_id:text:f"
	auditWantConstraints = "events_action_valid:c;events_actor_consistent:c;events_actor_kind_valid:c;events_group_needs_org:c;" +
		"events_ids_nonzero:c;events_operation_valid:c;events_outcome_valid:c;events_pkey:p;events_previous_role_valid:c;" +
		"events_reason_valid:c;events_request_id_valid:c;events_resource_pair:c;events_resource_type_valid:c;events_role_valid:c;" +
		"events_schema_version:c;events_trace_id_valid:c;events_ttl_valid:c"
	auditWantIndexes  = "events_pkey:id:-;events_tenant_time_idx:organization_id, occurred_at, id:(organization_id IS NOT NULL)"
	auditWantPolicies = "events_insert:a:true:ow+rt;events_select:r:true:ow+rt"
	auditWantACL      = "events:rt:SELECT;events.action:rt:INSERT;events.group_id:rt:INSERT;events.operation:rt:INSERT;" +
		"events.organization_id:rt:INSERT;events.outcome:rt:INSERT;events.previous_role:rt:INSERT;events.reason:rt:INSERT;" +
		"events.resource_id:rt:INSERT;events.resource_type:rt:INSERT;events.role:rt:INSERT;events.subject_user_id:rt:INSERT;" +
		"events.ttl_seconds:rt:INSERT"
	auditWantSchema    = "rt=USAGE"
	auditWantFunctions = "capture_business_profiles:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"capture_group_business_profiles:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"capture_groups:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"capture_organizations:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"capture_role_bindings:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"capture_session_revocations:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"capture_sessions:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"capture_stored_objects:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"capture_users:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"events_attribute:ow:false:false:v:search_path=pg_catalog, pg_temp:-;" +
		"events_refuse:ow:false:false:v:search_path=pg_catalog, pg_temp:-"
	auditWantTriggers = "events_attribute:O:CREATE TRIGGER events_attribute BEFORE INSERT ON audit.events FOR EACH ROW EXECUTE FUNCTION audit.events_attribute();" +
		"events_refuse:O:CREATE TRIGGER events_refuse BEFORE DELETE OR UPDATE ON audit.events FOR EACH ROW EXECUTE FUNCTION audit.events_refuse();" +
		"events_refuse_truncate:O:CREATE TRIGGER events_refuse_truncate BEFORE TRUNCATE ON audit.events FOR EACH STATEMENT EXECUTE FUNCTION audit.events_refuse()"
	// auditWantCapture is the exact definition (pg_get_triggerdef) and enabled state of every non-internal trigger of the
	// eight audited app tables: the 18 capture triggers and the 6 existing guards. auditWantSrc is md5 of each audit
	// function body. Both are pinned from the reviewed migration.
	auditWantCapture = "72ec620fbd6f5ad7b251c34f00064792"
	auditWantSrc     = "capture_business_profiles=54b25277fea8393e3cf96667abbfb944;capture_group_business_profiles=57adad683961a1e335525c75f88f6d32;" +
		"capture_groups=7d69196ecb61eda26dd751464d90ecdc;capture_organizations=16eeb1460258e225abc3e75974674d19;" +
		"capture_role_bindings=f2661dd8f68079a2a7a21a414a2d1f4b;capture_session_revocations=182fdad03f97751dcd4e07d2c7d61512;" +
		"capture_sessions=8ca2f9d79920646392db672be0c6e611;capture_stored_objects=c9f6f53a3b8e152122694607267e2e9e;" +
		"capture_users=0c547f00f1d29c013f49ca295dc00da2;events_attribute=f949e2ff067f3e78dd0d8ca7a2d414f5;" +
		"events_refuse=ee436a151f6bcf4bf893747a8cf243f9"
	// The two policy expressions (pg_get_expr, normalized): the read is the current org_owner of the active
	// Organization; the insert binds a user actor to the context user and the Organization to the context Organization.
	auditWantSelectExpr = "((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) = " +
		"'org_owner'::text)) WITH CHECK"
	auditWantInsertExpr = "WITH CHECK (((actor_kind <> 'user'::text) OR (actor_user_id = ( SELECT app.ctx_user() AS ctx_user))) AND " +
		"((organization_id IS NULL) OR (organization_id = ( SELECT app.ctx_org() AS ctx_org))))"
)

// auditParts reads the audit parts of the catalog, keyed by detector code.
func auditParts(t *testing.T, q catalogQuerier, cfg Config) map[string]string {
	t.Helper()
	args := rlsLabelArgs(cfg)
	one := func(sql string) string {
		var s *string
		if err := q.QueryRow(context.Background(), sql+rlsUseArgsWhere(sql), args...).Scan(&s); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		if s == nil {
			return ""
		}
		return *s
	}
	p := map[string]string{}
	p["audit-relation"] = one(`SELECT string_agg(format('%s:%s:%s:%s:%s', c.relname, c.relkind::text, ` + rlsLabel("c.relowner") + `,
  CASE WHEN c.relrowsecurity THEN 't' ELSE 'f' END, CASE WHEN c.relforcerowsecurity THEN 't' ELSE 'f' END), ';' ORDER BY c.relname)
FROM pg_class c WHERE c.relnamespace = 'audit'::regnamespace`)
	p["audit-columns"] = one(`SELECT string_agg(format('%s:%s:%s', a.attname, format_type(a.atttypid, a.atttypmod),
  CASE WHEN a.attnotnull THEN 't' ELSE 'f' END), ';' ORDER BY a.attnum)
FROM pg_attribute a WHERE a.attrelid = to_regclass('audit.events') AND a.attnum > 0 AND NOT a.attisdropped`)
	p["audit-constraints"] = one(`SELECT string_agg(conname || ':' || contype::text, ';' ORDER BY conname)
FROM pg_constraint WHERE conrelid = to_regclass('audit.events')`)
	p["audit-indexes"] = one(`SELECT string_agg(c.relname || ':' || (SELECT string_agg(a.attname, ', ' ORDER BY k.ord)
  FROM unnest(i.indkey) WITH ORDINALITY k(attnum, ord) JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = k.attnum) || ':' ||
  coalesce(pg_get_expr(i.indpred, i.indrelid), '-'), ';' ORDER BY c.relname)
FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE i.indrelid = to_regclass('audit.events')`)
	p["audit-policies"] = one(`SELECT string_agg(format('%s:%s:%s:%s', p.polname, p.polcmd, p.polpermissive::text,
  (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (SELECT ` + rlsLabel("x") + ` AS v FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname)
FROM pg_policy p WHERE p.polrelid = to_regclass('audit.events')`)
	p["audit-acl"] = one(`SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") FROM (
  SELECT c.relname::text AS k, ` + rlsLabel("a.grantee") + ` AS g, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
  FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
  WHERE c.relnamespace = 'audit'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
  UNION ALL
  SELECT c.relname || '.' || at.attname, ` + rlsLabel("a.grantee") + `, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
  FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
  CROSS JOIN LATERAL aclexplode(at.attacl) a
  WHERE c.relnamespace = 'audit'::regnamespace GROUP BY c.relname, at.attname, a.grantee) t`)
	p["audit-schema"] = one(`SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") FROM (
  SELECT ` + rlsLabel("a.grantee") + ` || '=' || a.privilege_type AS v
  FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a WHERE n.nspname = 'audit' AND a.grantee <> n.nspowner) l`)
	p["audit-functions"] = one(`SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname, ` + rlsLabel("p.proowner") + `,
  p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
  coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
     SELECT ` + rlsLabel("a.grantee") + ` || '=' || a.privilege_type AS v FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-')),
  ';' ORDER BY p.proname) FROM pg_proc p WHERE p.pronamespace = 'audit'::regnamespace`)
	p["audit-triggers"] = one(`SELECT string_agg(t.tgname || ':' || t.tgenabled::text || ':' || pg_get_triggerdef(t.oid), ';' ORDER BY t.tgname)
FROM pg_trigger t WHERE t.tgrelid = to_regclass('audit.events') AND NOT t.tgisinternal`)
	p["audit-capture"] = one(`SELECT md5(string_agg(c.relname || '.' || t.tgname || ':' || t.tgenabled::text || ':' || pg_get_triggerdef(t.oid),
  E'\n' ORDER BY c.relname, t.tgname))
FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
WHERE c.relnamespace = 'app'::regnamespace AND NOT t.tgisinternal`)
	p["audit-capture-text"] = one(`SELECT string_agg(c.relname || '.' || t.tgname || ':' || t.tgenabled::text || ':' || pg_get_triggerdef(t.oid),
  E'\n' ORDER BY c.relname, t.tgname)
FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
WHERE c.relnamespace = 'app'::regnamespace AND NOT t.tgisinternal`)
	p["audit-src"] = one(`SELECT string_agg(p.proname || '=' || md5(p.prosrc), ';' ORDER BY p.proname)
FROM pg_proc p WHERE p.pronamespace = 'audit'::regnamespace`)
	rows, err := q.Query(context.Background(), `SELECT p.polname::text, coalesce(pg_get_expr(p.polqual, p.polrelid), ''),
  coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '') FROM pg_policy p WHERE p.polrelid = to_regclass('audit.events') ORDER BY p.polname`)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	for rows.Next() {
		var name, qual, check string
		if err := rows.Scan(&name, &qual, &check); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		p["audit-policy-expr:"+name] = normalize(qual + " WITH CHECK " + check)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("catalog: %v", err)
	}
	return p
}

// rlsUseArgsWhere appends the six label parameters to a query that does not reference them all.
func rlsUseArgsWhere(sql string) string {
	if strings.Contains(sql, "$6") {
		return ""
	}
	if strings.Contains(sql, "WHERE") {
		return rlsUseArgs
	}
	return " WHERE true" + rlsUseArgs
}

// auditViolations is the version-8 detector: the sorted codes of every part that differs from its pin. It covers the
// unchanged version-7 app catalog (every objViolations part except the stored-object trigger list, which the capture
// inventory replaces, and the policy inventory with the two audit policies) and the audit boundary.
func auditViolations(t *testing.T, q catalogQuerier, cfg Config) []string {
	t.Helper()
	p := objCatalogParts(t, q, cfg)
	for k, v := range auditParts(t, q, cfg) {
		p[k] = v
	}
	want := map[string]string{"rls-flags": objWantFlags, "policy-inventory": auditWantAppPolicies, "acl": objWantACL,
		"functions": objWantFunctions, "schema": rlsWantSchema, "roles": rlsWantRoles, "memberships": rlsWantMemberships,
		"views": rlsWantViews, "obj-columns": objWantColumns, "obj-constraints": objWantConstraints, "obj-indexes": objWantIndexes,
		"obj-guard-src": objWantGuardSrc, "audit-relation": auditWantRelations, "audit-columns": auditWantColumns,
		"audit-constraints": auditWantConstraints, "audit-indexes": auditWantIndexes, "audit-policies": auditWantPolicies,
		"audit-acl": auditWantACL, "audit-schema": auditWantSchema, "audit-functions": auditWantFunctions,
		"audit-triggers": auditWantTriggers, "audit-capture": auditWantCapture, "audit-src": auditWantSrc}
	for name, expr := range rlsWantExprs {
		want["policy-expr:"+name] = expr
	}
	var codes []string
	for k, w := range want {
		if p[k] != w {
			codes = append(codes, k)
			t.Logf("detector %s: got %s", k, p[k])
			if k == "audit-capture" {
				t.Logf("detector audit-capture text:\n%s", p["audit-capture-text"])
			}
		}
	}
	for name, w := range map[string]string{"events_select": auditWantSelectExpr, "events_insert": auditWantInsertExpr} {
		if got := p["audit-policy-expr:"+name]; got != w {
			codes = append(codes, "audit-policy-expr:"+name)
			t.Logf("detector audit-policy-expr:%s: got %s", name, got)
		}
	}
	slices.Sort(codes)
	return slices.Compact(codes)
}

// auditCount is the number of audit rows, read by the attested superuser (row-level security does not apply to it).
func auditCount(t *testing.T, setup *pgx.Conn) int {
	t.Helper()
	var n int
	if err := setup.QueryRow(context.Background(), `SELECT count(*)::int FROM audit.events`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// auditPurge empties the disposable history through the attested superuser (the refusal trigger disabled around one
// TRUNCATE, in one transaction): a test-only privileged action, never production.
func auditPurge(t *testing.T, setup *pgx.Conn) {
	t.Helper()
	if _, err := setup.Exec(context.Background(), `BEGIN; ALTER TABLE audit.events DISABLE TRIGGER events_refuse_truncate; `+
		`TRUNCATE audit.events; ALTER TABLE audit.events ENABLE TRIGGER events_refuse_truncate; COMMIT`); err != nil {
		t.Fatalf("purge: %v", err)
	}
}

// auditRowsSnapshot hashes every app row (the ten version-6 tables and stored_objects) for the populated-upgrade
// comparison.
func auditRowsSnapshot(t *testing.T, setup *pgx.Conn) string {
	t.Helper()
	var h string
	if err := setup.QueryRow(context.Background(), `SELECT count(*) || ':' || coalesce(md5(string_agg(r::text, E'\n' ORDER BY r::text)), '-') FROM app.stored_objects r`).Scan(&h); err != nil {
		t.Fatal(err)
	}
	return rlsRowsSnapshot(t, setup) + ";stored_objects=" + h
}

// auditRetention runs fn as the retention role inside a rolled-back superuser transaction, after attesting the
// assumed role (current_user is retention: not superuser, no BYPASSRLS, a member of nothing that matters).
func auditRetention(t *testing.T, setup *pgx.Conn, cfg Config, fn func(tx pgx.Tx) error) error {
	t.Helper()
	ctx := context.Background()
	tx, err := setup.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE "+quote(cfg.RetentionRole())); err != nil {
		t.Fatal(err)
	}
	var got string
	if err := tx.QueryRow(ctx, `SELECT format('%s|%s|%s', current_user, r.rolsuper, r.rolbypassrls) FROM pg_roles r WHERE r.rolname = current_user`).Scan(&got); err != nil ||
		got != cfg.RetentionRole()+"|f|f" {
		t.Fatalf("retention attestation: %q (%v)", got, err)
	}
	return fn(tx)
}

// auditFS is the migration set through 00008 (G013 Plan r3 §11.3): the G011 catalog and migration evidence keeps its
// exact version-8 meaning (CT-AUD-3/5: retention holds nothing); 00009 is B-RETENTION-*'s.
func auditFS(t *testing.T) fs.FS {
	t.Helper()
	return subsetFS(t, 8)
}

func auditUp(t *testing.T, cfg Config) *Runner {
	t.Helper()
	r := runnerFor(t, cfg, auditFS(t))
	objRunUp(t, r, "current_version\t8\n")
	return r
}

// TestAuditCatalog (B-AUDIT-CATALOG): the exact version-8 catalog, runtime and retention refusals, owner FORCE, and the
// eight rolled-back missing-protection detectors.
func TestAuditCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r := auditUp(t, cfg)
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	w := seedRLSWorld(t, setup, "aud") // privileged fixture writes: their capture rows carry actor unknown

	t.Run("CT-AUD-1 relation, columns, constraints, index", func(t *testing.T) {
		p := auditParts(t, cat, cfg)
		for k, want := range map[string]string{"audit-relation": auditWantRelations, "audit-columns": auditWantColumns,
			"audit-constraints": auditWantConstraints, "audit-indexes": auditWantIndexes} {
			if p[k] != want {
				t.Errorf("%s:\n got %s\nwant %s", k, p[k], want)
			}
		}
		metadataThroughVersion8Only(t, cfg, r)
	})
	t.Run("CT-AUD-2 row-level security and policies", func(t *testing.T) {
		p := auditParts(t, cat, cfg)
		if p["audit-policies"] != auditWantPolicies || !strings.HasPrefix(p["audit-relation"], "events:r:ow:t:t") {
			t.Errorf("policies %s, relation %s", p["audit-policies"], p["audit-relation"])
		}
		if v := auditViolations(t, cat, cfg); slices.ContainsFunc(v, func(c string) bool { return strings.HasPrefix(c, "audit-policy") }) {
			t.Errorf("policy expressions: %v", v)
		}
	})
	t.Run("CT-AUD-3 every ACL", func(t *testing.T) {
		p := auditParts(t, cat, cfg)
		if p["audit-acl"] != auditWantACL || p["audit-schema"] != auditWantSchema {
			t.Errorf("acl %s; schema %s", p["audit-acl"], p["audit-schema"])
		}
		if p["audit-functions"] != auditWantFunctions {
			t.Errorf("functions %s", p["audit-functions"])
		}
		var retention bool
		if err := cat.QueryRow(ctx, `SELECT has_schema_privilege($1, 'audit', 'USAGE') OR has_table_privilege($1, 'audit.events', 'SELECT')`,
			cfg.RetentionRole()).Scan(&retention); err != nil || retention {
			t.Errorf("retention holds an audit privilege (%v)", err)
		}
		op := objCatalogParts(t, cat, cfg)
		if op["acl"] != objWantACL || op["schema"] != rlsWantSchema || op["functions"] != objWantFunctions {
			t.Error("the app ACLs changed")
		}
	})
	t.Run("CT-AUD-4 triggers, functions and bodies", func(t *testing.T) {
		v := auditViolations(t, cat, cfg)
		if len(v) != 0 {
			t.Fatalf("version-8 detector: %v", v)
		}
		// No row-dump construct; the only exception blocks are the two revocation-priority branches.
		var bad string
		if err := cat.QueryRow(ctx, `SELECT coalesce(string_agg(proname, ','), '') FROM pg_proc WHERE pronamespace = 'audit'::regnamespace
  AND (prosrc ~* '(row_to_json|to_jsonb|hstore|NEW\.\*|OLD\.\*)')`).Scan(&bad); err != nil || bad != "" {
			t.Errorf("row-dump constructs in %q (%v)", bad, err)
		}
		var exc string
		if err := cat.QueryRow(ctx, `SELECT string_agg(proname || ':' || (length(prosrc) - length(replace(prosrc, 'EXCEPTION WHEN', ''))) / 14, ',' ORDER BY proname)
FROM pg_proc WHERE pronamespace = 'audit'::regnamespace AND prosrc LIKE '%EXCEPTION WHEN%'`).Scan(&exc); err != nil ||
			exc != "capture_session_revocations:1,capture_users:1" {
			t.Errorf("exception blocks %q (%v)", exc, err)
		}
	})
	t.Run("CT-AUD-5 runtime and retention refusals", func(t *testing.T) {
		rt := rlsRuntimeConn(t, cfg)
		before, rows := auditViolations(t, cat, cfg), auditCount(t, setup)
		if rows == 0 {
			t.Fatal("no fixture history")
		}
		stmts := []struct{ what, sql string }{
			{"UPDATE", `UPDATE audit.events SET outcome = 'failed'`},
			{"DELETE", `DELETE FROM audit.events`},
			{"TRUNCATE", `TRUNCATE audit.events`},
			{"ALTER TABLE", `ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY`},
			{"DROP TABLE", `DROP TABLE audit.events`},
			{"CREATE in audit", `CREATE TABLE audit.runtime_ddl (id int)`},
			{"DISABLE TRIGGER", `ALTER TABLE audit.events DISABLE TRIGGER events_refuse`},
			{"DISABLE capture", `ALTER TABLE app.sessions DISABLE TRIGGER audit_capture_revocations`},
			{"SET ROLE owner", `SET ROLE ` + quote(cfg.OwnerRole())},
			{"row_security off", `SET LOCAL row_security = off; SELECT count(*) FROM audit.events`},
			{"INSERT actor", `INSERT INTO audit.events (action, outcome, actor_kind) VALUES ('auth.sign_in', 'failed', 'system')`},
			{"INSERT id", `INSERT INTO audit.events (id, action, outcome) VALUES (gen_random_uuid(), 'auth.sign_in', 'failed')`},
			{"EXECUTE capture", `SELECT audit.capture_users()`},
		}
		for _, s := range stmts {
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			_, err = tx.Exec(ctx, s.sql)
			_ = tx.Rollback(ctx)
			if got := sqlstateOf(err); got != "42501" && !(s.what == "EXECUTE capture" && got == "0A000") {
				t.Errorf("runtime %s: %v (want 42501)", s.what, err)
			}
			if s.what == "SET ROLE owner" {
				continue // the retention role is assumed by the superuser session, whose SET ROLE rights are the session user's
			}
			err = auditRetention(t, setup, cfg, func(tx pgx.Tx) error { _, err := tx.Exec(ctx, s.sql); return err })
			if got := sqlstateOf(err); got != "42501" && !(s.what == "EXECUTE capture" && got == "0A000") {
				t.Errorf("retention %s: %v (want 42501)", s.what, err)
			}
		}
		// GRANT by a non-owner grants nothing (a warning, no error, for runtime's own SELECT; an error otherwise).
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = tx.Exec(ctx, `GRANT SELECT ON audit.events TO PUBLIC`)
		_ = tx.Commit(ctx)
		if err := auditRetention(t, setup, cfg, func(tx pgx.Tx) error { _, err := tx.Exec(ctx, `GRANT SELECT ON audit.events TO PUBLIC`); return err }); sqlstateOf(err) != "42501" {
			t.Errorf("retention GRANT: %v", err)
		}
		if after := auditViolations(t, cat, cfg); !slices.Equal(after, before) || len(after) != 0 || auditCount(t, setup) != rows {
			t.Errorf("catalog or rows changed: %v -> %v, rows %d -> %d", before, after, rows, auditCount(t, setup))
		}
	})
	t.Run("CT-AUD-6 owner FORCE", func(t *testing.T) {
		own := rlsOwnerConn(t, cfg)
		var n int
		if err := own.QueryRow(ctx, `SELECT count(*)::int FROM audit.events`).Scan(&n); err != nil || n != 0 {
			t.Errorf("owner without context sees %d rows (%v)", n, err)
		}
		tx, err := own.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		rlsContext(t, tx, w.userA, w.orgA, "", "")
		var orgs string
		if err := tx.QueryRow(ctx, `SELECT coalesce(string_agg(DISTINCT organization_id::text, ','), '') FROM audit.events`).Scan(&orgs); err != nil || orgs != w.orgA {
			t.Errorf("owner in context A sees Organizations %q (%v)", orgs, err)
		}
		for _, s := range []string{`UPDATE audit.events SET outcome = 'failed'`, `DELETE FROM audit.events`} {
			if tag, err := tx.Exec(ctx, s); err != nil || tag.RowsAffected() != 0 {
				t.Errorf("owner %s: %v rows, %v", s, tag.RowsAffected(), err)
			}
		}
		_ = tx.Rollback(ctx)
		if _, err := own.Exec(ctx, `TRUNCATE audit.events`); sqlstateOf(err) != "42501" {
			t.Errorf("owner TRUNCATE: %v (want the append-only refusal)", err)
		}
	})
	t.Run("CT-AUD-NEG rolled-back missing-protection detectors", func(t *testing.T) {
		for _, n := range []struct{ id, sql, code string }{
			{"N1", `ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY`, "audit-relation"},
			{"N2", `GRANT UPDATE ON audit.events TO ` + quote(cfg.RuntimeRole()), "audit-acl"},
			{"N3", `GRANT INSERT (actor_user_id) ON audit.events TO ` + quote(cfg.RuntimeRole()), "audit-acl"},
			{"N4", `DROP TRIGGER events_refuse ON audit.events`, "audit-triggers"},
			{"N5", `ALTER TABLE app.sessions DISABLE TRIGGER audit_capture_revocations`, "audit-capture"},
			{"N6", `CREATE POLICY events_probe ON audit.events FOR SELECT TO ` + quote(cfg.RuntimeRole()) + ` USING (true)`, "audit-policies"},
			{"N7", `GRANT USAGE ON SCHEMA audit TO PUBLIC`, "audit-schema"},
			{"N8", `ALTER FUNCTION audit.capture_users() SECURITY DEFINER`, "audit-functions"},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			v := auditViolations(t, tx, cfg)
			_ = tx.Rollback(ctx)
			// N6 also adds a policy to the all-schema app inventory; every other control yields exactly its code.
			want := []string{n.code}
			if n.id == "N6" {
				want = []string{"audit-policies", "policy-inventory"}
			}
			if !slices.Equal(v, want) {
				t.Errorf("%s: the detector reported %v, want %v", n.id, v, want)
			}
		}
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", v)
		}
	})
}

// TestAuditMigration (B-AUDIT-MIGRATION): empty to 8; a populated version-7 world to 8 with every row and ACL identical
// and no backfilled history; repeat up; a missing or unclean audit schema refused with nothing changed; down refused
// while history exists; down and up on an empty history restore the exact version-7 and version-8 catalogs (and
// reapply); a rolled-back up restores the owner and the version-7 catalog.
func TestAuditMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	full := "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\n" +
		"applied\t00005_organization_policy.sql\napplied\t00006_row_level_security.sql\napplied\t00007_stored_objects.sql\n" +
		"applied\t00008_audit_events.sql\ncurrent_version\t8\n"

	t.Run("MG-AUD-1 empty to version 8", func(t *testing.T) {
		reset(t, cfg)
		r := runnerFor(t, cfg, auditFS(t))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != full {
			t.Fatalf("up: %q", out)
		}
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
		if auditCount(t, setup) != 0 {
			t.Fatal("a fresh history is not empty")
		}
		rlsOwnerRestored(t, cfg, r)
	})
	var populated, acl string
	t.Run("MG-AUD-2 populated version 7 to 8 keeps every row and ACL; repeat up", func(t *testing.T) {
		reset(t, cfg)
		objRunUp(t, runnerFor(t, cfg, subsetFS(t, 7)), "current_version\t7\n")
		seedRLSWorld(t, setup, "aum")
		populated, acl = auditRowsSnapshot(t, setup), objCatalogParts(t, cat, cfg)["acl"]
		r := runnerFor(t, cfg, auditFS(t))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00008_audit_events.sql\ncurrent_version\t8\n" {
			t.Fatalf("up: %q", out)
		}
		if got := auditRowsSnapshot(t, setup); got != populated {
			t.Fatalf("rows changed:\n%s\n%s", populated, got)
		}
		if got := objCatalogParts(t, cat, cfg)["acl"]; got != acl {
			t.Fatal("app ACLs changed")
		}
		if auditCount(t, setup) != 0 {
			t.Fatal("history was backfilled: it starts at version 8")
		}
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "no pending migrations\ncurrent_version\t8\n" {
			t.Fatalf("repeat up: %q", out)
		}
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
	})
	t.Run("MG-AUD-3 a missing or unclean audit schema is refused, nothing changed", func(t *testing.T) {
		r := runnerFor(t, cfg, auditFS(t))
		auditPurge(t, setup)
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out != "rolled_back\t00008_audit_events.sql\ncurrent_version\t7\n" {
			t.Fatalf("down: %q", out)
		}
		for _, c := range []struct{ name, break_, restore, want string }{
			{"stray object", `CREATE TABLE audit.stray (id int)`, `DROP TABLE audit.stray`, "schema audit is not empty and owner-only"},
			{"PUBLIC usage", `GRANT USAGE ON SCHEMA audit TO PUBLIC`, `REVOKE USAGE ON SCHEMA audit FROM PUBLIC`, "schema audit is not empty and owner-only"},
			{"missing", `ALTER SCHEMA audit RENAME TO audit_moved`, `ALTER SCHEMA audit_moved RENAME TO audit`, "schema audit is missing or not owned"},
			{"wrong owner", `ALTER SCHEMA audit OWNER TO ` + quote(os.Getenv("DB_SUPERUSER")), `ALTER SCHEMA audit OWNER TO ` + quote(cfg.OwnerRole()),
				"schema audit is missing or not owned"},
		} {
			if _, err := setup.Exec(ctx, c.break_); err != nil {
				t.Fatalf("%s: %v", c.name, err)
			}
			var buf bytes.Buffer
			err := r.Up(ctx, &buf)
			if _, rerr := setup.Exec(ctx, c.restore); rerr != nil {
				t.Fatalf("%s restore: %v", c.name, rerr)
			}
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("%s: up %v, want %q", c.name, err, c.want)
			}
			if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t7\n" {
				t.Errorf("%s: version %q", c.name, out)
			}
			if got := auditRowsSnapshot(t, setup); got != populated {
				t.Errorf("%s: rows changed", c.name)
			}
			rlsOwnerRestored(t, cfg, r)
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-7 detector after the refusals: %v", v)
		}
		objRunUp(t, r, "current_version\t8\n")
	})
	t.Run("MG-AUD-4 down refused while history exists", func(t *testing.T) {
		r := runnerFor(t, cfg, auditFS(t))
		if _, err := setup.Exec(ctx, `INSERT INTO app.organizations (name) VALUES ('Audit history')`); err != nil {
			t.Fatal(err)
		}
		rows, n := auditRowsSnapshot(t, setup), auditCount(t, setup)
		if n == 0 {
			t.Fatal("no history")
		}
		var buf bytes.Buffer
		err := r.Down(ctx, &buf)
		if err == nil || !strings.Contains(err.Error(), "refusing to roll back 00008") {
			t.Fatalf("down with history: %v", err)
		}
		if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t8\n" {
			t.Fatalf("version after the refused down: %q", out)
		}
		if auditRowsSnapshot(t, setup) != rows || auditCount(t, setup) != n {
			t.Fatal("rows or history changed")
		}
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector after the refused down: %v", v)
		}
	})
	t.Run("MG-AUD-5 down and up on an empty history; reapply", func(t *testing.T) {
		r := runnerFor(t, cfg, auditFS(t))
		auditPurge(t, setup)
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out != "rolled_back\t00008_audit_events.sql\ncurrent_version\t7\n" {
			t.Fatalf("down: %q", out)
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-7 detector after down: %v", v)
		}
		metadataThroughVersion7Only(t, cfg, r)
		var acl string
		if err := cat.QueryRow(ctx, `SELECT coalesce(array_to_string(nspacl, ','), '') FROM pg_namespace WHERE nspname = 'audit' AND nspowner = to_regrole($1)`,
			cfg.OwnerRole()).Scan(&acl); err != nil || strings.Contains(acl, cfg.RuntimeRole()) {
			t.Fatalf("schema audit after down: %q (%v)", acl, err)
		}
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00008_audit_events.sql\ncurrent_version\t8\n" {
			t.Fatalf("up: %q", out)
		}
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector after up: %v", v)
		}
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out != "rolled_back\t00008_audit_events.sql\napplied\t00008_audit_events.sql\ncurrent_version\t8\n" {
			t.Fatalf("reapply: %q", out)
		}
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector after reapply: %v", v)
		}
	})
	t.Run("MG-AUD-6 a failed up rolls back and restores the owner", func(t *testing.T) {
		r := runnerFor(t, cfg, auditFS(t))
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); !strings.HasSuffix(out, "current_version\t7\n") {
			t.Fatalf("down: %q", out)
		}
		if _, err := setup.Exec(ctx, `CREATE POLICY audit_unexpected ON app.organizations FOR SELECT TO `+quote(cfg.RuntimeRole())+` USING (false)`); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		err := r.Up(ctx, &buf)
		if _, derr := setup.Exec(ctx, `DROP POLICY audit_unexpected ON app.organizations`); derr != nil {
			t.Fatal(derr)
		}
		if err == nil || !strings.Contains(err.Error(), "version-7 precondition: policy inventory") {
			t.Fatalf("up with an unexpected policy: %v", err)
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-7 catalog after the rolled-back up: %v", v)
		}
		rlsOwnerRestored(t, cfg, r)
		objRunUp(t, r, "current_version\t8\n")
	})
}

// TestAuditPersistSeed and TestAuditPersistVerify (B-AUDIT-MIGRATION persistence): history written before a
// PostgreSQL stop/start is returned unchanged after it, to the superuser and to the Organization owner through
// row-level security.
func TestAuditPersistSeed(t *testing.T) {
	cfg := itConfig(t)
	reset(t, cfg)
	auditUp(t, cfg)
	setup := rlsSetupConn(t, cfg)
	w := seedRLSWorld(t, setup, "aup")
	var event, row string
	if err := setup.QueryRow(context.Background(), `SELECT id::text, md5(e::text) FROM audit.events e
WHERE action = 'organization.created' AND organization_id = $1::text::uuid`, w.orgA).Scan(&event, &row); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(map[string]string{"world": mustJSON(t, w), "event": event, "row": row})
	fmt.Println("AUDIT-PERSIST " + string(b))
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestAuditPersistVerify(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	var ids map[string]string
	var w rlsWorld
	if err := json.Unmarshal([]byte(os.Getenv("AUDIT_PERSIST_IDS")), &ids); err != nil || ids["event"] == "" {
		t.Fatalf("AUDIT_PERSIST_IDS: %v", err)
	}
	if err := json.Unmarshal([]byte(ids["world"]), &w); err != nil || w.orgA == "" {
		t.Fatalf("AUDIT_PERSIST_IDS world: %v", err)
	}
	r := runnerFor(t, cfg, auditFS(t))
	if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t8\n" {
		t.Fatalf("version after restart: %q", out)
	}
	if v := auditViolations(t, catalogConn(t, cfg), cfg); len(v) != 0 {
		t.Fatalf("detector after restart %v", v)
	}
	var row string
	if err := rlsSetupConn(t, cfg).QueryRow(ctx, `SELECT md5(e::text) FROM audit.events e WHERE id = $1::text::uuid`, ids["event"]).Scan(&row); err != nil || row != ids["row"] {
		t.Fatalf("event after restart: %q (%v)", row, err)
	}
	rt := rlsRuntimeConn(t, cfg)
	tx, err := rt.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rlsContext(t, tx, w.userA, w.orgA, "", "")
	var n int
	if err := tx.QueryRow(ctx, `SELECT count(*)::int FROM audit.events WHERE id = $1::text::uuid`, ids["event"]).Scan(&n); err != nil || n != 1 {
		t.Fatalf("the Organization owner reads the event %d times after restart (%v)", n, err)
	}
}
