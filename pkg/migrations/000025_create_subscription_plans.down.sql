ALTER TABLE users DROP COLUMN IF NOT EXISTS system_role;

ALTER TABLE organizations DROP COLUMN IF NOT EXISTS subscription_plan_id;

DROP TABLE IF EXISTS subscription_plans;
