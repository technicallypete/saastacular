//go:build integration

// B-RLS-CATALOG and B-RLS-MIGRATION (G008 Plan r3 §8.3) against a disposable PostgreSQL initialized by the rendered role
// script (infra/validation/run.sh, suite "migrations"):
//
//   - TestRLSCatalog: CT-RLS-1..6 and the rolled-back negative controls N1-N13 (each yields exactly its detector code;
//     the catalog is unchanged afterwards);
//   - TestRLSMigration: MG-RLS-1..5 and 5b, and the migration role-switch evidence (owner restored after success and
//     after a rolled-back failure, no helper CREATE after up, no helper grant or function after down);
//   - TestRLSPersistSeed / TestRLSPersistVerify: MG-RLS-6 around a PostgreSQL restart (the harness passes the seeded IDs
//     from the seed's stdout to the verify step's RLS_PERSIST_IDS; nothing is bind-mounted);
//   - TestRLSProvisioned: PROV-2/3 on an instance provisioned by infra/postgres/provision-roles.sql (harness only).
//
// Assertions on rows run as the actual runtime role (attested); fixtures, setup-side hashes and the role-attribute
// controls use the separately attested superuser of the disposable instance (DB_SUPERUSER, DB_SUPERUSER_PASSWORD);
// catalog reads use the owner. Failure messages print codes and hashes, never row values.
package migrate

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
)

// --- connections -------------------------------------------------------------------------------------------------------

// rlsSetupConn is the separately attested superuser connection (decided Q4): fixtures, setup-side reads and rolled-back
// role controls only, never an isolation assertion.
func rlsSetupConn(t *testing.T, cfg Config) *pgx.Conn {
	t.Helper()
	user, pw := os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD")
	if user == "" || pw == "" {
		t.Fatal("integration settings: DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required from version 6")
	}
	c := connectAs(t, cfg, user, pw)
	var super bool
	if err := c.QueryRow(context.Background(), `SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&super); err != nil || !super {
		t.Fatalf("setup connection is not a superuser (%v)", err)
	}
	return c
}

// rlsRuntimeConn is the direct runtime login, attested (R2): session_user = current_user = runtime; not superuser,
// BYPASSRLS, CREATEROLE or CREATEDB; no membership in the owner, migrator, retention or either helper.
func rlsRuntimeConn(t *testing.T, cfg Config) *pgx.Conn {
	t.Helper()
	c := runtimeConn(t, cfg)
	rlsAttestRuntime(t, cfg, c)
	return c
}

func rlsAttestRuntime(t *testing.T, cfg Config, c interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}) {
	t.Helper()
	var got string
	err := c.QueryRow(context.Background(), `
SELECT format('%s|%s|%s|%s|%s|%s|%s', session_user, current_user, r.rolsuper, r.rolbypassrls, r.rolcreaterole, r.rolcreatedb,
  (SELECT count(*) FROM unnest($1::text[]) x WHERE pg_has_role(r.oid, x, 'MEMBER')))
FROM pg_roles r WHERE r.rolname = current_user`, []string{cfg.OwnerRole(), cfg.MigratorRole(), cfg.Database + "_retention",
		cfg.Database + "_policy_reader", cfg.Database + "_auth_reader"}).Scan(&got)
	rt := cfg.RuntimeRole()
	if err != nil || got != rt+"|"+rt+"|f|f|f|f|0" {
		t.Fatalf("runtime attestation: %q (%v)", got, err)
	}
}

// ownerConn is the migrator assuming the owner role (FORCE controls).
func rlsOwnerConn(t *testing.T, cfg Config) *pgx.Conn {
	t.Helper()
	c := connectAs(t, cfg, cfg.MigratorRole(), cfg.Password)
	if _, err := c.Exec(context.Background(), "SET ROLE "+quote(cfg.OwnerRole())); err != nil {
		t.Fatal(err)
	}
	return c
}

// rlsContext sets the four transaction-local settings in tx (user, organization, group, binding; "" = absent).
func rlsContext(t *testing.T, tx pgx.Tx, user, org, group, binding string) {
	t.Helper()
	if _, err := tx.Exec(context.Background(), `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true),
       set_config('app.group_id', $3, true), set_config('app.login_binding', $4, true)`, user, org, group, binding); err != nil {
		t.Fatal(err)
	}
}

// --- the version-6 catalog detector -------------------------------------------------------------------------------------

// rlsLabels maps the application roles onto stable labels ($1..$6: runtime, owner, policy reader, auth reader,
// retention, migrator).
func rlsLabelArgs(cfg Config) []any {
	return []any{cfg.RuntimeRole(), cfg.OwnerRole(), cfg.Database + "_policy_reader", cfg.Database + "_auth_reader",
		cfg.Database + "_retention", cfg.MigratorRole()}
}

// rlsUseArgs references the six label parameters in a query that needs none of them (PostgreSQL must type every
// parameter it is sent).
const rlsUseArgs = ` AND concat($1::text, $2::text, $3::text, $4::text, $5::text, $6::text) IS NOT NULL`

// rlsLabel maps a role oid column onto the six parameter labels. The G019 claim-writer helper (<db>_claim_writer,
// provisioned by infra/postgres/init/10-roles.sh and provision-roles.sql at every version, D-14a) is derived from the
// connected database so that the six pinned parameters and every query built on them stay unchanged. The G062
// provision-writer helper (<db>_provision_writer, provisioned by the same two scripts at every version) is derived the
// same way.
func rlsLabel(col string) string {
	return `CASE ` + col + ` WHEN 0 THEN 'public' WHEN to_regrole($1) THEN 'rt' WHEN to_regrole($2) THEN 'ow' WHEN to_regrole($3) THEN 'pr' ` +
		`WHEN to_regrole($4) THEN 'ar' WHEN to_regrole($5) THEN 'rn' WHEN to_regrole($6) THEN 'mg' ` +
		`WHEN to_regrole(current_database() || '_claim_writer') THEN 'cw' ` +
		`WHEN to_regrole(current_database() || '_provision_writer') THEN 'pw' ELSE 'other' END`
}

// The pinned version-6 catalog (Plan r3 §3.1, §4.2): the same strings the migration's postcondition checks, plus the
// role, membership, view and policy-expression pins the migration does not check.
var (
	rlsWantFlags = "business_profiles:t:t,goose_db_version:f:f,group_business_profiles:t:t,groups:t:t,identity_principals:t:t," +
		"identity_provider_binding:f:f,login_transactions:t:t,organizations:t:t,role_bindings:t:t,sessions:t:t,users:t:t"
	rlsWantPolicies = "business_profiles_select:business_profiles:r:true:ow+rt;business_profiles_update:business_profiles:w:true:ow+rt;" +
		"group_business_profiles_select:group_business_profiles:r:true:ow+rt;groups_policy_reader:groups:r:true:pr;" +
		"groups_select:groups:r:true:ow+rt;groups_update:groups:w:true:ow+rt;" +
		"identity_principals_auth_reader:identity_principals:r:true:ar;" +
		"identity_principals_insert:identity_principals:a:true:ow+rt;identity_principals_select:identity_principals:r:true:ow+rt;" +
		"login_transactions_delete:login_transactions:d:true:ow+rt;login_transactions_expire:login_transactions:d:true:ow+rt;" +
		"login_transactions_insert:login_transactions:a:true:ow+rt;login_transactions_select:login_transactions:r:true:ow+rt;" +
		"organizations_select:organizations:r:true:ow+rt;organizations_update:organizations:w:true:ow+rt;" +
		"role_bindings_lock:role_bindings:w:true:ow+rt;role_bindings_policy_reader:role_bindings:r:true:pr;" +
		"role_bindings_select:role_bindings:r:true:ow+rt;sessions_auth_reader:sessions:r:true:ar;" +
		"sessions_insert:sessions:a:true:ow+rt;sessions_select:sessions:r:true:ow+rt;sessions_update:sessions:w:true:ow+rt;" +
		"users_auth_reader:users:r:true:ar;users_insert:users:a:true:ow+rt;users_select:users:r:true:ow+rt;" +
		"users_update:users:w:true:ow+rt"
	rlsWantACL = "business_profile_access_paths:rt:SELECT;business_profiles:rt:SELECT;business_profiles.description:rt:UPDATE;" +
		"business_profiles.display_name:rt:UPDATE;business_profiles.industry:rt:UPDATE;business_profiles.legal_name:rt:UPDATE;" +
		"business_profiles.website:rt:UPDATE;group_business_profiles:rt:SELECT;group_memberships:rt:SELECT;groups:rt:SELECT;" +
		"groups.id:pr:SELECT;groups.name:rt:UPDATE;groups.organization_id:pr:SELECT;groups.slug:rt:UPDATE;" +
		"identity_principals:rt:INSERT,SELECT;identity_principals.issuer:ar:SELECT;identity_principals.subject:ar:SELECT;" +
		"identity_principals.user_id:ar:SELECT;identity_provider_binding:rt:INSERT,SELECT;login_transactions:rt:DELETE,INSERT,SELECT;" +
		"organization_memberships:rt:SELECT;organizations:rt:SELECT;organizations.group_admins_edit_business_profiles:rt:UPDATE;" +
		"role_bindings.group_id:pr:SELECT;role_bindings.group_id:rt:SELECT;role_bindings.organization_id:pr:SELECT;" +
		"role_bindings.organization_id:rt:SELECT;role_bindings.role:pr:SELECT;role_bindings.role:rt:SELECT,UPDATE;" +
		"role_bindings.scope_id:pr:SELECT;role_bindings.scope_id:rt:SELECT;role_bindings.scope_type:pr:SELECT;" +
		"role_bindings.scope_type:rt:SELECT;role_bindings.source:pr:SELECT;role_bindings.source:rt:SELECT;" +
		"role_bindings.subject_user_id:pr:SELECT;role_bindings.subject_user_id:rt:SELECT;sessions:rt:INSERT,SELECT;" +
		"sessions.absolute_expires_at:ar:SELECT;sessions.auth_time:ar:SELECT;sessions.credential_hash:ar:SELECT;" +
		"sessions.csrf_token:ar:SELECT;sessions.id:ar:SELECT;sessions.idle_expires_at:ar:SELECT;sessions.idle_expires_at:rt:UPDATE;" +
		"sessions.mfa_assurance:ar:SELECT;sessions.provider_logout_ref_expires_at:ar:SELECT;" +
		"sessions.provider_logout_ref_expires_at:rt:UPDATE;sessions.provider_logout_ref_hash:ar:SELECT;" +
		"sessions.provider_logout_ref_hash:rt:UPDATE;sessions.provider_logout_returned_at:ar:SELECT;" +
		"sessions.provider_logout_returned_at:rt:UPDATE;sessions.provider_revocation:rt:UPDATE;" +
		"sessions.provider_session_id:ar:SELECT;sessions.renewed_at:ar:SELECT;sessions.renewed_at:rt:UPDATE;" +
		"sessions.revoke_reason:rt:UPDATE;sessions.revoked_at:ar:SELECT;sessions.revoked_at:rt:UPDATE;sessions.user_id:ar:SELECT;" +
		"users.created_at:rt:INSERT;users.disabled_at:rt:SELECT,UPDATE;users.display_name:ar:SELECT;" +
		"users.display_name:rt:INSERT,SELECT,UPDATE;users.id:ar:SELECT;users.id:rt:INSERT,SELECT;users.status:ar:SELECT;" +
		"users.status:rt:SELECT,UPDATE;users.updated_at:rt:INSERT,UPDATE"
	rlsWantFunctions = "auth_logout_ref:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;" +
		"auth_principal_subject:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;" +
		"auth_resolve_session:ar:true:false:s:search_path=pg_catalog, pg_temp:rt=EXECUTE;" +
		"business_profiles_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;" +
		"ctx_group:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;" +
		"ctx_login_binding:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;" +
		"ctx_org:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;" +
		"ctx_user:ow:false:false:s:search_path=pg_catalog, pg_temp:pr=EXECUTE+rt=EXECUTE;" +
		"group_business_profiles_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;" +
		"groups_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;" +
		"organizations_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;" +
		"rls_group_role:pr:true:false:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;" +
		"rls_org_role:pr:true:false:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;" +
		"role_bindings_keys_immutable:ow:false:false:v:search_path=pg_catalog:-"
	rlsWantSchema = "ar=USAGE,pr=USAGE,rn=USAGE,rt=USAGE"
	rlsWantRoles  = "ar:login=false,inherit=false,super=false,bypass=false,createrole=false,createdb=false,repl=false;" +
		"mg:login=true,inherit=true,super=false,bypass=false,createrole=false,createdb=false,repl=false;" +
		"ow:login=false,inherit=false,super=false,bypass=false,createrole=false,createdb=false,repl=false;" +
		"pr:login=false,inherit=false,super=false,bypass=false,createrole=false,createdb=false,repl=false;" +
		"rn:login=true,inherit=false,super=false,bypass=false,createrole=false,createdb=false,repl=false;" +
		"rt:login=true,inherit=false,super=false,bypass=false,createrole=false,createdb=false,repl=false"
	// Memberships touching an application role. An ADMIN-only grant (no INHERIT, no SET) to a non-application role (the
	// provisioning creator) is reported as "admin-only" and tolerated only where a test allows it.
	// The migrator's SET-only membership in the owner and the owner's in the four NOLOGIN NOINHERIT helpers (G008 ar/pr,
	// G019 cw, G062 pw), exactly as 10-roles.sh and provision-roles.sql grant them.
	rlsWantMemberships = "mg>ow:inherit=false,set=true,admin=false;ow>ar:inherit=false,set=true,admin=false;" +
		"ow>cw:inherit=false,set=true,admin=false;ow>pr:inherit=false,set=true,admin=false;" +
		"ow>pw:inherit=false,set=true,admin=false"
	rlsWantViews = "business_profile_access_paths:security_invoker=true;group_memberships:security_invoker=true;" +
		"organization_memberships:security_invoker=true"
)

// rlsCatalogParts reads the version-6 catalog parts, keyed by detector code.
func rlsCatalogParts(t *testing.T, q catalogQuerier, cfg Config) map[string]string {
	t.Helper()
	args := rlsLabelArgs(cfg)
	one := func(sql string) string {
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
	p["rls-flags"] = one(`SELECT string_agg(format('%s:%s:%s', c.relname, CASE WHEN c.relrowsecurity THEN 't' ELSE 'f' END,
  CASE WHEN c.relforcerowsecurity THEN 't' ELSE 'f' END), ',' ORDER BY c.relname) FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relkind = 'r'` + rlsUseArgs)
	p["policy-inventory"] = one(`SELECT string_agg(format('%s:%s:%s:%s:%s', p.polname, c.relname, p.polcmd, p.polpermissive::text,
  (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (SELECT ` + rlsLabel("x") + ` AS v FROM unnest(p.polroles) x) l)),
  ';' ORDER BY p.polname) FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid`)
	p["acl"] = one(`SELECT string_agg(t.k || ':' || t.g || ':' || t.privs, ';' ORDER BY t.k COLLATE "C", t.g COLLATE "C") FROM (
  SELECT c.relname::text AS k, ` + rlsLabel("a.grantee") + ` AS g, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type) AS privs
  FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
  WHERE c.relnamespace = 'app'::regnamespace AND a.grantee <> c.relowner GROUP BY c.relname, a.grantee
  UNION ALL
  SELECT c.relname || '.' || at.attname, ` + rlsLabel("a.grantee") + `, string_agg(a.privilege_type, ',' ORDER BY a.privilege_type)
  FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
  CROSS JOIN LATERAL aclexplode(at.attacl) a
  WHERE c.relnamespace = 'app'::regnamespace GROUP BY c.relname, at.attname, a.grantee) t`)
	p["functions"] = one(`SELECT string_agg(format('%s:%s:%s:%s:%s:%s:%s', p.proname, ` + rlsLabel("p.proowner") + `,
  p.prosecdef::text, p.proleakproof::text, p.provolatile, coalesce(array_to_string(p.proconfig, '|'), '-'),
  coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C") FROM (
     SELECT ` + rlsLabel("a.grantee") + ` || '=' || a.privilege_type AS v FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-')),
  ';' ORDER BY p.proname) FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace`)
	p["schema"] = one(`SELECT string_agg(l.v, ',' ORDER BY l.v COLLATE "C") FROM (
  SELECT ` + rlsLabel("a.grantee") + ` || '=' || a.privilege_type AS v
  FROM pg_namespace n CROSS JOIN LATERAL aclexplode(n.nspacl) a WHERE n.nspname = 'app' AND a.grantee <> n.nspowner) l`)
	p["roles"] = one(`SELECT string_agg(format('%s:login=%s,inherit=%s,super=%s,bypass=%s,createrole=%s,createdb=%s,repl=%s', l,
  r.rolcanlogin::text, r.rolinherit::text, r.rolsuper::text, r.rolbypassrls::text, r.rolcreaterole::text, r.rolcreatedb::text,
  r.rolreplication::text), ';' ORDER BY l COLLATE "C")
FROM (SELECT r.*, ` + rlsLabel("r.oid") + ` AS l FROM pg_roles r WHERE r.rolname IN ($1, $2, $3, $4, $5, $6)) r`)
	p["memberships"] = one(`SELECT string_agg(v, ';' ORDER BY v COLLATE "C") FROM (
  SELECT format('%s>%s:inherit=%s,set=%s,admin=%s', ` + rlsLabel("m.member") + `, ` + rlsLabel("m.roleid") + `,
    m.inherit_option::text, m.set_option::text, m.admin_option::text) AS v
  FROM pg_auth_members m
  WHERE (m.member IN (to_regrole($1), to_regrole($2), to_regrole($3), to_regrole($4), to_regrole($5), to_regrole($6))
         OR m.roleid IN (to_regrole($1), to_regrole($2), to_regrole($3), to_regrole($4), to_regrole($5), to_regrole($6)))
    AND NOT (m.member NOT IN (to_regrole($1), to_regrole($2), to_regrole($3), to_regrole($4), to_regrole($5), to_regrole($6))
             AND NOT m.inherit_option AND NOT m.set_option)) s`)
	p["views"] = one(`SELECT string_agg(c.relname || ':' || array_to_string(c.reloptions, ','), ';' ORDER BY c.relname)
FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace AND c.relkind = 'v'` + rlsUseArgs)
	rows, err := q.Query(context.Background(), `SELECT p.polname::text, coalesce(pg_get_expr(p.polqual, p.polrelid), ''),
  coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '') FROM pg_policy p ORDER BY p.polname`)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	for rows.Next() {
		var name, qual, check string
		if err := rows.Scan(&name, &qual, &check); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		p["policy-expr:"+name] = normalize(qual + " WITH CHECK " + check)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("catalog: %v", err)
	}
	return p
}

// rlsViolations is the version-6 detector: the sorted codes of every catalog part that differs from its pin. An
// unpinned policy shows only in policy-inventory.
func rlsViolations(t *testing.T, q catalogQuerier, cfg Config) []string {
	t.Helper()
	p := rlsCatalogParts(t, q, cfg)
	want := map[string]string{"rls-flags": rlsWantFlags, "policy-inventory": rlsWantPolicies, "acl": rlsWantACL,
		"functions": rlsWantFunctions, "schema": rlsWantSchema, "roles": rlsWantRoles, "memberships": rlsWantMemberships,
		"views": rlsWantViews}
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
	slices.Sort(codes)
	return codes
}

// --- B-RLS-CATALOG -------------------------------------------------------------------------------------------------------

// rlsWorld is a two-Organization fixture inserted by the setup connection.
type rlsWorld struct{ userA, userB, orgA, orgB, groupA, groupB string }

func seedRLSWorld(t *testing.T, setup *pgx.Conn, tag string) rlsWorld {
	t.Helper()
	a, b := seedPolicyFixture(t, setup, tag+"a"), seedPolicyFixture(t, setup, tag+"b")
	return rlsWorld{userA: a.user, userB: b.user, orgA: a.org, orgB: b.org, groupA: a.group, groupB: b.group}
}

// rlsRunner6 is a runner over the migrations through 00006 (G009 Plan r1 §5.6): the G008 checks exercise 00006 itself
// at version 6, with their exact version-6 expectations; 00007 is B-OBJ-*'s.
func rlsRunner6(t *testing.T, cfg Config) *Runner {
	t.Helper()
	return runnerFor(t, cfg, subsetFS(t, 6))
}

func TestRLSCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	// From version 7 (G009): migrate to version 7, then step down from 7 on the empty object table, so the version-6
	// catalog below is the one a rollback of 00007 restores. G011: version 7 through the migrations up to 00007 (the
	// version-7 evidence stays version-7; 00008 is B-AUDIT-*'s).
	r := runnerFor(t, cfg, subsetFS(t, 7))
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t7\n") {
		t.Fatalf("up: %q", out)
	}
	if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out != "rolled_back\t00007_stored_objects.sql\ncurrent_version\t6\n" {
		t.Fatalf("down from 7: %q", out)
	}
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	rt := rlsRuntimeConn(t, cfg)
	w := seedRLSWorld(t, setup, "cat")

	t.Run("CT-RLS-1 relations unchanged, nine forced tables, invoker views", func(t *testing.T) {
		metadataThroughVersion6Only(t, cfg, r)
		p := rlsCatalogParts(t, cat, cfg)
		if p["rls-flags"] != rlsWantFlags || p["views"] != rlsWantViews {
			t.Errorf("flags %q views %q", p["rls-flags"], p["views"])
		}
	})
	t.Run("CT-RLS-2 policy inventory, expressions and dependency graph", func(t *testing.T) {
		p := rlsCatalogParts(t, cat, cfg)
		if p["policy-inventory"] != rlsWantPolicies {
			t.Errorf("policy inventory:\n got %s\nwant %s", p["policy-inventory"], rlsWantPolicies)
		}
		for name, want := range rlsWantExprs {
			if got := p["policy-expr:"+name]; got != want {
				t.Errorf("policy %s expression:\n got %s\nwant %s", name, got, want)
			}
		}
		// §4.3: helper policies call only the parsers; runtime policies reference only the relations of the acyclic
		// graph (users -> role_bindings; business_profiles -> group_business_profiles, organizations); nothing reads its
		// own relation through runtime policies.
		edges := lines(t, cat, `SELECT DISTINCT c.relname || '>' || d.relname FROM pg_policy p JOIN pg_class c ON c.oid = p.polrelid
JOIN pg_depend dep ON dep.classid = 'pg_policy'::regclass AND dep.objid = p.oid AND dep.refclassid = 'pg_class'::regclass
JOIN pg_class d ON d.oid = dep.refobjid WHERE d.oid <> c.oid ORDER BY 1`)
		if edges != "business_profiles>group_business_profiles\nbusiness_profiles>organizations\nusers>role_bindings" {
			t.Errorf("policy relation edges:\n%s", edges)
		}
		fns := lines(t, cat, `SELECT DISTINCT p.polname || '>' || f.proname FROM pg_policy p
JOIN pg_depend dep ON dep.classid = 'pg_policy'::regclass AND dep.objid = p.oid AND dep.refclassid = 'pg_proc'::regclass
JOIN pg_proc f ON f.oid = dep.refobjid WHERE p.polname LIKE '%\_reader' ORDER BY 1`)
		for _, l := range strings.Split(fns, "\n") {
			if !strings.Contains(l, ">ctx_") {
				t.Errorf("a helper policy depends on %s", l)
			}
		}
	})
	t.Run("CT-RLS-3 ACLs, functions, schema and sequences", func(t *testing.T) {
		p := rlsCatalogParts(t, cat, cfg)
		for _, k := range []string{"acl", "functions", "schema"} {
			want := map[string]string{"acl": rlsWantACL, "functions": rlsWantFunctions, "schema": rlsWantSchema}[k]
			if p[k] != want {
				t.Errorf("%s:\n got %s\nwant %s", k, p[k], want)
			}
		}
		// The exact readable column sets, as the privilege functions see them (role-wide; no masking relation).
		for _, c := range []struct{ role, table, want string }{
			{cfg.RuntimeRole(), "users", "disabled_at,display_name,id,status"},
			{cfg.RuntimeRole(), "role_bindings", "group_id,organization_id,role,scope_id,scope_type,source,subject_user_id"},
			{cfg.Database + "_policy_reader", "role_bindings", "group_id,organization_id,role,scope_id,scope_type,source,subject_user_id"},
			{cfg.Database + "_policy_reader", "groups", "id,organization_id"},
			{cfg.Database + "_auth_reader", "sessions", "absolute_expires_at,auth_time,credential_hash,csrf_token,id,idle_expires_at," +
				"mfa_assurance,provider_logout_ref_expires_at,provider_logout_ref_hash,provider_logout_returned_at,provider_session_id," +
				"renewed_at,revoked_at,user_id"},
			{cfg.Database + "_auth_reader", "users", "display_name,id,status"},
			{cfg.Database + "_auth_reader", "identity_principals", "issuer,subject,user_id"},
		} {
			got := lines(t, cat, `SELECT string_agg(attname, ',' ORDER BY attname) FROM pg_attribute
WHERE attrelid = ('app.' || $2)::regclass AND attnum > 0 AND NOT attisdropped AND has_column_privilege($1, attrelid, attnum, 'SELECT')`,
				c.role, c.table)
			if got != c.want {
				t.Errorf("%s SELECT on %s: %s, want %s", c.role, c.table, got, c.want)
			}
		}
		if seq := lines(t, cat, `SELECT c.relname || ':' || coalesce(array_to_string(c.relacl, ','), '-') FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relkind = 'S' AND c.relacl IS NOT NULL
  AND EXISTS (SELECT 1 FROM aclexplode(c.relacl) a WHERE a.grantee <> c.relowner)`); seq != "" {
			t.Errorf("a sequence grants a privilege beyond its owner: %s", seq)
		}
	})
	t.Run("CT-RLS-4 role attributes and memberships", func(t *testing.T) {
		p := rlsCatalogParts(t, cat, cfg)
		if p["roles"] != rlsWantRoles || p["memberships"] != rlsWantMemberships {
			t.Errorf("roles %s\nmemberships %s", p["roles"], p["memberships"])
		}
		rlsAttestRuntime(t, cfg, rt)
	})
	t.Run("CT-RLS-5 runtime denials", func(t *testing.T) {
		quoteRole := func(s string) string { return quote(s) }
		stmts := []string{
			`CREATE TABLE app.rls_probe (id int)`, `ALTER TABLE app.groups DISABLE ROW LEVEL SECURITY`,
			`ALTER TABLE app.sessions NO FORCE ROW LEVEL SECURITY`, `CREATE POLICY rls_probe ON app.users USING (true)`,
			`ALTER POLICY users_select ON app.users USING (true)`, `DROP POLICY users_select ON app.users`,
			`CREATE FUNCTION app.rls_probe() RETURNS int LANGUAGE sql AS 'SELECT 1'`,
			`ALTER FUNCTION app.rls_org_role() SECURITY INVOKER`,
			`SET ROLE ` + quoteRole(cfg.OwnerRole()), `SET ROLE ` + quoteRole(cfg.MigratorRole()),
			`SET ROLE ` + quoteRole(cfg.Database+"_policy_reader"), `SET ROLE ` + quoteRole(cfg.Database+"_auth_reader"),
			`SET ROLE ` + quoteRole(cfg.Database+"_retention"),
			`SELECT 1 FROM app.goose_db_version`,
			`INSERT INTO app.organizations (name) VALUES ('X')`,
			`INSERT INTO app.groups (organization_id, name, slug) VALUES ('` + w.orgA + `', 'X', 'x')`,
			`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source) VALUES ('` + w.orgA +
				`', '` + w.userB + `', 'organization', '` + w.orgA + `', 'org_owner', 'membership')`,
			`INSERT INTO app.business_profiles (organization_id, display_name) VALUES ('` + w.orgA + `', 'X')`,
			`INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) SELECT organization_id, '` + w.groupA +
				`', id FROM app.business_profiles LIMIT 1`,
			`DELETE FROM app.role_bindings`, `DELETE FROM app.group_business_profiles`, `DELETE FROM app.users`, `DELETE FROM app.sessions`,
			`DELETE FROM app.identity_principals`, `DELETE FROM app.organizations`, `DELETE FROM app.groups`, `DELETE FROM app.business_profiles`,
			`DELETE FROM app.identity_provider_binding`,
			`UPDATE app.organizations SET name = 'X'`, `UPDATE app.groups SET organization_id = organization_id`,
			`UPDATE app.role_bindings SET subject_user_id = subject_user_id`, `UPDATE app.role_bindings SET source = source`,
			`UPDATE app.role_bindings SET scope_id = scope_id`, `UPDATE app.business_profiles SET organization_id = organization_id`,
			`UPDATE app.group_business_profiles SET group_id = group_id`, `UPDATE app.users SET created_at = now()`,
			`UPDATE app.sessions SET credential_hash = credential_hash`, `UPDATE app.sessions SET user_id = user_id`,
			`UPDATE app.sessions SET mfa_assurance = mfa_assurance`, `UPDATE app.sessions SET csrf_token = csrf_token`,
			`UPDATE app.identity_principals SET subject = subject`, `UPDATE app.identity_provider_binding SET issuer = issuer`,
			`SELECT id FROM app.role_bindings`, `SELECT created_at FROM app.role_bindings`, `SELECT created_at FROM app.users`,
			`SELECT updated_at FROM app.users`,
			`INSERT INTO app.organization_memberships (user_id, organization_id, role) VALUES ('` + w.userB + `', '` + w.orgA + `', 'org_owner')`,
		}
		for _, tbl := range []string{"organizations", "groups", "role_bindings", "business_profiles", "group_business_profiles", "users",
			"sessions", "identity_principals", "login_transactions", "identity_provider_binding"} {
			stmts = append(stmts, `TRUNCATE app.`+tbl)
		}
		for _, s := range stmts {
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			rlsContext(t, tx, w.userA, w.orgA, "", "")
			_, err = tx.Exec(ctx, s)
			if got := sqlstateOf(err); got != "42501" {
				t.Errorf("runtime %q: SQLSTATE %q, want 42501", s, got)
			}
			_ = tx.Rollback(ctx)
		}
		// row_security = off makes a protected read fail rather than bypass (the runtime role cannot bypass).
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `SET LOCAL row_security = off`); err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, `SELECT 1 FROM app.organizations`)
		if got := sqlstateOf(err); got != "42501" {
			t.Errorf("row_security = off then a protected read: SQLSTATE %q, want 42501", got)
		}
		_ = tx.Rollback(ctx)
	})
	t.Run("CT-RLS-6 owner FORCE", func(t *testing.T) {
		own := rlsOwnerConn(t, cfg)
		var total int
		if err := setup.QueryRow(ctx, `SELECT count(*) FROM app.organizations WHERE id IN ($1::uuid, $2::uuid)`, w.orgA, w.orgB).Scan(&total); err != nil || total != 2 {
			t.Fatalf("setup sees %d of the two fixture Organizations (%v)", total, err)
		}
		tx, err := own.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		var none int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.organizations`).Scan(&none); err != nil || none != 0 {
			t.Errorf("owner without context sees %d Organizations (%v), want 0", none, err)
		}
		rlsContext(t, tx, w.userA, w.orgA, "", "")
		var ids string
		if err := tx.QueryRow(ctx, `SELECT coalesce(string_agg(id::text, ','), '') FROM app.organizations`).Scan(&ids); err != nil || ids != w.orgA {
			t.Errorf("owner in context A sees %s (%v), want only A", ids, err)
		}
		var foreign int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.groups WHERE organization_id = $1::uuid`, w.orgB).Scan(&foreign); err != nil || foreign != 0 {
			t.Errorf("owner in context A sees %d foreign Groups (%v)", foreign, err)
		}
		if tag, err := tx.Exec(ctx, `DELETE FROM app.business_profiles WHERE organization_id = $1::uuid`, w.orgB); err != nil || tag.RowsAffected() != 0 {
			t.Errorf("owner foreign DELETE: %v rows (%v), want 0", tag.RowsAffected(), err)
		}
		_, err = tx.Exec(ctx, `SAVEPOINT s; INSERT INTO app.organizations (name) VALUES ('Owner Insert')`)
		if got := sqlstateOf(err); got != "42501" {
			t.Errorf("owner INSERT under FORCE: SQLSTATE %q, want 42501 (row-level security violation)", got)
		}
	})
	t.Run("CT-RLS-NEG rolled-back controls yield exactly their codes", func(t *testing.T) {
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("the unmodified catalog has violations %v", v)
		}
		pr, ar, rtq := quote(cfg.Database+"_policy_reader"), quote(cfg.Database+"_auth_reader"), quote(cfg.RuntimeRole())
		ow := quote(cfg.OwnerRole())
		type neg struct {
			name  string
			super bool
			stmts []string
			want  string
		}
		negs := []neg{
			{"N1 NO FORCE on groups", false, []string{`ALTER TABLE app.groups NO FORCE ROW LEVEL SECURITY`}, "rls-flags"},
			{"N2 DISABLE on sessions", false, []string{`ALTER TABLE app.sessions DISABLE ROW LEVEL SECURITY`}, "rls-flags"},
			{"N3 an extra USING (true) runtime policy", false, []string{`CREATE POLICY rls_neg_open ON app.groups FOR SELECT TO ` + rtq + ` USING (true)`}, "policy-inventory"},
			{"N4 a FOR ALL policy", false, []string{`CREATE POLICY rls_neg_all ON app.users FOR ALL TO ` + rtq + ` USING (id IS NOT NULL)`}, "policy-inventory"},
			{"N5 GRANT INSERT on role_bindings", false, []string{`GRANT INSERT ON app.role_bindings TO ` + rtq}, "acl"},
			{"N6 PUBLIC EXECUTE on rls_org_role", false, []string{`SET LOCAL ROLE ` + pr, `GRANT EXECUTE ON FUNCTION app.rls_org_role() TO PUBLIC`, `SET LOCAL ROLE ` + ow}, "functions"},
			{"N7 a helper with LOGIN", true, []string{`ALTER ROLE ` + pr + ` LOGIN`}, "roles"},
			{"N8 runtime granted membership in the policy reader", true, []string{`GRANT ` + pr + ` TO ` + rtq}, "memberships"},
			{"N9 a definer without search_path", false, []string{`SET LOCAL ROLE ` + pr, `ALTER FUNCTION app.rls_org_role() RESET search_path`, `SET LOCAL ROLE ` + ow}, "functions"},
			{"N10 a view with security_invoker = false", false, []string{`ALTER VIEW app.group_memberships SET (security_invoker = false)`}, "views"},
			{"N11 runtime SELECT on users.created_at", false, []string{`GRANT SELECT (created_at) ON app.users TO ` + rtq}, "acl"},
			{"N12 auth reader SELECT on sessions.revoke_reason", false, []string{`GRANT SELECT (revoke_reason) ON app.sessions TO ` + ar}, "acl"},
			{"N13 login_transactions_expire without its EMPTY-shape terms", false, []string{
				`DROP POLICY login_transactions_expire ON app.login_transactions`,
				`CREATE POLICY login_transactions_expire ON app.login_transactions FOR DELETE TO ` + rtq + `, ` + ow + ` USING (expires_at <= now())`},
				"policy-expr:login_transactions_expire"},
		}
		for _, n := range negs {
			c := cat
			if n.super {
				c = setup
			}
			tx, err := c.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range n.stmts {
				if _, err := tx.Exec(ctx, s); err != nil {
					t.Fatalf("%s: %s: %v", n.name, s, err)
				}
			}
			got := strings.Join(rlsViolations(t, tx, cfg), ",")
			_ = tx.Rollback(ctx)
			if got != n.want {
				t.Errorf("%s: detector codes %q, want exactly %q", n.name, got, n.want)
			}
		}
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("the catalog changed after the rolled-back controls: %v", v)
		}
	})
}

// --- B-RLS-MIGRATION -----------------------------------------------------------------------------------------------------

// rlsRowsSnapshot hashes every row of every application table (setup connection; row-level security does not apply).
func rlsRowsSnapshot(t *testing.T, setup *pgx.Conn) string {
	t.Helper()
	var parts []string
	for _, tbl := range []string{"identity_provider_binding", "users", "identity_principals", "sessions", "login_transactions",
		"organizations", "groups", "role_bindings", "business_profiles", "group_business_profiles"} {
		var h string
		if err := setup.QueryRow(context.Background(), `SELECT count(*) || ':' || coalesce(md5(string_agg(r::text, E'\n' ORDER BY r::text)), '-') FROM app.`+tbl+` r`).Scan(&h); err != nil {
			t.Fatal(err)
		}
		parts = append(parts, tbl+"="+h)
	}
	return strings.Join(parts, ";")
}

// rlsVersion5Catalog is the version-5 catalog fingerprint: the migrate snapshot plus every app ACL, policy, RLS flag,
// function and the helpers' schema privileges.
func rlsVersion5Catalog(t *testing.T, cfg Config, r *Runner, cat *pgx.Conn) string {
	t.Helper()
	p := rlsCatalogParts(t, cat, cfg)
	return snapshot(t, r) + "\n" + p["rls-flags"] + "\n" + p["policy-inventory"] + "\n" + p["acl"] + "\n" + p["functions"] + "\n" + p["schema"]
}

func TestRLSMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	full := "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\n" +
		"applied\t00005_organization_policy.sql\napplied\t00006_row_level_security.sql\ncurrent_version\t6\n"

	t.Run("MG-RLS-1 empty to version 6", func(t *testing.T) {
		reset(t, cfg)
		r := rlsRunner6(t, cfg)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != full {
			t.Fatalf("up: %q", out)
		}
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
		rlsOwnerRestored(t, cfg, r)
	})
	t.Run("MG-RLS-2 repeated up and status", func(t *testing.T) {
		r := rlsRunner6(t, cfg)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "no pending migrations\ncurrent_version\t6\n" {
			t.Fatalf("repeat up: %q", out)
		}
		out := run(t, "status", func(ctx context.Context, b *bytes.Buffer) error { return r.Status(ctx, b) })
		if !strings.Contains(out, "migration\t00006_row_level_security.sql\tapplied") || !strings.HasSuffix(out, "current_version\t6\n") || strings.Contains(out, "pending") {
			t.Fatalf("status: %q", out)
		}
	})
	var v5catalog, populated string
	t.Run("MG-RLS-3 populated version 5 to 6 keeps every row", func(t *testing.T) {
		reset(t, cfg)
		r5 := runnerFor(t, cfg, subsetFS(t, 5))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r5.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t5\n") {
			t.Fatalf("up to 5: %q", out)
		}
		v5catalog = rlsVersion5Catalog(t, cfg, r5, cat)
		rt := runtimeConn(t, cfg) // version 5: no row-level security yet, the accepted runtime grants seed rows
		seedPolicyFixture(t, rt, "mga")
		seedPolicyFixture(t, rt, "mgb")
		populated = rlsRowsSnapshot(t, setup)
		r := rlsRunner6(t, cfg)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00006_row_level_security.sql\ncurrent_version\t6\n" {
			t.Fatalf("up to 6: %q", out)
		}
		if got := rlsRowsSnapshot(t, setup); got != populated {
			t.Fatalf("rows changed by the upgrade:\n got %s\nwant %s", got, populated)
		}
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
	})
	t.Run("MG-RLS-4 reapply keeps the catalog", func(t *testing.T) {
		r := rlsRunner6(t, cfg)
		before := snapshot(t, r)
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out != "rolled_back\t00006_row_level_security.sql\napplied\t00006_row_level_security.sql\ncurrent_version\t6\n" {
			t.Fatalf("reapply: %q", out)
		}
		if after := snapshot(t, r); after != before {
			t.Fatal("catalog changed by reapply")
		}
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
		if got := rlsRowsSnapshot(t, setup); got != populated {
			t.Fatal("rows changed by reapply")
		}
	})
	t.Run("MG-RLS-5 down to 5 restores the exact version-5 catalog with rows unchanged; up again", func(t *testing.T) {
		r := rlsRunner6(t, cfg)
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); out != "rolled_back\t00006_row_level_security.sql\ncurrent_version\t5\n" {
			t.Fatalf("down: %q", out)
		}
		r5 := runnerFor(t, cfg, subsetFS(t, 5))
		if got := rlsVersion5Catalog(t, cfg, r5, cat); got != v5catalog {
			t.Fatalf("version-5 catalog after down differs:\n got %s\nwant %s", got, v5catalog)
		}
		if got := rlsRowsSnapshot(t, setup); got != populated {
			t.Fatal("rows changed by down")
		}
		// The helpers survive as cluster objects holding nothing in app.
		var leftovers string
		if err := cat.QueryRow(ctx, `SELECT format('functions=%s,usage=%s/%s,create=%s/%s',
  (SELECT count(*) FROM pg_proc WHERE pronamespace = 'app'::regnamespace AND proowner IN (to_regrole($1), to_regrole($2))),
  has_schema_privilege($1, 'app', 'USAGE'), has_schema_privilege($2, 'app', 'USAGE'),
  has_schema_privilege($1, 'app', 'CREATE'), has_schema_privilege($2, 'app', 'CREATE'))`,
			cfg.Database+"_policy_reader", cfg.Database+"_auth_reader").Scan(&leftovers); err != nil || leftovers != "functions=0,usage=f/f,create=f/f" {
			t.Fatalf("helper leftovers after down: %s (%v)", leftovers, err)
		}
		rlsOwnerRestored(t, cfg, r)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00006_row_level_security.sql\ncurrent_version\t6\n" {
			t.Fatalf("up again: %q", out)
		}
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
	})
	t.Run("MG-RLS-5b G008 code against version 5 fails closed", func(t *testing.T) {
		r := rlsRunner6(t, cfg)
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); !strings.HasSuffix(out, "current_version\t5\n") {
			t.Fatalf("down: %q", out)
		}
		defer func() {
			if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t6\n") {
				t.Fatalf("up: %q", out)
			}
		}()
		pool := runtimePool(t, cfg)
		svc := identity.NewService(identity.NewStore(pool), identity.Config{Idle: 30 * time.Minute, Max: 8 * time.Hour})
		raw := make([]byte, 32)
		_, _ = rand.Read(raw)
		if _, err := svc.Session(ctx, base64.RawURLEncoding.EncodeToString(raw), "https://org-a.localhost:543"); err != identity.ErrUnavailable {
			t.Fatalf("Session at version 5: %v, want identity_unavailable (the E1 reader does not exist)", err)
		}
		if _, err := svc.ProviderLogoutReturn(ctx, base64.RawURLEncoding.EncodeToString(raw)); err != identity.ErrUnavailable {
			t.Fatalf("ProviderLogoutReturn at version 5: %v, want identity_unavailable", err)
		}
	})
	t.Run("MG-RLS-C4 a failed up rolls back and restores the owner", func(t *testing.T) {
		r := rlsRunner6(t, cfg)
		if out := run(t, "down", func(ctx context.Context, b *bytes.Buffer) error { return r.Down(ctx, b) }); !strings.HasSuffix(out, "current_version\t5\n") {
			t.Fatalf("down: %q", out)
		}
		// An unexpected grant the version-6 postcondition rejects, after the role switches have run.
		if _, err := setup.Exec(ctx, `GRANT SELECT ON app.organizations TO `+quote(cfg.Database+"_retention")); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		err := r.Up(ctx, &buf)
		if err == nil || !strings.Contains(err.Error(), "version-6 ACLs") {
			t.Fatalf("up with an unexpected grant: %v", err)
		}
		if _, err := setup.Exec(ctx, `REVOKE SELECT ON app.organizations FROM `+quote(cfg.Database+"_retention")); err != nil {
			t.Fatal(err)
		}
		r5 := runnerFor(t, cfg, subsetFS(t, 5))
		if got := rlsVersion5Catalog(t, cfg, r5, cat); got != v5catalog {
			t.Fatalf("version-5 catalog after the rolled-back up differs:\n got %s\nwant %s", got, v5catalog)
		}
		rlsOwnerRestored(t, cfg, r)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t6\n") {
			t.Fatalf("up: %q", out)
		}
		rlsOwnerRestored(t, cfg, r)
		if v := rlsViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
	})
}

// rlsOwnerRestored checks that every connection of the runner's pool acts as the owner again after a migration that
// switched roles (the readers' grants and drops run as their helper), and that the helpers hold no CREATE on app.
func rlsOwnerRestored(t *testing.T, cfg Config, r *Runner) {
	t.Helper()
	ctx := context.Background()
	for range 3 {
		var cur, sess string
		if err := r.db.QueryRowContext(ctx, `SELECT current_user::text, session_user::text`).Scan(&cur, &sess); err != nil {
			t.Fatal(err)
		}
		if cur != cfg.OwnerRole() || sess != cfg.MigratorRole() {
			t.Fatalf("runner session acts as %s (session %s), want %s", cur, sess, cfg.OwnerRole())
		}
	}
	var create bool
	if err := r.db.QueryRowContext(ctx, `SELECT has_schema_privilege($1, 'app', 'CREATE') OR has_schema_privilege($2, 'app', 'CREATE')`,
		cfg.Database+"_policy_reader", cfg.Database+"_auth_reader").Scan(&create); err != nil || create {
		t.Fatalf("a helper holds CREATE on app (%v)", err)
	}
}

// --- MG-RLS-6 persistence ------------------------------------------------------------------------------------------------

func TestRLSPersistSeed(t *testing.T) {
	cfg := itConfig(t)
	reset(t, cfg)
	r := rlsRunner6(t, cfg)
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t6\n") {
		t.Fatalf("up: %q", out)
	}
	w := seedRLSWorld(t, rlsSetupConn(t, cfg), "per")
	b, _ := json.Marshal(w)
	fmt.Println("RLS-PERSIST " + string(b))
}

func TestRLSPersistVerify(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	var w rlsWorld
	if err := json.Unmarshal([]byte(os.Getenv("RLS_PERSIST_IDS")), &w); err != nil || w.orgA == "" {
		t.Fatalf("RLS_PERSIST_IDS: %v", err)
	}
	r := rlsRunner6(t, cfg)
	if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t6\n" {
		t.Fatalf("version after restart: %q", out)
	}
	if v := rlsViolations(t, catalogConn(t, cfg), cfg); len(v) != 0 {
		t.Fatalf("detector after restart %v", v)
	}
	rt := rlsRuntimeConn(t, cfg)
	tx, err := rt.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rlsContext(t, tx, w.userA, w.orgA, "", "")
	var ids string
	if err := tx.QueryRow(ctx, `SELECT coalesce(string_agg(id::text, ','), '') FROM app.organizations`).Scan(&ids); err != nil || ids != w.orgA {
		t.Fatalf("runtime in context A after restart sees %s (%v)", ids, err)
	}
}

// MarshalJSON / UnmarshalJSON carry the persistence IDs (unexported fields).
func (w rlsWorld) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"userA": w.userA, "userB": w.userB, "orgA": w.orgA, "orgB": w.orgB, "groupA": w.groupA, "groupB": w.groupB})
}

func (w *rlsWorld) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	*w = rlsWorld{userA: m["userA"], userB: m["userB"], orgA: m["orgA"], orgB: m["orgB"], groupA: m["groupA"], groupB: m["groupB"]}
	return nil
}

// --- PROV-2/3 ----------------------------------------------------------------------------------------------------------

// TestRLSProvisioned runs against an instance provisioned without the image init script (infra/postgres/provision-roles.sql,
// harness rls_provision): it migrates to 6, checks the detector (an ADMIN-only creator grant is not a membership the
// detector reports), and repeats representative runtime positives and negatives per protected table plus owner FORCE.
func TestRLSProvisioned(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	r := rlsRunner6(t, cfg)
	if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); !strings.HasSuffix(out, "current_version\t6\n") {
		t.Fatalf("up: %q", out)
	}
	cat := catalogConn(t, cfg)
	if v := rlsViolations(t, cat, cfg); len(v) != 0 {
		t.Fatalf("detector on the provisioned instance %v", v)
	}
	setup := rlsSetupConn(t, cfg)
	w := seedRLSWorld(t, setup, "prov")
	rt := rlsRuntimeConn(t, cfg)
	for _, tbl := range []string{"organizations", "groups", "role_bindings", "business_profiles", "group_business_profiles", "users",
		"sessions", "identity_principals"} {
		col := map[string]string{"organizations": "id", "users": "id", "sessions": "user_id", "identity_principals": "user_id"}[tbl]
		if col == "" {
			col = "organization_id"
		}
		want := map[string]string{"organizations": w.orgA, "users": w.userA, "sessions": w.userA, "identity_principals": w.userA}[tbl]
		if want == "" {
			want = w.orgA
		}
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		rlsContext(t, tx, w.userA, w.orgA, "", "")
		var got string
		if err := tx.QueryRow(ctx, `SELECT coalesce(string_agg(DISTINCT `+col+`::text, ','), '') FROM app.`+tbl).Scan(&got); err != nil {
			t.Fatalf("%s: %v", tbl, err)
		}
		if got != want {
			t.Errorf("%s in context A: %s, want only %s's rows", tbl, sha(got), sha(want))
		}
		_ = tx.Rollback(ctx)
	}
	for _, s := range []string{`DELETE FROM app.role_bindings`, `TRUNCATE app.users`, `SET ROLE ` + quote(cfg.OwnerRole())} {
		_, err := rt.Exec(ctx, s)
		if sqlstateOf(err) != "42501" {
			t.Errorf("runtime %q: %v, want 42501", s, err)
		}
	}
	own := rlsOwnerConn(t, cfg)
	var n int
	if err := own.QueryRow(ctx, `SELECT count(*) FROM app.organizations`).Scan(&n); err != nil || n != 0 {
		t.Errorf("owner without context sees %d Organizations (%v)", n, err)
	}
}

func sha(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))[:12] }

// rlsWantExprs pins every policy expression as PostgreSQL deparses it (USING, then WITH CHECK), whitespace-normalized:
// CT-RLS-2 and N13 detect a changed predicate that the migration's own postcondition does not check (F-1).
var rlsWantExprs = map[string]string{
	"business_profiles_select":        "((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND ((( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text])) OR (EXISTS ( SELECT 1 FROM app.group_business_profiles a WHERE ((a.organization_id = business_profiles.organization_id) AND (a.business_profile_id = business_profiles.id) AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (a.group_id = ( SELECT app.ctx_group() AS ctx_group))) AND (app.rls_group_role(a.group_id) = ANY (ARRAY['group_owner'::text, 'group_admin'::text, 'group_member'::text]))))))) WITH CHECK",
	"business_profiles_update":        "((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND ((( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text])) OR (EXISTS ( SELECT 1 FROM app.group_business_profiles a WHERE ((a.organization_id = business_profiles.organization_id) AND (a.business_profile_id = business_profiles.id) AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (a.group_id = ( SELECT app.ctx_group() AS ctx_group))) AND (app.rls_group_role(a.group_id) = ANY (ARRAY['group_owner'::text, 'group_admin'::text, 'group_member'::text]))))))) WITH CHECK ((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND ((( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text])) OR (( SELECT o.group_admins_edit_business_profiles FROM app.organizations o WHERE (o.id = ( SELECT app.ctx_org() AS ctx_org))) AND (EXISTS ( SELECT 1 FROM app.group_business_profiles a WHERE ((a.organization_id = business_profiles.organization_id) AND (a.business_profile_id = business_profiles.id) AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (a.group_id = ( SELECT app.ctx_group() AS ctx_group))) AND (app.rls_group_role(a.group_id) = ANY (ARRAY['group_owner'::text, 'group_admin'::text]))))))))",
	"group_business_profiles_select":  "((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (group_id = ( SELECT app.ctx_group() AS ctx_group))) AND ((( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text])) OR (app.rls_group_role(group_id) IS NOT NULL))) WITH CHECK",
	"groups_policy_reader":            "((id = app.ctx_group()) AND (organization_id = app.ctx_org())) WITH CHECK",
	"groups_select":                   "((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (id = ( SELECT app.ctx_group() AS ctx_group))) AND ((( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text])) OR (app.rls_group_role(id) IS NOT NULL))) WITH CHECK",
	"groups_update":                   "((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (id = ( SELECT app.ctx_group() AS ctx_group))) AND ((( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text])) OR (app.rls_group_role(id) IS NOT NULL))) WITH CHECK ((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (id = ( SELECT app.ctx_group() AS ctx_group))) AND ((( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text])) OR (app.rls_group_role(id) = ANY (ARRAY['group_owner'::text, 'group_admin'::text]))))",
	"identity_principals_auth_reader": "true WITH CHECK",
	"identity_principals_insert":      "WITH CHECK ((user_id = ( SELECT app.ctx_user() AS ctx_user)) AND (( SELECT app.ctx_org() AS ctx_org) IS NULL))",
	"identity_principals_select":      "(user_id = ( SELECT app.ctx_user() AS ctx_user)) WITH CHECK",
	"login_transactions_delete":       "(binding_hash = ( SELECT app.ctx_login_binding() AS ctx_login_binding)) WITH CHECK",
	"login_transactions_expire":       "((expires_at <= now()) AND (( SELECT app.ctx_user() AS ctx_user) IS NULL) AND (( SELECT app.ctx_login_binding() AS ctx_login_binding) IS NULL)) WITH CHECK",
	"login_transactions_insert":       "WITH CHECK (binding_hash = ( SELECT app.ctx_login_binding() AS ctx_login_binding))",
	"login_transactions_select":       "(binding_hash = ( SELECT app.ctx_login_binding() AS ctx_login_binding)) WITH CHECK",
	"organizations_select":            "((id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL)) WITH CHECK",
	"organizations_update":            "((id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL)) WITH CHECK ((id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) = 'org_owner'::text))",
	"role_bindings_lock":              "((( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND (organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (subject_user_id = ( SELECT app.ctx_user() AS ctx_user)) AND (source = 'membership'::text) AND ((scope_type = 'organization'::text) OR (( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (group_id = ( SELECT app.ctx_group() AS ctx_group)))) WITH CHECK false",
	"role_bindings_policy_reader":     "((subject_user_id = app.ctx_user()) AND (organization_id = app.ctx_org())) WITH CHECK",
	"role_bindings_select":            "((( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND (organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (source = 'membership'::text) AND (((subject_user_id = ( SELECT app.ctx_user() AS ctx_user)) AND ((scope_type = 'organization'::text) OR (( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (group_id = ( SELECT app.ctx_group() AS ctx_group)))) OR ((scope_type = 'organization'::text) AND (( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text]))) OR ((scope_type = 'group'::text) AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (group_id = ( SELECT app.ctx_group() AS ctx_group))) AND ((( SELECT app.rls_org_role() AS rls_org_role) = ANY (ARRAY['org_owner'::text, 'org_admin'::text])) OR (app.rls_group_role(group_id) = ANY (ARRAY['group_owner'::text, 'group_admin'::text])))))) WITH CHECK",
	"sessions_auth_reader":            "true WITH CHECK",
	"sessions_insert":                 "WITH CHECK ((user_id = ( SELECT app.ctx_user() AS ctx_user)) AND (( SELECT app.ctx_org() AS ctx_org) IS NULL) AND (revoked_at IS NULL))",
	"sessions_select":                 "(user_id = ( SELECT app.ctx_user() AS ctx_user)) WITH CHECK",
	"sessions_update":                 "(user_id = ( SELECT app.ctx_user() AS ctx_user)) WITH CHECK ((user_id = ( SELECT app.ctx_user() AS ctx_user)) AND (( SELECT app.ctx_org() AS ctx_org) IS NULL))",
	"users_auth_reader":               "true WITH CHECK",
	"users_insert":                    "WITH CHECK ((id = ( SELECT app.ctx_user() AS ctx_user)) AND (( SELECT app.ctx_org() AS ctx_org) IS NULL) AND (status = 'active'::text) AND (disabled_at IS NULL))",
	"users_select":                    "((id = ( SELECT app.ctx_user() AS ctx_user)) OR ((( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) AND (EXISTS ( SELECT 1 FROM app.role_bindings rb WHERE ((rb.subject_user_id = users.id) AND (rb.organization_id = ( SELECT app.ctx_org() AS ctx_org))))))) WITH CHECK",
	"users_update":                    "(id = ( SELECT app.ctx_user() AS ctx_user)) WITH CHECK ((id = ( SELECT app.ctx_user() AS ctx_user)) AND (( SELECT app.ctx_org() AS ctx_org) IS NULL))",
}
