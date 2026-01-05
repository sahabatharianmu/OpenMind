package payment

import (
	"context"
	"fmt"

	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentmethod"
	"go.uber.org/zap"
)

// stripeProvider implements PaymentProvider for Stripe
type stripeProvider struct {
	config *config.StripeConfig
	log    logger.Logger
}

// NewStripeProvider creates a new Stripe payment provider
func NewStripeProvider(cfg *config.StripeConfig, log logger.Logger) PaymentProvider {
	// Initialize Stripe with secret key
	stripe.Key = cfg.SecretKey

	return &stripeProvider{
		config: cfg,
		log:    log,
	}
}

// CreatePaymentMethod creates a payment method in Stripe
// token is the Stripe PaymentMethod ID (pm_xxx)
func (p *stripeProvider) CreatePaymentMethod(ctx context.Context, token string) (string, error) {
	// Retrieve the payment method to verify it exists and get details
	pm, err := paymentmethod.Get(token, nil)
	if err != nil {
		p.log.Error("Failed to retrieve payment method from Stripe", zap.Error(err), zap.String("token", token))
		return "", fmt.Errorf("failed to retrieve payment method: %w", err)
	}

	// Return the payment method ID
	return pm.ID, nil
}

// DeletePaymentMethod removes a payment method from Stripe
func (p *stripeProvider) DeletePaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		p.log.Error(
			"Failed to delete payment method from Stripe",
			zap.Error(err),
			zap.String("payment_method_id", paymentMethodID),
		)
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	return nil
}

// ListPaymentMethods retrieves all payment methods for the organization
// Note: Stripe doesn't have a direct way to list all payment methods for an organization
// This would typically require storing customer IDs. For now, we'll return an error
// indicating this needs to be handled at the service layer.
func (p *stripeProvider) ListPaymentMethods(ctx context.Context) ([]PaymentMethodInfo, error) {
	// This method is not directly supported by Stripe without a customer ID
	// The service layer should handle listing by querying our database
	return nil, fmt.Errorf("list payment methods not supported directly by Stripe provider - use service layer")
}

// GetPaymentMethod retrieves a specific payment method by ID
func (p *stripeProvider) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethodInfo, error) {
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		p.log.Error(
			"Failed to get payment method from Stripe",
			zap.Error(err),
			zap.String("payment_method_id", paymentMethodID),
		)
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}

	// Convert Stripe payment method to our standard format
	info := &PaymentMethodInfo{
		ID:          pm.ID,
		Last4:       pm.Card.Last4,
		Brand:       string(pm.Card.Brand),
		ExpiryMonth: int(pm.Card.ExpMonth),
		ExpiryYear:  int(pm.Card.ExpYear),
	}

	return info, nil
}
