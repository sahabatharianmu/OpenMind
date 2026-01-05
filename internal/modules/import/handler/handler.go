package handler

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/import/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/import/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"github.com/xuri/excelize/v2"
)

type ImportHandler struct {
	svc service.ImportService
}

func NewImportHandler(svc service.ImportService) *ImportHandler {
	return &ImportHandler{svc: svc}
}

func (h *ImportHandler) PreviewImport(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Get organization ID from user
	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	var req dto.ImportPreviewRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	result, err := h.svc.PreviewImport(context.Background(), req, orgID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Preview generated successfully", result))
}

func (h *ImportHandler) ExecuteImport(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Get organization ID from user
	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	var req dto.ImportExecuteRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	result, err := h.svc.ExecuteImport(context.Background(), req, orgID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Import completed", result))
}

func (h *ImportHandler) DownloadTemplate(_ context.Context, c *app.RequestContext) {
	importType := c.Param("type")
	format := c.Query("format") // "csv" or "xlsx"
	if importType == "" {
		response.BadRequest(c, "Import type is required", nil)
		return
	}

	if format == "" {
		format = "csv" // default to CSV
	}

	var content []byte
	var filename string
	var contentType string

	switch importType {
	case "patients":
		if format == "xlsx" {
			// Create XLSX file for patients
			f := excelize.NewFile()
			defer f.Close()

			sheetName := "Patients"
			f.NewSheet(sheetName)
			f.DeleteSheet("Sheet1")

			// Headers
			headers := []string{"first_name", "last_name", "date_of_birth", "email", "phone", "address"}
			for i, h := range headers {
				cell := fmt.Sprintf("%c1", 'A'+i)
				f.SetCellValue(sheetName, cell, h)
			}

			// Example rows
			examples := [][]interface{}{
				{"John", "Doe", "1990-01-15", "john.doe@example.com", "555-0100", "123 Main St"},
				{"Jane", "Smith", "1985-05-20", "jane.smith@example.com", "555-0101", "456 Oak Ave"},
			}
			for rowIdx, row := range examples {
				for colIdx, val := range row {
					cell := fmt.Sprintf("%c%d", 'A'+colIdx, rowIdx+2)
					f.SetCellValue(sheetName, cell, val)
				}
			}

			var buf bytes.Buffer
			if err := f.Write(&buf); err != nil {
				response.InternalServerError(c, "Failed to generate XLSX template")
				return
			}
			content = buf.Bytes()
			filename = "patients-import-template.xlsx"
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		} else {
			// CSV template for patients
			csvContent := "first_name,last_name,date_of_birth,email,phone,address\n" +
				"John,Doe,1990-01-15,john.doe@example.com,555-0100,123 Main St\n" +
				"Jane,Smith,1985-05-20,jane.smith@example.com,555-0101,456 Oak Ave"
			content = []byte(csvContent)
			filename = "patients-import-template.csv"
			contentType = "text/csv"
		}
	case "notes":
		if format == "xlsx" {
			// Create XLSX file for clinical notes
			f := excelize.NewFile()
			defer f.Close()

			sheetName := "Clinical Notes"
			f.NewSheet(sheetName)
			f.DeleteSheet("Sheet1")

			// Headers
			headers := []string{
				"patient_id",
				"appointment_id",
				"note_type",
				"icd10_code",
				"subjective",
				"objective",
				"assessment",
				"plan",
				"is_signed",
			}
			for i, h := range headers {
				cell := fmt.Sprintf("%c1", 'A'+i)
				f.SetCellValue(sheetName, cell, h)
			}

			// Example row
			example := []interface{}{
				"patient-uuid-here",
				"appointment-uuid-here",
				"soap",
				"E11.9",
				"Patient reports...",
				"Physical examination reveals...",
				"Assessment and diagnosis...",
				"Treatment plan includes...",
				"false",
			}
			for colIdx, val := range example {
				cell := fmt.Sprintf("%c2", 'A'+colIdx)
				f.SetCellValue(sheetName, cell, val)
			}

			var buf bytes.Buffer
			if err := f.Write(&buf); err != nil {
				response.InternalServerError(c, "Failed to generate XLSX template")
				return
			}
			content = buf.Bytes()
			filename = "clinical-notes-import-template.xlsx"
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		} else {
			// CSV template for clinical notes
			csvContent := "patient_id,appointment_id,note_type,icd10_code,subjective,objective,assessment,plan,is_signed\n" +
				"patient-uuid-here,appointment-uuid-here,soap,E11.9,Patient reports...,Physical examination reveals...,Assessment and diagnosis...,Treatment plan includes...,false"
			content = []byte(csvContent)
			filename = "clinical-notes-import-template.csv"
			contentType = "text/csv"
		}
	default:
		response.BadRequest(c, "Invalid import type. Must be 'patients' or 'notes'", nil)
		return
	}

	c.Response.Header.Set("Content-Type", contentType)
	c.Response.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response.SetStatusCode(consts.StatusOK)
	c.Response.SetBody(content)
}
