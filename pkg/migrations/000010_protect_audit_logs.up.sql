-- Migration: Protect audit logs from deletion and updates
-- Description: Creates PostgreSQL triggers to prevent audit log modifications
--              Audit logs must remain immutable for HIPAA compliance

-- Create function to prevent audit log deletion
CREATE OR REPLACE FUNCTION prevent_audit_log_delete()
RETURNS TRIGGER AS $$
BEGIN
    -- Audit logs are immutable and cannot be deleted
    RAISE EXCEPTION 'Cannot delete audit logs. Audit logs are immutable for compliance and must be retained.';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to prevent audit log deletion
DROP TRIGGER IF EXISTS trigger_prevent_audit_log_delete ON audit_logs;
CREATE TRIGGER trigger_prevent_audit_log_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_delete();

-- Create function to prevent audit log updates
CREATE OR REPLACE FUNCTION prevent_audit_log_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Audit logs are immutable and cannot be updated
    RAISE EXCEPTION 'Cannot update audit logs. Audit logs are immutable for compliance and must be retained as-is.';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to prevent audit log updates
DROP TRIGGER IF EXISTS trigger_prevent_audit_log_update ON audit_logs;
CREATE TRIGGER trigger_prevent_audit_log_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_update();

