package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	userRepo "github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type PatientService interface {
	Create(
		ctx context.Context,
		req dto.CreatePatientRequest,
		organizationID, createdBy uuid.UUID,
	) (*dto.PatientResponse, error)
	Update(
		ctx context.Context,
		id uuid.UUID,
		organizationID uuid.UUID,
		req dto.UpdatePatientRequest,
	) (*dto.PatientResponse, error)
	Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, userID uuid.UUID, userRole string) (*dto.PatientResponse, error)
	List(ctx context.Context, organizationID uuid.UUID, page, pageSize int, userID uuid.UUID, userRole string) ([]dto.PatientResponse, int64, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	AssignClinician(ctx context.Context, patientID uuid.UUID, req dto.AssignClinicianRequest, organizationID uuid.UUID, userID uuid.UUID) error
	UnassignClinician(ctx context.Context, patientID, clinicianID uuid.UUID, organizationID uuid.UUID, userID uuid.UUID) error
	GetAssignedClinicians(ctx context.Context, patientID uuid.UUID, organizationID uuid.UUID) ([]dto.ClinicianAssignmentResponse, error)
}

type patientService struct {
	repo     repository.PatientRepository
	userRepo userRepo.UserRepository
	log      logger.Logger
}

func NewPatientService(repo repository.PatientRepository, userRepo userRepo.UserRepository, log logger.Logger) PatientService {
	return &patientService{
		repo:     repo,
		userRepo: userRepo,
		log:      log,
	}
}

func (s *patientService) Create(
	ctx context.Context,
	req dto.CreatePatientRequest,
	organizationID, createdBy uuid.UUID,
) (*dto.PatientResponse, error) {
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, err
	}

	status := "active"
	if req.Status != "" {
		status = req.Status
	}

	patient := &entity.Patient{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    dob,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		Status:         status,
		CreatedBy:      createdBy,
	}

	if err := s.repo.Create(patient); err != nil {
		return nil, err
	}

	// Automatically assign creator as primary clinician
	if err := s.repo.AssignClinician(patient.ID, createdBy, "primary", createdBy); err != nil {
		s.log.Warn("Failed to auto-assign creator as primary clinician", zap.Error(err),
			zap.String("patient_id", patient.ID.String()),
			zap.String("created_by", createdBy.String()))
		// Don't fail patient creation if assignment fails, but log it
	}

	return s.mapEntityToResponse(patient), nil
}

func (s *patientService) Update(
	ctx context.Context,
	id uuid.UUID,
	organizationID uuid.UUID,
	req dto.UpdatePatientRequest,
) (*dto.PatientResponse, error) {
	patient, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if patient.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	if req.FirstName != "" {
		patient.FirstName = req.FirstName
	}
	if req.LastName != "" {
		patient.LastName = req.LastName
	}
	if req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return nil, err
		}
		patient.DateOfBirth = dob
	}
	if req.Email != nil {
		patient.Email = req.Email
	}
	if req.Phone != nil {
		patient.Phone = req.Phone
	}
	if req.Address != nil {
		patient.Address = req.Address
	}
	if req.Status != "" {
		patient.Status = req.Status
	}

	if err := s.repo.Update(patient); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(patient), nil
}

func (s *patientService) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	patient, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if patient.OrganizationID != organizationID {
		return response.ErrNotFound
	}

	return s.repo.Delete(id)
}

func (s *patientService) Get(
	ctx context.Context,
	id uuid.UUID,
	organizationID uuid.UUID,
	userID uuid.UUID,
	userRole string,
) (*dto.PatientResponse, error) {
	patient, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if patient.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	// Check access control: admin/owner can see all, others only assigned patients
	if userRole != constants.RoleAdmin && userRole != constants.RoleOwner {
		isAssigned, err := s.repo.IsPatientAssignedToClinician(id, userID)
		if err != nil {
			s.log.Error("Failed to check patient assignment", zap.Error(err))
			return nil, fmt.Errorf("failed to check access: %w", err)
		}
		if !isAssigned {
			return nil, response.ErrForbidden
		}
	}

	return s.mapEntityToResponse(patient), nil
}

func (s *patientService) List(
	ctx context.Context,
	organizationID uuid.UUID,
	page, pageSize int,
	userID uuid.UUID,
	userRole string,
) ([]dto.PatientResponse, int64, error) {
	offset := (page - 1) * pageSize

	// Get assigned patient IDs for non-admin users
	var assignedPatientIDs []uuid.UUID
	if userRole != constants.RoleAdmin && userRole != constants.RoleOwner {
		assignedIDs, err := s.repo.GetAssignedPatients(userID, organizationID)
		if err != nil {
			s.log.Error("Failed to get assigned patients", zap.Error(err),
				zap.String("user_id", userID.String()))
			return nil, 0, fmt.Errorf("failed to get assigned patients: %w", err)
		}
		assignedPatientIDs = assignedIDs
	}

	patients, total, err := s.repo.List(organizationID, pageSize, offset, assignedPatientIDs)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.PatientResponse
	for _, p := range patients {
		responses = append(responses, *s.mapEntityToResponse(&p))
	}

	return responses, total, nil
}

func (s *patientService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetOrganizationID(userID)
}

func (s *patientService) AssignClinician(
	ctx context.Context,
	patientID uuid.UUID,
	req dto.AssignClinicianRequest,
	organizationID uuid.UUID,
	userID uuid.UUID,
) error {
	// Verify patient exists and belongs to organization
	patient, err := s.repo.FindByID(patientID)
	if err != nil {
		return response.ErrNotFound
	}
	if patient.OrganizationID != organizationID {
		return response.ErrNotFound
	}

	// Verify clinician belongs to same organization (check via organization_members)
	clinicianOrgID, err := s.repo.GetOrganizationID(req.ClinicianID)
	if err != nil || clinicianOrgID != organizationID {
		return response.NewBadRequest("Clinician does not belong to this organization")
	}

	// Check if already assigned
	isAssigned, err := s.repo.IsPatientAssignedToClinician(patientID, req.ClinicianID)
	if err != nil {
		return fmt.Errorf("failed to check assignment: %w", err)
	}
	if isAssigned {
		return response.NewBadRequest("Clinician is already assigned to this patient")
	}

	// Assign clinician
	if err := s.repo.AssignClinician(patientID, req.ClinicianID, req.Role, userID); err != nil {
		s.log.Error("Failed to assign clinician", zap.Error(err))
		return fmt.Errorf("failed to assign clinician: %w", err)
	}

	return nil
}

func (s *patientService) UnassignClinician(
	ctx context.Context,
	patientID, clinicianID uuid.UUID,
	organizationID uuid.UUID,
	userID uuid.UUID,
) error {
	// Verify patient exists and belongs to organization
	patient, err := s.repo.FindByID(patientID)
	if err != nil {
		return response.ErrNotFound
	}
	if patient.OrganizationID != organizationID {
		return response.ErrNotFound
	}

	// Prevent removing last primary clinician
	primaryCount, err := s.repo.CountPrimaryClinicians(patientID)
	if err != nil {
		return fmt.Errorf("failed to count primary clinicians: %w", err)
	}

	// Check if the clinician being removed is primary
	isAssigned, err := s.repo.IsPatientAssignedToClinician(patientID, clinicianID)
	if err != nil {
		return fmt.Errorf("failed to check assignment: %w", err)
	}
	if !isAssigned {
		return response.NewBadRequest("Clinician is not assigned to this patient")
	}

	// Get assignment to check role
	assignments, err := s.repo.GetAssignedClinicians(patientID)
	if err != nil {
		return fmt.Errorf("failed to get assignments: %w", err)
	}

	var isPrimary bool
	for _, assignment := range assignments {
		if assignment.ClinicianID == clinicianID {
			isPrimary = assignment.Role == "primary"
			break
		}
	}

	if isPrimary && primaryCount <= 1 {
		return response.NewBadRequest("Cannot remove the last primary clinician from a patient")
	}

	// Unassign clinician
	if err := s.repo.UnassignClinician(patientID, clinicianID); err != nil {
		s.log.Error("Failed to unassign clinician", zap.Error(err))
		return fmt.Errorf("failed to unassign clinician: %w", err)
	}

	return nil
}

func (s *patientService) GetAssignedClinicians(
	ctx context.Context,
	patientID uuid.UUID,
	organizationID uuid.UUID,
) ([]dto.ClinicianAssignmentResponse, error) {
	// Verify patient exists and belongs to organization
	patient, err := s.repo.FindByID(patientID)
	if err != nil {
		return nil, response.ErrNotFound
	}
	if patient.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	// Get assignments
	assignments, err := s.repo.GetAssignedClinicians(patientID)
	if err != nil {
		s.log.Error("Failed to get assigned clinicians", zap.Error(err))
		return nil, fmt.Errorf("failed to get assigned clinicians: %w", err)
	}

	// Fetch user details for each assignment
	var responses []dto.ClinicianAssignmentResponse
	for _, assignment := range assignments {
		user, err := s.userRepo.GetByID(assignment.ClinicianID)
		if err != nil {
			s.log.Warn("Failed to fetch user for assignment", zap.Error(err),
				zap.String("clinician_id", assignment.ClinicianID.String()))
			continue
		}

		responses = append(responses, dto.ClinicianAssignmentResponse{
			ClinicianID: assignment.ClinicianID,
			FullName:    user.FullName,
			Email:       user.Email,
			Role:        assignment.Role,
			AssignedAt:  assignment.AssignedAt,
			AssignedBy:  assignment.AssignedBy,
		})
	}

	return responses, nil
}

func (s *patientService) mapEntityToResponse(p *entity.Patient) *dto.PatientResponse {
	return &dto.PatientResponse{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		DateOfBirth:    p.DateOfBirth.Format("2006-01-02"),
		Email:          p.Email,
		Phone:          p.Phone,
		Address:        p.Address,
		Status:         p.Status,
		CreatedBy:      p.CreatedBy,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
