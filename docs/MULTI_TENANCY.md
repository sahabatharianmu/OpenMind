# Multi-Tenancy Implementation Guide

## Overview

OpenMind Practice Cloud Version implements **schema-based multi-tenancy** using PostgreSQL schemas. Each organization (tenant) has its own isolated schema containing all tenant-specific data.

## Architecture

### Schema Isolation Pattern

- **Public Schema**: Contains shared tables (users, organizations, tenants, organization_members)
- **Tenant Schemas**: Each organization has a dedicated schema (e.g., `tenant_abc123...`) containing:
  - `patients`
  - `appointments`
  - `clinical_notes`
  - `invoices`

### How It Works

1. **Tenant Middleware**: After authentication, the middleware:
   - Extracts the user's organization ID
   - Looks up or creates the tenant record
   - Sets PostgreSQL `search_path` to `{tenant_schema}, public` for the request

2. **Automatic Query Routing**: PostgreSQL's `search_path` mechanism automatically routes queries:
   - Tables in tenant schema are accessed first
   - Tables in public schema are accessible as fallback
   - No code changes needed in repositories!

## Components

### 1. Tenant Entity (`internal/modules/tenant/entity/tenant.go`)
Maps organizations to their PostgreSQL schemas.

### 2. Tenant Repository (`internal/modules/tenant/repository/repository.go`)
CRUD operations for tenant records.

### 3. Tenant Service (`internal/modules/tenant/service/service.go`)
- Creates tenant schemas
- Manages schema migrations
- Generates schema names from organization IDs

### 4. Tenant Middleware (`internal/core/middleware/tenant.go`)
- Runs after authentication
- Sets tenant context per request
- Stores schema name in request context

### 5. Database Helpers (`internal/core/database/tenant.go`)
- Schema-qualified table name helpers
- Schema-scoped DB connections
- SQL injection protection

### 6. Migration System
- **000011**: Creates `tenants` table and helper functions
- **000012**: Migrates existing organizations to tenant schemas

## Usage

### Automatic Tenant Context

The tenant middleware is automatically applied to all protected routes. No code changes needed in handlers or repositories!

```go
// In any repository - works automatically!
func (r *patientRepository) List(organizationID uuid.UUID, limit, offset int) ([]entity.Patient, int64, error) {
    // This query automatically uses the tenant schema set by middleware
    query := r.db.Model(&entity.Patient{}).Where("organization_id = ?", organizationID)
    // ...
}
```

### Accessing Tenant Context

```go
import "github.com/sahabatharianmu/OpenMind/internal/core/middleware"

// Get tenant schema from context
schemaName, exists := middleware.GetTenantSchemaFromContext(c)
if !exists {
    // Handle error
}

// Get organization ID from context
orgID, exists := middleware.GetOrganizationIDFromContext(c)
if !exists {
    // Handle error
}
```

### Manual Schema Operations

```go
import "github.com/sahabatharianmu/OpenMind/internal/core/database"

// Get DB instance with schema set
dbWithSchema := database.GetDBWithSchema("tenant_abc123")

// Execute function in specific schema
err := database.ExecuteInSchema(db, "tenant_abc123", func(db *gorm.DB) error {
    // Your queries here automatically use the schema
    return nil
})
```

## Migration Process

### For New Organizations

When a new organization is created:
1. Tenant service automatically creates a tenant record
2. Creates the tenant schema
3. Runs migrations to create tables in the schema

### For Existing Organizations

Run migration `000012` which:
1. Creates tenant schemas for all existing organizations
2. Copies existing data from public schema to tenant schemas
3. Creates tenant records

```bash
# Migration runs automatically on startup
# Or manually:
migrate -path ./pkg/migrations -database "postgres://..." up
```

## Security Considerations

1. **Schema Isolation**: Complete data isolation at the database level
2. **SQL Injection Protection**: Schema names are sanitized before use
3. **Automatic Context**: Middleware ensures every request uses the correct schema
4. **No Cross-Tenant Access**: Impossible to access another tenant's data due to schema isolation

## Performance Considerations

1. **Connection Pooling**: Each request sets `search_path` on the connection
2. **Indexes**: Each tenant schema has its own indexes
3. **Scalability**: Can scale by:
   - Moving tenant schemas to different databases
   - Using read replicas per tenant
   - Partitioning large tenants

## Troubleshooting

### Tenant Not Found
- Check if organization exists
- Verify tenant record exists in `tenants` table
- Check schema exists: `SELECT schema_name FROM tenants WHERE organization_id = '...'`

### Schema Not Set
- Verify tenant middleware is registered in router
- Check middleware runs after authentication
- Verify `search_path` is set: `SHOW search_path`

### Data Not Appearing
- Verify data exists in tenant schema, not public schema
- Check `search_path` includes tenant schema
- Verify organization_id matches tenant's organization_id

## Future Enhancements

1. **Multi-Organization Users**: Support users belonging to multiple organizations
2. **Schema Migration Tooling**: Tools to move schemas between databases
3. **Tenant Analytics**: Cross-tenant analytics (with proper permissions)
4. **Schema Backup/Restore**: Per-tenant backup and restore capabilities

