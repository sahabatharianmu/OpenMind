-- Remove subscription_tier column from organizations table
ALTER TABLE organizations 
DROP CONSTRAINT IF EXISTS check_subscription_tier;

ALTER TABLE organizations 
DROP COLUMN IF EXISTS subscription_tier;

