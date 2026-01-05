package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/payment/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentMethodRepository defines the interface for payment method data operations
type PaymentMethodRepository interface {
	Create(paymentMethod *entity.PaymentMethod) error
	GetByID(id uuid.UUID) (*entity.PaymentMethod, error)
	GetByOrganizationID(organizationID uuid.UUID) ([]entity.PaymentMethod, error)
	GetDefaultByOrganizationID(organizationID uuid.UUID) (*entity.PaymentMethod, error)
	Update(paymentMethod *entity.PaymentMethod) error
	Delete(id uuid.UUID) error
	SetDefault(organizationID, paymentMethodID uuid.UUID) error
	CountByOrganizationID(organizationID uuid.UUID) (int64, error)
}

// paymentMethodRepository implements PaymentMethodRepository
type paymentMethodRepository struct {
	db  *gorm.DB
	log logger.Logger
}

// NewPaymentMethodRepository creates a new payment method repository
func NewPaymentMethodRepository(db *gorm.DB, log logger.Logger) PaymentMethodRepository {
	return &paymentMethodRepository{
		db:  db,
		log: log,
	}
}

// Create creates a new payment method
func (r *paymentMethodRepository) Create(paymentMethod *entity.PaymentMethod) error {
	if err := r.db.Create(paymentMethod).Error; err != nil {
		r.log.Error(
			"Failed to create payment method",
			zap.Error(err),
			zap.String("organization_id", paymentMethod.OrganizationID.String()),
		)
		return err
	}
	return nil
}

// GetByID retrieves a payment method by ID
func (r *paymentMethodRepository) GetByID(id uuid.UUID) (*entity.PaymentMethod, error) {
	var paymentMethod entity.PaymentMethod
	if err := r.db.Where("id = ?", id).First(&paymentMethod).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Error("Failed to get payment method by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, err
	}
	return &paymentMethod, nil
}

// GetByOrganizationID retrieves all payment methods for an organization
func (r *paymentMethodRepository) GetByOrganizationID(organizationID uuid.UUID) ([]entity.PaymentMethod, error) {
	var paymentMethods []entity.PaymentMethod
	if err := r.db.Where("organization_id = ?", organizationID).
		Order("is_default DESC, created_at DESC").
		Find(&paymentMethods).Error; err != nil {
		r.log.Error(
			"Failed to get payment methods by organization ID",
			zap.Error(err),
			zap.String("organization_id", organizationID.String()),
		)
		return nil, err
	}
	return paymentMethods, nil
}

// GetDefaultByOrganizationID retrieves the default payment method for an organization
func (r *paymentMethodRepository) GetDefaultByOrganizationID(organizationID uuid.UUID) (*entity.PaymentMethod, error) {
	var paymentMethod entity.PaymentMethod
	if err := r.db.Where("organization_id = ? AND is_default = ?", organizationID, true).
		First(&paymentMethod).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Error(
			"Failed to get default payment method",
			zap.Error(err),
			zap.String("organization_id", organizationID.String()),
		)
		return nil, err
	}
	return &paymentMethod, nil
}

// Update updates a payment method
func (r *paymentMethodRepository) Update(paymentMethod *entity.PaymentMethod) error {
	if err := r.db.Save(paymentMethod).Error; err != nil {
		r.log.Error("Failed to update payment method", zap.Error(err), zap.String("id", paymentMethod.ID.String()))
		return err
	}
	return nil
}

// Delete soft deletes a payment method
func (r *paymentMethodRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entity.PaymentMethod{}, id).Error; err != nil {
		r.log.Error("Failed to delete payment method", zap.Error(err), zap.String("id", id.String()))
		return err
	}
	return nil
}

// SetDefault sets a payment method as the default for an organization
// This will unset any existing default payment method for the organization
func (r *paymentMethodRepository) SetDefault(organizationID, paymentMethodID uuid.UUID) error {
	// Start a transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Unset all existing default payment methods for this organization
	if err := tx.Model(&entity.PaymentMethod{}).
		Where("organization_id = ? AND is_default = ?", organizationID, true).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		r.log.Error(
			"Failed to unset existing default payment methods",
			zap.Error(err),
			zap.String("organization_id", organizationID.String()),
		)
		return err
	}

	// Set the new default payment method
	if err := tx.Model(&entity.PaymentMethod{}).
		Where("id = ? AND organization_id = ?", paymentMethodID, organizationID).
		Update("is_default", true).Error; err != nil {
		tx.Rollback()
		r.log.Error(
			"Failed to set default payment method",
			zap.Error(err),
			zap.String("payment_method_id", paymentMethodID.String()),
		)
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Failed to commit transaction for setting default payment method", zap.Error(err))
		return err
	}

	return nil
}

// CountByOrganizationID counts payment methods for an organization
func (r *paymentMethodRepository) CountByOrganizationID(organizationID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&entity.PaymentMethod{}).
		Where("organization_id = ?", organizationID).
		Count(&count).Error; err != nil {
		r.log.Error(
			"Failed to count payment methods",
			zap.Error(err),
			zap.String("organization_id", organizationID.String()),
		)
		return 0, err
	}
	return count, nil
}
