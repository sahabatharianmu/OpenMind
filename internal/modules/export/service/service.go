package service

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	appointmentRepo "github.com/sahabatharianmu/OpenMind/internal/modules/appointment/repository"
	clinicalNoteRepo "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/repository"
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
	orgRepo          organizationRepo.OrganizationRepository
	patientRepo      patientRepo.PatientRepository
	appointmentRepo  appointmentRepo.AppointmentRepository
	clinicalNoteRepo clinicalNoteRepo.ClinicalNoteRepository
	invoiceRepo      invoiceRepo.InvoiceRepository
	log              logger.Logger
}

func NewExportService(
	orgRepo organizationRepo.OrganizationRepository,
	patientRepo patientRepo.PatientRepository,
	appointmentRepo appointmentRepo.AppointmentRepository,
	clinicalNoteRepo clinicalNoteRepo.ClinicalNoteRepository,
	invoiceRepo invoiceRepo.InvoiceRepository,
	log logger.Logger,
) ExportService {
	return &exportService{
		orgRepo:          orgRepo,
		patientRepo:      patientRepo,
		appointmentRepo:  appointmentRepo,
		clinicalNoteRepo: clinicalNoteRepo,
		invoiceRepo:      invoiceRepo,
		log:              log,
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

	// Export patients
	patients, _, err := s.patientRepo.List(org.ID, 1, 10000) // Large page size for export
	if err != nil {
		s.log.Error("Failed to fetch patients for export", zap.Error(err))
	} else {
		data, _ := json.MarshalIndent(patients, "", "  ")
		files["patients.json"] = data
	}

	// Export appointments
	appointments, _, err := s.appointmentRepo.List(org.ID, 1, 10000)
	if err != nil {
		s.log.Error("Failed to fetch appointments for export", zap.Error(err))
	} else {
		data, _ := json.MarshalIndent(appointments, "", "  ")
		files["appointments.json"] = data
	}

	// Export clinical notes
	notes, _, err := s.clinicalNoteRepo.List(org.ID, 1, 10000)
	if err != nil {
		s.log.Error("Failed to fetch clinical notes for export", zap.Error(err))
	} else {
		data, _ := json.MarshalIndent(notes, "", "  ")
		files["clinical_notes.json"] = data
	}

	// Export invoices
	invoices, _, err := s.invoiceRepo.List(org.ID, 1, 10000)
	if err != nil {
		s.log.Error("Failed to fetch invoices for export", zap.Error(err))
	} else {
		data, _ := json.MarshalIndent(invoices, "", "  ")
		files["invoices.json"] = data
	}

	// Export organization info
	orgData := map[string]interface{}{
		"id":   org.ID,
		"name": org.Name,
		"type": org.Type,
	}
	data, _ := json.MarshalIndent(orgData, "", "  ")
	files["organization.json"] = data

	s.log.Info("Data export completed", zap.String("org_id", org.ID.String()), zap.Int("file_count", len(files)))

	if len(files) == 0 {
		return nil, fmt.Errorf("no data to export")
	}

	return files, nil
}
