package song

import (
	"time"

	"github.com/google/uuid"
)

type SongEntity struct {
	ID         uuid.UUID  `db:"id"`
	Title      string     `db:"title"`
	Artist     string     `db:"artist"`
	Duration   int        `db:"duration"`
	SourceID   string     `db:"source_id"`
	UploadedBy *uuid.UUID `db:"uploaded_by"`
	CreatedAt  time.Time  `db:"created_at"`
}
