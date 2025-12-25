-- +goose Up
-- +goose StatementBegin
CREATE TABLE products
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255)                   NOT NULL,
    description TEXT,
    price       NUMERIC(10, 2)                 NOT NULL,
    stock       INT              DEFAULT 0 CHECK ( stock >= 0 ),
    slug        VARCHAR(255)                   NOT NULL,
    is_active   BOOLEAN          DEFAULT TRUE  NOT NULL,
    created_at  TIMESTAMPTZ      DEFAULT NOW() NOT NULL,
    updated_at  TIMESTAMPTZ      DEFAULT NOW() NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
