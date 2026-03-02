DROP TRIGGER IF EXISTS trigger_prevent_signed_note_delete ON clinical_notes;
DROP TRIGGER IF EXISTS trigger_prevent_signed_note_update ON clinical_notes;
DROP FUNCTION IF EXISTS prevent_signed_note_delete();
DROP FUNCTION IF EXISTS prevent_signed_note_update();
