-- Migration to create team_invitations table
-- This table stores team invitation requests with secure tokens

CREATE TABLE IF NOT EXISTS team_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    token VARCHAR(255) NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, accepted, expired, cancelled
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    accepted_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_team_invitations_organization_id ON team_invitations(organization_id);
CREATE INDEX IF NOT EXISTS idx_team_invitations_email ON team_invitations(email);
CREATE INDEX IF NOT EXISTS idx_team_invitations_token ON team_invitations(token);
CREATE INDEX IF NOT EXISTS idx_team_invitations_status ON team_invitations(status);
CREATE INDEX IF NOT EXISTS idx_team_invitations_deleted_at ON team_invitations(deleted_at);

-- Add constraint to ensure valid role
ALTER TABLE team_invitations ADD CONSTRAINT check_valid_role 
    CHECK (role IN ('owner', 'admin', 'clinician', 'case_manager', 'member'));

-- Add comment explaining the invitation flow
COMMENT ON TABLE team_invitations IS 'Stores team invitation requests. Invitations are sent via email with secure tokens that expire after a set period.';

