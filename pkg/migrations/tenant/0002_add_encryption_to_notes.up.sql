-- Tenant schema: Add encryption columns to clinical_notes
-- Source: original 000004_add_encryption_to_notes.up.sql

ALTER TABLE clinical_notes ADD COLUMN IF NOT EXISTS content_encrypted BYTEA;
ALTER TABLE clinical_notes ADD COLUMN IF NOT EXISTS key_id VARCHAR(255);
ALTER TABLE clinical_notes ADD COLUMN IF NOT EXISTS nonce BYTEA;
