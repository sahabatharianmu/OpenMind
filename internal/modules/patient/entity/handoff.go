package entity

import (
	"time"

	"github.com/google/uuid"
)

// PatientHandoff represents a patient handoff request between clinicians
type PatientHandoff struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PatientID             uuid.UUID  `gorm:"type:uuid;not null"                             json:"patient_id"`
	RequestingClinicianID uuid.UUID  `gorm:"type:uuid;not null"                             json:"requesting_clinician_id"`
	ReceivingClinicianID  uuid.UUID  `gorm:"type:uuid;not null"                             json:"receiving_clinician_id"`
	Status                string     `gorm:"type:varchar(50);not null;default:'requested'"  json:"status"`
	RequestedRole         *string    `gorm:"type:varchar(50)"                               json:"requested_role"` // Role receiving clinician should get (inherits if null)
	Message               *string    `gorm:"type:text"                                      json:"message"`
	RequestedAt           time.Time  `gorm:"autoCreateTime"                                 json:"requested_at"`
	RespondedAt           *time.Time `gorm:""                                               json:"responded_at"`
	RespondedBy           *uuid.UUID `gorm:"type:uuid"                                      json:"responded_by"`
	CreatedAt             time.Time  `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (PatientHandoff) TableName() string {
	return "patient_handoffs"
}

// Handoff status constants
const (
	StatusRequested = "requested"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusCancelled = "cancelled"
)

// IsPending returns true if the handoff is in a pending state (requested)
func (h *PatientHandoff) IsPending() bool {
	return h.Status == StatusRequested
}

// CanBeApproved returns true if the handoff can be approved
func (h *PatientHandoff) CanBeApproved() bool {
	return h.IsPending()
}

// CanBeRejected returns true if the handoff can be rejected
func (h *PatientHandoff) CanBeRejected() bool {
	return h.IsPending()
}

// CanBeCancelled returns true if the handoff can be cancelled
func (h *PatientHandoff) CanBeCancelled() bool {
	return h.IsPending()
}

// IsFinal returns true if the handoff is in a final state
func (h *PatientHandoff) IsFinal() bool {
	return h.Status == StatusApproved || h.Status == StatusRejected || h.Status == StatusCancelled
}
