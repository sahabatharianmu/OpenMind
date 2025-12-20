package payment

import (
	"context"
)

// PaymentProvider defines the interface for payment method operations
// This abstraction allows easy addition of new payment providers (Stripe, Square, etc.)
type PaymentProvider interface {
	// CreatePaymentMethod creates a payment method in the provider system
	// token is the payment method token from the frontend (e.g., Stripe PaymentMethod ID)
	// Returns the provider's payment method ID
	CreatePaymentMethod(ctx context.Context, token string) (string, error)

	// DeletePaymentMethod removes a payment method from the provider
	DeletePaymentMethod(ctx context.Context, paymentMethodID string) error

	// ListPaymentMethods retrieves all payment methods for the organization
	// Returns standardized payment method information
	ListPaymentMethods(ctx context.Context) ([]PaymentMethodInfo, error)

	// GetPaymentMethod retrieves a specific payment method by ID
	GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethodInfo, error)
}

// PaymentMethodInfo contains standardized payment method information
// This is returned by all payment providers in a consistent format
type PaymentMethodInfo struct {
	ID          string // Provider's payment method ID
	Last4       string // Last 4 digits of card
	Brand       string // Card brand (visa, mastercard, etc.)
	ExpiryMonth int    // Expiry month (1-12)
	ExpiryYear  int    // Expiry year (e.g., 2025)
}

// ProviderType represents the type of payment provider
type ProviderType string

const (
	ProviderStripe   ProviderType = "stripe"
	ProviderSquare   ProviderType = "square"
	ProviderMidtrans ProviderType = "midtrans"
)

// WebhookResult represents the result of processing a payment webhook
type WebhookResult struct {
	GatewayTransactionID string       // Transaction ID from payment gateway
	PaymentCode          string       // Our internal payment reference code
	Status               PaymentStatus // Payment status
	RawPayload           string       // Raw webhook payload for debugging
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

