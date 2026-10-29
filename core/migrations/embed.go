// Package migrations holds the normal, ordered database migration set. The
// files are embedded into the migrate binary, so nothing is read from disk at
// run time. Test-only migrations never live here.
package migrations

import "embed"

// FS is the normal migration set: goose SQL files named NNNNN_name.sql.
//
//go:embed *.sql
var FS embed.FS
