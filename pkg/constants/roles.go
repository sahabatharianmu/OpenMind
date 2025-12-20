package constants

// User roles in the organization (stored in organization_members table)
const (
	// RoleAdmin - Full access to all features including:
	// - Audit logs
	// - Export/Import
	// - Invoice management
	// - Organization settings
	// - All patient, appointment, and clinical note operations
	RoleAdmin = "admin"

	// RoleClinician - Clinical operations access:
	// - Create/Edit patients, appointments, clinical notes
	// - View invoices (read-only)
	// - Cannot access: audit logs, export/import, invoice management, org settings
	RoleClinician = "clinician"

	// RoleMember - Limited/Read-only access:
	// - View patients, appointments, clinical notes (read-only)
	// - View invoices (read-only)
	// - Cannot create/edit or access admin functions
	RoleMember = "member"

	// RoleCaseManager - Case management access:
	// - Similar to clinician but focused on case management
	// - Can manage patient cases and coordination
	// - View invoices (read-only)
	// - Cannot access: audit logs, export/import, invoice management, org settings
	RoleCaseManager = "case_manager"

	// RoleOwner - Full access (same as admin)
	// Assigned to the first user who creates the organization
	// Used to distinguish organization owners from regular admins
	RoleOwner = "owner"
)

// AllRoles returns a slice of all valid roles
func AllRoles() []string {
	return []string{
		RoleOwner,
		RoleAdmin,
		RoleClinician,
		RoleCaseManager,
		RoleMember,
	}
}

// IsValidRole checks if a role string is valid
func IsValidRole(role string) bool {
	for _, r := range AllRoles() {
		if r == role {
			return true
		}
	}
	return false
}

