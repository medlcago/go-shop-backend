-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories
(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(255) NOT NULL,
    slug      VARCHAR(255) NOT NULL,
    parent_id UUID,

    CONSTRAINT check_parent_not_self CHECK (parent_id != id OR parent_id IS NULL),

    FOREIGN KEY (parent_id) REFERENCES categories (id) ON DELETE CASCADE
);

CREATE INDEX idx_categories_parent_id ON categories (parent_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
