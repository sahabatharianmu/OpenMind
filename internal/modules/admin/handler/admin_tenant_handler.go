package handler

import (
	"context"
	"math"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AdminTenantHandler provides admin tenant management
type AdminTenantHandler struct {
	db  *gorm.DB
	log logger.Logger
}

// NewAdminTenantHandler creates a new admin tenant handler
func NewAdminTenantHandler(db *gorm.DB, log logger.Logger) *AdminTenantHandler {
	return &AdminTenantHandler{db: db, log: log}
}

// TenantListItem represents a tenant row in the admin list
type TenantListItem struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	SubscriptionTier string `json:"subscription_tier"`
	MemberCount      int64  `json:"member_count"`
	CreatedAt        string `json:"created_at"`
}

// TenantListResponse is the paginated response for /admin/tenants
type TenantListResponse struct {
	Tenants    []TenantListItem `json:"tenants"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PerPage    int              `json:"per_page"`
	TotalPages int              `json:"total_pages"`
}

// ListTenants returns a paginated list of organizations for admin
func (h *AdminTenantHandler) ListTenants(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(string(c.Query("per_page")))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	var total int64
	if err := h.db.Table("organizations").Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		h.log.Error("Failed to count organizations", zap.Error(err))
		response.InternalServerError(c, "Failed to count organizations")
		return
	}

	type orgRow struct {
		ID               string `gorm:"column:id"`
		Name             string `gorm:"column:name"`
		SubscriptionTier string `gorm:"column:subscription_tier"`
		CreatedAt        string `gorm:"column:created_at"`
	}

	var rows []orgRow
	if err := h.db.Table("organizations").
		Select("id, name, subscription_tier, created_at").
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		h.log.Error("Failed to list organizations", zap.Error(err))
		response.InternalServerError(c, "Failed to list organizations")
		return
	}

	tenants := make([]TenantListItem, 0, len(rows))
	for _, r := range rows {
		var memberCount int64
		h.db.Table("organization_members").Where("organization_id = ?", r.ID).Count(&memberCount)

		tenants = append(tenants, TenantListItem{
			ID:               r.ID,
			Name:             r.Name,
			SubscriptionTier: r.SubscriptionTier,
			MemberCount:      memberCount,
			CreatedAt:        r.CreatedAt,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	c.JSON(consts.StatusOK, response.Success("Tenants retrieved", TenantListResponse{
		Tenants:    tenants,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}))
}
