CREATE TABLE IF NOT EXISTS clinical_note_attachments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    note_id UUID NOT NULL REFERENCES clinical_notes(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    data_encrypted BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

