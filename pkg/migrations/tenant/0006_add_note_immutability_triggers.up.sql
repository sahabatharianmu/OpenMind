-- Tenant schema: Note immutability triggers for signed notes
-- Source: original 000009_add_note_immutability_triggers.up.sql

CREATE OR REPLACE FUNCTION prevent_signed_note_update()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.is_signed = TRUE THEN
        RAISE EXCEPTION 'Cannot update a signed clinical note. Signed notes are immutable for compliance.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION prevent_signed_note_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.is_signed = TRUE THEN
        RAISE EXCEPTION 'Cannot delete a signed clinical note. Signed notes are immutable for compliance.';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_prevent_signed_note_update ON clinical_notes;
CREATE TRIGGER trigger_prevent_signed_note_update
    BEFORE UPDATE ON clinical_notes
    FOR EACH ROW
    EXECUTE FUNCTION prevent_signed_note_update();

DROP TRIGGER IF EXISTS trigger_prevent_signed_note_delete ON clinical_notes;
CREATE TRIGGER trigger_prevent_signed_note_delete
    BEFORE DELETE ON clinical_notes
    FOR EACH ROW
    EXECUTE FUNCTION prevent_signed_note_delete();
