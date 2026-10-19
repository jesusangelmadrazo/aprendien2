CREATE TYPE exam_type AS ENUM ('theory', 'practice');
CREATE TYPE exam_status AS ENUM ('pending_review', 'approved', 'rejected');

CREATE TABLE exams (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type              exam_type NOT NULL,
    status            exam_status NOT NULL DEFAULT 'pending_review',
    photo_url         TEXT NOT NULL,
    rejection_reason  TEXT,
    reviewed_by       UUID REFERENCES users(id),
    submitted_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_exams_student ON exams(student_id, type);
CREATE INDEX idx_exams_pending ON exams(status) WHERE status = 'pending_review';
