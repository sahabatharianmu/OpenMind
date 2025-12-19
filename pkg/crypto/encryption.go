package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/config"
	tenantKeyRepo "github.com/sahabatharianmu/OpenMind/internal/modules/tenant/repository"
)

// EncryptionService handles encryption and decryption operations
// Supports both tenant-specific keys (HIPAA compliant) and legacy shared key
type EncryptionService struct {
	config              *config.Config
	tenantKeyRepository tenantKeyRepo.TenantEncryptionKeyRepository
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(cfg *config.Config) *EncryptionService {
	return &EncryptionService{
		config: cfg,
	}
}

// SetTenantKeyRepository sets the tenant encryption key repository
// This allows the service to retrieve tenant-specific keys
func (s *EncryptionService) SetTenantKeyRepository(repo tenantKeyRepo.TenantEncryptionKeyRepository) {
	s.tenantKeyRepository = repo
}

const KeySize = 32

// getMasterKey returns the master key used to encrypt/decrypt tenant keys
func (s *EncryptionService) getMasterKey() ([]byte, error) {
	key := []byte(s.config.Security.EncryptionKey)
	if len(key) != KeySize {
		return nil, errors.New("master encryption key must be 32 bytes")
	}
	return key, nil
}

// getTenantKey retrieves and decrypts a tenant's encryption key
// Returns the plaintext tenant key for encryption/decryption operations
func (s *EncryptionService) getTenantKey(organizationID uuid.UUID) ([]byte, error) {
	if s.tenantKeyRepository == nil {
		// Fallback to master key if repository not set (backward compatibility)
		return s.getMasterKey()
	}

	// Get encrypted tenant key from database
	tenantKey, err := s.tenantKeyRepository.GetByOrganizationID(organizationID)
	if err != nil {
		// If tenant key doesn't exist, fallback to master key (for backward compatibility)
		// In production, you might want to generate a key here instead
		return s.getMasterKey()
	}

	// Decrypt tenant key using master key
	masterKey, err := s.getMasterKey()
	if err != nil {
		return nil, err
	}

	// Decrypt the tenant key (encrypted key is stored as base64 string in bytes)
	encryptedKeyStr := string(tenantKey.EncryptedKey)
	decryptedKeyBase64, err := s.decryptWithKey(encryptedKeyStr, masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tenant key: %w", err)
	}

	// Decode from base64 to get the actual key bytes
	decryptedKey, err := base64.StdEncoding.DecodeString(decryptedKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tenant key: %w", err)
	}

	return decryptedKey, nil
}

// Encrypt encrypts a plaintext string using tenant-specific key (HIPAA compliant)
// If organizationID is provided, uses tenant-specific key
// Otherwise, falls back to master key (legacy mode)
func (s *EncryptionService) Encrypt(plaintext string, organizationID ...uuid.UUID) (string, error) {
	var key []byte
	var err error

	if len(organizationID) > 0 && organizationID[0] != uuid.Nil {
		// Use tenant-specific key (HIPAA compliant)
		key, err = s.getTenantKey(organizationID[0])
	} else {
		// Fallback to master key (legacy mode)
		key, err = s.getMasterKey()
	}

	if err != nil {
		return "", err
	}

	return s.encryptWithKey(plaintext, key)
}

// Decrypt decrypts a ciphertext string using tenant-specific key (HIPAA compliant)
// If organizationID is provided, uses tenant-specific key
// Otherwise, falls back to master key (legacy mode)
func (s *EncryptionService) Decrypt(ciphertext string, organizationID ...uuid.UUID) (string, error) {
	var key []byte
	var err error

	if len(organizationID) > 0 && organizationID[0] != uuid.Nil {
		// Use tenant-specific key (HIPAA compliant)
		key, err = s.getTenantKey(organizationID[0])
	} else {
		// Fallback to master key (legacy mode)
		key, err = s.getMasterKey()
	}

	if err != nil {
		return "", err
	}

	return s.decryptWithKey(ciphertext, key)
}

// encryptWithKey encrypts using a specific key
func (s *EncryptionService) encryptWithKey(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptWithKey decrypts using a specific key
func (s *EncryptionService) decryptWithKey(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// GenerateTenantKey generates a new encryption key for a tenant
// Returns the plaintext key (should be encrypted with master key before storage)
func (s *EncryptionService) GenerateTenantKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate tenant key: %w", err)
	}
	return key, nil
}

// EncryptTenantKey encrypts a tenant key with the master key for storage
func (s *EncryptionService) EncryptTenantKey(tenantKey []byte) ([]byte, error) {
	masterKey, err := s.getMasterKey()
	if err != nil {
		return nil, err
	}

	// Convert key to base64 string for encryption
	keyBase64 := base64.StdEncoding.EncodeToString(tenantKey)
	encrypted, err := s.encryptWithKey(keyBase64, masterKey)
	if err != nil {
		return nil, err
	}

	// Return as bytes
	return []byte(encrypted), nil
}
