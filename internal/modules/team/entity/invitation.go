package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamInvitation represents a team invitation sent to a user
type TeamInvitation struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Email          string         `gorm:"type:varchar(255);not null" json:"email"`
	Role           string         `gorm:"type:varchar(50);not null;default:'member'" json:"role"`
	Token          string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"token"`
	InvitedBy      uuid.UUID      `gorm:"type:uuid;not null" json:"invited_by"`
	Status         string         `gorm:"type:varchar(50);not null;default:'pending'" json:"status"` // pending, accepted, expired, cancelled
	ExpiresAt      time.Time      `gorm:"not null" json:"expires_at"`
	AcceptedAt     *time.Time     `gorm:"" json:"accepted_at,omitempty"`
	AcceptedBy     *uuid.UUID     `gorm:"type:uuid" json:"accepted_by,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (TeamInvitation) TableName() string {
	return "team_invitations"
}

// IsExpired checks if the invitation has expired
func (ti *TeamInvitation) IsExpired() bool {
	return time.Now().After(ti.ExpiresAt)
}

// IsPending checks if the invitation is still pending
func (ti *TeamInvitation) IsPending() bool {
	return ti.Status == "pending" && !ti.IsExpired()
}

// CanBeAccepted checks if the invitation can be accepted
func (ti *TeamInvitation) CanBeAccepted() bool {
	return ti.IsPending()
}

