-- Drop indexes
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_schema_name;
DROP INDEX IF EXISTS idx_tenants_organization_id;

-- Drop functions
DROP FUNCTION IF EXISTS drop_tenant_schema(TEXT);
DROP FUNCTION IF EXISTS create_tenant_schema(TEXT);

-- Drop tenants table
DROP TABLE IF EXISTS tenants;

