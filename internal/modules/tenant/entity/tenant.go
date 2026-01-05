package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tenant represents a tenant (organization) with its associated PostgreSQL schema
type Tenant struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"                  json:"organization_id"`
	SchemaName     string         `gorm:"type:varchar(63);not null;uniqueIndex"           json:"schema_name"`
	Status         string         `gorm:"type:varchar(50);not null;default:'active'"      json:"status"` // active, suspended, deleted
	CreatedAt      time.Time      `                                                       json:"created_at"`
	UpdatedAt      time.Time      `                                                       json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                                           json:"deleted_at,omitempty"`
}

func (Tenant) TableName() string {
	return "tenants"
}
