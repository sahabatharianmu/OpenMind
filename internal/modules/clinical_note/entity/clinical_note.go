package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClinicalNote struct {
	ID             uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null"                           json:"organization_id"`
	PatientID      uuid.UUID      `gorm:"type:uuid;not null"                           json:"patient_id"`
	ClinicianID    uuid.UUID      `gorm:"type:uuid;not null"                           json:"clinician_id"`
	AppointmentID  *uuid.UUID     `gorm:"type:uuid"                                    json:"appointment_id"`
	NoteType       string         `gorm:"not null"                                     json:"note_type"`
	Subjective     *string        `gorm:""                                             json:"subjective"`
	Objective      *string        `gorm:""                                             json:"objective"`
	Assessment     *string        `gorm:""                                             json:"assessment"`
	Plan           *string        `gorm:""                                             json:"plan"`
	IsSigned       bool           `gorm:"not null;default:false"                       json:"is_signed"`
	SignedAt       *time.Time     `gorm:""                                             json:"signed_at"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                               json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                               json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                                        json:"-"`
}

func (ClinicalNote) TableName() string {
	return "clinical_notes"
}
