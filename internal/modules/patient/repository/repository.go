package repository

import (
	"errors"

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
	List(organizationID uuid.UUID, limit, offset int) ([]entity.Patient, int64, error)
	GetOrganizationID(userID uuid.UUID) (uuid.UUID, error)
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

func (r *patientRepository) List(organizationID uuid.UUID, limit, offset int) ([]entity.Patient, int64, error) {
	var patients []entity.Patient
	var total int64

	query := r.db.Model(&entity.Patient{}).Where("organization_id = ?", organizationID)

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
