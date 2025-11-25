-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL,             -- Encrypted email (AES-256-GCM)
    email_hash TEXT NOT NULL UNIQUE, -- SHA-256 hash for lookups
    hashed_password TEXT NOT NULL,   -- bcrypt hash
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email_hash ON users(email_hash);

-- +goose Down
DROP TABLE IF EXISTS users;

