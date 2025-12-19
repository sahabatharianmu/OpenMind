package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type RBACMiddleware struct{}

func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{}
}

// HasRole creates middleware that checks if user has one of the allowed roles
// Role is retrieved from request context, which is set by tenant middleware
// from the organization_members table (organization-specific role)
func (m *RBACMiddleware) HasRole(allowedRoles ...string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get role from context (set by tenant middleware from organization_members table)
		roleVal, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "User role not found")
			c.Abort()
			return
		}

		role := roleVal.(string)

		// Admin can access everything (from organization_members table)
		// This includes: audit logs, export/import, invoices, organization settings, etc.
		if role == constants.RoleAdmin {
			c.Next(ctx)
			return
		}

		// Owner can also access everything (if using owner role)
		if role == constants.RoleOwner {
			c.Next(ctx)
			return
		}

		// Check if user has one of the allowed roles
		isAllowed := false
		for _, r := range allowedRoles {
			if r == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			response.Forbidden(c, "You do not have permission to perform this action")
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}
