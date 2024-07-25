-- +goose Up
-- +goose StatementBegin

-- table for users
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- email is unique
    CONSTRAINT "uq_users_email" UNIQUE (email),

    -- serial_id is used for pagination
    serial_id BIGSERIAL NOT NULL UNIQUE
);

CREATE INDEX "idx_users_email" ON users (email);
CREATE INDEX "idx_users_created_at" ON users (created_at);
CREATE INDEX "idx_users_updated_at" ON users (updated_at);
CREATE INDEX "idx_users_pagination" ON users (serial_id, id);

-- +goose StatementEnd
--
-- +goose Down
-- +goose StatementBegin

-- drop table users
DROP INDEX "idx_users_email";
DROP INDEX "idx_users_created_at";
DROP INDEX "idx_users_updated_at";
DROP INDEX "idx_users_pagination";

DROP TABLE IF EXISTS users;

-- +goose StatementEnd
