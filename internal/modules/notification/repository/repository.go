package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/notification/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *entity.Notification) error
	GetByUserID(userID uuid.UUID, limit, offset int) ([]entity.Notification, error)
	GetUnreadByUserID(userID uuid.UUID, limit, offset int) ([]entity.Notification, error)
	MarkAsRead(notificationID uuid.UUID) error
	MarkAllAsRead(userID uuid.UUID) error
	GetUnreadCount(userID uuid.UUID) (int64, error)
	GetByID(id uuid.UUID) (*entity.Notification, error)
}

type notificationRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewNotificationRepository(db *gorm.DB, log logger.Logger) NotificationRepository {
	return &notificationRepository{
		db:  db,
		log: log,
	}
}

func (r *notificationRepository) Create(notification *entity.Notification) error {
	if err := r.db.Create(notification).Error; err != nil {
		r.log.Error("Failed to create notification", zap.Error(err),
			zap.String("user_id", notification.UserID.String()),
			zap.String("type", notification.Type))
		return err
	}
	return nil
}

func (r *notificationRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]entity.Notification, error) {
	var notifications []entity.Notification
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	if err := query.Find(&notifications).Error; err != nil {
		r.log.Error("Failed to get notifications by user ID", zap.Error(err),
			zap.String("user_id", userID.String()))
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) GetUnreadByUserID(userID uuid.UUID, limit, offset int) ([]entity.Notification, error) {
	var notifications []entity.Notification
	query := r.db.Where("user_id = ? AND is_read = ?", userID, false).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	if err := query.Find(&notifications).Error; err != nil {
		r.log.Error("Failed to get unread notifications by user ID", zap.Error(err),
			zap.String("user_id", userID.String()))
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(notificationID uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&entity.Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error; err != nil {
		r.log.Error("Failed to mark notification as read", zap.Error(err),
			zap.String("notification_id", notificationID.String()))
		return err
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(userID uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error; err != nil {
		r.log.Error("Failed to mark all notifications as read", zap.Error(err),
			zap.String("user_id", userID.String()))
		return err
	}
	return nil
}

func (r *notificationRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		r.log.Error("Failed to get unread notification count", zap.Error(err),
			zap.String("user_id", userID.String()))
		return 0, err
	}
	return count, nil
}

func (r *notificationRepository) GetByID(id uuid.UUID) (*entity.Notification, error) {
	var notification entity.Notification
	if err := r.db.First(&notification, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Error("Failed to get notification by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, err
	}
	return &notification, nil
}
