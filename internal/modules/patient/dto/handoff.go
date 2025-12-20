package dto

import (
	"time"

	"github.com/google/uuid"
)

// RequestHandoffRequest represents a request to hand off a patient
type RequestHandoffRequest struct {
	ReceivingClinicianID uuid.UUID `json:"receiving_clinician_id" binding:"required"`
	Message              *string   `json:"message"`
	Role                 *string   `json:"role"` // Optional: role receiving clinician should get (inherits if not provided)
}

// ApproveHandoffRequest represents a request to approve a handoff
type ApproveHandoffRequest struct {
	Reason *string `json:"reason"` // Optional reason for approval
}

// RejectHandoffRequest represents a request to reject a handoff
type RejectHandoffRequest struct {
	Reason *string `json:"reason" binding:"required"` // Reason for rejection
}

// HandoffResponse represents a patient handoff with full details
type HandoffResponse struct {
	ID                    uuid.UUID  `json:"id"`
	PatientID             uuid.UUID  `json:"patient_id"`
	PatientName           string     `json:"patient_name"` // First + Last name
	RequestingClinicianID uuid.UUID  `json:"requesting_clinician_id"`
	RequestingClinicianName string   `json:"requesting_clinician_name"`
	RequestingClinicianEmail string  `json:"requesting_clinician_email"`
	ReceivingClinicianID  uuid.UUID  `json:"receiving_clinician_id"`
	ReceivingClinicianName string    `json:"receiving_clinician_name"`
	ReceivingClinicianEmail string   `json:"receiving_clinician_email"`
	Status                string     `json:"status"`
	RequestedRole         *string    `json:"requested_role"`
	Message               *string    `json:"message"`
	RequestedAt           time.Time  `json:"requested_at"`
	RespondedAt           *time.Time `json:"responded_at"`
	RespondedBy            *uuid.UUID `json:"responded_by"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// HandoffListResponse represents a list of handoffs
type HandoffListResponse struct {
	Handoffs []HandoffResponse `json:"handoffs"`
	Total    int64             `json:"total"`
}

