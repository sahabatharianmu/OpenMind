-- Tenant schema: Create payment_methods table
-- Source: extracted from 000023 DO-loop, converted to direct DDL

CREATE TABLE IF NOT EXISTS payment_methods (
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
    CONSTRAINT check_payment_provider CHECK (provider IN ('stripe', 'square')),
    CONSTRAINT check_expiry_month CHECK (expiry_month >= 1 AND expiry_month <= 12),
    CONSTRAINT check_expiry_year CHECK (expiry_year >= 2000 AND expiry_year <= 2100)
);

CREATE INDEX IF NOT EXISTS idx_payment_methods_organization_id
    ON payment_methods(organization_id);

CREATE INDEX IF NOT EXISTS idx_payment_methods_is_default
    ON payment_methods(organization_id, is_default)
    WHERE is_default = TRUE;

CREATE INDEX IF NOT EXISTS idx_payment_methods_provider
    ON payment_methods(provider);

CREATE INDEX IF NOT EXISTS idx_payment_methods_deleted_at
    ON payment_methods(deleted_at);
