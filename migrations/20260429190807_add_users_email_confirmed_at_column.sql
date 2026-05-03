-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN email_confirmed_at TIMESTAMPTZ;

CREATE INDEX idx_users_email_confirmed_at ON users (email_confirmed_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_users_email_confirmed_at;

ALTER TABLE users
    DROP COLUMN email_confirmed_at;
-- +goose StatementEnd
