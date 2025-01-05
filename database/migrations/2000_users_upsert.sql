-- +goose Up
-- +goose StatementBegin

-- table users
INSERT INTO users (id, first_name, last_name, email, password_hash, disabled) VALUES
('483688e9-b8af-42c7-bba4-fb4a29cb7887', 'user', '1',  'user.1@mail.com',  '$2a$10$KxiXRbNgEL5In8zZqfB9y.ossfaRfrw8uaK4OH/OIAInxekGP8yQO', FALSE); -- password is 'ThisIsAPassword123' hashed with bcrypt and salt -- password is 'ThisIsAPassword123' hashed with bcrypt and salt

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- delete users
DELETE FROM users;

-- +goose StatementEnd
