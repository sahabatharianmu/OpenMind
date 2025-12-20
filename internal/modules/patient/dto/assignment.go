package dto

import (
	"time"

	"github.com/google/uuid"
)

type AssignClinicianRequest struct {
	ClinicianID uuid.UUID `json:"clinician_id" binding:"required"`
	Role        string    `json:"role" binding:"required,oneof=primary secondary"`
}

type ClinicianAssignmentResponse struct {
	ClinicianID uuid.UUID `json:"clinician_id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	AssignedAt  time.Time `json:"assigned_at"`
	AssignedBy  uuid.UUID `json:"assigned_by"`
}

