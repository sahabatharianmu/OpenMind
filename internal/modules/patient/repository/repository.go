package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PatientRepository interface {
	Create(patient *entity.Patient) error
	Update(patient *entity.Patient) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*entity.Patient, error)
	List(organizationID uuid.UUID, limit, offset int, assignedPatientIDs []uuid.UUID) ([]entity.Patient, int64, error)
	GetOrganizationID(userID uuid.UUID) (uuid.UUID, error)
	AssignClinician(patientID, clinicianID uuid.UUID, role string, assignedBy uuid.UUID) error
	UnassignClinician(patientID, clinicianID uuid.UUID) error
	GetAssignedClinicians(patientID uuid.UUID) ([]entity.PatientAssignment, error)
	GetAssignedPatients(clinicianID, organizationID uuid.UUID) ([]uuid.UUID, error)
	IsPatientAssignedToClinician(patientID, clinicianID uuid.UUID) (bool, error)
	CountPrimaryClinicians(patientID uuid.UUID) (int64, error)
}

type patientRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewPatientRepository(db *gorm.DB, log logger.Logger) PatientRepository {
	return &patientRepository{
		db:  db,
		log: log,
	}
}

func (r *patientRepository) Create(patient *entity.Patient) error {
	if err := r.db.Create(patient).Error; err != nil {
		r.log.Error("Failed to create patient", zap.Error(err))
		return err
	}
	return nil
}

func (r *patientRepository) Update(patient *entity.Patient) error {
	if err := r.db.Save(patient).Error; err != nil {
		r.log.Error("Failed to update patient", zap.Error(err), zap.String("id", patient.ID.String()))
		return err
	}
	return nil
}

func (r *patientRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entity.Patient{}, "id = ?", id).Error; err != nil {
		r.log.Error("Failed to delete patient", zap.Error(err), zap.String("id", id.String()))
		return err
	}
	return nil
}

func (r *patientRepository) FindByID(id uuid.UUID) (*entity.Patient, error) {
	var patient entity.Patient
	if err := r.db.First(&patient, "id = ?", id).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find patient", zap.Error(err), zap.String("id", id.String()))
		}
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepository) List(
	organizationID uuid.UUID,
	limit, offset int,
	assignedPatientIDs []uuid.UUID,
) ([]entity.Patient, int64, error) {
	var patients []entity.Patient
	var total int64

	query := r.db.Model(&entity.Patient{}).Where("organization_id = ?", organizationID)

	// Filter by assigned patients if provided (for non-admin users)
	if len(assignedPatientIDs) > 0 {
		query = query.Where("id IN ?", assignedPatientIDs)
	}

	if err := query.Count(&total).Error; err != nil {
		r.log.Error("Failed to count patients", zap.Error(err))
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&patients).Error; err != nil {
		r.log.Error("Failed to list patients", zap.Error(err))
		return nil, 0, err
	}

	return patients, total, nil
}

func (r *patientRepository) GetOrganizationID(userID uuid.UUID) (uuid.UUID, error) {
	var orgIDStr string
	if err := r.db.Table("organization_members").Select("organization_id").Where("user_id = ?", userID).Limit(1).Scan(&orgIDStr).Error; err != nil {
		r.log.Error("Failed to get organization ID", zap.Error(err), zap.String("user_id", userID.String()))
		return uuid.Nil, err
	}
	if orgIDStr == "" {
		return uuid.Nil, errors.New("organization not found for user")
	}
	return uuid.Parse(orgIDStr)
}

func (r *patientRepository) AssignClinician(patientID, clinicianID uuid.UUID, role string, assignedBy uuid.UUID) error {
	// First verify the patient exists in the current schema context
	var patientExists bool
	if err := r.db.Raw("SELECT EXISTS(SELECT 1 FROM patients WHERE id = ?)", patientID).Scan(&patientExists).Error; err != nil {
		r.log.Error("Failed to check if patient exists", zap.Error(err),
			zap.String("patient_id", patientID.String()))
		return err
	}
	if !patientExists {
		r.log.Error("Patient not found in current schema", zap.String("patient_id", patientID.String()))
		return fmt.Errorf("patient not found: %s", patientID.String())
	}

	assignment := &entity.PatientAssignment{
		PatientID:   patientID,
		ClinicianID: clinicianID,
		Role:        role,
		AssignedBy:  assignedBy,
	}

	if err := r.db.Create(assignment).Error; err != nil {
		r.log.Error("Failed to assign clinician to patient", zap.Error(err),
			zap.String("patient_id", patientID.String()),
			zap.String("clinician_id", clinicianID.String()))
		return err
	}
	return nil
}

func (r *patientRepository) UnassignClinician(patientID, clinicianID uuid.UUID) error {
	if err := r.db.Where("patient_id = ? AND clinician_id = ?", patientID, clinicianID).
		Delete(&entity.PatientAssignment{}).Error; err != nil {
		r.log.Error("Failed to unassign clinician from patient", zap.Error(err),
			zap.String("patient_id", patientID.String()),
			zap.String("clinician_id", clinicianID.String()))
		return err
	}
	return nil
}

func (r *patientRepository) GetAssignedClinicians(patientID uuid.UUID) ([]entity.PatientAssignment, error) {
	var assignments []entity.PatientAssignment
	if err := r.db.Where("patient_id = ?", patientID).Find(&assignments).Error; err != nil {
		r.log.Error("Failed to get assigned clinicians", zap.Error(err),
			zap.String("patient_id", patientID.String()))
		return nil, err
	}
	return assignments, nil
}

func (r *patientRepository) GetAssignedPatients(clinicianID, organizationID uuid.UUID) ([]uuid.UUID, error) {
	var patientIDs []uuid.UUID

	// Join with patients table to ensure we only get patients from the same organization
	if err := r.db.Table("assigned_clinicians").
		Select("assigned_clinicians.patient_id").
		Joins("INNER JOIN patients ON assigned_clinicians.patient_id = patients.id").
		Where("assigned_clinicians.clinician_id = ? AND patients.organization_id = ?", clinicianID, organizationID).
		Pluck("assigned_clinicians.patient_id", &patientIDs).Error; err != nil {
		r.log.Error("Failed to get assigned patients", zap.Error(err),
			zap.String("clinician_id", clinicianID.String()),
			zap.String("organization_id", organizationID.String()))
		return nil, err
	}

	return patientIDs, nil
}

func (r *patientRepository) IsPatientAssignedToClinician(patientID, clinicianID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.PatientAssignment{}).
		Where("patient_id = ? AND clinician_id = ?", patientID, clinicianID).
		Count(&count).Error; err != nil {
		r.log.Error("Failed to check patient assignment", zap.Error(err),
			zap.String("patient_id", patientID.String()),
			zap.String("clinician_id", clinicianID.String()))
		return false, err
	}
	return count > 0, nil
}

func (r *patientRepository) CountPrimaryClinicians(patientID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&entity.PatientAssignment{}).
		Where("patient_id = ? AND role = ?", patientID, "primary").
		Count(&count).Error; err != nil {
		r.log.Error("Failed to count primary clinicians", zap.Error(err),
			zap.String("patient_id", patientID.String()))
		return 0, err
	}
	return count, nil
}
