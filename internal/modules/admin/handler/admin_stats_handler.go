package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	tenantRepo "github.com/sahabatharianmu/OpenMind/internal/modules/tenant/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AdminStatsHandler provides admin dashboard statistics
type AdminStatsHandler struct {
	db         *gorm.DB
	tenantRepo tenantRepo.TenantRepository
	log        logger.Logger
}

// NewAdminStatsHandler creates a new admin stats handler
func NewAdminStatsHandler(db *gorm.DB, tenantRepo tenantRepo.TenantRepository, log logger.Logger) *AdminStatsHandler {
	return &AdminStatsHandler{
		db:         db,
		tenantRepo: tenantRepo,
		log:        log,
	}
}

// StatsResponse holds the admin dashboard statistics
type StatsResponse struct {
	TotalTenants        int64   `json:"total_tenants"`
	TotalUsers          int64   `json:"total_users"`
	ActiveSubscriptions int64   `json:"active_subscriptions"`
	MonthlyRevenue      float64 `json:"monthly_revenue"` // In dollars
}

// GetStats returns platform-wide admin statistics
func (h *AdminStatsHandler) GetStats(_ context.Context, c *app.RequestContext) {
	var stats StatsResponse

	// Count tenants
	_, totalTenants, err := h.tenantRepo.List(1, 0)
	if err != nil {
		h.log.Error("Failed to count tenants", zap.Error(err))
		response.InternalServerError(c, "Failed to get stats")
		return
	}
	stats.TotalTenants = totalTenants

	// Count users
	var userCount int64
	if err := h.db.Table("users").Where("deleted_at IS NULL").Count(&userCount).Error; err != nil {
		h.log.Error("Failed to count users", zap.Error(err))
		response.InternalServerError(c, "Failed to get stats")
		return
	}
	stats.TotalUsers = userCount

	// Count organizations with active subscriptions (non-free tier)
	var activeSubCount int64
	if err := h.db.Table("organizations").
		Where("deleted_at IS NULL AND subscription_tier != ? AND subscription_tier != ?", "free", "").
		Count(&activeSubCount).Error; err != nil {
		h.log.Error("Failed to count active subscriptions", zap.Error(err))
		// Non-critical, continue with 0
	}
	stats.ActiveSubscriptions = activeSubCount

	// Calculate MRR from subscription plans
	var mrrCents int64
	if err := h.db.Table("organizations").
		Select("COALESCE(SUM(sp.price), 0)").
		Joins("LEFT JOIN subscription_plans sp ON organizations.subscription_plan_id = sp.id").
		Where("organizations.deleted_at IS NULL AND organizations.subscription_tier != ? AND organizations.subscription_tier != ?", "free", "").
		Scan(&mrrCents).Error; err != nil {
		h.log.Error("Failed to calculate MRR", zap.Error(err))
		// Non-critical, continue with 0
	}
	stats.MonthlyRevenue = float64(mrrCents) / 100.0

	c.JSON(consts.StatusOK, response.Success("Admin stats retrieved successfully", stats))
}
