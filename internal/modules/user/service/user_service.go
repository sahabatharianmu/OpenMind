package service

import (
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type UserService interface {
	GetProfile(userID uuid.UUID) (*dto.UserResponse, error)
	UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

type userService struct {
	repo repository.UserRepository
	log  logger.Logger
}

func NewUserService(repo repository.UserRepository, log logger.Logger) UserService {
	return &userService{
		repo: repo,
		log:  log,
	}
}

func (s *userService) GetProfile(userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("GetProfile failed", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, response.ErrNotFound
	}

	return &dto.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	}, nil
}

func (s *userService) UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("UpdateProfile failed: user not found", zap.Error(err))
		return nil, response.ErrNotFound
	}

	user.FullName = req.FullName

	if err := s.repo.Update(user); err != nil {
		s.log.Error("UpdateProfile failed: update error", zap.Error(err))
		return nil, err
	}

	s.log.Info("Profile updated successfully", zap.String("user_id", userID.String()))

	return &dto.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	}, nil
}
