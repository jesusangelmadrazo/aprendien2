DROP INDEX IF EXISTS idx_payments_instructor_pending;
ALTER TABLE payments DROP COLUMN IF EXISTS payout_id;
ALTER TABLE payments DROP COLUMN IF EXISTS instructor_id;
DROP TABLE IF EXISTS payouts;
DROP TABLE IF EXISTS instructor_payout_methods;
DROP TYPE IF EXISTS account_type;
