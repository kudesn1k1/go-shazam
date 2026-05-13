package files

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-shazam/internal/core/db"
)

type FilesRepositoryInterface interface {
	FindByHash(ctx context.Context, hash string) (*FileEntity, error)
	Insert(ctx context.Context, e *FileEntity) error
	BumpUpdatedAt(ctx context.Context, hash string) error
	UpdateStatus(ctx context.Context, hash, status string) error
	IsReferenced(ctx context.Context, hash string) (bool, error)
	ListExpired(ctx context.Context, olderThan time.Time, limit int) ([]string, error)
	DeleteByHash(ctx context.Context, hash string) error
}

type FilesRepository struct {
	db *db.Repository
}

func NewFilesRepository(db *db.Repository) FilesRepositoryInterface {
	return &FilesRepository{db: db}
}

func (r *FilesRepository) FindByHash(ctx context.Context, hash string) (*FileEntity, error) {
	query := `SELECT hash, content_type, size_bytes, status, created_at, updated_at FROM files WHERE hash = $1`
	var f FileEntity
	if err := r.db.Connection(ctx).GetContext(ctx, &f, query, hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("files: FindByHash: %w", err)
	}
	return &f, nil
}

func (r *FilesRepository) Insert(ctx context.Context, e *FileEntity) error {
	query := `
		INSERT INTO files (hash, content_type, size_bytes, status, created_at, updated_at)
		VALUES (:hash, :content_type, :size_bytes, :status, :created_at, :updated_at)
		ON CONFLICT (hash) DO UPDATE SET updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Connection(ctx).NamedExecContext(ctx, query, e)
	return err
}

func (r *FilesRepository) BumpUpdatedAt(ctx context.Context, hash string) error {
	query := `UPDATE files SET updated_at = (NOW() AT TIME ZONE 'UTC') WHERE hash = $1`
	_, err := r.db.Connection(ctx).ExecContext(ctx, query, hash)
	return err
}

func (r *FilesRepository) UpdateStatus(ctx context.Context, hash, status string) error {
	query := `UPDATE files SET status = $1, updated_at = (NOW() AT TIME ZONE 'UTC') WHERE hash = $2`
	_, err := r.db.Connection(ctx).ExecContext(ctx, query, status, hash)
	return err
}

func (r *FilesRepository) IsReferenced(ctx context.Context, hash string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE avatar_file_hash = $1)`
	var exists bool
	if err := r.db.Connection(ctx).GetContext(ctx, &exists, query, hash); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *FilesRepository) ListExpired(ctx context.Context, olderThan time.Time, limit int) ([]string, error) {
	query := `
		SELECT hash FROM files
		WHERE status = $1 AND updated_at < $2
		LIMIT $3
	`
	var hashes []string
	if err := r.db.Connection(ctx).SelectContext(ctx, &hashes, query, StatusTemporary, olderThan, limit); err != nil {
		return nil, err
	}
	return hashes, nil
}

func (r *FilesRepository) DeleteByHash(ctx context.Context, hash string) error {
	query := `DELETE FROM files WHERE hash = $1`
	_, err := r.db.Connection(ctx).ExecContext(ctx, query, hash)
	return err
}
