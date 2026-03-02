-- Tenant schema: Create clinical_note_addendums table
-- Source: original 000005_create_note_addendums.up.sql

CREATE TABLE IF NOT EXISTS clinical_note_addendums (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_id UUID NOT NULL REFERENCES clinical_notes(id) ON DELETE CASCADE,
    clinician_id UUID NOT NULL,
    content_encrypted BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    signed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
