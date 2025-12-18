package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Appointment struct {
	ID             uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null"                              json:"organization_id"`
	PatientID      uuid.UUID      `gorm:"type:uuid;not null"                              json:"patient_id"`
	ClinicianID    uuid.UUID      `gorm:"type:uuid;not null"                              json:"clinician_id"`
	StartTime      time.Time      `gorm:"not null"                                        json:"start_time"`
	EndTime        time.Time      `gorm:"not null"                                        json:"end_time"`
	Status         string         `gorm:"not null;default:'scheduled'"                    json:"status"`
	Type           string         `gorm:"column:appointment_type;not null"                json:"appointment_type"`
	Mode           string         `gorm:"not null"                                        json:"mode"`
	Notes          *string        `gorm:""                                                json:"notes"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                                  json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                                           json:"-"`
}

func (Appointment) TableName() string {
	return "appointments"
}
