package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"github.com/sahabatharianmu/OpenMind/pkg/security"
)

// Security returns a comprehensive security middleware
func Security() app.HandlerFunc {
	config := security.DefaultOWASPSecurityConfig()
	return security.HeadersMiddleware(config)
}

// InputValidation returns input validation middleware
func InputValidation() app.HandlerFunc {
	return security.InputValidationMiddleware()
}

// RateLimit returns rate limiting middleware
func RateLimit(requests int, window time.Duration) app.HandlerFunc {
	return security.RateLimitingMiddleware(requests, window)
}

// CSRFProtection returns CSRF protection middleware
func CSRFProtection() app.HandlerFunc {
	return security.CSRFProtectionMiddleware(32) //nolint:mnd
}

// FileUploadSecurity returns file upload security middleware
func FileUploadSecurity(_ *config.Config) app.HandlerFunc {
	owaspConfig := security.DefaultOWASPSecurityConfig()
	return security.FileUploadSecurityMiddleware(owaspConfig)
}

// SecurityLogging returns security logging middleware
func SecurityLogging() app.HandlerFunc {
	return security.LoggingMiddleware()
}

// CORS returns a CORS middleware that handles preflight and sets headers
func CORS(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.Request.Header.Peek("Origin"))
		allowOrigin := ""
		for _, o := range cfg.Security.CORSAllowOrigins {
			if o == "*" || o == origin {
				allowOrigin = o
				break
			}
		}

		if allowOrigin == "*" && origin != "" {
			// Reflect the request origin when using * to support credentials safely
			allowOrigin = origin
		}

		if allowOrigin != "" {
			c.Response.Header.Set("Access-Control-Allow-Origin", allowOrigin)
			c.Response.Header.Set("Vary", "Origin")
		}

		// Methods
		if len(cfg.Security.CORSAllowMethods) > 0 {
			c.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(cfg.Security.CORSAllowMethods, ", "))
		}
		// Headers
		if len(cfg.Security.CORSAllowHeaders) > 0 {
			c.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(cfg.Security.CORSAllowHeaders, ", "))
		}
		// Max-Age
		if cfg.Security.CORSMaxAge > 0 {
			c.Response.Header.Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.Security.CORSMaxAge))
		}
		// Credentials
		c.Response.Header.Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight
		if string(c.Request.Header.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}

		c.Next(ctx)
	}
}

// ComprehensiveSecurity returns all security middlewares combined
func ComprehensiveSecurity(cfg *config.Config) []app.HandlerFunc {
	return []app.HandlerFunc{
		Security(),
		InputValidation(),
		RateLimit(cfg.Security.RateLimitRequests, cfg.Security.RateLimitWindow),
		CSRFProtection(),
		FileUploadSecurity(cfg),
		SecurityLogging(),
	}
}

// SecurityErrorHandler handles security-related errors
func SecurityErrorHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				// Log security incident
				// In production, this would integrate with SIEM
				response.InternalServerError(c, "Security error occurred")
			}
		}()

		c.Next(ctx)
	}
}

// RequestSizeLimit returns request size limit middleware
func RequestSizeLimit(maxSize int64) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		contentLength := int64(c.Request.Header.ContentLength())
		if contentLength > 0 && contentLength > maxSize {
			response.Error(c, consts.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "Request entity too large", nil)
			return
		}

		// Check actual body size for chunked transfers
		if len(c.Request.Body()) > int(maxSize) {
			response.Error(c, consts.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "Request entity too large", nil)
			return
		}

		c.Next(ctx)
	}
}

// TimeoutMiddleware adds request timeout protection
func TimeoutMiddleware(_ time.Duration) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// For now, just call next without timeout handling
		// TODO: Implement proper timeout handling compatible with Hertz
		c.Next(ctx)
	}
}

// RequestIDMiddleware adds request ID for tracking
func RequestIDMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		requestID := string(c.Request.Header.Peek("X-Request-ID"))
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Response.Header.Set("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next(ctx)
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8)) //nolint:mnd
}

// randomString generates a random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
