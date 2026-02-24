package user

import "github.com/google/uuid"

type RoleEntity struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type UserRoleEntity struct {
	UserID uuid.UUID `db:"user_id"`
	RoleID int       `db:"role_id"`
}
