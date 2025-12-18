ALTER TABLE organizations ADD COLUMN tax_id VARCHAR(50);
ALTER TABLE organizations ADD COLUMN npi VARCHAR(50);
ALTER TABLE organizations ADD COLUMN address TEXT;

ALTER TABLE appointments ADD COLUMN cpt_code VARCHAR(20);

ALTER TABLE clinical_notes ADD COLUMN icd10_code VARCHAR(20);

