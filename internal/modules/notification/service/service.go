package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/notification/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/notification/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/notification/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

type NotificationService interface {
	CreateNotification(
		ctx context.Context,
		userID uuid.UUID,
		notificationType, title, message string,
		relatedEntityType *string,
		relatedEntityID *uuid.UUID,
	) error
	GetUserNotifications(
		ctx context.Context,
		userID uuid.UUID,
		limit, offset int,
	) ([]dto.NotificationResponse, int64, error)
	MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

type notificationService struct {
	repo repository.NotificationRepository
	log  logger.Logger
}

func NewNotificationService(repo repository.NotificationRepository, log logger.Logger) NotificationService {
	return &notificationService{
		repo: repo,
		log:  log,
	}
}

func (s *notificationService) CreateNotification(
	ctx context.Context,
	userID uuid.UUID,
	notificationType, title, message string,
	relatedEntityType *string,
	relatedEntityID *uuid.UUID,
) error {
	notification := &entity.Notification{
		UserID:            userID,
		Type:              notificationType,
		Title:             title,
		Message:           message,
		RelatedEntityType: relatedEntityType,
		RelatedEntityID:   relatedEntityID,
		IsRead:            false,
	}

	if err := s.repo.Create(notification); err != nil {
		s.log.Error("Failed to create notification", zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("type", notificationType))
		return err
	}

	s.log.Info("Notification created", zap.String("user_id", userID.String()),
		zap.String("type", notificationType))

	return nil
}

func (s *notificationService) GetUserNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.NotificationResponse, int64, error) {
	notifications, err := s.repo.GetByUserID(userID, limit, offset)
	if err != nil {
		s.log.Error("Failed to get user notifications", zap.Error(err),
			zap.String("user_id", userID.String()))
		return nil, 0, err
	}

	responses := make([]dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = dto.NotificationResponse{
			ID:                n.ID,
			UserID:            n.UserID,
			Type:              n.Type,
			Title:             n.Title,
			Message:           n.Message,
			RelatedEntityType: n.RelatedEntityType,
			RelatedEntityID:   n.RelatedEntityID,
			IsRead:            n.IsRead,
			ReadAt:            n.ReadAt,
			CreatedAt:         n.CreatedAt,
			UpdatedAt:         n.UpdatedAt,
		}
	}

	// Get total count (approximate, using limit if needed)
	total := int64(len(notifications))
	if limit > 0 && len(notifications) == limit {
		// Might be more, but we'll use this as approximation
		total = int64(offset + limit)
	}

	return responses, total, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	// Verify notification belongs to user
	notification, err := s.repo.GetByID(notificationID)
	if err != nil {
		s.log.Error("Failed to get notification", zap.Error(err),
			zap.String("notification_id", notificationID.String()))
		return err
	}

	if notification == nil {
		return fmt.Errorf("notification not found")
	}

	if notification.UserID != userID {
		return fmt.Errorf("notification does not belong to user")
	}

	if err := s.repo.MarkAsRead(notificationID); err != nil {
		s.log.Error("Failed to mark notification as read", zap.Error(err),
			zap.String("notification_id", notificationID.String()))
		return err
	}

	return nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.MarkAllAsRead(userID); err != nil {
		s.log.Error("Failed to mark all notifications as read", zap.Error(err),
			zap.String("user_id", userID.String()))
		return err
	}

	return nil
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repo.GetUnreadCount(userID)
	if err != nil {
		s.log.Error("Failed to get unread notification count", zap.Error(err),
			zap.String("user_id", userID.String()))
		return 0, err
	}

	return count, nil
}
