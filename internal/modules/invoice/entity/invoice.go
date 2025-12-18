package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Invoice struct {
	ID             uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null"                           json:"organization_id"`
	PatientID      uuid.UUID      `gorm:"type:uuid;not null"                           json:"patient_id"`
	AppointmentID  *uuid.UUID     `gorm:"type:uuid"                                    json:"appointment_id"`
	AmountCents    int            `gorm:"not null"                                     json:"amount_cents"`
	Status         string         `gorm:"not null;default:'pending'"                   json:"status"`
	DueDate        *time.Time     `gorm:""                                             json:"due_date"`
	PaidAt         *time.Time     `gorm:""                                             json:"paid_at"`
	PaymentMethod  *string        `gorm:""                                             json:"payment_method"`
	Notes          *string        `gorm:""                                             json:"notes"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                               json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                               json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                                        json:"-"`
}

func (Invoice) TableName() string {
	return "invoices"
}
