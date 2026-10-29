// Command migrate applies the embedded database migrations as the migrator
// role acting as the schema owner. It runs as a one-shot container; the core
// server never migrates and never receives migrator credentials.
//
//	migrate status|up|down|reapply|version
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"core/internal/migrate"
	"core/internal/observability"
	"core/migrations"
)

func main() {
	logger := observability.NewLogger(os.Stderr, slog.LevelInfo, []string{os.Getenv("DB_MIGRATOR_PASSWORD")})
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	code := migrate.Run(ctx, os.Args[1:], os.Getenv, migrations.FS, os.Stdout, func(msg string) {
		logger.Error(msg)
	})
	stop()
	os.Exit(code)
}
