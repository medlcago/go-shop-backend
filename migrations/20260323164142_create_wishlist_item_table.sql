-- +goose Up
-- +goose StatementBegin
CREATE TABLE wishlist_items
(
    id          UUID PRIMARY KEY                        DEFAULT gen_random_uuid(),
    wishlist_id UUID REFERENCES wishlists (id) NOT NULL,
    product_id  UUID REFERENCES products (id)  NOT NULL,
    note        VARCHAR(128),
    priority    INTEGER                                 DEFAULT 0 NOT NULL,
    created_at  TIMESTAMPTZ                    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ                    NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_wishlist_items_product_unique ON wishlist_items (wishlist_id, product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wishlist_items;
-- +goose StatementEnd
