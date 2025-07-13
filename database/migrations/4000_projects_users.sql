-- +goose Up
-- +goose StatementBegin

---------------------------------------------------------------------------------------------------
-- table projects_users
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS projects_users (
    projects_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    users_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- primary key to ensure unique project-user pairs
    PRIMARY KEY (projects_id, users_id)
);
-- indexes for projects_users
CREATE INDEX "idx_projects_users_projects_id" ON projects_users (projects_id);
CREATE INDEX "idx_projects_users_users_id" ON projects_users (users_id);
CREATE INDEX "idx_projects_users_created_at" ON projects_users (created_at);
CREATE INDEX "idx_projects_users_updated_at" ON projects_users (updated_at);
CREATE INDEX "idx_projects_users_project_user" ON projects_users (projects_id, users_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- drop indexes for projects_users
DROP INDEX IF EXISTS "idx_projects_users_projects_id";
DROP INDEX IF EXISTS "idx_projects_users_users_id";
DROP INDEX IF EXISTS "idx_projects_users_created_at";
DROP INDEX IF EXISTS "idx_projects_users_updated_at";
DROP INDEX IF EXISTS "idx_projects_users_project_user";

-- drop projects_users table
DROP TABLE IF EXISTS projects_users;


-- +goose StatementEnd