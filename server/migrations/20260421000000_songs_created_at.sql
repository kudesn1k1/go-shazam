-- +goose Up
ALTER TABLE songs ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC');
CREATE INDEX IF NOT EXISTS idx_songs_created_at ON songs(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_songs_created_at;
ALTER TABLE songs DROP COLUMN IF EXISTS created_at;
