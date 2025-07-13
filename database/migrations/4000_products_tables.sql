-- +goose Up
-- +goose StatementBegin

---------------------------------------------------------------------------------------------------
-- table products
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS products (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT uuidv7(),
    projects_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- projects_id and name are used to uniquely identify a product within a project
    CONSTRAINT unique_project_product_name UNIQUE (projects_id, name),

    -- serial_id is used for pagination
    serial_id BIGSERIAL NOT NULL UNIQUE
);

-- indexes for products
CREATE INDEX "idx_products_id" ON products (id);
CREATE INDEX "idx_products_projects_id" ON products (projects_id);
CREATE INDEX "idx_products_name" ON products (name);
CREATE INDEX "idx_products_created_at" ON products (created_at);
CREATE INDEX "idx_products_updated_at" ON products (updated_at);
CREATE INDEX "idx_products_pagination" ON products (serial_id, id);

-- +goose StatementEnd
--
-- +goose Down
-- +goose StatementBegin

-- drop indexes for products
DROP INDEX IF EXISTS "idx_products_id";
DROP INDEX IF EXISTS "idx_products_name";
DROP INDEX IF EXISTS "idx_products_created_at";
DROP INDEX IF EXISTS "idx_products_updated_at";
DROP INDEX IF EXISTS "idx_products_pagination";
-- drop products table
DROP TABLE IF EXISTS products;

-- +goose StatementEnd