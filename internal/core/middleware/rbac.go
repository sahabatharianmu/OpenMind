package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type RBACMiddleware struct{}

func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{}
}

func (m *RBACMiddleware) HasRole(allowedRoles ...string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		roleVal, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "User role not found")
			c.Abort()
			return
		}

		role := roleVal.(string)

		// Admin can access everything
		if role == "admin" {
			c.Next(ctx)
			return
		}

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
