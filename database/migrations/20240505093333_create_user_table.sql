-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
--
-- table for users
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "idx_users_email" ON users (email);
CREATE INDEX "idx_users_created_at" ON users (created_at);
CREATE INDEX "idx_users_updated_at" ON users (updated_at);
CREATE INDEX "idx_users_pagination" ON users (created_at, id);

-- +goose StatementEnd
--
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
--
DROP INDEX "idx_users_email";
DROP INDEX "idx_users_created_at";
DROP INDEX "idx_users_updated_at";
DROP INDEX "idx_users_pagination";

DROP TABLE IF EXISTS users;

-- +goose StatementEnd
