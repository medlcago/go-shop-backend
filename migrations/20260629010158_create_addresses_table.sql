-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS addresses
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID                           NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       VARCHAR(100)                   NOT NULL,
    street     VARCHAR(255)                   NOT NULL,
    house      VARCHAR(50)                    NOT NULL,
    city       VARCHAR(255)                   NOT NULL,
    country    VARCHAR(100)                   NOT NULL,

    floor      VARCHAR(50),
    entrance   VARCHAR(50),
    apartment  VARCHAR(50),
    comment    TEXT,
    is_default BOOLEAN                        NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_addresses_is_default ON addresses (is_default);
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses (user_id);
CREATE INDEX IF NOT EXISTS idx_addresses_deleted_at ON addresses (deleted_at);

ALTER TABLE orders
    ADD COLUMN address JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS addresses;
ALTER TABLE orders
    DROP COLUMN IF EXISTS address;
-- +goose StatementEnd
