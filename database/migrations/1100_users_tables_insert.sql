-- +goose Up
-- +goose StatementBegin

-- table users
INSERT INTO users (id, first_name, last_name, email, password_hash, disabled, admin) VALUES
-- user Administrator
('019791d2-adef-76d2-a865-5b19e5073e60',
 'Administrator',
 'Default',
 'admin@qu3ry.me',
 '$2a$10$GBQxEIIhh3MBQdFoHZ.wrej4l4ak26X1c5uvrLJmjrjSZ.VWXnX9G', -- password is 'ThisIsApassw0rd.,' hashed with bcrypt and salt
 FALSE,
 TRUE);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- drop all users
DELETE FROM user;

-- +goose StatementEnd
