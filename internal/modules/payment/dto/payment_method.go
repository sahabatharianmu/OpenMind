package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreatePaymentMethodRequest represents the request to create a payment method
type CreatePaymentMethodRequest struct {
	Token    string `json:"token"              binding:"required"` // Payment method token from frontend (e.g., Stripe PaymentMethod ID)
	Provider string `json:"provider,omitempty"`                    // Optional: payment provider (stripe, square). Defaults to configured default provider
}

// PaymentMethodResponse represents a payment method in API responses
// Note: Does not include sensitive data like encrypted_token
type PaymentMethodResponse struct {
	ID          uuid.UUID `json:"id"`
	Provider    string    `json:"provider"` // stripe, square
	Last4       string    `json:"last4"`
	Brand       string    `json:"brand"` // visa, mastercard, etc.
	ExpiryMonth int       `json:"expiry_month"`
	ExpiryYear  int       `json:"expiry_year"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdatePaymentMethodRequest represents the request to update a payment method
type UpdatePaymentMethodRequest struct {
	IsDefault *bool `json:"is_default,omitempty"` // Set as default payment method
}

// ListPaymentMethodsResponse represents the response for listing payment methods
type ListPaymentMethodsResponse struct {
	PaymentMethods []PaymentMethodResponse `json:"payment_methods"`
	Total          int                     `json:"total"`
}
