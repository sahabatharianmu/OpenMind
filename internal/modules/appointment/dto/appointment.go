package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateAppointmentRequest struct {
	PatientID   uuid.UUID `json:"patient_id" validate:"required"`
	ClinicianID uuid.UUID `json:"clinician_id" validate:"required"`
	StartTime   string    `json:"start_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"` // RFC3339
	EndTime     string    `json:"end_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`   // RFC3339
	Status      string    `json:"status" validate:"omitempty,oneof=scheduled completed cancelled no-show"`
	Type        string    `json:"appointment_type" validate:"required"`
	Mode        string    `json:"mode" validate:"required,oneof=in-person video phone"`
	Notes       *string   `json:"notes"`
}

type UpdateAppointmentRequest struct {
	StartTime *string `json:"start_time" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndTime   *string `json:"end_time" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Status    string  `json:"status" validate:"omitempty,oneof=scheduled completed cancelled no-show"`
	Type      string  `json:"appointment_type"`
	Mode      string  `json:"mode" validate:"omitempty,oneof=in-person video phone"`
	Notes     *string `json:"notes"`
}

type AppointmentResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	PatientID      uuid.UUID `json:"patient_id"`
	ClinicianID    uuid.UUID `json:"clinician_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Status         string    `json:"status"`
	Type           string    `json:"appointment_type"`
	Mode           string    `json:"mode"`
	Notes          *string   `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
