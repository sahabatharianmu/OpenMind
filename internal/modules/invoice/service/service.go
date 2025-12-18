package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/invoice/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/invoice/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/invoice/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
)

type InvoiceService interface {
	Create(ctx context.Context, req dto.CreateInvoiceRequest, organizationID uuid.UUID) (*dto.InvoiceResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateInvoiceRequest) (*dto.InvoiceResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (*dto.InvoiceResponse, error)
	List(ctx context.Context, organizationID uuid.UUID, page, pageSize int) ([]dto.InvoiceResponse, int64, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type invoiceService struct {
	repo repository.InvoiceRepository
	log  logger.Logger
}

func NewInvoiceService(repo repository.InvoiceRepository, log logger.Logger) InvoiceService {
	return &invoiceService{
		repo: repo,
		log:  log,
	}
}

func (s *invoiceService) Create(ctx context.Context, req dto.CreateInvoiceRequest, organizationID uuid.UUID) (*dto.InvoiceResponse, error) {
	status := "pending"
	if req.Status != "" {
		status = req.Status
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		t, err := s.parseTime(*req.DueDate)
		if err != nil {
			return nil, err
		}
		dueDate = &t
	}

	var paidAt *time.Time
	if req.PaidAt != nil {
		t, err := s.parseTime(*req.PaidAt)
		if err != nil {
			return nil, err
		}
		paidAt = &t
	}

	invoice := &entity.Invoice{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		PatientID:      req.PatientID,
		AppointmentID:  req.AppointmentID,
		AmountCents:    req.AmountCents,
		Status:         status,
		DueDate:        dueDate,
		PaidAt:         paidAt,
		PaymentMethod:  req.PaymentMethod,
		Notes:          req.Notes,
	}

	if err := s.repo.Create(invoice); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(invoice), nil
}

func (s *invoiceService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateInvoiceRequest) (*dto.InvoiceResponse, error) {
	invoice, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.AmountCents != nil {
		invoice.AmountCents = *req.AmountCents
	}
	if req.Status != "" {
		invoice.Status = req.Status
	}
	if req.DueDate != nil {
		t, err := s.parseTime(*req.DueDate)
		if err != nil {
			return nil, err
		}
		invoice.DueDate = &t
	}
	if req.PaidAt != nil {
		t, err := s.parseTime(*req.PaidAt)
		if err != nil {
			return nil, err
		}
		invoice.PaidAt = &t
	}
	if req.PaymentMethod != nil {
		invoice.PaymentMethod = req.PaymentMethod
	}
	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	if err := s.repo.Update(invoice); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(invoice), nil
}

func (s *invoiceService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *invoiceService) Get(ctx context.Context, id uuid.UUID) (*dto.InvoiceResponse, error) {
	invoice, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.mapEntityToResponse(invoice), nil
}

func (s *invoiceService) List(ctx context.Context, organizationID uuid.UUID, page, pageSize int) ([]dto.InvoiceResponse, int64, error) {
	offset := (page - 1) * pageSize
	invoices, total, err := s.repo.List(organizationID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.InvoiceResponse
	for _, i := range invoices {
		responses = append(responses, *s.mapEntityToResponse(&i))
	}

	return responses, total, nil
}

func (s *invoiceService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetOrganizationID(userID)
}

func (s *invoiceService) parseTime(dateStr string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	return time.Parse(time.DateOnly, dateStr)
}

func (s *invoiceService) mapEntityToResponse(i *entity.Invoice) *dto.InvoiceResponse {
	return &dto.InvoiceResponse{
		ID:             i.ID,
		OrganizationID: i.OrganizationID,
		PatientID:      i.PatientID,
		AppointmentID:  i.AppointmentID,
		AmountCents:    i.AmountCents,
		Status:         i.Status,
		DueDate:        i.DueDate,
		PaidAt:         i.PaidAt,
		PaymentMethod:  i.PaymentMethod,
		Notes:          i.Notes,
		CreatedAt:      i.CreatedAt,
		UpdatedAt:      i.UpdatedAt,
	}
}
