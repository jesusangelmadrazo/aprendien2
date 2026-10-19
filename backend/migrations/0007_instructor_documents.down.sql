ALTER TABLE instructor_profiles DROP COLUMN IF EXISTS verification_status;
DROP TABLE IF EXISTS instructor_documents;
DROP TYPE IF EXISTS instructor_verification_status;
DROP TYPE IF EXISTS document_status;
DROP TYPE IF EXISTS document_type;
