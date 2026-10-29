//go:build integration

package identity_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// B-IDENTITY-CATALOG (Plan r5 §6 / r2 §6.1; Spec R9/AC9): the exact identity objects, owner, grants, constraints and
// runtime denials on a real PostgreSQL migrated to the reviewed set. From version 6 (G008 RC-4) this is the final
// identity and authentication contract: row-level security on users, sessions, identity_principals and
// login_transactions (not on the public provider binding), the narrowed runtime table grants, runtime CRUD only inside
// its context shape, constraint controls through the attested setup connection, and runtime denials.

var identityTables = []string{"identity_principals", "identity_provider_binding", "login_transactions", "sessions", "users"}

func sqlstate(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func expectState(t *testing.T, conn *pgx.Conn, what, want, stmt string, args ...any) {
	t.Helper()
	ctx := context.Background()
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, stmt, args...)
	if got := sqlstate(err); got != want {
		t.Errorf("%s: want SQLSTATE %s, got %q (%v)", what, want, got, err)
	}
}

func TestIdentityCatalog(t *testing.T) {
	resetIdentity(t)
	ctx := context.Background()
	own := ownerConn(t)
	setup := setupConn(t)
	rt := runtimeConn(t)

	t.Run("objects and owner", func(t *testing.T) {
		var tables, owners string
		var rls int
		err := own.QueryRow(ctx, `
SELECT string_agg(c.relname, ',' ORDER BY c.relname), string_agg(DISTINCT c.relowner::regrole::text, ','),
       count(*) FILTER (WHERE c.relrowsecurity OR c.relforcerowsecurity)::int
FROM pg_class c
WHERE c.relnamespace = 'app'::regnamespace AND c.relkind = 'r' AND c.relname <> 'goose_db_version'`).Scan(&tables, &owners, &rls)
		if err != nil {
			t.Fatal(err)
		}
		// From version 3 the G005 tenancy tables (groups, organizations, role_bindings) and from version 4 the G006
		// BusinessProfile tables (business_profiles, group_business_profiles) share the schema; the set stays exact.
		// Their own catalogs are B-TENANCY-CATALOG and B-BUSINESS-PROFILE-CATALOG. From version 6 every table except the
		// provider binding and the migration metadata has row-level security (B-RLS-CATALOG names them exactly). From
		// version 7 the G009 stored_objects table joins the set, forced as well (B-OBJ-CATALOG). From version 10 the G014
		// mcp_connections table joins the set, forced as well (B-MCP-CATALOG). From version 11 the G016
		// organization_domains table joins the set, forced as well (TestDomainsCatalog). From version 13 the G017
		// organization_branding table joins the set, forced as well (TestBrandingCatalog). From version 15 the G019
		// invitations table joins the set, forced as well (TestMembershipCatalog), and from version 16 the G058
		// enrollment_requests, enrollment_restrictions and enrollment_sessions tables, forced as well (TestEnrollmentCatalog),
		// and from version 18 the G062 signup_sessions table, forced as well (TestSignupCatalog), and from version 20 the
		// G020 label_assignments and label_definitions tables, forced as well (TestLabelsCatalog).
		want := "business_profiles,enrollment_requests,enrollment_restrictions,enrollment_sessions,group_business_profiles,groups," +
			"identity_principals,identity_provider_binding,invitations,label_assignments,label_definitions,login_transactions," +
			"mcp_connections,organization_branding," +
			"organization_domains,organizations,role_bindings,sessions,signup_sessions,stored_objects,users"
		if tables != want || owners != it.cfg.OwnerRole() || rls != 20 {
			t.Fatalf("tables=%q owners=%q rls=%d", tables, owners, rls)
		}
		// No tenant, organization, group, role or vendor-organization column. The two exemptions are exactly
		// login_transactions.organization_id (G016 amendment 02): the Organization a tenant-host login was pinned to, a
		// routing fact re-checked at issuance, not a tenant or membership attribute of the identity; and, from version 12
		// (G016 security amendment A2, amendment 08), sessions.organization_id: the Organization of the tenant origin a
		// browser session is bound to (with sessions.origin, which this pattern does not match), likewise a routing
		// binding re-checked on every resolution, not a membership or authority attribute.
		var bad string
		if err := own.QueryRow(ctx, `
SELECT coalesce(string_agg(table_name || '.' || column_name, ','), '') FROM information_schema.columns
WHERE table_schema = 'app' AND table_name = ANY($1)
  AND column_name ~ '(organization|org_|tenant|group|membership|role|permission|entitlement)'
  AND NOT (table_name = 'login_transactions' AND column_name = 'organization_id')
  AND NOT (table_name = 'sessions' AND column_name = 'organization_id')`, identityTables).Scan(&bad); err != nil || bad != "" {
			t.Fatalf("forbidden columns: %q %v", bad, err)
		}
	})

	t.Run("exact runtime grants and no other grantee", func(t *testing.T) {
		var grants, others string
		err := own.QueryRow(ctx, `
SELECT coalesce(string_agg(table_name || ':' || privilege_type, ',' ORDER BY table_name, privilege_type) FILTER (WHERE grantee = $2), ''),
       coalesce(string_agg(grantee || '@' || table_name, ',') FILTER (WHERE grantee NOT IN ($2, $3)), '')
FROM information_schema.role_table_grants WHERE table_schema = 'app' AND table_name = ANY($1)`,
			identityTables, it.cfg.RuntimeRole(), it.cfg.OwnerRole()).Scan(&grants, &others)
		if err != nil {
			t.Fatal(err)
		}
		// Version 6: table-level grants only (users and the session UPDATE columns are column grants, B-RLS-CATALOG).
		want := "identity_principals:INSERT,identity_principals:SELECT," +
			"identity_provider_binding:INSERT,identity_provider_binding:SELECT," +
			"login_transactions:DELETE,login_transactions:INSERT,login_transactions:SELECT," +
			"sessions:INSERT,sessions:SELECT"
		if grants != want || others != "" {
			t.Fatalf("runtime grants %q (want %q); other grantees %q", grants, want, others)
		}
	})

	// Seed one valid user, principal and session as the runtime role: CRUD within its grants works inside its context
	// shape (IDENTITY for the subject, LOGIN for the binding), never outside it. From version 12 a browser session must
	// carry its tenant binding (origin and Organization; 00012 sessions_insert), so the fixture Organization org-a is
	// created first through the attested setup connection (seedTenants) and the session is bound to its origin.
	seedTenants(t)
	uid := "4d6f6e69-746f-4172-9a43-61746c6f6731"
	h32 := strings.Repeat("a", 32)
	identityTx := func(stmts ...func(pgx.Tx) error) {
		t.Helper()
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true)`, uid); err != nil {
			t.Fatal(err)
		}
		for _, f := range stmts {
			if err := f(tx); err != nil {
				t.Fatalf("runtime write in IDENTITY shape: %v", err)
			}
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	identityTx(
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO app.users (id, display_name) VALUES ($1, 'Catalog')`, uid)
			return err
		},
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE app.users SET display_name = 'Catalog 2' WHERE id = $1`, uid)
			return err
		},
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('https://iss.example', 'user_01CAT', $1)`, uid)
			return err
		},
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
INSERT INTO app.sessions (user_id, credential_hash, csrf_token, provider_session_id, auth_time, idle_expires_at, absolute_expires_at, provider_logout_ref_hash, provider_logout_ref_expires_at, origin, organization_id)
VALUES ($1, convert_to($2, 'UTF8'), convert_to($2, 'UTF8'), 'sid', now(), now() + interval '30 minutes', now() + interval '8 hours', convert_to($3, 'UTF8'), now() + interval '10 minutes', $4, $5::text::uuid)`,
				uid, h32, strings.Repeat("r", 32), originA, orgAID)
			return err
		},
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE app.sessions SET revoked_at = now(), revoke_reason = 'provider_logout' WHERE user_id = $1`, uid)
			return err
		},
	)
	// Outside a context the same runtime role reads and writes nothing (zero rows, WITH CHECK 42501).
	var visible int
	if err := rt.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.users) + (SELECT count(*) FROM app.sessions) + (SELECT count(*) FROM app.identity_principals)`).Scan(&visible); err != nil || visible != 0 {
		t.Fatalf("runtime without context sees %d identity rows (%v)", visible, err)
	}
	expectState(t, rt, "runtime INSERT users without context", "42501", `INSERT INTO app.users (id, display_name) VALUES ('5d6f6e69-746f-4172-9a43-61746c6f6731', 'X')`)
	if _, err := rt.Exec(ctx, `INSERT INTO app.identity_provider_binding (provider, issuer, client_id) VALUES ('workos', 'https://iss.example', 'client_01X')`); err != nil {
		t.Fatalf("runtime INSERT binding: %v", err)
	}
	loginTx := func(stmt string, args ...any) int64 {
		t.Helper()
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if _, err := tx.Exec(ctx, `SELECT set_config('app.login_binding', encode(convert_to($1, 'UTF8'), 'hex'), true)`, h32); err != nil {
			t.Fatal(err)
		}
		tag, err := tx.Exec(ctx, stmt, args...)
		if err != nil {
			t.Fatalf("runtime write in LOGIN shape: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		return tag.RowsAffected()
	}
	loginTx(`
INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, created_at, expires_at)
VALUES (convert_to($1, 'UTF8'), convert_to($1, 'UTF8'), $2, 'client_01X', 'https://o', '/', now(), now() + interval '10 minutes')`, h32, strings.Repeat("v", 43))
	if n := loginTx(`DELETE FROM app.login_transactions`); n != 1 {
		t.Fatalf("runtime DELETE login_transactions in its own binding: %d rows", n)
	}

	t.Run("constraint negatives", func(t *testing.T) {
		missing := "00000000-0000-0000-0000-000000000001"
		cases := []struct{ name, state, stmt string }{
			{"principal for a missing user", "23503", `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('i', 'user_x', '` + missing + `')`},
			{"session for a missing user", "23503", `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at) VALUES ('` + missing + `', convert_to(repeat('b',32),'UTF8'), convert_to(repeat('b',32),'UTF8'), now(), now(), now())`},
			{"delete a referenced user (RESTRICT)", "23503", `DELETE FROM app.users WHERE id = '` + uid + `'`},
			{"duplicate (issuer, subject)", "23505", `INSERT INTO app.identity_principals (issuer, subject, user_id) SELECT 'https://iss.example', 'user_01CAT', id FROM app.users WHERE id = '` + uid + `'`},
			{"second principal for a user", "23505", `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('https://iss.example', 'user_01OTHER', '` + uid + `')`},
			{"duplicate credential hash", "23505", `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at) VALUES ('` + uid + `', convert_to(repeat('a',32),'UTF8'), convert_to(repeat('c',32),'UTF8'), now(), now(), now())`},
			{"duplicate logout reference", "23505", `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at, provider_logout_ref_hash, provider_logout_ref_expires_at, revoked_at, revoke_reason) VALUES ('` + uid + `', convert_to(repeat('d',32),'UTF8'), convert_to(repeat('d',32),'UTF8'), now(), now(), now(), convert_to(repeat('r',32),'UTF8'), now(), now(), 'provider_logout')`},
			{"second binding row", "23505", `INSERT INTO app.identity_provider_binding (provider, issuer, client_id) VALUES ('workos', 'https://other', 'client_01Y')`},
			{"non-workos provider", "23514", `INSERT INTO app.identity_provider_binding (provider, issuer, client_id) VALUES ('oidc', 'https://other', 'client_01Y')`},
			{"empty issuer", "23514", `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('', 'user_x', '` + uid + `')`},
			{"oversize issuer", "23514", `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES (repeat('i', 2049), 'user_x', '` + uid + `')`},
			{"oversize subject", "23514", `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('i', 'user_' || repeat('x', 251), '` + uid + `')`},
			{"non user_ subject", "23514", `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('i', 'admin', '` + uid + `')`},
			{"subject with a LIKE wildcard prefix", "23514", `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('i', 'userX1', '` + uid + `')`},
			{"bad status", "23514", `UPDATE app.users SET status = 'root' WHERE id = '` + uid + `'`},
			{"disabled without disabled_at", "23514", `UPDATE app.users SET status = 'disabled' WHERE id = '` + uid + `'`},
			{"oversize display name", "23514", `UPDATE app.users SET display_name = repeat('n', 201) WHERE id = '` + uid + `'`},
			{"assurance other than unknown", "23514", `UPDATE app.sessions SET mfa_assurance = 'verified'`},
			{"bad revoke reason", "23514", `UPDATE app.sessions SET revoke_reason = 'expired'`},
			{"revoked without reason", "23514", `UPDATE app.sessions SET revoke_reason = NULL`},
			{"provider outcome without provider logout", "23514", `UPDATE app.sessions SET revoke_reason = 'logout', provider_revocation = 'confirmed'`},
			{"short credential hash", "23514", `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at) VALUES ('` + uid + `', convert_to('short','UTF8'), convert_to(repeat('e',32),'UTF8'), now(), now(), now())`},
			{"short csrf token", "23514", `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at) VALUES ('` + uid + `', convert_to(repeat('f',32),'UTF8'), convert_to('short','UTF8'), now(), now(), now())`},
			{"idle beyond absolute", "23514", `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at) VALUES ('` + uid + `', convert_to(repeat('g',32),'UTF8'), convert_to(repeat('g',32),'UTF8'), now(), now() + interval '2 hours', now() + interval '1 hour')`},
			{"bad return path", "23514", `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, created_at, expires_at) VALUES (convert_to(repeat('h',32),'UTF8'), convert_to(repeat('h',32),'UTF8'), repeat('v',43), 'c', 'o', '//evil', now(), now() + interval '1 minute')`},
			{"transaction TTL over 10 minutes", "23514", `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, created_at, expires_at) VALUES (convert_to(repeat('i',32),'UTF8'), convert_to(repeat('i',32),'UTF8'), repeat('v',43), 'c', 'o', '/', now(), now() + interval '11 minutes')`},
			{"short verifier", "23514", `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, created_at, expires_at) VALUES (convert_to(repeat('j',32),'UTF8'), convert_to(repeat('j',32),'UTF8'), 'v', 'c', 'o', '/', now(), now() + interval '1 minute')`},
			{"short state hash", "23514", `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, created_at, expires_at) VALUES (convert_to('s','UTF8'), convert_to(repeat('k',32),'UTF8'), repeat('v',43), 'c', 'o', '/', now(), now() + interval '1 minute')`},
		}
		// Constraints hold independently of row-level security: checked through the attested setup connection (the
		// owner is bound by FORCE and would see a policy violation first), every statement rolled back.
		for _, c := range cases {
			expectState(t, setup, c.name, c.state, c.stmt)
		}
	})

	t.Run("runtime denials", func(t *testing.T) {
		for _, c := range []struct{ name, stmt string }{
			{"CREATE TABLE in app", `CREATE TABLE app.runtime_ddl (id int)`},
			{"ALTER TABLE", `ALTER TABLE app.users ADD COLUMN x int`},
			{"DROP TABLE", `DROP TABLE app.sessions`},
			{"TRUNCATE", `TRUNCATE app.login_transactions`},
			{"DELETE users", `DELETE FROM app.users`},
			{"DELETE principals", `DELETE FROM app.identity_principals`},
			{"DELETE sessions", `DELETE FROM app.sessions`},
			{"DELETE binding", `DELETE FROM app.identity_provider_binding`},
			{"UPDATE binding", `UPDATE app.identity_provider_binding SET issuer = 'https://evil'`},
			{"SELECT migration metadata", `SELECT 1 FROM app.goose_db_version`},
			{"SET ROLE owner", `SET ROLE ` + pgx.Identifier{it.cfg.OwnerRole()}.Sanitize()},
			{"SET ROLE migrator", `SET ROLE ` + pgx.Identifier{it.cfg.MigratorRole()}.Sanitize()},
			{"SET ROLE policy reader", `SET ROLE ` + pgx.Identifier{it.cfg.Database + "_policy_reader"}.Sanitize()},
			{"SET ROLE auth reader", `SET ROLE ` + pgx.Identifier{it.cfg.Database + "_auth_reader"}.Sanitize()},
			{"UPDATE principals", `UPDATE app.identity_principals SET subject = 'user_01EVIL'`},
			{"UPDATE users created_at", `UPDATE app.users SET created_at = now()`},
			{"SELECT users created_at", `SELECT created_at FROM app.users`},
			{"UPDATE sessions credential", `UPDATE app.sessions SET credential_hash = credential_hash`},
		} {
			expectState(t, rt, c.name, "42501", c.stmt)
		}
	})

	resetIdentity(t)
}
