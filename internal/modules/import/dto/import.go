package dto

import "github.com/google/uuid"

// ImportPreviewRequest represents a request to preview an import
type ImportPreviewRequest struct {
	Type     string `json:"type"      binding:"required,oneof=patients notes"`
	FileData string `json:"file_data" binding:"required"` // Base64 encoded file
	FileName string `json:"file_name" binding:"required"`
}

// ImportExecuteRequest represents a request to execute an import
type ImportExecuteRequest struct {
	Type     string `json:"type"      binding:"required,oneof=patients notes"`
	FileData string `json:"file_data" binding:"required"` // Base64 encoded file
	FileName string `json:"file_name" binding:"required"`
}

// ImportPreviewResponse represents the preview results
type ImportPreviewResponse struct {
	TotalRows   int                      `json:"total_rows"`
	ValidRows   int                      `json:"valid_rows"`
	InvalidRows int                      `json:"invalid_rows"`
	Preview     []map[string]interface{} `json:"preview"` // First 10 valid rows
	Errors      []RowError               `json:"errors,omitempty"`
	Warnings    []RowWarning             `json:"warnings,omitempty"`
}

// ImportExecuteResponse represents the import execution results
type ImportExecuteResponse struct {
	TotalRows    int         `json:"total_rows"`
	SuccessCount int         `json:"success_count"`
	ErrorCount   int         `json:"error_count"`
	Errors       []RowError  `json:"errors,omitempty"`
	ImportedIDs  []uuid.UUID `json:"imported_ids,omitempty"`
}

// RowError represents an error for a specific row
type RowError struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// RowWarning represents a warning for a specific row
type RowWarning struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// PatientCSVRow represents a patient row in CSV format
type PatientCSVRow struct {
	FirstName   string `csv:"first_name"`
	LastName    string `csv:"last_name"`
	DateOfBirth string `csv:"date_of_birth"`
	Email       string `csv:"email"`
	Phone       string `csv:"phone"`
	Address     string `csv:"address"`
}

// ClinicalNoteCSVRow represents a clinical note row in CSV/XLSX format
type ClinicalNoteCSVRow struct {
	PatientID     string `csv:"patient_id"`
	AppointmentID string `csv:"appointment_id"`
	NoteType      string `csv:"note_type"`
	ICD10Code     string `csv:"icd10_code"`
	Subjective    string `csv:"subjective"`
	Objective     string `csv:"objective"`
	Assessment    string `csv:"assessment"`
	Plan          string `csv:"plan"`
	IsSigned      string `csv:"is_signed"`
}
