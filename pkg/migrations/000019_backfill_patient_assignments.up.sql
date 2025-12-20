-- Migration to backfill existing patients with assignments
-- Assigns each patient's creator as the primary clinician

INSERT INTO assigned_clinicians (patient_id, clinician_id, role, assigned_by, assigned_at)
SELECT 
    id AS patient_id,
    created_by AS clinician_id,
    'primary' AS role,
    created_by AS assigned_by,
    created_at AS assigned_at
FROM patients
WHERE NOT EXISTS (
    SELECT 1 
    FROM assigned_clinicians 
    WHERE assigned_clinicians.patient_id = patients.id
)
ON CONFLICT (patient_id, clinician_id) DO NOTHING;

