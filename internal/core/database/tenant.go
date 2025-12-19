package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// WithSchema returns a GORM DB instance scoped to a specific schema
// This sets the search_path for the connection
func WithSchema(db *gorm.DB, schemaName string) *gorm.DB {
	// Sanitize schema name to prevent SQL injection
	schemaName = sanitizeSchemaName(schemaName)
	
	// Set search_path for this session
	return db.Exec(fmt.Sprintf("SET search_path TO %s, public", schemaName))
}

// SchemaQualifiedTableName returns a schema-qualified table name
// Example: "tenant_abc123.patients"
func SchemaQualifiedTableName(schemaName, tableName string) string {
	schemaName = sanitizeSchemaName(schemaName)
	tableName = sanitizeTableName(tableName)
	return fmt.Sprintf("%s.%s", schemaName, tableName)
}

// GetDBWithSchema returns a DB instance with schema set in search_path
// This is the preferred method for tenant-scoped queries
func GetDBWithSchema(schemaName string) *gorm.DB {
	if schemaName == "" {
		return DB
	}
	return WithSchema(DB, schemaName)
}

// sanitizeSchemaName removes potentially dangerous characters from schema name
func sanitizeSchemaName(name string) string {
	// Remove any characters that aren't alphanumeric or underscore
	// PostgreSQL schema names can contain letters, digits, and underscores
	name = strings.ToLower(name)
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// sanitizeTableName removes potentially dangerous characters from table name
func sanitizeTableName(name string) string {
	// Similar to schema name, but we keep it as-is mostly
	// Just ensure it's safe for use in SQL
	name = strings.ToLower(name)
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// TableNameWithSchema is a helper for GORM models to return schema-qualified table names
// Usage in model: func (Model) TableName() string { return database.TableNameWithSchema("schema_name", "table_name") }
func TableNameWithSchema(schemaName, tableName string) string {
	return SchemaQualifiedTableName(schemaName, tableName)
}

