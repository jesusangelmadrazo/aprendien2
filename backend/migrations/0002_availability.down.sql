DROP INDEX IF EXISTS idx_lesson_requests_instructor_slot;
ALTER TABLE lesson_requests DROP COLUMN IF EXISTS scheduled_at;
DROP TABLE IF EXISTS instructor_availability;
