-- Migration to create payment_methods table in all tenant schemas
-- This table stores encrypted payment method tokens for organizations

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
        
        -- Create payment_methods table in tenant schema
        BEGIN
            EXECUTE format('
                CREATE TABLE IF NOT EXISTS %I.payment_methods (
                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                    organization_id UUID NOT NULL,
                    provider VARCHAR(50) NOT NULL,
                    encrypted_token BYTEA NOT NULL,
                    provider_payment_method_id VARCHAR(255) NOT NULL,
                    last4 VARCHAR(4) NOT NULL,
                    brand VARCHAR(50) NOT NULL,
                    expiry_month INTEGER NOT NULL,
                    expiry_year INTEGER NOT NULL,
                    is_default BOOLEAN NOT NULL DEFAULT FALSE,
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                    deleted_at TIMESTAMP WITH TIME ZONE,
                    CONSTRAINT check_payment_provider CHECK (provider IN (''stripe'', ''square'')),
                    CONSTRAINT check_expiry_month CHECK (expiry_month >= 1 AND expiry_month <= 12),
                    CONSTRAINT check_expiry_year CHECK (expiry_year >= 2000 AND expiry_year <= 2100)
                )
            ', tenant_schema_name);
            
            -- Create indexes
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_payment_methods_organization_id 
                ON %I.payment_methods(organization_id)
            ', tenant_schema_name);
            
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_payment_methods_is_default 
                ON %I.payment_methods(organization_id, is_default) 
                WHERE is_default = TRUE
            ', tenant_schema_name);
            
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_payment_methods_provider 
                ON %I.payment_methods(provider)
            ', tenant_schema_name);
            
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_payment_methods_deleted_at 
                ON %I.payment_methods(deleted_at)
            ', tenant_schema_name);
            
            RAISE NOTICE 'Created payment_methods table in schema %', tenant_schema_name;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not create payment_methods table in %: %', tenant_schema_name, SQLERRM;
        END;
    END LOOP;
END $$;

