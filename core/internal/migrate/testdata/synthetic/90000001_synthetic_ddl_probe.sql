-- Test-only probe for the migration integration tests: proves that migration
-- DDL runs as the owner role and rolls back cleanly. It is never part of the
-- normal migration set (core/migrations) and never runs outside a disposable
-- test database.

-- +goose Up
CREATE TABLE app.synthetic_ddl_probe (
  id bigint PRIMARY KEY,
  note text NOT NULL
);

-- +goose Down
DROP TABLE app.synthetic_ddl_probe;
