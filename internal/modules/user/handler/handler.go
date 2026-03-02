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
	svc     service.AuthService
	baseURL string
}

func NewAuthHandler(svc service.AuthService, baseURL string) *AuthHandler {
	return &AuthHandler{
		svc:     svc,
		baseURL: baseURL,
	}
}

func (h *AuthHandler) Register(_ context.Context, c *app.RequestContext) {
	var req dto.RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Register(req.Email, req.Password, req.FullName, req.PracticeName, h.baseURL)
	if err != nil {
		response.HandleError(c, err)
		return
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

func (h *AuthHandler) ForgotPassword(_ context.Context, c *app.RequestContext) {
	var req dto.ForgotPasswordRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Always return success to prevent email enumeration
	_ = h.svc.ForgotPassword(req.Email, h.baseURL)

	c.JSON(consts.StatusOK, response.Success("If an account with that email exists, a password reset link has been sent", nil))
}

func (h *AuthHandler) ResetPassword(_ context.Context, c *app.RequestContext) {
	var req dto.ResetPasswordRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := h.svc.ResetPassword(req.Token, req.NewPassword); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Password has been reset successfully", nil))
}

func (h *AuthHandler) RefreshToken(_ context.Context, c *app.RequestContext) {
	var req dto.RefreshTokenRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.RefreshToken(req.RefreshToken)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Token refreshed successfully", resp))
}
