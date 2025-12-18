package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type AppointmentHandler struct {
	svc service.AppointmentService
}

func NewAppointmentHandler(svc service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{svc: svc}
}

func (h *AppointmentHandler) Create(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	var req dto.CreateAppointmentRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Create(context.Background(), req, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, resp, "Appointment created successfully")
}

func (h *AppointmentHandler) List(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	resp, total, err := h.svc.List(context.Background(), orgID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Appointments retrieved successfully", map[string]interface{}{
		"items":     resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}

func (h *AppointmentHandler) Get(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID", nil)
		return
	}

	resp, err := h.svc.Get(context.Background(), id, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Appointment retrieved successfully", resp))
}

func (h *AppointmentHandler) Update(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID", nil)
		return
	}

	var req dto.UpdateAppointmentRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Update(context.Background(), id, orgID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Appointment updated successfully", resp))
}

func (h *AppointmentHandler) Delete(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID", nil)
		return
	}

	if err := h.svc.Delete(context.Background(), id, orgID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Appointment deleted successfully", nil))
}
