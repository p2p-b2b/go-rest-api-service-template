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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "idx_users_email" ON users (email);
CREATE INDEX "idx_users_created_at" ON users (created_at);
CREATE INDEX "idx_users_updated_at" ON users (updated_at);

-- function to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  IF row(NEW.*) IS DISTINCT FROM row(OLD.*) THEN
    NEW.updated_at = now();
    RETURN NEW;
  ELSE
    RETURN OLD;
  END IF;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_updated_at_column BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE  update_updated_at_column();

-- +goose StatementEnd
--
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
--
DROP INDEX "idx_users_email";
DROP INDEX "idx_users_created_at";
DROP INDEX "idx_users_updated_at";

DROP TRIGGER update_updated_at_column ON users;
DROP FUNCTION update_updated_at_column();

DROP TABLE IF EXISTS users;

-- +goose StatementEnd
