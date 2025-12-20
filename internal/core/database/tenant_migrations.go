package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/migrations"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunTenantMigrations runs migrations for a specific tenant schema
func RunTenantMigrations(db *gorm.DB, schemaName string, appLogger logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("Failed to get sql.DB from gorm.DB", zap.Error(err))
		return err
	}

	// Set search_path to the tenant schema
	_, err = sqlDB.Exec(fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		appLogger.Error("Failed to set search_path", zap.Error(err), zap.String("schema_name", schemaName))
		return err
	}

	// Create a new connection string with search_path set
	// We need to use a driver that supports schema-specific migrations
	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{
		DatabaseName:    "", // Will use current database
		SchemaName:      schemaName,
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		appLogger.Error("Failed to create postgres driver for tenant", zap.Error(err), zap.String("schema_name", schemaName))
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
		appLogger.Error("Failed to create migrate instance for tenant", zap.Error(err), zap.String("schema_name", schemaName))
		return err
	}

	appLogger.Info("Running tenant migrations", zap.String("schema_name", schemaName))
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		appLogger.Error("Failed to run tenant migrations", zap.Error(err), zap.String("schema_name", schemaName))
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		appLogger.Info("No new migrations to apply for tenant", zap.String("schema_name", schemaName))
	} else {
		appLogger.Info("Tenant migrations completed successfully", zap.String("schema_name", schemaName))
	}

	return nil
}

// CreateTenantSchemaTables creates all required tables in a tenant schema
// This is a simplified approach that creates tables directly in the schema
func CreateTenantSchemaTables(ctx context.Context, db *gorm.DB, schemaName string, appLogger logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set search_path to the tenant schema
	_, err = sqlDB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	// Create tables in the tenant schema
	// We'll use the same table definitions but in the tenant schema
	tables := []string{
		// Users table (if needed per tenant, otherwise keep in public)
		// For now, we'll keep users in public schema and only tenant-specific data in tenant schemas
		`
		CREATE TABLE IF NOT EXISTS patients (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID NOT NULL,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			date_of_birth DATE NOT NULL,
			email VARCHAR(255),
			phone VARCHAR(50),
			address TEXT,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			created_by UUID NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS appointments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID NOT NULL,
			patient_id UUID NOT NULL,
			clinician_id UUID NOT NULL,
			start_time TIMESTAMP WITH TIME ZONE NOT NULL,
			end_time TIMESTAMP WITH TIME ZONE NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'scheduled',
			appointment_type VARCHAR(100) NOT NULL,
			mode VARCHAR(50) NOT NULL,
			cpt_code VARCHAR(20),
			notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS clinical_notes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID NOT NULL,
			patient_id UUID NOT NULL,
			clinician_id UUID NOT NULL,
			appointment_id UUID,
			note_type VARCHAR(100) NOT NULL,
			icd10_code VARCHAR(20),
			subjective TEXT,
			objective TEXT,
			assessment TEXT,
			plan TEXT,
			content_encrypted BYTEA,
			key_id VARCHAR(255),
			nonce BYTEA,
			is_signed BOOLEAN NOT NULL DEFAULT FALSE,
			signed_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS invoices (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID NOT NULL,
			patient_id UUID NOT NULL,
			appointment_id UUID,
			amount_cents INTEGER NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			due_date TIMESTAMP WITH TIME ZONE,
			paid_at TIMESTAMP WITH TIME ZONE,
			payment_method VARCHAR(50),
			notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS assigned_clinicians (
			patient_id UUID NOT NULL,
			clinician_id UUID NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'primary',
			assigned_by UUID NOT NULL,
			assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (patient_id, clinician_id)
		);
		`,
		`
		-- Add foreign key constraints separately to ensure they reference the correct schema
		-- Drop existing constraint if it exists (might be from wrong schema) and recreate
		DO $$
		BEGIN
			-- Drop constraint if it exists in current schema
			IF EXISTS (
				SELECT 1 FROM pg_constraint c
				JOIN pg_namespace n ON n.oid = c.connamespace
				WHERE c.conname = 'fk_assigned_clinicians_patient'
				AND n.nspname = current_schema()
			) THEN
				ALTER TABLE assigned_clinicians DROP CONSTRAINT IF EXISTS fk_assigned_clinicians_patient;
			END IF;
			
			-- Add constraint referencing current schema's patients table
			ALTER TABLE assigned_clinicians 
			ADD CONSTRAINT fk_assigned_clinicians_patient 
			FOREIGN KEY (patient_id) 
			REFERENCES patients(id) 
			ON DELETE CASCADE;
		END $$;
		`,
		`
		DO $$
		BEGIN
			-- Drop constraint if it exists in current schema
			IF EXISTS (
				SELECT 1 FROM pg_constraint c
				JOIN pg_namespace n ON n.oid = c.connamespace
				WHERE c.conname = 'fk_assigned_clinicians_clinician'
				AND n.nspname = current_schema()
			) THEN
				ALTER TABLE assigned_clinicians DROP CONSTRAINT IF EXISTS fk_assigned_clinicians_clinician;
			END IF;
			
			-- Add constraint referencing public.users
			ALTER TABLE assigned_clinicians 
			ADD CONSTRAINT fk_assigned_clinicians_clinician 
			FOREIGN KEY (clinician_id) 
			REFERENCES public.users(id) 
			ON DELETE CASCADE;
		END $$;
		`,
		`
		DO $$
		BEGIN
			-- Drop constraint if it exists in current schema
			IF EXISTS (
				SELECT 1 FROM pg_constraint c
				JOIN pg_namespace n ON n.oid = c.connamespace
				WHERE c.conname = 'fk_assigned_clinicians_assigned_by'
				AND n.nspname = current_schema()
			) THEN
				ALTER TABLE assigned_clinicians DROP CONSTRAINT IF EXISTS fk_assigned_clinicians_assigned_by;
			END IF;
			
			-- Add constraint referencing public.users
			ALTER TABLE assigned_clinicians 
			ADD CONSTRAINT fk_assigned_clinicians_assigned_by 
			FOREIGN KEY (assigned_by) 
			REFERENCES public.users(id) 
			ON DELETE SET NULL;
		END $$;
		`,
		`
		CREATE INDEX IF NOT EXISTS idx_assigned_clinicians_clinician_id 
			ON assigned_clinicians(clinician_id);
		`,
		`
		CREATE INDEX IF NOT EXISTS idx_assigned_clinicians_patient_id 
			ON assigned_clinicians(patient_id);
		`,
		`
		ALTER TABLE assigned_clinicians ADD CONSTRAINT check_assignment_role 
			CHECK (role IN ('primary', 'secondary'));
		`,
		`
		CREATE TABLE IF NOT EXISTS patient_handoffs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			patient_id UUID NOT NULL,
			requesting_clinician_id UUID NOT NULL,
			receiving_clinician_id UUID NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'requested',
			requested_role VARCHAR(50),
			message TEXT,
			requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			responded_at TIMESTAMP WITH TIME ZONE,
			responded_by UUID,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT check_handoff_status 
				CHECK (status IN ('requested', 'approved', 'rejected', 'cancelled')),
			CONSTRAINT check_handoff_not_self 
				CHECK (requesting_clinician_id != receiving_clinician_id)
		);
		`,
		`
		-- Add foreign key constraints for patient_handoffs
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint c
				JOIN pg_namespace n ON n.oid = c.connamespace
				WHERE c.conname = 'fk_patient_handoffs_patient'
				AND n.nspname = current_schema()
			) THEN
				ALTER TABLE patient_handoffs 
				ADD CONSTRAINT fk_patient_handoffs_patient 
				FOREIGN KEY (patient_id) 
				REFERENCES patients(id) 
				ON DELETE CASCADE;
			END IF;
		END $$;
		`,
		`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint c
				JOIN pg_namespace n ON n.oid = c.connamespace
				WHERE c.conname = 'fk_patient_handoffs_requesting'
				AND n.nspname = current_schema()
			) THEN
				ALTER TABLE patient_handoffs 
				ADD CONSTRAINT fk_patient_handoffs_requesting 
				FOREIGN KEY (requesting_clinician_id) 
				REFERENCES public.users(id) 
				ON DELETE CASCADE;
			END IF;
		END $$;
		`,
		`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint c
				JOIN pg_namespace n ON n.oid = c.connamespace
				WHERE c.conname = 'fk_patient_handoffs_receiving'
				AND n.nspname = current_schema()
			) THEN
				ALTER TABLE patient_handoffs 
				ADD CONSTRAINT fk_patient_handoffs_receiving 
				FOREIGN KEY (receiving_clinician_id) 
				REFERENCES public.users(id) 
				ON DELETE CASCADE;
			END IF;
		END $$;
		`,
		`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint c
				JOIN pg_namespace n ON n.oid = c.connamespace
				WHERE c.conname = 'fk_patient_handoffs_responded_by'
				AND n.nspname = current_schema()
			) THEN
				ALTER TABLE patient_handoffs 
				ADD CONSTRAINT fk_patient_handoffs_responded_by 
				FOREIGN KEY (responded_by) 
				REFERENCES public.users(id) 
				ON DELETE SET NULL;
			END IF;
		END $$;
		`,
		`
		CREATE INDEX IF NOT EXISTS idx_patient_handoffs_patient_id 
			ON patient_handoffs(patient_id);
		`,
		`
		CREATE INDEX IF NOT EXISTS idx_patient_handoffs_requesting_clinician_id 
			ON patient_handoffs(requesting_clinician_id);
		`,
		`
		CREATE INDEX IF NOT EXISTS idx_patient_handoffs_receiving_clinician_id 
			ON patient_handoffs(receiving_clinician_id);
		`,
		`
		CREATE INDEX IF NOT EXISTS idx_patient_handoffs_status 
			ON patient_handoffs(status);
		`,
	}

	for _, tableSQL := range tables {
		_, err = sqlDB.ExecContext(ctx, tableSQL)
		if err != nil {
			appLogger.Error("Failed to create table in tenant schema", zap.Error(err), zap.String("schema_name", schemaName))
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	appLogger.Info("Tenant schema tables created successfully", zap.String("schema_name", schemaName))
	return nil
}

// EnsurePatientHandoffsTable ensures the patient_handoffs table exists in a tenant schema
// This is needed when rolling out new features to existing tenant schemas
func EnsurePatientHandoffsTable(ctx context.Context, db *gorm.DB, schemaName string, appLogger logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set search_path to the tenant schema
	_, err = sqlDB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	// Check if table exists
	var tableExists bool
	checkSQL := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = '%s' 
			AND table_name = 'patient_handoffs'
		)
	`, schemaName)
	
	if err := sqlDB.QueryRowContext(ctx, checkSQL).Scan(&tableExists); err != nil {
		appLogger.Error("Failed to check if patient_handoffs table exists", zap.Error(err), zap.String("schema_name", schemaName))
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists {
		appLogger.Debug("patient_handoffs table already exists", zap.String("schema_name", schemaName))
		return nil
	}

	// Table doesn't exist, create it with all constraints
	appLogger.Info("Creating patient_handoffs table in existing tenant schema", zap.String("schema_name", schemaName))
	
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS patient_handoffs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			patient_id UUID NOT NULL,
			requesting_clinician_id UUID NOT NULL,
			receiving_clinician_id UUID NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'requested',
			requested_role VARCHAR(50),
			message TEXT,
			requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			responded_at TIMESTAMP WITH TIME ZONE,
			responded_by UUID,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT check_handoff_status 
				CHECK (status IN ('requested', 'approved', 'rejected', 'cancelled')),
			CONSTRAINT check_handoff_not_self 
				CHECK (requesting_clinician_id != receiving_clinician_id)
		);
	`
	
	if _, err := sqlDB.ExecContext(ctx, createTableSQL); err != nil {
		appLogger.Error("Failed to create patient_handoffs table", zap.Error(err), zap.String("schema_name", schemaName))
		return fmt.Errorf("failed to create patient_handoffs table: %w", err)
	}

	// Add foreign key constraints
	constraints := []string{
		`ALTER TABLE patient_handoffs 
			ADD CONSTRAINT fk_patient_handoffs_patient 
			FOREIGN KEY (patient_id) 
			REFERENCES patients(id) 
			ON DELETE CASCADE;`,
		`ALTER TABLE patient_handoffs 
			ADD CONSTRAINT fk_patient_handoffs_requesting 
			FOREIGN KEY (requesting_clinician_id) 
			REFERENCES public.users(id) 
			ON DELETE CASCADE;`,
		`ALTER TABLE patient_handoffs 
			ADD CONSTRAINT fk_patient_handoffs_receiving 
			FOREIGN KEY (receiving_clinician_id) 
			REFERENCES public.users(id) 
			ON DELETE CASCADE;`,
		`ALTER TABLE patient_handoffs 
			ADD CONSTRAINT fk_patient_handoffs_responded_by 
			FOREIGN KEY (responded_by) 
			REFERENCES public.users(id) 
			ON DELETE SET NULL;`,
	}

	for _, constraintSQL := range constraints {
		if _, err := sqlDB.ExecContext(ctx, constraintSQL); err != nil {
			// Check if constraint already exists
			if !strings.Contains(err.Error(), "already exists") {
				appLogger.Warn("Failed to add constraint to patient_handoffs", zap.Error(err), zap.String("schema_name", schemaName))
				// Continue with other constraints
			}
		}
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_patient_handoffs_patient_id ON patient_handoffs(patient_id);`,
		`CREATE INDEX IF NOT EXISTS idx_patient_handoffs_requesting_clinician_id ON patient_handoffs(requesting_clinician_id);`,
		`CREATE INDEX IF NOT EXISTS idx_patient_handoffs_receiving_clinician_id ON patient_handoffs(receiving_clinician_id);`,
		`CREATE INDEX IF NOT EXISTS idx_patient_handoffs_status ON patient_handoffs(status);`,
	}

	for _, indexSQL := range indexes {
		if _, err := sqlDB.ExecContext(ctx, indexSQL); err != nil {
			appLogger.Warn("Failed to create index on patient_handoffs", zap.Error(err), zap.String("schema_name", schemaName))
			// Continue with other indexes
		}
	}

	appLogger.Info("patient_handoffs table created successfully", zap.String("schema_name", schemaName))
	return nil
}

// FixAssignedCliniciansConstraints fixes foreign key constraints in existing tenant schemas
// This is needed if constraints were created incorrectly (e.g., referencing wrong schema)
func FixAssignedCliniciansConstraints(ctx context.Context, db *gorm.DB, schemaName string, appLogger logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set search_path to the tenant schema
	_, err = sqlDB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	// Fix foreign key constraints
	fixSQL := `
		DO $$
		BEGIN
			-- Drop and recreate patient foreign key
			ALTER TABLE assigned_clinicians DROP CONSTRAINT IF EXISTS fk_assigned_clinicians_patient;
			ALTER TABLE assigned_clinicians 
			ADD CONSTRAINT fk_assigned_clinicians_patient 
			FOREIGN KEY (patient_id) 
			REFERENCES patients(id) 
			ON DELETE CASCADE;
			
			-- Drop and recreate clinician foreign key
			ALTER TABLE assigned_clinicians DROP CONSTRAINT IF EXISTS fk_assigned_clinicians_clinician;
			ALTER TABLE assigned_clinicians 
			ADD CONSTRAINT fk_assigned_clinicians_clinician 
			FOREIGN KEY (clinician_id) 
			REFERENCES public.users(id) 
			ON DELETE CASCADE;
			
			-- Drop and recreate assigned_by foreign key
			ALTER TABLE assigned_clinicians DROP CONSTRAINT IF EXISTS fk_assigned_clinicians_assigned_by;
			ALTER TABLE assigned_clinicians 
			ADD CONSTRAINT fk_assigned_clinicians_assigned_by 
			FOREIGN KEY (assigned_by) 
			REFERENCES public.users(id) 
			ON DELETE SET NULL;
		END $$;
	`

	_, err = sqlDB.ExecContext(ctx, fixSQL)
	if err != nil {
		appLogger.Error("Failed to fix assigned_clinicians constraints", zap.Error(err), zap.String("schema_name", schemaName))
		return fmt.Errorf("failed to fix constraints: %w", err)
	}

	appLogger.Info("Fixed assigned_clinicians constraints", zap.String("schema_name", schemaName))
	return nil
}

// CopyDataToTenantSchema copies existing data from public schema to tenant schema
func CopyDataToTenantSchema(ctx context.Context, db *gorm.DB, schemaName string, organizationID string, appLogger logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Tables to copy (only organization-specific data)
	tables := []string{"patients", "appointments", "clinical_notes", "invoices"}

	for _, tableName := range tables {
		// Copy data from public schema to tenant schema
		copySQL := fmt.Sprintf(`
			INSERT INTO %s.%s 
			SELECT * FROM public.%s 
			WHERE organization_id = $1
			ON CONFLICT (id) DO NOTHING
		`, schemaName, tableName, tableName)

		result, err := sqlDB.ExecContext(ctx, copySQL, organizationID)
		if err != nil {
			appLogger.Warn("Failed to copy data to tenant schema",
				zap.Error(err),
				zap.String("schema_name", schemaName),
				zap.String("table", tableName))
			// Continue with other tables even if one fails
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		appLogger.Info("Copied data to tenant schema",
			zap.String("schema_name", schemaName),
			zap.String("table", tableName),
			zap.Int64("rows", rowsAffected))
	}

	return nil
}

// GetTenantDB returns a GORM DB instance scoped to a tenant schema
func GetTenantDB(schemaName string) *gorm.DB {
	if schemaName == "" {
		return DB
	}
	return WithSchema(DB, schemaName)
}

// ExecuteInSchema executes a function with a specific schema set
func ExecuteInSchema(db *gorm.DB, schemaName string, fn func(*gorm.DB) error) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Save current search_path
	var currentPath string
	err = sqlDB.QueryRow("SHOW search_path").Scan(&currentPath)
	if err != nil {
		return fmt.Errorf("failed to get current search_path: %w", err)
	}

	// Set new search_path
	_, err = sqlDB.Exec(fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	// Execute function
	err = fn(db)

	// Restore original search_path
	_, restoreErr := sqlDB.Exec(fmt.Sprintf("SET search_path TO %s", currentPath))
	if restoreErr != nil {
		// Log but don't fail
		if err == nil {
			err = fmt.Errorf("failed to restore search_path: %w", restoreErr)
		}
	}

	return err
}
