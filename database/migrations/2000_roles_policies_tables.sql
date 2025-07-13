-- +goose Up
-- +goose StatementBegin

---------------------------------------------------------------------------------------------------
-- table roles
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    system BOOLEAN NOT NULL DEFAULT FALSE,
    auto_assign BOOLEAN NOT NULL DEFAULT FALSE, -- this is used to set the auto_assign role for new users
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- name are unique
    CONSTRAINT "roles_name" UNIQUE (name),

    -- serial_id is used for pagination
    serial_id BIGSERIAL NOT NULL UNIQUE
);

-- indexes for roles
CREATE INDEX "idx_roles_id" ON roles (id);
CREATE INDEX "idx_roles_name" ON roles (name);
CREATE INDEX "idx_roles_created_at" ON roles (created_at);
CREATE INDEX "idx_roles_updated_at" ON roles (updated_at);
CREATE INDEX "idx_roles_pagination" ON roles (serial_id, id);

-- trigger to restrict delete and update on system roles
CREATE OR REPLACE FUNCTION fn_restrict_delete_update_on_system_roles()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' AND OLD.system THEN
        RAISE EXCEPTION 'System roles cannot be deleted.';
    ELSIF TG_OP = 'UPDATE' AND OLD.system THEN
        RAISE EXCEPTION 'System roles cannot be updated.';
    ELSEIF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSEIF TG_OP = 'UPDATE' THEN
        OLD = NEW;
		    RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER tr_restrict_delete_update_on_system_roles
BEFORE DELETE OR UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION fn_restrict_delete_update_on_system_roles();

---------------------------------------------------------------------------------------------------
-- table resources
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS resources (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    action VARCHAR(255) NOT NULL,
    resource VARCHAR(512) NOT NULL,
    system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- name, action and resource are unique
    CONSTRAINT "permissions_name_action_resource" UNIQUE (name, action, resource),

    -- serial_id is used for pagination
    serial_id BIGSERIAL NOT NULL UNIQUE
);

-- indexes for resources
CREATE INDEX "idx_permissions_id" ON resources (id);
CREATE INDEX "idx_permissions_action" ON resources (action);
CREATE INDEX "idx_permissions_resource" ON resources (resource);
CREATE INDEX "idx_permissions_created_at" ON resources (created_at);
CREATE INDEX "idx_permissions_updated_at" ON resources (updated_at);
CREATE INDEX "idx_permissions_pagination" ON resources (serial_id, id);

-- trigger to restrict delete and update on system resources
CREATE OR REPLACE FUNCTION fn_restrict_delete_update_on_system_permissions()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' AND OLD.system THEN
        RAISE EXCEPTION 'System resources cannot be deleted.';
    ELSIF TG_OP = 'UPDATE' AND OLD.system THEN
        RAISE EXCEPTION 'System resources cannot be updated.';
    ELSEIF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSEIF TG_OP = 'UPDATE' THEN
        OLD = NEW;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_restrict_delete_update_on_system_permissions
BEFORE DELETE OR UPDATE ON resources
FOR EACH ROW
EXECUTE FUNCTION fn_restrict_delete_update_on_system_permissions();

---------------------------------------------------------------------------------------------------
-- table policies
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS policies (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT uuidv7(),
    resources_id uuid NOT NULL REFERENCES resources (id) ON DELETE CASCADE ON UPDATE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    allowed_action VARCHAR(255) NOT NULL,
    allowed_resource VARCHAR(512) NOT NULL,
    system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- name, allowed_action and allowed_resource are unique
    CONSTRAINT "policies_name_allowed_action_allowed_resource" UNIQUE (name, allowed_action, allowed_resource),

    -- serial_id is used for pagination
    serial_id BIGSERIAL NOT NULL UNIQUE
);

-- indexes for policies
CREATE INDEX "idx_policies_id" ON policies (id);
CREATE INDEX "idx_policies_resources_id" ON policies (resources_id);
CREATE INDEX "idx_policies_name" ON policies (name);
CREATE INDEX "idx_policies_allowed_action" ON policies (allowed_action);
CREATE INDEX "idx_policies_allowed_resource" ON policies (allowed_resource);
CREATE INDEX "idx_policies_created_at" ON policies (created_at);
CREATE INDEX "idx_policies_updated_at" ON policies (updated_at);
CREATE INDEX "idx_policies_pagination" ON policies (serial_id, id);

-- trigger to restrict delete and update on system policies
CREATE OR REPLACE FUNCTION fn_restrict_delete_update_on_system_policies()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' AND OLD.system THEN
        RAISE EXCEPTION 'System policies cannot be deleted.';
    ELSIF TG_OP = 'UPDATE' AND OLD.system THEN
        RAISE EXCEPTION 'System policies cannot be updated.';
    ELSEIF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSEIF TG_OP = 'UPDATE' THEN
        OLD = NEW;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_restrict_delete_update_on_system_policies
BEFORE DELETE OR UPDATE ON policies
FOR EACH ROW
EXECUTE FUNCTION fn_restrict_delete_update_on_system_policies();

---------------------------------------------------------------------------------------------------
-- table roles_policies
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles_policies (
    roles_id uuid NOT NULL REFERENCES roles (id) ON DELETE CASCADE ON UPDATE CASCADE,
    policies_id uuid NOT NULL REFERENCES policies (id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- PRIMARY KEY
    PRIMARY KEY (roles_id, policies_id)
);

-- indexes for roles_policies
CREATE INDEX "idx_roles_policies_roles_id" ON roles_policies (roles_id);
CREATE INDEX "idx_roles_policies_policies_id" ON roles_policies (policies_id);
CREATE INDEX "idx_roles_policies_created_at" ON roles_policies (created_at);
CREATE INDEX "idx_roles_policies_updated_at" ON roles_policies (updated_at);
CREATE INDEX "idx_roles_policies_roles_id_policies_id" ON roles_policies (roles_id, policies_id);


-- table users_roles
CREATE TABLE IF NOT EXISTS users_roles (
    users_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    roles_id uuid NOT NULL REFERENCES roles (id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- PRIMARY KEY
    PRIMARY KEY (users_id, roles_id)
);

-- indexes for users_roles
CREATE INDEX "idx_users_roles_users_id" ON users_roles (users_id);
CREATE INDEX "idx_users_roles_roles_id" ON users_roles (roles_id);
CREATE INDEX "idx_users_roles_created_at" ON users_roles (created_at);
CREATE INDEX "idx_users_roles_updated_at" ON users_roles (updated_at);
CREATE INDEX "idx_users_roles_id" ON users_roles (users_id, roles_id);

-- +goose StatementEnd
--
-- +goose Down
-- +goose StatementBegin

-- drop indexes for roles
DROP INDEX IF EXISTS "idx_roles_id";
DROP INDEX IF EXISTS "idx_roles_name";
DROP INDEX IF EXISTS "idx_roles_created_at";
DROP INDEX IF EXISTS "idx_roles_updated_at";
DROP INDEX IF EXISTS "idx_roles_pagination";

-- drop roles table
DROP TABLE IF EXISTS roles;

-- drop indexes for resources
DROP INDEX IF EXISTS "idx_permissions_id";
DROP INDEX IF EXISTS "idx_permissions_action";
DROP INDEX IF EXISTS "idx_permissions_resource";
DROP INDEX IF EXISTS "idx_permissions_created_at";
DROP INDEX IF EXISTS "idx_permissions_updated_at";
DROP INDEX IF EXISTS "idx_permissions_pagination";

-- drop resources table
DROP TABLE IF EXISTS resources;

-- drop indexes for policies
DROP INDEX IF EXISTS "idx_policies_id";
DROP INDEX IF EXISTS "idx_policies_resources_id";
DROP INDEX IF EXISTS "idx_policies_name";
DROP INDEX IF EXISTS "idx_policies_allowed_action";
DROP INDEX IF EXISTS "idx_policies_allowed_resource";
DROP INDEX IF EXISTS "idx_policies_created_at";
DROP INDEX IF EXISTS "idx_policies_updated_at";
DROP INDEX IF EXISTS "idx_policies_pagination";

-- drop policies table
DROP TABLE IF EXISTS policies;

-- drop indexes for roles_policies
DROP INDEX IF EXISTS "idx_roles_policies_roles_id";
DROP INDEX IF EXISTS "idx_roles_policies_policies_id";
DROP INDEX IF EXISTS "idx_roles_policies_created_at";
DROP INDEX IF EXISTS "idx_roles_policies_updated_at";
DROP INDEX IF EXISTS "idx_roles_policies_roles_id_policies_id";

-- drop roles_policies table
DROP TABLE IF EXISTS roles_policies;

-- +goose StatementEnd
