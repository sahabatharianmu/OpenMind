package entity

import (
	"time"

	"github.com/google/uuid"
)

// PatientAssignment represents the assignment of a clinician to a patient
type PatientAssignment struct {
	PatientID   uuid.UUID `gorm:"primaryKey;type:uuid" json:"patient_id"`
	ClinicianID uuid.UUID `gorm:"primaryKey;type:uuid" json:"clinician_id"`
	Role        string    `gorm:"type:varchar(50);not null;default:'primary'" json:"role"` // "primary" or "secondary"
	AssignedBy  uuid.UUID `gorm:"type:uuid;not null" json:"assigned_by"`
	AssignedAt  time.Time `gorm:"autoCreateTime" json:"assigned_at"`
}

func (PatientAssignment) TableName() string {
	return "assigned_clinicians"
}

