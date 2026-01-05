package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"github.com/sahabatharianmu/OpenMind/pkg/security"
	"go.uber.org/zap"
)

// Context keys for storing user information
type contextKey string

const (
	UserIDKey contextKey = "user_id"
	UserKey   contextKey = "user"
)

// JWTAuthMiddleware creates a JWT authentication middleware
func JWTAuthMiddleware(cfg *config.Config, logger logger.Logger) app.HandlerFunc {
	jwtService := security.NewJWTService(cfg)

	return func(ctx context.Context, c *app.RequestContext) {
		// Get Authorization header
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			response.Unauthorized(c, "Missing authorization header")
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			response.Unauthorized(c, "Missing token")
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("Invalid JWT token", zap.Error(err))
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Store user information in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("system_role", claims.SystemRole)

		logger.Debug("JWT authentication successful")

		c.Next(ctx)
	}
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c *app.RequestContext) (uuid.UUID, bool) {
	if userID, exists := c.Get(string(UserIDKey)); exists {
		if id, ok := userID.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

// GetUserEmailFromContext retrieves user email from context
func GetUserEmailFromContext(c *app.RequestContext) (string, bool) {
	if email, exists := c.Get("email"); exists {
		if emailStr, ok := email.(string); ok {
			return emailStr, true
		}
	}
	return "", false
}

// GetUserRoleFromContext retrieves user role from context
func GetUserRoleFromContext(c *app.RequestContext) (string, bool) {
	if role, exists := c.Get("role"); exists {
		if roleStr, ok := role.(string); ok {
			return roleStr, true
		}
	}
	return "", false
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roles ...string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		userRole, exists := GetUserRoleFromContext(c)
		if !exists {
			response.Forbidden(c, "User role not found")
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

// GetSystemRoleFromContext retrieves system role from context
func GetSystemRoleFromContext(c *app.RequestContext) (string, bool) {
	if role, exists := c.Get("system_role"); exists {
		if roleStr, ok := role.(string); ok {
			return roleStr, true
		}
	}
	return "", false
}

// RequireSystemRole creates a middleware that requires a specific system role
func RequireSystemRole(roles ...string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		userRole, exists := GetSystemRoleFromContext(c)
		if !exists {
			response.Forbidden(c, "System role not found")
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Forbidden(c, "Insufficient system permissions")
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

// OptionalAuthMiddleware creates an optional authentication middleware
// It validates the token if present but doesn't require it
func OptionalAuthMiddleware(cfg *config.Config, _ logger.Logger) app.HandlerFunc {
	jwtService := security.NewJWTService(cfg)

	return func(ctx context.Context, c *app.RequestContext) {
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.Next(ctx)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next(ctx)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.Next(ctx)
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			// For optional auth, we don't fail if token is invalid
			c.Next(ctx)
			return
		}

		// Store user information in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("system_role", claims.SystemRole)

		c.Next(ctx)
	}
}
