package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	appointmentRepo "github.com/sahabatharianmu/OpenMind/internal/modules/appointment/repository"
	auditLogService "github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/service"
	clinicalNoteService "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/service"
	invoiceRepo "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/repository"
	organizationRepo "github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	patientRepo "github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type ExportService interface {
	ExportAllData(userID uuid.UUID) (map[string][]byte, error)
}

type exportService struct {
	orgRepo         organizationRepo.OrganizationRepository
	patientRepo     patientRepo.PatientRepository
	appointmentRepo appointmentRepo.AppointmentRepository
	clinicalNoteSvc clinicalNoteService.ClinicalNoteService
	invoiceRepo     invoiceRepo.InvoiceRepository
	auditLogSvc     auditLogService.AuditLogService
	log             logger.Logger
}

func NewExportService(
	orgRepo organizationRepo.OrganizationRepository,
	patientRepo patientRepo.PatientRepository,
	appointmentRepo appointmentRepo.AppointmentRepository,
	clinicalNoteSvc clinicalNoteService.ClinicalNoteService,
	invoiceRepo invoiceRepo.InvoiceRepository,
	auditLogSvc auditLogService.AuditLogService,
	log logger.Logger,
) ExportService {
	return &exportService{
		orgRepo:         orgRepo,
		patientRepo:     patientRepo,
		appointmentRepo: appointmentRepo,
		clinicalNoteSvc: clinicalNoteSvc,
		invoiceRepo:     invoiceRepo,
		auditLogSvc:     auditLogSvc,
		log:             log,
	}
}

func (s *exportService) ExportAllData(userID uuid.UUID) (map[string][]byte, error) {
	// Get user's organization
	org, err := s.orgRepo.GetByUserID(userID)
	if err != nil {
		s.log.Error("ExportAllData: failed to get organization", zap.Error(err))
		return nil, response.ErrNotFound
	}

	files := make(map[string][]byte)

	// Export patients - use large limit with offset 0 to get all records
	// nil for assignedPatientIDs means no filter (admin export sees all)
	patients, _, err := s.patientRepo.List(org.ID, 10000, 0, nil)
	if err != nil {
		s.log.Error("Failed to fetch patients for export", zap.Error(err))
	} else {
		data, _ := json.MarshalIndent(patients, "", "  ")
		files["patients.json"] = data
	}

	// Export appointments - use large limit with offset 0 to get all records
	appointments, _, err := s.appointmentRepo.List(org.ID, 10000, 0)
	if err != nil {
		s.log.Error("Failed to fetch appointments for export", zap.Error(err))
	} else {
		data, _ := json.MarshalIndent(appointments, "", "  ")
		files["appointments.json"] = data
	}

	// Export clinical notes
	notes, _, err := s.clinicalNoteSvc.List(context.Background(), org.ID, 1, 10000)
	if err != nil {
		s.log.Error("Failed to fetch clinical notes for export", zap.Error(err))
	} else {
		data, _ := json.MarshalIndent(notes, "", "  ")
		files["clinical_notes.json"] = data

		// Also export raw attachment files (decrypted)
		for _, note := range notes {
			for _, att := range note.Attachments {
				_, data, _, err := s.clinicalNoteSvc.DownloadAttachment(context.Background(), att.ID, org.ID)
				if err == nil {
					files[fmt.Sprintf("attachments/%s_%s", att.ID.String()[:8], att.FileName)] = data
				}
			}
		}
	}

	// Export invoices - use large limit with offset 0 to get all records
	invoices, _, err := s.invoiceRepo.List(org.ID, 10000, 0)
	if err != nil {
		s.log.Error("Failed to fetch invoices for export", zap.Error(err))
	} else {
		data, _ := json.MarshalIndent(invoices, "", "  ")
		files["invoices.json"] = data
	}

	// Export organization info - include all fields
	orgData := map[string]interface{}{
		"id":       org.ID,
		"name":     org.Name,
		"type":     org.Type,
		"tax_id":   org.TaxID,
		"npi":      org.NPI,
		"address":  org.Address,
		"currency": org.Currency,
		"locale":   org.Locale,
	}
	data, _ := json.MarshalIndent(orgData, "", "  ")
	files["organization.json"] = data

	// Export audit logs
	logs, _, err := s.auditLogSvc.List(context.Background(), org.ID, 1, 10000, nil)
	if err != nil {
		s.log.Error("Failed to fetch audit logs for export", zap.Error(err))
	} else {
		data, _ = json.MarshalIndent(logs, "", "  ")
		files["audit_logs.json"] = data
	}

	s.log.Info("Data export completed", zap.String("org_id", org.ID.String()), zap.Int("file_count", len(files)))

	if len(files) == 0 {
		return nil, fmt.Errorf("no data to export")
	}

	return files, nil
}
