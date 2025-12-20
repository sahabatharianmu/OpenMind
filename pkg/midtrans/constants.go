package midtrans

import "net/http"

// BI-SNAP Response Codes for QRIS Payment
const (
	QRISResponseCodeSuccess    = "2004700" // Successful
	QRISResponseCodeProcessing = "2024700" // Transaction still on process

	QRISResponseCodeBadRequest          = "4004700" // General request failed error
	QRISResponseCodeInvalidFormat       = "4004701" // Invalid format
	QRISResponseCodeMissingField        = "4004702" // Missing or invalid format on mandatory field
	QRISResponseCodeUnauthorized        = "4014700" // General unauthorized error
	QRISResponseCodeInvalidToken        = "4014701" // Token found in request is invalid
	QRISResponseCodeTokenNotFound       = "4014703" // Token not found in the system
	QRISResponseCodeExpired             = "4034700" // Transaction expired
	QRISResponseCodeNotAllowed          = "4034701" // Merchant not allowed to call Direct Debit APIs
	QRISResponseCodeAmountLimit         = "4034702" // Exceeds Transaction Amount Limit
	QRISResponseCodeSuspectedFraud      = "4034703" // Suspected Fraud
	QRISResponseCodeTooManyRequests     = "4034704" // Too many request, Exceeds Transaction Frequency Limit
	QRISResponseCodeAbnormalStatus      = "4034705" // Account or User status is abnormal
	QRISResponseCodeDormantAccount      = "4034709" // The account is dormant
	QRISResponseCodeInsufficientFunds   = "4034714" // Insufficient Funds
	QRISResponseCodeNotPermitted        = "4034715" // Transaction Not Permitted
	QRISResponseCodeSuspendTransaction  = "4034716" // Suspend Transaction
	QRISResponseCodeInactiveAccount     = "4034718" // Indicates inactive account
	QRISResponseCodeMerchantSuspended   = "4034719" // Merchant is suspended from calling any APIs
	QRISResponseCodeDailyLimitExceeded  = "4034720" // Merchant aggregated purchase amount on that day exceeds the agreed limit
	QRISResponseCodeTokenLimitInvalid   = "4034722" // The token limit desired by the merchant is not within the agreed range
	QRISResponseCodeAccountDailyLimit   = "4034723" // Account aggregated purchase amount on that day exceeds the agreed limit
	QRISResponseCodeInvalidStatus       = "4044700" // Invalid transaction status
	QRISResponseCodeNotFound            = "4044701" // Transaction not found
	QRISResponseCodeInvalidRouting      = "4044702" // Invalid Routing
	QRISResponseCodeCancelledByCustomer = "4044704" // Transaction is cancelled by customer
	QRISResponseCodeMerchantNotExist    = "4044708" // Merchant does not exist or status abnormal
	QRISResponseCodeInvalidTransition   = "4044710" // Invalid API transition within a journey
	QRISResponseCodeAmountMismatch      = "4044713" // The amount doesn't match with what supposed to
	QRISResponseCodeTerminalNotExist    = "4044717" // Terminal does not exist in the system
	QRISResponseCodeInconsistentParam   = "4044718" // Inconsistent request parameter
	QRISResponseCodeNotSupported        = "4054700" // Requested function is not supported
	QRISResponseCodeOperationNotAllowed = "4054701" // Requested operation to cancel/refund transaction Is not allowed at this time
	QRISResponseCodeDuplicateExternalID = "4094700" // Cannot use same X-EXTERNAL-ID in same day
	QRISResponseCodeAlreadyProcessed    = "4094701" // Transaction has previously been processed
	QRISResponseCodeMaxLimitExceeded    = "4294700" // Maximum transaction limit exceeded

	QRISResponseCodeInternalError  = "5004700" // General Error
	QRISResponseCodeUnknownFailure = "5004701" // Unknown Internal Server Failure
	QRISResponseCodeBackendFailure = "5004702" // Backend system failure
	QRISResponseCodeTimeout        = "5044700" // Timeout from the issuer
)

const (
	AccessTokenResponseCodeSuccess       = "2007300" // Success
	AccessTokenResponseCodeUnauthorized  = "4017300" // Unauthorized Signature
	AccessTokenResponseCodeInternalError = "5007300" // Internal Server Error
)

const (
	WebhookResponseCodeSuccess       = "2005100" // Success
	WebhookResponseCodeInvalidField  = "4005102" // Invalid Mandatory Field
	WebhookResponseCodeUnauthorized  = "4015100" // Unauthorized
	WebhookResponseCodeInternalError = "5005101" // Internal Server Error
	WebhookResponseCodeTimeout       = "5045100" // Timeout
)

const (
	StatusSuccess = "success"
)

// QRISResponseCodeInfo contains response code information
type QRISResponseCodeInfo struct {
	Code        string
	HTTPStatus  int
	Description string
	Category    string // StatusSuccess, "client_error", "server_error"
}

// QRISResponseCodes maps response codes to their information
var QRISResponseCodes = map[string]QRISResponseCodeInfo{
	QRISResponseCodeSuccess: {
		Code:        QRISResponseCodeSuccess,
		HTTPStatus:  http.StatusOK,
		Description: "Successful",
		Category:    StatusSuccess,
	},
	QRISResponseCodeProcessing: {
		Code:        QRISResponseCodeProcessing,
		HTTPStatus:  http.StatusAccepted,
		Description: "Transaction still on process",
		Category:    StatusSuccess,
	},

	QRISResponseCodeBadRequest: {
		Code:        QRISResponseCodeBadRequest,
		HTTPStatus:  http.StatusBadRequest,
		Description: "General request failed error, including message parsing failed",
		Category:    "client_error",
	},
	QRISResponseCodeInvalidFormat: {
		Code:        QRISResponseCodeInvalidFormat,
		HTTPStatus:  http.StatusBadRequest,
		Description: "Invalid format",
		Category:    "client_error",
	},
	QRISResponseCodeMissingField: {
		Code:        QRISResponseCodeMissingField,
		HTTPStatus:  http.StatusBadRequest,
		Description: "Missing or invalid format on mandatory field",
		Category:    "client_error",
	},
	QRISResponseCodeUnauthorized: {
		Code:        QRISResponseCodeUnauthorized,
		HTTPStatus:  http.StatusUnauthorized,
		Description: "General unauthorized error",
		Category:    "client_error",
	},
	QRISResponseCodeInvalidToken: {
		Code:        QRISResponseCodeInvalidToken,
		HTTPStatus:  http.StatusUnauthorized,
		Description: "Token found in request is invalid",
		Category:    "client_error",
	},
	QRISResponseCodeTokenNotFound: {
		Code:        QRISResponseCodeTokenNotFound,
		HTTPStatus:  http.StatusUnauthorized,
		Description: "Token not found in the system",
		Category:    "client_error",
	},
	QRISResponseCodeExpired: {
		Code:        QRISResponseCodeExpired,
		HTTPStatus:  http.StatusForbidden,
		Description: "Transaction expired",
		Category:    "client_error",
	},
	QRISResponseCodeNotAllowed: {
		Code:        QRISResponseCodeNotAllowed,
		HTTPStatus:  http.StatusForbidden,
		Description: "This merchant is not allowed to call Direct Debit APIs",
		Category:    "client_error",
	},
	QRISResponseCodeAmountLimit: {
		Code:        QRISResponseCodeAmountLimit,
		HTTPStatus:  http.StatusForbidden,
		Description: "Exceeds Transaction Amount Limit",
		Category:    "client_error",
	},
	QRISResponseCodeSuspectedFraud: {
		Code:        QRISResponseCodeSuspectedFraud,
		HTTPStatus:  http.StatusForbidden,
		Description: "Suspected Fraud",
		Category:    "client_error",
	},
	QRISResponseCodeTooManyRequests: {
		Code:        QRISResponseCodeTooManyRequests,
		HTTPStatus:  http.StatusForbidden,
		Description: "Too many request, Exceeds Transaction Frequency Limit",
		Category:    "client_error",
	},
	QRISResponseCodeAbnormalStatus: {
		Code:        QRISResponseCodeAbnormalStatus,
		HTTPStatus:  http.StatusForbidden,
		Description: "Account or User status is abnormal",
		Category:    "client_error",
	},
	QRISResponseCodeDormantAccount: {
		Code:        QRISResponseCodeDormantAccount,
		HTTPStatus:  http.StatusForbidden,
		Description: "The account is dormant",
		Category:    "client_error",
	},
	QRISResponseCodeInsufficientFunds: {
		Code:        QRISResponseCodeInsufficientFunds,
		HTTPStatus:  http.StatusForbidden,
		Description: "Insufficient Funds",
		Category:    "client_error",
	},
	QRISResponseCodeNotPermitted: {
		Code:        QRISResponseCodeNotPermitted,
		HTTPStatus:  http.StatusForbidden,
		Description: "Transaction Not Permitted",
		Category:    "client_error",
	},
	QRISResponseCodeSuspendTransaction: {
		Code:        QRISResponseCodeSuspendTransaction,
		HTTPStatus:  http.StatusForbidden,
		Description: "Suspend Transaction",
		Category:    "client_error",
	},
	QRISResponseCodeInactiveAccount: {
		Code:        QRISResponseCodeInactiveAccount,
		HTTPStatus:  http.StatusForbidden,
		Description: "Indicates inactive account",
		Category:    "client_error",
	},
	QRISResponseCodeMerchantSuspended: {
		Code:        QRISResponseCodeMerchantSuspended,
		HTTPStatus:  http.StatusForbidden,
		Description: "Merchant is suspended from calling any APIs",
		Category:    "client_error",
	},
	QRISResponseCodeDailyLimitExceeded: {
		Code:        QRISResponseCodeDailyLimitExceeded,
		HTTPStatus:  http.StatusForbidden,
		Description: "Merchant aggregated purchase amount on that day exceeds the agreed limit",
		Category:    "client_error",
	},
	QRISResponseCodeTokenLimitInvalid: {
		Code:        QRISResponseCodeTokenLimitInvalid,
		HTTPStatus:  http.StatusForbidden,
		Description: "The token limit desired by the merchant is not within the agreed range",
		Category:    "client_error",
	},
	QRISResponseCodeAccountDailyLimit: {
		Code:        QRISResponseCodeAccountDailyLimit,
		HTTPStatus:  http.StatusForbidden,
		Description: "Account aggregated purchase amount on that day exceeds the agreed limit",
		Category:    "client_error",
	},
	QRISResponseCodeInvalidStatus: {
		Code:        QRISResponseCodeInvalidStatus,
		HTTPStatus:  http.StatusNotFound,
		Description: "Invalid transaction status",
		Category:    "client_error",
	},
	QRISResponseCodeNotFound: {
		Code:        QRISResponseCodeNotFound,
		HTTPStatus:  http.StatusNotFound,
		Description: "Transaction not found",
		Category:    "client_error",
	},
	QRISResponseCodeInvalidRouting: {
		Code:        QRISResponseCodeInvalidRouting,
		HTTPStatus:  http.StatusNotFound,
		Description: "Invalid Routing",
		Category:    "client_error",
	},
	QRISResponseCodeCancelledByCustomer: {
		Code:        QRISResponseCodeCancelledByCustomer,
		HTTPStatus:  http.StatusNotFound,
		Description: "Transaction is cancelled by customer",
		Category:    "client_error",
	},
	QRISResponseCodeMerchantNotExist: {
		Code:        QRISResponseCodeMerchantNotExist,
		HTTPStatus:  http.StatusNotFound,
		Description: "Merchant does not exist or status abnormal",
		Category:    "client_error",
	},
	QRISResponseCodeInvalidTransition: {
		Code:        QRISResponseCodeInvalidTransition,
		HTTPStatus:  http.StatusNotFound,
		Description: "Invalid API transition within a journey",
		Category:    "client_error",
	},
	QRISResponseCodeAmountMismatch: {
		Code:        QRISResponseCodeAmountMismatch,
		HTTPStatus:  http.StatusNotFound,
		Description: "The amount doesn't match with what supposed to",
		Category:    "client_error",
	},
	QRISResponseCodeTerminalNotExist: {
		Code:        QRISResponseCodeTerminalNotExist,
		HTTPStatus:  http.StatusNotFound,
		Description: "Terminal does not exist in the system",
		Category:    "client_error",
	},
	QRISResponseCodeInconsistentParam: {
		Code:        QRISResponseCodeInconsistentParam,
		HTTPStatus:  http.StatusNotFound,
		Description: "Inconsistent request parameter",
		Category:    "client_error",
	},
	QRISResponseCodeNotSupported: {
		Code:        QRISResponseCodeNotSupported,
		HTTPStatus:  http.StatusMethodNotAllowed,
		Description: "Requested function is not supported",
		Category:    "client_error",
	},
	QRISResponseCodeOperationNotAllowed: {
		Code:        QRISResponseCodeOperationNotAllowed,
		HTTPStatus:  http.StatusMethodNotAllowed,
		Description: "Requested operation to cancel/refund transaction Is not allowed at this time",
		Category:    "client_error",
	},
	QRISResponseCodeDuplicateExternalID: {
		Code:        QRISResponseCodeDuplicateExternalID,
		HTTPStatus:  http.StatusConflict,
		Description: "Cannot use same X-EXTERNAL-ID in same day",
		Category:    "client_error",
	},
	QRISResponseCodeAlreadyProcessed: {
		Code:        QRISResponseCodeAlreadyProcessed,
		HTTPStatus:  http.StatusConflict,
		Description: "Transaction has previously been processed",
		Category:    "client_error",
	},
	QRISResponseCodeMaxLimitExceeded: {
		Code:        QRISResponseCodeMaxLimitExceeded,
		HTTPStatus:  http.StatusTooManyRequests,
		Description: "Maximum transaction limit exceeded",
		Category:    "client_error",
	},

	// Server error codes (5xx)
	QRISResponseCodeInternalError: {
		Code:        QRISResponseCodeInternalError,
		HTTPStatus:  http.StatusInternalServerError,
		Description: "General Error",
		Category:    "server_error",
	},
	QRISResponseCodeUnknownFailure: {
		Code:        QRISResponseCodeUnknownFailure,
		HTTPStatus:  http.StatusInternalServerError,
		Description: "Unknown Internal Server Failure, Please retry the process again",
		Category:    "server_error",
	},
	QRISResponseCodeBackendFailure: {
		Code:        QRISResponseCodeBackendFailure,
		HTTPStatus:  http.StatusInternalServerError,
		Description: "Backend system failure, etc",
		Category:    "server_error",
	},
	QRISResponseCodeTimeout: {
		Code:        QRISResponseCodeTimeout,
		HTTPStatus:  http.StatusGatewayTimeout,
		Description: "Timeout from the issuer",
		Category:    "server_error",
	},
}

// AccessTokenResponseCodes maps access token response codes to their information
var AccessTokenResponseCodes = map[string]QRISResponseCodeInfo{
	AccessTokenResponseCodeSuccess: {
		Code:        AccessTokenResponseCodeSuccess,
		HTTPStatus:  http.StatusOK,
		Description: StatusSuccess,
		Category:    StatusSuccess,
	},
	AccessTokenResponseCodeUnauthorized: {
		Code:        AccessTokenResponseCodeUnauthorized,
		HTTPStatus:  http.StatusUnauthorized,
		Description: "Unauthorized Signature",
		Category:    "client_error",
	},
	AccessTokenResponseCodeInternalError: {
		Code:        AccessTokenResponseCodeInternalError,
		HTTPStatus:  http.StatusInternalServerError,
		Description: "Internal Server Error",
		Category:    "server_error",
	},
}

// WebhookResponseCodes maps webhook response codes to their information
var WebhookResponseCodes = map[string]QRISResponseCodeInfo{
	WebhookResponseCodeSuccess: {
		Code:        WebhookResponseCodeSuccess,
		HTTPStatus:  http.StatusOK,
		Description: StatusSuccess,
		Category:    StatusSuccess,
	},
	WebhookResponseCodeInvalidField: {
		Code:        WebhookResponseCodeInvalidField,
		HTTPStatus:  http.StatusBadRequest,
		Description: "Invalid Mandatory Field",
		Category:    "client_error",
	},
	WebhookResponseCodeUnauthorized: {
		Code:        WebhookResponseCodeUnauthorized,
		HTTPStatus:  http.StatusUnauthorized,
		Description: "Unauthorized",
		Category:    "client_error",
	},
	WebhookResponseCodeInternalError: {
		Code:        WebhookResponseCodeInternalError,
		HTTPStatus:  http.StatusInternalServerError,
		Description: "Internal Server Error",
		Category:    "server_error",
	},
	WebhookResponseCodeTimeout: {
		Code:        WebhookResponseCodeTimeout,
		HTTPStatus:  http.StatusGatewayTimeout,
		Description: "Timeout",
		Category:    "server_error",
	},
}

// IsSuccessCode checks if the response code indicates success
func IsSuccessCode(code string) bool {
	if info, exists := QRISResponseCodes[code]; exists {
		return info.Category == StatusSuccess
	}
	if info, exists := AccessTokenResponseCodes[code]; exists {
		return info.Category == StatusSuccess
	}
	if info, exists := WebhookResponseCodes[code]; exists {
		return info.Category == StatusSuccess
	}
	return false
}

// GetResponseCodeInfo returns information about a response code
func GetResponseCodeInfo(code string) (QRISResponseCodeInfo, bool) {
	if info, exists := QRISResponseCodes[code]; exists {
		return info, true
	}
	if info, exists := AccessTokenResponseCodes[code]; exists {
		return info, true
	}
	if info, exists := WebhookResponseCodes[code]; exists {
		return info, true
	}
	return QRISResponseCodeInfo{}, false
}

// IsRetryableError checks if the error is retryable (server errors or specific client errors)
func IsRetryableError(code string) bool {
	info, exists := GetResponseCodeInfo(code)
	if !exists {
		return false
	}

	if info.Category == "server_error" {
		return true
	}

	retryableClientErrors := map[string]bool{
		QRISResponseCodeTooManyRequests:  true,
		QRISResponseCodeMaxLimitExceeded: true,
	}

	return retryableClientErrors[code]
}

// FormatErrorMessage creates a user-friendly error message based on response code
func FormatErrorMessage(code string) string {
	info, exists := GetResponseCodeInfo(code)
	if !exists {
		return "Unknown error occurred"
	}

	switch info.Category {
	case StatusSuccess:
		return "Transaction completed successfully"
	case "client_error":
		switch code {
		case QRISResponseCodeInsufficientFunds:
			return "Insufficient funds. Please check your account balance."
		case QRISResponseCodeExpired:
			return "Transaction has expired. Please try again."
		case QRISResponseCodeInvalidToken:
			return "Authentication failed. Please try again."
		case QRISResponseCodeAmountLimit:
			return "Transaction amount exceeds the allowed limit."
		case QRISResponseCodeTooManyRequests:
			return "Too many requests. Please wait a moment and try again."
		case QRISResponseCodeNotFound:
			return "Transaction not found."
		case QRISResponseCodeCancelledByCustomer:
			return "Transaction was cancelled."
		default:
			return info.Description
		}
	case "server_error":
		return "Service temporarily unavailable. Please try again later."
	default:
		return info.Description
	}
}
