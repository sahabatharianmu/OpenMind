package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

// RequestLogging returns middleware that logs every HTTP request with structured fields.
// Logs at INFO for 2xx, WARN for 4xx, ERROR for 5xx.
func RequestLogging(log logger.Logger) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		// Process request
		c.Next(ctx)

		// Calculate latency
		latency := time.Since(start)
		status := c.Response.StatusCode()

		fields := []zap.Field{
			zap.String("method", string(c.Method())),
			zap.String("path", string(c.Path())),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("body_size", len(c.Response.Body())),
		}

		// Add query string if present
		if query := string(c.URI().QueryString()); query != "" {
			fields = append(fields, zap.String("query", query))
		}

		// Add user ID if authenticated
		if userID, exists := c.Get("userID"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		switch {
		case status >= 500:
			log.Error("Server error", fields...)
		case status >= 400:
			log.Warn("Client error", fields...)
		default:
			log.Info("Request", fields...)
		}
	}
}
