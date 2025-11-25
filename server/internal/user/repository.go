package user

import (
	"context"
	"database/sql"
	"errors"
	"go-shazam/internal/core/db"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepositoryInterface interface {
	Create(ctx context.Context, user *UserEntity) error
	FindByID(ctx context.Context, id uuid.UUID) (*UserEntity, error)
	FindByEmailHash(ctx context.Context, emailHash string) (*UserEntity, error)
	ExistsByEmailHash(ctx context.Context, emailHash string) (bool, error)
}

type UserRepository struct {
	db *db.Repository
}

func NewUserRepository(db *db.Repository) UserRepositoryInterface {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *UserEntity) error {
	query := `
		INSERT INTO users (id, email, email_hash, hashed_password, created_at, updated_at)
		VALUES (:id, :email, :email_hash, :hashed_password, :created_at, :updated_at)
	`
	_, err := r.db.Connection(ctx).NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*UserEntity, error) {
	query := "SELECT * FROM users WHERE id = $1"
	var user UserEntity
	if err := r.db.Connection(ctx).GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmailHash(ctx context.Context, emailHash string) (*UserEntity, error) {
	query := "SELECT * FROM users WHERE email_hash = $1"
	var user UserEntity
	if err := r.db.Connection(ctx).GetContext(ctx, &user, query, emailHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ExistsByEmailHash(ctx context.Context, emailHash string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email_hash = $1)"
	var exists bool
	if err := r.db.Connection(ctx).GetContext(ctx, &exists, query, emailHash); err != nil {
		return false, err
	}
	return exists, nil
}
