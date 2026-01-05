package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateQRISPaymentRequest represents the request to create a QRIS payment
type CreateQRISPaymentRequest struct {
	Amount   float64 `json:"amount"   binding:"required,min=0.01"` // Amount in USD
	Currency string  `json:"currency" binding:"required"`          // Currency code (USD, IDR, etc.)
	Type     string  `json:"type"     binding:"required"`          // subscription, one_time
}

// QRISPaymentResponse represents the response for QRIS payment creation
type QRISPaymentResponse struct {
	ID                 uuid.UUID  `json:"id"`
	TransactionID      string     `json:"transaction_id"`       // Provider transaction ID
	PartnerReferenceNo string     `json:"partner_reference_no"` // Our reference number
	QRCode             string     `json:"qr_code"`              // QR code string
	QRCodeURL          string     `json:"qr_code_url"`          // QR code URL
	QRCodeImage        string     `json:"qr_code_image"`        // QR code image base64
	Amount             float64    `json:"amount"`               // Amount in USD
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// CheckPaymentStatusResponse represents the response for payment status check
type CheckPaymentStatusResponse struct {
	ID                      uuid.UUID  `json:"id"`
	TransactionID           string     `json:"transaction_id"`
	Status                  string     `json:"status"`
	Amount                  float64    `json:"amount"`
	Currency                string     `json:"currency"`
	PaidAt                  *time.Time `json:"paid_at,omitempty"`
	LatestTransactionStatus string     `json:"latest_transaction_status,omitempty"` // From provider
	TransactionStatusDesc   string     `json:"transaction_status_desc,omitempty"`   // From provider
}
