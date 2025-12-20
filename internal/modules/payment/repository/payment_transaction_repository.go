package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/payment/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentTransactionRepository defines operations for payment transactions
type PaymentTransactionRepository interface {
	Create(transaction *entity.PaymentTransaction) error
	Update(transaction *entity.PaymentTransaction) error
	FindByID(id uuid.UUID) (*entity.PaymentTransaction, error)
	FindByPartnerReferenceNo(partnerReferenceNo string) (*entity.PaymentTransaction, error)
	FindByProviderTransactionID(providerTransactionID string) (*entity.PaymentTransaction, error)
	ListByOrganizationID(organizationID uuid.UUID, limit, offset int) ([]entity.PaymentTransaction, int64, error)
	UpdateStatus(id uuid.UUID, status string, paidAt *time.Time) error
}

type paymentTransactionRepository struct {
	db  *gorm.DB
	log logger.Logger
}

// NewPaymentTransactionRepository creates a new PaymentTransactionRepository
func NewPaymentTransactionRepository(db *gorm.DB, log logger.Logger) PaymentTransactionRepository {
	return &paymentTransactionRepository{
		db:  db,
		log: log,
	}
}

// Create a new payment transaction
func (r *paymentTransactionRepository) Create(transaction *entity.PaymentTransaction) error {
	if err := r.db.Create(transaction).Error; err != nil {
		r.log.Error("Failed to create payment transaction", zap.Error(err), zap.String("organization_id", transaction.OrganizationID.String()))
		return err
	}
	return nil
}

// Update an existing payment transaction
func (r *paymentTransactionRepository) Update(transaction *entity.PaymentTransaction) error {
	if err := r.db.Save(transaction).Error; err != nil {
		r.log.Error("Failed to update payment transaction", zap.Error(err), zap.String("id", transaction.ID.String()))
		return err
	}
	return nil
}

// FindByID finds a payment transaction by its ID
func (r *paymentTransactionRepository) FindByID(id uuid.UUID) (*entity.PaymentTransaction, error) {
	var transaction entity.PaymentTransaction
	if err := r.db.First(&transaction, "id = ?", id).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find payment transaction", zap.Error(err), zap.String("id", id.String()))
		}
		return nil, err
	}
	return &transaction, nil
}

// FindByPartnerReferenceNo finds a payment transaction by partner reference number
func (r *paymentTransactionRepository) FindByPartnerReferenceNo(partnerReferenceNo string) (*entity.PaymentTransaction, error) {
	var transaction entity.PaymentTransaction
	if err := r.db.Where("partner_reference_no = ?", partnerReferenceNo).First(&transaction).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find payment transaction by partner reference", zap.Error(err), zap.String("partner_reference_no", partnerReferenceNo))
		}
		return nil, err
	}
	return &transaction, nil
}

// FindByProviderTransactionID finds a payment transaction by provider transaction ID
func (r *paymentTransactionRepository) FindByProviderTransactionID(providerTransactionID string) (*entity.PaymentTransaction, error) {
	var transaction entity.PaymentTransaction
	if err := r.db.Where("provider_transaction_id = ?", providerTransactionID).First(&transaction).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find payment transaction by provider transaction ID", zap.Error(err), zap.String("provider_transaction_id", providerTransactionID))
		}
		return nil, err
	}
	return &transaction, nil
}

// ListByOrganizationID lists payment transactions for an organization with pagination
func (r *paymentTransactionRepository) ListByOrganizationID(organizationID uuid.UUID, limit, offset int) ([]entity.PaymentTransaction, int64, error) {
	var transactions []entity.PaymentTransaction
	var total int64

	// Count total
	if err := r.db.Model(&entity.PaymentTransaction{}).Where("organization_id = ?", organizationID).Count(&total).Error; err != nil {
		r.log.Error("Failed to count payment transactions", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.Where("organization_id = ?", organizationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		r.log.Error("Failed to list payment transactions", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, 0, err
	}

	return transactions, total, nil
}

// UpdateStatus updates the status of a payment transaction
func (r *paymentTransactionRepository) UpdateStatus(id uuid.UUID, status string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if paidAt != nil {
		updates["paid_at"] = paidAt
	}

	if err := r.db.Model(&entity.PaymentTransaction{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.log.Error("Failed to update payment transaction status", zap.Error(err), zap.String("id", id.String()), zap.String("status", status))
		return err
	}
	return nil
}

