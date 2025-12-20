-- Down migration: Drop team_invitations table

DROP INDEX IF EXISTS idx_team_invitations_deleted_at;
DROP INDEX IF EXISTS idx_team_invitations_status;
DROP INDEX IF EXISTS idx_team_invitations_token;
DROP INDEX IF EXISTS idx_team_invitations_email;
DROP INDEX IF EXISTS idx_team_invitations_organization_id;

DROP TABLE IF EXISTS team_invitations;

