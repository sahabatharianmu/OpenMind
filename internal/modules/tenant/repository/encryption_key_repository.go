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

type TenantEncryptionKeyRepository interface {
	Create(key *entity.TenantEncryptionKey) error
	GetByTenantID(tenantID uuid.UUID) (*entity.TenantEncryptionKey, error)
	GetByOrganizationID(organizationID uuid.UUID) (*entity.TenantEncryptionKey, error)
	Update(key *entity.TenantEncryptionKey) error
	Delete(tenantID uuid.UUID) error
}

type tenantEncryptionKeyRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewTenantEncryptionKeyRepository(db *gorm.DB, log logger.Logger) TenantEncryptionKeyRepository {
	return &tenantEncryptionKeyRepository{
		db:  db,
		log: log,
	}
}

func (r *tenantEncryptionKeyRepository) Create(key *entity.TenantEncryptionKey) error {
	if err := r.db.Create(key).Error; err != nil {
		r.log.Error("Failed to create tenant encryption key", zap.Error(err), zap.String("tenant_id", key.TenantID.String()))
		return err
	}
	return nil
}

func (r *tenantEncryptionKeyRepository) GetByTenantID(tenantID uuid.UUID) (*entity.TenantEncryptionKey, error) {
	var key entity.TenantEncryptionKey
	if err := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("encryption key not found for tenant %s", tenantID.String())
		}
		r.log.Error("Failed to get encryption key by tenant ID", zap.Error(err), zap.String("tenant_id", tenantID.String()))
		return nil, err
	}
	return &key, nil
}

func (r *tenantEncryptionKeyRepository) GetByOrganizationID(organizationID uuid.UUID) (*entity.TenantEncryptionKey, error) {
	var key entity.TenantEncryptionKey
	if err := r.db.Where("organization_id = ? AND deleted_at IS NULL", organizationID).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("encryption key not found for organization %s", organizationID.String())
		}
		r.log.Error("Failed to get encryption key by organization ID", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, err
	}
	return &key, nil
}

func (r *tenantEncryptionKeyRepository) Update(key *entity.TenantEncryptionKey) error {
	if err := r.db.Save(key).Error; err != nil {
		r.log.Error("Failed to update tenant encryption key", zap.Error(err), zap.String("id", key.ID.String()))
		return err
	}
	return nil
}

func (r *tenantEncryptionKeyRepository) Delete(tenantID uuid.UUID) error {
	if err := r.db.Delete(&entity.TenantEncryptionKey{}, "tenant_id = ?", tenantID).Error; err != nil {
		r.log.Error("Failed to delete tenant encryption key", zap.Error(err), zap.String("tenant_id", tenantID.String()))
		return err
	}
	return nil
}

