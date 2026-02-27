package role

import (
	"context"
	"go-shazam/internal/core/db"

	"github.com/google/uuid"
)

type RepositoryInterface interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]RoleEntity, error)
	FindByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]UserRoleRow, error)
	FindByName(ctx context.Context, name string) (*RoleEntity, error)
	FindAll(ctx context.Context) ([]RoleEntity, error)
	AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []int32) error
	RemoveAllForUser(ctx context.Context, userID uuid.UUID) error
}

type Repository struct {
	db *db.Repository
}

func NewRepository(db *db.Repository) RepositoryInterface {
	return &Repository{db: db}
}

func (r *Repository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]RoleEntity, error) {
	query := `
		SELECT ro.id, ro.name FROM roles ro
		INNER JOIN user_roles ur ON ur.role_id = ro.id
		WHERE ur.user_id = $1
	`
	var roles []RoleEntity
	if err := r.db.Connection(ctx).SelectContext(ctx, &roles, query, userID); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repository) FindByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]UserRoleRow, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	query := `
		SELECT ur.user_id, r.id AS role_id, r.name FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = ANY($1)
	`
	var rows []UserRoleRow
	if err := r.db.Connection(ctx).SelectContext(ctx, &rows, query, userIDs); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (*RoleEntity, error) {
	query := "SELECT id, name FROM roles WHERE name = $1"
	var role RoleEntity
	if err := r.db.Connection(ctx).GetContext(ctx, &role, query, name); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) FindAll(ctx context.Context) ([]RoleEntity, error) {
	query := "SELECT id, name FROM roles ORDER BY id"
	var roles []RoleEntity
	if err := r.db.Connection(ctx).SelectContext(ctx, &roles, query); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repository) AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []int32) error {
	if len(roleIDs) == 0 {
		return nil
	}
	query := `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, unnest($2::int[])
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.Connection(ctx).ExecContext(ctx, query, userID, roleIDs)
	return err
}

func (r *Repository) RemoveAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := "DELETE FROM user_roles WHERE user_id = $1"
	_, err := r.db.Connection(ctx).ExecContext(ctx, query, userID)
	return err
}
