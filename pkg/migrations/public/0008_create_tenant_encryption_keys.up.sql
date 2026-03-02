-- Public schema: Create tenant_encryption_keys table
-- Source: original 000015_create_tenant_encryption_keys.up.sql

CREATE TABLE IF NOT EXISTS tenant_encryption_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    encrypted_key BYTEA NOT NULL,
    key_version INTEGER NOT NULL DEFAULT 1,
    algorithm VARCHAR(50) NOT NULL DEFAULT 'AES-256-GCM',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_tenant_encryption_keys_tenant_id ON tenant_encryption_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_encryption_keys_organization_id ON tenant_encryption_keys(organization_id);
CREATE INDEX IF NOT EXISTS idx_tenant_encryption_keys_deleted_at ON tenant_encryption_keys(deleted_at);

COMMENT ON TABLE tenant_encryption_keys IS 'Stores tenant-specific encryption keys encrypted with master key. Platform cannot decrypt tenant data without tenant key. HIPAA compliant zero-knowledge encryption.';
