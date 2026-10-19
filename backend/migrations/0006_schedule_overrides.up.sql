-- One override per instructor+date, either a full day off or a single
-- custom time range that completely replaces the recurring weekly pattern
-- for that specific date (e.g. "next Monday I only work mornings").
CREATE TABLE instructor_schedule_overrides (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instructor_id UUID NOT NULL REFERENCES instructor_profiles(user_id) ON DELETE CASCADE,
    date          DATE NOT NULL,
    is_day_off    BOOLEAN NOT NULL DEFAULT false,
    start_time    TIME,
    end_time      TIME,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (instructor_id, date),
    CHECK (is_day_off OR (start_time IS NOT NULL AND end_time IS NOT NULL AND end_time > start_time))
);
CREATE INDEX idx_schedule_overrides_instructor_date ON instructor_schedule_overrides(instructor_id, date);
