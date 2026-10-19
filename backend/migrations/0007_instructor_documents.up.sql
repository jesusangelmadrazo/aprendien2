CREATE TYPE document_type AS ENUM ('license', 'vehicle_registration', 'insurance');
CREATE TYPE document_status AS ENUM ('pending_review', 'approved', 'rejected');
CREATE TYPE instructor_verification_status AS ENUM ('pending', 'approved');

CREATE TABLE instructor_documents (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instructor_id     UUID NOT NULL REFERENCES instructor_profiles(user_id) ON DELETE CASCADE,
    doc_type          document_type NOT NULL,
    photo_url         TEXT NOT NULL,
    expiry_date       DATE,
    status            document_status NOT NULL DEFAULT 'pending_review',
    rejection_reason  TEXT,
    reviewed_by       UUID REFERENCES users(id),
    submitted_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_instructor_documents_instructor ON instructor_documents(instructor_id, doc_type);
CREATE INDEX idx_instructor_documents_pending ON instructor_documents(status) WHERE status = 'pending_review';

-- Cached/denormalized so listing and gating queries (ListAll, "can go online")
-- don't need a correlated subquery over instructor_documents every time —
-- recomputed by DocumentService whenever an admin reviews a document.
ALTER TABLE instructor_profiles ADD COLUMN verification_status instructor_verification_status NOT NULL DEFAULT 'pending';
