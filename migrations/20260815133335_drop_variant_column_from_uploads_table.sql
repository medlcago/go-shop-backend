-- +goose Up
-- +goose StatementBegin
ALTER TABLE uploads
    DROP COLUMN IF EXISTS variant;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE uploads
    ADD COLUMN variant VARCHAR(50) NOT NULL DEFAULT 'original';
-- +goose StatementEnd
