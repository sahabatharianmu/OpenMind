package crypto

import (
	"fmt"

	"github.com/sahabatharianmu/OpenMind/config"
	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles password hashing and verification
type PasswordService struct {
	config *config.Config
}

// NewPasswordService creates a new password service
func NewPasswordService(cfg *config.Config) *PasswordService {
	return &PasswordService{
		config: cfg,
	}
}

// HashPassword hashes a plain text password
func (s *PasswordService) HashPassword(password string) (string, error) {
	// Use the configured hash cost
	hashCost := s.config.Security.HashCost
	if hashCost == 0 {
		hashCost = 10 // Default cost if not configured
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against a hash
func (s *PasswordService) VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	return nil
}

// ValidatePassword validates password against security requirements
func (s *PasswordService) ValidatePassword(password string) error {
	cfg := s.config.Security

	// Check minimum length
	if len(password) < cfg.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters long", cfg.PasswordMinLength)
	}

	// Check for uppercase letter
	if cfg.PasswordRequireUpper {
		hasUpper := false
		for _, char := range password {
			if char >= 'A' && char <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
	}

	// Check for lowercase letter
	if cfg.PasswordRequireLower {
		hasLower := false
		for _, char := range password {
			if char >= 'a' && char <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return fmt.Errorf("password must contain at least one lowercase letter")
		}
	}

	// Check for number
	if cfg.PasswordRequireNumber {
		hasNumber := false
		for _, char := range password {
			if char >= '0' && char <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return fmt.Errorf("password must contain at least one number")
		}
	}

	// Check for special character
	if cfg.PasswordRequireSpecial {
		hasSpecial := false
		specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		for _, char := range password {
			for _, special := range specialChars {
				if char == special {
					hasSpecial = true
					break
				}
			}
			if hasSpecial {
				break
			}
		}
		if !hasSpecial {
			return fmt.Errorf("password must contain at least one special character")
		}
	}

	return nil
}
