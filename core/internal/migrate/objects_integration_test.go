//go:build integration

// G009 B-OBJ-CATALOG and B-OBJ-MIGRATION (Plan r1 §3, §5.3, §5.2 TestObjectsMigration) against the disposable
// PostgreSQL of the migrations suite: TestObjectsCatalog pins the exact version-7 delta on top of the unchanged
// version-6 catalog, with eight rolled-back missing-protection detectors (N-OBJ-1..8); TestObjectsMigration covers
// empty to 7, a populated G008 world 6 to 7, repeat up, the down refusal with rows, down and up on an empty table and
// the exact v6 <-> v7 catalogs.
package migrate

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"testing"
)

// The version-7 catalog: the version-6 pins with exactly the stored-object delta (Plan r1 §3.2, §3.5).
var (
	objWantFlags    = strings.Replace(rlsWantFlags, "sessions:t:t,", "sessions:t:t,stored_objects:t:t,", 1)
	objWantPolicies = strings.Replace(rlsWantPolicies, "sessions_update:sessions:w:true:ow+rt;",
		"sessions_update:sessions:w:true:ow+rt;stored_objects_insert:stored_objects:a:true:ow+rt;"+
			"stored_objects_select:stored_objects:r:true:ow+rt;stored_objects_update:stored_objects:w:true:ow+rt;", 1)
	objWantACL = strings.Replace(rlsWantACL, "sessions.user_id:ar:SELECT;",
		"sessions.user_id:ar:SELECT;stored_objects:rt:SELECT;stored_objects.byte_size:rt:INSERT;stored_objects.content_type:rt:INSERT;"+
			"stored_objects.creator_user_id:rt:INSERT;stored_objects.group_id:rt:INSERT;stored_objects.id:rt:INSERT;"+
			"stored_objects.organization_id:rt:INSERT;stored_objects.sha256:rt:UPDATE;stored_objects.state:rt:UPDATE;", 1)
	objWantFunctions = rlsWantFunctions + ";stored_objects_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-"
	// The relation: columns (name:type:notnull:generated), constraints, the trigger and the index.
	objWantColumns = "id:uuid:t:;organization_id:uuid:t:;group_id:uuid:f:;creator_user_id:uuid:t:;storage_key:text:f:s;" +
		"state:text:t:;byte_size:bigint:t:;content_type:text:t:;sha256:bytea:f:;created_at:timestamp with time zone:t:;" +
		"ready_at:timestamp with time zone:f:;deleting_at:timestamp with time zone:f:;deleted_at:timestamp with time zone:f:;" +
		"deleted_by:uuid:f:"
	objWantConstraints = "stored_objects_content_type_valid:c;stored_objects_creator_fkey:f;stored_objects_deleted_by_fkey:f;" +
		"stored_objects_group_fkey:f;stored_objects_id_nonzero:c;stored_objects_lifecycle_check:c;stored_objects_organization_fkey:f;" +
		"stored_objects_pkey:p;stored_objects_sha256_valid:c;stored_objects_size_valid:c;stored_objects_state_valid:c"
	objWantTrigger = "stored_objects_guard:O:BEFORE UPDATE:stored_objects_guard"
	objWantIndexes = "stored_objects_organization_group_idx:organization_id, group_id;stored_objects_pkey:id"
	objWantKey     = "('objects/'::text || (organization_id)::text) || '/'::text) || (id)::text)"
	// objWantGuardSrc is md5(prosrc) of app.stored_objects_guard(), the exact 00007 body (S-review finding 1): the
	// creator-only finalize and the same-state refusal are pinned byte for byte, as the older migrate tests pin
	// function sources.
	objWantGuardSrc = "af1143d8f2143696c8462e530893e0e4"
)

// objCatalogParts extends the version-6 parts with the stored-object relation's own parts.
func objCatalogParts(t *testing.T, q catalogQuerier, cfg Config) map[string]string {
	t.Helper()
	p := rlsCatalogParts(t, q, cfg)
	one := func(sql string) string {
		var s *string
		if err := q.QueryRow(context.Background(), sql).Scan(&s); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		if s == nil {
			return ""
		}
		return *s
	}
	p["obj-columns"] = one(`SELECT string_agg(format('%s:%s:%s:%s', a.attname, format_type(a.atttypid, a.atttypmod),
  CASE WHEN a.attnotnull THEN 't' ELSE 'f' END, a.attgenerated::text), ';' ORDER BY a.attnum)
FROM pg_attribute a WHERE a.attrelid = to_regclass('app.stored_objects') AND a.attnum > 0 AND NOT a.attisdropped`)
	p["obj-constraints"] = one(`SELECT string_agg(conname || ':' || contype::text, ';' ORDER BY conname)
FROM pg_constraint WHERE conrelid = to_regclass('app.stored_objects')`)
	p["obj-trigger"] = one(`SELECT string_agg(format('%s:%s:%s:%s', t.tgname, t.tgenabled::text,
  CASE WHEN t.tgtype & 2 = 2 THEN 'BEFORE' ELSE 'AFTER' END || CASE WHEN t.tgtype & 16 = 16 THEN ' UPDATE' ELSE ' OTHER' END, p.proname), ';')
FROM pg_trigger t JOIN pg_proc p ON p.oid = t.tgfoid WHERE t.tgrelid = to_regclass('app.stored_objects') AND NOT t.tgisinternal`)
	p["obj-indexes"] = one(`SELECT string_agg(c.relname || ':' || (SELECT string_agg(a.attname, ', ' ORDER BY k.ord)
  FROM unnest(i.indkey) WITH ORDINALITY k(attnum, ord) JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = k.attnum), ';' ORDER BY c.relname)
FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE i.indrelid = to_regclass('app.stored_objects')`)
	p["obj-guard-src"] = one(`SELECT md5(p.prosrc) FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace AND p.proname = 'stored_objects_guard'`)
	p["obj-key"] = one(`SELECT (SELECT pg_get_expr(d.adbin, d.adrelid) FROM pg_attrdef d JOIN pg_attribute a ON a.attrelid = d.adrelid AND a.attnum = d.adnum
WHERE d.adrelid = to_regclass('app.stored_objects') AND a.attname = 'storage_key')`)
	return p
}

// objViolations is the version-7 detector: the sorted codes of every catalog part that differs from its pin, plus the
// stored-object policy expression checks (each names the context parsers and the Organization reader; none is
// constant). The version-6 policy expressions keep their exact pins.
func objViolations(t *testing.T, q catalogQuerier, cfg Config) []string {
	t.Helper()
	p := objCatalogParts(t, q, cfg)
	want := map[string]string{"rls-flags": objWantFlags, "policy-inventory": objWantPolicies, "acl": objWantACL,
		"functions": objWantFunctions, "schema": rlsWantSchema, "roles": rlsWantRoles, "memberships": rlsWantMemberships,
		"views": rlsWantViews, "obj-columns": objWantColumns, "obj-constraints": objWantConstraints, "obj-trigger": objWantTrigger,
		"obj-indexes": objWantIndexes, "obj-guard-src": objWantGuardSrc}
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
	if !strings.Contains(p["obj-key"], "'objects/'::text") || !strings.Contains(p["obj-key"], "organization_id") || !strings.Contains(p["obj-key"], "(id)::text") {
		codes = append(codes, "obj-key")
		t.Logf("detector obj-key: got %s", p["obj-key"])
	}
	for _, name := range []string{"stored_objects_select", "stored_objects_insert", "stored_objects_update"} {
		e := p["policy-expr:"+name]
		ok := strings.Contains(e, "app.ctx_org()") && strings.Contains(e, "app.rls_org_role()") && strings.Contains(e, "app.rls_group_role(group_id)")
		switch name {
		case "stored_objects_insert":
			ok = ok && strings.Contains(e, "app.ctx_user()") && strings.Contains(e, "'pending'::text")
		case "stored_objects_update":
			ok = ok && strings.Contains(e, "app.ctx_user()") && strings.Count(e, "WITH CHECK") == 1 &&
				normalize(strings.SplitN(e, " WITH CHECK ", 2)[0]) == normalize(strings.SplitN(e, " WITH CHECK ", 2)[1])
		}
		if !ok {
			codes = append(codes, "policy-expr:"+name)
			t.Logf("detector policy-expr:%s: got %s", name, e)
		}
	}
	slices.Sort(codes)
	return slices.Compact(codes)
}

func objRunUp(t *testing.T, r *Runner, want string) {
	t.Helper()
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, want) {
		t.Fatalf("up: %q, want suffix %q", out, want)
	}
}

// TestObjectsCatalog (B-OBJ-CATALOG): the exact version-7 delta and the unchanged version-6 objects, then the eight
// deliberate missing-protection detectors, each rolled back, each making the detector fail.
func TestObjectsCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r := runnerFor(t, cfg, subsetFS(t, 7))
	objRunUp(t, r, "current_version\t7\n")
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)

	t.Run("CT-OBJ-1 the exact version-7 catalog", func(t *testing.T) {
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-7 detector: %v", v)
		}
		metadataThroughVersion7Only(t, cfg, r)
	})
	t.Run("CT-OBJ-2 runtime privileges and denials", func(t *testing.T) {
		rt := rlsRuntimeConn(t, cfg)
		for _, c := range []struct{ what, sql string }{
			{"DELETE", `DELETE FROM app.stored_objects`},
			{"TRUNCATE", `TRUNCATE app.stored_objects`},
			{"UPDATE storage_key", `UPDATE app.stored_objects SET storage_key = DEFAULT`},
			{"UPDATE creator_user_id", `UPDATE app.stored_objects SET creator_user_id = creator_user_id`},
			{"UPDATE organization_id", `UPDATE app.stored_objects SET organization_id = organization_id`},
			{"UPDATE deleted_by", `UPDATE app.stored_objects SET deleted_by = NULL`},
			{"INSERT state", `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type, state) VALUES (gen_random_uuid(), gen_random_uuid(), gen_random_uuid(), 0, 'a/b', 'ready')`},
			{"INSERT created_at", `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type, created_at) VALUES (gen_random_uuid(), gen_random_uuid(), gen_random_uuid(), 0, 'a/b', now())`},
			{"ALTER TABLE", `ALTER TABLE app.stored_objects NO FORCE ROW LEVEL SECURITY`},
			{"DROP POLICY", `DROP POLICY stored_objects_select ON app.stored_objects`},
			{"EXECUTE guard", `SELECT app.stored_objects_guard()`},
		} {
			_, err := rt.Exec(ctx, c.sql)
			if got := sqlstateOf(err); got != "42501" && !(c.what == "EXECUTE guard" && got == "0A000") {
				t.Errorf("runtime %s: %v (want 42501)", c.what, err)
			}
		}
	})
	t.Run("CT-OBJ-NEG rolled-back missing-protection detectors", func(t *testing.T) {
		for _, n := range []struct{ id, sql, code string }{
			{"N-OBJ-1", `ALTER TABLE app.stored_objects NO FORCE ROW LEVEL SECURITY`, "rls-flags"},
			{"N-OBJ-2", `GRANT DELETE ON app.stored_objects TO ` + quote(cfg.RuntimeRole()), "acl"},
			{"N-OBJ-3", `ALTER POLICY stored_objects_select ON app.stored_objects USING (true)`, "policy-expr:stored_objects_select"},
			{"N-OBJ-4", `ALTER TABLE app.stored_objects DISABLE TRIGGER stored_objects_guard`, "obj-trigger"},
			{"N-OBJ-5", `ALTER FUNCTION app.stored_objects_guard() SECURITY DEFINER`, "functions"},
			{"N-OBJ-6", `GRANT SELECT ON app.stored_objects TO PUBLIC`, "acl"},
			{"N-OBJ-7", `ALTER TABLE app.stored_objects ALTER COLUMN storage_key DROP EXPRESSION`, "obj-columns"},
			// A replaced guard body with the same name, language, settings and security (only the source differs).
			{"N-OBJ-8", `CREATE OR REPLACE FUNCTION app.stored_objects_guard() RETURNS trigger LANGUAGE plpgsql ` +
				`SET search_path = pg_catalog, pg_temp AS $$BEGIN RETURN NEW; END$$`, "obj-guard-src"},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			v := objViolations(t, tx, cfg)
			_ = tx.Rollback(ctx)
			if !slices.Contains(v, n.code) {
				t.Errorf("%s: the detector reported %v, want %s", n.id, v, n.code)
			}
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", v)
		}
	})
}

// TestObjectsMigration (B-OBJ-MIGRATION): empty to 7; a populated G008 world from 6 to 7 with every row identical;
// repeat up; down refused while a row exists (version and rows intact); down and up on an empty table; the exact
// version-6 catalog after down and the exact version-7 catalog after up.
func TestObjectsMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	full := "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\n" +
		"applied\t00005_organization_policy.sql\napplied\t00006_row_level_security.sql\napplied\t00007_stored_objects.sql\ncurrent_version\t7\n"

	t.Run("MG-OBJ-1 empty to version 7", func(t *testing.T) {
		reset(t, cfg)
		r := runnerFor(t, cfg, subsetFS(t, 7))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != full {
			t.Fatalf("up: %q", out)
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
		rlsOwnerRestored(t, cfg, r)
	})
	var populated string
	var w rlsWorld
	t.Run("MG-OBJ-2 populated version 6 to 7 keeps every row; repeat up", func(t *testing.T) {
		reset(t, cfg)
		objRunUp(t, rlsRunner6(t, cfg), "current_version\t6\n")
		w = seedRLSWorld(t, setup, "obj")
		populated = rlsRowsSnapshot(t, setup)
		r := runnerFor(t, cfg, subsetFS(t, 7))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00007_stored_objects.sql\ncurrent_version\t7\n" {
			t.Fatalf("up to 7: %q", out)
		}
		if got := rlsRowsSnapshot(t, setup); got != populated {
			t.Fatalf("rows changed by the upgrade:\n got %s\nwant %s", got, populated)
		}
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "no pending migrations\ncurrent_version\t7\n" {
			t.Fatalf("repeat up: %q", out)
		}
		out := run(t, "status", func(ctx context.Context, b *bytes.Buffer) error { return r.Status(ctx, b) })
		if !strings.Contains(out, "migration\t00007_stored_objects.sql\tapplied") || strings.Contains(out, "pending") {
			t.Fatalf("status: %q", out)
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
	})
	t.Run("MG-OBJ-3 down is refused while a row exists", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 7))
		if _, err := setup.Exec(ctx, `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type)
VALUES (gen_random_uuid(), $1::text::uuid, $2::text::uuid, 0, 'application/octet-stream')`, w.orgA, w.userA); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = setup.Exec(ctx, `DELETE FROM app.stored_objects`) })
		rows := func() string {
			var s string
			if err := setup.QueryRow(ctx, `SELECT count(*) || ':' || md5(string_agg(o::text, E'\n' ORDER BY o.id)) FROM app.stored_objects o`).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
		before := rows()
		var buf bytes.Buffer
		err := r.Down(ctx, &buf)
		if err == nil || !strings.Contains(err.Error(), "refusing to roll back 00007") {
			t.Fatalf("down with a row: %v %q", err, buf.String())
		}
		if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t7\n" {
			t.Fatalf("version after the refused down: %q", out)
		}
		if rows() != before || rlsRowsSnapshot(t, setup) != populated {
			t.Fatal("rows changed by the refused down")
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("catalog changed by the refused down: %v", v)
		}
		rlsOwnerRestored(t, cfg, r)
		if _, err := setup.Exec(ctx, `DELETE FROM app.stored_objects`); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("MG-OBJ-4 down and up on an empty table: exact v6 <-> v7 catalogs", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 7))
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out != "rolled_back\t00007_stored_objects.sql\ncurrent_version\t6\n" {
			t.Fatalf("down: %q", out)
		}
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-6 detector after down: %v", v)
		}
		metadataThroughVersion6Only(t, cfg, r)
		if got := rlsRowsSnapshot(t, setup); got != populated {
			t.Fatal("rows changed by down")
		}
		rlsOwnerRestored(t, cfg, r)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00007_stored_objects.sql\ncurrent_version\t7\n" {
			t.Fatalf("up again: %q", out)
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-7 detector after up: %v", v)
		}
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out != "rolled_back\t00007_stored_objects.sql\napplied\t00007_stored_objects.sql\ncurrent_version\t7\n" {
			t.Fatalf("reapply: %q", out)
		}
		if v := objViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-7 detector after reapply: %v", v)
		}
	})
	t.Run("MG-OBJ-5 a failed up rolls back and restores the owner", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 7))
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); !strings.HasSuffix(out, "current_version\t6\n") {
			t.Fatalf("down: %q", out)
		}
		// An unexpected policy the version-6 precondition rejects.
		if _, err := setup.Exec(ctx, `CREATE POLICY obj_unexpected ON app.organizations FOR SELECT TO `+quote(cfg.RuntimeRole())+` USING (false)`); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		err := r.Up(ctx, &buf)
		if _, derr := setup.Exec(ctx, `DROP POLICY obj_unexpected ON app.organizations`); derr != nil {
			t.Fatal(derr)
		}
		if err == nil || !strings.Contains(err.Error(), "version-6 precondition: policy inventory") {
			t.Fatalf("up with an unexpected policy: %v", err)
		}
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-6 catalog after the rolled-back up: %v", v)
		}
		rlsOwnerRestored(t, cfg, r)
		objRunUp(t, r, "current_version\t7\n")
	})
}
