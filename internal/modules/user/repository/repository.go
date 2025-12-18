package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entity.User) error
	CreateWithOrganization(user *entity.User, organization *entity.Organization) error
	FindByEmail(email string) (*entity.User, error)
	GetByID(id uuid.UUID) (*entity.User, error)
	Update(user *entity.User) error
	CountUsers() (int64, error)
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

func (r *userRepository) CreateWithOrganization(user *entity.User, organization *entity.Organization) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			r.log.Error("Failed to create user", zap.Error(err), zap.String("email", user.Email))
			return err
		}

		// Organization.CreatedBy field does not exist yet.

		if err := tx.Create(organization).Error; err != nil {
			r.log.Error("Failed to create organization", zap.Error(err), zap.String("name", organization.Name))
			return err
		}

		member := entity.OrganizationMember{
			OrganizationID: organization.ID,
			UserID:         user.ID,
			Role:           "owner", // Creator is owner
		}

		if err := tx.Create(&member).Error; err != nil {
			r.log.Error("Failed to add user to organization", zap.Error(err))
			return err
		}

		return nil
	})
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

func (r *userRepository) GetByID(id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find user by ID", zap.Error(err), zap.String("id", id.String()))
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *entity.User) error {
	if err := r.db.Save(user).Error; err != nil {
		r.log.Error("Failed to update user", zap.Error(err), zap.String("id", user.ID.String()))
		return err
	}
	return nil
}

func (r *userRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&entity.User{}).Count(&count).Error
	if err != nil {
		r.log.Error("Failed to count users", zap.Error(err))
		return 0, err
	}
	return count, nil
}
