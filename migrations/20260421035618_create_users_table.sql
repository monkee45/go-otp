-- +goose Up
SELECT 'up SQL query';
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    otp TEXT,
    otpexpiry TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE users; 
