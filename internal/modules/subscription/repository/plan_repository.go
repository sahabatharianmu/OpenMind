package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/subscription/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PlanRepository interface {
	Create(plan *entity.SubscriptionPlan) error
	Update(plan *entity.SubscriptionPlan) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*entity.SubscriptionPlan, error)
	ListAll() ([]entity.SubscriptionPlan, error)
	ListActive() ([]entity.SubscriptionPlan, error)
}

type planRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewPlanRepository(db *gorm.DB, logger logger.Logger) PlanRepository {
	return &planRepository{
		db:     db,
		logger: logger,
	}
}

func (r *planRepository) Create(plan *entity.SubscriptionPlan) error {
	if err := r.db.Create(plan).Error; err != nil {
		r.logger.Error("Failed to create subscription plan", zap.Error(err))
		return err
	}
	return nil
}

func (r *planRepository) Update(plan *entity.SubscriptionPlan) error {
	if err := r.db.Save(plan).Error; err != nil {
		r.logger.Error("Failed to update subscription plan", zap.String("id", plan.ID.String()), zap.Error(err))
		return err
	}
	return nil
}

func (r *planRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entity.SubscriptionPlan{}, id).Error; err != nil {
		r.logger.Error("Failed to delete subscription plan", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}

func (r *planRepository) GetByID(id uuid.UUID) (*entity.SubscriptionPlan, error) {
	var plan entity.SubscriptionPlan
	if err := r.db.First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to get subscription plan", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) ListAll() ([]entity.SubscriptionPlan, error) {
	var plans []entity.SubscriptionPlan
	if err := r.db.Find(&plans).Error; err != nil {
		r.logger.Error("Failed to list all subscription plans", zap.Error(err))
		return nil, err
	}
	return plans, nil
}

func (r *planRepository) ListActive() ([]entity.SubscriptionPlan, error) {
	var plans []entity.SubscriptionPlan
	if err := r.db.Where("is_active = ?", true).Find(&plans).Error; err != nil {
		r.logger.Error("Failed to list active subscription plans", zap.Error(err))
		return nil, err
	}
	return plans, nil
}
