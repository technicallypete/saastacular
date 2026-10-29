//go:build integration

// G071 TestDefaultTenantCatalog (Plan r4 §3, S2, §5.1 R1) against the disposable PostgreSQL of the migrations suite. It
// pins the exact version-21 -> 22 delta: one new function, app.organization_id_by_slug(text), owned by <db>_auth_reader,
// SECURITY DEFINER, STABLE, search_path pinned, the reviewed text, EXECUTE for the runtime and the owner only, 89
// relations and the 00021 auth-reader bodies unchanged (CT-DT-1); the body under the runtime role in a rolled-back
// transaction: an exact slug resolves to its UUID, an unknown slug, a differently cased slug and a platform label to
// NULL, nothing is written, and the other helpers are refused with 42501 (CT-DT-2); down 22 -> 21 restores version 21
// exactly and up restores version 22 (CT-DT-3). The expected body is read from the embedded migration itself. Setup uses
// the attested superuser; catalog reads use the owner (search_path pg_catalog).
package migrate

import (
	"bytes"
	"context"
	"testing"
)

// The fixture: one slugged and labeled Organization (dtOrg), selected by its slug dtSlug; its platform label dtLabel is
// never a selection key.
const (
	dtOrg   = "0f071000-0000-4000-8000-000000000001"
	dtSlug  = "alpha-agency"
	dtLabel = "dt-fixed-label"
)

func TestDefaultTenantCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	objRunUp(t, runnerFor(t, cfg, subsetFS(t, 21)), "current_version\t21\n")
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	rt := rlsRuntimeConn(t, cfg)
	s21 := namesCatalog(t, cat)
	r := runnerFor(t, cfg, subsetFS(t, 22))
	objRunUp(t, r, "applied\t00022_default_tenant.sql\ncurrent_version\t22\n")
	s22 := namesCatalog(t, cat)
	o, rtr, ar := cfg.OwnerRole(), cfg.RuntimeRole(), cfg.Database+"_auth_reader"
	pr, cw, pw := cfg.Database+"_policy_reader", cfg.Database+"_claim_writer", cfg.Database+"_provision_writer"
	one := func(sql string, args ...any) string {
		t.Helper()
		var s *string
		if err := cat.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		if s == nil {
			return ""
		}
		return *s
	}
	// lbl maps a role oid column onto labels ($1..$6: runtime, owner, auth reader, provision writer, claim writer, policy
	// reader); every query using it passes roles.
	lbl := func(col string) string {
		return `CASE ` + col + ` WHEN 0 THEN 'public' WHEN $1::regrole THEN 'rt' WHEN $2::regrole THEN 'ow' WHEN $3::regrole THEN 'ar'
  WHEN $4::regrole THEN 'pw' WHEN $5::regrole THEN 'cw' WHEN $6::regrole THEN 'pr' ELSE 'other' END`
	}
	roles := []any{rtr, o, ar, pw, cw, pr}

	t.Run("CT-DT-1 the exact version-22 delta: one auth-reader function, 89 relations", func(t *testing.T) {
		if got := namesDelta(s21, s22); len(got) != 1 || got[0] != "fn:app.organization_id_by_slug(text)" {
			t.Fatalf("version-21 -> 22 delta: %v, want exactly [fn:app.organization_id_by_slug(text)]", got)
		}
		metadataThroughVersion22(t, cfg, r)
		cfgPath := "search_path=pg_catalog, pg_temp"
		if got := one(`SELECT p.proname || ':' || `+lbl("p.proowner")+` || ':' || p.prosecdef::text || ':' || p.provolatile::text || ':' ||
  format_type(p.prorettype, NULL) || ':' || coalesce(array_to_string(p.proconfig, '|'), '-') || ':' ||
  coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C")
  FROM (SELECT `+lbl("a.grantee")+` || '=' || a.privilege_type AS v FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-')
FROM pg_proc p WHERE p.oid = 'app.organization_id_by_slug(text)'::regprocedure`, roles...); got !=
			"organization_id_by_slug:ar:true:s:uuid:"+cfgPath+":ow=EXECUTE+rt=EXECUTE" {
			t.Errorf("00022 DEFINER body: %s", got)
		}
		if got, want := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.organization_id_by_slug(text)'::regprocedure`),
			md5hex(migBody(t, "00022_default_tenant.sql", "CREATE FUNCTION app.organization_id_by_slug")); got != want {
			t.Errorf("organization_id_by_slug body digest %s, want %s (the reviewed file)", got, want)
		}
		if src := one(`SELECT prosrc FROM pg_proc WHERE oid = 'app.organization_id_by_slug(text)'::regprocedure`); src !=
			"\nBEGIN\n  RETURN (SELECT o.id FROM app.organizations o WHERE o.slug = (p_slug COLLATE \"C\"));\nEND " {
			t.Errorf("organization_id_by_slug body: %q", src)
		}
		// EXECUTE per role: the runtime and the owner by grant, the auth reader as the owner of the body; no other helper,
		// the retention role, the migrator or PUBLIC.
		for _, c := range []struct {
			role string
			want string
		}{
			{rtr, "true"}, {o, "true"}, {ar, "true"}, {pr, "false"}, {cw, "false"}, {pw, "false"},
			{cfg.RetentionRole(), "false"}, {cfg.MigratorRole(), "false"}, {"public", "false"},
		} {
			if got := one(`SELECT has_function_privilege($1::name, 'app.organization_id_by_slug(text)'::regprocedure, 'EXECUTE')::text`, c.role); got != c.want {
				t.Errorf("%s EXECUTE on organization_id_by_slug: %s, want %s", c.role, got, c.want)
			}
		}
		// The two 00021 auth-reader bodies keep their reviewed text (the delta above also shows their catalog lines unchanged).
		for _, c := range []struct{ fn, marker string }{
			{"app.hostname_state(text, text)", "CREATE OR REPLACE FUNCTION app.hostname_state"},
			{"app.signup_slug_taken(text)", "CREATE FUNCTION app.signup_slug_taken"},
		} {
			if got, want := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure($1)`, c.fn), md5hex(migBody(t, "00021_platform_labels.sql", c.marker)); got != want {
				t.Errorf("%s body digest %s, want %s (00021)", c.fn, got, want)
			}
		}
		if got := one(`SELECT string_agg(proname, ',' ORDER BY proname COLLATE "C") FROM pg_proc WHERE proowner = $1::regrole`, ar); got !=
			"auth_logout_ref,auth_principal_subject,auth_resolve_enrollment,auth_resolve_session,auth_resolve_signup,hostname_state,"+
				"is_organization_member,organization_id_by_slug,public_branding,signup_slug_taken" {
			t.Errorf("%s owns %s, want the version-21 nine plus organization_id_by_slug", ar, got)
		}
		if got := one(`SELECT has_schema_privilege($1::name, 'app', 'CREATE')::text`, ar); got != "false" {
			t.Errorf("auth reader CREATE on app: %s, want false", got)
		}
		if got := one(`SELECT count(*)::text FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
  AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
  AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass)`); got != "89" {
			t.Errorf("schema app holds %s relations, want 89", got)
		}
	})

	// Fixture (committed): one slugged and labeled Organization. 00022's down has no refusal, so it stays through CT-DT-3.
	if _, err := setup.Exec(ctx, `INSERT INTO app.organizations (id, name, slug, platform_label) VALUES ($1, 'Default Tenant Catalog', $2, $3)`,
		dtOrg, dtSlug, dtLabel); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	orgCount := func(t *testing.T) string {
		t.Helper()
		var s string
		if err := setup.QueryRow(ctx, `SELECT count(*)::text FROM app.organizations`).Scan(&s); err != nil {
			t.Fatalf("organization count: %v", err)
		}
		return s
	}

	t.Run("CT-DT-2 the body under the runtime role: exact slug only, read-only, helpers refused", func(t *testing.T) {
		before := orgCount(t)
		func() {
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			// No request context is set: the definer reads under the auth reader's USING (true) policy, not the runtime's.
			// The exact slug resolves to its UUID; an unknown slug, a differently cased slug (COLLATE "C": no case folding),
			// the platform label (never a selection key) and an untrimmed slug (core trims the environment value) are NULL.
			for _, c := range []struct{ slug, want string }{
				{dtSlug, dtOrg}, {"nope", ""}, {"Alpha-Agency", ""}, {dtLabel, ""}, {" " + dtSlug, ""},
			} {
				var got string
				if err := tx.QueryRow(ctx, `SELECT coalesce(app.organization_id_by_slug($1)::text, '')`, c.slug).Scan(&got); err != nil {
					t.Fatalf("organization_id_by_slug(%q) as the runtime: %v", c.slug, err)
				}
				if got != c.want {
					t.Errorf("organization_id_by_slug(%q) as the runtime: %q, want %q", c.slug, got, c.want)
				}
			}
		}()
		// The other helpers have no EXECUTE (each attempt rolled back; the attested superuser assumes the role).
		for _, role := range []string{pr, cw, pw, cfg.RetentionRole()} {
			func() {
				tx, err := setup.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+quote(role)); err != nil {
					t.Fatalf("SET LOCAL ROLE %s: %v", role, err)
				}
				var s *string
				err = tx.QueryRow(ctx, `SELECT app.organization_id_by_slug($1)::text`, dtSlug).Scan(&s)
				if code := sqlstateOf(err); code != "42501" {
					t.Errorf("%s calling organization_id_by_slug: %q (%v), want 42501", role, code, err)
				}
			}()
		}
		if after := orgCount(t); after != before {
			t.Errorf("app.organizations rows %s -> %s across the reads, want unchanged", before, after)
		}
	})

	t.Run("CT-DT-3 down 22 -> 21 restores version 21 exactly; up restores version 22", func(t *testing.T) {
		var b bytes.Buffer
		if err := r.Down(ctx, &b); err != nil || b.String() != "rolled_back\t00022_default_tenant.sql\ncurrent_version\t21\n" {
			t.Fatalf("down: %v %q", err, b.String())
		}
		if got := namesDelta(s21, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down differs from version 21: %v", got)
		}
		metadataThroughVersion21(t, cfg, r)
		var gone bool
		if err := cat.QueryRow(ctx, `SELECT to_regprocedure('app.organization_id_by_slug(text)') IS NULL`).Scan(&gone); err != nil || !gone {
			t.Fatalf("organization_id_by_slug after down: absent=%v (%v)", gone, err)
		}
		objRunUp(t, r, "applied\t00022_default_tenant.sql\ncurrent_version\t22\n")
		if got := namesDelta(s22, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down/up differs from version 22: %v", got)
		}
		// The reader still resolves the fixture after down/up (as the runtime, rolled back).
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		var got string
		if err := tx.QueryRow(ctx, `SELECT coalesce(app.organization_id_by_slug($1)::text, '')`, dtSlug).Scan(&got); err != nil || got != dtOrg {
			t.Errorf("organization_id_by_slug(%q) after down/up: %q (%v), want %s", dtSlug, got, err, dtOrg)
		}
	})
}
