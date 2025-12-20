package dto

import (
	"time"

	"github.com/google/uuid"
)

// SendInvitationRequest represents a request to send a team invitation
type SendInvitationRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// AcceptInvitationRequest represents a request to accept an invitation
type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}

// RegisterWithInvitationRequest represents a request to register and accept an invitation
type RegisterWithInvitationRequest struct {
	Token     string `json:"token" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FullName  string `json:"full_name" binding:"required,min=2"`
}

// TeamInvitationResponse represents a team invitation response
type TeamInvitationResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ListInvitationsResponse represents a paginated list of invitations
type ListInvitationsResponse struct {
	Invitations []TeamInvitationResponse `json:"invitations"`
	Total       int64                    `json:"total"`
	Page        int                       `json:"page"`
	PageSize    int                       `json:"page_size"`
}

