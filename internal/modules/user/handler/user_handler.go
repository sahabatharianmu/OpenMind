package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type UserHandler struct {
	userSvc service.UserService
	authSvc service.AuthService
}

func NewUserHandler(userSvc service.UserService, authSvc service.AuthService) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		authSvc: authSvc,
	}
}

func (h *UserHandler) GetProfile(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	resp, err := h.userSvc.GetProfile(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Profile retrieved successfully", resp))
}

func (h *UserHandler) UpdateProfile(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req dto.UpdateProfileRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.userSvc.UpdateProfile(userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Profile updated successfully", resp))
}

func (h *UserHandler) ChangePassword(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req dto.ChangePasswordRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Password changed successfully", nil))
}
