package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppointmentRepository interface {
	Create(appointment *entity.Appointment) error
	Update(appointment *entity.Appointment) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*entity.Appointment, error)
	List(organizationID uuid.UUID, limit, offset int) ([]entity.Appointment, int64, error)
	CheckOverlap(
		organizationID uuid.UUID,
		clinicianID uuid.UUID,
		startTime, endTime time.Time,
		excludeID *uuid.UUID,
	) (bool, error)
	GetOrganizationID(userID uuid.UUID) (uuid.UUID, error)
}

type appointmentRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewAppointmentRepository(db *gorm.DB, log logger.Logger) AppointmentRepository {
	return &appointmentRepository{
		db:  db,
		log: log,
	}
}

func (r *appointmentRepository) Create(appointment *entity.Appointment) error {
	if err := r.db.Create(appointment).Error; err != nil {
		r.log.Error("Failed to create appointment", zap.Error(err))
		return err
	}
	return nil
}

func (r *appointmentRepository) Update(appointment *entity.Appointment) error {
	if err := r.db.Save(appointment).Error; err != nil {
		r.log.Error("Failed to update appointment", zap.Error(err), zap.String("id", appointment.ID.String()))
		return err
	}
	return nil
}

func (r *appointmentRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entity.Appointment{}, "id = ?", id).Error; err != nil {
		r.log.Error("Failed to delete appointment", zap.Error(err), zap.String("id", id.String()))
		return err
	}
	return nil
}

func (r *appointmentRepository) FindByID(id uuid.UUID) (*entity.Appointment, error) {
	var appointment entity.Appointment
	if err := r.db.First(&appointment, "id = ?", id).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find appointment", zap.Error(err), zap.String("id", id.String()))
		}
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepository) List(organizationID uuid.UUID, limit, offset int) ([]entity.Appointment, int64, error) {
	var appointments []entity.Appointment
	var total int64

	query := r.db.Model(&entity.Appointment{}).Where("organization_id = ?", organizationID)

	if err := query.Count(&total).Error; err != nil {
		r.log.Error("Failed to count appointments", zap.Error(err))
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Order("start_time desc").Find(&appointments).Error; err != nil {
		r.log.Error("Failed to list appointments", zap.Error(err))
		return nil, 0, err
	}

	return appointments, total, nil
}

func (r *appointmentRepository) CheckOverlap(
	organizationID uuid.UUID,
	clinicianID uuid.UUID,
	startTime, endTime time.Time,
	excludeID *uuid.UUID,
) (bool, error) {
	var count int64
	query := r.db.Model(&entity.Appointment{}).
		Where("organization_id = ?", organizationID).
		Where("clinician_id = ?", clinicianID).
		Where("status != ?", "cancelled").
		Where("((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?))",
			endTime, startTime, endTime, startTime, startTime, endTime)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *appointmentRepository) GetOrganizationID(userID uuid.UUID) (uuid.UUID, error) {
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
