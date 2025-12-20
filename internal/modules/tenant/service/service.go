package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/core/database"
	"github.com/sahabatharianmu/OpenMind/internal/modules/tenant/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/tenant/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TenantService interface {
	CreateTenantForOrganization(ctx context.Context, organizationID uuid.UUID) (*entity.Tenant, error)
	GetTenantByOrganizationID(ctx context.Context, organizationID uuid.UUID) (*entity.Tenant, error)
	GetTenantBySchemaName(ctx context.Context, schemaName string) (*entity.Tenant, error)
	SetSchemaForRequest(ctx context.Context, schemaName string) error
	CreateSchema(ctx context.Context, schemaName string) error
	DropSchema(ctx context.Context, schemaName string) error
	MigrateSchema(ctx context.Context, schemaName string) error
	GenerateSchemaName(organizationID uuid.UUID) string
	SetEncryptionService(encryptionSvc *crypto.EncryptionService)
	SetKeyRepository(keyRepo repository.TenantEncryptionKeyRepository)
	GenerateKeysForExistingTenants(ctx context.Context) error
	EnsureTenantHasKey(ctx context.Context, tenantID, organizationID uuid.UUID) error
}

type tenantService struct {
	repo            repository.TenantRepository
	keyRepo         repository.TenantEncryptionKeyRepository
	encryptionSvc   *crypto.EncryptionService
	log             logger.Logger
	db              *gorm.DB
}

func NewTenantService(
	repo repository.TenantRepository,
	db *gorm.DB,
	log logger.Logger,
) TenantService {
	return &tenantService{
		repo:  repo,
		log:   log,
		db:    db,
	}
}

// SetEncryptionService sets the encryption service for key generation
func (s *tenantService) SetEncryptionService(encryptionSvc *crypto.EncryptionService) {
	s.encryptionSvc = encryptionSvc
}

// SetKeyRepository sets the tenant encryption key repository
func (s *tenantService) SetKeyRepository(keyRepo repository.TenantEncryptionKeyRepository) {
	s.keyRepo = keyRepo
}

// GenerateSchemaName generates a safe schema name from organization ID
func (s *tenantService) GenerateSchemaName(organizationID uuid.UUID) string {
	// Use organization ID as base, remove hyphens, and prefix with 'tenant_'
	// PostgreSQL schema names must be lowercase and can contain underscores
	schemaName := strings.ReplaceAll(organizationID.String(), "-", "")
	return fmt.Sprintf("tenant_%s", schemaName)
}

// CreateTenantForOrganization creates a tenant record and its schema
func (s *tenantService) CreateTenantForOrganization(ctx context.Context, organizationID uuid.UUID) (*entity.Tenant, error) {
	schemaName := s.GenerateSchemaName(organizationID)

	// Check if tenant already exists
	existing, err := s.repo.GetByOrganizationID(organizationID)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Create schema first
	if err := s.CreateSchema(ctx, schemaName); err != nil {
		s.log.Error("Failed to create tenant schema", zap.Error(err), zap.String("schema_name", schemaName))
		return nil, fmt.Errorf("failed to create tenant schema: %w", err)
	}

	// Create tenant record
	tenant := &entity.Tenant{
		OrganizationID: organizationID,
		SchemaName:     schemaName,
		Status:         "active",
	}

	if err := s.repo.Create(tenant); err != nil {
		// If tenant creation fails, try to clean up schema
		_ = s.DropSchema(ctx, schemaName)
		s.log.Error("Failed to create tenant record", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, fmt.Errorf("failed to create tenant record: %w", err)
	}

	// Run migrations for the new schema
	if err := s.MigrateSchema(ctx, schemaName); err != nil {
		s.log.Error("Failed to migrate tenant schema", zap.Error(err), zap.String("schema_name", schemaName))
		// Don't fail the entire operation, but log the error
	}

	// Generate and store encryption key for this tenant (HIPAA compliant)
	if s.encryptionSvc != nil && s.keyRepo != nil {
		if err := s.generateTenantEncryptionKey(ctx, tenant.ID, organizationID); err != nil {
			s.log.Error("Failed to generate tenant encryption key", zap.Error(err),
				zap.String("tenant_id", tenant.ID.String()),
				zap.String("organization_id", organizationID.String()))
			// Don't fail tenant creation if key generation fails, but log the error
			// The tenant can still function with legacy encryption
		} else {
			s.log.Info("Tenant encryption key generated successfully",
				zap.String("tenant_id", tenant.ID.String()),
				zap.String("organization_id", organizationID.String()))
		}
	}

	s.log.Info("Tenant created successfully", zap.String("organization_id", organizationID.String()), zap.String("schema_name", schemaName))
	return tenant, nil
}

// generateTenantEncryptionKey generates and stores an encryption key for a tenant
func (s *tenantService) generateTenantEncryptionKey(ctx context.Context, tenantID, organizationID uuid.UUID) error {
	// Check if key already exists
	existing, err := s.keyRepo.GetByTenantID(tenantID)
	if err == nil && existing != nil {
		s.log.Info("Tenant encryption key already exists", zap.String("tenant_id", tenantID.String()))
		return nil
	}

	// Generate new tenant key (32 bytes, AES-256)
	tenantKey, err := s.encryptionSvc.GenerateTenantKey()
	if err != nil {
		return fmt.Errorf("failed to generate tenant key: %w", err)
	}

	// Encrypt tenant key with master key
	encryptedKey, err := s.encryptionSvc.EncryptTenantKey(tenantKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt tenant key: %w", err)
	}

	// Store encrypted key in database
	keyEntity := &entity.TenantEncryptionKey{
		TenantID:       tenantID,
		OrganizationID: organizationID,
		EncryptedKey:   encryptedKey,
		KeyVersion:     1,
		Algorithm:      "AES-256-GCM",
	}

	if err := s.keyRepo.Create(keyEntity); err != nil {
		return fmt.Errorf("failed to store tenant encryption key: %w", err)
	}

	return nil
}

// GetTenantByOrganizationID retrieves tenant by organization ID
// Also ensures required tables exist for existing tenant schemas
func (s *tenantService) GetTenantByOrganizationID(ctx context.Context, organizationID uuid.UUID) (*entity.Tenant, error) {
	tenant, err := s.repo.GetByOrganizationID(organizationID)
	if err != nil || tenant == nil {
		return tenant, err
	}

	// Ensure patient_handoffs table exists for existing tenant schemas
	// This is needed when rolling out new features to existing tenants
	if err := database.EnsurePatientHandoffsTable(ctx, s.db, tenant.SchemaName, s.log); err != nil {
		s.log.Warn("Failed to ensure patient_handoffs table exists", zap.Error(err), zap.String("schema_name", tenant.SchemaName))
		// Don't fail the operation, but log the warning
	}

	return tenant, nil
}

// GetTenantBySchemaName retrieves tenant by schema name
func (s *tenantService) GetTenantBySchemaName(ctx context.Context, schemaName string) (*entity.Tenant, error) {
	return s.repo.GetBySchemaName(schemaName)
}

// SetSchemaForRequest sets the search_path for the current database connection
func (s *tenantService) SetSchemaForRequest(ctx context.Context, schemaName string) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set search_path for this connection
	// This ensures all queries use the specified schema
	_, err = sqlDB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		s.log.Error("Failed to set search_path", zap.Error(err), zap.String("schema_name", schemaName))
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	return nil
}

// CreateSchema creates a PostgreSQL schema
func (s *tenantService) CreateSchema(ctx context.Context, schemaName string) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Use the create_tenant_schema function from migration
	query := fmt.Sprintf("SELECT create_tenant_schema('%s')", schemaName)
	_, err = sqlDB.ExecContext(ctx, query)
	if err != nil {
		s.log.Error("Failed to create schema", zap.Error(err), zap.String("schema_name", schemaName))
		return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
	}

	s.log.Info("Schema created successfully", zap.String("schema_name", schemaName))
	return nil
}

// DropSchema drops a PostgreSQL schema
func (s *tenantService) DropSchema(ctx context.Context, schemaName string) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Use the drop_tenant_schema function from migration
	query := fmt.Sprintf("SELECT drop_tenant_schema('%s')", schemaName)
	_, err = sqlDB.ExecContext(ctx, query)
	if err != nil {
		s.log.Error("Failed to drop schema", zap.Error(err), zap.String("schema_name", schemaName))
		return fmt.Errorf("failed to drop schema %s: %w", schemaName, err)
	}

	s.log.Info("Schema dropped successfully", zap.String("schema_name", schemaName))
	return nil
}

// MigrateSchema runs migrations for a specific tenant schema
func (s *tenantService) MigrateSchema(ctx context.Context, schemaName string) error {
	// Import the database package to use tenant migration functions
	// This creates all required tables in the tenant schema
	if err := database.CreateTenantSchemaTables(ctx, s.db, schemaName, s.log); err != nil {
		s.log.Error("Failed to create tenant schema tables", zap.Error(err), zap.String("schema_name", schemaName))
		return fmt.Errorf("failed to create tenant schema tables: %w", err)
	}

	// Fix foreign key constraints for assigned_clinicians table (in case they were created incorrectly)
	if err := database.FixAssignedCliniciansConstraints(ctx, s.db, schemaName, s.log); err != nil {
		s.log.Warn("Failed to fix assigned_clinicians constraints", zap.Error(err), zap.String("schema_name", schemaName))
		// Don't fail the migration if constraint fix fails, but log it
	}

	// Ensure patient_handoffs table exists (for existing tenant schemas when rolling out new features)
	if err := database.EnsurePatientHandoffsTable(ctx, s.db, schemaName, s.log); err != nil {
		s.log.Warn("Failed to ensure patient_handoffs table exists", zap.Error(err), zap.String("schema_name", schemaName))
		// Don't fail the migration if table creation fails, but log it
	}

	s.log.Info("Tenant schema migrated successfully", zap.String("schema_name", schemaName))
	return nil
}

// GetSchemaFromContext retrieves the schema name from context
func GetSchemaFromContext(ctx context.Context) (string, bool) {
	if schemaName, ok := ctx.Value("tenant_schema").(string); ok {
		return schemaName, true
	}
	return "", false
}

// SetSchemaInContext sets the schema name in context
func SetSchemaInContext(ctx context.Context, schemaName string) context.Context {
	return context.WithValue(ctx, "tenant_schema", schemaName)
}

