-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id            UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    full_name     VARCHAR(255),
    phone         VARCHAR(50),
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL ,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL ,
    deleted_at    TIMESTAMP WITH TIME ZONE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
