//go:build integration

// G074 B-PROFILE-CATALOG (Plan r3.1 §3.1-§3.3, §5): TestProfileCatalog against the disposable PostgreSQL of the
// migrations suite. CT-PRO-1 the exact version-22 -> 23 delta on the otherwise unchanged version-22 catalog (the two
// tables with their columns, CHECKs, keys and index, the two guards and the capture, the nine policies, the policy
// reader's two new readers, the replaced invitation-candidate, E1 and attribution bodies equal to the reviewed migration,
// the grants with no runtime DELETE and no retention grant, the helper inventories, 94 relations); CT-PRO-NEG
// rolled-back missing-protection negatives, each reported by exactly its detector part and, where a statement applies,
// admitting what the protection refuses; CT-PRO-UPDOWN the down to the exact version-22 catalog when empty, up again,
// and the down refused while a profile row, an avatar row or profile history exists. Setup uses the attested superuser;
// catalog reads use the owner (search_path pg_catalog).
package migrate

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"testing"
)

// The CT-PRO fixture: two users (proA owns proAvatarReady and proAvatarPending, proB owns proAvatarB).
const (
	proA             = "0f074000-0000-4000-8000-0000000000a1"
	proB             = "0f074000-0000-4000-8000-0000000000b1"
	proAvatarReady   = "0f074000-0000-4000-8000-000000000c01"
	proAvatarPending = "0f074000-0000-4000-8000-000000000c02"
	proAvatarB       = "0f074000-0000-4000-8000-000000000c03"
)

// proWantDelta is the exact set of namesCatalog keys version 23 adds or changes over version 22.
var proWantDelta = []string{
	"acl:app.account_avatars.byte_size", "acl:app.account_avatars.content_type", "acl:app.account_avatars.id",
	"acl:app.account_avatars.sha256", "acl:app.account_avatars.state", "acl:app.account_avatars.upload_started_at",
	"acl:app.account_avatars.user_id",
	"acl:app.account_profiles.avatar_object_id", "acl:app.account_profiles.avatar_version", "acl:app.account_profiles.name",
	"acl:app.account_profiles.updated_at", "acl:app.account_profiles.user_id", "acl:app.account_profiles.version",
	"col:app.account_avatars.byte_size", "col:app.account_avatars.content_type", "col:app.account_avatars.created_at",
	"col:app.account_avatars.deleted_at", "col:app.account_avatars.deleting_at", "col:app.account_avatars.id",
	"col:app.account_avatars.ready_at", "col:app.account_avatars.sha256", "col:app.account_avatars.state",
	"col:app.account_avatars.storage_key", "col:app.account_avatars.upload_started_at", "col:app.account_avatars.user_id",
	"col:app.account_profiles.avatar_object_id", "col:app.account_profiles.avatar_version", "col:app.account_profiles.name",
	"col:app.account_profiles.updated_at", "col:app.account_profiles.user_id", "col:app.account_profiles.version",
	"con:app.account_avatars.account_avatars_id_nonzero", "con:app.account_avatars.account_avatars_lifecycle_check",
	"con:app.account_avatars.account_avatars_pkey", "con:app.account_avatars.account_avatars_sha256_valid",
	"con:app.account_avatars.account_avatars_size_valid", "con:app.account_avatars.account_avatars_state_valid",
	"con:app.account_avatars.account_avatars_type_valid", "con:app.account_avatars.account_avatars_user_fkey",
	"con:app.account_profiles.account_profiles_avatar_fkey", "con:app.account_profiles.account_profiles_avatar_shape",
	"con:app.account_profiles.account_profiles_avatar_version_valid", "con:app.account_profiles.account_profiles_name_valid",
	"con:app.account_profiles.account_profiles_pkey", "con:app.account_profiles.account_profiles_user_fkey",
	"con:app.account_profiles.account_profiles_version_valid",
	"con:audit.events.events_action_valid", "con:audit.events.events_operation_valid", "con:audit.events.events_reason_valid",
	"con:audit.events.events_resource_type_valid",
	"fn:app.account_avatars_guard()", "fn:app.account_profiles_guard()",
	"fn:app.auth_resolve_session(bytea,timestamp with time zone,text)", "fn:app.rls_account_profile(uuid)",
	"fn:app.rls_group_invitation_candidates(uuid,uuid,integer)", "fn:app.rls_profile_visible(uuid)",
	"fn:audit.capture_account_profiles()", "fn:audit.events_attribute()",
	"pol:app.account_avatars.account_avatars_insert", "pol:app.account_avatars.account_avatars_policy_reader",
	"pol:app.account_avatars.account_avatars_select", "pol:app.account_avatars.account_avatars_update",
	"pol:app.account_profiles.account_profiles_auth_reader", "pol:app.account_profiles.account_profiles_insert",
	"pol:app.account_profiles.account_profiles_policy_reader", "pol:app.account_profiles.account_profiles_select",
	"pol:app.account_profiles.account_profiles_update",
	"rel:app.account_avatars", "rel:app.account_avatars_pkey", "rel:app.account_avatars_user_state_idx",
	"rel:app.account_profiles", "rel:app.account_profiles_pkey",
	"rls:app.account_avatars", "rls:app.account_profiles",
	"trg:app.account_avatars.account_avatars_guard", "trg:app.account_profiles.account_profiles_guard",
	"trg:app.account_profiles.audit_capture_account_profiles",
}

// proSources are the version-23 function bodies the detector fingerprints, keyed by part, with the marker that finds
// each one's reviewed text in the 00023 up section.
var proSources = map[string]string{
	"pro-avatars-guard-src":  "CREATE FUNCTION app.account_avatars_guard()",
	"pro-profiles-guard-src": "CREATE FUNCTION app.account_profiles_guard()",
	"pro-capture-src":        "CREATE FUNCTION audit.capture_account_profiles()",
	"pro-visible-src":        "CREATE FUNCTION app.rls_profile_visible(",
	"pro-projection-src":     "CREATE FUNCTION app.rls_account_profile(",
	"pro-picker-src":         "CREATE OR REPLACE FUNCTION app.rls_group_invitation_candidates(",
	"pro-e1-src":             "CREATE OR REPLACE FUNCTION app.auth_resolve_session(",
	"pro-attribute-src":      "CREATE OR REPLACE FUNCTION audit.events_attribute()",
}

// proSignatures are the regprocedure inputs of the fingerprinted bodies.
var proSignatures = map[string]string{
	"pro-avatars-guard-src":  "app.account_avatars_guard()",
	"pro-profiles-guard-src": "app.account_profiles_guard()",
	"pro-capture-src":        "audit.capture_account_profiles()",
	"pro-visible-src":        "app.rls_profile_visible(uuid)",
	"pro-projection-src":     "app.rls_account_profile(uuid)",
	"pro-picker-src":         "app.rls_group_invitation_candidates(uuid, uuid, integer)",
	"pro-e1-src":             "app.auth_resolve_session(bytea, timestamptz, text)",
	"pro-attribute-src":      "audit.events_attribute()",
}

// proParts reads the version-23 delta's catalog parts. Roles are symbolic in pro-functions (rt, ow, pr, ar, cw, pw,
// public) and by name elsewhere.
func proParts(t *testing.T, q catalogQuerier, cfg Config) map[string]string {
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
	const tables = `(to_regclass('app.account_avatars'), to_regclass('app.account_profiles'))`
	p := map[string]string{}
	p["pro-columns"] = one(`SELECT string_agg(format('%s.%s:%s:%s:%s:%s', c.relname, a.attname, format_type(a.atttypid, a.atttypmod),
  a.attnotnull, a.attgenerated, a.atthasdef), ';' ORDER BY c.relname, a.attnum)
FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
WHERE a.attrelid IN ` + tables + ` AND a.attnum > 0 AND NOT a.attisdropped`)
	p["pro-constraints"] = one(`SELECT string_agg(conname || ':' || pg_get_constraintdef(oid), ';' ORDER BY conname) FROM pg_constraint
WHERE conrelid IN ` + tables)
	p["pro-indexes"] = one(`SELECT string_agg(pg_get_indexdef(i.indexrelid), ';' ORDER BY c.relname) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
WHERE i.indrelid IN ` + tables)
	p["pro-triggers"] = one(`SELECT string_agg(t.tgname || ':' || t.tgenabled::text || ':' || pg_get_triggerdef(t.oid), E'\n' ORDER BY t.tgname)
FROM pg_trigger t WHERE t.tgrelid IN ` + tables + ` AND NOT t.tgisinternal`)
	p["pro-flags"] = one(`SELECT string_agg(relname || ':' || relrowsecurity::text || ':' || relforcerowsecurity::text, ';' ORDER BY relname)
FROM pg_class WHERE oid IN ` + tables)
	p["pro-policies"] = one(`SELECT string_agg(p.polname || ':' || p.polcmd::text || ':' || p.polpermissive::text || ':' ||
  array_to_string(ARRAY(SELECT r::regrole::text FROM unnest(p.polroles) r ORDER BY 1), '+') || ':' ||
  regexp_replace(coalesce(pg_get_expr(p.polqual, p.polrelid), ''), '\s+', ' ', 'g') || ':' ||
  regexp_replace(coalesce(pg_get_expr(p.polwithcheck, p.polrelid), ''), '\s+', ' ', 'g'), E'\n' ORDER BY p.polname)
FROM pg_policy p WHERE p.polrelid IN ` + tables)
	// Every table and column privilege on the two tables other than the owner's own.
	p["pro-acl"] = one(`SELECT string_agg(x, ';' ORDER BY x COLLATE "C") FROM (
  SELECT c.relname || ':' || a.grantee::regrole::text || ':' || a.privilege_type AS x
  FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
  WHERE c.oid IN ` + tables + ` AND a.grantee <> c.relowner
  UNION ALL
  SELECT c.relname || '.' || at.attname || ':' || a.grantee::regrole::text || ':' || a.privilege_type
  FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
  CROSS JOIN LATERAL aclexplode(at.attacl) a
  WHERE c.oid IN ` + tables + ` AND a.grantee <> c.relowner) s`)
	lbl := func(col string) string {
		return `CASE ` + col + ` WHEN 0 THEN 'public' WHEN $1::regrole THEN 'rt' WHEN $2::regrole THEN 'ow' WHEN $3::regrole THEN 'pr'
  WHEN $4::regrole THEN 'ar' WHEN $5::regrole THEN 'cw' WHEN $6::regrole THEN 'pw' ELSE 'other' END`
	}
	db := cfg.Database
	roles := []any{cfg.RuntimeRole(), cfg.OwnerRole(), db + "_policy_reader", db + "_auth_reader", db + "_claim_writer", db + "_provision_writer"}
	p["pro-functions"] = one(`SELECT string_agg(p.proname || ':' || `+lbl("p.proowner")+` || ':' || p.prosecdef::text || ':' || p.provolatile::text || ':' ||
  coalesce(array_to_string(p.proconfig, '|'), '-') || ':' || coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C")
  FROM (SELECT `+lbl("a.grantee")+` || '=' || a.privilege_type AS v FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'),
  ';' ORDER BY p.proname COLLATE "C")
FROM pg_proc p WHERE p.oid IN (to_regprocedure('app.account_avatars_guard()'), to_regprocedure('app.account_profiles_guard()'),
  to_regprocedure('audit.capture_account_profiles()'), to_regprocedure('audit.events_attribute()'),
  to_regprocedure('app.rls_profile_visible(uuid)'), to_regprocedure('app.rls_account_profile(uuid)'),
  to_regprocedure('app.rls_group_invitation_candidates(uuid, uuid, integer)'),
  to_regprocedure('app.auth_resolve_session(bytea, timestamptz, text)'))`, roles...)
	for part, sig := range proSignatures {
		p[part] = one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure($1)`, sig)
	}
	// The helper inventories (by owner) and their CREATE on schema app.
	p["pro-helpers"] = one(`SELECT string_agg(x, ';' ORDER BY x COLLATE "C") FROM (
  SELECT `+lbl("p.proowner")+` || '=' || string_agg(p.proname, ',' ORDER BY p.proname COLLATE "C") AS x
  FROM pg_proc p WHERE p.proowner IN ($3::regrole, $4::regrole, $5::regrole, $6::regrole) GROUP BY p.proowner
  UNION ALL
  SELECT 'create=' || has_schema_privilege($3::regrole, 'app', 'CREATE')::text || ',' || has_schema_privilege($4::regrole, 'app', 'CREATE')::text) s`,
		roles...)
	return p
}

// proViolations is the version-23 delta detector: the sorted parts that differ from want.
func proViolations(t *testing.T, q catalogQuerier, cfg Config, want map[string]string) []string {
	t.Helper()
	p := proParts(t, q, cfg)
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

func TestProfileCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r22 := runnerFor(t, cfg, subsetFS(t, 22))
	objRunUp(t, r22, "current_version\t22\n")
	s22 := snapshot(t, r22)
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	n22 := namesCatalog(t, cat)
	r := runnerFor(t, cfg, subsetFS(t, 23))
	objRunUp(t, r, "applied\t00023_account_profiles.sql\ncurrent_version\t23\n")
	n23 := namesCatalog(t, cat)
	clean := proParts(t, cat, cfg)
	o, rt, db := cfg.OwnerRole(), cfg.RuntimeRole(), cfg.Database
	pr, ar := db+"_policy_reader", db+"_auth_reader"

	t.Run("CT-PRO-1 the exact version-23 delta on the unchanged version-22 catalog", func(t *testing.T) {
		if got := namesDelta(n22, n23); !slices.Equal(got, proWantDelta) {
			t.Errorf("version-22 -> 23 delta:\n got %v\nwant %v", got, proWantDelta)
		}
		metadataThroughVersion23(t, cfg, r)
		for part, marker := range proSources {
			if clean[part] != md5hex(migBody(t, "00023_account_profiles.sql", marker)) {
				t.Errorf("%s differs from the reviewed migration body", part)
			}
		}
		if got, want := clean["pro-columns"], "account_avatars.id:uuid:t::f;account_avatars.user_id:uuid:t::f;"+
			"account_avatars.storage_key:text:f:s:t;account_avatars.state:text:t::t;account_avatars.byte_size:bigint:t::f;"+
			"account_avatars.content_type:text:t::f;account_avatars.sha256:bytea:f::f;"+
			"account_avatars.created_at:timestamp with time zone:t::t;account_avatars.ready_at:timestamp with time zone:f::f;"+
			"account_avatars.deleting_at:timestamp with time zone:f::f;account_avatars.deleted_at:timestamp with time zone:f::f;"+
			"account_avatars.upload_started_at:timestamp with time zone:f::f;"+
			"account_profiles.user_id:uuid:t::f;account_profiles.name:text:f::f;account_profiles.avatar_object_id:uuid:f::f;"+
			"account_profiles.avatar_version:integer:t::t;account_profiles.version:integer:t::t;"+
			"account_profiles.updated_at:timestamp with time zone:t::t"; got != want {
			t.Errorf("columns:\n got %s\nwant %s", got, want)
		}
		if got := clean["pro-flags"]; got != "account_avatars:true:true;account_profiles:true:true" {
			t.Errorf("row-level security flags: %s, want enabled and forced on both tables", got)
		}
		cons := clean["pro-constraints"]
		var names []string
		for _, c := range strings.Split(cons, ";") {
			n, _, _ := strings.Cut(c, ":")
			names = append(names, n)
		}
		if got := strings.Join(names, ","); got != "account_avatars_id_nonzero,account_avatars_lifecycle_check,account_avatars_pkey,"+
			"account_avatars_sha256_valid,account_avatars_size_valid,account_avatars_state_valid,account_avatars_type_valid,"+
			"account_avatars_user_fkey,account_profiles_avatar_fkey,account_profiles_avatar_shape,account_profiles_avatar_version_valid,"+
			"account_profiles_name_valid,account_profiles_pkey,account_profiles_user_fkey,account_profiles_version_valid" {
			t.Errorf("constraints: %s", got)
		}
		// Three RESTRICT foreign keys (dropping avatar metadata or a user never cascades), the avatar reference to the
		// avatar table only, and no Organization column or reference anywhere.
		if strings.Count(cons, "ON DELETE RESTRICT") != 3 || strings.Contains(cons, "CASCADE") || strings.Contains(cons, "organization") ||
			!strings.Contains(cons, "account_profiles_avatar_fkey:FOREIGN KEY (avatar_object_id) REFERENCES app.account_avatars(id)") ||
			!strings.Contains(cons, "BETWEEN 1 AND 524288") && !strings.Contains(cons, ">= 1) AND (byte_size <= 524288)") {
			t.Errorf("constraints: the foreign keys, the avatar reference or the size bound differ: %s", cons)
		}
		if got := clean["pro-indexes"]; got != "CREATE UNIQUE INDEX account_avatars_pkey ON app.account_avatars USING btree (id);"+
			"CREATE INDEX account_avatars_user_state_idx ON app.account_avatars USING btree (user_id, state);"+
			"CREATE UNIQUE INDEX account_profiles_pkey ON app.account_profiles USING btree (user_id)" {
			t.Errorf("indexes: %s", got)
		}
		trg := clean["pro-triggers"]
		if strings.Count(trg, "\n") != 2 ||
			!strings.Contains(trg, "account_avatars_guard:O:CREATE TRIGGER account_avatars_guard BEFORE UPDATE ON app.account_avatars FOR EACH ROW EXECUTE FUNCTION app.account_avatars_guard()") ||
			!strings.Contains(trg, "account_profiles_guard:O:CREATE TRIGGER account_profiles_guard BEFORE INSERT OR UPDATE ON app.account_profiles FOR EACH ROW EXECUTE FUNCTION app.account_profiles_guard()") ||
			!strings.Contains(trg, "audit_capture_account_profiles:O:CREATE TRIGGER audit_capture_account_profiles AFTER INSERT OR UPDATE OF name, avatar_object_id ON app.account_profiles FOR EACH ROW EXECUTE FUNCTION audit.capture_account_profiles()") {
			t.Errorf("triggers:\n%s", trg)
		}
		// The nine permissive policies: self-only reads and writes for the runtime and the owner (the profile read also
		// through the relationship predicate; the avatar insert only as pending), USING (true) for the two helpers only.
		selfRoles := []string{o, rt}
		slices.Sort(selfRoles)
		self := strings.Join(selfRoles, "+")
		pols := strings.Split(clean["pro-policies"], "\n")
		if len(pols) != 9 {
			t.Fatalf("policies: %q", clean["pro-policies"])
		}
		for _, l := range pols {
			f := strings.SplitN(l, ":", 6)
			if len(f) != 6 {
				t.Errorf("policy line: %q", l)
				continue
			}
			name, cmd, permissive, polRoles, qual, check := f[0], f[1], f[2], f[3], f[4], f[5]
			if permissive != "true" {
				t.Errorf("policy %s is restrictive", name)
			}
			switch name {
			case "account_avatars_policy_reader", "account_profiles_policy_reader", "account_profiles_auth_reader":
				want := pr
				if name == "account_profiles_auth_reader" {
					want = ar
				}
				if cmd != "r" || polRoles != want || qual != "true" || check != "" {
					t.Errorf("helper policy %s: %s", name, l)
				}
			case "account_avatars_select", "account_avatars_insert", "account_avatars_update", "account_profiles_insert",
				"account_profiles_update", "account_profiles_select":
				wantCmd := map[string]string{"account_avatars_select": "r", "account_avatars_insert": "a", "account_avatars_update": "w",
					"account_profiles_insert": "a", "account_profiles_update": "w", "account_profiles_select": "r"}[name]
				if cmd != wantCmd || polRoles != self || !strings.Contains(qual+check, "user_id = ( SELECT app.ctx_user()") {
					t.Errorf("self policy %s lacks its command, roles or the context user: %s", name, l)
				}
				if (name == "account_avatars_update" || name == "account_profiles_update") &&
					(!strings.Contains(qual, "app.ctx_user()") || !strings.Contains(check, "app.ctx_user()")) {
					t.Errorf("update policy %s is not self-only on both sides: %s", name, l)
				}
				if strings.Contains(qual+check, "rls_profile_visible") != (name == "account_profiles_select") {
					t.Errorf("policy %s and the relationship predicate: %s", name, l)
				}
				if name == "account_profiles_select" && !strings.Contains(qual, "app.rls_profile_visible(user_id)") {
					t.Errorf("profile read policy without the predicate on its own row: %s", l)
				}
				if name == "account_avatars_insert" && !strings.Contains(check, "'pending'") {
					t.Errorf("avatar insert policy without the pending state: %s", l)
				}
			default:
				t.Errorf("unexpected policy %s", l)
			}
		}
		var acl []string
		acl = append(acl, "account_avatars:"+rt+":SELECT", "account_profiles:"+rt+":SELECT")
		for _, c := range []string{"id", "user_id", "byte_size", "content_type"} {
			acl = append(acl, "account_avatars."+c+":"+rt+":INSERT")
		}
		for _, c := range []string{"state", "sha256", "upload_started_at"} {
			acl = append(acl, "account_avatars."+c+":"+rt+":UPDATE")
		}
		for _, c := range []string{"id", "user_id", "state", "content_type", "byte_size", "sha256"} {
			acl = append(acl, "account_avatars."+c+":"+pr+":SELECT")
		}
		for _, c := range []string{"user_id", "name", "avatar_object_id", "avatar_version", "version", "updated_at"} {
			acl = append(acl, "account_profiles."+c+":"+rt+":INSERT")
		}
		for _, c := range []string{"name", "avatar_object_id", "avatar_version", "version", "updated_at"} {
			acl = append(acl, "account_profiles."+c+":"+rt+":UPDATE")
		}
		for _, c := range []string{"user_id", "name", "avatar_object_id", "avatar_version"} {
			acl = append(acl, "account_profiles."+c+":"+pr+":SELECT")
		}
		for _, c := range []string{"user_id", "name"} {
			acl = append(acl, "account_profiles."+c+":"+ar+":SELECT")
		}
		slices.Sort(acl)
		if got := clean["pro-acl"]; got != strings.Join(acl, ";") {
			t.Errorf("privileges:\n got %s\nwant %s", got, strings.Join(acl, ";"))
		}
		// No DELETE, TRUNCATE, REFERENCES or TRIGGER for the runtime, and nothing at all for the retention role, the
		// migrator, the claim and provision writers or PUBLIC (tombstones and profile rows are kept; no purge grant).
		one := func(sql string, args ...any) string {
			t.Helper()
			var s string
			if err := cat.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
				t.Fatalf("catalog: %v", err)
			}
			return s
		}
		for _, tbl := range []string{"app.account_avatars", "app.account_profiles"} {
			if got := one(`SELECT has_table_privilege($1, $2, 'DELETE, TRUNCATE, REFERENCES, TRIGGER')::text`, rt, tbl); got != "false" {
				t.Errorf("runtime holds DELETE, TRUNCATE, REFERENCES or TRIGGER on %s", tbl)
			}
			for _, role := range []string{cfg.RetentionRole(), cfg.MigratorRole(), db + "_claim_writer", db + "_provision_writer", "public"} {
				if got := one(`SELECT (has_table_privilege($1, $2, 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER')
  OR has_any_column_privilege($1, $2, 'SELECT, INSERT, UPDATE, REFERENCES'))::text`, role, tbl); got != "false" {
					t.Errorf("%s holds a privilege on %s", role, tbl)
				}
			}
		}
		const sp = "search_path=pg_catalog, pg_temp"
		if got, want := clean["pro-functions"], "account_avatars_guard:ow:false:v:"+sp+":-;account_profiles_guard:ow:false:v:"+sp+":-;"+
			"auth_resolve_session:ar:true:s:"+sp+":rt=EXECUTE;capture_account_profiles:ow:false:v:"+sp+":-;"+
			"events_attribute:ow:false:v:"+sp+":-;rls_account_profile:pr:true:s:"+sp+":ow=EXECUTE+rt=EXECUTE;"+
			"rls_group_invitation_candidates:pr:true:s:"+sp+":ow=EXECUTE+rt=EXECUTE;"+
			"rls_profile_visible:pr:true:s:"+sp+":ow=EXECUTE+rt=EXECUTE"; got != want {
			t.Errorf("functions:\n got %s\nwant %s", got, want)
		}
		if got, want := clean["pro-helpers"], "ar=auth_logout_ref,auth_principal_subject,auth_resolve_enrollment,auth_resolve_session,"+
			"auth_resolve_signup,hostname_state,is_organization_member,organization_id_by_slug,public_branding,signup_slug_taken;"+
			"create=false,false;"+
			"cw=enrollment_admit,enrollment_lock_organization,enrollment_request,invitation_claim_group,invitation_claim_organization,"+
			"rls_invitation_pending,rls_user_enabled_lock;"+
			"pr=rls_account_profile,rls_dependent_group_bindings,rls_enrollment_mode,rls_group_invitation_candidates,rls_group_role,"+
			"rls_org_role,rls_own_organizations,rls_owner_count,rls_profile_eligible,rls_profile_visible,rls_sole_owner_scopes;"+
			"pw=signup_challenge_complete,signup_challenge_issue,signup_provision_organization"; got != want {
			t.Errorf("helper inventories:\n got %s\nwant %s", got, want)
		}
		// E1 keeps its identity and result columns; the invitation candidates keep theirs (no avatar column).
		if got := one(`SELECT pg_get_function_result('app.auth_resolve_session(bytea, timestamptz, text)'::regprocedure)`); got !=
			"TABLE(session_id uuid, user_id uuid, display_name text, mfa_assurance text, auth_time timestamp with time zone, "+
				"idle_expires_at timestamp with time zone, absolute_expires_at timestamp with time zone, renewed_at timestamp with time zone, "+
				"csrf_token bytea, provider_session_id text, origin text, organization_id uuid)" {
			t.Errorf("E1 result columns: %s", got)
		}
		if got := one(`SELECT pg_get_function_result('app.rls_group_invitation_candidates(uuid, uuid, integer)'::regprocedure)`); got !=
			"TABLE(user_id uuid, display_name text)" {
			t.Errorf("invitation candidates result columns: %s", got)
		}
		if got := one(`SELECT pg_get_function_result('app.rls_account_profile(uuid)'::regprocedure)`); got !=
			"TABLE(user_id uuid, name text, avatar_version integer, avatar_object_id uuid, avatar_type text, avatar_size bigint, avatar_sha256 bytea)" {
			t.Errorf("projection result columns: %s", got)
		}
		// The audit catalog names the profile action, reasons, operations and resource type.
		for con, vals := range map[string][]string{
			"events_action_valid":        {"'profile.updated'"},
			"events_reason_valid":        {"'name_changed'", "'avatar_set'", "'avatar_removed'"},
			"events_operation_valid":     {"'profile.get'", "'profile.update'", "'profile.avatar.get'", "'profile.avatar.set'", "'profile.avatar.remove'"},
			"events_resource_type_valid": {"'account_profile'"},
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
  AND c.oid NOT IN (SELECT i.indexrelid FROM pg_index i WHERE i.indrelid = 'app.goose_db_version'::regclass)`); got != "94" {
			t.Errorf("schema app holds %s relations, want 94", got)
		}
	})

	t.Run("CT-PRO-NEG rolled-back missing-protection negatives", func(t *testing.T) {
		// The probe world (superuser fixture, inside the probe's rolled-back transaction): proA with a ready current avatar
		// and a pending one, proB with a ready avatar; the context is proA's IDENTITY shape, so the guards' actor checks and
		// the capture's attribution see the owning user.
		seed := `INSERT INTO app.users (id, display_name) VALUES ('` + proA + `', 'Catalog A'), ('` + proB + `', 'Catalog B');
INSERT INTO app.account_avatars (id, user_id, state, byte_size, content_type, sha256, ready_at) VALUES
  ('` + proAvatarReady + `', '` + proA + `', 'ready', 10, 'image/png', sha256('a'::bytea), now()),
  ('` + proAvatarB + `', '` + proB + `', 'ready', 10, 'image/png', sha256('b'::bytea), now());
INSERT INTO app.account_avatars (id, user_id, byte_size, content_type) VALUES ('` + proAvatarPending + `', '` + proA + `', 10, 'image/png');
SELECT set_config('app.user_id', '` + proA + `', true), set_config('app.organization_id', '', true), set_config('app.group_id', '', true);
INSERT INTO app.account_profiles (user_id, avatar_object_id, avatar_version) VALUES ('` + proA + `', '` + proAvatarReady + `', 1)`
		ref := func(avatar string) string {
			return `UPDATE app.account_profiles SET avatar_object_id = '` + avatar + `', avatar_version = avatar_version + 1, version = version + 1
WHERE user_id = '` + proA + `'`
		}
		foreignRef, pendingRef := ref(proAvatarB), ref(proAvatarPending)
		withdrawCurrent := `UPDATE app.account_avatars SET state = 'deleting', deleting_at = now() WHERE id = '` + proAvatarReady + `'`
		contextFree := `DO $$ BEGIN IF NOT app.rls_profile_visible('` + proB + `'::uuid) THEN
  RAISE EXCEPTION 'not visible' USING ERRCODE = 'insufficient_privilege'; END IF; END $$`
		probe := func(id, mutation, stmt string) error {
			t.Helper()
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			for _, s := range []string{mutation, seed} {
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
		// The intact protections refuse every probe: another user's or a non-ready avatar reference (the profile guard),
		// withdrawing the current avatar (I-CUR), and the predicate without a valid context.
		for _, c := range []struct{ what, stmt, state string }{
			{"another user's avatar", foreignRef, "23514"}, {"a pending avatar", pendingRef, "23514"},
			{"the current avatar withdrawn", withdrawCurrent, "23000"}, {"the predicate without a context", contextFree, "42501"},
		} {
			if err := probe("intact", "", c.stmt); sqlstateOf(err) != c.state {
				t.Fatalf("intact protections: %s %v, want %s", c.what, err, c.state)
			}
		}
		for _, n := range []struct {
			id, sql string
			want    []string
			stmts   []string // statements the missing protection now admits
		}{
			{"N1 row-level security not forced", `ALTER TABLE app.account_profiles NO FORCE ROW LEVEL SECURITY`, []string{"pro-flags"}, nil},
			{"N2 runtime DELETE", `GRANT DELETE ON app.account_profiles TO ` + quote(rt), []string{"pro-acl"}, nil},
			{"N3 projection executable by PUBLIC", `GRANT EXECUTE ON FUNCTION app.rls_account_profile(uuid) TO PUBLIC`, []string{"pro-functions"}, nil},
			{"N4 predicate without the validity guard", `CREATE OR REPLACE FUNCTION app.rls_profile_visible(p_subject uuid) RETURNS boolean
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $f$ BEGIN RETURN p_subject IS NOT NULL; END $f$`,
				[]string{"pro-visible-src"}, []string{contextFree}},
			{"N5 invitation candidates without the gate", `CREATE OR REPLACE FUNCTION app.rls_group_invitation_candidates(p_group uuid, p_after uuid, p_limit integer)
  RETURNS TABLE (user_id uuid, display_name text)
  LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $f$
#variable_conflict use_column
BEGIN
  RETURN QUERY SELECT us.id, coalesce(ap.name, us.display_name) FROM app.users us LEFT JOIN app.account_profiles ap ON ap.user_id = us.id
  ORDER BY us.id LIMIT least(p_limit, 101);
END $f$`, []string{"pro-picker-src"}, nil},
			{"N6 profile guard admitting another user's or a non-ready avatar", `CREATE OR REPLACE FUNCTION app.account_profiles_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $f$ BEGIN RETURN NEW; END $f$`,
				[]string{"pro-profiles-guard-src"}, []string{foreignRef, pendingRef}},
			{"N7 avatar guard without I-CUR", `CREATE OR REPLACE FUNCTION app.account_avatars_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $f$ BEGIN RETURN NEW; END $f$`,
				[]string{"pro-avatars-guard-src"}, []string{withdrawCurrent}},
			{"N8 profile read policy without the self and relationship rule", `ALTER POLICY account_profiles_select ON app.account_profiles USING (true)`,
				[]string{"pro-policies"}, nil},
			{"N9 retention grant on avatar metadata", `GRANT SELECT ON app.account_avatars TO ` + quote(cfg.RetentionRole()), []string{"pro-acl"}, nil},
			{"N10 auth reader reading the avatar reference", `GRANT SELECT (avatar_object_id) ON app.account_profiles TO ` + quote(ar),
				[]string{"pro-acl"}, nil},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			v := proViolations(t, tx, cfg, clean)
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
		if v := proViolations(t, cat, cfg, clean); len(v) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", v)
		}
	})

	t.Run("CT-PRO-UPDOWN exact version-22 catalog when empty, up again, down refused with profile state or history", func(t *testing.T) {
		s23 := snapshot(t, r)
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
		if err != nil || out != "rolled_back\t00023_account_profiles.sql\ncurrent_version\t22\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		metadataThroughVersion22(t, cfg, r)
		if got := snapshot(t, r); got != s22 {
			t.Fatalf("catalog after down differs from version 22:\n%s\n---\n%s", got, s22)
		}
		if got := namesDelta(n22, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down differs from version 22: %v", got)
		}
		objRunUp(t, r, "applied\t00023_account_profiles.sql\ncurrent_version\t23\n")
		if got := snapshot(t, r); got != s23 {
			t.Fatalf("catalog after down/up differs:\n%s\n---\n%s", got, s23)
		}
		if got := namesDelta(n23, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down/up differs from version 23: %v", got)
		}
		if v := proViolations(t, cat, cfg, clean); len(v) != 0 {
			t.Fatalf("version-23 detector after up again: %v", v)
		}
		refused := func(what string) {
			t.Helper()
			if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00023:") {
				t.Fatalf("down with %s: %v %q", what, err, out)
			}
			if got := snapshot(t, r); got != s23 {
				t.Fatalf("the refused down (%s) changed the catalog", what)
			}
		}
		// A profile row without a name (no history is written for it) refuses the down.
		exec(`INSERT INTO app.users (id, display_name) VALUES ('` + proA + `', 'Catalog Down')`)
		exec(`INSERT INTO app.account_profiles (user_id) VALUES ('` + proA + `')`)
		refused("a profile row")
		exec(`DELETE FROM app.account_profiles WHERE user_id = '` + proA + `'`)
		// So does an avatar row alone (dropping it would orphan provider bytes).
		exec(`INSERT INTO app.account_avatars (id, user_id, byte_size, content_type) VALUES ('` + proAvatarPending + `', '` + proA + `', 10, 'image/png')`)
		refused("an avatar row")
		exec(`DELETE FROM app.account_avatars WHERE id = '` + proAvatarPending + `'`)
		// And profile history is never discarded: a name edit's profile.updated row still refuses with the row gone.
		func() {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			for _, s := range []string{
				`SELECT set_config('app.user_id', '` + proA + `', true), set_config('app.organization_id', '', true), set_config('app.group_id', '', true)`,
				`INSERT INTO app.account_profiles (user_id, name) VALUES ('` + proA + `', 'Catalog History')`,
				`DELETE FROM app.account_profiles WHERE user_id = '` + proA + `'`,
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
		if err := setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'profile.updated' AND reason = 'name_changed'
  AND subject_user_id = $1::uuid AND organization_id IS NULL`, proA).Scan(&hist); err != nil || hist != 1 {
			t.Fatalf("profile history rows: %d (%v), want 1", hist, err)
		}
		refused("profile history")
		reset(t, cfg)
	})
}
