package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                  json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                                           json:"-"`
	Email        string         `gorm:"uniqueIndex;not null"                            json:"email"`
	PasswordHash string         `gorm:"not null"                                        json:"-"` // Never return password hash in JSON
	FullName     string         `gorm:"not null"                                        json:"full_name"`
	SystemRole   string         `gorm:"type:varchar(50);not null;default:'user'"        json:"system_role"` // user, admin (platform admin)
	// Note: Role is now per-organization in organization_members table
}

type Organization struct {
	ID               uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name             string         `gorm:"not null"                                        json:"name"`
	Type             string         `gorm:"not null;default:'clinic'"                       json:"type"`
	SubscriptionTier string         `gorm:"type:varchar(50);not null;default:'free'"        json:"subscription_tier"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"                                  json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index"                                           json:"-"`
}

func (Organization) TableName() string {
	return "organizations"
}

type OrganizationMember struct {
	OrganizationID uuid.UUID `gorm:"primaryKey;type:uuid"      json:"organization_id"`
	UserID         uuid.UUID `gorm:"primaryKey;type:uuid"      json:"user_id"`
	Role           string    `gorm:"not null;default:'member'" json:"role"`
	CreatedAt      time.Time `gorm:"autoCreateTime"            json:"created_at"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}
