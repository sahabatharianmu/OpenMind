package migrations

import (
	"embed"
	"io/fs"
)

// MigrationsFS embeds the migration files
//
//go:embed *.sql
var migrationsFS embed.FS

// GetMigrationsFS returns the embedded migrations filesystem.
func GetMigrationsFS() fs.FS {
	return migrationsFS
}
