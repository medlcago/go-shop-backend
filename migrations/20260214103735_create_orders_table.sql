-- +goose Up
-- +goose StatementBegin

CREATE TYPE order_status AS ENUM (
    'draft',
    'pending',
    'paid',
    'canceled',
    'completed'
    );

CREATE TABLE orders
(
    id           UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    user_id      UUID REFERENCES users (id) ON DELETE RESTRICT,
    session_id   UUID         NOT NULL,
    status       order_status NOT NULL DEFAULT 'draft',
    total_amount BIGINT       NOT NULL DEFAULT 0 CHECK ( total_amount >= 0 ),
    created_at   TIMESTAMPTZ           DEFAULT NOW() NOT NULL,
    updated_at   TIMESTAMPTZ           DEFAULT NOW() NOT NULL,
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_session_id ON orders (session_id);
CREATE INDEX idx_orders_status ON orders (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
-- +goose StatementEnd