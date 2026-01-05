package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantEncryptionKey represents an encryption key for a tenant
// The key itself is encrypted with a master key (zero-knowledge encryption)
type TenantEncryptionKey struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"                  json:"tenant_id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"                  json:"organization_id"`
	EncryptedKey   []byte         `gorm:"type:bytea;not null"                             json:"-"` // Encrypted with master key
	KeyVersion     int            `gorm:"not null;default:1"                              json:"key_version"`
	Algorithm      string         `gorm:"type:varchar(50);not null;default:'AES-256-GCM'" json:"algorithm"`
	CreatedAt      time.Time      `                                                       json:"created_at"`
	UpdatedAt      time.Time      `                                                       json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                                           json:"deleted_at,omitempty"`
}

func (TenantEncryptionKey) TableName() string {
	return "tenant_encryption_keys"
}
