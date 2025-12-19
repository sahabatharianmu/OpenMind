package middleware

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	auditLogService "github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/service"
)

type AuditMiddleware struct {
	auditSvc auditLogService.AuditLogService
}

func NewAuditMiddleware(auditSvc auditLogService.AuditLogService) *AuditMiddleware {
	return &AuditMiddleware{auditSvc: auditSvc}
}

func (m *AuditMiddleware) Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		method := string(c.Method())
		path := string(c.Path())

		// Log all write operations (POST, PUT, DELETE) and sensitive READ (GET) operations
		isWrite := method == "POST" || method == "PUT" || method == "DELETE"
		isSensitiveRead := method == "GET" && isSensitivePath(path)

		if !isWrite && !isSensitiveRead {
			c.Next(ctx)
			return
		}

		// Extract user info from context (set by auth middleware)
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.Next(ctx)
			return
		}
		userID := userIDVal.(uuid.UUID)

		// Get organization ID
		orgID, err := m.auditSvc.GetOrganizationID(context.Background(), userID)
		if err != nil {
			c.Next(ctx)
			return
		}

		// Continue processing the request
		c.Next(ctx)

		// Only log if request was successful (2xx status)
		statusCode := c.Response.StatusCode()
		if statusCode < 200 || statusCode >= 300 {
			return
		}

		// Determine action and resource type from method and path
		action := methodToAction(method)
		resourceType, resourceID, isListOperation := parseResourceFromPath(path, c)

		// Build audit log details based on operation type
		var details map[string]interface{}
		if resourceType == "patient" {
			if isListOperation {
				// For list operations, indicate it's a list access
				details = map[string]interface{}{
					"operation": "list",
					"resource":  "patients",
				}
				// Try to extract patient count from response body
				if patientCount := extractPatientCountFromResponse(c); patientCount > 0 {
					details["patient_count"] = patientCount
				}
			} else if resourceID != nil {
				// For individual patient access, include patient_id in details
				details = map[string]interface{}{
					"operation":  "read",
					"resource":   "patient",
					"patient_id": resourceID.String(),
				}
			}
		}

		// Get client info
		ipAddr := c.ClientIP()
		userAgent := string(c.UserAgent())

		// Log the action (non-blocking)
		go func() {
			_ = m.auditSvc.Log(
				context.Background(),
				action,
				resourceType,
				resourceID,
				userID,
				orgID,
				details,
				&ipAddr,
				&userAgent,
			)
		}()
	}
}

func methodToAction(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	case "GET":
		return "read"
	default:
		return "unknown"
	}
}

func isSensitivePath(path string) bool {
	// Path format: /api/v1/{resource}/{id?}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return false
	}

	resource := parts[2]
	sensitiveResources := map[string]bool{
		"patients":       true,
		"clinical-notes": true,
		"appointments":   true,
		"invoices":       true,
		"export":         true,
	}

	return sensitiveResources[resource]
}

// parseResourceFromPath parses the resource type, ID, and determines if it's a list operation
// Returns: resourceType, resourceID, isListOperation
func parseResourceFromPath(path string, c *app.RequestContext) (string, *uuid.UUID, bool) {
	// Path format: /api/v1/{resource}/{id?}
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 3 {
		return "unknown", nil, false
	}

	resourceType := parts[2] // e.g., "patients", "appointments"

	// Normalize to singular form for resource_type
	resourceType = strings.TrimSuffix(resourceType, "s")
	resourceType = strings.ReplaceAll(resourceType, "-", "_")

	// Try to extract ID from path parameter
	idParam := c.Param("id")
	if idParam != "" {
		if id, err := uuid.Parse(idParam); err == nil {
			// Individual resource access (has ID)
			return resourceType, &id, false
		}
	}

	// No ID in path = list operation
	return resourceType, nil, true
}

// extractPatientCountFromResponse attempts to extract patient count from the response body
// Returns 0 if extraction fails or count is not available
func extractPatientCountFromResponse(c *app.RequestContext) int64 {
	// Check if response is JSON
	contentType := string(c.Response.Header.ContentType())
	if !strings.Contains(contentType, "application/json") {
		return 0
	}

	// Get response body
	body := c.Response.Body()
	if len(body) == 0 {
		return 0
	}

	// Parse JSON response
	var response struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return 0
	}

	// Extract total from data (data is a map[string]interface{})
	dataMap, ok := response.Data.(map[string]interface{})
	if !ok {
		return 0
	}

	// Extract total value (JSON numbers are float64)
	if total, ok := dataMap["total"].(float64); ok {
		return int64(total)
	}

	return 0
}
