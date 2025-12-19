package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClinicalNote struct {
	ID               uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	OrganizationID   uuid.UUID      `gorm:"type:uuid;not null"                              json:"organization_id"`
	PatientID        uuid.UUID      `gorm:"type:uuid;not null"                              json:"patient_id"`
	ClinicianID      uuid.UUID      `gorm:"type:uuid;not null"                              json:"clinician_id"`
	AppointmentID    *uuid.UUID     `gorm:"type:uuid"                                       json:"appointment_id"`
	NoteType         string         `gorm:"not null"                                        json:"note_type"`
	ICD10Code        string         `gorm:"type:varchar(20)"                                json:"icd10_code"`
	Subjective       *string        `gorm:"-"                                               json:"subjective"`
	Objective        *string        `gorm:"-"                                               json:"objective"`
	Assessment       *string        `gorm:"-"                                               json:"assessment"`
	Plan             *string        `gorm:"-"                                               json:"plan"`
	ContentEncrypted []byte         `gorm:"type:bytea"                                      json:"-"`
	KeyID            string         `gorm:"type:varchar(255)"                               json:"key_id"`
	Nonce            []byte         `gorm:"type:bytea"                                      json:"-"`
	IsSigned         bool           `gorm:"not null;default:false"                          json:"is_signed"`
	SignedAt         *time.Time     `gorm:""                                                json:"signed_at"`
	Addendums        []Addendum     `gorm:"foreignKey:NoteID"                               json:"addendums,omitempty"`
	Attachments      []Attachment   `gorm:"foreignKey:NoteID"                               json:"attachments,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"                                  json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index"                                           json:"-"`
}

type Addendum struct {
	ID               uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	NoteID           uuid.UUID `gorm:"type:uuid;not null"                              json:"note_id"`
	ClinicianID      uuid.UUID `gorm:"type:uuid;not null"                              json:"clinician_id"`
	Content          string    `gorm:"-"                                               json:"content"`
	ContentEncrypted []byte    `gorm:"type:bytea"                                      json:"-"`
	Nonce            []byte    `gorm:"type:bytea"                                      json:"-"`
	SignedAt         time.Time `gorm:"not null;autoCreateTime"                         json:"signed_at"`
}

func (Addendum) TableName() string {
	return "clinical_note_addendums"
}

type Attachment struct {
	ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	NoteID        uuid.UUID `gorm:"type:uuid;not null"                              json:"note_id"`
	FileName      string    `gorm:"type:varchar(255);not null"                      json:"file_name"`
	ContentType   string    `gorm:"type:varchar(100);not null"                      json:"content_type"`
	Size          int64     `gorm:"not null"                                        json:"size"`
	DataEncrypted []byte    `gorm:"type:bytea;not null"                             json:"-"`
	Nonce         []byte    `gorm:"type:bytea;not null"                             json:"-"`
	CreatedAt     time.Time `gorm:"autoCreateTime"                                  json:"created_at"`
}

func (Attachment) TableName() string {
	return "clinical_note_attachments"
}

func (ClinicalNote) TableName() string {
	return "clinical_notes"
}

// BeforeUpdate is a GORM hook that prevents updates to signed notes
// This provides application-level protection in addition to database triggers
func (n *ClinicalNote) BeforeUpdate(tx *gorm.DB) error {
	// Check if the note was already signed before this update
	// We need to query the database to get the current state
	var existingNote ClinicalNote
	if err := tx.Unscoped().Where("id = ?", n.ID).First(&existingNote).Error; err != nil {
		// If note doesn't exist, allow the operation (it will fail elsewhere)
		return nil
	}

	// If the note was already signed, prevent the update
	if existingNote.IsSigned {
		return errors.New("cannot update a signed clinical note: signed notes are immutable for compliance")
	}

	return nil
}

// BeforeDelete is a GORM hook that prevents deletion of signed notes
// This provides application-level protection in addition to database triggers
func (n *ClinicalNote) BeforeDelete(tx *gorm.DB) error {
	// Check if the note is signed
	// We need to query the database to get the current state
	var existingNote ClinicalNote
	if err := tx.Unscoped().Where("id = ?", n.ID).First(&existingNote).Error; err != nil {
		// If note doesn't exist, allow the operation (it will fail elsewhere)
		return nil
	}

	// If the note is signed, prevent deletion
	if existingNote.IsSigned {
		return errors.New("cannot delete a signed clinical note: signed notes are immutable for compliance")
	}

	return nil
}
