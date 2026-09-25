-- +goose Up
-- +goose StatementBegin
CREATE TABLE favorites
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id UUID        NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_favorites_user_product_unique ON favorites (user_id, product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS favorites;
-- +goose StatementEnd
