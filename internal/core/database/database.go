package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sahabatharianmu/OpenMind/config"
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

// RunMigrations runs database migrations using golang-migrate.
func RunMigrations(db *gorm.DB, appLogger logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("Failed to get sql.DB from gorm.DB", zap.Error(err))
		return err
	}

	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		appLogger.Error("Failed to create postgres driver", zap.Error(err))
		return err
	}
	migrationsFS := migrations.GetMigrationsFS()
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

	if appLogger != nil {
		appLogger.Info("Running database migrations...")
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		appLogger.Error("Failed to run migrations", zap.Error(err))
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		if appLogger != nil {
			appLogger.Info("No new migrations to apply")
		}
	} else {
		if appLogger != nil {
			appLogger.Info("Migrations completed successfully")
		}
	}

	return nil
}
