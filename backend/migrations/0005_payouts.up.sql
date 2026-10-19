CREATE TYPE account_type AS ENUM ('caja_ahorro', 'cuenta_corriente');

CREATE TABLE instructor_payout_methods (
    instructor_id           UUID PRIMARY KEY REFERENCES instructor_profiles(user_id) ON DELETE CASCADE,
    bank_name               TEXT NOT NULL,
    account_type            account_type NOT NULL,
    account_number          TEXT NOT NULL,
    account_holder_name     TEXT NOT NULL,
    account_holder_document TEXT NOT NULL,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payouts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instructor_id UUID NOT NULL REFERENCES users(id),
    amount_uyu    NUMERIC(10,2) NOT NULL,
    period_start  TIMESTAMPTZ NOT NULL,
    period_end    TIMESTAMPTZ NOT NULL,
    created_by    UUID NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payouts_instructor ON payouts(instructor_id);

-- Denormalized onto payments so wallet/payout queries don't need to join
-- through lesson_requests every time. payout_id is set once a weekly payout
-- (marked paid manually by an admin, real money moves outside the app) is
-- created for that payment.
ALTER TABLE payments ADD COLUMN instructor_id UUID REFERENCES users(id);
ALTER TABLE payments ADD COLUMN payout_id UUID REFERENCES payouts(id);
CREATE INDEX idx_payments_instructor_pending ON payments(instructor_id) WHERE status = 'approved' AND payout_id IS NULL;
