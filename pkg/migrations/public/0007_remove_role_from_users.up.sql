-- Public schema: Remove role column from users table
-- Role is now managed per-organization in organization_members table
-- Source: original 000013_remove_role_from_users.up.sql

ALTER TABLE users DROP COLUMN IF EXISTS role;
