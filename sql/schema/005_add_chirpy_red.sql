-- +goose Up
ALTER TABLE users ADD COLUMN is_chirpy_red boolean; -- Add as nullable first
UPDATE users SET is_chirpy_red = false; -- Set a default value for existing rows
ALTER TABLE users ALTER COLUMN is_chirpy_red SET NOT NULL; -- Then make it NOT NULL

-- +goose Down
ALTER TABLE users DROP COLUMN is_chirpy_red;
