package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/internal/modules/tenant/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/migrations"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database connection
func InitDB(cfg *config.Config, log logger.Logger) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance", zap.Error(err))
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info("Database connection established successfully")
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// runMigrationsWithFS runs migrations from a given filesystem against a specific schema.
// This is the core migration runner used by both public and tenant migration functions.
func runMigrationsWithFS(db *gorm.DB, migrationsFS fs.FS, appLogger logger.Logger, schemaName string, migrationsTableSuffix string) error {
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("Failed to get sql.DB from gorm.DB", zap.Error(err))
		return err
	}

	config := &migratepostgres.Config{
		// Use a distinct migrations table name so public and tenant tracks
		// don't collide if both ever target the same schema (shouldn't happen,
		// but this is defensive).
		MigrationsTable: "schema_migrations" + migrationsTableSuffix,
	}
	if schemaName != "" {
		config.SchemaName = schemaName
		appLogger.Info("Running migrations in schema",
			zap.String("schema", schemaName),
			zap.String("migrations_table", config.MigrationsTable))
	}

	driver, err := migratepostgres.WithInstance(sqlDB, config)
	if err != nil {
		appLogger.Error("Failed to create postgres driver", zap.Error(err))
		return err
	}

	source, err := iofs.New(migrationsFS, ".")
	if err != nil {
		appLogger.Error("Failed to create migration source", zap.Error(err))
		return err
	}

	m, err := migrate.NewWithInstance(
		"migrations",
		source,
		"postgres",
		driver,
	)
	if err != nil {
		appLogger.Error("Failed to create migrate instance", zap.Error(err))
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		appLogger.Error("Failed to run migrations", zap.Error(err),
			zap.String("schema", schemaName))
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		appLogger.Info("No new migrations to apply",
			zap.String("schema", schemaName))
	} else {
		appLogger.Info("Migrations completed successfully",
			zap.String("schema", schemaName))
	}

	return nil
}

// RunPublicMigrations runs migrations for the public schema only.
// These include: users, organizations, tenants, audit_logs, notifications, etc.
func RunPublicMigrations(db *gorm.DB, appLogger logger.Logger) error {
	appLogger.Info("Running PUBLIC schema migrations...")
	return runMigrationsWithFS(db, migrations.GetPublicMigrationsFS(), appLogger, "", "")
}

// RunTenantMigrations runs tenant-specific migrations for a single tenant schema.
// These include: patients, appointments, clinical_notes, invoices, etc.
func RunTenantMigrations(db *gorm.DB, appLogger logger.Logger, schemaName string) error {
	appLogger.Info("Running TENANT schema migrations...", zap.String("schema", schemaName))
	return runMigrationsWithFS(db, migrations.GetTenantMigrationsFS(), appLogger, schemaName, "")
}

// RunTenantMigrationsForAll runs tenant migrations for all active tenant schemas.
// This ensures all tenants are synced with the latest tenant migration files.
// Uses pagination to handle large numbers of tenants efficiently.
func RunTenantMigrationsForAll(ctx context.Context, db *gorm.DB, tenantRepo repository.TenantRepository, appLogger logger.Logger) error {
	const batchSize = 1000
	offset := 0
	successCount := 0
	failCount := 0
	totalProcessed := 0

	appLogger.Info("Starting tenant migrations with pagination", zap.Int("batch_size", batchSize))

	for {
		tenants, total, err := tenantRepo.List(batchSize, offset)
		if err != nil {
			appLogger.Error("Failed to list tenants for migration", zap.Error(err), zap.Int("offset", offset))
			return fmt.Errorf("failed to list tenants: %w", err)
		}

		if len(tenants) == 0 {
			break
		}

		appLogger.Info("Processing tenant batch",
			zap.Int("batch_size", len(tenants)),
			zap.Int("offset", offset),
			zap.Int64("total_tenants", total))

		for _, tenant := range tenants {
			if tenant.DeletedAt.Valid {
				continue
			}

			if err := RunTenantMigrations(db, appLogger, tenant.SchemaName); err != nil {
				appLogger.Error("Failed to run migrations for tenant",
					zap.Error(err),
					zap.String("schema_name", tenant.SchemaName),
					zap.String("organization_id", tenant.OrganizationID.String()))
				failCount++
				continue
			}

			appLogger.Info("Migrations completed for tenant",
				zap.String("schema_name", tenant.SchemaName),
				zap.String("organization_id", tenant.OrganizationID.String()))
			successCount++
			totalProcessed++
		}

		if len(tenants) < batchSize {
			break
		}

		offset += batchSize
	}

	if totalProcessed == 0 {
		appLogger.Info("No active tenants found, skipping tenant migrations")
		return nil
	}

	appLogger.Info("Tenant migrations summary",
		zap.Int("total_processed", totalProcessed),
		zap.Int("successful", successCount),
		zap.Int("failed", failCount))

	if failCount > 0 {
		return fmt.Errorf("migrations failed for %d tenant(s)", failCount)
	}

	return nil
}

// RunMigrations is kept for backward compatibility but delegates to RunPublicMigrations.
// Deprecated: Use RunPublicMigrations or RunTenantMigrations directly.
func RunMigrations(db *gorm.DB, appLogger logger.Logger, schemaName ...string) error {
	if len(schemaName) > 0 && schemaName[0] != "" {
		return RunTenantMigrations(db, appLogger, schemaName[0])
	}
	return RunPublicMigrations(db, appLogger)
}
