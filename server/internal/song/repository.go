package song

import (
	"context"
	"go-shazam/internal/core/db"

	"github.com/google/uuid"
)

type SongRepositoryInterface interface {
	Save(ctx context.Context, song *SongEntity) error
	FindByID(ctx context.Context, id uuid.UUID) (*SongEntity, error)
	FindByTitleAndArtist(ctx context.Context, title string, artist string) (*SongEntity, error)
}

type SongRepository struct {
	db *db.Repository
}

func NewSongRepository(db *db.Repository) SongRepositoryInterface {
	return &SongRepository{db: db}
}

func (r *SongRepository) FindByID(ctx context.Context, id uuid.UUID) (*SongEntity, error) {
	query := "SELECT * FROM songs WHERE id = $1"
	var song SongEntity
	if err := r.db.Connection(ctx).GetContext(ctx, &song, query, id); err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) FindByTitleAndArtist(ctx context.Context, title string, artist string) (*SongEntity, error) {
	query := "SELECT * FROM songs WHERE title = $1 AND artist = $2"
	var song SongEntity
	if err := r.db.Connection(ctx).GetContext(ctx, &song, query, title, artist); err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) Save(ctx context.Context, song *SongEntity) error {
	query := `
		INSERT INTO songs (id, title, artist, duration, source_id)
		VALUES (:id, :title, :artist, :duration, :source_id)
	`
	_, err := r.db.Connection(ctx).NamedExecContext(ctx, query, song)
	return err
}
