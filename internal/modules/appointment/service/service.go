package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type AppointmentService interface {
	Create(
		ctx context.Context,
		req dto.CreateAppointmentRequest,
		organizationID uuid.UUID,
	) (*dto.AppointmentResponse, error)
	Update(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req dto.UpdateAppointmentRequest) (*dto.AppointmentResponse, error)
	Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*dto.AppointmentResponse, error)
	List(ctx context.Context, organizationID uuid.UUID, page, pageSize int) ([]dto.AppointmentResponse, int64, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type appointmentService struct {
	repo repository.AppointmentRepository
	log  logger.Logger
}

func NewAppointmentService(repo repository.AppointmentRepository, log logger.Logger) AppointmentService {
	return &appointmentService{
		repo: repo,
		log:  log,
	}
}

func (s *appointmentService) Create(
	ctx context.Context,
	req dto.CreateAppointmentRequest,
	organizationID uuid.UUID,
) (*dto.AppointmentResponse, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, err
	}

	status := "scheduled"
	if req.Status != "" {
		status = req.Status
	}

	appointment := &entity.Appointment{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		PatientID:      req.PatientID,
		ClinicianID:    req.ClinicianID,
		StartTime:      startTime,
		EndTime:        endTime,
		Status:         status,
		Type:           req.Type,
		Mode:           req.Mode,
		Notes:          req.Notes,
	}

	if err := s.repo.Create(appointment); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(appointment), nil
}

func (s *appointmentService) Update(
	ctx context.Context,
	id uuid.UUID,
	organizationID uuid.UUID,
	req dto.UpdateAppointmentRequest,
) (*dto.AppointmentResponse, error) {
	appointment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if appointment.OrganizationID != organizationID {
		return nil, response.NewNotFound("Appointment not found")
	}

	if req.StartTime != nil {
		startTime, err := time.Parse(time.RFC3339, *req.StartTime)
		if err != nil {
			return nil, err
		}
		appointment.StartTime = startTime
	}
	if req.EndTime != nil {
		endTime, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			return nil, err
		}
		appointment.EndTime = endTime
	}
	if req.Status != "" {
		appointment.Status = req.Status
	}
	if req.Type != "" {
		appointment.Type = req.Type
	}
	if req.Mode != "" {
		appointment.Mode = req.Mode
	}
	if req.Notes != nil {
		appointment.Notes = req.Notes
	}

	if err := s.repo.Update(appointment); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(appointment), nil
}

func (s *appointmentService) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	appointment, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if appointment.OrganizationID != organizationID {
		return response.ErrNotFound
	}

	return s.repo.Delete(id)
}

func (s *appointmentService) Get(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*dto.AppointmentResponse, error) {
	appointment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if appointment.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	return s.mapEntityToResponse(appointment), nil
}

func (s *appointmentService) List(
	ctx context.Context,
	organizationID uuid.UUID,
	page, pageSize int,
) ([]dto.AppointmentResponse, int64, error) {
	offset := (page - 1) * pageSize
	appointments, total, err := s.repo.List(organizationID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.AppointmentResponse
	for _, a := range appointments {
		responses = append(responses, *s.mapEntityToResponse(&a))
	}

	return responses, total, nil
}

func (s *appointmentService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetOrganizationID(userID)
}

func (s *appointmentService) mapEntityToResponse(a *entity.Appointment) *dto.AppointmentResponse {
	return &dto.AppointmentResponse{
		ID:             a.ID,
		OrganizationID: a.OrganizationID,
		PatientID:      a.PatientID,
		ClinicianID:    a.ClinicianID,
		StartTime:      a.StartTime,
		EndTime:        a.EndTime,
		Status:         a.Status,
		Type:           a.Type,
		Mode:           a.Mode,
		Notes:          a.Notes,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}
