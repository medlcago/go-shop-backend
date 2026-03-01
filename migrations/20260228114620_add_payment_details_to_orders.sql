-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
    ADD COLUMN payment_id VARCHAR(255);

ALTER TABLE orders
    ADD COLUMN provider_name VARCHAR(255);

ALTER TABLE orders
    ADD COLUMN paid_at TIMESTAMPTZ;

ALTER TABLE orders
    ADD COLUMN canceled_at TIMESTAMPTZ;

ALTER TABLE orders
    ADD COLUMN expires_at TIMESTAMPTZ;

CREATE UNIQUE INDEX idx_orders_provider_payment_unique ON orders (provider_name, payment_id);
CREATE INDEX idx_orders_status_expires_at ON orders (status, expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_orders_provider_payment_unique;

ALTER TABLE orders
    DROP COLUMN payment_id;

ALTER TABLE orders
    DROP COLUMN provider_name;

ALTER TABLE orders
    DROP COLUMN paid_at;

ALTER TABLE orders
    DROP COLUMN canceled_at;

ALTER TABLE orders
    DROP COLUMN expires_at;
-- +goose StatementEnd
