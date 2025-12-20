-- Migration to create patient_handoffs table
-- This table stores patient handoff requests between clinicians

CREATE TABLE IF NOT EXISTS patient_handoffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL,
    requesting_clinician_id UUID NOT NULL,
    receiving_clinician_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'requested',
    requested_role VARCHAR(50), -- Role the receiving clinician should get (inherits from requesting if null)
    message TEXT,
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP WITH TIME ZONE,
    responded_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_patient_handoffs_patient 
        FOREIGN KEY (patient_id) 
        REFERENCES patients(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_patient_handoffs_requesting 
        FOREIGN KEY (requesting_clinician_id) 
        REFERENCES public.users(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_patient_handoffs_receiving 
        FOREIGN KEY (receiving_clinician_id) 
        REFERENCES public.users(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_patient_handoffs_responded_by 
        FOREIGN KEY (responded_by) 
        REFERENCES public.users(id) 
        ON DELETE SET NULL,
    CONSTRAINT check_handoff_status 
        CHECK (status IN ('requested', 'approved', 'rejected', 'cancelled')),
    CONSTRAINT check_handoff_not_self 
        CHECK (requesting_clinician_id != receiving_clinician_id)
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_patient_handoffs_patient_id ON patient_handoffs(patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_handoffs_requesting_clinician_id ON patient_handoffs(requesting_clinician_id);
CREATE INDEX IF NOT EXISTS idx_patient_handoffs_receiving_clinician_id ON patient_handoffs(receiving_clinician_id);
CREATE INDEX IF NOT EXISTS idx_patient_handoffs_status ON patient_handoffs(status);
CREATE INDEX IF NOT EXISTS idx_patient_handoffs_requested_at ON patient_handoffs(requested_at DESC);

-- Add comment explaining the handoff system
COMMENT ON TABLE patient_handoffs IS 'Stores patient handoff requests between clinicians. Workflow: requested → approved/rejected/cancelled. On approval, patient assignments are updated automatically.';

