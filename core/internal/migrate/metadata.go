package migrate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3/database"
)

// MetadataTable is goose's version table. It is system infrastructure, not a
// tenant-owned domain table: no organization_id, no row-level security and no
// grant to any role other than its owner.
const MetadataTable = "app.goose_db_version"

// protectedStore is goose's PostgreSQL store with one change: the metadata
// table and its identity sequence are made owner-only inside the transaction
// that creates them. goose (v3.28.0, without WithIsolateDDL) creates the table
// and inserts version 0 in a single transaction on the first use of any
// command, so the creation either commits owner-only or not at all. The
// schema's default privileges, which grant sequences to the runtime role,
// therefore never reach a committed metadata object.
type protectedStore struct {
	inner      database.Store
	owner      string   // quoted owner role
	revokeFrom []string // quoted roles that must hold nothing: runtime, retention, migrator
	skipRevoke bool     // test hook: skip the revokes so the postcondition must catch the default grant
}

func newProtectedStore(cfg Config) (*protectedStore, error) {
	inner, err := database.NewStore(database.DialectPostgres, MetadataTable)
	if err != nil {
		return nil, err
	}
	return &protectedStore{
		inner: inner,
		owner: quote(cfg.OwnerRole()),
		revokeFrom: []string{
			quote(cfg.RuntimeRole()),
			quote(cfg.RetentionRole()),
			quote(cfg.MigratorRole()),
		},
	}, nil
}

func quote(role string) string { return pgx.Identifier{role}.Sanitize() }

// protectionStatements returns the statements run after creation, in order:
// revoke on the exact table, then (in one DO block) revoke on its exact
// identity sequence and verify that both objects are owned by the owner role
// and grant nothing to anyone else. A failed verification raises, which rolls
// back goose's creation transaction.
func (s *protectedStore) protectionStatements() []string {
	grantees := "PUBLIC, " + strings.Join(s.revokeFrom, ", ")
	var stmts []string
	if !s.skipRevoke {
		stmts = append(stmts, fmt.Sprintf("REVOKE ALL ON TABLE %s FROM %s", MetadataTable, grantees))
	}
	revokeSeq := ""
	if !s.skipRevoke {
		revokeSeq = fmt.Sprintf("EXECUTE format('REVOKE ALL ON SEQUENCE %%s FROM %s', seq);\n", strings.ReplaceAll(grantees, "'", "''"))
	}
	stmts = append(stmts, fmt.Sprintf(`DO $protect$
DECLARE
  tbl regclass := '%[1]s'::regclass;
  seq regclass := pg_get_serial_sequence('%[1]s', 'id')::regclass;
  objects int;
  bad int;
BEGIN
  IF seq IS NULL THEN
    RAISE EXCEPTION 'migration metadata: identity sequence not found';
  END IF;
  %[2]sSELECT count(*) INTO objects FROM pg_class WHERE oid IN (tbl, seq);
  SELECT count(*) INTO bad
  FROM pg_class c
  LEFT JOIN LATERAL aclexplode(coalesce(c.relacl, acldefault(
    CASE WHEN c.relkind = 'S' THEN 's'::"char" ELSE 'r'::"char" END, c.relowner))) a
    ON a.grantee <> c.relowner
  WHERE c.oid IN (tbl, seq)
    AND (c.relowner <> '%[3]s'::regrole OR a.grantee IS NOT NULL);
  IF objects <> 2 OR bad <> 0 THEN
    RAISE EXCEPTION 'migration metadata is not owner-only';
  END IF;
END
$protect$`, MetadataTable, revokeSeq, strings.ReplaceAll(s.owner, "'", "''")))
	return stmts
}

// CreateVersionTable creates the table through goose's store and protects it
// on the same transaction handle.
func (s *protectedStore) CreateVersionTable(ctx context.Context, db database.DBTxConn) error {
	if err := s.inner.CreateVersionTable(ctx, db); err != nil {
		return err
	}
	for _, stmt := range s.protectionStatements() {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("protect migration metadata: %w", err)
		}
	}
	return nil
}

// TableExists keeps goose's catalog-based existence check when the inner
// store provides it.
func (s *protectedStore) TableExists(ctx context.Context, db database.DBTxConn) (bool, error) {
	if ext, ok := s.inner.(interface {
		TableExists(context.Context, database.DBTxConn) (bool, error)
	}); ok {
		return ext.TableExists(ctx, db)
	}
	return false, errors.ErrUnsupported
}

func (s *protectedStore) Tablename() string { return s.inner.Tablename() }

func (s *protectedStore) Insert(ctx context.Context, db database.DBTxConn, req database.InsertRequest) error {
	return s.inner.Insert(ctx, db, req)
}

func (s *protectedStore) Delete(ctx context.Context, db database.DBTxConn, version int64) error {
	return s.inner.Delete(ctx, db, version)
}

func (s *protectedStore) GetMigration(ctx context.Context, db database.DBTxConn, version int64) (*database.GetMigrationResult, error) {
	return s.inner.GetMigration(ctx, db, version)
}

func (s *protectedStore) GetLatestVersion(ctx context.Context, db database.DBTxConn) (int64, error) {
	return s.inner.GetLatestVersion(ctx, db)
}

func (s *protectedStore) ListMigrations(ctx context.Context, db database.DBTxConn) ([]*database.ListMigrationsResult, error) {
	return s.inner.ListMigrations(ctx, db)
}
