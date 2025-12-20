package dto

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	TaxID       string    `json:"tax_id"`
	NPI         string    `json:"npi"`
	Address     string    `json:"address"`
	Currency    string    `json:"currency"`
	Locale      string    `json:"locale"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateOrganizationRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Type     string `json:"type"`
	TaxID    string `json:"tax_id"`
	NPI      string `json:"npi"`
	Address  string `json:"address"`
	Currency string `json:"currency"`
	Locale   string `json:"locale"`
}

type TeamMemberResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}