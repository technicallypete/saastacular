//go:build integration

// G017 B-BRANDING-CATALOG (Plan r2 §4, §6.2): TestBrandingCatalog. CT-BRD-1 the exact version-13 delta (table, validator,
// guard, capture, runtime policies, both auth-reader policies, the definer projection, grants) with the function bodies
// equal to the reviewed migration; CT-BRD-NEG rolled-back missing-protection negatives, each reported by the detector
// and (where a statement applies) admitting what the protection refuses; CT-BRD-UPDOWN the down to the exact
// version-12 catalog when empty, up again, and the down refused while a branding row or branding history exists.
package migrate

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"core/migrations"
)

// brdBody returns the dollar-quoted body (prosrc) of the first function created after marker in the 00013 up section.
func brdBody(t *testing.T, marker string) string {
	t.Helper()
	b, err := fs.ReadFile(migrations.FS, "00013_organization_branding.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, _ := strings.Cut(string(b), "-- +goose Down")
	i := strings.Index(up, marker)
	if i < 0 {
		t.Fatalf("00013: %q not found", marker)
	}
	j := strings.Index(up[i:], "AS $$")
	if j < 0 {
		t.Fatalf("00013: no body after %q", marker)
	}
	start := i + j + len("AS $$")
	k := strings.Index(up[start:], "$$")
	return up[start : start+k]
}

// brdParts reads the version-13 delta's catalog parts.
func brdParts(t *testing.T, q catalogQuerier, cfg Config) map[string]string {
	t.Helper()
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
	p := map[string]string{}
	p["brd-columns"] = one(`SELECT string_agg(format('%s:%s:%s:%s', a.attname, format_type(a.atttypid, a.atttypmod),
  CASE WHEN a.attnotnull THEN 't' ELSE 'f' END, coalesce(pg_get_expr(d.adbin, d.adrelid), '')), ';' ORDER BY a.attnum)
FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = to_regclass('app.organization_branding') AND a.attnum > 0 AND NOT a.attisdropped`)
	p["brd-constraints"] = one(`SELECT string_agg(conname || ':' || pg_get_constraintdef(oid), ';' ORDER BY conname) FROM pg_constraint
WHERE conrelid = to_regclass('app.organization_branding')`)
	p["brd-indexes"] = one(`SELECT string_agg(pg_get_indexdef(i.indexrelid), ';' ORDER BY c.relname) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
WHERE i.indrelid = to_regclass('app.organization_branding')`)
	p["brd-triggers"] = one(`SELECT string_agg(t.tgname || ':' || t.tgenabled::text || ':' || pg_get_triggerdef(t.oid), E'\n' ORDER BY t.tgname)
FROM pg_trigger t WHERE t.tgrelid = to_regclass('app.organization_branding') AND NOT t.tgisinternal`)
	p["brd-flags"] = one(`SELECT relrowsecurity::text || ':' || relforcerowsecurity::text FROM pg_class WHERE oid = to_regclass('app.organization_branding')`)
	p["brd-policies"] = one(`SELECT string_agg(p.polname || ':' || p.polcmd::text || ':' || array_to_string(ARRAY(SELECT r::regrole::text FROM unnest(p.polroles) r ORDER BY 1), '+') ||
  ':' || regexp_replace(coalesce(pg_get_expr(p.polqual, p.polrelid), ''), '\s+', ' ', 'g') || ':' ||
  regexp_replace(coalesce(pg_get_expr(p.polwithcheck, p.polrelid), ''), '\s+', ' ', 'g'), E'\n' ORDER BY p.polname)
FROM pg_policy p WHERE p.polrelid = to_regclass('app.organization_branding') OR p.polname = 'stored_objects_auth_reader'`)
	// Table and column privileges on the branding table, and the auth reader's column privileges on stored_objects.
	p["brd-acl"] = one(`SELECT string_agg(x, ';' ORDER BY x) FROM (
  SELECT c.relname || ':' || a.grantee::regrole::text || ':' || a.privilege_type AS x
  FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
  WHERE c.oid = to_regclass('app.organization_branding') AND a.grantee <> c.relowner
  UNION ALL
  SELECT c.relname || '.' || at.attname || ':' || a.grantee::regrole::text || ':' || a.privilege_type
  FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
  CROSS JOIN LATERAL aclexplode(at.attacl) a
  WHERE (c.oid = to_regclass('app.organization_branding') OR (c.oid = to_regclass('app.stored_objects') AND a.grantee::regrole::text = $1))
    AND a.grantee <> c.relowner) s`, cfg.Database+"_auth_reader")
	p["brd-functions"] = one(`SELECT string_agg(p.pronamespace::regnamespace::text || '.' || p.proname || ':' || p.proowner::regrole::text || ':' ||
  p.prosecdef::text || ':' || coalesce(array_to_string(p.proacl, ','), '-') || ':' || coalesce(array_to_string(p.proconfig, '|'), '-'), ';'
  ORDER BY p.pronamespace::regnamespace::text, p.proname)
FROM pg_proc p WHERE p.oid IN (to_regprocedure('app.branding_overrides_valid(jsonb)'), to_regprocedure('app.organization_branding_guard()'),
  to_regprocedure('app.public_branding(uuid)'), to_regprocedure('audit.capture_organization_branding()'))`)
	p["brd-attribute-src"] = one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure('audit.events_attribute()')`)
	for name, sig := range map[string]string{"brd-validator-src": "app.branding_overrides_valid(jsonb)", "brd-guard-src": "app.organization_branding_guard()",
		"brd-reader-src": "app.public_branding(uuid)", "brd-capture-src": "audit.capture_organization_branding()"} {
		p[name] = one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure('` + sig + `')`)
	}
	return p
}

// brdViolations is the version-13 delta detector: the sorted parts that differ from want.
func brdViolations(t *testing.T, q catalogQuerier, cfg Config, want map[string]string) []string {
	t.Helper()
	p := brdParts(t, q, cfg)
	var codes []string
	for k, w := range want {
		if p[k] != w {
			codes = append(codes, k)
			t.Logf("detector %s: got %s", k, p[k])
		}
	}
	slices.Sort(codes)
	return codes
}

func TestBrandingCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r12 := runnerFor(t, cfg, subsetFS(t, 12))
	objRunUp(t, r12, "applied\t00012_session_binding_and_claim_exclusivity.sql\ncurrent_version\t12\n")
	s12 := snapshot(t, r12)
	r := runnerFor(t, cfg, subsetFS(t, 13))
	objRunUp(t, r, "applied\t00013_organization_branding.sql\ncurrent_version\t13\n")
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	clean := brdParts(t, cat, cfg)
	o, rt, ar := cfg.OwnerRole(), cfg.RuntimeRole(), cfg.Database+"_auth_reader"

	t.Run("CT-BRD-1 the exact version-13 delta", func(t *testing.T) {
		metadataThroughVersion13Only(t, cfg, r)
		for part, marker := range map[string]string{"brd-validator-src": "CREATE FUNCTION app.branding_overrides_valid(",
			"brd-guard-src": "CREATE FUNCTION app.organization_branding_guard()", "brd-reader-src": "CREATE FUNCTION app.public_branding(",
			"brd-capture-src": "CREATE FUNCTION audit.capture_organization_branding()"} {
			if clean[part] != md5hex(brdBody(t, marker)) {
				t.Errorf("%s differs from the reviewed migration body", part)
			}
		}
		if got, want := clean["brd-columns"], "organization_id:uuid:t:;display_label:text:f:;logo_object_id:uuid:f:;light:jsonb:t:'{}'::jsonb;"+
			"dark:jsonb:t:'{}'::jsonb;version:bigint:t:1;updated_at:timestamp with time zone:t:now();updated_by:uuid:t:"; got != want {
			t.Errorf("columns:\n got %s\nwant %s", got, want)
		}
		if got := clean["brd-flags"]; got != "true:true" {
			t.Errorf("row-level security flags: %s, want enabled and forced", got)
		}
		for _, c := range []string{"organization_branding_dark_valid:", "organization_branding_label_valid:", "organization_branding_light_valid:",
			"organization_branding_organization_fkey:", "organization_branding_pkey:", "organization_branding_updated_by_fkey:",
			"organization_branding_version_valid:"} {
			if !strings.Contains(clean["brd-constraints"], c) {
				t.Errorf("constraint %s missing: %s", c, clean["brd-constraints"])
			}
		}
		if !strings.Contains(clean["brd-constraints"], "IS NFC NORMALIZED") || strings.Contains(clean["brd-constraints"], "stored_objects") {
			t.Errorf("constraints: the label is not NFC-checked or a stored-object foreign key exists: %s", clean["brd-constraints"])
		}
		if got := strings.Count(clean["brd-triggers"], "\n") + 1; got != 2 || !strings.Contains(clean["brd-triggers"], "audit_capture_branding:O:CREATE TRIGGER audit_capture_branding AFTER INSERT OR UPDATE") ||
			!strings.Contains(clean["brd-triggers"], "organization_branding_guard:O:CREATE TRIGGER organization_branding_guard BEFORE INSERT OR UPDATE") {
			t.Errorf("triggers: %s", clean["brd-triggers"])
		}
		// The runtime policies carry the Organization context and the owner/admin predicate (the insert also version 1);
		// the auth reader's two policies are exactly "every branding row" and "only some Organization's current logo".
		pols := strings.Split(clean["brd-policies"], "\n")
		if len(pols) != 5 {
			t.Fatalf("policies: %q", clean["brd-policies"])
		}
		for _, l := range pols {
			name, _, _ := strings.Cut(l, ":")
			switch name {
			case "organization_branding_select", "organization_branding_insert", "organization_branding_update":
				if !strings.Contains(l, ":"+o+"+"+rt+":") && !strings.Contains(l, ":"+rt+"+"+o+":") ||
					strings.Count(l, "'org_owner'") < 1 || strings.Count(l, "'org_admin'") < 1 || !strings.Contains(l, "rls_org_role()") ||
					!strings.Contains(l, "ctx_org()") {
					t.Errorf("policy %s lacks the Organization owner/admin predicate or its roles: %s", name, l)
				}
				if name == "organization_branding_insert" && !strings.Contains(l, "version = 1") {
					t.Errorf("insert policy without version 1: %s", l)
				}
			case "organization_branding_auth_reader":
				if l != "organization_branding_auth_reader:r:"+ar+":true:" {
					t.Errorf("auth reader branding policy: %s", l)
				}
			case "stored_objects_auth_reader":
				if !strings.HasPrefix(l, "stored_objects_auth_reader:r:"+ar+":(EXISTS ( SELECT 1") || !strings.Contains(l, "b.logo_object_id = stored_objects.id") ||
					!strings.Contains(l, "b.organization_id = stored_objects.organization_id") {
					t.Errorf("auth reader stored-object policy: %s", l)
				}
			default:
				t.Errorf("unexpected policy %s", l)
			}
		}
		var acl []string
		acl = append(acl, "organization_branding:"+rt+":SELECT")
		for _, c := range []string{"organization_id", "display_label", "logo_object_id", "light", "dark", "version", "updated_by"} {
			acl = append(acl, "organization_branding."+c+":"+rt+":INSERT")
		}
		for _, c := range []string{"display_label", "logo_object_id", "light", "dark", "version", "updated_at", "updated_by"} {
			acl = append(acl, "organization_branding."+c+":"+rt+":UPDATE")
		}
		for _, c := range []string{"organization_id", "display_label", "logo_object_id", "light", "dark", "version"} {
			acl = append(acl, "organization_branding."+c+":"+ar+":SELECT")
		}
		for _, c := range []string{"id", "organization_id", "state", "content_type", "byte_size", "sha256", "storage_key"} {
			acl = append(acl, "stored_objects."+c+":"+ar+":SELECT")
		}
		slices.Sort(acl)
		if got := clean["brd-acl"]; got != strings.Join(acl, ";") {
			t.Errorf("privileges:\n got %s\nwant %s", got, strings.Join(acl, ";"))
		}
		sp := "search_path=pg_catalog, pg_temp"
		if got, want := clean["brd-functions"], fmt.Sprintf("app.branding_overrides_valid:%[1]s:false:%[1]s=X/%[1]s,%[2]s=X/%[1]s:%[4]s;"+
			"app.organization_branding_guard:%[1]s:false:%[1]s=X/%[1]s:%[4]s;app.public_branding:%[3]s:true:%[3]s=X/%[3]s,%[2]s=X/%[3]s:%[4]s;"+
			"audit.capture_organization_branding:%[1]s:false:%[1]s=X/%[1]s:%[4]s", o, rt, ar, sp); got != want {
			t.Errorf("functions:\n got %s\nwant %s", got, want)
		}
	})

	t.Run("CT-BRD-NEG rolled-back missing-protection negatives", func(t *testing.T) {
		seed := `INSERT INTO app.users (id, display_name) VALUES ('0f17c000-0000-4000-8000-0000000000c1', 'Catalog');
INSERT INTO app.organizations (id, name) VALUES ('0f17c000-0000-4000-8000-0000000000c2', 'Catalog');
INSERT INTO app.organization_branding (organization_id, display_label, updated_by)
VALUES ('0f17c000-0000-4000-8000-0000000000c2', 'Catalog', '0f17c000-0000-4000-8000-0000000000c1')`
		versionJump := `UPDATE app.organization_branding SET version = 5 WHERE organization_id = '0f17c000-0000-4000-8000-0000000000c2'`
		foreignLogo := `UPDATE app.organization_branding SET logo_object_id = gen_random_uuid(), version = version + 1
WHERE organization_id = '0f17c000-0000-4000-8000-0000000000c2'`
		badLabel := `UPDATE app.organization_branding SET display_label = E'Cat\nalog', version = version + 1
WHERE organization_id = '0f17c000-0000-4000-8000-0000000000c2'`
		probe := func(id, mutation, stmt string) error {
			t.Helper()
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			for _, s := range []string{mutation, seed} {
				if s == "" {
					continue
				}
				if _, err := tx.Exec(ctx, s); err != nil {
					t.Fatalf("%s setup: %v", id, err)
				}
			}
			_, err = tx.Exec(ctx, stmt)
			return err
		}
		sqlState := func(err error) string {
			if pg, ok := err.(*pgconn.PgError); ok {
				return pg.Code
			}
			return ""
		}
		for what, stmt := range map[string]string{"version jump": versionJump, "unknown logo": foreignLogo, "control in label": badLabel} {
			if err := probe("intact", "", stmt); sqlState(err) != "23514" {
				t.Fatalf("intact protections: %s %v, want 23514", what, err)
			}
		}
		for _, n := range []struct {
			id, sql string
			want    []string
			stmt    string // a statement the missing protection now admits (or "")
		}{
			{"N1 select policy without the owner/admin predicate", `ALTER POLICY organization_branding_select ON app.organization_branding USING (` +
				`organization_id = (SELECT app.ctx_org()))`, []string{"brd-policies"}, ""},
			{"N2 guard trigger missing", `DROP TRIGGER organization_branding_guard ON app.organization_branding`, []string{"brd-triggers"}, versionJump},
			{"N3 guard admitting any reference", `CREATE OR REPLACE FUNCTION app.organization_branding_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $f$ BEGIN RETURN NEW; END $f$`, []string{"brd-guard-src"}, foreignLogo},
			{"N4 auth reader seeing every stored object", `ALTER POLICY stored_objects_auth_reader ON app.stored_objects USING (true)`,
				[]string{"brd-policies"}, ""},
			{"N5 label check missing", `ALTER TABLE app.organization_branding DROP CONSTRAINT organization_branding_label_valid`,
				[]string{"brd-constraints"}, badLabel},
			{"N6 runtime DELETE", `GRANT DELETE ON app.organization_branding TO ` + quote(rt), []string{"brd-acl"}, ""},
			{"N7 projection executable by PUBLIC", `GRANT EXECUTE ON FUNCTION app.public_branding(uuid) TO PUBLIC`, []string{"brd-functions"}, ""},
			{"N8 row-level security not forced", `ALTER TABLE app.organization_branding NO FORCE ROW LEVEL SECURITY`, []string{"brd-flags"}, ""},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			v := brdViolations(t, tx, cfg, clean)
			_ = tx.Rollback(ctx)
			if !slices.Equal(v, n.want) {
				t.Errorf("%s: the detector reported %v, want %v", n.id, v, n.want)
			}
			if n.stmt != "" {
				if err := probe(n.id, n.sql, n.stmt); err != nil {
					t.Errorf("%s: the missing protection still refused the probe: %v", n.id, err)
				}
			}
		}
		if v := brdViolations(t, cat, cfg, clean); len(v) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", v)
		}
	})

	t.Run("CT-BRD-UPDOWN exact version-12 catalog when empty, up again, down refused with branding state or history", func(t *testing.T) {
		s13 := snapshot(t, r)
		down := func() (string, error) {
			var b bytes.Buffer
			err := r.Down(ctx, &b)
			return b.String(), err
		}
		exec := func(sql string) {
			t.Helper()
			if _, err := setup.Exec(ctx, sql); err != nil {
				t.Fatal(err)
			}
		}
		out, err := down()
		if err != nil || out != "rolled_back\t00013_organization_branding.sql\ncurrent_version\t12\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		metadataThroughVersion12Only(t, cfg, r)
		if got := snapshot(t, r); got != s12 {
			t.Fatalf("catalog after down differs from version 12:\n%s\n---\n%s", got, s12)
		}
		objRunUp(t, r, "applied\t00013_organization_branding.sql\ncurrent_version\t13\n")
		if got := snapshot(t, r); got != s13 {
			t.Fatalf("catalog after down/up differs:\n%s\n---\n%s", got, s13)
		}
		if v := brdViolations(t, cat, cfg, clean); len(v) != 0 {
			t.Fatalf("version-13 detector after up again: %v", v)
		}
		exec(`INSERT INTO app.users (id, display_name) VALUES ('0f17c000-0000-4000-8000-0000000000d1', 'Catalog Down')`)
		exec(`INSERT INTO app.organizations (id, name) VALUES ('0f17c000-0000-4000-8000-0000000000d2', 'Catalog Down')`)
		exec(`INSERT INTO app.organization_branding (organization_id, display_label, updated_by)
VALUES ('0f17c000-0000-4000-8000-0000000000d2', 'Catalog', '0f17c000-0000-4000-8000-0000000000d1')`)
		if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00013: branding rows exist") {
			t.Fatalf("down with a branding row: %v %q", err, out)
		}
		if got := snapshot(t, r); got != s13 {
			t.Fatal("the refused down changed the catalog")
		}
		// Branding history is never discarded: with the row gone, its organization.branding_changed row still refuses.
		exec(`DELETE FROM app.organization_branding WHERE organization_id = '0f17c000-0000-4000-8000-0000000000d2'`)
		if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00013: audit.events holds branding history") {
			t.Fatalf("down with branding history: %v %q", err, out)
		}
		reset(t, cfg)
	})
}
