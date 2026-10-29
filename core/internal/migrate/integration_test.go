//go:build integration

// Integration tests against a disposable PostgreSQL initialized by the
// rendered role script (infra/postgres/init). They are run by the project
// validation harness (infra/validation/run.sh, suite "migrations") and fail,
// never skip, when the database settings are absent.
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD,
// DB_RUNTIME_PASSWORD, APP_ENV.
package migrate

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/migrations"
)

func itConfig(t *testing.T) Config {
	t.Helper()
	cfg, err := LoadConfig(os.Getenv)
	if err != nil {
		t.Fatalf("integration settings: %v", err)
	}
	if os.Getenv("DB_RUNTIME_PASSWORD") == "" {
		t.Fatal("integration settings: DB_RUNTIME_PASSWORD is required")
	}
	return cfg
}

func connectAs(t *testing.T, cfg Config, user, password string) *pgx.Conn {
	t.Helper()
	cc, err := pgx.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	cc.Host, cc.Port, cc.Database, cc.User, cc.Password = cfg.Host, uint16(cfg.Port), cfg.Database, user, password
	cc.ConnectTimeout = 5 * time.Second
	cc.TLSConfig, cc.Fallbacks = nil, nil
	conn, err := pgx.ConnectConfig(context.Background(), cc)
	if err != nil {
		t.Fatalf("connect as %s: %v", user, err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func newRunner(t *testing.T, cfg Config, combined bool, opts openOptions) *Runner {
	t.Helper()
	var fsys fs.FS = migrations.FS
	if combined {
		fsys = combinedFS(t)
	}
	r, err := open(context.Background(), cfg, fsys, opts)
	if err != nil {
		t.Fatalf("open runner: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

// identityTables are the objects of 00002_identity.sql, in drop order.
var identityTables = []string{"login_transactions", "sessions", "identity_principals", "users", "identity_provider_binding"}

// tenancyDrops remove the objects of 00003_tenancy.sql (G005), in drop order. They run before the identity tables,
// because role_bindings references app.users.
var tenancyDrops = []string{
	"DROP VIEW IF EXISTS app.group_memberships", "DROP VIEW IF EXISTS app.organization_memberships",
	"DROP TABLE IF EXISTS app.role_bindings", "DROP TABLE IF EXISTS app.groups", "DROP TABLE IF EXISTS app.organizations",
	"DROP FUNCTION IF EXISTS app.role_bindings_keys_immutable()", "DROP FUNCTION IF EXISTS app.groups_keys_immutable()",
	"DROP FUNCTION IF EXISTS app.organizations_keys_immutable()",
	// G018: the 00014 bootstrap trigger function outlives its trigger on app.groups.
	"DROP FUNCTION IF EXISTS app.groups_bootstrap_owner()",
}

// storedObjectDrops remove the objects of 00007_stored_objects.sql (G009): the table (with its policies, trigger,
// indexes and grants), then its trigger function.
var storedObjectDrops = []string{"DROP TABLE IF EXISTS app.stored_objects", "DROP FUNCTION IF EXISTS app.stored_objects_guard()"}

// auditFunctions are the eleven functions of 00008_audit_events.sql (G011).
var auditFunctions = []string{"events_attribute", "events_refuse", "capture_organizations", "capture_groups", "capture_role_bindings",
	"capture_business_profiles", "capture_group_business_profiles", "capture_users", "capture_sessions",
	"capture_session_revocations", "capture_stored_objects"}

// auditDrops remove the objects of 00008_audit_events.sql (G011) and the runtime's schema USAGE, leaving the empty,
// owner-only provisioned schema: the table (with its policies, triggers, index and grants, whatever rows it holds; a
// disposable test database only), then the functions, CASCADE dropping the eighteen capture triggers on the app tables.
func auditDrops(cfg Config) []string {
	out := []string{"DROP TABLE IF EXISTS audit.events"}
	for _, f := range auditFunctions {
		out = append(out, "DROP FUNCTION IF EXISTS audit."+f+"() CASCADE")
	}
	return append(out, "DO $$ BEGIN IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'audit') THEN "+
		"EXECUTE 'REVOKE ALL ON SCHEMA audit FROM "+quote(cfg.RuntimeRole())+", "+quote(cfg.RetentionRole())+"'; END IF; END $$")
}

// retentionDrops remove the one object of 00009_retention_maintenance.sql (G013) that outlives its tables: the
// eligibility function. Its four policies and table grants go with audit.events and app.stored_objects, its schema
// USAGE with auditDrops' revoke.
var retentionDrops = []string{"DROP FUNCTION IF EXISTS app.retention_eligible(uuid, timestamptz, uuid[])"}

// mcpDrops remove the objects of 00010_mcp_connections.sql (G014) that outlive audit.events: the capture function
// (CASCADE drops its two triggers), the connection table (with its policies, guard trigger, keys and grants; it
// references sessions, users and organizations) and its guard function. The sessions kind column, the replaced
// sessions_insert policy and the E1 reader go with app.sessions and rlsReset.
var mcpDrops = []string{
	"DROP FUNCTION IF EXISTS audit.capture_mcp_connections() CASCADE", "DROP TABLE IF EXISTS app.mcp_connections",
	"DROP FUNCTION IF EXISTS app.mcp_connections_guard()",
}

// domainDrops remove the objects of 00011_organization_domains.sql (G016) that outlive audit.events and the app
// tables, in dependency order and without CASCADE: the domain table (its policies, guard and capture triggers, keys and
// grants go with it; it references organizations and users), then its owner-owned guard and audit capture functions.
// The slug and login-transaction columns, the organizations/role_bindings policies and grants go with their tables;
// domainLateDrops removes the slug guard after app.organizations (its trigger), and rlsReset the two auth-reader
// functions.
var domainDrops = []string{
	"DROP TABLE IF EXISTS app.organization_domains", "DROP FUNCTION IF EXISTS app.organization_domains_guard()",
	"DROP FUNCTION IF EXISTS audit.capture_organization_domains()",
}

// brandingDrops remove the objects of 00013_organization_branding.sql (G017) that outlive audit.events and the app
// tables, in dependency order and without CASCADE: the auth reader's stored-object policy (it reads the branding table),
// the branding table (its policies, guard and capture triggers, key and grants go with it; it references organizations
// and users), then its owner-owned guard, capture and validator functions. rlsReset drops the auth reader's
// app.public_branding.
var brandingDrops = []string{
	"DO $$ BEGIN IF to_regclass('app.stored_objects') IS NOT NULL THEN " +
		"DROP POLICY IF EXISTS stored_objects_auth_reader ON app.stored_objects; END IF; END $$",
	"DROP TABLE IF EXISTS app.organization_branding", "DROP FUNCTION IF EXISTS app.organization_branding_guard()",
	"DROP FUNCTION IF EXISTS audit.capture_organization_branding()", "DROP FUNCTION IF EXISTS app.branding_overrides_valid(jsonb)",
}

// membershipDrops remove the owner-side objects of 00015_membership_invitations.sql (G019) in the down migration's
// dependency order and without CASCADE (the catalog test leaves version 15, invitations and invitation audit rows after
// the history-preserving down refusal): the authority guard trigger, then its function; the helper policies on
// role_bindings, users and identity_principals (users_claim_* read app.role_bindings and would block its drop); the
// invitation table (its policies, triggers, keys, indexes and grants go with it; it references organizations, groups,
// users and role_bindings, and its insert policy calls the auth reader's is_organization_member); then its owner-owned
// trigger and capture functions and the claim writer's audit USAGE. rlsReset drops the five policy-reader and three
// claim-writer functions as their owners and the claim writer's app USAGE; events_insert_claim and the widened audit
// constraints go with audit.events, role_bindings_update/_delete and the table grants with their tables.
func membershipDrops(cfg Config) []string {
	return []string{
		"DO $$ BEGIN IF to_regclass('app.role_bindings') IS NOT NULL THEN " +
			"DROP TRIGGER IF EXISTS role_bindings_authority_guard ON app.role_bindings; " +
			"DROP POLICY IF EXISTS role_bindings_claim_select ON app.role_bindings; " +
			"DROP POLICY IF EXISTS role_bindings_insert_invited ON app.role_bindings; " +
			"DROP POLICY IF EXISTS role_bindings_claim_upgrade ON app.role_bindings; " +
			"DROP POLICY IF EXISTS role_bindings_scope_reader ON app.role_bindings; " +
			"DROP POLICY IF EXISTS role_bindings_retirement_reader ON app.role_bindings; END IF; END $$",
		"DO $$ BEGIN IF to_regclass('app.users') IS NOT NULL THEN " +
			"DROP POLICY IF EXISTS users_policy_reader ON app.users; DROP POLICY IF EXISTS users_claim_select ON app.users; " +
			"DROP POLICY IF EXISTS users_claim_lock ON app.users; END IF; END $$",
		"DO $$ BEGIN IF to_regclass('app.identity_principals') IS NOT NULL THEN " +
			"DROP POLICY IF EXISTS identity_principals_claim_select ON app.identity_principals; END IF; END $$",
		"DROP FUNCTION IF EXISTS app.role_bindings_authority_guard()",
		"DROP TABLE IF EXISTS app.invitations", "DROP FUNCTION IF EXISTS app.invitations_keys_immutable()",
		"DROP FUNCTION IF EXISTS app.invitations_bound()", "DROP FUNCTION IF EXISTS audit.capture_invitations()",
		"DO $$ BEGIN IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'audit') THEN " +
			"EXECUTE 'REVOKE ALL ON SCHEMA audit FROM " + quote(cfg.Database+"_claim_writer") + "'; END IF; END $$",
	}
}

// enrollmentDrops remove the owner-side objects of 00016_organization_enrollment.sql (G058 slice 1a) in the down
// migration's dependency order and without CASCADE: the departure trigger on app.role_bindings, then its owner-owned
// trigger function; the restriction ledger and the pending enrollment credentials (their policies, keys, indexes,
// identity sequence and grants go with them; they reference organizations and users). The organizations and
// login_transactions columns, constraints, helper policies and grants go with their tables, the widened audit
// constraints and events_attribute with audit.events and auditDrops; rlsReset drops the claim-writer, policy-reader and
// auth-reader functions as their owners (and with them the D-15(i) runtime EXECUTE).
var enrollmentDrops = []string{
	"DO $$ BEGIN IF to_regclass('app.role_bindings') IS NOT NULL THEN " +
		"DROP TRIGGER IF EXISTS role_bindings_enrollment_departure ON app.role_bindings; END IF; END $$",
	"DROP FUNCTION IF EXISTS app.enrollment_restrict_on_departure()",
	// Slice 1b: the two audit captures (CASCADE drops their triggers on app.organizations / app.enrollment_restrictions),
	// the owner-owned readiness functions and the mode/readiness guard (its trigger sits on app.organizations).
	"DROP FUNCTION IF EXISTS audit.capture_organization_enrollment() CASCADE",
	"DROP FUNCTION IF EXISTS audit.capture_enrollment_restrictions() CASCADE",
	"DO $$ BEGIN IF to_regclass('app.organizations') IS NOT NULL THEN " +
		"DROP TRIGGER IF EXISTS organizations_enrollment_guard ON app.organizations; END IF; END $$",
	"DROP FUNCTION IF EXISTS app.organizations_enrollment_guard()",
	"DROP FUNCTION IF EXISTS app.enrollment_readiness_set(uuid, text, text)",
	"DROP FUNCTION IF EXISTS app.enrollment_history_backfill(uuid, uuid, text)",
	// Request slice: the ledger references the requests (source_request_id), so it goes first; the request table takes
	// its decision, state-machine and capture triggers with it, then their three functions go.
	"DROP TABLE IF EXISTS app.enrollment_restrictions", "DROP TABLE IF EXISTS app.enrollment_requests",
	"DROP TABLE IF EXISTS app.enrollment_sessions",
	"DROP FUNCTION IF EXISTS app.enrollment_restrict_on_decision()", "DROP FUNCTION IF EXISTS app.enrollment_requests_immutable()",
	// G060: the 00017 owner-owned override guard (its trigger sits on app.organizations); the column, grants and replaced
	// bodies go with their tables and owners.
	"DO $$ BEGIN IF to_regclass('app.organizations') IS NOT NULL THEN " +
		"DROP TRIGGER IF EXISTS organizations_group_override_guard ON app.organizations; END IF; END $$",
	"DROP FUNCTION IF EXISTS app.organizations_group_override_guard()",
	"DROP FUNCTION IF EXISTS audit.capture_enrollment_requests()",
}

// labelDrops remove the owner-side objects of 00020_labels.sql (G020) without CASCADE: the assignment table, then the
// definition table (their policies, triggers, keys, indexes and grants go with them; they reference organizations and
// groups), then the three owner-owned trigger functions and the two audit captures. The widened audit constraints and
// events_attribute go with audit.events and auditDrops.
var labelDrops = []string{
	"DROP TABLE IF EXISTS app.label_assignments", "DROP TABLE IF EXISTS app.label_definitions",
	"DROP FUNCTION IF EXISTS app.label_definitions_bound()", "DROP FUNCTION IF EXISTS app.label_definitions_keys_immutable()",
	"DROP FUNCTION IF EXISTS app.label_assignments_keys_immutable()", "DROP FUNCTION IF EXISTS audit.capture_label_definitions()",
	"DROP FUNCTION IF EXISTS audit.capture_label_assignments()",
}

// signupDrops remove the owner-side objects of 00018_platform_signup.sql (G062): the signup table (its policies, keys,
// index and grants go with it; it references users and organizations) and the provision writer's audit USAGE. The
// login_transactions columns and constraints, the provision-writer policies on organizations, role_bindings, users and
// audit.events and the widened audit constraints go with their tables; rlsReset drops the three provision-writer bodies
// and the auth reader's app.auth_resolve_signup as their owners, and the provision writer's app USAGE.
func signupDrops(cfg Config) []string {
	return []string{
		"DROP TABLE IF EXISTS app.signup_sessions",
		"DO $$ BEGIN IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'audit') THEN " +
			"EXECUTE 'REVOKE ALL ON SCHEMA audit FROM " + quote(cfg.Database+"_provision_writer") + "'; END IF; END $$",
	}
}

// domainLateDrops remove the 00011 slug guard function once app.organizations (and with it its BEFORE UPDATE trigger)
// is gone.
var domainLateDrops = []string{"DROP FUNCTION IF EXISTS app.organizations_slug_guard()"}

// domainRelations are the 6 relations of 00011_organization_domains.sql (G016): the domain table, its primary key, its
// claim-fence unique key and two indexes, and the partial unique slug index on app.organizations.
var domainRelations = []string{"i:organization_domains_hostname_generation_key", "i:organization_domains_hostname_live",
	"i:organization_domains_organization", "i:organization_domains_pkey", "i:organizations_slug_key", "r:organization_domains"}

// domainRelations12 are the same 6 relations at version 12 (00012_session_binding_and_claim_exclusivity.sql): the
// live-hostname index is replaced by the proven-hostname index organization_domains_hostname_owned.
var domainRelations12 = []string{"i:organization_domains_hostname_generation_key", "i:organization_domains_hostname_owned",
	"i:organization_domains_organization", "i:organization_domains_pkey", "i:organizations_slug_key", "r:organization_domains"}

// mcpRelations are the 4 relations of 00010_mcp_connections.sql (G014): the table, its primary key and its two unique
// keys (no other index).
var mcpRelations = []string{"i:mcp_connections_pkey", "i:mcp_connections_resource_key", "i:mcp_connections_session_key",
	"r:mcp_connections"}

// businessProfileDrops remove the objects of 00004_business_profiles.sql (G006), in drop order. They run before the
// tenancy objects, because the association table references organizations and groups and the view reads role_bindings.
var businessProfileDrops = []string{
	"DROP VIEW IF EXISTS app.business_profile_access_paths", "DROP TABLE IF EXISTS app.group_business_profiles",
	"DROP TABLE IF EXISTS app.business_profiles", "DROP FUNCTION IF EXISTS app.group_business_profiles_keys_immutable()",
	"DROP FUNCTION IF EXISTS app.business_profiles_keys_immutable()",
}

// reset returns the database to "fresh": no migration metadata, no probe, no row-level security objects, no
// BusinessProfile or tenancy objects, no identity tables. 00005_organization_policy.sql (G007) adds no object of its
// own: its policy column is dropped with app.organizations. 00006_row_level_security.sql (G008) adds functions (the
// helper-owned readers are dropped as their helper, through the session's SET chain, in one transaction), policies
// (dropped with their tables) and the helpers' schema USAGE.
func reset(t *testing.T, cfg Config) {
	t.Helper()
	r := newRunner(t, cfg, false, openOptions{})
	// G011: the audit objects go first (their capture triggers sit on the app tables; their functions and policies use
	// the 00006 parsers and readers). G009: the stored-object relation references organizations, groups and users, and
	// its policies use the 00006 parsers: it goes next.
	// G013: the retention function goes after both tables (their retention policies use it). G014: the MCP capture
	// function, connection table and guard go before the identity and tenancy tables they reference. G016: the domain
	// table and its functions go first (the table references organizations and users).
	// G016 security amendment: 00012's session binding references organizations, so its foreign key goes first (the
	// columns go with app.sessions). G017: the branding table and its functions go before the domain objects. G019: the
	// 00015 owner-side objects go first (the invitation table references the tenancy tables and role_bindings; its capture
	// trigger and the helper policies sit on app tables). G058: the 00016 owner-side objects go before them (the
	// departure trigger sits on app.role_bindings; the two tables reference organizations and users).
	// G062: the 00018 signup table goes first (it references organizations and users). G020: the 00020 label tables and
	// functions go before it (the tables reference organizations and groups).
	stmts0 := slices.Concat(labelDrops, signupDrops(cfg), enrollmentDrops, membershipDrops(cfg),
		[]string{"ALTER TABLE IF EXISTS app.sessions DROP CONSTRAINT IF EXISTS sessions_organization_fkey"},
		brandingDrops, domainDrops, mcpDrops, auditDrops(cfg), storedObjectDrops, retentionDrops)
	for _, stmt := range stmts0 {
		if _, err := r.db.ExecContext(context.Background(), stmt); err != nil {
			t.Fatalf("reset: %v", err)
		}
	}
	rlsReset(t, cfg, r)
	stmts := []string{"DROP TABLE IF EXISTS app.synthetic_ddl_probe"}
	stmts = append(stmts, businessProfileDrops...)
	stmts = append(stmts, tenancyDrops...)
	stmts = append(stmts, domainLateDrops...)
	for _, tbl := range identityTables {
		stmts = append(stmts, "DROP TABLE IF EXISTS app."+tbl)
	}
	stmts = append(stmts, "DROP TABLE IF EXISTS app.goose_db_version")
	for _, stmt := range stmts {
		if _, err := r.db.ExecContext(context.Background(), stmt); err != nil {
			t.Fatalf("reset: %v", err)
		}
	}
	_ = r.Close()
}

// rlsReset removes the objects of 00006_row_level_security.sql, if present, in one transaction of r's owner session.
func rlsReset(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	pr, ar, ow := quote(cfg.Database+"_policy_reader"), quote(cfg.Database+"_auth_reader"), quote(cfg.OwnerRole())
	cw := quote(cfg.Database + "_claim_writer")
	pw := quote(cfg.Database + "_provision_writer")
	for _, stmt := range []string{
		// A helper needs USAGE on app to find (and drop) its own functions; an earlier reset revoked it.
		"GRANT USAGE ON SCHEMA app TO " + pr + ", " + ar + ", " + cw + ", " + pw,
		// G062: the three 00018 provision-writer bodies (no object depends on them).
		"SET LOCAL ROLE " + pw,
		"DROP FUNCTION IF EXISTS app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz)",
		// G057: the 00021 8-argument provisioning body that replaces it.
		"DROP FUNCTION IF EXISTS app.signup_provision_organization(bytea, text, text, text, bytea, bytea, text, timestamptz)",
		"DROP FUNCTION IF EXISTS app.signup_challenge_issue(bytea, text, uuid, bytea, bytea, timestamptz)",
		"DROP FUNCTION IF EXISTS app.signup_challenge_complete(bytea, bytea, text, uuid, timestamptz)",
		"SET LOCAL ROLE " + ow,
		// G019: the three 00015 claim-writer bodies (no object depends on them once membershipDrops ran).
		"SET LOCAL ROLE " + cw,
		// G058: the four 00016 claim-writer bodies (no object depends on them once enrollmentDrops ran).
		"DROP FUNCTION IF EXISTS app.enrollment_request(bytea, uuid)",
		"DROP FUNCTION IF EXISTS app.enrollment_admit(bytea, uuid, text, uuid)",
		"DROP FUNCTION IF EXISTS app.enrollment_lock_organization(uuid)",
		"DROP FUNCTION IF EXISTS app.rls_invitation_pending(uuid, uuid)",
		"DROP FUNCTION IF EXISTS app.invitation_claim_organization(uuid, bytea, text, text, text)",
		"DROP FUNCTION IF EXISTS app.invitation_claim_group(uuid, bytea)",
		"DROP FUNCTION IF EXISTS app.rls_user_enabled_lock(uuid, text, uuid)",
		"SET LOCAL ROLE " + pr,
		// G058: the 00016 public mode reader (no object depends on it).
		"DROP FUNCTION IF EXISTS app.rls_enrollment_mode(uuid)",
		// G019: the five 00015 policy-reader readers (no object depends on them once membershipDrops ran).
		"DROP FUNCTION IF EXISTS app.rls_group_invitation_candidates(uuid, uuid, integer)",
		"DROP FUNCTION IF EXISTS app.rls_own_organizations(uuid, integer)", "DROP FUNCTION IF EXISTS app.rls_sole_owner_scopes()",
		"DROP FUNCTION IF EXISTS app.rls_dependent_group_bindings(uuid)", "DROP FUNCTION IF EXISTS app.rls_owner_count(text, uuid)",
		// G018: the 00014 profile eligibility reader (CASCADE: the association insert policy that calls it).
		"DROP FUNCTION IF EXISTS app.rls_profile_eligible(uuid) CASCADE",
		"DROP FUNCTION IF EXISTS app.rls_group_role(uuid) CASCADE", "DROP FUNCTION IF EXISTS app.rls_org_role() CASCADE",
		"SET LOCAL ROLE " + ar,
		// G058: the 00016 enrollment bootstrap resolver (no object depends on it).
		"DROP FUNCTION IF EXISTS app.auth_resolve_enrollment(bytea, timestamptz)",
		// G062: the 00018 signup bootstrap resolver (no object depends on it).
		"DROP FUNCTION IF EXISTS app.auth_resolve_signup(bytea, timestamptz)",
		// G057: the 00021 slug-availability reader (no object depends on it).
		"DROP FUNCTION IF EXISTS app.signup_slug_taken(text)",
		// G071: the 00022 default-tenant reader (no object depends on it).
		"DROP FUNCTION IF EXISTS app.organization_id_by_slug(text)",
		// G016: the two 00011 auth-reader functions (no object depends on them).
		"DROP FUNCTION IF EXISTS app.hostname_state(text, text)", "DROP FUNCTION IF EXISTS app.is_organization_member(uuid, uuid)",
		"DROP FUNCTION IF EXISTS app.auth_logout_ref(bytea, timestamptz) CASCADE", "DROP FUNCTION IF EXISTS app.auth_principal_subject(text, text) CASCADE",
		"DROP FUNCTION IF EXISTS app.auth_resolve_session(bytea, timestamptz) CASCADE",
		// G016 security amendment: the 00012 bound reader replaces the 2-argument one.
		"DROP FUNCTION IF EXISTS app.auth_resolve_session(bytea, timestamptz, text) CASCADE",
		// G017: the 00013 public branding projection (no object depends on it).
		"DROP FUNCTION IF EXISTS app.public_branding(uuid)",
		"SET LOCAL ROLE " + ow,
		"DROP FUNCTION IF EXISTS app.ctx_login_binding() CASCADE", "DROP FUNCTION IF EXISTS app.ctx_group() CASCADE",
		"DROP FUNCTION IF EXISTS app.ctx_org() CASCADE", "DROP FUNCTION IF EXISTS app.ctx_user() CASCADE",
		"REVOKE ALL ON SCHEMA app FROM " + pr + ", " + ar + ", " + cw + ", " + pw,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("reset (%s): %v", stmt, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("reset: %v", err)
	}
}

func metadataExists(t *testing.T, rt *pgx.Conn) bool {
	t.Helper()
	var exists bool
	if err := rt.QueryRow(context.Background(), `SELECT to_regclass('app.goose_db_version') IS NOT NULL`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	return exists
}

func expectDenied(t *testing.T, conn *pgx.Conn, what, stmt string) {
	t.Helper()
	_, err := conn.Exec(context.Background(), stmt)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
		t.Errorf("%s: want SQLSTATE 42501, got %v", what, err)
	}
}

// assertP is assertion P (plan §8.5): the metadata table and its identity
// sequence exist, are owned by the owner role, have owner-only ACLs, and no
// other role (runtime, retention, migrator without SET ROLE, PUBLIC) holds any
// privilege on them; real runtime attempts are denied.
func assertP(t *testing.T, cfg Config, r *Runner, rt *pgx.Conn) {
	t.Helper()
	ctx := context.Background()
	rows, err := r.db.QueryContext(ctx, `
SELECT c.relkind::text, c.relowner::regrole::text,
       (SELECT count(*) FROM aclexplode(coalesce(c.relacl, acldefault(
          CASE WHEN c.relkind = 'S' THEN 's'::"char" ELSE 'r'::"char" END, c.relowner))) a
        WHERE a.grantee <> c.relowner)
FROM pg_class c
WHERE c.oid IN ('app.goose_db_version'::regclass,
                pg_get_serial_sequence('app.goose_db_version', 'id')::regclass)
ORDER BY c.relkind`)
	if err != nil {
		t.Fatalf("P: catalog: %v", err)
	}
	var kinds []string
	for rows.Next() {
		var kind, owner string
		var extra int
		if err := rows.Scan(&kind, &owner, &extra); err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, kind)
		if owner != cfg.OwnerRole() || extra != 0 {
			t.Errorf("P: relkind %s owned by %s with %d non-owner ACL entries", kind, owner, extra)
		}
	}
	rows.Close()
	if strings.Join(kinds, ",") != "S,r" {
		t.Fatalf("P: metadata objects = %v, want sequence and table", kinds)
	}
	var seq string
	if err := r.db.QueryRowContext(ctx, `SELECT pg_get_serial_sequence('app.goose_db_version', 'id')`).Scan(&seq); err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{cfg.RuntimeRole(), cfg.RetentionRole(), cfg.MigratorRole(), "public"} {
		var any bool
		err := r.db.QueryRowContext(ctx, `
SELECT bool_or(has_table_privilege($1, 'app.goose_db_version', p))
       OR bool_or(has_sequence_privilege($1, $2, s))
FROM unnest(ARRAY['SELECT','INSERT','UPDATE','DELETE','TRUNCATE','REFERENCES','TRIGGER']) p,
     unnest(ARRAY['USAGE','SELECT','UPDATE']) s`, role, seq).Scan(&any)
		if err != nil {
			t.Fatal(err)
		}
		if any {
			t.Errorf("P: %s holds a privilege on migration metadata", role)
		}
	}
	expectDenied(t, rt, "runtime SELECT metadata", `SELECT 1 FROM app.goose_db_version`)
	expectDenied(t, rt, "runtime nextval metadata sequence", `SELECT nextval('`+strings.ReplaceAll(seq, "'", "''")+`')`)
}

func snapshot(t *testing.T, r *Runner) string {
	t.Helper()
	var objects, history string
	err := r.db.QueryRowContext(context.Background(), `
SELECT coalesce(string_agg(format('%s.%s|%s|%s|%s', n.nspname, c.relname, c.relkind, c.relowner::regrole, coalesce(c.relacl::text, '')),
                           E'\n' ORDER BY n.nspname, c.relname), '')
FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname IN ('app', 'audit', 'public')`).Scan(&objects)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.db.QueryRowContext(context.Background(), `
SELECT string_agg(format('%s:%s', version_id, is_applied), ',' ORDER BY version_id) FROM app.goose_db_version`).Scan(&history); err != nil {
		t.Fatal(err)
	}
	return objects + "\nhistory " + history
}

// onlyMetadataAndNothingInPublic: app holds only the metadata table, its
// sequence and its indexes; public holds nothing owned by an application role.
func onlyMetadataAndNothingInPublic(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	var others, public int
	err := r.db.QueryRowContext(context.Background(), `
SELECT
  (SELECT count(*) FROM pg_class c
    WHERE c.relnamespace = 'app'::regnamespace
      AND c.oid NOT IN ('app.goose_db_version'::regclass, pg_get_serial_sequence('app.goose_db_version', 'id')::regclass)
      AND NOT EXISTS (SELECT 1 FROM pg_index i WHERE i.indexrelid = c.oid AND i.indrelid = 'app.goose_db_version'::regclass)),
  (SELECT count(*) FROM pg_class c
    WHERE c.relnamespace = 'public'::regnamespace AND c.relowner::regrole::text = ANY($1))`,
		[]string{cfg.OwnerRole(), cfg.MigratorRole(), cfg.RuntimeRole(), cfg.RetentionRole()}).Scan(&others, &public)
	if err != nil {
		t.Fatal(err)
	}
	if others != 0 || public != 0 {
		t.Errorf("schema objects beyond metadata: app=%d, public(owned by application roles)=%d", others, public)
	}
}

// metadataAndIdentityOnly: at version 2 (and after 00003 is rolled back), app holds the metadata (table, sequence, indexes) and exactly the five
// identity tables of 00002 with their own indexes, all owned by the owner role; public holds nothing owned by an
// application role. The detailed identity catalog (constraints, grants, negatives) is B-IDENTITY-CATALOG.
func metadataAndIdentityOnly(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	var tables, strays, wrongOwner, public string
	err := r.db.QueryRowContext(context.Background(), `
WITH meta AS (
  SELECT 'app.goose_db_version'::regclass AS oid
  UNION ALL SELECT pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
  UNION ALL SELECT i.indexrelid::regclass FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass
), objs AS (
  SELECT c.oid, c.relname, c.relkind, c.relowner, t.relname AS table_name, t.relnamespace AS table_ns
  FROM pg_class c
  LEFT JOIN pg_index i ON i.indexrelid = c.oid
  LEFT JOIN pg_class t ON t.oid = i.indrelid
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid NOT IN (SELECT oid FROM meta)
)
SELECT
  coalesce((SELECT string_agg(relname, ',' ORDER BY relname) FROM objs WHERE relkind = 'r'), ''),
  coalesce((SELECT string_agg(relname, ',' ORDER BY relname) FROM objs
            WHERE relkind NOT IN ('r', 'i')
               OR (relkind = 'i' AND (table_ns <> 'app'::regnamespace OR table_name <> ALL($2::text[])))), ''),
  coalesce((SELECT string_agg(relname, ',' ORDER BY relname) FROM objs WHERE relowner::regrole::text <> $1), ''),
  coalesce((SELECT string_agg(c.relname, ',') FROM pg_class c
            WHERE c.relnamespace = 'public'::regnamespace AND c.relowner::regrole::text = ANY($3)), '')`,
		cfg.OwnerRole(), identityTables,
		[]string{cfg.OwnerRole(), cfg.MigratorRole(), cfg.RuntimeRole(), cfg.RetentionRole()}).Scan(&tables, &strays, &wrongOwner, &public)
	if err != nil {
		t.Fatal(err)
	}
	if tables != "identity_principals,identity_provider_binding,login_transactions,sessions,users" || strays != "" || wrongOwner != "" || public != "" {
		t.Errorf("objects at version 2: tables=%q strays=%q not-owner=%q public=%q", tables, strays, wrongOwner, public)
	}
}

// tenancyRelations are the 13 relations of 00003_tenancy.sql (relkind:name): three tables, two views, eight indexes.
var tenancyRelations = []string{
	"i:groups_organization_id_key", "i:groups_organization_slug_key", "i:groups_pkey", "i:organizations_pkey",
	"i:role_bindings_group_idx", "i:role_bindings_pkey", "i:role_bindings_scope_idx", "i:role_bindings_subject_scope_key",
	"r:groups", "r:organizations", "r:role_bindings", "v:group_memberships", "v:organization_memberships",
}

// metadataIdentityAndTenancyOnly: at version 3, app holds the metadata (table, sequence, indexes), exactly the 14
// identity relations of 00002 and exactly the 13 tenancy relations of 00003 (G005), all owned by the owner role, plus
// exactly the three tenancy trigger functions; public holds nothing owned by an application role. The detailed
// catalogs are B-IDENTITY-CATALOG and B-TENANCY-CATALOG.
func metadataIdentityAndTenancyOnly(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	var objs, wrongOwner, funcs, public string
	err := r.db.QueryRowContext(context.Background(), `
WITH meta AS (
  SELECT 'app.goose_db_version'::regclass AS oid
  UNION ALL SELECT pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
  UNION ALL SELECT i.indexrelid::regclass FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass
), objs AS (
  SELECT c.relname, c.relkind, c.relowner FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid NOT IN (SELECT oid FROM meta)
)
SELECT
  coalesce((SELECT string_agg(relkind::text || ':' || relname, ',' ORDER BY relkind::text, relname) FROM objs), ''),
  coalesce((SELECT string_agg(relname, ',' ORDER BY relname) FROM objs WHERE relowner::regrole::text <> $1), ''),
  coalesce((SELECT string_agg(proname || ':' || proowner::regrole::text, ',' ORDER BY proname) FROM pg_proc
            WHERE pronamespace = 'app'::regnamespace), ''),
  coalesce((SELECT string_agg(c.relname, ',') FROM pg_class c
            WHERE c.relnamespace = 'public'::regnamespace AND c.relowner::regrole::text = ANY($2)), '')`,
		cfg.OwnerRole(), []string{cfg.OwnerRole(), cfg.MigratorRole(), cfg.RuntimeRole(), cfg.RetentionRole()}).Scan(&objs, &wrongOwner, &funcs, &public)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]string{
		"i:identity_principals_pkey", "i:identity_principals_user_id_key", "i:identity_provider_binding_pkey",
		"i:login_transactions_pkey", "i:sessions_credential_hash_key", "i:sessions_pkey",
		"i:sessions_provider_logout_ref_hash_key", "i:sessions_user_live_idx", "i:users_pkey", "r:identity_principals",
		"r:identity_provider_binding", "r:login_transactions", "r:sessions", "r:users",
	}, tenancyRelations...)
	slices.Sort(want)
	o := cfg.OwnerRole()
	wantFuncs := "groups_keys_immutable:" + o + ",organizations_keys_immutable:" + o + ",role_bindings_keys_immutable:" + o
	if objs != strings.Join(want, ",") || wrongOwner != "" || funcs != wantFuncs || public != "" {
		t.Errorf("objects at version 3: objects=%q not-owner=%q functions=%q public=%q", objs, wrongOwner, funcs, public)
	}
}

// businessProfileRelations are the 8 relations of 00004_business_profiles.sql (relkind:name): two tables, one view and
// five indexes.
var businessProfileRelations = []string{
	"i:business_profiles_organization_id_key", "i:business_profiles_pkey", "i:group_business_profiles_business_profile_idx",
	"i:group_business_profiles_group_key", "i:group_business_profiles_pkey", "r:business_profiles", "r:group_business_profiles",
	"v:business_profile_access_paths",
}

// metadataIdentityTenancyAndBusinessProfileOnly: at version 4 (and at version 5, whose G007 policy flag is a column of
// app.organizations, not a relation), app holds the metadata (table, sequence, indexes), the 14
// identity relations of 00002, the 13 tenancy relations of 00003 (G005) and the 8 BusinessProfile relations of 00004
// (G006), all owned by the owner role, plus exactly the five trigger functions of 00003 and 00004; public holds nothing
// owned by an application role. The detailed catalogs are B-IDENTITY-CATALOG, B-TENANCY-CATALOG and
// B-BUSINESS-PROFILE-CATALOG.
func metadataIdentityTenancyAndBusinessProfileOnly(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o := cfg.OwnerRole()
	metadataRelationsOnly(t, cfg, r, "4/5", "business_profiles_keys_immutable:"+o+",group_business_profiles_keys_immutable:"+o+
		",groups_keys_immutable:"+o+",organizations_keys_immutable:"+o+",role_bindings_keys_immutable:"+o)
}

// metadataThroughVersion6Only is the version-6 form (G008): the same 35 relations, plus exactly the four owner-owned
// context parsers and the five readers owned by their helper roles.
func metadataThroughVersion6Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	metadataRelationsOnly(t, cfg, r, "6", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+",ctx_org:"+o+",ctx_user:"+o+
		",group_business_profiles_keys_immutable:"+o+",groups_keys_immutable:"+o+",organizations_keys_immutable:"+o+
		",rls_group_role:"+pr+",rls_org_role:"+pr+",role_bindings_keys_immutable:"+o)
}

// storedObjectRelations are the 3 relations of 00007_stored_objects.sql (G009): the table, its primary key and the
// Organization/Group index.
var storedObjectRelations = []string{"i:stored_objects_organization_group_idx", "i:stored_objects_pkey", "r:stored_objects"}

// metadataThroughVersion7Only is the version-7 form (G009): the version-6 relations and functions plus the stored-object
// table, its two indexes and its owner-owned invoker trigger function (38 relations). B-OBJ-CATALOG names the rest.
func metadataThroughVersion7Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	metadataRelationsOnly(t, cfg, r, "7", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+",ctx_org:"+o+",ctx_user:"+o+
		",group_business_profiles_keys_immutable:"+o+",groups_keys_immutable:"+o+",organizations_keys_immutable:"+o+
		",rls_group_role:"+pr+",rls_org_role:"+pr+",role_bindings_keys_immutable:"+o+",stored_objects_guard:"+o, storedObjectRelations...)
	auditObjects(t, r, "")
}

// metadataThroughVersion8Only is the version-8 form (G011): the version-7 app relations and functions, unchanged (the
// audit objects live outside schema app), plus exactly the audit relation, its two indexes and its eleven functions.
// B-AUDIT-CATALOG names the rest.
func metadataThroughVersion8Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	metadataRelationsOnly(t, cfg, r, "8", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+",ctx_org:"+o+",ctx_user:"+o+
		",group_business_profiles_keys_immutable:"+o+",groups_keys_immutable:"+o+",organizations_keys_immutable:"+o+
		",rls_group_role:"+pr+",rls_org_role:"+pr+",role_bindings_keys_immutable:"+o+",stored_objects_guard:"+o, storedObjectRelations...)
	fns := slices.Clone(auditFunctions)
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// metadataThroughVersion9Only is the version-9 form (G013): the version-8 relations, unchanged, plus the owner-owned
// retention eligibility function in schema app. B-RETENTION-CATALOG names the rest.
func metadataThroughVersion9Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	metadataRelationsOnly(t, cfg, r, "9", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+",ctx_org:"+o+",ctx_user:"+o+
		",group_business_profiles_keys_immutable:"+o+",groups_keys_immutable:"+o+",organizations_keys_immutable:"+o+
		",retention_eligible:"+o+",rls_group_role:"+pr+",rls_org_role:"+pr+",role_bindings_keys_immutable:"+o+
		",stored_objects_guard:"+o, storedObjectRelations...)
	fns := slices.Clone(auditFunctions)
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// metadataThroughVersion10Only is the version-10 form (G014): the version-9 relations plus the connection table and its
// three keys, the owner-owned connection guard in schema app, and the audit capture function in schema audit.
// B-MCP-CATALOG names the rest.
func metadataThroughVersion10Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	metadataRelationsOnly(t, cfg, r, "10", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+",ctx_org:"+o+",ctx_user:"+o+
		",group_business_profiles_keys_immutable:"+o+",groups_keys_immutable:"+o+",mcp_connections_guard:"+o+
		",organizations_keys_immutable:"+o+",retention_eligible:"+o+",rls_group_role:"+pr+",rls_org_role:"+pr+
		",role_bindings_keys_immutable:"+o+",stored_objects_guard:"+o, append(slices.Clone(storedObjectRelations), mcpRelations...)...)
	fns := append(slices.Clone(auditFunctions), "capture_mcp_connections")
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// metadataThroughVersion11Only is the version-11 form (G016), a separate counterpart that leaves the version-10 form
// unchanged: the version-10 relations plus the domain table, its three keys/indexes and the slug index, the two
// auth-reader functions (hostname_state, is_organization_member), the owner-owned domain and slug guards in schema app,
// and the domain capture function in schema audit. TestDomainsCatalog names the rest.
func metadataThroughVersion11Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	extra := slices.Concat(storedObjectRelations, mcpRelations, domainRelations)
	metadataRelationsOnly(t, cfg, r, "11", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+",ctx_org:"+o+",ctx_user:"+o+
		",group_business_profiles_keys_immutable:"+o+",groups_keys_immutable:"+o+",hostname_state:"+ar+
		",is_organization_member:"+ar+",mcp_connections_guard:"+o+",organization_domains_guard:"+o+
		",organizations_keys_immutable:"+o+",organizations_slug_guard:"+o+",retention_eligible:"+o+",rls_group_role:"+pr+
		",rls_org_role:"+pr+",role_bindings_keys_immutable:"+o+",stored_objects_guard:"+o, extra...)
	fns := append(slices.Clone(auditFunctions), "capture_mcp_connections", "capture_organization_domains")
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// metadataThroughVersion12Only is the version-12 form (G016 security amendment): the version-11 objects with the
// live-hostname index replaced by organization_domains_hostname_owned; the function names are unchanged (the bound
// session reader replaces the 2-argument one under the same name). TestDomainsCatalog CT-DOM-12 names the rest.
func metadataThroughVersion12Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	extra := slices.Concat(storedObjectRelations, mcpRelations, domainRelations12)
	metadataRelationsOnly(t, cfg, r, "12", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+",ctx_org:"+o+",ctx_user:"+o+
		",group_business_profiles_keys_immutable:"+o+",groups_keys_immutable:"+o+",hostname_state:"+ar+
		",is_organization_member:"+ar+",mcp_connections_guard:"+o+",organization_domains_guard:"+o+
		",organizations_keys_immutable:"+o+",organizations_slug_guard:"+o+",retention_eligible:"+o+",rls_group_role:"+pr+
		",rls_org_role:"+pr+",role_bindings_keys_immutable:"+o+",stored_objects_guard:"+o, extra...)
	fns := append(slices.Clone(auditFunctions), "capture_mcp_connections", "capture_organization_domains")
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// brandingRelations are the 2 relations of 00013_organization_branding.sql (G017): the branding table and its primary
// key (no other index).
var brandingRelations = []string{"i:organization_branding_pkey", "r:organization_branding"}

// metadataThroughVersion13Only is the version-13 form (G017), a separate counterpart that leaves the version-12 form
// unchanged: the version-12 relations plus the branding table and its key, the owner-owned override validator and
// branding guard and the auth reader's public_branding in schema app, and the branding capture function in schema
// audit. TestBrandingCatalog names the rest.
func metadataThroughVersion13Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	extra := slices.Concat(storedObjectRelations, mcpRelations, domainRelations12, brandingRelations)
	metadataRelationsOnly(t, cfg, r, "13", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",branding_overrides_valid:"+o+",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+
		",ctx_org:"+o+",ctx_user:"+o+",group_business_profiles_keys_immutable:"+o+",groups_keys_immutable:"+o+
		",hostname_state:"+ar+",is_organization_member:"+ar+",mcp_connections_guard:"+o+",organization_branding_guard:"+o+
		",organization_domains_guard:"+o+",organizations_keys_immutable:"+o+",organizations_slug_guard:"+o+
		",public_branding:"+ar+",retention_eligible:"+o+",rls_group_role:"+pr+",rls_org_role:"+pr+
		",role_bindings_keys_immutable:"+o+",stored_objects_guard:"+o, extra...)
	fns := append(slices.Clone(auditFunctions), "capture_mcp_connections", "capture_organization_branding", "capture_organization_domains")
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// metadataThroughVersion14Only is the version-14 form (G018), a separate counterpart that leaves the version-13 form
// unchanged: no new relation; the owner-owned SECURITY INVOKER bootstrap trigger function groups_bootstrap_owner and the
// policy reader's rls_profile_eligible in schema app. TestOnboardingCatalog names the rest.
func metadataThroughVersion14Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader"
	extra := slices.Concat(storedObjectRelations, mcpRelations, domainRelations12, brandingRelations)
	metadataRelationsOnly(t, cfg, r, "14", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",branding_overrides_valid:"+o+",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+
		",ctx_org:"+o+",ctx_user:"+o+",group_business_profiles_keys_immutable:"+o+",groups_bootstrap_owner:"+o+
		",groups_keys_immutable:"+o+",hostname_state:"+ar+",is_organization_member:"+ar+",mcp_connections_guard:"+o+
		",organization_branding_guard:"+o+",organization_domains_guard:"+o+",organizations_keys_immutable:"+o+
		",organizations_slug_guard:"+o+",public_branding:"+ar+",retention_eligible:"+o+",rls_group_role:"+pr+
		",rls_org_role:"+pr+",rls_profile_eligible:"+pr+",role_bindings_keys_immutable:"+o+",stored_objects_guard:"+o, extra...)
	fns := append(slices.Clone(auditFunctions), "capture_mcp_connections", "capture_organization_branding", "capture_organization_domains")
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// membershipRelations are the 6 relations of 00015_membership_invitations.sql (G019): the invitation table, its primary
// and token-digest keys and its scope, pending and recipient indexes.
var membershipRelations = []string{"i:invitations_pending_idx", "i:invitations_pkey", "i:invitations_recipient_idx",
	"i:invitations_scope_idx", "i:invitations_token_digest_key", "r:invitations"}

// metadataThroughVersion15Only is the version-15 form (G019), a separate counterpart that leaves the version-14 form
// unchanged: the version-14 relations plus the invitation relations; in schema app the owner-owned invitation triggers
// and authority guard, the policy reader's five membership readers and the claim writer's three bodies; in schema audit
// the invitation capture function. TestMembershipCatalog names the rest.
func metadataThroughVersion15Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	o, pr, ar, cw := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader", cfg.Database+"_claim_writer"
	extra := slices.Concat(storedObjectRelations, mcpRelations, domainRelations12, brandingRelations, membershipRelations)
	metadataRelationsOnly(t, cfg, r, "15", "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_session:"+ar+
		",branding_overrides_valid:"+o+",business_profiles_keys_immutable:"+o+",ctx_group:"+o+",ctx_login_binding:"+o+
		",ctx_org:"+o+",ctx_user:"+o+",group_business_profiles_keys_immutable:"+o+",groups_bootstrap_owner:"+o+
		",groups_keys_immutable:"+o+",hostname_state:"+ar+",invitation_claim_group:"+cw+",invitation_claim_organization:"+cw+
		",invitations_bound:"+o+",invitations_keys_immutable:"+o+",is_organization_member:"+ar+",mcp_connections_guard:"+o+
		",organization_branding_guard:"+o+",organization_domains_guard:"+o+",organizations_keys_immutable:"+o+
		",organizations_slug_guard:"+o+",public_branding:"+ar+",retention_eligible:"+o+
		",rls_dependent_group_bindings:"+pr+",rls_group_invitation_candidates:"+pr+",rls_group_role:"+pr+",rls_org_role:"+pr+
		",rls_own_organizations:"+pr+",rls_owner_count:"+pr+",rls_profile_eligible:"+pr+",rls_sole_owner_scopes:"+pr+
		",rls_user_enabled_lock:"+cw+",role_bindings_authority_guard:"+o+",role_bindings_keys_immutable:"+o+
		",stored_objects_guard:"+o, extra...)
	fns := append(slices.Clone(auditFunctions), "capture_invitations", "capture_mcp_connections", "capture_organization_branding",
		"capture_organization_domains")
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// enrollmentRelations are the 15 relations of 00016_organization_enrollment.sql (G058 slice 1a and the request slice):
// the three tables, their primary keys, the credential-digest and seq unique keys, the three 1a partial indexes, the
// request table's one_pending unique partial index and its scope and user indexes, and the ledger's identity sequence.
var enrollmentRelations = []string{"S:enrollment_restrictions_seq_seq", "i:enrollment_requests_one_pending",
	"i:enrollment_requests_pkey", "i:enrollment_requests_scope_idx", "i:enrollment_requests_user_idx",
	"i:enrollment_restrictions_open_idx",
	"i:enrollment_restrictions_pkey", "i:enrollment_restrictions_seq_key", "i:enrollment_sessions_credential_hash_key",
	"i:enrollment_sessions_pkey", "i:enrollment_sessions_provider_idx", "i:enrollment_sessions_user_idx",
	"r:enrollment_requests", "r:enrollment_restrictions", "r:enrollment_sessions"}

// metadataThroughVersion16Only is the version-16 form (G058 slice 1a), a separate counterpart that leaves the version-15
// form unchanged: the version-15 relations plus the enrollment relations; in schema app the policy reader's mode reader,
// the auth reader's enrollment resolver, the claim writer's four bodies, the owner-owned departure and decision trigger
// bodies and the request state machine; in schema audit the enrollment captures (events_attribute is replaced in
// place). TestEnrollmentCatalog names the rest.
func metadataThroughVersion16Only(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	metadataThroughVersion(t, cfg, r, "16")
}

// metadataThroughVersion17 is the version-17 form (G060): no relation is added (71); the app functions gain exactly the
// owner-owned override guard; schema audit is unchanged.
func metadataThroughVersion17(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	metadataThroughVersion(t, cfg, r, "17")
}

// signupRelations are the 8 relations of 00018_platform_signup.sql (G062): the signup table, its primary key, its five
// digest unique keys and the live-user index.
var signupRelations = []string{"i:signup_sessions_challenge_hash_key", "i:signup_sessions_credential_hash_key",
	"i:signup_sessions_nonce_hash_key", "i:signup_sessions_pkey", "i:signup_sessions_proof_hash_key",
	"i:signup_sessions_ref_hash_key", "i:signup_sessions_user_idx", "r:signup_sessions"}

// metadataThroughVersion18 is the version-18 form (G062): the version-17 objects plus the signup relations (79); in
// schema app the auth reader's signup resolver and the provision writer's three bodies; schema audit is unchanged
// (events_attribute is replaced in place). TestSignupCatalog names the rest.
func metadataThroughVersion18(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	metadataThroughVersion(t, cfg, r, "18")
}

// metadataThroughVersion19 is the version-19 form (G063): exactly the version-18 objects (79). 00019 adds one column, its
// CHECK, two column grants and one trigger on app.mcp_connections and replaces two audit bodies in schema audit; it adds
// no relation and no app function. TestMCPNamesCatalog names the rest.
func metadataThroughVersion19(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	metadataThroughVersion(t, cfg, r, "19")
}

// labelRelations are the 9 relations of 00020_labels.sql (G020): the two label tables, their primary keys, the
// definitions' scoped-name and owner unique keys, the assignments' unique key and its target and Group indexes.
var labelRelations = []string{"i:label_assignments_group_idx", "i:label_assignments_pkey", "i:label_assignments_target_idx",
	"i:label_assignments_unique", "i:label_definitions_owner_key", "i:label_definitions_pkey", "i:label_definitions_scope_name_key",
	"r:label_assignments", "r:label_definitions"}

// metadataThroughVersion20 is the version-20 form (G020): the version-19 objects plus the label relations (88); in schema
// app the three owner-owned label trigger functions (the two keys-immutable guards and the definitions bound); in schema
// audit the two label captures (events_attribute is replaced in place). TestLabelsCatalog names the rest.
func metadataThroughVersion20(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	metadataThroughVersion(t, cfg, r, "20")
}

// platformLabelRelations is the 1 relation of 00021_platform_labels.sql (G057): the partial unique index on
// app.organizations.platform_label (the column and its CHECKs are no relation).
var platformLabelRelations = []string{"i:organizations_platform_label_key"}

// metadataThroughVersion21 is the version-21 form (G057): the version-20 objects plus the platform-label index (89); in
// schema app the auth reader's slug-availability reader (app.hostname_state and the provisioning body are replaced under
// their names and owners); schema audit is unchanged. TestPlatformLabelsCatalog names the rest.
func metadataThroughVersion21(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	metadataThroughVersion(t, cfg, r, "21")
}

// metadataThroughVersion22 is the version-22 form (G071): exactly the version-21 relations (89); in schema app the auth
// reader's default-tenant reader app.organization_id_by_slug; schema audit is unchanged. TestDefaultTenantCatalog names
// the rest.
func metadataThroughVersion22(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	metadataThroughVersion(t, cfg, r, "22")
}

func metadataThroughVersion(t *testing.T, cfg Config, r *Runner, version string) {
	t.Helper()
	o, pr, ar, cw := cfg.OwnerRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader", cfg.Database+"_claim_writer"
	pw := cfg.Database + "_provision_writer"
	extra := slices.Concat(storedObjectRelations, mcpRelations, domainRelations12, brandingRelations, membershipRelations,
		enrollmentRelations)
	guard, signupReader, signupBodies, labelBodies, tenantReader := "", "", "", "", ""
	if version == "17" || version == "18" || version == "19" || version == "20" || version == "21" || version == "22" {
		guard = ",organizations_group_override_guard:" + o
	}
	if version == "18" || version == "19" || version == "20" || version == "21" || version == "22" {
		extra = append(extra, signupRelations...)
		signupReader = ",auth_resolve_signup:" + ar
		signupBodies = ",signup_challenge_complete:" + pw + ",signup_challenge_issue:" + pw + ",signup_provision_organization:" + pw
	}
	if version == "20" || version == "21" || version == "22" {
		extra = append(extra, labelRelations...)
		labelBodies = ",label_assignments_keys_immutable:" + o + ",label_definitions_bound:" + o + ",label_definitions_keys_immutable:" + o
	}
	if version == "21" || version == "22" {
		extra = append(extra, platformLabelRelations...)
		signupBodies += ",signup_slug_taken:" + ar
	}
	if version == "22" {
		// G071: no relation; the auth reader's default-tenant reader (byte order: after organization_domains_guard).
		tenantReader = ",organization_id_by_slug:" + ar
	}
	metadataRelationsOnly(t, cfg, r, version, "auth_logout_ref:"+ar+",auth_principal_subject:"+ar+",auth_resolve_enrollment:"+ar+
		",auth_resolve_session:"+ar+signupReader+",branding_overrides_valid:"+o+",business_profiles_keys_immutable:"+o+",ctx_group:"+o+
		",ctx_login_binding:"+o+",ctx_org:"+o+",ctx_user:"+o+",enrollment_admit:"+cw+",enrollment_history_backfill:"+o+
		",enrollment_lock_organization:"+cw+",enrollment_readiness_set:"+o+",enrollment_request:"+cw+
		",enrollment_requests_immutable:"+o+",enrollment_restrict_on_decision:"+o+
		",enrollment_restrict_on_departure:"+o+",group_business_profiles_keys_immutable:"+o+",groups_bootstrap_owner:"+o+
		",groups_keys_immutable:"+o+",hostname_state:"+ar+",invitation_claim_group:"+cw+",invitation_claim_organization:"+cw+
		",invitations_bound:"+o+",invitations_keys_immutable:"+o+",is_organization_member:"+ar+labelBodies+",mcp_connections_guard:"+o+
		",organization_branding_guard:"+o+",organization_domains_guard:"+o+tenantReader+",organizations_enrollment_guard:"+o+guard+
		",organizations_keys_immutable:"+o+
		",organizations_slug_guard:"+o+",public_branding:"+ar+",retention_eligible:"+o+
		",rls_dependent_group_bindings:"+pr+",rls_enrollment_mode:"+pr+",rls_group_invitation_candidates:"+pr+
		",rls_group_role:"+pr+",rls_invitation_pending:"+cw+",rls_org_role:"+pr+
		",rls_own_organizations:"+pr+",rls_owner_count:"+pr+",rls_profile_eligible:"+pr+",rls_sole_owner_scopes:"+pr+
		",rls_user_enabled_lock:"+cw+",role_bindings_authority_guard:"+o+",role_bindings_keys_immutable:"+o+signupBodies+
		",stored_objects_guard:"+o, extra...)
	// Slice 1b: + the two enrollment captures (policy/readiness and backfill); request slice: + the request capture.
	fns := append(slices.Clone(auditFunctions), "capture_invitations", "capture_mcp_connections", "capture_organization_branding",
		"capture_organization_domains", "capture_enrollment_restrictions", "capture_organization_enrollment",
		"capture_enrollment_requests")
	if version == "20" || version == "21" || version == "22" {
		fns = append(fns, "capture_label_assignments", "capture_label_definitions")
	}
	slices.Sort(fns)
	auditObjects(t, r, "i:events_pkey,i:events_tenant_time_idx,r:events|"+strings.Join(fns, ","))
}

// auditObjects asserts the objects of schema audit ("relations|functions"; "" for an empty schema) and that it exists.
func auditObjects(t *testing.T, r *Runner, want string) {
	t.Helper()
	var got string
	err := r.db.QueryRowContext(context.Background(), `
SELECT coalesce((SELECT string_agg(relkind::text || ':' || relname, ',' ORDER BY relkind::text, relname) FROM pg_class
                 WHERE relnamespace = 'audit'::regnamespace), '') || '|' ||
       coalesce((SELECT string_agg(proname, ',' ORDER BY proname) FROM pg_proc WHERE pronamespace = 'audit'::regnamespace), '')`).Scan(&got)
	if want == "" {
		want = "|"
	}
	if err != nil || got != want {
		t.Errorf("schema audit objects: %q (%v), want %q", got, err, want)
	}
}

func metadataRelationsOnly(t *testing.T, cfg Config, r *Runner, version, wantFuncs string, extra ...string) {
	t.Helper()
	var objs, wrongOwner, funcs, public string
	// G058: the relations are ordered in byte order (COLLATE "C"), the order of slices.Sort below; from version 16 the
	// ledger's identity sequence (relkind S) is the first relation whose order differs under a linguistic collation.
	err := r.db.QueryRowContext(context.Background(), `
WITH meta AS (
  SELECT 'app.goose_db_version'::regclass AS oid
  UNION ALL SELECT pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
  UNION ALL SELECT i.indexrelid::regclass FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass
), objs AS (
  SELECT c.relname, c.relkind, c.relowner FROM pg_class c
  WHERE c.relnamespace = 'app'::regnamespace AND c.oid NOT IN (SELECT oid FROM meta)
)
SELECT
  coalesce((SELECT string_agg(relkind::text || ':' || relname, ',' ORDER BY relkind::text COLLATE "C", relname COLLATE "C") FROM objs), ''),
  coalesce((SELECT string_agg(relname, ',' ORDER BY relname) FROM objs WHERE relowner::regrole::text <> $1), ''),
  coalesce((SELECT string_agg(proname || ':' || proowner::regrole::text, ',' ORDER BY proname) FROM pg_proc
            WHERE pronamespace = 'app'::regnamespace), ''),
  coalesce((SELECT string_agg(c.relname, ',') FROM pg_class c
            WHERE c.relnamespace = 'public'::regnamespace AND c.relowner::regrole::text = ANY($2)), '')`,
		cfg.OwnerRole(), []string{cfg.OwnerRole(), cfg.MigratorRole(), cfg.RuntimeRole(), cfg.RetentionRole()}).Scan(&objs, &wrongOwner, &funcs, &public)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]string{
		"i:identity_principals_pkey", "i:identity_principals_user_id_key", "i:identity_provider_binding_pkey",
		"i:login_transactions_pkey", "i:sessions_credential_hash_key", "i:sessions_pkey",
		"i:sessions_provider_logout_ref_hash_key", "i:sessions_user_live_idx", "i:users_pkey", "r:identity_principals",
		"r:identity_provider_binding", "r:login_transactions", "r:sessions", "r:users",
	}, tenancyRelations...)
	want = append(want, businessProfileRelations...)
	want = append(want, extra...)
	slices.Sort(want)
	if objs != strings.Join(want, ",") || wrongOwner != "" || funcs != wantFuncs || public != "" {
		t.Errorf("objects at version %s: objects=%q not-owner=%q functions=%q public=%q", version, objs, wrongOwner, funcs, public)
	}
}

func run(t *testing.T, name string, f func(context.Context, *bytes.Buffer) error) string {
	t.Helper()
	var out bytes.Buffer
	if err := f(context.Background(), &out); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return out.String()
}

func TestMigrationLifecycle(t *testing.T) {
	cfg := itConfig(t)
	rt := connectAs(t, cfg, cfg.RuntimeRole(), os.Getenv("DB_RUNTIME_PASSWORD"))

	t.Run("01 fresh version creates protected metadata", func(t *testing.T) {
		reset(t, cfg)
		r := newRunner(t, cfg, false, openOptions{})
		out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) })
		if out != "current_version\t0\n" || !metadataExists(t, rt) {
			t.Fatalf("version on a fresh database: %q", out)
		}
		assertP(t, cfg, r, rt)
	})

	t.Run("02 fresh status before the baseline", func(t *testing.T) {
		reset(t, cfg)
		r := newRunner(t, cfg, false, openOptions{})
		out := run(t, "status", func(ctx context.Context, b *bytes.Buffer) error { return r.Status(ctx, b) })
		if !strings.Contains(out, "migration\t00001_baseline.sql\tpending") || !strings.Contains(out, "migration\t00002_identity.sql\tpending") ||
			!strings.Contains(out, "migration\t00003_tenancy.sql\tpending") || !strings.Contains(out, "migration\t00004_business_profiles.sql\tpending") ||
			!strings.Contains(out, "migration\t00005_organization_policy.sql\tpending") ||
			!strings.Contains(out, "migration\t00006_row_level_security.sql\tpending") || !strings.HasSuffix(out, "current_version\t0\n") {
			t.Fatalf("status on a fresh database: %q", out)
		}
		assertP(t, cfg, r, rt)
	})

	t.Run("03 unprotected creation is never committed", func(t *testing.T) {
		reset(t, cfg)
		r := newRunner(t, cfg, false, openOptions{skipRevoke: true})
		var out bytes.Buffer
		err := r.Status(context.Background(), &out)
		if err == nil || !strings.Contains(err.Error(), "not owner-only") {
			t.Fatalf("status without the revokes must fail on the postcondition, got %v", err)
		}
		if metadataExists(t, rt) {
			t.Fatal("metadata table committed despite the failed postcondition")
		}
	})

	t.Run("04 concurrent first use", func(t *testing.T) {
		reset(t, cfg)
		a := newRunner(t, cfg, false, openOptions{})
		b := newRunner(t, cfg, false, openOptions{})
		var wg sync.WaitGroup
		errs := make([]error, 2)
		for i, r := range []*Runner{a, b} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs[i] = r.Status(context.Background(), &bytes.Buffer{})
			}()
		}
		wg.Wait()
		if errs[0] != nil || errs[1] != nil {
			t.Fatalf("concurrent status: %v / %v", errs[0], errs[1])
		}
		var n int
		if err := rt.QueryRow(context.Background(), `SELECT count(*) FROM pg_class WHERE relnamespace = 'app'::regnamespace AND relname = 'goose_db_version'`).Scan(&n); err != nil || n != 1 {
			t.Fatalf("metadata tables = %d (%v)", n, err)
		}
		assertP(t, cfg, a, rt)
	})

	var s1 string
	normal := newRunner(t, cfg, false, openOptions{})

	t.Run("05 up applies the reviewed set", func(t *testing.T) {
		out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return normal.Up(ctx, b) })
		if out != "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\n"+
			"applied\t00005_organization_policy.sql\napplied\t00006_row_level_security.sql\napplied\t00007_stored_objects.sql\n"+
			"applied\t00008_audit_events.sql\napplied\t00009_retention_maintenance.sql\napplied\t00010_mcp_connections.sql\n"+
			"applied\t00011_organization_domains.sql\napplied\t00012_session_binding_and_claim_exclusivity.sql\n"+
			"applied\t00013_organization_branding.sql\napplied\t00014_group_onboarding.sql\n"+
			"applied\t00015_membership_invitations.sql\napplied\t00016_organization_enrollment.sql\n"+
			"applied\t00017_group_authority.sql\napplied\t00018_platform_signup.sql\napplied\t00019_mcp_connection_names.sql\n"+
			"applied\t00020_labels.sql\napplied\t00021_platform_labels.sql\napplied\t00022_default_tenant.sql\ncurrent_version\t22\n" {
			t.Fatalf("up: %q", out)
		}
		// Versions 5 and 6 add no relation: the same 35 objects as version 4 (the G007 policy flag is a column; G008's
		// row-level security adds policies, functions and grants only). Version 7 adds the G009 stored-object table and
		// its two indexes (38) and the invoker trigger function. Version 8 (G011) adds nothing in schema app: the audit
		// relation, its indexes and functions live in schema audit. Version 9 (G013) adds no relation: one app function
		// (the retention eligibility check), policies and grants. Version 10 (G014) adds the connection table and its three
		// keys (42), the connection guard and the audit capture function. Version 11 (G016) adds the domain table, its
		// keys and the slug index (48), four app functions and the domain capture function. Version 12 (G016 security
		// amendment) swaps one domain index (48) and replaces the session reader under the same name. Version 13 (G017)
		// adds the branding table and its key (50), three app functions and the branding capture function. Version 14
		// (G018) adds no relation: the bootstrap trigger function and the policy reader's eligibility reader. Version 15
		// (G019) adds the invitation table, its two keys and three indexes (56), eleven app functions (the two invitation
		// triggers, the authority guard, five policy-reader readers, three claim-writer bodies) and the invitation capture.
		// Version 16 (G058 slice 1a) adds the enrollment credential and restriction tables, their keys, three indexes and
		// the ledger's identity sequence (66) and six app functions (the policy reader's mode reader, the auth reader's
		// resolver, three claim-writer bodies, the owner-owned departure trigger body); slice 1b adds the readiness
		// functions, the mode guard and two audit captures; the request slice adds the request table, its key and three
		// indexes (71), the claim writer's enrollment_request, the owner-owned decision body and state machine and the
		// request capture.
		// Version 17 (G060) adds no relation (71): one column, three column grants, the owner-owned override guard and its
		// trigger, re-created policies and replaced bodies under their existing names.
		// Version 18 (G062) adds the signup table, its key, five digest keys and one index (79), the auth reader's signup
		// resolver and the provision writer's three bodies; events_attribute is replaced in place.
		// Version 19 (G063) adds no relation and no app function (79): one column with its CHECK, two column grants and
		// the name capture trigger; events_attribute and the connection capture are replaced in place.
		// Version 20 (G020) adds the two label tables, their primary keys, the definitions' scoped-name and owner keys,
		// the assignments' unique key and two indexes (88), three owner-owned app trigger functions and the two label
		// captures; events_attribute is replaced in place.
		// Version 21 (G057) adds the partial unique platform-label index (89) and the auth reader's slug-availability
		// reader; app.hostname_state and the provisioning body are replaced under their names and owners.
		// Version 22 (G071) adds no relation (89): the auth reader's default-tenant reader app.organization_id_by_slug.
		metadataThroughVersion22(t, cfg, normal)
		groupAuthorityAtVersion17(t, normal)
		assertP(t, cfg, normal, rt)
		s1 = snapshot(t, normal)
	})

	combined := newRunner(t, cfg, true, openOptions{})

	t.Run("06 synthetic DDL runs as owner; runtime has no authority", func(t *testing.T) {
		out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return combined.Up(ctx, b) })
		if out != "applied\t90000001_synthetic_ddl_probe.sql\ncurrent_version\t90000001\n" {
			t.Fatalf("synthetic up: %q", out)
		}
		var owner string
		if err := combined.db.QueryRowContext(context.Background(), `SELECT relowner::regrole::text FROM pg_class WHERE oid = 'app.synthetic_ddl_probe'::regclass`).Scan(&owner); err != nil || owner != cfg.OwnerRole() {
			t.Fatalf("probe table owner = %q (%v), want %s", owner, err, cfg.OwnerRole())
		}
		assertP(t, cfg, combined, rt)

		expectDenied(t, rt, "runtime CREATE TABLE in app", `CREATE TABLE app.runtime_ddl (id int)`)
		expectDenied(t, rt, "runtime CREATE TABLE in public", `CREATE TABLE public.runtime_ddl (id int)`)
		expectDenied(t, rt, "runtime SELECT probe", `SELECT 1 FROM app.synthetic_ddl_probe`)
		expectDenied(t, rt, "runtime SELECT metadata", `SELECT 1 FROM app.goose_db_version`)
		for _, role := range []string{cfg.OwnerRole(), cfg.MigratorRole(), "postgres"} {
			expectDenied(t, rt, "runtime SET ROLE "+role, `SET ROLE `+quote(role))
		}
		var super, bypass, memberOwner, memberMigrator bool
		var owned int
		err := rt.QueryRow(context.Background(), `
SELECT r.rolsuper, r.rolbypassrls, pg_has_role(r.rolname, $1, 'MEMBER'), pg_has_role(r.rolname, $2, 'MEMBER'),
       (SELECT count(*) FROM pg_class WHERE relowner = r.oid)
FROM pg_roles r WHERE r.rolname = current_user`, cfg.OwnerRole(), cfg.MigratorRole()).Scan(&super, &bypass, &memberOwner, &memberMigrator, &owned)
		if err != nil {
			t.Fatal(err)
		}
		if super || bypass || memberOwner || memberMigrator || owned != 0 {
			t.Errorf("runtime attributes: super=%v bypassrls=%v member(owner)=%v member(migrator)=%v owned=%d", super, bypass, memberOwner, memberMigrator, owned)
		}
		mig := connectAs(t, cfg, cfg.MigratorRole(), cfg.Password)
		expectDenied(t, mig, "migrator without SET ROLE: CREATE TABLE in app", `CREATE TABLE app.migrator_ddl (id int)`)
	})

	t.Run("07 synthetic down restores the reviewed state", func(t *testing.T) {
		out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return combined.Down(ctx, b) })
		if out != "rolled_back\t90000001_synthetic_ddl_probe.sql\ncurrent_version\t22\n" {
			t.Fatalf("synthetic down: %q", out)
		}
		if got := snapshot(t, normal); got != s1 {
			t.Fatalf("catalog after synthetic down differs from S1:\n%s\n---\n%s", got, s1)
		}
		assertP(t, cfg, normal, rt)
	})

	t.Run("08 reapply", func(t *testing.T) {
		out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return normal.Reapply(ctx, b) })
		if out != "rolled_back\t00022_default_tenant.sql\napplied\t00022_default_tenant.sql\ncurrent_version\t22\n" {
			t.Fatalf("reapply: %q", out)
		}
		if got := snapshot(t, normal); got != s1 {
			t.Fatalf("catalog after reapply differs from S1:\n%s\n---\n%s", got, s1)
		}
		assertP(t, cfg, normal, rt)
	})

	t.Run("09 down to zero keeps only metadata; up restores", func(t *testing.T) {
		out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00022_default_tenant.sql\ncurrent_version\t21\n" {
			t.Fatalf("down from 22: %q", out)
		}
		metadataThroughVersion21(t, cfg, normal)
		groupAuthorityAtVersion17(t, normal)
		assertP(t, cfg, normal, rt)

		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00021_platform_labels.sql\ncurrent_version\t20\n" {
			t.Fatalf("down from 21: %q", out)
		}
		metadataThroughVersion20(t, cfg, normal)
		groupAuthorityAtVersion17(t, normal)
		assertP(t, cfg, normal, rt)

		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00020_labels.sql\ncurrent_version\t19\n" {
			t.Fatalf("down from 20: %q", out)
		}
		metadataThroughVersion19(t, cfg, normal)
		groupAuthorityAtVersion17(t, normal)
		assertP(t, cfg, normal, rt)

		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00019_mcp_connection_names.sql\ncurrent_version\t18\n" {
			t.Fatalf("down from 19: %q", out)
		}
		metadataThroughVersion18(t, cfg, normal)
		groupAuthorityAtVersion17(t, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00018_platform_signup.sql\ncurrent_version\t17\n" {
			t.Fatalf("down from 18: %q", out)
		}
		metadataThroughVersion17(t, cfg, normal)
		groupAuthorityAtVersion17(t, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00017_group_authority.sql\ncurrent_version\t16\n" {
			t.Fatalf("down from 17: %q", out)
		}
		metadataThroughVersion16Only(t, cfg, normal)
		groupAuthorityAbsent(t, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00016_organization_enrollment.sql\ncurrent_version\t15\n" {
			t.Fatalf("down from 16: %q", out)
		}
		metadataThroughVersion15Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00015_membership_invitations.sql\ncurrent_version\t14\n" {
			t.Fatalf("down from 15: %q", out)
		}
		metadataThroughVersion14Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00014_group_onboarding.sql\ncurrent_version\t13\n" {
			t.Fatalf("down from 14: %q", out)
		}
		metadataThroughVersion13Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00013_organization_branding.sql\ncurrent_version\t12\n" {
			t.Fatalf("down from 13: %q", out)
		}
		metadataThroughVersion12Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00012_session_binding_and_claim_exclusivity.sql\ncurrent_version\t11\n" {
			t.Fatalf("down from 12: %q", out)
		}
		metadataThroughVersion11Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00011_organization_domains.sql\ncurrent_version\t10\n" {
			t.Fatalf("down from 11: %q", out)
		}
		metadataThroughVersion10Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00010_mcp_connections.sql\ncurrent_version\t9\n" {
			t.Fatalf("down from 10: %q", out)
		}
		metadataThroughVersion9Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00009_retention_maintenance.sql\ncurrent_version\t8\n" {
			t.Fatalf("down from 9: %q", out)
		}
		metadataThroughVersion8Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00008_audit_events.sql\ncurrent_version\t7\n" {
			t.Fatalf("down from 8: %q", out)
		}
		metadataThroughVersion7Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00007_stored_objects.sql\ncurrent_version\t6\n" {
			t.Fatalf("down from 7: %q", out)
		}
		metadataThroughVersion6Only(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00006_row_level_security.sql\ncurrent_version\t5\n" {
			t.Fatalf("down from 6: %q", out)
		}
		metadataIdentityTenancyAndBusinessProfileOnly(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00005_organization_policy.sql\ncurrent_version\t4\n" {
			t.Fatalf("down from 5: %q", out)
		}
		metadataIdentityTenancyAndBusinessProfileOnly(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00004_business_profiles.sql\ncurrent_version\t3\n" {
			t.Fatalf("down from 4: %q", out)
		}
		metadataIdentityAndTenancyOnly(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00003_tenancy.sql\ncurrent_version\t2\n" {
			t.Fatalf("down from 3: %q", out)
		}
		metadataAndIdentityOnly(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00002_identity.sql\ncurrent_version\t1\n" {
			t.Fatalf("down from 2: %q", out)
		}
		onlyMetadataAndNothingInPublic(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		out = run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return normal.Down(ctx, b) })
		if out != "rolled_back\t00001_baseline.sql\ncurrent_version\t0\n" {
			t.Fatalf("down from 1: %q", out)
		}
		onlyMetadataAndNothingInPublic(t, cfg, normal)
		assertP(t, cfg, normal, rt)
		var buf bytes.Buffer
		if err := normal.Down(context.Background(), &buf); err == nil {
			t.Fatal("down at version 0 must fail")
		}
		out = run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return normal.Up(ctx, b) })
		if out != "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\n"+
			"applied\t00005_organization_policy.sql\napplied\t00006_row_level_security.sql\napplied\t00007_stored_objects.sql\n"+
			"applied\t00008_audit_events.sql\napplied\t00009_retention_maintenance.sql\napplied\t00010_mcp_connections.sql\n"+
			"applied\t00011_organization_domains.sql\napplied\t00012_session_binding_and_claim_exclusivity.sql\n"+
			"applied\t00013_organization_branding.sql\napplied\t00014_group_onboarding.sql\n"+
			"applied\t00015_membership_invitations.sql\napplied\t00016_organization_enrollment.sql\n"+
			"applied\t00017_group_authority.sql\napplied\t00018_platform_signup.sql\napplied\t00019_mcp_connection_names.sql\n"+
			"applied\t00020_labels.sql\napplied\t00021_platform_labels.sql\napplied\t00022_default_tenant.sql\ncurrent_version\t22\n" {
			t.Fatalf("up again: %q", out)
		}
		if got := snapshot(t, normal); got != s1 {
			t.Fatalf("catalog after down/up differs from S1:\n%s\n---\n%s", got, s1)
		}
		assertP(t, cfg, normal, rt)
	})

	t.Run("10 preflight refuses without the owner role", func(t *testing.T) {
		_ = normal.Close()
		_ = combined.Close()
		reset(t, cfg)
		_, err := open(context.Background(), cfg, migrations.FS, openOptions{skipSetRole: true})
		if err == nil || !strings.Contains(err.Error(), "refusing to migrate") {
			t.Fatalf("open without SET ROLE: %v", err)
		}
		if metadataExists(t, rt) {
			t.Fatal("refused run created metadata")
		}
	})
}

// groupAuthorityAtVersion17 is the version-17 delta of 00017_group_authority.sql (G060 correction02 C02-1): the
// boolean column NOT NULL DEFAULT false, exactly three column grants on it (runtime UPDATE, claim writer and policy
// reader SELECT), the per-grantee column grants of the two helpers, and the owner-owned SECURITY INVOKER guard with its
// BEFORE UPDATE OF trigger.
func groupAuthorityAtVersion17(t *testing.T, r *Runner) {
	t.Helper()
	ctx := context.Background()
	q := func(sql string, args ...any) string {
		t.Helper()
		var out string
		if err := r.db.QueryRowContext(ctx, sql, args...).Scan(&out); err != nil {
			t.Fatalf("version 17: %v", err)
		}
		return out
	}
	db := q(`SELECT current_database()`)
	rt, cw, pr := db+"_runtime", db+"_claim_writer", db+"_policy_reader"
	if got := q(`SELECT format_type(a.atttypid, a.atttypmod) || ':' || a.attnotnull::text || ':' || pg_get_expr(d.adbin, d.adrelid)
FROM pg_attribute a JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = 'app.organizations'::regclass AND a.attname = 'group_management_override'`); got != "boolean:true:false" {
		t.Errorf("group_management_override column: %s", got)
	}
	colgrants := func(grantee string) string {
		return q(`SELECT coalesce(string_agg(x, ',' ORDER BY x COLLATE "C"), '') FROM (SELECT at.attname || ':' || a.privilege_type AS x
FROM pg_attribute at CROSS JOIN LATERAL aclexplode(at.attacl) a
WHERE at.attrelid = 'app.organizations'::regclass AND at.attnum > 0 AND NOT at.attisdropped AND a.grantee = $1::regrole) s`, grantee)
	}
	for role, want := range map[string]string{
		cw: "enrollment_mode:SELECT,enrollment_mode:UPDATE,group_management_override:SELECT,id:SELECT",
		pr: "enrollment_mode:SELECT,group_management_override:SELECT,id:SELECT",
	} {
		if got := colgrants(role); got != want {
			t.Errorf("%s column grants on app.organizations: %s, want %s", role, got, want)
		}
	}
	if got := q(`SELECT coalesce(string_agg(a.grantee::regrole::text || ':' || a.privilege_type, ',' ORDER BY a.grantee::regrole::text COLLATE "C"), '')
FROM pg_attribute at CROSS JOIN LATERAL aclexplode(at.attacl) a
WHERE at.attrelid = 'app.organizations'::regclass AND at.attname = 'group_management_override'`); got != cw+":SELECT,"+pr+":SELECT,"+rt+":UPDATE" {
		t.Errorf("group_management_override ACL: %s", got)
	}
	if got := q(`SELECT p.prosecdef::text || ':' || (p.proowner = (SELECT nspowner FROM pg_namespace WHERE nspname = 'app'))::text || ':' ||
  coalesce((SELECT string_agg(a.grantee::regrole::text, '+') FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner), '-')
FROM pg_proc p WHERE p.oid = 'app.organizations_group_override_guard()'::regprocedure`); got != "false:true:-" {
		t.Errorf("override guard: %s", got)
	}
	if got := q(`SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE tgrelid = 'app.organizations'::regclass
  AND tgname = 'organizations_group_override_guard'`); strings.Replace(got, "EXECUTE FUNCTION app.", "EXECUTE FUNCTION ", 1) !=
		"CREATE TRIGGER organizations_group_override_guard BEFORE UPDATE OF group_management_override ON app.organizations "+
			"FOR EACH ROW EXECUTE FUNCTION organizations_group_override_guard()" {
		t.Errorf("override guard trigger: %s", got)
	}
}

// groupAuthorityAbsent is the version-16 state after 00017's down: no column, guard or trigger remains.
func groupAuthorityAbsent(t *testing.T, r *Runner) {
	t.Helper()
	var n int
	if err := r.db.QueryRowContext(context.Background(), `SELECT
  (SELECT count(*) FROM pg_attribute WHERE attrelid = 'app.organizations'::regclass AND attname = 'group_management_override' AND NOT attisdropped)
  + (SELECT count(*) FROM pg_proc WHERE proname = 'organizations_group_override_guard')
  + (SELECT count(*) FROM pg_trigger WHERE tgname = 'organizations_group_override_guard')`).Scan(&n); err != nil || n != 0 {
		t.Errorf("00017 objects after down: %d (%v)", n, err)
	}
}
