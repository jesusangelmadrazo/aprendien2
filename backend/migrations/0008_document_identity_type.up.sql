-- "cédula" is the person's national ID (documento de identidad), distinct
-- from vehicle_registration (propiedad del auto) — they were wrongly
-- conflated in the first pass.
ALTER TYPE document_type ADD VALUE 'identity';
