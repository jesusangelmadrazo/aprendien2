CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('student', 'instructor', 'admin');

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    phone           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    full_name       TEXT NOT NULL,
    role            user_role NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE instructor_status AS ENUM ('offline', 'online', 'busy');

CREATE TABLE instructor_profiles (
    user_id             UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    vehicle_make        TEXT NOT NULL,
    vehicle_model       TEXT NOT NULL,
    vehicle_plate       TEXT NOT NULL,
    vehicle_year        INT,
    hourly_rate_uyu     NUMERIC(10,2) NOT NULL,
    bio                 TEXT,
    status              instructor_status NOT NULL DEFAULT 'offline',
    avg_rating          NUMERIC(3,2) NOT NULL DEFAULT 0,
    total_lessons       INT NOT NULL DEFAULT 0,
    last_known_location GEOGRAPHY(Point, 4326),
    last_seen_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE lesson_status AS ENUM (
    'requested',
    'searching',
    'accepted',
    'instructor_enroute',
    'in_progress',
    'completed',
    'cancelled',
    'no_instructor_found'
);

CREATE TABLE lesson_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id          UUID NOT NULL REFERENCES users(id),
    instructor_id       UUID REFERENCES users(id),
    status              lesson_status NOT NULL DEFAULT 'requested',
    pickup_location     GEOGRAPHY(Point, 4326) NOT NULL,
    pickup_address_text TEXT,
    duration_minutes    INT NOT NULL DEFAULT 60,
    estimated_price_uyu NUMERIC(10,2) NOT NULL,
    final_price_uyu     NUMERIC(10,2),
    requested_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at         TIMESTAMPTZ,
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    cancel_reason       TEXT
);
CREATE INDEX idx_lesson_requests_student ON lesson_requests(student_id, status);
CREATE INDEX idx_lesson_requests_instructor ON lesson_requests(instructor_id, status);

CREATE TYPE offer_result AS ENUM ('pending', 'accepted', 'rejected', 'timed_out');

CREATE TABLE lesson_offers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_request_id   UUID NOT NULL REFERENCES lesson_requests(id) ON DELETE CASCADE,
    instructor_id       UUID NOT NULL REFERENCES users(id),
    sequence_order      INT NOT NULL,
    result              offer_result NOT NULL DEFAULT 'pending',
    offered_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    responded_at        TIMESTAMPTZ
);
CREATE INDEX idx_lesson_offers_request ON lesson_offers(lesson_request_id);

CREATE TYPE payment_status AS ENUM ('pending', 'processing', 'approved', 'rejected', 'refunded');

CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_request_id   UUID NOT NULL REFERENCES lesson_requests(id),
    student_id          UUID NOT NULL REFERENCES users(id),
    amount_uyu          NUMERIC(10,2) NOT NULL,
    mp_preference_id    TEXT,
    mp_payment_id       TEXT,
    status              payment_status NOT NULL DEFAULT 'pending',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payments_lesson ON payments(lesson_request_id);

CREATE TABLE reviews (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_request_id   UUID NOT NULL REFERENCES lesson_requests(id),
    reviewer_id         UUID NOT NULL REFERENCES users(id),
    reviewee_id         UUID NOT NULL REFERENCES users(id),
    rating              SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment             TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(lesson_request_id, reviewer_id)
);
