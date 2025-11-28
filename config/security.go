package config

import (
	"fmt"
	"strings"
	"time"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// CORS settings
	CORSAllowOrigins []string `mapstructure:"cors_allow_origins"`
	CORSAllowMethods []string `mapstructure:"cors_allow_methods"`
	CORSAllowHeaders []string `mapstructure:"cors_allow_headers"`
	CORSMaxAge       int      `mapstructure:"cors_max_age"`

	// Rate limiting
	RateLimitRequests int           `mapstructure:"rate_limit_requests"`
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window"`

	// JWT settings
	JWTSecretKey     string        `mapstructure:"jwt_secret_key"`
	JWTAccessExpiry  time.Duration `mapstructure:"jwt_access_expiry"`
	JWTRefreshExpiry time.Duration `mapstructure:"jwt_refresh_expiry"`

	// Password settings
	PasswordMinLength      int  `mapstructure:"password_min_length"`
	PasswordRequireUpper   bool `mapstructure:"password_require_upper"`
	PasswordRequireLower   bool `mapstructure:"password_require_lower"`
	PasswordRequireNumber  bool `mapstructure:"password_require_number"`
	PasswordRequireSpecial bool `mapstructure:"password_require_special"`

	// Session settings
	SessionTimeout  time.Duration `mapstructure:"session_timeout"`
	SessionSecure   bool          `mapstructure:"session_secure"`
	SessionHTTPOnly bool          `mapstructure:"session_httponly"`

	// File upload security
	MaxFileSize      int64    `mapstructure:"max_file_size"`
	AllowedFileTypes []string `mapstructure:"allowed_file_types"`

	// API security
	APIKeyHeader string `mapstructure:"api_key_header"`
	EnableAPIKey bool   `mapstructure:"enable_api_key"`
	EnableCSRF   bool   `mapstructure:"enable_csrf"`
	EnableCSP    bool   `mapstructure:"enable_csp"`

	// Database security
	DBMaxOpenConns    int           `mapstructure:"db_max_open_conns"`
	DBMaxIdleConns    int           `mapstructure:"db_max_idle_conns"`
	DBConnMaxLifetime time.Duration `mapstructure:"db_conn_max_lifetime"`

	// Logging and monitoring
	EnableSecurityLogging bool `mapstructure:"enable_security_logging"`
	LogSecurityEvents     bool `mapstructure:"log_security_events"`

	// Encryption settings
	EncryptionKey string `mapstructure:"encryption_key"`
	HashCost      int    `mapstructure:"hash_cost"`

	// Request limits
	MaxRequestSize    int64         `mapstructure:"max_request_size"`
	RequestTimeout    time.Duration `mapstructure:"request_timeout"`
	MaxRequestPerIP   int           `mapstructure:"max_request_per_ip"`
	MaxRequestPerUser int           `mapstructure:"max_request_per_user"`
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		// CORS settings
		CORSAllowOrigins: []string{"http://localhost:3000", "https://smatax.id"},
		CORSAllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		CORSAllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "X-Request-ID"},
		CORSMaxAge:       86400, // 24 hours

		// Rate limiting
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,

		// JWT settings
		JWTSecretKey:     "your-secret-key-change-this-in-production",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,

		// Password settings
		PasswordMinLength:      8,
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireNumber:  true,
		PasswordRequireSpecial: true,

		// Session settings
		SessionTimeout:  30 * time.Minute,
		SessionSecure:   true,
		SessionHTTPOnly: true,

		// File upload security
		MaxFileSize:      5 * 1024 * 1024, // 5MB
		AllowedFileTypes: []string{".pdf", ".jpg", ".jpeg", ".png", ".xlsx", ".csv"},

		// API security
		APIKeyHeader: "X-API-Key",
		EnableAPIKey: false,
		EnableCSRF:   true,
		EnableCSP:    true,

		// Database security
		DBMaxOpenConns:    25,
		DBMaxIdleConns:    5,
		DBConnMaxLifetime: time.Hour,

		// Logging and monitoring
		EnableSecurityLogging: true,
		LogSecurityEvents:     true,

		// Encryption settings
		EncryptionKey: "your-encryption-key-change-this-in-production",
		HashCost:      12, // bcrypt cost factor

		// Request limits
		MaxRequestSize:    10 * 1024 * 1024, // 10MB
		RequestTimeout:    30 * time.Second,
		MaxRequestPerIP:   1000,
		MaxRequestPerUser: 100,
	}
}

// Validate validates security configuration
func (s *SecurityConfig) Validate() error {
	if s.PasswordMinLength < 6 {
		return fmt.Errorf("password minimum length must be at least 6")
	}

	if s.JWTAccessExpiry < 5*time.Minute {
		return fmt.Errorf("JWT access expiry must be at least 5 minutes")
	}

	if s.MaxFileSize > 50*1024*1024 {
		return fmt.Errorf("max file size cannot exceed 50MB")
	}

	if s.RateLimitRequests < 10 {
		return fmt.Errorf("rate limit requests must be at least 10")
	}

	if s.HashCost < 10 || s.HashCost > 20 {
		return fmt.Errorf("hash cost must be between 10 and 20")
	}

	return nil
}

// GetAllowedFileExtensions returns map of allowed file extensions
func (s *SecurityConfig) GetAllowedFileExtensions() map[string]bool {
	extensions := make(map[string]bool)
	for _, ext := range s.AllowedFileTypes {
		extensions[strings.ToLower(ext)] = true
	}
	return extensions
}

// IsValidPassword checks if password meets security requirements
func (s *SecurityConfig) IsValidPassword(password string) bool {
	if len(password) < s.PasswordMinLength {
		return false
	}

	if s.PasswordRequireUpper && !containsUpper(password) {
		return false
	}

	if s.PasswordRequireLower && !containsLower(password) {
		return false
	}

	if s.PasswordRequireNumber && !containsNumber(password) {
		return false
	}

	if s.PasswordRequireSpecial && !containsSpecial(password) {
		return false
	}

	return true
}

// Helper functions for password validation
func containsUpper(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func containsLower(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsSpecial(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			return true
		}
	}
	return false
}
