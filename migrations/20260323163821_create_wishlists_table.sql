-- +goose Up
-- +goose StatementBegin
CREATE TABLE wishlists
(
    id          UUID PRIMARY KEY                    DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users (id) NOT NULL,
    title       VARCHAR(255)               NOT NULL,
    is_public   BOOLEAN                    NOT NULL DEFAULT FALSE,
    share_token VARCHAR(64)                NOT NULL,
    created_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_wishlists_share_token_unique ON wishlists (share_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wishlists;
-- +goose StatementEnd
