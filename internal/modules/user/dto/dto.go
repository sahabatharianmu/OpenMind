package dto

import "github.com/google/uuid"

type RegisterRequest struct {
	Email        string `json:"email"         binding:"required,email"`
	Password     string `json:"password"      binding:"required,min=8"`
	FullName     string `json:"full_name"     binding:"required,min=2"`
	PracticeName string `json:"practice_name" binding:"required,min=2"`
}

type RegisterResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SetupStatusResponse struct {
	IsSetupRequired bool `json:"is_setup_required"`
	HasUsers        bool `json:"has_users"`
}
