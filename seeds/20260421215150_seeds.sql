-- +goose Up
SELECT 'up SQL query';

INSERT INTO users (name, email, otpExpiry)
VALUES
  ('Alice Smith', 'alice@example.com', ),
  ('Bob Johnson', 'bob@example.com')
ON CONFLICT (email) DO NOTHING;

-- +goose Down
SELECT 'down SQL query';
