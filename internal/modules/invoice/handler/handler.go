package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/invoice/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/invoice/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type InvoiceHandler struct {
	svc service.InvoiceService
}

func NewInvoiceHandler(svc service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

func (h *InvoiceHandler) Create(_ context.Context, c *app.RequestContext) {
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

	var req dto.CreateInvoiceRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Create(context.Background(), req, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, resp, "Invoice created successfully")
}

func (h *InvoiceHandler) List(_ context.Context, c *app.RequestContext) {
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

	c.JSON(consts.StatusOK, response.Success("Invoices retrieved successfully", map[string]interface{}{
		"items":     resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}

func (h *InvoiceHandler) Get(_ context.Context, c *app.RequestContext) {
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
		response.BadRequest(c, "Invalid invoice ID", nil)
		return
	}

	resp, err := h.svc.Get(context.Background(), id, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Invoice retrieved successfully", resp))
}

func (h *InvoiceHandler) Update(_ context.Context, c *app.RequestContext) {
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
		response.BadRequest(c, "Invalid invoice ID", nil)
		return
	}

	var req dto.UpdateInvoiceRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Update(context.Background(), id, orgID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Invoice updated successfully", resp))
}

func (h *InvoiceHandler) Delete(_ context.Context, c *app.RequestContext) {
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
		response.BadRequest(c, "Invalid invoice ID", nil)
		return
	}

	if err := h.svc.Delete(context.Background(), id, orgID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Invoice deleted successfully", nil))
}
