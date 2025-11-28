package repository

import (
	"errors"

	"github.com/sahabatharianmu/OpenMind/internal/modules/auth/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entity.User) error
	FindByEmail(email string) (*entity.User, error)
}

type userRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewUserRepository(db *gorm.DB, log logger.Logger) UserRepository {
	return &userRepository{
		db:  db,
		log: log,
	}
}

func (r *userRepository) Create(user *entity.User) error {
	if err := r.db.Create(user).Error; err != nil {
		r.log.Error("Failed to create user", zap.Error(err), zap.String("email", user.Email))
		return err
	}
	return nil
}

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find user by email", zap.Error(err), zap.String("email", email))
		}
		return nil, err
	}
	return &user, nil
}
