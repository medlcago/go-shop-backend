-- +goose Up
-- +goose StatementBegin
CREATE TABLE products
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        VARCHAR(255)                           NOT NULL,
    description TEXT,
    price       NUMERIC(10, 2)                         NOT NULL,
    stock       INT                      DEFAULT 0 CHECK ( stock >= 0 ),
    slug        VARCHAR(255)                           NOT NULL,
    is_active   BOOLEAN                  DEFAULT TRUE  NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_products_code ON products (slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_code;
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
