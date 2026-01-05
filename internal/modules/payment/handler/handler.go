package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/organization/service"
	"github.com/sahabatharianmu/OpenMind/internal/modules/payment/dto"
	paymentService "github.com/sahabatharianmu/OpenMind/internal/modules/payment/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

// PaymentMethodHandler handles HTTP requests for payment methods
type PaymentMethodHandler struct {
	svc        paymentService.PaymentMethodService
	orgService service.OrganizationService
}

// NewPaymentMethodHandler creates a new payment method handler
func NewPaymentMethodHandler(
	svc paymentService.PaymentMethodService,
	orgService service.OrganizationService,
) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		svc:        svc,
		orgService: orgService,
	}
}

// getOrganizationID retrieves the organization ID for the current user
func (h *PaymentMethodHandler) getOrganizationID(c *app.RequestContext) (uuid.UUID, error) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, response.ErrUnauthorized
	}
	userID := userIDVal.(uuid.UUID)

	org, err := h.orgService.GetMyOrganization(userID)
	if err != nil {
		return uuid.Nil, err
	}

	return org.ID, nil
}

// CreatePaymentMethod handles POST /payment-methods
func (h *PaymentMethodHandler) CreatePaymentMethod(_ context.Context, c *app.RequestContext) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.CreatePaymentMethodRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.CreatePaymentMethod(context.Background(), orgID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, resp, "Payment method added successfully")
}

// ListPaymentMethods handles GET /payment-methods
func (h *PaymentMethodHandler) ListPaymentMethods(_ context.Context, c *app.RequestContext) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	resp, err := h.svc.ListPaymentMethods(context.Background(), orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Payment methods retrieved successfully", resp))
}

// GetPaymentMethod handles GET /payment-methods/:id
func (h *PaymentMethodHandler) GetPaymentMethod(_ context.Context, c *app.RequestContext) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	paymentMethodIDStr := c.Param("id")
	paymentMethodID, err := uuid.Parse(paymentMethodIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", nil)
		return
	}

	resp, err := h.svc.GetPaymentMethod(context.Background(), orgID, paymentMethodID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Payment method retrieved successfully", resp))
}

// DeletePaymentMethod handles DELETE /payment-methods/:id
func (h *PaymentMethodHandler) DeletePaymentMethod(_ context.Context, c *app.RequestContext) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	paymentMethodIDStr := c.Param("id")
	paymentMethodID, err := uuid.Parse(paymentMethodIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", nil)
		return
	}

	err = h.svc.DeletePaymentMethod(context.Background(), orgID, paymentMethodID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Payment method deleted successfully", nil))
}

// SetDefaultPaymentMethod handles PUT /payment-methods/:id/default
func (h *PaymentMethodHandler) SetDefaultPaymentMethod(_ context.Context, c *app.RequestContext) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	paymentMethodIDStr := c.Param("id")
	paymentMethodID, err := uuid.Parse(paymentMethodIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", nil)
		return
	}

	err = h.svc.SetDefaultPaymentMethod(context.Background(), orgID, paymentMethodID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Default payment method updated successfully", nil))
}
