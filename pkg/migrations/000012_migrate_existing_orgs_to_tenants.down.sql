-- Rollback migration: Drop tenant schemas and remove tenant records
-- WARNING: This will delete all tenant data!

DO $$
DECLARE
    tenant_record RECORD;
BEGIN
    -- Loop through all tenants and drop their schemas
    FOR tenant_record IN 
        SELECT schema_name FROM tenants WHERE deleted_at IS NULL
    LOOP
        -- Drop tenant schema (CASCADE will drop all tables)
        EXECUTE format('DROP SCHEMA IF EXISTS %I CASCADE', tenant_record.schema_name);
        RAISE NOTICE 'Dropped tenant schema %', tenant_record.schema_name;
    END LOOP;
    
    -- Delete all tenant records
    DELETE FROM tenants;
END $$;

