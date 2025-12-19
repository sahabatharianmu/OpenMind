-- Migration: Rollback audit log protection triggers
-- Description: Removes PostgreSQL triggers that prevent audit log modifications

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_prevent_audit_log_delete ON audit_logs;
DROP TRIGGER IF EXISTS trigger_prevent_audit_log_update ON audit_logs;

-- Drop functions
DROP FUNCTION IF EXISTS prevent_audit_log_delete();
DROP FUNCTION IF EXISTS prevent_audit_log_update();

