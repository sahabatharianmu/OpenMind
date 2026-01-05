package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name               string         `gorm:"type:varchar(255);not null"                       json:"name"`
	Type               string         `gorm:"type:varchar(50);not null"                        json:"type"`
	SubscriptionPlanID *uuid.UUID     `gorm:"type:uuid"                                        json:"subscription_plan_id"`
	SubscriptionTier   string         `gorm:"type:varchar(50);not null;default:'free'"        json:"subscription_tier"`
	TaxID              string         `gorm:"type:varchar(50)"                                 json:"tax_id"`
	NPI                string         `gorm:"type:varchar(50)"                                 json:"npi"`
	Address            string         `gorm:"type:text"                                        json:"address"`
	Currency           string         `gorm:"type:varchar(10);not null;default:'USD'"          json:"currency"`
	Locale             string         `gorm:"type:varchar(10);not null;default:'en-US'"        json:"locale"`
	CreatedAt          time.Time      `                                                        json:"created_at"`
	UpdatedAt          time.Time      `                                                        json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index"                                            json:"deleted_at,omitempty"`
}

type OrganizationMember struct {
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"        json:"organization_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"        json:"user_id"`
	Role           string    `gorm:"type:varchar(50);not null" json:"role"` // owner, admin, member
	CreatedAt      time.Time `                                 json:"created_at"`
}
