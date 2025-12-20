package dto

import (
	"time"

	"github.com/google/uuid"
)

// NotificationResponse represents a notification
type NotificationResponse struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Type             string     `json:"type"`
	Title            string     `json:"title"`
	Message          string     `json:"message"`
	RelatedEntityType *string   `json:"related_entity_type"`
	RelatedEntityID   *uuid.UUID `json:"related_entity_id"`
	IsRead           bool       `json:"is_read"`
	ReadAt           *time.Time  `json:"read_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// NotificationListResponse represents a list of notifications
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	UnreadCount   int64                  `json:"unread_count"`
}

