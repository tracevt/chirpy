-- +goose Up
CREATE TABLE refresh_tokens(
  token text PRIMARY KEY,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at timestamp NOT NULL,
  revoked_at timestamp
);

-- +goose Down
DROP TABLE refresh_tokens;
