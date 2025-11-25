package fingerprint

import "github.com/google/uuid"

type Peak struct {
	Frequency float64
	Magnitude float64
	Time      float64
	BandIndex int
}

type Hash struct {
	HashValue  int64     `db:"hash"`
	SongID     uuid.UUID `db:"song_id"`
	TimeOffset float64   `db:"time_offset"`
}
