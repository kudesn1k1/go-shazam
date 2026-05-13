-- +goose Up
CREATE TABLE IF NOT EXISTS files (
    hash          TEXT      PRIMARY KEY,
    content_type  TEXT      NOT NULL,
    size_bytes    INTEGER   NOT NULL,
    status        TEXT      NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at    TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX IF NOT EXISTS idx_files_status ON files(status);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS avatar_file_hash TEXT NULL
    REFERENCES files(hash) ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS avatar_file_hash;
DROP INDEX IF EXISTS idx_files_status;
DROP TABLE IF EXISTS files;
