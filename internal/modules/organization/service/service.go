package service

import (
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/organization/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type OrganizationService interface {
	GetMyOrganization(userID uuid.UUID) (*dto.OrganizationResponse, error)
	UpdateOrganization(userID uuid.UUID, req dto.UpdateOrganizationRequest) (*dto.OrganizationResponse, error)
}

type organizationService struct {
	repo repository.OrganizationRepository
	log  logger.Logger
}

func NewOrganizationService(repo repository.OrganizationRepository, log logger.Logger) OrganizationService {
	return &organizationService{
		repo: repo,
		log:  log,
	}
}

func (s *organizationService) GetMyOrganization(userID uuid.UUID) (*dto.OrganizationResponse, error) {
	org, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.Error("GetMyOrganization failed", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, response.ErrNotFound
	}

	memberCount, err := s.repo.GetMemberCount(org.ID)
	if err != nil {
		s.log.Error("Failed to get member count", zap.Error(err))
		memberCount = 0 // Continue with 0 if error
	}

	return &dto.OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Type:        org.Type,
		TaxID:       org.TaxID,
		NPI:         org.NPI,
		Address:     org.Address,
		Currency:    org.Currency,
		Locale:      org.Locale,
		MemberCount: int(memberCount),
		CreatedAt:   org.CreatedAt,
	}, nil
}

func (s *organizationService) UpdateOrganization(
	userID uuid.UUID,
	req dto.UpdateOrganizationRequest,
) (*dto.OrganizationResponse, error) {
	org, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.Error("UpdateOrganization failed: org not found", zap.Error(err))
		return nil, response.ErrNotFound
	}

	org.Name = req.Name
	if req.TaxID != "" {
		org.TaxID = req.TaxID
	}
	if req.NPI != "" {
		org.NPI = req.NPI
	}
	if req.Address != "" {
		org.Address = req.Address
	}
	if req.Currency != "" {
		org.Currency = req.Currency
	}
	if req.Locale != "" {
		org.Locale = req.Locale
	}

	if err := s.repo.Update(org); err != nil {
		s.log.Error("UpdateOrganization failed: update error", zap.Error(err))
		return nil, err
	}

	memberCount, _ := s.repo.GetMemberCount(org.ID)

	s.log.Info("Organization updated successfully", zap.String("org_id", org.ID.String()))

	return &dto.OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Type:        org.Type,
		TaxID:       org.TaxID,
		NPI:         org.NPI,
		Address:     org.Address,
		Currency:    org.Currency,
		Locale:      org.Locale,
		MemberCount: int(memberCount),
		CreatedAt:   org.CreatedAt,
	}, nil
}
