-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         CITEXT                         NOT NULL,
    password_hash VARCHAR(255)                   NOT NULL,
    full_name     VARCHAR(255),
    phone         VARCHAR(50),
    created_at    TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    updated_at    TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_email_unique ON users (email);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
