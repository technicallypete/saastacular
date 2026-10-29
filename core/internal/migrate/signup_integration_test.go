//go:build integration

// G062 B-SIGNUP-CATALOG (Plan r5 §3, §5.2 SUP-4/SUP-12 with the r5 ACCEPT pending-row clarification): TestSignupCatalog.
//   - CT-SUP-1_precondition (version 17): 00018 refuses while <db>_provision_writer is missing or has INHERIT, with
//     nothing applied; the role is provisioning-owned and never created by the migration.
//   - CT-SUP-1_tables the signup table (columns, constraints, foreign keys, index, RLS/FORCE), the login_transactions
//     delta and the version-18 relation/function inventory.
//   - CT-SUP-1_policies the exact policy inventory on app.signup_sessions and the provision writer's policies elsewhere
//     (each TO the provision writer only), with their load-bearing predicates: the provision writer's SELECT/UPDATE admit
//     every row of an ACTIVE user (pending and provisioned alike), its UPDATE keeps the proof-reset WITH CHECK.
//   - CT-SUP-1_privileges the table and column ACLs (runtime INSERT 8 / UPDATE 4 columns, provision writer UPDATE of exactly
//     14 columns) and the provision writer's non-grants: no EXECUTE on hostname_state / is_organization_member, no
//     login_transactions privilege, no audit SELECT, no DELETE/TRUNCATE anywhere, no CREATE on app; EXECUTE only on the four
//     context parsers.
//   - CT-SUP-1_definers the four SECURITY DEFINER bodies (owner, volatility, search_path, EXECUTE holders) and the owned-
//     function inventories (provision writer 3, auth reader 8, claim writer 7).
//   - CT-SUP-1_negatives (every case rolled back): the bodies refuse the TENANT, LOGIN and empty shapes (42501 with their
//     named constraint); the continuation bodies refuse outside the LOGIN shape; completion with another Organization
//     matches nothing while the right one matches once; the runtime and the owner cannot INSERT an Organization or an
//     org_owner binding (42501); runtime UPDATE of a provision-writer column and a TENANT-shape signup INSERT are 42501.
//   - CT-SUP-1_provision_writer (rolled back, owner SET LOCAL ROLE to the provision writer, IDENTITY context): the pending
//     and the provisioned row of the active user are both visible; a disabled user's row is not; a proof set without a
//     completion violates the WITH CHECK (42501); an ungranted column, DELETE, login_transactions, hostname_state and a
//     depth-0 audit INSERT are 42501.
//   - CT-SUP-1_return_path (rolled back): /signup/details is admitted, look-alikes and a handoff without its expected user
//     are 23514.
//   - CT-SUP-1_updown the down refused while a signup row exists (catalog unchanged), then a clean down to the exact
//     version-17 catalog and up again to the exact version-18 catalog.
package migrate

import (
	"context"
	"crypto/sha256"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	sigUserA  = "0f62c000-0000-4000-8000-0000000000a1" // active; one pending and one provisioned row
	sigUserD  = "0f62c000-0000-4000-8000-0000000000d1" // disabled; one pending row
	sigOrg    = "0f62c000-0000-4000-8000-000000000001" // the provisioned row's Organization
	sigOrgX   = "0f62c000-0000-4000-8000-000000000002" // another Organization (completion mismatch)
	sigRowA0  = "0f62c000-0000-4000-8000-000000000a00" // A's pending row
	sigRowA1  = "0f62c000-0000-4000-8000-000000000a01" // A's provisioned row, ready for completion
	sigRowD0  = "0f62c000-0000-4000-8000-000000000d00" // D's pending row
	sigRoot   = "https://localhost:543"
	sigTarget = "https://sigcat.localhost:543"
)

// sigDigest is a fixture SHA-256 digest (32 bytes).
func sigDigest(s string) []byte { b := sha256.Sum256([]byte("g062-catalog/" + s)); return b[:] }

func sigPgErr(err error) (code, constraint string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code, pgErr.ConstraintName
	}
	return "", ""
}

func TestSignupCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r17 := runnerFor(t, cfg, subsetFS(t, 17))
	objRunUp(t, r17, "current_version\t17\n")
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	rt := rlsRuntimeConn(t, cfg)
	own := rlsOwnerConn(t, cfg)
	o, rtr, ar, cw, pr := cfg.OwnerRole(), cfg.RuntimeRole(), cfg.Database+"_auth_reader", cfg.Database+"_claim_writer",
		cfg.Database+"_policy_reader"
	pw := cfg.Database + "_provision_writer"
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := setup.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("fixture: %v", err)
		}
	}
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
	// lbl maps a role oid column onto labels ($1..$6: runtime, owner, auth reader, provision writer, claim writer, policy
	// reader); every query using it passes roles.
	lbl := func(col string) string {
		return `CASE ` + col + ` WHEN 0 THEN 'public' WHEN $1::regrole THEN 'rt' WHEN $2::regrole THEN 'ow' WHEN $3::regrole THEN 'ar'
  WHEN $4::regrole THEN 'pw' WHEN $5::regrole THEN 'cw' WHEN $6::regrole THEN 'pr' ELSE 'other' END`
	}
	roles := []any{rtr, o, ar, pw, cw, pr}
	// state is the catalog an 00018 up or down changes (the TestEnrollmentCatalog state): relations/ACLs/history, every
	// app and audit function, policy, column ACL, RLS flag, schema ACL, constraint and non-internal trigger.
	state := func(r *Runner) string {
		t.Helper()
		return snapshot(t, r) + "\n" + one(`SELECT concat_ws(E'\n--\n',
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s|%s|%s|%s', p.oid::regprocedure, p.proowner::regrole,
     p.prosecdef, coalesce(p.proacl::text, ''), md5(p.prosrc)) AS x FROM pg_proc p
     WHERE p.pronamespace IN ('app'::regnamespace, 'audit'::regnamespace)) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s|%s|%s|%s|%s', p.polrelid::regclass, p.polname,
     p.polcmd, p.polroles::regrole[]::text, coalesce(pg_get_expr(p.polqual, p.polrelid), ''),
     coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) AS x FROM pg_policy p) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s.%s|%s', c.oid::regclass, a.attname, a.attacl::text) AS x
     FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
     WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND cardinality(a.attacl) > 0) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s|%s', c.oid::regclass, c.relrowsecurity,
     c.relforcerowsecurity) AS x FROM pg_class c
     WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND c.relkind = 'r') s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s', n.nspname, coalesce(n.nspacl::text, '')) AS x
     FROM pg_namespace n WHERE n.nspname IN ('app', 'audit')) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s|%s', c.conrelid::regclass, c.conname,
     md5(pg_get_constraintdef(c.oid))) AS x FROM pg_constraint c
     WHERE c.connamespace IN ('app'::regnamespace, 'audit'::regnamespace)) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s', t.tgrelid::regclass, t.tgname) AS x
     FROM pg_trigger t WHERE NOT t.tgisinternal) s))`)
	}

	// SUP-12: the migration requires the provisioning-owned fourth helper with safe attributes; a refusal applies nothing.
	t.Run("CT-SUP-1_precondition", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 18))
		up := func() error { return r.Up(ctx, nopWriter{}) }
		exec(`ALTER ROLE ` + quote(pw) + ` RENAME TO ` + quote(pw+"_away"))
		err := up()
		exec(`ALTER ROLE ` + quote(pw+"_away") + ` RENAME TO ` + quote(pw))
		if err == nil || !strings.Contains(err.Error(), "roles for database "+cfg.Database+" are missing") {
			t.Errorf("up without the provision writer: %v, want the missing-roles refusal", err)
		}
		exec(`ALTER ROLE ` + quote(pw) + ` INHERIT`)
		err = up()
		exec(`ALTER ROLE ` + quote(pw) + ` NOINHERIT`)
		if err == nil || !strings.Contains(err.Error(), "must be NOLOGIN NOINHERIT") {
			t.Errorf("up with an INHERIT provision writer: %v, want the helper-attribute refusal", err)
		}
		if got := one(`SELECT max(version_id)::text FROM app.goose_db_version`); got != "17" {
			t.Fatalf("version after the refused ups: %s, want 17", got)
		}
		if got := one(`SELECT coalesce(to_regclass('app.signup_sessions')::text, 'absent')`); got != "absent" {
			t.Errorf("signup table after the refused ups: %s", got)
		}
	})
	s17 := state(r17)
	r := runnerFor(t, cfg, subsetFS(t, 18))
	objRunUp(t, r, "applied\t00018_platform_signup.sql\ncurrent_version\t18\n")
	s18 := state(r)

	// Fixtures (committed; removed before the clean down): A active with a pending row and a provisioned row whose
	// continuation is ready for completion (R consumed, C and N stored, N consumed, P stored); D, disabled, with a pending row;
	// the provisioned row's Organization and another one.
	exec(`INSERT INTO app.users (id, display_name, status, disabled_at) VALUES ($1, 'Sig A', 'active', NULL), ($2, 'Sig D', 'disabled', now())`,
		sigUserA, sigUserD)
	exec(`INSERT INTO app.organizations (id, name, slug) VALUES ($1, 'Sig Catalog', 'sigcat'), ($2, 'Sig Other', 'sigother')`, sigOrg, sigOrgX)
	exec(`INSERT INTO app.signup_sessions (id, user_id, origin, credential_hash, csrf_token, provider_session_id, expires_at)
VALUES ($1, $2, $3, $4, $5, 'sid-a0', now() + interval '25 minutes'), ($6, $7, $3, $8, $9, 'sid-d0', now() + interval '25 minutes')`,
		sigRowA0, sigUserA, sigRoot, sigDigest("cred-a0"), sigDigest("csrf-a0"), sigRowD0, sigUserD, sigDigest("cred-d0"), sigDigest("csrf-d0"))
	exec(`INSERT INTO app.signup_sessions (id, user_id, origin, credential_hash, csrf_token, provider_session_id, expires_at, intent_hash,
  organization_id, provisioned_at, target_origin, continuation_generation, continuation_expires_at, ref_hash, ref_consumed_at,
  challenge_hash, nonce_hash, nonce_consumed_at, proof_hash)
VALUES ($1, $2, $3, $4, $5, 'sid-a1', now() + interval '25 minutes', $6, $7, now(), $8, 1, now() + interval '20 minutes', $9, now(),
  $10, $11, now(), $12)`, sigRowA1, sigUserA, sigRoot, sigDigest("cred-a1"), sigDigest("csrf-a1"), sigDigest("intent-a1"), sigOrg,
		sigTarget, sigDigest("ref-a1"), sigDigest("challenge-a1"), sigDigest("nonce-a1"), sigDigest("proof-a1"))

	t.Run("CT-SUP-1_tables", func(t *testing.T) {
		metadataThroughVersion18(t, cfg, r)
		if got := one(`SELECT string_agg(a.attname || ':' || format_type(a.atttypid, a.atttypmod) || ':' || a.attnotnull::text, ',' ORDER BY a.attnum)
FROM pg_attribute a WHERE a.attrelid = 'app.signup_sessions'::regclass AND a.attnum > 0 AND NOT a.attisdropped`); got !=
			"id:uuid:true,user_id:uuid:true,origin:text:true,credential_hash:bytea:true,csrf_token:bytea:true,provider_session_id:text:true,"+
				"created_at:timestamp with time zone:true,expires_at:timestamp with time zone:true,revoked_at:timestamp with time zone:false,"+
				"revoke_reason:text:false,intent_hash:bytea:false,organization_id:uuid:false,provisioned_at:timestamp with time zone:false,"+
				"target_origin:text:false,continuation_generation:integer:true,continuation_expires_at:timestamp with time zone:false,"+
				"ref_hash:bytea:false,ref_consumed_at:timestamp with time zone:false,challenge_hash:bytea:false,nonce_hash:bytea:false,"+
				"nonce_consumed_at:timestamp with time zone:false,proof_hash:bytea:false,proof_consumed_at:timestamp with time zone:false,"+
				"completed_at:timestamp with time zone:false" {
			t.Errorf("signup_sessions columns (24): %s", got)
		}
		if got := one(`SELECT string_agg(conname || ':' || contype::text, ',' ORDER BY conname COLLATE "C") FROM pg_constraint
WHERE conrelid = 'app.signup_sessions'::regclass`); got != "signup_sessions_challenge_hash_key:u,signup_sessions_challenge_shape:c,"+
			"signup_sessions_completion_shape:c,signup_sessions_continuation_requires_provisioned:c,signup_sessions_credential_hash_key:u,"+
			"signup_sessions_digest_lengths:c,signup_sessions_generation_valid:c,signup_sessions_lifetime:c,"+
			"signup_sessions_nonce_hash_key:u,signup_sessions_nonce_shape:c,signup_sessions_organization_fkey:f,"+
			"signup_sessions_origin_length:c,signup_sessions_pkey:p,signup_sessions_proof_hash_key:u,signup_sessions_proof_shape:c,"+
			"signup_sessions_provider_session_id_length:c,signup_sessions_provisioned_shape:c,signup_sessions_ref_hash_key:u,"+
			"signup_sessions_revocation_shape:c,signup_sessions_target_origin_length:c,signup_sessions_user_fkey:f" {
			t.Errorf("signup_sessions constraints: %s", got)
		}
		if got := one(`SELECT string_agg(conname || ':' || confdeltype::text || ':' || confrelid::regclass::text, ',' ORDER BY conname COLLATE "C")
FROM pg_constraint WHERE contype = 'f' AND (conrelid = 'app.signup_sessions'::regclass
  OR (conrelid = 'app.login_transactions'::regclass AND conname = 'login_transactions_expected_user_fkey'))`); got !=
			"login_transactions_expected_user_fkey:c:app.users,signup_sessions_organization_fkey:r:app.organizations,"+
				"signup_sessions_user_fkey:r:app.users" {
			t.Errorf("signup foreign keys: %s", got)
		}
		if got := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = 'signup_sessions_lifetime'`); !strings.Contains(got, "00:30:00") {
			t.Errorf("signup lifetime: %s, want 30 minutes", got)
		}
		if got := normalize(one(`SELECT pg_get_indexdef('app.signup_sessions_user_idx'::regclass)`)); got !=
			"CREATE INDEX signup_sessions_user_idx ON app.signup_sessions USING btree (user_id, provider_session_id) WHERE (revoked_at IS NULL)" {
			t.Errorf("signup user index: %s", got)
		}
		if got := one(`SELECT relrowsecurity::text || ':' || relforcerowsecurity::text FROM pg_class WHERE oid = 'app.signup_sessions'::regclass`); got != "true:true" {
			t.Errorf("signup_sessions RLS/FORCE: %s", got)
		}
		if got := one(`SELECT string_agg(a.attname || ':' || format_type(a.atttypid, a.atttypmod) || ':' || a.attnotnull::text || ':' ||
  coalesce(pg_get_expr(d.adbin, d.adrelid), '-'), ',' ORDER BY a.attname)
FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = 'app.login_transactions'::regclass AND a.attname IN ('signup_intent', 'signup_handoff', 'expected_user_id')
  AND NOT a.attisdropped`); got != "expected_user_id:uuid:false:-,signup_handoff:boolean:true:false,signup_intent:boolean:true:false" {
			t.Errorf("login_transactions signup columns: %s", got)
		}
		def := func(name string) string {
			return normalize(one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'app.login_transactions'::regclass AND conname = $1`, name))
		}
		for name, frags := range map[string][]string{
			"login_transactions_signup_intent_shape":  {"NOT signup_intent", "organization_id IS NULL", "NOT enrollment_intent"},
			"login_transactions_signup_handoff_shape": {"NOT signup_handoff", "organization_id IS NOT NULL", "expected_user_id IS NOT NULL", "NOT signup_intent"},
			"login_transactions_expected_user_shape":  {"signup_handoff OR (expected_user_id IS NULL)"},
			"login_transactions_return_path_check":    {"'/'::text", "'/account'::text", "'/join/status'::text", "'/signup/details'::text"},
		} {
			got := def(name)
			for _, f := range frags {
				if !strings.Contains(got, f) {
					t.Errorf("%s = %s, missing %q", name, got, f)
				}
			}
		}
	})

	t.Run("CT-SUP-1_policies", func(t *testing.T) {
		if got := one(`SELECT string_agg(p.polname || ':' || p.polcmd::text || ':' || (SELECT string_agg(l, '+' ORDER BY l COLLATE "C")
  FROM (SELECT `+lbl("x")+` AS l FROM unnest(p.polroles) x) s), ';' ORDER BY p.polname COLLATE "C")
FROM pg_policy p WHERE p.polrelid = 'app.signup_sessions'::regclass`, roles...); got !=
			"signup_sessions_auth_reader:r:ar;signup_sessions_expire:d:ow+rt;signup_sessions_insert:a:ow+rt;"+
				"signup_sessions_prove:w:ow+rt;signup_sessions_provision_select:r:pw;signup_sessions_provision_update:w:pw;"+
				"signup_sessions_revoke:w:ow+rt;signup_sessions_select:r:ow+rt" {
			t.Errorf("signup_sessions policies: %s", got)
		}
		// Every policy naming the provision writer names it alone.
		if got := one(`SELECT string_agg(p.polrelid::regclass::text || ':' || p.polname || ':' || p.polcmd::text || ':' || cardinality(p.polroles)::text,
  ';' ORDER BY p.polname COLLATE "C") FROM pg_policy p WHERE $1::regrole = ANY (p.polroles)`, pw); got !=
			"audit.events:events_insert_provision:a:1;app.organizations:organizations_provision_insert:a:1;"+
				"app.role_bindings:role_bindings_insert_owner:a:1;app.signup_sessions:signup_sessions_provision_select:r:1;"+
				"app.signup_sessions:signup_sessions_provision_update:w:1;app.users:users_provision_lock:w:1;app.users:users_provision_select:r:1" {
			t.Errorf("provision-writer policies: %s", got)
		}
		expr := func(name string) string {
			return normalize(one(`SELECT coalesce(pg_get_expr(polqual, polrelid), '') || ' WITH CHECK ' || coalesce(pg_get_expr(polwithcheck, polrelid), '')
FROM pg_policy WHERE polname = $1`, name))
		}
		for name, c := range map[string]struct{ has, not []string }{
			// r5 clarification: an active user's rows, pending (first create) and provisioned (retry, continuation) alike.
			"signup_sessions_provision_select": {has: []string{"u.id = signup_sessions.user_id", "u.status = 'active'::text"},
				not: []string{"provisioned_at", "ctx_user"}},
			"signup_sessions_provision_update": {has: []string{"u.id = signup_sessions.user_id", "u.status = 'active'::text",
				"WITH CHECK ((proof_hash IS NULL) OR (completed_at IS NOT NULL))"}, not: []string{"provisioned_at"}},
			"users_provision_lock":           {has: []string{"id = app.ctx_user()", "WITH CHECK false"}},
			"organizations_provision_insert": {has: []string{"pg_trigger_depth() = 0", "app.ctx_user() IS NOT NULL", "app.ctx_org() IS NULL", "slug IS NOT NULL"}},
			"role_bindings_insert_owner": {has: []string{"subject_user_id = app.ctx_user()", "app.ctx_org() IS NULL", "scope_id = organization_id",
				"group_id IS NULL", "'org_owner'::text", "'membership'::text"}},
			"events_insert_provision": {has: []string{"pg_trigger_depth() >= 1", "actor_user_id = app.ctx_user()", "organization_id IS NOT NULL",
				"'organization.created'::text", "'membership.granted'::text"}},
			"signup_sessions_insert": {has: []string{"user_id = ( SELECT app.ctx_user() AS ctx_user)", "( SELECT app.ctx_org() AS ctx_org) IS NULL",
				"revoked_at IS NULL", "provisioned_at IS NULL"}},
			"signup_sessions_prove":  {has: []string{"user_id = ( SELECT app.ctx_user() AS ctx_user)", "( SELECT app.ctx_org() AS ctx_org) IS NULL"}},
			"signup_sessions_expire": {has: []string{"expires_at <= now()", "( SELECT app.ctx_user() AS ctx_user) IS NULL", "( SELECT app.ctx_login_binding() AS ctx_login_binding) IS NULL"}},
		} {
			got := expr(name)
			for _, f := range c.has {
				if !strings.Contains(got, f) {
					t.Errorf("%s = %s, missing %q", name, got, f)
				}
			}
			for _, f := range c.not {
				if strings.Contains(got, f) {
					t.Errorf("%s = %s, must not reference %q", name, got, f)
				}
			}
		}
		// No runtime or owner INSERT path on app.organizations (or an org_owner binding) is added by 00018: the only
		// runtime/owner INSERT policy there stays the 00014 trigger-bound group_owner bootstrap.
		if got := one(`SELECT coalesce(string_agg(p.polname, ',' ORDER BY p.polname COLLATE "C"), '') FROM pg_policy p
WHERE p.polcmd IN ('a', '*') AND p.polrelid IN ('app.organizations'::regclass, 'app.role_bindings'::regclass)
  AND ($1::regrole = ANY (p.polroles) OR $2::regrole = ANY (p.polroles))`, rtr, o); got != "role_bindings_insert_bootstrap" {
			t.Errorf("runtime/owner INSERT policies on organizations/role_bindings: %s", got)
		}
		if got := normalize(one(`SELECT pg_get_expr(polwithcheck, polrelid) FROM pg_policy WHERE polname = 'role_bindings_insert_bootstrap'`)); !strings.Contains(got,
			"pg_trigger_depth() > 0") || !strings.Contains(got, "'group_owner'::text") {
			t.Errorf("role_bindings_insert_bootstrap = %s, want the trigger-bound group_owner bootstrap only", got)
		}
	})

	t.Run("CT-SUP-1_privileges", func(t *testing.T) {
		if got := one(`SELECT string_agg(x, ',' ORDER BY x COLLATE "C") FROM (SELECT `+lbl("a.grantee")+` || ':' || a.privilege_type AS x
FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a WHERE c.oid = 'app.signup_sessions'::regclass AND a.grantee <> c.relowner) s`,
			roles...); got != "ar:SELECT,pw:SELECT,rt:DELETE,rt:SELECT" {
			t.Errorf("signup_sessions table ACL: %s", got)
		}
		var want []string
		for _, c := range []string{"id", "user_id", "origin", "credential_hash", "csrf_token", "provider_session_id", "created_at", "expires_at"} {
			want = append(want, "rt:INSERT:"+c)
		}
		for _, c := range []string{"revoked_at", "revoke_reason", "nonce_consumed_at", "proof_hash"} {
			want = append(want, "rt:UPDATE:"+c)
		}
		for _, c := range []string{"intent_hash", "organization_id", "provisioned_at", "target_origin", "continuation_generation",
			"continuation_expires_at", "ref_hash", "ref_consumed_at", "challenge_hash", "nonce_hash", "nonce_consumed_at", "proof_hash",
			"proof_consumed_at", "completed_at"} {
			want = append(want, "pw:UPDATE:"+c)
		}
		sort.Strings(want)
		if got := one(`SELECT string_agg(x, ',' ORDER BY x COLLATE "C") FROM (SELECT `+lbl("a.grantee")+` || ':' || a.privilege_type || ':' || at.attname AS x
FROM pg_attribute at CROSS JOIN LATERAL aclexplode(at.attacl) a
WHERE at.attrelid = 'app.signup_sessions'::regclass AND at.attnum > 0 AND NOT at.attisdropped) s`, roles...); got != strings.Join(want, ",") {
			t.Errorf("signup_sessions column ACL:\n got %s\nwant %s", got, strings.Join(want, ","))
		}
		// The provision writer's column grants elsewhere: users SELECT (id, status) + UPDATE (status); organizations INSERT
		// (id, name, slug) + SELECT (id, slug); role_bindings INSERT of the eight 00015 columns; audit.events INSERT of the
		// nine claim-writer columns.
		if got := one(`SELECT string_agg(x, ',' ORDER BY x COLLATE "C") FROM (SELECT c.relname || '.' || at.attname || ':' || a.privilege_type AS x
FROM pg_attribute at JOIN pg_class c ON c.oid = at.attrelid CROSS JOIN LATERAL aclexplode(at.attacl) a
WHERE a.grantee = $1::regrole AND c.relname <> 'signup_sessions') s`, pw); got !=
			"events.action:INSERT,events.group_id:INSERT,events.organization_id:INSERT,events.outcome:INSERT,events.previous_role:INSERT,"+
				"events.resource_id:INSERT,events.resource_type:INSERT,events.role:INSERT,events.subject_user_id:INSERT,"+
				"organizations.id:INSERT,organizations.id:SELECT,organizations.name:INSERT,organizations.slug:INSERT,organizations.slug:SELECT,"+
				"role_bindings.group_id:INSERT,role_bindings.id:INSERT,role_bindings.organization_id:INSERT,role_bindings.role:INSERT,"+
				"role_bindings.scope_id:INSERT,role_bindings.scope_type:INSERT,role_bindings.source:INSERT,role_bindings.subject_user_id:INSERT,"+
				"users.id:SELECT,users.status:SELECT,users.status:UPDATE" {
			t.Errorf("provision-writer column grants: %s", got)
		}
		// The non-grants (Plan r5 SUP-4): no reader, no login transaction, no audit read, no delete, no DDL.
		if got := one(`SELECT format('%s|%s|%s|%s|%s|%s|%s', has_function_privilege($1, 'app.hostname_state(text, text)', 'EXECUTE'),
  has_function_privilege($1, 'app.is_organization_member(uuid, uuid)', 'EXECUTE'),
  has_table_privilege($1, 'app.login_transactions', 'SELECT, INSERT, UPDATE, DELETE'), has_table_privilege($1, 'audit.events', 'SELECT'),
  has_schema_privilege($1, 'app', 'CREATE'),
  (SELECT count(*) FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a WHERE a.grantee = $1::regrole AND a.privilege_type IN ('DELETE', 'TRUNCATE')),
  has_table_privilege($1, 'app.organizations', 'UPDATE, DELETE'))`, pw); got != "f|f|f|f|f|0|f" {
			t.Errorf("provision-writer non-grants (hostname_state|is_member|login_transactions|audit select|app create|deletes|org update): %s", got)
		}
		if got := one(`SELECT string_agg(p.proname, ',' ORDER BY p.proname COLLATE "C") FROM pg_proc p CROSS JOIN LATERAL aclexplode(p.proacl) a
WHERE a.grantee = $1::regrole AND a.grantee <> p.proowner`, pw); got != "ctx_group,ctx_login_binding,ctx_org,ctx_user" {
			t.Errorf("provision-writer EXECUTE grants: %s", got)
		}
		if got := one(`SELECT string_agg(n.nspname || ':' || a.privilege_type, ',' ORDER BY n.nspname COLLATE "C") FROM pg_namespace n
CROSS JOIN LATERAL aclexplode(n.nspacl) a WHERE a.grantee = $1::regrole`, pw); got != "app:USAGE,audit:USAGE" {
			t.Errorf("provision-writer schema grants: %s", got)
		}
	})

	t.Run("CT-SUP-1_definers", func(t *testing.T) {
		cfgPath := "search_path=pg_catalog, pg_temp"
		if got := one(`SELECT string_agg(p.proname || ':' || `+lbl("p.proowner")+` || ':' || p.prosecdef::text || ':' || p.provolatile::text || ':' ||
  coalesce(array_to_string(p.proconfig, '|'), '-') || ':' || coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C")
  FROM (SELECT `+lbl("a.grantee")+` || '=' || a.privilege_type AS v FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'),
  ';' ORDER BY p.proname COLLATE "C")
FROM pg_proc p WHERE p.oid IN ('app.auth_resolve_signup(bytea, timestamptz)'::regprocedure,
  'app.signup_provision_organization(bytea, text, text, bytea, bytea, text, timestamptz)'::regprocedure,
  'app.signup_challenge_issue(bytea, text, uuid, bytea, bytea, timestamptz)'::regprocedure,
  'app.signup_challenge_complete(bytea, bytea, text, uuid, timestamptz)'::regprocedure)`, roles...); got !=
			"auth_resolve_signup:ar:true:s:"+cfgPath+":ow=EXECUTE+rt=EXECUTE;"+
				"signup_challenge_complete:pw:true:v:"+cfgPath+":ow=EXECUTE+rt=EXECUTE;"+
				"signup_challenge_issue:pw:true:v:"+cfgPath+":ow=EXECUTE+rt=EXECUTE;"+
				"signup_provision_organization:pw:true:v:"+cfgPath+":ow=EXECUTE+rt=EXECUTE" {
			t.Errorf("00018 DEFINER bodies: %s", got)
		}
		if got := one(`SELECT format('%s|%s|%s', (SELECT count(*) FROM pg_proc WHERE proowner = $1::regrole),
  (SELECT count(*) FROM pg_proc WHERE proowner = $2::regrole), (SELECT count(*) FROM pg_proc WHERE proowner = $3::regrole))`, pw, ar, cw); got != "3|8|7" {
			t.Errorf("owned functions (provision writer|auth reader|claim writer): %s, want 3|8|7", got)
		}
	})

	// rolledBack runs f in a transaction of c that is always rolled back.
	rolledBack := func(t *testing.T, c *pgx.Conn, f func(tx pgx.Tx)) {
		t.Helper()
		tx, err := c.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		f(tx)
	}
	// attempt runs stmt under a savepoint and returns its SQLSTATE and constraint ("" when it succeeded).
	attempt := func(t *testing.T, tx pgx.Tx, stmt string, args ...any) (string, string) {
		t.Helper()
		sp, err := tx.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = sp.Rollback(ctx) }()
		_, err = sp.Exec(ctx, stmt, args...)
		return sigPgErr(err)
	}
	const provision = `SELECT * FROM app.signup_provision_organization($1, 'Acme', 'acme', $2, $3, 'https://acme.localhost:543', now())`
	const issue = `SELECT * FROM app.signup_challenge_issue($1, $2, $3::uuid, $4, $5, now())`
	const complete = `SELECT user_id::text FROM app.signup_challenge_complete($1, $2, $3, $4::uuid, now())`

	t.Run("CT-SUP-1_negatives", func(t *testing.T) {
		pargs := []any{sigDigest("cred-a0"), sigDigest("intent-x"), sigDigest("ref-x")}
		for _, c := range []struct{ name, user, org, binding, constraint string }{
			{"TENANT shape", sigUserA, sigOrg, "", "signup_context"},
			{"LOGIN shape", "", "", strings.Repeat("b1", 32), "signup_context"},
			{"no context", "", "", "", "signup_context"},
		} {
			rolledBack(t, rt, func(tx pgx.Tx) {
				rlsContext(t, tx, c.user, c.org, "", c.binding)
				if code, con := attempt(t, tx, provision, pargs...); code != "42501" || con != c.constraint {
					t.Errorf("provision in the %s: %s %s, want 42501 %s", c.name, code, con, c.constraint)
				}
			})
		}
		for _, c := range []struct{ name, user, org string }{{"IDENTITY shape", sigUserA, ""}, {"TENANT shape", sigUserA, sigOrg}, {"no context", "", ""}} {
			rolledBack(t, rt, func(tx pgx.Tx) {
				rlsContext(t, tx, c.user, c.org, "", "")
				if code, con := attempt(t, tx, issue, sigDigest("ref-a1"), sigTarget, sigOrg, sigDigest("c2"), sigDigest("n2")); code != "42501" ||
					con != "signup_continuation_context" {
					t.Errorf("challenge issue in the %s: %s %s, want 42501 signup_continuation_context", c.name, code, con)
				}
				if code, con := attempt(t, tx, complete, sigDigest("challenge-a1"), sigDigest("proof-a1"), sigTarget, sigOrg); code != "42501" ||
					con != "signup_continuation_context" {
					t.Errorf("challenge complete in the %s: %s %s, want 42501 signup_continuation_context", c.name, code, con)
				}
			})
		}
		// LOGIN shape: another Organization matches nothing and spends nothing; the right one matches exactly once.
		rolledBack(t, rt, func(tx pgx.Tx) {
			rlsContext(t, tx, "", "", "", strings.Repeat("b2", 32))
			n := func(org string) int {
				t.Helper()
				rows, err := tx.Query(ctx, complete, sigDigest("challenge-a1"), sigDigest("proof-a1"), sigTarget, org)
				if err != nil {
					t.Fatalf("complete: %v", err)
				}
				k := 0
				for rows.Next() {
					k++
				}
				rows.Close()
				if err := rows.Err(); err != nil {
					t.Fatalf("complete: %v", err)
				}
				return k
			}
			if k := n(sigOrgX); k != 0 {
				t.Errorf("completion with another Organization matched %d rows", k)
			}
			if k := n(sigOrg); k != 1 {
				t.Errorf("completion with the row's Organization matched %d rows, want 1", k)
			}
			if k := n(sigOrg); k != 0 {
				t.Errorf("completion replay matched %d rows", k)
			}
		})
		// No runtime or owner INSERT path on Organizations or org_owner bindings; the runtime cannot write the provision
		// writer's columns or a TENANT-shape signup row.
		for name, c := range map[string]*pgx.Conn{"runtime": rt, "owner": own} {
			rolledBack(t, c, func(tx pgx.Tx) {
				rlsContext(t, tx, sigUserA, "", "", "")
				if code, _ := attempt(t, tx, `INSERT INTO app.organizations (id, name, slug) VALUES (gen_random_uuid(), 'Direct', 'direct-org')`); code != "42501" {
					t.Errorf("%s Organization INSERT: %q, want 42501", name, code)
				}
				if code, _ := attempt(t, tx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1, $2, 'organization', $1, NULL, 'org_owner', 'membership')`, sigOrgX, sigUserA); code != "42501" {
					t.Errorf("%s org_owner binding INSERT: %q, want 42501", name, code)
				}
			})
		}
		rolledBack(t, rt, func(tx pgx.Tx) {
			rlsContext(t, tx, sigUserA, "", "", "")
			if code, _ := attempt(t, tx, `UPDATE app.signup_sessions SET ref_hash = $1 WHERE id = $2`, sigDigest("ref-y"), sigRowA1); code != "42501" {
				t.Errorf("runtime UPDATE of ref_hash: %q, want 42501", code)
			}
			if code, _ := attempt(t, tx, `UPDATE app.signup_sessions SET completed_at = now() WHERE id = $1`, sigRowA1); code != "42501" {
				t.Errorf("runtime UPDATE of completed_at: %q, want 42501", code)
			}
		})
		rolledBack(t, rt, func(tx pgx.Tx) {
			rlsContext(t, tx, sigUserA, sigOrg, "", "")
			if code, _ := attempt(t, tx, `INSERT INTO app.signup_sessions (id, user_id, origin, credential_hash, csrf_token, provider_session_id, expires_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, 'sid-t', now() + interval '10 minutes')`, sigUserA, sigRoot, sigDigest("cred-t"), sigDigest("csrf-t")); code != "42501" {
				t.Errorf("runtime TENANT-shape signup INSERT: %q, want 42501", code)
			}
		})
	})

	t.Run("CT-SUP-1_provision_writer", func(t *testing.T) {
		rolledBack(t, own, func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+quote(pw)); err != nil {
				t.Fatalf("SET ROLE provision writer: %v", err)
			}
			rlsContext(t, tx, sigUserA, "", "", "")
			var got string
			if err := tx.QueryRow(ctx, `SELECT coalesce(string_agg(id::text, ',' ORDER BY id), '') FROM app.signup_sessions`).Scan(&got); err != nil ||
				got != sigRowA0+","+sigRowA1 {
				t.Errorf("provision-writer visible rows: %q %v, want the pending and the provisioned row of the active user only", got, err)
			}
			if code, _ := attempt(t, tx, `UPDATE app.signup_sessions SET proof_hash = $1 WHERE id = $2`, sigDigest("proof-z"), sigRowA1); code != "42501" {
				t.Errorf("provision-writer proof set without completion: %q, want 42501 (WITH CHECK)", code)
			}
			if code, _ := attempt(t, tx, `UPDATE app.signup_sessions SET continuation_generation = continuation_generation WHERE id = $1`, sigRowA0); code != "" {
				t.Errorf("provision-writer UPDATE of the pending row: %q, want admitted", code)
			}
			for what, stmt := range map[string]string{
				"ungranted column":   `UPDATE app.signup_sessions SET revoked_at = now(), revoke_reason = 'x' WHERE id = '` + sigRowA0 + `'`,
				"DELETE":             `DELETE FROM app.signup_sessions`,
				"login_transactions": `SELECT 1 FROM app.login_transactions`,
				"hostname_state":     `SELECT * FROM app.hostname_state('platform', 'sigcat')`,
				"is_member":          `SELECT app.is_organization_member('` + sigUserA + `', '` + sigOrg + `')`,
				"audit SELECT":       `SELECT 1 FROM audit.events`,
				"depth-0 audit INSERT": `INSERT INTO audit.events (action, outcome, organization_id) VALUES ('organization.created', 'success', '` +
					sigOrg + `')`,
			} {
				if code, _ := attempt(t, tx, stmt); code != "42501" {
					t.Errorf("provision writer %s: %q, want 42501", what, code)
				}
			}
		})
	})

	t.Run("CT-SUP-1_return_path", func(t *testing.T) {
		rolledBack(t, setup, func(tx pgx.Tx) {
			ins := `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, expires_at)
VALUES ($1, $2, repeat('v', 43), 'c', $3, $4, now() + interval '1 minute')`
			if code, _ := attempt(t, tx, ins, sigDigest("s-ok"), sigDigest("b"), sigRoot, "/signup/details"); code != "" {
				t.Errorf("/signup/details: %q, want admitted", code)
			}
			for _, p := range []string{"/signup", "/signup/details/", "//evil", "https://evil", "/%2Fevil", "/signup/../account"} {
				if code, con := attempt(t, tx, ins, sigDigest("s-"+p), sigDigest("b"), sigRoot, p); code != "23514" || con != "login_transactions_return_path_check" {
					t.Errorf("return path %q: %s %s, want 23514 login_transactions_return_path_check", p, code, con)
				}
			}
			if code, con := attempt(t, tx, `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path,
  expires_at, organization_id, signup_handoff) VALUES ($1, $2, repeat('v', 43), 'c', $3, '/', now() + interval '1 minute', $4, true)`,
				sigDigest("s-h"), sigDigest("b"), sigTarget, sigOrg); code != "23514" || con != "login_transactions_signup_handoff_shape" {
				t.Errorf("handoff without its expected user: %s %s, want 23514 login_transactions_signup_handoff_shape", code, con)
			}
			if code, con := attempt(t, tx, `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path,
  expires_at, organization_id, signup_intent) VALUES ($1, $2, repeat('v', 43), 'c', $3, '/signup/details', now() + interval '1 minute', $4, true)`,
				sigDigest("s-i"), sigDigest("b"), sigRoot, sigOrg); code != "23514" || con != "login_transactions_signup_intent_shape" {
				t.Errorf("signup intent with an Organization: %s %s, want 23514 login_transactions_signup_intent_shape", code, con)
			}
		})
	})

	t.Run("CT-SUP-1_updown", func(t *testing.T) {
		down := func() (string, error) {
			var b strings.Builder
			err := r.Down(ctx, &b)
			return b.String(), err
		}
		withRows := state(r)
		if out, err := down(); err == nil || !strings.Contains(err.Error(),
			"refusing to roll back 00018: 3 signup sessions / 0 login transactions / 0 audit rows (history is never discarded)") {
			t.Fatalf("down with signup rows: %v %q", err, out)
		}
		if got := state(r); got != withRows {
			t.Fatal("the refused down changed the catalog")
		}
		exec(`DELETE FROM app.signup_sessions`)
		out, err := down()
		if err != nil || out != "rolled_back\t00018_platform_signup.sql\ncurrent_version\t17\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		metadataThroughVersion17(t, cfg, r)
		if got := state(r); got != s17 {
			t.Fatalf("catalog after down differs from version 17:\n got %s\nwant %s", got, s17)
		}
		objRunUp(t, r, "applied\t00018_platform_signup.sql\ncurrent_version\t18\n")
		if got := state(r); got != s18 {
			t.Fatalf("catalog after down/up differs from version 18:\n got %s\nwant %s", got, s18)
		}
	})
}

// nopWriter discards a runner's output.
type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }
