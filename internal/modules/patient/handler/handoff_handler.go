package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type PatientHandoffHandler struct {
	service service.PatientHandoffService
	baseURL string
}

func NewPatientHandoffHandler(service service.PatientHandoffService, baseURL string) *PatientHandoffHandler {
	return &PatientHandoffHandler{
		service: service,
		baseURL: baseURL,
	}
}

// RequestHandoff handles POST /patients/:id/handoff
func (h *PatientHandoffHandler) RequestHandoff(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgIDVal, exists := c.Get("organization_id")
	if !exists {
		response.Unauthorized(c, "Organization not found")
		return
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid organization ID")
		return
	}

	patientIDStr := c.Param("id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid patient ID", nil)
		return
	}

	var req dto.RequestHandoffRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	handoff, err := h.service.RequestHandoff(
		context.Background(),
		patientID,
		req.ReceivingClinicianID,
		userID,
		orgID,
		req.Message,
		req.Role,
		h.baseURL,
	)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Handoff request created successfully", handoff))
}

// ApproveHandoff handles POST /patients/handoffs/:id/approve
func (h *PatientHandoffHandler) ApproveHandoff(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgIDVal, exists := c.Get("organization_id")
	if !exists {
		response.Unauthorized(c, "Organization not found")
		return
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid organization ID")
		return
	}

	handoffIDStr := c.Param("id")
	handoffID, err := uuid.Parse(handoffIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid handoff ID", nil)
		return
	}

	var req dto.ApproveHandoffRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := h.service.ApproveHandoff(
		context.Background(),
		handoffID,
		userID,
		orgID,
		req.Reason,
		h.baseURL,
	); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Handoff approved successfully", nil))
}

// RejectHandoff handles POST /patients/handoffs/:id/reject
func (h *PatientHandoffHandler) RejectHandoff(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgIDVal, exists := c.Get("organization_id")
	if !exists {
		response.Unauthorized(c, "Organization not found")
		return
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid organization ID")
		return
	}

	handoffIDStr := c.Param("id")
	handoffID, err := uuid.Parse(handoffIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid handoff ID", nil)
		return
	}

	var req dto.RejectHandoffRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := h.service.RejectHandoff(
		context.Background(),
		handoffID,
		userID,
		orgID,
		req.Reason,
		h.baseURL,
	); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Handoff rejected successfully", nil))
}

// CancelHandoff handles POST /patients/handoffs/:id/cancel
func (h *PatientHandoffHandler) CancelHandoff(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgIDVal, exists := c.Get("organization_id")
	if !exists {
		response.Unauthorized(c, "Organization not found")
		return
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid organization ID")
		return
	}

	handoffIDStr := c.Param("id")
	handoffID, err := uuid.Parse(handoffIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid handoff ID", nil)
		return
	}

	if err := h.service.CancelHandoff(
		context.Background(),
		handoffID,
		userID,
		orgID,
	); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Handoff cancelled successfully", nil))
}

// GetHandoff handles GET /patients/handoffs/:id
func (h *PatientHandoffHandler) GetHandoff(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgIDVal, exists := c.Get("organization_id")
	if !exists {
		response.Unauthorized(c, "Organization not found")
		return
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid organization ID")
		return
	}

	handoffIDStr := c.Param("id")
	handoffID, err := uuid.Parse(handoffIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid handoff ID", nil)
		return
	}

	handoff, err := h.service.GetHandoff(context.Background(), handoffID, userID, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Handoff retrieved successfully", handoff))
}

// ListHandoffs handles GET /patients/:id/handoffs
func (h *PatientHandoffHandler) ListHandoffs(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgIDVal, exists := c.Get("organization_id")
	if !exists {
		response.Unauthorized(c, "Organization not found")
		return
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid organization ID")
		return
	}

	patientIDStr := c.Param("id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid patient ID", nil)
		return
	}

	handoffs, err := h.service.ListHandoffs(context.Background(), patientID, userID, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Handoffs retrieved successfully", handoffs))
}

// ListPendingHandoffs handles GET /patients/handoffs/pending
func (h *PatientHandoffHandler) ListPendingHandoffs(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgIDVal, exists := c.Get("organization_id")
	if !exists {
		response.Unauthorized(c, "Organization not found")
		return
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid organization ID")
		return
	}

	handoffs, err := h.service.ListPendingHandoffs(context.Background(), userID, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Pending handoffs retrieved successfully", handoffs))
}
