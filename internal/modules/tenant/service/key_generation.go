package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GenerateKeysForExistingTenants generates encryption keys for all existing tenants
// that don't have keys yet. This should be run once after deploying the encryption feature.
func (s *tenantService) GenerateKeysForExistingTenants(ctx context.Context) error {
	if s.encryptionSvc == nil || s.keyRepo == nil {
		s.log.Warn("Encryption service or key repository not set, skipping key generation")
		return nil
	}

	// Get all tenants without encryption keys
	tenants, _, err := s.repo.List(1000, 0) // Get up to 1000 tenants
	if err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}

	generatedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, tenant := range tenants {
		// Check if key already exists
		existing, err := s.keyRepo.GetByTenantID(tenant.ID)
		if err == nil && existing != nil {
			skippedCount++
			continue
		}

		// Generate key for this tenant
		if err := s.generateTenantEncryptionKey(ctx, tenant.ID, tenant.OrganizationID); err != nil {
			s.log.Error("Failed to generate key for tenant",
				zap.Error(err),
				zap.String("tenant_id", tenant.ID.String()),
				zap.String("organization_id", tenant.OrganizationID.String()))
			errorCount++
			continue
		}

		generatedCount++
		s.log.Info("Generated encryption key for existing tenant",
			zap.String("tenant_id", tenant.ID.String()),
			zap.String("organization_id", tenant.OrganizationID.String()))
	}

	s.log.Info("Completed key generation for existing tenants",
		zap.Int("generated", generatedCount),
		zap.Int("skipped", skippedCount),
		zap.Int("errors", errorCount))

	if errorCount > 0 {
		return fmt.Errorf("failed to generate keys for %d tenants", errorCount)
	}

	return nil
}

// EnsureTenantHasKey ensures a tenant has an encryption key, generating one if needed
// This is called automatically when a tenant is created, but can also be called manually
func (s *tenantService) EnsureTenantHasKey(ctx context.Context, tenantID, organizationID uuid.UUID) error {
	if s.encryptionSvc == nil || s.keyRepo == nil {
		// If encryption service is not set, skip key generation (backward compatibility)
		return nil
	}

	return s.generateTenantEncryptionKey(ctx, tenantID, organizationID)
}

