package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateClinicalNoteRequest struct {
	PatientID     uuid.UUID  `json:"patient_id"     validate:"required"`
	ClinicianID   uuid.UUID  `json:"clinician_id"   validate:"required"`
	AppointmentID *uuid.UUID `json:"appointment_id"`
	NoteType      string     `json:"note_type"      validate:"required"`
	Subjective    *string    `json:"subjective"`
	Objective     *string    `json:"objective"`
	Assessment    *string    `json:"assessment"`
	Plan          *string    `json:"plan"`
	IsSigned      bool       `json:"is_signed"`
}

type UpdateClinicalNoteRequest struct {
	NoteType   string  `json:"note_type"`
	Subjective *string `json:"subjective"`
	Objective  *string `json:"objective"`
	Assessment *string `json:"assessment"`
	Plan       *string `json:"plan"`
	IsSigned   *bool   `json:"is_signed"`
}

type ClinicalNoteResponse struct {
	ID             uuid.UUID            `json:"id"`
	OrganizationID uuid.UUID            `json:"organization_id"`
	PatientID      uuid.UUID            `json:"patient_id"`
	ClinicianID    uuid.UUID            `json:"clinician_id"`
	AppointmentID  *uuid.UUID           `json:"appointment_id"`
	NoteType       string               `json:"note_type"`
	Subjective     *string              `json:"subjective"`
	Objective      *string              `json:"objective"`
	Assessment     *string              `json:"assessment"`
	Plan           *string              `json:"plan"`
	IsSigned       bool                 `json:"is_signed"`
	SignedAt       *time.Time           `json:"signed_at"`
	Addendums      []AddendumResponse   `json:"addendums,omitempty"`
	Attachments    []AttachmentResponse `json:"attachments,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type AddAddendumRequest struct {
	Content     string    `json:"content"      validate:"required"`
	ClinicianID uuid.UUID `json:"clinician_id" validate:"required"`
}

type AddendumResponse struct {
	ID          uuid.UUID `json:"id"`
	ClinicianID uuid.UUID `json:"clinician_id"`
	Content     string    `json:"content"`
	SignedAt    time.Time `json:"signed_at"`
}

type AttachmentResponse struct {
	ID          uuid.UUID `json:"id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}
