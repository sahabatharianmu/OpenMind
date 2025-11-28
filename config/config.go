package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Application ApplicationConfig `mapstructure:"application"`
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Security    SecurityConfig    `mapstructure:"security"`
	Email       EmailConfig       `mapstructure:"email"`
	SMS         SMSConfig         `mapstructure:"sms"`
	Storage     StorageConfig     `mapstructure:"storage"`
	Payment     PaymentConfig     `mapstructure:"payment"`
}

// ApplicationConfig holds application configuration
type ApplicationConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	ExitTimeout  time.Duration `mapstructure:"exit_timeout"`
	Mode         string        `mapstructure:"mode"` // development, staging, production
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"db_name"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// SecurityConfig is defined in security.go file

// EmailConfig holds email service configuration
type EmailConfig struct {
	Provider  string         `mapstructure:"provider"` // sendgrid, aws_ses, smtp
	FromEmail string         `mapstructure:"from_email"`
	FromName  string         `mapstructure:"from_name"`
	SendGrid  SendGridConfig `mapstructure:"sendgrid"`
	AWSES     AWSESConfig    `mapstructure:"aws_es"`
	SMTP      SMTPConfig     `mapstructure:"smtp"`
}

// SendGridConfig holds SendGrid configuration
type SendGridConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// AWSESConfig holds AWS SES configuration
type AWSESConfig struct {
	Region    string `mapstructure:"region"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	TLS      bool   `mapstructure:"tls"`
}

// SMSConfig holds SMS service configuration
type SMSConfig struct {
	Provider string       `mapstructure:"provider"` // twilio, local
	Twilio   TwilioConfig `mapstructure:"twilio"`
}

// TwilioConfig holds Twilio configuration
type TwilioConfig struct {
	AccountSID string `mapstructure:"account_sid"`
	AuthToken  string `mapstructure:"auth_token"`
	FromNumber string `mapstructure:"from_number"`
}

// StorageConfig holds file storage configuration
type StorageConfig struct {
	Provider   string           `mapstructure:"provider"` // aws_s3, gcp_storage, local
	AWSS3      AWSS3Config      `mapstructure:"aws_s3"`
	GCPStorage GCPStorageConfig `mapstructure:"gcp_storage"`
	Local      LocalConfig      `mapstructure:"local"`
}

// AWSS3Config holds AWS S3 configuration
type AWSS3Config struct {
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// GCPStorageConfig holds Google Cloud Storage configuration
type GCPStorageConfig struct {
	Bucket  string `mapstructure:"bucket"`
	KeyFile string `mapstructure:"key_file"`
}

// LocalConfig holds local storage configuration
type LocalConfig struct {
	Path string `mapstructure:"path"`
}

// PaymentConfig holds payment gateway configuration
type PaymentConfig struct {
	Provider string         `mapstructure:"provider"` // xendit, midtrans, doku
	Xendit   XenditConfig   `mapstructure:"xendit"`
	Midtrans MidtransConfig `mapstructure:"midtrans"`
	Doku     DokuConfig     `mapstructure:"doku"`
}

// XenditConfig holds Xendit configuration
type XenditConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	PublicKey string `mapstructure:"public_key"`
}

// MidtransConfig holds Midtrans configuration
type MidtransConfig struct {
	ClientKey    string `mapstructure:"client_key"`
	ServerKey    string `mapstructure:"server_key"`
	IsProduction bool   `mapstructure:"is_production"`
}

// DokuConfig holds Doku configuration
type DokuConfig struct {
	ClientID     string `mapstructure:"client_id"`
	SecretKey    string `mapstructure:"secret_key"`
	IsProduction bool   `mapstructure:"is_production"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/openmind/")

	// Set environment variable prefix
	viper.SetEnvPrefix("OPENMIND")
	viper.AutomaticEnv()

	// Set default values
	setDefaults()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Application defaults
	viper.SetDefault("application.name", "OpenMind")
	viper.SetDefault("application.version", "0.0.1")
	viper.SetDefault("application.environment", "development")

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.exit_timeout", "60s")
	viper.SetDefault("server.mode", viper.GetString("application.environment"))

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.db_name", "smetax")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Security defaults
	securityConfig := DefaultSecurityConfig()
	viper.SetDefault("security.cors_allow_origins", securityConfig.CORSAllowOrigins)
	viper.SetDefault("security.cors_allow_methods", securityConfig.CORSAllowMethods)
	viper.SetDefault("security.cors_allow_headers", securityConfig.CORSAllowHeaders)
	viper.SetDefault("security.cors_max_age", securityConfig.CORSMaxAge)
	viper.SetDefault("security.rate_limit_requests", securityConfig.RateLimitRequests)
	viper.SetDefault("security.rate_limit_window", securityConfig.RateLimitWindow)
	viper.SetDefault("security.jwt_secret_key", securityConfig.JWTSecretKey)
	viper.SetDefault("security.jwt_access_expiry", securityConfig.JWTAccessExpiry)
	viper.SetDefault("security.jwt_refresh_expiry", securityConfig.JWTRefreshExpiry)
	viper.SetDefault("security.password_min_length", securityConfig.PasswordMinLength)
	viper.SetDefault("security.password_require_upper", securityConfig.PasswordRequireUpper)
	viper.SetDefault("security.password_require_lower", securityConfig.PasswordRequireLower)
	viper.SetDefault("security.password_require_number", securityConfig.PasswordRequireNumber)
	viper.SetDefault("security.password_require_special", securityConfig.PasswordRequireSpecial)
	viper.SetDefault("security.session_timeout", securityConfig.SessionTimeout)
	viper.SetDefault("security.session_secure", securityConfig.SessionSecure)
	viper.SetDefault("security.session_httponly", securityConfig.SessionHTTPOnly)
	viper.SetDefault("security.max_file_size", securityConfig.MaxFileSize)
	viper.SetDefault("security.allowed_file_types", securityConfig.AllowedFileTypes)
	viper.SetDefault("security.api_key_header", securityConfig.APIKeyHeader)
	viper.SetDefault("security.enable_api_key", securityConfig.EnableAPIKey)
	viper.SetDefault("security.enable_csrf", securityConfig.EnableCSRF)
	viper.SetDefault("security.enable_csp", securityConfig.EnableCSP)
	viper.SetDefault("security.db_max_open_conns", securityConfig.DBMaxOpenConns)
	viper.SetDefault("security.db_max_idle_conns", securityConfig.DBMaxIdleConns)
	viper.SetDefault("security.db_conn_max_lifetime", securityConfig.DBConnMaxLifetime)
	viper.SetDefault("security.enable_security_logging", securityConfig.EnableSecurityLogging)
	viper.SetDefault("security.log_security_events", securityConfig.LogSecurityEvents)
	viper.SetDefault("security.encryption_key", securityConfig.EncryptionKey)
	viper.SetDefault("security.hash_cost", securityConfig.HashCost)
	viper.SetDefault("security.max_request_size", securityConfig.MaxRequestSize)
	viper.SetDefault("security.request_timeout", securityConfig.RequestTimeout)
	viper.SetDefault("security.max_request_per_ip", securityConfig.MaxRequestPerIP)
	viper.SetDefault("security.max_request_per_user", securityConfig.MaxRequestPerUser)
}
