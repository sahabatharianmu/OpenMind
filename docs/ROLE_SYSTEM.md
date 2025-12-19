# Role-Based Access Control (RBAC) System

## Overview

OpenMind Practice uses **organization-scoped roles** stored in the `organization_members` table. Each user's role is specific to their organization, allowing for proper multi-tenant access control.

## Role Storage

- **Location**: `organization_members.role` column
- **NOT stored in**: `users` table (removed for multi-tenancy support)
- **Set by**: Tenant middleware from `organization_members` table per request

## Available Roles

**Note**: All role constants are defined in `pkg/constants/roles.go` for reuse across the codebase.

### 1. `owner` (`constants.RoleOwner`)
- **Full access** to all features (same as admin)
- Can access:
  - Audit logs
  - Export/Import
  - Invoices (create, update, delete)
  - Organization settings
  - All patient, appointment, and clinical note operations
- **First user** who creates an organization gets this role
- Used to distinguish organization owners from regular admins

### 2. `admin` (`constants.RoleAdmin`)
- **Full access** to all features (same as owner)
- Can access:
  - Audit logs
  - Export/Import
  - Invoices (create, update, delete)
  - Organization settings
  - All patient, appointment, and clinical note operations
- **Granted** to users by organization owners/admins

### 3. `clinician` (`constants.RoleClinician`)
- **Clinical operations** access
- Can access:
  - Create/Edit patients
  - Create/Edit appointments
  - Create/Edit clinical notes
  - View invoices (read-only)
- **Cannot access**:
  - Audit logs
  - Export/Import
  - Invoice management (create/update/delete)
  - Organization settings

### 4. `member` (`constants.RoleMember`)
- **Limited/Read-only** access
- Can access:
  - View patients, appointments, clinical notes (read-only)
  - View invoices (read-only)
- **Cannot access**:
  - Create/Edit operations
  - Admin functions


## How It Works

### Request Flow

1. **Authentication** → JWT middleware validates token
2. **Tenant Middleware** → 
   - Gets user's organization
   - Gets role from `organization_members` table
   - Sets role in request context: `c.Set("role", role)`
3. **RBAC Middleware** → 
   - Reads role from context: `c.Get("role")`
   - Checks if role matches required permissions
   - Allows or denies access

### Example

```go
// In router.go
auditLogs.Use(rbacMiddleware.HasRole("admin"))
// This checks the role from context (set by tenant middleware)
// which comes from organization_members.role
```

## Role Assignment

### During Registration
- First user gets `"owner"` role in `organization_members` table
- Automatically assigned when organization is created
- Owners can then grant `"admin"` role to other users

### Adding New Users
When adding users to an organization, assign appropriate role:
- `"owner"` - Only for the first user who creates the organization (automatically assigned)
- `"admin"` - For organization administrators (granted by owners)
- `"clinician"` - For clinical staff
- `"member"` - For read-only users

## Permission Matrix

| Feature | owner | admin | clinician | member |
|---------|-------|-------|-----------|--------|
| View Patients | ✅ | ✅ | ✅ | ✅ |
| Create/Edit Patients | ✅ | ✅ | ✅ | ❌ |
| Delete Patients | ✅ | ✅ | ❌ | ❌ |
| View Appointments | ✅ | ✅ | ✅ | ✅ |
| Create/Edit Appointments | ✅ | ✅ | ✅ | ❌ |
| View Clinical Notes | ✅ | ✅ | ✅ | ✅ |
| Create/Edit Clinical Notes | ✅ | ✅ | ✅ | ❌ |
| View Invoices | ✅ | ✅ | ✅ | ✅ |
| Manage Invoices | ✅ | ✅ | ❌ | ❌ |
| Audit Logs | ✅ | ✅ | ❌ | ❌ |
| Export/Import | ✅ | ✅ | ❌ | ❌ |
| Organization Settings | ✅ | ✅ | ❌ | ❌ |

## Implementation Details

### Role Constants
All roles are defined as constants in `pkg/constants/roles.go`:
```go
const (
    RoleAdmin     = "admin"
    RoleClinician = "clinician"
    RoleMember    = "member"
    RoleOwner     = "owner"
)
```

### Tenant Middleware
```go
// Sets role from organization_members table
role, err := orgRepo.GetMemberRole(orgID, userID)
c.Set("role", role) // Overrides JWT role with org-specific role
```

### RBAC Middleware
```go
// Reads role from context
role := c.Get("role").(string)
// Admin can access everything
if role == constants.RoleAdmin {
    c.Next(ctx)
    return
}
// Check specific role permissions
```

### Router Usage
```go
// Use constants instead of hardcoded strings
rbacMiddleware.HasRole(constants.RoleAdmin)
rbacMiddleware.HasRole(constants.RoleClinician)
```

## Migration Notes

- **Migration 000013**: Removes `role` column from `users` table
- All existing code now uses role from `organization_members` table
- JWT token still contains role (for backward compatibility), but it's overridden by tenant middleware

## Future Enhancements

1. **Custom Roles**: Support for organization-specific custom roles
2. **Role Permissions**: Fine-grained permission system per role
3. **Multi-Organization Users**: Users with different roles in different organizations

