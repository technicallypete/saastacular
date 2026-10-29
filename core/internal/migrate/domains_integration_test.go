//go:build integration

// G016 B-DOMAINS-CATALOG (Plan r4 §4, §6.2) against the disposable PostgreSQL of the migrations suite, on the full
// normal set (version 11). TestDomainsCatalog: CT-DOM-1 the exact version-11 delta (00011's own inventory postcondition
// held at apply time; the version-10 detector reports exactly the delta's parts; the reviewed function bodies equal
// the embedded migration; the auth reader returns exactly four columns; every owner/admin predicate is present),
// CT-DOM-NEG rolled-back missing-protection negatives (each detected by the catalog parts, and each protection shown
// to hold behaviourally on the intact catalog), and CT-DOM-UPDOWN down refused while domains, slugs, tenant-host login
// transactions or domain history exist, allowed when empty with the exact version-10 catalog, then up again with an
// identical catalog. These three run on the version-11 set (subsetFS 11). CT-DOM-12 (G016 security amendment, Plan
// amendment A2) is the seed-before-12 transition: legacy NULL-bound browser rows seeded at version 11 are unchanged by
// the upgrade to version 12, refused by the bound reader at every origin, still revocable by the runtime and expirable
// by the restore-time invalidation, never revived by the down migration, and the down refuses bound sessions, live
// legacy sessions and duplicate non-removed hostnames. Setup and rolled-back controls use the attested superuser;
// catalog reads use the owner; reader, revocation and runtime negatives use the attested runtime login.
package migrate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"core/migrations"
)

// domBody returns the dollar-quoted body (prosrc) of the first function created after marker in the 00011 up section.
func domBody(t *testing.T, marker string) string {
	t.Helper()
	b, err := fs.ReadFile(migrations.FS, "00011_organization_domains.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, _ := strings.Cut(string(b), "-- +goose Down")
	i := strings.Index(up, marker)
	if i < 0 {
		t.Fatalf("00011: %q not found", marker)
	}
	j := strings.Index(up[i:], "AS $$")
	if j < 0 {
		t.Fatalf("00011: no body after %q", marker)
	}
	start := i + j + len("AS $$")
	k := strings.Index(up[start:], "$$")
	return up[start : start+k]
}

// domParts reads the version-11 delta's catalog parts.
func domParts(t *testing.T, q catalogQuerier) map[string]string {
	t.Helper()
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
	p := map[string]string{}
	p["dom-columns"] = one(`SELECT string_agg(format('%s:%s:%s:%s', a.attname, format_type(a.atttypid, a.atttypmod),
  CASE WHEN a.attnotnull THEN 't' ELSE 'f' END, coalesce(pg_get_expr(d.adbin, d.adrelid), '')), ';' ORDER BY a.attnum)
FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = to_regclass('app.organization_domains') AND a.attnum > 0 AND NOT a.attisdropped`)
	p["dom-constraints"] = one(`SELECT string_agg(conname || ':' || pg_get_constraintdef(oid), ';' ORDER BY conname) FROM pg_constraint
WHERE conrelid IN (to_regclass('app.organization_domains'), to_regclass('app.organizations'), to_regclass('app.login_transactions'))`)
	p["dom-indexes"] = one(`SELECT string_agg(pg_get_indexdef(i.indexrelid), ';' ORDER BY c.relname) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
WHERE i.indrelid IN (to_regclass('app.organization_domains'), to_regclass('app.organizations'))`)
	p["dom-triggers"] = one(`SELECT string_agg(c.relname || '.' || t.tgname || ':' || t.tgenabled::text || ':' || pg_get_triggerdef(t.oid),
  E'\n' ORDER BY c.relname, t.tgname)
FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
WHERE c.oid IN (to_regclass('app.organization_domains'), to_regclass('app.organizations')) AND NOT t.tgisinternal`)
	p["dom-policies"] = one(`SELECT string_agg(p.polname || ':' || p.polcmd::text || ':' || coalesce(pg_get_expr(p.polqual, p.polrelid), '') ||
  ':' || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), ''), E'\n' ORDER BY p.polname)
FROM pg_policy p WHERE p.polrelid IN (to_regclass('app.organization_domains'), to_regclass('app.organizations'), to_regclass('app.role_bindings'))`)
	p["dom-acl"] = one(`SELECT string_agg(x, ';' ORDER BY x) FROM (
  SELECT c.relname || ':' || a.grantee::regrole::text || ':' || a.privilege_type AS x
  FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
  WHERE c.oid IN (to_regclass('app.organization_domains'), to_regclass('app.organizations'), to_regclass('app.role_bindings')) AND a.grantee <> c.relowner
  UNION ALL
  SELECT c.relname || '.' || at.attname || ':' || a.grantee::regrole::text || ':' || a.privilege_type
  FROM pg_class c JOIN pg_attribute at ON at.attrelid = c.oid AND at.attnum > 0 AND NOT at.attisdropped
  CROSS JOIN LATERAL aclexplode(at.attacl) a
  WHERE c.oid IN (to_regclass('app.organization_domains'), to_regclass('app.organizations'), to_regclass('app.role_bindings'))) s`)
	p["dom-functions"] = one(`SELECT string_agg(p.pronamespace::regnamespace::text || '.' || p.proname || ':' || p.proowner::regrole::text || ':' ||
  p.prosecdef::text || ':' || coalesce(array_to_string(p.proacl, ','), '-'), ';' ORDER BY p.pronamespace::regnamespace::text, p.proname)
FROM pg_proc p WHERE p.pronamespace IN ('app'::regnamespace, 'audit'::regnamespace)`)
	p["dom-reader-columns"] = one(`SELECT string_agg(n || ':' || format_type(ty, NULL), ',' ORDER BY o) FROM pg_proc p,
  unnest(p.proargnames, p.proallargtypes::oid[], p.proargmodes::text[]) WITH ORDINALITY AS u(n, ty, m, o)
WHERE p.oid = to_regprocedure('app.hostname_state(text, text)') AND m = 't'`)
	for name, sig := range map[string]string{"dom-guard-src": "app.organization_domains_guard()", "dom-slug-guard-src": "app.organizations_slug_guard()",
		"dom-reader-src": "app.hostname_state(text, text)", "dom-member-src": "app.is_organization_member(uuid, uuid)",
		"dom-capture-src": "audit.capture_organization_domains()"} {
		p[name] = one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = to_regprocedure('` + sig + `')`)
	}
	return p
}

// domViolations is the version-11 delta detector: the sorted parts that differ from want.
func domViolations(t *testing.T, q catalogQuerier, want map[string]string) []string {
	t.Helper()
	p := domParts(t, q)
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

// TestDomainsCatalog (B-DOMAINS-CATALOG): CT-DOM-1, CT-DOM-NEG and CT-DOM-UPDOWN on the version-11 set, then CT-DOM-12.
func TestDomainsCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r := runnerFor(t, cfg, subsetFS(t, 11))
	objRunUp(t, r, "applied\t00011_organization_domains.sql\ncurrent_version\t11\n")
	cat := catalogConn(t, cfg)
	setup := rlsSetupConn(t, cfg)
	clean := domParts(t, cat)

	t.Run("CT-DOM-1 the exact version-11 delta", func(t *testing.T) {
		metadataThroughVersion11Only(t, cfg, r)
		for part, marker := range map[string]string{"dom-guard-src": "CREATE FUNCTION app.organization_domains_guard()",
			"dom-slug-guard-src": "CREATE FUNCTION app.organizations_slug_guard()", "dom-reader-src": "CREATE FUNCTION app.hostname_state(",
			"dom-member-src": "CREATE FUNCTION app.is_organization_member(", "dom-capture-src": "CREATE FUNCTION audit.capture_organization_domains()"} {
			if clean[part] != md5hex(domBody(t, marker)) {
				t.Errorf("%s differs from the reviewed migration body", part)
			}
		}
		if got := clean["dom-reader-columns"]; got != "organization_id:uuid,kind:text,routing:boolean,issuance:boolean" {
			t.Errorf("auth reader result columns: %q", got)
		}
		for _, pol := range []string{"organization_domains_select:r:", "organization_domains_insert:a:", "organization_domains_update:w:",
			"organizations_update_slug:w:"} {
			found := false
			for _, l := range strings.Split(clean["dom-policies"], "\n") {
				if strings.HasPrefix(l, pol) {
					found = true
					if strings.Count(l, "'org_owner'") < 1 || strings.Count(l, "'org_admin'") < 1 || !strings.Contains(l, "rls_org_role()") {
						t.Errorf("policy %s lacks the owner/admin predicate: %s", pol, l)
					}
				}
			}
			if !found {
				t.Errorf("policy %s missing", pol)
			}
		}
		// Everything outside the delta equals version 10: the version-10 detector reports exactly the delta's parts.
		got := mcpViolations(t, cat, cfg)
		want := []string{"acl", "audit-capture-rest", "audit-src", "audit-functions", "functions", "policy-inventory", "rls-flags"}
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("version-10 detector at version 11: %v, want exactly the delta %v", got, want)
		}
	})

	t.Run("CT-DOM-NEG rolled-back missing-protection negatives", func(t *testing.T) {
		ar := quote(cfg.Database + "_auth_reader")
		// The protections hold on the intact catalog (rolled back): a removed-row proof update and a second live claim of
		// one hostname are refused.
		seed := `INSERT INTO app.users (id, display_name) VALUES ('0f16c000-0000-4000-8000-0000000000c1', 'Catalog');
INSERT INTO app.organizations (id, name) VALUES ('0f16c000-0000-4000-8000-0000000000c2', 'Catalog');
INSERT INTO app.organization_domains (id, organization_id, hostname, generation, challenge_value, challenge_expires_at, created_by)
VALUES ('0f16c000-0000-4000-8000-0000000000c3', '0f16c000-0000-4000-8000-0000000000c2', 'portal.catalog.example', 7,
        repeat('a', 52), now() + interval '1 hour', '0f16c000-0000-4000-8000-0000000000c1');
UPDATE app.organization_domains SET state = 'removed', removed_at = now(), tls_state = 'withdrawing', version = version + 1
WHERE id = '0f16c000-0000-4000-8000-0000000000c3'`
		removedUpdate := `UPDATE app.organization_domains SET verified_until = now() + interval '1 day', version = version + 1
WHERE id = '0f16c000-0000-4000-8000-0000000000c3'`
		secondLive := `INSERT INTO app.organization_domains (id, organization_id, hostname, generation, challenge_value, challenge_expires_at, created_by)
VALUES (gen_random_uuid(), '0f16c000-0000-4000-8000-0000000000c2', 'live.catalog.example', 8, repeat('b', 52), now() + interval '1 hour',
        '0f16c000-0000-4000-8000-0000000000c1'), (gen_random_uuid(), '0f16c000-0000-4000-8000-0000000000c2', 'live.catalog.example', 9,
        repeat('c', 52), now() + interval '1 hour', '0f16c000-0000-4000-8000-0000000000c1')`
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
		sqlState := func(err error) string {
			if pg, ok := err.(*pgconn.PgError); ok {
				return pg.Code
			}
			return ""
		}
		if err := probe("intact", "", removedUpdate); sqlState(err) != "23514" {
			t.Fatalf("intact guard: removed-row proof update %v, want 23514", err)
		}
		if err := probe("intact", "", secondLive); sqlState(err) != "23505" {
			t.Fatalf("intact live-hostname key: second live claim %v, want 23505", err)
		}
		readerExtra := `GRANT CREATE ON SCHEMA app TO ` + ar + `; SET LOCAL ROLE ` + ar + `; DROP FUNCTION app.hostname_state(text, text);
CREATE FUNCTION app.hostname_state(p_kind text, p_key text) RETURNS TABLE (organization_id uuid, kind text, routing boolean, issuance boolean,
  challenge_value text) LANGUAGE sql STABLE SECURITY DEFINER SET search_path = pg_catalog, pg_temp AS $f$
  SELECT d.organization_id, 'custom'::text, false, false, d.challenge_value FROM app.organization_domains d WHERE d.hostname = p_key $f$;
RESET ROLE; REVOKE CREATE ON SCHEMA app FROM ` + ar
		for _, n := range []struct {
			id, sql string
			want    []string
			stmt    string // a statement the missing protection now admits (or "")
		}{
			{"N1 insert policy without the owner/admin predicate", `ALTER POLICY organization_domains_insert ON app.organization_domains WITH CHECK (` +
				`organization_id = (SELECT app.ctx_org()) AND state = 'pending' AND version = 0 AND generation >= 1)`, []string{"dom-policies"}, ""},
			{"N2 slug policy without the owner/admin predicate", `ALTER POLICY organizations_update_slug ON app.organizations USING (` +
				`id = (SELECT app.ctx_org())) WITH CHECK (id = (SELECT app.ctx_org()))`, []string{"dom-policies"}, ""},
			{"N3 guard trigger missing", `DROP TRIGGER organization_domains_guard ON app.organization_domains`, []string{"dom-triggers"}, removedUpdate},
			{"N4 removed-row update admitted", `CREATE OR REPLACE FUNCTION app.organization_domains_guard() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $f$ BEGIN RETURN NEW; END $f$`, []string{"dom-guard-src"}, removedUpdate},
			{"N5 duplicate live hostname", `DROP INDEX app.organization_domains_hostname_live`, []string{"dom-indexes"}, secondLive},
			{"N6 auth reader returning extra columns", readerExtra, []string{"dom-functions", "dom-reader-columns", "dom-reader-src"}, ""},
			{"N7 slug guard missing", `DROP TRIGGER organizations_slug_guard ON app.organizations`, []string{"dom-triggers"}, ""},
		} {
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, n.sql); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("%s: %v", n.id, err)
			}
			v := domViolations(t, tx, clean)
			_ = tx.Rollback(ctx)
			if !slices.Equal(v, n.want) {
				t.Errorf("%s: the detector reported %v, want %v", n.id, v, n.want)
			}
			if n.stmt != "" {
				if err := probe(n.id, n.sql, n.stmt); err != nil {
					t.Errorf("%s: the missing protection still refused the probe: %v", n.id, err)
				}
			}
		}
		if v := domViolations(t, cat, clean); len(v) != 0 {
			t.Fatalf("catalog changed by the rolled-back controls: %v", v)
		}
	})

	t.Run("CT-DOM-UPDOWN down refused with domain state, exact version-10 catalog when empty, up again", func(t *testing.T) {
		s11 := snapshot(t, r)
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
		exec(`INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, created_at, expires_at,
  callback_origin) VALUES (convert_to(repeat('d', 32), 'UTF8'), convert_to(repeat('d', 32), 'UTF8'), repeat('v', 43), 'c', 'https://o', '/',
  now(), now() + interval '5 minutes', 'https://auth.localhost:543')`)
		if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00011: domains, slugs or tenant-host login transactions exist") {
			t.Fatalf("down with a tenant-host login transaction: %v %q", err, out)
		}
		exec(`DELETE FROM app.login_transactions WHERE callback_origin IS NOT NULL`)
		if got := snapshot(t, r); got != s11 {
			t.Fatal("the refused down changed the catalog")
		}
		out, err := down()
		if err != nil || out != "rolled_back\t00011_organization_domains.sql\ncurrent_version\t10\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		metadataThroughVersion10Only(t, cfg, r)
		if v := mcpViolations(t, cat, cfg); len(v) != 0 {
			t.Fatalf("version-10 detector after down: %v", v)
		}
		objRunUp(t, r, "applied\t00011_organization_domains.sql\ncurrent_version\t11\n")
		if got := snapshot(t, r); got != s11 {
			t.Fatalf("catalog after down/up differs:\n%s\n---\n%s", got, s11)
		}
		if v := domViolations(t, cat, clean); len(v) != 0 {
			t.Fatalf("version-11 detector after up again: %v", v)
		}
		// Domain history is never discarded: a slug change leaves organization.slug_changed and the down refuses.
		exec(`INSERT INTO app.organizations (id, name, slug) VALUES ('0f16c000-0000-4000-8000-0000000000d1', 'Catalog Slug', 'catalog-slug')`)
		exec(`UPDATE app.organizations SET slug = NULL WHERE id = '0f16c000-0000-4000-8000-0000000000d1'`)
		if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00011: audit.events holds domain history") {
			t.Fatalf("down with domain history: %v %q", err, out)
		}
		reset(t, cfg)
	})

	t.Run("CT-DOM-12 legacy sessions across the version-12 transition", func(t *testing.T) {
		reset(t, cfg)
		r11 := runnerFor(t, cfg, subsetFS(t, 11))
		objRunUp(t, r11, "applied\t00011_organization_domains.sql\ncurrent_version\t11\n")
		exec := func(sql string, args ...any) {
			t.Helper()
			if _, err := setup.Exec(ctx, sql, args...); err != nil {
				t.Fatal(err)
			}
		}
		const (
			userA, userB = "0f16c000-0000-4000-8000-0000000012a1", "0f16c000-0000-4000-8000-0000000012b1"
			orgA         = "0f16c000-0000-4000-8000-0000000012c1"
			// legacyRevoke (userA) is revoked by the runtime; legacyExpire (userB) starts live and is expired by the
			// restore-time invalidation; legacyDead (userB) is already past its absolute expiry.
			legacyRevoke, legacyExpire, legacyDead = "0f16c000-0000-4000-8000-0000000012d1", "0f16c000-0000-4000-8000-0000000012d2",
				"0f16c000-0000-4000-8000-0000000012d3"
		)
		hash := func(id string) []byte { h := sha256.Sum256([]byte("ct-dom-12:" + id)); return h[:] }
		exec(`INSERT INTO app.users (id, display_name) VALUES ($1::text::uuid, 'Legacy A'), ($2::text::uuid, 'Legacy B')`, userA, userB)
		exec(`INSERT INTO app.organizations (id, name) VALUES ($1::text::uuid, 'Legacy Org')`, orgA)
		exec(`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership')`, orgA, userA)
		exec(`INSERT INTO app.sessions (id, user_id, kind, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at) VALUES
  ($1::text::uuid, $4::text::uuid, 'browser', $6, sha256($6), now(), now() + interval '1 hour', now() + interval '1 day'),
  ($2::text::uuid, $5::text::uuid, 'browser', $7, sha256($7), now(), now() + interval '1 hour', now() + interval '1 day'),
  ($3::text::uuid, $5::text::uuid, 'browser', $8, sha256($8), now() - interval '2 days', now() - interval '1 day 1 hour',
   now() - interval '1 day')`, legacyRevoke, legacyExpire, legacyDead, userA, userB, hash(legacyRevoke), hash(legacyExpire), hash(legacyDead))
		rowsHash := func() string {
			t.Helper()
			var s string
			if err := setup.QueryRow(ctx, `SELECT count(*)::text || ':' || md5(string_agg(format('%s|%s|%s|%s|%s|%s|%s|%s|%s', id, user_id, kind,
  encode(credential_hash, 'hex'), encode(csrf_token, 'hex'), revoked_at, revoke_reason, idle_expires_at, absolute_expires_at), ',' ORDER BY id))
FROM app.sessions`).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
		before := rowsHash()

		r12 := runnerFor(t, cfg, subsetFS(t, 12))
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r12.Up(ctx, b) }); out !=
			"applied\t00012_session_binding_and_claim_exclusivity.sql\ncurrent_version\t12\n" {
			t.Fatalf("up to 12: %q", out)
		}
		metadataThroughVersion12Only(t, cfg, r12)
		s12 := snapshot(t, r12)
		if got := rowsHash(); got != before {
			t.Fatal("legacy session rows changed by the upgrade")
		}
		var bound int
		if err := setup.QueryRow(ctx, `SELECT count(*) FROM app.sessions WHERE origin IS NOT NULL OR organization_id IS NOT NULL`).Scan(&bound); err != nil || bound != 0 {
			t.Fatalf("the upgrade bound %d legacy rows (%v)", bound, err)
		}

		rt := rlsRuntimeConn(t, cfg)
		resolves := func(id, origin string) int {
			t.Helper()
			var n int
			if err := rt.QueryRow(ctx, `SELECT count(*) FROM app.auth_resolve_session($1, now(), $2)`, hash(id), origin).Scan(&n); err != nil {
				t.Fatalf("bound reader: %v", err)
			}
			return n
		}
		for _, origin := range []string{"https://org-a.localhost:543", "https://portal.example.test"} {
			for _, id := range []string{legacyRevoke, legacyExpire} {
				if n := resolves(id, origin); n != 0 {
					t.Errorf("live legacy session resolves at an origin after version 12 (%d rows)", n)
				}
			}
		}
		down := func() (string, error) {
			var b bytes.Buffer
			err := r12.Down(ctx, &b)
			return b.String(), err
		}
		if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00012: live legacy browser sessions would become usable by the global reader") {
			t.Fatalf("down with live legacy sessions: %v %q", err, out)
		}
		if got := snapshot(t, r12); got != s12 {
			t.Fatal("the refused down changed the catalog")
		}

		// Runtime revocation (the store.go revokeAll statement) on a legacy row: allowed, audited once.
		var events0, events1 int
		countRevoked := func() int {
			t.Helper()
			var n int
			if err := setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'session.revoked'`).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		events0 = countRevoked()
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		rlsContext(t, tx, userA, "", "", "")
		tag, err := tx.Exec(ctx, `UPDATE app.sessions SET revoked_at = $2, revoke_reason = $3 WHERE user_id = $1 AND revoked_at IS NULL`,
			userA, time.Now(), "logout_all")
		if err != nil || tag.RowsAffected() != 1 {
			_ = tx.Rollback(ctx)
			t.Fatalf("runtime revocation of a legacy row: %v (%d rows)", err, tag.RowsAffected())
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if events1 = countRevoked(); events1-events0 != 1 {
			t.Errorf("session.revoked events for the legacy revocation: %d, want 1", events1-events0)
		}

		// Restore-time invalidation (infra/recovery/sql/invalidate.sql, its sessions UPDATE verbatim) as the superuser.
		invalidate := `UPDATE app.sessions SET idle_expires_at = least(idle_expires_at, now()), absolute_expires_at = least(absolute_expires_at, now()),
         csrf_token = uuid_send(gen_random_uuid()) || uuid_send(gen_random_uuid())
  WHERE revoked_at IS NULL RETURNING 1`
		if b64 := os.Getenv("MCP_INVALIDATE_SQL_B64"); b64 != "" {
			src, err := base64.StdEncoding.DecodeString(b64)
			if err != nil || !strings.Contains(normalize(string(src)), normalize(invalidate)) {
				t.Fatalf("the invalidation statement differs from infra/recovery/sql/invalidate.sql (%v)", err)
			}
		} else {
			t.Log("MCP_INVALIDATE_SQL_B64 unset: the invalidation statement is the cited copy")
		}
		tag, err = setup.Exec(ctx, invalidate)
		if err != nil || tag.RowsAffected() != 2 {
			t.Fatalf("restore-time invalidation of the legacy rows: %v (%d rows), want the two unrevoked rows", err, tag.RowsAffected())
		}
		var live, revoked int
		if err := setup.QueryRow(ctx, `SELECT count(*) FILTER (WHERE revoked_at IS NULL AND idle_expires_at > now() AND absolute_expires_at > now()),
  count(*) FILTER (WHERE revoked_at IS NOT NULL) FROM app.sessions WHERE id = ANY($1::text[]::uuid[])`,
			[]string{legacyRevoke, legacyExpire, legacyDead}).Scan(&live, &revoked); err != nil || live != 0 || revoked != 1 {
			t.Fatalf("after invalidation: %d live, %d revoked (%v), want 0 live and only the runtime revocation", live, revoked, err)
		}

		// Runtime negatives at version 12, each rolled back: an unbound browser row is refused by the policy, an unbound
		// mcp row in its TENANT-ORG shape is admitted, a bound browser row is admitted and resolves only at its own
		// origin, and the binding columns cannot be changed.
		probe := func(id, user, org, stmt string, args ...any) error {
			t.Helper()
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			rlsContext(t, tx, user, org, "", "")
			_, err = tx.Exec(ctx, stmt, args...)
			return err
		}
		insert := `INSERT INTO app.sessions (user_id, kind, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at, origin,
  organization_id) VALUES ($1::text::uuid, $2, sha256(gen_random_uuid()::text::bytea), sha256(gen_random_uuid()::text::bytea), now(),
  now() + interval '1 hour', now() + interval '1 day', $3, $4::text::uuid)`
		if err := probe("unbound browser", userA, "", insert, userA, "browser", nil, nil); sqlstateOf(err) != "42501" {
			t.Errorf("runtime INSERT of an unbound browser row: %v, want 42501", err)
		}
		if err := probe("origin without organization", userA, "", insert, userA, "browser", "https://org-a.localhost:543", nil); err == nil {
			t.Error("runtime INSERT of a half-bound browser row was admitted")
		}
		if err := probe("unbound mcp", userA, orgA, insert, userA, "mcp", nil, nil); err != nil {
			t.Errorf("runtime INSERT of an unbound mcp row in its TENANT-ORG shape: %v", err)
		}
		if err := probe("bound mcp", userA, orgA, insert, userA, "mcp", "https://org-a.localhost:543", orgA); err == nil {
			t.Error("runtime INSERT of a bound mcp row was admitted")
		}
		{
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			rlsContext(t, tx, userA, "", "", "")
			cred := hash("bound-control")
			if _, err := tx.Exec(ctx, `INSERT INTO app.sessions (user_id, kind, credential_hash, csrf_token, auth_time, idle_expires_at,
  absolute_expires_at, origin, organization_id) VALUES ($1::text::uuid, 'browser', $2, sha256($2), now(), now() + interval '1 hour',
  now() + interval '1 day', 'https://org-a.localhost:543', $3::text::uuid)`, userA, cred, orgA); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("runtime INSERT of a bound browser row: %v", err)
			}
			var own, other int
			var gotOrigin, gotOrg string
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.auth_resolve_session($1, now(), 'https://org-b.localhost:543')`, cred).Scan(&other); err != nil {
				t.Fatal(err)
			}
			if err := tx.QueryRow(ctx, `SELECT count(*), min(origin), min(organization_id::text) FROM app.auth_resolve_session($1, now(), 'https://org-a.localhost:543')`,
				cred).Scan(&own, &gotOrigin, &gotOrg); err != nil {
				t.Fatal(err)
			}
			if own != 1 || other != 0 || gotOrigin != "https://org-a.localhost:543" || gotOrg != orgA {
				t.Errorf("bound control: own origin %d rows (%s, %s), other origin %d rows", own, gotOrigin, gotOrg, other)
			}
			if _, err := tx.Exec(ctx, `UPDATE app.sessions SET origin = 'https://org-b.localhost:543' WHERE credential_hash = $1`, cred); sqlstateOf(err) != "42501" {
				t.Errorf("runtime UPDATE of the binding origin: %v, want 42501", err)
			}
			_ = tx.Rollback(ctx)
		}
		if err := probe("binding organization", userA, "", `UPDATE app.sessions SET organization_id = NULL WHERE id = $1::text::uuid`, legacyRevoke); sqlstateOf(err) != "42501" {
			t.Errorf("runtime UPDATE of the binding organization: %v, want 42501", err)
		}

		// No live legacy row remains: the down is allowed, restores the exact version-11 catalog and revives nothing.
		out, err := down()
		if err != nil || out != "rolled_back\t00012_session_binding_and_claim_exclusivity.sql\ncurrent_version\t11\n" {
			t.Fatalf("down after revocation and invalidation: %v %q", err, out)
		}
		metadataThroughVersion11Only(t, cfg, r11)
		if v := domViolations(t, cat, clean); len(v) != 0 {
			t.Fatalf("version-11 detector after the 00012 down: %v", v)
		}
		for _, id := range []string{legacyRevoke, legacyExpire, legacyDead} {
			var n int
			if err := rt.QueryRow(ctx, `SELECT count(*) FROM app.auth_resolve_session($1, now())`, hash(id)).Scan(&n); err != nil || n != 0 {
				t.Errorf("the restored global reader revives a legacy row (%d rows, %v)", n, err)
			}
		}
		if out := run(t, "up", func(ctx context.Context, b *bytes.Buffer) error { return r12.Up(ctx, b) }); out !=
			"applied\t00012_session_binding_and_claim_exclusivity.sql\ncurrent_version\t12\n" {
			t.Fatalf("up again: %q", out)
		}
		if got := snapshot(t, r12); got != s12 {
			t.Fatalf("catalog after down/up differs:\n%s\n---\n%s", got, s12)
		}

		// The remaining refusals: two non-removed rows of one hostname (pending claims coexist at version 12), then a
		// bound session (binding history is never discarded). Each refused down leaves the catalog unchanged.
		exec(`INSERT INTO app.organization_domains (id, organization_id, hostname, generation, challenge_value, challenge_expires_at, created_by)
VALUES (gen_random_uuid(), $1::text::uuid, 'shared.legacy.example', 11, repeat('e', 52), now() + interval '1 hour', $2::text::uuid),
       (gen_random_uuid(), $1::text::uuid, 'shared.legacy.example', 12, repeat('f', 52), now() + interval '1 hour', $2::text::uuid)`, orgA, userA)
		if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00012: a hostname has more than one non-removed row") {
			t.Fatalf("down with two pending claims of one hostname: %v %q", err, out)
		}
		exec(`INSERT INTO app.sessions (user_id, kind, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at, origin,
  organization_id) VALUES ($1::text::uuid, 'browser', $2, sha256($2), now(), now() + interval '1 hour', now() + interval '1 day',
  'https://org-a.localhost:543', $3::text::uuid)`, userA, hash("bound-history"), orgA)
		if out, err := down(); err == nil || !strings.Contains(err.Error(), "refusing to roll back 00012: bound sessions exist (binding history is never discarded)") {
			t.Fatalf("down with a bound session: %v %q", err, out)
		}
		if got := snapshot(t, r12); got != s12 {
			t.Fatal("a refused down changed the catalog")
		}
		reset(t, cfg)
	})
}
