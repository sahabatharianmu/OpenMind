-- Migration to create assigned_clinicians junction table
-- This table stores patient-clinician assignments with primary/secondary roles

CREATE TABLE IF NOT EXISTS assigned_clinicians (
    patient_id UUID NOT NULL,
    clinician_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'primary', -- 'primary' or 'secondary'
    assigned_by UUID NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (patient_id, clinician_id),
    CONSTRAINT fk_assigned_clinicians_patient 
        FOREIGN KEY (patient_id) 
        REFERENCES patients(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_assigned_clinicians_clinician 
        FOREIGN KEY (clinician_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_assigned_clinicians_assigned_by 
        FOREIGN KEY (assigned_by) 
        REFERENCES users(id) 
        ON DELETE SET NULL
);

-- Index for fast lookups by clinician
CREATE INDEX IF NOT EXISTS idx_assigned_clinicians_clinician_id 
    ON assigned_clinicians(clinician_id);

-- Index for fast lookups by patient
CREATE INDEX IF NOT EXISTS idx_assigned_clinicians_patient_id 
    ON assigned_clinicians(patient_id);

-- Add constraint to ensure valid role
ALTER TABLE assigned_clinicians ADD CONSTRAINT check_assignment_role 
    CHECK (role IN ('primary', 'secondary'));

-- Add comment explaining the assignment system
COMMENT ON TABLE assigned_clinicians IS 'Junction table for patient-clinician assignments. All patients must be assigned to at least one clinician. Primary and secondary clinicians have full access.';

