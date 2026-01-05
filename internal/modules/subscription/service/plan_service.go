package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/subscription/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/subscription/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
)

type PlanService interface {
	CreatePlan(plan *entity.SubscriptionPlan) error
	UpdatePlan(plan *entity.SubscriptionPlan) error
	DeletePlan(id uuid.UUID) error
	GetPlan(id uuid.UUID) (*entity.SubscriptionPlan, error)
	ListAllPlans() ([]entity.SubscriptionPlan, error)
	ListActivePlans() ([]entity.SubscriptionPlan, error)
}

type planService struct {
	repo   repository.PlanRepository
	logger logger.Logger
}

func NewPlanService(repo repository.PlanRepository, logger logger.Logger) PlanService {
	return &planService{
		repo:   repo,
		logger: logger,
	}
}

func (s *planService) CreatePlan(plan *entity.SubscriptionPlan) error {
	if plan.Name == "" {
		return errors.New("plan name is required")
	}
	// Add more validation if needed
	return s.repo.Create(plan)
}

func (s *planService) UpdatePlan(plan *entity.SubscriptionPlan) error {
	// Verify exists
	existing, err := s.repo.GetByID(plan.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("plan not found")
	}
	return s.repo.Update(plan)
}

func (s *planService) DeletePlan(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *planService) GetPlan(id uuid.UUID) (*entity.SubscriptionPlan, error) {
	return s.repo.GetByID(id)
}

func (s *planService) ListAllPlans() ([]entity.SubscriptionPlan, error) {
	return s.repo.ListAll()
}

func (s *planService) ListActivePlans() ([]entity.SubscriptionPlan, error) {
	return s.repo.ListActive()
}
