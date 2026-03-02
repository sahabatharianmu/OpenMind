-- Tenant schema: Add superbill fields to tenant tables
-- Source: tenant-only portion of 000006_add_superbill_fields.up.sql

ALTER TABLE appointments ADD COLUMN IF NOT EXISTS cpt_code VARCHAR(20);
ALTER TABLE clinical_notes ADD COLUMN IF NOT EXISTS icd10_code VARCHAR(20);
