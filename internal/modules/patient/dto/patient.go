package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreatePatientRequest struct {
	FirstName   string  `json:"first_name" validate:"required"`
	LastName    string  `json:"last_name" validate:"required"`
	DateOfBirth string  `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
	Email       *string `json:"email" validate:"omitempty,email"`
	Phone       *string `json:"phone"`
	Address     *string `json:"address"`
	Status      string  `json:"status" validate:"omitempty,oneof=active inactive archived"`
}

type UpdatePatientRequest struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	DateOfBirth string  `json:"date_of_birth" validate:"omitempty,datetime=2006-01-02"`
	Email       *string `json:"email" validate:"omitempty,email"`
	Phone       *string `json:"phone"`
	Address     *string `json:"address"`
	Status      string  `json:"status" validate:"omitempty,oneof=active inactive archived"`
}

type PatientResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	DateOfBirth    string    `json:"date_of_birth"`
	Email          *string   `json:"email"`
	Phone          *string   `json:"phone"`
	Address        *string   `json:"address"`
	Status         string    `json:"status"`
	CreatedBy      uuid.UUID `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
