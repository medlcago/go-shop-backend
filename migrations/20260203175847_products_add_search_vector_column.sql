-- +goose Up
-- +goose StatementBegin
ALTER TABLE products
    ADD COLUMN search_vector tsvector
        GENERATED ALWAYS AS (
            setweight(to_tsvector('russian', COALESCE(name, '')), 'A') ||
            setweight(to_tsvector('russian', COALESCE(description, '')), 'B')
            ) STORED;

CREATE INDEX idx_products_search_vector ON products USING GIN (search_vector);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_products_search_vector;
ALTER TABLE products
    DROP COLUMN search_vector;
-- +goose StatementEnd
