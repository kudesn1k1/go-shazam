package song

import (
	"context"
	"database/sql"
	"errors"
	"go-shazam/internal/core/db"

	"github.com/google/uuid"
)

type SongRepositoryInterface interface {
	Save(ctx context.Context, song *SongEntity) error
	FindByID(ctx context.Context, id uuid.UUID) (*SongEntity, error)
	FindByTitleAndArtist(ctx context.Context, title string, artist string) (*SongEntity, error)
	FindByUploadedBy(ctx context.Context, userID uuid.UUID, limit, offset int) ([]SongEntity, error)
	CountByUploadedBy(ctx context.Context, userID uuid.UUID) (int, error)
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) FindByTitleAndArtist(ctx context.Context, title string, artist string) (*SongEntity, error) {
	query := "SELECT * FROM songs WHERE title = $1 AND artist = $2"
	var song SongEntity
	if err := r.db.Connection(ctx).GetContext(ctx, &song, query, title, artist); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) Save(ctx context.Context, song *SongEntity) error {
	query := `
		INSERT INTO songs (id, title, artist, duration, source_id, uploaded_by)
		VALUES (:id, :title, :artist, :duration, :source_id, :uploaded_by)
	`
	_, err := r.db.Connection(ctx).NamedExecContext(ctx, query, song)
	return err
}

func (r *SongRepository) FindByUploadedBy(ctx context.Context, userID uuid.UUID, limit, offset int) ([]SongEntity, error) {
	query := "SELECT * FROM songs WHERE uploaded_by = $1 ORDER BY id DESC LIMIT $2 OFFSET $3"
	var songs []SongEntity
	if err := r.db.Connection(ctx).SelectContext(ctx, &songs, query, userID, limit, offset); err != nil {
		return nil, err
	}
	return songs, nil
}

func (r *SongRepository) CountByUploadedBy(ctx context.Context, userID uuid.UUID) (int, error) {
	query := "SELECT COUNT(*) FROM songs WHERE uploaded_by = $1"
	var count int
	if err := r.db.Connection(ctx).GetContext(ctx, &count, query, userID); err != nil {
		return 0, err
	}
	return count, nil
}
