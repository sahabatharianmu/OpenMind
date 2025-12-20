package payment

import (
	"fmt"
	"sync"

	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

// PaymentProviderManager manages multiple payment provider instances
// This allows organizations to use different payment gateways simultaneously
type PaymentProviderManager struct {
	providers map[ProviderType]PaymentProvider
	defaultProvider ProviderType
	log logger.Logger
	mu sync.RWMutex
}

// NewPaymentProviderManager creates a new payment provider manager
// Initializes all configured payment providers
func NewPaymentProviderManager(cfg *config.PaymentConfig, log logger.Logger) (*PaymentProviderManager, error) {
	manager := &PaymentProviderManager{
		providers: make(map[ProviderType]PaymentProvider),
		defaultProvider: ProviderType(cfg.Provider),
		log: log,
	}

	// Initialize Stripe provider if configured
	if cfg.Stripe.SecretKey != "" {
		stripeProvider := NewStripeProvider(&cfg.Stripe, log)
		manager.providers[ProviderStripe] = stripeProvider
		log.Info("Stripe payment provider initialized")
	}

	// Initialize Square provider if configured (when implemented)
	if cfg.Square.AccessToken != "" {
		// TODO: Uncomment when Square provider is implemented
		// squareProvider := NewSquareProvider(&cfg.Square, log)
		// manager.providers[ProviderSquare] = squareProvider
		log.Info("Square payment provider configuration found but not yet implemented")
	}

	// Initialize Midtrans provider if configured
	if cfg.Midtrans.BISnapClientID != "" && cfg.Midtrans.BISnapClientSecret != "" {
		midtransProvider, err := NewMidtransProvider(&cfg.Midtrans, log)
		if err != nil {
			log.Warn("Failed to initialize Midtrans provider", zap.Error(err))
		} else {
			manager.providers[ProviderMidtrans] = midtransProvider
			log.Info("Midtrans payment provider initialized")
		}
	}

	// Validate that at least one provider is configured
	if len(manager.providers) == 0 {
		return nil, fmt.Errorf("no payment providers configured")
	}

	// Validate that default provider is configured
	if _, exists := manager.providers[manager.defaultProvider]; !exists {
		return nil, fmt.Errorf("default payment provider '%s' is not configured", manager.defaultProvider)
	}

	return manager, nil
}

// GetProvider returns the payment provider for the given provider type
// Returns the default provider if the requested provider is not found
func (m *PaymentProviderManager) GetProvider(providerType ProviderType) (PaymentProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[providerType]
	if !exists {
		// Fallback to default provider
		m.log.Warn("Requested payment provider not found, using default", 
			zap.String("requested", string(providerType)),
			zap.String("default", string(m.defaultProvider)))
		provider, exists = m.providers[m.defaultProvider]
		if !exists {
			return nil, fmt.Errorf("payment provider '%s' not configured", providerType)
		}
	}

	return provider, nil
}

// GetDefaultProvider returns the default payment provider
func (m *PaymentProviderManager) GetDefaultProvider() (PaymentProvider, error) {
	return m.GetProvider(m.defaultProvider)
}

// GetDefaultProviderType returns the default provider type
func (m *PaymentProviderManager) GetDefaultProviderType() ProviderType {
	return m.defaultProvider
}

// IsProviderAvailable checks if a specific provider is configured
func (m *PaymentProviderManager) IsProviderAvailable(providerType ProviderType) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.providers[providerType]
	return exists
}

// GetAvailableProviders returns a list of all configured provider types
func (m *PaymentProviderManager) GetAvailableProviders() []ProviderType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	providers := make([]ProviderType, 0, len(m.providers))
	for providerType := range m.providers {
		providers = append(providers, providerType)
	}
	return providers
}

