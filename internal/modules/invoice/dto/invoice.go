package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateInvoiceRequest struct {
	PatientID     uuid.UUID  `json:"patient_id" validate:"required"`
	AppointmentID *uuid.UUID `json:"appointment_id"`
	AmountCents   int        `json:"amount_cents" validate:"required,min=0"`
	Status        string     `json:"status" validate:"omitempty,oneof=pending paid void overdue"`
	DueDate       *string    `json:"due_date" validate:"omitempty"`
	PaidAt        *string    `json:"paid_at" validate:"omitempty"`
	PaymentMethod *string    `json:"payment_method"`
	Notes         *string    `json:"notes"`
}

type UpdateInvoiceRequest struct {
	AmountCents   *int    `json:"amount_cents" validate:"omitempty,min=0"`
	Status        string  `json:"status" validate:"omitempty,oneof=pending paid void overdue"`
	DueDate       *string `json:"due_date" validate:"omitempty"`
	PaidAt        *string `json:"paid_at" validate:"omitempty"`
	PaymentMethod *string `json:"payment_method"`
	Notes         *string `json:"notes"`
}

type InvoiceResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	PatientID      uuid.UUID  `json:"patient_id"`
	AppointmentID  *uuid.UUID `json:"appointment_id"`
	AmountCents    int        `json:"amount_cents"`
	Status         string     `json:"status"`
	DueDate        *time.Time `json:"due_date"`
	PaidAt         *time.Time `json:"paid_at"`
	PaymentMethod  *string    `json:"payment_method"`
	Notes          *string    `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
