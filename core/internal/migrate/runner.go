// Package migrate runs database migrations. It logs in as the migrator role,
// assumes the NOLOGIN owner role with SET ROLE on every connection, verifies
// that identity before doing anything, and applies the embedded migration set
// with goose. The core runtime binary never imports this package, and the
// runtime role is never used here.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// Database names leave room for the role suffixes (_retention is the longest).
var dbNameRe = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,47}$`)

var appEnvs = map[string]bool{"development": true, "test": true, "production": true}

// Config is the migration runner's configuration, read from the environment.
type Config struct {
	Database string // DATABASE_NAME; also the prefix of every role name
	Host     string // DB_HOST
	Port     int    // DB_PORT
	Password string // DB_MIGRATOR_PASSWORD
	AppEnv   string // APP_ENV
}

func (c Config) OwnerRole() string     { return c.Database + "_owner" }
func (c Config) MigratorRole() string  { return c.Database + "_migrator" }
func (c Config) RuntimeRole() string   { return c.Database + "_runtime" }
func (c Config) RetentionRole() string { return c.Database + "_retention" }

// LoadConfig reads and validates the configuration, reporting every problem
// at once. Error messages name variables, never their values.
func LoadConfig(getenv func(string) string) (Config, error) {
	var problems []string
	req := func(name string) string {
		v := strings.TrimSpace(getenv(name))
		if v == "" {
			problems = append(problems, name+" is required")
		}
		return v
	}
	c := Config{
		Database: req("DATABASE_NAME"),
		Host:     req("DB_HOST"),
		Password: getenv("DB_MIGRATOR_PASSWORD"),
		AppEnv:   req("APP_ENV"),
	}
	if c.Password == "" {
		problems = append(problems, "DB_MIGRATOR_PASSWORD is required")
	}
	if c.Database != "" && !dbNameRe.MatchString(c.Database) {
		problems = append(problems, "DATABASE_NAME is not a valid database name")
	}
	if c.AppEnv != "" && !appEnvs[c.AppEnv] {
		problems = append(problems, "APP_ENV must be development, test or production")
	}
	if raw := req("DB_PORT"); raw != "" {
		p, err := strconv.Atoi(raw)
		if err != nil || p < 1 || p > 65535 {
			problems = append(problems, "DB_PORT must be a port number")
		}
		c.Port = p
	}
	if len(problems) > 0 {
		return Config{}, errors.New("invalid configuration: " + strings.Join(problems, "; "))
	}
	return c, nil
}

// Runner applies one migration set to one database.
type Runner struct {
	cfg      Config
	db       *sql.DB
	provider *goose.Provider
}

// openOptions are test hooks; production code always uses the zero value.
type openOptions struct {
	skipSetRole bool // connect without assuming the owner role (preflight must refuse)
	skipRevoke  bool // create metadata without the protection revokes (postcondition must refuse)
}

// Open connects, verifies the migration identity and prepares goose. Nothing
// is created before the identity check passes.
func Open(ctx context.Context, cfg Config, migrations fs.FS) (*Runner, error) {
	return open(ctx, cfg, migrations, openOptions{})
}

func open(ctx context.Context, cfg Config, migrations fs.FS, opts openOptions) (*Runner, error) {
	connCfg, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}
	connCfg.Host = cfg.Host
	connCfg.Port = uint16(cfg.Port)
	connCfg.Database = cfg.Database
	connCfg.User = cfg.MigratorRole()
	connCfg.Password = cfg.Password
	connCfg.ConnectTimeout = 5 * time.Second
	connCfg.RuntimeParams["application_name"] = "migrate"
	// Exactly one plaintext connection to Host:Port on the internal data
	// network, as for core. Nothing derived by ParseConfig (default host,
	// sslmode=prefer fallbacks, PG* variables) may add another endpoint.
	connCfg.TLSConfig = nil
	connCfg.Fallbacks = nil

	var dbOpts []stdlib.OptionOpenDB
	if !opts.skipSetRole {
		setRole := "SET ROLE " + quote(cfg.OwnerRole())
		dbOpts = append(dbOpts, stdlib.OptionAfterConnect(func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, setRole)
			return err
		}))
	}
	db := stdlib.OpenDB(*connCfg, dbOpts...)
	db.SetMaxOpenConns(4)

	if err := preflight(ctx, db, cfg); err != nil {
		_ = db.Close()
		return nil, err
	}
	store, err := newProtectedStore(cfg)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	store.skipRevoke = opts.skipRevoke
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	provider, err := goose.NewProvider(goose.DialectCustom, db, migrations, providerOptions(store, locker)...)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Runner{cfg: cfg, db: db, provider: provider}, nil
}

// providerOptions is the complete goose configuration. WithIsolateDDL is
// deliberately absent: metadata protection relies on goose creating the
// version table inside one transaction.
func providerOptions(store *protectedStore, locker lock.SessionLocker) []goose.ProviderOption {
	return []goose.ProviderOption{
		goose.WithStore(store),
		goose.WithSessionLocker(locker),
	}
}

// preflight refuses to continue unless the connection is the migrator role
// acting as the owner of schema app, without superuser or BYPASSRLS.
func preflight(ctx context.Context, db *sql.DB, cfg Config) error {
	var current, session string
	var schemaOwner sql.NullString
	var super, bypass bool
	err := db.QueryRowContext(ctx, `
SELECT current_user::text, session_user::text,
       (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app'),
       r.rolsuper, r.rolbypassrls
FROM pg_roles r WHERE r.rolname = session_user`).Scan(&current, &session, &schemaOwner, &super, &bypass)
	if err != nil {
		return fmt.Errorf("connect as %s: %w", cfg.MigratorRole(), err)
	}
	var problems []string
	if session != cfg.MigratorRole() {
		problems = append(problems, "session user is "+session+", expected "+cfg.MigratorRole())
	}
	if current != cfg.OwnerRole() {
		problems = append(problems, "current role is "+current+", expected "+cfg.OwnerRole())
	}
	if !schemaOwner.Valid || schemaOwner.String != cfg.OwnerRole() {
		problems = append(problems, "schema app must exist and be owned by "+cfg.OwnerRole())
	}
	if super || bypass {
		problems = append(problems, "the migrator role must not be superuser or BYPASSRLS")
	}
	if len(problems) > 0 {
		return errors.New("refusing to migrate: " + strings.Join(problems, "; "))
	}
	return nil
}

// Close releases the provider and its database handle.
func (r *Runner) Close() error {
	return r.provider.Close()
}

// Status prints every known migration with its state, then the current version.
func (r *Runner) Status(ctx context.Context, w io.Writer) error {
	states, err := r.provider.Status(ctx)
	if err != nil {
		return err
	}
	for _, s := range states {
		applied := ""
		if s.State == goose.StateApplied {
			applied = s.AppliedAt.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(w, "migration\t%s\t%s\t%s\n", s.Source.Path, s.State, applied)
	}
	return r.Version(ctx, w)
}

// Version prints the current version.
func (r *Runner) Version(ctx context.Context, w io.Writer) error {
	v, err := r.provider.GetDBVersion(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "current_version\t%d\n", v)
	return nil
}

// Up applies every pending migration in order.
func (r *Runner) Up(ctx context.Context, w io.Writer) error {
	results, err := r.provider.Up(ctx)
	for _, res := range results {
		fmt.Fprintf(w, "applied\t%s\n", res.Source.Path)
	}
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Fprintln(w, "no pending migrations")
	}
	return r.Version(ctx, w)
}

// Down rolls back the latest applied migration only.
func (r *Runner) Down(ctx context.Context, w io.Writer) error {
	v, err := r.provider.GetDBVersion(ctx)
	if err != nil {
		return err
	}
	if v == 0 {
		return errors.New("no applied migration to roll back")
	}
	res, err := r.provider.Down(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "rolled_back\t%s\n", res.Source.Path)
	return r.Version(ctx, w)
}

// Reapply rolls back the latest applied migration and applies it again.
func (r *Runner) Reapply(ctx context.Context, w io.Writer) error {
	v, err := r.provider.GetDBVersion(ctx)
	if err != nil {
		return err
	}
	if v == 0 {
		return errors.New("no applied migration to reapply")
	}
	down, err := r.provider.Down(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "rolled_back\t%s\n", down.Source.Path)
	up, err := r.provider.ApplyVersion(ctx, v, true)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "applied\t%s\n", up.Source.Path)
	return r.Version(ctx, w)
}

// Exit codes of the migrate command.
const (
	ExitOK      = 0
	ExitFailure = 1
	ExitUsage   = 2 // usage, configuration or refusal; nothing was changed
)

const usage = "usage: migrate status|up|down|reapply|version"

// destructive commands are refused when APP_ENV=production.
var commands = map[string]bool{"status": false, "up": false, "down": true, "reapply": true, "version": false}

// Run is the migrate command. Results go to stdout; errors go to the logger
// on stderr, which redacts the migrator password.
func Run(ctx context.Context, args []string, getenv func(string) string, migrations fs.FS, stdout io.Writer, logf func(msg string)) int {
	if len(args) != 1 {
		logf(usage)
		return ExitUsage
	}
	cmd := args[0]
	destructive, known := commands[cmd]
	if !known {
		logf(usage)
		return ExitUsage
	}
	cfg, err := LoadConfig(getenv)
	if err != nil {
		logf(err.Error())
		return ExitUsage
	}
	if destructive && cfg.AppEnv == "production" {
		logf("refusing " + cmd + " with APP_ENV=production")
		return ExitUsage
	}
	r, err := Open(ctx, cfg, migrations)
	if err != nil {
		logf(err.Error())
		return ExitFailure
	}
	defer r.Close()
	run := map[string]func(context.Context, io.Writer) error{
		"status": r.Status, "up": r.Up, "down": r.Down, "reapply": r.Reapply, "version": r.Version,
	}[cmd]
	if err := run(ctx, stdout); err != nil {
		logf(cmd + " failed: " + err.Error())
		return ExitFailure
	}
	return ExitOK
}
