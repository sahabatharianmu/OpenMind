package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SubscriptionPlan represents a subscription package (e.g. Free, Pro, Enterprise)
type SubscriptionPlan struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;unique" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Price       int64          `gorm:"not null;default:0" json:"price"` // In cents (e.g. 1000 = $10.00)
	Currency    string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"`
	Limits      datatypes.JSON `gorm:"type:jsonb" json:"limits"` // e.g. {"max_patients": 100, "max_clinicians": 5}
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}
