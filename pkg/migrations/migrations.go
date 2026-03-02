package migrations

import (
	"embed"
	"io/fs"
)

// PublicMigrationsFS embeds the public schema migration files
//
//go:embed public/*.sql
var publicMigrationsFS embed.FS

// TenantMigrationsFS embeds the tenant schema migration files
//
//go:embed tenant/*.sql
var tenantMigrationsFS embed.FS

// GetPublicMigrationsFS returns the embedded public migrations filesystem.
func GetPublicMigrationsFS() fs.FS {
	sub, err := fs.Sub(publicMigrationsFS, "public")
	if err != nil {
		panic("failed to get public migrations subdirectory: " + err.Error())
	}
	return sub
}

// GetTenantMigrationsFS returns the embedded tenant migrations filesystem.
func GetTenantMigrationsFS() fs.FS {
	sub, err := fs.Sub(tenantMigrationsFS, "tenant")
	if err != nil {
		panic("failed to get tenant migrations subdirectory: " + err.Error())
	}
	return sub
}
