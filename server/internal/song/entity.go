package song

import "github.com/google/uuid"

type SongEntity struct {
	ID       uuid.UUID `db:"id"`
	Title    string    `db:"title"`
	Artist   string    `db:"artist"`
	Duration int       `db:"duration"`
	SourceID string    `db:"source_id"`
}
