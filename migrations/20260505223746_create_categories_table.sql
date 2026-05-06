-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories
(
    id          UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    parent_id   UUID         REFERENCES categories (id) ON DELETE SET NULL,

    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL,

    description TEXT,

    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order  INT          NOT NULL DEFAULT 0,

    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,

    -- protection against self-references
    CHECK (parent_id IS NULL OR id <> parent_id)
);

CREATE INDEX idx_categories_parent_id ON categories (parent_id);

CREATE UNIQUE INDEX idx_categories_slug_unique ON categories (slug)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_categories_deleted_at ON categories (deleted_at);

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
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS product_categories;
-- +goose StatementEnd
