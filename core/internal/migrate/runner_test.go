package migrate

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/pressly/goose/v3/lock"

	"core/migrations"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func validEnv() map[string]string {
	return map[string]string{
		"DATABASE_NAME":        "sample_db",
		"DB_HOST":              "postgres",
		"DB_PORT":              "5432",
		"DB_MIGRATOR_PASSWORD": "s3cret-migrator-value",
		"APP_ENV":              "development",
	}
}

func TestLoadConfigValid(t *testing.T) {
	cfg, err := LoadConfig(env(validEnv()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 5432 || cfg.Database != "sample_db" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	roles := []string{cfg.OwnerRole(), cfg.MigratorRole(), cfg.RuntimeRole(), cfg.RetentionRole()}
	want := []string{"sample_db_owner", "sample_db_migrator", "sample_db_runtime", "sample_db_retention"}
	for i := range want {
		if roles[i] != want[i] {
			t.Errorf("role %d = %q, want %q", i, roles[i], want[i])
		}
	}
}

func TestLoadConfigReportsEveryProblemWithoutValues(t *testing.T) {
	_, err := LoadConfig(env(map[string]string{}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, name := range []string{"DATABASE_NAME", "DB_HOST", "DB_PORT", "DB_MIGRATOR_PASSWORD", "APP_ENV"} {
		if !strings.Contains(err.Error(), name+" is required") {
			t.Errorf("missing %s not reported: %v", name, err)
		}
	}
	bad := validEnv()
	bad["DATABASE_NAME"] = "Bad-Name"
	bad["APP_ENV"] = "staging"
	bad["DB_PORT"] = "99999"
	_, err = LoadConfig(env(bad))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"DATABASE_NAME is not a valid", "APP_ENV must be", "DB_PORT must be"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q not reported: %v", want, err)
		}
	}
	if strings.Contains(err.Error(), bad["DB_MIGRATOR_PASSWORD"]) {
		t.Fatal("error message contains the migrator password")
	}
	long := validEnv()
	long["DATABASE_NAME"] = strings.Repeat("a", 49)
	if _, err := LoadConfig(env(long)); err == nil {
		t.Fatal("a 49-character database name leaves no room for role suffixes")
	}
}

func TestRunRefusesBeforeConnecting(t *testing.T) {
	prod := validEnv()
	prod["APP_ENV"] = "production"
	prod["DB_HOST"] = "host.invalid" // a connection attempt would fail with ExitFailure, not ExitUsage
	cases := []struct {
		name string
		args []string
		env  map[string]string
		want string
	}{
		{"no command", nil, validEnv(), "usage"},
		{"two commands", []string{"up", "down"}, validEnv(), "usage"},
		{"unknown command", []string{"drop"}, validEnv(), "usage"},
		{"invalid config", []string{"status"}, map[string]string{}, "invalid configuration"},
		{"down in production", []string{"down"}, prod, "refusing down with APP_ENV=production"},
		{"reapply in production", []string{"reapply"}, prod, "refusing reapply with APP_ENV=production"},
	}
	for _, c := range cases {
		var logged []string
		code := Run(context.Background(), c.args, env(c.env), migrations.FS, &bytes.Buffer{}, func(m string) { logged = append(logged, m) })
		if code != ExitUsage {
			t.Errorf("%s: exit %d, want %d", c.name, code, ExitUsage)
		}
		if len(logged) != 1 || !strings.Contains(logged[0], c.want) {
			t.Errorf("%s: logged %q, want %q", c.name, logged, c.want)
		}
	}
}

// The normal set is exactly the reviewed set (the baseline, the G004 identity migration, the G005 tenancy migration, the
// G006 BusinessProfile migration, the G007 Organization policy migration, the G008 row-level security migration and
// the G009 stored-object migration, the G011 audit migration, the G013 retention migration, the G014 MCP connection
// migration, the G016 organization-domain migration, the G016 security amendment's session-binding migration, the G017
// branding migration, the G018 Group onboarding migration, the G019 membership/invitation migration, the G058 enrollment migration and
// the G060 Group-authority migration, the G062 platform-signup migration, the G063 MCP connection-name migration, the G020 labels migration, the G057
// platform-label migration, the G071 default-tenant reader migration); the
// synthetic probe is test data only.
func TestNormalSetIsTheReviewedSet(t *testing.T) {
	names, err := fs.Glob(migrations.FS, "*")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(names, ",") != "00001_baseline.sql,00002_identity.sql,00003_tenancy.sql,00004_business_profiles.sql,00005_organization_policy.sql,"+
		"00006_row_level_security.sql,00007_stored_objects.sql,00008_audit_events.sql,00009_retention_maintenance.sql,"+
		"00010_mcp_connections.sql,00011_organization_domains.sql,00012_session_binding_and_claim_exclusivity.sql,00013_organization_branding.sql,"+
		"00014_group_onboarding.sql,00015_membership_invitations.sql,00016_organization_enrollment.sql,00017_group_authority.sql,"+
		"00018_platform_signup.sql,00019_mcp_connection_names.sql,00020_labels.sql,00021_platform_labels.sql,00022_default_tenant.sql" {
		t.Fatalf("normal migration set = %v, want [00001_baseline.sql 00002_identity.sql 00003_tenancy.sql 00004_business_profiles.sql "+
			"00005_organization_policy.sql 00006_row_level_security.sql 00007_stored_objects.sql 00008_audit_events.sql "+
			"00009_retention_maintenance.sql 00010_mcp_connections.sql 00011_organization_domains.sql "+
			"00012_session_binding_and_claim_exclusivity.sql 00013_organization_branding.sql 00014_group_onboarding.sql "+
			"00015_membership_invitations.sql 00016_organization_enrollment.sql 00017_group_authority.sql "+
			"00018_platform_signup.sql 00019_mcp_connection_names.sql 00020_labels.sql 00021_platform_labels.sql "+
			"00022_default_tenant.sql]", names)
	}
	synthetic, err := fs.Glob(syntheticFS(t), "*.sql")
	if err != nil || len(synthetic) != 1 {
		t.Fatalf("synthetic probe: %v %v", synthetic, err)
	}
	if _, err := fs.Stat(migrations.FS, synthetic[0]); err == nil {
		t.Fatal("the synthetic probe must not be part of the normal set")
	}
}

func testConfig() Config {
	cfg, _ := LoadConfig(env(validEnv()))
	return cfg
}

func TestProtectionTargetsExactObjects(t *testing.T) {
	s, err := newProtectedStore(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	stmts := s.protectionStatements()
	if len(stmts) != 2 {
		t.Fatalf("want 2 statements, got %d", len(stmts))
	}
	grantees := `PUBLIC, "sample_db_runtime", "sample_db_retention", "sample_db_migrator"`
	if stmts[0] != "REVOKE ALL ON TABLE app.goose_db_version FROM "+grantees {
		t.Errorf("table revoke = %q", stmts[0])
	}
	do := stmts[1]
	for _, want := range []string{
		"pg_get_serial_sequence('app.goose_db_version', 'id')",
		"'app.goose_db_version'::regclass",
		"REVOKE ALL ON SEQUENCE %s FROM " + grantees,
		`'"sample_db_owner"'::regrole`,
		"RAISE EXCEPTION 'migration metadata is not owner-only'",
	} {
		if !strings.Contains(do, want) {
			t.Errorf("protection block lacks %q:\n%s", want, do)
		}
	}
	for _, stmt := range stmts {
		// The sequence is resolved through the catalog, never by a guessed name.
		for _, pattern := range []string{" LIKE ", " ~ ", "SIMILAR TO", "goose_db_version%", "_id_seq"} {
			if strings.Contains(stmt, pattern) {
				t.Errorf("statement uses name matching (%q):\n%s", pattern, stmt)
			}
		}
	}
	s.skipRevoke = true
	hooked := s.protectionStatements()
	if len(hooked) != 1 || strings.Contains(hooked[0], "REVOKE") || !strings.Contains(hooked[0], "not owner-only") {
		t.Fatalf("skipRevoke must keep only the verification:\n%v", hooked)
	}
}

// recorder is both a fake inner store and a fake transaction handle, so the
// order of creation and protection is observable.
type recorder struct {
	calls  []string
	execs  []string
	failAt int // fail the n-th Exec (1-based); 0 never fails
}

func (r *recorder) ExecContext(_ context.Context, q string, _ ...any) (sql.Result, error) {
	r.execs = append(r.execs, q)
	r.calls = append(r.calls, "exec")
	if r.failAt == len(r.execs) {
		return nil, errors.New("exec failed")
	}
	return driverResult{}, nil
}
func (r *recorder) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	panic("unexpected query")
}
func (r *recorder) QueryRowContext(context.Context, string, ...any) *sql.Row {
	panic("unexpected query")
}

type driverResult struct{}

func (driverResult) LastInsertId() (int64, error) { return 0, nil }
func (driverResult) RowsAffected() (int64, error) { return 0, nil }

func (r *recorder) Tablename() string { r.calls = append(r.calls, "tablename"); return MetadataTable }
func (r *recorder) CreateVersionTable(_ context.Context, db database.DBTxConn) error {
	if db != database.DBTxConn(r) {
		panic("creation must use the caller's transaction handle")
	}
	r.calls = append(r.calls, "create")
	return nil
}
func (r *recorder) Insert(context.Context, database.DBTxConn, database.InsertRequest) error {
	r.calls = append(r.calls, "insert")
	return nil
}
func (r *recorder) Delete(context.Context, database.DBTxConn, int64) error {
	r.calls = append(r.calls, "delete")
	return nil
}
func (r *recorder) GetMigration(context.Context, database.DBTxConn, int64) (*database.GetMigrationResult, error) {
	r.calls = append(r.calls, "get")
	return nil, nil
}
func (r *recorder) GetLatestVersion(context.Context, database.DBTxConn) (int64, error) {
	r.calls = append(r.calls, "latest")
	return 0, nil
}
func (r *recorder) ListMigrations(context.Context, database.DBTxConn) ([]*database.ListMigrationsResult, error) {
	r.calls = append(r.calls, "list")
	return nil, nil
}
func (r *recorder) TableExists(context.Context, database.DBTxConn) (bool, error) {
	r.calls = append(r.calls, "exists")
	return true, nil
}

func TestCreateVersionTableProtectsOnTheSameTransaction(t *testing.T) {
	ctx := context.Background()
	rec := &recorder{}
	s, _ := newProtectedStore(testConfig())
	want := s.protectionStatements()
	s.inner = rec
	if err := s.CreateVersionTable(ctx, rec); err != nil {
		t.Fatal(err)
	}
	if strings.Join(rec.calls, ",") != "create,exec,exec" || strings.Join(rec.execs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("calls %v, execs %v", rec.calls, rec.execs)
	}
	for failAt := 1; failAt <= 2; failAt++ {
		rec = &recorder{failAt: failAt}
		s.inner = rec
		if err := s.CreateVersionTable(ctx, rec); err == nil {
			t.Fatalf("a failing protection statement %d must fail the creation", failAt)
		}
	}
}

func TestStoreDelegatesEverythingElse(t *testing.T) {
	ctx := context.Background()
	rec := &recorder{}
	s := &protectedStore{inner: rec}
	_ = s.Tablename()
	_ = s.Insert(ctx, rec, database.InsertRequest{})
	_ = s.Delete(ctx, rec, 1)
	_, _ = s.GetMigration(ctx, rec, 1)
	_, _ = s.GetLatestVersion(ctx, rec)
	_, _ = s.ListMigrations(ctx, rec)
	_, _ = s.TableExists(ctx, rec)
	if got := strings.Join(rec.calls, ","); got != "tablename,insert,delete,get,latest,list,exists" || len(rec.execs) != 0 {
		t.Fatalf("delegation calls = %s, execs = %v", got, rec.execs)
	}
}

// goose accepts the protected store and the complete option set without a
// connection; the option set is exactly the store and the session locker.
func TestProviderAcceptsProtectedStore(t *testing.T) {
	db, err := sql.Open("pgx", "host=unused.invalid")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s, _ := newProtectedStore(testConfig())
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		t.Fatal(err)
	}
	opts := providerOptions(s, locker)
	if len(opts) != 2 {
		t.Fatalf("want exactly WithStore and WithSessionLocker, got %d options", len(opts))
	}
	p, err := goose.NewProvider(goose.DialectCustom, db, migrations.FS, opts...)
	if err != nil {
		t.Fatal(err)
	}
	// The normal set lists versions 1..22 (G017: 00013_organization_branding; G018: 00014_group_onboarding; G019:
	// 00015_membership_invitations; G058: 00016_organization_enrollment; G060: 00017_group_authority; G062:
	// 00018_platform_signup; G063: 00019_mcp_connection_names; G020: 00020_labels; G057: 00021_platform_labels; G071:
	// 00022_default_tenant).
	if srcs := p.ListSources(); len(srcs) != 22 || srcs[0].Version != 1 || srcs[1].Version != 2 || srcs[2].Version != 3 ||
		srcs[3].Version != 4 || srcs[4].Version != 5 || srcs[5].Version != 6 || srcs[6].Version != 7 ||
		srcs[7].Version != 8 || srcs[8].Version != 9 || srcs[9].Version != 10 || srcs[10].Version != 11 || srcs[11].Version != 12 ||
		srcs[12].Version != 13 || srcs[13].Version != 14 || srcs[14].Version != 15 || srcs[15].Version != 16 || srcs[16].Version != 17 ||
		srcs[17].Version != 18 || srcs[18].Version != 19 || srcs[19].Version != 20 || srcs[20].Version != 21 ||
		srcs[21].Version != 22 {
		t.Fatalf("sources = %v", srcs)
	}
}

func syntheticFS(t *testing.T) fs.FS {
	t.Helper()
	sub, err := fs.Sub(testdataFS(), "testdata/synthetic")
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

// combinedFS is the normal set plus the synthetic probe (tests only).
func combinedFS(t *testing.T) fs.FS {
	t.Helper()
	out := fstest.MapFS{}
	for _, f := range []fs.FS{migrations.FS, syntheticFS(t)} {
		names, err := fs.Glob(f, "*.sql")
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range names {
			b, err := fs.ReadFile(f, n)
			if err != nil {
				t.Fatal(err)
			}
			out[n] = &fstest.MapFile{Data: b}
		}
	}
	return out
}

//go:embed testdata/synthetic/*.sql
var testdata embed.FS

func testdataFS() fs.FS { return testdata }
