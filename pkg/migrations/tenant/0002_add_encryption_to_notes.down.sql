ALTER TABLE clinical_notes DROP COLUMN IF EXISTS nonce;
ALTER TABLE clinical_notes DROP COLUMN IF EXISTS key_id;
ALTER TABLE clinical_notes DROP COLUMN IF EXISTS content_encrypted;
