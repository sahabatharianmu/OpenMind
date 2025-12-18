package entity

import (
	"time"

	"github.com/google/uuid"
)

type Patient struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"                              json:"organization_id"`
	FirstName      string    `gorm:"not null"                                        json:"first_name"`
	LastName       string    `gorm:"not null"                                        json:"last_name"`
	DateOfBirth    time.Time `gorm:"type:date;not null"                              json:"date_of_birth"`
	Email          *string   `gorm:""                                                json:"email"`
	Phone          *string   `gorm:""                                                json:"phone"`
	Address        *string   `gorm:""                                                json:"address"`
	Status         string    `gorm:"not null;default:'active'"                       json:"status"`
	CreatedBy      uuid.UUID `gorm:"type:uuid;not null"                              json:"created_by"`
	CreatedAt      time.Time `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (Patient) TableName() string {
	return "patients"
}
