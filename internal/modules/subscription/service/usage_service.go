package service

import (
	"context"

	"github.com/google/uuid"
	orgRepo "github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	patientRepo "github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

// UsageStats represents the current usage statistics for an organization
type UsageStats struct {
	PatientCount   int64 `json:"patient_count"`
	ClinicianCount int64 `json:"clinician_count"`
}

// UsageService provides methods to track usage statistics
type UsageService interface {
	GetPatientCount(ctx context.Context, organizationID uuid.UUID) (int64, error)
	GetClinicianCount(ctx context.Context, organizationID uuid.UUID) (int64, error)
	GetUsageStats(ctx context.Context, organizationID uuid.UUID) (*UsageStats, error)
}

type usageService struct {
	patientRepo patientRepo.PatientRepository
	orgRepo     orgRepo.OrganizationRepository
	log         logger.Logger
}

// NewUsageService creates a new usage tracking service
func NewUsageService(
	patientRepo patientRepo.PatientRepository,
	orgRepo orgRepo.OrganizationRepository,
	log logger.Logger,
) UsageService {
	return &usageService{
		patientRepo: patientRepo,
		orgRepo:     orgRepo,
		log:         log,
	}
}

// GetPatientCount counts the number of patients for an organization
// Note: This relies on the tenant schema being set in the database connection
func (s *usageService) GetPatientCount(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	// Use the patient repository's List method to count
	// We pass empty assignedPatientIDs to count all patients
	var assignedPatientIDs []uuid.UUID
	_, total, err := s.patientRepo.List(organizationID, 1, 0, assignedPatientIDs)
	if err != nil {
		s.log.Error("Failed to count patients", zap.Error(err),
			zap.String("organization_id", organizationID.String()))
		return 0, err
	}

	return total, nil
}

// GetClinicianCount counts the number of clinicians (organization members) for an organization
func (s *usageService) GetClinicianCount(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	count, err := s.orgRepo.GetMemberCount(organizationID)
	if err != nil {
		s.log.Error("Failed to count clinicians", zap.Error(err),
			zap.String("organization_id", organizationID.String()))
		return 0, err
	}

	return count, nil
}

// GetUsageStats returns both patient and clinician counts for an organization
func (s *usageService) GetUsageStats(ctx context.Context, organizationID uuid.UUID) (*UsageStats, error) {
	patientCount, err := s.GetPatientCount(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	clinicianCount, err := s.GetClinicianCount(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	return &UsageStats{
		PatientCount:   patientCount,
		ClinicianCount: clinicianCount,
	}, nil
}

