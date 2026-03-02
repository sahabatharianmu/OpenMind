ALTER TABLE users DROP COLUMN IF EXISTS system_role;
DROP INDEX IF EXISTS idx_organizations_subscription_plan_id;
ALTER TABLE organizations DROP COLUMN IF EXISTS subscription_plan_id;
DROP TABLE IF EXISTS subscription_plans;
