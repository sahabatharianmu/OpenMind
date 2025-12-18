package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/invoice/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	Create(invoice *entity.Invoice) error
	Update(invoice *entity.Invoice) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*entity.Invoice, error)
	List(organizationID uuid.UUID, limit, offset int) ([]entity.Invoice, int64, error)
	GetOrganizationID(userID uuid.UUID) (uuid.UUID, error)
}

type invoiceRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewInvoiceRepository(db *gorm.DB, log logger.Logger) InvoiceRepository {
	return &invoiceRepository{
		db:  db,
		log: log,
	}
}

func (r *invoiceRepository) Create(invoice *entity.Invoice) error {
	if err := r.db.Create(invoice).Error; err != nil {
		r.log.Error("Failed to create invoice", zap.Error(err))
		return err
	}
	return nil
}

func (r *invoiceRepository) Update(invoice *entity.Invoice) error {
	if err := r.db.Save(invoice).Error; err != nil {
		r.log.Error("Failed to update invoice", zap.Error(err), zap.String("id", invoice.ID.String()))
		return err
	}
	return nil
}

func (r *invoiceRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entity.Invoice{}, "id = ?", id).Error; err != nil {
		r.log.Error("Failed to delete invoice", zap.Error(err), zap.String("id", id.String()))
		return err
	}
	return nil
}

func (r *invoiceRepository) FindByID(id uuid.UUID) (*entity.Invoice, error) {
	var invoice entity.Invoice
	if err := r.db.First(&invoice, "id = ?", id).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find invoice", zap.Error(err), zap.String("id", id.String()))
		}
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) List(organizationID uuid.UUID, limit, offset int) ([]entity.Invoice, int64, error) {
	var invoices []entity.Invoice
	var total int64

	query := r.db.Model(&entity.Invoice{}).Where("organization_id = ?", organizationID)

	if err := query.Count(&total).Error; err != nil {
		r.log.Error("Failed to count invoices", zap.Error(err))
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&invoices).Error; err != nil {
		r.log.Error("Failed to list invoices", zap.Error(err))
		return nil, 0, err
	}

	return invoices, total, nil
}

func (r *invoiceRepository) GetOrganizationID(userID uuid.UUID) (uuid.UUID, error) {
	var orgIDStr string
	if err := r.db.Table("organization_members").Select("organization_id").Where("user_id = ?", userID).Limit(1).Scan(&orgIDStr).Error; err != nil {
		r.log.Error("Failed to get organization ID", zap.Error(err), zap.String("user_id", userID.String()))
		return uuid.Nil, err
	}
	if orgIDStr == "" {
		return uuid.Nil, errors.New("organization not found for user")
	}
	return uuid.Parse(orgIDStr)
}
