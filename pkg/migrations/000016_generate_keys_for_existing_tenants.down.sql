-- Down migration: Remove encryption keys for tenants
-- WARNING: This will make encrypted data unrecoverable!

DO $$
DECLARE
    tenant_record RECORD;
BEGIN
    -- Remove all tenant encryption keys
    -- WARNING: This will make all encrypted data unrecoverable!
    DELETE FROM tenant_encryption_keys;
    
    RAISE NOTICE 'All tenant encryption keys removed (WARNING: Data may be unrecoverable!)';
END $$;

