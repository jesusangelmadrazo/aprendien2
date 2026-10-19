-- Postgres can't drop a single enum value directly; recreate the type
-- without it. Any 'identity' rows are removed first (dev-only migration,
-- acceptable data loss on rollback).
DELETE FROM instructor_documents WHERE doc_type = 'identity';

ALTER TYPE document_type RENAME TO document_type_old;
CREATE TYPE document_type AS ENUM ('license', 'vehicle_registration', 'insurance');
ALTER TABLE instructor_documents ALTER COLUMN doc_type TYPE document_type USING doc_type::text::document_type;
DROP TYPE document_type_old;
