CREATE TABLE IF NOT EXISTS clinical_note_addendums (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    note_id UUID NOT NULL REFERENCES clinical_notes(id) ON DELETE CASCADE,
    clinician_id UUID NOT NULL REFERENCES users(id),
    content_encrypted BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    signed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

