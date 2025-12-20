package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/payment/dto"
	paymentService "github.com/sahabatharianmu/OpenMind/internal/modules/payment/service"
	"github.com/sahabatharianmu/OpenMind/pkg/midtrans"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

// PaymentTransactionHandler handles HTTP requests for payment transactions
type PaymentTransactionHandler struct {
	svc paymentService.PaymentTransactionService
}

// NewPaymentTransactionHandler creates a new payment transaction handler
func NewPaymentTransactionHandler(svc paymentService.PaymentTransactionService) *PaymentTransactionHandler {
	return &PaymentTransactionHandler{
		svc: svc,
	}
}

// getOrganizationID retrieves the organization ID for the current user
func (h *PaymentTransactionHandler) getOrganizationID(c *app.RequestContext) (uuid.UUID, error) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, response.ErrUnauthorized
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		return uuid.Nil, err
	}

	return orgID, nil
}

// CreateQRISPayment handles POST /payments/qris/create
func (h *PaymentTransactionHandler) CreateQRISPayment(_ context.Context, c *app.RequestContext) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req dto.CreateQRISPaymentRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Set default currency if not provided
	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Set default type if not provided
	if req.Type == "" {
		req.Type = "subscription"
	}

	resp, err := h.svc.CreateQRISPayment(context.Background(), orgID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, resp, "QRIS payment created successfully")
}

// CheckPaymentStatus handles GET /payments/qris/status/:id
func (h *PaymentTransactionHandler) CheckPaymentStatus(_ context.Context, c *app.RequestContext) {
	orgID, err := h.getOrganizationID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	transactionIDStr := c.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", nil)
		return
	}

	resp, err := h.svc.CheckPaymentStatus(context.Background(), orgID, transactionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Payment status retrieved successfully", resp))
}

// HandleQRISWebhook handles POST /webhooks/midtrans/v1.0/qr/qr-mpm-notify
// This is a public endpoint that Midtrans calls to notify us of QRIS payment status changes
func (h *PaymentTransactionHandler) HandleQRISWebhook(_ context.Context, c *app.RequestContext) {
	// Read raw request body (needed for signature validation)
	bodyBytes := c.Request.Body()
	if len(bodyBytes) == 0 {
		response.BadRequest(c, "Empty request body", nil)
		return
	}

	headers := map[string]string{
		midtrans.HeaderXTimestamp:  c.Request.Header.Get(midtrans.HeaderXTimestamp),
		midtrans.HeaderXSignature:  c.Request.Header.Get(midtrans.HeaderXSignature),
		midtrans.HeaderXPartnerID:  c.Request.Header.Get(midtrans.HeaderXPartnerID),
		midtrans.HeaderXExternalID: c.Request.Header.Get(midtrans.HeaderXExternalID),
		midtrans.HeaderChannelID:   c.Request.Header.Get(midtrans.HeaderChannelID),
	}

	// Process QRIS webhook (includes signature validation via midtrans.HandleQRISWebhook)
	if err := h.svc.ProcessQRISWebhook(context.Background(), bodyBytes, headers); err != nil {
		response.HandleError(c, err)
		return
	}

	// Return success response (Midtrans expects 200 OK)
	c.JSON(consts.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Webhook processed successfully",
	})
}
