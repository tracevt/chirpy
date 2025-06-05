-- +goose Up
ALTER TABLE users ADD COLUMN hashed_password text; -- Add as nullable first
UPDATE users SET hashed_password = 'unset'; -- Set a default value for existing rows
ALTER TABLE users ALTER COLUMN hashed_password SET NOT NULL; -- Then make it NOT NULL

-- +goose Down
ALTER TABLE users DROP COLUMN hashed_password;
