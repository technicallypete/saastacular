package migrations

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

var nameRe = regexp.MustCompile(`^[0-9]{5}_[a-z0-9_]+\.sql$`)

func files(t *testing.T) []string {
	t.Helper()
	names, err := fs.Glob(FS, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("no migrations embedded")
	}
	return names
}

func TestEveryMigrationIsVersionedAndReversible(t *testing.T) {
	for _, name := range files(t) {
		if !nameRe.MatchString(name) {
			t.Errorf("%s: name must match NNNNN_name.sql", name)
		}
		b, err := fs.ReadFile(FS, name)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		up := strings.Index(src, "-- +goose Up")
		down := strings.Index(src, "-- +goose Down")
		if up < 0 || down < 0 || down < up {
			t.Errorf("%s: needs a '-- +goose Up' section followed by a '-- +goose Down' section", name)
		}
	}
}

// The migration set is exactly 00001-00022 (G071 adds 00022_default_tenant.sql; G057 adds 00021_platform_labels.sql; G020 adds 00020_labels.sql; G063 adds 00019_mcp_connection_names.sql; G062 adds 00018_platform_signup.sql; G060 adds 00017_group_authority.sql; G008 adds 00006_row_level_security.sql, G009 00007_stored_objects.sql, G011
// 00008_audit_events.sql, G013 00009_retention_maintenance.sql, G014 00010_mcp_connections.sql, G016
// 00011_organization_domains.sql and its security amendment 00012_session_binding_and_claim_exclusivity.sql, G017
// 00013_organization_branding.sql, G018 00014_group_onboarding.sql, G019 00015_membership_invitations.sql, G058 00016_organization_enrollment.sql after the
// accepted history), and the accepted 00006, 00007, 00008 and 00009 stay byte-identical (G009 Plan r1 §5.6; G011 Plan §5.7; G013
// Plan r3 §6.1; G014 Plan r8 §5).
func TestMigrationSet(t *testing.T) {
	want := "00001_baseline.sql,00002_identity.sql,00003_tenancy.sql,00004_business_profiles.sql,00005_organization_policy.sql," +
		"00006_row_level_security.sql,00007_stored_objects.sql,00008_audit_events.sql,00009_retention_maintenance.sql," +
		"00010_mcp_connections.sql,00011_organization_domains.sql,00012_session_binding_and_claim_exclusivity.sql," +
		"00013_organization_branding.sql,00014_group_onboarding.sql,00015_membership_invitations.sql,00016_organization_enrollment.sql," +
		"00017_group_authority.sql,00018_platform_signup.sql,00019_mcp_connection_names.sql,00020_labels.sql,00021_platform_labels.sql," +
		"00022_default_tenant.sql"
	if got := strings.Join(files(t), ","); got != want {
		t.Fatalf("migrations = %s, want %s", got, want)
	}
	for name, sum := range map[string]string{"00006_row_level_security.sql": migration00006SHA, "00007_stored_objects.sql": migration00007SHA,
		"00008_audit_events.sql": migration00008SHA, "00009_retention_maintenance.sql": migration00009SHA} {
		b, err := fs.ReadFile(FS, name)
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != sum {
			t.Fatalf("%s changed: sha256 %s, want %s", name, got, sum)
		}
	}
}

// migration00006SHA is the accepted G008 00006_row_level_security.sql (as integrated).
const migration00006SHA = "fae1813b6b747a4d385eb16a4b900c92290b3eb82331ca85bb6d41a44c7384ed"

// migration00007SHA is the accepted G009 00007_stored_objects.sql (newly pinned by G011).
const migration00007SHA = "28f0a073631a0b10ca17602e2b5897a6c02a17e305ca795c959a89d403100d12"

// migration00008SHA is the accepted G011 00008_audit_events.sql (newly pinned by G013).
const migration00008SHA = "b1dedf1f56f28c97a9f8c59796ff1b615c1ed79e6b8ab8f2cfb602098c597233"

// migration00009SHA is the accepted G013 00009_retention_maintenance.sql (newly pinned by G014).
const migration00009SHA = "7fbfe7f9671d702e9ea1b4ed89f5e7c925e2409e2e3dd2dfc60e68022ade9b2b"

// The retention migration (G013 Plan r3 §6.1) is additive and privilege-only: no table, column, index, view, sequence,
// type, extension, schema, role or trigger; exactly one new function (app.retention_eligible, SECURITY INVOKER, STABLE,
// pinned search_path), the replaced audit.events_refuse body, exactly four permissive command-specific policies TO the
// retention role only, and grants to the retention role only (schema USAGE, the minimal column SELECT, DELETE on the two
// tables, EXECUTE on the function); no INSERT, UPDATE, TRUNCATE, REFERENCES, TRIGGER or ALL grant, no DML, no SECURITY
// DEFINER, LEAKPROOF, BYPASSRLS, FOR ALL, RESTRICTIVE, NO FORCE or DISABLE TRIGGER, no exception block. Its down section
// drops the four policies and the function, restores the version-8 refusal body and revokes the grants.
func TestRetentionMigrationShape(t *testing.T) {
	b, err := fs.ReadFile(FS, "00009_retention_maintenance.sql")
	if err != nil {
		t.Fatal(err)
	}
	rawUp, rawDown, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	strip := regexp.MustCompile(`--[^\n]*`)
	up, down := strip.ReplaceAllString(rawUp, " "), strip.ReplaceAllString(rawDown, " ")
	for _, re := range []string{
		`(?i)\b(CREATE|DROP)\s+(TABLE|INDEX|VIEW|SEQUENCE|TYPE|EXTENSION|SCHEMA|ROLE|TRIGGER)\b`,
		`(?i)\bINSERT\s+INTO\b|\bUPDATE\s+(app|audit)\.|\bDELETE\s+FROM\b|\bTRUNCATE\b|\bALTER\s+(TABLE|ROLE)\b`,
		`(?i)\bLEAKPROOF\b|\bBYPASSRLS\b|\bFOR\s+ALL\b|\bAS\s+RESTRICTIVE\b|SECURITY\s+DEFINER|\bNO\s+FORCE\b|DISABLE\s+TRIGGER|EXCEPTION\s+WHEN`,
		`(?i)\bGRANT\b[^;]*\b(INSERT|UPDATE|TRUNCATE|REFERENCES|TRIGGER|ALL)\b`,
	} {
		if m := regexp.MustCompile(re).FindString(up); m != "" {
			t.Errorf("00009 up contains %q", m)
		}
	}
	funcs := regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+([a-z_]+\.[a-z_]+)`).FindAllStringSubmatch(up, -1)
	if len(funcs) != 2 || funcs[0][2] != "app.retention_eligible" || funcs[0][1] != "" || funcs[1][2] != "audit.events_refuse" || funcs[1][1] == "" {
		t.Errorf("functions = %v", funcs)
	}
	if !strings.Contains(up, "RETURNS boolean\n  LANGUAGE plpgsql STABLE SET search_path = pg_catalog, pg_temp AS $$") {
		t.Error("app.retention_eligible must be a STABLE invoker function with the pinned search_path")
	}
	var got []string
	for _, m := range regexp.MustCompile(`CREATE POLICY ([a-z_]+) ON ([a-z_.]+) FOR (SELECT|DELETE) TO %I USING`).FindAllStringSubmatch(up, -1) {
		got = append(got, m[1]+":"+m[2]+":"+m[3])
	}
	if strings.Join(got, ",") != "events_retention_select:audit.events:SELECT,events_retention_delete:audit.events:DELETE,"+
		"stored_objects_retention_select:app.stored_objects:SELECT,stored_objects_retention_delete:app.stored_objects:DELETE" ||
		strings.Count(up, "CREATE POLICY ") != 4 {
		t.Errorf("policies = %v", got)
	}
	grants := regexp.MustCompile(`GRANT [^']*TO %I`).FindAllString(up, -1)
	if strings.Join(grants, "|") != "GRANT EXECUTE ON FUNCTION app.retention_eligible(uuid, timestamptz, uuid[]) TO %I|"+
		"GRANT USAGE ON SCHEMA audit TO %I|GRANT SELECT (id, occurred_at, organization_id), DELETE ON audit.events TO %I|"+
		"GRANT SELECT (id, organization_id, state, deleted_at), DELETE ON app.stored_objects TO %I" {
		t.Errorf("grants = %v", grants)
	}
	if n := strings.Count(up, ", rn::text);"); n != 9 {
		t.Errorf("%d statements formatted with the retention role, want 9 (four grants, four policies, the refusal body)", n)
	}
	for _, d := range []string{"DROP POLICY events_retention_select ON audit.events;", "DROP POLICY events_retention_delete ON audit.events;",
		"DROP POLICY stored_objects_retention_select ON app.stored_objects;", "DROP POLICY stored_objects_retention_delete ON app.stored_objects;",
		"DROP FUNCTION app.retention_eligible(uuid, timestamptz, uuid[]);", "REVOKE USAGE ON SCHEMA audit FROM %I",
		"RAISE EXCEPTION 'audit events are append-only' USING ERRCODE = 'insufficient_privilege';\nEND $b$;"} {
		if !strings.Contains(down, d) {
			t.Errorf("down lacks %q", d)
		}
	}
}

// The audit migration (G011 Plan §5) creates objects only in schema audit (one table, one index, eleven SECURITY
// INVOKER functions with a pinned search_path, three triggers on audit.events) plus the eighteen capture triggers on
// eight app tables; exactly two permissive command-specific policies on audit.events, ENABLE and FORCE row-level
// security, the schema USAGE, SELECT and twelve-column INSERT runtime grants; no app table, column, policy, function or
// grant change, no role, membership, SECURITY DEFINER, LEAKPROOF, BYPASSRLS, foreign key, retention grant or DML on an
// app table. The exception blocks are exactly the two revocation-priority branches and hold only the audit insert. Its
// down section refuses while rows exist.
func TestAuditMigrationShape(t *testing.T) {
	b, err := fs.ReadFile(FS, "00008_audit_events.sql")
	if err != nil {
		t.Fatal(err)
	}
	rawUp, rawDown, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	strip := regexp.MustCompile(`--[^\n]*`)
	up, down := strip.ReplaceAllString(rawUp, " "), strip.ReplaceAllString(rawDown, " ")
	for _, re := range []string{
		`(?i)\b(CREATE|DROP)\s+(VIEW|SEQUENCE|TYPE|EXTENSION|SCHEMA|ROLE)\b`,
		`(?i)\bUPDATE\s+app\.|\bDELETE\s+FROM\b|\bTRUNCATE\s+app\.|\bINSERT\s+INTO\s+app\.`,
		`(?i)\bLEAKPROOF\b|\bBYPASSRLS\b|\bFOR\s+ALL\b|\bAS\s+RESTRICTIVE\b|SECURITY\s+DEFINER|\bREFERENCES\b|\bALTER\s+TABLE\s+app\.`,
		`(?i)\bGRANT\b[^;]*\b(UPDATE|DELETE|TRUNCATE|TRIGGER)\b`, `(?i)\bGRANT\b[^;]*_retention`, `(?i)row_to_json|to_jsonb|hstore|NEW\.\*|OLD\.\*`,
		`(?i)\bCREATE\s+POLICY\b[^;]*\bON\s+app\.`, `(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+app\.`,
	} {
		if m := regexp.MustCompile(re).FindString(up); m != "" {
			t.Errorf("00008 up contains %q", m)
		}
	}
	var created []string
	for _, m := range regexp.MustCompile(`(?i)\bCREATE\s+(TABLE|INDEX|FUNCTION)\s+(audit\.)?([a-z_]+)`).FindAllStringSubmatch(up, -1) {
		created = append(created, m[3])
	}
	if strings.Join(created, ",") != "events,events_tenant_time_idx,events_attribute,events_refuse,capture_organizations,capture_groups,"+
		"capture_role_bindings,capture_business_profiles,capture_group_business_profiles,capture_users,capture_sessions,"+
		"capture_session_revocations,capture_stored_objects" {
		t.Errorf("created objects = %v", created)
	}
	if n := len(regexp.MustCompile(`(?i)\bCREATE\s+FUNCTION\s+audit\.[a-z_]+\(\)\s+RETURNS\s+trigger\s+LANGUAGE\s+plpgsql\s+SET\s+search_path\s+=\s+pg_catalog,\s+pg_temp\s+AS`).FindAllString(up, -1)); n != 11 {
		t.Errorf("%d invoker trigger functions with the pinned search_path, want 11", n)
	}
	triggers := regexp.MustCompile(`(?i)\bCREATE\s+TRIGGER\s+([a-z_]+)\s+(BEFORE|AFTER)[^;]*?\bON\s+(app|audit)\.([a-z_]+)`).FindAllStringSubmatch(up, -1)
	var onApp, onAudit int
	for _, m := range triggers {
		if m[3] == "app" {
			onApp++
		} else {
			onAudit++
		}
	}
	if onApp != 18 || onAudit != 3 {
		t.Errorf("%d app capture triggers and %d audit.events triggers, want 18 and 3", onApp, onAudit)
	}
	if n := strings.Count(up, "REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows"); n != 1 {
		t.Errorf("%d transition-table triggers, want 1", n)
	}
	if n := strings.Count(up, "CREATE POLICY "); n != 2 {
		t.Errorf("%d CREATE POLICY statements, want 2", n)
	}
	for _, p := range []string{"events_select ON audit.events FOR SELECT", "events_insert ON audit.events FOR INSERT",
		"ALTER TABLE audit.events ENABLE ROW LEVEL SECURITY", "ALTER TABLE audit.events FORCE ROW LEVEL SECURITY"} {
		if !strings.Contains(up, p) {
			t.Errorf("missing %s", p)
		}
	}
	grants := regexp.MustCompile(`GRANT [^']*TO %I`).FindAllString(up, -1)
	if strings.Join(grants, "|") != "GRANT USAGE ON SCHEMA audit TO %I|GRANT SELECT ON audit.events TO %I|"+
		"GRANT INSERT (action, outcome, reason, operation, resource_type, resource_id, subject_user_id, organization_id, group_id, role, "+
		"previous_role, ttl_seconds) ON audit.events TO %I" {
		t.Errorf("runtime grants = %v", grants)
	}
	// The only exception blocks: the users active -> disabled branch and the session revocation priority branch, each
	// holding only the audit insert; the handler only records the gap marker.
	blocks := regexp.MustCompile(`(?s)\bBEGIN\s+(INSERT INTO audit\.events [^;]*;)\s+EXCEPTION WHEN OTHERS THEN\s+(PERFORM pg_catalog\.set_config\('audit\.revocation_gap',[^;]*;)\s+END;`).FindAllStringSubmatch(up, -1)
	if len(blocks) != 2 || strings.Count(up, "EXCEPTION WHEN") != 2 {
		t.Fatalf("%d bounded exception blocks (%d EXCEPTION WHEN), want exactly 2", len(blocks), strings.Count(up, "EXCEPTION WHEN"))
	}
	if !strings.Contains(blocks[0][1], "'user.disabled'") || !strings.Contains(blocks[1][1], "IN ('logout', 'logout_all', 'provider_logout', 'disabled')") ||
		strings.Contains(blocks[1][1], "superseded") || !strings.Contains(blocks[0][2], "'user.disabled'") || !strings.Contains(blocks[1][2], "'session.revoked'") {
		t.Errorf("exception blocks = %v", blocks)
	}
	if !strings.Contains(down, "IF EXISTS (SELECT 1 FROM audit.events) THEN") || !strings.Contains(down, "DROP TABLE audit.events;") ||
		strings.Count(down, "DROP TRIGGER audit_capture_") != 18 || strings.Count(down, "DROP FUNCTION audit.") != 11 {
		t.Error("down must refuse while rows exist, then drop the 18 capture triggers, the table and the 11 functions")
	}
}

// The stored-object migration (G009 Plan r1 §3) is additive: exactly one table and one index, one SECURITY INVOKER
// trigger function with a pinned search_path and its BEFORE UPDATE trigger, exactly three permissive command-specific
// policies on app.stored_objects (no FOR ALL, no RESTRICTIVE), ENABLE and FORCE row-level security, and only the three
// runtime grants; no role, membership, SECURITY DEFINER, LEAKPROOF, BYPASSRLS, DML, DELETE grant or change of any other
// relation. Its down section refuses while rows exist, then drops the table and the function.
func TestStoredObjectsMigrationShape(t *testing.T) {
	b, err := fs.ReadFile(FS, "00007_stored_objects.sql")
	if err != nil {
		t.Fatal(err)
	}
	rawUp, rawDown, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	strip := regexp.MustCompile(`--[^\n]*`)
	up, down := strip.ReplaceAllString(rawUp, " "), strip.ReplaceAllString(rawDown, " ")
	for _, re := range []string{
		`(?i)\b(CREATE|DROP)\s+(VIEW|SEQUENCE|TYPE|EXTENSION|SCHEMA|ROLE)\b`,
		`(?i)\bINSERT\s+INTO\b|\bUPDATE\s+app\.|\bDELETE\s+FROM\b|\bTRUNCATE\b`,
		`(?i)\bLEAKPROOF\b|\bBYPASSRLS\b|\bFOR\s+ALL\b|\bAS\s+RESTRICTIVE\b|SECURITY\s+DEFINER|\bGRANT\s+\S+\s+TO\b`,
		`(?i)\bGRANT\b[^;]*\b(DELETE|TRUNCATE|REFERENCES|TRIGGER)\b`,
	} {
		if m := regexp.MustCompile(re).FindString(up); m != "" {
			t.Errorf("00007 up contains %q", m)
		}
	}
	if got := regexp.MustCompile(`(?i)\bCREATE\s+(TABLE|INDEX|FUNCTION|TRIGGER)\s+(app\.)?([a-z_]+)`).FindAllStringSubmatch(up, -1); len(got) != 4 ||
		got[0][3] != "stored_objects" || got[1][3] != "stored_objects_organization_group_idx" || got[2][3] != "stored_objects_guard" ||
		got[3][3] != "stored_objects_guard" {
		t.Errorf("created objects = %v", got)
	}
	if regexp.MustCompile(`(?i)\bALTER\s+TABLE\s+app\.`).FindAllString(up, -1) != nil {
		for _, m := range regexp.MustCompile(`(?i)\bALTER\s+TABLE\s+app\.[a-z_]+\s+[A-Z ]+`).FindAllString(up, -1) {
			if m != "ALTER TABLE app.stored_objects ENABLE ROW LEVEL SECURITY" && m != "ALTER TABLE app.stored_objects FORCE ROW LEVEL SECURITY" {
				t.Errorf("00007 up alters a table: %q", m)
			}
		}
	}
	if n := strings.Count(up, "CREATE POLICY "); n != 3 {
		t.Errorf("%d CREATE POLICY statements, want 3", n)
	}
	for _, p := range []string{"stored_objects_select ON app.stored_objects FOR SELECT", "stored_objects_insert ON app.stored_objects FOR INSERT",
		"stored_objects_update ON app.stored_objects FOR UPDATE"} {
		if !strings.Contains(up, p) {
			t.Errorf("missing policy %s", p)
		}
	}
	grants := regexp.MustCompile(`GRANT [^']*ON app\.stored_objects TO %I`).FindAllString(up, -1)
	if strings.Join(grants, "|") != "GRANT SELECT ON app.stored_objects TO %I|"+
		"GRANT INSERT (id, organization_id, group_id, creator_user_id, byte_size, content_type) ON app.stored_objects TO %I|"+
		"GRANT UPDATE (state, sha256) ON app.stored_objects TO %I" {
		t.Errorf("runtime grants = %v", grants)
	}
	if !strings.Contains(up, "storage_key     text GENERATED ALWAYS AS ('objects/' || organization_id::text || '/' || id::text) STORED") {
		t.Error("the provider key is not the generated objects/<organization>/<id> column")
	}
	if !strings.Contains(down, "IF EXISTS (SELECT 1 FROM app.stored_objects) THEN") || !strings.Contains(down, "DROP TABLE app.stored_objects;") ||
		!strings.Contains(down, "DROP FUNCTION app.stored_objects_guard();") ||
		strings.Index(down, "IF EXISTS (SELECT 1 FROM app.stored_objects)") > strings.Index(down, "DROP TABLE") {
		t.Error("down must refuse while rows exist, then drop the table and the guard")
	}
}

// The row-level security migration (G008) changes no data and no relation: no table, column, index, constraint,
// trigger or DML; it creates exactly the four parsers and five readers (SECURITY DEFINER only for the readers, every
// function with a pinned search_path, none LEAKPROOF), ENABLE and FORCE on exactly the nine tables, exactly 26
// permissive command-specific policies (no FOR ALL), no BYPASSRLS, and its down section drops what it created.
func TestRowLevelSecurityMigrationShape(t *testing.T) {
	b, err := fs.ReadFile(FS, "00006_row_level_security.sql")
	if err != nil {
		t.Fatal(err)
	}
	rawUp, rawDown, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	strip := regexp.MustCompile(`--[^\n]*`)
	up, down := strip.ReplaceAllString(rawUp, " "), strip.ReplaceAllString(rawDown, " ")
	for _, re := range []string{
		`(?i)\b(CREATE|DROP)\s+(TABLE|INDEX|TRIGGER|VIEW|SEQUENCE|TYPE|EXTENSION|SCHEMA|ROLE)\b`,
		`(?i)\bALTER\s+TABLE\s+app\.\S+\s+(ADD|DROP|ALTER|RENAME|OWNER|SET)\b`,
		`(?i)\bINSERT\s+INTO\b|\bUPDATE\s+app\.|\bDELETE\s+FROM\b|\bTRUNCATE\b`,
		`(?i)\bADD\s+COLUMN\b|\bLEAKPROOF\b|\bBYPASSRLS\b|\bFOR\s+ALL\b|\bAS\s+RESTRICTIVE\b`,
	} {
		if m := regexp.MustCompile(re).FindString(up); m != "" {
			t.Errorf("00006 up contains %q", m)
		}
	}
	funcs := regexp.MustCompile(`(?s)CREATE FUNCTION app\.([a-z_]+)\(.*?END \$\$;`).FindAllStringSubmatch(up, -1)
	var names []string
	for _, f := range funcs {
		names = append(names, f[1])
		definer := strings.Contains(f[0], "SECURITY DEFINER")
		if definer != (strings.HasPrefix(f[1], "rls_") || strings.HasPrefix(f[1], "auth_")) {
			t.Errorf("%s: SECURITY DEFINER = %v", f[1], definer)
		}
		if !strings.Contains(f[0], "STABLE") || !strings.Contains(f[0], "SET search_path = pg_catalog, pg_temp") {
			t.Errorf("%s: must be STABLE with search_path pg_catalog, pg_temp", f[1])
		}
	}
	if got := strings.Join(names, ","); got != "ctx_user,ctx_org,ctx_group,ctx_login_binding,rls_org_role,rls_group_role,"+
		"auth_resolve_session,auth_principal_subject,auth_logout_ref" {
		t.Errorf("created functions = %s", got)
	}
	for _, f := range funcs {
		if regexp.MustCompile(`(?i)\bEXECUTE\b`).MatchString(f[0]) {
			t.Errorf("%s declares dynamic SQL", f[1])
		}
	}
	if n := strings.Count(up, "CREATE POLICY "); n != 26 {
		t.Errorf("%d CREATE POLICY statements, want 26", n)
	}
	tables := "'organizations', 'groups', 'role_bindings', 'business_profiles', 'group_business_profiles',\n" +
		"                             'users', 'sessions', 'identity_principals', 'login_transactions'"
	if !strings.Contains(up, tables) || !strings.Contains(up, "ENABLE ROW LEVEL SECURITY") || !strings.Contains(up, "FORCE ROW LEVEL SECURITY") {
		t.Error("up does not ENABLE and FORCE row-level security on exactly the nine tables")
	}
	if !strings.Contains(down, tables) || !strings.Contains(down, "NO FORCE ROW LEVEL SECURITY") || !strings.Contains(down, "DISABLE ROW LEVEL SECURITY") {
		t.Error("down does not remove row-level security from the nine tables")
	}
	for _, f := range names {
		if !regexp.MustCompile(`DROP FUNCTION app\.` + f + `\(`).MatchString(down) {
			t.Errorf("down does not drop %s", f)
		}
	}
	if !strings.Contains(down, "DROP POLICY") {
		t.Error("down does not drop the policies")
	}
	// The E5 cleanup policy keeps both EMPTY-shape terms (N13's static twin).
	if !regexp.MustCompile(`login_transactions_expire ON app\.login_transactions FOR DELETE TO %I, %I USING \('\s*\|\|\s*'expires_at <= now\(\) AND %s IS NULL AND %s IS NULL\)', rt, ow, U, B\)`).MatchString(up) {
		t.Error("login_transactions_expire must be expires_at <= now() AND U IS NULL AND B IS NULL")
	}
}

// The baseline is an empty-schema marker: it may assert, but it creates no
// schema objects and grants nothing.
func TestBaselineCreatesNothing(t *testing.T) {
	b, err := fs.ReadFile(FS, "00001_baseline.sql")
	if err != nil {
		t.Fatal(err)
	}
	// Only SQL is checked: comments and string literals are removed first.
	code := regexp.MustCompile(`--[^\n]*|'(?:[^']|'')*'`).ReplaceAllString(string(b), " ")
	forbidden := regexp.MustCompile(`(?i)\b(CREATE|ALTER|DROP)\s+(TABLE|VIEW|MATERIALIZED|FUNCTION|PROCEDURE|POLICY|SCHEMA|SEQUENCE|TYPE|INDEX|TRIGGER|ROLE)\b|\b(GRANT|REVOKE)\b`)
	if m := forbidden.FindString(code); m != "" {
		t.Fatalf("baseline must not create objects or change privileges; found %q", m)
	}
}

// The identity migration (G004) creates only the five global identity tables, with no tenant, role or
// vendor-organization column and no row-level security; it grants DELETE only on login transactions, keeps MFA
// assurance fixed to 'unknown', and its down section drops exactly those tables.
func TestIdentityMigrationShape(t *testing.T) {
	b, err := fs.ReadFile(FS, "00002_identity.sql")
	if err != nil {
		t.Fatal(err)
	}
	rawUp, rawDown, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	strip := regexp.MustCompile(`--[^\n]*`)
	up, down := strip.ReplaceAllString(rawUp, " "), strip.ReplaceAllString(rawDown, " ")
	created := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+app\.([a-z_]+)`).FindAllStringSubmatch(up, -1)
	var names []string
	for _, m := range created {
		names = append(names, m[1])
	}
	if got := strings.Join(names, ","); got != "identity_provider_binding,users,identity_principals,sessions,login_transactions" {
		t.Fatalf("created tables = %s", got)
	}
	for _, re := range []string{
		`(?i)\b(organization|group|tenant|membership|role|permission|org)_?(id|s)?\b\s+(uuid|text|bigint|integer)`,
		`(?i)ROW\s+LEVEL\s+SECURITY|CREATE\s+POLICY|\bBYPASSRLS\b`,
		`(?i)CREATE\s+(FUNCTION|PROCEDURE|TRIGGER|SEQUENCE|VIEW|SCHEMA|ROLE|EXTENSION)\b`,
	} {
		if m := regexp.MustCompile(re).FindString(up); m != "" {
			t.Errorf("identity migration contains %q", m)
		}
	}
	grants := regexp.MustCompile(`(?i)GRANT\s+([A-Z, ]+?)\s+ON\s+([a-z_., ]+?)\s+TO`).FindAllStringSubmatch(up, -1)
	if len(grants) != 3 {
		t.Fatalf("want exactly three GRANT statements, got %d", len(grants))
	}
	for _, g := range grants {
		if strings.Contains(strings.ToUpper(g[1]), "DELETE") && strings.TrimSpace(g[2]) != "app.login_transactions" {
			t.Errorf("DELETE granted on %s", g[2])
		}
		for _, p := range []string{"TRUNCATE", "REFERENCES", "TRIGGER", "ALL"} {
			if strings.Contains(strings.ToUpper(g[1]), p) {
				t.Errorf("%s granted on %s", p, g[2])
			}
		}
	}
	if !regexp.MustCompile(`mfa_assurance\s+text NOT NULL DEFAULT 'unknown' CHECK \(mfa_assurance = 'unknown'\)`).MatchString(up) {
		t.Error("mfa_assurance must be constrained to 'unknown'")
	}
	dropped := regexp.MustCompile(`(?i)DROP\s+TABLE\s+app\.([a-z_]+);`).FindAllStringSubmatch(down, -1)
	names = names[:0]
	for _, m := range dropped {
		names = append(names, m[1])
	}
	if got := strings.Join(names, ","); got != "login_transactions,sessions,identity_principals,users,identity_provider_binding" {
		t.Errorf("down drops %s", got)
	}
}

// The MCP connection migration (G014 Plan r8 §5) adds one table (app.mcp_connections: no foreign key on the selected
// Group, no index beyond its primary and two unique keys), one sessions column, two new SECURITY INVOKER trigger functions
// with a pinned search_path (the connection guard and the audit capture), three triggers on the new table, three
// permissive command-specific policies on it plus the replaced sessions_insert check, and runtime grants of SELECT, the
// eight INSERT columns and the five UPDATE columns only. The only replaced functions are audit.events_attribute and the
// E1 reader (the latter by its auth-reader owner, with the kind filter); no new SECURITY DEFINER function, no DELETE,
// TRUNCATE, REFERENCES or TRIGGER grant, no DML, role, LEAKPROOF, BYPASSRLS, FOR ALL, RESTRICTIVE or exception block. Its
// down section refuses while MCP state or history exists and restores the version-9 objects.
func TestMCPMigrationShape(t *testing.T) {
	b, err := fs.ReadFile(FS, "00010_mcp_connections.sql")
	if err != nil {
		t.Fatal(err)
	}
	rawUp, rawDown, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	strip := regexp.MustCompile(`--[^\n]*`)
	up, down := strip.ReplaceAllString(rawUp, " "), strip.ReplaceAllString(rawDown, " ")
	for _, re := range []string{
		`(?i)\b(CREATE|DROP)\s+(INDEX|VIEW|SEQUENCE|TYPE|EXTENSION|SCHEMA|ROLE)\b`,
		`(?i)\bINSERT\s+INTO\s+app\.|\bUPDATE\s+app\.|\bDELETE\s+FROM\b|\bTRUNCATE\b`,
		`(?i)\bLEAKPROOF\b|\bBYPASSRLS\b|\bFOR\s+ALL\b|\bAS\s+RESTRICTIVE\b|\bNO\s+FORCE\b|DISABLE\s+TRIGGER|EXCEPTION\s+WHEN`,
		`(?i)\bGRANT\b[^;]*\b(DELETE|TRUNCATE|REFERENCES|TRIGGER|ALL)\b`, `(?i)\bGRANT\b[^;]*_retention`,
		`(?i)selected_group_id\s+uuid[^,]*REFERENCES`,
	} {
		if m := regexp.MustCompile(re).FindString(up); m != "" {
			t.Errorf("00010 up contains %q", m)
		}
	}
	var created []string
	for _, m := range regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?(TABLE|FUNCTION)\s+([a-z_]+\.[a-z_]+)`).FindAllStringSubmatch(up, -1) {
		kind := "new"
		if m[1] != "" {
			kind = "replaced"
		}
		created = append(created, m[3]+":"+kind)
	}
	if strings.Join(created, ",") != "app.mcp_connections:new,app.mcp_connections_guard:new,audit.events_attribute:replaced,"+
		"audit.capture_mcp_connections:new,app.auth_resolve_session:replaced" {
		t.Errorf("created objects = %v", created)
	}
	if n := len(regexp.MustCompile(`(?i)\bCREATE\s+FUNCTION\s+[a-z_]+\.[a-z_]+\(\)\s+RETURNS\s+trigger\s+LANGUAGE\s+plpgsql\s+SET\s+search_path\s+=\s+pg_catalog,\s+pg_temp\s+AS`).FindAllString(up, -1)); n != 2 {
		t.Errorf("%d new invoker trigger functions with the pinned search_path, want 2", n)
	}
	if n := strings.Count(up, "SECURITY DEFINER"); n != 1 || !strings.Contains(up, "AND u.status = 'active' AND s.kind = 'browser';\nEND $$\n$e1$") {
		t.Errorf("the only SECURITY DEFINER text must be the replaced E1 reader with the kind filter (%d)", n)
	}
	var triggers []string
	for _, m := range regexp.MustCompile(`(?i)\bCREATE\s+TRIGGER\s+([a-z_]+)\s+(BEFORE|AFTER)[^;]*?\bON\s+([a-z_.]+)`).FindAllStringSubmatch(up, -1) {
		triggers = append(triggers, m[1]+":"+m[2]+":"+m[3])
	}
	if strings.Join(triggers, ",") != "mcp_connections_guard:BEFORE:app.mcp_connections,audit_capture_state:AFTER:app.mcp_connections,"+
		"audit_capture_selection:AFTER:app.mcp_connections" {
		t.Errorf("triggers = %v", triggers)
	}
	var policies []string
	for _, m := range regexp.MustCompile(`CREATE POLICY ([a-z_]+) ON ([a-z_.]+) FOR (SELECT|INSERT|UPDATE) TO %I, %I`).FindAllStringSubmatch(up, -1) {
		policies = append(policies, m[1]+":"+m[2]+":"+m[3])
	}
	if strings.Join(policies, ",") != "mcp_connections_select:app.mcp_connections:SELECT,mcp_connections_insert:app.mcp_connections:INSERT,"+
		"mcp_connections_update:app.mcp_connections:UPDATE" || strings.Count(up, "CREATE POLICY ") != 3 ||
		strings.Count(up, "ALTER POLICY sessions_insert ON app.sessions WITH CHECK") != 1 {
		t.Errorf("policies = %v", policies)
	}
	grants := regexp.MustCompile(`GRANT [^']*TO %I`).FindAllString(up, -1)
	if strings.Join(grants, "|") != "GRANT SELECT (kind) ON app.sessions TO %I|GRANT CREATE ON SCHEMA app TO %I|"+
		"GRANT SELECT ON app.mcp_connections TO %I|GRANT UPDATE (state, client_id, activated_at, selected_group_id, context_version) ON app.mcp_connections TO %I" {
		t.Errorf("grants = %v", grants)
	}
	if !strings.Contains(up, "GRANT INSERT (id, session_id, user_id, organization_id, host, resource, pending_expires_at, created_at) ") {
		t.Error("runtime INSERT columns differ")
	}
	for _, d := range []string{"refusing to roll back 00010: MCP connections or mcp sessions exist",
		"refusing to roll back 00010: audit.events holds MCP history", "DROP TABLE app.mcp_connections;",
		"DROP FUNCTION app.mcp_connections_guard();", "DROP FUNCTION audit.capture_mcp_connections();",
		"ALTER TABLE app.sessions DROP COLUMN kind;", "AND u.status = 'active';\nEND $$\n$e1$",
		"user_id = (SELECT app.ctx_user()) AND (SELECT app.ctx_org()) IS NULL AND revoked_at IS NULL);",
		"ALTER TABLE app.sessions FORCE ROW LEVEL SECURITY;", "ALTER TABLE audit.events FORCE ROW LEVEL SECURITY;"} {
		if !strings.Contains(down, d) {
			t.Errorf("down lacks %q", d)
		}
	}
	// The down section restores the exact version-8 attribution body.
	v8, err := fs.ReadFile(FS, "00008_audit_events.sql")
	if err != nil {
		t.Fatal(err)
	}
	body := func(s string) string {
		i := strings.Index(s, "audit.events_attribute() RETURNS trigger")
		if i < 0 {
			t.Fatal("events_attribute body not found")
		}
		j := strings.Index(s[i:], "END $$;")
		if j < 0 {
			t.Fatal("events_attribute body end not found")
		}
		return s[i : i+j]
	}
	if body(string(v8)) != body(rawDown) {
		t.Error("down does not restore the exact version-8 events_attribute body")
	}
}

// The session-binding migration (G016 security amendment; Plan amendments A1/A2) adds no table, view, sequence, type,
// trigger or policy and no DML: two nullable session columns with VALID shape checks (no unvalidated CHECK, so legacy
// rows stay revocable), one auth-reader column grant, the bound 3-argument reader replacing the 2-argument one (no
// unbound overload), the proven-only hostname_state, the live-hostname index swapped for the proven-hostname index, and
// the sessions_insert predicate that requires a binding on every new browser row. NO FORCE appears only in the down
// section's refusal block, which restores FORCE before any shape change; no trigger or constraint is disabled.
func TestSessionBindingMigrationShape(t *testing.T) {
	b, err := fs.ReadFile(FS, "00012_session_binding_and_claim_exclusivity.sql")
	if err != nil {
		t.Fatal(err)
	}
	rawUp, rawDown, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	strip := regexp.MustCompile(`--[^\n]*`)
	up, down := strip.ReplaceAllString(rawUp, " "), strip.ReplaceAllString(rawDown, " ")
	for _, re := range []string{
		`(?i)\b(CREATE|DROP)\s+(TABLE|VIEW|SEQUENCE|TYPE|EXTENSION|SCHEMA|ROLE|TRIGGER|POLICY)\b`,
		`(?i)\bINSERT\s+INTO\b|\bUPDATE\s+app\.|\bDELETE\s+FROM\b|\bTRUNCATE\b`,
		`(?i)\bNOT\s+VALID\b|\bLEAKPROOF\b|\bBYPASSRLS\b|\bNO\s+FORCE\b|DISABLE\s+TRIGGER|EXCEPTION\s+WHEN`,
		`(?i)\bCREATE\s+OR\s+REPLACE\s+FUNCTION\s+app\.auth_resolve_session\b`,
		`(?i)\bGRANT\b[^;]*\b(INSERT|UPDATE|DELETE|TRUNCATE|REFERENCES|TRIGGER|ALL)\b`, `(?i)\bGRANT\b[^;]*_retention`,
	} {
		if m := regexp.MustCompile(re).FindString(up); m != "" {
			t.Errorf("00012 up contains %q", m)
		}
	}
	for _, m := range []string{
		"ADD COLUMN origin text NULL CONSTRAINT sessions_origin_valid",
		"ADD COLUMN organization_id uuid NULL CONSTRAINT sessions_organization_fkey",
		"ADD CONSTRAINT sessions_binding_shape CHECK ((origin IS NULL) = (organization_id IS NULL))",
		"ADD CONSTRAINT sessions_mcp_unbound CHECK (kind <> 'mcp' OR origin IS NULL)",
		"GRANT SELECT (origin, organization_id) ON app.sessions TO %I",
		"DROP FUNCTION app.auth_resolve_session(bytea, timestamptz);",
		"CREATE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz, p_origin text)",
		"AND u.status = 'active' AND s.kind = 'browser' AND s.origin IS NOT NULL AND s.organization_id IS NOT NULL\n    AND s.origin = p_origin;",
		"REVOKE ALL ON FUNCTION app.auth_resolve_session(bytea, timestamptz, text) FROM PUBLIC;",
		"WHERE d.hostname = (p_key COLLATE \"C\") AND d.state IN ('verified', 'active', 'suspended');",
		"DROP INDEX app.organization_domains_hostname_live;",
		"CREATE UNIQUE INDEX organization_domains_hostname_owned ON app.organization_domains (hostname)\n  WHERE state IN ('verified', 'active', 'suspended');",
		"(kind = ''browser'' AND %s IS NULL AND origin IS NOT NULL '\n    || 'AND organization_id IS NOT NULL) '",
		"(kind = ''mcp'' AND %s IS NOT NULL AND %s IS NULL AND %s IS NOT NULL AND origin IS NULL)",
	} {
		if strings.Count(up, m) != 1 {
			t.Errorf("00012 up lacks (or repeats) %q", m)
		}
	}
	if n := strings.Count(up, "SECURITY DEFINER"); n != 2 {
		t.Errorf("00012 up declares %d SECURITY DEFINER functions, want the two replaced readers", n)
	}
	if strings.Count(up, "ALTER POLICY sessions_insert ON app.sessions WITH CHECK") != 1 {
		t.Error("00012 up must alter the sessions_insert predicate exactly once")
	}
	// Down: the three refusals under NO FORCE, FORCE restored on both tables, then the exact version-11 shapes.
	for _, d := range []string{
		"ALTER TABLE app.sessions NO FORCE ROW LEVEL SECURITY;", "ALTER TABLE app.organization_domains NO FORCE ROW LEVEL SECURITY;",
		"refusing to roll back 00012: bound sessions exist (binding history is never discarded)",
		"refusing to roll back 00012: live legacy browser sessions would become usable by the global reader",
		"refusing to roll back 00012: a hostname has more than one non-removed row",
		"ALTER TABLE app.sessions FORCE ROW LEVEL SECURITY;", "ALTER TABLE app.organization_domains FORCE ROW LEVEL SECURITY;",
		"DROP FUNCTION app.auth_resolve_session(bytea, timestamptz, text);",
		"CREATE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz)",
		"AND u.status = 'active' AND s.kind = 'browser';\nEND $$\n$e1$",
		"user_id = %s AND revoked_at IS NULL AND ((kind = ''browser'' AND %s IS NULL) '\n    || 'OR (kind = ''mcp'' AND %s IS NOT NULL AND %s IS NULL AND %s IS NOT NULL)))'",
		"CREATE UNIQUE INDEX organization_domains_hostname_live ON app.organization_domains (hostname) WHERE state <> 'removed';",
		"DROP COLUMN origin, DROP COLUMN organization_id;",
	} {
		if !strings.Contains(down, d) {
			t.Errorf("down lacks %q", d)
		}
	}
	if regexp.MustCompile(`(?i)DISABLE\s+TRIGGER|\bDROP\s+CONSTRAINT\s+sessions_(origin_valid|organization_fkey)\b|\bDELETE\s+FROM\b|\bUPDATE\s+app\.`).MatchString(down) {
		t.Error("down disables a protection or changes rows")
	}
	if i, j := strings.Index(down, "ALTER TABLE app.organization_domains FORCE ROW LEVEL SECURITY;"), strings.Index(down, "ALTER POLICY"); i < 0 || j < i {
		t.Error("down changes shapes before its refusals restore FORCE")
	}
	// The version-11 bodies are restored exactly: the 2-argument reader as 00010 replaced it, hostname_state as 00011
	// created it.
	body := func(src, marker string) string {
		t.Helper()
		i := strings.Index(src, marker)
		if i < 0 {
			t.Fatalf("%q not found", marker)
		}
		return src[i+strings.Index(src[i:], "AS $$") : i+strings.Index(src[i:], "END $$")]
	}
	for _, c := range []struct{ file, marker string }{
		{"00010_mcp_connections.sql", "app.auth_resolve_session(p_hash bytea, p_at timestamptz)"},
		{"00011_organization_domains.sql", "app.hostname_state(p_kind text, p_key text)"},
	} {
		prev, err := fs.ReadFile(FS, c.file)
		if err != nil {
			t.Fatal(err)
		}
		prevUp, _, _ := strings.Cut(string(prev), "-- +goose Down")
		if body(prevUp, c.marker) != body(rawDown, c.marker) {
			t.Errorf("down does not restore the exact %s body of %s", c.marker, c.file)
		}
	}
}
