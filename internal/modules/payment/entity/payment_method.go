package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentMethod represents a payment method stored in the tenant schema
type PaymentMethod struct {
	ID                      uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	OrganizationID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	Provider                string         `gorm:"type:varchar(50);not null;index" json:"provider"` // stripe, square
	EncryptedToken          []byte         `gorm:"type:bytea;not null" json:"-"`                    // Encrypted payment method token
	ProviderPaymentMethodID string         `gorm:"type:varchar(255);not null" json:"provider_payment_method_id"` // Provider's payment method ID
	Last4                   string         `gorm:"type:varchar(4);not null" json:"last4"`
	Brand                   string         `gorm:"type:varchar(50);not null" json:"brand"` // visa, mastercard, etc.
	ExpiryMonth             int            `gorm:"not null" json:"expiry_month"`          // 1-12
	ExpiryYear              int            `gorm:"not null" json:"expiry_year"`           // e.g., 2025
	IsDefault               bool           `gorm:"not null;default:false;index" json:"is_default"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for GORM
func (PaymentMethod) TableName() string {
	return "payment_methods"
}

