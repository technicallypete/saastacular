//go:build integration

// G006 BusinessProfile catalog and migration tests against a disposable PostgreSQL initialized by the rendered role
// script (infra/validation/run.sh, suite "migrations"): B-BUSINESS-PROFILE-CATALOG (TestBusinessProfileCatalog) and
// B-BUSINESS-PROFILE-MIGRATION (TestBusinessProfileMigration, then TestBusinessProfilePersistSeed and
// TestBusinessProfilePersistVerify around a PostgreSQL restart). They fail, never skip, when the database settings are
// absent. Runtime cases run as the actual runtime role; owner-level cases use the migrator with SET ROLE <db>_owner.
// The service methods of G006 are not used here: rows are written through runtime-role SQL (Plan r2 MG-BP-3/6).
// Since G007 these predecessor tests run the normal set up to version 4 (runnerFor with subsetFS 4), so that every G006
// assertion keeps its exact version-4 meaning (G007 Plan r1 RC-5).
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, APP_ENV; the
// persist phases also BUSINESS_PROFILE_PERSIST_FILE.
package migrate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

// The relations of 00004_business_profiles.sql besides its five indexes, and its trigger functions.
var (
	businessProfileRels      = []string{"business_profile_access_paths", "business_profiles", "group_business_profiles"}
	businessProfileFunctions = []string{"business_profiles_keys_immutable", "group_business_profiles_keys_immutable"}
)

// accessPathsDef is the normalized pg_get_viewdef (search_path pg_catalog) of app.business_profile_access_paths.
const accessPathsDef = "SELECT rb.subject_user_id AS user_id, gbp.organization_id, gbp.business_profile_id, gbp.group_id, " +
	"rb.role AS group_role FROM (app.group_business_profiles gbp JOIN app.role_bindings rb ON " +
	"(((rb.organization_id = gbp.organization_id) AND (rb.group_id = gbp.group_id)))) " +
	"WHERE ((rb.scope_type = 'group'::text) AND (rb.source = 'membership'::text));"

// businessProfileAuthority is the version-4 inventory of the CT-6 detector: the two G005 membership projections,
// unchanged, plus the read-only access-path projection, which reads exactly group_business_profiles and role_bindings
// and takes its group_role column from role_bindings.role.
var businessProfileAuthority = authorityInventory{
	roleColumns: []string{"app.organization_memberships.role", "app.group_memberships.role", "app.business_profile_access_paths.group_role"},
	projections: map[string]projectionPin{
		"app.organization_memberships": pinnedProjections["app.organization_memberships"],
		"app.group_memberships":        pinnedProjections["app.group_memberships"],
		"app.business_profile_access_paths": {"user_id,organization_id,business_profile_id,group_id,group_role", accessPathsDef,
			[]string{"app.group_business_profiles", "app.role_bindings"}},
	},
}

// catalogPartQueries are tenancyCatalog's seven queries; catalogParts returns their rows unjoined, so that a multi-line
// view definition stays one row. joinParts(catalogParts) equals tenancyCatalog (asserted in CT-BP-1).
var catalogPartQueries = []string{
	`SELECT c.relkind::text || ':' || c.relname || '|' || c.relowner::regrole::text || '|' || coalesce(c.relacl::text, '')
FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace ORDER BY c.relname`,
	`SELECT c.relname || '.' || a.attname || '|' || a.attacl::text FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE c.relnamespace = 'app'::regnamespace AND a.attacl IS NOT NULL ORDER BY c.relname, a.attname`,
	`SELECT conrelid::regclass::text || '|' || conname || '|' || pg_get_constraintdef(oid) FROM pg_constraint
WHERE connamespace = 'app'::regnamespace ORDER BY conrelid::regclass::text, conname`,
	`SELECT pg_get_indexdef(i.indexrelid) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
WHERE c.relnamespace = 'app'::regnamespace ORDER BY c.relname`,
	`SELECT c.relname || '|' || coalesce(c.reloptions::text, '') || '|' || pg_get_viewdef(c.oid) FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relkind = 'v' ORDER BY c.relname`,
	`SELECT p.proname || '|' || p.prosecdef::text || '|' || coalesce(p.proconfig::text, '') || '|' || coalesce(p.proacl::text, '')
       || '|' || p.proowner::regrole::text || '|' || md5(p.prosrc) FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace ORDER BY p.proname`,
	`SELECT pg_get_triggerdef(g.oid) || '|' || g.tgenabled::text FROM pg_trigger g JOIN pg_class c ON c.oid = g.tgrelid
WHERE c.relnamespace = 'app'::regnamespace AND NOT g.tgisinternal ORDER BY g.tgname`,
}

func catalogParts(t *testing.T, q catalogQuerier) [][]string {
	t.Helper()
	out := make([][]string, len(catalogPartQueries))
	for i, sql := range catalogPartQueries {
		rows, err := q.Query(context.Background(), sql)
		if err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		vals, err := pgx.CollectRows(rows, pgx.RowTo[*string])
		if err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		for _, v := range vals {
			if v != nil {
				out[i] = append(out[i], *v)
			}
		}
	}
	return out
}

func joinParts(parts [][]string) string {
	s := make([]string, len(parts))
	for i, p := range parts {
		s[i] = strings.Join(p, "\n")
	}
	return strings.Join(s, "\n--\n")
}

// g006Row reports whether a catalog row describes a G006 object: every G006 relation, constraint, index, function and
// trigger name contains business_profile, and no predecessor name does (CT-1 of G005 forbids it at version 3).
func g006Row(row string) bool { return strings.Contains(row, "business_profile") }

// tenancyRowsSnapshot hashes every Organization, Group and role binding (owner connection).
func tenancyRowsSnapshot(t *testing.T, q catalogQuerier) string {
	t.Helper()
	return lines(t, q, `SELECT 'organizations:' || coalesce(md5(string_agg(o::text, E'\n' ORDER BY o.id)), '') FROM app.organizations o
UNION ALL SELECT 'groups:' || coalesce(md5(string_agg(g::text, E'\n' ORDER BY g.id)), '') FROM app.groups g
UNION ALL SELECT 'role_bindings:' || coalesce(md5(string_agg(b::text, E'\n' ORDER BY b.id)), '') FROM app.role_bindings b`)
}

func businessProfileObjectsAbsent(t *testing.T, q catalogQuerier) {
	t.Helper()
	got := lines(t, q, `SELECT n FROM unnest($1::text[]) n WHERE to_regclass('app.' || n) IS NOT NULL
UNION ALL SELECT p.proname FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace AND p.proname = ANY ($2::text[])`,
		businessProfileRels, businessProfileFunctions)
	if got != "" {
		t.Fatalf("G006 objects present after down: %s", got)
	}
}

// insertProfileAndLink writes one profile and its link for a Group through runtime-role SQL and returns their IDs.
func insertProfileAndLink(t *testing.T, rt *pgx.Conn, org, group, name string) (profile, link string) {
	t.Helper()
	ctx := context.Background()
	if err := rt.QueryRow(ctx, `INSERT INTO app.business_profiles (organization_id, display_name, website) VALUES ($1::text::uuid, $2, 'https://example.com/')
RETURNING id::text`, org, name).Scan(&profile); err != nil {
		t.Fatalf("runtime INSERT business_profiles: %v", err)
	}
	if err := rt.QueryRow(ctx, `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id)
VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid) RETURNING id::text`, org, group, profile).Scan(&link); err != nil {
		t.Fatalf("runtime INSERT group_business_profiles: %v", err)
	}
	return profile, link
}

// --- B-BUSINESS-PROFILE-CATALOG ---------------------------------------------------------------------------------------

func TestBusinessProfileCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r3 := runnerFor(t, cfg, subsetFS(t, 3))
	if out := run(t, "up to 3", func(ctx context.Context, b *bytes.Buffer) error { return r3.Up(ctx, b) }); out !=
		"applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\ncurrent_version\t3\n" {
		t.Fatalf("up to 3: %q", out)
	}
	_ = r3.Close()
	own := catalogConn(t, cfg)
	s3, parts3 := tenancyCatalog(t, own), catalogParts(t, own)
	r := runnerFor(t, cfg, subsetFS(t, 4))
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out !=
		"applied\t00004_business_profiles.sql\ncurrent_version\t4\n" {
		t.Fatalf("up to 4: %q", out)
	}
	rt := runtimeConn(t, cfg)
	o, rtr := cfg.OwnerRole(), cfg.RuntimeRole()

	t.Run("runtime identity", func(t *testing.T) {
		var cur, sess string
		var super, bypass bool
		var memberships int
		err := rt.QueryRow(ctx, `SELECT current_user::text, session_user::text, r.rolsuper, r.rolbypassrls,
  (SELECT count(*) FROM pg_auth_members m WHERE m.member = r.oid)::int
FROM pg_roles r WHERE r.rolname = current_user`).Scan(&cur, &sess, &super, &bypass, &memberships)
		if err != nil {
			t.Fatal(err)
		}
		if cur != rtr || sess != rtr || super || bypass || memberships != 0 {
			t.Fatalf("runtime connection: current=%s session=%s super=%v bypassrls=%v memberships=%d", cur, sess, super, bypass, memberships)
		}
	})

	t.Run("CT-BP-1 additive catalog", func(t *testing.T) {
		metadataIdentityTenancyAndBusinessProfileOnly(t, cfg, r)
		if joinParts(parts3) != s3 {
			t.Fatal("catalogParts does not reproduce tenancyCatalog at version 3")
		}
		parts4 := catalogParts(t, own)
		if joinParts(parts4) != tenancyCatalog(t, own) {
			t.Fatal("catalogParts does not reproduce tenancyCatalog at version 4")
		}
		// S4 minus the G006 rows equals S3: no predecessor row changed, including role_bindings_scope_consistent.
		names := []string{"relations", "column ACLs", "constraints", "indexes", "views", "functions", "triggers"}
		wantAdded := []int{8, 9, 15, 5, 1, 2, 2}
		for i := range parts4 {
			var kept []string
			added := 0
			for _, row := range parts4[i] {
				if g006Row(row) {
					added++
				} else {
					kept = append(kept, row)
				}
			}
			for _, row := range parts3[i] {
				if g006Row(row) {
					t.Fatalf("a version-3 %s row names a G006 object", names[i])
				}
			}
			if strings.Join(kept, "\n") != strings.Join(parts3[i], "\n") {
				t.Errorf("%s: a predecessor row changed:\n%s\n---\n%s", names[i], strings.Join(parts3[i], "\n"), strings.Join(kept, "\n"))
			}
			if added != wantAdded[i] {
				t.Errorf("%s: %d G006 rows, want %d", names[i], added, wantAdded[i])
			}
		}
	})

	t.Run("CT-BP-2 ownership and functions", func(t *testing.T) {
		got := lines(t, own, `SELECT c.relname || ':' || c.relowner::regrole::text FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($1::text[]) ORDER BY c.relname`, businessProfileRels)
		want := []string{"business_profile_access_paths:" + o, "business_profiles:" + o, "group_business_profiles:" + o}
		if got != strings.Join(want, "\n") {
			t.Fatalf("owners:\n%s", got)
		}
		got = lines(t, own, `SELECT p.proname || '|' || p.prosecdef::text || '|' || coalesce(p.proconfig::text, '') || '|' || coalesce(p.proacl::text, '')
  || '|' || p.proowner::regrole::text || '|' || p.prorettype::regtype::text || '|' || l.lanname
FROM pg_proc p JOIN pg_language l ON l.oid = p.prolang WHERE p.pronamespace = 'app'::regnamespace AND p.proname = ANY ($1::text[]) ORDER BY p.proname`,
			businessProfileFunctions)
		var w []string
		for _, f := range businessProfileFunctions {
			w = append(w, fmt.Sprintf("%s|false|{search_path=pg_catalog}|{%s=X/%s}|%s|trigger|plpgsql", f, o, o, o))
		}
		if got != strings.Join(w, "\n") {
			t.Fatalf("functions:\n got %s\nwant %s", got, strings.Join(w, "\n"))
		}
		for _, role := range []string{rtr, cfg.RetentionRole(), cfg.MigratorRole()} {
			if got := lines(t, own, `SELECT string_agg(p.proname, ',') FROM pg_proc p
WHERE p.pronamespace = 'app'::regnamespace AND p.proname = ANY ($2::text[]) AND has_function_privilege($1::text::name, p.oid, 'EXECUTE')`,
				role, businessProfileFunctions); got != "" {
				t.Errorf("%s may execute %s", role, got)
			}
		}
	})

	t.Run("CT-BP-3 table and column ACLs", func(t *testing.T) {
		got := lines(t, own, `SELECT c.relname || '|' || c.relacl::text FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($1::text[]) ORDER BY c.relname`, businessProfileRels)
		acl := func(rt string) string { return fmt.Sprintf("{%s=arwdDxtm/%s,%s=%s/%s}", o, o, rtr, rt, o) }
		want := strings.Join([]string{
			"business_profile_access_paths|" + acl("r"), "business_profiles|" + acl("r"), "group_business_profiles|" + acl("rd"),
		}, "\n")
		if got != want {
			t.Fatalf("table ACLs:\n got %s\nwant %s", got, want)
		}
		got = lines(t, own, `SELECT c.relname || '.' || a.attname || '|' || a.attacl::text FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($1::text[]) AND a.attacl IS NOT NULL ORDER BY c.relname, a.attname`, businessProfileRels)
		col := func(p string) string { return fmt.Sprintf("{%s=%s/%s}", rtr, p, o) }
		want = strings.Join([]string{
			"business_profiles.description|" + col("aw"), "business_profiles.display_name|" + col("aw"),
			"business_profiles.industry|" + col("aw"), "business_profiles.legal_name|" + col("aw"),
			"business_profiles.organization_id|" + col("a"), "business_profiles.website|" + col("aw"),
			"group_business_profiles.business_profile_id|" + col("a"), "group_business_profiles.group_id|" + col("a"),
			"group_business_profiles.organization_id|" + col("a"),
		}, "\n")
		if got != want {
			t.Fatalf("column ACLs:\n got %s\nwant %s", got, want)
		}
		// PUBLIC holds no entry on any G006 relation or column.
		if got := lines(t, own, `SELECT count(*)::text FROM pg_class c, aclexplode(c.relacl) a
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($1::text[]) AND a.grantee = 0`, businessProfileRels); got != "0" {
			t.Errorf("PUBLIC table entries: %s", got)
		}
		if got := lines(t, own, `SELECT count(*)::text FROM pg_attribute at JOIN pg_class c ON c.oid = at.attrelid, aclexplode(at.attacl) a
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($1::text[]) AND a.grantee = 0`, businessProfileRels); got != "0" {
			t.Errorf("PUBLIC column entries: %s", got)
		}
		// The effective privilege matrix: runtime exactly as granted; retention and migrator (without SET ROLE) nothing.
		matrix := `SELECT string_agg(c.relname || ':' || p, ',' ORDER BY c.relname, p) FROM pg_class c,
  unnest(ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'TRUNCATE', 'REFERENCES', 'TRIGGER', 'MAINTAIN']) p
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($2::text[]) AND has_table_privilege($1::text::name, c.oid, p)`
		cols := `SELECT string_agg(c.relname || '.' || a.attname || ':' || p, ',' ORDER BY c.relname, a.attname, p)
FROM pg_class c JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped,
  unnest(ARRAY['INSERT', 'UPDATE', 'REFERENCES']) p
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($2::text[]) AND has_column_privilege($1::text::name, c.oid, a.attnum, p)`
		if got := lines(t, own, matrix, rtr, businessProfileRels); got != "business_profile_access_paths:SELECT,business_profiles:SELECT,"+
			"group_business_profiles:DELETE,group_business_profiles:SELECT" {
			t.Errorf("runtime table privileges: %s", got)
		}
		if got := lines(t, own, cols, rtr, businessProfileRels); got != "business_profiles.description:INSERT,business_profiles.description:UPDATE,"+
			"business_profiles.display_name:INSERT,business_profiles.display_name:UPDATE,business_profiles.industry:INSERT,"+
			"business_profiles.industry:UPDATE,business_profiles.legal_name:INSERT,business_profiles.legal_name:UPDATE,"+
			"business_profiles.organization_id:INSERT,business_profiles.website:INSERT,business_profiles.website:UPDATE,"+
			"group_business_profiles.business_profile_id:INSERT,group_business_profiles.group_id:INSERT,"+
			"group_business_profiles.organization_id:INSERT" {
			t.Errorf("runtime column privileges: %s", got)
		}
		for _, role := range []string{cfg.RetentionRole(), cfg.MigratorRole()} {
			if got := lines(t, own, matrix, role, businessProfileRels) + lines(t, own, cols, role, businessProfileRels); got != "" {
				t.Errorf("%s holds %s", role, got)
			}
		}
	})

	t.Run("CT-BP-4 constraints", func(t *testing.T) {
		got := lines(t, own, `SELECT conrelid::regclass::text || ' ' || conname || ' ' || contype::text || ' ' || pg_get_constraintdef(oid)
FROM pg_constraint WHERE conrelid = ANY (ARRAY['app.business_profiles'::regclass, 'app.group_business_profiles'::regclass])
ORDER BY conrelid::regclass::text, conname`)
		want := []string{
			`app.business_profiles business_profiles_description_valid c CHECK (((description IS NULL) OR (((char_length(description) >= 1) AND (char_length(description) <= 4000)) AND (description = btrim(description)) AND ((description COLLATE "C") !~ '[\u0001-\u0008\u000b-\u001f\u007f-\u009f]'::text))))`,
			`app.business_profiles business_profiles_display_name_valid c CHECK ((((char_length(display_name) >= 1) AND (char_length(display_name) <= 255)) AND (display_name = btrim(display_name)) AND ((display_name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'::text)))`,
			`app.business_profiles business_profiles_id_nonzero c CHECK ((id <> '00000000-0000-0000-0000-000000000000'::uuid))`,
			`app.business_profiles business_profiles_industry_valid c CHECK (((industry IS NULL) OR (((char_length(industry) >= 1) AND (char_length(industry) <= 255)) AND (industry = btrim(industry)) AND ((industry COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'::text))))`,
			`app.business_profiles business_profiles_legal_name_valid c CHECK (((legal_name IS NULL) OR (((char_length(legal_name) >= 1) AND (char_length(legal_name) <= 255)) AND (legal_name = btrim(legal_name)) AND ((legal_name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'::text))))`,
			`app.business_profiles business_profiles_organization_fkey f FOREIGN KEY (organization_id) REFERENCES app.organizations(id) ON DELETE RESTRICT`,
			`app.business_profiles business_profiles_organization_id_key u UNIQUE (organization_id, id)`,
			`app.business_profiles business_profiles_pkey p PRIMARY KEY (id)`,
			`app.business_profiles business_profiles_website_valid c CHECK (((website IS NULL) OR (((char_length(website) >= 1) AND (char_length(website) <= 2048)) AND ((website COLLATE "C") ~ '^https?://[\u0021-\u005b\u005d-\u007e]+$'::text))))`,
			`app.group_business_profiles group_business_profiles_business_profile_fkey f FOREIGN KEY (organization_id, business_profile_id) REFERENCES app.business_profiles(organization_id, id) ON DELETE RESTRICT`,
			`app.group_business_profiles group_business_profiles_group_fkey f FOREIGN KEY (organization_id, group_id) REFERENCES app.groups(organization_id, id) ON DELETE RESTRICT`,
			`app.group_business_profiles group_business_profiles_group_key u UNIQUE (organization_id, group_id)`,
			`app.group_business_profiles group_business_profiles_id_nonzero c CHECK ((id <> '00000000-0000-0000-0000-000000000000'::uuid))`,
			`app.group_business_profiles group_business_profiles_organization_fkey f FOREIGN KEY (organization_id) REFERENCES app.organizations(id) ON DELETE RESTRICT`,
			`app.group_business_profiles group_business_profiles_pkey p PRIMARY KEY (id)`,
		}
		var gotN []string
		for _, l := range strings.Split(got, "\n") {
			gotN = append(gotN, normalize(l))
		}
		if strings.Join(gotN, "\n") != strings.Join(want, "\n") {
			t.Fatalf("constraints:\n got %s\nwant %s", strings.Join(gotN, "\n"), strings.Join(want, "\n"))
		}
		if got := lines(t, own, `SELECT string_agg(conname || ':' || confdeltype::text || confupdtype::text || confmatchtype::text, ',' ORDER BY conname)
FROM pg_constraint WHERE contype = 'f' AND conrelid = ANY (ARRAY['app.business_profiles'::regclass, 'app.group_business_profiles'::regclass])`); got !=
			"business_profiles_organization_fkey:ras,group_business_profiles_business_profile_fkey:ras,group_business_profiles_group_fkey:ras,"+
				"group_business_profiles_organization_fkey:ras" {
			t.Fatalf("foreign keys (delete restrict, update no action, match simple): %s", got)
		}
		// Uniqueness: a Group has at most one association; nothing is unique on a profile ID or an Organization ID alone.
		unique := `SELECT string_agg(c.relname || ' (' || (SELECT string_agg(a.attname, ', ' ORDER BY k.n) FROM unnest(i.indkey) WITH ORDINALITY k(attnum, n)
  JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = k.attnum) || ')', '; ' ORDER BY c.relname)
FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE i.indrelid = $1::text::regclass AND i.indisunique`
		if got := lines(t, own, unique, "app.group_business_profiles"); got != "group_business_profiles_group_key (organization_id, group_id); group_business_profiles_pkey (id)" {
			t.Errorf("association unique indexes: %s", got)
		}
		if got := lines(t, own, unique, "app.business_profiles"); got != "business_profiles_organization_id_key (organization_id, id); business_profiles_pkey (id)" {
			t.Errorf("profile unique indexes: %s", got)
		}
	})

	t.Run("CT-BP-5 indexes view and triggers", func(t *testing.T) {
		got := lines(t, own, `SELECT pg_get_indexdef(i.indexrelid) FROM pg_index i
WHERE i.indrelid = ANY (ARRAY['app.business_profiles'::regclass, 'app.group_business_profiles'::regclass]) ORDER BY i.indexrelid::regclass::text`)
		want := strings.Join([]string{
			"CREATE UNIQUE INDEX business_profiles_organization_id_key ON app.business_profiles USING btree (organization_id, id)",
			"CREATE UNIQUE INDEX business_profiles_pkey ON app.business_profiles USING btree (id)",
			"CREATE INDEX group_business_profiles_business_profile_idx ON app.group_business_profiles USING btree (organization_id, business_profile_id)",
			"CREATE UNIQUE INDEX group_business_profiles_group_key ON app.group_business_profiles USING btree (organization_id, group_id)",
			"CREATE UNIQUE INDEX group_business_profiles_pkey ON app.group_business_profiles USING btree (id)",
		}, "\n")
		if got != want {
			t.Fatalf("indexes:\n got %s\nwant %s", got, want)
		}
		var opts, def, cols string
		if err := own.QueryRow(ctx, `SELECT coalesce(reloptions::text, ''), pg_get_viewdef(oid),
  (SELECT string_agg(a.attname, ',' ORDER BY a.attnum) FROM pg_attribute a WHERE a.attrelid = c.oid AND a.attnum > 0)
FROM pg_class c WHERE oid = 'app.business_profile_access_paths'::regclass`).Scan(&opts, &def, &cols); err != nil {
			t.Fatal(err)
		}
		if opts != "{security_invoker=true}" || normalize(def) != accessPathsDef || cols != "user_id,organization_id,business_profile_id,group_id,group_role" {
			t.Errorf("access-path view: options %s columns %s definition %q", opts, cols, normalize(def))
		}
		got = lines(t, own, `SELECT pg_get_triggerdef(g.oid) || ' ' || g.tgenabled::text FROM pg_trigger g
WHERE g.tgrelid = ANY (ARRAY['app.business_profiles'::regclass, 'app.group_business_profiles'::regclass]) AND NOT g.tgisinternal ORDER BY g.tgname`)
		want = strings.Join([]string{
			"CREATE TRIGGER business_profiles_keys_immutable BEFORE UPDATE ON app.business_profiles FOR EACH ROW EXECUTE FUNCTION app.business_profiles_keys_immutable() O",
			"CREATE TRIGGER group_business_profiles_keys_immutable BEFORE UPDATE ON app.group_business_profiles FOR EACH ROW EXECUTE FUNCTION app.group_business_profiles_keys_immutable() O",
		}, "\n")
		if got != want {
			t.Fatalf("triggers:\n got %s\nwant %s", got, want)
		}
	})

	t.Run("CT-BP-6 single authority at version 4", func(t *testing.T) {
		if v := authorityViolations(t, own, cfg, businessProfileAuthority); len(v) != 0 {
			t.Fatalf("authority violations at version 4: %v", v)
		}
		// The version-3 inventory does not silently accept the new projection.
		if v := authorityViolations(t, own, cfg, tenancyAuthority); strings.Join(v, ",") != "unexpected_role_projection" {
			t.Fatalf("version-3 inventory at version 4: %v, want exactly [unexpected_role_projection]", v)
		}
	})

	t.Run("CT-BP-6-NEG detector controls at version 4", func(t *testing.T) {
		before := tenancyCatalog(t, own)
		rtID := quote(rtr)
		for _, c := range []struct {
			name, ddl string
			want      []string
		}{
			// G005 N1-N9b, unchanged, with the version-4 inventory.
			{"N1", `ALTER TABLE app.groups ADD COLUMN role text`, []string{"stored_role_outside_authority"}},
			{"N2", `CREATE TABLE app.membership_mirror (user_id uuid, organization_id uuid, role text)`, []string{"stored_role_outside_authority"}},
			{"N3", `CREATE MATERIALIZED VIEW app.membership_cache AS SELECT subject_user_id, role FROM app.role_bindings`, []string{"stored_role_outside_authority"}},
			{"N4", `CREATE VIEW app.extra_memberships AS SELECT subject_user_id, role FROM app.role_bindings`, []string{"unexpected_role_projection"}},
			{"N5", `CREATE OR REPLACE VIEW app.organization_memberships WITH (security_invoker = true) AS
  SELECT subject_user_id AS user_id, organization_id, 'org_owner'::text AS role
  FROM app.role_bindings WHERE scope_type = 'organization' AND source = 'membership'`, []string{"projection_definition", "projection_source"}},
			{"N6", `GRANT INSERT ON app.group_memberships TO ` + rtID, []string{"writable_projection"}},
			{"N7a", `CREATE FUNCTION app.n7_projection_insert() RETURNS trigger LANGUAGE plpgsql AS $f$ BEGIN RETURN NULL; END $f$;
CREATE TRIGGER n7_projection_insert INSTEAD OF INSERT ON app.organization_memberships FOR EACH ROW EXECUTE FUNCTION app.n7_projection_insert()`, []string{"projection_write_path"}},
			{"N7b", `CREATE RULE n7_projection_rule AS ON INSERT TO app.organization_memberships DO INSTEAD NOTHING`, []string{"projection_write_path"}},
			{"N8", `ALTER TABLE app.groups ADD COLUMN role text; GRANT UPDATE (role) ON app.groups TO ` + rtID, []string{"stored_role_outside_authority", "writable_role_column"}},
			{"N9a", `ALTER TABLE app.groups ADD COLUMN owner_user_id uuid`, []string{"stored_role_outside_authority"}},
			{"N9b", `ALTER TABLE app.groups ADD COLUMN created_by uuid`, []string{"stored_role_outside_authority"}},
			// G006: no profile role, owner, mirror, writable projection, altered projection or cached authority.
			{"N10", `ALTER TABLE app.business_profiles ADD COLUMN role text`, []string{"stored_role_outside_authority"}},
			// N11 also reports writable_role_column: the runtime role holds table-level DELETE on associations (as N8).
			{"N11", `ALTER TABLE app.group_business_profiles ADD COLUMN role text`, []string{"stored_role_outside_authority", "writable_role_column"}},
			{"N12", `ALTER TABLE app.business_profiles ADD COLUMN owner_user_id uuid`, []string{"stored_role_outside_authority"}},
			{"N13", `CREATE TABLE app.business_profile_memberships (user_id uuid, business_profile_id uuid, role text)`, []string{"stored_role_outside_authority"}},
			{"N14", `GRANT INSERT ON app.business_profile_access_paths TO ` + rtID, []string{"writable_projection"}},
			{"N15", `CREATE OR REPLACE VIEW app.business_profile_access_paths WITH (security_invoker = true) AS
  SELECT rb.subject_user_id AS user_id, gbp.organization_id, gbp.business_profile_id, gbp.group_id, 'group_owner'::text AS group_role
  FROM app.group_business_profiles gbp
  JOIN app.role_bindings rb ON rb.organization_id = gbp.organization_id AND rb.group_id = gbp.group_id
  WHERE rb.scope_type = 'group' AND rb.source = 'membership'`, []string{"projection_definition", "projection_source"}},
			{"N16", `CREATE MATERIALIZED VIEW app.business_profile_access_cache AS SELECT * FROM app.business_profile_access_paths`, []string{"stored_role_outside_authority"}},
		} {
			t.Run(c.name, func(t *testing.T) {
				tx, err := own.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				if _, err := tx.Exec(ctx, c.ddl); err != nil {
					t.Fatalf("defect: %v", err)
				}
				if got := authorityViolations(t, tx, cfg, businessProfileAuthority); strings.Join(got, ",") != strings.Join(c.want, ",") {
					t.Fatalf("violations %v, want exactly %v", got, c.want)
				}
			})
		}
		if v := authorityViolations(t, own, cfg, businessProfileAuthority); len(v) != 0 {
			t.Fatalf("violations after the rolled-back controls: %v", v)
		}
		if after := tenancyCatalog(t, own); after != before {
			t.Fatalf("catalog changed by the rolled-back controls:\n%s\n---\n%s", before, after)
		}
	})

	t.Run("CT-BP-7 no row-level security yet", func(t *testing.T) {
		// G008's obligation, recorded in docs/foundation/TENANCY.md; this is not isolation evidence.
		if got := lines(t, own, `SELECT string_agg(relname || ':' || relrowsecurity::text || relforcerowsecurity::text, ',' ORDER BY relname)
FROM pg_class WHERE oid = ANY (ARRAY['app.business_profiles'::regclass, 'app.group_business_profiles'::regclass])`); got !=
			"business_profiles:falsefalse,group_business_profiles:falsefalse" {
			t.Fatalf("row security: %s", got)
		}
		if got := lines(t, own, `SELECT count(*)::text FROM pg_policy`); got != "0" {
			t.Fatalf("policies: %s", got)
		}
	})

	t.Run("CT-BP-8 no premature surface", func(t *testing.T) {
		wantCols := map[string]string{
			"business_profiles": "id uuid not null default gen_random_uuid(), organization_id uuid not null, display_name text not null, " +
				"legal_name text, description text, industry text, website text, created_at timestamp with time zone not null default now()",
			"group_business_profiles": "id uuid not null default gen_random_uuid(), organization_id uuid not null, group_id uuid not null, " +
				"business_profile_id uuid not null, created_at timestamp with time zone not null default now()",
		}
		for tbl, want := range wantCols {
			got := lines(t, own, `SELECT string_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod)
  || CASE WHEN a.attnotnull THEN ' not null' ELSE '' END
  || coalesce(' default ' || pg_get_expr(d.adbin, d.adrelid), ''), ', ' ORDER BY a.attnum)
FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = ('app.' || $1::text)::regclass AND a.attnum > 0 AND NOT a.attisdropped`, tbl)
			if got != want {
				t.Errorf("%s columns:\n got %s\nwant %s", tbl, got, want)
			}
		}
		// No billing, payment, tax, invoice, onboarding, completion, policy, capability, permission or tenant-context
		// relation or column anywhere in app or public.
		forbidden := `(billing|payment|tax|invoice|onboard|complet|polic|capabilit|permission|tenant_context)`
		if got := lines(t, own, `SELECT c.relname FROM pg_class c WHERE c.relnamespace IN ('app'::regnamespace, 'public'::regnamespace) AND c.relname ~ $1
UNION ALL SELECT c.relname || '.' || a.attname FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE c.relnamespace IN ('app'::regnamespace, 'public'::regnamespace) AND a.attnum > 0 AND NOT a.attisdropped AND a.attname ~ $1`, forbidden); got != "" {
			t.Fatalf("forbidden objects: %s", got)
		}
		// role_bindings keeps exactly its two G005 scope branches: no BusinessProfile scope.
		def := lines(t, own, `SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'app.role_bindings'::regclass AND conname = 'role_bindings_scope_consistent'`)
		if strings.Count(def, "scope_type = ") != 2 || strings.Contains(def, "business") {
			t.Fatalf("role_bindings_scope_consistent: %s", def)
		}
		// No business_profile_id on any predecessor table.
		if got := lines(t, own, `SELECT c.relname FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE c.relnamespace = 'app'::regnamespace AND c.relkind = 'r' AND a.attname = 'business_profile_id' AND c.relname <> 'group_business_profiles'`); got != "" {
			t.Fatalf("business_profile_id on %s", got)
		}
	})

	t.Run("CT-BP-9 runtime denials", func(t *testing.T) {
		for what, stmt := range map[string]string{
			"CREATE TABLE":                `CREATE TABLE app.runtime_business_profile (id int)`,
			"ALTER business_profiles":     `ALTER TABLE app.business_profiles ADD COLUMN x int`,
			"ALTER associations":          `ALTER TABLE app.group_business_profiles DROP CONSTRAINT group_business_profiles_group_key`,
			"DROP business_profiles":      `DROP TABLE app.business_profiles`,
			"DROP associations":           `DROP TABLE app.group_business_profiles`,
			"DROP view":                   `DROP VIEW app.business_profile_access_paths`,
			"TRUNCATE business_profiles":  `TRUNCATE app.business_profiles`,
			"TRUNCATE associations":       `TRUNCATE app.group_business_profiles`,
			"DELETE business_profiles":    `DELETE FROM app.business_profiles`,
			"UPDATE association group":    `UPDATE app.group_business_profiles SET group_id = group_id`,
			"UPDATE association profile":  `UPDATE app.group_business_profiles SET business_profile_id = business_profile_id`,
			"UPDATE profile Organization": `UPDATE app.business_profiles SET organization_id = organization_id`,
			"UPDATE profile id":           `UPDATE app.business_profiles SET id = id`,
			"CREATE TRIGGER":              `CREATE TRIGGER x BEFORE UPDATE ON app.business_profiles FOR EACH ROW EXECUTE FUNCTION app.business_profiles_keys_immutable()`,
			"DISABLE TRIGGER":             `ALTER TABLE app.business_profiles DISABLE TRIGGER business_profiles_keys_immutable`,
			"SET ROLE owner":              `SET ROLE ` + quote(cfg.OwnerRole()),
			"SET ROLE migrator":           `SET ROLE ` + quote(cfg.MigratorRole()),
			"metadata read":               `SELECT 1 FROM app.goose_db_version`,
		} {
			expectDenied(t, rt, "runtime "+what, stmt)
		}
	})
}

// --- B-BUSINESS-PROFILE-MIGRATION -----------------------------------------------------------------------------------

func TestBusinessProfileMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	const full = "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\ncurrent_version\t4\n"

	t.Run("MG-BP-1 empty to full", func(t *testing.T) {
		reset(t, cfg)
		r := runnerFor(t, cfg, subsetFS(t, 4))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != full {
			t.Fatalf("up: %q", out)
		}
		metadataIdentityTenancyAndBusinessProfileOnly(t, cfg, r)
	})

	t.Run("MG-BP-2 repeated up and status", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 4))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "no pending migrations\ncurrent_version\t4\n" {
			t.Fatalf("repeated up: %q", out)
		}
		out := run(t, "status", func(ctx context.Context, b *bytes.Buffer) error { return r.Status(ctx, b) })
		for _, m := range []string{"00001_baseline.sql", "00002_identity.sql", "00003_tenancy.sql", "00004_business_profiles.sql"} {
			if !regexp.MustCompile(`(?m)^migration\t` + regexp.QuoteMeta(m) + `\tapplied\t\S+$`).MatchString(out) {
				t.Errorf("status lacks applied %s: %q", m, out)
			}
		}
		if !strings.HasSuffix(out, "current_version\t4\n") || strings.Contains(out, "pending") {
			t.Fatalf("status: %q", out)
		}
	})

	t.Run("MG-BP-4 reapply keeps the catalog", func(t *testing.T) {
		own := catalogConn(t, cfg)
		before := tenancyCatalog(t, own)
		r := runnerFor(t, cfg, subsetFS(t, 4))
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out !=
			"rolled_back\t00004_business_profiles.sql\napplied\t00004_business_profiles.sql\ncurrent_version\t4\n" {
			t.Fatalf("reapply: %q", out)
		}
		if after := tenancyCatalog(t, own); after != before {
			t.Fatalf("catalog after reapply differs:\n%s\n---\n%s", before, after)
		}
	})

	var preG006, idSnap, tenSnap string
	t.Run("MG-BP-3 populated upgrade", func(t *testing.T) {
		reset(t, cfg)
		r3 := runnerFor(t, cfg, subsetFS(t, 3))
		if out := run(t, "up to 3", func(ctx context.Context, b *bytes.Buffer) error { return r3.Up(ctx, b) }); out !=
			"applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\ncurrent_version\t3\n" {
			t.Fatalf("up to 3: %q", out)
		}
		_ = r3.Close()
		rt := runtimeConn(t, cfg)
		users := seedIdentity(t, rt, 3, "bp")
		svc := tenancy.NewService(runtimePool(t, cfg), nil)
		org, err := svc.CreateOrganization(ctx, "Upgrade Organization")
		if err != nil {
			t.Fatal(err)
		}
		creator, _ := tenancy.ParseUserID(users[0])
		admin, _ := tenancy.ParseUserID(users[1])
		member, _ := tenancy.ParseUserID(users[2])
		g, _, err := svc.CreateGroup(ctx, org.ID, creator, "Upgrade Group", "")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GrantOrganizationRole(ctx, org.ID, admin, tenancy.OrgAdmin); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GrantGroupRole(ctx, org.ID, g.ID, member, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		own := catalogConn(t, cfg)
		idSnap, tenSnap, preG006 = identitySnapshot(t, own), tenancyRowsSnapshot(t, own), tenancyCatalog(t, own)
		r := runnerFor(t, cfg, subsetFS(t, 4))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00004_business_profiles.sql\ncurrent_version\t4\n" {
			t.Fatalf("upgrade: %q", out)
		}
		if identitySnapshot(t, own) != idSnap || tenancyRowsSnapshot(t, own) != tenSnap {
			t.Fatal("identity or tenancy rows changed by the upgrade")
		}
		// Runtime smoke through runtime-role SQL: a profile, its link to the Group and the derived access paths.
		profile, _ := insertProfileAndLink(t, rt, org.ID.String(), g.ID.String(), "Upgrade Profile")
		got := lines(t, rt, `SELECT user_id::text || ' ' || business_profile_id::text || ' ' || group_id::text || ' ' || group_role
FROM app.business_profile_access_paths WHERE organization_id = $1::text::uuid ORDER BY group_role, user_id`, org.ID.String())
		want := strings.Join([]string{
			member.String() + " " + profile + " " + g.ID.String() + " group_member",
			creator.String() + " " + profile + " " + g.ID.String() + " group_owner",
		}, "\n")
		if got != want {
			t.Fatalf("access paths after the upgrade smoke:\n got %s\nwant %s", got, want)
		}
		if got := lines(t, own, `SELECT (SELECT count(*) FROM app.business_profiles)::text || '/' || (SELECT count(*) FROM app.group_business_profiles)::text`); got != "1/1" {
			t.Fatalf("G006 rows after the smoke: %s", got)
		}
		if identitySnapshot(t, own) != idSnap || tenancyRowsSnapshot(t, own) != tenSnap {
			t.Fatal("identity or tenancy rows changed by G006 writes")
		}
	})

	t.Run("MG-BP-5 down to 3 with G006 rows present", func(t *testing.T) {
		// Down destroys the G006 rows with their tables (documented); identity rows, every Organization, Group, binding
		// and slug, and the catalog of version 3 are unchanged.
		r := runnerFor(t, cfg, subsetFS(t, 4))
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out != "rolled_back\t00004_business_profiles.sql\ncurrent_version\t3\n" {
			t.Fatalf("down: %q", out)
		}
		own := catalogConn(t, cfg)
		businessProfileObjectsAbsent(t, own)
		if identitySnapshot(t, own) != idSnap || tenancyRowsSnapshot(t, own) != tenSnap {
			t.Fatal("identity or tenancy rows changed by down")
		}
		if got := tenancyCatalog(t, own); got != preG006 {
			t.Fatalf("catalog after down differs from version 3:\n%s\n---\n%s", preG006, got)
		}
		metadataIdentityAndTenancyOnly(t, cfg, r)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00004_business_profiles.sql\ncurrent_version\t4\n" {
			t.Fatalf("reapply after down: %q", out)
		}
		if identitySnapshot(t, own) != idSnap || tenancyRowsSnapshot(t, own) != tenSnap {
			t.Fatal("identity or tenancy rows changed by up after down")
		}
		if got := lines(t, own, `SELECT (SELECT count(*) FROM app.business_profiles)::text || '/' || (SELECT count(*) FROM app.group_business_profiles)::text`); got != "0/0" {
			t.Fatalf("G006 rows after down and up: %s", got)
		}
	})

	t.Run("MG-BP-7 identity fixture truncate at version 4", func(t *testing.T) {
		rt := runtimeConn(t, cfg)
		own := catalogConn(t, cfg)
		users := seedIdentity(t, rt, 2, "bpesc")
		svc := tenancy.NewService(runtimePool(t, cfg), nil)
		org, err := svc.CreateOrganization(ctx, "Fixture Organization")
		if err != nil {
			t.Fatal(err)
		}
		u0, _ := tenancy.ParseUserID(users[0])
		u1, _ := tenancy.ParseUserID(users[1])
		g, _, err := svc.CreateGroup(ctx, org.ID, u0, "Fixture Group", "")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GrantGroupRole(ctx, org.ID, g.ID, u1, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		insertProfileAndLink(t, rt, org.ID.String(), g.ID.String(), "Fixture Profile")
		counts := `SELECT (SELECT count(*) FROM app.organizations)::text || '/' || (SELECT count(*) FROM app.groups)::text || '/' ||
  (SELECT count(*) FROM app.business_profiles)::text || '/' || (SELECT count(*) FROM app.group_business_profiles)::text || '/' ||
  (SELECT count(*) FROM app.role_bindings)::text || '/' || (SELECT count(*) FROM app.users)::text || '/' ||
  (SELECT count(*) FROM app.sessions)::text || '/' || (SELECT count(*) FROM app.identity_principals)::text`
		before := strings.Split(lines(t, own, counts), "/")
		profiles := lines(t, own, `SELECT coalesce((SELECT md5(string_agg(p::text, E'\n' ORDER BY p.id)) FROM app.business_profiles p), '') || '|' ||
  coalesce((SELECT md5(string_agg(l::text, E'\n' ORDER BY l.id)) FROM app.group_business_profiles l), '')`)
		tx, err := own.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, strings.TrimSuffix(identityFixtureTruncate, " CASCADE"))
		_ = tx.Rollback(ctx)
		if sqlstateOf(err) != "0A000" {
			t.Fatalf("plain TRUNCATE at version 4: %v, want 0A000", err)
		}
		if _, err := own.Exec(ctx, identityFixtureTruncate); err != nil {
			t.Fatalf("fixture TRUNCATE ... CASCADE: %v", err)
		}
		// Bindings are emptied with their users; Organizations, Groups, profiles and links remain unchanged.
		want := strings.Join(append(slices.Clone(before[:4]), "0", "0", "0", "0"), "/")
		if got := lines(t, own, counts); got != want {
			t.Fatalf("after the fixture reset: %s, want %s", got, want)
		}
		if got := lines(t, own, `SELECT coalesce((SELECT md5(string_agg(p::text, E'\n' ORDER BY p.id)) FROM app.business_profiles p), '') || '|' ||
  coalesce((SELECT md5(string_agg(l::text, E'\n' ORDER BY l.id)) FROM app.group_business_profiles l), '')`); got != profiles {
			t.Fatal("profiles or links changed by the fixture reset")
		}
		if got := lines(t, own, `SELECT count(*)::text FROM app.business_profile_access_paths`); got != "0" {
			t.Fatalf("access paths after the fixture reset: %s", got)
		}
		metadataIdentityTenancyAndBusinessProfileOnly(t, cfg, runnerFor(t, cfg, subsetFS(t, 4)))
	})
}

// TestBusinessProfilePersistSeed (MG-BP-6, before the harness restarts PostgreSQL with its volume kept) seeds G006 rows
// at version 4 and records their IDs in BUSINESS_PROFILE_PERSIST_FILE (0600). TestBusinessProfilePersistVerify reads
// them back afterwards.
func TestBusinessProfilePersistSeed(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	file := os.Getenv("BUSINESS_PROFILE_PERSIST_FILE")
	if file == "" {
		t.Fatal("BUSINESS_PROFILE_PERSIST_FILE is required")
	}
	// G064: the seed runs on its own fresh copy of the current template; it establishes version 4 itself (reset to
	// empty, then the bounded Up) instead of relying on TestBusinessProfileMigration's end state.
	reset(t, cfg)
	r := runnerFor(t, cfg, subsetFS(t, 4))
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t4\n") {
		t.Fatalf("up: %q", out)
	}
	rt := runtimeConn(t, cfg)
	users := seedIdentity(t, rt, 1, "bppersist")
	svc := tenancy.NewService(runtimePool(t, cfg), nil)
	org, err := svc.CreateOrganization(ctx, "Persisted Profile Organization")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := tenancy.ParseUserID(users[0])
	g, _, err := svc.CreateGroup(ctx, org.ID, u, "Persisted Profile Group", "persisted-profile-group")
	if err != nil {
		t.Fatal(err)
	}
	profile, link := insertProfileAndLink(t, rt, org.ID.String(), g.ID.String(), "Persisted Profile")
	content := strings.Join([]string{org.ID.String(), g.ID.String(), u.String(), profile, link}, "\n") + "\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestBusinessProfilePersistVerify(t *testing.T) {
	cfg := itConfig(t)
	file := os.Getenv("BUSINESS_PROFILE_PERSIST_FILE")
	if file == "" {
		t.Fatal("BUSINESS_PROFILE_PERSIST_FILE is required")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	f := strings.Fields(string(b))
	if len(f) != 5 {
		t.Fatalf("persist file has %d fields", len(f))
	}
	_, err1 := tenancy.ParseOrganizationID(f[0])
	_, err2 := tenancy.ParseGroupID(f[1])
	_, err3 := tenancy.ParseUserID(f[2])
	if err := errors.Join(err1, err2, err3); err != nil {
		t.Fatal(err)
	}
	r := runnerFor(t, cfg, subsetFS(t, 4))
	if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t4\n" {
		t.Fatalf("version after restart: %q", out)
	}
	rt := runtimeConn(t, cfg)
	if got := lines(t, rt, `SELECT organization_id::text || ' ' || display_name || ' ' || coalesce(website, '-') FROM app.business_profiles WHERE id = $1::text::uuid`, f[3]); got !=
		f[0]+" Persisted Profile https://example.com/" {
		t.Fatalf("profile after restart: %q", got)
	}
	if got := lines(t, rt, `SELECT organization_id::text || ' ' || group_id::text || ' ' || business_profile_id::text FROM app.group_business_profiles WHERE id = $1::text::uuid`, f[4]); got !=
		f[0]+" "+f[1]+" "+f[3] {
		t.Fatalf("link after restart: %q", got)
	}
	if got := lines(t, rt, `SELECT user_id::text || ' ' || group_role FROM app.business_profile_access_paths
WHERE organization_id = $1::text::uuid AND business_profile_id = $2::text::uuid AND group_id = $3::text::uuid`, f[0], f[3], f[1]); got != f[2]+" group_owner" {
		t.Fatalf("access path after restart: %q", got)
	}
}
