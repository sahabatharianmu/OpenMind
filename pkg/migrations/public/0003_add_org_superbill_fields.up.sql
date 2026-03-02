-- Public schema: Add superbill fields to organizations
-- Source: public-only portion of 000006_add_superbill_fields.up.sql

ALTER TABLE organizations ADD COLUMN IF NOT EXISTS tax_id VARCHAR(50);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS npi VARCHAR(50);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS address TEXT;
