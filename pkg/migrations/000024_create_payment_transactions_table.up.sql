-- Migration to create payment_transactions table in tenant schemas
-- This table stores payment transactions (QRIS, credit card, virtual account, etc.)

CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    provider_transaction_id VARCHAR(255),
    partner_reference_no VARCHAR(255) UNIQUE,
    external_id VARCHAR(255),
    qr_code TEXT,
    qr_code_url TEXT,
    qr_code_image TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_payment_transactions_organization 
        FOREIGN KEY (organization_id) 
        REFERENCES public.organizations(id) 
        ON DELETE CASCADE,
    CONSTRAINT check_payment_type CHECK (type IN ('subscription', 'one_time')),
    CONSTRAINT check_payment_method_type CHECK (payment_method IN ('qris', 'credit_card', 'virtual_account')),
    CONSTRAINT check_payment_provider_type CHECK (provider IN ('midtrans', 'stripe', 'square')),
    CONSTRAINT check_payment_status CHECK (status IN ('pending', 'paid', 'failed', 'cancelled', 'expired'))
);

CREATE INDEX IF NOT EXISTS idx_payment_transactions_organization_id ON payment_transactions(organization_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_provider_transaction_id ON payment_transactions(provider_transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_partner_reference_no ON payment_transactions(partner_reference_no);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_external_id ON payment_transactions(external_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_expires_at ON payment_transactions(expires_at);

COMMENT ON TABLE payment_transactions IS 'Stores payment transactions for subscriptions and one-time payments. Supports multiple payment methods (QRIS, credit card, virtual account) and providers (Midtrans, Stripe, Square).';

