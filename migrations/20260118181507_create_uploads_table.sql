-- +goose Up
-- +goose StatementBegin
CREATE TABLE uploads
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_key   VARCHAR(1024) NOT NULL,
    entity_type  VARCHAR(255)  NOT NULL,
    entity_id    UUID          NOT NULL,

    file_size    BIGINT        NOT NULL,
    content_type VARCHAR(255),
    purpose      VARCHAR(255)  NOT NULL,

    created_at   TIMESTAMPTZ      DEFAULT NOW(),
    updated_at   TIMESTAMPTZ      DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_uploads_object_key_unique
    ON uploads (object_key)
    WHERE deleted_at IS NULL;


CREATE INDEX idx_uploads_entity
    ON uploads (entity_type, entity_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_uploads_deleted_at
    ON uploads (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS uploads;
-- +goose StatementEnd
