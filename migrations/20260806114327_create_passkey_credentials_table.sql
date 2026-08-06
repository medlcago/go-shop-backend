-- +goose Up
-- +goose StatementBegin
CREATE TABLE passkey_credentials
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),

    user_id       UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    credential_id BYTEA       NOT NULL,
    credential    JSONB       NOT NULL,

    name          VARCHAR(255),

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    last_used_at  TIMESTAMPTZ
);

CREATE INDEX idx_passkey_credentials_user_id
    ON passkey_credentials(user_id);

CREATE UNIQUE INDEX idx_passkey_credentials_credential_id_unique
    ON passkey_credentials (credential_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_passkey_credentials_deleted_at
    ON passkey_credentials (deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS passkey_credentials;
-- +goose StatementEnd
