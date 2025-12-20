package entity

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents an in-app notification for a user
type Notification struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Type             string     `gorm:"type:varchar(100);not null" json:"type"`
	Title            string     `gorm:"type:varchar(255);not null" json:"title"`
	Message          string     `gorm:"type:text;not null" json:"message"`
	RelatedEntityType *string   `gorm:"type:varchar(100)" json:"related_entity_type"`
	RelatedEntityID   *uuid.UUID `gorm:"type:uuid" json:"related_entity_id"`
	IsRead           bool       `gorm:"not null;default:false" json:"is_read"`
	ReadAt           *time.Time `gorm:"" json:"read_at"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// Notification type constants
const (
	TypeHandoffRequest   = "handoff_request"
	TypeHandoffApproved  = "handoff_approved"
	TypeHandoffRejected  = "handoff_rejected"
	TypeHandoffCancelled = "handoff_cancelled"
	TypeTeamInvitation   = "team_invitation"
)

// Related entity type constants
const (
	RelatedEntityTypePatientHandoff = "patient_handoff"
	RelatedEntityTypeTeamInvitation = "team_invitation"
)

