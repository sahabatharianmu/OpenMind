package payment

import (
	"fmt"

	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
)

// NewPaymentProvider creates a payment provider instance based on configuration
// This factory pattern allows easy addition of new payment providers
func NewPaymentProvider(cfg *config.PaymentConfig, log logger.Logger) (PaymentProvider, error) {
	switch cfg.Provider {
	case "stripe":
		if cfg.Stripe.SecretKey == "" {
			return nil, fmt.Errorf("stripe secret key is required")
		}
		return NewStripeProvider(&cfg.Stripe, log), nil

	case "square":
		// Placeholder for future Square implementation
		return nil, fmt.Errorf("square payment provider not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", cfg.Provider)
	}
}

