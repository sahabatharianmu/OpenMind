package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/repository"
	patientRepo "github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	userRepo "github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type AppointmentService interface {
	Create(
		ctx context.Context,
		req dto.CreateAppointmentRequest,
		organizationID uuid.UUID,
	) (*dto.AppointmentResponse, error)
	Update(
		ctx context.Context,
		id uuid.UUID,
		organizationID uuid.UUID,
		req dto.UpdateAppointmentRequest,
	) (*dto.AppointmentResponse, error)
	Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, userID uuid.UUID, userRole string) (*dto.AppointmentResponse, error)
	List(ctx context.Context, organizationID uuid.UUID, page, pageSize int, userID uuid.UUID, userRole string) ([]dto.AppointmentResponse, int64, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type appointmentService struct {
	repo        repository.AppointmentRepository
	patientRepo patientRepo.PatientRepository
	userRepo    userRepo.UserRepository
	log         logger.Logger
}

func NewAppointmentService(repo repository.AppointmentRepository, patientRepo patientRepo.PatientRepository, userRepo userRepo.UserRepository, log logger.Logger) AppointmentService {
	return &appointmentService{
		repo:        repo,
		patientRepo: patientRepo,
		userRepo:    userRepo,
		log:         log,
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

	// Conflict Detection
	overlap, err := s.repo.CheckOverlap(organizationID, req.ClinicianID, startTime, endTime, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check for schedule conflicts: %w", err)
	}
	if overlap {
		return nil, response.NewConflict(
			"Scheduling conflict: This clinician already has an appointment during this time.",
		)
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

	// Conflict Detection for Update
	overlap, err := s.repo.CheckOverlap(
		organizationID,
		appointment.ClinicianID,
		appointment.StartTime,
		appointment.EndTime,
		&id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check for schedule conflicts: %w", err)
	}
	if overlap {
		return nil, response.NewConflict(
			"Scheduling conflict: This clinician already has an appointment during this time.",
		)
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

func (s *appointmentService) Get(
	ctx context.Context,
	id uuid.UUID,
	organizationID uuid.UUID,
	userID uuid.UUID,
	userRole string,
) (*dto.AppointmentResponse, error) {
	appointment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if appointment.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	// Check access control: admin/owner can see all, others only appointments for assigned patients
	if userRole != constants.RoleAdmin && userRole != constants.RoleOwner {
		isAssigned, err := s.patientRepo.IsPatientAssignedToClinician(appointment.PatientID, userID)
		if err != nil {
			s.log.Error("Failed to check patient assignment", zap.Error(err))
			return nil, fmt.Errorf("failed to check access: %w", err)
		}
		if !isAssigned {
			return nil, response.NewForbidden("You can only view appointments for patients you are assigned to")
		}
	}

	return s.mapEntityToResponse(appointment), nil
}

func (s *appointmentService) List(
	ctx context.Context,
	organizationID uuid.UUID,
	page, pageSize int,
	userID uuid.UUID,
	userRole string,
) ([]dto.AppointmentResponse, int64, error) {
	offset := (page - 1) * pageSize
	appointments, total, err := s.repo.List(organizationID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Filter appointments by patient assignment for non-admin users
	var filteredAppointments []entity.Appointment
	if userRole != constants.RoleAdmin && userRole != constants.RoleOwner {
		// Get assigned patient IDs for this user
		assignedPatientIDs, err := s.patientRepo.GetAssignedPatients(userID, organizationID)
		if err != nil {
			s.log.Error("Failed to get assigned patients", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to get assigned patients: %w", err)
		}

		// Create a map for quick lookup
		assignedMap := make(map[uuid.UUID]bool)
		for _, pid := range assignedPatientIDs {
			assignedMap[pid] = true
		}

		// Filter appointments to only those for assigned patients
		for i := range appointments {
			if assignedMap[appointments[i].PatientID] {
				filteredAppointments = append(filteredAppointments, appointments[i])
			}
		}
		appointments = filteredAppointments
		total = int64(len(appointments)) // Update total count
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
	// Fetch clinician information
	var clinicianName *string
	var clinicianEmail *string
	
	clinician, err := s.userRepo.GetByID(a.ClinicianID)
	if err == nil && clinician != nil {
		clinicianName = &clinician.FullName
		clinicianEmail = &clinician.Email
	} else {
		s.log.Warn("Failed to fetch clinician information for appointment", 
			zap.String("appointment_id", a.ID.String()),
			zap.String("clinician_id", a.ClinicianID.String()),
			zap.Error(err))
	}

	return &dto.AppointmentResponse{
		ID:             a.ID,
		OrganizationID: a.OrganizationID,
		PatientID:      a.PatientID,
		ClinicianID:    a.ClinicianID,
		ClinicianName:  clinicianName,
		ClinicianEmail: clinicianEmail,
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
