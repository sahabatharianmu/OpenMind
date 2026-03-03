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

// AdminUserHandler provides admin user management
type AdminUserHandler struct {
	db  *gorm.DB
	log logger.Logger
}

// NewAdminUserHandler creates a new admin user handler
func NewAdminUserHandler(db *gorm.DB, log logger.Logger) *AdminUserHandler {
	return &AdminUserHandler{db: db, log: log}
}

// UserListItem represents a user row in the admin list
type UserListItem struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	FullName   string  `json:"full_name"`
	SystemRole string  `json:"system_role"`
	OrgName    *string `json:"org_name"`
	CreatedAt  string  `json:"created_at"`
}

// UserListResponse is the paginated response for /admin/users
type UserListResponse struct {
	Users      []UserListItem `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// ListUsers returns a paginated list of all platform users
func (h *AdminUserHandler) ListUsers(_ context.Context, c *app.RequestContext) {
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
	if err := h.db.Table("users").Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		h.log.Error("Failed to count users", zap.Error(err))
		response.InternalServerError(c, "Failed to count users")
		return
	}

	type userRow struct {
		ID         string  `gorm:"column:id"`
		Email      string  `gorm:"column:email"`
		FullName   string  `gorm:"column:full_name"`
		SystemRole string  `gorm:"column:system_role"`
		OrgName    *string `gorm:"column:org_name"`
		CreatedAt  string  `gorm:"column:created_at"`
	}

	var rows []userRow
	if err := h.db.Table("users u").
		Select("u.id, u.email, u.full_name, u.system_role, o.name as org_name, u.created_at").
		Joins("LEFT JOIN organization_members om ON om.user_id = u.id").
		Joins("LEFT JOIN organizations o ON o.id = om.organization_id AND o.deleted_at IS NULL").
		Where("u.deleted_at IS NULL").
		Order("u.created_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		h.log.Error("Failed to list users", zap.Error(err))
		response.InternalServerError(c, "Failed to list users")
		return
	}

	users := make([]UserListItem, 0, len(rows))
	for _, r := range rows {
		users = append(users, UserListItem{
			ID:         r.ID,
			Email:      r.Email,
			FullName:   r.FullName,
			SystemRole: r.SystemRole,
			OrgName:    r.OrgName,
			CreatedAt:  r.CreatedAt,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	c.JSON(consts.StatusOK, response.Success("Users retrieved", UserListResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}))
}

// UpdateUserRoleRequest is the request body for PATCH /admin/users/:id
type UpdateUserRoleRequest struct {
	SystemRole string `json:"system_role" vd:"required,oneof=user admin"`
}

// UpdateUserRole updates a user's system role
func (h *AdminUserHandler) UpdateUserRole(_ context.Context, c *app.RequestContext) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "User ID is required", nil)
		return
	}

	var req UpdateUserRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	result := h.db.Table("users").
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("system_role", req.SystemRole)

	if result.Error != nil {
		h.log.Error("Failed to update user role", zap.Error(result.Error))
		response.InternalServerError(c, "Failed to update user role")
		return
	}

	if result.RowsAffected == 0 {
		response.NotFound(c, "User not found")
		return
	}

	c.JSON(consts.StatusOK, response.Success("User role updated", nil))
}
