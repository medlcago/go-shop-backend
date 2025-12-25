-- +goose Up
-- +goose StatementBegin
CREATE TABLE product_categories
(
    product_id  UUID,
    category_id UUID,
    PRIMARY KEY (product_id, category_id),
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_categories;
-- +goose StatementEnd
