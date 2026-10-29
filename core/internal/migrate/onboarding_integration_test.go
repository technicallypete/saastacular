//go:build integration

// G018 B-ONBOARDING-CATALOG (Plan r5 §4, §6.2; R0 addendum; audit-operation amendment
// management/v0/results/G018-audit-operation-amendment-codex-review.md): TestOnboardingCatalog.
//   - CT-ONB-1 the version-14 delta: column-limited runtime INSERT grants (no DELETE/TRIGGER, no provenance column),
//     the policy reader's columns, the six policies, the bootstrap trigger and the boolean reader (owner, SECURITY
//     DEFINER, volatility, search_path, ACL, body), the insertion-provenance columns (type, NOT NULL, defaults, backfill)
//     and the replaced immutability body with its owner refusal, the pg_control_system()/pg_current_xact_id() privilege
//     pins, and the runtime behaviour under actual grants (the member-safe keyed Group insert and its trigger-written
//     owner binding, the depth-0 binding refusal, the insertion predicate, the reader without context).
//   - CT-ONB-AUD the widened events_operation_valid (29 + 7 = 36 operations) and the refusal of an unknown operation.
//   - CT-ONB-UPDOWN a clean down to the exact version-13 catalog and up again; the down refused, without losing history,
//     while audit.events holds a G018 operation.
//   - CT-ONB-1/R-ONB-1-neg and -pos the logical-restore collision boundary (a restored row whose created_xid equals the
//     evaluating transaction's id under a foreign cluster identifier is never eligible; a fresh insert is).
package migrate

import (
	"bytes"
	"context"
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/migrations"
)

const (
	onbOrg     = "0f18c000-0000-4000-8000-000000000001"
	onbOwner   = "0f18c000-0000-4000-8000-000000000002" // org_owner
	onbMember  = "0f18c000-0000-4000-8000-000000000003" // org_member, group_owner of onbGroup
	onbGOnly   = "0f18c000-0000-4000-8000-000000000004" // group_admin of onbGroup, no Organization binding
	onbGroup   = "0f18c000-0000-4000-8000-000000000005"
	onbProfile = "0f18c000-0000-4000-8000-000000000006" // committed before 00014 (backfilled), unattached
	onbOrgB    = "0f18c000-0000-4000-8000-000000000007"
)

// onbOps are the seven G018 access.denied operations (amendment).
var onbOps = []string{"group.suggest_slug", "group.create", "group.list", "group.onboarding.get",
	"group.onboarding.complete_new", "group.onboarding.complete_existing", "business_profile.list"}

// onbBody returns the dollar-quoted body (prosrc) of the first function created after marker in the 00014 up section.
func onbBody(t *testing.T, marker string) string {
	t.Helper()
	b, err := fs.ReadFile(migrations.FS, "00014_group_onboarding.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, _ := strings.Cut(string(b), "-- +goose Down")
	i := strings.Index(up, marker)
	if i < 0 {
		t.Fatalf("00014: %q not found", marker)
	}
	j := strings.Index(up[i:], "AS $$")
	if j < 0 {
		t.Fatalf("00014: no body after %q", marker)
	}
	start := i + j + len("AS $$")
	k := strings.Index(up[start:], "$$")
	return up[start : start+k]
}

func TestOnboardingCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r13 := runnerFor(t, cfg, subsetFS(t, 13))
	objRunUp(t, r13, "applied\t00013_organization_branding.sql\ncurrent_version\t13\n")
	s13 := snapshot(t, r13)
	setup := rlsSetupConn(t, cfg)
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := setup.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("fixture: %v", err)
		}
	}
	// Populated upgrade: the fixture (and one profile to backfill) exists at version 13.
	exec(`INSERT INTO app.users (id, display_name) VALUES ('` + onbOwner + `', 'Onb Owner'), ('` + onbMember + `', 'Onb Member'),
  ('` + onbGOnly + `', 'Onb Group Only')`)
	exec(`INSERT INTO app.organizations (id, name) VALUES ('` + onbOrg + `', 'Onboarding'), ('` + onbOrgB + `', 'Onboarding B')`)
	exec(`INSERT INTO app.groups (id, organization_id, name, slug) VALUES ('` + onbGroup + `', '` + onbOrg + `', 'Existing', 'existing')`)
	exec(`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + onbOrg + `', '` + onbOwner + `', 'organization', '` + onbOrg + `', NULL, 'org_owner', 'membership'),
  ('` + onbOrg + `', '` + onbMember + `', 'organization', '` + onbOrg + `', NULL, 'org_member', 'membership'),
  ('` + onbOrg + `', '` + onbMember + `', 'group', '` + onbGroup + `', '` + onbGroup + `', 'group_owner', 'membership'),
  ('` + onbOrg + `', '` + onbGOnly + `', 'group', '` + onbGroup + `', '` + onbGroup + `', 'group_admin', 'membership')`)
	exec(`INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ('` + onbProfile + `', '` + onbOrg + `', 'Backfilled')`)

	r := runnerFor(t, cfg, subsetFS(t, 14))
	objRunUp(t, r, "applied\t00014_group_onboarding.sql\ncurrent_version\t14\n")
	s14 := snapshot(t, r)
	cat := catalogConn(t, cfg)
	rt := rlsRuntimeConn(t, cfg)
	o, rtr, pr := cfg.OwnerRole(), cfg.RuntimeRole(), cfg.Database+"_policy_reader"
	one := func(q catalogQuerier, sql string, args ...any) string {
		t.Helper()
		var s *string
		if err := q.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		if s == nil {
			return ""
		}
		return *s
	}
	// inTx runs f in a runtime transaction with the given context ("" = absent) and rolls it back.
	inTx := func(user, org, group string, f func(tx pgx.Tx)) {
		t.Helper()
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if user != "" {
			rlsContext(t, tx, user, org, group, "")
		}
		f(tx)
	}
	eligible := func(tx pgx.Tx, id string) bool {
		t.Helper()
		var b bool
		if err := tx.QueryRow(ctx, `SELECT app.rls_profile_eligible($1)`, id).Scan(&b); err != nil {
			t.Fatalf("reader: %v", err)
		}
		return b
	}

	t.Run("CT-ONB-1 the version-14 delta", func(t *testing.T) {
		metadataThroughVersion14Only(t, cfg, r)
		// Column-limited runtime INSERT; the policy reader's columns; no table-level INSERT/DELETE/TRIGGER/TRUNCATE.
		cols := func(grantee, priv string) string {
			return one(cat, `SELECT string_agg(c.relname || '.' || at.attname, ',' ORDER BY c.relname, at.attnum)
FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
CROSS JOIN LATERAL aclexplode(at.attacl) a
WHERE c.oid IN ('app.groups'::regclass, 'app.role_bindings'::regclass, 'app.business_profiles'::regclass,
  'app.group_business_profiles'::regclass) AND a.grantee = $1::regrole AND a.privilege_type = $2`, grantee, priv)
		}
		if got, want := cols(rtr, "INSERT"), "business_profiles.id,business_profiles.organization_id,business_profiles.display_name,"+
			"business_profiles.legal_name,business_profiles.description,business_profiles.industry,business_profiles.website,"+
			"group_business_profiles.organization_id,group_business_profiles.group_id,group_business_profiles.business_profile_id,"+
			"groups.id,groups.organization_id,groups.name,groups.slug,"+
			"role_bindings.organization_id,role_bindings.subject_user_id,role_bindings.scope_type,role_bindings.scope_id,"+
			"role_bindings.group_id,role_bindings.role,role_bindings.source"; got != want {
			t.Errorf("runtime INSERT columns:\n got %s\nwant %s", got, want)
		}
		if got, want := cols(rtr, "UPDATE"), "business_profiles.display_name,business_profiles.legal_name,business_profiles.description,"+
			"business_profiles.industry,business_profiles.website,groups.name,groups.slug,role_bindings.role"; got != want {
			t.Errorf("runtime UPDATE columns:\n got %s\nwant %s", got, want)
		}
		if got := one(cat, `SELECT string_agg(c.relname || '.' || at.attname, ',' ORDER BY c.relname, at.attnum)
FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
CROSS JOIN LATERAL aclexplode(at.attacl) a
WHERE c.oid IN ('app.business_profiles'::regclass, 'app.group_business_profiles'::regclass) AND a.grantee = $1::regrole`, pr); got !=
			"business_profiles.id,business_profiles.organization_id,business_profiles.created_xid,business_profiles.created_sysid,"+
				"group_business_profiles.organization_id,group_business_profiles.group_id,group_business_profiles.business_profile_id" {
			t.Errorf("policy reader columns: %s", got)
		}
		if got := one(cat, `SELECT coalesce(string_agg(c.relname || ':' || a.grantee::regrole::text || ':' || a.privilege_type, ','), '')
FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
WHERE c.oid IN ('app.groups'::regclass, 'app.role_bindings'::regclass, 'app.business_profiles'::regclass,
  'app.group_business_profiles'::regclass) AND a.grantee <> c.relowner AND a.privilege_type <> 'SELECT'`); got != "" {
			t.Errorf("table-level non-SELECT privileges: %s", got)
		}
		for _, c := range []string{"created_xid", "created_sysid", "created_at"} {
			for _, p := range []string{"INSERT", "UPDATE"} {
				var has bool
				if err := cat.QueryRow(ctx, `SELECT has_column_privilege($1, 'app.business_profiles', $2, $3)`, rtr, c, p).Scan(&has); err != nil || has {
					t.Errorf("runtime %s on business_profiles.%s: %v %v", p, c, has, err)
				}
			}
		}
		// The two built-ins are executable by the runtime and policy-reader roles (PostgreSQL's default PUBLIC grant).
		if got := one(cat, `SELECT string_agg(has_function_privilege(r, f, 'EXECUTE')::text, ',' ORDER BY r, f)
FROM unnest($1::text[]) r CROSS JOIN unnest(ARRAY['pg_catalog.pg_control_system()', 'pg_catalog.pg_current_xact_id()']) f`,
			[]string{rtr, pr}); got != "true,true,true,true" {
			t.Errorf("pg_control_system()/pg_current_xact_id() privileges: %s", got)
		}
		// The six policies: name, command, roles and the load-bearing predicates.
		pols := one(cat, `SELECT string_agg(p.polname || ':' || p.polcmd::text || ':' ||
  array_to_string(ARRAY(SELECT x::regrole::text FROM unnest(p.polroles) x ORDER BY 1), '+') || ':' ||
  regexp_replace(coalesce(pg_get_expr(p.polqual, p.polrelid), ''), '\s+', ' ', 'g') || ':' ||
  regexp_replace(coalesce(pg_get_expr(p.polwithcheck, p.polrelid), ''), '\s+', ' ', 'g'), E'\n' ORDER BY p.polname)
FROM pg_policy p WHERE p.polname IN ('groups_insert', 'role_bindings_insert_bootstrap', 'business_profiles_insert',
  'group_business_profiles_insert', 'business_profiles_policy_reader', 'group_business_profiles_policy_reader')`)
		roles := strings.Join(sortedPair(o, rtr), "+")
		want := map[string][]string{
			"business_profiles_insert":              {":a:" + roles + ":", "rls_org_role()", "'group_owner'", "'group_admin'", "ctx_group()"},
			"business_profiles_policy_reader":       {":r:" + pr + ":", "ctx_org()"},
			"group_business_profiles_insert":        {":a:" + roles + ":", "rls_group_role(group_id)", "rls_profile_eligible(business_profile_id)", "ctx_group()"},
			"group_business_profiles_policy_reader": {":r:" + pr + ":", "ctx_org()"},
			"groups_insert":                         {":a:" + roles + ":", "ctx_group()", "IS NULL", "rls_org_role()", "IS NOT NULL"},
			"role_bindings_insert_bootstrap":        {":a:" + roles + ":", "pg_trigger_depth() > 0", "'group_owner'", "'membership'", "ctx_user()"},
		}
		lines := strings.Split(pols, "\n")
		if len(lines) != len(want) {
			t.Fatalf("policies: %q", pols)
		}
		for _, l := range lines {
			name, _, _ := strings.Cut(l, ":")
			for _, frag := range want[name] {
				if !strings.Contains(l, frag) {
					t.Errorf("policy %s lacks %q: %s", name, frag, l)
				}
			}
		}
		// The two new functions and the replaced immutability body.
		sp := "search_path=pg_catalog, pg_temp"
		if got, want := one(cat, `SELECT string_agg(p.proname || ':' || p.proowner::regrole::text || ':' || p.prosecdef::text || ':' ||
  p.provolatile::text || ':' || coalesce(array_to_string(p.proconfig, '|'), '-') || ':' || coalesce(array_to_string(p.proacl, ','), '-'), ';'
  ORDER BY p.proname) FROM pg_proc p WHERE p.oid IN (to_regprocedure('app.groups_bootstrap_owner()'),
  to_regprocedure('app.rls_profile_eligible(uuid)'))`),
			"groups_bootstrap_owner:"+o+":false:v:search_path=pg_catalog:"+o+"=X/"+o+";"+
				"rls_profile_eligible:"+pr+":true:s:"+sp+":"+pr+"=X/"+pr+","+rtr+"=X/"+pr+","+o+"=X/"+pr; got != want {
			t.Errorf("functions:\n got %s\nwant %s", got, want)
		}
		if got := one(cat, `SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.rls_profile_eligible(uuid)'::regprocedure`); got !=
			md5hex(onbBody(t, "CREATE FUNCTION app.rls_profile_eligible(")) {
			t.Error("rls_profile_eligible differs from the reviewed migration body")
		}
		if got := one(cat, `SELECT md5(prosrc) FROM pg_proc WHERE oid = 'app.business_profiles_keys_immutable()'::regprocedure`); got !=
			md5hex(onbBody(t, "CREATE OR REPLACE FUNCTION app.business_profiles_keys_immutable()")) {
			t.Error("business_profiles_keys_immutable differs from the reviewed migration body")
		}
		if src := one(cat, `SELECT prosrc FROM pg_proc WHERE oid = 'app.groups_bootstrap_owner()'::regprocedure`); !strings.Contains(src,
			"current_user::text NOT IN ('"+rtr+"', '"+o+"') OR app.ctx_user() IS NULL") ||
			!strings.Contains(src, "VALUES (NEW.organization_id, app.ctx_user(), 'group', NEW.id, NEW.id, 'group_owner', 'membership')") {
			t.Errorf("groups_bootstrap_owner body: %s", src)
		}
		if got := one(cat, `SELECT string_agg(t.tgname || ':' || pg_get_triggerdef(t.oid), E'\n' ORDER BY t.tgname)
FROM pg_trigger t WHERE t.tgrelid = 'app.groups'::regclass AND t.tgname = 'groups_bootstrap_owner'`); got !=
			"groups_bootstrap_owner:CREATE TRIGGER groups_bootstrap_owner AFTER INSERT ON app.groups FOR EACH ROW EXECUTE FUNCTION app.groups_bootstrap_owner()" {
			t.Errorf("bootstrap trigger: %s", got)
		}
		// The provenance columns, their defaults and the backfill of the pre-existing row.
		if got := one(cat, `SELECT string_agg(format('%s:%s:%s:%s', a.attname, format_type(a.atttypid, a.atttypmod), a.attnotnull,
  coalesce(pg_get_expr(d.adbin, d.adrelid), '')), ';' ORDER BY a.attnum)
FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = 'app.business_profiles'::regclass AND a.attnum > 8 AND NOT a.attisdropped`); got !=
			"created_xid:xid8:t:pg_current_xact_id();created_sysid:bigint:t:(pg_control_system()).system_identifier" {
			t.Errorf("provenance columns: %s", got)
		}
		var backfill string
		if err := setup.QueryRow(ctx, `SELECT format('%s:%s', created_xid <> pg_current_xact_id(),
  created_sysid = (pg_control_system()).system_identifier) FROM app.business_profiles WHERE id = $1`, onbProfile).Scan(&backfill); err != nil ||
			backfill != "t:t" {
			t.Errorf("backfilled row: %q %v", backfill, err)
		}
		// The FORCE-bound owner cannot move either provenance value (rolled back; FORCE lifted inside the transaction only).
		for _, c := range []string{"created_xid = '1'::xid8", "created_sysid = created_sysid + 1", "id = gen_random_uuid()"} {
			tx, err := cat.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			_, err = tx.Exec(ctx, `ALTER TABLE app.business_profiles NO FORCE ROW LEVEL SECURITY`)
			if err == nil {
				_, err = tx.Exec(ctx, `UPDATE app.business_profiles SET `+c+` WHERE id = '`+onbProfile+`'`)
			}
			_ = tx.Rollback(ctx)
			if sqlstateOf(err) != "23000" {
				t.Errorf("owner UPDATE %s: %v, want 23000", c, err)
			}
		}

		// Runtime behaviour under actual grants.
		newGroup := "0f18c000-0000-4000-8000-0000000000a1"
		for _, creator := range []string{onbOwner, onbMember} {
			inTx(creator, onbOrg, "", func(tx pgx.Tx) {
				// S-G1 keyed insert without RETURNING; the trigger writes the creator's owner binding; S-G2 reads the row.
				if _, err := tx.Exec(ctx, `INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1, $2, 'Fresh Group', 'fresh-group')`,
					newGroup, onbOrg); err != nil {
					t.Fatalf("creator %s: keyed Group insert: %v", creator, err)
				}
				var got string
				if err := tx.QueryRow(ctx, `SELECT g.slug || ':' || (SELECT string_agg(rb.subject_user_id::text || ':' || rb.role || ':' ||
  rb.source, ',') FROM app.role_bindings rb WHERE rb.scope_type = 'group' AND rb.scope_id = g.id) FROM app.groups g WHERE g.id = $1`,
					newGroup).Scan(&got); err != nil || got != "fresh-group:"+creator+":group_owner:membership" {
					t.Errorf("creator %s: S-G2 %q %v", creator, got, err)
				}
			})
		}
		denied := func(what, user, org, group, sql string, args ...any) {
			t.Helper()
			inTx(user, org, group, func(tx pgx.Tx) {
				if _, err := tx.Exec(ctx, sql, args...); sqlstateOf(err) != "42501" {
					t.Errorf("%s: %v, want 42501", what, err)
				}
			})
		}
		ins := `INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1, $2, 'Fresh Group', 'fresh-group')`
		denied("Group insert in selected-Group context", onbOwner, onbOrg, onbGroup, ins, newGroup, onbOrg)
		denied("Group insert by a Group-role-only subject", onbGOnly, onbOrg, "", ins, newGroup, onbOrg)
		denied("Group insert into a foreign Organization", onbOwner, onbOrg, "", ins, newGroup, onbOrgB)
		denied("Group insert without context", "", "", "", ins, newGroup, onbOrg)
		denied("depth-0 owner binding on an existing Group", onbMember, onbOrg, "", `INSERT INTO app.role_bindings
  (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES ($1, $2, 'group', $3, $3, 'group_owner', 'membership')`,
			onbOrg, onbOwner, onbGroup)
		denied("depth-0 own owner binding", onbOwner, onbOrg, "", `INSERT INTO app.role_bindings
  (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES ($1, $2, 'group', $3, $3, 'group_owner', 'membership')`,
			onbOrg, onbOwner, onbGroup)
		for _, tbl := range []string{"app.groups", "app.role_bindings", "app.business_profiles", "app.group_business_profiles"} {
			denied("runtime DELETE "+tbl, onbOwner, onbOrg, "", `DELETE FROM `+tbl)
		}
		// The insertion predicate: true for a same-transaction insert; false for a committed row, for a committed row
		// UPDATEd in this transaction, and without context.
		fresh := "0f18c000-0000-4000-8000-0000000000a2"
		inTx(onbOwner, onbOrg, "", func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, `INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ($1, $2, 'Fresh')`,
				fresh, onbOrg); err != nil {
				t.Fatalf("keyed profile insert: %v", err)
			}
			if !eligible(tx, fresh) {
				t.Error("a profile inserted by this transaction is not eligible")
			}
			if eligible(tx, onbProfile) {
				t.Error("a committed unattached profile is eligible")
			}
			if _, err := tx.Exec(ctx, `UPDATE app.business_profiles SET description = 'updated' WHERE id = $1`, onbProfile); err != nil {
				t.Fatalf("owner update: %v", err)
			}
			if eligible(tx, onbProfile) {
				t.Error("a committed profile UPDATEd by this transaction is eligible")
			}
		})
		inTx("", "", "", func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, `INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ($1, $2, 'Fresh')`,
				fresh, onbOrg); sqlstateOf(err) != "42501" {
				t.Errorf("profile insert without context: %v, want 42501", err)
			}
		})
		inTx("", "", "", func(tx pgx.Tx) {
			if eligible(tx, onbProfile) {
				t.Error("the reader is true without context")
			}
		})
	})

	t.Run("CT-ONB-1/R-ONB-1-neg a restored row colliding with the evaluating transaction id is never eligible", func(t *testing.T) {
		restored := "0f18c000-0000-4000-8000-0000000000b1"
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if _, err := tx.Exec(ctx, `INSERT INTO app.business_profiles (id, organization_id, display_name, created_xid, created_sysid)
VALUES ($1, $2, 'Restored', pg_current_xact_id(), (pg_control_system()).system_identifier + 1)`, restored, onbOrg); err != nil {
			t.Fatalf("restore-shaped insert: %v", err)
		}
		var collides bool
		if err := tx.QueryRow(ctx, `SELECT created_xid = pg_current_xact_id() FROM app.business_profiles WHERE id = $1`, restored).Scan(&collides); err != nil || !collides {
			t.Fatalf("the collision is not real: %v %v", collides, err)
		}
		if tag, err := tx.Exec(ctx, `UPDATE app.business_profiles SET description = 'refreshed' WHERE id = $1`, restored); err != nil || tag.RowsAffected() != 1 {
			t.Fatalf("superuser update: %v %v", tag, err)
		}
		if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+quote(rtr)); err != nil {
			t.Fatal(err)
		}
		rlsContext(t, tx, onbMember, onbOrg, "", "")
		if eligible(tx, restored) {
			t.Error("the restored colliding row is eligible")
		}
		_, err = tx.Exec(ctx, `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1, $2, $3)`,
			onbOrg, onbGroup, restored)
		if sqlstateOf(err) != "42501" {
			t.Errorf("attaching the restored row: %v, want 42501", err)
		}
	})

	t.Run("CT-ONB-1/R-ONB-1-pos a fresh insert is eligible and attaches", func(t *testing.T) {
		fresh := "0f18c000-0000-4000-8000-0000000000b2"
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		rlsContext(t, tx, onbMember, onbOrg, "", "")
		if _, err := tx.Exec(ctx, `INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ($1, $2, 'Fresh')`,
			fresh, onbOrg); err != nil {
			t.Fatalf("S-P1: %v", err)
		}
		if !eligible(tx, fresh) {
			t.Error("the fresh insert is not eligible")
		}
		if tag, err := tx.Exec(ctx, `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1, $2, $3)`,
			onbOrg, onbGroup, fresh); err != nil || tag.RowsAffected() != 1 {
			t.Fatalf("S-P2: %v %v", tag, err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		var got string
		if err := setup.QueryRow(ctx, `SELECT format('%s:%s', created_sysid = (pg_control_system()).system_identifier,
  created_xid <> pg_current_xact_id()) FROM app.business_profiles WHERE id = $1`, fresh).Scan(&got); err != nil || got != "t:t" {
			t.Errorf("fresh row provenance: %q %v", got, err)
		}
	})

	t.Run("CT-ONB-AUD the widened operation constraint", func(t *testing.T) {
		def := one(cat, `SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'audit.events'::regclass
  AND conname = 'events_operation_valid'`)
		ops := regexp.MustCompile(`'([a-z_.]+)'::text`).FindAllStringSubmatch(def, -1)
		if len(ops) != 36 {
			t.Errorf("events_operation_valid lists %d operations, want 36: %s", len(ops), def)
		}
		got := map[string]bool{}
		for _, m := range ops {
			got[m[1]] = true
		}
		for _, op := range append(slicesClone(onbOps), "access.resolve", "branding.remove_logo", "organization.set_slug") {
			if !got[op] {
				t.Errorf("operation %s missing: %s", op, def)
			}
		}
		// The constraint itself (triggers disabled for the superuser fixture only; rolled back).
		probe := func(op string) error {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
				t.Fatal(err)
			}
			_, err = tx.Exec(ctx, `INSERT INTO audit.events (action, outcome, actor_kind, operation) VALUES ('access.denied', 'denied', 'system', $1)`, op)
			return err
		}
		if err := probe("group.unknown"); sqlstateOf(err) != "23514" {
			t.Errorf("unknown operation: %v, want 23514", err)
		}
		for _, op := range onbOps {
			if err := probe(op); err != nil {
				t.Errorf("operation %s refused: %v", op, err)
			}
		}
	})

	t.Run("CT-ONB-UPDOWN clean down/up and the history-preserving down refusal", func(t *testing.T) {
		down := func() (string, error) {
			var b bytes.Buffer
			err := r.Down(ctx, &b)
			return b.String(), err
		}
		// The rows written above at version 14 are tenancy rows; the down keeps them (it drops only provenance).
		out, err := down()
		if err != nil || out != "rolled_back\t00014_group_onboarding.sql\ncurrent_version\t13\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		metadataThroughVersion13Only(t, cfg, r)
		if got := snapshot(t, r); got != s13 {
			t.Fatalf("catalog after down differs from version 13:\n%s\n---\n%s", got, s13)
		}
		objRunUp(t, r, "applied\t00014_group_onboarding.sql\ncurrent_version\t14\n")
		if got := snapshot(t, r); got != s14 {
			t.Fatalf("catalog after down/up differs:\n%s\n---\n%s", got, s14)
		}
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range []string{`SET LOCAL session_replication_role = replica`,
			`INSERT INTO audit.events (action, outcome, actor_kind, operation) VALUES ('access.denied', 'denied', 'system', 'group.create')`} {
			if _, err := tx.Exec(ctx, s); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatal(err)
			}
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00014: audit.events holds onboarding history") {
			t.Fatalf("down with onboarding history: %v %q", err, out)
		}
		if got := snapshot(t, r); got != s14 {
			t.Fatal("the refused down changed the catalog")
		}
		var n int
		if err := setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE operation = 'group.create'`).Scan(&n); err != nil || n != 1 {
			t.Fatalf("history after the refused down: %d %v", n, err)
		}
		if got := one(cat, `SELECT relforcerowsecurity::text FROM pg_class WHERE oid = 'audit.events'::regclass`); got != "true" {
			t.Fatalf("audit.events FORCE after the refused down: %s", got)
		}
		reset(t, cfg)
	})
}

func slicesClone(s []string) []string { return append([]string(nil), s...) }
