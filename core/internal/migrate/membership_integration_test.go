//go:build integration

// G019 B-MEMBERSHIP-CATALOG (Plan r1 §4, §6.2 CT-MEM-1 as amended by the r2 addendum §5/§6 (D-14a), the r3 addendum C1,
// the r4 addendum (events_insert_claim at depth >= 1), Result addendum A1-01 (the own-Organization reader: eight DEFINER
// helpers = five policy-reader readers + three claim-writer bodies) and the DV-1/DV-2 disposition): TestMembershipCatalog.
//   - CT-MEM-1/preflight the legacy owner preflight: an ownerless Organization and, separately, an ownerless Group (synthetic
//     rows committed at version 14 as the superuser) each refuse the up with counts only; after each refusal and its
//     rollback all three lifted tables keep RLS and FORCE, the version stays 14 and no partial 00015 object, policy, grant
//     or body remains (DV-1); each passes once resolved through an owner binding.
//   - CT-MEM-1 the version-15 delta: the invitation table, its CHECKs, indexes and triggers; every new or replaced function
//     (owner, SECURITY DEFINER, volatility, search_path, ACL; the eight DEFINER helpers and the reviewed bodies); the claim
//     writer's whole authority (role attributes, membership, column grants, policies, the three-function inventory); the
//     policy strings with their load-bearing predicates; the events_attribute body = the 00013 body + the invitation shape
//     (hash-diff); the retained RLS/FORCE flags.
//   - CT-MEM-1 least privilege under actual grants: the runtime never reads the digest, writes the consumption columns,
//     deletes an invitation or executes the Organization claim; a foreign Organization's rows stay invisible.
//   - CT-MEM-AUD the widened audit constraints (three invitation actions, 36 + 17 = 53 operations, the invitation resource).
//   - CT-MEM-UPDOWN a clean down to the exact version-14 catalog (policies, grants and bodies restored) and up again; the
//     down refused transactionally, without losing history, while an invitation or an invitation audit row exists.
//
// Assertions are written from the Plan and the reviewed migration, not from an observed run (NOT EXECUTED until the
// native DB qualification).
package migrate

import (
	"bytes"
	"context"
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/migrations"
)

const (
	memOrgA     = "0f19c000-0000-4000-8000-000000000001"
	memOrgB     = "0f19c000-0000-4000-8000-000000000002"
	memOrgX     = "0f19c000-0000-4000-8000-000000000003" // ownerless Organization (preflight)
	memGroupA   = "0f19c000-0000-4000-8000-000000000004"
	memGroupY   = "0f19c000-0000-4000-8000-000000000005" // ownerless Group in A (preflight)
	memOwnerA   = "0f19c000-0000-4000-8000-000000000011" // org_owner A, group_owner memGroupA
	memAdminA   = "0f19c000-0000-4000-8000-000000000012" // org_admin A
	memMemberA  = "0f19c000-0000-4000-8000-000000000013" // org_member A
	memOwnerB   = "0f19c000-0000-4000-8000-000000000014" // org_owner B
	memInviteB  = "0f19c000-0000-4000-8000-000000000021"
	memInviteUD = "0f19c000-0000-4000-8000-000000000022"
)

// memOps are the seventeen G019 access.denied operations (r2 §7; the 00015 down list).
var memOps = []string{"membership.list", "membership.self", "membership.change_role", "membership.remove", "membership.leave",
	"ownership.promote", "ownership.demote", "ownership.transfer", "invitation.list", "invitation.mine",
	"invitation.candidates", "invitation.issue", "invitation.revoke", "invitation.get", "invitation.accept",
	"invitation.consume", "retirement.eligibility"}

// memBody returns the dollar-quoted body (prosrc) of the first function created after marker in the up section of file.
func memBody(t *testing.T, file, marker string) string {
	t.Helper()
	b, err := fs.ReadFile(migrations.FS, file)
	if err != nil {
		t.Fatal(err)
	}
	up, _, _ := strings.Cut(string(b), "-- +goose Down")
	i := strings.Index(up, marker)
	if i < 0 {
		t.Fatalf("%s: %q not found", file, marker)
	}
	j := strings.Index(up[i:], "AS $$")
	if j < 0 {
		t.Fatalf("%s: no body after %q", file, marker)
	}
	start := i + j + len("AS $$")
	k := strings.Index(up[start:], "$$")
	return up[start : start+k]
}

func TestMembershipCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r14 := runnerFor(t, cfg, subsetFS(t, 14))
	objRunUp(t, r14, "current_version\t14\n")
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	rt := rlsRuntimeConn(t, cfg)
	o, rtr, pr, ar, cw := cfg.OwnerRole(), cfg.RuntimeRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader",
		cfg.Database+"_claim_writer"
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
	// replica runs sql as the superuser with triggers (and FK triggers) disabled, committed: history and foreign rows
	// written without their capture, the onboarding/domains catalog pattern.
	replica := func(sql string, args ...any) {
		t.Helper()
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("replica fixture: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	// state is the catalog a partial or committed 00015 would change: relations, ACLs and history (snapshot); every app
	// and audit function with owner, SECURITY DEFINER, ACL and body hash; every policy with roles and predicates; every
	// column ACL; the RLS/FORCE flags; the schema ACLs; the constraints and the non-internal triggers.
	state := func() string {
		t.Helper()
		return snapshot(t, r14) + "\n" + one(`SELECT concat_ws(E'\n--\n',
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
	flags := func() string {
		t.Helper()
		return one(`SELECT string_agg(relname || ':' || relrowsecurity::text || ':' || relforcerowsecurity::text, ',' ORDER BY relname)
FROM pg_class WHERE oid IN ('app.organizations'::regclass, 'app.groups'::regclass, 'app.role_bindings'::regclass)`)
	}
	const flagsOn = "groups:true:true,organizations:true:true,role_bindings:true:true"

	// The version-14 fixture: two owned Organizations, an owned Group, an admin and a member.
	exec(`INSERT INTO app.users (id, display_name) VALUES ('` + memOwnerA + `', 'Mem Owner A'), ('` + memAdminA + `', 'Mem Admin A'),
  ('` + memMemberA + `', 'Mem Member A'), ('` + memOwnerB + `', 'Mem Owner B')`)
	exec(`INSERT INTO app.organizations (id, name) VALUES ('` + memOrgA + `', 'Membership A'), ('` + memOrgB + `', 'Membership B')`)
	exec(`INSERT INTO app.groups (id, organization_id, name, slug) VALUES ('` + memGroupA + `', '` + memOrgA + `', 'Owned', 'owned')`)
	exec(`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + memOrgA + `', '` + memOwnerA + `', 'organization', '` + memOrgA + `', NULL, 'org_owner', 'membership'),
  ('` + memOrgA + `', '` + memAdminA + `', 'organization', '` + memOrgA + `', NULL, 'org_admin', 'membership'),
  ('` + memOrgA + `', '` + memMemberA + `', 'organization', '` + memOrgA + `', NULL, 'org_member', 'membership'),
  ('` + memOrgA + `', '` + memOwnerA + `', 'group', '` + memGroupA + `', '` + memGroupA + `', 'group_owner', 'membership'),
  ('` + memOrgB + `', '` + memOwnerB + `', 'organization', '` + memOrgB + `', NULL, 'org_owner', 'membership')`)
	s14 := state()
	if got := flags(); got != flagsOn {
		t.Fatalf("version-14 flags: %s", got)
	}
	r := runnerFor(t, cfg, subsetFS(t, 15))
	// refused asserts the preflight refusal with counts only and the DV-1 postconditions after its rollback.
	refused := func(t *testing.T, wantMsg string, ids ...string) {
		t.Helper()
		var b bytes.Buffer
		err := r.Up(ctx, &b)
		if err == nil || !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("up with an ownerless record: %v %q, want %q", err, b.String(), wantMsg)
		}
		if sqlstateOf(err) != "" && sqlstateOf(err) != "23000" {
			t.Errorf("preflight SQLSTATE %s, want 23000", sqlstateOf(err))
		}
		for _, id := range ids {
			if strings.Contains(err.Error(), id) {
				t.Errorf("the preflight names a record (%s): %v", id, err)
			}
		}
		if got := flags(); got != flagsOn {
			t.Errorf("RLS/FORCE after the rolled-back preflight: %s, want %s", got, flagsOn)
		}
		if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t14\n" {
			t.Errorf("version after the refusal: %q", out)
		}
		if got := one(`SELECT format('%s:%s:%s', to_regclass('app.invitations') IS NULL,
  (SELECT count(*) FROM pg_proc WHERE proowner = $1::regrole), to_regprocedure('app.rls_owner_count(text, uuid)') IS NULL)`, cw); got != "t:0:t" {
			t.Errorf("partial 00015 objects after the refusal: %s", got)
		}
		if got := state(); got != s14 {
			t.Errorf("catalog after the refused up differs from version 14:\n got %s\nwant %s", got, s14)
		}
	}

	t.Run("CT-MEM-1/preflight an ownerless Organization is refused with counts only, then resolved", func(t *testing.T) {
		exec(`INSERT INTO app.organizations (id, name) VALUES ('` + memOrgX + `', 'Ownerless X')`)
		refused(t, "legacy owner preflight: 1 organizations / 0 groups without an owner binding", memOrgX, "Ownerless X")
		// Resolved through an owner binding (the documented restricted owner procedure's result).
		exec(`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + memOrgX + `', '` + memOwnerA + `', 'organization', '` + memOrgX + `', NULL, 'org_owner', 'membership')`)
	})

	t.Run("CT-MEM-1/preflight an ownerless Group is refused with counts only, then resolved", func(t *testing.T) {
		exec(`INSERT INTO app.groups (id, organization_id, name, slug) VALUES ('` + memGroupY + `', '` + memOrgA + `', 'Ownerless Y', 'ownerless-y')`)
		refused(t, "legacy owner preflight: 0 organizations / 1 groups without an owner binding", memGroupY, "Ownerless Y")
		exec(`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + memOrgA + `', '` + memOwnerA + `', 'group', '` + memGroupY + `', '` + memGroupY + `', 'group_owner', 'membership')`)
	})

	// Resolved: the up applies.
	objRunUp(t, r, "applied\t00015_membership_invitations.sql\ncurrent_version\t15\n")
	s15 := state()

	t.Run("CT-MEM-1 the version-15 delta", func(t *testing.T) {
		metadataThroughVersion15Only(t, cfg, r)
		if got := one(`SELECT string_agg(relname || ':' || relrowsecurity::text || ':' || relforcerowsecurity::text, ',' ORDER BY relname)
FROM pg_class WHERE oid IN ('app.organizations'::regclass, 'app.groups'::regclass, 'app.role_bindings'::regclass,
  'app.invitations'::regclass)`); got != "groups:true:true,invitations:true:true,organizations:true:true,role_bindings:true:true" {
			t.Errorf("RLS/FORCE at version 15: %s", got)
		}
		// The invitation table: columns, CHECKs, keys and indexes.
		if got := one(`SELECT string_agg(a.attname || ':' || format_type(a.atttypid, a.atttypmod) || ':' || a.attnotnull::text, ',' ORDER BY a.attnum)
FROM pg_attribute a WHERE a.attrelid = 'app.invitations'::regclass AND a.attnum > 0 AND NOT a.attisdropped`); got !=
			"id:uuid:true,organization_id:uuid:true,scope_type:text:true,scope_id:uuid:true,group_id:uuid:false,role:text:true,"+
				"issuer_user_id:uuid:true,intended_user_id:uuid:false,intended_address:text:false,token_digest:bytea:true,"+
				"created_at:timestamp with time zone:true,expires_at:timestamp with time zone:true,revoked_at:timestamp with time zone:false,"+
				"consumed_at:timestamp with time zone:false,consumed_by_user_id:uuid:false,consumed_binding_id:uuid:false" {
			t.Errorf("invitation columns: %s", got)
		}
		if got := one(`SELECT string_agg(conname || ':' || contype::text, ',' ORDER BY conname) FROM pg_constraint
WHERE conrelid = 'app.invitations'::regclass`); got != "invitations_address_normalized:c,invitations_consumed_by_fkey:f,"+
			"invitations_consumed_is_intended:c,invitations_consumption_shape:c,invitations_group_fkey:f,invitations_id_nonzero:c,"+
			"invitations_intended_user_fkey:f,invitations_issuer_fkey:f,invitations_lifetime:c,invitations_organization_fkey:f,"+
			"invitations_pkey:p,invitations_scope_consistent:c,invitations_token_digest_key:u,invitations_token_digest_length:c" {
			t.Errorf("invitation constraints: %s", got)
		}
		// Every foreign key into the tenancy and identity tables restricts deletion (the FK inventory).
		if got := one(`SELECT string_agg(conname || ':' || confdeltype::text, ',' ORDER BY conname) FROM pg_constraint
WHERE conrelid = 'app.invitations'::regclass AND contype = 'f'`); got != "invitations_consumed_by_fkey:r,invitations_group_fkey:r,"+
			"invitations_intended_user_fkey:r,invitations_issuer_fkey:r,invitations_organization_fkey:r" {
			t.Errorf("invitation foreign keys: %s", got)
		}
		if def := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = 'invitations_scope_consistent'`); !strings.Contains(def,
			"'org_admin'") || !strings.Contains(def, "'group_member'") || strings.Contains(def, "owner") {
			t.Errorf("scope CHECK admits an owner role or lacks an invitable role: %s", def)
		}
		if def := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = 'invitations_lifetime'`); !strings.Contains(def, "30 days") {
			t.Errorf("lifetime CHECK: %s", def)
		}
		if got := one(`SELECT string_agg(c.relname, ',' ORDER BY c.relname) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
WHERE i.indrelid = 'app.invitations'::regclass`); got != "invitations_pending_idx,invitations_pkey,invitations_recipient_idx,"+
			"invitations_scope_idx,invitations_token_digest_key" {
			t.Errorf("invitation indexes: %s", got)
		}
		// Triggers (postcondition 5) with the exact firing definitions.
		if got := one(`SELECT string_agg(pg_get_triggerdef(t.oid), E'\n' ORDER BY t.tgname) FROM pg_trigger t
WHERE t.tgrelid = 'app.invitations'::regclass AND NOT t.tgisinternal`); got !=
			"CREATE TRIGGER audit_capture_insert AFTER INSERT ON app.invitations FOR EACH ROW EXECUTE FUNCTION audit.capture_invitations()\n"+
				"CREATE TRIGGER audit_capture_update AFTER UPDATE ON app.invitations FOR EACH ROW EXECUTE FUNCTION audit.capture_invitations()\n"+
				"CREATE TRIGGER invitations_bound BEFORE INSERT ON app.invitations FOR EACH ROW EXECUTE FUNCTION app.invitations_bound()\n"+
				"CREATE TRIGGER invitations_keys_immutable BEFORE UPDATE ON app.invitations FOR EACH ROW EXECUTE FUNCTION app.invitations_keys_immutable()" {
			t.Errorf("invitation triggers:\n%s", got)
		}
		if got := one(`SELECT string_agg(t.tgname, ',' ORDER BY t.tgname) FROM pg_trigger t
WHERE t.tgrelid = 'app.role_bindings'::regclass AND NOT t.tgisinternal`); got !=
			"audit_capture_delete,audit_capture_insert,audit_capture_update,role_bindings_authority_guard,role_bindings_keys_immutable" {
			t.Errorf("role_bindings triggers: %s", got)
		}
		if got := one(`SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE tgrelid = 'app.role_bindings'::regclass
  AND tgname = 'role_bindings_authority_guard'`); got != "CREATE TRIGGER role_bindings_authority_guard BEFORE DELETE OR UPDATE OF role "+
			"ON app.role_bindings FOR EACH ROW EXECUTE FUNCTION app.role_bindings_authority_guard()" {
			t.Errorf("authority guard trigger: %s", got)
		}

		// Postcondition 1 repeated: every new or replaced function (owner, DEFINER, volatility, configuration, ACL).
		labels := func(col string) string {
			return `CASE ` + col + ` WHEN 0 THEN 'public' WHEN $1::regrole THEN 'rt' WHEN $2::regrole THEN 'ow' WHEN $3::regrole THEN 'pr'
  WHEN $4::regrole THEN 'cw' WHEN $5::regrole THEN 'ar' ELSE 'other' END`
		}
		roles := []any{rtr, o, pr, cw, ar}
		if got := one(`SELECT string_agg(p.proname || ':' || `+labels("p.proowner")+` || ':' || p.prosecdef::text || ':' || p.provolatile::text
  || ':' || coalesce(array_to_string(p.proconfig, '|'), '-') || ':' || coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C")
  FROM (SELECT `+labels("a.grantee")+` || '=' || a.privilege_type AS v FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'),
  ';' ORDER BY p.proname)
FROM pg_proc p WHERE p.oid IN ('app.invitations_keys_immutable()'::regprocedure, 'app.invitations_bound()'::regprocedure,
  'audit.capture_invitations()'::regprocedure, 'app.role_bindings_authority_guard()'::regprocedure,
  'app.rls_owner_count(text, uuid)'::regprocedure, 'app.rls_dependent_group_bindings(uuid)'::regprocedure,
  'app.rls_sole_owner_scopes()'::regprocedure, 'app.rls_own_organizations(uuid, integer)'::regprocedure,
  'app.rls_group_invitation_candidates(uuid, uuid, integer)'::regprocedure, 'app.rls_user_enabled_lock(uuid, text, uuid)'::regprocedure,
  'app.invitation_claim_group(uuid, bytea)'::regprocedure, 'app.invitation_claim_organization(uuid, bytea, text, text, text)'::regprocedure)`,
			roles...); got != "capture_invitations:ow:false:v:search_path=pg_catalog, pg_temp:-;"+
			"invitation_claim_group:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
			"invitation_claim_organization:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE;"+
			"invitations_bound:ow:false:v:search_path=pg_catalog:-;"+
			"invitations_keys_immutable:ow:false:v:search_path=pg_catalog:-;"+
			"rls_dependent_group_bindings:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
			"rls_group_invitation_candidates:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
			"rls_own_organizations:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
			"rls_owner_count:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
			"rls_sole_owner_scopes:pr:true:s:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
			"rls_user_enabled_lock:cw:true:v:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
			"role_bindings_authority_guard:ow:false:v:search_path=pg_catalog:-" {
			t.Errorf("membership functions:\n got %s", got)
		}
		// The DEFINER allowance of 00015: exactly eight helpers, five policy-reader readers and three claim-writer bodies.
		if got := one(`SELECT string_agg(`+labels("p.proowner")+` || ':' || p.proname, ',' ORDER BY p.proname)
FROM pg_proc p WHERE p.prosecdef AND p.pronamespace = 'app'::regnamespace AND p.proowner IN ($3::regrole, $4::regrole)
  AND p.proname IN ('rls_owner_count', 'rls_dependent_group_bindings', 'rls_sole_owner_scopes', 'rls_own_organizations',
    'rls_group_invitation_candidates', 'rls_user_enabled_lock', 'invitation_claim_group', 'invitation_claim_organization')`, roles...); got !=
			"cw:invitation_claim_group,cw:invitation_claim_organization,pr:rls_dependent_group_bindings,pr:rls_group_invitation_candidates,"+
				"pr:rls_own_organizations,pr:rls_owner_count,pr:rls_sole_owner_scopes,cw:rls_user_enabled_lock" {
			t.Errorf("00015 DEFINER helpers: %s", got)
		}
		// The owned-function inventories: the claim writer owns exactly its three bodies (postcondition 2); the policy
		// reader the three earlier readers plus its five; the auth reader's six are unchanged (no write body added).
		for _, c := range []struct{ role, want string }{
			{cw, "invitation_claim_group,invitation_claim_organization,rls_user_enabled_lock"},
			{pr, "rls_dependent_group_bindings,rls_group_invitation_candidates,rls_group_role,rls_org_role,rls_own_organizations," +
				"rls_owner_count,rls_profile_eligible,rls_sole_owner_scopes"},
			{ar, "auth_logout_ref,auth_principal_subject,auth_resolve_session,hostname_state,is_organization_member,public_branding"},
		} {
			if got := one(`SELECT string_agg(proname, ',' ORDER BY proname) FROM pg_proc WHERE proowner = $1::regrole`, c.role); got != c.want {
				t.Errorf("functions owned by %s: %s, want %s", c.role, got, c.want)
			}
		}
		// The reviewed bodies: the eight helpers and the invitation capture equal the migration text.
		for _, f := range []struct{ sig, marker string }{
			{"app.rls_owner_count(text, uuid)", "CREATE FUNCTION app.rls_owner_count("},
			{"app.rls_dependent_group_bindings(uuid)", "CREATE FUNCTION app.rls_dependent_group_bindings("},
			{"app.rls_sole_owner_scopes()", "CREATE FUNCTION app.rls_sole_owner_scopes("},
			{"app.rls_own_organizations(uuid, integer)", "CREATE FUNCTION app.rls_own_organizations("},
			{"app.rls_group_invitation_candidates(uuid, uuid, integer)", "CREATE FUNCTION app.rls_group_invitation_candidates("},
			{"app.rls_user_enabled_lock(uuid, text, uuid)", "CREATE FUNCTION app.rls_user_enabled_lock("},
			{"app.invitation_claim_group(uuid, bytea)", "CREATE FUNCTION app.invitation_claim_group("},
			{"app.invitation_claim_organization(uuid, bytea, text, text, text)", "CREATE FUNCTION app.invitation_claim_organization("},
			{"audit.capture_invitations()", "CREATE FUNCTION audit.capture_invitations("},
		} {
			if got := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = $1::regprocedure`, f.sig); got !=
				md5hex(memBody(t, "00015_membership_invitations.sql", f.marker)) {
				t.Errorf("%s differs from the reviewed migration body", f.sig)
			}
		}
		// The trigger functions lock with one command and count with a separate one (READ COMMITTED: the count's snapshot
		// is taken after the lock wait); the guard's role set is runtime, owner and claim writer.
		bound := one(`SELECT prosrc FROM pg_proc WHERE oid = 'app.invitations_bound()'::regprocedure`)
		if !strings.Contains(bound, "FOR NO KEY UPDATE;\n") || !strings.Contains(bound, "SELECT count(*) INTO n FROM app.invitations i") ||
			strings.Index(bound, "FOR NO KEY UPDATE") > strings.Index(bound, "SELECT count(*)") || !strings.Contains(bound, "n >= 100") ||
			!strings.Contains(bound, "current_user::text NOT IN ('"+rtr+"', '"+o+"')") {
			t.Errorf("invitations_bound body: %s", bound)
		}
		guard := one(`SELECT prosrc FROM pg_proc WHERE oid = 'app.role_bindings_authority_guard()'::regprocedure`)
		for _, frag := range []string{"current_user::text NOT IN ('" + rtr + "', '" + o + "', '" + cw + "')",
			"app.rls_user_enabled_lock(OLD.subject_user_id, OLD.scope_type, OLD.scope_id)",
			"app.is_organization_member(OLD.subject_user_id, OLD.organization_id)", "app.ctx_group() IS NOT NULL",
			"app.rls_dependent_group_bindings(OLD.subject_user_id) > 0", "FOR NO KEY UPDATE;\n",
			"app.rls_owner_count(OLD.scope_type, OLD.scope_id) - 1 < 1", "CONSTRAINT = 'last_owner'", "CONSTRAINT = 'owner_authority'",
			"CONSTRAINT = 'recipient_ineligible'", "CONSTRAINT = 'dependent_bindings'", "CONSTRAINT = 'context_required'"} {
			if !strings.Contains(guard, frag) {
				t.Errorf("authority guard lacks %q", frag)
			}
		}
		if strings.Index(guard, "FOR NO KEY UPDATE") > strings.Index(guard, "app.rls_owner_count(") {
			t.Error("authority guard counts before it locks the scope row")
		}
		// events_attribute = the 00013 body + the invitation shape (hash-diff); no context parser added.
		b15 := memBody(t, "00015_membership_invitations.sql", "CREATE OR REPLACE FUNCTION audit.events_attribute()")
		if got := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure`); got != md5hex(b15) {
			t.Error("events_attribute differs from the reviewed 00015 body")
		}
		stripped := strings.Replace(b15, "'organization.branding_changed',\n    'invitation.issued', 'invitation.revoked', 'invitation.consumed');",
			"'organization.branding_changed');", 1)
		i := strings.Index(stripped, "    WHEN NEW.action IN ('invitation.issued', 'invitation.revoked', 'invitation.consumed') THEN\n")
		j := strings.Index(stripped, "    WHEN NEW.action = 'auth.sign_in' THEN\n")
		if stripped == b15 || i < 0 || j < i {
			t.Fatalf("events_attribute: the invitation shape is not where the review placed it (%d, %d)", i, j)
		}
		if stripped = stripped[:i] + stripped[j:]; md5hex(stripped) != md5hex(memBody(t, "00013_organization_branding.sql",
			"CREATE OR REPLACE FUNCTION audit.events_attribute()")) {
			t.Error("events_attribute without the invitation shape differs from the 00013 body")
		}
		if got := one(`SELECT string_agg(proname, ',' ORDER BY proname) FROM pg_proc
WHERE pronamespace = 'app'::regnamespace AND proname LIKE 'ctx\_%'`); got != "ctx_group,ctx_login_binding,ctx_org,ctx_user" {
			t.Errorf("context parsers: %s", got)
		}
		// The audit capture functions stay SECURITY INVOKER (their execution mode is the caller's).
		if got := one(`SELECT string_agg(proname || ':' || prosecdef::text, ',' ORDER BY proname) FROM pg_proc
WHERE oid IN ('audit.capture_role_bindings()'::regprocedure, 'audit.capture_invitations()'::regprocedure)`); got !=
			"capture_invitations:false,capture_role_bindings:false" {
			t.Errorf("capture execution mode: %s", got)
		}

		// The claim writer: NOLOGIN NOINHERIT, no bypass, no credential; the owner its only member (INHERIT FALSE, SET TRUE);
		// a member of nothing.
		var attrs string
		if err := setup.QueryRow(ctx, `SELECT format('%s:%s:%s:%s:%s:%s', rolcanlogin, rolinherit, rolbypassrls, rolsuper, rolcreaterole,
  rolpassword IS NULL) FROM pg_authid WHERE rolname = $1`, cw).Scan(&attrs); err != nil || attrs != "f:f:f:f:f:t" {
			t.Errorf("claim writer attributes: %q %v", attrs, err)
		}
		if got := one(`SELECT string_agg(m.member::regrole::text || ':' || m.inherit_option::text || ':' || m.set_option::text, ',')
FROM pg_auth_members m WHERE m.roleid = $1::regrole`, cw); got != o+":false:true" {
			t.Errorf("claim writer members: %s", got)
		}
		if got := one(`SELECT count(*)::text FROM pg_auth_members WHERE member = $1::regrole`, cw); got != "0" {
			t.Errorf("claim writer memberships: %s", got)
		}
		// Its whole authority: column grants, one table-level SELECT, schema usage and EXECUTE on exactly six functions.
		colgrants := func(grantee string) string {
			return one(`SELECT string_agg(x, ',' ORDER BY x COLLATE "C") FROM (SELECT c.oid::regclass::text || '.' || at.attname || ':' ||
  a.privilege_type AS x FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
  CROSS JOIN LATERAL aclexplode(at.attacl) a WHERE a.grantee = $1::regrole AND c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace)) s`, grantee)
		}
		var want []string
		add := func(rel, priv string, cols ...string) {
			for _, c := range cols {
				want = append(want, rel+"."+c+":"+priv)
			}
		}
		add("app.identity_principals", "SELECT", "issuer", "subject", "user_id")
		add("app.invitations", "UPDATE", "consumed_at", "consumed_by_user_id", "intended_user_id", "consumed_binding_id")
		rbCols := []string{"id", "organization_id", "subject_user_id", "scope_type", "scope_id", "group_id", "role", "source"}
		add("app.role_bindings", "SELECT", rbCols...)
		add("app.role_bindings", "INSERT", rbCols...)
		add("app.role_bindings", "UPDATE", "role")
		add("app.users", "SELECT", "id", "status")
		add("app.users", "UPDATE", "status")
		add("audit.events", "INSERT", "action", "outcome", "organization_id", "group_id", "resource_type", "resource_id",
			"subject_user_id", "role", "previous_role")
		if got, w := colgrants(cw), sortedJoin(want); got != w {
			t.Errorf("claim writer column grants:\n got %s\nwant %s", got, w)
		}
		if got := one(`SELECT coalesce(string_agg(c.oid::regclass::text || ':' || a.privilege_type, ',' ORDER BY c.oid::regclass::text,
  a.privilege_type), '') FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a WHERE a.grantee = $1::regrole`, cw); got != "app.invitations:SELECT" {
			t.Errorf("claim writer table grants: %s", got)
		}
		if got := one(`SELECT string_agg(n.nspname || ':' || a.privilege_type, ',' ORDER BY n.nspname) FROM pg_namespace n
CROSS JOIN LATERAL aclexplode(n.nspacl) a WHERE a.grantee = $1::regrole`, cw); got != "app:USAGE,audit:USAGE" {
			t.Errorf("claim writer schema grants: %s", got)
		}
		if got := one(`SELECT string_agg(x, ',' ORDER BY x COLLATE "C") FROM (SELECT p.oid::regprocedure::text AS x FROM pg_proc p
CROSS JOIN LATERAL aclexplode(p.proacl) a WHERE a.grantee = $1::regrole AND p.proowner <> $1::regrole) s`, cw); got !=
			"app.ctx_group(),app.ctx_org(),app.ctx_user(),app.is_organization_member(uuid,uuid),app.rls_group_role(uuid),app.rls_org_role()" {
			t.Errorf("claim writer EXECUTE: %s", got)
		}
		// Runtime column grants on the invitation table: SELECT without the digest, INSERT with it, UPDATE (revoked_at) only.
		rtInv := one(`SELECT string_agg(x, ',' ORDER BY x COLLATE "C") FROM (SELECT at.attname || ':' || a.privilege_type AS x
FROM pg_attribute at CROSS JOIN LATERAL aclexplode(at.attacl) a
WHERE at.attrelid = 'app.invitations'::regclass AND at.attnum > 0 AND a.grantee = $1::regrole) s`, rtr)
		want = nil
		add("i", "SELECT", "id", "organization_id", "scope_type", "scope_id", "group_id", "role", "issuer_user_id", "intended_user_id",
			"intended_address", "created_at", "expires_at", "revoked_at", "consumed_at", "consumed_by_user_id", "consumed_binding_id")
		add("i", "INSERT", "id", "organization_id", "scope_type", "scope_id", "group_id", "role", "issuer_user_id", "intended_user_id",
			"intended_address", "token_digest", "expires_at")
		add("i", "UPDATE", "revoked_at")
		for k := range want {
			want[k] = strings.TrimPrefix(want[k], "i.")
		}
		if w := sortedJoin(want); rtInv != w {
			t.Errorf("runtime invitation column grants:\n got %s\nwant %s", rtInv, w)
		}
		if got := one(`SELECT coalesce(string_agg(a.grantee::regrole::text || ':' || a.privilege_type, ',' ORDER BY a.grantee::regrole::text,
  a.privilege_type), '') FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a WHERE c.oid = 'app.invitations'::regclass AND a.grantee <> c.relowner`); got !=
			cw+":SELECT" {
			t.Errorf("invitation table-level grants: %s", got)
		}
		// Runtime on role_bindings gains DELETE and SELECT (id) and no new INSERT shape; the policy reader reads users.
		for _, c := range []struct {
			role, rel, col, priv string
			want                 bool
		}{
			{rtr, "app.role_bindings", "id", "SELECT", true}, {rtr, "app.role_bindings", "id", "INSERT", false},
			{pr, "app.users", "id", "SELECT", true}, {pr, "app.users", "display_name", "SELECT", true}, {pr, "app.users", "status", "SELECT", true},
			{rtr, "app.invitations", "token_digest", "SELECT", false}, {rtr, "app.invitations", "consumed_at", "UPDATE", false},
			{rtr, "app.invitations", "consumed_by_user_id", "UPDATE", false}, {rtr, "app.invitations", "intended_user_id", "UPDATE", false},
		} {
			if got := one(`SELECT has_column_privilege($1, $2, $3, $4)::text`, c.role, c.rel, c.col, c.priv); got != map[bool]string{true: "true", false: "false"}[c.want] {
				t.Errorf("%s %s on %s.%s: %s", c.role, c.priv, c.rel, c.col, got)
			}
		}
		// The table owner retains DELETE on its own table (an ACL fact); FORCE RLS with no DELETE policy is the separate
		// behavioral control, asserted with the policy strings. The runtime and claim-writer negatives stay.
		for _, c := range []struct {
			role, rel, priv string
			want            bool
		}{
			{rtr, "app.role_bindings", "DELETE", true}, {rtr, "app.invitations", "DELETE", false}, {o, "app.invitations", "DELETE", true},
			{cw, "app.role_bindings", "DELETE", false}, {cw, "audit.events", "SELECT", false}, {cw, "app.invitations", "DELETE", false},
		} {
			if got := one(`SELECT has_table_privilege($1, $2, $3)::text`, c.role, c.rel, c.priv); got != map[bool]string{true: "true", false: "false"}[c.want] {
				t.Errorf("%s %s on %s: %s", c.role, c.priv, c.rel, got)
			}
		}
		// is_organization_member: executable by the runtime, the owner and the claim writer; the Organization claim by the
		// owner only.
		for _, c := range []struct {
			role, fn string
			want     bool
		}{
			{rtr, "app.is_organization_member(uuid, uuid)", true}, {o, "app.is_organization_member(uuid, uuid)", true},
			{cw, "app.is_organization_member(uuid, uuid)", true},
			{rtr, "app.invitation_claim_organization(uuid, bytea, text, text, text)", false},
			{o, "app.invitation_claim_organization(uuid, bytea, text, text, text)", true},
			{pr, "app.invitation_claim_group(uuid, bytea)", false},
		} {
			if got := one(`SELECT has_function_privilege($1, $2, 'EXECUTE')::text`, c.role, c.fn); got != map[bool]string{true: "true", false: "false"}[c.want] {
				t.Errorf("%s EXECUTE %s: %s", c.role, c.fn, got)
			}
		}

		// Postcondition 3 repeated (name, command, roles; role_bindings_lock dropped), then the load-bearing predicates.
		if got := one(`SELECT string_agg(format('%s:%s:%s', p.polname, p.polcmd, (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C")
  FROM (SELECT `+labels("x")+` AS v FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname)
FROM pg_policy p WHERE p.polname IN ('invitations_select', 'invitations_insert', 'invitations_update', 'invitations_claim_select',
  'invitations_claim_update', 'role_bindings_update', 'role_bindings_delete', 'role_bindings_scope_reader',
  'role_bindings_retirement_reader', 'role_bindings_claim_select', 'role_bindings_insert_invited', 'role_bindings_claim_upgrade',
  'users_policy_reader', 'users_claim_select', 'users_claim_lock', 'identity_principals_claim_select', 'events_insert_claim',
  'role_bindings_lock')`, roles...); got != "events_insert_claim:a:cw;identity_principals_claim_select:r:cw;"+
			"invitations_claim_select:r:cw;invitations_claim_update:w:cw;invitations_insert:a:ow+rt;"+
			"invitations_select:r:ow+rt;invitations_update:w:ow+rt;role_bindings_claim_select:r:cw;"+
			"role_bindings_claim_upgrade:w:cw;role_bindings_delete:d:ow+rt;role_bindings_insert_invited:a:cw;"+
			"role_bindings_retirement_reader:r:pr;role_bindings_scope_reader:r:pr;role_bindings_update:w:ow+rt;"+
			"users_claim_lock:w:cw;users_claim_select:r:cw;users_policy_reader:r:pr" {
			t.Errorf("membership policies:\n got %s", got)
		}
		pol := func(name string) string {
			return one(`SELECT regexp_replace(coalesce(pg_get_expr(polqual, polrelid), ''), '\s+', ' ', 'g') || ' || ' ||
  regexp_replace(coalesce(pg_get_expr(polwithcheck, polrelid), ''), '\s+', ' ', 'g') FROM pg_policy WHERE polname = $1`, name)
		}
		for name, frags := range map[string][]string{
			"invitations_select":              {"ctx_org()", "intended_user_id", "consumed_by_user_id", "ctx_user()", "rls_org_role()", "rls_group_role(group_id)"},
			"invitations_insert":              {"ctx_org()", "issuer_user_id", "ctx_user()", "is_organization_member(intended_user_id", "ctx_group()"},
			"invitations_update":              {"ctx_org()", "rls_org_role()", "rls_group_role(group_id)"},
			"invitations_claim_update":        {"consumed_by_user_id = app.ctx_user()"},
			"role_bindings_update":            {"ctx_org()", "'membership'", "subject_user_id", "ctx_group()", "rls_org_role()", "rls_group_role(group_id)"},
			"role_bindings_delete":            {"ctx_org()", "'membership'", "subject_user_id", "ctx_group()"},
			"role_bindings_scope_reader":      {"organization_id = app.ctx_org()"},
			"role_bindings_retirement_reader": {"subject_user_id = app.ctx_user()", "'org_owner'", "'group_owner'"},
			"role_bindings_insert_invited":    {"subject_user_id = app.ctx_user()", "'membership'", "<> ALL", "'org_owner'", "'group_owner'"},
			"role_bindings_claim_upgrade":     {"subject_user_id = app.ctx_user()", "'membership'", "<> ALL", "'org_owner'", "'group_owner'"},
			"users_claim_select":              {"id = app.ctx_user()", "rls_org_role()", "rb.subject_user_id = users.id", "rb.organization_id = app.ctx_org()"},
			"users_claim_lock":                {"id = app.ctx_user()", "rls_org_role()", "|| false"},
			"events_insert_claim": {"pg_trigger_depth() >= 1", "actor_kind = 'user'", "actor_user_id = app.ctx_user()",
				"organization_id IS NOT NULL", "'membership.granted'", "'membership.role_changed'", "'invitation.consumed'"},
		} {
			got := pol(name)
			for _, f := range frags {
				if !strings.Contains(got, f) {
					t.Errorf("policy %s lacks %q: %s", name, f, got)
				}
			}
		}
		// The claim-scope audit policy has no context-Organization term (the owner-only consumer's context has none).
		if got := pol("events_insert_claim"); strings.Contains(got, "ctx_org()") {
			t.Errorf("events_insert_claim names the context Organization: %s", got)
		}
		if got := one(`SELECT count(*)::text FROM pg_policy WHERE polname = 'role_bindings_lock'`); got != "0" {
			t.Errorf("role_bindings_lock remains at version 15: %s", got)
		}
	})

	t.Run("CT-MEM-1 least privilege under actual grants", func(t *testing.T) {
		inTx := func(user, org string, f func(tx pgx.Tx)) {
			t.Helper()
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if user != "" {
				rlsContext(t, tx, user, org, "", "")
			}
			f(tx)
		}
		denied := func(what, sql string, args ...any) {
			t.Helper()
			inTx(memOwnerA, memOrgA, func(tx pgx.Tx) {
				if _, err := tx.Exec(ctx, sql, args...); sqlstateOf(err) != "42501" {
					t.Errorf("%s: %v, want 42501", what, err)
				}
			})
		}
		denied("runtime reads the token digest", `SELECT token_digest FROM app.invitations`)
		denied("runtime writes a consumption column", `UPDATE app.invitations SET consumed_at = now()`)
		denied("runtime moves the intended recipient", `UPDATE app.invitations SET intended_user_id = $1`, memMemberA)
		denied("runtime deletes an invitation", `DELETE FROM app.invitations`)
		denied("runtime executes the Organization claim", `SELECT * FROM app.invitation_claim_organization($1, $2, 'i', 's', 'a@example.test')`,
			memInviteB, make([]byte, 32))
		denied("runtime writes an owner binding at depth 0", `INSERT INTO app.role_bindings
  (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES ($1, $2, 'group', $3, $3, 'group_owner', 'membership')`,
			memOrgA, memMemberA, memGroupA)
		// A foreign Organization's invitation and bindings stay invisible to an owner of A (committed fixture, then removed).
		replica(`INSERT INTO app.invitations (id, organization_id, scope_type, scope_id, group_id, role, issuer_user_id, intended_address,
  token_digest, expires_at) VALUES ($1, $2, 'organization', $2, NULL, 'org_member', $3, 'invitee@example.test', $4, now() + interval '1 day')`,
			memInviteB, memOrgB, memOwnerB, bytes.Repeat([]byte{0xb1}, 32))
		inTx(memOwnerA, memOrgA, func(tx pgx.Tx) {
			var inv, rb int
			if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.invitations WHERE organization_id = $1),
  (SELECT count(*) FROM app.role_bindings WHERE organization_id = $1)`, memOrgB).Scan(&inv, &rb); err != nil || inv != 0 || rb != 0 {
				t.Errorf("foreign rows visible to an owner of A: invitations=%d bindings=%d %v", inv, rb, err)
			}
		})
		inTx(memOwnerB, memOrgB, func(tx pgx.Tx) {
			var n int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.invitations WHERE id = $1`, memInviteB).Scan(&n); err != nil || n != 1 {
				t.Errorf("the owner of B does not see its own invitation: %d %v", n, err)
			}
		})
		inTx("", "", func(tx pgx.Tx) {
			var n int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.invitations`).Scan(&n); err != nil || n != 0 {
				t.Errorf("invitations visible without context: %d %v", n, err)
			}
		})
		exec(`DELETE FROM app.invitations WHERE id = $1`, memInviteB)
	})

	t.Run("CT-MEM-AUD the widened audit constraints", func(t *testing.T) {
		def := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'audit.events'::regclass
  AND conname = 'events_operation_valid'`)
		ops := regexp.MustCompile(`'([a-z_.]+)'::text`).FindAllStringSubmatch(def, -1)
		if len(ops) != 53 {
			t.Errorf("events_operation_valid lists %d operations, want 53: %s", len(ops), def)
		}
		got := map[string]bool{}
		for _, m := range ops {
			got[m[1]] = true
		}
		for _, op := range append(slicesClone(memOps), onbOps...) {
			if !got[op] {
				t.Errorf("operation %s missing", op)
			}
		}
		acts := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'audit.events'::regclass AND conname = 'events_action_valid'`)
		for _, a := range []string{"'invitation.issued'", "'invitation.revoked'", "'invitation.consumed'", "'membership.granted'"} {
			if !strings.Contains(acts, a) {
				t.Errorf("action %s missing: %s", a, acts)
			}
		}
		if res := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'audit.events'::regclass
  AND conname = 'events_resource_type_valid'`); !strings.Contains(res, "'invitation'") {
			t.Errorf("resource type invitation missing: %s", res)
		}
	})

	t.Run("CT-MEM-UPDOWN clean down/up and the history-preserving down refusal", func(t *testing.T) {
		down := func() (string, error) {
			var b bytes.Buffer
			err := r.Down(ctx, &b)
			return b.String(), err
		}
		out, err := down()
		if err != nil || out != "rolled_back\t00015_membership_invitations.sql\ncurrent_version\t14\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		metadataThroughVersion14Only(t, cfg, r)
		// Policies (role_bindings_lock restored), grants (is_organization_member runtime-only again, the claim writer
		// holding nothing) and bodies (events_attribute = 00013) are exactly the version-14 catalog.
		if got := state(); got != s14 {
			t.Fatalf("catalog after down differs from version 14:\n got %s\nwant %s", got, s14)
		}
		if got := flags(); got != flagsOn {
			t.Errorf("RLS/FORCE after down: %s", got)
		}
		objRunUp(t, r, "applied\t00015_membership_invitations.sql\ncurrent_version\t15\n")
		if got := state(); got != s15 {
			t.Fatalf("catalog after down/up differs from version 15:\n got %s\nwant %s", got, s15)
		}
		// Refused while an invitation exists (any tenant), transactionally: catalog, row and FORCE unchanged.
		replica(`INSERT INTO app.invitations (id, organization_id, scope_type, scope_id, group_id, role, issuer_user_id, intended_address,
  token_digest, expires_at) VALUES ($1, $2, 'organization', $2, NULL, 'org_member', $3, 'kept@example.test', $4, now() + interval '1 day')`,
			memInviteUD, memOrgA, memOwnerA, bytes.Repeat([]byte{0xc2}, 32))
		check := func(wantMsg, countSQL string) {
			t.Helper()
			if out, err := down(); err == nil || !strings.Contains(err.Error(), wantMsg) {
				t.Fatalf("down with history: %v %q, want %q", err, out, wantMsg)
			}
			if got := state(); got != s15 {
				t.Fatal("the refused down changed the catalog")
			}
			var n int
			if err := setup.QueryRow(ctx, countSQL).Scan(&n); err != nil || n != 1 {
				t.Fatalf("history after the refused down: %d %v", n, err)
			}
			if got := one(`SELECT string_agg(relname || ':' || relforcerowsecurity::text, ',' ORDER BY relname) FROM pg_class
WHERE oid IN ('app.invitations'::regclass, 'audit.events'::regclass)`); got != "events:true,invitations:true" {
				t.Fatalf("FORCE after the refused down: %s", got)
			}
			if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t15\n" {
				t.Fatalf("version after the refused down: %q", out)
			}
		}
		check("refusing to roll back 00015: 1 invitations / 0 audit rows (history is never discarded)",
			`SELECT count(*) FROM app.invitations WHERE id = '`+memInviteUD+`'`)
		exec(`DELETE FROM app.invitations WHERE id = $1`, memInviteUD)
		// Refused while an invitation-era audit row exists.
		replica(`INSERT INTO audit.events (action, outcome, actor_kind, operation) VALUES ('access.denied', 'denied', 'system', 'invitation.list')`)
		check("refusing to roll back 00015: 0 invitations / 1 audit rows (history is never discarded)",
			`SELECT count(*) FROM audit.events WHERE operation = 'invitation.list'`)
		reset(t, cfg)
	})
}

// sortedJoin returns the strings joined by "," in byte order (the COLLATE "C" order of the catalog queries).
func sortedJoin(s []string) string {
	c := slices.Clone(s)
	slices.Sort(c)
	return strings.Join(c, ",")
}
