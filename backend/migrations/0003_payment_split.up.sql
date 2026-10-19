-- Rename to provider-agnostic column names (payments.mp_* predates the
-- PaymentProvider interface, which is meant to work with any gateway).
ALTER TABLE payments RENAME COLUMN mp_preference_id TO provider_ref;
ALTER TABLE payments RENAME COLUMN mp_payment_id TO provider_payment_id;

-- Marketplace split: the instructor is always paid exactly what they quoted;
-- the platform fee is added on top of that to get amount_uyu (what the
-- student is actually charged).
ALTER TABLE payments ADD COLUMN instructor_price_uyu NUMERIC(10,2) NOT NULL DEFAULT 0;
ALTER TABLE payments ADD COLUMN platform_fee_uyu NUMERIC(10,2) NOT NULL DEFAULT 0;
