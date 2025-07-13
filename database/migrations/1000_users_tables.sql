-- +goose Up
-- +goose StatementBegin

---------------------------------------------------------------------------------------------------
-- table for users
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT uuidv7(),
    first_name VARCHAR(25) NOT NULL,
    last_name VARCHAR(25) NOT NULL,
    email VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    disabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- email is unique
    CONSTRAINT "users_email" UNIQUE (email),

    -- serial_id is used for pagination
    serial_id BIGSERIAL NOT NULL UNIQUE
);

-- indexes for users
CREATE INDEX "idx_users_id" ON users (id);
CREATE INDEX "idx_users_email" ON users (email);
CREATE INDEX "idx_users_created_at" ON users (created_at);
CREATE INDEX "idx_users_updated_at" ON users (updated_at);
CREATE INDEX "idx_users_pagination" ON users (serial_id, id);

-- +goose StatementEnd
--
-- +goose Down
-- +goose StatementBegin

-- drop indexes for users
DROP INDEX IF EXISTS "idx_users_id";
DROP INDEX IF EXISTS "idx_users_email";
DROP INDEX IF EXISTS "idx_users_created_at";
DROP INDEX IF EXISTS "idx_users_updated_at";
DROP INDEX IF EXISTS "idx_users_pagination";

-- drop users table
DROP TABLE IF EXISTS users;

-- +goose StatementEnd