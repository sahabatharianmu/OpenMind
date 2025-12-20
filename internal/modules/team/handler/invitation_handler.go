package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/core/middleware"
	teamDto "github.com/sahabatharianmu/OpenMind/internal/modules/team/dto"
	teamEntity "github.com/sahabatharianmu/OpenMind/internal/modules/team/entity"
	teamService "github.com/sahabatharianmu/OpenMind/internal/modules/team/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type TeamInvitationHandler struct {
	service teamService.TeamInvitationService
}

func NewTeamInvitationHandler(service teamService.TeamInvitationService) *TeamInvitationHandler {
	return &TeamInvitationHandler{
		service: service,
	}
}

// SendInvitation sends a team invitation
// POST /api/v1/team/invitations
func (h *TeamInvitationHandler) SendInvitation(ctx context.Context, c *app.RequestContext) {
	// Get organization ID from context
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Organization context not found")
		return
	}

	// Get user ID from context
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User context not found")
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req teamDto.SendInvitationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	invitation, err := h.service.SendInvitation(ctx, orgID, userID, req.Email, req.Role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Invitation sent successfully", h.mapToResponse(invitation)))
}

// AcceptInvitation accepts a team invitation (for authenticated users)
// POST /api/v1/team/invitations/accept
func (h *TeamInvitationHandler) AcceptInvitation(ctx context.Context, c *app.RequestContext) {
	var req teamDto.AcceptInvitationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context (user must be authenticated)
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User must be authenticated to accept invitation")
		return
	}
	userID := userIDVal.(uuid.UUID)
	userIDPtr := &userID

	if err := h.service.AcceptInvitation(ctx, req.Token, userIDPtr); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Invitation accepted successfully", nil))
}

// RegisterAndAcceptInvitation creates a new user account and accepts the invitation
// POST /api/v1/team/invitations/register
func (h *TeamInvitationHandler) RegisterAndAcceptInvitation(ctx context.Context, c *app.RequestContext) {
	var req teamDto.RegisterWithInvitationRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	userID, err := h.service.RegisterAndAcceptInvitation(ctx, req.Token, req.Email, req.Password, req.FullName)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Account created and invitation accepted successfully", map[string]interface{}{
		"user_id": userID,
	}))
}

// GetInvitation retrieves an invitation by token (for public access)
// GET /api/v1/team/invitations/:token
func (h *TeamInvitationHandler) GetInvitation(ctx context.Context, c *app.RequestContext) {
	token := c.Param("token")
	if token == "" {
		response.BadRequest(c, "Token is required", nil)
		return
	}

	invitation, err := h.service.GetInvitationByToken(ctx, token)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("", h.mapToResponse(invitation)))
}

// ListInvitations lists all invitations for the current organization
// GET /api/v1/team/invitations
func (h *TeamInvitationHandler) ListInvitations(ctx context.Context, c *app.RequestContext) {
	// Get organization ID from context
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Organization context not found")
		return
	}

	// Get pagination parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	invitations, total, err := h.service.ListInvitations(ctx, orgID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	responses := make([]teamDto.TeamInvitationResponse, len(invitations))
	for i, inv := range invitations {
		responses[i] = h.mapToResponse(&inv)
	}

	response.PaginatedResponse(c, teamDto.ListInvitationsResponse{
		Invitations: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, page, pageSize, total, "Invitations retrieved successfully")
}

// CancelInvitation cancels an invitation
// DELETE /api/v1/team/invitations/:id
func (h *TeamInvitationHandler) CancelInvitation(ctx context.Context, c *app.RequestContext) {
	// Get organization ID from context
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Organization context not found")
		return
	}

	invitationIDStr := c.Param("id")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid invitation ID", nil)
		return
	}

	if err := h.service.CancelInvitation(ctx, invitationID, orgID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Invitation cancelled successfully", nil))
}

// ResendInvitation resends an invitation email
// POST /api/v1/team/invitations/:id/resend
func (h *TeamInvitationHandler) ResendInvitation(ctx context.Context, c *app.RequestContext) {
	// Get organization ID from context
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Organization context not found")
		return
	}

	invitationIDStr := c.Param("id")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid invitation ID", nil)
		return
	}

	if err := h.service.ResendInvitation(ctx, invitationID, orgID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Invitation resent successfully", nil))
}

// mapToResponse maps entity to DTO
func (h *TeamInvitationHandler) mapToResponse(invitation *teamEntity.TeamInvitation) teamDto.TeamInvitationResponse {
	return teamDto.TeamInvitationResponse{
		ID:             invitation.ID,
		OrganizationID: invitation.OrganizationID,
		Email:          invitation.Email,
		Role:           invitation.Role,
		Status:         invitation.Status,
		ExpiresAt:      invitation.ExpiresAt,
		AcceptedAt:     invitation.AcceptedAt,
		CreatedAt:      invitation.CreatedAt,
	}
}

