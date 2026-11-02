//go:build integration

// G075 B-GROUP-PROFILE-CATALOG (Plan r2 §3.1-§3.3, §3.6, §5): TestGroupProfileCatalog against the disposable PostgreSQL
// of the migrations suite. CT-GMP-1 the exact version-23 -> 24 delta on the otherwise unchanged version-23 catalog (the
// Group member profile table with its columns, CHECKs, keys and foreign-key actions (exactly one CASCADE, from the
// membership binding; no other table cascades from app.role_bindings), the guard and the capture, the four policies,
// the policy reader's card projection, the replaced attribution body equal to the reviewed migration, the grants with no
// runtime DELETE and no retention grant, the helper inventories, the audit catalog at 54 actions, 85 reasons, 87
// operations and 14 resource types, 97 relations); CT-GMP-NEG rolled-back missing-protection negatives, each reported
// by exactly its detector parts and, where a statement applies, admitting what the protection refuses (with the intact
// cascade as the positive control); CT-GMP-UPDOWN the down to the exact version-23 catalog when empty, up again, and the
// down refused while a label row or label history exists. Setup uses the attested superuser; catalog reads use the
// owner (search_path pg_catalog).
package migrate

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"testing"
)

// The CT-GMP fixture: gmpMember holds the Group membership gmpBindingID of gmpGroup in gmpOrg; gmpOther is another
// member of gmpOrg without a Group.
const (
	gmpOrg       = "0f075000-0000-4000-8000-000000000001"
	gmpGroup     = "0f075000-0000-4000-8000-000000000201"
	gmpBindingID = "0f075000-0000-4000-8000-000000000301"
	gmpMember    = "0f075000-0000-4000-8000-0000000000a1"
	gmpOther     = "0f075000-0000-4000-8000-0000000000b1"
)

// gmpSeed is the CT-GMP world (attested superuser statements).
const gmpSeed = `INSERT INTO app.users (id, display_name) VALUES ('` + gmpMember + `', 'Catalog Member'), ('` + gmpOther + `', 'Catalog Other');
INSERT INTO app.organizations (id, name) VALUES ('` + gmpOrg + `', 'Group Profiles Catalog');
INSERT INTO app.groups (id, organization_id, name, slug) VALUES ('` + gmpGroup + `', '` + gmpOrg + `', 'Catalog Group', 'gmp-catalog-group');
INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + gmpOrg + `', '` + gmpMember + `', 'organization', '` + gmpOrg + `', NULL, 'org_member', 'membership'),
  ('` + gmpOrg + `', '` + gmpOther + `', 'organization', '` + gmpOrg + `', NULL, 'org_member', 'membership');
INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + gmpBindingID + `', '` + gmpOrg + `', '` + gmpMember + `', 'group', '` + gmpGroup + `', '` + gmpGroup + `', 'group_member', 'membership')`

// gmpWantDelta is the exact set of namesCatalog keys version 24 adds or changes over version 23.
var gmpWantDelta = []string{
	"acl:app.group_member_profiles.binding_id", "acl:app.group_member_profiles.group_id",
	"acl:app.group_member_profiles.organization_id", "acl:app.group_member_profiles.role_label",
	"acl:app.group_member_profiles.updated_at", "acl:app.group_member_profiles.user_id", "acl:app.group_member_profiles.version",
	"col:app.group_member_profiles.binding_id", "col:app.group_member_profiles.group_id",
	"col:app.group_member_profiles.organization_id", "col:app.group_member_profiles.role_label",
	"col:app.group_member_profiles.updated_at", "col:app.group_member_profiles.user_id", "col:app.group_member_profiles.version",
	"con:app.group_member_profiles.group_member_profiles_binding_fkey", "con:app.group_member_profiles.group_member_profiles_group_fkey",
	"con:app.group_member_profiles.group_member_profiles_label_valid",
	"con:app.group_member_profiles.group_member_profiles_one_per_membership",
	"con:app.group_member_profiles.group_member_profiles_organization_fkey", "con:app.group_member_profiles.group_member_profiles_pkey",
	"con:app.group_member_profiles.group_member_profiles_user_fkey", "con:app.group_member_profiles.group_member_profiles_version_valid",
	"con:audit.events.events_action_valid", "con:audit.events.events_operation_valid", "con:audit.events.events_reason_valid",
	"con:audit.events.events_resource_type_valid",
	"fn:app.group_member_profiles_guard()", "fn:app.rls_group_member_cards(uuid)", "fn:audit.capture_group_member_profiles()",
	"fn:audit.events_attribute()",
	"pol:app.group_member_profiles.group_member_profiles_insert", "pol:app.group_member_profiles.group_member_profiles_policy_reader",
	"pol:app.group_member_profiles.group_member_profiles_select", "pol:app.group_member_profiles.group_member_profiles_update",
	"rel:app.group_member_profiles", "rel:app.group_member_profiles_one_per_membership", "rel:app.group_member_profiles_pkey",
	"rls:app.group_member_profiles",
	"trg:app.group_member_profiles.audit_capture_group_member_profiles", "trg:app.group_member_profiles.group_member_profiles_guard",
}

// gmpSources are the version-24 function bodies the detector fingerprints, keyed by part, with the marker that finds
// each one's reviewed text in the 00024 up section.
var gmpSources = map[string]string{
	"gmp-guard-src":     "CREATE FUNCTION app.group_member_profiles_guard()",
	"gmp-capture-src":   "CREATE FUNCTION audit.capture_group_member_profiles()",
	"gmp-cards-src":     "CREATE FUNCTION app.rls_group_member_cards(",
	"gmp-attribute-src": "CREATE OR REPLACE FUNCTION audit.events_attribute()",
}

// gmpSignatures are the regprocedure inputs of the fingerprinted bodies.
var gmpSignatures = map[string]string{
	"gmp-guard-src":     "app.group_member_profiles_guard()",
	"gmp-capture-src":   "audit.capture_group_member_profiles()",
	"gmp-cards-src":     "app.rls_group_member_cards(uuid)",
	"gmp-attribute-src": "audit.events_attribute()",
}

// gmpParts reads the version-24 delta's catalog parts. Roles are symbolic in gmp-functions and gmp-helpers (rt, ow,
// pr, ar, cw, pw, public) and by name elsewhere. Policy lines separate their fields with '|' (the expressions hold casts).
func gmpParts(t *testing.T, q catalogQuerier, cfg Config) map[string]string {
	t.Helper()
	one := func(sql string, args ...any) string {
		t.Helper()
		var s *string
		if err := q.QueryRow(context.Background(), sql, args...).Scan(&s); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		if s == nil {
			return ""
		}
		return *s
	}
	const table = `to_regclass('app.group_member_profiles')`
	p := map[string]string{}
	p["gmp-columns"] = one(`SELECT string_agg(format('%s.%s:%s:%s:%s:%s', c.relname, a.attname, format_type(a.atttypid, a.atttypmod),
  a.attnotnull, a.attgenerated, a.atthasdef), ';' ORDER BY a.attnum)
FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE a.attrelid = ` + table + ` AND a.attnum > 0 AND NOT a.attisdropped`)
	p["gmp-constraints"] = one(`SELECT string_agg(conname || ':' || pg_get_constraintdef(oid), ';' ORDER BY conname) FROM pg_constraint
WHERE conrelid = ` + table)
	// Every cascading foreign key onto app.role_bindings (the membership key only).
	p["gmp-cascades"] = one(`SELECT string_agg(conrelid::regclass::text || '.' || conname, ';' ORDER BY conname) FROM pg_constraint
WHERE confrelid = 'app.role_bindings'::regclass AND contype = 'f' AND confdeltype = 'c'`)
	p["gmp-indexes"] = one(`SELECT string_agg(pg_get_indexdef(i.indexrelid), ';' ORDER BY c.relname) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
WHERE i.indrelid = ` + table)
	p["gmp-triggers"] = one(`SELECT string_agg(t.tgname || ':' || t.tgenabled::text || ':' || pg_get_triggerdef(t.oid), E'\n' ORDER BY t.tgname)
FROM pg_trigger t WHERE t.tgrelid = ` + table + ` AND NOT t.tgisinternal`)
	p["gmp-flags"] = one(`SELECT string_agg(relname || ':' || relrowsecurity::text || ':' || relforcerowsecurity::text, ';' ORDER BY relname)
FROM pg_class WHERE oid = ` + table)
	p["gmp-policies"] = one(`SELECT string_agg(p.polname || '|' || p.polcmd::text || '|' || p.polpermissive::text || '|' ||
  array_to_string(ARRAY(SELECT r::regrole::text FROM unnest(p.polroles) r ORDER BY 1), '+') || '|' ||
  regexp_replace(coalesce(pg_get_expr(p.polqual, p.polrelid), ''), '\s+', ' ', 'g') || '|' ||
  regexp_replace(coalesce(pg_get_expr(p.polwithcheck, p.polrelid), ''), '\s+', ' ', 'g'), E'\n' ORDER BY p.polname)
FROM pg_policy p WHERE p.polrelid = ` + table)
	// Every table and column privilege on the table other than the owner's own.
	p["gmp-acl"] = one(`SELECT string_agg(x, ';' ORDER BY x COLLATE "C") FROM (
  SELECT c.relname || ':' || a.grantee::regrole::text || ':' || a.privilege_type AS x
  FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
  WHERE c.oid = ` + table + ` AND a.grantee <> c.relowner
  UNION ALL
  SELECT c.relname || '.' || at.attname || ':' || a.grantee::regrole::text || ':' || a.privilege_type
  FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
  CROSS JOIN LATERAL aclexplode(at.attacl) a
  WHERE c.oid = ` + table + ` AND a.grantee <> c.relowner) s`)
	lbl := func(col string) string {
		return `CASE ` + col + ` WHEN 0 THEN 'public' WHEN $1::regrole THEN 'rt' WHEN $2::regrole THEN 'ow' WHEN $3::regrole THEN 'pr'
  WHEN $4::regrole THEN 'ar' WHEN $5::regrole THEN 'cw' WHEN $6::regrole THEN 'pw' ELSE 'other' END`
	}
	db := cfg.Database
	roles := []any{cfg.RuntimeRole(), cfg.OwnerRole(), db + "_policy_reader", db + "_auth_reader", db + "_claim_writer", db + "_provision_writer"}
	p["gmp-functions"] = one(`SELECT string_agg(p.proname || ':' || `+lbl("p.proowner")+` || ':' || p.prosecdef::text || ':' || p.provolatile::text || ':' ||
  coalesce(array_to_string(p.proconfig, '|'), '-') || ':' || coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C")
  FROM (SELECT `+lbl("a.grantee")+` || '=' || a.privilege_type AS v FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'),
  ';' ORDER BY p.proname COLLATE "C")
FROM pg_proc p WHERE p.oid IN (to_regprocedure('app.group_member_profiles_guard()'), to_regprocedure('audit.capture_group_member_profiles()'),
  to_regprocedure('audit.events_attribute()'), to_regprocedure('app.rls_group_member_cards(uuid)'))`, roles...)
	for part, sig := range gmpSignatures {
		p[part] = one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure($1)`, sig)
	}
	// The helper inventories (by owner) and their CREATE on schema app.
	p["gmp-helpers"] = one(`SELECT string_agg(x, ';' ORDER BY x COLLATE "C") FROM (
  SELECT `+lbl("p.proowner")+` || '=' || string_agg(p.proname, ',' ORDER BY p.proname COLLATE "C") AS x
  FROM pg_proc p WHERE p.proowner IN ($3::regrole, $4::regrole, $5::regrole, $6::regrole) GROUP BY p.proowner
  UNION ALL
  SELECT 'create=' || has_schema_privilege($3::regrole, 'app', 'CREATE')::text || ',' || has_schema_privilege($4::regrole, 'app', 'CREATE')::text) s`,
		roles...)
	// The four audit catalog constraints by their value counts (one '::text' per value, the 00024 postcondition's rule).
	p["gmp-catalog"] = one(`SELECT string_agg(s.conname || '=' || ((length(s.def) - length(replace(s.def, '::text', ''))) / length('::text'))::text, ';'
  ORDER BY s.conname COLLATE "C")
FROM (SELECT c.conname::text AS conname, pg_get_constraintdef(c.oid) AS def FROM pg_constraint c
  WHERE c.conrelid = 'audit.events'::regclass
    AND c.conname IN ('events_action_valid', 'events_reason_valid', 'events_operation_valid', 'events_resource_type_valid')) s`)
	return p
}

// gmpViolations is the version-24 delta detector: the sorted parts that differ from want.
func gmpViolations(t *testing.T, q catalogQuerier, cfg Config, want map[string]string) []string {
	t.Helper()
	p := gmpParts(t, q, cfg)
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

func TestGroupProfileCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r23 := runnerFor(t, cfg, subsetFS(t, 23))
	objRunUp(t, r23, "current_version\t23\n")
	s23 := snapshot(t, r23)
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	n23 := namesCatalog(t, cat)
	r := runnerFor(t, cfg, subsetFS(t, 24))
	objRunUp(t, r, "applied\t00024_group_member_profiles.sql\ncurrent_version\t24\n")
	n24 := namesCatalog(t, cat)
	clean := gmpParts(t, cat, cfg)
	o, rt, db := cfg.OwnerRole(), cfg.RuntimeRole(), cfg.Database
	pr := db + "_policy_reader"

	t.Run("CT-GMP-1 the exact version-24 delta on the unchanged version-23 catalog", func(t *testing.T) {
		if got := namesDelta(n23, n24); !slices.Equal(got, gmpWantDelta) {
			t.Errorf("version-23 -> 24 delta:\n got %v\nwant %v", got, gmpWantDelta)
		}
		metadataThroughVersion24(t, cfg, r)
		for part, marker := range gmpSources {
			if clean[part] != md5hex(migBody(t, "00024_group_member_profiles.sql", marker)) {
				t.Errorf("%s differs from the reviewed migration body", part)
			}
		}
		if got, want := clean["gmp-columns"], "group_member_profiles.binding_id:uuid:t::f;group_member_profiles.organization_id:uuid:t::f;"+
			"group_member_profiles.group_id:uuid:t::f;group_member_profiles.user_id:uuid:t::f;group_member_profiles.role_label:text:f::f;"+
			"group_member_profiles.version:integer:t::t;group_member_profiles.updated_at:timestamp with time zone:t::t"; got != want {
			t.Errorf("columns:\n got %s\nwant %s", got, want)
		}
		if got := clean["gmp-flags"]; got != "group_member_profiles:true:true" {
			t.Errorf("row-level security flags: %s, want enabled and forced", got)
		}
		cons := clean["gmp-constraints"]
		var names []string
		for _, c := range strings.Split(cons, ";") {
			n, _, _ := strings.Cut(c, ":")
			names = append(names, n)
		}
		if got := strings.Join(names, ","); got != "group_member_profiles_binding_fkey,group_member_profiles_group_fkey,"+
			"group_member_profiles_label_valid,group_member_profiles_one_per_membership,group_member_profiles_organization_fkey,"+
			"group_member_profiles_pkey,group_member_profiles_user_fkey,group_member_profiles_version_valid" {
			t.Errorf("constraints: %s", got)
		}
		// Exactly one cascading foreign key, from the membership binding (the label goes with the membership); the
		// Organization, Group and user keys restrict; the label bound and grammar are the defensive CHECK.
		for _, want := range []string{
			"group_member_profiles_binding_fkey:FOREIGN KEY (binding_id) REFERENCES app.role_bindings(id) ON DELETE CASCADE",
			"group_member_profiles_group_fkey:FOREIGN KEY (organization_id, group_id) REFERENCES app.groups(organization_id, id) ON DELETE RESTRICT",
			"group_member_profiles_organization_fkey:FOREIGN KEY (organization_id) REFERENCES app.organizations(id) ON DELETE RESTRICT",
			"group_member_profiles_user_fkey:FOREIGN KEY (user_id) REFERENCES app.users(id) ON DELETE RESTRICT",
			"group_member_profiles_one_per_membership:UNIQUE (organization_id, user_id, group_id)",
			"group_member_profiles_pkey:PRIMARY KEY (binding_id)",
		} {
			if !strings.Contains(cons, want) {
				t.Errorf("constraints lack %q", want)
			}
		}
		if strings.Count(cons, "ON DELETE CASCADE") != 1 || strings.Count(cons, "ON DELETE RESTRICT") != 3 ||
			!strings.Contains(cons, "btrim(role_label)") || !strings.Contains(cons, `\u2028`) ||
			!strings.Contains(cons, "BETWEEN 1 AND 100") && !strings.Contains(cons, "<= 100)") {
			t.Errorf("constraints: the foreign-key actions, the label bound or the label grammar differ: %s", cons)
		}
		if got := clean["gmp-cascades"]; got != "app.group_member_profiles.group_member_profiles_binding_fkey" {
			t.Errorf("cascading foreign keys onto app.role_bindings: %s, want the membership key only", got)
		}
		if got := clean["gmp-indexes"]; got != "CREATE UNIQUE INDEX group_member_profiles_one_per_membership ON app.group_member_profiles "+
			"USING btree (organization_id, user_id, group_id);"+
			"CREATE UNIQUE INDEX group_member_profiles_pkey ON app.group_member_profiles USING btree (binding_id)" {
			t.Errorf("indexes: %s", got)
		}
		trg := clean["gmp-triggers"]
		if strings.Count(trg, "\n") != 1 ||
			!strings.Contains(trg, "audit_capture_group_member_profiles:O:CREATE TRIGGER audit_capture_group_member_profiles AFTER INSERT OR UPDATE OF role_label ON app.group_member_profiles FOR EACH ROW EXECUTE FUNCTION audit.capture_group_member_profiles()") ||
			!strings.Contains(trg, "group_member_profiles_guard:O:CREATE TRIGGER group_member_profiles_guard BEFORE INSERT OR UPDATE ON app.group_member_profiles FOR EACH ROW EXECUTE FUNCTION app.group_member_profiles_guard()") {
			t.Errorf("triggers (no DELETE capture):\n%s", trg)
		}
		// The four permissive policies: the two writer paths (the member's own row; a current Group owner/admin of that
		// Group), in the context Organization with a current Organization role and the selected Group only, for the
		// runtime and the owner; USING (true) for the policy reader only. None reads the profile predicate or the override.
		writerRoles := []string{o, rt}
		slices.Sort(writerRoles)
		writers := strings.Join(writerRoles, "+")
		pols := strings.Split(clean["gmp-policies"], "\n")
		if len(pols) != 4 {
			t.Fatalf("policies: %q", clean["gmp-policies"])
		}
		for _, l := range pols {
			f := strings.SplitN(l, "|", 6)
			if len(f) != 6 {
				t.Errorf("policy line: %q", l)
				continue
			}
			name, cmd, permissive, polRoles, qual, check := f[0], f[1], f[2], f[3], f[4], f[5]
			if permissive != "true" {
				t.Errorf("policy %s is restrictive", name)
			}
			switch name {
			case "group_member_profiles_policy_reader":
				if cmd != "r" || polRoles != pr || qual != "true" || check != "" {
					t.Errorf("helper policy %s: %s", name, l)
				}
			case "group_member_profiles_select", "group_member_profiles_insert", "group_member_profiles_update":
				wantCmd := map[string]string{"group_member_profiles_select": "r", "group_member_profiles_insert": "a",
					"group_member_profiles_update": "w"}[name]
				if cmd != wantCmd || polRoles != writers {
					t.Errorf("writer policy %s lacks its command or roles: %s", name, l)
				}
				for _, side := range []struct{ what, expr string }{{"USING", qual}, {"WITH CHECK", check}} {
					if side.expr == "" {
						continue
					}
					for _, part := range []string{"( SELECT app.ctx_org()", "( SELECT app.rls_org_role()", "( SELECT app.ctx_group()",
						"user_id = ( SELECT app.ctx_user()", "app.rls_group_role(group_id)", "'group_owner'", "'group_admin'"} {
						if !strings.Contains(side.expr, part) {
							t.Errorf("policy %s %s lacks %q: %s", name, side.what, part, side.expr)
						}
					}
					if strings.Contains(side.expr, "rls_profile_visible") || strings.Contains(side.expr, "override") {
						t.Errorf("policy %s %s reads the profile predicate or the override: %s", name, side.what, side.expr)
					}
				}
				if (name == "group_member_profiles_select") != (check == "") || (name == "group_member_profiles_insert") != (qual == "") ||
					(name == "group_member_profiles_update" && qual != check) {
					t.Errorf("policy %s has the wrong USING / WITH CHECK shape: %s", name, l)
				}
			default:
				t.Errorf("unexpected policy %s", l)
			}
		}
		acl := []string{"group_member_profiles:" + rt + ":SELECT"}
		for _, c := range []string{"binding_id", "organization_id", "group_id", "user_id", "role_label", "version", "updated_at"} {
			acl = append(acl, "group_member_profiles."+c+":"+rt+":INSERT")
		}
		for _, c := range []string{"role_label", "version", "updated_at"} {
			acl = append(acl, "group_member_profiles."+c+":"+rt+":UPDATE")
		}
		for _, c := range []string{"binding_id", "organization_id", "group_id", "user_id", "role_label", "version"} {
			acl = append(acl, "group_member_profiles."+c+":"+pr+":SELECT")
		}
		slices.Sort(acl)
		if got := clean["gmp-acl"]; got != strings.Join(acl, ";") {
			t.Errorf("privileges:\n got %s\nwant %s", got, strings.Join(acl, ";"))
		}
		one := func(sql string, args ...any) string {
			t.Helper()
			var s string
			if err := cat.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
				t.Fatalf("catalog: %v", err)
			}
			return s
		}
		// No DELETE, TRUNCATE, REFERENCES or TRIGGER for the runtime (the cascade runs as the table owner), and nothing at
		// all for the retention role, the migrator, the auth reader, the claim and provision writers or PUBLIC.
		if got := one(`SELECT has_table_privilege($1, 'app.group_member_profiles', 'DELETE, TRUNCATE, REFERENCES, TRIGGER')::text`, rt); got != "false" {
			t.Error("runtime holds DELETE, TRUNCATE, REFERENCES or TRIGGER on app.group_member_profiles")
		}
		for _, role := range []string{cfg.RetentionRole(), cfg.MigratorRole(), db + "_auth_reader", db + "_claim_writer", db + "_provision_writer", "public"} {
			if got := one(`SELECT (has_table_privilege($1, 'app.group_member_profiles', 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER')
  OR has_any_column_privilege($1, 'app.group_member_profiles', 'SELECT, INSERT, UPDATE, REFERENCES'))::text`, role); got != "false" {
				t.Errorf("%s holds a privilege on app.group_member_profiles", role)
			}
		}
		const sp = "search_path=pg_catalog, pg_temp"
		if got, want := clean["gmp-functions"], "capture_group_member_profiles:ow:false:v:"+sp+":-;events_attribute:ow:false:v:"+sp+":-;"+
			"group_member_profiles_guard:ow:false:v:"+sp+":-;rls_group_member_cards:pr:true:s:"+sp+":ow=EXECUTE+rt=EXECUTE"; got != want {
			t.Errorf("functions:\n got %s\nwant %s", got, want)
		}
		if got, want := clean["gmp-helpers"], "ar=auth_logout_ref,auth_principal_subject,auth_resolve_enrollment,auth_resolve_session,"+
			"auth_resolve_signup,hostname_state,is_organization_member,organization_id_by_slug,public_branding,signup_slug_taken;"+
			"create=false,false;"+
			"cw=enrollment_admit,enrollment_lock_organization,enrollment_request,invitation_claim_group,invitation_claim_organization,"+
			"rls_invitation_pending,rls_user_enabled_lock;"+
			"pr=rls_account_profile,rls_dependent_group_bindings,rls_enrollment_mode,rls_group_invitation_candidates,rls_group_member_cards,"+
			"rls_group_role,rls_org_role,rls_own_organizations,rls_owner_count,rls_profile_eligible,rls_profile_visible,rls_sole_owner_scopes;"+
			"pw=signup_challenge_complete,signup_challenge_issue,signup_provision_organization"; got != want {
			t.Errorf("helper inventories:\n got %s\nwant %s", got, want)
		}
		// The projection's result columns: no Group name (joined by the caller under its own row-level security), no
		// permission role, no other member.
		if got := one(`SELECT pg_get_function_result('app.rls_group_member_cards(uuid)'::regprocedure)`); got !=
			"TABLE(group_id uuid, role_label text, can_edit boolean, version integer)" {
			t.Errorf("projection result columns: %s", got)
		}
		// The audit catalog: 54 actions, 85 reasons, 87 operations and 14 resource types, naming the Group member profile
		// action, reason, the two operations and the resource type.
		if got := clean["gmp-catalog"]; got != "events_action_valid=54;events_operation_valid=87;events_reason_valid=85;events_resource_type_valid=14" {
			t.Errorf("audit catalog counts: %s", got)
		}
		for con, vals := range map[string][]string{
			"events_action_valid":        {"'group_member_profile.updated'"},
			"events_reason_valid":        {"'role_label_changed'"},
			"events_operation_valid":     {"'profile.groups.list'", "'profile.group_label.update'"},
			"events_resource_type_valid": {"'group_member_profile'"},
		} {
			def := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'audit.events'::regclass AND conname = $1`, con)
			for _, v := range vals {
				if !strings.Contains(def, v) {
					t.Errorf("%s lacks %s", con, v)
				}
			}
		}
		if got := one(`SELECT count(*)::text FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.oid <> 'app.goose_db_version'::regclass
  AND c.oid IS DISTINCT FROM pg_get_serial_sequence('app.goose_db_version', 'id')::regclass
  AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass)`); got != "97" {
			t.Errorf("schema app holds %s relations, want 97", got)
		}
	})

	t.Run("CT-GMP-NEG rolled-back missing-protection negatives", func(t *testing.T) {
		// The probe world (superuser fixture, inside the probe's rolled-back transaction): gmpMember's Group membership
		// and gmpOther without one. The rows carry no label, so the capture records nothing and no context is needed.
		forged := `INSERT INTO app.group_member_profiles (binding_id, organization_id, group_id, user_id) VALUES ('` + gmpBindingID + `', '` +
			gmpOrg + `', '` + gmpGroup + `', '` + gmpOther + `')`
		cascade := `DO $$ BEGIN
  INSERT INTO app.group_member_profiles (binding_id, organization_id, group_id, user_id) VALUES ('` + gmpBindingID + `', '` + gmpOrg + `', '` +
			gmpGroup + `', '` + gmpMember + `');
  DELETE FROM app.role_bindings WHERE id = '` + gmpBindingID + `';
  IF EXISTS (SELECT 1 FROM app.group_member_profiles WHERE binding_id = '` + gmpBindingID + `') THEN
    RAISE EXCEPTION 'the label survived its membership';
  END IF;
END $$`
		probe := func(id, mutation, stmt string) error {
			t.Helper()
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			for _, s := range []string{mutation, gmpSeed} {
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
		// The intact protections: the guard refuses a row whose user is not the binding's subject; the positive control is
		// the cascade (a removed membership takes its label row with it).
		if err := probe("intact", "", forged); sqlstateOf(err) != "23514" {
			t.Fatalf("intact protections: the forged row %v, want 23514", err)
		}
		if err := probe("intact", "", cascade); err != nil {
			t.Fatalf("intact protections: the membership removal kept its label row (%v)", err)
		}
		cardsHead := `CREATE OR REPLACE FUNCTION app.rls_group_member_cards(p_subject uuid)
  RETURNS TABLE (group_id uuid, role_label text, can_edit boolean, version integer)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $f$
#variable_conflict use_column
BEGIN
`
		for _, n := range []struct {
			id, sql string
			want    []string
			stmts   []string // statements the missing protection now admits
		}{
			{"N1 row-level security not forced", `ALTER TABLE app.group_member_profiles NO FORCE ROW LEVEL SECURITY`, []string{"gmp-flags"}, nil},
			{"N2 runtime DELETE", `GRANT DELETE ON app.group_member_profiles TO ` + quote(rt), []string{"gmp-acl"}, nil},
			{"N3 projection executable by PUBLIC", `GRANT EXECUTE ON FUNCTION app.rls_group_member_cards(uuid) TO PUBLIC`, []string{"gmp-functions"}, nil},
			{"N4 projection without the profile-predicate gate", cardsHead + `  RETURN QUERY SELECT sb.scope_id, NULL::text, false, 0 FROM app.role_bindings sb
  WHERE sb.organization_id = app.ctx_org() AND sb.scope_type = 'group' AND sb.subject_user_id = p_subject AND sb.source = 'membership'
  ORDER BY sb.scope_id LIMIT 201;
END $f$`, []string{"gmp-cards-src"}, nil},
			{"N5 projection without the 201-row bound", cardsHead + `  IF p_subject IS NULL OR NOT app.rls_profile_visible(p_subject) THEN
    RETURN;
  END IF;
  RETURN QUERY SELECT sb.scope_id, NULL::text, false, 0 FROM app.role_bindings sb
  WHERE sb.organization_id = app.ctx_org() AND sb.scope_type = 'group' AND sb.subject_user_id = p_subject AND sb.source = 'membership'
  ORDER BY sb.scope_id;
END $f$`, []string{"gmp-cards-src"}, nil},
			{"N6 write policy without the Group-role branch", `ALTER POLICY group_member_profiles_insert ON app.group_member_profiles
  WITH CHECK (organization_id = (SELECT app.ctx_org()))`, []string{"gmp-policies"}, nil},
			{"N7 membership key without CASCADE", `ALTER TABLE app.group_member_profiles DROP CONSTRAINT group_member_profiles_binding_fkey,
  ADD CONSTRAINT group_member_profiles_binding_fkey FOREIGN KEY (binding_id) REFERENCES app.role_bindings (id)`,
				[]string{"gmp-cascades", "gmp-constraints"}, nil},
			{"N8 guard without the binding check", `CREATE OR REPLACE FUNCTION app.group_member_profiles_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $f$ BEGIN NEW.updated_at := now(); RETURN NEW; END $f$`,
				[]string{"gmp-guard-src"}, []string{forged}},
			{"N9 retention grant on the labels", `GRANT SELECT ON app.group_member_profiles TO ` + quote(cfg.RetentionRole()), []string{"gmp-acl"}, nil},
			{"N10 capture without the audit row", `CREATE OR REPLACE FUNCTION audit.capture_group_member_profiles() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $f$ BEGIN RETURN NULL; END $f$`, []string{"gmp-capture-src"}, nil},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			v := gmpViolations(t, tx, cfg, clean)
			_ = tx.Rollback(ctx)
			if !slices.Equal(v, n.want) {
				t.Errorf("%s: the detector reported %v, want %v", n.id, v, n.want)
			}
			for _, stmt := range n.stmts {
				if err := probe(n.id, n.sql, stmt); err != nil {
					t.Errorf("%s: the missing protection still refused the probe: %v", n.id, err)
				}
			}
		}
		if v := gmpViolations(t, cat, cfg, clean); len(v) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", v)
		}
	})

	t.Run("CT-GMP-UPDOWN exact version-23 catalog when empty, up again, down refused with a label row or label history", func(t *testing.T) {
		s24 := snapshot(t, r)
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
		if err != nil || out != "rolled_back\t00024_group_member_profiles.sql\ncurrent_version\t23\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		metadataThroughVersion23(t, cfg, r)
		if got := snapshot(t, r); got != s23 {
			t.Fatalf("catalog after down differs from version 23:\n%s\n---\n%s", got, s23)
		}
		if got := namesDelta(n23, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down differs from version 23: %v", got)
		}
		objRunUp(t, r, "applied\t00024_group_member_profiles.sql\ncurrent_version\t24\n")
		if got := snapshot(t, r); got != s24 {
			t.Fatalf("catalog after down/up differs:\n%s\n---\n%s", got, s24)
		}
		if got := namesDelta(n24, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down/up differs from version 24: %v", got)
		}
		if v := gmpViolations(t, cat, cfg, clean); len(v) != 0 {
			t.Fatalf("version-24 detector after up again: %v", v)
		}
		refused := func(what string) {
			t.Helper()
			if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00024:") {
				t.Fatalf("down with %s: %v %q", what, err, out)
			}
			if got := snapshot(t, r); got != s24 {
				t.Fatalf("the refused down (%s) changed the catalog", what)
			}
		}
		exec(gmpSeed)
		// A label row without a label (the capture records nothing for it) refuses the down.
		exec(`INSERT INTO app.group_member_profiles (binding_id, organization_id, group_id, user_id) VALUES ('` + gmpBindingID + `', '` +
			gmpOrg + `', '` + gmpGroup + `', '` + gmpMember + `')`)
		refused("a label row")
		exec(`DELETE FROM app.group_member_profiles WHERE binding_id = '` + gmpBindingID + `'`)
		// And label history is never discarded: a label's group_member_profile.updated row still refuses with the row gone.
		func() {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			for _, s := range []string{
				`SELECT set_config('app.user_id', '` + gmpMember + `', true), set_config('app.organization_id', '` + gmpOrg + `', true),
  set_config('app.group_id', '', true)`,
				`INSERT INTO app.group_member_profiles (binding_id, organization_id, group_id, user_id, role_label) VALUES ('` + gmpBindingID + `', '` +
					gmpOrg + `', '` + gmpGroup + `', '` + gmpMember + `', 'Catalog History')`,
				`DELETE FROM app.group_member_profiles WHERE binding_id = '` + gmpBindingID + `'`,
			} {
				if _, err := tx.Exec(ctx, s); err != nil {
					t.Fatalf("history fixture: %v", err)
				}
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatalf("history fixture: %v", err)
			}
		}()
		var hist int
		if err := setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'group_member_profile.updated' AND reason = 'role_label_changed'
  AND subject_user_id = $1::uuid AND organization_id = $2::uuid AND group_id = $3::uuid AND resource_id = $4::uuid`,
			gmpMember, gmpOrg, gmpGroup, gmpBindingID).Scan(&hist); err != nil || hist != 1 {
			t.Fatalf("label history rows: %d (%v), want 1", hist, err)
		}
		refused("label history")
		reset(t, cfg)
	})
}
