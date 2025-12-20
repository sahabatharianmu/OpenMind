package service

import (
	"context"

	"github.com/google/uuid"
	orgRepo "github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	"github.com/sahabatharianmu/OpenMind/internal/modules/tenant/service"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/email"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"github.com/sahabatharianmu/OpenMind/pkg/security"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(email, password, fullName, practiceName, baseURL string) (*dto.RegisterResponse, error)
	Login(email, password string) (*dto.LoginResponse, error)
	ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error
}

type authService struct {
	repo            repository.UserRepository
	orgRepo         orgRepo.OrganizationRepository
	jwt             *security.JWTService
	passwordService *crypto.PasswordService
	tenantService   service.TenantService
	emailService    *email.EmailService
	log             logger.Logger
}

func NewAuthService(
	repo repository.UserRepository,
	orgRepo orgRepo.OrganizationRepository,
	jwt *security.JWTService,
	passwordService *crypto.PasswordService,
	tenantService service.TenantService,
	emailService *email.EmailService,
	log logger.Logger,
) AuthService {
	return &authService{
		repo:            repo,
		orgRepo:         orgRepo,
		jwt:             jwt,
		passwordService: passwordService,
		tenantService:   tenantService,
		emailService:    emailService,
		log:             log,
	}
}

func (s *authService) Register(email, password, fullName, practiceName, baseURL string) (*dto.RegisterResponse, error) {
	existingUser, _ := s.repo.FindByEmail(email)
	if existingUser != nil {
		s.log.Warn("Registration failed: email already registered", zap.String("email", email))
		return nil, response.ErrConflict
	}

	hashedPassword, err := s.passwordService.HashPassword(password)
	if err != nil {
		s.log.Error("Registration failed: password hashing error", zap.Error(err))
		return nil, err
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
		// Role is set in organization_members table, not here
	}

	organization := &entity.Organization{
		ID:               uuid.New(),
		Name:             practiceName,
		Type:             "clinic",
		SubscriptionTier: constants.TierFree, // Default to free tier
	}

	if err := s.repo.CreateWithOrganization(user, organization); err != nil {
		return nil, err
	}

	// Create tenant for the new organization
	ctx := context.Background()
	_, err = s.tenantService.CreateTenantForOrganization(ctx, organization.ID)
	if err != nil {
		s.log.Error("Failed to create tenant for organization during registration",
			zap.Error(err),
			zap.String("organization_id", organization.ID.String()))
		// Don't fail registration if tenant creation fails - it can be created later
		// But log the error for monitoring
	}

	// Get role from organization_members (should be "owner" for creator)
	role, err := s.orgRepo.GetMemberRole(organization.ID, user.ID)
	if err != nil {
		s.log.Error("Failed to get role after registration", zap.Error(err))
		// Default to "owner" if we can't get it (first user is owner)
		role = constants.RoleOwner
	}

	// Generate JWT tokens for auto-login
	accessToken, refreshToken, err := s.jwt.GenerateTokens(user.ID, user.Email, role)
	if err != nil {
		s.log.Error("Registration failed: token generation error", zap.Error(err))
		return nil, response.ErrInternalServerError
	}

	// Send welcome email (don't fail registration if email fails)
	dashboardURL := baseURL + "/dashboard"
	if err := s.emailService.SendWelcomeEmail(email, fullName, practiceName, dashboardURL); err != nil {
		s.log.Warn("Failed to send welcome email after registration",
			zap.Error(err),
			zap.String("email", email))
		// Don't fail registration if email fails
	}

	s.log.Info("User registered successfully with organization",
		zap.String("email", email),
		zap.String("practice", practiceName),
		zap.String("organization_id", organization.ID.String()),
		zap.String("role", role))

	return &dto.RegisterResponse{
		ID:           user.ID,
		Email:        user.Email,
		Role:         role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Login(email, password string) (*dto.LoginResponse, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		s.log.Warn("Login failed: user not found", zap.String("email", email))
		return nil, response.ErrUnauthorized
	}

	if verifyErr := s.passwordService.VerifyPassword(password, user.PasswordHash); verifyErr != nil {
		s.log.Warn("Login failed: invalid password", zap.String("email", email))
		return nil, response.ErrUnauthorized
	}

	// Get user's organization and role from organization_members
	org, err := s.orgRepo.GetByUserID(user.ID)
	if err != nil {
		s.log.Warn("Login failed: user has no organization", zap.String("email", email), zap.Error(err))
		return nil, response.ErrUnauthorized
	}

	// Get role from organization_members table
	role, err := s.orgRepo.GetMemberRole(org.ID, user.ID)
	if err != nil {
		s.log.Warn("Login failed: could not get user role", zap.String("email", email), zap.Error(err))
		return nil, response.ErrUnauthorized
	}

	accessToken, refreshToken, err := s.jwt.GenerateTokens(user.ID, user.Email, role)
	if err != nil {
		s.log.Error("Login failed: token generation error", zap.Error(err))
		return nil, response.ErrInternalServerError
	}

	s.log.Info("User logged in successfully", zap.String("email", email), zap.String("role", role))
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) SetupStatus() (*dto.SetupStatusResponse, error) {
	count, err := s.repo.CountUsers()
	if err != nil {
		s.log.Error("SetupStatus failed: error counting users", zap.Error(err))
		return nil, response.ErrInternalServerError
	}

	return &dto.SetupStatusResponse{
		IsSetupRequired: count == 0,
		HasUsers:        count > 0,
	}, nil
}

func (s *authService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("ChangePassword failed: user not found", zap.Error(err))
		return response.ErrNotFound
	}

	// Verify old password
	if verifyErr := s.passwordService.VerifyPassword(oldPassword, user.PasswordHash); verifyErr != nil {
		s.log.Warn("ChangePassword failed: invalid old password", zap.String("user_id", userID.String()))
		return response.ErrUnauthorized
	}

	// Hash new password
	hashedPassword, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		s.log.Error("ChangePassword failed: password hashing error", zap.Error(err))
		return err
	}

	// Update password
	user.PasswordHash = hashedPassword
	if err := s.repo.Update(user); err != nil {
		s.log.Error("ChangePassword failed: update error", zap.Error(err))
		return err
	}

	s.log.Info("Password changed successfully", zap.String("user_id", userID.String()))
	return nil
}
