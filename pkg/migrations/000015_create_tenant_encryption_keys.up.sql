-- Migration to create tenant_encryption_keys table
-- This table stores encryption keys for each tenant, encrypted with a master key
-- This ensures HIPAA compliance: platform cannot decrypt tenant data

CREATE TABLE IF NOT EXISTS tenant_encryption_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    -- Encrypted key (encrypted with master key, stored as base64)
    encrypted_key BYTEA NOT NULL,
    -- Key version for key rotation support
    key_version INTEGER NOT NULL DEFAULT 1,
    -- Key metadata
    algorithm VARCHAR(50) NOT NULL DEFAULT 'AES-256-GCM',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Index for fast lookups
CREATE INDEX IF NOT EXISTS idx_tenant_encryption_keys_tenant_id ON tenant_encryption_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_encryption_keys_organization_id ON tenant_encryption_keys(organization_id);
CREATE INDEX IF NOT EXISTS idx_tenant_encryption_keys_deleted_at ON tenant_encryption_keys(deleted_at);

-- Add comment explaining the security model
COMMENT ON TABLE tenant_encryption_keys IS 'Stores tenant-specific encryption keys encrypted with master key. Platform cannot decrypt tenant data without tenant key. HIPAA compliant zero-knowledge encryption.';

