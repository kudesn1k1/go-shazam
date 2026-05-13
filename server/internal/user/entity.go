package user

import (
	"time"

	"github.com/google/uuid"
)

type UserEntity struct {
	ID             uuid.UUID `db:"id"`
	Email          string    `db:"email"`
	EmailHash      string    `db:"email_hash"`
	HashedPassword string    `db:"hashed_password"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	AvatarFileHash *string   `db:"avatar_file_hash"`
}
