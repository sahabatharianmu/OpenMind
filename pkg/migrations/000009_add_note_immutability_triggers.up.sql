-- Create function to prevent signed note updates
CREATE OR REPLACE FUNCTION prevent_signed_note_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if the note was already signed before this update
    -- Allow signing (is_signed: false -> true) but prevent any other updates to signed notes
    IF OLD.is_signed = TRUE THEN
        RAISE EXCEPTION 'Cannot update a signed clinical note. Signed notes are immutable for compliance.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create function to prevent signed note deletion
CREATE OR REPLACE FUNCTION prevent_signed_note_delete()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if the note is signed
    IF OLD.is_signed = TRUE THEN
        RAISE EXCEPTION 'Cannot delete a signed clinical note. Signed notes are immutable for compliance.';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to prevent signed note updates
DROP TRIGGER IF EXISTS trigger_prevent_signed_note_update ON clinical_notes;
CREATE TRIGGER trigger_prevent_signed_note_update
    BEFORE UPDATE ON clinical_notes
    FOR EACH ROW
    EXECUTE FUNCTION prevent_signed_note_update();

-- Create trigger to prevent signed note deletion
DROP TRIGGER IF EXISTS trigger_prevent_signed_note_delete ON clinical_notes;
CREATE TRIGGER trigger_prevent_signed_note_delete
    BEFORE DELETE ON clinical_notes
    FOR EACH ROW
    EXECUTE FUNCTION prevent_signed_note_delete();

