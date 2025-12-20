package midtrans

// AccessTokenRequest for BI-SNAP authentication
type AccessTokenRequest struct {
	GrantType string `json:"grantType"`
}

// AccessTokenResponse from BI-SNAP
type AccessTokenResponse struct {
	ResponseCode    string   `json:"responseCode"`
	ResponseMessage string   `json:"responseMessage"`
	AccessToken     string   `json:"accessToken"`
	TokenType       string   `json:"tokenType"`
	ExpiresIn       string   `json:"expiresIn"`
	AdditionalInfo  struct{} `json:"additionalInfo"`
}

// PaymentRequest represents a payment request to Midtrans
type PaymentRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	CustomerDetails    CustomerDetails    `json:"customer_details"`
	ItemDetails        []ItemDetail       `json:"item_details"`
	PaymentType        string             `json:"payment_type"`
	QRIS               *QRISDetails       `json:"qris,omitempty"`
}

// QRISPaymentRequest represents the request for creating QRIS payment using BI-SNAP
type QRISPaymentRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo,omitempty"`
	Amount             struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	MerchantID     string `json:"merchantId,omitempty"`
	ValidityPeriod string `json:"validityPeriod,omitempty"`
	AdditionalInfo struct {
		Acquirer string `json:"acquirer,omitempty"`
	} `json:"additionalInfo,omitempty"`
}

// TransactionDetails contains transaction information
type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

// CustomerDetails contains customer information
type CustomerDetails struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email,omitempty"`
}

// ItemDetail represents an item in the transaction
type ItemDetail struct {
	ID       string `json:"id"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

// QRISDetails contains QRIS-specific configuration
type QRISDetails struct {
	Acquirer string `json:"acquirer"`
}

// PaymentResponse represents Midtrans payment response
type PaymentResponse struct {
	StatusCode        string                 `json:"status_code"`
	StatusMessage     string                 `json:"status_message"`
	TransactionID     string                 `json:"transaction_id"`
	OrderID           string                 `json:"order_id"`
	GrossAmount       string                 `json:"gross_amount"`
	PaymentType       string                 `json:"payment_type"`
	TransactionTime   string                 `json:"transaction_time"`
	TransactionStatus string                 `json:"transaction_status"`
	FraudStatus       string                 `json:"fraud_status"`
	Actions           []PaymentAction        `json:"actions"`
	QRString          string                 `json:"qr_string,omitempty"`
	QrURL             string                 `json:"qr_url,omitempty"`
	QrImage           string                 `json:"qr_image,omitempty"`
	Acquirer          string                 `json:"acquirer,omitempty"`
	RawResponse       map[string]interface{} `json:"-"`
}

// QRISPaymentResponse represents the response from BI-SNAP QRIS payment
type QRISPaymentResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	ReferenceNo        string `json:"referenceNo,omitempty"`
	PartnerReferenceNo string `json:"partnerReferenceNo,omitempty"`
	QrContent          string `json:"qrContent"`
	QrURL              string `json:"qrURL"`
	QrImage            string `json:"qrImage"`
	AdditionalInfo     struct {
		Acquirer interface{} `json:"acquirer,omitempty"`
	} `json:"additionalInfo,omitempty"`
}

// TransactionStatusResponse represents the response from BI-SNAP transaction status check
type TransactionStatusResponse struct {
	ResponseCode               string `json:"responseCode"`
	ResponseMessage            string `json:"responseMessage"`
	OriginalExternalID         string `json:"originalExternalId"`
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string `json:"originalReferenceNo"`
	ServiceCode                string `json:"serviceCode"`
	LatestTransactionStatus    string `json:"latestTransactionStatus"`
	TransactionStatusDesc      string `json:"transactionStatusDesc"`
	PaidTime                   string `json:"paidTime"`
	Amount                     struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	TerminalID     string `json:"terminalId"`
	AdditionalInfo struct {
		RefundHistory []struct {
			RefundStatus       string `json:"refundStatus"`
			Reason             string `json:"reason"`
			RefundNo           string `json:"refundNo"`
			PartnerReferenceNo string `json:"partnerReferenceNo"`
			RefundDate         string `json:"refundDate"`
			RefundAmount       struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"refundAmount"`
		} `json:"refundHistory"`
	} `json:"additionalInfo"`
}

// PaymentAction represents available payment actions
type PaymentAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}
