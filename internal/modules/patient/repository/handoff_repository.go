package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PatientHandoffRepository interface {
	Create(handoff *entity.PatientHandoff) error
	GetByID(id uuid.UUID) (*entity.PatientHandoff, error)
	GetByPatientID(patientID uuid.UUID) ([]entity.PatientHandoff, error)
	GetPendingByReceivingClinician(clinicianID uuid.UUID) ([]entity.PatientHandoff, error)
	GetPendingByRequestingClinician(clinicianID uuid.UUID) ([]entity.PatientHandoff, error)
	Update(handoff *entity.PatientHandoff) error
	GetByStatus(status string, limit, offset int) ([]entity.PatientHandoff, error)
	GetPendingByPatientAndClinician(patientID, clinicianID uuid.UUID) (*entity.PatientHandoff, error)
	HasPendingHandoffAsReceivingClinician(patientID, clinicianID uuid.UUID) (bool, error)
}

type patientHandoffRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewPatientHandoffRepository(db *gorm.DB, log logger.Logger) PatientHandoffRepository {
	return &patientHandoffRepository{
		db:  db,
		log: log,
	}
}

func (r *patientHandoffRepository) Create(handoff *entity.PatientHandoff) error {
	if err := r.db.Create(handoff).Error; err != nil {
		r.log.Error("Failed to create patient handoff", zap.Error(err),
			zap.String("patient_id", handoff.PatientID.String()),
			zap.String("requesting_clinician_id", handoff.RequestingClinicianID.String()))
		return err
	}
	return nil
}

func (r *patientHandoffRepository) GetByID(id uuid.UUID) (*entity.PatientHandoff, error) {
	var handoff entity.PatientHandoff
	if err := r.db.First(&handoff, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Error("Failed to get patient handoff by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, err
	}
	return &handoff, nil
}

func (r *patientHandoffRepository) GetByPatientID(patientID uuid.UUID) ([]entity.PatientHandoff, error) {
	var handoffs []entity.PatientHandoff
	if err := r.db.Where("patient_id = ?", patientID).
		Order("requested_at DESC").
		Find(&handoffs).Error; err != nil {
		r.log.Error("Failed to get patient handoffs by patient ID", zap.Error(err),
			zap.String("patient_id", patientID.String()))
		return nil, err
	}
	return handoffs, nil
}

func (r *patientHandoffRepository) GetPendingByReceivingClinician(clinicianID uuid.UUID) ([]entity.PatientHandoff, error) {
	var handoffs []entity.PatientHandoff
	if err := r.db.Where("receiving_clinician_id = ? AND status = ?", clinicianID, entity.StatusRequested).
		Order("requested_at DESC").
		Find(&handoffs).Error; err != nil {
		r.log.Error("Failed to get pending handoffs for receiving clinician", zap.Error(err),
			zap.String("clinician_id", clinicianID.String()))
		return nil, err
	}
	return handoffs, nil
}

func (r *patientHandoffRepository) GetPendingByRequestingClinician(clinicianID uuid.UUID) ([]entity.PatientHandoff, error) {
	var handoffs []entity.PatientHandoff
	if err := r.db.Where("requesting_clinician_id = ? AND status = ?", clinicianID, entity.StatusRequested).
		Order("requested_at DESC").
		Find(&handoffs).Error; err != nil {
		r.log.Error("Failed to get pending handoffs for requesting clinician", zap.Error(err),
			zap.String("clinician_id", clinicianID.String()))
		return nil, err
	}
	return handoffs, nil
}

func (r *patientHandoffRepository) Update(handoff *entity.PatientHandoff) error {
	if err := r.db.Save(handoff).Error; err != nil {
		r.log.Error("Failed to update patient handoff", zap.Error(err),
			zap.String("id", handoff.ID.String()))
		return err
	}
	return nil
}

func (r *patientHandoffRepository) GetByStatus(status string, limit, offset int) ([]entity.PatientHandoff, error) {
	var handoffs []entity.PatientHandoff
	query := r.db.Where("status = ?", status).Order("requested_at DESC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	if err := query.Find(&handoffs).Error; err != nil {
		r.log.Error("Failed to get patient handoffs by status", zap.Error(err), zap.String("status", status))
		return nil, err
	}
	return handoffs, nil
}

func (r *patientHandoffRepository) GetPendingByPatientAndClinician(patientID, clinicianID uuid.UUID) (*entity.PatientHandoff, error) {
	var handoff entity.PatientHandoff
	if err := r.db.Where("patient_id = ? AND (requesting_clinician_id = ? OR receiving_clinician_id = ?) AND status = ?",
		patientID, clinicianID, clinicianID, entity.StatusRequested).
		First(&handoff).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Error("Failed to get pending handoff by patient and clinician", zap.Error(err),
			zap.String("patient_id", patientID.String()),
			zap.String("clinician_id", clinicianID.String()))
		return nil, err
	}
	return &handoff, nil
}

func (r *patientHandoffRepository) HasPendingHandoffAsReceivingClinician(patientID, clinicianID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.PatientHandoff{}).
		Where("patient_id = ? AND receiving_clinician_id = ? AND status = ?",
			patientID, clinicianID, entity.StatusRequested).
		Count(&count).Error; err != nil {
		r.log.Error("Failed to check pending handoff as receiving clinician", zap.Error(err),
			zap.String("patient_id", patientID.String()),
			zap.String("clinician_id", clinicianID.String()))
		return false, err
	}
	return count > 0, nil
}

