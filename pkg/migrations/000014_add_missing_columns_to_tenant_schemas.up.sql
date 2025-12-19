-- Migration to add missing columns to existing tenant schemas
-- This adds columns that were added in later migrations but were missing from tenant schema creation

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
        
        -- Add missing columns to clinical_notes table if they don't exist
        BEGIN
            EXECUTE format('ALTER TABLE %I.clinical_notes ADD COLUMN IF NOT EXISTS icd10_code VARCHAR(20)', tenant_schema_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add icd10_code to %: %', tenant_schema_name, SQLERRM;
        END;
        
        BEGIN
            EXECUTE format('ALTER TABLE %I.clinical_notes ADD COLUMN IF NOT EXISTS key_id VARCHAR(255)', tenant_schema_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add key_id to %: %', tenant_schema_name, SQLERRM;
        END;
        
        BEGIN
            EXECUTE format('ALTER TABLE %I.clinical_notes ADD COLUMN IF NOT EXISTS nonce BYTEA', tenant_schema_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add nonce to %: %', tenant_schema_name, SQLERRM;
        END;
        
        -- Add missing column to appointments table if it doesn't exist
        BEGIN
            EXECUTE format('ALTER TABLE %I.appointments ADD COLUMN IF NOT EXISTS cpt_code VARCHAR(20)', tenant_schema_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add cpt_code to %: %', tenant_schema_name, SQLERRM;
        END;
        
        RAISE NOTICE 'Updated schema % with missing columns', tenant_schema_name;
    END LOOP;
END $$;

