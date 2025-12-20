package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	"github.com/sahabatharianmu/OpenMind/internal/modules/payment/dto"
	paymentEntity "github.com/sahabatharianmu/OpenMind/internal/modules/payment/entity"
	paymentRepo "github.com/sahabatharianmu/OpenMind/internal/modules/payment/repository"
	tenantService "github.com/sahabatharianmu/OpenMind/internal/modules/tenant/service"
	userRepository "github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/email"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/midtrans"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentTransactionService defines the interface for payment transaction operations
type PaymentTransactionService interface {
	CreateQRISPayment(ctx context.Context, organizationID uuid.UUID, req dto.CreateQRISPaymentRequest) (*dto.QRISPaymentResponse, error)
	CheckPaymentStatus(ctx context.Context, organizationID, transactionID uuid.UUID) (*dto.CheckPaymentStatusResponse, error)
	ProcessQRISWebhook(ctx context.Context, payload []byte, headers map[string]string) error
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type paymentTransactionService struct {
	transactionRepo  paymentRepo.PaymentTransactionRepository
	organizationRepo repository.OrganizationRepository
	userRepo         userRepository.UserRepository
	emailService     *email.EmailService
	midtransService  *midtrans.Service
	tenantService    tenantService.TenantService
	db               *gorm.DB
	log              logger.Logger
	appURL           string
}

// NewPaymentTransactionService creates a new PaymentTransactionService
func NewPaymentTransactionService(
	transactionRepo paymentRepo.PaymentTransactionRepository,
	organizationRepo repository.OrganizationRepository,
	userRepo userRepository.UserRepository,
	emailService *email.EmailService,
	midtransService *midtrans.Service,
	tenantService tenantService.TenantService,
	db *gorm.DB,
	log logger.Logger,
	appURL string,
) PaymentTransactionService {
	return &paymentTransactionService{
		transactionRepo:  transactionRepo,
		organizationRepo: organizationRepo,
		userRepo:         userRepo,
		emailService:     emailService,
		midtransService:  midtransService,
		tenantService:    tenantService,
		db:               db,
		log:              log,
		appURL:           appURL,
	}
}

// CreateQRISPayment creates a QRIS payment transaction
func (s *paymentTransactionService) CreateQRISPayment(
	ctx context.Context,
	organizationID uuid.UUID,
	req dto.CreateQRISPaymentRequest,
) (*dto.QRISPaymentResponse, error) {
	// Validate type
	if req.Type != "subscription" && req.Type != "one_time" {
		return nil, response.NewBadRequest("invalid payment type, must be 'subscription' or 'one_time'")
	}

	// Convert USD to IDR for Midtrans QRIS (Midtrans QRIS uses IDR)
	// Exchange rate: 1 USD = ~15000 IDR (this should be configurable or fetched from API)
	usdAmount := decimal.NewFromFloat(req.Amount)
	exchangeRate := decimal.NewFromInt(15000) // TODO: Make this configurable or fetch from API
	idrAmount := usdAmount.Mul(exchangeRate)

	// Generate unique partner reference number
	partnerReferenceNo := fmt.Sprintf("PAY-%s-%d", organizationID.String()[:8], time.Now().Unix())

	// Create QRIS payment via Midtrans
	midtransResp, err := s.midtransService.CreateQRISPayment(ctx, partnerReferenceNo, idrAmount)
	if err != nil {
		s.log.Error("Failed to create QRIS payment via Midtrans", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, response.NewInternalServerError(fmt.Sprintf("Failed to create QRIS payment: %v", err))
	}

	// Calculate expiry time (15 minutes from now, as per Midtrans default)
	expiresAt := time.Now().Add(15 * time.Minute)

	// Convert amount to cents for storage (storing USD amount in cents)
	amountInCents := int64(req.Amount * 100)

	// Create payment transaction entity
	transaction := &paymentEntity.PaymentTransaction{
		OrganizationID:        organizationID,
		Type:                  req.Type,
		PaymentMethod:         "qris",
		Provider:              "midtrans",
		Amount:                amountInCents,
		Currency:              req.Currency,
		Status:                "pending",
		ProviderTransactionID: midtransResp.TransactionID,
		PartnerReferenceNo:    partnerReferenceNo,
		ExternalID:            partnerReferenceNo,
		QRCode:                midtransResp.QRString,
		QRCodeURL:             midtransResp.QrURL,
		QRCodeImage:           midtransResp.QrImage,
		ExpiresAt:             &expiresAt,
	}

	// Save to database
	if err := s.transactionRepo.Create(transaction); err != nil {
		s.log.Error("Failed to save payment transaction", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, response.NewInternalServerError("Failed to save payment transaction")
	}

	// Return response
	return &dto.QRISPaymentResponse{
		ID:                 transaction.ID,
		TransactionID:      transaction.ProviderTransactionID,
		PartnerReferenceNo: transaction.PartnerReferenceNo,
		QRCode:             transaction.QRCode,
		QRCodeURL:          transaction.QRCodeURL,
		QRCodeImage:        transaction.QRCodeImage,
		Amount:             req.Amount,
		Currency:           req.Currency,
		Status:             transaction.Status,
		ExpiresAt:          transaction.ExpiresAt,
		CreatedAt:          transaction.CreatedAt,
	}, nil
}

// CheckPaymentStatus checks the status of a payment transaction
func (s *paymentTransactionService) CheckPaymentStatus(
	ctx context.Context,
	organizationID, transactionID uuid.UUID,
) (*dto.CheckPaymentStatusResponse, error) {
	// Get transaction from database
	transaction, err := s.transactionRepo.FindByID(transactionID)
	if err != nil {
		s.log.Error("Failed to find payment transaction", zap.Error(err), zap.String("transaction_id", transactionID.String()))
		return nil, response.NewNotFound("Payment transaction not found")
	}

	// Verify organization ownership
	if transaction.OrganizationID != organizationID {
		return nil, response.NewForbidden("You do not have permission to access this transaction")
	}

	// Return current status from database (updated by webhook, no need to call Midtrans API)
	return s.mapTransactionToStatusResponse(transaction), nil
}

// GetOrganizationID retrieves the organization ID for a given user
func (s *paymentTransactionService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	org, err := s.organizationRepo.GetByUserID(userID)
	if err != nil {
		s.log.Error("Failed to get organization by user ID", zap.Error(err), zap.String("user_id", userID.String()))
		return uuid.Nil, response.NewInternalServerError("Failed to retrieve organization information")
	}
	return org.ID, nil
}

// mapTransactionToStatusResponse maps a payment transaction to status response DTO
func (s *paymentTransactionService) mapTransactionToStatusResponse(transaction *paymentEntity.PaymentTransaction) *dto.CheckPaymentStatusResponse {
	amountInUSD := float64(transaction.Amount) / 100.0 // Convert cents to USD

	return &dto.CheckPaymentStatusResponse{
		ID:                      transaction.ID,
		TransactionID:           transaction.ProviderTransactionID,
		Status:                  transaction.Status,
		Amount:                  amountInUSD,
		Currency:                transaction.Currency,
		PaidAt:                  transaction.PaidAt,
		LatestTransactionStatus: "", // Will be populated from provider response if available
		TransactionStatusDesc:   "", // Will be populated from provider response if available
	}
}

// ProcessMidtransWebhook processes incoming webhook notifications from Midtrans
func (s *paymentTransactionService) ProcessMidtransWebhook(
	ctx context.Context,
	webhookPayload map[string]interface{},
	requestBody, signature, timestamp, endpoint string,
) error {
	// Validate webhook signature first
	if signature != "" && timestamp != "" && endpoint != "" {
		isValid := s.midtransService.ValidateWebhookSignatureBISnap("POST", endpoint, requestBody, timestamp, signature)
		if !isValid {
			s.log.Error("Webhook signature validation failed",
				zap.String("endpoint", endpoint),
				zap.String("timestamp", timestamp),
			)
			return response.NewUnauthorized("Invalid webhook signature")
		}
		s.log.Info("Webhook signature validated successfully")
	} else {
		s.log.Warn("Webhook signature validation skipped - missing signature, timestamp, or endpoint",
			zap.Bool("has_signature", signature != ""),
			zap.Bool("has_timestamp", timestamp != ""),
			zap.Bool("has_endpoint", endpoint != ""),
		)
	}
	// Extract partner reference number from webhook payload
	partnerRefNo, ok := webhookPayload["partnerReferenceNo"].(string)
	if !ok {
		// Try alternative field names
		if originalPartnerRefNo, ok := webhookPayload["originalPartnerReferenceNo"].(string); ok {
			partnerRefNo = originalPartnerRefNo
		} else {
			s.log.Error("Failed to extract partner reference number from webhook", zap.Any("payload", webhookPayload))
			return response.NewBadRequest("Missing partner reference number in webhook payload")
		}
	}

	// Find transaction by partner reference number
	transaction, err := s.transactionRepo.FindByPartnerReferenceNo(partnerRefNo)
	if err != nil {
		s.log.Error("Failed to find payment transaction for webhook", zap.Error(err), zap.String("partner_reference_no", partnerRefNo))
		return response.NewNotFound("Payment transaction not found")
	}

	// Extract transaction status from webhook
	// Midtrans BI-SNAP webhook status codes: "00" = success, "03" = pending, "04" = refunded, etc.
	latestStatus, ok := webhookPayload["latestTransactionStatus"].(string)
	if !ok {
		// Try alternative field names
		if transactionStatus, ok := webhookPayload["transactionStatus"].(string); ok {
			latestStatus = transactionStatus
		} else {
			s.log.Warn("Failed to extract transaction status from webhook", zap.Any("payload", webhookPayload))
			return response.NewBadRequest("Missing transaction status in webhook payload")
		}
	}

	// Map Midtrans status to our status
	var newStatus string
	var paidAt *time.Time

	switch latestStatus {
	case "00": // Success
		newStatus = "paid"
		// Extract paid time if available
		if paidTimeStr, ok := webhookPayload["paidTime"].(string); ok && paidTimeStr != "" {
			paidTime, err := time.Parse(time.RFC3339, paidTimeStr)
			if err == nil {
				paidAt = &paidTime
			}
		} else {
			now := time.Now()
			paidAt = &now
		}
	case "03": // Pending
		newStatus = "pending"
	case "04": // Refunded
		newStatus = "failed"
	default:
		// For other statuses, keep as pending or map appropriately
		s.log.Warn("Unknown Midtrans status code", zap.String("status", latestStatus), zap.String("partner_reference_no", partnerRefNo))
		newStatus = "pending"
	}

	// Update transaction status
	if err := s.transactionRepo.UpdateStatus(transaction.ID, newStatus, paidAt); err != nil {
		s.log.Error("Failed to update payment transaction status from webhook", zap.Error(err), zap.String("transaction_id", transaction.ID.String()))
		return response.NewInternalServerError("Failed to update payment transaction status")
	}

	s.log.Info("Payment transaction status updated from webhook",
		zap.String("transaction_id", transaction.ID.String()),
		zap.String("old_status", transaction.Status),
		zap.String("new_status", newStatus),
		zap.String("partner_reference_no", partnerRefNo),
	)

	// If payment is successful and type is subscription, upgrade organization tier
	if newStatus == "paid" && transaction.Type == "subscription" {
		if err := s.handleSubscriptionUpgrade(ctx, transaction); err != nil {
			s.log.Error("Failed to handle subscription upgrade", zap.Error(err), zap.String("transaction_id", transaction.ID.String()))
			// Don't fail the webhook, just log the error
		}
	}

	return nil
}

// ProcessQRISWebhook processes incoming QRIS webhook notifications from Midtrans
func (s *paymentTransactionService) ProcessQRISWebhook(
	ctx context.Context,
	payload []byte,
	headers map[string]string,
) error {
	// Use Midtrans HandleQRISWebhook to validate and parse webhook
	webhookResult, err := s.midtransService.HandleQRISWebhook(ctx, payload, headers)
	if err != nil {
		s.log.Error("Failed to process QRIS webhook", zap.Error(err))
		return response.NewBadRequest(fmt.Sprintf("Failed to process webhook: %v", err))
	}

	// Find transaction by partner reference number (PaymentCode)
	// Since webhook is public, we need to search across all tenant schemas
	partnerRefNo := webhookResult.PaymentCode

	// Query all tenants to find which schema contains this transaction
	sqlDB, err := s.db.DB()
	if err != nil {
		s.log.Error("Failed to get sql.DB", zap.Error(err))
		return response.NewInternalServerError("Failed to process webhook")
	}

	// Get all active tenants
	rows, err := sqlDB.QueryContext(ctx, "SELECT organization_id, schema_name FROM tenants WHERE deleted_at IS NULL")
	if err != nil {
		s.log.Error("Failed to query tenants", zap.Error(err))
		return response.NewInternalServerError("Failed to process webhook")
	}
	defer rows.Close()

	var transaction *paymentEntity.PaymentTransaction
	var foundSchemaName string

	// Search each tenant schema for the transaction
	for rows.Next() {
		var orgID uuid.UUID
		var schemaName string
		if err := rows.Scan(&orgID, &schemaName); err != nil {
			continue
		}

		// For new tenants, tables are created directly on tenant creation
		// Migrations are only needed for existing tenants or future schema changes
		// Since we're not in production yet, we can skip lazy migration for now

		// Set search_path to this tenant schema
		_, err = sqlDB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName))
		if err != nil {
			continue
		}

		// Try to find transaction in this schema
		var tx paymentEntity.PaymentTransaction
		err = s.db.Where("partner_reference_no = ?", partnerRefNo).First(&tx).Error
		if err == nil {
			transaction = &tx
			foundSchemaName = schemaName
			break
		}
	}

	if transaction == nil {
		s.log.Error("Failed to find payment transaction for webhook", zap.String("partner_reference_no", partnerRefNo))
		return response.NewNotFound("Payment transaction not found")
	}

	// Set tenant schema for subsequent operations
	if err := s.tenantService.SetSchemaForRequest(ctx, foundSchemaName); err != nil {
		s.log.Error("Failed to set tenant schema", zap.Error(err), zap.String("schema_name", foundSchemaName))
		return response.NewInternalServerError("Failed to process webhook")
	}

	// Map midtrans.PaymentStatus to our transaction status string
	var newStatus string
	var paidAt *time.Time

	switch webhookResult.Status {
	case midtrans.PaymentStatusCompleted:
		newStatus = "paid"
		now := time.Now()
		paidAt = &now
	case midtrans.PaymentStatusPending:
		newStatus = "pending"
	case midtrans.PaymentStatusFailed:
		newStatus = "failed"
	case midtrans.PaymentStatusCancelled:
		newStatus = "cancelled"
	case midtrans.PaymentStatusRefunded:
		newStatus = "failed" // We treat refunded as failed
	default:
		newStatus = "pending"
	}

	// Update transaction status
	if err := s.transactionRepo.UpdateStatus(transaction.ID, newStatus, paidAt); err != nil {
		s.log.Error("Failed to update payment transaction status from webhook", zap.Error(err), zap.String("transaction_id", transaction.ID.String()))
		return response.NewInternalServerError("Failed to update payment transaction status")
	}

	s.log.Info("Payment transaction status updated from webhook",
		zap.String("transaction_id", transaction.ID.String()),
		zap.String("old_status", transaction.Status),
		zap.String("new_status", newStatus),
		zap.String("payment_code", webhookResult.PaymentCode),
		zap.String("gateway_transaction_id", webhookResult.GatewayTransactionID),
	)

	// If payment is successful and type is subscription, upgrade organization tier
	if newStatus == "paid" && transaction.Type == "subscription" {
		if err := s.handleSubscriptionUpgrade(ctx, transaction); err != nil {
			s.log.Error("Failed to handle subscription upgrade", zap.Error(err), zap.String("transaction_id", transaction.ID.String()))
			// Don't fail the webhook, just log the error
		}
	}

	return nil
}

// handleSubscriptionUpgrade upgrades organization subscription tier and sends notifications
func (s *paymentTransactionService) handleSubscriptionUpgrade(ctx context.Context, transaction *paymentEntity.PaymentTransaction) error {
	// Get organization
	org, err := s.organizationRepo.GetByID(transaction.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Only upgrade if currently on free tier
	if org.SubscriptionTier != "free" {
		s.log.Info("Organization already on paid tier, skipping upgrade",
			zap.String("organization_id", org.ID.String()),
			zap.String("current_tier", org.SubscriptionTier),
		)
		return nil
	}

	// Upgrade to paid tier
	org.SubscriptionTier = "paid"
	if err := s.organizationRepo.Update(org); err != nil {
		return fmt.Errorf("failed to update organization subscription tier: %w", err)
	}

	s.log.Info("Organization subscription upgraded to paid",
		zap.String("organization_id", org.ID.String()),
		zap.String("transaction_id", transaction.ID.String()),
	)

	// Get organization members (owners and admins) to send notifications
	members, err := s.organizationRepo.ListMembers(org.ID)
	if err != nil {
		s.log.Warn("Failed to get organization members for notification", zap.Error(err))
		// Continue without sending emails
		return nil
	}

	// Send email notifications to owners and admins
	amountInUSD := float64(transaction.Amount) / 100.0
	for _, member := range members {
		// Only notify owners and admins
		if member.Role != "owner" && member.Role != "admin" {
			continue
		}

		user, err := s.userRepo.GetByID(member.UserID)
		if err != nil {
			s.log.Warn("Failed to get user for payment notification", zap.Error(err), zap.String("user_id", member.UserID.String()))
			continue
		}

		// Send payment success email
		if err := s.sendPaymentSuccessEmail(user.Email, user.FullName, org.Name, amountInUSD, transaction.Currency); err != nil {
			s.log.Warn("Failed to send payment success email", zap.Error(err), zap.String("email", user.Email))
			// Continue with other members
		}
	}

	return nil
}

// sendPaymentSuccessEmail sends an email notification when payment is successful
func (s *paymentTransactionService) sendPaymentSuccessEmail(to, userName, orgName string, amount float64, currency string) error {
	subject := "Payment Successful - Subscription Upgraded"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Payment Successful</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<h2 style="color: #27ae60;">Payment Successful!</h2>
		<p>Hello %s,</p>
		<p>Your payment of <strong>%s %.2f</strong> has been successfully processed.</p>
		<p>Your organization <strong>%s</strong> has been upgraded to the <strong>Paid</strong> subscription tier.</p>
		
		<h3 style="color: #2c3e50; margin-top: 30px;">What's Next?</h3>
		<p>You now have access to:</p>
		<ul style="line-height: 2;">
			<li>Unlimited patients</li>
			<li>Unlimited clinicians</li>
			<li>All premium features</li>
		</ul>

		<div style="text-align: center; margin: 30px 0;">
			<a href="%s/dashboard" style="background-color: #3498db; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Go to Dashboard</a>
		</div>

		<p style="color: #95a5a6; font-size: 12px; margin-top: 30px; border-top: 1px solid #ecf0f1; padding-top: 20px;">
			If you have any questions about your subscription, please contact our support team.
		</p>
		<p style="color: #95a5a6; font-size: 12px;">
			Thank you for choosing OpenMind Practice!<br>
			The OpenMind Team
		</p>
	</div>
</body>
</html>
`, userName, currency, amount, orgName, s.appURL)

	textBody := fmt.Sprintf(`
Payment Successful!

Hello %s,

Your payment of %s %.2f has been successfully processed.

Your organization %s has been upgraded to the Paid subscription tier.

What's Next?
- Unlimited patients
- Unlimited clinicians
- All premium features

Go to Dashboard: %s/dashboard

If you have any questions about your subscription, please contact our support team.

Thank you for choosing OpenMind Practice!
The OpenMind Team
`, userName, currency, amount, orgName, s.appURL)

	if err := s.emailService.SendEmail(to, subject, htmlBody); err != nil {
		s.log.Warn("Failed to send HTML payment success email, trying plain text", zap.Error(err))
		return s.emailService.SendEmail(to, subject, textBody)
	}

	return nil
}
