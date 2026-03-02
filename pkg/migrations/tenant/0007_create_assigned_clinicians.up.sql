-- Tenant schema: Create assigned_clinicians junction table
-- Source: original 000018_create_assigned_clinicians_table.up.sql (FK refs changed to tenant-local)

CREATE TABLE IF NOT EXISTS assigned_clinicians (
    patient_id UUID NOT NULL,
    clinician_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'primary',
    assigned_by UUID NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (patient_id, clinician_id),
    CONSTRAINT fk_assigned_clinicians_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assigned_clinicians_clinician_id
    ON assigned_clinicians(clinician_id);

CREATE INDEX IF NOT EXISTS idx_assigned_clinicians_patient_id
    ON assigned_clinicians(patient_id);

ALTER TABLE assigned_clinicians ADD CONSTRAINT check_assignment_role
    CHECK (role IN ('primary', 'secondary'));

COMMENT ON TABLE assigned_clinicians IS 'Junction table for patient-clinician assignments. All patients must be assigned to at least one clinician. Primary and secondary clinicians have full access.';
