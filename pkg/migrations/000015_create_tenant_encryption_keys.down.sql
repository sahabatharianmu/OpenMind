-- Down migration: Drop tenant_encryption_keys table

DROP INDEX IF EXISTS idx_tenant_encryption_keys_deleted_at;
DROP INDEX IF EXISTS idx_tenant_encryption_keys_organization_id;
DROP INDEX IF EXISTS idx_tenant_encryption_keys_tenant_id;

DROP TABLE IF EXISTS tenant_encryption_keys;

