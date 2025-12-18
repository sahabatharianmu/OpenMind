package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
)

type ClinicalNoteService interface {
	Create(ctx context.Context, req dto.CreateClinicalNoteRequest, organizationID uuid.UUID) (*dto.ClinicalNoteResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateClinicalNoteRequest) (*dto.ClinicalNoteResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (*dto.ClinicalNoteResponse, error)
	List(ctx context.Context, organizationID uuid.UUID, page, pageSize int) ([]dto.ClinicalNoteResponse, int64, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type clinicalNoteService struct {
	repo repository.ClinicalNoteRepository
	log  logger.Logger
}

func NewClinicalNoteService(repo repository.ClinicalNoteRepository, log logger.Logger) ClinicalNoteService {
	return &clinicalNoteService{
		repo: repo,
		log:  log,
	}
}

func (s *clinicalNoteService) Create(ctx context.Context, req dto.CreateClinicalNoteRequest, organizationID uuid.UUID) (*dto.ClinicalNoteResponse, error) {
	var signedAt *time.Time
	if req.IsSigned {
		now := time.Now()
		signedAt = &now
	}

	note := &entity.ClinicalNote{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		PatientID:      req.PatientID,
		ClinicianID:    req.ClinicianID,
		AppointmentID:  req.AppointmentID,
		NoteType:       req.NoteType,
		Subjective:     req.Subjective,
		Objective:      req.Objective,
		Assessment:     req.Assessment,
		Plan:           req.Plan,
		IsSigned:       req.IsSigned,
		SignedAt:       signedAt,
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(note), nil
}

func (s *clinicalNoteService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateClinicalNoteRequest) (*dto.ClinicalNoteResponse, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.NoteType != "" {
		note.NoteType = req.NoteType
	}
	if req.Subjective != nil {
		note.Subjective = req.Subjective
	}
	if req.Objective != nil {
		note.Objective = req.Objective
	}
	if req.Assessment != nil {
		note.Assessment = req.Assessment
	}
	if req.Plan != nil {
		note.Plan = req.Plan
	}
	if req.IsSigned != nil {
		note.IsSigned = *req.IsSigned
		if *req.IsSigned && note.SignedAt == nil {
			now := time.Now()
			note.SignedAt = &now
		} else if !*req.IsSigned {
			note.SignedAt = nil
		}
	}

	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(note), nil
}

func (s *clinicalNoteService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *clinicalNoteService) Get(ctx context.Context, id uuid.UUID) (*dto.ClinicalNoteResponse, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.mapEntityToResponse(note), nil
}

func (s *clinicalNoteService) List(ctx context.Context, organizationID uuid.UUID, page, pageSize int) ([]dto.ClinicalNoteResponse, int64, error) {
	offset := (page - 1) * pageSize
	notes, total, err := s.repo.List(organizationID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.ClinicalNoteResponse
	for _, n := range notes {
		responses = append(responses, *s.mapEntityToResponse(&n))
	}

	return responses, total, nil
}

func (s *clinicalNoteService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetOrganizationID(userID)
}

func (s *clinicalNoteService) mapEntityToResponse(n *entity.ClinicalNote) *dto.ClinicalNoteResponse {
	return &dto.ClinicalNoteResponse{
		ID:             n.ID,
		OrganizationID: n.OrganizationID,
		PatientID:      n.PatientID,
		ClinicianID:    n.ClinicianID,
		AppointmentID:  n.AppointmentID,
		NoteType:       n.NoteType,
		Subjective:     n.Subjective,
		Objective:      n.Objective,
		Assessment:     n.Assessment,
		Plan:           n.Plan,
		IsSigned:       n.IsSigned,
		SignedAt:       n.SignedAt,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
	}
}
