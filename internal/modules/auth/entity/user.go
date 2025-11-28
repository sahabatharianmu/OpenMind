package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID      `gorm:"primaryKey"                   json:"id"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"               json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"               json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                        json:"-"`
	Email        string         `gorm:"uniqueIndex;not null"         json:"email"`
	PasswordHash string         `gorm:"not null"                     json:"-"` // Never return password hash in JSON
	Role         string         `gorm:"not null;default:'clinician'" json:"role"`
}
