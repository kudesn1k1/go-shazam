package fingerprint

import (
	"context"
	"fmt"
	"go-shazam/internal/core/db"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *db.Repository
}

func NewRepository(db *db.Repository) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindHashesByValues(ctx context.Context, hashValues []int64) ([]Hash, error) {
	if len(hashValues) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In("SELECT hash, song_id, time_offset FROM fingerprints WHERE hash IN (?)", hashValues)
	if err != nil {
		return nil, err
	}
	//TODO: check rebind
	query = sqlx.Rebind(sqlx.DOLLAR, query)

	var hashes []Hash
	if err := r.db.Connection(ctx).SelectContext(ctx, &hashes, query, args...); err != nil {
		return nil, err
	}

	return hashes, nil
}

func (r *Repository) SaveFingerprints(ctx context.Context, hashes []Hash) error {
	if len(hashes) == 0 {
		return nil
	}

	//TODO: reconsider chunk size to reduce database load
	chunkSize := 1000
	for i := 0; i < len(hashes); i += chunkSize {
		end := i + chunkSize
		if end > len(hashes) {
			end = len(hashes)
		}

		chunk := hashes[i:end]
		if err := r.insertChunk(ctx, chunk); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) insertChunk(ctx context.Context, chunk []Hash) error {
	query := "INSERT INTO fingerprints (hash, song_id, time_offset) VALUES "
	values := []interface{}{}
	placeholders := []string{}

	for i, h := range chunk {
		base := i * 3
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", base+1, base+2, base+3))
		values = append(values, h.HashValue, h.SongID, h.TimeOffset)
	}

	query += strings.Join(placeholders, ", ")

	_, err := r.db.Connection(ctx).ExecContext(ctx, query, values...)
	return err
}
