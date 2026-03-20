-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN two_fa_enabled BOOLEAN DEFAULT FALSE NOT NULL;

ALTER TABLE users
    ADD COLUMN two_fa_secret VARCHAR(255);

ALTER TABLE users
    ADD COLUMN two_fa_confirmed_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN two_fa_enabled;

ALTER TABLE users
    DROP COLUMN two_fa_secret;

ALTER TABLE users
    DROP COLUMN two_fa_confirmed_at;
-- +goose StatementEnd
