package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	orgRepo "github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	tenantService "github.com/sahabatharianmu/OpenMind/internal/modules/tenant/service"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

// TenantContextKey is the key for storing tenant information in context
type TenantContextKey string

const (
	TenantSchemaKey TenantContextKey = "tenant_schema"
	TenantIDKey     TenantContextKey = "tenant_id"
	OrgIDKey        TenantContextKey = "organization_id"
)

// TenantContextMiddleware creates middleware that sets tenant context per request
func TenantContextMiddleware(tenantSvc tenantService.TenantService, orgRepo orgRepo.OrganizationRepository, log logger.Logger) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get user ID from context (set by auth middleware)
		userIDVal, exists := c.Get("userID")
		if !exists {
			// If no user ID, this might be a public endpoint - skip tenant context
			c.Next(ctx)
			return
		}

		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			log.Warn("Invalid user ID type in context")
			response.Unauthorized(c, "Invalid user context")
			c.Abort()
			return
		}

		// Get organization ID from user
		// We need to get the organization the user belongs to
		// This assumes the user belongs to one organization
		// For multi-org support, you'd need to get org from request header/param
		org, err := orgRepo.GetByUserID(userID)
		if err != nil {
			log.Error("Failed to get organization for user", zap.Error(err), zap.String("user_id", userID.String()))
			response.InternalServerError(c, "Failed to determine organization")
			c.Abort()
			return
		}

		if org == nil {
			log.Error("Organization not found for user", zap.String("user_id", userID.String()))
			response.InternalServerError(c, "Organization not found")
			c.Abort()
			return
		}

		orgID := org.ID
		if orgID == uuid.Nil {
			log.Error("Organization ID is nil", zap.String("user_id", userID.String()))
			response.InternalServerError(c, "Invalid organization")
			c.Abort()
			return
		}

		// Get tenant for this organization
		tenant, err := tenantSvc.GetTenantByOrganizationID(ctx, orgID)
		if err != nil {
			log.Error("Failed to get tenant for organization", zap.Error(err), zap.String("organization_id", orgID.String()))
			// If tenant doesn't exist, create it
			tenant, err = tenantSvc.CreateTenantForOrganization(ctx, orgID)
			if err != nil {
				log.Error("Failed to create tenant for organization", zap.Error(err), zap.String("organization_id", orgID.String()))
				response.InternalServerError(c, "Failed to initialize tenant")
				c.Abort()
				return
			}
		}

		// Get user's role from organization_members
		role, err := orgRepo.GetMemberRole(orgID, userID)
		if err != nil {
			log.Error("Failed to get user role from organization", zap.Error(err),
				zap.String("organization_id", orgID.String()),
				zap.String("user_id", userID.String()))
			response.InternalServerError(c, "Failed to determine user role")
			c.Abort()
			return
		}

		// Set schema for this request
		if err := tenantSvc.SetSchemaForRequest(ctx, tenant.SchemaName); err != nil {
			log.Error("Failed to set tenant schema", zap.Error(err), zap.String("schema_name", tenant.SchemaName))
			response.InternalServerError(c, "Failed to set tenant context")
			c.Abort()
			return
		}

		// Store tenant information and role in context
		// Role is from organization_members table (organization-specific)
		// This role is used by RBAC middleware for all permission checks
		c.Set(string(TenantSchemaKey), tenant.SchemaName)
		c.Set(string(TenantIDKey), tenant.ID)
		c.Set(string(OrgIDKey), orgID)
		c.Set("role", role) // Organization-specific role from organization_members table

		log.Debug("Tenant context set", zap.String("schema_name", tenant.SchemaName), zap.String("organization_id", orgID.String()))

		c.Next(ctx)
	}
}

// GetTenantSchemaFromContext retrieves the tenant schema name from context
func GetTenantSchemaFromContext(c *app.RequestContext) (string, bool) {
	if schemaName, exists := c.Get(string(TenantSchemaKey)); exists {
		if schema, ok := schemaName.(string); ok {
			return schema, true
		}
	}
	return "", false
}

// GetOrganizationIDFromContext retrieves the organization ID from context
func GetOrganizationIDFromContext(c *app.RequestContext) (uuid.UUID, bool) {
	if orgIDVal, exists := c.Get(string(OrgIDKey)); exists {
		if orgID, ok := orgIDVal.(uuid.UUID); ok {
			return orgID, true
		}
	}
	return uuid.Nil, false
}
