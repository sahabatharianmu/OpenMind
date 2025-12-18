package middleware

import (
	"context"
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
		resourceType, resourceID := parseResourceFromPath(path, c)

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
				nil, // details can be enhanced later
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

func parseResourceFromPath(path string, c *app.RequestContext) (string, *uuid.UUID) {
	// Path format: /api/v1/{resource}/{id?}
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 3 {
		return "unknown", nil
	}

	resourceType := parts[2] // e.g., "patients", "appointments"

	// Normalize to singular form for resource_type
	resourceType = strings.TrimSuffix(resourceType, "s")
	resourceType = strings.ReplaceAll(resourceType, "-", "_")

	// Try to extract ID from path parameter
	idParam := c.Param("id")
	if idParam != "" {
		if id, err := uuid.Parse(idParam); err == nil {
			return resourceType, &id
		}
	}

	return resourceType, nil
}
