-- +goose Up
-- +goose StatementBegin
CREATE TABLE products
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255)                   NOT NULL,
    description TEXT,
    price       BIGINT                         NOT NULL CHECK ( price >= 0 ),
    stock       INT              DEFAULT 0 CHECK ( stock >= 0 ),
    reserved    INT              DEFAULT 0 CHECK ( reserved >= 0 ),
    slug        VARCHAR(255)                   NOT NULL,
    is_active   BOOLEAN          DEFAULT TRUE  NOT NULL,
    created_at  TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    updated_at  TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_products_deleted_at ON products (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
