package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/sahabatharianmu/OpenMind/internal/modules/subscription/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type PublicPlanHandler struct {
	service service.PlanService
}

func NewPublicPlanHandler(service service.PlanService) *PublicPlanHandler {
	return &PublicPlanHandler{
		service: service,
	}
}

func (h *PublicPlanHandler) ListActivePlans(ctx context.Context, c *app.RequestContext) {
	plans, err := h.service.ListActivePlans()
	if err != nil {
		response.InternalServerError(c, "Failed to list plans")
		return
	}

	c.JSON(consts.StatusOK, response.Success("Plans retrieved successfully", plans))
}
