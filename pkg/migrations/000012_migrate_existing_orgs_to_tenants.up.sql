-- Migration script to create tenant schemas for existing organizations
-- and copy existing data to tenant schemas

DO $$
DECLARE
    org_record RECORD;
    schema_name TEXT;
    tenant_id UUID;
BEGIN
    -- Loop through all existing organizations
    FOR org_record IN 
        SELECT id, name FROM organizations WHERE deleted_at IS NULL
    LOOP
        -- Generate schema name from organization ID
        schema_name := 'tenant_' || REPLACE(org_record.id::TEXT, '-', '');
        
        -- Create tenant schema if it doesn't exist
        EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
        
        -- Create tenant record if it doesn't exist
        INSERT INTO tenants (organization_id, schema_name, status, created_at, updated_at)
        VALUES (org_record.id, schema_name, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        ON CONFLICT (organization_id) DO NOTHING;
        
        -- Create tables in tenant schema
        EXECUTE format('
            CREATE TABLE IF NOT EXISTS %I.patients (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                organization_id UUID NOT NULL,
                first_name VARCHAR(255) NOT NULL,
                last_name VARCHAR(255) NOT NULL,
                date_of_birth DATE NOT NULL,
                email VARCHAR(255),
                phone VARCHAR(50),
                address TEXT,
                status VARCHAR(50) NOT NULL DEFAULT ''active'',
                created_by UUID NOT NULL,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
            )
        ', schema_name);
        
        EXECUTE format('
            CREATE TABLE IF NOT EXISTS %I.appointments (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                organization_id UUID NOT NULL,
                patient_id UUID NOT NULL,
                clinician_id UUID NOT NULL,
                start_time TIMESTAMP WITH TIME ZONE NOT NULL,
                end_time TIMESTAMP WITH TIME ZONE NOT NULL,
                status VARCHAR(50) NOT NULL DEFAULT ''scheduled'',
                appointment_type VARCHAR(100) NOT NULL,
                mode VARCHAR(50) NOT NULL,
                cpt_code VARCHAR(20),
                notes TEXT,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                deleted_at TIMESTAMP WITH TIME ZONE
            )
        ', schema_name);
        
        EXECUTE format('
            CREATE TABLE IF NOT EXISTS %I.clinical_notes (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                organization_id UUID NOT NULL,
                patient_id UUID NOT NULL,
                clinician_id UUID NOT NULL,
                appointment_id UUID,
                note_type VARCHAR(100) NOT NULL,
                icd10_code VARCHAR(20),
                subjective TEXT,
                objective TEXT,
                assessment TEXT,
                plan TEXT,
                content_encrypted BYTEA,
                key_id VARCHAR(255),
                nonce BYTEA,
                is_signed BOOLEAN NOT NULL DEFAULT FALSE,
                signed_at TIMESTAMP WITH TIME ZONE,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                deleted_at TIMESTAMP WITH TIME ZONE
            )
        ', schema_name);
        
        EXECUTE format('
            CREATE TABLE IF NOT EXISTS %I.invoices (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                organization_id UUID NOT NULL,
                patient_id UUID NOT NULL,
                appointment_id UUID,
                amount_cents INTEGER NOT NULL,
                status VARCHAR(50) NOT NULL DEFAULT ''pending'',
                due_date TIMESTAMP WITH TIME ZONE,
                paid_at TIMESTAMP WITH TIME ZONE,
                payment_method VARCHAR(50),
                notes TEXT,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                deleted_at TIMESTAMP WITH TIME ZONE
            )
        ', schema_name);
        
        -- Copy existing data from public schema to tenant schema
        -- Only copy data that belongs to this organization
        
        EXECUTE format('
            INSERT INTO %I.patients 
            SELECT * FROM public.patients 
            WHERE organization_id = $1
            ON CONFLICT (id) DO NOTHING
        ', schema_name) USING org_record.id;
        
        EXECUTE format('
            INSERT INTO %I.appointments 
            SELECT * FROM public.appointments 
            WHERE organization_id = $1
            ON CONFLICT (id) DO NOTHING
        ', schema_name) USING org_record.id;
        
        EXECUTE format('
            INSERT INTO %I.clinical_notes 
            SELECT * FROM public.clinical_notes 
            WHERE organization_id = $1
            ON CONFLICT (id) DO NOTHING
        ', schema_name) USING org_record.id;
        
        EXECUTE format('
            INSERT INTO %I.invoices 
            SELECT * FROM public.invoices 
            WHERE organization_id = $1
            ON CONFLICT (id) DO NOTHING
        ', schema_name) USING org_record.id;
        
        RAISE NOTICE 'Migrated organization % (ID: %) to schema %', org_record.name, org_record.id, schema_name;
    END LOOP;
END $$;

