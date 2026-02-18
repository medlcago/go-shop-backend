-- +goose Up
-- +goose StatementBegin
CREATE TABLE order_items
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID                           NOT NULL REFERENCES orders (id) ON DELETE RESTRICT,
    product_id   UUID                           NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    product_name VARCHAR(255)                   NOT NULL,
    quantity     INTEGER                        NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price   BIGINT                         NOT NULL CHECK (unit_price >= 0),
    created_at   TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    updated_at   TIMESTAMPTZ      DEFAULT NOW() NOT NULL
);

CREATE UNIQUE INDEX uniq_order_items_order_product
    ON order_items (order_id, product_id);

CREATE INDEX idx_order_items_order_id
    ON order_items (order_id);

CREATE INDEX idx_order_items_product_id
    ON order_items (product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_items;
-- +goose StatementEnd
