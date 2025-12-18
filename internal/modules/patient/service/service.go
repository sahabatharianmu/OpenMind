package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
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
	Get(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*dto.PatientResponse, error)
	List(ctx context.Context, organizationID uuid.UUID, page, pageSize int) ([]dto.PatientResponse, int64, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type patientService struct {
	repo repository.PatientRepository
	log  logger.Logger
}

func NewPatientService(repo repository.PatientRepository, log logger.Logger) PatientService {
	return &patientService{
		repo: repo,
		log:  log,
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
) (*dto.PatientResponse, error) {
	patient, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if patient.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	return s.mapEntityToResponse(patient), nil
}

func (s *patientService) List(
	ctx context.Context,
	organizationID uuid.UUID,
	page, pageSize int,
) ([]dto.PatientResponse, int64, error) {
	offset := (page - 1) * pageSize
	patients, total, err := s.repo.List(organizationID, pageSize, offset)
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
