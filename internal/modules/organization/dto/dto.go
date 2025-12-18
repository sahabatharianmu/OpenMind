package dto

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}
