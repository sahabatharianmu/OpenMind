package service

import (
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"github.com/sahabatharianmu/OpenMind/pkg/security"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(email, password, fullName, practiceName string) (*entity.User, error)
	Login(email, password string) (*dto.LoginResponse, error)
}

type authService struct {
	repo            repository.UserRepository
	jwt             *security.JWTService
	passwordService *crypto.PasswordService
	log             logger.Logger
}

func NewAuthService(
	repo repository.UserRepository,
	jwt *security.JWTService,
	passwordService *crypto.PasswordService,
	log logger.Logger,
) AuthService {
	return &authService{
		repo:            repo,
		jwt:             jwt,
		passwordService: passwordService,
		log:             log,
	}
}

func (s *authService) Register(email, password, fullName, practiceName string) (*entity.User, error) {
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
		Role:         "clinician", // Default role
	}

	organization := &entity.Organization{
		ID:   uuid.New(),
		Name: practiceName,
		Type: "clinic",
	}

	if err := s.repo.CreateWithOrganization(user, organization); err != nil {
		return nil, err
	}

	s.log.Info("User registered successfully with organization", zap.String("email", email), zap.String("practice", practiceName))
	return user, nil
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

	accessToken, refreshToken, err := s.jwt.GenerateTokens(user.ID, user.Email, user.Role)
	if err != nil {
		s.log.Error("Login failed: token generation error", zap.Error(err))
		return nil, response.ErrInternalServerError
	}

	s.log.Info("User logged in successfully", zap.String("email", email))
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
