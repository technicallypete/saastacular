//go:build integration

// G007 Organization policy catalog and migration tests against a disposable PostgreSQL initialized by the rendered role
// script (infra/validation/run.sh, suite "migrations"): B-POLICY-CATALOG (TestPolicyCatalog) and B-POLICY-MIGRATION
// (TestPolicyMigration, then TestPolicyPersistSeed and TestPolicyPersistVerify around a PostgreSQL restart; G007 Plan r1
// §6.2). They fail, never skip, when the database settings are absent. Runtime cases run as the actual runtime role;
// owner-level cases use the migrator with SET ROLE <db>_owner. Rows are written through runtime-role SQL or the tenancy
// model; the guarded facade is B-POLICY-IT's.
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, APP_ENV; the
// persist phases also POLICY_PERSIST_FILE.
package migrate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

// policyColumn is the one column of 00005_organization_policy.sql.
const policyColumn = "group_admins_edit_business_profiles"

// policyCatalogQueries is the CT-POL catalog: tenancyCatalog's seven queries, with dropped attributes excluded from the
// column ACLs (a down leaves a dropped-attribute slot), plus the column definitions of every app relation.
var policyCatalogQueries = []string{
	catalogPartQueries[0],
	`SELECT c.relname || '.' || a.attname || '|' || a.attacl::text FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE c.relnamespace = 'app'::regnamespace AND a.attacl IS NOT NULL AND NOT a.attisdropped ORDER BY c.relname, a.attname`,
	catalogPartQueries[2], catalogPartQueries[3], catalogPartQueries[4], catalogPartQueries[5], catalogPartQueries[6],
	`SELECT c.relname || '.' || a.attname || ' ' || format_type(a.atttypid, a.atttypmod) || CASE WHEN a.attnotnull THEN ' not null' ELSE '' END
       || coalesce(' default ' || pg_get_expr(d.adbin, d.adrelid), '')
FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE c.relnamespace = 'app'::regnamespace AND a.attnum > 0 AND NOT a.attisdropped ORDER BY c.relname, a.attnum`,
}

var policyCatalogNames = []string{"relations", "column ACLs", "constraints", "indexes", "views", "functions", "triggers", "columns"}

func policyCatalog(t *testing.T, q catalogQuerier) [][]string {
	t.Helper()
	out := make([][]string, len(policyCatalogQueries))
	for i, sql := range policyCatalogQueries {
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

// policyDelta asserts that v5 is v4 plus exactly the policy column and its runtime column ACL, every other row
// unchanged.
func policyDelta(t *testing.T, v4, v5 [][]string, runtime, owner string) {
	t.Helper()
	wantAdded := map[int]string{
		1: "organizations." + policyColumn + "|{" + runtime + "=w/" + owner + "}",
		7: "organizations." + policyColumn + " boolean not null default true",
	}
	for i := range v5 {
		var kept, added []string
		for _, row := range v5[i] {
			if strings.Contains(row, policyColumn) {
				added = append(added, row)
			} else {
				kept = append(kept, row)
			}
		}
		for _, row := range v4[i] {
			if strings.Contains(row, policyColumn) {
				t.Fatalf("a version-4 %s row names the policy column", policyCatalogNames[i])
			}
		}
		if strings.Join(kept, "\n") != strings.Join(v4[i], "\n") {
			t.Errorf("%s: a predecessor row changed:\n%s\n---\n%s", policyCatalogNames[i], strings.Join(v4[i], "\n"), strings.Join(kept, "\n"))
		}
		if want, ok := wantAdded[i]; ok {
			if strings.Join(added, "\n") != want {
				t.Errorf("%s: added %q, want exactly %q", policyCatalogNames[i], added, want)
			}
		} else if len(added) != 0 {
			t.Errorf("%s: unexpected policy rows %q", policyCatalogNames[i], added)
		}
	}
}

// predecessorRows hashes every predecessor row: the Organizations by their version-4 columns, Groups, bindings,
// profiles and associations (owner connection).
func predecessorRows(t *testing.T, q catalogQuerier) string {
	t.Helper()
	return lines(t, q, `SELECT 'organizations:' || coalesce(md5(string_agg((o.id, o.name, o.created_at)::text, E'\n' ORDER BY o.id)), '') FROM app.organizations o
UNION ALL SELECT 'groups:' || coalesce(md5(string_agg(g::text, E'\n' ORDER BY g.id)), '') FROM app.groups g
UNION ALL SELECT 'role_bindings:' || coalesce(md5(string_agg(b::text, E'\n' ORDER BY b.id)), '') FROM app.role_bindings b
UNION ALL SELECT 'business_profiles:' || coalesce(md5(string_agg(p::text, E'\n' ORDER BY p.id)), '') FROM app.business_profiles p
UNION ALL SELECT 'group_business_profiles:' || coalesce(md5(string_agg(a::text, E'\n' ORDER BY a.id)), '') FROM app.group_business_profiles a`)
}

// stmtState runs one statement in its own transaction, consumes its rows, rolls back and returns the SQLSTATE ("" on
// success) and the number of rows returned or affected.
func stmtState(t *testing.T, conn *pgx.Conn, stmt string, args ...any) (string, int64) {
	t.Helper()
	ctx := context.Background()
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, stmt, args...)
	if err != nil {
		return sqlstateOf(err), 0
	}
	var n int64
	for rows.Next() {
		n++
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return sqlstateOf(err), 0
	}
	if tag := rows.CommandTag(); !strings.HasPrefix(tag.String(), "SELECT") {
		n = tag.RowsAffected()
	}
	return "", n
}

func wantState(t *testing.T, what string, gotState string, gotRows int64, state string, rows int64) {
	t.Helper()
	if gotState != state || (state == "" && gotRows != rows) {
		t.Errorf("%s: SQLSTATE %q with %d rows, want %q with %d", what, gotState, gotRows, state, rows)
	}
}

// policyFixture seeds, through runtime-role SQL, one user with a live session, an Organization, a Group, the user's
// Organization and Group bindings, and a profile linked to the Group. It returns their IDs.
type policyFixture struct{ user, org, group, profile, link string }

func seedPolicyFixture(t *testing.T, rt *pgx.Conn, tag string) policyFixture {
	t.Helper()
	ctx := context.Background()
	f := policyFixture{user: seedIdentity(t, rt, 1, tag)[0]}
	if err := rt.QueryRow(ctx, `INSERT INTO app.organizations (name) VALUES ($1) RETURNING id::text`, "Policy "+tag).Scan(&f.org); err != nil {
		t.Fatalf("seed organization: %v", err)
	}
	if err := rt.QueryRow(ctx, `INSERT INTO app.groups (organization_id, name, slug) VALUES ($1::text::uuid, $2, $3) RETURNING id::text`,
		f.org, "Policy Group "+tag, "policy-"+tag).Scan(&f.group); err != nil {
		t.Fatalf("seed group: %v", err)
	}
	// The fixture user owns both scopes: every Group carries a group_owner binding, as the model has always created one
	// (G005 IT-OWN-1) and as migration 00015's legacy owner preflight requires of the rows this fixture leaves in the
	// shared harness database at a lower version (TestPolicyPersistSeed, seedRLSWorld) before a later check migrates up.
	if _, err := rt.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_owner', 'membership'),
       ($1::text::uuid, $2::text::uuid, 'group', $3::text::uuid, $3::text::uuid, 'group_owner', 'membership')`, f.org, f.user, f.group); err != nil {
		t.Fatalf("seed bindings: %v", err)
	}
	f.profile, f.link = insertProfileAndLink(t, rt, f.org, f.group, "Policy Profile "+tag)
	return f
}

// --- B-POLICY-CATALOG ----------------------------------------------------------------------------------------------

func TestPolicyCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r4 := runnerFor(t, cfg, subsetFS(t, 4))
	if out := run(t, "up to 4", func(ctx context.Context, b *bytes.Buffer) error { return r4.Up(ctx, b) }); out !=
		"applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\ncurrent_version\t4\n" {
		t.Fatalf("up to 4: %q", out)
	}
	_ = r4.Close()
	own := catalogConn(t, cfg)
	v4 := policyCatalog(t, own)
	filenode := lines(t, own, `SELECT pg_relation_filenode('app.organizations'::regclass)::text`)
	r := runnerFor(t, cfg, subsetFS(t, 5))
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out !=
		"applied\t00005_organization_policy.sql\ncurrent_version\t5\n" {
		t.Fatalf("up to 5: %q", out)
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

	t.Run("CT-POL-1 the version-5 delta is one column and one column ACL", func(t *testing.T) {
		metadataIdentityTenancyAndBusinessProfileOnly(t, cfg, r)
		policyDelta(t, v4, policyCatalog(t, own), rtr, o)
		// A catalog-only fast default: no table rewrite, and the stored rows carry the default as a missing value.
		if got := lines(t, own, `SELECT pg_relation_filenode('app.organizations'::regclass)::text`); got != filenode {
			t.Errorf("app.organizations was rewritten (filenode %s -> %s)", filenode, got)
		}
		if got := lines(t, own, `SELECT a.atthasmissing::text || ' ' || a.attnum::text FROM pg_attribute a
WHERE a.attrelid = 'app.organizations'::regclass AND a.attname = $1`, policyColumn); got != "true 4" {
			t.Errorf("policy column attribute: %q, want a fast default at attnum 4", got)
		}
	})

	t.Run("CT-POL-2 exact runtime privilege matrix", func(t *testing.T) {
		if got := lines(t, own, `SELECT c.relacl::text FROM pg_class c WHERE c.oid = 'app.organizations'::regclass`); got !=
			fmt.Sprintf("{%s=arwdDxtm/%s,%s=r/%s}", o, o, rtr, o) {
			t.Errorf("organizations table ACL: %s", got)
		}
		if got := lines(t, own, `SELECT a.attname || '|' || a.attacl::text FROM pg_attribute a
WHERE a.attrelid = 'app.organizations'::regclass AND a.attacl IS NOT NULL AND NOT a.attisdropped ORDER BY a.attname`); got !=
			fmt.Sprintf("%s|{%s=w/%s}\nname|{%s=aw/%s}", policyColumn, rtr, o, rtr, o) {
			t.Errorf("organizations column ACLs:\n%s", got)
		}
		matrix := `SELECT coalesce(string_agg(p, ',' ORDER BY p), '') FROM unnest(ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'TRUNCATE',
  'REFERENCES', 'TRIGGER', 'MAINTAIN']) p WHERE has_table_privilege($1::text::name, 'app.organizations'::regclass, p)`
		cols := `SELECT coalesce(string_agg(a.attname || ':' || p, ',' ORDER BY a.attname, p), '')
FROM pg_attribute a, unnest(ARRAY['INSERT', 'UPDATE', 'REFERENCES']) p
WHERE a.attrelid = 'app.organizations'::regclass AND a.attnum > 0 AND NOT a.attisdropped
  AND has_column_privilege($1::text::name, a.attrelid, a.attnum, p)`
		if got := lines(t, own, matrix, rtr); got != "SELECT" {
			t.Errorf("runtime table privileges: %s", got)
		}
		if got := lines(t, own, cols, rtr); got != policyColumn+":UPDATE,name:INSERT,name:UPDATE" {
			t.Errorf("runtime column privileges: %s", got)
		}
		for _, role := range []string{cfg.RetentionRole(), cfg.MigratorRole()} {
			if got := lines(t, own, matrix, role) + lines(t, own, cols, role); got != "" {
				t.Errorf("%s holds %s", role, got)
			}
		}
		if got := lines(t, own, `SELECT (SELECT count(*) FROM pg_class c, aclexplode(c.relacl) a WHERE c.oid = 'app.organizations'::regclass AND a.grantee = 0)
  + (SELECT count(*) FROM pg_attribute at, aclexplode(at.attacl) a WHERE at.attrelid = 'app.organizations'::regclass AND a.grantee = 0)`); got != "0" {
			t.Errorf("PUBLIC entries on organizations: %s", got)
		}
		// A creator cannot pre-set the flag; a new Organization receives the configured default.
		st, n := stmtState(t, rt, `INSERT INTO app.organizations (name, `+policyColumn+`) VALUES ('Preset', false) RETURNING id`)
		wantState(t, "runtime INSERT of the flag", st, n, "42501", 0)
		var flag bool
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		err = tx.QueryRow(ctx, `INSERT INTO app.organizations (name) VALUES ('Default Flag') RETURNING `+policyColumn).Scan(&flag)
		_ = tx.Rollback(ctx)
		if err != nil || !flag {
			t.Errorf("a new Organization's flag: %v %v, want true", flag, err)
		}
	})

	f := seedPolicyFixture(t, rt, "cat")

	t.Run("CT-POL-3 malformed values and row-lock privileges", func(t *testing.T) {
		row := func() string {
			return lines(t, own, `SELECT o::text FROM app.organizations o WHERE o.id = $1::text::uuid`, f.org)
		}
		before := row()
		if !strings.HasSuffix(before, ",t)") {
			t.Fatal("the seeded Organization row does not end with the true flag")
		}
		for _, c := range []struct {
			what  string
			conn  *pgx.Conn
			value string
			state string
		}{
			{"runtime NULL", rt, "NULL", "23502"}, {"runtime text", rt, "'maybe'", "22P02"},
			{"owner NULL", own, "NULL", "23502"}, {"owner text", own, "'maybe'", "22P02"},
		} {
			st, n := stmtState(t, c.conn, `UPDATE app.organizations SET `+policyColumn+` = `+c.value+` WHERE id = $1::text::uuid`, f.org)
			wantState(t, c.what+" assignment", st, n, c.state, 0)
		}
		st, n := stmtState(t, rt, `UPDATE app.organizations SET `+policyColumn+` = false WHERE id = $1::text::uuid`, f.org)
		wantState(t, "runtime boolean assignment (rolled back)", st, n, "", 1)
		if row() != before {
			t.Fatal("the Organization row changed")
		}
		// FOR SHARE needs UPDATE on at least one column: runtime holds none on group_business_profiles, which is why the
		// guarded writes and Unlink use the Group row as the association's lock surrogate (G007 Plan r1 §2, §4).
		st, n = stmtState(t, rt, `SELECT 1 FROM app.group_business_profiles WHERE organization_id = $1::text::uuid FOR SHARE`, f.org)
		wantState(t, "runtime FOR SHARE on group_business_profiles", st, n, "42501", 0)
		for _, c := range []struct{ what, stmt, arg string }{
			{"users FOR SHARE", `SELECT 1 FROM app.users WHERE id = $1::text::uuid AND status = 'active' FOR SHARE`, f.user},
			{"sessions FOR SHARE", `SELECT 1 FROM app.sessions WHERE user_id = $1::text::uuid AND revoked_at IS NULL FOR SHARE`, f.user},
			{"organizations FOR SHARE", `SELECT 1 FROM app.organizations WHERE id = $1::text::uuid FOR SHARE`, f.org},
			{"organizations FOR NO KEY UPDATE", `SELECT 1 FROM app.organizations WHERE id = $1::text::uuid FOR NO KEY UPDATE`, f.org},
			{"groups FOR SHARE", `SELECT 1 FROM app.groups WHERE organization_id = $1::text::uuid FOR SHARE`, f.org},
			{"groups FOR NO KEY UPDATE", `SELECT 1 FROM app.groups WHERE organization_id = $1::text::uuid FOR NO KEY UPDATE`, f.org},
		} {
			st, n := stmtState(t, rt, c.stmt, c.arg)
			wantState(t, "runtime "+c.what, st, n, "", 1)
		}
		// A row lock through a SELECT-only projection is refused, so the membership-fact locks name role_bindings with the
		// projections' filters (Plan r1 §11 fallback; Codex review 432d5f17 §1): 42501 through both projections, success
		// on the direct scoped locks, and the projections' write and ACL restrictions unchanged.
		for _, view := range []string{"organization_memberships", "group_memberships"} {
			st, n := stmtState(t, rt, `SELECT 1 FROM app.`+view+` WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid FOR SHARE`, f.org, f.user)
			wantState(t, "runtime FOR SHARE through "+view, st, n, "42501", 0)
		}
		for what, stmt := range map[string]string{
			"Organization binding lock": `SELECT organization_id, role FROM app.role_bindings WHERE organization_id = $1::text::uuid
AND subject_user_id = $2::text::uuid AND scope_type = 'organization' AND source = 'membership' FOR SHARE`,
			"all Group bindings lock": `SELECT organization_id, group_id, role FROM app.role_bindings WHERE organization_id = $1::text::uuid
AND subject_user_id = $2::text::uuid AND scope_type = 'group' AND source = 'membership' ORDER BY group_id FOR SHARE`,
		} {
			st, n := stmtState(t, rt, stmt, f.org, f.user)
			wantState(t, "runtime "+what, st, n, "", 1)
		}
		st, n = stmtState(t, rt, `SELECT organization_id, group_id, role FROM app.role_bindings WHERE organization_id = $1::text::uuid
AND subject_user_id = $2::text::uuid AND scope_type = 'group' AND source = 'membership' AND group_id = ANY (ARRAY[$3::text::uuid]) ORDER BY group_id FOR SHARE`,
			f.org, f.user, f.group)
		wantState(t, "runtime candidate Group bindings lock", st, n, "", 1)
		for what, stmt := range map[string]string{
			"INSERT through organization_memberships": `INSERT INTO app.organization_memberships (user_id, organization_id, role) VALUES ($2::text::uuid, $1::text::uuid, 'org_owner')`,
			"UPDATE through organization_memberships": `UPDATE app.organization_memberships SET role = 'org_owner' WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid`,
			"DELETE through group_memberships":        `DELETE FROM app.group_memberships WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid`,
			"UPDATE through group_memberships":        `UPDATE app.group_memberships SET role = 'group_owner' WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid`,
		} {
			st, n := stmtState(t, rt, stmt, f.org, f.user)
			wantState(t, "runtime "+what, st, n, "42501", 0)
		}
		if got := lines(t, own, `SELECT c.relname || '|' || c.relacl::text FROM pg_class c WHERE c.oid IN ('app.organization_memberships'::regclass,
  'app.group_memberships'::regclass, 'app.role_bindings'::regclass) ORDER BY c.relname`); got != fmt.Sprintf(
			"group_memberships|{%s=arwdDxtm/%s,%s=r/%s}\norganization_memberships|{%s=arwdDxtm/%s,%s=r/%s}\nrole_bindings|{%s=arwdDxtm/%s,%s=rd/%s}",
			o, o, rtr, o, o, o, rtr, o, o, o, rtr, o) {
			t.Errorf("projection and role_bindings table ACLs:\n%s", got)
		}
		col := fmt.Sprintf("{%s=a/%s}", rtr, o)
		if got := lines(t, own, `SELECT c.relname || '.' || a.attname || '|' || a.attacl::text FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE c.oid IN ('app.organization_memberships'::regclass, 'app.group_memberships'::regclass, 'app.role_bindings'::regclass)
  AND a.attacl IS NOT NULL AND NOT a.attisdropped ORDER BY c.relname, a.attname`); got != strings.Join([]string{
			"role_bindings.group_id|" + col, "role_bindings.organization_id|" + col, "role_bindings.role|" + fmt.Sprintf("{%s=aw/%s}", rtr, o),
			"role_bindings.scope_id|" + col, "role_bindings.scope_type|" + col, "role_bindings.source|" + col, "role_bindings.subject_user_id|" + col,
		}, "\n") {
			t.Errorf("projection and role_bindings column ACLs:\n%s", got)
		}
	})

	t.Run("CT-POL-4 no row-level security, new object or authority", func(t *testing.T) {
		// G008's obligation (docs/foundation/AUTHORIZATION.md); this is not isolation evidence.
		if got := lines(t, own, `SELECT coalesce(string_agg(relname, ','), '') FROM pg_class
WHERE relnamespace = 'app'::regnamespace AND (relrowsecurity OR relforcerowsecurity)`); got != "" {
			t.Fatalf("row security on %s", got)
		}
		if got := lines(t, own, `SELECT count(*)::text FROM pg_policy`); got != "0" {
			t.Fatalf("policies: %s", got)
		}
		if got := lines(t, own, `SELECT count(*)::text FROM pg_proc WHERE pronamespace = 'app'::regnamespace AND prosecdef`); got != "0" {
			t.Fatalf("SECURITY DEFINER functions: %s", got)
		}
		// The single-authority detector keeps the version-4 inventory: the flag is no role, mirror or grant store.
		if v := authorityViolations(t, own, cfg, businessProfileAuthority); len(v) != 0 {
			t.Fatalf("authority violations at version 5: %v", v)
		}
	})

	t.Run("CT-POL-5 runtime DDL, role and metadata denials", func(t *testing.T) {
		for what, stmt := range map[string]string{
			"DROP the policy column":       `ALTER TABLE app.organizations DROP COLUMN ` + policyColumn,
			"change the policy default":    `ALTER TABLE app.organizations ALTER COLUMN ` + policyColumn + ` SET DEFAULT false`,
			"drop the policy NOT NULL":     `ALTER TABLE app.organizations ALTER COLUMN ` + policyColumn + ` DROP NOT NULL`,
			"add an Organization column":   `ALTER TABLE app.organizations ADD COLUMN policy jsonb`,
			"CREATE TABLE":                 `CREATE TABLE app.organization_policies (organization_id uuid, key text, value text)`,
			"ENABLE ROW LEVEL SECURITY":    `ALTER TABLE app.organizations ENABLE ROW LEVEL SECURITY`,
			"CREATE POLICY":                `CREATE POLICY p ON app.organizations USING (true)`,
			"COMMENT ON the policy column": `COMMENT ON COLUMN app.organizations.` + policyColumn + ` IS 'x'`,
			"TRUNCATE organizations":       `TRUNCATE app.organizations`,
			"DELETE organizations":         `DELETE FROM app.organizations`,
			"SET ROLE owner":               `SET ROLE ` + quote(cfg.OwnerRole()),
			"SET ROLE migrator":            `SET ROLE ` + quote(cfg.MigratorRole()),
			"metadata read":                `SELECT 1 FROM app.goose_db_version`,
		} {
			expectDenied(t, rt, "runtime "+what, stmt)
		}
	})
}

// --- B-POLICY-MIGRATION --------------------------------------------------------------------------------------------

func TestPolicyMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	const full = "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\n" +
		"applied\t00005_organization_policy.sql\ncurrent_version\t5\n"

	t.Run("MG-POL-1 empty to version 5", func(t *testing.T) {
		reset(t, cfg)
		r := runnerFor(t, cfg, subsetFS(t, 5))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != full {
			t.Fatalf("up: %q", out)
		}
		metadataIdentityTenancyAndBusinessProfileOnly(t, cfg, r)
		if got := lines(t, catalogConn(t, cfg), `SELECT format_type(atttypid, atttypmod) || ' ' || attnotnull::text FROM pg_attribute
WHERE attrelid = 'app.organizations'::regclass AND attname = $1 AND NOT attisdropped`, policyColumn); got != "boolean true" {
			t.Fatalf("policy column: %q", got)
		}
	})

	t.Run("MG-POL-2 repeated up and status", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 5))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "no pending migrations\ncurrent_version\t5\n" {
			t.Fatalf("repeated up: %q", out)
		}
		out := run(t, "status", func(ctx context.Context, b *bytes.Buffer) error { return r.Status(ctx, b) })
		for _, m := range []string{"00001_baseline.sql", "00002_identity.sql", "00003_tenancy.sql", "00004_business_profiles.sql", "00005_organization_policy.sql"} {
			if !regexp.MustCompile(`(?m)^migration\t` + regexp.QuoteMeta(m) + `\tapplied\t\S+$`).MatchString(out) {
				t.Errorf("status lacks applied %s: %q", m, out)
			}
		}
		if !strings.HasSuffix(out, "current_version\t5\n") || strings.Contains(out, "pending") {
			t.Fatalf("status: %q", out)
		}
	})

	var v4 [][]string
	var idSnap, rowSnap string
	var f policyFixture
	t.Run("MG-POL-3 populated version-4 upgrade", func(t *testing.T) {
		reset(t, cfg)
		r4 := runnerFor(t, cfg, subsetFS(t, 4))
		if out := run(t, "up to 4", func(ctx context.Context, b *bytes.Buffer) error { return r4.Up(ctx, b) }); out !=
			"applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\ncurrent_version\t4\n" {
			t.Fatalf("up to 4: %q", out)
		}
		_ = r4.Close()
		rt := runtimeConn(t, cfg)
		f = seedPolicyFixture(t, rt, "upgrade")
		seedPolicyFixture(t, rt, "second")
		own := catalogConn(t, cfg)
		v4, idSnap, rowSnap = policyCatalog(t, own), identitySnapshot(t, own), predecessorRows(t, own)
		r := runnerFor(t, cfg, subsetFS(t, 5))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out !=
			"applied\t00005_organization_policy.sql\ncurrent_version\t5\n" {
			t.Fatalf("upgrade: %q", out)
		}
		if identitySnapshot(t, own) != idSnap || predecessorRows(t, own) != rowSnap {
			t.Fatal("identity or predecessor rows changed by the upgrade")
		}
		if got := lines(t, own, `SELECT count(*)::text || '/' || count(*) FILTER (WHERE `+policyColumn+`)::text FROM app.organizations`); got != "2/2" {
			t.Fatalf("existing Organizations and their true flags: %s", got)
		}
		policyDelta(t, v4, policyCatalog(t, own), cfg.RuntimeRole(), cfg.OwnerRole())
		// The tenancy model keeps working at version 5, and a new Organization reads true.
		svc := tenancy.NewService(runtimePool(t, cfg), nil)
		org, err := svc.CreateOrganization(ctx, "Upgrade Organization")
		if err != nil {
			t.Fatal(err)
		}
		if got := lines(t, own, `SELECT `+policyColumn+`::text FROM app.organizations WHERE id = $1::text::uuid`, org.ID.String()); got != "true" {
			t.Fatalf("new Organization flag: %s", got)
		}
		if _, err := own.Exec(ctx, `DELETE FROM app.organizations WHERE id = $1::text::uuid`, org.ID.String()); err != nil {
			t.Fatal(err)
		}
		if predecessorRows(t, own) != rowSnap {
			t.Fatal("predecessor rows differ after removing the new Organization")
		}
	})

	t.Run("MG-POL-4 reapply keeps the catalog", func(t *testing.T) {
		own := catalogConn(t, cfg)
		before := policyCatalog(t, own)
		r := runnerFor(t, cfg, subsetFS(t, 5))
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out !=
			"rolled_back\t00005_organization_policy.sql\napplied\t00005_organization_policy.sql\ncurrent_version\t5\n" {
			t.Fatalf("reapply: %q", out)
		}
		after := policyCatalog(t, own)
		for i := range before {
			if strings.Join(before[i], "\n") != strings.Join(after[i], "\n") {
				t.Errorf("%s after reapply differ:\n%s\n---\n%s", policyCatalogNames[i], strings.Join(before[i], "\n"), strings.Join(after[i], "\n"))
			}
		}
		if identitySnapshot(t, own) != idSnap || predecessorRows(t, own) != rowSnap {
			t.Fatal("identity or predecessor rows changed by reapply")
		}
	})

	t.Run("MG-POL-5 down to 4 with flags set", func(t *testing.T) {
		// Down destroys only the flag values (documented); identity rows, every predecessor row and the version-4 catalog
		// are unchanged.
		rt := runtimeConn(t, cfg)
		if tag, err := rt.Exec(ctx, `UPDATE app.organizations SET `+policyColumn+` = false WHERE id = $1::text::uuid`, f.org); err != nil || tag.RowsAffected() != 1 {
			t.Fatalf("set the flag false: %v", err)
		}
		r := runnerFor(t, cfg, subsetFS(t, 5))
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out !=
			"rolled_back\t00005_organization_policy.sql\ncurrent_version\t4\n" {
			t.Fatalf("down: %q", out)
		}
		own := catalogConn(t, cfg)
		if identitySnapshot(t, own) != idSnap || predecessorRows(t, own) != rowSnap {
			t.Fatal("identity or predecessor rows changed by down")
		}
		after := policyCatalog(t, own)
		for i := range v4 {
			if strings.Join(v4[i], "\n") != strings.Join(after[i], "\n") {
				t.Errorf("%s after down differ from version 4:\n%s\n---\n%s", policyCatalogNames[i], strings.Join(v4[i], "\n"), strings.Join(after[i], "\n"))
			}
		}
		metadataIdentityTenancyAndBusinessProfileOnly(t, cfg, r)
		// The version-4 tenancy model works on the downgraded database.
		svc := tenancy.NewService(runtimePool(t, cfg), nil)
		org, err := tenancy.ParseOrganizationID(f.org)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GetOrganization(ctx, org); err != nil {
			t.Fatalf("model read at version 4: %v", err)
		}
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out !=
			"applied\t00005_organization_policy.sql\ncurrent_version\t5\n" {
			t.Fatalf("up after down: %q", out)
		}
		if got := lines(t, own, `SELECT count(*) FILTER (WHERE NOT `+policyColumn+`)::text FROM app.organizations`); got != "0" {
			t.Fatalf("flags after down and up: %s false", got)
		}
		if identitySnapshot(t, own) != idSnap || predecessorRows(t, own) != rowSnap {
			t.Fatal("identity or predecessor rows changed by up after down")
		}
	})
}

// TestPolicyPersistSeed (MG-POL-6, before the harness restarts PostgreSQL with its volume kept) seeds two
// Organizations at version 5, one with the flag false and one true, and records their IDs in POLICY_PERSIST_FILE (0600).
// TestPolicyPersistVerify reads them back afterwards.
func TestPolicyPersistSeed(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	file := os.Getenv("POLICY_PERSIST_FILE")
	if file == "" {
		t.Fatal("POLICY_PERSIST_FILE is required")
	}
	// G064: the seed runs on its own fresh copy of the current template; it establishes version 5 itself (reset to
	// empty, then the bounded Up) instead of relying on TestPolicyMigration's end state.
	reset(t, cfg)
	r := runnerFor(t, cfg, subsetFS(t, 5))
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t5\n") {
		t.Fatalf("up: %q", out)
	}
	rt := runtimeConn(t, cfg)
	off := seedPolicyFixture(t, rt, "persist-off")
	on := seedPolicyFixture(t, rt, "persist-on")
	if tag, err := rt.Exec(ctx, `UPDATE app.organizations SET `+policyColumn+` = false WHERE id = $1::text::uuid`, off.org); err != nil || tag.RowsAffected() != 1 {
		t.Fatalf("set the flag false: %v", err)
	}
	if err := os.WriteFile(file, []byte(off.org+"\n"+on.org+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPolicyPersistVerify(t *testing.T) {
	cfg := itConfig(t)
	file := os.Getenv("POLICY_PERSIST_FILE")
	if file == "" {
		t.Fatal("POLICY_PERSIST_FILE is required")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	f := strings.Fields(string(b))
	if len(f) != 2 {
		t.Fatalf("persist file has %d fields", len(f))
	}
	_, err1 := tenancy.ParseOrganizationID(f[0])
	_, err2 := tenancy.ParseOrganizationID(f[1])
	if err := errors.Join(err1, err2); err != nil {
		t.Fatal(err)
	}
	r := runnerFor(t, cfg, subsetFS(t, 5))
	if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t5\n" {
		t.Fatalf("version after restart: %q", out)
	}
	if got := lines(t, runtimeConn(t, cfg), `SELECT string_agg(`+policyColumn+`::text, ',' ORDER BY array_position($1::text[], id::text))
FROM app.organizations WHERE id::text = ANY ($1::text[])`, f); got != "false,true" {
		t.Fatalf("flags after restart: %q, want false,true", got)
	}
}
