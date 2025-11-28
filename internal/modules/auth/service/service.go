package service

import (
	"github.com/sahabatharianmu/OpenMind/internal/modules/auth/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/auth/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/auth/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/apperrors"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/security"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(email, password, role string) (*entity.User, error)
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

func (s *authService) Register(email, password, role string) (*entity.User, error) {
	existingUser, _ := s.repo.FindByEmail(email)
	if existingUser != nil {
		s.log.Warn("Registration failed: email already registered", zap.String("email", email))
		return nil, apperrors.NewConflict("email already registered")
	}

	hashedPassword, err := s.passwordService.HashPassword(password)
	if err != nil {
		s.log.Error("Registration failed: password hashing error", zap.Error(err))
		return nil, apperrors.NewInternalServerError(err.Error())
	}

	user := &entity.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, apperrors.NewInternalServerError(err.Error())
	}

	s.log.Info("User registered successfully", zap.String("email", email), zap.String("role", role))
	return user, nil
}

func (s *authService) Login(email, password string) (*dto.LoginResponse, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		s.log.Warn("Login failed: user not found", zap.String("email", email))
		return nil, apperrors.NewUnauthorized("invalid credentials")
	}

	if err := s.passwordService.VerifyPassword(password, user.PasswordHash); err != nil {
		s.log.Warn("Login failed: invalid password", zap.String("email", email))
		return nil, apperrors.NewUnauthorized("invalid credentials")
	}

	accessToken, refreshToken, err := s.jwt.GenerateTokens(user.ID, user.Email, user.Role)
	if err != nil {
		s.log.Error("Login failed: token generation error", zap.Error(err))
		return nil, apperrors.NewInternalServerError(err.Error())
	}

	s.log.Info("User logged in successfully", zap.String("email", email))
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
