//go:build integration

// G020 B-LABELS-CATALOG (Plan r2 §3, §5.2 T7, §8 AC1/AC5/AC6) against the disposable PostgreSQL of the migrations suite.
// TestLabelsCatalog pins the exact version-19 -> 20 delta: the two tables, their columns, ACLs and FORCE (CT-LAB-1); the
// production-closed target storage of app.label_assignments (CT-LAB-2: the closed target-shape CHECK body, the SELECT
// policy body and the column list, all group-only); the seven policies, five triggers and five functions, no
// Group-management override or trigger-depth predicate in any label policy, and the capture shapes (CT-LAB-3); the
// unchanged helper inventories and the one-index-probe EXISTS of the assignments SELECT policy (CT-LAB-4).
// TestLabelsMigration covers the populated version-19 world to 20, the history-preserving down refusal with a label row
// and with label history only, down to 19 and up/reapply restoring the exact catalogs, and a rolled-back up (MG-LAB-1..4).
// Expected function bodies are read from the embedded migration itself, so the applied catalog must equal the reviewed
// file. Setup and rolled-back controls use the attested superuser; catalog reads use the owner (search_path pg_catalog).
// This database is never the one the test-only fixture widening (core/internal/tenancy/access labels_fixture_test.go)
// alters: the migrate package owns its own disposable database.
package migrate

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// The production storage fragments of app.label_assignments (CT-LAB-2), exported so a fixture-teardown test can compare
// the restored production bodies byte for byte. LabelsWantTargetShape is pg_get_constraintdef of
// label_assignments_target_shape; LabelsWantAssignmentsSelect is pg_get_expr of the label_assignments_select USING
// expression, whitespace-normalized (strings.Fields joined by one space) and printed with search_path = pg_catalog;
// LabelsWantAssignmentColumns is the production column list in attnum order (no per-type target key column exists in
// production storage; the first one arrives with a future target type's own reviewed migration).
const (
	LabelsWantTargetShape = `CHECK (((owner_scope_type = 'organization'::text) AND (owner_scope_id = organization_id) ` +
		`AND (target_type = 'group'::text) AND (target_group_id = target_id)))`
	LabelsWantAssignmentsSelect = `((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) ` +
		`AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (target_group_id = ( SELECT app.ctx_group() AS ctx_group))) ` +
		`AND (EXISTS ( SELECT 1 FROM app.label_definitions d WHERE ((d.id = label_assignments.label_id) ` +
		`AND (d.organization_id = label_assignments.organization_id)))) AND (target_type = 'group'::text))`
	LabelsWantAssignmentColumns = "id,organization_id,label_id,owner_scope_type,owner_scope_id,target_type,target_id,target_group_id,created_at"
)

// labBS is one backslash: the G063 control class is assembled from it so that no tool rewrites its escapes.
const labBS = "\x5c"

var (
	// labWantNameCheck is label_definitions_name_valid as PostgreSQL prints it (the 00019 display-name grammar, 1..80).
	labWantNameCheck = `CHECK ((((char_length(name) >= 1) AND (char_length(name) <= 80)) AND (name = btrim(name)) AND ((name COLLATE "C") !~ '[` +
		labBS + `u0001-` + labBS + `u001f` + labBS + `u007f-` + labBS + `u009f` + labBS + `u2028` + labBS + `u2029]'::text)))`

	labWantColumns = map[string]string{
		"label_definitions": "id:uuid:t::-,organization_id:uuid:t::-,owner_scope_type:text:t::-,owner_scope_id:uuid:t::-," +
			"group_id:uuid:f::-,name:text:t::-,name_key:text:f:s:lower((name COLLATE \"C\")),visibility:text:t::-," +
			"created_at:timestamp with time zone:t::now()",
		"label_assignments": "id:uuid:t::-,organization_id:uuid:t::-,label_id:uuid:t::-,owner_scope_type:text:t::-," +
			"owner_scope_id:uuid:t::-,target_type:text:t::-,target_id:uuid:t::-,target_group_id:uuid:t::-," +
			"created_at:timestamp with time zone:t::now()",
	}
	labWantTableACL  = map[string]string{"label_definitions": "rt:DELETE,rt:SELECT", "label_assignments": "rt:DELETE,rt:SELECT"}
	labWantColumnACL = map[string]string{
		"label_definitions": "group_id:rt:INSERT,id:rt:INSERT,name:rt:INSERT,name:rt:UPDATE,organization_id:rt:INSERT," +
			"owner_scope_id:rt:INSERT,owner_scope_type:rt:INSERT,visibility:rt:INSERT,visibility:rt:UPDATE",
		"label_assignments": "id:rt:INSERT,label_id:rt:INSERT,organization_id:rt:INSERT,owner_scope_id:rt:INSERT," +
			"owner_scope_type:rt:INSERT,target_group_id:rt:INSERT,target_id:rt:INSERT,target_type:rt:INSERT",
	}
	labWantPolicies = "label_assignments_delete:d:true:ow+rt,label_assignments_insert:a:true:ow+rt,label_assignments_select:r:true:ow+rt," +
		"label_definitions_delete:d:true:ow+rt,label_definitions_insert:a:true:ow+rt,label_definitions_select:r:true:ow+rt," +
		"label_definitions_update:w:true:ow+rt"
	labWantTriggers = []string{
		"CREATE TRIGGER audit_capture_label_assignments AFTER INSERT OR DELETE ON app.label_assignments FOR EACH ROW " +
			"EXECUTE FUNCTION audit.capture_label_assignments()",
		"CREATE TRIGGER label_assignments_keys_immutable BEFORE UPDATE ON app.label_assignments FOR EACH ROW " +
			"EXECUTE FUNCTION app.label_assignments_keys_immutable()",
		"CREATE TRIGGER audit_capture_label_definitions AFTER INSERT OR DELETE OR UPDATE OF name, visibility ON app.label_definitions " +
			"FOR EACH ROW EXECUTE FUNCTION audit.capture_label_definitions()",
		"CREATE TRIGGER label_definitions_bound BEFORE INSERT ON app.label_definitions FOR EACH ROW " +
			"EXECUTE FUNCTION app.label_definitions_bound()",
		"CREATE TRIGGER label_definitions_keys_immutable BEFORE UPDATE ON app.label_definitions FOR EACH ROW " +
			"EXECUTE FUNCTION app.label_definitions_keys_immutable()",
	}
	labFunctions = []string{"app.label_assignments_keys_immutable()", "app.label_definitions_bound()",
		"app.label_definitions_keys_immutable()", "audit.capture_label_assignments()", "audit.capture_label_definitions()"}
	// labAuditDelta are the existing keys version 20 changes (namesCatalog): the four catalog constraints and the
	// attribution body. Every other delta key names a label object.
	labAuditDelta = []string{"con:audit.events.events_action_valid", "con:audit.events.events_operation_valid",
		"con:audit.events.events_reason_valid", "con:audit.events.events_resource_type_valid", "fn:audit.events_attribute()"}
)

// labWantLabelKeys is the number of namesCatalog keys naming a label object: 9 relations, 18 columns, 15 column ACLs,
// 5 functions, 7 policies, 2 FORCE flags, 15 constraints and 5 triggers.
const labWantLabelKeys = 9 + 18 + 15 + 5 + 7 + 2 + 15 + 5

const (
	labOrg  = "0f020000-0000-4000-8000-000000000101"
	labGrp  = "0f020000-0000-4000-8000-000000000201"
	labDefO = "0f020000-0000-4000-8000-000000000301" // Organization-owned
	labDefG = "0f020000-0000-4000-8000-000000000302" // Group-owned
	labAsg  = "0f020000-0000-4000-8000-000000000401"
)

func isLabelKey(k string) bool {
	return strings.Contains(k, "label_definitions") || strings.Contains(k, "label_assignments") || strings.Contains(k, "capture_label_")
}

// labExec runs each statement on the transaction, failing the test on any error.
func labExec(t *testing.T, tx pgx.Tx, stmts ...string) {
	t.Helper()
	for _, s := range stmts {
		if _, err := tx.Exec(context.Background(), s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
}

var labSeedSQL = []string{
	`INSERT INTO app.organizations (id, name) VALUES ('` + labOrg + `', 'Labels Catalog')`,
	`INSERT INTO app.groups (id, organization_id, name, slug) VALUES ('` + labGrp + `', '` + labOrg + `', 'Labels Group', 'labels-group')`,
}

var labRowsSQL = []string{
	`INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility) VALUES
  ('` + labDefO + `', '` + labOrg + `', 'organization', '` + labOrg + `', NULL, 'Priority Secret', 'owner_scope'),
  ('` + labDefG + `', '` + labOrg + `', 'group', '` + labGrp + `', '` + labGrp + `', 'Design Secret', 'owner_scope')`,
	`INSERT INTO app.label_assignments (id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id,
  target_group_id) VALUES ('` + labAsg + `', '` + labOrg + `', '` + labDefO + `', 'organization', '` + labOrg + `', 'group', '` +
		labGrp + `', '` + labGrp + `')`,
}

// labPlanNodes collects every plan node (with its InitPlans and SubPlans) of an EXPLAIN (FORMAT JSON) document.
func labPlanNodes(n map[string]any, out *[]map[string]any) {
	*out = append(*out, n)
	if kids, ok := n["Plans"].([]any); ok {
		for _, k := range kids {
			if m, ok := k.(map[string]any); ok {
				labPlanNodes(m, out)
			}
		}
	}
}

func TestLabelsCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	objRunUp(t, runnerFor(t, cfg, subsetFS(t, 19)), "current_version\t19\n")
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	rt := quote(cfg.RuntimeRole())
	s19 := namesCatalog(t, cat)
	r := runnerFor(t, cfg, subsetFS(t, 20))
	objRunUp(t, r, "applied\t00020_labels.sql\ncurrent_version\t20\n")
	s20 := namesCatalog(t, cat)
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
	const roleName = `CASE WHEN a.grantee = $1::regrole THEN 'rt' WHEN a.grantee = (SELECT nspowner FROM pg_namespace WHERE nspname = 'app') ` +
		`THEN 'ow' WHEN a.grantee = 0 THEN 'PUBLIC' ELSE a.grantee::regrole::text END`

	t.Run("CT-LAB-1 the exact version-20 delta: tables, columns, ACLs, FORCE", func(t *testing.T) {
		var labelKeys []string
		for k := range s20 {
			if isLabelKey(k) {
				labelKeys = append(labelKeys, k)
				if _, ok := s19[k]; ok {
					t.Errorf("%s exists at version 19", k)
				}
			}
		}
		if len(labelKeys) != labWantLabelKeys {
			t.Errorf("version 20 names %d label catalog keys, want %d: %v", len(labelKeys), labWantLabelKeys, labelKeys)
		}
		want := append(labelKeys, labAuditDelta...)
		slices.Sort(want)
		if got := namesDelta(s19, s20); !slices.Equal(got, want) {
			t.Fatalf("version-19 -> 20 delta: %v, want exactly %v", got, want)
		}
		for tbl, w := range labWantColumns {
			if got := one(`SELECT string_agg(format('%s:%s:%s:%s:%s', a.attname, format_type(a.atttypid, a.atttypmod), a.attnotnull, a.attgenerated,
  coalesce(pg_get_expr(d.adbin, d.adrelid), '-')), ',' ORDER BY a.attnum)
FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = ('app.' || $1)::regclass AND a.attnum > 0 AND NOT a.attisdropped`, tbl); got != w {
				t.Errorf("%s columns:\n got %s\nwant %s", tbl, got, w)
			}
		}
		for tbl, w := range labWantTableACL {
			if got := one(`SELECT coalesce(string_agg(x, ',' ORDER BY x COLLATE "C"), '') FROM (SELECT `+roleName+` || ':' || a.privilege_type AS x
  FROM pg_class c, aclexplode(c.relacl) a WHERE c.oid = ('app.' || $2)::regclass AND a.grantee <> c.relowner) s`, cfg.RuntimeRole(), tbl); got != w {
				t.Errorf("%s table ACL: %s, want %s", tbl, got, w)
			}
		}
		for tbl, w := range labWantColumnACL {
			if got := one(`SELECT coalesce(string_agg(x, ',' ORDER BY x COLLATE "C"), '') FROM (SELECT at.attname || ':' || `+roleName+` || ':' ||
  a.privilege_type AS x FROM pg_attribute at, aclexplode(at.attacl) a WHERE at.attrelid = ('app.' || $2)::regclass AND at.attnum > 0
  AND NOT at.attisdropped) s`, cfg.RuntimeRole(), tbl); got != w {
				t.Errorf("%s column ACL:\n got %s\nwant %s", tbl, got, w)
			}
		}
		for _, k := range []string{"rls:app.label_definitions", "rls:app.label_assignments"} {
			if s20[k] != "t|t" {
				t.Errorf("%s: %q, want ENABLE and FORCE row-level security", k, s20[k])
			}
		}
		if got := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'app.label_definitions'::regclass
  AND conname = 'label_definitions_name_valid'`); got != labWantNameCheck {
			t.Errorf("name CHECK:\n got %s\nwant %s", got, labWantNameCheck)
		}
		// The four foreign keys: RESTRICT, validated, and the exact referencing/referenced columns.
		if got := one(`SELECT string_agg(format('%s>%s:%s:%s', conname, confrelid::regclass, confdeltype, convalidated), ',' ORDER BY conname::text COLLATE "C")
FROM pg_constraint WHERE conrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass) AND contype = 'f'`); got !=
			"label_assignments_label_fkey>app.label_definitions:r:t,label_assignments_target_group_fkey>app.groups:r:t,"+
				"label_definitions_group_fkey>app.groups:r:t,label_definitions_organization_fkey>app.organizations:r:t" {
			t.Errorf("label foreign keys: %s", got)
		}
		if got := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'app.label_assignments'::regclass
  AND conname = 'label_assignments_label_fkey'`); got != "FOREIGN KEY (organization_id, label_id, owner_scope_type, owner_scope_id) "+
			"REFERENCES app.label_definitions(organization_id, id, owner_scope_type, owner_scope_id) ON DELETE RESTRICT" {
			t.Errorf("assignment -> definition key: %s", got)
		}
	})

	t.Run("CT-LAB-2 production assignment storage is closed to group targets", func(t *testing.T) {
		if got := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'app.label_assignments'::regclass
  AND conname = 'label_assignments_target_shape'`); got != LabelsWantTargetShape {
			t.Errorf("target-shape CHECK:\n got %s\nwant %s", got, LabelsWantTargetShape)
		}
		if got := normalize(one(`SELECT pg_get_expr(polqual, polrelid) FROM pg_policy WHERE polrelid = 'app.label_assignments'::regclass
  AND polname = 'label_assignments_select'`)); got != LabelsWantAssignmentsSelect {
			t.Errorf("assignments SELECT policy:\n got %s\nwant %s", got, LabelsWantAssignmentsSelect)
		}
		if got := one(`SELECT string_agg(attname, ',' ORDER BY attnum) FROM pg_attribute WHERE attrelid = 'app.label_assignments'::regclass
  AND attnum > 0 AND NOT attisdropped`); got != LabelsWantAssignmentColumns {
			t.Errorf("assignment columns: %s, want %s", got, LabelsWantAssignmentColumns)
		}
		// Robust fragments independent of the exact deparse: group is the only target kind and no fixture object exists.
		if got := one(`SELECT count(*)::text FROM pg_class WHERE relnamespace = 'app'::regnamespace AND relname LIKE '%fixture%'`); got != "0" {
			t.Errorf("%s fixture relations in production storage", got)
		}
		for k, v := range s20 {
			if isLabelKey(k) && (strings.Contains(k, "fixture") || strings.Contains(v, "fixture")) {
				t.Errorf("%s names a fixture: %s", k, v)
			}
		}
		if def := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = 'label_assignments_target_shape'`); strings.Count(def, "target_type") != 1 ||
			!strings.Contains(def, "(target_type = 'group'::text)") || strings.Contains(def, " OR ") {
			t.Errorf("target shape is not closed to group: %s", def)
		}
	})

	t.Run("CT-LAB-3 policies, triggers, functions and capture shapes", func(t *testing.T) {
		if got := one(`SELECT string_agg(p.polname::text || ':' || p.polcmd::text || ':' || p.polpermissive::text || ':' ||
  (SELECT string_agg(k, '+' ORDER BY k COLLATE "C") FROM (SELECT CASE WHEN x = $1::regrole THEN 'rt'
     WHEN x = (SELECT nspowner FROM pg_namespace WHERE nspname = 'app') THEN 'ow' ELSE x::regrole::text END AS k FROM unnest(p.polroles) x) s),
  ',' ORDER BY p.polname::text COLLATE "C")
FROM pg_policy p WHERE p.polrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass)`, cfg.RuntimeRole()); got != labWantPolicies {
			t.Errorf("label policies:\n got %s\nwant %s", got, labWantPolicies)
		}
		rows, err := cat.Query(ctx, `SELECT p.polname, coalesce(pg_get_expr(p.polqual, p.polrelid), '') || ' | ' || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')
FROM pg_policy p WHERE p.polrelid IN ('app.label_definitions'::regclass, 'app.label_assignments'::regclass)`)
		if err != nil {
			t.Fatal(err)
		}
		exprs := map[string]string{}
		for rows.Next() {
			var n, e string
			if err := rows.Scan(&n, &e); err != nil {
				t.Fatal(err)
			}
			exprs[n] = e
		}
		rows.Close()
		for n, e := range exprs {
			// No Group-management override (OVR), departure-cascade depth (DEP) or override column in any label policy.
			for _, bad := range []string{"group_management_override", "pg_trigger_depth", "app.organizations"} {
				if strings.Contains(e, bad) {
					t.Errorf("%s reads %s: %s", n, bad, e)
				}
			}
			if !strings.Contains(e, "( SELECT app.ctx_org() AS ctx_org)") || !strings.Contains(e, "( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL") {
				t.Errorf("%s does not bind the context Organization with a current role: %s", n, e)
			}
		}
		// Organization-owned rows need the Organization owner/admin; Group-owned rows need a role in the exact Group.
		for _, n := range []string{"label_definitions_select", "label_definitions_insert", "label_definitions_update", "label_definitions_delete"} {
			if e := exprs[n]; !strings.Contains(e, "(owner_scope_type = 'organization'::text)") || !strings.Contains(e, "app.rls_group_role(group_id)") ||
				!strings.Contains(e, "ARRAY['org_owner'::text, 'org_admin'::text]") {
				t.Errorf("%s: %s", n, e)
			}
		}
		if e := exprs["label_definitions_select"]; !strings.Contains(e, "(visibility = 'target_members'::text)") ||
			!strings.Contains(e, "(app.rls_group_role(group_id) IS NOT NULL)") {
			t.Errorf("label_definitions_select lacks the target_members member branch: %s", e)
		}
		for _, n := range []string{"label_definitions_insert", "label_definitions_update", "label_definitions_delete", "label_assignments_insert",
			"label_assignments_delete"} {
			if e := exprs[n]; strings.Contains(e, "target_members") || strings.Contains(e, "rls_group_role(group_id) IS NOT NULL") ||
				strings.Contains(e, "rls_group_role(target_group_id) IS NOT NULL") || !strings.Contains(e, "ARRAY['group_owner'::text, 'group_admin'::text]") {
				t.Errorf("%s admits a non-manager: %s", n, e)
			}
		}
		var trg []string
		tr, err := cat.Query(ctx, `SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE tgrelid IN ('app.label_definitions'::regclass,
  'app.label_assignments'::regclass) AND NOT tgisinternal AND tgenabled = 'O' ORDER BY tgrelid::regclass::text COLLATE "C", tgname::text COLLATE "C"`)
		if err != nil {
			t.Fatal(err)
		}
		for tr.Next() {
			var d string
			if err := tr.Scan(&d); err != nil {
				t.Fatal(err)
			}
			trg = append(trg, d)
		}
		tr.Close()
		if !slices.Equal(trg, labWantTriggers) {
			t.Errorf("label triggers:\n got %v\nwant %v", trg, labWantTriggers)
		}
		for _, fn := range append(slices.Clone(labFunctions), "audit.events_attribute()") {
			if got := one(`SELECT format('%s|%s|%s|%s', p.proowner = (SELECT nspowner FROM pg_namespace WHERE nspname = 'app'), p.prosecdef,
  coalesce(array_to_string(p.proconfig, ';'), '-'), coalesce((SELECT string_agg(a.grantee::regrole::text, '+') FROM aclexplode(p.proacl) a
  WHERE a.grantee <> p.proowner), '-')) FROM pg_proc p WHERE p.oid = to_regprocedure($1)`, fn); got != "t|f|search_path=pg_catalog, pg_temp|-" {
				t.Errorf("%s: %s, want an owner-owned invoker body with the pinned search_path and no other grantee", fn, got)
			}
		}
		for fn, marker := range map[string]string{
			"app.label_definitions_keys_immutable()": "CREATE FUNCTION app.label_definitions_keys_immutable()",
			"app.label_assignments_keys_immutable()": "CREATE FUNCTION app.label_assignments_keys_immutable()",
			"audit.capture_label_definitions()":      "CREATE FUNCTION audit.capture_label_definitions()",
			"audit.capture_label_assignments()":      "CREATE FUNCTION audit.capture_label_assignments()",
			"audit.events_attribute()":               "CREATE OR REPLACE FUNCTION audit.events_attribute()",
		} {
			if got, want := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure($1)`, fn), md5hex(migBody(t, "00020_labels.sql", marker)); got != want {
				t.Errorf("%s body digest %s, want %s (the reviewed file)", fn, got, want)
			}
		}
		if src := one(`SELECT prosrc FROM pg_proc WHERE oid = 'app.label_definitions_bound()'::regprocedure`); !strings.Contains(src, "FOR NO KEY UPDATE") ||
			!strings.Contains(src, "IF n >= 200 THEN") || !strings.Contains(src, "CONSTRAINT = 'label_definitions_bound'") ||
			!strings.Contains(src, "'"+cfg.RuntimeRole()+"'") {
			t.Errorf("label_definitions_bound body: %s", src)
		}
		for name, want := range map[string]int{"events_action_valid": 52, "events_reason_valid": 81, "events_operation_valid": 80,
			"events_resource_type_valid": 12} {
			if n := strings.Count(one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'audit.events'::regclass AND conname = $1`, name),
				"::text"); n != want {
				t.Errorf("%s names %d values, want %d", name, n, want)
			}
		}

		// The captures under the reviewed shapes (rolled back, superuser): five actions, the definition as the resource, the
		// owning or target Group, renamed and visibility_changed as separate events, a no-op update silent, no name.
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		labExec(t, tx, labSeedSQL...)
		labExec(t, tx, labRowsSQL...)
		labExec(t, tx,
			`UPDATE app.label_definitions SET name = 'Design Renamed', visibility = 'target_members' WHERE id = '`+labDefG+`'`,
			`UPDATE app.label_definitions SET name = 'Design Renamed', visibility = 'target_members' WHERE id = '`+labDefG+`'`,
			`DELETE FROM app.label_assignments WHERE id = '`+labAsg+`'`,
			`DELETE FROM app.label_definitions WHERE id = '`+labDefO+`'`)
		var got string
		if err := tx.QueryRow(ctx, `SELECT string_agg(x, ',' ORDER BY x COLLATE "C") FROM (SELECT e.action || ':' || e.outcome || ':' ||
  coalesce(e.reason, '-') || ':' || CASE WHEN e.group_id IS NULL THEN '-' WHEN e.group_id = $2::uuid THEN 'g' ELSE '?' END || ':' ||
  CASE e.resource_id WHEN $3::uuid THEN 'O' WHEN $4::uuid THEN 'G' ELSE '?' END || ':' || e.resource_type || ':' ||
  (e.organization_id = $1::uuid)::text || ':' || (e.subject_user_id IS NULL AND e.role IS NULL AND e.operation IS NULL)::text AS x
  FROM audit.events e WHERE e.action LIKE 'label.%') s`, labOrg, labGrp, labDefO, labDefG).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if want := "label.assigned:success:-:g:O:label:true:true,label.created:success:-:-:O:label:true:true," +
			"label.created:success:-:g:G:label:true:true,label.deleted:success:-:-:O:label:true:true," +
			"label.unassigned:success:-:g:O:label:true:true,label.updated:success:renamed:g:G:label:true:true," +
			"label.updated:success:visibility_changed:g:G:label:true:true"; got != want {
			t.Errorf("label audit rows:\n got %s\nwant %s", got, want)
		}
		var leaks int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM audit.events e WHERE position('Secret' IN e::text) > 0 OR position('Renamed' IN e::text) > 0`).Scan(&leaks); err != nil || leaks != 0 {
			t.Errorf("%d audit rows carry a label name (%v)", leaks, err)
		}
	})

	t.Run("CT-LAB-4 helper inventories unchanged; the assignments EXISTS is one index probe", func(t *testing.T) {
		db := cfg.Database
		for role, want := range map[string]string{
			db + "_policy_reader": "rls_dependent_group_bindings,rls_enrollment_mode,rls_group_invitation_candidates,rls_group_role,rls_org_role," +
				"rls_own_organizations,rls_owner_count,rls_profile_eligible,rls_sole_owner_scopes",
			db + "_claim_writer": "enrollment_admit,enrollment_lock_organization,enrollment_request,invitation_claim_group,invitation_claim_organization," +
				"rls_invitation_pending,rls_user_enabled_lock",
			db + "_auth_reader": "auth_logout_ref,auth_principal_subject,auth_resolve_enrollment,auth_resolve_session,auth_resolve_signup,hostname_state," +
				"is_organization_member,public_branding",
			db + "_provision_writer": "signup_challenge_complete,signup_challenge_issue,signup_provision_organization",
		} {
			if got := one(`SELECT string_agg(proname, ',' ORDER BY proname COLLATE "C") FROM pg_proc WHERE proowner = $1::regrole`, role); got != want {
				t.Errorf("%s owns %s, want %s", role, got, want)
			}
			if got := one(`SELECT (has_table_privilege($1, 'app.label_definitions', 'SELECT, INSERT, UPDATE, DELETE')
  OR has_table_privilege($1, 'app.label_assignments', 'SELECT, INSERT, UPDATE, DELETE')
  OR has_any_column_privilege($1, 'app.label_definitions', 'SELECT, INSERT, UPDATE')
  OR has_any_column_privilege($1, 'app.label_assignments', 'SELECT, INSERT, UPDATE'))::text`, role); got != "false" {
				t.Errorf("%s holds a label table privilege", role)
			}
		}
		for _, role := range []string{cfg.RetentionRole(), cfg.MigratorRole()} {
			if got := one(`SELECT (has_table_privilege($1, 'app.label_definitions', 'SELECT, INSERT, UPDATE, DELETE')
  OR has_table_privilege($1, 'app.label_assignments', 'SELECT, INSERT, UPDATE, DELETE'))::text`, role); got != "false" {
				t.Errorf("%s holds a label table privilege", role)
			}
		}
		if got := one(`SELECT count(*)::text FROM pg_proc WHERE prosecdef AND proname LIKE 'label%'`); got != "0" {
			t.Errorf("%s SECURITY DEFINER label functions", got)
		}

		// E-PG-style plan summary (risk r2): under the runtime role, with the hash and merge strategies and sequential
		// scans disabled, every access to app.label_definitions made by the assignments SELECT policy's EXISTS is an index
		// probe on its primary or owner key; there is no sequential scan of the definitions.
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		labExec(t, tx, `SET LOCAL ROLE `+rt, `SET LOCAL enable_seqscan = off`, `SET LOCAL enable_hashjoin = off`,
			`SET LOCAL enable_mergejoin = off`, `SET LOCAL enable_bitmapscan = off`)
		var raw []byte
		if err := tx.QueryRow(ctx, `EXPLAIN (FORMAT JSON, COSTS OFF) SELECT a.id FROM app.label_assignments a`).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var doc []map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil || len(doc) != 1 {
			t.Fatalf("plan: %v %s", err, raw)
		}
		root, _ := doc[0]["Plan"].(map[string]any)
		var nodes []map[string]any
		labPlanNodes(root, &nodes)
		probes := 0
		for _, n := range nodes {
			if n["Relation Name"] != "label_definitions" {
				continue
			}
			typ, idx := n["Node Type"], n["Index Name"]
			if (typ != "Index Scan" && typ != "Index Only Scan") || (idx != "label_definitions_pkey" && idx != "label_definitions_owner_key") {
				t.Errorf("assignments SELECT reads label_definitions by %v %v, want an index probe on its key: %s", typ, idx, raw)
				continue
			}
			probes++
		}
		if probes != 1 {
			t.Errorf("assignments SELECT probes label_definitions %d times, want exactly one index probe: %s", probes, raw)
		}
	})
}

// TestLabelsMigration (B-LABELS-CATALOG): MG-LAB-1..4.
func TestLabelsMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	down := func(r *Runner) (string, error) {
		var b bytes.Buffer
		err := r.Down(ctx, &b)
		return b.String(), err
	}
	commit := func(stmts ...string) {
		t.Helper()
		for _, s := range stmts {
			if _, err := setup.Exec(ctx, s); err != nil {
				t.Fatalf("%s: %v", s, err)
			}
		}
	}
	var s19, s20 map[string]string
	r := runnerFor(t, cfg, subsetFS(t, 20))

	t.Run("MG-LAB-1 a populated version-19 world to 20", func(t *testing.T) {
		reset(t, cfg)
		objRunUp(t, runnerFor(t, cfg, subsetFS(t, 19)), "current_version\t19\n")
		commit(labSeedSQL...)
		s19 = namesCatalog(t, cat)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00020_labels.sql\ncurrent_version\t20\n" {
			t.Fatalf("up: %q", out)
		}
		s20 = namesCatalog(t, cat)
		metadataThroughVersion20(t, cfg, r)
		var n int
		if err := setup.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.label_definitions) + (SELECT count(*) FROM app.label_assignments)
  + (SELECT count(*) FROM audit.events WHERE action LIKE 'label.%')`).Scan(&n); err != nil || n != 0 {
			t.Fatalf("after the upgrade: %d label rows or events (%v), want 0", n, err)
		}
	})

	t.Run("MG-LAB-2 down refused with a label row or label history, catalog unchanged", func(t *testing.T) {
		commit(labRowsSQL...)
		if out, err := down(r); err == nil || !strings.Contains(err.Error(),
			"refusing to roll back 00020: 2 label definitions / 1 label assignments / 3 label audit rows (history is never discarded)") {
			t.Fatalf("down with label rows: %v %q", err, out)
		}
		commit(`DELETE FROM app.label_assignments`, `DELETE FROM app.label_definitions`)
		if out, err := down(r); err == nil || !strings.Contains(err.Error(),
			"refusing to roll back 00020: 0 label definitions / 0 label assignments / 6 label audit rows (history is never discarded)") {
			t.Fatalf("down with label history: %v %q", err, out)
		}
		if got := namesDelta(s20, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("the refused down changed the catalog: %v", got)
		}
	})

	t.Run("MG-LAB-3 down and up on a label-free world restore the exact 19 and 20 catalogs", func(t *testing.T) {
		auditPurge(t, setup)
		out, err := down(r)
		if err != nil || out != "rolled_back\t00020_labels.sql\ncurrent_version\t19\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		if got := namesDelta(s19, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down differs from version 19: %v", got)
		}
		metadataThroughVersion19(t, cfg, r)
		objRunUp(t, r, "applied\t00020_labels.sql\ncurrent_version\t20\n")
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out !=
			"rolled_back\t00020_labels.sql\napplied\t00020_labels.sql\ncurrent_version\t20\n" {
			t.Fatalf("reapply: %q", out)
		}
		if got := namesDelta(s20, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down/up/reapply differs from version 20: %v", got)
		}
	})

	t.Run("MG-LAB-4 a rolled-back up leaves version 19 and the owner", func(t *testing.T) {
		if out, err := down(r); err != nil || !strings.HasSuffix(out, "current_version\t19\n") {
			t.Fatalf("down: %v %q", err, out)
		}
		// A stray relation in schema app breaks the version-19 relation precondition: the up fails and rolls back entirely.
		commit(`CREATE TABLE app.label_stray (id int)`)
		var buf bytes.Buffer
		if err := r.Up(ctx, &buf); err == nil || !strings.Contains(err.Error(), "version-19 precondition: schema app holds 80 relations") {
			t.Fatalf("up with a stray relation: %v %q", err, buf.String())
		}
		if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t19\n" {
			t.Fatalf("version after the failed up: %q", out)
		}
		commit(`DROP TABLE app.label_stray`)
		if got := namesDelta(s19, namesCatalog(t, cat)); len(got) != 0 {
			t.Errorf("after the failed up the catalog differs from version 19 by %v", got)
		}
		rlsOwnerRestored(t, cfg, r)
		objRunUp(t, r, "applied\t00020_labels.sql\ncurrent_version\t20\n")
	})
}
