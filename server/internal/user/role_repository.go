package user

import (
	"context"
	"go-shazam/internal/core/db"

	"github.com/google/uuid"
)

type RoleRepositoryInterface interface {
	FindRolesByUserID(ctx context.Context, userID uuid.UUID) ([]RoleEntity, error)
	FindRoleByName(ctx context.Context, name string) (*RoleEntity, error)
	AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error
	RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error
	FindAll(ctx context.Context) ([]RoleEntity, error)
}

type RoleRepository struct {
	db *db.Repository
}

func NewRoleRepository(db *db.Repository) RoleRepositoryInterface {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindRolesByUserID(ctx context.Context, userID uuid.UUID) ([]RoleEntity, error) {
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

func (r *RoleRepository) FindRoleByName(ctx context.Context, name string) (*RoleEntity, error) {
	query := "SELECT id, name FROM roles WHERE name = $1"
	var role RoleEntity
	if err := r.db.Connection(ctx).GetContext(ctx, &role, query, name); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	query := "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	_, err := r.db.Connection(ctx).ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *RoleRepository) RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error {
	query := "DELETE FROM user_roles WHERE user_id = $1"
	_, err := r.db.Connection(ctx).ExecContext(ctx, query, userID)
	return err
}

func (r *RoleRepository) FindAll(ctx context.Context) ([]RoleEntity, error) {
	query := "SELECT id, name FROM roles ORDER BY id"
	var roles []RoleEntity
	if err := r.db.Connection(ctx).SelectContext(ctx, &roles, query); err != nil {
		return nil, err
	}
	return roles, nil
}
