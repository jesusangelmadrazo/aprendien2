ALTER TABLE payments DROP COLUMN IF EXISTS platform_fee_uyu;
ALTER TABLE payments DROP COLUMN IF EXISTS instructor_price_uyu;
ALTER TABLE payments RENAME COLUMN provider_payment_id TO mp_payment_id;
ALTER TABLE payments RENAME COLUMN provider_ref TO mp_preference_id;
