package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/organization/entity"
	orgRepo "github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	subscriptionService "github.com/sahabatharianmu/OpenMind/internal/modules/subscription/service"
	teamEntity "github.com/sahabatharianmu/OpenMind/internal/modules/team/entity"
	teamRepo "github.com/sahabatharianmu/OpenMind/internal/modules/team/repository"
	userEntity "github.com/sahabatharianmu/OpenMind/internal/modules/user/entity"
	userRepo "github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/email"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type TeamInvitationService interface {
	SendInvitation(ctx context.Context, organizationID, invitedBy uuid.UUID, email, role string) (*teamEntity.TeamInvitation, error)
	AcceptInvitation(ctx context.Context, token string, userID *uuid.UUID) error
	RegisterAndAcceptInvitation(ctx context.Context, token string, email, password, fullName string) (*uuid.UUID, error)
	GetInvitationByToken(ctx context.Context, token string) (*teamEntity.TeamInvitation, error)
	ListInvitations(ctx context.Context, organizationID uuid.UUID, page, pageSize int) ([]teamEntity.TeamInvitation, int64, error)
	CancelInvitation(ctx context.Context, invitationID, organizationID uuid.UUID) error
	ResendInvitation(ctx context.Context, invitationID, organizationID uuid.UUID) error
}

type teamInvitationService struct {
	invitationRepo  teamRepo.TeamInvitationRepository
	orgRepo         orgRepo.OrganizationRepository
	userRepo        userRepo.UserRepository
	passwordService *crypto.PasswordService
	emailService    *email.EmailService
	gatingService   subscriptionService.FeatureGatingService
	log             logger.Logger
	baseURL         string // Base URL for invitation links
}

func NewTeamInvitationService(
	invitationRepo teamRepo.TeamInvitationRepository,
	orgRepo orgRepo.OrganizationRepository,
	userRepo userRepo.UserRepository,
	passwordService *crypto.PasswordService,
	emailService *email.EmailService,
	gatingService subscriptionService.FeatureGatingService,
	log logger.Logger,
	baseURL string,
) TeamInvitationService {
	return &teamInvitationService{
		invitationRepo:  invitationRepo,
		orgRepo:         orgRepo,
		userRepo:        userRepo,
		passwordService: passwordService,
		emailService:    emailService,
		gatingService:   gatingService,
		log:             log,
		baseURL:         baseURL,
	}
}

// generateInvitationToken generates a secure random token for invitation
func (s *teamInvitationService) generateInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SendInvitation sends a team invitation via email
func (s *teamInvitationService) SendInvitation(
	ctx context.Context,
	organizationID, invitedBy uuid.UUID,
	email, role string,
) (*teamEntity.TeamInvitation, error) {
	// Check clinician limit before sending invitation
	if s.gatingService != nil {
		if err := s.gatingService.CheckClinicianLimit(ctx, organizationID); err != nil {
			return nil, err
		}
	}

	// Validate role
	if !constants.IsValidRole(role) {
		return nil, response.NewBadRequest(fmt.Sprintf("Invalid role: %s", role))
	}

	// Check if user already exists - existing users cannot be invited (they already have their own organization/tenant)
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil {
		// Check if user is already a member of this organization
		org, err := s.orgRepo.GetByUserID(existingUser.ID)
		if err == nil && org != nil {
			if org.ID == organizationID {
				return nil, response.NewConflict("User is already a member of this organization")
			}
			// User exists in a different organization - prevent cross-organization invitations
			return nil, response.NewBadRequest("This email is already registered with another organization. Users cannot be members of multiple organizations.")
		}
		// User exists but not in any organization (edge case) - still reject
		// because they should have their own organization when they register
		return nil, response.NewBadRequest("This email is already registered. Existing users cannot accept invitations as they already belong to their own organization.")
	}

	// Cancel any pending invitations for this email and organization
	_ = s.invitationRepo.CancelPendingInvitations(email, organizationID)

	// Generate secure token
	token, err := s.generateInvitationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	// Create invitation
	invitation := &teamEntity.TeamInvitation{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Email:          email,
		Role:           role,
		Token:          token,
		InvitedBy:      invitedBy,
		Status:         "pending",
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days expiration
	}

	if err := s.invitationRepo.Create(invitation); err != nil {
		s.log.Error("Failed to create invitation", zap.Error(err),
			zap.String("email", email),
			zap.String("organization_id", organizationID.String()))
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Get organization and inviter details for email
	org, err := s.orgRepo.GetByID(organizationID)
	if err != nil {
		s.log.Error("Failed to get organization", zap.Error(err))
		return invitation, nil // Return invitation even if email fails
	}

	inviter, err := s.userRepo.GetByID(invitedBy)
	if err != nil {
		s.log.Error("Failed to get inviter", zap.Error(err))
		return invitation, nil // Return invitation even if email fails
	}

	// Send invitation email
	if err := s.emailService.SendInvitationEmail(
		email,
		inviter.FullName,
		org.Name,
		token,
		s.baseURL,
	); err != nil {
		s.log.Error("Failed to send invitation email", zap.Error(err),
			zap.String("email", email))
		// Don't fail the operation if email fails, but log it
		// The invitation is still created and can be resent
	}

	s.log.Info("Team invitation sent", zap.String("email", email),
		zap.String("organization_id", organizationID.String()),
		zap.String("role", role))

	return invitation, nil
}

// AcceptInvitation is deprecated - existing users cannot accept invitations
// They already belong to their own organization/tenant
// This method is kept for backward compatibility but will always reject
func (s *teamInvitationService) AcceptInvitation(ctx context.Context, token string, userID *uuid.UUID) error {
	if userID == nil {
		return response.NewBadRequest("User ID is required")
	}

	// Get invitation by token
	invitation, err := s.invitationRepo.GetByToken(token)
	if err != nil {
		return response.ErrNotFound
	}

	// Get user
	user, err := s.userRepo.GetByID(*userID)
	if err != nil {
		return response.ErrNotFound
	}

	// Verify email matches
	if user.Email != invitation.Email {
		return response.NewBadRequest("Invitation email does not match your account email")
	}

	// Reject - existing users cannot accept invitations
	// They already have their own organization/tenant
	return response.NewBadRequest("Existing users cannot accept invitations. You already belong to your own organization. Only new users can accept invitations to join an organization.")
}

// RegisterAndAcceptInvitation creates a new user account and accepts the invitation
func (s *teamInvitationService) RegisterAndAcceptInvitation(ctx context.Context, token string, email, password, fullName string) (*uuid.UUID, error) {
	// Get invitation by token
	invitation, err := s.invitationRepo.GetByToken(token)
	if err != nil {
		return nil, response.ErrNotFound
	}

	// Check if invitation can be accepted
	if !invitation.CanBeAccepted() {
		if invitation.IsExpired() {
			invitation.Status = "expired"
			_ = s.invitationRepo.Update(invitation)
			return nil, response.NewBadRequest("Invitation has expired")
		}
		return nil, response.NewBadRequest("Invitation is no longer valid")
	}

	// Verify email matches invitation
	if email != invitation.Email {
		return nil, response.NewBadRequest("Email does not match the invitation")
	}

	// Check if user already exists - reject if they do (they already have their own organization/tenant)
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil {
		return nil, response.NewBadRequest("This email is already registered. Existing users cannot accept invitations as they already belong to their own organization.")
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(password)
	if err != nil {
		s.log.Error("Failed to hash password during invitation signup", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &userEntity.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.log.Error("Failed to create user during invitation signup", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Add user to organization
	member := &entity.OrganizationMember{
		OrganizationID: invitation.OrganizationID,
		UserID:         user.ID,
		Role:           invitation.Role,
		CreatedAt:      time.Now(),
	}

	if err := s.orgRepo.AddMember(member); err != nil {
		s.log.Error("Failed to add user to organization", zap.Error(err),
			zap.String("organization_id", invitation.OrganizationID.String()),
			zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Update invitation status
	invitation.Status = "accepted"
	now := time.Now()
	invitation.AcceptedAt = &now
	invitation.AcceptedBy = &user.ID

	if err := s.invitationRepo.Update(invitation); err != nil {
		s.log.Error("Failed to update invitation status", zap.Error(err))
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	s.log.Info("User registered and invitation accepted", zap.String("email", email),
		zap.String("organization_id", invitation.OrganizationID.String()),
		zap.String("user_id", user.ID.String()))

	return &user.ID, nil
}

// GetInvitationByToken retrieves an invitation by token
func (s *teamInvitationService) GetInvitationByToken(ctx context.Context, token string) (*teamEntity.TeamInvitation, error) {
	return s.invitationRepo.GetByToken(token)
}

// ListInvitations lists all invitations for an organization
func (s *teamInvitationService) ListInvitations(
	ctx context.Context,
	organizationID uuid.UUID,
	page, pageSize int,
) ([]teamEntity.TeamInvitation, int64, error) {
	offset := (page - 1) * pageSize
	return s.invitationRepo.ListByOrganization(organizationID, pageSize, offset)
}

// CancelInvitation cancels an invitation
func (s *teamInvitationService) CancelInvitation(ctx context.Context, invitationID, organizationID uuid.UUID) error {
	invitation, err := s.invitationRepo.GetByID(invitationID)
	if err != nil {
		return response.ErrNotFound
	}

	if invitation.OrganizationID != organizationID {
		return response.ErrForbidden
	}

	if invitation.Status != "pending" {
		return response.NewBadRequest("Only pending invitations can be cancelled")
	}

	invitation.Status = "cancelled"
	if err := s.invitationRepo.Update(invitation); err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	return nil
}

// ResendInvitation resends an invitation email
func (s *teamInvitationService) ResendInvitation(ctx context.Context, invitationID, organizationID uuid.UUID) error {
	invitation, err := s.invitationRepo.GetByID(invitationID)
	if err != nil {
		return response.ErrNotFound
	}

	if invitation.OrganizationID != organizationID {
		return response.ErrForbidden
	}

	if invitation.Status != "pending" {
		return response.NewBadRequest("Only pending invitations can be resent")
	}

	// Check if expired, if so, extend expiration
	if invitation.IsExpired() {
		invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	// Get organization and inviter details
	org, err := s.orgRepo.GetByID(organizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	inviter, err := s.userRepo.GetByID(invitation.InvitedBy)
	if err != nil {
		return fmt.Errorf("failed to get inviter: %w", err)
	}

	// Resend email
	if err := s.emailService.SendInvitationEmail(
		invitation.Email,
		inviter.FullName,
		org.Name,
		invitation.Token,
		s.baseURL,
	); err != nil {
		s.log.Error("Failed to resend invitation email", zap.Error(err))
		return fmt.Errorf("failed to resend email: %w", err)
	}

	// Update invitation
	if err := s.invitationRepo.Update(invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	return nil
}

