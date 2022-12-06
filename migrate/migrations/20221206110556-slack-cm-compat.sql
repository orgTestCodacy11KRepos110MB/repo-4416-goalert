-- +migrate Up notransaction
ALTER TYPE enum_user_contact_method_type
ADD VALUE IF NOT EXISTS 'SLACK_DM';

ALTER TYPE engine_processing_type
ADD VALUE IF NOT EXISTS 'compat';

INSERT INTO engine_processing_versions (type_id, version)
VALUES ('compat', 1) ON CONFLICT DO NOTHING;

-- +migrate Down
DELETE FROM engine_processing_versions
WHERE type_id = 'compat';
