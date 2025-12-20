package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

// UpgradePrompt contains information about the upgrade prompt
type UpgradePrompt struct {
	Feature    string `json:"feature"`     // "patients" or "clinicians"
	Current    int64  `json:"current"`     // Current usage
	Limit      int64  `json:"limit"`       // Limit for current tier
	UpgradeURL string `json:"upgrade_url"` // URL to upgrade page
}

// LimitReachedError is a custom error that includes upgrade prompt information
type LimitReachedError struct {
	*response.AppError
	UpgradePrompt *UpgradePrompt `json:"upgrade_prompt,omitempty"`
}

func (e *LimitReachedError) Error() string {
	return e.Message
}

// GetUpgradePrompt returns the upgrade prompt for error handling
func (e *LimitReachedError) GetUpgradePrompt() interface{} {
	return e.UpgradePrompt
}

// GetAppError returns the underlying AppError
func (e *LimitReachedError) GetAppError() *response.AppError {
	return e.AppError
}

// FeatureGatingService provides methods to check feature limits
type FeatureGatingService interface {
	CheckPatientLimit(ctx context.Context, organizationID uuid.UUID) error
	CheckClinicianLimit(ctx context.Context, organizationID uuid.UUID) error
	GetTierLimits(tier string) constants.TierLimits
}

type featureGatingService struct {
	orgRepo      repository.OrganizationRepository
	usageService UsageService
	log          logger.Logger
	baseURL      string
}

// NewFeatureGatingService creates a new feature gating service
func NewFeatureGatingService(
	orgRepo repository.OrganizationRepository,
	usageService UsageService,
	log logger.Logger,
	baseURL string,
) FeatureGatingService {
	return &featureGatingService{
		orgRepo:      orgRepo,
		usageService: usageService,
		log:          log,
		baseURL:      baseURL,
	}
}

// GetTierLimits returns the limits for a given subscription tier
func (s *featureGatingService) GetTierLimits(tier string) constants.TierLimits {
	return constants.GetTierLimits(tier)
}

// CheckPatientLimit verifies that the organization hasn't reached its patient limit
func (s *featureGatingService) CheckPatientLimit(ctx context.Context, organizationID uuid.UUID) error {
	// Get organization to check subscription tier
	org, err := s.orgRepo.GetByID(organizationID)
	if err != nil {
		s.log.Error("Failed to get organization for limit check", zap.Error(err),
			zap.String("organization_id", organizationID.String()))
		return response.NewInternalServerError("Failed to check patient limit")
	}

	// Get tier limits
	limits := s.GetTierLimits(org.SubscriptionTier)

	// If unlimited, no check needed
	if constants.IsUnlimited(limits.MaxPatients) {
		return nil
	}

	// Get current patient count
	patientCount, err := s.usageService.GetPatientCount(ctx, organizationID)
	if err != nil {
		s.log.Error("Failed to get patient count for limit check", zap.Error(err),
			zap.String("organization_id", organizationID.String()))
		return response.NewInternalServerError("Failed to check patient limit")
	}

	// Check if limit is reached
	if patientCount >= int64(limits.MaxPatients) {
		return &LimitReachedError{
			AppError: response.NewBadRequest(
				fmt.Sprintf("Patient limit reached. You have %d of %d patients. Upgrade to add more patients.", patientCount, limits.MaxPatients),
			),
			UpgradePrompt: &UpgradePrompt{
				Feature:    "patients",
				Current:    patientCount,
				Limit:      int64(limits.MaxPatients),
				UpgradeURL: s.baseURL + "/pricing",
			},
		}
	}

	return nil
}

// CheckClinicianLimit verifies that the organization hasn't reached its clinician limit
func (s *featureGatingService) CheckClinicianLimit(ctx context.Context, organizationID uuid.UUID) error {
	// Get organization to check subscription tier
	org, err := s.orgRepo.GetByID(organizationID)
	if err != nil {
		s.log.Error("Failed to get organization for limit check", zap.Error(err),
			zap.String("organization_id", organizationID.String()))
		return response.NewInternalServerError("Failed to check clinician limit")
	}

	// Get tier limits
	limits := s.GetTierLimits(org.SubscriptionTier)

	// If unlimited, no check needed
	if constants.IsUnlimited(limits.MaxClinicians) {
		return nil
	}

	// Get current clinician count
	clinicianCount, err := s.usageService.GetClinicianCount(ctx, organizationID)
	if err != nil {
		s.log.Error("Failed to get clinician count for limit check", zap.Error(err),
			zap.String("organization_id", organizationID.String()))
		return response.NewInternalServerError("Failed to check clinician limit")
	}

	// Check if limit is reached
	if clinicianCount >= int64(limits.MaxClinicians) {
		return &LimitReachedError{
			AppError: response.NewBadRequest(
				fmt.Sprintf("Clinician limit reached. You have %d of %d clinicians. Upgrade to add more team members.", clinicianCount, limits.MaxClinicians),
			),
			UpgradePrompt: &UpgradePrompt{
				Feature:    "clinicians",
				Current:    clinicianCount,
				Limit:      int64(limits.MaxClinicians),
				UpgradeURL: s.baseURL + "/pricing",
			},
		}
	}

	return nil
}

