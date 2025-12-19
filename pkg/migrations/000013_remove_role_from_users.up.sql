-- Remove role column from users table
-- Role is now managed per-organization in organization_members table
ALTER TABLE users DROP COLUMN IF EXISTS role;

