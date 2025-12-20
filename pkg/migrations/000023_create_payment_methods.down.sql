-- Migration to drop payment_methods table from all tenant schemas

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
        
        -- Drop payment_methods table from tenant schema
        BEGIN
            EXECUTE format('DROP TABLE IF EXISTS %I.payment_methods CASCADE', tenant_schema_name);
            RAISE NOTICE 'Dropped payment_methods table from schema %', tenant_schema_name;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not drop payment_methods table from %: %', tenant_schema_name, SQLERRM;
        END;
    END LOOP;
END $$;

