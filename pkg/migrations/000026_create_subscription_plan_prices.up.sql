CREATE TABLE IF NOT EXISTS subscription_plan_prices (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    plan_id UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    currency VARCHAR(10) NOT NULL,
    price BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(plan_id, currency)
);

CREATE INDEX IF NOT EXISTS idx_subscription_plan_prices_plan_id ON subscription_plan_prices(plan_id);

-- Migrate existing prices
INSERT INTO subscription_plan_prices (plan_id, currency, price)
SELECT id, currency, price FROM subscription_plans;

-- Drop old columns
ALTER TABLE subscription_plans DROP COLUMN IF EXISTS price;
ALTER TABLE subscription_plans DROP COLUMN IF EXISTS currency;
