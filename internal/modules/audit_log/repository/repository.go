package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(log *entity.AuditLog) error
	List(
		organizationID uuid.UUID,
		limit, offset int,
		filters *dto.FilterOptions,
	) ([]entity.AuditLog, int64, error)
	GetOrganizationID(userID uuid.UUID) (uuid.UUID, error)
}

type auditLogRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewAuditLogRepository(db *gorm.DB, log logger.Logger) AuditLogRepository {
	return &auditLogRepository{
		db:  db,
		log: log,
	}
}

func (r *auditLogRepository) Create(log *entity.AuditLog) error {
	if err := r.db.Create(log).Error; err != nil {
		r.log.Error("Failed to create audit log", zap.Error(err))
		return err
	}
	return nil
}

func (r *auditLogRepository) List(
	organizationID uuid.UUID,
	limit, offset int,
	filters *dto.FilterOptions,
) ([]entity.AuditLog, int64, error) {
	var logs []entity.AuditLog
	var total int64

	query := r.db.Model(&entity.AuditLog{}).Where("organization_id = ?", organizationID)

	// Apply filters
	if filters != nil {
		if filters.ResourceType != nil {
			query = query.Where("resource_type = ?", *filters.ResourceType)
		}
		if filters.UserID != nil {
			query = query.Where("user_id = ?", *filters.UserID)
		}
		if filters.Action != nil {
			query = query.Where("action = ?", *filters.Action)
		}
		if filters.StartDate != nil {
			query = query.Where("created_at >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			// Add 1 day to include the entire end date
			endOfDay := filters.EndDate.Add(24 * time.Hour)
			query = query.Where("created_at < ?", endOfDay)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		r.log.Error("Failed to count audit logs", zap.Error(err))
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&logs).Error; err != nil {
		r.log.Error("Failed to list audit logs", zap.Error(err))
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *auditLogRepository) GetOrganizationID(userID uuid.UUID) (uuid.UUID, error) {
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
