package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	clinicalNoteEntity "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/entity"
	clinicalNoteRepository "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/repository"
	clinicalNoteService "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/service"
	"github.com/sahabatharianmu/OpenMind/internal/modules/import/dto"
	patientEntity "github.com/sahabatharianmu/OpenMind/internal/modules/patient/entity"
	patientRepository "github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	patientService "github.com/sahabatharianmu/OpenMind/internal/modules/patient/service"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ImportService interface {
	PreviewImport(
		ctx context.Context,
		req dto.ImportPreviewRequest,
		organizationID uuid.UUID,
		userID uuid.UUID,
	) (*dto.ImportPreviewResponse, error)
	ExecuteImport(
		ctx context.Context,
		req dto.ImportExecuteRequest,
		organizationID uuid.UUID,
		userID uuid.UUID,
	) (*dto.ImportExecuteResponse, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type importService struct {
	patientRepo      patientRepository.PatientRepository
	clinicalNoteRepo clinicalNoteRepository.ClinicalNoteRepository
	patientSvc       patientService.PatientService
	clinicalNoteSvc  clinicalNoteService.ClinicalNoteService
	encryptSvc       *crypto.EncryptionService
	db               *gorm.DB
	log              logger.Logger
}

func NewImportService(
	patientRepo patientRepository.PatientRepository,
	clinicalNoteRepo clinicalNoteRepository.ClinicalNoteRepository,
	patientSvc patientService.PatientService,
	clinicalNoteSvc clinicalNoteService.ClinicalNoteService,
	encryptSvc *crypto.EncryptionService,
	db *gorm.DB,
	log logger.Logger,
) ImportService {
	return &importService{
		patientRepo:      patientRepo,
		clinicalNoteRepo: clinicalNoteRepo,
		patientSvc:       patientSvc,
		clinicalNoteSvc:  clinicalNoteSvc,
		encryptSvc:       encryptSvc,
		db:               db,
		log:              log,
	}
}

func (s *importService) PreviewImport(
	ctx context.Context,
	req dto.ImportPreviewRequest,
	organizationID uuid.UUID,
	userID uuid.UUID,
) (*dto.ImportPreviewResponse, error) {
	// Decode base64 file data
	fileData, err := base64.StdEncoding.DecodeString(req.FileData)
	if err != nil {
		return nil, fmt.Errorf("invalid file data: %w", err)
	}

	switch req.Type {
	case "patients":
		return s.previewPatientsImport(ctx, fileData, req.FileName)
	case "notes":
		return s.previewNotesImport(ctx, fileData, req.FileName, organizationID)
	default:
		return nil, fmt.Errorf("unsupported import type: %s", req.Type)
	}
}

func (s *importService) ExecuteImport(
	ctx context.Context,
	req dto.ImportExecuteRequest,
	organizationID uuid.UUID,
	userID uuid.UUID,
) (*dto.ImportExecuteResponse, error) {
	// Decode base64 file data
	fileData, err := base64.StdEncoding.DecodeString(req.FileData)
	if err != nil {
		return nil, fmt.Errorf("invalid file data: %w", err)
	}

	switch req.Type {
	case "patients":
		return s.executePatientsImport(ctx, fileData, req.FileName, organizationID, userID)
	case "notes":
		return s.executeNotesImport(ctx, fileData, req.FileName, organizationID, userID)
	default:
		return nil, fmt.Errorf("unsupported import type: %s", req.Type)
	}
}

// Helper functions for CSV/XLSX parsing
func isXLSXFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return ext == ".xlsx" || ext == ".xls"
}

func parseCSV(fileData []byte) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(string(fileData)))
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	return reader.ReadAll()
}

func parseXLSX(fileData []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX file: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("no sheets found in XLSX file")
	}

	// Get all rows with cell values
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	// Get the actual cell values with proper formatting
	// Excel stores dates as numbers, but GetRows might return formatted strings
	// We need to check both the string value and the numeric value
	now := time.Now()
	minReasonableYear := now.Year() - 150 // Allow dates up to 150 years in the past
	maxReasonableYear := now.Year() + 100 // Allow dates up to 100 years in the future

	// Get the total number of rows and columns
	maxRow := len(rows)
	if maxRow == 0 {
		return rows, nil
	}
	maxCol := 0
	for _, row := range rows {
		if len(row) > maxCol {
			maxCol = len(row)
		}
	}

	// Convert Excel date numbers to properly formatted date strings
	for i := 0; i < maxRow; i++ {
		for j := 0; j < maxCol; j++ {
			cellName, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				continue
			}

			// Excel dates are stored as numbers, but GetRows returns formatted strings
			// We need to get the raw numeric value using GetCellFormula or check the cell type
			// First, try to get the cell value as a number
			cellValue, err := f.GetCellValue(sheetName, cellName)
			if err != nil {
				continue
			}

			// Try to parse as Excel date serial number (numeric value)
			if cellValue != "" {
				if num, err := strconv.ParseFloat(cellValue, 64); err == nil {
					// Excel dates are typically between 1 (1900-01-01) and ~50000 (2037+)
					if num > 1 && num < 100000 {
						// Convert Excel serial number to date
						excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
						days := int(num)
						date := excelEpoch.AddDate(0, 0, days)
						// Check if the resulting date is reasonable
						if date.Year() >= minReasonableYear && date.Year() <= maxReasonableYear {
							// Update the cell in rows array
							if i < len(rows) {
								if j >= len(rows[i]) {
									// Extend row if needed
									for len(rows[i]) <= j {
										rows[i] = append(rows[i], "")
									}
								}
								rows[i][j] = date.Format("2006-01-02")
							}
							continue
						}
					}
				}
			}

			// If it's a string (formatted date), try to parse it with our date parser
			// This handles cases where Excel returns formatted date strings like "01-15-90"
			if i < len(rows) && j < len(rows[i]) {
				cellStr := strings.TrimSpace(rows[i][j])
				if cellStr != "" {
					// Try to parse the string as a date
					if t, err := parseDate(cellStr); err == nil {
						// If it parses successfully, format it to standard YYYY-MM-DD
						rows[i][j] = t.Format("2006-01-02")
					} else {
						// If parseDate fails, it might be a 2-digit year format from Excel
						// Try to detect and convert 2-digit years to 4-digit years
						// Pattern: MM-DD-YY or MM/DD/YY
						if converted, ok := convertTwoDigitYear(cellStr); ok {
							if t, err := parseDate(converted); err == nil {
								rows[i][j] = t.Format("2006-01-02")
							}
						}
					}
				}
			}
		}
	}

	return rows, nil
}

func parseFile(fileData []byte, fileName string) ([][]string, error) {
	if isXLSXFile(fileName) {
		return parseXLSX(fileData)
	}
	return parseCSV(fileData)
}

// Patient Import Functions
func (s *importService) previewPatientsImport(
	ctx context.Context,
	fileData []byte,
	fileName string,
) (*dto.ImportPreviewResponse, error) {
	records, err := parseFile(fileData, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// First row is header
	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Validate required headers
	requiredHeaders := []string{"first_name", "last_name", "date_of_birth"}
	for _, req := range requiredHeaders {
		if _, ok := headerMap[req]; !ok {
			return nil, fmt.Errorf("missing required column: %s", req)
		}
	}

	var errors []dto.RowError
	var warnings []dto.RowWarning
	var preview []map[string]interface{}
	validCount := 0

	// Process rows (skip header)
	for i := 1; i < len(records); i++ {
		row := records[i]
		rowNum := i + 1

		// Build row map
		rowMap := make(map[string]interface{})
		for key, idx := range headerMap {
			if idx < len(row) {
				rowMap[key] = strings.TrimSpace(row[idx])
			} else {
				rowMap[key] = ""
			}
		}

		// Validate row
		rowValid := true
		firstName := getStringValue(rowMap, "first_name")
		lastName := getStringValue(rowMap, "last_name")
		dobStr := getStringValue(rowMap, "date_of_birth")

		if firstName == "" {
			errors = append(errors, dto.RowError{
				Row:     rowNum,
				Field:   "first_name",
				Message: "First name is required",
			})
			rowValid = false
		}

		if lastName == "" {
			errors = append(errors, dto.RowError{
				Row:     rowNum,
				Field:   "last_name",
				Message: "Last name is required",
			})
			rowValid = false
		}

		if dobStr == "" {
			errors = append(errors, dto.RowError{
				Row:     rowNum,
				Field:   "date_of_birth",
				Message: "Date of birth is required",
			})
			rowValid = false
		} else {
			_, err := parseDate(dobStr)
			if err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Field:   "date_of_birth",
					Message: fmt.Sprintf("Invalid date format. Supported formats: YYYY-MM-DD, MM/DD/YYYY, M/D/YYYY, MM-DD-YYYY, M-D-YYYY. Got: %s", dobStr),
				})
				rowValid = false
			}
		}

		// Validate email if provided
		email := getStringValue(rowMap, "email")
		if email != "" && !strings.Contains(email, "@") {
			warnings = append(warnings, dto.RowWarning{
				Row:     rowNum,
				Field:   "email",
				Message: "Email format appears invalid",
			})
		}

		if rowValid {
			validCount++
			if len(preview) < 10 {
				preview = append(preview, rowMap)
			}
		}
	}

	return &dto.ImportPreviewResponse{
		TotalRows:   len(records) - 1,
		ValidRows:   validCount,
		InvalidRows: len(records) - 1 - validCount,
		Preview:     preview,
		Errors:      errors,
		Warnings:    warnings,
	}, nil
}

func (s *importService) executePatientsImport(
	ctx context.Context,
	fileData []byte,
	fileName string,
	organizationID uuid.UUID,
	userID uuid.UUID,
) (*dto.ImportExecuteResponse, error) {
	records, err := parseFile(fileData, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var errors []dto.RowError
	var importedIDs []uuid.UUID
	successCount := 0

	// Use transaction for atomicity
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for i := 1; i < len(records); i++ {
			row := records[i]
			rowNum := i + 1

			rowMap := make(map[string]interface{})
			for key, idx := range headerMap {
				if idx < len(row) {
					rowMap[key] = strings.TrimSpace(row[idx])
				} else {
					rowMap[key] = ""
				}
			}

			firstName := getStringValue(rowMap, "first_name")
			lastName := getStringValue(rowMap, "last_name")
			dobStr := getStringValue(rowMap, "date_of_birth")

			// Validate required fields
			if firstName == "" || lastName == "" || dobStr == "" {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Message: "Missing required fields: first_name, last_name, or date_of_birth",
				})
				continue
			}

			dob, err := parseDate(dobStr)
			if err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Field:   "date_of_birth",
					Message: fmt.Sprintf("Invalid date format: %s. Supported formats: YYYY-MM-DD, MM/DD/YYYY, M/D/YYYY, MM-DD-YYYY, M-D-YYYY", dobStr),
				})
				continue
			}

			// Create patient
			email := getStringValue(rowMap, "email")
			phone := getStringValue(rowMap, "phone")
			address := getStringValue(rowMap, "address")

			var emailPtr *string
			if email != "" {
				emailPtr = &email
			}
			var phonePtr *string
			if phone != "" {
				phonePtr = &phone
			}
			var addressPtr *string
			if address != "" {
				addressPtr = &address
			}

			patient := &patientEntity.Patient{
				ID:             uuid.New(),
				OrganizationID: organizationID,
				FirstName:      firstName,
				LastName:       lastName,
				DateOfBirth:    dob,
				Email:          emailPtr,
				Phone:          phonePtr,
				Address:        addressPtr,
				Status:         "active",
				CreatedBy:      userID,
			}

			if err := tx.Create(patient).Error; err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Message: fmt.Sprintf("Failed to create patient: %v", err),
				})
				continue
			}

			importedIDs = append(importedIDs, patient.ID)
			successCount++
		}

		// If too many errors, rollback
		if len(errors) > len(records)/2 {
			return fmt.Errorf("too many errors, rolling back transaction")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.ImportExecuteResponse{
		TotalRows:    len(records) - 1,
		SuccessCount: successCount,
		ErrorCount:   len(errors),
		Errors:       errors,
		ImportedIDs:  importedIDs,
	}, nil
}

// Clinical Note Import Functions
func (s *importService) previewNotesImport(
	ctx context.Context,
	fileData []byte,
	fileName string,
	organizationID uuid.UUID,
) (*dto.ImportPreviewResponse, error) {
	records, err := parseFile(fileData, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	// First row is header
	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Validate required headers
	requiredHeaders := []string{"patient_id", "note_type"}
	for _, req := range requiredHeaders {
		if _, ok := headerMap[req]; !ok {
			return nil, fmt.Errorf("missing required column: %s", req)
		}
	}

	var errors []dto.RowError
	var warnings []dto.RowWarning
	var preview []map[string]interface{}
	validCount := 0

	// Process rows (skip header)
	for i := 1; i < len(records); i++ {
		row := records[i]
		rowNum := i + 1

		// Build row map
		rowMap := make(map[string]interface{})
		for key, idx := range headerMap {
			if idx < len(row) {
				rowMap[key] = strings.TrimSpace(row[idx])
			} else {
				rowMap[key] = ""
			}
		}

		rowValid := true
		patientIDStr := getStringValue(rowMap, "patient_id")
		noteType := getStringValue(rowMap, "note_type")

		// Validate patient_id
		if patientIDStr == "" {
			errors = append(errors, dto.RowError{
				Row:     rowNum,
				Field:   "patient_id",
				Message: "Patient ID is required",
			})
			rowValid = false
		} else {
			if _, err := uuid.Parse(patientIDStr); err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Field:   "patient_id",
					Message: "Invalid UUID format",
				})
				rowValid = false
			} else {
				_, err := s.patientRepo.FindByID(uuid.MustParse(patientIDStr))
				if err != nil {
					warnings = append(warnings, dto.RowWarning{
						Row:     rowNum,
						Field:   "patient_id",
						Message: "Patient not found in database",
					})
				}
			}
		}

		// Validate note_type
		if noteType == "" {
			errors = append(errors, dto.RowError{
				Row:     rowNum,
				Field:   "note_type",
				Message: "Note type is required",
			})
			rowValid = false
		}

		if rowValid {
			validCount++
			if len(preview) < 10 {
				preview = append(preview, rowMap)
			}
		}
	}

	return &dto.ImportPreviewResponse{
		TotalRows:   len(records) - 1,
		ValidRows:   validCount,
		InvalidRows: len(records) - 1 - validCount,
		Preview:     preview,
		Errors:      errors,
		Warnings:    warnings,
	}, nil
}

func (s *importService) executeNotesImport(
	ctx context.Context,
	fileData []byte,
	fileName string,
	organizationID uuid.UUID,
	userID uuid.UUID,
) (*dto.ImportExecuteResponse, error) {
	records, err := parseFile(fileData, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var errors []dto.RowError
	var importedIDs []uuid.UUID
	successCount := 0

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for i := 1; i < len(records); i++ {
			row := records[i]
			rowNum := i + 1

			rowMap := make(map[string]interface{})
			for key, idx := range headerMap {
				if idx < len(row) {
					rowMap[key] = strings.TrimSpace(row[idx])
				} else {
					rowMap[key] = ""
				}
			}

			patientIDStr := getStringValue(rowMap, "patient_id")
			noteType := getStringValue(rowMap, "note_type")

			// Validate patient_id
			patientID, err := uuid.Parse(patientIDStr)
			if err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Field:   "patient_id",
					Message: "Invalid patient ID format",
				})
				continue
			}

			// Verify patient exists
			_, err = s.patientRepo.FindByID(patientID)
			if err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Field:   "patient_id",
					Message: "Patient not found",
				})
				continue
			}

			// Validate note_type
			if noteType == "" {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Field:   "note_type",
					Message: "Note type is required",
				})
				continue
			}

			// Parse appointment_id if provided
			var appointmentID *uuid.UUID
			appointmentIDStr := getStringValue(rowMap, "appointment_id")
			if appointmentIDStr != "" {
				aptID, err := uuid.Parse(appointmentIDStr)
				if err != nil {
					errors = append(errors, dto.RowError{
						Row:     rowNum,
						Field:   "appointment_id",
						Message: "Invalid appointment ID format",
					})
					continue
				}
				appointmentID = &aptID
			}

			// Parse is_signed
			isSignedStr := strings.ToLower(getStringValue(rowMap, "is_signed"))
			isSigned := isSignedStr == "true" || isSignedStr == "1" || isSignedStr == "yes"

			var signedAt *time.Time
			if isSigned {
				now := time.Now()
				signedAt = &now
			}

			icd10Code := getStringValue(rowMap, "icd10_code")
			subjective := getStringValue(rowMap, "subjective")
			objective := getStringValue(rowMap, "objective")
			assessment := getStringValue(rowMap, "assessment")
			plan := getStringValue(rowMap, "plan")

			// Create clinical note entity
			clinicalNote := &clinicalNoteEntity.ClinicalNote{
				ID:             uuid.New(),
				OrganizationID: organizationID,
				PatientID:      patientID,
				ClinicianID:    userID,
				AppointmentID:  appointmentID,
				NoteType:       noteType,
				ICD10Code:      icd10Code,
				Subjective:     &subjective,
				Objective:      &objective,
				Assessment:     &assessment,
				Plan:           &plan,
				IsSigned:       isSigned,
				SignedAt:       signedAt,
			}

			// Encrypt the note content
			content := clinicalNoteContent{
				Subjective: clinicalNote.Subjective,
				Objective:  clinicalNote.Objective,
				Assessment: clinicalNote.Assessment,
				Plan:       clinicalNote.Plan,
			}

			jsonData, err := json.Marshal(content)
			if err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Message: fmt.Sprintf("Failed to marshal note content: %v", err),
				})
				continue
			}

			encryptedBase64, err := s.encryptSvc.Encrypt(string(jsonData))
			if err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Message: fmt.Sprintf("Failed to encrypt note: %v", err),
				})
				continue
			}

			encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedBase64)
			if err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Message: fmt.Sprintf("Failed to decode encrypted data: %v", err),
				})
				continue
			}

			const nonceSize = 12
			if len(encryptedBytes) < nonceSize {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Message: "Encrypted data too short",
				})
				continue
			}

			clinicalNote.ContentEncrypted = encryptedBytes
			clinicalNote.Nonce = encryptedBytes[:nonceSize]
			clinicalNote.KeyID = "v1"

			if err := tx.Create(clinicalNote).Error; err != nil {
				errors = append(errors, dto.RowError{
					Row:     rowNum,
					Message: fmt.Sprintf("Failed to create note: %v", err),
				})
				continue
			}

			importedIDs = append(importedIDs, clinicalNote.ID)
			successCount++
		}

		if len(errors) > (len(records)-1)/2 {
			return fmt.Errorf("too many errors, rolling back transaction")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.ImportExecuteResponse{
		TotalRows:    len(records) - 1,
		SuccessCount: successCount,
		ErrorCount:   len(errors),
		Errors:       errors,
		ImportedIDs:  importedIDs,
	}, nil
}

func (s *importService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.patientSvc.GetOrganizationID(ctx, userID)
}

// Helper functions
func getStringValue(m map[string]interface{}, key string) string {
	val, ok := m[key]
	if !ok {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(str)
}

// parseDate attempts to parse a date string in multiple common formats
// and returns a standardized time.Time. Only supports 4-digit years (YYYY) for clarity.
// Supported formats:
// - YYYY-MM-DD (2006-01-02) - standard ISO format
// - MM/DD/YYYY (01/15/1990) - US format with slashes
// - M/D/YYYY (1/15/1990) - US format without leading zeros
// - MM-DD-YYYY (01-15-1990) - US format with dashes
// - M-D-YYYY (1-15-1990) - US format with dashes, no leading zeros
func parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// List of date formats to try (only 4-digit years)
	formats := []string{
		"2006-01-02", // YYYY-MM-DD (standard ISO format)
		"01/02/2006", // MM/DD/YYYY
		"1/2/2006",   // M/D/YYYY
		"01-02-2006", // MM-DD-YYYY
		"1-2-2006",   // M-D-YYYY
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s. Supported formats: YYYY-MM-DD, MM/DD/YYYY, M/D/YYYY, MM-DD-YYYY, M-D-YYYY", dateStr)
}

// convertTwoDigitYear attempts to convert a date string with 2-digit year to 4-digit year
// This is a fallback for Excel files that export dates with 2-digit years
// Converts: MM-DD-YY -> MM-DD-YYYY, MM/DD/YY -> MM/DD/YYYY
func convertTwoDigitYear(dateStr string) (string, bool) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return "", false
	}

	// Check for MM-DD-YY pattern
	if strings.Contains(dateStr, "-") {
		parts := strings.Split(dateStr, "-")
		if len(parts) == 3 {
			yearStr := strings.TrimSpace(parts[2])
			if len(yearStr) == 2 {
				year, err := strconv.Atoi(yearStr)
				if err == nil {
					if year < 50 {
						year += 2000
					} else {
						year += 1900
					}
					return fmt.Sprintf("%s-%s-%04d", parts[0], parts[1], year), true
				}
			}
		}
	}

	// Check for MM/DD/YY pattern
	if strings.Contains(dateStr, "/") {
		parts := strings.Split(dateStr, "/")
		if len(parts) == 3 {
			yearStr := strings.TrimSpace(parts[2])
			if len(yearStr) == 2 {
				year, err := strconv.Atoi(yearStr)
				if err == nil {
					if year < 50 {
						year += 2000
					} else {
						year += 1900
					}
					return fmt.Sprintf("%s/%s/%04d", parts[0], parts[1], year), true
				}
			}
		}
	}

	return "", false
}

// clinicalNoteContent matches the structure used in clinical_note service
type clinicalNoteContent struct {
	Subjective *string `json:"subjective,omitempty"`
	Objective  *string `json:"objective,omitempty"`
	Assessment *string `json:"assessment,omitempty"`
	Plan       *string `json:"plan,omitempty"`
}
