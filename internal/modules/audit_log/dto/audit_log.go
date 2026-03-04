package dto

import (
	"time"

	"github.com/google/uuid"
)

type AuditLogResponse struct {
	ID             uuid.UUID   `json:"id"`
	OrganizationID uuid.UUID   `json:"organization_id"`
	UserID         uuid.UUID   `json:"user_id"`
	UserName       string      `json:"user_name"`
	Action         string      `json:"action"`
	ResourceType   string      `json:"resource_type"`
	ResourceID     *uuid.UUID  `json:"resource_id"`
	Details        interface{} `json:"details"`
	IPAddress      *string     `json:"ip_address"`
	CreatedAt      time.Time   `json:"created_at"`
}

type FilterOptions struct {
	ResourceType *string
	UserID       *uuid.UUID
	Action       *string
	StartDate    *time.Time
	EndDate      *time.Time
}
