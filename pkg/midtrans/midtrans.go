package midtrans

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	ContextTimeout        = 30 * time.Second
	MaxRetries            = 3
	Buffer                = 300
	PaymentDecimalPlaces  = 2
	PaymentValidityPeriod = 15 * time.Minute
	// TimestampFormatMilliseconds is the timestamp format used for Midtrans API requests with milliseconds
	TimestampFormatMilliseconds = "2006-01-02T15:04:05.000Z"
)

const (
	HeaderAuthorization = "Authorization"
	BearerPrefix        = "Bearer "
	HeaderXTimestamp    = "X-TIMESTAMP"
	HeaderXSignature    = "X-SIGNATURE"
	HeaderXClientKey    = "X-CLIENT-KEY"
	HeaderXPartnerID    = "X-PARTNER-ID"
	HeaderXExternalID   = "X-EXTERNAL-ID"
	HeaderXDeviceID     = "X-DEVICE-ID"
	HeaderChannelID     = "CHANNEL-ID"
	HeaderContentType   = "Content-Type"
	// ErrorMessageFormat is the format string for error messages with response code
	ErrorMessageFormat = "%s (Code: %s)"
	// ErrFailedToGetAccessToken is the error message format when access token retrieval fails
	ErrFailedToGetAccessToken = "failed to get access token: %w"
)

// SimpleCache is a simple in-memory cache for access tokens
type SimpleCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      string
	expiration time.Time
}

func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		items: make(map[string]cacheItem),
	}
}

func (c *SimpleCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.items[key]
	if !exists {
		return "", false
	}
	if time.Now().After(item.expiration) {
		delete(c.items, key)
		return "", false
	}
	return item.value, true
}

func (c *SimpleCache) Set(key string, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	return nil
}

type Service struct {
	midtransConf *config.MidtransConfig
	cache        *SimpleCache
	logger       logger.Logger
	client       *fasthttp.Client
	baseURL      string
	accessToken  string
	tokenExpiry  time.Time
	mu           sync.RWMutex
}

func NewMidtransService(
	midtransConf *config.MidtransConfig,
	log logger.Logger,
) (*Service, error) {
	decodedConf := decodeBase64Keys(midtransConf, log)

	baseURL := "https://api.sandbox.midtrans.com/v2"
	if decodedConf.IsProduction {
		baseURL = "https://api.midtrans.com/v2"
	}

	return &Service{
		midtransConf: decodedConf,
		cache:        NewSimpleCache(),
		logger:       log,
		client: &fasthttp.Client{
			ReadTimeout:  ContextTimeout,
			WriteTimeout: ContextTimeout,
		},
		baseURL: baseURL,
	}, nil
}

// decodeBase64Keys decodes base64-encoded keys in the config if they are base64 encoded.
// It creates a copy of the config with decoded keys to avoid modifying the original.
// Keys stored as base64 in config (for easier environment variable storage) are decoded to PEM format.
func decodeBase64Keys(conf *config.MidtransConfig, log logger.Logger) *config.MidtransConfig {
	decodedConf := *conf

	if decodedConf.BISnapPublicKey != "" {
		decoded, err := decodeIfBase64(decodedConf.BISnapPublicKey)
		if err == nil {
			decodedConf.BISnapPublicKey = decoded
		}
	}

	if decodedConf.MerchantPrivateKey != "" {
		decoded, err := decodeIfBase64(decodedConf.MerchantPrivateKey)
		if err == nil {
			decodedConf.MerchantPrivateKey = decoded
		}
	}

	if decodedConf.MerchantPublicKey != "" {
		decoded, err := decodeIfBase64(decodedConf.MerchantPublicKey)
		if err == nil {
			decodedConf.MerchantPublicKey = decoded
		}
	}

	if decodedConf.NotificationPublicKey != "" {
		decoded, err := decodeIfBase64(decodedConf.NotificationPublicKey)
		if err == nil {
			decodedConf.NotificationPublicKey = decoded
		}
	}

	return &decodedConf
}

// decodeIfBase64 attempts to decode a string from base64.
// It handles both single-line and multiline base64 strings.
// If the string is not valid base64, it returns the original string (assuming it's already PEM format).
func decodeIfBase64(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return s, nil
	}

	// Remove all whitespace/newlines for decoding
	cleaned := strings.ReplaceAll(s, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	decoded, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return s, fmt.Errorf("not base64 encoded: %w", err)
	}

	decodedStr := string(decoded)
	if strings.Contains(decodedStr, "BEGIN") && strings.Contains(decodedStr, "END") {
		return decodedStr, nil
	}

	return decodedStr, nil
}

func (m *Service) getBISnapBaseURL() string {
	if m.midtransConf.IsProduction {
		return "https://merchants.midtrans.com"
	}
	return "https://merchants.sbx.midtrans.com"
}

// func (m *Service) getBISnapAuthURL() string {
//	if m.midtransConf.IsProduction {
//		return "https://merchants-app.midtrans.com"
//	}
//	return "https://merchants-app.sbx.midtrans.com"
//}

// minifyJSON removes unnecessary whitespace from JSON string
func minifyJSON(jsonStr string) string {
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, []byte(jsonStr)); err != nil {
		return jsonStr
	}
	return buffer.String()
}

// generateSignature generates HMAC-SHA512 signature for BI-SNAP API
// Formula: HTTPMethod + ":" + EndpointUrl + ":" + AccessToken + ":" + Lowercase(HexEncode(SHA-256(minify(RequestBody)))) + ":" + TimeStamp
func (m *Service) generateSignature(method, endpoint, accessToken, requestBody, timestamp string) string {
	minifiedBody := minifyJSON(requestBody)

	hash := sha256.Sum256([]byte(minifiedBody))
	requestBodyHash := strings.ToLower(hex.EncodeToString(hash[:]))

	stringToSign := fmt.Sprintf("%s:%s:%s:%s:%s", method, endpoint, accessToken, requestBodyHash, timestamp)

	h := hmac.New(sha512.New, []byte(m.midtransConf.BISnapClientSecret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

func (m *Service) getLegacyBasicAuth() string {
	auth := base64.StdEncoding.EncodeToString(
		[]byte(m.midtransConf.ClientKey + ":" + m.midtransConf.ServerKey),
	)
	return "Basic " + auth
}

// generateAccessTokenSignature creates SHA256withRSA signature for B2B Access Token API
func (m *Service) generateAccessTokenSignature(clientID, timestamp string) (string, error) {
	privateKeyStr := m.midtransConf.MerchantPrivateKey

	if privateKeyStr == "" {
		return "", fmt.Errorf("merchant private key is empty")
	}

	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA")
	}

	stringToSign := fmt.Sprintf("%s|%s", clientID, timestamp)

	hash := sha256.Sum256([]byte(stringToSign))

	// Midtrans BI-SNAP API specifies "SHA256withRSA" which uses RSASSA-PKCS1-v1_5 (PKCS1v15 padding)
	// In standard cryptographic naming, "SHA256withRSA" refers to PKCS1v15, not RSA-PSS.
	// RSA-PSS would be named "SHA256withRSAandMGF1" or "SHA256withRSA/PSS".
	// See: https://docs.oracle.com/javase/12/docs/specs/security/standard-names.html
	// Midtrans docs: https://docs.midtrans.com/reference/signature-generation#asymmetric-signature-sha256withrsa

	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// getAccessToken obtains access token for BI-SNAP API with Redis caching and retry logic
func (m *Service) getAccessToken(ctx context.Context) (string, error) {
	return m.getAccessTokenWithRetry(ctx, MaxRetries)
}

// GetAccessToken is the public method to get access token for external use
func (m *Service) GetAccessToken(ctx context.Context) (string, error) {
	return m.getAccessToken(ctx)
}

// getAccessTokenWithRetry implements retry logic for access token retrieval
func (m *Service) getAccessTokenWithRetry(ctx context.Context, maxRetries int) (string, error) {
	cachedToken, found := m.getCachedAccessToken(ctx)
	if found {
		return cachedToken, nil
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		m.logger.Debug("Retrieving access token from API", zap.Int("attempt", attempt))

		token, expiresIn, err := m.fetchAccessTokenFromAPI(ctx)
		if err == nil {
			m.saveToken(ctx, token, expiresIn)
			return token, nil
		}

		lastErr = err
		m.logger.Warn("Failed to get access token", zap.Error(err), zap.Int("attempt", attempt))

		if attempt < maxRetries && m.isRetryableError(err) {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	return "", fmt.Errorf("failed to get access token after %d attempts: %w", maxRetries, lastErr)
}

func (m *Service) getCachedAccessToken(ctx context.Context) (string, bool) {
	cacheKey := fmt.Sprintf("midtrans:access_token:%s", m.midtransConf.BISnapClientID)
	cachedToken, found := m.cache.Get(cacheKey)
	if found {
		m.logger.Debug("Using cached access token")
		return cachedToken, true
	}

	m.logger.Debug("No valid cached token found")
	return "", false
}

func (m *Service) saveToken(ctx context.Context, token string, expiresIn int) {
	m.mu.Lock()
	m.accessToken = token
	m.tokenExpiry = time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)
	m.mu.Unlock()

	cacheKey := fmt.Sprintf("midtrans:access_token:%s", m.midtransConf.BISnapClientID)
	cacheExpiry := time.Duration(expiresIn-Buffer) * time.Second
	if cacheExpiry > 0 {
		if err := m.cache.Set(cacheKey, token, cacheExpiry); err != nil {
			m.logger.Error("Failed to cache access token", zap.Error(err))
		}
	}
}

// fetchAccessTokenFromAPI fetches a new access token from the Midtrans API
func (m *Service) fetchAccessTokenFromAPI(ctx context.Context) (string, int, error) {
	tokenRequest := AccessTokenRequest{
		GrantType: "client_credentials",
	}

	jsonData, err := sonic.Marshal(tokenRequest)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal token request: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	signature, err := m.generateAccessTokenSignature(m.midtransConf.BISnapClientID, timestamp)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate access token signature: %w", err)
	}

	var tokenResp AccessTokenResponse
	opts := requestOptions{
		ctx:    ctx,
		method: "POST",
		url:    fmt.Sprintf("%s/v1.0/access-token/b2b", m.getBISnapBaseURL()),
		headers: map[string]string{
			HeaderXTimestamp: timestamp,
			HeaderXSignature: signature,
			HeaderXClientKey: m.midtransConf.BISnapClientID,
		},
		body:         jsonData,
		responseDest: &tokenResp,
	}

	m.logger.Debug("Access token request headers:")
	m.logger.Debug(HeaderXTimestamp, zap.String(HeaderXTimestamp, timestamp))
	m.logger.Debug(HeaderXSignature, zap.String(HeaderXSignature, signature))
	m.logger.Debug(HeaderXClientKey, zap.String(HeaderXClientKey, m.midtransConf.BISnapClientID))

	resp, err := m.doRequest(opts)
	if resp != nil {
		defer fasthttp.ReleaseResponse(resp)
	}
	if err != nil {
		return "", 0, err
	}

	if !IsSuccessCode(tokenResp.ResponseCode) {
		userMessage := FormatErrorMessage(tokenResp.ResponseCode)
		return "", 0, fmt.Errorf(ErrorMessageFormat, userMessage, tokenResp.ResponseCode)
	}

	expiresIn, _ := strconv.Atoi(tokenResp.ExpiresIn)
	if expiresIn == 0 {
		expiresIn = 3600
	}

	return tokenResp.AccessToken, expiresIn, nil
}

// isRetryableError determines if an error is retryable
func (m *Service) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()

	// Network errors are retryable
	if strings.Contains(errorStr, "connection") ||
		strings.Contains(errorStr, "timeout") ||
		strings.Contains(errorStr, "network") {
		return true
	}

	// Server errors (5xx) are retryable
	if strings.Contains(errorStr, "500") ||
		strings.Contains(errorStr, "502") ||
		strings.Contains(errorStr, "503") ||
		strings.Contains(errorStr, "504") {
		return true
	}

	// Rate limiting is retryable
	if strings.Contains(errorStr, "429") ||
		strings.Contains(errorStr, "rate limit") {
		return true
	}

	return false
}

// CreateQRISPayment creates a QRIS payment transaction using BI-SNAP API
func (m *Service) CreateQRISPayment(
	ctx context.Context,
	paymentCode string,
	amount decimal.Decimal,
) (*PaymentResponse, error) {
	accessToken, err := m.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf(ErrFailedToGetAccessToken, err)
	}

	paymentRequest := QRISPaymentRequest{
		PartnerReferenceNo: paymentCode,
		Amount: struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		}{
			Value:    amount.StringFixed(PaymentDecimalPlaces),
			Currency: "IDR",
		},
		MerchantID: m.midtransConf.BISnapPartnerID,
		ValidityPeriod: time.Now().
			Add(PaymentValidityPeriod).
			UTC().
			Format("2006-01-02T15:04:05-07:00"),
		// ISO 8601 format
		AdditionalInfo: struct {
			Acquirer string `json:"acquirer,omitempty"`
		}{
			Acquirer: "GOPAY",
		},
	}

	jsonData, err := sonic.Marshal(paymentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	timestamp := time.Now().UTC().Format(TimestampFormatMilliseconds)

	endpoint := "/v1.0/qr/qr-mpm-generate"
	signature := m.generateSignature("POST", endpoint, accessToken, string(jsonData), timestamp)

	var qrisResp QRISPaymentResponse
	opts := requestOptions{
		ctx:    ctx,
		method: "POST",
		url:    fmt.Sprintf("%s%s", m.getBISnapBaseURL(), endpoint),
		headers: map[string]string{
			HeaderAuthorization: BearerPrefix + accessToken,
			HeaderXPartnerID:    m.midtransConf.BISnapPartnerID,
			HeaderXExternalID:   paymentCode,
			HeaderXTimestamp:    timestamp,
			HeaderXSignature:    signature,
			HeaderXDeviceID:     "web-friends-v1.0",
			HeaderChannelID:     generateNumericHash(),
		},
		body:         jsonData,
		responseDest: &qrisResp,
	}

	resp, err := m.doRequest(opts)
	if resp != nil {
		defer fasthttp.ReleaseResponse(resp)
	}
	if err != nil {
		return nil, err
	}

	if !IsSuccessCode(qrisResp.ResponseCode) {
		userMessage := FormatErrorMessage(qrisResp.ResponseCode)
		return nil, fmt.Errorf(ErrorMessageFormat, userMessage, qrisResp.ResponseCode)
	}

	return &PaymentResponse{
		TransactionID:     qrisResp.ReferenceNo,
		OrderID:           paymentCode,
		QRString:          qrisResp.QrContent,
		QrURL:             qrisResp.QrURL,
		QrImage:           qrisResp.QrImage,
		TransactionStatus: "pending",
		GrossAmount:       amount.StringFixed(0),
		PaymentType:       "qris",
		TransactionTime:   time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// makePaymentRequest sends payment request to Midtrans
// func (m *Service) makePaymentRequest(ctx context.Context, request PaymentRequest) (*PaymentResponse, error) {
//	jsonData, err := sonic.Marshal(request)
//	if err != nil {
//		return nil, fmt.Errorf("failed to marshal request: %w", err)
//	}
//
//	var paymentResp PaymentResponse
//	opts := requestOptions{
//		ctx:    ctx,
//		method: "POST",
//		url:    fmt.Sprintf("%s/charge", m.baseURL),
//		headers: map[string]string{
//			"Accept":        "application/json",
//			"Authorization": m.getLegacyBasicAuth(),
//		},
//		body:         jsonData,
//		responseDest: &paymentResp,
//	}
//
//	resp, err := m.doRequest(opts)
//	if resp != nil {
//		defer fasthttp.ReleaseResponse(resp)
//	}
//	if err != nil {
//		return nil, err
//	}
//
//	if resp != nil && resp.StatusCode() >= 400 {
//		return nil, fmt.Errorf("payment request failed: %s - %s", paymentResp.StatusCode, paymentResp.StatusMessage)
//	}
//
//	return &paymentResp, nil
//}

// GetTransactionStatus retrieves transaction status from Midtrans
func (m *Service) GetTransactionStatus(ctx context.Context, orderID string) (*PaymentResponse, error) {
	var paymentResp PaymentResponse
	opts := requestOptions{
		ctx:    ctx,
		method: "GET",
		url:    fmt.Sprintf("%s/%s/status", m.baseURL, orderID),
		headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": m.getLegacyBasicAuth(),
		},
		responseDest: &paymentResp,
	}

	resp, err := m.doRequest(opts)
	if resp != nil {
		defer fasthttp.ReleaseResponse(resp)
	}
	if err != nil {
		return nil, err
	}

	if resp != nil && resp.StatusCode() >= 400 {
		return nil, fmt.Errorf(
			"failed to get transaction status: %s - %s",
			paymentResp.StatusCode,
			paymentResp.StatusMessage,
		)
	}

	return &paymentResp, nil
}

// GetPaymentStatus retrieves payment status using BI-SNAP API
func (m *Service) GetPaymentStatus(ctx context.Context, referenceNo string) (*TransactionStatusResponse, error) {
	accessToken, err := m.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf(ErrFailedToGetAccessToken, err)
	}

	externalID := uuid.New().String()

	statusRequest := map[string]interface{}{
		"originalReferenceNo": referenceNo,
	}

	jsonData, err := sonic.Marshal(statusRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status request: %w", err)
	}

	timestamp := time.Now().UTC().Format(TimestampFormatMilliseconds)

	endpoint := "/v1.0/transfer-va/status"
	signature := m.generateSignature("POST", endpoint, accessToken, string(jsonData), timestamp)

	url := fmt.Sprintf("%s%s", m.getBISnapBaseURL(), endpoint)

	var statusResp TransactionStatusResponse
	opts := requestOptions{
		ctx:    ctx,
		method: "POST",
		url:    url,
		headers: map[string]string{
			HeaderContentType:   "application/json",
			HeaderAuthorization: BearerPrefix + accessToken,
			HeaderXPartnerID:    m.midtransConf.BISnapPartnerID,
			HeaderXExternalID:   externalID,
			HeaderXTimestamp:    timestamp,
			HeaderXSignature:    signature,
			HeaderXDeviceID:     "web-closaf-v1.0",
			HeaderChannelID:     generateNumericHash(),
		},
		body:         jsonData,
		responseDest: &statusResp,
	}

	resp, err := m.doRequest(opts)
	if resp != nil {
		defer fasthttp.ReleaseResponse(resp)
	}
	if err != nil {
		return nil, err
	}

	if !IsSuccessCode(statusResp.ResponseCode) {
		userMessage := FormatErrorMessage(statusResp.ResponseCode)
		return nil, fmt.Errorf(ErrorMessageFormat, userMessage, statusResp.ResponseCode)
	}

	return &statusResp, nil
}

// CheckTransactionStatus checks the status of a QRIS transaction using BI-SNAP API
func (m *Service) CheckTransactionStatus(
	ctx context.Context,
	transactionID string,
) (*TransactionStatusResponse, error) {
	accessToken, err := m.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf(ErrFailedToGetAccessToken, err)
	}

	url := fmt.Sprintf("%s/v1.0/qr/qr-mpm-query", m.getBISnapBaseURL())
	timestamp := time.Now().UTC().Format(TimestampFormatMilliseconds)
	externalID := fmt.Sprintf("query-%d", time.Now().Unix())

	requestBody := map[string]interface{}{
		"originalPartnerReferenceNo": transactionID,
		"merchantId":                 m.midtransConf.BISnapPartnerID,
		"serviceCode":                "47", // QRIS payment service code
	}

	requestJSON, err := sonic.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	signature := m.generateSignature("POST", "/v1.0/qr/qr-mpm-query", accessToken, string(requestJSON), timestamp)

	var response TransactionStatusResponse
	opts := requestOptions{
		ctx:    ctx,
		method: "POST",
		url:    url,
		headers: map[string]string{
			HeaderAuthorization: BearerPrefix + accessToken,
			HeaderXPartnerID:    m.midtransConf.BISnapPartnerID,
			HeaderXExternalID:   externalID,
			HeaderXTimestamp:    timestamp,
			HeaderXSignature:    signature,
			HeaderXDeviceID:     "web-friends-v1.0",
			HeaderChannelID:     generateNumericHash(),
		},
		body:         requestJSON,
		responseDest: &response,
	}

	resp, err := m.doRequest(opts)
	if resp != nil {
		defer fasthttp.ReleaseResponse(resp)
	}
	if err != nil {
		return nil, err
	}

	if !IsSuccessCode(response.ResponseCode) {
		userMessage := FormatErrorMessage(response.ResponseCode)
		return nil, fmt.Errorf(ErrorMessageFormat, userMessage, response.ResponseCode)
	}

	return &response, nil
}

func (m *Service) ValidateWebhookSignatureBISnap(
	method, endpoint, requestBody, timestamp, receivedSignature string,
) bool {
	minified := minifyJSON(requestBody)

	hash := sha256.Sum256([]byte(minified))
	requestBodyHash := strings.ToLower(hex.EncodeToString(hash[:]))

	stringToSign := fmt.Sprintf("%s:%s:%s:%s", method, endpoint, requestBodyHash, timestamp)

	block, _ := pem.Decode([]byte(m.midtransConf.NotificationPublicKey))
	if block == nil {
		m.logger.Error("Failed to decode public key PEM block")
		return false
	}

	var rsaPublicKey *rsa.PublicKey
	var err error

	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		rsaPublicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return false
		}
	} else {
		var ok bool
		rsaPublicKey, ok = pubKeyInterface.(*rsa.PublicKey)
		if !ok {
			return false
		}
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(receivedSignature)
	if err != nil {
		m.logger.Error("Failed to decode signature", zap.Error(err))
		return false
	}

	hashed := sha256.Sum256([]byte(stringToSign))

	m.logger.Debug("Webhook signature validation: ",
		zap.String("string_to_sign", stringToSign),
		zap.String("hashed", fmt.Sprintf("%x", hashed)),
		zap.String("received_signature", receivedSignature),
	)

	// Midtrans BI-SNAP API specifies "SHA256withRSA" which uses RSASSA-PKCS1-v1_5 (PKCS1v15 padding)
	// In standard cryptographic naming, "SHA256withRSA" refers to PKCS1v15, not RSA-PSS.
	// RSA-PSS would be named "SHA256withRSAandMGF1" or "SHA256withRSA/PSS".
	// See: https://docs.oracle.com/en/java/javase/12/docs/specs/security/standard-names.html
	// Midtrans docs: https://docs.midtrans.com/reference/signature-generation#asymmetric-signature-sha256withrsa

	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signatureBytes)
	if err != nil {
		m.logger.Error("PKCS1v15 signature verification failed", zap.Error(err))
		return false
	}

	m.logger.Info("Webhook signature validation successful using PKCS1v15 (Midtrans requirement)")
	return true
}

// GetName returns the gateway's name.
func (m *Service) GetName() string {
	return "midtrans"
}

// WebhookResult represents the result of processing a payment webhook
type WebhookResult struct {
	GatewayTransactionID string        // Transaction ID from payment gateway
	PaymentCode          string        // Our internal payment reference code
	Status               PaymentStatus // Payment status
	RawPayload           string        // Raw webhook payload for debugging
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

func (m *Service) HandleQRISWebhook(
	ctx context.Context,
	payload []byte,
	headers map[string]string,
) (*WebhookResult, error) {
	method := "POST"
	endpoint := "/v1.0/qr/qr-mpm-notify"
	timestamp := headers[HeaderXTimestamp]
	signature := headers[HeaderXSignature]
	externalID := headers[HeaderXExternalID]

	if timestamp == "" || signature == "" || externalID == "" {
		return nil, fmt.Errorf("missing required webhook headers")
	}

	select {
	case <-ctx.Done():
		m.logger.Warn("Context cancelled before webhook validation")
		return nil, ctx.Err()
	default:
	}

	rawPayload := payload

	if !m.ValidateWebhookSignatureBISnap(method, endpoint, string(rawPayload), timestamp, signature) {
		m.logger.Warn(
			"Invalid webhook signature",
			zap.String(HeaderXTimestamp, timestamp),
			zap.String(HeaderXSignature, signature),
			zap.String(HeaderXExternalID, externalID),
		)
		return nil, fmt.Errorf("invalid webhook signature")
	}

	var payloadMap map[string]interface{}
	if err := sonic.Unmarshal(rawPayload, &payloadMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook payload: %w", err)
	}

	latestTransactionStatus, ok := payloadMap["latestTransactionStatus"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid webhook payload structure")
	}

	originalPartnerReferenceNo, ok := payloadMap["originalPartnerReferenceNo"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid webhook payload structure: missing originalPartnerReferenceNo")
	}

	originalReferenceNo, ok := payloadMap["originalReferenceNo"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid webhook payload structure: missing originalReferenceNo")
	}

	m.logger.Info("Received webhook",
		zap.String("originalPartnerReferenceNo", originalPartnerReferenceNo),
		zap.String("originalReferenceNo", originalReferenceNo),
		zap.String("latestTransactionStatus", latestTransactionStatus),
	)

	var newStatus PaymentStatus
	switch latestTransactionStatus {
	case "00":
		newStatus = PaymentStatusCompleted
	case "03":
		newStatus = PaymentStatusPending
	case "04":
		newStatus = PaymentStatusRefunded
	case "05":
		newStatus = PaymentStatusCancelled
	case "06", "07", "08", "09":
		newStatus = PaymentStatusFailed
	default:
		newStatus = PaymentStatusPending
	}

	return &WebhookResult{
		GatewayTransactionID: originalReferenceNo,
		PaymentCode:          originalPartnerReferenceNo,
		Status:               newStatus,
		RawPayload:           string(rawPayload),
	}, nil
}
