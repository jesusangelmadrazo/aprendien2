CREATE TABLE instructor_availability (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instructor_id UUID NOT NULL REFERENCES instructor_profiles(user_id) ON DELETE CASCADE,
    day_of_week   SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0 = Sunday .. 6 = Saturday, matches time.Weekday
    start_time    TIME NOT NULL,
    end_time      TIME NOT NULL CHECK (end_time > start_time),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_instructor_availability_instructor ON instructor_availability(instructor_id);

ALTER TABLE lesson_requests ADD COLUMN scheduled_at TIMESTAMPTZ;

-- Makes double-booking a specific slot impossible even under concurrent
-- requests: the DB itself rejects the second insert rather than relying on
-- application-level locking.
CREATE UNIQUE INDEX idx_lesson_requests_instructor_slot
    ON lesson_requests(instructor_id, scheduled_at)
    WHERE scheduled_at IS NOT NULL AND status != 'cancelled';
