package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type PatientHandler struct {
	svc service.PatientService
}

func NewPatientHandler(svc service.PatientService) *PatientHandler {
	return &PatientHandler{svc: svc}
}

func (h *PatientHandler) Create(_ context.Context, c *app.RequestContext) {
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

	var req dto.CreatePatientRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Create(context.Background(), req, orgID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, resp, "Patient created successfully")
}

func (h *PatientHandler) List(_ context.Context, c *app.RequestContext) {
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

	c.JSON(consts.StatusOK, response.Success("Patients retrieved successfully", map[string]interface{}{
		"items":     resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}

func (h *PatientHandler) Get(_ context.Context, c *app.RequestContext) {
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
		response.BadRequest(c, "Invalid patient ID", nil)
		return
	}

	resp, err := h.svc.Get(context.Background(), id, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Patient retrieved successfully", resp))
}

func (h *PatientHandler) Update(_ context.Context, c *app.RequestContext) {
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
		response.BadRequest(c, "Invalid patient ID", nil)
		return
	}

	var req dto.UpdatePatientRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Update(context.Background(), id, orgID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Patient updated successfully", resp))
}

func (h *PatientHandler) Delete(_ context.Context, c *app.RequestContext) {
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
		response.BadRequest(c, "Invalid patient ID", nil)
		return
	}

	if err := h.svc.Delete(context.Background(), id, orgID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Patient deleted successfully", nil))
}
