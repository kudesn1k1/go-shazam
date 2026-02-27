package role

import "github.com/google/uuid"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) String() string { return string(r) }

type RoleEntity struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type UserRoleRow struct {
	UserID uuid.UUID `db:"user_id"`
	RoleID int       `db:"role_id"`
	Name   string    `db:"name"`
}
