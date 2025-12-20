package service

import (
	"context"
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/internal/modules/payment/dto"
	paymentEntity "github.com/sahabatharianmu/OpenMind/internal/modules/payment/entity"
	paymentRepo "github.com/sahabatharianmu/OpenMind/internal/modules/payment/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/payment"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

// PaymentMethodService defines the interface for payment method operations
type PaymentMethodService interface {
	CreatePaymentMethod(ctx context.Context, organizationID uuid.UUID, req dto.CreatePaymentMethodRequest) (*dto.PaymentMethodResponse, error)
	ListPaymentMethods(ctx context.Context, organizationID uuid.UUID) (*dto.ListPaymentMethodsResponse, error)
	GetPaymentMethod(ctx context.Context, organizationID, paymentMethodID uuid.UUID) (*dto.PaymentMethodResponse, error)
	DeletePaymentMethod(ctx context.Context, organizationID, paymentMethodID uuid.UUID) error
	SetDefaultPaymentMethod(ctx context.Context, organizationID, paymentMethodID uuid.UUID) error
}

// paymentMethodService implements PaymentMethodService
type paymentMethodService struct {
	repo            paymentRepo.PaymentMethodRepository
	providerManager *payment.PaymentProviderManager
	encryptSvc      *crypto.EncryptionService
	config          *config.Config
	log             logger.Logger
}

// NewPaymentMethodService creates a new payment method service
func NewPaymentMethodService(
	repo paymentRepo.PaymentMethodRepository,
	providerManager *payment.PaymentProviderManager,
	encryptSvc *crypto.EncryptionService,
	cfg *config.Config,
	log logger.Logger,
) PaymentMethodService {
	return &paymentMethodService{
		repo:            repo,
		providerManager: providerManager,
		encryptSvc:      encryptSvc,
		config:          cfg,
		log:             log,
	}
}

// CreatePaymentMethod creates a new payment method
func (s *paymentMethodService) CreatePaymentMethod(
	ctx context.Context,
	organizationID uuid.UUID,
	req dto.CreatePaymentMethodRequest,
) (*dto.PaymentMethodResponse, error) {
	// Validate token
	if req.Token == "" {
		return nil, response.NewBadRequest("payment method token is required")
	}

	// Determine which provider to use
	// Use provider from request, or default to configured default provider
	providerType := payment.ProviderType(req.Provider)
	if providerType == "" {
		// Use default provider type from manager
		providerType = s.providerManager.GetDefaultProviderType()
	}

	// Get the appropriate provider
	provider, err := s.providerManager.GetProvider(providerType)
	if err != nil {
		s.log.Error("Failed to get payment provider", zap.Error(err), zap.String("provider_type", string(providerType)))
		return nil, response.NewBadRequest("payment provider not available")
	}

	// Create payment method in provider (e.g., Stripe, Square)
	// This verifies the token and retrieves payment method details
	providerPaymentMethodID, err := provider.CreatePaymentMethod(ctx, req.Token)
	if err != nil {
		s.log.Error("Failed to create payment method in provider", zap.Error(err), zap.String("token", req.Token), zap.String("provider", string(providerType)))
		return nil, response.NewBadRequest("invalid payment method token")
	}

	// Get payment method details from provider
	pmInfo, err := provider.GetPaymentMethod(ctx, providerPaymentMethodID)
	if err != nil {
		s.log.Error("Failed to get payment method details from provider", zap.Error(err), zap.String("provider_payment_method_id", providerPaymentMethodID))
		return nil, response.NewInternalServerError("failed to retrieve payment method details")
	}

	// Encrypt the payment method token for storage
	// Use tenant-specific encryption key (HIPAA compliant)
	encryptedTokenStr, err := s.encryptSvc.Encrypt(req.Token, organizationID)
	if err != nil {
		s.log.Error("Failed to encrypt payment method token", zap.Error(err))
		return nil, response.NewInternalServerError("failed to encrypt payment method")
	}

	// Convert encrypted string to bytes for storage
	encryptedTokenBytes, err := base64.StdEncoding.DecodeString(encryptedTokenStr)
	if err != nil {
		s.log.Error("Failed to decode encrypted token", zap.Error(err))
		return nil, response.NewInternalServerError("failed to process payment method")
	}

	// Determine if this should be the default payment method
	// If this is the first payment method for the organization, make it default
	count, err := s.repo.CountByOrganizationID(organizationID)
	if err != nil {
		s.log.Error("Failed to count payment methods", zap.Error(err))
		return nil, response.NewInternalServerError("failed to check existing payment methods")
	}
	isDefault := count == 0

	// If this is being set as default, unset existing defaults
	if isDefault {
		existingDefault, err := s.repo.GetDefaultByOrganizationID(organizationID)
		if err == nil && existingDefault != nil {
			// Unset existing default
			existingDefault.IsDefault = false
			if err := s.repo.Update(existingDefault); err != nil {
				s.log.Warn("Failed to unset existing default payment method", zap.Error(err))
			}
		}
	}

	// Create payment method entity
	paymentMethod := &paymentEntity.PaymentMethod{
		OrganizationID:          organizationID,
		Provider:                string(providerType), // Use the provider type from request or default
		EncryptedToken:          encryptedTokenBytes,
		ProviderPaymentMethodID: providerPaymentMethodID,
		Last4:                   pmInfo.Last4,
		Brand:                   pmInfo.Brand,
		ExpiryMonth:             pmInfo.ExpiryMonth,
		ExpiryYear:              pmInfo.ExpiryYear,
		IsDefault:               isDefault,
	}

	// Save to database
	if err := s.repo.Create(paymentMethod); err != nil {
		s.log.Error("Failed to create payment method in database", zap.Error(err))
		return nil, response.NewInternalServerError("failed to save payment method")
	}

	// Return response
	return s.mapEntityToResponse(paymentMethod), nil
}

// ListPaymentMethods lists all payment methods for an organization
func (s *paymentMethodService) ListPaymentMethods(
	ctx context.Context,
	organizationID uuid.UUID,
) (*dto.ListPaymentMethodsResponse, error) {
	paymentMethods, err := s.repo.GetByOrganizationID(organizationID)
	if err != nil {
		s.log.Error("Failed to list payment methods", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, response.NewInternalServerError("failed to list payment methods")
	}

	responses := make([]dto.PaymentMethodResponse, 0, len(paymentMethods))
	for _, pm := range paymentMethods {
		responses = append(responses, *s.mapEntityToResponse(&pm))
	}

	return &dto.ListPaymentMethodsResponse{
		PaymentMethods: responses,
		Total:          len(responses),
	}, nil
}

// GetPaymentMethod retrieves a specific payment method
func (s *paymentMethodService) GetPaymentMethod(
	ctx context.Context,
	organizationID, paymentMethodID uuid.UUID,
) (*dto.PaymentMethodResponse, error) {
	paymentMethod, err := s.repo.GetByID(paymentMethodID)
	if err != nil {
		s.log.Error("Failed to get payment method", zap.Error(err), zap.String("payment_method_id", paymentMethodID.String()))
		return nil, response.NewInternalServerError("failed to get payment method")
	}

	if paymentMethod == nil {
		return nil, response.ErrNotFound
	}

	// Verify the payment method belongs to the organization
	if paymentMethod.OrganizationID != organizationID {
		return nil, response.ErrForbidden
	}

	return s.mapEntityToResponse(paymentMethod), nil
}

// DeletePaymentMethod deletes a payment method
func (s *paymentMethodService) DeletePaymentMethod(
	ctx context.Context,
	organizationID, paymentMethodID uuid.UUID,
) error {
	// Get payment method to verify ownership
	paymentMethod, err := s.repo.GetByID(paymentMethodID)
	if err != nil {
		s.log.Error("Failed to get payment method for deletion", zap.Error(err), zap.String("payment_method_id", paymentMethodID.String()))
		return response.NewInternalServerError("failed to get payment method")
	}

	if paymentMethod == nil {
		return response.ErrNotFound
	}

	// Verify the payment method belongs to the organization
	if paymentMethod.OrganizationID != organizationID {
		return response.ErrForbidden
	}

	// Get the provider for this payment method
	providerType := payment.ProviderType(paymentMethod.Provider)
	provider, err := s.providerManager.GetProvider(providerType)
	if err != nil {
		s.log.Warn("Failed to get payment provider for deletion, continuing with database deletion", 
			zap.Error(err), 
			zap.String("provider", paymentMethod.Provider),
			zap.String("payment_method_id", paymentMethodID.String()))
	} else {
		// Delete from provider first
		if err := provider.DeletePaymentMethod(ctx, paymentMethod.ProviderPaymentMethodID); err != nil {
			s.log.Error("Failed to delete payment method from provider", 
				zap.Error(err), 
				zap.String("provider_payment_method_id", paymentMethod.ProviderPaymentMethodID),
				zap.String("provider", paymentMethod.Provider))
			// Continue with database deletion even if provider deletion fails
			// (payment method might already be deleted in provider)
		}
	}

	// Delete from database
	if err := s.repo.Delete(paymentMethodID); err != nil {
		s.log.Error("Failed to delete payment method from database", zap.Error(err), zap.String("payment_method_id", paymentMethodID.String()))
		return response.NewInternalServerError("failed to delete payment method")
	}

	return nil
}

// SetDefaultPaymentMethod sets a payment method as the default for an organization
func (s *paymentMethodService) SetDefaultPaymentMethod(
	ctx context.Context,
	organizationID, paymentMethodID uuid.UUID,
) error {
	// Verify the payment method exists and belongs to the organization
	paymentMethod, err := s.repo.GetByID(paymentMethodID)
	if err != nil {
		s.log.Error("Failed to get payment method", zap.Error(err), zap.String("payment_method_id", paymentMethodID.String()))
		return response.NewInternalServerError("failed to get payment method")
	}

	if paymentMethod == nil {
		return response.ErrNotFound
	}

	// Verify the payment method belongs to the organization
	if paymentMethod.OrganizationID != organizationID {
		return response.ErrForbidden
	}

	// Set as default
	if err := s.repo.SetDefault(organizationID, paymentMethodID); err != nil {
		s.log.Error("Failed to set default payment method", zap.Error(err), zap.String("payment_method_id", paymentMethodID.String()))
		return response.NewInternalServerError("failed to set default payment method")
	}

	return nil
}

// mapEntityToResponse maps a payment method entity to a response DTO
func (s *paymentMethodService) mapEntityToResponse(pm *paymentEntity.PaymentMethod) *dto.PaymentMethodResponse {
	return &dto.PaymentMethodResponse{
		ID:          pm.ID,
		Provider:   pm.Provider,
		Last4:      pm.Last4,
		Brand:      pm.Brand,
		ExpiryMonth: pm.ExpiryMonth,
		ExpiryYear:  pm.ExpiryYear,
		IsDefault:  pm.IsDefault,
		CreatedAt:  pm.CreatedAt,
		UpdatedAt:  pm.UpdatedAt,
	}
}

