-- Down migration: Remove columns from tenant schemas
-- Note: This is a destructive operation and should be used with caution

DO $$
DECLARE
    tenant_record RECORD;
    tenant_schema_name TEXT;
BEGIN
    -- Loop through all tenant schemas
    FOR tenant_record IN 
        SELECT t.schema_name FROM tenants t WHERE t.status = 'active' AND t.deleted_at IS NULL
    LOOP
        tenant_schema_name := tenant_record.schema_name;
        
        -- Remove columns from clinical_notes table
        BEGIN
            EXECUTE format('ALTER TABLE %I.clinical_notes DROP COLUMN IF EXISTS icd10_code', tenant_schema_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not remove icd10_code from %: %', tenant_schema_name, SQLERRM;
        END;
        
        BEGIN
            EXECUTE format('ALTER TABLE %I.clinical_notes DROP COLUMN IF EXISTS key_id', tenant_schema_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not remove key_id from %: %', tenant_schema_name, SQLERRM;
        END;
        
        BEGIN
            EXECUTE format('ALTER TABLE %I.clinical_notes DROP COLUMN IF EXISTS nonce', tenant_schema_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not remove nonce from %: %', tenant_schema_name, SQLERRM;
        END;
        
        -- Remove column from appointments table
        BEGIN
            EXECUTE format('ALTER TABLE %I.appointments DROP COLUMN IF EXISTS cpt_code', tenant_schema_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not remove cpt_code from %: %', tenant_schema_name, SQLERRM;
        END;
        
        RAISE NOTICE 'Removed columns from schema %', tenant_schema_name;
    END LOOP;
END $$;

