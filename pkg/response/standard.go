package response

import (
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/sahabatharianmu/OpenMind/pkg/apperrors"
)

// StandardResponse represents a standard API response
type StandardResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
	Meta    *MetaData    `json:"meta,omitempty"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// MetaData represents pagination and metadata
type MetaData struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// Success creates a success response
func Success(message string, data ...interface{}) StandardResponse {
	res := StandardResponse{
		Success: true,
		Message: message,
	}

	if len(data) > 0 {
		res.Data = data[0]
	}

	return res
}

// Created creates a created response
func Created(c *app.RequestContext, data interface{}, message string) {
	response := StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(consts.StatusCreated, response)
}

// Error creates an error response
func Error(c *app.RequestContext, code int, errorCode string, message string, details map[string]interface{}) {
	response := StandardResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
	c.JSON(code, response)
}

// BadRequest creates a bad request response
func BadRequest(c *app.RequestContext, message string, details map[string]interface{}) {
	Error(c, consts.StatusBadRequest, ErrorCodeBadRequest, message, details)
}

// Unauthorized creates an unauthorized response
func Unauthorized(c *app.RequestContext, message string) {
	Error(c, consts.StatusUnauthorized, ErrorCodeAuthentication, message, nil)
}

// Forbidden creates a forbidden response
func Forbidden(c *app.RequestContext, message string) {
	Error(c, consts.StatusForbidden, ErrorCodeAuthorization, message, nil)
}

// NotFound creates a not found response
func NotFound(c *app.RequestContext, resource string) {
	message := "Resource not found"
	if resource != "" {
		message = resource + " not found"
	}
	Error(c, consts.StatusNotFound, ErrorCodeNotFound, message, nil)
}

// InternalServerError creates an internal server error response
func InternalServerError(c *app.RequestContext, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(c, consts.StatusInternalServerError, ErrorCodeInternal, message, nil)
}

// ValidationError creates a validation error response
func ValidationError(c *app.RequestContext, field string, message string) {
	details := map[string]interface{}{
		"field":   field,
		"message": message,
	}
	Error(c, consts.StatusBadRequest, ErrorCodeValidation, "Validation failed", details)
}

// PaginatedResponse creates a paginated response
func PaginatedResponse(c *app.RequestContext, data interface{}, page, limit int, total int64, message string) {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta: &MetaData{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	c.JSON(consts.StatusOK, response)
}

func HandleError(c *app.RequestContext, err error) {
	appErr := &apperrors.AppError{}
	if errors.As(err, &appErr) {
		Error(c, appErr.Code, getErrorCode(appErr.Code), appErr.Message, nil)
		return
	}

	InternalServerError(c, err.Error())
}

func getErrorCode(statusCode int) string {
	switch statusCode {
	case consts.StatusBadRequest:
		return ErrorCodeBadRequest
	case consts.StatusUnauthorized:
		return ErrorCodeAuthentication
	case consts.StatusForbidden:
		return ErrorCodeAuthorization
	case consts.StatusNotFound:
		return ErrorCodeNotFound
	case consts.StatusConflict:
		return "CONFLICT"
	default:
		return ErrorCodeInternal
	}
}
