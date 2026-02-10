-- +goose Up
-- +goose StatementBegin
CREATE TABLE cart_items
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id    UUID                           NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    product_id UUID                           NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    quantity   INTEGER                        NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price DECIMAL(10, 2)                 NOT NULL CHECK (unit_price >= 0),
    created_at TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ      DEFAULT NOW() NOT NULL
);

CREATE UNIQUE INDEX idx_cart_items_cart_product_unique ON cart_items (cart_id, product_id);
CREATE INDEX idx_cart_items_cart_id ON cart_items (cart_id);
CREATE INDEX idx_cart_items_product_id ON cart_items (product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cart_items;
-- +goose StatementEnd
