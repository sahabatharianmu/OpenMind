package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SubscriptionPlan represents a subscription package (e.g. Free, Pro, Enterprise)
type SubscriptionPlan struct {
	ID          uuid.UUID               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string                  `gorm:"type:varchar(100);not null;unique"               json:"name"`
	Description string                  `gorm:"type:text"                                       json:"description"`
	Prices      []SubscriptionPlanPrice `gorm:"foreignKey:PlanID"                               json:"prices"`
	Limits      datatypes.JSON          `gorm:"type:jsonb"                                      json:"limits"` // e.g. {"max_patients": 100, "max_clinicians": 5}
	IsActive    bool                    `gorm:"not null;default:true"                           json:"is_active"`
	CreatedAt   time.Time               `                                                       json:"created_at"`
	UpdatedAt   time.Time               `                                                       json:"updated_at"`
	DeletedAt   gorm.DeletedAt          `gorm:"index"                                           json:"deleted_at,omitempty"`
}

// SubscriptionPlanPrice represents a price for a subscription plan in a specific currency
type SubscriptionPlanPrice struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PlanID    uuid.UUID      `gorm:"type:uuid;not null"                              json:"plan_id"`
	Currency  string         `gorm:"type:varchar(10);not null"                       json:"currency"`
	Price     int64          `gorm:"not null;default:0"                              json:"price"` // In cents
	CreatedAt time.Time      `                                                       json:"created_at"`
	UpdatedAt time.Time      `                                                       json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                           json:"deleted_at,omitempty"`
}

func (SubscriptionPlanPrice) TableName() string {
	return "subscription_plan_prices"
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}
