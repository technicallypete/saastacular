//go:build integration

// B-TENANCY-IT: the tenancy model against a disposable PostgreSQL initialized by the rendered role script and migrated
// here to the reviewed set (infra/validation/run.sh, suite "migrations"). Every model operation runs through a pool
// logged in as the actual runtime role; owner-level cases use the migrator with SET ROLE <db>_owner. Assertions are
// scoped to this run's fixture rows, compare row-set hashes before and after each rejected or failed operation, and
// name the intended SQLSTATE and constraint or error code. Fails, never skips, when settings are absent.
//
// Required environment: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, APP_ENV,
// TENANCY_CANARY_FILE (per-run canaries, one per line) and TENANCY_LOG_FILE (where the operation log is written).
package tenancy_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/migrate"
	"core/internal/observability"
	"core/internal/tenancy"
	"core/migrations"
)

type env struct {
	cfg      migrate.Config
	pw       string
	pool     *pgxpool.Pool
	own      *pgx.Conn // migrator with SET ROLE owner
	rt       *pgx.Conn // runtime
	logs     *syncBuf
	svc      *tenancy.Service
	canaries []string
}

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.b.Write(p) }
func (s *syncBuf) String() string              { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

func connect(t *testing.T, cfg migrate.Config, user, pw string) *pgx.Conn {
	t.Helper()
	cc, err := pgx.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	cc.Host, cc.Port, cc.Database, cc.User, cc.Password = cfg.Host, uint16(cfg.Port), cfg.Database, user, pw
	cc.ConnectTimeout = 5 * time.Second
	cc.TLSConfig, cc.Fallbacks = nil, nil
	c, err := pgx.ConnectConfig(context.Background(), cc)
	if err != nil {
		t.Fatalf("connect as %s: %v", user, err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })
	return c
}

func (e *env) ownerConn(t *testing.T) *pgx.Conn {
	t.Helper()
	c := connect(t, e.cfg, e.cfg.MigratorRole(), e.cfg.Password)
	if _, err := c.Exec(context.Background(), "SET ROLE "+pgx.Identifier{e.cfg.OwnerRole()}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	return c
}

func (e *env) runtimeConn(t *testing.T) *pgx.Conn {
	t.Helper()
	return connect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
}

func setup(t *testing.T) *env {
	t.Helper()
	ctx := context.Background()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil {
		t.Fatalf("integration settings: %v", err)
	}
	e := &env{cfg: cfg, pw: os.Getenv("DB_RUNTIME_PASSWORD")}
	canaryFile, logFile := os.Getenv("TENANCY_CANARY_FILE"), os.Getenv("TENANCY_LOG_FILE")
	if e.pw == "" || canaryFile == "" || logFile == "" {
		t.Fatal("integration settings: DB_RUNTIME_PASSWORD, TENANCY_CANARY_FILE and TENANCY_LOG_FILE are required")
	}
	b, err := os.ReadFile(canaryFile)
	if err != nil {
		t.Fatalf("canary file: %v", err)
	}
	e.canaries = strings.Fields(string(b))
	if len(e.canaries) < 3 {
		t.Fatal("canary file: at least three canaries are required")
	}
	// G008 RC-7: the model primitives are trusted setup paths whose runtime grants version 6 removes; these tests keep
	// their accepted meaning at version 5. A database left at version 6 by an earlier test is rolled back one step (its
	// rows are kept), then the version-5 set is applied. From version 7 (G009 Plan r1 §5.6) a database at version 7 is
	// first rolled back to 6 (its object table is empty: the object tests clean their rows), then to 5. From version 8
	// (G011 Plan §10.2) the disposable database's audit history is purged first through the attested superuser (a
	// test-only privileged action: the append-only refusal trigger is disabled around one TRUNCATE), then 8 steps down
	// to 7. From version 9 (G013 Plan r3 §11.3) 9 first steps down to 8 (00009 adds privileges and policies only; its
	// down needs no purge). From version 10 (G014 Plan r8 §0.2) the disposable MCP state and audit history are purged
	// through the attested superuser (00010's down refuses while either exists), then 10 steps down to 9. From version 11
	// (G016) the disposable domain state (domains, slugs, tenant-host login transactions) and audit history are purged
	// through the attested superuser (00011's down refuses while any exists), then 11 steps down to 10. From version 12
	// (G016 security amendment) the disposable browser sessions and domain rows are purged through the attested
	// superuser (00012's down refuses bound or live legacy sessions and duplicate hostnames), then 12 steps down to 11.
	// From version 13 (G017) the disposable branding rows and audit history are purged through the attested superuser
	// (00013's down refuses while any branding row or branding history exists), then 13 steps down to 12. From version
	// 14 (G018) the audit history is purged through the attested superuser (00014's down refuses while any audit row
	// carries a G018 operation), then 14 steps down to 13. From version 15 (G019) the disposable invitations and audit
	// history are purged through the attested superuser (00015's down refuses while any invitation or invitation-era
	// audit row exists), then 15 steps down to 14. From version 16 (G058) the disposable enrollment credentials and
	// restriction ledger rows and the audit history are purged through the attested superuser (00016's down refuses while
	// any credential, ledger row, enrollment-shaped login transaction or 00016-reason/operation audit row exists), then 16
	// steps down to 15.
	full, err := migrate.Open(ctx, cfg, migrations.FS)
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	var out bytes.Buffer
	if err := full.Version(ctx, &out); err != nil {
		t.Fatalf("migrate version: %v", err)
	}
	// G064 (Plan r2 §3.4): the observed starting version, before the ladder. The harness starts every invocation on a
	// fresh copy at the embedded set's top version, and this ladder leaves a database at 5; any other start fails closed.
	// The top version is derived from the embedded migrations, never a literal.
	embedded, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil || len(embedded) == 0 {
		t.Fatalf("embedded migrations: %v (%d files)", err, len(embedded))
	}
	slices.Sort(embedded)
	topVersion, _, _ := strings.Cut(embedded[len(embedded)-1], "_")
	topVersion = strings.TrimLeft(topVersion, "0")
	startVersion := "unknown"
	if i := strings.LastIndex(out.String(), "current_version\t"); i >= 0 {
		startVersion = strings.TrimSpace(out.String()[i+len("current_version\t"):])
	}
	t.Logf("TENANCY-SETUP version=%s", startVersion)
	if startVersion != topVersion && startVersion != "5" {
		t.Fatalf("TENANCY-SETUP: starting version %s is neither the embedded top version %s nor 5", startVersion, topVersion)
	}
	if strings.HasSuffix(out.String(), "current_version\t22\n") {
		// G071: 00022 adds only the default-tenant reader function; its down needs no data cleanup, then 22 steps down to 21.
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t21\n") {
			t.Fatalf("migrate down to 21: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t21\n") {
		// G057: 00021's down refuses while a live or re-mintable signup continuation references a platform label: the
		// disposable signup rows are purged through the attested superuser, then 21 steps down to 20 (the version-18 branch
		// below repeats the signup purge).
		purgeDomainState(t, cfg, `DELETE FROM app.signup_sessions; `)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t20\n") {
			t.Fatalf("migrate down to 20: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t20\n") {
		// G020: 00020's down refuses while any label definition or assignment exists or audit.events holds a label action,
		// operation, reason or resource: the disposable label rows (assignments first: the definition key is RESTRICT) and
		// the audit history are purged through the attested superuser, then 20 steps down to 19 (the version-19 branch
		// below continues to 18; its domain purge repeats).
		purgeDomainState(t, cfg, `DELETE FROM app.label_assignments; DELETE FROM app.label_definitions; `)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t19\n") {
			t.Fatalf("migrate down to 19: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t19\n") {
		// G063: 00019's down refuses while any connection has a friendly name or audit.events holds an
		// mcp.connection_renamed event: the names are cleared and the audit history purged through the attested
		// superuser, then 19 steps down to 18 (the version-18 branch below continues to 17; its domain purge repeats).
		purgeDomainState(t, cfg, `UPDATE app.mcp_connections SET display_name = NULL WHERE display_name IS NOT NULL; `)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t18\n") {
			t.Fatalf("migrate down to 18: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t18\n") {
		// G062: 00018's down refuses while app.signup_sessions holds any row, app.login_transactions holds a signup-shaped
		// or /signup/details row, or audit.events holds an 00018 reason or operation: the disposable signup rows and the
		// audit history are purged through the attested superuser, then 18 steps down to 17 (the version-17 branch below
		// continues to 16).
		purgeDomainState(t, cfg, `DELETE FROM app.signup_sessions; DELETE FROM app.login_transactions WHERE signup_intent OR signup_handoff `+
			`OR return_path = '/signup/details'; `)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t17\n") {
			t.Fatalf("migrate down to 17: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t17\n") {
		// G060: 00017's down refuses while any Organization has the Group-management override enabled or audit.events
		// holds an 00017 action, reason or operation: the override is reset and the audit history purged through the
		// attested superuser, then 17 steps down to 16 (the version-16 branch below continues to 15).
		purgeDomainState(t, cfg, `UPDATE app.organizations SET group_management_override = false WHERE group_management_override; `)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t16\n") {
			t.Fatalf("migrate down to 16: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t16\n") {
		// The ledger references the requests (source_request_id): it goes before them. The request events go with the
		// audit purge below.
		purgeDomainState(t, cfg, `DELETE FROM app.enrollment_sessions; DELETE FROM app.enrollment_restrictions; DELETE FROM app.enrollment_requests; `+
			`DELETE FROM app.login_transactions WHERE enrollment_intent OR invitation_id IS NOT NULL OR request_on_arrival `+
			`OR return_path = '/join/status'; `)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t15\n") {
			t.Fatalf("migrate down to 15: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t15\n") {
		purgeDomainState(t, cfg, `DELETE FROM app.invitations; `)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t14\n") {
			t.Fatalf("migrate down to 14: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t14\n") {
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t13\n") {
			t.Fatalf("migrate down to 13: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t13\n") {
		purgeDomainState(t, cfg, `DELETE FROM app.organization_branding; `)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t12\n") {
			t.Fatalf("migrate down to 12: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t12\n") {
		purgeDomainState(t, cfg, `DELETE FROM app.sessions WHERE kind = 'browser'; `)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t11\n") {
			t.Fatalf("migrate down to 11: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t11\n") {
		purgeDomainState(t, cfg)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t10\n") {
			t.Fatalf("migrate down to 10: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t10\n") {
		purgeMCPState(t, cfg)
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t9\n") {
			t.Fatalf("migrate down to 9: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t9\n") {
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t8\n") {
			t.Fatalf("migrate down to 8: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t8\n") {
		purgeAuditHistory(t, cfg)
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t7\n") {
			t.Fatalf("migrate down to 7: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t7\n") {
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t6\n") {
			t.Fatalf("migrate down to 6: %v %q", err, out.String())
		}
	}
	if strings.HasSuffix(out.String(), "current_version\t6\n") {
		out.Reset()
		if err := full.Down(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t5\n") {
			t.Fatalf("migrate down to 5: %v %q", err, out.String())
		}
	}
	_ = full.Close()
	v5 := fstest.MapFS{}
	names, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		if n < "00006" {
			b, err := fs.ReadFile(migrations.FS, n)
			if err != nil {
				t.Fatal(err)
			}
			v5[n] = &fstest.MapFile{Data: b}
		}
	}
	r, err := migrate.Open(ctx, cfg, v5)
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	out.Reset()
	if err := r.Up(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t5\n") {
		t.Fatalf("migrate up: %v %q", err, out.String())
	}
	_ = r.Close()
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = cfg.Host, uint16(cfg.Port), cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = cfg.RuntimeRole(), e.pw
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	pc.MaxConns = 48
	if e.pool, err = pgxpool.NewWithConfig(ctx, pc); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.pool.Close)
	e.own, e.rt = e.ownerConn(t), e.runtimeConn(t)
	e.logs = &syncBuf{}
	e.svc = tenancy.NewService(e.pool, observability.NewLogger(e.logs, slog.LevelInfo, nil))
	t.Cleanup(func() {
		// World-readable: it holds bounded codes only, and the nonroot log checker reads it.
		if err := os.WriteFile(logFile, []byte(e.logs.String()), 0o644); err != nil {
			t.Errorf("write the operation log: %v", err)
		}
	})
	return e
}

// --- helpers ---------------------------------------------------------------------------------------------------------

// purgeAuditHistory empties audit.events of the disposable test database through the attested superuser
// (DB_SUPERUSER), so that 00008's down, which refuses while any row exists, can run. Never a production path.
func purgeAuditHistory(t *testing.T, cfg migrate.Config) {
	t.Helper()
	ctx := context.Background()
	user, pw := os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD")
	if user == "" || pw == "" {
		t.Fatal("integration settings: DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required to step down from version 8")
	}
	c := connect(t, cfg, user, pw)
	defer func() { _ = c.Close(ctx) }() // the ladder from the top version opens 19 purge connections; none outlives its helper
	var super bool
	if err := c.QueryRow(ctx, `SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&super); err != nil || !super {
		t.Fatalf("purge connection is not a superuser (%v)", err)
	}
	if _, err := c.Exec(ctx, `BEGIN; ALTER TABLE audit.events DISABLE TRIGGER events_refuse_truncate; TRUNCATE audit.events; `+
		`ALTER TABLE audit.events ENABLE TRIGGER events_refuse_truncate; COMMIT`); err != nil {
		t.Fatalf("purge the disposable audit history: %v", err)
	}
}

// purgeMCPState removes every MCP connection and mcp session of the disposable database through the attested superuser
// (a test-only privileged action; row-level security does not apply to a superuser), so 00010's down can run.
func purgeMCPState(t *testing.T, cfg migrate.Config) {
	t.Helper()
	ctx := context.Background()
	user, pw := os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD")
	if user == "" || pw == "" {
		t.Fatal("integration settings: DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required to step down from version 10")
	}
	c := connect(t, cfg, user, pw)
	defer func() { _ = c.Close(ctx) }()
	var super bool
	if err := c.QueryRow(ctx, `SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&super); err != nil || !super {
		t.Fatalf("purge connection is not a superuser (%v)", err)
	}
	if _, err := c.Exec(ctx, `BEGIN; DELETE FROM app.mcp_connections; DELETE FROM app.sessions WHERE kind = 'mcp'; COMMIT`); err != nil {
		t.Fatalf("purge the disposable MCP state: %v", err)
	}
}

// purgeDomainState removes every organization domain, Organization slug and tenant-host login transaction of the
// disposable database through the attested superuser (a test-only privileged action; row-level security does not apply
// to a superuser), so 00011's down can run once the audit history is purged too; extra statements (the version-12
// browser-session purge) run first in the same transaction.
func purgeDomainState(t *testing.T, cfg migrate.Config, extra ...string) {
	t.Helper()
	ctx := context.Background()
	user, pw := os.Getenv("DB_SUPERUSER"), os.Getenv("DB_SUPERUSER_PASSWORD")
	if user == "" || pw == "" {
		t.Fatal("integration settings: DB_SUPERUSER and DB_SUPERUSER_PASSWORD are required to step down from version 11")
	}
	c := connect(t, cfg, user, pw)
	defer func() { _ = c.Close(ctx) }()
	var super bool
	if err := c.QueryRow(ctx, `SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&super); err != nil || !super {
		t.Fatalf("purge connection is not a superuser (%v)", err)
	}
	if _, err := c.Exec(ctx, `BEGIN; `+strings.Join(extra, "")+`DELETE FROM app.organization_domains; UPDATE app.organizations SET slug = NULL WHERE slug IS NOT NULL; `+
		`DELETE FROM app.login_transactions WHERE organization_id IS NOT NULL OR callback_origin IS NOT NULL; COMMIT`); err != nil {
		t.Fatalf("purge the disposable domain state: %v", err)
	}
}

func sqlErr(err error) (code, constraint string) {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return pg.Code, pg.ConstraintName
	}
	return "", ""
}

func wantCode(t *testing.T, what string, err error, want tenancy.Code) {
	t.Helper()
	if got := tenancy.CodeOf(err); got != want {
		t.Errorf("%s: %v, want %s", what, err, want)
	}
}

// expectSQL runs stmt in its own transaction, rolls it back and asserts the SQLSTATE and, when given, the constraint.
func expectSQL(t *testing.T, conn *pgx.Conn, what, state, constraint, stmt string, args ...any) {
	t.Helper()
	ctx := context.Background()
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, stmt, args...)
	code, con := sqlErr(err)
	if code != state || (constraint != "" && con != constraint) {
		t.Errorf("%s: SQLSTATE %q constraint %q (%v), want %s %s", what, code, con, err, state, constraint)
	}
}

// rows is the fixture row-set snapshot: Organizations, their Groups and their bindings, each as an md5 over the
// ordered row texts.
func (e *env) rows(t *testing.T, orgs ...tenancy.OrganizationID) string {
	t.Helper()
	ids := make([]string, len(orgs))
	for i, o := range orgs {
		ids[i] = o.String()
	}
	var s string
	err := e.own.QueryRow(context.Background(), `
SELECT coalesce((SELECT md5(string_agg(o::text, E'\n' ORDER BY o.id)) FROM app.organizations o WHERE o.id::text = ANY ($1::text[])), '-') || '|' ||
       coalesce((SELECT md5(string_agg(g::text, E'\n' ORDER BY g.id)) FROM app.groups g WHERE g.organization_id::text = ANY ($1::text[])), '-') || '|' ||
       coalesce((SELECT md5(string_agg(b::text, E'\n' ORDER BY b.id)) FROM app.role_bindings b WHERE b.organization_id::text = ANY ($1::text[])), '-')`,
		ids).Scan(&s)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func (e *env) count(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := e.own.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	return n
}

var userSeq atomic.Int64

// user seeds one synthetic global user through runtime SQL, with a principal and a live session unless bare.
func (e *env) user(t *testing.T, bare bool) tenancy.UserID {
	t.Helper()
	ctx := context.Background()
	var id string
	n := userSeq.Add(1)
	if err := e.rt.QueryRow(ctx, `INSERT INTO app.users (display_name) VALUES ($1) RETURNING id::text`, fmt.Sprintf("Tenancy fixture %d", n)).Scan(&id); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if !bare {
		if _, err := e.rt.Exec(ctx, `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('https://issuer.example.invalid', $1, $2::text::uuid)`,
			"user_t"+id[:8]+fmt.Sprint(n), id); err != nil {
			t.Fatalf("seed principal: %v", err)
		}
		if _, err := e.rt.Exec(ctx, `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at)
VALUES ($1::text::uuid, sha256(convert_to($2::text || 'c', 'UTF8')), sha256(convert_to($2::text || 'x', 'UTF8')), now(), now() + interval '30 minutes', now() + interval '8 hours')`, id, id); err != nil {
			t.Fatalf("seed session: %v", err)
		}
	}
	u, err := tenancy.ParseUserID(id)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func (e *env) org(t *testing.T, name string) tenancy.Organization {
	t.Helper()
	o, err := e.svc.CreateOrganization(context.Background(), name)
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	return o
}

func (e *env) group(t *testing.T, org tenancy.OrganizationID, creator tenancy.UserID, name, slug string) (tenancy.Group, tenancy.Binding) {
	t.Helper()
	g, b, err := e.svc.CreateGroup(context.Background(), org, creator, name, slug)
	if err != nil {
		t.Fatalf("create group %s: %v", slug, err)
	}
	return g, b
}

// bindingRow reads one binding's columns as text (owner connection).
func (e *env) bindingRow(t *testing.T, id tenancy.BindingID) string {
	t.Helper()
	var s string
	if err := e.own.QueryRow(context.Background(), `SELECT organization_id::text || ' ' || subject_user_id::text || ' ' || scope_type || ' ' ||
  scope_id::text || ' ' || coalesce(group_id::text, '-') || ' ' || role || ' ' || source FROM app.role_bindings WHERE id = $1::text::uuid`, id.String()).Scan(&s); err != nil {
		t.Fatalf("binding %s: %v", id, err)
	}
	return s
}

// waitBlocked waits until another backend of this database waits for a lock: an ungranted pg_locks entry of a
// pg_stat_activity backend of the current database (both are readable by the runtime role; other roles' wait events in
// pg_stat_activity are not).
func (e *env) waitBlocked(t *testing.T) {
	t.Helper()
	obs := e.runtimeConn(t)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		var n int
		if err := obs.QueryRow(context.Background(), `SELECT count(*) FROM pg_locks l JOIN pg_stat_activity a ON a.pid = l.pid
WHERE NOT l.granted AND a.datname = current_database() AND a.pid <> pg_backend_pid()`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no lock wait observed")
}

func isV4(id [16]byte) bool { return id[6]>>4 == 4 && id[8]>>6 == 2 }

// h is what a failure message prints for a value that may carry a name: a short hash prefix, never the value.
func h(v any) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%+v", v))))[:12] }

// --- fault injection ---------------------------------------------------------------------------------------------------

// faultDB wraps the runtime pool. Hooks run after a statement whose SQL starts with a prefix; failures replace a
// statement or the commit. Every injected failure is a real transaction abort: the wrapper rolls back.
type faultDB struct {
	pool       *pgxpool.Pool
	failOn     string        // statement prefix whose execution is replaced by a connection failure
	failCommit bool          // commit is replaced by a rollback and an error
	after      string        // statement prefix after which hook runs
	hook       func()        // run after the statement matching after
	cancel     func()        // cancels the operation's context after the statement matching after
	closed     *pgxpool.Pool // when set, a closed pool is used instead
}

func (f *faultDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	if f.closed != nil {
		return f.closed.BeginTx(ctx, o)
	}
	tx, err := f.pool.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}
	return &faultTx{Tx: tx, f: f}, nil
}

type faultTx struct {
	pgx.Tx
	f *faultDB
}

var errInjected = &pgconn.PgError{Code: "08006", Message: "injected connection failure"}

type errRow struct{ err error }

func (r errRow) Scan(...any) error { return r.err }

type afterRow struct {
	row  pgx.Row
	then func()
}

func (r afterRow) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	r.then()
	return err
}

func (t *faultTx) afterHook(sql string) func() {
	if t.f.after == "" || !strings.HasPrefix(sql, t.f.after) {
		return func() {}
	}
	return func() {
		if t.f.hook != nil {
			t.f.hook()
		}
		if t.f.cancel != nil {
			t.f.cancel()
		}
	}
}

func (t *faultTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if t.f.failOn != "" && strings.HasPrefix(sql, t.f.failOn) {
		return errRow{errInjected}
	}
	return afterRow{row: t.Tx.QueryRow(ctx, sql, args...), then: t.afterHook(sql)}
}

func (t *faultTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if t.f.failOn != "" && strings.HasPrefix(sql, t.f.failOn) {
		return pgconn.CommandTag{}, errInjected
	}
	tag, err := t.Tx.Exec(ctx, sql, args...)
	t.afterHook(sql)()
	return tag, err
}

func (t *faultTx) Commit(ctx context.Context) error {
	if t.f.failCommit {
		_ = t.Tx.Rollback(ctx)
		return errors.New("injected commit failure")
	}
	return t.Tx.Commit(ctx)
}

// --- the model test ------------------------------------------------------------------------------------------------------

func TestTenancyModel(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	svc := e.svc

	t.Run("runtime identity", func(t *testing.T) {
		var cur, sess string
		var super, bypass bool
		var memberships int
		err := e.pool.QueryRow(ctx, `SELECT current_user::text, session_user::text, r.rolsuper, r.rolbypassrls,
  (SELECT count(*) FROM pg_auth_members m WHERE m.member = r.oid)::int FROM pg_roles r WHERE r.rolname = current_user`).Scan(&cur, &sess, &super, &bypass, &memberships)
		if err != nil {
			t.Fatal(err)
		}
		if cur != e.cfg.RuntimeRole() || sess != cur || super || bypass || memberships != 0 {
			t.Fatalf("model pool: current=%s session=%s super=%v bypassrls=%v memberships=%d", cur, sess, super, bypass, memberships)
		}
	})

	// Shared fixtures: creator F, Organizations A and B, Groups A/g1 and A/g2, and B/g3 with g1's slug.
	F := e.user(t, false)
	A, B := e.org(t, "Fixture A"), e.org(t, "Fixture B")
	g1, g1Owner := e.group(t, A.ID, F, "Acme Revenue", "")
	g2, _ := e.group(t, A.ID, F, "Acme Operations", "acme-operations")
	g3, _ := e.group(t, B.ID, F, "Acme Revenue", "acme-revenue")

	t.Run("IT-ID-1 identities and creator ownership", func(t *testing.T) {
		ids := map[[16]byte]bool{}
		for _, id := range [][16]byte{A.ID, B.ID, g1.ID, g2.ID, g3.ID, g1Owner.ID} {
			if id == [16]byte{} || !isV4(id) || ids[id] {
				t.Fatalf("not a distinct nonzero v4 UUID: %x", id)
			}
			ids[id] = true
		}
		if g1.OrganizationID != A.ID || g2.OrganizationID != A.ID || g3.OrganizationID != B.ID || g1.Slug != "acme-revenue" || g3.Slug != g1.Slug {
			t.Fatalf("groups: %s %s %s", h(g1), h(g2), h(g3))
		}
		for _, g := range []tenancy.Group{g1, g2, g3} {
			ms, err := svc.ListGroupMemberships(ctx, g.OrganizationID, g.ID)
			if err != nil || len(ms) != 1 || ms[0].UserID != F || ms[0].Role != tenancy.GroupOwner {
				t.Fatalf("group %s memberships: %+v %v", g.Slug, ms, err)
			}
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE subject_user_id = $1::text::uuid`, F.String()); n != 3 {
			t.Fatalf("creator bindings: %d", n)
		}
		if _, err := svc.GetOrganizationMembership(ctx, A.ID, F); tenancy.CodeOf(err) != tenancy.CodeMembershipNotFound {
			t.Fatalf("creator Organization membership: %v", err)
		}
	})

	t.Run("IT-ID-2 renames and slug edit keep identity", func(t *testing.T) {
		o := e.org(t, "Rename Organization")
		g, owner := e.group(t, o.ID, F, "Rename Group", "rename-group")
		before := e.bindingRow(t, owner.ID)
		o2, err := svc.RenameOrganization(ctx, o.ID, "Renamed Organization")
		if err != nil || o2.ID != o.ID || !o2.CreatedAt.Equal(o.CreatedAt) || o2.Name != "Renamed Organization" {
			t.Fatalf("rename organization: %s %v", h(o2), err)
		}
		r, err := svc.RenameGroup(ctx, o.ID, g.ID, "Renamed Group")
		if err != nil || r.ID != g.ID || r.OrganizationID != o.ID || r.Slug != "rename-group" || !r.CreatedAt.Equal(g.CreatedAt) {
			t.Fatalf("rename group (slug must not change): %s %v", h(r), err)
		}
		c, err := svc.ChangeGroupSlug(ctx, o.ID, g.ID, "renamed-group")
		if err != nil || c.ID != g.ID || c.OrganizationID != o.ID || c.Name != "Renamed Group" || c.Slug != "renamed-group" {
			t.Fatalf("change slug: %s %v", h(c), err)
		}
		if same, err := svc.ChangeGroupSlug(ctx, o.ID, g.ID, "renamed-group"); err != nil || same.Slug != "renamed-group" {
			t.Fatalf("same slug is a no-op success: %v", err)
		}
		if got := e.bindingRow(t, owner.ID); got != before {
			t.Fatalf("binding changed: %s -> %s", before, got)
		}
	})

	t.Run("IT-ID-3 runtime cannot change or supply keys", func(t *testing.T) {
		o := e.org(t, "Key Organization")
		g, b := e.group(t, o.ID, F, "Key Group", "key-group")
		before := e.rows(t, o.ID, B.ID)
		for what, stmt := range map[string]string{
			"groups.organization_id":        `UPDATE app.groups SET organization_id = $1::text::uuid WHERE id = $2::text::uuid`,
			"groups.id":                     `UPDATE app.groups SET id = gen_random_uuid() WHERE id = $2::text::uuid AND $1::text IS NOT NULL`,
			"organizations.id":              `UPDATE app.organizations SET id = gen_random_uuid() WHERE id = $1::text::uuid AND $2::text IS NOT NULL`,
			"organizations.created_at":      `UPDATE app.organizations SET created_at = now() WHERE id = $1::text::uuid AND $2::text IS NOT NULL`,
			"role_bindings.id":              `UPDATE app.role_bindings SET id = gen_random_uuid() WHERE id = $2::text::uuid AND $1::text IS NOT NULL`,
			"role_bindings.organization_id": `UPDATE app.role_bindings SET organization_id = $1::text::uuid WHERE id = $2::text::uuid`,
			"role_bindings.subject_user_id": `UPDATE app.role_bindings SET subject_user_id = subject_user_id WHERE id = $2::text::uuid AND $1::text IS NOT NULL`,
			"role_bindings.scope_type":      `UPDATE app.role_bindings SET scope_type = 'organization' WHERE id = $2::text::uuid AND $1::text IS NOT NULL`,
			"role_bindings.scope_id":        `UPDATE app.role_bindings SET scope_id = $1::text::uuid WHERE id = $2::text::uuid`,
			"role_bindings.group_id":        `UPDATE app.role_bindings SET group_id = NULL WHERE id = $2::text::uuid AND $1::text IS NOT NULL`,
			"role_bindings.source":          `UPDATE app.role_bindings SET source = 'membership' WHERE id = $2::text::uuid AND $1::text IS NOT NULL`,
			"organizations explicit id":     `INSERT INTO app.organizations (id, name) VALUES (gen_random_uuid(), 'x' || $1::text || $2::text)`,
			"organizations explicit time":   `INSERT INTO app.organizations (name, created_at) VALUES ('x' || $1::text || $2::text, now())`,
			"groups explicit id":            `INSERT INTO app.groups (id, organization_id, name, slug) VALUES (gen_random_uuid(), $1::text::uuid, 'x', 'x' || length($2::text))`,
			"groups explicit time":          `INSERT INTO app.groups (organization_id, name, slug, created_at) VALUES ($1::text::uuid, 'x', 'x' || length($2::text), now())`,
			"role_bindings explicit id": `INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, role, source)
SELECT gen_random_uuid(), organization_id, subject_user_id, 'organization', organization_id, 'org_member', 'membership' FROM app.role_bindings WHERE id = $2::text::uuid AND $1::text IS NOT NULL`,
			"role_bindings explicit time": `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source, created_at)
SELECT organization_id, subject_user_id, 'organization', organization_id, 'org_member', 'membership', now() FROM app.role_bindings WHERE id = $2::text::uuid AND $1::text IS NOT NULL`,
		} {
			arg2 := g.ID.String()
			if strings.HasPrefix(what, "role_bindings") {
				arg2 = b.ID.String()
			}
			if strings.HasPrefix(what, "organizations.") {
				expectSQL(t, e.rt, what, "42501", "", stmt, o.ID.String(), arg2)
				continue
			}
			expectSQL(t, e.rt, what, "42501", "", stmt, B.ID.String(), arg2)
		}
		if after := e.rows(t, o.ID, B.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-ID-4 owner key changes hit the immutability triggers", func(t *testing.T) {
		o := e.org(t, "Trigger Organization")
		g, b := e.group(t, o.ID, F, "Trigger Group", "trigger-group")
		before := e.rows(t, o.ID)
		og, ob := g.ID.String(), b.ID.String()
		for what, c := range map[string]struct{ constraint, stmt, arg string }{
			"organizations.id":              {"organizations_keys_immutable", `UPDATE app.organizations SET id = gen_random_uuid() WHERE id = $1::text::uuid`, o.ID.String()},
			"groups.id":                     {"groups_keys_immutable", `UPDATE app.groups SET id = gen_random_uuid() WHERE id = $1::text::uuid`, og},
			"groups.organization_id":        {"groups_keys_immutable", `UPDATE app.groups SET organization_id = '` + B.ID.String() + `' WHERE id = $1::text::uuid`, og},
			"role_bindings.id":              {"role_bindings_keys_immutable", `UPDATE app.role_bindings SET id = gen_random_uuid() WHERE id = $1::text::uuid`, ob},
			"role_bindings.organization_id": {"role_bindings_keys_immutable", `UPDATE app.role_bindings SET organization_id = '` + B.ID.String() + `' WHERE id = $1::text::uuid`, ob},
			"role_bindings.subject_user_id": {"role_bindings_keys_immutable", `UPDATE app.role_bindings SET subject_user_id = (SELECT id FROM app.users WHERE id <> subject_user_id LIMIT 1) WHERE id = $1::text::uuid`, ob},
			"role_bindings.scope_type":      {"role_bindings_keys_immutable", `UPDATE app.role_bindings SET scope_type = 'organization' WHERE id = $1::text::uuid`, ob},
			"role_bindings.scope_id":        {"role_bindings_keys_immutable", `UPDATE app.role_bindings SET scope_id = organization_id WHERE id = $1::text::uuid`, ob},
			"role_bindings.group_id":        {"role_bindings_keys_immutable", `UPDATE app.role_bindings SET group_id = NULL WHERE id = $1::text::uuid`, ob},
			"role_bindings.source":          {"role_bindings_keys_immutable", `UPDATE app.role_bindings SET source = 'invitation' WHERE id = $1::text::uuid`, ob},
		} {
			expectSQL(t, e.own, what, "23000", c.constraint, c.stmt, c.arg)
		}
		if after := e.rows(t, o.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-ID-5 zero UUIDs are refused", func(t *testing.T) {
		const zero = "00000000-0000-0000-0000-000000000000"
		u := e.user(t, true)
		expectSQL(t, e.own, "organization id", "23514", "organizations_id_nonzero", `INSERT INTO app.organizations (id, name) VALUES ($1::text::uuid, 'Zero')`, zero)
		expectSQL(t, e.own, "group id", "23514", "groups_id_nonzero", `INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1::text::uuid, $2::text::uuid, 'Zero', 'zero-id')`, zero, A.ID.String())
		expectSQL(t, e.own, "binding id", "23514", "role_bindings_id_nonzero", `INSERT INTO app.role_bindings (id, organization_id, subject_user_id, scope_type, scope_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid, 'organization', $2::text::uuid, 'org_member', 'membership')`, zero, A.ID.String(), u.String())
		expectSQL(t, e.own, "subject", "23514", "role_bindings_subject_nonzero", `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, 'org_member', 'membership')`, A.ID.String(), zero)
		expectSQL(t, e.rt, "runtime subject", "23514", "role_bindings_subject_nonzero", `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, 'org_member', 'membership')`, A.ID.String(), zero)
	})

	t.Run("IT-REF-1 runtime grants of every role", func(t *testing.T) {
		o := e.org(t, "Grant Organization")
		g, _ := e.group(t, o.ID, F, "Grant Group", "grant-group")
		for _, r := range []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin, tenancy.OrgMember} {
			u := e.user(t, false)
			b, err := svc.GrantOrganizationRole(ctx, o.ID, u, r)
			if err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("%s %s organization %s - %s membership", o.ID, u, o.ID, r)
			if got := e.bindingRow(t, b.ID); got != want || b.GroupID != nil || b.ScopeType != tenancy.ScopeOrganization || b.Role != string(r) {
				t.Fatalf("organization %s binding: %s (returned %+v)", r, got, b)
			}
		}
		for _, r := range []tenancy.GroupRole{tenancy.GroupOwner, tenancy.GroupAdmin, tenancy.GroupMember} {
			u := e.user(t, false)
			b, err := svc.GrantGroupRole(ctx, o.ID, g.ID, u, r)
			if err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("%s %s group %s %s %s membership", o.ID, u, g.ID, g.ID, r)
			if got := e.bindingRow(t, b.ID); got != want || b.GroupID == nil || *b.GroupID != g.ID {
				t.Fatalf("group %s binding: %s", r, got)
			}
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid`, o.ID.String()); n != 7 {
			t.Fatalf("bindings: %d, want 6 grants plus the creator", n)
		}
	})

	t.Run("IT-REF-2 unknown, foreign and zero selectors", func(t *testing.T) {
		u := e.user(t, false)
		before := e.rows(t, A.ID, B.ID)
		unknownOrg, _ := tenancy.ParseOrganizationID("7f000000-0000-4000-8000-000000000001")
		unknownGroup, _ := tenancy.ParseGroupID("7f000000-0000-4000-8000-000000000002")
		unknownUser, _ := tenancy.ParseUserID("7f000000-0000-4000-8000-000000000003")
		_, e1 := svc.GrantOrganizationRole(ctx, unknownOrg, u, tenancy.OrgMember)
		wantCode(t, "unknown organization", e1, tenancy.CodeOrganizationNotFound)
		_, e2 := svc.GrantGroupRole(ctx, A.ID, unknownGroup, u, tenancy.GroupMember)
		_, e3 := svc.GrantGroupRole(ctx, A.ID, g3.ID, u, tenancy.GroupMember) // B's Group with Organization A
		wantCode(t, "unknown group", e2, tenancy.CodeGroupNotFound)
		wantCode(t, "foreign group", e3, tenancy.CodeGroupNotFound)
		if e2 == nil || e3 == nil || e2.Error() != e3.Error() {
			t.Fatalf("unknown and foreign errors differ: %v / %v", e2, e3)
		}
		_, e4 := svc.GrantOrganizationRole(ctx, A.ID, unknownUser, tenancy.OrgMember)
		wantCode(t, "unknown user", e4, tenancy.CodeUserNotFound)
		_, e5 := svc.GrantGroupRole(ctx, A.ID, g1.ID, unknownUser, tenancy.GroupMember)
		wantCode(t, "unknown user (group)", e5, tenancy.CodeUserNotFound)
		_, e6 := svc.GrantOrganizationRole(ctx, A.ID, tenancy.UserID{}, tenancy.OrgMember)
		wantCode(t, "zero user", e6, tenancy.CodeInvalidID)
		_, e7 := tenancy.ParseUserID("not-a-uuid")
		wantCode(t, "malformed user", e7, tenancy.CodeInvalidID)
		_, e8 := svc.GetGroup(ctx, A.ID, unknownGroup)
		_, e9 := svc.GetGroup(ctx, A.ID, g3.ID)
		if tenancy.CodeOf(e8) != tenancy.CodeGroupNotFound || e8.Error() != e9.Error() {
			t.Fatalf("get unknown/foreign group: %v / %v", e8, e9)
		}
		if after := e.rows(t, A.ID, B.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-REF-3 runtime direct SQL negatives", func(t *testing.T) {
		u := e.user(t, true)
		unknown := "7f000000-0000-4000-8000-000000000009"
		before := e.rows(t, A.ID, B.ID)
		ins := `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES `
		a, gB, g := A.ID.String(), g3.ID.String(), g1.ID.String()
		expectSQL(t, e.rt, "unknown subject", "23503", "role_bindings_subject_fkey",
			ins+`($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership')`, a, unknown)
		expectSQL(t, e.rt, "unknown organization", "23503", "role_bindings_organization_fkey",
			ins+`($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership')`, unknown, u.String())
		expectSQL(t, e.rt, "group of B with Organization A", "23503", "role_bindings_group_fkey",
			ins+`($1::text::uuid, $2::text::uuid, 'group', $3::text::uuid, $3::text::uuid, 'group_member', 'membership')`, a, u.String(), gB)
		expectSQL(t, e.rt, "Organization scope with group_id", "23514", "role_bindings_scope_consistent",
			ins+`($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, $3::text::uuid, 'org_member', 'membership')`, a, u.String(), g)
		expectSQL(t, e.rt, "Group scope without group_id", "23514", "role_bindings_scope_consistent",
			ins+`($1::text::uuid, $2::text::uuid, 'group', $3::text::uuid, NULL, 'group_member', 'membership')`, a, u.String(), g)
		expectSQL(t, e.rt, "Organization scope_id differs", "23514", "role_bindings_scope_consistent",
			ins+`($1::text::uuid, $2::text::uuid, 'organization', $3::text::uuid, NULL, 'org_member', 'membership')`, a, u.String(), B.ID.String())
		expectSQL(t, e.rt, "Group scope_id differs", "23514", "role_bindings_scope_consistent",
			ins+`($1::text::uuid, $2::text::uuid, 'group', $3::text::uuid, $4::text::uuid, 'group_member', 'membership')`, a, u.String(), g2.ID.String(), g)
		expectSQL(t, e.rt, "zero subject", "23514", "role_bindings_subject_nonzero",
			ins+`($1::text::uuid, '00000000-0000-0000-0000-000000000000', 'organization', $1::text::uuid, NULL, 'org_member', 'membership')`, a)
		// A referenced user's key change fails on the subject foreign key (G004 grants runtime table-level UPDATE on users).
		bare := e.user(t, true)
		if _, err := svc.GrantOrganizationRole(ctx, A.ID, bare, tenancy.OrgMember); err != nil {
			t.Fatal(err)
		}
		before = e.rows(t, A.ID, B.ID)
		expectSQL(t, e.rt, "referenced user key change", "23503", "role_bindings_subject_fkey",
			`UPDATE app.users SET id = gen_random_uuid() WHERE id = $1::text::uuid`, bare.String())
		if after := e.rows(t, A.ID, B.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-REF-4 parent deletes race grants", func(t *testing.T) {
		// Deterministic orders first: (1) an uncommitted owner delete blocks the grant, which then fails; (2) an
		// uncommitted runtime insert blocks the owner delete, which then fails.
		type target struct {
			name   string
			parent func(t *testing.T) (del string, arg string, grant func() error, insert func(tx pgx.Tx) error)
		}
		mk := func(t *testing.T, kind string) (string, string, func() error, func(tx pgx.Tx) error) {
			u := e.user(t, true)
			switch kind {
			case "user":
				return `DELETE FROM app.users WHERE id = $1::text::uuid`, u.String(),
					func() error { _, err := svc.GrantOrganizationRole(ctx, A.ID, u, tenancy.OrgMember); return err },
					func(tx pgx.Tx) error {
						_, err := tx.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, 'org_member', 'membership')`, A.ID.String(), u.String())
						return err
					}
			case "organization":
				o := e.org(t, "Race Organization")
				return `DELETE FROM app.organizations WHERE id = $1::text::uuid`, o.ID.String(),
					func() error { _, err := svc.GrantOrganizationRole(ctx, o.ID, u, tenancy.OrgMember); return err },
					func(tx pgx.Tx) error {
						_, err := tx.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, 'org_member', 'membership')`, o.ID.String(), u.String())
						return err
					}
			default: // group: its creator binding is removed first so that only the race decides
				g, _ := e.group(t, A.ID, F, "Race Group", fmt.Sprintf("race-%d", userSeq.Add(1)))
				if err := svc.RemoveGroupMembership(ctx, A.ID, g.ID, F); err != nil {
					t.Fatal(err)
				}
				return `DELETE FROM app.groups WHERE id = $1::text::uuid`, g.ID.String(),
					func() error { _, err := svc.GrantGroupRole(ctx, A.ID, g.ID, u, tenancy.GroupMember); return err },
					func(tx pgx.Tx) error {
						_, err := tx.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'group', $3::text::uuid, $3::text::uuid, 'group_member', 'membership')`, A.ID.String(), u.String(), g.ID.String())
						return err
					}
			}
		}
		notFound := map[string]tenancy.Code{"user": tenancy.CodeUserNotFound, "organization": tenancy.CodeOrganizationNotFound, "group": tenancy.CodeGroupNotFound}
		fkName := map[string]string{"user": "role_bindings_subject_fkey", "organization": "role_bindings_organization_fkey", "group": "role_bindings_group_fkey"}
		dangling := `SELECT count(*) FROM app.role_bindings b WHERE NOT EXISTS (SELECT 1 FROM app.users u WHERE u.id = b.subject_user_id)
  OR NOT EXISTS (SELECT 1 FROM app.organizations o WHERE o.id = b.organization_id)
  OR (b.group_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM app.groups g WHERE g.id = b.group_id AND g.organization_id = b.organization_id))`
		for _, kind := range []string{"user", "organization", "group"} {
			// Order 1: the delete commits while the grant waits.
			del, arg, grant, _ := mk(t, kind)
			own := e.ownerConn(t)
			tx, err := own.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, del, arg); err != nil {
				t.Fatalf("%s delete: %v", kind, err)
			}
			done := make(chan error, 1)
			go func() { done <- grant() }()
			e.waitBlocked(t)
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			wantCode(t, kind+" grant after a committed delete", <-done, notFound[kind])
			// Order 2: the grant's insert commits while the delete waits.
			del, arg, _, insert := mk(t, kind)
			rtx, err := e.runtimeConn(t).Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if err := insert(rtx); err != nil {
				t.Fatalf("%s insert: %v", kind, err)
			}
			res := make(chan error, 1)
			go func() { _, err := own.Exec(ctx, del, arg); res <- err }()
			e.waitBlocked(t)
			if err := rtx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if code, con := sqlErr(<-res); code != "23503" || con != fkName[kind] {
				t.Errorf("%s delete after a committed grant: %s %s", kind, code, con)
			}
		}
		// 50 randomized rounds: exactly one side wins, never a dangling row.
		for i := range 50 {
			kind := []string{"user", "organization", "group"}[i%3]
			del, arg, grant, _ := mk(t, kind)
			own := e.ownerConn(t)
			var gErr, dErr error
			var wg sync.WaitGroup
			wg.Add(2)
			go func() { defer wg.Done(); time.Sleep(time.Duration(rand.IntN(3)) * time.Millisecond); gErr = grant() }()
			go func() {
				defer wg.Done()
				time.Sleep(time.Duration(rand.IntN(3)) * time.Millisecond)
				_, dErr = own.Exec(ctx, del, arg)
			}()
			wg.Wait()
			dCode, _ := sqlErr(dErr)
			switch {
			case gErr == nil && dCode == "23503":
			case tenancy.CodeOf(gErr) == notFound[kind] && dErr == nil:
			default:
				t.Fatalf("round %d (%s): grant %v, delete %v", i, kind, gErr, dErr)
			}
		}
		if n := e.count(t, dangling); n != 0 {
			t.Fatalf("dangling bindings: %d", n)
		}
	})

	t.Run("IT-VOC-1 the closed scope, role and source matrix", func(t *testing.T) {
		u := e.user(t, true)
		valid := map[string]bool{}
		for _, r := range []string{"org_owner", "org_admin", "org_member"} {
			valid["organization|"+r+"|membership"] = true
		}
		for _, r := range []string{"group_owner", "group_admin", "group_member"} {
			valid["group|"+r+"|membership"] = true
		}
		roles := []string{"org_owner", "org_admin", "org_member", "group_owner", "group_admin", "group_member", "owner", "admin", "member",
			"be_owner", "be_admin", "business_profile_owner", "advisor", "principal", "associate", ""}
		inserted := 0
		for _, scope := range []string{"organization", "group", "business_profile", "business_entity", "", "Organization"} {
			for _, role := range roles {
				for _, source := range []string{"membership", "direct_grant", "invitation", "system", ""} {
					groupID, scopeID := "NULLIF($4::text, $4::text)::uuid", "$1::text::uuid"
					if scope == "group" {
						groupID, scopeID = "$4::text::uuid", "$4::text::uuid"
					}
					stmt := `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, $3, ` + scopeID + `, ` + groupID + `, $5, $6)`
					key := scope + "|" + role + "|" + source
					tx, err := e.rt.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					_, err = tx.Exec(ctx, stmt, A.ID.String(), u.String(), scope, g1.ID.String(), role, source)
					_ = tx.Rollback(ctx)
					code, con := sqlErr(err)
					switch {
					case valid[key] && err == nil:
						inserted++
					case !valid[key] && code == "23514" && con == "role_bindings_scope_consistent":
					default:
						t.Errorf("%s: %v", key, err)
					}
				}
			}
		}
		if inserted != 6 {
			t.Fatalf("valid combinations inserted: %d, want 6", inserted)
		}
		for _, r := range []string{"group_owner", "advisor", "principal", "associate", "be_owner", "business_profile_owner"} {
			_, err := svc.GrantOrganizationRole(ctx, A.ID, u, tenancy.OrganizationRole(r))
			wantCode(t, "service organization role "+r, err, tenancy.CodeInvalidRole)
		}
		for _, r := range []string{"org_owner", "advisor", "principal", "associate", "be_admin"} {
			_, err := svc.GrantGroupRole(ctx, A.ID, g1.ID, u, tenancy.GroupRole(r))
			wantCode(t, "service group role "+r, err, tenancy.CodeInvalidRole)
		}
	})

	t.Run("IT-UNIQ-1 concurrent duplicate grants", func(t *testing.T) {
		o := e.org(t, "Unique Organization")
		g, _ := e.group(t, o.ID, F, "Unique Group", "unique-group")
		u := e.user(t, false)
		for round := range 20 {
			for _, scope := range []string{"organization", "group"} {
				var ok, exists, other atomic.Int64
				var wg sync.WaitGroup
				for range 32 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						var err error
						if scope == "organization" {
							_, err = svc.GrantOrganizationRole(ctx, o.ID, u, tenancy.OrgMember)
						} else {
							_, err = svc.GrantGroupRole(ctx, o.ID, g.ID, u, tenancy.GroupMember)
						}
						switch tenancy.CodeOf(err) {
						case "":
							ok.Add(1)
						case tenancy.CodeMembershipExists:
							exists.Add(1)
						default:
							other.Add(1)
						}
					}()
				}
				wg.Wait()
				n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid AND scope_type = $3`, o.ID.String(), u.String(), scope)
				if ok.Load() != 1 || exists.Load() != 31 || other.Load() != 0 || n != 1 {
					t.Fatalf("round %d %s: ok=%d exists=%d other=%d rows=%d", round, scope, ok.Load(), exists.Load(), other.Load(), n)
				}
				var err error
				if scope == "organization" {
					err = svc.RemoveOrganizationMembership(ctx, o.ID, u)
				} else {
					err = svc.RemoveGroupMembership(ctx, o.ID, g.ID, u)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		// The creator already holds its Group's one binding.
		_, err := svc.GrantGroupRole(ctx, o.ID, g.ID, F, tenancy.GroupMember)
		wantCode(t, "creator regrant", err, tenancy.CodeMembershipExists)
	})

	t.Run("IT-UNIQ-2 deterministic duplicate insert", func(t *testing.T) {
		u := e.user(t, true)
		ins := `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, 'org_member', 'membership')`
		txA, err := e.runtimeConn(t).Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := txA.Exec(ctx, ins, B.ID.String(), u.String()); err != nil {
			t.Fatal(err)
		}
		connB := e.runtimeConn(t)
		res := make(chan error, 1)
		go func() { _, err := connB.Exec(ctx, ins, B.ID.String(), u.String()); res <- err }()
		e.waitBlocked(t)
		if err := txA.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if code, con := sqlErr(<-res); code != "23505" || con != "role_bindings_subject_scope_key" {
			t.Fatalf("second insert: %s %s", code, con)
		}
	})

	t.Run("IT-PROJ-1 grant, change, remove and regrant through one binding", func(t *testing.T) {
		o := e.org(t, "Projection Organization")
		g, owner := e.group(t, o.ID, F, "Projection Group", "projection-group")
		ownerRow := e.bindingRow(t, owner.ID)
		other := e.user(t, false)
		if _, err := svc.GrantOrganizationRole(ctx, o.ID, other, tenancy.OrgAdmin); err != nil {
			t.Fatal(err)
		}
		u := e.user(t, false)
		unrelated := func() string {
			var s string
			if err := e.own.QueryRow(ctx, `SELECT coalesce(md5(string_agg(b::text, E'\n' ORDER BY b.id)), '-') FROM app.role_bindings b
WHERE b.subject_user_id <> $1::text::uuid`, u.String()).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
		before := unrelated()
		bindings := func(scope string) int {
			return e.count(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid AND scope_type = $3`, o.ID.String(), u.String(), scope)
		}
		// Organization scope.
		b1, err := svc.GrantOrganizationRole(ctx, o.ID, u, tenancy.OrgMember)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range []tenancy.OrganizationRole{tenancy.OrgMember, tenancy.OrgAdmin, tenancy.OrgOwner} {
			if r != tenancy.OrgMember {
				b, err := svc.ChangeOrganizationRole(ctx, o.ID, u, r)
				if err != nil || b.ID != b1.ID || b.Role != string(r) {
					t.Fatalf("change to %s: %+v %v", r, b, err)
				}
			}
			m, err := svc.GetOrganizationMembership(ctx, o.ID, u)
			if err != nil || m != (tenancy.OrganizationMembership{UserID: u, OrganizationID: o.ID, Role: r}) || bindings("organization") != 1 {
				t.Fatalf("projection after %s: %+v %v", r, m, err)
			}
			if unrelated() != before {
				t.Fatalf("unrelated bindings changed at %s", r)
			}
		}
		if err := svc.RemoveOrganizationMembership(ctx, o.ID, u); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GetOrganizationMembership(ctx, o.ID, u); tenancy.CodeOf(err) != tenancy.CodeMembershipNotFound || bindings("organization") != 0 {
			t.Fatalf("after removal: %v", err)
		}
		b2, err := svc.GrantOrganizationRole(ctx, o.ID, u, tenancy.OrgMember)
		if err != nil || b2.ID == b1.ID || b2.Role != string(tenancy.OrgMember) || bindings("organization") != 1 {
			t.Fatalf("regrant: %+v %v", b2, err)
		}
		// Group scope.
		gb1, err := svc.GrantGroupRole(ctx, o.ID, g.ID, u, tenancy.GroupMember)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range []tenancy.GroupRole{tenancy.GroupMember, tenancy.GroupAdmin, tenancy.GroupOwner} {
			if r != tenancy.GroupMember {
				b, err := svc.ChangeGroupRole(ctx, o.ID, g.ID, u, r)
				if err != nil || b.ID != gb1.ID || b.Role != string(r) {
					t.Fatalf("change to %s: %+v %v", r, b, err)
				}
			}
			m, err := svc.GetGroupMembership(ctx, o.ID, g.ID, u)
			if err != nil || m != (tenancy.GroupMembership{UserID: u, OrganizationID: o.ID, GroupID: g.ID, Role: r}) || bindings("group") != 1 {
				t.Fatalf("projection after %s: %+v %v", r, m, err)
			}
		}
		if err := svc.RemoveGroupMembership(ctx, o.ID, g.ID, u); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GetGroupMembership(ctx, o.ID, g.ID, u); tenancy.CodeOf(err) != tenancy.CodeMembershipNotFound || bindings("group") != 0 {
			t.Fatalf("after removal: %v", err)
		}
		gb2, err := svc.GrantGroupRole(ctx, o.ID, g.ID, u, tenancy.GroupMember)
		if err != nil || gb2.ID == gb1.ID || bindings("group") != 1 {
			t.Fatalf("regrant: %+v %v", gb2, err)
		}
		if e.bindingRow(t, owner.ID) != ownerRow || unrelated() != before {
			t.Fatal("the creator's owner binding or an unrelated binding changed")
		}
		// Projection shapes, including the Group view's Organization anchor.
		var cols string
		if err := e.rt.QueryRow(ctx, `SELECT (SELECT string_agg(attname, ',' ORDER BY attnum) FROM pg_attribute WHERE attrelid = 'app.organization_memberships'::regclass AND attnum > 0)
  || '|' || (SELECT string_agg(attname, ',' ORDER BY attnum) FROM pg_attribute WHERE attrelid = 'app.group_memberships'::regclass AND attnum > 0)`).Scan(&cols); err != nil {
			t.Fatal(err)
		}
		if cols != "user_id,organization_id,role|user_id,organization_id,group_id,role" {
			t.Fatalf("projection columns: %s", cols)
		}
	})

	t.Run("IT-PROJ-2 projections are read-only", func(t *testing.T) {
		u := e.user(t, true)
		before := e.rows(t, A.ID)
		a, g := A.ID.String(), g1.ID.String()
		expectSQL(t, e.rt, "insert organization projection", "42501", "", `INSERT INTO app.organization_memberships (user_id, organization_id, role) VALUES ($2::text::uuid, $1::text::uuid, 'org_owner')`, a, u.String())
		expectSQL(t, e.rt, "update organization projection", "42501", "", `UPDATE app.organization_memberships SET role = 'org_owner' WHERE organization_id = $1::text::uuid`, a)
		expectSQL(t, e.rt, "delete organization projection", "42501", "", `DELETE FROM app.organization_memberships WHERE organization_id = $1::text::uuid`, a)
		expectSQL(t, e.rt, "insert group projection", "42501", "", `INSERT INTO app.group_memberships (user_id, organization_id, group_id, role) VALUES ($2::text::uuid, $1::text::uuid, $3::text::uuid, 'group_owner')`, a, u.String(), g)
		expectSQL(t, e.rt, "update group projection", "42501", "", `UPDATE app.group_memberships SET role = 'group_owner' WHERE group_id = $1::text::uuid`, g)
		expectSQL(t, e.rt, "delete group projection", "42501", "", `DELETE FROM app.group_memberships WHERE group_id = $1::text::uuid`, g)
		if after := e.rows(t, A.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-PROJ-3 readers see committed state only", func(t *testing.T) {
		o := e.org(t, "Reader Organization")
		u := e.user(t, false)
		reader := e.runtimeConn(t)
		read := func() (int, string) {
			var n int
			var role *string
			if err := reader.QueryRow(ctx, `SELECT count(*)::int, max(role) FROM app.organization_memberships WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid`,
				o.ID.String(), u.String()).Scan(&n, &role); err != nil {
				t.Fatal(err)
			}
			if role == nil {
				return n, ""
			}
			return n, *role
		}
		if _, err := svc.GrantOrganizationRole(ctx, o.ID, u, tenancy.OrgMember); err != nil {
			t.Fatal(err)
		}
		if n, r := read(); n != 1 || r != "org_member" {
			t.Fatalf("after grant: %d %s", n, r)
		}
		// A concurrent reader loop while an uncommitted change is open and while committed changes happen.
		stop := make(chan struct{})
		var bad atomic.Int64
		var seen sync.Map
		loopConn := e.runtimeConn(t)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				var n int
				var role *string
				if err := loopConn.QueryRow(ctx, `SELECT count(*)::int, max(role) FROM app.organization_memberships WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid`,
					o.ID.String(), u.String()).Scan(&n, &role); err != nil || n > 1 || (role != nil && *role == "org_owner") {
					bad.Add(1)
				}
				if role != nil {
					seen.Store(*role, true)
				}
			}
		}()
		tx, err := e.runtimeConn(t).Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE app.role_bindings SET role = 'org_owner' WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid`, o.ID.String(), u.String()); err != nil {
			t.Fatal(err)
		}
		time.Sleep(200 * time.Millisecond)
		if n, r := read(); n != 1 || r != "org_member" {
			t.Fatalf("uncommitted change visible: %d %s", n, r)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.ChangeOrganizationRole(ctx, o.ID, u, tenancy.OrgAdmin); err != nil {
			t.Fatal(err)
		}
		if n, r := read(); n != 1 || r != "org_admin" {
			t.Fatalf("after change: %d %s", n, r)
		}
		time.Sleep(100 * time.Millisecond)
		if err := svc.RemoveOrganizationMembership(ctx, o.ID, u); err != nil {
			t.Fatal(err)
		}
		if n, _ := read(); n != 0 {
			t.Fatalf("after removal: %d rows", n)
		}
		close(stop)
		wg.Wait()
		if bad.Load() != 0 {
			t.Fatalf("reader saw two rows or an uncommitted role %d times", bad.Load())
		}
		if _, ok := seen.Load("org_member"); !ok {
			t.Fatal("reader loop never observed the committed role")
		}
	})

	t.Run("IT-MULTI-1 one global user, independent scoped roles", func(t *testing.T) {
		U := e.user(t, false)
		identity := func() string {
			var s string
			if err := e.own.QueryRow(ctx, `SELECT md5((SELECT u::text FROM app.users u WHERE u.id = $1::text::uuid) || '|' ||
  (SELECT string_agg(p::text, ',' ORDER BY p.subject) FROM app.identity_principals p WHERE p.user_id = $1::text::uuid) || '|' ||
  (SELECT string_agg(s::text, ',' ORDER BY s.id) FROM app.sessions s WHERE s.user_id = $1::text::uuid))`, U.String()).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
		idBefore := identity()
		mustB := func(b tenancy.Binding, err error) tenancy.Binding {
			t.Helper()
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
		bA := mustB(svc.GrantOrganizationRole(ctx, A.ID, U, tenancy.OrgOwner))
		bB := mustB(svc.GrantOrganizationRole(ctx, B.ID, U, tenancy.OrgMember))
		g4, b4 := e.group(t, A.ID, U, "Multi Group", "multi-group")
		b1 := mustB(svc.GrantGroupRole(ctx, A.ID, g1.ID, U, tenancy.GroupAdmin))
		b2 := mustB(svc.GrantGroupRole(ctx, A.ID, g2.ID, U, tenancy.GroupMember))
		b3 := mustB(svc.GrantGroupRole(ctx, B.ID, g3.ID, U, tenancy.GroupMember))
		keep := []tenancy.BindingID{bB.ID, b4.ID, b2.ID, b3.ID}
		rowsOf := func() []string {
			var out []string
			for _, id := range keep {
				out = append(out, e.bindingRow(t, id))
			}
			return out
		}
		before := rowsOf()
		if c := mustB(svc.ChangeOrganizationRole(ctx, A.ID, U, tenancy.OrgAdmin)); c.ID != bA.ID {
			t.Fatal("binding UUID changed")
		}
		if err := svc.RemoveGroupMembership(ctx, A.ID, g1.ID, U); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(rowsOf(), before) {
			t.Fatal("other scoped bindings changed")
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE id = $1::text::uuid`, b1.ID.String()); n != 0 {
			t.Fatal("removed binding remains")
		}
		if identity() != idBefore {
			t.Fatal("the user's identity, principal or session rows changed")
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE subject_user_id = $1::text::uuid`, U.String()); n != 5 {
			t.Fatalf("U's bindings: %d, want 5", n)
		}
		if m, err := svc.GetGroupMembership(ctx, A.ID, g4.ID, U); err != nil || m.Role != tenancy.GroupOwner {
			t.Fatalf("U's own Group: %+v %v", m, err)
		}
	})

	t.Run("IT-MULTI-2 nothing is synthesized across scopes", func(t *testing.T) {
		V, W, C := e.user(t, false), e.user(t, false), e.user(t, false)
		if _, err := svc.GrantGroupRole(ctx, A.ID, g2.ID, V, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GrantOrganizationRole(ctx, B.ID, W, tenancy.OrgMember); err != nil {
			t.Fatal(err)
		}
		gc, _ := e.group(t, B.ID, C, "Creator Only", "creator-only")
		_, err := svc.GetOrganizationMembership(ctx, A.ID, V)
		wantCode(t, "V Organization membership", err, tenancy.CodeMembershipNotFound)
		_, err = svc.GetOrganizationMembership(ctx, B.ID, C)
		wantCode(t, "C Organization membership", err, tenancy.CodeMembershipNotFound)
		for user, want := range map[tenancy.UserID]int{V: 1, W: 0, C: 1} {
			if n := e.count(t, `SELECT count(*) FROM app.group_memberships WHERE user_id = $1::text::uuid`, user.String()); n != want {
				t.Errorf("group projection rows of %s: %d, want %d", user, n, want)
			}
		}
		if n := e.count(t, `SELECT count(*) FROM app.organization_memberships WHERE user_id = $1::text::uuid`, C.String()); n != 0 {
			t.Fatal("Group creation minted an Organization membership")
		}
		if m, err := svc.GetGroupMembership(ctx, B.ID, gc.ID, C); err != nil || m.Role != tenancy.GroupOwner {
			t.Fatalf("creator membership: %+v %v", m, err)
		}
	})

	t.Run("IT-OWN-1 creation with the creator's owner binding", func(t *testing.T) {
		C := e.user(t, false)
		before := e.count(t, `SELECT count(*) FROM app.organization_memberships WHERE user_id = $1::text::uuid`, C.String())
		g, b, err := svc.CreateGroup(ctx, A.ID, C, "Owner Group", "owner-group")
		if err != nil {
			t.Fatal(err)
		}
		if b.SubjectUserID != C || b.ScopeType != tenancy.ScopeGroup || b.Role != string(tenancy.GroupOwner) || b.Source != tenancy.SourceMembership ||
			b.OrganizationID != A.ID || b.GroupID == nil || *b.GroupID != g.ID || b.ScopeID != [16]byte(g.ID) {
			t.Fatalf("owner binding: %+v", b)
		}
		if got := e.bindingRow(t, b.ID); got != fmt.Sprintf("%s %s group %s %s group_owner membership", A.ID, C, g.ID, g.ID) {
			t.Fatalf("committed binding: %s", got)
		}
		stored, err := svc.GetGroup(ctx, A.ID, g.ID)
		if err != nil || stored.ID != g.ID || stored.Name != g.Name || stored.Slug != g.Slug || !stored.CreatedAt.Equal(g.CreatedAt) {
			t.Fatalf("committed group: %s %v", h(stored), err)
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE group_id = $1::text::uuid`, g.ID.String()); n != 1 {
			t.Fatalf("bindings of the new Group: %d", n)
		}
		if after := e.count(t, `SELECT count(*) FROM app.organization_memberships WHERE user_id = $1::text::uuid`, C.String()); after != before {
			t.Fatal("an Organization membership was minted")
		}
	})

	t.Run("IT-OWN-2 creator and Organization prechecks", func(t *testing.T) {
		before := e.rows(t, A.ID, B.ID)
		unknownUser, _ := tenancy.ParseUserID("7f000000-0000-4000-8000-0000000000aa")
		unknownOrg, _ := tenancy.ParseOrganizationID("7f000000-0000-4000-8000-0000000000ab")
		C := e.user(t, false)
		_, _, err := svc.CreateGroup(ctx, A.ID, unknownUser, "Nobody", "nobody")
		wantCode(t, "unknown creator", err, tenancy.CodeUserNotFound)
		_, _, err = svc.CreateGroup(ctx, A.ID, tenancy.UserID{}, "Nobody", "nobody")
		wantCode(t, "zero creator", err, tenancy.CodeInvalidID)
		_, err = tenancy.ParseUserID("{" + C.String() + "}")
		wantCode(t, "malformed creator", err, tenancy.CodeInvalidID)
		_, _, err = svc.CreateGroup(ctx, unknownOrg, C, "Nobody", "nobody")
		wantCode(t, "unknown organization", err, tenancy.CodeOrganizationNotFound)
		if after := e.rows(t, A.ID, B.ID); after != before {
			t.Fatal("rows changed")
		}
		if n := e.count(t, `SELECT count(*) FROM app.groups WHERE slug = 'nobody'`); n != 0 {
			t.Fatal("a Group was created")
		}
	})

	t.Run("IT-OWN-3 injected failures leave neither row", func(t *testing.T) {
		C := e.user(t, false)
		before := e.rows(t, A.ID)
		check := func(name string, db *faultDB, ctx context.Context) {
			t.Helper()
			s := tenancy.NewService(db, nil)
			g, b, err := s.CreateGroup(ctx, A.ID, C, "Fault Group", "fault-group")
			wantCode(t, name, err, tenancy.CodeUnavailable)
			if g != (tenancy.Group{}) || b.ID != (tenancy.BindingID{}) {
				t.Errorf("%s returned a result", name)
			}
			if after := e.rows(t, A.ID); after != before {
				t.Fatalf("%s: rows changed", name)
			}
		}
		check("binding insert failure", &faultDB{pool: e.pool, failOn: "INSERT INTO app.role_bindings"}, ctx)
		check("commit failure", &faultDB{pool: e.pool, failCommit: true}, ctx)
		cctx, cancel := context.WithCancel(ctx)
		defer cancel()
		check("cancellation between the inserts", &faultDB{pool: e.pool, after: "INSERT INTO app.groups", cancel: cancel}, cctx)
		if !errors.Is(func() error {
			cctx, cancel := context.WithCancel(ctx)
			defer cancel()
			_, _, err := tenancy.NewService(&faultDB{pool: e.pool, after: "INSERT INTO app.groups", cancel: cancel}, nil).CreateGroup(cctx, A.ID, C, "Fault Group", "fault-group")
			return err
		}(), context.Canceled) {
			t.Fatal("cancellation cause not exposed")
		}
	})

	t.Run("IT-OWN-4 creator deletion races", func(t *testing.T) {
		// (a) The creator is deleted (owner-level, committed) between the creator precheck and the binding insert.
		C := e.user(t, true)
		o := e.org(t, "Creator Race Organization")
		own := e.ownerConn(t)
		db := &faultDB{pool: e.pool, after: "INSERT INTO app.groups", hook: func() {
			if _, err := own.Exec(ctx, `DELETE FROM app.users WHERE id = $1::text::uuid`, C.String()); err != nil {
				t.Errorf("delete creator: %v", err)
			}
		}}
		_, _, err := tenancy.NewService(db, nil).CreateGroup(ctx, o.ID, C, "Race Group", "race-group")
		wantCode(t, "creator deleted mid-creation", err, tenancy.CodeUserNotFound)
		if n := e.count(t, `SELECT count(*) FROM app.groups WHERE organization_id = $1::text::uuid`, o.ID.String()); n != 0 {
			t.Fatal("a Group without its owner remains")
		}
		// (b) The owner-level delete starts after the binding insert and waits; creation commits; the delete fails.
		C2 := e.user(t, true)
		own2 := e.ownerConn(t)
		res := make(chan error, 1)
		db2 := &faultDB{pool: e.pool, after: "INSERT INTO app.role_bindings", hook: func() {
			go func() {
				_, err := own2.Exec(ctx, `DELETE FROM app.users WHERE id = $1::text::uuid`, C2.String())
				res <- err
			}()
			e.waitBlocked(t)
		}}
		g, b, err := tenancy.NewService(db2, nil).CreateGroup(ctx, o.ID, C2, "Race Group", "race-group")
		if err != nil || b.SubjectUserID != C2 {
			t.Fatalf("creation: %v", err)
		}
		if code, con := sqlErr(<-res); code != "23503" || con != "role_bindings_subject_fkey" {
			t.Fatalf("creator delete after creation: %s %s", code, con)
		}
		if ms, err := svc.ListGroupMemberships(ctx, o.ID, g.ID); err != nil || len(ms) != 1 || ms[0].UserID != C2 {
			t.Fatalf("owner binding: %+v %v", ms, err)
		}
	})

	t.Run("IT-OWN-5 concurrent creation of one slug", func(t *testing.T) {
		o := e.org(t, "Concurrent Creation Organization")
		var creators []tenancy.UserID
		for range 16 {
			creators = append(creators, e.user(t, true))
		}
		results := make([]error, 16)
		groups := make([]tenancy.Group, 16)
		var wg sync.WaitGroup
		for i, c := range creators {
			wg.Add(1)
			go func() {
				defer wg.Done()
				groups[i], _, results[i] = svc.CreateGroup(ctx, o.ID, c, "Contended", "contended")
			}()
		}
		wg.Wait()
		winner := -1
		for i, err := range results {
			switch tenancy.CodeOf(err) {
			case "":
				if winner >= 0 {
					t.Fatal("two winners")
				}
				winner = i
			case tenancy.CodeSlugConflict:
			default:
				t.Fatalf("creation %d: %v", i, err)
			}
		}
		if winner < 0 {
			t.Fatal("no winner")
		}
		if n := e.count(t, `SELECT count(*) FROM app.groups WHERE organization_id = $1::text::uuid`, o.ID.String()); n != 1 {
			t.Fatalf("groups: %d", n)
		}
		ms, err := svc.ListGroupMemberships(ctx, o.ID, groups[winner].ID)
		if err != nil || len(ms) != 1 || ms[0].UserID != creators[winner] || ms[0].Role != tenancy.GroupOwner {
			t.Fatalf("owner: %+v %v", ms, err)
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid`, o.ID.String()); n != 1 {
			t.Fatalf("bindings: %d", n)
		}
	})

	t.Run("IT-OWN-6 owner bindings are ordinary model rows", func(t *testing.T) {
		o := e.org(t, "Ordinary Owner Organization")
		C := e.user(t, false)
		g, owner := e.group(t, o.ID, C, "Ordinary Owner Group", "ordinary-owner")
		other := e.user(t, false)
		ob, err := svc.GrantGroupRole(ctx, o.ID, g.ID, other, tenancy.GroupMember)
		if err != nil {
			t.Fatal(err)
		}
		otherRow := e.bindingRow(t, ob.ID)
		c, err := svc.ChangeGroupRole(ctx, o.ID, g.ID, C, tenancy.GroupAdmin)
		if err != nil || c.ID != owner.ID || c.Role != string(tenancy.GroupAdmin) {
			t.Fatalf("change owner to admin: %+v %v", c, err)
		}
		if err := svc.RemoveGroupMembership(ctx, o.ID, g.ID, C); err != nil {
			t.Fatalf("remove former owner: %v", err)
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE id = $1::text::uuid`, owner.ID.String()); n != 0 {
			t.Fatal("binding remains")
		}
		if e.bindingRow(t, ob.ID) != otherRow {
			t.Fatal("another binding changed")
		}
	})

	t.Run("IT-SLUG-1 every reserved entry", func(t *testing.T) {
		before := e.rows(t, A.ID)
		for _, s := range tenancy.ReservedSlugs() {
			_, _, err := svc.CreateGroup(ctx, A.ID, F, "Reserved", s)
			wantCode(t, "create "+s, err, tenancy.CodeSlugReserved)
			_, err = svc.ChangeGroupSlug(ctx, A.ID, g2.ID, s)
			wantCode(t, "change "+s, err, tenancy.CodeSlugReserved)
			expectSQL(t, e.rt, "runtime "+s, "23514", "groups_slug_reserved", `INSERT INTO app.groups (organization_id, name, slug) VALUES ($1::text::uuid, 'Reserved', $2)`, A.ID.String(), s)
		}
		if len(tenancy.ReservedSlugs()) != 24 {
			t.Fatal("reserved set size")
		}
		if after := e.rows(t, A.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-SLUG-2 grammar negatives", func(t *testing.T) {
		before := e.rows(t, A.ID)
		for _, s := range []string{"", strings.Repeat("a", 64), "Acme", "acme_revenue", "acme.revenue", "-acme", "acme-", "acme--revenue",
			"acme%2frevenue", "%2e%2e", "..", "a/b", "a b", "é", "ａｃｍｅ", "асme", " acme", "acme\n"} {
			if s != "" {
				_, _, err := svc.CreateGroup(ctx, A.ID, F, "Grammar", s)
				wantCode(t, fmt.Sprintf("create %q", s), err, tenancy.CodeInvalidSlug)
			}
			_, err := svc.ChangeGroupSlug(ctx, A.ID, g2.ID, s)
			wantCode(t, fmt.Sprintf("change %q", s), err, tenancy.CodeInvalidSlug)
			expectSQL(t, e.rt, fmt.Sprintf("runtime %q", s), "23514", "groups_slug_format", `INSERT INTO app.groups (organization_id, name, slug) VALUES ($1::text::uuid, 'Grammar', $2)`, A.ID.String(), s)
		}
		if after := e.rows(t, A.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-SLUG-3 Organization-local uniqueness", func(t *testing.T) {
		if g1.Slug != g3.Slug || g1.OrganizationID == g3.OrganizationID {
			t.Fatal("same slug across Organizations was not created")
		}
		before := e.rows(t, A.ID)
		_, _, err := svc.CreateGroup(ctx, A.ID, F, "Duplicate", g1.Slug)
		wantCode(t, "same-Organization create", err, tenancy.CodeSlugConflict)
		_, err = svc.ChangeGroupSlug(ctx, A.ID, g2.ID, g1.Slug)
		wantCode(t, "same-Organization edit", err, tenancy.CodeSlugConflict)
		if after := e.rows(t, A.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-SLUG-4 concurrent slug collisions", func(t *testing.T) {
		o := e.org(t, "Slug Race Organization")
		creators := make([]tenancy.UserID, 16)
		for i := range creators {
			creators[i] = e.user(t, true)
		}
		var wins atomic.Int64
		var wg sync.WaitGroup
		for _, c := range creators {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err := svc.CreateGroup(ctx, o.ID, c, "Same Slug", "same-slug")
				if err == nil {
					wins.Add(1)
				} else if tenancy.CodeOf(err) != tenancy.CodeSlugConflict {
					t.Errorf("create: %v", err)
				}
			}()
		}
		wg.Wait()
		if wins.Load() != 1 {
			t.Fatalf("concurrent creates won: %d", wins.Load())
		}
		var gs []tenancy.Group
		for i := range 16 {
			g, _ := e.group(t, o.ID, F, fmt.Sprintf("Edit %d", i), fmt.Sprintf("edit-%d", i))
			gs = append(gs, g)
		}
		wins.Store(0)
		for _, g := range gs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := svc.ChangeGroupSlug(ctx, o.ID, g.ID, "edited")
				if err == nil {
					wins.Add(1)
				} else if tenancy.CodeOf(err) != tenancy.CodeSlugConflict {
					t.Errorf("edit: %v", err)
				}
			}()
		}
		wg.Wait()
		if wins.Load() != 1 {
			t.Fatalf("concurrent edits won: %d", wins.Load())
		}
		changed := 0
		for _, g := range gs {
			now, err := svc.GetGroup(ctx, o.ID, g.ID)
			if err != nil || now.Name != g.Name || now.ID != g.ID {
				t.Fatalf("group %s: %s %v", g.Slug, h(now), err)
			}
			if now.Slug == "edited" {
				changed++
			} else if now.Slug != g.Slug {
				t.Fatalf("a losing edit changed the slug of %s to %s", g.Slug, now.Slug)
			}
		}
		if changed != 1 {
			t.Fatalf("groups with the edited slug: %d", changed)
		}
		// Create versus edit, both racing for one slug.
		target := gs[0]
		if target.Slug == "edited" {
			target = gs[1]
		}
		var cErr, eErr error
		wg.Add(2)
		go func() { defer wg.Done(); _, _, cErr = svc.CreateGroup(ctx, o.ID, F, "Create Side", "contested") }()
		go func() { defer wg.Done(); _, eErr = svc.ChangeGroupSlug(ctx, o.ID, target.ID, "contested") }()
		wg.Wait()
		if (cErr == nil) == (eErr == nil) || (cErr != nil && tenancy.CodeOf(cErr) != tenancy.CodeSlugConflict) || (eErr != nil && tenancy.CodeOf(eErr) != tenancy.CodeSlugConflict) {
			t.Fatalf("create vs edit: %v / %v", cErr, eErr)
		}
		// Deterministic interleaving: an uncommitted insert of a slug blocks an edit to it, which then conflicts.
		tx, err := e.runtimeConn(t).Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO app.groups (organization_id, name, slug) VALUES ($1::text::uuid, 'Interleaved', 'interleaved')`, o.ID.String()); err != nil {
			t.Fatal(err)
		}
		res := make(chan error, 1)
		go func() { _, err := svc.ChangeGroupSlug(ctx, o.ID, target.ID, "interleaved"); res <- err }()
		e.waitBlocked(t)
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		wantCode(t, "edit after a committed insert", <-res, tenancy.CodeSlugConflict)
	})

	t.Run("IT-SLUG-5 advisory suggestions", func(t *testing.T) {
		o := e.org(t, "Suggestion Organization")
		if s, err := svc.SuggestGroupSlug(ctx, o.ID, "Acme Revenue"); err != nil || s != "acme-revenue" {
			t.Fatalf("free base: %q %v", s, err)
		}
		e.group(t, o.ID, F, "Acme Revenue", "")
		if s, err := svc.SuggestGroupSlug(ctx, o.ID, "Acme Revenue"); err != nil || s != "acme-revenue-2" {
			t.Fatalf("taken base: %q %v", s, err)
		}
		if s, err := svc.SuggestGroupSlug(ctx, o.ID, "API"); err != nil || s != "api-2" {
			t.Fatalf("reserved base: %q %v", s, err)
		}
		_, err := svc.SuggestGroupSlug(ctx, o.ID, "東京チーム")
		wantCode(t, "unsluggable", err, tenancy.CodeSlugRequired)
		unknownOrg, _ := tenancy.ParseOrganizationID("7f000000-0000-4000-8000-0000000000ac")
		_, err = svc.SuggestGroupSlug(ctx, unknownOrg, "Acme")
		wantCode(t, "unknown organization", err, tenancy.CodeOrganizationNotFound)
		s, err := svc.SuggestGroupSlug(ctx, o.ID, "Acme Revenue")
		if err != nil {
			t.Fatal(err)
		}
		e.group(t, o.ID, F, "Taken First", s)
		_, _, err = svc.CreateGroup(ctx, o.ID, F, "Acme Revenue", s)
		wantCode(t, "suggestion taken before create", err, tenancy.CodeSlugConflict)
		// The base-only create reports the collision deterministically.
		_, _, err = svc.CreateGroup(ctx, o.ID, F, "Acme Revenue", "")
		wantCode(t, "base-only create", err, tenancy.CodeSlugConflict)
		_, _, err = svc.CreateGroup(ctx, o.ID, F, "API", "")
		wantCode(t, "reserved base create", err, tenancy.CodeSlugReserved)
	})

	t.Run("IT-SLUG-6 slug lookup is Organization-scoped", func(t *testing.T) {
		o := e.org(t, "Lookup Organization")
		e.group(t, B.ID, F, "Only In B", "only-in-b")
		_, err := svc.GetGroupBySlug(ctx, o.ID, "only-in-b")
		wantCode(t, "slug of another Organization", err, tenancy.CodeGroupNotFound)
		if g, err := svc.GetGroupBySlug(ctx, B.ID, "only-in-b"); err != nil || g.OrganizationID != B.ID {
			t.Fatalf("own lookup: %v", err)
		}
	})

	t.Run("IT-SCOPE-1 foreign Group selectors", func(t *testing.T) {
		u := e.user(t, false)
		if _, err := svc.GrantGroupRole(ctx, B.ID, g3.ID, u, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		before := e.rows(t, B.ID)
		unknownGroup, _ := tenancy.ParseGroupID("7f000000-0000-4000-8000-0000000000ad")
		calls := map[string]func(tenancy.GroupID) error{
			"get":    func(g tenancy.GroupID) error { _, err := svc.GetGroup(ctx, A.ID, g); return err },
			"rename": func(g tenancy.GroupID) error { _, err := svc.RenameGroup(ctx, A.ID, g, "Taken Over"); return err },
			"slug":   func(g tenancy.GroupID) error { _, err := svc.ChangeGroupSlug(ctx, A.ID, g, "taken-over"); return err },
			"grant": func(g tenancy.GroupID) error {
				_, err := svc.GrantGroupRole(ctx, A.ID, g, u, tenancy.GroupAdmin)
				return err
			},
			"change": func(g tenancy.GroupID) error {
				_, err := svc.ChangeGroupRole(ctx, A.ID, g, u, tenancy.GroupAdmin)
				return err
			},
			"remove": func(g tenancy.GroupID) error { return svc.RemoveGroupMembership(ctx, A.ID, g, u) },
			"member": func(g tenancy.GroupID) error { _, err := svc.GetGroupMembership(ctx, A.ID, g, u); return err },
			"list":   func(g tenancy.GroupID) error { _, err := svc.ListGroupMemberships(ctx, A.ID, g); return err },
		}
		for name, call := range calls {
			foreign, unknown := call(g3.ID), call(unknownGroup)
			wantCode(t, name+" foreign", foreign, tenancy.CodeGroupNotFound)
			if foreign == nil || unknown == nil || foreign.Error() != unknown.Error() {
				t.Errorf("%s: foreign %v, unknown %v", name, foreign, unknown)
			}
		}
		_, err := svc.ChangeOrganizationRole(ctx, A.ID, u, tenancy.OrgAdmin)
		wantCode(t, "change without an Organization membership", err, tenancy.CodeMembershipNotFound)
		err = svc.RemoveOrganizationMembership(ctx, A.ID, u)
		wantCode(t, "remove without an Organization membership", err, tenancy.CodeMembershipNotFound)
		if after := e.rows(t, B.ID); after != before {
			t.Fatal("B's rows changed")
		}
	})

	t.Run("IT-ATOM-1 rejected inputs change nothing", func(t *testing.T) {
		u := e.user(t, false)
		before := e.rows(t, A.ID, B.ID)
		var zero tenancy.OrganizationID
		for name, c := range map[string]struct {
			call func() error
			want tenancy.Code
		}{
			"bad name":       {func() error { _, err := svc.CreateOrganization(ctx, "a\tb"); return err }, tenancy.CodeInvalidName},
			"long name":      {func() error { _, err := svc.RenameOrganization(ctx, A.ID, strings.Repeat("x", 201)); return err }, tenancy.CodeInvalidName},
			"zero org":       {func() error { _, err := svc.RenameOrganization(ctx, zero, "x"); return err }, tenancy.CodeInvalidID},
			"bad slug":       {func() error { _, _, err := svc.CreateGroup(ctx, A.ID, u, "x", "X"); return err }, tenancy.CodeInvalidSlug},
			"reserved":       {func() error { _, err := svc.ChangeGroupSlug(ctx, A.ID, g1.ID, "admin"); return err }, tenancy.CodeSlugReserved},
			"slug required":  {func() error { _, _, err := svc.CreateGroup(ctx, A.ID, u, "🚀", ""); return err }, tenancy.CodeSlugRequired},
			"bad role":       {func() error { _, err := svc.GrantGroupRole(ctx, A.ID, g1.ID, u, "advisor"); return err }, tenancy.CodeInvalidRole},
			"conflict":       {func() error { _, _, err := svc.CreateGroup(ctx, A.ID, u, "x", g1.Slug); return err }, tenancy.CodeSlugConflict},
			"existing":       {func() error { _, err := svc.GrantGroupRole(ctx, A.ID, g1.ID, F, tenancy.GroupMember); return err }, tenancy.CodeMembershipExists},
			"no membership":  {func() error { return svc.RemoveGroupMembership(ctx, A.ID, g1.ID, u) }, tenancy.CodeMembershipNotFound},
			"foreign rename": {func() error { _, err := svc.RenameGroup(ctx, B.ID, g1.ID, "x"); return err }, tenancy.CodeGroupNotFound},
		} {
			wantCode(t, name, c.call(), c.want)
		}
		if after := e.rows(t, A.ID, B.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-ATOM-2 injected store failures", func(t *testing.T) {
		u := e.user(t, false)
		before := e.rows(t, A.ID)
		closed, err := pgxpool.New(ctx, e.pool.Config().ConnString())
		if err != nil {
			t.Fatal(err)
		}
		closed.Close()
		for name, dbs := range map[string][2]*faultDB{
			"commit failure":    {{pool: e.pool, failCommit: true}, {pool: e.pool, failCommit: true}},
			"statement failure": {{pool: e.pool, failOn: "UPDATE app.groups"}, {pool: e.pool, failOn: "INSERT INTO app.role_bindings"}},
			"closed pool":       {{pool: e.pool, closed: closed}, {pool: e.pool, closed: closed}},
		} {
			g, err := tenancy.NewService(dbs[0], nil).RenameGroup(ctx, A.ID, g1.ID, "Should Not Stick")
			wantCode(t, name+" rename", err, tenancy.CodeUnavailable)
			if g != (tenancy.Group{}) {
				t.Errorf("%s returned a result", name)
			}
			b, err := tenancy.NewService(dbs[1], nil).GrantGroupRole(ctx, A.ID, g1.ID, u, tenancy.GroupMember)
			wantCode(t, name+" grant", err, tenancy.CodeUnavailable)
			if b.ID != (tenancy.BindingID{}) {
				t.Errorf("%s returned a binding", name)
			}
		}
		rmDB := &faultDB{pool: e.pool, failOn: "DELETE FROM app.role_bindings"}
		wantCode(t, "remove statement failure", tenancy.NewService(rmDB, nil).RemoveGroupMembership(ctx, A.ID, g1.ID, F), tenancy.CodeUnavailable)
		cctx, cancel := context.WithCancel(ctx)
		_, err = tenancy.NewService(&faultDB{pool: e.pool, after: "SELECT EXISTS", cancel: cancel}, nil).GrantGroupRole(cctx, A.ID, g1.ID, u, tenancy.GroupMember)
		cancel()
		wantCode(t, "cancellation mid-transaction", err, tenancy.CodeUnavailable)
		if after := e.rows(t, A.ID); after != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-CONC-1 concurrent role changes", func(t *testing.T) {
		o := e.org(t, "Concurrent Change Organization")
		u := e.user(t, false)
		b, err := svc.GrantOrganizationRole(ctx, o.ID, u, tenancy.OrgMember)
		if err != nil {
			t.Fatal(err)
		}
		roles := []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin, tenancy.OrgMember}
		var wg sync.WaitGroup
		for i := range 32 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if c, err := svc.ChangeOrganizationRole(ctx, o.ID, u, roles[i%3]); err != nil || c.ID != b.ID {
					t.Errorf("change: %+v %v", c, err)
				}
			}()
		}
		wg.Wait()
		var id, role string
		if err := e.own.QueryRow(ctx, `SELECT id::text, role FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid`, o.ID.String(), u.String()).Scan(&id, &role); err != nil {
			t.Fatal(err)
		}
		if id != b.ID.String() || !slices.Contains(roles, tenancy.OrganizationRole(role)) {
			t.Fatalf("after concurrent changes: %s %s", id, role)
		}
		if m, err := svc.GetOrganizationMembership(ctx, o.ID, u); err != nil || string(m.Role) != role {
			t.Fatalf("projection differs from the row: %+v %v", m, err)
		}
		// Deterministic: A changes to admin (uncommitted), B (the service) changes to owner and waits; A commits; B wins.
		tx, err := e.runtimeConn(t).Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE app.role_bindings SET role = 'org_admin' WHERE id = $1::text::uuid`, b.ID.String()); err != nil {
			t.Fatal(err)
		}
		res := make(chan error, 1)
		go func() { _, err := svc.ChangeOrganizationRole(ctx, o.ID, u, tenancy.OrgOwner); res <- err }()
		e.waitBlocked(t)
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if err := <-res; err != nil {
			t.Fatal(err)
		}
		if m, _ := svc.GetOrganizationMembership(ctx, o.ID, u); m.Role != tenancy.OrgOwner {
			t.Fatalf("last committed change: %s", m.Role)
		}
		// Change versus remove: a committed removal makes the waiting change find nothing.
		tx, err = e.runtimeConn(t).Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM app.role_bindings WHERE id = $1::text::uuid`, b.ID.String()); err != nil {
			t.Fatal(err)
		}
		go func() { _, err := svc.ChangeOrganizationRole(ctx, o.ID, u, tenancy.OrgAdmin); res <- err }()
		e.waitBlocked(t)
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		wantCode(t, "change after a committed removal", <-res, tenancy.CodeMembershipNotFound)
		// Remove versus change: a committed change, then the waiting removal deletes the row.
		b2, err := svc.GrantOrganizationRole(ctx, o.ID, u, tenancy.OrgMember)
		if err != nil {
			t.Fatal(err)
		}
		tx, err = e.runtimeConn(t).Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE app.role_bindings SET role = 'org_admin' WHERE id = $1::text::uuid`, b2.ID.String()); err != nil {
			t.Fatal(err)
		}
		go func() { res <- svc.RemoveOrganizationMembership(ctx, o.ID, u) }()
		e.waitBlocked(t)
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if err := <-res; err != nil {
			t.Fatalf("removal after a committed change: %v", err)
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid`, o.ID.String()); n != 0 {
			t.Fatalf("bindings: %d", n)
		}
	})

	t.Run("IT-LOG-1 operation records", func(t *testing.T) {
		// Operations run inside observability.Middleware for correlation; canaries are used as names on success,
		// rejected and failed paths. Rejections that carry a canary are refused by service validation before SQL, and
		// database-reached failures only show key columns, so no canary can reach the PostgreSQL log either.
		start := len(e.logs.String())
		c0, c1, c2 := e.canaries[0], e.canaries[1], e.canaries[2]
		u := e.user(t, false)
		var created tenancy.Organization
		handler := observability.Middleware(slog.New(slog.DiscardHandler), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var err error
			created, err = svc.CreateOrganization(ctx, c0) // ok
			if err != nil {
				t.Errorf("create: %v", err)
			}
			_, _ = svc.RenameOrganization(ctx, created.ID, c1+"\x00")               // rejected: invalid_name
			_, _, _ = svc.CreateGroup(ctx, created.ID, u, c1, "acme--x")            // rejected: invalid_slug
			_, _, _ = svc.CreateGroup(ctx, created.ID, u, c2, "log-group")          // ok
			_, _, _ = svc.CreateGroup(ctx, created.ID, u, c2, "log-group")          // rejected: slug_conflict (key columns only)
			_, _ = svc.GrantOrganizationRole(ctx, created.ID, u, tenancy.OrgMember) // ok, organization scope
			_, _ = svc.ChangeOrganizationRole(ctx, created.ID, u, "advisor")        // rejected: invalid_role
			_ = svc.RemoveOrganizationMembership(ctx, created.ID, u)                // ok
			failing := tenancy.NewService(&faultDB{pool: e.pool, failCommit: true}, observability.NewLogger(e.logs, slog.LevelInfo, nil))
			_, _ = failing.RenameOrganization(ctx, created.ID, c2)                     // failed: unavailable
			_, _ = failing.GrantOrganizationRole(ctx, created.ID, u, tenancy.OrgAdmin) // failed: unavailable
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		req.Header.Set("X-Request-Id", "req-tenancy-0001")
		req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
		handler.ServeHTTP(httptest.NewRecorder(), req)
		// One call outside any request: no correlation keys.
		if _, err := svc.RenameOrganization(ctx, created.ID, "Outside"); err != nil {
			t.Fatal(err)
		}
		type rec struct {
			Level, Event, Operation, Outcome string
			ScopeType                        string `json:"scope_type"`
			ErrorCode                        string `json:"error_code"`
			RequestID                        string `json:"request_id"`
			TraceID                          string `json:"trace_id"`
			Service                          string
		}
		var got []rec
		for _, l := range strings.Split(strings.TrimSpace(e.logs.String()[start:]), "\n") {
			var r rec
			if err := json.Unmarshal([]byte(l), &r); err != nil {
				t.Fatalf("log line is not JSON: %v", err)
			}
			got = append(got, r)
		}
		want := []rec{
			{"info", "tenancy.operation", "organization.create", "ok", "", "", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"warn", "tenancy.operation", "organization.rename", "rejected", "", "invalid_name", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"warn", "tenancy.operation", "group.create", "rejected", "", "invalid_slug", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"info", "tenancy.operation", "group.create", "ok", "", "", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"warn", "tenancy.operation", "group.create", "rejected", "", "slug_conflict", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"info", "tenancy.operation", "membership.grant", "ok", "organization", "", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"warn", "tenancy.operation", "membership.change", "rejected", "organization", "invalid_role", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"info", "tenancy.operation", "membership.remove", "ok", "organization", "", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"error", "tenancy.operation", "organization.rename", "failed", "", "unavailable", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"error", "tenancy.operation", "membership.grant", "failed", "organization", "unavailable", "req-tenancy-0001", "0af7651916cd43dd8448eb211c80319c", "core"},
			{"info", "tenancy.operation", "organization.rename", "ok", "", "", "", "", "core"},
		}
		if !slices.Equal(got, want) {
			t.Fatalf("records:\n got %+v\nwant %+v", got, want)
		}
		logs := e.logs.String()
		for _, c := range e.canaries {
			if strings.Contains(logs, c) {
				t.Fatal("a canary reached the operation log")
			}
		}
		for _, id := range []string{created.ID.String(), u.String()} {
			if strings.Contains(logs, id) {
				t.Fatal("a UUID reached the operation log")
			}
		}
		allowed := map[string]bool{"time": true, "level": true, "service": true, "event": true, "operation": true, "scope_type": true,
			"outcome": true, "error_code": true, "request_id": true, "trace_id": true}
		for _, l := range strings.Split(strings.TrimSpace(logs), "\n") {
			var m map[string]any
			if err := json.Unmarshal([]byte(l), &m); err != nil {
				t.Fatalf("log line is not JSON: %v", err)
			}
			for k := range m {
				if !allowed[k] {
					t.Fatalf("operation log carries the key %s", k)
				}
			}
		}
	})

	// MEM-1 (G019 Plan r1 §6.2; F-01 refinement a). NOT EXECUTED while the native test hold applies.
	t.Run("MEM-1 atomic bootstrap with an active owner", func(t *testing.T) {
		orgs := func() int { return e.count(t, `SELECT count(*) FROM app.organizations`) }
		owner := e.user(t, true)
		org, b, err := svc.CreateOrganizationWithOwner(ctx, "Bootstrap Organization", owner)
		if err != nil || b.OrganizationID != org.ID || b.SubjectUserID != owner || b.Role != string(tenancy.OrgOwner) {
			t.Fatalf("create: %v, binding %+v", err, b)
		}
		if n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid`, org.ID.String()); n != 1 {
			t.Errorf("bindings: %d, want the one owner binding", n)
		}

		// A fault between the two inserts: neither row.
		before := orgs()
		cctx, cancel := context.WithCancel(ctx)
		_, _, err = tenancy.NewService(&faultDB{pool: e.pool, after: "INSERT INTO app.organizations", cancel: cancel}, nil).
			CreateOrganizationWithOwner(cctx, "Fault Bootstrap", owner)
		cancel()
		if err == nil || orgs() != before {
			t.Errorf("fault: %v, %d Organizations (want %d)", err, orgs(), before)
		}

		// An unknown owner stays user_not_found; nothing is written.
		var unknown tenancy.UserID
		unknown[0], unknown[6], unknown[8] = 0x5a, 0x40, 0x80
		if _, _, err := svc.CreateOrganizationWithOwner(ctx, "Unknown Owner", unknown); tenancy.CodeOf(err) != tenancy.CodeUserNotFound || orgs() != before {
			t.Errorf("unknown owner: %v", err)
		}
	})

	t.Run("MEM-1 inactive owner and the post-disable wait refuse", func(t *testing.T) {
		orgs := func() int { return e.count(t, `SELECT count(*) FROM app.organizations`) }
		disabled := e.user(t, true)
		if _, err := e.rt.Exec(ctx, `UPDATE app.users SET status = 'disabled', disabled_at = now() WHERE id = $1::text::uuid`, disabled.String()); err != nil {
			t.Fatal(err)
		}
		before := orgs()
		if _, _, err := svc.CreateOrganizationWithOwner(ctx, "Disabled Owner", disabled); tenancy.CodeOf(err) != tenancy.CodeRecipientIneligible || orgs() != before {
			t.Errorf("disabled owner: %v, %d Organizations (want %d)", err, orgs(), before)
		}

		// The creation waits on a disable that holds the owner's row FOR UPDATE, then reads the committed status.
		owner := e.user(t, true)
		x, err := e.pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = x.Rollback(ctx) }()
		if _, err := x.Exec(ctx, `SELECT 1 FROM app.users WHERE id = $1::text::uuid FOR UPDATE`, owner.String()); err != nil {
			t.Fatal(err)
		}
		if _, err := x.Exec(ctx, `UPDATE app.users SET status = 'disabled', disabled_at = now() WHERE id = $1::text::uuid`, owner.String()); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() {
			_, _, err := svc.CreateOrganizationWithOwner(ctx, "Waiting Bootstrap", owner)
			done <- err
		}()
		e.waitBlocked(t)
		if err := x.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if err := <-done; tenancy.CodeOf(err) != tenancy.CodeRecipientIneligible || orgs() != before {
			t.Errorf("after the disable committed: %v, %d Organizations (want %d)", err, orgs(), before)
		}
	})
}
