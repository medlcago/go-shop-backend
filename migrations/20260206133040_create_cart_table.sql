-- +goose Up
-- +goose StatementBegin
CREATE TABLE carts
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users (id) ON DELETE CASCADE,
    session_id UUID,
    created_at TIMESTAMPTZ      DEFAULT NOW(),
    updated_at TIMESTAMPTZ      DEFAULT NOW(),

    CONSTRAINT chk_cart_owner CHECK (
        (user_id IS NOT NULL AND session_id IS NULL) OR
        (user_id IS NULL AND session_id IS NOT NULL)
        )
);

CREATE UNIQUE INDEX idx_carts_user_id_unique ON carts (user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_carts_session_id_unique ON carts (session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_carts_user_id ON carts (user_id);
CREATE INDEX idx_carts_session_id ON carts (session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS carts;
-- +goose StatementEnd
