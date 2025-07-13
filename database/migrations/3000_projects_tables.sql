-- +goose Up
-- +goose StatementBegin

---------------------------------------------------------------------------------------------------
-- table projects
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT uuidv7(),
    name VARCHAR(70) NOT NULL,
    description TEXT NOT NULL,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    system BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- name is unique
    CONSTRAINT unique_project_name UNIQUE (name),

    -- serial_id is used for pagination
    serial_id BIGSERIAL NOT NULL UNIQUE
);

-- indexes for projects
CREATE INDEX "idx_projects_id" ON projects (id);
CREATE INDEX "idx_projects_name" ON projects (name);
CREATE INDEX "idx_projects_created_at" ON projects (created_at);
CREATE INDEX "idx_projects_updated_at" ON projects (updated_at);
CREATE INDEX "idx_projects_pagination" ON projects (serial_id, id);

-- trigger to restrict delete and update on system projects
CREATE OR REPLACE FUNCTION fn_restrict_delete_update_on_system_projects()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' AND OLD.system THEN
        RAISE EXCEPTION 'System projects cannot be deleted.';
    ELSIF TG_OP = 'UPDATE' AND OLD.system THEN
        RAISE EXCEPTION 'System projects cannot be updated.';
    ELSEIF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSEIF TG_OP = 'UPDATE' THEN
        OLD = NEW;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_restrict_delete_update_on_system_projects
BEFORE DELETE OR UPDATE ON projects
FOR EACH ROW
EXECUTE FUNCTION fn_restrict_delete_update_on_system_projects();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- drop indexes for projects
DROP INDEX IF EXISTS "idx_projects_id";
DROP INDEX IF EXISTS "idx_projects_name";
DROP INDEX IF EXISTS "idx_projects_created_at";
DROP INDEX IF EXISTS "idx_projects_updated_at";
DROP INDEX IF EXISTS "idx_projects_pagination";

-- drop projects table
DROP TABLE IF EXISTS projects;

-- +goose StatementEnd