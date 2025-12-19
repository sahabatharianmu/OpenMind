package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/tenant/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(tenant *entity.Tenant) error
	GetByOrganizationID(organizationID uuid.UUID) (*entity.Tenant, error)
	GetBySchemaName(schemaName string) (*entity.Tenant, error)
	GetByID(id uuid.UUID) (*entity.Tenant, error)
	Update(tenant *entity.Tenant) error
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]entity.Tenant, int64, error)
}

type tenantRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewTenantRepository(db *gorm.DB, log logger.Logger) TenantRepository {
	return &tenantRepository{
		db:  db,
		log: log,
	}
}

func (r *tenantRepository) Create(tenant *entity.Tenant) error {
	if err := r.db.Create(tenant).Error; err != nil {
		r.log.Error("Failed to create tenant", zap.Error(err), zap.String("organization_id", tenant.OrganizationID.String()))
		return err
	}
	return nil
}

func (r *tenantRepository) GetByOrganizationID(organizationID uuid.UUID) (*entity.Tenant, error) {
	var tenant entity.Tenant
	if err := r.db.Where("organization_id = ? AND deleted_at IS NULL", organizationID).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found for organization %s", organizationID.String())
		}
		r.log.Error("Failed to get tenant by organization ID", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) GetBySchemaName(schemaName string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	if err := r.db.Where("schema_name = ? AND deleted_at IS NULL", schemaName).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found for schema %s", schemaName)
		}
		r.log.Error("Failed to get tenant by schema name", zap.Error(err), zap.String("schema_name", schemaName))
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) GetByID(id uuid.UUID) (*entity.Tenant, error) {
	var tenant entity.Tenant
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found: %s", id.String())
		}
		r.log.Error("Failed to get tenant by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) Update(tenant *entity.Tenant) error {
	if err := r.db.Save(tenant).Error; err != nil {
		r.log.Error("Failed to update tenant", zap.Error(err), zap.String("id", tenant.ID.String()))
		return err
	}
	return nil
}

func (r *tenantRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entity.Tenant{}, "id = ?", id).Error; err != nil {
		r.log.Error("Failed to delete tenant", zap.Error(err), zap.String("id", id.String()))
		return err
	}
	return nil
}

func (r *tenantRepository) List(limit, offset int) ([]entity.Tenant, int64, error) {
	var tenants []entity.Tenant
	var total int64

	if err := r.db.Model(&entity.Tenant{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		r.log.Error("Failed to count tenants", zap.Error(err))
		return nil, 0, err
	}

	if err := r.db.Where("deleted_at IS NULL").Limit(limit).Offset(offset).Find(&tenants).Error; err != nil {
		r.log.Error("Failed to list tenants", zap.Error(err))
		return nil, 0, err
	}

	return tenants, total, nil
}

