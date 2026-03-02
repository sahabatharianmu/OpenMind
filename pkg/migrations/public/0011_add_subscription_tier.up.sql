-- Public schema: Add subscription_tier column to organizations
-- Source: original 000022_add_subscription_tier.up.sql

ALTER TABLE organizations
ADD COLUMN IF NOT EXISTS subscription_tier VARCHAR(50) NOT NULL DEFAULT 'free';

ALTER TABLE organizations
ADD CONSTRAINT check_subscription_tier
CHECK (subscription_tier IN ('free', 'paid'));

UPDATE organizations
SET subscription_tier = 'free'
WHERE subscription_tier IS NULL OR subscription_tier = '';
