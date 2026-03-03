package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

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
	ForgotPassword(email, baseURL string) error
	ResetPassword(token, newPassword string) error
	RefreshToken(refreshToken string) (*dto.LoginResponse, error)
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
	accessToken, refreshToken, err := s.jwt.GenerateTokens(user.ID, user.Email, role, "user")
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

	var role string

	// Platform admins may not have an organization — allow them through
	if user.SystemRole == "admin" {
		org, err := s.orgRepo.GetByUserID(user.ID)
		if err != nil {
			// Admin without org — use "admin" as their role
			role = "admin"
		} else {
			role, _ = s.orgRepo.GetMemberRole(org.ID, user.ID)
			if role == "" {
				role = "admin"
			}
		}
	} else {
		// Regular users must have an organization
		org, err := s.orgRepo.GetByUserID(user.ID)
		if err != nil {
			s.log.Warn("Login failed: user has no organization", zap.String("email", email), zap.Error(err))
			return nil, response.ErrUnauthorized
		}

		role, err = s.orgRepo.GetMemberRole(org.ID, user.ID)
		if err != nil {
			s.log.Warn("Login failed: could not get user role", zap.String("email", email), zap.Error(err))
			return nil, response.ErrUnauthorized
		}
	}

	accessToken, refreshToken, err := s.jwt.GenerateTokens(user.ID, user.Email, role, user.SystemRole)
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

// ForgotPassword generates a password reset token, saves it to the database,
// and sends a password reset email to the user.
func (s *authService) ForgotPassword(emailAddr, baseURL string) error {
	user, err := s.repo.FindByEmail(emailAddr)
	if err != nil {
		// Don't reveal whether email exists — always return success to prevent enumeration
		s.log.Info("ForgotPassword: email not found (returning success to prevent enumeration)",
			zap.String("email", emailAddr))
		return nil
	}

	// Generate secure random token (32 bytes → 64 hex characters)
	tokenBytes := make([]byte, 32) //nolint:mnd
	if _, err := rand.Read(tokenBytes); err != nil {
		s.log.Error("ForgotPassword: failed to generate reset token", zap.Error(err))
		return response.ErrInternalServerError
	}
	token := hex.EncodeToString(tokenBytes)

	// Token expires in 1 hour
	expiresAt := time.Now().Add(1 * time.Hour)
	user.PasswordResetToken = &token
	user.PasswordResetExpiresAt = &expiresAt

	if err := s.repo.Update(user); err != nil {
		s.log.Error("ForgotPassword: failed to save reset token", zap.Error(err))
		return response.ErrInternalServerError
	}

	// Send password reset email
	resetURL := baseURL + "/auth?mode=reset&token=" + token
	if err := s.emailService.SendPasswordResetEmail(emailAddr, user.FullName, resetURL); err != nil {
		s.log.Error("ForgotPassword: failed to send reset email",
			zap.Error(err), zap.String("email", emailAddr))
		// Still return nil to prevent enumeration
	}

	s.log.Info("Password reset token generated", zap.String("email", emailAddr))
	return nil
}

// ResetPassword validates the reset token, updates the password, and clears
// the reset token from the database.
func (s *authService) ResetPassword(token, newPassword string) error {
	user, err := s.repo.FindByResetToken(token)
	if err != nil {
		s.log.Warn("ResetPassword: invalid or expired token")
		return response.ErrInvalidInput
	}

	// Hash new password
	hashedPassword, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		s.log.Error("ResetPassword: password hashing error", zap.Error(err))
		return response.ErrInternalServerError
	}

	// Update password and clear reset token
	user.PasswordHash = hashedPassword
	user.PasswordResetToken = nil
	user.PasswordResetExpiresAt = nil

	if err := s.repo.Update(user); err != nil {
		s.log.Error("ResetPassword: failed to update password", zap.Error(err))
		return response.ErrInternalServerError
	}

	s.log.Info("Password reset successfully", zap.String("user_id", user.ID.String()))
	return nil
}

// RefreshToken generates a new access token (and refresh token) from a valid refresh token.
func (s *authService) RefreshToken(refreshToken string) (*dto.LoginResponse, error) {
	// Validate refresh token and extract user ID
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.log.Warn("RefreshToken: invalid refresh token", zap.Error(err))
		return nil, response.ErrUnauthorized
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		s.log.Error("RefreshToken: invalid subject in token", zap.Error(err))
		return nil, response.ErrUnauthorized
	}

	// Look up current user details from DB
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Warn("RefreshToken: user not found", zap.String("user_id", userID.String()))
		return nil, response.ErrUnauthorized
	}

	// Get the user's current org role
	org, err := s.orgRepo.GetByUserID(user.ID)
	if err != nil {
		s.log.Warn("RefreshToken: user has no organization", zap.String("user_id", userID.String()))
		return nil, response.ErrUnauthorized
	}

	role, err := s.orgRepo.GetMemberRole(org.ID, user.ID)
	if err != nil {
		s.log.Warn("RefreshToken: could not get user role", zap.String("user_id", userID.String()))
		return nil, response.ErrUnauthorized
	}

	// Generate new token pair
	newAccessToken, newRefreshToken, err := s.jwt.GenerateTokens(user.ID, user.Email, role, user.SystemRole)
	if err != nil {
		s.log.Error("RefreshToken: token generation error", zap.Error(err))
		return nil, response.ErrInternalServerError
	}

	return &dto.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
