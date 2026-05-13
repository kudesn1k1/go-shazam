package song

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-shazam/internal/core/db"
	"strings"

	"github.com/google/uuid"
)

type SongRepositoryInterface interface {
	Save(ctx context.Context, song *SongEntity) error
	FindByID(ctx context.Context, id uuid.UUID) (*SongEntity, error)
	FindByTitleAndArtist(ctx context.Context, title string, artist string) (*SongEntity, error)
	FindFiltered(ctx context.Context, f SongFilter) ([]SongEntity, error)
	CountFiltered(ctx context.Context, f SongFilter) (int, error)
	Delete(ctx context.Context, id uuid.UUID) error
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
		INSERT INTO songs (id, title, artist, duration, source_id, uploaded_by, created_at)
		VALUES (:id, :title, :artist, :duration, :source_id, :uploaded_by, :created_at)
	`
	_, err := r.db.Connection(ctx).NamedExecContext(ctx, query, song)
	return err
}

func (r *SongRepository) FindFiltered(ctx context.Context, f SongFilter) ([]SongEntity, error) {
	var sb strings.Builder
	args := make([]any, 0, 8)
	sb.WriteString("SELECT * FROM songs")
	r.appendFilterClauses(&sb, &args, f)
	r.appendOrderClause(&sb, f)
	sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2))
	args = append(args, f.Limit, f.Offset)

	var songs []SongEntity
	if err := r.db.Connection(ctx).SelectContext(ctx, &songs, sb.String(), args...); err != nil {
		return nil, err
	}
	return songs, nil
}

func (r *SongRepository) CountFiltered(ctx context.Context, f SongFilter) (int, error) {
	var sb strings.Builder
	args := make([]any, 0, 8)
	sb.WriteString("SELECT COUNT(*) FROM songs")
	r.appendFilterClauses(&sb, &args, f)

	var count int
	if err := r.db.Connection(ctx).GetContext(ctx, &count, sb.String(), args...); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SongRepository) appendFilterClauses(sb *strings.Builder, args *[]any, f SongFilter) {
	clauses := []string{}

	if f.Q != "" {
		esc := escapeLike(f.Q)
		*args = append(*args, "%"+esc+"%")
		clauses = append(clauses, fmt.Sprintf("(title ILIKE $%d ESCAPE '\\' OR artist ILIKE $%d ESCAPE '\\')", len(*args), len(*args)))
	}
	if f.Artist != "" {
		*args = append(*args, f.Artist)
		clauses = append(clauses, fmt.Sprintf("artist = $%d", len(*args)))
	}
	if f.UploadedBy != nil {
		*args = append(*args, *f.UploadedBy)
		clauses = append(clauses, fmt.Sprintf("uploaded_by = $%d", len(*args)))
	}
	if f.CreatedAfter != nil {
		*args = append(*args, *f.CreatedAfter)
		clauses = append(clauses, fmt.Sprintf("created_at >= $%d", len(*args)))
	}
	if f.CreatedBefore != nil {
		*args = append(*args, *f.CreatedBefore)
		clauses = append(clauses, fmt.Sprintf("created_at < $%d", len(*args)))
	}

	if len(clauses) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(clauses, " AND "))
	}
}

func (r *SongRepository) appendOrderClause(sb *strings.Builder, f SongFilter) {
	col := string(f.Sort)
	dir := "DESC"
	if f.Order == SortAsc {
		dir = "ASC"
	}
	sb.WriteString(fmt.Sprintf(" ORDER BY %s %s", col, dir))
}

func escapeLike(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")
	return r.Replace(s)
}

func (r *SongRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM songs WHERE id = $1"
	_, err := r.db.Connection(ctx).ExecContext(ctx, query, id)
	return err
}
