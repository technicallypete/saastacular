//go:build integration

// G013 B-RETENTION-CATALOG (Plan r3 §6.1, §10.1) against the disposable PostgreSQL of the migrations suite.
// TestRetentionCatalog pins the exact version-9 delta on top of the unchanged version-8 catalog (CT-RET-1), the
// retention role's attributes and the unchanged runtime, owner and helper privileges (CT-RET-2), and eight rolled-back
// missing-protection detectors (CT-RET-NEG). TestRetentionMigration covers empty to 9, a populated version-8 world to 9
// with every row and non-delta ACL identical, down to 8 with rows present, up again and reapply, and a rolled-back up.
// Setup and rolled-back controls use the attested superuser; catalog reads use the owner. From version 10 (G014) every
// runner here is bound to the 00001-00009 subset, so these checks keep their exact version-9 meaning.
package migrate

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// The version-9 catalog: the version-8 pins plus the delta of 00009.
var (
	retWantAppPolicies = strings.Replace(strings.Replace(auditWantAppPolicies,
		"events_insert:events:a:true:ow+rt;",
		"events_insert:events:a:true:ow+rt;events_retention_delete:events:d:true:rn;events_retention_select:events:r:true:rn;", 1),
		"stored_objects_insert:stored_objects:a:true:ow+rt;",
		"stored_objects_insert:stored_objects:a:true:ow+rt;stored_objects_retention_delete:stored_objects:d:true:rn;"+
			"stored_objects_retention_select:stored_objects:r:true:rn;", 1)
	retWantACL = strings.Replace(strings.Replace(objWantACL,
		"stored_objects:rt:SELECT;stored_objects.byte_size:rt:INSERT;stored_objects.content_type:rt:INSERT;stored_objects.creator_user_id:rt:INSERT;",
		"stored_objects:rn:DELETE;stored_objects:rt:SELECT;stored_objects.byte_size:rt:INSERT;stored_objects.content_type:rt:INSERT;"+
			"stored_objects.creator_user_id:rt:INSERT;stored_objects.deleted_at:rn:SELECT;", 1),
		"stored_objects.group_id:rt:INSERT;stored_objects.id:rt:INSERT;stored_objects.organization_id:rt:INSERT;stored_objects.sha256:rt:UPDATE;stored_objects.state:rt:UPDATE;",
		"stored_objects.group_id:rt:INSERT;stored_objects.id:rn:SELECT;stored_objects.id:rt:INSERT;stored_objects.organization_id:rn:SELECT;"+
			"stored_objects.organization_id:rt:INSERT;stored_objects.sha256:rt:UPDATE;stored_objects.state:rn:SELECT;stored_objects.state:rt:UPDATE;", 1)
	retWantFunctions = strings.Replace(objWantFunctions, "organizations_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;",
		"organizations_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;retention_eligible:ow:false:false:s:search_path=pg_catalog, pg_temp:rn=EXECUTE;", 1)
	retWantAuditPolicies = "events_insert:a:true:ow+rt;events_retention_delete:d:true:rn;events_retention_select:r:true:rn;events_select:r:true:ow+rt"
	retWantAuditACL      = strings.Replace(strings.Replace(auditWantACL,
		"events:rt:SELECT;", "events:rn:DELETE;events:rt:SELECT;", 1),
		"events.group_id:rt:INSERT;events.operation:rt:INSERT;events.organization_id:rt:INSERT;",
		"events.group_id:rt:INSERT;events.id:rn:SELECT;events.occurred_at:rn:SELECT;events.operation:rt:INSERT;events.organization_id:rn:SELECT;events.organization_id:rt:INSERT;", 1)
	retWantAuditSchema = "rn=USAGE,rt=USAGE"
	// md5 of the two reviewed bodies (the refusal body with the embedded role name normalized to <db>_retention), as
	// 00009's postconditions check them.
	retWantRefuseSrc   = "c6c0b51ae2876ba67658dc4041eec5e9"
	retWantEligibleSrc = "89c969a9ec364eae864aab3e9fda0d58"
	retWantAuditSrc    = strings.Replace(auditWantSrc, "events_refuse=ee436a151f6bcf4bf893747a8cf243f9", "events_refuse="+retWantRefuseSrc, 1)
	// The four retention policy expressions (pg_get_expr, normalized; no WITH CHECK).
	retWantEventsExpr  = "app.retention_eligible(organization_id, occurred_at, ARRAY[organization_id, group_id, actor_user_id, subject_user_id, resource_id]) WITH CHECK"
	retWantObjectsExpr = "((state = 'deleted'::text) AND app.retention_eligible(organization_id, deleted_at, ARRAY[organization_id, group_id, id, creator_user_id, deleted_by])) WITH CHECK"
)

// retentionParts reads every version-8 part plus the version-9 extras: the normalized audit sources, the eligibility
// function body and the two stored-object retention policy expressions.
func retentionParts(t *testing.T, q catalogQuerier, cfg Config) map[string]string {
	t.Helper()
	p := objCatalogParts(t, q, cfg)
	for k, v := range auditParts(t, q, cfg) {
		p[k] = v
	}
	one := func(sql string, args ...any) string {
		var s *string
		if err := q.QueryRow(context.Background(), sql, args...).Scan(&s); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		if s == nil {
			return ""
		}
		return *s
	}
	p["audit-src"] = one(`SELECT string_agg(p.proname || '=' || md5(replace(p.prosrc, $1, '<db>_retention')), ';' ORDER BY p.proname)
FROM pg_proc p WHERE p.pronamespace = 'audit'::regnamespace`, cfg.RetentionRole())
	p["ret-eligible-src"] = one(`SELECT md5(p.prosrc) FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace AND p.proname = 'retention_eligible'`)
	return p
}

// retentionViolations is the version-9 detector: the sorted codes of every part that differs from its pin.
func retentionViolations(t *testing.T, q catalogQuerier, cfg Config) []string {
	t.Helper()
	p := retentionParts(t, q, cfg)
	want := map[string]string{"rls-flags": objWantFlags, "policy-inventory": retWantAppPolicies, "acl": retWantACL,
		"functions": retWantFunctions, "schema": rlsWantSchema, "roles": rlsWantRoles, "memberships": rlsWantMemberships,
		"views": rlsWantViews, "obj-columns": objWantColumns, "obj-constraints": objWantConstraints, "obj-indexes": objWantIndexes,
		"obj-guard-src": objWantGuardSrc, "audit-relation": auditWantRelations, "audit-columns": auditWantColumns,
		"audit-constraints": auditWantConstraints, "audit-indexes": auditWantIndexes, "audit-policies": retWantAuditPolicies,
		"audit-acl": retWantAuditACL, "audit-schema": retWantAuditSchema, "audit-functions": auditWantFunctions,
		"audit-triggers": auditWantTriggers, "audit-capture": auditWantCapture, "audit-src": retWantAuditSrc,
		"ret-eligible-src": retWantEligibleSrc,
		"policy-expr:stored_objects_retention_select": retWantObjectsExpr, "policy-expr:stored_objects_retention_delete": retWantObjectsExpr,
		"audit-policy-expr:events_select": auditWantSelectExpr, "audit-policy-expr:events_insert": auditWantInsertExpr,
		"audit-policy-expr:events_retention_select": retWantEventsExpr, "audit-policy-expr:events_retention_delete": retWantEventsExpr}
	for name, expr := range rlsWantExprs {
		want["policy-expr:"+name] = expr
	}
	var codes []string
	for k, w := range want {
		if p[k] != w {
			codes = append(codes, k)
			t.Logf("detector %s: got %s", k, p[k])
		}
	}
	slices.Sort(codes)
	return slices.Compact(codes)
}

func retentionUp(t *testing.T, cfg Config) *Runner {
	t.Helper()
	r := runnerFor(t, cfg, subsetFS(t, 9))
	objRunUp(t, r, "current_version\t9\n")
	return r
}

// TestRetentionCatalog (B-RETENTION-CATALOG): the exact version-9 delta, the role and privilege boundary, and the
// rolled-back detectors.
func TestRetentionCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r := retentionUp(t, cfg)
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)

	t.Run("CT-RET-1 the exact version-9 delta", func(t *testing.T) {
		if v := retentionViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-9 detector: %v", v)
		}
		metadataThroughVersion9Only(t, cfg, r)
		// Everything outside the delta equals version 8: the version-8 detector reports exactly the delta's parts.
		got := auditViolations(t, cat, cfg)
		want := []string{"acl", "audit-acl", "audit-policies", "audit-schema", "audit-src", "functions", "policy-inventory"}
		if !slices.Equal(got, want) {
			t.Errorf("version-8 detector at version 9: %v, want exactly the delta %v", got, want)
		}
	})

	t.Run("CT-RET-2 role attributes and the unchanged privileges", func(t *testing.T) {
		p := retentionParts(t, cat, cfg)
		if p["roles"] != rlsWantRoles || p["memberships"] != rlsWantMemberships {
			t.Errorf("roles %s; memberships %s", p["roles"], p["memberships"])
		}
		// The retention role: DELETE only on the two tables, no INSERT, UPDATE, TRUNCATE, REFERENCES or TRIGGER anywhere,
		// SELECT only on its columns; runtime and PUBLIC cannot execute the eligibility function.
		var bad string
		if err := setup.QueryRow(ctx, `
SELECT coalesce(string_agg(format('%s:%s', c.oid::regclass, p), ',' ORDER BY 1), '')
FROM pg_class c CROSS JOIN unnest(ARRAY['INSERT','UPDATE','TRUNCATE','REFERENCES','TRIGGER','DELETE','SELECT']) p
WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND c.relkind IN ('r', 'v')
  AND has_table_privilege($1, c.oid, p)
  AND NOT (p = 'DELETE' AND c.oid IN ('audit.events'::regclass, 'app.stored_objects'::regclass))`, cfg.RetentionRole()).Scan(&bad); err != nil || bad != "" {
			t.Errorf("retention table privileges beyond DELETE on the two tables: %q (%v)", bad, err)
		}
		var cols string
		if err := setup.QueryRow(ctx, `
SELECT string_agg(format('%s.%s', c.relname, a.attname), ',' ORDER BY c.relname, a.attname)
FROM pg_class c JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped
WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND has_column_privilege($1, c.oid, a.attnum, 'SELECT')`,
			cfg.RetentionRole()).Scan(&cols); err != nil || cols != "events.id,events.occurred_at,events.organization_id,stored_objects.deleted_at,stored_objects.id,stored_objects.organization_id,stored_objects.state" {
			t.Errorf("retention readable columns: %q (%v)", cols, err)
		}
		var rtExec, rnExec bool
		if err := setup.QueryRow(ctx, `SELECT has_function_privilege($1, 'app.retention_eligible(uuid, timestamptz, uuid[])', 'EXECUTE'),
  has_function_privilege($2, 'app.retention_eligible(uuid, timestamptz, uuid[])', 'EXECUTE')`, cfg.RuntimeRole(), cfg.RetentionRole()).Scan(&rtExec, &rnExec); err != nil || rtExec || !rnExec {
			t.Errorf("eligibility EXECUTE: runtime %v, retention %v (%v)", rtExec, rnExec, err)
		}
		// Runtime and owner: still no DELETE on audit history or tombstones beyond what version 8 grants.
		var rtDel bool
		if err := setup.QueryRow(ctx, `SELECT has_table_privilege($1, 'audit.events', 'DELETE') OR has_table_privilege($1, 'app.stored_objects', 'DELETE')`,
			cfg.RuntimeRole()).Scan(&rtDel); err != nil || rtDel {
			t.Errorf("runtime DELETE on audit or tombstones (%v)", err)
		}
	})

	t.Run("CT-RET-NEG rolled-back missing-protection detectors", func(t *testing.T) {
		rn, rt := quote(cfg.RetentionRole()), quote(cfg.RuntimeRole())
		for _, n := range []struct {
			id, sql string
			want    []string
		}{
			{"N1 retention UPDATE grant", `GRANT UPDATE ON audit.events TO ` + rn, []string{"audit-acl"}},
			{"N2 retention INSERT grant", `GRANT INSERT ON audit.events TO ` + rn, []string{"audit-acl"}},
			{"N3 retention TRUNCATE grant", `GRANT TRUNCATE ON app.stored_objects TO ` + rn, []string{"acl"}},
			{"N4 retention policy also TO runtime", `ALTER POLICY events_retention_delete ON audit.events TO ` + rn + `, ` + rt,
				[]string{"audit-policies", "policy-inventory"}},
			{"N5 refusal without the role check", `CREATE OR REPLACE FUNCTION audit.events_refuse() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $b$
BEGIN
  IF TG_OP = 'DELETE' THEN
    IF app.retention_eligible(OLD.organization_id, OLD.occurred_at,
         ARRAY[OLD.organization_id, OLD.group_id, OLD.actor_user_id, OLD.subject_user_id, OLD.resource_id]) THEN
      RETURN OLD;
    END IF;
  END IF;
  RAISE EXCEPTION 'audit events are append-only' USING ERRCODE = 'insufficient_privilege';
END $b$`, []string{"audit-src"}},
			{"N6 SECURITY DEFINER eligibility", `ALTER FUNCTION app.retention_eligible(uuid, timestamptz, uuid[]) SECURITY DEFINER`, []string{"functions"}},
			{"N7 PUBLIC EXECUTE", `GRANT EXECUTE ON FUNCTION app.retention_eligible(uuid, timestamptz, uuid[]) TO PUBLIC`, []string{"functions"}},
			{"N8 retention policy on another app table", `CREATE POLICY users_retention_probe ON app.users FOR SELECT TO ` + rn + ` USING (true)`,
				[]string{"policy-inventory"}},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			v := retentionViolations(t, tx, cfg)
			_ = tx.Rollback(ctx)
			if !slices.Equal(v, n.want) {
				t.Errorf("%s: the detector reported %v, want %v", n.id, v, n.want)
			}
		}
		if v := retentionViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", v)
		}
	})
}

// retSeed writes a version-8 world (two Organizations through seedRLSWorld, whose capture triggers record history) plus
// a deleted tombstone, as the attested superuser.
func retSeed(t *testing.T, setup *pgx.Conn) {
	t.Helper()
	w := seedRLSWorld(t, setup, "ret")
	if _, err := setup.Exec(context.Background(), `INSERT INTO app.stored_objects (id, organization_id, group_id, creator_user_id, state,
  byte_size, content_type, sha256, ready_at, deleting_at, deleted_at, deleted_by)
VALUES (gen_random_uuid(), $1::text::uuid, NULL, $2::text::uuid, 'deleted', 1, 'text/plain', sha256('x'::bytea), now(), now(), now(), $2::text::uuid)`,
		w.orgA, w.userA); err != nil {
		t.Fatalf("seed tombstone: %v", err)
	}
}

func retRows(t *testing.T, setup *pgx.Conn) string {
	t.Helper()
	var h string
	if err := setup.QueryRow(context.Background(), `SELECT count(*) || ':' || coalesce(md5(string_agg(e::text, E'\n' ORDER BY e.id)), '-') FROM audit.events e`).Scan(&h); err != nil {
		t.Fatal(err)
	}
	return auditRowsSnapshot(t, setup) + ";events=" + h
}

// TestRetentionMigration (B-RETENTION-CATALOG): MG-RET-1..5.
func TestRetentionMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	full := "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\n" +
		"applied\t00005_organization_policy.sql\napplied\t00006_row_level_security.sql\napplied\t00007_stored_objects.sql\n" +
		"applied\t00008_audit_events.sql\napplied\t00009_retention_maintenance.sql\ncurrent_version\t9\n"
	up := func(r *Runner) string {
		return run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) })
	}

	t.Run("MG-RET-1 empty to version 9", func(t *testing.T) {
		reset(t, cfg)
		if out := up(runnerFor(t, cfg, subsetFS(t, 9))); out != full {
			t.Fatalf("up: %q", out)
		}
		if v := retentionViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
	})

	var rows string
	t.Run("MG-RET-2 a populated version-8 world to 9", func(t *testing.T) {
		reset(t, cfg)
		objRunUp(t, runnerFor(t, cfg, subsetFS(t, 8)), "current_version\t8\n")
		retSeed(t, setup)
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-8 detector before the upgrade: %v", v)
		}
		rows = retRows(t, setup)
		r := runnerFor(t, cfg, subsetFS(t, 9))
		if out := up(r); out != "applied\t00009_retention_maintenance.sql\ncurrent_version\t9\n" {
			t.Fatalf("up: %q", out)
		}
		if v := retentionViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
		if got := retRows(t, setup); got != rows {
			t.Fatal("rows changed by the upgrade")
		}
		if out := up(r); out != "no pending migrations\ncurrent_version\t9\n" {
			t.Fatalf("repeat up: %q", out)
		}
	})

	t.Run("MG-RET-3 down to 8 with rows present restores the exact version-8 catalog", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 9))
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out != "rolled_back\t00009_retention_maintenance.sql\ncurrent_version\t8\n" {
			t.Fatalf("down: %q", out)
		}
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-8 detector after down: %v", v)
		}
		if got := retRows(t, setup); got != rows {
			t.Fatal("rows changed by the rollback")
		}
	})

	t.Run("MG-RET-4 up again and reapply", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 9))
		if out := up(r); out != "applied\t00009_retention_maintenance.sql\ncurrent_version\t9\n" {
			t.Fatalf("up: %q", out)
		}
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out != "rolled_back\t00009_retention_maintenance.sql\napplied\t00009_retention_maintenance.sql\ncurrent_version\t9\n" {
			t.Fatalf("reapply: %q", out)
		}
		if v := retentionViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
		if got := retRows(t, setup); got != rows {
			t.Fatal("rows changed by up/reapply")
		}
	})

	t.Run("MG-RET-5 a rolled-back up leaves version 8", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 9))
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); !strings.HasSuffix(out, "current_version\t8\n") {
			t.Fatalf("down: %q", out)
		}
		// A stray retention grant breaks the version-8 precondition: the up fails and rolls back entirely.
		if _, err := setup.Exec(ctx, `GRANT SELECT ON audit.events TO `+quote(cfg.RetentionRole())); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := r.Up(ctx, &buf); err == nil || !strings.Contains(err.Error(), "version-8 precondition") {
			t.Fatalf("up with a stray grant: %v %q", err, buf.String())
		}
		if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t8\n" {
			t.Fatalf("version after the failed up: %q", out)
		}
		if v := auditViolations(t, cat, cfg); !slices.Equal(v, []string{"audit-acl"}) {
			t.Errorf("after the failed up the detector reports %v, want only the stray grant", v)
		}
		if _, err := setup.Exec(ctx, `REVOKE SELECT ON audit.events FROM `+quote(cfg.RetentionRole())); err != nil {
			t.Fatal(err)
		}
		if v := auditViolations(t, cat, cfg); len(v) != 0 {
			t.Errorf("version-8 detector: %v", v)
		}
		rlsOwnerRestored(t, cfg, r)
		if got := retRows(t, setup); got != rows {
			t.Fatal("rows changed")
		}
		objRunUp(t, r, "current_version\t9\n")
	})
}
