package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// OWASPSecurityConfig holds OWASP security configuration
type OWASPSecurityConfig struct {
    ContentSecurityPolicy string
    XFrameOptions         string
    XContentTypeOptions   string
    XXSSProtection        string
    ReferrerPolicy        string
    PermissionsPolicy     string
    StrictTransportSecurity string
    XPoweredBy            bool
    ServerHeader          string
    MaxRequestSize        int64
    AllowedFileTypes      []string
    MaxFileSize          int64
}

// DefaultOWASPSecurityConfig returns default OWASP security configuration
func DefaultOWASPSecurityConfig() *OWASPSecurityConfig {
    return &OWASPSecurityConfig{
        ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'nonce-{nonce}' https://cdn.jsdelivr.net; style-src 'self' 'nonce-{nonce}' https://fonts.googleapis.com https://cdn.jsdelivr.net; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https:;",
        XFrameOptions:         "DENY",
        XContentTypeOptions:   "nosniff",
        XXSSProtection:        "1; mode=block",
        ReferrerPolicy:        "strict-origin-when-cross-origin",
        PermissionsPolicy:     "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=()",
        StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
        XPoweredBy:            false,
        ServerHeader:          "SME-Tax-Platform",
        MaxRequestSize:        10 * 1024 * 1024, // 10MB
        AllowedFileTypes:      []string{".pdf", ".jpg", ".jpeg", ".png", ".xlsx", ".csv"},
        MaxFileSize:          5 * 1024 * 1024, // 5MB
    }
}

// SecurityHeadersMiddleware adds OWASP security headers
func SecurityHeadersMiddleware(config *OWASPSecurityConfig) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        nonceBytes := make([]byte, 16)
        rand.Read(nonceBytes)
        nonce := base64.StdEncoding.EncodeToString(nonceBytes)
        c.Set("csp_nonce", nonce)
        if config.ContentSecurityPolicy != "" {
            csp := strings.ReplaceAll(config.ContentSecurityPolicy, "{nonce}", nonce)
            c.Response.Header.Set("Content-Security-Policy", csp)
        }

		// X-Frame-Options
		if config.XFrameOptions != "" {
			c.Response.Header.Set("X-Frame-Options", config.XFrameOptions)
		}

		// X-Content-Type-Options
		if config.XContentTypeOptions != "" {
			c.Response.Header.Set("X-Content-Type-Options", config.XContentTypeOptions)
		}

		// X-XSS-Protection
		if config.XXSSProtection != "" {
			c.Response.Header.Set("X-XSS-Protection", config.XXSSProtection)
		}

		// Referrer Policy
		if config.ReferrerPolicy != "" {
			c.Response.Header.Set("Referrer-Policy", config.ReferrerPolicy)
		}

		// Permissions Policy
		if config.PermissionsPolicy != "" {
			c.Response.Header.Set("Permissions-Policy", config.PermissionsPolicy)
		}

		// Strict Transport Security (HTTPS only)
        if config.StrictTransportSecurity != "" && strings.HasPrefix(string(c.Request.URI().Scheme()), "https") {
            c.Response.Header.Set("Strict-Transport-Security", config.StrictTransportSecurity)
        }

		// Remove X-Powered-By if disabled
		if !config.XPoweredBy {
			c.Response.Header.Del("X-Powered-By")
		}

		// Set custom Server header
		if config.ServerHeader != "" {
			c.Response.Header.Set("Server", config.ServerHeader)
		}

        c.Next(ctx)
    }
}

// InputValidationMiddleware validates and sanitizes input
func InputValidationMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if len(c.Request.Body()) > 10*1024*1024 { // 10MB limit
			c.AbortWithStatus(consts.StatusRequestEntityTooLarge)
			return
		}

		c.Request.URI().QueryArgs().VisitAll(func(key, value []byte) {
			sanitizedValue := html.EscapeString(string(value))
			c.Request.URI().QueryArgs().Set(string(key), sanitizedValue)
		})

		path := string(c.Request.URI().Path())
		if containsDangerousPatterns(path) {
			c.AbortWithStatus(consts.StatusBadRequest)
			return
		}

		c.Next(ctx)
	}
}

// RateLimitingMiddleware implements rate limiting
func RateLimitingMiddleware(requests int, window time.Duration) app.HandlerFunc {
	type clientInfo struct {
		requests int
		window   time.Time
	}

	clients := make(map[string]*clientInfo)
	var mu sync.RWMutex

	return func(ctx context.Context, c *app.RequestContext) {
		ip := c.ClientIP()

		mu.Lock()
		defer mu.Unlock()

		now := time.Now()
		client, exists := clients[ip]

		if !exists || now.Sub(client.window) > window {
			client = &clientInfo{
				requests: 1,
				window:   now,
			}
			clients[ip] = client
		} else {
			client.requests++
			if client.requests > requests {
				c.Response.Header.Set("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
				c.Response.Header.Set("X-RateLimit-Remaining", "0")
				c.Response.Header.Set("X-RateLimit-Reset", fmt.Sprintf("%d", client.window.Add(window).Unix()))
				c.AbortWithStatus(consts.StatusTooManyRequests)
				return
			}
		}

		c.Response.Header.Set("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		c.Response.Header.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", requests-client.requests))
		c.Response.Header.Set("X-RateLimit-Reset", fmt.Sprintf("%d", client.window.Add(window).Unix()))

		c.Next(ctx)
	}
}

// CSRFProtectionMiddleware provides CSRF protection
func CSRFProtectionMiddleware(tokenLength int) app.HandlerFunc {
	if tokenLength < 32 {
		tokenLength = 32
	}

	return func(ctx context.Context, c *app.RequestContext) {
		// Skip CSRF for GET, HEAD, OPTIONS requests
		method := string(c.Request.Header.Method())
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next(ctx)
			return
		}

		// Generate CSRF token for GET requests
		if method == "GET" {
			csrfToken := generateCSRFToken(tokenLength)
			c.Response.Header.Set("X-CSRF-Token", csrfToken)
			c.Next(ctx)
			return
		}

		// Validate CSRF token for POST, PUT, DELETE requests
		csrfToken := string(c.Request.Header.Peek("X-CSRF-Token"))
		if csrfToken == "" {
			csrfToken = string(c.Request.Header.Peek("X-Xsrf-Token"))
		}

		if csrfToken == "" {
			c.AbortWithStatus(consts.StatusForbidden)
			return
		}

		// Validate token (in production, this should be validated against session/cache)
		if len(csrfToken) < tokenLength {
			c.AbortWithStatus(consts.StatusForbidden)
			return
		}

		c.Next(ctx)
	}
}

// FileUploadSecurityMiddleware validates file uploads
func FileUploadSecurityMiddleware(config *OWASPSecurityConfig) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		contentType := string(c.Request.Header.ContentType())
		
		// Check if this is a multipart upload
		if !strings.Contains(contentType, "multipart/form-data") {
			c.Next(ctx)
			return
		}

		// Validate total request size
		if int64(len(c.Request.Body())) > config.MaxRequestSize {
			c.AbortWithStatus(consts.StatusRequestEntityTooLarge)
			return
		}

		// TODO: Parse multipart data here and validate each file individually
		
		c.Next(ctx)
	}
}

// SecurityLoggingMiddleware logs security events
func SecurityLoggingMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		
		c.Next(ctx)
		
		duration := time.Since(start)
		status := c.Response.StatusCode()
		
		// Log suspicious activities
		if status >= 400 {
			// Log security events
			// In production, this would integrate with your logging system
			fmt.Printf("[SECURITY] %s %s %d %v %s\n", 
				string(c.Request.Header.Method()), 
				string(c.Request.URI().Path()), 
				status, 
				duration,
				c.ClientIP())
		}
	}
}

func containsDangerousPatterns(input string) bool {
	dangerousPatterns := []string{
		"../", "..\\", "<script", "javascript:", "data:", "vbscript:", 
		"onload=", "onerror=", "onclick=", "eval(", "expression(", 
		"${", "#{", "<%", ">>>", "&&", "||", ";", "`", "$()",
	}
	
	input = strings.ToLower(input)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	return false
}

func generateCSRFToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}