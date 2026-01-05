-- Add back columns
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS price BIGINT NOT NULL DEFAULT 0;
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS currency VARCHAR(10) NOT NULL DEFAULT 'USD';

-- Restore data (Best effort: try to find USD price, otherwise take any price)
UPDATE subscription_plans sp
SET price = spp.price, currency = spp.currency
FROM subscription_plan_prices spp
WHERE sp.id = spp.plan_id AND (spp.currency = 'USD' OR spp.currency = (SELECT currency FROM subscription_plan_prices WHERE plan_id = sp.id LIMIT 1));

DROP TABLE IF EXISTS subscription_plan_prices;
