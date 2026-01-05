package payment

import (
	"context"
	"fmt"

	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/midtrans"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// midtransProvider implements PaymentProvider for Midtrans QRIS
// Note: Midtrans QRIS is transaction-based, not payment method-based like Stripe cards
// This provider adapts QRIS to work with the PaymentProvider interface
type midtransProvider struct {
	service *midtrans.Service
	log     logger.Logger
}

// NewMidtransProvider creates a new Midtrans payment provider
func NewMidtransProvider(cfg *config.MidtransConfig, log logger.Logger) (PaymentProvider, error) {
	service, err := midtrans.NewMidtransService(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Midtrans service: %w", err)
	}

	return &midtransProvider{
		service: service,
		log:     log,
	}, nil
}

// CreatePaymentMethod creates a QRIS payment transaction
// For Midtrans QRIS, the "token" parameter is actually a payment code/order ID
// and the amount should be passed separately. This is a limitation of the interface.
// Returns the transaction ID as the payment method ID
func (p *midtransProvider) CreatePaymentMethod(ctx context.Context, token string) (string, error) {
	// For QRIS, we need amount and payment code
	// Since the interface only provides token, we'll need to parse it or use a different approach
	// For now, we'll create a QRIS payment with a default amount
	// TODO: Extend the interface to support amount for QRIS payments

	p.log.Warn(
		"CreatePaymentMethod called for Midtrans QRIS - QRIS requires amount which is not available in the interface",
		zap.String("token", token),
	)

	// Return an error indicating that QRIS payments need amount
	return "", fmt.Errorf("Midtrans QRIS payments require amount - use CreateQRISPayment directly")
}

// DeletePaymentMethod is not applicable for QRIS transactions
// QRIS transactions are one-time and cannot be "deleted" like payment methods
func (p *midtransProvider) DeletePaymentMethod(ctx context.Context, paymentMethodID string) error {
	p.log.Warn("DeletePaymentMethod called for Midtrans QRIS - not applicable for QRIS transactions",
		zap.String("payment_method_id", paymentMethodID))
	return fmt.Errorf("QRIS transactions cannot be deleted - they are one-time payment transactions")
}

// ListPaymentMethods is not applicable for QRIS
// QRIS transactions are not stored payment methods
func (p *midtransProvider) ListPaymentMethods(ctx context.Context) ([]PaymentMethodInfo, error) {
	p.log.Warn("ListPaymentMethods called for Midtrans QRIS - not applicable for QRIS transactions")
	return nil, fmt.Errorf("QRIS does not support listing payment methods - use transaction status check instead")
}

// GetPaymentMethod retrieves QRIS transaction status
// The paymentMethodID is the transaction ID
func (p *midtransProvider) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethodInfo, error) {
	// Check transaction status
	status, err := p.service.CheckTransactionStatus(ctx, paymentMethodID)
	if err != nil {
		p.log.Error(
			"Failed to check QRIS transaction status",
			zap.Error(err),
			zap.String("transaction_id", paymentMethodID),
		)
		return nil, fmt.Errorf("failed to get transaction status: %w", err)
	}

	// Convert transaction status to PaymentMethodInfo
	// Note: This is a workaround since QRIS doesn't have card details
	info := &PaymentMethodInfo{
		ID:          paymentMethodID,
		Last4:       "QRIS", // Placeholder
		Brand:       "QRIS",
		ExpiryMonth: 0,
		ExpiryYear:  0,
	}

	// Log transaction status
	p.log.Info("QRIS transaction status retrieved",
		zap.String("transaction_id", paymentMethodID),
		zap.String("status", status.LatestTransactionStatus))

	return info, nil
}

// CreateQRISPayment is a helper method to create QRIS payments with amount
// This should be called directly from the service layer, not through the PaymentProvider interface
func (p *midtransProvider) CreateQRISPayment(
	ctx context.Context,
	paymentCode string,
	amount decimal.Decimal,
) (*midtrans.PaymentResponse, error) {
	return p.service.CreateQRISPayment(ctx, paymentCode, amount)
}

// CheckTransactionStatus is a helper method to check QRIS transaction status
func (p *midtransProvider) CheckTransactionStatus(
	ctx context.Context,
	transactionID string,
) (*midtrans.TransactionStatusResponse, error) {
	return p.service.CheckTransactionStatus(ctx, transactionID)
}
