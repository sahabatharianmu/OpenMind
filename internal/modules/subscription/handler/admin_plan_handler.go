package handler

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/subscription/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/subscription/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"gorm.io/datatypes"
)

type AdminPlanHandler struct {
	service service.PlanService
}

func NewAdminPlanHandler(service service.PlanService) *AdminPlanHandler {
	return &AdminPlanHandler{
		service: service,
	}
}

type PlanPriceRequest struct {
	Currency string `json:"currency" validate:"required,len=3"`
	Price    int64  `json:"price"    validate:"gte=0"`
}

type CreatePlanRequest struct {
	Name        string                 `json:"name"        validate:"required"`
	Description string                 `json:"description"`
	Prices      []PlanPriceRequest     `json:"prices"      validate:"required,min=1,dive"`
	Limits      map[string]interface{} `json:"limits"`
	IsActive    bool                   `json:"is_active"`
}

type UpdatePlanRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Prices      []PlanPriceRequest     `json:"prices"      validate:"dive"`
	Limits      map[string]interface{} `json:"limits"`
	IsActive    *bool                  `json:"is_active"`
}

func (h *AdminPlanHandler) CreatePlan(ctx context.Context, c *app.RequestContext) {
	var req CreatePlanRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	var limitsJSON datatypes.JSON
	if req.Limits != nil {
		bytes, err := json.Marshal(req.Limits)
		if err == nil {
			limitsJSON = datatypes.JSON(bytes)
		}
	}

	var prices []entity.SubscriptionPlanPrice
	for _, p := range req.Prices {
		prices = append(prices, entity.SubscriptionPlanPrice{
			Currency: p.Currency,
			Price:    p.Price,
		})
	}

	plan := &entity.SubscriptionPlan{
		Name:        req.Name,
		Description: req.Description,
		Prices:      prices,
		Limits:      limitsJSON,
		IsActive:    req.IsActive,
	}

	if err := h.service.CreatePlan(plan); err != nil {
		response.InternalServerError(c, "Failed to create plan")
		return
	}

	response.Created(c, plan, "Plan created successfully")
}

func (h *AdminPlanHandler) UpdatePlan(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID", nil)
		return
	}

	var req UpdatePlanRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	plan, err := h.service.GetPlan(id)
	if err != nil {
		response.InternalServerError(c, "Failed to get plan")
		return
	}
	if plan == nil {
		response.NotFound(c, "Plan not found")
		return
	}

	if req.Name != "" {
		plan.Name = req.Name
	}
	if req.Description != "" {
		plan.Description = req.Description
	}
	if req.Prices != nil {
		var prices []entity.SubscriptionPlanPrice
		for _, p := range req.Prices {
			prices = append(prices, entity.SubscriptionPlanPrice{
				PlanID:   plan.ID,
				Currency: p.Currency,
				Price:    p.Price,
			})
		}
		// logic to replace prices or update would go here.
		// For now simple assignment for update (Note: this might duplicate if IDs are not managed,
		// but since we are creating new structs they won't have IDs, so GORM will likely insert new ones.
		// Proper way relies on Repository/Service handling association replacement)
		plan.Prices = prices
	}
	if req.Limits != nil {
		bytes, err := json.Marshal(req.Limits)
		if err == nil {
			plan.Limits = datatypes.JSON(bytes)
		}
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}

	if err := h.service.UpdatePlan(plan); err != nil {
		response.InternalServerError(c, "Failed to update plan")
		return
	}

	c.JSON(consts.StatusOK, response.Success("Plan updated successfully", plan))
}

func (h *AdminPlanHandler) ListPlans(ctx context.Context, c *app.RequestContext) {
	plans, err := h.service.ListAllPlans()
	if err != nil {
		response.InternalServerError(c, "Failed to list plans")
		return
	}

	c.JSON(consts.StatusOK, response.Success("Plans retrieved successfully", plans))
}

func (h *AdminPlanHandler) GetPlan(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID", nil)
		return
	}

	plan, err := h.service.GetPlan(id)
	if err != nil {
		response.InternalServerError(c, "Failed to get plan")
		return
	}
	if plan == nil {
		response.NotFound(c, "Plan not found")
		return
	}

	c.JSON(consts.StatusOK, response.Success("Plan retrieved successfully", plan))
}
