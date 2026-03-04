-- Drop audit_logs table and indexes
DROP INDEX IF EXISTS idx_audit_logs_deleted;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_user;
DROP INDEX IF EXISTS idx_audit_logs_org;
DROP TABLE IF EXISTS audit_logs;
