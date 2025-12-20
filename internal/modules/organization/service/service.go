package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/organization/dto"
	orgRepo "github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	userRepo "github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type OrganizationService interface {
	GetMyOrganization(userID uuid.UUID) (*dto.OrganizationResponse, error)
	UpdateOrganization(userID uuid.UUID, req dto.UpdateOrganizationRequest) (*dto.OrganizationResponse, error)
	ListTeamMembers(userID uuid.UUID) ([]dto.TeamMemberResponse, error)
	UpdateMemberRole(userID, targetUserID uuid.UUID, role string) error
	RemoveMember(userID, targetUserID uuid.UUID) error
}

type organizationService struct {
	repo     orgRepo.OrganizationRepository
	userRepo userRepo.UserRepository
	log      logger.Logger
}

func NewOrganizationService(repo orgRepo.OrganizationRepository, userRepo userRepo.UserRepository, log logger.Logger) OrganizationService {
	return &organizationService{
		repo:     repo,
		userRepo: userRepo,
		log:      log,
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
	if req.Type != "" {
		org.Type = req.Type
	}
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

func (s *organizationService) ListTeamMembers(userID uuid.UUID) ([]dto.TeamMemberResponse, error) {
	org, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.Error("ListTeamMembers failed: org not found", zap.Error(err))
		return nil, response.ErrNotFound
	}

	members, err := s.repo.ListMembers(org.ID)
	if err != nil {
		s.log.Error("ListTeamMembers failed: failed to list members", zap.Error(err))
		return nil, err
	}

	result := make([]dto.TeamMemberResponse, 0, len(members))
	for _, member := range members {
		user, err := s.userRepo.GetByID(member.UserID)
		if err != nil {
			s.log.Warn("ListTeamMembers: failed to get user", zap.Error(err), zap.String("user_id", member.UserID.String()))
			continue
		}

		result = append(result, dto.TeamMemberResponse{
			UserID:   user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     member.Role,
			JoinedAt: member.CreatedAt,
		})
	}

	return result, nil
}

func (s *organizationService) UpdateMemberRole(userID, targetUserID uuid.UUID, role string) error {
	// Validate role
	if !constants.IsValidRole(role) {
		return response.NewBadRequest(fmt.Sprintf("Invalid role: %s", role))
	}

	org, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.Error("UpdateMemberRole failed: org not found", zap.Error(err))
		return response.ErrNotFound
	}

	// Check if target user is a member
	_, err = s.repo.GetMemberRole(org.ID, targetUserID)
	if err != nil {
		s.log.Error("UpdateMemberRole failed: target user is not a member", zap.Error(err))
		return response.ErrNotFound
	}

	// Prevent changing owner role (only owner can change their own role, but not remove it)
	currentRole, _ := s.repo.GetMemberRole(org.ID, targetUserID)
	if currentRole == constants.RoleOwner && role != constants.RoleOwner {
		return response.NewBadRequest("Cannot change owner role. Organization must have at least one owner.")
	}

	if err := s.repo.UpdateMemberRole(org.ID, targetUserID, role); err != nil {
		s.log.Error("UpdateMemberRole failed: update error", zap.Error(err))
		return err
	}

	s.log.Info("Member role updated successfully",
		zap.String("org_id", org.ID.String()),
		zap.String("target_user_id", targetUserID.String()),
		zap.String("new_role", role))

	return nil
}

func (s *organizationService) RemoveMember(userID, targetUserID uuid.UUID) error {
	org, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.Error("RemoveMember failed: org not found", zap.Error(err))
		return response.ErrNotFound
	}

	// Check if target user is a member
	role, err := s.repo.GetMemberRole(org.ID, targetUserID)
	if err != nil {
		s.log.Error("RemoveMember failed: target user is not a member", zap.Error(err))
		return response.ErrNotFound
	}

	// Prevent removing owner
	if role == constants.RoleOwner {
		return response.NewBadRequest("Cannot remove organization owner")
	}

	// Prevent removing yourself
	if userID == targetUserID {
		return response.NewBadRequest("Cannot remove yourself from the organization")
	}

	if err := s.repo.RemoveMember(org.ID, targetUserID); err != nil {
		s.log.Error("RemoveMember failed: remove error", zap.Error(err))
		return err
	}

	s.log.Info("Member removed successfully",
		zap.String("org_id", org.ID.String()),
		zap.String("target_user_id", targetUserID.String()))

	return nil
}
