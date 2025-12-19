-- Migration placeholder for tenant encryption keys
-- Note: This migration only ensures the table exists.
-- Actual key generation is handled by the application on startup
-- via TenantService.GenerateKeysForExistingTenants()
-- 
-- This ensures:
-- 1. Keys are properly encrypted with the master key from application config
-- 2. Keys are generated using cryptographically secure random
-- 3. Keys are managed by the application, not SQL

-- The table should already exist from migration 000015
-- This migration is here for documentation and to ensure proper ordering

-- Application will automatically generate keys for existing tenants on startup
-- See: internal/modules/tenant/service/key_generation.go

