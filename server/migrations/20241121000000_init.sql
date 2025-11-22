-- +goose Up
CREATE TABLE IF NOT EXISTS songs (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    artist TEXT NOT NULL,
    duration INTEGER NOT NULL, -- Duration in milliseconds
    source_id TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_songs_title_artist ON songs(title, artist);

CREATE TABLE IF NOT EXISTS fingerprints (
    id BIGSERIAL PRIMARY KEY,
    hash BIGINT NOT NULL,
    song_id UUID NOT NULL,
    time_offset DOUBLE PRECISION NOT NULL,
    CONSTRAINT fk_song
        FOREIGN KEY(song_id) 
        REFERENCES songs(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_fingerprints_hash ON fingerprints(hash);
CREATE INDEX IF NOT EXISTS idx_fingerprints_song_id ON fingerprints(song_id);

-- +goose Down
DROP TABLE IF EXISTS fingerprints;
DROP TABLE IF EXISTS songs;

