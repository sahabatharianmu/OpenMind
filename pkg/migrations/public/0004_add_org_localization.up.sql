-- Public schema: Add localization columns to organizations
-- Source: original 000008_add_org_localization.up.sql

ALTER TABLE organizations ADD COLUMN IF NOT EXISTS currency VARCHAR(10) NOT NULL DEFAULT 'USD';
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS locale VARCHAR(10) NOT NULL DEFAULT 'en-US';
