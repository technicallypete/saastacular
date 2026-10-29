//go:build integration

// G014 B-MCP-CATALOG (Plan r8 §5, §10.4) against the disposable PostgreSQL of the migrations suite. TestMCPCatalog pins
// the exact version-10 delta on top of the unchanged version-9 catalog (CT-MCP-1) and ten rolled-back missing-protection
// detectors (CT-MCP-NEG). TestMCPMigration covers empty to 10, a populated version-9 world to 10 with every row
// identical, down refused with MCP state, down to 9 when empty with the exact version-9 catalog, up again and reapply,
// and a rolled-back up (MG-MCP-1..5). The expected function bodies are read from the embedded migration itself, so the
// applied catalog must equal the reviewed file. Setup and rolled-back controls use the attested superuser; catalog
// reads use the owner. Since G016 these version-10 cases run the normal set up to version 10 (runnerFor with subsetFS
// 10), so every assertion keeps its exact version-10 meaning; the full set (version 11) is covered by
// TestMigrationLifecycle and TestDomainsCatalog.
package migrate

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/migrations"
)

// mcpBody returns the dollar-quoted body (prosrc) of the first function created after marker in the 00010 up section.
func mcpBody(t *testing.T, marker string) string {
	t.Helper()
	b, err := fs.ReadFile(migrations.FS, "00010_mcp_connections.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, _ := strings.Cut(string(b), "-- +goose Down")
	i := strings.Index(up, marker)
	if i < 0 {
		t.Fatalf("00010: %q not found", marker)
	}
	j := strings.Index(up[i:], "AS $$")
	if j < 0 {
		t.Fatalf("00010: no body after %q", marker)
	}
	start := i + j + len("AS $$")
	k := strings.Index(up[start:], "$$")
	return up[start : start+k]
}

func md5hex(s string) string { return fmt.Sprintf("%x", md5.Sum([]byte(s))) }

// The version-10 delta.
var (
	mcpWantFlags    = strings.Replace(objWantFlags, "login_transactions:t:t,", "login_transactions:t:t,mcp_connections:t:t,", 1)
	mcpWantPolicies = strings.Replace(retWantAppPolicies, "login_transactions_select:login_transactions:r:true:ow+rt;",
		"login_transactions_select:login_transactions:r:true:ow+rt;mcp_connections_insert:mcp_connections:a:true:ow+rt;"+
			"mcp_connections_select:mcp_connections:r:true:ow+rt;mcp_connections_update:mcp_connections:w:true:ow+rt;", 1)
	mcpWantACL = strings.Replace(strings.Replace(retWantACL,
		"login_transactions:rt:DELETE,INSERT,SELECT;",
		"login_transactions:rt:DELETE,INSERT,SELECT;mcp_connections:rt:SELECT;mcp_connections.activated_at:rt:UPDATE;"+
			"mcp_connections.client_id:rt:UPDATE;mcp_connections.context_version:rt:UPDATE;mcp_connections.created_at:rt:INSERT;"+
			"mcp_connections.host:rt:INSERT;mcp_connections.id:rt:INSERT;mcp_connections.organization_id:rt:INSERT;"+
			"mcp_connections.pending_expires_at:rt:INSERT;mcp_connections.resource:rt:INSERT;mcp_connections.selected_group_id:rt:UPDATE;"+
			"mcp_connections.session_id:rt:INSERT;mcp_connections.state:rt:UPDATE;mcp_connections.user_id:rt:INSERT;", 1),
		"sessions.idle_expires_at:rt:UPDATE;", "sessions.idle_expires_at:rt:UPDATE;sessions.kind:ar:SELECT;", 1)
	mcpWantFunctions = strings.Replace(retWantFunctions, "groups_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;",
		"groups_keys_immutable:ow:false:false:v:search_path=pg_catalog:-;mcp_connections_guard:ow:false:false:v:search_path=pg_catalog, pg_temp:-;", 1)
	mcpWantAuditFunctions = strings.Replace(auditWantFunctions, "capture_groups:ow:false:false:v:search_path=pg_catalog, pg_temp:-;",
		"capture_groups:ow:false:false:v:search_path=pg_catalog, pg_temp:-;capture_mcp_connections:ow:false:false:v:search_path=pg_catalog, pg_temp:-;", 1)
	mcpWantColumns = "id:uuid:t:;session_id:uuid:t:;user_id:uuid:t:;organization_id:uuid:t:;host:text:t:;resource:text:t:;" +
		"state:text:t:'pending'::text;pending_expires_at:timestamp with time zone:t:;client_id:text:f:;" +
		"activated_at:timestamp with time zone:f:;selected_group_id:uuid:f:;context_version:bigint:t:0;" +
		"created_at:timestamp with time zone:t:now()"
	mcpWantConstraints = "mcp_connections_active_shape:c;mcp_connections_client_valid:c;mcp_connections_host_valid:c;" +
		"mcp_connections_id_nonzero:c;mcp_connections_organization_fkey:f;mcp_connections_pkey:p;mcp_connections_resource_key:u;" +
		"mcp_connections_resource_valid:c;mcp_connections_selection_shape:c;mcp_connections_session_fkey:f;" +
		"mcp_connections_session_key:u;mcp_connections_state_valid:c;mcp_connections_user_fkey:f;mcp_connections_version_valid:c"
	mcpWantIndexes  = "mcp_connections_pkey:id;mcp_connections_resource_key:resource;mcp_connections_session_key:session_id"
	mcpWantTriggers = "mcp_connections.audit_capture_selection:O:CREATE TRIGGER audit_capture_selection AFTER UPDATE OF selected_group_id " +
		"ON app.mcp_connections FOR EACH ROW WHEN ((old.selected_group_id IS DISTINCT FROM new.selected_group_id)) EXECUTE FUNCTION " +
		"audit.capture_mcp_connections()\nmcp_connections.audit_capture_state:O:CREATE TRIGGER audit_capture_state AFTER UPDATE OF state " +
		"ON app.mcp_connections FOR EACH ROW WHEN ((old.state IS DISTINCT FROM new.state)) EXECUTE FUNCTION audit.capture_mcp_connections()\n" +
		"mcp_connections.mcp_connections_guard:O:CREATE TRIGGER mcp_connections_guard BEFORE UPDATE ON app.mcp_connections FOR EACH ROW " +
		"EXECUTE FUNCTION app.mcp_connections_guard()"
	mcpWantSessionsKind = "kind:text:t:'browser'::text;sessions_kind_valid:CHECK ((kind = ANY (ARRAY['browser'::text, 'mcp'::text])))"
	mcpWantExprs        = map[string]string{
		"mcp_connections_select": "((user_id = ( SELECT app.ctx_user() AS ctx_user)) AND (organization_id = ( SELECT app.ctx_org() AS ctx_org))) WITH CHECK",
		"mcp_connections_insert": "WITH CHECK ((user_id = ( SELECT app.ctx_user() AS ctx_user)) AND (organization_id = ( SELECT app.ctx_org() AS ctx_org)) " +
			"AND (( SELECT app.ctx_group() AS ctx_group) IS NULL) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) " +
			"AND (state = 'pending'::text) AND (selected_group_id IS NULL) AND (context_version = 0))",
		"mcp_connections_update": "((user_id = ( SELECT app.ctx_user() AS ctx_user)) AND (organization_id = ( SELECT app.ctx_org() AS ctx_org)) " +
			"AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL)) WITH CHECK ((user_id = ( SELECT app.ctx_user() AS ctx_user)) " +
			"AND (organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL))",
		"sessions_insert": "WITH CHECK ((user_id = ( SELECT app.ctx_user() AS ctx_user)) AND (revoked_at IS NULL) AND (((kind = 'browser'::text) " +
			"AND (( SELECT app.ctx_org() AS ctx_org) IS NULL)) OR ((kind = 'mcp'::text) AND (( SELECT app.ctx_org() AS ctx_org) IS NOT NULL) " +
			"AND (( SELECT app.ctx_group() AS ctx_group) IS NULL) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL))))",
	}
	// The new catalog values each widened audit constraint must name.
	mcpWantAuditValues = map[string][]string{
		"events_action_valid":        {"'mcp.connection_established'", "'mcp.context_changed'"},
		"events_reason_valid":        {"'resource_mismatch'"},
		"events_operation_valid":     {"'mcp.request'"},
		"events_resource_type_valid": {"'mcp_connection'"},
	}
)

// mcpAuditSrc is the version-9 audit source pin with the replaced attribution body and the new capture function.
func mcpAuditSrc(t *testing.T) string {
	t.Helper()
	m := map[string]string{}
	for _, kv := range strings.Split(retWantAuditSrc, ";") {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	m["events_attribute"] = md5hex(mcpBody(t, "CREATE OR REPLACE FUNCTION audit.events_attribute()"))
	m["capture_mcp_connections"] = md5hex(mcpBody(t, "CREATE FUNCTION audit.capture_mcp_connections()"))
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	var out []string
	for _, k := range keys {
		out = append(out, k+"="+m[k])
	}
	return strings.Join(out, ";")
}

// mcpParts reads every version-9 part plus the version-10 extras.
func mcpParts(t *testing.T, q catalogQuerier, cfg Config) map[string]string {
	t.Helper()
	p := retentionParts(t, q, cfg)
	one := func(sql string) string {
		var s *string
		if err := q.QueryRow(context.Background(), sql).Scan(&s); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		if s == nil {
			return ""
		}
		return *s
	}
	p["mcp-columns"] = one(`SELECT string_agg(format('%s:%s:%s:%s', a.attname, format_type(a.atttypid, a.atttypmod),
  CASE WHEN a.attnotnull THEN 't' ELSE 'f' END, coalesce(pg_get_expr(d.adbin, d.adrelid), '')), ';' ORDER BY a.attnum)
FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = to_regclass('app.mcp_connections') AND a.attnum > 0 AND NOT a.attisdropped`)
	p["mcp-constraints"] = one(`SELECT string_agg(conname || ':' || contype::text, ';' ORDER BY conname) FROM pg_constraint
WHERE conrelid = to_regclass('app.mcp_connections')`)
	p["mcp-indexes"] = one(`SELECT string_agg(c.relname || ':' || (SELECT string_agg(a.attname, ', ' ORDER BY k.ord)
  FROM unnest(i.indkey) WITH ORDINALITY k(attnum, ord) JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = k.attnum),
  ';' ORDER BY c.relname)
FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE i.indrelid = to_regclass('app.mcp_connections')`)
	p["mcp-triggers"] = one(`SELECT string_agg(c.relname || '.' || t.tgname || ':' || t.tgenabled::text || ':' || pg_get_triggerdef(t.oid),
  E'\n' ORDER BY c.relname, t.tgname)
FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid WHERE c.oid = to_regclass('app.mcp_connections') AND NOT t.tgisinternal`)
	p["sessions-kind"] = one(`SELECT (SELECT format('%s:%s:%s:%s', a.attname, format_type(a.atttypid, a.atttypmod),
  CASE WHEN a.attnotnull THEN 't' ELSE 'f' END, coalesce(pg_get_expr(d.adbin, d.adrelid), ''))
  FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
  WHERE a.attrelid = 'app.sessions'::regclass AND a.attname = 'kind' AND NOT a.attisdropped) || ';' ||
  (SELECT conname || ':' || pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'app.sessions'::regclass AND conname = 'sessions_kind_valid')`)
	p["mcp-guard-src"] = one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure('app.mcp_connections_guard()')`)
	p["e1-src"] = one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure('app.auth_resolve_session(bytea, timestamptz)')`)
	// The app capture triggers other than the connection table's: unchanged from version 9.
	var lines []string
	for _, l := range strings.Split(p["audit-capture-text"], "\n") {
		if !strings.HasPrefix(l, "mcp_connections.") {
			lines = append(lines, l)
		}
	}
	p["audit-capture-rest"] = md5hex(strings.Join(lines, "\n"))
	for name, values := range mcpWantAuditValues {
		def := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = to_regclass('audit.events') AND conname = '` + name + `'`)
		ok := def != ""
		for _, v := range values {
			ok = ok && strings.Contains(def, v)
		}
		p["audit-check:"+name] = fmt.Sprint(ok)
	}
	return p
}

// mcpViolations is the version-10 detector: the sorted codes of every part that differs from its pin.
func mcpViolations(t *testing.T, q catalogQuerier, cfg Config) []string {
	t.Helper()
	p := mcpParts(t, q, cfg)
	want := map[string]string{"rls-flags": mcpWantFlags, "policy-inventory": mcpWantPolicies, "acl": mcpWantACL,
		"functions": mcpWantFunctions, "schema": rlsWantSchema, "roles": rlsWantRoles, "memberships": rlsWantMemberships,
		"views": rlsWantViews, "obj-columns": objWantColumns, "obj-constraints": objWantConstraints, "obj-indexes": objWantIndexes,
		"obj-guard-src": objWantGuardSrc, "audit-relation": auditWantRelations, "audit-columns": auditWantColumns,
		"audit-constraints": auditWantConstraints, "audit-indexes": auditWantIndexes, "audit-policies": retWantAuditPolicies,
		"audit-acl": retWantAuditACL, "audit-schema": retWantAuditSchema, "audit-functions": mcpWantAuditFunctions,
		"audit-triggers": auditWantTriggers, "audit-capture-rest": auditWantCapture, "audit-src": mcpAuditSrc(t),
		"ret-eligible-src": retWantEligibleSrc,
		"mcp-columns":      mcpWantColumns, "mcp-constraints": mcpWantConstraints, "mcp-indexes": mcpWantIndexes,
		"mcp-triggers": mcpWantTriggers, "sessions-kind": mcpWantSessionsKind,
		"mcp-guard-src": md5hex(mcpBody(t, "CREATE FUNCTION app.mcp_connections_guard()")),
		"e1-src":        md5hex(mcpBody(t, "CREATE OR REPLACE FUNCTION app.auth_resolve_session")),
		"policy-expr:stored_objects_retention_select": retWantObjectsExpr, "policy-expr:stored_objects_retention_delete": retWantObjectsExpr,
		"audit-policy-expr:events_select": auditWantSelectExpr, "audit-policy-expr:events_insert": auditWantInsertExpr,
		"audit-policy-expr:events_retention_select": retWantEventsExpr, "audit-policy-expr:events_retention_delete": retWantEventsExpr}
	for name, expr := range rlsWantExprs {
		want["policy-expr:"+name] = expr
	}
	for name, expr := range mcpWantExprs {
		want["policy-expr:"+name] = expr
	}
	for name := range mcpWantAuditValues {
		want["audit-check:"+name] = "true"
	}
	var codes []string
	for k, w := range want {
		if p[k] != w {
			codes = append(codes, k)
			t.Logf("detector %s: got %s", k, p[k])
		}
	}
	slices.Sort(codes)
	return slices.Compact(codes)
}

func mcpUp(t *testing.T, cfg Config) *Runner {
	t.Helper()
	r := runnerFor(t, cfg, subsetFS(t, 10))
	objRunUp(t, r, "current_version\t10\n")
	return r
}

// TestMCPCatalog (B-MCP-CATALOG): CT-MCP-1 and CT-MCP-NEG.
func TestMCPCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r := mcpUp(t, cfg)
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)

	t.Run("CT-MCP-1 the exact version-10 delta", func(t *testing.T) {
		if v := mcpViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-10 detector: %v", v)
		}
		metadataThroughVersion10Only(t, cfg, r)
		// Everything outside the delta equals version 9: the version-9 detector reports exactly the delta's parts.
		got := retentionViolations(t, cat, cfg)
		want := []string{"acl", "audit-capture", "audit-functions", "audit-src", "functions", "policy-expr:sessions_insert", "policy-inventory",
			"rls-flags"}
		if !slices.Equal(got, want) {
			t.Errorf("version-9 detector at version 10: %v, want exactly the delta %v", got, want)
		}
	})

	t.Run("CT-MCP-NEG rolled-back missing-protection detectors", func(t *testing.T) {
		rt, ar := quote(cfg.RuntimeRole()), quote(cfg.Database+"_auth_reader")
		for _, n := range []struct {
			id, sql string
			want    []string
		}{
			{"N1 runtime DELETE on connections", `GRANT DELETE ON app.mcp_connections TO ` + rt, []string{"acl"}},
			{"N2 runtime UPDATE of a binding column", `GRANT UPDATE (host, resource) ON app.mcp_connections TO ` + rt, []string{"acl"}},
			{"N3 insert policy without the Organization role", `ALTER POLICY mcp_connections_insert ON app.mcp_connections WITH CHECK (` +
				`user_id = (SELECT app.ctx_user()) AND organization_id = (SELECT app.ctx_org()) AND state = 'pending')`,
				[]string{"policy-expr:mcp_connections_insert"}},
			{"N4 update policy without the Organization role", `ALTER POLICY mcp_connections_update ON app.mcp_connections USING (` +
				`user_id = (SELECT app.ctx_user()) AND organization_id = (SELECT app.ctx_org()))`, []string{"policy-expr:mcp_connections_update"}},
			{"N5 select policy without the Organization predicate", `ALTER POLICY mcp_connections_select ON app.mcp_connections USING (` +
				`user_id = (SELECT app.ctx_user()))`, []string{"policy-expr:mcp_connections_select"}},
			{"N6 sessions_insert admitting browser rows in a tenant shape", `ALTER POLICY sessions_insert ON app.sessions WITH CHECK (` +
				`user_id = (SELECT app.ctx_user()) AND revoked_at IS NULL)`, []string{"policy-expr:sessions_insert"}},
			{"N7 E1 without the kind filter", `GRANT CREATE ON SCHEMA app TO ` + ar + `; SET LOCAL ROLE ` + ar + `; ` +
				strings.Replace(strings.Replace(mcpE1(t), " AND s.kind = 'browser'", "", 1), "$$", "$f$", 2) + `; RESET ROLE; ` +
				`REVOKE CREATE ON SCHEMA app FROM ` + ar, []string{"e1-src"}},
			{"N8 missing guard trigger", `DROP TRIGGER mcp_connections_guard ON app.mcp_connections`, []string{"mcp-triggers"}},
			{"N9 activation allowed twice", `CREATE OR REPLACE FUNCTION app.mcp_connections_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $f$ BEGIN RETURN NEW; END $f$`, []string{"mcp-guard-src"}},
			{"N10 definer reader on the table", `CREATE FUNCTION app.mcp_probe_reader() RETURNS SETOF uuid LANGUAGE sql STABLE SECURITY DEFINER ` +
				`SET search_path = pg_catalog, pg_temp AS $f$ SELECT id FROM app.mcp_connections $f$`, []string{"functions"}},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			v := mcpViolations(t, tx, cfg)
			_ = tx.Rollback(ctx)
			if !slices.Equal(v, n.want) {
				t.Errorf("%s: the detector reported %v, want %v", n.id, v, n.want)
			}
		}
		if v := mcpViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", v)
		}
	})

	t.Run("CT-MCP-GUARD the selection guard under the runtime role", func(t *testing.T) {
		mcpGuardCases(t, ctx, setup, cfg)
	})
}

// mcpGuardCases runs the selection guard's refusals as the runtime role (SET LOCAL ROLE inside rolled-back superuser
// transactions, so FORCE row-level security applies): a foreign target, a target the subject cannot see, a previous
// selection the subject can no longer see (a clear still passes), a write in a selected-Group shape, and the negative
// control: a guard without the previous-selection clause lets that switch through, so its omission is detectable.
func mcpGuardCases(t *testing.T, ctx context.Context, setup *pgx.Conn, cfg Config) {
	t.Helper()
	rt := quote(cfg.RuntimeRole())
	var user, o1, o2, g1, g2, g3, gx, conn string
	if err := setup.QueryRow(ctx, `WITH u AS (INSERT INTO app.users (display_name) VALUES ('Guard probe') RETURNING id),
  a AS (INSERT INTO app.organizations (name) VALUES ('Guard A') RETURNING id),
  b AS (INSERT INTO app.organizations (name) VALUES ('Guard B') RETURNING id)
SELECT u.id::text, a.id::text, b.id::text FROM u, a, b`).Scan(&user, &o1, &o2); err != nil {
		t.Fatal(err)
	}
	group := func(org, slug string) string {
		var id string
		if err := setup.QueryRow(ctx, `INSERT INTO app.groups (organization_id, name, slug) VALUES ($1::text::uuid, $2, $2) RETURNING id::text`,
			org, slug+"-"+user[:8]).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	g1, g2, g3, gx = group(o1, "guard-one"), group(o1, "guard-two"), group(o1, "guard-three"), group(o2, "guard-foreign")
	if _, err := setup.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership'),
  ($1::text::uuid, $2::text::uuid, 'group', $3::text::uuid, $3::text::uuid, 'group_member', 'membership'),
  ($1::text::uuid, $2::text::uuid, 'group', $4::text::uuid, $4::text::uuid, 'group_member', 'membership')`, o1, user, g1, g2); err != nil {
		t.Fatal(err)
	}
	var session string
	if err := setup.QueryRow(ctx, `INSERT INTO app.sessions (user_id, kind, credential_hash, csrf_token, auth_time, idle_expires_at,
  absolute_expires_at) VALUES ($1::text::uuid, 'mcp', sha256(gen_random_uuid()::text::bytea), sha256(gen_random_uuid()::text::bytea), now(),
  now() + interval '1 day', now() + interval '2 days') RETURNING id::text`, user).Scan(&session); err != nil {
		t.Fatal(err)
	}
	if err := setup.QueryRow(ctx, `INSERT INTO app.mcp_connections (id, session_id, user_id, organization_id, host, resource, pending_expires_at)
  VALUES (gen_random_uuid(), $1::text::uuid, $2::text::uuid, $3::text::uuid, 'org-a.localhost:543', 'https://org-a.localhost:543/mcp/c/guard',
  now() + interval '30 minutes') RETURNING id::text`, session, user, o1).Scan(&conn); err != nil {
		t.Fatal(err)
	}
	if _, err := setup.Exec(ctx, `UPDATE app.mcp_connections SET state = 'active', client_id = 'client_01GUARD', activated_at = now() WHERE id = $1::text::uuid`, conn); err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = setup.Exec(context.Background(), `DELETE FROM app.mcp_connections WHERE id = $1::text::uuid; DELETE FROM app.sessions WHERE id = $2::text::uuid`, conn, session)
	})
	// try switches the connection to group ("" clears) as the runtime role in the given context Group ("" for the
	// Organization-only shape), after the optional setup statement, and rolls back; it returns the rows affected (so a
	// row hidden by row-level security never passes vacuously) and the error.
	try := func(ctxGroup, group, prelude string) (int64, error) {
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if prelude != "" {
			if _, err := tx.Exec(ctx, prelude); err != nil {
				t.Fatalf("prelude: %v", err)
			}
		}
		if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+rt); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true),
  set_config('app.group_id', $3, true), set_config('app.login_binding', '', true)`, user, o1, ctxGroup); err != nil {
			t.Fatal(err)
		}
		var arg *string
		if group != "" {
			arg = &group
		}
		tag, err := tx.Exec(ctx, `UPDATE app.mcp_connections SET selected_group_id = $2::text::uuid, context_version = context_version + 1
  WHERE id = $1::text::uuid`, conn, arg)
		return tag.RowsAffected(), err
	}
	refusedByGuard := func(err error) bool {
		var pe *pgconn.PgError
		return errors.As(err, &pe) && pe.ConstraintName == "mcp_connections_guard"
	}
	if n, err := try("", g1, ""); err != nil || n != 1 {
		t.Fatalf("a visible target in the Organization-only shape: %d rows, %v", n, err)
	}
	for name, c := range map[string]struct{ ctxGroup, group string }{
		"foreign target":       {"", gx},
		"invisible target":     {"", g3},
		"selected-Group shape": {g1, g1},
	} {
		if _, err := try(c.ctxGroup, c.group, ""); !refusedByGuard(err) {
			t.Errorf("%s: %v, want the guard's refusal", name, err)
		}
	}
	// A previous selection the subject can no longer see: commit g1, then remove its binding.
	if _, err := setup.Exec(ctx, `UPDATE app.mcp_connections SET selected_group_id = $2::text::uuid, context_version = context_version + 1
  WHERE id = $1::text::uuid`, conn, g1); err != nil {
		t.Fatalf("select g1: %v", err)
	}
	if _, err := setup.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND group_id = $2::text::uuid`, user, g1); err != nil {
		t.Fatal(err)
	}
	if _, err := try("", g2, ""); !refusedByGuard(err) {
		t.Errorf("switch away from an invisible previous selection: %v, want the guard's refusal", err)
	}
	if n, err := try("", "", ""); err != nil || n != 1 {
		t.Errorf("clear after the loss: %d rows, %v", n, err)
	}
	// Negative control: the guard without the previous-selection clause admits that switch.
	src := mcpBody(t, "CREATE FUNCTION app.mcp_connections_guard()")
	clause := `
                     AND (OLD.selected_group_id IS NULL
                          OR EXISTS (SELECT 1 FROM app.groups y
                                     WHERE y.id = OLD.selected_group_id AND y.organization_id = NEW.organization_id))`
	if !strings.Contains(src, clause) {
		t.Fatal("the guard source does not hold the previous-selection clause")
	}
	weak := `CREATE OR REPLACE FUNCTION app.mcp_connections_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $f$` + strings.Replace(src, clause, "", 1) + `$f$`
	if n, err := try("", g2, weak); err != nil || n != 1 {
		t.Errorf("negative control: the weakened guard still refused (%d rows, %v), so the clause is not what refuses it", n, err)
	}
}

// mcpE1 is the version-10 E1 statement as 00010 executes it.

// mcpE1 is the version-10 E1 statement as 00010 executes it.
func mcpE1(t *testing.T) string {
	t.Helper()
	b, err := fs.ReadFile(migrations.FS, "00010_mcp_connections.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, _ := strings.Cut(string(b), "-- +goose Down")
	i := strings.Index(up, "CREATE OR REPLACE FUNCTION app.auth_resolve_session")
	j := strings.Index(up[i:], "\n$e1$")
	return up[i : i+j]
}

// mcpRows hashes every app and audit row as JSON without the sessions kind column, so a version-9 world and the same
// world at version 10 (where every existing session is a browser session) compare equal.
func mcpRows(t *testing.T, setup *pgx.Conn) string {
	t.Helper()
	var parts []string
	for _, tbl := range []string{"app.identity_provider_binding", "app.users", "app.identity_principals", "app.sessions",
		"app.login_transactions", "app.organizations", "app.groups", "app.role_bindings", "app.business_profiles",
		"app.group_business_profiles", "app.stored_objects", "audit.events"} {
		var h string
		if err := setup.QueryRow(context.Background(), `SELECT count(*) || ':' || coalesce(md5(string_agg((to_jsonb(r) - 'kind')::text, E'\n'
  ORDER BY (to_jsonb(r) - 'kind')::text)), '-') FROM `+tbl+` r`).Scan(&h); err != nil {
			t.Fatal(err)
		}
		parts = append(parts, tbl+"="+h)
	}
	return strings.Join(parts, ";")
}

// TestMCPMigration (B-MCP-CATALOG): MG-MCP-1..5.
func TestMCPMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	full := "applied\t00001_baseline.sql\napplied\t00002_identity.sql\napplied\t00003_tenancy.sql\napplied\t00004_business_profiles.sql\n" +
		"applied\t00005_organization_policy.sql\napplied\t00006_row_level_security.sql\napplied\t00007_stored_objects.sql\n" +
		"applied\t00008_audit_events.sql\napplied\t00009_retention_maintenance.sql\napplied\t00010_mcp_connections.sql\ncurrent_version\t10\n"
	up := func(r *Runner) string {
		return run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) })
	}
	down := func(r *Runner) (string, error) {
		var b bytes.Buffer
		err := r.Down(ctx, &b)
		return b.String(), err
	}

	t.Run("MG-MCP-1 empty to version 10", func(t *testing.T) {
		reset(t, cfg)
		if out := up(runnerFor(t, cfg, subsetFS(t, 10))); out != full {
			t.Fatalf("up: %q", out)
		}
		if v := mcpViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
	})

	var rows string
	t.Run("MG-MCP-2 a populated version-9 world to 10", func(t *testing.T) {
		reset(t, cfg)
		objRunUp(t, runnerFor(t, cfg, subsetFS(t, 9)), "current_version\t9\n")
		retSeed(t, setup)
		if v := retentionViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-9 detector before the upgrade: %v", v)
		}
		rows = mcpRows(t, setup)
		r := runnerFor(t, cfg, subsetFS(t, 10))
		if out := up(r); out != "applied\t00010_mcp_connections.sql\ncurrent_version\t10\n" {
			t.Fatalf("up: %q", out)
		}
		if v := mcpViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
		if got := mcpRows(t, setup); got != rows {
			t.Fatal("rows changed by the upgrade")
		}
		var browser, mcp int
		if err := setup.QueryRow(ctx, `SELECT count(*) FILTER (WHERE kind = 'browser'), count(*) FILTER (WHERE kind = 'mcp') FROM app.sessions`).
			Scan(&browser, &mcp); err != nil || browser == 0 || mcp != 0 {
			t.Fatalf("existing sessions must become browser sessions: %d browser, %d mcp (%v)", browser, mcp, err)
		}
	})

	t.Run("MG-MCP-3 down refused with MCP state, allowed when empty, exact version-9 catalog", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 10))
		var user, org string
		if err := setup.QueryRow(ctx, `SELECT u.id::text, rb.organization_id::text FROM app.users u
  JOIN app.role_bindings rb ON rb.subject_user_id = u.id AND rb.scope_type = 'organization' ORDER BY u.id LIMIT 1`).Scan(&user, &org); err != nil {
			t.Fatal(err)
		}
		var session string
		if err := setup.QueryRow(ctx, `INSERT INTO app.sessions (user_id, kind, credential_hash, csrf_token, auth_time, idle_expires_at,
  absolute_expires_at) VALUES ($1::text::uuid, 'mcp', sha256(gen_random_uuid()::text::bytea), sha256(gen_random_uuid()::text::bytea), now(),
  now() + interval '1 day', now() + interval '2 days') RETURNING id::text`, user).Scan(&session); err != nil {
			t.Fatal(err)
		}
		if _, err := setup.Exec(ctx, `INSERT INTO app.mcp_connections (id, session_id, user_id, organization_id, host, resource, pending_expires_at)
  VALUES (gen_random_uuid(), $1::text::uuid, $2::text::uuid, $3::text::uuid, 'org-a.localhost:543', 'https://org-a.localhost:543/mcp/c/probe',
  now() + interval '30 minutes')`, session, user, org); err != nil {
			t.Fatal(err)
		}
		if out, err := down(r); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00010") {
			t.Fatalf("down with MCP state: %v %q", err, out)
		}
		if v := mcpViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("the refused down changed the catalog: %v", v)
		}
		if _, err := setup.Exec(ctx, `DELETE FROM app.mcp_connections; DELETE FROM app.sessions WHERE kind = 'mcp'`); err != nil {
			t.Fatal(err)
		}
		rows = mcpRows(t, setup) // the probe session's session.issued event stays in the history
		out, err := down(r)
		if err != nil || out != "rolled_back\t00010_mcp_connections.sql\ncurrent_version\t9\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		if v := retentionViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-9 detector after down: %v", v)
		}
		metadataThroughVersion9Only(t, cfg, r)
		if got := mcpRows(t, setup); got != rows {
			t.Fatal("rows changed by the rollback")
		}
	})

	t.Run("MG-MCP-4 up again and reapply", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 10))
		if out := up(r); out != "applied\t00010_mcp_connections.sql\ncurrent_version\t10\n" {
			t.Fatalf("up: %q", out)
		}
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out != "rolled_back\t00010_mcp_connections.sql\napplied\t00010_mcp_connections.sql\ncurrent_version\t10\n" {
			t.Fatalf("reapply: %q", out)
		}
		if v := mcpViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("detector %v", v)
		}
		if got := mcpRows(t, setup); got != rows {
			t.Fatal("rows changed by up/reapply")
		}
	})

	t.Run("MG-MCP-5 a rolled-back up leaves version 9", func(t *testing.T) {
		r := runnerFor(t, cfg, subsetFS(t, 10))
		if out, err := down(r); err != nil || !strings.HasSuffix(out, "current_version\t9\n") {
			t.Fatalf("down: %v %q", err, out)
		}
		// A stray runtime grant breaks the version-9 precondition: the up fails and rolls back entirely.
		if _, err := setup.Exec(ctx, `GRANT DELETE ON app.sessions TO `+quote(cfg.RuntimeRole())); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := r.Up(ctx, &buf); err == nil || !strings.Contains(err.Error(), "version-9 precondition") {
			t.Fatalf("up with a stray grant: %v %q", err, buf.String())
		}
		if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t9\n" {
			t.Fatalf("version after the failed up: %q", out)
		}
		if v := retentionViolations(t, cat, cfg); !slices.Equal(v, []string{"acl"}) {
			t.Errorf("after the failed up the version-9 detector reports %v, want only the stray grant", v)
		}
		if _, err := setup.Exec(ctx, `REVOKE DELETE ON app.sessions FROM `+quote(cfg.RuntimeRole())); err != nil {
			t.Fatal(err)
		}
		if v := retentionViolations(t, cat, cfg); len(v) != 0 {
			t.Errorf("version-9 detector: %v", v)
		}
		rlsOwnerRestored(t, cfg, r)
		if got := mcpRows(t, setup); got != rows {
			t.Fatal("rows changed")
		}
		objRunUp(t, r, "current_version\t10\n")
	})
}

// --- G063 B-MCP-CATALOG: TestMCPNamesCatalog (CT-MCPN-*) and TestMCPNamesMigration (MG-MCPN-*) -----------------------
//
// G063 Plan r1 §3, §6: the exact version-19 delta over the version-18 catalog. namesCatalog keys every app and audit
// relation, column, column ACL, function, policy, RLS flag, schema ACL, constraint and non-internal trigger; the delta
// between versions 18 and 19 is exactly the name column (with its CHECK and the two runtime column privileges), the
// name capture trigger, the two replaced audit bodies and the widened action CHECK. Policies, the guard and every other
// object are unchanged. CT-MCPN-NEG rolls back six missing-protection detectors and requires each to change exactly its
// key. The expected bodies are read from the embedded 00019 itself, so the applied catalog must equal the reviewed file.

const (
	mcpnUser    = "0f63c000-0000-4000-8000-0000000000a1" // org_owner of mcpnOrg
	mcpnOther   = "0f63c000-0000-4000-8000-0000000000b1" // a user without any binding in mcpnOrg
	mcpnOrg     = "0f63c000-0000-4000-8000-000000000001"
	mcpnSessP   = "0f63c000-0000-4000-8000-000000000501" // the pending connection's 'mcp' session
	mcpnSessA   = "0f63c000-0000-4000-8000-000000000502" // the active connection's 'mcp' session
	mcpnPending = "0f63c000-0000-4000-8000-000000000c01"
	mcpnActive  = "0f63c000-0000-4000-8000-000000000c02"
)

// migBody returns the dollar-quoted body (prosrc) of the first function created after marker in the up section of file.
func migBody(t *testing.T, file, marker string) string {
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

// namesCatalog returns the keyed catalog lines ("key" -> "value") of schemas app and audit.
func namesCatalog(t *testing.T, q catalogQuerier) map[string]string {
	t.Helper()
	rows, err := q.Query(context.Background(), `
SELECT 'rel:' || c.oid::regclass::text, format('%s|%s|%s', c.relkind, c.relowner::regrole, coalesce(c.relacl::text, ''))
  FROM pg_class c WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace)
UNION ALL
SELECT 'col:' || c.oid::regclass::text || '.' || a.attname, format('%s|%s|%s', format_type(a.atttypid, a.atttypmod), a.attnotnull, a.atthasdef)
  FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
  WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND c.relkind = 'r' AND a.attnum > 0 AND NOT a.attisdropped
UNION ALL
SELECT 'acl:' || c.oid::regclass::text || '.' || a.attname, a.attacl::text
  FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
  WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND cardinality(a.attacl) > 0
UNION ALL
SELECT 'fn:' || p.oid::regprocedure::text, format('%s|%s|%s|%s|%s', p.proowner::regrole, p.prosecdef, coalesce(p.proacl::text, ''),
    coalesce(p.proconfig::text, ''), md5(p.prosrc))
  FROM pg_proc p WHERE p.pronamespace IN ('app'::regnamespace, 'audit'::regnamespace)
UNION ALL
SELECT 'pol:' || p.polrelid::regclass::text || '.' || p.polname, format('%s|%s|%s|%s', p.polcmd, p.polroles::regrole[]::text,
    coalesce(pg_get_expr(p.polqual, p.polrelid), ''), coalesce(pg_get_expr(p.polwithcheck, p.polrelid), ''))
  FROM pg_policy p
UNION ALL
SELECT 'rls:' || c.oid::regclass::text, format('%s|%s', c.relrowsecurity, c.relforcerowsecurity)
  FROM pg_class c WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND c.relkind = 'r'
UNION ALL
SELECT 'nsp:' || n.nspname, coalesce(n.nspacl::text, '') FROM pg_namespace n WHERE n.nspname IN ('app', 'audit')
UNION ALL
SELECT 'con:' || c.conrelid::regclass::text || '.' || c.conname, format('%s|%s', c.convalidated, md5(pg_get_constraintdef(c.oid)))
  FROM pg_constraint c WHERE c.connamespace IN ('app'::regnamespace, 'audit'::regnamespace)
UNION ALL
SELECT 'trg:' || t.tgrelid::regclass::text || '.' || t.tgname, format('%s|%s', t.tgenabled, md5(pg_get_triggerdef(t.oid)))
  FROM pg_trigger t WHERE NOT t.tgisinternal`)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		out[k] = v
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("catalog: %v", err)
	}
	return out
}

// namesDelta returns the sorted keys whose value differs (or that exist on one side only).
func namesDelta(a, b map[string]string) []string {
	var keys []string
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			keys = append(keys, k)
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return slices.Compact(keys)
}

// mcpnWantDelta is the exact set of keys version 19 adds or changes over version 18.
var mcpnWantDelta = []string{
	"acl:app.mcp_connections.display_name",
	"col:app.mcp_connections.display_name",
	"con:app.mcp_connections.mcp_connections_display_name_valid",
	"con:audit.events.events_action_valid",
	"fn:audit.capture_mcp_connections()",
	"fn:audit.events_attribute()",
	"trg:app.mcp_connections.audit_capture_name",
}

const (
	mcpnWantCheck = `CHECK (((display_name IS NULL) OR (((char_length(display_name) >= 1) AND (char_length(display_name) <= 80)) ` +
		`AND (display_name = btrim(display_name)) AND ((display_name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f\u2028\u2029]'::text))))`
	mcpnWantTrigger = `CREATE TRIGGER audit_capture_name AFTER UPDATE OF display_name ON app.mcp_connections FOR EACH ROW ` +
		`WHEN ((old.display_name IS DISTINCT FROM new.display_name)) EXECUTE FUNCTION audit.capture_mcp_connections()`
)

// mcpnSeed creates the committed fixture at the current version: an org_owner, an unbound user, the Organization, a
// pending and an active connection with their 'mcp' sessions (the activation records mcp.connection_established).
func mcpnSeed(t *testing.T, setup *pgx.Conn) {
	t.Helper()
	ctx := context.Background()
	for _, s := range []string{
		`INSERT INTO app.users (id, display_name) VALUES ('` + mcpnUser + `', 'Names Owner'), ('` + mcpnOther + `', 'Names Other')`,
		`INSERT INTO app.organizations (id, name) VALUES ('` + mcpnOrg + `', 'Names Catalog')`,
		`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + mcpnOrg + `', '` + mcpnUser + `', 'organization', '` + mcpnOrg + `', NULL, 'org_owner', 'membership')`,
		`INSERT INTO app.sessions (id, user_id, kind, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at) VALUES
  ('` + mcpnSessP + `', '` + mcpnUser + `', 'mcp', sha256('names-p'::bytea), sha256('names-pc'::bytea), now(), now() + interval '1 day', now() + interval '2 days'),
  ('` + mcpnSessA + `', '` + mcpnUser + `', 'mcp', sha256('names-a'::bytea), sha256('names-ac'::bytea), now(), now() + interval '1 day', now() + interval '2 days')`,
		`INSERT INTO app.mcp_connections (id, session_id, user_id, organization_id, host, resource, pending_expires_at) VALUES
  ('` + mcpnPending + `', '` + mcpnSessP + `', '` + mcpnUser + `', '` + mcpnOrg + `', 'names.localhost:543', 'https://names.localhost:543/mcp/c/p', now() + interval '30 minutes'),
  ('` + mcpnActive + `', '` + mcpnSessA + `', '` + mcpnUser + `', '` + mcpnOrg + `', 'names.localhost:543', 'https://names.localhost:543/mcp/c/a', now() + interval '30 minutes')`,
		`UPDATE app.mcp_connections SET state = 'active', client_id = 'client_01NAMES', activated_at = now() WHERE id = '` + mcpnActive + `'`,
	} {
		if _, err := setup.Exec(ctx, s); err != nil {
			t.Fatalf("names fixture: %v", err)
		}
	}
}

// mcpnRows hashes every touched app and audit row as JSON without the display_name key, so a version-18 world and the
// same world at version 19 (every existing connection unnamed) compare equal.
func mcpnRows(t *testing.T, setup *pgx.Conn) string {
	t.Helper()
	var parts []string
	for _, tbl := range []string{"app.users", "app.sessions", "app.organizations", "app.role_bindings", "app.mcp_connections", "audit.events"} {
		var h string
		if err := setup.QueryRow(context.Background(), `SELECT count(*) || ':' || coalesce(md5(string_agg((to_jsonb(r) - 'display_name')::text, E'\n'
  ORDER BY (to_jsonb(r) - 'display_name')::text)), '-') FROM `+tbl+` r`).Scan(&h); err != nil {
			t.Fatal(err)
		}
		parts = append(parts, tbl+"="+h)
	}
	return strings.Join(parts, ";")
}

// TestMCPNamesCatalog (B-MCP-CATALOG): CT-MCPN-1, CT-MCPN-CHECK, CT-MCPN-DML and CT-MCPN-NEG.
func TestMCPNamesCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r18 := runnerFor(t, cfg, subsetFS(t, 18))
	objRunUp(t, r18, "current_version\t18\n")
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	rt := quote(cfg.RuntimeRole())
	s18 := namesCatalog(t, cat)
	r := runnerFor(t, cfg, subsetFS(t, 19))
	objRunUp(t, r, "applied\t00019_mcp_connection_names.sql\ncurrent_version\t19\n")
	s19 := namesCatalog(t, cat)
	mcpnSeed(t, setup)
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

	t.Run("CT-MCPN-1 the exact version-19 delta", func(t *testing.T) {
		if got := namesDelta(s18, s19); !slices.Equal(got, mcpnWantDelta) {
			t.Fatalf("version-18 -> 19 delta: %v, want exactly %v", got, mcpnWantDelta)
		}
		if got := s19["col:app.mcp_connections.display_name"]; got != "text|f|f" {
			t.Errorf("display_name column: %s, want nullable text without a default", got)
		}
		if got := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'app.mcp_connections'::regclass
  AND conname = 'mcp_connections_display_name_valid'`); got != mcpnWantCheck {
			t.Errorf("name CHECK:\n got %s\nwant %s", got, mcpnWantCheck)
		}
		if got := one(`SELECT pg_get_triggerdef(t.oid) FROM pg_trigger t WHERE t.tgrelid = 'app.mcp_connections'::regclass
  AND t.tgname = 'audit_capture_name'`); got != mcpnWantTrigger {
			t.Errorf("name trigger:\n got %s\nwant %s", got, mcpnWantTrigger)
		}
		// The column ACL names the runtime only, with INSERT and UPDATE.
		if got := one(`SELECT string_agg(format('%s:%s', CASE WHEN x.grantee = $1::regrole THEN 'rt' ELSE x.grantee::regrole::text END,
  x.privilege_type), ',' ORDER BY x.privilege_type) FROM pg_attribute a, aclexplode(a.attacl) x
  WHERE a.attrelid = 'app.mcp_connections'::regclass AND a.attname = 'display_name'`, cfg.RuntimeRole()); got != "rt:INSERT,rt:UPDATE" {
			t.Errorf("display_name column ACL: %s, want rt:INSERT,rt:UPDATE", got)
		}
		for fn, want := range map[string]string{
			"audit.events_attribute()":        md5hex(migBody(t, "00019_mcp_connection_names.sql", "CREATE OR REPLACE FUNCTION audit.events_attribute()")),
			"audit.capture_mcp_connections()": md5hex(migBody(t, "00019_mcp_connection_names.sql", "CREATE OR REPLACE FUNCTION audit.capture_mcp_connections()")),
			"app.mcp_connections_guard()":     md5hex(mcpBody(t, "CREATE FUNCTION app.mcp_connections_guard()")),
		} {
			if got := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure($1)`, fn); got != want {
				t.Errorf("%s body digest %s, want %s (the reviewed file)", fn, got, want)
			}
		}
		def := one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'audit.events'::regclass AND conname = 'events_action_valid'`)
		if n := strings.Count(def, "::text"); n != 47 || !strings.Contains(def, "'mcp.connection_renamed'::text") {
			t.Errorf("events_action_valid names %d actions (want 47 including mcp.connection_renamed): %s", n, def)
		}
		for _, k := range []string{"pol:app.mcp_connections.mcp_connections_select", "pol:app.mcp_connections.mcp_connections_insert",
			"pol:app.mcp_connections.mcp_connections_update", "pol:app.sessions.sessions_insert", "fn:app.mcp_connections_guard()",
			"trg:app.mcp_connections.mcp_connections_guard", "rls:app.mcp_connections", "rls:audit.events"} {
			if s19[k] == "" || s19[k] != s18[k] {
				t.Errorf("%s changed or missing: %q -> %q", k, s18[k], s19[k])
			}
		}
	})

	t.Run("CT-MCPN-CHECK the stored-shape CHECK and the strict name capture", func(t *testing.T) {
		try := func(name string) (code, constraint string, events int) {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `UPDATE app.mcp_connections SET display_name = $2 WHERE id = $1`, mcpnActive, name); err != nil {
				var pe *pgconn.PgError
				if errors.As(err, &pe) {
					return pe.Code, pe.ConstraintName, 0
				}
				t.Fatalf("update: %v", err)
			}
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'mcp.connection_renamed' AND resource_id = $1::uuid
  AND organization_id = $2::uuid AND subject_user_id = $3::uuid AND resource_type = 'mcp_connection' AND outcome = 'success'
  AND reason IS NULL AND group_id IS NULL AND operation IS NULL`, mcpnActive, mcpnOrg, mcpnUser).Scan(&events); err != nil {
				t.Fatal(err)
			}
			return "", "", events
		}
		for _, ok := range []string{strings.Repeat("a", 80), strings.Repeat("\U0001F600", 80), "a b", "<script>alert(1)</script>", "\uFEFFx"} {
			if code, con, n := try(ok); code != "" || n != 1 {
				t.Errorf("name %q: %s %s, %d rename events; want stored with exactly one event", ok, code, con, n)
			}
		}
		for _, bad := range []string{"", strings.Repeat("a", 81), "a\nb", "a\tb", " a", "a ", "a\u0085b", "a\u2028b", "a\u2029b", "a\x7fb", "a\u009fb"} {
			if code, con, _ := try(bad); code != "23514" || con != "mcp_connections_display_name_valid" {
				t.Errorf("name %q: %s %s, want 23514 mcp_connections_display_name_valid", bad, code, con)
			}
		}
		// A same-value write records nothing; a clear records one event.
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		count := func() int {
			var c int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'mcp.connection_renamed' AND resource_id = $1::uuid`,
				mcpnPending).Scan(&c); err != nil {
				t.Fatal(err)
			}
			return c
		}
		for i, s := range []string{`UPDATE app.mcp_connections SET display_name = 'Same' WHERE id = $1`,
			`UPDATE app.mcp_connections SET display_name = 'Same' WHERE id = $1`, `UPDATE app.mcp_connections SET display_name = NULL WHERE id = $1`} {
			if tag, err := tx.Exec(ctx, s, mcpnPending); err != nil || tag.RowsAffected() != 1 {
				t.Fatalf("write %d: %v", i, err)
			}
			if want := []int{1, 1, 2}[i]; count() != want {
				t.Errorf("after write %d (set, same value, clear): %d rename events, want %d", i, count(), want)
			}
		}
	})

	t.Run("CT-MCPN-DML the runtime name path writes the name only, own row, with a current role", func(t *testing.T) {
		try := func(user, sql string) (int64, string) {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+rt); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true),
  set_config('app.group_id', '', true), set_config('app.login_binding', '', true)`, user, mcpnOrg); err != nil {
				t.Fatal(err)
			}
			tag, err := tx.Exec(ctx, sql, mcpnActive)
			var pe *pgconn.PgError
			if errors.As(err, &pe) {
				return 0, pe.Code
			} else if err != nil {
				t.Fatalf("%s: %v", sql, err)
			}
			return tag.RowsAffected(), ""
		}
		if n, code := try(mcpnUser, `UPDATE app.mcp_connections SET display_name = 'Runtime' WHERE id = $1`); n != 1 || code != "" {
			t.Errorf("own row with a current role: %d rows, %s", n, code)
		}
		if n, code := try(mcpnOther, `UPDATE app.mcp_connections SET display_name = 'Runtime' WHERE id = $1`); n != 0 || code != "" {
			t.Errorf("another user's row: %d rows, %s, want 0 rows", n, code)
		}
		for _, col := range []string{"host = 'x.localhost'", "resource = 'https://x/mcp/c/x'", "user_id = user_id", "organization_id = organization_id",
			"session_id = session_id", "pending_expires_at = now()", "created_at = now()", "id = id"} {
			if _, code := try(mcpnUser, `UPDATE app.mcp_connections SET display_name = 'Runtime', `+col+` WHERE id = $1`); code != "42501" {
				t.Errorf("name write with %s: %s, want 42501", col, code)
			}
		}
		if _, code := try(mcpnUser, `DELETE FROM app.mcp_connections WHERE id = $1`); code != "42501" {
			t.Errorf("runtime DELETE: %s, want 42501", code)
		}
	})

	t.Run("CT-MCPN-NEG rolled-back missing-protection detectors", func(t *testing.T) {
		for _, n := range []struct{ id, sql, want string }{
			{"N1 missing name CHECK", `ALTER TABLE app.mcp_connections DROP CONSTRAINT mcp_connections_display_name_valid`,
				"con:app.mcp_connections.mcp_connections_display_name_valid"},
			{"N2 runtime UPDATE of another column", `GRANT UPDATE (host) ON app.mcp_connections TO ` + rt, "acl:app.mcp_connections.host"},
			{"N3 NO FORCE on connections", `ALTER TABLE app.mcp_connections NO FORCE ROW LEVEL SECURITY`, "rls:app.mcp_connections"},
			{"N4 dropped name capture", `DROP TRIGGER audit_capture_name ON app.mcp_connections`, "trg:app.mcp_connections.audit_capture_name"},
			{"N5 definer capture", `ALTER FUNCTION audit.capture_mcp_connections() SECURITY DEFINER`, "fn:audit.capture_mcp_connections()"},
			{"N6 runtime DELETE on connections", `GRANT DELETE ON app.mcp_connections TO ` + rt, "rel:app.mcp_connections"},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			got := namesDelta(s19, namesCatalog(t, tx))
			_ = tx.Rollback(ctx)
			if !slices.Equal(got, []string{n.want}) {
				t.Errorf("%s: the detector reported %v, want [%s]", n.id, got, n.want)
			}
		}
		if got := namesDelta(s19, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", got)
		}
	})
}

// TestMCPNamesMigration (B-MCP-CATALOG): MG-MCPN-1..4.
func TestMCPNamesMigration(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	down := func(r *Runner) (string, error) {
		var b bytes.Buffer
		err := r.Down(ctx, &b)
		return b.String(), err
	}
	var s18, s19 map[string]string
	var rows string
	r := runnerFor(t, cfg, subsetFS(t, 19))

	t.Run("MG-MCPN-1 a populated version-18 world to 19, every row unnamed and otherwise identical", func(t *testing.T) {
		reset(t, cfg)
		objRunUp(t, runnerFor(t, cfg, subsetFS(t, 18)), "current_version\t18\n")
		mcpnSeed(t, setup)
		s18 = namesCatalog(t, cat)
		rows = mcpnRows(t, setup)
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r.Up(ctx, b) }); out != "applied\t00019_mcp_connection_names.sql\ncurrent_version\t19\n" {
			t.Fatalf("up: %q", out)
		}
		s19 = namesCatalog(t, cat)
		if got := namesDelta(s18, s19); !slices.Equal(got, mcpnWantDelta) {
			t.Fatalf("delta %v, want %v", got, mcpnWantDelta)
		}
		if got := mcpnRows(t, setup); got != rows {
			t.Fatal("rows changed by the upgrade")
		}
		var named, established int
		if err := setup.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.mcp_connections WHERE display_name IS NOT NULL),
  (SELECT count(*) FROM audit.events WHERE action = 'mcp.connection_established')`).Scan(&named, &established); err != nil || named != 0 || established != 1 {
			t.Fatalf("after the upgrade: %d named rows (want 0), %d established events (want 1): %v", named, established, err)
		}
	})

	t.Run("MG-MCPN-2 down refused with a named row or a rename event, catalog unchanged", func(t *testing.T) {
		if _, err := setup.Exec(ctx, `UPDATE app.mcp_connections SET display_name = 'Laptop' WHERE id = $1`, mcpnActive); err != nil {
			t.Fatal(err)
		}
		if out, err := down(r); err == nil || !strings.Contains(err.Error(),
			"refusing to roll back 00019: 1 named connections / 1 rename audit rows (history is never discarded)") {
			t.Fatalf("down with a named row: %v %q", err, out)
		}
		if _, err := setup.Exec(ctx, `UPDATE app.mcp_connections SET display_name = NULL WHERE id = $1`, mcpnActive); err != nil {
			t.Fatal(err)
		}
		if out, err := down(r); err == nil || !strings.Contains(err.Error(),
			"refusing to roll back 00019: 0 named connections / 2 rename audit rows (history is never discarded)") {
			t.Fatalf("down with rename history: %v %q", err, out)
		}
		if got := namesDelta(s19, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("the refused down changed the catalog: %v", got)
		}
	})

	t.Run("MG-MCPN-3 down and up on an unnamed world restore the exact 18 and 19 catalogs", func(t *testing.T) {
		auditPurge(t, setup)
		rows = mcpnRows(t, setup)
		out, err := down(r)
		if err != nil || out != "rolled_back\t00019_mcp_connection_names.sql\ncurrent_version\t18\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		if got := namesDelta(s18, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down differs from version 18: %v", got)
		}
		metadataThroughVersion18(t, cfg, r)
		if got := mcpnRows(t, setup); got != rows {
			t.Fatal("rows changed by the rollback")
		}
		objRunUp(t, r, "applied\t00019_mcp_connection_names.sql\ncurrent_version\t19\n")
		if out := run(t, "reapply", func(ctx context.Context, b *bytes.Buffer) error { return r.Reapply(ctx, b) }); out !=
			"rolled_back\t00019_mcp_connection_names.sql\napplied\t00019_mcp_connection_names.sql\ncurrent_version\t19\n" {
			t.Fatalf("reapply: %q", out)
		}
		if got := namesDelta(s19, namesCatalog(t, cat)); len(got) != 0 {
			t.Fatalf("catalog after down/up/reapply differs from version 19: %v", got)
		}
		if got := mcpnRows(t, setup); got != rows {
			t.Fatal("rows changed by up/reapply")
		}
	})

	t.Run("MG-MCPN-4 a rolled-back up leaves version 18 and the owner", func(t *testing.T) {
		if out, err := down(r); err != nil || !strings.HasSuffix(out, "current_version\t18\n") {
			t.Fatalf("down: %v %q", err, out)
		}
		// A stray runtime column grant breaks the version-18 precondition: the up fails and rolls back entirely.
		if _, err := setup.Exec(ctx, `GRANT UPDATE (host) ON app.mcp_connections TO `+quote(cfg.RuntimeRole())); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := r.Up(ctx, &buf); err == nil || !strings.Contains(err.Error(), "version-18 precondition") {
			t.Fatalf("up with a stray grant: %v %q", err, buf.String())
		}
		if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t18\n" {
			t.Fatalf("version after the failed up: %q", out)
		}
		if got := namesDelta(s18, namesCatalog(t, cat)); !slices.Equal(got, []string{"acl:app.mcp_connections.host"}) {
			t.Errorf("after the failed up the catalog differs by %v, want only the stray grant", got)
		}
		if _, err := setup.Exec(ctx, `REVOKE UPDATE (host) ON app.mcp_connections FROM `+quote(cfg.RuntimeRole())); err != nil {
			t.Fatal(err)
		}
		rlsOwnerRestored(t, cfg, r)
		if got := mcpnRows(t, setup); got != rows {
			t.Fatal("rows changed")
		}
		objRunUp(t, r, "applied\t00019_mcp_connection_names.sql\ncurrent_version\t19\n")
	})
}
