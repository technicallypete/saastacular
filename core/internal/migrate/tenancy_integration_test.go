//go:build integration

// G005 tenancy catalog and migration tests against a disposable PostgreSQL initialized by the rendered role script
// (infra/validation/run.sh, suite "migrations"): B-TENANCY-CATALOG (TestTenancyCatalog) and B-TENANCY-MIGRATION
// (TestTenancyMigration, then TestTenancyPersistSeed and TestTenancyPersistVerify around a PostgreSQL restart). They
// fail, never skip, when the database settings are absent. Model operations run as the actual runtime role; owner-level
// cases use the migrator with SET ROLE <db>_owner. Since G006 these predecessor tests run the normal set up to version
// 3 (runnerFor with subsetFS 3), so that every G005 assertion keeps its exact version-3 meaning.
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, APP_ENV; the
// persist phases also TENANCY_PERSIST_FILE.
package migrate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/tenancy"
	"core/migrations"
)

// tenancyTables, tenancyViews and tenancyFunctions are the objects of 00003_tenancy.sql besides its eight indexes.
var (
	tenancyTables    = []string{"organizations", "groups", "role_bindings"}
	tenancyViews     = []string{"organization_memberships", "group_memberships"}
	tenancyFunctions = []string{"groups_keys_immutable", "organizations_keys_immutable", "role_bindings_keys_immutable"}
)

// identityFixtureTruncate is the G004 identity fixtures' owner reset (core/internal/identity/integration_test.go and
// core/internal/identity/workos/staging_test.go), which needs CASCADE once role_bindings references app.users.
const identityFixtureTruncate = `TRUNCATE app.login_transactions, app.sessions, app.identity_principals, app.users, app.identity_provider_binding CASCADE`

// catalogConn connects as the migrator, assumes the owner role and pins search_path to pg_catalog, so that every
// catalog definition is printed schema-qualified and deterministic.
func catalogConn(t *testing.T, cfg Config) *pgx.Conn {
	t.Helper()
	c := connectAs(t, cfg, cfg.MigratorRole(), cfg.Password)
	if _, err := c.Exec(context.Background(), "SET ROLE "+quote(cfg.OwnerRole())+"; SET search_path = pg_catalog"); err != nil {
		t.Fatal(err)
	}
	return c
}

func runtimeConn(t *testing.T, cfg Config) *pgx.Conn {
	t.Helper()
	return connectAs(t, cfg, cfg.RuntimeRole(), os.Getenv("DB_RUNTIME_PASSWORD"))
}

func runtimePool(t *testing.T, cfg Config) *pgxpool.Pool {
	t.Helper()
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = cfg.Host, uint16(cfg.Port), cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = cfg.RuntimeRole(), os.Getenv("DB_RUNTIME_PASSWORD")
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	p, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

// subsetFS is the normal set up to and including maxVersion (a populated-upgrade starting point).
func subsetFS(t *testing.T, maxVersion int) fs.FS {
	t.Helper()
	out := fstest.MapFS{}
	names, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		v, err := strconv.Atoi(n[:5])
		if err != nil {
			t.Fatal(err)
		}
		if v <= maxVersion {
			b, err := fs.ReadFile(migrations.FS, n)
			if err != nil {
				t.Fatal(err)
			}
			out[n] = &fstest.MapFile{Data: b}
		}
	}
	return out
}

func runnerFor(t *testing.T, cfg Config, fsys fs.FS) *Runner {
	t.Helper()
	r, err := open(context.Background(), cfg, fsys, openOptions{})
	if err != nil {
		t.Fatalf("open runner: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func normalize(s string) string { return strings.Join(strings.Fields(s), " ") }

func sqlstateOf(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// lines runs a query returning one text column and joins the non-NULL rows with newlines.
func lines(t *testing.T, q catalogQuerier, sql string, args ...any) string {
	t.Helper()
	rows, err := q.Query(context.Background(), sql, args...)
	if err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	vals, err := pgx.CollectRows(rows, pgx.RowTo[*string])
	if err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	var out []string
	for _, v := range vals {
		if v != nil {
			out = append(out, *v)
		}
	}
	return strings.Join(out, "\n")
}

type catalogQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// tenancyCatalog is the CT-1..CT-5 catalog snapshot of the tenancy objects: relations with owner and table ACL, column
// ACLs, constraints, indexes, views, functions and triggers. It is also MG-4's equality basis.
func tenancyCatalog(t *testing.T, q catalogQuerier) string {
	t.Helper()
	parts := []string{
		lines(t, q, `SELECT c.relkind::text || ':' || c.relname || '|' || c.relowner::regrole::text || '|' || coalesce(c.relacl::text, '')
FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace ORDER BY c.relname`),
		lines(t, q, `SELECT c.relname || '.' || a.attname || '|' || a.attacl::text FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE c.relnamespace = 'app'::regnamespace AND a.attacl IS NOT NULL ORDER BY c.relname, a.attname`),
		lines(t, q, `SELECT conrelid::regclass::text || '|' || conname || '|' || pg_get_constraintdef(oid) FROM pg_constraint
WHERE connamespace = 'app'::regnamespace ORDER BY conrelid::regclass::text, conname`),
		lines(t, q, `SELECT pg_get_indexdef(i.indexrelid) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
WHERE c.relnamespace = 'app'::regnamespace ORDER BY c.relname`),
		lines(t, q, `SELECT c.relname || '|' || coalesce(c.reloptions::text, '') || '|' || pg_get_viewdef(c.oid) FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relkind = 'v' ORDER BY c.relname`),
		lines(t, q, `SELECT p.proname || '|' || p.prosecdef::text || '|' || coalesce(p.proconfig::text, '') || '|' || coalesce(p.proacl::text, '')
       || '|' || p.proowner::regrole::text || '|' || md5(p.prosrc) FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace ORDER BY p.proname`),
		lines(t, q, `SELECT pg_get_triggerdef(g.oid) || '|' || g.tgenabled::text FROM pg_trigger g JOIN pg_class c ON c.oid = g.tgrelid
WHERE c.relnamespace = 'app'::regnamespace AND NOT g.tgisinternal ORDER BY g.tgname`),
	}
	return strings.Join(parts, "\n--\n")
}

// --- CT-6: one durable authority versus exactly two read-only projections ---------------------------------------------

// The role-bearing and owner-bearing column rules (Plan r4 §5.2).
const authorityColumnsSQL = `
SELECT n.nspname || '.' || c.relname, c.relkind::text, a.attname,
       (a.attname ~ '^(role|roles)$' OR a.attname ~ '_roles?$'),
       (a.attname ~ '(^|_)(owner|owners|creator)(_|$)' OR a.attname ~ '^created_by(_|$)')
FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped
WHERE c.relkind IN ('r', 'p', 'm', 'f', 'v')
  AND (n.nspname IN ('app', 'public') OR c.relowner::regrole::text = ANY ($1::text[]))
  AND (a.attname ~ '^(role|roles)$' OR a.attname ~ '_roles?$'
       OR a.attname ~ '(^|_)(owner|owners|creator)(_|$)' OR a.attname ~ '^created_by(_|$)')
ORDER BY 1, 3`

// projectionPin is one pinned read-only role-bearing projection: its columns, its normalized pg_get_viewdef (search_path
// pg_catalog) and the exact set of relations its _RETURN rule reads.
type projectionPin struct {
	cols, def string
	sources   []string
}

// The pinned projections of version 3 (G005): each reads role_bindings only.
var pinnedProjections = map[string]projectionPin{
	"app.organization_memberships": {"user_id,organization_id,role",
		"SELECT subject_user_id AS user_id, organization_id, role FROM app.role_bindings WHERE ((scope_type = 'organization'::text) AND (source = 'membership'::text));",
		[]string{"app.role_bindings"}},
	"app.group_memberships": {"user_id,organization_id,group_id,role",
		"SELECT subject_user_id AS user_id, organization_id, group_id, role FROM app.role_bindings WHERE ((scope_type = 'group'::text) AND (source = 'membership'::text));",
		[]string{"app.role_bindings"}},
}

// authorityInventory is the expected role-bearing projection set of one schema version: the projection role columns
// ("app.<view>.<column>") and each projection's pin. Durable storage is role_bindings.role at every version.
type authorityInventory struct {
	roleColumns []string
	projections map[string]projectionPin
}

// tenancyAuthority is the version-3 inventory (G005): exactly the two membership projections.
var tenancyAuthority = authorityInventory{
	roleColumns: []string{"app.organization_memberships.role", "app.group_memberships.role"},
	projections: pinnedProjections,
}

const projectionSQL = `
SELECT c.relkind::text, coalesce(c.reloptions::text, ''),
  (SELECT string_agg(a.attname, ',' ORDER BY a.attnum) FROM pg_catalog.pg_attribute a WHERE a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped),
  CASE WHEN c.relkind = 'v' THEN pg_catalog.pg_get_viewdef(c.oid) ELSE '' END,
  (SELECT count(*) FROM pg_catalog.pg_rewrite rw WHERE rw.ev_class = c.oid AND rw.rulename <> '_RETURN'),
  (SELECT count(*) FROM pg_catalog.pg_trigger tg WHERE tg.tgrelid = c.oid),
  coalesce((SELECT string_agg(DISTINCT d.refobjid::regclass::text, ',' ORDER BY d.refobjid::regclass::text) FROM pg_catalog.pg_rewrite rw
     JOIN pg_catalog.pg_depend d ON d.classid = 'pg_catalog.pg_rewrite'::regclass AND d.objid = rw.oid AND d.refclassid = 'pg_catalog.pg_class'::regclass
     WHERE rw.ev_class = c.oid AND rw.rulename = '_RETURN' AND d.refobjid <> c.oid), ''),
  EXISTS (SELECT 1 FROM pg_catalog.pg_rewrite rw
     JOIN pg_catalog.pg_depend d ON d.classid = 'pg_catalog.pg_rewrite'::regclass AND d.objid = rw.oid
       AND d.refclassid = 'pg_catalog.pg_class'::regclass AND d.refobjid = 'app.role_bindings'::regclass
     JOIN pg_catalog.pg_attribute a ON a.attrelid = d.refobjid AND a.attnum = d.refobjsubid
     WHERE rw.ev_class = c.oid AND rw.rulename = '_RETURN' AND a.attname = 'role')
FROM pg_catalog.pg_class c WHERE c.oid = pg_catalog.to_regclass($1::text)`

const tableWriteSQL = `SELECT p FROM unnest(ARRAY['INSERT', 'UPDATE', 'DELETE', 'TRUNCATE']) p
WHERE has_table_privilege($1::text::name, $2::text::regclass, p) ORDER BY p`
const columnWriteSQL = `SELECT col || ':' || p FROM unnest($3::text[]) col, unnest(ARRAY['INSERT', 'UPDATE']) p
WHERE has_column_privilege($1::text::name, $2::text::regclass, col, p) ORDER BY 1`
const publicWriteSQL = `
SELECT 'table:' || a.privilege_type FROM pg_catalog.pg_class c,
  pg_catalog.aclexplode(coalesce(c.relacl, pg_catalog.acldefault('r', c.relowner))) a
WHERE c.oid = $1::text::regclass AND a.grantee = 0 AND a.privilege_type IN ('INSERT', 'UPDATE', 'DELETE', 'TRUNCATE')
UNION ALL
SELECT at.attname || ':' || a.privilege_type FROM pg_catalog.pg_attribute at, pg_catalog.aclexplode(at.attacl) a
WHERE at.attrelid = $1::text::regclass AND at.attname = ANY ($2::text[]) AND a.grantee = 0 AND a.privilege_type IN ('INSERT', 'UPDATE')`

// authorityViolations is the CT-6 detector. It returns the sorted violation codes; the empty set means that role_bindings
// is the single durable authority, the inventory's projections are its only role-bearing projections (exact
// definitions, exactly the expected source relations, the role column sourced from role_bindings.role, no write path),
// and no role or PUBLIC can write a role-bearing relation except the runtime role's INSERT/UPDATE of role_bindings.role
// and DELETE on role_bindings. G005 passes the version-3 inventory (tenancyAuthority); G006 the version-4 one.
func authorityViolations(t *testing.T, q catalogQuerier, cfg Config, inv authorityInventory) []string {
	t.Helper()
	ctx := context.Background()
	appRoles := []string{cfg.OwnerRole(), cfg.MigratorRole(), cfg.RuntimeRole(), cfg.RetentionRole()}
	rows, err := q.Query(ctx, authorityColumnsSQL, appRoles)
	if err != nil {
		t.Fatalf("authority columns: %v", err)
	}
	type relInfo struct {
		kind string
		cols []string
	}
	storageRole, projectionRole := map[string]bool{}, map[string]bool{}
	ownerCols := 0
	roleRels := map[string]*relInfo{}
	for rows.Next() {
		var rel, kind, col string
		var roleBearing, ownerBearing bool
		if err := rows.Scan(&rel, &kind, &col, &roleBearing, &ownerBearing); err != nil {
			t.Fatal(err)
		}
		if roleBearing {
			if kind == "v" {
				projectionRole[rel+"."+col] = true
			} else {
				storageRole[rel+"."+col] = true
			}
			if roleRels[rel] == nil {
				roleRels[rel] = &relInfo{kind: kind}
			}
			roleRels[rel].cols = append(roleRels[rel].cols, col)
		}
		if ownerBearing && kind != "v" {
			ownerCols++
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	setIs := func(m map[string]bool, want ...string) bool {
		if len(m) != len(want) {
			return false
		}
		for _, w := range want {
			if !m[w] {
				return false
			}
		}
		return true
	}
	// A1 durable storage.
	if !setIs(storageRole, "app.role_bindings.role") || ownerCols != 0 {
		codes["stored_role_outside_authority"] = true
	}
	// A2 projections.
	if !setIs(projectionRole, inv.roleColumns...) {
		codes["unexpected_role_projection"] = true
	}
	for name, pin := range inv.projections {
		var kind, opts, cols, def, sources string
		var otherRules, triggers int
		var roleSource bool
		err := q.QueryRow(ctx, projectionSQL, name).Scan(&kind, &opts, &cols, &def, &otherRules, &triggers, &sources, &roleSource)
		if errors.Is(err, pgx.ErrNoRows) {
			continue // absent: the projection set above already differs
		}
		if err != nil {
			t.Fatalf("projection %s: %v", name, err)
		}
		if kind != "v" || opts != "{security_invoker=true}" || cols != pin.cols || normalize(def) != pin.def {
			codes["projection_definition"] = true
		}
		if sources != strings.Join(pin.sources, ",") || !roleSource {
			codes["projection_source"] = true
		}
		if otherRules != 0 || triggers != 0 {
			codes["projection_write_path"] = true
		}
	}
	// A3 no other write path: named application roles through the privilege functions, PUBLIC through its ACL entries.
	allowed := map[string]bool{
		cfg.RuntimeRole() + "|app.role_bindings|DELETE":      true,
		cfg.RuntimeRole() + "|app.role_bindings|role:INSERT": true,
		cfg.RuntimeRole() + "|app.role_bindings|role:UPDATE": true,
	}
	for rel, info := range roleRels {
		code := "writable_role_column"
		if info.kind == "v" {
			code = "writable_projection"
		}
		var found []string
		for _, role := range []string{cfg.RuntimeRole(), cfg.RetentionRole(), cfg.MigratorRole()} {
			for _, p := range strings.Split(lines(t, q, tableWriteSQL, role, rel), "\n") {
				if p != "" {
					found = append(found, role+"|"+rel+"|"+p)
				}
			}
			for _, p := range strings.Split(lines(t, q, columnWriteSQL, role, rel, info.cols), "\n") {
				if p != "" {
					found = append(found, role+"|"+rel+"|"+p)
				}
			}
		}
		for _, p := range strings.Split(lines(t, q, publicWriteSQL, rel, info.cols), "\n") {
			if p != "" {
				found = append(found, "PUBLIC|"+rel+"|"+p)
			}
		}
		for _, f := range found {
			if !allowed[f] {
				codes[code] = true
			}
		}
	}
	out := make([]string, 0, len(codes))
	for c := range codes {
		out = append(out, c)
	}
	slices.Sort(out)
	return out
}

// --- B-TENANCY-CATALOG ---------------------------------------------------------------------------------------------

func TestTenancyCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r := runnerFor(t, cfg, subsetFS(t, 3))
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out !=
		"applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\ncurrent_version\t3\n" {
		t.Fatalf("up: %q", out)
	}
	own := catalogConn(t, cfg)
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

	t.Run("CT-1 object set and columns", func(t *testing.T) {
		metadataIdentityAndTenancyOnly(t, cfg, r)
		// No duplicate tenant, Agency/ClientOrganization, BusinessProfile, billing, policy, label, domain, onboarding,
		// staffing or audit object in app or public.
		if got := lines(t, own, `SELECT c.relname FROM pg_class c WHERE c.relnamespace IN ('app'::regnamespace, 'public'::regnamespace)
  AND c.relname ~ '(tenant|agenc|client_org|business_profile|business_entit|billing|polic|label|domain|onboard|staff|engagement|audit|invitation)'`); got != "" {
			t.Fatalf("forbidden objects: %s", got)
		}
		if got := lines(t, own, `SELECT count(*)::text FROM pg_policy`); got != "0" {
			t.Fatalf("policies: %s", got)
		}
		wantCols := map[string]string{
			"organizations": "id uuid not null default gen_random_uuid(), name text not null, created_at timestamp with time zone not null default now()",
			"groups": "id uuid not null default gen_random_uuid(), organization_id uuid not null, name text not null, " +
				"slug text collate \"C\" not null, created_at timestamp with time zone not null default now()",
			"role_bindings": "id uuid not null default gen_random_uuid(), organization_id uuid not null, subject_user_id uuid not null, " +
				"scope_type text not null, scope_id uuid not null, group_id uuid, role text not null, source text not null, " +
				"created_at timestamp with time zone not null default now()",
		}
		for tbl, want := range wantCols {
			got := lines(t, own, `SELECT string_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod)
  || CASE WHEN a.attcollation <> 0 AND a.attcollation <> (SELECT oid FROM pg_collation WHERE collname = 'default') THEN ' collate "' || co.collname || '"' ELSE '' END
  || CASE WHEN a.attnotnull THEN ' not null' ELSE '' END
  || coalesce(' default ' || pg_get_expr(d.adbin, d.adrelid), ''), ', ' ORDER BY a.attnum)
FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
LEFT JOIN pg_collation co ON co.oid = a.attcollation
WHERE a.attrelid = ('app.' || $1::text)::regclass AND a.attnum > 0 AND NOT a.attisdropped`, tbl)
			if got != want {
				t.Errorf("%s columns:\n got %s\nwant %s", tbl, got, want)
			}
		}
	})

	t.Run("CT-2 ownership and functions", func(t *testing.T) {
		got := lines(t, own, `SELECT c.relname || ':' || c.relowner::regrole::text FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($1::text[]) ORDER BY c.relname`, append(slices.Clone(tenancyTables), tenancyViews...))
		want := []string{"group_memberships:" + o, "groups:" + o, "organization_memberships:" + o, "organizations:" + o, "role_bindings:" + o}
		if got != strings.Join(want, "\n") {
			t.Fatalf("owners:\n%s", got)
		}
		got = lines(t, own, `SELECT p.proname || '|' || p.prosecdef::text || '|' || coalesce(p.proconfig::text, '') || '|' || coalesce(p.proacl::text, '')
  || '|' || p.proowner::regrole::text || '|' || p.prorettype::regtype::text || '|' || l.lanname
FROM pg_proc p JOIN pg_language l ON l.oid = p.prolang WHERE p.pronamespace = 'app'::regnamespace ORDER BY p.proname`)
		var w []string
		for _, f := range tenancyFunctions {
			w = append(w, fmt.Sprintf("%s|false|{search_path=pg_catalog}|{%s=X/%s}|%s|trigger|plpgsql", f, o, o, o))
		}
		if got != strings.Join(w, "\n") {
			t.Fatalf("functions:\n got %s\nwant %s", got, strings.Join(w, "\n"))
		}
		for _, role := range []string{rtr, cfg.RetentionRole(), cfg.MigratorRole()} {
			if got := lines(t, own, `SELECT string_agg(p.proname, ',') FROM pg_proc p
WHERE p.pronamespace = 'app'::regnamespace AND has_function_privilege($1::text::name, p.oid, 'EXECUTE')`, role); got != "" {
				t.Errorf("%s may execute %s", role, got)
			}
		}
	})

	t.Run("CT-3 table and column ACLs", func(t *testing.T) {
		got := lines(t, own, `SELECT c.relname || '|' || c.relacl::text FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($1::text[]) ORDER BY c.relname`, append(slices.Clone(tenancyTables), tenancyViews...))
		acl := func(rt string) string { return fmt.Sprintf("{%s=arwdDxtm/%s,%s=%s/%s}", o, o, rtr, rt, o) }
		want := strings.Join([]string{
			"group_memberships|" + acl("r"), "groups|" + acl("r"), "organization_memberships|" + acl("r"),
			"organizations|" + acl("r"), "role_bindings|" + acl("rd"),
		}, "\n")
		if got != want {
			t.Fatalf("table ACLs:\n got %s\nwant %s", got, want)
		}
		got = lines(t, own, `SELECT c.relname || '.' || a.attname || '|' || a.attacl::text FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE c.relnamespace = 'app'::regnamespace AND a.attacl IS NOT NULL ORDER BY c.relname, a.attname`)
		col := func(p string) string { return fmt.Sprintf("{%s=%s/%s}", rtr, p, o) }
		want = strings.Join([]string{
			"groups.name|" + col("aw"), "groups.organization_id|" + col("a"), "groups.slug|" + col("aw"),
			"organizations.name|" + col("aw"),
			"role_bindings.group_id|" + col("a"), "role_bindings.organization_id|" + col("a"), "role_bindings.role|" + col("aw"),
			"role_bindings.scope_id|" + col("a"), "role_bindings.scope_type|" + col("a"), "role_bindings.source|" + col("a"),
			"role_bindings.subject_user_id|" + col("a"),
		}, "\n")
		if got != want {
			t.Fatalf("column ACLs:\n got %s\nwant %s", got, want)
		}
		// The effective privilege matrix: runtime exactly as granted; retention and migrator (without SET ROLE) nothing.
		matrix := `SELECT string_agg(c.relname || ':' || p, ',' ORDER BY c.relname, p) FROM pg_class c,
  unnest(ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'TRUNCATE', 'REFERENCES', 'TRIGGER', 'MAINTAIN']) p
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($2::text[]) AND has_table_privilege($1::text::name, c.oid, p)`
		cols := `SELECT string_agg(c.relname || '.' || a.attname || ':' || p, ',' ORDER BY c.relname, a.attname, p)
FROM pg_class c JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped,
  unnest(ARRAY['INSERT', 'UPDATE', 'REFERENCES']) p
WHERE c.relnamespace = 'app'::regnamespace AND c.relname = ANY ($2::text[]) AND has_column_privilege($1::text::name, c.oid, a.attnum, p)`
		rels := append(slices.Clone(tenancyTables), tenancyViews...)
		if got := lines(t, own, matrix, rtr, rels); got != "group_memberships:SELECT,groups:SELECT,organization_memberships:SELECT,organizations:SELECT,role_bindings:DELETE,role_bindings:SELECT" {
			t.Errorf("runtime table privileges: %s", got)
		}
		if got := lines(t, own, cols, rtr, rels); got != "groups.name:INSERT,groups.name:UPDATE,groups.organization_id:INSERT,groups.slug:INSERT,groups.slug:UPDATE,"+
			"organizations.name:INSERT,organizations.name:UPDATE,role_bindings.group_id:INSERT,role_bindings.organization_id:INSERT,"+
			"role_bindings.role:INSERT,role_bindings.role:UPDATE,role_bindings.scope_id:INSERT,role_bindings.scope_type:INSERT,"+
			"role_bindings.source:INSERT,role_bindings.subject_user_id:INSERT" {
			t.Errorf("runtime column privileges: %s", got)
		}
		for _, role := range []string{cfg.RetentionRole(), cfg.MigratorRole()} {
			if got := lines(t, own, matrix, role, rels) + lines(t, own, cols, role, rels); got != "" {
				t.Errorf("%s holds %s", role, got)
			}
		}
	})

	t.Run("CT-4 constraints", func(t *testing.T) {
		got := lines(t, own, `SELECT conrelid::regclass::text || ' ' || conname || ' ' || contype::text || ' ' || pg_get_constraintdef(oid)
FROM pg_constraint WHERE conrelid = ANY (ARRAY['app.organizations'::regclass, 'app.groups'::regclass, 'app.role_bindings'::regclass])
ORDER BY conrelid::regclass::text, conname`)
		want := []string{
			`app.groups groups_id_nonzero c CHECK ((id <> '00000000-0000-0000-0000-000000000000'::uuid))`,
			`app.groups groups_name_valid c CHECK ((((char_length(name) >= 1) AND (char_length(name) <= 200)) AND (name = btrim(name)) AND ((name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'::text)))`,
			`app.groups groups_organization_fkey f FOREIGN KEY (organization_id) REFERENCES app.organizations(id) ON DELETE RESTRICT`,
			`app.groups groups_organization_id_key u UNIQUE (organization_id, id)`,
			`app.groups groups_organization_slug_key u UNIQUE (organization_id, slug)`,
			`app.groups groups_pkey p PRIMARY KEY (id)`,
			`app.groups groups_slug_format c CHECK ((((char_length(slug) >= 1) AND (char_length(slug) <= 63)) AND (octet_length(slug) = char_length(slug)) AND (slug ~ '^[0-9a-z]+(-[0-9a-z]+)*$'::text)))`,
			`app.groups groups_slug_reserved c CHECK ((slug <> ALL (ARRAY['account'::text, 'admin'::text, 'api'::text, 'auth'::text, 'billing'::text, 'blueprint'::text, 'fonts'::text, 'health'::text, 'healthz'::text, 'internal'::text, 'login'::text, 'logout'::text, 'mcp'::text, 'members'::text, 'message'::text, 'messages'::text, 'new'::text, 'onboarding'::text, 'readyz'::text, 'rpc'::text, 'settings'::text, 'sign-in'::text, 'sse'::text, 'system'::text])))`,
			`app.organizations organizations_id_nonzero c CHECK ((id <> '00000000-0000-0000-0000-000000000000'::uuid))`,
			`app.organizations organizations_name_valid c CHECK ((((char_length(name) >= 1) AND (char_length(name) <= 200)) AND (name = btrim(name)) AND ((name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'::text)))`,
			`app.organizations organizations_pkey p PRIMARY KEY (id)`,
			`app.role_bindings role_bindings_group_fkey f FOREIGN KEY (organization_id, group_id) REFERENCES app.groups(organization_id, id) ON DELETE RESTRICT`,
			`app.role_bindings role_bindings_id_nonzero c CHECK ((id <> '00000000-0000-0000-0000-000000000000'::uuid))`,
			`app.role_bindings role_bindings_organization_fkey f FOREIGN KEY (organization_id) REFERENCES app.organizations(id) ON DELETE RESTRICT`,
			`app.role_bindings role_bindings_pkey p PRIMARY KEY (id)`,
			`app.role_bindings role_bindings_scope_consistent c CHECK ((((scope_type = 'organization'::text) AND (scope_id = organization_id) AND (group_id IS NULL) AND (role = ANY (ARRAY['org_owner'::text, 'org_admin'::text, 'org_member'::text])) AND (source = 'membership'::text)) OR ((scope_type = 'group'::text) AND (group_id IS NOT NULL) AND (scope_id = group_id) AND (role = ANY (ARRAY['group_owner'::text, 'group_admin'::text, 'group_member'::text])) AND (source = 'membership'::text))))`,
			`app.role_bindings role_bindings_subject_fkey f FOREIGN KEY (subject_user_id) REFERENCES app.users(id) ON DELETE RESTRICT`,
			`app.role_bindings role_bindings_subject_nonzero c CHECK ((subject_user_id <> '00000000-0000-0000-0000-000000000000'::uuid))`,
			`app.role_bindings role_bindings_subject_scope_key u UNIQUE (subject_user_id, scope_type, scope_id)`,
		}
		var gotN []string
		for _, l := range strings.Split(got, "\n") {
			gotN = append(gotN, normalize(l))
		}
		if strings.Join(gotN, "\n") != strings.Join(want, "\n") {
			t.Fatalf("constraints:\n got %s\nwant %s", strings.Join(gotN, "\n"), strings.Join(want, "\n"))
		}
		if got := lines(t, own, `SELECT string_agg(conname || ':' || confdeltype::text || confupdtype::text || confmatchtype::text, ',' ORDER BY conname)
FROM pg_constraint WHERE contype = 'f' AND conrelid = ANY (ARRAY['app.groups'::regclass, 'app.role_bindings'::regclass])`); got !=
			"groups_organization_fkey:ras,role_bindings_group_fkey:ras,role_bindings_organization_fkey:ras,role_bindings_subject_fkey:ras" {
			t.Fatalf("foreign keys (delete restrict, update no action, match simple): %s", got)
		}
	})

	t.Run("CT-5 indexes views and triggers", func(t *testing.T) {
		got := lines(t, own, `SELECT pg_get_indexdef(i.indexrelid) FROM pg_index i
WHERE i.indrelid = ANY (ARRAY['app.organizations'::regclass, 'app.groups'::regclass, 'app.role_bindings'::regclass]) ORDER BY i.indexrelid::regclass::text`)
		want := strings.Join([]string{
			"CREATE UNIQUE INDEX groups_organization_id_key ON app.groups USING btree (organization_id, id)",
			"CREATE UNIQUE INDEX groups_organization_slug_key ON app.groups USING btree (organization_id, slug)",
			"CREATE UNIQUE INDEX groups_pkey ON app.groups USING btree (id)",
			"CREATE UNIQUE INDEX organizations_pkey ON app.organizations USING btree (id)",
			"CREATE INDEX role_bindings_group_idx ON app.role_bindings USING btree (organization_id, group_id) WHERE (group_id IS NOT NULL)",
			"CREATE UNIQUE INDEX role_bindings_pkey ON app.role_bindings USING btree (id)",
			"CREATE INDEX role_bindings_scope_idx ON app.role_bindings USING btree (organization_id, scope_type, scope_id)",
			"CREATE UNIQUE INDEX role_bindings_subject_scope_key ON app.role_bindings USING btree (subject_user_id, scope_type, scope_id)",
		}, "\n")
		if got != want {
			t.Fatalf("indexes:\n got %s\nwant %s", got, want)
		}
		for name, pin := range pinnedProjections {
			var opts, def string
			if err := own.QueryRow(ctx, `SELECT coalesce(reloptions::text, ''), pg_get_viewdef(oid) FROM pg_class WHERE oid = $1::text::regclass`, name).Scan(&opts, &def); err != nil {
				t.Fatal(err)
			}
			if opts != "{security_invoker=true}" || normalize(def) != pin.def {
				t.Errorf("view %s: options %s definition %q", name, opts, normalize(def))
			}
		}
		got = lines(t, own, `SELECT pg_get_triggerdef(g.oid) || ' ' || g.tgenabled::text FROM pg_trigger g
WHERE g.tgrelid = ANY (ARRAY['app.organizations'::regclass, 'app.groups'::regclass, 'app.role_bindings'::regclass]) AND NOT g.tgisinternal ORDER BY g.tgname`)
		want = strings.Join([]string{
			"CREATE TRIGGER groups_keys_immutable BEFORE UPDATE ON app.groups FOR EACH ROW EXECUTE FUNCTION app.groups_keys_immutable() O",
			"CREATE TRIGGER organizations_keys_immutable BEFORE UPDATE ON app.organizations FOR EACH ROW EXECUTE FUNCTION app.organizations_keys_immutable() O",
			"CREATE TRIGGER role_bindings_keys_immutable BEFORE UPDATE ON app.role_bindings FOR EACH ROW EXECUTE FUNCTION app.role_bindings_keys_immutable() O",
		}, "\n")
		if got != want {
			t.Fatalf("triggers:\n got %s\nwant %s", got, want)
		}
	})

	t.Run("CT-6 single authority", func(t *testing.T) {
		if v := authorityViolations(t, own, cfg, tenancyAuthority); len(v) != 0 {
			t.Fatalf("authority violations on the migrated schema: %v", v)
		}
	})

	t.Run("CT-6-NEG detector controls", func(t *testing.T) {
		before := tenancyCatalog(t, own)
		rtID := quote(rtr)
		for _, c := range []struct {
			name, ddl string
			want      []string
		}{
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
				if got := authorityViolations(t, tx, cfg, tenancyAuthority); strings.Join(got, ",") != strings.Join(c.want, ",") {
					t.Fatalf("violations %v, want exactly %v", got, c.want)
				}
			})
		}
		if v := authorityViolations(t, own, cfg, tenancyAuthority); len(v) != 0 {
			t.Fatalf("violations after the rolled-back controls: %v", v)
		}
		if after := tenancyCatalog(t, own); after != before {
			t.Fatalf("catalog changed by the rolled-back controls:\n%s\n---\n%s", before, after)
		}
	})

	t.Run("CT-7 no row-level security yet", func(t *testing.T) {
		// G008's obligation, recorded in docs/foundation/TENANCY.md; this is not isolation evidence.
		if got := lines(t, own, `SELECT string_agg(relname || ':' || relrowsecurity::text || relforcerowsecurity::text, ',' ORDER BY relname)
FROM pg_class WHERE oid = ANY (ARRAY['app.organizations'::regclass, 'app.groups'::regclass, 'app.role_bindings'::regclass])`); got !=
			"groups:falsefalse,organizations:falsefalse,role_bindings:falsefalse" {
			t.Fatalf("row security: %s", got)
		}
		if got := lines(t, own, `SELECT count(*)::text FROM pg_policy WHERE polrelid = ANY (ARRAY['app.organizations'::regclass, 'app.groups'::regclass, 'app.role_bindings'::regclass])`); got != "0" {
			t.Fatalf("policies: %s", got)
		}
	})

	t.Run("runtime DDL, role and metadata denials", func(t *testing.T) {
		for what, stmt := range map[string]string{
			"CREATE TABLE":         `CREATE TABLE app.runtime_tenancy (id int)`,
			"ALTER groups":         `ALTER TABLE app.groups ADD COLUMN x int`,
			"ALTER role_bindings":  `ALTER TABLE app.role_bindings DROP CONSTRAINT role_bindings_scope_consistent`,
			"DROP organizations":   `DROP TABLE app.organizations`,
			"DROP view":            `DROP VIEW app.group_memberships`,
			"TRUNCATE bindings":    `TRUNCATE app.role_bindings`,
			"TRUNCATE groups":      `TRUNCATE app.groups`,
			"CREATE TRIGGER":       `CREATE TRIGGER x BEFORE UPDATE ON app.groups FOR EACH ROW EXECUTE FUNCTION app.groups_keys_immutable()`,
			"DISABLE TRIGGER":      `ALTER TABLE app.groups DISABLE TRIGGER groups_keys_immutable`,
			"SET ROLE owner":       `SET ROLE ` + quote(cfg.OwnerRole()),
			"SET ROLE migrator":    `SET ROLE ` + quote(cfg.MigratorRole()),
			"metadata read":        `SELECT 1 FROM app.goose_db_version`,
			"DELETE organizations": `DELETE FROM app.organizations`,
			"DELETE groups":        `DELETE FROM app.groups`,
		} {
			expectDenied(t, rt, "runtime "+what, stmt)
		}
	})
}

// --- B-TENANCY-MIGRATION -------------------------------------------------------------------------------------------

// identitySnapshot hashes the four G004 identity tables that runtime fixtures populate.
func identitySnapshot(t *testing.T, q catalogQuerier) string {
	t.Helper()
	return lines(t, q, `SELECT 'users:' || coalesce(md5(string_agg(u::text, E'\n' ORDER BY u.id)), '') FROM app.users u
UNION ALL SELECT 'principals:' || coalesce(md5(string_agg(p::text, E'\n' ORDER BY p.issuer, p.subject)), '') FROM app.identity_principals p
UNION ALL SELECT 'sessions:' || coalesce(md5(string_agg(s::text, E'\n' ORDER BY s.id)), '') FROM app.sessions s
UNION ALL SELECT 'binding:' || coalesce(md5(string_agg(b::text, E'\n' ORDER BY b.provider)), '') FROM app.identity_provider_binding b`)
}

// seedIdentity inserts n synthetic users with a principal and a live session each (and the provider binding once), as
// the runtime role, and returns the user IDs.
func seedIdentity(t *testing.T, rt *pgx.Conn, n int, tag string) []string {
	t.Helper()
	ctx := context.Background()
	if _, err := rt.Exec(ctx, `INSERT INTO app.identity_provider_binding (provider, issuer, client_id)
VALUES ('workos', 'https://issuer.example.invalid', 'client_01TENANCY') ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("seed binding: %v", err)
	}
	var ids []string
	for i := range n {
		var id string
		if err := rt.QueryRow(ctx, `INSERT INTO app.users (display_name) VALUES ($1) RETURNING id::text`, fmt.Sprintf("Fixture %s %d", tag, i)).Scan(&id); err != nil {
			t.Fatalf("seed user: %v", err)
		}
		if _, err := rt.Exec(ctx, `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('https://issuer.example.invalid', $1, $2)`,
			fmt.Sprintf("user_%s%d%s", tag, i, id[:8]), id); err != nil {
			t.Fatalf("seed principal: %v", err)
		}
		if _, err := rt.Exec(ctx, `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at)
VALUES ($1::text::uuid, sha256(convert_to($2::text || 'c', 'UTF8')), sha256(convert_to($2::text || 'x', 'UTF8')), now(), now() + interval '30 minutes', now() + interval '8 hours')`, id, id); err != nil {
			t.Fatalf("seed session: %v", err)
		}
		ids = append(ids, id)
	}
	return ids
}

func tenancyObjectsAbsent(t *testing.T, q catalogQuerier) {
	t.Helper()
	got := lines(t, q, `SELECT n FROM unnest($1::text[]) n WHERE to_regclass('app.' || n) IS NOT NULL
UNION ALL SELECT p.proname FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace`,
		append(slices.Clone(tenancyTables), tenancyViews...))
	if got != "" {
		t.Fatalf("tenancy objects present after down: %s", got)
	}
}

func TestTenancyMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	const full = "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\ncurrent_version\t3\n"

	t.Run("MG-1 empty to full", func(t *testing.T) {
		reset(t, cfg)
		r := runnerFor(t, cfg, subsetFS(t, 3))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != full {
			t.Fatalf("up: %q", out)
		}
		metadataIdentityAndTenancyOnly(t, cfg, r)
	})

	t.Run("MG-2 repeated up and status", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 3))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "no pending migrations\ncurrent_version\t3\n" {
			t.Fatalf("repeated up: %q", out)
		}
		out := run(t, "status", func(ctx context.Context, b *bytes.Buffer) error { return r.Status(ctx, b) })
		for _, m := range []string{"00001_baseline.sql", "00002_identity.sql", "00003_tenancy.sql"} {
			if !regexp.MustCompile(`(?m)^migration\t` + regexp.QuoteMeta(m) + `\tapplied\t\S+$`).MatchString(out) {
				t.Errorf("status lacks applied %s: %q", m, out)
			}
		}
		if !strings.HasSuffix(out, "current_version\t3\n") || strings.Contains(out, "pending") {
			t.Fatalf("status: %q", out)
		}
	})

	t.Run("MG-4 reapply keeps the catalog", func(t *testing.T) {
		own := catalogConn(t, cfg)
		before := tenancyCatalog(t, own)
		r := runnerFor(t, cfg, subsetFS(t, 3))
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out !=
			"rolled_back\t00003_tenancy.sql\napplied\t00003_tenancy.sql\ncurrent_version\t3\n" {
			t.Fatalf("reapply: %q", out)
		}
		if after := tenancyCatalog(t, own); after != before {
			t.Fatalf("catalog after reapply differs:\n%s\n---\n%s", before, after)
		}
	})

	var preTenancy, idSnap string
	t.Run("MG-3 populated upgrade", func(t *testing.T) {
		reset(t, cfg)
		r2 := runnerFor(t, cfg, subsetFS(t, 2))
		if out := run(t, "up to 2", func(ctx context.Context, b *bytes.Buffer) error { return r2.Up(ctx, b) }); out !=
			"applied\t00001_baseline.sql\napplied\t00002_identity.sql\ncurrent_version\t2\n" {
			t.Fatalf("up to 2: %q", out)
		}
		_ = r2.Close()
		rt := runtimeConn(t, cfg)
		users := seedIdentity(t, rt, 3, "mg")
		own := catalogConn(t, cfg)
		idSnap = identitySnapshot(t, own)
		preTenancy = tenancyCatalog(t, own)
		r := runnerFor(t, cfg, subsetFS(t, 3))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00003_tenancy.sql\ncurrent_version\t3\n" {
			t.Fatalf("upgrade: %q", out)
		}
		if got := identitySnapshot(t, own); got != idSnap {
			t.Fatalf("identity rows changed by the upgrade:\n%s\n---\n%s", idSnap, got)
		}
		// Runtime CRUD smoke on the new objects, including CreateGroup with a seeded creator.
		svc := tenancy.NewService(runtimePool(t, cfg), nil)
		org, err := svc.CreateOrganization(ctx, "Upgrade Organization")
		if err != nil {
			t.Fatal(err)
		}
		creator, _ := tenancy.ParseUserID(users[0])
		g, owner, err := svc.CreateGroup(ctx, org.ID, creator, "Upgrade Group", "")
		if err != nil {
			t.Fatal(err)
		}
		if g.Slug != "upgrade-group" || owner.Role != string(tenancy.GroupOwner) || owner.SubjectUserID != creator || owner.GroupID == nil || *owner.GroupID != g.ID {
			t.Fatalf("created group (slug %s) or its owner binding (%s %s) differs", g.Slug, owner.Role, owner.ScopeType)
		}
		member, _ := tenancy.ParseUserID(users[1])
		if _, err := svc.GrantOrganizationRole(ctx, org.ID, member, tenancy.OrgMember); err != nil {
			t.Fatal(err)
		}
		if m, err := svc.GetGroupMembership(ctx, org.ID, g.ID, creator); err != nil || m.Role != tenancy.GroupOwner {
			t.Fatalf("owner projection: %+v %v", m, err)
		}
		if got := lines(t, own, `SELECT (SELECT count(*) FROM app.organizations)::text || '/' || (SELECT count(*) FROM app.groups)::text || '/' ||
  (SELECT count(*) FROM app.role_bindings)::text`); got != "1/1/2" {
			t.Fatalf("rows after the smoke: %s", got)
		}
		if got := identitySnapshot(t, own); got != idSnap {
			t.Fatal("identity rows changed by tenancy writes")
		}
	})

	t.Run("MG-5 down to 2", func(t *testing.T) {
		// Down destroys the G005 fixture rows with their tables (documented); identity rows and the catalog of version 2
		// are unchanged.
		r := runnerFor(t, cfg, subsetFS(t, 3))
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out != "rolled_back\t00003_tenancy.sql\ncurrent_version\t2\n" {
			t.Fatalf("down: %q", out)
		}
		own := catalogConn(t, cfg)
		tenancyObjectsAbsent(t, own)
		if got := identitySnapshot(t, own); got != idSnap {
			t.Fatal("identity rows changed by down")
		}
		if got := tenancyCatalog(t, own); got != preTenancy {
			t.Fatalf("catalog after down differs from version 2:\n%s\n---\n%s", preTenancy, got)
		}
		metadataAndIdentityOnly(t, cfg, r)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00003_tenancy.sql\ncurrent_version\t3\n" {
			t.Fatalf("reapply after down: %q", out)
		}
		if got := identitySnapshot(t, own); got != idSnap {
			t.Fatal("identity rows changed by up after down")
		}
	})

	t.Run("ESC-1 identity fixture truncate at version 3", func(t *testing.T) {
		// Both G004 fixture literals are byte-identical and carry CASCADE (Plan r4 ESC-1 option B).
		for _, f := range []string{"../identity/integration_test.go", "../identity/workos/staging_test.go"} {
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			src := string(b)
			if strings.Count(src, "`"+identityFixtureTruncate+"`") != 1 || strings.Count(src, "TRUNCATE ") != 1 {
				t.Fatalf("%s: the identity reset statement differs from %q", f, identityFixtureTruncate)
			}
		}
		rt := runtimeConn(t, cfg)
		own := catalogConn(t, cfg)
		users := seedIdentity(t, rt, 2, "esc")
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
		if _, err := svc.GrantOrganizationRole(ctx, org.ID, u1, tenancy.OrgAdmin); err != nil {
			t.Fatal(err)
		}
		counts := `SELECT (SELECT count(*) FROM app.organizations)::text || '/' || (SELECT count(*) FROM app.groups)::text || '/' ||
  (SELECT count(*) FROM app.role_bindings)::text || '/' || (SELECT count(*) FROM app.users)::text || '/' ||
  (SELECT count(*) FROM app.sessions)::text || '/' || (SELECT count(*) FROM app.identity_principals)::text`
		before := lines(t, own, counts)
		// Without CASCADE the statement is refused even though a binding exists (and would be even without one).
		tx, err := own.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, strings.TrimSuffix(identityFixtureTruncate, " CASCADE"))
		_ = tx.Rollback(ctx)
		if sqlstateOf(err) != "0A000" {
			t.Fatalf("plain TRUNCATE at version 3: %v, want 0A000", err)
		}
		if _, err := own.Exec(ctx, identityFixtureTruncate); err != nil {
			t.Fatalf("fixture TRUNCATE ... CASCADE: %v", err)
		}
		orgs, groups := strings.Split(before, "/")[0], strings.Split(before, "/")[1]
		// Bindings are emptied with their users; Organizations and Groups remain; nothing else is cascaded into.
		if got := lines(t, own, counts); got != orgs+"/"+groups+"/0/0/0/0" {
			t.Fatalf("after the fixture reset: %s (before %s)", got, before)
		}
		if _, err := svc.GetGroup(ctx, org.ID, g.ID); err != nil {
			t.Fatalf("group after the fixture reset: %v", err)
		}
		if ms, err := svc.ListGroupMemberships(ctx, org.ID, g.ID); err != nil || len(ms) != 0 {
			t.Fatalf("group memberships after the fixture reset: %v %v", ms, err)
		}
		metadataIdentityAndTenancyOnly(t, cfg, runnerFor(t, cfg, subsetFS(t, 3)))
	})
}

// TestTenancyPersistSeed (MG-6, before the harness restarts PostgreSQL with its volume kept) seeds tenancy rows at
// version 3 and records their IDs in TENANCY_PERSIST_FILE (0600). TestTenancyPersistVerify reads them back afterwards.
func TestTenancyPersistSeed(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	file := os.Getenv("TENANCY_PERSIST_FILE")
	if file == "" {
		t.Fatal("TENANCY_PERSIST_FILE is required")
	}
	// G064: the seed runs on its own fresh copy of the current template; it establishes version 3 itself (reset to
	// empty, then the bounded Up) instead of relying on TestTenancyMigration's end state.
	reset(t, cfg)
	r := runnerFor(t, cfg, subsetFS(t, 3))
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t3\n") {
		t.Fatalf("up: %q", out)
	}
	users := seedIdentity(t, runtimeConn(t, cfg), 1, "persist")
	svc := tenancy.NewService(runtimePool(t, cfg), nil)
	org, err := svc.CreateOrganization(ctx, "Persisted Organization")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := tenancy.ParseUserID(users[0])
	g, owner, err := svc.CreateGroup(ctx, org.ID, u, "Persisted Group", "persisted-group")
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.GrantOrganizationRole(ctx, org.ID, u, tenancy.OrgOwner)
	if err != nil {
		t.Fatal(err)
	}
	content := strings.Join([]string{org.ID.String(), g.ID.String(), u.String(), owner.ID.String(), b.ID.String()}, "\n") + "\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestTenancyPersistVerify(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	file := os.Getenv("TENANCY_PERSIST_FILE")
	if file == "" {
		t.Fatal("TENANCY_PERSIST_FILE is required")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	f := strings.Fields(string(b))
	if len(f) != 5 {
		t.Fatalf("persist file has %d fields", len(f))
	}
	org, err1 := tenancy.ParseOrganizationID(f[0])
	gid, err2 := tenancy.ParseGroupID(f[1])
	u, err3 := tenancy.ParseUserID(f[2])
	if err := errors.Join(err1, err2, err3); err != nil {
		t.Fatal(err)
	}
	r := runnerFor(t, cfg, subsetFS(t, 3))
	if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t3\n" {
		t.Fatalf("version after restart: %q", out)
	}
	svc := tenancy.NewService(runtimePool(t, cfg), nil)
	if o, err := svc.GetOrganization(ctx, org); err != nil || o.Name != "Persisted Organization" {
		t.Fatalf("organization after restart: %v", err)
	}
	if g, err := svc.GetGroupBySlug(ctx, org, "persisted-group"); err != nil || g.ID != gid {
		t.Fatalf("group after restart: %v", err)
	}
	if m, err := svc.GetGroupMembership(ctx, org, gid, u); err != nil || m.Role != tenancy.GroupOwner {
		t.Fatalf("owner membership after restart: %v", err)
	}
	if m, err := svc.GetOrganizationMembership(ctx, org, u); err != nil || m.Role != tenancy.OrgOwner {
		t.Fatalf("organization membership after restart: %v", err)
	}
	if got := lines(t, catalogConn(t, cfg), `SELECT string_agg(id::text, ',' ORDER BY id) FROM app.role_bindings WHERE organization_id = $1::text::uuid`, f[0]); got !=
		strings.Join(sortedPair(f[3], f[4]), ",") {
		t.Fatalf("binding IDs after restart: %s", got)
	}
}

func sortedPair(a, b string) []string {
	if b < a {
		return []string{b, a}
	}
	return []string{a, b}
}
