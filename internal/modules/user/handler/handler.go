package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(_ context.Context, c *app.RequestContext) {
	var req dto.RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	user, err := h.svc.Register(req.Email, req.Password, req.FullName, req.PracticeName)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	resp := dto.RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	response.Created(c, resp, "User registered successfully")
}

func (h *AuthHandler) Login(_ context.Context, c *app.RequestContext) {
	var req dto.LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Login successful", resp))
}

func (h *AuthHandler) SetupStatus(_ context.Context, c *app.RequestContext) {
	resp, err := h.svc.SetupStatus()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Setup status retrieved", resp))
}
