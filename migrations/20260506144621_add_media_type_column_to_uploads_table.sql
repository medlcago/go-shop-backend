-- +goose Up
-- +goose StatementBegin
ALTER TABLE uploads
    ADD COLUMN media_type VARCHAR(100) NOT NULL DEFAULT 'default';

ALTER TABLE uploads
    ADD COLUMN variant VARCHAR(50) NOT NULL DEFAULT 'original';

DROP INDEX IF EXISTS idx_uploads_is_main_unique;

ALTER TABLE uploads
    DROP COLUMN is_main;

CREATE INDEX idx_uploads_entity_media_type
    ON uploads (entity_type, entity_id, media_type)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE uploads
    ADD COLUMN is_main BOOLEAN NOT NULL DEFAULT FALSE;

DROP INDEX IF EXISTS idx_uploads_entity_media_type;

ALTER TABLE uploads
    DROP COLUMN media_type;

CREATE UNIQUE INDEX idx_uploads_is_main_unique
    ON uploads (entity_type, entity_id)
    WHERE is_main = true AND deleted_at IS NULL;
-- +goose StatementEnd
