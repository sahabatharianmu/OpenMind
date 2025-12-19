-- Drop triggers
DROP TRIGGER IF EXISTS trigger_prevent_audit_log_update ON audit_logs;
DROP TRIGGER IF EXISTS trigger_prevent_audit_log_delete ON audit_logs;

-- Drop functions
DROP FUNCTION IF EXISTS prevent_audit_log_update();
DROP FUNCTION IF EXISTS prevent_audit_log_delete();

