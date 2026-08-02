-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_payment_methods
(
    id                         UUID PRIMARY KEY      DEFAULT gen_random_uuid(),

    user_id                    UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title                      VARCHAR(255) NOT NULL,
    provider                   VARCHAR(255) NOT NULL,
    provider_customer_id       VARCHAR(255),
    provider_payment_method_id VARCHAR(255) NOT NULL,

    type                       VARCHAR(100) NOT NULL,

    brand                      VARCHAR(100),
    last4                      VARCHAR(4),
    exp_month                  SMALLINT,
    exp_year                   SMALLINT,

    is_default                 BOOLEAN      NOT NULL DEFAULT FALSE,

    last_used_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at                 TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at                 TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_user_payment_methods_user_method_unique
    ON user_payment_methods (user_id, provider, provider_payment_method_id) WHERE deleted_at is NULL;

CREATE INDEX idx_user_payment_methods_user
    ON user_payment_methods (user_id);

CREATE INDEX idx_user_payment_methods_deleted_at
    ON user_payment_methods (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_payment_methods;
-- +goose StatementEnd
