package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	notificationService "github.com/sahabatharianmu/OpenMind/internal/modules/notification/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type NotificationHandler struct {
	service notificationService.NotificationService
}

func NewNotificationHandler(service notificationService.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

// GetNotifications handles GET /notifications
func (h *NotificationHandler) GetNotifications(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse pagination parameters
	limit := 50 // default
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	notifications, total, err := h.service.GetUserNotifications(context.Background(), userID, limit, offset)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	unreadCount, err := h.service.GetUnreadCount(context.Background(), userID)
	if err != nil {
		// Log but don't fail the request
		unreadCount = 0
	}

	c.JSON(consts.StatusOK, response.Success("Notifications retrieved successfully", map[string]interface{}{
		"notifications": notifications,
		"total":         total,
		"unread_count":  unreadCount,
	}))
}

// MarkAsRead handles PUT /notifications/:id/read
func (h *NotificationHandler) MarkAsRead(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid notification ID", nil)
		return
	}

	if err := h.service.MarkAsRead(context.Background(), notificationID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Notification marked as read", nil))
}

// MarkAllAsRead handles PUT /notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	if err := h.service.MarkAllAsRead(context.Background(), userID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("All notifications marked as read", nil))
}

