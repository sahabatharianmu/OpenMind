package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentTransaction represents a payment transaction (e.g., QRIS payment, subscription payment)
type PaymentTransaction struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"  json:"id"`
	OrganizationID        uuid.UUID      `gorm:"type:uuid;not null;index"                          json:"organization_id"`
	Type                  string         `gorm:"type:varchar(50);not null"                         json:"type"`           // subscription, one_time, etc.
	PaymentMethod         string         `gorm:"type:varchar(50);not null"                         json:"payment_method"` // qris, credit_card, virtual_account
	Provider              string         `gorm:"type:varchar(50);not null"                         json:"provider"`       // midtrans, stripe, etc.
	Amount                int64          `gorm:"not null"                                          json:"amount"`         // Amount in cents (e.g., 2900 for $29.00)
	Currency              string         `gorm:"type:varchar(10);not null;default:'USD'"           json:"currency"`
	Status                string         `gorm:"type:varchar(50);not null;default:'pending';index" json:"status"`                  // pending, paid, failed, cancelled, expired
	ProviderTransactionID string         `gorm:"type:varchar(255);index"                           json:"provider_transaction_id"` // Transaction ID from provider
	PartnerReferenceNo    string         `gorm:"type:varchar(255);uniqueIndex"                     json:"partner_reference_no"`    // Our reference number
	ExternalID            string         `gorm:"type:varchar(255);index"                           json:"external_id"`             // External ID for provider
	QRCode                string         `gorm:"type:text"                                         json:"qr_code,omitempty"`       // QR code string for QRIS
	QRCodeURL             string         `gorm:"type:text"                                         json:"qr_code_url,omitempty"`   // QR code image URL
	QRCodeImage           string         `gorm:"type:text"                                         json:"qr_code_image,omitempty"` // QR code image base64
	ExpiresAt             *time.Time     `gorm:"index"                                             json:"expires_at,omitempty"`    // When the payment expires
	PaidAt                *time.Time     `                                                         json:"paid_at,omitempty"`       // When the payment was completed
	Metadata              *string        `gorm:"type:jsonb"                                        json:"metadata,omitempty"`      // Additional metadata (JSON)
	CreatedAt             time.Time      `                                                         json:"created_at"`
	UpdatedAt             time.Time      `                                                         json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index"                                             json:"deleted_at,omitempty"`
}

// TableName specifies the table name for GORM
func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}
