-- Add subscription_tier column to organizations table
-- Default tier is 'free' for all organizations
ALTER TABLE organizations 
ADD COLUMN IF NOT EXISTS subscription_tier VARCHAR(50) NOT NULL DEFAULT 'free';

-- Add check constraint to ensure valid tier values
ALTER TABLE organizations 
ADD CONSTRAINT check_subscription_tier 
CHECK (subscription_tier IN ('free', 'paid'));

-- Update existing organizations to 'free' tier if they don't have one
UPDATE organizations 
SET subscription_tier = 'free' 
WHERE subscription_tier IS NULL OR subscription_tier = '';

