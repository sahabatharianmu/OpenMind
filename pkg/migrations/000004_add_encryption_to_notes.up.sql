ALTER TABLE clinical_notes ADD COLUMN content_encrypted BYTEA;
ALTER TABLE clinical_notes ADD COLUMN key_id VARCHAR(255);
ALTER TABLE clinical_notes ADD COLUMN nonce BYTEA;

-- We keep the original columns for now but they will be empty if we use the encrypted blob
-- Or we could drop them if we are sure. The plan doesn't say to drop them.

