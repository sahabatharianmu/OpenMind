package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID             uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_audit_logs_org"     json:"organization_id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index:idx_audit_logs_user"    json:"user_id"`
	Action         string         `gorm:"type:varchar(50);not null"                       json:"action"`
	ResourceType   string         `gorm:"type:varchar(50);not null;index"                 json:"resource_type"`
	ResourceID     *uuid.UUID     `gorm:"type:uuid;index"                                 json:"resource_id"`
	Details        datatypes.JSON `gorm:"type:jsonb"                                      json:"details"`
	IPAddress      *string        `gorm:"type:varchar(45)"                                json:"ip_address"`
	UserAgent      *string        `gorm:"type:text"                                       json:"user_agent"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;index:idx_audit_logs_org"         json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                                           json:"-"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeUpdate is a GORM hook that prevents updates to audit logs
// This provides application-level protection in addition to database triggers
func (a *AuditLog) BeforeUpdate(tx *gorm.DB) error {
	// Audit logs are immutable and cannot be updated
	return errors.New("cannot update audit logs: audit logs are immutable for compliance and must be retained as-is")
}

// BeforeDelete is a GORM hook that prevents deletion of audit logs
// This provides application-level protection in addition to database triggers
func (a *AuditLog) BeforeDelete(tx *gorm.DB) error {
	// Audit logs are immutable and cannot be deleted
	return errors.New("cannot delete audit logs: audit logs are immutable for compliance and must be retained")
}
